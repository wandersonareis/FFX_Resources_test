package spira

import (
	"ffxresources/backend/core"
	"ffxresources/backend/models"
)

type treeNodeIcon struct {
	Icon string
	Type models.NodeType
}

var icons = []treeNodeIcon{
	{
		Icon: "pi pi-folder",
		Type: models.Folder,
	},
	{
		Icon: "pi pi-file-word",
		Type: models.Dialogs,
	},
	{
		Icon: "pi pi-file-pdf",
		Type: models.Tutorial,
	},
	{
		Icon: "pi pi-file-plus",
		Type: models.Dcp,
	},
	{
		Icon: "pi pi-file-word",
		Type: models.Kernel,
	},
}

func getTreeNodeIcon(source *core.SpiraFileInfo) string {
	nodeIcon := "pi pi-file"

	for _, icon := range icons {
		if icon.Type == source.Type {
			nodeIcon = icon.Icon
			break
		}
	}

	return nodeIcon
}
