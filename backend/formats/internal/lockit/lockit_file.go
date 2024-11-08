package lockit

import (
	"ffxresources/backend/common"
	"ffxresources/backend/events"
	"ffxresources/backend/formats/internal/lockit/internal"
	"ffxresources/backend/formats/lib"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"fmt"
)

type LockitFile struct {
	DataInfo *interactions.GameDataInfo
	Parts    *[]internal.LockitFileParts
}

func NewLockitFile(dataInfo *interactions.GameDataInfo) *LockitFile {
	partsLength := interactions.NewInteraction().GamePartOptions.LockitPartsLength

	parts := make([]internal.LockitFileParts, 0, partsLength)

	gameFilesPath := interactions.NewInteraction().GameLocation.TargetDirectory

	relative := common.GetDifferencePath(dataInfo.GameData.AbsolutePath, gameFilesPath)

	dataInfo.GameData.RelativePath = relative

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	if err := lib.FindFileParts(&parts, dataInfo.ExtractLocation.TargetPath, common.LOCKIT_FILE_PARTS_PATTERN, internal.NewLockitFileParts); err != nil {
		events.NotifyError(err)
		return nil
	}

	return &LockitFile{
		DataInfo: dataInfo,
		Parts:    &parts,
	}
}

func (l *LockitFile) GetFileInfo() *interactions.GameDataInfo {
	return l.DataInfo
}

func (l *LockitFile) Extract() {
	currentPartsLength := len(*l.Parts)
	expectedPartsLength := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().LockitPartsLength

	if err := l.ensurePartsExists(currentPartsLength, expectedPartsLength); err != nil {
		events.NotifyError(err)
		return
	}

	worker := common.NewWorker[internal.LockitFileParts]()

	worker.ParallelForEach(*l.Parts,
		func(index int, part internal.LockitFileParts) {
			if index > 0 && index%2 == 0 {
				part.Extract(internal.LocEnc)
			} else {
				part.Extract(internal.FfxEnc)
			}
		})
}

func (l *LockitFile) Compress() {
	interactions := interactions.NewInteraction()

	lockitSizesLength := interactions.GamePartOptions.GetGamePartOptions().LockitPartsLength

	translatedParts := make([]internal.LockitFileParts, 0, lockitSizesLength)

	if err := lib.FindFileParts(&translatedParts, l.DataInfo.TranslateLocation.TargetPath, common.LOCKIT_TXT_PARTS_PATTERN, internal.NewLockitFileParts); err != nil {
		events.NotifyError(err)
		return
	}

	if len(translatedParts) != lockitSizesLength+1 {
		events.NotifyError(fmt.Errorf("invalid number of translated parts: %d expected: %d", len(translatedParts), lockitSizesLength))
		return
	}

	worker := common.NewWorker[internal.LockitFileParts]()

	worker.ParallelForEach(*l.Parts,
		func(index int, part internal.LockitFileParts) {
			if index > 0 && index%2 == 0 {
				part.Compress(internal.LocEnc)
			} else {
				part.Compress(internal.FfxEnc)
			}
		})

	lockitSizes := interactions.GamePartOptions.GetGamePartOptions().LockitPartsSizes

	if err := ffx2LockitJoiner(l.DataInfo, lockitSizes); err != nil {
		events.NotifyError(err)
		return
	}
}

func (l *LockitFile) ensurePartsExists(current, expected int) error {
	partsSizes := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().LockitPartsSizes
	if current != expected {
		if err := ffx2Xplitter(l.DataInfo, partsSizes); err != nil {
			return err
		}

		newLockitFile := NewLockitFile(l.DataInfo)
		l.DataInfo = newLockitFile.GetFileInfo()
		l.Parts = newLockitFile.Parts
	}

	return nil
}

func ffx2Xplitter(dataInfo *interactions.GameDataInfo, sizes []int) error {
	handler := internal.NewLockitFileXplit(dataInfo)

	extractLocation := dataInfo.ExtractLocation

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return err
	}

	if err := handler.XplitFile(sizes, common.LOCKIT_NAME_BASE, extractLocation.TargetPath); err != nil {
		return err
	}

	return nil
}

func ffx2LockitJoiner(dataInfo *interactions.GameDataInfo, sizes []int) error {
	joiner := internal.NewLockitFileJoin(dataInfo)

	if err := dataInfo.TranslateLocation.ProvideTargetPath(); err != nil {
		return err
	}

	if err := joiner.JoinFile(&sizes); err != nil {
		return err
	}

	return nil
}
