# ADB GUI 🚀

一个强大、现代且自包含的 Android 管理工具，基于 **Wails**、**React** 和 **Ant Design** 构建。

> ✨ **注意**：本应用为纯 **vibecoding** 产物。

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)

## ✨ 功能特性

### 📱 设备管理
- 实时监控已连接设备。
- 查看设备 ID、型号和连接状态。
- 一键访问应用管理、Shell、Logcat 和屏幕镜像。

### 📦 应用管理
- 列出所有已安装的包（系统应用和用户应用）。
- 按名称或类型过滤和搜索应用。
- **操作**：强制停止、清除数据、启用/禁用和卸载。
- **快速 Logcat**：直接从应用列表跳转到特定应用的日志。

### 🖥️ 屏幕镜像 (Scrcpy)
- **内置 Scrcpy**：无需额外安装任何软件。
- 精细控制：
  - 视频比特率和最大 FPS。
  - 分辨率（最大尺寸）。
  - 保持唤醒和关闭屏幕选项。
  - 窗口置顶。
  - 音频流开关。

### 📜 高级 Logcat
- 带有自动滚动的实时日志流。
- **特定应用过滤**：按特定包名过滤日志。
- **自动监控**：在应用打开前开始记录；工具会自动检测 PID 并在此应用启动后开始过滤。
- 关键词搜索/过滤。

### 💻 ADB Shell
- 集成终端，用于运行 ADB 命令。
- 快速执行命令并带有输出历史记录。

---

## 🛠️ 内置二进制文件

本应用完全自包含。它捆绑了：
- **ADB** (Android Debug Bridge)
- **Scrcpy** 可执行文件
- **Scrcpy-server**

启动时，这些文件会被提取到临时目录并自动使用。您无需配置系统环境变量。

---

## ⚠️ 小米/Poco/红米用户重要提示

要在 Scrcpy 中启用**触摸控制**，您必须：
1. 进入 **开发者选项**。
2. 启用 **USB 调试**。
3. 启用 **USB 调试（安全设置）**。
   *(注意：在大多数小米设备上，这需要插入 SIM 卡并登录小米账号)*。

---

## 🚀 开始使用

### 前置条件
- **Go** (v1.21)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
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

### 发布
本项目使用 GitHub Actions 自动进行多平台构建。要创建新版本：
1. 为您的提交打标签：`git tag v1.0.0`
2. 推送标签：`git push origin v1.0.0`
GitHub Action 将自动为 macOS、Windows 和 Linux 构建，并将产物上传到 Release 页面。

---

## 🔧 故障排除

### macOS: "应用已损坏，无法打开"
如果您从 GitHub 下载应用并看到 *"adbGUI.app 已损坏，无法打开"* 的错误提示，这是由于 macOS Gatekeeper 的隔离机制导致的。

要解决此问题，请在终端中运行以下命令：
```bash
sudo xattr -cr /path/to/adbGUI.app
```
*(请将 `/path/to/adbGUI.app` 替换为您下载应用的实际路径)*

> **或者选择自己构建：** 如果您不想绕过 Gatekeeper，您可以轻松地[从源码构建应用](#-开始使用)。只需几分钟即可完成！

### Windows: "Windows 已保护你的电脑"
如果看到蓝色的 SmartScreen 窗口阻止应用启动：
1. 点击 **更多信息 (More info)**。
2. 点击 **仍要运行 (Run anyway)**。

---

## 📄 许可证
本项目采用 MIT 许可证。
