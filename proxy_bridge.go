package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"Gaze/proxy"
)

// proxyDeviceId tracks which device the proxy is monitoring for session events
var (
	proxyDeviceId string
	proxyDeviceMu sync.RWMutex
)

// SetProxyDevice sets which device the proxy is monitoring (for session events)
func (a *App) SetProxyDevice(deviceId string) {
	proxyDeviceMu.Lock()
	proxyDeviceId = deviceId
	proxyDeviceMu.Unlock()
}

// GetProxyDevice returns the currently monitored device
func (a *App) GetProxyDevice() string {
	proxyDeviceMu.RLock()
	defer proxyDeviceMu.RUnlock()
	return proxyDeviceId
}

// StartProxy starts the internal HTTP/HTTPS proxy
func (a *App) StartProxy(port int) (string, error) {
	err := proxy.GetProxy().Start(port, func(req proxy.RequestLog) {
		// All network events go through Session (unified event source)
		// Skip partial updates to avoid duplicates - only emit completed requests
		// Filter out pending requests (StatusCode 0) to avoid duplicates.
		// We only want to emit the final event when the response is complete.
		if req.StatusCode == 0 {
			return
		}

		proxyDeviceMu.RLock()
		deviceId := proxyDeviceId
		proxyDeviceMu.RUnlock()

		if deviceId == "" {
			return
		}

		// Determine level based on status code
		level := "info"
		if req.StatusCode >= 400 && req.StatusCode < 500 {
			level = "warn"
		} else if req.StatusCode >= 500 {
			level = "error"
		}

		title := fmt.Sprintf("%s %s → %d", req.Method, req.URL, req.StatusCode)
		if len(title) > 100 {
			title = title[:97] + "..."
		}

		// Calculate duration from ID (SessionId-TimestampNano)
		var durationMs int64
		parts := strings.Split(req.Id, "-")
		if len(parts) >= 2 {
			if startNano, err := strconv.ParseInt(parts[len(parts)-1], 10, 64); err == nil {
				durationMs = (time.Now().UnixNano() - startNano) / 1e6
			}
		}

		a.EmitSessionEvent(deviceId, "network_request", "network", level, title,
			map[string]interface{}{
				"id":              req.Id,
				"method":          req.Method,
				"url":             req.URL,
				"statusCode":      req.StatusCode,
				"contentType":     req.ContentType,
				"bodySize":        req.BodySize,
				"duration":        durationMs,
				"isHttps":         req.IsHTTPS,
				"isWs":            req.IsWs,
				"requestHeaders":  req.Headers,
				"requestBody":     req.Body,
				"responseHeaders": req.RespHeaders,
				"responseBody":    req.RespBody,
			})
	})
	if err != nil {
		return "", err
	}
	return "Proxy started successfully", nil
}

// StopProxy stops the internal proxy
func (a *App) StopProxy() (string, error) {
	err := proxy.GetProxy().Stop()
	if err != nil {
		return "", err
	}
	return "Proxy stopped successfully", nil
}

// GetProxyStatus returns true if the proxy is running
func (a *App) GetProxyStatus() bool {
	return proxy.GetProxy().IsRunning()
}

// SetProxyLimit sets the upload and download speed limits for the proxy server (bytes per second)
func (a *App) SetProxyLimit(uploadSpeed, downloadSpeed int) {
	proxy.GetProxy().SetLimits(uploadSpeed, downloadSpeed)
}

// SetProxyWSEnabled enables or disables WebSocket support
func (a *App) SetProxyWSEnabled(enabled bool) {
	proxy.GetProxy().SetWSEnabled(enabled)
}

// SetProxyMITM enables or disables HTTPS Decryption (MITM)
func (a *App) SetProxyMITM(enabled bool) {
	proxy.GetProxy().SetProxyMITM(enabled)
}

// SetMITMBypassPatterns sets the keywords/domains to bypass MITM
func (a *App) SetMITMBypassPatterns(patterns []string) {
	proxy.GetProxy().SetMITMBypassPatterns(patterns)
}

// GetMITMBypassPatterns returns the current bypass patterns
func (a *App) GetMITMBypassPatterns() []string {
	return proxy.GetProxy().GetMITMBypassPatterns()
}

// GetProxySettings returns the current proxy settings
func (a *App) GetProxySettings() map[string]interface{} {
	return map[string]interface{}{
		"wsEnabled":      proxy.GetProxy().IsWSEnabled(),
		"mitmEnabled":    proxy.GetProxy().IsMITMEnabled(),
		"bypassPatterns": proxy.GetProxy().GetMITMBypassPatterns(),
	}
}

// InstallProxyCert pushes the generated CA certificate to the device
func (a *App) InstallProxyCert(deviceId string) (string, error) {
	certPath := proxy.GetProxy().GetCertPath()
	if certPath == "" {
		return "", nil
	}

	dest := "/sdcard/Download/Gaze-CA.crt"

	cmd := exec.Command(a.adbPath, "-s", deviceId, "push", certPath, dest)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", err
	} else {
		_ = out
	}

	return dest, nil
}

// SetProxyLatency sets the artificial latency in milliseconds
func (a *App) SetProxyLatency(latencyMs int) {
	proxy.GetProxy().SetLatency(latencyMs)
}

// isStaticResource checks if a request is for a static resource (image, css, js, font, etc.)
// Returns true if it should be filtered out from the timeline
func isStaticResource(url, contentType string) bool {
	// Check by content type first (most reliable)
	if contentType != "" {
		staticContentTypes := []string{
			"image/",
			"text/css",
			"text/javascript",
			"application/javascript",
			"application/x-javascript",
			"font/",
			"application/font",
			"application/x-font",
			"video/",
			"audio/",
		}

		// Normalize content type (remove charset, etc.)
		ctLower := contentType
		for idx := 0; idx < len(contentType); idx++ {
			if contentType[idx] == ';' {
				ctLower = contentType[:idx]
				break
			}
		}

		for _, ct := range staticContentTypes {
			if len(ctLower) >= len(ct) {
				match := true
				for j := 0; j < len(ct); j++ {
					c1 := ctLower[j]
					c2 := ct[j]
					if c1 >= 'A' && c1 <= 'Z' {
						c1 += 32
					}
					if c1 != byte(c2) {
						match = false
						break
					}
				}
				if match && (len(ctLower) == len(ct) || ctLower[len(ct)] == '/' || ctLower[len(ct)] == ';') {
					return true
				}
			}
		}
	}

	// Check by URL extension as fallback
	// Extract path part (before ? or #)
	pathEnd := len(url)
	for i := 0; i < len(url); i++ {
		if url[i] == '?' || url[i] == '#' {
			pathEnd = i
			break
		}
	}

	if pathEnd == 0 {
		return false
	}

	path := url[:pathEnd]

	// Find the last dot in the path
	lastDot := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			lastDot = i
			break
		}
		if path[i] == '/' {
			// No extension found before path separator
			break
		}
	}

	if lastDot == -1 || lastDot == len(path)-1 {
		return false
	}

	// Extract extension (including the dot)
	ext := path[lastDot:]

	// Convert to lowercase
	extLower := ""
	for i := 0; i < len(ext); i++ {
		c := ext[i]
		if c >= 'A' && c <= 'Z' {
			extLower += string(c + 32)
		} else {
			extLower += string(c)
		}
	}

	// Check against known static extensions
	staticExtensions := []string{
		".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".ico", ".bmp",
		".css", ".js", ".mjs",
		".woff", ".woff2", ".ttf", ".eot", ".otf",
		".mp4", ".webm", ".ogg", ".mp3", ".wav",
		".pdf", ".zip", ".tar", ".gz",
	}

	for _, staticExt := range staticExtensions {
		if extLower == staticExt {
			return true
		}
	}

	return false
}
