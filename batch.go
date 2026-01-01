package main

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ExecuteBatchOperation executes an operation on multiple devices in parallel
func (a *App) ExecuteBatchOperation(op BatchOperation) BatchOperationResult {
	result := BatchOperationResult{
		TotalDevices: len(op.DeviceIDs),
		Results:      make([]BatchResult, 0, len(op.DeviceIDs)),
	}

	if len(op.DeviceIDs) == 0 {
		return result
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	resultsChan := make(chan BatchResult, len(op.DeviceIDs))

	for _, deviceID := range op.DeviceIDs {
		wg.Add(1)
		go func(devID string) {
			defer wg.Done()

			var br BatchResult
			br.DeviceID = devID

			switch op.Type {
			case "install":
				br = a.batchInstall(devID, op.APKPath)
			case "uninstall":
				br = a.batchUninstall(devID, op.PackageName)
			case "clear":
				br = a.batchClearData(devID, op.PackageName)
			case "stop":
				br = a.batchForceStop(devID, op.PackageName)
			case "shell":
				br = a.batchShellCommand(devID, op.Command)
			case "push":
				br = a.batchPushFile(devID, op.LocalPath, op.RemotePath)
			case "reboot":
				br = a.batchReboot(devID)
			default:
				br.Error = fmt.Sprintf("unknown operation type: %s", op.Type)
			}

			br.DeviceID = devID
			resultsChan <- br

			// Emit progress event
			wailsRuntime.EventsEmit(a.ctx, "batch-progress", br)
		}(deviceID)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for br := range resultsChan {
		mu.Lock()
		result.Results = append(result.Results, br)
		if br.Success {
			result.SuccessCount++
		} else {
			result.FailureCount++
		}
		mu.Unlock()
	}

	return result
}

func (a *App) batchInstall(deviceID, apkPath string) BatchResult {
	br := BatchResult{DeviceID: deviceID}

	if apkPath == "" {
		br.Error = "no APK path specified"
		return br
	}

	cmd := exec.Command(a.adbPath, "-s", deviceID, "install", "-r", apkPath)
	output, err := cmd.CombinedOutput()
	br.Output = string(output)

	if err != nil {
		br.Error = err.Error()
		return br
	}

	if strings.Contains(br.Output, "Success") {
		br.Success = true
	} else if strings.Contains(br.Output, "Failure") {
		br.Error = br.Output
	} else {
		br.Success = true
	}

	return br
}

func (a *App) batchUninstall(deviceID, packageName string) BatchResult {
	br := BatchResult{DeviceID: deviceID}

	if packageName == "" {
		br.Error = "no package name specified"
		return br
	}

	// Try standard uninstall first
	cmd := exec.Command(a.adbPath, "-s", deviceID, "uninstall", packageName)
	output, err := cmd.CombinedOutput()
	br.Output = string(output)

	if err == nil && !strings.Contains(br.Output, "Failure") {
		br.Success = true
		return br
	}

	// Try pm uninstall for system apps
	cmd2 := exec.Command(a.adbPath, "-s", deviceID, "shell", "pm", "uninstall", "-k", "--user", "0", packageName)
	output2, err2 := cmd2.CombinedOutput()
	br.Output = string(output2)

	if err2 != nil || strings.Contains(br.Output, "Failure") {
		br.Error = br.Output
		return br
	}

	br.Success = true
	return br
}

func (a *App) batchClearData(deviceID, packageName string) BatchResult {
	br := BatchResult{DeviceID: deviceID}

	if packageName == "" {
		br.Error = "no package name specified"
		return br
	}

	cmd := exec.Command(a.adbPath, "-s", deviceID, "shell", "pm", "clear", packageName)
	output, err := cmd.CombinedOutput()
	br.Output = string(output)

	if err != nil {
		br.Error = err.Error()
		return br
	}

	if strings.Contains(br.Output, "Success") {
		br.Success = true
	} else {
		br.Error = br.Output
	}

	return br
}

func (a *App) batchForceStop(deviceID, packageName string) BatchResult {
	br := BatchResult{DeviceID: deviceID}

	if packageName == "" {
		br.Error = "no package name specified"
		return br
	}

	cmd := exec.Command(a.adbPath, "-s", deviceID, "shell", "am", "force-stop", packageName)
	output, err := cmd.CombinedOutput()
	br.Output = string(output)

	if err != nil {
		br.Error = err.Error()
		return br
	}

	br.Success = true
	return br
}

func (a *App) batchShellCommand(deviceID, command string) BatchResult {
	br := BatchResult{DeviceID: deviceID}

	if command == "" {
		br.Error = "no command specified"
		return br
	}

	cmd := exec.Command(a.adbPath, "-s", deviceID, "shell", command)
	output, err := cmd.CombinedOutput()
	br.Output = string(output)

	if err != nil {
		br.Error = err.Error()
		return br
	}

	br.Success = true
	return br
}

func (a *App) batchPushFile(deviceID, localPath, remotePath string) BatchResult {
	br := BatchResult{DeviceID: deviceID}

	if localPath == "" || remotePath == "" {
		br.Error = "local path and remote path are required"
		return br
	}

	cmd := exec.Command(a.adbPath, "-s", deviceID, "push", localPath, remotePath)
	output, err := cmd.CombinedOutput()
	br.Output = string(output)

	if err != nil {
		br.Error = err.Error()
		return br
	}

	br.Success = true
	return br
}

func (a *App) batchReboot(deviceID string) BatchResult {
	br := BatchResult{DeviceID: deviceID}

	cmd := exec.Command(a.adbPath, "-s", deviceID, "reboot")
	output, err := cmd.CombinedOutput()
	br.Output = string(output)

	if err != nil {
		br.Error = err.Error()
		return br
	}

	br.Success = true
	return br
}

// SelectAPKForBatch opens a file dialog to select an APK file
func (a *App) SelectAPKForBatch() (string, error) {
	path, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select APK",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Android Package (*.apk)", Pattern: "*.apk"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

// SelectFileForBatch opens a file dialog to select a file for pushing
func (a *App) SelectFileForBatch() (string, error) {
	path, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select File to Push",
	})
	if err != nil {
		return "", err
	}
	return path, nil
}
