package lib

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Notification struct {
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type Severity int

const (
	SeveritySuccess Severity = iota
	SeverityInfo
	SeverityWarn
	SeverityError
)


func (nt Severity) String() string {
	switch nt {
	case SeveritySuccess:
		return "success"
	case SeverityInfo:
		return "info"
	case SeverityWarn:
		return "warn"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

func EmitError(ctx context.Context, err error) {
	runtime.EventsEmit(ctx, "ApplicationError", err.Error())
	runtime.LogDebug(ctx, err.Error())
}

func Notify(ctx context.Context, notification Severity, message string) {
	notify := Notification{
		Message:  message,
		Severity: notification.String(),
	}

	runtime.EventsEmit(ctx, "Notify", notify)
	runtime.LogPrint(ctx, message)
	fmt.Println(message)
}
