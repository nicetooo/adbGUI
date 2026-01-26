# Gaze

**Wails**, **React**, **Ant Design**으로 구축된 강력하고 현대적인 독립형 Android 기기 관리 및 자동화 도구입니다. 완전한 기기 동작 추적을 위한 통합 **Session-Event** 아키텍처, 테스트 자동화를 위한 비주얼 **Workflow** 엔진, AI 기반 기기 제어를 위한 완전한 **MCP** (Model Context Protocol) 통합을 제공합니다.

> **참고**: 이 애플리케이션은 순수한 **vibecoding**의 결과물입니다.

[English](README.md) | [简体中文](locales/README.zh-CN.md) | [繁體中文](locales/README.zh-TW.md) | [日本語](locales/README.ja.md) | [한국어](locales/README.ko.md) | [Español](locales/README.es.md) | [Português](locales/README.pt.md) | [Français](locales/README.fr.md) | [Deutsch](locales/README.de.md) | [Русский](locales/README.ru.md) | [Tiếng Việt](locales/README.vi.md) | [العربية](locales/README.ar.md)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)
[![Website](https://img.shields.io/website?up_message=online&url=https%3A%2F%2Fgaze.nicetooo.com)](https://gaze.nicetooo.com)

## Gaze를 선택해야 하는 이유

- **현대적이고 빠름**: Wails (Go + React)로 구축되어 최소한의 리소스 오버헤드로 네이티브에 가까운 경험을 제공합니다.
- **완전한 독립형**: 시스템에 `adb`, `scrcpy`, `aapt`, `ffmpeg`, `ffprobe`를 설치할 필요가 없습니다. 모든 것이 번들로 제공되어 바로 사용할 수 있습니다.
- **안정적인 파일 전송**: macOS에서 종종 불안정한 *Android File Transfer*의 견고한 대안입니다.
- **다중 기기 지원**: 여러 기기의 독립적인 동시 백그라운드 녹화를 지원합니다.
- **Session-Event 아키텍처**: 모든 기기 활동(로그, 네트워크, 터치, 앱 수명 주기)을 단일 타임라인에서 통합 추적합니다.
- **비주얼 Workflow 자동화**: 드래그 앤 드롭 노드 편집기로 복잡한 테스트 플로우를 구축합니다 -- 코드가 필요 없습니다.
- **MCP를 통한 AI 지원**: Model Context Protocol을 통해 50개 이상의 도구를 제공하여 Claude Desktop 및 Cursor와 같은 AI 클라이언트와 원활하게 통합됩니다.
- **개발자 우선**: 개발자를 위해, 개발자가 설계한 통합 Logcat, Shell, MITM Proxy 및 UI Inspector를 제공합니다.

## 앱 스크린샷

| 기기 미러링 | 파일 관리자 |
|:---:|:---:|
| <img src="screenshots/mirror.png" width="400" /> | <img src="screenshots/files.png" width="400" /> |

| 앱 관리 | Logcat 뷰어 |
|:---:|:---:|
| <img src="screenshots/apps.png" width="400" /> | <img src="screenshots/logcat.png" width="400" /> |

| ADB Shell | 시스템 트레이 |
|:---:|:---:|
| <img src="screenshots/shell.png" width="400" /> | <img src="screenshots/tray.png" width="400" /> |

| 프록시 및 네트워크 |
|:---:|
| <img src="screenshots/proxy.png" width="820" /> |

---

## 기능

### 기기 관리
- **통합 기기 목록**: 자동 USB/Wi-Fi 병합으로 물리적 기기와 무선 기기를 원활하게 관리합니다.
- **무선 연결**: mDNS 지원과 함께 IP/포트 페어링을 통해 간편하게 연결합니다.
- **기기 기록**: 이전에 연결된 오프라인 기기에 빠르게 액세스합니다.
- **기기 고정**: 자주 사용하는 기기를 목록 상단에 고정합니다.
- **기기 모니터링**: 배터리, 네트워크 및 화면 상태 변경을 실시간으로 추적합니다.
- **일괄 작업**: 여러 기기에서 동시에 작업을 실행합니다.

### 앱 관리
- **완벽한 패키지 제어**: 설치(드래그 앤 드롭), 제거, 활성화, 비활성화, 강제 중지, 데이터 삭제.
- **APK 관리**: 설치된 APK 내보내기, 일괄 설치.
- **스마트 필터링**: 시스템/사용자 앱 검색 및 필터링.
- **빠른 작업**: 앱을 실행하거나 해당 앱의 로그로 직접 이동합니다.

### 화면 미러링 (Scrcpy)
- **고성능**: Scrcpy 기반의 저지연 미러링.
- **녹화**: 여러 기기의 동시 녹화를 지원하는 독립적인 백그라운드 녹화 및 원클릭 폴더 액세스.
- **오디오 포워딩**: 기기 오디오를 컴퓨터로 스트리밍 (Android 11+).
- **사용자 지정**: 해상도, 비트레이트, FPS, 코덱(H.264/H.265) 조정.
- **제어**: 멀티 터치 지원, 화면 켜짐 유지, 화면 끄기 모드.

### 파일 관리
- **다기능 탐색기**: 탐색, 복사, 잘라내기, 붙여넣기, 이름 바꾸기, 삭제, 폴더 생성.
- **드래그 앤 드롭**: 파일을 창으로 드래그하여 업로드합니다.
- **다운로드**: 기기에서 컴퓨터로 간편하게 파일을 전송합니다.
- **미리보기**: 호스트 컴퓨터에서 파일을 직접 엽니다.

### 고급 Logcat
- **실시간 스트리밍**: 자동 스크롤 제어 기능이 있는 실시간 로그 뷰어.
- **강력한 필터링**: 로그 레벨, 태그, PID 또는 사용자 지정 정규식으로 필터링.
- **앱 중심**: 특정 애플리케이션의 로그를 자동으로 필터링합니다.
- **JSON 포맷팅**: 탐지된 JSON 로그 세그먼트를 보기 좋게 출력합니다.

### 네트워크 및 프록시 (MITM)
- **자동화된 캡처**: 클릭 한 번으로 HTTP/HTTPS 프록시 서버를 시작하고 ADB를 통해 기기 프록시 설정을 자동 구성합니다.
- **HTTPS 복호화 (MITM)**: 자동 CA 인증서 생성 및 배포를 지원하여 SSL 트래픽을 복호화합니다.
- **WebSocket 지원**: 실시간 WebSocket 트래픽을 캡처하고 검사합니다.
- **대용량 데이터 처리**: 최대 100MB의 전체 본문 캡처를 지원하며, 5000개의 로그 버퍼를 제공합니다.
- **트래픽 쉐이핑**: 기기별 다운로드/업로드 대역폭 제한 및 인공 지연 설정을 통해 실제 네트워크 환경을 시뮬레이션합니다.
- **시각적 지표**: 선택된 기기의 RX/TX 속도를 실시간으로 모니터링합니다.

### Session 및 Event 추적
- **통합 Event Pipeline**: 모든 기기 활동(로그, 네트워크 요청, 터치 이벤트, 앱 수명 주기, 어설션)이 이벤트로 캡처되어 Session 타임라인에 연결됩니다.
- **자동 Session 관리**: 이벤트 발생 시 자동으로 Session이 생성되거나, 사용자 지정 설정(logcat, 녹화, 프록시, 모니터링)으로 수동 생성할 수 있습니다.
- **Event 타임라인**: 시간 기반 인덱싱 및 탐색으로 모든 이벤트를 다중 레인에서 시각화합니다.
- **전문 검색**: SQLite FTS5를 사용하여 모든 이벤트를 검색합니다.
- **백프레셔 제어**: 고부하 시 자동 이벤트 샘플링으로 중요 이벤트(오류, 네트워크, Workflow)를 보호합니다.
- **Event 어설션**: 자동화된 검증을 위해 이벤트 스트림에 대한 어설션을 정의하고 평가합니다.
- **비디오 동기화**: 시각적 디버깅을 위해 이벤트 타임스탬프에 동기화된 비디오 프레임을 추출합니다.

### UI Inspector 및 자동화
- **UI 계층 구조 Inspector**: 모든 화면의 전체 UI 트리를 탐색하고 분석합니다.
- **Element Picker**: UI 요소를 클릭하여 선택하고 속성(resource-id, text, bounds, class)을 검사합니다.
- **터치 녹화**: 터치 인터랙션을 녹화하고 자동화 스크립트로 재생합니다.
- **Element 기반 작업**: 선택자(id, text, contentDesc, className, xpath)를 사용하여 UI 요소에 대해 클릭, 롱클릭, 텍스트 입력, 스와이프, 대기, 어설션을 수행합니다.

### 비주얼 Workflow 엔진
- **노드 기반 편집기**: XYFlow 기반의 드래그 앤 드롭 인터페이스로 자동화 플로우를 시각적으로 구축합니다.
- **30개 이상의 단계 유형**: 탭, 스와이프, 요소 인터랙션, 앱 제어, 키 이벤트, 화면 제어, 대기, ADB 명령, 변수, 분기, 하위 Workflow, Session 제어.
- **조건 분기**: exists/not_exists/text_equals/text_contains 조건으로 지능적인 플로우를 생성합니다.
- **변수 및 수식**: 산술 수식 지원과 함께 Workflow 변수를 사용합니다 (`{{count}} + 1`).
- **단계별 디버깅**: 각 Workflow 단계에서 일시 중지, 단계별 실행, 변수 상태 검사가 가능합니다.
- **Session 통합**: 포괄적인 테스트 보고를 위해 Workflow 내에서 추적 Session을 시작/중지합니다.

### ADB Shell
- **통합 콘솔**: 앱 내에서 직접 ADB 명령을 실행합니다.
- **명령 기록**: 이전에 실행한 명령에 빠르게 액세스합니다.

### 시스템 트레이
- **빠른 액세스**: 메뉴 바/시스템 트레이에서 미러링을 제어하고 기기 상태를 확인합니다.
- **기기 고정**: 기본 기기를 목록 및 트레이 메뉴 상단에 고정합니다.
- **트레이 기능**: 트레이에서 고정된 기기의 Logcat, Shell, 파일 관리자에 직접 액세스합니다.
- **녹화 표시기**: 녹화가 활성화되면 트레이에 빨간색 점 표시기가 나타납니다.
- **백그라운드 작동**: 즉각적인 액세스를 위해 앱을 백그라운드에서 실행 상태로 유지합니다.

---

## MCP 통합 (Model Context Protocol)

Gaze에는 50개 이상의 도구와 5개의 리소스를 제공하는 내장 **MCP 서버**가 포함되어 있어, AI 클라이언트가 자연어를 통해 Android 기기를 완전히 제어할 수 있습니다. 이를 통해 Gaze는 AI와 Android 사이의 다리 역할을 합니다.

### 지원되는 AI 클라이언트

| 클라이언트 | 전송 방식 | 설정 |
|--------|-----------|---------------|
| **Claude Desktop** | SSE | `claude_desktop_config.json` |
| **Claude Code (CLI)** | SSE | `.claude/settings.json` |
| **Cursor** | SSE | Cursor MCP 설정 |

### 빠른 설정

MCP 서버는 Gaze와 함께 `http://localhost:23816/mcp/sse`에서 자동으로 시작됩니다.

**Claude Desktop** (`claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "gaze": {
      "url": "http://localhost:23816/mcp/sse"
    }
  }
}
```

**Claude Code**:
```bash
claude mcp add gaze --transport sse http://localhost:23816/mcp/sse
```

**Cursor**: Cursor의 MCP 설정에서 MCP 서버 URL `http://localhost:23816/mcp/sse`를 추가합니다.

### MCP 도구 (50개 이상)

| 카테고리 | 도구 | 설명 |
|----------|-------|-------------|
| **Device** | `device_list`, `device_info`, `device_connect`, `device_disconnect`, `device_pair`, `device_wireless`, `device_ip` | 기기 검색, 연결 및 정보 |
| **CLI Tools** | `adb_execute`, `aapt_execute`, `ffmpeg_execute`, `ffprobe_execute` | 번들된 CLI 도구 실행 (ADB, AAPT, FFmpeg, FFprobe) |
| **Apps** | `app_list`, `app_info`, `app_start`, `app_stop`, `app_running`, `app_install`, `app_uninstall`, `app_clear_data` | 완전한 애플리케이션 수명 주기 관리 |
| **Screen** | `screen_screenshot`, `screen_record_start`, `screen_record_stop`, `screen_recording_status` | 스크린샷 (base64) 및 녹화 제어 |
| **UI Automation** | `ui_hierarchy`, `ui_search`, `ui_tap`, `ui_swipe`, `ui_input`, `ui_resolution` | UI 검사, 요소 인터랙션 및 입력 |
| **Sessions** | `session_create`, `session_end`, `session_active`, `session_list`, `session_events`, `session_stats` | Session 수명 주기 및 이벤트 쿼리 |
| **Workflows** | `workflow_list`, `workflow_get`, `workflow_create`, `workflow_update`, `workflow_delete`, `workflow_run`, `workflow_stop`, `workflow_pause`, `workflow_resume`, `workflow_step_next`, `workflow_status`, `workflow_execute_step` | 완전한 Workflow CRUD, 실행 및 디버깅 |
| **Proxy** | `proxy_start`, `proxy_stop`, `proxy_status` | 네트워크 프록시 제어 |
| **Video** | `video_frame`, `video_metadata`, `session_video_frame`, `session_video_info` | 비디오 프레임 추출 및 메타데이터 |

### MCP 리소스

| URI | 설명 |
|-----|-------------|
| `gaze://devices` | 연결된 기기 목록 |
| `gaze://devices/{deviceId}` | 상세 기기 정보 |
| `gaze://sessions` | 활성 및 최근 Session |
| `workflow://list` | 저장된 모든 Workflow |
| `workflow://{workflowId}` | 단계를 포함한 Workflow 상세 정보 |

### AI가 Gaze로 할 수 있는 것

MCP 통합을 통해 AI 클라이언트는 다음을 수행할 수 있습니다:
- **테스트 자동화**: 자연어 지시를 통해 UI 테스트 Workflow를 생성하고 실행합니다.
- **문제 디버깅**: 스크린샷을 찍고, UI 계층 구조를 검사하고, 로그를 읽고, 네트워크 트래픽을 분석합니다.
- **기기 관리**: 여러 기기에 걸쳐 앱을 설치하고, 파일을 전송하고, 설정을 구성합니다.
- **Workflow 구축**: 분기 로직과 변수 관리를 포함한 복잡한 자동화 Workflow를 생성합니다.
- **Session 모니터링**: 이벤트 기반 Session 녹화로 시간에 따른 기기 동작을 추적합니다.

---

## 내장 바이너리

이 애플리케이션은 완전한 독립형입니다. 다음을 번들로 포함합니다:
- **ADB** (Android Debug Bridge)
- **Scrcpy** (화면 미러링 및 녹화)
- **AAPT** (Android Asset Packaging Tool)
- **FFmpeg** (비디오/오디오 처리)
- **FFprobe** (미디어 분석)

시작 시 이러한 파일은 임시 디렉토리에 추출되어 자동으로 사용됩니다. 시스템 PATH를 구성할 필요가 없습니다.

---

## Xiaomi/Poco/Redmi 사용자를 위한 중요 사항

Scrcpy에서 **터치 제어**를 활성화하려면 다음을 수행해야 합니다:
1. **개발자 옵션**으로 이동합니다.
2. **USB 디버깅**을 활성화합니다.
3. **USB 디버깅(보안 설정)**을 활성화합니다.
   *(참고: 대부분의 Xiaomi 기기에서는 SIM 카드와 Mi 계정 로그인이 필요합니다).*

---

## 시작하기

### 필수 조건
- **Go** (v1.23+)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### 개발
```bash
wails dev
```

### 빌드
```bash
wails build
```
컴파일된 애플리케이션은 `build/bin`에서 확인할 수 있습니다.

### 테스트 실행
```bash
go test ./...
```

### 릴리스
이 프로젝트는 GitHub Actions를 사용하여 다중 플랫폼 빌드를 자동화합니다. 새 릴리스를 만들려면:
1. 커밋에 태그 달기: `git tag v1.0.0`
2. 태그 푸시: `git push origin v1.0.0`
GitHub Action은 macOS, Windows 및 Linux용으로 자동 빌드하고 아티팩트를 릴리스 페이지에 업로드합니다.

---

## 아키텍처 개요

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

## 기술 스택

| 계층 | 기술 |
|-------|-----------|
| **데스크톱 프레임워크** | Wails v2 |
| **백엔드** | Go 1.23+ |
| **프론트엔드** | React 18, TypeScript, Ant Design 6 |
| **상태 관리** | Zustand |
| **Workflow 편집기** | XYFlow + Dagre |
| **데이터베이스** | SQLite (WAL mode, FTS5) |
| **프록시** | goproxy |
| **MCP** | mcp-go (Model Context Protocol) |
| **i18n** | i18next (5개 언어) |
| **로깅** | zerolog |
| **차트** | Recharts |

---

## 문제 해결

### macOS: "앱이 손상되었기 때문에 열 수 없습니다"
GitHub에서 앱을 다운로드할 때 *"Gaze.app이(가) 손상되었기 때문에 열 수 없습니다"* 오류가 표시된다면, 이는 macOS Gatekeeper 격리 기능 때문입니다.

이 문제를 해결하려면 터미널에서 다음 명령어를 실행하세요:
```bash
sudo xattr -cr /path/to/Gaze.app
```
*(`/path/to/Gaze.app`을 실제 다운로드한 애플리케이션 경로로 변경하세요)*

> **직접 빌드하기:** Gatekeeper를 우회하고 싶지 않다면, 로컬에서 [소스 코드로 앱을 빌드](#시작하기)할 수 있습니다. 몇 분이면 충분합니다!

### Windows: "Windows의 PC 보호"
파란색 SmartScreen 팝업이 앱 실행을 차단하는 경우:
1. **추가 정보**를 클릭합니다.
2. **실행**을 클릭합니다.

---

## 라이선스
이 프로젝트는 MIT 라이선스에 따라 라이선스가 부여됩니다.
