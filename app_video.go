package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// ========================================
// Video APIs (exposed to frontend)
// ========================================

var (
	videoService   *VideoService
	videoServiceMu sync.RWMutex
)

// getVideoService returns the video service, initializing if needed
func (a *App) getVideoService() *VideoService {
	videoServiceMu.Lock()
	defer videoServiceMu.Unlock()

	if videoService == nil {
		// Use embedded FFmpeg if available, otherwise fallback to PATH search
		videoService = NewVideoServiceWithPaths(a.ctx, a.dataDir, a.ffmpegPath, a.ffprobePath)
	}
	return videoService
}

// VideoServiceInfo represents video service status
type VideoServiceInfo struct {
	Available   bool   `json:"available"`
	FFmpegPath  string `json:"ffmpegPath,omitempty"`
	FFprobePath string `json:"ffprobePath,omitempty"`
}

// GetVideoServiceInfo returns video service availability info
func (a *App) GetVideoServiceInfo() VideoServiceInfo {
	svc := a.getVideoService()
	return VideoServiceInfo{
		Available:   svc.IsAvailable(),
		FFmpegPath:  svc.ffmpegPath,
		FFprobePath: svc.ffprobePath,
	}
}

// GetVideoMetadata returns metadata for a video file
func (a *App) GetVideoMetadata(videoPath string) (*VideoMetadata, error) {
	svc := a.getVideoService()
	return svc.GetMetadata(videoPath)
}

// GetVideoFrame extracts a single frame from a video
func (a *App) GetVideoFrame(videoPath string, timeMs int64, width int) (string, error) {
	svc := a.getVideoService()
	return svc.ExtractFrameBase64(videoPath, timeMs, width)
}

// GetVideoThumbnails generates thumbnails at regular intervals
func (a *App) GetVideoThumbnails(videoPath string, intervalMs int64, width int) ([]VideoThumbnail, error) {
	svc := a.getVideoService()
	if intervalMs <= 0 {
		intervalMs = 5000 // Default 5 seconds
	}
	if width <= 0 {
		width = 160 // Default thumbnail width
	}
	return svc.GenerateThumbnails(videoPath, intervalMs, width)
}

// GetSessionVideoInfo returns video info for a session
func (a *App) GetSessionVideoInfo(sessionID string) (map[string]interface{}, error) {
	if a.eventStore == nil {
		return nil, fmt.Errorf("event store not initialized")
	}

	// Get session
	session, err := a.eventStore.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.VideoPath == "" {
		return map[string]interface{}{
			"hasVideo": false,
		}, nil
	}

	// Check if video file exists
	if _, err := os.Stat(session.VideoPath); os.IsNotExist(err) {
		return map[string]interface{}{
			"hasVideo":  false,
			"videoPath": session.VideoPath,
			"error":     "Video file not found",
		}, nil
	}

	// Get video metadata
	svc := a.getVideoService()
	meta, err := svc.GetMetadata(session.VideoPath)
	if err != nil {
		return map[string]interface{}{
			"hasVideo":  true,
			"videoPath": session.VideoPath,
			"error":     err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"hasVideo":    true,
		"videoPath":   session.VideoPath,
		"videoOffset": session.VideoOffset,
		"metadata":    meta,
	}, nil
}

// GetVideoFileURL returns a URL that can be used to access the video
// This is needed because browsers can't directly access local files
func (a *App) GetVideoFileURL(videoPath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return "", fmt.Errorf("video file not found: %s", videoPath)
	}

	// Return path with video:// protocol
	// The frontend will handle this through a custom handler
	return "video://" + videoPath, nil
}

// ServeVideoFile serves a video file for streaming
// This is called by the frontend through a special URL
func (a *App) ServeVideoFile(w http.ResponseWriter, r *http.Request, videoPath string) {
	// Open the file
	file, err := os.Open(videoPath)
	if err != nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Cannot read video", http.StatusInternalServerError)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")

	// Handle range requests for video seeking
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		// Parse range header
		if strings.HasPrefix(rangeHeader, "bytes=") {
			rangeSpec := strings.TrimPrefix(rangeHeader, "bytes=")
			parts := strings.Split(rangeSpec, "-")
			if len(parts) == 2 {
				start, _ := strconv.ParseInt(parts[0], 10, 64)
				end := stat.Size() - 1
				if parts[1] != "" {
					end, _ = strconv.ParseInt(parts[1], 10, 64)
				}

				// Seek to start
				_, err = file.Seek(start, 0)
				if err != nil {
					http.Error(w, "Seek error", http.StatusInternalServerError)
					return
				}

				// Set partial content headers
				w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, stat.Size()))
				w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
				w.WriteHeader(http.StatusPartialContent)

				// Copy the requested range
				io.CopyN(w, file, end-start+1)
				return
			}
		}
	}

	// Full file response
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	io.Copy(w, file)
}

// ReadVideoFileAsDataURL reads a video file and returns it as a data URL
// Warning: This loads the entire video into memory, use only for small videos
func (a *App) ReadVideoFileAsDataURL(videoPath string) (string, error) {
	// Check file size first
	stat, err := os.Stat(videoPath)
	if err != nil {
		return "", err
	}

	// Limit to 50MB
	if stat.Size() > 50*1024*1024 {
		return "", fmt.Errorf("video file too large for data URL (max 50MB)")
	}

	data, err := os.ReadFile(videoPath)
	if err != nil {
		return "", err
	}

	// Determine MIME type
	mimeType := "video/mp4"
	ext := strings.ToLower(filepath.Ext(videoPath))
	switch ext {
	case ".webm":
		mimeType = "video/webm"
	case ".ogg", ".ogv":
		mimeType = "video/ogg"
	case ".mov":
		mimeType = "video/quicktime"
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
}

// GetRecordingsDir returns the recordings directory path
func (a *App) GetRecordingsDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	return filepath.Join(configDir, "Gaze", "recordings")
}

// ListRecordings lists all video recordings
func (a *App) ListRecordings() ([]map[string]interface{}, error) {
	recordingsDir := a.GetRecordingsDir()

	entries, err := os.ReadDir(recordingsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]interface{}{}, nil
		}
		return nil, err
	}

	var recordings []map[string]interface{}
	svc := a.getVideoService()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".mp4" && ext != ".webm" && ext != ".mkv" {
			continue
		}

		path := filepath.Join(recordingsDir, entry.Name())
		info, _ := entry.Info()

		recording := map[string]interface{}{
			"name":     entry.Name(),
			"path":     path,
			"size":     info.Size(),
			"modified": info.ModTime().Unix(),
		}

		// Try to get video metadata
		if meta, err := svc.GetMetadata(path); err == nil {
			recording["duration"] = meta.DurationMs
			recording["width"] = meta.Width
			recording["height"] = meta.Height
		}

		recordings = append(recordings, recording)
	}

	return recordings, nil
}
