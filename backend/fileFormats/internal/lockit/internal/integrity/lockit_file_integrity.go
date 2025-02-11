package integrity

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/lockit/internal"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"ffxresources/backend/models"
	"fmt"
	"os"
	"path/filepath"
)

type (
	ILockitFileIntregrity interface {
		ValidateFileLineBreaksCount(destination locations.IDestination, fileOptions core.ILockitFileOptions) error
		VerifyFileIntegrity(file string, lockitEncoding ffxencoding.IFFXTextLockitEncoding, options core.ILockitFileOptions) error
	}
	lockitFileIntegrity struct {
		formatter interfaces.ITextFormatter
		log       logger.ILoggerHandler
	}
)

func newLockitFileIntegrity(logger logger.ILoggerHandler) ILockitFileIntregrity {
	return &lockitFileIntegrity{
		formatter: interactions.NewInteractionService().TextFormatter(),
		log:       logger,
	}
}

func (lfi *lockitFileIntegrity) ValidateFileLineBreaksCount(
	destination locations.IDestination,
	fileOptions core.ILockitFileOptions) error {
	importTargetFile := destination.Import().Get().GetTargetFile()
	if err := destination.Import().Get().Validate(); err != nil {
		return fmt.Errorf("reimport file not exists: %s | %w", importTargetFile, err)
	}

	if err := lfi.checkFileIntegrity(importTargetFile, fileOptions); err != nil {
		return err
	}
	return nil
}

func (lfi *lockitFileIntegrity) VerifyFileIntegrity(file string, lockitEncoding ffxencoding.IFFXTextLockitEncoding, fileOptions core.ILockitFileOptions) error {
	source, destination := lfi.createTemporaryFile(file)

	splitter := internal.NewLockitFileSplitter()

	if err := splitter.FileSplitter(source, destination.Extract().Get(), fileOptions); err != nil {
		return err
	}

	tempExtractedBinaryPartsList := components.NewList[lockitParts.LockitFileParts](fileOptions.GetPartsLength())
	defer tempExtractedBinaryPartsList.Clear()

	tempExtractedBinaryPath := destination.Extract().Get().GetTargetPath()

	if err := lfi.populateTemporaryBinaryPartsList(tempExtractedBinaryPartsList, tempExtractedBinaryPath, fileOptions); err != nil {
		return fmt.Errorf("error when checking lockit file integrity:: %w", err)
	}

	if err := lfi.temporaryPartsDecoder(tempExtractedBinaryPartsList, lockitEncoding); err != nil {
		return err
	}

	if err := lfi.temporaryPartsComparer(tempExtractedBinaryPartsList); err != nil {
		return err
	}

	return nil
}

func (lfi *lockitFileIntegrity) checkFileIntegrity(filePath string, options core.ILockitFileOptions) error {
	targetLineBreaksCount, err := lfi.getLineBreakCount(filePath)
	if err != nil {
		return err
	}

	if err := lfi.ensureLineBreaksCount(targetLineBreaksCount, options.GetLineBreaksCount()); err != nil {
		return err
	}

	return nil
}

func (lfi *lockitFileIntegrity) getLineBreakCount(file string) (int, error) {
	lockitFileData, err := os.ReadFile(file)
	if err != nil {
		return 0, fmt.Errorf("error when reading imported lockit file %s", err.Error())
	}

	targetLineBreaksCount := bytes.Count(lockitFileData, []byte("\r\n"))

	return targetLineBreaksCount, nil
}

func (fv *lockitFileIntegrity) ensureLineBreaksCount(targetCount, expectedCount int) error {
	if targetCount != expectedCount {
		return fmt.Errorf("error when ensuring line breaks count: parts length is %d, expected %d", targetCount, expectedCount)
	}

	return nil
}

func (lfi *lockitFileIntegrity) populateTemporaryBinaryPartsList(tempPartsList components.IList[lockitParts.LockitFileParts], tempDir string, fileOptions core.ILockitFileOptions) error {
	if err := lockitParts.PopulateLockitBinaryFileParts(tempPartsList, tempDir); err != nil {
		return fmt.Errorf("error when checking lockit file integrity:: %w", err)
	}

	if tempPartsList.GetLength() != fileOptions.GetPartsLength() {
		return fmt.Errorf("error checking lockit parts integrity: expected %d, got %d",
			fileOptions.GetPartsLength(), tempPartsList.GetLength())
	}

	setExtractTemporaryDirectory := func(part lockitParts.LockitFileParts) {
		newPartFile := filepath.Join(tempDir, common.GetFileName(part.GetDestination().Extract().Get().GetTargetFile()))

		part.GetDestination().Extract().Get().SetTargetFile(newPartFile)
		part.GetDestination().Extract().Get().SetTargetPath(tempDir)
	}

	tempPartsList.ForEach(setExtractTemporaryDirectory)

	return nil
}

func (lfi *lockitFileIntegrity) temporaryPartsDecoder(tempPartsList components.IList[lockitParts.LockitFileParts], lockitEncoding ffxencoding.IFFXTextLockitEncoding) error {
	defaultIntegrityError := fmt.Errorf("error when checking lockit file integrity")

	if tempPartsList.IsEmpty() {
		lfi.log.LogError(defaultIntegrityError, "")
		return defaultIntegrityError
	}

	filePartsDecoder := lockitParts.NewLockitFilePartsDecoder()
	if err := filePartsDecoder.DecodeFileParts(tempPartsList, lockitEncoding); err != nil {
		lfi.log.LogError(err, "error when decoding temporary lockit file parts")
		return defaultIntegrityError
	}

	return nil
}

func (lfi *lockitFileIntegrity) temporaryPartsComparer(partsList components.IList[lockitParts.LockitFileParts]) error {
	if partsList.IsEmpty() {
		lfi.log.LogError(nil, "error when comparing temporary lockit file parts: parts list is empty")

		return fmt.Errorf("error when checking lockit file integrity")
	}

	compareFilesList := components.NewList[models.FileComparisonEntry](partsList.GetLength())
	defer compareFilesList.Clear()

	partsList.ForEach(func(part lockitParts.LockitFileParts) {
		compareFilesList.Add(models.FileComparisonEntry{
			FromFile: part.GetDestination().Translate().Get().GetTargetFile(),
			ToFile:   part.GetDestination().Extract().Get().GetTargetFile(),
		})
	})

	filePartsIntegrity := NewLockitFilePartsIntegrity(lfi.log)

	if err := filePartsIntegrity.ComparePartsContent(compareFilesList); err != nil {
		return fmt.Errorf("error when comparing text parts: %s", err.Error())
	}

	return nil
}

func (lfi *lockitFileIntegrity) createTemporaryFile(file string) (interfaces.ISource, locations.IDestination) {
	source, err := locations.NewSource(file)
	if err != nil {
		return nil, nil
	}

	destination := locations.NewDestination()
	//gameFileLocation := interactions.NewInteractionService().GameLocation.GetTargetDirectory()

	//destination.CreateRelativePath(source, gameFileLocation)
	destination.InitializeLocations(source, lfi.formatter)

	tmp := common.NewTempProvider("", "")
	tmpDirectory := filepath.Join(tmp.TempFilePath, "tmpLockit")

	destination.Extract().Get().SetTargetPath(tmpDirectory)
	destination.Extract().Get().SetTargetFile(tmp.TempFile)

	return source, destination
}
