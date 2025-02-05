package lockit

import (
	"ffxresources/backend/core"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/verify"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
)

type LockitFile struct {
	*base.FormatsBase

	source         interfaces.ISource
	destination    locations.IDestination
	fileCompressor ILockitFileCompressor
	fileExtractor  ILockitFileExtractor

	lockitPartsIntegrity verify.ILockitFilePartsIntegrity
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
	lockitEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	
	lf.lockitPartsIntegrity = verify.NewLockitFilePartsIntegrity(lf.log)
	defer lockitEncoding.Dispose()
	defer lf.lockitPartsIntegrity.Dispose()

	
	lf.fileExtractor = newLockitFileExtractor(lf.source, lf.destination, lockitEncoding, lf.log)
	
	lf.log.LogInfo("Extracting lockit file in path: %s", lf.destination.Extract().Get().GetTargetPath())
	
	if err := lf.fileExtractor.Extract(); err != nil {
		return err
	}
	
	lf.log.LogInfo("Verifying lockit file parts in path: %s", lf.destination.Extract().Get().GetTargetPath())
	
	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber()
	fileOptions := core.NewLockitFileOptions(gameVersion)

	//TODO: Implement this function
	if err := lf.lockitPartsIntegrity.ValidatePartsLineBreaksCount(nil, fileOptions); err != nil {
		return err
	}

	lf.log.LogInfo("Lockit file extracted successfully in path: %s", lf.destination.Extract().Get().GetTargetPath())

	return nil
}

func (lf *LockitFile) Compress() error {
	lockitEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()

	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber()
	fileOptions := core.NewLockitFileOptions(gameVersion)

	lf.fileCompressor = newLockitFileCompressor(lf.source, lf.destination, lockitEncoding, fileOptions, lf.log)

	defer lockitEncoding.Dispose()
	defer lf.dispose()

	lf.lockitFileIntegrity = NewLockitFileIntegrity(lf.log)
	lf.lockitPartsIntegrity = verify.NewLockitFilePartsIntegrity(lf.log)

	lf.log.LogInfo("Compressing lockit file in path: %s", lf.destination.Translate().Get().GetTargetPath())

	if err := lf.fileCompressor.Compress(); err != nil {
		return err
	}

	lf.log.LogInfo("Verifying lockit file parts in path: %s", lf.destination.Translate().Get().GetTargetPath())

	if err := lf.lockitFileIntegrity.ValidateFileLineBreaksCount(lf.Destination(), fileOptions); err != nil {
		return err
	}

	if err := lf.lockitFileIntegrity.VerifyFileIntegrity(lf.Destination().Import().Get().GetTargetFile(), lockitEncoding, fileOptions); err != nil {
		return err
	}

	lf.log.LogInfo("Translated Lockit file monted successfully!")

	return nil
}

func (lf *LockitFile) dispose() {
	if lf.fileCompressor != nil {
		lf.fileCompressor.Dispose()
		lf.fileCompressor = nil
	}

	if lf.fileExtractor != nil {
		lf.fileExtractor = nil
	}

	if lf.lockitFileIntegrity != nil {
		lf.lockitFileIntegrity = nil
	}

	if lf.lockitPartsIntegrity != nil {
		lf.lockitPartsIntegrity.Dispose()
		lf.lockitPartsIntegrity = nil
	}
}
