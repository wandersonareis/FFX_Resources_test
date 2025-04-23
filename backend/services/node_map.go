package services

import "maps"

import "ffxresources/backend/fileFormats"

type NodeStore struct {
	nodes fileFormats.TreeMapNode
}

var NodeDataStore *NodeStore

func NewNodeStore(nodes fileFormats.TreeMapNode) *NodeStore {
	cloned := maps.Clone(nodes)
	return &NodeStore{nodes: cloned}
}


func (ns *NodeStore) Get(path string) (*fileFormats.MapNode, bool) {
	node, ok := ns.nodes[path]
	return node, ok
}

func (ns *NodeStore) Len() int {
	return len(ns.nodes)
}

// IsNode checks whether the provided MapNode is valid by ensuring that the node itself,
// its Data field, and the Data's Source field are all non-nil.
// Returns true if all checks pass, otherwise returns false.
func (ns *NodeStore) IsNode(node *fileFormats.MapNode) bool {
	if node == nil {
		return false
	}

	if node.Data == nil {
		return false
	}

	if node.Data.Source == nil {
		return false
	}
	
	return true
}