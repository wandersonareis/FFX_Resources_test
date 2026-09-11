package macrodic

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
	"fmt"
)

// MacroDictionaryChunkCount is the number of 4-byte chunk offsets in the
// macro dictionary header. Each offset is a uint32 pointing to the start
// of a chunk within the container Bytes.
const MacroDictionaryChunkCount = 16

// macroDictionaryHeaderLength is the byte length of the chunk offset table.
const macroDictionaryHeaderLength = MacroDictionaryChunkCount * 4

// macroDictionaryTerminator marks the end of the chunk offset table.
const macroDictionaryTerminator = 0xFFFFFFFF

// MacroDictionaryTextContainer holds the raw bytes of a macro dictionary file along
// with chunk offsets computed from those bytes. The container is the single source
// of truth for binary reconstruction.
type MacroDictionaryTextContainer struct {
	Bytes        []byte
	ChunkOffsets []uint32
	Localization string
	Version      int
}

// NewMacroDictionaryTextContainer creates a container from raw bytes, parsing the
// chunk offset table via mapBytes.
func NewMacroDictionaryTextContainer(data []byte, localization string, version int) (*MacroDictionaryTextContainer, error) {
	if len(data) < macroDictionaryHeaderLength {
		return nil, fmt.Errorf("insufficient data to create MacroDictionaryTextContainer: have %d bytes, need at least %d", len(data), macroDictionaryHeaderLength)
	}

	c := &MacroDictionaryTextContainer{
		Bytes:        data,
		Localization: localization,
		Version:      version,
	}
	if err := c.mapBytes(); err != nil {
		return nil, err
	}
	return c, nil
}

// mapBytes parses the chunk offset table at the start of the container Bytes.
// Each entry is read as uint32. Zero offsets are kept as empty-chunk placeholders.
//
// The parsing mirrors components.BytesToChunks semantics so the container holds
// exactly the offsets the legacy NewMacroDictionaryFile path would use: when the
// 0xFFFFFFFF terminator is found at slot i, only the first i-1 offsets are kept.
func (c *MacroDictionaryTextContainer) mapBytes() error {
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
func (c *MacroDictionaryTextContainer) boundsOf(chunkIndex int) (chunkBounds, bool) {
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
func (c *MacroDictionaryTextContainer) Files() ([]*MacroDictionaryTextFile, error) {
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
func (c *MacroDictionaryTextContainer) FileAt(chunkIndex int) (*MacroDictionaryTextFile, error) {
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

// PublishStrings publishes every chunk of the container into the datastore,
// mirroring the legacy MacroDictionaryFile.PublishStrings behavior on the new
// format. MCR lookups in getStringAtLookupOffsetBinary resolve through
// datastore.GetMacro(gameVersion, chunk*0x100 + index), populated here via
// datastore.SetMacro.
func (c *MacroDictionaryTextContainer) PublishStrings() {
	for i := range c.ChunkOffsets {
		c.publishStringsOfChunk(i)
	}
}

func (c *MacroDictionaryTextContainer) publishStringsOfChunk(chunkIndex int) {
	f, err := c.FileAt(chunkIndex)
	if err != nil || len(f.Segments) == 0 {
		return
	}
	gameVersion := models.GameVersion(c.Version)
	for j, macroStr := range f.ToMacroStrings() {
		if macroStr == nil {
			continue
		}
		key := chunkIndex*0x100 + j
		macroO, exist := datastore.GetMacro(gameVersion, key)
		if !exist {
			macroO = NewLocalizedMacroStringObject()
			datastore.SetMacro(gameVersion, key, macroO)
		}
		macroO.SetLocalizedContent(c.Localization, macroStr)
	}
}

// ExportToJson exports this container to the merged JSON format (objectfile
// ExportToJson pattern): the entry Name/SimplifiedName maps hold this
// container localization texts.
func (c *MacroDictionaryTextContainer) ExportToJson() *MacroDictionaryJsonExport {
	return ExportToJson(map[string]*MacroDictionaryTextContainer{c.Localization: c})
}

// ImportFromJson imports a merged JSON representation, rebuilding one binary
// container per localization found in the data. A missing/empty SimplifiedName
// entry falls back to the Name pointer in the rebuilt binary.
func (c *MacroDictionaryTextContainer) ImportFromJson(data *MacroDictionaryJsonImport, version int) (map[string]*MacroDictionaryTextContainer, error) {
	return ImportFromJson(data, version)
}
