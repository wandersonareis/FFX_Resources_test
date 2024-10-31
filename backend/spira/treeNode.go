package spira

import (
	"ffxresources/backend/core"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
)

func getTreeNodeIcon(source *core.Source) string {
	var icon string

	switch source.Type {
	case models.Folder:
		icon = "pi pi-folder"
	case models.File:
		icon = "pi pi-file"
	case models.Dialogs:
		icon = "pi pi-file-word"
	case models.Kernel:
		icon = "pi pi-file-word"
	case models.Dcp:
		icon = "pi pi-file-plus"
	case models.Tutorial:
		icon = "pi pi-file-pdf"
	default:
		icon = ""
	}

	return icon
}

func generateNode(key string, source *core.Source) (interactions.TreeNode, error) {
	/* fileInfo := &lib.FileInfo{}
	lib.UpdateFileInfoFromSource(fileInfo, source) */
	fileInfo := interactions.NewGameDataInfo(source.Path)

	fileProcessor := NewFileProcessor(fileInfo)
	if fileProcessor == nil {
		return interactions.TreeNode{}, nil
	}

	dataInfo := fileProcessor.GetFileInfo()

	var node = interactions.TreeNode{
		Key:   key,
		Label: source.Name,
		Data:  *dataInfo,
	}

	return node, nil
}

func CreateTreeNode(key string, source *core.Source, childrens []interactions.TreeNode) (interactions.TreeNode, error) {
	node, err := generateNode(key, source)
	if err != nil {
		return interactions.TreeNode{}, err
	}

	node.Icon = getTreeNodeIcon(source)
	node.Children = childrens

	return node, nil
}
