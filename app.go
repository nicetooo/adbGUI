package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"Gaze/pkg/cache"
)

// Binaries are embedded in platform-specific files (bin_*.go) and bin_common.go

// App struct
type App struct {
	ctx          context.Context
	adbPath      string
	scrcpyPath   string
	serverPath   string
	aaptPath     string
	ffmpegPath   string
	ffprobePath  string
	logcatCmd    *exec.Cmd
	logcatCancel context.CancelFunc

	// Generic mutex for shared state
	mu sync.Mutex

	// Services
	cacheService *cache.Service

	// History (still managed by device.go)
	historyMu sync.Mutex

	// Scrcpy process management
	scrcpyCmds      map[string]*exec.Cmd
	scrcpyRecordCmd map[string]*exec.Cmd
	scrcpyMu        sync.Mutex

	// File open process management
	openFileCmds map[string]*exec.Cmd
	openFileMu   sync.Mutex

	// Wireless Server
	httpServer *http.Server
	localAddr  string

	version string

	// Runtime logs
	runtimeLogs []string
	logsMu      sync.Mutex

	// Device tracking
	lastDevCount int
	idToSerial   map[string]string
	idToSerialMu sync.RWMutex

	// Wireless stability
	reconnectCooldown map[string]time.Time
	reconnectMu       sync.Mutex

	// Device monitor
	deviceMonitorCancel context.CancelFunc
	deviceMonitorMu     sync.Mutex

	// Session device monitors (per device)
	sessionMonitors   map[string]*DeviceMonitor
	sessionMonitorsMu sync.Mutex

	// Event System (new)
	eventStore      *EventStore
	eventPipeline   *EventPipeline
	assertionEngine *AssertionEngine
	eventSystemMu   sync.RWMutex
	dataDir         string

	// AI Service
	aiService   *AIService
	aiConfigMgr *AIConfigManager
	aiServiceMu sync.RWMutex
}

// NewApp creates a new App instance
func NewApp(version string) *App {
	app := &App{
		scrcpyCmds:        make(map[string]*exec.Cmd),
		scrcpyRecordCmd:   make(map[string]*exec.Cmd),
		openFileCmds:      make(map[string]*exec.Cmd),
		idToSerial:        make(map[string]string),
		reconnectCooldown: make(map[string]time.Time),
		sessionMonitors:   make(map[string]*DeviceMonitor),
		version:           version,
	}
	app.initCacheService()
	return app
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.setupBinaries()
	a.initEventSystem() // Initialize new event system
	a.StartDeviceMonitor()
	a.StartBatchSync() // Start session event batch sync (legacy, for compatibility)
	a.LoadMockRules()  // Load saved mock rules
}

// Shutdown is called when the application is closing
func (a *App) Shutdown(ctx context.Context) {
	LogAppState(StateShuttingDown, map[string]interface{}{
		"reason": "application_close",
	})

	// Stop proxy and clean up device settings to prevent network issues
	if a.GetProxyStatus() {
		a.StopProxy()
	}

	a.StopBatchSync()       // Stop session event batch sync (legacy)
	a.shutdownEventSystem() // Shutdown new event system
	a.scrcpyMu.Lock()
	for id, cmd := range a.scrcpyCmds {
		if cmd.Process != nil {
			os.Stderr.WriteString(fmt.Sprintf("\n[SHUTDOWN] Killing mirroring for %s\n", id))
			_ = cmd.Process.Kill()
		}
	}
	for id, cmd := range a.scrcpyRecordCmd {
		if cmd.Process != nil {
			os.Stderr.WriteString(fmt.Sprintf("\n[SHUTDOWN] Killing recording for %s\n", id))
			_ = cmd.Process.Kill()
		}
	}
	a.scrcpyMu.Unlock()
	a.StopLogcat()
	a.StopDeviceMonitor()

	LogAppState(StateStopped, nil)
	CloseLogger()
}

// GetAppVersion returns the application version
func (a *App) GetAppVersion() string {
	return a.version
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Log adds a message to the runtime logs (legacy method, forwards to zerolog)
func (a *App) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	// Forward to structured logger
	LogInfo("app").Msg(msg)

	// Keep legacy runtime logs for frontend display
	a.logsMu.Lock()
	defer a.logsMu.Unlock()
	timestampedMsg := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
	a.runtimeLogs = append(a.runtimeLogs, timestampedMsg)
	if len(a.runtimeLogs) > 1000 {
		a.runtimeLogs = a.runtimeLogs[len(a.runtimeLogs)-1000:]
	}
}

// GetBackendLogs returns the captured backend logs
func (a *App) GetBackendLogs() []string {
	a.logsMu.Lock()
	defer a.logsMu.Unlock()
	logs := make([]string, len(a.runtimeLogs))
	copy(logs, a.runtimeLogs)
	return logs
}

// updateLastActive updates the last active timestamp for a device
func (a *App) updateLastActive(deviceId string) {
	if deviceId == "" || a.cacheService == nil {
		return
	}

	serial := deviceId
	a.idToSerialMu.RLock()
	if s, ok := a.idToSerial[deviceId]; ok {
		serial = s
	}
	a.idToSerialMu.RUnlock()

	a.cacheService.SetLastActive(serial, time.Now().Unix())
	go a.saveSettings()
}

// Initialization functions

func (a *App) initCacheService() {
	svc, err := cache.New(cache.Config{
		LogFunc: a.Log,
	})
	if err != nil {
		a.Log("Error initializing cache service: %v", err)
		return
	}
	a.cacheService = svc
}

// saveSettings delegates to cache service
func (a *App) saveSettings() {
	if a.cacheService != nil {
		_ = a.cacheService.SaveSettings()
	}
}

// saveCache delegates to cache service
func (a *App) saveCache() {
	if a.cacheService != nil {
		_ = a.cacheService.SaveCache()
	}
}

func (a *App) setupBinaries() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	appBinDir := filepath.Join(configDir, "Gaze", "bin")
	_ = os.MkdirAll(appBinDir, 0755)

	extract := func(name string, data []byte) string {
		if len(data) == 0 {
			return ""
		}

		path := filepath.Join(appBinDir, name)
		if runtime.GOOS == "windows" && !strings.HasSuffix(name, ".exe") && name != "scrcpy-server" {
			path += ".exe"
		}

		info, err := os.Stat(path)
		if err != nil || info.Size() != int64(len(data)) {
			err = os.WriteFile(path, data, 0755)
			if err != nil {
				fmt.Printf("Error extracting %s: %v\n", name, err)
			}
		}

		if runtime.GOOS != "windows" {
			_ = os.Chmod(path, 0755)
			if runtime.GOOS == "darwin" {
				_ = exec.Command("xattr", "-d", "com.apple.quarantine", path).Run()
			}
		}

		return path
	}

	// Prefer system ADB if available
	if path, err := exec.LookPath("adb"); err == nil {
		a.adbPath = path
		fmt.Printf("Using system adb found in PATH: %s\n", a.adbPath)
	} else {
		a.adbPath = extract("adb", adbBinary)
		if a.adbPath != "" {
			fmt.Printf("Using bundled adb at: %s\n", a.adbPath)
		}
	}

	a.scrcpyPath = extract("scrcpy", scrcpyBinary)
	a.serverPath = extract("scrcpy-server", scrcpyServerBinary)

	if len(aaptBinary) > 0 {
		a.aaptPath = extract("aapt", aaptBinary)
		fmt.Printf("AAPT setup at: %s\n", a.aaptPath)
	}

	// Setup FFmpeg and FFprobe
	if len(ffmpegBinary) > 0 {
		a.ffmpegPath = extract("ffmpeg", ffmpegBinary)
		fmt.Printf("FFmpeg setup at: %s\n", a.ffmpegPath)
	}
	if len(ffprobeBinary) > 0 {
		a.ffprobePath = extract("ffprobe", ffprobeBinary)
		fmt.Printf("FFprobe setup at: %s\n", a.ffprobePath)
	}

	a.Log("Binaries setup at: %s", appBinDir)
	a.Log("Final ADB path: %s", a.adbPath)
}

// Command helper functions

// newAdbCommand creates an exec.Cmd with a clean environment to avoid proxy issues
func (a *App) newAdbCommand(ctx context.Context, args ...string) *exec.Cmd {
	var cmd *exec.Cmd
	if ctx != nil {
		cmd = exec.CommandContext(ctx, a.adbPath, args...)
	} else {
		cmd = exec.Command(a.adbPath, args...)
	}

	env := os.Environ()
	newEnv := make([]string, 0, len(env))
	proxyVars := []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "all_proxy", "no_proxy"}

	for _, e := range env {
		isProxy := false
		for _, v := range proxyVars {
			if strings.HasPrefix(e, v+"=") {
				isProxy = true
				break
			}
		}
		if !isProxy {
			newEnv = append(newEnv, e)
		}
	}
	cmd.Env = newEnv
	return cmd
}

// newScrcpyCommand creates an exec.Cmd for scrcpy with a clean environment
func (a *App) newScrcpyCommand(args ...string) *exec.Cmd {
	return a.newScrcpyCommandContext(context.Background(), args...)
}

func (a *App) newScrcpyCommandContext(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, a.scrcpyPath, args...)

	env := os.Environ()
	newEnv := make([]string, 0, len(env))
	proxyVars := []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "all_proxy", "no_proxy"}

	for _, e := range env {
		isProxy := false
		for _, v := range proxyVars {
			if strings.HasPrefix(e, v+"=") {
				isProxy = true
				break
			}
		}
		if !isProxy {
			newEnv = append(newEnv, e)
		}
	}

	newEnv = append(newEnv,
		"SCRCPY_SERVER_PATH="+a.serverPath,
		"ADB="+a.adbPath,
	)

	cmd.Env = newEnv
	return cmd
}

// ========================================
// Event System
// ========================================

// initEventSystem initializes the new event storage and pipeline
func (a *App) initEventSystem() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	a.dataDir = filepath.Join(configDir, "Gaze", "data")

	// Create event store
	store, err := NewEventStore(a.dataDir)
	if err != nil {
		a.Log("Failed to initialize event store: %v", err)
		return
	}
	a.eventStore = store

	// Create event pipeline
	a.eventPipeline = NewEventPipeline(context.Background(), a.ctx, store)
	a.eventPipeline.Start()

	// Create assertion engine
	a.assertionEngine = NewAssertionEngine(a, store, a.eventPipeline)

	// Initialize AI service
	a.initAIService()

	a.Log("Event system initialized at: %s", a.dataDir)
}

// initAIService initializes the AI service
func (a *App) initAIService() {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	// Initialize config manager
	a.aiConfigMgr = NewAIConfigManager(a.dataDir)
	if err := a.aiConfigMgr.Load(); err != nil {
		a.Log("Failed to load AI config: %v", err)
	}

	config := a.aiConfigMgr.GetConfig()

	// Initialize AI service
	aiService, err := NewAIService(a.ctx, config)
	if err != nil {
		a.Log("Failed to initialize AI service: %v", err)
		return
	}
	a.aiService = aiService

	a.Log("AI service initialized (status: %s)", aiService.GetStatus())
}

// shutdownEventSystem shuts down the event system
func (a *App) shutdownEventSystem() {
	// Shutdown AI service first
	a.shutdownAIService()

	a.eventSystemMu.Lock()
	defer a.eventSystemMu.Unlock()

	if a.eventPipeline != nil {
		a.eventPipeline.Stop()
		a.eventPipeline = nil
	}
	if a.eventStore != nil {
		if err := a.eventStore.Close(); err != nil {
			a.Log("Error closing event store: %v", err)
		}
		a.eventStore = nil
	}
}

// shutdownAIService shuts down the AI service
func (a *App) shutdownAIService() {
	a.aiServiceMu.Lock()
	defer a.aiServiceMu.Unlock()

	if a.aiService != nil {
		a.aiService.Shutdown()
		a.aiService = nil
	}
}

// ========================================
// Event System APIs (exposed to frontend)
// ========================================

// EmitEvent emits a new event through the pipeline
func (a *App) EmitEvent(deviceID string, source string, eventType string, level string, title string, data interface{}) {
	if a.eventPipeline == nil {
		return
	}
	a.eventPipeline.EmitRaw(
		deviceID,
		EventSource(source),
		eventType,
		EventLevel(level),
		title,
		data,
	)
}

// StartNewSession starts a new session for a device (legacy API, no config)
func (a *App) StartNewSession(deviceID, sessionType, name string) string {
	log.Printf("[StartNewSession] Called with deviceID=%s, type=%s, name=%s", deviceID, sessionType, name)
	if a.eventPipeline == nil {
		log.Printf("[StartNewSession] ERROR: eventPipeline is nil!")
		return ""
	}
	sessionID := a.eventPipeline.StartSession(deviceID, sessionType, name, nil)
	log.Printf("[StartNewSession] Created session: %s", sessionID)
	return sessionID
}

// StartSessionWithConfig starts a new session with configuration (logcat, recording, proxy)
func (a *App) StartSessionWithConfig(deviceID, name string, config SessionConfig) string {
	log.Printf("[StartSessionWithConfig] Called with deviceID=%s, name=%s, config=%+v", deviceID, name, config)
	if a.eventPipeline == nil {
		log.Printf("[StartSessionWithConfig] ERROR: eventPipeline is nil!")
		return ""
	}

	// 创建 session
	sessionID := a.eventPipeline.StartSession(deviceID, "manual", name, &config)
	log.Printf("[StartSessionWithConfig] Created session: %s", sessionID)

	// 根据配置启动相关功能
	if config.Logcat.Enabled {
		log.Printf("[StartSessionWithConfig] Starting logcat for package: %s", config.Logcat.PackageName)
		go func() {
			if err := a.StartLogcat(deviceID, config.Logcat.PackageName,
				config.Logcat.PreFilter, false,
				config.Logcat.ExcludeFilter, false); err != nil {
				log.Printf("[StartSessionWithConfig] Failed to start logcat: %v", err)
			}
		}()
	}

	if config.Recording.Enabled {
		log.Printf("[StartSessionWithConfig] Starting headless recording")
		go func() {
			// 生成录屏文件路径
			recordPath := a.generateRecordPath(deviceID, sessionID)
			scrcpyConfig := ScrcpyConfig{
				RecordPath: recordPath,
			}
			// 根据质量设置参数
			switch config.Recording.Quality {
			case "low":
				scrcpyConfig.MaxSize = 480
				scrcpyConfig.BitRate = 2000000
			case "high":
				scrcpyConfig.MaxSize = 1080
				scrcpyConfig.BitRate = 8000000
			default: // medium
				scrcpyConfig.MaxSize = 720
				scrcpyConfig.BitRate = 4000000
			}
			if err := a.StartRecording(deviceID, scrcpyConfig); err != nil {
				log.Printf("[StartSessionWithConfig] Failed to start recording: %v", err)
			} else {
				// 更新 session 的 videoPath
				a.updateSessionVideoPath(sessionID, recordPath)
			}
		}()
	}

	if config.Proxy.Enabled {
		log.Printf("[StartSessionWithConfig] Starting proxy on port: %d", config.Proxy.Port)
		go func() {
			port := config.Proxy.Port
			if port == 0 {
				port = 8080
			}
			a.SetProxyMITM(config.Proxy.MitmEnabled)
			if _, err := a.StartProxy(port); err != nil {
				log.Printf("[StartSessionWithConfig] Failed to start proxy: %v", err)
			} else {
				a.SetProxyDevice(deviceID)
			}
		}()
	}

	if config.Monitor.Enabled {
		log.Printf("[StartSessionWithConfig] Starting device monitor")
		a.sessionMonitorsMu.Lock()
		// Stop existing monitor for this device if any
		if existing := a.sessionMonitors[deviceID]; existing != nil {
			existing.Stop()
		}
		monitor := NewDeviceMonitor(a, deviceID)
		a.sessionMonitors[deviceID] = monitor
		a.sessionMonitorsMu.Unlock()
		monitor.Start()
	}

	return sessionID
}

// generateRecordPath generates a unique recording path for a session
func (a *App) generateRecordPath(deviceID, sessionID string) string {
	homeDir, _ := os.UserHomeDir()
	recordDir := filepath.Join(homeDir, ".adbGUI", "recordings")
	os.MkdirAll(recordDir, 0755)
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return filepath.Join(recordDir, fmt.Sprintf("%s_%s.mp4", timestamp, sessionID[:8]))
}

// updateSessionVideoPath updates the video path for a session
func (a *App) updateSessionVideoPath(sessionID, videoPath string) {
	if a.eventPipeline == nil {
		return
	}
	a.eventPipeline.UpdateSessionVideoPath(sessionID, videoPath)
}

// EndActiveSession ends the current session for a device
func (a *App) EndActiveSession(sessionID, status string) {
	if a.eventPipeline == nil {
		return
	}

	// 获取 session 配置，自动停止相关功能
	session := a.eventPipeline.GetSession(sessionID)
	if session != nil {
		if session.Config.Logcat.Enabled {
			log.Printf("[EndActiveSession] Stopping logcat")
			a.StopLogcat()
		}
		if session.Config.Recording.Enabled {
			log.Printf("[EndActiveSession] Stopping recording")
			a.StopRecording(session.DeviceID)
		}
		if session.Config.Proxy.Enabled {
			log.Printf("[EndActiveSession] Stopping proxy")
			a.StopProxy()
		}
		if session.Config.Monitor.Enabled {
			log.Printf("[EndActiveSession] Stopping device monitor")
			a.sessionMonitorsMu.Lock()
			if monitor := a.sessionMonitors[session.DeviceID]; monitor != nil {
				monitor.Stop()
				delete(a.sessionMonitors, session.DeviceID)
			}
			a.sessionMonitorsMu.Unlock()
		}
	}

	a.eventPipeline.EndSession(sessionID, status)
}

// GetDeviceActiveSession returns the active session for a device
func (a *App) GetDeviceActiveSession(deviceID string) *DeviceSession {
	if a.eventPipeline == nil {
		return nil
	}
	return a.eventPipeline.GetActiveSession(deviceID)
}

// GetInstalledPackages returns a list of installed packages on the device
func (a *App) GetInstalledPackages(deviceID string, thirdPartyOnly bool) ([]string, error) {
	args := []string{"-s", deviceID, "shell", "pm", "list", "packages"}
	if thirdPartyOnly {
		args = append(args, "-3") // Only third-party (user-installed) apps
	}

	cmd := exec.Command(a.adbPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %v", err)
	}

	var packages []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package:") {
			pkg := strings.TrimPrefix(line, "package:")
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// ========================================
// Persistent Log API Methods (exposed to frontend)
// ========================================

// GetLogFilePath returns the current log file path
func (a *App) GetLogFilePath() string {
	return GetLogFilePath()
}

// GetLogDir returns the log directory path
func (a *App) GetLogDir() string {
	return GetLogDir()
}

// ListLogFiles returns all log files
func (a *App) ListLogFiles() ([]string, error) {
	return ListLogFiles()
}

// ReadRecentLogs reads the most recent log lines
func (a *App) ReadRecentLogs(lines int) ([]string, error) {
	return ReadRecentLogs(lines)
}

// OpenLogDir opens the log directory in the system file manager
func (a *App) OpenLogDir() error {
	logDir := GetLogDir()
	if logDir == "" {
		return fmt.Errorf("log directory not available")
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", logDir)
	case "windows":
		cmd = exec.Command("explorer", logDir)
	default:
		cmd = exec.Command("xdg-open", logDir)
	}
	return cmd.Start()
}
