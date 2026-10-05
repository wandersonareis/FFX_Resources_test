package objectsfile

import (
	"bytes"
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

	// source escolhe a árvore de leitura (SourcePreferred = mods-first).
	// SourceData monta o ORIGINAL de data/ sem cair em mods/.
	source common.FileSource
}

// NewObjectBinaryFileStore constrói um ObjectBinaryFileStore.
func NewObjectBinaryFileStore(patternPath string, creator CreatorFunc, languageCode string, version common.GameVersion) *ObjectBinaryFileStore {
	return &ObjectBinaryFileStore{
		Header:       NewBinaryHeader(version),
		patternPath:  patternPath,
		languageCode: languageCode,
		creator:      creator,
		Version:      version,
		source:       common.SourcePreferred,
	}
}

// WithFileSource aponta a leitura para uma árvore específica (deve vir antes
// de LoadFromBinary).
func (b *ObjectBinaryFileStore) WithFileSource(src common.FileSource) *ObjectBinaryFileStore {
	b.source = src
	return b
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
	// Fontes de .vbf não têm atalho em disco: o conteúdo vem do overlay em
	// memória, direto pelo accessor.
	if base := interactionGameFilesDirStore(); base != "" && !b.source.IsVbfSource() {
		if data, err := b.readFileFromBase(base); err == nil {
			return data, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	fileAccessor, err := common.NewFileAccessorFrom(b.resolveFilePath(), b.source)
	if err != nil {
		common.LogVerbose("Error accessing file: %v", err)
		return nil, errors.New("failed to access file")
	}

	if !fileAccessor.Exists {
		common.LogVerbose("File does not exist (%s): %s", b.source, b.patternPath)
		return nil, errors.New("file does not exist")
	}

	data, err := fileAccessor.ReadBytes()
	if err != nil {
		common.LogVerbose("Error reading file: %v", err)
		return nil, errors.New("failed to read file")
	}
	return data, nil
}

// readFileFromBase lê dentro da raiz (target directory) respeitando a fonte:
// SourceData lê SÓ data/, SourceMods lê SÓ mods/ e SourcePreferred faz o
// fluxo mods-first habitual (mods/ quando EnableMods, senão data/).
func (b *ObjectBinaryFileStore) readFileFromBase(base string) ([]byte, error) {
	rel := b.resolveFilePath()
	readAt := func(p string) ([]byte, error) {
		data, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				common.LogVerbose("File does not exist (%s): %s", b.source, b.patternPath)
				return nil, err
			}
			common.LogVerbose("Error reading file: %v", err)
			return nil, errors.New("failed to read file")
		}
		return data, nil
	}

	switch b.source {
	case common.SourceData:
		return readAt(filepath.Join(base, rel))
	case common.SourceMods:
		return readAt(filepath.Join(base, common.ModsFolder, rel))
	case common.SourceVbf, common.SourceVbfPreferred:
		// Fonte de .vbf não tem atalho em disco: deixa o accessor resolver.
		return nil, os.ErrNotExist
	}

	if !common.AreModsEnabled() {
		return readAt(filepath.Join(base, rel))
	}

	if data, err := os.ReadFile(filepath.Join(base, common.ModsFolder, rel)); err == nil {
		return data, nil
	}
	return readAt(filepath.Join(base, rel))
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

	PopulateDataObjectLocalizationsFrom(b.patternPath, b.Objects, b.creator, b.Version, b.source)
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

func (b *ObjectBinaryFileStore) SaveToBinary(filePath string) error {
	return SaveBinaryFileStore(b, filePath)
}

func (b *ObjectBinaryFileStore) GetObjects() components.IList[datastore.IGlobalLocalizedTextObject] {
	return b.Objects
}
