package services

import (
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/notifications"
	"ffxresources/backend/spira"
	"fmt"
	"runtime"
)

type CollectionService struct{}

var NodeMap fileFormats.TreeMapNode

func NewCollectionService() *CollectionService {
	return &CollectionService{}
}

func (c *CollectionService) BuildTree() []spira.TreeNode {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	beforeAlloc := m.Alloc

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

	NodeMap = spira.CreateFileTreeMap(gameVersion, formatter)
	rootTreeNode := spira.BuildTreeFromMap(NodeMap, path)

	runtime.ReadMemStats(&m)
	afterAlloc := m.Alloc

	fmt.Printf("Memory allocated: %d bytes\n", afterAlloc-beforeAlloc)
	fmt.Printf("Total allocations: %d\n", m.Mallocs-m.Frees)

	return []spira.TreeNode{*rootTreeNode}
}
