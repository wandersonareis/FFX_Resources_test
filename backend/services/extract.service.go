package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/models"
	"ffxresources/backend/notifications"
	"fmt"
)

type ExtractService struct {
	dirExtractService IDirectoryService
}

func NewExtractService() *ExtractService {
	return &ExtractService{}
}

func (e *ExtractService) Extract(path string) {
	common.CheckArgumentNil(path, "path")

	cb := func() error {
		return e.extract(path)
	}

	if err := common.RecoverFn(cb); err != nil {
		notifications.NotifyError(err)
	}
}

func (e *ExtractService) extract(path string) error {
	common.CheckArgumentNil(NodeMap, "NodeMap")

	node, ok := NodeMap[path]
	if !ok {
		return fmt.Errorf("node not found for path: %s", path)
	}

	if node.Data.Source.Type == models.Folder {
		if e.dirExtractService == nil {
			e.dirExtractService = NewDirectoryExtractService()
		}

		if err := e.dirExtractService.ProcessDirectory(path, NodeMap); err != nil {
			return err
		}

		notifications.NotifySuccess(fmt.Sprintf("Directory %s extracted successfully!", node.Label))
		return nil
	}

	processor := node.Data.FileProcessor
	if processor != nil {
		if err := processor.Extract(); err != nil {
			return err
		}

		notifications.NotifySuccess(fmt.Sprintf("File %s extracted successfully!", node.Label))
	}

	return nil
}
