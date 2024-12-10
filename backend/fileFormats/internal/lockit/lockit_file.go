package lockit

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/joiner"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/fileFormats/internal/lockit/internal/splitter"
	"ffxresources/backend/fileFormats/internal/lockit/internal/verify"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"

	"github.com/rs/zerolog"
)

type LockitFile struct {
	*base.FormatsBase

	fileVerifier verify.ILockitFileVerifier
	fileSplitter splitter.IFileSplitter
	options      interactions.LockitFileOptions
	parts        components.IList[parts.LockitFileParts]
	partsJoiner  joiner.IPartsJoiner

	log zerolog.Logger
}

func NewLockitFile(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	partsList := components.NewEmptyList[parts.LockitFileParts]()

	dataInfo.CreateRelativePath()

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	err := components.GenerateGameFileParts(
		partsList,
		dataInfo.GetExtractLocation().TargetPath,
		lib.LOCKIT_FILE_PARTS_PATTERN,
		parts.NewLockitFileParts)

	if err != nil {
		l := logger.Get()
		l.Error().Err(err).Msg("error when finding lockit parts")
		return nil
	}

	return &LockitFile{
		FormatsBase:  base.NewFormatsBase(dataInfo),
		fileVerifier: verify.NewLockitFileVerifier(dataInfo),
		fileSplitter: splitter.NewLockitFileSplitter(),
		options:      interactions.NewInteraction().GamePartOptions.GetLockitFileOptions(),
		partsJoiner:  joiner.NewLockitFileJoiner(dataInfo, partsList),
		parts:        partsList,
		log:          logger.Get().With().Str("module", "lockit_file").Logger(),
	}
}

func (l *LockitFile) Extract() error {
	l.log.Info().Msgf("Verifying lockit file parts in path: %s", l.GetExtractLocation().TargetPath)

	l.log.Info().Msgf("Parts found: %d", l.parts.GetLength())

	if l.parts.GetLength() != l.options.PartsLength {
		l.log.Info().Msg("Extracting lockit file parts...")

		l.fileSplitter.FileSplitter(l.GetFileInfo(), l.options)

		newLockitFile := NewLockitFile(l.GetFileInfo()).(*LockitFile)

		l.SetFileInfo(newLockitFile.GetFileInfo())
		l.parts = newLockitFile.parts
	}

	l.log.Info().Msg("Decoding lockit file parts...")

	l.fileSplitter.DecoderPartsFiles(l.parts)

	l.log.Info().Msgf("Verifying lockit file parts: %s", l.GetExtractLocation().TargetPath)

	if err := l.fileVerifier.VerifyExtract(l.parts, l.GetExtractLocation(), l.options); err != nil {
		l.log.Error().Err(err).Send()
		return fmt.Errorf("failed to extract lockit file: %s", l.GetGameData().Name)
	}

	l.log.Info().Msgf("Lockit file extracted: %s", l.GetGameData().Name)

	return nil
}

func (l *LockitFile) Compress() error {
	l.log.Info().Msg("Compressing lockit file parts...")
	l.log.Info().Msgf("Verifying splited parts before compressing: %s", l.GetExtractLocation().TargetPath)

	if err := l.fileVerifier.VerifyExtract(l.parts, l.GetExtractLocation(), l.options); err != nil {
		l.log.Error().
			Err(err).
			Send()
		return fmt.Errorf("failed to verify lockit file parts: %s", l.GetGameData().Name)
	}

	l.log.Info().Msgf("Finding translated text parts on: %s", l.GetTranslateLocation().TargetPath)

	translatedParts, err := l.partsJoiner.FindTranslatedTextParts()
	if err != nil {
		l.log.Error().
			Err(err).
			Msg("error when finding translated text parts")
		return fmt.Errorf("failed to find translated text parts: %s", l.GetGameData().Name)
	}

	if translatedParts.GetLength() != l.options.PartsLength {
		l.log.Error().Msg("invalid number of translated parts")
		return fmt.Errorf("invalid number of translated parts: %s", l.GetGameData().Name)
	}

	l.log.Info().Msgf("Parts found: %d", translatedParts.GetLength())

	l.log.Info().Msgf("Encoding files parts to: %s", l.GetImportLocation().TargetPath)

	if err := l.partsJoiner.EncodeFilesParts(); err != nil {
		l.log.Error().
			Err(err).
			Msg("error when encoding lockit file parts")
		return fmt.Errorf("failed to encode lockit file parts: %s", l.GetGameData().Name)
	}

	l.log.Info().Msgf("Joining file parts to: %s", l.GetImportLocation().TargetFile)

	if err := l.partsJoiner.JoinFileParts(); err != nil {
		l.log.Error().
			Err(err).
			Msg("error when joining lockit file parts")
		return fmt.Errorf("failed to join lockit file parts: %s", l.GetGameData().Name)
	}

	l.log.Info().Msgf("Verifying monted lockit file: %s", l.GetImportLocation().TargetFile)

	if err := l.fileVerifier.VerifyCompress(l.GetFileInfo(), l.options); err != nil {
		l.log.Error().
			Err(err).
			Msg("error when verifying monted lockit file")
		return fmt.Errorf("failed to verify monted lockit file: %s", l.GetGameData().Name)
	}

	l.log.Info().
		Msgf("Lockit file monted: %s", l.GetGameData().Name)

	return nil
}
