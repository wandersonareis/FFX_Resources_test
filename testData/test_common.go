package testcommon

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetTestDataRootDirectory searches upwards from the current test file's directory to locate the project root directory,
// identified by the presence of a "go.mod" file. It returns the path to the directory containing "go.mod".
// If no such file is found in any parent directory, the function panics.
func GetTestDataRootDirectory() string {
	// Obtém o diretório do arquivo de teste
	return filepath.Dir(getCurrentFilePath())
}

// SetBuildBinPath sets the "APP_BASE_PATH" environment variable to the build/bin directory
// within the project root. It checks for the existence of the "resources.asar" file in that
// directory and returns an error if the file is not found or if setting the environment
// variable fails.
func SetBuildBinPath() error {
	projectRoot := findProjectRoot()
	buildBinPath := filepath.Join(projectRoot, "build", "bin")
	asarFilePath := filepath.Join(buildBinPath, "resources.asar")

	if _, err := os.Stat(asarFilePath); os.IsNotExist(err) {
		return fmt.Errorf("resources.asar file not found in the build directory: %v", err)
	}

	if err := os.Setenv("APP_BASE_PATH", buildBinPath); err != nil {
		return fmt.Errorf("error by setting the construction directory path: %v", err)
	}

	return nil
}

// FindProjectRoot searches upwards from the current test file's directory to locate the project root directory,
// identified by the presence of a "go.mod" file. It returns the path to the directory containing "go.mod".
// If no such file is found in any parent directory, the function panics.
func findProjectRoot() string {
	// Obtém o diretório do arquivo de teste
	testDir := filepath.Dir(getCurrentFilePath())
	
	// Sobe na hierarquia até encontrar go.mod
	currentDir := testDir
	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			panic("Arquivo go.mod não encontrado na hierarquia de diretórios")
		}
		currentDir = parentDir
	}
}

// getCurrentFilePath returns the file path of the caller function.
// It uses runtime.Caller to retrieve the filename of the calling function's source file.
// If the file path cannot be determined, the function panics with an error message.
func getCurrentFilePath() string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("Erro ao obter o caminho do arquivo de teste")
	}
	return filename
}