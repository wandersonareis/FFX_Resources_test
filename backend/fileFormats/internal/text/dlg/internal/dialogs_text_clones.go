package internal

import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
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
	dataInfo interactions.IGameDataInfo
	log      zerolog.Logger
}

func NewDlgClones(dataInfo interactions.IGameDataInfo) *dialogsClones {
	return &dialogsClones{
		dataInfo: dataInfo,
		log:      logger.Get().With().Str("module", "dialogs_clones").Logger(),
	}
}

func (dc *dialogsClones) Clone() {
	if dc.dataInfo.GetGameData().ClonedItems != nil {
		for _, clone := range dc.dataInfo.GetGameData().ClonedItems {
			cloneReimportPath := filepath.Join(dc.dataInfo.GetImportLocation().TargetDirectory, clone)

			if err := dc.duplicateFile(dc.dataInfo.GetImportLocation().TargetFile, cloneReimportPath); err != nil {
				dc.log.Error().Err(err).Str("File", clone).Str("TargetPath", cloneReimportPath).Msg("Error duplicating dialog file")
				continue
			}
		}

		dc.log.Info().Msgf("All duplicated dialog files for %s have been created", dc.dataInfo.GetGameData().Name)
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