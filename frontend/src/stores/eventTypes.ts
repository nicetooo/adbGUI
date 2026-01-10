/**
 * Unified Event Types - matches backend event_types.go
 * Core types for the new event system with SQLite persistence
 */

// ========================================
// Event Source - 事件来源
// ========================================

export type EventSource =
  | 'logcat'     // 设备日志
  | 'network'    // 网络请求
  | 'device'     // 设备状态
  | 'app'        // 应用生命周期
  | 'ui'         // UI 状态变化
  | 'touch'      // 触摸事件
  | 'workflow'   // 自动化流程
  | 'perf'       // 性能指标
  | 'system'     // 系统事件
  | 'assertion'; // 断言结果

// ========================================
// Event Category - 事件大类
// ========================================

export type EventCategory =
  | 'log'
  | 'network'
  | 'state'
  | 'interaction'
  | 'automation'
  | 'diagnostic';

// ========================================
// Event Level - 事件级别
// ========================================

export type EventLevel =
  | 'verbose'
  | 'debug'
  | 'info'
  | 'warn'
  | 'error'
  | 'fatal';

// ========================================
// UnifiedEvent - 统一事件结构
// ========================================

export interface UnifiedEvent {
  // 标识字段
  id: string;
  sessionId: string;
  deviceId: string;

  // 时间字段
  timestamp: number;      // Unix 毫秒 (绝对时间)
  relativeTime: number;   // 相对 Session 开始的毫秒偏移
  duration?: number;      // 持续时间 (ms)

  // 分类字段
  source: EventSource;
  category: EventCategory;
  type: string;           // 具体类型 (如 "http_request", "activity_start")
  level: EventLevel;

  // 内容字段
  title: string;
  summary?: string;

  // 扩展数据 (JSON)
  data?: any;

  // Legacy alias for backward compatibility
  detail?: any;

  // 关联字段
  parentId?: string;
  stepId?: string;
  traceId?: string;

  // 聚合信息
  aggregateCount?: number;
  aggregateFirst?: number;
  aggregateLast?: number;
}

// ========================================
// Session Config - Session 启动配置
// ========================================

export interface LogcatConfig {
  enabled: boolean;
  packageName?: string;
  preFilter?: string;
  excludeFilter?: string;
}

export interface RecordingConfig {
  enabled: boolean;
  quality?: 'low' | 'medium' | 'high';
}

export interface ProxyConfig {
  enabled: boolean;
  port?: number;
  mitmEnabled?: boolean;
}

export interface MonitorConfig {
  enabled: boolean; // 设备状态监控 (电池、网络、屏幕、应用生命周期)
}

export interface SessionConfig {
  logcat: LogcatConfig;
  recording: RecordingConfig;
  proxy: ProxyConfig;
  monitor: MonitorConfig;
}

// 默认配置 - 全部启用
export const defaultSessionConfig: SessionConfig = {
  logcat: { enabled: true, packageName: '' },
  recording: { enabled: true, quality: 'medium' },
  proxy: { enabled: true, port: 8080, mitmEnabled: true },
  monitor: { enabled: true },
};

// ========================================
// DeviceSession - 设备会话
// ========================================

export interface DeviceSession {
  id: string;
  deviceId: string;
  type: string;       // "manual", "workflow", "recording", "debug", "auto"
  name: string;
  startTime: number;  // Unix ms
  endTime: number;    // 0 = active
  status: 'active' | 'completed' | 'failed' | 'cancelled';
  eventCount: number;

  // Session config (启用的功能)
  config?: SessionConfig;

  // Recording info
  videoPath?: string;
  videoDuration?: number;
  videoOffset?: number;

  // Metadata
  metadata?: Record<string, any>;
}

// ========================================
// Type-Specific Data Structures
// ========================================

// Logcat 事件数据
export interface LogcatData {
  tag: string;
  message: string;
  androidLevel: string; // V/D/I/W/E/F
  pid?: number;
  tid?: number;
  packageName?: string;
  raw?: string;
}

// 聚合的 logcat 数据
export interface LogcatAggregatedData {
  entries: LogcatData[];
  tag: string;
  count: number;
}

// 网络请求数据
export interface NetworkRequestData {
  requestId: string;
  method: string;
  url: string;
  host?: string;
  path?: string;
  statusCode: number;
  contentType?: string;

  // 请求
  requestHeaders?: Record<string, string>;
  requestBody?: string;
  requestBodySize?: number;

  // 响应
  responseHeaders?: Record<string, string>;
  responseBody?: string;
  responseBodySize?: number;

  // 时间
  startTime?: number;
  endTime?: number;
  dnsTime?: number;
  connectTime?: number;
  tlsTime?: number;
  firstByteTime?: number;

  // 状态
  isHttps: boolean;
  isWs: boolean;
  error?: string;
}

// 设备状态数据
export interface DeviceStateData {
  stateType: 'battery' | 'network' | 'screen' | 'storage' | 'memory';

  // Battery
  batteryLevel?: number;
  batteryStatus?: 'charging' | 'discharging' | 'full';
  batteryTemp?: number;

  // Network
  networkType?: 'wifi' | 'mobile' | 'none';
  wifiSsid?: string;
  signalStrength?: number;

  // Screen
  screenState?: 'on' | 'off';
  brightness?: number;
  orientation?: 'portrait' | 'landscape';

  // Memory
  memoryTotal?: number;
  memoryAvail?: number;
  memoryUsed?: number;
}

// 应用生命周期数据
export interface AppLifecycleData {
  packageName: string;
  activityName?: string;
  action: 'start' | 'stop' | 'resume' | 'pause' | 'crash' | 'anr';
  pid?: number;

  // Crash Info
  crashType?: 'java' | 'native';
  crashMessage?: string;
  stackTrace?: string;
}

// 触摸事件数据
export interface TouchEventData {
  action: 'down' | 'up' | 'move';
  x: number;
  y: number;
  pressure?: number;
  pointerId?: number;

  // 手势识别
  gestureType?: 'tap' | 'double_tap' | 'long_press' | 'swipe' | 'pinch';
  swipeDir?: 'up' | 'down' | 'left' | 'right';
}

// 自动化事件数据
export interface WorkflowEventData {
  workflowId: string;
  workflowName: string;
  stepId?: string;
  stepName?: string;
  stepType?: string;

  action: 'start' | 'step_start' | 'step_end' | 'complete' | 'error';
  success: boolean;
  errorMessage?: string;

  // 元素信息
  selector?: {
    type: string;
    value: string;
    index?: number;
  };
  foundElement?: boolean;
}

// 性能数据
export interface PerfData {
  metricType: 'cpu' | 'memory' | 'fps' | 'network_speed';
  value: number;
  unit: string;

  // CPU
  cpuUsage?: number;
  cpuAppUsage?: number;

  // Memory
  memoryRss?: number;
  memoryHeap?: number;

  // FPS
  fps?: number;
  frameTime?: number;
  jankCount?: number;
}

// 断言数据
export interface AssertionData {
  assertionId: string;
  assertionType: 'exists' | 'not_exists' | 'sequence' | 'timing' | 'condition';
  expression: string;

  passed: boolean;
  actualValue?: any;
  expectedValue?: any;
  errorMessage?: string;

  // 关联的事件
  matchedEvents?: string[];
}

// ========================================
// Query Types
// ========================================

export interface EventQuery {
  sessionId?: string;
  deviceId?: string;
  sources?: EventSource[];
  categories?: EventCategory[];
  types?: string[];
  levels?: EventLevel[];
  startTime?: number;   // 相对时间 (ms)
  endTime?: number;
  searchText?: string;
  parentId?: string;
  stepId?: string;
  traceId?: string;
  limit?: number;
  offset?: number;
  orderDesc?: boolean;  // true = 时间倒序
}

export interface EventQueryResult {
  events: UnifiedEvent[];
  total: number;
  hasMore: boolean;
}

// ========================================
// Time Index & Bookmarks
// ========================================

export interface TimeIndexEntry {
  second: number;        // 相对 session 开始的秒数
  eventCount: number;
  firstEventId: string;
  hasError: boolean;
}

export interface Bookmark {
  id: string;
  sessionId: string;
  relativeTime: number;  // 相对 session 开始的毫秒
  label: string;
  color?: string;
  type: 'user' | 'error' | 'milestone' | 'assertion_fail';
  createdAt: number;
}

// ========================================
// UI Display Helpers
// ========================================

// Source 显示配置
export const sourceConfig: Record<EventSource, { label: string; icon: string; color: string }> = {
  logcat: { label: 'Logcat', icon: '📝', color: '#52c41a' },
  network: { label: 'Network', icon: '🌐', color: '#722ed1' },
  device: { label: 'Device', icon: '📱', color: '#13c2c2' },
  app: { label: 'App', icon: '📦', color: '#1890ff' },
  ui: { label: 'UI', icon: '🖼️', color: '#eb2f96' },
  touch: { label: 'Touch', icon: '👆', color: '#fa8c16' },
  workflow: { label: 'Workflow', icon: '⚙️', color: '#2f54eb' },
  perf: { label: 'Perf', icon: '📊', color: '#faad14' },
  system: { label: 'System', icon: '🔧', color: '#8c8c8c' },
  assertion: { label: 'Assert', icon: '✅', color: '#52c41a' },
};

// Category 显示配置
export const categoryConfig: Record<EventCategory, { label: string; color: string }> = {
  log: { label: 'Log', color: '#52c41a' },
  network: { label: 'Network', color: '#722ed1' },
  state: { label: 'State', color: '#1890ff' },
  interaction: { label: 'Interaction', color: '#fa8c16' },
  automation: { label: 'Automation', color: '#2f54eb' },
  diagnostic: { label: 'Diagnostic', color: '#8c8c8c' },
};

// Level 显示配置
export const levelConfig: Record<EventLevel, { label: string; color: string; icon: string; priority: number }> = {
  verbose: { label: 'Verbose', color: '#bfbfbf', icon: '📋', priority: 0 },
  debug: { label: 'Debug', color: '#8c8c8c', icon: '🔧', priority: 1 },
  info: { label: 'Info', color: '#1890ff', icon: 'ℹ️', priority: 2 },
  warn: { label: 'Warn', color: '#faad14', icon: '⚠️', priority: 3 },
  error: { label: 'Error', color: '#ff4d4f', icon: '❌', priority: 4 },
  fatal: { label: 'Fatal', color: '#cf1322', icon: '💀', priority: 5 },
};

// Event Type 显示配置
export const eventTypeConfig: Record<string, { label: string; icon: string }> = {
  // Session
  session_start: { label: 'Session Start', icon: '🚀' },
  session_end: { label: 'Session End', icon: '🏁' },

  // Logcat
  logcat: { label: 'Log', icon: '📝' },
  logcat_aggregated: { label: 'Logs', icon: '📝' },

  // Network
  http_request: { label: 'HTTP Request', icon: '🌐' },
  websocket_message: { label: 'WebSocket', icon: '🔌' },

  // Device State
  battery_change: { label: 'Battery', icon: '🔋' },
  network_change: { label: 'Network', icon: '📶' },
  screen_change: { label: 'Screen', icon: '📱' },

  // App Lifecycle
  app_start: { label: 'App Start', icon: '▶️' },
  app_stop: { label: 'App Stop', icon: '⏹️' },
  activity_start: { label: 'Activity', icon: '📄' },
  activity_stop: { label: 'Activity Stop', icon: '📄' },
  app_crash: { label: 'Crash', icon: '💥' },
  app_anr: { label: 'ANR', icon: '⏰' },

  // Touch
  touch: { label: 'Touch', icon: '👆' },
  gesture: { label: 'Gesture', icon: '🤚' },

  // Workflow
  workflow_start: { label: 'Workflow Start', icon: '▶️' },
  workflow_step_start: { label: 'Step Start', icon: '🔷' },
  workflow_step_end: { label: 'Step End', icon: '✅' },
  workflow_complete: { label: 'Workflow Complete', icon: '🏁' },
  workflow_error: { label: 'Workflow Error', icon: '❌' },

  // Performance
  perf_sample: { label: 'Perf Sample', icon: '📊' },

  // Assertion
  assertion_result: { label: 'Assertion', icon: '✅' },

  // Recording
  recording_start: { label: 'Recording Start', icon: '🔴' },
  recording_end: { label: 'Recording End', icon: '⏹️' },
};

// ========================================
// Utility Functions
// ========================================

/**
 * 格式化相对时间为可读字符串
 */
export function formatRelativeTime(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const millis = ms % 1000;

  if (hours > 0) {
    return `${hours}:${pad(minutes)}:${pad(seconds)}.${pad3(millis)}`;
  }
  return `${minutes}:${pad(seconds)}.${pad3(millis)}`;
}

/**
 * 格式化绝对时间戳为时间字符串
 */
export function formatTimestamp(timestamp: number): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }) + '.' + pad3(date.getMilliseconds());
}

/**
 * 格式化持续时间
 */
export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  const mins = Math.floor(ms / 60000);
  const secs = Math.floor((ms % 60000) / 1000);
  return `${mins}m ${secs}s`;
}

/**
 * 获取事件显示图标
 */
export function getEventIcon(event: UnifiedEvent): string {
  const typeInfo = eventTypeConfig[event.type];
  if (typeInfo) return typeInfo.icon;

  const sourceInfo = sourceConfig[event.source];
  if (sourceInfo) return sourceInfo.icon;

  return levelConfig[event.level]?.icon || '•';
}

/**
 * 获取事件显示颜色
 */
export function getEventColor(event: UnifiedEvent): string {
  // 优先级: level error/fatal > category > source
  if (event.level === 'error' || event.level === 'fatal') {
    return levelConfig[event.level].color;
  }
  if (event.level === 'warn') {
    return levelConfig.warn.color;
  }
  return categoryConfig[event.category]?.color || sourceConfig[event.source]?.color || '#8c8c8c';
}

/**
 * 获取事件显示标签
 */
export function getEventLabel(event: UnifiedEvent): string {
  const typeInfo = eventTypeConfig[event.type];
  return typeInfo?.label || event.type;
}

/**
 * 判断事件是否为关键事件
 */
export function isCriticalEvent(event: UnifiedEvent): boolean {
  return (
    event.level === 'error' ||
    event.level === 'fatal' ||
    event.type === 'app_crash' ||
    event.type === 'app_anr'
  );
}

/**
 * 从事件数据中提取网络请求信息
 */
export function extractNetworkData(event: UnifiedEvent): NetworkRequestData | null {
  if (event.source !== 'network' || !event.data) return null;
  return event.data as NetworkRequestData;
}

/**
 * 从事件数据中提取 Logcat 信息
 */
export function extractLogcatData(event: UnifiedEvent): LogcatData | LogcatAggregatedData | null {
  if (event.source !== 'logcat' || !event.data) return null;
  return event.data as LogcatData | LogcatAggregatedData;
}

// Helper functions
function pad(n: number): string {
  return n.toString().padStart(2, '0');
}

function pad3(n: number): string {
  return n.toString().padStart(3, '0');
}

// ========================================
// Legacy Compatibility - 兼容旧类型
// ========================================

// 兼容旧的 SessionEvent 类型
export type SessionEvent = UnifiedEvent;

// 兼容旧的 Session 类型
export type Session = DeviceSession;

// 兼容旧的 SessionFilter 类型
export interface SessionFilter {
  categories?: string[];
  types?: string[];
  levels?: string[];
  stepId?: string;
  startTime?: number;
  endTime?: number;
  searchText?: string;
}

// 将旧的 SessionFilter 转换为 EventQuery
export function sessionFilterToEventQuery(sessionId: string, filter: SessionFilter): EventQuery {
  return {
    sessionId,
    categories: filter.categories as EventCategory[] | undefined,
    types: filter.types,
    levels: filter.levels as EventLevel[] | undefined,
    stepId: filter.stepId,
    startTime: filter.startTime,
    endTime: filter.endTime,
    searchText: filter.searchText,
  };
}
