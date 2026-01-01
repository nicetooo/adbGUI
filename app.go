package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	a.StartDeviceMonitor()
}

// Shutdown is called when the application is closing
func (a *App) Shutdown(ctx context.Context) {
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
