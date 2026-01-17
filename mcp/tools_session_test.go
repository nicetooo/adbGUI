package mcp

import (
	"context"
	"strings"
	"testing"
)

// ==================== session_create ====================

func TestHandleSessionCreate_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.CreateSessionResult = "session_123456"
	server := NewMCPServer(mock)

	result, err := server.handleSessionCreate(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"name":      "Test Session",
		"type":      "manual",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "session_123456") {
		t.Error("Result should contain session ID")
	}

	// Verify arguments
	lastCall := mock.GetLastCall()
	if lastCall.Args[0] != "device1" {
		t.Errorf("Expected device_id 'device1', got %v", lastCall.Args[0])
	}
	if lastCall.Args[1] != "manual" {
		t.Errorf("Expected type 'manual', got %v", lastCall.Args[1])
	}
	if lastCall.Args[2] != "Test Session" {
		t.Errorf("Expected name 'Test Session', got %v", lastCall.Args[2])
	}
}

func TestHandleSessionCreate_DefaultType(t *testing.T) {
	mock := NewMockGazeApp()
	mock.CreateSessionResult = "session_123"
	server := NewMCPServer(mock)

	_, err := server.handleSessionCreate(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Default type should be "manual"
	lastCall := mock.GetLastCall()
	if lastCall.Args[1] != "manual" {
		t.Errorf("Expected default type 'manual', got %v", lastCall.Args[1])
	}
}

func TestHandleSessionCreate_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleSessionCreate(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

// ==================== session_end ====================

func TestHandleSessionEnd_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleSessionEnd(context.Background(), makeToolRequest(map[string]interface{}{
		"session_id": "session_123",
		"status":     "completed",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "end") || !strings.Contains(strings.ToLower(text), "session") {
		t.Error("Result should mention ending session")
	}

	// Verify arguments
	lastCall := mock.GetLastCall()
	if lastCall.Args[0] != "session_123" {
		t.Errorf("Expected session_id 'session_123', got %v", lastCall.Args[0])
	}
	if lastCall.Args[1] != "completed" {
		t.Errorf("Expected status 'completed', got %v", lastCall.Args[1])
	}
}

func TestHandleSessionEnd_DefaultStatus(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleSessionEnd(context.Background(), makeToolRequest(map[string]interface{}{
		"session_id": "session_123",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Default status should be "completed"
	lastCall := mock.GetLastCall()
	if lastCall.Args[1] != "completed" {
		t.Errorf("Expected default status 'completed', got %v", lastCall.Args[1])
	}
}

func TestHandleSessionEnd_MissingSessionId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleSessionEnd(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing session_id")
	}
}

func TestHandleSessionEnd_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("EndSession", ErrSessionNotFound)
	server := NewMCPServer(mock)

	_, err := server.handleSessionEnd(context.Background(), makeToolRequest(map[string]interface{}{
		"session_id": "nonexistent",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== session_active ====================

func TestHandleSessionActive_HasActive(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetActiveSessionResult = "session_active_123"
	server := NewMCPServer(mock)

	result, err := server.handleSessionActive(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "session_active_123") {
		t.Error("Result should contain active session ID")
	}
}

func TestHandleSessionActive_NoActive(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetActiveSessionResult = ""
	server := NewMCPServer(mock)

	result, err := server.handleSessionActive(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "no") && !strings.Contains(strings.ToLower(text), "active") {
		t.Errorf("Result should indicate no active session, got: %s", text)
	}
}

func TestHandleSessionActive_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleSessionActive(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

// ==================== session_list ====================

func TestHandleSessionList_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ListStoredSessionsResult = []DeviceSession{
		SampleSession("sess1", "device1"),
		SampleSession("sess2", "device1"),
	}
	server := NewMCPServer(mock)

	result, err := server.handleSessionList(context.Background(), makeToolRequest(map[string]interface{}{}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "sess1") {
		t.Error("Result should contain session 1")
	}
	if !strings.Contains(text, "sess2") {
		t.Error("Result should contain session 2")
	}
}

func TestHandleSessionList_WithDeviceFilter(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ListStoredSessionsResult = []DeviceSession{}
	server := NewMCPServer(mock)

	_, err := server.handleSessionList(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify device_id was passed
	lastCall := mock.GetLastCall()
	if lastCall.Args[0] != "device1" {
		t.Errorf("Expected device_id 'device1', got %v", lastCall.Args[0])
	}
}

func TestHandleSessionList_WithLimit(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ListStoredSessionsResult = []DeviceSession{}
	server := NewMCPServer(mock)

	_, err := server.handleSessionList(context.Background(), makeToolRequest(map[string]interface{}{
		"limit": float64(10),
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify limit was passed
	lastCall := mock.GetLastCall()
	if lastCall.Args[1] != 10 {
		t.Errorf("Expected limit 10, got %v", lastCall.Args[1])
	}
}

func TestHandleSessionList_Empty(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ListStoredSessionsResult = []DeviceSession{}
	server := NewMCPServer(mock)

	result, err := server.handleSessionList(context.Background(), makeToolRequest(map[string]interface{}{}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "no") {
		t.Errorf("Result should indicate no sessions, got: %s", text)
	}
}

func TestHandleSessionList_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("ListStoredSessions", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleSessionList(context.Background(), makeToolRequest(map[string]interface{}{}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== session_events ====================

func TestHandleSessionEvents_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.QuerySessionEventsResult = &EventQueryResult{
		Events: []interface{}{
			map[string]interface{}{"type": "touch", "title": "Tap at 100,200"},
			map[string]interface{}{"type": "logcat", "title": "App started"},
		},
		Total:   2,
		HasMore: false,
	}
	server := NewMCPServer(mock)

	result, err := server.handleSessionEvents(context.Background(), makeToolRequest(map[string]interface{}{
		"session_id": "session_123",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "2") {
		t.Error("Result should contain event count")
	}
}

func TestHandleSessionEvents_WithFilters(t *testing.T) {
	mock := NewMockGazeApp()
	mock.QuerySessionEventsResult = &EventQueryResult{
		Events:  []interface{}{},
		Total:   0,
		HasMore: false,
	}
	server := NewMCPServer(mock)

	_, err := server.handleSessionEvents(context.Background(), makeToolRequest(map[string]interface{}{
		"session_id": "session_123",
		"types":      "touch,logcat",
		"sources":    "device,app",
		"levels":     "info,error",
		"search":     "test",
		"limit":      float64(50),
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify query was constructed properly
	lastCall := mock.GetLastCall()
	query, ok := lastCall.Args[0].(EventQuery)
	if !ok {
		t.Fatal("Expected EventQuery argument")
	}
	if query.SessionID != "session_123" {
		t.Errorf("Expected session_id 'session_123', got %s", query.SessionID)
	}
	if query.SearchText != "test" {
		t.Errorf("Expected search 'test', got %s", query.SearchText)
	}
	if query.Limit != 50 {
		t.Errorf("Expected limit 50, got %d", query.Limit)
	}
}

func TestHandleSessionEvents_MissingSessionId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleSessionEvents(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing session_id")
	}
}

func TestHandleSessionEvents_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("QuerySessionEvents", ErrSessionNotFound)
	server := NewMCPServer(mock)

	_, err := server.handleSessionEvents(context.Background(), makeToolRequest(map[string]interface{}{
		"session_id": "nonexistent",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== session_stats ====================

func TestHandleSessionStats_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionStatsResult = map[string]interface{}{
		"totalEvents": 150,
		"duration":    60000,
		"eventTypes": map[string]int{
			"touch":  50,
			"logcat": 100,
		},
	}
	server := NewMCPServer(mock)

	result, err := server.handleSessionStats(context.Background(), makeToolRequest(map[string]interface{}{
		"session_id": "session_123",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "150") || !strings.Contains(text, "event") {
		t.Error("Result should contain stats")
	}
}

func TestHandleSessionStats_MissingSessionId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleSessionStats(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing session_id")
	}
}

func TestHandleSessionStats_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("GetSessionStats", ErrSessionNotFound)
	server := NewMCPServer(mock)

	_, err := server.handleSessionStats(context.Background(), makeToolRequest(map[string]interface{}{
		"session_id": "nonexistent",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
