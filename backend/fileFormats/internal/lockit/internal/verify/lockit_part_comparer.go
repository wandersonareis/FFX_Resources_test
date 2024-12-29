package verify

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitFileParts"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
	"fmt"
	"os"
)

type IPartComparer interface {
	// CompareGameDataBinaryParts compares the gamedata binary parts with imported binary parts of the given LockitFileParts slice.
	// It iterates over each part and compares the game data file path with the import location target file.
	// If any comparison fails, it returns an error.
	//
	// Parameters:
	//   partsList []LockitFileParts - A slice of LockitFileParts to be compared.
	//
	// Returns:
	//   error - An error if any comparison fails, otherwise nil.
	CompareGameDataBinaryParts(partsList components.IList[lockitFileParts.LockitFileParts]) error

	// CompareTranslatedTextParts compares the translated text parts with extracted text parts of the given LockitFileParts.
	// It iterates over each part and compares the target files of the translate
	// and extract locations using the PartComparer.
	//
	// Parameters:
	//
	//	partsList []LockitFileParts - A slice of LockitFileParts to be compared.
	//
	// Returns:
	//
	//	error - An error if any comparison fails, otherwise nil.
	CompareTranslatedTextParts(partsList components.IList[lockitFileParts.LockitFileParts]) error
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

func (pc PartComparer) CompareGameDataBinaryParts(partsList components.IList[lockitFileParts.LockitFileParts]) error {
	errChan := make(chan error, partsList.GetLength())
	defer close(errChan)

	go notifications.ProcessError(errChan, pc.GetLogger())

	compareBinaryParts := func(part lockitFileParts.LockitFileParts) {
		if err := pc.compare(part.Source().Get().Path, part.Destination().Import().Get().GetTargetFile()); err != nil {
			errChan <- err
			return
		}
	}

	partsList.ForEach(compareBinaryParts)

	return nil
}

func (pc PartComparer) CompareTranslatedTextParts(partsList components.IList[lockitFileParts.LockitFileParts]) error {
	errChan := make(chan error, partsList.GetLength())
	defer close(errChan)

	go notifications.ProcessError(errChan, pc.GetLogger())

	compareTextParts := func(item lockitFileParts.LockitFileParts) {
		if err := pc.compare(item.Destination().Translate().Get().GetTargetFile(), item.Destination().Extract().Get().GetTargetFile()); err != nil {
			errChan <- err
			return
		}
	}

	partsList.ForEach(compareTextParts)

	return nil
}

func (pc PartComparer) compare(fromFile, toFile string) error {
	newExtractedPartData, err := os.ReadFile(fromFile)
	if err != nil {
		return fmt.Errorf("error when reading extracted part: %s", common.GetFileName(fromFile))
	}

	importedPartData, err := os.ReadFile(toFile)
	if err != nil {
		return fmt.Errorf("error when reading imported part: %s", common.GetFileName(toFile))
	}

	if !bytes.Equal(newExtractedPartData, importedPartData) {
		return fmt.Errorf("file part: %s Expected size: %d Got size: %d", common.GetFileName(fromFile), len(importedPartData), len(newExtractedPartData))
	}

	return nil
}
