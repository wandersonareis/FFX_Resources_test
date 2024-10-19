package services

import (
	"ffxresources/backend/lib"
	"ffxresources/backend/spira"
	"fmt"
)

type ExtractService struct{}

func NewExtractService() *ExtractService {
	return &ExtractService{}
}

func (e *ExtractService) Extract(fileInfo *lib.FileInfo) {
	fileProcessor := spira.NewFileProcessor(fileInfo)
	if fileProcessor == nil {
		lib.NotifyError(fmt.Errorf("invalid file type: %s", fileInfo.Name))
		return
	}
	
	fileProcessor.Extract()
	lib.NotifySuccess("Extraction completed")
}
