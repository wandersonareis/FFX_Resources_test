package notifications

import (
	"ffxresources/backend/interactions"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	captalize := func(s string) string {
		if len(s) == 0 {
			return s
		}

		return cases.Upper(language.BrazilianPortuguese).String(s[:1]) + s[1:]
	}

	context := interactions.NewInteractionService().Ctx
	notify := Notification{
		Message:  captalize(message),
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
