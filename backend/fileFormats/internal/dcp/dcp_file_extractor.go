package dcp

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpFileHandler"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpParts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"ffxresources/backend/models"
	"fmt"
)

type (
	IDcpFileExtractor interface {
		Extract() error
	}

	dcpFileExtractor struct {
		source            interfaces.ISource
		destination       locations.IDestination
		formatter         interfaces.ITextFormatter
		dcpFileSplitter   dcpFileHandler.IDcpFileSplitter
		dcpFileProperties models.IDcpFileProperties
		log               loggingService.ILoggerService
	}
)

func NewDcpFileExtractor(
	source interfaces.ISource,
	destination locations.IDestination,
	formatter interfaces.ITextFormatter,
	log loggingService.ILoggerService) IDcpFileExtractor {
	return &dcpFileExtractor{
		dcpFileSplitter:   dcpFileHandler.NewDcpFileSplitter(),
		dcpFileProperties: models.NewDcpFileOptions(source.GetVersion()),
		source:            source,
		destination:       destination,
		formatter:         formatter,
		log:               log,
	}
}

func (dfe *dcpFileExtractor) Extract() error {
	dcpBinaryPartsList := components.NewList[dcpParts.DcpFileParts](dfe.dcpFileProperties.GetPartsLength())
	defer dcpBinaryPartsList.Clear()

	if err := dfe.populateDcpBinaryFileParts(dcpBinaryPartsList); err != nil {
		return err
	}

	if err := dfe.ensureAllDcpBinaryFileParts(dcpBinaryPartsList); err != nil {
		return err
	}

	if err := dfe.decodeFilesParts(dcpBinaryPartsList); err != nil {
		return err
	}

	return nil
}

func (dfe *dcpFileExtractor) populateDcpBinaryFileParts(binaryPartsList components.IList[dcpParts.DcpFileParts]) error {
	err := dcpParts.PopulateDcpBinaryFileParts(
		binaryPartsList,
		dfe.destination.Extract().GetTargetPath(),
		dfe.formatter,
	)
	if err != nil {
		return fmt.Errorf("failed to populate dcp binary file parts: %w", err)
	}
	return nil
}

func (dfe *dcpFileExtractor) populateMissingDcpBinaryFileParts(binaryPartsList components.IList[dcpParts.DcpFileParts]) error {
	err := dcpParts.PopulateDcpBinaryFileParts(
		binaryPartsList,
		dfe.destination.Extract().GetTargetPath(),
		dfe.formatter,
	)
	if err != nil {
		return fmt.Errorf("failed to populate dcp binary file parts: %w", err)
	}

	if binaryPartsList.IsEmpty() {
		return fmt.Errorf("no dcp binary file parts found")
	}

	return nil
}

func (dfe *dcpFileExtractor) ensureAllDcpBinaryFileParts(binaryPartsList components.IList[dcpParts.DcpFileParts]) error {
	dcpFilePartsLen := dfe.dcpFileProperties.GetPartsLength()
	binaryPartsListLen := binaryPartsList.GetLength()

	if binaryPartsListLen == dcpFilePartsLen {
		return nil
	}

	dfe.log.Info("Missing dcp file parts detected. Attempting to extract...")

	if err := dfe.extractMissingDcpFileParts(); err != nil {
		return err
	}

	if err := dfe.populateMissingDcpBinaryFileParts(binaryPartsList); err != nil {
		return err
	}

	if err := lib.EnsurePartsListCount(dcpFilePartsLen, binaryPartsList.GetLength()); err != nil {
		return err
	}

	return nil
}

func (dfe *dcpFileExtractor) extractMissingDcpFileParts() error {
	if err := dfe.dcpFileSplitter.Split(dfe.source, dfe.destination, dfe.dcpFileProperties); err != nil {
		return fmt.Errorf("error when extracting dcp file parts: %w", err)
	}

	return nil
}

func (dfe *dcpFileExtractor) decodeFilesParts(binaryPartsList components.IList[dcpParts.DcpFileParts]) error {
	dfe.log.Info("Decoding files parts to: %s", dfe.destination.Import().GetTargetPath())

	var hasError bool

	extractParts := func(part dcpParts.DcpFileParts) {
		if err := common.CheckPathExists(part.GetSource().GetPath()); err != nil {
			dfe.log.Error(err, "error when checking dcp file part path: %s", part.GetSource().GetPath())
			hasError = true
		}

		if err := part.Extract(); err != nil {
			dfe.log.Error(err, "error when extracting dcp file part: %s", part.GetSource().GetPath())
			hasError = true
		}
	}

	binaryPartsList.ForEach(extractParts)

	if hasError {
		return fmt.Errorf("error when decoding dcp file parts")
	}

	return nil
}
