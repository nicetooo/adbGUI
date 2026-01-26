package main

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

const (
	adbKeyboardPackage = "com.android.adbkeyboard"
	adbKeyboardIME     = "com.android.adbkeyboard/.AdbIME"
)

// IsADBKeyboardInstalled checks if ADBKeyboard is installed on the device
func (a *App) IsADBKeyboardInstalled(deviceId string) bool {
	output, err := a.RunAdbCommand(deviceId, "shell pm list packages "+adbKeyboardPackage)
	return err == nil && strings.Contains(output, "package:"+adbKeyboardPackage)
}

// IsADBKeyboardActive checks if ADBKeyboard is the current active IME
func (a *App) IsADBKeyboardActive(deviceId string) bool {
	output, err := a.RunAdbCommand(deviceId, "shell settings get secure default_input_method")
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) == adbKeyboardIME
}

// EnsureADBKeyboard installs (if needed) and activates ADBKeyboard on the device.
// This is called lazily on first Unicode text input.
// Returns (ready, installOccurred, error).
func (a *App) EnsureADBKeyboard(deviceId string) (bool, bool, error) {
	if a.adbKeyboardPath == "" {
		return false, false, fmt.Errorf("ADBKeyboard APK not available (not embedded in this build)")
	}

	installed := false

	// Step 1: Install if not present
	if !a.IsADBKeyboardInstalled(deviceId) {
		LogDebug("adb_keyboard").Str("deviceId", deviceId).Msg("Installing ADBKeyboard APK...")
		// Use newAdbCommand with separate args to handle paths with spaces
		cmd := a.newAdbCommand(nil, "-s", deviceId, "install", "-r", a.adbKeyboardPath)
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		if err != nil {
			return false, false, fmt.Errorf("failed to install ADBKeyboard: %w (output: %s)", err, output)
		}
		if !strings.Contains(output, "Success") {
			return false, false, fmt.Errorf("ADBKeyboard install did not succeed: %s", output)
		}
		installed = true
		LogDebug("adb_keyboard").Str("deviceId", deviceId).Msg("ADBKeyboard installed successfully")

		// Emit event so frontend timeline shows the installation
		if a.eventPipeline != nil {
			a.eventPipeline.EmitRaw(deviceId, SourceApp, "app_install", LevelInfo,
				"ADBKeyboard installed for Unicode text input support",
				map[string]interface{}{
					"packageName": adbKeyboardPackage,
					"action":      "auto_install",
					"reason":      "unicode_text_input",
				})
		}

		// Small delay after install to let the system register the IME
		time.Sleep(500 * time.Millisecond)
	}

	// Step 2: Enable and activate if not already active
	if !a.IsADBKeyboardActive(deviceId) {
		_, err := a.RunAdbCommand(deviceId, "shell ime enable "+adbKeyboardIME)
		if err != nil {
			return false, installed, fmt.Errorf("failed to enable ADBKeyboard IME: %w", err)
		}
		_, err = a.RunAdbCommand(deviceId, "shell ime set "+adbKeyboardIME)
		if err != nil {
			return false, installed, fmt.Errorf("failed to activate ADBKeyboard IME: %w", err)
		}
		LogDebug("adb_keyboard").Str("deviceId", deviceId).Msg("ADBKeyboard activated as current IME")

		// Wait for IME service to fully initialize and bind to the input field.
		// Without this delay, the first broadcast after activation is lost.
		time.Sleep(1 * time.Second)
	}

	return true, installed, nil
}

// InputTextViaADBKeyboard inputs text using ADBKeyboard's base64 broadcast.
// Supports any Unicode text including CJK, emoji, etc.
func (a *App) InputTextViaADBKeyboard(deviceId string, text string) error {
	ready, _, err := a.EnsureADBKeyboard(deviceId)
	if !ready {
		return fmt.Errorf("ADBKeyboard not ready: %w", err)
	}

	// Encode text as base64 to avoid shell escaping issues
	encoded := base64.StdEncoding.EncodeToString([]byte(text))

	cmd := fmt.Sprintf("shell am broadcast -a ADB_INPUT_B64 --es msg %s", encoded)
	output, err := a.RunAdbCommand(deviceId, cmd)
	if err != nil {
		return fmt.Errorf("ADBKeyboard broadcast failed: %w", err)
	}

	// ADBKeyboard broadcast returns "result=0" or "result=-1" on success
	if strings.Contains(output, "result=0") || strings.Contains(output, "result=-1") {
		return nil
	}

	LogDebug("adb_keyboard").Str("deviceId", deviceId).Str("output", output).Msg("Unexpected broadcast result")
	return nil
}

// ClearTextViaADBKeyboard clears the focused text field using ADBKeyboard
func (a *App) ClearTextViaADBKeyboard(deviceId string) error {
	ready, _, err := a.EnsureADBKeyboard(deviceId)
	if !ready {
		return fmt.Errorf("ADBKeyboard not ready: %w", err)
	}

	_, err = a.RunAdbCommand(deviceId, "shell am broadcast -a ADB_CLEAR_TEXT")
	return err
}

// containsNonASCII checks if a string contains any non-ASCII characters
func containsNonASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return true
		}
	}
	return false
}

// InputText is the unified text input entry point.
// For ASCII-only text, uses native "adb shell input text" (fast, no extra setup).
// For Unicode text, automatically installs and uses ADBKeyboard (lazy install).
func (a *App) InputText(deviceId string, text string) error {
	if containsNonASCII(text) {
		return a.InputTextViaADBKeyboard(deviceId, text)
	}

	// ASCII path: use native adb input text
	escaped := escapeForAdbInput(text)
	_, err := a.RunAdbCommand(deviceId, fmt.Sprintf("shell input text %s", escaped))
	return err
}

// escapeForAdbInput escapes a string for safe use with "adb shell input text".
// Only suitable for ASCII text.
func escapeForAdbInput(text string) string {
	// adb input text uses %s for spaces
	result := strings.ReplaceAll(text, " ", "%s")

	// Escape shell special characters
	shellSpecials := []string{
		"'", "\"", "`", "\\", "$",
		"(", ")", "{", "}", "[", "]",
		"&", "|", ";", "<", ">",
		"#", "!", "~", "*", "?",
	}
	for _, ch := range shellSpecials {
		result = strings.ReplaceAll(result, ch, "\\"+ch)
	}

	return result
}
