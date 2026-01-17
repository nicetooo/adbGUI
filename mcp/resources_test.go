package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Helper to create a ReadResourceRequest
func makeResourceRequest(uri string) mcp.ReadResourceRequest {
	return mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: uri,
		},
	}
}

// Helper to get text from resource contents
func getResourceText(contents []mcp.ResourceContents) string {
	if len(contents) == 0 {
		return ""
	}
	if tc, ok := contents[0].(mcp.TextResourceContents); ok {
		return tc.Text
	}
	return ""
}

// ==================== gaze://devices ====================

func TestHandleDevicesResource_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithDevices(
		SampleDevice("device1"),
		SampleDevice("device2"),
	)
	server := NewMCPServer(mock)

	contents, err := server.handleDevicesResource(context.Background(), makeResourceRequest("gaze://devices"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(contents) == 0 {
		t.Fatal("Expected at least one content item")
	}

	text := getResourceText(contents)

	// Should be valid JSON
	var devices []Device
	if err := json.Unmarshal([]byte(text), &devices); err != nil {
		t.Fatalf("Result should be valid JSON: %v", err)
	}

	if len(devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(devices))
	}
}

func TestHandleDevicesResource_Empty(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetDevicesResult = []Device{}
	server := NewMCPServer(mock)

	contents, err := server.handleDevicesResource(context.Background(), makeResourceRequest("gaze://devices"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getResourceText(contents)

	var devices []Device
	if err := json.Unmarshal([]byte(text), &devices); err != nil {
		t.Fatalf("Result should be valid JSON: %v", err)
	}

	if len(devices) != 0 {
		t.Errorf("Expected 0 devices, got %d", len(devices))
	}
}

func TestHandleDevicesResource_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("GetDevices", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleDevicesResource(context.Background(), makeResourceRequest("gaze://devices"))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== gaze://devices/{deviceId} ====================

func TestHandleDeviceInfoResource_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetDeviceInfoResult = SampleDeviceInfo()
	server := NewMCPServer(mock)

	contents, err := server.handleDeviceInfoResource(context.Background(), makeResourceRequest("gaze://devices/device1"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getResourceText(contents)

	// Should be valid JSON
	var info DeviceInfo
	if err := json.Unmarshal([]byte(text), &info); err != nil {
		t.Fatalf("Result should be valid JSON: %v", err)
	}

	if info.Model != "Pixel 6" {
		t.Errorf("Expected model 'Pixel 6', got '%s'", info.Model)
	}

	// Verify correct device ID was extracted
	if !mock.WasMethodCalled("GetDeviceInfo") {
		t.Error("GetDeviceInfo should have been called")
	}
	lastCall := mock.GetLastCall()
	if lastCall.Args[0] != "device1" {
		t.Errorf("Expected device ID 'device1', got %v", lastCall.Args[0])
	}
}

func TestHandleDeviceInfoResource_InvalidURI(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleDeviceInfoResource(context.Background(), makeResourceRequest("gaze://devices"))
	if err == nil {
		t.Error("Expected error for invalid URI")
	}
}

func TestHandleDeviceInfoResource_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("GetDeviceInfo", ErrDeviceNotFound)
	server := NewMCPServer(mock)

	_, err := server.handleDeviceInfoResource(context.Background(), makeResourceRequest("gaze://devices/nonexistent"))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== gaze://sessions ====================

func TestHandleSessionsResource_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ListStoredSessionsResult = []DeviceSession{
		SampleSession("sess1", "device1"),
		SampleSession("sess2", "device2"),
	}
	server := NewMCPServer(mock)

	contents, err := server.handleSessionsResource(context.Background(), makeResourceRequest("gaze://sessions"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getResourceText(contents)

	var sessions []DeviceSession
	if err := json.Unmarshal([]byte(text), &sessions); err != nil {
		t.Fatalf("Result should be valid JSON: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}
}

func TestHandleSessionsResource_Empty(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ListStoredSessionsResult = []DeviceSession{}
	server := NewMCPServer(mock)

	contents, err := server.handleSessionsResource(context.Background(), makeResourceRequest("gaze://sessions"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getResourceText(contents)

	var sessions []DeviceSession
	if err := json.Unmarshal([]byte(text), &sessions); err != nil {
		t.Fatalf("Result should be valid JSON: %v", err)
	}

	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}
}

func TestHandleSessionsResource_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("ListStoredSessions", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleSessionsResource(context.Background(), makeResourceRequest("gaze://sessions"))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== workflow://list ====================

func TestHandleWorkflowsResource_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Login Flow"),
		SampleWorkflow("wf2", "Purchase Flow"),
	}
	server := NewMCPServer(mock)

	contents, err := server.handleWorkflowsResource(context.Background(), makeResourceRequest("workflow://list"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getResourceText(contents)

	var workflows []Workflow
	if err := json.Unmarshal([]byte(text), &workflows); err != nil {
		t.Fatalf("Result should be valid JSON: %v", err)
	}

	if len(workflows) != 2 {
		t.Errorf("Expected 2 workflows, got %d", len(workflows))
	}
}

func TestHandleWorkflowsResource_Empty(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{}
	server := NewMCPServer(mock)

	contents, err := server.handleWorkflowsResource(context.Background(), makeResourceRequest("workflow://list"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getResourceText(contents)

	var workflows []Workflow
	if err := json.Unmarshal([]byte(text), &workflows); err != nil {
		t.Fatalf("Result should be valid JSON: %v", err)
	}

	if len(workflows) != 0 {
		t.Errorf("Expected 0 workflows, got %d", len(workflows))
	}
}

func TestHandleWorkflowsResource_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("LoadWorkflows", ErrPermissionDenied)
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowsResource(context.Background(), makeResourceRequest("workflow://list"))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== workflow://{workflowId} ====================

func TestHandleWorkflowResource_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Login Flow"),
		SampleWorkflow("wf2", "Purchase Flow"),
	}
	server := NewMCPServer(mock)

	contents, err := server.handleWorkflowResource(context.Background(), makeResourceRequest("workflow://wf1"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getResourceText(contents)

	var workflow Workflow
	if err := json.Unmarshal([]byte(text), &workflow); err != nil {
		t.Fatalf("Result should be valid JSON: %v", err)
	}

	if workflow.ID != "wf1" {
		t.Errorf("Expected workflow ID 'wf1', got '%s'", workflow.ID)
	}
	if workflow.Name != "Login Flow" {
		t.Errorf("Expected workflow name 'Login Flow', got '%s'", workflow.Name)
	}
}

func TestHandleWorkflowResource_NotFound(t *testing.T) {
	mock := NewMockGazeApp()
	mock.LoadWorkflowsResult = []Workflow{
		SampleWorkflow("wf1", "Login Flow"),
	}
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowResource(context.Background(), makeResourceRequest("workflow://nonexistent"))
	if err == nil {
		t.Error("Expected error for nonexistent workflow")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found', got: %v", err)
	}
}

func TestHandleWorkflowResource_InvalidURI(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowResource(context.Background(), makeResourceRequest("workflow://"))
	if err == nil {
		t.Error("Expected error for invalid URI")
	}
}

func TestHandleWorkflowResource_LoadError(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("LoadWorkflows", ErrPermissionDenied)
	server := NewMCPServer(mock)

	_, err := server.handleWorkflowResource(context.Background(), makeResourceRequest("workflow://wf1"))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
