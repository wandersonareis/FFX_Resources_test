package common

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetFileName(path string) string {
	return filepath.Base(path)
}

func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error when reading file %s: %s", GetFileName(path), err)
	}

	return data, nil
}
