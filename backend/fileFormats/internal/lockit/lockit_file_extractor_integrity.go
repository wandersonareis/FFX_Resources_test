package lockit

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/fileFormats/internal/lockit/internal/verify"
	"ffxresources/backend/logger"
)

type (
	ILockitFileExtractorIntegrity interface {
		VerifyFileIntegrity(destination locations.IDestination, fileOptions core.ILockitFileOptions) error
		Dispose()
	}

	lockitFileExtractorIntegrity struct {
		destination        locations.IDestination
		fileOptions        core.ILockitFileOptions
		filePartsIntegrity verify.ILockitFilePartsIntegrity

		logger logger.ILoggerHandler
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

	if lockitBinaryPartsList.GetLength() != partsLength {
		return lib.ErrLockitFilePartsCountMismatch(partsLength, lockitBinaryPartsList.GetLength())
	}

	lockitBinaryPartsPathList := lfei.createPartsPathsList(lockitBinaryPartsList)
	defer lockitBinaryPartsPathList.Clear()

	if lockitBinaryPartsPathList.GetLength() != partsLength {
		return lib.ErrLockitFilePartsCountMismatch(partsLength, lockitBinaryPartsPathList.GetLength())
	}

	return lfei.validateLineBreaksCount(lockitBinaryPartsPathList)
}

func (lfei *lockitFileExtractorIntegrity) ensureAllLockitExtractedTextFileParts(lockitTextPartsList components.IList[lockitParts.LockitFileParts]) error {
	partsLength := lfei.fileOptions.GetPartsLength()

	if lockitTextPartsList.GetLength() != partsLength {
		err := lib.ErrLockitFilePartsCountMismatch(partsLength, lockitTextPartsList.GetLength())

		lfei.logger.LogError(err, "error ensuring translated lockit text parts")

		return err
	}

	translatedTextList := lfei.createPartsPathsList(lockitTextPartsList)
	defer translatedTextList.Clear()

	return lfei.validateLineBreaksCount(translatedTextList)
}

func (lfei *lockitFileExtractorIntegrity) createPartsPathsList(lockitFilePartsList components.IList[lockitParts.LockitFileParts]) components.IList[string] {
	pathsList := components.NewList[string](lfei.fileOptions.GetPartsLength())
	defer pathsList.Clear()

	lockitFilePartsList.ForEach(func(part lockitParts.LockitFileParts) {
		pathsList.Add(part.Source().Get().Path)
	})

	return pathsList
}

func (lfei *lockitFileExtractorIntegrity) validateLineBreaksCount(filePaths components.IList[string]) error {
	if lfei.filePartsIntegrity == nil {
		lfei.filePartsIntegrity = verify.NewLockitFilePartsIntegrity(lfei.logger)
	}

	err := lfei.filePartsIntegrity.ValidatePartsLineBreaksCount(
		filePaths,
		lfei.fileOptions,
	)

	return err
}

func (lfei *lockitFileExtractorIntegrity) Dispose() {
	if lfei.filePartsIntegrity != nil {
		lfei.filePartsIntegrity.Dispose()
		lfei.filePartsIntegrity = nil
	}
}