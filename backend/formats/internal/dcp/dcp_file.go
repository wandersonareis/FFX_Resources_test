package dcp

import (
	"ffxresources/backend/common"
	"ffxresources/backend/formats/internal/dcp/internal"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
	"sync"
	"time"
)

type DcpFile struct {
	dataInfo *interactions.GameDataInfo
	Parts    *[]dcp_internal.DcpFileParts
}

var errorCount = 0

func NewDcpFile(dataInfo *interactions.GameDataInfo) *DcpFile {
	parts := make([]dcp_internal.DcpFileParts, 0, 7)

	gameFilesPath := interactions.NewInteraction().GameLocation.TargetDirectory

	relative := common.GetDifferencePath(dataInfo.GameData.AbsolutePath, gameFilesPath)

	dataInfo.GameData.RelativePath = relative

	dataInfo.ExtractLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)

	if err := dcp_internal.FindDcpParts(&parts, dataInfo.ExtractLocation.TargetPath, common.MACRODIC_PATTERN+"$"); err != nil {
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

func (d DcpFile) Extract() {
	err := dcp_internal.DcpFileXpliter(d.GetFileInfo())
	if err != nil {
		lib.NotifyError(err)
		return
	}

	//Wait for the xpliter to finish
	time.Sleep(400 * time.Millisecond)

	macrodicPath := d.dataInfo.ExtractLocation.TargetPath
	macrodicXplitedFiles := make([]string, 0, 7)
	err = common.EnumerateFilesByPattern(&macrodicXplitedFiles, macrodicPath, common.MACRODIC_PATTERN+"$")
	if err != nil {
		lib.NotifyError(err)
		return
	}

	if len(macrodicXplitedFiles) != 7 {
		lib.NotifyError(fmt.Errorf("invalid number of xplited files: %d", len(macrodicXplitedFiles)))
		return
	}

	err = processFilesParallel(&macrodicXplitedFiles, func(extractor *dcp_internal.DcpFileParts) {
		if err := extractor.Validate(); err != nil {
			lib.NotifyError(err)
			return
		}

		extractor.Extract()
	})

	if err != nil {
		lib.NotifyError(err)
		return
	}
}

func (d DcpFile) Compress() {
	if len(*d.Parts) != 7 {
		dcpExtractedPath := common.PathJoin(d.dataInfo.ExtractLocation.TargetDirectory, common.DCP_PARTS_TARGET_DIR_NAME)
		dcpXplitedFiles := make([]string, 0, 7)
		dcpFilesPartsPattern := common.MACRODIC_PATTERN + "$"

		if err := common.EnumerateFilesByPattern(&dcpXplitedFiles, dcpExtractedPath, dcpFilesPartsPattern); err != nil {
			lib.NotifyError(err)
			dcp_internal.DcpFileXpliter(d.GetFileInfo())
			return
		}
	}
	//macrodicImportPartsPath := common.PathJoin(d.dataInfo.ExtractLocation.TargetDirectory, common.DCP_PARTS_TARGET_DIR_NAME)
	dcpTranslatedPartsTextPath := d.dataInfo.TranslateLocation.TargetPath

	dcpTextFilesPattern := common.MACRODIC_PATTERN + "\\.txt"

	if len(*d.Parts) != 7 {
		lib.NotifyError(fmt.Errorf("invalid number of xplited files"))
		return
	}

	dcpXplitedTextFiles := make([]string, 0, 7)
	if err := common.EnumerateFilesByPattern(&dcpXplitedTextFiles, dcpTranslatedPartsTextPath, dcpTextFilesPattern); err != nil {
		lib.NotifyError(err)
		return
	}

	if len(dcpXplitedTextFiles) != 7 {
		lib.NotifyError(fmt.Errorf("invalid number of xplited text files: %d", len(dcpXplitedTextFiles)))
		return
	}

	if err := processFilesParallelDev(d.Parts, func(compressor *dcp_internal.DcpFileParts) {
		compressor.Compress()
	}); err != nil {
		lib.NotifyError(err)
		return
	}

	targetReimportFile := d.dataInfo.ImportLocation.TargetFile

	if err := dcp_internal.DcpFileJoiner(d.dataInfo, d.Parts, targetReimportFile); err != nil {
		lib.NotifyError(err)
		return
	}
}

func processFilesParallel(xplitedFiles *[]string, callback func(handler *dcp_internal.DcpFileParts)) error {
	var wg sync.WaitGroup

	errChan := make(chan error, len(*xplitedFiles))

	for _, xplitedFile := range *xplitedFiles {
		wg.Add(1)

		go func(file string) {
			defer wg.Done()

			dcpPartInfo := interactions.NewGameDataInfo(file)

			fileHandle := dcp_internal.NewDcpFileParts(dcpPartInfo)
			if fileHandle == nil {
				errChan <- fmt.Errorf("invalid file type: %s", dcpPartInfo.GameData.Name)
				return
			}

			callback(fileHandle)

		}(xplitedFile)
	}

	wg.Wait()

	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func processFilesParallelDev(dcpParts *[]dcp_internal.DcpFileParts, callback func(handler *dcp_internal.DcpFileParts)) error {
	var wg sync.WaitGroup

	errChan := make(chan error, len(*dcpParts))

	for _, dcpPart := range *dcpParts {
		wg.Add(1)

		go func(file *dcp_internal.DcpFileParts) {
			defer wg.Done()

			callback(&dcpPart)

		}(&dcpPart)
	}

	wg.Wait()

	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
