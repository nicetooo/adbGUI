# Gaze

Một công cụ quản lý và tự động hóa thiết bị Android mạnh mẽ, hiện đại và hoàn toàn độc lập, được xây dựng bằng **Wails**, **React** và **Ant Design**. Tích hợp kiến trúc **Session-Event** thống nhất để theo dõi toàn diện hành vi thiết bị, công cụ **Workflow** trực quan để tự động hóa kiểm thử, và tích hợp đầy đủ **MCP** (Model Context Protocol) để điều khiển thiết bị bằng AI.


[English](README.md) | [简体中文](locales/README.zh-CN.md) | [繁體中文](locales/README.zh-TW.md) | [日本語](locales/README.ja.md) | [한국어](locales/README.ko.md) | [Español](locales/README.es.md) | [Português](locales/README.pt.md) | [Français](locales/README.fr.md) | [Deutsch](locales/README.de.md) | [Русский](locales/README.ru.md) | [Tiếng Việt](locales/README.vi.md) | [العربية](locales/README.ar.md)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)
[![Website](https://img.shields.io/website?up_message=online&url=https%3A%2F%2Fgaze.nicetooo.com)](https://gaze.nicetooo.com)

## Tại sao chọn Gaze?

- **Hiện đại & Nhanh**: Xây dựng bằng Wails (Go + React), mang lại trải nghiệm gần như ứng dụng gốc với mức tiêu thụ tài nguyên tối thiểu.
- **Hoàn toàn Độc lập**: Không cần cài đặt `adb`, `scrcpy`, `aapt`, `ffmpeg` hay `ffprobe` trên hệ thống. Tất cả đã được tích hợp sẵn và sẵn sàng sử dụng.
- **Truyền Tệp Tin cậy**: Một giải pháp thay thế ổn định cho ứng dụng *Android File Transfer* thường gặp lỗi trên macOS.
- **Hỗ trợ Đa thiết bị**: Hỗ trợ ghi màn hình nền độc lập, đồng thời cho nhiều thiết bị cùng lúc.
- **Kiến trúc Session-Event**: Theo dõi thống nhất tất cả hoạt động thiết bị (nhật ký, mạng, cảm ứng, vòng đời ứng dụng) trên một dòng thời gian duy nhất.
- **Tự động hóa Workflow trực quan**: Xây dựng các luồng kiểm thử phức tạp bằng trình biên tập kéo thả — không cần viết mã.
- **Sẵn sàng cho AI qua MCP**: Hơn 50 công cụ được cung cấp thông qua Model Context Protocol để tích hợp liền mạch với các ứng dụng khách AI như Claude Desktop và Cursor.
- **Ưu tiên Lập trình viên**: Tích hợp Logcat, Shell, MITM Proxy và UI Inspector được thiết kế bởi lập trình viên, dành cho lập trình viên.

## Ảnh chụp Ứng dụng

| Quản lý Thiết bị | Phản chiếu Màn hình |
|:---:|:---:|
| <img src="screenshots/devices.png" width="400" /> | <img src="screenshots/mirror.png" width="400" /> |

| Quản lý Tệp | Quản lý Ứng dụng |
|:---:|:---:|
| <img src="screenshots/files.png" width="400" /> | <img src="screenshots/apps.png" width="400" /> |

| Giám sát Hiệu năng | Dòng thời gian Phiên |
|:---:|:---:|
| <img src="screenshots/performance.png" width="400" /> | <img src="screenshots/session.png" width="400" /> |

| Danh sách Phiên | Xem Logcat |
|:---:|:---:|
| <img src="screenshots/session-list.png" width="400" /> | <img src="screenshots/logcat.png" width="400" /> |

| Trình chỉnh sửa Luồng công việc | Danh sách Luồng công việc |
|:---:|:---:|
| <img src="screenshots/workflow-editor.png" width="400" /> | <img src="screenshots/workflow.png" width="400" /> |

| Trình kiểm tra UI | Ghi lại Cảm ứng |
|:---:|:---:|
| <img src="screenshots/ui-inspector.png" width="400" /> | <img src="screenshots/recording.png" width="400" /> |

| Proxy Mạng (MITM) | ADB Shell |
|:---:|:---:|
| <img src="screenshots/proxy.png" width="400" /> | <img src="screenshots/shell.png" width="400" /> |

---

## Tính năng

### Quản lý Thiết bị
- **Danh sách Thiết bị Thống nhất**: Quản lý liền mạch các thiết bị vật lý và không dây với tính năng tự động hợp nhất USB/Wi-Fi.
- **Kết nối Không dây**: Kết nối dễ dàng qua ghép nối IP/Cổng với hỗ trợ mDNS.
- **Lịch sử Thiết bị**: Truy cập nhanh vào các thiết bị ngoại tuyến đã kết nối trước đó.
- **Ghim Thiết bị**: Ghim thiết bị thường dùng nhất để luôn ở đầu danh sách.
- **Giám sát Thiết bị**: Theo dõi thời gian thực các thay đổi về pin, mạng và trạng thái màn hình.
- **Thao tác Hàng loạt**: Thực hiện các thao tác trên nhiều thiết bị cùng lúc.

### Quản lý Ứng dụng
- **Kiểm soát Gói Đầy đủ**: Cài đặt (Kéo & Thả), Gỡ cài đặt, Bật, Tắt, Buộc dừng, Xóa dữ liệu.
- **Quản lý APK**: Xuất các APK đã cài đặt, Cài đặt hàng loạt.
- **Lọc Thông minh**: Tìm kiếm và lọc theo ứng dụng Hệ thống/Người dùng.
- **Hành động Nhanh**: Khởi chạy ứng dụng hoặc chuyển trực tiếp đến nhật ký của chúng.

### Phản chiếu Màn hình (Scrcpy)
- **Hiệu suất Cao**: Phản chiếu độ trễ thấp được hỗ trợ bởi Scrcpy.
- **Ghi hình**: Ghi nền độc lập với hỗ trợ nhiều thiết bị đồng thời và truy cập thư mục bằng một cú nhấp.
- **Chuyển tiếp Âm thanh**: Truyền phát âm thanh thiết bị đến máy tính của bạn (Android 11+).
- **Tùy chỉnh**: Điều chỉnh Độ phân giải, Tốc độ bit, FPS và Codec (H.264/H.265).
- **Điều khiển**: Hỗ trợ cảm ứng đa điểm, Giữ màn hình bật, Chế độ tắt màn hình.

### Quản lý Tệp
- **Trình Khám phá Đầy đủ Tính năng**: Duyệt, Sao chép, Cắt, Dán, Đổi tên, Xóa và Tạo thư mục.
- **Kéo & Thả**: Tải tệp lên bằng cách kéo chúng vào cửa sổ.
- **Tải xuống**: Dễ dàng chuyển tệp từ thiết bị sang máy tính.
- **Xem trước**: Mở tệp trực tiếp trên máy chủ.

### Logcat Nâng cao
- **Phát trực tuyến Thời gian thực**: Trình xem nhật ký trực tiếp với điều khiển tự động cuộn.
- **Lọc Mạnh mẽ**: Lọc theo Mức nhật ký, Thẻ, PID hoặc Regex tùy chỉnh.
- **Tập trung vào Ứng dụng**: Tự động lọc nhật ký cho một ứng dụng cụ thể.
- **Định dạng JSON**: Tự động trình bày đẹp các đoạn nhật ký JSON được phát hiện.

### Mạng & Proxy (MITM)
- **Thu thập Tự động**: Một cú nhấp để khởi động máy chủ proxy HTTP/HTTPS và tự động cấu hình cài đặt proxy cho thiết bị qua ADB.
- **Giải mã HTTPS (MITM)**: Hỗ trợ giải mã lưu lượng SSL với tính năng tự động tạo và triển khai chứng chỉ CA.
- **Hỗ trợ WebSocket**: Thu thập và kiểm tra lưu lượng WebSocket thời gian thực.
- **Xử lý Dữ liệu Lớn**: Hỗ trợ thu thập toàn bộ body (lên đến 100MB) không bị cắt, với bộ đệm nhật ký 5000 mục.
- **Định hình Lưu lượng**: Mô phỏng điều kiện mạng thực tế với giới hạn băng thông Tải xuống/Tải lên theo thiết bị và độ trễ nhân tạo.
- **Chỉ số Trực quan**: Giám sát tốc độ RX/TX thời gian thực cho thiết bị được chọn.

### Session & Theo dõi Sự kiện
- **Đường ống Sự kiện Thống nhất**: Tất cả hoạt động thiết bị (nhật ký, yêu cầu mạng, sự kiện cảm ứng, vòng đời ứng dụng, kiểm định) được thu thập dưới dạng sự kiện và liên kết với dòng thời gian session.
- **Quản lý Session Tự động**: Session được tạo tự động khi sự kiện xảy ra, hoặc thủ công với cấu hình tùy chỉnh (logcat, ghi hình, proxy, giám sát).
- **Dòng thời gian Sự kiện**: Trực quan hóa đa làn tất cả sự kiện với lập chỉ mục và điều hướng theo thời gian.
- **Tìm kiếm Toàn văn**: Tìm kiếm trên tất cả sự kiện sử dụng SQLite FTS5.
- **Kiểm soát Áp lực ngược**: Tự động lấy mẫu sự kiện khi tải cao trong khi bảo vệ các sự kiện quan trọng (lỗi, mạng, workflow).
- **Kiểm định Sự kiện**: Định nghĩa và đánh giá các kiểm định đối với luồng sự kiện để xác nhận tự động.
- **Đồng bộ Video**: Trích xuất khung hình video đồng bộ với mốc thời gian sự kiện để gỡ lỗi trực quan.

### Kiểm tra UI & Tự động hóa
- **Trình Kiểm tra Phân cấp UI**: Duyệt và phân tích toàn bộ cây UI của bất kỳ màn hình nào.
- **Chọn Phần tử**: Nhấp chọn phần tử UI và kiểm tra thuộc tính của chúng (resource-id, text, bounds, class).
- **Ghi Cảm ứng**: Ghi lại các tương tác cảm ứng và phát lại chúng dưới dạng kịch bản tự động hóa.
- **Hành động dựa trên Phần tử**: Nhấp, nhấn giữ, nhập văn bản, vuốt, chờ và kiểm định trên các phần tử UI sử dụng bộ chọn (id, text, contentDesc, className, xpath).

### Công cụ Workflow Trực quan
- **Trình Biên tập dựa trên Node**: Xây dựng các luồng tự động hóa trực quan với giao diện kéo thả được hỗ trợ bởi XYFlow.
- **Hơn 30 Loại Bước**: Chạm, vuốt, tương tác phần tử, điều khiển ứng dụng, sự kiện phím, điều khiển màn hình, chờ, lệnh ADB, biến, rẽ nhánh, workflow con và điều khiển session.
- **Rẽ nhánh Có điều kiện**: Tạo các luồng thông minh với điều kiện exists/not_exists/text_equals/text_contains.
- **Biến & Biểu thức**: Sử dụng biến workflow với hỗ trợ biểu thức số học (`{{count}} + 1`).
- **Gỡ lỗi Từng bước**: Tạm dừng, bước qua và kiểm tra trạng thái biến tại mỗi bước workflow.
- **Tích hợp Session**: Bắt/dừng session theo dõi trong workflow để báo cáo kiểm thử toàn diện.

### ADB Shell
- **Console Tích hợp**: Chạy các lệnh ADB trực tiếp trong ứng dụng.
- **Lịch sử Lệnh**: Truy cập nhanh các lệnh đã thực hiện trước đó.

### Khay Hệ thống
- **Truy cập Nhanh**: Điều khiển phản chiếu và xem trạng thái thiết bị từ thanh menu/khay hệ thống.
- **Ghim Thiết bị**: Ghim thiết bị chính của bạn để xuất hiện ở đầu danh sách và menu khay.
- **Chức năng Khay**: Truy cập trực tiếp Logcat, Shell và Trình quản lý Tệp cho các thiết bị đã ghim từ khay.
- **Chỉ báo Ghi hình**: Chỉ báo trực quan "chấm đỏ" trên khay khi đang ghi hình.
- **Hoạt động Nền**: Giữ ứng dụng chạy trong nền để truy cập tức thì.

---

## Tích hợp MCP (Model Context Protocol)

Gaze bao gồm một **máy chủ MCP** tích hợp sẵn cung cấp hơn 50 công cụ và 5 tài nguyên, cho phép các ứng dụng khách AI điều khiển hoàn toàn thiết bị Android thông qua ngôn ngữ tự nhiên. Điều này biến Gaze thành cầu nối giữa AI và Android.

### Các Ứng dụng Khách AI được Hỗ trợ

| Ứng dụng khách | Giao thức | Cấu hình |
|--------|-----------|---------------|
| **Claude Desktop** | SSE | `claude_desktop_config.json` |
| **Claude Code (CLI)** | SSE | `.claude/settings.json` |
| **Cursor** | SSE | Cài đặt MCP của Cursor |

### Thiết lập Nhanh

Máy chủ MCP tự động khởi động cùng Gaze tại `http://localhost:23816/mcp/sse`.

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

**Cursor**: Thêm URL máy chủ MCP `http://localhost:23816/mcp/sse` trong cài đặt MCP của Cursor.

### Các Công cụ MCP (hơn 50)

| Danh mục | Công cụ | Mô tả |
|----------|-------|-------------|
| **Thiết bị** | `device_list`, `device_info`, `device_connect`, `device_disconnect`, `device_pair`, `device_wireless`, `device_ip` | Phát hiện, kết nối và thông tin thiết bị |
| **Công cụ CLI** | `adb_execute`, `aapt_execute`, `ffmpeg_execute`, `ffprobe_execute` | Thực thi các công cụ CLI tích hợp sẵn (ADB, AAPT, FFmpeg, FFprobe) |
| **Ứng dụng** | `app_list`, `app_info`, `app_start`, `app_stop`, `app_running`, `app_install`, `app_uninstall`, `app_clear_data` | Quản lý toàn bộ vòng đời ứng dụng |
| **Màn hình** | `screen_screenshot`, `screen_record_start`, `screen_record_stop`, `screen_recording_status` | Chụp màn hình (base64) và điều khiển ghi hình |
| **Tự động hóa UI** | `ui_hierarchy`, `ui_search`, `ui_tap`, `ui_swipe`, `ui_input`, `ui_resolution` | Kiểm tra UI, tương tác phần tử và nhập liệu |
| **Session** | `session_create`, `session_end`, `session_active`, `session_list`, `session_events`, `session_stats` | Vòng đời session và truy vấn sự kiện |
| **Workflow** | `workflow_list`, `workflow_get`, `workflow_create`, `workflow_update`, `workflow_delete`, `workflow_run`, `workflow_stop`, `workflow_pause`, `workflow_resume`, `workflow_step_next`, `workflow_status`, `workflow_execute_step` | CRUD, thực thi và gỡ lỗi workflow đầy đủ |
| **Proxy** | `proxy_start`, `proxy_stop`, `proxy_status` | Điều khiển proxy mạng |
| **Video** | `video_frame`, `video_metadata`, `session_video_frame`, `session_video_info` | Trích xuất khung hình video và metadata |

### Tài nguyên MCP

| URI | Mô tả |
|-----|-------------|
| `gaze://devices` | Danh sách các thiết bị đã kết nối |
| `gaze://devices/{deviceId}` | Thông tin chi tiết thiết bị |
| `gaze://sessions` | Các session đang hoạt động và gần đây |
| `workflow://list` | Tất cả các workflow đã lưu |
| `workflow://{workflowId}` | Chi tiết workflow với các bước |

### AI có thể làm gì với Gaze?

Với tích hợp MCP, các ứng dụng khách AI có thể:
- **Tự động hóa Kiểm thử**: Tạo và chạy các workflow kiểm thử UI thông qua hướng dẫn bằng ngôn ngữ tự nhiên.
- **Gỡ lỗi Vấn đề**: Chụp màn hình, kiểm tra phân cấp UI, đọc nhật ký và phân tích lưu lượng mạng.
- **Quản lý Thiết bị**: Cài đặt ứng dụng, truyền tệp, cấu hình cài đặt trên nhiều thiết bị.
- **Xây dựng Workflow**: Tạo các workflow tự động hóa phức tạp với logic rẽ nhánh và quản lý biến.
- **Giám sát Session**: Theo dõi hành vi thiết bị theo thời gian với ghi session dựa trên sự kiện.

---

## Binary Tích hợp Sẵn

Ứng dụng này hoàn toàn độc lập. Nó bao gồm:
- **ADB** (Android Debug Bridge)
- **Scrcpy** (Phản chiếu & ghi màn hình)
- **AAPT** (Android Asset Packaging Tool)
- **FFmpeg** (Xử lý video/âm thanh)
- **FFprobe** (Phân tích media)

Khi khởi động, các tệp này được giải nén vào một thư mục tạm thời và được sử dụng tự động. Bạn không cần cấu hình PATH hệ thống.

---

## Lưu ý Quan trọng cho Người dùng Xiaomi/Poco/Redmi

Để bật **điều khiển cảm ứng** trong Scrcpy, bạn phải:
1. Vào **Tùy chọn Nhà phát triển**.
2. Bật **Gỡ lỗi USB**.
3. Bật **Gỡ lỗi USB (Cài đặt bảo mật)**.
   *(Lưu ý: Điều này yêu cầu thẻ SIM và đăng nhập Tài khoản Mi trên hầu hết các thiết bị Xiaomi).*

---

## Bắt đầu

### Điều kiện Tiên quyết
- **Go** (v1.23+)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### Phát triển
```bash
wails dev
```

### Biên dịch
```bash
wails build
```
Ứng dụng đã biên dịch sẽ có sẵn trong `build/bin`.

### Chạy Kiểm thử
```bash
go test ./...
```

### Phát hành
Dự án này sử dụng GitHub Actions để tự động hóa việc biên dịch đa nền tảng. Để tạo một bản phát hành mới:
1. Gắn thẻ cho commit của bạn: `git tag v1.0.0`
2. Đẩy thẻ lên: `git push origin v1.0.0`
GitHub Action sẽ tự động biên dịch cho macOS, Windows và Linux, và tải các tệp lên trang Release.

---

## Tổng quan Kiến trúc

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

## Ngăn xếp Công nghệ

| Tầng | Công nghệ |
|-------|-----------|
| **Framework Desktop** | Wails v2 |
| **Backend** | Go 1.23+ |
| **Frontend** | React 18, TypeScript, Ant Design 6 |
| **Quản lý Trạng thái** | Zustand |
| **Trình Biên tập Workflow** | XYFlow + Dagre |
| **Cơ sở Dữ liệu** | SQLite (WAL mode, FTS5) |
| **Proxy** | goproxy |
| **MCP** | mcp-go (Model Context Protocol) |
| **i18n** | i18next (5 ngôn ngữ) |
| **Ghi nhật ký** | zerolog |
| **Biểu đồ** | Recharts |

---

## Khắc phục Sự cố

### macOS: "Ứng dụng bị hỏng và không thể mở được"
Nếu bạn tải xuống ứng dụng từ GitHub và gặp lỗi *"Gaze.app bị hỏng và không thể mở được"*, điều này là do tính năng cách ly Gatekeeper của macOS.

Để khắc phục điều này, hãy chạy lệnh sau trong terminal của bạn:
```bash
sudo xattr -cr /path/to/Gaze.app
```
*(Thay thế `/path/to/Gaze.app` bằng đường dẫn thực tế đến ứng dụng đã tải xuống của bạn)*

> **Hoặc tự build:** Nếu bạn không muốn bỏ qua Gatekeeper, bạn có thể dễ dàng [build ứng dụng từ mã nguồn](#bat-dau) ngay trên máy của mình. Chỉ mất vài phút!

### Windows: "Windows đã bảo vệ PC của bạn"
Nếu bạn thấy cửa sổ SmartScreen màu xanh ngăn ứng dụng khởi động:
1. Nhấp vào **Thông tin thêm**.
2. Nhấp vào **Vẫn chạy**.

---

## Giấy phép
Dự án này được cấp phép theo Giấy phép MIT.
