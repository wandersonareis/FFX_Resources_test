package integrity

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	dcpCore "ffxresources/backend/fileFormats/internal/dcp/internal/core"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpParts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"path/filepath"
)

type (
	IDcpFileCompressorIntegrity interface {
		Verify(targetFile string, formatter interfaces.ITextFormatter, options core.IDcpFileOptions) error
	}

	dcpFileCompressorIntegrity struct {
		log logger.ILoggerHandler
	}
)

func NewDcpFileCompressorVerify(logger logger.ILoggerHandler) IDcpFileCompressorIntegrity {
	return &dcpFileCompressorIntegrity{
		log: logger,
	}
}

func (dfci *dcpFileCompressorIntegrity) Verify(targetFile string, formatter interfaces.ITextFormatter, fileOptions core.IDcpFileOptions) error {
	common.CheckArgumentNil(targetFile, "targetFile")
	common.CheckArgumentNil(formatter, "formatter")
	common.CheckArgumentNil(fileOptions, "fileOptions")

	if err := common.CheckPathExists(targetFile); err != nil {
		return err
	}

	source, destination, err := dfci.generateTempFile(targetFile, formatter)
	if err != nil {
		return err
	}

	if err := dfci.temporaryFileSplitter(source, destination, fileOptions); err != nil {
		return err
	}

	tempExtractedBinaryPartsList := components.NewList[dcpParts.DcpFileParts](fileOptions.GetPartsLength())
	defer tempExtractedBinaryPartsList.Clear()

	if err := dfci.populateTemporaryBinaryPartsList(
		tempExtractedBinaryPartsList,
		destination.Extract().GetTargetPath(),
		formatter,
		fileOptions); err != nil {
		return fmt.Errorf("error when checking lockit file integrity:: %s", err.Error())
	}

	if err := dfci.temporaryPartsDecoder(tempExtractedBinaryPartsList); err != nil {
		return err
	}

	if err := dfci.temporaryPartsIntegrity(tempExtractedBinaryPartsList); err != nil {
		return err
	}

	return nil
}

func (dfci *dcpFileCompressorIntegrity) populateTemporaryBinaryPartsList(
	tempPartsList components.IList[dcpParts.DcpFileParts],
	tempDir string,
	formatter interfaces.ITextFormatter,
	fileOptions core.IDcpFileOptions) error {
	if err := dcpParts.PopulateDcpBinaryFileParts(tempPartsList, tempDir, formatter); err != nil {
		return fmt.Errorf("error when checking lockit file integrity:: %w", err)
	}

	if err := lib.EnsurePartsListCount(tempPartsList.GetLength(), fileOptions.GetPartsLength()); err != nil {
		return err
	}

	setExtractTemporaryDirectory := func(part dcpParts.DcpFileParts) {
		newPartFile := filepath.Join(tempDir, common.GetFileName(part.GetDestination().Extract().GetTargetFile()))

		part.GetDestination().Extract().SetTargetFile(newPartFile)
		part.GetDestination().Extract().SetTargetPath(tempDir)
	}

	tempPartsList.ForEach(setExtractTemporaryDirectory)

	return nil
}

func (dfci *dcpFileCompressorIntegrity) temporaryPartsDecoder(tempPartsList components.IList[dcpParts.DcpFileParts]) error {
	defaultIntegrityError := fmt.Errorf("error when checking lockit file integrity")

	if tempPartsList.IsEmpty() {
		return defaultIntegrityError
	}

	errChan := make(chan error, tempPartsList.GetLength())

	tempPartsList.ForEach(func(part dcpParts.DcpFileParts) {
		if err := common.CheckPathExists(part.GetSource().Get().Path); err != nil {
			errChan <- err
			return
		}

		if err := part.Extract(); err != nil {
			errChan <- fmt.Errorf("failed to extract file part: %s", part.GetSource().Get().Name)
		}
	})

	close(errChan)

	for err := range errChan {
		if err != nil {
			dfci.log.LogError(err, "error when decoding temporary dcp file parts")
			return defaultIntegrityError
		}
	}

	return nil
}

func (dfci *dcpFileCompressorIntegrity) temporaryPartsIntegrity(tempPartsList components.IList[dcpParts.DcpFileParts]) error {
	filesToCompareList := dfci.createCompareTextList(tempPartsList)
	defer filesToCompareList.Clear()

	filesToCompareList.ForEach(func(item components.IFileComparer) {
		if err := item.CompareFiles(); err != nil {
			dfci.log.LogError(err, "failed to compare parts")
		}
	})

	return nil
}

func (dfci *dcpFileCompressorIntegrity) generateTempFile(file string, formatter interfaces.ITextFormatter) (interfaces.ISource, locations.IDestination, error) {
	source, err := locations.NewSource(file)
	if err != nil {
		return nil, nil, err
	}

	destination := locations.NewDestination()

	if err := destination.InitializeLocations(source, formatter); err != nil {
		return nil, nil, err
	}

	tmp := common.NewTempProvider("", "")
	tmpDirectory := filepath.Join(tmp.TempFilePath, "tmpDcp")

	destination.Extract().SetTargetPath(tmpDirectory)
	destination.Extract().SetTargetFile(tmp.TempFile)

	return source, destination, nil
}

func (dfci *dcpFileCompressorIntegrity) temporaryFileSplitter(source interfaces.ISource, destination locations.IDestination, fileOptions core.IDcpFileOptions) error {
	splitter := dcpCore.NewDcpFileSpliter()

	return splitter.FileSplitter(source, destination, fileOptions)
}
func (dfci *dcpFileCompressorIntegrity) createCompareTextList(partsList components.IList[dcpParts.DcpFileParts]) components.IList[components.IFileComparer] {
	filesToCompareList := components.NewList[components.IFileComparer](partsList.GetLength())

	partsList.ForEach(func(item dcpParts.DcpFileParts) {
		if item.GetDestination().Translate().GetTargetExtension() != ".txt" ||
			item.GetDestination().Extract().GetTargetExtension() != ".txt" {
			return
		}

		translatedFile := item.GetDestination().Translate().GetTargetFile()
		extractedFile := item.GetDestination().Extract().GetTargetFile()

		if err := common.CheckPathExists(translatedFile); err != nil {
			dfci.log.LogError(err, "failed to check path")
			return
		}

		if common.CountSegments(translatedFile) <= 0 {
			dfci.log.LogError(nil, "error when counting segments in translated part: %s", translatedFile)
		}

		if err := common.CheckPathExists(extractedFile); err != nil {
			dfci.log.LogError(err, "failed to check path")
			return
		}

		if common.CountSegments(extractedFile) <= 0 {
			dfci.log.LogError(nil, "error when counting segments in extracted part: %s", extractedFile)
		}

		filesToCompareList.Add(&components.FileComparisonEntry{
			FromFile: item.GetDestination().Translate().GetTargetFile(),
			ToFile:   item.GetDestination().Extract().GetTargetFile(),
		})
	})

	return filesToCompareList
}
