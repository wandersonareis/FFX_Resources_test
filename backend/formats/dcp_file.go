package formats

import (
	"ffxresources/backend/lib"
	"fmt"
	"sync"
	"time"
)

const pattern = "macrodic\\..*?\\.00[0-6]"

type DcpFile struct {
	FileInfo *lib.FileInfo
}

func NewDcpFile(fileInfo *lib.FileInfo) *DcpFile {
	relativePath, err := lib.GetRelativePathFromMarker(fileInfo)
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	fileInfo.RelativePath = relativePath

	fileInfo.ExtractLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)
	fileInfo.TranslateLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)
	fileInfo.ImportLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)

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
	xplitedFiles, err := lib.EnumerateFilesByPattern(macrodicPath, pattern+"$")
	if err != nil {
		lib.NotifyError(err)
		return
	}

	if len(xplitedFiles) != 7 {
		lib.NotifyError(fmt.Errorf("invalid number of xplited files: %d", len(xplitedFiles)))
		return
	}

	err = processFilesParallel(xplitedFiles, func(extractor *DcpFileParts) {
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
	macrodicPartsPath := d.FileInfo.ExtractLocation.TargetPath
	macrodicPartsTextPath := d.FileInfo.TranslateLocation.TargetPath

	macrodicFilesPattern := pattern + "$"
	macrodicTextFilesPattern := pattern + "\\.txt"

	xplitedFiles, err := lib.EnumerateFilesByPattern(macrodicPartsPath, macrodicFilesPattern)
	if err != nil {
		lib.NotifyError(err)
		return
	}

	if len(xplitedFiles) != 7 {
		lib.NotifyError(fmt.Errorf("invalid number of xplited files: %d", len(xplitedFiles)))
		return
	}

	xplitedTextFiles, err := lib.EnumerateFilesByPattern(macrodicPartsTextPath, macrodicTextFilesPattern)
	if err != nil {
		lib.NotifyError(err)
		return
	}

	if len(xplitedTextFiles) != 7 {
		lib.NotifyError(fmt.Errorf("invalid number of xplited text files: %d", len(xplitedTextFiles)))
		return
	}

	err = processFilesParallel(xplitedFiles, func(compressor *DcpFileParts) {
		compressor.Compress()
	})
	if err != nil {
		lib.NotifyError(err)
		return
	}

	reimportedDcpPartsDirectory := lib.PathJoin(d.FileInfo.ImportLocation.TargetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME, d.FileInfo.Name)

	err = dcpFileJoiner(d.FileInfo, reimportedDcpPartsDirectory)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}

func processFilesParallel(xplitedFiles []string, callback func(handler *DcpFileParts)) error {
	var wg sync.WaitGroup

	errChan := make(chan error, len(xplitedFiles))

	for _, xplitedFile := range xplitedFiles {
		wg.Add(1)

		go func(file string) {
			defer wg.Done()

			sourcePart, err := lib.NewSource(file)
			if err != nil {
				errChan <- err
				return
			}

			dcpPartInfo := &lib.FileInfo{}
			lib.UpdateFileInfoFromSource(dcpPartInfo, sourcePart)

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
