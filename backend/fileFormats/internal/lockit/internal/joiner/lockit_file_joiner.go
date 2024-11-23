package joiner

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
	"slices"
)

type IPartsJoiner interface {
	JoinFileParts() error
	EncodeFilesParts() error
	FindTranslatedTextParts() (*[]parts.LockitFileParts, error)
}

type lockitFileJoiner struct {
	*base.FormatsBase
	partsList *[]parts.LockitFileParts
	options   interactions.LockitFileOptions
	worker    common.IWorker[parts.LockitFileParts]
}

func NewLockitFileJoiner(dataInfo interactions.IGameDataInfo, partsList *[]parts.LockitFileParts) IPartsJoiner {
	return &lockitFileJoiner{
		FormatsBase: base.NewFormatsBase(dataInfo),
		options:     interactions.NewInteraction().GamePartOptions.GetLockitFileOptions(),
		partsList:   partsList,
		worker:      common.NewWorker[parts.LockitFileParts](),
	}
}

func (lj *lockitFileJoiner) FindTranslatedTextParts() (*[]parts.LockitFileParts, error) {
	partsList := []parts.LockitFileParts{}
	err := util.FindFileParts(
		&partsList,
		lj.GetTranslateLocation().TargetPath,
		lib.LOCKIT_TXT_PARTS_PATTERN,
		parts.NewLockitFileParts)

	if err != nil {
		return nil, err
	}

	partsList = slices.Clip(partsList)

	return &partsList, nil
}

func (lj *lockitFileJoiner) EncodeFilesParts() error {
	lj.worker.ParallelForEach(lj.partsList,
		func(index int, part parts.LockitFileParts) {
			if index > 0 && index%2 == 0 {
				part.Compress(parts.LocEnc)
			} else {
				part.Compress(parts.FfxEnc)
			}
		})

	if err := lj.GetTranslateLocation().ProvideTargetPath(); err != nil {
		return err
	}

	return nil
}

func (lj *lockitFileJoiner) JoinFileParts() error {
	importLocation := lj.GetImportLocation()

	if len(*lj.partsList) != lj.options.PartsLength {
		return fmt.Errorf("invalid number of parts: %d expected: %d", len(*lj.partsList), lj.options.PartsLength)
	}

	var combinedBuffer bytes.Buffer

	err := lj.worker.ForEach(lj.partsList, func(_ int, part parts.LockitFileParts) error {
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
	}

	if err := importLocation.ProvideTargetPath(); err != nil {
		return err
	}

	if err := os.WriteFile(importLocation.TargetFile, combinedBuffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("error when creating output file: %v", err)
	}

	lj.worker.Close()

	return nil
}
