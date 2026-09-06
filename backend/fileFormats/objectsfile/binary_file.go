package objectsfile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// IBinaryHeader define o contrato para leitura e escrita do cabeçalho
type IBinaryHeader interface {
	Read(r *bytes.Reader) error
	Write(w *bytes.Buffer) error
	GetMinIndex() int
	GetMaxIndex() int
	GetIndividualLength() int
	GetTotalLength() int
	GetDataLength() int
}

// --- VERSÃO 1 ---
type BinaryHeaderV1 struct {
	Signature        [8]byte
	MinIndex         uint16
	MaxIndex         uint16
	IndividualLength uint16
	TotalLength      uint16
	Unknown1         uint32 // 4 bytes de padding
}

func (h *BinaryHeaderV1) Read(r *bytes.Reader) error {
	return binary.Read(r, binary.LittleEndian, h)
}
func (h *BinaryHeaderV1) Write(w *bytes.Buffer) error {
	return binary.Write(w, binary.LittleEndian, h)
}
func (h *BinaryHeaderV1) GetMinIndex() int         { return int(h.MinIndex) }
func (h *BinaryHeaderV1) GetMaxIndex() int         { return int(h.MaxIndex) }
func (h *BinaryHeaderV1) GetIndividualLength() int { return int(h.IndividualLength) }
func (h *BinaryHeaderV1) GetTotalLength() int      { return int(h.TotalLength) }
func (h *BinaryHeaderV1) GetDataLength() int {
	return int(h.TotalLength + uint16(h.Unknown1))
}

// --- VERSÃO 2 ---
type BinaryHeaderV2 struct {
	Signature        [4]byte
	Unknown1         uint32
	Unknown2         uint32
	MinIndex         uint32
	MaxIndex         uint32
	IndividualLength uint32
	TotalLength      uint32
	HeaderSize       uint32
}

func (h *BinaryHeaderV2) Read(r *bytes.Reader) error {
	return binary.Read(r, binary.LittleEndian, h)
}
func (h *BinaryHeaderV2) Write(w *bytes.Buffer) error {
	return binary.Write(w, binary.LittleEndian, h)
}
func (h *BinaryHeaderV2) GetMinIndex() int         { return 0 } // V2 sempre começa em 0
func (h *BinaryHeaderV2) GetMaxIndex() int         { return int(h.MaxIndex) }
func (h *BinaryHeaderV2) GetIndividualLength() int { return int(h.IndividualLength) }
func (h *BinaryHeaderV2) GetTotalLength() int      { return int(h.TotalLength) }
func (h *BinaryHeaderV2) GetDataLength() int {
	return int(h.HeaderSize)
}

// NewBinaryHeader instância o header correto sem poluir o resto do código
func NewBinaryHeader() IBinaryHeader {
	if common.GetGameVersionString() == "ffx2" {
		return &BinaryHeaderV2{}
	}
	return &BinaryHeaderV1{}
}

// CreatorFunc é a função que sabe instanciar o chunk correto (NameDesc, Plate, etc)
type CreatorFunc func(chunkBytes []byte, stringBytes []byte, headerLength int, languageCode string) datastore.IGlobalLocalizedTextObject

// JsonExporterFunc exporta objetos para JSON
type JsonExporterFunc func(objects components.IList[datastore.IGlobalLocalizedTextObject], fileName string) error

// JsonImporterFunc importa objetos de JSON
type JsonImporterFunc func(fileName string, objectsList components.IList[datastore.IGlobalLocalizedTextObject]) error

var _ datastore.IBinaryFile = (*BinaryFile)(nil)

// BinaryFile orquestra todo o ciclo de vida de um arquivo binário de localização
type BinaryFile struct {
	Header       IBinaryHeader
	Objects      components.IList[datastore.IGlobalLocalizedTextObject]
	StringBytes  []byte
	creator      CreatorFunc
	languageCode string
	patternPath  string
	relativePath string
}

// NewBinaryFile cria o orquestrador. Você passa a função que cria o objeto correto.
func NewBinaryFile(patternPath string, creator CreatorFunc, languageCode string) *BinaryFile {
	return &BinaryFile{
		Header:       NewBinaryHeader(),
		patternPath:  patternPath,
		relativePath: filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath),
		languageCode: languageCode,
		creator:      creator,
	}
}

func (b *BinaryFile) fileAcessor() ([]byte, error) {
	fileAccessor, err := common.NewFileAccessor(b.relativePath)
	if err != nil {
		common.LogVerbose("Error accessing file: %v", err)
		return nil, errors.New("failed to access file")
	}

	if !fileAccessor.Exists {
		common.LogVerbose("File does not exist: %s", b.patternPath)
		return nil, errors.New("file does not exist")
	}

	data, err := os.ReadFile(fileAccessor.ResolvedPath)
	if err != nil {
		common.LogVerbose("Error reading file: %v", err)
		return nil, errors.New("failed to read file")
	}
	return data, nil
}

func (b *BinaryFile) LoadFromBinary() error {
	data, err := b.fileAcessor()
	if err != nil {
		return err
	}

	reader := bytes.NewReader(data)

	if err := b.Header.Read(reader); err != nil {
		return fmt.Errorf("error reading header: %w", err)
	}

	dataBytes := make([]byte, b.Header.GetTotalLength())
	if _, err := io.ReadFull(reader, dataBytes); err != nil {
		return fmt.Errorf("error reading chunks: %w", err)
	}

	b.StringBytes = make([]byte, reader.Len())
	io.ReadFull(reader, b.StringBytes)

	count := b.Header.GetMaxIndex() - b.Header.GetMinIndex()
	b.Objects = components.NewList[datastore.IGlobalLocalizedTextObject](count + 1)

	individualLength := b.Header.GetIndividualLength()
	for i := 0; i <= count; i++ {
		from := i * individualLength
		to := (i + 1) * individualLength
		if to > len(dataBytes) {
			break
		}

		chunk := bytes.Clone(dataBytes[from:to])
		// Passa individualLength como headerLength para o objeto (chunk)
		obj := b.creator(chunk, b.StringBytes, individualLength, b.languageCode)
		if obj == nil {
			common.LogVerbose("Skipping invalid V2 object at index %d", i+b.Header.GetMinIndex())
			continue
		}

		b.Objects.Add(obj)
	}

	PopulateDataObjectLocalizationsWithIlist(b.patternPath, b.Objects, b.creator)

	return nil
}

func (b *BinaryFile) ExportToJson(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("json file not configured")
	}
	return ExportToJSON(b.Objects, filePath)
}

func (b *BinaryFile) ImportFromJson(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("json file not configured")
	}
	return ImportFromJson(filePath, b.Objects)
}

// SaveToBinary salva de volta para o binário no formato V2
func (b *BinaryFile) SaveToBinary(filePath string) error {
	var buf bytes.Buffer

	for localizationKey := range common.SupportedLanguages {
		if localizationKey != "us" {
			continue // Skip non-US localizations for now
		}
		localizationRoot := common.GetLocalizationRoot(localizationKey)
		localePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, localizationRoot, filePath)
		localePath = filepath.FromSlash(localePath)

		var allKeyedStrings []datastore.IGlobalKeyedString

		b.Objects.RangeIndex(func(i int, obj datastore.IGlobalLocalizedTextObject) {
			keyedStrings := obj.GetLocalizedKeyedStrings(localizationKey)
			for _, ks := range keyedStrings {
				if ks != nil {
					allKeyedStrings = append(allKeyedStrings, ks)
				} else {
					common.LogVerbose("Keyed string is nil for object at index %d", obj.GetName(common.DefaultLocalization))
				}
			}
		})
		/* for _, obj := range b.Objects.Items() {
			keyedStrings := obj.GetLocalizedKeyedStrings(localizationKey)
			for _, ks := range keyedStrings {
				if ks != nil {
					allKeyedStrings = append(allKeyedStrings, ks)
				} else {
					// TODO: delete this
					common.LogVerbose("Keyed string is nil for object at index %d", obj.GetName(common.DefaultLocalization))
				}
			}
		} */

		charset := ffxencoding.GetCharsetForLanguage(localizationKey)
		b.StringBytes = RebuildKeyedStrings(allKeyedStrings, charset)

		// 1. Escreve o Header
		if err := b.Header.Write(&buf); err != nil {
			return fmt.Errorf("error writing header: %w", err)
		}

		// 2. Escreve os Chunks (ToBytes de cada objeto)
		b.Objects.RangeIndex(func(i int, obj datastore.IGlobalLocalizedTextObject) {
			if obj != nil {
				// Se você recalculou os offsets internos do chunk, eles serão salvos aqui
				chunkBytes := obj.ToBytes(b.languageCode)
				buf.Write(chunkBytes)
			}
		})

		// 3. Escreve o Bloco de Strings
		// Se você modificou textos, este StringBytes deve ser o NOVO bloco de strings
		// recalculado. Aqui assumimos que b.StringBytes foi atualizado antes de chamar esta função.
		buf.Write(b.StringBytes)

		dir := filepath.Dir(localePath)
		if err := common.EnsurePathExists(dir); err != nil {
			return fmt.Errorf("error when creating directory %s: %w", dir, err)
		}

		if err := common.WriteBytesToFile(localePath, buf.Bytes()); err != nil {
			return fmt.Errorf("error when writing file %s: %w", localePath, err)
		}

		common.LogVerbose("Wrote localized data to %s (%d bytes)", localePath, len(buf.Bytes()))
	}

	// 4. Salva no disco
	return common.WriteBytesToFile(filePath, buf.Bytes())
}

// GetObjects permite acessar a lista de objetos
func (b *BinaryFile) GetObjects() components.IList[datastore.IGlobalLocalizedTextObject] {
	return b.Objects
}
