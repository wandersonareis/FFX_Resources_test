package spira

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interfaces"
)

type GameDataInfo struct {
	FilePath  string                      `json:"file_path"`
	Source    core.SpiraFileInfo          `json:"source"`
	Extract   locations.ExtractLocation   `json:"extract_location"`
	Translate locations.TranslateLocation `json:"translate_location"`
	FileProcessor interfaces.IFileProcessor
}

type TreeNode struct {
	Key      string       `json:"key"`
	Label    string       `json:"label"`
	Data     GameDataInfo `json:"data"`
	Icon     string       `json:"icon"`
	Children []TreeNode   `json:"children"`
}

type TreeMapNode = map[string]*MapNode

func CreateTreeNode(source interfaces.ISource, destination locations.IDestination) TreeNode {
	var node TreeNode

	node.Data = createTreeNodeData(source, destination)

	node.Icon = getTreeNodeIcon(source.Get().Type)

	return node
}

func createTreeNodeData(source interfaces.ISource, destination locations.IDestination) GameDataInfo {
	fileProcessor := fileFormats.NewFileProcessor(source, destination)

	gameDataInfo := GameDataInfo{
		FilePath:  source.Get().Path,
		Source:    *source.Get(),
		Extract:   *destination.Extract().Get(),
		Translate: *destination.Translate().Get(),

		FileProcessor: fileProcessor,
	}

	return gameDataInfo
}
