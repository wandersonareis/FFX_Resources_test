package integrity

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/logger"
	"ffxresources/backend/models"
)

type (
	ILockitFilePartsIntegrity interface {
		ComparePartsLineBreaksCount(fileList components.IList[string], lockitFileOptions core.ILockitFileOptions) error
		ComparePartsContent(partsList components.IList[models.FileComparisonEntry]) error
	}
	LockitFilePartsIntegrity struct {
		lockitFileLineBreaksCounter ILineBreakCounter
		lockitFileContentComparer   IComparerContent
		log                         logger.ILoggerHandler
	}
)

func NewLockitFilePartsIntegrity(logger logger.ILoggerHandler) ILockitFilePartsIntegrity {
	return &LockitFilePartsIntegrity{
		lockitFileLineBreaksCounter: NewLineBreakCounter(logger),
		lockitFileContentComparer:   NewComparerContent(logger),

		log: logger,
	}
}

func (lfpi *LockitFilePartsIntegrity) ComparePartsLineBreaksCount(fileList components.IList[string], lockitFileOptions core.ILockitFileOptions) error {
	if err := lfpi.lockitFileLineBreaksCounter.VerifyLineBreaks(fileList, lockitFileOptions); err != nil {
		return err
	}

	return nil
}

func (lfpi *LockitFilePartsIntegrity) ComparePartsContent(partsList components.IList[models.FileComparisonEntry]) error {
	if err := lfpi.lockitFileContentComparer.CompareContent(partsList); err != nil {
		return err
	}

	return nil
}
