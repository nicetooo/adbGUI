package mcp

import (
	"errors"
	"sync"
)

// MockCall records a method call for verification
type MockCall struct {
	Method string
	Args   []interface{}
}

// MockGazeApp is a mock implementation of GazeApp for testing
type MockGazeApp struct {
	mu    sync.Mutex
	Calls []MockCall

	// Device Management
	GetDevicesResult       []Device
	GetDevicesError        error
	GetDeviceInfoResult    DeviceInfo
	GetDeviceInfoError     error
	AdbConnectResult       string
	AdbConnectError        error
	AdbDisconnectResult    string
	AdbDisconnectError     error
	AdbPairResult          string
	AdbPairError           error
	SwitchToWirelessResult string
	SwitchToWirelessError  error
	GetDeviceIPResult      string
	GetDeviceIPError       error

	// App Management
	ListPackagesResult []AppPackage
	ListPackagesError  error
	GetAppInfoResult   AppPackage
	GetAppInfoError    error
	StartAppResult     string
	StartAppError      error
	ForceStopAppResult string
	ForceStopAppError  error
	InstallAPKResult   string
	InstallAPKError    error
	UninstallAppResult string
	UninstallAppError  error
	ClearAppDataResult string
	ClearAppDataError  error
	IsAppRunningResult bool
	IsAppRunningError  error

	// Screen Control
	TakeScreenshotResult string
	TakeScreenshotError  error
	StartRecordingError  error
	StopRecordingError   error
	IsRecordingResult    bool

	// UI Automation
	GetUIHierarchyResult      *UIHierarchyResult
	GetUIHierarchyError       error
	SearchUIElementsResult    []map[string]interface{}
	SearchUIElementsError     error
	PerformNodeActionError    error
	GetDeviceResolutionResult string
	GetDeviceResolutionError  error

	// Session Management
	CreateSessionResult          string
	StartSessionWithConfigResult string
	EndSessionError              error
	GetActiveSessionResult       string
	ListStoredSessionsResult     []DeviceSession
	ListStoredSessionsError      error
	QuerySessionEventsResult     *EventQueryResult
	QuerySessionEventsError      error
	GetSessionStatsResult        map[string]interface{}
	GetSessionStatsError         error

	// Workflow
	LoadWorkflowsResult          []Workflow
	LoadWorkflowsError           error
	GetWorkflowResult            *Workflow
	GetWorkflowError             error
	SaveWorkflowError            error
	DeleteWorkflowError          error
	RunWorkflowError             error
	ExecuteSingleStepError       error
	IsWorkflowRunningResult      bool
	WorkflowExecutionResultValue *WorkflowExecutionResult
	// StopWorkflow has no return

	// Proxy
	StartProxyResult     string
	StartProxyError      error
	StopProxyResult      string
	StopProxyError       error
	GetProxyStatusResult bool

	// Video
	GetVideoFrameResult       string
	GetVideoFrameError        error
	GetVideoMetadataResult    *VideoMetadata
	GetVideoMetadataError     error
	GetSessionVideoInfoResult map[string]interface{}
	GetSessionVideoInfoError  error

	// Utility
	AppVersion string
}

// NewMockGazeApp creates a new MockGazeApp with sensible defaults
func NewMockGazeApp() *MockGazeApp {
	return &MockGazeApp{
		Calls:      make([]MockCall, 0),
		AppVersion: "1.0.0-test",
		// Default empty results
		GetDevicesResult:         []Device{},
		ListPackagesResult:       []AppPackage{},
		LoadWorkflowsResult:      []Workflow{},
		ListStoredSessionsResult: []DeviceSession{},
		SearchUIElementsResult:   []map[string]interface{}{},
	}
}

// recordCall records a method call
func (m *MockGazeApp) recordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{Method: method, Args: args})
}

// GetCalls returns all recorded calls
func (m *MockGazeApp) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MockCall{}, m.Calls...)
}

// ResetCalls clears all recorded calls
func (m *MockGazeApp) ResetCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]MockCall, 0)
}

// GetLastCall returns the last recorded call
func (m *MockGazeApp) GetLastCall() *MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Calls) == 0 {
		return nil
	}
	return &m.Calls[len(m.Calls)-1]
}

// WasMethodCalled checks if a method was called
func (m *MockGazeApp) WasMethodCalled(method string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.Calls {
		if call.Method == method {
			return true
		}
	}
	return false
}

// GetLastCallByMethod returns the last call to a specific method
func (m *MockGazeApp) GetLastCallByMethod(method string) *MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.Calls) - 1; i >= 0; i-- {
		if m.Calls[i].Method == method {
			return &m.Calls[i]
		}
	}
	return nil
}

// === Device Management ===

func (m *MockGazeApp) GetDevices(forceLog bool) ([]Device, error) {
	m.recordCall("GetDevices", forceLog)
	return m.GetDevicesResult, m.GetDevicesError
}

func (m *MockGazeApp) GetDeviceInfo(deviceId string) (DeviceInfo, error) {
	m.recordCall("GetDeviceInfo", deviceId)
	return m.GetDeviceInfoResult, m.GetDeviceInfoError
}

func (m *MockGazeApp) AdbConnect(address string) (string, error) {
	m.recordCall("AdbConnect", address)
	return m.AdbConnectResult, m.AdbConnectError
}

func (m *MockGazeApp) AdbDisconnect(address string) (string, error) {
	m.recordCall("AdbDisconnect", address)
	return m.AdbDisconnectResult, m.AdbDisconnectError
}

func (m *MockGazeApp) AdbPair(address string, code string) (string, error) {
	m.recordCall("AdbPair", address, code)
	return m.AdbPairResult, m.AdbPairError
}

func (m *MockGazeApp) SwitchToWireless(deviceId string) (string, error) {
	m.recordCall("SwitchToWireless", deviceId)
	return m.SwitchToWirelessResult, m.SwitchToWirelessError
}

func (m *MockGazeApp) GetDeviceIP(deviceId string) (string, error) {
	m.recordCall("GetDeviceIP", deviceId)
	return m.GetDeviceIPResult, m.GetDeviceIPError
}

// === App Management ===

func (m *MockGazeApp) ListPackages(deviceId string, packageType string) ([]AppPackage, error) {
	m.recordCall("ListPackages", deviceId, packageType)
	return m.ListPackagesResult, m.ListPackagesError
}

func (m *MockGazeApp) GetAppInfo(deviceId, packageName string, force bool) (AppPackage, error) {
	m.recordCall("GetAppInfo", deviceId, packageName, force)
	return m.GetAppInfoResult, m.GetAppInfoError
}

func (m *MockGazeApp) StartApp(deviceId, packageName string) (string, error) {
	m.recordCall("StartApp", deviceId, packageName)
	return m.StartAppResult, m.StartAppError
}

func (m *MockGazeApp) ForceStopApp(deviceId, packageName string) (string, error) {
	m.recordCall("ForceStopApp", deviceId, packageName)
	return m.ForceStopAppResult, m.ForceStopAppError
}

func (m *MockGazeApp) InstallAPK(deviceId string, path string) (string, error) {
	m.recordCall("InstallAPK", deviceId, path)
	return m.InstallAPKResult, m.InstallAPKError
}

func (m *MockGazeApp) UninstallApp(deviceId, packageName string) (string, error) {
	m.recordCall("UninstallApp", deviceId, packageName)
	return m.UninstallAppResult, m.UninstallAppError
}

func (m *MockGazeApp) ClearAppData(deviceId, packageName string) (string, error) {
	m.recordCall("ClearAppData", deviceId, packageName)
	return m.ClearAppDataResult, m.ClearAppDataError
}

func (m *MockGazeApp) IsAppRunning(deviceId, packageName string) (bool, error) {
	m.recordCall("IsAppRunning", deviceId, packageName)
	return m.IsAppRunningResult, m.IsAppRunningError
}

// === Screen Control ===

func (m *MockGazeApp) TakeScreenshot(deviceId, savePath string) (string, error) {
	m.recordCall("TakeScreenshot", deviceId, savePath)
	return m.TakeScreenshotResult, m.TakeScreenshotError
}

func (m *MockGazeApp) StartRecording(deviceId string, config ScrcpyConfig) error {
	m.recordCall("StartRecording", deviceId, config)
	return m.StartRecordingError
}

func (m *MockGazeApp) StopRecording(deviceId string) error {
	m.recordCall("StopRecording", deviceId)
	return m.StopRecordingError
}

func (m *MockGazeApp) IsRecording(deviceId string) bool {
	m.recordCall("IsRecording", deviceId)
	return m.IsRecordingResult
}

// === UI Automation ===

func (m *MockGazeApp) GetUIHierarchy(deviceId string) (*UIHierarchyResult, error) {
	m.recordCall("GetUIHierarchy", deviceId)
	return m.GetUIHierarchyResult, m.GetUIHierarchyError
}

func (m *MockGazeApp) SearchUIElements(deviceId string, query string) ([]map[string]interface{}, error) {
	m.recordCall("SearchUIElements", deviceId, query)
	return m.SearchUIElementsResult, m.SearchUIElementsError
}

func (m *MockGazeApp) PerformNodeAction(deviceId string, bounds string, actionType string) error {
	m.recordCall("PerformNodeAction", deviceId, bounds, actionType)
	return m.PerformNodeActionError
}

func (m *MockGazeApp) GetDeviceResolution(deviceId string) (string, error) {
	m.recordCall("GetDeviceResolution", deviceId)
	return m.GetDeviceResolutionResult, m.GetDeviceResolutionError
}

// === Session Management ===

func (m *MockGazeApp) CreateSession(deviceId, sessionType, name string) string {
	m.recordCall("CreateSession", deviceId, sessionType, name)
	return m.CreateSessionResult
}

func (m *MockGazeApp) StartSessionWithConfig(deviceId, name string, config MCPSessionConfig) string {
	m.recordCall("StartSessionWithConfig", deviceId, name, config)
	if m.StartSessionWithConfigResult != "" {
		return m.StartSessionWithConfigResult
	}
	return m.CreateSessionResult
}

func (m *MockGazeApp) EndSession(sessionId string, status string) error {
	m.recordCall("EndSession", sessionId, status)
	return m.EndSessionError
}

func (m *MockGazeApp) GetActiveSession(deviceId string) string {
	m.recordCall("GetActiveSession", deviceId)
	return m.GetActiveSessionResult
}

func (m *MockGazeApp) ListStoredSessions(deviceID string, limit int) ([]DeviceSession, error) {
	m.recordCall("ListStoredSessions", deviceID, limit)
	return m.ListStoredSessionsResult, m.ListStoredSessionsError
}

func (m *MockGazeApp) QuerySessionEvents(query EventQuery) (*EventQueryResult, error) {
	m.recordCall("QuerySessionEvents", query)
	return m.QuerySessionEventsResult, m.QuerySessionEventsError
}

func (m *MockGazeApp) GetSessionStats(sessionID string) (map[string]interface{}, error) {
	m.recordCall("GetSessionStats", sessionID)
	return m.GetSessionStatsResult, m.GetSessionStatsError
}

// === Workflow ===

func (m *MockGazeApp) LoadWorkflows() ([]Workflow, error) {
	m.recordCall("LoadWorkflows")
	return m.LoadWorkflowsResult, m.LoadWorkflowsError
}

func (m *MockGazeApp) GetWorkflow(workflowID string) (*Workflow, error) {
	m.recordCall("GetWorkflow", workflowID)
	if m.GetWorkflowError != nil {
		return nil, m.GetWorkflowError
	}
	if m.GetWorkflowResult != nil {
		return m.GetWorkflowResult, nil
	}
	// Fallback: search in LoadWorkflowsResult
	for _, wf := range m.LoadWorkflowsResult {
		if wf.ID == workflowID {
			return &wf, nil
		}
	}
	return nil, errors.New("workflow not found")
}

func (m *MockGazeApp) SaveWorkflow(workflow Workflow) error {
	m.recordCall("SaveWorkflow", workflow)
	return m.SaveWorkflowError
}

func (m *MockGazeApp) DeleteWorkflow(id string) error {
	m.recordCall("DeleteWorkflow", id)
	return m.DeleteWorkflowError
}

func (m *MockGazeApp) RunWorkflow(device Device, workflow Workflow) error {
	m.recordCall("RunWorkflow", device, workflow)
	return m.RunWorkflowError
}

func (m *MockGazeApp) StopWorkflow(device Device) {
	m.recordCall("StopWorkflow", device)
}

func (m *MockGazeApp) PauseTask(deviceId string) {
	m.recordCall("PauseTask", deviceId)
}

func (m *MockGazeApp) ResumeTask(deviceId string) {
	m.recordCall("ResumeTask", deviceId)
}

func (m *MockGazeApp) ExecuteSingleWorkflowStep(deviceId string, step WorkflowStep) error {
	m.recordCall("ExecuteSingleWorkflowStep", deviceId, step)
	return m.ExecuteSingleStepError
}

func (m *MockGazeApp) IsWorkflowRunning(deviceId string) bool {
	m.recordCall("IsWorkflowRunning", deviceId)
	return m.IsWorkflowRunningResult
}

func (m *MockGazeApp) GetWorkflowExecutionResult(deviceId string) *WorkflowExecutionResult {
	m.recordCall("GetWorkflowExecutionResult", deviceId)
	return m.WorkflowExecutionResultValue
}

func (m *MockGazeApp) StepNextWorkflow(deviceId string) (*WorkflowExecutionResult, error) {
	m.recordCall("StepNextWorkflow", deviceId)
	return m.WorkflowExecutionResultValue, nil
}

// === Proxy ===

func (m *MockGazeApp) StartProxy(port int) (string, error) {
	m.recordCall("StartProxy", port)
	return m.StartProxyResult, m.StartProxyError
}

func (m *MockGazeApp) StopProxy() (string, error) {
	m.recordCall("StopProxy")
	return m.StopProxyResult, m.StopProxyError
}

func (m *MockGazeApp) GetProxyStatus() bool {
	m.recordCall("GetProxyStatus")
	return m.GetProxyStatusResult
}

// === Video ===

func (m *MockGazeApp) GetVideoFrame(videoPath string, timeMs int64, width int) (string, error) {
	m.recordCall("GetVideoFrame", videoPath, timeMs, width)
	return m.GetVideoFrameResult, m.GetVideoFrameError
}

func (m *MockGazeApp) GetVideoMetadata(videoPath string) (*VideoMetadata, error) {
	m.recordCall("GetVideoMetadata", videoPath)
	return m.GetVideoMetadataResult, m.GetVideoMetadataError
}

func (m *MockGazeApp) GetSessionVideoInfo(sessionID string) (map[string]interface{}, error) {
	m.recordCall("GetSessionVideoInfo", sessionID)
	return m.GetSessionVideoInfoResult, m.GetSessionVideoInfoError
}

// === ADB ===

func (m *MockGazeApp) RunAdbCommand(deviceId string, command string) (string, error) {
	m.recordCall("RunAdbCommand", deviceId, command)
	return "", nil
}

// === CLI Tools ===

func (m *MockGazeApp) RunAaptCommand(command string, timeoutSec int) (string, error) {
	m.recordCall("RunAaptCommand", command, timeoutSec)
	return "", nil
}

func (m *MockGazeApp) RunFfmpegCommand(command string, timeoutSec int) (string, error) {
	m.recordCall("RunFfmpegCommand", command, timeoutSec)
	return "", nil
}

func (m *MockGazeApp) RunFfprobeCommand(command string, timeoutSec int) (string, error) {
	m.recordCall("RunFfprobeCommand", command, timeoutSec)
	return "", nil
}

// === Utility ===

func (m *MockGazeApp) GetAppVersion() string {
	m.recordCall("GetAppVersion")
	return m.AppVersion
}

// === Test Helper Functions ===

// SetupWithDevices configures mock with sample devices
func (m *MockGazeApp) SetupWithDevices(devices ...Device) *MockGazeApp {
	m.GetDevicesResult = devices
	return m
}

// SetupWithError configures a specific method to return an error
func (m *MockGazeApp) SetupWithError(method string, err error) *MockGazeApp {
	switch method {
	case "GetDevices":
		m.GetDevicesError = err
	case "GetDeviceInfo":
		m.GetDeviceInfoError = err
	case "AdbConnect":
		m.AdbConnectError = err
	case "AdbDisconnect":
		m.AdbDisconnectError = err
	case "AdbPair":
		m.AdbPairError = err
	case "SwitchToWireless":
		m.SwitchToWirelessError = err
	case "GetDeviceIP":
		m.GetDeviceIPError = err
	case "ListPackages":
		m.ListPackagesError = err
	case "GetAppInfo":
		m.GetAppInfoError = err
	case "StartApp":
		m.StartAppError = err
	case "ForceStopApp":
		m.ForceStopAppError = err
	case "InstallAPK":
		m.InstallAPKError = err
	case "UninstallApp":
		m.UninstallAppError = err
	case "ClearAppData":
		m.ClearAppDataError = err
	case "IsAppRunning":
		m.IsAppRunningError = err
	case "TakeScreenshot":
		m.TakeScreenshotError = err
	case "StartRecording":
		m.StartRecordingError = err
	case "StopRecording":
		m.StopRecordingError = err
	case "GetUIHierarchy":
		m.GetUIHierarchyError = err
	case "SearchUIElements":
		m.SearchUIElementsError = err
	case "PerformNodeAction":
		m.PerformNodeActionError = err
	case "GetDeviceResolution":
		m.GetDeviceResolutionError = err
	case "EndSession":
		m.EndSessionError = err
	case "ListStoredSessions":
		m.ListStoredSessionsError = err
	case "QuerySessionEvents":
		m.QuerySessionEventsError = err
	case "GetSessionStats":
		m.GetSessionStatsError = err
	case "LoadWorkflows":
		m.LoadWorkflowsError = err
	case "RunWorkflow":
		m.RunWorkflowError = err
	case "StartProxy":
		m.StartProxyError = err
	case "StopProxy":
		m.StopProxyError = err
	}
	return m
}

// Common test errors
var (
	ErrDeviceNotFound      = errors.New("device not found")
	ErrDeviceOffline       = errors.New("device offline")
	ErrAppNotFound         = errors.New("app not found")
	ErrAppNotRunning       = errors.New("app not running")
	ErrSessionNotFound     = errors.New("session not found")
	ErrWorkflowNotFound    = errors.New("workflow not found")
	ErrProxyAlreadyRunning = errors.New("proxy already running")
	ErrProxyNotRunning     = errors.New("proxy not running")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrTimeout             = errors.New("operation timed out")
)

// Sample test data factories

// SampleDevice returns a sample device for testing
func SampleDevice(id string) Device {
	return Device{
		ID:         id,
		Serial:     id,
		State:      "device",
		Model:      "Pixel 6",
		Brand:      "Google",
		Type:       "wired",
		IDs:        []string{id},
		LastActive: 1700000000000,
	}
}

// SampleDeviceInfo returns sample device info for testing
func SampleDeviceInfo() DeviceInfo {
	return DeviceInfo{
		Model:        "Pixel 6",
		Brand:        "Google",
		Manufacturer: "Google",
		AndroidVer:   "14",
		SDK:          "34",
		ABI:          "arm64-v8a",
		Serial:       "abc123",
		Resolution:   "1080x2400",
		Density:      "420",
		CPU:          "Tensor",
		Memory:       "8GB",
		Props:        map[string]string{"ro.build.id": "AP1A.240405.002"},
	}
}

// SampleAppPackage returns a sample app package for testing
func SampleAppPackage(name string) AppPackage {
	return AppPackage{
		Name:             name,
		Label:            "Sample App",
		Type:             "user",
		State:            "enabled",
		VersionName:      "1.0.0",
		VersionCode:      "1",
		TargetSdkVersion: "34",
		Permissions:      []string{"android.permission.INTERNET"},
		Activities:       []string{name + ".MainActivity"},
	}
}

// SampleWorkflow returns a sample workflow for testing
func SampleWorkflow(id, name string) Workflow {
	return Workflow{
		ID:          id,
		Name:        name,
		Description: "Test workflow",
		Version:     2,
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-01T00:00:00Z",
		Steps: []WorkflowStep{
			{
				ID:   "start",
				Type: "start",
				Name: "Start",
				Common: StepCommon{
					OnError: "stop",
					Loop:    1,
				},
				Connections: StepConnections{
					SuccessStepId: "step_1",
				},
				Layout: StepLayout{PosX: 20, PosY: 20},
			},
			{
				ID:   "step_1",
				Type: "tap",
				Name: "Tap button",
				Tap:  &TapParams{X: 540, Y: 960},
				Common: StepCommon{
					Timeout:   5000,
					PostDelay: 500,
					OnError:   "stop",
					Loop:      1,
				},
				Connections: StepConnections{},
				Layout:      StepLayout{PosX: 20, PosY: 180},
			},
		},
	}
}

// SampleSession returns a sample session for testing
func SampleSession(id, deviceID string) DeviceSession {
	return DeviceSession{
		ID:        id,
		DeviceID:  deviceID,
		Name:      "Test Session",
		Type:      "manual",
		Status:    "active",
		StartTime: 1700000000000,
	}
}
