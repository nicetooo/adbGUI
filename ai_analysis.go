package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ========================================
// AI Analysis Service - AI 分析服务
// 包含自然语言搜索、日志分析、崩溃分析等功能
// ========================================

// ========================================
// Natural Language Search - 自然语言搜索
// ========================================

// NLQueryResult represents the result of parsing a natural language query
type NLQueryResult struct {
	Query       NLParsedQuery `json:"query"`       // Structured query
	Explanation string        `json:"explanation"` // Explanation of the parsing
	Confidence  float32       `json:"confidence"`  // Confidence score
	Suggestions []string      `json:"suggestions"` // Alternative query suggestions
}

// NLParsedQuery represents a parsed natural language query (separate from EventQuery in event_store.go)
type NLParsedQuery struct {
	Types      []string        `json:"types,omitempty"`      // Event types to filter
	Sources    []string        `json:"sources,omitempty"`    // Event sources
	Levels     []string        `json:"levels,omitempty"`     // Event levels
	Keywords   []string        `json:"keywords,omitempty"`   // Keywords to search
	TimeRange  *NLTimeRange    `json:"timeRange,omitempty"`  // Time range filter
	Context    string          `json:"context,omitempty"`    // Contextual filter (e.g., "after login")
}

// NLTimeRange represents a time range for natural language queries
type NLTimeRange struct {
	StartMs int64  `json:"startMs,omitempty"`
	EndMs   int64  `json:"endMs,omitempty"`
	Last    string `json:"last,omitempty"` // e.g., "5m", "1h", "1d"
}

// ParseNaturalQuery parses a natural language query into a structured query
func (a *App) ParseNaturalQuery(query string, sessionID string) (*NLQueryResult, error) {
	a.aiServiceMu.RLock()
	aiService := a.aiService
	config := a.aiConfigMgr.GetConfig()
	a.aiServiceMu.RUnlock()

	// Check if AI is enabled and ready
	if aiService == nil || !aiService.IsReady() || !config.Features.NaturalSearch {
		// Fall back to simple keyword parsing
		return parseQuerySimple(query), nil
	}

	// Build prompt for LLM
	prompt := buildNLQueryPrompt(query)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := aiService.Complete(ctx, &CompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: nlQuerySystemPrompt},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   500,
		Temperature: 0.1, // Low temperature for more consistent parsing
	})

	if err != nil {
		// Fall back to simple parsing on error
		LogWarn("ai_analysis").Err(err).Msg("LLM query parsing failed, using simple parser")
		return parseQuerySimple(query), nil
	}

	// Parse LLM response
	result, err := parseNLQueryResponse(resp.Content)
	if err != nil {
		return parseQuerySimple(query), nil
	}

	return result, nil
}

const nlQuerySystemPrompt = `You are a query parser for an Android device monitoring system. Parse natural language queries into structured JSON queries.

Available event types:
- logcat, logcat_aggregated: Device logs
- http_request, http_response, websocket_message: Network events
- app_start, app_stop, app_crash, app_anr: App lifecycle
- touch, gesture: Touch events
- activity_start, activity_stop: Activity lifecycle
- battery_change, network_change, screen_change: Device state
- workflow_start, workflow_step, workflow_complete, workflow_error: Automation

Available sources: logcat, network, device, app, touch, workflow, ui, perf, system, assertion
Available levels: verbose, debug, info, warn, error, fatal

Output JSON format:
{
  "query": {
    "types": ["event_type1", "event_type2"],
    "sources": ["source1"],
    "levels": ["level1"],
    "keywords": ["keyword1", "keyword2"],
    "timeRange": {"last": "5m"},
    "context": "contextual description"
  },
  "explanation": "Brief explanation of the parsed query",
  "confidence": 0.95
}`

func buildNLQueryPrompt(query string) string {
	return fmt.Sprintf(`Parse this query: "%s"

Return only valid JSON, no markdown or explanation outside the JSON.`, query)
}

func parseNLQueryResponse(response string) (*NLQueryResult, error) {
	// Clean response (remove markdown if present)
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	var result NLQueryResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &result, nil
}

// parseQuerySimple provides a simple keyword-based query parser
func parseQuerySimple(query string) *NLQueryResult {
	query = strings.ToLower(query)
	result := &NLQueryResult{
		Query:      NLParsedQuery{},
		Confidence: 0.5,
	}

	// Detect event types
	typeKeywords := map[string][]string{
		"app_crash":     {"crash", "崩溃", "闪退"},
		"app_anr":       {"anr", "无响应", "not responding"},
		"http_request":  {"http", "network", "request", "api", "网络", "请求"},
		"logcat":        {"log", "日志"},
		"touch":         {"touch", "tap", "click", "点击", "触摸"},
		"app_start":     {"start", "launch", "启动"},
		"app_stop":      {"stop", "close", "关闭"},
		"battery_change": {"battery", "电池"},
	}

	for eventType, keywords := range typeKeywords {
		for _, kw := range keywords {
			if strings.Contains(query, kw) {
				result.Query.Types = append(result.Query.Types, eventType)
				break
			}
		}
	}

	// Detect levels
	levelKeywords := map[string][]string{
		"error": {"error", "错误"},
		"warn":  {"warn", "warning", "警告"},
		"fatal": {"fatal", "致命"},
	}

	for level, keywords := range levelKeywords {
		for _, kw := range keywords {
			if strings.Contains(query, kw) {
				result.Query.Levels = append(result.Query.Levels, level)
				break
			}
		}
	}

	// Detect time ranges
	if strings.Contains(query, "last 5 min") || strings.Contains(query, "最近5分钟") {
		result.Query.TimeRange = &NLTimeRange{Last: "5m"}
	} else if strings.Contains(query, "last hour") || strings.Contains(query, "最近一小时") {
		result.Query.TimeRange = &NLTimeRange{Last: "1h"}
	}

	// Extract remaining keywords
	commonWords := []string{"the", "a", "an", "in", "on", "at", "to", "for", "of", "and", "or", "after", "before", "during"}
	words := strings.Fields(query)
	for _, word := range words {
		isCommon := false
		for _, cw := range commonWords {
			if word == cw {
				isCommon = true
				break
			}
		}
		if !isCommon && len(word) > 2 {
			result.Query.Keywords = append(result.Query.Keywords, word)
		}
	}

	result.Explanation = "Parsed using simple keyword matching"
	return result
}

// ========================================
// Log Analysis - 智能日志分析
// ========================================

// LogAnalysisResult represents the result of analyzing a log entry
type LogAnalysisResult struct {
	Classification  string   `json:"classification"`  // "error", "warning", "noise", "important"
	Severity        float32  `json:"severity"`        // 0.0 - 1.0
	RelatedTags     []string `json:"relatedTags"`     // Related log tags
	SuggestedAction string   `json:"suggestedAction"` // "investigate", "ignore", "monitor"
	Summary         string   `json:"summary"`         // Brief summary
}

// AnalyzeLog analyzes a single log entry
func (a *App) AnalyzeLog(tag, message, level string) (*LogAnalysisResult, error) {
	a.aiServiceMu.RLock()
	aiService := a.aiService
	config := a.aiConfigMgr.GetConfig()
	a.aiServiceMu.RUnlock()

	// Default classification based on level
	result := &LogAnalysisResult{
		Classification:  "info",
		Severity:        0.3,
		SuggestedAction: "ignore",
	}

	// Simple heuristic classification
	switch strings.ToLower(level) {
	case "e", "error":
		result.Classification = "error"
		result.Severity = 0.8
		result.SuggestedAction = "investigate"
	case "w", "warn", "warning":
		result.Classification = "warning"
		result.Severity = 0.5
		result.SuggestedAction = "monitor"
	case "f", "fatal":
		result.Classification = "error"
		result.Severity = 1.0
		result.SuggestedAction = "investigate"
	}

	// Check for known noise patterns
	noisePatterns := []string{
		"GC_CONCURRENT",
		"dalvikvm",
		"FinalizerDaemon",
		"HeapTaskDaemon",
		"ReferenceQueueDaemon",
	}
	for _, pattern := range noisePatterns {
		if strings.Contains(tag, pattern) || strings.Contains(message, pattern) {
			result.Classification = "noise"
			result.Severity = 0.1
			result.SuggestedAction = "ignore"
			return result, nil
		}
	}

	// Check for important patterns
	importantPatterns := []string{
		"Exception",
		"Error",
		"FATAL",
		"crash",
		"ANR",
		"OutOfMemory",
	}
	for _, pattern := range importantPatterns {
		if strings.Contains(message, pattern) {
			result.Classification = "important"
			result.Severity = 0.9
			result.SuggestedAction = "investigate"
			break
		}
	}

	// Use AI for more sophisticated analysis if available
	if aiService != nil && aiService.IsReady() && config.Features.LogAnalysis {
		aiResult, err := a.analyzeLogWithAI(tag, message, level)
		if err == nil && aiResult != nil {
			return aiResult, nil
		}
	}

	return result, nil
}

func (a *App) analyzeLogWithAI(tag, message, level string) (*LogAnalysisResult, error) {
	aiService := a.aiService
	if aiService == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	prompt := fmt.Sprintf(`Analyze this Android log entry:
Tag: %s
Level: %s
Message: %s

Classify as: error, warning, noise, or important
Rate severity: 0.0-1.0
Suggest action: investigate, ignore, or monitor

Return JSON: {"classification": "...", "severity": 0.0, "suggestedAction": "...", "summary": "..."}`, tag, level, message)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := aiService.Complete(ctx, &CompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are an Android log analyzer. Return only valid JSON."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   200,
		Temperature: 0.1,
	})

	if err != nil {
		return nil, err
	}

	var result LogAnalysisResult
	response := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ========================================
// Crash Analysis - 崩溃分析
// ========================================

// RootCauseAnalysis represents crash root cause analysis result
type RootCauseAnalysis struct {
	CrashEventID   string           `json:"crashEventId"`
	ProbableCauses []CauseCandidate `json:"probableCauses"`
	RelatedEvents  []string         `json:"relatedEvents"` // Event IDs
	Summary        string           `json:"summary"`
	Recommendations []string        `json:"recommendations"`
	Confidence     float32          `json:"confidence"`
}

// CauseCandidate represents a probable cause
type CauseCandidate struct {
	EventID     string  `json:"eventId,omitempty"`
	Type        string  `json:"type"`        // "network_error", "memory", "exception", etc.
	Description string  `json:"description"`
	Probability float32 `json:"probability"`
}

// AnalyzeCrashRootCause analyzes a crash event and its preceding events
func (a *App) AnalyzeCrashRootCause(crashEventID string, sessionID string) (*RootCauseAnalysis, error) {
	if a.eventStore == nil {
		return nil, fmt.Errorf("event store not initialized")
	}

	// Get all events from the session to find the crash event
	result, err := a.eventStore.QueryEvents(EventQuery{
		SessionID: sessionID,
		Limit:     1000,
	})
	if err != nil {
		return nil, err
	}

	var crashEvent *UnifiedEvent
	for i := range result.Events {
		if result.Events[i].ID == crashEventID {
			crashEvent = &result.Events[i]
			break
		}
	}

	if crashEvent == nil {
		return nil, fmt.Errorf("crash event not found")
	}

	// Get preceding events (last 30 seconds before crash)
	startTime := crashEvent.Timestamp - 30000 // 30 seconds before
	precedingResult, err := a.eventStore.QueryEvents(EventQuery{
		SessionID: sessionID,
		StartTime: startTime,
		EndTime:   crashEvent.Timestamp,
		Limit:     100,
	})
	if err != nil {
		return nil, err
	}

	precedingEvents := precedingResult.Events

	// Analyze with simple heuristics first
	analysis := &RootCauseAnalysis{
		CrashEventID:   crashEventID,
		ProbableCauses: []CauseCandidate{},
		RelatedEvents:  []string{},
		Confidence:     0.5,
	}

	// Look for common crash patterns
	for _, event := range precedingEvents {
		// Check for network errors
		if event.Source == SourceNetwork && event.Level == LevelError {
			analysis.ProbableCauses = append(analysis.ProbableCauses, CauseCandidate{
				EventID:     event.ID,
				Type:        "network_error",
				Description: fmt.Sprintf("Network error: %s", event.Title),
				Probability: 0.6,
			})
			analysis.RelatedEvents = append(analysis.RelatedEvents, event.ID)
		}

		// Check for memory issues
		if strings.Contains(event.Title, "OutOfMemory") || strings.Contains(event.Title, "OOM") {
			analysis.ProbableCauses = append(analysis.ProbableCauses, CauseCandidate{
				EventID:     event.ID,
				Type:        "memory",
				Description: "Out of memory condition detected",
				Probability: 0.9,
			})
			analysis.RelatedEvents = append(analysis.RelatedEvents, event.ID)
		}

		// Check for exceptions in logs
		if event.Type == "logcat" && event.Level == LevelError {
			if strings.Contains(event.Title, "Exception") {
				analysis.ProbableCauses = append(analysis.ProbableCauses, CauseCandidate{
					EventID:     event.ID,
					Type:        "exception",
					Description: fmt.Sprintf("Exception: %s", event.Title),
					Probability: 0.7,
				})
				analysis.RelatedEvents = append(analysis.RelatedEvents, event.ID)
			}
		}
	}

	// Try AI analysis if available
	a.aiServiceMu.RLock()
	aiService := a.aiService
	config := a.aiConfigMgr.GetConfig()
	a.aiServiceMu.RUnlock()

	if aiService != nil && aiService.IsReady() && config.Features.CrashAnalysis {
		aiAnalysis, err := a.analyzeCrashWithAI(crashEvent, precedingEvents)
		if err == nil && aiAnalysis != nil {
			return aiAnalysis, nil
		}
	}

	// Generate summary
	if len(analysis.ProbableCauses) > 0 {
		analysis.Summary = fmt.Sprintf("Found %d potential causes for the crash", len(analysis.ProbableCauses))
		analysis.Recommendations = []string{
			"Review the identified error events",
			"Check memory usage patterns",
			"Verify network connectivity handling",
		}
	} else {
		analysis.Summary = "No obvious cause found in preceding events"
		analysis.Recommendations = []string{
			"Review the crash stack trace",
			"Check for threading issues",
			"Analyze device state at crash time",
		}
	}

	return analysis, nil
}

func (a *App) analyzeCrashWithAI(crashEvent *UnifiedEvent, precedingEvents []UnifiedEvent) (*RootCauseAnalysis, error) {
	aiService := a.aiService
	if aiService == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	// Build context from events
	var eventSummaries []string
	for i, e := range precedingEvents {
		if i >= 20 { // Limit to 20 events
			break
		}
		eventSummaries = append(eventSummaries, fmt.Sprintf("[%s] %s: %s", e.Source, e.Type, e.Title))
	}

	prompt := fmt.Sprintf(`Analyze this Android app crash:

Crash Event: %s - %s

Events before crash (chronological):
%s

Identify probable causes and provide recommendations.

Return JSON:
{
  "probableCauses": [{"type": "...", "description": "...", "probability": 0.0}],
  "summary": "...",
  "recommendations": ["..."],
  "confidence": 0.0
}`, crashEvent.Type, crashEvent.Title, strings.Join(eventSummaries, "\n"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := aiService.Complete(ctx, &CompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are an Android crash analyst. Analyze crash events and identify root causes. Return only valid JSON."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   800,
		Temperature: 0.3,
	})

	if err != nil {
		return nil, err
	}

	var result RootCauseAnalysis
	response := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	result.CrashEventID = crashEvent.ID
	return &result, nil
}

// ========================================
// Assertion Suggestion - 断言建议
// ========================================

// AssertionSuggestion represents a suggested assertion
type AssertionSuggestion struct {
	Type        string   `json:"type"`        // Assertion type
	Condition   string   `json:"condition"`   // Assertion condition
	Description string   `json:"description"` // Human-readable description
	Confidence  float32  `json:"confidence"`
	Examples    []string `json:"examples"`    // Example event IDs that support this assertion
}

// SuggestAssertions suggests assertions based on session events
func (a *App) SuggestAssertions(sessionID string) ([]AssertionSuggestion, error) {
	if a.eventStore == nil {
		return nil, fmt.Errorf("event store not initialized")
	}

	// Get session events
	result, err := a.eventStore.QueryEvents(EventQuery{
		SessionID: sessionID,
		Limit:     500,
	})
	if err != nil {
		return nil, err
	}

	events := result.Events
	suggestions := []AssertionSuggestion{}

	// Analyze event patterns
	typeCounts := make(map[string]int)
	errorCount := 0
	hasNetworkRequests := false
	hasAppStart := false

	for _, e := range events {
		typeCounts[e.Type]++
		if e.Level == LevelError {
			errorCount++
		}
		if e.Source == SourceNetwork {
			hasNetworkRequests = true
		}
		if e.Type == "app_start" {
			hasAppStart = true
		}
	}

	// Suggest error count assertion
	if errorCount == 0 {
		suggestions = append(suggestions, AssertionSuggestion{
			Type:        "event_count",
			Condition:   `level == "error" && count == 0`,
			Description: "No errors should occur during the session",
			Confidence:  0.8,
		})
	}

	// Suggest app start assertion
	if hasAppStart {
		suggestions = append(suggestions, AssertionSuggestion{
			Type:        "event_exists",
			Condition:   `type == "app_start"`,
			Description: "App should start successfully",
			Confidence:  0.9,
		})
	}

	// Suggest network success assertion
	if hasNetworkRequests {
		suggestions = append(suggestions, AssertionSuggestion{
			Type:        "event_count",
			Condition:   `source == "network" && level == "error" && count <= 0`,
			Description: "No network errors should occur",
			Confidence:  0.7,
		})
	}

	// Try AI suggestions if available
	a.aiServiceMu.RLock()
	aiService := a.aiService
	config := a.aiConfigMgr.GetConfig()
	a.aiServiceMu.RUnlock()

	if aiService != nil && aiService.IsReady() && config.Features.AssertionGen {
		aiSuggestions, err := a.suggestAssertionsWithAI(events)
		if err == nil && len(aiSuggestions) > 0 {
			suggestions = append(suggestions, aiSuggestions...)
		}
	}

	return suggestions, nil
}

// ========================================
// Session Summary - 会话总结
// ========================================

// SessionSummary represents an AI-generated session summary
type SessionSummary struct {
	SessionID    string              `json:"sessionId"`
	Duration     int64               `json:"duration"`     // Duration in ms
	EventCount   int                 `json:"eventCount"`   // Total events
	Overview     string              `json:"overview"`     // Brief overview
	KeyFindings  []SessionFinding    `json:"keyFindings"`  // Important findings
	Timeline     []TimelineHighlight `json:"timeline"`     // Key moments
	Statistics   SessionStats        `json:"statistics"`   // Event statistics
	Suggestions  []string            `json:"suggestions"`  // Improvement suggestions
	GeneratedAt  int64               `json:"generatedAt"`  // When generated
}

// SessionFinding represents a key finding in the session
type SessionFinding struct {
	Type        string  `json:"type"`        // "error", "warning", "info", "success"
	Title       string  `json:"title"`       // Short title
	Description string  `json:"description"` // Detailed description
	EventIDs    []string `json:"eventIds,omitempty"` // Related event IDs
	Severity    float32 `json:"severity"`    // 0.0 - 1.0
}

// TimelineHighlight represents a key moment in the timeline
type TimelineHighlight struct {
	Timestamp   int64  `json:"timestamp"`
	RelativeMs  int64  `json:"relativeMs"` // Relative to session start
	Title       string `json:"title"`
	Description string `json:"description"`
	EventID     string `json:"eventId,omitempty"`
}

// SessionStats contains session statistics
type SessionStats struct {
	TotalEvents     int            `json:"totalEvents"`
	EventsBySource  map[string]int `json:"eventsBySource"`
	EventsByLevel   map[string]int `json:"eventsByLevel"`
	ErrorCount      int            `json:"errorCount"`
	WarningCount    int            `json:"warningCount"`
	CrashCount      int            `json:"crashCount"`
	NetworkRequests int            `json:"networkRequests"`
	TouchEvents     int            `json:"touchEvents"`
}

// SummarizeSession generates an AI summary of a session
func (a *App) SummarizeSession(sessionID string) (*SessionSummary, error) {
	if a.eventStore == nil {
		return nil, fmt.Errorf("event store not initialized")
	}

	// Get session info
	session, err := a.eventStore.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Get all events from the session
	result, err := a.eventStore.QueryEvents(EventQuery{
		SessionID: sessionID,
		Limit:     2000, // Get more events for comprehensive summary
	})
	if err != nil {
		return nil, err
	}

	events := result.Events

	// Calculate basic statistics
	stats := calculateSessionStats(events)

	// Calculate duration
	var duration int64
	if session.EndTime > 0 {
		duration = session.EndTime - session.StartTime
	} else {
		duration = time.Now().UnixMilli() - session.StartTime
	}

	// Build basic summary
	summary := &SessionSummary{
		SessionID:   sessionID,
		Duration:    duration,
		EventCount:  len(events),
		Statistics:  stats,
		GeneratedAt: time.Now().UnixMilli(),
		KeyFindings: []SessionFinding{},
		Timeline:    []TimelineHighlight{},
		Suggestions: []string{},
	}

	// Extract key findings from events (heuristic analysis)
	summary.KeyFindings = extractKeyFindings(events, session.StartTime)
	summary.Timeline = extractTimelineHighlights(events, session.StartTime)

	// Generate basic overview
	summary.Overview = generateBasicOverview(stats, duration)

	// Try AI-enhanced summary if available
	a.aiServiceMu.RLock()
	aiService := a.aiService
	config := a.aiConfigMgr.GetConfig()
	a.aiServiceMu.RUnlock()

	if aiService != nil && aiService.IsReady() && config.Enabled {
		aiSummary, err := a.summarizeSessionWithAI(session, events, stats)
		if err == nil && aiSummary != nil {
			// Merge AI summary with basic summary
			summary.Overview = aiSummary.Overview
			if len(aiSummary.KeyFindings) > 0 {
				summary.KeyFindings = aiSummary.KeyFindings
			}
			if len(aiSummary.Suggestions) > 0 {
				summary.Suggestions = aiSummary.Suggestions
			}
		}
	}

	return summary, nil
}

// calculateSessionStats calculates statistics from events
func calculateSessionStats(events []UnifiedEvent) SessionStats {
	stats := SessionStats{
		TotalEvents:    len(events),
		EventsBySource: make(map[string]int),
		EventsByLevel:  make(map[string]int),
	}

	for _, e := range events {
		stats.EventsBySource[string(e.Source)]++
		stats.EventsByLevel[string(e.Level)]++

		switch e.Level {
		case LevelError, LevelFatal:
			stats.ErrorCount++
		case LevelWarn:
			stats.WarningCount++
		}

		switch e.Type {
		case "app_crash", "app_anr":
			stats.CrashCount++
		case "http_request":
			stats.NetworkRequests++
		case "touch", "gesture":
			stats.TouchEvents++
		}
	}

	return stats
}

// extractKeyFindings extracts important findings from events
func extractKeyFindings(events []UnifiedEvent, _ int64) []SessionFinding {
	findings := []SessionFinding{}

	// Track crashes
	var crashEvents []UnifiedEvent
	var errorEvents []UnifiedEvent
	var networkErrors []UnifiedEvent

	for _, e := range events {
		if e.Type == "app_crash" || e.Type == "app_anr" {
			crashEvents = append(crashEvents, e)
		}
		if e.Level == LevelError || e.Level == LevelFatal {
			errorEvents = append(errorEvents, e)
		}
		if e.Source == SourceNetwork && e.Level == LevelError {
			networkErrors = append(networkErrors, e)
		}
	}

	// Report crashes
	if len(crashEvents) > 0 {
		eventIDs := make([]string, len(crashEvents))
		for i, e := range crashEvents {
			eventIDs[i] = e.ID
		}
		findings = append(findings, SessionFinding{
			Type:        "error",
			Title:       fmt.Sprintf("%d Crash(es) Detected", len(crashEvents)),
			Description: fmt.Sprintf("Application crashed or became unresponsive %d time(s)", len(crashEvents)),
			EventIDs:    eventIDs,
			Severity:    1.0,
		})
	}

	// Report error concentration
	if len(errorEvents) > 5 {
		findings = append(findings, SessionFinding{
			Type:        "warning",
			Title:       fmt.Sprintf("%d Errors Logged", len(errorEvents)),
			Description: "High number of error-level events detected during session",
			Severity:    0.7,
		})
	}

	// Report network issues
	if len(networkErrors) > 0 {
		findings = append(findings, SessionFinding{
			Type:        "warning",
			Title:       fmt.Sprintf("%d Network Errors", len(networkErrors)),
			Description: "Network request failures detected during session",
			Severity:    0.6,
		})
	}

	// Report successful session if no major issues
	if len(crashEvents) == 0 && len(errorEvents) < 3 {
		findings = append(findings, SessionFinding{
			Type:        "success",
			Title:       "Session Completed Successfully",
			Description: "No crashes or significant errors detected",
			Severity:    0.0,
		})
	}

	return findings
}

// extractTimelineHighlights extracts key moments from the timeline
func extractTimelineHighlights(events []UnifiedEvent, sessionStart int64) []TimelineHighlight {
	highlights := []TimelineHighlight{}

	for _, e := range events {
		// Add crashes
		if e.Type == "app_crash" || e.Type == "app_anr" {
			highlights = append(highlights, TimelineHighlight{
				Timestamp:   e.Timestamp,
				RelativeMs:  e.Timestamp - sessionStart,
				Title:       "App Crash",
				Description: e.Title,
				EventID:     e.ID,
			})
		}

		// Add app starts
		if e.Type == "app_start" {
			highlights = append(highlights, TimelineHighlight{
				Timestamp:   e.Timestamp,
				RelativeMs:  e.Timestamp - sessionStart,
				Title:       "App Started",
				Description: e.Title,
				EventID:     e.ID,
			})
		}

		// Add significant errors (limit to avoid too many)
		if (e.Level == LevelError || e.Level == LevelFatal) && len(highlights) < 20 {
			highlights = append(highlights, TimelineHighlight{
				Timestamp:   e.Timestamp,
				RelativeMs:  e.Timestamp - sessionStart,
				Title:       "Error",
				Description: e.Title,
				EventID:     e.ID,
			})
		}
	}

	// Limit highlights
	if len(highlights) > 15 {
		highlights = highlights[:15]
	}

	return highlights
}

// generateBasicOverview generates a basic text overview
func generateBasicOverview(stats SessionStats, durationMs int64) string {
	durationSec := durationMs / 1000
	var durationStr string
	if durationSec < 60 {
		durationStr = fmt.Sprintf("%d seconds", durationSec)
	} else if durationSec < 3600 {
		durationStr = fmt.Sprintf("%d minutes", durationSec/60)
	} else {
		durationStr = fmt.Sprintf("%.1f hours", float64(durationSec)/3600)
	}

	overview := fmt.Sprintf("Session lasted %s with %d total events. ", durationStr, stats.TotalEvents)

	if stats.CrashCount > 0 {
		overview += fmt.Sprintf("Detected %d crash(es). ", stats.CrashCount)
	}
	if stats.ErrorCount > 0 {
		overview += fmt.Sprintf("Found %d error(s). ", stats.ErrorCount)
	}
	if stats.NetworkRequests > 0 {
		overview += fmt.Sprintf("Made %d network request(s). ", stats.NetworkRequests)
	}
	if stats.TouchEvents > 0 {
		overview += fmt.Sprintf("Recorded %d touch event(s).", stats.TouchEvents)
	}

	return overview
}

// summarizeSessionWithAI uses AI to generate a more detailed summary
func (a *App) summarizeSessionWithAI(session *DeviceSession, events []UnifiedEvent, stats SessionStats) (*SessionSummary, error) {
	aiService := a.aiService
	if aiService == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	// Build event summary for AI
	var eventSummary strings.Builder
	eventSummary.WriteString("Session Statistics:\n")
	eventSummary.WriteString(fmt.Sprintf("- Duration: %d ms\n", session.EndTime-session.StartTime))
	eventSummary.WriteString(fmt.Sprintf("- Total events: %d\n", stats.TotalEvents))
	eventSummary.WriteString(fmt.Sprintf("- Errors: %d\n", stats.ErrorCount))
	eventSummary.WriteString(fmt.Sprintf("- Warnings: %d\n", stats.WarningCount))
	eventSummary.WriteString(fmt.Sprintf("- Crashes: %d\n", stats.CrashCount))
	eventSummary.WriteString(fmt.Sprintf("- Network requests: %d\n", stats.NetworkRequests))
	eventSummary.WriteString(fmt.Sprintf("- Touch events: %d\n", stats.TouchEvents))

	eventSummary.WriteString("\nEvents by source:\n")
	for src, count := range stats.EventsBySource {
		eventSummary.WriteString(fmt.Sprintf("- %s: %d\n", src, count))
	}

	// Add sample of important events
	eventSummary.WriteString("\nKey events (sample):\n")
	count := 0
	for _, e := range events {
		if count >= 30 {
			break
		}
		if e.Level == LevelError || e.Level == LevelFatal || e.Type == "app_crash" || e.Type == "app_anr" || e.Type == "app_start" {
			eventSummary.WriteString(fmt.Sprintf("- [%s/%s] %s\n", e.Source, e.Level, e.Title))
			count++
		}
	}

	prompt := fmt.Sprintf(`Analyze this Android app testing session and provide a summary:

%s

Provide:
1. A brief overview (2-3 sentences)
2. Key findings (list important issues or observations)
3. Suggestions for improvement

Return JSON:
{
  "overview": "...",
  "keyFindings": [{"type": "error|warning|info|success", "title": "...", "description": "...", "severity": 0.0}],
  "suggestions": ["..."]
}`, eventSummary.String())

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := aiService.Complete(ctx, &CompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a mobile app testing analyst. Analyze session data and provide actionable insights. Return only valid JSON."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1000,
		Temperature: 0.3,
	})

	if err != nil {
		return nil, err
	}

	var result SessionSummary
	response := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (a *App) suggestAssertionsWithAI(events []UnifiedEvent) ([]AssertionSuggestion, error) {
	aiService := a.aiService
	if aiService == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	// Build event summary
	var eventSummary strings.Builder
	eventSummary.WriteString("Event types and counts:\n")
	typeCounts := make(map[string]int)
	for _, e := range events {
		typeCounts[e.Type]++
	}
	for t, c := range typeCounts {
		eventSummary.WriteString(fmt.Sprintf("- %s: %d\n", t, c))
	}

	prompt := fmt.Sprintf(`Based on this session's event patterns, suggest test assertions:

%s

Suggest 3-5 assertions that would help validate this behavior in future runs.

Return JSON array:
[{"type": "event_count|event_exists|timing", "condition": "...", "description": "...", "confidence": 0.0}]`, eventSummary.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := aiService.Complete(ctx, &CompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a test automation expert. Suggest assertions for mobile app testing. Return only valid JSON array."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   500,
		Temperature: 0.3,
	})

	if err != nil {
		return nil, err
	}

	var suggestions []AssertionSuggestion
	response := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	if err := json.Unmarshal([]byte(response), &suggestions); err != nil {
		return nil, err
	}

	return suggestions, nil
}
