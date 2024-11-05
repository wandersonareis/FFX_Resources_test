package services

import (
	"ffxresources/backend/core"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"ffxresources/backend/spira"
)

type CollectionService struct{}

func NewCollectionService() *CollectionService {
	return &CollectionService{}
}

func (c *CollectionService) PopulateTree() []interactions.TreeNode {
	path := interactions.NewInteraction().GameLocation.GetPath()
	if path == "" {
		return nil
	}

	if err := interactions.NewInteraction().GameLocation.IsSpira(); err != nil {
		lib.NotifyError(err)
		return nil
	}

	source, err := core.NewSource(path)
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	tree := make([]interactions.TreeNode, 0, 1)

	if err := spira.BuildFileTree(&tree, source); err != nil {
		lib.NotifyError(err)
		return nil
	}
	
	return tree
}
