package main

import (
	"encoding/json"

	"Gaze/mcp"
)

// MCPBridge bridges the main App to the MCP server
type MCPBridge struct {
	app *App
}

// NewMCPBridge creates a new MCP bridge
func NewMCPBridge(app *App) *MCPBridge {
	return &MCPBridge{app: app}
}

// Implement mcp.GazeApp interface

func (b *MCPBridge) GetDevices(forceLog bool) ([]mcp.Device, error) {
	devices, err := b.app.GetDevices(forceLog)
	if err != nil {
		return nil, err
	}
	result := make([]mcp.Device, len(devices))
	for i, d := range devices {
		result[i] = mcp.Device{
			ID:         d.ID,
			Serial:     d.Serial,
			State:      d.State,
			Model:      d.Model,
			Brand:      d.Brand,
			Type:       d.Type,
			IDs:        d.IDs,
			WifiAddr:   d.WifiAddr,
			LastActive: d.LastActive,
			IsPinned:   d.IsPinned,
		}
	}
	return result, nil
}

func (b *MCPBridge) GetDeviceInfo(deviceId string) (mcp.DeviceInfo, error) {
	info, err := b.app.GetDeviceInfo(deviceId)
	if err != nil {
		return mcp.DeviceInfo{}, err
	}
	return mcp.DeviceInfo{
		Model:        info.Model,
		Brand:        info.Brand,
		Manufacturer: info.Manufacturer,
		AndroidVer:   info.AndroidVer,
		SDK:          info.SDK,
		ABI:          info.ABI,
		Serial:       info.Serial,
		Resolution:   info.Resolution,
		Density:      info.Density,
		CPU:          info.CPU,
		Memory:       info.Memory,
		Props:        info.Props,
	}, nil
}

func (b *MCPBridge) AdbConnect(address string) (string, error) {
	return b.app.AdbConnect(address)
}

func (b *MCPBridge) AdbDisconnect(address string) (string, error) {
	return b.app.AdbDisconnect(address)
}

func (b *MCPBridge) AdbPair(address string, code string) (string, error) {
	return b.app.AdbPair(address, code)
}

func (b *MCPBridge) SwitchToWireless(deviceId string) (string, error) {
	return b.app.SwitchToWireless(deviceId)
}

func (b *MCPBridge) GetDeviceIP(deviceId string) (string, error) {
	return b.app.GetDeviceIP(deviceId)
}

func (b *MCPBridge) ListPackages(deviceId string, packageType string) ([]mcp.AppPackage, error) {
	packages, err := b.app.ListPackages(deviceId, packageType)
	if err != nil {
		return nil, err
	}
	result := make([]mcp.AppPackage, len(packages))
	for i, p := range packages {
		result[i] = mcp.AppPackage{
			Name:                 p.Name,
			Label:                p.Label,
			Icon:                 p.Icon,
			Type:                 p.Type,
			State:                p.State,
			VersionName:          p.VersionName,
			VersionCode:          p.VersionCode,
			MinSdkVersion:        p.MinSdkVersion,
			TargetSdkVersion:     p.TargetSdkVersion,
			Permissions:          p.Permissions,
			Activities:           p.Activities,
			LaunchableActivities: p.LaunchableActivities,
		}
	}
	return result, nil
}

func (b *MCPBridge) GetAppInfo(deviceId, packageName string, force bool) (mcp.AppPackage, error) {
	info, err := b.app.GetAppInfo(deviceId, packageName, force)
	if err != nil {
		return mcp.AppPackage{}, err
	}
	return mcp.AppPackage{
		Name:                 info.Name,
		Label:                info.Label,
		Icon:                 info.Icon,
		Type:                 info.Type,
		State:                info.State,
		VersionName:          info.VersionName,
		VersionCode:          info.VersionCode,
		MinSdkVersion:        info.MinSdkVersion,
		TargetSdkVersion:     info.TargetSdkVersion,
		Permissions:          info.Permissions,
		Activities:           info.Activities,
		LaunchableActivities: info.LaunchableActivities,
	}, nil
}

func (b *MCPBridge) StartApp(deviceId, packageName string) (string, error) {
	return b.app.StartApp(deviceId, packageName)
}

func (b *MCPBridge) ForceStopApp(deviceId, packageName string) (string, error) {
	return b.app.ForceStopApp(deviceId, packageName)
}

func (b *MCPBridge) InstallAPK(deviceId string, path string) (string, error) {
	return b.app.InstallAPK(deviceId, path)
}

func (b *MCPBridge) UninstallApp(deviceId, packageName string) (string, error) {
	return b.app.UninstallApp(deviceId, packageName)
}

func (b *MCPBridge) ClearAppData(deviceId, packageName string) (string, error) {
	return b.app.ClearAppData(deviceId, packageName)
}

func (b *MCPBridge) IsAppRunning(deviceId, packageName string) (bool, error) {
	return b.app.IsAppRunning(deviceId, packageName)
}

func (b *MCPBridge) TakeScreenshot(deviceId, savePath string) (string, error) {
	return b.app.TakeScreenshot(deviceId, savePath)
}

func (b *MCPBridge) StartRecording(deviceId string, config mcp.ScrcpyConfig) error {
	return b.app.StartRecording(deviceId, ScrcpyConfig{
		MaxSize:     config.MaxSize,
		BitRate:     config.BitRate,
		MaxFps:      config.MaxFps,
		ShowTouches: config.ShowTouches,
	})
}

func (b *MCPBridge) StopRecording(deviceId string) error {
	return b.app.StopRecording(deviceId)
}

func (b *MCPBridge) IsRecording(deviceId string) bool {
	return b.app.IsRecording(deviceId)
}

func (b *MCPBridge) GetUIHierarchy(deviceId string) (*mcp.UIHierarchyResult, error) {
	result, err := b.app.GetUIHierarchy(deviceId)
	if err != nil {
		return nil, err
	}
	return &mcp.UIHierarchyResult{
		Root:   result.Root,
		RawXML: result.RawXML,
	}, nil
}

func (b *MCPBridge) SearchUIElements(deviceId string, query string) ([]map[string]interface{}, error) {
	return b.app.SearchUIElements(deviceId, query)
}

func (b *MCPBridge) PerformNodeAction(deviceId string, bounds string, actionType string) error {
	return b.app.PerformNodeAction(deviceId, bounds, actionType)
}

func (b *MCPBridge) GetDeviceResolution(deviceId string) (string, error) {
	return b.app.GetDeviceResolution(deviceId)
}

func (b *MCPBridge) CreateSession(deviceId, sessionType, name string) string {
	// Use StartNewSession which persists to EventStore (database)
	return b.app.StartNewSession(deviceId, sessionType, name)
}

func (b *MCPBridge) EndSession(sessionId string, status string) error {
	// Use EndActiveSession which properly closes via EventPipeline
	b.app.EndActiveSession(sessionId, status)
	return nil
}

func (b *MCPBridge) GetActiveSession(deviceId string) string {
	// Use GetDeviceActiveSession which reads from EventPipeline
	session := b.app.GetDeviceActiveSession(deviceId)
	if session != nil {
		return session.ID
	}
	return ""
}

func (b *MCPBridge) ListStoredSessions(deviceID string, limit int) ([]mcp.DeviceSession, error) {
	sessions, err := b.app.ListStoredSessions(deviceID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]mcp.DeviceSession, len(sessions))
	for i, s := range sessions {
		result[i] = mcp.DeviceSession{
			ID:            s.ID,
			DeviceID:      s.DeviceID,
			Name:          s.Name,
			Type:          s.Type,
			Status:        s.Status,
			StartTime:     s.StartTime,
			EndTime:       s.EndTime,
			EventCount:    s.EventCount,
			VideoPath:     s.VideoPath,
			VideoDuration: s.VideoDuration,
		}
	}
	return result, nil
}

func (b *MCPBridge) QuerySessionEvents(query mcp.EventQuery) (*mcp.EventQueryResult, error) {
	// Convert string arrays to typed arrays
	sources := make([]EventSource, len(query.Sources))
	for i, s := range query.Sources {
		sources[i] = EventSource(s)
	}
	levels := make([]EventLevel, len(query.Levels))
	for i, l := range query.Levels {
		levels[i] = EventLevel(l)
	}

	result, err := b.app.QuerySessionEvents(EventQuery{
		SessionID:   query.SessionID,
		DeviceID:    query.DeviceID,
		Types:       query.Types,
		Sources:     sources,
		Levels:      levels,
		StartTime:   query.StartTime,
		EndTime:     query.EndTime,
		SearchText:  query.SearchText,
		Limit:       query.Limit,
		Offset:      query.Offset,
		IncludeData: true, // MCP needs full event data including coordinates
	})
	if err != nil {
		return nil, err
	}
	events := make([]interface{}, len(result.Events))
	for i, e := range result.Events {
		// Convert UnifiedEvent to map for MCP serialization
		eventMap := map[string]interface{}{
			"id":           e.ID,
			"sessionId":    e.SessionID,
			"deviceId":     e.DeviceID,
			"timestamp":    e.Timestamp,
			"relativeTime": e.RelativeTime,
			"source":       string(e.Source),
			"category":     string(e.Category),
			"type":         e.Type,
			"level":        string(e.Level),
			"title":        e.Title,
			"summary":      e.Summary,
		}
		// Parse data JSON if present
		if len(e.Data) > 0 {
			var data map[string]interface{}
			if err := json.Unmarshal(e.Data, &data); err == nil {
				eventMap["data"] = data
			}
		}
		if e.Duration > 0 {
			eventMap["duration"] = e.Duration
		}
		events[i] = eventMap
	}
	return &mcp.EventQueryResult{
		Events:  events,
		Total:   result.Total,
		HasMore: result.HasMore,
	}, nil
}

func (b *MCPBridge) GetSessionStats(sessionID string) (map[string]interface{}, error) {
	return b.app.GetSessionStats(sessionID)
}

func (b *MCPBridge) LoadWorkflows() ([]mcp.Workflow, error) {
	workflows, err := b.app.LoadWorkflows()
	if err != nil {
		return nil, err
	}
	result := make([]mcp.Workflow, len(workflows))
	for i, wf := range workflows {
		converted := b.convertWorkflowToMCP(&wf)
		result[i] = *converted
	}
	return result, nil
}

func (b *MCPBridge) RunWorkflow(device mcp.Device, workflow mcp.Workflow) error {
	// Convert back to main types
	mainDevice := Device{
		ID:         device.ID,
		Serial:     device.Serial,
		State:      device.State,
		Model:      device.Model,
		Brand:      device.Brand,
		Type:       device.Type,
		IDs:        device.IDs,
		WifiAddr:   device.WifiAddr,
		LastActive: device.LastActive,
		IsPinned:   device.IsPinned,
	}
	mainWorkflow := b.convertWorkflowFromMCP(workflow)
	return b.app.RunWorkflow(mainDevice, mainWorkflow)
}

func (b *MCPBridge) StopWorkflow(device mcp.Device) {
	mainDevice := Device{
		ID:         device.ID,
		Serial:     device.Serial,
		State:      device.State,
		Model:      device.Model,
		Brand:      device.Brand,
		Type:       device.Type,
		IDs:        device.IDs,
		WifiAddr:   device.WifiAddr,
		LastActive: device.LastActive,
		IsPinned:   device.IsPinned,
	}
	b.app.StopWorkflow(mainDevice)
}

func (b *MCPBridge) PauseTask(deviceId string) {
	b.app.PauseTask(deviceId)
}

func (b *MCPBridge) ResumeTask(deviceId string) {
	b.app.ResumeTask(deviceId)
}

func (b *MCPBridge) GetWorkflow(workflowID string) (*mcp.Workflow, error) {
	wf, err := b.app.GetWorkflow(workflowID)
	if err != nil {
		return nil, err
	}
	return b.convertWorkflowToMCP(wf), nil
}

func (b *MCPBridge) SaveWorkflow(workflow mcp.Workflow) error {
	mainWorkflow := b.convertWorkflowFromMCP(workflow)
	return b.app.SaveWorkflow(mainWorkflow)
}

func (b *MCPBridge) DeleteWorkflow(id string) error {
	return b.app.DeleteWorkflow(id)
}

func (b *MCPBridge) ExecuteSingleWorkflowStep(deviceId string, step mcp.WorkflowStep) error {
	mainStep := b.convertStepFromMCP(step)
	return b.app.ExecuteSingleWorkflowStep(deviceId, mainStep)
}

func (b *MCPBridge) IsWorkflowRunning(deviceId string) bool {
	return b.app.IsPlayingTouch(deviceId)
}

func (b *MCPBridge) GetWorkflowExecutionResult(deviceId string) *mcp.WorkflowExecutionResult {
	result := b.app.GetWorkflowExecutionResult(deviceId)
	if result == nil {
		return nil
	}
	return &mcp.WorkflowExecutionResult{
		WorkflowID:   result.WorkflowID,
		WorkflowName: result.WorkflowName,
		Status:       result.Status,
		Error:        result.Error,
		StartTime:    result.StartTime.UnixMilli(),
		EndTime:      result.EndTime.UnixMilli(),
		Duration:     result.Duration,
		StepsTotal:   result.StepsTotal,
	}
}

// Helper functions for workflow type conversion
// Both packages now use the same nested structure format

func (b *MCPBridge) convertWorkflowToMCP(wf *Workflow) *mcp.Workflow {
	if wf == nil {
		return nil
	}
	steps := make([]mcp.WorkflowStep, len(wf.Steps))
	for i, s := range wf.Steps {
		steps[i] = b.convertStepToMCP(s)
	}
	return &mcp.Workflow{
		ID:          wf.ID,
		Name:        wf.Name,
		Description: wf.Description,
		Version:     2,
		Steps:       steps,
		Variables:   wf.Variables,
		CreatedAt:   wf.CreatedAt,
		UpdatedAt:   wf.UpdatedAt,
	}
}

func (b *MCPBridge) convertStepToMCP(s WorkflowStep) mcp.WorkflowStep {
	result := mcp.WorkflowStep{
		ID:   s.ID,
		Type: s.Type,
		Name: s.Name,
	}

	// Convert Common (value to value)
	result.Common = mcp.StepCommon{
		Timeout:   s.Common.Timeout,
		OnError:   s.Common.OnError,
		Loop:      s.Common.Loop,
		PostDelay: s.Common.PostDelay,
		PreWait:   s.Common.PreWait,
	}

	// Convert Connections (value to value)
	result.Connections = mcp.StepConnections{
		SuccessStepId: s.Connections.SuccessStepId,
		ErrorStepId:   s.Connections.ErrorStepId,
		TrueStepId:    s.Connections.TrueStepId,
		FalseStepId:   s.Connections.FalseStepId,
	}

	// Convert Layout (value to value)
	result.Layout = mcp.StepLayout{
		PosX: s.Layout.PosX,
		PosY: s.Layout.PosY,
	}
	if s.Layout.Handles != nil {
		result.Layout.Handles = make(map[string]mcp.HandleInfo)
		for k, v := range s.Layout.Handles {
			result.Layout.Handles[k] = mcp.HandleInfo{
				SourceHandle: v.SourceHandle,
				TargetHandle: v.TargetHandle,
			}
		}
	}

	// Copy type-specific params (pointer to pointer)
	if s.Tap != nil {
		result.Tap = &mcp.TapParams{X: s.Tap.X, Y: s.Tap.Y}
	}
	if s.Swipe != nil {
		result.Swipe = &mcp.SwipeParams{
			X: s.Swipe.X, Y: s.Swipe.Y, X2: s.Swipe.X2, Y2: s.Swipe.Y2,
			Direction: s.Swipe.Direction, Distance: s.Swipe.Distance, Duration: s.Swipe.Duration,
		}
	}
	if s.Element != nil {
		result.Element = &mcp.ElementParams{
			Selector:      mcp.ElementSelector{Type: s.Element.Selector.Type, Value: s.Element.Selector.Value, Index: s.Element.Selector.Index},
			Action:        s.Element.Action,
			InputText:     s.Element.InputText,
			SwipeDir:      s.Element.SwipeDir,
			SwipeDistance: s.Element.SwipeDistance,
			SwipeDuration: s.Element.SwipeDuration,
		}
	}
	if s.App != nil {
		result.App = &mcp.AppParams{PackageName: s.App.PackageName, Action: s.App.Action}
	}
	if s.Branch != nil {
		result.Branch = &mcp.BranchParams{
			Condition:     s.Branch.Condition,
			ExpectedValue: s.Branch.ExpectedValue,
			VariableName:  s.Branch.VariableName,
		}
		if s.Branch.Selector != nil {
			result.Branch.Selector = &mcp.ElementSelector{
				Type: s.Branch.Selector.Type, Value: s.Branch.Selector.Value, Index: s.Branch.Selector.Index,
			}
		}
	}
	if s.Wait != nil {
		result.Wait = &mcp.WaitParams{DurationMs: s.Wait.DurationMs}
	}
	if s.Script != nil {
		result.Script = &mcp.ScriptParams{ScriptName: s.Script.ScriptName}
	}
	if s.Variable != nil {
		result.Variable = &mcp.VariableParams{Name: s.Variable.Name, Value: s.Variable.Value}
	}
	if s.ADB != nil {
		result.ADB = &mcp.ADBParams{Command: s.ADB.Command}
	}
	if s.Workflow != nil {
		result.Workflow = &mcp.SubWorkflowParams{WorkflowId: s.Workflow.WorkflowId}
	}
	if s.ReadToVariable != nil {
		result.ReadToVariable = &mcp.ReadToVariableParams{
			Selector:     mcp.ElementSelector{Type: s.ReadToVariable.Selector.Type, Value: s.ReadToVariable.Selector.Value, Index: s.ReadToVariable.Selector.Index},
			VariableName: s.ReadToVariable.VariableName,
			Attribute:    s.ReadToVariable.Attribute,
			Regex:        s.ReadToVariable.Regex,
			DefaultValue: s.ReadToVariable.DefaultValue,
		}
	}

	return result
}

func (b *MCPBridge) convertWorkflowFromMCP(wf mcp.Workflow) Workflow {
	steps := make([]WorkflowStep, len(wf.Steps))
	for i, s := range wf.Steps {
		steps[i] = b.convertStepFromMCP(s)
	}
	return Workflow{
		ID:          wf.ID,
		Name:        wf.Name,
		Description: wf.Description,
		Steps:       steps,
		Variables:   wf.Variables,
		CreatedAt:   wf.CreatedAt,
		UpdatedAt:   wf.UpdatedAt,
	}
}

func (b *MCPBridge) convertStepFromMCP(s mcp.WorkflowStep) WorkflowStep {
	result := WorkflowStep{
		ID:   s.ID,
		Type: s.Type,
		Name: s.Name,
	}

	// Convert Common (value to value)
	result.Common = StepCommon{
		Timeout:   s.Common.Timeout,
		OnError:   s.Common.OnError,
		Loop:      s.Common.Loop,
		PostDelay: s.Common.PostDelay,
		PreWait:   s.Common.PreWait,
	}

	// Convert Connections (value to value)
	result.Connections = StepConnections{
		SuccessStepId: s.Connections.SuccessStepId,
		ErrorStepId:   s.Connections.ErrorStepId,
		TrueStepId:    s.Connections.TrueStepId,
		FalseStepId:   s.Connections.FalseStepId,
	}

	// Convert Layout (value to value)
	result.Layout = StepLayout{
		PosX: s.Layout.PosX,
		PosY: s.Layout.PosY,
	}
	if s.Layout.Handles != nil {
		result.Layout.Handles = make(map[string]HandleInfo)
		for k, v := range s.Layout.Handles {
			result.Layout.Handles[k] = HandleInfo{
				SourceHandle: v.SourceHandle,
				TargetHandle: v.TargetHandle,
			}
		}
	}

	// Copy type-specific params (pointer to pointer)
	if s.Tap != nil {
		result.Tap = &TapParams{X: s.Tap.X, Y: s.Tap.Y}
	}
	if s.Swipe != nil {
		result.Swipe = &SwipeParams{
			X: s.Swipe.X, Y: s.Swipe.Y, X2: s.Swipe.X2, Y2: s.Swipe.Y2,
			Direction: s.Swipe.Direction, Distance: s.Swipe.Distance, Duration: s.Swipe.Duration,
		}
	}
	if s.Element != nil {
		result.Element = &ElementParams{
			Selector:      ElementSelector{Type: s.Element.Selector.Type, Value: s.Element.Selector.Value, Index: s.Element.Selector.Index},
			Action:        s.Element.Action,
			InputText:     s.Element.InputText,
			SwipeDir:      s.Element.SwipeDir,
			SwipeDistance: s.Element.SwipeDistance,
			SwipeDuration: s.Element.SwipeDuration,
		}
	}
	if s.App != nil {
		result.App = &AppParams{PackageName: s.App.PackageName, Action: s.App.Action}
	}
	if s.Branch != nil {
		result.Branch = &BranchParams{
			Condition:     s.Branch.Condition,
			ExpectedValue: s.Branch.ExpectedValue,
			VariableName:  s.Branch.VariableName,
		}
		if s.Branch.Selector != nil {
			result.Branch.Selector = &ElementSelector{
				Type: s.Branch.Selector.Type, Value: s.Branch.Selector.Value, Index: s.Branch.Selector.Index,
			}
		}
	}
	if s.Wait != nil {
		result.Wait = &WaitParams{DurationMs: s.Wait.DurationMs}
	}
	if s.Script != nil {
		result.Script = &ScriptParams{ScriptName: s.Script.ScriptName}
	}
	if s.Variable != nil {
		result.Variable = &VariableParams{Name: s.Variable.Name, Value: s.Variable.Value}
	}
	if s.ADB != nil {
		result.ADB = &ADBParams{Command: s.ADB.Command}
	}
	if s.Workflow != nil {
		result.Workflow = &SubWorkflowParams{WorkflowId: s.Workflow.WorkflowId}
	}
	if s.ReadToVariable != nil {
		result.ReadToVariable = &ReadToVariableParams{
			Selector:     ElementSelector{Type: s.ReadToVariable.Selector.Type, Value: s.ReadToVariable.Selector.Value, Index: s.ReadToVariable.Selector.Index},
			VariableName: s.ReadToVariable.VariableName,
			Attribute:    s.ReadToVariable.Attribute,
			Regex:        s.ReadToVariable.Regex,
			DefaultValue: s.ReadToVariable.DefaultValue,
		}
	}

	return result
}

func (b *MCPBridge) StartProxy(port int) (string, error) {
	return b.app.StartProxy(port)
}

func (b *MCPBridge) StopProxy() (string, error) {
	return b.app.StopProxy()
}

func (b *MCPBridge) GetProxyStatus() bool {
	return b.app.GetProxyStatus()
}

func (b *MCPBridge) GetAppVersion() string {
	return b.app.GetAppVersion()
}

func (b *MCPBridge) GetVideoFrame(videoPath string, timeMs int64, width int) (string, error) {
	return b.app.GetVideoFrame(videoPath, timeMs, width)
}

func (b *MCPBridge) GetVideoMetadata(videoPath string) (*mcp.VideoMetadata, error) {
	meta, err := b.app.GetVideoMetadata(videoPath)
	if err != nil {
		return nil, err
	}
	return &mcp.VideoMetadata{
		Path:        meta.Path,
		Duration:    meta.Duration,
		DurationMs:  meta.DurationMs,
		Width:       meta.Width,
		Height:      meta.Height,
		FrameRate:   meta.FrameRate,
		Codec:       meta.Codec,
		BitRate:     meta.BitRate,
		TotalFrames: meta.TotalFrames,
	}, nil
}

func (b *MCPBridge) GetSessionVideoInfo(sessionID string) (map[string]interface{}, error) {
	return b.app.GetSessionVideoInfo(sessionID)
}

// StartMCPServer starts the MCP server with the given app
func StartMCPServer(app *App) {
	bridge := NewMCPBridge(app)
	mcpServer := mcp.NewMCPServer(bridge)
	if err := mcpServer.Start(); err != nil {
		LogError("mcp").Err(err).Msg("Failed to start MCP server")
	}
}
