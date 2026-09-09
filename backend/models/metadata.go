package models

import (
	"bytes"
	"encoding/json"
	"ffxresources/backend/common"
	"os"
	"path/filepath"
)

// FileMetadata wraps the SpiraFileInfo of an ORIGINAL game binary so an exported
// artifact can be mapped back to the file it came from (for binary reconstruction).
type FileMetadata struct {
	FileInfo SpiraFileInfo `json:"file_info"`
}

// NewFileMetadata creates a FileMetadata from a SpiraFileInfo (nil-safe).
func NewFileMetadata(info *SpiraFileInfo) *FileMetadata {
	if info == nil {
		return nil
	}
	cp := *info
	return &FileMetadata{FileInfo: cp}
}

// ---- source binary path helpers -------------------------------------------------

// ObjectFileBinaryPath returns the absolute path of an objectsfile-type binary
// (e.g. "battle/kernel/important.bin") for the default localization.
func ObjectFileBinaryPath(patternPath string) string {
	return filepath.Join(common.GameFilesRoot, common.ModsFolder, common.GetLocalizationRoot(common.DefaultLocalization), patternPath)
}

// EventBinaryPath returns the absolute path of an event binary given its ID
// (e.g. "ev001" -> .../event/obj_ps3/ev/ev001/ev001.bin).
func EventBinaryPath(id string) string {
	if len(id) < 2 {
		return ""
	}
	rel := filepath.Join("event/obj_ps3", id[:2], id, id+".bin")
	return filepath.Join(common.GameFilesRoot, common.ModsFolder, common.GetLocalizationRoot(common.DefaultLocalization), rel)
}

// MacroBinaryPath returns the absolute path of a macro dictionary binary for a
// given localization (e.g. "us" -> .../new_uspc/menu/macrodic.dcp).
func MacroBinaryPath(localization string) string {
	rel := filepath.Join("menu", "macrodic.dcp")
	return filepath.Join(common.GameFilesRoot, common.ModsFolder, common.GetLocalizationRoot(localization), rel)
}

// ---- export file wrapper -------------------------------------------------------

// DataWrapper wraps any payload under a "data" key as the exported file format.
type DataWrapper[T any] struct {
	Data T `json:"data"`
}

// SaveDataFile writes payload wrapped as {"data": payload} WITHOUT HTML escaping,
// so paths containing &, <, > stay human-readable.
func SaveDataFile[T any](payload T, filePath string) error {
	return saveNoEscape(DataWrapper[T]{Data: payload}, filePath)
}

// LoadDataFile reads a {"data": payload} file, falling back to a bare payload when
// the file is in the legacy (pre-metadata) format.
func LoadDataFile[T any](filePath string) (T, error) {
	var zero T
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return zero, err
	}

	var probe map[string]json.RawMessage
	if json.Unmarshal(raw, &probe) == nil {
		if _, ok := probe["data"]; ok {
			var wrapped DataWrapper[T]
			if err := json.Unmarshal(raw, &wrapped); err == nil {
				return wrapped.Data, nil
			}
		}
	}

	if err := json.Unmarshal(raw, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}

func saveNoEscape(v any, filePath string) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

// ---- export payload types ------------------------------------------------------

// EventStringDataExport is the JSON form of a single event string.
type EventStringDataExport struct {
	Index int               `json:"index"`
	Text  map[string]string `json:"text"`
}

// EventFileExport is one event in the events export, carrying its binary metadata.
type EventFileExport struct {
	Metadata *FileMetadata           `json:"metadata"`
	ID       string                  `json:"id"`
	Strings  []EventStringDataExport `json:"strings"`
}

// MacroStringExport is the JSON form of a single macro string.
type MacroStringExport struct {
	Index          int    `json:"index"`
	RegularText    string `json:"regular_text"`
	SimplifiedText string `json:"simplified_text"`
	HasDistinct    bool   `json:"has_distinct_simplified"`
}

// MacroChunkExport is the JSON form of a macro chunk.
type MacroChunkExport struct {
	ChunkIndex int                  `json:"chunk_index"`
	Strings    []MacroStringExport `json:"strings"`
}

// MacroLocalizationExport is one localization in the macro dictionary export.
type MacroLocalizationExport struct {
	Metadata    *FileMetadata       `json:"metadata"`
	Localization string             `json:"localization"`
	Chunks       []MacroChunkExport `json:"chunks"`
}

// ObjectsFileExport is an objectsfile export: binary metadata + the text strings.
// Strings is kept as raw JSON so the objectsfile package can (un)marshal its own types.
type ObjectsFileExport struct {
	Metadata *FileMetadata   `json:"metadata"`
	Strings  json.RawMessage `json:"strings"`
}

// NewFileInfoFromPath builds a SpiraFileInfo for an arbitrary file (such as a game binary)
// without requiring an FFX/FFX-2 prefix in the path. The version resolves to Unknown (0)
// when the prefix is absent. It never returns an error.
func NewFileInfoFromPath(path string) *SpiraFileInfo {
	aPath, err := filepath.Abs(path)
	if err != nil {
		aPath = path
	}

	info, statErr := os.Stat(aPath)
	if info == nil || statErr != nil {
		return &SpiraFileInfo{
			Name:       common.RecursiveRemoveFileExtension(filepath.Base(path)),
			NamePrefix: common.RemoveOneFileExtension(filepath.Base(path)),
			Extension:  filepath.Ext(path),
			Path:       path,
			Parent:     filepath.Dir(path),
			Type:       guessFileType(path),
			Version:    GameVersion(getVersionFromPrefix(aPath)),
		}
	}

	fileInfo := &SpiraFileInfo{
		Name:       common.RecursiveRemoveFileExtension(info.Name()),
		NamePrefix: common.RemoveOneFileExtension(info.Name()),
		Extension:  filepath.Ext(path),
		IsDir:      info.IsDir(),
		Path:       path,
		Parent:     filepath.Dir(path),
		Type:       guessFileType(path),
		Version:    GameVersion(getVersionFromPrefix(aPath)),
	}

	if !info.IsDir() && fileInfo.Type != DcpParts {
		if v, verr := common.CheckFFXPath(aPath); verr == nil {
			fileInfo.Version = GameVersion(v)
			if rp, rerr := common.RelativePathFromMatch(aPath); rerr == nil {
				fileInfo.RelativePath = rp
			}
		}
	}

	if !info.IsDir() {
		fileInfo.Size = info.Size()
	}

	return fileInfo
}
