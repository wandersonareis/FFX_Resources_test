package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/models"
	"ffxresources/backend/notifications"
	"fmt"
	"sync"
)

type CompressService struct {
	dirCompressServideOnce sync.Once
	dirCompressService     IDirectoryService
}

func NewCompressService() *CompressService {
	return &CompressService{}
}

func (c *CompressService) Compress(path string) {
	common.CheckArgumentNil(path, "path")

	cb := func() error {
		return c.compress(path)
	}

	if err := common.RecoverFn(cb); err != nil {
		notifications.NotifyError(err)
	}
}

func (c *CompressService) compress(path string) error {
	common.CheckArgumentNil(nodeStore, "nodeStore")

	node, ok := nodeStore.Get(path)
	if !ok {
		return fmt.Errorf("node not found for path: %s", path)
	}

	if node == nil {
		return fmt.Errorf("node is nil for path: %s", path)
	}

	if node.Data.Source.Type == models.Folder {
		c.dirCompressServideOnce.Do(func() {
			c.dirCompressService = NewDirectoryCompressService()
		})

		if err := c.dirCompressService.ProcessDirectory(path, nodeStore); err != nil {
			return err
		}

		notifications.NotifySuccess(fmt.Sprintf("Directory %s compressed successfully!", node.Label))
		return nil
	}

	processor := node.Data.FileProcessor
	if processor != nil {
		if err := processor.Compress(); err != nil {
			return err
		}

		notifications.NotifySuccess(fmt.Sprintf("File %s compressed successfully!", node.Label))
	}

	return nil
}
