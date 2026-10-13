package main

import (
	"embed"
	"fmt"
	"log"
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/out
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	// Create an instance of the app structure
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				fmt.Println("Recovered from panic:", e.Error())
			} else {
				fmt.Printf("Recovered from panic: %v\n", r)
			}
		}
	}()

	appInstance := NewApp()

	// v2: wails.Run(&options.App{...}) combinava opções, janela e execução.
	// v3 separa em application.New (app + serviços), Window.NewWithOptions
	// (janela) e app.Run (event loop).
	app := application.New(application.Options{
		Name: "FINAL FANTASY X and X-2 HD Remaster Resources Editor",
		Icon: icon,
		// App (sessão/ciclo de vida) + serviços por domínio.
		Services: appServices(appInstance),
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		// v2: OnBeforeClose. shouldQuit devolve true para PERMITIR a saída,
		// o inverso do `prevent` do v2.
		ShouldQuit: appInstance.shouldQuit,
		// v2: LogLevel logger.TRACE — o v3 não tem TRACE; Debug é o nível
		// mais baixo disponível.
		LogLevel: slog.LevelDebug,
		Windows: application.WindowsOptions{
			// v2: WebviewUserDataPath: "" (default %APPDATA%\<binário>).
			WebviewUserDataPath: "",
		},
	})

	// v2: Title/Width/Height/MinWidth/MinHeight/BackgroundColour/WindowStartState
	// moravam em options.App; no v3 viraram opções da janela. O menu e o menu
	// de contexto padrão continuam desligados/ligados por default
	// (EnableDefaultContextMenu: true vira DefaultContextMenuDisabled: false).
	_ = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "FINAL FANTASY X and X-2 HD Remaster Resources Editor",
		Width:            1280,
		Height:           800,
		MinWidth:         1024,
		MinHeight:        600,
		StartState:       application.WindowStateNormal,
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL:              "/",
		// v2: mac.Options{TitleBar, Appearance} — no v3 essas opções passaram
		// para a janela.
		Mac: application.MacWindow{
			TitleBar: application.MacTitleBar{
				AppearsTransparent:   true,
				HideToolbarSeparator: true,
			},
			Appearance: application.NSAppearanceNameDarkAqua,
		},
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
