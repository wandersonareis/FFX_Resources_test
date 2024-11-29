package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte
var logFile *os.File

func logToFile() {
	currentTime := time.Now().Format("02-01-2006")
	fileName := fmt.Sprintf("tracker-%s.log", currentTime)

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}

	logFile = file
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.LstdFlags | log.Lshortfile)
}

func main() {
	// Create an instance of the app structure

	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r.(error).Error())
		}
	}()

	logToFile()
	app := NewApp()
	

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "Final Fantasy X/X2 HD Remaster Resources Editor",
		Width:     800,
		Height:    450,
		MinWidth:  700,
		MinHeight: 450,
		//MaxWidth:          1280,
		//MaxHeight:         800,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Menu:                     nil,
		EnableDefaultContextMenu: true,
		Logger:                   nil,
		LogLevel:                 logger.TRACE,
		OnStartup:                app.startup,
		OnDomReady:               app.domReady,
		OnBeforeClose:            app.beforeClose,
		OnShutdown:               app.shutdown,
		WindowStartState:         options.Normal,
		Bind: []interface{}{
			app,
			app.CollectionService,
			app.ExtractService,
			app.CompressService,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// DisableFramelessWindowDecorations: false,
			WebviewUserDataPath: "",
			ZoomFactor:          1.0,
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "FFX Resources",
				Message: "",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
