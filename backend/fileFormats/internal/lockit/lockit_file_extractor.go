package lockit

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/baseFormats"
	"ffxresources/backend/fileFormats/internal/lockit/internal"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"fmt"
)

type (
	ILockitFileExtractor interface {
		Extract() error
	}

	LockitFileExtractor struct {
		baseFormats.IBaseFileFormat

		filePartsDecoder lockitParts.ILockitFilePartsDecoder
		lockitEncoding   ffxencoding.IFFXTextLockitEncoding
		options          core.ILockitFileOptions

		log loggingService.ILoggerService
	}
)

func NewLockitFileExtractor(
	source interfaces.ISource,
	destination locations.IDestination,
	lockitEncoding ffxencoding.IFFXTextLockitEncoding,
	fileOptions core.ILockitFileOptions,
	logger loggingService.ILoggerService) *LockitFileExtractor {
	common.CheckArgumentNil(source, "source")
	common.CheckArgumentNil(destination, "destination")
	common.CheckArgumentNil(lockitEncoding, "lockitEncoding")
	common.CheckArgumentNil(fileOptions, "fileOptions")
	common.CheckArgumentNil(logger, "logger")
	return &LockitFileExtractor{
		IBaseFileFormat:  baseFormats.NewFormatsBase(source, destination),
		filePartsDecoder: lockitParts.NewLockitFilePartsDecoder(),
		lockitEncoding:   lockitEncoding,

		options: fileOptions,
		log:     logger,
	}
}

func (lfe *LockitFileExtractor) Extract() error {
	partsLength := lfe.options.GetPartsLength()

	partsList := components.NewList[lockitParts.LockitFileParts](partsLength)
	defer partsList.Clear()

	if err := lfe.populateLockitBinaryFileParts(partsList); err != nil {
		return err
	}

	if err := lfe.ensureAllLockitBinaryFileParts(partsList, partsLength); err != nil {
		return err
	}

	if err := lfe.decodeFileParts(partsList); err != nil {
		return err
	}

	lfe.log.Info("Lockit file extracted: %s", lfe.GetDestination().Extract().GetTargetPath())

	return nil
}

func (lfe *LockitFileExtractor) populateLockitBinaryFileParts(partsList components.IList[lockitParts.LockitFileParts]) error {
	lfe.log.Info("Populating lockit binary file parts...")
	return lockitParts.PopulateLockitBinaryFileParts(
		partsList,
		lfe.GetDestination().Extract().GetTargetPath(),
	)
}

func (lfe *LockitFileExtractor) ensureAllLockitBinaryFileParts(partsList components.IList[lockitParts.LockitFileParts], partsLength int) error {
	lfe.log.Info("Ensuring all lockit binary file parts...")

	if partsList.GetLength() == partsLength {
		return nil
	}

	if err := lfe.extractMissingLockitFileParts(); err != nil {
		return err
	}

	if err := lfe.populateLockitBinaryFileParts(partsList); err != nil {
		return err
	}

	if partsList.GetLength() != partsLength {
		return fmt.Errorf("error ensuring splitted lockit parts: expected %d, got %d on path: %s",
			partsLength, partsList.GetLength(), lfe.GetDestination().Extract().GetTargetPath())
	}

	return nil
}

func (lfe *LockitFileExtractor) extractMissingLockitFileParts() error {
	lfe.log.Info("Missing lockit file parts detected. Attempting to extract...")

	splitter := internal.NewLockitFileSplitter()
	return splitter.FileSplitter(lfe.GetSource(), lfe.GetDestination().Extract(), lfe.options)
}

func (lfe *LockitFileExtractor) decodeFileParts(partsList components.IList[lockitParts.LockitFileParts]) error {
	if partsList.IsEmpty() {
		return fmt.Errorf("error ensuring lockit parts: list is empty")
	}

	lfe.log.Info("Decoding lockit file parts...")

	// TODO: Implement a way to get the game version from the source file
	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()
	filePartsDecoder := lockitParts.NewLockitFilePartsDecoder()
	if err := filePartsDecoder.DecodeFileParts(partsList, lfe.lockitEncoding, gameVersion); err != nil {
		lfe.log.Error(err, "failed to decode lockit file parts")
		return fmt.Errorf("failed to decode lockit file: %s", lfe.GetSource().GetName())
	}

	return nil
}
