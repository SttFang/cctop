package data

// StatsCache represents the parsed ~/.claude/stats-cache.json structure.
type StatsCache struct {
	Version          int                `json:"version"`
	LastComputedDate string             `json:"lastComputedDate"`
	DailyActivity    []DailyActivity    `json:"dailyActivity"`
	DailyModelTokens []DailyModelTokens `json:"dailyModelTokens"`
	ModelUsage       map[string]ModelUsageEntry `json:"modelUsage"`
	TotalSessions    int                `json:"totalSessions"`
	TotalMessages    int                `json:"totalMessages"`
	LongestSession   LongestSession     `json:"longestSession"`
	FirstSessionDate string             `json:"firstSessionDate"`
	HourCounts       map[string]int     `json:"hourCounts"`
}

type DailyActivity struct {
	Date          string `json:"date"`
	MessageCount  int    `json:"messageCount"`
	SessionCount  int    `json:"sessionCount"`
	ToolCallCount int    `json:"toolCallCount"`
}

type DailyModelTokens struct {
	Date          string           `json:"date"`
	TokensByModel map[string]int64 `json:"tokensByModel"`
}

type ModelUsageEntry struct {
	InputTokens              int64   `json:"inputTokens"`
	OutputTokens             int64   `json:"outputTokens"`
	CacheReadInputTokens     int64   `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int64   `json:"cacheCreationInputTokens"`
	WebSearchRequests        int     `json:"webSearchRequests"`
	CostUSD                  float64 `json:"costUSD"`
}

type LongestSession struct {
	SessionID    string `json:"sessionId"`
	Duration     int64  `json:"duration"`
	MessageCount int    `json:"messageCount"`
	Timestamp    string `json:"timestamp"`
}

// SessionMessage represents a single message from a JSONL session file.
type SessionMessage struct {
	Type      string           `json:"type"`
	SessionID string           `json:"sessionId"`
	UUID      string           `json:"uuid"`
	Timestamp string           `json:"timestamp"`
	CWD       string           `json:"cwd"`
	Message   *MessageContent  `json:"message,omitempty"`
}

type MessageContent struct {
	Role      string      `json:"role"`
	Content   interface{} `json:"content"`
	ID        string      `json:"id,omitempty"`
	Model     string      `json:"model,omitempty"`
	Usage     *TokenUsage `json:"usage,omitempty"`
}

type TokenUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
}

// SessionSummary is a processed summary of a session for display.
type SessionSummary struct {
	SessionID    string
	Project      string
	Model        string
	StartTime    string
	DurationMs   int64
	InputTokens  int64
	OutputTokens int64
	CacheRead    int64
	CacheCreate  int64
	TotalTokens  int64
	Cost         float64
	MessageCount int
	FirstPrompt  string
	ToolCalls    map[string]int
}
