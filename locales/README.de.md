# Gaze

Ein leistungsstarkes, modernes und eigenstaendiges Android-Geraeteverwaltungs- und Automatisierungstool, entwickelt mit **Wails**, **React** und **Ant Design**. Es bietet eine einheitliche **Session-Event**-Architektur fuer vollstaendiges Tracking des Geraeteverhaltens, eine visuelle **Workflow**-Engine fuer Testautomatisierung und eine vollstaendige **MCP**-Integration (Model Context Protocol) fuer KI-gesteuerte Geraetesteuerung.


[English](README.md) | [简体中文](locales/README.zh-CN.md) | [繁體中文](locales/README.zh-TW.md) | [日本語](locales/README.ja.md) | [한국어](locales/README.ko.md) | [Español](locales/README.es.md) | [Português](locales/README.pt.md) | [Français](locales/README.fr.md) | [Deutsch](locales/README.de.md) | [Русский](locales/README.ru.md) | [Tiếng Việt](locales/README.vi.md) | [العربية](locales/README.ar.md)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)
[![Website](https://img.shields.io/website?up_message=online&url=https%3A%2F%2Fgaze.nicetooo.com)](https://gaze.nicetooo.com)

## Warum Gaze?

- **Modern und schnell**: Entwickelt mit Wails (Go + React) fuer ein natives Erlebnis mit minimalem Ressourcenverbrauch.
- **Vollstaendig eigenstaendig**: Keine Installation von `adb`, `scrcpy`, `aapt`, `ffmpeg` oder `ffprobe` auf Ihrem System erforderlich. Alles ist integriert und sofort einsatzbereit.
- **Zuverlaessige Dateituebertragung**: Eine robuste Alternative zum oft unzuverlaessigen *Android File Transfer* unter macOS.
- **Multi-Geraete-Leistung**: Unterstuetzt unabhaengige, gleichzeitige Hintergrundaufnahmen fuer mehrere Geraete.
- **Session-Event-Architektur**: Einheitliches Tracking aller Geraeteaktivitaeten (Protokolle, Netzwerk, Touch, App-Lebenszyklus) auf einer einzelnen Zeitleiste.
- **Visuelle Workflow-Automatisierung**: Erstellen Sie komplexe Testablaeufe mit einem Drag-and-Drop-Node-Editor -- kein Code erforderlich.
- **KI-bereit ueber MCP**: Ueber 50 Tools, die ueber das Model Context Protocol bereitgestellt werden, fuer nahtlose Integration mit KI-Clients wie Claude Desktop und Cursor.
- **Entwicklerorientiert**: Integriertes Logcat, Shell, MITM-Proxy und UI-Inspector -- von Entwicklern fuer Entwickler.

## App-Screenshots

| Geräteverwaltung | Bildschirmspiegelung |
|:---:|:---:|
| <img src="screenshots/devices.png" width="400" /> | <img src="screenshots/mirror.png" width="400" /> |

| Dateimanager | App-Verwaltung |
|:---:|:---:|
| <img src="screenshots/files.png" width="400" /> | <img src="screenshots/apps.png" width="400" /> |

| Leistungsmonitor | Sitzungs-Timeline |
|:---:|:---:|
| <img src="screenshots/performance.png" width="400" /> | <img src="screenshots/session.png" width="400" /> |

| Sitzungsliste | Logcat-Ansicht |
|:---:|:---:|
| <img src="screenshots/session-list.png" width="400" /> | <img src="screenshots/logcat.png" width="400" /> |

| Visueller Workflow-Editor | Workflow-Liste |
|:---:|:---:|
| <img src="screenshots/workflow-editor.png" width="400" /> | <img src="screenshots/workflow.png" width="400" /> |

| UI-Inspektor | Touch-Aufnahme |
|:---:|:---:|
| <img src="screenshots/ui-inspector.png" width="400" /> | <img src="screenshots/recording.png" width="400" /> |

| Netzwerk-Proxy (MITM) | ADB Shell |
|:---:|:---:|
| <img src="screenshots/proxy.png" width="400" /> | <img src="screenshots/shell.png" width="400" /> |

---

## Funktionen

### Geraeteverwaltung
- **Einheitliche Geraetelist**: Nahtlose Verwaltung physischer und drahtloser Geraete mit automatischer USB/Wi-Fi-Zusammenfuehrung.
- **Drahtlose Verbindung**: Muehelos per IP/Port-Pairing mit mDNS-Unterstuetzung verbinden.
- **Geraetehistorie**: Schneller Zugriff auf zuvor verbundene Offline-Geraete.
- **Geraete-Pinning**: Heften Sie Ihr meistgenutztes Geraet an, damit es immer ganz oben in der Liste steht.
- **Geraeteueberwachung**: Echtzeit-Tracking von Akku-, Netzwerk- und Bildschirmstatusaenderungen.
- **Batch-Operationen**: Operationen gleichzeitig auf mehreren Geraeten ausfuehren.

### App-Verwaltung
- **Volle Paketkontrolle**: Installieren (Drag & Drop), Deinstallieren, Aktivieren, Deaktivieren, Stoppen erzwingen, Daten loeschen.
- **APK-Verwaltung**: Exportieren installierter APKs, Batch-Installation.
- **Intelligente Filterung**: Suchen und Filtern nach System-/Benutzer-Apps.
- **Schnellaktionen**: Apps starten oder direkt zu deren Protokollen springen.

### Bildschirmspiegelung (Scrcpy)
- **Hohe Leistung**: Spiegelung mit geringer Latenz, angetrieben von Scrcpy.
- **Aufnahme**: Unabhaengige Hintergrundaufnahme mit Unterstuetzung fuer mehrere Geraete gleichzeitig und Ein-Klick-Ordnerzugriff.
- **Audio-Weiterleitung**: Geraeteaudio auf Ihren Computer streamen (Android 11+).
- **Anpassung**: Aufloesung, Bitrate, FPS und Codec (H.264/H.265) anpassen.
- **Steuerung**: Multi-Touch-Unterstuetzung, Wach-bleiben-Modus, Bildschirm-Aus-Modus.

### Dateiverwaltung
- **Voll ausgestatteter Explorer**: Durchsuchen, Kopieren, Ausschneiden, Einfuegen, Umbenennen, Loeschen und Ordner erstellen.
- **Drag & Drop**: Dateien hochladen, indem Sie sie einfach in das Fenster ziehen.
- **Downloads**: Einfache Dateituebertragung vom Geraet auf den Computer.
- **Vorschau**: Dateien direkt auf dem Host-Computer oeffnen.

### Erweitertes Logcat
- **Echtzeit-Streaming**: Live-Protokollansicht mit automatischer Scroll-Steuerung.
- **Leistungsstarke Filterung**: Filtern nach Protokollebene, Tag, PID oder benutzerdefiniertem Regex.
- **App-zentriert**: Automatisches Filtern von Protokollen fuer eine bestimmte Anwendung.
- **JSON-Formatierung**: Erkannte JSON-Protokollsegmente uebersichtlich formatieren.

### Netzwerk und Proxy (MITM)
- **Automatisierte Erfassung**: Ein Klick zum Starten eines HTTP/HTTPS-Proxyservers und automatische Konfiguration der Geraete-Proxy-Einstellungen ueber ADB.
- **HTTPS-Entschluesselung (MITM)**: Unterstuetzung fuer die Entschluesselung von SSL-Datenverkehr mit automatischer CA-Zertifikatsgenerierung und -bereitstellung.
- **WebSocket-Unterstuetzung**: Erfassung und Inspektion von Echtzeit-WebSocket-Datenverkehr.
- **Verarbeitung grosser Datenmengen**: Unterstuetzung fuer vollstaendige Body-Erfassung (bis zu 100 MB) ohne Kuerzung, mit einem Protokollpuffer fuer 5000 Eintraege.
- **Traffic Shaping**: Simulieren Sie reale Netzwerkbedingungen mit geraetespezifischen Download-/Upload-Bandbreitenlimits und kuenstlicher Latenz.
- **Visuelle Metriken**: Echtzeit-RX/TX-Geschwindigkeitsueberwachung fuer das ausgewaehlte Geraet.

### Session und Event Tracking
- **Einheitliche Event Pipeline**: Alle Geraeteaktivitaeten (Protokolle, Netzwerkanfragen, Touch-Events, App-Lebenszyklus, Assertions) werden als Events erfasst und mit einer Session-Zeitleiste verknuepft.
- **Automatisches Session-Management**: Sessions werden automatisch erstellt, wenn Events auftreten, oder manuell mit benutzerdefinierten Konfigurationen (Logcat, Aufnahme, Proxy, Monitoring).
- **Event-Zeitleiste**: Mehrspurige Visualisierung aller Events mit zeitbasierter Indizierung und Navigation.
- **Volltextsuche**: Suche ueber alle Events mit SQLite FTS5.
- **Backpressure-Kontrolle**: Automatisches Event-Sampling bei hoher Last unter Schutz kritischer Events (Fehler, Netzwerk, Workflow).
- **Event Assertions**: Definition und Auswertung von Assertions gegen Event-Stroeme fuer automatisierte Validierung.
- **Video-Synchronisierung**: Extraktion von Videoframes synchronisiert mit Event-Zeitstempeln fuer visuelles Debugging.

### UI Inspector und Automatisierung
- **UI-Hierarchie-Inspektor**: Durchsuchen und Analysieren des vollstaendigen UI-Baums jedes Bildschirms.
- **Element-Picker**: Klicken, um UI-Elemente auszuwaehlen und deren Eigenschaften zu inspizieren (Resource-ID, Text, Bounds, Klasse).
- **Touch-Aufnahme**: Touch-Interaktionen aufnehmen und als Automatisierungsskripte wiedergeben.
- **Elementbasierte Aktionen**: Klicken, langes Klicken, Texteingabe, Wischen, Warten und Assertions auf UI-Elementen mittels Selektoren (ID, Text, contentDesc, className, XPath).

### Visuelle Workflow-Engine
- **Node-basierter Editor**: Automatisierungsablaeufe visuell erstellen mit einer Drag-and-Drop-Oberflaeche, angetrieben von XYFlow.
- **Ueber 30 Schritttypen**: Tippen, Wischen, Element-Interaktion, App-Steuerung, Tastenereignisse, Bildschirmsteuerung, Warten, ADB-Befehle, Variablen, Verzweigung, Sub-Workflows und Session-Steuerung.
- **Bedingte Verzweigung**: Intelligente Ablaeufe mit exists/not_exists/text_equals/text_contains-Bedingungen erstellen.
- **Variablen und Ausdruecke**: Workflow-Variablen mit arithmetischer Ausdrucksunterstuetzung verwenden (`{{count}} + 1`).
- **Schritt-fuer-Schritt-Debugging**: Anhalten, schrittweise durchgehen und Variablenzustand bei jedem Workflow-Schritt inspizieren.
- **Session-Integration**: Tracking-Sessions innerhalb von Workflows starten/stoppen fuer umfassende Testberichte.

### ADB Shell
- **Integrierte Konsole**: Rohe ADB-Befehle direkt in der App ausfuehren.
- **Befehlsverlauf**: Schneller Zugriff auf zuvor ausgefuehrte Befehle.

### Systemablage
- **Schnellzugriff**: Spiegelung steuern und Geraetestatus ueber die Menuleiste/Systemablage anzeigen.
- **Geraete-Pinning**: Hauptgeraet anheften, damit es oben in der Liste und im Systemablage-Menue erscheint.
- **Ablage-Funktionen**: Direkter Zugriff auf Logcat, Shell und Dateimanager fuer angeheftete Geraete ueber die Systemablage.
- **Aufnahme-Indikatoren**: Visueller roter Punkt in der Systemablage, wenn eine Aufnahme aktiv ist.
- **Hintergrundbetrieb**: App im Hintergrund laufen lassen fuer sofortigen Zugriff.

---

## MCP-Integration (Model Context Protocol)

Gaze enthaelt einen integrierten **MCP-Server**, der ueber 50 Tools und 5 Ressourcen bereitstellt. Dies ermoeglicht KI-Clients die vollstaendige Steuerung von Android-Geraeten durch natuerliche Sprache. Gaze wird so zur Bruecke zwischen KI und Android.

### Unterstuetzte KI-Clients

| Client | Transport | Konfiguration |
|--------|-----------|---------------|
| **Claude Desktop** | SSE | `claude_desktop_config.json` |
| **Claude Code (CLI)** | SSE | `.claude/settings.json` |
| **Cursor** | SSE | Cursor MCP-Einstellungen |

### Schnelleinrichtung

Der MCP-Server startet automatisch mit Gaze unter `http://localhost:23816/mcp/sse`.

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

**Cursor**: Fuegen Sie die MCP-Server-URL `http://localhost:23816/mcp/sse` in den MCP-Einstellungen von Cursor hinzu.

### MCP-Tools (50+)

| Kategorie | Tools | Beschreibung |
|-----------|-------|--------------|
| **Geraete** | `device_list`, `device_info`, `device_connect`, `device_disconnect`, `device_pair`, `device_wireless`, `device_ip` | Geraeteerkennung, Verbindung und Informationen |
| **CLI-Tools** | `adb_execute`, `aapt_execute`, `ffmpeg_execute`, `ffprobe_execute` | Ausfuehrung integrierter CLI-Tools (ADB, AAPT, FFmpeg, FFprobe) |
| **Apps** | `app_list`, `app_info`, `app_start`, `app_stop`, `app_running`, `app_install`, `app_uninstall`, `app_clear_data` | Vollstaendiges Anwendungs-Lifecycle-Management |
| **Bildschirm** | `screen_screenshot`, `screen_record_start`, `screen_record_stop`, `screen_recording_status` | Screenshots (Base64) und Aufnahmesteuerung |
| **UI-Automatisierung** | `ui_hierarchy`, `ui_search`, `ui_tap`, `ui_swipe`, `ui_input`, `ui_resolution` | UI-Inspektion, Element-Interaktion und Eingabe |
| **Sessions** | `session_create`, `session_end`, `session_active`, `session_list`, `session_events`, `session_stats` | Session-Lebenszyklus und Event-Abfragen |
| **Workflows** | `workflow_list`, `workflow_get`, `workflow_create`, `workflow_update`, `workflow_delete`, `workflow_run`, `workflow_stop`, `workflow_pause`, `workflow_resume`, `workflow_step_next`, `workflow_status`, `workflow_execute_step` | Vollstaendiges Workflow-CRUD, Ausfuehrung und Debugging |
| **Proxy** | `proxy_start`, `proxy_stop`, `proxy_status` | Netzwerk-Proxy-Steuerung |
| **Video** | `video_frame`, `video_metadata`, `session_video_frame`, `session_video_info` | Videoframe-Extraktion und Metadaten |

### MCP-Ressourcen

| URI | Beschreibung |
|-----|--------------|
| `gaze://devices` | Liste verbundener Geraete |
| `gaze://devices/{deviceId}` | Detaillierte Geraeteinformationen |
| `gaze://sessions` | Aktive und kuerzliche Sessions |
| `workflow://list` | Alle gespeicherten Workflows |
| `workflow://{workflowId}` | Workflow-Details mit Schritten |

### Was kann KI mit Gaze tun?

Mit der MCP-Integration koennen KI-Clients:
- **Tests automatisieren**: UI-Test-Workflows durch natuerlichsprachliche Anweisungen erstellen und ausfuehren.
- **Probleme debuggen**: Screenshots aufnehmen, UI-Hierarchie inspizieren, Protokolle lesen und Netzwerkverkehr analysieren.
- **Geraete verwalten**: Apps installieren, Dateien uebertragen, Einstellungen auf mehreren Geraeten konfigurieren.
- **Workflows erstellen**: Komplexe Automatisierungs-Workflows mit Verzweigungslogik und Variablenverwaltung generieren.
- **Sessions ueberwachen**: Geraeteverhalten ueber die Zeit mit eventbasierter Session-Aufnahme verfolgen.

---

## Integrierte Binaerdateien

Diese Anwendung ist vollstaendig eigenstaendig. Sie enthaelt:
- **ADB** (Android Debug Bridge)
- **Scrcpy** (Bildschirmspiegelung und Aufnahme)
- **AAPT** (Android Asset Packaging Tool)
- **FFmpeg** (Video-/Audioverarbeitung)
- **FFprobe** (Medienanalyse)

Beim Start werden diese in ein temporaeres Verzeichnis extrahiert und automatisch verwendet. Sie muessen Ihren System-PATH nicht konfigurieren.

---

## Wichtige Hinweise fuer Xiaomi/Poco/Redmi-Benutzer

Um die **Touch-Steuerung** in Scrcpy zu aktivieren, muessen Sie:
1. Zu den **Entwickleroptionen** gehen.
2. **USB-Debugging** aktivieren.
3. **USB-Debugging (Sicherheitseinstellungen)** aktivieren.
   *(Hinweis: Dies erfordert bei den meisten Xiaomi-Geraeten eine SIM-Karte und eine Anmeldung mit dem Mi-Konto).*

---

## Erste Schritte

### Voraussetzungen
- **Go** (v1.23+)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### Entwicklung
```bash
wails dev -tags fts5
```

### Build
```bash
wails build -tags fts5
```
Die kompilierte Anwendung befindet sich in `build/bin`.

### Tests ausfuehren
```bash
go test -tags fts5 ./...
```

### Release
Dieses Projekt verwendet GitHub Actions zur Automatisierung von Multi-Plattform-Builds. So erstellen Sie ein neues Release:
1. Taggen Sie Ihren Commit: `git tag v1.0.0`
2. Pushen Sie den Tag: `git push origin v1.0.0`
Die GitHub Action erstellt automatisch Builds fuer macOS, Windows und Linux und laedt die Artefakte auf die Release-Seite hoch.

---

## Architekturuebersicht

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

## Technologie-Stack

| Schicht | Technologie |
|---------|-------------|
| **Desktop-Framework** | Wails v2 |
| **Backend** | Go 1.23+ |
| **Frontend** | React 18, TypeScript, Ant Design 6 |
| **State Management** | Zustand |
| **Workflow-Editor** | XYFlow + Dagre |
| **Datenbank** | SQLite (WAL-Modus, FTS5) |
| **Proxy** | goproxy |
| **MCP** | mcp-go (Model Context Protocol) |
| **i18n** | i18next (5 Sprachen) |
| **Logging** | zerolog |
| **Diagramme** | Recharts |

---

## Fehlerbehebung

### macOS: "App ist beschaedigt und kann nicht geoeffnet werden"
Wenn Sie die App von GitHub herunterladen und den Fehler *"Gaze.app ist beschaedigt und kann nicht geoeffnet werden"* sehen, liegt dies an der macOS-Gatekeeper-Quarantaene.

Fuehren Sie folgenden Befehl in Ihrem Terminal aus, um das Problem zu beheben:
```bash
sudo xattr -cr /path/to/Gaze.app
```
*(Ersetzen Sie `/path/to/Gaze.app` durch den tatsaechlichen Pfad zu Ihrer heruntergeladenen Anwendung)*

> **Oder selbst kompilieren:** Wenn Sie Gatekeeper nicht umgehen moechten, koennen Sie die [App ganz einfach lokal aus dem Quellcode kompilieren](#erste-schritte). Das dauert nur wenige Minuten!

### Windows: "Der Computer wurde durch Windows geschuetzt"
Wenn ein blaues SmartScreen-Fenster den Start der App verhindert:
1. Klicken Sie auf **Weitere Informationen**.
2. Klicken Sie auf **Trotzdem ausfuehren**.

---

## Lizenz
Dieses Projekt ist unter der MIT-Lizenz lizenziert.
