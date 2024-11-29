package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/fileFormats/internal/dcp/internal/splitter"
	"ffxresources/backend/interactions"
	"fmt"
)

type DcpFileVerify struct {
	*base.FormatsBase

	fileValidator  IFileValidator
	segmentCounter ISegmentCounter
	fileSpliter    splitter.IDcpFileSpliter
	worker         common.IWorker[parts.DcpFileParts]
}

func NewDcpFileVerify(dataInfo interactions.IGameDataInfo) *DcpFileVerify {
	return &DcpFileVerify{
		FormatsBase:    base.NewFormatsBase(dataInfo),
		fileValidator:  newFileValidator(),
		segmentCounter: new(segmentCounter),
		fileSpliter:    new(splitter.DcpFileSpliter),
		worker:         common.NewWorker[parts.DcpFileParts](),
	}
}

func (lv *DcpFileVerify) VerifyExtract(dcpFileParts *[]parts.DcpFileParts, options interactions.DcpFileOptions) error {
	if len(*dcpFileParts) != options.PartsLength {
		return fmt.Errorf("error when ensuring splited lockit parts: expected %d | got %d", options.PartsLength, len(*dcpFileParts))
	}

	if err := lv.segmentCounter.CountBinaryParts(dcpFileParts, options); err != nil {
		return fmt.Errorf("error when counting binary line breaks in splited files: %w", err)
	}

	if err := lv.segmentCounter.CountTextParts(dcpFileParts, options); err != nil {
		return fmt.Errorf("error when counting text segments in splited files: %w", err)
	}

	return nil
}

func (lv *DcpFileVerify) VerifyCompress(dataInfo interactions.IGameDataInfo, options interactions.DcpFileOptions) error {
	errChan := make(chan error, 10)

	lv.Log.Info().Msgf("Verifying reimported macrodic file: %s", dataInfo.GetImportLocation().TargetFile)

	if err := dataInfo.GetImportLocation().Validate(); err != nil {
		errChan <- fmt.Errorf("reimport file not exists: %s | %w", dataInfo.GetImportLocation().TargetFile, err)
		return <-errChan
	}

	if err := lv.fileValidator.Validate(dataInfo.GetImportLocation().TargetFile, options); err != nil {
		errChan <- err
		return <-errChan
	}

	return nil
}
