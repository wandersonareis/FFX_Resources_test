package main

import (
	"context"
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"ffxresources/backend/services"
	"fmt"
	"log"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type AppConfig struct {
	filePath          string
	GameFilesLocation string `json:"gameFilesLocation"`
	GamePart          int    `json:"gamePart"`
	ExtractLocation   string `json:"extractLocation"`
	TranslateLocation string `json:"translateLocation"`
	ReimportLocation  string `json:"reimportLocation"`
}

// App struct
type App struct {
	//ctx               context.Context
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
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()
	interactions.NewInteractionWithCtx(ctx)

	err := a.loadConfig(a.appConfig)
	if err != nil {
		lib.LogSeverity(lib.SeverityError, err.Error())
	}

	interactions.NewInteraction().GamePart.SetGamePartNumber(a.appConfig.GamePart)

	interactions.NewInteraction().GameLocation.SetPath(a.appConfig.GameFilesLocation)
	interactions.NewInteraction().ExtractLocation.SetPath(a.appConfig.ExtractLocation)
	interactions.NewInteraction().TranslateLocation.SetPath(a.appConfig.TranslateLocation)
	interactions.NewInteraction().ImportLocation.SetPath(a.appConfig.ReimportLocation)
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
	runtime.EventsOn(ctx, "GameLocationChanged", func(data ...any) {
		fmt.Println("GameLocationChanged", data)
		interactions.NewInteraction().GameLocation.SetPath(data[0].(string))
	})
	runtime.EventsOn(ctx, "ExtractLocationChanged", func(data ...any) {
		fmt.Println("ExtractLocationChanged", data)
		interactions.NewInteraction().ExtractLocation.SetPath(data[0].(string))
	})

	runtime.EventsOn(ctx, "TranslateLocationChanged", func(data ...any) {
		fmt.Println("TranslateLocationChanged", data)
		interactions.NewInteraction().TranslateLocation.SetPath(data[0].(string))
	})

	runtime.EventsOn(ctx, "ReimportLocationChanged", func(data ...any) {
		fmt.Println("ReimportLocationChanged", data)
		interactions.NewInteraction().ImportLocation.SetPath(data[0].(string))
	})

	runtime.EventsEmit(ctx, "GameFilesLocation", interactions.NewInteraction().GameLocation.GetPath())
	runtime.EventsEmit(ctx, "ExtractLocation", interactions.NewInteraction().ExtractLocation.GetPath())
	runtime.EventsEmit(ctx, "TranslateLocation", interactions.NewInteraction().TranslateLocation.GetPath())
	runtime.EventsEmit(ctx, "ReimportLocation", interactions.NewInteraction().ImportLocation.GetPath())

	runtime.EventsOn(ctx, "GetGameLocationDirectory", func(data ...any) {
		fmt.Println("GetGameLocationDirectory", data)
		runtime.EventsEmit(ctx, "GameFilesLocation_value", interactions.NewInteraction().GameLocation.GetPath())
	})

	testPath := "F:\\ffxWails\\FFX_Resources\\build\\bin\\data\\ffx-2_data\\gamedata\\ps3data\\lockit\\ffx2_loc_kit_ps3_us.bin"
	services.TestExtractFile(testPath, false, true)
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	config := AppConfig{
		GameFilesLocation: interactions.NewInteraction().GameLocation.GetPath(),
		GamePart:          interactions.NewInteraction().GamePart.GetGamePartNumber(),
		ExtractLocation:   interactions.NewInteraction().ExtractLocation.GetPath(),
		TranslateLocation: interactions.NewInteraction().TranslateLocation.GetPath(),
		ReimportLocation:  interactions.NewInteraction().ImportLocation.GetPath(),
	}

	err := a.saveConfig(config)
	if err != nil {
		fmt.Println(err)
	}

	_, err = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.QuestionDialog,
		Title:   "Quit?",
		Message: "Are you sure you want to quit?",
	})

	if err != nil {
		return false
	}

	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

func (a *App) ReadFileAsString(dataInfo interactions.GameDataInfo) string {
	content, err := os.ReadFile(dataInfo.ExtractLocation.TargetFile)
	if err != nil {
		return ""
	}
	fmt.Println(string(content))
	return string(content)
}

func (a *App) WriteTextFile(dataInfo interactions.GameDataInfo, text string) {
	err := os.WriteFile(dataInfo.ExtractLocation.TargetFile, []byte(text), 0644)
	if err != nil {
		lib.LogSeverity(lib.SeverityError, err.Error())

		runtime.EventsEmit(interactions.NewInteraction().Ctx, "Notify", err.Error())
	}
}

func (a *App) SelectDirectory(title string) string {
	selection, err := runtime.OpenDirectoryDialog(interactions.NewInteraction().Ctx, runtime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: common.GetExecDir(),
	})

	if err != nil {
		return ""
	}
	return selection
}

func (a *App) GetExtractDirectory() string {
	return interactions.NewInteraction().ExtractLocation.GetPath()
}

func (a *App) GetTranslateDirectory() string {
	return interactions.NewInteraction().TranslateLocation.GetPath()
}

func (a *App) SetExtractDirectory(path string) {
	interactions.NewInteraction().ExtractLocation.SetPath(path)
}

func (a *App) saveConfig(config AppConfig) error {
	filePath := a.appConfig.filePath
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (a *App) loadConfig(config *AppConfig) error {
	filePath := a.appConfig.filePath

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &config)
	return err
}
