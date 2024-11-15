package common

import (
	"os"
	"path/filepath"
)

func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func RemoveFile(path string) error {
	return os.RemoveAll(path)
}

func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

/* func DuplicateFile(src string, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error when accessing the origin file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path of origin is not a file")
	}

	outputDirectory := filepath.Dir(dst)

	err = EnsurePathExists(outputDirectory)
	if err != nil {
		return err
	}

	inputFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error when opening the origin file:%w", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error when creating the destination file: %w", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return fmt.Errorf("error when copying the contents: %w", err)
	}

	err = outputFile.Sync()
	if err != nil {
		return fmt.Errorf("error when synchronizing the destination file: %w", err)
	}

	return nil
} */

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
