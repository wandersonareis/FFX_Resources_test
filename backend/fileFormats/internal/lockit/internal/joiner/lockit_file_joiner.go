package joiner

import (
	"bytes"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type IPartsJoiner interface {
	JoinFileParts() error
	EncodeFilesParts() error
	FindTranslatedTextParts() (components.IList[parts.LockitFileParts], error)
}

type lockitFileJoiner struct {
	*base.FormatsBase

	partsList components.IList[parts.LockitFileParts]
	options   interactions.LockitFileOptions
	log       zerolog.Logger
}

func NewLockitFileJoiner(source interfaces.ISource, destination locations.IDestination, partsList components.IList[parts.LockitFileParts]) IPartsJoiner {
	return &lockitFileJoiner{
		FormatsBase: base.NewFormatsBaseDev(source, destination),
		options:     interactions.NewInteraction().GamePartOptions.GetLockitFileOptions(),
		partsList:   partsList,

		log: logger.Get().With().Str("module", "lockit_file_joiner").Logger(),
	}
}

func (lj *lockitFileJoiner) FindTranslatedTextParts() (components.IList[parts.LockitFileParts], error) {
	partsList := components.NewEmptyList[parts.LockitFileParts]()

	err := components.GenerateGameFilePartsDev(
		partsList,
		lj.Destination().Translate().Get().GetTargetPath(),
		lib.LOCKIT_TXT_PARTS_PATTERN,
		parts.NewLockitFileParts)

	if err != nil {
		lj.log.Error().
			Err(err).
			Str("path", lj.Destination().Translate().Get().GetTargetPath()).
			Msg("error when finding translated text parts")

		return nil, err
	}

	partsList.Clip()

	return partsList, nil
}

func (lj *lockitFileJoiner) EncodeFilesParts() error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	defer encoding.Dispose()

	compressorFunc := func(index int, part parts.LockitFileParts) {
		if index > 0 && index%2 == 0 {
			part.Compress(parts.LocEnc, encoding)
		} else {
			part.Compress(parts.FfxEnc, encoding)
		}
	}

	lj.partsList.ParallelForEach(compressorFunc)

	if err := lj.Destination().Translate().Get().ProvideTargetPath(); err != nil {
		return err
	}

	return nil
}

func (lj *lockitFileJoiner) JoinFileParts() error {
	importLocation := lj.Destination().Import().Get()

	if lj.partsList.GetLength() != lj.options.PartsLength {
		return fmt.Errorf("invalid number of parts: %d expected: %d", lj.partsList.GetLength(), lj.options.PartsLength)
	}

	var combinedBuffer bytes.Buffer

	errChan := make(chan error, lj.partsList.GetLength())
	defer close(errChan)

	go notifications.ProcessError(errChan, lj.log)

	combineFilesFunc := func(part parts.LockitFileParts) {
		fileName := part.Destination().Import().Get().GetTargetFile()

		partData, err := os.ReadFile(fileName)
		if err != nil {
			errChan <- fmt.Errorf("error when reading file part %s: %w", part.Source().Get().Path, err)
			return
		}

		combinedBuffer.Write(partData)
	}

	lj.partsList.ForEach(combineFilesFunc)

	if err := importLocation.ProvideTargetPath(); err != nil {
		return err
	}

	if err := os.WriteFile(importLocation.GetTargetFile(), combinedBuffer.Bytes(), 0644); err != nil {
		lj.log.Error().
			Err(err).
			Str("file", importLocation.GetTargetFile()).
			Msg("error when creating output file")

		return fmt.Errorf("error when creating output file: %v", err)
	}

	return nil
}
