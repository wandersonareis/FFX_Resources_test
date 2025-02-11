package services

import (
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/notifications"
	"ffxresources/backend/spira"
)

type CollectionService struct{}

func NewCollectionService() *CollectionService {
	return &CollectionService{}
}

func (c *CollectionService) BuildTree() []spira.TreeNode {
	path := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
	if path == "" {
		return nil
	}

	if err := interactions.NewInteractionService().GameLocation.IsSpira(); err != nil {
		notifications.NotifyError(err)
		return nil
	}

	formatter := formatters.NewTxtFormatter()

	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()

	NodeMap = spira.CreateNodeMap(gameVersion, formatter)
	rootTreeNode := spira.BuildTreeFromMap(NodeMap, path)

	return []spira.TreeNode{*rootTreeNode}
}
