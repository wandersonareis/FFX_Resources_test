package objectsfile

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"fmt"
	"path/filepath"
)

// CreatorFunc é a função que sabe instanciar o chunk correto (NameDesc, Plate, etc)
type CreatorFunc func(chunkBytes []byte, stringBytes []byte, headerLength int, languageCode string) datastore.IGlobalLocalizedTextObject

// JsonExporterFunc exporta objetos para JSON
type JsonExporterFunc func(objects components.IList[datastore.IGlobalLocalizedTextObject], fileName string) error

// JsonImporterFunc importa objetos de JSON
type JsonImporterFunc func(fileName string, objectsList components.IList[datastore.IGlobalLocalizedTextObject]) error

var _ datastore.IBinaryFile = (*BinaryFile)(nil)

// BinaryFile orquestra todo o ciclo de vida de um arquivo binário de localização
type BinaryFile struct {
	RawHeader    []byte
	Objects      components.IList[datastore.IGlobalLocalizedTextObject]
	StringBytes  []byte
	HeaderLength int

	JsonExporter JsonExporterFunc
	JsonImporter JsonImporterFunc

	creator  CreatorFunc
	filePath string
}

// NewBinaryFile cria o orquestrador. Você passa a função que cria o objeto correto.
func NewBinaryFile(creator CreatorFunc, jsonExporter JsonExporterFunc, jsonImporter JsonImporterFunc) *BinaryFile {
	return &BinaryFile{
		creator:      creator,
		JsonExporter: jsonExporter,
		JsonImporter: jsonImporter,
	}
}

// LoadFromBinary faz o parse do binário no formato V2 (header 32 bytes, campos uint32)
func (b *BinaryFile) LoadFromBinary(data []byte) error {
	if len(data) < 32 {
		return fmt.Errorf("data too small for valid binary format v2")
	}

	b.RawHeader = make([]byte, 32)
	copy(b.RawHeader, data[:32])

	offset := 16

	maxIndex := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	individualLength := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	totalLength := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	expectedTotalLength := (maxIndex + 1) * individualLength
	if totalLength != expectedTotalLength {
		common.LogVerbose("TotalLength divergente, aplicando fallback calculado")
		totalLength = expectedTotalLength
	}

	b.HeaderLength = 32

	offset = 32
	if offset+totalLength > len(data) {
		return fmt.Errorf("insufficient data for specified total length")
	}

	dataBytes := data[offset : offset+totalLength]
	offset += totalLength

	b.StringBytes = data[offset:]

	count := maxIndex
	b.Objects = components.NewList[datastore.IGlobalLocalizedTextObject](count + 1)

	for i := 0; i <= count; i++ {
		from := i * individualLength
		to := (i + 1) * individualLength
		if to > len(dataBytes) {
			break
		}
		objData := dataBytes[from:to]
		obj := b.creator(objData, b.StringBytes, individualLength, common.DefaultLocalization)
		if obj != nil {
			b.Objects.Add(obj)
		}
	}

	return nil
}

// ExportToJson exporta para JSON usando a função injetada
func (b *BinaryFile) ExportToJson(filePath string) error {
	if b.JsonExporter == nil {
		return fmt.Errorf("json exporter not configured")
	}
	return b.JsonExporter(b.Objects, filePath)
}

// ImportFromJson importa do JSON (atualiza os ponteiros na memória)
func (b *BinaryFile) ImportFromJson(filePath string) error {
	if b.JsonImporter == nil {
		return fmt.Errorf("json importer not configured")
	}
	return b.JsonImporter(filePath, b.Objects)
}

// SaveToBinary salva de volta para o binário no formato V2
func (b *BinaryFile) SaveToBinary(filePath string) error {
	var fileBuffer bytes.Buffer

	fileBuffer.Write(b.RawHeader)

	objects := b.Objects.Items()

	var allKeyedStrings []datastore.IGlobalKeyedString
	for _, obj := range objects {
		keyedStrings := obj.GetLocalizedKeyedStrings(common.DefaultLocalization)
		for _, ks := range keyedStrings {
			if ks != nil {
				allKeyedStrings = append(allKeyedStrings, ks)
			}
		}
	}

	charset := components.GetCharsetForLanguage(common.DefaultLocalization)
	stringBytes := RebuildKeyedStrings(allKeyedStrings, charset)

	for _, obj := range objects {
		fileBuffer.Write(obj.ToBytes(common.DefaultLocalization))
	}

	fileBuffer.Write(stringBytes)

	dir := filepath.Dir(filePath)
	if err := common.EnsurePathExists(dir); err != nil {
		return err
	}

	return common.WriteBytesToFile(filePath, fileBuffer.Bytes())
}

// GetObjects permite acessar a lista de objetos
func (b *BinaryFile) GetObjects() components.IList[datastore.IGlobalLocalizedTextObject] {
	return b.Objects
}
