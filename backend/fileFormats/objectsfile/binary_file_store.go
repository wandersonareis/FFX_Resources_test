package objectsfile

import (
	"bytes"
	"errors"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
)

// Cópia Store de ObjectBinaryFile.
// Espelha byte-a-byte o ciclo usado pelos Read* (via newObjectBinaryFile):
// header -> chunks -> strings -> buildObjects -> Populate localizações.
// O original em binary_file.go está congelado como oráculo.

// ObjectBinaryFileStore é a implementação do store para arquivos de objetos
// com strings localizadas.
type ObjectBinaryFileStore struct {
	Header       IBinaryHeader
	Objects      components.IList[datastore.IGlobalLocalizedTextObject]
	StringBytes  []byte
	creator      CreatorFunc
	languageCode string
	patternPath  string
	Version      common.GameVersion
}

// NewObjectBinaryFileStore constrói um ObjectBinaryFileStore.
func NewObjectBinaryFileStore(patternPath string, creator CreatorFunc, languageCode string, version common.GameVersion) interfaces.IBinaryFile[datastore.IGlobalLocalizedTextObject] {
	return &ObjectBinaryFileStore{
		Header:       NewBinaryHeader(version),
		patternPath:  patternPath,
		languageCode: languageCode,
		creator:      creator,
		Version:      version,
	}
}

func (b *ObjectBinaryFileStore) resolveFilePath() string {
	return b.resolveFilePathStore()
}

// resolveFilePathStore monta o caminho relativo usando a versão do próprio
// arquivo (lastmiss resolve sob ffx2), sem depender da versão global.
func (b *ObjectBinaryFileStore) resolveFilePathStore() string {
	return filepath.Join(common.GetLocalizationRootForVersion(b.Version, b.languageCode), b.patternPath)
}

func interactionGameFilesDirStore() string {
	svc := interactions.NewInteractionService()
	if svc == nil || svc.GameLocation == nil {
		return ""
	}
	return svc.GameLocation.GetTargetDirectory()
}

func (b *ObjectBinaryFileStore) readFile() ([]byte, error) {
	if base := interactionGameFilesDirStore(); base != "" {
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

func (b *ObjectBinaryFileStore) readFileFromBase(base string) ([]byte, error) {
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

// LoadFromBinaryStore carrega header, chunks, strings e constrói os objetos.
func (b *ObjectBinaryFileStore) LoadFromBinary() error {
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

	PopulateDataObjectLocalizationsWithIlistStore(b.patternPath, b.Objects, b.creator, b.Version)
	return nil
}

func (b *ObjectBinaryFileStore) readHeader(r *bytes.Reader) error {
	if err := b.Header.Read(r); err != nil {
		return fmt.Errorf("error reading header: %w", err)
	}
	return nil
}

func (b *ObjectBinaryFileStore) readChunks(r *bytes.Reader) ([]byte, error) {
	dataBytes := make([]byte, b.Header.GetTotalLength())
	if _, err := io.ReadFull(r, dataBytes); err != nil {
		return nil, fmt.Errorf("error reading chunks: %w", err)
	}
	return dataBytes, nil
}

func (b *ObjectBinaryFileStore) readStrings(r *bytes.Reader) error {
	b.StringBytes = make([]byte, r.Len())
	if _, err := io.ReadFull(r, b.StringBytes); err != nil {
		return fmt.Errorf("error reading strings: %w", err)
	}
	return nil
}

func (b *ObjectBinaryFileStore) buildObjects(dataBytes []byte) {
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

// newObjectBinaryFileStore monta, carrega e registra os objetos em
// datastore.Commands. Espelha newObjectBinaryFile do caminho Read.
func newObjectBinaryFileStore(patternPath string, creatorFunc CreatorFunc, gameVersion common.GameVersion, formatter StringFormatterStore) *ObjectBinaryFileStore {
	binaryDataFile := &ObjectBinaryFileStore{
		Header:       NewBinaryHeader(gameVersion),
		patternPath:  patternPath,
		languageCode: common.DefaultLocalization,
		creator: func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
			obj, err := creatorFunc(cBytes, sBytes, hLen, lang)
			if err != nil {
				return nil, err
			}
			if formatter != nil {
				if kf, ok := obj.(*KeyedStringFileStore); ok {
					kf.SetFormatter(formatter)
				}
			}
			return obj, nil
		},
		Version: gameVersion,
	}

	if err := binaryDataFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading binary data: %v", err)
		return nil
	}

	if objects := binaryDataFile.GetObjects(); objects != nil && !objects.IsEmpty() {
		common.LogVerbose("Loaded %d objects with all localizations", objects.Len())
		datastore.Commands = objects
	}
	return binaryDataFile
}

func (b *ObjectBinaryFileStore) ExportToJson(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("json file not configured")
	}
	return ExportToJSON(b.Objects, filePath)
}

func (b *ObjectBinaryFileStore) ImportFromJson(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("json file not configured")
	}
	return ImportFromJson(filePath, b.Objects)
}

func (b *ObjectBinaryFileStore) SaveToBinary(filePath string) error {
	return SaveBinaryFileStore(b, filePath)
}

func (b *ObjectBinaryFileStore) GetObjects() components.IList[datastore.IGlobalLocalizedTextObject] {
	return b.Objects
}
