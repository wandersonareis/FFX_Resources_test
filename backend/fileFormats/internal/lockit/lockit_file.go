package lockit

import (
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"
)

type LockitFile struct {
	*base.FormatsBase
	internal.ILockitFileVerifier

	options *interactions.LockitFileOptions
	Parts   *[]internal.LockitFileParts
}

func NewLockitFile(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	parts := []internal.LockitFileParts{}

	dataInfo.CreateRelativePath()

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	if err := util.FindFileParts(&parts,
		dataInfo.GetExtractLocation().TargetPath,
		internal.LOCKIT_FILE_PARTS_PATTERN,
		internal.NewLockitFileParts); err != nil {
		l := logger.Get()
		l.Error().Err(err).Msg("error when finding lockit parts")
		return nil
	}

	return &LockitFile{
		FormatsBase:         base.NewFormatsBase(dataInfo),
		ILockitFileVerifier: internal.NewLockitFileVerifier(dataInfo),
		options:             interactions.NewInteraction().GamePartOptions.GetLockitFileOptions(),
		Parts:               &parts,
	}
}

func (l *LockitFile) Extract() {
	errChan := make(chan error, 1)

	go func() {
		for err := range errChan {
			fmt.Printf("Captured error: %s\n", err)
			l.Log.Error().Err(err).Msg("error when verifying monted lockit file")
			l.GetFileInfo().GetImportLocation().DisposeTargetFile()
			return
		}
	}()

	if len(*l.Parts) != l.options.PartsLength {
		internal.FileSplitter(l.GetFileInfo(), *l.options)

		newLockitFile := NewLockitFile(l.GetFileInfo()).(*LockitFile)

		l.SetFileInfo(newLockitFile.GetFileInfo())
		l.Parts = newLockitFile.Parts
	}

	internal.DecoderPartsFiles(l.Parts)

	if err := l.VerifyExtract(*l.Parts, l.GetExtractLocation(), *l.options); err != nil {
		l.Log.Error().Err(err).Send()
		return
	}

	l.Log.Info().Msgf("Lockit file extracted: %s", l.GetGameData().Name)
}

func (l *LockitFile) Compress() {
	errChan := make(chan error, 1)

	go func() {
		for err := range errChan {
			fmt.Printf("Captured error: %s\n", err)
			l.Log.Error().Err(err).Msg("error when verifying monted lockit file")
			l.GetFileInfo().GetImportLocation().DisposeTargetFile()
			return
		}
	}()

	partsJoiner := internal.NewLockitFileJoiner(l.GetFileInfo(), *l.Parts)

	translatedParts, err := partsJoiner.FindTranslatedTextParts()
	if err != nil {
		errChan <- err
		return
	}

	if len(translatedParts) != l.options.PartsLength {
		errChan <- fmt.Errorf("invalid number of translated parts")
		return
	}

	if err := partsJoiner.EncodeFilesParts(); err != nil {
		l.Log.Error().Err(err).Interface("LockitFile", l.GetFileInfo()).Msg("error when encoding lockit file parts")
		errChan <- err
		return
	}

	if err := partsJoiner.JoinFileParts(); err != nil {
		errChan <- err
		return
	}

	if err := l.VerifyCompress(l.GetFileInfo(), l.options); err != nil {
		errChan <- err
		return
	}

	l.Log.Info().Msgf("Lockit file monted: %s", l.GetGameData().Name)
}
