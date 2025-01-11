package spira

import "ffxresources/backend/models"

type labelSetter interface {
	SetNodeLabel(label string)
}

func addTreeNodeLabel[T labelSetter](item T, gameVersion models.GameVersion) {
	label := getTreeRootLabel(gameVersion)
	
	item.SetNodeLabel(label)
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