package main

import _ "embed"

//go:embed bin/scrcpy-server
var scrcpyServerBinary []byte

//go:embed bin/ADBKeyboard.apk
var adbKeyboardAPK []byte
