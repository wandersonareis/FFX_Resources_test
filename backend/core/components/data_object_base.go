package components

import (
	"encoding/binary"
	"ffxresources/backend/common"
	"fmt"
	"os"
)

// Stringer interface para objetos que têm método String
type Stringer interface {
	String() string
}

// DataObject interface que combina Stringer com método ToString
type DataObject interface {
	Stringer
	ToString(localization string) string
}

// LocalizationSetter interface para objetos que podem receber localizações
type LocalizationSetter interface {
	SetLocalizations(other LocalizationSetter)
}

type NameDescriptionGetter interface {
	GetNameDescriptionTextObject() *NameDescriptionTextObject
}

// DataObjectWithLocalizations interface que combina DataObject com LocalizationSetter
type DataObjectWithLocalizations interface {
	DataObject
	LocalizationSetter
}

// DataObjectBase é a estrutura base para todos os objetos de dados que herdam NameDescriptionTextObject
type DataObjectBase[T any] struct {
	*NameDescriptionTextObject
}

// DataObjectCreator é uma função que cria um objeto T a partir dos dados
type DataObjectCreator[T any] func([]byte, []byte, string) T

// DataIndexWriter é uma função que formata um índice para impressão
type DataIndexWriter func(int) string

// NewDataObjectBase cria uma nova instância base
func NewDataObjectBase[T any](data []byte, stringBytes []byte, localization string) *DataObjectBase[T] {
	nameDescObj := NewNameDescriptionTextObject(data, stringBytes, localization)

	return &DataObjectBase[T]{
		NameDescriptionTextObject: nameDescObj,
	}
}

// GetName implementa a interface Nameable
func (d *DataObjectBase[T]) GetName(localization string) string {
	return d.Name.GetLocalizedString(localization)
}

// GetKeyedString recupera uma string localizada específica por título
func (d *DataObjectBase[T]) GetKeyedString(title string) *LocalizedKeyedStringObject {
	switch title {
	case "name":
		return d.Name
	case "description":
		return d.Description
	default:
		return nil
	}
}

// String retorna uma representação em string usando localização padrão
func (d *DataObjectBase[T]) String() string {
	return d.ToString(common.DefaultLocalization)
}

// ToString retorna uma representação em string formatada para uma localização específica
func (d *DataObjectBase[T]) ToString(languageCode string) string {
	descriptionStr := d.Description.GetLocalizedContent(languageCode).String()
	nameStr := d.GetName(languageCode)
	return fmt.Sprintf("%-22s %s", nameStr, descriptionStr)
}

// ReadDataArray lê um arquivo e retorna um array de objetos T
func ReadDataArray[T DataObject](filename string, languageCode string, creator DataObjectCreator[T]) []T {
	list := readDataList(filename, languageCode, creator)
	if list == nil {
		return nil
	}
	return list
}

// ReadDataList lê um arquivo e retorna uma slice de objetos T
func readDataList[T DataObject](filename string, languageCode string, creator DataObjectCreator[T]) []T {
	fileAccessor, err := common.NewFileAccessor(filename)
	if err != nil {
		if common.IsVerboseMode() {
			fmt.Printf("Erro ao acessar arquivo: %v\n", err)
		}
		return nil
	}

	if !fileAccessor.Exists {
		if common.IsVerboseMode() {
			fmt.Printf("Arquivo não existe: %s\n", filename)
		}
		return nil
	}

	data, err := os.ReadFile(fileAccessor.ResolvedPath)
	if err != nil {
		if common.IsVerboseMode() {
			fmt.Printf("Erro ao ler arquivo: %v\n", err)
		}
		return nil
	}

	if len(data) < 16 { // Minimum header size
		if common.IsVerboseMode() {
			fmt.Printf("Arquivo muito pequeno: %s\n", filename)
		}
		return nil
	}

	// Skip first 8 bytes
	offset := 8

	// Read min and max indices (2 bytes each, little endian)
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
		if common.IsVerboseMode() {
			fmt.Printf("Dados insuficientes no arquivo: %s\n", filename)
		}
		return nil
	}

	// Read data bytes
	dataBytes := data[offset : offset+totalLength]
	offset += totalLength

	// Read string bytes (remaining data)
	stringBytes := data[offset:]

	// Create objects slice
	objectCount := maxIndex - minIndex + 1
	objects := make([]T, 0, objectCount)

	for i := 0; i <= maxIndex-minIndex; i++ {
		from := i * individualLength
		to := (i + 1) * individualLength

		if to > len(dataBytes) {
			break
		}
		objData := dataBytes[from:to]
		obj := creator(objData, stringBytes, languageCode)
		objects = append(objects, obj)

		if common.IsVerboseMode() {
			offsetStr := fmt.Sprintf("%04X", (i*individualLength)+0x14)
			indexStr := fmt.Sprintf("%d", i+minIndex)
			fmt.Printf("%s (Offset %s) - %s\n", indexStr, offsetStr, obj.ToString(languageCode))
		}
	}

	return objects
}
