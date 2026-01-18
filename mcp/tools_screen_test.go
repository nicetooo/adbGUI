package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Helper to create a test PNG file (minimal valid PNG)
func createTestPNG(path string) error {
	// Minimal valid 1x1 PNG
	png := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
		0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59,
		0xE7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82, // IEND chunk
	}
	return os.WriteFile(path, png, 0644)
}

// Helper to check if result contains image content
func hasImageContent(result *mcp.CallToolResult) bool {
	for _, content := range result.Content {
		if _, ok := content.(mcp.ImageContent); ok {
			return true
		}
	}
	return false
}

// ==================== screen_screenshot ====================

func TestHandleScreenshot_Success(t *testing.T) {
	// Create a temp PNG file
	tempFile := filepath.Join(os.TempDir(), "test_screenshot.png")
	if err := createTestPNG(tempFile); err != nil {
		t.Fatalf("Failed to create test PNG: %v", err)
	}
	defer os.Remove(tempFile)

	mock := NewMockGazeApp()
	mock.TakeScreenshotResult = tempFile
	server := NewMCPServer(mock)

	result, err := server.handleScreenshot(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain image content (base64)
	if !hasImageContent(result) {
		t.Error("Result should contain image content")
	}

	// Should contain text content
	text := getTextContent(result)
	if !strings.Contains(text, "device1") {
		t.Error("Result should mention device ID")
	}
}

func TestHandleScreenshot_WithUIHierarchy(t *testing.T) {
	// Create a temp PNG file
	tempFile := filepath.Join(os.TempDir(), "test_screenshot_ui.png")
	if err := createTestPNG(tempFile); err != nil {
		t.Fatalf("Failed to create test PNG: %v", err)
	}
	defer os.Remove(tempFile)

	mock := NewMockGazeApp()
	mock.TakeScreenshotResult = tempFile
	mock.GetUIHierarchyResult = &UIHierarchyResult{
		RawXML: `<?xml version="1.0"?><hierarchy><node text="Hello" bounds="[0,0][100,100]"/></hierarchy>`,
	}
	server := NewMCPServer(mock)

	result, err := server.handleScreenshot(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":  "device1",
		"include_ui": true,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain image content
	if !hasImageContent(result) {
		t.Error("Result should contain image content")
	}

	// Should contain UI hierarchy in text
	text := getTextContent(result)
	if !strings.Contains(text, "UI Hierarchy") {
		t.Error("Result should contain UI hierarchy")
	}
	if !strings.Contains(text, "Hello") {
		t.Error("Result should contain element text from hierarchy")
	}
}

func TestHandleScreenshot_AutoGeneratePath(t *testing.T) {
	// Create a temp PNG file
	tempFile := filepath.Join(os.TempDir(), "test_screenshot_auto.png")
	if err := createTestPNG(tempFile); err != nil {
		t.Fatalf("Failed to create test PNG: %v", err)
	}
	defer os.Remove(tempFile)

	mock := NewMockGazeApp()
	mock.TakeScreenshotResult = tempFile
	server := NewMCPServer(mock)

	result, err := server.handleScreenshot(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		// No save_path - should auto-generate
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should still return a result with image
	if result == nil {
		t.Error("Result should not be nil")
	}
	if !hasImageContent(result) {
		t.Error("Result should contain image content")
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
