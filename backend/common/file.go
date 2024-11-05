package common

import (
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func RemoveFile(path string) error {
	return os.RemoveAll(path)
}

func OpenFile(path string) (*os.File, error) {
	return os.Open(path)
}

func ReadFile(path string) (string, error) {
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

func RemoveFileExtension(filePath string) string {
	ext := filepath.Ext(filePath)
	return filePath[:len(filePath)-len(ext)]
}
