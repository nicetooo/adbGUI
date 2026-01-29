package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // register JPEG decoder for image.DecodeConfig
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ========================================
// Video Service - 视频处理服务
// ========================================

// VideoService handles video processing operations
type VideoService struct {
	ctx         context.Context
	ffmpegPath  string
	ffprobePath string
	cacheDir    string
	cacheMu     sync.RWMutex
	cache       map[string]*VideoMetadata
}

// VideoMetadata contains video file metadata
type VideoMetadata struct {
	Path          string  `json:"path"`
	Duration      float64 `json:"duration"`   // Duration in seconds
	DurationMs    int64   `json:"durationMs"` // Duration in milliseconds
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	FrameRate     float64 `json:"frameRate"`
	Codec         string  `json:"codec"`
	BitRate       int64   `json:"bitRate"`
	TotalFrames   int64   `json:"totalFrames"`
	ThumbnailPath string  `json:"thumbnailPath,omitempty"`
}

// VideoFrame represents an extracted video frame
type VideoFrame struct {
	TimeMs int64  `json:"timeMs"` // Time position in milliseconds
	Data   []byte `json:"data"`   // Raw image data (JPEG)
	Base64 string `json:"base64"` // Base64 encoded image
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// VideoThumbnail represents a thumbnail at a specific time
type VideoThumbnail struct {
	TimeMs int64  `json:"timeMs"`
	Base64 string `json:"base64"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// NewVideoServiceWithPaths creates a video service with pre-configured FFmpeg paths
func NewVideoServiceWithPaths(ctx context.Context, dataDir string, ffmpegPath, ffprobePath string) *VideoService {
	cacheDir := filepath.Join(dataDir, "video_cache")
	_ = os.MkdirAll(cacheDir, 0755)

	svc := &VideoService{
		ctx:         ctx,
		cacheDir:    cacheDir,
		cache:       make(map[string]*VideoMetadata),
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
	}

	// If paths not provided, try to find in PATH
	if svc.ffmpegPath == "" || svc.ffprobePath == "" {
		svc.findFFmpeg()
	} else {
		LogInfo("video_service").Str("ffmpeg", svc.ffmpegPath).Str("ffprobe", svc.ffprobePath).Msg("Using embedded ffmpeg binaries")
	}

	return svc
}

// SetFFmpegPaths sets the FFmpeg and FFprobe paths
func (s *VideoService) SetFFmpegPaths(ffmpegPath, ffprobePath string) {
	if ffmpegPath != "" {
		s.ffmpegPath = ffmpegPath
		LogInfo("video_service").Str("ffmpeg", ffmpegPath).Msg("Set ffmpeg path")
	}
	if ffprobePath != "" {
		s.ffprobePath = ffprobePath
		LogInfo("video_service").Str("ffprobe", ffprobePath).Msg("Set ffprobe path")
	}
}

// findFFmpeg locates ffmpeg and ffprobe binaries
func (s *VideoService) findFFmpeg() {
	// Try to find ffmpeg in PATH if not already set
	if s.ffmpegPath == "" {
		if path, err := exec.LookPath("ffmpeg"); err == nil {
			s.ffmpegPath = path
		}
	}
	if s.ffprobePath == "" {
		if path, err := exec.LookPath("ffprobe"); err == nil {
			s.ffprobePath = path
		}
	}

	// Log status
	if s.ffmpegPath != "" {
		LogInfo("video_service").Str("ffmpeg", s.ffmpegPath).Msg("Found ffmpeg")
	} else {
		LogWarn("video_service").Msg("ffmpeg not found in PATH")
	}
	if s.ffprobePath != "" {
		LogInfo("video_service").Str("ffprobe", s.ffprobePath).Msg("Found ffprobe")
	} else {
		LogWarn("video_service").Msg("ffprobe not found in PATH")
	}
}

// IsAvailable checks if video processing is available
func (s *VideoService) IsAvailable() bool {
	return s.ffmpegPath != "" && s.ffprobePath != ""
}

// GetMetadata retrieves video metadata
func (s *VideoService) GetMetadata(videoPath string) (*VideoMetadata, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("ffmpeg/ffprobe not available")
	}

	// Check cache
	s.cacheMu.RLock()
	if meta, ok := s.cache[videoPath]; ok {
		s.cacheMu.RUnlock()
		return meta, nil
	}
	s.cacheMu.RUnlock()

	// Check if file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("video file not found: %s", videoPath)
	}

	// Run ffprobe to get metadata
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	// Parse ffprobe output
	var probeResult struct {
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
		Streams []struct {
			CodecType  string `json:"codec_type"`
			CodecName  string `json:"codec_name"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
			NbFrames   string `json:"nb_frames"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &probeResult); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	meta := &VideoMetadata{
		Path: videoPath,
	}

	// Parse duration
	if probeResult.Format.Duration != "" {
		if duration, err := strconv.ParseFloat(probeResult.Format.Duration, 64); err == nil {
			meta.Duration = duration
			meta.DurationMs = int64(duration * 1000)
		}
	}

	// Parse bitrate
	if probeResult.Format.BitRate != "" {
		if bitRate, err := strconv.ParseInt(probeResult.Format.BitRate, 10, 64); err == nil {
			meta.BitRate = bitRate
		}
	}

	// Find video stream
	for _, stream := range probeResult.Streams {
		if stream.CodecType == "video" {
			meta.Width = stream.Width
			meta.Height = stream.Height
			meta.Codec = stream.CodecName

			// Parse frame rate (e.g., "30/1" or "29.97")
			if stream.RFrameRate != "" {
				parts := strings.Split(stream.RFrameRate, "/")
				if len(parts) == 2 {
					num, _ := strconv.ParseFloat(parts[0], 64)
					den, _ := strconv.ParseFloat(parts[1], 64)
					if den > 0 {
						meta.FrameRate = num / den
					}
				} else {
					meta.FrameRate, _ = strconv.ParseFloat(stream.RFrameRate, 64)
				}
			}

			// Parse total frames
			if stream.NbFrames != "" {
				meta.TotalFrames, _ = strconv.ParseInt(stream.NbFrames, 10, 64)
			} else if meta.Duration > 0 && meta.FrameRate > 0 {
				meta.TotalFrames = int64(meta.Duration * meta.FrameRate)
			}
			break
		}
	}

	// Cache the result
	s.cacheMu.Lock()
	s.cache[videoPath] = meta
	s.cacheMu.Unlock()

	return meta, nil
}

// ExtractFrame extracts a single frame at the specified time
func (s *VideoService) ExtractFrame(videoPath string, timeMs int64, width int) (*VideoFrame, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("ffmpeg not available")
	}

	// Build ffmpeg command
	timeStr := formatFFmpegTime(timeMs)

	args := []string{
		"-ss", timeStr,
		"-i", videoPath,
		"-vframes", "1",
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
	}

	// Add scaling if width is specified
	if width > 0 {
		args = append(args, "-vf", fmt.Sprintf("scale=%d:-1", width))
	}

	args = append(args, "-")

	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.ffmpegPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %w, stderr: %s", err, stderr.String())
	}

	frameData := stdout.Bytes()
	if len(frameData) == 0 {
		return nil, fmt.Errorf("no frame data extracted")
	}

	// Get image dimensions
	img, _, err := image.DecodeConfig(bytes.NewReader(frameData))
	if err != nil {
		// Still return the frame even if we can't decode dimensions
		return &VideoFrame{
			TimeMs: timeMs,
			Data:   frameData,
			Base64: base64.StdEncoding.EncodeToString(frameData),
		}, nil
	}

	return &VideoFrame{
		TimeMs: timeMs,
		Data:   frameData,
		Base64: base64.StdEncoding.EncodeToString(frameData),
		Width:  img.Width,
		Height: img.Height,
	}, nil
}

// ExtractFrameBase64 extracts a frame and returns base64 string (for frontend)
func (s *VideoService) ExtractFrameBase64(videoPath string, timeMs int64, width int) (string, error) {
	frame, err := s.ExtractFrame(videoPath, timeMs, width)
	if err != nil {
		return "", err
	}
	return "data:image/jpeg;base64," + frame.Base64, nil
}

// GenerateThumbnails generates thumbnails at regular intervals
func (s *VideoService) GenerateThumbnails(videoPath string, intervalMs int64, width int) ([]VideoThumbnail, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("ffmpeg not available")
	}

	// Get video metadata first
	meta, err := s.GetMetadata(videoPath)
	if err != nil {
		return nil, err
	}

	var thumbnails []VideoThumbnail
	for timeMs := int64(0); timeMs < meta.DurationMs; timeMs += intervalMs {
		frame, err := s.ExtractFrame(videoPath, timeMs, width)
		if err != nil {
			continue // Skip failed frames
		}

		thumbnails = append(thumbnails, VideoThumbnail{
			TimeMs: timeMs,
			Base64: "data:image/jpeg;base64," + frame.Base64,
			Width:  frame.Width,
			Height: frame.Height,
		})
	}

	return thumbnails, nil
}

// formatFFmpegTime formats milliseconds to ffmpeg time format (HH:MM:SS.mmm)
func formatFFmpegTime(ms int64) string {
	seconds := float64(ms) / 1000.0
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := seconds - float64(hours*3600) - float64(minutes*60)
	return fmt.Sprintf("%02d:%02d:%06.3f", hours, minutes, secs)
}
