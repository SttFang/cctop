package data

import (
	"testing"
)

func TestParseSessionFile(t *testing.T) {
	summary, err := ParseSessionFile("../../testdata/session-sample.jsonl")
	if err != nil {
		t.Fatalf("ParseSessionFile error: %v", err)
	}

	if summary.SessionID != "session-sample" {
		t.Errorf("SessionID = %q, want %q", summary.SessionID, "session-sample")
	}

	// 3 user messages + 3 assistant messages = but we only count user+assistant with msg content
	// Actually: 2 user + 3 assistant = 5 messages
	if summary.MessageCount != 5 {
		t.Errorf("MessageCount = %d, want 5", summary.MessageCount)
	}

	if summary.Model != "claude-opus-4-6" {
		t.Errorf("Model = %q, want %q", summary.Model, "claude-opus-4-6")
	}

	// Token totals: sum of 3 assistant messages
	// Input: 1000+800+600 = 2400
	if summary.InputTokens != 2400 {
		t.Errorf("InputTokens = %d, want 2400", summary.InputTokens)
	}
	// Output: 500+1200+300 = 2000
	if summary.OutputTokens != 2000 {
		t.Errorf("OutputTokens = %d, want 2000", summary.OutputTokens)
	}
	// CacheRead: 5000+4000+3000 = 12000
	if summary.CacheRead != 12000 {
		t.Errorf("CacheRead = %d, want 12000", summary.CacheRead)
	}

	// First prompt
	if summary.FirstPrompt != "帮我重构认证模块" {
		t.Errorf("FirstPrompt = %q, want %q", summary.FirstPrompt, "帮我重构认证模块")
	}

	// Tool calls: Read(1) + Edit(2) + Bash(1) = 4 total
	if summary.ToolCalls["Read"] != 1 {
		t.Errorf("ToolCalls[Read] = %d, want 1", summary.ToolCalls["Read"])
	}
	if summary.ToolCalls["Edit"] != 2 {
		t.Errorf("ToolCalls[Edit] = %d, want 2", summary.ToolCalls["Edit"])
	}
	if summary.ToolCalls["Bash"] != 1 {
		t.Errorf("ToolCalls[Bash] = %d, want 1", summary.ToolCalls["Bash"])
	}

	// Duration: 14:22:00 to 14:27:00 = 5 minutes = 300000ms
	if summary.DurationMs != 300000 {
		t.Errorf("DurationMs = %d, want 300000", summary.DurationMs)
	}

	// Cost should be computed
	if summary.Cost <= 0 {
		t.Errorf("Cost = %f, should be > 0", summary.Cost)
	}
}

func TestParseSessionFile_Empty(t *testing.T) {
	summary, err := ParseSessionFile("../../testdata/session-empty.jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.MessageCount != 0 {
		t.Errorf("MessageCount = %d, want 0", summary.MessageCount)
	}
	if summary.TotalTokens != 0 {
		t.Errorf("TotalTokens = %d, want 0", summary.TotalTokens)
	}
}

func TestParseSessionFile_Malformed(t *testing.T) {
	summary, err := ParseSessionFile("../../testdata/session-malformed.jsonl")
	if err != nil {
		t.Fatalf("should not error on malformed lines, got: %v", err)
	}
	// Malformed lines should be skipped
	if summary.MessageCount != 0 {
		t.Errorf("MessageCount = %d, want 0 for malformed file", summary.MessageCount)
	}
}

func TestExtractProjectName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"-Users-fanghanjun-ClawUI", "ClawUI"},
		{"-Users-fanghanjun-dev-my-app", "dev-my-app"},
		{"-Users-fanghanjun", "fanghanjun"},
		{"-tmp", "tmp"},
		{"simple", "simple"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExtractProjectName(tt.input)
			if got != tt.want {
				t.Errorf("ExtractProjectName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
