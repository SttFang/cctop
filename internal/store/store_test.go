package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSchemaCreation(t *testing.T) {
	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Verify tables exist
	tables := []string{"sessions", "tool_calls", "sync_state"}
	for _, table := range tables {
		var name string
		err := s.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestSchemaIdempotent(t *testing.T) {
	s, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Creating schema again should not error
	_, err = s.db.Exec(schemaSQL)
	if err != nil {
		t.Errorf("schema creation should be idempotent: %v", err)
	}
}

func TestSyncSessions(t *testing.T) {
	s, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Create a temp directory with test data
	tmpDir := t.TempDir()
	projDir := filepath.Join(tmpDir, "-Users-test-MyProject")
	os.MkdirAll(projDir, 0755)

	// Copy test fixture
	srcData, err := os.ReadFile("../../testdata/session-sample.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	testFile := filepath.Join(projDir, "test-session.jsonl")
	os.WriteFile(testFile, srcData, 0644)

	// First sync
	result, err := s.SyncSessions(tmpDir)
	if err != nil {
		t.Fatalf("SyncSessions error: %v", err)
	}
	if result.FilesProcessed != 1 {
		t.Errorf("FilesProcessed = %d, want 1", result.FilesProcessed)
	}

	// Verify session was indexed
	count, err := s.GetSessionCount()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("session count = %d, want 1", count)
	}

	// Second sync (no changes) should skip
	result2, err := s.SyncSessions(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if result2.FilesProcessed != 0 {
		t.Errorf("second sync FilesProcessed = %d, want 0", result2.FilesProcessed)
	}
	if result2.FilesSkipped != 1 {
		t.Errorf("second sync FilesSkipped = %d, want 1", result2.FilesSkipped)
	}
}

func TestSyncSessions_NewFile(t *testing.T) {
	s, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	tmpDir := t.TempDir()
	projDir := filepath.Join(tmpDir, "-Users-test-MyProject")
	os.MkdirAll(projDir, 0755)

	srcData, _ := os.ReadFile("../../testdata/session-sample.jsonl")
	os.WriteFile(filepath.Join(projDir, "session-1.jsonl"), srcData, 0644)

	// First sync
	s.SyncSessions(tmpDir)

	// Add a new file
	os.WriteFile(filepath.Join(projDir, "session-2.jsonl"), srcData, 0644)

	result, err := s.SyncSessions(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if result.FilesProcessed != 1 {
		t.Errorf("FilesProcessed = %d, want 1 (new file)", result.FilesProcessed)
	}

	count, _ := s.GetSessionCount()
	if count != 2 {
		t.Errorf("session count = %d, want 2", count)
	}
}

func TestGetSessions(t *testing.T) {
	s := seedTestStore(t)
	defer s.Close()

	// Default sort by time descending
	sessions, total, err := s.GetSessions(SessionFilter{SortDesc: true, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total < 1 {
		t.Errorf("total = %d, want >= 1", total)
	}
	if len(sessions) < 1 {
		t.Error("should return at least 1 session")
	}

	// Filter by project
	sessions2, _, err := s.GetSessions(SessionFilter{Project: "MyProject", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	for _, sess := range sessions2 {
		if sess.Project != "MyProject" {
			t.Errorf("filtered session project = %q, want MyProject", sess.Project)
		}
	}
}

func TestGetToolFrequency(t *testing.T) {
	s := seedTestStore(t)
	defer s.Close()

	tools, err := s.GetToolFrequency("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) == 0 {
		t.Error("should return tool frequency data")
	}

	// Should be sorted by total calls descending
	for i := 1; i < len(tools); i++ {
		if tools[i].TotalCalls > tools[i-1].TotalCalls {
			t.Error("tools should be sorted by calls descending")
		}
	}
}

// seedTestStore creates a store with test session data.
func seedTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	projDir := filepath.Join(tmpDir, "-Users-test-MyProject")
	os.MkdirAll(projDir, 0755)

	srcData, _ := os.ReadFile("../../testdata/session-sample.jsonl")
	os.WriteFile(filepath.Join(projDir, "test-session.jsonl"), srcData, 0644)

	_, err = s.SyncSessions(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
