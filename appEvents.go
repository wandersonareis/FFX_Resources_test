package main

import (
	"ffxresources/backend/interactions"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Este arquivo emite o estado inicial de localizações quando o frontend fica
// pronto (domReady). Os handlers FE→BE (GameVersionChanged, etc.) foram
// movidos para ConfigService como métodos bindados.

func EventsOnStartup(bus *application.EventManager) {
	emitLocationsEvents(bus)
}

func emitLocationsEvents(bus *application.EventManager) {
	emitGameVersion(bus)
	emitGameLocation(bus)
	emitGameExeLocation(bus)
	emitTranslateLocation(bus)
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
func emitTranslateLocation(bus *application.EventManager) {
	translateLocation := interactions.NewInteractionService().TranslateLocation.GetTargetDirectory()

	if translateLocation == "" {
		interactions.NewInteractionService().TranslateLocation.ProvideTargetDirectory()

		emitTranslateLocation(bus)
		return
	}
	bus.Emit("TranslateLocation", translateLocation)
}
