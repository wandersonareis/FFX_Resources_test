package verify

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/logger"
)

type (
	ILockitFilePartsIntegrity interface {
		ValidatePartsLineBreaksCount(fileList components.IList[string], lockitFileOptions core.ILockitFileOptions) error
		ComparePartsContent(partsList components.IList[FileComparisonEntry]) error
		Dispose()
	}
	LockitFilePartsIntegrity struct {
		filePartsLineBreakCounter ILineBreakCounter
		filePartsCompareContent   IComparerContent

		log logger.ILoggerHandler
	}
)

func NewLockitFilePartsIntegrity(logger logger.ILoggerHandler) ILockitFilePartsIntegrity {
	return &LockitFilePartsIntegrity{
		log: logger,
	}
}

func (lfpi *LockitFilePartsIntegrity) ValidatePartsLineBreaksCount(fileList components.IList[string], lockitFileOptions core.ILockitFileOptions) error {
	lfpi.filePartsLineBreakCounter = NewLineBreakCounter(lfpi.log)

	return lfpi.filePartsLineBreakCounter.VerifyLineBreaks(fileList, lockitFileOptions)
}

func (lfpi *LockitFilePartsIntegrity) ComparePartsContent(partsList components.IList[FileComparisonEntry]) error {
	lfpi.filePartsCompareContent = NewComparerContent(lfpi.log)

	return lfpi.filePartsCompareContent.CompareContent(partsList)
}

func (lfpi *LockitFilePartsIntegrity) Dispose() {
	if lfpi.filePartsLineBreakCounter != nil {
		lfpi.filePartsLineBreakCounter = nil
	}

	if lfpi.filePartsCompareContent != nil {
		lfpi.filePartsCompareContent = nil
	}
}