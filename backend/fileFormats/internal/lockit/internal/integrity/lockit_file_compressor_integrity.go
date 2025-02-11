package integrity

import (
	"ffxresources/backend/core"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
)

type (
	ILockitFileCompressorIntegrity interface {
		VerifyFileIntegrity(
			destination locations.IDestination,
			lockitEncoding ffxencoding.IFFXTextLockitEncoding,
			fileOptions core.ILockitFileOptions) error
	}

	lockitFileCompressorIntegrity struct {
		formatter interfaces.ITextFormatter
		log       logger.ILoggerHandler
	}
)

func NewLockitFileCompressorIntegrity(logger logger.ILoggerHandler) ILockitFileCompressorIntegrity {
	return &lockitFileCompressorIntegrity{
		formatter: interactions.NewInteractionService().TextFormatter(),
		log: logger}
}

func (lfci *lockitFileCompressorIntegrity) VerifyFileIntegrity(
	destination locations.IDestination,
	lockitEncoding ffxencoding.IFFXTextLockitEncoding,
	fileOptions core.ILockitFileOptions) error {

	return lfci.verify(destination, lockitEncoding, fileOptions)
}

func (lfci *lockitFileCompressorIntegrity) verify(
	destination locations.IDestination,
	lockitEncoding ffxencoding.IFFXTextLockitEncoding,
	fileOptions core.ILockitFileOptions) error {
	lockitFileIntegrity := newLockitFileIntegrity(lfci.log)

	if err := lockitFileIntegrity.ValidateFileLineBreaksCount(destination, fileOptions); err != nil {
		return err
	}

	return lockitFileIntegrity.VerifyFileIntegrity(destination.Import().Get().GetTargetFile(), lockitEncoding, fileOptions)
}
