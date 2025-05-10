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

func CreateNodeMap(gameVersion models.GameVersion, formatter interfaces.ITextFormatter) fileFormats.TreeMapNode {
	nodeMap := make(fileFormats.TreeMapNode, 1800)

	rootDir := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
	if rootDir == "" {
		return nil
	}

	entrySource, err := locations.NewSource(rootDir)
	if err != nil {
		return nil
	}

	entrySource.PopulateDuplicatesFiles(gameVersion)

	destination := locations.NewDestination()
	destination.InitializeLocations(entrySource, formatter)

	rootMapNode := &fileFormats.MapNode{}

	rootMapNode.SetNodeKey(rootDir)

	addTreeNodeLabel(rootMapNode, gameVersion)
	addTreeNodeIcon(rootMapNode, entrySource.GetType())
	addTreeNodeData(rootMapNode, entrySource, destination)

	nodeMap[rootDir] = rootMapNode

	err = filepath.WalkDir(rootDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil || path == rootDir {
			return err
		}

		entrySource, err := locations.NewSource(path)
		if err != nil {
			return err
		}

		if entrySource.GetSize() == 0 {
			return nil
		}

		entrySource.PopulateDuplicatesFiles(gameVersion)

		destination := locations.NewDestination()
		destination.InitializeLocations(entrySource, formatter)

		childNode := &fileFormats.MapNode{}

		childNode.SetNodeKey(path)
		childNode.SetNodeLabel(info.Name())

		addTreeNodeIcon(childNode, entrySource.GetType())
		addTreeNodeData(childNode, entrySource, destination)

		nodeMap[path] = childNode

		parent := entrySource.GetParentPath()
		if parentNode, ok := nodeMap[parent]; ok {
			parentNode.AddChildKey(path)
		}

		return nil
	})

	if err != nil {
		return nil
	}

	return nodeMap
}

func BuildTreeFromMap(nodeMap fileFormats.TreeMapNode, rootKey string) *TreeNode {
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

	treeNode := &TreeNode{
		Data: &fileFormats.TreeNodeData{},
	}

	treeNode.SetNodeKey(node.Key)
	treeNode.SetNodeLabel(node.Label)
	treeNode.SetNodeIcon(node.Icon)
	treeNode.SetNodeData(node.Data)

	return treeNode
}
