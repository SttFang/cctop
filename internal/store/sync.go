package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fanghanjun/cctop/internal/data"
)

// SyncResult contains statistics about a sync operation.
type SyncResult struct {
	FilesProcessed int
	FilesSkipped   int
	Errors         int
}

// SyncSessions scans the projects directory for JSONL files and indexes them.
func (s *Store) SyncSessions(projectsDir string) (SyncResult, error) {
	var result SyncResult

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return result, fmt.Errorf("reading projects dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projDir := filepath.Join(projectsDir, entry.Name())
		files, err := filepath.Glob(filepath.Join(projDir, "*.jsonl"))
		if err != nil {
			continue
		}

		for _, filePath := range files {
			// Skip agent files
			base := filepath.Base(filePath)
			if strings.HasPrefix(base, "agent-") {
				continue
			}

			processed, err := s.syncFile(filePath)
			if err != nil {
				result.Errors++
				continue
			}
			if processed {
				result.FilesProcessed++
			} else {
				result.FilesSkipped++
			}
		}
	}

	return result, nil
}

// syncFile processes a single JSONL file if it has been modified.
func (s *Store) syncFile(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	mtime := info.ModTime().Unix()
	size := info.Size()

	// Check if file has changed
	var storedMtime, storedSize int64
	err = s.db.QueryRow("SELECT mtime, size FROM sync_state WHERE file_path = ?", filePath).
		Scan(&storedMtime, &storedSize)

	if err == nil && storedMtime == mtime && storedSize == size {
		return false, nil // not modified
	}

	// Parse the session file
	summary, err := data.ParseSessionFile(filePath)
	if err != nil {
		return false, err
	}

	if summary.MessageCount == 0 {
		// Skip empty sessions
		return false, nil
	}

	// Upsert session
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT OR REPLACE INTO sessions
		(session_id, project, model, start_time, duration_ms, input_tokens, output_tokens,
		 cache_read, cache_create, total_tokens, cost, message_count, first_prompt, file_path, file_mtime)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		summary.SessionID, summary.Project, summary.Model, summary.StartTime,
		summary.DurationMs, summary.InputTokens, summary.OutputTokens,
		summary.CacheRead, summary.CacheCreate, summary.TotalTokens,
		summary.Cost, summary.MessageCount, summary.FirstPrompt, filePath, mtime)
	if err != nil {
		return false, err
	}

	// Delete old tool calls and insert new ones
	_, err = tx.Exec("DELETE FROM tool_calls WHERE session_id = ?", summary.SessionID)
	if err != nil {
		return false, err
	}

	for toolName, count := range summary.ToolCalls {
		_, err = tx.Exec("INSERT INTO tool_calls (session_id, tool_name, call_count) VALUES (?, ?, ?)",
			summary.SessionID, toolName, count)
		if err != nil {
			return false, err
		}
	}

	// Update sync state
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO sync_state (file_path, mtime, size) VALUES (?, ?, ?)`,
		filePath, mtime, size)
	if err != nil {
		return false, err
	}

	return true, tx.Commit()
}

// GetSyncState returns the number of indexed files.
func (s *Store) GetSyncState() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sync_state").Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return count, nil
}
