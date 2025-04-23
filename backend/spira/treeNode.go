package spira

import "ffxresources/backend/fileFormats"

type TreeNode struct {
	Key      string                   `json:"key"`
	Label    string                   `json:"label"`
	Data     *fileFormats.TreeNodeData `json:"data"`
	Icon     string                   `json:"icon"`
	Children []*TreeNode              `json:"children"`
}

func (treeNode *TreeNode) SetNodeKey(key string) {
	treeNode.Key = key
}

func (treeNode *TreeNode) SetNodeLabel(label string) {
	treeNode.Label = label
}

func (treeNode *TreeNode) SetNodeIcon(icon string) {
	treeNode.Icon = icon
}

func (treeNode *TreeNode) SetNodeData(data *fileFormats.DataInfo) {
	treeNode.Data.Source = data.Source
	treeNode.Data.Extract = data.Extract
	treeNode.Data.Translate = data.Translate
}

func (treeNode *TreeNode) AddNodeChild(child *TreeNode) {
	treeNode.Children = append(treeNode.Children, child)
}
