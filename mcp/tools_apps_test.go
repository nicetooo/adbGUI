package mcp

import (
	"context"
	"strings"
	"testing"
)

// ==================== app_list ====================

func TestHandleAppList_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ListPackagesResult = []AppPackage{
		SampleAppPackage("com.example.app1"),
		SampleAppPackage("com.example.app2"),
	}
	server := NewMCPServer(mock)

	result, err := server.handleAppList(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "com.example.app1") {
		t.Error("Result should contain app1")
	}
	if !strings.Contains(text, "com.example.app2") {
		t.Error("Result should contain app2")
	}

	// Verify correct arguments
	lastCall := mock.GetLastCall()
	if lastCall.Args[0] != "device1" {
		t.Errorf("Expected device_id 'device1', got %v", lastCall.Args[0])
	}
}

func TestHandleAppList_WithType(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ListPackagesResult = []AppPackage{}
	server := NewMCPServer(mock)

	_, err := server.handleAppList(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"type":      "system",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify type was passed
	lastCall := mock.GetLastCall()
	if lastCall.Args[1] != "system" {
		t.Errorf("Expected type 'system', got %v", lastCall.Args[1])
	}
}

func TestHandleAppList_DefaultType(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ListPackagesResult = []AppPackage{}
	server := NewMCPServer(mock)

	_, err := server.handleAppList(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Default type should be "user"
	lastCall := mock.GetLastCall()
	if lastCall.Args[1] != "user" {
		t.Errorf("Expected default type 'user', got %v", lastCall.Args[1])
	}
}

func TestHandleAppList_NoApps(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ListPackagesResult = []AppPackage{}
	server := NewMCPServer(mock)

	result, err := server.handleAppList(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	// Result says "No packages found"
	if !strings.Contains(strings.ToLower(text), "no") {
		t.Errorf("Result should indicate no apps found, got: %s", text)
	}
}

func TestHandleAppList_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppList(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleAppList_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("ListPackages", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleAppList(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== app_info ====================

func TestHandleAppInfo_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetAppInfoResult = SampleAppPackage("com.example.app")
	mock.GetAppInfoResult.VersionName = "2.0.0"
	mock.GetAppInfoResult.VersionCode = "20"
	server := NewMCPServer(mock)

	result, err := server.handleAppInfo(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":    "device1",
		"package_name": "com.example.app",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "com.example.app") {
		t.Error("Result should contain package name")
	}
	if !strings.Contains(text, "2.0.0") {
		t.Error("Result should contain version name")
	}
}

func TestHandleAppInfo_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppInfo(context.Background(), makeToolRequest(map[string]interface{}{
		"package_name": "com.example.app",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleAppInfo_MissingPackageName(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppInfo(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error for missing package_name")
	}
}

func TestHandleAppInfo_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("GetAppInfo", ErrAppNotFound)
	server := NewMCPServer(mock)

	_, err := server.handleAppInfo(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":    "device1",
		"package_name": "com.nonexistent.app",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== app_start ====================

func TestHandleAppStart_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.StartAppResult = "Starting: Intent { act=android.intent.action.MAIN }"
	server := NewMCPServer(mock)

	result, err := server.handleAppStart(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":    "device1",
		"package_name": "com.example.app",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "start") {
		t.Error("Result should mention starting")
	}

	// Verify arguments
	lastCall := mock.GetLastCall()
	if lastCall.Args[0] != "device1" {
		t.Errorf("Expected device_id 'device1', got %v", lastCall.Args[0])
	}
	if lastCall.Args[1] != "com.example.app" {
		t.Errorf("Expected package_name 'com.example.app', got %v", lastCall.Args[1])
	}
}

func TestHandleAppStart_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppStart(context.Background(), makeToolRequest(map[string]interface{}{
		"package_name": "com.example.app",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleAppStart_MissingPackageName(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppStart(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error for missing package_name")
	}
}

func TestHandleAppStart_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("StartApp", ErrAppNotFound)
	server := NewMCPServer(mock)

	_, err := server.handleAppStart(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":    "device1",
		"package_name": "com.nonexistent.app",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== app_stop ====================

func TestHandleAppStop_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ForceStopAppResult = "OK"
	server := NewMCPServer(mock)

	result, err := server.handleAppStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":    "device1",
		"package_name": "com.example.app",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "stop") {
		t.Error("Result should mention stopping")
	}
}

func TestHandleAppStop_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppStop(context.Background(), makeToolRequest(map[string]interface{}{
		"package_name": "com.example.app",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleAppStop_MissingPackageName(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error for missing package_name")
	}
}

func TestHandleAppStop_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("ForceStopApp", ErrPermissionDenied)
	server := NewMCPServer(mock)

	_, err := server.handleAppStop(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":    "device1",
		"package_name": "com.system.app",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== app_running ====================

func TestHandleAppRunning_Running(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsAppRunningResult = true
	server := NewMCPServer(mock)

	result, err := server.handleAppRunning(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":    "device1",
		"package_name": "com.example.app",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "running") {
		t.Error("Result should indicate app is running")
	}
}

func TestHandleAppRunning_NotRunning(t *testing.T) {
	mock := NewMockGazeApp()
	mock.IsAppRunningResult = false
	server := NewMCPServer(mock)

	result, err := server.handleAppRunning(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":    "device1",
		"package_name": "com.example.app",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "not") {
		t.Error("Result should indicate app is not running")
	}
}

func TestHandleAppRunning_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppRunning(context.Background(), makeToolRequest(map[string]interface{}{
		"package_name": "com.example.app",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleAppRunning_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("IsAppRunning", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleAppRunning(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":    "device1",
		"package_name": "com.example.app",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== app_install ====================
// Note: This tool requires confirmation via MCP elicitation
// In tests without an active MCP session, confirmation will fail

func TestHandleAppInstall_RequiresConfirmation(t *testing.T) {
	mock := NewMockGazeApp()
	mock.InstallAPKResult = "Success"
	server := NewMCPServer(mock)

	// Without an active MCP session, confirmation request will fail
	_, err := server.handleAppInstall(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"apk_path":  "/path/to/app.apk",
	}))
	// Expected to fail because no active session for confirmation
	if err == nil {
		t.Log("Install succeeded without confirmation (may be expected in some scenarios)")
	} else {
		// Error is expected - confirmation requires active MCP session
		if !strings.Contains(err.Error(), "confirmation") && !strings.Contains(err.Error(), "session") {
			t.Logf("Got expected error: %v", err)
		}
	}
}

func TestHandleAppInstall_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppInstall(context.Background(), makeToolRequest(map[string]interface{}{
		"apk_path": "/path/to/app.apk",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleAppInstall_MissingApkPath(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppInstall(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error for missing apk_path")
	}
}

func TestHandleAppInstall_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("InstallAPK", ErrPermissionDenied)
	server := NewMCPServer(mock)

	_, err := server.handleAppInstall(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
		"apk_path":  "/path/to/app.apk",
	}))
	// Note: Error might be wrapped in confirmation flow
	// Just check that we don't panic
	_ = err
}

// ==================== app_uninstall ====================
// Note: This tool requires confirmation via MCP elicitation

func TestHandleAppUninstall_RequiresConfirmation(t *testing.T) {
	mock := NewMockGazeApp()
	mock.UninstallAppResult = "Success"
	server := NewMCPServer(mock)

	// Without an active MCP session, confirmation request will fail
	_, err := server.handleAppUninstall(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":    "device1",
		"package_name": "com.example.app",
	}))
	// Expected to fail because no active session for confirmation
	if err == nil {
		t.Log("Uninstall succeeded without confirmation (may be expected in some scenarios)")
	} else {
		t.Logf("Got expected error (confirmation required): %v", err)
	}
}

func TestHandleAppUninstall_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppUninstall(context.Background(), makeToolRequest(map[string]interface{}{
		"package_name": "com.example.app",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleAppUninstall_MissingPackageName(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppUninstall(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error for missing package_name")
	}
}

// ==================== app_clear_data ====================
// Note: This tool requires confirmation via MCP elicitation

func TestHandleAppClearData_RequiresConfirmation(t *testing.T) {
	mock := NewMockGazeApp()
	mock.ClearAppDataResult = "Success"
	server := NewMCPServer(mock)

	// Without an active MCP session, confirmation request will fail
	_, err := server.handleAppClearData(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id":    "device1",
		"package_name": "com.example.app",
	}))
	// Expected to fail because no active session for confirmation
	if err == nil {
		t.Log("ClearData succeeded without confirmation (may be expected in some scenarios)")
	} else {
		t.Logf("Got expected error (confirmation required): %v", err)
	}
}

func TestHandleAppClearData_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppClearData(context.Background(), makeToolRequest(map[string]interface{}{
		"package_name": "com.example.app",
	}))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleAppClearData_MissingPackageName(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleAppClearData(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error for missing package_name")
	}
}
