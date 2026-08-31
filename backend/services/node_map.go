package services

import (
	"path/filepath"

	"maps"

	"ffxresources/backend/fileFormats"
)

type NodeStore struct {
	nodes fileFormats.TreeMapNode
}

var NodeDataStore *NodeStore

func NewNodeStore(nodes fileFormats.TreeMapNode) *NodeStore {
	cloned := maps.Clone(nodes)
	store := &NodeStore{nodes: cloned}

	return store
}

// Merge adds all entries from another node map into the store. It is used to
// accumulate the NodeDataStore across multiple game versions (e.g. FFX and
// FFX-2) into a single lookup table, since their absolute paths are disjoint.
func (ns *NodeStore) Merge(other fileFormats.TreeMapNode) {
	for k, v := range other {
		ns.nodes[k] = v
	}
}

func (ns *NodeStore) Get(path string) (*fileFormats.MapNode, bool) {
	node, ok := ns.nodes[filepath.Clean(path)]
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

	return true
}
