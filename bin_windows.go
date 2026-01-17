//go:build windows

package main

import _ "embed"

//go:embed bin/windows/adb.exe
var adbBinary []byte

//go:embed bin/windows/scrcpy.exe
var scrcpyBinary []byte

//go:embed bin/windows/aapt.exe
var aaptBinary []byte

//go:embed bin/windows/ffmpeg.exe
var ffmpegBinary []byte

//go:embed bin/windows/ffprobe.exe
var ffprobeBinary []byte
