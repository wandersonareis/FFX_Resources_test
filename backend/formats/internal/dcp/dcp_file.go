package dcp

import (
	"ffxresources/backend/common"
	"ffxresources/backend/formats/internal/dcp/internal"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
)

type DcpFile struct {
	dataInfo *interactions.GameDataInfo
	Parts    *[]internal.DcpFileParts
}

//var errorCount = 0

func NewDcpFile(dataInfo *interactions.GameDataInfo) *DcpFile {
	partsLength := interactions.NewInteraction().GamePartOptions.DcpPartsLength
	parts := make([]internal.DcpFileParts, 0, partsLength)

	gameFilesPath := interactions.NewInteraction().GameLocation.TargetDirectory

	relative := common.GetDifferencePath(dataInfo.GameData.AbsolutePath, gameFilesPath)

	dataInfo.GameData.RelativePath = relative

	dataInfo.ExtractLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)

	if err := internal.FindDcpParts(&parts, dataInfo.ExtractLocation.TargetPath, common.MACRODIC_PATTERN+"$"); err != nil {
		lib.NotifyError(err)
		return nil
	}

	return &DcpFile{
		dataInfo: dataInfo,
		Parts:    &parts,
	}
}

func (d DcpFile) GetFileInfo() *interactions.GameDataInfo {
	return d.dataInfo
}

func (d *DcpFile) Extract() {
	partsLength := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().DcpPartsLength

	if len(*d.Parts) != partsLength {
		err := internal.DcpFileXpliter(d.GetFileInfo())
		if err != nil {
			lib.NotifyError(err)
			return
		}

		newDcpFile := NewDcpFile(d.GetFileInfo())
		d.dataInfo = newDcpFile.GetFileInfo()
		d.Parts = newDcpFile.Parts
	}
	
	if len(*d.Parts) != partsLength {
		lib.NotifyError(fmt.Errorf("invalid number of xplited files: %d", partsLength))
		return
	}

	worker := lib.NewWorker[internal.DcpFileParts]()
	worker.Process(*d.Parts, func(i int, extractor internal.DcpFileParts) {
		if err := extractor.Validate(); err != nil {
			lib.NotifyError(err)
			return
		}

		extractor.Extract()
	})
}

func (d DcpFile) Compress() {
	partsLength := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().DcpPartsLength

	if len(*d.Parts) != partsLength {
		if err := internal.DcpFileXpliter(d.GetFileInfo()); err != nil {
			lib.NotifyError(err)
			return
		}

		newDcpFile := NewDcpFile(d.GetFileInfo())
		d.dataInfo = newDcpFile.GetFileInfo()
		d.Parts = newDcpFile.Parts
	}

	dcpTranslatedPartsTextPath := d.dataInfo.TranslateLocation.TargetPath

	dcpTextFilesPattern := common.MACRODIC_PATTERN + "\\.txt"

	if len(*d.Parts) != partsLength {
		lib.NotifyError(fmt.Errorf("invalid number of xplited files"))
		return
	}

	dcpXplitedTextFiles := make([]string, 0, partsLength)
	if err := common.EnumerateFilesByPattern(&dcpXplitedTextFiles, dcpTranslatedPartsTextPath, dcpTextFilesPattern); err != nil {
		lib.NotifyError(err)
		return
	}

	if len(dcpXplitedTextFiles) != partsLength {
		lib.NotifyError(fmt.Errorf("invalid number of xplited text files: %d", len(dcpXplitedTextFiles)))
		return
	}

	worker := lib.NewWorker[internal.DcpFileParts]()
	worker.Process(*d.Parts, func(i int, compressor internal.DcpFileParts) {
		compressor.Compress()
	})

	targetReimportFile := d.dataInfo.ImportLocation.TargetFile

	if err := internal.DcpFileJoiner(d.dataInfo, d.Parts, targetReimportFile); err != nil {
		lib.NotifyError(err)
		return
	}
}
