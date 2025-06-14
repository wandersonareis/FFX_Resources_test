package common

import (
	"fmt"
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
	DisableMods   = true
)

func SetGameFilesRoot(path string) {
	if path == "" {
		fmt.Println("Invalid path provided for GameFilesRoot")
		return
	}
	GameFilesRoot = filepath.Clean(path)
	fmt.Printf("GameFilesRoot set to: %s\n", GameFilesRoot)
}

func getRealFile(path string) string {
	return filepath.Join(GameFilesRoot, path)
}

func getModdedFile(path string) string {
	return filepath.Join(GameFilesRoot, ModsFolder, path)
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

// WriteBytesToFile writes a slice of integers as bytes to a file
// Creates necessary directories before writing
func WriteBytesToFile(path string, bytes []byte) error {
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

type FileAccessor struct {
	RootPath     string
	ResolvedPath string
	Info         os.FileInfo
	Size         int64
	Exists       bool
}

// NewFileAccessor creates a new FileAccessor instance for the given path
// This function handles both existing and non-existing paths gracefully
//
// Parameters:
//   - path: File path to access (can be relative or absolute)
//
// Returns:
//   - *FileAccessor: FileAccessor instance with file information
//   - error: Error if path resolution fails (not if file doesn't exist)
func NewFileAccessor(path string) (FileAccessor, error) {
	resolvedPath, err := resolvePath(path)
	if err != nil {
		return FileAccessor{}, err
	}

	fileInfo, exists := getFileInfo(resolvedPath)

	return FileAccessor{
		RootPath:     path,
		ResolvedPath: resolvedPath,
		Info:         fileInfo,
		Size:         getFileSize(fileInfo),
		Exists:       exists,
	}, nil
}

// resolvePath resolves the given path using the same logic as ResolveFile
// but without checking if the file exists
func resolvePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	if DisableMods {
		return getRealFile(path), nil
	}

	return resolveModdedPath(path), nil
}

// resolveModdedPath checks for modded version first, then falls back to original
func resolveModdedPath(path string) string {
	moddedPath := getModdedFile(path)
	if fileExists(moddedPath) {
		return moddedPath
	}
	return getRealFile(path)
}

// getFileInfo attempts to get file info, returns nil if file doesn't exist
func getFileInfo(path string) (os.FileInfo, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	return info, true
}

// getFileSize safely extracts file size from FileInfo
func getFileSize(info os.FileInfo) int64 {
	if info == nil {
		return 0
	}
	return info.Size()
}

// fileExists checks if a file exists without returning detailed error info
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
