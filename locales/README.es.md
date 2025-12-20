# ADB GUI 🚀

Una herramienta de gestión de Android potente, moderna y autónoma construida con **Wails**, **React** y **Ant Design**.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)

## ✨ Características

### 📱 Gestión de Dispositivos
- Monitoreo en tiempo real de dispositivos conectados.
- Ver ID del dispositivo, modelo y estado de conexión.
- Acceso con un solo clic a Apps, Shell, Logcat y Mirroring.

### 📦 Gestión de Aplicaciones
- Listar todos los paquetes instalados (aplicaciones del sistema y de usuario).
- Filtrar y buscar aplicaciones por nombre o tipo.
- **Acciones**: Forzar detención, borrar datos, habilitar/deshabilitar y desinstalar.
- **Logcat rápido**: Saltar a los registros de una aplicación específica directamente desde la lista de aplicaciones.

### 🖥️ Espejo de Pantalla (Scrcpy)
- **Scrcpy integrado**: No es necesario instalar nada externamente.
- Control detallado sobre:
  - Tasa de bits de video y FPS máximos.
  - Resolución (tamaño máximo).
  - Opciones de mantener despierto y apagar pantalla.
  - Ventana siempre al frente.
  - Alternar transmisión de audio.

### 📜 Logcat Avanzado
- Transmisión de registros en tiempo real con desplazamiento automático.
- **Filtrado específico de la aplicación**: Filtrar registros por un nombre de paquete específico.
- **Monitoreo automático**: Iniciar el registro antes de que se abra una aplicación; la herramienta detectará automáticamente el PID y comenzará a filtrar una vez que la aplicación se inicie.
- Búsqueda/filtrado por palabras clave.

### 💻 ADB Shell
- Terminal integrada para ejecutar comandos ADB.
- Ejecución rápida de comandos con historial de salida.

---

## 🛠️ Binarios Integrados

Esta aplicación es totalmente autónoma. Incluye:
- **ADB** (Android Debug Bridge)
- Ejecutable **Scrcpy**
- **Scrcpy-server**

Al inicio, estos se extraen en un directorio temporal y se usan automáticamente. No es necesario configurar el PATH de su sistema.

---

## ⚠️ Notas Importantes para Usuarios de Xiaomi/Poco/Redmi

Para habilitar el **control táctil** en Scrcpy, debe:
1. Ir a **Opciones de Desarrollador**.
2. Habilitar **Depuración USB**.
3. Habilitar **Depuración USB (ajustes de seguridad)**.
   *(Nota: Esto requiere una tarjeta SIM e iniciar sesión en la cuenta Mi en la mayoría de los dispositivos Xiaomi).*

---

## 🚀 Empezando

### Prerrequisitos
- **Go** (v1.21+)
- **Node.js** (v18+)
- **Wails CLI** (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Desarrollo
```bash
wails dev
```

### Construcción
```bash
wails build
```
La aplicación compilada estará disponible en `build/bin`.

### Lanzamiento
Este proyecto utiliza GitHub Actions para automatizar las construcciones multiplataforma. Para crear un nuevo lanzamiento:
1. Etiquete su commit: `git tag v1.0.0`
2. Empuje la etiqueta: `git push origin v1.0.0`
La GitHub Action construirá automáticamente para macOS, Windows y Linux, y subirá los artefactos a la página de Lanzamientos.

---

## 📄 Licencia
Este proyecto está bajo la Licencia MIT.
