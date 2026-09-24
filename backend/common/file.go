package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetFileName(path string) string {
	return filepath.Base(path)
}

func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error when reading file %s: %s", GetFileName(path), err)
	}

	return data, nil
}

func OpenFile(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error when opening file %s: %s", GetFileName(path), err)
	}

	return file, nil
}

func RemoveOneFileExtension(filePath string) string {
	ext := filepath.Ext(filePath)
	return filePath[:len(filePath)-len(ext)]
}

func RecursiveRemoveFileExtension(filePath string) string {
	base := filepath.Base(filePath)

	parts := strings.Split(base, ".")

	if len(parts) <= 2 {
		return filePath
	}

	trimmed := strings.Join(parts[:len(parts)-1], ".")

	return RecursiveRemoveFileExtension(filepath.Join(filepath.Dir(filePath), trimmed))
}
