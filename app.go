package main

import (
	"context"
	"encoding/json"
	"ffxresources/backend/lib"
	"ffxresources/backend/services"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type AppConfig struct {
	filePath           string
	OriginalDirectory  string
	ExtractDirectory   string
	TranslateDirectory string
	GameLocation       lib.GameLocation
	ExtractLocation    lib.ExtractLocation
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
			filePath: lib.PathJoin(lib.GetExecDir(), "config.json"),
		},
	}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
	a.ExtractService.Ctx = ctx
	a.CollectionService.Ctx = ctx
	a.CompressService.Ctx = ctx

	config, err := a.LoadConfig(a.appConfig.filePath)
	if err != nil {
		fmt.Println(err)
	}
	a.appConfig.OriginalDirectory = config.OriginalDirectory
	a.appConfig.ExtractDirectory = config.ExtractDirectory
	a.appConfig.TranslateDirectory = config.TranslateDirectory

	lib.NewInteraction().GameLocation.SetPath(config.OriginalDirectory)
	lib.NewInteraction().WorkingLocation.ExtractedDirectory = config.ExtractDirectory
	lib.NewInteraction().WorkingLocation.TranslatedDirectory = config.TranslateDirectory

	runtime.LogInfo(ctx, "Original startup: "+lib.NewInteraction().GameLocation.GetPath())
	runtime.LogInfo(ctx, "Extracted startup: "+lib.NewInteraction().WorkingLocation.ExtractedDirectory)
	runtime.LogInfo(ctx, "Translated startup: "+lib.NewInteraction().WorkingLocation.TranslatedDirectory)
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
	runtime.LogInfo(ctx, "domReady")

	runtime.EventsOn(ctx, "GameLocationChanged", func(data ...any) {
		lib.NewInteraction().GameLocation.SetPath(data[0].(string))
	})
	runtime.EventsOn(ctx, "ExtractedDirectoryChanged", func(data ...any) {
		lib.NewInteraction().WorkingLocation.ExtractedDirectory = data[0].(string)
	})

	runtime.EventsOn(ctx, "TranslatedDirectoryChanged", func(data ...any) {
		lib.NewInteraction().WorkingLocation.TranslatedDirectory = data[0].(string)
	})

	runtime.EventsEmit(ctx, "GameDirectory", lib.NewInteraction().GameLocation.GetPath())
	runtime.EventsEmit(ctx, "ExtractedDirectory", lib.NewInteraction().WorkingLocation.ExtractedDirectory)
	runtime.EventsEmit(ctx, "TranslatedDirectory", lib.NewInteraction().WorkingLocation.TranslatedDirectory)

	runtime.EventsOn(ctx, "GetGameLocationDirectory", func(data ...any) {
		fmt.Println("GetGameLocationDirectory", data)
		runtime.EventsEmit(ctx, "GameDirectory_value", lib.NewInteraction().GameLocation.GetPath())
	})
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	config := AppConfig{
		OriginalDirectory:  lib.NewInteraction().GameLocation.GetPath(),
		ExtractDirectory:   lib.NewInteraction().WorkingLocation.ExtractedDirectory,
		TranslateDirectory: lib.NewInteraction().WorkingLocation.TranslatedDirectory,
	}

	err := a.SaveConfig(config)
	if err != nil {
		fmt.Println(err)
	}
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
		DefaultDirectory: lib.GetExecDir(),
	})

	if err != nil {
		return ""
	}
	return selection
}

func (a *App) GetExtractDirectory() string {
	return lib.GetExtractDirectory()
}

func (a *App) GetTranslateDirectory() string {
	return lib.GetTranslateDirectory()
}

func (a *App) SetExtractDirectory(path string) {
	lib.NewInteraction().WorkingLocation.ExtractedDirectory = path
}

func (a *App) SaveConfig(config AppConfig) error {
	filePath := a.appConfig.filePath
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (a *App) LoadConfig(filePath string) (AppConfig, error) {
	var config AppConfig

	// Lê o conteúdo do arquivo
	data, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	// Decodifica o conteúdo JSON para a struct
	err = json.Unmarshal(data, &config)
	return config, err
}
