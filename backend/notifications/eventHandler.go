package notifications

import (
	"ffxresources/backend/interactions"

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

func Notify(notification Severity, message string) {
	context := interactions.NewInteraction().Ctx
	notify := Notification{
		Message:  message,
		Severity: notification.String(),
	}

	runtime.EventsEmit(context, "Notify", notify)
}

func NotifyError(err error) {
	Notify(SeverityError, err.Error())
}

func NotifyInfo(message string) {
	Notify(SeverityInfo, message)
}

func NotifyWarn(message string) {
	Notify(SeverityWarn, message)
}

func NotifySuccess(message string) {
	Notify(SeveritySuccess, message)
}
