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
	parts   *[]LockitFileParts
	options *interactions.LockitFileOptions
}

func NewLockitFileJoiner(dataInfo interactions.IGameDataInfo, parts *[]LockitFileParts) *lockitFileJoin {
	return &lockitFileJoin{
		FormatsBase: base.NewFormatsBase(dataInfo),
		options:     interactions.NewInteraction().GamePartOptions.GetLockitFileOptions(),
		parts:       parts,
	}
}

func (lj *lockitFileJoin) FindTextParts() ([]LockitFileParts, error) {
	parts := []LockitFileParts{}
	err := util.FindFileParts(
		&parts,
		lj.GetTranslateLocation().TargetPath,
		LOCKIT_TXT_PARTS_PATTERN,
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

	if len(*lj.parts) != lj.options.PartsLength {
		return fmt.Errorf("invalid number of parts: %d expected: %d", len(*lj.parts), lj.options.PartsLength)
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

	err = lj.valideLockit(lj.GetGameData().FullFilePath, combinedBuffer.Bytes())
	if err != nil {
		return err
	}

	if err := os.WriteFile(importLocation.TargetFile, combinedBuffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("error when creating output file: %v", err)
	}

	return nil
}

func (lj *lockitFileJoin) countAllLineEndings(buffer []byte) int {
	return bytes.Count(buffer, []byte("\r\n"))
}

func (lj *lockitFileJoin) valideLockit(file string, buffer []byte) error {
	bufferLineBreaksCount := lj.countAllLineEndings(buffer)

	if lj.options.LineBreaksCount != bufferLineBreaksCount {
		return fmt.Errorf("line breaks count does not match. Expected: %d, got: %d", lj.options.LineBreaksCount, bufferLineBreaksCount)
	}

	originalData, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("erro ao ler o arquivo original: %v", err)
	}

	isExactMatch := bytes.Equal(originalData, buffer)
	if !isExactMatch {
		return fmt.Errorf("arquivos não correspondem")
	} else {
		fmt.Println("Arquivos lockit correspondem")
	}

	return nil
}
