package verify

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
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
	CompareGameDataBinaryParts(partsList *[]parts.LockitFileParts) error

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
	CompareTranslatedTextParts(partsList *[]parts.LockitFileParts) error
}

type PartComparer struct {
	worker common.IWorker[parts.LockitFileParts]
}

func newPartComparer() IPartComparer {
	worker := common.NewWorker[parts.LockitFileParts]()
	return &PartComparer{
		worker: worker,
	}
}

func (pc PartComparer) CompareGameDataBinaryParts(partsList *[]parts.LockitFileParts) error {
	compareBinaryParts := func(index int, part parts.LockitFileParts) error {
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

func (pc PartComparer) CompareTranslatedTextParts(partsList *[]parts.LockitFileParts) error {
	compareTextParts := func(index int, item parts.LockitFileParts) error {
		if err := pc.compare(item.GetTranslateLocation().TargetFile, item.GetExtractLocation().TargetFile); err != nil {
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
