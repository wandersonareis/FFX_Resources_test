package spira

import (
	"ffxresources/backend/core"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interactions"
)

func CreateTreeNode(key string, source *core.Source, childrens []interactions.TreeNode) (interactions.TreeNode, error) {
	node, err := generateNode(key, source)
	if err != nil {
		return interactions.TreeNode{}, err
	}

	node.Icon = getTreeNodeIcon(source)
	node.Children = childrens

	return node, nil
}

func generateNode(key string, source *core.Source) (interactions.TreeNode, error) {
	fileInfo := interactions.NewGameDataInfo(source.Path)

	fileProcessor := fileFormats.NewFileProcessor(fileInfo)
	if fileProcessor == nil {
		return interactions.TreeNode{}, nil
	}

	dataInfo := fileProcessor.GetFileInfo()

	var node = interactions.TreeNode{
		Key:   key,
		Label: source.Name,
		Data:  *dataInfo.GetGameDataInfo(),
	}

	return node, nil
}
