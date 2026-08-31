package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/spira"
	"fmt"
	"path/filepath"
)

type CollectionService struct {
	notifier INotificationService
}

func NewCollectionService(notifier INotificationService) *CollectionService {
	return &CollectionService{notifier: notifier}
}

func (c *CollectionService) BuildTree(path string) []spira.TreeNode {
	if path == "" {
		return nil
	}

	formatter := formatters.NewTxtFormatter()

	rawMap := c.CreateNodeDataStore(path, formatter)

	if err := common.CheckArgumentNil(rawMap, "BuildTree"); err != nil {
		c.notifier.NotifyError(fmt.Errorf("failed to create node data store"))
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		c.notifier.NotifyError(fmt.Errorf("failed to resolve path: %w", err))
		return nil
	}

	rootTreeNode := spira.BuildTreeFromMap(rawMap, absPath)

	return []spira.TreeNode{*rootTreeNode}
}

func (c *CollectionService) CreateNodeDataStore(
	path string,
	formatter interfaces.ITextFormatter) fileFormats.TreeMapNode {
	rawMap := spira.CreateNodeMap(path, formatter)

	if NodeDataStore == nil {
		NodeDataStore = NewNodeStore(rawMap)
	} else {
		NodeDataStore.Merge(rawMap)
	}

	return rawMap
}
