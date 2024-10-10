package spira

import (
	"ffxresources/backend/lib"
	"fmt"
	"strconv"
)

func ListFilesAndDirectories(source *lib.Source, prefix string) ([]lib.TreeNode, error) {
	var result []lib.TreeNode

	if !source.IsDir {
		return nil, nil
	}

	entries, err := source.ReadDir()
	if err != nil {
		return nil, err
	}

	for i, entry := range entries {
		entryPath := source.JoinEntryPath(entry)
		key := prefix + strconv.Itoa(i)

		if source.IsDir {
			entrySource, err := lib.NewSource(entryPath)
			if err != nil {
				return nil, err
			}

			children, err := ListFilesAndDirectories(entrySource, key+"-")
			if err != nil {
				return nil, err
			}

			node, err := CreateTreeNode(key, entrySource, children)
			if err != nil {
				return nil, err
			}

			result = append(result, node)
		} else {
			isSpira := lib.NewInteraction().GameLocation.IsSpiraPath(entryPath)
			if !isSpira {
				return nil, fmt.Errorf("invalid not spira path: %s", entryPath)
			}

			node, err := CreateTreeNode(key, source, nil)
			if err != nil {
				return nil, err
			}

			result = append(result, node)
		}
	}

	return result, nil
}
