package lib

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

func getTempDir() string {
	return os.TempDir()
}