package dcp

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
)

type (
	IDcpFileExtractor interface {
		Extract() error
	}

	dcpFileExtractor struct {
		source      interfaces.ISource
		destination locations.IDestination
		formatter   interfaces.ITextFormatter
		options     core.IDcpFileOptions

		log logger.ILoggerHandler
	}
)

func NewDcpFileExtractor(
	source interfaces.ISource,
	destination locations.IDestination,
	formatter interfaces.ITextFormatter,
	options core.IDcpFileOptions,
	log logger.ILoggerHandler) IDcpFileExtractor {
	common.CheckArgumentNil(source, "source")
	common.CheckArgumentNil(destination, "destination")
	common.CheckArgumentNil(formatter, "formatter")
	common.CheckArgumentNil(options, "options")
	common.CheckArgumentNil(log, "log")

	return &dcpFileExtractor{
		source:      source,
		destination: destination,
		formatter:   formatter,
		options:     options,
		log:         log,
	}
}

func (d *dcpFileExtractor) Extract() error {
	dcpBinaryPartsList := components.NewList[dcpParts.DcpFileParts](d.options.GetPartsLength())
	defer dcpBinaryPartsList.Clear()

	if err := d.populateDcpBinaryFileParts(dcpBinaryPartsList); err != nil {
		return err
	}

	if err := d.ensureAllDcpBinaryFileParts(dcpBinaryPartsList); err != nil {
		return err
	}

	if err := d.decodeFilesParts(dcpBinaryPartsList); err != nil {
		return err
	}

	return nil
}

func (dfe *dcpFileExtractor) populateDcpBinaryFileParts(binaryPartsList components.IList[dcpParts.DcpFileParts]) error {
	return dcpParts.PopulateDcpBinaryFileParts(
		binaryPartsList,
		dfe.destination.Extract().Get().GetTargetPath(),
		dfe.formatter,
	)
}

func (dfe *dcpFileExtractor) ensureAllDcpBinaryFileParts(binaryPartsList components.IList[dcpParts.DcpFileParts]) error {
	if binaryPartsList.GetLength() == dfe.options.GetPartsLength() {
		return nil
	}

	dfe.log.LogInfo("Missing dcp file parts detected. Attempting to extract...")

	if err := dfe.extractMissingDcpFileParts(); err != nil {
		return err
	}

	if err := dfe.populateDcpBinaryFileParts(binaryPartsList); err != nil {
		return err
	}

	if err := lib.EnsurePartsListCount(dfe.options.GetPartsLength(), binaryPartsList.GetLength()); err != nil {
		return err
	}

	return nil
}

func (dfe *dcpFileExtractor) extractMissingDcpFileParts() error {
	splitter := dcpCore.NewDcpFileSpliter()
	return splitter.FileSplitter(dfe.source, dfe.destination, dfe.options)
}

func (dfe *dcpFileExtractor) decodeFilesParts(binaryPartsList components.IList[dcpParts.DcpFileParts]) error {
	dfe.log.LogInfo("Decoding files parts to: %s", dfe.destination.Import().Get().GetTargetPath())

	var hasError bool

	extractParts := func(part dcpParts.DcpFileParts) {
		if err := common.CheckPathExists(part.GetSource().Get().Path); err != nil {
			dfe.log.LogError(err, "error when checking dcp file part path: %s", part.GetSource().Get().Path)
			hasError = true
		}

		if err := part.Extract(); err != nil {
			dfe.log.LogError(err, "error when extracting dcp file part: %s", part.GetSource().Get().Path)
			hasError = true
		}
	}

	binaryPartsList.ForEach(extractParts)

	if hasError {
		return fmt.Errorf("error when decoding dcp file parts")
	}

	return nil
}
