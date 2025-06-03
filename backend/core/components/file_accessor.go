package components

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	ModsFolder   = "mods/"
	csvLineRegex = `,(?=([^\"]*\"[^\"]*\")*[^\"]*$)`
)

var (
	ResourcesRoot = `D:\Steam\steamapps\common\FINAL FANTASY FFX&FFX-2 HD Remaster\data\FFX_Data`
	GameFilesRoot = ResourcesRoot
	DisableMods   = false
)

func getRealFile(path string) string {
	return filepath.Join(GameFilesRoot, path)
}

func GetModdedFile(path string) string {
	return filepath.Join(GameFilesRoot, ModsFolder, path)
}

// ResolveFile resolves a file path, handling both relative and absolute paths
// For relative paths: combines with GameFilesRoot and checks for modded versions
// For absolute paths: returns the absolute version of the path without modification
//
// Parameters:
//   - path: File path to resolve (can be relative or absolute)
//   - print: If true, prints the resolved path for debugging
//
// Returns:
//   - string: The resolved absolute file path
//   - error: Error if the file doesn't exist or path resolution fails
//
// Behavior:
//   - Absolute paths: Uses filepath.Abs() to ensure proper absolute path format
//   - Relative paths: Combines with GameFilesRoot, checks mods folder first (unless disabled)
//   - Always verifies the final path exists before returning
func ResolveFile(path string, print bool) (string, error) {
	var filePath string

	// Check if path is already absolute
	if filepath.IsAbs(path) {
		cPath := filepath.Clean(path)
		filePath = cPath
	} else {
		// Handle relative paths with mod support
		if DisableMods {
			filePath = getRealFile(path)
		} else {
			modded := GetModdedFile(path)
			if _, err := os.Stat(modded); err == nil {
				filePath = modded
			} else {
				filePath = getRealFile(path)
			}
		}
	}

	if print {
		fmt.Println("---", filePath, "---")
	}

	if _, err := os.Stat(filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

// WriteStringToFile writes a string to a file using UTF-8 encoding
// Creates necessary directories before writing
func WriteStringToFile(path string, content string) error {
	CreateDirectories(path)

	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Failed to write file")
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println("Failed to write file")
		return err
	}

	return file.Sync()
}

// WriteByteArrayToFile writes a slice of integers as bytes to a file
// Creates necessary directories before writing
func WriteByteArrayToFile(path string, bytes []byte) error {
	CreateDirectories(path)

	err := os.WriteFile(path, bytes, 0644)
	if err != nil {
		fmt.Println("Failed to write file")
		return err
	}

	return nil
}

// CreateDirectories creates the necessary directories for a given file path
// If the path is a directory, it ensures the directory exists
// If the path is a file, it creates all necessary parent directories
func CreateDirectories(path string) {
	info, err := os.Stat(path)
	var dirPath string

	if err == nil && info.IsDir() {
		dirPath = path
	} else {
		dirPath = filepath.Dir(path)
	}

	if dirPath == "" || dirPath == "." {
		return
	}

	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		fmt.Println("Failed to create directories: ", err)
	}
}

// ExampleResolveFileUsage demonstrates how ResolveFile handles different path types
func ExampleResolveFileUsage() {
	// Example 1: Relative path (will be combined with GameFilesRoot)
	relativePath := "event/obj/06/0601/0601.ebp"
	resolved1, err1 := ResolveFile(relativePath, true)
	if err1 == nil {
		fmt.Printf("Relative path resolved: %s\n", resolved1)
		// Output example: D:\Steam\...\FFX_Data\event\obj\06\0601\0601.ebp
	}

	// Example 2: Absolute path (will be returned as absolute without GameFilesRoot)
	absolutePath := `C:\MyCustomMods\event\custom_event.ebp`
	resolved2, err2 := ResolveFile(absolutePath, true)
	if err2 == nil {
		fmt.Printf("Absolute path resolved: %s\n", resolved2)
		// Output: C:\MyCustomMods\event\custom_event.ebp
	}

	// Example 3: Relative path with mods (will check mods folder first)
	moddedFile := "battle/kernel/kernel.bin"
	resolved3, err3 := ResolveFile(moddedFile, true)
	if err3 == nil {
		fmt.Printf("Modded file resolved: %s\n", resolved3)
		// Will check: GameFilesRoot/mods/battle/kernel/kernel.bin first
		// Then fall back to: GameFilesRoot/battle/kernel/kernel.bin
	}

	// Example 4: Current directory relative path
	currentDirPath := "./local_file.txt"
	resolved4, err4 := ResolveFile(currentDirPath, true)
	if err4 == nil {
		fmt.Printf("Current dir path resolved: %s\n", resolved4)
	}
}
