package integrity

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/logger"
)

type (
	ILockitFilePartsIntegrity interface {
		ValidatePartsLineBreaksCount(fileList components.IList[string], lockitFileOptions core.ILockitFileOptions) error
		ComparePartsContent(partsList components.IList[FileComparisonEntry]) error
	}
	LockitFilePartsIntegrity struct {
		log logger.ILoggerHandler
	}
)

func NewLockitFilePartsIntegrity(logger logger.ILoggerHandler) ILockitFilePartsIntegrity {
	return &LockitFilePartsIntegrity{log: logger}
}

func (lfpi *LockitFilePartsIntegrity) ValidatePartsLineBreaksCount(fileList components.IList[string], lockitFileOptions core.ILockitFileOptions) error {
	filePartsLineBreakCounter := NewLineBreakCounter(lfpi.log)

	return filePartsLineBreakCounter.VerifyLineBreaks(fileList, lockitFileOptions)
}

func (lfpi *LockitFilePartsIntegrity) ComparePartsContent(partsList components.IList[FileComparisonEntry]) error {
	filePartsCompareContent := NewComparerContent(lfpi.log)

	return filePartsCompareContent.CompareContent(partsList)
}
