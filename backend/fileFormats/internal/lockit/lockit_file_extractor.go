package lockit

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type (
	ILockitFileExtractor interface {
		Extract() error
	}

	LockitFileExtractor struct {
		*base.FormatsBase

		filePartsList    components.IList[lockitParts.LockitFileParts]
		filePartsDecoder lockitParts.ILockitFilePartsDecoder
		lockitEncoding   ffxencoding.IFFXTextLockitEncoding
		options          core.ILockitFileOptions

		log logger.ILoggerHandler
	}
)

func newLockitFileExtractor(
	source interfaces.ISource,
	destination locations.IDestination,
	lockitEncoding ffxencoding.IFFXTextLockitEncoding,
	logger logger.ILoggerHandler) *LockitFileExtractor {
	return &LockitFileExtractor{
		FormatsBase:      base.NewFormatsBase(source, destination),
		filePartsDecoder: lockitParts.NewLockitFilePartsDecoder(),
		lockitEncoding:   lockitEncoding,
		options:          core.NewLockitFileOptions(interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber()),

		log: logger,
	}
}

func (l *LockitFileExtractor) Extract() error {
	partsLength := l.options.GetPartsLength()

	l.initializeFilePartsList(partsLength)
	defer l.filePartsList.Clear()

	if err := l.populateLockitBinaryFileParts(); err != nil {
		return err
	}

	if err := l.ensureAllLockitBinaryFileParts(partsLength); err != nil {
		return err
	}

	if err := l.decodeFileParts(); err != nil {
		return err
	}

	l.log.LogInfo("Lockit file extracted: %s", l.GetDestination().Extract().Get().GetTargetPath())

	return nil
}

func (l *LockitFileExtractor) initializeFilePartsList(partsLength int) {
	l.filePartsList = components.NewList[lockitParts.LockitFileParts](partsLength)
}

func (l *LockitFileExtractor) populateLockitBinaryFileParts() error {
	return lockitParts.PopulateLockitBinaryFileParts(
		l.filePartsList,
		l.GetDestination().Extract().Get().GetTargetPath(),
	)
}

func (l *LockitFileExtractor) ensureAllLockitBinaryFileParts(partsLength int) error {
	if l.filePartsList.GetLength() == partsLength {
		return nil
	}

	l.log.LogInfo("Missing lockit file parts detected. Attempting to extract...")

	if err := l.extractMissingLockitFileParts(); err != nil {
		return err
	}

	if err := l.populateLockitBinaryFileParts(); err != nil {
		return err
	}

	if l.filePartsList.GetLength() != partsLength {
		return fmt.Errorf("error ensuring splitted lockit parts: expected %d, got %d",
			partsLength, l.filePartsList.GetLength())
	}

	return nil
}

func (l *LockitFileExtractor) extractMissingLockitFileParts() error {
	splitter := internal.NewLockitFileSplitter()
	return splitter.FileSplitter(l.GetSource(), l.GetDestination().Extract().Get(), l.options)
}

func (l *LockitFileExtractor) decodeFileParts() error {
	l.log.LogInfo("Decoding lockit file parts...")

	if err := l.filePartsDecoder.DecodeFileParts(l.filePartsList, l.lockitEncoding); err != nil {
		return fmt.Errorf("failed to decode lockit file: %s", l.GetSource().Get().Name)
	}

	return nil
}
