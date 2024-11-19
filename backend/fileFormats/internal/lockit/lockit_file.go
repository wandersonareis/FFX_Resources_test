package lockit

import (
	"ffxresources/backend/fileFormats/internal/lockit/internal"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
)

type LockitFile struct {
	*internal.LockitFileVerify

	options *interactions.LockitFileOptions
	Parts   *[]internal.LockitFileParts
}

func NewLockitFile(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	parts := []internal.LockitFileParts{}

	dataInfo.CreateRelativePath()

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	if err := util.FindFileParts(&parts, dataInfo.GetExtractLocation().TargetPath, internal.LOCKIT_FILE_PARTS_PATTERN, internal.NewLockitFileParts); err != nil {
		l := logger.Get()
		l.Error().Err(err).Msg("error when finding lockit parts")
		return nil
	}

	return &LockitFile{
		LockitFileVerify: internal.NewLockitFileVerify(dataInfo),
		options:     interactions.NewInteraction().GamePartOptions.GetLockitFileOptions(),
		Parts:       &parts,
	}
}

func (l *LockitFile) Extract() {
	xplitter := internal.NewLockitFileXplitter(l.GetFileInfo())

	if len(*l.Parts) != l.options.PartsLength {
		if err := xplitter.EnsurePartsExists(); err != nil {
			l.Log.Error().Err(err).Msg("error when ensuring lockit parts exist")
			return
		}

		newLockitFile := NewLockitFile(l.GetFileInfo()).(*LockitFile)

		l.SetFileInfo(newLockitFile.GetFileInfo())
		l.Parts = newLockitFile.Parts
	}

	xplitter.SegmentFile(l.Parts)

	l.VerifyExtract(l.GetExtractLocation().TargetPath, l.options)
}

func (l *LockitFile) Compress() {
	partsJoiner := internal.NewLockitFileJoiner(l.GetFileInfo(), l.Parts)

	translatedParts, err := partsJoiner.FindTextParts()
	if err != nil {
		l.Log.Error().Err(err).Msg("error when finding lockit text parts")
		return
	}

	if len(translatedParts) != l.options.PartsLength {
		l.Log.Error().Int("TranslatedParts", len(translatedParts)).Int("ExpectedParts", l.options.PartsLength).Msg("invalid number of translated parts")
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

	if err := l.VerifyCompress(l.GetFileInfo(), l.options); err != nil {
		l.Log.Error().Err(err).Send()
	}
}
