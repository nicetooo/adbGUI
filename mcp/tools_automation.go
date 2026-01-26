package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// registerAutomationTools registers UI automation tools
func (s *MCPServer) registerAutomationTools() {
	// ui_hierarchy - Get UI hierarchy
	s.server.AddTool(
		mcp.NewTool("ui_hierarchy",
			mcp.WithDescription("Get the current UI hierarchy/layout of the device screen"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
		),
		s.handleUIHierarchy,
	)

	// ui_search - Search for UI elements
	s.server.AddTool(
		mcp.NewTool("ui_search",
			mcp.WithDescription("Search for UI elements matching a query"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Search query - can be text, resource-id, class name, or XPath"),
			),
		),
		s.handleUISearch,
	)

	// ui_tap - Tap on screen
	s.server.AddTool(
		mcp.NewTool("ui_tap",
			mcp.WithDescription("Tap at a specific location or element on the screen"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
			mcp.WithString("bounds",
				mcp.Required(),
				mcp.Description("Element bounds in format [x1,y1][x2,y2] or coordinates x,y"),
			),
		),
		s.handleUITap,
	)

	// ui_swipe - Swipe on screen
	s.server.AddTool(
		mcp.NewTool("ui_swipe",
			mcp.WithDescription("Perform a swipe gesture on the screen"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
			mcp.WithString("bounds",
				mcp.Required(),
				mcp.Description("Element bounds for swipe area"),
			),
			mcp.WithString("direction",
				mcp.Description("Swipe direction: up, down, left, right (default: down)"),
			),
		),
		s.handleUISwipe,
	)

	// ui_input - Input text
	s.server.AddTool(
		mcp.NewTool("ui_input",
			mcp.WithDescription(`Input text into the focused field.

Supports both ASCII and Unicode text (Chinese, Japanese, Korean, emoji, etc.).

For ASCII-only text, uses native 'adb shell input text' (fast, no setup needed).
For Unicode text, automatically installs ADBKeyboard on the device (first use only),
temporarily activates it for input, then restores the previous IME.
ADBKeyboard is a lightweight IME (~30KB) that enables Unicode input via ADB.

NOTE: First Unicode input on a device may take a few seconds for ADBKeyboard installation.
The field must already be focused before calling this tool. Use ui_tap to focus first if needed.`),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
			mcp.WithString("text",
				mcp.Required(),
				mcp.Description("Text to input"),
			),
		),
		s.handleUIInput,
	)

	// keyboard_setup - Install and activate ADBKeyboard
	s.server.AddTool(
		mcp.NewTool("keyboard_setup",
			mcp.WithDescription(`Pre-install ADBKeyboard on a device for Unicode text input support.

ADBKeyboard is a lightweight Android IME (~30KB) that enables inputting any Unicode text
(Chinese, Japanese, Korean, emoji, accented characters, etc.) via ADB commands.

This only installs and enables ADBKeyboard in the IME list. It does NOT switch the active IME.
ADBKeyboard is temporarily activated only during actual Unicode text input via ui_input,
and the previous IME is restored immediately after.

This is called automatically when Unicode text is first input via ui_input, but you can
call it proactively to avoid the install delay during actual input.`),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
		),
		s.handleKeyboardSetup,
	)

	// ui_resolution - Get device resolution
	s.server.AddTool(
		mcp.NewTool("ui_resolution",
			mcp.WithDescription("Get the device screen resolution"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
		),
		s.handleUIResolution,
	)
}

// Tool handlers

func (s *MCPServer) handleUIHierarchy(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	hierarchy, err := s.app.GetUIHierarchy(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get UI hierarchy: %w", err)
	}

	jsonData, err := json.MarshalIndent(hierarchy, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize hierarchy: %w", err)
	}

	// Truncate if too large
	result := string(jsonData)
	if len(result) > 50000 {
		result = result[:50000] + "\n... (truncated, hierarchy too large)"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("UI Hierarchy for device %s:\n\n```json\n%s\n```", deviceID, result)),
		},
	}, nil
}

func (s *MCPServer) handleUISearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	elements, err := s.app.SearchUIElements(deviceID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search UI elements: %w", err)
	}

	if len(elements) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("No elements found matching '%s'", query)),
			},
		}, nil
	}

	result := fmt.Sprintf("Found %d element(s) matching '%s':\n\n", len(elements), query)
	for i, elem := range elements {
		result += fmt.Sprintf("%d. ", i+1)
		if text, ok := elem["text"].(string); ok && text != "" {
			result += fmt.Sprintf("Text: \"%s\" ", text)
		}
		if class, ok := elem["class"].(string); ok {
			result += fmt.Sprintf("Class: %s ", class)
		}
		if bounds, ok := elem["bounds"].(string); ok {
			result += fmt.Sprintf("Bounds: %s", bounds)
		}
		result += "\n"
	}

	jsonData, _ := json.MarshalIndent(elements, "", "  ")
	result += fmt.Sprintf("\nJSON:\n```json\n%s\n```", string(jsonData))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}

func (s *MCPServer) handleUITap(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	bounds, ok := args["bounds"].(string)
	if !ok || bounds == "" {
		return nil, fmt.Errorf("bounds is required")
	}

	err := s.app.PerformNodeAction(deviceID, bounds, "click")
	if err != nil {
		return nil, fmt.Errorf("failed to tap: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Tapped at %s", bounds)),
		},
	}, nil
}

func (s *MCPServer) handleUISwipe(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	bounds, ok := args["bounds"].(string)
	if !ok || bounds == "" {
		return nil, fmt.Errorf("bounds is required")
	}

	direction := "down"
	if d, ok := args["direction"].(string); ok && d != "" {
		direction = d
	}

	actionType := "swipe_" + direction
	err := s.app.PerformNodeAction(deviceID, bounds, actionType)
	if err != nil {
		return nil, fmt.Errorf("failed to swipe: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Swiped %s at %s", direction, bounds)),
		},
	}, nil
}

func (s *MCPServer) handleUIInput(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	text, ok := args["text"].(string)
	if !ok || text == "" {
		return nil, fmt.Errorf("text is required")
	}

	// Use InputText: supports both ASCII and Unicode via ADBKeyboard
	err := s.app.InputText(deviceID, text)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error: %v", err))},
			IsError: true,
		}, nil
	}

	// Indicate which method was used
	method := "adb input text"
	for _, r := range text {
		if r > 127 {
			method = "ADBKeyboard (base64)"
			break
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Input text: \"%s\" (method: %s)", text, method)),
		},
	}, nil
}

func (s *MCPServer) handleKeyboardSetup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	ready, installed, err := s.app.EnsureADBKeyboard(deviceID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error setting up ADBKeyboard: %v", err))},
			IsError: true,
		}, nil
	}

	var msg string
	if installed {
		msg = fmt.Sprintf("ADBKeyboard installed and enabled on device %s. It will be temporarily activated during Unicode text input via ui_input.", deviceID)
	} else if ready {
		msg = fmt.Sprintf("ADBKeyboard is already installed on device %s. Ready for Unicode text input.", deviceID)
	} else {
		msg = fmt.Sprintf("ADBKeyboard setup on device %s: ready=%v", deviceID, ready)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(msg)},
	}, nil
}

func (s *MCPServer) handleUIResolution(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	resolution, err := s.app.GetDeviceResolution(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resolution: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Device %s resolution: %s", deviceID, resolution)),
		},
	}, nil
}
