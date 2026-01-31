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

	// Default port should be 8080
	lastCall := mock.GetLastCall()
	if lastCall.Args[0] != 8080 {
		t.Errorf("Expected default port 8080, got %v", lastCall.Args[0])
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

// ==================== map_remote_add ====================

func TestHandleMapRemoteAdd_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleMapRemoteAdd(context.Background(), makeToolRequest(map[string]interface{}{
		"source_pattern": "*/api/*",
		"target_url":     "http://localhost:3000/api/*",
		"method":         "GET",
		"description":    "Redirect to local",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "map-test-id") {
		t.Error("Result should contain the returned ID")
	}

	lastCall := mock.GetLastCallByMethod("AddMapRemoteRule")
	if lastCall == nil {
		t.Fatal("AddMapRemoteRule was not called")
	}
	if lastCall.Args[0] != "*/api/*" {
		t.Errorf("Expected source_pattern '*/api/*', got %v", lastCall.Args[0])
	}
	if lastCall.Args[1] != "http://localhost:3000/api/*" {
		t.Errorf("Expected target_url, got %v", lastCall.Args[1])
	}
	if lastCall.Args[2] != "GET" {
		t.Errorf("Expected method 'GET', got %v", lastCall.Args[2])
	}
	if lastCall.Args[3] != "Redirect to local" {
		t.Errorf("Expected description, got %v", lastCall.Args[3])
	}
}

func TestHandleMapRemoteAdd_MissingRequired(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	// Missing source_pattern
	result, err := server.handleMapRemoteAdd(context.Background(), makeToolRequest(map[string]interface{}{
		"target_url": "http://localhost:3000/*",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Should return error when source_pattern is missing")
	}

	// Missing target_url
	result, err = server.handleMapRemoteAdd(context.Background(), makeToolRequest(map[string]interface{}{
		"source_pattern": "*/api/*",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Should return error when target_url is missing")
	}
}

// ==================== map_remote_list ====================

func TestHandleMapRemoteList_Empty(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleMapRemoteList(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "no map remote") {
		t.Errorf("Expected 'no map remote' message, got: %s", text)
	}

	if !mock.WasMethodCalled("GetMapRemoteRules") {
		t.Error("GetMapRemoteRules should have been called")
	}
}

// ==================== map_remote_update ====================

func TestHandleMapRemoteUpdate_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleMapRemoteUpdate(context.Background(), makeToolRequest(map[string]interface{}{
		"id":             "map-123",
		"source_pattern": "*/v2/*",
		"target_url":     "http://staging/*",
		"method":         "POST",
		"enabled":        false,
		"description":    "Updated rule",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "map-123") {
		t.Error("Result should contain the rule ID")
	}

	lastCall := mock.GetLastCallByMethod("UpdateMapRemoteRule")
	if lastCall == nil {
		t.Fatal("UpdateMapRemoteRule was not called")
	}
	if lastCall.Args[0] != "map-123" {
		t.Errorf("Expected id 'map-123', got %v", lastCall.Args[0])
	}
	if lastCall.Args[1] != "*/v2/*" {
		t.Errorf("Expected source_pattern, got %v", lastCall.Args[1])
	}
	if lastCall.Args[2] != "http://staging/*" {
		t.Errorf("Expected target_url, got %v", lastCall.Args[2])
	}
	if lastCall.Args[3] != "POST" {
		t.Errorf("Expected method 'POST', got %v", lastCall.Args[3])
	}
	if lastCall.Args[4] != false {
		t.Errorf("Expected enabled=false, got %v", lastCall.Args[4])
	}
	if lastCall.Args[5] != "Updated rule" {
		t.Errorf("Expected description, got %v", lastCall.Args[5])
	}
}

// ==================== map_remote_remove ====================

func TestHandleMapRemoteRemove_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleMapRemoteRemove(context.Background(), makeToolRequest(map[string]interface{}{
		"id": "map-456",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "map-456") {
		t.Error("Result should contain the rule ID")
	}

	lastCall := mock.GetLastCallByMethod("RemoveMapRemoteRule")
	if lastCall == nil {
		t.Fatal("RemoveMapRemoteRule was not called")
	}
	if lastCall.Args[0] != "map-456" {
		t.Errorf("Expected id 'map-456', got %v", lastCall.Args[0])
	}
}

func TestHandleMapRemoteRemove_MissingID(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleMapRemoteRemove(context.Background(), makeToolRequest(map[string]interface{}{}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Should return error when id is missing")
	}
}

// ==================== map_remote_toggle ====================

func TestHandleMapRemoteToggle_Enable(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleMapRemoteToggle(context.Background(), makeToolRequest(map[string]interface{}{
		"id":      "map-789",
		"enabled": true,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "enabled") {
		t.Errorf("Expected 'enabled' in result, got: %s", text)
	}

	lastCall := mock.GetLastCallByMethod("ToggleMapRemoteRule")
	if lastCall == nil {
		t.Fatal("ToggleMapRemoteRule was not called")
	}
	if lastCall.Args[0] != "map-789" {
		t.Errorf("Expected id 'map-789', got %v", lastCall.Args[0])
	}
	if lastCall.Args[1] != true {
		t.Errorf("Expected enabled=true, got %v", lastCall.Args[1])
	}
}

func TestHandleMapRemoteToggle_Disable(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleMapRemoteToggle(context.Background(), makeToolRequest(map[string]interface{}{
		"id":      "map-789",
		"enabled": false,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "disabled") {
		t.Errorf("Expected 'disabled' in result, got: %s", text)
	}
}

// ==================== rewrite_rule_add ====================

func TestHandleRewriteRuleAdd_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleRewriteRuleAdd(context.Background(), makeToolRequest(map[string]interface{}{
		"url_pattern": "*/api/*",
		"phase":       "response",
		"target":      "body",
		"match":       `"role":"guest"`,
		"replace":     `"role":"admin"`,
		"method":      "GET",
		"header_name": "",
		"description": "Elevate role",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "rw-test-id") {
		t.Error("Result should contain the returned ID")
	}

	lastCall := mock.GetLastCallByMethod("AddRewriteRule")
	if lastCall == nil {
		t.Fatal("AddRewriteRule was not called")
	}
	// Verify all 8 parameters
	if lastCall.Args[0] != "*/api/*" {
		t.Errorf("Expected url_pattern, got %v", lastCall.Args[0])
	}
	if lastCall.Args[1] != "GET" {
		t.Errorf("Expected method 'GET', got %v", lastCall.Args[1])
	}
	if lastCall.Args[2] != "response" {
		t.Errorf("Expected phase 'response', got %v", lastCall.Args[2])
	}
	if lastCall.Args[3] != "body" {
		t.Errorf("Expected target 'body', got %v", lastCall.Args[3])
	}
	if lastCall.Args[4] != "" {
		t.Errorf("Expected empty header_name, got %v", lastCall.Args[4])
	}
	if lastCall.Args[5] != `"role":"guest"` {
		t.Errorf("Expected match pattern, got %v", lastCall.Args[5])
	}
	if lastCall.Args[6] != `"role":"admin"` {
		t.Errorf("Expected replace string, got %v", lastCall.Args[6])
	}
	if lastCall.Args[7] != "Elevate role" {
		t.Errorf("Expected description, got %v", lastCall.Args[7])
	}
}

func TestHandleRewriteRuleAdd_MissingRequired(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	// Missing url_pattern
	result, _ := server.handleRewriteRuleAdd(context.Background(), makeToolRequest(map[string]interface{}{
		"phase":   "response",
		"target":  "body",
		"match":   "old",
		"replace": "new",
	}))
	if !result.IsError {
		t.Error("Should error when url_pattern missing")
	}

	// Missing phase
	result, _ = server.handleRewriteRuleAdd(context.Background(), makeToolRequest(map[string]interface{}{
		"url_pattern": "*",
		"target":      "body",
		"match":       "old",
		"replace":     "new",
	}))
	if !result.IsError {
		t.Error("Should error when phase missing")
	}

	// Missing target
	result, _ = server.handleRewriteRuleAdd(context.Background(), makeToolRequest(map[string]interface{}{
		"url_pattern": "*",
		"phase":       "response",
		"match":       "old",
		"replace":     "new",
	}))
	if !result.IsError {
		t.Error("Should error when target missing")
	}

	// Missing match
	result, _ = server.handleRewriteRuleAdd(context.Background(), makeToolRequest(map[string]interface{}{
		"url_pattern": "*",
		"phase":       "response",
		"target":      "body",
		"replace":     "new",
	}))
	if !result.IsError {
		t.Error("Should error when match missing")
	}
}

// ==================== rewrite_rule_list ====================

func TestHandleRewriteRuleList_Empty(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleRewriteRuleList(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return empty JSON array
	text := getTextContent(result)
	if !strings.Contains(text, "[]") {
		t.Errorf("Expected empty array, got: %s", text)
	}

	if !mock.WasMethodCalled("GetRewriteRules") {
		t.Error("GetRewriteRules should have been called")
	}
}

// ==================== rewrite_rule_update ====================

func TestHandleRewriteRuleUpdate_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleRewriteRuleUpdate(context.Background(), makeToolRequest(map[string]interface{}{
		"id":          "rw-123",
		"url_pattern": "*/v2/*",
		"phase":       "request",
		"target":      "header",
		"match":       "Bearer.*",
		"replace":     "Bearer new-token",
		"method":      "POST",
		"header_name": "Authorization",
		"enabled":     false,
		"description": "Replace auth token",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "rw-123") {
		t.Error("Result should contain the rule ID")
	}

	lastCall := mock.GetLastCallByMethod("UpdateRewriteRule")
	if lastCall == nil {
		t.Fatal("UpdateRewriteRule was not called")
	}
	// Verify all 10 parameters
	if lastCall.Args[0] != "rw-123" {
		t.Errorf("Expected id, got %v", lastCall.Args[0])
	}
	if lastCall.Args[1] != "*/v2/*" {
		t.Errorf("Expected url_pattern, got %v", lastCall.Args[1])
	}
	if lastCall.Args[2] != "POST" {
		t.Errorf("Expected method, got %v", lastCall.Args[2])
	}
	if lastCall.Args[3] != "request" {
		t.Errorf("Expected phase, got %v", lastCall.Args[3])
	}
	if lastCall.Args[4] != "header" {
		t.Errorf("Expected target, got %v", lastCall.Args[4])
	}
	if lastCall.Args[5] != "Authorization" {
		t.Errorf("Expected header_name, got %v", lastCall.Args[5])
	}
	if lastCall.Args[6] != "Bearer.*" {
		t.Errorf("Expected match, got %v", lastCall.Args[6])
	}
	if lastCall.Args[7] != "Bearer new-token" {
		t.Errorf("Expected replace, got %v", lastCall.Args[7])
	}
	if lastCall.Args[8] != false {
		t.Errorf("Expected enabled=false, got %v", lastCall.Args[8])
	}
	if lastCall.Args[9] != "Replace auth token" {
		t.Errorf("Expected description, got %v", lastCall.Args[9])
	}
}

// ==================== rewrite_rule_remove ====================

func TestHandleRewriteRuleRemove_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleRewriteRuleRemove(context.Background(), makeToolRequest(map[string]interface{}{
		"id": "rw-456",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "rw-456") {
		t.Error("Result should contain the rule ID")
	}

	lastCall := mock.GetLastCallByMethod("RemoveRewriteRule")
	if lastCall == nil {
		t.Fatal("RemoveRewriteRule was not called")
	}
	if lastCall.Args[0] != "rw-456" {
		t.Errorf("Expected id 'rw-456', got %v", lastCall.Args[0])
	}
}

func TestHandleRewriteRuleRemove_MissingID(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, _ := server.handleRewriteRuleRemove(context.Background(), makeToolRequest(map[string]interface{}{}))
	if !result.IsError {
		t.Error("Should return error when id is missing")
	}
}

// ==================== rewrite_rule_toggle ====================

func TestHandleRewriteRuleToggle_Enable(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleRewriteRuleToggle(context.Background(), makeToolRequest(map[string]interface{}{
		"id":      "rw-789",
		"enabled": true,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "enabled") {
		t.Errorf("Expected 'enabled' in result, got: %s", text)
	}

	lastCall := mock.GetLastCallByMethod("ToggleRewriteRule")
	if lastCall == nil {
		t.Fatal("ToggleRewriteRule was not called")
	}
	if lastCall.Args[0] != "rw-789" {
		t.Errorf("Expected id 'rw-789', got %v", lastCall.Args[0])
	}
	if lastCall.Args[1] != true {
		t.Errorf("Expected enabled=true, got %v", lastCall.Args[1])
	}
}

func TestHandleRewriteRuleToggle_Disable(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleRewriteRuleToggle(context.Background(), makeToolRequest(map[string]interface{}{
		"id":      "rw-789",
		"enabled": false,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "disabled") {
		t.Errorf("Expected 'disabled' in result, got: %s", text)
	}
}

func TestHandleRewriteRuleToggle_MissingID(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, _ := server.handleRewriteRuleToggle(context.Background(), makeToolRequest(map[string]interface{}{
		"enabled": true,
	}))
	if !result.IsError {
		t.Error("Should return error when id is missing")
	}
}

// ==================== mock_rule_export ====================

func TestHandleMockRuleExport_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleMockRuleExport(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if text != "[]" {
		t.Errorf("Expected '[]' from mock, got: %s", text)
	}

	if !mock.WasMethodCalled("ExportMockRules") {
		t.Error("ExportMockRules should have been called")
	}
}

// ==================== mock_rule_import ====================

func TestHandleMockRuleImport_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	jsonInput := `[{"urlPattern":"*/api/*","statusCode":200,"body":"{}"}]`
	result, err := server.handleMockRuleImport(context.Background(), makeToolRequest(map[string]interface{}{
		"json": jsonInput,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "imported") {
		t.Errorf("Expected 'imported' in result, got: %s", text)
	}

	lastCall := mock.GetLastCallByMethod("ImportMockRules")
	if lastCall == nil {
		t.Fatal("ImportMockRules was not called")
	}
	if lastCall.Args[0] != jsonInput {
		t.Errorf("Expected JSON input to be passed through, got %v", lastCall.Args[0])
	}
}

func TestHandleMockRuleImport_MissingJSON(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleMockRuleImport(context.Background(), makeToolRequest(map[string]interface{}{}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Should return error when json is missing")
	}
}
