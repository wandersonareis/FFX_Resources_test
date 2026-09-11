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

// SaveBinaryFile persists a ObjectBinaryFile's objects back to binary form.
// It mirrors the lifecycle used in examples/main.go: objects already loaded
// in memory are re-encoded for the default (us) localization, written to the
// localized mods tree and finally to filePath itself.
func SaveBinaryFile(b *ObjectBinaryFile, filePath string) error {
	var lastBuf []byte

	for localizationKey := range common.SupportedLanguages {
		if localizationKey != "us" {
			continue // Skip non-US localizations for now
		}

		buf, err := encodeBinaryLanguage(b, localizationKey)
		if err != nil {
			return err
		}

		if err := writeBinaryLocalizedFile(b, localizationKey, filePath, buf.Bytes()); err != nil {
			return err
		}
		lastBuf = buf.Bytes()
	}

	return common.WriteBytesToFile(filePath, lastBuf)
}

func encodeBinaryLanguage(b *ObjectBinaryFile, localizationKey string) (*bytes.Buffer, error) {
	keyedStrings := collectBinaryKeyedStrings(b, localizationKey)
	charset := ffxencoding.GetCharsetForLanguage(localizationKey)
	stringBytes := RebuildKeyedStrings(keyedStrings, charset, models.GameVersion(b.Version))

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

func collectBinaryKeyedStrings(b *ObjectBinaryFile, localizationKey string) []datastore.IGlobalKeyedString {
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

func writeBinaryLocalizedFile(b *ObjectBinaryFile, localizationKey, filePath string, data []byte) error {
	_ = b
	localePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, common.GetLocalizationRoot(localizationKey), filePath)
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

// writeStringSegments writes a contiguous sequence of 4-byte string reference segments
// into a pre-allocated byte slice, populating each segment from the corresponding
// IGlobalLocalizedKeyedStringObject's localized content.
//
// Each segment is exactly 4 bytes: 2 bytes for the string table offset (uint16 LE)
// and 2 bytes for the string key (uint16 LE). Segments are written sequentially
// starting at the given offset, with each subsequent segment at start+i*4. The caller
// must ensure that `data` is large enough to hold all segments.
//
// This function is the write-side counterpart of readStringSegments. It is intended
// exclusively for headers where ALL string segments are stored contiguously, with no
// gaps or non-string bytes between them. Examples:
//   - CommandTextObject: Name → SimplifiedName → Description → SimplifiedDescription
//   - NameOnlyTextObject: Name → SimplifiedName
//   - CommandTextObjectV2: Name → Description
//
// It must NOT be used for segments that are written at arbitrary or non-sequential
// positions within the binary chunk (e.g., JobTextObject where the Effect
// segment is at a separate position offset). For those cases, use direct
// models.WriteSegmentAt instead.
//
// Segments with nil localized content are skipped, preserving the original bytes
// in the data slice. The function stops and returns an error on the first write
// failure without processing remaining segments.
//
// Parameters:
//   - data: The pre-allocated byte slice to write into.
//   - start: The byte offset where the first segment is written.
//   - languageCode: The localization language code (e.g., "us", "jp", "de") used to
//     resolve each segment's localized content via GetLocalizedContent.
//   - segments: One or more IGlobalLocalizedKeyedStringObject instances to serialize,
//     listed in the exact order they should appear in the binary output.
//
// Returns: nil on success, or a wrapped fmt.Errorf indicating which segment index
// failed to write (e.g., "writing segment 2: <underlying error>").
func writeStringSegments(data []byte, start int, languageCode string, segments ...datastore.IGlobalLocalizedKeyedStringObject) error {
	for i, seg := range segments {
		content := seg.GetLocalizedContent(languageCode)
		if content == nil {
			continue
		}
		if err := models.WriteSegmentAt(data, start+i*4, getSegment(content)); err != nil {
			return fmt.Errorf("writing segment %d: %w", i, err)
		}
	}
	return nil
}