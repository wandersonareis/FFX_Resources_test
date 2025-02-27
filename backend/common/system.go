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
