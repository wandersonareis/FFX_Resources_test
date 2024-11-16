package internal

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
	"slices"
)

type lockitFileJoin struct {
	*base.FormatsBase
	parts               *[]LockitFileParts
	partsSizes          *[]int
	expectedPartsLength int
}

func NewLockitFileJoiner(dataInfo interactions.IGameDataInfo, parts *[]LockitFileParts) *lockitFileJoin {
	lockitSizes := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().LockitPartsSizes
	return &lockitFileJoin{
		FormatsBase:         base.NewFormatsBase(dataInfo),
		parts:               parts,
		partsSizes:          &lockitSizes,
		expectedPartsLength: len(lockitSizes) + 1,
	}
}

func (lj *lockitFileJoin) FindTextParts() ([]LockitFileParts, error) {
	parts := []LockitFileParts{}
	err := util.FindFileParts(
		&parts,
		lj.GetTranslateLocation().TargetPath,
		util.LOCKIT_TXT_PARTS_PATTERN,
		NewLockitFileParts)

	if err != nil {
		return nil, err
	}

	return slices.Clip(parts), nil
}

func (lj *lockitFileJoin) EncodeFilesParts() error {
	worker := common.NewWorker[LockitFileParts]()

	worker.ParallelForEach(*lj.parts,
		func(index int, part LockitFileParts) {
			if index > 0 && index%2 == 0 {
				part.Compress(LocEnc)
			} else {
				part.Compress(FfxEnc)
			}
		})

	if err := lj.GetTranslateLocation().ProvideTargetPath(); err != nil {
		return err
	}

	return nil
}

func (lj *lockitFileJoin) JoinFileParts() error {
	importLocation := lj.GetImportLocation()

	if len(*lj.parts) != lj.expectedPartsLength {
		return fmt.Errorf("invalid number of parts: %d expected: %d", len(*lj.parts), lj.expectedPartsLength)
	}

	var combinedBuffer bytes.Buffer

	worker := common.NewWorker[LockitFileParts]()

	err := worker.ForEach(*lj.parts, func(_ int, part LockitFileParts) error {
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

	err = valideLockit(importLocation.TargetFile, combinedBuffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func valideLockit(file string, buffer []byte) error {
	originalData, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("erro ao ler o arquivo original: %v", err)
	}

	isExactMatch := bytes.Equal(originalData, buffer)
	if !isExactMatch {
		return fmt.Errorf("arquivos não correspondem")
	} else {
		fmt.Println("Arquivos lockkt correspondem")
	}

	return nil
}
