# ADB GUI 🚀

A powerful, modern, and self-contained Android management tool built with **Wails**, **React**, and **Ant Design**.

> ✨ **Note**: This application is a product of pure **vibecoding**.

[English](README.md) | [简体中文](locales/README.zh-CN.md) | [繁體中文](locales/README.zh-TW.md) | [日本語](locales/README.ja.md) | [한국어](locales/README.ko.md) | [Español](locales/README.es.md) | [Português](locales/README.pt.md) | [Français](locales/README.fr.md) | [Deutsch](locales/README.de.md) | [Русский](locales/README.ru.md) | [Tiếng Việt](locales/README.vi.md) | [العربية](locales/README.ar.md)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)

## 📸 App Screenshots

| Device Mirror | File Manager |
|:---:|:---:|
| <img src="screenshots/mirror.png" width="400" /> | <img src="screenshots/files.png" width="400" /> |

| App Management | Logcat Viewer |
|:---:|:---:|
| <img src="screenshots/apps.png" width="400" /> | <img src="screenshots/logcat.png" width="400" /> |

| ADB Shell | System Tray |
|:---:|:---:|
| <img src="screenshots/shell.png" width="400" /> | <img src="screenshots/tray.png" width="400" /> |


### 📱 Device Management
- **Unified Device List**: Seamlessly manage unified physical and wireless devices.
- **Wireless Connection**: Connect effortlessly via IP/Port pairing.
- **Device History**: Quick access to previously connected offline devices.
- **Detailed Info**: View real-time device stats, model, and ID.

### 📦 App Management
- **Full Package Control**: Install (Drag & Drop), Uninstall, Enable, Disable, Force Stop, Clear Data.
- **APK Management**: Export installed APKs, Batch Install.
- **Smart Filtering**: Search and filter by System/User apps.
- **Quick Actions**: Launch apps or jump directly to their logs.

### 🖥️ Screen Mirroring (Scrcpy)
- **High Performance**: Low-latency mirroring powered by Scrcpy.
- **Recording**: Independent background recording with one-click folder access.
- **Audio Forwarding**: Stream device audio to your computer (Android 11+).
- **Customization**: Adjust Resolution, Bitrate, FPS, and Codec (H.264/H.265).
- **Control**: Multi-touch support, Keep Awake, Screen Off mode.

### 📂 File Management
- **Full-Featured Explorer**: Browse, Copy, Cut, Paste, Rename, Delete, and Create Folders.
- **Drag & Drop**: Upload files by simply dragging them to the window.
- **Downloads**: Easy file transfer from device to computer.
- **Preview**: Open files directly on the host machine.

### 📜 Advanced Logcat
- **Real-time Streaming**: Live log viewer with auto-scroll control.
- **Powerful Filtering**: Filter by Log Level, Tag, PID, or custom Regex.
- **App-Centric**: Auto-filter logs for a specific application.

### 💻 ADB Shell
- **Integrated Console**: Run raw ADB commands directly within the app.
- **Command History**: Quick access to previously executed commands.

### 🔌 System Tray
- **Quick Access**: Control mirroring and view device status from the menu bar/system tray.
- **Background Operation**: Keep the app running in the background for instant access.

---

## 🛠️ Built-in Binaries

This application is fully self-contained. It bundles:
- **ADB** (Android Debug Bridge)
- **Scrcpy** executable
- **Scrcpy-server**

At startup, these are extracted to a temporary directory and used automatically. You don't need to configure your system PATH.

---

## ⚠️ Important Notes for Xiaomi/Poco/Redmi Users

To enable **touch control** in Scrcpy, you must:
1. Go to **Developer Options**.
2. Enable **USB Debugging**.
3. Enable **USB Debugging (Security settings)**. 
   *(Note: This requires a SIM card and Mi Account login on most Xiaomi devices).*

---

## 🚀 Getting Started

### Prerequisites
- **Go** (v1.21)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### Development
```bash
wails dev
```

### Build
```bash
wails build
```
The compiled application will be available in `build/bin`.

### Release
This project uses GitHub Actions to automate multi-platform builds. To create a new release:
1. Tag your commit: `git tag v1.0.0`
2. Push the tag: `git push origin v1.0.0`
The GitHub Action will automatically build for macOS, Windows, and Linux, and upload the artifacts to the Release page.

---

## 🔧 Troubleshooting

### macOS: "App is damaged and can't be opened"
If you download the app from GitHub and see the error *"adbGUI.app is damaged and can't be opened"*, this is due to macOS Gatekeeper quarantine.

To fix this, run the following command in your terminal:
```bash
sudo xattr -cr /path/to/adbGUI.app
```
*(Replace `/path/to/adbGUI.app` with the actual path to your downloaded application)*

> **Or build it yourself:** If you prefer not to bypass Gatekeeper, you can easily [build the app from source](#-getting-started) locally. It only takes a few minutes!

### Windows: "Windows protected your PC"
If you see a blue SmartScreen popup preventing the app from starting:
1. Click **More info**.
2. Click **Run anyway**.

---

## 📄 License
This project is licensed under the MIT License.
