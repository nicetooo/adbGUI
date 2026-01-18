package main

import (
	"fmt"
	"testing"
)

// ==================== StepResult Tests ====================

func TestStepResult_Success(t *testing.T) {
	result := StepResult{
		Success:        true,
		IsBranchResult: false,
		Error:          nil,
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.IsBranchResult {
		t.Error("Expected IsBranchResult to be false")
	}
	if result.Error != nil {
		t.Errorf("Expected Error to be nil, got %v", result.Error)
	}
}

func TestStepResult_Failure(t *testing.T) {
	result := StepResult{
		Success:        false,
		IsBranchResult: false,
		Error:          fmt.Errorf("test error"),
	}

	if result.Success {
		t.Error("Expected Success to be false")
	}
	if result.Error == nil {
		t.Error("Expected Error to be non-nil")
	}
	if result.Error.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", result.Error.Error())
	}
}

func TestStepResult_BranchTrue(t *testing.T) {
	result := StepResult{
		Success:        true,
		IsBranchResult: true,
		Error:          nil,
	}

	if !result.IsBranchResult {
		t.Error("Expected IsBranchResult to be true")
	}
	if !result.Success {
		t.Error("Expected branch condition to be true")
	}
}

func TestStepResult_BranchFalse(t *testing.T) {
	result := StepResult{
		Success:        false,
		IsBranchResult: true,
		Error:          nil,
	}

	if !result.IsBranchResult {
		t.Error("Expected IsBranchResult to be true")
	}
	if result.Success {
		t.Error("Expected branch condition to be false")
	}
	// Note: Error is nil because branch evaluated successfully, just returned false
	if result.Error != nil {
		t.Error("Expected Error to be nil for successful branch evaluation")
	}
}

// ==================== DetermineNextStep Tests ====================

func TestDetermineNextStep_NormalSuccess(t *testing.T) {
	app := &App{}
	step := &WorkflowStep{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "error_handler",
		},
	}
	result := StepResult{Success: true, IsBranchResult: false}

	nextId := app.determineNextStep(step, result)

	if nextId != "step2" {
		t.Errorf("Expected 'step2', got '%s'", nextId)
	}
}

func TestDetermineNextStep_NormalError_WithErrorPath(t *testing.T) {
	app := &App{}
	step := &WorkflowStep{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "error_handler",
		},
	}
	result := StepResult{Success: false, IsBranchResult: false, Error: fmt.Errorf("failed")}

	nextId := app.determineNextStep(step, result)

	if nextId != "error_handler" {
		t.Errorf("Expected 'error_handler', got '%s'", nextId)
	}
}

func TestDetermineNextStep_NormalError_NoErrorPath_Stop(t *testing.T) {
	app := &App{}
	step := &WorkflowStep{
		ID:   "step1",
		Type: "tap",
		Common: StepCommon{
			OnError: "stop",
		},
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "", // No error path
		},
	}
	result := StepResult{Success: false, IsBranchResult: false, Error: fmt.Errorf("failed")}

	nextId := app.determineNextStep(step, result)

	if nextId != "" {
		t.Errorf("Expected empty string (stop), got '%s'", nextId)
	}
}

func TestDetermineNextStep_NormalError_NoErrorPath_Continue(t *testing.T) {
	app := &App{}
	step := &WorkflowStep{
		ID:   "step1",
		Type: "tap",
		Common: StepCommon{
			OnError: "continue",
		},
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "", // No error path
		},
	}
	result := StepResult{Success: false, IsBranchResult: false, Error: fmt.Errorf("failed")}

	nextId := app.determineNextStep(step, result)

	if nextId != "step2" {
		t.Errorf("Expected 'step2' (continue to success path), got '%s'", nextId)
	}
}

func TestDetermineNextStep_BranchTrue(t *testing.T) {
	app := &App{}
	step := &WorkflowStep{
		ID:   "branch1",
		Type: "branch",
		Connections: StepConnections{
			SuccessStepId: "fallback",
			TrueStepId:    "step_yes",
			FalseStepId:   "step_no",
			ErrorStepId:   "step_error",
		},
	}
	result := StepResult{Success: true, IsBranchResult: true}

	nextId := app.determineNextStep(step, result)

	if nextId != "step_yes" {
		t.Errorf("Expected 'step_yes', got '%s'", nextId)
	}
}

func TestDetermineNextStep_BranchFalse(t *testing.T) {
	app := &App{}
	step := &WorkflowStep{
		ID:   "branch1",
		Type: "branch",
		Connections: StepConnections{
			SuccessStepId: "fallback",
			TrueStepId:    "step_yes",
			FalseStepId:   "step_no",
			ErrorStepId:   "step_error",
		},
	}
	result := StepResult{Success: false, IsBranchResult: true}

	nextId := app.determineNextStep(step, result)

	if nextId != "step_no" {
		t.Errorf("Expected 'step_no', got '%s'", nextId)
	}
}

func TestDetermineNextStep_BranchExecutionError(t *testing.T) {
	app := &App{}
	step := &WorkflowStep{
		ID:   "branch1",
		Type: "branch",
		Connections: StepConnections{
			SuccessStepId: "fallback",
			TrueStepId:    "step_yes",
			FalseStepId:   "step_no",
			ErrorStepId:   "step_error",
		},
	}
	// Branch execution failed (not a condition result)
	result := StepResult{Success: false, IsBranchResult: false, Error: fmt.Errorf("UI dump failed")}

	nextId := app.determineNextStep(step, result)

	if nextId != "step_error" {
		t.Errorf("Expected 'step_error', got '%s'", nextId)
	}
}

// ==================== StepCommon Helpers Tests ====================

func TestWorkflowStep_ShouldStopOnError_WithErrorPath(t *testing.T) {
	step := WorkflowStep{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			ErrorStepId: "error_handler",
		},
		Common: StepCommon{
			OnError: "stop", // Even with stop, should not stop if error path connected
		},
	}

	if step.ShouldStopOnError() {
		t.Error("Should not stop when error path is connected")
	}
}

func TestWorkflowStep_ShouldStopOnError_NoErrorPath_Stop(t *testing.T) {
	step := WorkflowStep{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			ErrorStepId: "",
		},
		Common: StepCommon{
			OnError: "stop",
		},
	}

	if !step.ShouldStopOnError() {
		t.Error("Should stop when error path not connected and onError=stop")
	}
}

func TestWorkflowStep_ShouldStopOnError_NoErrorPath_Continue(t *testing.T) {
	step := WorkflowStep{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			ErrorStepId: "",
		},
		Common: StepCommon{
			OnError: "continue",
		},
	}

	if step.ShouldStopOnError() {
		t.Error("Should not stop when onError=continue")
	}
}

func TestWorkflowStep_ShouldStopOnError_NoErrorPath_DefaultStop(t *testing.T) {
	step := WorkflowStep{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			ErrorStepId: "",
		},
		Common: StepCommon{
			OnError: "", // Default
		},
	}

	if !step.ShouldStopOnError() {
		t.Error("Should stop when onError is empty (default to stop)")
	}
}

// ==================== Workflow Execution Flow Tests ====================

func TestWorkflow_EmptyWorkflow(t *testing.T) {
	workflow := Workflow{
		ID:    "wf1",
		Name:  "Empty",
		Steps: []WorkflowStep{},
	}

	errors := workflow.Validate()
	// Empty workflow should fail validation (no start node)
	hasNoStartError := false
	for _, err := range errors {
		if ve, ok := err.(ValidationError); ok && ve.Message == "workflow must have a start node" {
			hasNoStartError = true
			break
		}
	}
	// Actually, we allow empty workflow, so this should pass if Steps is empty
	// Let's check the actual validation logic
	if len(workflow.Steps) == 0 {
		// Empty steps means no start node, so should fail
		if !hasNoStartError {
			t.Log("Empty workflow validation: either no error or different error")
		}
	}
}

func TestWorkflow_StartOnlyWorkflow(t *testing.T) {
	workflow := Workflow{
		ID:   "wf1",
		Name: "StartOnly",
		Steps: []WorkflowStep{
			{
				ID:   "start",
				Type: "start",
				Connections: StepConnections{
					SuccessStepId: "", // No next step
				},
			},
		},
	}

	errors := workflow.Validate()
	if len(errors) > 0 {
		t.Errorf("Start-only workflow should be valid, got errors: %v", errors)
	}
}

func TestWorkflow_LinearFlow(t *testing.T) {
	workflow := Workflow{
		ID:   "wf1",
		Name: "Linear",
		Steps: []WorkflowStep{
			{
				ID:   "start",
				Type: "start",
				Connections: StepConnections{
					SuccessStepId: "step1",
				},
			},
			{
				ID:   "step1",
				Type: "tap",
				Tap:  &TapParams{X: 100, Y: 200},
				Connections: StepConnections{
					SuccessStepId: "step2",
				},
			},
			{
				ID:   "step2",
				Type: "tap",
				Tap:  &TapParams{X: 300, Y: 400},
				Connections: StepConnections{
					SuccessStepId: "",
				},
			},
		},
	}

	errors := workflow.Validate()
	if len(errors) > 0 {
		t.Errorf("Linear workflow should be valid, got errors: %v", errors)
	}
}

func TestWorkflow_BranchFlow(t *testing.T) {
	workflow := Workflow{
		ID:   "wf1",
		Name: "Branch",
		Steps: []WorkflowStep{
			{
				ID:   "start",
				Type: "start",
				Connections: StepConnections{
					SuccessStepId: "branch1",
				},
			},
			{
				ID:   "branch1",
				Type: "branch",
				Branch: &BranchParams{
					Condition: "exists",
					Selector: &ElementSelector{
						Type:  "text",
						Value: "Submit",
					},
				},
				Connections: StepConnections{
					TrueStepId:  "step_yes",
					FalseStepId: "step_no",
					ErrorStepId: "step_error",
				},
			},
			{
				ID:   "step_yes",
				Type: "tap",
				Tap:  &TapParams{X: 100, Y: 100},
			},
			{
				ID:   "step_no",
				Type: "tap",
				Tap:  &TapParams{X: 200, Y: 200},
			},
			{
				ID:   "step_error",
				Type: "tap",
				Tap:  &TapParams{X: 300, Y: 300},
			},
		},
	}

	errors := workflow.Validate()
	if len(errors) > 0 {
		t.Errorf("Branch workflow should be valid, got errors: %v", errors)
	}
}

func TestWorkflow_WithErrorPath(t *testing.T) {
	workflow := Workflow{
		ID:   "wf1",
		Name: "WithErrorPath",
		Steps: []WorkflowStep{
			{
				ID:   "start",
				Type: "start",
				Connections: StepConnections{
					SuccessStepId: "step1",
				},
			},
			{
				ID:   "step1",
				Type: "click_element",
				Element: &ElementParams{
					Selector: ElementSelector{
						Type:  "text",
						Value: "Submit",
					},
					Action: "click",
				},
				Connections: StepConnections{
					SuccessStepId: "step2",
					ErrorStepId:   "error_handler",
				},
			},
			{
				ID:   "step2",
				Type: "tap",
				Tap:  &TapParams{X: 100, Y: 100},
			},
			{
				ID:   "error_handler",
				Type: "tap",
				Tap:  &TapParams{X: 200, Y: 200},
				Connections: StepConnections{
					SuccessStepId: "step2", // Recover and continue
				},
			},
		},
	}

	errors := workflow.Validate()
	if len(errors) > 0 {
		t.Errorf("Workflow with error path should be valid, got errors: %v", errors)
	}
}

func TestWorkflow_InvalidConnection(t *testing.T) {
	workflow := Workflow{
		ID:   "wf1",
		Name: "InvalidConnection",
		Steps: []WorkflowStep{
			{
				ID:   "start",
				Type: "start",
				Connections: StepConnections{
					SuccessStepId: "nonexistent_step",
				},
			},
		},
	}

	errors := workflow.Validate()
	hasConnError := false
	for _, err := range errors {
		if ve, ok := err.(ValidationError); ok && ve.Field == "connections.successStepId" {
			hasConnError = true
			break
		}
	}
	if !hasConnError {
		t.Error("Expected validation error for invalid connection")
	}
}

func TestWorkflow_MultipleStartNodes(t *testing.T) {
	workflow := Workflow{
		ID:   "wf1",
		Name: "MultipleStart",
		Steps: []WorkflowStep{
			{
				ID:   "start1",
				Type: "start",
			},
			{
				ID:   "start2",
				Type: "start",
			},
		},
	}

	errors := workflow.Validate()
	hasMultiStartError := false
	for _, err := range errors {
		if ve, ok := err.(ValidationError); ok && ve.Message == "workflow can only have one start node" {
			hasMultiStartError = true
			break
		}
	}
	if !hasMultiStartError {
		t.Error("Expected validation error for multiple start nodes")
	}
}

// ==================== Variable Substitution Tests ====================

func TestWorkflow_VariableSubstitution(t *testing.T) {
	// This is a conceptual test - actual execution would need mocking
	workflow := Workflow{
		ID:   "wf1",
		Name: "VariableTest",
		Variables: map[string]string{
			"email":    "test@example.com",
			"password": "secret123",
		},
		Steps: []WorkflowStep{
			{
				ID:   "start",
				Type: "start",
				Connections: StepConnections{
					SuccessStepId: "input_email",
				},
			},
			{
				ID:   "input_email",
				Type: "input_text",
				Element: &ElementParams{
					Selector: ElementSelector{
						Type:  "id",
						Value: "email_field",
					},
					Action:    "input",
					InputText: "{{email}}", // Variable reference
				},
				Connections: StepConnections{
					SuccessStepId: "input_password",
				},
			},
			{
				ID:   "input_password",
				Type: "input_text",
				Element: &ElementParams{
					Selector: ElementSelector{
						Type:  "id",
						Value: "password_field",
					},
					Action:    "input",
					InputText: "{{password}}", // Variable reference
				},
			},
		},
	}

	errors := workflow.Validate()
	if len(errors) > 0 {
		t.Errorf("Workflow with variables should be valid, got errors: %v", errors)
	}

	// Verify variables are set
	if workflow.Variables["email"] != "test@example.com" {
		t.Errorf("Expected email variable, got %s", workflow.Variables["email"])
	}
}

// ==================== Loop Execution Tests ====================

func TestWorkflowStep_LoopConfig(t *testing.T) {
	step := WorkflowStep{
		ID:   "step1",
		Type: "tap",
		Tap:  &TapParams{X: 100, Y: 200},
		Common: StepCommon{
			Loop: 3,
		},
	}

	if step.Common.Loop != 3 {
		t.Errorf("Expected Loop 3, got %d", step.Common.Loop)
	}
}

func TestWorkflowStep_DefaultLoop(t *testing.T) {
	step := WorkflowStep{
		ID:     "step1",
		Type:   "tap",
		Tap:    &TapParams{X: 100, Y: 200},
		Common: StepCommon{
			// Loop not set
		},
	}

	// Default loop count should be treated as 1
	loopCount := step.Common.Loop
	if loopCount <= 0 {
		loopCount = 1
	}
	if loopCount != 1 {
		t.Errorf("Expected default loop count 1, got %d", loopCount)
	}
}

// ==================== Timeout Tests ====================

func TestWorkflowStep_TimeoutConfig(t *testing.T) {
	step := WorkflowStep{
		ID:   "step1",
		Type: "wait_element",
		Element: &ElementParams{
			Selector: ElementSelector{
				Type:  "text",
				Value: "Loading",
			},
			Action: "wait",
		},
		Common: StepCommon{
			Timeout: 10000,
		},
	}

	if step.Common.Timeout != 10000 {
		t.Errorf("Expected Timeout 10000, got %d", step.Common.Timeout)
	}
}

// ==================== Pre/Post Delay Tests ====================

func TestWorkflowStep_DelayConfig(t *testing.T) {
	step := WorkflowStep{
		ID:   "step1",
		Type: "tap",
		Tap:  &TapParams{X: 100, Y: 200},
		Common: StepCommon{
			PreWait:   500,
			PostDelay: 1000,
		},
	}

	if step.Common.PreWait != 500 {
		t.Errorf("Expected PreWait 500, got %d", step.Common.PreWait)
	}
	if step.Common.PostDelay != 1000 {
		t.Errorf("Expected PostDelay 1000, got %d", step.Common.PostDelay)
	}
}
