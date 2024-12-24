package services

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
	"ffxresources/backend/spira"
	"fmt"
)

type CollectionService struct{}

func NewCollectionService() *CollectionService {
	return &CollectionService{}
}

func (c *CollectionService) BuildTree() []spira.TreeNode {
	errChan := make(chan error)
	go notifications.PanicRecover(errChan, logger.Get().With().Str("service", "collection").Logger())
	defer close(errChan)

	path := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
	if path == "" {
		return nil
	}

	fmt.Println("Building tree for", interactions.NewInteractionService().ImportLocation.GetTargetDirectory())

	if err := interactions.NewInteractionService().GameLocation.IsSpira(); err != nil {
		notifications.NotifyError(err)
		return nil
	}

	source, err := locations.NewSource(path, interactions.NewInteractionService().FFXGameVersion().GetGameVersion())
	if err != nil {
		notifications.NotifyError(err)
		return nil
	}

	destination := locations.NewDestination()
	destination.InitializeLocations(source, formatters.NewTxtFormatterDev())

	tree := components.NewEmptyList[spira.TreeNode]()

	if err := spira.BuildFileTree(tree, source); err != nil {
		notifications.NotifyError(err)
		return nil
	}

	return tree.GetItems()
}
