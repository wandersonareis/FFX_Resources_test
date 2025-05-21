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
	"ffxresources/backend/loggingService"
	"ffxresources/backend/models"
	"fmt"
	"os"
	"path/filepath"
)

type (
	ILockitFileCompressorIntegrity interface {
		Verify(
			destination locations.IDestination,
			lockitEncoding ffxencoding.IFFXTextLockitEncoding,
			fileOptions core.ILockitFileOptions) error
	}
	lockitFileCompressorIntegrity struct {
		formatter interfaces.ITextFormatter
		log       loggingService.ILoggerService
	}
)

func NewLockitFileIntegrity(logger loggingService.ILoggerService) ILockitFileCompressorIntegrity {
	return &lockitFileCompressorIntegrity{
		formatter: interactions.NewInteractionService().TextFormatter(),
		log:       logger,
	}
}

func (lfi *lockitFileCompressorIntegrity) Verify(
	destination locations.IDestination,
	lockitEncoding ffxencoding.IFFXTextLockitEncoding,
	fileOptions core.ILockitFileOptions) error {
	importTargetFile := destination.Import().GetTargetFile()

	if err := common.CheckPathExists(importTargetFile); err != nil {
		return fmt.Errorf("reimport file not exists: %s | %w", importTargetFile, err)
	}

	if err := lfi.verifyInitialIntegrity(importTargetFile, fileOptions); err != nil {
		return err
	}

	if err := lfi.verifyDataIntegrity(importTargetFile, lockitEncoding, fileOptions); err != nil {
		return err
	}

	return nil
}
func (lfi *lockitFileCompressorIntegrity) verifyDataIntegrity(file string, lockitEncoding ffxencoding.IFFXTextLockitEncoding, fileOptions core.ILockitFileOptions) error {
	source, destination := lfi.createTemporaryFile(file)

	splitter := internal.NewLockitFileSplitter()

	if err := splitter.FileSplitter(source, destination.Extract(), fileOptions); err != nil {
		return err
	}

	tempExtractedBinaryPartsList := components.NewList[lockitParts.LockitFileParts](fileOptions.GetPartsLength())
	defer tempExtractedBinaryPartsList.Clear()

	tempExtractedBinaryPath := destination.Extract().GetTargetPath()

	if err := lfi.populateTemporaryBinaryPartsList(tempExtractedBinaryPartsList, tempExtractedBinaryPath, fileOptions); err != nil {
		return fmt.Errorf("error when checking lockit file integrity:: %w", err)
	}

	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()
	if err := lfi.temporaryPartsDecoder(tempExtractedBinaryPartsList, lockitEncoding, gameVersion); err != nil {
		return err
	}

	if err := lfi.temporaryPartsComparer(tempExtractedBinaryPartsList); err != nil {
		return err
	}

	return nil
}

func (lfi *lockitFileCompressorIntegrity) verifyInitialIntegrity(filePath string, options core.ILockitFileOptions) error {
	targetLineBreaksCount, err := lfi.getLineBreakCount(filePath)
	if err != nil {
		return err
	}

	if err := lfi.ensureLineBreaksCount(targetLineBreaksCount, options.GetLineBreaksCount()); err != nil {
		return err
	}

	return nil
}

func (lfi *lockitFileCompressorIntegrity) getLineBreakCount(file string) (int, error) {
	lockitFileData, err := os.ReadFile(file)
	if err != nil {
		return 0, fmt.Errorf("error when reading imported lockit file %s", err.Error())
	}

	targetLineBreaksCount := bytes.Count(lockitFileData, []byte("\r\n"))

	return targetLineBreaksCount, nil
}

func (fv *lockitFileCompressorIntegrity) ensureLineBreaksCount(targetCount, expectedCount int) error {
	if targetCount != expectedCount {
		return fmt.Errorf("error when ensuring line breaks count: parts length is %d, expected %d", targetCount, expectedCount)
	}

	return nil
}

func (lfi *lockitFileCompressorIntegrity) populateTemporaryBinaryPartsList(tempPartsList components.IList[lockitParts.LockitFileParts], tempDir string, fileOptions core.ILockitFileOptions) error {
	if err := lockitParts.PopulateLockitBinaryFileParts(tempPartsList, tempDir); err != nil {
		return fmt.Errorf("error when checking lockit file integrity:: %w", err)
	}

	if tempPartsList.GetLength() != fileOptions.GetPartsLength() {
		return fmt.Errorf("error checking lockit parts integrity: expected %d, got %d",
			fileOptions.GetPartsLength(), tempPartsList.GetLength())
	}

	setExtractTemporaryDirectory := func(part lockitParts.LockitFileParts) {
		newPartFile := filepath.Join(tempDir, common.GetFileName(part.GetDestination().Extract().GetTargetFile()))

		part.GetDestination().Extract().SetTargetFile(newPartFile)
		part.GetDestination().Extract().SetTargetPath(tempDir)
	}

	tempPartsList.ForEach(setExtractTemporaryDirectory)

	return nil
}

func (lfi *lockitFileCompressorIntegrity) temporaryPartsDecoder(tempPartsList components.IList[lockitParts.LockitFileParts], lockitEncoding ffxencoding.IFFXTextLockitEncoding, gameVersion models.GameVersion) error {
	defaultIntegrityError := fmt.Errorf("error when checking lockit file integrity")

	if tempPartsList.IsEmpty() {
		lfi.log.Error(defaultIntegrityError, "")
		return defaultIntegrityError
	}

	filePartsDecoder := lockitParts.NewLockitFilePartsDecoder()
	if err := filePartsDecoder.DecodeFileParts(tempPartsList, lockitEncoding, gameVersion); err != nil {
		lfi.log.Error(err, "error when decoding temporary lockit file parts")
		return defaultIntegrityError
	}

	return nil
}

func (lfi *lockitFileCompressorIntegrity) temporaryPartsComparer(partsList components.IList[lockitParts.LockitFileParts]) error {
	if partsList.IsEmpty() {
		lfi.log.Error(nil, "error when comparing temporary lockit file parts: parts list is empty")

		return fmt.Errorf("error when checking lockit file integrity")
	}

	compareFilesList := components.NewList[models.FileComparisonEntry](partsList.GetLength())
	defer compareFilesList.Clear()

	partsList.ForEach(func(part lockitParts.LockitFileParts) {
		compareFilesList.Add(models.FileComparisonEntry{
			FromFile: part.GetDestination().Translate().GetTargetFile(),
			ToFile:   part.GetDestination().Extract().GetTargetFile(),
		})
	})

	filePartsIntegrity := NewLockitFilePartsIntegrity(lfi.log)

	if err := filePartsIntegrity.ComparePartsContent(compareFilesList); err != nil {
		return fmt.Errorf("error when comparing text parts: %s", err.Error())
	}

	return nil
}

func (lfi *lockitFileCompressorIntegrity) createTemporaryFile(file string) (interfaces.ISource, locations.IDestination) {
	source, err := locations.NewSource(file)
	if err != nil {
		return nil, nil
	}

	destination := locations.NewDestination(source.GetVersion().String())
	if err := destination.InitializeLocations(source, lfi.formatter); err != nil {
		return nil, nil
	}

	tmp := common.NewTempProvider("", "")
	tmpDirectory := filepath.Join(tmp.TempFilePath, "tmpLockit")

	destination.Extract().SetTargetPath(tmpDirectory)
	destination.Extract().SetTargetFile(tmp.TempFile)

	return source, destination
}
