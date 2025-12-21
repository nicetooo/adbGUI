package main

import (
	"context"
	"embed"
	"runtime"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/icon.svg
var iconData []byte

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	var shouldQuit bool

	// Create application menu
	var applicationMenu *menu.Menu
	if runtime.GOOS == "darwin" {
		applicationMenu = menu.NewMenu()
		applicationMenu.Append(menu.AppMenu())
		applicationMenu.Append(menu.WindowMenu())
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "adbGUI",
		Width:     1280,
		Height:    720,
		MinWidth:  1280,
		MinHeight: 720,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Menu:             applicationMenu,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			// Initialize system tray
			if runtime.GOOS == "darwin" {
				start, _ := systray.RunWithExternalLoop(func() {
					systray.SetIcon(iconData)
					systray.SetTooltip("adbGUI")

					mShow := systray.AddMenuItem("Open adbGUI", "Show the main window")
					mShow.Click(func() {
						wailsRuntime.WindowShow(ctx)
					})

					mQuit := systray.AddMenuItem("Quit", "Quit adbGUI")
					mQuit.Click(func() {
						shouldQuit = true
						systray.Quit()
						wailsRuntime.Quit(ctx)
					})
				}, func() {})
				start()
			}
		},
		WindowStartState: options.Normal,
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			if runtime.GOOS == "darwin" && !shouldQuit {
				wailsRuntime.WindowHide(ctx)
				return true // Prevent closing
			}
			return false
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "adbGUI",
				Message: "A modern ADB GUI tool",
			},
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
