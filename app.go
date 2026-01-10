package main

import (
	"context"
	"encoding/json"
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
)

// Binaries are embedded in platform-specific files (bin_*.go) and bin_common.go

// App struct
type App struct {
	ctx          context.Context
	adbPath      string
	scrcpyPath   string
	serverPath   string
	aaptPath     string
	logcatCmd    *exec.Cmd
	logcatCancel context.CancelFunc

	// Generic mutex for shared state
	mu sync.Mutex

	// aaptCache caches app label & icon so each package is processed at most once.
	aaptCache   map[string]AppPackage
	aaptCacheMu sync.RWMutex
	cachePath   string

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

	// History and Settings
	historyPath  string
	settingsPath string
	historyMu    sync.Mutex

	version string

	// Last active tracking
	lastActive   map[string]int64
	lastActiveMu sync.RWMutex

	// Device pinning
	pinnedSerial string
	pinnedMu     sync.RWMutex

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

	// Event System (new)
	eventStore       *EventStore
	eventPipeline    *EventPipeline
	assertionEngine  *AssertionEngine
	dataDir          string
}

// NewApp creates a new App instance
func NewApp(version string) *App {
	app := &App{
		aaptCache:         make(map[string]AppPackage),
		scrcpyCmds:        make(map[string]*exec.Cmd),
		scrcpyRecordCmd:   make(map[string]*exec.Cmd),
		openFileCmds:      make(map[string]*exec.Cmd),
		lastActive:        make(map[string]int64),
		idToSerial:        make(map[string]string),
		reconnectCooldown: make(map[string]time.Time),
		version:           version,
	}
	app.initPersistentCache()
	return app
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.setupBinaries()
	a.initPersistentCache()
	a.initEventSystem() // Initialize new event system
	a.StartDeviceMonitor()
	a.StartBatchSync() // Start session event batch sync (legacy, for compatibility)
}

// Shutdown is called when the application is closing
func (a *App) Shutdown(ctx context.Context) {
	a.StopBatchSync() // Stop session event batch sync (legacy)
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
}

// GetAppVersion returns the application version
func (a *App) GetAppVersion() string {
	return a.version
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Log adds a message to the runtime logs
func (a *App) Log(format string, args ...interface{}) {
	a.logsMu.Lock()
	defer a.logsMu.Unlock()
	msg := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
	a.runtimeLogs = append(a.runtimeLogs, msg)
	if len(a.runtimeLogs) > 1000 {
		a.runtimeLogs = a.runtimeLogs[len(a.runtimeLogs)-1000:]
	}
	fmt.Println(msg)
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
	if deviceId == "" {
		return
	}

	serial := deviceId
	a.idToSerialMu.RLock()
	if s, ok := a.idToSerial[deviceId]; ok {
		serial = s
	}
	a.idToSerialMu.RUnlock()

	a.lastActiveMu.Lock()
	a.lastActive[serial] = time.Now().Unix()
	a.lastActiveMu.Unlock()

	go a.saveSettings()
}

// Initialization functions

func (a *App) initPersistentCache() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	appConfigDir := filepath.Join(configDir, "Gaze")
	_ = os.MkdirAll(appConfigDir, 0755)
	a.cachePath = filepath.Join(appConfigDir, "aapt_cache.json")
	a.historyPath = filepath.Join(appConfigDir, "history.json")
	a.settingsPath = filepath.Join(appConfigDir, "settings.json")

	a.loadCache()
	a.loadSettings()
}

func (a *App) loadSettings() {
	if a.settingsPath == "" {
		return
	}
	data, err := os.ReadFile(a.settingsPath)
	if err != nil {
		return
	}
	var settings AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return
	}

	a.lastActiveMu.Lock()
	if settings.LastActive != nil {
		a.lastActive = settings.LastActive
	}
	a.lastActiveMu.Unlock()

	a.pinnedMu.Lock()
	a.pinnedSerial = settings.PinnedSerial
	a.pinnedMu.Unlock()
}

func (a *App) saveSettings() {
	if a.settingsPath == "" {
		return
	}

	a.lastActiveMu.RLock()
	lastActive := make(map[string]int64)
	for k, v := range a.lastActive {
		lastActive[k] = v
	}
	a.lastActiveMu.RUnlock()

	a.pinnedMu.RLock()
	pinnedSerial := a.pinnedSerial
	a.pinnedMu.RUnlock()

	settings := AppSettings{
		LastActive:   lastActive,
		PinnedSerial: pinnedSerial,
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return
	}
	_ = os.WriteFile(a.settingsPath, data, 0644)
}

func (a *App) loadCache() {
	a.aaptCacheMu.Lock()
	defer a.aaptCacheMu.Unlock()

	data, err := os.ReadFile(a.cachePath)
	if err != nil {
		return
	}

	_ = json.Unmarshal(data, &a.aaptCache)
}

func (a *App) saveCache() {
	a.aaptCacheMu.RLock()
	data, err := json.Marshal(a.aaptCache)
	a.aaptCacheMu.RUnlock()

	if err != nil {
		a.Log("Error marshaling cache: %v", err)
		return
	}

	err = os.WriteFile(a.cachePath, data, 0644)
	if err != nil {
		a.Log("Error saving cache to %s: %v", a.cachePath, err)
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

	a.Log("Event system initialized at: %s", a.dataDir)
}

// shutdownEventSystem shuts down the event system
func (a *App) shutdownEventSystem() {
	if a.eventPipeline != nil {
		a.eventPipeline.Stop()
	}
	if a.eventStore != nil {
		if err := a.eventStore.Close(); err != nil {
			a.Log("Error closing event store: %v", err)
		}
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

// QuerySessionEvents queries events from a session
func (a *App) QuerySessionEvents(query EventQuery) (*EventQueryResult, error) {
	log.Printf("[QuerySessionEvents] Called with sessionId=%s, startTime=%d, endTime=%d, limit=%d",
		query.SessionID, query.StartTime, query.EndTime, query.Limit)
	if a.eventStore == nil {
		log.Printf("[QuerySessionEvents] ERROR: eventStore is nil!")
		return &EventQueryResult{Events: []UnifiedEvent{}}, nil
	}
	result, err := a.eventStore.QueryEvents(query)
	if err != nil {
		log.Printf("[QuerySessionEvents] ERROR: %v", err)
	} else {
		log.Printf("[QuerySessionEvents] Returned %d events, total=%d", len(result.Events), result.Total)
	}
	return result, err
}

// GetStoredEvent gets a single event by ID
func (a *App) GetStoredEvent(eventID string) (*UnifiedEvent, error) {
	if a.eventStore == nil {
		return nil, nil
	}
	return a.eventStore.GetEvent(eventID)
}

// GetStoredSession gets a session by ID
func (a *App) GetStoredSession(sessionID string) (*DeviceSession, error) {
	log.Printf("[GetStoredSession] Called with sessionID=%s", sessionID)
	if a.eventStore == nil {
		log.Printf("[GetStoredSession] ERROR: eventStore is nil!")
		return nil, nil
	}
	session, err := a.eventStore.GetSession(sessionID)
	log.Printf("[GetStoredSession] Result: session=%+v, err=%v", session, err)
	return session, err
}

// ListStoredSessions lists sessions from storage
func (a *App) ListStoredSessions(deviceID string, limit int) ([]DeviceSession, error) {
	if a.eventStore == nil {
		return []DeviceSession{}, nil
	}
	return a.eventStore.ListSessions(deviceID, limit)
}

// DeleteStoredSession deletes a session and its events
func (a *App) DeleteStoredSession(sessionID string) error {
	if a.eventStore == nil {
		return nil
	}
	return a.eventStore.DeleteSession(sessionID)
}

// GetSessionTimeIndex gets the time index for a session
func (a *App) GetSessionTimeIndex(sessionID string) ([]TimeIndexEntry, error) {
	if a.eventStore == nil {
		return []TimeIndexEntry{}, nil
	}
	return a.eventStore.GetTimeIndex(sessionID)
}

// GetSessionStats gets statistics for a session
func (a *App) GetSessionStats(sessionID string) (map[string]interface{}, error) {
	if a.eventStore == nil {
		return nil, nil
	}
	return a.eventStore.GetSessionStats(sessionID)
}

// GetRecentSessionEvents gets recent events from memory for a session
func (a *App) GetRecentSessionEvents(sessionID string, count int) []UnifiedEvent {
	if a.eventPipeline == nil {
		return nil
	}
	return a.eventPipeline.GetRecentEvents(sessionID, count)
}

// CreateSessionBookmark creates a bookmark in a session
func (a *App) CreateSessionBookmark(sessionID string, relativeTime int64, label, color, bookmarkType string) error {
	if a.eventStore == nil {
		return nil
	}
	bookmark := &Bookmark{
		ID:           fmt.Sprintf("bm_%d", time.Now().UnixNano()),
		SessionID:    sessionID,
		RelativeTime: relativeTime,
		Label:        label,
		Color:        color,
		Type:         bookmarkType,
		CreatedAt:    time.Now().UnixMilli(),
	}
	return a.eventStore.CreateBookmark(bookmark)
}

// GetSessionBookmarks gets bookmarks for a session
func (a *App) GetSessionBookmarks(sessionID string) ([]Bookmark, error) {
	if a.eventStore == nil {
		return []Bookmark{}, nil
	}
	return a.eventStore.GetBookmarks(sessionID)
}

// DeleteSessionBookmark deletes a bookmark
func (a *App) DeleteSessionBookmark(bookmarkID string) error {
	if a.eventStore == nil {
		return nil
	}
	return a.eventStore.DeleteBookmark(bookmarkID)
}

// CleanupOldSessionData cleans up old session data
func (a *App) CleanupOldSessionData(maxAgeDays int) (int, error) {
	if a.eventStore == nil {
		return 0, nil
	}
	return a.eventStore.CleanupOldSessions(time.Duration(maxAgeDays) * 24 * time.Hour)
}

// GetEventSystemStats returns statistics about the event system
func (a *App) GetEventSystemStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if a.eventPipeline != nil {
		stats["backpressure"] = a.eventPipeline.GetBackpressureStats()
	}

	if a.eventStore != nil {
		stats["dataDir"] = a.dataDir
	}

	return stats
}

// ========================================
// Assertion API Methods
// ========================================

// ExecuteAssertion executes an assertion against stored events
func (a *App) ExecuteAssertion(assertion Assertion) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// ExecuteAssertionJSON executes an assertion from JSON (for frontend)
func (a *App) ExecuteAssertionJSON(assertionJSON string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	var assertion Assertion
	if err := json.Unmarshal([]byte(assertionJSON), &assertion); err != nil {
		return nil, fmt.Errorf("invalid assertion JSON: %v", err)
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// GetAssertionResult gets a specific assertion result
func (a *App) GetAssertionResult(resultID string) *AssertionResult {
	if a.assertionEngine == nil {
		return nil
	}
	return a.assertionEngine.GetResult(resultID)
}

// ListAssertionResults lists assertion results for a session
func (a *App) ListAssertionResults(sessionID string, limit int) []*AssertionResult {
	if a.assertionEngine == nil {
		return nil
	}
	return a.assertionEngine.ListResults(sessionID, limit)
}

// QuickAssertExists creates and executes a quick "exists" assertion
func (a *App) QuickAssertExists(sessionID, deviceID, eventType, titleMatch string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	assertion := Assertion{
		ID:        fmt.Sprintf("quick_%d", time.Now().UnixNano()),
		Name:      fmt.Sprintf("Quick check: %s exists", eventType),
		Type:      AssertExists,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Criteria: EventCriteria{
			Types:      []string{eventType},
			TitleMatch: titleMatch,
		},
		Expected: AssertionExpected{
			Exists: true,
		},
		CreatedAt: time.Now().UnixMilli(),
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// QuickAssertCount creates and executes a quick "count" assertion
func (a *App) QuickAssertCount(sessionID, deviceID, eventType string, minCount, maxCount int) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	assertion := Assertion{
		ID:        fmt.Sprintf("quick_%d", time.Now().UnixNano()),
		Name:      fmt.Sprintf("Quick count: %s", eventType),
		Type:      AssertCount,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Criteria: EventCriteria{
			Types: []string{eventType},
		},
		Expected: AssertionExpected{
			MinCount: &minCount,
			MaxCount: &maxCount,
		},
		CreatedAt: time.Now().UnixMilli(),
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// QuickAssertNoErrors creates and executes a quick "no errors" assertion
func (a *App) QuickAssertNoErrors(sessionID, deviceID string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	assertion := Assertion{
		ID:        fmt.Sprintf("quick_%d", time.Now().UnixNano()),
		Name:      "Quick check: no errors",
		Type:      AssertNotExists,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Criteria: EventCriteria{
			Levels: []EventLevel{LevelError, LevelFatal},
		},
		Expected: AssertionExpected{
			Exists: false,
		},
		CreatedAt: time.Now().UnixMilli(),
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// QuickAssertNoCrashes creates and executes a quick "no crashes" assertion
func (a *App) QuickAssertNoCrashes(sessionID, deviceID string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	assertion := Assertion{
		ID:        fmt.Sprintf("quick_%d", time.Now().UnixNano()),
		Name:      "Quick check: no crashes/ANR",
		Type:      AssertNotExists,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Criteria: EventCriteria{
			Types: []string{"app_crash", "app_anr"},
		},
		Expected: AssertionExpected{
			Exists: false,
		},
		CreatedAt: time.Now().UnixMilli(),
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// QuickAssertSequence creates and executes a quick "sequence" assertion
func (a *App) QuickAssertSequence(sessionID, deviceID string, eventTypes []string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	sequence := make([]EventCriteria, len(eventTypes))
	for i, eventType := range eventTypes {
		sequence[i] = EventCriteria{Types: []string{eventType}}
	}

	assertion := Assertion{
		ID:        fmt.Sprintf("quick_%d", time.Now().UnixNano()),
		Name:      fmt.Sprintf("Quick sequence: %s", strings.Join(eventTypes, " -> ")),
		Type:      AssertSequence,
		SessionID: sessionID,
		DeviceID:  deviceID,
		Criteria:  EventCriteria{},
		Expected: AssertionExpected{
			Sequence: sequence,
			Ordered:  true,
		},
		CreatedAt: time.Now().UnixMilli(),
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// ========================================
// Assertion Management API Methods
// ========================================

// CreateStoredAssertion creates and persists a new assertion
func (a *App) CreateStoredAssertion(assertion Assertion, saveAsTemplate bool) error {
	if a.assertionEngine == nil {
		return fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.CreateAssertion(&assertion, saveAsTemplate)
}

// CreateStoredAssertionJSON creates and persists a new assertion from JSON
func (a *App) CreateStoredAssertionJSON(assertionJSON string, saveAsTemplate bool) error {
	if a.assertionEngine == nil {
		return fmt.Errorf("assertion engine not initialized")
	}

	var assertion Assertion
	if err := json.Unmarshal([]byte(assertionJSON), &assertion); err != nil {
		return fmt.Errorf("invalid assertion JSON: %v", err)
	}

	return a.assertionEngine.CreateAssertion(&assertion, saveAsTemplate)
}

// GetStoredAssertion retrieves a stored assertion by ID
func (a *App) GetStoredAssertion(assertionID string) (*StoredAssertion, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.GetStoredAssertion(assertionID)
}

// ListStoredAssertions lists stored assertions
func (a *App) ListStoredAssertions(sessionID, deviceID string, templatesOnly bool, limit int) ([]StoredAssertion, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.ListStoredAssertions(sessionID, deviceID, templatesOnly, limit)
}

// ListAssertionTemplates lists assertion templates (convenience method)
func (a *App) ListAssertionTemplates(limit int) ([]StoredAssertion, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.ListStoredAssertions("", "", true, limit)
}

// DeleteStoredAssertion deletes a stored assertion
func (a *App) DeleteStoredAssertion(assertionID string) error {
	if a.assertionEngine == nil {
		return fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.DeleteStoredAssertion(assertionID)
}

// ExecuteStoredAssertion executes a stored assertion by ID
func (a *App) ExecuteStoredAssertion(assertionID string) (*AssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}

	stored, err := a.assertionEngine.GetStoredAssertion(assertionID)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return nil, fmt.Errorf("assertion not found: %s", assertionID)
	}

	// Convert StoredAssertion to Assertion
	var criteria EventCriteria
	if err := json.Unmarshal(stored.Criteria, &criteria); err != nil {
		return nil, fmt.Errorf("invalid criteria: %v", err)
	}
	var expected AssertionExpected
	if err := json.Unmarshal(stored.Expected, &expected); err != nil {
		return nil, fmt.Errorf("invalid expected: %v", err)
	}
	var metadata map[string]interface{}
	if len(stored.Metadata) > 0 {
		json.Unmarshal(stored.Metadata, &metadata)
	}

	assertion := Assertion{
		ID:          stored.ID,
		Name:        stored.Name,
		Description: stored.Description,
		Type:        AssertionType(stored.Type),
		SessionID:   stored.SessionID,
		DeviceID:    stored.DeviceID,
		Criteria:    criteria,
		Expected:    expected,
		Timeout:     stored.Timeout,
		Tags:        stored.Tags,
		Metadata:    metadata,
		CreatedAt:   stored.CreatedAt,
	}

	if stored.TimeRange != nil {
		assertion.TimeRange = &TimeRange{
			Start: stored.TimeRange.Start,
			End:   stored.TimeRange.End,
		}
	}

	return a.assertionEngine.ExecuteAssertion(&assertion)
}

// ListStoredAssertionResults lists persisted assertion results
func (a *App) ListStoredAssertionResults(sessionID string, limit int) ([]StoredAssertionResult, error) {
	if a.assertionEngine == nil {
		return nil, fmt.Errorf("assertion engine not initialized")
	}
	return a.assertionEngine.ListStoredResults(sessionID, limit)
}
