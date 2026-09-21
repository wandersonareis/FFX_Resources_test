package dto

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"ffxresources/backend/common"
)

// Metadata é o contrato mínimo de localização de um texto exportado.
//
//	key      - chave canônica do store (version/pattern, com `/`): ex
//	           ffx/event/obj_ps3/az/azit0000/azit0000.bin
//	row_count - número de rows exportadas (conferência; importer ignora)
//	id       - identificador universal = chave da Collection:
//	           events = eventID, objects = basename sem extensão,
//	           macro = chunk_NN
//	is_dir   - omitido quando arquivo; true quando diretório
//
// Todos os demais campos do formato antigo (version, dir_pattern,
// file_name, shortened, mid_path, event_file_path, localization_pattern,
// file_info, chunk_index) são derivados de key (+id) via ParseKey/
// ExpandLegacy e nunca serializados (quebra total do formato antigo).
type Metadata struct {
	Key      string `json:"key"`
	RowCount int    `json:"row_count,omitempty"`
	ID       string `json:"id,omitempty"`
	IsDir    bool   `json:"is_dir,omitempty"`
}

// WithRowCount preenche a contagem de rows (len(entry.Rows) no export).
func (m Metadata) WithRowCount(n int) Metadata {
	m.RowCount = n
	return m
}

// NewEventMetadata gera a metadata de um evento a partir do id.
// Chave canônica: <version>/event/obj_ps3/<short>/<id>/<id>.bin,
// onde short = id[:2]. Sem dependência dos geradores legados.
func NewEventMetadata(eventID string, version common.GameVersion) Metadata {
	id := strings.TrimSpace(eventID)
	if len(id) < 2 {
		return Metadata{ID: id}
	}
	short := id[:2]
	key := common.VersionPathName(version) + "/event/obj_ps3/" + short + "/" + id + "/" + id + ".bin"
	return Metadata{Key: key, ID: id}
}

// NewObjectMetadata gera a metadata de um objectsfile.
// id = basename sem extensão (important, command, btl_txt, ...).
func NewObjectMetadata(version common.GameVersion, dirPattern, fileName, key string) Metadata {
	_ = version
	_ = dirPattern
	id := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	if key == "" {
		key = strings.Trim(strings.TrimSpace(dirPattern)+"/"+strings.TrimSpace(fileName), "/")
	}
	return Metadata{Key: filepath.ToSlash(key), ID: id}
}

// MacroChunkID é o id universal de um chunk: chunk_NN.
func MacroChunkID(chunk int) string {
	return fmt.Sprintf("chunk_%02d", chunk)
}

// ChunkIndexFromID extrai o índice de ids chunk_NN.
func ChunkIndexFromID(id string) (int, bool) {
	rest, ok := strings.CutPrefix(strings.TrimSpace(id), "chunk_")
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// NewMacroMetadata gera a metadata de um chunk de macrodic.
// A key é compartilhada por todos os chunks; a identidade é o id (chunk_NN).
func NewMacroMetadata(version common.GameVersion, chunkIndex int) Metadata {
	return Metadata{
		Key: common.VersionPathName(version) + "/menu/macrodic.dcp",
		ID:  MacroChunkID(chunkIndex),
	}
}

// ParsedKey é a decomposição de uma key canônica.
type ParsedKey struct {
	Version             string // primeiro segmento (ffx, ffx2, lastmiss)
	DirPattern          string // segmentos intermediários (event/obj_ps3, battle/kernel, menu)
	FileName            string // basename (azit0000.bin)
	Stem                string // basename sem extensão (azit0000)
	LocalizationPattern string // key sem o primeiro segmento
}

// ParseKey decompõe version/pattern em partes. Exige ao menos 2 segmentos.
func ParseKey(key string) (ParsedKey, bool) {
	clean := strings.Trim(filepath.ToSlash(strings.TrimSpace(key)), "/")
	if clean == "" {
		return ParsedKey{}, false
	}
	segs := strings.Split(clean, "/")
	if len(segs) < 2 {
		return ParsedKey{}, false
	}
	fileName := segs[len(segs)-1]
	stem := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return ParsedKey{
		Version:             segs[0],
		DirPattern:          strings.Join(segs[1:len(segs)-1], "/"),
		FileName:            fileName,
		Stem:                stem,
		LocalizationPattern: strings.Join(segs[1:], "/"),
	}, true
}

// LegacyView deriva os campos do formato antigo a partir da key,
// para prova de equivalência e reconstrução no importer.
// O prefixo de build (pack root) vem de common.PackRootForVersion.
type LegacyView struct {
	Version             string
	DirPattern          string
	FileName            string
	Shortened           string
	EventID             string
	LocalizationPattern string
	MidPath             string
	EventFilePath       string
}

// ExpandLegacy deriva a visão legada. Retorna ok=false se a key for inválida.
// Shortened/MidPath/EventID só fazem sentido para events; fora disso,
// Shortened/MidPath ficam vazios e EventID = stem.
func (m Metadata) ExpandLegacy(version common.GameVersion, localization string) (LegacyView, bool) {
	p, ok := ParseKey(m.Key)
	if !ok {
		return LegacyView{}, false
	}
	if localization == "" {
		localization = common.DefaultLocalization
	}
	out := LegacyView{
		Version:             p.Version,
		DirPattern:          p.DirPattern,
		FileName:            p.FileName,
		LocalizationPattern: p.LocalizationPattern,
		EventID:             p.Stem,
	}
	// Forma event: event/obj_ps3/<short>/<id>/<file> → dir_pattern = key[1:3],
	// shortened = key[3], event_id = key[4], mid_path = key[3:] sem extensão.
	if segs := strings.Split(p.LocalizationPattern, "/"); len(segs) == 5 &&
		segs[0] == "event" && segs[1] == "obj_ps3" {
		out.DirPattern = "event/obj_ps3"
		out.Shortened = segs[2]
		out.EventID = segs[3]
		out.MidPath = strings.TrimSuffix(strings.Join(segs[2:], "/"), filepath.Ext(segs[4]))
	} else if m.ID != "" {
		out.EventID = m.ID
	}
	out.EventFilePath = filepath.ToSlash(filepath.Join(
		common.PackRootForVersion(version, localization),
		p.LocalizationPattern,
	))
	return out, true
}
