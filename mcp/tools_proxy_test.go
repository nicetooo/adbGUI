package mcp

import (
	"context"
	"strings"
	"testing"
)

// ==================== proxy_start ====================

func TestHandleProxyStart_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.StartProxyResult = "Proxy started on port 8888"
	server := NewMCPServer(mock)

	result, err := server.handleProxyStart(context.Background(), makeToolRequest(map[string]interface{}{}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "start") {
		t.Error("Result should indicate proxy started")
	}
	if !strings.Contains(text, "8888") {
		t.Error("Result should contain port number")
	}
}

func TestHandleProxyStart_CustomPort(t *testing.T) {
	mock := NewMockGazeApp()
	mock.StartProxyResult = "Proxy started on port 9999"
	server := NewMCPServer(mock)

	_, err := server.handleProxyStart(context.Background(), makeToolRequest(map[string]interface{}{
		"port": float64(9999),
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify port was passed
	lastCall := mock.GetLastCall()
	if lastCall.Args[0] != 9999 {
		t.Errorf("Expected port 9999, got %v", lastCall.Args[0])
	}
}

func TestHandleProxyStart_DefaultPort(t *testing.T) {
	mock := NewMockGazeApp()
	mock.StartProxyResult = "Proxy started"
	server := NewMCPServer(mock)

	_, err := server.handleProxyStart(context.Background(), makeToolRequest(map[string]interface{}{}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Default port should be 8888
	lastCall := mock.GetLastCall()
	if lastCall.Args[0] != 8888 {
		t.Errorf("Expected default port 8888, got %v", lastCall.Args[0])
	}
}

func TestHandleProxyStart_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("StartProxy", ErrProxyAlreadyRunning)
	server := NewMCPServer(mock)

	_, err := server.handleProxyStart(context.Background(), makeToolRequest(map[string]interface{}{}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== proxy_stop ====================

func TestHandleProxyStop_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.StopProxyResult = "Proxy stopped"
	server := NewMCPServer(mock)

	result, err := server.handleProxyStop(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "stop") {
		t.Error("Result should indicate proxy stopped")
	}

	// Verify StopProxy was called
	if !mock.WasMethodCalled("StopProxy") {
		t.Error("StopProxy should have been called")
	}
}

func TestHandleProxyStop_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("StopProxy", ErrProxyNotRunning)
	server := NewMCPServer(mock)

	_, err := server.handleProxyStop(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== proxy_status ====================

func TestHandleProxyStatus_Running(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetProxyStatusResult = true
	server := NewMCPServer(mock)

	result, err := server.handleProxyStatus(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "running") {
		t.Error("Result should indicate proxy is running")
	}
}

func TestHandleProxyStatus_NotRunning(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetProxyStatusResult = false
	server := NewMCPServer(mock)

	result, err := server.handleProxyStatus(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	// Result says "stopped" when proxy is not running
	if !strings.Contains(strings.ToLower(text), "stop") && !strings.Contains(strings.ToLower(text), "not") {
		t.Errorf("Result should indicate proxy is not running, got: %s", text)
	}
}

func TestHandleProxyStatus_CalledCorrectly(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, _ = server.handleProxyStatus(context.Background(), makeToolRequest(nil))

	// Verify GetProxyStatus was called
	if !mock.WasMethodCalled("GetProxyStatus") {
		t.Error("GetProxyStatus should have been called")
	}
}
