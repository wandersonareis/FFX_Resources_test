package lockit

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitFileParts"
	"ffxresources/backend/fileFormats/internal/lockit/internal/splitter"
	"ffxresources/backend/fileFormats/internal/lockit/internal/verify"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type LockitFileExtractor struct {
	*base.FormatsBase
	logger.ILoggerHandler

	filePartsVerifier verify.ILockitFileVerifier
	filePartsSplitter splitter.IFileSplitter
	filePartsList     components.IList[lockitFileParts.LockitFileParts]
	filePartsDecoder  lockitFileParts.ILockitFilePartsDecoder
	options           interactions.LockitFileOptions
}

func newLockitFileExtractor(source interfaces.ISource, destination locations.IDestination, partsList components.IList[lockitFileParts.LockitFileParts], ) *LockitFileExtractor {
	return &LockitFileExtractor{
		FormatsBase:       base.NewFormatsBase(source, destination),
		ILoggerHandler:    &logger.LogHandler{Logger: logger.Get().With().Str("module", "lockit_file_extractor").Logger()},
		filePartsVerifier: verify.NewLockitFileVerifier(source, destination),
		filePartsSplitter: splitter.NewLockitFileSplitter(),
		filePartsList:     partsList,
		options: interactions.NewInteractionService().DcpAndLockitOptions.GetLockitFileOptions(),
	}
}

func (l *LockitFileExtractor) Extract() error {
	l.LogInfo("Verifying lockit file parts in path: %s", l.Destination().Extract().Get().GetTargetPath())

	if l.filePartsList.GetLength() != l.options.PartsLength {
		l.LogInfo("Extracting lockit file parts...")

		l.filePartsSplitter.FileSplitter(l.Source(), l.Destination().Extract().Get(), l.options)

		newLockitFile := NewLockitFile(l.Source(), l.Destination()).(*LockitFile)

		l.SetFileInfo(newLockitFile.GetFileInfo())
		l.filePartsList = newLockitFile.lockitFileParts
		l.filePartsDecoder = lockitFileParts.NewLockitFilePartsDecoder(l.filePartsList)
	}

	l.LogInfo("Decoding lockit file parts...")

	if err := l.filePartsDecoder.DecodeFileParts(); err != nil {
		l.LogError(err, "failed to decode lockit file parts: %s", l.Source().Get().Name)
		return fmt.Errorf("failed to decode lockit file: %s", l.Source().Get().Name)
	}

	l.LogInfo("Verifying lockit file parts: %s", l.Destination().Extract().Get().GetTargetPath())

	if err := l.filePartsVerifier.VerifyExtract(l.filePartsList, l.options); err != nil {
		l.LogError(err, "failed to verify lockit file parts: %s", l.Source().Get().Name)
		return fmt.Errorf("failed to extract lockit file: %s", l.Source().Get().Name)
	}

	l.LogInfo("Lockit file: %s", l.Source().Get().Name)
	l.LogInfo("Lockit file extracted: %s", l.Destination().Extract().Get().GetTargetPath())

	return nil
}
