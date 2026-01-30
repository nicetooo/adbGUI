// Package mcp provides MCP (Model Context Protocol) server implementation for Gaze
// This allows external AI clients (like Claude Desktop) to interact with Android devices
package mcp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"Gaze/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Type aliases from shared types package
// This avoids code duplication and ensures type consistency
type (
	Device            = types.Device
	DeviceInfo        = types.DeviceInfo
	AppPackage        = types.AppPackage
	ScrcpyConfig      = types.ScrcpyConfig
	UIHierarchyResult = types.UIHierarchyResult
	EventQuery        = types.EventQuery
	EventQueryResult  = types.EventQueryResult
	DeviceSession     = types.DeviceSession
	VideoMetadata     = types.VideoMetadata

	// Workflow types
	Workflow                = types.Workflow
	WorkflowStep            = types.WorkflowStep
	StepConnections         = types.StepConnections
	StepCommon              = types.StepCommon
	StepLayout              = types.StepLayout
	HandleInfo              = types.HandleInfo
	ElementSelector         = types.ElementSelector
	TapParams               = types.TapParams
	SwipeParams             = types.SwipeParams
	ElementParams           = types.ElementParams
	AppParams               = types.AppParams
	BranchParams            = types.BranchParams
	WaitParams              = types.WaitParams
	ScriptParams            = types.ScriptParams
	VariableParams          = types.VariableParams
	ADBParams               = types.ADBParams
	SubWorkflowParams       = types.SubWorkflowParams
	ReadToVariableParams    = types.ReadToVariableParams
	SessionParams           = types.SessionParams
	WorkflowExecutionResult = types.WorkflowExecutionResult
)

// PerfMonitorConfig is the performance monitor configuration for MCP interface
type PerfMonitorConfig struct {
	PackageName   string `json:"packageName,omitempty"`
	IntervalMs    int    `json:"intervalMs"`
	EnableCPU     bool   `json:"enableCPU"`
	EnableMemory  bool   `json:"enableMemory"`
	EnableFPS     bool   `json:"enableFPS"`
	EnableNetwork bool   `json:"enableNetwork"`
	EnableBattery bool   `json:"enableBattery"`
}

// ProcessPerfData represents a single process performance entry for MCP interface
type ProcessPerfData struct {
	PID       int     `json:"pid"`
	Name      string  `json:"name"`
	CPU       float64 `json:"cpu"`
	MemoryKB  int     `json:"memoryKB"`
	User      float64 `json:"user"`
	Kernel    float64 `json:"kernel"`
	LinuxUser string  `json:"linuxUser"`
	PPID      int     `json:"ppid"`
	VSZKB     int     `json:"vszKB"`
	State     string  `json:"state"`
}

// ProcessMemoryCategory represents a memory category from dumpsys meminfo App Summary
type ProcessMemoryCategory struct {
	Name  string `json:"name"`
	PssKB int    `json:"pssKB"`
	RssKB int    `json:"rssKB"`
}

// ProcessObjects represents Android object counts from dumpsys meminfo
type ProcessObjects struct {
	Views           int `json:"views"`
	ViewRootImpl    int `json:"viewRootImpl"`
	AppContexts     int `json:"appContexts"`
	Activities      int `json:"activities"`
	Assets          int `json:"assets"`
	AssetManagers   int `json:"assetManagers"`
	LocalBinders    int `json:"localBinders"`
	ProxyBinders    int `json:"proxyBinders"`
	DeathRecipients int `json:"deathRecipients"`
	WebViews        int `json:"webViews"`
}

// ProcessDetail represents detailed process info for MCP interface
type ProcessDetail struct {
	PID               int                     `json:"pid"`
	PackageName       string                  `json:"packageName"`
	TotalPSSKB        int                     `json:"totalPssKB"`
	TotalRSSKB        int                     `json:"totalRssKB"`
	SwapPSSKB         int                     `json:"swapPssKB"`
	Memory            []ProcessMemoryCategory `json:"memory"`
	JavaHeapSizeKB    int                     `json:"javaHeapSizeKB"`
	JavaHeapAllocKB   int                     `json:"javaHeapAllocKB"`
	JavaHeapFreeKB    int                     `json:"javaHeapFreeKB"`
	NativeHeapSizeKB  int                     `json:"nativeHeapSizeKB"`
	NativeHeapAllocKB int                     `json:"nativeHeapAllocKB"`
	NativeHeapFreeKB  int                     `json:"nativeHeapFreeKB"`
	Objects           ProcessObjects          `json:"objects"`
	Threads           int                     `json:"threads"`
	FDSize            int                     `json:"fdSize"`
	VmSwapKB          int                     `json:"vmSwapKB"`
	OomScoreAdj       int                     `json:"oomScoreAdj"`
	UID               int                     `json:"uid"`
}

// PerfSampleData is a performance sample snapshot for MCP interface
type PerfSampleData struct {
	CPUUsage     float64           `json:"cpuUsage"`
	CPUApp       float64           `json:"cpuApp"`
	CPUCores     int               `json:"cpuCores"`
	CPUFreqMHz   int               `json:"cpuFreqMHz"`
	CPUTempC     float64           `json:"cpuTempC"`
	MemTotalMB   int               `json:"memTotalMB"`
	MemUsedMB    int               `json:"memUsedMB"`
	MemFreeMB    int               `json:"memFreeMB"`
	MemUsage     float64           `json:"memUsage"`
	MemAppMB     int               `json:"memAppMB"`
	FPS          float64           `json:"fps"`
	JankCount    int               `json:"jankCount"`
	NetRxKBps    float64           `json:"netRxKBps"`
	NetTxKBps    float64           `json:"netTxKBps"`
	NetRxTotalMB float64           `json:"netRxTotalMB"`
	NetTxTotalMB float64           `json:"netTxTotalMB"`
	BatteryLevel int               `json:"batteryLevel"`
	BatteryTemp  float64           `json:"batteryTemp"`
	PackageName  string            `json:"packageName,omitempty"`
	Processes    []ProcessPerfData `json:"processes,omitempty"`
}

// MCPSessionConfig is a simplified session config for MCP interface
// This avoids coupling with internal event_types.go definitions
type MCPSessionConfig struct {
	// Logcat config
	LogcatEnabled       bool   `json:"logcatEnabled,omitempty"`
	LogcatPackageName   string `json:"logcatPackageName,omitempty"`
	LogcatPreFilter     string `json:"logcatPreFilter,omitempty"`
	LogcatExcludeFilter string `json:"logcatExcludeFilter,omitempty"`

	// Recording config
	RecordingEnabled bool   `json:"recordingEnabled,omitempty"`
	RecordingQuality string `json:"recordingQuality,omitempty"` // "low", "medium", "high"

	// Proxy config
	ProxyEnabled     bool `json:"proxyEnabled,omitempty"`
	ProxyPort        int  `json:"proxyPort,omitempty"`
	ProxyMitmEnabled bool `json:"proxyMitmEnabled,omitempty"`

	// Monitor config
	MonitorEnabled bool `json:"monitorEnabled,omitempty"`
}

// GazeApp interface defines the methods that MCP server needs from the main App
// This allows loose coupling between MCP and the main application
type GazeApp interface {
	// Device Management
	GetDevices(forceLog bool) ([]Device, error)
	GetDeviceInfo(deviceId string) (DeviceInfo, error)
	AdbConnect(address string) (string, error)
	AdbDisconnect(address string) (string, error)
	AdbPair(address string, code string) (string, error)
	SwitchToWireless(deviceId string) (string, error)
	GetDeviceIP(deviceId string) (string, error)

	// App Management
	ListPackages(deviceId string, packageType string) ([]AppPackage, error)
	GetAppInfo(deviceId, packageName string, force bool) (AppPackage, error)
	StartApp(deviceId, packageName string) (string, error)
	ForceStopApp(deviceId, packageName string) (string, error)
	InstallAPK(deviceId string, path string) (string, error)
	UninstallApp(deviceId, packageName string) (string, error)
	ClearAppData(deviceId, packageName string) (string, error)
	IsAppRunning(deviceId, packageName string) (bool, error)

	// Screen Control
	TakeScreenshot(deviceId, savePath string) (string, error)
	StartRecording(deviceId string, config ScrcpyConfig) error
	StopRecording(deviceId string) error
	IsRecording(deviceId string) bool

	// UI Automation
	GetUIHierarchy(deviceId string) (*UIHierarchyResult, error)
	SearchUIElements(deviceId string, query string) ([]map[string]interface{}, error)
	PerformNodeAction(deviceId string, bounds string, actionType string) error
	GetDeviceResolution(deviceId string) (string, error)
	InputText(deviceId string, text string) error
	EnsureADBKeyboard(deviceId string) (bool, bool, error)
	IsADBKeyboardInstalled(deviceId string) bool

	// Session Management
	CreateSession(deviceId, sessionType, name string) string
	StartSessionWithConfig(deviceId, name string, config MCPSessionConfig) string
	EndSession(sessionId string, status string) error
	GetActiveSession(deviceId string) string
	ListStoredSessions(deviceID string, limit int) ([]DeviceSession, error)
	QuerySessionEvents(query EventQuery) (*EventQueryResult, error)
	GetSessionStats(sessionID string) (map[string]interface{}, error)

	// Workflow
	LoadWorkflows() ([]Workflow, error)
	GetWorkflow(workflowID string) (*Workflow, error)
	SaveWorkflow(workflow Workflow) error
	DeleteWorkflow(id string) error
	RunWorkflow(device Device, workflow Workflow) error
	StopWorkflow(device Device)
	PauseTask(deviceId string)
	ResumeTask(deviceId string)
	ExecuteSingleWorkflowStep(deviceId string, step WorkflowStep) error
	IsWorkflowRunning(deviceId string) bool
	GetWorkflowExecutionResult(deviceId string) *WorkflowExecutionResult
	StepNextWorkflow(deviceId string) (*WorkflowExecutionResult, error)

	// Proxy
	StartProxy(port int) (string, error)
	StopProxy() (string, error)
	GetProxyStatus() bool
	SetProxyDevice(deviceId string)
	GetProxyDevice() string
	SetupProxyForDevice(deviceId string, port int) error
	CleanupProxyForDevice(deviceId string, port int) error
	SetProxyMITM(enabled bool)
	SetProxyWSEnabled(enabled bool)
	SetProxyLimit(uploadSpeed, downloadSpeed int)
	SetProxyLatency(latencyMs int)
	SetMITMBypassPatterns(patterns []string)
	GetMITMBypassPatterns() []string
	GetProxySettings() map[string]interface{}
	InstallProxyCert(deviceId string) (string, error)
	CheckCertTrust(deviceId string) string

	// Mock Rules
	AddMockRule(urlPattern, method string, statusCode int, headers map[string]string, body string, delay int, description string, conditions []MCPMockCondition) string
	UpdateMockRule(id, urlPattern, method string, statusCode int, headers map[string]string, body string, delay int, enabled bool, description string, conditions []MCPMockCondition) error
	RemoveMockRule(ruleID string)
	GetMockRules() []MCPMockRule
	ToggleMockRule(ruleID string, enabled bool) error
	ResendRequest(method, url string, headers map[string]string, body string) (map[string]interface{}, error)

	// Breakpoint Rules
	AddBreakpointRule(urlPattern, method, phase, description string) string
	UpdateBreakpointRule(id, urlPattern, method, phase string, enabled bool, description string) error
	RemoveBreakpointRule(ruleID string)
	GetBreakpointRules() []MCPBreakpointRule
	ToggleBreakpointRule(ruleID string, enabled bool) error
	ResolveBreakpoint(breakpointID string, action string, modifications map[string]interface{}) error
	GetPendingBreakpoints() []MCPPendingBreakpointInfo
	ForwardAllBreakpoints()

	// Video
	GetVideoFrame(videoPath string, timeMs int64, width int) (string, error)
	GetVideoMetadata(videoPath string) (*VideoMetadata, error)
	GetSessionVideoInfo(sessionID string) (map[string]interface{}, error)

	// ADB
	RunAdbCommand(deviceId string, command string) (string, error)

	// CLI Tools
	RunAaptCommand(command string, timeoutSec int) (string, error)
	RunFfmpegCommand(command string, timeoutSec int) (string, error)
	RunFfprobeCommand(command string, timeoutSec int) (string, error)

	// File Management
	UploadFile(deviceId, localPath, remotePath string) error
	ListFiles(deviceId, pathStr string) ([]map[string]interface{}, error)

	// Session Export/Import
	ExportSessionToPath(sessionID, outputPath string) (string, error)
	ImportSessionFromPath(inputPath string) (string, error)

	// Performance Monitoring
	StartPerfMonitor(deviceId string, config PerfMonitorConfig) string
	StopPerfMonitor(deviceId string) string
	IsPerfMonitorRunning(deviceId string) bool
	GetPerfSnapshot(deviceId string, packageName string) (*PerfSampleData, error)
	GetProcessDetail(deviceId string, pid int) (*ProcessDetail, error)

	// Protobuf Management
	AddProtoFile(name, content string) (string, error)
	UpdateProtoFile(id, name, content string) error
	RemoveProtoFile(id string) error
	GetProtoFiles() []MCPProtoFile
	AddProtoMapping(urlPattern, messageType, direction, description string) (string, error)
	UpdateProtoMapping(id, urlPattern, messageType, direction, description string) error
	RemoveProtoMapping(id string) error
	GetProtoMappings() []MCPProtoMapping
	GetProtoMessageTypes() []string
	LoadProtoFromURL(rawURL string) ([]string, error)

	// Utility
	GetAppVersion() string
}

// MCPProtoFile represents a .proto file entry for MCP interface
type MCPProtoFile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	LoadedAt int64  `json:"loadedAt"`
}

// MCPProtoMapping represents a URL→message type mapping for MCP interface
type MCPProtoMapping struct {
	ID          string `json:"id"`
	URLPattern  string `json:"urlPattern"`
	MessageType string `json:"messageType"`
	Direction   string `json:"direction"`
	Description string `json:"description"`
}

// MCPMockCondition represents a conditional match for mock rules
type MCPMockCondition struct {
	Type     string `json:"type"`     // "header", "query", "body"
	Key      string `json:"key"`      // header name or query param name (unused for body type)
	Operator string `json:"operator"` // "equals", "contains", "regex", "exists", "not_exists"
	Value    string `json:"value"`    // expected value (unused for exists/not_exists)
}

// MCPMockRule represents a mock rule for MCP interface
type MCPMockRule struct {
	ID          string             `json:"id"`
	URLPattern  string             `json:"urlPattern"`
	Method      string             `json:"method"`
	StatusCode  int                `json:"statusCode"`
	Headers     map[string]string  `json:"headers"`
	Body        string             `json:"body"`
	Delay       int                `json:"delay"`
	Enabled     bool               `json:"enabled"`
	Description string             `json:"description"`
	Conditions  []MCPMockCondition `json:"conditions,omitempty"`
}

// MCPBreakpointRule represents a breakpoint rule for MCP interface
type MCPBreakpointRule struct {
	ID          string `json:"id"`
	URLPattern  string `json:"urlPattern"`
	Method      string `json:"method"` // empty = match all
	Phase       string `json:"phase"`  // "request", "response", "both"
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"createdAt"`
}

// MCPPendingBreakpointInfo represents a pending breakpoint for MCP interface
type MCPPendingBreakpointInfo struct {
	ID     string `json:"id"`
	RuleID string `json:"ruleId"`
	Phase  string `json:"phase"` // "request" or "response"

	// Request info
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    string              `json:"body,omitempty"`

	// Response info (only for response phase)
	StatusCode  int                 `json:"statusCode,omitempty"`
	RespHeaders map[string][]string `json:"respHeaders,omitempty"`
	RespBody    string              `json:"respBody,omitempty"`

	CreatedAt int64 `json:"createdAt"` // unix ms
}

// MCPServer wraps the MCP server and provides Gaze-specific functionality
type MCPServer struct {
	app       GazeApp
	server    *server.MCPServer
	stdio     *server.StdioServer
	mu        sync.Mutex
	isRunning bool
}

// NewMCPServer creates a new MCP server for Gaze
func NewMCPServer(app GazeApp) *MCPServer {
	mcpServer := server.NewMCPServer(
		"gaze-android-manager",
		app.GetAppVersion(),
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithElicitation(), // Enable elicitation for dangerous operations
		server.WithLogging(),
	)

	s := &MCPServer{
		app:    app,
		server: mcpServer,
	}

	// Register all tools
	s.registerTools()

	// Register resources
	s.registerResources()

	return s
}

// registerTools registers all MCP tools
func (s *MCPServer) registerTools() {
	// Device Management Tools
	s.registerDeviceTools()

	// App Management Tools
	s.registerAppTools()

	// Screen Control Tools
	s.registerScreenTools()

	// UI Automation Tools
	s.registerAutomationTools()

	// Session Management Tools
	s.registerSessionTools()

	// Workflow Tools
	s.registerWorkflowTools()

	// Proxy Tools
	s.registerProxyTools()

	// Video Tools
	s.registerVideoTools()

	// Performance Monitoring Tools
	s.registerPerfTools()

	// Protobuf Management Tools
	s.registerProtoTools()
}

// registerResources registers all MCP resources
func (s *MCPServer) registerResources() {
	// Device list resource
	s.server.AddResource(
		mcp.NewResource(
			"gaze://devices",
			"Connected Android devices",
			mcp.WithMIMEType("application/json"),
		),
		s.handleDevicesResource,
	)

	// Device info resource template
	s.server.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"gaze://devices/{deviceId}",
			"Device information",
		),
		s.handleDeviceInfoResource,
	)

	// Session list resource
	s.server.AddResource(
		mcp.NewResource(
			"gaze://sessions",
			"Active and recent sessions",
			mcp.WithMIMEType("application/json"),
		),
		s.handleSessionsResource,
	)

	// Workflow list resource
	s.server.AddResource(
		mcp.NewResource(
			"workflow://list",
			"All saved workflows",
			mcp.WithMIMEType("application/json"),
		),
		s.handleWorkflowsResource,
	)

	// Individual workflow resource template
	s.server.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"workflow://{workflowId}",
			"Workflow details",
		),
		s.handleWorkflowResource,
	)
}

// Start starts the MCP server (blocking - for CLI mode)
// This method blocks until the server shuts down
func (s *MCPServer) Start() error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("MCP server is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	return s.run()
}

// StartAsync starts the MCP server in a goroutine (non-blocking)
func (s *MCPServer) StartAsync() error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("MCP server is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	go s.run()
	return nil
}

// run runs the MCP server (blocking)
func (s *MCPServer) run() error {
	s.stdio = server.NewStdioServer(s.server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		cancel()
	}()

	fmt.Fprintln(os.Stderr, "[MCP] Gaze MCP Server started")
	err := s.stdio.Listen(ctx, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[MCP] Server error: %v\n", err)
	}

	s.mu.Lock()
	s.isRunning = false
	s.mu.Unlock()

	return err
}

// Stop stops the MCP server
func (s *MCPServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// The server will stop when stdin is closed or context is cancelled
	s.isRunning = false
}

// IsRunning returns whether the MCP server is running
func (s *MCPServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}

// requestConfirmation requests user confirmation for dangerous operations
func (s *MCPServer) requestConfirmation(ctx context.Context, operation, details string) (bool, error) {
	elicitationRequest := mcp.ElicitationRequest{
		Params: mcp.ElicitationParams{
			Message: fmt.Sprintf("⚠️ Dangerous Operation: %s\n\nDetails: %s\n\nDo you want to proceed?", operation, details),
			RequestedSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"confirm": map[string]any{
						"type":        "boolean",
						"description": "Confirm to proceed with this operation",
					},
				},
				"required": []string{"confirm"},
			},
		},
	}

	result, err := s.server.RequestElicitation(ctx, elicitationRequest)
	if err != nil {
		return false, fmt.Errorf("failed to request confirmation: %w", err)
	}

	if result.Action != mcp.ElicitationResponseActionAccept {
		return false, nil
	}

	data, ok := result.Content.(map[string]any)
	if !ok {
		return false, fmt.Errorf("unexpected response format")
	}

	confirm, ok := data["confirm"].(bool)
	if !ok {
		return false, fmt.Errorf("invalid confirmation response")
	}

	return confirm, nil
}
