package interaction

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

func NewTempProvider() *TempProvider {
	return baseTempProvider("", "")
}

func NewTempProviderWithPrefix(prefix string) *TempProvider {
	return baseTempProvider(prefix, "")
}

func NewTempProviderWithPrefixAndExtension(prefix string, extension string) *TempProvider {
	return baseTempProvider(prefix, extension)
}

func baseTempProvider(filePrefix string, extension string) *TempProvider {
	tempPath := os.TempDir()
	uuid := uuid.New().String()
	return &TempProvider{
		FilePrefix: filePrefix,
		Extension:  extension,
		FilePath:   filepath.Join(tempPath, filePrefix+uuid+validExtension(extension)),
	}
}

func (tp *TempProvider) ProvideTempFile(filePrefix string) *TempProvider {
	return NewTempProviderWithPrefix(filePrefix)
}

func (tp *TempProvider) ProvideTempFileWithExtension(filePrefix string, extension string) *TempProvider {
	return NewTempProviderWithPrefixAndExtension(filePrefix, extension)
}

func validExtension(extension string) string {
	if strings.HasPrefix(extension, ".") {
		return extension
	}
	return "." + extension
}
