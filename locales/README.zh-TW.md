# Gaze

一個強大、現代且自包含的 Android 裝置管理與自動化工具，基於 **Wails**、**React** 和 **Ant Design** 構建。採用統一的 **Session-Event** 架構實現完整的裝置行為追蹤，搭配視覺化 **Workflow** 引擎進行測試自動化，並完整整合 **MCP**（Model Context Protocol）以實現 AI 驅動的裝置控制。


[English](README.md) | [简体中文](locales/README.zh-CN.md) | [繁體中文](locales/README.zh-TW.md) | [日本語](locales/README.ja.md) | [한국어](locales/README.ko.md) | [Español](locales/README.es.md) | [Português](locales/README.pt.md) | [Français](locales/README.fr.md) | [Deutsch](locales/README.de.md) | [Русский](locales/README.ru.md) | [Tiếng Việt](locales/README.vi.md) | [العربية](locales/README.ar.md)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)
[![Website](https://img.shields.io/website?up_message=online&url=https%3A%2F%2Fgaze.nicetooo.com)](https://gaze.nicetooo.com)

## 為何選擇 Gaze？

- **現代且快速**：基於 Wails（Go + React）構建，提供近似原生的體驗，資源消耗極低。
- **真正自包含**：無需在系統上安裝 `adb`、`scrcpy`、`aapt`、`ffmpeg` 或 `ffprobe`。所有工具皆已內建，開箱即用。
- **可靠的檔案傳輸**：macOS 上經常不穩定的 *Android File Transfer* 的可靠替代方案。
- **多裝置支援**：支援多裝置同時獨立背景錄影。
- **Session-Event 架構**：在單一時間軸上統一追蹤所有裝置活動（日誌、網路、觸控、應用程式生命週期）。
- **視覺化 Workflow 自動化**：透過拖放式節點編輯器構建複雜的測試流程——無需撰寫程式碼。
- **透過 MCP 支援 AI**：超過 50 個工具透過 Model Context Protocol 公開，可與 Claude Desktop 和 Cursor 等 AI 客戶端無縫整合。
- **開發者優先**：整合 Logcat、Shell、MITM 代理和 UI 檢查器，由開發者設計，為開發者打造。

## 應用程式截圖

| 裝置管理 | 螢幕投射 |
|:---:|:---:|
| <img src="screenshots/devices.png" width="400" /> | <img src="screenshots/mirror.png" width="400" /> |

| 檔案管理 | 應用程式管理 |
|:---:|:---:|
| <img src="screenshots/files.png" width="400" /> | <img src="screenshots/apps.png" width="400" /> |

| 效能監控 | Session 時間線 |
|:---:|:---:|
| <img src="screenshots/performance.png" width="400" /> | <img src="screenshots/session.png" width="400" /> |

| Session 列表 | Logcat 檢視器 |
|:---:|:---:|
| <img src="screenshots/session-list.png" width="400" /> | <img src="screenshots/logcat.png" width="400" /> |

| 視覺化工作流編輯器 | 工作流列表 |
|:---:|:---:|
| <img src="screenshots/workflow-editor.png" width="400" /> | <img src="screenshots/workflow.png" width="400" /> |

| UI 檢查器 | 觸控錄製 |
|:---:|:---:|
| <img src="screenshots/ui-inspector.png" width="400" /> | <img src="screenshots/recording.png" width="400" /> |

| 網路代理 (MITM) | ADB Shell |
|:---:|:---:|
| <img src="screenshots/proxy.png" width="400" /> | <img src="screenshots/shell.png" width="400" /> |

---

## 功能特性

### 裝置管理
- **統一裝置列表**：無縫管理實體和無線裝置，支援 USB/Wi-Fi 自動合併。
- **無線連接**：透過 IP/連接埠配對輕鬆連接，支援 mDNS。
- **裝置記錄**：快速存取之前連接過的離線裝置。
- **裝置置頂**：將常用裝置置頂，始終顯示在列表最上方。
- **裝置監控**：即時追蹤電池、網路和螢幕狀態變化。
- **批次操作**：同時在多台裝置上執行操作。

### 應用程式管理
- **完整套件控制**：安裝（拖放）、解除安裝、啟用、停用、強制停止、清除資料。
- **APK 管理**：匯出已安裝的 APK、批次安裝。
- **智慧過濾**：按系統/使用者應用程式進行搜尋和過濾。
- **快速操作**：啟動應用程式或直接跳轉至其日誌。

### 螢幕投屏 (Scrcpy)
- **高效能**：由 Scrcpy 驅動的低延遲投屏。
- **錄影**：獨立背景錄製，支援多裝置同時錄製，一鍵開啟資料夾。
- **音訊轉發**：將裝置音訊串流傳輸至電腦（Android 11+）。
- **自訂設定**：調整解析度、位元率、FPS 和編解碼器（H.264/H.265）。
- **控制功能**：多點觸控支援、保持喚醒、關閉螢幕模式。

### 檔案管理
- **全功能檔案瀏覽器**：瀏覽、複製、剪下、貼上、重新命名、刪除和建立資料夾。
- **拖放上傳**：只需將檔案拖曳至視窗即可上傳。
- **下載檔案**：輕鬆將檔案從裝置傳輸至電腦。
- **預覽功能**：直接在主機上開啟檔案。

### 進階 Logcat
- **即時串流**：帶有自動捲動控制的即時日誌檢視器。
- **強大過濾**：按日誌等級、Tag、PID 或自訂正規表示式過濾。
- **以應用程式為中心**：自動過濾特定應用程式的日誌。
- **JSON 格式化**：自動美化偵測到的 JSON 日誌片段。

### 網路與代理 (MITM)
- **自動化抓包**：一鍵啟動 HTTP/HTTPS 代理伺服器，並透過 ADB 自動配置裝置代理設定。
- **HTTPS 解密 (MITM)**：支援 SSL 流量解密，自動產生和部署 CA 憑證。
- **WebSocket 支援**：即時捕獲並檢視 WebSocket 傳輸內容。
- **大資料處理**：支援捕獲高達 100MB 的完整封包內容而不截斷，並擁有 5000 筆日誌緩衝區。
- **流量整形**：模擬真實網路環境，支援按裝置限制下行/上行頻寬以及設定人工延遲。
- **視覺化指標**：即時監控所選裝置的 RX/TX 速率。

### Session 與 Event 追蹤
- **統一 Event Pipeline**：所有裝置活動（日誌、網路請求、觸控事件、應用程式生命週期、斷言）均被捕獲為事件並連結至 Session 時間軸。
- **自動 Session 管理**：當事件發生時自動建立 Session，或手動建立帶有自訂設定（logcat、錄影、代理、監控）的 Session。
- **Event 時間軸**：多通道視覺化呈現所有事件，支援基於時間的索引和導覽。
- **全文搜尋**：使用 SQLite FTS5 在所有事件中進行搜尋。
- **背壓控制**：高負載時自動事件採樣，同時保護關鍵事件（錯誤、網路、工作流程）。
- **Event 斷言**：定義並評估針對事件流的斷言，實現自動化驗證。
- **影片同步**：擷取與事件時間戳同步的影片畫面，用於視覺化除錯。

### UI 檢查器與自動化
- **UI 階層檢查器**：瀏覽和分析任何畫面的完整 UI 樹狀結構。
- **元素選取器**：點擊選取 UI 元素並檢視其屬性（resource-id、text、bounds、class）。
- **觸控錄製**：錄製觸控互動並將其重播為自動化腳本。
- **基於元素的操作**：使用選擇器（id、text、contentDesc、className、xpath）對 UI 元素執行點擊、長按、輸入文字、滑動、等待和斷言操作。

### 視覺化 Workflow 引擎
- **節點式編輯器**：透過由 XYFlow 驅動的拖放介面視覺化建構自動化流程。
- **30+ 步驟類型**：點擊、滑動、元素互動、應用程式控制、按鍵事件、螢幕控制、等待、ADB 指令、變數、條件分支、子工作流程和 Session 控制。
- **條件分支**：使用 exists/not_exists/text_equals/text_contains 條件建立智慧化流程。
- **變數與表達式**：使用工作流程變數搭配算術表達式支援（`{{count}} + 1`）。
- **逐步除錯**：暫停、逐步執行並在每個工作流程步驟檢視變數狀態。
- **Session 整合**：在工作流程中啟動/停止追蹤 Session，實現完整的測試報告。

### ADB Shell
- **整合控制台**：直接在應用程式內執行原始 ADB 指令。
- **指令記錄**：快速存取之前執行過的指令。

### 系統匣
- **快速存取**：從選單列/系統匣控制投屏並查看裝置狀態。
- **裝置置頂**：將常用裝置置頂，使其顯示在列表和系統匣選單的最上方。
- **系統匣功能**：從系統匣直接存取置頂裝置的 Logcat、Shell 和檔案管理器。
- **錄影指示燈**：錄影啟動時，系統匣圖示顯示紅色圓點狀態。
- **背景運行**：保持應用程式在背景運行，以便即時存取。

---

## MCP 整合（Model Context Protocol）

Gaze 內建 **MCP 伺服器**，公開超過 50 個工具和 5 個資源，使 AI 客戶端能透過自然語言完整控制 Android 裝置。這讓 Gaze 成為 AI 與 Android 之間的橋樑。

### 支援的 AI 客戶端

| 客戶端 | 傳輸方式 | 設定方式 |
|--------|-----------|---------------|
| **Claude Desktop** | SSE | `claude_desktop_config.json` |
| **Claude Code (CLI)** | SSE | `.claude/settings.json` |
| **Cursor** | SSE | Cursor MCP 設定 |

### 快速設定

MCP 伺服器會隨 Gaze 自動啟動，地址為 `http://localhost:23816/mcp/sse`。

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

**Cursor**：在 Cursor 的 MCP 設定中新增 MCP 伺服器 URL `http://localhost:23816/mcp/sse`。

### MCP 工具（50+）

| 分類 | 工具 | 說明 |
|----------|-------|-------------|
| **裝置** | `device_list`, `device_info`, `device_connect`, `device_disconnect`, `device_pair`, `device_wireless`, `device_ip` | 裝置發現、連接和資訊查詢 |
| **CLI 工具** | `adb_execute`, `aapt_execute`, `ffmpeg_execute`, `ffprobe_execute` | 執行內建 CLI 工具（ADB、AAPT、FFmpeg、FFprobe） |
| **應用程式** | `app_list`, `app_info`, `app_start`, `app_stop`, `app_running`, `app_install`, `app_uninstall`, `app_clear_data` | 完整的應用程式生命週期管理 |
| **螢幕** | `screen_screenshot`, `screen_record_start`, `screen_record_stop`, `screen_recording_status` | 截圖（base64）和錄影控制 |
| **UI 自動化** | `ui_hierarchy`, `ui_search`, `ui_tap`, `ui_swipe`, `ui_input`, `ui_resolution` | UI 檢查、元素互動和輸入 |
| **Sessions** | `session_create`, `session_end`, `session_active`, `session_list`, `session_events`, `session_stats` | Session 生命週期和事件查詢 |
| **Workflows** | `workflow_list`, `workflow_get`, `workflow_create`, `workflow_update`, `workflow_delete`, `workflow_run`, `workflow_stop`, `workflow_pause`, `workflow_resume`, `workflow_step_next`, `workflow_status`, `workflow_execute_step` | 完整的 Workflow CRUD、執行和除錯 |
| **代理** | `proxy_start`, `proxy_stop`, `proxy_status` | 網路代理控制 |
| **影片** | `video_frame`, `video_metadata`, `session_video_frame`, `session_video_info` | 影片畫面擷取和中繼資料 |

### MCP 資源

| URI | 說明 |
|-----|-------------|
| `gaze://devices` | 已連接的裝置列表 |
| `gaze://devices/{deviceId}` | 詳細裝置資訊 |
| `gaze://sessions` | 活躍和最近的 Sessions |
| `workflow://list` | 所有已儲存的 Workflows |
| `workflow://{workflowId}` | 含步驟的 Workflow 詳細資訊 |

### AI 能透過 Gaze 做什麼？

透過 MCP 整合，AI 客戶端可以：
- **自動化測試**：透過自然語言指令建立和執行 UI 測試工作流程。
- **除錯問題**：截圖、檢查 UI 階層、讀取日誌和分析網路流量。
- **管理裝置**：安裝應用程式、傳輸檔案、跨多台裝置配置設定。
- **建構 Workflows**：產生具備條件分支邏輯和變數管理的複雜自動化工作流程。
- **監控 Sessions**：透過基於事件的 Session 錄製追蹤裝置行為。

---

## 內建執行檔

本應用完全自包含，內建以下工具：
- **ADB**（Android Debug Bridge）
- **Scrcpy**（螢幕投屏與錄影）
- **AAPT**（Android Asset Packaging Tool）
- **FFmpeg**（影音處理）
- **FFprobe**（媒體分析）

啟動時，這些工具會被解壓至臨時目錄並自動使用。您無需配置系統 PATH。

---

## 小米/Poco/紅米使用者重要提示

要在 Scrcpy 中啟用**觸控控制**，您必須：
1. 進入**開發者選項**。
2. 啟用 **USB 偵錯**。
3. 啟用 **USB 偵錯（安全設定）**。
   *（注意：在大多數小米裝置上，這需要插入 SIM 卡並登入小米帳號）*。

---

## 開始使用

### 前置條件
- **Go**（v1.23+）
- **Node.js**（v18 LTS）
- **Wails CLI**（v2.9.2）
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### 開發
```bash
wails dev -tags fts5
```

### 建構
```bash
wails build -tags fts5
```
編譯後的應用將位於 `build/bin` 目錄。

### 執行測試
```bash
go test -tags fts5 ./...
```

### 發佈
本專案使用 GitHub Actions 自動進行多平台建構。要建立新版本：
1. 為您的提交打標籤：`git tag v1.0.0`
2. 推送標籤：`git push origin v1.0.0`
GitHub Action 將自動為 macOS、Windows 和 Linux 建構，並將產物上傳至 Release 頁面。

---

## 架構概覽

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

## 技術棧

| 層級 | 技術 |
|-------|-----------|
| **桌面框架** | Wails v2 |
| **後端** | Go 1.23+ |
| **前端** | React 18, TypeScript, Ant Design 6 |
| **狀態管理** | Zustand |
| **Workflow 編輯器** | XYFlow + Dagre |
| **資料庫** | SQLite（WAL 模式、FTS5） |
| **代理** | goproxy |
| **MCP** | mcp-go（Model Context Protocol） |
| **國際化** | i18next（5 種語言） |
| **日誌** | zerolog |
| **圖表** | Recharts |

---

## 疑難排解

### macOS：「應用程式已損毀，無法打開」
如果您從 GitHub 下載應用程式並看到 *「Gaze.app 已損毀，無法打開」* 的錯誤提示，這是由於 macOS Gatekeeper 的隔離機制所導致。

要解決此問題，請在終端機中執行以下指令：
```bash
sudo xattr -cr /path/to/Gaze.app
```
*（請將 `/path/to/Gaze.app` 替換為您下載應用程式的實際路徑）*

> **或者自行建構：** 如果您不想繞過 Gatekeeper，可以輕鬆地[從原始碼建構應用](#開始使用)。只需幾分鐘即可完成！

### Windows：「Windows 已保護您的電腦」
如果看到藍色的 SmartScreen 快顯視窗阻止應用程式啟動：
1. 點擊**其他資訊**。
2. 點擊**仍要執行**。

---

## 授權條款
本專案採用 MIT 授權條款。
