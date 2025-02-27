package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type TempProvider struct {
	TempFile     string
	TempFilePath string
}

func NewTempProvider(fileName, ext string) *TempProvider {
	prefix := "temp"
	if fileName != "" {
		prefix = fileName
	}

	tempProvider := &TempProvider{}

	tmpExt := ".tmp"
	if ext != "" {
		tmpExt = tempProvider.validExtension(ext)
	}

	tempPath := filepath.Join(os.TempDir(), "ffxresources")
	if err := EnsurePathExists(tempPath); err != nil {
		panic(fmt.Sprintf("Error creating temp directory: %s", err))
	}

	tempProvider.TempFilePath = tempPath

	tmpFileName := fmt.Sprintf("%s_%s.%s", prefix, GetUUID(), tmpExt)
	file := filepath.Join(tempPath, tmpFileName)

	tempProvider.TempFile = file
	tempProvider.TempFilePath = tempPath

	return tempProvider
}

// Dispose removes the temporary file associated with the TempProvider instance.
// It calls os.Remove on the file path stored in the tp.File field.
func (tp *TempProvider) Dispose() {
	os.Remove(tp.TempFile)
	RemoveDir(tp.TempFilePath)
	tp.TempFile = ""
	tp.TempFilePath = ""
}

func (tp *TempProvider) cleanExtension(extension string) string {
    if !strings.HasPrefix(extension, ".") {
        return extension
    }
    
    return tp.cleanExtension(strings.TrimPrefix(extension, "."))
}

// validExtension takes a file extension and returns a valid file extension.
// If the given extension already starts with a '.', it is returned as is.
// Otherwise, a '.' is prepended to the extension and it is returned.
func (tp *TempProvider) validExtension(extension string) string {
	if extension == "" {
		return ""
	}

	extension = tp.cleanExtension(extension)
	
	return extension
}
