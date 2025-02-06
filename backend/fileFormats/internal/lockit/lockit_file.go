package lockit

import (
	"ffxresources/backend/core"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
)

type LockitFile struct {
	*base.FormatsBase

	source         interfaces.ISource
	destination    locations.IDestination
	//fileCompressor ILockitFileCompressor
	//fileExtractor  ILockitFileExtractor
	fileOptions    core.ILockitFileOptions

	//lockitPartsIntegrity verify.ILockitFilePartsIntegrity
	lockitFileIntegrity  ILockitFileIntregrity

	log logger.ILoggerHandler
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
	//lockitEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()

	//lf.lockitPartsIntegrity = verify.NewLockitFilePartsIntegrity(lf.log)
	//defer lockitEncoding.Dispose()
	//defer lf.lockitPartsIntegrity.Dispose()

	//lf.fileExtractor = newLockitFileExtractor(lf.source, lf.destination, lockitEncoding, lf.log)

	//lf.log.LogInfo("Extracting lockit file in path: %s", lf.destination.Extract().Get().GetTargetPath())

	/* if err := lf.fileExtractor.Extract(); err != nil {
		return err
	} */

	if err := lf.extract(); err != nil {
		return err
	}

	if err := lf.extractVerify(); err != nil {
		return err
	}

	/* if err := lf.lockitPartsIntegrity.ValidatePartsLineBreaksCount(nil, fileOptions); err != nil {
		return err
	} */

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

	fileExtractor := newLockitFileExtractor(lf.source, lf.destination, lockitEncoding, lf.log)

	return fileExtractor.Extract()
}

func (lf *LockitFile) extractVerify() error {
	lf.log.LogInfo("Verifying lockit file parts in path: %s", lf.destination.Extract().Get().GetTargetPath())

	extractVerifier := NewLockitFileExtractorIntegrity(lf.log)
	defer extractVerifier.Dispose()

	lf.ensureFileOptions()

	return extractVerifier.VerifyFileIntegrity(lf.Destination(), lf.fileOptions)
}

func (lf *LockitFile) Compress() error {
	lockitEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	defer lockitEncoding.Dispose()

	/* gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber()
	fileOptions := core.NewLockitFileOptions(gameVersion) */

	//lf.ensureFileOptions()

	//lf.fileCompressor = newLockitFileCompressor(lf.source, lf.destination, lockitEncoding, lf.fileOptions, lf.log)


	lf.lockitFileIntegrity = NewLockitFileIntegrity(lf.log)
	//lf.lockitPartsIntegrity = verify.NewLockitFilePartsIntegrity(lf.log)

	//lf.log.LogInfo("Compressing lockit file in path: %s", lf.destination.Translate().Get().GetTargetPath())

	/* if err := lf.fileCompressor.Compress(); err != nil {
		return err
	} */

	if err := lf.compress(lockitEncoding); err != nil {
		return err
	}

	lf.log.LogInfo("Verifying lockit file parts in path: %s", lf.destination.Translate().Get().GetTargetPath())

	if err := lf.lockitFileIntegrity.ValidateFileLineBreaksCount(lf.Destination(), lf.fileOptions); err != nil {
		return err
	}

	if err := lf.lockitFileIntegrity.VerifyFileIntegrity(lf.Destination().Import().Get().GetTargetFile(), lockitEncoding, lf.fileOptions); err != nil {
		return err
	}

	lf.log.LogInfo("Translated Lockit file monted successfully!")

	return nil
}

func (lf *LockitFile) compress(lockitEncoding ffxencoding.IFFXTextLockitEncoding) error {
	lf.log.LogInfo("Compressing lockit file in path: %s", lf.destination.Translate().Get().GetTargetPath())

	lf.ensureFileOptions()

	fileCompressor := newLockitFileCompressor(lf.source, lf.destination, lockitEncoding, lf.fileOptions, lf.log)
	defer fileCompressor.Dispose()

	return fileCompressor.Compress()
}

func (lf *LockitFile) Dispose() {
	/* if lf.fileCompressor != nil {
		lf.fileCompressor.Dispose()
		lf.fileCompressor = nil
	} */

	/* if lf.fileExtractor != nil {
		lf.fileExtractor = nil
	} */

	if lf.lockitFileIntegrity != nil {
		lf.lockitFileIntegrity = nil
	}

	/* if lf.lockitPartsIntegrity != nil {
		lf.lockitPartsIntegrity.Dispose()
		lf.lockitPartsIntegrity = nil
	} */
}
