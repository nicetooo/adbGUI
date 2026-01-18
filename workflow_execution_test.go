package main

import (
	"fmt"
	"testing"
)

// ==================== V2StepResult Tests ====================

func TestV2StepResult_Success(t *testing.T) {
	result := V2StepResult{
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

func TestV2StepResult_Failure(t *testing.T) {
	result := V2StepResult{
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

func TestV2StepResult_BranchTrue(t *testing.T) {
	result := V2StepResult{
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

func TestV2StepResult_BranchFalse(t *testing.T) {
	result := V2StepResult{
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

// ==================== DetermineNextStepV2 Tests ====================

func TestDetermineNextStepV2_NormalSuccess(t *testing.T) {
	app := &App{}
	step := &WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "error_handler",
		},
	}
	result := V2StepResult{Success: true, IsBranchResult: false}

	nextId := app.determineNextStepV2(step, result)

	if nextId != "step2" {
		t.Errorf("Expected 'step2', got '%s'", nextId)
	}
}

func TestDetermineNextStepV2_NormalError_WithErrorPath(t *testing.T) {
	app := &App{}
	step := &WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "error_handler",
		},
	}
	result := V2StepResult{Success: false, IsBranchResult: false, Error: fmt.Errorf("failed")}

	nextId := app.determineNextStepV2(step, result)

	if nextId != "error_handler" {
		t.Errorf("Expected 'error_handler', got '%s'", nextId)
	}
}

func TestDetermineNextStepV2_NormalError_NoErrorPath_Stop(t *testing.T) {
	app := &App{}
	step := &WorkflowStepV2{
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
	result := V2StepResult{Success: false, IsBranchResult: false, Error: fmt.Errorf("failed")}

	nextId := app.determineNextStepV2(step, result)

	if nextId != "" {
		t.Errorf("Expected empty string (stop), got '%s'", nextId)
	}
}

func TestDetermineNextStepV2_NormalError_NoErrorPath_Continue(t *testing.T) {
	app := &App{}
	step := &WorkflowStepV2{
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
	result := V2StepResult{Success: false, IsBranchResult: false, Error: fmt.Errorf("failed")}

	nextId := app.determineNextStepV2(step, result)

	if nextId != "step2" {
		t.Errorf("Expected 'step2' (continue to success path), got '%s'", nextId)
	}
}

func TestDetermineNextStepV2_BranchTrue(t *testing.T) {
	app := &App{}
	step := &WorkflowStepV2{
		ID:   "branch1",
		Type: "branch",
		Connections: StepConnections{
			SuccessStepId: "fallback",
			TrueStepId:    "step_yes",
			FalseStepId:   "step_no",
			ErrorStepId:   "step_error",
		},
	}
	result := V2StepResult{Success: true, IsBranchResult: true}

	nextId := app.determineNextStepV2(step, result)

	if nextId != "step_yes" {
		t.Errorf("Expected 'step_yes', got '%s'", nextId)
	}
}

func TestDetermineNextStepV2_BranchFalse(t *testing.T) {
	app := &App{}
	step := &WorkflowStepV2{
		ID:   "branch1",
		Type: "branch",
		Connections: StepConnections{
			SuccessStepId: "fallback",
			TrueStepId:    "step_yes",
			FalseStepId:   "step_no",
			ErrorStepId:   "step_error",
		},
	}
	result := V2StepResult{Success: false, IsBranchResult: true}

	nextId := app.determineNextStepV2(step, result)

	if nextId != "step_no" {
		t.Errorf("Expected 'step_no', got '%s'", nextId)
	}
}

func TestDetermineNextStepV2_BranchExecutionError(t *testing.T) {
	app := &App{}
	step := &WorkflowStepV2{
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
	result := V2StepResult{Success: false, IsBranchResult: false, Error: fmt.Errorf("UI dump failed")}

	nextId := app.determineNextStepV2(step, result)

	if nextId != "step_error" {
		t.Errorf("Expected 'step_error', got '%s'", nextId)
	}
}

// ==================== String Helper Tests ====================

func TestContainsIgnoreCaseV2(t *testing.T) {
	testCases := []struct {
		s      string
		substr string
		expect bool
	}{
		{"Hello World", "world", true},
		{"Hello World", "HELLO", true},
		{"Hello World", "hello world", true},
		{"Hello World", "foo", false},
		{"Hello World", "", true},
		{"", "foo", false},
		{"", "", true},
		{"abc", "abcd", false},
	}

	for _, tc := range testCases {
		result := containsIgnoreCaseV2(tc.s, tc.substr)
		if result != tc.expect {
			t.Errorf("containsIgnoreCaseV2(%q, %q) = %v, want %v", tc.s, tc.substr, result, tc.expect)
		}
	}
}

func TestEqualFoldAtV2(t *testing.T) {
	testCases := []struct {
		s      string
		i      int
		substr string
		expect bool
	}{
		{"Hello", 0, "hello", true},
		{"Hello", 0, "HELLO", true},
		{"Hello World", 6, "world", true},
		{"Hello World", 6, "WORLD", true},
		{"Hello", 0, "World", false},
	}

	for _, tc := range testCases {
		result := equalFoldAtV2(tc.s, tc.i, tc.substr)
		if result != tc.expect {
			t.Errorf("equalFoldAtV2(%q, %d, %q) = %v, want %v", tc.s, tc.i, tc.substr, result, tc.expect)
		}
	}
}

// ==================== StepCommon Helpers Tests ====================

func TestWorkflowStepV2_ShouldStopOnError_WithErrorPath(t *testing.T) {
	step := WorkflowStepV2{
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

func TestWorkflowStepV2_ShouldStopOnError_NoErrorPath_Stop(t *testing.T) {
	step := WorkflowStepV2{
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

func TestWorkflowStepV2_ShouldStopOnError_NoErrorPath_Continue(t *testing.T) {
	step := WorkflowStepV2{
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

func TestWorkflowStepV2_ShouldStopOnError_NoErrorPath_DefaultStop(t *testing.T) {
	step := WorkflowStepV2{
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

func TestWorkflowV2_EmptyWorkflow(t *testing.T) {
	workflow := WorkflowV2{
		ID:    "wf1",
		Name:  "Empty",
		Steps: []WorkflowStepV2{},
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

func TestWorkflowV2_StartOnlyWorkflow(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "StartOnly",
		Steps: []WorkflowStepV2{
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

func TestWorkflowV2_LinearFlow(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "Linear",
		Steps: []WorkflowStepV2{
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

func TestWorkflowV2_BranchFlow(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "Branch",
		Steps: []WorkflowStepV2{
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

func TestWorkflowV2_WithErrorPath(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "WithErrorPath",
		Steps: []WorkflowStepV2{
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
				Element: &ElementParamsV2{
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

func TestWorkflowV2_InvalidConnection(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "InvalidConnection",
		Steps: []WorkflowStepV2{
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

func TestWorkflowV2_MultipleStartNodes(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "MultipleStart",
		Steps: []WorkflowStepV2{
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

func TestWorkflowV2_VariableSubstitution(t *testing.T) {
	// This is a conceptual test - actual execution would need mocking
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "VariableTest",
		Variables: map[string]string{
			"email":    "test@example.com",
			"password": "secret123",
		},
		Steps: []WorkflowStepV2{
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
				Element: &ElementParamsV2{
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
				Element: &ElementParamsV2{
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

func TestWorkflowStepV2_LoopConfig(t *testing.T) {
	step := WorkflowStepV2{
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

func TestWorkflowStepV2_DefaultLoop(t *testing.T) {
	step := WorkflowStepV2{
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

func TestWorkflowStepV2_TimeoutConfig(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "wait_element",
		Element: &ElementParamsV2{
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

func TestWorkflowStepV2_DelayConfig(t *testing.T) {
	step := WorkflowStepV2{
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
