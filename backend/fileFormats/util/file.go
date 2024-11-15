package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DuplicateFile copies a file from the source path to the destination path.
// It ensures that the destination directory exists before creating the file.
// If the source path is a directory, an error is returned.
//
// Parameters:
//   - src: The path to the source file.
//   - dst: The path to the destination file.
//
// Returns:
//   - error: An error if any issue occurs during the file duplication process.
func DuplicateFile(src string, dst string) error {
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

	if _, err = io.Copy(outputFile, inputFile); err != nil {
		return fmt.Errorf("error when copying the contents: %w", err)
	}

	if err = outputFile.Sync(); err != nil {
		return fmt.Errorf("error when synchronizing the destination file: %w", err)
	}

	return nil
}
