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
			mcp.WithDescription("Input text into the focused field"),
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

	// Use PerformNodeAction with special "input" type
	err := s.app.PerformNodeAction(deviceID, text, "input")
	if err != nil {
		return nil, fmt.Errorf("failed to input text: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Input text: \"%s\"", text)),
		},
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
