package main

import (
	"context"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func EventsOnStartup(ctx context.Context) {
	go notifications.PanicRecoverLogger(logger.Get().With().Str("module", "events").Logger())

	emitLocationsEvents(ctx)
	eventsOnLocations(ctx)
}

func emitLocationsEvents(ctx context.Context) {
	emitGameVersion(ctx)
	emitGameLocation(ctx)
	emitExtractLocation(ctx)
	emitTranslateLocation(ctx)
	emitimportLocation(ctx)
}

func eventsOnLocations(ctx context.Context) {
	eventOnSetGameVersion(ctx)
	eventOnSetGameLocation(ctx)
	eventOnSetExtractLocation(ctx)
	eventOnSetTranslateLocation(ctx)
	eventOnSetImportLocation(ctx)
}

func EventsOnSaveConfig(ctx context.Context) {
	runtime.EventsOn(ctx, "SaveConfig", func(data ...any) {
		fmt.Println("SaveConfig", data)
		
		interactions.NewInteractionService().FFXAppConfig().ToJson()
	})
}

func emitGameVersion(ctx context.Context) {
	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber()
	runtime.EventsEmit(ctx, "GameVersion", gameVersion)
}

func emitGameLocation(ctx context.Context) {
	gameLocation := interactions.NewInteractionService().GameLocation.GetTargetDirectory()

	if gameLocation == "" {
		interactions.NewInteractionService().GameLocation.ProvideTargetDirectory()

		emitGameLocation(ctx)
		return
	}
	
	runtime.EventsEmit(ctx, "GameFilesLocation", gameLocation)
}

func emitExtractLocation(ctx context.Context) {
	extractLocation := interactions.NewInteractionService().ExtractLocation.GetTargetDirectory()

	if extractLocation == "" {
		interactions.NewInteractionService().ExtractLocation.ProvideTargetDirectory()

		emitExtractLocation(ctx)
		return
	}
	runtime.EventsEmit(ctx, "ExtractLocation", extractLocation)
}

func emitTranslateLocation(ctx context.Context) {
	translateLocation := interactions.NewInteractionService().TranslateLocation.GetTargetDirectory()

	if translateLocation == "" {
		interactions.NewInteractionService().TranslateLocation.ProvideTargetDirectory()

		emitTranslateLocation(ctx)
		return
	}
	runtime.EventsEmit(ctx, "TranslateLocation", translateLocation)
}

func emitimportLocation(ctx context.Context) {
	importLocation := interactions.NewInteractionService().ImportLocation.GetTargetDirectory()

	if importLocation == "" {
		interactions.NewInteractionService().ImportLocation.ProvideTargetDirectory()

		emitimportLocation(ctx)
		return
	}
	runtime.EventsEmit(ctx, "ReimportLocation", importLocation)
}

func eventOnSetGameVersion(ctx context.Context) {
	updateGameVersionNumber := func(version int) {
		interactions.NewInteractionService().FFXGameVersion().SetGameVersionNumber(version)
		
		interactions.NewInteractionService().FFXAppConfig().ToJson()
	}

	runtime.EventsOn(ctx, "GameVersionChanged", func(data ...any) {
		fmt.Println("GameVersionChanged", data[0])
		
		updateGameVersionNumber(int(data[0].(float64)))

		emitGameVersion(ctx)
	})
}

func eventOnSetGameLocation(ctx context.Context) {
	runtime.EventsOn(ctx, "GameLocationChanged", func(data ...any) {
		fmt.Println("GameLocationChanged", data)

		interactions.NewInteractionService().GameLocation.SetTargetDirectory(data[0].(string))

		emitGameLocation(ctx)
	})
}

func eventOnSetExtractLocation(ctx context.Context) {
	runtime.EventsOn(ctx, "ExtractLocationChanged", func(data ...any) {
		fmt.Println("ExtractLocationChanged", data)

		interactions.NewInteractionService().ExtractLocation.SetTargetDirectory(data[0].(string))

		emitExtractLocation(ctx)
	})
}

func eventOnSetTranslateLocation(ctx context.Context) {
	runtime.EventsOn(ctx, "TranslateLocationChanged", func(data ...any) {
		fmt.Println("TranslateLocationChanged", data)

		interactions.NewInteractionService().TranslateLocation.SetTargetDirectory(data[0].(string))

		emitTranslateLocation(ctx)
	})
}

func eventOnSetImportLocation(ctx context.Context) {
	runtime.EventsOn(ctx, "ReimportLocationChanged", func(data ...any) {
		fmt.Println("ReimportLocationChanged", data)

		interactions.NewInteractionService().ImportLocation.SetTargetDirectory(data[0].(string))

		emitimportLocation(ctx)
	})
}
