package macrodic

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// RebuildChunkData rebuilds one macro dictionary file from its raw segment texts,
// from zero. Layout rule (matches the original builder, verified byte-identical
// on real files): first offset = segment count * 4; each Name text is appended
// in order (never deduplicated across entries), including empty texts which get
// their own null byte; SimplifiedName shares the Name offset when its bytes are
// equal and is appended otherwise (an empty SimplifiedName therefore never
// aliases the content start, unlike the legacy contentOffset fallback).
// Nil segments rebuild as zero entries, preserving index alignment. Only the
// input slices are read — nothing is mutated.
func RebuildChunkData(segments []*MacroDictionaryTextSegment) []byte {
	count := len(segments)
	contentOffset := count * 4
	var buf bytes.Buffer
	type pair struct{ name, simplified int }
	offs := make([]pair, len(segments))
	for i, seg := range segments {
		if seg == nil {
			continue // zero entry
		}
		offs[i].name = contentOffset + buf.Len()
		buf.Write(seg.NameBytes)
		buf.WriteByte(0x00)
		if bytes.Equal(seg.SimplifiedNameBytes, seg.NameBytes) {
			offs[i].simplified = offs[i].name
		} else {
			offs[i].simplified = contentOffset + buf.Len()
			buf.Write(seg.SimplifiedNameBytes)
			buf.WriteByte(0x00)
		}
	}

	out := make([]byte, count*4+buf.Len())
	for i, o := range offs {
		binary.LittleEndian.PutUint16(out[i*4:], uint16(o.name))
		binary.LittleEndian.PutUint16(out[i*4+2:], uint16(o.simplified))
	}
	copy(out[count*4:], buf.Bytes())
	return out
}

// BuildContainerBinary assembles the full container binary from the given text
// files, recalculating every chunk offset from the rebuilt lengths: the first
// file starts at 0x40 and each following non-zero pointer is the previous offset
// plus the previous file length. Empty files keep zero slots. Only fresh slices
// are produced — existing chunks, offsets and text bytes are never altered.
func BuildContainerBinary(files []*MacroDictionaryTextFile) ([]byte, error) {
	if len(files) > MacroDictionaryChunkCount {
		return nil, fmt.Errorf("too many files to build macro dictionary: have %d, need at most %d", len(files), MacroDictionaryChunkCount)
	}

	fileBytes := make([][]byte, MacroDictionaryChunkCount)
	for i, f := range files {
		if f == nil || len(f.Segments) == 0 {
			continue // zero slot preserved
		}
		b, err := f.ToBytes()
		if err != nil {
			return nil, fmt.Errorf("rebuilding file %d: %w", i, err)
		}
		if len(b) == 0 {
			continue
		}
		fileBytes[i] = b
	}

	header := make([]byte, macroDictionaryHeaderLength)
	currentOffset := uint32(macroDictionaryHeaderLength)
	for i, b := range fileBytes {
		if len(b) == 0 {
			continue // zero slot preserved
		}
		binary.LittleEndian.PutUint32(header[i*4:], currentOffset)
		currentOffset += uint32(len(b))
	}

	containerBytes := make([]byte, 0, int(currentOffset))
	containerBytes = append(containerBytes, header...)
	for _, b := range fileBytes {
		containerBytes = append(containerBytes, b...)
	}

	return containerBytes, nil
}
