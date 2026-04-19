package jsonlog

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	if logger.out != &buf {
		t.Fatal("expected output destination to be set")
	}
	if logger.minLevel != LevelInfo {
		t.Fatalf("expected minLevel to be LevelInfo, got %v", logger.minLevel)
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelInfo, "INFO"},
		{LevelError, "ERROR"},
		{LevelFatal, "FATAL"},
		{LevelOff, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.expected)
			}
		})
	}
}

func TestPrintInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	logger.PrintInfo("test message", map[string]string{"key": "value"})

	var entry struct {
		Level      string            `json:"level"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties"`
		Trace      string            `json:"trace"`
	}

	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if entry.Level != "INFO" {
		t.Errorf("expected level INFO, got %q", entry.Level)
	}
	if entry.Message != "test message" {
		t.Errorf("expected message 'test message', got %q", entry.Message)
	}
	if entry.Properties["key"] != "value" {
		t.Errorf("expected property key=value, got key=%q", entry.Properties["key"])
	}
	if entry.Trace != "" {
		t.Error("expected no trace for INFO level")
	}
}

func TestPrintErrorIncludesTrace(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	logger.PrintError(errForTest("something went wrong"), nil)

	var entry struct {
		Level   string `json:"level"`
		Message string `json:"message"`
		Trace   string `json:"trace"`
	}

	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if entry.Level != "ERROR" {
		t.Errorf("expected level ERROR, got %q", entry.Level)
	}
	if entry.Trace == "" {
		t.Error("expected stack trace for ERROR level, got empty")
	}
}

func TestMinLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelError)

	logger.PrintInfo("should be filtered out", nil)

	if buf.Len() != 0 {
		t.Error("expected INFO message to be filtered out when minLevel is ERROR")
	}

	logger.PrintError(errForTest("should appear"), nil)

	if buf.Len() == 0 {
		t.Error("expected ERROR message to appear when minLevel is ERROR")
	}
}

func TestWriteMethod(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	n, err := logger.Write([]byte("test via Write"))
	if err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}
	if n == 0 {
		t.Error("Write() returned 0 bytes written")
	}

	var entry struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if entry.Level != "ERROR" {
		t.Errorf("expected Write() to log at ERROR level, got %q", entry.Level)
	}
}

// errForTest is a simple error implementation for testing.
type errForTest string

func (e errForTest) Error() string {
	return string(e)
}
