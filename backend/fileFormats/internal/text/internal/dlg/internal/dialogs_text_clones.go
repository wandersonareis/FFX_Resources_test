package internal

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type IDlgClones interface {
	Clone(source interfaces.ISource, destination locations.IDestination)
}

type dialogsClones struct {
	log logger.ILoggerHandler
}

func NewDlgClones(logger logger.ILoggerHandler) *dialogsClones {
	return &dialogsClones{
		log: logger,
	}
}

func (dc *dialogsClones) Clone(source interfaces.ISource, destination locations.IDestination) {
	importTargetFile := destination.Import().Get().GetTargetFile()

	fileClones := source.Get().ClonedItems
	if len(fileClones) == 0 {
		return
	}

	dc.log.LogInfo("Clones from: %s", importTargetFile)

	for _, clone := range fileClones {
		cloneReimportPath := filepath.Join(destination.Import().Get().GetTargetDirectory(), clone)

		if err := dc.duplicateFile(importTargetFile, cloneReimportPath); err != nil {
			dc.log.LogError(err, "Error duplicating file: %s", cloneReimportPath)
			continue
		}
	}

	dc.log.LogInfo("Create %d files clones for %s successfully", len(fileClones), source.Get().Name)
}

// It ensures that the destination directory exists before creating the file.
// If the source path is a directory, an error is returned.
//
// Parameters:
//   - src: The path to the source file.
//   - dst: The path to the destination file.
//
// Returns:
//   - error: An error if any issue occurs during the file duplication process.
func (dc *dialogsClones) duplicateFile(src string, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error when accessing the origin file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path of origin is not a file")
	}

	outputDirectory := filepath.Dir(dst)

	err = util.EnsurePathExists(outputDirectory)
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
