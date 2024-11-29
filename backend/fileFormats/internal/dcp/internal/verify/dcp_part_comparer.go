package verify

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/logger"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type IPartComparer interface {
	// CompareGameDataBinaryParts compares the gamedata binary parts with imported binary parts of the given DcpFileParts slice.
	// It iterates over each part and compares the game data file path with the import location target file.
	// If any comparison fails, it returns an error.
	//
	// Parameters:
	//   partsList []DcpFileParts - A slice of DcpFileParts to be compared.
	//
	// Returns:
	//   error - An error if any comparison fails, otherwise nil.
	CompareGameDataBinaryParts(partsList *[]parts.DcpFileParts) error

	// CompareTranslatedTextParts compares the translated text parts with extracted text parts of the given DcpFileParts.
	// It iterates over each part and compares the target files of the translate
	// and extract locations using the PartComparer.
	//
	// Parameters:
	//
	//	partsList []DcpFileParts - A slice of DcpFileParts to be compared.
	//
	// Returns:
	//
	//	error - An error if any comparison fails, otherwise nil.
	CompareTranslatedTextParts(partsList *[]parts.DcpFileParts) error
}

type PartComparer struct {
	log    zerolog.Logger
	worker common.IWorker[parts.DcpFileParts]
}

func newPartComparer() IPartComparer {
	worker := common.NewWorker[parts.DcpFileParts]()
	return &PartComparer{
		log:    logger.Get().With().Str("module", "dcp_parts_verify").Logger(),
		worker: worker,
	}
}

func (pc PartComparer) CompareGameDataBinaryParts(partsList *[]parts.DcpFileParts) error {
	compareBinaryParts := func(index int, part parts.DcpFileParts) error {
		if err := pc.compare(part.GetGameData().FullFilePath, part.GetImportLocation().TargetFile); err != nil {
			return err
		}

		return nil
	}

	if err := pc.worker.ForEach(partsList, compareBinaryParts); err != nil {
		return err
	}

	return nil
}

func (pc PartComparer) CompareTranslatedTextParts(partsList *[]parts.DcpFileParts) error {
	compareTextParts := func(index int, item parts.DcpFileParts) error {
		if err := pc.compare(item.GetTranslateLocation().TargetFile, item.GetExtractLocation().TargetFile); err != nil {
			pc.log.Error().
				Err(err).
				Str("part", item.GetImportLocation().TargetFile).
				Msg("Error when comparing translated text parts")

			return err
		}

		return nil
	}

	if err := pc.worker.ForEach(partsList, compareTextParts); err != nil {
		return err
	}

	return nil
}

func (pc PartComparer) compare(fromFile, toFile string) error {
	newExtractedPartData, err := os.ReadFile(fromFile)
	if err != nil {
		pc.log.Error().
			Err(err).
			Str("file", fromFile).
			Msg("Error when reading extracted part")

		return fmt.Errorf("error when reading extracted part")
	}

	importedPartData, err := os.ReadFile(toFile)
	if err != nil {
		pc.log.Error().
			Err(err).
			Str("file", toFile).
			Msg("Error when reading imported part")

		return fmt.Errorf("error when reading imported part")
	}

	if !bytes.Equal(newExtractedPartData, importedPartData) {
		pc.log.Error().
			Str("fromFile", fromFile).
			Str("toFile", toFile).
			Int("expectedLength", len(importedPartData)).
			Int("gotLength", len(newExtractedPartData)).
			Msg("Extracted part is different from imported part")

		return fmt.Errorf("extracted part is different from imported part")
	}

	return nil
}
