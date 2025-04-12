package services

import "maps"

import "ffxresources/backend/fileFormats"

type NodeStore struct {
	nodes fileFormats.TreeMapNode
}

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