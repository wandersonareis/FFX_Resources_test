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
		
		interactions.NewInteraction().FFXAppConfig().ToJson()
	})
}

func emitGameVersion(ctx context.Context) {
	gameVersion := interactions.NewInteraction().FFXGameVersion().GetGameVersionNumber()
	runtime.EventsEmit(ctx, "GameVersion", gameVersion)
}

func emitGameLocation(ctx context.Context) {
	gameLocation := interactions.NewInteraction().GameLocation.GetTargetDirectory()
	runtime.EventsEmit(ctx, "GameFilesLocation", gameLocation)
}

func emitExtractLocation(ctx context.Context) {
	extractLocation := interactions.NewInteraction().ExtractLocation.GetTargetDirectory()

	if extractLocation == "" {
		interactions.NewInteraction().ExtractLocation.ProvideTargetDirectory()

		emitExtractLocation(ctx)
		return
	}
	runtime.EventsEmit(ctx, "ExtractLocation", extractLocation)
}

func emitTranslateLocation(ctx context.Context) {
	translateLocation := interactions.NewInteraction().TranslateLocation.GetTargetDirectory()

	if translateLocation == "" {
		interactions.NewInteraction().TranslateLocation.ProvideTargetDirectory()

		emitTranslateLocation(ctx)
		return
	}
	runtime.EventsEmit(ctx, "TranslateLocation", translateLocation)
}

func emitimportLocation(ctx context.Context) {
	importLocation := interactions.NewInteraction().ImportLocation.GetTargetDirectory()

	if importLocation == "" {
		interactions.NewInteraction().ImportLocation.ProvideTargetDirectory()

		emitimportLocation(ctx)
		return
	}
	runtime.EventsEmit(ctx, "ReimportLocation", importLocation)
}

func eventOnSetGameVersion(ctx context.Context) {
	runtime.EventsOn(ctx, "GameVersionChanged", func(data ...any) {
		fmt.Println("GameVersionChanged", data[0])

		interactions.NewInteraction().FFXGameVersion().SetGameVersionNumber(int(data[0].(float64)))
		interactions.NewInteraction().FFXAppConfig().ToJson()

		emitGameVersion(ctx)
	})
}

func eventOnSetGameLocation(ctx context.Context) {
	runtime.EventsOn(ctx, "GameLocationChanged", func(data ...any) {
		fmt.Println("GameLocationChanged", data)

		interactions.NewInteraction().GameLocation.SetTargetDirectory(data[0].(string))

		emitGameLocation(ctx)
	})
}

func eventOnSetExtractLocation(ctx context.Context) {
	runtime.EventsOn(ctx, "ExtractLocationChanged", func(data ...any) {
		fmt.Println("ExtractLocationChanged", data)

		interactions.NewInteraction().ExtractLocation.SetTargetDirectory(data[0].(string))

		emitExtractLocation(ctx)
	})
}

func eventOnSetTranslateLocation(ctx context.Context) {
	runtime.EventsOn(ctx, "TranslateLocationChanged", func(data ...any) {
		fmt.Println("TranslateLocationChanged", data)

		interactions.NewInteraction().TranslateLocation.SetTargetDirectory(data[0].(string))

		emitTranslateLocation(ctx)
	})
}

func eventOnSetImportLocation(ctx context.Context) {
	runtime.EventsOn(ctx, "ReimportLocationChanged", func(data ...any) {
		fmt.Println("ReimportLocationChanged", data)

		interactions.NewInteraction().ImportLocation.SetTargetDirectory(data[0].(string))

		emitimportLocation(ctx)
	})
}
