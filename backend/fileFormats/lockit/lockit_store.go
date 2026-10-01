package lockit

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ffxresources/backend/common"
)

// FileStore mantém os LockitFile carregados por chave canônica
// (<version>/<pattern>).
type FileStore struct {
	files map[string]*LockitFile
	// stamps registra, por chave, o carimbo físico dos binários de cada
	// idioma na carga (path/size/mtime). Permite detectar mudanças no disco
	// (tradução copiada/manual para mods/) e recarregar na próxima leitura.
	stamps map[string]map[string]fileStamp
}

// DataStore é o store global dos arquivos lockit.
var DataStore = NewFileStore()

// fileStamp identifica o estado físico de um binário carregado.
type fileStamp struct {
	path    string
	size    int64
	modTime time.Time
}

// NewFileStore constrói um store vazio.
func NewFileStore() *FileStore {
	return &FileStore{
		files:  map[string]*LockitFile{},
		stamps: map[string]map[string]fileStamp{},
	}
}

// Register publica um arquivo carregado. Sem carimbo: a entrada é confiada
// (o chamador responde pela frescura) e LoadFromStore não a recarrega.
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
	clear(s.stamps)
}

// stampOf devolve o carimbo registrado para a chave (nil = sem carimbo).
func (s *FileStore) stampOf(key string) map[string]fileStamp {
	if s == nil {
		return nil
	}
	return s.stamps[key]
}

// stampForResolve carimba o binário de cada idioma do layout na árvore
// preferida (mesma resolução da leitura: mods-first). Idioma ausente no disco
// não tem entrada no mapa — a carga o ignora da mesma forma.
func stampFor(l Layout) map[string]fileStamp {
	if len(l.Languages) == 0 {
		l.Languages = common.SupportedLanguageCodes()
	}
	out := make(map[string]fileStamp, len(l.Languages))
	for _, lang := range l.Languages {
		acc, err := common.NewFileAccessorFrom(filepath.FromSlash(l.RelPath(lang)), common.SourcePreferred)
		if err != nil || !acc.Exists {
			continue
		}
		out[lang] = fileStamp{
			path:    acc.ResolvedPath,
			size:    acc.Size,
			modTime: acc.Info.ModTime(),
		}
	}
	return out
}

// stampsEqual compara dois carimbos por idioma (path + size + mtime).
func stampsEqual(a, b map[string]fileStamp) bool {
	if len(a) != len(b) {
		return false
	}
	for k, sa := range a {
		if sb, ok := b[k]; !ok || sa != sb {
			return false
		}
	}
	return true
}

// LoadFromStore carrega (ou reutiliza) o arquivo do layout e o registra.
//
// A reutilização é vigiada: o carimbo físico (path/size/mtime por idioma) é
// recomparado a cada chamada — tradução copiada/editada manualmente em mods/
// (ou qualquer mudança nos binários) recarrega o arquivo na próxima leitura,
// sem reiniciar o app. O próprio save do app também muda o mods: a leitura
// seguinte recai na recarga e serve exatamente o que foi gravado.
func LoadFromStore(l Layout) (*LockitFile, error) {
	stamp := stampFor(l)
	if f, ok := DataStore.GetByLayout(l); ok && f != nil {
		stored := DataStore.stampOf(l.Key())
		// Sem carimbo (Register direto): entrada confiada, sem vigiação.
		if stored == nil || stampsEqual(stored, stamp) {
			return f, nil
		}
		// Binário mudou no disco (tradução copiada/editada em mods/):
		// descarta a instância e recarrega.
		common.LogVerbose("[lockit] %s: binário mudou no disco — recarregando", l.Stem)
	}
	f, err := Load(l)
	if err != nil {
		return nil, err
	}
	DataStore.Register(l.Key(), f)
	DataStore.stamps[l.Key()] = stamp
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
