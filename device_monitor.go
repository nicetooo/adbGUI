package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ========================================
// Device Monitor - 设备状态监控
// 监控电池、网络、屏幕状态和应用生命周期
// ========================================

// mustMarshal 序列化数据，失败时返回空 JSON 对象
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		LogWarn("device_monitor").Err(err).Msg("Failed to marshal event data")
		return json.RawMessage("{}")
	}
	return data
}

type DeviceMonitor struct {
	app      *App
	deviceID string
	ctx      context.Context
	cancel   context.CancelFunc

	// 状态缓存
	lastBattery    *BatteryState
	lastNetwork    *NetworkState
	lastScreen     *ScreenState
	lastActivity   string
	lastPackage    string
	stateMu        sync.RWMutex

	// Logcat 监听 (用于应用事件)
	logcatCmd    *exec.Cmd
	logcatCancel context.CancelFunc

	// Touch 事件监听
	touchCmd    *exec.Cmd
	touchCancel context.CancelFunc
}

// BatteryState 电池状态
type BatteryState struct {
	Level       int    `json:"level"`
	Status      string `json:"status"` // charging, discharging, full, not_charging
	Temperature int    `json:"temperature"`
	Voltage     int    `json:"voltage"`
	Health      string `json:"health"`
	Plugged     string `json:"plugged"` // ac, usb, wireless, none
}

// NetworkState 网络状态
type NetworkState struct {
	Type          string `json:"type"` // wifi, mobile, none
	Connected     bool   `json:"connected"`
	WifiSSID      string `json:"wifiSsid,omitempty"`
	WifiRSSI      int    `json:"wifiRssi,omitempty"`
	MobileType    string `json:"mobileType,omitempty"` // LTE, 5G, etc.
	AirplaneMode  bool   `json:"airplaneMode"`
}

// ScreenState 屏幕状态
type ScreenState struct {
	On          bool   `json:"on"`
	Brightness  int    `json:"brightness"`
	Orientation string `json:"orientation"` // portrait, landscape
	Locked      bool   `json:"locked"`
}

// ========================================
// Monitor Lifecycle
// ========================================

// NewDeviceMonitor 创建设备监控器
func NewDeviceMonitor(app *App, deviceID string) *DeviceMonitor {
	// 继承 app.ctx，确保应用关闭时监控器也能正确停止
	ctx, cancel := context.WithCancel(app.ctx)
	return &DeviceMonitor{
		app:      app,
		deviceID: deviceID,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start 启动监控
func (m *DeviceMonitor) Start() {
	LogDebug("device_monitor").Msgf("[DeviceMonitor] Starting for device: %s", m.deviceID)
	// 启动状态轮询
	go m.pollDeviceState()

	// 启动 logcat 监听应用事件
	go m.watchAppEvents()

	// 启动触摸事件监听
	go m.watchTouchEvents()
}

// Stop 停止监控
func (m *DeviceMonitor) Stop() {
	m.cancel()
	if m.logcatCancel != nil {
		m.logcatCancel()
	}
	if m.logcatCmd != nil && m.logcatCmd.Process != nil {
		_ = m.logcatCmd.Process.Kill()
	}
	if m.touchCancel != nil {
		m.touchCancel()
	}
	if m.touchCmd != nil && m.touchCmd.Process != nil {
		_ = m.touchCmd.Process.Kill()
	}
}

// ========================================
// State Polling
// ========================================

// pollDeviceState 轮询设备状态
func (m *DeviceMonitor) pollDeviceState() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	m.checkBattery()
	m.checkNetwork()
	m.checkScreen()
	m.checkCurrentActivity()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkBattery()
			m.checkNetwork()
			m.checkScreen()
			m.checkCurrentActivity()
		}
	}
}

// checkBattery 检查电池状态
func (m *DeviceMonitor) checkBattery() {
	ctx, cancel := context.WithTimeout(m.ctx, 3*time.Second)
	defer cancel()

	cmd := m.app.newAdbCommand(ctx, "-s", m.deviceID, "shell", "dumpsys battery")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	state := parseBatteryDump(string(output))
	if state == nil {
		return
	}

	m.stateMu.Lock()
	changed := m.lastBattery == nil ||
		m.lastBattery.Level != state.Level ||
		m.lastBattery.Status != state.Status ||
		m.lastBattery.Plugged != state.Plugged
	m.lastBattery = state
	m.stateMu.Unlock()

	if changed {
		m.emitBatteryEvent(state)
	}
}

// checkNetwork 检查网络状态
func (m *DeviceMonitor) checkNetwork() {
	ctx, cancel := context.WithTimeout(m.ctx, 3*time.Second)
	defer cancel()

	// 获取 WiFi 状态
	cmd := m.app.newAdbCommand(ctx, "-s", m.deviceID, "shell", "dumpsys wifi | grep 'Wi-Fi is\\|mWifiInfo\\|SSID\\|RSSI'")
	wifiOutput, _ := cmd.Output()

	// 获取网络连接状态
	cmd2 := m.app.newAdbCommand(ctx, "-s", m.deviceID, "shell", "dumpsys connectivity | head -50")
	connOutput, _ := cmd2.Output()

	state := parseNetworkState(string(wifiOutput), string(connOutput))
	if state == nil {
		return
	}

	m.stateMu.Lock()
	changed := m.lastNetwork == nil ||
		m.lastNetwork.Type != state.Type ||
		m.lastNetwork.Connected != state.Connected ||
		m.lastNetwork.WifiSSID != state.WifiSSID
	m.lastNetwork = state
	m.stateMu.Unlock()

	if changed {
		m.emitNetworkEvent(state)
	}
}

// checkScreen 检查屏幕状态
func (m *DeviceMonitor) checkScreen() {
	ctx, cancel := context.WithTimeout(m.ctx, 3*time.Second)
	defer cancel()

	// 检查屏幕开关状态
	cmd := m.app.newAdbCommand(ctx, "-s", m.deviceID, "shell", "dumpsys power | grep 'Display Power\\|mScreenOn\\|mWakefulness'")
	output, _ := cmd.Output()

	// 检查屏幕方向
	cmd2 := m.app.newAdbCommand(ctx, "-s", m.deviceID, "shell", "dumpsys display | grep 'mCurrentOrientation'")
	orientOutput, _ := cmd2.Output()

	state := parseScreenState(string(output), string(orientOutput))
	if state == nil {
		return
	}

	m.stateMu.Lock()
	changed := m.lastScreen == nil ||
		m.lastScreen.On != state.On ||
		m.lastScreen.Orientation != state.Orientation
	m.lastScreen = state
	m.stateMu.Unlock()

	if changed {
		m.emitScreenEvent(state)
	}
}

// checkCurrentActivity 检查当前 Activity
func (m *DeviceMonitor) checkCurrentActivity() {
	ctx, cancel := context.WithTimeout(m.ctx, 3*time.Second)
	defer cancel()

	cmd := m.app.newAdbCommand(ctx, "-s", m.deviceID, "shell", "dumpsys activity activities | grep 'mResumedActivity'")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	activity, pkg := parseCurrentActivity(string(output))
	if activity == "" {
		return
	}

	m.stateMu.Lock()
	changed := m.lastActivity != activity
	m.lastActivity = activity
	m.lastPackage = pkg
	m.stateMu.Unlock()

	if changed {
		m.emitActivityEvent(pkg, activity, "resume")
	}
}

// ========================================
// App Events via Logcat
// ========================================

// watchAppEvents 监听应用事件 (崩溃、ANR 等)
func (m *DeviceMonitor) watchAppEvents() {
	LogDebug("device_monitor").Msgf("[DeviceMonitor] watchAppEvents starting for device: %s", m.deviceID)
	ctx, cancel := context.WithCancel(m.ctx)
	m.logcatCancel = cancel

	// 清除旧日志
	clearCmd := m.app.newAdbCommand(ctx, "-s", m.deviceID, "logcat", "-c")
	_ = clearCmd.Run()

	// 监听关键事件
	// ActivityTaskManager: Activity 启动/显示 (Android 10+)
	// ActivityManager: 应用启动/停止 (旧版本)
	// AndroidRuntime: 崩溃
	// ANRManager: ANR
	cmd := m.app.newAdbCommand(ctx, "-s", m.deviceID, "logcat",
		"-v", "time",
		"ActivityTaskManager:I", "ActivityManager:I", "AndroidRuntime:E", "ANRManager:E", "*:S")
	m.logcatCmd = cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		LogDebug("device_monitor").Msgf("[DeviceMonitor] watchAppEvents stdout pipe error: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		LogDebug("device_monitor").Msgf("[DeviceMonitor] watchAppEvents start error: %v", err)
		return
	}

	LogDebug("device_monitor").Msgf("[DeviceMonitor] watchAppEvents logcat started successfully")

	reader := bufio.NewReader(stdout)
	lineCount := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			LogDebug("device_monitor").Msgf("[DeviceMonitor] watchAppEvents read error: %v", err)
			break
		}

		lineCount++
		if lineCount <= 5 || lineCount%100 == 0 {
			LogDebug("device_monitor").Msgf("[DeviceMonitor] watchAppEvents line %d: %s", lineCount, strings.TrimSpace(line))
		}

		m.processAppLogLine(line)
	}
	LogDebug("device_monitor").Msgf("[DeviceMonitor] watchAppEvents ended, total lines: %d", lineCount)
}

// processAppLogLine 处理应用相关日志
func (m *DeviceMonitor) processAppLogLine(line string) {
	line = strings.TrimSpace(line)

	// 检测应用启动
	if strings.Contains(line, "START u0") && strings.Contains(line, "cmp=") {
		LogDebug("device_monitor").Msgf("[DeviceMonitor] Detected START: %s", line)
		// 解析 cmp=pkg/activity
		if match := regexp.MustCompile(`cmp=([^/]+)/([^\s}]+)`).FindStringSubmatch(line); len(match) >= 3 {
			pkg := match[1]
			activity := match[2]
			LogDebug("device_monitor").Msgf("[DeviceMonitor] Emitting activity_start: %s/%s", pkg, activity)
			m.emitActivityEvent(pkg, activity, "start")
		}
	}

	// 检测 Activity 恢复
	if strings.Contains(line, "Displayed") {
		LogDebug("device_monitor").Msgf("[DeviceMonitor] Detected Displayed: %s", line)
		if match := regexp.MustCompile(`Displayed ([^/]+)/([^\s:]+)`).FindStringSubmatch(line); len(match) >= 3 {
			pkg := match[1]
			activity := match[2]
			// 提取启动时间
			var launchTime int64
			if timeMatch := regexp.MustCompile(`\+(\d+)ms`).FindStringSubmatch(line); len(timeMatch) >= 2 {
				if t, err := strconv.ParseInt(timeMatch[1], 10, 64); err == nil {
					launchTime = t
				}
			}
			LogDebug("device_monitor").Msgf("[DeviceMonitor] Emitting activity_displayed: %s/%s (%dms)", pkg, activity, launchTime)
			m.emitActivityDisplayed(pkg, activity, launchTime)
		}
	}

	// 检测应用崩溃
	if strings.Contains(line, "FATAL EXCEPTION") || strings.Contains(line, "AndroidRuntime") && strings.Contains(line, "E/") {
		m.emitCrashEvent(line)
	}

	// 检测 ANR
	if strings.Contains(line, "ANR in") {
		if match := regexp.MustCompile(`ANR in ([^\s]+)`).FindStringSubmatch(line); len(match) >= 2 {
			m.emitANREvent(match[1], line)
		}
	}

	// 检测进程终止
	if strings.Contains(line, "Process") && strings.Contains(line, "has died") {
		if match := regexp.MustCompile(`Process ([^\s]+) \(pid (\d+)\) has died`).FindStringSubmatch(line); len(match) >= 3 {
			m.emitProcessDied(match[1], match[2])
		}
	}
}

// ========================================
// Touch Events via getevent
// ========================================

// watchTouchEvents 监听触摸事件
func (m *DeviceMonitor) watchTouchEvents() {
	LogDebug("device_monitor").Msgf("[DeviceMonitor] watchTouchEvents starting for device: %s", m.deviceID)
	ctx, cancel := context.WithCancel(m.ctx)
	m.touchCancel = cancel

	// 使用 getevent 监听触摸事件
	// -l: 使用标签而不是数字
	// -t: 显示时间戳
	cmd := m.app.newAdbCommand(ctx, "-s", m.deviceID, "shell", "getevent", "-lt")
	m.touchCmd = cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		LogDebug("device_monitor").Msgf("[DeviceMonitor] watchTouchEvents stdout pipe error: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		LogDebug("device_monitor").Msgf("[DeviceMonitor] watchTouchEvents start error: %v", err)
		return
	}

	LogDebug("device_monitor").Msgf("[DeviceMonitor] watchTouchEvents getevent started successfully")

	reader := bufio.NewReader(stdout)
	var currentTouch *touchState
	lineCount := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			LogDebug("device_monitor").Msgf("[DeviceMonitor] watchTouchEvents read error: %v", err)
			break
		}

		lineCount++
		line = strings.TrimSpace(line)

		// 解析触摸事件
		// 格式: [timestamp] /dev/input/eventX: EV_ABS ABS_MT_POSITION_X value
		if strings.Contains(line, "ABS_MT_POSITION_X") {
			x := parseTouchValue(line)
			if currentTouch == nil {
				currentTouch = &touchState{}
			}
			currentTouch.x = x
		} else if strings.Contains(line, "ABS_MT_POSITION_Y") {
			y := parseTouchValue(line)
			if currentTouch == nil {
				currentTouch = &touchState{}
			}
			currentTouch.y = y
		} else if strings.Contains(line, "BTN_TOUCH") && strings.Contains(line, "DOWN") {
			if currentTouch == nil {
				currentTouch = &touchState{}
			}
			currentTouch.action = "down"
			currentTouch.timestamp = time.Now().UnixMilli()
		} else if strings.Contains(line, "BTN_TOUCH") && strings.Contains(line, "UP") {
			if currentTouch != nil && currentTouch.action == "down" {
				// 触摸抬起，发送完整的触摸事件
				duration := time.Now().UnixMilli() - currentTouch.timestamp
				m.emitTouchEvent(currentTouch.x, currentTouch.y, "tap", duration)
				currentTouch = nil
			}
		} else if strings.Contains(line, "SYN_REPORT") {
			// 同步事件，可以在这里处理滑动等手势
			if currentTouch != nil && currentTouch.action == "" {
				// 移动事件
				currentTouch.action = "move"
			}
		}
	}

	LogDebug("device_monitor").Msgf("[DeviceMonitor] watchTouchEvents ended, total lines: %d", lineCount)
}

// touchState 触摸状态
type touchState struct {
	x         int
	y         int
	action    string // down, up, move
	timestamp int64
}

// parseTouchValue 从 getevent 行中解析值
func parseTouchValue(line string) int {
	// 格式: [timestamp] /dev/input/eventX: EV_ABS ABS_MT_POSITION_X 00000123
	parts := strings.Fields(line)
	if len(parts) >= 4 {
		// 最后一个是十六进制值
		hexVal := parts[len(parts)-1]
		if val, err := strconv.ParseInt(hexVal, 16, 32); err == nil {
			return int(val)
		}
	}
	return 0
}

// emitTouchEvent 发送触摸事件
func (m *DeviceMonitor) emitTouchEvent(x, y int, action string, duration int64) {
	if m.app.eventPipeline == nil {
		return
	}

	title := fmt.Sprintf("👆 Touch %s at (%d, %d)", action, x, y)
	if duration > 500 {
		title = fmt.Sprintf("👆 Long press at (%d, %d) - %dms", x, y, duration)
		action = "long_press"
	}

	LogDebug("device_monitor").Msgf("[DeviceMonitor] emitTouchEvent: %s", title)

	data := mustMarshal(map[string]interface{}{
		"action":   action,
		"x":        x,
		"y":        y,
		"duration": duration,
	})

	m.app.eventPipeline.Emit(UnifiedEvent{
		DeviceID:  m.deviceID,
		Timestamp: time.Now().UnixMilli(),
		Duration:  duration,
		Source:    SourceTouch,
		Category:  CategoryInteraction,
		Type:      "touch",
		Level:     LevelInfo,
		Title:     title,
		Data:      data,
	})
}

// ========================================
// Event Emitters
// ========================================

func (m *DeviceMonitor) emitBatteryEvent(state *BatteryState) {
	LogDebug("device_monitor").Msgf("[DeviceMonitor] emitBatteryEvent: level=%d, status=%s", state.Level, state.Status)
	if m.app.eventPipeline == nil {
		LogDebug("device_monitor").Msgf("[DeviceMonitor] emitBatteryEvent: eventPipeline is nil!")
		return
	}

	level := LevelInfo
	title := fmt.Sprintf("Battery: %d%% (%s)", state.Level, state.Status)

	if state.Level <= 15 {
		level = LevelWarn
		title = fmt.Sprintf("⚠️ Low Battery: %d%%", state.Level)
	}

	data := mustMarshal(state)

	m.app.eventPipeline.Emit(UnifiedEvent{
		DeviceID:  m.deviceID,
		Timestamp: time.Now().UnixMilli(),
		Source:    SourceDevice,
		Category:  CategoryState,
		Type:      "battery_change",
		Level:     level,
		Title:     title,
		Data:      data,
	})
}

func (m *DeviceMonitor) emitNetworkEvent(state *NetworkState) {
	if m.app.eventPipeline == nil {
		return
	}

	title := fmt.Sprintf("Network: %s", state.Type)
	if state.Type == "wifi" && state.WifiSSID != "" {
		title = fmt.Sprintf("WiFi: %s", state.WifiSSID)
	}
	if !state.Connected {
		title = "Network: Disconnected"
	}

	data := mustMarshal(state)

	m.app.eventPipeline.Emit(UnifiedEvent{
		DeviceID:  m.deviceID,
		Timestamp: time.Now().UnixMilli(),
		Source:    SourceDevice,
		Category:  CategoryState,
		Type:      "network_change",
		Level:     LevelInfo,
		Title:     title,
		Data:      data,
	})
}

func (m *DeviceMonitor) emitScreenEvent(state *ScreenState) {
	if m.app.eventPipeline == nil {
		return
	}

	title := "Screen: Off"
	if state.On {
		title = fmt.Sprintf("Screen: On (%s)", state.Orientation)
	}

	data := mustMarshal(state)

	m.app.eventPipeline.Emit(UnifiedEvent{
		DeviceID:  m.deviceID,
		Timestamp: time.Now().UnixMilli(),
		Source:    SourceDevice,
		Category:  CategoryState,
		Type:      "screen_change",
		Level:     LevelDebug,
		Title:     title,
		Data:      data,
	})
}

func (m *DeviceMonitor) emitActivityEvent(pkg, activity, action string) {
	if m.app.eventPipeline == nil {
		return
	}

	eventType := "activity_" + action
	title := fmt.Sprintf("%s: %s/%s", strings.Title(action), pkg, activity)

	data := mustMarshal(map[string]interface{}{
		"packageName":  pkg,
		"activityName": activity,
		"action":       action,
	})

	m.app.eventPipeline.Emit(UnifiedEvent{
		DeviceID:  m.deviceID,
		Timestamp: time.Now().UnixMilli(),
		Source:    SourceApp,
		Category:  CategoryState,
		Type:      eventType,
		Level:     LevelInfo,
		Title:     title,
		Data:      data,
	})
}

func (m *DeviceMonitor) emitActivityDisplayed(pkg, activity string, launchTime int64) {
	if m.app.eventPipeline == nil {
		return
	}

	title := fmt.Sprintf("Displayed: %s/%s", pkg, activity)
	if launchTime > 0 {
		title += fmt.Sprintf(" (+%dms)", launchTime)
	}

	data := mustMarshal(map[string]interface{}{
		"packageName":  pkg,
		"activityName": activity,
		"launchTime":   launchTime,
	})

	m.app.eventPipeline.Emit(UnifiedEvent{
		DeviceID:  m.deviceID,
		Timestamp: time.Now().UnixMilli(),
		Source:    SourceApp,
		Category:  CategoryState,
		Type:      "activity_displayed",
		Level:     LevelInfo,
		Title:     title,
		Duration:  launchTime,
		Data:      data,
	})
}

func (m *DeviceMonitor) emitCrashEvent(logLine string) {
	if m.app.eventPipeline == nil {
		return
	}

	data := mustMarshal(map[string]interface{}{
		"crashType": "java",
		"raw":       logLine,
	})

	m.app.eventPipeline.Emit(UnifiedEvent{
		DeviceID:  m.deviceID,
		Timestamp: time.Now().UnixMilli(),
		Source:    SourceApp,
		Category:  CategoryDiagnostic,
		Type:      "app_crash",
		Level:     LevelError,
		Title:     "App Crash Detected",
		Data:      data,
	})
}

func (m *DeviceMonitor) emitANREvent(pkg, logLine string) {
	if m.app.eventPipeline == nil {
		return
	}

	data := mustMarshal(map[string]interface{}{
		"packageName": pkg,
		"raw":         logLine,
	})

	m.app.eventPipeline.Emit(UnifiedEvent{
		DeviceID:  m.deviceID,
		Timestamp: time.Now().UnixMilli(),
		Source:    SourceApp,
		Category:  CategoryDiagnostic,
		Type:      "app_anr",
		Level:     LevelError,
		Title:     fmt.Sprintf("ANR in %s", pkg),
		Data:      data,
	})
}

func (m *DeviceMonitor) emitProcessDied(pkg, pid string) {
	if m.app.eventPipeline == nil {
		return
	}

	data := mustMarshal(map[string]interface{}{
		"packageName": pkg,
		"pid":         pid,
	})

	m.app.eventPipeline.Emit(UnifiedEvent{
		DeviceID:  m.deviceID,
		Timestamp: time.Now().UnixMilli(),
		Source:    SourceApp,
		Category:  CategoryState,
		Type:      "app_stop",
		Level:     LevelInfo,
		Title:     fmt.Sprintf("Process died: %s (pid %s)", pkg, pid),
		Data:      data,
	})
}

// ========================================
// Parsers
// ========================================

func parseBatteryDump(output string) *BatteryState {
	state := &BatteryState{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "level:") {
			fmt.Sscanf(line, "level: %d", &state.Level)
		} else if strings.HasPrefix(line, "status:") {
			var status int
			fmt.Sscanf(line, "status: %d", &status)
			switch status {
			case 2:
				state.Status = "charging"
			case 3:
				state.Status = "discharging"
			case 4:
				state.Status = "not_charging"
			case 5:
				state.Status = "full"
			default:
				state.Status = "unknown"
			}
		} else if strings.HasPrefix(line, "temperature:") {
			fmt.Sscanf(line, "temperature: %d", &state.Temperature)
		} else if strings.HasPrefix(line, "voltage:") {
			fmt.Sscanf(line, "voltage: %d", &state.Voltage)
		} else if strings.HasPrefix(line, "AC powered:") && strings.Contains(line, "true") {
			state.Plugged = "ac"
		} else if strings.HasPrefix(line, "USB powered:") && strings.Contains(line, "true") {
			if state.Plugged == "" {
				state.Plugged = "usb"
			}
		} else if strings.HasPrefix(line, "Wireless powered:") && strings.Contains(line, "true") {
			if state.Plugged == "" {
				state.Plugged = "wireless"
			}
		}
	}

	if state.Plugged == "" {
		state.Plugged = "none"
	}

	return state
}

func parseNetworkState(wifiOutput, connOutput string) *NetworkState {
	state := &NetworkState{
		Type: "none",
	}

	// 解析 WiFi 状态
	if strings.Contains(wifiOutput, "Wi-Fi is enabled") {
		// 查找 SSID
		if match := regexp.MustCompile(`SSID: "?([^",\n]+)"?`).FindStringSubmatch(wifiOutput); len(match) >= 2 {
			state.Type = "wifi"
			state.WifiSSID = strings.Trim(match[1], "\"")
			state.Connected = true
		}
		// 查找 RSSI
		if match := regexp.MustCompile(`RSSI: (-?\d+)`).FindStringSubmatch(wifiOutput); len(match) >= 2 {
			state.WifiRSSI, _ = strconv.Atoi(match[1])
		}
	}

	// 解析连接状态
	if strings.Contains(connOutput, "MOBILE") && strings.Contains(connOutput, "CONNECTED") {
		if state.Type == "none" {
			state.Type = "mobile"
			state.Connected = true
		}
	}

	// 检查飞行模式
	state.AirplaneMode = strings.Contains(connOutput, "airplane_mode: true")

	return state
}

func parseScreenState(powerOutput, orientOutput string) *ScreenState {
	state := &ScreenState{}

	// 解析屏幕开关状态
	if strings.Contains(powerOutput, "mWakefulness=Awake") ||
		strings.Contains(powerOutput, "mScreenOn=true") ||
		strings.Contains(powerOutput, "Display Power: state=ON") {
		state.On = true
	}

	// 解析屏幕方向
	if strings.Contains(orientOutput, "mCurrentOrientation=0") ||
		strings.Contains(orientOutput, "mCurrentOrientation=2") {
		state.Orientation = "portrait"
	} else if strings.Contains(orientOutput, "mCurrentOrientation=1") ||
		strings.Contains(orientOutput, "mCurrentOrientation=3") {
		state.Orientation = "landscape"
	} else {
		state.Orientation = "portrait"
	}

	return state
}

func parseCurrentActivity(output string) (activity, pkg string) {
	// mResumedActivity: ActivityRecord{xxx u0 com.example/.MainActivity t123}
	match := regexp.MustCompile(`u0 ([^/\s]+)/([^\s}]+)`).FindStringSubmatch(output)
	if len(match) >= 3 {
		return match[2], match[1]
	}
	return "", ""
}

// ========================================
// Manager for multiple devices
// ========================================

var (
	deviceStateMonitors   = make(map[string]*DeviceMonitor)
	deviceStateMonitorsMu sync.Mutex
)

// StartDeviceStateMonitor 启动设备状态监控 (电池/网络/屏幕/应用)
func (a *App) StartDeviceStateMonitor(deviceID string) {
	deviceStateMonitorsMu.Lock()
	defer deviceStateMonitorsMu.Unlock()

	// 如果已存在,先停止
	if m, ok := deviceStateMonitors[deviceID]; ok {
		m.Stop()
	}

	// 创建新监控
	monitor := NewDeviceMonitor(a, deviceID)
	deviceStateMonitors[deviceID] = monitor
	monitor.Start()
}

// StopDeviceStateMonitor 停止设备状态监控
func (a *App) StopDeviceStateMonitor(deviceID string) {
	deviceStateMonitorsMu.Lock()
	defer deviceStateMonitorsMu.Unlock()

	if m, ok := deviceStateMonitors[deviceID]; ok {
		m.Stop()
		delete(deviceStateMonitors, deviceID)
	}
}

// StopAllDeviceStateMonitors 停止所有设备状态监控
func (a *App) StopAllDeviceStateMonitors() {
	deviceStateMonitorsMu.Lock()
	defer deviceStateMonitorsMu.Unlock()

	for _, m := range deviceStateMonitors {
		m.Stop()
	}
	deviceStateMonitors = make(map[string]*DeviceMonitor)
}
