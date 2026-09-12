package macrodic

import "ffxresources/backend/common"

// MapChunks returns the chunk offsets parsed from the container Bytes.
// Returns a slice of uint32 offsets corresponding to each chunk in the binary data.
// Each offset points to the start of the chunk data within the container Bytes.
func (c *MacroDictionaryBinaryFile) MapChunks() []uint32 {
	if c.Bytes == nil || len(c.ChunkOffsets) == 0 {
		return c.ChunkOffsets
	}
	return c.ChunkOffsets
}

// MapStrings parses a single chunk of binary data into an array of MacroString.
// It resolves the chunk through the container FileAt boundary scan and converts
// the file segments.
func (c *MacroDictionaryBinaryFile) MapStrings(chunkIndex int) []*MacroString {
	f, err := c.FileAt(chunkIndex)
	if err != nil || len(f.Segments) == 0 {
		return nil
	}
	return f.ToMacroStrings()
}

// MapAllStrings maps all chunks in the container to their respective MacroString arrays.
// Returns a 2D slice where each row corresponds to a chunk's strings.
func (c *MacroDictionaryBinaryFile) MapAllStrings() [][]*MacroString {
	if c.Bytes == nil || len(c.ChunkOffsets) == 0 {
		return nil
	}
	result := make([][]*MacroString, len(c.ChunkOffsets))
	for i := range c.ChunkOffsets {
		result[i] = c.MapStrings(i)
	}
	return result
}

// MapStringsWithLocalization maps all chunks with the given localization string.
// This is the main entry point for reading a macro dictionary binary into
// MacroString objects.
func MapStringsWithLocalization(data []byte, localization string, version common.GameVersion) [][]*MacroString {
	container, err := NewMacroDictionaryBinaryFileFromBytes(data, localization, version)
	if err != nil {
		return nil
	}
	return container.MapAllStrings()
}
