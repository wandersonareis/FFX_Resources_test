package objectsfile

import (
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"fmt"
	"os"
	"path/filepath"
)

// ReadNameOnlyDataObjectsWithIlist reads binary data from a specified pattern path
// and creates NameOnlyDataObject entries with all available localizations.
//
// This function reads LocalizedTextObject entries containing only name information
// for various game elements that don't require descriptions. Each entry includes
// localized text for all supported languages in the game, making it suitable for
// simple text elements like labels, button text, or single-word entries.
//
// File format: name only data (binary format)
// Pattern path: Variable, passed as parameter (e.g., "battle/kernel/btl_txt.bin")
//
// Parameters:
//   - patternPath: Relative path to the binary file within the localization directory
//
// Returns: IList[ILocalizedTextObject] containing NameOnlyDataObject entries with full localization data
func ReadNameOnlyDataObjectsWithIlist(patternPath string) components.IList[datastore.IGlobalLocalizedTextObject] {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

	creator := func(data []byte, stringBytes []byte, headerLength int, localization string) datastore.IGlobalLocalizedTextObject {
		var obj datastore.IGlobalLocalizedTextObject
		if common.GetGameVersionString() == "ffx2" {
			obj = NewNameOnlyDataObjectV2(data, stringBytes, headerLength, localization)
		} else {
			obj = NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
		}
		if obj == nil {
			return nil
		}
		return obj
	}

	nameObjects := ReadDataListWithIlist(filePath, common.DefaultLocalization, creator)
	if nameObjects == nil || nameObjects.IsEmpty() {
		common.LogVerbose("No name-only data objects found for %s\n", patternPath)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	PopulateDataObjectLocalizationsWithIlist(patternPath, nameObjects, creator)

	common.LogVerbose("Loading %d name-only data objects...\n", nameObjects.Len())

	return nameObjects
}

// ReadNameDescriptionObjectsWithIlist reads binary data from a specified pattern path
// and creates NameDescriptionTextObject entries with all available localizations.
//
// This function reads LocalizedTextObject entries containing both name and description information
// for various game elements that require detailed text. Each entry includes localized text for all
// supported languages in the game, making it suitable for complex game objects like items, commands,
// abilities, and other elements that need both a title and detailed description.
//
// File format: name and description data (binary format)
// Pattern path: Variable, passed as parameter (e.g., "battle/kernel/command.bin")
//
// Parameters:
//   - patternPath: Relative path to the binary file within the localization directory
//
// Returns: IList[ILocalizedTextObject] containing NameDescriptionTextObject entries with full localization data
func ReadNameDescriptionObjectsWithIlist(patternPath string) components.IList[datastore.IGlobalLocalizedTextObject] {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

	creator := func(data []byte, stringBytes []byte, headerLength int, localization string) datastore.IGlobalLocalizedTextObject {
		if common.GetGameVersionString() == "ffx2" {
			if obj := NewNameDescriptionTextObjectV2(data, stringBytes, headerLength, localization); obj != nil {
				return obj
			}
			return nil
		}
		if obj := NewNameDescriptionTextObject(data, stringBytes, headerLength, localization); obj != nil {
			return obj
		}
		return nil
	}

	var nameDescObjects components.IList[datastore.IGlobalLocalizedTextObject]
	if common.GetGameVersionString() == "ffx2" {
		nameDescObjects = ReadDataListWithIlistV2(filePath, common.DefaultLocalization, creator)
	} else {
		nameDescObjects = ReadDataListWithIlist(filePath, common.DefaultLocalization, creator)
	}
	if nameDescObjects == nil || nameDescObjects.IsEmpty() {
		common.LogVerbose("No name and description objects found for %s\n", patternPath)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	PopulateDataObjectLocalizationsWithIlist(patternPath, nameDescObjects, creator)

	common.LogVerbose("Loading %d name and description objects...\n", nameDescObjects.Len())

	return nameDescObjects
}

// ReadJobObjectsWithIlist reads job data from the job.bin file (FFX-2 only).
// and creates JobTextObject entries with all available localizations.
func ReadJobObjectsWithIlist(patternPath string) components.IList[datastore.IGlobalLocalizedTextObject] {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

	creator := func(data []byte, stringBytes []byte, headerLength int, localization string) datastore.IGlobalLocalizedTextObject {
		if common.GetGameVersionString() == "ffx2" {
			if obj := NewJobTextObject(data, stringBytes, headerLength, localization); obj != nil {
				return obj
			}
			return nil
		}
		if obj := NewNameDescriptionTextObject(data, stringBytes, headerLength, localization); obj != nil {
			return obj
		}
		return nil
	}

	var jobObjects components.IList[datastore.IGlobalLocalizedTextObject]
	if common.GetGameVersionString() == "ffx2" {
		jobObjects = ReadDataListWithIlistV2(filePath, common.DefaultLocalization, creator)
	} else {
		jobObjects = ReadDataListWithIlist(filePath, common.DefaultLocalization, creator)
	}
	if jobObjects == nil || jobObjects.IsEmpty() {
		common.LogVerbose("No job objects found for %s\n", patternPath)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	PopulateDataObjectLocalizationsWithIlist(patternPath, jobObjects, creator)

	common.LogVerbose("Loading %d job objects...\n", jobObjects.Len())

	return jobObjects
}

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
func PopulateDataObjectLocalizationsWithIlist(path string, objects components.IList[datastore.IGlobalLocalizedTextObject], creator func([]byte, []byte, int, string) datastore.IGlobalLocalizedTextObject) {
	if objects == nil || objects.IsEmpty() {
		return
	}

	for locKey := range common.SupportedLanguages {
		fullPath := filepath.Join(common.GetLocalizationRoot(locKey), path)

		var localizationData components.IList[datastore.IGlobalLocalizedTextObject]
		if common.GetGameVersionString() == "ffx2" {
			localizationData = ReadDataListWithIlistV2(fullPath, locKey, creator)
		} else {
			localizationData = ReadDataListWithIlist(fullPath, locKey, creator)
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
//
// Returns: IList[ILocalizedTextObject] containing localized text entries, or nil if reading fails
func ReadDataListWithIlist(filename string, languageCode string, creator func([]byte, []byte, int, string) datastore.IGlobalLocalizedTextObject) components.IList[datastore.IGlobalLocalizedTextObject] {
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

	return ParseDataListWithIlist(data, languageCode, creator)
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
func ParseDataListWithIlist(data []byte, languageCode string, creator func([]byte, []byte, int, string) datastore.IGlobalLocalizedTextObject) components.IList[datastore.IGlobalLocalizedTextObject] {
	// No FFX-2 (v2) o cabeçalho é de 0x20 bytes e usa campos uint32 a partir do
	// offset 16. Quando a versão do jogo é FFX-2, delegamos ao parser V2 (que lê
	// offset := 16 e uint32), idêntico a ParseDataListWithIlistV2.
	if common.GetGameVersionString() == "ffx2" {
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
		obj := creator(objData, stringBytes, individualLength, languageCode)
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
func ReadDataListWithIlistV2(filename string, languageCode string, creator func([]byte, []byte, int, string) datastore.IGlobalLocalizedTextObject) components.IList[datastore.IGlobalLocalizedTextObject] {
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
// ParseDataListWithIlistV2 faz o parse do formato V2 com fallbacks estáticos
func ParseDataListWithIlistV2(data []byte, languageCode string, creator func([]byte, []byte, int, string) datastore.IGlobalLocalizedTextObject) components.IList[datastore.IGlobalLocalizedTextObject] {
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

    for i := 0; i <= count; i++ {
        from := i * individualLength
        to := (i + 1) * individualLength

        if to > len(dataBytes) {
            break
        }

        objData := dataBytes[from:to]
        obj := creator(objData, stringBytes, individualLength, languageCode)
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
