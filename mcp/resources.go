package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleDevicesResource handles the gaze://devices resource
func (s *MCPServer) handleDevicesResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	devices, err := s.app.GetDevices(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	jsonData, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize devices: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// handleDeviceInfoResource handles the gaze://devices/{deviceId} resource template
func (s *MCPServer) handleDeviceInfoResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Extract device ID from URI: gaze://devices/{deviceId}
	uri := request.Params.URI
	parts := strings.Split(uri, "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid URI format: %s", uri)
	}
	deviceID := parts[3]

	info, err := s.app.GetDeviceInfo(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize device info: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// handleSessionsResource handles the gaze://sessions resource
func (s *MCPServer) handleSessionsResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	sessions, err := s.app.ListStoredSessions("", 50)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	jsonData, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize sessions: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// handleWorkflowsResource handles the workflow://list resource
func (s *MCPServer) handleWorkflowsResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	workflows, err := s.app.LoadWorkflows()
	if err != nil {
		return nil, fmt.Errorf("failed to load workflows: %w", err)
	}

	jsonData, err := json.MarshalIndent(workflows, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize workflows: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// handleWorkflowResource handles the workflow://{workflowId} resource template
func (s *MCPServer) handleWorkflowResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Extract workflow ID from URI: workflow://{workflowId}
	uri := request.Params.URI
	// URI format: workflow://wf_123456
	workflowID := strings.TrimPrefix(uri, "workflow://")
	if workflowID == "" || workflowID == uri {
		return nil, fmt.Errorf("invalid workflow URI format: %s", uri)
	}

	// Load all workflows and find the matching one
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

	jsonData, err := json.MarshalIndent(targetWorkflow, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize workflow: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}
