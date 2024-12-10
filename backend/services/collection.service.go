package services

import (
	"ffxresources/backend/core"
	"ffxresources/backend/interactions"
	"ffxresources/backend/notifications"
	"ffxresources/backend/spira"
	"fmt"
)

type CollectionService struct{}

func NewCollectionService() *CollectionService {
	return &CollectionService{}
}

func (c *CollectionService) BuildTree() []interactions.TreeNode {
	path := interactions.NewInteraction().GameLocation.GetTargetDirectory()
	if path == "" {
		return nil
	}

	fmt.Println("Building tree for", interactions.NewInteraction().ImportLocation.GetTargetDirectory())

	if err := interactions.NewInteraction().GameLocation.IsSpira(); err != nil {
		notifications.NotifyError(err)
		return nil
	}

	source, err := core.NewSource(path)
	if err != nil {
		notifications.NotifyError(err)
		return nil
	}

	tree := make([]interactions.TreeNode, 0, 1)

	if err := spira.BuildFileTree(&tree, source); err != nil {
		notifications.NotifyError(err)
		return nil
	}

	return tree
}
