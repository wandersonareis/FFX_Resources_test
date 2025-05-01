package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/models"
	"fmt"
)

type ExtractService struct {
	dirExtractService   IDirectoryService
	NotificationService INotificationService
	ProgressService     IProgressService
}

func NewExtractService(notificationService INotificationService, progressService IProgressService) *ExtractService {
	return &ExtractService{
		NotificationService: notificationService,
		ProgressService:     progressService,
	}
}

func (e *ExtractService) Extract(path string) error {
	if err := common.CheckArgumentNil(path, "path"); err != nil {
		return err
	}

	if err := common.CheckArgumentNil(NodeDataStore, "nodeStore"); err != nil {
		return err
	}

	node, ok := NodeDataStore.Get(path)
	if !ok {
		return fmt.Errorf("node not found for path: %s", path)
	}

	if !NodeDataStore.IsNode(node) {
		return fmt.Errorf("node is invalid for path: %s", path)
	}

	var cb func(node *fileFormats.MapNode) error

	switch node.Data.Source.Type {
	case models.Folder:
		cb = e.extractDirectory
	default:
		cb = e.extractFile
	}

	if err := cb(node); err != nil {
		e.NotificationService.NotifyError(err)
	}

	return nil
}

func (e *ExtractService) extractFile(node *fileFormats.MapNode) error {
	if !NodeDataStore.IsNode(node) {
		return fmt.Errorf("node is invalid")
	}

	if node.Data.Source.Type == models.Folder {
		return fmt.Errorf("node is not a file")
	}

	if node.Data.FileProcessor == nil {
		return fmt.Errorf("file processor is nil")
	}

	if err := node.Data.FileProcessor.Extract(); err != nil {
		return err
	}

	e.NotificationService.NotifySuccess(fmt.Sprintf("File %s extracted successfully!", node.Data.Source.Name))
	return nil
}

func (e *ExtractService) extractDirectory(node *fileFormats.MapNode) error {
	if !NodeDataStore.IsNode(node) {
		return fmt.Errorf("node is invalid")
	}

	if node.Data.Source.Type != models.Folder {
		return fmt.Errorf("node is not a directory")
	}

	if e.dirExtractService == nil {
		e.dirExtractService = NewDirectoryExtractService(e.NotificationService, e.ProgressService)
	}

	if err := e.dirExtractService.ProcessDirectory(node.Data.Source.Path, NodeDataStore); err != nil {
		return err
	}

	e.NotificationService.NotifySuccess(fmt.Sprintf("Directory %s extracted successfully!", node.Data.Source.Name))
	return nil
}
