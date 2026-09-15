package objectsfile

import (
	"fmt"
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

// LoadObjectFile constrói, carrega e registra um ObjectBinaryFile a partir de um FileLayout.
// Se o layout tiver Creator (ex: weapons), usa o creator; caso contrário monta um
// KeyedStringFile genérico com o layout e formatter fornecidos.
func LoadObjectFile(l FileLayout) (datastore.IBinaryFile, error) {
	patternPath := l.PatternPath()
	var creator CreatorFunc
	if l.Creator != nil {
		creator = l.Creator
	} else {
		layouts := LayoutSet{l.Version: l.Fields}
		creator = func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
			return NewKeyedStringFileAt(cBytes, sBytes, hLen, lang, l.Version, layouts, "", l.Start)
		}
	}

	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		creator,
		common.DefaultLocalization,
		l.Version,
	)
	if err := binaryDataFile.LoadFromBinary(); err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", patternPath, err)
	}

	if objects := binaryDataFile.GetObjects(); objects != nil && !objects.IsEmpty() {
		if l.IndexCount > 0 && objects.Len() != l.IndexCount {
			return nil, fmt.Errorf("index count mismatch for %s: expected %d, got %d", patternPath, l.IndexCount, objects.Len())
		}
		datastore.Commands = objects
	}

	return binaryDataFile, nil
}

// LoadObjectFileFromStore busca o layout pelo key (versão + patternPath), carrega o
// arquivo e registra a instância no ObjectFileDataStore, retornando-a.
func LoadObjectFileFromStore(version common.GameVersion, patternPath string) (datastore.IBinaryFile, error) {
	key := FileLayoutKey(version, patternPath)
	layout, ok := FileLayoutFor(version, patternPath)
	if !ok {
		return nil, fmt.Errorf("no layout registered for key %s", key)
	}
	binFile, err := LoadObjectFile(layout)
	if err != nil {
		return nil, err
	}
	ObjectFileDataStore.Register(key, binFile)
	return binFile, nil
}