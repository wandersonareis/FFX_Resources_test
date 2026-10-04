// Package eventtable trata os binários do FFX-2 que usam a MESMA tabela de
// strings dos eventos (stride 8 por entrada: regOff u16 + flags u8 +
// choices u8, simpOff u16 + flags u8 + choices u8; first u16 = tamanho da
// tabela; offsets relativos ao início do arquivo).
//
// Famílias cobertas (todas com um binário por localização):
//
//   - textos de batalha: battle/btl/<id>/<id>.bin (kind "battletext")
//   - cloudsave:        cloudsave/cloud.bin, cloudsave/cloudv.bin
//     (kind "cloud", ids "cloud", "cloudv")
//   - tutorial:          menu/tutorial.msb (kind "tutorial", id "tutorial")
//
// O codec (FieldString) é compartilhado com o pacote event; este pacote só
// acrescenta a resolução de caminho por família e o store. Os builders
// (backend/builders) montam o DTO, os formatters serializam — como nos
// demais formatos, este pacote desconhece JSON/.strings.
package eventtable

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/event"
)

// Kinds servidos ao frontend (mesmos valores do DTO/entryKindsFor).
const (
	KindBattleText = "battletext"
	KindCloud      = "cloud"
	KindTutorial   = "tutorial"
	// KindMenuMain é o texto do menu principal do FFX (menu/menumain.bin,
	// FFX-only — o FFX-2 reorganizou essas telas).
	KindMenuMain = "menumain"
)

// fixedIDs são os artefatos estáticos por família (battletext é o único com
// descoberta dinâmica em diretório).
var fixedIDs = map[string][]string{
	KindCloud:     {"cloud"},
	KindTutorial:  {"tutorial"},
	KindMenuMain:  {"menumain"},
}

// RelPath devolve o caminho do binário de strings, relativo à raiz de
// localização da versão (ex.: battle/btl/bika07_228/bika07_228.bin).
// ok=false para kind desconhecido ou id fora da família.
func RelPath(kind, id string) (string, bool) {
	switch kind {
	case KindBattleText:
		id = strings.TrimSpace(id)
		if id == "" || strings.ContainsAny(id, `/\`) {
			return "", false
		}
		return filepath.ToSlash(filepath.Join("battle", "btl", id, id+".bin")), true
	case KindCloud:
		// cloud.bin + cloudv.bin são um único artefato (id "cloud"): o par é
		// sempre lido/exportado/importado JUNTO — o dedup ($hash) entre os dois
		// arquivos só funciona num artefato único, e o Save recompõe os dois
		// binários. Bins com contagens iguais por idioma (régua events).
		switch strings.ToLower(strings.TrimSpace(id)) {
		case "cloud", "cloudsave":
			return "cloudsave/cloud.bin", true
		}
		return "", false
	case KindTutorial:
		if strings.EqualFold(strings.TrimSpace(id), "tutorial") {
			return "menu/tutorial.msb", true
		}
		return "", false
	case KindMenuMain:
		if strings.EqualFold(strings.TrimSpace(id), "menumain") {
			return "menu/menumain.bin", true
		}
		return "", false
	}
	return "", false
}

// RelPaths lista todos os bins físicos de um artefato (cloud = par cloud+cloudv).
func RelPaths(kind, id string) []string {
	if kind == KindCloud {
		if id == "cloud" || id == "cloudsave" {
			return []string{"cloudsave/cloud.bin", "cloudsave/cloudv.bin"}
		}
		return nil
	}
	if rel, ok := RelPath(kind, id); ok {
		return []string{rel}
	}
	return nil
}

// Key devolve a chave canônica da metadata (<version>/<pattern>).
func Key(version common.GameVersion, kind, id string) (string, bool) {
	rel, ok := RelPath(kind, id)
	if !ok {
		return "", false
	}
	return common.VersionPathName(version) + "/" + rel, true
}

// RelPathLoc devolve o caminho relativo ao GameFilesRoot para um idioma:
// <raiz de localização da versão>/<pattern>.
func RelPathLoc(version common.GameVersion, loc, kind, id string) (string, bool) {
	rel, ok := RelPath(kind, id)
	if !ok {
		return "", false
	}
	return filepath.ToSlash(filepath.Join(common.GetLocalizationRootForVersion(version, loc), rel)), true
}

// IDs descobre os ids disponíveis da família na árvore original (data/).
// Cloud/tutorial têm lista estática; battletext varre battle/btl/<dir>/<dir>.bin
// na localização padrão da versão (existe em FFX e FFX-2).
func IDs(kind string, version common.GameVersion) []string {
	if ids, ok := fixedIDs[kind]; ok {
		return append([]string(nil), ids...)
	}
	if kind != KindBattleText {
		return nil
	}
	scanDir := filepath.Join(common.GameFilesRoot,
		common.GetLocalizationRootForVersion(version, common.DefaultLocalization),
		"battle", "btl")
	entries, err := os.ReadDir(scanDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		bin := filepath.Join(scanDir, e.Name(), e.Name()+".bin")
		if _, err := os.Stat(bin); err == nil {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// BinSpec descreve um dos binários físicos de um artefato: quais strings (por
// faixa de File.Strings) pertencem a ele. Cloud funde cloud.bin+cloudv.bin
// para dedupe crescer junto; o Save reescreve um por BinSpec.
type BinSpec struct {
	Name  string // "cloud", "cloudv" ou o Nome do artefato
	Rel   string // caminho relativo à raiz de localização
	Start int
	Len   int
}

// File é um artefato carregado: as strings por posição, com conteúdo por
// localização (como os eventos).
type File struct {
	Kind    string
	ID      string
	Version common.GameVersion
	Strings []*event.LocalizedFieldStringObject
	Bins    []BinSpec
}

// Load lê todos os idiomas disponíveis do artefato (resolução mods-first)
// e devolve o File lógico. Não registra no store.
func Load(kind string, version common.GameVersion, id string) (*File, error) {
	if kind == KindCloud {
		return loadCloud(version)
	}
	rel, ok := RelPath(kind, id)
	if !ok {
		return nil, fmt.Errorf("%s: id desconhecido %q", kind, id)
	}
	strings_ := event.ReadLocalizedStringFilesFrom(rel, version, common.SourcePreferred)
	if len(strings_) == 0 {
		return nil, fmt.Errorf("%s: %s sem localizações", kind, rel)
	}
	return &File{Kind: kind, ID: id, Version: version, Strings: strings_, Bins: []BinSpec{{Name: id, Rel: rel, Start: 0, Len: len(strings_)}}}, nil
}

// loadCloud lê cloud.bin + cloudv.bin juntos (mesmo artefato), preservando
// a faixa de cada um para o Save reescrever os dois binários.
func loadCloud(version common.GameVersion) (*File, error) {
	cloud := event.ReadLocalizedStringFilesFrom("cloudsave/cloud.bin", version, common.SourcePreferred)
	cloudv := event.ReadLocalizedStringFilesFrom("cloudsave/cloudv.bin", version, common.SourcePreferred)
	if len(cloud) == 0 && len(cloudv) == 0 {
		return nil, fmt.Errorf("cloudsave: nenhum conteúdo")
	}
	all := make([]*event.LocalizedFieldStringObject, 0, len(cloud)+len(cloudv))
	all = append(all, cloud...)
	all = append(all, cloudv...)
	return &File{
		Kind:    KindCloud,
		ID:      "cloud",
		Version: version,
		Strings: all,
		Bins: []BinSpec{
			{Name: "cloud", Rel: "cloudsave/cloud.bin", Start: 0, Len: len(cloud)},
			{Name: "cloudv", Rel: "cloudsave/cloudv.bin", Start: len(cloud), Len: len(cloudv)},
		},
	}, nil
}

// ReadLocalizedStringsFrom lê as strings direto da árvore indicada, sem
// tocar no store (é o caminho do ORIGINAL / .vbf).
func ReadLocalizedStringsFrom(kind string, id string, version common.GameVersion, src common.FileSource) ([]*event.LocalizedFieldStringObject, error) {
	if kind == KindCloud {
		cloud := event.ReadLocalizedStringFilesFrom("cloudsave/cloud.bin", version, src)
		cloudv := event.ReadLocalizedStringFilesFrom("cloudsave/cloudv.bin", version, src)
		all := append(append([]*event.LocalizedFieldStringObject{}, cloud...), cloudv...)
		if len(all) == 0 {
			return nil, fmt.Errorf("cloudsave: nenhum conteúdo (%s)", src)
		}
		return all, nil
	}
	rel, ok := RelPath(kind, id)
	if !ok {
		return nil, fmt.Errorf("%s: id desconhecido %q", kind, id)
	}
	strings_ := event.ReadLocalizedStringFilesFrom(rel, version, src)
	if len(strings_) == 0 {
		return nil, fmt.Errorf("%s: %s sem conteúdo para %s", kind, id, src)
	}
	return strings_, nil
}

// Save grava TODAS as localizações do File na árvore mods/ (o original do
// gamefiles fica intacto), no mesmo padrão dos eventos.
func (f *File) Save() error {
	if len(f.Strings) == 0 {
		return nil
	}
	bins := f.Bins
	if len(bins) == 0 {
		rel, ok := RelPath(f.Kind, f.ID)
		if !ok {
			return fmt.Errorf("%s: id desconhecido %q", f.Kind, f.ID)
		}
		bins = []BinSpec{{Name: f.ID, Rel: rel, Start: 0, Len: len(f.Strings)}}
	}
	for _, loc := range common.SupportedLanguageCodes() {
		for _, bin := range bins {
			buf, err := event.EncodeLocalizedStrings(f.Strings[bin.Start:bin.Start+bin.Len], loc, f.Version)
			if err != nil {
				return fmt.Errorf("%s: codificar %s/%s/%s: %w", f.Kind, f.ID, bin.Name, loc, err)
			}
			localeRel := filepath.ToSlash(filepath.Join(common.GetLocalizationRootForVersion(f.Version, loc), bin.Rel))
			target := filepath.Join(common.GameFilesRoot, common.ModsFolder, filepath.FromSlash(localeRel))
			if err := common.EnsurePathExists(filepath.Dir(target)); err != nil {
				return fmt.Errorf("%s: %w", f.Kind, err)
			}
			if err := common.WriteBytesToFile(target, buf); err != nil {
				return fmt.Errorf("%s: gravar %s: %w", f.Kind, target, err)
			}
		}
	}
	return nil
}

// ---- store -------------------------------------------------------------------

var (
	storeMu sync.Mutex
	store   = map[string]*File{} // key: kind/version/id
	stamps  = map[string]map[string]common.FileStamp{}
)

func keyOf(kind string, version common.GameVersion, id string) string {
	return kind + "/" + common.VersionPathName(version) + "/" + id
}

// stampFor carimba o binário de cada idioma na árvore preferida (mods-first).
func stampFor(kind string, version common.GameVersion, id string) map[string]common.FileStamp {
	out := make(map[string]common.FileStamp, len(common.SupportedLanguageCodes()))
	for _, loc := range common.SupportedLanguageCodes() {
		for _, rel := range RelPaths(kind, id) {
			full := filepath.ToSlash(filepath.Join(common.GetLocalizationRootForVersion(version, loc), rel))
			if stamp, ok := common.StampFile(filepath.FromSlash(full)); ok {
				out[loc] = stamp
				break
			}
		}
	}
	return out
}

func stampsEqual(a, b map[string]common.FileStamp) bool {
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

// LoadFromStore carrega (ou reutiliza) o artefato com vigia de carimbo:
// binário modificado no disco (tradução copiada para mods/) recarrega na
// próxima leitura. É o caminho de GetCollection — igual ao lockit.
func LoadFromStore(kind string, version common.GameVersion, id string) (*File, error) {
	k := keyOf(kind, version, id)
	stamp := stampFor(kind, version, id)
	storeMu.Lock()
	f := store[k]
	stored := stamps[k]
	storeMu.Unlock()
	if f != nil && (stored == nil || stampsEqual(stored, stamp)) {
		return f, nil
	}
	if f != nil {
		common.LogVerbose("[eventtable] %s/%s: binário mudou no disco — recarregando", kind, id)
	}
	loaded, err := Load(kind, version, id)
	if err != nil {
		return nil, err
	}
	storeMu.Lock()
	store[k] = loaded
	stamps[k] = stamp
	storeMu.Unlock()
	return loaded, nil
}

// Get devolve o File sem carregar (nil quando ausente).
func Get(kind string, version common.GameVersion, id string) *File {
	storeMu.Lock()
	defer storeMu.Unlock()
	return store[keyOf(kind, version, id)]
}

// ClearStore esvazia a store (higiene ao trocar GameFilesLocation/testes).
func ClearStore() {
	storeMu.Lock()
	defer storeMu.Unlock()
	store = map[string]*File{}
	stamps = map[string]map[string]common.FileStamp{}
}

// List devolve os ids descobertos da família (ordenados).
func List(kind string, version common.GameVersion) []string {
	ids := IDs(kind, version)
	sort.Strings(ids)
	return ids
}

// Validate garante que o kind é servido por este pacote.
func Validate(kind string) error {
	switch kind {
	case KindBattleText, KindCloud, KindTutorial, KindMenuMain:
		return nil
	}
	return fmt.Errorf("eventtable: kind desconhecido %q", kind)
}
