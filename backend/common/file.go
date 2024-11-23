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

func ReadFileAsString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ChangeExtension(path, newExt string) string {
	ext := filepath.Ext(path)
	return path[:len(path)-len(ext)] + newExt
}

func AddExtension(path, newExt string) string {
	ext := filepath.Ext(path)
	if ext == newExt {
		return path
	}

	return path + newExt
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

func RecursiveRemoveAllExtensions(filePath string) string {
	base := filepath.Base(filePath)

	parts := strings.Split(base, ".")

	if len(parts) == 1 {
		return filePath
	}

	trimmed := parts[0]

	return RecursiveRemoveAllExtensions(filepath.Join(filepath.Dir(filePath), trimmed))
}
