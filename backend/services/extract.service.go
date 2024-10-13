package services

import (
	"context"
	"ffxresources/backend/lib"
	"ffxresources/backend/spira"
	"fmt"
)

type ExtractService struct {
	Ctx context.Context
}

func NewExtractService() *ExtractService {
	return &ExtractService{
		Ctx: nil,
	}
}

func (e *ExtractService) Extract(fileInfo lib.FileInfo) {	
	var fileProcessor lib.IExtractor = spira.NewFileProcessor(e.Ctx, fileInfo)
	if fileProcessor == nil {
		lib.Notify(e.Ctx, lib.SeverityError, fmt.Sprintf("invalid file type: %s", fileInfo.Name))
		return
	}
	
	fileProcessor.Extract()
	lib.Notify(e.Ctx, lib.SeveritySuccess, "Extraction completed")
}
