package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
			mcp.WithDescription(`Take a screenshot of the device screen and return as base64 image.
Optionally includes UI hierarchy XML for element analysis.
Returns: base64 PNG image + optional UI hierarchy JSON`),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
			mcp.WithBoolean("include_ui",
				mcp.Description("Include UI hierarchy in response (default: false)"),
			),
			mcp.WithString("save_path",
				mcp.Description("Also save screenshot to this path (optional)"),
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

	includeUI := false
	if v, ok := args["include_ui"].(bool); ok {
		includeUI = v
	}

	// Generate temp path for screenshot
	tempDir := os.TempDir()
	filename := fmt.Sprintf("screenshot_%s_%s.png", deviceID, time.Now().Format("20060102_150405"))
	tempPath := filepath.Join(tempDir, filename)

	// Take screenshot
	path, err := s.app.TakeScreenshot(deviceID, tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	// Ensure temp file is always cleaned up immediately after we're done
	defer os.Remove(path)

	// Read screenshot file and convert to base64
	imageData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot: %w", err)
	}
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// Also save to user-specified path if provided
	savedPath := ""
	if savePath, ok := args["save_path"].(string); ok && savePath != "" {
		if err := os.WriteFile(savePath, imageData, 0644); err == nil {
			savedPath = savePath
		}
	}

	// Build response content
	contents := []mcp.Content{}

	// Add image content
	contents = append(contents, mcp.NewImageContent(base64Image, "image/png"))

	// Build text description
	textInfo := fmt.Sprintf("Screenshot captured for device %s", deviceID)
	if savedPath != "" {
		textInfo += fmt.Sprintf("\nSaved to: %s", savedPath)
	}

	// Include UI hierarchy if requested
	if includeUI {
		hierarchy, err := s.app.GetUIHierarchy(deviceID)
		if err == nil {
			jsonData, err := json.Marshal(hierarchy)
			if err == nil {
				textInfo += fmt.Sprintf("\n\nUI Hierarchy:\n```json\n%s\n```", string(jsonData))
			}
		} else {
			textInfo += fmt.Sprintf("\n\nUI Hierarchy: failed to get (%v)", err)
		}
	}

	contents = append(contents, mcp.NewTextContent(textInfo))

	return &mcp.CallToolResult{
		Content: contents,
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
