package services

import (
	"context"
	"reflect"

	"github.com/wailsapp/wails/v2/pkg/runtime"
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

	NotificationService struct {
		ctx context.Context
	}

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

func NewEventNotifier(ctx context.Context) INotificationService {
	return &NotificationService{ctx: ctx}
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

	if isEmptyContext(e.ctx) {
		return
	}

	runtime.EventsEmit(e.ctx, "Notify", notify)
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

func isEmptyContext(ctx context.Context) bool {
	ctxString := reflect.TypeOf(ctx).String()
	return ctxString == "*context.emptyCtx" || ctxString == "context.backgroundCtx"
}
