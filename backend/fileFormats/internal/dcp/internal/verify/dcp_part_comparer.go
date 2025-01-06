package verify

import (
	"bytes"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/logger"
	"fmt"
	"os"
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
	CompareGameDataBinaryParts(partsList components.IList[parts.DcpFileParts]) error

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
	CompareTranslatedTextParts(partsList components.IList[parts.DcpFileParts]) error
}

type PartComparer struct {
	logger.ILoggerHandler
}

func newPartComparer() IPartComparer {
	return &PartComparer{
		ILoggerHandler: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "part_comparer").Logger(),
		},
	}
}

func (pc PartComparer) CompareGameDataBinaryParts(partsList components.IList[parts.DcpFileParts]) error {
	errChan := make(chan error, partsList.GetLength())
	successChan := make(chan string, partsList.GetLength())
	defer close(errChan)

	compareBinaryParts := func(part parts.DcpFileParts) {
		if err := pc.compare(part.Source().Get().Path, part.Destination().Import().Get().GetTargetFile()); err != nil {
			pc.LogError(err, "failed to compare gamedata binary parts: Part: %s", part.Destination().Import().Get().GetTargetFile())

			errChan <- err
			return
		}
		successChan <- part.Source().Get().Name
	}

	partsList.ForEach(compareBinaryParts)

	select {
	case err := <-errChan:
		pc.LogError(err, "failed to compare gamedata binary parts")
	case <-successChan:
		pc.LogInfo("gamedata binary parts are equal: %s", <-successChan)
	}

	return nil
}

func (pc PartComparer) CompareTranslatedTextParts(partsList components.IList[parts.DcpFileParts]) error {
	errChan := make(chan error, partsList.GetLength())
	successChan := make(chan string, partsList.GetLength())
	defer close(errChan)
	defer close(successChan)

	compareTextParts := func(item parts.DcpFileParts) {
		if err := pc.compare(item.Destination().Translate().Get().GetTargetFile(), item.Destination().Extract().Get().GetTargetFile()); err != nil {
			pc.LogError(err, "failed to compare translated text parts: Part: %s", item.Destination().Import().Get().GetTargetFile())

			errChan <- err
			return
		}
		successChan <- item.Source().Get().Name
	}

	partsList.ForEach(compareTextParts)

	select {
	case err := <-errChan:
		pc.LogError(err, "failed to compare translated text parts")
	case <-successChan:
		pc.LogInfo("translated text parts are equal: %s", <-successChan)
	}

	return nil
}

func (pc PartComparer) compare(fromFile, toFile string) error {
	newExtractedPartData, err := os.ReadFile(fromFile)
	if err != nil {
		pc.LogError(err, "error when reading extracted part: %s", fromFile)

		return fmt.Errorf("error when reading extracted part")
	}

	importedPartData, err := os.ReadFile(toFile)
	if err != nil {
		pc.LogError(err, "error when reading imported part: %s", toFile)

		return fmt.Errorf("error when reading imported part")
	}

	if !bytes.Equal(newExtractedPartData, importedPartData) {
	/* 	pc.log.Error().
			Str("fromFile", fromFile).
			Str("toFile", toFile).
			Int("expectedLength", len(importedPartData)).
			Int("gotLength", len(newExtractedPartData)).
			Msg("Extracted part is different from imported part") */

		pc.LogError(nil, "extracted part is different from imported part: %s", fromFile)

		return fmt.Errorf("extracted part is different from imported part")
	}

	return nil
}
