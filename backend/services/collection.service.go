package services

import (
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/notifications"
	"ffxresources/backend/spira"
	"fmt"
	"runtime"
)

type CollectionService struct{}

var nodeMap spira.TreeMapNode

func NewCollectionService() *CollectionService {
	return &CollectionService{}
}

func (c *CollectionService) BuildTree() []spira.TreeNode {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	beforeAlloc := m.Alloc

	/* errChan := make(chan error)
	go notifications.PanicRecover(errChan, logger.Get().With().Str("service", "collection").Logger())
	defer close(errChan) */

	path := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
	if path == "" {
		return nil
	}

	if err := interactions.NewInteractionService().GameLocation.IsSpira(); err != nil {
		notifications.NotifyError(err)
		return nil
	}

	/* source, err := locations.NewSource(path, interactions.NewInteractionService().FFXGameVersion().GetGameVersion())
	if err != nil {
		notifications.NotifyError(err)
		return nil
	} */

	formatter := formatters.NewTxtFormatter()

	//destination := locations.NewDestination()
	//destination.InitializeLocations(source, formatter)

	//tree := components.NewEmptyList[spira.TreeNode]()

	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()

	/* if err := spira.BuildFileTree(tree, source, gameVersion, formatter); err != nil {
		notifications.NotifyError(err)
		return nil
	} */

	nodeMap = spira.CreateFileTreeMap(gameVersion, formatter)
	rootTreeNode := spira.BuildTreeFromMap(nodeMap, path)

	runtime.ReadMemStats(&m)
	afterAlloc := m.Alloc
	

	fmt.Printf("Memory allocated: %d bytes\n", afterAlloc-beforeAlloc)
	fmt.Printf("Total allocations: %d\n", m.Mallocs-m.Frees)


	return []spira.TreeNode{*rootTreeNode}
}
