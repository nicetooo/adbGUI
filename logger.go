package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// ========================================
// Structured Logger - 结构化日志系统
// ========================================

// Logger 全局日志实例
var Logger zerolog.Logger

// persistentLogger 持久化日志管理器
var persistentLogger *PersistentLogger

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// LogConfig 日志配置
type LogConfig struct {
	Level       LogLevel
	Console     bool   // 是否输出到控制台
	File        bool   // 是否输出到文件
	FilePath    string // 日志文件路径
	MaxSizeMB   int    // 单个日志文件最大大小 (MB)
	MaxAgeDays  int    // 日志保留天数
	MaxBackups  int    // 最大备份数量
	Compress    bool   // 是否压缩旧日志
	TimeFormat  string // 时间格式
	AppDataPath string // 应用数据目录
}

// DefaultLogConfig 返回默认日志配置
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Level:      LogLevelInfo,
		Console:    true,
		File:       false,
		MaxSizeMB:  10,
		MaxAgeDays: 7,
		MaxBackups: 5,
		Compress:   true,
		TimeFormat: time.RFC3339,
	}
}

// PersistentLogConfig 返回持久化日志配置
func PersistentLogConfig(appDataPath string) LogConfig {
	logDir := filepath.Join(appDataPath, "logs")
	return LogConfig{
		Level:       LogLevelInfo,
		Console:     true,
		File:        true,
		FilePath:    filepath.Join(logDir, "gaze.log"),
		MaxSizeMB:   10,
		MaxAgeDays:  7,
		MaxBackups:  5,
		Compress:    true,
		TimeFormat:  time.RFC3339,
		AppDataPath: appDataPath,
	}
}

// ========================================
// PersistentLogger - 持久化日志管理器
// ========================================

// PersistentLogger 管理日志文件轮转和清理
type PersistentLogger struct {
	mu          sync.Mutex
	config      LogConfig
	currentFile *os.File
	currentSize int64
	logDir      string
}

// NewPersistentLogger 创建持久化日志管理器
func NewPersistentLogger(config LogConfig) (*PersistentLogger, error) {
	logDir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	pl := &PersistentLogger{
		config: config,
		logDir: logDir,
	}

	if err := pl.openFile(); err != nil {
		return nil, err
	}

	// 启动清理协程
	go pl.cleanupRoutine()

	return pl, nil
}

// Write 实现 io.Writer 接口
func (pl *PersistentLogger) Write(p []byte) (n int, err error) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	// 检查是否需要轮转
	if pl.config.MaxSizeMB > 0 && pl.currentSize+int64(len(p)) > int64(pl.config.MaxSizeMB)*1024*1024 {
		if err := pl.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = pl.currentFile.Write(p)
	pl.currentSize += int64(n)
	return n, err
}

// openFile 打开日志文件
func (pl *PersistentLogger) openFile() error {
	file, err := os.OpenFile(pl.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	pl.currentFile = file
	pl.currentSize = info.Size()
	return nil
}

// rotate 轮转日志文件
func (pl *PersistentLogger) rotate() error {
	if pl.currentFile != nil {
		pl.currentFile.Close()
	}

	// 生成轮转文件名
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	rotatedPath := filepath.Join(pl.logDir, fmt.Sprintf("gaze_%s.log", timestamp))

	// 重命名当前日志文件
	if err := os.Rename(pl.config.FilePath, rotatedPath); err != nil {
		// 如果重命名失败，尝试直接打开新文件
		return pl.openFile()
	}

	// 压缩旧文件
	if pl.config.Compress {
		go pl.compressFile(rotatedPath)
	}

	return pl.openFile()
}

// compressFile 压缩日志文件
func (pl *PersistentLogger) compressFile(filePath string) {
	src, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(filePath + ".gz")
	if err != nil {
		return
	}
	defer dst.Close()

	gz := gzip.NewWriter(dst)
	defer gz.Close()

	if _, err := io.Copy(gz, src); err != nil {
		os.Remove(filePath + ".gz")
		return
	}

	// 删除原文件
	os.Remove(filePath)
}

// cleanupRoutine 定期清理旧日志
func (pl *PersistentLogger) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// 启动时立即清理一次
	pl.cleanup()

	for range ticker.C {
		pl.cleanup()
	}
}

// cleanup 清理旧日志文件
func (pl *PersistentLogger) cleanup() {
	files, err := filepath.Glob(filepath.Join(pl.logDir, "gaze_*.log*"))
	if err != nil {
		return
	}

	// 按修改时间排序
	type fileInfo struct {
		path    string
		modTime time.Time
	}
	var fileInfos []fileInfo

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{path: f, modTime: info.ModTime()})
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime.After(fileInfos[j].modTime)
	})

	now := time.Now()
	for i, fi := range fileInfos {
		// 删除超过保留天数的文件
		if pl.config.MaxAgeDays > 0 && now.Sub(fi.modTime) > time.Duration(pl.config.MaxAgeDays)*24*time.Hour {
			os.Remove(fi.path)
			continue
		}

		// 删除超过备份数量的文件
		if pl.config.MaxBackups > 0 && i >= pl.config.MaxBackups {
			os.Remove(fi.path)
		}
	}
}

// Close 关闭日志文件
func (pl *PersistentLogger) Close() error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	if pl.currentFile != nil {
		return pl.currentFile.Close()
	}
	return nil
}

// ========================================
// 日志初始化
// ========================================

// InitLogger 初始化日志系统
func InitLogger(config LogConfig) error {
	var writers []io.Writer

	// 控制台输出 (带颜色)
	if config.Console {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
		writers = append(writers, consoleWriter)
	}

	// 文件输出 (使用持久化日志管理器)
	if config.File && config.FilePath != "" {
		pl, err := NewPersistentLogger(config)
		if err != nil {
			return err
		}
		persistentLogger = pl
		writers = append(writers, pl)
	}

	// 如果没有配置任何输出，默认输出到控制台
	if len(writers) == 0 {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		})
	}

	// 创建多输出 writer
	multi := zerolog.MultiLevelWriter(writers...)

	// 设置日志级别
	var level zerolog.Level
	switch config.Level {
	case LogLevelDebug:
		level = zerolog.DebugLevel
	case LogLevelInfo:
		level = zerolog.InfoLevel
	case LogLevelWarn:
		level = zerolog.WarnLevel
	case LogLevelError:
		level = zerolog.ErrorLevel
	default:
		level = zerolog.InfoLevel
	}

	// 创建 Logger
	Logger = zerolog.New(multi).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	return nil
}

// CloseLogger 关闭日志系统
func CloseLogger() {
	if persistentLogger != nil {
		persistentLogger.Close()
	}
}

// ========================================
// 便捷日志函数
// ========================================

// LogDebug 输出 Debug 级别日志
func LogDebug(module string) *zerolog.Event {
	return Logger.Debug().Str("module", module)
}

// LogInfo 输出 Info 级别日志
func LogInfo(module string) *zerolog.Event {
	return Logger.Info().Str("module", module)
}

// LogWarn 输出 Warn 级别日志
func LogWarn(module string) *zerolog.Event {
	return Logger.Warn().Str("module", module)
}

// LogError 输出 Error 级别日志
func LogError(module string) *zerolog.Event {
	return Logger.Error().Str("module", module)
}

// ========================================
// 模块特定日志
// ========================================

// DeviceLog 设备管理日志
func DeviceLog() *zerolog.Event {
	return Logger.Info().Str("module", "device")
}

// SessionLog Session 管理日志
func SessionLog() *zerolog.Event {
	return Logger.Info().Str("module", "session")
}

// EventLog 事件管道日志
func EventLog() *zerolog.Event {
	return Logger.Info().Str("module", "event")
}

// ProxyLog 代理日志
func ProxyLog() *zerolog.Event {
	return Logger.Info().Str("module", "proxy")
}

// AutomationLog 自动化日志
func AutomationLog() *zerolog.Event {
	return Logger.Info().Str("module", "automation")
}

// ========================================
// 用户交互日志 - 记录用户行为
// ========================================

// UserAction 用户操作类型
type UserAction string

const (
	// 设备相关
	ActionDeviceConnect    UserAction = "device_connect"
	ActionDeviceDisconnect UserAction = "device_disconnect"
	ActionDeviceSelect     UserAction = "device_select"

	// Session 相关
	ActionSessionStart UserAction = "session_start"
	ActionSessionEnd   UserAction = "session_end"

	// 代理相关
	ActionProxyStart UserAction = "proxy_start"
	ActionProxyStop  UserAction = "proxy_stop"
	ActionMockCreate UserAction = "mock_create"
	ActionMockDelete UserAction = "mock_delete"

	// 自动化相关
	ActionWorkflowStart  UserAction = "workflow_start"
	ActionWorkflowStop   UserAction = "workflow_stop"
	ActionScriptRun      UserAction = "script_run"
	ActionScriptStop     UserAction = "script_stop"
	ActionRecordingStart UserAction = "recording_start"
	ActionRecordingStop  UserAction = "recording_stop"

	// 屏幕相关
	ActionScrcpyStart    UserAction = "scrcpy_start"
	ActionScrcpyStop     UserAction = "scrcpy_stop"
	ActionScreenshot     UserAction = "screenshot"
	ActionScreenRecord   UserAction = "screen_record"
	ActionTouchAction    UserAction = "touch_action"
	ActionElementClick   UserAction = "element_click"
	ActionElementInspect UserAction = "element_inspect"

	// 文件相关
	ActionFilePush UserAction = "file_push"
	ActionFilePull UserAction = "file_pull"
	ActionAppInstall UserAction = "app_install"
	ActionAppUninstall UserAction = "app_uninstall"

	// Shell 相关
	ActionShellCommand UserAction = "shell_command"

	// 设置相关
	ActionSettingsChange UserAction = "settings_change"

	// UI 相关
	ActionUINavigation UserAction = "ui_navigation"
	ActionUISearch     UserAction = "ui_search"
	ActionUIFilter     UserAction = "ui_filter"
)

// UserInteractionLog 用户交互日志记录器
type UserInteractionLog struct {
	logger zerolog.Logger
}

// userInteractionLog 全局用户交互日志实例
var userInteractionLog *UserInteractionLog

// InitUserInteractionLog 初始化用户交互日志
func InitUserInteractionLog() {
	userInteractionLog = &UserInteractionLog{
		logger: Logger.With().Str("category", "user_interaction").Logger(),
	}
}

// LogUserAction 记录用户操作
func LogUserAction(action UserAction, deviceID string, details map[string]interface{}) {
	if userInteractionLog == nil {
		InitUserInteractionLog()
	}

	event := userInteractionLog.logger.Info().
		Str("action", string(action)).
		Str("device_id", deviceID).
		Time("timestamp", time.Now())

	// 添加详细信息
	for k, v := range details {
		switch val := v.(type) {
		case string:
			event.Str(k, val)
		case int:
			event.Int(k, val)
		case int64:
			event.Int64(k, val)
		case float64:
			event.Float64(k, val)
		case bool:
			event.Bool(k, val)
		case error:
			event.Err(val)
		default:
			event.Interface(k, val)
		}
	}

	event.Msg("User action")
}

// ========================================
// 运行状态日志
// ========================================

// AppState 应用状态
type AppState string

const (
	StateStarting     AppState = "starting"
	StateReady        AppState = "ready"
	StateShuttingDown AppState = "shutting_down"
	StateStopped      AppState = "stopped"
)

// LogAppState 记录应用状态变化
func LogAppState(state AppState, details map[string]interface{}) {
	event := Logger.Info().
		Str("category", "app_state").
		Str("state", string(state)).
		Time("timestamp", time.Now())

	for k, v := range details {
		switch val := v.(type) {
		case string:
			event.Str(k, val)
		case int:
			event.Int(k, val)
		case int64:
			event.Int64(k, val)
		case float64:
			event.Float64(k, val)
		case bool:
			event.Bool(k, val)
		case error:
			event.Err(val)
		default:
			event.Interface(k, val)
		}
	}

	event.Msg("App state changed")
}

// LogSystemMetrics 记录系统指标
func LogSystemMetrics(metrics map[string]interface{}) {
	event := Logger.Info().
		Str("category", "system_metrics").
		Time("timestamp", time.Now())

	for k, v := range metrics {
		switch val := v.(type) {
		case string:
			event.Str(k, val)
		case int:
			event.Int(k, val)
		case int64:
			event.Int64(k, val)
		case float64:
			event.Float64(k, val)
		case bool:
			event.Bool(k, val)
		default:
			event.Interface(k, val)
		}
	}

	event.Msg("System metrics")
}

// ========================================
// 错误追踪日志
// ========================================

// LogErrorWithContext 记录带上下文的错误
func LogErrorWithContext(module string, err error, context map[string]interface{}) {
	event := Logger.Error().
		Str("module", module).
		Err(err).
		Time("timestamp", time.Now())

	for k, v := range context {
		switch val := v.(type) {
		case string:
			event.Str(k, val)
		case int:
			event.Int(k, val)
		case int64:
			event.Int64(k, val)
		case float64:
			event.Float64(k, val)
		case bool:
			event.Bool(k, val)
		default:
			event.Interface(k, val)
		}
	}

	event.Msg("Error occurred")
}

// LogPanic 记录 panic 信息
func LogPanic(module string, recovered interface{}, stack string) {
	Logger.Error().
		Str("module", module).
		Str("category", "panic").
		Interface("recovered", recovered).
		Str("stack", stack).
		Time("timestamp", time.Now()).
		Msg("Panic recovered")
}

// ========================================
// 性能日志
// ========================================

// OperationTimer 操作计时器
type OperationTimer struct {
	module    string
	operation string
	startTime time.Time
	details   map[string]interface{}
}

// StartOperation 开始计时
func StartOperation(module, operation string) *OperationTimer {
	return &OperationTimer{
		module:    module,
		operation: operation,
		startTime: time.Now(),
		details:   make(map[string]interface{}),
	}
}

// AddDetail 添加详细信息
func (t *OperationTimer) AddDetail(key string, value interface{}) *OperationTimer {
	t.details[key] = value
	return t
}

// End 结束计时并记录日志
func (t *OperationTimer) End() {
	duration := time.Since(t.startTime)

	event := Logger.Info().
		Str("module", t.module).
		Str("category", "performance").
		Str("operation", t.operation).
		Dur("duration", duration).
		Int64("duration_ms", duration.Milliseconds())

	for k, v := range t.details {
		switch val := v.(type) {
		case string:
			event.Str(k, val)
		case int:
			event.Int(k, val)
		case int64:
			event.Int64(k, val)
		case float64:
			event.Float64(k, val)
		case bool:
			event.Bool(k, val)
		default:
			event.Interface(k, val)
		}
	}

	event.Msg("Operation completed")
}

// EndWithError 结束计时并记录错误
func (t *OperationTimer) EndWithError(err error) {
	duration := time.Since(t.startTime)

	event := Logger.Error().
		Str("module", t.module).
		Str("category", "performance").
		Str("operation", t.operation).
		Dur("duration", duration).
		Int64("duration_ms", duration.Milliseconds()).
		Err(err)

	for k, v := range t.details {
		switch val := v.(type) {
		case string:
			event.Str(k, val)
		case int:
			event.Int(k, val)
		case int64:
			event.Int64(k, val)
		case float64:
			event.Float64(k, val)
		case bool:
			event.Bool(k, val)
		default:
			event.Interface(k, val)
		}
	}

	event.Msg("Operation failed")
}

// ========================================
// 日志查询接口 (供前端调用)
// ========================================

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Module    string                 `json:"module"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// GetLogFilePath 获取日志文件路径
func GetLogFilePath() string {
	if persistentLogger != nil {
		return persistentLogger.config.FilePath
	}
	return ""
}

// GetLogDir 获取日志目录
func GetLogDir() string {
	if persistentLogger != nil {
		return persistentLogger.logDir
	}
	return ""
}

// ListLogFiles 列出所有日志文件
func ListLogFiles() ([]string, error) {
	if persistentLogger == nil {
		return nil, fmt.Errorf("persistent logger not initialized")
	}

	pattern := filepath.Join(persistentLogger.logDir, "gaze*.log*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// 按修改时间排序 (最新在前)
	type fileWithTime struct {
		path    string
		modTime time.Time
	}
	var filesWithTime []fileWithTime
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		filesWithTime = append(filesWithTime, fileWithTime{path: f, modTime: info.ModTime()})
	}

	sort.Slice(filesWithTime, func(i, j int) bool {
		return filesWithTime[i].modTime.After(filesWithTime[j].modTime)
	})

	result := make([]string, len(filesWithTime))
	for i, f := range filesWithTime {
		result[i] = f.path
	}
	return result, nil
}

// ReadRecentLogs 读取最近的日志 (最后 n 行)
func ReadRecentLogs(lines int) ([]string, error) {
	if persistentLogger == nil {
		return nil, fmt.Errorf("persistent logger not initialized")
	}

	content, err := os.ReadFile(persistentLogger.config.FilePath)
	if err != nil {
		return nil, err
	}

	allLines := strings.Split(string(content), "\n")
	if len(allLines) <= lines {
		return allLines, nil
	}

	return allLines[len(allLines)-lines:], nil
}

// ========================================
// 初始化
// ========================================

func init() {
	// 默认初始化 (控制台输出)
	_ = InitLogger(DefaultLogConfig())
}
