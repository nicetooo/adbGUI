package proxy

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"golang.org/x/time/rate"
)

func copyHeader(h http.Header) http.Header {
	if h == nil {
		return nil
	}
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

// MockCondition defines an additional match condition for mock rules.
// All conditions must match (AND logic) for a rule to apply.
type MockCondition struct {
	Type     string `json:"type"`     // "header", "query", "body"
	Key      string `json:"key"`      // header name or query param name (unused for body type)
	Operator string `json:"operator"` // "equals", "contains", "regex", "exists", "not_exists"
	Value    string `json:"value"`    // expected value (unused for exists/not_exists)
}

// MockRule defines a rule for mocking HTTP responses
type MockRule struct {
	ID         string
	URLPattern string
	Method     string
	StatusCode int
	Headers    map[string]string
	Body       string
	Delay      int
	Conditions []MockCondition // Additional match conditions (AND logic)
}

// ProxyServer handles the HTTP/HTTPS logic using goproxy
type ProxyServer struct {
	server             *http.Server
	proxy              *goproxy.ProxyHttpServer
	listener           net.Listener
	mu                 sync.Mutex
	running            bool
	port               int
	OnRequest          func(RequestLog) // Callback for request logging
	mitmEnabled        bool             // HTTPS Decrypt
	wsEnabled          bool             // WebSocket support
	mitmBypassPatterns []string
	certMgr            *CertManager

	upLimiter   *rate.Limiter
	downLimiter *rate.Limiter
	latency     time.Duration // Artificial latency

	mockRules map[string]*MockRule // Mock response rules

	hasDecryptedHTTPS bool // Track if we've seen decrypted HTTPS traffic

	// reqBodyCache stores captured request bodies keyed by request ID.
	// Written by the request TransparentReadCloser, read by the response one.
	reqBodyCache   map[string]cachedReqBody
	reqBodyCacheMu sync.Mutex
}

// cachedReqBody holds both the display text and raw bytes for a request body.
type cachedReqBody struct {
	text     string
	rawBytes []byte // non-nil when binary data detected
}

// GetPort returns the port the proxy is running on
func (p *ProxyServer) GetPort() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.port
}

// SetLatency sets the artificial latency in milliseconds
func (p *ProxyServer) SetLatency(latencyMs int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.latency = time.Duration(latencyMs) * time.Millisecond
}

// AddMockRule adds a mock response rule
func (p *ProxyServer) AddMockRule(id, urlPattern, method string, statusCode int, headers map[string]string, body string, delay int, conditions []MockCondition) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(os.Stderr, "[MOCK DEBUG] AddMockRule: id=%s pattern=%s method=%s status=%d conditions=%d\n", id, urlPattern, method, statusCode, len(conditions))
	if p.mockRules == nil {
		p.mockRules = make(map[string]*MockRule)
	}
	p.mockRules[id] = &MockRule{
		ID:         id,
		URLPattern: urlPattern,
		Method:     method,
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
		Delay:      delay,
		Conditions: conditions,
	}
}

// RemoveMockRule removes a mock response rule
func (p *ProxyServer) RemoveMockRule(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mockRules != nil {
		delete(p.mockRules, id)
	}
}

// matchMockRule checks if a request matches any mock rule and returns it.
// It accepts the full http.Request to evaluate header, query, and body conditions.
// bodyBytes is the pre-read request body (may be nil if no rules need body matching).
func (p *ProxyServer) matchMockRule(r *http.Request, bodyBytes []byte) *MockRule {
	p.mu.Lock()
	defer p.mu.Unlock()

	method := r.Method
	url := r.URL.String()

	fmt.Fprintf(os.Stderr, "[MOCK DEBUG] Checking %s %s against %d rules\n", method, url, len(p.mockRules))

	for id, rule := range p.mockRules {
		fmt.Fprintf(os.Stderr, "[MOCK DEBUG]   Rule %s: method=%s pattern=%s conditions=%d\n", id, rule.Method, rule.URLPattern, len(rule.Conditions))
		// Check method (empty means match all)
		if rule.Method != "" && rule.Method != method {
			continue
		}

		// Check URL pattern (supports * wildcard)
		if !MatchPattern(url, rule.URLPattern) {
			continue
		}

		// Check additional conditions (AND logic)
		if len(rule.Conditions) > 0 && !evaluateConditions(rule.Conditions, r, bodyBytes) {
			fmt.Fprintf(os.Stderr, "[MOCK DEBUG]     Conditions not met\n")
			continue
		}

		fmt.Fprintf(os.Stderr, "[MOCK DEBUG]     MATCHED\n")
		return rule
	}
	return nil
}

// hasBodyConditions checks if any registered mock rule has body-type conditions.
func (p *ProxyServer) hasBodyConditions() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, rule := range p.mockRules {
		for _, c := range rule.Conditions {
			if c.Type == "body" {
				return true
			}
		}
	}
	return false
}

// evaluateConditions checks all conditions against the request (AND logic).
func evaluateConditions(conditions []MockCondition, r *http.Request, bodyBytes []byte) bool {
	for _, cond := range conditions {
		if !evaluateCondition(cond, r, bodyBytes) {
			return false
		}
	}
	return true
}

// evaluateCondition checks a single condition against the request.
func evaluateCondition(cond MockCondition, r *http.Request, bodyBytes []byte) bool {
	switch cond.Type {
	case "header":
		return evaluateHeaderCondition(cond, r.Header)
	case "query":
		return evaluateQueryCondition(cond, r.URL.Query())
	case "body":
		return evaluateBodyCondition(cond, bodyBytes)
	default:
		return true // unknown condition type → pass
	}
}

func evaluateHeaderCondition(cond MockCondition, headers http.Header) bool {
	value := headers.Get(cond.Key)
	exists := value != "" || headers[http.CanonicalHeaderKey(cond.Key)] != nil
	return matchOperator(cond.Operator, value, cond.Value, exists)
}

func evaluateQueryCondition(cond MockCondition, query map[string][]string) bool {
	values, exists := query[cond.Key]
	value := ""
	if len(values) > 0 {
		value = values[0]
	}
	return matchOperator(cond.Operator, value, cond.Value, exists)
}

func evaluateBodyCondition(cond MockCondition, bodyBytes []byte) bool {
	body := string(bodyBytes)
	return matchOperator(cond.Operator, body, cond.Value, len(bodyBytes) > 0)
}

// matchOperator applies an operator to compare actual vs expected values.
func matchOperator(operator, actual, expected string, exists bool) bool {
	switch operator {
	case "equals":
		return actual == expected
	case "contains":
		return strings.Contains(actual, expected)
	case "regex":
		re, err := regexp.Compile(expected)
		if err != nil {
			return false
		}
		return re.MatchString(actual)
	case "exists":
		return exists
	case "not_exists":
		return !exists
	default:
		return actual == expected // default to equals
	}
}

// MatchPattern checks if a URL matches a pattern with * wildcards
func MatchPattern(url, pattern string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}

	// Split pattern by *
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		// No wildcard, exact match
		return url == pattern
	}

	// Check if URL matches pattern with wildcards
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
			// First part must match at start
			return false
		}
		pos += idx + len(part)
	}

	// If pattern doesn't end with *, URL must end exactly
	if !strings.HasSuffix(pattern, "*") && pos != len(url) {
		return false
	}

	return true
}

func (p *ProxyServer) simulateLatency() {
	p.mu.Lock()
	latency := p.latency
	p.mu.Unlock()

	if latency > 0 {
		time.Sleep(latency)
	}
}

func (p *ProxyServer) debugLog(format string, args ...interface{}) {
	// Disk IO is slow. Perform it asynchronously to never block the proxy pipeline.
	msg := fmt.Sprintf("[%s] ", time.Now().Format("15:04:05.000")) + fmt.Sprintf(format, args...) + "\n"
	go func() {
		_ = os.MkdirAll(".log", 0755)
		logPath := filepath.Join(".log", "proxy_debug.log")
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()
		f.WriteString(msg)
	}()
}

// RequestLog contains details about a proxied request
type RequestLog struct {
	Id            string              `json:"id"`
	Time          string              `json:"time"`
	ClientIP      string              `json:"clientIp"`
	Method        string              `json:"method"`
	URL           string              `json:"url"`
	IsHTTPS       bool                `json:"isHttps"`
	Headers       map[string][]string `json:"headers"`     // Request Headers
	Body          string              `json:"previewBody"` // Request Body
	RespHeaders   map[string][]string `json:"respHeaders"` // Response Headers
	RespBody      string              `json:"respBody"`    // Response Body
	RespBodyRaw   []byte              `json:"-"`           // Raw response body bytes (for binary/protobuf decoding, not serialized)
	ReqBodyRaw    []byte              `json:"-"`           // Raw request body bytes (for binary/protobuf decoding, not serialized)
	StatusCode    int                 `json:"statusCode"`
	ContentType   string              `json:"contentType"`
	BodySize      int64               `json:"bodySize"`
	IsWs          bool                `json:"isWs"`
	PartialUpdate bool                `json:"partialUpdate"` // If true, only update specific fields in UI
	Mocked        bool                `json:"mocked"`        // If true, response was from mock rule
}

var currentProxy *ProxyServer
var proxyLock sync.Mutex

func GetProxy() *ProxyServer {
	proxyLock.Lock()
	defer proxyLock.Unlock()
	if currentProxy == nil {
		currentProxy = &ProxyServer{
			wsEnabled:   true, // Default ON
			mitmEnabled: true, // Default ON to match UI
			mitmBypassPatterns: []string{
				"cdn", "static", "img", "image", "video", "asset",
				"akamai", "byte", "tos-", "mon.", "snssdk",
			},
			reqBodyCache: make(map[string]cachedReqBody),
		}
	}
	return currentProxy
}

func (p *ProxyServer) Start(port int, onRequest func(RequestLog)) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("proxy already running")
	}
	p.OnRequest = onRequest
	p.port = port
	p.hasDecryptedHTTPS = false // Reset on start

	// Initialize CertManager
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".adbGUI")
	_ = os.MkdirAll(dataDir, 0755)
	p.certMgr = NewCertManager(dataDir)
	if err := p.certMgr.EnsureCert(); err != nil {
		fmt.Fprintf(os.Stderr, "[Proxy] Warning: Failed to ensure CA cert: %v\n", err)
	}
	if err := p.certMgr.LoadToGoproxy(); err != nil {
		return fmt.Errorf("failed to load CA cert: %v", err)
	}

	// Initialize goproxy
	p.proxy = goproxy.NewProxyHttpServer()
	p.proxy.Verbose = false

	// CRITICAL: Disable automatic decompression.
	// Go's DefaultTransport automatically decompresses gzip, which breaks
	// binary transparency for Apps that expect original compressed bytes.
	// CRITICAL: Disable automatic decompression.
	// This ensures the proxy is a transparent pipe. We handle decompression
	// for UI display separately in the background to avoid interfering with
	// the actual data stream.
	p.proxy.Tr = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    true, // CRITICAL: Stop Go from decompressing and changing headers
		ForceAttemptHTTP2:     true, // Support H2 to match modern app expectations
	}

	// Configure Connect Logic (MITM vs Tunnel)
	p.proxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		p.debugLog("CONNECT: %s (Session: %d)", host, ctx.Session)
		p.simulateLatency() // Simulate network RTT for connection

		p.mu.Lock()
		mitm := p.mitmEnabled
		p.mu.Unlock()

		if mitm {
			// 🚀 DYNAMIC HEURISTIC:
			// Use user-configurable patterns to bypass MITM for sensitive domains (CDNs, etc.)
			hostLower := strings.ToLower(host)

			p.mu.Lock()
			patterns := p.mitmBypassPatterns
			p.mu.Unlock()

			shouldBypass := false
			for _, pat := range patterns {
				if strings.Contains(hostLower, pat) {
					shouldBypass = true
					break
				}
			}

			if shouldBypass {
				p.debugLog("  -> PASS-THROUGH (Bypassed by pattern %s)", host)
				return goproxy.OkConnect, host
			}

			p.debugLog("  -> MITM (Decrypting %s)", host)
			return goproxy.MitmConnect, host
		}

		// Fallback for non-MITM mode: Apply rate limits via Hijack
		p.debugLog("  -> HIJACK (Tunneling %s with rate limit)", host)
		return &goproxy.ConnectAction{
			Action: goproxy.ConnectHijack,
			Hijack: p.handleHijackConnect,
		}, host
	})

	// Configure Request Handling (WS Check, Request Rate Limit)
	p.proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		id := fmt.Sprintf("%d-%d", ctx.Session, time.Now().UnixNano())
		ctx.UserData = id // Sync ID to response

		p.debugLog("REQ: %s %s (Session: %d, ID: %s)", r.Method, r.URL.String(), ctx.Session, id)
		p.simulateLatency()

		p.mu.Lock()
		wsAllowed := p.wsEnabled
		upLimiter := p.upLimiter
		p.mu.Unlock()

		// Log request immediately (notify=false: StatusCode=0, always filtered by bridge)
		p.logRequest(id, r, nil, false)

		// 1. Transparent Request Body Copy
		// Use a Tee-like wrapper to capture the request body as it's sent to the server.
		if r.Body != nil && r.Body != http.NoBody {
			r.Body = &TransparentReadCloser{
				rc:       r.Body,
				p:        p,
				id:       id,
				isReq:    true,
				captured: new(bytes.Buffer),
				limit:    100 * 1024 * 1024, // Full capture (100MB per request)
			}
		}

		// Check WS Block
		if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" && !wsAllowed {
			p.debugLog("  -> WS BLOCKED")
			return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusForbidden, "WebSocket disabled")
		}

		// Apply Upload Limit to Request Body
		if upLimiter != nil && r.Body != nil {
			p.debugLog("  -> APPLY UPLOAD LIMIT")
			r.Body = &RateLimitedReadCloser{
				rc:      r.Body,
				limiter: upLimiter,
			}
		}

		// Read request body for condition matching (only if needed)
		var mockBodyBytes []byte
		if p.hasBodyConditions() && r.Body != nil && r.Body != http.NoBody {
			mockBodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(mockBodyBytes)) // reset for upstream
		}

		// Check for mock rules
		if mockRule := p.matchMockRule(r, mockBodyBytes); mockRule != nil {
			fmt.Fprintf(os.Stderr, "[MOCK DEBUG] >>> RETURNING MOCK RESPONSE for %s %s (Rule: %s, Status: %d)\n", r.Method, r.URL.String(), mockRule.ID, mockRule.StatusCode)
			p.debugLog("  -> MOCK RESPONSE (Rule: %s)", mockRule.ID)

			// Mark as mocked in UserData
			ctx.UserData = id + "|mocked"

			// Apply mock delay
			if mockRule.Delay > 0 {
				time.Sleep(time.Duration(mockRule.Delay) * time.Millisecond)
			}

			// Create mock response
			mockResp := &http.Response{
				StatusCode: mockRule.StatusCode,
				Status:     fmt.Sprintf("%d %s", mockRule.StatusCode, http.StatusText(mockRule.StatusCode)),
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(mockRule.Body)),
				Request:    r,
			}

			// Set mock headers
			for k, v := range mockRule.Headers {
				mockResp.Header.Set(k, v)
			}

			// Set Content-Length if not set
			if mockResp.Header.Get("Content-Length") == "" {
				mockResp.Header.Set("Content-Length", fmt.Sprintf("%d", len(mockRule.Body)))
			}

			return r, mockResp
		}

		return r, nil
	})

	// Configure Response Handling (Logging, Download Rate Limit)
	p.proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		userData, _ := ctx.UserData.(string)
		id := userData
		mocked := false
		if strings.HasSuffix(userData, "|mocked") {
			id = strings.TrimSuffix(userData, "|mocked")
			mocked = true
		}
		if id == "" {
			id = fmt.Sprintf("%d-%d", ctx.Session, time.Now().UnixNano())
		}

		req := resp.Request
		if req == nil {
			req = ctx.Req
		}

		p.debugLog("RESP: %d for %s (Proto: %s, Type: %s, Mocked: %v)", resp.StatusCode, id, resp.Proto, resp.Header.Get("Content-Type"), mocked)

		// Track if we've successfully decrypted HTTPS traffic
		// This is used to verify certificate trust status
		if req != nil && req.URL != nil && req.URL.Scheme == "https" && resp.StatusCode > 0 {
			p.mu.Lock()
			p.hasDecryptedHTTPS = true
			p.mu.Unlock()
		}

		// Determine if TransparentReadCloser will handle the final OnRequest callback.
		// If so, logRequest must NOT notify (to avoid duplicate events entering DB).
		isWS := resp.StatusCode == 101 || (req != nil && strings.ToLower(req.Header.Get("Upgrade")) == "websocket")
		willWrapBody := resp.Body != nil && !isWS && (req == nil || req.Method != "CONNECT")

		// Update log with response headers
		// notify=!willWrapBody: only fire OnRequest if TransparentReadCloser won't do it
		log := p.logRequest(id, req, resp, !willWrapBody)
		log.Mocked = mocked
		_ = log // Ensure used if Body is nil

		p.mu.Lock()
		downLimiter := p.downLimiter
		p.mu.Unlock()

		if resp.Body != nil {
			rc := resp.Body

			// 2. Universal Transparent Mirroring
			// We wrap EVERYTHING to ensure we have size monitoring and logs.
			// But we use a 'Shadow Capture' limit to avoid messing with heavy binaries.

			contentType := resp.Header.Get("Content-Type")
			isBinary := strings.Contains(contentType, "image/") ||
				strings.Contains(contentType, "video/") ||
				strings.Contains(contentType, "audio/")

			captureLimit := 100 * 1024 * 1024 // 100MB: Practically full capture for API
			if isBinary {
				captureLimit = 0 // Mirrors the flow without storing any bytes
			}

			if !isWS && (req == nil || req.Method != "CONNECT") {
				rc = &TransparentReadCloser{
					rc:       rc,
					p:        p,
					id:       id,
					isReq:    false,
					log:      log,
					captured: new(bytes.Buffer),
					limit:    int64(captureLimit),
				}
			}

			if downLimiter != nil {
				rc = &RateLimitedReadCloser{
					rc:      rc,
					limiter: downLimiter,
				}
			}
			resp.Body = rc
		}

		return resp
	})

	p.mu.Unlock()

	// Bind to localhost only for security - external access via adb reverse
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	p.listener = ln // Raw listener now

	p.server = &http.Server{
		Handler: p.proxy,
	}

	p.mu.Lock()
	p.running = true
	fmt.Fprintf(os.Stderr, "[PROXY DEBUG] Proxy started: running=%v, port=%d, proxy=%p\n", p.running, p.port, p)
	p.mu.Unlock()

	go func() {
		err := p.server.Serve(p.listener)
		fmt.Fprintf(os.Stderr, "[PROXY DEBUG] Serve() returned: err=%v, proxy=%p\n", err, p)
		p.mu.Lock()
		p.running = false
		p.mu.Unlock()
	}()

	return nil
}

// handleHijackConnect implements a custom TCP tunnel to support rate limiting without MITM
func (p *ProxyServer) handleHijackConnect(req *http.Request, clientConn net.Conn, ctx *goproxy.ProxyCtx) {
	// 1. Dial destination
	destConn, err := net.DialTimeout("tcp", req.Host, 10*time.Second)
	if err != nil {
		clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		clientConn.Close()
		return
	}

	// 2. Respond 200 OK to client to establish tunnel
	clientConn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

	// 3. Rate Limiters
	p.mu.Lock()
	up := p.upLimiter
	down := p.downLimiter
	p.mu.Unlock()

	// 4. Bidirectional Copy
	// Client -> Dest (Upload)
	go p.transfer(destConn, clientConn, up)
	// Dest -> Client (Download)
	go p.transfer(clientConn, destConn, down)
}

func (p *ProxyServer) transfer(dst net.Conn, src net.Conn, limiter *rate.Limiter) {
	defer dst.Close()
	defer src.Close()

	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			// Apply Limit
			if limiter != nil {
				ctx := context.Background()
				burst := limiter.Burst()
				remaining := n
				for remaining > 0 {
					take := remaining
					if take > burst {
						take = burst
					}
					if err := limiter.WaitN(ctx, take); err != nil {
						// context canceled or error
						return
					}
					remaining -= take
				}
			}
			// Write
			if _, wErr := dst.Write(buf[:n]); wErr != nil {
				break
			}
		}
		if err != nil {
			break
		}
	}
}

func (p *ProxyServer) logRequest(id string, r *http.Request, resp *http.Response, notify bool) RequestLog {
	if p.OnRequest == nil || r == nil {
		return RequestLog{}
	}

	host := ""
	if r.RemoteAddr != "" {
		h, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			host = h
		} else {
			host = r.RemoteAddr
		}
	}

	isHTTPS := false
	if r.URL != nil {
		isHTTPS = r.URL.Scheme == "https"
	}

	method := r.Method
	if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
		method = "WS"
	}

	// Clone request headers to avoid data race
	reqHeaders := copyHeader(r.Header)
	urlStr := ""
	if r.URL != nil {
		urlStr = r.URL.String()
	}

	// Extract Response details
	statusCode := 0
	var contentType string
	var bodySize int64
	var respHeaders map[string][]string

	if resp != nil {
		statusCode = resp.StatusCode
		contentType = resp.Header.Get("Content-Type")
		bodySize = resp.ContentLength
		respHeaders = copyHeader(resp.Header)
		// Don't cap bodySize at 0 if it's -1 (chunked)
	} else if method == "CONNECT" {
		statusCode = 200 // Connection Established
	}

	log := RequestLog{
		Id:          id,
		Time:        time.Now().Format("2006-01-02 15:04:05"),
		ClientIP:    host,
		Method:      method,
		URL:         urlStr,
		IsHTTPS:     isHTTPS,
		Headers:     reqHeaders,
		RespHeaders: respHeaders,
		StatusCode:  statusCode,
		ContentType: contentType,
		BodySize:    bodySize,
		IsWs:        method == "WS",
	}

	// Emit log to UI in background, only if notify is true.
	// When TransparentReadCloser will wrap the response body, it sends the
	// final (complete) callback itself, so logRequest should NOT notify to
	// avoid duplicate events.
	go func() {
		if p.OnRequest != nil && notify {
			p.OnRequest(log)
		}
	}()

	return log
}

func (p *ProxyServer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.running || p.server == nil {
		return nil
	}
	// Clear request body cache
	p.reqBodyCacheMu.Lock()
	p.reqBodyCache = make(map[string]cachedReqBody)
	p.reqBodyCacheMu.Unlock()
	return p.server.Shutdown(context.Background())
}

func (p *ProxyServer) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(os.Stderr, "[PROXY DEBUG] IsRunning called: running=%v, port=%d, proxy=%p\n", p.running, p.port, p)
	return p.running
}

func (p *ProxyServer) SetLimits(uploadSpeed, downloadSpeed int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Low burst to ensure even small requests feel the throttle
	// But enough to handle basic TCP headers/segments efficiently
	// 4KB burst means we enforce check every 4KB
	const burstSize = 4 * 1024

	if uploadSpeed > 0 {
		p.upLimiter = rate.NewLimiter(rate.Limit(uploadSpeed), burstSize)
	} else {
		p.upLimiter = nil
	}

	if downloadSpeed > 0 {
		p.downLimiter = rate.NewLimiter(rate.Limit(downloadSpeed), burstSize)
	} else {
		p.downLimiter = nil
	}
}

func (p *ProxyServer) SetWSEnabled(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.wsEnabled = enabled
}

func (p *ProxyServer) IsWSEnabled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.wsEnabled
}

func (p *ProxyServer) SetProxyMITM(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mitmEnabled = enabled
	p.debugLog("PROXY MITM: %v", enabled)
}

func (p *ProxyServer) SetMITMBypassPatterns(patterns []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mitmBypassPatterns = patterns
	p.debugLog("PROXY MITM Bypass Patterns Updated: %v", patterns)
}

func (p *ProxyServer) GetMITMBypassPatterns() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mitmBypassPatterns
}

func (p *ProxyServer) IsMITMEnabled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mitmEnabled
}

func (p *ProxyServer) GetCertPath() string {
	if p.certMgr != nil {
		return p.certMgr.CertPath
	}
	return ""
}

// HasRecentDecryptedHTTPS checks if there are any successfully decrypted HTTPS requests
func (p *ProxyServer) HasRecentDecryptedHTTPS() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.hasDecryptedHTTPS
}

// --- Helpers ---

// RateLimitedReadCloser wraps an io.ReadCloser with rate limiting
type RateLimitedReadCloser struct {
	rc      io.ReadCloser
	limiter *rate.Limiter
}

func (r *RateLimitedReadCloser) Read(p []byte) (n int, err error) {
	n, err = r.rc.Read(p)
	if n > 0 && r.limiter != nil {
		ctx := context.Background()
		burst := r.limiter.Burst()
		remaining := n
		for remaining > 0 {
			take := remaining
			if take > burst {
				take = burst
			}
			if wErr := r.limiter.WaitN(ctx, take); wErr != nil {
				// Should we fail? Yes, probably context canceled.
				break
			}
			remaining -= take
		}
	}
	return
}

func (r *RateLimitedReadCloser) Close() error {
	return r.rc.Close()
}

// MultiReadCloser aids in restoring the body after peeking
type MultiReadCloser struct {
	io.Reader
	Closer io.Closer
}

func (m *MultiReadCloser) Close() error {
	return m.Closer.Close()
}

// TransparentReadCloser captures body and counts size during transfer without affecting the stream
type TransparentReadCloser struct {
	rc         io.ReadCloser
	p          *ProxyServer
	id         string
	isReq      bool
	log        RequestLog // Cached log metadata for response updates
	captured   *bytes.Buffer
	limit      int64
	totalSize  int64
	lastUpdate time.Time
	doneCalled bool // prevents double update(true) from Read EOF + Close
	mu         sync.Mutex
}

func (r *TransparentReadCloser) Read(p []byte) (n int, err error) {
	n, err = r.rc.Read(p)
	if n > 0 {
		r.mu.Lock()
		r.totalSize += int64(n)
		// Capture copy for analysis (up to limit)
		if r.limit > 0 && int64(r.captured.Len()) < r.limit {
			toWrite := n
			if int64(r.captured.Len())+int64(toWrite) > r.limit {
				toWrite = int(r.limit - int64(r.captured.Len()))
			}
			r.captured.Write(p[:toWrite])
		}

		// Update UI infrequently (every 512KB) or on timer
		shouldUpdate := r.totalSize%(512*1024) < int64(n) || time.Since(r.lastUpdate) > 2*time.Second
		r.mu.Unlock()

		if shouldUpdate {
			r.update(false)
		}
	}
	if err == io.EOF {
		r.update(true)
	}
	return n, err
}

func (r *TransparentReadCloser) update(done bool) {
	r.mu.Lock()
	if done {
		if r.doneCalled {
			r.mu.Unlock()
			return
		}
		r.doneCalled = true
	}
	r.lastUpdate = time.Now()
	size := r.totalSize

	var capturedCopy []byte
	// Only capture on final update or if limit reached
	if (done || (r.limit > 0 && int64(r.captured.Len()) >= r.limit)) && r.captured.Len() > 0 {
		capturedCopy = append([]byte(nil), r.captured.Bytes()...)
	}
	r.mu.Unlock()

	// Update UI asynchronously in background
	go func() {
		if r.p.OnRequest == nil {
			return
		}

		if r.isReq {
			// Store captured request body into cache for the response handler to pick up.
			// We no longer send a PartialUpdate here; instead the response's final
			// update will include the request body from the cache.
			if len(capturedCopy) > 0 {
				result := r.p.analyzeBodyFull(capturedCopy, "", "")
				r.p.reqBodyCacheMu.Lock()
				r.p.reqBodyCache[r.id] = cachedReqBody{
					text:     result.Text,
					rawBytes: result.RawBytes, // non-nil for binary (e.g. protobuf)
				}
				r.p.reqBodyCacheMu.Unlock()
			}
		} else {
			// For Response:
			// If not done, only send a PartialUpdate with the current size to avoid overwriting headers
			if !done {
				r.p.OnRequest(RequestLog{
					Id:            r.id,
					BodySize:      size,
					PartialUpdate: true,
				})
				return
			}

			// Final update: Send the full log (or a meaningful subset)
			logToSend := r.log
			logToSend.BodySize = size
			if len(capturedCopy) > 0 {
				encoding := ""
				if vv, ok := logToSend.RespHeaders["Content-Encoding"]; ok && len(vv) > 0 {
					encoding = vv[0]
				}
				result := r.p.analyzeBodyFull(capturedCopy, encoding, logToSend.ContentType)
				logToSend.RespBody = result.Text
				if result.IsBinary {
					logToSend.RespBodyRaw = result.RawBytes
				}
			}

			// Merge cached request body (captured by the request TransparentReadCloser)
			r.p.reqBodyCacheMu.Lock()
			if cached, ok := r.p.reqBodyCache[r.id]; ok {
				logToSend.Body = cached.text
				if len(cached.rawBytes) > 0 {
					logToSend.ReqBodyRaw = cached.rawBytes
				}
				delete(r.p.reqBodyCache, r.id)
			}
			r.p.reqBodyCacheMu.Unlock()

			r.p.OnRequest(logToSend)
		}
	}()
}

func (r *TransparentReadCloser) Close() error {
	r.update(true)
	return r.rc.Close()
}

// AnalyzedBody holds the result of body analysis.
type AnalyzedBody struct {
	Text     string // String representation for display
	RawBytes []byte // Raw decompressed bytes (set when binary data detected)
	IsBinary bool   // Whether the data is binary
}

// analyzeBody handles decompression and string conversion for UI display in a non-interfering way.
func (p *ProxyServer) analyzeBody(data []byte, encoding string, contentType string) string {
	result := p.analyzeBodyFull(data, encoding, contentType)
	return result.Text
}

// analyzeBodyFull handles decompression and returns both text and raw bytes.
// Raw bytes are preserved when binary data is detected (e.g. protobuf).
func (p *ProxyServer) analyzeBodyFull(data []byte, encoding string, contentType string) AnalyzedBody {
	if len(data) == 0 {
		return AnalyzedBody{}
	}

	raw := data
	// 1. Decompress if needed (Shadow copy only)
	if strings.Contains(encoding, "gzip") {
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err == nil {
			decompressed, err := io.ReadAll(gr)
			if err == nil {
				raw = decompressed
			}
			gr.Close()
		}
	}

	// 2. Binary Detection
	isBinary := false
	limit := len(raw)
	if limit > 512 {
		limit = 512
	}
	for i := 0; i < limit; i++ {
		if raw[i] == 0 {
			isBinary = true
			break
		}
	}

	if isBinary {
		return AnalyzedBody{
			Text:     fmt.Sprintf("[Binary Data: %d bytes]", len(raw)),
			RawBytes: append([]byte(nil), raw...), // copy
			IsBinary: true,
		}
	}

	// 3. String Truncation removed: Send full string to UI as requested
	return AnalyzedBody{
		Text: string(raw),
	}
}
