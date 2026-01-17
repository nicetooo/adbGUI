package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Helper to create a CallToolRequest with arguments
func makeToolRequest(args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

// Helper to get text content from result
func getTextContent(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}

// ==================== device_list ====================

func TestHandleDeviceList_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithDevices(
		SampleDevice("device1"),
		SampleDevice("device2"),
	)
	server := NewMCPServer(mock)

	result, err := server.handleDeviceList(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "device1") {
		t.Error("Result should contain device1")
	}
	if !strings.Contains(text, "device2") {
		t.Error("Result should contain device2")
	}
	if !strings.Contains(text, "2 device") {
		t.Error("Result should mention 2 devices")
	}
}

func TestHandleDeviceList_NoDevices(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	result, err := server.handleDeviceList(context.Background(), makeToolRequest(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "no device") {
		t.Errorf("Result should indicate no devices, got: %s", text)
	}
}

func TestHandleDeviceList_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("GetDevices", ErrDeviceNotFound)
	server := NewMCPServer(mock)

	_, err := server.handleDeviceList(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== device_info ====================

func TestHandleDeviceInfo_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetDeviceInfoResult = SampleDeviceInfo()
	server := NewMCPServer(mock)

	result, err := server.handleDeviceInfo(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "Pixel 6") {
		t.Error("Result should contain model name")
	}
	if !strings.Contains(text, "1080x2400") {
		t.Error("Result should contain resolution")
	}

	// Verify correct device ID was passed
	if !mock.WasMethodCalled("GetDeviceInfo") {
		t.Error("GetDeviceInfo should have been called")
	}
	lastCall := mock.GetLastCall()
	if lastCall.Args[0] != "device1" {
		t.Errorf("Expected device_id 'device1', got %v", lastCall.Args[0])
	}
}

func TestHandleDeviceInfo_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleDeviceInfo(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
	if !strings.Contains(err.Error(), "device_id") {
		t.Errorf("Error should mention device_id, got: %v", err)
	}
}

func TestHandleDeviceInfo_EmptyDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleDeviceInfo(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "",
	}))
	if err == nil {
		t.Error("Expected error for empty device_id")
	}
}

func TestHandleDeviceInfo_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("GetDeviceInfo", ErrDeviceNotFound)
	server := NewMCPServer(mock)

	_, err := server.handleDeviceInfo(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "nonexistent",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== device_connect ====================

func TestHandleDeviceConnect_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.AdbConnectResult = "connected to 192.168.1.100:5555"
	server := NewMCPServer(mock)

	result, err := server.handleDeviceConnect(context.Background(), makeToolRequest(map[string]interface{}{
		"address": "192.168.1.100:5555",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "connected") {
		t.Error("Result should indicate connection success")
	}

	// Verify correct address was passed
	lastCall := mock.GetLastCall()
	if lastCall.Method != "AdbConnect" {
		t.Errorf("Expected AdbConnect call, got %s", lastCall.Method)
	}
	if lastCall.Args[0] != "192.168.1.100:5555" {
		t.Errorf("Expected address '192.168.1.100:5555', got %v", lastCall.Args[0])
	}
}

func TestHandleDeviceConnect_MissingAddress(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleDeviceConnect(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing address")
	}
}

func TestHandleDeviceConnect_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("AdbConnect", ErrTimeout)
	server := NewMCPServer(mock)

	_, err := server.handleDeviceConnect(context.Background(), makeToolRequest(map[string]interface{}{
		"address": "192.168.1.100:5555",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== device_disconnect ====================

func TestHandleDeviceDisconnect_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.AdbDisconnectResult = "disconnected 192.168.1.100:5555"
	server := NewMCPServer(mock)

	result, err := server.handleDeviceDisconnect(context.Background(), makeToolRequest(map[string]interface{}{
		"address": "192.168.1.100:5555",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "disconnected") {
		t.Error("Result should indicate disconnection")
	}
}

func TestHandleDeviceDisconnect_MissingAddress(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleDeviceDisconnect(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing address")
	}
}

func TestHandleDeviceDisconnect_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("AdbDisconnect", ErrDeviceNotFound)
	server := NewMCPServer(mock)

	_, err := server.handleDeviceDisconnect(context.Background(), makeToolRequest(map[string]interface{}{
		"address": "192.168.1.100:5555",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== device_pair ====================

func TestHandleDevicePair_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.AdbPairResult = "Successfully paired"
	server := NewMCPServer(mock)

	result, err := server.handleDevicePair(context.Background(), makeToolRequest(map[string]interface{}{
		"address": "192.168.1.100:37123",
		"code":    "123456",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(strings.ToLower(text), "pair") {
		t.Error("Result should mention pairing")
	}

	// Verify both arguments were passed
	lastCall := mock.GetLastCall()
	if lastCall.Args[0] != "192.168.1.100:37123" {
		t.Errorf("Expected address '192.168.1.100:37123', got %v", lastCall.Args[0])
	}
	if lastCall.Args[1] != "123456" {
		t.Errorf("Expected code '123456', got %v", lastCall.Args[1])
	}
}

func TestHandleDevicePair_MissingAddress(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleDevicePair(context.Background(), makeToolRequest(map[string]interface{}{
		"code": "123456",
	}))
	if err == nil {
		t.Error("Expected error for missing address")
	}
}

func TestHandleDevicePair_MissingCode(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleDevicePair(context.Background(), makeToolRequest(map[string]interface{}{
		"address": "192.168.1.100:37123",
	}))
	if err == nil {
		t.Error("Expected error for missing code")
	}
}

func TestHandleDevicePair_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("AdbPair", ErrTimeout)
	server := NewMCPServer(mock)

	_, err := server.handleDevicePair(context.Background(), makeToolRequest(map[string]interface{}{
		"address": "192.168.1.100:37123",
		"code":    "123456",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== device_wireless ====================

func TestHandleDeviceWireless_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SwitchToWirelessResult = "192.168.1.100:5555"
	server := NewMCPServer(mock)

	result, err := server.handleDeviceWireless(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "192.168.1.100") {
		t.Error("Result should contain the wireless address")
	}
}

func TestHandleDeviceWireless_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleDeviceWireless(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleDeviceWireless_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("SwitchToWireless", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleDeviceWireless(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ==================== device_ip ====================

func TestHandleDeviceIP_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetDeviceIPResult = "192.168.1.100"
	server := NewMCPServer(mock)

	result, err := server.handleDeviceIP(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	text := getTextContent(result)
	if !strings.Contains(text, "192.168.1.100") {
		t.Error("Result should contain the IP address")
	}
}

func TestHandleDeviceIP_MissingDeviceId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	_, err := server.handleDeviceIP(context.Background(), makeToolRequest(nil))
	if err == nil {
		t.Error("Expected error for missing device_id")
	}
}

func TestHandleDeviceIP_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.SetupWithError("GetDeviceIP", ErrDeviceOffline)
	server := NewMCPServer(mock)

	_, err := server.handleDeviceIP(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestHandleDeviceIP_EmptyResult(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetDeviceIPResult = ""
	server := NewMCPServer(mock)

	result, err := server.handleDeviceIP(context.Background(), makeToolRequest(map[string]interface{}{
		"device_id": "device1",
	}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should still return a result, possibly indicating no IP found
	if result == nil {
		t.Error("Result should not be nil")
	}
}
