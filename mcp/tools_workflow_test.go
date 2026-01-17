package mcp

import (
	"context"
	"strings"
	"testing"
)

// ==================== workflow_list ====================

func TestHandleWorkflowList_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Login Flow"),
		SampleWorkflow("wf2", "Purchase Flow"),
	}
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowList(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "Login Flow") {
		t.Error("Result should contain workflow 1 name")
	}
	if !strings.Contains(text, "Purchase Flow") {
		t.Error("Result should contain workflow 2 name")
	}
	if !strings.Contains(text, "wf1") {
		t.Error("Result should contain workflow 1 ID")
	}
	if !strings.Contains(text, "2") {
		t.Error("Result should mention number of workflows")
	}
}

func TestHandleWorkflowList_Empty(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{}
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowList(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "no") {
		t.Errorf("Result should indicate no workflows, got: %s", text)
	}
}

func TestHandleWorkflowList_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("LoadWorkflows", ErrPermissionDenied)
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowList(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== workflow_run ====================

func TestHandleWorkflowRun_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Test Workflow"),
	}
	mock.SetupWithDevices(SampleDevice("device1"))
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowRun(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":   "device1",
		"workflow_id": "wf1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "start") {
		t.Error("Result should indicate workflow started")
	}
	if !strings.Contains(text, "Test Workflow") {
		t.Error("Result should contain workflow name")
	}
}

func TestHandleWorkflowRun_WorkflowNotFound(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{}
	mock.SetupWithDevices(SampleDevice("device1"))
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowRun(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":   "device1",
		"workflow_id": "nonexistent",
	}))
	if err == nil {
		t.Error("Expected error for nonexistent workflow")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found', got: %v", err)
	}
}

func TestHandleWorkflowRun_DeviceNotFound(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Test Workflow"),
	}
	mock.GetDevicesResult = []Device{} // No devices
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowRun(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":   "nonexistent",
		"workflow_id": "wf1",
	}))
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestHandleWorkflowRun_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowRun(context.Background(), makeToolRequest(map[string]interface{}{
		"workflow_id": "wf1",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleWorkflowRun_MissingWorkflowId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowRun(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error for missing workflow_id")
	}
}

func TestHandleWorkflowRun_LoadWorkflowsError(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("LoadWorkflows", ErrPermissionDenied)
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowRun(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":   "device1",
		"workflow_id": "wf1",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== workflow_stop ====================

func TestHandleWorkflowStop_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithDevices(SampleDevice("device1"))
	mock.IsWorkflowRunningResult = true // Workflow must be running to stop it
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "stop") {
		t.Error("Result should indicate workflow stopped")
	}

	// Verify StopWorkflow was called
	if !mock.WasMethodCalled("StopWorkflow") {
		t.Error("StopWorkflow should have been called")
	}
}

func TestHandleWorkflowStop_NotRunning(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithDevices(SampleDevice("device1"))
	mock.IsWorkflowRunningResult = false // No workflow running
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "no workflow") {
		t.Error("Result should indicate no workflow is running")
	}

	// StopWorkflow should NOT have been called
	if mock.WasMethodCalled("StopWorkflow") {
		t.Error("StopWorkflow should not have been called when no workflow is running")
	}
}

func TestHandleWorkflowStop_DeviceNotFound(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetDevicesResult = []Device{}  // No devices
	mock.IsWorkflowRunningResult = true // Workflow is running (so we try to stop)
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "nonexistent",
	}))
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestHandleWorkflowStop_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowStop(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleWorkflowStop_GetDevicesError(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("GetDevices", ErrDeviceOffline)
	mock.IsWorkflowRunningResult = true // Workflow is running (so we try to stop)
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
