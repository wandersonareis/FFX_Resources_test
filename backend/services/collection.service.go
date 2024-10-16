package services

import (
	"ffxresources/backend/lib"
	"ffxresources/backend/spira"
)

type CollectionService struct{}

func NewCollectionService() *CollectionService {
	return &CollectionService{}
}

func (c *CollectionService) PopulateTree() []lib.TreeNode {
	path := lib.NewInteraction().GameLocation.GetPath()
	if path == "" {
		return nil
	}

	err := lib.NewInteraction().GameLocation.IsSpira()
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	source, err := lib.NewSource(path)
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	tree := make([]lib.TreeNode, 0, 1)

	err = spira.BuildFileTree(&tree, source)
	if err != nil {
		lib.NotifyError(err)
		return nil
	}
	return tree
}
