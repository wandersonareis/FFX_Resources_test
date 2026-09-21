package builders

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/formatters/hash"
)

// Macro name rows: o formato atual (chunks com ChunkIndex + strings
// posicionais) é mantido — o idx é necessário para reconstrução do
// binário na mesma ordem. Cada string vira até duas rows:
// Name="name" e, quando houver simplificado distinto, Name="simplifiedName".
const (
	macroNameField           = "name"
	macroSimplifiedNameField = "simplifiedName"
)

// MacroChunkKey é a chave da Collection para um chunk: basename sem
// extensão na forma chunk_NN.
func MacroChunkKey(chunk int) string {
	return fmt.Sprintf("chunk_%02d", chunk)
}

// parseMacroChunkKey extrai o índice do chunk de chaves "chunk_NN".
func parseMacroChunkKey(key string) (int, bool) {
	rest, ok := strings.CutPrefix(key, "chunk_")
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// BuildMacroDTO monta a Collection pronta (texto + metadata + hash) lendo
// os containers do disco para a versão dada.
func BuildMacroDTO(version common.GameVersion) (dto.Collection, error) {
	containers, err := macrodic.ReadMacroDictionaryContainers(version)
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, fmt.Errorf("no macro dictionary data found")
	}
	return BuildMacroDTOFromContainers(version, containers)
}

// BuildMacroDTOFromContainers monta a Collection a partir de containers já
// carregados (texto raw + metadata + hash xxHash64 por frase/idioma).
func BuildMacroDTOFromContainers(version common.GameVersion, containers map[string]*macrodic.MacroDictionaryBinaryFile) (dto.Collection, error) {
	locs := make([]string, 0, len(containers))
	for loc := range containers {
		locs = append(locs, loc)
	}
	sort.Strings(locs)

	type entryKey struct {
		chunk int
		index int
	}
	names := make(map[entryKey]map[string]string)
	simplified := make(map[entryKey]map[string]string)
	// seen registra toda posição (chunk, index) com segmento não-nulo,
	// mesmo sem texto em nenhum idioma: a posição precisa existir no
	// rebuild (vira entrada zero), como no formato legado.
	seen := make(map[entryKey]bool)
	chunks := make(map[int]bool)

	for _, loc := range locs {
		c := containers[loc]
		if c == nil {
			continue
		}
		for chunkIndex, strings := range c.MapAllStrings() {
			chunks[chunkIndex] = true
			for stringIndex, s := range strings {
				if s == nil {
					continue
				}
				k := entryKey{chunk: chunkIndex, index: stringIndex}
				seen[k] = true
				if text := s.GetRegularString(); text != "" {
					if names[k] == nil {
						names[k] = make(map[string]string)
					}
					names[k][loc] = text
				}
				if s.HasDistinctSimplified() {
					if text := s.GetSimplifiedString(); text != "" {
						if simplified[k] == nil {
							simplified[k] = make(map[string]string)
						}
						simplified[k][loc] = text
					}
				}
			}
		}
	}

	out := make(dto.Collection, len(chunks))
	for chunk := range chunks {
		key := MacroChunkKey(chunk)
		entry := dto.FileEntry{
			Metadata: dto.NewMacroMetadata(version, chunk),
			Rows:     []dto.TextRow{},
		}
		out[key] = entry
	}
	// Emite uma row por posição vista, mesmo sem texto (placeholder
	// posicional com Text vazio, sem hash): o rebuild precisa do índice.
	ordered := make([]entryKey, 0, len(seen))
	for k := range seen {
		ordered = append(ordered, k)
	}
	sort.Slice(ordered, func(a, b int) bool {
		if ordered[a].chunk != ordered[b].chunk {
			return ordered[a].chunk < ordered[b].chunk
		}
		return ordered[a].index < ordered[b].index
	})
	for _, k := range ordered {
		key := MacroChunkKey(k.chunk)
		entry := out[key]
		texts := names[k]
		if texts == nil {
			texts = make(map[string]string)
		}
		entry.Rows = append(entry.Rows, dto.TextRow{
			Index: k.index,
			Name:  macroNameField,
			Hash:  hash.Texts(texts),
			Text:  texts,
		})
		if simp, ok := simplified[k]; ok {
			entry.Rows = append(entry.Rows, dto.TextRow{
				Index: k.index,
				Name:  macroSimplifiedNameField,
				Hash:  hash.Texts(simp),
				Text:  simp,
			})
		}
		out[key] = entry
	}
	for key, entry := range out {
		// Ordena por (index, name) para saída determinística; a reconstrução
		// usa o Index posicional, então a ordem entre name/simplifiedName
		// no arquivo não importa.
		sort.Slice(entry.Rows, func(a, b int) bool {
			if entry.Rows[a].Index != entry.Rows[b].Index {
				return entry.Rows[a].Index < entry.Rows[b].Index
			}
			return entry.Rows[a].Name < entry.Rows[b].Name
		})
		entry.Metadata = entry.Metadata.WithRowCount(len(entry.Rows))
		out[key] = entry
	}
	// TODO: deletar quando colisão xxHash64 for considerada segura —
	// guarda temporária de desencargo: reprova DTO com mesmo hash para textos diferentes.
	if err := hash.ValidateNoCollision(out); err != nil {
		return nil, err
	}
	return out, nil
}

// RebuildMacroContainers aplica o DTO de volta em containers binários
// (parse DTO → binário, sem tocar em disco). A ordem posicional vem do
// Index; gaps viram entradas zero, como no rebuild legado.
func RebuildMacroContainers(version common.GameVersion, c dto.Collection) (map[string]*macrodic.MacroDictionaryBinaryFile, error) {
	type entryKey struct {
		chunk int
		index int
	}
	names := make(map[entryKey]map[string]string)
	simplified := make(map[entryKey]map[string]string)
	for _, key := range c.SortedKeys() {
		entry := c[key]
		chunk := -1
		if n, ok := dto.ChunkIndexFromID(entry.Metadata.ID); ok {
			chunk = n
		} else if n, ok := parseMacroChunkKey(key); ok {
			chunk = n
		} else if n, ok := dto.ChunkIndexFromID(key); ok {
			chunk = n
		} else {
			return nil, fmt.Errorf("macro entry %q without chunk index", key)
		}
		dto.SortRows(entry.Rows)
		for _, row := range entry.Rows {
			if row.Index < 0 {
				continue
			}
			k := entryKey{chunk: chunk, index: row.Index}
			switch row.Name {
			case "", macroNameField:
				// Registra o índice mesmo sem texto (placeholder posicional):
				// o rebuild cria a entrada zero correspondente.
				if names[k] == nil {
					names[k] = make(map[string]string)
				}
				for lang, text := range row.Text {
					if text == "" {
						continue
					}
					if names[k] == nil {
						names[k] = make(map[string]string)
					}
					names[k][lang] = text
				}
			case macroSimplifiedNameField:
				for lang, text := range row.Text {
					if text == "" {
						continue
					}
					if simplified[k] == nil {
						simplified[k] = make(map[string]string)
					}
					simplified[k][lang] = text
				}
			default:
				common.LogVerbose("unknown macro row name %q (chunk %d index %d), skipping", row.Name, chunk, row.Index)
			}
		}
	}
	byChunk := make(map[int][]macrodic.MacroStringJsonImport)
	for k, texts := range names {
		byChunk[k.chunk] = append(byChunk[k.chunk], macrodic.MacroStringJsonImport{
			Index:          k.index,
			Name:           texts,
			SimplifiedName: simplified[k],
		})
	}
	// Strings só-simplified (sem name) também precisam existir para o rebuild
	// posicional: viram entrada com Name vazio (fallback do rebuild aponta
	// o simplificado para os bytes do nome vazio, como no legado).
	for k, texts := range simplified {
		if _, ok := names[k]; !ok {
			byChunk[k.chunk] = append(byChunk[k.chunk], macrodic.MacroStringJsonImport{
				Index:          k.index,
				SimplifiedName: texts,
			})
		}
	}
	chunks := make([]macrodic.MacroChunkJsonImport, 0, len(byChunk))
	for chunk, strs := range byChunk {
		sort.Slice(strs, func(a, b int) bool { return strs[a].Index < strs[b].Index })
		chunks = append(chunks, macrodic.MacroChunkJsonImport{ChunkIndex: chunk, Strings: strs})
	}
	sort.Slice(chunks, func(a, b int) bool { return chunks[a].ChunkIndex < chunks[b].ChunkIndex })
	return macrodic.ImportFromJson(&macrodic.MacroDictionaryJsonImport{Chunks: chunks}, version)
}

// ApplyMacroDTO reconstrói os containers a partir do DTO e salva cada
// binário de volta no seu game file. Outro responsável (formatter) fez
// a leitura do arquivo; este applier só parseia DTO → binário.
func ApplyMacroDTO(version common.GameVersion, c dto.Collection) error {
	if len(c) == 0 {
		return fmt.Errorf("no macro data in DTO to apply")
	}
	containers, err := RebuildMacroContainers(version, c)
	if err != nil {
		return err
	}
	if err := macrodic.SaveMacroDictionaryBinaries(containers); err != nil {
		return err
	}
	common.LogVerbose("Macro dictionary processed successfully (%d container(s))", len(containers))
	return nil
}
