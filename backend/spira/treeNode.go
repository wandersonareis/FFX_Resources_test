package spira

import (
	"context"
	"ffxresources/backend/lib"
)

func getTreeNodeIcon(source *lib.Source) string {
	var icon string

	switch source.Type {
	case lib.Folder:
		icon = "pi pi-folder"
	case lib.File:
		icon = "pi pi-file"
	case lib.Dialogs:
		icon = "pi pi-file-word"
	case lib.Kernel:
		icon = "pi pi-file-word"
	case lib.Dcp:
		icon = "pi pi-file-plus"
	case lib.Tutorial:
		icon = "pi pi-file-pdf"
	default:
		icon = ""
	}

	return icon
}

func generateNode(key string, source *lib.Source) (lib.TreeNode, error) {
	fileInfo, err := lib.CreateFileInfo(source)
	if err != nil {
		return lib.TreeNode{}, err
	}

	var node = lib.TreeNode{
		Key:   key,
		Label: source.Name,
		Data:  NewFileProcessor(context.Background(), fileInfo).GetFileInfo(),
	}

	return node, nil
}

func CreateTreeNode(key string, source *lib.Source, childrens []lib.TreeNode) (lib.TreeNode, error) {
	node, err := generateNode(key, source)
	if err != nil {
		return lib.TreeNode{}, err
	}

	node.Icon = getTreeNodeIcon(source)
	node.Children = childrens

	return node, nil
}
