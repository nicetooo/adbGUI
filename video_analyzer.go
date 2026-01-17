package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ========================================
// Video Analyzer - 视频AI分析
// ========================================

// VideoFrameAnalysis represents the analysis result of a video frame
type VideoFrameAnalysis struct {
	FrameTime   int64             `json:"frameTime"`   // Relative time in ms
	SceneType   string            `json:"sceneType"`   // "login", "home", "settings", "error", "loading", etc.
	OCRText     []string          `json:"ocrText"`     // Detected text on screen
	IsAnomaly   bool              `json:"isAnomaly"`   // Whether this frame shows an anomaly
	AnomalyType string            `json:"anomalyType"` // "anr", "crash", "blank", "loading_stuck"
	UIElements  []DetectedElement `json:"uiElements"`  // Detected UI elements
	Confidence  float32           `json:"confidence"`  // Analysis confidence
	Description string            `json:"description"` // Human-readable description
}

// DetectedElement represents a detected UI element in a frame
type DetectedElement struct {
	Type   string      `json:"type"`   // "button", "input", "text", "image", "dialog"
	Text   string      `json:"text"`   // Text content if any
	Bounds [4]float64  `json:"bounds"` // [x, y, width, height] - can be normalized (0-1) or pixel values
	Score  float32     `json:"score"`  // Detection confidence
}

// VideoAnalysisResult represents the full analysis of a video
type VideoAnalysisResult struct {
	VideoPath      string               `json:"videoPath"`
	DurationMs     int64                `json:"durationMs"`
	KeyFrames      []VideoFrameAnalysis `json:"keyFrames"`
	AnomalyFrames  []VideoFrameAnalysis `json:"anomalyFrames"`
	SceneChanges   []int64              `json:"sceneChanges"` // Timestamps where scene changes
	Summary        string               `json:"summary"`
	AnalyzedAt     int64                `json:"analyzedAt"`
}

// VideoAnalyzer handles video analysis
type VideoAnalyzer struct {
	videoService *VideoService
	aiService    *AIService
	cacheDir     string
}

// NewVideoAnalyzer creates a new video analyzer
func NewVideoAnalyzer(videoService *VideoService, aiService *AIService, cacheDir string) *VideoAnalyzer {
	return &VideoAnalyzer{
		videoService: videoService,
		aiService:    aiService,
		cacheDir:     cacheDir,
	}
}

// AnalyzeVideo performs full analysis on a video
func (va *VideoAnalyzer) AnalyzeVideo(ctx context.Context, videoPath string, intervalMs int64) (*VideoAnalysisResult, error) {
	if va.videoService == nil || !va.videoService.IsAvailable() {
		return nil, fmt.Errorf("video service not available")
	}

	// Get video metadata
	meta, err := va.videoService.GetMetadata(videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get video metadata: %w", err)
	}

	result := &VideoAnalysisResult{
		VideoPath:     videoPath,
		DurationMs:    meta.DurationMs,
		KeyFrames:     []VideoFrameAnalysis{},
		AnomalyFrames: []VideoFrameAnalysis{},
		SceneChanges:  []int64{},
		AnalyzedAt:    time.Now().Unix(),
	}

	// Extract and analyze frames at intervals
	if intervalMs <= 0 {
		intervalMs = 2000 // Default 2 seconds
	}

	var lastSceneType string
	for timeMs := int64(0); timeMs < meta.DurationMs; timeMs += intervalMs {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		analysis, err := va.AnalyzeFrame(ctx, videoPath, timeMs)
		if err != nil {
			continue // Skip frames that fail to analyze
		}

		result.KeyFrames = append(result.KeyFrames, *analysis)

		// Track scene changes
		if analysis.SceneType != lastSceneType && lastSceneType != "" {
			result.SceneChanges = append(result.SceneChanges, timeMs)
		}
		lastSceneType = analysis.SceneType

		// Track anomalies
		if analysis.IsAnomaly {
			result.AnomalyFrames = append(result.AnomalyFrames, *analysis)
		}
	}

	// Generate summary
	if va.aiService != nil && va.aiService.IsReady() {
		result.Summary = va.generateSummary(ctx, result)
	} else {
		result.Summary = fmt.Sprintf("Analyzed %d frames, found %d anomalies, %d scene changes",
			len(result.KeyFrames), len(result.AnomalyFrames), len(result.SceneChanges))
	}

	return result, nil
}

// AnalyzeFrame analyzes a single video frame
func (va *VideoAnalyzer) AnalyzeFrame(ctx context.Context, videoPath string, timeMs int64) (*VideoFrameAnalysis, error) {
	// Extract frame as base64
	frameBase64, err := va.videoService.ExtractFrameBase64(videoPath, timeMs, 720)
	if err != nil {
		return nil, fmt.Errorf("failed to extract frame: %w", err)
	}

	analysis := &VideoFrameAnalysis{
		FrameTime:  timeMs,
		Confidence: 0.5,
	}

	// Check AI availability and vision support
	aiReady := va.aiService != nil && va.aiService.IsReady()
	supportsVision := va.aiService != nil && va.aiService.SupportsVision()

	LogDebug("video_analyzer").
		Bool("aiServiceNil", va.aiService == nil).
		Bool("aiServiceReady", aiReady).
		Bool("supportsVision", supportsVision).
		Int64("timeMs", timeMs).
		Msg("Checking AI for frame analysis")

	if aiReady && supportsVision {
		aiAnalysis, err := va.analyzeFrameWithAI(ctx, frameBase64)
		if err != nil {
			LogWarn("video_analyzer").Err(err).Int64("timeMs", timeMs).Msg("AI frame analysis failed")
		}
		if err == nil && aiAnalysis != nil {
			LogInfo("video_analyzer").Int64("timeMs", timeMs).Str("sceneType", aiAnalysis.SceneType).Msg("AI frame analysis success")
			return aiAnalysis, nil
		}
	}

	// Fallback: basic heuristic analysis
	analysis.SceneType = "unknown"
	if !aiReady {
		analysis.Description = "Frame analysis not available (AI service not ready)"
	} else if !supportsVision {
		analysis.Description = fmt.Sprintf("Frame analysis skipped (model %s does not support vision)", va.aiService.GetProvider().Model())
	} else {
		analysis.Description = "Frame analysis failed"
	}
	LogDebug("video_analyzer").Int64("timeMs", timeMs).Str("reason", analysis.Description).Msg("Falling back to heuristic analysis")

	return analysis, nil
}

// analyzeFrameWithAI uses LLM to analyze a frame
func (va *VideoAnalyzer) analyzeFrameWithAI(ctx context.Context, frameBase64 string) (*VideoFrameAnalysis, error) {
	if va.aiService == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	// Strong prompt to enforce JSON output
	prompt := `IMPORTANT: You MUST respond with ONLY a valid JSON object. No text before or after the JSON.

Analyze this mobile app screenshot and return this exact JSON structure:
{"sceneType":"home","ocrText":["text1","text2"],"isAnomaly":false,"anomalyType":"none","uiElements":[{"type":"button","text":"Click","bounds":[0.1,0.2,0.3,0.1],"score":0.9}],"confidence":0.9,"description":"Brief screen description"}

Field requirements:
- sceneType: one of "login", "home", "settings", "error", "loading", "dialog", "list", "detail", "other"
- ocrText: array of visible text strings on screen
- isAnomaly: true only for error/crash/ANR screens
- anomalyType: "none", "anr", "crash", "blank", or "loading_stuck"
- uiElements: 2-4 main UI elements with bounds as [x,y,width,height] normalized 0-1
- confidence: 0.0-1.0
- description: one sentence describing the screen

RESPOND WITH ONLY THE JSON OBJECT, NOTHING ELSE.`

	// Extract base64 data without the data URL prefix
	imageData := frameBase64
	if strings.HasPrefix(frameBase64, "data:") {
		// Remove "data:image/jpeg;base64," prefix
		parts := strings.SplitN(frameBase64, ",", 2)
		if len(parts) == 2 {
			imageData = parts[1]
		}
	}

	resp, err := va.aiService.Complete(ctx, &CompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a JSON-only API. You analyze mobile app screenshots and return ONLY valid JSON objects. Never include explanations, markdown, or any text outside the JSON. Your entire response must be parseable as JSON."},
			{
				Role:    "user",
				Content: prompt,
				Images: []ImageContent{
					{
						Type:   "base64",
						Data:   imageData,
						Format: "jpeg",
					},
				},
			},
		},
		MaxTokens:   500, // Reduced since we only need JSON
		Temperature: 0.1, // Lower temperature for more consistent output
		Timeout:     180 * time.Second, // Vision models need more time
	})

	if err != nil {
		return nil, err
	}

	var analysis VideoFrameAnalysis
	response := strings.TrimSpace(resp.Content)

	// Log raw response for debugging
	LogDebug("video_analyzer").
		Str("rawResponse", response[:min(len(response), 500)]).
		Msg("AI raw response")

	// Try to extract JSON from the response
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("failed to extract JSON from AI response: %s", response[:min(len(response), 200)])
	}

	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse AI response JSON: %w, content: %s", err, jsonStr[:min(len(jsonStr), 200)])
	}

	return &analysis, nil
}

// extractJSON tries to extract a JSON object from a string that may contain other text
func extractJSON(s string) string {
	s = strings.TrimSpace(s)

	// Remove markdown code blocks
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	}
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	s = strings.TrimSpace(s)

	// Find JSON object boundaries
	start := strings.Index(s, "{")
	if start == -1 {
		return ""
	}

	// Find matching closing brace
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}

	// If we didn't find a matching brace, try to return what we have
	if end := strings.LastIndex(s, "}"); end > start {
		return s[start : end+1]
	}

	return ""
}

// generateSummary generates a summary of the video analysis
func (va *VideoAnalyzer) generateSummary(ctx context.Context, result *VideoAnalysisResult) string {
	if va.aiService == nil || !va.aiService.IsReady() {
		return ""
	}

	// Build summary of frames
	var sceneSummary []string
	sceneCount := make(map[string]int)
	for _, frame := range result.KeyFrames {
		sceneCount[frame.SceneType]++
	}
	for scene, count := range sceneCount {
		sceneSummary = append(sceneSummary, fmt.Sprintf("%s: %d frames", scene, count))
	}

	var anomalySummary []string
	for _, anomaly := range result.AnomalyFrames {
		anomalySummary = append(anomalySummary, fmt.Sprintf("- %s at %dms: %s", anomaly.AnomalyType, anomaly.FrameTime, anomaly.Description))
	}

	prompt := fmt.Sprintf(`Summarize this video analysis:

Duration: %dms
Scenes: %s
Scene changes: %d
Anomalies: %d
%s

Provide a brief 2-3 sentence summary of what happened in this video session.`,
		result.DurationMs,
		strings.Join(sceneSummary, ", "),
		len(result.SceneChanges),
		len(result.AnomalyFrames),
		strings.Join(anomalySummary, "\n"))

	resp, err := va.aiService.Complete(ctx, &CompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a mobile app testing analyst. Summarize video analysis results concisely."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   200,
		Temperature: 0.3,
	})

	if err != nil {
		return ""
	}

	return strings.TrimSpace(resp.Content)
}

// DetectAnomalies quickly scans for anomaly frames
func (va *VideoAnalyzer) DetectAnomalies(ctx context.Context, videoPath string) ([]VideoFrameAnalysis, error) {
	// Use scene change detection to find potential anomaly points
	if va.videoService == nil {
		return nil, fmt.Errorf("video service not available")
	}

	meta, err := va.videoService.GetMetadata(videoPath)
	if err != nil {
		return nil, err
	}

	var anomalies []VideoFrameAnalysis

	// Sample at higher frequency for anomaly detection
	intervalMs := int64(1000) // 1 second
	for timeMs := int64(0); timeMs < meta.DurationMs; timeMs += intervalMs {
		select {
		case <-ctx.Done():
			return anomalies, ctx.Err()
		default:
		}

		analysis, err := va.AnalyzeFrame(ctx, videoPath, timeMs)
		if err != nil {
			continue
		}

		if analysis.IsAnomaly {
			anomalies = append(anomalies, *analysis)
		}
	}

	return anomalies, nil
}

// SaveAnalysisCache saves analysis result to cache
func (va *VideoAnalyzer) SaveAnalysisCache(videoPath string, result *VideoAnalysisResult) error {
	cacheFile := va.getCacheFilePath(videoPath)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cacheFile), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

// LoadAnalysisCache loads analysis result from cache
func (va *VideoAnalyzer) LoadAnalysisCache(videoPath string) (*VideoAnalysisResult, error) {
	cacheFile := va.getCacheFilePath(videoPath)

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var result VideoAnalysisResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (va *VideoAnalyzer) getCacheFilePath(videoPath string) string {
	// Create a unique cache file name based on video path
	hash := base64.URLEncoding.EncodeToString([]byte(videoPath))
	if len(hash) > 32 {
		hash = hash[:32]
	}
	return filepath.Join(va.cacheDir, "video_analysis", hash+".json")
}

// ========================================
// API Methods for App
// ========================================

// AnalyzeSessionVideo analyzes the video for a session
func (a *App) AnalyzeSessionVideo(sessionID string, intervalMs int64) (*VideoAnalysisResult, error) {
	// Get session info
	videoInfo, err := a.GetSessionVideoInfo(sessionID)
	if err != nil {
		return nil, err
	}

	hasVideo, ok := videoInfo["hasVideo"].(bool)
	if !ok || !hasVideo {
		return nil, fmt.Errorf("session has no video")
	}

	videoPath, ok := videoInfo["videoPath"].(string)
	if !ok || videoPath == "" {
		return nil, fmt.Errorf("video path not found")
	}

	// Create analyzer
	a.aiServiceMu.RLock()
	aiService := a.aiService
	a.aiServiceMu.RUnlock()

	videoService := a.getVideoService()
	analyzer := NewVideoAnalyzer(videoService, aiService, a.dataDir)

	// Check cache first
	cached, err := analyzer.LoadAnalysisCache(videoPath)
	if err == nil && cached != nil {
		return cached, nil
	}

	// Perform analysis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := analyzer.AnalyzeVideo(ctx, videoPath, intervalMs)
	if err != nil {
		return nil, err
	}

	// Save to cache
	_ = analyzer.SaveAnalysisCache(videoPath, result)

	return result, nil
}

// AnalyzeVideoFrame analyzes a single frame
func (a *App) AnalyzeVideoFrame(videoPath string, timeMs int64) (*VideoFrameAnalysis, error) {
	a.aiServiceMu.RLock()
	aiService := a.aiService
	a.aiServiceMu.RUnlock()

	videoService := a.getVideoService()
	analyzer := NewVideoAnalyzer(videoService, aiService, a.dataDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return analyzer.AnalyzeFrame(ctx, videoPath, timeMs)
}

// DetectVideoAnomalies detects anomalies in a video
func (a *App) DetectVideoAnomalies(videoPath string) ([]VideoFrameAnalysis, error) {
	a.aiServiceMu.RLock()
	aiService := a.aiService
	a.aiServiceMu.RUnlock()

	videoService := a.getVideoService()
	analyzer := NewVideoAnalyzer(videoService, aiService, a.dataDir)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	return analyzer.DetectAnomalies(ctx, videoPath)
}
