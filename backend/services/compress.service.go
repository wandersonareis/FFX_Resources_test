package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/lib"
	"ffxresources/backend/spira"
	"fmt"
)

type CompressService struct{}

func NewCompressService() *CompressService {
	return &CompressService{}
}

func (c *CompressService) Compress(fileInfo *lib.FileInfo) {
	if !common.FileExists(fileInfo.AbsolutePath) {
		lib.NotifyError(fmt.Errorf("original text file %s not found", fileInfo.Name))
		return
	}

	if !fileInfo.TranslateLocation.TargetFileExists() && fileInfo.Type != common.Dcp {
		lib.NotifyError(fmt.Errorf("translated text file %s not found", fileInfo.TranslateLocation.TargetFileName))
		return
	}

	err := common.EnsureWindowsFormat(fileInfo.TranslateLocation.TargetFile, fileInfo.Type)
	if err != nil {
		lib.NotifyError(err)
		return
	}

	if common.CountSeparators(fileInfo.TranslateLocation.TargetFile) < 0 {
		lib.NotifyError(fmt.Errorf("text file contains no separators: %s", fileInfo.Name))
		return
	}

	var fileProcessor lib.ICompressor = spira.NewFileProcessor(fileInfo)
	if fileProcessor == nil {
		lib.NotifyError(fmt.Errorf("invalid file type: %s", fileInfo.Name))
		return
	}
	
	fileProcessor.Compress()
}
