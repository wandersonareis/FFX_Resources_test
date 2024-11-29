package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type TempProvider struct {
	filePrefix string
	extension  string
	File       string
	FilePath   string
}

// NewTempProvider returns a new TempProvider with empty file prefix and extension.
// The FilePath will be set to a new temporary file path in the OS temp directory.
func NewTempProvider() *TempProvider {
	tempPath := os.TempDir()
	return &TempProvider{
		FilePath: tempPath,
	}
}

//TODO: Remove this function
func NewTempProviderDev(prefix, ext string) *TempProvider {
	tempPath := os.TempDir()
	uuid := uuid.New().String()

	fileName := fmt.Sprintf("%s-%s.%s", prefix, uuid, ext)
	file := filepath.Join(tempPath, fileName)
	return &TempProvider{
		FilePath: tempPath,
		File:     file,
	}
}

// baseTempProvider creates a new TempProvider with the given file prefix and extension.
// The file will be written to the OS temp directory with a UUID appended to the prefix.
// The extension will be sanitized to ensure it is in the correct format.
// The returned TempProvider will have the FilePath set to the full path of the temp file.
func (tp *TempProvider) baseTempProvider(filePrefix string, extension string) *TempProvider {
	tempPath := os.TempDir()
	uuid := uuid.New().String()
	file := filepath.Join(tempPath, filePrefix+uuid+validExtension(extension))
	return &TempProvider{
		filePrefix: filePrefix,
		extension:  extension,
		File:       file,
		FilePath:   tempPath,
	}
}

// ProvideTempFile returns a new TempProvider with the given file prefix.
// The extension will be empty.
func (tp *TempProvider) ProvideTempFile(filePrefix string) *TempProvider {
	return tp.baseTempProvider(filePrefix, "")
}

func (tp *TempProvider) ProvideTempFilePath() string {
	return tp.FilePath
}

// ProvideTempFileWithExtension returns a new TempProvider with the given file prefix and extension.
// The returned TempProvider will have the FilePath set to the full path of the temp file.
func (tp *TempProvider) ProvideTempFileWithExtension(filePrefix string, extension string) *TempProvider {
	return tp.baseTempProvider(filePrefix, extension)
}

// ProvideTempDir generates a unique temporary folder path.
// It uses the system's temporary directory and appends a new UUID to ensure uniqueness.
// Returns the full path to the temporary folder as a string.
func (tp *TempProvider) ProvideTempDir() string {
	tempPath := os.TempDir()
	uuid := uuid.New().String()
	return filepath.Join(tempPath, uuid)
}

// Dispose removes the temporary file associated with the TempProvider instance.
// It calls os.Remove on the file path stored in the tp.File field.
func (tp *TempProvider) Dispose() {
	os.Remove(tp.File)
}

// validExtension takes a file extension and returns a valid file extension.
// If the given extension already starts with a '.', it is returned as is.
// Otherwise, a '.' is prepended to the extension and it is returned.
func validExtension(extension string) string {
	if extension == "" {
		return ""
	}

	if strings.HasPrefix(extension, ".") {
		return extension
	}
	return "." + extension
}
