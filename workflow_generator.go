package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ========================================
// Workflow Generator - 从 Session 录制生成 Workflow
// ========================================

// WorkflowGeneratorConfig configures workflow generation
type WorkflowGeneratorConfig struct {
	MinTouchDuration    int64  // Minimum touch duration to consider (ms)
	MaxImplicitWait     int64  // Maximum implicit wait time (ms)
	SelectorPreference  string // "id", "text", "xpath", "coordinates"
	IncludeAssertions   bool   // Generate verification steps
	DetectLoops         bool   // Detect and generate loops
	SimplifyCoordinates bool   // Use element centers instead of exact coordinates
	UseVideoAnalysis    bool   // Use video frame analysis for step context
	VideoAnalysisWidth  int    // Width for video frame extraction (default 720)
}

// DefaultWorkflowGeneratorConfig returns default config
func DefaultWorkflowGeneratorConfig() WorkflowGeneratorConfig {
	return WorkflowGeneratorConfig{
		MinTouchDuration:    50,
		MaxImplicitWait:     5000,
		SelectorPreference:  "id",
		IncludeAssertions:   true,
		DetectLoops:         true,
		SimplifyCoordinates: true,
		UseVideoAnalysis:    true,
		VideoAnalysisWidth:  720,
	}
}

// GeneratedWorkflow represents a generated workflow
type GeneratedWorkflow struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []WorkflowStep `json:"steps"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Suggestions []string       `json:"suggestions,omitempty"`
	Confidence  float64        `json:"confidence,omitempty"`
	Loops       []LoopPattern  `json:"loops,omitempty"`
	Branches    []any          `json:"branches,omitempty"`
}

// LoopPattern represents a detected loop in the workflow
type LoopPattern struct {
	StartIndex int     `json:"startIndex"`
	EndIndex   int     `json:"endIndex"`
	Iterations int     `json:"iterations"`
	Confidence float64 `json:"confidence"`
}

// TouchEventForGeneration represents a touch event with context
type TouchEventForGeneration struct {
	Event         *UnifiedEvent          `json:"event"`
	UIBefore      map[string]interface{} `json:"uiBefore,omitempty"`
	UIAfter       map[string]interface{} `json:"uiAfter,omitempty"`
	TargetElement map[string]interface{} `json:"targetElement,omitempty"`
}

// StepVideoContext contains video frame analysis context for a step
type StepVideoContext struct {
	FrameBase64  string   `json:"frameBase64,omitempty"`  // Base64 encoded frame image
	SceneType    string   `json:"sceneType,omitempty"`    // Detected scene type
	VisibleText  []string `json:"visibleText,omitempty"`  // OCR text on screen
	Description  string   `json:"description,omitempty"`  // AI description of the screen
	UIElements   []string `json:"uiElements,omitempty"`   // Detected UI elements summary
}

// GenerateWorkflowFromSession generates a workflow from session events
func (a *App) GenerateWorkflowFromSession(sessionID string, config *WorkflowGeneratorConfig) (*GeneratedWorkflow, error) {
	if a.eventStore == nil {
		return nil, fmt.Errorf("event store not initialized")
	}

	if config == nil {
		defaultConfig := DefaultWorkflowGeneratorConfig()
		config = &defaultConfig
	}

	LogInfo("workflow_generator").
		Str("sessionId", sessionID).
		Bool("useVideoAnalysis", config.UseVideoAnalysis).
		Int("videoAnalysisWidth", config.VideoAnalysisWidth).
		Msg("Starting workflow generation")

	// Send initial progress
	a.emitWorkflowGenProgress("loading_session", 5, "Loading session data...")

	// Get session info
	session, err := a.eventStore.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	LogInfo("workflow_generator").
		Str("sessionId", sessionID).
		Str("videoPath", session.VideoPath).
		Int64("videoDuration", session.VideoDuration).
		Int64("videoOffset", session.VideoOffset).
		Msg("Session loaded")

	a.emitWorkflowGenProgress("querying_events", 10, "Querying touch events...")

	// Query touch events with full data (needed for X, Y coordinates)
	touchResult, err := a.eventStore.QueryEvents(EventQuery{
		SessionID:   sessionID,
		Sources:     []EventSource{SourceTouch},
		Limit:       1000,
		IncludeData: true, // Need full data for touch coordinates
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query touch events: %w", err)
	}
	touchEvents := touchResult.Events

	LogInfo("workflow_generator").
		Int("touchEventCount", len(touchEvents)).
		Str("sessionId", sessionID).
		Msg("Touch events queried")

	a.emitWorkflowGenProgress("querying_ui", 15, "Querying UI events...")

	// Query UI events (for context) with full data
	uiResult, err := a.eventStore.QueryEvents(EventQuery{
		SessionID:   sessionID,
		Sources:     []EventSource{SourceUI},
		Limit:       500,
		IncludeData: true, // Need full data for UI hierarchy
	})
	var uiEvents []UnifiedEvent
	if err != nil {
		// Continue without UI events
		uiEvents = []UnifiedEvent{}
	} else {
		uiEvents = uiResult.Events
	}

	// Check if video analysis is available and enabled
	var videoContexts map[int]*StepVideoContext
	if config.UseVideoAnalysis && session.VideoPath != "" {
		a.emitWorkflowGenProgress("video_analysis", 20, fmt.Sprintf("Extracting video frames for %d events (video: %s)...", len(touchEvents), session.VideoPath))
		videoContexts = a.extractVideoContextsForEvents(session, touchEvents, config)
		if videoContexts != nil && len(videoContexts) > 0 {
			a.emitWorkflowGenProgress("video_success", 42, fmt.Sprintf("Video analysis complete: %d frames extracted", len(videoContexts)))
		} else {
			a.emitWorkflowGenProgress("video_failed", 42, "Video analysis failed or returned no frames")
		}
	} else if !config.UseVideoAnalysis {
		a.emitWorkflowGenProgress("skip_video", 40, "Video analysis disabled in config")
	} else {
		a.emitWorkflowGenProgress("skip_video", 40, fmt.Sprintf("No video available for session (VideoPath: '%s')", session.VideoPath))
	}

	a.emitWorkflowGenProgress("generating_steps", 50, fmt.Sprintf("Generating workflow from %d touch events...", len(touchEvents)))

	// Generate workflow
	hasVideoCtx := videoContexts != nil && len(videoContexts) > 0
	workflow := &GeneratedWorkflow{
		Name:        fmt.Sprintf("Workflow from %s", session.Name),
		Description: fmt.Sprintf("Auto-generated from session %s", sessionID),
		Steps:       []WorkflowStep{},
		Metadata: map[string]any{
			"sourceSessionId":   sessionID,
			"generatedAt":       time.Now().Unix(),
			"hasVideoContext":   hasVideoCtx,
			"videoFrameCount":   len(videoContexts),
			"sessionVideoPath":  session.VideoPath,
			"useVideoAnalysis":  config.UseVideoAnalysis,
		},
		Suggestions: []string{},
		Confidence:  0.8,
		Loops:       []LoopPattern{},
		Branches:    []any{},
	}

	LogInfo("workflow_generator").
		Bool("hasVideoContext", hasVideoCtx).
		Int("videoFrameCount", len(videoContexts)).
		Str("sessionVideoPath", session.VideoPath).
		Msg("Metadata set")

	// Convert touch events to workflow steps
	var lastEventTime int64 = 0
	for i, event := range touchEvents {
		step, warnings := a.touchEventToStep(&event, i, lastEventTime, config, uiEvents)
		if step != nil {
			workflow.Steps = append(workflow.Steps, *step)
		}
		workflow.Suggestions = append(workflow.Suggestions, warnings...)
		lastEventTime = event.Timestamp
	}

	// Add assertions if enabled
	if config.IncludeAssertions && len(workflow.Steps) > 0 {
		a.emitWorkflowGenProgress("adding_assertions", 60, "Adding assertions...")
		a.addAutoAssertions(workflow, touchEvents, uiEvents)
	}

	// Try AI enhancement if available
	a.aiServiceMu.RLock()
	aiService := a.aiService
	aiConfig := a.aiConfigMgr.GetConfig()
	a.aiServiceMu.RUnlock()

	LogInfo("workflow_generator").
		Bool("aiServiceNil", aiService == nil).
		Bool("aiServiceReady", aiService != nil && aiService.IsReady()).
		Bool("workflowGenEnabled", aiConfig.Features.WorkflowGeneration).
		Msg("Checking AI enhancement conditions")

	if aiService != nil && aiService.IsReady() && aiConfig.Features.WorkflowGeneration {
		a.emitWorkflowGenProgress("ai_enhancement", 70, "AI analyzing workflow and naming steps...")
		enhanced, err := a.enhanceWorkflowWithAI(workflow, videoContexts)
		if err != nil {
			LogError("workflow_generator").Err(err).Msg("AI enhancement failed")
			a.emitWorkflowGenProgress("ai_error", 75, fmt.Sprintf("AI enhancement failed: %v", err))
		} else if enhanced != nil {
			workflow = enhanced
			LogInfo("workflow_generator").Msg("AI enhancement completed")
		}
	} else {
		LogWarn("workflow_generator").
			Bool("aiServiceNil", aiService == nil).
			Bool("aiServiceReady", aiService != nil && aiService.IsReady()).
			Bool("workflowGenEnabled", aiConfig.Features.WorkflowGeneration).
			Msg("AI enhancement skipped")
	}

	a.emitWorkflowGenProgress("completed", 100, "Workflow generation completed!")
	return workflow, nil
}

// emitWorkflowGenProgress sends workflow generation progress to frontend
func (a *App) emitWorkflowGenProgress(stage string, percent int, message string) {
	a.emitWorkflowGenProgressWithData(stage, percent, message, nil)
}

// emitWorkflowGenProgressWithData sends workflow generation progress with additional data
func (a *App) emitWorkflowGenProgressWithData(stage string, percent int, message string, extraData map[string]interface{}) {
	if a.ctx == nil {
		return
	}
	data := map[string]interface{}{
		"stage":   stage,
		"percent": percent,
		"message": message,
	}
	// Merge extra data
	for k, v := range extraData {
		data[k] = v
	}
	wailsRuntime.EventsEmit(a.ctx, "workflow-gen-progress", data)
}

// extractVideoContextsForEvents extracts video frames at each touch event timestamp
func (a *App) extractVideoContextsForEvents(session *DeviceSession, touchEvents []UnifiedEvent, config *WorkflowGeneratorConfig) map[int]*StepVideoContext {
	videoService := a.getVideoService()
	if videoService == nil || !videoService.IsAvailable() {
		LogWarn("workflow_generator").Msg("Video service not available for frame extraction")
		a.emitWorkflowGenProgress("video_error", 25, "Video service not available (ffmpeg not found?)")
		return nil
	}

	// Check video file exists
	if _, err := os.Stat(session.VideoPath); os.IsNotExist(err) {
		LogWarn("workflow_generator").Str("path", session.VideoPath).Msg("Video file not found")
		a.emitWorkflowGenProgress("video_error", 25, "Video file not found")
		return nil
	}

	contexts := make(map[int]*StepVideoContext)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var completed int32

	totalEvents := len(touchEvents)
	a.emitWorkflowGenProgress("video_extraction", 22, fmt.Sprintf("Extracting %d video frames...", totalEvents))

	// Get AI service for frame analysis
	a.aiServiceMu.RLock()
	aiService := a.aiService
	a.aiServiceMu.RUnlock()

	LogInfo("workflow_generator").
		Bool("aiServiceNil", aiService == nil).
		Bool("aiServiceReady", aiService != nil && aiService.IsReady()).
		Msg("AI service status for frame analysis")

	// Extract frames in parallel (limit concurrency)
	sem := make(chan struct{}, 4) // Max 4 concurrent extractions

	for i, event := range touchEvents {
		wg.Add(1)
		go func(idx int, evt UnifiedEvent) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			// Calculate video time from event timestamp
			// RelativeTime is the time since session start in milliseconds
			videoTimeMs := evt.RelativeTime
			if videoTimeMs < 0 {
				videoTimeMs = 0
			}

			// Add video offset if exists
			if session.VideoOffset > 0 {
				videoTimeMs += session.VideoOffset
			}

			width := config.VideoAnalysisWidth
			if width <= 0 {
				width = 720
			}

			// Extract frame
			frameBase64, err := videoService.ExtractFrameBase64(session.VideoPath, videoTimeMs, width)
			if err != nil {
				LogDebug("workflow_generator").
					Int("step", idx).
					Int64("timeMs", videoTimeMs).
					Err(err).
					Msg("Failed to extract frame")
				atomic.AddInt32(&completed, 1)
				return
			}

			ctx := &StepVideoContext{
				FrameBase64: frameBase64,
			}

			// Update progress counter first
			done := atomic.AddInt32(&completed, 1)
			progress := 22 + int(float64(done)/float64(totalEvents)*18) // Progress from 22% to 40%

			// If AI service is available, analyze the frame
			LogDebug("workflow_generator").
				Int("step", idx).
				Bool("aiServiceNil", aiService == nil).
				Bool("aiServiceReady", aiService != nil && aiService.IsReady()).
				Msg("Checking AI for frame analysis")

			if aiService != nil && aiService.IsReady() {
				analyzer := NewVideoAnalyzer(videoService, aiService, a.dataDir)
				frameAnalysis, err := analyzer.AnalyzeFrame(context.Background(), session.VideoPath, videoTimeMs)
				if err != nil {
					LogWarn("workflow_generator").
						Int("step", idx).
						Int64("timeMs", videoTimeMs).
						Err(err).
						Msg("Frame AI analysis failed")
					a.emitWorkflowGenProgress("frame_analysis_error", progress,
						fmt.Sprintf("Frame %d AI analysis failed: %v", idx, err))
				} else if frameAnalysis != nil {
					ctx.SceneType = frameAnalysis.SceneType
					ctx.VisibleText = frameAnalysis.OCRText
					ctx.Description = frameAnalysis.Description

					// Summarize UI elements
					var elementSummary []string
					for _, el := range frameAnalysis.UIElements {
						if el.Text != "" {
							elementSummary = append(elementSummary, fmt.Sprintf("%s:%s", el.Type, el.Text))
						} else {
							elementSummary = append(elementSummary, el.Type)
						}
					}
					ctx.UIElements = elementSummary

					// Log frame analysis result
					LogInfo("workflow_generator").
						Int("step", idx).
						Int64("timeMs", videoTimeMs).
						Str("sceneType", frameAnalysis.SceneType).
						Str("description", frameAnalysis.Description).
						Int("uiElementsCount", len(frameAnalysis.UIElements)).
						Msg("Frame analyzed")

					a.emitWorkflowGenProgressWithData("frame_analyzed", progress,
						fmt.Sprintf("Frame %d @ %dms: Scene=%s, Desc=%s, Elements=%d",
							idx, videoTimeMs, frameAnalysis.SceneType, frameAnalysis.Description, len(frameAnalysis.UIElements)),
						map[string]interface{}{
							"frameIndex":   idx,
							"frameTimeMs":  videoTimeMs,
							"frameBase64":  frameBase64,
							"sceneType":    frameAnalysis.SceneType,
							"description":  frameAnalysis.Description,
							"ocrText":      frameAnalysis.OCRText,
							"uiElements":   len(frameAnalysis.UIElements),
						})
				}
			} else {
				LogDebug("workflow_generator").
					Int("step", idx).
					Msg("AI service not available for frame analysis")
			}

			mu.Lock()
			contexts[idx] = ctx
			mu.Unlock()
			a.emitWorkflowGenProgress("video_extraction", progress,
				fmt.Sprintf("Analyzed frame %d/%d...", done, totalEvents))
		}(i, event)
	}

	wg.Wait()

	a.emitWorkflowGenProgress("video_done", 40,
		fmt.Sprintf("Video analysis complete: %d frames extracted", len(contexts)))

	LogInfo("workflow_generator").
		Int("totalEvents", len(touchEvents)).
		Int("extractedFrames", len(contexts)).
		Msg("Video frame extraction completed")

	return contexts
}

// touchEventToStep converts a touch event to a workflow step
func (a *App) touchEventToStep(event *UnifiedEvent, index int, lastEventTime int64, config *WorkflowGeneratorConfig, uiEvents []UnifiedEvent) (*WorkflowStep, []string) {
	var warnings []string

	// Parse touch data
	var touchData TouchEventData
	if event.Data != nil {
		if err := json.Unmarshal(event.Data, &touchData); err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to parse touch event data: %v", err))
			return nil, warnings
		}
		LogDebug("workflow_generator").
			Int("step", index).
			Float64("x", touchData.X).
			Float64("y", touchData.Y).
			Str("action", touchData.Action).
			Str("gestureType", touchData.GestureType).
			Str("rawData", string(event.Data)).
			Msg("Parsed touch data")
	} else {
		LogWarn("workflow_generator").Int("step", index).Msg("Touch event has nil Data")
	}

	// Skip very short touches (might be accidental)
	// Use event.Duration since TouchEventData doesn't have Duration field
	if event.Duration < config.MinTouchDuration {
		return nil, warnings
	}

	// Create base step - Type is a string like "click", "swipe", "long_press"
	stepType := "click"
	stepName := fmt.Sprintf("Step %d", index+1)

	// Determine action type based on gesture
	if touchData.GestureType == "swipe" {
		stepType = "swipe"
	} else if event.Duration > 500 {
		stepType = "long_press"
	}

	step := &WorkflowStep{
		ID:   fmt.Sprintf("step_%d", index+1),
		Type: stepType,
		Name: stepName,
	}

	// Always store coordinates in Value field
	step.Value = fmt.Sprintf("%.0f,%.0f", touchData.X, touchData.Y)

	// Try to find matching UI element
	element := a.findUIElementAtPoint(int(touchData.X), int(touchData.Y), event.Timestamp, uiEvents)
	if element != nil {
		selector := a.buildSelectorFromElement(element, config.SelectorPreference)
		if selector != nil {
			step.Selector = selector
			// Update step name with selector info
			step.Name = fmt.Sprintf("Tap on %s", selector.Value)
		}
	}

	// Add warning if no selector found
	if step.Selector == nil {
		warnings = append(warnings, fmt.Sprintf("Step %d: No UI element found, using coordinates", index+1))
	}

	// Add implicit wait if there's a significant gap
	if lastEventTime > 0 {
		gap := event.Timestamp - lastEventTime
		if gap > 500 && gap < config.MaxImplicitWait {
			step.PreWait = int(gap)
		}
	}

	return step, warnings
}

// findUIElementAtPoint finds a UI element at the given coordinates
func (a *App) findUIElementAtPoint(x, y int, timestamp int64, uiEvents []UnifiedEvent) map[string]interface{} {
	// Find the closest UI snapshot before this timestamp
	var closestEvent *UnifiedEvent
	for i := range uiEvents {
		if uiEvents[i].Timestamp <= timestamp {
			if closestEvent == nil || uiEvents[i].Timestamp > closestEvent.Timestamp {
				closestEvent = &uiEvents[i]
			}
		}
	}

	if closestEvent == nil || closestEvent.Data == nil {
		return nil
	}

	// Parse UI hierarchy and find element at point
	var uiData map[string]interface{}
	if err := json.Unmarshal(closestEvent.Data, &uiData); err != nil {
		return nil
	}

	// Look for element at coordinates (simplified - real implementation would traverse hierarchy)
	if elements, ok := uiData["elements"].([]interface{}); ok {
		for _, el := range elements {
			if elem, ok := el.(map[string]interface{}); ok {
				if bounds, ok := elem["bounds"].(map[string]interface{}); ok {
					left, _ := bounds["left"].(float64)
					top, _ := bounds["top"].(float64)
					right, _ := bounds["right"].(float64)
					bottom, _ := bounds["bottom"].(float64)

					if float64(x) >= left && float64(x) <= right &&
						float64(y) >= top && float64(y) <= bottom {
						return elem
					}
				}
			}
		}
	}

	return nil
}

// buildSelectorFromElement builds a selector from a UI element
// ElementSelector has Type ("text", "id", "xpath", "advanced") and Value fields
func (a *App) buildSelectorFromElement(element map[string]interface{}, preference string) *ElementSelector {
	// Extract element properties
	resourceID, _ := element["resourceId"].(string)
	text, _ := element["text"].(string)
	contentDesc, _ := element["contentDescription"].(string)
	className, _ := element["className"].(string)

	// Build selector based on preference
	switch preference {
	case "id":
		if resourceID != "" {
			return &ElementSelector{
				Type:  "id",
				Value: resourceID,
			}
		}
	case "text":
		if text != "" {
			return &ElementSelector{
				Type:  "text",
				Value: text,
			}
		}
	}

	// Fallback: try all available attributes in priority order
	if resourceID != "" {
		return &ElementSelector{
			Type:  "id",
			Value: resourceID,
		}
	}
	if text != "" {
		return &ElementSelector{
			Type:  "text",
			Value: text,
		}
	}
	if contentDesc != "" {
		return &ElementSelector{
			Type:  "text",
			Value: contentDesc,
		}
	}
	if className != "" {
		return &ElementSelector{
			Type:  "xpath",
			Value: fmt.Sprintf("//%s", className),
		}
	}

	return nil
}

// addAutoAssertions adds automatic assertion steps to the workflow
func (a *App) addAutoAssertions(workflow *GeneratedWorkflow, touchEvents []UnifiedEvent, uiEvents []UnifiedEvent) {
	// Add assertion after each significant interaction
	// This is a simplified implementation

	newSteps := make([]WorkflowStep, 0, len(workflow.Steps)*2)

	for _, step := range workflow.Steps {
		newSteps = append(newSteps, step)

		// Add wait_element assertion after clicks that have a selector
		if step.Type == "click" && step.Selector != nil {
			assertStep := WorkflowStep{
				ID:       fmt.Sprintf("%s_assert", step.ID),
				Type:     "wait",
				Name:     fmt.Sprintf("Verify %s", step.Name),
				Selector: step.Selector,
				Timeout:  5000,
			}
			newSteps = append(newSteps, assertStep)
		}
	}

	workflow.Steps = newSteps
}

// enhanceWorkflowWithAI uses AI to improve the generated workflow
func (a *App) enhanceWorkflowWithAI(workflow *GeneratedWorkflow, videoContexts map[int]*StepVideoContext) (*GeneratedWorkflow, error) {
	aiService := a.aiService
	if aiService == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	// Build workflow summary for AI with step details
	type stepInfo struct {
		ID           string   `json:"id"`
		Type         string   `json:"type"`
		Name         string   `json:"name"`
		Selector     string   `json:"selector,omitempty"`
		Value        string   `json:"value,omitempty"`
		SceneType    string   `json:"sceneType,omitempty"`
		VisibleText  []string `json:"visibleText,omitempty"`
		ScreenDesc   string   `json:"screenDescription,omitempty"`
		UIElements   []string `json:"uiElements,omitempty"`
	}

	stepsInfo := make([]stepInfo, len(workflow.Steps))
	for i, step := range workflow.Steps {
		selectorStr := ""
		if step.Selector != nil {
			selectorStr = fmt.Sprintf("%s=%s", step.Selector.Type, step.Selector.Value)
		}
		info := stepInfo{
			ID:       step.ID,
			Type:     step.Type,
			Name:     step.Name,
			Selector: selectorStr,
			Value:    step.Value,
		}

		// Add video context if available
		if videoContexts != nil {
			if ctx, ok := videoContexts[i]; ok && ctx != nil {
				info.SceneType = ctx.SceneType
				info.VisibleText = ctx.VisibleText
				info.ScreenDesc = ctx.Description
				info.UIElements = ctx.UIElements
			}
		}

		stepsInfo[i] = info
	}

	stepsJSON, _ := json.Marshal(stepsInfo)

	// Build prompt based on whether we have video context
	hasVideoContext := videoContexts != nil && len(videoContexts) > 0

	var prompt string
	if hasVideoContext {
		prompt = fmt.Sprintf(`You are a mobile test automation expert. Analyze this auto-generated workflow with video frame context and provide meaningful names.

Current workflow name: %s
Steps (with screen context from video frames):
%s

The steps include video analysis context:
- sceneType: The detected type of screen (login, home, settings, list, dialog, etc.)
- visibleText: OCR-detected text visible on screen
- screenDescription: AI description of what's shown on screen
- uiElements: Detected UI elements on screen

Please provide:
1. A clear, descriptive workflow name that describes the overall user journey based on the screens shown (e.g., "User Login Flow", "Add Item to Cart", "Profile Settings Update")
2. A brief description of what this workflow does based on the screen transitions
3. Meaningful names for each step that describe the user action and screen context (e.g., "Tap Login Button on Login Screen", "Enter Username in Form", "Navigate to Settings from Home")

Return ONLY valid JSON in this exact format:
{
  "name": "descriptive workflow name",
  "description": "brief description of the workflow",
  "steps": [
    {"id": "step_1", "name": "descriptive step name"},
    {"id": "step_2", "name": "descriptive step name"}
  ],
  "suggestions": ["optional improvement suggestion"]
}`, workflow.Name, string(stepsJSON))
	} else {
		prompt = fmt.Sprintf(`You are a mobile test automation expert. Analyze this auto-generated workflow and provide meaningful names.

Current workflow name: %s
Steps:
%s

Please provide:
1. A clear, descriptive workflow name that describes the overall user journey (e.g., "User Login Flow", "Add Item to Cart", "Profile Settings Update")
2. A brief description of what this workflow does
3. Meaningful names for each step that describe the user action (e.g., "Tap Login Button", "Enter Username", "Scroll to Settings")

Return ONLY valid JSON in this exact format:
{
  "name": "descriptive workflow name",
  "description": "brief description of the workflow",
  "steps": [
    {"id": "step_1", "name": "descriptive step name"},
    {"id": "step_2", "name": "descriptive step name"}
  ],
  "suggestions": ["optional improvement suggestion"]
}`, workflow.Name, string(stepsJSON))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Log and emit AI prompt for debugging
	LogInfo("workflow_generator").
		Int("promptLength", len(prompt)).
		Bool("hasVideoContext", hasVideoContext).
		Msg("Sending prompt to AI")

	a.emitWorkflowGenProgress("ai_prompt", 72, fmt.Sprintf("AI Prompt (%d chars):\n%s", len(prompt), prompt))

	resp, err := aiService.Complete(ctx, &CompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a mobile test automation expert. Generate clear, descriptive names for workflow steps based on their actions and selectors. Return only valid JSON without markdown formatting."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1000,
		Temperature: 0.3,
	})

	if err != nil {
		LogError("workflow_generator").Err(err).Msg("AI completion failed")
		a.emitWorkflowGenProgress("ai_error", 75, fmt.Sprintf("AI Error: %v", err))
		return nil, err
	}

	// Log and emit AI response for debugging
	LogInfo("workflow_generator").
		Int("responseLength", len(resp.Content)).
		Msg("Received AI response")

	a.emitWorkflowGenProgress("ai_response", 78, fmt.Sprintf("AI Response:\n%s", resp.Content))

	// Parse AI response
	var aiResponse struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Steps       []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"steps"`
		Suggestions []string `json:"suggestions"`
	}

	response := strings.TrimSpace(resp.Content)
	// Remove markdown code blocks if present
	if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	if err := json.Unmarshal([]byte(response), &aiResponse); err != nil {
		// Try to extract JSON from the response
		startIdx := strings.Index(response, "{")
		endIdx := strings.LastIndex(response, "}")
		if startIdx >= 0 && endIdx > startIdx {
			response = response[startIdx : endIdx+1]
			if err := json.Unmarshal([]byte(response), &aiResponse); err != nil {
				return nil, fmt.Errorf("failed to parse AI response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse AI response: %w", err)
		}
	}

	// Apply AI-generated workflow name and description
	if aiResponse.Name != "" {
		workflow.Name = aiResponse.Name
	}
	if aiResponse.Description != "" {
		workflow.Description = aiResponse.Description
	}

	// Apply AI-generated step names
	stepNameMap := make(map[string]string)
	for _, s := range aiResponse.Steps {
		if s.ID != "" && s.Name != "" {
			stepNameMap[s.ID] = s.Name
		}
	}

	for i := range workflow.Steps {
		if newName, ok := stepNameMap[workflow.Steps[i].ID]; ok {
			workflow.Steps[i].Name = newName
		}
	}

	// Add suggestions as warnings
	if len(aiResponse.Suggestions) > 0 {
		workflow.Suggestions = append(workflow.Suggestions, aiResponse.Suggestions...)
	}

	workflow.Metadata["aiEnhanced"] = true
	workflow.Metadata["aiModel"] = aiService.GetProvider().Model()
	if videoContexts != nil && len(videoContexts) > 0 {
		workflow.Metadata["videoFramesAnalyzed"] = len(videoContexts)
	}

	return workflow, nil
}
