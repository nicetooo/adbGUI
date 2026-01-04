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
	a.updateLastActive(deviceId)

	if a.logcatCmd != nil {
		a.StopLogcat()
	}

	ctx, cancel := context.WithCancel(context.Background())
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
						wailsRuntime.EventsEmit(a.ctx, "logcat-data", status)
					} else {
						wailsRuntime.EventsEmit(a.ctx, "logcat-data", fmt.Sprintf("--- Waiting for %s processes... ---", packageName))
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

	go func() {
		reader := bufio.NewReader(stdout)

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

			// Emit ALL logs to Session (no rate limiting - batch sync handles performance)
			if level, tag, message, ok := parseLogcatLine(line); ok {
				sessionLevel := logcatLevelToSessionLevel(level)
				a.EmitSessionEvent(deviceId, "logcat", "log", sessionLevel,
					fmt.Sprintf("[%s] %s", tag, message),
					map[string]interface{}{
						"tag":         tag,
						"message":     message,
						"level":       level,
						"packageName": packageName,
						"raw":         strings.TrimSpace(line),
					})
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
