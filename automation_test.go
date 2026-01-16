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
