package services

import (
	"context"
	"ffxresources/backend/lib"
	"ffxresources/backend/spira"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type CollectionService struct {
	Ctx context.Context
}

func NewCollectionService() *CollectionService {
	return &CollectionService{
		Ctx: nil,
	}
}

func (c *CollectionService) PopulateTree() []lib.TreeNode {
	path := lib.NewInteraction().GameLocation.GetPath()
	if path == "" {
		return nil
	}

	err := lib.NewInteraction().GameLocation.IsSpira()
	if err != nil {
		runtime.EventsEmit(c.Ctx, "ApplicationError", err.Error())
		return nil
	}
	
	source, err := lib.NewSource(path)
	if err != nil {
		runtime.EventsEmit(c.Ctx, "ApplicationError", err.Error())
		return nil
	}

	tree, err := spira.ListFilesAndDirectories(source, "")
	if err != nil {
		runtime.EventsEmit(c.Ctx, "ApplicationError", err.Error())
		return nil
	}
	return tree
}
