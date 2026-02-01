# Gaze

一个强大、现代且完全自包含的 Android 设备管理与自动化工具，基于 **Wails**、**React** 和 **Ant Design** 构建。采用统一的 **Session-Event** 架构实现完整的设备行为追踪，配备可视化 **Workflow** 引擎用于测试自动化，并全面集成 **MCP**（Model Context Protocol）实现 AI 驱动的设备控制。


[English](README.md) | [简体中文](locales/README.zh-CN.md) | [繁體中文](locales/README.zh-TW.md) | [日本語](locales/README.ja.md) | [한국어](locales/README.ko.md) | [Español](locales/README.es.md) | [Português](locales/README.pt.md) | [Français](locales/README.fr.md) | [Deutsch](locales/README.de.md) | [Русский](locales/README.ru.md) | [Tiếng Việt](locales/README.vi.md) | [العربية](locales/README.ar.md)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)
[![Website](https://img.shields.io/website?up_message=online&url=https%3A%2F%2Fgaze.nicetooo.com)](https://gaze.nicetooo.com)

## 为什么选择 Gaze？

- **现代且高效**：基于 Wails（Go + React）构建，提供接近原生的体验，资源占用极低。
- **真正的自包含**：无需在系统上安装 `adb`、`scrcpy`、`aapt`、`ffmpeg` 或 `ffprobe`，一切已内置，开箱即用。
- **可靠的文件传输**：macOS 上常见的 *Android File Transfer* 卡顿问题的可靠替代方案。
- **多设备能力**：支持多台设备独立、同时后台录屏。
- **Session-Event 架构**：在统一的时间线上追踪所有设备活动（日志、网络、触摸、应用生命周期）。
- **可视化 Workflow 自动化**：通过拖拽节点编辑器构建复杂的测试流程，无需编写代码。
- **通过 MCP 接入 AI**：50+ 工具通过 Model Context Protocol 暴露，与 Claude Desktop 和 Cursor 等 AI 客户端无缝集成。
- **开发者优先**：集成 Logcat、Shell、MITM 代理和 UI 检查器，由开发者设计，为开发者服务。

## 应用截图

| 设备管理 | 屏幕镜像 |
|:---:|:---:|
| <img src="screenshots/devices.png" width="400" /> | <img src="screenshots/mirror.png" width="400" /> |

| 文件管理 | 应用管理 |
|:---:|:---:|
| <img src="screenshots/files.png" width="400" /> | <img src="screenshots/apps.png" width="400" /> |

| 性能监控 | Session 时间线 |
|:---:|:---:|
| <img src="screenshots/performance.png" width="400" /> | <img src="screenshots/session.png" width="400" /> |

| Session 列表 | Logcat 查看器 |
|:---:|:---:|
| <img src="screenshots/session-list.png" width="400" /> | <img src="screenshots/logcat.png" width="400" /> |

| 可视化工作流编辑器 | 工作流列表 |
|:---:|:---:|
| <img src="screenshots/workflow-editor.png" width="400" /> | <img src="screenshots/workflow.png" width="400" /> |

| UI 检查器 | 触摸录制 |
|:---:|:---:|
| <img src="screenshots/ui-inspector.png" width="400" /> | <img src="screenshots/recording.png" width="400" /> |

| 网络代理 (MITM) | ADB Shell |
|:---:|:---:|
| <img src="screenshots/proxy.png" width="400" /> | <img src="screenshots/shell.png" width="400" /> |

---

## 功能特性

### 设备管理
- **统一设备列表**：无缝管理物理和无线设备，支持 USB/Wi-Fi 自动合并。
- **无线连接**：通过 IP/端口配对轻松连接，支持 mDNS。
- **设备历史**：快速访问以前连接过的离线设备。
- **设备置顶**：将常用设备置顶，始终显示在列表最上方。
- **设备监控**：实时追踪电池、网络和屏幕状态变化。
- **批量操作**：同时在多台设备上执行操作。

### 应用管理
- **全面的包管理**：安装（拖放）、卸载、启用、禁用、强制停止、清除数据。
- **APK 管理**：导出已安装的 APK、批量安装。
- **智能过滤**：按系统/用户应用进行搜索和过滤。
- **快速操作**：启动应用或直接跳转到其日志。

### 屏幕镜像（Scrcpy）
- **高性能**：由 Scrcpy 驱动的低延迟镜像。
- **录屏**：独立后台录制，支持多设备同时录制，一键访问录制文件夹。
- **音频转发**：将设备音频流式传输到电脑（Android 11+）。
- **自定义**：调整分辨率、码率、FPS 和编解码器（H.264/H.265）。
- **控制**：多点触控支持、保持唤醒、息屏模式。

### 文件管理
- **全功能文件管理器**：浏览、复制、剪切、粘贴、重命名、删除和创建文件夹。
- **拖放上传**：只需将文件拖到窗口即可上传。
- **下载**：轻松将文件从设备传输到电脑。
- **预览**：直接在主机上打开文件。

### 高级 Logcat
- **实时流**：带有自动滚动控制的实时日志查看器。
- **强大过滤**：按日志级别、Tag、PID 或自定义正则表达式过滤。
- **以应用为中心**：自动过滤特定应用的日志。
- **JSON 格式化**：自动美化检测到的 JSON 日志段落。

### 网络与代理（MITM）
- **自动化抓包**：一键启动 HTTP/HTTPS 代理服务器，并通过 ADB 自动配置设备代理设置。
- **HTTPS 解密（MITM）**：支持 SSL 流量解密，自动生成和部署 CA 证书。
- **WebSocket 支持**：捕获并检查实时 WebSocket 流量。
- **大数据处理**：支持完整的请求/响应体捕获（最大 100MB），不截断，5000 条日志缓冲区。
- **弱网模拟**：模拟真实网络条件，支持按设备限制下行/上行带宽以及设置人工延迟。
- **可视化指标**：实时监控所选设备的 RX/TX 速率。

### Session 与 Event 追踪
- **统一 Event Pipeline**：所有设备活动（日志、网络请求、触摸事件、应用生命周期、断言）均作为 Event 捕获并关联到 Session 时间线。
- **自动 Session 管理**：Event 产生时自动创建 Session，也可手动创建带自定义配置（logcat、录屏、代理、监控）的 Session。
- **Event 时间线**：多泳道可视化展示所有 Event，支持基于时间的索引和导航。
- **全文搜索**：使用 SQLite FTS5 在所有 Event 中搜索。
- **背压控制**：高负载时自动采样 Event，同时保护关键 Event（错误、网络、Workflow）。
- **Event 断言**：定义和评估针对 Event 流的断言，实现自动化验证。
- **视频同步**：提取与 Event 时间戳同步的视频帧，用于可视化调试。

### UI 检查器与自动化
- **UI 层级检查器**：浏览和分析任何界面的完整 UI 树。
- **元素选择器**：点击选取 UI 元素并检查其属性（resource-id、text、bounds、class）。
- **触摸录制**：录制触摸交互并作为自动化脚本回放。
- **基于元素的操作**：使用选择器（id、text、contentDesc、className、xpath）对 UI 元素执行点击、长按、输入文本、滑动、等待和断言操作。

### 可视化 Workflow 引擎
- **节点式编辑器**：通过基于 XYFlow 的拖拽界面可视化构建自动化流程。
- **30+ 步骤类型**：点击、滑动、元素交互、应用控制、按键事件、屏幕控制、等待、ADB 命令、变量、分支、子 Workflow 和 Session 控制。
- **条件分支**：通过 exists/not_exists/text_equals/text_contains 条件创建智能流程。
- **变量与表达式**：支持 Workflow 变量和算术表达式（`{{count}} + 1`）。
- **单步调试**：暂停、逐步执行，并在每个 Workflow 步骤检查变量状态。
- **Session 集成**：在 Workflow 中启动/停止追踪 Session，生成完整的测试报告。

### ADB Shell
- **集成控制台**：直接在应用内运行原始 ADB 命令。
- **命令历史**：快速访问以前执行过的命令。

### 系统托盘
- **快速访问**：从菜单栏/系统托盘控制镜像并查看设备状态。
- **设备置顶**：将常用设备置顶，始终显示在列表和托盘菜单的最上方。
- **托盘功能**：从托盘菜单直接访问置顶设备的 Logcat、Shell 和文件管理器。
- **录制指示灯**：录制激活时，托盘显示红色圆点指示器。
- **后台运行**：保持应用在后台运行，以便即时访问。

---

## MCP 集成（Model Context Protocol）

Gaze 内置 **MCP 服务器**，暴露 50+ 工具和 5 个资源，使 AI 客户端能够通过自然语言完全控制 Android 设备。Gaze 是连接 AI 与 Android 的桥梁。

### 支持的 AI 客户端

| 客户端 | 传输方式 | 配置 |
|--------|----------|------|
| **Claude Desktop** | SSE | `claude_desktop_config.json` |
| **Claude Code (CLI)** | SSE | `.claude/settings.json` |
| **Cursor** | SSE | Cursor MCP 设置 |

### 快速设置

MCP 服务器随 Gaze 自动启动，地址为 `http://localhost:23816/mcp/sse`。

**Claude Desktop**（`claude_desktop_config.json`）：
```json
{
  "mcpServers": {
    "gaze": {
      "url": "http://localhost:23816/mcp/sse"
    }
  }
}
```

**Claude Code**：
```bash
claude mcp add gaze --transport sse http://localhost:23816/mcp/sse
```

**Cursor**：在 Cursor 的 MCP 设置中添加 MCP 服务器 URL `http://localhost:23816/mcp/sse`。

### MCP 工具（50+）

| 分类 | 工具 | 描述 |
|------|------|------|
| **设备** | `device_list`, `device_info`, `device_connect`, `device_disconnect`, `device_pair`, `device_wireless`, `device_ip` | 设备发现、连接和信息查询 |
| **CLI 工具** | `adb_execute`, `aapt_execute`, `ffmpeg_execute`, `ffprobe_execute` | 执行内置 CLI 工具（ADB、AAPT、FFmpeg、FFprobe） |
| **应用** | `app_list`, `app_info`, `app_start`, `app_stop`, `app_running`, `app_install`, `app_uninstall`, `app_clear_data` | 完整的应用生命周期管理 |
| **屏幕** | `screen_screenshot`, `screen_record_start`, `screen_record_stop`, `screen_recording_status` | 截图（base64）和录屏控制 |
| **UI 自动化** | `ui_hierarchy`, `ui_search`, `ui_tap`, `ui_swipe`, `ui_input`, `ui_resolution` | UI 检查、元素交互和输入 |
| **Session** | `session_create`, `session_end`, `session_active`, `session_list`, `session_events`, `session_stats` | Session 生命周期和 Event 查询 |
| **Workflow** | `workflow_list`, `workflow_get`, `workflow_create`, `workflow_update`, `workflow_delete`, `workflow_run`, `workflow_stop`, `workflow_pause`, `workflow_resume`, `workflow_step_next`, `workflow_status`, `workflow_execute_step` | 完整的 Workflow CRUD、执行和调试 |
| **代理** | `proxy_start`, `proxy_stop`, `proxy_status` | 网络代理控制 |
| **视频** | `video_frame`, `video_metadata`, `session_video_frame`, `session_video_info` | 视频帧提取和元数据查询 |

### MCP 资源

| URI | 描述 |
|-----|------|
| `gaze://devices` | 已连接设备列表 |
| `gaze://devices/{deviceId}` | 设备详细信息 |
| `gaze://sessions` | 活跃和近期的 Session |
| `workflow://list` | 所有已保存的 Workflow |
| `workflow://{workflowId}` | Workflow 详情（含步骤） |

### AI 能用 Gaze 做什么？

通过 MCP 集成，AI 客户端可以：
- **自动化测试**：通过自然语言指令创建和运行 UI 测试 Workflow。
- **调试问题**：截图、检查 UI 层级、读取日志、分析网络流量。
- **管理设备**：在多台设备上安装应用、传输文件、配置设置。
- **构建 Workflow**：生成包含分支逻辑和变量管理的复杂自动化 Workflow。
- **监控 Session**：通过基于 Event 的 Session 录制追踪设备行为。

---

## 内置二进制文件

本应用完全自包含，捆绑了以下工具：
- **ADB**（Android Debug Bridge）
- **Scrcpy**（屏幕镜像与录制）
- **AAPT**（Android Asset Packaging Tool）
- **FFmpeg**（视频/音频处理）
- **FFprobe**（媒体分析）

启动时，这些文件会被提取到临时目录并自动使用。您无需配置系统 PATH。

---

## 小米/Poco/红米用户重要提示

要在 Scrcpy 中启用**触摸控制**，您必须：
1. 进入**开发者选项**。
2. 启用 **USB 调试**。
3. 启用 **USB 调试（安全设置）**。
   *（注意：在大多数小米设备上，这需要插入 SIM 卡并登录小米账号）*

---

## 开始使用

### 前置条件
- **Go**（v1.23+）
- **Node.js**（v18 LTS）
- **Wails CLI**（v2.9.2）
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### 开发
```bash
wails dev
```

### 构建
```bash
wails build
```
编译后的应用将位于 `build/bin` 目录。

### 运行测试
```bash
go test ./...
```

### 发布
本项目使用 GitHub Actions 自动进行多平台构建。要创建新版本：
1. 为您的提交打标签：`git tag v1.0.0`
2. 推送标签：`git push origin v1.0.0`
GitHub Action 将自动为 macOS、Windows 和 Linux 构建，并将产物上传到 Release 页面。

---

## 架构概览

```
                    +-----------------+
                    |   Wails (GUI)   |
                    +--------+--------+
                             |
              +--------------+--------------+
              |                             |
     +--------v--------+          +--------v--------+
     |  React Frontend |          |   Go Backend    |
     |  (Ant Design,   |          |  (App, Device,  |
     |   Zustand,      |          |   Automation,   |
     |   XYFlow)       |          |   Workflow)     |
     +-----------------+          +--------+--------+
                                           |
                         +-----------------+-----------------+
                         |                 |                 |
                +--------v------+  +-------v-------+  +-----v-------+
                | Event Pipeline|  |  MCP Server   |  |   Proxy     |
                | (Session,     |  |  (50+ tools,  |  |  (MITM,     |
                |  SQLite,      |  |   5 resources)|  |   goproxy)  |
                |  FTS5)        |  +---------------+  +-------------+
                +---------------+
```

---

## 技术栈

| 层级 | 技术 |
|------|------|
| **桌面框架** | Wails v2 |
| **后端** | Go 1.23+ |
| **前端** | React 18, TypeScript, Ant Design 6 |
| **状态管理** | Zustand |
| **Workflow 编辑器** | XYFlow + Dagre |
| **数据库** | SQLite（WAL 模式，FTS5） |
| **代理** | goproxy |
| **MCP** | mcp-go（Model Context Protocol） |
| **国际化** | i18next（5 种语言） |
| **日志** | zerolog |
| **图表** | Recharts |

---

## 故障排除

### macOS："应用已损坏，无法打开"
如果您从 GitHub 下载应用并看到 *"Gaze.app 已损坏，无法打开"* 的错误提示，这是由于 macOS Gatekeeper 的隔离机制导致的。

要解决此问题，请在终端中运行以下命令：
```bash
sudo xattr -cr /path/to/Gaze.app
```
*（请将 `/path/to/Gaze.app` 替换为您下载应用的实际路径）*

> **或者自行构建：** 如果您不想绕过 Gatekeeper，可以轻松地[从源码构建应用](#开始使用)。只需几分钟即可完成！

### Windows："Windows 已保护你的电脑"
如果看到蓝色的 SmartScreen 弹窗阻止应用启动：
1. 点击**更多信息**。
2. 点击**仍要运行**。

---

## 许可证
本项目采用 MIT 许可证。
