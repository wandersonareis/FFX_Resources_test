package main

import (
	"context"
	"ffxresources/backend/common"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/loggingService"
	"ffxresources/backend/services"
	"ffxresources/backend/spira"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type AppConfig struct {
	GameFilesLocation string `json:"gameFilesLocation"`
	GamePart          int    `json:"gamePart"`
	ExtractLocation   string `json:"extractLocation"`
	TranslateLocation string `json:"translateLocation"`
	ReimportLocation  string `json:"reimportLocation"`
}

// App struct
type App struct {
	noticationService services.INotificationService

	CollectionService *services.CollectionService
	ExtractService    *services.ExtractService
	CompressService   *services.CompressService
}

// NewApp creates a new App application struct
func NewApp() *App {
	notifier := services.NewEventNotifier(context.Background())
	progress := services.NewProgressService(context.Background())
	return &App{
		CollectionService: services.NewCollectionService(notifier),
		ExtractService:    services.NewExtractService(notifier, progress),
		CompressService:   services.NewCompressService(notifier, progress),
	}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)

			l := loggingService.Get()
			l.Fatal().Caller(2).Err(err.(error)).Msg("panic occurred")
		}
	}()

	// Initialize services
	a.initServices(ctx)

	interactions.NewInteractionWithCtx(ctx)
	interactions.NewInteractionWithTextFormatter(formatters.NewTxtFormatter())
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)

			l := loggingService.Get()
			l.Fatal().Caller(2).Err(err.(error)).Msg("panic occurred")
		}
	}()

	EventsOnStartup(ctx)

	EventsOnSaveConfig(ctx)
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	interactions.NewInteractionService().FFXAppConfig().ToJson()

	answer, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.QuestionDialog,
		Title:   "Quit?",
		Message: "Are you sure you want to quit?",
	})
	if err != nil {
		return false
	}

	fmt.Println("Answer:", answer)
	return answer != "Yes"
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

func (a *App) initServices(ctx context.Context) {
	notification := services.NewEventNotifier(ctx)
	progress := services.NewProgressService(ctx)

	a.noticationService = notification

	// Initialize services
	a.CollectionService = services.NewCollectionService(notification)
	a.ExtractService = services.NewExtractService(notification, progress)
	a.CompressService = services.NewCompressService(notification, progress)
}

func (a *App) BuildTree() []spira.TreeNode {
	var tree []spira.TreeNode

	root := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
	if root == "" {
		return nil
	}

	dirs, err := os.ReadDir(root)
	if err != nil {
		a.noticationService.NotifyError(fmt.Errorf("failed to read root directory: %s", err))
		return tree
	}

	for _, d := range dirs {
		if d.IsDir() {
			buildTreeNode := a.CollectionService.BuildTree(filepath.Join(root, d.Name()))
			tree = append(tree, buildTreeNode...)
		}
	}

	if err := common.CheckArgumentNil(tree, "BuildTree"); err != nil {
		a.noticationService.NotifyError(fmt.Errorf("failed to build files tree"))
		return nil
	}

	return tree
}

func (a *App) Extract(path string) {
	if err := common.CheckArgumentNil(path, "path"); err != nil {
		a.noticationService.NotifyError(err)
		return
	}

	if err := a.ExtractService.Extract(path); err != nil {
		a.noticationService.NotifyError(err)
	}
}

func (a *App) Compress(path string) {
	if err := common.CheckArgumentNil(path, "path"); err != nil {
		a.noticationService.NotifyError(err)
		return
	}

	if err := a.CompressService.Compress(path); err != nil {
		a.noticationService.NotifyError(err)
	}
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
		a.noticationService.NotifyError(err)
	}
}

func (a *App) SelectDirectory(title string) string {
	selection, err := runtime.OpenDirectoryDialog(interactions.NewInteractionService().Ctx, runtime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: common.GetExecDir(),
	})

	if err != nil {
		return ""
	}
	return selection
}
