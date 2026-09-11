package objectsfile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/interactions"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
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
	SignatureA       uint32
	SignatureB       uint32
	MinIndex         uint16
	MaxIndex         uint16
	IndividualLength uint16
	TotalLength      uint16
	Padding          uint32
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
	return int(h.TotalLength + uint16(h.Padding))
}

// --- VERSÃO 2 ---
type BinaryHeaderV2 struct {
	SignatureA       uint32
	SignatureB       uint32
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

func NewBinaryHeader(version int) IBinaryHeader {
	if version == 2 {
		return &BinaryHeaderV2{}
	}
	return &BinaryHeaderV1{}
}

type CreatorFunc func(chunkBytes []byte, stringBytes []byte, headerLength int, languageCode string) (datastore.IGlobalLocalizedTextObject, error)

type BinaryFile struct {
	Header       IBinaryHeader
	Objects      components.IList[datastore.IGlobalLocalizedTextObject]
	StringBytes  []byte
	creator      CreatorFunc
	languageCode string
	patternPath  string
	Version      int
}

func NewBinaryFile(patternPath string, creator CreatorFunc, languageCode string, version int) *BinaryFile {
	return &BinaryFile{
		Header:       NewBinaryHeader(version),
		patternPath:  patternPath,
		languageCode: languageCode,
		creator:      creator,
		Version:      version,
	}
}

func (b *BinaryFile) resolveFilePath() string {
	return filepath.Join(common.GetLocalizationRoot(b.languageCode), b.patternPath)
}

func interactionGameFilesDir() string {
	svc := interactions.NewInteractionService()
	if svc == nil || svc.GameLocation == nil {
		return ""
	}
	return svc.GameLocation.GetTargetDirectory()
}

func (b *BinaryFile) readFile() ([]byte, error) {
	if base := interactionGameFilesDir(); base != "" {
		if data, err := b.readFileFromBase(base); err == nil {
			return data, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	fileAccessor, err := common.NewFileAccessor(b.resolveFilePath())
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

func (b *BinaryFile) readFileFromBase(base string) ([]byte, error) {
	rel := b.resolveFilePath()

	if !common.AreModsEnabled() {
		data, err := os.ReadFile(filepath.Join(base, rel))
		if err != nil {
			if os.IsNotExist(err) {
				common.LogVerbose("File does not exist: %s", b.patternPath)
				return nil, err
			}
			common.LogVerbose("Error reading file: %v", err)
			return nil, errors.New("failed to read file")
		}
		return data, nil
	}

	if data, err := os.ReadFile(filepath.Join(base, common.ModsFolder, rel)); err == nil {
		return data, nil
	}

	data, err := os.ReadFile(filepath.Join(base, rel))
	if err != nil {
		if os.IsNotExist(err) {
			common.LogVerbose("File does not exist: %s", b.patternPath)
			return nil, err
		}
		common.LogVerbose("Error reading file: %v", err)
		return nil, errors.New("failed to read file")
	}
	return data, nil
}

func (b *BinaryFile) LoadFromBinary() error {
	data, err := b.readFile()
	if err != nil {
		return err
	}

	reader := bytes.NewReader(data)

	if err := b.readHeader(reader); err != nil {
		return err
	}

	dataBytes, err := b.readChunks(reader)
	if err != nil {
		return err
	}

	if err := b.readStrings(reader); err != nil {
		return err
	}

	b.buildObjects(dataBytes)

	PopulateDataObjectLocalizationsWithIlist(b.patternPath, b.Objects, b.creator, b.Version)
	return nil
}

func (b *BinaryFile) readHeader(r *bytes.Reader) error {
	if err := b.Header.Read(r); err != nil {
		return fmt.Errorf("error reading header: %w", err)
	}
	return nil
}

func (b *BinaryFile) readChunks(r *bytes.Reader) ([]byte, error) {
	dataBytes := make([]byte, b.Header.GetTotalLength())
	if _, err := io.ReadFull(r, dataBytes); err != nil {
		return nil, fmt.Errorf("error reading chunks: %w", err)
	}
	return dataBytes, nil
}

func (b *BinaryFile) readStrings(r *bytes.Reader) error {
	b.StringBytes = make([]byte, r.Len())
	if _, err := io.ReadFull(r, b.StringBytes); err != nil {
		return fmt.Errorf("error reading strings: %w", err)
	}
	return nil
}

func (b *BinaryFile) buildObjects(dataBytes []byte) {
	count := b.Header.GetMaxIndex() - b.Header.GetMinIndex()
	b.Objects = components.NewList[datastore.IGlobalLocalizedTextObject](count + 1)

	individualLength := b.Header.GetIndividualLength()
	minIndex := b.Header.GetMinIndex()

	for i := 0; i <= count; i++ {
		from := i * individualLength
		to := (i + 1) * individualLength
		if to > len(dataBytes) {
			break
		}

		chunk := slices.Clone(dataBytes[from:to])
		obj, err := b.creator(chunk, b.StringBytes, individualLength, b.languageCode)
		if err != nil {
			common.LogVerbose("Error creating object at index %d: %v", i+minIndex, err)
			continue
		}
		if obj == nil {
			common.LogVerbose("Skipping invalid V2 object at index %d", i+minIndex)
			continue
		}
		b.Objects.Add(obj)
	}
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

func (b *BinaryFile) SaveToBinary(filePath string) error {
	return SaveBinaryFile(b, filePath)
}

func (b *BinaryFile) GetObjects() components.IList[datastore.IGlobalLocalizedTextObject] {
	return b.Objects
}
