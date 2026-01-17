package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Network Monitor State
var (
	monitorCancels = make(map[string]context.CancelFunc)
	monitorMu      sync.Mutex
)

// StartNetworkMonitor starts a goroutine to poll /proc/net/dev for a specific device
func (a *App) StartNetworkMonitor(deviceId string) {
	a.StopNetworkMonitor(deviceId)

	monitorMu.Lock()
	ctx, cancel := context.WithCancel(a.ctx)
	monitorCancels[deviceId] = cancel
	monitorMu.Unlock()

	go func() {
		var lastStats NetworkStats
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats, err := a.getNetworkStats(deviceId)
				if err != nil {
					continue
				}
				stats.DeviceId = deviceId

				if lastStats.Time > 0 && stats.Time > lastStats.Time {
					duration := float64(stats.Time - lastStats.Time)
					if duration > 0 {
						if stats.RxBytes >= lastStats.RxBytes {
							stats.RxSpeed = uint64(float64(stats.RxBytes-lastStats.RxBytes) / duration)
						}
						if stats.TxBytes >= lastStats.TxBytes {
							stats.TxSpeed = uint64(float64(stats.TxBytes-lastStats.TxBytes) / duration)
						}
					}
				}
				lastStats = stats

				wailsRuntime.EventsEmit(a.ctx, "network-stats", stats)
			}
		}
	}()
}

// StopNetworkMonitor stops the monitoring goroutine for a specific device
func (a *App) StopNetworkMonitor(deviceId string) {
	monitorMu.Lock()
	defer monitorMu.Unlock()
	if cancel, ok := monitorCancels[deviceId]; ok {
		cancel()
		delete(monitorCancels, deviceId)
	}
}

// StopAllNetworkMonitors stops all network monitoring
func (a *App) StopAllNetworkMonitors() {
	monitorMu.Lock()
	defer monitorMu.Unlock()
	for id, cancel := range monitorCancels {
		cancel()
		delete(monitorCancels, id)
	}
}

// SetDeviceNetworkLimit sets the ingress rate limit (Android 13+)
func (a *App) SetDeviceNetworkLimit(deviceId string, bytesPerSecond int) (string, error) {
	cmd := exec.Command(a.adbPath, "-s", deviceId, "shell", "settings", "put", "global", "ingress_rate_limit_bytes_per_second", fmt.Sprintf("%d", bytesPerSecond))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %v", string(output), err)
	}
	return "Network limit set successfully", nil
}

func (a *App) getNetworkStats(deviceId string) (NetworkStats, error) {
	var stats NetworkStats
	stats.Time = time.Now().Unix()

	cmd := exec.Command(a.adbPath, "-s", deviceId, "shell", "cat", "/proc/net/dev")
	output, err := cmd.Output()
	if err != nil {
		return stats, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "wlan0:") {
			fields := strings.Fields(strings.TrimPrefix(line, "wlan0:"))
			if len(fields) >= 9 {
				fmt.Sscanf(fields[0], "%d", &stats.RxBytes)
				fmt.Sscanf(fields[8], "%d", &stats.TxBytes)
			}
			break
		}
	}
	return stats, nil
}
