package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestLoggerInit(t *testing.T) {
	// Test default config
	config := DefaultLogConfig()
	if config.Level != LogLevelInfo {
		t.Errorf("Expected default level Info, got %d", config.Level)
	}
	if !config.Console {
		t.Error("Expected console output to be enabled by default")
	}
	if config.File {
		t.Error("Expected file output to be disabled by default")
	}
}

func TestLoggerLevels(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a test logger
	testLogger := zerolog.New(&buf).With().Timestamp().Logger()

	// Test each level
	testLogger.Debug().Msg("debug message")
	testLogger.Info().Msg("info message")
	testLogger.Warn().Msg("warn message")
	testLogger.Error().Msg("error message")

	output := buf.String()

	// Verify messages were logged
	if !strings.Contains(output, "debug message") {
		t.Error("Expected debug message in output")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Expected info message in output")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Expected warn message in output")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Expected error message in output")
	}
}

func TestLoggerStructuredFields(t *testing.T) {
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf).With().Logger()

	// Log with structured fields
	testLogger.Info().
		Str("module", "test").
		Str("deviceId", "device-123").
		Int("eventCount", 42).
		Msg("test message")

	output := buf.String()

	// Verify fields are in JSON output
	if !strings.Contains(output, "module") {
		t.Error("Expected 'module' field in output")
	}
	if !strings.Contains(output, "test") {
		t.Error("Expected 'test' value in output")
	}
	if !strings.Contains(output, "deviceId") {
		t.Error("Expected 'deviceId' field in output")
	}
	if !strings.Contains(output, "device-123") {
		t.Error("Expected 'device-123' value in output")
	}
	if !strings.Contains(output, "eventCount") {
		t.Error("Expected 'eventCount' field in output")
	}
	if !strings.Contains(output, "42") {
		t.Error("Expected '42' value in output")
	}
}

func TestLogFunctions(t *testing.T) {
	// Test that log functions don't panic
	// These use the global Logger

	// Reinit with a test config
	err := InitLogger(DefaultLogConfig())
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// Test convenience functions
	LogDebug("test").Msg("debug test")
	LogInfo("test").Msg("info test")
	LogWarn("test").Msg("warn test")
	LogError("test").Msg("error test")

	// Test module-specific functions
	SessionLog().Msg("session test")
}

func TestLogConfigLevels(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected zerolog.Level
	}{
		{LogLevelDebug, zerolog.DebugLevel},
		{LogLevelInfo, zerolog.InfoLevel},
		{LogLevelWarn, zerolog.WarnLevel},
		{LogLevelError, zerolog.ErrorLevel},
	}

	for _, tt := range tests {
		config := LogConfig{
			Level:   tt.level,
			Console: true,
		}
		err := InitLogger(config)
		if err != nil {
			t.Errorf("Failed to init logger with level %d: %v", tt.level, err)
		}
	}
}

func TestPersistentLogConfig(t *testing.T) {
	tempDir := t.TempDir()
	config := PersistentLogConfig(tempDir)

	if !config.File {
		t.Error("Expected File to be enabled")
	}
	if !config.Console {
		t.Error("Expected Console to be enabled")
	}
	if config.MaxSizeMB != 10 {
		t.Errorf("Expected MaxSizeMB 10, got %d", config.MaxSizeMB)
	}
	if config.MaxAgeDays != 7 {
		t.Errorf("Expected MaxAgeDays 7, got %d", config.MaxAgeDays)
	}
	if config.MaxBackups != 5 {
		t.Errorf("Expected MaxBackups 5, got %d", config.MaxBackups)
	}
	expectedPath := filepath.Join(tempDir, "logs", "gaze.log")
	if config.FilePath != expectedPath {
		t.Errorf("Expected FilePath %s, got %s", expectedPath, config.FilePath)
	}
}

func TestPersistentLogger(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	config := LogConfig{
		Level:      LogLevelInfo,
		Console:    false,
		File:       true,
		FilePath:   logPath,
		MaxSizeMB:  1,
		MaxAgeDays: 7,
		MaxBackups: 5,
	}

	pl, err := NewPersistentLogger(config)
	if err != nil {
		t.Fatalf("Failed to create persistent logger: %v", err)
	}
	defer pl.Close()

	// Write some data
	testData := []byte("Test log message\n")
	n, err := pl.Write(testData)
	if err != nil {
		t.Errorf("Failed to write: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
	}

	// Verify file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Read file contents
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if !strings.Contains(string(content), "Test log message") {
		t.Error("Log file does not contain expected message")
	}
}

func TestUserActionLog(t *testing.T) {
	// Initialize logger
	err := InitLogger(DefaultLogConfig())
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// Test logging user actions (should not panic)
	LogUserAction(ActionDeviceConnect, "device-123", map[string]interface{}{
		"address": "192.168.1.100:5555",
		"success": true,
	})

	LogUserAction(ActionProxyStart, "", map[string]interface{}{
		"port": 8888,
	})

	LogUserAction(ActionScrcpyStart, "device-456", map[string]interface{}{
		"max_fps":  60,
		"bit_rate": 8,
	})
}

func TestAppStateLog(t *testing.T) {
	err := InitLogger(DefaultLogConfig())
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// Test app state logging (should not panic)
	LogAppState(StateStarting, map[string]interface{}{
		"version": "1.0.0",
	})

	LogAppState(StateReady, nil)

	LogAppState(StateShuttingDown, map[string]interface{}{
		"reason": "user_request",
	})
}

func TestOperationTimer(t *testing.T) {
	err := InitLogger(DefaultLogConfig())
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// Test successful operation
	timer := StartOperation("test_module", "test_operation")
	timer.AddDetail("key1", "value1")
	timer.AddDetail("key2", 123)
	time.Sleep(10 * time.Millisecond)
	timer.End()

	// Test failed operation
	timer2 := StartOperation("test_module", "failing_operation")
	time.Sleep(5 * time.Millisecond)
	timer2.EndWithError(os.ErrNotExist)
}

func TestLogRotation(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "rotate_test.log")

	// Create a config with very small max size to trigger rotation
	config := LogConfig{
		Level:      LogLevelInfo,
		Console:    false,
		File:       true,
		FilePath:   logPath,
		MaxSizeMB:  0, // Will be set to trigger rotation on small writes
		MaxAgeDays: 7,
		MaxBackups: 5,
		Compress:   false, // Disable compression for easier testing
	}

	pl, err := NewPersistentLogger(config)
	if err != nil {
		t.Fatalf("Failed to create persistent logger: %v", err)
	}
	defer pl.Close()

	// Write initial data
	_, err = pl.Write([]byte("Initial log entry\n"))
	if err != nil {
		t.Errorf("Failed to write initial entry: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestCloseLogger(t *testing.T) {
	tempDir := t.TempDir()
	config := PersistentLogConfig(tempDir)

	err := InitLogger(config)
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// Write something
	LogInfo("test").Msg("test message before close")

	// Close should not panic
	CloseLogger()

	// Reinit with console only for subsequent tests
	_ = InitLogger(DefaultLogConfig())
}
