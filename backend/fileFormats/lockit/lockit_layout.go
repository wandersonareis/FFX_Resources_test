// Package lockit trata os arquivos "loc kit" (por exemplo,
// ffx_data/gamedata/ps3data/lockit/ffx_loc_kit_ps3_us.bin) — um conjunto de
// registros de texto que mistura duas codificações no MESMO arquivo:
//
//   - game: bytes no charset customizado do jogo (core/encoding), cuja tabela
//     depende do idioma (us/default/jp/kr/ch);
//   - utf8: texto UTF-8 puro (launcher/PC).
//
// O arquivo é separado por CRLF em registros e existe UMA variante por idioma
// (<stem>_<lang>.bin). A estrutura (qual registro é game e qual é utf8) é
// fixa por versão e idêntica entre idiomas — por isso a extração agrupa tudo
// em um único arquivo lógico, com todos os idiomas por row.
//
// O texto flui como DTO (backend/dto): este pacote lê/escreve o binário e os
// builders (backend/builders) montam/aplicam o DTO. O pacote desconhece
// JSON/.strings, igual a objectsfile/macrodic.
package lockit

import (
	"path/filepath"
	"sort"
	"strings"

	"ffxresources/backend/common"
)

// Kind é a codificação de armazenamento de um registro.
type Kind uint8

const (
	// KindGame é um registro no charset customizado do jogo.
	KindGame Kind = iota
	// KindUTF8 é um registro em UTF-8 puro.
	KindUTF8
)

// String devolve o rótulo usado no campo Name do DTO.
func (k Kind) String() string {
	switch k {
	case KindGame:
		return "game"
	case KindUTF8:
		return "utf8"
	default:
		return "unknown"
	}
}

// KindFromName converte o rótulo do DTO de volta em Kind.
func KindFromName(name string) (Kind, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "game":
		return KindGame, true
	case "utf8":
		return KindUTF8, true
	}
	return 0, false
}

// Layout descreve os arquivos físicos de uma versão: o diretório, o prefixo
// (stem) e os idiomas que podem existir. Cada idioma é um arquivo
// <Dir>/<Stem>_<lang>.bin.
type Layout struct {
	Version   common.GameVersion
	Dir       string   // relativo ao GameFilesRoot, com "/" (ex.: ffx_data/gamedata/ps3data/lockit)
	Stem      string   // ex.: ffx_loc_kit_ps3
	Languages []string // idiomas que podem existir (us/jp/kr/ch/de/fr/it/sp)
}

// PatternPath é o caminho canônico usado na metadata.key/id.
func (l Layout) PatternPath() string {
	return "gamedata/ps3data/lockit/" + l.Stem + ".bin"
}

// Key é a chave canônica da Collection (<version>/<pattern>).
func (l Layout) Key() string {
	return common.VersionPathName(l.Version) + "/" + l.PatternPath()
}

// ID é a chave da Collection (stem, sem extensão).
func (l Layout) ID() string { return l.Stem }

// FileName é o nome lógico do arquivo (sem idioma).
func (l Layout) FileName() string { return l.Stem + ".bin" }

// RelPath devolve o caminho físico de um idioma, relativo ao GameFilesRoot.
func (l Layout) RelPath(lang string) string {
	return filepath.ToSlash(filepath.Join(l.Dir, l.Stem+"_"+lang+".bin"))
}

// layouts é o registro declarativo por versão. Apenas FFX e FFX-2 têm lockit
// (LastMiss divide a árvore do FFX-2, mas não tem kit próprio).
var layouts = map[string]Layout{
	"ffx/gamedata/ps3data/lockit/ffx_loc_kit_ps3.bin": {
		Version: common.GameVersionFFX,
		Dir:     "ffx_data/gamedata/ps3data/lockit",
		Stem:    "ffx_loc_kit_ps3",
	},
	"ffx2/gamedata/ps3data/lockit/ffx2_loc_kit_ps3.bin": {
		Version: common.GameVersionFFX2,
		Dir:     "ffx-2_data/gamedata/ps3data/lockit",
		Stem:    "ffx2_loc_kit_ps3",
	},
}

// AllLayouts devolve os layouts registrados (ordem determinística).
func AllLayouts() []Layout {
	keys := make([]string, 0, len(layouts))
	for k := range layouts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]Layout, 0, len(keys))
	for _, k := range keys {
		l := layouts[k]
		if len(l.Languages) == 0 {
			l.Languages = common.SupportedLanguageCodes()
		}
		out = append(out, l)
	}
	return out
}

// LayoutsForVersion devolve os layouts de uma versão (match exato: LastMiss
// não tem lockit próprio, embora divida a árvore com o FFX-2).
func LayoutsForVersion(version common.GameVersion) []Layout {
	var out []Layout
	for _, l := range AllLayouts() {
		if l.Version == version {
			out = append(out, l)
		}
	}
	return out
}

// LayoutFor resolve um layout por versão + key/pattern (aceita com ou sem o
// prefixo de versão).
func LayoutFor(version common.GameVersion, pattern string) (Layout, bool) {
	clean := strings.Trim(filepath.ToSlash(strings.TrimSpace(pattern)), "/")
	if clean == "" {
		return Layout{}, false
	}
	key := clean
	if !strings.HasPrefix(key, "ffx/") && !strings.HasPrefix(key, "ffx2/") {
		key = common.VersionPathName(version) + "/" + key
	}
	l, ok := layouts[key]
	if !ok || l.Version != version {
		return Layout{}, false
	}
	if len(l.Languages) == 0 {
		l.Languages = common.SupportedLanguageCodes()
	}
	return l, true
}

// LayoutForID resolve um layout pelo id (stem) numa versão.
func LayoutForID(version common.GameVersion, id string) (Layout, bool) {
	clean := strings.TrimSpace(id)
	for _, l := range LayoutsForVersion(version) {
		if strings.EqualFold(l.ID(), clean) {
			return l, true
		}
	}
	return Layout{}, false
}

// LanguageCharset devolve o charset do jogo para um idioma (us/default/jp/kr/ch).
func LanguageCharset(lang string) string {
	return common.LanguageCodeToCharset(lang)
}
