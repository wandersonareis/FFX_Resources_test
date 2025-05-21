package integrity

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpParts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"ffxresources/backend/models"
	"fmt"
)

type (
	IDcpFileExtractorIntegrity interface {
		Verify(targetPath string, formatter interfaces.ITextFormatter, fileOptions models.IDcpFileProperties) error
	}

	dcpFileExtractorIntegrity struct {
		log loggingService.ILoggerService
	}
)

func NewDcpFileExtractorIntegrity(logger loggingService.ILoggerService) IDcpFileExtractorIntegrity {
	return &dcpFileExtractorIntegrity{
		log: logger,
	}
}

func (dfei *dcpFileExtractorIntegrity) Verify(targetPath string, formatter interfaces.ITextFormatter, fileOptions models.IDcpFileProperties) error {
	if err := common.CheckArgumentNil(targetPath, "targetPath"); err != nil {
		return err
	}
	if err := common.CheckArgumentNil(formatter, "formatter"); err != nil {
		return err
	}
	if err := common.CheckArgumentNil(fileOptions, "fileOptions"); err != nil {
		return err
	}

	if err := common.CheckPathExists(targetPath); err != nil {
		return lib.ErrDcpFileExtractIntegrityFailed(err)
	}

	if err := dfei.verifyDcpBinaryParts(targetPath, formatter, fileOptions.GetPartsLength()); err != nil {
		return lib.ErrDcpFileExtractIntegrityFailed(err)
	}

	if err := dfei.verifyDcpTextParts(targetPath, formatter, fileOptions.GetPartsLength()); err != nil {
		return lib.ErrDcpFileExtractIntegrityFailed(err)
	}

	return nil
}

func (dfei *dcpFileExtractorIntegrity) verifyDcpBinaryParts(targetPath string, formatter interfaces.ITextFormatter, expectedCount int) error {
	dcpBinaryPartsList := components.NewList[dcpParts.DcpFileParts](expectedCount)
	defer dcpBinaryPartsList.Clear()

	if err := dfei.populateBinaryPartsList(dcpBinaryPartsList, targetPath, formatter); err != nil {
		return err
	}

	if err := dfei.ensureAllDcpBinaryFileParts(dcpBinaryPartsList, expectedCount); err != nil {
		return err
	}

	return nil
}

func (dfei *dcpFileExtractorIntegrity) verifyDcpTextParts(targetPath string, formatter interfaces.ITextFormatter, expectedCount int) error {
	dcpTextPartsList := components.NewList[dcpParts.DcpFileParts](expectedCount)
	defer dcpTextPartsList.Clear()

	if err := dfei.populateTextPartsList(dcpTextPartsList, targetPath, formatter); err != nil {
		return err
	}

	if err := dfei.ensureAllDcpTextFileParts(dcpTextPartsList, expectedCount); err != nil {
		return err
	}

	return nil
}
func (dfei *dcpFileExtractorIntegrity) populateBinaryPartsList(
	binaryPartsList components.IList[dcpParts.DcpFileParts],
	path string,
	formatter interfaces.ITextFormatter) error {
	if err := dcpParts.PopulateDcpBinaryFileParts(binaryPartsList, path, formatter); err != nil {
		return err
	}

	return nil
}

func (dfei *dcpFileExtractorIntegrity) populateTextPartsList(
	textPartsList components.IList[dcpParts.DcpFileParts],
	path string,
	formatter interfaces.ITextFormatter) error {
	if err := dcpParts.PopulateDcpTextFileParts(textPartsList, path, formatter); err != nil {
		return err
	}

	return nil
}

func (dfei *dcpFileExtractorIntegrity) ensureAllDcpBinaryFileParts(binaryPartsList components.IList[dcpParts.DcpFileParts], expectedCount int) error {
	if err := lib.EnsurePartsListCount(binaryPartsList.GetLength(), expectedCount); err != nil {
		return fmt.Errorf("failed to ensure all DCP file binary parts: %w", err)
	}

	return nil
}

func (dfei *dcpFileExtractorIntegrity) ensureAllDcpTextFileParts(textPartsList components.IList[dcpParts.DcpFileParts], expectedCount int) error {
	if err := lib.EnsurePartsListCount(textPartsList.GetLength(), expectedCount); err != nil {
		return fmt.Errorf("failed to ensure all DCP file text parts: %w", err)
	}

	return nil
}
