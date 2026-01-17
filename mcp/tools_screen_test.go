package mcp

import (
	"context"
	"strings"
	"testing"
)

// ==================== screen_screenshot ====================

func TestHandleScreenshot_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.TakeScreenshotResult = "/tmp/screenshot.png"
	server := NewMCPServer(mock)

	result, err := server.handleScreenshot(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"save_path": "/tmp/screenshot.png",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "/tmp/screenshot.png") {
		t.Error("Result should contain save path")
	}
}

func TestHandleScreenshot_AutoGeneratePath(t *testing.T) {
	mock := NewMockGazeApp()
	mock.TakeScreenshotResult = "/Users/test/Downloads/screenshot_device1_123456.png"
	server := NewMCPServer(mock)

	result, err := server.handleScreenshot(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		// No save_path - should auto-generate
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should still return a result with auto-generated path
	if result == nil {
		t.Error("Result should not be nil")
	}

	// Verify TakeScreenshot was called
	if !mock.WasMethodCalled("TakeScreenshot") {
		t.Error("TakeScreenshot should have been called")
	}
}

func TestHandleScreenshot_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleScreenshot(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleScreenshot_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("TakeScreenshot", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleScreenshot(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== screen_record_start ====================

func TestHandleRecordStart_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleRecordStart(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "record") {
		t.Error("Result should mention recording")
	}

	// Verify StartRecording was called
	if !mock.WasMethodCalled("StartRecording") {
		t.Error("StartRecording should have been called")
	}
}

func TestHandleRecordStart_WithOptions(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleRecordStart(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"max_size":  float64(720),
		"bit_rate":  float64(4),
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify StartRecording was called with config
	lastCall := mock.GetLastCall()
	if lastCall.Method != "StartRecording" {
		t.Errorf("Expected StartRecording call, got %s", lastCall.Method)
	}
	config, ok := lastCall.Args[1].(ScrcpyConfig)
	if !ok {
		t.Fatal("Second argument should be ScrcpyConfig")
	}
	if config.MaxSize != 720 {
		t.Errorf("Expected MaxSize 720, got %d", config.MaxSize)
	}
}

func TestHandleRecordStart_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleRecordStart(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleRecordStart_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("StartRecording", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleRecordStart(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== screen_record_stop ====================

func TestHandleRecordStop_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleRecordStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "stop") {
		t.Error("Result should mention stopping")
	}

	// Verify StopRecording was called
	if !mock.WasMethodCalled("StopRecording") {
		t.Error("StopRecording should have been called")
	}
}

func TestHandleRecordStop_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleRecordStop(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleRecordStop_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("StopRecording", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleRecordStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== screen_recording_status ====================

func TestHandleRecordingStatus_Recording(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsRecordingResult = true
	server := NewMCPServer(mock)

	result, err := server.handleRecordingStatus(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "recording") {
		t.Error("Result should indicate recording status")
	}
}

func TestHandleRecordingStatus_NotRecording(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsRecordingResult = false
	server := NewMCPServer(mock)

	result, err := server.handleRecordingStatus(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "not") {
		t.Error("Result should indicate not recording")
	}
}

func TestHandleRecordingStatus_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleRecordingStatus(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}
