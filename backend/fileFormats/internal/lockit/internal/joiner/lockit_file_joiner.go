package joiner

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
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
	worker    common.IWorker[parts.LockitFileParts]
}

func NewLockitFileJoiner(dataInfo interactions.IGameDataInfo, partsList components.IList[parts.LockitFileParts]) IPartsJoiner {
	return &lockitFileJoiner{
		FormatsBase: base.NewFormatsBase(dataInfo),
		options:     interactions.NewInteraction().GamePartOptions.GetLockitFileOptions(),
		partsList:   partsList,
		worker:      common.NewWorker[parts.LockitFileParts](),
	}
}

func (lj *lockitFileJoiner) FindTranslatedTextParts() (components.IList[parts.LockitFileParts], error) {
	partsList := components.NewEmptyList[parts.LockitFileParts]()

	err := components.GenerateGameFileParts(
		partsList,
		lj.GetTranslateLocation().TargetPath,
		lib.LOCKIT_TXT_PARTS_PATTERN,
		parts.NewLockitFileParts)

	if err != nil {
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

	/* lj.worker.ParallelForEach(lj.partsList,
	func(index int, part parts.LockitFileParts) {
		if index > 0 && index%2 == 0 {
			part.Compress(parts.LocEnc, encoding)
		} else {
			part.Compress(parts.FfxEnc, encoding)
		}
	}) */

	if err := lj.GetTranslateLocation().ProvideTargetPath(); err != nil {
		return err
	}

	return nil
}

func (lj *lockitFileJoiner) JoinFileParts() error {
	importLocation := lj.GetImportLocation()

	if lj.partsList.GetLength() != lj.options.PartsLength {
		return fmt.Errorf("invalid number of parts: %d expected: %d", lj.partsList.GetLength(), lj.options.PartsLength)
	}

	var combinedBuffer bytes.Buffer

	errChan := make(chan error, 1)

	combineFilesFunc := func(part parts.LockitFileParts) {
		fileName := part.GetFileInfo().GetImportLocation().TargetFile

		partData, err := os.ReadFile(fileName)
		if err != nil {
			errChan <- fmt.Errorf("error reading the separate %s: %v", fileName, err)
			return
		}

		combinedBuffer.Write(partData)
	}

	lj.partsList.ForEach(combineFilesFunc)

	var err error
	if err = <-errChan; err != nil {
		return err
	}

	/* err = lj.worker.ForEach(lj.partsList, func(_ int, part parts.LockitFileParts) error {
		fileName := part.GetFileInfo().GetImportLocation().TargetFile

		partData, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("error reading the separate %s: %v", fileName, err)
		}
		combinedBuffer.Write(partData)

		return nil
	})

	if err != nil {
		return err
	} */

	if err := importLocation.ProvideTargetPath(); err != nil {
		return err
	}

	if err := os.WriteFile(importLocation.TargetFile, combinedBuffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("error when creating output file: %v", err)
	}

	lj.worker.Close()

	return nil
}
