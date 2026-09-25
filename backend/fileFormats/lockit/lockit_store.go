package lockit

import (
	"sort"
	"strings"

	"ffxresources/backend/common"
)

// FileStore mantém os LockitFile carregados por chave canônica
// (<version>/<pattern>).
type FileStore struct {
	files map[string]*LockitFile
}

// DataStore é o store global dos arquivos lockit.
var DataStore = NewFileStore()

// NewFileStore constrói um store vazio.
func NewFileStore() *FileStore {
	return &FileStore{files: map[string]*LockitFile{}}
}

// Register publica um arquivo carregado.
func (s *FileStore) Register(key string, f *LockitFile) {
	if s == nil || f == nil {
		return
	}
	s.files[key] = f
}

// Get recupera um arquivo carregado.
func (s *FileStore) Get(key string) (*LockitFile, bool) {
	if s == nil {
		return nil, false
	}
	f, ok := s.files[key]
	return f, ok
}

// GetByLayout recupera pelo layout.
func (s *FileStore) GetByLayout(l Layout) (*LockitFile, bool) {
	return s.Get(l.Key())
}

// Range itera os arquivos registrados.
func (s *FileStore) Range(fn func(key string, file *LockitFile)) {
	if s == nil || fn == nil {
		return
	}
	for k, v := range s.files {
		fn(k, v)
	}
}

// Clear limpa o store.
func (s *FileStore) Clear() {
	if s == nil {
		return
	}
	clear(s.files)
}

// LoadFromStore carrega (ou reutiliza) o arquivo do layout e o registra.
func LoadFromStore(l Layout) (*LockitFile, error) {
	if f, ok := DataStore.GetByLayout(l); ok && f != nil {
		return f, nil
	}
	f, err := Load(l)
	if err != nil {
		return nil, err
	}
	DataStore.Register(l.Key(), f)
	return f, nil
}

// LoadAllFromStore carrega todos os layouts da versão e devolve os arquivos.
func LoadAllFromStore(version common.GameVersion) ([]*LockitFile, error) {
	var out []*LockitFile
	for _, l := range LayoutsForVersion(version) {
		f, err := LoadFromStore(l)
		if err != nil {
			common.LogVerbose("[lockit] pular %s: %v", l.Stem, err)
			continue
		}
		out = append(out, f)
	}
	return out, nil
}

// KeyForID resolve a chave canônica a partir de um id (stem).
func KeyForID(version common.GameVersion, id string) (string, bool) {
	l, ok := LayoutForID(version, id)
	if !ok {
		return "", false
	}
	return l.Key(), true
}

// IsLockitKey informa se uma metadata.key pertence ao formato lockit.
func IsLockitKey(key string) bool {
	return strings.Contains(strings.ToLower(key), "/gamedata/ps3data/lockit/")
}

// SortedIDs devolve os ids (stems) dos layouts da versão, ordenados.
func SortedIDs(version common.GameVersion) []string {
	ls := LayoutsForVersion(version)
	ids := make([]string, 0, len(ls))
	for _, l := range ls {
		ids = append(ids, l.ID())
	}
	sort.Strings(ids)
	return ids
}
