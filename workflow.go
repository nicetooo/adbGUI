package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// StepResult represents the result of executing a step
type StepResult struct {
	Success        bool  // Whether the step executed successfully
	IsBranchResult bool  // For branch nodes: true if this is a condition result (not execution error)
	Error          error // Error if execution failed
}

// getWorkflowsPath returns the path to the workflows directory
func (a *App) getWorkflowsPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	workflowsPath := filepath.Join(configDir, "Gaze", "workflows")
	_ = os.MkdirAll(workflowsPath, 0755)
	return workflowsPath
}

// SaveWorkflow saves a workflow to file
func (a *App) SaveWorkflow(workflow Workflow) error {
	workflowsPath := a.getWorkflowsPath()

	safeName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(workflow.ID, "_")
	if safeName == "" {
		safeName = fmt.Sprintf("wf_%d", time.Now().Unix())
	}

	filePath := filepath.Join(workflowsPath, safeName+".json")

	data, err := SerializeWorkflow(&workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	return nil
}

// LoadWorkflows loads all saved workflows
func (a *App) LoadWorkflows() ([]Workflow, error) {
	workflowsPath := a.getWorkflowsPath()

	entries, err := os.ReadDir(workflowsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Workflow{}, nil
		}
		return nil, fmt.Errorf("failed to read workflows directory: %w", err)
	}

	workflows := make([]Workflow, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(workflowsPath, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		workflow, err := ParseWorkflow(data)
		if err != nil {
			LogWarn("workflow").Err(err).Str("file", entry.Name()).Msg("Failed to parse workflow")
			continue
		}

		workflows = append(workflows, *workflow)
	}

	return workflows, nil
}

// LoadWorkflowByID loads a single workflow by ID
func (a *App) LoadWorkflowByID(id string) (*Workflow, error) {
	workflowsPath := a.getWorkflowsPath()
	safeName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(id, "_")
	filePath := filepath.Join(workflowsPath, safeName+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("workflow not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read workflow: %w", err)
	}

	return ParseWorkflow(data)
}

// GetWorkflow returns a workflow by ID
func (a *App) GetWorkflow(workflowID string) (*Workflow, error) {
	return a.LoadWorkflowByID(workflowID)
}

// DeleteWorkflow deletes a saved workflow
func (a *App) DeleteWorkflow(id string) error {
	workflowsPath := a.getWorkflowsPath()

	safeName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(id, "_")
	filePath := filepath.Join(workflowsPath, safeName+".json")

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("workflow not found")
		}
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	return nil
}

// StopWorkflow stops the currently running workflow on the device
func (a *App) StopWorkflow(device Device) {
	a.StopTask(device.ID)
}

// ExecuteSingleWorkflowStep executes a single workflow step on the device
func (a *App) ExecuteSingleWorkflowStep(deviceId string, step WorkflowStep) error {
	if step.Type == "start" {
		return nil
	}

	ctx := a.ctx
	vars := make(map[string]string)

	if step.Common.PreWait > 0 {
		time.Sleep(time.Duration(step.Common.PreWait) * time.Millisecond)
	}

	result := a.executeStep(ctx, deviceId, &step, vars, "")

	if step.Common.PostDelay > 0 {
		time.Sleep(time.Duration(step.Common.PostDelay) * time.Millisecond)
	}

	return result.Error
}

// RunWorkflow executes a workflow on the specified device
func (a *App) RunWorkflow(device Device, workflow Workflow) error {
	deviceId := device.ID

	touchPlaybackMu.Lock()
	if _, exists := touchPlaybackCancel[deviceId]; exists {
		touchPlaybackMu.Unlock()
		return fmt.Errorf("workflow execution already in progress")
	}

	ctx, cancel := context.WithCancel(a.ctx)
	touchPlaybackCancel[deviceId] = cancel
	touchPlaybackMu.Unlock()

	sessionId := a.EnsureActiveSession(deviceId)

	a.EmitSessionEventFull(SessionEvent{
		DeviceID: deviceId,
		Type:     "workflow_start",
		Category: "workflow",
		Level:    "info",
		Title:    fmt.Sprintf("Workflow started: %s", workflow.Name),
		Detail: map[string]interface{}{
			"workflowId": workflow.ID,
			"totalSteps": len(workflow.Steps),
			"sessionId":  sessionId,
		},
	})

	wailsRuntime.EventsEmit(a.ctx, "workflow-started", map[string]interface{}{
		"workflowId":   workflow.ID,
		"workflowName": workflow.Name,
		"deviceId":     deviceId,
	})

	go func() {
		startTime := time.Now()
		sessionStatus := "completed"

		err := a.runWorkflowInternal(ctx, deviceId, workflow, nil)

		duration := time.Since(startTime).Milliseconds()

		touchPlaybackMu.Lock()
		delete(touchPlaybackCancel, deviceId)
		touchPlaybackMu.Unlock()

		if err != nil {
			sessionStatus = "error"
			if err == context.Canceled {
				sessionStatus = "cancelled"
			}

			a.EmitSessionEventFull(SessionEvent{
				DeviceID: deviceId,
				Type:     "workflow_error",
				Category: "workflow",
				Level:    "error",
				Title:    fmt.Sprintf("Workflow error: %s", workflow.Name),
				Duration: duration,
				Detail: map[string]interface{}{
					"workflowId": workflow.ID,
					"error":      err.Error(),
					"status":     sessionStatus,
				},
			})

			wailsRuntime.EventsEmit(a.ctx, "workflow-error", map[string]interface{}{
				"workflowId":   workflow.ID,
				"workflowName": workflow.Name,
				"deviceId":     deviceId,
				"error":        err.Error(),
			})
		} else {
			a.EmitSessionEventFull(SessionEvent{
				DeviceID: deviceId,
				Type:     "workflow_completed",
				Category: "workflow",
				Level:    "info",
				Title:    fmt.Sprintf("Workflow completed: %s", workflow.Name),
				Duration: duration,
				Detail: map[string]interface{}{
					"workflowId": workflow.ID,
					"status":     sessionStatus,
				},
			})

			wailsRuntime.EventsEmit(a.ctx, "workflow-completed", map[string]interface{}{
				"workflowId":   workflow.ID,
				"workflowName": workflow.Name,
				"deviceId":     deviceId,
				"duration":     duration,
			})
		}
	}()

	return nil
}

// runWorkflowInternal is the core execution loop
func (a *App) runWorkflowInternal(ctx context.Context, deviceId string, workflow Workflow, parentVars map[string]string) error {
	stepMap := make(map[string]*WorkflowStep)
	for i := range workflow.Steps {
		stepMap[workflow.Steps[i].ID] = &workflow.Steps[i]
	}

	vars := make(map[string]string)
	for k, v := range workflow.Variables {
		vars[k] = v
	}
	for k, v := range parentVars {
		vars[k] = v
	}

	var startStep *WorkflowStep
	for i := range workflow.Steps {
		if workflow.Steps[i].Type == "start" {
			startStep = &workflow.Steps[i]
			break
		}
	}

	if startStep == nil {
		return fmt.Errorf("workflow has no start node")
	}

	currentStepId := startStep.Connections.SuccessStepId
	if currentStepId == "" {
		return nil
	}

	maxSteps := 2000
	stepCount := 0

	for currentStepId != "" && stepCount < maxSteps {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		step, exists := stepMap[currentStepId]
		if !exists {
			return fmt.Errorf("step not found: %s", currentStepId)
		}

		stepCount++

		wailsRuntime.EventsEmit(a.ctx, "workflow-step-running", map[string]interface{}{
			"workflowId": workflow.ID,
			"stepId":     step.ID,
			"stepType":   step.Type,
			"stepName":   step.Name,
		})

		a.EmitSessionEventFull(SessionEvent{
			DeviceID: deviceId,
			Type:     "workflow_step_start",
			Category: "workflow",
			Level:    "info",
			Title:    fmt.Sprintf("Step: %s (%s)", step.Name, step.Type),
			Detail: map[string]interface{}{
				"workflowId": workflow.ID,
				"stepId":     step.ID,
				"stepType":   step.Type,
			},
		})

		loopCount := step.Common.Loop
		if loopCount <= 0 {
			loopCount = 1
		}

		var result StepResult
		for i := 0; i < loopCount; i++ {
			if step.Common.PreWait > 0 {
				time.Sleep(time.Duration(step.Common.PreWait) * time.Millisecond)
			}

			result = a.executeStep(ctx, deviceId, step, vars, workflow.ID)

			if step.Common.PostDelay > 0 {
				time.Sleep(time.Duration(step.Common.PostDelay) * time.Millisecond)
			}

			if !result.Success && step.ShouldStopOnError() {
				break
			}
		}

		nextStepId := a.determineNextStep(step, result)

		logLevel := "info"
		if !result.Success {
			logLevel = "warn"
		}
		a.EmitSessionEventFull(SessionEvent{
			DeviceID: deviceId,
			Type:     "workflow_step_end",
			Category: "workflow",
			Level:    logLevel,
			Title:    fmt.Sprintf("Step completed: %s", step.Name),
			Detail: map[string]interface{}{
				"workflowId": workflow.ID,
				"stepId":     step.ID,
				"success":    result.Success,
				"nextStepId": nextStepId,
			},
		})

		if result.Error != nil && nextStepId == "" && step.ShouldStopOnError() {
			return fmt.Errorf("step %s failed: %w", step.ID, result.Error)
		}

		currentStepId = nextStepId
	}

	if stepCount >= maxSteps {
		return fmt.Errorf("maximum step count exceeded (%d)", maxSteps)
	}

	return nil
}

// determineNextStep determines the next step based on execution result
func (a *App) determineNextStep(step *WorkflowStep, result StepResult) string {
	if step.Type == "branch" && result.IsBranchResult {
		if result.Success {
			return step.Connections.TrueStepId
		}
		return step.Connections.FalseStepId
	}

	if result.Success {
		return step.Connections.SuccessStepId
	}

	if step.Connections.ErrorStepId != "" {
		return step.Connections.ErrorStepId
	}

	if step.Common.OnError == "continue" {
		return step.Connections.SuccessStepId
	}

	return ""
}

// executeStep executes a single step and returns the result
func (a *App) executeStep(ctx context.Context, deviceId string, step *WorkflowStep, vars map[string]string, workflowId string) StepResult {
	switch step.Type {
	case "start":
		return StepResult{Success: true}

	case "tap":
		if step.Tap == nil {
			return StepResult{Success: false, Error: fmt.Errorf("tap params missing")}
		}
		_, err := a.RunAdbCommand(deviceId, fmt.Sprintf("shell input tap %d %d", step.Tap.X, step.Tap.Y))
		return StepResult{Success: err == nil, Error: err}

	case "swipe":
		if step.Swipe == nil {
			return StepResult{Success: false, Error: fmt.Errorf("swipe params missing")}
		}
		x1, y1 := step.Swipe.X, step.Swipe.Y
		x2, y2 := step.Swipe.X2, step.Swipe.Y2
		duration := step.Swipe.Duration
		if duration <= 0 {
			duration = 300
		}

		if step.Swipe.Direction != "" && step.Swipe.Distance > 0 {
			distance := step.Swipe.Distance
			switch step.Swipe.Direction {
			case "up":
				y2 = y1 - distance
				x2 = x1
			case "down":
				y2 = y1 + distance
				x2 = x1
			case "left":
				x2 = x1 - distance
				y2 = y1
			case "right":
				x2 = x1 + distance
				y2 = y1
			}
		}

		_, err := a.RunAdbCommand(deviceId, fmt.Sprintf("shell input swipe %d %d %d %d %d", x1, y1, x2, y2, duration))
		return StepResult{Success: err == nil, Error: err}

	case "click_element", "long_click_element", "input_text", "swipe_element", "wait_element", "wait_gone", "assert_element":
		return a.executeElementStep(ctx, deviceId, step, vars)

	case "launch_app", "stop_app", "clear_app", "open_settings":
		return a.executeAppStep(deviceId, step)

	case "branch":
		return a.executeBranchStep(deviceId, step, vars)

	case "wait":
		if step.Wait == nil {
			return StepResult{Success: false, Error: fmt.Errorf("wait params missing")}
		}
		time.Sleep(time.Duration(step.Wait.DurationMs) * time.Millisecond)
		return StepResult{Success: true}

	case "script":
		if step.Script == nil {
			return StepResult{Success: false, Error: fmt.Errorf("script params missing")}
		}
		err := a.executeScriptStep(ctx, deviceId, step.Script.ScriptName)
		return StepResult{Success: err == nil, Error: err}

	case "set_variable":
		if step.Variable == nil {
			return StepResult{Success: false, Error: fmt.Errorf("variable params missing")}
		}
		processedValue := a.processWorkflowVariables(step.Variable.Value, vars)
		vars[step.Variable.Name] = processedValue
		return StepResult{Success: true}

	case "adb":
		if step.ADB == nil {
			return StepResult{Success: false, Error: fmt.Errorf("adb params missing")}
		}
		processedCmd := a.processWorkflowVariables(step.ADB.Command, vars)
		_, err := a.RunAdbCommand(deviceId, processedCmd)
		return StepResult{Success: err == nil, Error: err}

	case "run_workflow":
		if step.Workflow == nil {
			return StepResult{Success: false, Error: fmt.Errorf("workflow params missing")}
		}
		err := a.executeSubWorkflow(ctx, deviceId, step.Workflow.WorkflowId, vars, 0)
		return StepResult{Success: err == nil, Error: err}

	case "key_back":
		_, err := a.RunAdbCommand(deviceId, "shell input keyevent 4")
		return StepResult{Success: err == nil, Error: err}

	case "key_home":
		_, err := a.RunAdbCommand(deviceId, "shell input keyevent 3")
		return StepResult{Success: err == nil, Error: err}

	case "key_recent":
		_, err := a.RunAdbCommand(deviceId, "shell input keyevent 187")
		return StepResult{Success: err == nil, Error: err}

	case "key_power":
		_, err := a.RunAdbCommand(deviceId, "shell input keyevent 26")
		return StepResult{Success: err == nil, Error: err}

	case "key_volume_up":
		_, err := a.RunAdbCommand(deviceId, "shell input keyevent 24")
		return StepResult{Success: err == nil, Error: err}

	case "key_volume_down":
		_, err := a.RunAdbCommand(deviceId, "shell input keyevent 25")
		return StepResult{Success: err == nil, Error: err}

	case "screen_on":
		_, err := a.RunAdbCommand(deviceId, "shell input keyevent 224")
		return StepResult{Success: err == nil, Error: err}

	case "screen_off":
		_, err := a.RunAdbCommand(deviceId, "shell input keyevent 223")
		return StepResult{Success: err == nil, Error: err}

	default:
		return StepResult{Success: false, Error: fmt.Errorf("unknown step type: %s", step.Type)}
	}
}

// executeElementStep handles element-based actions
func (a *App) executeElementStep(ctx context.Context, deviceId string, step *WorkflowStep, vars map[string]string) StepResult {
	if step.Element == nil {
		return StepResult{Success: false, Error: fmt.Errorf("element params missing")}
	}

	selector := &ElementSelector{
		Type:  step.Element.Selector.Type,
		Value: a.processWorkflowVariables(step.Element.Selector.Value, vars),
		Index: step.Element.Selector.Index,
	}

	timeout := step.Common.Timeout
	if timeout <= 0 {
		timeout = 5000
	}

	config := &ElementActionConfig{
		Timeout: timeout,
	}

	var err error
	switch step.Element.Action {
	case "click":
		err = a.ClickElement(ctx, deviceId, selector, config)

	case "long_click":
		err = a.LongClickElement(ctx, deviceId, selector, 1000, config)

	case "input":
		inputText := a.processWorkflowVariables(step.Element.InputText, vars)
		err = a.InputTextToElement(ctx, deviceId, selector, inputText, false, config)

	case "swipe":
		distance := step.Element.SwipeDistance
		if distance <= 0 {
			distance = 300
		}
		duration := step.Element.SwipeDuration
		if duration <= 0 {
			duration = 300
		}
		err = a.SwipeOnElement(ctx, deviceId, selector, step.Element.SwipeDir, distance, duration, config)

	case "wait":
		err = a.WaitForElement(ctx, deviceId, selector, timeout)

	case "wait_gone":
		err = a.WaitElementGone(ctx, deviceId, selector, timeout)

	case "assert":
		err = a.WaitForElement(ctx, deviceId, selector, timeout)

	default:
		err = fmt.Errorf("unknown element action: %s", step.Element.Action)
	}

	return StepResult{Success: err == nil, Error: err}
}

// executeAppStep handles app operations
func (a *App) executeAppStep(deviceId string, step *WorkflowStep) StepResult {
	if step.App == nil {
		return StepResult{Success: false, Error: fmt.Errorf("app params missing")}
	}

	var err error
	switch step.App.Action {
	case "launch":
		_, err = a.StartApp(deviceId, step.App.PackageName)
	case "stop":
		_, err = a.ForceStopApp(deviceId, step.App.PackageName)
	case "clear":
		_, err = a.ClearAppData(deviceId, step.App.PackageName)
	case "settings":
		_, err = a.OpenSettings(deviceId, "android.settings.APPLICATION_DETAILS_SETTINGS", "package:"+step.App.PackageName)
	default:
		err = fmt.Errorf("unknown app action: %s", step.App.Action)
	}

	return StepResult{Success: err == nil, Error: err}
}

// executeBranchStep handles branch condition evaluation
func (a *App) executeBranchStep(deviceId string, step *WorkflowStep, vars map[string]string) StepResult {
	if step.Branch == nil {
		return StepResult{Success: false, Error: fmt.Errorf("branch params missing")}
	}

	condition := step.Branch.Condition
	if condition == "" {
		condition = "exists"
	}

	if condition == "variable_equals" {
		varName := step.Branch.VariableName
		expected := step.Branch.ExpectedValue
		actual := vars[varName]
		result := actual == expected
		return StepResult{Success: result, IsBranchResult: true}
	}

	if step.Branch.Selector == nil {
		return StepResult{Success: false, Error: fmt.Errorf("branch selector required for condition: %s", condition)}
	}

	hierarchy, err := a.GetUIHierarchy(deviceId)
	if err != nil {
		return StepResult{Success: false, Error: err}
	}

	selectorValue := a.processWorkflowVariables(step.Branch.Selector.Value, vars)

	var conditionResult bool
	switch condition {
	case "exists":
		node := a.findElementInHierarchy(hierarchy, step.Branch.Selector.Type, selectorValue)
		conditionResult = node != nil

	case "not_exists":
		node := a.findElementInHierarchy(hierarchy, step.Branch.Selector.Type, selectorValue)
		conditionResult = node == nil

	case "text_equals":
		node := a.findElementInHierarchy(hierarchy, step.Branch.Selector.Type, selectorValue)
		if node != nil {
			expected := step.Branch.ExpectedValue
			conditionResult = node.Text == expected
		}

	case "text_contains":
		node := a.findElementInHierarchy(hierarchy, step.Branch.Selector.Type, selectorValue)
		if node != nil {
			expected := step.Branch.ExpectedValue
			conditionResult = strings.Contains(node.Text, expected)
		}

	default:
		return StepResult{Success: false, Error: fmt.Errorf("unknown branch condition: %s", condition)}
	}

	return StepResult{Success: conditionResult, IsBranchResult: true}
}

// executeSubWorkflow runs a sub-workflow
func (a *App) executeSubWorkflow(ctx context.Context, deviceId string, workflowId string, vars map[string]string, depth int) error {
	if depth > 10 {
		return fmt.Errorf("maximum workflow nesting depth exceeded")
	}

	workflow, err := a.LoadWorkflowByID(workflowId)
	if err != nil {
		return fmt.Errorf("failed to load sub-workflow: %w", err)
	}

	return a.runWorkflowInternal(ctx, deviceId, *workflow, vars)
}

// executeScriptStep runs a recorded touch script
func (a *App) executeScriptStep(ctx context.Context, deviceId string, scriptName string) error {
	scriptsPath := a.getScriptsPath()
	safeName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(scriptName, "_")
	filePath := filepath.Join(scriptsPath, safeName+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		filePath = filepath.Join(scriptsPath, scriptName+".json")
		data, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("script not found: %s", scriptName)
		}
	}

	var script TouchScript
	if err := json.Unmarshal(data, &script); err != nil {
		return fmt.Errorf("failed to parse script: %w", err)
	}

	return a.playTouchScriptSync(ctx, deviceId, script, nil)
}

// processWorkflowVariables replaces placeholders like {{var}} with actual values
func (a *App) processWorkflowVariables(text string, vars map[string]string) string {
	if text == "" {
		return ""
	}
	for k, v := range vars {
		placeholder := "{{" + k + "}}"
		text = strings.ReplaceAll(text, placeholder, v)
	}
	return text
}

// findElementInHierarchy finds an element in the UI hierarchy
func (a *App) findElementInHierarchy(hierarchy *UIHierarchyResult, selectorType, selectorValue string) *UINode {
	if hierarchy == nil || hierarchy.Root == nil {
		return nil
	}

	if selectorType == "xpath" {
		results := a.SearchElementsXPath(hierarchy.Root, selectorValue)
		if len(results) > 0 {
			return results[0].Node
		}
		return nil
	}

	selector := &ElementSelector{Type: selectorType, Value: selectorValue, Index: 0}
	return a.FindElementBySelector(hierarchy.Root, selector)
}
