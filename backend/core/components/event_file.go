package components

import "strings"

// EventFile represents an event file with script and localized text data
type EventFile struct {
	ID                 string
	EventScript        *AtelScriptObject // This would need to be defined separately
	ScriptBytes        []byte
	JapaneseTextBytes  []byte
	UnknownChunk2Bytes []byte
	FtcxBytes          []byte
	EnglishTextBytes   []byte
	Strings            []*LocalizedFieldStringObject
}

const (
	DefaultAssumedChunkCount = 10
)

// NewEventFile creates a new EventFile from ID and byte data
func NewEventFile(id string, bytes []byte) *EventFile {
	ef := &EventFile{
		ID: id,
	}

	chunks := BytesToChunks(bytes, DefaultAssumedChunkCount, 4)
	ef.mapChunks(chunks)
	ef.mapStrings()

	return ef
}

// bytesToChunks converts byte array to chunks (similar to BytesHelper.bytesToChunks)
/* func (ef *EventFile) bytesToChunks(bytes []byte, assumedChunkCount, pointerSize int) []*Chunk {
	if len(bytes) < pointerSize {
		return []*Chunk{}
	}

	chunks := make([]*Chunk, 0, assumedChunkCount)

	// Read chunk offsets from header
	offsets := make([]int, 0, assumedChunkCount)
	for i := 0; i < len(bytes) && i < assumedChunkCount*pointerSize; i += pointerSize {
		if i+pointerSize <= len(bytes) {
			offset := Read4Bytes(bytes, i)
			if offset == 0 {
				break
			}
			offsets = append(offsets, offset)
		}
	}

	// Create chunks from offsets
	for i, offset := range offsets {
		if offset >= len(bytes) {
			continue
		}

		var chunkBytes []byte
		if i+1 < len(offsets) {
			nextOffset := offsets[i+1]
			if nextOffset <= len(bytes) && nextOffset > offset {
				chunkBytes = bytes[offset:nextOffset]
			}
		} else {
			chunkBytes = bytes[offset:]
		}

		chunk := &Chunk{
			Bytes:  chunkBytes,
			Offset: offset,
			Length: len(chunkBytes),
		}
		chunks = append(chunks, chunk)
	}

	return chunks
} */

// mapChunks maps the chunks to specific byte arrays
func (ef *EventFile) mapChunks(chunks []Chunk) {
	if len(chunks) > 0 {
		ef.ScriptBytes = chunks[0].Bytes
	}
	if len(chunks) > 1 {
		ef.JapaneseTextBytes = chunks[1].Bytes
	}
	if len(chunks) > 2 {
		ef.UnknownChunk2Bytes = chunks[2].Bytes
	}
	if len(chunks) > 3 {
		ef.FtcxBytes = chunks[3].Bytes
	}
	if len(chunks) > 4 && chunks[4].Offset != 0 {
		ef.EnglishTextBytes = chunks[4].Bytes
	}
}

// mapStrings converts byte data to localized string objects
func (ef *EventFile) mapStrings() {
	// Map Japanese strings
	if len(ef.JapaneseTextBytes) > 0 {
		japaneseStrings, err := FromFieldStringData(ef.JapaneseTextBytes, false, "jp")
		if err == nil && japaneseStrings != nil {
			localizedJpStringObjects := make([]*LocalizedFieldStringObject, len(japaneseStrings))
			for i, str := range japaneseStrings {
				localizedJpStringObjects[i] = NewLocalizedFieldStringObject()
				localizedJpStringObjects[i].SetLocalizedContent("jp", str)
			}
			ef.AddLocalizations(localizedJpStringObjects)
		}
	}

	// Map English strings
	if len(ef.EnglishTextBytes) > 0 {
		englishStrings, err := FromFieldStringData(ef.EnglishTextBytes, false, "us")
		if err == nil && englishStrings != nil {
			localizedUsStringObjects := make([]*LocalizedFieldStringObject, len(englishStrings))
			for i, str := range englishStrings {
				localizedUsStringObjects[i] = NewLocalizedFieldStringObject()
				localizedUsStringObjects[i].SetLocalizedContent("us", str)
			}
			ef.AddLocalizations(localizedUsStringObjects)
		}
	}
}

// AddLocalizations adds or merges localized string objects
func (ef *EventFile) AddLocalizations(strings []*LocalizedFieldStringObject) {
	if ef.Strings == nil {
		ef.Strings = strings
		return
	}

	for i, localizationStringObject := range strings {
		if i < len(ef.Strings) {
			stringObject := ef.Strings[i]
			if stringObject != nil && localizationStringObject != nil {
				localizationStringObject.CopyInto(stringObject)
			}
		} else {
			ef.Strings = append(ef.Strings, localizationStringObject)
		}
	}
}

// ToBytes converts the EventFile back to byte format
// Rebuilds the file in the same chunk order as mapChunks separates it
func (ef *EventFile) ToBytes() []byte {
	chunks := make([][]byte, 0, 5)

	// Chunk 0: Script bytes
	chunks = append(chunks, ef.ScriptBytes)

	// Chunk 1: Japanese text bytes (updated from strings if available)
	ef.UpdateTextBytes("jp")
	chunks = append(chunks, ef.JapaneseTextBytes)

	// Chunk 2: Unknown chunk
	chunks = append(chunks, ef.UnknownChunk2Bytes)

	// Chunk 3: FTCX bytes
	chunks = append(chunks, ef.FtcxBytes)

	// Chunk 4: English text bytes (if available, updated from strings)
	if len(ef.EnglishTextBytes) > 0 {
		ef.UpdateTextBytes("us")
		chunks = append(chunks, ef.EnglishTextBytes)
	}

	return ef.chunksToBytes(chunks)
}

// UpdateTextBytes updates the Japanese or English text bytes from current strings
// Accepts "jp" or "us" to specify which localization to update
// This function takes the strings array and converts them back to the appropriate text bytes
func (ef *EventFile) UpdateTextBytes(localization string) {
	textBytes := ef.stringsToStringFileBytes(localization)

	switch strings.ToLower(localization) {
	case "jp", "japanese":
		ef.JapaneseTextBytes = textBytes
	case "us", "english":
		ef.EnglishTextBytes = textBytes
	default:
		// For other localizations, you might want to handle differently
		// For now, default to English
		ef.EnglishTextBytes = textBytes
	}
}

// stringsToStringFileBytes converts localized strings to byte format for specific localization
func (ef *EventFile) stringsToStringFileBytes(localization string) []byte {
	if len(ef.Strings) == 0 {
		return []byte{}
	}

	// Extract FieldString objects for the specified localization
	fieldStrings := make([]*FieldString, 0, len(ef.Strings))
	charset := LocalizationToCharset(localization)

	for _, localizedObj := range ef.Strings {
		if localizedObj != nil {
			fieldString := localizedObj.GetLocalizedContent(localization)
			if fieldString != nil {
				fieldStrings = append(fieldStrings, fieldString)
			} else {
				// Create empty field string if no content for this localization
				emptyFieldString := &FieldString{
					Charset: charset,
				}
				fieldStrings = append(fieldStrings, emptyFieldString)
			}
		}
	}

	// Use RebuildFieldStrings to convert back to bytes
	return RebuildFieldStrings(fieldStrings, charset, true)
}

// chunksToBytes converts chunks back to a single byte array with proper headers
func (ef *EventFile) chunksToBytes(chunks [][]byte) []byte {
	if len(chunks) == 0 {
		return []byte{}
	}

	// Calculate header size (4 bytes per chunk + 4 bytes for end marker)
	headerSize := (len(chunks) + 1) * 4

	// Calculate chunk offsets
	offsets := make([]int, len(chunks)+1)
	currentOffset := headerSize

	for i, chunk := range chunks {
		if len(chunk) == 0 {
			offsets[i] = 0
		} else {
			offsets[i] = currentOffset
			currentOffset += len(chunk)
		}
	}
	offsets[len(chunks)] = currentOffset // End marker

	// Build the result
	result := make([]byte, currentOffset)

	// Write header (offsets)
	for i, offset := range offsets {
		write4BytesLE(result, i*4, uint32(offset))
	}

	// Write chunk data
	for i, chunk := range chunks {
		if len(chunk) > 0 && offsets[i] != 0 {
			copy(result[offsets[i]:], chunk)
		}
	}

	return result
}

// write4BytesLE writes a 4-byte integer in little-endian format
func write4BytesLE(data []byte, offset int, value uint32) {
	if offset+3 < len(data) {
		data[offset] = byte(value & 0xFF)
		data[offset+1] = byte((value >> 8) & 0xFF)
		data[offset+2] = byte((value >> 16) & 0xFF)
		data[offset+3] = byte((value >> 24) & 0xFF)
	}
}

// String returns string representation of the EventFile
func (ef *EventFile) String() string {
	var builder strings.Builder
	builder.WriteString(ef.GetName())
	builder.WriteString("\n")
	return builder.String()
}

// GetName returns the name of the event file
func (ef *EventFile) GetName() string {
	return ef.ID
}

// GetNameWithLocalization returns the name with localization-specific area name
func (ef *EventFile) GetNameWithLocalization(localization string) string {
	// This would need access to MACRO_LOOKUP and EventScript implementation
	// For now, return the basic name
	return ef.GetName()
}

// AtelScriptObject placeholder - this would need to be implemented based on your needs
type AtelScriptObject struct {
	AreaNameIndexes []int
	// Add other fields as needed
}
