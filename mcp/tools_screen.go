package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// registerScreenTools registers screen control tools
func (s *MCPServer) registerScreenTools() {
	// screen_screenshot - Take a screenshot
	s.server.AddTool(
		mcp.NewTool("screen_screenshot",
			mcp.WithDescription("Take a screenshot of the device screen"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
			mcp.WithString("save_path",
				mcp.Description("Path to save the screenshot (optional, auto-generated if not provided)"),
			),
		),
		s.handleScreenshot,
	)

	// screen_record_start - Start recording
	s.server.AddTool(
		mcp.NewTool("screen_record_start",
			mcp.WithDescription("Start recording the device screen"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
			mcp.WithNumber("max_size",
				mcp.Description("Maximum video dimension (default: 1280)"),
			),
			mcp.WithNumber("bit_rate",
				mcp.Description("Video bit rate in Mbps (default: 8)"),
			),
		),
		s.handleRecordStart,
	)

	// screen_record_stop - Stop recording
	s.server.AddTool(
		mcp.NewTool("screen_record_stop",
			mcp.WithDescription("Stop recording the device screen"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
		),
		s.handleRecordStop,
	)

	// screen_recording_status - Check recording status
	s.server.AddTool(
		mcp.NewTool("screen_recording_status",
			mcp.WithDescription("Check if device screen is being recorded"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
		),
		s.handleRecordingStatus,
	)
}

// Tool handlers

func (s *MCPServer) handleScreenshot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	savePath := ""
	if p, ok := args["save_path"].(string); ok && p != "" {
		savePath = p
	} else {
		// Auto-generate save path in Downloads folder
		home, err := os.UserHomeDir()
		if err != nil {
			home = os.TempDir()
		}
		downloadsDir := filepath.Join(home, "Downloads")
		if _, err := os.Stat(downloadsDir); os.IsNotExist(err) {
			downloadsDir = home
		}
		filename := fmt.Sprintf("screenshot_%s_%s.png", deviceID, time.Now().Format("20060102_150405"))
		savePath = filepath.Join(downloadsDir, filename)
	}

	path, err := s.app.TakeScreenshot(deviceID, savePath)
	if err != nil {
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Screenshot saved to: %s", path)),
		},
	}, nil
}

func (s *MCPServer) handleRecordStart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	config := ScrcpyConfig{
		MaxSize: 1280,
		BitRate: 8000000,
		MaxFps:  30,
	}

	if maxSize, ok := args["max_size"].(float64); ok {
		config.MaxSize = int(maxSize)
	}
	if bitRate, ok := args["bit_rate"].(float64); ok {
		config.BitRate = int(bitRate * 1000000) // Convert Mbps to bps
	}

	err := s.app.StartRecording(deviceID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to start recording: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Started recording device %s", deviceID)),
		},
	}, nil
}

func (s *MCPServer) handleRecordStop(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	err := s.app.StopRecording(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to stop recording: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Stopped recording device %s", deviceID)),
		},
	}, nil
}

func (s *MCPServer) handleRecordingStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	isRecording := s.app.IsRecording(deviceID)
	status := "not recording"
	if isRecording {
		status = "recording"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Device %s is %s", deviceID, status)),
		},
	}, nil
}
