package data

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fanghanjun/cctop/internal/config"
)

// ParseSessionFile reads a JSONL session file and extracts a SessionSummary.
func ParseSessionFile(path string) (*SessionSummary, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening session file: %w", err)
	}
	defer f.Close()

	summary := &SessionSummary{
		SessionID: strings.TrimSuffix(filepath.Base(path), ".jsonl"),
		ToolCalls: make(map[string]int),
	}

	var firstTime, lastTime time.Time
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		var msg SessionMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}

		// Skip non-message types
		if msg.Type != "user" && msg.Type != "assistant" {
			continue
		}

		if msg.SessionID != "" && summary.SessionID == "" {
			summary.SessionID = msg.SessionID
		}

		// Parse timestamp
		if msg.Timestamp != "" {
			t, err := time.Parse(time.RFC3339Nano, msg.Timestamp)
			if err == nil {
				if firstTime.IsZero() || t.Before(firstTime) {
					firstTime = t
				}
				if t.After(lastTime) {
					lastTime = t
				}
			}
		}

		if msg.Message == nil {
			continue
		}

		summary.MessageCount++

		// Extract first user prompt
		if msg.Message.Role == "user" && summary.FirstPrompt == "" {
			summary.FirstPrompt = extractTextContent(msg.Message.Content)
			if len(summary.FirstPrompt) > 200 {
				summary.FirstPrompt = summary.FirstPrompt[:200] + "..."
			}
		}

		// Extract model and token usage from assistant messages
		if msg.Message.Role == "assistant" {
			if msg.Message.Model != "" {
				summary.Model = msg.Message.Model
			}
			if msg.Message.Usage != nil {
				u := msg.Message.Usage
				summary.InputTokens += u.InputTokens
				summary.OutputTokens += u.OutputTokens
				summary.CacheRead += u.CacheReadInputTokens
				summary.CacheCreate += u.CacheCreationInputTokens
			}

			// Extract tool use from content
			extractToolCalls(msg.Message.Content, summary.ToolCalls)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning session file: %w", err)
	}

	// Compute derived fields
	if !firstTime.IsZero() {
		summary.StartTime = firstTime.Format(time.RFC3339)
	}
	if !firstTime.IsZero() && !lastTime.IsZero() {
		summary.DurationMs = lastTime.Sub(firstTime).Milliseconds()
	}
	summary.TotalTokens = summary.InputTokens + summary.OutputTokens + summary.CacheRead + summary.CacheCreate
	summary.Cost = config.ComputeCost(summary.Model, summary.InputTokens, summary.OutputTokens, summary.CacheRead, summary.CacheCreate)

	// Extract project from directory path
	dir := filepath.Dir(path)
	summary.Project = ExtractProjectName(filepath.Base(dir))

	return summary, nil
}

// ExtractProjectName extracts a readable project name from the directory name.
// e.g., "-Users-fanghanjun-ClawUI" -> "ClawUI"
func ExtractProjectName(dirName string) string {
	parts := strings.Split(dirName, "-")
	// Find the last meaningful segment(s)
	// Pattern: -Users-username-ProjectName or -Users-username-path-to-project
	userIdx := -1
	for i, p := range parts {
		if p == "Users" || p == "home" {
			userIdx = i
			break
		}
	}
	if userIdx >= 0 && userIdx+2 < len(parts) {
		// Skip Users and username, take the rest
		remaining := parts[userIdx+2:]
		return strings.Join(remaining, "-")
	}
	// Fallback: return last non-empty part
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return dirName
}

// extractTextContent gets text from various content formats.
func extractTextContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if m["type"] == "text" {
					if text, ok := m["text"].(string); ok {
						return text
					}
				}
			}
		}
	}
	return ""
}

// extractToolCalls counts tool_use blocks in assistant message content.
func extractToolCalls(content interface{}, toolCalls map[string]int) {
	items, ok := content.([]interface{})
	if !ok {
		return
	}
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if m["type"] == "tool_use" {
			if name, ok := m["name"].(string); ok {
				toolCalls[name]++
			}
		}
	}
}
