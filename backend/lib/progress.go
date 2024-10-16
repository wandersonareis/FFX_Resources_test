package lib

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Progress struct {
	Total      int `json:"total"`
	Processed  int `json:"processed"`
	Percentage int `json:"percentage"`
}

func ShowProgressBar(ctx context.Context) {
	runtime.EventsEmit(ctx, "ShowProgress", true)
}

func HideProgressBar(ctx context.Context) {
	runtime.EventsEmit(ctx, "ShowProgress", false)
}

func SendProgress(ctx context.Context, progress Progress) {
	runtime.EventsEmit(ctx, "Progress", progress)
}
