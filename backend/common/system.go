package common

import (
	"log"
	"os"
	"path/filepath"
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

const (
	colorReset   = "\033[0m"
	colorVerbose = "\033[33m"
	colorInfo    = "\033[32m"
	colorError   = "\033[31m"
)

func LogVerbose(format string, args ...any) {
	if IsVerboseMode() {
		log.Printf(colorVerbose+format+colorReset, args...)
	}
}

func LogInfo(format string, args ...any) {
	log.Printf(colorInfo+format+colorReset, args...)
}

func LogError(format string, args ...any) {
	log.Printf(colorError+format+colorReset, args...)
}

func AreModsEnabled() bool {
	return !DisableMods
}
