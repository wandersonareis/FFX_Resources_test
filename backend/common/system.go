package common

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func GetNumCpu() int {
	return runtime.NumCPU()
}

func GetToolExcutable(exeName string) (string, error) {
	currentDirectory := GetExecDir()

	executablePath := filepath.Join(currentDirectory, exeName)

	return exec.LookPath(executablePath)
}

func GetExecDir() string {
	exePath, _ := os.Executable()
	currentDirectory := filepath.Dir(exePath)

	return currentDirectory
}

func GetBasePath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Error when obtaining source code path")
	}

	srcPath := filepath.Dir(filename)

	if base := os.Getenv("APP_BASE_PATH"); base != "" {
		return base
	}

	exePath, err := os.Executable()
	if err == nil {
		return filepath.Dir(exePath)
	}

	return srcPath
}

func GetTempDir() string {
	return os.TempDir()
}

func GetGameVersionString() string {
	version := os.Getenv("GAME_VERSION")
	switch version {
	case "ffx":
		return "ffx"
	case "ffx2":
		return "ffx2"
	default:
		return "ffx" // Default to ffx if not set
	}
}
func SetGameVersion(version int) {
	if version < 1 || version > 2 {
		panic("Invalid game version. Must be 1 or 2.")
	}

	switch version {
	case 1:
		os.Setenv("GAME_VERSION", "ffx")
	case 2:
		os.Setenv("GAME_VERSION", "ffx2")
	default:
		os.Setenv("GAME_VERSION", "ffx") // Default to ffx if invalid
	}
}
