package integrity

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	dcpfilehandler "ffxresources/backend/fileFormats/internal/dcp/internal/dcpFileHandler"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpParts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"fmt"
	"path/filepath"
)

type dcpCompressionVerificationStrategy struct {
	dcpFileSplitter  dcpfilehandler.IDcpFileSplitter
	formatterFactory formatters.IFormatterFactory
	verifyService    components.IVerificationService
	targetFile       string
}

func NewDcpCompressionVerificationStrategy(targetFile string) components.IVerificationStrategy {
	return &dcpCompressionVerificationStrategy{
		dcpFileSplitter:  dcpfilehandler.NewDcpFileSplitter(),
		formatterFactory: formatters.NewFormatterFactory(),
		verifyService:    components.NewVerificationService(),
		targetFile:       targetFile,
	}
}

func (dcv *dcpCompressionVerificationStrategy) Verify(source interfaces.ISource, destination locations.IDestination) error {
	dcpFileProperties := models.NewDcpFileOptions(source.GetVersion())
	if err := common.CheckArgumentNil(dcpFileProperties, "dcpFileProperties"); err != nil {
		return err
	}

	targetExt := destination.Extract().GetTargetExtension()
	formatter, err := dcv.formatterFactory.CreateFormatter(targetExt)
	if err != nil {
		return fmt.Errorf("failed to create formatter for target extension %s: %w", targetExt, err)
	}

	if err := common.CheckArgumentNil(formatter, "formatter"); err != nil {
		return err
	}

	if err := dcv.verify(formatter, dcpFileProperties); err != nil {
		return fmt.Errorf("an error occurred while verifying the content of the compressed file '%s': %v", dcv.targetFile, err)
	}

	return nil
}
func (dcv *dcpCompressionVerificationStrategy) verify(formatter interfaces.ITextFormatter, dcpFileProperties models.IDcpFileProperties) error {
	if err := common.CheckArgumentNil(dcv.targetFile, "targetFile"); err != nil {
		return err
	}

	source, destination, err := dcv.generateTempFile(dcv.targetFile, formatter)
	if err != nil {
		return err
	}

	if err := dcv.temporaryFileSplitter(source, destination, dcpFileProperties); err != nil {
		return err
	}

	tempExtractedBinaryPartsList := components.NewList[dcpParts.DcpFileParts](dcpFileProperties.GetPartsLength())
	defer tempExtractedBinaryPartsList.Clear()

	if err := dcv.populateTemporaryBinaryPartsList(
		tempExtractedBinaryPartsList,
		destination.Extract().GetTargetPath(),
		formatter,
		dcpFileProperties); err != nil {
		return fmt.Errorf("error when checking lockit file integrity:: %s", err.Error())
	}

	if err := dcv.temporaryPartsDecoder(tempExtractedBinaryPartsList); err != nil {
		return err
	}

	if err := dcv.temporaryPartsIntegrity(tempExtractedBinaryPartsList); err != nil {
		return err
	}

	return nil
}

func (dcv *dcpCompressionVerificationStrategy) populateTemporaryBinaryPartsList(
	tempPartsList components.IList[dcpParts.DcpFileParts],
	tempDir string,
	formatter interfaces.ITextFormatter,
	dcpFileProperties models.IDcpFileProperties) error {
	if err := dcpParts.PopulateDcpBinaryFileParts(tempPartsList, tempDir, formatter); err != nil {
		return fmt.Errorf("error when checking lockit file integrity:: %w", err)
	}

	if err := lib.EnsurePartsListCount(tempPartsList.GetLength(), dcpFileProperties.GetPartsLength()); err != nil {
		return err
	}

	setExtractTemporaryDirectory := func(part dcpParts.DcpFileParts) {
		targetFile := part.GetDestination().Extract().GetTargetFile()
		targetFileName := common.GetFileName(targetFile)
		newPartFile := filepath.Join(tempDir, targetFileName)

		part.GetDestination().Extract().SetTargetFile(newPartFile)
		part.GetDestination().Extract().SetTargetPath(tempDir)
	}

	tempPartsList.ForEach(setExtractTemporaryDirectory)

	return nil
}

func (dcv *dcpCompressionVerificationStrategy) temporaryPartsDecoder(tempPartsList components.IList[dcpParts.DcpFileParts]) error {
	if tempPartsList.IsEmpty() {
		return fmt.Errorf("error when checking dcp file integrity")
	}

	errChan := make(chan error, tempPartsList.GetLength())

	tempPartsList.ForEach(func(part dcpParts.DcpFileParts) {
		if err := common.CheckPathExists(part.GetSource().GetPath()); err != nil {
			errChan <- err
			return
		}

		if err := part.Extract(); err != nil {
			errChan <- fmt.Errorf("failed to extract file part: %s", part.GetSource().GetName())
		}
	})

	close(errChan)

	for err := range errChan {
		if err != nil {
			return fmt.Errorf("error when decoding temporary dcp file parts: %w", err)
		}
	}

	return nil
}

func (dcv *dcpCompressionVerificationStrategy) temporaryPartsIntegrity(tempPartsList components.IList[dcpParts.DcpFileParts]) error {
	filesToCompareList, err := dcv.createCompareTextList(tempPartsList)
	if err != nil {
		return fmt.Errorf("error when creating compare text list: %w", err)
	}
	defer filesToCompareList.Clear()

	errorChan := make(chan error, filesToCompareList.GetLength())

	filesToCompareList.ForEach(func(item components.IFileComparer) {
		if err := item.CompareFiles(); err != nil {
			errorChan <- fmt.Errorf("failed to compare parts: %w", err)
		}
	})

	close(errorChan)

	for err := range errorChan {
		if err != nil {
			return fmt.Errorf("error when comparing dcp file parts: %w", err)
		}
	}

	return nil
}

func (dcv *dcpCompressionVerificationStrategy) generateTempFile(file string, formatter interfaces.ITextFormatter) (interfaces.ISource, locations.IDestination, error) {
	source, err := locations.NewSource(file)
	if err != nil {
		return nil, nil, err
	}

	destination := locations.NewDestination(source.GetVersion().String())

	if err := destination.InitializeLocations(source, formatter); err != nil {
		return nil, nil, err
	}

	tmp := common.NewTempProvider("", "")
	tmpDirectory := filepath.Join(tmp.TempFilePath, "tmpDcp", source.GetVersion().String())

	destination.Extract().SetTargetPath(tmpDirectory)
	destination.Extract().SetTargetFile(tmp.TempFile)

	return source, destination, nil
}

func (dcv *dcpCompressionVerificationStrategy) temporaryFileSplitter(source interfaces.ISource, destination locations.IDestination, dcpFileProperties models.IDcpFileProperties) error {
	if err := dcv.dcpFileSplitter.Split(source, destination, dcpFileProperties); err != nil {
		return fmt.Errorf("error when splitting temporary dcp file parts: %w", err)
	}
	return nil
}
func (dcv *dcpCompressionVerificationStrategy) createCompareTextList(partsList components.IList[dcpParts.DcpFileParts]) (components.IList[components.IFileComparer], error) {
	filesToCompareList := components.NewList[components.IFileComparer](partsList.GetLength())

	errorChan := make(chan error, partsList.GetLength())

	partsList.ForEach(func(item dcpParts.DcpFileParts) {
		if item.GetDestination().Translate().GetTargetExtension() != ".txt" ||
			item.GetDestination().Extract().GetTargetExtension() != ".txt" {
			return
		}

		translatedFile := item.GetDestination().Translate().GetTargetFile()
		extractedFile := item.GetDestination().Extract().GetTargetFile()

		if err := common.CheckPathExists(translatedFile); err != nil {
			errorChan <- fmt.Errorf("failed to check path: %w", err)
			return
		}

		if common.CountSegments(translatedFile) <= 0 {
			errorChan <- fmt.Errorf("error when counting segments in translated part: %s", translatedFile)
		}

		if err := common.CheckPathExists(extractedFile); err != nil {
			errorChan <- fmt.Errorf("failed to check path: %w", err)
			return
		}

		if common.CountSegments(extractedFile) <= 0 {
			errorChan <- fmt.Errorf("error when counting segments in extracted part: %s", extractedFile)
		}

		filesToCompareList.Add(&components.FileComparisonEntry{
			FromFile: item.GetDestination().Translate().GetTargetFile(),
			ToFile:   item.GetDestination().Extract().GetTargetFile(),
		})
	})

	close(errorChan)
	for err := range errorChan {
		if err != nil {
			return nil, fmt.Errorf("error when creating compare text list: %w", err)
		}
	}

	return filesToCompareList, nil
}
