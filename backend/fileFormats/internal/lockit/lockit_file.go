package lockit

import (
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
)

type LockitFile struct {
	*base.FormatsBase
	//DataInfo interactions.IGameDataInfo
	Parts    *[]internal.LockitFileParts
}

func NewLockitFile(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	partsLength := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().LockitPartsLength

	parts := make([]internal.LockitFileParts, 0, partsLength)

	dataInfo.CreateRelativePath()

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	if err := util.FindFileParts(&parts, dataInfo.GetExtractLocation().TargetPath, util.LOCKIT_FILE_PARTS_PATTERN, internal.NewLockitFileParts); err != nil {
		l := logger.Get()
		l.Error().Err(err).Msg("error when finding lockit parts")
		return nil
	}

	return &LockitFile{
		FormatsBase: base.NewFormatsBase(dataInfo),
		//DataInfo: dataInfo.GetGameDataInfo(),
		Parts:    &parts,
	}
}

func (l *LockitFile) Extract() {
	currentPartsLength := len(*l.Parts)
	expectedPartsLength := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().LockitPartsLength + 1

	if currentPartsLength != expectedPartsLength {
		if err := internal.EnsurePartsExists(l.GetFileInfo()); err != nil {
			l.Log.Error().Err(err).Msg("error when ensuring lockit parts exist")
			return
		}

		newLockitFile := NewLockitFile(l.GetFileInfo()).(*LockitFile)

		//l.DataInfo = newLockitFile.GetFileInfo()
		l.SetFileInfo(newLockitFile.GetFileInfo())
		l.Parts = newLockitFile.Parts
	}

	internal.SegmentFile(l.Parts)
}

func (l *LockitFile) Compress() {
	interactions := interactions.NewInteraction()

	lockitSizesLength := interactions.GamePartOptions.GetGamePartOptions().LockitPartsLength

	translatedParts := make([]internal.LockitFileParts, 0, lockitSizesLength)

	if err := util.FindFileParts(
		&translatedParts,
		l.GetTranslateLocation().TargetPath,
		util.LOCKIT_TXT_PARTS_PATTERN,
		internal.NewLockitFileParts); err != nil {
		l.Log.Error().Err(err).Str("Path", l.GetTranslateLocation().TargetPath).Msg("error when finding lockit parts")
		return
	}

	if len(translatedParts) != lockitSizesLength+1 {
		l.Log.Error().Int("TranslatedParts", len(translatedParts)).Int("ExpectedParts", lockitSizesLength).Msg("invalid number of translated parts")
		return
	}

	partsJoiner := internal.NewLockitFileJoiner(l.GetFileInfo(), l.Parts)

	if err := partsJoiner.UnSegmenterFile(); err != nil {
		l.Log.Error().Err(err).Interface("LockitFile", l.GetFileInfo()).Msg("error when unsegmenting lockit file")
		return
	}
}
