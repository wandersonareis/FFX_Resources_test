package macrodic

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"ffxresources/backend/common"
	"os"
	"testing"
)

const (
	testFFXDcpPath  = "../../../testData/FFX/binary/ffx_ps2/ffx/master/new_uspc/menu/macrodic.dcp"
	testFFX2DcpPath = "../../../testData/FFX-2/binary/ffx_ps2/ffx2/master/new_uspc/menu/macrodic.dcp"
)

func mustLoadDcp(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("macrodic file not available: %v", err)
	}
	return raw
}

func hashHex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func roundtripContainers(t *testing.T, raw []byte, loc string, version common.GameVersion) map[string]*MacroDictionaryBinaryFile {
	t.Helper()
	c, err := NewMacroDictionaryBinaryFileFromBytes(raw, loc, version)
	if err != nil {
		t.Fatalf("parse original: %v", err)
	}
	js, err := MarshalToJson(ExportToJson(map[string]*MacroDictionaryBinaryFile{loc: c}))
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	imp, err := UnmarshalJson(js)
	if err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	back, err := ImportFromJson(imp, version)
	if err != nil {
		t.Fatalf("import JSON: %v", err)
	}
	return back
}

// TestUSReimportHashMatch validates that a full JSON roundtrip
// (binary -> merged JSON -> binary) reproduces the original us bytes exactly.
func TestUSReimportHashMatch(t *testing.T) {
	raw := mustLoadDcp(t, testFFXDcpPath)
	want := hashHex(raw)

	back := roundtripContainers(t, raw, "us", common.GameVersionFFX)
	rc, ok := back["us"]
	if !ok || len(rc.Bytes) == 0 {
		t.Fatalf("import did not rebuild us container")
	}

	if got := hashHex(rc.Bytes); got != want {
		t.Fatalf("reimported hash mismatch:\n got=%s\nwant=%s", got, want)
	}
	t.Logf("us reimport hash match OK (%d bytes)", len(raw))
}

// TestFFX2ReimportTexts validates the FFX-2 us file roundtrips through the
// merged JSON with all texts intact. No hash match is required: the original
// FFX-2 builder packs strings with suffix sharing, so a from-zero rebuild
// recalculates different (but equivalent) offsets.
func TestFFX2ReimportTexts(t *testing.T) {
	raw := mustLoadDcp(t, testFFX2DcpPath)

	orig, err := NewMacroDictionaryBinaryFileFromBytes(raw, "us", common.GameVersionFFX2)
	if err != nil {
		t.Fatalf("parse original: %v", err)
	}
	back := roundtripContainers(t, raw, "us", common.GameVersionFFX2)
	rc, ok := back["us"]
	if !ok || len(rc.Bytes) == 0 {
		t.Fatalf("import did not rebuild us container")
	}
	if _, err := NewMacroDictionaryBinaryFileFromBytes(rc.Bytes, "us", common.GameVersionFFX2); err != nil {
		t.Fatalf("rebuilt binary does not reparse: %v", err)
	}

	for i := range orig.ChunkOffsets {
		of, err := orig.FileAt(i)
		if err != nil {
			t.Fatalf("original FileAt(%d): %v", i, err)
		}
		rf, err := rc.FileAt(i)
		if err != nil {
			t.Fatalf("rebuilt FileAt(%d): %v", i, err)
		}
		if len(rf.Segments) != len(of.Segments) {
			t.Fatalf("chunk %d segment count: rebuilt=%d original=%d", i, len(rf.Segments), len(of.Segments))
		}
		for j := range of.Segments {
			o, r := of.Segments[j], rf.Segments[j]
			if (o == nil) != (r == nil) {
				t.Fatalf("chunk %d string %d nil mismatch", i, j)
			}
			if o == nil {
				continue
			}
			if !bytes.Equal(r.NameBytes, o.NameBytes) {
				t.Fatalf("chunk %d string %d name differs after roundtrip", i, j)
			}
			// Documented fallback: an originally empty SimplifiedName with a
			// distinct pointer rebuilds sharing the Name pointer.
			wantSimp := o.SimplifiedNameBytes
			if len(wantSimp) == 0 && o.SimplifiedNameOffset != o.NameOffset {
				wantSimp = o.NameBytes
			}
			if !bytes.Equal(r.SimplifiedNameBytes, wantSimp) {
				t.Fatalf("chunk %d string %d simplified differs after roundtrip", i, j)
			}
		}
	}
	t.Logf("FFX-2 us reimport texts OK (%d chunks)", len(orig.ChunkOffsets))
}

// TestOtherLocalizationsReimport validates that every other available language
// roundtrips through the merged JSON (parse -> export -> import -> rebuild ->
// reparse). No hash match is required here, only successful reimport.
func TestOtherLocalizationsReimport(t *testing.T) {
	raw := mustLoadDcp(t, testFFXDcpPath)

	otherLocs := []string{"de", "fr", "it", "sp", "jp", "ch", "kr"}
	containers := map[string]*MacroDictionaryBinaryFile{}
	for _, loc := range append([]string{"us"}, otherLocs...) {
		c, err := NewMacroDictionaryBinaryFileFromBytes(raw, loc, common.GameVersionFFX)
		if err != nil {
			t.Fatalf("parse as %s: %v", loc, err)
		}
		containers[loc] = c
	}

	exp := ExportToJson(containers)
	if len(exp.Chunks) == 0 {
		t.Fatalf("empty merged export")
	}

	// every localization must show up in at least one entry
	seen := make(map[string]bool)
	for _, ch := range exp.Chunks {
		for _, s := range ch.Strings {
			for loc := range s.Name {
				seen[loc] = true
			}
			for loc := range s.SimplifiedName {
				seen[loc] = true
			}
		}
	}
	for loc := range containers {
		if !seen[loc] {
			t.Errorf("localization %s missing from merged export", loc)
		}
	}

	js, err := MarshalToJson(exp)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	imp, err := UnmarshalJson(js)
	if err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	back, err := ImportFromJson(imp, common.GameVersionFFX)
	if err != nil {
		t.Fatalf("import JSON: %v", err)
	}

	for _, loc := range append([]string{"us"}, otherLocs...) {
		rc, ok := back[loc]
		if !ok || len(rc.Bytes) == 0 {
			t.Errorf("localization %s was not reimported", loc)
			continue
		}
		if _, err := NewMacroDictionaryBinaryFileFromBytes(rc.Bytes, loc, common.GameVersionFFX); err != nil {
			t.Errorf("rebuilt %s does not reparse: %v", loc, err)
		}
	}
}
