package joiner

import (
	"bytes"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitFileParts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type IPartsJoiner interface {
	JoinFileParts() error
	FindTranslatedTextParts() (components.IList[lockitFileParts.LockitFileParts], error)
}

type lockitFileJoiner struct {
	*base.FormatsBase
	logger.ILoggerHandler

	partsList components.IList[lockitFileParts.LockitFileParts]
	options   interactions.LockitFileOptions
	log       zerolog.Logger
}

func NewLockitFileJoiner(source interfaces.ISource, destination locations.IDestination, partsList components.IList[lockitFileParts.LockitFileParts]) IPartsJoiner {
	return &lockitFileJoiner{
		FormatsBase: base.NewFormatsBase(source, destination),
		ILoggerHandler: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "lockit_file_joiner").Logger(),
		},
		options:     interactions.NewInteractionService().DcpAndLockitOptions.GetLockitFileOptions(),
		partsList:   partsList,
	}
}

func (lj *lockitFileJoiner) FindTranslatedTextParts() (components.IList[lockitFileParts.LockitFileParts], error) {
	partsList := components.NewEmptyList[lockitFileParts.LockitFileParts]()

	err := components.GenerateGameFileParts(
		partsList,
		lj.Destination().Translate().Get().GetTargetPath(),
		lib.LOCKIT_TXT_PARTS_PATTERN,
		lockitFileParts.NewLockitFileParts)

	if err != nil {
		lj.LogError(err, "error when finding translated text parts", "path", lj.Destination().Translate().Get().GetTargetPath())

		return nil, err
	}

	partsList.Clip()

	return partsList, nil
}

func (lj *lockitFileJoiner) JoinFileParts() error {
	importLocation := lj.Destination().Import().Get()

	if lj.partsList.GetLength() != lj.options.PartsLength {
		return fmt.Errorf("invalid number of parts: %d expected: %d", lj.partsList.GetLength(), lj.options.PartsLength)
	}

	var combinedBuffer bytes.Buffer

	errChan := make(chan error, lj.partsList.GetLength())

	combineFilesFunc := func(part lockitFileParts.LockitFileParts) {
		fileName := part.Destination().Import().Get().GetTargetFile()

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
		lj.LogError(err, "error when creating output file", "file", importLocation.GetTargetFile())

		return fmt.Errorf("error when creating output file: %v", err)
	}

	return nil
}
