package mcp

import (
	"context"
	"encoding/json"
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

	// mock_rule_list - List all mock rules
	s.server.AddTool(
		mcp.NewTool("mock_rule_list",
			mcp.WithDescription(`List all HTTP mock response rules.

Returns all configured mock rules with their URL patterns, methods, status codes,
response bodies, delays, and enabled states. Rules are sorted by creation time.`),
		),
		s.handleMockRuleList,
	)

	// mock_rule_add - Add a mock rule
	s.server.AddTool(
		mcp.NewTool("mock_rule_add",
			mcp.WithDescription(`Add a new HTTP mock response rule.

When the proxy intercepts a request matching the URL pattern (and optional method),
it returns the configured mock response instead of forwarding to the real server.

URL patterns support * wildcard:
  - "*/api/users*" matches any URL containing /api/users
  - "https://example.com/api/*" matches paths under /api/
  - "*" matches all URLs

Examples:
  Mock a JSON API: urlPattern="*/api/login", statusCode=200, body='{"token":"abc"}'
  Simulate error:  urlPattern="*/api/data", statusCode=500, body='{"error":"Internal Server Error"}'
  Add delay:       urlPattern="*/api/slow", statusCode=200, body='{}', delay=3000`),
			mcp.WithString("url_pattern",
				mcp.Required(),
				mcp.Description("URL wildcard pattern to match (e.g., '*/api/users*')"),
			),
			mcp.WithString("method",
				mcp.Description("HTTP method to match (empty = match all). e.g., 'GET', 'POST'"),
			),
			mcp.WithNumber("status_code",
				mcp.Required(),
				mcp.Description("HTTP status code for the mock response (e.g., 200, 404, 500)"),
			),
			mcp.WithString("body",
				mcp.Description("Response body content (JSON, XML, HTML, or plain text)"),
			),
			mcp.WithNumber("delay",
				mcp.Description("Delay in milliseconds before responding (default: 0)"),
			),
			mcp.WithString("headers_json",
				mcp.Description(`Response headers as JSON object, e.g., '{"Content-Type":"application/json","X-Custom":"value"}'`),
			),
			mcp.WithString("description",
				mcp.Description("Optional description of what this mock rule does"),
			),
		),
		s.handleMockRuleAdd,
	)

	// mock_rule_update - Update a mock rule
	s.server.AddTool(
		mcp.NewTool("mock_rule_update",
			mcp.WithDescription(`Update an existing mock rule. All fields are replaced with the new values.`),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("ID of the mock rule to update"),
			),
			mcp.WithString("url_pattern",
				mcp.Required(),
				mcp.Description("URL wildcard pattern to match"),
			),
			mcp.WithString("method",
				mcp.Description("HTTP method to match (empty = match all)"),
			),
			mcp.WithNumber("status_code",
				mcp.Required(),
				mcp.Description("HTTP status code for the mock response"),
			),
			mcp.WithString("body",
				mcp.Description("Response body content"),
			),
			mcp.WithNumber("delay",
				mcp.Description("Delay in milliseconds before responding"),
			),
			mcp.WithString("headers_json",
				mcp.Description("Response headers as JSON object"),
			),
			mcp.WithBoolean("enabled",
				mcp.Description("Whether this rule is active (default: true)"),
			),
			mcp.WithString("description",
				mcp.Description("Optional description"),
			),
		),
		s.handleMockRuleUpdate,
	)

	// mock_rule_remove - Remove a mock rule
	s.server.AddTool(
		mcp.NewTool("mock_rule_remove",
			mcp.WithDescription("Remove a mock rule by ID."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("ID of the mock rule to remove"),
			),
		),
		s.handleMockRuleRemove,
	)

	// mock_rule_toggle - Enable/disable a mock rule
	s.server.AddTool(
		mcp.NewTool("mock_rule_toggle",
			mcp.WithDescription("Enable or disable a mock rule without removing it."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("ID of the mock rule to toggle"),
			),
			mcp.WithBoolean("enabled",
				mcp.Required(),
				mcp.Description("Whether to enable (true) or disable (false) the rule"),
			),
		),
		s.handleMockRuleToggle,
	)

	// resend_request - Resend an HTTP request
	s.server.AddTool(
		mcp.NewTool("resend_request",
			mcp.WithDescription(`Send an HTTP request with optional modifications. Useful for testing APIs.

Mock rules are checked first — if a matching mock rule exists, the mock response
is returned without making a real network request.

Examples:
  GET request:  method="GET", url="https://api.example.com/users"
  POST request: method="POST", url="https://api.example.com/login",
                headers_json='{"Content-Type":"application/json"}',
                body='{"username":"test","password":"pass"}'`),
			mcp.WithString("method",
				mcp.Required(),
				mcp.Description("HTTP method (GET, POST, PUT, DELETE, PATCH, etc.)"),
			),
			mcp.WithString("url",
				mcp.Required(),
				mcp.Description("Full URL to send the request to"),
			),
			mcp.WithString("headers_json",
				mcp.Description("Request headers as JSON object, e.g., '{\"Content-Type\":\"application/json\"}'"),
			),
			mcp.WithString("body",
				mcp.Description("Request body content"),
			),
		),
		s.handleResendRequest,
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

// --- Mock Rule Handlers ---

func (s *MCPServer) handleMockRuleList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rules := s.app.GetMockRules()

	if len(rules) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("No mock rules configured.")},
		}, nil
	}

	data, _ := json.MarshalIndent(rules, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Mock rules (%d):\n%s", len(rules), string(data)))},
	}, nil
}

func (s *MCPServer) handleMockRuleAdd(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	urlPattern, _ := args["url_pattern"].(string)
	if urlPattern == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("Error: url_pattern is required")}, IsError: true}, nil
	}
	statusCode := 200
	if sc, ok := args["status_code"].(float64); ok {
		statusCode = int(sc)
	}

	method, _ := args["method"].(string)
	body, _ := args["body"].(string)
	description, _ := args["description"].(string)
	delay := 0
	if d, ok := args["delay"].(float64); ok {
		delay = int(d)
	}

	var headers map[string]string
	if hj, ok := args["headers_json"].(string); ok && hj != "" {
		if err := json.Unmarshal([]byte(hj), &headers); err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error parsing headers_json: %v", err))}, IsError: true}, nil
		}
	}

	id := s.app.AddMockRule(urlPattern, method, statusCode, headers, body, delay, description)

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Mock rule added successfully.\nID: %s\nPattern: %s\nStatus: %d", id, urlPattern, statusCode))},
	}, nil
}

func (s *MCPServer) handleMockRuleUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	id, _ := args["id"].(string)
	if id == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("Error: id is required")}, IsError: true}, nil
	}
	urlPattern, _ := args["url_pattern"].(string)
	if urlPattern == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("Error: url_pattern is required")}, IsError: true}, nil
	}
	statusCode := 200
	if sc, ok := args["status_code"].(float64); ok {
		statusCode = int(sc)
	}

	method, _ := args["method"].(string)
	body, _ := args["body"].(string)
	description, _ := args["description"].(string)
	delay := 0
	if d, ok := args["delay"].(float64); ok {
		delay = int(d)
	}
	enabled := true
	if e, ok := args["enabled"].(bool); ok {
		enabled = e
	}

	var headers map[string]string
	if hj, ok := args["headers_json"].(string); ok && hj != "" {
		if err := json.Unmarshal([]byte(hj), &headers); err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error parsing headers_json: %v", err))}, IsError: true}, nil
		}
	}

	if err := s.app.UpdateMockRule(id, urlPattern, method, statusCode, headers, body, delay, enabled, description); err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error: %v", err))}, IsError: true}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Mock rule updated: %s", id))},
	}, nil
}

func (s *MCPServer) handleMockRuleRemove(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	id, _ := args["id"].(string)
	if id == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("Error: id is required")}, IsError: true}, nil
	}

	s.app.RemoveMockRule(id)

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Mock rule removed: %s", id))},
	}, nil
}

func (s *MCPServer) handleMockRuleToggle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	id, _ := args["id"].(string)
	if id == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("Error: id is required")}, IsError: true}, nil
	}
	enabled, _ := args["enabled"].(bool)

	if err := s.app.ToggleMockRule(id, enabled); err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error: %v", err))}, IsError: true}, nil
	}

	state := "disabled"
	if enabled {
		state = "enabled"
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Mock rule %s: %s", state, id))},
	}, nil
}

func (s *MCPServer) handleResendRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	method, _ := args["method"].(string)
	if method == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("Error: method is required")}, IsError: true}, nil
	}
	url, _ := args["url"].(string)
	if url == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("Error: url is required")}, IsError: true}, nil
	}
	body, _ := args["body"].(string)

	var headers map[string]string
	if hj, ok := args["headers_json"].(string); ok && hj != "" {
		if err := json.Unmarshal([]byte(hj), &headers); err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error parsing headers_json: %v", err))}, IsError: true}, nil
		}
	}

	result, err := s.app.ResendRequest(method, url, headers, body)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Request failed: %v", err))}, IsError: true}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(string(data))},
	}, nil
}
