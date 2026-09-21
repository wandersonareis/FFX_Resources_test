package common

import "testing"

func TestCharsetVersion(t *testing.T) {
	if got := CharsetVersion(GameVersionFFX); got != GameVersionFFX {
		t.Fatalf("ffx identity: got %v", got)
	}
	if got := CharsetVersion(GameVersionFFX2); got != GameVersionFFX2 {
		t.Fatalf("ffx2 identity: got %v", got)
	}
	if got := CharsetVersion(GameVersionLastMiss); got != GameVersionFFX2 {
		t.Fatalf("lastmiss must resolve ffx2: got %v", got)
	}
}

func TestCharsetVersionPanicsOnUnknown(t *testing.T) {
	for name, gv := range map[string]GameVersion{"zero": {}, "invalid": newGameVersion("nope")} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic for %s version", name)
				}
			}()
			_ = CharsetVersion(gv)
		}()
	}
}
