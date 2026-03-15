package store

import (
	"database/sql"
	"fmt"

	"github.com/fanghanjun/cctop/internal/data"
)

// SessionFilter defines parameters for querying sessions.
type SessionFilter struct {
	Project  string
	Model    string
	Search   string
	SortBy   string // "time", "cost", "tokens", "duration"
	SortDesc bool
	Limit    int
	Offset   int
}

// SessionRow is a session row from the database.
type SessionRow struct {
	data.SessionSummary
}

// GetSessions queries sessions with filtering and sorting.
func (s *Store) GetSessions(f SessionFilter) ([]SessionRow, int, error) {
	where := "1=1"
	var args []any

	if f.Project != "" {
		where += " AND project = ?"
		args = append(args, f.Project)
	}
	if f.Model != "" {
		where += " AND model LIKE ?"
		args = append(args, "%"+f.Model+"%")
	}
	if f.Search != "" {
		where += " AND (first_prompt LIKE ? OR project LIKE ?)"
		args = append(args, "%"+f.Search+"%", "%"+f.Search+"%")
	}

	// Count total
	var total int
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM sessions WHERE %s", where)
	if err := s.db.QueryRow(countQ, args...).Scan(&total); err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	// Sort
	orderCol := "start_time"
	switch f.SortBy {
	case "cost":
		orderCol = "cost"
	case "tokens":
		orderCol = "total_tokens"
	case "duration":
		orderCol = "duration_ms"
	}
	order := "DESC"
	if !f.SortDesc {
		order = "ASC"
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(`
		SELECT session_id, project, model, start_time, duration_ms,
		       input_tokens, output_tokens, cache_read, cache_create,
		       total_tokens, cost, message_count, first_prompt
		FROM sessions WHERE %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		where, orderCol, order)

	args = append(args, limit, f.Offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sessions []SessionRow
	for rows.Next() {
		var r SessionRow
		err := rows.Scan(
			&r.SessionID, &r.Project, &r.Model, &r.StartTime, &r.DurationMs,
			&r.InputTokens, &r.OutputTokens, &r.CacheRead, &r.CacheCreate,
			&r.TotalTokens, &r.Cost, &r.MessageCount, &r.FirstPrompt)
		if err != nil {
			continue
		}
		sessions = append(sessions, r)
	}

	return sessions, total, nil
}

// GetProjects returns a list of distinct project names.
func (s *Store) GetProjects() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT project FROM sessions WHERE project != '' ORDER BY project")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err == nil {
			projects = append(projects, p)
		}
	}
	return projects, nil
}

// ToolFreq represents tool usage frequency.
type ToolFreq struct {
	Name      string
	TotalCalls int
	Percent    float64
}

// GetToolFrequency returns tool usage frequency, optionally filtered by time.
func (s *Store) GetToolFrequency(startTime, endTime string) ([]ToolFreq, error) {
	where := "1=1"
	var args []any

	if startTime != "" {
		where += " AND s.start_time >= ?"
		args = append(args, startTime)
	}
	if endTime != "" {
		where += " AND s.start_time <= ?"
		args = append(args, endTime)
	}

	query := fmt.Sprintf(`
		SELECT tc.tool_name, SUM(tc.call_count) as total
		FROM tool_calls tc
		JOIN sessions s ON tc.session_id = s.session_id
		WHERE %s
		GROUP BY tc.tool_name
		ORDER BY total DESC`, where)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []ToolFreq
	var grandTotal int
	for rows.Next() {
		var tf ToolFreq
		if err := rows.Scan(&tf.Name, &tf.TotalCalls); err == nil {
			grandTotal += tf.TotalCalls
			tools = append(tools, tf)
		}
	}

	// Compute percentages
	for i := range tools {
		if grandTotal > 0 {
			tools[i].Percent = float64(tools[i].TotalCalls) / float64(grandTotal) * 100
		}
	}

	return tools, nil
}

// GetSessionToolCalls returns tool calls for a specific session.
func (s *Store) GetSessionToolCalls(sessionID string) (map[string]int, error) {
	rows, err := s.db.Query(
		"SELECT tool_name, call_count FROM tool_calls WHERE session_id = ? ORDER BY call_count DESC",
		sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tools := make(map[string]int)
	for rows.Next() {
		var name string
		var count int
		if err := rows.Scan(&name, &count); err == nil {
			tools[name] = count
		}
	}
	return tools, nil
}

// GetSessionCount returns the total number of indexed sessions.
func (s *Store) GetSessionCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
	return count, err
}

// GetTotalStats returns aggregate stats from indexed sessions.
type TotalStats struct {
	TotalSessions int
	TotalCost     float64
	TotalTokens   int64
}

func (s *Store) GetTotalStats() (TotalStats, error) {
	var ts TotalStats
	err := s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(cost), 0), COALESCE(SUM(total_tokens), 0)
		FROM sessions`).Scan(&ts.TotalSessions, &ts.TotalCost, &ts.TotalTokens)
	return ts, err
}
