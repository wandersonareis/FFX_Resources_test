package spira

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"io/fs"
	"path/filepath"
)

func CreateFileTreeMap(gameVersion models.GameVersion, formatter interfaces.ITextFormatter) fileFormats.TreeMapNode {
	nodeMap := make(fileFormats.TreeMapNode)

	rootDir := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
	if rootDir == "" {
		return nil
	}

	entrySource, err := locations.NewSource(rootDir, gameVersion)
	if err != nil {
		return nil
	}

	destination := locations.NewDestination()
	destination.InitializeLocations(entrySource, formatter)

	rootMapNode := &fileFormats.MapNode{}

	rootMapNode.SetNodeKey(rootDir)

	addTreeNodeLabel(rootMapNode, gameVersion)
	addTreeNodeIcon(rootMapNode, entrySource.Get().Type)
	addTreeNodeData(rootMapNode, entrySource, destination)

	nodeMap[rootDir] = rootMapNode

	filepath.WalkDir(rootDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil || path == rootDir {
			return err
		}

		entrySource, err := locations.NewSource(path, gameVersion)
		if err != nil {
			return err
		}

		if entrySource.Get().Size == 0 {
			return filepath.SkipDir
		}

		destination := locations.NewDestination()
		destination.InitializeLocations(entrySource, formatter)

		childNode := &fileFormats.MapNode{}

		childNode.SetNodeKey(path)
		childNode.SetNodeLabel(info.Name())

		addTreeNodeIcon(childNode, entrySource.Get().Type)
		addTreeNodeData(childNode, entrySource, destination)

		nodeMap[path] = childNode

		parent := entrySource.Get().Parent
		if parentNode, ok := nodeMap[parent]; ok {
			parentNode.AddChildKey(path)
		}

		return nil
	})

	return nodeMap
}

func BuildTreeFromMap(nodeMap map[string]*fileFormats.MapNode, rootKey string) *TreeNode {
	rootNodeData, exists := nodeMap[rootKey]
	if !exists {
		return nil
	}

	treeNode := convertMapNodeToTreeNode(rootNodeData)

	for _, childKey := range rootNodeData.ChildKeys {
		childNode := BuildTreeFromMap(nodeMap, childKey)
		if childNode != nil {
			treeNode.AddNodeChild(childNode)
		}
	}

	return treeNode
}

func convertMapNodeToTreeNode(node *fileFormats.MapNode) *TreeNode {
	if node == nil {
		return nil
	}

	treeNode := &TreeNode{}

	treeNode.SetNodeKey(node.Key)
	treeNode.SetNodeLabel(node.Label)
	treeNode.SetNodeData(node.Data)
	treeNode.SetNodeIcon(node.Icon)

	return treeNode
}
