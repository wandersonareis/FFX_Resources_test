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
	Parts *[]internal.LockitFileParts
}

func NewLockitFile(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	parts := []internal.LockitFileParts{}

	dataInfo.CreateRelativePath()

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	if err := util.FindFileParts(&parts, dataInfo.GetExtractLocation().TargetPath, util.LOCKIT_FILE_PARTS_PATTERN, internal.NewLockitFileParts); err != nil {
		l := logger.Get()
		l.Error().Err(err).Msg("error when finding lockit parts")
		return nil
	}

	return &LockitFile{
		FormatsBase: base.NewFormatsBase(dataInfo),
		Parts:       &parts,
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

		l.SetFileInfo(newLockitFile.GetFileInfo())
		l.Parts = newLockitFile.Parts
	}

	internal.SegmentFile(l.Parts)
}

func (l *LockitFile) Compress() {
	interactions := interactions.NewInteraction()

	lockitSizesLength := interactions.GamePartOptions.GetGamePartOptions().LockitPartsLength + 1

	partsJoiner := internal.NewLockitFileJoiner(l.GetFileInfo(), l.Parts)

	translatedParts, err := partsJoiner.FindTextParts()
	if err != nil {
		l.Log.Error().Err(err).Msg("error when finding lockit text parts")
		return
	}

	if len(translatedParts) != lockitSizesLength {
		l.Log.Error().Int("TranslatedParts", len(translatedParts)).Int("ExpectedParts", lockitSizesLength).Msg("invalid number of translated parts")
		return
	}

	if err := partsJoiner.EncodeFilesParts(); err != nil {
		l.Log.Error().Err(err).Interface("LockitFile", l.GetFileInfo()).Msg("error when encoding lockit file parts")
		return
	}

	if err := partsJoiner.JoinFileParts(); err != nil {
		l.Log.Error().Err(err).Interface("LockitFile", l.GetFileInfo()).Msg("error when joining lockit file parts")
		return
	}
}
