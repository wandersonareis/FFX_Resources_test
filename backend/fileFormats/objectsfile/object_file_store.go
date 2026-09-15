package objectsfile

import (
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

// ObjectFileStore mantém instâncias de arquivos binários de objetos carregadas
// por key (versão + patternPath), espelhando o NodeDataStore.
type ObjectFileStore struct {
	files map[string]datastore.IBinaryFile
}

var ObjectFileDataStore = NewObjectFileStore()

func NewObjectFileStore() *ObjectFileStore {
	return &ObjectFileStore{files: map[string]datastore.IBinaryFile{}}
}

func (s *ObjectFileStore) Register(key string, f datastore.IBinaryFile) {
	if s == nil || f == nil {
		return
	}
	s.files[key] = f
}

func (s *ObjectFileStore) Get(key string) (datastore.IBinaryFile, bool) {
	if s == nil {
		return nil, false
	}
	f, ok := s.files[key]
	return f, ok
}

func (s *ObjectFileStore) GetByVersion(version common.GameVersion, patternPath string) (datastore.IBinaryFile, bool) {
	return s.Get(FileLayoutKey(version, patternPath))
}

func (s *ObjectFileStore) All() map[string]datastore.IBinaryFile {
	if s == nil {
		return nil
	}
	out := make(map[string]datastore.IBinaryFile, len(s.files))
	for k, v := range s.files {
		out[k] = v
	}
	return out
}

func (s *ObjectFileStore) Range(fn func(key string, file datastore.IBinaryFile)) {
	if s == nil || fn == nil {
		return
	}
	for k, v := range s.files {
		fn(k, v)
	}
}

func (s *ObjectFileStore) Len() int {
	if s == nil {
		return 0
	}
	return len(s.files)
}