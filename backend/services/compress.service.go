package services

import (
	"context"
	"ffxresources/backend/lib"
	"ffxresources/backend/spira"
	"fmt"
)

type CompressService struct {
	Ctx context.Context
}

func NewCompressService() *CompressService {
	return &CompressService{
		Ctx: context.Background(),
	}
}

func (c *CompressService) Compress(fileInfo lib.FileInfo) {
	if !ensureFilesExist(fileInfo) {
		lib.EmitError(c.Ctx, fmt.Errorf("original or extracted text file not found: %s", fileInfo.Name))
		return
	}

	err := lib.EnsureWindowsFormat(fileInfo)
	if err != nil {
		lib.EmitError(c.Ctx, err)
		return
	}

	if lib.CountSeparators(fileInfo) < 0 {
		lib.EmitError(c.Ctx, fmt.Errorf("text file contains no separators: %s", fileInfo.Name))
		return
	}

	var fileProcessor lib.ICompressor = spira.NewFileProcessor(c.Ctx, fileInfo)
	if fileProcessor == nil {
		lib.EmitError(c.Ctx, fmt.Errorf("invalid file type: %s", fileInfo.Name))
		return
	}
	fileProcessor.Compress()
}

func ensureFilesExist(fileInfo lib.FileInfo) bool {
	var isFine = true
	if !lib.FileExists(fileInfo.AbsolutePath) {
		isFine = false
	}

	if !lib.FileExists(fileInfo.ExtractLocation.TargetFile) {
		isFine = false
	}

	return isFine
}
