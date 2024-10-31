package formats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
	"sync"
	"time"
)

type DcpFile struct {
	DataInfo *interactions.GameDataInfo
}

var errorCount = 0
func NewDcpFile(dataInfo *interactions.GameDataInfo) *DcpFile {
	dataInfo.ExtractLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)

	return &DcpFile{
		DataInfo: dataInfo,
	}
}

func (d DcpFile) GetFileInfo() *interactions.GameDataInfo {
	return d.DataInfo
}

func (d DcpFile) Extract() {
	err := dcpFileXpliter(d.GetFileInfo())
	if err != nil {
		lib.NotifyError(err)
		return
	}

	//Wait for the xpliter to finish
	time.Sleep(400 * time.Millisecond)

	macrodicPath := d.DataInfo.ExtractLocation.TargetPath
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

	err = processFilesParallel(&macrodicXplitedFiles, func(extractor *DcpFileParts) {
		if extractor.GetFileInfo().ExtractLocation.TargetFileExists() {
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
	macrodicImportPartsPath := common.PathJoin(d.DataInfo.ImportLocation.TargetDirectory, common.DCP_PARTS_TARGET_DIR_NAME)
	macrodicTranslatedPartsTextPath := d.DataInfo.TranslateLocation.TargetPath

	macrodicFilesPattern := common.MACRODIC_PATTERN + "$"
	macrodicTextFilesPattern := common.MACRODIC_PATTERN + "\\.txt"

	macrodicXplitedFiles := make([]string, 0, 7)
	if err := common.EnumerateFilesByPattern(&macrodicXplitedFiles, macrodicImportPartsPath, macrodicFilesPattern); err != nil {
		errorCount++

		if errorCount > 3 {
			lib.NotifyError(err)
			return
		}

		lib.LogSeverity(lib.SeverityInfo, fmt.Sprintf("wrong dcp xplited files count: %d", len(macrodicXplitedFiles)))
		dcpFileXpliter(d.GetFileInfo())
	}

	if len(macrodicXplitedFiles) != 7 {
		lib.NotifyError(fmt.Errorf("invalid number of xplited files: %d", len(macrodicXplitedFiles)))
		return
	}

	macrodicXplitedTextFiles := make([]string, 0, 7)
	if err := common.EnumerateFilesByPattern(&macrodicXplitedTextFiles, macrodicTranslatedPartsTextPath, macrodicTextFilesPattern); err != nil {
		lib.NotifyError(err)
		return
	}

	if len(macrodicXplitedTextFiles) != 7 {
		lib.NotifyError(fmt.Errorf("invalid number of xplited text files: %d", len(macrodicXplitedTextFiles)))
		return
	}

	if err := processFilesParallel(&macrodicXplitedFiles, func(compressor *DcpFileParts) {
		compressor.Compress()
	}); err != nil {
		lib.NotifyError(err)
		return
	}

	targetReimportFile := d.DataInfo.ImportLocation.TargetFile

	if err := dcpFileJoiner(d.DataInfo, &macrodicXplitedFiles, targetReimportFile); err != nil {
		lib.NotifyError(err)
		return
	}
}

func processFilesParallel(xplitedFiles *[]string, callback func(handler *DcpFileParts)) error {
	var wg sync.WaitGroup

	errChan := make(chan error, len(*xplitedFiles))

	for _, xplitedFile := range *xplitedFiles {
		wg.Add(1)

		go func(file string) {
			defer wg.Done()

			 dcpPartInfo := interactions.NewGameDataInfo(file)

			fileHandle := NewDcpFileParts(dcpPartInfo)
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
