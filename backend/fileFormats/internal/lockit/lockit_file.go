package lockit

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/joiner"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/fileFormats/internal/lockit/internal/splitter"
	"ffxresources/backend/fileFormats/internal/lockit/internal/verify"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
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

func NewLockitFile(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	partsList := components.NewEmptyList[parts.LockitFileParts]()

	destination.CreateRelativePath(source, interactions.NewInteraction().GameLocation.GetTargetDirectory())

	destination.InitializeLocations(source, formatters.NewTxtFormatterDev())

	err := components.GenerateGameFilePartsDev(
		partsList,
		destination.Extract().Get().GetTargetPath(),
		lib.LOCKIT_FILE_PARTS_PATTERN,
		parts.NewLockitFileParts)

	if err != nil {
		l := logger.Get()
		l.Error().Err(err).Msg("error when finding lockit parts")
		return nil
	}

	return &LockitFile{
		FormatsBase:  base.NewFormatsBaseDev(source, destination),
		fileVerifier: verify.NewLockitFileVerifier(source, destination),
		fileSplitter: splitter.NewLockitFileSplitter(),
		options:      interactions.NewInteraction().DcpAndLockitOptions.GetLockitFileOptions(),
		partsJoiner:  joiner.NewLockitFileJoiner(source, destination, partsList),
		parts:        partsList,
		log:          logger.Get().With().Str("module", "lockit_file").Logger(),
	}
}

func (l *LockitFile) Extract() error {
	l.log.Info().Msgf("Verifying lockit file parts in path: %s", l.Destination().Extract().Get().GetTargetPath())

	l.log.Info().Msgf("Parts found: %d", l.parts.GetLength())

	if l.parts.GetLength() != l.options.PartsLength {
		l.log.Info().Msg("Extracting lockit file parts...")

		l.fileSplitter.FileSplitter(l.Source(), l.Destination().Extract().Get(), l.options)

		newLockitFile := NewLockitFile(l.Source(), l.Destination()).(*LockitFile)

		l.SetFileInfo(newLockitFile.GetFileInfo())
		l.parts = newLockitFile.parts
	}

	l.log.Info().Msg("Decoding lockit file parts...")

	l.fileSplitter.DecoderPartsFiles(l.parts)

	l.log.Info().Msgf("Verifying lockit file parts: %s", l.Destination().Extract().Get().GetTargetPath())

	if err := l.fileVerifier.VerifyExtract(l.parts, l.options); err != nil {
		l.log.Error().Err(err).Send()
		return fmt.Errorf("failed to extract lockit file: %s", l.Source().Get().Name)
	}

	l.log.Info().Msgf("Lockit file extracted: %s", l.Source().Get().Name)

	return nil
}

func (l *LockitFile) Compress() error {
	l.log.Info().Msg("Compressing lockit file parts...")
	l.log.Info().Msgf("Verifying splited parts before compressing: %s", l.Destination().Extract().Get().GetTargetPath())

	if err := l.fileVerifier.VerifyExtract(l.parts, l.options); err != nil {
		l.log.Error().
			Err(err).
			Send()
		return fmt.Errorf("failed to verify lockit file parts: %s", l.Source().Get().Name)
	}

	l.log.Info().Msgf("Finding translated text parts on: %s", l.Destination().Translate().Get().GetTargetPath())

	translatedParts, err := l.partsJoiner.FindTranslatedTextParts()
	if err != nil {
		l.log.Error().
			Err(err).
			Msg("error when finding translated text parts")
		return fmt.Errorf("failed to find translated text parts: %s", l.Source().Get().Name)
	}

	if translatedParts.GetLength() != l.options.PartsLength {
		l.log.Error().Msg("invalid number of translated parts")
		return fmt.Errorf("invalid number of translated parts: %s", l.Source().Get().Name)
	}

	l.log.Info().Msgf("Parts found: %d", translatedParts.GetLength())

	l.log.Info().Msgf("Encoding files parts to: %s", l.Destination().Import().Get().GetTargetPath())

	if err := l.partsJoiner.EncodeFilesParts(); err != nil {
		l.log.Error().
			Err(err).
			Msg("error when encoding lockit file parts")
		return fmt.Errorf("failed to encode lockit file parts: %s", l.Source().Get().Name)
	}

	l.log.Info().Msgf("Joining file parts to: %s", l.Destination().Import().Get().GetTargetFile())

	if err := l.partsJoiner.JoinFileParts(); err != nil {
		l.log.Error().
			Err(err).
			Msg("error when joining lockit file parts")
		return fmt.Errorf("failed to join lockit file parts: %s", l.Source().Get().Name)
	}

	l.log.Info().Msgf("Verifying monted lockit file: %s", l.Destination().Import().Get().GetTargetFile())

	if err := l.fileVerifier.VerifyCompress(l.Destination(), l.options); err != nil {
		l.log.Error().
			Err(err).
			Msg("error when verifying monted lockit file")
		return fmt.Errorf("failed to verify monted lockit file: %s", l.Source().Get().Name)
	}

	l.log.Info().
		Msgf("Lockit file monted: %s", l.Source().Get().Name)

	return nil
}
