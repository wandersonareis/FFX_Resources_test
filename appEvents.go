package main

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/services"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Este arquivo era todo parametrizado por um context.Context que existia só
// para runtime.EventsOn/EventsEmit. No Wails v3 os eventos vivem no
// EventManager da aplicação, então o parâmetro threaded passou a ser o
// *application.EventManager — mesma forma, tipo novo.

func EventsOnStartup(bus *application.EventManager) {
	emitLocationsEvents(bus)
	eventsOnLocations(bus)
}

func emitLocationsEvents(bus *application.EventManager) {
	emitGameVersion(bus)
	emitGameLocation(bus)
	emitGameExeLocation(bus)
	emitExtractLocation(bus)
	emitTranslateLocation(bus)
	emitimportLocation(bus)
}

func eventsOnLocations(bus *application.EventManager) {
	eventOnSetGameVersion(bus)
	eventOnSetGameLocation(bus)
	eventOnSetGameExeLocation(bus)
	eventOnSetExtractLocation(bus)
	eventOnSetTranslateLocation(bus)
	eventOnSetImportLocation(bus)
}

func EventsOnSaveConfig(bus *application.EventManager) {
	bus.On("SaveConfig", func(e *application.CustomEvent) {
		fmt.Println("SaveConfig", e.Data)

		interactions.NewInteractionService().FFXAppConfig().ToJson()
	})
}

func emitGameVersion(bus *application.EventManager) {
	gameVersion := interactions.CurrentGameVersion()
	bus.Emit("GameVersion", gameVersion)
}

func emitGameLocation(bus *application.EventManager) {
	gameLocation := interactions.NewInteractionService().GameLocation.GetTargetDirectory()

	if gameLocation == "" {
		interactions.NewInteractionService().GameLocation.ProvideTargetDirectory()

		emitGameLocation(bus)
		return
	}

	bus.Emit("GameFilesLocation", gameLocation)
}

// emitGameExeLocation devolve o executável do jogo escolhido ("") — é dele
// que o navegador de .vbf deriva a pasta de busca dos containers.
func emitGameExeLocation(bus *application.EventManager) {
	bus.Emit("GameExeLocation",
		interactions.NewInteractionService().FFXAppConfig().GetGameExeLocation())
}

func emitExtractLocation(bus *application.EventManager) {
	extractLocation := interactions.NewInteractionService().ExtractLocation.GetTargetDirectory()

	if extractLocation == "" {
		interactions.NewInteractionService().ExtractLocation.ProvideTargetDirectory()

		emitExtractLocation(bus)
		return
	}
	bus.Emit("ExtractLocation", extractLocation)
}

func emitTranslateLocation(bus *application.EventManager) {
	translateLocation := interactions.NewInteractionService().TranslateLocation.GetTargetDirectory()

	if translateLocation == "" {
		interactions.NewInteractionService().TranslateLocation.ProvideTargetDirectory()

		emitTranslateLocation(bus)
		return
	}
	bus.Emit("TranslateLocation", translateLocation)
}

func emitimportLocation(bus *application.EventManager) {
	importLocation := interactions.NewInteractionService().ImportLocation.GetTargetDirectory()

	if importLocation == "" {
		interactions.NewInteractionService().ImportLocation.ProvideTargetDirectory()

		emitimportLocation(bus)
		return
	}
	bus.Emit("ReimportLocation", importLocation)
}

func eventOnSetGameVersion(bus *application.EventManager) {
	updateGameVersion := func(version common.GameVersion) {
		interactions.NewInteractionService().FFXAppConfig().SetGameVersion(version)
	}

	bus.On("GameVersionChanged", func(e *application.CustomEvent) {
		fmt.Println("GameVersionChanged", e.Data)

		gameVer, err := common.ParseGameVersionStrict(e.Data.(string))
		if err != nil {
			fmt.Println("ParseGameVersion error:", err)
			return
		}
		updateGameVersion(gameVer)

		emitGameVersion(bus)
	})
}

func eventOnSetGameLocation(bus *application.EventManager) {
	bus.On("GameLocationChanged", func(e *application.CustomEvent) {
		fmt.Println("GameLocationChanged", e.Data)

		service := interactions.NewInteractionService()
		oldGameDir := service.GameLocation.GetTargetDirectory()
		oldDefaultTranslate := common.DefaultTranslatedDir(oldGameDir)

		service.GameLocation.SetTargetDirectory(e.Data.(string))

		// A árvore de gamefiles mudou: tudo que foi derivado dela (DTOs
		// deduplicados, original pristine de data/ e os stores de formato)
		// pertence ao diretório anterior — descarta, senão a view serviria
		// texto de outro diretório.
		services.InvalidateViewCaches()

		// Se o translated ainda apontava para o default derivado do
		// gamefiles anterior, acompanha o novo gamefiles.
		if service.TranslateLocation.GetTargetDirectory() == oldDefaultTranslate {
			newDefaultTranslate := common.DefaultTranslatedDir(service.GameLocation.GetTargetDirectory())
			_ = service.TranslateLocation.SetTargetDirectory(newDefaultTranslate)
			emitTranslateLocation(bus)
		}

		emitGameLocation(bus)
		// Aplica imediatamente: config já foi persistido por
		// SetTargetDirectory; avisa o frontend para recarregar a árvore.
		bus.Emit("Refresh_Tree")
	})
}

// eventOnSetGameExeLocation grava o executável do jogo (caminho de ARQUIVO,
// fora do map Locations) e recarrega a árvore VBF.
func eventOnSetGameExeLocation(bus *application.EventManager) {
	bus.On("GameExeLocationChanged", func(e *application.CustomEvent) {
		fmt.Println("GameExeLocationChanged", e.Data)

		interactions.NewInteractionService().FFXAppConfig().
			SetGameExeLocation(e.Data.(string))

		// A pasta de busca dos .vbf mudou: os containers abertos seguram
		// handle de leitura do caminho anterior — fecha; a próxima listagem
		// reabre os da nova pasta.
		services.CloseVbfArchives()
		emitGameExeLocation(bus)
		bus.Emit("Refresh_Tree")
	})
}

func eventOnSetExtractLocation(bus *application.EventManager) {
	bus.On("ExtractLocationChanged", func(e *application.CustomEvent) {
		fmt.Println("ExtractLocationChanged", e.Data)

		interactions.NewInteractionService().ExtractLocation.SetTargetDirectory(e.Data.(string))

		emitExtractLocation(bus)
	})
}

func eventOnSetTranslateLocation(bus *application.EventManager) {
	bus.On("TranslateLocationChanged", func(e *application.CustomEvent) {
		fmt.Println("TranslateLocationChanged", e.Data)

		interactions.NewInteractionService().TranslateLocation.SetTargetDirectory(e.Data.(string))

		emitTranslateLocation(bus)
		// Fonte da tradução do reimport (mods/translated); recarrega
		// metadados da árvore para refletir disponibilidade.
		bus.Emit("Refresh_Tree")
	})
}

func eventOnSetImportLocation(bus *application.EventManager) {
	bus.On("ReimportLocationChanged", func(e *application.CustomEvent) {
		fmt.Println("ReimportLocationChanged", e.Data)

		interactions.NewInteractionService().ImportLocation.SetTargetDirectory(e.Data.(string))

		emitimportLocation(bus)
	})
}
