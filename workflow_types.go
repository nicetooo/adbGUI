package main

import (
	"encoding/json"
	"fmt"
)

// ============== Step Connections ==============

// StepConnections defines the flow control for a workflow step
type StepConnections struct {
	// All nodes have these
	SuccessStepId string `json:"successStepId,omitempty"` // Next step on success
	ErrorStepId   string `json:"errorStepId,omitempty"`   // Next step on error

	// Branch node additional outputs
	TrueStepId  string `json:"trueStepId,omitempty"`  // Branch condition true
	FalseStepId string `json:"falseStepId,omitempty"` // Branch condition false
}

// ============== Step Common Config ==============

// StepCommon contains common configuration for all step types
type StepCommon struct {
	Timeout   int    `json:"timeout,omitempty"`   // Timeout in milliseconds
	OnError   string `json:"onError,omitempty"`   // "stop" or "continue"
	Loop      int    `json:"loop,omitempty"`      // Number of times to repeat
	PostDelay int    `json:"postDelay,omitempty"` // Delay after execution (ms)
	PreWait   int    `json:"preWait,omitempty"`   // Wait before execution (ms)
}

// ============== Step Layout ==============

// HandleInfo stores React Flow handle connection info
type HandleInfo struct {
	SourceHandle string `json:"sourceHandle,omitempty"`
	TargetHandle string `json:"targetHandle,omitempty"`
}

// StepLayout contains UI layout information (separated from business logic)
type StepLayout struct {
	PosX float64 `json:"posX,omitempty"`
	PosY float64 `json:"posY,omitempty"`

	// React Flow handle mappings
	Handles map[string]HandleInfo `json:"handles,omitempty"`
}

// ============== Type-Specific Parameters ==============

// TapParams for tap action
type TapParams struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// SwipeParams for swipe action
type SwipeParams struct {
	// Coordinate mode
	X  int `json:"x,omitempty"`
	Y  int `json:"y,omitempty"`
	X2 int `json:"x2,omitempty"`
	Y2 int `json:"y2,omitempty"`

	// Direction mode
	Direction string `json:"direction,omitempty"` // up/down/left/right
	Distance  int    `json:"distance,omitempty"`

	// Common
	Duration int `json:"duration,omitempty"` // Swipe duration in ms
}

// ElementParams for element-based actions
type ElementParams struct {
	Selector ElementSelector `json:"selector"`
	Action   string          `json:"action"` // click/long_click/input/swipe/wait/wait_gone/assert

	// Action-specific parameters
	InputText     string `json:"inputText,omitempty"`     // For action=input
	SwipeDir      string `json:"swipeDir,omitempty"`      // For action=swipe (up/down/left/right)
	SwipeDistance int    `json:"swipeDistance,omitempty"` // For action=swipe
	SwipeDuration int    `json:"swipeDuration,omitempty"` // For action=swipe
}

// AppParams for app operations
type AppParams struct {
	PackageName string `json:"packageName"`
	Action      string `json:"action"` // launch/stop/clear/settings
}

// BranchParams for conditional branching
type BranchParams struct {
	Condition string `json:"condition"` // exists/not_exists/text_equals/text_contains/variable_equals

	// Condition parameters
	Selector      *ElementSelector `json:"selector,omitempty"`
	ExpectedValue string           `json:"expectedValue,omitempty"`
	VariableName  string           `json:"variableName,omitempty"`
}

// WaitParams for wait action
type WaitParams struct {
	DurationMs int `json:"durationMs"`
}

// ScriptParams for running recorded scripts
type ScriptParams struct {
	ScriptName string `json:"scriptName"`
}

// VariableParams for setting variables
type VariableParams struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ADBParams for raw ADB commands
type ADBParams struct {
	Command string `json:"command"`
}

// SubWorkflowParams for running sub-workflows
type SubWorkflowParams struct {
	WorkflowId string `json:"workflowId"`
}

// ============== WorkflowStep ==============

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`

	// Common configuration
	Common StepCommon `json:"common,omitempty"`

	// Flow connections
	Connections StepConnections `json:"connections,omitempty"`

	// Type-specific parameters (only one should be set based on Type)
	Tap      *TapParams         `json:"tap,omitempty"`
	Swipe    *SwipeParams       `json:"swipe,omitempty"`
	Element  *ElementParams     `json:"element,omitempty"`
	App      *AppParams         `json:"app,omitempty"`
	Branch   *BranchParams      `json:"branch,omitempty"`
	Wait     *WaitParams        `json:"wait,omitempty"`
	Script   *ScriptParams      `json:"script,omitempty"`
	Variable *VariableParams    `json:"variable,omitempty"`
	ADB      *ADBParams         `json:"adb,omitempty"`
	Workflow *SubWorkflowParams `json:"workflow,omitempty"`

	// UI Layout (separated from business logic)
	Layout StepLayout `json:"layout,omitempty"`
}

// ============== Workflow ==============

// Workflow represents a complete workflow definition
type Workflow struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Version     int               `json:"version,omitempty"` // Schema version (2 = current V2 format)
	Steps       []WorkflowStep    `json:"steps"`
	Variables   map[string]string `json:"variables,omitempty"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
}

// ============== Validation ==============

// ValidationError represents a validation error
type ValidationError struct {
	StepID  string `json:"stepId,omitempty"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	if e.StepID != "" {
		return fmt.Sprintf("step %s: %s - %s", e.StepID, e.Field, e.Message)
	}
	return fmt.Sprintf("%s - %s", e.Field, e.Message)
}

// Validate validates the workflow step
func (s *WorkflowStep) Validate() error {
	if s.ID == "" {
		return ValidationError{Field: "id", Message: "step id is required"}
	}
	if s.Type == "" {
		return ValidationError{StepID: s.ID, Field: "type", Message: "step type is required"}
	}

	switch s.Type {
	case "start":
		// Start node has no specific params
		return nil

	case "tap":
		if s.Tap == nil {
			return ValidationError{StepID: s.ID, Field: "tap", Message: "tap params required for tap step"}
		}

	case "swipe":
		if s.Swipe == nil {
			return ValidationError{StepID: s.ID, Field: "swipe", Message: "swipe params required for swipe step"}
		}
		// Either coordinate mode or direction mode
		hasCoords := s.Swipe.X != 0 || s.Swipe.Y != 0 || s.Swipe.X2 != 0 || s.Swipe.Y2 != 0
		hasDirection := s.Swipe.Direction != ""
		if !hasCoords && !hasDirection {
			return ValidationError{StepID: s.ID, Field: "swipe", Message: "swipe requires coordinates or direction"}
		}

	case "click_element", "long_click_element", "input_text", "swipe_element", "wait_element", "wait_gone", "assert_element":
		if s.Element == nil {
			return ValidationError{StepID: s.ID, Field: "element", Message: "element params required for element step"}
		}
		if s.Element.Selector.Type == "" || s.Element.Selector.Value == "" {
			return ValidationError{StepID: s.ID, Field: "element.selector", Message: "selector type and value are required"}
		}
		if s.Type == "input_text" && s.Element.InputText == "" {
			return ValidationError{StepID: s.ID, Field: "element.inputText", Message: "inputText required for input_text step"}
		}

	case "launch_app", "stop_app", "clear_app", "open_settings":
		if s.App == nil {
			return ValidationError{StepID: s.ID, Field: "app", Message: "app params required for app step"}
		}
		if s.App.PackageName == "" {
			return ValidationError{StepID: s.ID, Field: "app.packageName", Message: "packageName is required"}
		}

	case "branch":
		if s.Branch == nil {
			return ValidationError{StepID: s.ID, Field: "branch", Message: "branch params required for branch step"}
		}
		if s.Branch.Condition == "" {
			return ValidationError{StepID: s.ID, Field: "branch.condition", Message: "condition is required"}
		}
		// For element-based conditions, selector is required
		if s.Branch.Condition != "variable_equals" && s.Branch.Selector == nil {
			return ValidationError{StepID: s.ID, Field: "branch.selector", Message: "selector required for element condition"}
		}

	case "wait":
		if s.Wait == nil {
			return ValidationError{StepID: s.ID, Field: "wait", Message: "wait params required for wait step"}
		}
		if s.Wait.DurationMs <= 0 {
			return ValidationError{StepID: s.ID, Field: "wait.durationMs", Message: "durationMs must be positive"}
		}

	case "script":
		if s.Script == nil {
			return ValidationError{StepID: s.ID, Field: "script", Message: "script params required for script step"}
		}
		if s.Script.ScriptName == "" {
			return ValidationError{StepID: s.ID, Field: "script.scriptName", Message: "scriptName is required"}
		}

	case "set_variable":
		if s.Variable == nil {
			return ValidationError{StepID: s.ID, Field: "variable", Message: "variable params required for set_variable step"}
		}
		if s.Variable.Name == "" {
			return ValidationError{StepID: s.ID, Field: "variable.name", Message: "variable name is required"}
		}

	case "adb":
		if s.ADB == nil {
			return ValidationError{StepID: s.ID, Field: "adb", Message: "adb params required for adb step"}
		}
		if s.ADB.Command == "" {
			return ValidationError{StepID: s.ID, Field: "adb.command", Message: "command is required"}
		}

	case "run_workflow":
		if s.Workflow == nil {
			return ValidationError{StepID: s.ID, Field: "workflow", Message: "workflow params required for run_workflow step"}
		}
		if s.Workflow.WorkflowId == "" {
			return ValidationError{StepID: s.ID, Field: "workflow.workflowId", Message: "workflowId is required"}
		}

	case "key_back", "key_home", "key_recent", "key_power", "key_volume_up", "key_volume_down", "screen_on", "screen_off":
		// System key events have no specific params
		return nil

	default:
		return ValidationError{StepID: s.ID, Field: "type", Message: fmt.Sprintf("unknown step type: %s", s.Type)}
	}

	return nil
}

// Validate validates the entire workflow
func (w *Workflow) Validate() []error {
	var errors []error

	if w.ID == "" {
		errors = append(errors, ValidationError{Field: "id", Message: "workflow id is required"})
	}
	if w.Name == "" {
		errors = append(errors, ValidationError{Field: "name", Message: "workflow name is required"})
	}

	// Check for start node
	hasStart := false
	stepMap := make(map[string]*WorkflowStep)
	for i := range w.Steps {
		step := &w.Steps[i]
		stepMap[step.ID] = step
		if step.Type == "start" {
			if hasStart {
				errors = append(errors, ValidationError{Field: "steps", Message: "workflow can only have one start node"})
			}
			hasStart = true
		}
	}

	if !hasStart && len(w.Steps) > 0 {
		errors = append(errors, ValidationError{Field: "steps", Message: "workflow must have a start node"})
	}

	// Validate each step
	for i := range w.Steps {
		if err := w.Steps[i].Validate(); err != nil {
			errors = append(errors, err)
		}
	}

	// Validate connections
	for i := range w.Steps {
		step := &w.Steps[i]
		connErrors := validateStepConnections(step, stepMap)
		errors = append(errors, connErrors...)
	}

	return errors
}

// validateStepConnections validates that step connections point to existing steps
func validateStepConnections(step *WorkflowStep, stepMap map[string]*WorkflowStep) []error {
	var errors []error

	if step.Connections.SuccessStepId != "" {
		if _, exists := stepMap[step.Connections.SuccessStepId]; !exists {
			errors = append(errors, ValidationError{
				StepID:  step.ID,
				Field:   "connections.successStepId",
				Message: fmt.Sprintf("target step '%s' not found", step.Connections.SuccessStepId),
			})
		}
	}

	if step.Connections.ErrorStepId != "" {
		if _, exists := stepMap[step.Connections.ErrorStepId]; !exists {
			errors = append(errors, ValidationError{
				StepID:  step.ID,
				Field:   "connections.errorStepId",
				Message: fmt.Sprintf("target step '%s' not found", step.Connections.ErrorStepId),
			})
		}
	}

	if step.Connections.TrueStepId != "" {
		if _, exists := stepMap[step.Connections.TrueStepId]; !exists {
			errors = append(errors, ValidationError{
				StepID:  step.ID,
				Field:   "connections.trueStepId",
				Message: fmt.Sprintf("target step '%s' not found", step.Connections.TrueStepId),
			})
		}
	}

	if step.Connections.FalseStepId != "" {
		if _, exists := stepMap[step.Connections.FalseStepId]; !exists {
			errors = append(errors, ValidationError{
				StepID:  step.ID,
				Field:   "connections.falseStepId",
				Message: fmt.Sprintf("target step '%s' not found", step.Connections.FalseStepId),
			})
		}
	}

	return errors
}

// ============== Helper Methods ==============

// GetNextStepId returns the next step to execute based on execution result
func (s *WorkflowStep) GetNextStepId(success bool, isBranchResult bool) string {
	if s.Type == "branch" && isBranchResult {
		// For branch nodes, success/false indicates condition result
		if success {
			return s.Connections.TrueStepId
		}
		return s.Connections.FalseStepId
	}

	// For all other nodes (or branch execution error)
	if success {
		return s.Connections.SuccessStepId
	}
	return s.Connections.ErrorStepId
}

// ShouldStopOnError returns whether workflow should stop when this step fails
// and error path is not connected
func (s *WorkflowStep) ShouldStopOnError() bool {
	// If error path is connected, don't stop
	if s.Connections.ErrorStepId != "" {
		return false
	}
	// Otherwise, check OnError setting
	return s.Common.OnError != "continue"
}

// GetFallbackStepId returns the step to go to when error occurs but onError=continue
func (s *WorkflowStep) GetFallbackStepId() string {
	if s.Connections.ErrorStepId != "" {
		return s.Connections.ErrorStepId
	}
	if s.Common.OnError == "continue" {
		return s.Connections.SuccessStepId
	}
	return ""
}

// ============== JSON Helpers ==============

// SerializeWorkflow serializes a workflow to JSON
func SerializeWorkflow(workflow *Workflow) ([]byte, error) {
	return json.MarshalIndent(workflow, "", "  ")
}

// ParseWorkflow parses workflow JSON
func ParseWorkflow(data []byte) (*Workflow, error) {
	var workflow Workflow
	if err := json.Unmarshal(data, &workflow); err != nil {
		return nil, err
	}
	return &workflow, nil
}
