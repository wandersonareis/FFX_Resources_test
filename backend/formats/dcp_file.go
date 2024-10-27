package formats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/lib"
	"fmt"
	"sync"
	"time"
)

type DcpFile struct {
	FileInfo *lib.FileInfo
}

func NewDcpFile(fileInfo *lib.FileInfo) *DcpFile {
	//relativePath, err := common.GetRelativePathFromMarker(fileInfo.AbsolutePath)
	/* if err != nil {
		lib.NotifyError(err)
		return nil
	} */

	//fileInfo.RelativePath = relativePath

	fileInfo.ExtractLocation.GenerateTargetOutput(NewTxtFormatter(), fileInfo)
	fileInfo.TranslateLocation.GenerateTargetOutput(NewTxtFormatter(), fileInfo)
	fileInfo.ImportLocation.GenerateTargetOutput(NewTxtFormatter(), fileInfo)

	return &DcpFile{
		FileInfo: fileInfo,
	}
}

func (d DcpFile) GetFileInfo() *lib.FileInfo {
	return d.FileInfo
}

func (d DcpFile) Extract() {
	err := dcpFileXpliter(d.GetFileInfo())
	if err != nil {
		lib.NotifyError(err)
		return
	}

	//Wait for the xpliter to finish
	time.Sleep(400 * time.Millisecond)

	macrodicPath := d.FileInfo.ExtractLocation.TargetPath
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
	macrodicImportPartsPath := common.PathJoin(d.FileInfo.ImportLocation.TargetDirectory, common.DCP_PARTS_TARGET_DIR_NAME)
	macrodicTranslatedPartsTextPath := d.FileInfo.TranslateLocation.TargetPath

	macrodicFilesPattern := common.MACRODIC_PATTERN + "$"
	macrodicTextFilesPattern := common.MACRODIC_PATTERN + "\\.txt"

	macrodicXplitedFiles := make([]string, 0, 7)
	if err := common.EnumerateFilesByPattern(&macrodicXplitedFiles, macrodicImportPartsPath, macrodicFilesPattern); err != nil {
		lib.NotifyError(err)
		return
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

	//reimportedDcpPartsDirectory := lib.PathJoin(d.FileInfo.ImportLocation.TargetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME, d.FileInfo.Name)

	targetReimportFile := d.FileInfo.ImportLocation.TargetFile

	if err := dcpFileJoiner(d.FileInfo, &macrodicXplitedFiles, targetReimportFile); err != nil {
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

			/* sourcePart, err := lib.NewSource(file)
			if err != nil {
				errChan <- err
				return
			}

			dcpPartInfo := &lib.FileInfo{}
			lib.UpdateFileInfoFromSource(dcpPartInfo, sourcePart) */

			dcpPartInfo, err := lib.CreateFileInfoFromPath(file)
			if err != nil {
				errChan <- err
				return
			}

			fileHandle := NewDcpFileParts(dcpPartInfo)
			if fileHandle == nil {
				errChan <- fmt.Errorf("invalid file type: %s", dcpPartInfo.Name)
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
