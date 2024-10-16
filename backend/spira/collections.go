package spira

import (
	"ffxresources/backend/lib"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func ListFilesAndDirectories(source *lib.Source, prefix string) ([]lib.TreeNode, error) {
	var result []lib.TreeNode

	if !source.IsDir {
		return nil, nil
	}

	entries, err := source.ReadDir()
	if err != nil {
		return nil, err
	}

	for i, entry := range entries {
		entryPath := source.JoinEntryPath(entry)
		key := prefix + strconv.Itoa(i)

		if source.IsDir {
			entrySource, err := lib.NewSource(entryPath)
			if err != nil {
				return nil, err
			}

			children, err := ListFilesAndDirectories(entrySource, key+"-")
			if err != nil {
				return nil, err
			}

			node, err := CreateTreeNode(key, entrySource, children)
			if err != nil {
				return nil, err
			}

			result = append(result, node)
		} else {
			isSpira := lib.NewInteraction().GameLocation.IsSpiraPath(entryPath)
			if !isSpira {
				return nil, fmt.Errorf("invalid not spira path: %s", entryPath)
			}

			node, err := CreateTreeNode(key, source, nil)
			if err != nil {
				return nil, err
			}

			result = append(result, node)
		}
	}

	return result, nil
}

// BuildFileTree percorre o diretório especificado e constrói uma árvore de arquivos e diretórios.
func BuildFileTree(source *lib.Source, nodes *[]lib.TreeNode) error {
    if !source.IsDir {
        return nil // Se não for um diretório, retorna nil
    }

    // Mapa para armazenar nós temporariamente
    nodeMap := make(map[string]*lib.TreeNode)

    err := filepath.Walk(source.Path, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err // Retorna erro se houver
        }

        // Ignora o diretório raiz
        if path == source.Path {
            return nil
        }

        relativePath, err := filepath.Rel(source.Path, path)
        if err != nil {
            return err
        }

        key := relativePath // Usar o caminho relativo como chave
        //label := info.Name() // Usar o nome do arquivo/diretório como label

        // Cria um novo nó
        var node lib.TreeNode
        if info.IsDir() {
            // Se for um diretório, cria um novo Source
            entrySource, err := lib.NewSource(path)
            if err != nil {
                return err
            }

            // Cria um nó para o diretório usando CreateTreeNode
            node, err = CreateTreeNodeDev(key, entrySource)
            if err != nil {
                return err
            }
            node.Children = []lib.TreeNode{} // Inicializa como um slice vazio
        } else {
            // Se for um arquivo, verifica se é um caminho válido
            isSpira := lib.NewInteraction().GameLocation.IsSpiraPath(path)
            if !isSpira {
                return fmt.Errorf("invalid not spira path: %s", path)
            }

            // Cria um nó para o arquivo usando CreateTreeNode
            /* fileInfo := &lib.FileInfo{
                Name: info.Name(),
                Path: path,
            } */
            node, err = CreateTreeNodeDev(key, source)
            if err != nil {
                return err
            }
            //node.Data = fileInfo // Atribui o FileInfo ao nó
        }

        // Adiciona o nó ao mapa
        nodeMap[key] = &node

        // Adiciona o nó ao seu pai, se houver
        parentKey := filepath.Dir(key)
        if parentKey != "." {
            if parentNode, exists := nodeMap[parentKey]; exists {
                parentNode.Children = append(parentNode.Children, node)
            } else {
                // Se o pai ainda não existe, cria um nó pai vazio
                parentNode := lib.TreeNode{
                    Key:      parentKey,
                    Label:    filepath.Base(parentKey),
                    Children: []lib.TreeNode{node},
                }
                nodeMap[parentKey] = &parentNode
            }
        } else {
            // Se não houver pai, adiciona diretamente ao slice de nós
            *nodes = append(*nodes, node)
        }

        return nil
    })

    // Adiciona nós pai que não foram adicionados ao slice principal
    for _, node := range nodeMap {
        if node.Children == nil {
            *nodes = append(*nodes, *node)
        }
    }

    return err // Retorna o erro, se houver
}
