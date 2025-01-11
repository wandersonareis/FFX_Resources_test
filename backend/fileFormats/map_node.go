package fileFormats

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
)

type (
	TreeMapNode = map[string]*MapNode

	DataInfo struct {
		FilePath      string                      `json:"file_path"`
		Source        core.SpiraFileInfo          `json:"source"`
		Extract       locations.ExtractLocation   `json:"extract_location"`
		Translate     locations.TranslateLocation `json:"translate_location"`
		FileProcessor interfaces.IFileProcessor
	}

	MapNode struct {
		Key       string   `json:"key"`
		Label     string   `json:"label"`
		Data      DataInfo `json:"data"`
		Icon      string   `json:"icon"`
		ChildKeys []string `json:"childKeys"`
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

func (mapNode *MapNode) SetNodeData(data DataInfo) {
	mapNode.Data = data
}

func (mapNode *MapNode) AddChildKey(key string) {
	mapNode.ChildKeys = append(mapNode.ChildKeys, key)
}
