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

// ==================== workflow_create ====================

func TestHandleWorkflowCreate_Success(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowCreate(context.Background(), makeToolRequest(map[string]interface{}{
		"name":        "Test Workflow",
		"description": "A test workflow",
		"steps_json":  `[{"type":"tap","tap":{"x":540,"y":960}}]`,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "Test Workflow") {
		t.Error("Result should contain workflow name")
	}
	if !strings.Contains(strings.ToLower(text), "created") {
		t.Error("Result should indicate workflow was created")
	}
	// Verify SaveWorkflow was called
	if !mock.WasMethodCalled("SaveWorkflow") {
		t.Error("SaveWorkflow should have been called")
	}
}

func TestHandleWorkflowCreate_WithVariables(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowCreate(context.Background(), makeToolRequest(map[string]interface{}{
		"name":           "Login Workflow",
		"description":    "Workflow with variables",
		"steps_json":     `[{"type":"input_text","element":{"selector":{"type":"id","value":"com.app:id/username"},"action":"input","inputText":"{{username}}"}}]`,
		"variables_json": `{"username":"test_user","password":"test_pass"}`,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "Login Workflow") {
		t.Error("Result should contain workflow name")
	}
	if !strings.Contains(text, "Variables: 2") {
		t.Errorf("Result should mention 2 variables, got: %s", text)
	}
	if !strings.Contains(text, "username") {
		t.Error("Result should list username variable")
	}
	if !strings.Contains(text, "password") {
		t.Error("Result should list password variable")
	}

	// Verify SaveWorkflow was called
	if !mock.WasMethodCalled("SaveWorkflow") {
		t.Error("SaveWorkflow should have been called")
	}
}

func TestHandleWorkflowCreate_InvalidVariablesJSON(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowCreate(context.Background(), makeToolRequest(map[string]interface{}{
		"name":           "Test Workflow",
		"variables_json": `{"invalid json`,
	}))
	if err == nil {
		t.Error("Expected error for invalid variables JSON")
	}
	if !strings.Contains(err.Error(), "variables_json") {
		t.Errorf("Error should mention variables_json, got: %v", err)
	}
}

func TestHandleWorkflowCreate_InvalidStepsJSON(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowCreate(context.Background(), makeToolRequest(map[string]interface{}{
		"name":       "Test Workflow",
		"steps_json": `[{"invalid json`,
	}))
	if err == nil {
		t.Error("Expected error for invalid steps JSON")
	}
	if !strings.Contains(err.Error(), "steps_json") {
		t.Errorf("Error should mention steps_json, got: %v", err)
	}
}

func TestHandleWorkflowCreate_MissingName(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowCreate(context.Background(), makeToolRequest(map[string]interface{}{
		"description": "A test workflow",
	}))
	if err == nil {
		t.Error("Expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("Error should mention name, got: %v", err)
	}
}

// ==================== workflow_update ====================

func TestHandleWorkflowUpdate_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Original Name"),
	}
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowUpdate(context.Background(), makeToolRequest(map[string]interface{}{
		"workflow_id": "wf1",
		"name":        "Updated Name",
		"description": "Updated description",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "Updated Name") {
		t.Error("Result should contain updated workflow name")
	}
	if !strings.Contains(strings.ToLower(text), "updated") {
		t.Error("Result should indicate workflow was updated")
	}
}

func TestHandleWorkflowUpdate_WithVariables(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Test Workflow"),
	}
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowUpdate(context.Background(), makeToolRequest(map[string]interface{}{
		"workflow_id":    "wf1",
		"variables_json": `{"api_url":"https://api.example.com","timeout":"30"}`,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "Variables: 2") {
		t.Errorf("Result should mention 2 variables, got: %s", text)
	}

	// Verify SaveWorkflow was called
	if !mock.WasMethodCalled("SaveWorkflow") {
		t.Error("SaveWorkflow should have been called")
	}
}

func TestHandleWorkflowUpdate_InvalidVariablesJSON(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Test Workflow"),
	}
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowUpdate(context.Background(), makeToolRequest(map[string]interface{}{
		"workflow_id":    "wf1",
		"variables_json": `not valid json`,
	}))
	if err == nil {
		t.Error("Expected error for invalid variables JSON")
	}
	if !strings.Contains(err.Error(), "variables_json") {
		t.Errorf("Error should mention variables_json, got: %v", err)
	}
}

func TestHandleWorkflowUpdate_WorkflowNotFound(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{} // No workflows
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowUpdate(context.Background(), makeToolRequest(map[string]interface{}{
		"workflow_id": "nonexistent",
		"name":        "New Name",
	}))
	if err == nil {
		t.Error("Expected error for nonexistent workflow")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found', got: %v", err)
	}
}

func TestHandleWorkflowUpdate_NoFieldsProvided(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Test Workflow"),
	}
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowUpdate(context.Background(), makeToolRequest(map[string]interface{}{
		"workflow_id": "wf1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "no fields") {
		t.Errorf("Result should indicate no fields to update, got: %s", text)
	}
}

// ==================== workflow_run with variables ====================

func TestHandleWorkflowRun_WithRuntimeVariables(t *testing.T) {
	mock := NewMockGazeApp()
	// Create a workflow with default variables
	wf := SampleWorkflow("wf1", "Test Workflow")
	wf.Variables = map[string]string{
		"username": "default_user",
		"password": "default_pass",
	}
	mock.LoadWorkflowsResult = []Workflow{wf}
	mock.SetupWithDevices(SampleDevice("device1"))
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowRun(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":      "device1",
		"workflow_id":    "wf1",
		"variables_json": `{"username":"runtime_user","new_var":"new_value"}`,
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "start") {
		t.Error("Result should indicate workflow started")
	}

	// Verify RunWorkflow was called
	if !mock.WasMethodCalled("RunWorkflow") {
		t.Error("RunWorkflow should have been called")
	}
}

func TestHandleWorkflowRun_InvalidVariablesJSON(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Test Workflow"),
	}
	mock.SetupWithDevices(SampleDevice("device1"))
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowRun(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":      "device1",
		"workflow_id":    "wf1",
		"variables_json": `{"broken json`,
	}))
	if err == nil {
		t.Error("Expected error for invalid variables JSON")
	}
	if !strings.Contains(err.Error(), "variables_json") {
		t.Errorf("Error should mention variables_json, got: %v", err)
	}
}

func TestHandleWorkflowRun_AlreadyRunning(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Test Workflow"),
	}
	mock.SetupWithDevices(SampleDevice("device1"))
	mock.IsWorkflowRunningResult = true // Workflow already running
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowRun(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":   "device1",
		"workflow_id": "wf1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "already running") {
		t.Errorf("Result should indicate workflow already running, got: %s", text)
	}
}

// ==================== workflow_get ====================

func TestHandleWorkflowGet_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Test Workflow"),
	}
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowGet(context.Background(), makeToolRequest(map[string]interface{}{
		"workflow_id": "wf1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "Test Workflow") {
		t.Error("Result should contain workflow name")
	}
	if !strings.Contains(text, "wf1") {
		t.Error("Result should contain workflow ID")
	}
	if !strings.Contains(text, "Steps") {
		t.Error("Result should show steps information")
	}
}

func TestHandleWorkflowGet_NotFound(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{}
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowGet(context.Background(), makeToolRequest(map[string]interface{}{
		"workflow_id": "nonexistent",
	}))
	if err == nil {
		t.Error("Expected error for nonexistent workflow")
	}
}

func TestHandleWorkflowGet_MissingWorkflowId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowGet(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing workflow_id")
	}
}

// ==================== workflow_delete ====================

func TestHandleWorkflowDelete_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Test Workflow"),
	}
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowDelete(context.Background(), makeToolRequest(map[string]interface{}{
		"workflow_id": "wf1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "deleted") {
		t.Error("Result should indicate workflow was deleted")
	}
	if !strings.Contains(text, "Test Workflow") {
		t.Error("Result should contain workflow name")
	}

	// Verify DeleteWorkflow was called
	if !mock.WasMethodCalled("DeleteWorkflow") {
		t.Error("DeleteWorkflow should have been called")
	}
}

func TestHandleWorkflowDelete_NotFound(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{}
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowDelete(context.Background(), makeToolRequest(map[string]interface{}{
		"workflow_id": "nonexistent",
	}))
	if err == nil {
		t.Error("Expected error for nonexistent workflow")
	}
}

// ==================== workflow_status ====================

func TestHandleWorkflowStatus_Running(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsWorkflowRunningResult = true
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowStatus(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToUpper(text), "RUNNING") {
		t.Errorf("Result should indicate RUNNING, got: %s", text)
	}
}

func TestHandleWorkflowStatus_Idle(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsWorkflowRunningResult = false
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowStatus(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToUpper(text), "IDLE") {
		t.Errorf("Result should indicate IDLE, got: %s", text)
	}
}

func TestHandleWorkflowStatus_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowStatus(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

// ==================== workflow_pause/resume ====================

func TestHandleWorkflowPause_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsWorkflowRunningResult = true
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowPause(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "pause") {
		t.Errorf("Result should indicate paused, got: %s", text)
	}

	// Verify PauseTask was called
	if !mock.WasMethodCalled("PauseTask") {
		t.Error("PauseTask should have been called")
	}
}

func TestHandleWorkflowPause_NotRunning(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsWorkflowRunningResult = false
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowPause(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "no workflow") {
		t.Errorf("Result should indicate no workflow running, got: %s", text)
	}
}

func TestHandleWorkflowResume_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsWorkflowRunningResult = true
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowResume(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "resume") {
		t.Errorf("Result should indicate resumed, got: %s", text)
	}

	// Verify ResumeTask was called
	if !mock.WasMethodCalled("ResumeTask") {
		t.Error("ResumeTask should have been called")
	}
}

func TestHandleWorkflowResume_NotRunning(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsWorkflowRunningResult = false
	server := NewMCPServer(mock)

	result, err := server.handleWorkflowResume(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "no workflow") {
		t.Errorf("Result should indicate no workflow running, got: %s", text)
	}
}
