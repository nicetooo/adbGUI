package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// registerDeviceTools registers device management tools
func (s *MCPServer) registerDeviceTools() {
	// device_list - List connected devices
	s.server.AddTool(
		mcp.NewTool("device_list",
			mcp.WithDescription("List all connected Android devices"),
		),
		s.handleDeviceList,
	)

	// device_info - Get device information
	s.server.AddTool(
		mcp.NewTool("device_info",
			mcp.WithDescription("Get detailed information about a specific device"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID to get information for"),
			),
		),
		s.handleDeviceInfo,
	)

	// device_connect - Connect to a wireless device
	s.server.AddTool(
		mcp.NewTool("device_connect",
			mcp.WithDescription("Connect to a device via ADB over network (IP:port)"),
			mcp.WithString("address",
				mcp.Required(),
				mcp.Description("Device address in format IP:port (e.g., 192.168.1.100:5555)"),
			),
		),
		s.handleDeviceConnect,
	)

	// device_disconnect - Disconnect a wireless device
	s.server.AddTool(
		mcp.NewTool("device_disconnect",
			mcp.WithDescription("Disconnect a device from ADB"),
			mcp.WithString("address",
				mcp.Required(),
				mcp.Description("Device address to disconnect"),
			),
		),
		s.handleDeviceDisconnect,
	)

	// device_pair - Pair with a device
	s.server.AddTool(
		mcp.NewTool("device_pair",
			mcp.WithDescription("Pair with a device using wireless debugging"),
			mcp.WithString("address",
				mcp.Required(),
				mcp.Description("Device pairing address (IP:port)"),
			),
			mcp.WithString("code",
				mcp.Required(),
				mcp.Description("6-digit pairing code from device"),
			),
		),
		s.handleDevicePair,
	)

	// device_wireless - Switch device to wireless mode
	s.server.AddTool(
		mcp.NewTool("device_wireless",
			mcp.WithDescription("Switch a USB-connected device to wireless ADB mode"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID to switch to wireless"),
			),
		),
		s.handleDeviceWireless,
	)

	// device_ip - Get device IP address
	s.server.AddTool(
		mcp.NewTool("device_ip",
			mcp.WithDescription("Get the IP address of a connected device"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID to get IP for"),
			),
		),
		s.handleDeviceIP,
	)

	// adb_execute - Execute arbitrary ADB command
	s.server.AddTool(
		mcp.NewTool("adb_execute",
			mcp.WithDescription("Execute an arbitrary ADB command on a device. Supports shell commands (e.g., 'shell pm list packages'), file operations (e.g., 'push local remote'), and other ADB commands."),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID to execute the command on"),
			),
			mcp.WithString("command",
				mcp.Required(),
				mcp.Description("ADB command to execute (e.g., 'shell ls /sdcard', 'shell pm list packages', 'shell getprop ro.build.version.sdk')"),
			),
			mcp.WithNumber("timeout",
				mcp.Description("Command timeout in seconds (default: 30, max: 300)"),
			),
		),
		s.handleAdbExecute,
	)
}

// Tool handlers

func (s *MCPServer) handleDeviceList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	devices, err := s.app.GetDevices(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	if len(devices) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("No devices connected"),
			},
		}, nil
	}

	// Format device list
	result := fmt.Sprintf("Found %d device(s):\n\n", len(devices))
	for i, d := range devices {
		connType := ""
		if d.Type == "wireless" || d.Type == "both" {
			connType = " [wireless]"
		}
		result += fmt.Sprintf("%d. %s (%s)%s\n   Model: %s, Brand: %s, State: %s\n",
			i+1, d.ID, d.Serial, connType, d.Model, d.Brand, d.State)
	}

	// Also include JSON for structured access
	jsonData, _ := json.MarshalIndent(devices, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
			mcp.NewTextContent(fmt.Sprintf("\nJSON data:\n```json\n%s\n```", string(jsonData))),
		},
	}, nil
}

func (s *MCPServer) handleDeviceInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	info, err := s.app.GetDeviceInfo(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	result := fmt.Sprintf("Device: %s\n\n", deviceID)
	result += fmt.Sprintf("Model: %s\n", info.Model)
	result += fmt.Sprintf("Brand: %s\n", info.Brand)
	result += fmt.Sprintf("Manufacturer: %s\n", info.Manufacturer)
	result += fmt.Sprintf("Android Version: %s\n", info.AndroidVer)
	result += fmt.Sprintf("SDK Level: %s\n", info.SDK)
	result += fmt.Sprintf("ABI: %s\n", info.ABI)
	result += fmt.Sprintf("Serial: %s\n", info.Serial)
	result += fmt.Sprintf("Resolution: %s\n", info.Resolution)
	result += fmt.Sprintf("Density: %s\n", info.Density)
	result += fmt.Sprintf("CPU: %s\n", info.CPU)
	result += fmt.Sprintf("Memory: %s\n", info.Memory)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}

func (s *MCPServer) handleDeviceConnect(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	address, ok := args["address"].(string)
	if !ok || address == "" {
		return nil, fmt.Errorf("address is required")
	}

	result, err := s.app.AdbConnect(address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}

func (s *MCPServer) handleDeviceDisconnect(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	address, ok := args["address"].(string)
	if !ok || address == "" {
		return nil, fmt.Errorf("address is required")
	}

	result, err := s.app.AdbDisconnect(address)
	if err != nil {
		return nil, fmt.Errorf("failed to disconnect: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}

func (s *MCPServer) handleDevicePair(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	address, ok := args["address"].(string)
	if !ok || address == "" {
		return nil, fmt.Errorf("address is required")
	}
	code, ok := args["code"].(string)
	if !ok || code == "" {
		return nil, fmt.Errorf("pairing code is required")
	}

	result, err := s.app.AdbPair(address, code)
	if err != nil {
		return nil, fmt.Errorf("failed to pair: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}

func (s *MCPServer) handleDeviceWireless(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	result, err := s.app.SwitchToWireless(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to switch to wireless: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}

func (s *MCPServer) handleDeviceIP(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	ip, err := s.app.GetDeviceIP(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device IP: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Device %s IP address: %s", deviceID, ip)),
		},
	}, nil
}

func (s *MCPServer) handleAdbExecute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	command, ok := args["command"].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf("command is required")
	}

	// Execute the ADB command
	output, err := s.app.RunAdbCommand(deviceID, command)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Command failed: %v\n\nOutput:\n%s", err, output)),
			},
			IsError: true,
		}, nil
	}

	// Format result
	result := fmt.Sprintf("Command: adb -s %s %s\n\n", deviceID, command)
	if output == "" {
		result += "Command executed successfully (no output)"
	} else {
		result += fmt.Sprintf("Output:\n%s", output)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}
