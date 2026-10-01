package objectsfile

import (
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

type ObjectFileStore struct {
	files map[string]datastore.IBinaryFile
	// stamps registra, por chave, o carimbo físico do binário na carga
	// (path/size/mtime). Permite detectar binário modificado no disco
	// (tradução copiada/manual para mods/) e recarregar na próxima leitura.
	stamps map[string]common.FileStamp
}

var ObjectFileDataStore = NewObjectFileStore()

func NewObjectFileStore() *ObjectFileStore {
	return &ObjectFileStore{
		files:  map[string]datastore.IBinaryFile{},
		stamps: map[string]common.FileStamp{},
	}
}

func (s *ObjectFileStore) Register(key string, f datastore.IBinaryFile) {
	if s == nil || f == nil {
		return
	}
	s.files[key] = f
}

// RegisterWithStamp publica o arquivo com o carimbo físico vigiado.
func (s *ObjectFileStore) RegisterWithStamp(key string, f datastore.IBinaryFile, stamp common.FileStamp) {
	if s == nil || f == nil {
		return
	}
	s.files[key] = f
	s.stamps[key] = stamp
}

// StampOf devolve o carimbo registrado (ok=false = sem carimbo/vigia).
func (s *ObjectFileStore) StampOf(key string) (common.FileStamp, bool) {
	if s == nil {
		return common.FileStamp{}, false
	}
	stamp, ok := s.stamps[key]
	return stamp, ok
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

// GetByLayout busca pelo layout (versão + dir + arquivo) em vez das
// primitivas separadas.
func (s *ObjectFileStore) GetByLayout(l FileLayout) (datastore.IBinaryFile, bool) {
	return s.Get(FileLayoutKey(l.Version, l.PatternPath()))
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

func (s *ObjectFileStore) IsEmpty() bool {
	return s.Len() == 0
}

func (s *ObjectFileStore) Clear() {
	if s == nil {
		return
	}
	clear(s.files)
	clear(s.stamps)
}
