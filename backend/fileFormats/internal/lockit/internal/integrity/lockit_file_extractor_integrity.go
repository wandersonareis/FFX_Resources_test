package integrity

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/logger"
)

type (
	ILockitFileExtractorIntegrity interface {
		Verify(targetPath string, fileOptions core.ILockitFileOptions) error
	}

	lockitFileExtractorIntegrity struct {
		logger logger.ILoggerHandler
	}
)

func NewLockitFileExtractorIntegrity(logger logger.ILoggerHandler) ILockitFileExtractorIntegrity {
	return &lockitFileExtractorIntegrity{logger: logger}
}

func (lfei *lockitFileExtractorIntegrity) Verify(targetPath string, fileOptions core.ILockitFileOptions) error {
	extractedLockitBinaryList := components.NewList[lockitParts.LockitFileParts](fileOptions.GetPartsLength())
	defer extractedLockitBinaryList.Clear()

	if err := lfei.populateLockitExtractedBinaryFileParts(targetPath, extractedLockitBinaryList); err != nil {
		return err
	}

	if err := lfei.ensureAllLockitExtractedBinaryFileParts(extractedLockitBinaryList, fileOptions); err != nil {
		return err
	}

	extractedLockitTextList := components.NewList[lockitParts.LockitFileParts](fileOptions.GetPartsLength())
	defer extractedLockitTextList.Clear()

	if err := lfei.populateLockitExtractedTextFileParts(targetPath, extractedLockitTextList); err != nil {
		return err
	}

	if err := lfei.ensureAllLockitExtractedTextFileParts(extractedLockitTextList, fileOptions); err != nil {
		return err
	}

	return nil
}

func (lfei *lockitFileExtractorIntegrity) populateLockitExtractedBinaryFileParts(targetPath string, lockitBinaryList components.IList[lockitParts.LockitFileParts]) error {
	return lockitParts.PopulateLockitBinaryFileParts(lockitBinaryList, targetPath)
}

func (lfei *lockitFileExtractorIntegrity) populateLockitExtractedTextFileParts(targetPath string, lockitTextList components.IList[lockitParts.LockitFileParts]) error {
	return lockitParts.PopulateLockitTextFileParts(lockitTextList, targetPath)
}

func (lfei *lockitFileExtractorIntegrity) ensureAllLockitExtractedBinaryFileParts(lockitBinaryPartsList components.IList[lockitParts.LockitFileParts], fileOptions core.ILockitFileOptions) error {
	partsLength := fileOptions.GetPartsLength()

	if err := lfei.ensurePartsListCount(lockitBinaryPartsList, partsLength); err != nil {
		return err
	}

	lockitBinaryPaths, err := lfei.createPartsPathsList(lockitBinaryPartsList, fileOptions)
	defer lockitBinaryPaths.Clear()
	if err != nil {
		return err
	}

	if lockitBinaryPaths.IsEmpty() {
		return lib.ErrLockitFilePartsListEmpty()
	}

	return lfei.validateLineBreaksCount(lockitBinaryPaths, fileOptions)
}

func (lfei *lockitFileExtractorIntegrity) ensureAllLockitExtractedTextFileParts(lockitTextPartsList components.IList[lockitParts.LockitFileParts], fileOptions core.ILockitFileOptions) error {
	partsLength := fileOptions.GetPartsLength()

	if err := lfei.ensurePartsListCount(lockitTextPartsList, partsLength); err != nil {
		lfei.logger.LogError(err, "error ensuring translated lockit text parts")

		return err
	}

	translatedTextList, err := lfei.createPartsPathsList(lockitTextPartsList, fileOptions)
	defer translatedTextList.Clear()
	if err != nil {
		return err
	}

	return lfei.validateLineBreaksCount(translatedTextList, fileOptions)
}

func (lfei *lockitFileExtractorIntegrity) ensurePartsListCount(partsList components.IList[lockitParts.LockitFileParts], partsLength int) error {
	if partsList.GetLength() != partsLength {
		return lib.ErrLockitFilePartsCountMismatch(partsLength, partsList.GetLength())
	}

	return nil
}

func (lfei *lockitFileExtractorIntegrity) createPartsPathsList(lockitFilePartsList components.IList[lockitParts.LockitFileParts], fileOptions core.ILockitFileOptions) (components.IList[string], error) {
	if lockitFilePartsList.IsEmpty() {
		return nil, lib.ErrLockitFilePartsListEmpty()
	}

	pathsList := components.NewList[string](fileOptions.GetPartsLength())

	lockitFilePartsList.ForEach(func(part lockitParts.LockitFileParts) {
		pathsList.Add(part.GetSource().Get().Path)
	})

	return pathsList, nil
}

func (lfei *lockitFileExtractorIntegrity) validateLineBreaksCount(filePaths components.IList[string], fileOptions core.ILockitFileOptions) error {
	filePartsIntegrity := NewLockitFilePartsIntegrity(lfei.logger)

	return filePartsIntegrity.ValidatePartsLineBreaksCount(filePaths, fileOptions)
}
