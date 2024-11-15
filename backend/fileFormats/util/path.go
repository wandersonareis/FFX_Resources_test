package util

import (
	"fmt"
	"os"
	"path/filepath"
)

func EnsurePathExists(path string) error {
	cPath := filepath.Clean(path)

	if filepath.Ext(cPath) != "" {
		cPath = filepath.Dir(cPath)
	}

	if err := os.MkdirAll(cPath, os.ModePerm); err != nil {
		return fmt.Errorf("error when creating the destination directory: %w", err)
	}

	return nil
}
