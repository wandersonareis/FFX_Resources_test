package spira

import "ffxresources/backend/models"

type IconSetter interface {
	SetNodeIcon(icon string)
}

var nodeTypeIcons = map[models.NodeType]string{
	models.Dialogs:  "pi pi-file-word",
	models.Tutorial: "pi pi-file-pdf",
	models.Dcp:      "pi pi-file-plus",
	models.Kernel:   "pi pi-file-word",
	models.Folder:   "pi pi-folder",
}

func addTreeNodeIcon[T IconSetter](item T, nodeType models.NodeType) {
	newIcon := "pi pi-file"

	if iconString, ok := nodeTypeIcons[nodeType]; ok {
		newIcon = iconString
	}

	item.SetNodeIcon(newIcon)
}
