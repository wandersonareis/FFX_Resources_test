package macrodic

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"fmt"
	"path/filepath"
	"sort"
)

// MacroDictionaryChunkCount is the number of 4-byte chunk offsets in the
// macro dictionary header. Each offset is a uint32 pointing to the start
// of a chunk within the container Bytes.
const MacroDictionaryChunkCount = 16

// macroDictionaryHeaderLength is the byte length of the chunk offset table.
const macroDictionaryHeaderLength = MacroDictionaryChunkCount * 4

// macroDictionaryTerminator marks the end of the chunk offset table.
const macroDictionaryTerminator = 0xFFFFFFFF

// MacroDictionaryPatternPath is the localization-relative path of a macro
// dictionary binary, mirroring ObjectBinaryFile.patternPath.
const MacroDictionaryPatternPath = "menu/macrodic.dcp"

// MacroDictionaryBinaryFile holds the raw bytes of one localization of a
// macro dictionary file along with chunk offsets computed from those bytes.
// The container is the single source of truth for binary reconstruction.
//
// It implements IMacroDictionaryFile: LoadFromBinary carrega do disco tudo o
// necessário para PublishStrings no datastore; ExportToJson, ImportFromJson e
// SaveToBinary operam em arquivos, como no ObjectBinaryFile.
type MacroDictionaryBinaryFile struct {
	Bytes        []byte
	ChunkOffsets []uint32
	Localization string
	Version      common.GameVersion
	patternPath  string
}

// NewMacroDictionaryBinaryFile creates an empty file handle for the given
// localization; call LoadFromBinary to populate it from disk.
func NewMacroDictionaryBinaryFile(localization string, version common.GameVersion) *MacroDictionaryBinaryFile {
	return &MacroDictionaryBinaryFile{
		Localization: localization,
		Version:      version,
		patternPath:  MacroDictionaryPatternPath,
	}
}

// NewMacroDictionaryBinaryFileFromBytes creates a container from raw bytes,
// parsing the chunk offset table via mapBytes.
func NewMacroDictionaryBinaryFileFromBytes(data []byte, localization string, version common.GameVersion) (*MacroDictionaryBinaryFile, error) {
	if len(data) < macroDictionaryHeaderLength {
		return nil, fmt.Errorf("insufficient data to create MacroDictionaryBinaryFile: have %d bytes, need at least %d", len(data), macroDictionaryHeaderLength)
	}

	c := &MacroDictionaryBinaryFile{
		Bytes:        data,
		Localization: localization,
		Version:      version,
		patternPath:  MacroDictionaryPatternPath,
	}
	if err := c.mapBytes(); err != nil {
		return nil, err
	}
	return c, nil
}

// LoadFromBinary reads this localization binary from disk and parses the
// chunk offset table, loading everything PublishStrings needs.
func (c *MacroDictionaryBinaryFile) LoadFromBinary() error {
	path := filepath.Join(common.GetLocalizationRoot(c.Localization), c.patternPath)
	accessor, err := common.NewFileAccessor(path)
	if err != nil {
		return err
	}
	data := accessor.ReadBytes()
	if len(data) == 0 {
		return fmt.Errorf("missing macro dictionary file: %s", path)
	}
	parsed, err := NewMacroDictionaryBinaryFileFromBytes(data, c.Localization, c.Version)
	if err != nil {
		return err
	}
	c.Bytes = parsed.Bytes
	c.ChunkOffsets = parsed.ChunkOffsets
	return nil
}

// GetLocalization returns this container localization.
func (c *MacroDictionaryBinaryFile) GetLocalization() string {
	return c.Localization
}

// GetVersion returns this container game version.
func (c *MacroDictionaryBinaryFile) GetVersion() common.GameVersion {
	return c.Version
}

// mapBytes parses the chunk offset table at the start of the container Bytes.
// Each entry is read as uint32. Zero offsets are kept as empty-chunk placeholders.
//
// The parsing mirrors components.BytesToChunks semantics so the container holds
// exactly the offsets the legacy NewMacroDictionaryFile path would use: when the
// 0xFFFFFFFF terminator is found at slot i, only the first i-1 offsets are kept.
func (c *MacroDictionaryBinaryFile) mapBytes() error {
	r := bytes.NewReader(c.Bytes[:macroDictionaryHeaderLength:macroDictionaryHeaderLength])

	c.ChunkOffsets = make([]uint32, 0, MacroDictionaryChunkCount)
	for range MacroDictionaryChunkCount {
		var off uint32
		if err := binary.Read(r, binary.LittleEndian, &off); err != nil {
			return fmt.Errorf("reading chunk offset: %w", err)
		}
		if off == macroDictionaryTerminator {
			if len(c.ChunkOffsets) > 0 {
				c.ChunkOffsets = c.ChunkOffsets[:len(c.ChunkOffsets)-1]
			}
			break
		}
		c.ChunkOffsets = append(c.ChunkOffsets, off)
	}
	return nil
}

// chunkBounds is a temporary from-to byte range resolving where one chunk
// file lives inside the container Bytes.
type chunkBounds struct {
	from uint32
	to   uint32
}

// boundsOf resolves the [from, to) range of the chunk at the given index:
// from is the first valid (non-zero) offset of the chunk itself, to is the
// next valid offset (non-zero and greater than or equal to from) or the end
// of the container Bytes when there is no next valid offset. Zero offsets
// hold no file data, so they resolve as invalid. The same rule applies to v1
// and v2 containers regardless of how many files each one holds.
func (c *MacroDictionaryBinaryFile) boundsOf(chunkIndex int) (chunkBounds, bool) {
	if chunkIndex < 0 || chunkIndex >= len(c.ChunkOffsets) {
		return chunkBounds{}, false
	}
	from := c.ChunkOffsets[chunkIndex]
	if from == 0 || uint64(from) > uint64(len(c.Bytes)) {
		return chunkBounds{}, false
	}
	to := uint32(len(c.Bytes))
	for _, next := range c.ChunkOffsets[chunkIndex+1:] {
		if next != 0 && next >= from {
			to = next
			break
		}
	}
	if to > uint32(len(c.Bytes)) {
		to = uint32(len(c.Bytes))
	}
	if to < from {
		return chunkBounds{}, false
	}
	return chunkBounds{from: from, to: to}, true
}

// Files returns one text file object per valid chunk in the container.
// Only offsets present in the header become files — zero slots hold no file
// data and are skipped. Each file carries its container chunk Index.
func (c *MacroDictionaryBinaryFile) Files() ([]*MacroDictionaryTextFile, error) {
	charset := ffxencoding.GetCharsetForLanguage(c.Localization)
	files := make([]*MacroDictionaryTextFile, 0, len(c.ChunkOffsets))
	for i := range c.ChunkOffsets {
		bounds, ok := c.boundsOf(i)
		if !ok {
			continue
		}
		f, err := NewMacroDictionaryTextFile(c.Bytes[bounds.from:bounds.to], charset, c.Version)
		if err != nil {
			return nil, fmt.Errorf("mapping chunk %d: %w", i, err)
		}
		f.Index = i
		files = append(files, f)
	}
	return files, nil
}

// FileAt returns the text file for the chunk at the given index.
//
// The file Bytes span the temporarily resolved from-to range (see boundsOf).
// Zero offsets yield an empty file. The returned slice references the
// container Bytes (no copy).
func (c *MacroDictionaryBinaryFile) FileAt(chunkIndex int) (*MacroDictionaryTextFile, error) {
	if chunkIndex < 0 || chunkIndex >= len(c.ChunkOffsets) {
		return nil, fmt.Errorf("chunk index out of range: %d", chunkIndex)
	}
	charset := ffxencoding.GetCharsetForLanguage(c.Localization)
	if c.ChunkOffsets[chunkIndex] == 0 {
		return &MacroDictionaryTextFile{Bytes: []byte{}, Charset: charset, Version: c.Version, Index: chunkIndex}, nil
	}
	bounds, ok := c.boundsOf(chunkIndex)
	if !ok {
		return nil, fmt.Errorf("chunk offset out of range: %d", c.ChunkOffsets[chunkIndex])
	}
	f, err := NewMacroDictionaryTextFile(c.Bytes[bounds.from:bounds.to], charset, c.Version)
	if err != nil {
		return nil, err
	}
	f.Index = chunkIndex
	return f, nil
}

// MaxIndex returns how many 4-byte string entries a macro chunk holds.
// Each entry is a uint16 pair (regular offset + simplified offset), so the
// first offset — the byte length of the entry table — divided by 4 yields
// the number of unique texts in the chunk.
func MaxIndex(chunkData []byte) int {
	if len(chunkData) < 2 {
		return 0
	}
	return int(binary.LittleEndian.Uint16(chunkData)) / 4
}

// GetMapObjects builds the macro map of this localization from the already
// parsed chunks, without touching the datastore: key chunk*0x100+index.
func (c *MacroDictionaryBinaryFile) GetMapObjects() datastore.MacroObjects {
	result := components.NewEmptyMap[int, datastore.IGlobalLocalizedMacroStringObject]()
	for i := range c.ChunkOffsets {
		f, err := c.FileAt(i)
		if err != nil || len(f.Segments) == 0 {
			continue
		}
		for j, macroStr := range f.ToMacroStrings() {
			if macroStr == nil {
				continue
			}
			key := i*0x100 + j
			macroO := NewLocalizedMacroStringObject()
			result.Add(key, macroO)
			macroO.SetLocalizedContent(c.Localization, macroStr)
		}
	}
	return result
}

// GetObjects adapts the macro map to an IList, ordered by macro ID, so this
// file satisfies interfaces.IBinaryFile.
func (c *MacroDictionaryBinaryFile) GetObjects() components.IList[datastore.IGlobalLocalizedMacroStringObject] {
	macros := c.GetMapObjects()
	keys := macros.Keys()
	sort.Ints(keys)
	objects := components.NewList[datastore.IGlobalLocalizedMacroStringObject](macros.Count())
	for _, key := range keys {
		if macroO, ok := macros.Get(key); ok && macroO != nil {
			objects.Add(macroO)
		}
	}
	return objects
}

// PublishStrings merges every chunk of this localization into the datastore,
// mirroring PopulateDataObjectLocalizationsWithIlist: objects are shared by
// ID across localizations, so entries published first (default localization)
// gain extra localized content as the other containers publish afterwards.
// MCR lookups in getStringAtLookupOffsetBinary resolve through
// datastore.GetMacro(gameVersion, chunk*0x100 + index).
func (c *MacroDictionaryBinaryFile) PublishStrings() error {
		gameVersion := c.Version
	macros := c.GetMapObjects()
	var firstErr error
	macros.ForEach(func(key int, macroO datastore.IGlobalLocalizedMacroStringObject) {
		if macroO == nil {
			return
		}
		if datastore.TryAddMacro(gameVersion, key, macroO) {
			return
		}
		existing, ok := datastore.GetMacro(gameVersion, key)
		if !ok || existing == nil {
			datastore.SetMacro(gameVersion, key, macroO)
			return
		}
		content, ok := macroO.GetLocalizedContent(c.Localization)
		if !ok || content == nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("missing content for macro %d localization %s", key, c.Localization)
			}
			return
		}
		existing.SetLocalizedContent(c.Localization, content)
	})
	return firstErr
}

// ExportToJson exports this container localization to a merged-shape JSON
// file (objectfile ExportToJson pattern).
func (c *MacroDictionaryBinaryFile) ExportToJson(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("json file not configured")
	}
	raw, err := MarshalToJson(ExportToJson(map[string]*MacroDictionaryBinaryFile{c.Localization: c}))
	if err != nil {
		return fmt.Errorf("failed to marshal macro dictionary JSON: %w", err)
	}
	if err := common.WriteBytesToFile(filePath, raw); err != nil {
		return fmt.Errorf("failed to write macro dictionary JSON file %s: %w", filePath, err)
	}
	return nil
}

// ImportFromJson imports a merged JSON file, rebuilding every localization
// container found in it, and adopts this container localization entry.
func (c *MacroDictionaryBinaryFile) ImportFromJson(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("json file not configured")
	}
	raw, err := common.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read macro dictionary JSON file %s: %w", filePath, err)
	}
	data, err := UnmarshalJson(raw)
	if err != nil {
		return err
	}
	containers, err := ImportFromJson(data, c.Version)
	if err != nil {
		return err
	}
	rebuilt, ok := containers[c.Localization]
	if !ok || rebuilt == nil {
		return fmt.Errorf("localization %s not found in the JSON file", c.Localization)
	}
	c.Bytes = rebuilt.Bytes
	c.ChunkOffsets = rebuilt.ChunkOffsets
	return nil
}

// SaveToBinary rebuilds this container binary from zero (pointers
// recalculated from the text bytes, first pointer = count*4) and writes it
// to filePath, defaulting to this localization game file.
func (c *MacroDictionaryBinaryFile) SaveToBinary(filePath string) error {
	if filePath == "" {
		filePath = common.MacroBinaryPath(c.Localization)
	}
	files, err := c.Files()
	if err != nil {
		return err
	}
	slotted := make([]*MacroDictionaryTextFile, MacroDictionaryChunkCount)
	for _, f := range files {
		if f == nil {
			continue
		}
		if f.Index < 0 || f.Index >= MacroDictionaryChunkCount {
			return fmt.Errorf("file chunk index out of range: %d", f.Index)
		}
		slotted[f.Index] = f
	}
	raw, err := BuildContainerBinary(slotted)
	if err != nil {
		return err
	}
	if err := common.WriteBytesToFile(filePath, raw); err != nil {
		return fmt.Errorf("failed to write macro dictionary binary %s: %w", filePath, err)
	}
	c.Bytes = raw
	return c.mapBytes()
}
