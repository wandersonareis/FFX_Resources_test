package lockit

import (
	"ffxresources/backend/common"
	"ffxresources/backend/formats/internal/lockit/internal"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
	"sync"
)

type LockitFile struct {
	dataInfo *interactions.GameDataInfo
	Parts    *[]internal.LockitFileParts
}

var ffxLockitSizes = []int{}
var ffx2LockitSizes = []int{80, 88, 90, 93, 94, 95, 102, 1223, 1224, 1230, 1232, 1233, 1240, 1241, 1502, 1534}

func NewLockitFile(dataInfo *interactions.GameDataInfo) *LockitFile {
	partsLength := interactions.NewInteraction().GamePartOptions.LockitPartsLength

	parts := make([]internal.LockitFileParts, 0, partsLength)

	gameFilesPath := interactions.NewInteraction().GameLocation.TargetDirectory

	relative := common.GetDifferencePath(dataInfo.GameData.AbsolutePath, gameFilesPath)
	
	dataInfo.GameData.RelativePath = relative

	dataInfo.ExtractLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)

	if err := internal.FindLockitParts(&parts, dataInfo.ExtractLocation.TargetPath, common.LOCKIT_FILE_PARTS_PATTERN); err != nil {
		lib.NotifyError(err)
		return nil
	}

	return &LockitFile{
		dataInfo: dataInfo,
		Parts:    &parts,
	}
}

func (l *LockitFile) GetFileInfo() *interactions.GameDataInfo {
	return l.dataInfo
}

func (l *LockitFile) Extract() {
	if len(*l.Parts) != len(getLockitFileSizes()) {
		if err := ffx2Xplitter(l.dataInfo); err != nil {
			lib.NotifyError(err)
			return
		}

		newLockitFile := NewLockitFile(l.dataInfo)
		l.dataInfo = newLockitFile.GetFileInfo()
		l.Parts = newLockitFile.Parts
	}

	var wg sync.WaitGroup

	for i, part := range *l.Parts {
		wg.Add(1)

		go func(index int, extractor *internal.LockitFileParts) {
			defer wg.Done()

			if index > 0 && index%2 == 0 {
				extractor.Extract(internal.LocEnc)
			} else {
				extractor.Extract(internal.FfxEnc)
			}
		}(i, &part)
	}

	wg.Wait()
}

func (l *LockitFile) Compress() {
	sizes := getLockitFileSizes()
	translatedParts := make([]internal.LockitFileParts, 0, len(sizes))

	if err := internal.FindLockitParts(&translatedParts, l.dataInfo.TranslateLocation.TargetPath, common.LOCKIT_TXT_PARTS_PATTERN); err != nil {
		lib.NotifyError(err)
		return
	}

	if len(translatedParts) != len(sizes)+1 {
		lib.NotifyError(fmt.Errorf("invalid number of translated parts: %d expected: %d", len(translatedParts), len(sizes)))
		return
	}

	var wg sync.WaitGroup

	for index, part := range *l.Parts {
		wg.Add(1)

		go func(index int, compressor *internal.LockitFileParts) {
			defer wg.Done()

			if index > 0 && index%2 == 0 {
				compressor.Compress(internal.LocEnc)
			} else {
				compressor.Compress(internal.FfxEnc)
			}
		}(index, &part)
	}

	wg.Wait()

	if err := ffx2LockitJoiner(l.dataInfo); err != nil {
		lib.NotifyError(err)
		return
	}
}

func ffx2Xplitter(dataInfo *interactions.GameDataInfo) error {
	handler := internal.NewLockitFileXplit(dataInfo)

	if err := dataInfo.ExtractLocation.ProvideTargetPath(); err != nil {
		return err
	}

	if err := handler.XplitFile(ffx2LockitSizes, common.LOCKIT_NAME_BASE, dataInfo.ExtractLocation.TargetPath); err != nil {
		return err
	}

	return nil
}

func ffx2LockitJoiner(dataInfo *interactions.GameDataInfo) error {
	joiner := internal.NewLockitFileJoin(dataInfo)

	if err := dataInfo.TranslateLocation.ProvideTargetPath(); err != nil {
		return err
	}

	sizes := getLockitFileSizes()

	if err := joiner.JoinFile(&sizes); err != nil {
		return err
	}

	return nil
}

func getLockitFileSizes() []int {
	gamePart := interactions.NewInteraction().GamePart.GetGamePart()

	if gamePart == interactions.Ffx {
		return ffxLockitSizes
	}

	if gamePart == interactions.Ffx2 {
		return ffx2LockitSizes
	}

	return nil
}
