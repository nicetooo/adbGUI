package main

import (
	"encoding/json"
	"testing"
)

// ==================== Serialization Tests ====================

func TestWorkflowStepV2Serialization_Tap(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Name: "Click button",
		Tap:  &TapParams{X: 100, Y: 200},
		Common: StepCommon{
			Timeout:   5000,
			OnError:   "stop",
			PostDelay: 500,
		},
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "error_handler",
		},
		Layout: StepLayout{
			PosX: 50.0,
			PosY: 100.0,
		},
	}

	// Marshal
	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var decoded WorkflowStepV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify
	if decoded.ID != "step1" {
		t.Errorf("Expected ID 'step1', got '%s'", decoded.ID)
	}
	if decoded.Type != "tap" {
		t.Errorf("Expected Type 'tap', got '%s'", decoded.Type)
	}
	if decoded.Tap == nil {
		t.Fatal("Tap params should not be nil")
	}
	if decoded.Tap.X != 100 || decoded.Tap.Y != 200 {
		t.Errorf("Expected Tap (100, 200), got (%d, %d)", decoded.Tap.X, decoded.Tap.Y)
	}
	if decoded.Swipe != nil {
		t.Error("Swipe params should be nil for tap step")
	}
	if decoded.Common.Timeout != 5000 {
		t.Errorf("Expected Timeout 5000, got %d", decoded.Common.Timeout)
	}
	if decoded.Connections.SuccessStepId != "step2" {
		t.Errorf("Expected SuccessStepId 'step2', got '%s'", decoded.Connections.SuccessStepId)
	}
	if decoded.Layout.PosX != 50.0 {
		t.Errorf("Expected PosX 50.0, got %f", decoded.Layout.PosX)
	}
}

func TestWorkflowStepV2Serialization_Swipe_Coordinates(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "swipe",
		Swipe: &SwipeParams{
			X:        100,
			Y:        500,
			X2:       100,
			Y2:       100,
			Duration: 300,
		},
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WorkflowStepV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Swipe == nil {
		t.Fatal("Swipe params should not be nil")
	}
	if decoded.Swipe.X != 100 || decoded.Swipe.Y != 500 {
		t.Errorf("Expected start (100, 500), got (%d, %d)", decoded.Swipe.X, decoded.Swipe.Y)
	}
	if decoded.Swipe.X2 != 100 || decoded.Swipe.Y2 != 100 {
		t.Errorf("Expected end (100, 100), got (%d, %d)", decoded.Swipe.X2, decoded.Swipe.Y2)
	}
	if decoded.Swipe.Duration != 300 {
		t.Errorf("Expected Duration 300, got %d", decoded.Swipe.Duration)
	}
}

func TestWorkflowStepV2Serialization_Swipe_Direction(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "swipe",
		Swipe: &SwipeParams{
			X:         200,
			Y:         400,
			Direction: "up",
			Distance:  500,
			Duration:  300,
		},
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WorkflowStepV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Swipe.Direction != "up" {
		t.Errorf("Expected Direction 'up', got '%s'", decoded.Swipe.Direction)
	}
	if decoded.Swipe.Distance != 500 {
		t.Errorf("Expected Distance 500, got %d", decoded.Swipe.Distance)
	}
}

func TestWorkflowStepV2Serialization_Element(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "click_element",
		Element: &ElementParamsV2{
			Selector: ElementSelector{
				Type:  "text",
				Value: "Submit",
				Index: 0,
			},
			Action: "click",
		},
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WorkflowStepV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Element == nil {
		t.Fatal("Element params should not be nil")
	}
	if decoded.Element.Selector.Type != "text" {
		t.Errorf("Expected Selector.Type 'text', got '%s'", decoded.Element.Selector.Type)
	}
	if decoded.Element.Selector.Value != "Submit" {
		t.Errorf("Expected Selector.Value 'Submit', got '%s'", decoded.Element.Selector.Value)
	}
	if decoded.Element.Action != "click" {
		t.Errorf("Expected Action 'click', got '%s'", decoded.Element.Action)
	}
}

func TestWorkflowStepV2Serialization_ElementInput(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "input_text",
		Element: &ElementParamsV2{
			Selector: ElementSelector{
				Type:  "id",
				Value: "email",
			},
			Action:    "input",
			InputText: "test@example.com",
		},
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WorkflowStepV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Element.InputText != "test@example.com" {
		t.Errorf("Expected InputText 'test@example.com', got '%s'", decoded.Element.InputText)
	}
}

func TestWorkflowStepV2Serialization_App(t *testing.T) {
	testCases := []struct {
		name   string
		action string
	}{
		{"launch", "launch"},
		{"stop", "stop"},
		{"clear", "clear"},
		{"settings", "settings"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			step := WorkflowStepV2{
				ID:   "step1",
				Type: tc.name + "_app",
				App: &AppParams{
					PackageName: "com.example.app",
					Action:      tc.action,
				},
			}

			data, err := json.Marshal(step)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			var decoded WorkflowStepV2
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if decoded.App == nil {
				t.Fatal("App params should not be nil")
			}
			if decoded.App.PackageName != "com.example.app" {
				t.Errorf("Expected PackageName 'com.example.app', got '%s'", decoded.App.PackageName)
			}
			if decoded.App.Action != tc.action {
				t.Errorf("Expected Action '%s', got '%s'", tc.action, decoded.App.Action)
			}
		})
	}
}

func TestWorkflowStepV2Serialization_Branch(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "branch1",
		Type: "branch",
		Branch: &BranchParams{
			Condition: "exists",
			Selector: &ElementSelector{
				Type:  "text",
				Value: "Success",
			},
		},
		Connections: StepConnections{
			TrueStepId:  "step_yes",
			FalseStepId: "step_no",
			ErrorStepId: "step_error",
		},
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WorkflowStepV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Branch == nil {
		t.Fatal("Branch params should not be nil")
	}
	if decoded.Branch.Condition != "exists" {
		t.Errorf("Expected Condition 'exists', got '%s'", decoded.Branch.Condition)
	}
	if decoded.Connections.TrueStepId != "step_yes" {
		t.Errorf("Expected TrueStepId 'step_yes', got '%s'", decoded.Connections.TrueStepId)
	}
	if decoded.Connections.FalseStepId != "step_no" {
		t.Errorf("Expected FalseStepId 'step_no', got '%s'", decoded.Connections.FalseStepId)
	}
	if decoded.Connections.ErrorStepId != "step_error" {
		t.Errorf("Expected ErrorStepId 'step_error', got '%s'", decoded.Connections.ErrorStepId)
	}
}

func TestWorkflowStepV2Serialization_Wait(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "wait",
		Wait: &WaitParams{
			DurationMs: 2000,
		},
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WorkflowStepV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Wait == nil {
		t.Fatal("Wait params should not be nil")
	}
	if decoded.Wait.DurationMs != 2000 {
		t.Errorf("Expected DurationMs 2000, got %d", decoded.Wait.DurationMs)
	}
}

func TestWorkflowStepV2Serialization_Script(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "script",
		Script: &ScriptParams{
			ScriptName: "login_flow",
		},
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WorkflowStepV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Script == nil {
		t.Fatal("Script params should not be nil")
	}
	if decoded.Script.ScriptName != "login_flow" {
		t.Errorf("Expected ScriptName 'login_flow', got '%s'", decoded.Script.ScriptName)
	}
}

func TestWorkflowStepV2Serialization_Variable(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "set_variable",
		Variable: &VariableParams{
			Name:  "username",
			Value: "test_user",
		},
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WorkflowStepV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Variable == nil {
		t.Fatal("Variable params should not be nil")
	}
	if decoded.Variable.Name != "username" {
		t.Errorf("Expected Name 'username', got '%s'", decoded.Variable.Name)
	}
	if decoded.Variable.Value != "test_user" {
		t.Errorf("Expected Value 'test_user', got '%s'", decoded.Variable.Value)
	}
}

func TestWorkflowStepV2Serialization_ADB(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "adb",
		ADB: &ADBParams{
			Command: "shell input keyevent 3",
		},
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WorkflowStepV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ADB == nil {
		t.Fatal("ADB params should not be nil")
	}
	if decoded.ADB.Command != "shell input keyevent 3" {
		t.Errorf("Expected Command 'shell input keyevent 3', got '%s'", decoded.ADB.Command)
	}
}

func TestWorkflowStepV2Serialization_SubWorkflow(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "run_workflow",
		Workflow: &SubWorkflowParams{
			WorkflowId: "wf-sub-123",
		},
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WorkflowStepV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Workflow == nil {
		t.Fatal("Workflow params should not be nil")
	}
	if decoded.Workflow.WorkflowId != "wf-sub-123" {
		t.Errorf("Expected WorkflowId 'wf-sub-123', got '%s'", decoded.Workflow.WorkflowId)
	}
}

// ==================== Validation Tests ====================

func TestWorkflowStepV2Validation_TapWithoutParams(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Tap:  nil,
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for tap without params")
	}
}

func TestWorkflowStepV2Validation_TapValid(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Tap:  &TapParams{X: 100, Y: 200},
	}

	err := step.Validate()
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

func TestWorkflowStepV2Validation_SwipeWithoutParams(t *testing.T) {
	step := WorkflowStepV2{
		ID:    "step1",
		Type:  "swipe",
		Swipe: nil,
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for swipe without params")
	}
}

func TestWorkflowStepV2Validation_SwipeWithoutCoordsOrDirection(t *testing.T) {
	step := WorkflowStepV2{
		ID:    "step1",
		Type:  "swipe",
		Swipe: &SwipeParams{},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for swipe without coordinates or direction")
	}
}

func TestWorkflowStepV2Validation_ElementWithoutSelector(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "click_element",
		Element: &ElementParamsV2{
			Selector: ElementSelector{},
			Action:   "click",
		},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for element without selector")
	}
}

func TestWorkflowStepV2Validation_InputTextWithoutInputText(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "input_text",
		Element: &ElementParamsV2{
			Selector: ElementSelector{
				Type:  "id",
				Value: "email",
			},
			Action:    "input",
			InputText: "",
		},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for input_text without inputText")
	}
}

func TestWorkflowStepV2Validation_AppWithoutPackageName(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "launch_app",
		App: &AppParams{
			PackageName: "",
			Action:      "launch",
		},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for app without packageName")
	}
}

func TestWorkflowStepV2Validation_BranchWithoutCondition(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "branch1",
		Type: "branch",
		Branch: &BranchParams{
			Condition: "",
		},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for branch without condition")
	}
}

func TestWorkflowStepV2Validation_BranchWithoutSelector(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "branch1",
		Type: "branch",
		Branch: &BranchParams{
			Condition: "exists",
			Selector:  nil,
		},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for element condition without selector")
	}
}

func TestWorkflowStepV2Validation_BranchVariableEqualsWithoutSelector(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "branch1",
		Type: "branch",
		Branch: &BranchParams{
			Condition:    "variable_equals",
			VariableName: "status",
			Selector:     nil, // No selector needed for variable_equals
		},
	}

	err := step.Validate()
	if err != nil {
		t.Errorf("Unexpected validation error for variable_equals without selector: %v", err)
	}
}

func TestWorkflowStepV2Validation_WaitNegativeDuration(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "wait",
		Wait: &WaitParams{
			DurationMs: -100,
		},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for negative duration")
	}
}

func TestWorkflowStepV2Validation_WaitZeroDuration(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "wait",
		Wait: &WaitParams{
			DurationMs: 0,
		},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for zero duration")
	}
}

func TestWorkflowStepV2Validation_ScriptWithoutName(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "script",
		Script: &ScriptParams{
			ScriptName: "",
		},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for script without name")
	}
}

func TestWorkflowStepV2Validation_VariableWithoutName(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "set_variable",
		Variable: &VariableParams{
			Name:  "",
			Value: "test",
		},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for variable without name")
	}
}

func TestWorkflowStepV2Validation_ADBWithoutCommand(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "adb",
		ADB: &ADBParams{
			Command: "",
		},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for ADB without command")
	}
}

func TestWorkflowStepV2Validation_SubWorkflowWithoutId(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "run_workflow",
		Workflow: &SubWorkflowParams{
			WorkflowId: "",
		},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for run_workflow without id")
	}
}

func TestWorkflowStepV2Validation_SystemKeys(t *testing.T) {
	keyTypes := []string{"key_back", "key_home", "key_recent", "key_power", "key_volume_up", "key_volume_down", "screen_on", "screen_off"}

	for _, keyType := range keyTypes {
		t.Run(keyType, func(t *testing.T) {
			step := WorkflowStepV2{
				ID:   "step1",
				Type: keyType,
			}

			err := step.Validate()
			if err != nil {
				t.Errorf("Unexpected validation error for %s: %v", keyType, err)
			}
		})
	}
}

func TestWorkflowStepV2Validation_UnknownType(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "unknown_type",
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for unknown type")
	}
}

func TestWorkflowStepV2Validation_MissingId(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "",
		Type: "tap",
		Tap:  &TapParams{X: 100, Y: 200},
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for missing id")
	}
}

func TestWorkflowStepV2Validation_MissingType(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "",
	}

	err := step.Validate()
	if err == nil {
		t.Error("Expected validation error for missing type")
	}
}

// ==================== Workflow Validation Tests ====================

func TestWorkflowV2Validation_Valid(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "Test Workflow",
		Steps: []WorkflowStepV2{
			{ID: "start", Type: "start", Connections: StepConnections{SuccessStepId: "step1"}},
			{ID: "step1", Type: "tap", Tap: &TapParams{X: 100, Y: 200}},
		},
	}

	errors := workflow.Validate()
	if len(errors) > 0 {
		t.Errorf("Unexpected validation errors: %v", errors)
	}
}

func TestWorkflowV2Validation_MissingId(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "",
		Name: "Test",
		Steps: []WorkflowStepV2{
			{ID: "start", Type: "start"},
		},
	}

	errors := workflow.Validate()
	if len(errors) == 0 {
		t.Error("Expected validation error for missing workflow id")
	}
}

func TestWorkflowV2Validation_MissingName(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "",
		Steps: []WorkflowStepV2{
			{ID: "start", Type: "start"},
		},
	}

	errors := workflow.Validate()
	if len(errors) == 0 {
		t.Error("Expected validation error for missing workflow name")
	}
}

func TestWorkflowV2Validation_NoStartNode(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "Test",
		Steps: []WorkflowStepV2{
			{ID: "step1", Type: "tap", Tap: &TapParams{X: 100, Y: 200}},
		},
	}

	errors := workflow.Validate()
	hasStartError := false
	for _, err := range errors {
		if ve, ok := err.(ValidationError); ok && ve.Message == "workflow must have a start node" {
			hasStartError = true
			break
		}
	}
	if !hasStartError {
		t.Error("Expected validation error for missing start node")
	}
}

func TestWorkflowV2Validation_MultipleStartNodes(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "Test",
		Steps: []WorkflowStepV2{
			{ID: "start1", Type: "start"},
			{ID: "start2", Type: "start"},
		},
	}

	errors := workflow.Validate()
	hasMultipleStartError := false
	for _, err := range errors {
		if ve, ok := err.(ValidationError); ok && ve.Message == "workflow can only have one start node" {
			hasMultipleStartError = true
			break
		}
	}
	if !hasMultipleStartError {
		t.Error("Expected validation error for multiple start nodes")
	}
}

func TestWorkflowV2Validation_InvalidConnection(t *testing.T) {
	workflow := WorkflowV2{
		ID:   "wf1",
		Name: "Test",
		Steps: []WorkflowStepV2{
			{ID: "start", Type: "start", Connections: StepConnections{SuccessStepId: "nonexistent"}},
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

// ==================== Helper Method Tests ====================

func TestGetNextStepId_NormalNode_Success(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "error_handler",
		},
	}

	next := step.GetNextStepId(true, false)
	if next != "step2" {
		t.Errorf("Expected 'step2', got '%s'", next)
	}
}

func TestGetNextStepId_NormalNode_Error(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "error_handler",
		},
	}

	next := step.GetNextStepId(false, false)
	if next != "error_handler" {
		t.Errorf("Expected 'error_handler', got '%s'", next)
	}
}

func TestGetNextStepId_Branch_True(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "branch1",
		Type: "branch",
		Connections: StepConnections{
			TrueStepId:  "step_yes",
			FalseStepId: "step_no",
		},
	}

	next := step.GetNextStepId(true, true)
	if next != "step_yes" {
		t.Errorf("Expected 'step_yes', got '%s'", next)
	}
}

func TestGetNextStepId_Branch_False(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "branch1",
		Type: "branch",
		Connections: StepConnections{
			TrueStepId:  "step_yes",
			FalseStepId: "step_no",
		},
	}

	next := step.GetNextStepId(false, true)
	if next != "step_no" {
		t.Errorf("Expected 'step_no', got '%s'", next)
	}
}

func TestShouldStopOnError_ErrorPathConnected(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			ErrorStepId: "error_handler",
		},
		Common: StepCommon{OnError: "stop"},
	}

	if step.ShouldStopOnError() {
		t.Error("Should not stop when error path is connected")
	}
}

func TestShouldStopOnError_NoErrorPath_Stop(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			ErrorStepId: "",
		},
		Common: StepCommon{OnError: "stop"},
	}

	if !step.ShouldStopOnError() {
		t.Error("Should stop when error path not connected and onError=stop")
	}
}

func TestShouldStopOnError_NoErrorPath_Continue(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			ErrorStepId: "",
		},
		Common: StepCommon{OnError: "continue"},
	}

	if step.ShouldStopOnError() {
		t.Error("Should not stop when onError=continue")
	}
}

func TestGetFallbackStepId_ErrorPathConnected(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "error_handler",
		},
	}

	fallback := step.GetFallbackStepId()
	if fallback != "error_handler" {
		t.Errorf("Expected 'error_handler', got '%s'", fallback)
	}
}

func TestGetFallbackStepId_NoErrorPath_Continue(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "",
		},
		Common: StepCommon{OnError: "continue"},
	}

	fallback := step.GetFallbackStepId()
	if fallback != "step2" {
		t.Errorf("Expected 'step2' (success path as fallback), got '%s'", fallback)
	}
}

func TestGetFallbackStepId_NoErrorPath_Stop(t *testing.T) {
	step := WorkflowStepV2{
		ID:   "step1",
		Type: "tap",
		Connections: StepConnections{
			SuccessStepId: "step2",
			ErrorStepId:   "",
		},
		Common: StepCommon{OnError: "stop"},
	}

	fallback := step.GetFallbackStepId()
	if fallback != "" {
		t.Errorf("Expected empty string (stop), got '%s'", fallback)
	}
}

// ==================== Full Workflow Serialization ====================

func TestWorkflowV2Serialization(t *testing.T) {
	workflow := WorkflowV2{
		ID:          "wf1",
		Name:        "Login Test",
		Description: "Test login flow",
		Version:     WorkflowSchemaVersion,
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
					ErrorStepId:   "error_handler",
				},
			},
			{
				ID:   "step2",
				Type: "input_text",
				Element: &ElementParamsV2{
					Selector:  ElementSelector{Type: "id", Value: "email"},
					Action:    "input",
					InputText: "test@example.com",
				},
			},
			{
				ID:   "error_handler",
				Type: "tap",
				Tap:  &TapParams{X: 50, Y: 50},
			},
		},
		Variables: map[string]string{
			"email": "test@example.com",
		},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(workflow)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded WorkflowV2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ID != "wf1" {
		t.Errorf("Expected ID 'wf1', got '%s'", decoded.ID)
	}
	if decoded.Version != WorkflowSchemaVersion {
		t.Errorf("Expected Version %d, got %d", WorkflowSchemaVersion, decoded.Version)
	}
	if len(decoded.Steps) != 4 {
		t.Errorf("Expected 4 steps, got %d", len(decoded.Steps))
	}
	if decoded.Variables["email"] != "test@example.com" {
		t.Errorf("Expected variable email='test@example.com', got '%s'", decoded.Variables["email"])
	}
}
