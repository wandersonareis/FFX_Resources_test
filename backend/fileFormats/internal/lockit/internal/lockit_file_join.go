package internal

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
)

type LockitFileJoin struct {
	*base.FormatsBase
	//dataInfo            interactions.IGameDataInfo
	parts               *[]LockitFileParts
	partsSizes          *[]int
	expectedPartsLength int
}

func NewLockitFileJoiner(dataInfo interactions.IGameDataInfo, parts *[]LockitFileParts) *LockitFileJoin {
	lockitSizes := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().LockitPartsSizes
	return &LockitFileJoin{
		FormatsBase:         base.NewFormatsBase(dataInfo),
		//dataInfo:            dataInfo,
		parts:               parts,
		partsSizes:          &lockitSizes,
		expectedPartsLength: len(lockitSizes) + 1,
	}
}

func (lj *LockitFileJoin) UnSegmenterFile() error {
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

	if err := lj.joinFile(); err != nil {
		return err
	}

	return nil
}

func (lj *LockitFileJoin) joinFile() error {
	importLocation := lj.GetImportLocation()

	if len(*lj.parts) != lj.expectedPartsLength {
		return fmt.Errorf("invalid number of parts: %d expected: %d", len(*lj.parts), lj.expectedPartsLength)
	}

	var combinedBuffer bytes.Buffer

	worker := common.NewWorker[LockitFileParts]()

	err := worker.ForEach(*lj.parts, func(_ int, part LockitFileParts) error {
		fileName := part.dataInfo.GetImportLocation().TargetFile

		partData, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("erro ao ler a parte %s: %v", fileName, err)
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
		return fmt.Errorf("erro ao criar arquivo de saída: %v", err)
	}

	originalData, err := os.ReadFile(lj.GetGameData().FullFilePath)
	if err != nil {
		return fmt.Errorf("erro ao ler o arquivo original: %v", err)
	}

	isExactMatch := bytes.Equal(originalData, combinedBuffer.Bytes())
	if !isExactMatch {
		return fmt.Errorf("arquivos não correspondem")
	} else {
		fmt.Println("Arquivos lockkt correspondem")
	}

	return nil
}
