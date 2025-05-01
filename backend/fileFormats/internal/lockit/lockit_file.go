package lockit

import (
	"ffxresources/backend/core"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/baseFormats"
	"ffxresources/backend/fileFormats/internal/lockit/internal/integrity"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
)

type LockitFile struct {
	baseFormats.IBaseFileFormat

	source      interfaces.ISource
	destination locations.IDestination
	log         logger.ILoggerHandler
}

func NewLockitFile(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	destination.CreateRelativePath(source, interactions.NewInteractionService().GameLocation.GetTargetDirectory())

	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	return &LockitFile{
		IBaseFileFormat: baseFormats.NewFormatsBase(source, destination),
		log:             logger.NewLoggerHandler("lockit_file"),
		source:          source,
		destination:     destination,
	}
}

func (lf *LockitFile) Extract() error {
	lockitEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	defer lockitEncoding.Dispose()

	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber()
	fileOptions := core.NewLockitFileOptions(gameVersion)

	if err := lf.extract(lockitEncoding, fileOptions); err != nil {
		return err
	}

	if err := lf.extractVerify(fileOptions); err != nil {
		return err
	}

	lf.log.LogInfo("Lockit file extracted successfully in path: %s", lf.destination.Extract().GetTargetPath())

	return nil
}

func (lf *LockitFile) extract(lockitEncoding ffxencoding.IFFXTextLockitEncoding, fileOptions core.ILockitFileOptions) error {
	lf.log.LogInfo("Extracting lockit file inside path: %s", lf.destination.Extract().GetTargetPath())

	fileExtractor := NewLockitFileExtractor(lf.source, lf.destination, lockitEncoding, fileOptions, lf.log)

	return fileExtractor.Extract()
}

func (lf *LockitFile) extractVerify(fileOptions core.ILockitFileOptions) error {
	lf.log.LogInfo("Verifying lockit file parts in path: %s", lf.destination.Extract().GetTargetPath())

	extractVerifier := integrity.NewLockitFileExtractorIntegrity(lf.log)

	return extractVerifier.Verify(lf.GetDestination().Extract().GetTargetPath(), fileOptions)
}

func (lf *LockitFile) Compress() error {
	lockitEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	defer lockitEncoding.Dispose()

	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber()
	fileOptions := core.NewLockitFileOptions(gameVersion)

	if err := lf.compress(lockitEncoding, fileOptions); err != nil {
		return err
	}

	if err := lf.compressVerify(lockitEncoding, fileOptions); err != nil {
		return err
	}

	lf.log.LogInfo("Translated Lockit file monted successfully!")

	return nil
}

func (lf *LockitFile) compress(lockitEncoding ffxencoding.IFFXTextLockitEncoding, fileOptions core.ILockitFileOptions) error {
	lf.log.LogInfo("Compressing lockit file in path: %s", lf.destination.Translate().Get().GetTargetPath())

	fileCompressor := NewLockitFileCompressor(lf.source, lf.destination, lockitEncoding, fileOptions, lf.log)

	return fileCompressor.Compress()
}

func (lf *LockitFile) compressVerify(lockitEncoding ffxencoding.IFFXTextLockitEncoding, fileOptions core.ILockitFileOptions) error {
	lf.log.LogInfo("Verifying translated lockit file in path: %s", lf.destination.Translate().Get().GetTargetPath())

	compressVerify := integrity.NewLockitFileIntegrity(lf.log)

	return compressVerify.Verify(lf.GetDestination(), lockitEncoding, fileOptions)
}
