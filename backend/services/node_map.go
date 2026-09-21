package services

import (
	"path/filepath"

	"ffxresources/backend/fileFormats"
)

// NodeStore é um índice local de MapNodes usado apenas por operações de
// diretório (Extract/Compress de uma pasta): o nó da pasta é montado sob
// demanda e o store cobre a subárvore. Não há mais store global populado
// pela árvore completa de diretórios.
type NodeStore struct {
	nodes fileFormats.TreeMapNode
}

func NewNodeStore(nodes fileFormats.TreeMapNode) *NodeStore {
	return &NodeStore{nodes: nodes}
}

func (ns *NodeStore) Get(path string) (*fileFormats.MapNode, bool) {
	node, ok := ns.nodes[filepath.Clean(path)]
	return node, ok
}

func (ns *NodeStore) Len() int {
	return len(ns.nodes)
}

// IsNode valida o MapNode: nó, Data e Source não nulos.
func (ns *NodeStore) IsNode(node *fileFormats.MapNode) bool {
	if node == nil {
		return false
	}
	if node.Data == nil {
		return false
	}
	return true
}
