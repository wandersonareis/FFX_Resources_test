package main

import (
	"context"
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/lib"
	"ffxresources/backend/services"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type AppConfig struct {
	filePath          string
	GameFilesLocation string
	ExtractLocation   string
	TranslateLocation string
	ReimportLocation  string
}

// App struct
type App struct {
	ctx               context.Context
	appConfig         *AppConfig
	CollectionService *services.CollectionService
	ExtractService    *services.ExtractService
	CompressService   *services.CompressService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		CollectionService: services.NewCollectionService(),
		ExtractService:    services.NewExtractService(),
		CompressService:   services.NewCompressService(),

		appConfig: &AppConfig{
			filePath: common.PathJoin(common.GetExecDir(), "config.json"),
		},
	}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx

	lib.NewInteractionWithCtx(ctx)

	config, err := a.loadConfig(a.appConfig.filePath)
	if err != nil {
		fmt.Println(err)
	}
	a.appConfig.GameFilesLocation = config.GameFilesLocation
	a.appConfig.ExtractLocation = config.ExtractLocation
	a.appConfig.TranslateLocation = config.TranslateLocation
	a.appConfig.ReimportLocation = config.ReimportLocation

	lib.NewInteraction().GameLocation.SetPath(config.GameFilesLocation)
	lib.NewInteraction().ExtractLocation.SetPath(config.ExtractLocation)
	lib.NewInteraction().TranslateLocation.SetPath(config.TranslateLocation)
	lib.NewInteraction().ImportLocation.SetPath(config.ReimportLocation)
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
	runtime.EventsOn(ctx, "GameLocationChanged", func(data ...any) {
		fmt.Println("GameLocationChanged", data)
		lib.NewInteraction().GameLocation.SetPath(data[0].(string))
	})
	runtime.EventsOn(ctx, "ExtractLocationChanged", func(data ...any) {
		fmt.Println("ExtractLocationChanged", data)
		lib.NewInteraction().ExtractLocation.SetPath(data[0].(string))
	})

	runtime.EventsOn(ctx, "TranslateLocationChanged", func(data ...any) {
		fmt.Println("TranslateLocationChanged", data)
		lib.NewInteraction().TranslateLocation.SetPath(data[0].(string))
	})

	runtime.EventsOn(ctx, "ReimportLocationChanged", func(data ...any) {
		fmt.Println("ReimportLocationChanged", data)
		lib.NewInteraction().ImportLocation.SetPath(data[0].(string))
	})

	runtime.EventsEmit(ctx, "GameFilesLocation", lib.NewInteraction().GameLocation.GetPath())
	runtime.EventsEmit(ctx, "ExtractLocation", lib.NewInteraction().ExtractLocation.GetPath())
	runtime.EventsEmit(ctx, "TranslateLocation", lib.NewInteraction().TranslateLocation.GetPath())
	runtime.EventsEmit(ctx, "ReimportLocation", lib.NewInteraction().ImportLocation.GetPath())

	runtime.EventsOn(ctx, "GetGameLocationDirectory", func(data ...any) {
		fmt.Println("GetGameLocationDirectory", data)
		runtime.EventsEmit(ctx, "GameFilesLocation_value", lib.NewInteraction().GameLocation.GetPath())
	})

	testPath := "F:\\ffxWails\\FFX_Resources\\build\\bin\\data\\ffx_ps2\\ffx2\\master\\new_uspc\\menu\\macrodic.dcp"
	services.TestExtractFile(testPath, false, true)
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	config := AppConfig{
		GameFilesLocation: lib.NewInteraction().GameLocation.GetPath(),
		ExtractLocation:   lib.NewInteraction().ExtractLocation.GetPath(),
		TranslateLocation: lib.NewInteraction().TranslateLocation.GetPath(),
		ReimportLocation:  lib.NewInteraction().ImportLocation.GetPath(),
	}

	err := a.saveConfig(config)
	if err != nil {
		fmt.Println(err)
	}

	dialog, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.QuestionDialog,
		Title:   "Quit?",
		Message: "Are you sure you want to quit?",
	})

	if err != nil {
		return false
	}
	fmt.Println(dialog)
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

func (a *App) ReadFileAsString(fileInfo lib.FileInfo) string {
	content, err := os.ReadFile(fileInfo.ExtractLocation.TargetFile)
	if err != nil {
		return ""
	}
	fmt.Println(string(content))
	return string(content)
}

func (a *App) WriteTextFile(fileInfo lib.FileInfo, text string) {
	err := os.WriteFile(fileInfo.ExtractLocation.TargetFile, []byte(text), 0644)
	if err != nil {
		runtime.LogError(a.ctx, err.Error())
		runtime.EventsEmit(a.ctx, "ApplicationError", err.Error())
	}
}

func (a *App) SelectDirectory(title string) string {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: common.GetExecDir(),
	})

	if err != nil {
		return ""
	}
	return selection
}

func (a *App) GetExtractDirectory() string {
	return lib.NewInteraction().ExtractLocation.GetPath()
}

func (a *App) GetTranslateDirectory() string {
	return lib.NewInteraction().TranslateLocation.GetPath()
}

func (a *App) SetExtractDirectory(path string) {
	lib.NewInteraction().ExtractLocation.SetPath(path)
}

func (a *App) saveConfig(config AppConfig) error {
	filePath := a.appConfig.filePath
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (a *App) loadConfig(filePath string) (AppConfig, error) {
	var config AppConfig

	data, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	return config, err
}
