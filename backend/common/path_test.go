package common

import (
	"path/filepath"
	"testing"
)

// TestMacroBinaryPathByVersion valida o caminho do binário de macro por
// versão explícita: ffx/eternalcalm na árvore ffx, ffx2/lastmiss na ffx2.
func TestMacroBinaryPath(t *testing.T) {
	prevRoot := GameFilesRoot
	GameFilesRoot = filepath.Join("tmp", "gamefiles")
	defer func() { GameFilesRoot = prevRoot }()

	wantSuffix := filepath.Join("mods", "ffx_ps2", "ffx", "master", "new_uspc", "menu", "macrodic.dcp")
	if got := MacroBinaryPath(GameVersionFFX, "us"); filepath.Base(got) == "" || !endsWith(got, wantSuffix) {
		t.Fatalf("ffx/us: got %q, want suffix %q", got, wantSuffix)
	}
	wantSuffix2 := filepath.Join("mods", "ffx_ps2", "ffx2", "master", "new_uspc", "menu", "macrodic.dcp")
	if got := MacroBinaryPath(GameVersionFFX2, "us"); !endsWith(got, wantSuffix2) {
		t.Fatalf("ffx2/us: got %q, want suffix %q", got, wantSuffix2)
	}
	// lastmiss divide a árvore com o ffx2.
	if got := MacroBinaryPath(GameVersionLastMiss, "us"); !endsWith(got, wantSuffix2) {
		t.Fatalf("lastmiss/us: got %q, want suffix %q", got, wantSuffix2)
	}
	// eternalcalm divide a árvore com o ffx.
	if got := MacroBinaryPath(GameVersionEternalCalm, "us"); !endsWith(got, wantSuffix) {
		t.Fatalf("eternalcalm/us: got %q, want suffix %q", got, wantSuffix)
	}
}

// TestVersionPathName valida a normalização do diretório de versão:
// expansões resolvem para o jogo-pai (eternalcalm → ffx, lastmiss → ffx2).
func TestVersionPathName(t *testing.T) {
	cases := map[GameVersion]string{
		GameVersionFFX:         "ffx",
		GameVersionEternalCalm: "ffx",
		GameVersionFFX2:        "ffx2",
		GameVersionLastMiss:    "ffx2",
	}
	for gv, want := range cases {
		if got := VersionPathName(gv); got != want {
			t.Fatalf("VersionPathName(%v): got %q, want %q", gv, got, want)
		}
	}
}

func endsWith(path, suffix string) bool {
	return len(path) >= len(suffix) && filepath.ToSlash(path[len(path)-len(suffix):]) == filepath.ToSlash(suffix)
}
