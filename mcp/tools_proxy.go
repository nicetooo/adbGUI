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
				mcp.Description("Port to listen on (default: 8080)"),
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
			mcp.WithString("conditions_json",
				mcp.Description(`Optional JSON array of match conditions (AND logic). Each condition:
  {"type":"header|query|body", "key":"header-name or param-name", "operator":"equals|contains|regex|exists|not_exists", "value":"expected"}
Example: '[{"type":"header","key":"Authorization","operator":"exists"},{"type":"query","key":"page","operator":"equals","value":"1"}]'`),
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
			mcp.WithString("conditions_json",
				mcp.Description(`Optional JSON array of match conditions (AND logic). Same format as mock_rule_add.`),
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

	// proxy_configure - Configure proxy settings
	s.server.AddTool(
		mcp.NewTool("proxy_configure",
			mcp.WithDescription(`Configure proxy settings. All parameters are optional — only provided settings will be changed.

Settings:
- mitm: Enable/disable HTTPS MITM decryption
- ws_enabled: Enable/disable WebSocket pass-through
- upload_speed: Upload speed limit in KB/s (0 = unlimited)
- download_speed: Download speed limit in KB/s (0 = unlimited)
- latency: Artificial latency in milliseconds (0 = no latency)
- bypass_patterns: JSON array of domain patterns to bypass MITM (replaces existing list)

Examples:
  Enable MITM: mitm=true
  Throttle to 100KB/s: upload_speed=100, download_speed=100
  Add 200ms latency: latency=200
  Set bypass: bypass_patterns='["cdn","static","*.google.com"]'`),
			mcp.WithBoolean("mitm",
				mcp.Description("Enable/disable HTTPS MITM decryption"),
			),
			mcp.WithBoolean("ws_enabled",
				mcp.Description("Enable/disable WebSocket pass-through"),
			),
			mcp.WithNumber("upload_speed",
				mcp.Description("Upload speed limit in KB/s (0 = unlimited)"),
			),
			mcp.WithNumber("download_speed",
				mcp.Description("Download speed limit in KB/s (0 = unlimited)"),
			),
			mcp.WithNumber("latency",
				mcp.Description("Artificial latency in milliseconds (0 = none)"),
			),
			mcp.WithString("bypass_patterns",
				mcp.Description(`JSON array of MITM bypass domain patterns, e.g. '["cdn","*.google.com"]'. Replaces existing list.`),
			),
		),
		s.handleProxyConfigure,
	)

	// proxy_settings - Get current proxy settings
	s.server.AddTool(
		mcp.NewTool("proxy_settings",
			mcp.WithDescription("Get current proxy settings including MITM state, WebSocket state, bypass patterns, and connected device."),
		),
		s.handleProxySettings,
	)

	// proxy_device_setup - Set up proxy on a device
	s.server.AddTool(
		mcp.NewTool("proxy_device_setup",
			mcp.WithDescription(`Set up the proxy on a connected Android device.

This performs:
1. Sets the proxy device ID
2. Creates an adb reverse tunnel (device port -> host proxy port)
3. Configures the device HTTP proxy settings

The device will route HTTP/HTTPS traffic through the proxy after setup.`),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device serial number"),
			),
			mcp.WithNumber("port",
				mcp.Description("Proxy port (default: 8080)"),
			),
		),
		s.handleProxyDeviceSetup,
	)

	// proxy_device_cleanup - Remove proxy from device
	s.server.AddTool(
		mcp.NewTool("proxy_device_cleanup",
			mcp.WithDescription(`Remove proxy configuration from an Android device.

This performs:
1. Removes the adb reverse tunnel
2. Clears the device HTTP proxy settings
3. Clears the proxy device ID`),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device serial number"),
			),
			mcp.WithNumber("port",
				mcp.Description("Proxy port used during setup (default: 8080)"),
			),
		),
		s.handleProxyDeviceCleanup,
	)

	// proxy_cert_install - Push CA certificate to device
	s.server.AddTool(
		mcp.NewTool("proxy_cert_install",
			mcp.WithDescription(`Push the proxy CA certificate to an Android device for HTTPS interception.

Pushes the CA certificate to /sdcard/Download/Gaze-CA.crt.
After pushing, the user must install it on the device:
  Settings > Security > Install from storage > Select Gaze-CA.crt`),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device serial number"),
			),
		),
		s.handleProxyCertInstall,
	)

	// proxy_cert_trust_check - Check certificate trust status
	s.server.AddTool(
		mcp.NewTool("proxy_cert_trust_check",
			mcp.WithDescription(`Check if a device trusts the proxy CA certificate.

Returns "trusted", "untrusted", or "unknown" based on whether
recent HTTPS traffic has been successfully decrypted.`),
			mcp.WithString("device_id",
				mcp.Required(),
				mcp.Description("Device serial number"),
			),
		),
		s.handleProxyCertTrustCheck,
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
	port := 8080
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

	var conditions []MCPMockCondition
	if cj, ok := args["conditions_json"].(string); ok && cj != "" {
		if err := json.Unmarshal([]byte(cj), &conditions); err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error parsing conditions_json: %v", err))}, IsError: true}, nil
		}
	}

	id := s.app.AddMockRule(urlPattern, method, statusCode, headers, body, delay, description, conditions)

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

	var conditions []MCPMockCondition
	if cj, ok := args["conditions_json"].(string); ok && cj != "" {
		if err := json.Unmarshal([]byte(cj), &conditions); err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error parsing conditions_json: %v", err))}, IsError: true}, nil
		}
	}

	if err := s.app.UpdateMockRule(id, urlPattern, method, statusCode, headers, body, delay, enabled, description, conditions); err != nil {
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

// --- Proxy Configuration Handlers ---

func (s *MCPServer) handleProxyConfigure(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	var changes []string

	if mitm, ok := args["mitm"].(bool); ok {
		s.app.SetProxyMITM(mitm)
		changes = append(changes, fmt.Sprintf("MITM: %v", mitm))
	}
	if ws, ok := args["ws_enabled"].(bool); ok {
		s.app.SetProxyWSEnabled(ws)
		changes = append(changes, fmt.Sprintf("WebSocket: %v", ws))
	}

	ulSet, dlSet := false, false
	ul, dl := 0, 0
	if u, ok := args["upload_speed"].(float64); ok {
		ul = int(u)
		ulSet = true
	}
	if d, ok := args["download_speed"].(float64); ok {
		dl = int(d)
		dlSet = true
	}
	if ulSet || dlSet {
		s.app.SetProxyLimit(ul, dl)
		changes = append(changes, fmt.Sprintf("Speed limit: UL=%dKB/s DL=%dKB/s", ul, dl))
	}

	if lat, ok := args["latency"].(float64); ok {
		s.app.SetProxyLatency(int(lat))
		changes = append(changes, fmt.Sprintf("Latency: %dms", int(lat)))
	}

	if bp, ok := args["bypass_patterns"].(string); ok && bp != "" {
		var patterns []string
		if err := json.Unmarshal([]byte(bp), &patterns); err != nil {
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error parsing bypass_patterns: %v", err))}, IsError: true}, nil
		}
		s.app.SetMITMBypassPatterns(patterns)
		changes = append(changes, fmt.Sprintf("Bypass patterns: %v", patterns))
	}

	if len(changes) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("No settings changed. Provide at least one parameter.")},
		}, nil
	}

	msg := "Proxy settings updated:\n"
	for _, c := range changes {
		msg += "  - " + c + "\n"
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(msg)},
	}, nil
}

func (s *MCPServer) handleProxySettings(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	settings := s.app.GetProxySettings()
	settings["running"] = s.app.GetProxyStatus()
	settings["proxyDevice"] = s.app.GetProxyDevice()
	settings["bypassPatterns"] = s.app.GetMITMBypassPatterns()

	data, _ := json.MarshalIndent(settings, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(string(data))},
	}, nil
}

func (s *MCPServer) handleProxyDeviceSetup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceId, _ := args["device_id"].(string)
	if deviceId == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("Error: device_id is required")}, IsError: true}, nil
	}
	port := 8080
	if p, ok := args["port"].(float64); ok {
		port = int(p)
	}

	s.app.SetProxyDevice(deviceId)
	if err := s.app.SetupProxyForDevice(deviceId, port); err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error setting up proxy on device: %v", err))}, IsError: true}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Proxy configured on device %s (port %d).\nadb reverse tunnel created and HTTP proxy settings applied.", deviceId, port))},
	}, nil
}

func (s *MCPServer) handleProxyDeviceCleanup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceId, _ := args["device_id"].(string)
	if deviceId == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("Error: device_id is required")}, IsError: true}, nil
	}
	port := 8080
	if p, ok := args["port"].(float64); ok {
		port = int(p)
	}

	if err := s.app.CleanupProxyForDevice(deviceId, port); err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error cleaning up proxy: %v", err))}, IsError: true}, nil
	}
	s.app.SetProxyDevice("")

	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Proxy removed from device %s. Reverse tunnel and HTTP proxy settings cleared.", deviceId))},
	}, nil
}

func (s *MCPServer) handleProxyCertInstall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceId, _ := args["device_id"].(string)
	if deviceId == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("Error: device_id is required")}, IsError: true}, nil
	}

	result, err := s.app.InstallProxyCert(deviceId)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error pushing certificate: %v", err))}, IsError: true}, nil
	}

	msg := fmt.Sprintf("CA certificate pushed to device %s.\nPath: /sdcard/Download/Gaze-CA.crt\n%s\n\nNext steps: Install the certificate on the device via Settings > Security > Install from storage.", deviceId, result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(msg)},
	}, nil
}

func (s *MCPServer) handleProxyCertTrustCheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	deviceId, _ := args["device_id"].(string)
	if deviceId == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("Error: device_id is required")}, IsError: true}, nil
	}

	status := s.app.CheckCertTrust(deviceId)
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Certificate trust status for %s: %s", deviceId, status))},
	}, nil
}
