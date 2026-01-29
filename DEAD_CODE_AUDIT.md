# Gaze 项目死代码与功能重叠完整审计报告

**审计日期**: 2026-01-29  
**代码规模**: Go 后端 ~27,787 行 / 前端 ~15,000+ 行  
**审计范围**: 全部 `.go` 文件（根目录 + mcp/ + proxy/ + pkg/）及全部前端 `.ts/.tsx` 文件

---

## 一、概览统计

| 类别 | Go 后端 | 前端 | 合计 |
|------|---------|------|------|
| 死代码函数/方法 | ~34 | - | 34 |
| 死代码类型/接口 | ~10 | ~15+ | 25+ |
| 完全死掉的文件 | 0 (但 session_manager.go 大部分是死代码) | **7 个文件** (~1,706 行) | 7 |
| 功能重叠（严重） | 3 | 3 | 6 |
| 功能重叠（一般） | 4 | 3 | 7 |

---

## 二、严重问题：双重 Session 管理系统（后端核心架构问题）

这是整个项目最大的架构冗余。**两套系统并行运行**，通过桥接层转换。

### 旧系统 (`session_manager.go`) vs 新系统 (`event_pipeline.go`)

| 操作 | 旧系统 (session_manager.go) | 新系统 (event_pipeline.go) |
|------|---------------------------|--------------------------|
| Session 存储 | 包级 `sessions` map + `activeSession` map | `EventPipeline.sessions` map + SQLite |
| 创建 Session | `App.CreateSession()` | `EventPipeline.StartSession()` |
| 结束 Session | `App.EndSession()` | `EventPipeline.EndSession()` |
| 获取活跃 Session | `App.GetActiveSession()` → 旧 map | `EventPipeline.GetActiveSession()` → 新 map |
| 发送事件 | `EmitSessionEvent()` → `bridgeToNewPipeline()` → 转换 → `Emit()` | 直接 `Emit()` / `EmitRaw()` |
| 数据结构 | `Session` 结构体 | `DeviceSession` 结构体 |

**问题**：两套状态可能不一致。旧系统通过 `bridgeToNewPipeline()` 将 `SessionEvent` 转换为 `UnifiedEvent`，增加了不必要的中间层。

**仍在使用旧系统的调用点**：
- `device.go:403` → `EnsureActiveSession()`
- `workflow.go:225` → `EnsureActiveSession()`
- `logcat.go:281` → `EmitSessionEvent()`
- `proxy_bridge.go:178` → `EmitSessionEventFull()`

### 三套获取活跃 Session 的方法

| 方法 | 读取位置 | 返回类型 |
|------|----------|----------|
| `App.GetActiveSession()` | 旧 `activeSession` map | `string` |
| `App.GetDeviceActiveSession()` | `EventPipeline.sessions` map | `*DeviceSession` |
| `EventPipeline.GetActiveSession()` | `EventPipeline.sessions` map | `string` |

### 桥接反模式

`EnsureActiveSession()` (session_manager.go:163) 在旧系统中创建 session，然后构造 `SessionEvent`，再调用 `bridgeToNewPipeline()` (session_manager.go:345) 转换为 `UnifiedEvent` 并转发到 EventPipeline。这是一个不必要的间接层。

### 双重事件发送路径

**路径 1（旧）**: `EmitSessionEvent` / `EmitSessionEventFull` → `emitEventInternal()` → `bridgeToNewPipeline()` → SessionEvent→UnifiedEvent 转换 → `eventPipeline.Emit()`

**路径 2（新）**: 直接 `eventPipeline.Emit()` 或 `eventPipeline.EmitRaw()`

---

## 三、后端死代码清单

### 3.1 完全无调用的函数/方法（27 个）

| # | 函数 | 文件:行号 | 说明 |
|---|------|----------|------|
| 1 | `App.Greet()` | `app.go:266` | 脚手架残留 |
| 2 | `isStaticResource()` | `proxy_bridge.go:372` | 定义但未调用 |
| 3 | `MustValidateDeviceID()` | `device.go:57` | 仅测试使用 |
| 4 | `App.EmitSessionEventWithStep()` | `session_manager.go:285` | 未调用 |
| 5 | `App.GetRecordingStatus()` | `automation_recording.go:107` | 未调用 |
| 6 | `App.GetRecentEvents()` | `session_manager.go:487` | 未调用，被新系统取代 |
| 7 | `App.CleanupOldSessions()` | `session_manager.go:555` | 未调用，新系统有自己的清理 |
| 8 | `App.ClearSession()` | `session_manager.go:573` | 未调用 |
| 9 | `App.GetSession()` | `session_manager.go:221` | 未调用，被新系统取代 |
| 10 | `EventStore.GetActiveSession()` | `event_store.go:579` | 未调用，Pipeline 用内存 map |
| 11 | `EventStore.GetAssertionResult()` | `event_store.go:1761` | 未调用 |
| 12 | `App.GetAssertionResult()` | `app_assertion.go:37` | 未调用 |
| 13 | `App.ListStoredAssertionResults()` | `app_assertion.go:378` | 未调用 |
| 14 | `App.ExecuteStoredAssertion()` | `app_assertion.go:258` | 未调用 |
| 15 | `App.ExecuteStoredAssertionInSession()` | `app_assertion.go:313` | 未调用 |
| 16 | `App.ListAssertionTemplates()` | `app_assertion.go:242` | 未调用 |
| 17 | `App.GetBestSelector()` | `selector.go:435` | 未调用 |
| 18 | `App.GetSelectorMatchCount()` | `selector.go:465` | 未调用 |
| 19 | `App.GenerateSelectorSuggestions()` | `selector.go:333` | 未调用 |
| 20 | `VideoService.ExtractKeyFrames()` | `video_service.go:366` | 未调用 |
| 21 | `VideoService.CreateThumbnail()` | `video_service.go:427` | 未调用 |
| 22 | `VideoService.GetVideoURL()` | `video_service.go:466` | 未调用 |
| 23 | `ResizeImage()` | `video_service.go:486` | 未调用 |
| 24 | `NewVideoService()` | `video_service.go:67` | 未调用，用的是 `NewVideoServiceWithPaths()` |
| 25 | `cache.Service.ClearPackageCache()` | `pkg/cache/service.go:125` | 未调用 |
| 26 | `cache.Service.CachePath()` | `pkg/cache/service.go:265` | 未调用 |
| 27 | `cache.Service.SettingsPath()` | `pkg/cache/service.go:275` | 未调用 |

### 3.2 仅测试使用的 Logger 便捷函数（7 个）

| # | 函数 | 文件:行号 |
|---|------|----------|
| 28 | `DeviceLog()` | `logger.go:375` |
| 29 | `EventLog()` | `logger.go:385` |
| 30 | `ProxyLog()` | `logger.go:390` |
| 31 | `AutomationLog()` | `logger.go:395` |
| 32 | `LogSystemMetrics()` | `logger.go:550` |
| 33 | `LogPanic()` | `logger.go:607` |
| 34 | `LogErrorWithContext()` | `logger.go:580` |

### 3.3 从未实例化的数据结构（8 个）

`event_types.go` 中以下结构体定义了完整的字段和 JSON 标签，但代码中实际发送事件时使用的是 `map[string]interface{}`：

| # | 类型 | 位置 |
|---|------|------|
| 35 | `LogcatAggregatedData` | `event_types.go:168` |
| 36 | `NetworkRequestData` | `event_types.go:175` |
| 37 | `DeviceStateData` | `event_types.go:209` |
| 38 | `AppLifecycleData` | `event_types.go:234` |
| 39 | `TouchEventData` | `event_types.go:247` |
| 40 | `WorkflowEventData` | `event_types.go:260` |
| 41 | `PerfData` | `event_types.go:277` |
| 42 | `AssertionData` | `event_types.go:297` |

### 3.4 从未使用的 UnifiedEvent 字段（5 个）

| 字段 | 位置 | 说明 |
|------|------|------|
| `ParentID` | `event_types.go:86` | 仅测试代码赋值 |
| `TraceID` | `event_types.go:88` | 仅测试代码赋值 |
| `AggregateCount` | `event_types.go:91` | 从未赋值 |
| `AggregateFirst` | `event_types.go:92` | 从未赋值 |
| `AggregateLast` | `event_types.go:93` | 从未赋值 |

---

## 四、后端功能重叠清单

### 4.1 `Shutdown()` vs `ShutdownWithoutGUI()` — 近乎相同的清理逻辑

- **文件**: `app.go:125-166` vs `app.go:193-236`
- **重叠**: 两者执行几乎完全一样的清理步骤（停止代理、关闭事件系统、杀 scrcpy、停止 logcat 等）
- **差异**: `Shutdown()` 多停了 workflow watcher；`ShutdownWithoutGUI()` 多了 `ctxCancel`

### 4.2 `startup()` vs `InitializeWithoutGUI()` — 近乎相同的初始化

- **文件**: `app.go:110-122` vs `app.go:174-185`
- **重叠**: 两者都调 `setupBinaries()`、`initEventSystem()`、`StartDeviceMonitor()`、`LoadMockRules()`
- **差异**: `startup()` 多启了 WorkflowWatcher；`InitializeWithoutGUI()` 设 `mcpMode = true`

### 4.3 `GetInstalledPackages()` vs `ListPackages()` — 重叠的包列表

- `GetInstalledPackages()` (`app.go:745`) → 返回 `[]string`（仅包名）
- `ListPackages()` (`apps.go:23`) → 返回 `[]AppPackage`（完整信息含标签、版本等）
- 两者都执行 `pm list packages`

### 4.4 重复的 `matchPattern` / `MockRule`

- `proxy/proxy.go:136` → `matchPattern()`
- `proxy_bridge.go:505` → `matchPatternLocal()` — **逐字符复制**
- `MockRule` 结构体在两个包中各定义一次

### 4.5 重复的触摸事件解析与发送

- `automation.go:2827` → `emitTouchEvent()` + 自己的 getevent 解析器
- `device_monitor.go:488` → `emitTouchEvent()` + 自己的 getevent 解析器
- 两者如果同时运行，会产生**重复事件**

### 4.6 event_store.go 中三对重复的 Scan 函数

| Row 版 | Rows 版 |
|--------|---------|
| `scanSession()` (line 633) | `scanSessionRow()` (line 664) |
| `scanAssertion()` (line 1631) | `scanAssertionRow()` (line 1677) |
| `scanEventSingle()` (line 980) | `scanEventRow()` (line 917) |

每对的字段映射逻辑完全一样，仅 `row.Scan` vs `rows.Scan` 的区别。

### 4.7 MCP Bridge 大量直通代码

`mcp_bridge.go` (821 行) 中大部分方法是简单的直通转发：

```go
func (b *MCPBridge) StartApp(deviceId, packageName string) error {
    return b.app.StartApp(deviceId, packageName)
}
```

**根本原因**: `types.go`（main 包）和 `pkg/types/` 包之间存在重复类型定义。如果 main 包直接使用 `pkg/types`，bridge 中的类型转换代码（~240 行）可以消除。

---

## 五、前端 — 可整体删除的文件（7 个文件，~1,706 行）

| # | 文件 | 行数 | 原因 |
|---|------|------|------|
| 1 | `components/SyncTimeline.tsx` | ~324 | 没有任何地方导入或渲染 |
| 2 | `components/VideoPlayer.tsx` | ~417 | 没有任何地方导入或渲染 |
| 3 | `components/TouchOverlay.tsx` | ~139 | 没有任何地方导入或渲染 |
| 4 | `stores/sessionStore.ts` | ~306 | 完全被 eventStore.ts 取代，无消费者 |
| 5 | `stores/recordingStore.ts` | ~95 | 完全被 automationStore.ts 复制，无消费者 |
| 6 | `stores/timelineStore.ts` | ~39 | 被 eventTimelineStore.ts / eventStore.ts 取代 |
| 7 | `stores/videoStore.ts` | ~386 | 仅被死掉的 VideoPlayer.tsx 引用，传递性死代码 |

---

## 六、前端 — 存活文件中的死导出

### 6.1 `eventTypes.ts` — 大量死导出

| 导出 | 行号 | 说明 |
|------|------|------|
| `getEventIconEmoji()` | 559 | 无消费者 |
| `getEventLabel()` | 586 | 无消费者 |
| `extractNetworkData()` | 606 | 无消费者 |
| `extractLogcatData()` | 614 | 无消费者 |
| `sessionFilterToEventQuery()` | 650 | 无消费者 |
| `formatTimestamp()` | 517 | 仅被死掉的 sessionStore 引用 |
| 8 个 Data 接口 | 199-355 | `LogcatData`, `NetworkRequestData`, `DeviceStateData` 等 — 无消费者 |

### 6.2 `elementStore.ts` — 多个死工具函数

| 导出 | 说明 |
|------|------|
| `parseElementBounds` / `parseBounds` | 无外部消费者 |
| `getBoundsCenter()` | 无外部消费者 |
| `findElementAtPoint()` | 无外部消费者 |
| `findElementsBySelector()` | 无外部消费者 |
| `getBestSelector()` | 无外部消费者 |
| `isUniqueSelector()` | 无外部消费者 |
| `generateSelectorSuggestions()` | 无外部消费者 |
| `BoundsRect` 接口 | 无外部消费者 |
| `ElementInfo` 接口 | 无外部消费者 |

### 6.3 `types/workflow.ts` — 大量死辅助函数

| 导出 | 说明 |
|------|------|
| `shouldStopOnError()` | 无消费者，后端逻辑前端重写但从未使用 |
| `getNextStepId()` | 无消费者 |
| `getFallbackStepId()` | 无消费者 |
| `STEP_CATEGORIES` | 无消费者 |
| `getStepCategory()` | 无消费者 |
| `getStepHandles()` | 无消费者 |
| `HandleType` 类型 | 无消费者 |
| `HandleInfo` 接口 | 无消费者 |

### 6.4 `uiStore.ts` — 死状态和方法

| 导出 | 说明 |
|------|------|
| `navigateToView()` | 定义但从未调用 |
| `syncTimelineContainerWidth` | 仅被死掉的 SyncTimeline 使用 |
| `setSyncTimelineContainerWidth` | 仅被死掉的 SyncTimeline 使用 |

### 6.5 `eventStore.ts` — 死 Hook

| 导出 | 行号 | 说明 |
|------|------|------|
| `useCurrentTimeIndex()` | 813 | 无消费者 |
| `useCurrentSession()` | 824 | 无消费者 |

### 6.6 其他零散死代码

| 文件 | 导出 | 说明 |
|------|------|------|
| `automationStore.ts` | `ScriptTaskModel` 类型 (line 46) | 无消费者 |
| `mirrorStore.ts` | `openRecordPath` (line 197) | 无消费者 |
| `stores/index.ts` | 对所有死 store 的重导出 | 传递性死代码 |
| `stores/types.ts` | `VIEW_NAME_MAP` (line 62) | 仅被死方法 `navigateToView` 使用 |

---

## 七、前端功能重叠清单

### 7.1 录制状态三重定义

**同一份 UI 状态存在于三个 Store**：
- `recordingStore.ts`（死代码）
- `automationStore.ts`
- `automationViewStore.ts`

字段完全重复：`selectedScriptNames`, `saveModalVisible`, `scriptName`, `renameModalVisible`, `editingScriptName`, `newScriptName`, `selectedScript`

### 7.2 `formatDuration` 五处实现

| 位置 | 参数 | 死代码？ |
|------|------|----------|
| `eventTypes.ts:529` | ms → string | 否 |
| `sessionStore.ts:288` | ms → string | 是 (死 store) |
| `SessionManager.tsx:50` | ms → string | 否（本地） |
| `DevicesView.tsx:360` | **seconds** → string | 否（本地） |
| `RecordingView.tsx:429` | **seconds** → string | 否（本地） |
| `AutomationView.tsx:535` | **seconds** → string | 否（本地） |

### 7.3 `getEventIcon` / `getEventColor` 多处实现

| 位置 | 返回类型 | 死代码？ |
|------|----------|----------|
| `eventTypes.ts:540` | `React.ReactNode` (icon 组件) | 否 |
| `eventTypes.ts:559` | `string` (emoji) — `getEventIconEmoji` | 是 |
| `sessionStore.ts:296` | `string` (emoji) — `getEventIcon` | 是 (死 store) |
| `SyncTimeline.tsx:38` | local `getEventIcon` | 是 (死组件) |
| `eventTypes.ts:572` | `string` (color name) — `getEventColor` | 否 |
| `sessionStore.ts:300` | `string` (hex color) — `getEventColor` | 是 (死 store) |
| `SyncTimeline.tsx:54` | local `getEventColor` | 是 (死组件) |

### 7.4 `UINode` / `ElementSelector` 类型三处定义

| 类型 | 定义位置 |
|------|----------|
| `UINode` | `elementStore.ts:13`, `automationStore.ts:48`, `wails-models.ts:960` |
| `ElementSelector` | `elementStore.ts:33`, `types/workflow.ts:40`, `wails-models.ts:580` |

---

## 八、测试覆盖现状

### 后端测试覆盖

| 文件 | 有测试？ | 覆盖程度 |
|------|----------|----------|
| `event_pipeline.go` | 部分 | RingBuffer/BackpressureController/LRUCache 有直接测试；StartSession/EndSession/Emit 有间接测试 (mcp_bridge_test.go) |
| `event_store.go` | 是 | Session CRUD、Event 读写查询、时间索引、并发写入 |
| `session_manager.go` | **否** | **零测试** — 全文件无任何测试覆盖 |
| `app.go` (startup/shutdown) | **否** | 深度耦合 Wails，难以单元测试 |
| `proxy_bridge.go` (matchPattern) | **否** | 纯函数，容易测试但没写 |
| `device_monitor.go` (emitTouchEvent) | **否** | 无测试 |
| `automation.go` (emitTouchEvent) | **否** | 无测试 |
| `workflow_execution.go` | 是 | StepResult, determineNextStep, 验证逻辑 |
| `workflow_types.go` | 是 | 序列化/验证/辅助方法 |
| `logger.go` | 是 | 全部 logger 功能 |
| `mcp/*.go` | 是 | 全部 MCP 工具（通过 mock 测试） |

### 前端测试覆盖

| 文件 | 有测试？ |
|------|----------|
| `workflowStore.ts` | 是 (`workflowStore.test.ts`) |
| 所有其他 stores | **否** |
| 所有组件 | **否** |

---

## 九、清理计划（按优先级排列）

### Phase 1: 消除 Session 双系统（高影响，高工作量）

**前置**：为 session_manager.go 写特征测试

1. 迁移所有 `EmitSessionEvent*` 调用者到 `eventPipeline.EmitRaw()`
   - `logcat.go:281`
   - `proxy_bridge.go:178`
2. 迁移 `EnsureActiveSession` 调用者到 EventPipeline 的 `getOrCreateSession()`
   - `device.go:403`
   - `workflow.go:225`
3. 整体删除 `session_manager.go`（~632 行）
   - `Session` / `SessionEvent` 结构体
   - `bridgeToNewPipeline()` 转换
   - 重复的 `sessions` / `activeSession` 内存 map
   - 所有死方法

**预计减少**: ~632 行

### Phase 2: 删除后端死代码（低风险）

1. 删除 27 个从未调用的函数/方法
2. 删除 7 个仅测试用的 logger 便捷函数（保留测试中使用的）
3. 关键文件：`app_assertion.go`, `video_service.go`, `selector.go`, `app.go`, `event_store.go`, `pkg/cache/`

**预计减少**: ~400 行

### Phase 3: 删除后端死类型（低风险）

1. 删除 8 个从未实例化的 event data 结构体
2. 评估是否删除 5 个未使用的 UnifiedEvent 字段（或标记为 reserved）

**预计减少**: ~100 行

### Phase 4: 统一后端重叠代码（中风险）

**前置**：为 matchPattern、Shutdown/startup 写测试

1. 提取 `Shutdown` / `ShutdownWithoutGUI` 共享逻辑到 `cleanupAll()`
2. 提取 `startup` / `InitializeWithoutGUI` 共享逻辑到 `commonInit()`
3. 导出 `matchPattern` / `MockRule` 从 `proxy/` 包，删除 `proxy_bridge.go` 中的副本
4. 考虑抽象 `event_store.go` 中的 scan 函数对

**预计减少**: ~200 行

### Phase 5: 防止触摸事件重复发送（中风险）

1. 添加守卫：当触摸录制活跃时，DeviceMonitor 跳过触摸事件发送
2. 考虑统一 getevent 解析逻辑

### Phase 6: 删除前端死文件（零风险）

1. 删除 7 个完全死掉的文件 (~1,706 行)
2. 清理 `stores/index.ts` 中对死 store 的重导出

### Phase 7: 清理前端存活文件中的死导出（低风险）

1. `eventTypes.ts`: 删除 6 个死函数 + 8 个死接口
2. `elementStore.ts`: 删除 8 个死工具函数
3. `types/workflow.ts`: 删除 8 个死辅助函数
4. `uiStore.ts`: 删除死状态/方法
5. `eventStore.ts`: 删除 2 个死 hook
6. 其他零散清理

### Phase 8: 统一前端重叠代码（低风险）

1. 创建共享 `formatDuration` 工具函数
2. 统一 `UINode` / `ElementSelector` 类型定义（使用单一来源）
3. 清理录制状态三重定义（保留 `automationViewStore` 中的）

### Phase 9: 评估 MCP Bridge 简化（低优先级）

1. 评估让 main 包直接使用 `pkg/types` 的可行性
2. 如可行，消除 ~400-500 行类型转换代码

---

## 十、预计总清理量

| 类别 | 预计减少行数 |
|------|-------------|
| Phase 1: Session 双系统 | ~632 行 |
| Phase 2: 后端死代码 | ~400 行 |
| Phase 3: 后端死类型 | ~100 行 |
| Phase 4: 后端重叠 | ~200 行 |
| Phase 6: 前端死文件 | ~1,706 行 |
| Phase 7: 前端死导出 | ~300 行 |
| Phase 8: 前端重叠 | ~100 行 |
| **总计** | **~3,400+ 行** |

同时显著降低架构复杂度和维护负担。最具影响力的单一改动是 **Phase 1**（删除 session_manager.go），它消除了双路径事件发送、双 Session 状态和桥接转换层。
