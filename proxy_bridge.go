package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"Gaze/proxy"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// proxyDeviceId tracks which device the proxy is monitoring for session events
var (
	proxyDeviceId string
	proxyDeviceMu sync.RWMutex

	// Track which session started the proxy (empty = proxy was started manually/externally)
	proxyOwnerSessionID string
	proxyOwnerMu        sync.Mutex

	// Track emitted request IDs to avoid duplicates (with TTL-based cleanup)
	emittedRequests    = make(map[string]int64) // requestId -> timestamp (unix nano)
	emittedRequestsMu  sync.Mutex
	emittedRequestsTTL = int64(5 * 60 * 1e9) // 5 minutes in nanoseconds
	emittedRequestsMax = 5000                // Max entries before cleanup
)

// SetProxyDevice sets which device the proxy is monitoring (for session events)
func (a *App) SetProxyDevice(deviceId string) {
	proxyDeviceMu.Lock()
	proxyDeviceId = deviceId
	proxyDeviceMu.Unlock()
}

// GetProxyDevice returns the currently monitored device
func (a *App) GetProxyDevice() string {
	proxyDeviceMu.RLock()
	defer proxyDeviceMu.RUnlock()
	return proxyDeviceId
}

// setProxyOwnerSession records which session started the proxy
func setProxyOwnerSession(sessionID string) {
	proxyOwnerMu.Lock()
	proxyOwnerSessionID = sessionID
	proxyOwnerMu.Unlock()
}

// getProxyOwnerSession returns the session that started the proxy
func getProxyOwnerSession() string {
	proxyOwnerMu.Lock()
	defer proxyOwnerMu.Unlock()
	return proxyOwnerSessionID
}

// clearProxyOwnerSession clears the proxy owner
func clearProxyOwnerSession() {
	proxyOwnerMu.Lock()
	proxyOwnerSessionID = ""
	proxyOwnerMu.Unlock()
}

// StartProxy starts the internal HTTP/HTTPS proxy
func (a *App) StartProxy(port int) (string, error) {
	LogUserAction(ActionProxyStart, "", map[string]interface{}{
		"port": port,
	})

	// Clear emitted requests map on start
	emittedRequestsMu.Lock()
	emittedRequests = make(map[string]int64)
	emittedRequestsMu.Unlock()

	err := proxy.GetProxy().Start(port, func(req proxy.RequestLog) {
		// Skip partial updates (body-only size updates during transfer)
		if req.PartialUpdate {
			return
		}

		// Skip pending requests (status=0) - wait for response
		if req.StatusCode == 0 {
			return
		}

		// Safety-net dedup: emit only once per request ID.
		// With the proxy.go fixes (logRequest notify=false when TransparentReadCloser
		// will handle it, and doneCalled guard in update), each request should only
		// reach here once. This map is a defensive safeguard.
		emittedRequestsMu.Lock()
		now := time.Now().UnixNano()
		if _, alreadyEmitted := emittedRequests[req.Id]; alreadyEmitted {
			emittedRequestsMu.Unlock()
			return
		}
		emittedRequests[req.Id] = now

		// TTL-based cleanup when map grows large
		if len(emittedRequests) > emittedRequestsMax {
			for id, ts := range emittedRequests {
				if now-ts > emittedRequestsTTL {
					delete(emittedRequests, id)
				}
			}
			if len(emittedRequests) > emittedRequestsMax {
				count := 0
				target := len(emittedRequests) / 2
				for id := range emittedRequests {
					if count >= target {
						break
					}
					delete(emittedRequests, id)
					count++
				}
			}
		}
		emittedRequestsMu.Unlock()

		// Calculate duration from ID (SessionId-TimestampNano)
		var durationMs int64
		parts := strings.Split(req.Id, "-")
		if len(parts) >= 2 {
			if startNano, err := strconv.ParseInt(parts[len(parts)-1], 10, 64); err == nil {
				durationMs = (time.Now().UnixNano() - startNano) / 1e6
			}
		}

		proxyDeviceMu.RLock()
		deviceId := proxyDeviceId
		proxyDeviceMu.RUnlock()

		// Debug: Log proxy device ID
		fmt.Printf("[ProxyBridge] Emitting network event: proxyDeviceId=%q, method=%s, url=%s\n",
			deviceId, req.Method, req.URL)

		// Determine level based on status code
		level := "info"
		if req.StatusCode >= 400 && req.StatusCode < 500 {
			level = "warn"
		} else if req.StatusCode >= 500 {
			level = "error"
		}

		title := fmt.Sprintf("%s %s → %d", req.Method, req.URL, req.StatusCode)
		if len(title) > 100 {
			title = title[:97] + "..."
		}

		// Emit via session manager - this ensures events go to session-events-batch
		// which ProxyView is listening to
		detail := map[string]interface{}{
			"id":              req.Id,
			"method":          req.Method,
			"url":             req.URL,
			"statusCode":      req.StatusCode,
			"contentType":     req.ContentType,
			"bodySize":        req.BodySize,
			"duration":        durationMs,
			"isHttps":         req.IsHTTPS,
			"isWs":            req.IsWs,
			"clientIp":        req.ClientIP,
			"requestHeaders":  req.Headers,
			"requestBody":     req.Body,
			"responseHeaders": req.RespHeaders,
			"responseBody":    req.RespBody,
			"mocked":          req.Mocked,
		}

		a.EmitSessionEventFull(SessionEvent{
			ID:        uuid.New().String(),
			DeviceID:  deviceId,
			Timestamp: time.Now().UnixMilli(),
			Type:      "network_request",
			Category:  "network",
			Level:     level,
			Title:     title,
			Detail:    detail,
			Duration:  durationMs,
		})
	})
	if err != nil {
		return "", err
	}

	// Register enabled mock rules with the proxy
	mockRulesMu.RLock()
	for _, rule := range mockRules {
		if rule.Enabled {
			proxy.GetProxy().AddMockRule(rule.ID, rule.URLPattern, rule.Method, rule.StatusCode, rule.Headers, rule.Body, rule.Delay)
		}
	}
	mockRulesMu.RUnlock()

	a.emitProxyStatus(true, port)
	return "Proxy started successfully", nil
}

// SetupProxyForDevice configures adb reverse and device proxy settings
// This allows the device to connect to the proxy via localhost through adb tunnel
func (a *App) SetupProxyForDevice(deviceId string, port int) error {
	if deviceId == "" {
		return fmt.Errorf("no device specified")
	}

	// 1. Setup adb reverse: device's localhost:port -> host's localhost:port
	reverseCmd := a.newAdbCommand(nil, "-s", deviceId, "reverse", "tcp:"+strconv.Itoa(port), "tcp:"+strconv.Itoa(port))
	if out, err := reverseCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("adb reverse failed: %v, output: %s", err, string(out))
	}

	// 2. Set device proxy to localhost (traffic goes through adb tunnel)
	proxyCmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "settings", "put", "global", "http_proxy", fmt.Sprintf("127.0.0.1:%d", port))
	if out, err := proxyCmd.CombinedOutput(); err != nil {
		// Try to clean up reverse on failure
		a.newAdbCommand(nil, "-s", deviceId, "reverse", "--remove", "tcp:"+strconv.Itoa(port)).Run()
		return fmt.Errorf("set proxy failed: %v, output: %s", err, string(out))
	}

	return nil
}

// CleanupProxyForDevice removes adb reverse and clears device proxy settings
func (a *App) CleanupProxyForDevice(deviceId string, port int) error {
	if deviceId == "" {
		return nil
	}

	// 1. Clear device proxy
	a.newAdbCommand(nil, "-s", deviceId, "shell", "settings", "put", "global", "http_proxy", ":0").Run()

	// 2. Remove adb reverse
	a.newAdbCommand(nil, "-s", deviceId, "reverse", "--remove", "tcp:"+strconv.Itoa(port)).Run()

	return nil
}

// StopProxy stops the internal proxy and cleans up device settings
func (a *App) StopProxy() (string, error) {
	LogUserAction(ActionProxyStop, "", nil)

	// Get device and port before stopping
	deviceId := a.GetProxyDevice()
	port := proxy.GetProxy().GetPort()

	// Clean up device proxy settings first (before stopping proxy)
	if deviceId != "" && port > 0 {
		a.CleanupProxyForDevice(deviceId, port)
	}

	// Clear the tracked device and owner
	a.SetProxyDevice("")
	clearProxyOwnerSession()

	err := proxy.GetProxy().Stop()
	if err != nil {
		return "", err
	}
	a.emitProxyStatus(false, port)
	return "Proxy stopped successfully", nil
}

// GetProxyStatus returns true if the proxy is running
func (a *App) GetProxyStatus() bool {
	return proxy.GetProxy().IsRunning()
}

// emitProxyStatus notifies the frontend of proxy status changes
func (a *App) emitProxyStatus(running bool, port int) {
	if a.ctx != nil && !a.mcpMode {
		wailsRuntime.EventsEmit(a.ctx, "proxy-status-changed", map[string]interface{}{
			"running": running,
			"port":    port,
		})
	}
}

// SetProxyLimit sets the upload and download speed limits for the proxy server (bytes per second)
func (a *App) SetProxyLimit(uploadSpeed, downloadSpeed int) {
	proxy.GetProxy().SetLimits(uploadSpeed, downloadSpeed)
}

// SetProxyWSEnabled enables or disables WebSocket support
func (a *App) SetProxyWSEnabled(enabled bool) {
	proxy.GetProxy().SetWSEnabled(enabled)
}

// SetProxyMITM enables or disables HTTPS Decryption (MITM)
func (a *App) SetProxyMITM(enabled bool) {
	proxy.GetProxy().SetProxyMITM(enabled)
}

// SetMITMBypassPatterns sets the keywords/domains to bypass MITM
func (a *App) SetMITMBypassPatterns(patterns []string) {
	proxy.GetProxy().SetMITMBypassPatterns(patterns)
}

// GetMITMBypassPatterns returns the current bypass patterns
func (a *App) GetMITMBypassPatterns() []string {
	return proxy.GetProxy().GetMITMBypassPatterns()
}

// GetProxySettings returns the current proxy settings
func (a *App) GetProxySettings() map[string]interface{} {
	return map[string]interface{}{
		"wsEnabled":      proxy.GetProxy().IsWSEnabled(),
		"mitmEnabled":    proxy.GetProxy().IsMITMEnabled(),
		"bypassPatterns": proxy.GetProxy().GetMITMBypassPatterns(),
	}
}

// InstallProxyCert pushes the generated CA certificate to the device
func (a *App) InstallProxyCert(deviceId string) (string, error) {
	certPath := proxy.GetProxy().GetCertPath()
	if certPath == "" {
		return "", nil
	}

	dest := "/sdcard/Download/Gaze-CA.crt"

	cmd := a.newAdbCommand(nil, "-s", deviceId, "push", certPath, dest)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", err
	} else {
		_ = out
	}

	return dest, nil
}

// CheckCertTrust checks if the device has trusted the proxy CA certificate
// Returns: "trusted", "not_trusted", "unknown", or "no_proxy"
func (a *App) CheckCertTrust(deviceId string) string {
	// Check if proxy is running
	p := proxy.GetProxy()
	if !p.IsRunning() {
		return "no_proxy"
	}

	// Check if MITM is enabled
	if !p.IsMITMEnabled() {
		return "unknown"
	}

	// Check recent proxy logs - if we have any successfully decrypted HTTPS requests
	// (requests with response body from HTTPS URLs), the cert is trusted
	hasDecryptedHTTPS := p.HasRecentDecryptedHTTPS()
	if hasDecryptedHTTPS {
		return "trusted"
	}

	// No decrypted HTTPS yet - could be trusted but no traffic, or not trusted
	// Return "pending" to indicate we need more traffic to determine
	return "pending"
}

// SetProxyLatency sets the artificial latency in milliseconds
func (a *App) SetProxyLatency(latencyMs int) {
	proxy.GetProxy().SetLatency(latencyMs)
}

// isStaticResource checks if a request is for a static resource (image, css, js, font, etc.)
// Returns true if it should be filtered out from the timeline
func isStaticResource(url, contentType string) bool {
	// Check by content type first (most reliable)
	if contentType != "" {
		staticContentTypes := []string{
			"image/",
			"text/css",
			"text/javascript",
			"application/javascript",
			"application/x-javascript",
			"font/",
			"application/font",
			"application/x-font",
			"video/",
			"audio/",
		}

		// Normalize content type (remove charset, etc.)
		ctLower := contentType
		for idx := 0; idx < len(contentType); idx++ {
			if contentType[idx] == ';' {
				ctLower = contentType[:idx]
				break
			}
		}

		for _, ct := range staticContentTypes {
			if len(ctLower) >= len(ct) {
				match := true
				for j := 0; j < len(ct); j++ {
					c1 := ctLower[j]
					c2 := ct[j]
					if c1 >= 'A' && c1 <= 'Z' {
						c1 += 32
					}
					if c1 != byte(c2) {
						match = false
						break
					}
				}
				if match && (len(ctLower) == len(ct) || ctLower[len(ct)] == '/' || ctLower[len(ct)] == ';') {
					return true
				}
			}
		}
	}

	// Check by URL extension as fallback
	// Extract path part (before ? or #)
	pathEnd := len(url)
	for i := 0; i < len(url); i++ {
		if url[i] == '?' || url[i] == '#' {
			pathEnd = i
			break
		}
	}

	if pathEnd == 0 {
		return false
	}

	path := url[:pathEnd]

	// Find the last dot in the path
	lastDot := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			lastDot = i
			break
		}
		if path[i] == '/' {
			// No extension found before path separator
			break
		}
	}

	if lastDot == -1 || lastDot == len(path)-1 {
		return false
	}

	// Extract extension (including the dot)
	ext := path[lastDot:]

	// Convert to lowercase
	extLower := ""
	for i := 0; i < len(ext); i++ {
		c := ext[i]
		if c >= 'A' && c <= 'Z' {
			extLower += string(c + 32)
		} else {
			extLower += string(c)
		}
	}

	// Check against known static extensions
	staticExtensions := []string{
		".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".ico", ".bmp",
		".css", ".js", ".mjs",
		".woff", ".woff2", ".ttf", ".eot", ".otf",
		".mp4", ".webm", ".ogg", ".mp3", ".wav",
		".pdf", ".zip", ".tar", ".gz",
	}

	for _, staticExt := range staticExtensions {
		if extLower == staticExt {
			return true
		}
	}

	return false
}

// matchMockRuleLocal checks if a request matches any enabled mock rule
func matchMockRuleLocal(method, url string) *MockRule {
	mockRulesMu.RLock()
	defer mockRulesMu.RUnlock()

	for _, rule := range mockRules {
		if !rule.Enabled {
			continue
		}
		// Check method (empty means match all)
		if rule.Method != "" && rule.Method != method {
			continue
		}
		// Check URL pattern
		if matchPatternLocal(url, rule.URLPattern) {
			return rule
		}
	}
	return nil
}

// matchPatternLocal checks if a URL matches a pattern with * wildcards
func matchPatternLocal(url, pattern string) bool {
	if pattern == "*" {
		return true
	}
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return url == pattern
	}
	pos := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(url[pos:], part)
		if idx == -1 {
			return false
		}
		if i == 0 && idx != 0 {
			return false
		}
		pos += idx + len(part)
	}
	if !strings.HasSuffix(pattern, "*") && pos != len(url) {
		return false
	}
	return true
}

// ResendRequest sends an HTTP request with optional modifications
// Returns the response status, headers, and body
// Checks mock rules first, then sends actual request if no match
func (a *App) ResendRequest(method, url string, headers map[string]string, body string) (map[string]interface{}, error) {
	// Check for matching mock rule first (works without proxy running)
	if mockRule := matchMockRuleLocal(method, url); mockRule != nil {
		// Apply mock delay
		if mockRule.Delay > 0 {
			time.Sleep(time.Duration(mockRule.Delay) * time.Millisecond)
		}

		// Build mock response headers
		respHeaders := make(map[string]string)
		for k, v := range mockRule.Headers {
			respHeaders[k] = v
		}
		if respHeaders["Content-Type"] == "" {
			respHeaders["Content-Type"] = "application/json"
		}

		return map[string]interface{}{
			"statusCode":  mockRule.StatusCode,
			"status":      fmt.Sprintf("%d %s", mockRule.StatusCode, http.StatusText(mockRule.StatusCode)),
			"headers":     respHeaders,
			"body":        mockRule.Body,
			"bodySize":    len(mockRule.Body),
			"duration":    int64(mockRule.Delay),
			"contentType": respHeaders["Content-Type"],
			"mocked":      true,
		}, nil
	}

	// No mock match, send actual request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var reqBody io.Reader
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime).Milliseconds()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Build response headers map
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		respHeaders[k] = strings.Join(v, ", ")
	}

	return map[string]interface{}{
		"statusCode":  resp.StatusCode,
		"status":      resp.Status,
		"headers":     respHeaders,
		"body":        string(respBody),
		"bodySize":    len(respBody),
		"duration":    duration,
		"contentType": resp.Header.Get("Content-Type"),
	}, nil
}

// MockRule defines a rule for mocking HTTP responses
type MockRule struct {
	ID          string            `json:"id"`
	URLPattern  string            `json:"urlPattern"`  // URL pattern to match (supports * wildcard)
	Method      string            `json:"method"`      // HTTP method to match (empty = all)
	StatusCode  int               `json:"statusCode"`  // Response status code
	Headers     map[string]string `json:"headers"`     // Response headers
	Body        string            `json:"body"`        // Response body
	Delay       int               `json:"delay"`       // Delay in milliseconds before responding
	Enabled     bool              `json:"enabled"`     // Whether this rule is active
	Description string            `json:"description"` // Optional description
}

var (
	mockRules   = make(map[string]*MockRule)
	mockRulesMu sync.RWMutex
)

// getMockRulesPath returns the path to the mock rules JSON file
func getMockRulesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".adbGUI", "mock_rules.json")
}

// saveMockRules saves mock rules to disk
func saveMockRules() error {
	rules := make([]*MockRule, 0, len(mockRules))
	for _, rule := range mockRules {
		rules = append(rules, rule)
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}

	path := getMockRulesPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// LoadMockRules loads mock rules from disk (called on app startup)
func (a *App) LoadMockRules() error {
	mockRulesMu.Lock()
	defer mockRulesMu.Unlock()

	path := getMockRulesPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No saved rules, that's fine
		}
		return err
	}

	var rules []*MockRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return err
	}

	mockRules = make(map[string]*MockRule)
	for _, rule := range rules {
		mockRules[rule.ID] = rule
		// Register enabled rules with proxy if it's running
		if rule.Enabled && proxy.GetProxy().IsRunning() {
			proxy.GetProxy().AddMockRule(rule.ID, rule.URLPattern, rule.Method, rule.StatusCode, rule.Headers, rule.Body, rule.Delay)
		}
	}

	return nil
}

// AddMockRule adds a new mock response rule
func (a *App) AddMockRule(rule MockRule) string {
	mockRulesMu.Lock()
	defer mockRulesMu.Unlock()

	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	rule.Enabled = true
	mockRules[rule.ID] = &rule

	// Register with proxy
	proxy.GetProxy().AddMockRule(rule.ID, rule.URLPattern, rule.Method, rule.StatusCode, rule.Headers, rule.Body, rule.Delay)

	// Persist to disk
	saveMockRules()

	return rule.ID
}

// UpdateMockRule updates an existing mock rule
func (a *App) UpdateMockRule(rule MockRule) error {
	mockRulesMu.Lock()
	defer mockRulesMu.Unlock()

	if _, exists := mockRules[rule.ID]; !exists {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}

	mockRules[rule.ID] = &rule

	// Update in proxy
	proxy.GetProxy().RemoveMockRule(rule.ID)
	if rule.Enabled {
		proxy.GetProxy().AddMockRule(rule.ID, rule.URLPattern, rule.Method, rule.StatusCode, rule.Headers, rule.Body, rule.Delay)
	}

	// Persist to disk
	saveMockRules()

	return nil
}

// RemoveMockRule removes a mock response rule
func (a *App) RemoveMockRule(ruleID string) {
	mockRulesMu.Lock()
	defer mockRulesMu.Unlock()

	delete(mockRules, ruleID)
	proxy.GetProxy().RemoveMockRule(ruleID)

	// Persist to disk
	saveMockRules()
}

// GetMockRules returns all mock response rules
func (a *App) GetMockRules() []*MockRule {
	mockRulesMu.RLock()
	defer mockRulesMu.RUnlock()

	rules := make([]*MockRule, 0, len(mockRules))
	for _, rule := range mockRules {
		rules = append(rules, rule)
	}
	return rules
}

// ToggleMockRule enables or disables a mock rule
func (a *App) ToggleMockRule(ruleID string, enabled bool) error {
	mockRulesMu.Lock()
	defer mockRulesMu.Unlock()

	rule, exists := mockRules[ruleID]
	if !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	rule.Enabled = enabled

	if enabled {
		proxy.GetProxy().AddMockRule(rule.ID, rule.URLPattern, rule.Method, rule.StatusCode, rule.Headers, rule.Body, rule.Delay)
	} else {
		proxy.GetProxy().RemoveMockRule(rule.ID)
	}

	// Persist to disk
	saveMockRules()

	return nil
}
