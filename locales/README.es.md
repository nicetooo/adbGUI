# Gaze

Una herramienta potente, moderna y autocontenida para la gestion y automatizacion de dispositivos Android, construida con **Wails**, **React** y **Ant Design**. Cuenta con una arquitectura unificada **Session-Event** para el seguimiento completo del comportamiento del dispositivo, un motor visual de **Workflow** para la automatizacion de pruebas, e integracion completa con **MCP** (Model Context Protocol) para el control de dispositivos mediante IA.

> **Nota**: Esta aplicacion es fruto del puro **vibecoding**.

[English](README.md) | [简体中文](locales/README.zh-CN.md) | [繁體中文](locales/README.zh-TW.md) | [日本語](locales/README.ja.md) | [한국어](locales/README.ko.md) | [Español](locales/README.es.md) | [Português](locales/README.pt.md) | [Français](locales/README.fr.md) | [Deutsch](locales/README.de.md) | [Русский](locales/README.ru.md) | [Tiếng Việt](locales/README.vi.md) | [العربية](locales/README.ar.md)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)
[![Website](https://img.shields.io/website?up_message=online&url=https%3A%2F%2Fgaze.nicetooo.com)](https://gaze.nicetooo.com)

## Por que Gaze?

- **Moderno y Rapido**: Construido con Wails (Go + React), ofrece una experiencia nativa con un consumo minimo de recursos.
- **Verdaderamente Autocontenido**: No es necesario instalar `adb`, `scrcpy`, `aapt`, `ffmpeg` ni `ffprobe` en tu sistema. Todo esta incluido y listo para usar.
- **Transferencia de Archivos Fiable**: Una alternativa robusta a la frecuentemente inestable *Android File Transfer* en macOS.
- **Potencia Multi-Dispositivo**: Soporta grabacion independiente y simultanea en segundo plano para multiples dispositivos.
- **Arquitectura Session-Event**: Seguimiento unificado de todas las actividades del dispositivo (registros, red, tactil, ciclo de vida de apps) en una sola linea de tiempo.
- **Automatizacion Visual con Workflow**: Construye flujos de prueba complejos con un editor de nodos de arrastrar y soltar, sin necesidad de codigo.
- **Preparado para IA via MCP**: Mas de 50 herramientas expuestas a traves del Model Context Protocol para una integracion fluida con clientes de IA como Claude Desktop y Cursor.
- **Pensado para Desarrolladores**: Logcat integrado, Shell, Proxy MITM e Inspector de UI disenados por desarrolladores, para desarrolladores.

## Capturas de Pantalla

| Gestión de Dispositivos | Espejo de Pantalla |
|:---:|:---:|
| <img src="screenshots/devices.png" width="400" /> | <img src="screenshots/mirror.png" width="400" /> |

| Gestor de Archivos | Gestión de Aplicaciones |
|:---:|:---:|
| <img src="screenshots/files.png" width="400" /> | <img src="screenshots/apps.png" width="400" /> |

| Monitor de Rendimiento | Línea de Tiempo de Sesión |
|:---:|:---:|
| <img src="screenshots/performance.png" width="400" /> | <img src="screenshots/session.png" width="400" /> |

| Lista de Sesiones | Visor de Logcat |
|:---:|:---:|
| <img src="screenshots/session-list.png" width="400" /> | <img src="screenshots/logcat.png" width="400" /> |

| Editor Visual de Flujos de Trabajo | Lista de Flujos de Trabajo |
|:---:|:---:|
| <img src="screenshots/workflow-editor.png" width="400" /> | <img src="screenshots/workflow.png" width="400" /> |

| Inspector de UI | Grabación Táctil |
|:---:|:---:|
| <img src="screenshots/ui-inspector.png" width="400" /> | <img src="screenshots/recording.png" width="400" /> |

| Proxy de Red (MITM) | ADB Shell |
|:---:|:---:|
| <img src="screenshots/proxy.png" width="400" /> | <img src="screenshots/shell.png" width="400" /> |

---

## Caracteristicas

### Gestion de Dispositivos
- **Lista Unificada de Dispositivos**: Gestiona dispositivos fisicos e inalambricos de forma transparente con fusion automatica USB/Wi-Fi.
- **Conexion Inalambrica**: Conecta facilmente mediante emparejamiento IP/Puerto con soporte mDNS.
- **Historial de Dispositivos**: Acceso rapido a dispositivos desconectados previamente conectados.
- **Fijado de Dispositivos**: Fija tu dispositivo mas usado para que siempre aparezca en la parte superior de la lista.
- **Monitorizacion de Dispositivos**: Seguimiento en tiempo real de cambios de bateria, red y estado de pantalla.
- **Operaciones por Lotes**: Ejecuta operaciones en multiples dispositivos simultaneamente.

### Gestion de Aplicaciones
- **Control Total de Paquetes**: Instalar (Arrastrar y Soltar), Desinstalar, Habilitar, Deshabilitar, Forzar Detencion, Borrar Datos.
- **Gestion de APK**: Exportar APKs instalados, Instalacion por Lotes.
- **Filtrado Inteligente**: Buscar y filtrar por aplicaciones del Sistema/Usuario.
- **Acciones Rapidas**: Lanza aplicaciones o salta directamente a sus registros.

### Espejo de Pantalla (Scrcpy)
- **Alto Rendimiento**: Duplicacion de baja latencia impulsada por Scrcpy.
- **Grabacion**: Grabacion independiente en segundo plano con soporte para multiples dispositivos simultaneamente y acceso a la carpeta con un solo clic.
- **Reenvio de Audio**: Transmite el audio del dispositivo a tu computadora (Android 11+).
- **Personalizacion**: Ajusta Resolucion, Tasa de bits, FPS y Codec (H.264/H.265).
- **Control**: Soporte multitactil, Mantener despierto, Modo pantalla apagada.

### Gestion de Archivos
- **Explorador Completo**: Navegar, Copiar, Cortar, Pegar, Renombrar, Eliminar y Crear Carpetas.
- **Arrastrar y Soltar**: Sube archivos simplemente arrastandolos a la ventana.
- **Descargas**: Transferencia facil de archivos del dispositivo a la computadora.
- **Vista Previa**: Abre archivos directamente en la maquina host.

### Logcat Avanzado
- **Transmision en Tiempo Real**: Visor de registros en vivo con control de desplazamiento automatico.
- **Filtrado Potente**: Filtra por Nivel de Registro, Etiqueta, PID o Regex personalizado.
- **Centrado en la Aplicacion**: Filtra automaticamente los registros de una aplicacion especifica.
- **Formato JSON**: Visualizacion con formato de segmentos de registro JSON detectados.

### Red y Proxy (MITM)
- **Captura Automatizada**: Un clic para iniciar un servidor proxy HTTP/HTTPS y configurar automaticamente los ajustes de proxy del dispositivo via ADB.
- **Descifrado HTTPS (MITM)**: Soporte para descifrar trafico SSL con generacion e implementacion automatica de certificados CA.
- **Soporte WebSocket**: Captura e inspeccion de trafico WebSocket en tiempo real.
- **Manejo de Grandes Volumenes**: Soporte para captura completa del cuerpo (hasta 100MB) sin truncamiento, con un buffer de 5000 entradas de registro.
- **Modelado de Trafico**: Simula condiciones de red reales con limites de ancho de banda de Descarga/Subida por dispositivo y latencia artificial.
- **Metricas Visuales**: Monitorizacion en tiempo real de velocidad RX/TX para el dispositivo seleccionado.

### Session y Seguimiento de Eventos
- **Pipeline de Eventos Unificado**: Todas las actividades del dispositivo (registros, peticiones de red, eventos tactiles, ciclo de vida de apps, aserciones) se capturan como eventos y se vinculan a una linea de tiempo de session.
- **Gestion Automatica de Sessions**: Las sessions se crean automaticamente cuando ocurren eventos, o manualmente con configuraciones personalizadas (logcat, grabacion, proxy, monitorizacion).
- **Linea de Tiempo de Eventos**: Visualizacion multicarril de todos los eventos con indexacion y navegacion basada en tiempo.
- **Busqueda de Texto Completo**: Busca en todos los eventos usando SQLite FTS5.
- **Control de Contrapresion**: Muestreo automatico de eventos bajo alta carga, protegiendo eventos criticos (errores, red, workflow).
- **Aserciones de Eventos**: Define y evalua aserciones contra flujos de eventos para validacion automatizada.
- **Sincronizacion de Video**: Extrae fotogramas de video sincronizados con marcas de tiempo de eventos para depuracion visual.

### Inspector de UI y Automatizacion
- **Inspector de Jerarquia de UI**: Navega y analiza el arbol completo de UI de cualquier pantalla.
- **Selector de Elementos**: Haz clic para seleccionar elementos de UI e inspeccionar sus propiedades (resource-id, text, bounds, class).
- **Grabacion Tactil**: Graba interacciones tactiles y reproducelas como scripts de automatizacion.
- **Acciones Basadas en Elementos**: Clic, clic largo, introduccion de texto, deslizamiento, espera y asercion sobre elementos de UI usando selectores (id, text, contentDesc, className, xpath).

### Motor Visual de Workflow
- **Editor Basado en Nodos**: Construye flujos de automatizacion visualmente con una interfaz de arrastrar y soltar impulsada por XYFlow.
- **Mas de 30 Tipos de Pasos**: Toque, deslizamiento, interaccion con elementos, control de apps, eventos de teclado, control de pantalla, espera, comandos ADB, variables, ramificacion, sub-workflows y control de session.
- **Ramificacion Condicional**: Crea flujos inteligentes con condiciones exists/not_exists/text_equals/text_contains.
- **Variables y Expresiones**: Usa variables de workflow con soporte de expresiones aritmeticas (`{{count}} + 1`).
- **Depuracion Paso a Paso**: Pausa, avanza paso a paso e inspecciona el estado de las variables en cada paso del workflow.
- **Integracion con Sessions**: Inicia/detiene sessions de seguimiento dentro de workflows para reportes de pruebas completos.

### ADB Shell
- **Consola Integrada**: Ejecuta comandos ADB sin procesar directamente dentro de la aplicacion.
- **Historial de Comandos**: Acceso rapido a comandos ejecutados anteriormente.

### Bandeja del Sistema
- **Acceso Rapido**: Controla el espejo y visualiza el estado del dispositivo desde la barra de menu/bandeja del sistema.
- **Fijado de Dispositivos**: Fija tu dispositivo principal para que aparezca en la parte superior de la lista y del menu de la bandeja.
- **Funciones de Bandeja**: Acceso directo a Logcat, Shell y Gestor de Archivos para dispositivos fijados desde la bandeja.
- **Indicadores de Grabacion**: Indicador visual de punto rojo en la bandeja cuando la grabacion esta activa.
- **Operacion en Segundo Plano**: Mantiene la aplicacion ejecutandose en segundo plano para acceso instantaneo.

---

## Integracion MCP (Model Context Protocol)

Gaze incluye un **servidor MCP** integrado que expone mas de 50 herramientas y 5 recursos, permitiendo a los clientes de IA controlar completamente dispositivos Android a traves de lenguaje natural. Esto convierte a Gaze en el puente entre la IA y Android.

### Clientes de IA Soportados

| Cliente | Transporte | Configuracion |
|---------|------------|---------------|
| **Claude Desktop** | SSE | `claude_desktop_config.json` |
| **Claude Code (CLI)** | SSE | `.claude/settings.json` |
| **Cursor** | SSE | Configuracion MCP de Cursor |

### Configuracion Rapida

El servidor MCP se inicia automaticamente con Gaze en `http://localhost:23816/mcp/sse`.

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

**Cursor**: Agrega la URL del servidor MCP `http://localhost:23816/mcp/sse` en la configuracion MCP de Cursor.

### Herramientas MCP (50+)

| Categoria | Herramientas | Descripcion |
|-----------|-------------|-------------|
| **Device** | `device_list`, `device_info`, `device_connect`, `device_disconnect`, `device_pair`, `device_wireless`, `device_ip` | Descubrimiento, conexion e informacion de dispositivos |
| **CLI Tools** | `adb_execute`, `aapt_execute`, `ffmpeg_execute`, `ffprobe_execute` | Ejecucion de herramientas CLI incluidas (ADB, AAPT, FFmpeg, FFprobe) |
| **Apps** | `app_list`, `app_info`, `app_start`, `app_stop`, `app_running`, `app_install`, `app_uninstall`, `app_clear_data` | Gestion completa del ciclo de vida de aplicaciones |
| **Screen** | `screen_screenshot`, `screen_record_start`, `screen_record_stop`, `screen_recording_status` | Capturas de pantalla (base64) y control de grabacion |
| **UI Automation** | `ui_hierarchy`, `ui_search`, `ui_tap`, `ui_swipe`, `ui_input`, `ui_resolution` | Inspeccion de UI, interaccion con elementos y entrada |
| **Sessions** | `session_create`, `session_end`, `session_active`, `session_list`, `session_events`, `session_stats` | Ciclo de vida de sessions y consulta de eventos |
| **Workflows** | `workflow_list`, `workflow_get`, `workflow_create`, `workflow_update`, `workflow_delete`, `workflow_run`, `workflow_stop`, `workflow_pause`, `workflow_resume`, `workflow_step_next`, `workflow_status`, `workflow_execute_step` | CRUD completo de workflows, ejecucion y depuracion |
| **Proxy** | `proxy_start`, `proxy_stop`, `proxy_status` | Control del proxy de red |
| **Video** | `video_frame`, `video_metadata`, `session_video_frame`, `session_video_info` | Extraccion de fotogramas de video y metadatos |

### Recursos MCP

| URI | Descripcion |
|-----|-------------|
| `gaze://devices` | Lista de dispositivos conectados |
| `gaze://devices/{deviceId}` | Informacion detallada del dispositivo |
| `gaze://sessions` | Sessions activas y recientes |
| `workflow://list` | Todos los workflows guardados |
| `workflow://{workflowId}` | Detalles del workflow con sus pasos |

### Que Puede Hacer la IA con Gaze?

Con la integracion MCP, los clientes de IA pueden:
- **Automatizar Pruebas**: Crear y ejecutar workflows de pruebas de UI mediante instrucciones en lenguaje natural.
- **Depurar Problemas**: Tomar capturas de pantalla, inspeccionar la jerarquia de UI, leer registros y analizar el trafico de red.
- **Gestionar Dispositivos**: Instalar aplicaciones, transferir archivos, configurar ajustes en multiples dispositivos.
- **Construir Workflows**: Generar workflows de automatizacion complejos con logica de ramificacion y gestion de variables.
- **Monitorizar Sessions**: Hacer seguimiento del comportamiento del dispositivo a lo largo del tiempo con grabacion de sessions basada en eventos.

---

## Binarios Integrados

Esta aplicacion es completamente autocontenida. Incluye:
- **ADB** (Android Debug Bridge)
- **Scrcpy** (Espejo de pantalla y grabacion)
- **AAPT** (Android Asset Packaging Tool)
- **FFmpeg** (Procesamiento de video/audio)
- **FFprobe** (Analisis de medios)

Al inicio, estos se extraen a un directorio temporal y se usan automaticamente. No necesitas configurar el PATH de tu sistema.

---

## Notas Importantes para Usuarios de Xiaomi/Poco/Redmi

Para habilitar el **control tactil** en Scrcpy, debes:
1. Ir a **Opciones de Desarrollador**.
2. Habilitar **Depuracion USB**.
3. Habilitar **Depuracion USB (ajustes de seguridad)**.
   *(Nota: Esto requiere una tarjeta SIM e iniciar sesion en la cuenta Mi en la mayoria de los dispositivos Xiaomi).*

---

## Primeros Pasos

### Prerrequisitos
- **Go** (v1.23+)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### Desarrollo
```bash
wails dev
```

### Compilacion
```bash
wails build
```
La aplicacion compilada estara disponible en `build/bin`.

### Ejecucion de Pruebas
```bash
go test ./...
```

### Lanzamiento
Este proyecto utiliza GitHub Actions para automatizar las compilaciones multiplataforma. Para crear un nuevo lanzamiento:
1. Etiqueta tu commit: `git tag v1.0.0`
2. Empuja la etiqueta: `git push origin v1.0.0`
La GitHub Action compilara automaticamente para macOS, Windows y Linux, y subira los artefactos a la pagina de Lanzamientos.

---

## Vision General de la Arquitectura

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

## Stack Tecnologico

| Capa | Tecnologia |
|------|-----------|
| **Framework de Escritorio** | Wails v2 |
| **Backend** | Go 1.23+ |
| **Frontend** | React 18, TypeScript, Ant Design 6 |
| **Gestion de Estado** | Zustand |
| **Editor de Workflows** | XYFlow + Dagre |
| **Base de Datos** | SQLite (modo WAL, FTS5) |
| **Proxy** | goproxy |
| **MCP** | mcp-go (Model Context Protocol) |
| **i18n** | i18next (5 idiomas) |
| **Logging** | zerolog |
| **Graficos** | Recharts |

---

## Solucion de Problemas

### macOS: "La aplicacion esta danada y no se puede abrir"
Si descargas la aplicacion desde GitHub y ves el error *"Gaze.app esta danada y no se puede abrir"*, esto se debe a la cuarentena de macOS Gatekeeper.

Para solucionarlo, ejecuta el siguiente comando en tu terminal:
```bash
sudo xattr -cr /path/to/Gaze.app
```
*(Reemplaza `/path/to/Gaze.app` con la ruta real de tu aplicacion descargada)*

> **O compilalo tu mismo:** Si prefieres no eludir Gatekeeper, puedes [compilar la aplicacion desde el codigo fuente](#primeros-pasos) facilmente de forma local. Solo toma unos minutos.

### Windows: "Windows protegio su PC"
Si ves una ventana emergente azul de SmartScreen impidiendo que la aplicacion se inicie:
1. Haz clic en **Mas informacion**.
2. Haz clic en **Ejecutar de todas formas**.

---

## Licencia
Este proyecto esta bajo la Licencia MIT.
