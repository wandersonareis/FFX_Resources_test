package spira

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"io/fs"
	"path/filepath"
)

type MapNode struct {
	Key       string       `json:"key"`
	Label     string       `json:"label"`
	Data      GameDataInfo `json:"data"`
	Icon      string       `json:"icon"`
	ChildKeys []string     `json:"childKeys"`
}

func CreateFileTreeMap(gameVersion models.GameVersion, formatter interfaces.ITextFormatterDev, paths ...string) TreeMapNode {
	nodeMap := make(TreeMapNode)

	var rootDir string
	if len(paths) > 0 && paths[0] != "" {
		rootDir = paths[0]
	} else {
		rootDir = interactions.NewInteractionService().GameLocation.GetTargetDirectory()
	}
	if rootDir == "" {
		return nil
	}

	entrySource, err := locations.NewSource(rootDir, gameVersion)
	if err != nil {
		return nil
	}
	destination := locations.NewDestination()
	destination.InitializeLocations(entrySource, formatter)

	rootTree := CreateTreeNode(entrySource, destination)
	rootNode := &MapNode{
		Key:   rootDir,
		Label: getTreeRootLabel(gameVersion),
		Data:  rootTree.Data,
		Icon:  rootTree.Icon,
	}
	nodeMap[rootDir] = rootNode

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
		childTree := CreateTreeNode(entrySource, destination)

		childNode := &MapNode{
			Key:   path,
			Label: info.Name(),
			Data:  childTree.Data,
			Icon:  childTree.Icon,
		}
		nodeMap[path] = childNode

		parent := filepath.Dir(path)
		if parentNode, ok := nodeMap[parent]; ok {
			parentNode.ChildKeys = append(parentNode.ChildKeys, path)
		}

		return nil
	})

	return nodeMap
}

func BuildTreeFromMap(nodeMap map[string]*MapNode, rootKey string) *TreeNode {
	rootNodeData, exists := nodeMap[rootKey]
	if !exists {
		return nil
	}

	treeNode := &TreeNode{
		Key:   rootNodeData.Data.Source.Name,
		Label: rootNodeData.Label,
		Data:  rootNodeData.Data,
		Icon:  rootNodeData.Icon,
	}

	for _, childKey := range rootNodeData.ChildKeys {
		childNode := BuildTreeFromMap(nodeMap, childKey)
		if childNode != nil {
			treeNode.Children = append(treeNode.Children, *childNode)
		}
	}

	return treeNode
}

func getTreeRootLabel(gameVersion models.GameVersion) string {
	var rootNodeLabel string

	switch gameVersion {
	case models.FFX:
		rootNodeLabel = "Final Fantasy X"
	case models.FFX2:
		rootNodeLabel = "Final Fantasy X-2"
	default:
		rootNodeLabel = "Unknown version"
	}

	return rootNodeLabel
}
