package mcp

import (
	"errors"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// ========================================
// video_frame Tests
// ========================================

func TestHandleVideoFrame_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetVideoFrameResult = "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"video_path": "/path/to/video.mp4",
		"time_ms":    float64(5000),
		"width":      float64(720),
	}

	result, err := server.handleVideoFrame(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Content) < 2 {
		t.Fatal("Expected at least 2 content items (text + image)")
	}

	// Verify text content
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "5000ms") {
		t.Error("Expected time position in output")
	}
	if !strings.Contains(text, "720") {
		t.Error("Expected width in output")
	}

	// Verify mock was called with correct args
	if !mock.WasMethodCalled("GetVideoFrame") {
		t.Error("Expected GetVideoFrame to be called")
	}
	call := mock.GetLastCallByMethod("GetVideoFrame")
	if call.Args[0] != "/path/to/video.mp4" {
		t.Errorf("Expected video path '/path/to/video.mp4', got %v", call.Args[0])
	}
	if call.Args[1] != int64(5000) {
		t.Errorf("Expected time_ms 5000, got %v", call.Args[1])
	}
	if call.Args[2] != 720 {
		t.Errorf("Expected width 720, got %v", call.Args[2])
	}
}

func TestHandleVideoFrame_DefaultWidth(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetVideoFrameResult = "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"video_path": "/path/to/video.mp4",
		"time_ms":    float64(1000),
		// width not specified, should default to 720
	}

	_, err := server.handleVideoFrame(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	call := mock.GetLastCallByMethod("GetVideoFrame")
	if call.Args[2] != 720 {
		t.Errorf("Expected default width 720, got %v", call.Args[2])
	}
}

func TestHandleVideoFrame_MissingVideoPath(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"time_ms": float64(5000),
	}

	_, err := server.handleVideoFrame(nil, request)
	if err == nil {
		t.Fatal("Expected error for missing video_path")
	}
	if !strings.Contains(err.Error(), "video_path") {
		t.Errorf("Expected error about video_path, got: %v", err)
	}
}

func TestHandleVideoFrame_EmptyVideoPath(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"video_path": "",
		"time_ms":    float64(5000),
	}

	_, err := server.handleVideoFrame(nil, request)
	if err == nil {
		t.Fatal("Expected error for empty video_path")
	}
}

func TestHandleVideoFrame_MissingTimeMs(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"video_path": "/path/to/video.mp4",
	}

	_, err := server.handleVideoFrame(nil, request)
	if err == nil {
		t.Fatal("Expected error for missing time_ms")
	}
	if !strings.Contains(err.Error(), "time_ms") {
		t.Errorf("Expected error about time_ms, got: %v", err)
	}
}

func TestHandleVideoFrame_ExtractError(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetVideoFrameError = errors.New("ffmpeg not found")
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"video_path": "/path/to/video.mp4",
		"time_ms":    float64(5000),
	}

	_, err := server.handleVideoFrame(nil, request)
	if err == nil {
		t.Fatal("Expected error when extraction fails")
	}
	if !strings.Contains(err.Error(), "ffmpeg not found") {
		t.Errorf("Expected ffmpeg error, got: %v", err)
	}
}

func TestHandleVideoFrame_ZeroTimeMs(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetVideoFrameResult = "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"video_path": "/path/to/video.mp4",
		"time_ms":    float64(0),
	}

	_, err := server.handleVideoFrame(nil, request)
	if err != nil {
		t.Fatalf("Expected no error for time_ms=0, got %v", err)
	}

	call := mock.GetLastCallByMethod("GetVideoFrame")
	if call.Args[1] != int64(0) {
		t.Errorf("Expected time_ms 0, got %v", call.Args[1])
	}
}

// ========================================
// video_metadata Tests
// ========================================

func TestHandleVideoMetadata_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetVideoMetadataResult = &VideoMetadata{
		Path:        "/path/to/video.mp4",
		Duration:    120.5,
		DurationMs:  120500,
		Width:       1920,
		Height:      1080,
		FrameRate:   30.0,
		Codec:       "h264",
		BitRate:     5000000,
		TotalFrames: 3615,
	}
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"video_path": "/path/to/video.mp4",
	}

	result, err := server.handleVideoMetadata(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text

	// Verify all metadata fields are in output
	expectedContents := []string{
		"120.50s",     // Duration
		"120500ms",    // DurationMs
		"1920x1080",   // Resolution
		"30.00 fps",   // Frame Rate
		"3615",        // Total Frames
		"h264",        // Codec
		"5000000 bps", // Bit Rate
	}

	for _, expected := range expectedContents {
		if !strings.Contains(text, expected) {
			t.Errorf("Expected '%s' in output, got: %s", expected, text)
		}
	}
}

func TestHandleVideoMetadata_MissingVideoPath(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{}

	_, err := server.handleVideoMetadata(nil, request)
	if err == nil {
		t.Fatal("Expected error for missing video_path")
	}
}

func TestHandleVideoMetadata_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetVideoMetadataError = errors.New("file not found")
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"video_path": "/nonexistent/video.mp4",
	}

	_, err := server.handleVideoMetadata(nil, request)
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("Expected 'file not found' error, got: %v", err)
	}
}

// ========================================
// session_video_info Tests
// ========================================

func TestHandleSessionVideoInfo_WithVideo(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoResult = map[string]interface{}{
		"hasVideo":    true,
		"videoPath":   "/path/to/recording.mp4",
		"videoOffset": int64(1000),
		"metadata": map[string]interface{}{
			"durationMs": float64(60000),
			"width":      float64(1080),
			"height":     float64(1920),
			"frameRate":  float64(30),
			"codec":      "h264",
		},
	}
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id": "session-123",
	}

	result, err := server.handleSessionVideoInfo(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "recording.mp4") {
		t.Error("Expected video path in output")
	}
	if !strings.Contains(text, "1000ms") {
		t.Error("Expected video offset in output")
	}
	if !strings.Contains(text, "Metadata") {
		t.Error("Expected metadata section in output")
	}
}

func TestHandleSessionVideoInfo_WithVideoMetadataStruct(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoResult = map[string]interface{}{
		"hasVideo":    true,
		"videoPath":   "/path/to/recording.mp4",
		"videoOffset": float64(2000),
		"metadata": &VideoMetadata{
			Duration:  60.5,
			Width:     1080,
			Height:    1920,
			FrameRate: 30.0,
			Codec:     "h264",
		},
	}
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id": "session-123",
	}

	result, err := server.handleSessionVideoInfo(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "2000ms") {
		t.Error("Expected video offset in output")
	}
}

func TestHandleSessionVideoInfo_NoVideo(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoResult = map[string]interface{}{
		"hasVideo": false,
	}
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id": "session-456",
	}

	result, err := server.handleSessionVideoInfo(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "no video recording") {
		t.Error("Expected 'no video recording' message")
	}
}

func TestHandleSessionVideoInfo_NoVideoWithError(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoResult = map[string]interface{}{
		"hasVideo": false,
		"error":    "Video file was deleted",
	}
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id": "session-789",
	}

	result, err := server.handleSessionVideoInfo(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Video file was deleted") {
		t.Error("Expected error message in output")
	}
}

func TestHandleSessionVideoInfo_MissingSessionId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{}

	_, err := server.handleSessionVideoInfo(nil, request)
	if err == nil {
		t.Fatal("Expected error for missing session_id")
	}
	if !strings.Contains(err.Error(), "session_id") {
		t.Errorf("Expected error about session_id, got: %v", err)
	}
}

func TestHandleSessionVideoInfo_Error(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoError = errors.New("session not found")
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id": "nonexistent",
	}

	_, err := server.handleSessionVideoInfo(nil, request)
	if err == nil {
		t.Fatal("Expected error for nonexistent session")
	}
}

// ========================================
// session_video_frame Tests
// ========================================

func TestHandleSessionVideoFrame_Success(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoResult = map[string]interface{}{
		"hasVideo":    true,
		"videoPath":   "/path/to/recording.mp4",
		"videoOffset": int64(500),
	}
	mock.GetVideoFrameResult = "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id":    "session-123",
		"event_time_ms": float64(5000),
	}

	result, err := server.handleSessionVideoFrame(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Content) < 2 {
		t.Fatal("Expected at least 2 content items")
	}

	// Verify text shows both event time and video time
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "5000ms") {
		t.Error("Expected event time in output")
	}
	if !strings.Contains(text, "4500ms") {
		t.Error("Expected video time (5000-500=4500) in output")
	}

	// Verify GetVideoFrame was called with correct video time
	call := mock.GetLastCallByMethod("GetVideoFrame")
	if call.Args[1] != int64(4500) {
		t.Errorf("Expected video time 4500 (5000-500), got %v", call.Args[1])
	}
}

func TestHandleSessionVideoFrame_WithFloatOffset(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoResult = map[string]interface{}{
		"hasVideo":    true,
		"videoPath":   "/path/to/recording.mp4",
		"videoOffset": float64(1000), // float64 instead of int64
	}
	mock.GetVideoFrameResult = "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id":    "session-123",
		"event_time_ms": float64(3000),
	}

	_, err := server.handleSessionVideoFrame(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify GetVideoFrame was called with correct video time
	call := mock.GetLastCallByMethod("GetVideoFrame")
	if call.Args[1] != int64(2000) {
		t.Errorf("Expected video time 2000 (3000-1000), got %v", call.Args[1])
	}
}

func TestHandleSessionVideoFrame_NegativeVideoTime(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoResult = map[string]interface{}{
		"hasVideo":    true,
		"videoPath":   "/path/to/recording.mp4",
		"videoOffset": int64(5000), // offset larger than event time
	}
	mock.GetVideoFrameResult = "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id":    "session-123",
		"event_time_ms": float64(2000), // less than offset
	}

	_, err := server.handleSessionVideoFrame(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify video time was clamped to 0
	call := mock.GetLastCallByMethod("GetVideoFrame")
	if call.Args[1] != int64(0) {
		t.Errorf("Expected video time to be clamped to 0, got %v", call.Args[1])
	}
}

func TestHandleSessionVideoFrame_NoVideo(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoResult = map[string]interface{}{
		"hasVideo": false,
	}
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id":    "session-456",
		"event_time_ms": float64(5000),
	}

	result, err := server.handleSessionVideoFrame(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "no video recording") {
		t.Error("Expected 'no video recording' message")
	}

	// GetVideoFrame should not be called
	if mock.WasMethodCalled("GetVideoFrame") {
		t.Error("GetVideoFrame should not be called when there's no video")
	}
}

func TestHandleSessionVideoFrame_MissingSessionId(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"event_time_ms": float64(5000),
	}

	_, err := server.handleSessionVideoFrame(nil, request)
	if err == nil {
		t.Fatal("Expected error for missing session_id")
	}
}

func TestHandleSessionVideoFrame_MissingEventTime(t *testing.T) {
	mock := NewMockGazeApp()
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id": "session-123",
	}

	_, err := server.handleSessionVideoFrame(nil, request)
	if err == nil {
		t.Fatal("Expected error for missing event_time_ms")
	}
}

func TestHandleSessionVideoFrame_CustomWidth(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoResult = map[string]interface{}{
		"hasVideo":  true,
		"videoPath": "/path/to/recording.mp4",
	}
	mock.GetVideoFrameResult = "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id":    "session-123",
		"event_time_ms": float64(1000),
		"width":         float64(1080),
	}

	_, err := server.handleSessionVideoFrame(nil, request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	call := mock.GetLastCallByMethod("GetVideoFrame")
	if call.Args[2] != 1080 {
		t.Errorf("Expected width 1080, got %v", call.Args[2])
	}
}

func TestHandleSessionVideoFrame_GetVideoInfoError(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoError = errors.New("session not found")
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id":    "nonexistent",
		"event_time_ms": float64(1000),
	}

	_, err := server.handleSessionVideoFrame(nil, request)
	if err == nil {
		t.Fatal("Expected error for session not found")
	}
}

func TestHandleSessionVideoFrame_ExtractError(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoResult = map[string]interface{}{
		"hasVideo":  true,
		"videoPath": "/path/to/recording.mp4",
	}
	mock.GetVideoFrameError = errors.New("frame extraction failed")
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id":    "session-123",
		"event_time_ms": float64(1000),
	}

	_, err := server.handleSessionVideoFrame(nil, request)
	if err == nil {
		t.Fatal("Expected error for extraction failure")
	}
}

func TestHandleSessionVideoFrame_EmptyVideoPath(t *testing.T) {
	mock := NewMockGazeApp()
	mock.GetSessionVideoInfoResult = map[string]interface{}{
		"hasVideo":  true,
		"videoPath": "", // empty path
	}
	server := NewMCPServer(mock)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"session_id":    "session-123",
		"event_time_ms": float64(1000),
	}

	_, err := server.handleSessionVideoFrame(nil, request)
	if err == nil {
		t.Fatal("Expected error for empty video path")
	}
}
