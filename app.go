package main

import (
	"context"
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"ffxresources/backend/notifications"
	"ffxresources/backend/services"
	"log"
	"os"
	"path/filepath"

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
			filePath: filepath.Join(common.GetExecDir(), "config.json"),
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

	if err := lib.LoadFromJson(a.appConfig, a.appConfig.filePath); err != nil {
		notifications.NotifyError(err)
	}

	interactions.NewInteraction().GamePart.SetGamePartNumber(a.appConfig.GamePart)

	a.initializeLocationsDirectory()
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
	EventsOnLocations(ctx)

	EmitLocationsEvents(ctx)

	EventsOnSaveConfig(ctx, a.appConfig.filePath)

	testPath := "F:\\ffxWails\\FFX_Resources\\build\\bin\\data\\ffx-2_data\\gamedata\\ps3data\\lockit\\ffx2_loc_kit_ps3_us.bin"
	services.TestExtractFile(testPath, false, false)

	testPath = `F:\ffxWails\FFX_Resources\build\bin\data\ffx_ps2\ffx2\master\new_uspc\menu\macrodic.dcp`
	services.TestExtractFile(testPath, true, false)

	testPath = `F:\ffxWails\FFX_Resources\build\bin\data\ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_235\bika07_235.bin`
	services.TestExtractFile(testPath, false, false)

	testPath = `build/bin/data/ffx_ps2/ffx2/master/new_uspc/event/obj_ps3/dn/dnfr0100/dnfr0100.bin`
	services.TestExtractFile(testPath, false, false)

	testPath = `build\bin\data\ffx_ps2\ffx2\master\new_uspc\lastmiss\kernel\lm_accesary.bin`
	services.TestExtractFile(testPath, false, false)

	testPath = `build\bin\data\ffx_ps2\ffx2\master\new_uspc\lastmiss\kernel`
	services.TestExtractDir(testPath, false, false)

}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	config := AppConfig{
		GameFilesLocation: interactions.NewInteraction().GameLocation.GetTargetDirectory(),
		GamePart:          interactions.NewInteraction().GamePart.GetGamePartNumber(),
		ExtractLocation:   interactions.NewInteraction().ExtractLocation.GetTargetDirectory(),
		TranslateLocation: interactions.NewInteraction().TranslateLocation.GetTargetDirectory(),
		ReimportLocation:  interactions.NewInteraction().ImportLocation.GetTargetDirectory(),
	}

	if err := lib.SaveToJSONFile(config, a.appConfig.filePath); err != nil {
		notifications.NotifyError(err)
	}

	if _, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.QuestionDialog,
		Title:   "Quit?",
		Message: "Are you sure you want to quit?",
	}); err != nil {
		return false
	}

	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

func (a *App) ReadFileAsString(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	//fmt.Println(string(content))
	return string(content)
}

func (a *App) WriteTextFile(file string, text string) {
	err := os.WriteFile(file, []byte(text), 0644)
	if err != nil {
		notifications.NotifyError(err)

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

func (a *App) initializeLocationsDirectory() {
	interactions.NewInteraction().GameLocation.SetTargetDirectory(a.appConfig.GameFilesLocation)
	interactions.NewInteraction().ExtractLocation.SetTargetDirectory(a.appConfig.ExtractLocation)
	interactions.NewInteraction().TranslateLocation.SetTargetDirectory(a.appConfig.TranslateLocation)
	interactions.NewInteraction().ImportLocation.SetTargetDirectory(a.appConfig.ReimportLocation)
}
