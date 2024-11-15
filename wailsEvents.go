package main

import (
	"context"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func EmitLocationsEvents(ctx context.Context) {
	runtime.EventsEmit(ctx, "GameFilesLocation", interactions.NewInteraction().GameLocation.GetTargetDirectory())
	runtime.EventsEmit(ctx, "ExtractLocation", interactions.NewInteraction().ExtractLocation.GetTargetDirectory())
	runtime.EventsEmit(ctx, "TranslateLocation", interactions.NewInteraction().TranslateLocation.GetTargetDirectory())
	runtime.EventsEmit(ctx, "ReimportLocation", interactions.NewInteraction().ImportLocation.GetTargetDirectory())
}

func EventsOnLocations(ctx context.Context) {
	runtime.EventsOn(ctx, "GameLocationChanged", func(data ...any) {
		fmt.Println("GameLocationChanged", data)
		interactions.NewInteraction().GameLocation.SetTargetDirectory(data[0].(string))
		runtime.EventsEmit(ctx, "GameFilesLocation", interactions.NewInteraction().GameLocation.GetTargetDirectory())
	})
	runtime.EventsOn(ctx, "ExtractLocationChanged", func(data ...any) {
		fmt.Println("ExtractLocationChanged", data)
		interactions.NewInteraction().ExtractLocation.SetTargetDirectory(data[0].(string))
		runtime.EventsEmit(ctx, "ExtractLocation", interactions.NewInteraction().ExtractLocation.GetTargetDirectory())
	})

	runtime.EventsOn(ctx, "TranslateLocationChanged", func(data ...any) {
		fmt.Println("TranslateLocationChanged", data)
		interactions.NewInteraction().TranslateLocation.SetTargetDirectory(data[0].(string))
		runtime.EventsEmit(ctx, "TranslateLocation", interactions.NewInteraction().TranslateLocation.GetTargetDirectory())
	})

	runtime.EventsOn(ctx, "ReimportLocationChanged", func(data ...any) {
		fmt.Println("ReimportLocationChanged", data)
		interactions.NewInteraction().ImportLocation.SetTargetDirectory(data[0].(string))
		runtime.EventsEmit(ctx, "ReimportLocation", interactions.NewInteraction().ImportLocation.GetTargetDirectory())
	})
}

func EventsOnSaveConfig(ctx context.Context, filePath string) {
	runtime.EventsOn(ctx, "SaveConfig", func(data ...any) {
		config := AppConfig{
			GameFilesLocation: interactions.NewInteraction().GameLocation.GetTargetDirectory(),
			GamePart:          interactions.NewInteraction().GamePart.GetGamePartNumber(),
			ExtractLocation:   interactions.NewInteraction().ExtractLocation.GetTargetDirectory(),
			TranslateLocation: interactions.NewInteraction().TranslateLocation.GetTargetDirectory(),
			ReimportLocation:  interactions.NewInteraction().ImportLocation.GetTargetDirectory(),
		}

		lib.SaveToJSONFile(config, filePath)
	})
}
