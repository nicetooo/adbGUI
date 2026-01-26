package main

import (
	"archive/zip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"Gaze/pkg/cache"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ListPackages returns a list of installed packages with their type and state
func (a *App) ListPackages(deviceId string, packageType string) ([]AppPackage, error) {
	if err := ValidateDeviceID(deviceId); err != nil {
		return nil, err
	}

	if packageType == "" {
		packageType = "user"
	}

	// Get list of disabled packages
	disabledPackages := make(map[string]bool)
	cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "pm", "list", "packages", "-d")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "package:") {
				disabledPackages[strings.TrimPrefix(line, "package:")] = true
			}
		}
	}

	var packages []AppPackage

	fetch := func(flag, typeName string) error {
		cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "pm", "list", "packages", flag)
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "package:") {
				name := strings.TrimPrefix(line, "package:")
				state := "enabled"
				if disabledPackages[name] {
					state = "disabled"
				}
				packages = append(packages, AppPackage{
					Name:  name,
					Type:  typeName,
					State: state,
				})
			}
		}
		return nil
	}

	if packageType == "all" {
		if err := fetch("-s", "system"); err != nil {
			return nil, fmt.Errorf("failed to list system packages: %w", err)
		}
		if err := fetch("-3", "user"); err != nil {
			return nil, fmt.Errorf("failed to list user packages: %w", err)
		}
	} else if packageType == "system" {
		if err := fetch("-s", "system"); err != nil {
			return nil, fmt.Errorf("failed to list system packages: %w", err)
		}
	} else {
		if err := fetch("-3", "user"); err != nil {
			return nil, fmt.Errorf("failed to list user packages: %w", err)
		}
	}

	// Fetch labels and icons from cache in parallel
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for i := range packages {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			pkg := &packages[idx]

			if a.cacheService != nil {
				if cached, ok := a.cacheService.GetCachedPackage(pkg.Name); ok {
					if cached.Label != "" {
						pkg.Label = cached.Label
					}
					if cached.Icon != "" {
						pkg.Icon = cached.Icon
					}
					if cached.VersionName != "" {
						pkg.VersionName = cached.VersionName
					}
					if cached.VersionCode != "" {
						pkg.VersionCode = cached.VersionCode
					}
					if cached.MinSdkVersion != "" {
						pkg.MinSdkVersion = cached.MinSdkVersion
					}
					if cached.TargetSdkVersion != "" {
						pkg.TargetSdkVersion = cached.TargetSdkVersion
					}
					if len(cached.Permissions) > 0 {
						pkg.Permissions = cached.Permissions
					}
				}
			}

			if pkg.Label == "" {
				brandMap := map[string]string{
					"com.google.android.youtube": "YouTube",
					"com.google.android.gms":     "Google Play Services",
					"com.android.vending":        "Google Play Store",
					"com.whatsapp":               "WhatsApp",
					"com.facebook.katana":        "Facebook",
					"com.facebook.orca":          "Messenger",
					"com.instagram.android":      "Instagram",
				}

				if brand, ok := brandMap[pkg.Name]; ok {
					pkg.Label = brand
				} else {
					parts := strings.Split(pkg.Name, ".")
					var meaningful []string
					skip := map[string]bool{
						"com": true, "net": true, "org": true, "android": true,
						"google": true, "ss": true, "ugc": true, "app": true,
					}
					for _, p := range parts {
						if !skip[strings.ToLower(p)] && len(p) > 2 {
							meaningful = append(meaningful, p)
						}
					}
					if len(meaningful) == 0 {
						meaningful = parts[len(parts)-1:]
					}
					for i, p := range meaningful {
						meaningful[i] = strings.ToUpper(p[:1]) + p[1:]
					}
					pkg.Label = strings.Join(meaningful, " ")
				}
			}
		}(i)
	}
	wg.Wait()

	return packages, nil
}

// GetAppInfo returns detailed information for a specific package
func (a *App) GetAppInfo(deviceId, packageName string, force bool) (AppPackage, error) {
	pkg, _ := a.getAdbDetailedInfo(deviceId, packageName)

	var cached cache.AppPackage
	var hasCache bool
	if a.cacheService != nil {
		cached, hasCache = a.cacheService.GetCachedPackage(packageName)
	}

	if force || !hasCache || cached.Label == "" || cached.LaunchableActivities == nil {
		detailedPkg, err := a.getAppInfoWithAapt(deviceId, packageName)
		if err == nil {
			pkg.Label = detailedPkg.Label
			pkg.Icon = detailedPkg.Icon
			pkg.VersionName = detailedPkg.VersionName
			pkg.VersionCode = detailedPkg.VersionCode
			pkg.MinSdkVersion = detailedPkg.MinSdkVersion
			pkg.TargetSdkVersion = detailedPkg.TargetSdkVersion
			pkg.LaunchableActivities = detailedPkg.LaunchableActivities

			if len(detailedPkg.Activities) > 0 {
				seen := make(map[string]bool)
				for _, act := range pkg.Activities {
					seen[act] = true
				}
				for _, act := range detailedPkg.Activities {
					if !seen[act] {
						pkg.Activities = append(pkg.Activities, act)
						seen[act] = true
					}
				}
			}
		}
	} else {
		pkg.Label = cached.Label
		pkg.Icon = cached.Icon
		pkg.VersionName = cached.VersionName
		pkg.VersionCode = cached.VersionCode
		pkg.MinSdkVersion = cached.MinSdkVersion
		pkg.TargetSdkVersion = cached.TargetSdkVersion
		pkg.LaunchableActivities = cached.LaunchableActivities

		if len(pkg.Activities) == 0 {
			pkg.Activities = cached.Activities
		}
	}

	return pkg, nil
}

func (a *App) getAdbDetailedInfo(deviceId, packageName string) (AppPackage, error) {
	var pkg AppPackage
	pkg.Name = packageName

	cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "dumpsys", "package", packageName)
	output, err := cmd.Output()
	if err != nil {
		return pkg, err
	}

	outputStr := string(output)
	pkg.Activities = a.parseActivitiesFromDumpsys(outputStr, packageName)
	pkg.Permissions = a.parsePermissionsFromDumpsys(outputStr)

	return pkg, nil
}

func (a *App) parseActivitiesFromDumpsys(output, packageName string) []string {
	var activities []string
	seen := make(map[string]bool)
	lines := strings.Split(output, "\n")
	inActivities := false

	pkgPattern := regexp.QuoteMeta(packageName)
	activityRegex := regexp.MustCompile(`(?i)(` + pkgPattern + `\/[\.\w\$]+)`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.EqualFold(trimmed, "Activities:") {
			inActivities = true
			continue
		}

		if inActivities {
			if len(line) > 0 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && strings.HasSuffix(trimmed, ":") {
				inActivities = false
				continue
			}

			matches := activityRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				act := a.normalizeActivityName(match[1], packageName)
				if !seen[act] {
					activities = append(activities, act)
					seen[act] = true
				}
			}
		}
	}

	if len(activities) == 0 {
		matches := activityRegex.FindAllStringSubmatch(output, -1)
		for _, match := range matches {
			act := a.normalizeActivityName(match[1], packageName)
			if !seen[act] {
				activities = append(activities, act)
				seen[act] = true
			}
		}
	}

	return activities
}

func (a *App) parsePermissionsFromDumpsys(output string) []string {
	var permissions []string
	lines := strings.Split(output, "\n")
	inPermissions := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "requested permissions:") {
			inPermissions = true
			continue
		}
		if inPermissions {
			if strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "android.permission") {
				inPermissions = false
				continue
			}
			if strings.HasPrefix(trimmed, "android.permission") || strings.Contains(trimmed, "permission") {
				perm := strings.Split(trimmed, ":")[0]
				permissions = append(permissions, strings.TrimSpace(perm))
			}
		}
	}
	return permissions
}

func (a *App) normalizeActivityName(activity, packageName string) string {
	if !strings.Contains(activity, "/") {
		if strings.HasPrefix(activity, ".") {
			return packageName + "/" + packageName + activity
		}
		return packageName + "/" + activity
	}

	parts := strings.SplitN(activity, "/", 2)
	pkg := parts[0]
	class := parts[1]

	if strings.HasPrefix(class, ".") {
		return pkg + "/" + pkg + class
	}
	return activity
}

func (a *App) getAppInfoWithAapt(deviceId, packageName string) (AppPackage, error) {
	var pkg AppPackage
	pkg.Name = packageName

	if a.aaptPath == "" {
		return pkg, fmt.Errorf("aapt not available (binary not embedded)")
	}

	if info, err := os.Stat(a.aaptPath); err != nil || info.Size() == 0 {
		return pkg, fmt.Errorf("aapt not available (file missing or empty)")
	}

	cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "pm", "path", packageName)
	output, err := cmd.Output()
	if err != nil {
		return pkg, fmt.Errorf("failed to get APK path: %w", err)
	}

	remotePath := strings.TrimSpace(string(output))
	if remotePath == "" {
		return pkg, fmt.Errorf("empty output from pm path for %s", packageName)
	}

	lines := strings.Split(remotePath, "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "package:") {
		return pkg, fmt.Errorf("unexpected output from pm path: %s", remotePath)
	}
	remotePath = strings.TrimPrefix(lines[0], "package:")

	tmpDir := filepath.Join(os.TempDir(), "adb-gui-apk")
	_ = os.MkdirAll(tmpDir, 0755)
	tmpAPK := filepath.Join(tmpDir, packageName+".apk")
	defer os.Remove(tmpAPK)

	pullCmd := a.newAdbCommand(nil, "-s", deviceId, "pull", remotePath, tmpAPK)
	pullOutput, err := pullCmd.CombinedOutput()
	if err != nil {
		return pkg, fmt.Errorf("failed to pull APK: %w (output: %s)", err, string(pullOutput))
	}

	aaptCmd := exec.Command(a.aaptPath, "dump", "badging", tmpAPK)
	aaptOutput, err := aaptCmd.CombinedOutput()
	if err != nil {
		return pkg, fmt.Errorf("failed to run aapt: %w, output: %s", err, string(aaptOutput))
	}

	outputStr := string(aaptOutput)
	pkg.Label = a.parseLabelFromAapt(outputStr)
	pkg.VersionName, pkg.VersionCode = a.parseVersionFromAapt(outputStr)
	pkg.MinSdkVersion = a.parseSdkVersionFromAapt(outputStr, "sdkVersion:")
	pkg.TargetSdkVersion = a.parseSdkVersionFromAapt(outputStr, "targetSdkVersion:")
	pkg.LaunchableActivities = a.parseActivitiesFromAapt(outputStr, packageName)
	pkg.Activities = pkg.LaunchableActivities

	icon, err := a.extractIconWithAapt(tmpAPK)
	if err == nil {
		pkg.Icon = icon
	}

	if a.cacheService != nil {
		// Convert to cache.AppPackage
		cachePkg := cache.AppPackage{
			Name:                 pkg.Name,
			Label:                pkg.Label,
			Icon:                 pkg.Icon,
			Type:                 pkg.Type,
			State:                pkg.State,
			VersionName:          pkg.VersionName,
			VersionCode:          pkg.VersionCode,
			MinSdkVersion:        pkg.MinSdkVersion,
			TargetSdkVersion:     pkg.TargetSdkVersion,
			Permissions:          pkg.Permissions,
			Activities:           pkg.Activities,
			LaunchableActivities: pkg.LaunchableActivities,
		}
		a.cacheService.SetCachedPackage(packageName, cachePkg)
		go a.saveCache()
	}

	return pkg, nil
}

func (a *App) parseVersionFromAapt(output string) (versionName, versionCode string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "package:") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "versionCode=") {
					versionCode = strings.Trim(strings.TrimPrefix(part, "versionCode="), "'\"")
				}
				if strings.HasPrefix(part, "versionName=") {
					versionName = strings.Trim(strings.TrimPrefix(part, "versionName="), "'\"")
				}
			}
			return
		}
	}
	return
}

func (a *App) parseSdkVersionFromAapt(output, prefix string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			val := strings.TrimPrefix(line, prefix)
			return strings.Trim(val, "'\"")
		}
	}
	return ""
}

func (a *App) parseLabelFromAapt(output string) string {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "application-label:") {
			label := strings.TrimPrefix(line, "application-label:")
			label = strings.Trim(label, "'\"")
			label = strings.TrimSpace(label)
			if label != "" {
				return label
			}
		}
	}

	preferredLocales := []string{"en", "zh-TW", "zh-CN", "zh", ""}
	for _, locale := range preferredLocales {
		prefix := "application-label"
		if locale != "" {
			prefix = fmt.Sprintf("application-label-%s", locale)
		}

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, prefix+":") {
				label := strings.TrimPrefix(line, prefix+":")
				label = strings.Trim(label, "'\"")
				label = strings.TrimSpace(label)
				if label != "" {
					return label
				}
			}
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "application-label-") && strings.Contains(line, ":") {
			idx := strings.Index(line, ":")
			if idx > 0 && idx < len(line)-1 {
				label := line[idx+1:]
				label = strings.Trim(label, "'\"")
				label = strings.TrimSpace(label)
				if label != "" {
					return label
				}
			}
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "application:") && strings.Contains(line, "label='") {
			idx := strings.Index(line, "label='")
			if idx > 0 {
				start := idx + 7
				end := strings.Index(line[start:], "'")
				if end > 0 {
					label := line[start : start+end]
					if label != "" {
						return label
					}
				}
			}
		}
	}

	return ""
}

func (a *App) parseActivitiesFromAapt(output, packageName string) []string {
	var activities []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "launchable-activity:") {
			idx := strings.Index(line, "name='")
			if idx > 0 {
				start := idx + 6
				end := strings.Index(line[start:], "'")
				if end > 0 {
					name := line[start : start+end]
					name = a.normalizeActivityName(name, packageName)
					activities = append(activities, name)
				}
			}
		}
	}
	return activities
}

func (a *App) extractIconWithAapt(apkPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, a.aaptPath, "dump", "badging", apkPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run aapt dump badging: %w, output: %s", err, string(output))
	}

	outputStr := string(output)
	iconPath := a.parseIconPathFromAapt(outputStr)
	if iconPath == "" {
		iconPath = a.parseIconPathFromAapt2(outputStr)
	}
	if iconPath == "" {
		return "", fmt.Errorf("icon path not found in aapt output")
	}

	iconData, err := a.extractFileFromAPK(apkPath, iconPath)
	if err != nil {
		altPaths := a.getAlternativeIconPaths(iconPath)
		for _, altPath := range altPaths {
			if data, err2 := a.extractFileFromAPK(apkPath, altPath); err2 == nil {
				iconData = data
				iconPath = altPath
				err = nil
				break
			}
		}
		if err != nil {
			return "", fmt.Errorf("failed to extract icon from APK: %w", err)
		}
	}

	var mimeType string
	if strings.HasSuffix(iconPath, ".png") {
		mimeType = "image/png"
	} else if strings.HasSuffix(iconPath, ".jpg") || strings.HasSuffix(iconPath, ".jpeg") {
		mimeType = "image/jpeg"
	} else if strings.HasSuffix(iconPath, ".webp") {
		mimeType = "image/webp"
	} else {
		mimeType = "image/png"
	}

	base64Str := base64.StdEncoding.EncodeToString(iconData)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str), nil
}

func (a *App) parseIconPathFromAapt(output string) string {
	if output == "" {
		return ""
	}

	lines := strings.Split(output, "\n")

	iconSizes := []string{"480", "320", "240", "160", "120", "80", "48"}
	for _, size := range iconSizes {
		prefix := fmt.Sprintf("application-icon-%s:", size)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, prefix) {
				iconPath := strings.TrimPrefix(line, prefix)
				iconPath = strings.Trim(iconPath, "'\"")
				iconPath = strings.TrimSpace(iconPath)
				if iconPath != "" {
					return iconPath
				}
			}
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "application-icon:") {
			iconPath := strings.TrimPrefix(line, "application-icon:")
			iconPath = strings.Trim(iconPath, "'\"")
			iconPath = strings.TrimSpace(iconPath)
			if iconPath != "" {
				return iconPath
			}
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "application:") && strings.Contains(line, "icon='") {
			idx := strings.Index(line, "icon='")
			if idx > 0 {
				start := idx + 6
				end := strings.Index(line[start:], "'")
				if end > 0 {
					iconPath := line[start : start+end]
					if iconPath != "" {
						return iconPath
					}
				}
			}
		}
	}

	return ""
}

func (a *App) parseIconPathFromAapt2(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "icon=") {
			parts := strings.Split(line, "icon=")
			if len(parts) >= 2 {
				iconPath := strings.Trim(parts[1], "'\"")
				iconPath = strings.TrimSpace(iconPath)
				if idx := strings.IndexAny(iconPath, " \t\n"); idx > 0 {
					iconPath = iconPath[:idx]
				}
				if iconPath != "" && (strings.HasSuffix(iconPath, ".png") ||
					strings.HasSuffix(iconPath, ".jpg") ||
					strings.HasSuffix(iconPath, ".jpeg") ||
					strings.HasSuffix(iconPath, ".webp")) {
					return iconPath
				}
			}
		}
		if strings.Contains(line, "package:") && strings.Contains(line, "icon") {
			if idx := strings.Index(line, "icon='"); idx > 0 {
				start := idx + 6
				if end := strings.Index(line[start:], "'"); end > 0 {
					iconPath := line[start : start+end]
					if iconPath != "" {
						return iconPath
					}
				}
			}
		}
	}
	return ""
}

func (a *App) getAlternativeIconPaths(originalPath string) []string {
	var alternatives []string

	densities := []string{"xxxhdpi", "xxhdpi", "xhdpi", "hdpi", "mdpi", "ldpi"}
	for _, density := range densities {
		if strings.Contains(originalPath, "mipmap-") {
			alt := strings.Replace(originalPath, "mipmap-", "mipmap-"+density+"-", 1)
			alternatives = append(alternatives, alt)
		}
		if strings.Contains(originalPath, "drawable-") {
			alt := strings.Replace(originalPath, "drawable-", "drawable-"+density+"-", 1)
			alternatives = append(alternatives, alt)
		}
	}

	iconNames := []string{"ic_launcher.png", "ic_launcher_foreground.png", "ic_launcher_round.png", "icon.png"}
	baseDir := filepath.Dir(originalPath)
	for _, iconName := range iconNames {
		alternatives = append(alternatives, filepath.Join(baseDir, iconName))
	}

	return alternatives
}

func (a *App) extractFileFromAPK(apkPath, filePath string) ([]byte, error) {
	r, err := zip.OpenReader(apkPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == filePath {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}

	if strings.HasPrefix(filePath, "res/") {
		filePath = strings.TrimPrefix(filePath, "res/")
		for _, f := range r.File {
			if f.Name == filePath || strings.HasSuffix(f.Name, filePath) {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
	}

	fileName := filepath.Base(filePath)
	for _, f := range r.File {
		if strings.Contains(f.Name, fileName) && (strings.Contains(f.Name, "mipmap") || strings.Contains(f.Name, "drawable")) {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err == nil {
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("file not found in APK: %s", filePath)
}

// App control functions

// UninstallApp uninstalls an app
func (a *App) UninstallApp(deviceId, packageName string) (string, error) {
	a.updateLastActive(deviceId)
	if err := ValidateDeviceID(deviceId); err != nil {
		return "", err
	}

	a.Log("Uninstalling %s from %s", packageName, deviceId)

	cmd := a.newAdbCommand(nil, "-s", deviceId, "uninstall", packageName)
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err == nil && !strings.Contains(outStr, "Failure") {
		return outStr, nil
	}

	LogDebug("apps").Str("package", packageName).Str("output", outStr).Msg("Standard uninstall failed, trying pm uninstall --user 0")
	cmd2 := a.newAdbCommand(nil, "-s", deviceId, "shell", "pm", "uninstall", "-k", "--user", "0", packageName)
	output2, err2 := cmd2.CombinedOutput()
	outStr2 := string(output2)
	if err2 != nil || strings.Contains(outStr2, "Failure") {
		return outStr2, fmt.Errorf("failed to uninstall: %s", outStr2)
	}

	return outStr2, nil
}

// ClearAppData clears the application data
func (a *App) ClearAppData(deviceId, packageName string) (string, error) {
	if deviceId == "" {
		return "", fmt.Errorf("no device specified")
	}
	cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "pm", "clear", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to clear data: %w", err)
	}
	return string(output), nil
}

// ForceStopApp force stops the application
func (a *App) ForceStopApp(deviceId, packageName string) (string, error) {
	if deviceId == "" {
		return "", fmt.Errorf("no device specified")
	}
	cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "am", "force-stop", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to force stop: %w", err)
	}
	return string(output), nil
}

// StartApp launches the application using monkey command
func (a *App) StartApp(deviceId, packageName string) (string, error) {
	a.updateLastActive(deviceId)
	if deviceId == "" {
		return "", fmt.Errorf("no device specified")
	}
	cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "monkey", "-p", packageName, "-c", "android.intent.category.LAUNCHER", "1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to start app: %w", err)
	}
	return string(output), nil
}

// EnableApp enables the application
func (a *App) EnableApp(deviceId, packageName string) (string, error) {
	if deviceId == "" {
		return "", fmt.Errorf("no device specified")
	}
	cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "pm", "enable", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to enable app: %w", err)
	}
	return string(output), nil
}

// DisableApp disables the application
func (a *App) DisableApp(deviceId, packageName string) (string, error) {
	if deviceId == "" {
		return "", fmt.Errorf("no device specified")
	}
	cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "pm", "disable-user", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to disable app: %w", err)
	}
	return string(output), nil
}

// StartActivity launches a specific activity
func (a *App) StartActivity(deviceId, activityName string) (string, error) {
	if deviceId == "" {
		return "", fmt.Errorf("no device specified")
	}
	cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "am", "start", "-n", activityName)
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err != nil {
		return outStr, fmt.Errorf("failed to start activity: %w", err)
	}

	if strings.Contains(outStr, "Error:") || strings.Contains(outStr, "Exception") || strings.Contains(outStr, "requires") {
		return outStr, fmt.Errorf("failed to start activity: %s", outStr)
	}

	return outStr, nil
}

// OpenSettings opens a specific system settings page
func (a *App) OpenSettings(deviceId string, action string, data string) (string, error) {
	if deviceId == "" {
		return "", fmt.Errorf("no device specified")
	}

	if action == "" {
		action = "android.settings.SETTINGS"
	}

	args := []string{"-s", deviceId, "shell", "am", "start", "-a", action}
	if data != "" {
		args = append(args, "-d", data)
	}

	cmd := a.newAdbCommand(nil, args...)
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err != nil {
		return outStr, fmt.Errorf("failed to open settings: %w", err)
	}

	if strings.Contains(outStr, "Error:") || strings.Contains(outStr, "Exception") {
		return outStr, fmt.Errorf("failed to open settings: %s", outStr)
	}

	return outStr, nil
}

// IsAppRunning checks if the given package is currently running on the device
func (a *App) IsAppRunning(deviceId, packageName string) (bool, error) {
	if deviceId == "" || packageName == "" {
		return false, nil
	}
	cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "pidof", packageName)
	out, _ := cmd.Output()
	if len(strings.TrimSpace(string(out))) > 0 {
		return true, nil
	}

	cmd2 := a.newAdbCommand(nil, "-s", deviceId, "shell", "pgrep", "-f", packageName)
	out2, _ := cmd2.Output()
	if len(strings.TrimSpace(string(out2))) > 0 {
		return true, nil
	}

	return false, nil
}

// InstallAPK installs an APK to the specified device
func (a *App) InstallAPK(deviceId string, path string) (string, error) {
	if err := ValidateDeviceID(deviceId); err != nil {
		return "", err
	}

	a.Log("Installing APK %s to device %s", path, deviceId)

	cmd := a.newAdbCommand(nil, "-s", deviceId, "install", "-r", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to install APK: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// ExportAPK extracts an installed APK from the device to the local machine
func (a *App) ExportAPK(deviceId string, packageName string) (string, error) {
	if err := ValidateDeviceID(deviceId); err != nil {
		return "", err
	}

	cmd := a.newAdbCommand(nil, "-s", deviceId, "shell", "pm", "path", packageName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get APK path: %w", err)
	}

	remotePath := strings.TrimSpace(string(output))
	lines := strings.Split(remotePath, "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "package:") {
		return "", fmt.Errorf("unexpected output from pm path: %s", remotePath)
	}
	remotePath = strings.TrimPrefix(lines[0], "package:")

	fileName := packageName + ".apk"
	defaultDir, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(defaultDir, "Downloads")
	if _, err := os.Stat(downloadsDir); err == nil {
		defaultDir = downloadsDir
	}

	savePath, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		DefaultFilename: fileName,
		Title:           "Export APK",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Android Package (*.apk)", Pattern: "*.apk"},
		},
		DefaultDirectory: defaultDir,
	})

	if err != nil {
		return "", fmt.Errorf("failed to open save dialog: %w", err)
	}
	if savePath == "" {
		return "", nil
	}

	pullCmd := a.newAdbCommand(nil, "-s", deviceId, "pull", remotePath, savePath)
	pullOutput, err := pullCmd.CombinedOutput()
	if err != nil {
		return string(pullOutput), fmt.Errorf("failed to pull APK: %w (output: %s)", err, string(pullOutput))
	}

	return savePath, nil
}
