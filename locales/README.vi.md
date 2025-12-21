# ADB GUI 🚀

Một công cụ quản lý Android mạnh mẽ, hiện đại và độc lập được xây dựng bằng **Wails**, **React** và **Ant Design**.

> ✨ **Lưu ý**: Ứng dụng này là kết quả của quá trình **vibecoding** thuần túy.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)

## ✨ Tính năng

### 📱 Quản lý thiết bị
- Theo dõi thời gian thực các thiết bị đã kết nối.
- Xem ID thiết bị, kiểu máy và trạng thái kết nối.
- Truy cập một lần nhấp vào Ứng dụng, Shell, Logcat và Phản chiếu màn hình.

### 📦 Quản lý ứng dụng
- Liệt kê tất cả các gói đã cài đặt (ứng dụng hệ thống & người dùng).
- Lọc và tìm kiếm ứng dụng theo tên hoặc loại.
- **Hành động**: Buộc dừng, Xóa dữ liệu, Bật/Tắt và Gỡ cài đặt.
- **Logcat nhanh**: Chuyển đến nhật ký của một ứng dụng cụ thể trực tiếp từ danh sách ứng dụng.

### 🖥️ Phản chiếu màn hình (Scrcpy)
- **Tích hợp sẵn Scrcpy**: Không cần cài đặt thêm bất cứ thứ gì bên ngoài.
- Kiểm soát chi tiết về:
  - Tốc độ bit video & FPS tối đa.
  - Độ phân giải (Kích thước tối đa).
  - Tùy chọn Luôn bật & Tắt màn hình.
  - Cửa sổ luôn ở trên cùng.
  - Chuyển đổi phát trực tuyến âm thanh.

### 📜 Logcat nâng cao
- Luồng nhật ký thời gian thực với tính năng tự động cuộn.
- **Lọc theo ứng dụng cụ thể**: Lọc nhật ký theo tên gói cụ thể.
- **Tự động giám sát**: Bắt đầu ghi nhật ký trước khi ứng dụng mở; công cụ sẽ tự động phát hiện PID và bắt đầu lọc sau khi ứng dụng khởi chạy.
- Tìm kiếm/lọc theo từ khóa.

### 💻 ADB Shell
- Terminal tích hợp để chạy các lệnh ADB.
- Thực thi lệnh nhanh chóng với lịch sử đầu ra.

---

## 🛠️ Binary tích hợp sẵn

Ứng dụng này hoàn toàn độc lập. Nó bao gồm:
- **ADB** (Android Debug Bridge)
- Tệp thực thi **Scrcpy**
- **Scrcpy-server**

Khi khởi động, các tệp này được giải nén vào một thư mục tạm thời và được sử dụng tự động. Bạn không cần cấu hình PATH hệ thống.

---

## ⚠️ Lưu ý quan trọng cho người dùng Xiaomi/Poco/Redmi

Để bật **điều khiển bằng cảm ứng** trong Scrcpy, bạn phải:
1. Vào **Tùy chọn nhà phát triển**.
2. Bật **Gỡ lỗi USB**.
3. Bật **Gỡ lỗi USB (Cài đặt bảo mật)**.
   *(Lưu ý: Điều này yêu cầu thẻ SIM và đăng nhập tài khoản Mi trên hầu hết các thiết bị Xiaomi).*

---

## 🚀 Bắt đầu

### Điều kiện tiên quyết
- **Go** (v1.21)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### Phát triển
```bash
wails dev
```

### Biên dịch (Build)
```bash
wails build
```
Ứng dụng đã biên dịch sẽ có sẵn trong `build/bin`.

### Phát hành (Release)
Dự án này sử dụng GitHub Actions để tự động hóa việc biên dịch đa nền tảng. Để tạo một bản phát hành mới:
1. Gắn thẻ (tag) cho commit của bạn: `git tag v1.0.0`
2. Đẩy thẻ lên: `git push origin v1.0.0`
GitHub Action sẽ tự động biên dịch cho macOS, Windows và Linux, và tải các tệp lên trang Release.

---

## 🔧 Khắc phục sự cố

### macOS: "Ứng dụng bị hỏng và không thể mở được"
Nếu bạn tải xuống ứng dụng từ GitHub và gặp lỗi *"adbGUI.app bị hỏng và không thể mở được"*, điều này là do tính năng cách ly Gatekeeper của macOS.

Để khắc phục điều này, hãy chạy lệnh sau trong terminal của bạn:
```bash
sudo xattr -cr /path/to/adbGUI.app
```
*(Thay thế `/path/to/adbGUI.app` bằng đường dẫn thực tế đến ứng dụng đã tải xuống của bạn)*

> **Hoặc tự build:** Nếu bạn không muốn bỏ qua Gatekeeper, bạn có thể dễ dàng [build ứng dụng từ mã nguồn](#-bắt-đầu) ngay trên máy của mình. Chỉ mất vài phút thôi!

### Windows: "Windows đã bảo vệ PC của bạn"
Nếu bạn thấy cửa sổ SmartScreen màu xanh ngăn ứng dụng khởi động:
1. Nhấp vào **Thông tin thêm (More info)**.
2. Nhấp vào **Vẫn chạy (Run anyway)**.

---

## 📄 Giấy phép
Dự án này được cấp phép theo Giấy phép MIT.
