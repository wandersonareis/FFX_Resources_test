package lib

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func EmitError(ctx context.Context, err error) {
	runtime.EventsEmit(ctx, "ApplicationError", err.Error())
	runtime.LogDebug(ctx, err.Error())
}

func InvalidFileType(ctx context.Context, fileName string) {
	runtime.EventsEmit(ctx, "ApplicationError", "Invalid file type")
	EmitError(ctx, fmt.Errorf("invalid file type: %s", fileName))
	runtime.LogDebug(ctx, "Invalid file type: "+fileName)
}