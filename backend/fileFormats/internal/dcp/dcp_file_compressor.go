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
	"os"
	"path/filepath"
)

type (
	IDcpFileCompressor interface {
		Compress() error
	}

	dcpFileCompressor struct {
		source      interfaces.ISource
		destination locations.IDestination
		formatter   interfaces.ITextFormatter
		options     core.IDcpFileOptions

		log logger.ILoggerHandler
	}
)

func NewDcpFileCompressor(
	source interfaces.ISource,
	destination locations.IDestination,
	formatter interfaces.ITextFormatter,
	options core.IDcpFileOptions,
	log logger.ILoggerHandler) IDcpFileCompressor {
	common.CheckArgumentNil(source, "source")
	common.CheckArgumentNil(destination, "destination")
	common.CheckArgumentNil(formatter, "formatter")
	common.CheckArgumentNil(options, "options")
	common.CheckArgumentNil(log, "log")

	return &dcpFileCompressor{
		source:      source,
		destination: destination,
		formatter:   formatter,
		options:     options,
		log:         log,
	}
}

func (dfc *dcpFileCompressor) Compress() error {
	dcpTranslatedTextPartsList := components.NewList[dcpParts.DcpFileParts](dfc.options.GetPartsLength())
	defer dcpTranslatedTextPartsList.Clear()

	if err := dfc.populateDcpTranslatedTextFileParts(dcpTranslatedTextPartsList); err != nil {
		return err
	}

	if err := dfc.ensureAllDcpTranslatedTextFileParts(dcpTranslatedTextPartsList); err != nil {
		return err
	}

	dcpExtractedBinaryPartsList := components.NewList[dcpParts.DcpFileParts](dfc.options.GetPartsLength())
	defer dcpExtractedBinaryPartsList.Clear()

	if err := dfc.populateDcpExtractedBinaryFileParts(dcpExtractedBinaryPartsList); err != nil {
		return err
	}

	if err := dfc.ensureAllDcpExtractedBinaryFileParts(dcpExtractedBinaryPartsList); err != nil {
		return err
	}

	if err := dfc.compressFilesParts(dcpExtractedBinaryPartsList); err != nil {
		return err
	}

	dcpTranslatedBinaryPartsList := components.NewList[dcpParts.DcpFileParts](dfc.options.GetPartsLength())
	defer dfc.disposePartsList(dcpTranslatedBinaryPartsList)

	if err := dfc.populateDcpTranslatedBinaryFileParts(dcpTranslatedBinaryPartsList); err != nil {
		return err
	}

	if err := dfc.ensureAllDcpTranslatedBinaryFileParts(dcpTranslatedBinaryPartsList); err != nil {
		return err
	}

	if err := dfc.joinFilesParts(dcpTranslatedBinaryPartsList); err != nil {
		return err
	}

	return nil
}

func (dfc *dcpFileCompressor) populateDcpExtractedBinaryFileParts(binaryPartsList components.IList[dcpParts.DcpFileParts]) error {
	return dcpParts.PopulateDcpBinaryFileParts(
		binaryPartsList,
		dfc.destination.Extract().GetTargetPath(),
		dfc.formatter,
	)
}

func (dfc *dcpFileCompressor) populateDcpTranslatedBinaryFileParts(binaryTranslatedPartsList components.IList[dcpParts.DcpFileParts]) error {
	dfc.log.LogInfo("Populating dcp translated binary file parts...")

	translatedBinaryPartsPath := filepath.Join(dfc.destination.Import().GetTargetDirectory(), lib.DCP_PARTS_TARGET_DIR_NAME)

	return dcpParts.PopulateDcpBinaryFileParts(
		binaryTranslatedPartsList,
		translatedBinaryPartsPath,
		dfc.formatter,
	)
}

func (dfc *dcpFileCompressor) populateDcpTranslatedTextFileParts(translatedTextPartsList components.IList[dcpParts.DcpFileParts]) error {
	dfc.log.LogInfo("Populating dcp translated text file parts...")

	return dcpParts.PopulateDcpTextFileParts(
		translatedTextPartsList,
		dfc.destination.Translate().Get().GetTargetPath(),
		dfc.formatter,
	)
}

func (dfc *dcpFileCompressor) ensureAllDcpExtractedBinaryFileParts(binaryExtractedPartsList components.IList[dcpParts.DcpFileParts]) error {
	if binaryExtractedPartsList.GetLength() == dfc.options.GetPartsLength() {
		return nil
	}

	dfc.log.LogInfo("Missing dcp file parts detected. Attempting to extract...")

	if err := dfc.extractMissingDcpFileParts(); err != nil {
		return err
	}

	if err := dfc.populateDcpExtractedBinaryFileParts(binaryExtractedPartsList); err != nil {
		return err
	}

	if err := lib.EnsurePartsListCount(dfc.options.GetPartsLength(), binaryExtractedPartsList.GetLength()); err != nil {
		return err
	}

	return nil
}

func (dfc *dcpFileCompressor) ensureAllDcpTranslatedBinaryFileParts(binaryTranslatedPartsList components.IList[dcpParts.DcpFileParts]) error {
	if err := lib.EnsurePartsListCount(dfc.options.GetPartsLength(), binaryTranslatedPartsList.GetLength()); err != nil {
		return err
	}

	errChan := make(chan error, binaryTranslatedPartsList.GetLength())

	binaryTranslatedPartsList.ForEach(func(part dcpParts.DcpFileParts) {
		if err := common.CheckPathExists(part.GetSource().Get().Path); err != nil {
			errChan <- err
		}

		if part.GetSource().Get().Size <= 0 {
			errChan <- lib.ErrInvalidFileSize(part.GetSource().Get().Path)
		}
	})

	return nil
}

func (dfc *dcpFileCompressor) ensureAllDcpTranslatedTextFileParts(translatedTextPartsList components.IList[dcpParts.DcpFileParts]) error {
	if err := lib.EnsurePartsListCount(dfc.options.GetPartsLength(), translatedTextPartsList.GetLength()); err != nil {
		return err
	}

	return nil
}

func (dfc *dcpFileCompressor) extractMissingDcpFileParts() error {
	splitter := dcpCore.NewDcpFileSpliter()
	return splitter.FileSplitter(dfc.source, dfc.destination, dfc.options)
}

func (dfc *dcpFileCompressor) compressFilesParts(partsList components.IList[dcpParts.DcpFileParts]) error {
	dfc.log.LogInfo("Compressing dcp file parts...")

	errChan := make(chan error, partsList.GetLength())

	compressor := func(part dcpParts.DcpFileParts) {
		if err := part.Compress(); err != nil {
			errChan <- err
		}
	}

	partsList.ForEach(compressor)

	close(errChan)

	var hasError bool

	for err := range errChan {
		hasError = true
		dfc.log.LogError(err, "failed to compress file part")
	}

	if hasError {
		return lib.ErrFailedToCompressFileParts()
	}

	return nil
}

func (dfc *dcpFileCompressor) joinFilesParts(dcpTranslatedBinaryPartsList components.IList[dcpParts.DcpFileParts]) error {
	dfc.log.LogInfo("Joining dcp file parts...")

	outputFile := dfc.destination.Import().GetTargetFile()

	dcpFileJoiner := dcpCore.NewDcpFileJoiner()
	if err := dcpFileJoiner.DcpFileJoiner(dfc.source, dfc.destination, dcpTranslatedBinaryPartsList, outputFile); err != nil {
		dfc.log.LogError(err, "error joining macrodic file: %s", outputFile)

		return fmt.Errorf("error joining macrodic file: %s", outputFile)
	}

	return nil
}
func (dfc *dcpFileCompressor) disposePartsList(partsList components.IList[dcpParts.DcpFileParts]) error {
	if partsList.IsEmpty() {
		return fmt.Errorf("cannot dispose empty parts list")
	}

	item := partsList.GetItems()[0]
	dir := item.GetSource().Get().Parent

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("error removing directory: %s", dir)
	}

	partsList.Clear()

	return nil
}
