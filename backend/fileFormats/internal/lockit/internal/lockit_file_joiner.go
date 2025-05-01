package internal

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/logger"
	"fmt"
	"os"
)

type ILockitPartsJoiner interface {
	JoinFileParts(destination locations.IDestination, lockitPartsList components.IList[lockitParts.LockitFileParts], fileOptions core.ILockitFileOptions) error
}

type lockitFileJoiner struct {
	log logger.ILoggerHandler
}

func NewLockitFileJoiner(logger logger.ILoggerHandler) ILockitPartsJoiner {
	return &lockitFileJoiner{log: logger}
}

func (lj *lockitFileJoiner) JoinFileParts(destination locations.IDestination, lockitPartsList components.IList[lockitParts.LockitFileParts], fileOptions core.ILockitFileOptions) error {
	importLocation := destination.Import()

	if lockitPartsList.GetLength() != fileOptions.GetPartsLength() {
		return fmt.Errorf("invalid number of parts: %d expected: %d", lockitPartsList.GetLength(), fileOptions.GetPartsLength())
	}

	var combinedBuffer bytes.Buffer

	errChan := make(chan error, lockitPartsList.GetLength())

	combineFilesFunc := func(part lockitParts.LockitFileParts) {
		translatedTextFile := part.GetDestination().Translate().Get().GetTargetFile()
		fileName := common.RemoveOneFileExtension(translatedTextFile) // remove .txt extension

		partData, err := os.ReadFile(fileName)
		if err != nil {
			errChan <- fmt.Errorf("error when reading file part %s: %w", part.GetSource().Get().Path, err)
			return
		}

		combinedBuffer.Write(partData)
	}

	lockitPartsList.ForEach(combineFilesFunc)

	close(errChan)
	if err := <-errChan; err != nil {
		return err
	}

	if err := importLocation.ProvideTargetPath(); err != nil {
		return err
	}

	if err := os.WriteFile(importLocation.GetTargetFile(), combinedBuffer.Bytes(), 0644); err != nil {
		lj.log.LogError(err, "error when creating output file", "file", importLocation.GetTargetFile())

		return fmt.Errorf("error when creating output file: %v", err)
	}

	return nil
}
