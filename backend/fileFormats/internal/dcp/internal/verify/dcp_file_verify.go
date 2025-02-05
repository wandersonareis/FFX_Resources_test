package verify

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type DcpFileVerify struct {
	fileValidator  IFileValidator
	segmentCounter ISegmentCounter

	log logger.ILoggerHandler
}

func NewDcpFileVerify(logger logger.ILoggerHandler) *DcpFileVerify {
	return &DcpFileVerify{
		fileValidator:  newFileValidator(logger),
		segmentCounter: new(segmentCounter),

		log: logger,
	}
}

func (lv *DcpFileVerify) VerifyExtract(dcpFileParts components.IList[parts.DcpFileParts], fileOptions core.IDcpFileOptions) error {
	if dcpFileParts.GetLength() != fileOptions.GetPartsLength() {
		lv.log.LogError(
			fmt.Errorf("invalid number of split files. Expected: %d Got: %d",
				fileOptions.GetPartsLength(),
				dcpFileParts.GetLength()), "Invalid number of split files")

		return fmt.Errorf("error when ensuring splited lockit parts")
	}

	if err := lv.segmentCounter.CountBinaryParts(dcpFileParts, fileOptions); err != nil {
		lv.log.LogError(err, "Error when counting binary parts in splited files")

		return fmt.Errorf("error when counting binary line breaks in splited files")
	}

	if err := lv.segmentCounter.CountTextParts(dcpFileParts); err != nil {
		lv.log.LogError(err, "Error when counting text segments in splited files")

		return fmt.Errorf("error when counting text segments in splited files")
	}

	return nil
}

func (lv *DcpFileVerify) VerifyCompress(destination locations.IDestination, formatter interfaces.ITextFormatter, options core.IDcpFileOptions) error {
	targetFile := destination.Import().Get().GetTargetFile()

	if err := destination.Import().Get().Validate(); err != nil {
		lv.log.LogError(err, "Error when validating reimported macrodic file: %s", targetFile)

		return fmt.Errorf("reimport file not exists: %s", err.Error())
	}

	if err := lv.fileValidator.Validate(targetFile, formatter, options); err != nil {
		lv.log.LogError(err, "Error when validating reimported macrodic file: %s", targetFile)

		return fmt.Errorf("error when validating reimported macrodic file: %w", err)
	}

	return nil
}
