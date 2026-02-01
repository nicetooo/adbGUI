//go:build linux

package main

import _ "embed"

//go:embed bin/linux/adb
var adbBinary []byte

//go:embed bin/linux/scrcpy
var scrcpyBinary []byte

//go:embed bin/linux/aapt
var aaptBinary []byte

//go:embed bin/linux/ffmpeg
var ffmpegBinary []byte

//go:embed bin/linux/ffprobe
var ffprobeBinary []byte

//go:embed bin/linux/protoc
var protocBinary []byte
