package lockit

import (
	"ffxresources/backend/events"
	"ffxresources/backend/fileFormats/internal/lockit/internal"
	"ffxresources/backend/fileFormats/lib"
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

	dataInfo.CreateRelativePath()

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	if err := lib.FindFileParts(&parts, dataInfo.ExtractLocation.TargetPath, lib.LOCKIT_FILE_PARTS_PATTERN, internal.NewLockitFileParts); err != nil {
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
	expectedPartsLength := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().LockitPartsLength + 1

	if currentPartsLength != expectedPartsLength {
		if err := internal.EnsurePartsExists(l.DataInfo); err != nil {
			events.NotifyError(err)
			return
		}

		newLockitFile := NewLockitFile(l.DataInfo)
		l.DataInfo = newLockitFile.GetFileInfo()
		l.Parts = newLockitFile.Parts
	}

	internal.SegmentFile(l.Parts)
}

func (l *LockitFile) Compress() {
	interactions := interactions.NewInteraction()

	lockitSizesLength := interactions.GamePartOptions.GetGamePartOptions().LockitPartsLength

	translatedParts := make([]internal.LockitFileParts, 0, lockitSizesLength)

	if err := lib.FindFileParts(&translatedParts, l.DataInfo.TranslateLocation.TargetPath, lib.LOCKIT_TXT_PARTS_PATTERN, internal.NewLockitFileParts); err != nil {
		events.NotifyError(err)
		return
	}

	if len(translatedParts) != lockitSizesLength+1 {
		events.NotifyError(fmt.Errorf("invalid number of translated parts: %d expected: %d", len(translatedParts), lockitSizesLength))
		return
	}

	partsJoiner := internal.NewLockitFileJoiner(l.DataInfo, l.Parts)

	if err := partsJoiner.UnSegmenterFile(); err != nil {
		events.NotifyError(err)
		return
	}
}
