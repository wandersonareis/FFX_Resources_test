package testcommon

import (
	"bufio"
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

// RemoveFirstNLines removes the first n lines from the file specified by filePath.
// It reads the file, skips the first n lines, and writes the remaining lines back to the file.
// Returns an error if the file cannot be opened, read, or written.
func RemoveFirstNLines(filePath string, n int) error {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			var remainingLines []string
			scanner := bufio.NewScanner(file)

			lineNumber := 0
			for scanner.Scan() {
				if lineNumber >= n {
					remainingLines = append(remainingLines, scanner.Text())
				}
				lineNumber++
			}

			if err := scanner.Err(); err != nil {
				return err
			}

			file, err = os.Create(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			writer := bufio.NewWriter(file)
			for _, line := range remainingLines {
				_, err := writer.WriteString(line + "\n")
				if err != nil {
					return err
				}
			}

			return writer.Flush()
		}