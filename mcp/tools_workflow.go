package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// registerWorkflowTools registers workflow management tools
func (s *MCPServer) registerWorkflowTools() {
	// workflow_list - List workflows
	s.server.AddTool(
		mcp.NewTool("workflow_list",
			mcp.WithDescription("List all saved workflows"),
		),
		s.handleWorkflowList,
	)

	// workflow_run - Run a workflow
	s.server.AddTool(
		mcp.NewTool("workflow_run",
			mcp.WithDescription("Run a workflow on a device"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID to run the workflow on"),
			),
			mcp.WithString("workflow_id",
				mcp.Required(),
				mcp.Description("Workflow ID to run"),
			),
		),
		s.handleWorkflowRun,
	)

	// workflow_stop - Stop a running workflow
	s.server.AddTool(
		mcp.NewTool("workflow_stop",
			mcp.WithDescription("Stop a running workflow on a device"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
		),
		s.handleWorkflowStop,
	)
}

// Tool handlers

func (s *MCPServer) handleWorkflowList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workflows, err := s.app.LoadWorkflows()
	if err != nil {
		return nil, fmt.Errorf("failed to load workflows: %w", err)
	}

	if len(workflows) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("No workflows found"),
			},
		}, nil
	}

	result := fmt.Sprintf("Found %d workflow(s):\n\n", len(workflows))
	for i, wf := range workflows {
		result += fmt.Sprintf("%d. %s (ID: %s)\n", i+1, wf.Name, wf.ID)
		if wf.Description != "" {
			result += fmt.Sprintf("   Description: %s\n", wf.Description)
		}
		result += fmt.Sprintf("   Steps: %d\n", len(wf.Steps))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}

func (s *MCPServer) handleWorkflowRun(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	workflowID, ok := args["workflow_id"].(string)
	if !ok || workflowID == "" {
		return nil, fmt.Errorf("workflow_id is required")
	}

	// Find the workflow
	workflows, err := s.app.LoadWorkflows()
	if err != nil {
		return nil, fmt.Errorf("failed to load workflows: %w", err)
	}

	var targetWorkflow *Workflow
	for _, wf := range workflows {
		if wf.ID == workflowID {
			wfCopy := wf
			targetWorkflow = &wfCopy
			break
		}
	}

	if targetWorkflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Get device info to construct Device struct
	devices, err := s.app.GetDevices(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	var targetDevice *Device
	for _, d := range devices {
		if d.ID == deviceID {
			dCopy := d
			targetDevice = &dCopy
			break
		}
	}

	if targetDevice == nil {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	// Run the workflow in a goroutine (non-blocking)
	go func() {
		err := s.app.RunWorkflow(*targetDevice, *targetWorkflow)
		if err != nil {
			fmt.Printf("[MCP] Workflow %s failed: %v\n", workflowID, err)
		}
	}()

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Started workflow '%s' on device %s\n\nWorkflow has %d steps and is running in background.", targetWorkflow.Name, deviceID, len(targetWorkflow.Steps))),
		},
	}, nil
}

func (s *MCPServer) handleWorkflowStop(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	// Get device info
	devices, err := s.app.GetDevices(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	var targetDevice *Device
	for _, d := range devices {
		if d.ID == deviceID {
			dCopy := d
			targetDevice = &dCopy
			break
		}
	}

	if targetDevice == nil {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	s.app.StopWorkflow(*targetDevice)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Stopped workflow on device %s", deviceID)),
		},
	}, nil
}

// Unused but kept for future use
var _ = json.Marshal
