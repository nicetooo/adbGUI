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
│   ├── workflow.go          # 工作流执行
│   └── mcp_bridge.go        # MCP 桥接层
│
├── mcp/                     # ⭐ MCP 服务器模块
│   ├── server.go            # MCP 服务器入口和 GazeApp 接口
│   ├── tools_device.go      # 设备管理工具 (adb_execute, aapt_execute, ffmpeg_execute, ffprobe_execute)
│   ├── tools_apps.go        # 应用管理工具
│   ├── tools_automation.go  # UI 自动化工具
│   ├── tools_screen.go      # 屏幕控制工具
│   ├── tools_session.go     # 会话管理工具
│   ├── tools_workflow.go    # 工作流工具
│   ├── tools_proxy.go       # 代理工具
│   ├── tools_video.go       # 视频处理工具
│   └── resources.go         # MCP 资源定义
│
├── frontend/src/
│   ├── stores/
│   │   ├── eventStore.ts    # ⭐ 前端事件状态
│   │   ├── sessionStore.ts  # ⭐ 前端会话状态
│   │   └── ...              # 其他状态管理
│   ├── components/
│   │   ├── EventTimeline.tsx  # 事件时间线组件
│   │   ├── SessionManager.tsx
│   │   ├── MCPInfoSection.tsx # ⭐ MCP 集成文档（主页展示）
│   │   └── ...
│   └── locales/             # 多语言 (en/zh/ja/ko/zh-TW)
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

## MCP (Model Context Protocol) 模块

### AI 架构原则

**重要**: 应用内不包含任何 AI 功能实现。所有 AI 能力必须通过 MCP 协议由外部 AI 客户端（如 Claude Desktop）提供。

```
正确的架构:
用户 → MCP 工具 → 外部 AI (Claude Desktop) → MCP 工具 → 设备操作

错误的架构 (已移除):
用户 → 应用内 AI → 设备操作  ❌
```

这意味着：
- **禁止** 在应用中集成 OpenAI/Anthropic 等 AI SDK
- **禁止** 创建 AI 相关的 Store（如 aiStore）
- **禁止** 实现 AI 驱动的自动化功能（如 AI 生成工作流、AI 崩溃分析）
- **允许** 通过 MCP 工具暴露数据和操作，让外部 AI 客户端调用

### ⭐ 新增功能必须同步暴露 MCP 工具

**所有新增的后端功能都必须通过 MCP 暴露给外部 AI 客户端。** 这是核心开发规范之一。

当你添加任何新功能时，必须同步完成以下工作：

1. **后端实现** → 在 `app.go` 或相关文件中实现功能方法
2. **MCP 接口暴露** → 在 `mcp/server.go` 的 `GazeApp` 接口添加方法签名
3. **MCP 桥接** → 在 `mcp_bridge.go` 中添加桥接方法
4. **MCP 工具注册** → 在 `mcp/tools_*.go` 中注册工具并编写详细描述
5. **前端 MCP 文档更新** → 更新 `MCPInfoSection.tsx` 组件中的工具列表和分类
6. **多语言翻译更新** → 在所有 locale 文件 (`en.json`, `zh.json`, `ja.json`, `ko.json`, `zh-TW.json`) 的 `mcp.tool.*` 下添加新工具的翻译

**不允许** 只在前端 UI 中添加功能而不暴露 MCP 工具。AI 客户端必须能通过 MCP 完成用户在 GUI 中能做的所有操作。

### MCP 工具文档位置

**重要**: MCP 工具的接口文档直接写在 Go 代码的 `mcp.WithDescription()` 中，这些描述会通过 MCP 协议传递给 AI 客户端。

```
mcp/
├── tools_device.go      # adb_execute, aapt_execute, ffmpeg_execute, ffprobe_execute
├── tools_apps.go        # app_list, app_info, app_start, app_stop, app_install...
├── tools_automation.go  # ui_hierarchy, ui_tap, ui_swipe, ui_input, ui_search...
├── tools_screen.go      # screen_screenshot, screen_record_start/stop...
├── tools_session.go     # session_create, session_end, session_events...
├── tools_workflow.go    # workflow_create, workflow_run, workflow_execute_step...
├── tools_proxy.go       # proxy_start, proxy_stop, proxy_status
└── tools_video.go       # video_frame, video_metadata, session_video_frame...
```

### 内置 CLI 工具

项目内嵌了以下命令行工具，全部通过 MCP 暴露：

| 工具 | MCP 工具名 | 用途 |
|------|-----------|------|
| adb | `adb_execute` | Android 调试桥，设备交互 |
| aapt | `aapt_execute` | APK 分析 (包信息、权限、资源) |
| ffmpeg | `ffmpeg_execute` | 视频/音频处理 |
| ffprobe | `ffprobe_execute` | 媒体文件分析 |
| scrcpy | 内部使用 | 屏幕镜像和录制 |

### 添加新 MCP 工具的步骤

1. **在 `mcp/server.go` 的 `GazeApp` 接口中添加方法签名**:
   ```go
   type GazeApp interface {
       // ...existing methods...
       MyNewMethod(param string) (string, error)
   }
   ```

2. **在主应用中实现方法** (如 `device.go`, `app.go` 等):
   ```go
   func (a *App) MyNewMethod(param string) (string, error) {
       // 实现逻辑
   }
   ```

3. **在 `mcp_bridge.go` 中添加桥接方法**:
   ```go
   func (b *MCPBridge) MyNewMethod(param string) (string, error) {
       return b.app.MyNewMethod(param)
   }
   ```

4. **在对应的 `mcp/tools_*.go` 中注册工具**:
   ```go
   s.server.AddTool(
       mcp.NewTool("my_new_tool",
           mcp.WithDescription(`详细的工具描述，包括：
   
   用途说明...
   
   常用命令/参数:
   - 命令1: 说明
   - 命令2: 说明
   
   示例:
     example1
     example2
   
   注意事项...`),
           mcp.WithString("param",
               mcp.Required(),
               mcp.Description("参数说明"),
           ),
       ),
       s.handleMyNewTool,
   )
   ```

5. **实现 handler 函数**:
   ```go
   func (s *MCPServer) handleMyNewTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
       args := request.GetArguments()
       param, _ := args["param"].(string)
       
       result, err := s.app.MyNewMethod(param)
       if err != nil {
           return &mcp.CallToolResult{
               Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Error: %v", err))},
               IsError: true,
           }, nil
       }
       
       return &mcp.CallToolResult{
           Content: []mcp.Content{mcp.NewTextContent(result)},
       }, nil
   }
   ```

6. **更新前端 MCP 文档组件** (`frontend/src/components/MCPInfoSection.tsx`):
   - 在 `TOOL_CATEGORIES` 数组对应分类的 `tools` 中添加新工具名
   - 如果是全新分类，在 `TOOL_CATEGORIES` 中新增分类项

7. **更新多语言翻译** (所有 locale 文件: `en.json`, `zh.json`, `ja.json`, `ko.json`, `zh-TW.json`):
   - 在 `mcp.tool.*` 下添加新工具的描述翻译
   - 如果新增了分类，同步添加 `mcp.category.*` 和 `mcp.category.*_desc` 翻译

### 前端 MCP 文档组件

应用主页（DevicesView）底部包含一个可折叠的 **MCP 集成文档区域**，由 `MCPInfoSection.tsx` 组件实现，内容包括：

- MCP 功能概述
- 各 AI 客户端配置方式（Claude Desktop / Claude Code / Cursor）
- 全部 MCP 工具列表（按分类组织，含描述）
- 可用资源列表

**关键文件**:
- 组件: `frontend/src/components/MCPInfoSection.tsx`
- 翻译: 各 locale 文件的 `mcp` 字段

每次新增/修改 MCP 工具后，**必须同步更新此组件和对应翻译**。

### MCP 工具描述规范

工具描述应该详细且结构化，参考 `tools_workflow.go` 和 `tools_device.go` 的风格：

- 简短的功能概述
- 分类列出常用命令/操作
- 提供具体的使用示例
- 说明参数格式和限制
- 添加必要的注意事项
