package spira

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"path/filepath"
)

func BuildFileTree(result components.IList[TreeNode], source interfaces.ISource) error {
	if !source.Get().IsDir {
		return nil
	}

	entries, err := source.Get().ReadDir()
	if err != nil {
		return err
	}

	gamePart := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()

	for _, entry := range entries {
		entryPath := filepath.Join(source.Get().Path, entry.Name())
		key := entry.Name()

		entrySource, err := locations.NewSource(entryPath, gamePart)
		if err != nil {
			return err
		}

		childrenCapacity := len(entries)

		var children = components.NewEmptyList[TreeNode]()
		if childrenCapacity > 0 {
			children = components.NewList[TreeNode](childrenCapacity)
		}

		if entrySource.Get().IsDir {
			err = BuildFileTree(children, entrySource)
			if err != nil {
				return err
			}
		}

		destination := locations.NewDestination()
		destination.InitializeLocations(entrySource, formatters.NewTxtFormatterDev())

		node, err := CreateTreeNode(key, entrySource, destination, children)
		if err != nil {
			return err
		}

		result.Add(node)
	}

	return nil
}
