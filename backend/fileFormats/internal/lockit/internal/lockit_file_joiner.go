package internal

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"os"
)

type IPartsJoiner interface {
	JoinFileParts() error
}

type lockitFileJoiner struct {
	*base.FormatsBase

	partsList components.IList[lockitParts.LockitFileParts]
	options   core.ILockitFileOptions
	log       logger.ILoggerHandler
}

func NewLockitFileJoiner(source interfaces.ISource, destination locations.IDestination, partsList components.IList[lockitParts.LockitFileParts]) IPartsJoiner {
	return &lockitFileJoiner{
		FormatsBase: base.NewFormatsBase(source, destination),
		options:     core.NewLockitFileOptions(interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber()),
		partsList:   partsList,

		log: logger.NewLoggerHandler("lockit_file_joiner"),
	}
}

func (lj *lockitFileJoiner) JoinFileParts() error {
	importLocation := lj.Destination().Import().Get()

	if lj.partsList.GetLength() != lj.options.GetPartsLength() {
		return fmt.Errorf("invalid number of parts: %d expected: %d", lj.partsList.GetLength(), lj.options.GetPartsLength())
	}

	var combinedBuffer bytes.Buffer

	errChan := make(chan error, lj.partsList.GetLength())

	combineFilesFunc := func(part lockitParts.LockitFileParts) {
		translatedTextFile := part.Destination().Translate().Get().GetTargetFile()
		fileName := common.RemoveOneFileExtension(translatedTextFile) // remove .txt extension

		partData, err := os.ReadFile(fileName)
		if err != nil {
			errChan <- fmt.Errorf("error when reading file part %s: %w", part.Source().Get().Path, err)
			return
		}

		combinedBuffer.Write(partData)
	}

	lj.partsList.ForEach(combineFilesFunc)

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
