package lockit

import (
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/joiner"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/fileFormats/internal/lockit/internal/splitter"
	"ffxresources/backend/fileFormats/internal/lockit/internal/verify"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"
)

type LockitFile struct {
	*base.FormatsBase

	fileVerifier verify.ILockitFileVerifier
	fileSplitter splitter.IFileSplitter
	options      interactions.LockitFileOptions
	parts        *[]parts.LockitFileParts
	partsJoiner  joiner.IPartsJoiner
}

func NewLockitFile(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	partsList := &[]parts.LockitFileParts{}

	dataInfo.CreateRelativePath()

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	if err := util.FindFileParts(partsList,
		dataInfo.GetExtractLocation().TargetPath,
		lib.LOCKIT_FILE_PARTS_PATTERN,
		parts.NewLockitFileParts); err != nil {
		l := logger.Get()
		l.Error().Err(err).Msg("error when finding lockit parts")
		return nil
	}

	return &LockitFile{
		FormatsBase:  base.NewFormatsBase(dataInfo),
		fileVerifier: verify.NewLockitFileVerifier(dataInfo),
		fileSplitter: splitter.NewLockitFileSplitter(),
		options:      interactions.NewInteraction().GamePartOptions.GetLockitFileOptions(),
		parts:        partsList,
		partsJoiner:  joiner.NewLockitFileJoiner(dataInfo, partsList),
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
			l.Log.Error().Err(err).Msg("error when verifying monted lockit file")

			return
		}
	}()

	l.Log.Info().Msg("Extracting lockit file parts...")
	l.Log.Info().Msgf("Parts found: %d", len(*l.parts))

	if len(*l.parts) != l.options.PartsLength {
		l.Log.Info().Msg("Ensuring splited lockit parts...")

		l.fileSplitter.FileSplitter(l.GetFileInfo(), l.options)

		newLockitFile := NewLockitFile(l.GetFileInfo()).(*LockitFile)

		l.SetFileInfo(newLockitFile.GetFileInfo())
		l.parts = newLockitFile.parts
	}

	l.Log.Info().Msg("Decoding lockit file parts...")

	l.fileSplitter.DecoderPartsFiles(l.parts)

	l.Log.Info().Msgf("Verifying splited lockit file: %s", l.GetExtractLocation().TargetPath)

	if err := l.fileVerifier.VerifyExtract(l.parts, l.GetExtractLocation(), l.options); err != nil {
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
			l.Log.Error().Err(err).Msg("error when verifying monted lockit file")

			return
		}
	}()

	l.Log.Info().Msg("Compressing lockit file parts...")
	l.Log.Info().Msgf("Verifying splited parts before compressing: %s", l.GetExtractLocation().TargetPath)
	
	if err := l.fileVerifier.VerifyExtract(l.parts, l.GetExtractLocation(), l.options); err != nil {
		l.Log.Error().Err(err).Send()
		return
	}

	l.Log.Info().Msgf("Finding translated text parts on: %s", l.GetTranslateLocation().TargetPath)

	translatedParts, err := l.partsJoiner.FindTranslatedTextParts()
	if err != nil {
		errChan <- err
		return
	}

	if len(*translatedParts) != l.options.PartsLength {
		errChan <- fmt.Errorf("invalid number of translated parts")
		return
	}

	l.Log.Info().Msgf("Parts found: %d", len(*translatedParts))

	l.Log.Info().Msgf("Encoding files parts to: %s", l.GetImportLocation().TargetPath)

	if err := l.partsJoiner.EncodeFilesParts(); err != nil {
		l.Log.Error().Err(err).Interface("LockitFile", l.GetFileInfo()).Msg("error when encoding lockit file parts")
		errChan <- err
		return
	}

	l.Log.Info().Msgf("Joining file parts to: %s", l.GetImportLocation().TargetFile)

	if err := l.partsJoiner.JoinFileParts(); err != nil {
		errChan <- err
		return
	}

	l.Log.Info().Msgf("Verifying monted lockit file: %s", l.GetImportLocation().TargetFile)

	if err := l.fileVerifier.VerifyCompress(l.GetFileInfo(), l.options); err != nil {
		errChan <- err
		return
	}

	l.Log.Info().Msgf("Lockit file monted: %s", l.GetGameData().Name)
}
