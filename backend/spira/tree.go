package spira

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"io/fs"
	"path/filepath"
)

func CreateNodeMap(rootDir string, formatter interfaces.ITextFormatter) fileFormats.TreeMapNode {
	nodeMap := make(fileFormats.TreeMapNode, 1800)

	if rootDir == "" {
		return nil
	}

	entrySource, err := newEntrySource(rootDir)
	if err != nil {
		return nil
	}

	prepareSource(entrySource)

	destination, err := newDestination(entrySource, formatter)
	if err != nil {
		return nil
	}

	rootMapNode := createRootMapNode(rootDir, entrySource, destination)
	nodeMap[rootDir] = rootMapNode

	err = filepath.WalkDir(rootDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil || path == rootDir {
			return err
		}

		entrySource, err := newEntrySource(path)
		if err != nil {
			return err
		}

		if entrySource.GetSize() == 0 {
			return nil
		}

		prepareSource(entrySource)

		destination, err := newDestination(entrySource, formatter)
		if err != nil {
			return err
		}

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

func newEntrySource(rootDir string) (interfaces.ISource, error) {
	return locations.NewSource(rootDir)
}

func prepareSource(src interfaces.ISource) {
	if src.GetType() == models.Dialogs {
		src.PopulateDuplicatesFiles()
	}
}

func newDestination(src interfaces.ISource, formatter interfaces.ITextFormatter) (locations.IDestination, error) {
	dest := locations.NewDestination()
	if err := dest.InitializeLocations(src, formatter); err != nil {
		return nil, err
	}
	return dest, nil
}

func createRootMapNode(rootDir string, src interfaces.ISource, dest locations.IDestination) *fileFormats.MapNode {
	rootMapNode := &fileFormats.MapNode{}
	rootMapNode.SetNodeKey(rootDir)
	addTreeNodeLabel(rootMapNode, src.Get().Version)
	addTreeNodeIcon(rootMapNode, src.GetType())
	addTreeNodeData(rootMapNode, src, dest)
	return rootMapNode
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
