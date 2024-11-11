package spira

import (
	"ffxresources/backend/core"
	"ffxresources/backend/interactions"
	"path/filepath"
)

func BuildFileTree(result *[]interactions.TreeNode, source *core.Source) error {
	if !source.IsDir {
		return nil
	}

	entries, err := source.ReadDir()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(source.Path, entry.Name())
		key := entry.Name()

		entrySource, err := core.NewSource(entryPath)
		if err != nil {
			return err
		}

        childrenCapacity := len(entries)
		
		var children []interactions.TreeNode
        if childrenCapacity > 0 {
            children = make([]interactions.TreeNode, 0, childrenCapacity)
        }

		if entrySource.IsDir {
			err = BuildFileTree(&children, entrySource)
			if err != nil {
				return err
			}
		}

		node, err := CreateTreeNode(key, entrySource, children)
		if err != nil {
			return err
		}

		*result = append(*result, node)
	}

	return nil
}
