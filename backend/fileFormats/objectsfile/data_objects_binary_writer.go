package objectsfile

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
	"fmt"
	"path/filepath"
)

// ExportLocalizedTextData exports localized text data from an IList to binary files
// for all supported localizations. This function converts in-memory localized objects
// back to their original binary format and writes them to the appropriate files.
//
// The function validates that objects are loaded, converts the IList to a slice,
// and delegates to the lower-level writing functions to handle the actual file operations.
//
// Parameters:
//   - objects: IList containing ILocalizedTextObject entries to be exported
//   - pathPattern: Relative path pattern for the binary files (e.g., "battle/kernel/command.bin")
//
// Returns: error if no objects are loaded or if the write operation fails
func ExportLocalizedTextData(objects components.IList[datastore.IGlobalLocalizedTextObject], pathPattern string) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("nenhum objeto carregado")
	}

	dataObjects := make([]datastore.IGlobalLocalizedTextObject, objects.Len())
	for i, obj := range objects.Items() {
		dataObjects[i] = obj
	}

	from := 0
	to := len(dataObjects)

	return writeLocalizedDataObjectsInAllLocalizations(pathPattern, dataObjects, from, to)
}

// writeLocalizedDataObjectsInAllLocalizations writes localized data objects to binary files
// for all supported localizations. This function handles the core logic of converting
// in-memory objects back to binary format and writing them to the appropriate file paths.
//
// The function iterates through all supported languages, converts objects to bytes for each
// localization, ensures directory structure exists, and writes the binary data to files.
//
// Parameters:
//   - pathPattern: Relative path pattern for the binary files
//   - objects: Slice of ILocalizedTextObject instances to be written
//   - startIndex: Starting index in the objects slice (inclusive)
//   - endIndex: Ending index in the objects slice (exclusive)
//
// Returns: error if indices are invalid or if any write operation fails
func writeLocalizedDataObjectsInAllLocalizations(pathPattern string, objects []datastore.IGlobalLocalizedTextObject, startIndex, endIndex int) error {
	common.LogVerbose("Writing localized data objects to: %s (from %d to %d)", pathPattern, startIndex, endIndex)

	if startIndex < 0 || endIndex > len(objects) || startIndex >= endIndex {
		return fmt.Errorf("invalid indices: from=%d, to=%d, length=%d", startIndex, endIndex, len(objects))
	}

	for localizationKey := range common.SupportedLanguages {
		if localizationKey != "us" {
			continue // Skip non-US localizations for now
		}
		localizationRoot := common.GetLocalizationRoot(localizationKey)
		localePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, localizationRoot, pathPattern)
		localePath = filepath.FromSlash(localePath)

		dataBytes, err := ConvertFFXLocalizedDataToBytes(objects, startIndex, endIndex, localizationKey)
		if err != nil {
			return fmt.Errorf("error when converting objects to bytes (localization %s): %w", localizationKey, err)
		}

		dir := filepath.Dir(localePath)
		if err := common.EnsurePathExists(dir); err != nil {
			return fmt.Errorf("error when creating directory %s: %w", dir, err)
		}

		if err := common.WriteBytesToFile(localePath, dataBytes); err != nil {
			return fmt.Errorf("error when writing file %s: %w", localePath, err)
		}

		common.LogVerbose("Wrote localized data to %s (%d bytes)", localePath, len(dataBytes))
	}

	return nil
}

// ConvertFFXLocalizedDataToBytes converts a slice of ILocalizedTextObject instances to binary format
// for a specific localization. This function recreates the original binary file structure including
// headers, object data, and string data sections.
//
// The binary format consists of:
// 1. FFX header (8 bytes)
// 2. Object range information (4 bytes)
// 3. Header length and total length (4 bytes)
// 4. Unknown header bytes (4 bytes)
// 5. Object data section (variable length)
// 6. String data section (variable length)
//
// Parameters:
//   - objects: Slice of ILocalizedTextObject instances to convert
//   - from: Starting index in the objects slice
//   - to: Ending index in the objects slice
//   - languageCode: Language code for the target localization
//
// Returns: byte slice containing the binary data, or error if conversion fails
func ConvertFFXLocalizedDataToBytes(objects []datastore.IGlobalLocalizedTextObject, from, to int, languageCode string) ([]byte, error) {
	if len(objects) == 0 {
		return nil, fmt.Errorf("objects slice is empty")
	}

	objectHeaderLength := objects[0].GetHeaderLength()

	ffxHeader := []byte{1, 0, 0, 0, 0, 0, 0, 0}
	ffxUnknownBytesOfHeader := []byte{0x14, 0x00, 0x00, 0x00}

	var allKeyedStrings []datastore.IGlobalKeyedString
	for _, obj := range objects {
		keyedStrings := obj.GetLocalizedKeyedStrings(languageCode)
		for _, ks := range keyedStrings {
			if ks != nil {
				allKeyedStrings = append(allKeyedStrings, ks)
			} else {
				// TODO: delete this
				common.LogVerbose("Keyed string is nil for object at index %d", obj.GetName(common.DefaultLocalization))
			}
		}
	}

	charset := ffxencoding.GetCharsetForLanguage(languageCode)
	stringBytes := RebuildKeyedStrings(allKeyedStrings, charset)

	var buf bytes.Buffer

	// Write FFX header
	buf.Write(ffxHeader)

	// Write object range (from index)
	binary.Write(&buf, binary.LittleEndian, uint16(from))

	// Write object range (to index - 1)
	toMinus1 := to - 1
	binary.Write(&buf, binary.LittleEndian, uint16(toMinus1))

	// Write header length per object
	binary.Write(&buf, binary.LittleEndian, uint16(objectHeaderLength))

	// Write total header section length
	totalLength := len(objects) * objectHeaderLength
	binary.Write(&buf, binary.LittleEndian, uint16(totalLength))

	// Write unknown header bytes
	buf.Write(ffxUnknownBytesOfHeader)

	// Write object data section
	for i, obj := range objects {
		objBytes, err := obj.ToBytes(languageCode)
		if err != nil {
			return nil, fmt.Errorf("error converting object %d to bytes: %w", i, err)
		}
		buf.Write(objBytes)
	}

	// Write string data section
	buf.Write(stringBytes)

	return buf.Bytes(), nil
}

// writeStringSegments writes a contiguous sequence of 4-byte string reference segments
// into a pre-allocated byte slice, populating each segment from the corresponding
// IGlobalLocalizedKeyedStringObject's localized content.
//
// Each segment is exactly 4 bytes: 2 bytes for the string table offset (uint16 LE)
// and 2 bytes for the string key (uint16 LE). Segments are written sequentially
// starting at the given offset, with each subsequent segment at start+i*4. The caller
// must ensure that `data` is large enough to hold all segments.
//
// This function is the write-side counterpart of readStringSegments. It is intended
// exclusively for headers where ALL string segments are stored contiguously, with no
// gaps or non-string bytes between them. Examples:
//   - CommandTextObject: Name → SimplifiedName → Description → SimplifiedDescription
//   - NameOnlyTextObject: Name → SimplifiedName
//   - CommandTextObjectV2: Name → Description
//
// It must NOT be used for segments that are written at arbitrary or non-sequential
// positions within the binary chunk (e.g., JobTextObject where the Effect
// segment is at a separate position offset). For those cases, use direct
// models.WriteSegmentAt instead.
//
// Segments with nil localized content are skipped, preserving the original bytes
// in the data slice. The function stops and returns an error on the first write
// failure without processing remaining segments.
//
// Parameters:
//   - data: The pre-allocated byte slice to write into.
//   - start: The byte offset where the first segment is written.
//   - languageCode: The localization language code (e.g., "us", "jp", "de") used to
//     resolve each segment's localized content via GetLocalizedContent.
//   - segments: One or more IGlobalLocalizedKeyedStringObject instances to serialize,
//     listed in the exact order they should appear in the binary output.
//
// Returns: nil on success, or a wrapped fmt.Errorf indicating which segment index
// failed to write (e.g., "writing segment 2: <underlying error>").
func writeStringSegments(data []byte, start int, languageCode string, segments ...datastore.IGlobalLocalizedKeyedStringObject) error {
	for i, seg := range segments {
		content := seg.GetLocalizedContent(languageCode)
		if content == nil {
			continue
		}
		if err := models.WriteSegmentAt(data, start+i*4, getSegment(content)); err != nil {
			return fmt.Errorf("writing segment %d: %w", i, err)
		}
	}
	return nil
}