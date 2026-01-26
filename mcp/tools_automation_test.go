package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// ==================== ui_hierarchy ====================

func TestHandleUIHierarchy_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetUIHierarchyResult = &UIHierarchyResult{
		Root: map[string]interface{}{
			"class":      "android.widget.FrameLayout",
			"text":       "",
			"resourceId": "",
			"children": []interface{}{
				map[string]interface{}{
					"class": "android.widget.Button",
					"text":  "Click Me",
				},
			},
		},
		RawXML: "<hierarchy></hierarchy>",
	}
	server := NewMCPServer(mock)

	result, err := server.handleUIHierarchy(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "FrameLayout") && !strings.Contains(text, "Button") {
		t.Error("Result should contain UI element information")
	}
}

func TestHandleUIHierarchy_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUIHierarchy(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleUIHierarchy_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("GetUIHierarchy", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleUIHierarchy(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestHandleUIHierarchy_NilResult(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetUIHierarchyResult = nil
	server := NewMCPServer(mock)

	_, err := server.handleUIHierarchy(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	// Should handle nil result gracefully
	if err == nil {
		// If no error, result should still be valid
		t.Log("Nil result handled without error")
	}
}

// ==================== ui_search ====================

func TestHandleUISearch_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SearchUIElementsResult = []map[string]interface{}{
		{
			"text":       "Login",
			"class":      "android.widget.Button",
			"bounds":     "[100,200][300,250]",
			"resourceId": "com.example:id/login_btn",
		},
		{
			"text":   "Username",
			"class":  "android.widget.EditText",
			"bounds": "[100,100][300,150]",
		},
	}
	server := NewMCPServer(mock)

	result, err := server.handleUISearch(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"query":     "Login",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "Login") {
		t.Error("Result should contain matched elements")
	}
	if !strings.Contains(text, "2") || !strings.Contains(strings.ToLower(text), "element") {
		t.Error("Result should mention number of elements found")
	}
}

func TestHandleUISearch_NoResults(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SearchUIElementsResult = []map[string]interface{}{}
	server := NewMCPServer(mock)

	result, err := server.handleUISearch(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"query":     "NonExistent",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "no") {
		t.Errorf("Result should indicate no elements found, got: %s", text)
	}
}

func TestHandleUISearch_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUISearch(context.Background(), makeToolRequest(map[string]interface{}{
		"query": "Login",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleUISearch_MissingQuery(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUISearch(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error for missing query")
	}
}

func TestHandleUISearch_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("SearchUIElements", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleUISearch(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"query":     "Login",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== ui_tap ====================

func TestHandleUITap_SuccessWithBounds(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleUITap(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"bounds":    "[100,200][300,250]",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "tap") {
		t.Error("Result should mention tap")
	}

	// Verify PerformNodeAction was called with correct arguments
	lastCall := mock.GetLastCall()
	if lastCall.Method != "PerformNodeAction" {
		t.Errorf("Expected PerformNodeAction call, got %s", lastCall.Method)
	}
	if lastCall.Args[1] != "[100,200][300,250]" {
		t.Errorf("Expected bounds '[100,200][300,250]', got %v", lastCall.Args[1])
	}
	if lastCall.Args[2] != "click" {
		t.Errorf("Expected action 'click', got %v", lastCall.Args[2])
	}
}

func TestHandleUITap_SuccessWithCoordinates(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleUITap(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"bounds":    "500,600",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Result should not be nil")
	}

	// Verify coordinates were passed
	lastCall := mock.GetLastCall()
	if lastCall.Args[1] != "500,600" {
		t.Errorf("Expected bounds '500,600', got %v", lastCall.Args[1])
	}
}

func TestHandleUITap_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUITap(context.Background(), makeToolRequest(map[string]interface{}{
		"bounds": "[100,200][300,250]",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleUITap_MissingBounds(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUITap(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error for missing bounds")
	}
}

func TestHandleUITap_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("PerformNodeAction", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleUITap(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"bounds":    "[100,200][300,250]",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== ui_swipe ====================

func TestHandleUISwipe_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleUISwipe(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"bounds":    "[100,200][300,800]",
		"direction": "up",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "swipe") {
		t.Error("Result should mention swipe")
	}

	// Verify action type
	lastCall := mock.GetLastCall()
	if lastCall.Args[2] != "swipe_up" {
		t.Errorf("Expected action 'swipe_up', got %v", lastCall.Args[2])
	}
}

func TestHandleUISwipe_DefaultDirection(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUISwipe(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"bounds":    "[100,200][300,800]",
		// No direction - should default to "down"
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	lastCall := mock.GetLastCall()
	if lastCall.Args[2] != "swipe_down" {
		t.Errorf("Expected default action 'swipe_down', got %v", lastCall.Args[2])
	}
}

func TestHandleUISwipe_AllDirections(t *testing.T) {
	directions := []string{"up", "down", "left", "right"}

	for _, dir := range directions {
		mock := NewMockGazeApp()
		server := NewMCPServer(mock)

		_, err := server.handleUISwipe(context.Background(), makeToolRequest(map[string]interface{}{
			"device_id": "device1",
			"bounds":    "[0,0][1080,1920]",
			"direction": dir,
		}))
		if err != nil {
			t.Errorf("Unexpected error for direction %s: %v", dir, err)
		}

		lastCall := mock.GetLastCall()
		expectedAction := "swipe_" + dir
		if lastCall.Args[2] != expectedAction {
			t.Errorf("Expected action '%s', got %v", expectedAction, lastCall.Args[2])
		}
	}
}

func TestHandleUISwipe_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUISwipe(context.Background(), makeToolRequest(map[string]interface{}{
		"bounds": "[100,200][300,800]",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleUISwipe_MissingBounds(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUISwipe(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error for missing bounds")
	}
}

// ==================== ui_input ====================

func TestHandleUIInput_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleUIInput(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"text":      "Hello World",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "input") {
		t.Error("Result should mention input")
	}

	// Verify InputText was called with correct args
	if !mock.WasMethodCalled("InputText") {
		t.Error("InputText should have been called")
	}
	lastCall := mock.GetLastCallByMethod("InputText")
	if lastCall.Args[0] != "device1" {
		t.Errorf("Expected device 'device1', got %v", lastCall.Args[0])
	}
	if lastCall.Args[1] != "Hello World" {
		t.Errorf("Expected text 'Hello World', got %v", lastCall.Args[1])
	}
}

func TestHandleUIInput_SpecialCharacters(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUIInput(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"text":      "test@email.com",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	lastCall := mock.GetLastCallByMethod("InputText")
	if lastCall == nil {
		t.Fatal("InputText should have been called")
	}
	if lastCall.Args[1] != "test@email.com" {
		t.Errorf("Expected text 'test@email.com', got %v", lastCall.Args[1])
	}
}

func TestHandleUIInput_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUIInput(context.Background(), makeToolRequest(map[string]interface{}{
		"text": "Hello",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleUIInput_MissingText(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUIInput(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error for missing text")
	}
}

func TestHandleUIInput_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.InputTextError = ErrDeviceOffline
	server := NewMCPServer(mock)

	result, err := server.handleUIInput(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"text":      "Hello",
	}))
	if err != nil {
		t.Fatalf("Unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected IsError=true in result")
	}
	text := getTextContent(result)
	if !strings.Contains(text, "device offline") {
		t.Errorf("Error message should contain 'device offline', got: %s", text)
	}
}

// ==================== ui_resolution ====================

func TestHandleUIResolution_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetDeviceResolutionResult = "1080x2400"
	server := NewMCPServer(mock)

	result, err := server.handleUIResolution(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "1080") || !strings.Contains(text, "2400") {
		t.Error("Result should contain resolution")
	}
}

func TestHandleUIResolution_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleUIResolution(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleUIResolution_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("GetDeviceResolution", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleUIResolution(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// Suppress unused import warning
var _ = json.Marshal
