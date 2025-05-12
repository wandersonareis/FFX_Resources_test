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
	"ffxresources/backend/loggingService"
)

type LockitFile struct {
	baseFormats.IBaseFileFormat

	source                      interfaces.ISource
	destination                 locations.IDestination
	lockitFileExtractIntegrity  integrity.ILockitFileExtractorIntegrity
	lockitFileCompressIntegrity integrity.ILockitFileCompressorIntegrity
	options                     core.ILockitFileOptions
	log                         loggingService.ILoggerService
}

func NewLockitFile(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	destination.CreateRelativePath(source, interactions.NewInteractionService().GameLocation.GetTargetDirectory())

	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	options := core.NewLockitFileOptions(interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber())

	return &LockitFile{
		IBaseFileFormat: baseFormats.NewFormatsBase(source, destination),
		lockitFileExtractIntegrity: integrity.NewLockitFileExtractorIntegrity(
			options,
			loggingService.NewLoggerHandler("lockit_file_integrity"),
		),
		source:      source,
		destination: destination,
		options:     options,
		log:         loggingService.NewLoggerHandler("lockit_file"),
	}
}

func (lf *LockitFile) Extract() error {
	lockitEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextUTF8Encoding()
	defer lockitEncoding.Dispose()

	if err := lf.extract(lockitEncoding); err != nil {
		return err
	}

	if err := lf.extractVerify(); err != nil {
		return err
	}

	lf.log.Info("Lockit file extracted successfully in path: %s", lf.destination.Extract().GetTargetPath())

	return nil
}

func (lf *LockitFile) extract(lockitEncoding ffxencoding.IFFXTextLockitEncoding) error {
	lf.log.Info("Extracting lockit file inside path: %s", lf.destination.Extract().GetTargetPath())

	fileExtractor := NewLockitFileExtractor(lf.source, lf.destination, lockitEncoding, lf.options, lf.log)

	return fileExtractor.Extract()
}

func (lf *LockitFile) extractVerify() error {
	lf.log.Info("Verifying lockit file parts in path: %s", lf.destination.Extract().GetTargetPath())

	if err := lf.lockitFileExtractIntegrity.Verify(lf.GetDestination().Extract().GetTargetPath()); err != nil {
		return err
	}

	lf.log.Info("Lockit file parts verified successfully in path: %s", lf.destination.Extract().GetTargetPath())
	return nil
}

func (lf *LockitFile) Compress() error {
	lockitEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextUTF8Encoding()
	defer lockitEncoding.Dispose()

	if err := lf.compress(lockitEncoding); err != nil {
		return err
	}

	if err := lf.compressVerify(lockitEncoding); err != nil {
		return err
	}

	lf.log.Info("Translated Lockit file monted successfully!")

	return nil
}

func (lf *LockitFile) compress(lockitEncoding ffxencoding.IFFXTextLockitEncoding) error {
	lf.log.Info("Compressing lockit file in path: %s", lf.destination.Translate().GetTargetPath())

	fileCompressor := NewLockitFileCompressor(lf.source, lf.destination, lockitEncoding, lf.options, lf.log)

	if err := fileCompressor.Compress(); err != nil {
		return err
	}

	lf.log.Info("Lockit file compressed successfully in path: %s", lf.destination.Translate().GetTargetPath())

	return nil
}

func (lf *LockitFile) compressVerify(lockitEncoding ffxencoding.IFFXTextLockitEncoding) error {
	lf.log.Info("Verifying translated lockit file in path: %s", lf.destination.Translate().GetTargetPath())

	if err := lf.lockitFileCompressIntegrity.Verify(lf.GetDestination(), lockitEncoding, lf.options); err != nil {
		return err
	}

	lf.log.Info("Translated lockit file verified successfully")

	return nil
}
