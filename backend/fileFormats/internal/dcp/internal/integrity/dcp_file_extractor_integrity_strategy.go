package integrity

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpParts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"fmt"
)

type dcpExtractionVerificationStrategy struct {
	formatterFactory formatters.IFormatterFactory
	verifyService    components.IVerificationService
}

func NewDcpExtractionVerificationStrategy() components.IVerificationStrategy {
	return &dcpExtractionVerificationStrategy{
		formatterFactory: formatters.NewFormatterFactory(),
		verifyService: components.NewVerificationService(),
	}
}

func (dev *dcpExtractionVerificationStrategy) Verify(source interfaces.ISource, destination locations.IDestination) error {
	dcpFileProperties := models.NewDcpFileOptions(source.GetVersion())
	if err := common.CheckArgumentNil(dcpFileProperties, "dcpFileProperties"); err != nil {
		return err
	}

	targetExt := destination.Extract().GetTargetExtension()
	formatter, err := dev.formatterFactory.CreateFormatter(targetExt)
	if err != nil {
		return fmt.Errorf("failed to create formatter for target extension %s: %w", targetExt, err)
	}

	if err := common.CheckArgumentNil(formatter, "formatter"); err != nil {
		return err
	}

	targetPath := destination.Extract().GetTargetPath()
	if err := common.CheckArgumentNil(targetPath, "targetPath"); err != nil {
		return err
	}

	if err := common.CheckPathExists(targetPath); err != nil {
		return lib.ErrDcpFileExtractIntegrityFailed(err)
	}

	if err := dev.verifyDcpBinaryParts(targetPath, formatter, dcpFileProperties.GetPartsLength()); err != nil {
		return lib.ErrDcpFileExtractIntegrityFailed(err)
	}

	if err := dev.verifyDcpTextParts(targetPath, formatter, dcpFileProperties.GetPartsLength()); err != nil {
		return lib.ErrDcpFileExtractIntegrityFailed(err)
	}

	return nil
}

func (dev *dcpExtractionVerificationStrategy) verifyDcpBinaryParts(targetPath string, formatter interfaces.ITextFormatter, expectedCount int) error {
	dcpBinaryPartsList := components.NewList[dcpParts.DcpFileParts](expectedCount)
	defer dcpBinaryPartsList.Clear()

	if err := dev.populateBinaryPartsList(dcpBinaryPartsList, targetPath, formatter); err != nil {
		return err
	}

	if err := dev.ensureAllDcpBinaryFileParts(dcpBinaryPartsList, expectedCount); err != nil {
		return err
	}

	return nil
}

func (dev *dcpExtractionVerificationStrategy) verifyDcpTextParts(targetPath string, formatter interfaces.ITextFormatter, expectedCount int) error {
	dcpTextPartsList := components.NewList[dcpParts.DcpFileParts](expectedCount)
	defer dcpTextPartsList.Clear()

	if err := dev.populateTextPartsList(dcpTextPartsList, targetPath, formatter); err != nil {
		return err
	}

	if err := dev.ensureAllDcpTextFileParts(dcpTextPartsList, expectedCount); err != nil {
		return err
	}

	return nil
}

func (dev *dcpExtractionVerificationStrategy) populateBinaryPartsList(
	binaryPartsList components.IList[dcpParts.DcpFileParts],
	path string,
	formatter interfaces.ITextFormatter) error {
	if err := dcpParts.PopulateDcpBinaryFileParts(binaryPartsList, path, formatter); err != nil {
		return err
	}

	return nil
}

func (dev *dcpExtractionVerificationStrategy) populateTextPartsList(
	textPartsList components.IList[dcpParts.DcpFileParts],
	path string,
	formatter interfaces.ITextFormatter) error {
	if err := dcpParts.PopulateDcpTextFileParts(textPartsList, path, formatter); err != nil {
		return err
	}

	return nil
}

func (dev *dcpExtractionVerificationStrategy) ensureAllDcpBinaryFileParts(binaryPartsList components.IList[dcpParts.DcpFileParts], expectedCount int) error {
	if err := lib.EnsurePartsListCount(binaryPartsList.GetLength(), expectedCount); err != nil {
		return fmt.Errorf("failed to validate of all DCP binary file parts: %w", err)
	}

	return nil
}

func (dev *dcpExtractionVerificationStrategy) ensureAllDcpTextFileParts(textPartsList components.IList[dcpParts.DcpFileParts], expectedCount int) error {
	if err := lib.EnsurePartsListCount(textPartsList.GetLength(), expectedCount); err != nil {
		return fmt.Errorf("failed to validate of all DCP text parts: %w", err)
	}

	return nil
}
