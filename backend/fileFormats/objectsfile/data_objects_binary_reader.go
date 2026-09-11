package objectsfile

import (
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// PopulateDataObjectLocalizationsWithIlist populates localization data for all supported languages
// by reading corresponding binary files for each language and merging the localized content
// into the existing LocalizedTextObject instances within the provided IList.
//
// This function iterates through all supported languages, reads the binary files for each
// language, and applies the localized text content to the corresponding objects in the IList,
// leveraging IList's content() method for efficient access to the underlying collection.
//
// Parameters:
//   - path: Relative path to the binary file pattern (e.g., "battle/kernel/command.bin")
//   - objects: IList[ILocalizedTextObject] containing instances to be populated with localizations
//   - creator: Function that creates LocalizedTextObject instances from binary data
//   - version: Game version as int (1 = FFX, 2 = FFX-2)
func PopulateDataObjectLocalizationsWithIlist(path string, objects components.IList[datastore.IGlobalLocalizedTextObject], creator func([]byte, []byte, int, string) (datastore.IGlobalLocalizedTextObject, error), version int) {
	if objects == nil || objects.IsEmpty() {
		return
	}

	for locKey := range common.SupportedLanguages {
		fullPath := filepath.Join(common.GetLocalizationRoot(locKey), path)

		var localizationData components.IList[datastore.IGlobalLocalizedTextObject]
		if version == 2 {
			localizationData = ReadDataListWithIlistV2(fullPath, locKey, creator)
		} else {
			localizationData = ReadDataListWithIlist(fullPath, locKey, creator, version)
		}
		if localizationData != nil {
			maxLen := min(localizationData.Len(), objects.Len())
			items := objects.Items()
			locItems := localizationData.Items()
			for i := range maxLen {
				if items[i] == nil || locItems[i] == nil {
					continue
				}
				items[i].SetLocalizations(locItems[i])
			}
		}
	}
}

// ReadDataListWithIlist reads binary data files and creates localized text objects using a custom creator function.
//
// This is a high-level function that handles file reading and delegates to ParseDataListWithIlist for data processing.
// It provides convenient access to binary data files by handling file system operations and error checking,
// then passes the raw binary data to the parsing function for object creation.
//
// Parameters:
//   - filename: Path to the binary file to read
//   - languageCode: Language code for localization (e.g., "us", "jp", "de")
//   - creator: Function that creates ILocalizedTextObject instances from binary data
//     Parameters: (objData []byte, stringBytes []byte, headerLength int, languageCode string)
//   - version: Game version as int (1 = FFX, 2 = FFX-2)
//
// Returns: IList[ILocalizedTextObject] containing localized text entries, or nil if reading fails
func ReadDataListWithIlist(filename string, languageCode string, creator func([]byte, []byte, int, string) (datastore.IGlobalLocalizedTextObject, error), version int) components.IList[datastore.IGlobalLocalizedTextObject] {
	fileAccessor, err := common.NewFileAccessor(filename)
	if err != nil {
		common.LogVerbose("Error accessing file: %v", err)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	if !fileAccessor.Exists {
		common.LogVerbose("File does not exist: %s", filename)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	data, err := os.ReadFile(fileAccessor.ResolvedPath)
	if err != nil {
		common.LogVerbose("Error reading file: %v", err)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	return ParseDataListWithIlist(data, languageCode, creator, version)
}

// ParseDataListWithIlist parses binary data and creates localized text objects using a custom creator function.
//
// This is a low-level function that provides flexible binary data parsing for any type of localized text object.
// It handles the standard binary format used by the game's localization files, which includes a header section
// with object metadata and data sections containing object-specific information and string data.
//
// The function parses the binary data structure:
// - 8 bytes: header information
// - 2 bytes: minimum index value
// - 2 bytes: maximum index value
// - 2 bytes: individual object length
// - 2 bytes: total data section length
// - 4 bytes: reserved/padding
// - Variable: object data section
// - Variable: string data section
//
// Parameters:
//   - data: Binary data to parse
//   - languageCode: Language code for localization (e.g., "us", "jp", "de")
//   - creator: Function that creates ILocalizedTextObject instances from binary data
//     Parameters: (objData []byte, stringBytes []byte, headerLength int, languageCode string)
//
// Returns: IList[ILocalizedTextObject] containing localized text entries, or nil if parsing fails
func ParseDataListWithIlist(data []byte, languageCode string, creator func([]byte, []byte, int, string) (datastore.IGlobalLocalizedTextObject, error), version int) components.IList[datastore.IGlobalLocalizedTextObject] {
	// No FFX-2 (v2) o cabeçalho é de 0x20 bytes e usa campos uint32 a partir do
	// offset 16. Quando a versão do jogo é FFX-2, delegamos ao parser V2 (que lê
	// offset := 16 e uint32), idêntico a ParseDataListWithIlistV2.
	if version == 2 {
		return ParseDataListWithIlistV2(data, languageCode, creator)
	}

	if len(data) < 16 { // Minimum header size
		common.LogVerbose("Data too small for valid binary format")
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	offset := 8

	minIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	maxIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Read individual length and total length
	individualLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	totalLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Skip 4 bytes
	offset += 4

	// Verify we have enough data
	if offset+totalLength > len(data) {
		common.LogVerbose("Insufficient data for specified total length")
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	// Read data bytes
	dataBytes := data[offset : offset+totalLength]
	offset += totalLength

	// Read string bytes (remaining data)
	stringBytes := data[offset:]

	// Create IList with pre-calculated capacity for optimal memory allocation
	count := maxIndex - minIndex
	objectCount := count + 1
	objects := components.NewList[datastore.IGlobalLocalizedTextObject](objectCount)

	for i := 0; i <= count; i++ {
		from := i * individualLength
		to := (i + 1) * individualLength

		if to > len(dataBytes) {
			break
		}
		objData := dataBytes[from:to]
		obj, err := creator(objData, stringBytes, individualLength, languageCode)
		if err != nil {
			common.LogVerbose("Error creating object at index %d: %v", i+minIndex, err)
			continue
		}
		if obj == nil {
			continue
		}
		objects.Add(obj)

		if common.IsVerboseMode() {
			offsetStr := fmt.Sprintf("%04X", (i*individualLength)+0x14)
			indexStr := fmt.Sprintf("%d", i+minIndex)
			objectString := obj.ToString(languageCode)
			fmt.Printf("%s (Offset %s) - %s\n", indexStr, offsetStr, objectString)
		}
	}

	return objects
}

// ReadDataListWithIlistV2 is the FFX-2 (version 2) counterpart of ReadDataListWithIlist.
// It reads the binary file and delegates to ParseDataListWithIlistV2 for parsing.
func ReadDataListWithIlistV2(filename string, languageCode string, creator func([]byte, []byte, int, string) (datastore.IGlobalLocalizedTextObject, error)) components.IList[datastore.IGlobalLocalizedTextObject] {
	fileAccessor, err := common.NewFileAccessor(filename)
	if err != nil {
		common.LogVerbose("Error accessing file: %v", err)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	if !fileAccessor.Exists {
		common.LogVerbose("File does not exist: %s", filename)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	data, err := os.ReadFile(fileAccessor.ResolvedPath)
	if err != nil {
		common.LogVerbose("Error reading file: %v", err)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	return ParseDataListWithIlistV2(data, languageCode, creator)
}

// ParseDataListWithIlistV2 parses the FFX-2 (version 2) binary data format and creates
// localized text objects using a custom creator function.
//
// The V2 header is 0x20 bytes long and uses uint32 fields instead of the uint16 fields
// used by the V1 format:
//   - 8 bytes: signature / unknown
//   - uint32: minimum index
//   - uint32: maximum index
//   - uint32: individual object length
//   - uint32: total data section length
//   - 8 bytes: reserved (0x18..0x1F)
//   - Variable: object data section (totalLength bytes)
//   - Variable: string data section (remainder)
//
// ParseDataListWithIlistV2 faz o parse do formato V2 com fallbacks estáticos
func ParseDataListWithIlistV2(data []byte, languageCode string, creator func([]byte, []byte, int, string) (datastore.IGlobalLocalizedTextObject, error)) components.IList[datastore.IGlobalLocalizedTextObject] {
	if len(data) < 32 {
		common.LogVerbose("Data too small for valid binary format v2")
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	offset := 16

	// 2. Lê o número de chunks (age como maxIndex na V2, já que minIndex é sempre 0)
	maxIndex := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	minIndex := 0 // Na V2, o índice inicial é sempre 0

	individualLength := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	totalLength := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	expectedTotalLength := (maxIndex - minIndex + 1) * individualLength
	if totalLength != expectedTotalLength {
		common.LogVerbose("TotalLength divergente, aplicando fallback calculado")
		totalLength = expectedTotalLength
	}

	offset = 32

	if offset+totalLength > len(data) {
		common.LogVerbose("Insufficient data for specified total length")
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	dataBytes := data[offset : offset+totalLength]
	offset += totalLength

	stringBytes := data[offset:]

	count := maxIndex - minIndex
	objectCount := count + 1
	objects := components.NewList[datastore.IGlobalLocalizedTextObject](objectCount)

	chunkData := make([]byte, individualLength)
	defer func() {
		chunkData = nil
	}()

	for i := 0; i <= count; i++ {
		from := i * individualLength
		to := (i + 1) * individualLength

		if to > len(dataBytes) {
			break
		}
		copy(chunkData, dataBytes[from:to])
		obj, err := creator(chunkData, stringBytes, individualLength, languageCode)
		if err != nil {
			common.LogVerbose("Error creating V2 object at index %d: %v", i+minIndex, err)
			continue
		}
		if obj == nil {
			common.LogVerbose("Skipping invalid V2 object at index %d", i+minIndex)
			continue
		}
		objects.Add(obj)

		if common.IsVerboseMode() {
			offsetStr := fmt.Sprintf("%04X", 0x20+(i*individualLength))
			indexStr := fmt.Sprintf("%d", i+minIndex)
			objectString := obj.ToString(languageCode)
			fmt.Printf("%s (Offset %s) - %s\n", indexStr, offsetStr, objectString)
		}
	}

	return objects
}

// readStringSegments reads a contiguous sequence of 4-byte string reference segments
// from a binary reader and populates the corresponding IGlobalLocalizedKeyedStringObject
// instances with their localized content.
//
// Each segment is exactly 4 bytes: 2 bytes for the string table offset (uint16 LE)
// and 2 bytes for the string key (uint16 LE). The segments are read sequentially
// from the reader's current position, one after another, in the order they appear
// in the binary data. No seeking or position calculation is performed — the reader
// must already be positioned at the first segment.
//
// This function is intended exclusively for headers where ALL string segments are
// stored contiguously at the beginning of the chunk, with no gaps or non-string
// bytes between them. Examples:
//   - CommandTextObject: Name → SimplifiedName → Description → SimplifiedDescription
//   - MonsterTextObject: Name → SensorText → SimplifiedSensorText → ScanText → SimplifiedScanText
//   - NameOnlyTextObject: Name → SimplifiedName
//
// It must NOT be used for segments that are located at arbitrary or non-sequential
// positions within the binary chunk (e.g., JobTextObject where the Effect
// segment is at a separate position offset). For those cases, use direct
// io.Reader.Seek + models.ReadSegment instead.
//
// Parameters:
//   - r: The io.Reader positioned exactly at the start of the first segment to read.
//     The caller is responsible for seeking or positioning the reader beforehand.
//   - stringBytes: The complete raw string table blob from the binary file, used by
//     each segment to resolve its offset+key into actual localized string content.
//   - languageCode: The localization language code (e.g., "us", "jp", "de") that
//     determines the character encoding used when interpreting the raw bytes.
//   - segments: One or more IGlobalLocalizedKeyedStringObject instances to populate,
//     listed in the exact order they appear in the binary stream. Each instance will
//     have its localized content set via ReadAndSetLocalizedContent.
//
// Returns: nil on success, or a wrapped fmt.Errorf indicating which segment index
// failed to read (e.g., "reading segment 2: EOF").
func readStringSegments(r io.Reader, stringBytes []byte, languageCode string, version int, segments ...datastore.IGlobalLocalizedKeyedStringObject) error {
	for i, seg := range segments {
		s, err := models.ReadSegment(r)
		if err != nil {
			return fmt.Errorf("reading segment %d: %w", i, err)
		}
		seg.ReadAndSetLocalizedContent(languageCode, stringBytes, s.Offset, s.Key, version)
	}
	return nil
}
