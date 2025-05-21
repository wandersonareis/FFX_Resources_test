package dcp

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	dcpCore "ffxresources/backend/fileFormats/internal/dcp/internal/core"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpParts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"ffxresources/backend/models"
	"fmt"
	"os"
	"path/filepath"
)

type (
	IDcpFileCompressor interface {
		Compress() error
	}

	dcpFileCompressor struct {
		source            interfaces.ISource
		destination       locations.IDestination
		formatter         interfaces.ITextFormatter
		dcpFileSplitter   dcpCore.IDcpFileSpliter
		dcpFileProperties models.IDcpFileProperties
		log               loggingService.ILoggerService
	}
)

func NewDcpFileCompressor(
	source interfaces.ISource,
	destination locations.IDestination,
	formatter interfaces.ITextFormatter,
	log loggingService.ILoggerService) IDcpFileCompressor {
	return &dcpFileCompressor{
		dcpFileSplitter:   dcpCore.NewDcpFileSpliter(),
		dcpFileProperties: models.NewDcpFileOptions(source.GetVersion()),
		source:            source,
		destination:       destination,
		formatter:         formatter,
		log:               log,
	}
}

func (dfc *dcpFileCompressor) Compress() error {
	dcpFilePartsLen := dfc.dcpFileProperties.GetPartsLength()
	if dcpFilePartsLen == 0 {
		return fmt.Errorf("dcp file parts length is zero")
	}

	dcpTranslatedTextPartsList := components.NewList[dcpParts.DcpFileParts](dcpFilePartsLen)
	defer dcpTranslatedTextPartsList.Clear()

	if err := dfc.populateDcpTranslatedTextFileParts(dcpTranslatedTextPartsList); err != nil {
		return err
	}

	if err := dfc.ensureAllDcpTranslatedTextFileParts(dcpTranslatedTextPartsList); err != nil {
		return err
	}

	dcpExtractedBinaryPartsList := components.NewList[dcpParts.DcpFileParts](dcpFilePartsLen)
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

	dcpTranslatedBinaryPartsList := components.NewList[dcpParts.DcpFileParts](dcpFilePartsLen)
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
	err := dcpParts.PopulateDcpBinaryFileParts(
		binaryPartsList,
		dfc.destination.Extract().GetTargetPath(),
		dfc.formatter,
	)
	if err != nil {
		return fmt.Errorf("failed to populate dcp extracted binary file parts: %w", err)
	}

	if binaryPartsList.IsEmpty() {
		return fmt.Errorf("no dcp extracted binary file parts found")
	}

	return nil
}

func (dfc *dcpFileCompressor) populateDcpTranslatedBinaryFileParts(binaryTranslatedPartsList components.IList[dcpParts.DcpFileParts]) error {
	dfc.log.Info("Populating dcp translated binary file parts...")

	targetDirectory := dfc.destination.Import().GetTargetDirectory()
	translatedBinaryPartsPath := filepath.Join(targetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME)

	err := dcpParts.PopulateDcpBinaryFileParts(
		binaryTranslatedPartsList,
		translatedBinaryPartsPath,
		dfc.formatter,
	)
	if err != nil {
		return fmt.Errorf("failed to populate dcp translated binary file parts: %w", err)
	}

	if binaryTranslatedPartsList.IsEmpty() {
		return fmt.Errorf("no dcp translated binary file parts found")
	}

	return nil
}

func (dfc *dcpFileCompressor) populateDcpTranslatedTextFileParts(translatedTextPartsList components.IList[dcpParts.DcpFileParts]) error {
	dfc.log.Info("Populating dcp translated text file parts...")

	err := dcpParts.PopulateDcpTextFileParts(
		translatedTextPartsList,
		dfc.destination.Translate().GetTargetPath(),
		dfc.formatter,
	)
	if err != nil {
		return fmt.Errorf("failed to populate dcp translated text file parts: %w", err)
	}

	if translatedTextPartsList.IsEmpty() {
		return fmt.Errorf("no dcp translated text file parts found")
	}

	return nil
}

func (dfc *dcpFileCompressor) ensureAllDcpExtractedBinaryFileParts(binaryExtractedPartsList components.IList[dcpParts.DcpFileParts]) error {
	dcpFilePartsLen := dfc.dcpFileProperties.GetPartsLength()

	if binaryExtractedPartsList.GetLength() == dcpFilePartsLen {
		return nil
	}

	dfc.log.Info("Missing dcp file parts detected. Attempting to extract...")

	if err := dfc.extractMissingDcpFileParts(); err != nil {
		return err
	}

	if err := dfc.populateDcpExtractedBinaryFileParts(binaryExtractedPartsList); err != nil {
		return err
	}

	if err := lib.EnsurePartsListCount(dcpFilePartsLen, binaryExtractedPartsList.GetLength()); err != nil {
		return err
	}

	return nil
}

func (dfc *dcpFileCompressor) ensureAllDcpTranslatedBinaryFileParts(binaryTranslatedPartsList components.IList[dcpParts.DcpFileParts]) error {
	if err := lib.EnsurePartsListCount(dfc.dcpFileProperties.GetPartsLength(), binaryTranslatedPartsList.GetLength()); err != nil {
		return err
	}

	errChan := make(chan error, binaryTranslatedPartsList.GetLength())

	binaryTranslatedPartsList.ForEach(func(part dcpParts.DcpFileParts) {
		if err := common.CheckPathExists(part.GetSource().GetPath()); err != nil {
			errChan <- fmt.Errorf("error checking path for %s: %w", part.GetSource().GetPath(), err)
		}

		if part.GetSource().GetSize() <= 0 {
			errChan <- fmt.Errorf("invalid file size for %s", part.GetSource().GetPath())
		}
	})

	close(errChan)

	for err := range errChan {
		if err != nil {
			return fmt.Errorf("translated binary file part error: %w", err)
		}
	}

	return nil
}

func (dfc *dcpFileCompressor) ensureAllDcpTranslatedTextFileParts(translatedTextPartsList components.IList[dcpParts.DcpFileParts]) error {
	if err := lib.EnsurePartsListCount(dfc.dcpFileProperties.GetPartsLength(), translatedTextPartsList.GetLength()); err != nil {
		return err
	}
	return nil
}

func (dfc *dcpFileCompressor) extractMissingDcpFileParts() error {
	splitter := dcpCore.NewDcpFileSpliter()
	return splitter.FileSplitter(dfc.source, dfc.destination, dfc.options)
}

func (dfc *dcpFileCompressor) compressFilesParts(partsList components.IList[dcpParts.DcpFileParts]) error {
	dfc.log.Info("Compressing dcp file parts...")

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
		dfc.log.Error(err, "failed to compress file part")
	}

	if hasError {
		return lib.ErrFailedToCompressFileParts()
	}
	return nil
}

func (dfc *dcpFileCompressor) joinFilesParts(dcpTranslatedBinaryPartsList components.IList[dcpParts.DcpFileParts]) error {
	dfc.log.Info("Joining dcp file parts...")

	outputFile := dfc.destination.Import().GetTargetFile()

	dcpFileJoiner := dcpCore.NewDcpFileJoiner()
	if err := dcpFileJoiner.DcpFileJoiner(dfc.source, dfc.destination, dcpTranslatedBinaryPartsList, outputFile); err != nil {
		dfc.log.Error(err, "error joining macrodic file: %s", outputFile)

		return fmt.Errorf("error joining macrodic file: %s", outputFile)
	}
	return nil
}

func (dfc *dcpFileCompressor) disposePartsList(partsList components.IList[dcpParts.DcpFileParts]) error {
	if partsList.IsEmpty() {
		return fmt.Errorf("cannot dispose empty parts list")
	}

	item := partsList.GetItems()[0]
	dir := item.GetSource().GetParentPath()

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("error removing directory: %s", dir)
	}

	partsList.Clear()

	return nil
}
