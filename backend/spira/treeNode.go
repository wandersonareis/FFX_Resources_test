package spira

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interfaces"
)

type GameDataInfo struct {
	FilePath  string                      `json:"file_path"`
	Extract   locations.ExtractLocation   `json:"extract_location"`
	Translate locations.TranslateLocation `json:"translate_location"`
	Import    locations.ImportLocation    `json:"import_location"`
}

type TreeNode struct {
	Key      string       `json:"key"`
	Label    string       `json:"label"`
	Data     GameDataInfo `json:"data"`
	Icon     string       `json:"icon"`
	Children []TreeNode   `json:"children"`
}

func CreateTreeNode(key string, source interfaces.ISource, destination locations.IDestination, childrens components.IList[TreeNode]) (TreeNode, error) {
	node, err := generateNode(key, source, destination)
	if err != nil {
		return TreeNode{}, err
	}

	node.Icon = getTreeNodeIcon(source.Get().Type)
	node.Children = childrens.GetItems()

	return node, nil
}

func generateNode(key string, source interfaces.ISource, destination locations.IDestination) (TreeNode, error) {
	//fileInfo := interactions.NewGameDataInfo(source.Path)
	//source := locations.NewSource(source.Path, interactions.NewInteraction().GamePart.GetGamePart())

	fileProcessor := fileFormats.NewFileProcessor(source, destination)
	if fileProcessor == nil {
		return TreeNode{}, nil
	}

	//dataInfo := fileProcessor.GetFileInfo()
	gameDataInfo := GameDataInfo{
		FilePath:  source.Get().Path,
		Extract:   *destination.Extract().Get(),
		Translate: *destination.Translate().Get(),
		Import:    *destination.Import().Get(),
	}

	var node = TreeNode{
		Key:   key,
		Label: source.Get().Name,
		Data:  gameDataInfo,
	}

	return node, nil
}
