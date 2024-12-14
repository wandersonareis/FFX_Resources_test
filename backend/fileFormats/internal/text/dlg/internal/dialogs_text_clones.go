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

	"github.com/rs/zerolog"
)

type IDlgClones interface {
	Clone()
}

type dialogsClones struct {
	source      interfaces.ISource
	destination locations.IDestination

	log zerolog.Logger
}

func NewDlgClones(source interfaces.ISource, destination locations.IDestination) *dialogsClones {
	return &dialogsClones{
		source:      source,
		destination: destination,

		log: logger.Get().With().Str("module", "dialogs_clones").Logger(),
	}
}

func (dc *dialogsClones) Clone() {
	importTargetFile := dc.destination.Import().Get().GetTargetFile()
	dc.log.Info().
		Str("Clones from: ", importTargetFile).
		Msg("Creating duplicated files for")

	if dc.source.Get().ClonedItems != nil {
		for _, clone := range dc.source.Get().ClonedItems {
			cloneReimportPath := filepath.Join(dc.destination.Import().Get().GetTargetDirectory(), clone)

			if err := dc.duplicateFile(importTargetFile, cloneReimportPath); err != nil {
				dc.log.Error().
					Err(err).
					Str("Clone from: ", importTargetFile).
					Str("Clone path: ", clone).
					Msg("Error duplicating dialog file")

				continue
			}
		}

		dc.log.Info().
			Str("Clone from: ", importTargetFile).
			Msgf("Create files clones for %s successfully", dc.source.Get().Name)
	}
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
