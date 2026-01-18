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
	WorkflowExecutionResult = types.WorkflowExecutionResult
)

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

	// Session Management
	CreateSession(deviceId, sessionType, name string) string
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

	// Proxy
	StartProxy(port int) (string, error)
	StopProxy() (string, error)
	GetProxyStatus() bool

	// Video
	GetVideoFrame(videoPath string, timeMs int64, width int) (string, error)
	GetVideoMetadata(videoPath string) (*VideoMetadata, error)
	GetSessionVideoInfo(sessionID string) (map[string]interface{}, error)

	// ADB
	RunAdbCommand(deviceId string, command string) (string, error)

	// Utility
	GetAppVersion() string
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
