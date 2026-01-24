package main

import (
	"encoding/json"
)

// ========================================
// Event Source - 事件来源
// ========================================

type EventSource string

const (
	SourceLogcat    EventSource = "logcat"    // 设备日志
	SourceNetwork   EventSource = "network"   // 网络请求
	SourceDevice    EventSource = "device"    // 设备状态
	SourceApp       EventSource = "app"       // 应用生命周期
	SourceUI        EventSource = "ui"        // UI 状态变化
	SourceTouch     EventSource = "touch"     // 触摸事件
	SourceWorkflow  EventSource = "workflow"  // 自动化流程
	SourcePerf      EventSource = "perf"      // 性能指标
	SourceSystem    EventSource = "system"    // 系统事件
	SourceAssertion EventSource = "assertion" // 断言结果
)

// ========================================
// Event Category - 事件大类
// ========================================

type EventCategory string

const (
	CategoryLog         EventCategory = "log"
	CategoryNetwork     EventCategory = "network"
	CategoryState       EventCategory = "state"
	CategoryInteraction EventCategory = "interaction"
	CategoryAutomation  EventCategory = "automation"
	CategoryDiagnostic  EventCategory = "diagnostic"
)

// ========================================
// Event Level - 事件级别
// ========================================

type EventLevel string

const (
	LevelVerbose EventLevel = "verbose"
	LevelDebug   EventLevel = "debug"
	LevelInfo    EventLevel = "info"
	LevelWarn    EventLevel = "warn"
	LevelError   EventLevel = "error"
	LevelFatal   EventLevel = "fatal"
)

// ========================================
// UnifiedEvent - 统一事件结构
// ========================================

type UnifiedEvent struct {
	// === 标识字段 ===
	ID        string `json:"id"`        // UUID
	SessionID string `json:"sessionId"` // 所属会话
	DeviceID  string `json:"deviceId"`  // 设备 ID

	// === 时间字段 ===
	Timestamp    int64 `json:"timestamp"`          // Unix 毫秒 (绝对时间)
	RelativeTime int64 `json:"relativeTime"`       // 相对 Session 开始的毫秒偏移
	Duration     int64 `json:"duration,omitempty"` // 持续时间 (ms)

	// === 分类字段 ===
	Source   EventSource   `json:"source"`   // 事件来源
	Category EventCategory `json:"category"` // 事件大类
	Type     string        `json:"type"`     // 具体类型 (如 "http_request", "activity_start")
	Level    EventLevel    `json:"level"`    // 事件级别

	// === 内容字段 ===
	Title   string `json:"title"`             // 简短标题 (用于列表显示)
	Summary string `json:"summary,omitempty"` // 摘要 (可搜索)

	// === 扩展数据 (JSON) ===
	Data   json.RawMessage `json:"data,omitempty"`   // 类型特定的详细数据
	Detail json.RawMessage `json:"detail,omitempty"` // 别名，向后兼容

	// === 关联字段 ===
	ParentID string `json:"parentId,omitempty"` // 父事件 ID
	StepID   string `json:"stepId,omitempty"`   // Workflow Step ID
	TraceID  string `json:"traceId,omitempty"`  // 跨事件追踪 ID

	// === 聚合信息 ===
	AggregateCount int   `json:"aggregateCount,omitempty"` // 聚合的事件数量
	AggregateFirst int64 `json:"aggregateFirst,omitempty"` // 聚合的第一条时间
	AggregateLast  int64 `json:"aggregateLast,omitempty"`  // 聚合的最后一条时间
}

// ========================================
// DeviceSession - 设备会话
// ========================================

// SessionConfig - Session 启动配置
type SessionConfig struct {
	Logcat    LogcatConfig    `json:"logcat"`
	Recording RecordingConfig `json:"recording"`
	Proxy     ProxyConfig     `json:"proxy"`
	Monitor   MonitorConfig   `json:"monitor"`
}

type LogcatConfig struct {
	Enabled       bool   `json:"enabled"`
	PackageName   string `json:"packageName,omitempty"`
	PreFilter     string `json:"preFilter,omitempty"`
	ExcludeFilter string `json:"excludeFilter,omitempty"`
}

type RecordingConfig struct {
	Enabled bool   `json:"enabled"`
	Quality string `json:"quality,omitempty"` // "low", "medium", "high"
}

type ProxyConfig struct {
	Enabled     bool `json:"enabled"`
	Port        int  `json:"port,omitempty"`
	MitmEnabled bool `json:"mitmEnabled,omitempty"`
}

type MonitorConfig struct {
	Enabled bool `json:"enabled"` // 设备状态监控 (电池、网络、屏幕、应用生命周期)
}

type DeviceSession struct {
	ID         string `json:"id"`
	DeviceID   string `json:"deviceId"`
	Type       string `json:"type"` // "manual", "workflow", "recording", "debug", "auto"
	Name       string `json:"name"`
	StartTime  int64  `json:"startTime"` // Unix ms
	EndTime    int64  `json:"endTime"`   // 0 = active
	Status     string `json:"status"`    // "active", "completed", "failed", "cancelled"
	EventCount int    `json:"eventCount"`

	// Session config (启用的功能)
	Config SessionConfig `json:"config"`

	// Recording info
	VideoPath     string `json:"videoPath,omitempty"`
	VideoDuration int64  `json:"videoDuration,omitempty"`
	VideoOffset   int64  `json:"videoOffset,omitempty"` // 视频相对 session 开始的偏移

	// Metadata
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ========================================
// Type-Specific Data Structures
// ========================================

// LogcatData logcat 事件数据
type LogcatData struct {
	Tag          string `json:"tag"`
	Message      string `json:"message"`
	AndroidLevel string `json:"androidLevel"` // V/D/I/W/E/F
	PID          int    `json:"pid,omitempty"`
	TID          int    `json:"tid,omitempty"`
	PackageName  string `json:"packageName,omitempty"`
	Raw          string `json:"raw,omitempty"`
}

// LogcatAggregatedData 聚合的 logcat 数据
type LogcatAggregatedData struct {
	Entries []LogcatData `json:"entries"`
	Tag     string       `json:"tag"`
	Count   int          `json:"count"`
}

// NetworkRequestData 网络请求数据
type NetworkRequestData struct {
	RequestID   string `json:"requestId"`
	Method      string `json:"method"`
	URL         string `json:"url"`
	Host        string `json:"host,omitempty"`
	Path        string `json:"path,omitempty"`
	StatusCode  int    `json:"statusCode"`
	ContentType string `json:"contentType,omitempty"`

	// 请求
	RequestHeaders  map[string]string `json:"requestHeaders,omitempty"`
	RequestBody     string            `json:"requestBody,omitempty"`
	RequestBodySize int64             `json:"requestBodySize,omitempty"`

	// 响应
	ResponseHeaders  map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody     string            `json:"responseBody,omitempty"`
	ResponseBodySize int64             `json:"responseBodySize,omitempty"`

	// 时间
	StartTime     int64 `json:"startTime,omitempty"`
	EndTime       int64 `json:"endTime,omitempty"`
	DNSTime       int64 `json:"dnsTime,omitempty"`
	ConnectTime   int64 `json:"connectTime,omitempty"`
	TLSTime       int64 `json:"tlsTime,omitempty"`
	FirstByteTime int64 `json:"firstByteTime,omitempty"`

	// 状态
	IsHTTPS bool   `json:"isHttps"`
	IsWS    bool   `json:"isWs"`
	Error   string `json:"error,omitempty"`
}

// DeviceStateData 设备状态数据
type DeviceStateData struct {
	StateType string `json:"stateType"` // battery, network, screen, storage, memory

	// Battery
	BatteryLevel  int    `json:"batteryLevel,omitempty"`
	BatteryStatus string `json:"batteryStatus,omitempty"` // charging, discharging, full
	BatteryTemp   int    `json:"batteryTemp,omitempty"`

	// Network
	NetworkType    string `json:"networkType,omitempty"` // wifi, mobile, none
	WifiSSID       string `json:"wifiSsid,omitempty"`
	SignalStrength int    `json:"signalStrength,omitempty"`

	// Screen
	ScreenState string `json:"screenState,omitempty"` // on, off
	Brightness  int    `json:"brightness,omitempty"`
	Orientation string `json:"orientation,omitempty"` // portrait, landscape

	// Memory
	MemoryTotal int64 `json:"memoryTotal,omitempty"`
	MemoryAvail int64 `json:"memoryAvail,omitempty"`
	MemoryUsed  int64 `json:"memoryUsed,omitempty"`
}

// AppLifecycleData 应用生命周期数据
type AppLifecycleData struct {
	PackageName  string `json:"packageName"`
	ActivityName string `json:"activityName,omitempty"`
	Action       string `json:"action"` // start, stop, resume, pause, crash, anr
	PID          int    `json:"pid,omitempty"`

	// Crash Info
	CrashType    string `json:"crashType,omitempty"` // java, native
	CrashMessage string `json:"crashMessage,omitempty"`
	StackTrace   string `json:"stackTrace,omitempty"`
}

// TouchEventData 触摸事件数据
type TouchEventData struct {
	Action    string  `json:"action"` // down, up, move
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Pressure  float64 `json:"pressure,omitempty"`
	PointerID int     `json:"pointerId,omitempty"`

	// 手势识别
	GestureType string `json:"gestureType,omitempty"` // tap, double_tap, long_press, swipe, pinch
	SwipeDir    string `json:"swipeDir,omitempty"`    // up, down, left, right
}

// WorkflowEventData 自动化事件数据
type WorkflowEventData struct {
	WorkflowID   string `json:"workflowId"`
	WorkflowName string `json:"workflowName"`
	StepID       string `json:"stepId,omitempty"`
	StepName     string `json:"stepName,omitempty"`
	StepType     string `json:"stepType,omitempty"`

	Action       string `json:"action"` // start, step_start, step_end, complete, error
	Success      bool   `json:"success"`
	ErrorMessage string `json:"errorMessage,omitempty"`

	// 元素信息
	Selector     *ElementSelector `json:"selector,omitempty"`
	FoundElement bool             `json:"foundElement,omitempty"`
}

// PerfData 性能数据
type PerfData struct {
	MetricType string  `json:"metricType"` // cpu, memory, fps, network_speed
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`

	// CPU
	CPUUsage    float64 `json:"cpuUsage,omitempty"`
	CPUAppUsage float64 `json:"cpuAppUsage,omitempty"`

	// Memory
	MemoryRSS  int64 `json:"memoryRss,omitempty"`
	MemoryHeap int64 `json:"memoryHeap,omitempty"`

	// FPS
	FPS       float64 `json:"fps,omitempty"`
	FrameTime float64 `json:"frameTime,omitempty"`
	JankCount int     `json:"jankCount,omitempty"`
}

// AssertionData 断言数据
type AssertionData struct {
	AssertionID   string `json:"assertionId"`
	AssertionType string `json:"assertionType"` // exists, not_exists, sequence, timing, condition
	Expression    string `json:"expression"`    // 原始断言表达式

	Passed        bool   `json:"passed"`
	ActualValue   any    `json:"actualValue,omitempty"`
	ExpectedValue any    `json:"expectedValue,omitempty"`
	ErrorMessage  string `json:"errorMessage,omitempty"`

	// 关联的事件
	MatchedEvents []string `json:"matchedEvents,omitempty"` // 匹配到的事件 ID 列表
}

// ========================================
// Time Index Entry - 时间索引条目
// ========================================

type TimeIndexEntry struct {
	Second       int    `json:"second"`       // 相对 session 开始的秒数
	EventCount   int    `json:"eventCount"`   // 该秒内的事件数量
	FirstEventID string `json:"firstEventId"` // 该秒内第一个事件的 ID
	HasError     bool   `json:"hasError"`     // 该秒内是否有错误事件
}

// ========================================
// Bookmark - 时间书签
// ========================================

type Bookmark struct {
	ID           string `json:"id"`
	SessionID    string `json:"sessionId"`
	RelativeTime int64  `json:"relativeTime"` // 相对 session 开始的毫秒
	Label        string `json:"label"`
	Color        string `json:"color,omitempty"`
	Type         string `json:"type"` // user, error, milestone, assertion_fail
	CreatedAt    int64  `json:"createdAt"`
}

// ========================================
// Event Type Registry - 事件类型注册表
// ========================================

type EventTypeInfo struct {
	Type        string        // 类型标识
	Source      EventSource   // 来源
	Category    EventCategory // 大类
	Description string        // 描述
}

var EventRegistry = map[string]EventTypeInfo{
	// === Logcat 事件 ===
	"logcat": {
		Type: "logcat", Source: SourceLogcat, Category: CategoryLog,
		Description: "Android logcat output",
	},
	"logcat_aggregated": {
		Type: "logcat_aggregated", Source: SourceLogcat, Category: CategoryLog,
		Description: "Aggregated logcat entries",
	},

	// === Network 事件 ===
	"http_request": {
		Type: "http_request", Source: SourceNetwork, Category: CategoryNetwork,
		Description: "HTTP/HTTPS request",
	},
	"websocket_message": {
		Type: "websocket_message", Source: SourceNetwork, Category: CategoryNetwork,
		Description: "WebSocket message",
	},

	// === Device State 事件 ===
	"battery_change": {
		Type: "battery_change", Source: SourceDevice, Category: CategoryState,
		Description: "Battery level or status change",
	},
	"network_change": {
		Type: "network_change", Source: SourceDevice, Category: CategoryState,
		Description: "Network connectivity change",
	},
	"screen_change": {
		Type: "screen_change", Source: SourceDevice, Category: CategoryState,
		Description: "Screen state change",
	},

	// === App Lifecycle 事件 ===
	"app_start": {
		Type: "app_start", Source: SourceApp, Category: CategoryState,
		Description: "Application started",
	},
	"app_stop": {
		Type: "app_stop", Source: SourceApp, Category: CategoryState,
		Description: "Application stopped",
	},
	"activity_start": {
		Type: "activity_start", Source: SourceApp, Category: CategoryState,
		Description: "Activity started",
	},
	"activity_stop": {
		Type: "activity_stop", Source: SourceApp, Category: CategoryState,
		Description: "Activity stopped",
	},
	"app_crash": {
		Type: "app_crash", Source: SourceApp, Category: CategoryState,
		Description: "Application crash",
	},
	"app_anr": {
		Type: "app_anr", Source: SourceApp, Category: CategoryState,
		Description: "Application Not Responding",
	},

	// === Touch 事件 ===
	"touch": {
		Type: "touch", Source: SourceTouch, Category: CategoryInteraction,
		Description: "Touch event",
	},
	"gesture": {
		Type: "gesture", Source: SourceTouch, Category: CategoryInteraction,
		Description: "Recognized gesture",
	},

	// === Workflow 事件 ===
	"workflow_start": {
		Type: "workflow_start", Source: SourceWorkflow, Category: CategoryAutomation,
		Description: "Workflow execution started",
	},
	"workflow_step_start": {
		Type: "workflow_step_start", Source: SourceWorkflow, Category: CategoryAutomation,
		Description: "Workflow step started",
	},
	"workflow_step_end": {
		Type: "workflow_step_end", Source: SourceWorkflow, Category: CategoryAutomation,
		Description: "Workflow step completed",
	},
	"workflow_complete": {
		Type: "workflow_complete", Source: SourceWorkflow, Category: CategoryAutomation,
		Description: "Workflow execution completed",
	},
	"workflow_error": {
		Type: "workflow_error", Source: SourceWorkflow, Category: CategoryAutomation,
		Description: "Workflow execution error",
	},

	// === Performance 事件 ===
	"perf_sample": {
		Type: "perf_sample", Source: SourcePerf, Category: CategoryDiagnostic,
		Description: "Performance metric sample",
	},

	// === Assertion 事件 ===
	"assertion_result": {
		Type: "assertion_result", Source: SourceAssertion, Category: CategoryAutomation,
		Description: "Assertion evaluation result",
	},

	// === System 事件 ===
	"session_start": {
		Type: "session_start", Source: SourceSystem, Category: CategoryState,
		Description: "Session started",
	},
	"session_end": {
		Type: "session_end", Source: SourceSystem, Category: CategoryState,
		Description: "Session ended",
	},
	"recording_start": {
		Type: "recording_start", Source: SourceSystem, Category: CategoryState,
		Description: "Screen recording started",
	},
	"recording_end": {
		Type: "recording_end", Source: SourceSystem, Category: CategoryState,
		Description: "Screen recording ended",
	},
}

// GetCategoryForType 根据事件类型获取分类
func GetCategoryForType(eventType string) EventCategory {
	if info, ok := EventRegistry[eventType]; ok {
		return info.Category
	}
	return CategoryDiagnostic
}
