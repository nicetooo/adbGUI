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

// getCurrentIME returns the current active IME identifier
func (a *App) getCurrentIME(deviceId string) string {
	output, err := a.RunAdbCommand(deviceId, "shell settings get secure default_input_method")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}

// EnsureADBKeyboard installs ADBKeyboard on the device if not already present,
// and enables it in the IME list (but does NOT activate it as the current IME).
// This avoids interfering with the device's normal keyboard.
// Returns (installed, justInstalled, error).
func (a *App) EnsureADBKeyboard(deviceId string) (bool, bool, error) {
	if a.adbKeyboardPath == "" {
		return false, false, fmt.Errorf("ADBKeyboard APK not available (not embedded in this build)")
	}

	justInstalled := false

	if !a.IsADBKeyboardInstalled(deviceId) {
		LogDebug("adb_keyboard").Str("deviceId", deviceId).Msg("Installing ADBKeyboard APK...")
		cmd := a.newAdbCommand(nil, "-s", deviceId, "install", "-r", a.adbKeyboardPath)
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		if err != nil {
			return false, false, fmt.Errorf("failed to install ADBKeyboard: %w (output: %s)", err, output)
		}
		if !strings.Contains(output, "Success") {
			return false, false, fmt.Errorf("ADBKeyboard install did not succeed: %s", output)
		}
		justInstalled = true
		LogDebug("adb_keyboard").Str("deviceId", deviceId).Msg("ADBKeyboard installed successfully")

		if a.eventPipeline != nil {
			a.eventPipeline.EmitRaw(deviceId, SourceApp, "app_install", LevelInfo,
				"ADBKeyboard installed for Unicode text input support",
				map[string]interface{}{
					"packageName": adbKeyboardPackage,
					"action":      "auto_install",
					"reason":      "unicode_text_input",
				})
		}

		// Let the system register the new IME
		time.Sleep(500 * time.Millisecond)
	}

	// Enable in IME list (does NOT switch the active IME)
	a.RunAdbCommand(deviceId, "shell ime enable "+adbKeyboardIME)

	return true, justInstalled, nil
}

// activateADBKeyboard temporarily switches the active IME to ADBKeyboard.
// Returns the previous IME so it can be restored after input.
func (a *App) activateADBKeyboard(deviceId string) string {
	previousIME := a.getCurrentIME(deviceId)

	if previousIME == adbKeyboardIME {
		return previousIME // Already active
	}

	a.RunAdbCommand(deviceId, "shell ime set "+adbKeyboardIME)
	LogDebug("adb_keyboard").Str("deviceId", deviceId).Msg("ADBKeyboard temporarily activated")

	// Wait for IME service to bind to the input field
	time.Sleep(800 * time.Millisecond)

	return previousIME
}

// restoreIME switches back to the previous IME after ADBKeyboard input
func (a *App) restoreIME(deviceId string, previousIME string) {
	if previousIME == "" || previousIME == adbKeyboardIME {
		return
	}
	_, err := a.RunAdbCommand(deviceId, "shell ime set "+previousIME)
	if err != nil {
		LogDebug("adb_keyboard").Str("deviceId", deviceId).Err(err).Msg("Failed to restore previous IME")
	} else {
		LogDebug("adb_keyboard").Str("deviceId", deviceId).Str("ime", previousIME).Msg("Restored previous IME")
	}
}

// InputTextViaADBKeyboard inputs text using ADBKeyboard's base64 broadcast.
// Temporarily activates ADBKeyboard, sends input, then restores the previous IME.
func (a *App) InputTextViaADBKeyboard(deviceId string, text string) error {
	// Step 1: Ensure installed
	ready, _, err := a.EnsureADBKeyboard(deviceId)
	if !ready {
		return fmt.Errorf("ADBKeyboard not ready: %w", err)
	}

	// Step 2: Temporarily activate ADBKeyboard (save previous IME)
	previousIME := a.activateADBKeyboard(deviceId)

	// Step 3: Send input via base64 broadcast
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	cmd := fmt.Sprintf("shell am broadcast -a ADB_INPUT_B64 --es msg %s", encoded)
	output, broadcastErr := a.RunAdbCommand(deviceId, cmd)

	// Step 4: Restore previous IME immediately after input
	a.restoreIME(deviceId, previousIME)

	if broadcastErr != nil {
		return fmt.Errorf("ADBKeyboard broadcast failed: %w", broadcastErr)
	}

	if strings.Contains(output, "result=0") || strings.Contains(output, "result=-1") {
		return nil
	}

	LogDebug("adb_keyboard").Str("deviceId", deviceId).Str("output", output).Msg("Unexpected broadcast result")
	return nil
}

// ClearTextViaADBKeyboard clears the focused text field using ADBKeyboard.
// Temporarily activates ADBKeyboard, sends clear, then restores the previous IME.
func (a *App) ClearTextViaADBKeyboard(deviceId string) error {
	ready, _, err := a.EnsureADBKeyboard(deviceId)
	if !ready {
		return fmt.Errorf("ADBKeyboard not ready: %w", err)
	}

	previousIME := a.activateADBKeyboard(deviceId)
	_, clearErr := a.RunAdbCommand(deviceId, "shell am broadcast -a ADB_CLEAR_TEXT")
	a.restoreIME(deviceId, previousIME)

	return clearErr
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
// ADBKeyboard is only active during the brief moment of Unicode input,
// then the previous IME is restored to avoid interfering with normal device usage.
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
