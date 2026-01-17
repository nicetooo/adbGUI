package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
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

	// workflow_get - Get workflow details
	s.server.AddTool(
		mcp.NewTool("workflow_get",
			mcp.WithDescription("Get detailed information about a specific workflow including all steps"),
			mcp.WithString("workflow_id",
				mcp.Required(),
				mcp.Description("Workflow ID"),
			),
		),
		s.handleWorkflowGet,
	)

	// workflow_create - Create a new workflow
	s.server.AddTool(
		mcp.NewTool("workflow_create",
			mcp.WithDescription("Create a new workflow with the given name and steps"),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Workflow name"),
			),
			mcp.WithString("description",
				mcp.Description("Workflow description"),
			),
			mcp.WithString("steps_json",
				mcp.Description("JSON array of workflow steps. Each step should have: type (tap/swipe/input/wait/back/home/launch), and optionally: selector (with type and value), value, timeout, postDelay"),
			),
		),
		s.handleWorkflowCreate,
	)

	// workflow_delete - Delete a workflow
	s.server.AddTool(
		mcp.NewTool("workflow_delete",
			mcp.WithDescription("Delete a saved workflow (requires confirmation)"),
			mcp.WithString("workflow_id",
				mcp.Required(),
				mcp.Description("Workflow ID to delete"),
			),
		),
		s.handleWorkflowDelete,
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

	// workflow_status - Get workflow running status
	s.server.AddTool(
		mcp.NewTool("workflow_status",
			mcp.WithDescription("Check if a workflow is currently running on a device"),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
		),
		s.handleWorkflowStatus,
	)

	// workflow_execute_step - Execute a single workflow step
	s.server.AddTool(
		mcp.NewTool("workflow_execute_step",
			mcp.WithDescription("Execute a single workflow step on a device. Useful for building workflows step by step or testing individual actions."),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device ID"),
			),
			mcp.WithString("step_type",
				mcp.Required(),
				mcp.Description("Step type: tap, swipe, input, wait, back, home, launch, screenshot"),
			),
			mcp.WithString("selector_type",
				mcp.Description("Element selector type: resourceId, text, contentDesc, className, xpath (required for tap/swipe/input on elements)"),
			),
			mcp.WithString("selector_value",
				mcp.Description("Element selector value (required when selector_type is specified)"),
			),
			mcp.WithString("value",
				mcp.Description("Input value for 'input' type, package name for 'launch' type, or duration(ms) for 'wait' type"),
			),
			mcp.WithNumber("timeout",
				mcp.Description("Timeout in milliseconds for finding elements (default: 5000)"),
			),
			mcp.WithNumber("post_delay",
				mcp.Description("Delay in milliseconds after executing the step (default: 500)"),
			),
		),
		s.handleWorkflowExecuteStep,
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
		if wf.CreatedAt != "" {
			result += fmt.Sprintf("   Created: %s\n", wf.CreatedAt)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}

func (s *MCPServer) handleWorkflowGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	workflowID, ok := args["workflow_id"].(string)
	if !ok || workflowID == "" {
		return nil, fmt.Errorf("workflow_id is required")
	}

	workflow, err := s.app.GetWorkflow(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	result := fmt.Sprintf("Workflow: %s\n", workflow.Name)
	result += fmt.Sprintf("ID: %s\n", workflow.ID)
	if workflow.Description != "" {
		result += fmt.Sprintf("Description: %s\n", workflow.Description)
	}
	if workflow.CreatedAt != "" {
		result += fmt.Sprintf("Created: %s\n", workflow.CreatedAt)
	}
	if workflow.UpdatedAt != "" {
		result += fmt.Sprintf("Updated: %s\n", workflow.UpdatedAt)
	}
	result += fmt.Sprintf("\nSteps (%d):\n", len(workflow.Steps))

	for i, step := range workflow.Steps {
		result += fmt.Sprintf("\n%d. [%s] %s\n", i+1, step.Type, step.Name)
		if step.Selector != nil {
			result += fmt.Sprintf("   Selector: %s = %s\n", step.Selector.Type, step.Selector.Value)
		}
		if step.Value != "" {
			result += fmt.Sprintf("   Value: %s\n", step.Value)
		}
		if step.Timeout > 0 {
			result += fmt.Sprintf("   Timeout: %dms\n", step.Timeout)
		}
		if step.PostDelay > 0 {
			result += fmt.Sprintf("   Post Delay: %dms\n", step.PostDelay)
		}
	}

	// Also return JSON for programmatic access
	jsonData, _ := json.MarshalIndent(workflow, "", "  ")
	result += fmt.Sprintf("\n\nJSON:\n%s", string(jsonData))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}

func (s *MCPServer) handleWorkflowCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name is required")
	}

	description := ""
	if d, ok := args["description"].(string); ok {
		description = d
	}

	// Parse steps from JSON
	var steps []WorkflowStep
	if stepsJSON, ok := args["steps_json"].(string); ok && stepsJSON != "" {
		if err := json.Unmarshal([]byte(stepsJSON), &steps); err != nil {
			return nil, fmt.Errorf("invalid steps_json format: %w", err)
		}
	}

	// Generate IDs for steps that don't have one
	for i := range steps {
		if steps[i].ID == "" {
			steps[i].ID = fmt.Sprintf("step_%s", uuid.New().String()[:8])
		}
	}

	// Create workflow
	now := time.Now().Format(time.RFC3339)
	workflow := Workflow{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Steps:       steps,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.app.SaveWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("failed to save workflow: %w", err)
	}

	result := fmt.Sprintf("Created workflow '%s'\n", name)
	result += fmt.Sprintf("ID: %s\n", workflow.ID)
	result += fmt.Sprintf("Steps: %d\n", len(steps))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}

func (s *MCPServer) handleWorkflowDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	workflowID, ok := args["workflow_id"].(string)
	if !ok || workflowID == "" {
		return nil, fmt.Errorf("workflow_id is required")
	}

	// Get workflow info first
	workflow, err := s.app.GetWorkflow(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	if err := s.app.DeleteWorkflow(workflowID); err != nil {
		return nil, fmt.Errorf("failed to delete workflow: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Deleted workflow '%s' (ID: %s)", workflow.Name, workflowID)),
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

	// Check if already running
	if s.app.IsWorkflowRunning(deviceID) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("A workflow is already running on device %s. Use workflow_stop to stop it first.", deviceID)),
			},
		}, nil
	}

	// Find the workflow
	workflow, err := s.app.GetWorkflow(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
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
		err := s.app.RunWorkflow(*targetDevice, *workflow)
		if err != nil {
			fmt.Printf("[MCP] Workflow %s failed: %v\n", workflowID, err)
		}
	}()

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Started workflow '%s' on device %s\n\nWorkflow has %d steps and is running in background.\n\nUse workflow_status to check progress or workflow_stop to stop.", workflow.Name, deviceID, len(workflow.Steps))),
		},
	}, nil
}

func (s *MCPServer) handleWorkflowStop(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	// Check if running
	if !s.app.IsWorkflowRunning(deviceID) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("No workflow is running on device %s", deviceID)),
			},
		}, nil
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

func (s *MCPServer) handleWorkflowStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	isRunning := s.app.IsWorkflowRunning(deviceID)

	status := "idle"
	if isRunning {
		status = "running"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Device %s workflow status: %s", deviceID, status)),
		},
	}, nil
}

func (s *MCPServer) handleWorkflowExecuteStep(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceID, ok := args["device_id"].(string)
	if !ok || deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	stepType, ok := args["step_type"].(string)
	if !ok || stepType == "" {
		return nil, fmt.Errorf("step_type is required")
	}

	// Validate step type
	validTypes := map[string]bool{
		"tap": true, "swipe": true, "input": true, "wait": true,
		"back": true, "home": true, "launch": true, "screenshot": true,
	}
	if !validTypes[stepType] {
		return nil, fmt.Errorf("invalid step_type '%s'. Valid types: tap, swipe, input, wait, back, home, launch, screenshot", stepType)
	}

	// Build the step
	step := WorkflowStep{
		ID:   fmt.Sprintf("mcp_step_%s", uuid.New().String()[:8]),
		Type: stepType,
		Name: fmt.Sprintf("MCP %s step", stepType),
	}

	// Handle selector
	if selectorType, ok := args["selector_type"].(string); ok && selectorType != "" {
		selectorValue, _ := args["selector_value"].(string)
		if selectorValue == "" {
			return nil, fmt.Errorf("selector_value is required when selector_type is specified")
		}
		step.Selector = &ElementSelector{
			Type:  selectorType,
			Value: selectorValue,
		}
	}

	// Handle value
	if value, ok := args["value"].(string); ok {
		step.Value = value
	}

	// Handle timeout
	step.Timeout = 5000 // default
	if timeout, ok := args["timeout"].(float64); ok && timeout > 0 {
		step.Timeout = int(timeout)
	}

	// Handle post delay
	step.PostDelay = 500 // default
	if postDelay, ok := args["post_delay"].(float64); ok && postDelay >= 0 {
		step.PostDelay = int(postDelay)
	}

	// Special handling for 'wait' type
	if stepType == "wait" && step.Value != "" {
		if duration, err := strconv.Atoi(step.Value); err == nil {
			step.Timeout = duration
		}
	}

	// Execute the step
	startTime := time.Now()
	err := s.app.ExecuteSingleWorkflowStep(deviceID, step)
	duration := time.Since(startTime)

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Step execution failed after %dms: %v", duration.Milliseconds(), err)),
			},
		}, nil
	}

	result := fmt.Sprintf("Step executed successfully in %dms\n", duration.Milliseconds())
	result += fmt.Sprintf("Type: %s\n", stepType)
	if step.Selector != nil {
		result += fmt.Sprintf("Selector: %s = %s\n", step.Selector.Type, step.Selector.Value)
	}
	if step.Value != "" {
		result += fmt.Sprintf("Value: %s\n", step.Value)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(result),
		},
	}, nil
}
