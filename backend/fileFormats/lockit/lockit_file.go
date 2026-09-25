package lockit

import (
	"fmt"
	"os"
	"path/filepath"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	ffxencoding "ffxresources/backend/core/encoding"
)

// Record é um registro do lockit: a codificação de armazenamento (kind), o
// texto decodificado por idioma e os bytes originais de cada idioma.
//
// Os bytes originais são preservados: o encode do jogo tem representações
// alternativas para alguns slots (ex.: prefixo 0x04 vs lead 0x2F, mesmo
// caractere) e o re-encode só é usado quando o texto é de fato editado. Isso
// garante round-trip byte-a-byte dos registros não modificados.
type Record struct {
	kind  Kind
	text  map[string]string
	raw   map[string][]byte
	dirty map[string]bool
}

// Kind devolve a codificação de armazenamento do registro.
func (r *Record) Kind() Kind { return r.kind }

// Text devolve o texto do idioma ("" se ausente).
func (r *Record) Text(lang string) string { return r.text[lang] }

// SetText define o texto de um idioma, marcando-o para re-encode quando difere
// do texto originalmente decodificado.
func (r *Record) SetText(lang, s string) {
	if r.text[lang] == s {
		return
	}
	r.text[lang] = s
	if r.dirty == nil {
		r.dirty = make(map[string]bool)
	}
	r.dirty[lang] = true
}

// TextMap devolve o mapa de textos por idioma (nunca nil).
func (r *Record) TextMap() map[string]string { return r.text }

// encode devolve os bytes do registro para um idioma: os originais quando o
// texto não foi editado, senão o re-encode.
func (r *Record) encode(lang string, version common.GameVersion) ([]byte, error) {
	if !r.dirty[lang] {
		if raw, ok := r.raw[lang]; ok {
			return raw, nil
		}
	}
	return encodeRecord(r.text[lang], r.kind, lang, version)
}

// LockitFile é o conjunto lógico de registros de uma versão, agregando todos
// os idiomas disponíveis em rows alinhadas por posição física.
type LockitFile struct {
	layout     Layout
	terminated bool
	langs      []string
	records    []*Record
}

// Layout devolve o layout de origem.
func (f *LockitFile) Layout() Layout { return f.layout }

// Records devolve os registros em ordem física.
func (f *LockitFile) Records() []*Record { return f.records }

// Len devolve o número de registros.
func (f *LockitFile) Len() int { return len(f.records) }

// Languages devolve os idiomas efetivamente carregados (na ordem do layout).
func (f *LockitFile) Languages() []string { return f.langs }

// HasLanguage informa se o idioma foi carregado.
func (f *LockitFile) HasLanguage(lang string) bool {
	for _, l := range f.langs {
		if l == lang {
			return true
		}
	}
	return false
}

// Terminated informa se o arquivo termina com quebra de linha.
func (f *LockitFile) Terminated() bool { return f.terminated }

// IndexesByKind devolve os índices físicos dos registros game e utf8, cada um
// em ordem física. É a base da numeração do DTO (game 0..G-1, utf8 G..G+U-1).
func (f *LockitFile) IndexesByKind() (game, utf8 []int) {
	for i, r := range f.records {
		if r.kind == KindGame {
			game = append(game, i)
		} else {
			utf8 = append(utf8, i)
		}
	}
	return game, utf8
}

// Load lê os arquivos de todos os idiomas disponíveis do layout e monta o
// conjunto lógico de registros. Idiomas ausentes são simplesmente ignorados.
func Load(l Layout) (*LockitFile, error) {
	if len(l.Languages) == 0 {
		l.Languages = common.SupportedLanguageCodes()
	}
	if err := ffxencoding.PrepareVersionCharsets(common.CharsetVersion(l.Version)); err != nil {
		return nil, fmt.Errorf("lockit %s: preparar charsets: %w", l.Stem, err)
	}

	rawByLang := make(map[string][][]byte, len(l.Languages))
	var langs []string
	n := -1
	terminated := true

	for _, lang := range l.Languages {
		data, err := readFile(l.RelPath(lang))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("lockit %s: ler %s: %w", l.Stem, lang, err)
		}
		recs, term := splitRecords(data)
		if n == -1 {
			n = len(recs)
			terminated = term
		} else if len(recs) != n {
			return nil, fmt.Errorf("lockit %s: %s tem %d registros, esperado %d", l.Stem, lang, len(recs), n)
		}
		rawByLang[lang] = recs
		langs = append(langs, lang)
	}
	if len(langs) == 0 {
		return nil, fmt.Errorf("lockit %s: nenhum arquivo de idioma encontrado", l.Stem)
	}

	kinds := detectKinds(rawByLang, n)
	records := make([]*Record, n)
	for i := 0; i < n; i++ {
		rec := &Record{
			kind: kinds[i],
			text: make(map[string]string, len(langs)),
			raw:  make(map[string][]byte, len(langs)),
		}
		for _, lang := range langs {
			rec.raw[lang] = rawByLang[lang][i]
			rec.text[lang] = decodeRecord(rawByLang[lang][i], kinds[i], lang, l.Version)
		}
		records[i] = rec
	}

	return &LockitFile{layout: l, terminated: terminated, langs: langs, records: records}, nil
}

// Bytes codifica o arquivo de cada idioma carregado (lang -> bytes), sem gravar.
func (f *LockitFile) Bytes() (map[string][]byte, error) {
	out := make(map[string][]byte, len(f.langs))
	for _, lang := range f.langs {
		encoded := make([][]byte, len(f.records))
		for i, rec := range f.records {
			b, err := rec.encode(lang, f.layout.Version)
			if err != nil {
				return nil, fmt.Errorf("lockit %s: codificar %s registro %d: %w", f.layout.Stem, lang, i, err)
			}
			encoded[i] = b
		}
		out[lang] = joinRecords(encoded, f.terminated)
	}
	return out, nil
}

// Save grava cada idioma carregado no binário (mods-first quando habilitado).
func (f *LockitFile) Save() error {
	return f.SaveLanguages(f.langs...)
}

// SaveLanguages grava apenas os idiomas pedidos no binário (mods-first quando
// habilitado). O import usa apenas a localização padrão (us), como nos demais
// formatos; os outros idiomas do lockit são referência de tradução.
func (f *LockitFile) SaveLanguages(langs ...string) error {
	byLang, err := f.Bytes()
	if err != nil {
		return err
	}
	for _, lang := range langs {
		if !f.HasLanguage(lang) {
			continue
		}
		if err := writeFile(f.layout.RelPath(lang), byLang[lang]); err != nil {
			return fmt.Errorf("lockit %s: gravar %s: %w", f.layout.Stem, lang, err)
		}
	}
	return nil
}

// decodeRecord decodifica um registro bruto conforme a codificação.
func decodeRecord(raw []byte, kind Kind, lang string, version common.GameVersion) string {
	if kind == KindGame {
		return converter.BytesToString(raw, lang, version)
	}
	return string(raw)
}

// encodeRecord codifica um texto conforme a codificação.
func encodeRecord(text string, kind Kind, lang string, version common.GameVersion) ([]byte, error) {
	if kind == KindGame {
		return converter.StringToBytes(text, LanguageCharset(lang), version)
	}
	return []byte(text), nil
}

// readFile lê um arquivo relativo ao GameFilesRoot, preferindo mods (igual ao
// restante do fluxo). Retorna os.ErrNotExist quando o idioma não existe.
func readFile(rel string) ([]byte, error) {
	accessor, err := common.NewFileAccessor(filepath.FromSlash(rel))
	if err != nil {
		return nil, err
	}
	if !accessor.Exists {
		return nil, os.ErrNotExist
	}
	return os.ReadFile(accessor.ResolvedPath)
}

// writeFile grava um arquivo relativo ao GameFilesRoot, no mods quando
// habilitado (para não sobrescrever o original) e no original caso contrário.
func writeFile(rel string, data []byte) error {
	base := common.GameFilesRoot
	if common.AreModsEnabled() {
		base = filepath.Join(base, common.ModsFolder)
	}
	full := filepath.Join(base, filepath.FromSlash(rel))
	if err := common.EnsurePathExists(full); err != nil {
		return err
	}
	return common.WriteBytesToFile(full, data)
}
