package strings

import (
	"encoding/hex"
	"fmt"
	"strconv"
	stdstrings "strings"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

// parsedLine é uma linha de conteúdo decomposta.
type parsedLine struct {
	id      string
	name    string
	index   int
	lang    string
	hash    string // hex bare
	ref     bool
	value   string // literal (com escapes desfeitos) ou "" quando ref
}

// Unmarshal parseia o formato Strings de volta para a Collection (DTO),
// com rows ordenados por Index para reconstrução posicional do binário.
//
// O header (key/row_count) é informativo: a key ancora a metadata da
// entrada e row_count é conferido com LogVerbose (o import reconta).
// Como no JSON, o hash é ponteiro opaco no import: refs "$hash" são
// resolvidas pela tabela hash→texto do próprio arquivo, e defs com
// xxh64(literal) != hash da chave geram só LogVerbose (texto vence).
func Unmarshal(data []byte) (dto.Collection, error) {
	entries := make(map[string]*dto.FileEntry)
	rows := make(map[string]map[string]*dto.TextRow) // entryID -> "index\x00name" -> row
	type pendingRef struct {
		entryID string
		rkey    string
		lang    string
		hash    string
	}
	var refs []pendingRef
	table := make(map[string]string) // bare hash -> texto (defs do arquivo)
	blockKey := ""
	expectedRows := make(map[string]int) // metadata.key -> row_count do header
	lineno := 0

	for _, raw := range stdstrings.Split(string(data), "\n") {
		lineno++
		line := stdstrings.TrimRight(raw, "\r")
		if stdstrings.TrimSpace(line) == "" {
			continue
		}
		if key, n, ok := parseHeader(line); ok {
			blockKey = key
			expectedRows[key] = n
			continue
		}
		pl, err := parseLine(line, lineno)
		if err != nil {
			return nil, err
		}
		entry := entries[pl.id]
		if entry == nil {
			entry = &dto.FileEntry{Metadata: dto.Metadata{Key: blockKey, ID: pl.id}}
			entries[pl.id] = entry
			rows[pl.id] = make(map[string]*dto.TextRow)
		} else if entry.Metadata.Key == "" && blockKey != "" {
			entry.Metadata.Key = blockKey
		}
		rkey := strconv.Itoa(pl.index) + "\x00" + pl.name
		row := rows[pl.id][rkey]
		if row == nil {
			entry.Rows = append(entry.Rows, dto.TextRow{Index: pl.index, Name: pl.name})
			row = &entry.Rows[len(entry.Rows)-1]
			rows[pl.id][rkey] = row
		}
		if row.Hash == nil {
			row.Hash = make(map[string]string)
		}
		if row.Text == nil {
			row.Text = make(map[string]string)
		}
		row.Hash[pl.lang] = pl.hash
		if pl.ref {
			refs = append(refs, pendingRef{entryID: pl.id, rkey: rkey, lang: pl.lang, hash: pl.hash})
			continue
		}
		if pl.hash != "" {
			if got := hash.Sum64Hex(pl.value); got != pl.hash {
				common.LogVerbose("strings hash mismatch for %s[%d] lang %s: file %s vs text %s",
					pl.id, pl.index, pl.lang, pl.hash, got)
			}
			if _, dup := table[pl.hash]; !dup && pl.value != "" {
				table[pl.hash] = pl.value
			}
		}
		row.Text[pl.lang] = pl.value
	}
	// Segunda passada: resolve refs pela tabela do próprio arquivo.
	for _, r := range refs {
		row := rows[r.entryID][r.rkey]
		if row == nil {
			continue
		}
		if v, ok := table[r.hash]; ok {
			row.Text[r.lang] = v
		} else {
			common.LogVerbose("strings orphan dedup reference $%s kept as-is", r.hash)
			row.Text[r.lang] = hash.Prefix(r.hash)
		}
	}
	out := make(dto.Collection, len(entries))
	for id, entry := range entries {
		dto.SortRows(entry.Rows)
		entry.Metadata = entry.Metadata.WithRowCount(len(entry.Rows))
		if want, ok := expectedRows[entry.Metadata.Key]; ok && len(entry.Rows) != want {
			common.LogVerbose("strings row_count mismatch for %s: header %d vs %d rows",
				entry.Metadata.Key, want, len(entry.Rows))
		}
		out[id] = *entry
	}
	return out, nil
}

// parseHeader lê /*key=<key> row_count=<n>*/.
func parseHeader(line string) (key string, rowCount int, ok bool) {
	t := stdstrings.TrimSpace(line)
	if !stdstrings.HasPrefix(t, "/*key=") || !stdstrings.HasSuffix(t, "*/") {
		return "", 0, false
	}
	inner := stdstrings.TrimSuffix(stdstrings.TrimPrefix(t, "/*key="), "*/")
	idx := stdstrings.LastIndex(inner, " row_count=")
	if idx < 0 {
		return "", 0, false
	}
	key = stdstrings.TrimSpace(inner[:idx])
	n, err := strconv.Atoi(stdstrings.TrimSpace(inner[idx+len(" row_count="):]))
	if err != nil {
		return "", 0, false
	}
	if key == "" {
		return "", 0, false
	}
	return key, n, true
}

// parseLine decompõe uma linha de conteúdo com âncora à direita:
// último campo = lang, penúltimo = index, resto = prefixo opaco
// ([versão:]... id [:name]).
func parseLine(line string, lineno int) (parsedLine, error) {
	fail := func(format string, args ...any) (parsedLine, error) {
		return parsedLine{}, fmt.Errorf("strings line %d: %s", lineno, fmt.Sprintf(format, args...))
	}
	ki := stdstrings.Index(line, keyValueSep)
	if ki < 0 {
		return fail("missing %q separator", keyValueSep)
	}
	keyPart, rest := line[:ki], line[ki+len(keyValueSep):]
	eq := stdstrings.Index(rest, "=")
	if eq < 0 {
		return fail("missing = separator")
	}
	rawHash := stdstrings.TrimSpace(rest[:eq])
	bare, ok := hash.Strip(rawHash)
	if !ok || bare == "" {
		return fail("missing $hash before =")
	}
	if len(bare) != 16 {
		return fail("hash must be 16 hex chars, got %q", bare)
	}
	if _, err := hex.DecodeString(bare); err != nil {
		return fail("hash is not hex: %q", bare)
	}
	value := stdstrings.TrimPrefix(rest[eq+1:], " ")
	fields := stdstrings.Split(keyPart, ":")
	if len(fields) < 3 {
		return fail("key needs at least id:index:lang, got %q", keyPart)
	}
	lang := fields[len(fields)-1]
	if stdstrings.TrimSpace(lang) == "" {
		return fail("empty lang in %q", keyPart)
	}
	index, err := strconv.Atoi(fields[len(fields)-2])
	if err != nil || index < 0 {
		return fail("invalid index in %q", keyPart)
	}
	left := fields[:len(fields)-2]
	if len(left) > 0 {
		if _, verr := common.ParseGameVersionStrict(left[0]); verr == nil {
			left = left[1:]
		}
	}
	var id, name string
	switch len(left) {
	case 0:
		return fail("missing id in %q", keyPart)
	case 1:
		id = left[0]
	default:
		// Crescimento futuro entra pela esquerda: id = penúltimo,
		// name = último do prefixo.
		id = left[len(left)-2]
		name = left[len(left)-1]
	}
	if stdstrings.TrimSpace(id) == "" {
		return fail("empty id in %q", keyPart)
	}
	pl := parsedLine{id: id, name: name, index: index, lang: lang, hash: bare}
	if rv, isRef := hash.Strip(value); isRef && rv == bare {
		pl.ref = true
		return pl, nil
	}
	pl.value = unescapeValue(value)
	return pl, nil
}
