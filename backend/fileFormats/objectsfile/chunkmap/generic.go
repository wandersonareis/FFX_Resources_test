package chunkmap

import (
	"fmt"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/formatters/hash"
)

// SchemaLoader carrega um arquivo binario e devolve o DTO pronto.
type SchemaLoader func(version common.GameVersion, patternPath string) (dto.Collection, error)

// SchemaSourceLoader carrega de uma FONTE especifica (SourceData para o
// original de data/, SourcePreferred para mods-first).
type SchemaSourceLoader func(version common.GameVersion, patternPath string, src common.FileSource) (dto.Collection, error)

// SchemaApplier aplica um DTO de volta no binario e persiste em mods/.
type SchemaApplier func(version common.GameVersion, entry dto.FileEntry) error

// HasChunkmapSchema informa se o arquivo tem schema chunkmap registrado.
func HasChunkmapSchema(version common.GameVersion, patternPath string) bool {
	_, ok := SchemaRegistry[objectsfile.FileLayoutKey(version, patternPath)]
	return ok
}

// BuildCollectionFromChunks monta a dto.Collection a partir das rows ja
// montadas pelo chamador. Chave = basename sem extensao.
func BuildCollectionFromChunks(version common.GameVersion, patternPath string, rows []dto.TextRow) (dto.Collection, error) {
	key := strings.TrimSuffix(filepath.Base(patternPath), filepath.Ext(patternPath))
	entry := dto.FileEntry{
		Metadata: dto.NewObjectMetadata(
			version,
			filepath.ToSlash(filepath.Dir(patternPath)),
			filepath.Base(patternPath),
			objectsfile.FileLayoutKey(version, patternPath),
		).WithRowCount(len(rows)),
		Rows: rows,
	}
	c := dto.Collection{key: entry}
	if err := hash.ValidateNoCollision(c); err != nil {
		return nil, err
	}
	return c, nil
}

// loadGeneric carrega qualquer arquivo binario via chunkmap e monta o DTO
// com rows nomeadas pelo campo Go (Name.Standard, Help.Simplified, ...).
func loadGeneric[T any](version common.GameVersion, patternPath string) (dto.Collection, error) {
	return loadGenericFrom[T](version, patternPath, common.SourcePreferred)
}

// loadGenericFrom carrega de uma FONTE especifica (SourceData para original).
func loadGenericFrom[T any](version common.GameVersion, patternPath string, src common.FileSource) (dto.Collection, error) {
	if err := ffxencoding.EnsureAllCharsetsLoaded(version); err != nil {
		return nil, fmt.Errorf("%s: charsets: %w", patternPath, err)
	}

	defaultLang := common.DefaultLocalization
	main, err := LoadFileFromSource[T](version, patternPath, src)
	if err != nil {
		return nil, err
	}

	files := map[string]*File[T]{defaultLang: main}
	for _, lang := range common.SupportedLanguageCodes() {
		if lang == defaultLang {
			continue
		}
		langData, lerr := readFileBytesFrom(version, lang, patternPath, src)
		if lerr != nil {
			common.LogVerbose("[chunkmap] %s/%s: %v", patternPath, lang, lerr)
			continue
		}
		langFile, lerr := LoadFileFromBytesForLanguage[T](version, patternPath, lang, langData)
		if lerr != nil {
			common.LogVerbose("[chunkmap] %s/%s: %v", patternPath, lang, lerr)
			continue
		}
		files[lang] = langFile
	}

	rows, err := GenericRows[T](main, files)
	if err != nil {
		return nil, err
	}
	return BuildCollectionFromChunks(version, patternPath, rows)
}

// SchemaMeasurer aplica as rows em memória e devolve o tamanho da string
// table reconstruída no idioma da carga (sem gravar).
type SchemaMeasurer func(version common.GameVersion, patternPath string, rows []dto.TextRow) (int, error)
// measureGeneric aplica as rows em memória e devolve o tamanho da string
// table reconstruída (com dedup) no idioma da carga. Sem gravar nada.
func measureGeneric[T any](version common.GameVersion, patternPath string, rows []dto.TextRow) (int, error) {
	if err := ffxencoding.EnsureAllCharsetsLoaded(version); err != nil {
		return 0, fmt.Errorf("%s: charsets: %w", patternPath, err)
	}
	f, err := LoadFile[T](version, patternPath)
	if err != nil {
		return 0, err
	}
	return f.MeasureStringTable(rows, common.DefaultLocalization)
}
// applyGeneric aplica rows de volta no binario e persiste em mods/.
func applyGeneric[T any](version common.GameVersion, patternPath string, entry dto.FileEntry) error {
	if err := ffxencoding.EnsureAllCharsetsLoaded(version); err != nil {
		return fmt.Errorf("%s: charsets: %w", patternPath, err)
	}

	file, err := LoadFile[T](version, patternPath)
	if err != nil {
		return err
	}
	if err := GenericApply[T](file, entry.Rows); err != nil {
		return err
	}
	return file.SaveToBinary("")
}

// GenericRows monta as rows de export de qualquer File[T]. Cada campo
// textual vira uma row: Index = posicao do chunk, Name = caminho Go do
// campo (ex: "Name.Standard"). Todos os idiomas sao mesclados na row.
func GenericRows[T any](f *File[T], files map[string]*File[T]) ([]dto.TextRow, error) {
	if f == nil {
		return nil, fmt.Errorf("chunkmap: arquivo ausente")
	}

	// Coleta textos por (chunk, field) de todos os idiomas.
	type fieldAddress struct{ chunk, field int }
	texts := make(map[fieldAddress]map[string]string)

	for lang, langFile := range files {
		if langFile == nil {
			langFile = f
		}
		exported, err := langFile.ExportText(lang)
		if err != nil {
			return nil, fmt.Errorf("chunkmap: exportar %s: %w", lang, err)
		}
		for _, chunk := range exported {
			for _, field := range chunk.Fields {
				address := fieldAddress{chunk: chunk.Index, field: field.Index}
				merged := texts[address]
				if merged == nil {
					merged = make(map[string]string)
					texts[address] = merged
				}
				merged[lang] = field.Texts[lang]
			}
		}
	}

	// Monta as rows na ordem dos chunks e campos.
	rows := make([]dto.TextRow, 0, f.Count()*4)
	for chunk := 0; chunk < f.Count(); chunk++ {
		obj := f.Get(chunk)
		names := obj.OrderedFieldNames()
		for fieldIdx, name := range names {
			address := fieldAddress{chunk: chunk, field: fieldIdx}
			t := texts[address]
			if len(t) == 0 {
				continue
			}
			rows = append(rows, dto.TextRow{
				Index: chunk,
				Name:  name,
				Hash:  hash.Texts(t),
				Text:  t,
			})
		}
	}
	return rows, nil
}

// GenericApply aplica as rows de volta nos chunks. Index e Name sao a
// identidade: Index localiza o chunk, Name o campo textual dentro dele.
func GenericApply[T any](f *File[T], rows []dto.TextRow) error {
	if f == nil {
		return fmt.Errorf("chunkmap: arquivo ausente")
	}

	// Agrupa fields por chunk para aplicacao em lote.
	byChunk := make(map[int][]TextField)
	for _, r := range rows {
		if r.Index < 0 || r.Index >= f.Count() {
			return fmt.Errorf("chunkmap: chunk %d fora da faixa (%d chunks)", r.Index, f.Count())
		}
		obj := f.Get(r.Index)
		names := obj.OrderedFieldNames()
		fieldIdx := -1
		for i, n := range names {
			if n == r.Name {
				fieldIdx = i
				break
			}
		}
		if fieldIdx < 0 {
			return fmt.Errorf("chunkmap: campo %q nao existe no chunk %d", r.Name, r.Index)
		}
		byChunk[r.Index] = append(byChunk[r.Index], TextField{
			Index: fieldIdx,
			Name:  r.Name,
			Texts: r.Text,
		})
	}

	for chunkIdx, fields := range byChunk {
		if err := f.Get(chunkIdx).ImportText(fields); err != nil {
			return fmt.Errorf("chunkmap: chunk %d: %w", chunkIdx, err)
		}
	}
	return nil
}
