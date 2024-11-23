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

	fileVerifier internal.ILockitFileVerifier
	fileSplitter internal.IFileSplitter
	options      interactions.LockitFileOptions
	parts        *[]internal.LockitFileParts
	partsJoiner  internal.IPartsJoiner
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
		FormatsBase:  base.NewFormatsBase(dataInfo),
		fileVerifier: internal.NewLockitFileVerifier(dataInfo),
		fileSplitter: internal.NewLockitFileSplitter(),
		options:      interactions.NewInteraction().GamePartOptions.GetLockitFileOptions(),
		parts:        &parts,
		partsJoiner:  internal.NewLockitFileJoiner(dataInfo, parts),
	}
}

func (l *LockitFile) Extract() {
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			l.Log.Info().Msgf("Disposing target file: %s", l.GetFileInfo().GetImportLocation().TargetFile)

			l.GetFileInfo().GetImportLocation().DisposeTargetFile()

			close(errChan)
		}()

		for err := range errChan {
			fmt.Printf("Captured error: %s\n", err)
			l.Log.Error().Err(err).Msg("error when verifying monted lockit file")
			return
		}
	}()

	if len(*l.parts) != l.options.PartsLength {
		l.fileSplitter.FileSplitter(l.GetFileInfo(), l.options)

		newLockitFile := NewLockitFile(l.GetFileInfo()).(*LockitFile)

		l.SetFileInfo(newLockitFile.GetFileInfo())
		l.parts = newLockitFile.parts
	}

	l.fileSplitter.DecoderPartsFiles(l.parts)

	l.Log.Info().Msgf("Verifying splited lockit file: %s", l.GetExtractLocation().TargetPath)

	if err := l.fileVerifier.VerifyExtract(*l.parts, l.GetExtractLocation(), l.options); err != nil {
		l.Log.Error().Err(err).Send()
		return
	}

	l.Log.Info().Msgf("Lockit file extracted: %s", l.GetGameData().Name)
}

func (l *LockitFile) Compress() {
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			l.Log.Info().Msgf("Disposing target file: %s", l.GetFileInfo().GetImportLocation().TargetFile)

			l.GetFileInfo().GetImportLocation().DisposeTargetFile()

			close(errChan)
		}()

		for err := range errChan {
			fmt.Printf("Captured error: %s\n", err)
			l.Log.Error().Err(err).Msg("error when verifying monted lockit file")
			return
		}
	}()

	l.Log.Info().Msgf("Verifying splited parts before compressing: %s", l.GetExtractLocation().TargetPath)

	if err := l.fileVerifier.VerifyExtract(*l.parts, l.GetExtractLocation(), l.options); err != nil {
		l.Log.Error().Err(err).Send()
		return
	}

	translatedParts, err := l.partsJoiner.FindTranslatedTextParts()
	if err != nil {
		errChan <- err
		return
	}

	if len(translatedParts) != l.options.PartsLength {
		errChan <- fmt.Errorf("invalid number of translated parts")
		return
	}

	if err := l.partsJoiner.EncodeFilesParts(); err != nil {
		l.Log.Error().Err(err).Interface("LockitFile", l.GetFileInfo()).Msg("error when encoding lockit file parts")
		errChan <- err
		return
	}

	if err := l.partsJoiner.JoinFileParts(); err != nil {
		errChan <- err
		return
	}

	if err := l.fileVerifier.VerifyCompress(l.GetFileInfo(), l.options); err != nil {
		errChan <- err
		return
	}

	l.Log.Info().Msgf("Lockit file monted: %s", l.GetGameData().Name)
}
