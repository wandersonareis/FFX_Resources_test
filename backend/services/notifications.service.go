package services

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type (
	INotificationService interface {
		Notify(notification Severity, message string)
		NotifyError(err error)
		NotifyInfo(message string)
		NotifyWarn(message string)
		NotifySuccess(message string)
	}

	// NotificationService publica avisos no frontend pelo bus de eventos do
	// Wails v3 — não guarda mais um context.Context, que no v2 existia só
	// para runtime.EventsEmit.
	NotificationService struct{}

	Notification struct {
		Message  string `json:"message"`
		Severity string `json:"severity"`
	}

	Severity int
)

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

func NewEventNotifier() INotificationService {
	return &NotificationService{}
}

func (e *NotificationService) Notify(notification Severity, message string) {
	captalize := func(s string) string {
		if len(s) == 0 {
			return s
		}
		return cases.Upper(language.BrazilianPortuguese).String(s[:1]) + s[1:]
	}

	notify := Notification{
		Message:  captalize(message),
		Severity: notification.String(),
	}

	if !HasEventBus() {
		return
	}

	EmitEvent("Notify", notify)
}

func (e *NotificationService) NotifyError(err error) {
	e.Notify(SeverityError, err.Error())
}

func (e *NotificationService) NotifyInfo(message string) {
	e.Notify(SeverityInfo, message)
}

func (e *NotificationService) NotifyWarn(message string) {
	e.Notify(SeverityWarn, message)
}

func (e *NotificationService) NotifySuccess(message string) {
	e.Notify(SeveritySuccess, message)
}
