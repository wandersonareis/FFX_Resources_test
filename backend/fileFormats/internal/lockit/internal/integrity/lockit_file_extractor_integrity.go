package integrity

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/logger"
)

type (
	ILockitFileExtractorIntegrity interface {
		VerifyFileIntegrity(destination locations.IDestination, fileOptions core.ILockitFileOptions) error
	}

	lockitFileExtractorIntegrity struct {
		destination locations.IDestination
		fileOptions core.ILockitFileOptions
		logger      logger.ILoggerHandler
	}
)

func NewLockitFileExtractorIntegrity(logger logger.ILoggerHandler) ILockitFileExtractorIntegrity {
	return &lockitFileExtractorIntegrity{logger: logger}
}

func (lfei *lockitFileExtractorIntegrity) VerifyFileIntegrity(destination locations.IDestination, fileOptions core.ILockitFileOptions) error {
	lfei.initializeFields(destination, fileOptions)

	extractedLockitBinaryList := components.NewList[lockitParts.LockitFileParts](fileOptions.GetPartsLength())
	defer extractedLockitBinaryList.Clear()

	if err := lfei.populateLockitExtractedBinaryFileParts(extractedLockitBinaryList); err != nil {
		return err
	}

	if err := lfei.ensureAllLockitExtractedBinaryFileParts(extractedLockitBinaryList); err != nil {
		return err
	}

	extractedLockitTextList := components.NewList[lockitParts.LockitFileParts](fileOptions.GetPartsLength())
	defer extractedLockitTextList.Clear()

	if err := lfei.populateLockitExtractedTextFileParts(extractedLockitTextList); err != nil {
		return err
	}

	if err := lfei.ensureAllLockitExtractedTextFileParts(extractedLockitTextList); err != nil {
		return err
	}

	return nil
}

func (lfei *lockitFileExtractorIntegrity) initializeFields(destination locations.IDestination, fileOptions core.ILockitFileOptions) {
	lfei.destination = destination
	lfei.fileOptions = fileOptions
}

func (lfei *lockitFileExtractorIntegrity) populateLockitExtractedBinaryFileParts(binaryList components.IList[lockitParts.LockitFileParts]) error {
	return lockitParts.PopulateLockitBinaryFileParts(
		binaryList,
		lfei.destination.Extract().Get().GetTargetPath(),
	)
}

func (lfei *lockitFileExtractorIntegrity) populateLockitExtractedTextFileParts(lockitTextList components.IList[lockitParts.LockitFileParts]) error {
	return lockitParts.PopulateLockitTextFileParts(
		lockitTextList,
		lfei.destination.Extract().Get().GetTargetPath(),
	)
}

func (lfei *lockitFileExtractorIntegrity) ensureAllLockitExtractedBinaryFileParts(lockitBinaryPartsList components.IList[lockitParts.LockitFileParts]) error {
	partsLength := lfei.fileOptions.GetPartsLength()

	if err := lfei.ensurePartsListCount(lockitBinaryPartsList, partsLength); err != nil {
		return err
	}

	lockitBinaryPartsPathList, err := lfei.createPartsPathsList(lockitBinaryPartsList)
	defer lockitBinaryPartsPathList.Clear()
	if err != nil {
		return err
	}

	if err := lfei.validateLineBreaksCount(lockitBinaryPartsPathList); err != nil {
		return err
	}

	return lfei.validateLineBreaksCount(lockitBinaryPartsPathList)
}

func (lfei *lockitFileExtractorIntegrity) ensureAllLockitExtractedTextFileParts(lockitTextPartsList components.IList[lockitParts.LockitFileParts]) error {
	partsLength := lfei.fileOptions.GetPartsLength()

	if err := lfei.ensurePartsListCount(lockitTextPartsList, partsLength); err != nil {
		lfei.logger.LogError(err, "error ensuring translated lockit text parts")

		return err
	}

	translatedTextList, err := lfei.createPartsPathsList(lockitTextPartsList)
	defer translatedTextList.Clear()
	if err != nil {
		return err
	}

	return lfei.validateLineBreaksCount(translatedTextList)
}

func (lfei *lockitFileExtractorIntegrity) ensurePartsListCount(partsList components.IList[lockitParts.LockitFileParts], partsLength int) error {
	if partsList.GetLength() != partsLength {
		return lib.ErrLockitFilePartsCountMismatch(partsLength, partsList.GetLength())
	}

	return nil
}

func (lfei *lockitFileExtractorIntegrity) createPartsPathsList(lockitFilePartsList components.IList[lockitParts.LockitFileParts]) (components.IList[string], error) {
	if lockitFilePartsList.IsEmpty() {
		return nil, lib.ErrLockitFilePartsListEmpty()
	}

	pathsList := components.NewList[string](lfei.fileOptions.GetPartsLength())
	defer pathsList.Clear()

	lockitFilePartsList.ForEach(func(part lockitParts.LockitFileParts) {
		pathsList.Add(part.GetSource().Get().Path)
	})

	return pathsList, nil
}

func (lfei *lockitFileExtractorIntegrity) validateLineBreaksCount(filePaths components.IList[string]) error {
	filePartsIntegrity := NewLockitFilePartsIntegrity(lfei.logger)

	err := filePartsIntegrity.ValidatePartsLineBreaksCount(
		filePaths,
		lfei.fileOptions,
	)

	return err
}
