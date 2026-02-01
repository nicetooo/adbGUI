package mcp

import (
	"context"
	"strings"
	"testing"
)

// ==================== touch_record_start ====================

func TestHandleTouchRecordStart_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchRecordStart(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"mode":      "fast",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "recording started") {
		t.Errorf("Expected 'recording started' in result, got: %s", text)
	}
	if !strings.Contains(text, "fast") {
		t.Errorf("Expected 'fast' mode in result, got: %s", text)
	}
}

func TestHandleTouchRecordStart_DefaultMode(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchRecordStart(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "fast") {
		t.Errorf("Expected default 'fast' mode in result, got: %s", text)
	}
}

func TestHandleTouchRecordStart_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchRecordStart(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result for missing device_id")
	}
}

func TestHandleTouchRecordStart_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("StartTouchRecording", ErrDeviceOffline)
	server := NewMCPServer(mock)

	result, err := server.handleTouchRecordStart(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result")
	}
}

// ==================== touch_record_stop ====================

func TestHandleTouchRecordStop_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.StopTouchRecordingResult = &TouchScript{
		Name:       "recorded_script",
		DeviceID:   "device1",
		Resolution: "1080x1920",
		Events: []TouchEvent{
			{Type: "tap", X: 540, Y: 960, Timestamp: 0},
			{Type: "tap", X: 100, Y: 200, Timestamp: 1000},
		},
	}
	server := NewMCPServer(mock)

	result, err := server.handleTouchRecordStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "recorded_script") {
		t.Errorf("Expected script name in result, got: %s", text)
	}
	if !strings.Contains(text, "1080x1920") {
		t.Errorf("Expected resolution in result, got: %s", text)
	}
}

func TestHandleTouchRecordStop_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchRecordStop(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result for missing device_id")
	}
}

func TestHandleTouchRecordStop_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("StopTouchRecording", ErrDeviceOffline)
	server := NewMCPServer(mock)

	result, err := server.handleTouchRecordStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result")
	}
}

// ==================== touch_record_status ====================

func TestHandleTouchRecordStatus_Recording(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsRecordingTouchResult = true
	server := NewMCPServer(mock)

	result, err := server.handleTouchRecordStatus(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "true") {
		t.Errorf("Expected isRecording=true, got: %s", text)
	}
}

func TestHandleTouchRecordStatus_NotRecording(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsRecordingTouchResult = false
	server := NewMCPServer(mock)

	result, err := server.handleTouchRecordStatus(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "false") {
		t.Errorf("Expected isRecording=false, got: %s", text)
	}
}

func TestHandleTouchRecordStatus_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchRecordStatus(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result for missing device_id")
	}
}

// ==================== touch_script_list ====================

func TestHandleTouchScriptList_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadTouchScriptsResult = []TouchScript{
		{Name: "script1", Resolution: "1080x1920", DeviceModel: "Pixel 7", Events: []TouchEvent{{Type: "tap"}, {Type: "swipe"}}},
		{Name: "script2", Resolution: "720x1280", Events: []TouchEvent{{Type: "tap"}}},
	}
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptList(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "script1") || !strings.Contains(text, "script2") {
		t.Errorf("Expected both script names in result, got: %s", text)
	}
	if !strings.Contains(text, "Pixel 7") {
		t.Errorf("Expected device model in result, got: %s", text)
	}
}

func TestHandleTouchScriptList_Empty(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadTouchScriptsResult = []TouchScript{}
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptList(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "[]") {
		t.Errorf("Expected empty list, got: %s", text)
	}
}

func TestHandleTouchScriptList_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("LoadTouchScripts", ErrDeviceOffline)
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptList(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result")
	}
}

// ==================== touch_script_play ====================

func TestHandleTouchScriptPlay_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadTouchScriptsResult = []TouchScript{
		{Name: "my_script", Resolution: "1080x1920", Events: []TouchEvent{{Type: "tap", X: 100, Y: 200}}},
	}
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptPlay(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":   "device1",
		"script_name": "my_script",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "Playing") && !strings.Contains(text, "my_script") {
		t.Errorf("Expected play confirmation, got: %s", text)
	}
}

func TestHandleTouchScriptPlay_ScriptNotFound(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadTouchScriptsResult = []TouchScript{
		{Name: "other_script"},
	}
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptPlay(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":   "device1",
		"script_name": "nonexistent",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for nonexistent script")
	}
	text := getTextContent(result)
	if !strings.Contains(text, "not found") {
		t.Errorf("Expected 'not found' in error, got: %s", text)
	}
}

func TestHandleTouchScriptPlay_MissingParams(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	// Missing both
	result, err := server.handleTouchScriptPlay(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for missing params")
	}

	// Missing script_name
	result, err = server.handleTouchScriptPlay(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for missing script_name")
	}
}

func TestHandleTouchScriptPlay_PlayError(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadTouchScriptsResult = []TouchScript{
		{Name: "my_script", Events: []TouchEvent{{Type: "tap"}}},
	}
	mock.SetupWithError("PlayTouchScript", ErrDeviceOffline)
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptPlay(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":   "device1",
		"script_name": "my_script",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result from play failure")
	}
}

// ==================== touch_script_save ====================

func TestHandleTouchScriptSave_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptSave(context.Background(), makeToolRequest(map[string]interface{}{
		"script_json": `{"name":"test_script","resolution":"1080x1920","events":[{"type":"tap","x":540,"y":960}]}`,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "saved") {
		t.Errorf("Expected 'saved' confirmation, got: %s", text)
	}
}

func TestHandleTouchScriptSave_InvalidJSON(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptSave(context.Background(), makeToolRequest(map[string]interface{}{
		"script_json": `{invalid json}`,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for invalid JSON")
	}
}

func TestHandleTouchScriptSave_MissingName(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptSave(context.Background(), makeToolRequest(map[string]interface{}{
		"script_json": `{"resolution":"1080x1920","events":[]}`,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for missing script name")
	}
}

func TestHandleTouchScriptSave_MissingParam(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptSave(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for missing script_json")
	}
}

func TestHandleTouchScriptSave_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("SaveTouchScript", ErrDeviceOffline)
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptSave(context.Background(), makeToolRequest(map[string]interface{}{
		"script_json": `{"name":"test","events":[]}`,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result")
	}
}

// ==================== touch_script_delete ====================

func TestHandleTouchScriptDelete_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptDelete(context.Background(), makeToolRequest(map[string]interface{}{
		"script_name": "my_script",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "deleted") {
		t.Errorf("Expected 'deleted' confirmation, got: %s", text)
	}
}

func TestHandleTouchScriptDelete_MissingName(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptDelete(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for missing script_name")
	}
}

func TestHandleTouchScriptDelete_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("DeleteTouchScript", ErrDeviceOffline)
	server := NewMCPServer(mock)

	result, err := server.handleTouchScriptDelete(context.Background(), makeToolRequest(map[string]interface{}{
		"script_name": "my_script",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result")
	}
}

// ==================== touch_playback_stop ====================

func TestHandleTouchPlaybackStop_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchPlaybackStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "stopped") {
		t.Errorf("Expected 'stopped' in result, got: %s", text)
	}
}

func TestHandleTouchPlaybackStop_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleTouchPlaybackStop(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for missing device_id")
	}
}
