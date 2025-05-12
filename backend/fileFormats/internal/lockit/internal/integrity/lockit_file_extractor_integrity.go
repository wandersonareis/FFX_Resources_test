package integrity

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/logger"
	"fmt"
)

type (
	ILockitFileExtractorIntegrity interface {
		Verify(targetPath string) error
	}

	lockitFileExtractorIntegrity struct {
		lockitFilePartsIntegrity ILockitFilePartsIntegrity

		options core.ILockitFileOptions
		logger  logger.ILoggerHandler
	}
)

func NewLockitFileExtractorIntegrity(
	options core.ILockitFileOptions,
	logger logger.ILoggerHandler) ILockitFileExtractorIntegrity {
	return &lockitFileExtractorIntegrity{
		lockitFilePartsIntegrity: NewLockitFilePartsIntegrity(logger),

		options: options,
		logger:  logger,
	}
}

func (lfei *lockitFileExtractorIntegrity) Verify(targetPath string) error {
	extractedLockitBinaryList := components.NewList[lockitParts.LockitFileParts](lfei.options.GetPartsLength())
	defer extractedLockitBinaryList.Clear()

	if err := lfei.populateLockitExtractedBinaryFileParts(targetPath, extractedLockitBinaryList); err != nil {
		return err
	}

	if err := lfei.ensureAllLockitExtractedBinaryFileParts(extractedLockitBinaryList); err != nil {
		return err
	}

	extractedLockitTextList := components.NewList[lockitParts.LockitFileParts](lfei.options.GetPartsLength())
	defer extractedLockitTextList.Clear()

	if err := lfei.populateLockitExtractedTextFileParts(targetPath, extractedLockitTextList); err != nil {
		return err
	}

	if err := lfei.ensureAllLockitExtractedTextFileParts(extractedLockitTextList); err != nil {
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

func (lfei *lockitFileExtractorIntegrity) ensureAllLockitExtractedBinaryFileParts(lockitBinaryPartsList components.IList[lockitParts.LockitFileParts]) error {
	if lockitBinaryPartsList.GetLength() != lfei.options.GetPartsLength() {
		return fmt.Errorf("extracted lockit binary file parts list length mismatch: expected %d, got %d",
			lfei.options.GetPartsLength(), lockitBinaryPartsList.GetLength())
	}

	lockitBinaryPaths, err := lfei.createPartsPathsList(lockitBinaryPartsList)
	defer lockitBinaryPaths.Clear()
	if err != nil {
		return err
	}

	if lockitBinaryPaths.IsEmpty() {
		return fmt.Errorf("lockit extracted binary file parts list is empty")
	}

	if err := lfei.validateLineBreaksCount(lockitBinaryPaths); err != nil {
		return fmt.Errorf("error validating line breaks count for extracted lockit binary file parts: %w", err)
	}

	return nil
}

func (lfei *lockitFileExtractorIntegrity) ensureAllLockitExtractedTextFileParts(lockitTextPartsList components.IList[lockitParts.LockitFileParts]) error {
	if lockitTextPartsList.GetLength() != lfei.options.GetPartsLength() {
		return fmt.Errorf("extracted lockit text file parts list length mismatch: expected %d, got %d",
			lfei.options.GetPartsLength(), lockitTextPartsList.GetLength())
	}


	translatedTextList, err := lfei.createPartsPathsList(lockitTextPartsList)
	defer translatedTextList.Clear()
	if err != nil {
		return err
	}

	if err := lfei.validateLineBreaksCount(translatedTextList); err != nil {
		return fmt.Errorf("error validating line breaks count for extracted lockit text file parts: %w", err)
	}

	return nil
}

func (lfei *lockitFileExtractorIntegrity) createPartsPathsList(lockitFilePartsList components.IList[lockitParts.LockitFileParts]) (components.IList[string], error) {
	if lockitFilePartsList.IsEmpty() {
		return nil, fmt.Errorf("lockit file parts list is empty")
	}

	pathsList := components.NewList[string](lfei.options.GetPartsLength())

	lockitFilePartsList.ForEach(func(part lockitParts.LockitFileParts) {
		pathsList.Add(part.GetSource().GetPath())
	})

	return pathsList, nil
}

func (lfei *lockitFileExtractorIntegrity) validateLineBreaksCount(filesList components.IList[string]) error {
	if err := lfei.lockitFilePartsIntegrity.ComparePartsLineBreaksCount(
		filesList,
		lfei.options,
	); err != nil {
		return fmt.Errorf("error validating line breaks count for lockit file parts: %w", err)
	}

	return nil
}
