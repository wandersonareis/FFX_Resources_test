package spira

import "ffxresources/backend/common"

type labelSetter interface {
	SetNodeLabel(label string)
}

func addTreeNodeLabel[T labelSetter](item T, gameVersion common.GameVersion) {
	label := getTreeRootLabel(gameVersion)
	
	item.SetNodeLabel(label)
}

func getTreeRootLabel(gameVersion common.GameVersion) string {
	var rootNodeLabel string

	switch gameVersion {
	case common.FFX:
		rootNodeLabel = "Final Fantasy X"
	case common.FFX2:
		rootNodeLabel = "Final Fantasy X-2"
	default:
		rootNodeLabel = "Unknown version"
	}

	return rootNodeLabel
}