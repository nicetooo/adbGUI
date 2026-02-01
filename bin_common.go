package main

import "embed"

//go:embed bin/scrcpy-server
var scrcpyServerBinary []byte

//go:embed bin/ADBKeyboard.apk
var adbKeyboardAPK []byte

// Protoc well-known type includes (platform-independent)
//
//go:embed bin/protoc-include
var protocIncludeFS embed.FS
