package main

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// logcatLinePattern matches Android logcat time format: "01-04 12:34:56.789 D/Tag( 1234): message"
var logcatLinePattern = regexp.MustCompile(`^\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}\.\d{3}\s+([VDIWEF])/([^(]+)\(\s*\d+\):\s*(.*)`)

// parseLogcatLine extracts level, tag, and message from a logcat line
func parseLogcatLine(line string) (level, tag, message string, ok bool) {
	matches := logcatLinePattern.FindStringSubmatch(line)
	if len(matches) < 4 {
		return "", "", "", false
	}
	return matches[1], strings.TrimSpace(matches[2]), matches[3], true
}

// logcatLevelToSessionLevel converts Android log level to session level
func logcatLevelToSessionLevel(androidLevel string) string {
	switch androidLevel {
	case "E", "F": // Error, Fatal
		return "error"
	case "W":
		return "warn"
	case "I":
		return "info"
	case "D":
		return "debug"
	default: // V (Verbose)
		return "verbose"
	}
}

// StartLogcat starts the logcat stream for a device
func (a *App) StartLogcat(deviceId, packageName, preFilter string, preUseRegex bool, excludeFilter string, excludeUseRegex bool) error {
	// 验证 deviceId 格式
	if err := ValidateDeviceID(deviceId); err != nil {
		return err
	}

	a.updateLastActive(deviceId)

	if a.logcatCmd != nil {
		a.StopLogcat()
	}

	ctx, cancel := context.WithCancel(a.ctx)
	a.logcatCancel = cancel

	var cmd *exec.Cmd
	shellCmd := "logcat -v time"

	if preFilter != "" {
		grepCmd := "grep -i"
		if preUseRegex {
			grepCmd += "E"
		}
		safeFilter := strings.ReplaceAll(preFilter, "'", "'\\''")
		shellCmd += fmt.Sprintf(" | %s '%s'", grepCmd, safeFilter)
	}

	if excludeFilter != "" {
		grepCmd := "grep -iv"
		if excludeUseRegex {
			grepCmd += "E"
		}
		safeExclude := strings.ReplaceAll(excludeFilter, "'", "'\\''")
		shellCmd += fmt.Sprintf(" | %s '%s'", grepCmd, safeExclude)
	}

	if preFilter != "" || excludeFilter != "" {
		cmd = exec.CommandContext(ctx, a.adbPath, "-s", deviceId, "shell", shellCmd)
	} else {
		cmd = exec.CommandContext(ctx, a.adbPath, "-s", deviceId, "logcat", "-v", "time")
	}
	a.logcatCmd = cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		a.logcatCmd = nil
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		a.logcatCmd = nil
		return fmt.Errorf("failed to start logcat: %w", err)
	}

	var currentPids []string
	var currentUid string
	var pidMutex sync.RWMutex

	if packageName != "" {
		uidCmd := exec.Command(a.adbPath, "-s", deviceId, "shell", "pm list packages -U "+packageName)
		uidOut, _ := uidCmd.Output()
		uidStr := string(uidOut)
		if strings.Contains(uidStr, "uid:") {
			parts := strings.Split(uidStr, "uid:")
			if len(parts) > 1 {
				currentUid = strings.TrimSpace(strings.Fields(parts[1])[0])
			}
		}
	}

	if packageName != "" {
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			checkPid := func() {
				c := exec.Command(a.adbPath, "-s", deviceId, "shell", "pgrep -f", packageName)
				out, _ := c.Output()
				raw := strings.TrimSpace(string(out))

				if raw == "" {
					c2 := exec.Command(a.adbPath, "-s", deviceId, "shell", "pidof", packageName)
					out2, _ := c2.Output()
					raw = strings.TrimSpace(string(out2))
				}

				if raw == "" {
					c3 := exec.Command(a.adbPath, "-s", deviceId, "shell", "ps -A")
					out3, _ := c3.Output()
					lines := strings.Split(string(out3), "\n")
					var matchedPids []string
					for _, line := range lines {
						if strings.Contains(line, packageName) {
							fields := strings.Fields(line)
							if len(fields) > 1 {
								matchedPids = append(matchedPids, fields[1])
							}
						}
					}
					raw = strings.Join(matchedPids, " ")
				}

				pids := strings.Fields(raw)

				pidMutex.Lock()
				changed := len(pids) != len(currentPids)
				if !changed {
					for i, p := range pids {
						if p != currentPids[i] {
							changed = true
							break
						}
					}
				}

				if changed {
					currentPids = pids
					if len(pids) > 0 {
						status := fmt.Sprintf("--- Monitoring %s (UID: %s, PIDs: %s) ---", packageName, currentUid, strings.Join(pids, ", "))
						if !a.mcpMode {
							wailsRuntime.EventsEmit(a.ctx, "logcat-data", status)
						}
					} else {
						if !a.mcpMode {
							wailsRuntime.EventsEmit(a.ctx, "logcat-data", fmt.Sprintf("--- Waiting for %s processes... ---", packageName))
						}
					}
				}
				pidMutex.Unlock()
			}

			checkPid()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					checkPid()
				}
			}
		}()
	}

	// Channel for parsed log events
	logEvtChan := make(chan map[string]interface{}, 1000)

	// Reader & Filter Routine
	go func() {
		reader := bufio.NewReader(stdout)
		defer close(logEvtChan)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			if packageName != "" {
				pidMutex.RLock()
				pids := currentPids
				uid := currentUid
				pidMutex.RUnlock()

				if len(pids) > 0 {
					found := false
					for _, pid := range pids {
						if strings.Contains(line, "("+pid+")") ||
							strings.Contains(line, "( "+pid+")") ||
							strings.Contains(line, "("+pid+" )") ||
							strings.Contains(line, "["+pid+"]") ||
							strings.Contains(line, "[ "+pid+"]") ||
							strings.Contains(line, " "+pid+":") ||
							strings.Contains(line, "/"+pid+"(") ||
							strings.Contains(line, " "+pid+" ") ||
							strings.Contains(line, " "+pid+"):") ||
							strings.Contains(line, " "+pid+":") {
							found = true
							break
						}
					}

					if !found && uid != "" && strings.Contains(line, " "+uid+" ") {
						found = true
					}

					if !found {
						continue
					}
				} else {
					continue
				}
			}

			if level, tag, message, ok := parseLogcatLine(line); ok {
				logEvtChan <- map[string]interface{}{
					"tag":         tag,
					"message":     message,
					"level":       level,
					"packageName": packageName,
					"raw":         strings.TrimSpace(line),
				}
			}
		}
	}()

	// Aggregator & Emitter Routine
	go func() {
		var buffer []map[string]interface{}
		var lastTag string
		var lastLevel string
		var lastActivityTime time.Time

		// Regular ticker as requested (300ms)
		flushTicker := time.NewTicker(300 * time.Millisecond)
		defer flushTicker.Stop()

		flush := func() {
			if len(buffer) == 0 {
				return
			}

			sessionLevel := logcatLevelToSessionLevel(lastLevel)
			title := fmt.Sprintf("[%s] %s", lastTag, buffer[0]["message"])

			if len(buffer) > 1 {
				title = fmt.Sprintf("Logcat Output (%d entries) - %s", len(buffer), lastTag)
			}

			// Emit complete aggregated event
			a.EmitSessionEvent(deviceId, "logcat", "log", sessionLevel, title, buffer)

			// Clear buffer
			buffer = nil
		}

		for {
			select {
			case evt, ok := <-logEvtChan:
				if !ok {
					flush()
					return
				}

				tag := evt["tag"].(string)
				level := evt["level"].(string)
				now := time.Now()

				// Flush IMMEDIATELY if Tag or Level changes
				// This ensures we don't mix different types
				if len(buffer) > 0 {
					if tag != lastTag || level != lastLevel {
						flush()
					}
				}

				if len(buffer) == 0 {
					lastTag = tag
					lastLevel = level
				}

				buffer = append(buffer, evt)
				lastActivityTime = now

				// Safety valve: Flush if buffer gets massive (e.g. infinite loop of same log)
				if len(buffer) >= 2000 {
					flush()
				}

			case <-flushTicker.C:
				// Ticker fires every 300ms.
				// User requirement: "If aggregating (active), wait. If done/idle, flush."

				// If buffer is empty, nothing to do.
				if len(buffer) == 0 {
					continue
				}

				// Check if we are "active"
				// If we received a log recently (e.g. < 100ms ago), we assume we are in the middle of a burst.
				// In this case, we SKIP the timer flush to avoid splitting the burst.
				timeSinceLastLog := time.Since(lastActivityTime)
				if timeSinceLastLog < 100*time.Millisecond {
					continue
				}

				// If we haven't received logs for >100ms, assume the specific aggregation "block" has finished (paused).
				flush()

			case <-ctx.Done():
				flush()
				return
			}
		}
	}()

	return nil
}

// StopLogcat stops the logcat stream
func (a *App) StopLogcat() {
	if a.logcatCancel != nil {
		a.logcatCancel()
	}
	if a.logcatCmd != nil && a.logcatCmd.Process != nil {
		_ = a.logcatCmd.Process.Kill()
	}
	a.logcatCmd = nil
	a.logcatCancel = nil
}
