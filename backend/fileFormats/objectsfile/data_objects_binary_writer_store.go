package objectsfile

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
	"fmt"
	"path/filepath"
)

// Cópias Store do writer binário. Espelham data_objects_binary_writer.go
// (congelado); operam em ObjectBinaryFileStore / KeyedStringFileStore.

// SaveBinaryFileStore persiste os objetos do store de volta ao binário.
func SaveBinaryFileStore(b *ObjectBinaryFileStore, filePath string) error {
	var lastBuf []byte

	for localizationKey := range common.SupportedLanguages {
		if localizationKey != "us" {
			continue
		}

		buf, err := encodeBinaryLanguageStore(b, localizationKey)
		if err != nil {
			return err
		}

		if err := writeBinaryLocalizedFileStore(b, b.Version, localizationKey, buf.Bytes()); err != nil {
			return err
		}
		lastBuf = buf.Bytes()
	}

	return common.WriteBytesToFile(filePath, lastBuf)
}

func encodeBinaryLanguageStore(b *ObjectBinaryFileStore, localizationKey string) (*bytes.Buffer, error) {
	keyedStrings := collectBinaryKeyedStringsStore(b, localizationKey)
	charset := ffxencoding.GetCharsetForLanguage(localizationKey)
	stringBytes := RebuildKeyedStrings(keyedStrings, charset, b.Version)

	buf := bytes.NewBuffer(make([]byte, 0, b.Header.GetDataLength()+len(stringBytes)+0x20))

	if err := b.Header.Write(buf); err != nil {
		return nil, fmt.Errorf("error writing header: %w", err)
	}

	var writeErr error
	b.Objects.RangeIndex(func(i int, obj datastore.IGlobalLocalizedTextObject) {
		if obj != nil && writeErr == nil {
			chunkBytes, err := obj.ToBytes(localizationKey)
			if err != nil {
				writeErr = fmt.Errorf("error converting object %d to bytes: %w", i, err)
				return
			}
			buf.Write(chunkBytes)
		}
	})
	if writeErr != nil {
		return nil, writeErr
	}

	buf.Write(stringBytes)
	return buf, nil
}

func collectBinaryKeyedStringsStore(b *ObjectBinaryFileStore, localizationKey string) []datastore.IGlobalKeyedString {
	var all []datastore.IGlobalKeyedString
	b.Objects.RangeIndex(func(_ int, obj datastore.IGlobalLocalizedTextObject) {
		for _, ks := range obj.GetLocalizedKeyedStrings(localizationKey) {
			if ks != nil {
				all = append(all, ks)
			} else {
				common.LogVerbose("Keyed string is nil for object at index %d", obj.GetName(common.DefaultLocalization))
			}
		}
	})
	return all
}

func writeBinaryLocalizedFileStore(b *ObjectBinaryFileStore, version common.GameVersion, localizationKey, data []byte) error {
	// A cópia na árvore de mods espelha a localização canônica do arquivo
	// (patternPath) — nunca o filePath do chamador, que pode ser absoluto
	// (ex.: testes gravando em temp dir) e não deve ser embutido no join.
	localePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, common.GetLocalizationRootForVersion(version, localizationKey), b.patternPath)
	localePath = filepath.FromSlash(localePath)

	dir := filepath.Dir(localePath)
	if err := common.EnsurePathExists(dir); err != nil {
		return fmt.Errorf("error when creating directory %s: %w", dir, err)
	}

	if err := common.WriteBytesToFile(localePath, data); err != nil {
		return fmt.Errorf("error when writing file %s: %w", localePath, err)
	}

	common.LogVerbose("Wrote localized data to %s (%d bytes)", localePath, len(data))
	return nil
}

// writeStringSegmentsAtStore escreve segmentos de 4 bytes nos offsets absolutos
// dados. Espelha writeStringSegmentsAt.
func writeStringSegmentsAtStore(data []byte, offsets []int, languageCode string, segments ...datastore.IGlobalLocalizedKeyedStringObject) error {
	for i, seg := range segments {
		content := seg.GetLocalizedContent(languageCode)
		if content == nil {
			continue
		}
		if err := models.WriteSegmentAt(data, offsets[i], getSegment(content)); err != nil {
			return fmt.Errorf("writing segment %d: %w", i, err)
		}
	}
	return nil
}
