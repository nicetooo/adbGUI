package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

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

	// Use ID as filename for uniqueness
	safeName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(workflow.ID, "_")
	if safeName == "" {
		safeName = fmt.Sprintf("wf_%d", time.Now().Unix())
	}

	filePath := filepath.Join(workflowsPath, safeName+".json")

	data, err := json.MarshalIndent(workflow, "", "  ")
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

		var workflow Workflow
		if err := json.Unmarshal(data, &workflow); err != nil {
			continue
		}

		workflows = append(workflows, workflow)
	}

	return workflows, nil
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
	// Re-use logic from automation.go
	a.StopTask(device.ID)
}

// RunWorkflow executes a workflow on the specified device
func (a *App) RunWorkflow(device Device, workflow Workflow) error {
	deviceId := device.ID

	// Check if already running
	touchPlaybackMu.Lock()
	if _, exists := touchPlaybackCancel[deviceId]; exists {
		touchPlaybackMu.Unlock()
		return fmt.Errorf("workflow execution already in progress")
	}

	ctx, cancel := context.WithCancel(context.Background())
	touchPlaybackCancel[deviceId] = cancel
	touchPlaybackMu.Unlock()

	go func() {
		defer func() {
			touchPlaybackMu.Lock()
			delete(touchPlaybackCancel, deviceId)
			touchPlaybackMu.Unlock()

			wailsRuntime.EventsEmit(a.ctx, "workflow-completed", map[string]interface{}{
				"deviceId":     deviceId,
				"workflowName": workflow.Name,
			})
		}()

		wailsRuntime.EventsEmit(a.ctx, "workflow-started", map[string]interface{}{
			"deviceId":     deviceId,
			"workflowName": workflow.Name,
			"steps":        len(workflow.Steps),
		})

		for i, step := range workflow.Steps {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Check pause (reusing automation.go logic)
			a.checkPause(deviceId)

			wailsRuntime.EventsEmit(a.ctx, "workflow-step-running", map[string]interface{}{
				"deviceId":  deviceId,
				"stepIndex": i, // Frontend uses index for highlighting
				"stepId":    step.ID,
				"stepName":  step.Name,
				"stepType":  step.Type,
			})

			loopCount := step.Loop
			if loopCount < 1 {
				loopCount = 1
			}

			for l := 0; l < loopCount; l++ {
				if err := a.runWorkflowStep(ctx, deviceId, step, l+1, loopCount, 0); err != nil {
					// Handle error
					if step.OnError == "continue" {
						fmt.Printf("[Workflow] Step failed but continuing: %v\n", err)
						continue
					}

					wailsRuntime.EventsEmit(a.ctx, "workflow-error", map[string]interface{}{
						"deviceId": deviceId,
						"error":    err.Error(),
					})
					return // Stop execution
				}
				// Check cancel between loops
				select {
				case <-ctx.Done():
					return
				default:
				}
			}

			// Post Delay
			if step.PostDelay > 0 {
				time.Sleep(time.Duration(step.PostDelay) * time.Millisecond)
			}
		}
	}()

	return nil
}

func (a *App) runWorkflowStep(ctx context.Context, deviceId string, step WorkflowStep, currentLoop, totalLoops, depth int) error {
	// Recursion guard
	if depth > 10 {
		return fmt.Errorf("maximum workflow nesting depth exceeded")
	}

	switch step.Type {
	case "wait":
		duration := 1000
		if val, err := strconv.Atoi(step.Value); err == nil {
			duration = val
		}
		time.Sleep(time.Duration(duration) * time.Millisecond)

	case "adb":
		// Raw ADB Command
		if _, err := a.RunAdbCommand(deviceId, step.Value); err != nil {
			return fmt.Errorf("adb command failed: %w", err)
		}

	case "run_workflow":
		// Load the sub-workflow
		// Value is expected to be the Workflow ID
		workflowID := step.Value
		workflowsPath := a.getWorkflowsPath()

		// Try to find file by ID (sanitized)
		safeName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(workflowID, "_")
		filePath := filepath.Join(workflowsPath, safeName+".json")

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("workflow not found: %s", workflowID)
		}

		var subWorkflow Workflow
		if err := json.Unmarshal(data, &subWorkflow); err != nil {
			return fmt.Errorf("failed to parse sub-workflow: %w", err)
		}

		// Execute steps of the sub-workflow
		for _, subStep := range subWorkflow.Steps {
			// Check context
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			subLoopCount := subStep.Loop
			if subLoopCount < 1 {
				subLoopCount = 1
			}

			for l := 0; l < subLoopCount; l++ {
				// Recursive call with incremented depth
				if err := a.runWorkflowStep(ctx, deviceId, subStep, l+1, subLoopCount, depth+1); err != nil {
					if subStep.OnError == "continue" {
						continue
					}
					return err
				}
			}

			if subStep.PostDelay > 0 {
				time.Sleep(time.Duration(subStep.PostDelay) * time.Millisecond)
			}
		}

	case "script":
		// Run recorded script
		// 1. Load script (inefficient to load every time, but specific enough)
		scriptsPath := a.getScriptsPath()
		// Try exact match or safe name
		scriptName := step.Value
		safeName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(scriptName, "_")
		filePath := filepath.Join(scriptsPath, safeName+".json")

		data, err := os.ReadFile(filePath)
		if err != nil {
			// Try assuming the Value IS the safe name
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

		// Use existing player
		return a.playTouchScriptSync(ctx, deviceId, script, nil)

	case "click_element", "long_click_element", "input_text", "assert_element", "wait_element", "wait_gone":
		return a.handleElementAction(ctx, deviceId, step)

	case "swipe_element":
		// TODO: Implement swipe on element
		// For now, treat as generic swipe if no selector?
		return fmt.Errorf("swipe_element not fully implemented yet")

	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}

	return nil
}

func (a *App) handleElementAction(ctx context.Context, deviceId string, step WorkflowStep) error {
	// Default timeout 10s
	timeout := 10000
	if step.Timeout > 0 {
		timeout = step.Timeout
	}

	startTime := time.Now()

	// Polling loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Dump UI
		// Reuse GetUIHierarchy from automation.go
		hierarchy, err := a.GetUIHierarchy(deviceId)
		if err != nil {
			// Retry if dump fails
			time.Sleep(1 * time.Second)
			if time.Since(startTime) > time.Duration(timeout)*time.Millisecond {
				return fmt.Errorf("UI dump failed: %w", err)
			}
			continue
		}

		// Search content
		var foundNode *UINode

		if step.Selector != nil {
			// "advanced" usually implies XPath or complex logic.
			if step.Selector.Type == "xpath" {
				results := a.SearchElementsXPath(hierarchy.Root, step.Selector.Value)
				if len(results) > 0 {
					// Use index if specified, else 0
					idx := step.Selector.Index
					if idx < len(results) {
						foundNode = results[idx].Node
					}
				}
			} else {
				// Simple FindElement (recursive)
				// Note: FindElement returns bool, we need the node.
				// Let's make a helper that returns the node.
				foundNode = a.findElementNode(hierarchy.Root, step.Selector.Type, step.Selector.Value)
			}
		}

		// Handle "wait_gone" logic
		if step.Type == "wait_gone" {
			if foundNode == nil {
				return nil // It's gone!
			}
			// It exists, keep waiting
			if time.Since(startTime) > time.Duration(timeout)*time.Millisecond {
				return fmt.Errorf("timeout waiting for element to disappear")
			}
			time.Sleep(1 * time.Second)
			continue
		}

		// Handle finding logic
		if foundNode != nil {
			// Element Found! Perform Action.

			if step.Type == "wait_element" || step.Type == "assert_element" {
				return nil // Success
			}

			// Calculate center point
			bounds := foundNode.Bounds // "[0,0][1080,220]"
			x, y, err := parseBoundsCenter(bounds)
			if err != nil {
				return fmt.Errorf("invalid bounds: %s", bounds)
			}

			if step.Type == "click_element" {
				_, err := a.RunAdbCommand(deviceId, fmt.Sprintf("shell input tap %d %d", x, y))
				return err
			}

			if step.Type == "long_click_element" {
				_, err := a.RunAdbCommand(deviceId, fmt.Sprintf("shell input swipe %d %d %d %d 1000", x, y, x, y))
				return err
			}

			if step.Type == "input_text" {
				// Click first to focus?
				a.RunAdbCommand(deviceId, fmt.Sprintf("shell input tap %d %d", x, y))
				time.Sleep(500 * time.Millisecond)

				// Escape string for shell
				text := step.Value
				text = strings.ReplaceAll(text, " ", "%s") // Android input text uses %s for space
				text = strings.ReplaceAll(text, "'", "\\'")

				_, err := a.RunAdbCommand(deviceId, fmt.Sprintf("shell input text '%s'", text))
				return err
			}

			return nil
		}

		// Not found yet
		if time.Since(startTime) > time.Duration(timeout)*time.Millisecond {
			return fmt.Errorf("element not found within timeout")
		}

		time.Sleep(1 * time.Second)
	}
}

// Helper to find node and return pointer
func (a *App) findElementNode(node *UINode, checkType, checkValue string) *UINode {
	if node == nil {
		return nil
	}

	match := false
	switch checkType {
	case "text":
		match = node.Text == checkValue
	case "id":
		match = node.ResourceID == checkValue || strings.HasSuffix(node.ResourceID, ":id/"+checkValue)
	case "class":
		match = node.Class == checkValue
	case "contains":
		match = strings.Contains(node.Text, checkValue) || strings.Contains(node.ContentDesc, checkValue)
	case "description":
		match = node.ContentDesc == checkValue
	}

	if match {
		return node
	}

	for i := range node.Nodes {
		found := a.findElementNode(&node.Nodes[i], checkType, checkValue)
		if found != nil {
			return found
		}
	}

	return nil
}

func parseBoundsCenter(bounds string) (int, int, error) {
	// Format: [x1,y1][x2,y2]
	re := regexp.MustCompile(`\[(\d+),(\d+)\]\[(\d+),(\d+)\]`)
	matches := re.FindStringSubmatch(bounds)
	if len(matches) != 5 {
		return 0, 0, fmt.Errorf("invalid bounds format")
	}

	x1, _ := strconv.Atoi(matches[1])
	y1, _ := strconv.Atoi(matches[2])
	x2, _ := strconv.Atoi(matches[3])
	y2, _ := strconv.Atoi(matches[4])

	centerX := x1 + (x2-x1)/2
	centerY := y1 + (y2-y1)/2

	return centerX, centerY, nil
}
