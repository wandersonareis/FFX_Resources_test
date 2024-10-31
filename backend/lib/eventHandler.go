package lib

import (
	"ffxresources/backend/interactions"
	"fmt"
	"log"

	goRT "runtime"

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

func logInFile(msg string) {
	log.Println(msg)

}
func captureTrace() (string, string, int) {
	// Pega o frame da stack do local de onde essa função foi chamada
	pc, file, line, ok := goRT.Caller(2) // 2 níveis acima (porque estamos chamando de dentro do logger)
	if !ok {
		return "unknown", "unknown", 0
	}

	// Obtém o nome da função a partir do ponteiro para o programa counter (pc)
	fn := goRT.FuncForPC(pc)
	if fn == nil {
		return file, "unknown", line
	}

	return file, fn.Name(), line
}

func LogSeverity(severity Severity, message string) {
	file, funcName, line := captureTrace()

	trace := "file: " + file + ", func: " + funcName + ", line: " + string(line)
	debugLine := fmt.Sprintf("Severity: %s, Message: %s, Trace: %s", severity.String(), message, trace)

	switch severity {
	case SeveritySuccess:
		fallthrough
	case SeverityInfo:
		logInFile("Info: " + message)
	case SeverityWarn:
		logInFile("Warn: " + message)
	case SeverityError:
		logInFile("Error: " + debugLine)
	default:
		logInFile(debugLine)
	}
}

func Notify(notification Severity, message string) {
	context := interactions.NewInteraction().Ctx
	notify := Notification{
		Message:  message,
		Severity: notification.String(),
	}

	runtime.EventsEmit(context, "Notify", notify)
	LogSeverity(notification, message)
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
