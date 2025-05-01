package fileFormats

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
)

type (
	TreeMapNode = map[string]*MapNode

	TreeNodeData struct {
		Source    *core.SpiraFileInfo          `json:"source"`
		Extract   locations.IExtractLocation   `json:"extract_location"`
		Translate locations.ITranslateLocation `json:"translate_location"`
	}

	DataInfo struct {
		Source    *core.SpiraFileInfo          `json:"source"`
		Extract   locations.IExtractLocation   `json:"extract_location"`
		Translate locations.ITranslateLocation `json:"translate_location"`

		FileProcessor interfaces.IFileProcessor
	}

	node struct {
		Key       string   `json:"key"`
		Label     string   `json:"label"`
		Icon      string   `json:"icon"`
		ChildKeys []string `json:"childKeys"`
	}
	MapNode struct {
		node
		Data *DataInfo `json:"data"`
	}

	TreeNode struct {
		node
		DataInfo TreeNodeData `json:"data"`
	}
)

func (mapNode *MapNode) SetNodeKey(key string) {
	mapNode.Key = key
}

func (mapNode *MapNode) SetNodeLabel(label string) {
	mapNode.Label = label
}

func (mapNode *MapNode) SetNodeIcon(icon string) {
	mapNode.Icon = icon
}

func (mapNode *MapNode) SetNodeData(data *DataInfo) {
	mapNode.Data = data
}

func (mapNode *MapNode) AddChildKey(key string) {
	mapNode.ChildKeys = append(mapNode.ChildKeys, key)
}
