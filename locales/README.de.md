# ADB GUI 🚀

Ein leistungsstarkes, modernes und eigenständiges Android-Verwaltungstool, entwickelt mit **Wails**, **React** und **Ant Design**.

> ✨ **Hinweis**: Diese Anwendung ist das Ergebnis von reinem **vibecoding**.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)

## ✨ Funktionen

### 📱 Geräteverwaltung
- Echtzeit-Überwachung verbundener Geräte.
- Anzeige von Geräte-ID, Modell und Verbindungsstatus.
- Ein-Klick-Zugriff auf Apps, Shell, Logcat und Spiegelung.

### 📦 App-Verwaltung
- Auflistung aller installierten Pakete (System- & Benutzer-Apps).
- Filtern und Suchen von Apps nach Name oder Typ.
- **Aktionen**: Stoppen erzwingen, Daten löschen, Aktivieren/Deaktivieren und Deinstallieren.
- **Schnell-Logcat**: Direkt aus der App-Liste zu den Protokollen einer bestimmten App springen.

### 🖥️ Bildschirmspiegelung (Scrcpy)
- **Integriertes Scrcpy**: Keine externe Installation erforderlich.
- Detaillierte Kontrolle über:
  - Video-Bitrate & maximale FPS.
  - Auflösung (maximale Größe).
  - Optionen für "Wach bleiben" & "Bildschirm ausschalten".
  - Fenster immer im Vordergrund.
  - Audio-Streaming umschalten.

### 📜 Erweitertes Logcat
- Echtzeit-Protokoll-Streaming mit automatischem Scrollen.
- **App-spezifische Filterung**: Protokolle nach einem bestimmten Paketnamen filtern.
- **Automatische Überwachung**: Protokollierung starten, bevor eine App geöffnet wird; das Tool erkennt automatisch die PID und beginnt mit der Filterung, sobald die App startet.
- Stichwortsuche/-filterung.

### 💻 ADB Shell
- Integriertes Terminal zum Ausführen von ADB-Befehlen.
- Schnelle Befehlsausführung mit Ausgabeverlauf.

---

## 🛠️ Integrierte Binärdateien

Diese Anwendung ist vollständig eigenständig. Sie enthält:
- **ADB** (Android Debug Bridge)
- **Scrcpy** ausführbare Datei
- **Scrcpy-server**

Beim Start werden diese in ein temporäres Verzeichnis extrahiert und automatisch verwendet. Sie müssen Ihren System-PATH nicht konfigurieren.

---

## ⚠️ Wichtige Hinweise für Xiaomi/Poco/Redmi-Benutzer

Um die **Touch-Steuerung** in Scrcpy zu aktivieren, müssen Sie:
1. Zu den **Entwickleroptionen** gehen.
2. **USB-Debugging** aktivieren.
3. **USB-Debugging (Sicherheitseinstellungen)** aktivieren.
   *(Hinweis: Dies erfordert bei den meisten Xiaomi-Geräten eine SIM-Karte und eine Anmeldung im Mi-Konto).*

---

## 🚀 Erste Schritte

### Voraussetzungen
- **Go** (v1.21)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### Entwicklung
```bash
wails dev
```

### Build
```bash
wails build
```
Die kompilierte Anwendung wird in `build/bin` verfügbar sein.

### Release
Dieses Projekt verwendet GitHub Actions, um Multi-Plattform-Builds zu automatisieren. So erstellen Sie ein neues Release:
1. Taggen Sie Ihren Commit: `git tag v1.0.0`
2. Pushen Sie den Tag: `git push origin v1.0.0`
Die GitHub Action wird automatisch für macOS, Windows und Linux bauen und die Artefakte auf die Release-Seite hochladen.

---

## 🔧 Fehlerbehebung

### macOS: "App ist beschädigt und kann nicht geöffnet werden"
Wenn Sie die App von GitHub herunterladen und den Fehler *"adbGUI.app ist beschädigt und kann nicht geöffnet werden"* sehen, liegt dies an der macOS Gatekeeper Quarantäne.

Um dies zu beheben, führen Sie folgenden Befehl im Terminal aus:
```bash
sudo xattr -cr /path/to/adbGUI.app
```
*(Ersetzen Sie `/path/to/adbGUI.app` durch den tatsächlichen Pfad zu Ihrer heruntergeladenen Anwendung)*

> **Oder selbst kompilieren:** Wenn Sie Gatekeeper nicht umgehen möchten, können Sie die [App ganz einfach lokal aus dem Quellcode kompilieren](#-erste-schritte). Das dauert nur wenige Minuten!

### Windows: "Der Computer wurde durch Windows geschützt"
Wenn ein blaues SmartScreen-Fenster den Start verhindert:
1. Klicken Sie auf **Weitere Informationen (More info)**.
2. Klicken Sie auf **Trotzdem ausführen (Run anyway)**.

---

## 📄 Lizenz
Dieses Projekt ist unter der MIT-Lizenz lizenziert.
