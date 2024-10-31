package common

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type TempProvider struct {
	FilePrefix string
	Extension  string
	FilePath   string
}

// NewTempProvider returns a new TempProvider with empty file prefix and extension.
// The FilePath will be set to a new temporary file path in the OS temp directory.
func NewTempProvider() *TempProvider {
	return baseTempProvider("", "")
}

// NewTempProviderWithPrefix returns a new TempProvider with the given file prefix.
// The file extension will be empty.
/* func NewTempProviderWithPrefix(prefix string) *TempProvider {
	return baseTempProvider(prefix, "")
} */

/* // NewTempProviderWithPrefixAndExtension returns a new TempProvider with the given file prefix and extension.
// The returned TempProvider will have the FilePath set to the full path of the temp file.
func NewTempProviderWithPrefixAndExtension(prefix string, extension string) *TempProvider {
	return baseTempProvider(prefix, extension)
} */

// baseTempProvider creates a new TempProvider with the given file prefix and extension.
// The file will be written to the OS temp directory with a UUID appended to the prefix.
// The extension will be sanitized to ensure it is in the correct format.
// The returned TempProvider will have the FilePath set to the full path of the temp file.
func baseTempProvider(filePrefix string, extension string) *TempProvider {
	tempPath := os.TempDir()
	uuid := uuid.New().String()
	return &TempProvider{
		FilePrefix: filePrefix,
		Extension:  extension,
		FilePath:   filepath.Join(tempPath, filePrefix+uuid+validExtension(extension)),
	}
}

// ProvideTempFile returns a new TempProvider with the given file prefix.
// The extension will be empty.
func (tp *TempProvider) ProvideTempFile(filePrefix string) *TempProvider {
	return baseTempProvider(filePrefix, "")
}

// ProvideTempFileWithExtension returns a new TempProvider with the given file prefix and extension.
// The returned TempProvider will have the FilePath set to the full path of the temp file.
func (tp *TempProvider) ProvideTempFileWithExtension(filePrefix string, extension string) *TempProvider {
	return baseTempProvider(filePrefix, extension)
}

// validExtension takes a file extension and returns a valid file extension.
// If the given extension already starts with a '.', it is returned as is.
// Otherwise, a '.' is prepended to the extension and it is returned.
func validExtension(extension string) string {
	if strings.HasPrefix(extension, ".") {
		return extension
	}
	return "." + extension
}
