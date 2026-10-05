package common

import (
	"os"
	"path/filepath"

	"ffxresources/backend/loggingService"
)

func GetExecDir() string {
	exePath, _ := os.Executable()
	currentDirectory := filepath.Dir(exePath)

	return currentDirectory
}

func SetVerboseMode(enabled bool) {
	if enabled {
		os.Setenv("VERBOSE_MODE", "1")
	} else {
		os.Setenv("VERBOSE_MODE", "0")
	}
}
func IsVerboseMode() bool {
	verbose := os.Getenv("VERBOSE_MODE")
	return verbose == "1"
}

// Façade de log: os helpers do common delegam ao loggingService — console
// colorido pelo nível + persistência em arquivo (JSON). Logging não pode
// derrubar a aplicação: erros de escrita são ignorados pelo zerolog.

func LogVerbose(format string, args ...any) {
	loggingService.Verbose(format, args...)
}

func LogInfo(format string, args ...any) {
	loggingService.Info(format, args...)
}

// LogWarning registra um aviso recuperável (entrada ignorada, divergência de
// árvore, degradação controlada) — visível sempre, sem ser erro.
func LogWarning(format string, args ...any) {
	loggingService.Warn(format, args...)
}

func LogError(format string, args ...any) {
	loggingService.Error(format, args...)
}

func AreModsEnabled() bool {
	return !DisableMods
}
