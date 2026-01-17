package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ========================================
// Workflow AI Executor - AI 智能执行引擎
// ========================================

// WorkflowExecutionMode defines how the workflow handles anomalies
type WorkflowExecutionMode string

const (
	ModeStrict     WorkflowExecutionMode = "strict"     // Stop on any error
	ModeAssisted   WorkflowExecutionMode = "assisted"   // AI analyzes + user confirms
	ModeAutonomous WorkflowExecutionMode = "autonomous" // AI handles known anomalies automatically
)

// WorkflowExecutionConfig configures how workflows are executed
type WorkflowExecutionConfig struct {
	Mode                    WorkflowExecutionMode `json:"mode"`
	AutoHandlePermissions   bool                  `json:"autoHandlePermissions"`
	AutoHandleSystemDialogs bool                  `json:"autoHandleSystemDialogs"`
	AutoUpdateWorkflow      bool                  `json:"autoUpdateWorkflow"`
	MaxRetries              int                   `json:"maxRetries"`
	AnomalyTimeout          int                   `json:"anomalyTimeout"` // seconds
}

// DefaultExecutionConfig returns default execution config
func DefaultExecutionConfig() WorkflowExecutionConfig {
	return WorkflowExecutionConfig{
		Mode:                    ModeAssisted,
		AutoHandlePermissions:   true,
		AutoHandleSystemDialogs: true,
		AutoUpdateWorkflow:      false,
		MaxRetries:              3,
		AnomalyTimeout:          30,
	}
}

// WorkflowAIExecutor handles AI-assisted workflow execution
type WorkflowAIExecutor struct {
	app             *App
	config          WorkflowExecutionConfig
	anomalyHandler  *WorkflowAnomalyHandler
	currentWorkflow *Workflow
	currentStep     int
	isPaused        bool
	mu              sync.RWMutex

	// Channels for async operations
	anomalyChan   chan *AnomalyContext
	responseChan  chan *AnomalyResponse
	stopChan      chan struct{}
}

// AnomalyContext represents the context when an anomaly is detected
type AnomalyContext struct {
	WorkflowID    string         `json:"workflowId"`
	StepID        string         `json:"stepId"`
	StepIndex     int            `json:"stepIndex"`
	ExpectedState *UIState       `json:"expectedState,omitempty"`
	ActualState   *UIState       `json:"actualState,omitempty"`
	Screenshot    string         `json:"screenshot,omitempty"` // Base64
	UIHierarchy   string         `json:"uiHierarchy,omitempty"`
	ErrorMessage  string         `json:"errorMessage"`
	Timestamp     int64          `json:"timestamp"`
}

// UIState represents the expected or actual UI state
type UIState struct {
	Activity    string            `json:"activity,omitempty"`
	Elements    []ElementState    `json:"elements,omitempty"`
	ScreenText  []string          `json:"screenText,omitempty"`
}

// ElementState represents the state of a UI element
type ElementState struct {
	Selector    *ElementSelector `json:"selector,omitempty"`
	Visible     bool             `json:"visible"`
	Enabled     bool             `json:"enabled"`
	Text        string           `json:"text,omitempty"`
}

// AnomalyResponse represents the user/AI response to an anomaly
type AnomalyResponse struct {
	Action       string         `json:"action"`       // "execute", "skip", "pause", "stop", "retry"
	Steps        []WorkflowStep `json:"steps,omitempty"` // Steps to execute (for "execute" action)
	UpdateWorkflow bool         `json:"updateWorkflow"`  // Whether to save this handling to workflow
	Remember     bool           `json:"remember"`        // Remember this choice for future
}

// StepExecutionResult represents the result of executing a step
type StepExecutionResult struct {
	StepID      string           `json:"stepId"`
	Success     bool             `json:"success"`
	Error       string           `json:"error,omitempty"`
	Duration    int64            `json:"duration"` // ms
	Anomaly     *AnomalyAnalysis `json:"anomaly,omitempty"`
	Handled     bool             `json:"handled"` // Whether anomaly was handled
}

// NewWorkflowAIExecutor creates a new AI executor
func NewWorkflowAIExecutor(app *App, config WorkflowExecutionConfig) *WorkflowAIExecutor {
	handler := NewWorkflowAnomalyHandler(app)
	return &WorkflowAIExecutor{
		app:            app,
		config:         config,
		anomalyHandler: handler,
		anomalyChan:    make(chan *AnomalyContext, 1),
		responseChan:   make(chan *AnomalyResponse, 1),
		stopChan:       make(chan struct{}),
	}
}

// ExecuteWorkflow executes a workflow with AI assistance
func (e *WorkflowAIExecutor) ExecuteWorkflow(ctx context.Context, deviceID string, workflow *Workflow) ([]StepExecutionResult, error) {
	e.mu.Lock()
	e.currentWorkflow = workflow
	e.currentStep = 0
	e.isPaused = false
	e.mu.Unlock()

	results := make([]StepExecutionResult, 0, len(workflow.Steps))

	for i, step := range workflow.Steps {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case <-e.stopChan:
			return results, fmt.Errorf("execution stopped by user")
		default:
		}

		e.mu.Lock()
		e.currentStep = i
		e.mu.Unlock()

		result := e.executeStep(ctx, deviceID, &step, i)
		results = append(results, result)

		if !result.Success {
			if e.config.Mode == ModeStrict {
				return results, fmt.Errorf("step %s failed: %s", step.ID, result.Error)
			}

			// Try to handle the anomaly
			handled, err := e.handleStepFailure(ctx, deviceID, &step, i, result)
			if err != nil {
				if e.config.Mode == ModeAssisted {
					// Wait for user response
					continue
				}
				return results, err
			}

			result.Handled = handled
		}
	}

	return results, nil
}

// executeStep executes a single workflow step
func (e *WorkflowAIExecutor) executeStep(ctx context.Context, deviceID string, step *WorkflowStep, index int) StepExecutionResult {
	start := time.Now()
	result := StepExecutionResult{
		StepID: step.ID,
	}

	// Execute the step using the existing automation engine
	// This is a simplified version - real implementation would use app.ExecuteWorkflowStep
	err := e.app.executeAutomationStep(deviceID, step)

	result.Duration = time.Since(start).Milliseconds()

	if err != nil {
		result.Success = false
		result.Error = err.Error()

		// Analyze the anomaly
		if e.config.Mode != ModeStrict {
			anomaly, _ := e.anomalyHandler.AnalyzeAnomaly(ctx, deviceID, step, err.Error())
			result.Anomaly = anomaly
		}
	} else {
		result.Success = true
	}

	return result
}

// handleStepFailure handles a failed step
func (e *WorkflowAIExecutor) handleStepFailure(ctx context.Context, deviceID string, step *WorkflowStep, index int, result StepExecutionResult) (bool, error) {
	if result.Anomaly == nil {
		return false, nil
	}

	anomaly := result.Anomaly

	// Check if we can auto-handle this anomaly
	if e.canAutoHandle(anomaly) {
		return e.autoHandleAnomaly(ctx, deviceID, anomaly)
	}

	// In autonomous mode, try AI handling
	if e.config.Mode == ModeAutonomous {
		suggestions := anomaly.SuggestedActions
		if len(suggestions) > 0 && suggestions[0].AutoExecute {
			return e.executeAnomalyAction(ctx, deviceID, &suggestions[0])
		}
	}

	// In assisted mode, notify frontend and wait for user
	if e.config.Mode == ModeAssisted {
		// Send anomaly to frontend
		e.notifyAnomalyDetected(anomaly, step, index)

		// Wait for response with timeout
		select {
		case response := <-e.responseChan:
			return e.processAnomalyResponse(ctx, deviceID, anomaly, response)
		case <-time.After(time.Duration(e.config.AnomalyTimeout) * time.Second):
			return false, fmt.Errorf("anomaly response timeout")
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}

	return false, nil
}

// canAutoHandle checks if an anomaly can be automatically handled
func (e *WorkflowAIExecutor) canAutoHandle(anomaly *AnomalyAnalysis) bool {
	switch anomaly.AnomalyType {
	case "permission_dialog":
		return e.config.AutoHandlePermissions
	case "system_dialog":
		return e.config.AutoHandleSystemDialogs
	default:
		return false
	}
}

// autoHandleAnomaly automatically handles known anomaly types
func (e *WorkflowAIExecutor) autoHandleAnomaly(ctx context.Context, deviceID string, anomaly *AnomalyAnalysis) (bool, error) {
	for _, action := range anomaly.SuggestedActions {
		if action.AutoExecute {
			return e.executeAnomalyAction(ctx, deviceID, &action)
		}
	}
	return false, nil
}

// executeAnomalyAction executes a suggested action
func (e *WorkflowAIExecutor) executeAnomalyAction(ctx context.Context, deviceID string, action *SuggestedAction) (bool, error) {
	for _, step := range action.Steps {
		err := e.app.executeAutomationStep(deviceID, &step)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// processAnomalyResponse processes user response to an anomaly
func (e *WorkflowAIExecutor) processAnomalyResponse(ctx context.Context, deviceID string, anomaly *AnomalyAnalysis, response *AnomalyResponse) (bool, error) {
	switch response.Action {
	case "execute":
		for _, step := range response.Steps {
			if err := e.app.executeAutomationStep(deviceID, &step); err != nil {
				return false, err
			}
		}
		return true, nil

	case "skip":
		return true, nil

	case "retry":
		return false, nil

	case "pause":
		e.mu.Lock()
		e.isPaused = true
		e.mu.Unlock()
		return false, fmt.Errorf("execution paused")

	case "stop":
		return false, fmt.Errorf("execution stopped by user")

	default:
		return false, fmt.Errorf("unknown action: %s", response.Action)
	}
}

// notifyAnomalyDetected sends anomaly notification to frontend
func (e *WorkflowAIExecutor) notifyAnomalyDetected(anomaly *AnomalyAnalysis, step *WorkflowStep, index int) {
	// Emit event to frontend via Wails runtime
	if e.app.ctx != nil {
		data := map[string]interface{}{
			"anomaly":   anomaly,
			"step":      step,
			"stepIndex": index,
		}
		// This would be emitted to the frontend
		_ = data
	}
}

// SendAnomalyResponse sends a response to a pending anomaly
func (e *WorkflowAIExecutor) SendAnomalyResponse(response *AnomalyResponse) {
	select {
	case e.responseChan <- response:
	default:
		// Response channel full or no listener
	}
}

// Stop stops the current execution
func (e *WorkflowAIExecutor) Stop() {
	close(e.stopChan)
	e.stopChan = make(chan struct{})
}

// Pause pauses the current execution
func (e *WorkflowAIExecutor) Pause() {
	e.mu.Lock()
	e.isPaused = true
	e.mu.Unlock()
}

// Resume resumes the current execution
func (e *WorkflowAIExecutor) Resume() {
	e.mu.Lock()
	e.isPaused = false
	e.mu.Unlock()
}

// GetCurrentStep returns the current step index
func (e *WorkflowAIExecutor) GetCurrentStep() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentStep
}

// IsPaused returns whether execution is paused
func (e *WorkflowAIExecutor) IsPaused() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isPaused
}

// ========================================
// App Methods for AI Execution
// ========================================

// GetWorkflow loads a workflow by ID
func (a *App) GetWorkflow(workflowID string) (*Workflow, error) {
	workflows, err := a.LoadWorkflows()
	if err != nil {
		return nil, fmt.Errorf("failed to load workflows: %w", err)
	}

	for _, w := range workflows {
		if w.ID == workflowID {
			return &w, nil
		}
	}

	return nil, fmt.Errorf("workflow not found: %s", workflowID)
}

// ExecuteWorkflowWithAI executes a workflow with AI assistance
func (a *App) ExecuteWorkflowWithAI(deviceID string, workflowID string, config *WorkflowExecutionConfig) ([]StepExecutionResult, error) {
	workflow, err := a.GetWorkflow(workflowID)
	if err != nil {
		return nil, err
	}

	if config == nil {
		defaultConfig := DefaultExecutionConfig()
		config = &defaultConfig
	}

	executor := NewWorkflowAIExecutor(a, *config)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	return executor.ExecuteWorkflow(ctx, deviceID, workflow)
}

// executeAutomationStep is a helper to execute a single step
func (a *App) executeAutomationStep(deviceID string, step *WorkflowStep) error {
	ctx := context.Background()

	// Use the existing automation engine
	switch step.Type {
	case "click", "tap", "click_element":
		if step.Selector != nil {
			cfg := DefaultElementActionConfig()
			return a.ClickElement(ctx, deviceID, step.Selector, &cfg)
		}
		// Fall back to coordinates if available
		if step.Value != "" {
			var x, y int
			_, err := fmt.Sscanf(step.Value, "%d,%d", &x, &y)
			if err == nil {
				return a.performTap(deviceID, x, y)
			}
		}
		return fmt.Errorf("no selector or coordinates for click step")

	case "long_press":
		if step.Selector != nil {
			cfg := DefaultElementActionConfig()
			return a.LongClickElement(ctx, deviceID, step.Selector, 1000, &cfg)
		}
		return fmt.Errorf("no selector for long_press step")

	case "swipe":
		// Parse swipe parameters from Value or use defaults
		return a.performSwipe(deviceID, 500, 1000, 500, 200, 300)

	case "input", "type":
		if step.Selector != nil {
			cfg := DefaultElementActionConfig()
			if err := a.ClickElement(ctx, deviceID, step.Selector, &cfg); err != nil {
				return err
			}
		}
		return a.performInputText(deviceID, step.Value)

	case "wait":
		if step.Timeout > 0 {
			time.Sleep(time.Duration(step.Timeout) * time.Millisecond)
		} else {
			time.Sleep(time.Second)
		}
		return nil

	case "back":
		return a.performPressBack(deviceID)

	case "home":
		return a.performPressHome(deviceID)

	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// Helper methods for basic ADB actions
func (a *App) performTap(deviceID string, x, y int) error {
	_, err := a.RunAdbCommand(deviceID, fmt.Sprintf("shell input tap %d %d", x, y))
	return err
}

func (a *App) performSwipe(deviceID string, x1, y1, x2, y2, duration int) error {
	_, err := a.RunAdbCommand(deviceID, fmt.Sprintf("shell input swipe %d %d %d %d %d", x1, y1, x2, y2, duration))
	return err
}

func (a *App) performInputText(deviceID, text string) error {
	text = strings.ReplaceAll(text, " ", "%s")
	_, err := a.RunAdbCommand(deviceID, fmt.Sprintf("shell input text '%s'", text))
	return err
}

func (a *App) performPressBack(deviceID string) error {
	_, err := a.RunAdbCommand(deviceID, "shell input keyevent 4")
	return err
}

func (a *App) performPressHome(deviceID string) error {
	_, err := a.RunAdbCommand(deviceID, "shell input keyevent 3")
	return err
}
