# Gaze (adbGUI) - Claude Code 项目指南

## 项目概述

Gaze 是一个跨平台的 Android 设备管理工具，使用 **Go + React** 构建，基于 **Wails** 框架。核心理念是通过统一的 **Session-Event** 架构，将设备的所有活动（日志、网络、触摸、应用状态等）关联到时间线上，实现完整的设备行为追踪和自动化测试。

## 最关键事项：Session 统一贯穿设备事件

### 核心架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        Session (会话)                            │
│  - 每个设备在任意时刻只有一个活跃 Session                          │
│  - Session 是所有事件的容器和时间基准                              │
│  - 所有模块产生的事件都必须关联到 Session                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     EventPipeline (事件管道)                      │
│  - 统一的事件入口：所有事件通过 Emit() 进入管道                     │
│  - 自动关联 Session：根据 DeviceID 查找活跃 Session               │
│  - 计算 RelativeTime：事件时间相对于 Session 开始的偏移            │
│  - 背压控制：防止事件洪泛，保护关键事件                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     EventStore (事件存储)                         │
│  - SQLite 持久化：支持大量事件的高效存储和查询                      │
│  - 时间索引：按秒聚合，支持快速跳转                                 │
│  - 全文搜索：FTS5 支持事件内容搜索                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 事件来源 (EventSource)

所有设备活动都统一为事件，通过 `Source` 字段区分来源：

| Source      | 描述         | 典型事件类型                                        |
| ----------- | ------------ | --------------------------------------------------- |
| `logcat`    | 设备日志     | logcat, logcat_aggregated                           |
| `network`   | 网络请求     | http_request, websocket_message                     |
| `device`    | 设备状态     | battery_change, network_change, screen_change       |
| `app`       | 应用生命周期 | app_start, app_stop, app_crash, app_anr             |
| `touch`     | 触摸事件     | touch, gesture                                      |
| `workflow`  | 自动化流程   | workflow*start, workflow_step*\*, workflow_complete |
| `ui`        | UI 状态      | element_found, element_click                        |
| `perf`      | 性能指标     | perf_sample                                         |
| `assertion` | 断言结果     | assertion_result                                    |
| `system`    | 系统事件     | session_start, session_end, recording_start         |

### 关键代码路径

1. **Session 创建**: `event_pipeline.go:StartSession()` / `getOrCreateSession()`
2. **事件发送**: `event_pipeline.go:Emit()` / `EmitRaw()`
3. **事件处理**: `event_pipeline.go:processEvent()` - 关联 Session、计算相对时间、写入存储
4. **事件存储**: `event_store.go:WriteEvent()` / `QueryEvents()`
5. **前端同步**: `event_pipeline.go:frontendEmitter()` - 批量推送到前端

### 发送事件的正确方式

```go
// 方式 1: 使用 EmitRaw (推荐，自动填充字段)
a.eventPipeline.EmitRaw(
    deviceID,           // 设备 ID
    SourceNetwork,      // 事件来源
    "http_request",     // 事件类型
    LevelInfo,          // 事件级别
    "GET /api/users",   // 标题
    networkRequestData, // 详细数据 (会被 JSON 序列化)
)

// 方式 2: 构造完整事件
a.eventPipeline.Emit(UnifiedEvent{
    ID:        uuid.New().String(),
    DeviceID:  deviceID,
    Timestamp: time.Now().UnixMilli(),
    Source:    SourceLogcat,
    Category:  CategoryLog,
    Type:      "logcat",
    Level:     LevelInfo,
    Title:     "ActivityManager: Starting activity",
    Data:      json.RawMessage(`{"tag":"ActivityManager","message":"..."}`),
})
```

### 添加新事件类型的步骤

1. **定义事件类型** (`event_types.go`):

   ```go
   // 在 EventRegistry 中注册
   "my_new_event": {
       Type: "my_new_event",
       Source: SourceApp,
       Category: CategoryState,
       Description: "My new event type",
   },

   // 如需要，定义数据结构
   type MyNewEventData struct {
       Field1 string `json:"field1"`
       Field2 int    `json:"field2"`
   }
   ```

2. **发送事件** (在相应模块中):

   ```go
   a.eventPipeline.EmitRaw(deviceID, SourceApp, "my_new_event", LevelInfo, "标题", MyNewEventData{...})
   ```

3. **前端处理** (`frontend/src/stores/eventStore.ts`):
   ```typescript
   // 事件会自动通过 "events-batch" 推送到前端
   // 在 eventStore 中按需添加处理逻辑
   ```

## 技术栈

### 后端 (Go)

- **Wails v2**: 桌面应用框架
- **SQLite**: 事件存储 (WAL 模式)
- **goproxy**: HTTP/HTTPS 代理

### 前端 (React/TypeScript)

- **React 18** + **Ant Design**: UI
- **Zustand**: 状态管理
- **XYFlow**: 工作流可视化
- **i18next**: 国际化

## 项目结构

```
adbGUI/
├── 核心后端文件
│   ├── main.go              # 应用入口
│   ├── app.go               # App 结构和初始化
│   ├── event_types.go       # ⭐ 事件类型定义
│   ├── event_store.go       # ⭐ SQLite 事件存储
│   ├── event_pipeline.go    # ⭐ 事件处理管道
│   ├── session_manager.go   # ⭐ 会话管理 (兼容层)
│   ├── event_assertion.go   # 断言引擎
│   ├── device.go            # 设备管理
│   ├── device_monitor.go    # 设备状态监控
│   ├── automation.go        # 自动化引擎
│   ├── proxy_bridge.go      # 代理事件桥接
│   └── workflow.go          # 工作流执行
│
├── frontend/src/
│   ├── stores/
│   │   ├── eventStore.ts    # ⭐ 前端事件状态
│   │   ├── sessionStore.ts  # ⭐ 前端会话状态
│   │   └── ...              # 其他状态管理
│   ├── components/
│   │   ├── EventTimeline.tsx # 事件时间线组件
│   │   ├── SessionManager.tsx
│   │   └── ...
│   └── locales/             # 多语言 (en/zh/ja/ko)
│
└── proxy/                   # HTTP/HTTPS 代理模块
    ├── proxy.go
    └── cert.go
```

## 常用命令

```bash
# 开发模式
wails dev

# 构建
wails build

# 运行测试
go test ./...
```

## 开发注意事项

### 前端状态管理规范

**所有应用状态必须在 Zustand Store 中管理**，禁止在组件中使用独立的 `useState` 管理业务状态。

```
frontend/src/stores/
├── eventStore.ts       # 事件和时间线状态
├── sessionStore.ts     # 会话状态（兼容层）
├── deviceStore.ts      # 设备列表和选中状态
├── workflowStore.ts    # 工作流编辑和执行状态
├── automationStore.ts  # 触摸录制和任务状态
├── proxyStore.ts       # 代理和网络请求状态
├── aiStore.ts          # AI 服务配置和调用状态
├── uiStore.ts          # 全局 UI 状态（导航、模态框）
└── ...                 # 其他领域状态
```

**规范要点**:

1. **禁止使用 `useState`**: 组件中不允许使用 `useState` 管理任何状态
2. **所有状态 → Store**: 包括 UI 状态、表单状态、临时状态都必须放在对应的 Store
3. **后端事件 → Store**: 所有 Wails 事件订阅的数据更新必须写入 Store
4. **新增状态**: 如需新状态，在现有 Store 中添加或创建新的 Store 文件

**正确示例**:
```typescript
// ✅ 在 Store 中管理业务状态
const useDeviceStore = create<DeviceState>((set) => ({
  devices: [],
  selectedDevice: null,
  setSelectedDevice: (device) => set({ selectedDevice: device }),
}));

// ✅ 组件中使用 Store
function DeviceList() {
  const { devices, selectedDevice, setSelectedDevice } = useDeviceStore();
  return <List data={devices} onSelect={setSelectedDevice} />;
}
```

**错误示例**:
```typescript
// ❌ 在组件中用 useState 管理业务状态
function DeviceList() {
  const [devices, setDevices] = useState([]);  // 错误！应该放在 Store
  const [selected, setSelected] = useState(null);  // 错误！应该放在 Store
}
```

### Session 生命周期

1. **自动创建**: 当设备产生事件但无活跃 Session 时，自动创建 "auto" 类型 Session
2. **手动创建**: 用户开始录制/工作流时，创建带配置的 Session
3. **结束 Session**: 调用 `EndSession(sessionID, status)` 或新 Session 替换旧 Session

### 事件级别 (EventLevel)

- `fatal`: 致命错误，永不丢弃
- `error`: 错误，永不丢弃
- `warn`: 警告
- `info`: 普通信息
- `debug`: 调试信息，高负载时可能被采样
- `verbose`: 详细日志，高负载时可能被丢弃

### 背压控制

`BackpressureController` 在高负载时自动：

1. 保护关键事件 (error/fatal/network/workflow)
2. 采样 verbose/debug 级别事件
3. 防止事件队列溢出

### 数据库 Schema

核心表:

- `sessions`: 会话信息
- `events`: 事件主表 (不含大数据)
- `event_data`: 事件详细数据 (分离存储)
- `time_index`: 时间索引 (按秒聚合)
- `bookmarks`: 用户书签
- `assertions`: 断言定义
- `assertion_results`: 断言结果

## 调试技巧

优先使用go后段日志直接进行后台运行日志监控

### 查看事件流

```go
// 在 event_pipeline.go:processEvent() 添加日志
log.Printf("[Event] %s: %s - %s", event.Source, event.Type, event.Title)
```

### 检查 Session 状态

```go
// 获取设备的活跃 Session
session := a.eventPipeline.GetActiveSession(deviceID)
log.Printf("Active session: %v", session)
```

### 前端调试

```typescript
// 在浏览器控制台监听事件
window.runtime.EventsOn("events-batch", (events) => {
  console.log("Events received:", events);
});
```

## 扩展指南

### 添加新的事件源模块

1. 确定事件来源 (SourceXxx) 和事件类型
2. 在 `event_types.go` 注册事件类型
3. 在模块代码中通过 `eventPipeline.EmitRaw()` 发送事件
4. 前端在 `eventStore` 中处理新事件类型

### 添加新的断言类型

1. 在 `event_assertion.go` 的 `AssertionEvaluator` 中添加评估逻辑
2. 在前端 `AssertionsPanel.tsx` 添加 UI 配置

## 性能考虑

- 事件写入使用缓冲批量提交 (500ms / 500条)
- 前端同步使用批量推送 (500ms)
- 时间索引使用内存缓存 + 定期持久化 (5s)
- 大数据 (请求/响应体) 分离存储，列表查询不加载
