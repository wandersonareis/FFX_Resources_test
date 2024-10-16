package spira

import (
	"ffxresources/backend/lib"
)

func BuildFileTree(result *[]lib.TreeNode, source *lib.Source) error {
	if !source.IsDir {
		return nil
	}

	entries, err := source.ReadDir()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := source.JoinEntryPath(entry)
		key := entry.Name()

		entrySource, err := lib.NewSource(entryPath)
		if err != nil {
			return err
		}

        childrenCapacity := len(entries)
		var children []lib.TreeNode
        if childrenCapacity > 0 {
            children = make([]lib.TreeNode, 0, childrenCapacity)
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
