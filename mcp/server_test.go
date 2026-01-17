package mcp

import (
	"testing"
)

// TestNewMCPServer tests server creation
func TestNewMCPServer(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	if server == nil {
		t.Fatal("NewMCPServer should not return nil")
	}

	if server.app == nil {
		t.Error("server.app should not be nil")
	}

	if server.server == nil {
		t.Error("server.server (underlying MCP server) should not be nil")
	}

	// Verify GetAppVersion was called during initialization
	if !mock.WasMethodCalled("GetAppVersion") {
		t.Error("GetAppVersion should be called during server creation")
	}
}

// TestMCPServer_IsRunning tests the IsRunning method
func TestMCPServer_IsRunning(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	// Initially should not be running
	if server.IsRunning() {
		t.Error("Server should not be running initially")
	}
}

// TestMCPServer_Stop tests the Stop method
func TestMCPServer_Stop(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	// Stop should not panic even when not running
	server.Stop()

	if server.IsRunning() {
		t.Error("Server should not be running after Stop")
	}
}

// TestMockGazeApp_Interface verifies MockGazeApp implements GazeApp
func TestMockGazeApp_Interface(t *testing.T) {
	var _ GazeApp = (*MockGazeApp)(nil)
}

// TestMockGazeApp_RecordsCalls tests call recording
func TestMockGazeApp_RecordsCalls(t *testing.T) {
	mock := NewMockGazeApp()

	// Make some calls
	mock.GetDevices(false)
	mock.GetDeviceInfo("device1")
	mock.StartApp("device1", "com.example.app")

	calls := mock.GetCalls()
	if len(calls) != 3 {
		t.Errorf("Expected 3 calls, got %d", len(calls))
	}

	// Verify call order and arguments
	if calls[0].Method != "GetDevices" {
		t.Errorf("Expected first call to be GetDevices, got %s", calls[0].Method)
	}

	if calls[1].Method != "GetDeviceInfo" {
		t.Errorf("Expected second call to be GetDeviceInfo, got %s", calls[1].Method)
	}

	if calls[1].Args[0] != "device1" {
		t.Errorf("Expected device1 argument, got %v", calls[1].Args[0])
	}

	if calls[2].Method != "StartApp" {
		t.Errorf("Expected third call to be StartApp, got %s", calls[2].Method)
	}
}

// TestMockGazeApp_ResetCalls tests clearing call history
func TestMockGazeApp_ResetCalls(t *testing.T) {
	mock := NewMockGazeApp()

	mock.GetDevices(false)
	mock.ResetCalls()

	if len(mock.GetCalls()) != 0 {
		t.Error("Calls should be empty after ResetCalls")
	}
}

// TestMockGazeApp_GetLastCall tests getting the last call
func TestMockGazeApp_GetLastCall(t *testing.T) {
	mock := NewMockGazeApp()

	// No calls yet
	if mock.GetLastCall() != nil {
		t.Error("GetLastCall should return nil when no calls made")
	}

	mock.GetDevices(false)
	mock.GetDeviceInfo("device1")

	last := mock.GetLastCall()
	if last == nil {
		t.Fatal("GetLastCall should not return nil")
	}

	if last.Method != "GetDeviceInfo" {
		t.Errorf("Expected last call to be GetDeviceInfo, got %s", last.Method)
	}
}

// TestMockGazeApp_WasMethodCalled tests method call checking
func TestMockGazeApp_WasMethodCalled(t *testing.T) {
	mock := NewMockGazeApp()

	if mock.WasMethodCalled("GetDevices") {
		t.Error("GetDevices should not have been called yet")
	}

	mock.GetDevices(false)

	if !mock.WasMethodCalled("GetDevices") {
		t.Error("GetDevices should have been called")
	}

	if mock.WasMethodCalled("GetDeviceInfo") {
		t.Error("GetDeviceInfo should not have been called")
	}
}

// TestMockGazeApp_SetupWithDevices tests the helper method
func TestMockGazeApp_SetupWithDevices(t *testing.T) {
	mock := NewMockGazeApp()
	device := SampleDevice("test123")

	mock.SetupWithDevices(device)

	devices, err := mock.GetDevices(false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(devices))
	}

	if devices[0].ID != "test123" {
		t.Errorf("Expected device ID test123, got %s", devices[0].ID)
	}
}

// TestMockGazeApp_SetupWithError tests the error configuration
func TestMockGazeApp_SetupWithError(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("GetDevices", ErrDeviceNotFound)

	_, err := mock.GetDevices(false)
	if err != ErrDeviceNotFound {
		t.Errorf("Expected ErrDeviceNotFound, got %v", err)
	}
}

// TestSampleDevice tests the sample device factory
func TestSampleDevice(t *testing.T) {
	device := SampleDevice("device123")

	if device.ID != "device123" {
		t.Errorf("Expected ID device123, got %s", device.ID)
	}

	if device.State != "device" {
		t.Errorf("Expected state 'device', got %s", device.State)
	}

	if device.Model == "" {
		t.Error("Model should not be empty")
	}
}

// TestSampleDeviceInfo tests the sample device info factory
func TestSampleDeviceInfo(t *testing.T) {
	info := SampleDeviceInfo()

	if info.Model == "" {
		t.Error("Model should not be empty")
	}

	if info.AndroidVer == "" {
		t.Error("AndroidVer should not be empty")
	}

	if info.Resolution == "" {
		t.Error("Resolution should not be empty")
	}
}

// TestSampleAppPackage tests the sample app package factory
func TestSampleAppPackage(t *testing.T) {
	pkg := SampleAppPackage("com.example.app")

	if pkg.Name != "com.example.app" {
		t.Errorf("Expected name com.example.app, got %s", pkg.Name)
	}

	if pkg.Type != "user" {
		t.Errorf("Expected type 'user', got %s", pkg.Type)
	}
}

// TestSampleWorkflow tests the sample workflow factory
func TestSampleWorkflow(t *testing.T) {
	workflow := SampleWorkflow("wf1", "Test Workflow")

	if workflow.ID != "wf1" {
		t.Errorf("Expected ID wf1, got %s", workflow.ID)
	}

	if workflow.Name != "Test Workflow" {
		t.Errorf("Expected name 'Test Workflow', got %s", workflow.Name)
	}

	if len(workflow.Steps) == 0 {
		t.Error("Workflow should have at least one step")
	}
}

// TestSampleSession tests the sample session factory
func TestSampleSession(t *testing.T) {
	session := SampleSession("sess1", "device1")

	if session.ID != "sess1" {
		t.Errorf("Expected ID sess1, got %s", session.ID)
	}

	if session.DeviceID != "device1" {
		t.Errorf("Expected DeviceID device1, got %s", session.DeviceID)
	}

	if session.Status != "active" {
		t.Errorf("Expected status 'active', got %s", session.Status)
	}
}
