package dto

import (
	"ffxresources/backend/common"
	"ffxresources/backend/models"
)

// Metadata unifica dados e geração no mesmo lugar: os mesmos campos
// produzidos pelos geradores legados (models.NewEventFileInfo,
// models.NewObjectFileMetadata/Keyed), copiados campo a campo para
// compatibilidade. O source antigo em models é mantido e continua
// sendo a origem dos valores.
//
// Todos os campos usam omitempty: formatos que não fornecem um campo
// não o emitem (sem nulos nem vazios no JSON). Version é ponteiro
// justamente para omitempty funcionar (struct não é omitida).
type Metadata struct {
	EventID             string                `json:"event_id,omitempty"`
	Shortened           string                `json:"shortened,omitempty"`
	MidPath             string                `json:"mid_path,omitempty"`
	EventFilePath       string                `json:"event_file_path,omitempty"`
	LocalizationPattern string                `json:"localization_pattern,omitempty"`
	Version             *common.GameVersion   `json:"version,omitempty"`
	DirPattern          string                `json:"dir_pattern,omitempty"`
	FileName            string                `json:"file_name,omitempty"`
	Key                 string                `json:"key,omitempty"`
	FileInfo            *models.SpiraFileInfo `json:"file_info,omitempty"`
	ChunkIndex          *int                  `json:"chunk_index,omitempty"`
}

// NewEventMetadata gera a metadata de um evento chamando o gerador
// legado models.NewEventFileInfo e copiando os mesmos campos.
func NewEventMetadata(eventID string, version common.GameVersion) Metadata {
	info := models.NewEventFileInfo(eventID, version)
	if info == nil {
		return Metadata{EventID: eventID, Version: &version}
	}
	v := info.Version
	return Metadata{
		EventID:             info.EventID,
		Shortened:           info.Shortened,
		MidPath:             info.MidPath,
		EventFilePath:       info.EventFilePath,
		LocalizationPattern: info.LocalizationPattern,
		Version:             &v,
		DirPattern:          info.DirPattern,
		FileName:            info.FileName,
		Key:                 info.Key,
	}
}

// NewObjectMetadata gera a metadata de um objectsfile chamando o gerador
// legado models.NewObjectFileMetadataKeyed e copiando os mesmos campos.
func NewObjectMetadata(version common.GameVersion, dirPattern, fileName, key string) Metadata {
	meta := models.NewObjectFileMetadataKeyed(version, dirPattern, fileName, key)
	if meta == nil {
		return Metadata{Version: &version, DirPattern: dirPattern, FileName: fileName, Key: key}
	}
	out := Metadata{
		DirPattern: dirPattern,
		FileName:   fileName,
		Key:        key,
	}
	fi := meta.FileInfo
	out.FileInfo = &fi
	if meta.Version != nil {
		v := *meta.Version
		out.Version = &v
	} else {
		v := version
		out.Version = &v
	}
	return out
}

// NewMacroMetadata gera a metadata de um chunk de macrodic. O formato atual
// (chunks com ChunkIndex + strings posicionais) é mantido: o idx é
// necessário para reconstrução do binário na mesma ordem.
func NewMacroMetadata(version common.GameVersion, chunkIndex int) Metadata {
	idx := chunkIndex
	return Metadata{
		Version:    &version,
		DirPattern: "menu",
		FileName:   "macrodic.dcp",
		Key:        common.VersionPathName(version) + "/menu/macrodic.dcp",
		ChunkIndex: &idx,
	}
}
