package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/models"
	"fmt"
	"sync"
)

type CompressService struct {
	dirCompressServideOnce sync.Once
	dirCompressService     IDirectoryService
	notifierService        INotificationService
	progressService        IProgressService
}

func NewCompressService(notificationService INotificationService, progressService IProgressService) *CompressService {
	return &CompressService{
		notifierService: notificationService,
		progressService: progressService,
	}
}

func (c *CompressService) Compress(path string) error {
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
		cb = c.compressDirectory
	default:
		cb = c.compressFile
	}

	if err := cb(node); err != nil {
		c.notifierService.NotifyError(err)
	}

	return nil
}

func (c *CompressService) compressFile(node *fileFormats.MapNode) error {
	if !NodeDataStore.IsNode(node) {
		return fmt.Errorf("node is invalid")
	}

	if node.Data.FileProcessor == nil {
		return fmt.Errorf("file processor is nil")
	}

	if err := node.Data.FileProcessor.Compress(); err != nil {
		return err
	}

	c.notifierService.NotifySuccess(fmt.Sprintf("File %s compressed successfully!", node.Data.Source.Name))
	return nil
}
func (c *CompressService) compressDirectory(node *fileFormats.MapNode) error {
	if node == nil {
		return fmt.Errorf("node is nil")
	}

	if node.Data.Source.Type != models.Folder {
		return fmt.Errorf("node is not a directory")
	}

	if c.dirCompressService == nil {
		c.dirCompressServideOnce.Do(func() {
			c.dirCompressService = NewDirectoryCompressService(c.notifierService, c.progressService)
		})
	}

	if err := c.dirCompressService.ProcessDirectory(node.Data.Source.Path, NodeDataStore); err != nil {
		return err
	}

	c.notifierService.NotifySuccess(fmt.Sprintf("Directory %s compressed successfully!", node.Data.Source.Name))
	return nil
}
