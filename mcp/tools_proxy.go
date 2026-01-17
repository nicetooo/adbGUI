package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// registerProxyTools registers proxy management tools
func (s *MCPServer) registerProxyTools() {
	// proxy_start - Start the proxy
	s.server.AddTool(
		mcp.NewTool("proxy_start",
			mcp.WithDescription("Start the HTTP/HTTPS proxy for network interception"),
			mcp.WithNumber("port",
				mcp.Description("Port to listen on (default: 8888)"),
			),
		),
		s.handleProxyStart,
	)

	// proxy_stop - Stop the proxy
	s.server.AddTool(
		mcp.NewTool("proxy_stop",
			mcp.WithDescription("Stop the HTTP/HTTPS proxy"),
		),
		s.handleProxyStop,
	)

	// proxy_status - Get proxy status
	s.server.AddTool(
		mcp.NewTool("proxy_status",
			mcp.WithDescription("Get the current proxy status"),
		),
		s.handleProxyStatus,
	)
}

// Tool handlers

func (s *MCPServer) handleProxyStart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	port := 8888
	if p, ok := args["port"].(float64); ok {
		port = int(p)
	}

	result, err := s.app.StartProxy(port)
	if err != nil {
		return nil, fmt.Errorf("failed to start proxy: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Proxy started on port %d\n%s", port, result)),
		},
	}, nil
}

func (s *MCPServer) handleProxyStop(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := s.app.StopProxy()
	if err != nil {
		return nil, fmt.Errorf("failed to stop proxy: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Proxy stopped\n%s", result)),
		},
	}, nil
}

func (s *MCPServer) handleProxyStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	running := s.app.GetProxyStatus()
	status := "stopped"
	if running {
		status = "running"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Proxy status: %s", status)),
		},
	}, nil
}
