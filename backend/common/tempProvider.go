package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type TempProvider struct {
	filePrefix   string
	extension    string
	TempFile     string
	TempFilePath string
}

// NewTempProvider returns a new TempProvider with empty file prefix and extension.
// The FilePath will be set to a new temporary file path in the OS temp directory.
func NewTempProvider() *TempProvider {
	tempPath := os.TempDir()
	return &TempProvider{
		TempFilePath: tempPath,
	}
}

func NewTempProviderDev(fileName, ext string) *TempProvider {
	prefix := "temp"
	if fileName != "" {
		prefix = fileName
	}

	tempProvider := &TempProvider{}

	tmpExt := ".tmp"
	if ext != "" {
		tmpExt = tempProvider.validExtension(ext)
	}

	tempPath := os.TempDir()

	tempProvider.filePrefix = prefix
	tempProvider.extension = tmpExt
	tempProvider.TempFilePath = tempPath

	uuid := uuid.New().String()

	tmpFileName := fmt.Sprintf("%s-%s.%s", prefix, uuid, tmpExt)
	file := filepath.Join(tempPath, tmpFileName)

	tempProvider.TempFile = file

	return tempProvider
}

// Dispose removes the temporary file associated with the TempProvider instance.
// It calls os.Remove on the file path stored in the tp.File field.
func (tp *TempProvider) Dispose() {
	os.Remove(tp.TempFile)
}

// validExtension takes a file extension and returns a valid file extension.
// If the given extension already starts with a '.', it is returned as is.
// Otherwise, a '.' is prepended to the extension and it is returned.
func (tp *TempProvider) validExtension(extension string) string {
	if extension == "" {
		return ""
	}

	if strings.HasPrefix(extension, ".") {
		return extension
	}
	return "." + extension
}
