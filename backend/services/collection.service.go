package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"ffxresources/backend/spira"
	"fmt"
)

type CollectionService struct {
	notifier INotificationService
}

func NewCollectionService(notifier INotificationService) *CollectionService {
	return &CollectionService{notifier: notifier}
}

func (c *CollectionService) BuildTree() []spira.TreeNode {
	path := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
	if path == "" {
		return nil
	}

	if err := interactions.NewInteractionService().GameLocation.IsSpira(); err != nil {
		c.notifier.NotifyError(err)
		return nil
	}

	formatter := formatters.NewTxtFormatter()

	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()

	rawMap := c.CreateNodeDataStore(gameVersion, formatter)

	if err := common.CheckArgumentNil(rawMap, "BuildTree"); err != nil {
		c.notifier.NotifyError(fmt.Errorf("failed to create node data store"))
		return nil
	}

	rootTreeNode := spira.BuildTreeFromMap(rawMap, path)

	return []spira.TreeNode{*rootTreeNode}
}

func (c *CollectionService) CreateNodeDataStore(
	gameVersion models.GameVersion,
	formatter interfaces.ITextFormatter) fileFormats.TreeMapNode {
	rawMap := spira.CreateNodeMap(gameVersion, formatter)
	NodeDataStore = NewNodeStore(rawMap)

	return rawMap
}
