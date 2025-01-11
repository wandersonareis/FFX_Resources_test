package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/logger"
	"ffxresources/backend/models"
	"ffxresources/backend/notifications"
	"fmt"
	"sync"
)

type ExtractService struct {
	dirExtractServideOnce sync.Once
	dirExtractService     IDirectoryExtractService
}

func NewExtractService() *ExtractService {
	return &ExtractService{}
}

func (e *ExtractService) Extract(path string) {
	defer func() {
		l := logger.Get()
		if r := recover(); r != nil {
			l.Error().
				Interface("recover", r).
				Str("file", common.GetFileName(path)).
				Msg("Panic occurred during extraction")

			notifications.NotifyError(fmt.Errorf("panic occurred: %v", r))
		}
	}()

	node, ok := NodeMap["path"]
	if !ok {
		notifications.NotifyError(fmt.Errorf("node not found for path: %s", path))
		return
	}

	if node.Data.Source.Type == models.Folder {
		e.dirExtractServideOnce.Do(func() {
			e.dirExtractService = &directoryExtractService{}
		})

		if err := e.dirExtractService.ProcessDirectory(path, NodeMap); err != nil {
			notifications.NotifyError(err)
			return
		}

		notifications.NotifySuccess(fmt.Sprintf("Directory %s extracted successfully!", node.Label))
		return
	}

	processor := node.Data.FileProcessor
	if processor != nil {
		if err := processor.Extract(); err != nil {
			notifications.NotifyError(err)
			return
		}
		notifications.NotifySuccess(fmt.Sprintf("File %s extracted successfully!", node.Label))
	}

}
