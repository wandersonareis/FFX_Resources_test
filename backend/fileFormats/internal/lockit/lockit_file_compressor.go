package lockit

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/joiner"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitFileParts"
	"ffxresources/backend/fileFormats/internal/lockit/internal/verify"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type lockitFileCompressor struct {
	*base.FormatsBase
	logger.ILoggerHandler

	fileVerifier    verify.ILockitFileVerifier
	lockitFileParts components.IList[lockitFileParts.LockitFileParts]
	options         interactions.LockitFileOptions
	partsEncoder    lockitFileParts.ILockitFilePartsEncoder
	partsJoiner     joiner.IPartsJoiner
}

func newLockitFileCompressor(source interfaces.ISource, destination locations.IDestination, lockitFilePartsList components.IList[lockitFileParts.LockitFileParts]) *lockitFileCompressor {
	return &lockitFileCompressor{
		FormatsBase: base.NewFormatsBase(source, destination),
		fileVerifier: verify.NewLockitFileVerifier(source, destination),
		lockitFileParts: lockitFilePartsList,
		options: interactions.NewInteractionService().DcpAndLockitOptions.GetLockitFileOptions(),
		partsEncoder: lockitFileParts.NewLockitFilePartsEncoder(lockitFilePartsList),
		partsJoiner: joiner.NewLockitFileJoiner(source, destination, lockitFilePartsList),
		ILoggerHandler: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "lockit_file_compressor").Logger(),
		},
	}
}

func (l *lockitFileCompressor) Compress() error {
	l.LogInfo("Compressing lockit file parts...")
	l.LogInfo("Verifying splited parts before compressing: %s", l.Destination().Extract().Get().GetTargetPath())

	if err := l.fileVerifier.VerifyExtract(l.lockitFileParts, l.options); err != nil {
		l.LogError(err, "failed to verify lockit file parts: %s", l.Source().Get().Name)
		return fmt.Errorf("failed to verify lockit file parts: %s", l.Source().Get().Name)
	}

	l.LogInfo("Finding translated text parts on: %s", l.Destination().Translate().Get().GetTargetPath())

	translatedParts, err := l.partsJoiner.FindTranslatedTextParts()
	if err != nil {
		l.LogError(err, "error when finding translated text parts")
		return fmt.Errorf("failed to find translated text parts: %s", l.Source().Get().Name)
	}

	if translatedParts.GetLength() != l.options.PartsLength {
		l.LogError(nil, "invalid number of translated parts")
		return fmt.Errorf("invalid number of translated parts: %s", l.Source().Get().Name)
	}

	l.LogInfo("Parts found: %d", translatedParts.GetLength())
	l.LogInfo("Encoding files parts to: %s", l.Destination().Import().Get().GetTargetPath())

	l.partsEncoder.EncodeFilesParts()

	l.LogInfo("Joining file parts to: %s", l.Destination().Import().Get().GetTargetFile())
	if err := l.partsJoiner.JoinFileParts(); err != nil {
		l.LogError(err, "error when joining lockit file parts")
		return fmt.Errorf("failed to join lockit file parts: %s", l.Source().Get().Name)
	}

	l.LogInfo("Verifying monted lockit file: %s", l.Destination().Import().Get().GetTargetFile())
	if err := l.fileVerifier.VerifyCompress(l.Destination(), l.options); err != nil {
		l.LogError(err, "error when verifying monted lockit file")
		return fmt.Errorf("failed to verify monted lockit file: %s", l.Source().Get().Name)
	}

	l.LogInfo("Lockit file monted: %s", l.Source().Get().Name)
	return nil
}