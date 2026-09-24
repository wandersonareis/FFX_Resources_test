package objectsfile

import (
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
	"fmt"
	"os"
	"path/filepath"
)

// Cópias com sufixo Store do parse binário usado pelos Read*.
// Lógica idêntica aos originais em data_objects_binary_reader.go
// (congelado como oráculo); pertencem ao store novo.

// PopulateDataObjectLocalizationsWithIlistStore popula localizações de todos
// os idiomas suportados nos objetos já carregados (idioma default).
func PopulateDataObjectLocalizationsWithIlistStore(path string, objects components.IList[datastore.IGlobalLocalizedTextObject], creator func([]byte, []byte, int, string) (datastore.IGlobalLocalizedTextObject, error), version common.GameVersion) {
	if objects == nil || objects.IsEmpty() {
		return
	}

	for locKey := range common.SupportedLanguages {
		fullPath := filepath.Join(common.GetLocalizationRootForVersion(version, locKey), path)

		var localizationData components.IList[datastore.IGlobalLocalizedTextObject]
		if version == common.GameVersionFFX {
			localizationData = ReadDataListWithIlistStore(fullPath, locKey, creator, version)
		} else {
			localizationData = ReadDataListWithIlistV2Store(fullPath, locKey, creator)
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

// ReadDataListWithIlistStore lê o arquivo e delega ao ParseDataListWithIlistStore.
func ReadDataListWithIlistStore(filename string, languageCode string, creator func([]byte, []byte, int, string) (datastore.IGlobalLocalizedTextObject, error), version common.GameVersion) components.IList[datastore.IGlobalLocalizedTextObject] {
	fileAccessor, err := common.NewFileAccessor(filename)
	if err != nil {
		common.LogVerbose("Error accessing file: %v", err)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	if !fileAccessor.Exists {
		common.LogVerbose("File does not exist: %s", filename)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	data, err := fileAccessor.ReadBytes()
	if err != nil {
		common.LogVerbose("Error reading file: %v", err)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	return ParseDataListWithIlistStore(data, languageCode, creator, version)
}

// ParseDataListWithIlistStore faz o parse do formato V1 (16 bytes de header útil).
func ParseDataListWithIlistStore(data []byte, languageCode string, creator func([]byte, []byte, int, string) (datastore.IGlobalLocalizedTextObject, error), version common.GameVersion) components.IList[datastore.IGlobalLocalizedTextObject] {
	if version != common.GameVersionFFX {
		return ParseDataListWithIlistV2Store(data, languageCode, creator)
	}

	if len(data) < 16 {
		common.LogVerbose("Data too small for valid binary format")
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	offset := 8

	minIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	maxIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	individualLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	totalLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	offset += 4

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

// ReadDataListWithIlistV2Store é a contraparte V2 de ReadDataListWithIlistStore.
func ReadDataListWithIlistV2Store(filename string, languageCode string, creator func([]byte, []byte, int, string) (datastore.IGlobalLocalizedTextObject, error)) components.IList[datastore.IGlobalLocalizedTextObject] {
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

	return ParseDataListWithIlistV2Store(data, languageCode, creator)
}

// ParseDataListWithIlistV2Store faz o parse do formato V2 (header de 0x20 bytes).
func ParseDataListWithIlistV2Store(data []byte, languageCode string, creator func([]byte, []byte, int, string) (datastore.IGlobalLocalizedTextObject, error)) components.IList[datastore.IGlobalLocalizedTextObject] {
	if len(data) < 32 {
		common.LogVerbose("Data too small for valid binary format v2")
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	offset := 16

	maxIndex := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	minIndex := 0

	individualLength := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	totalLength := int(binary.LittleEndian.Uint32(data[offset : offset+4]))

	expectedTotalLength := (maxIndex - minIndex + 1) * individualLength
	if totalLength != expectedTotalLength {
		headerConsistent := individualLength > 0 &&
			totalLength%individualLength == 0 &&
			32+totalLength <= len(data)
		if headerConsistent && totalLength < expectedTotalLength {
			common.LogVerbose("TotalLength do cabeçalho confere com o arquivo, mantendo valor original")
		} else {
			common.LogVerbose("TotalLength divergente, aplicando fallback calculado")
			totalLength = expectedTotalLength
		}
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

// readStringSegmentsAtStore lê segmentos de 4 bytes nos offsets absolutos dados,
// populando cada IGlobalLocalizedKeyedStringObject com seu conteúdo localizado.
// Espelha readStringSegmentsAt; offsets vêm de segmentOffsetsStore.
func readStringSegmentsAtStore(data []byte, offsets []int, stringBytes []byte, languageCode string, version common.GameVersion, segments ...datastore.IGlobalLocalizedKeyedStringObject) error {
	for i, seg := range segments {
		s, err := models.ReadSegmentAt(data, offsets[i])
		if err != nil {
			return fmt.Errorf("reading segment %d: %w", i, err)
		}
		seg.ReadAndSetLocalizedContent(languageCode, stringBytes, s.Offset, s.Key, version)
	}
	return nil
}

// segmentOffsetsStore calcula onde cada segmento começa dentro do chunk.
// Cada segmento ocupa 4 bytes; gaps[i] = bytes pulados APÓS o segmento i.
// O primeiro segmento sempre começa em 0 (sem Start).
func segmentOffsetsStore(gaps []int, n int) []int {
	offsets := make([]int, n)
	pos := 0
	for i := 0; i < n; i++ {
		offsets[i] = pos
		pos += 4
		if i < len(gaps) {
			pos += gaps[i]
		}
	}
	return offsets
}
