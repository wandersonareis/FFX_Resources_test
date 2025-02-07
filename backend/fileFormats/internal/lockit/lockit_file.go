package lockit

import (
	"ffxresources/backend/core"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/integrity"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
)

type LockitFile struct {
	*base.FormatsBase

	source      interfaces.ISource
	destination locations.IDestination
	fileOptions core.ILockitFileOptions
	log         logger.ILoggerHandler
}

func NewLockitFile(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	destination.CreateRelativePath(source, interactions.NewInteractionService().GameLocation.GetTargetDirectory())

	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	return &LockitFile{
		FormatsBase: base.NewFormatsBase(source, destination),
		log:         logger.NewLoggerHandler("lockit_file"),
		source:      source,
		destination: destination,
	}
}

func (lf *LockitFile) Extract() error {
	if err := lf.extract(); err != nil {
		return err
	}

	if err := lf.extractVerify(); err != nil {
		return err
	}

	lf.log.LogInfo("Lockit file extracted successfully in path: %s", lf.destination.Extract().Get().GetTargetPath())

	return nil
}

func (lf *LockitFile) ensureFileOptions() {
	if lf.fileOptions == nil {
		lf.fileOptions = core.NewLockitFileOptions(interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber())
	}
}

func (lf *LockitFile) extract() error {
	lf.log.LogInfo("Extracting lockit file inside path: %s", lf.destination.Extract().Get().GetTargetPath())

	lockitEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	defer lockitEncoding.Dispose()

	lf.ensureFileOptions()

	fileExtractor := newLockitFileExtractor(lf.source, lf.destination, lockitEncoding, lf.log)

	return fileExtractor.Extract()
}

func (lf *LockitFile) extractVerify() error {
	lf.log.LogInfo("Verifying lockit file parts in path: %s", lf.destination.Extract().Get().GetTargetPath())

	extractVerifier := integrity.NewLockitFileExtractorIntegrity(lf.log)

	return extractVerifier.VerifyFileIntegrity(lf.Destination(), lf.fileOptions)
}

func (lf *LockitFile) Compress() error {
	lockitEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	defer lockitEncoding.Dispose()

	lf.ensureFileOptions()

	if err := lf.compress(lockitEncoding); err != nil {
		return err
	}

	if err := lf.compressVerify(lockitEncoding); err != nil {
		return err
	}

	lf.log.LogInfo("Translated Lockit file monted successfully!")

	return nil
}

func (lf *LockitFile) compress(lockitEncoding ffxencoding.IFFXTextLockitEncoding) error {
	lf.log.LogInfo("Compressing lockit file in path: %s", lf.destination.Translate().Get().GetTargetPath())

	fileCompressor := newLockitFileCompressor(lf.source, lf.destination, lockitEncoding, lf.fileOptions, lf.log)
	defer fileCompressor.Dispose()

	return fileCompressor.Compress()
}

func (lf *LockitFile) compressVerify(lockitEncoding ffxencoding.IFFXTextLockitEncoding) error {
	lf.log.LogInfo("Verifying translated lockit file in path: %s", lf.destination.Translate().Get().GetTargetPath())

	compressVerify := integrity.NewLockitFileCompressorIntegrity(lf.log)

	return compressVerify.VerifyFileIntegrity(lf.Destination(), lockitEncoding, lf.fileOptions)
}
