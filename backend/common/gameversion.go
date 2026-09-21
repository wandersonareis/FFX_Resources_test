package common

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GameVersion é a identidade única da versão do jogo.
// Valores válidos: "ffx", "ffx2", "lastmiss".
// O campo s é privado — só ParseGameVersionStrict e unmarshaling JSON criam instâncias válidas.
type GameVersion struct {
	s string
}

// Construtor interno para uso por constantes e unmarshaling.
func newGameVersion(s string) GameVersion {
	return GameVersion{s: s}
}

var (
	GameVersionFFX      = newGameVersion("ffx")
	GameVersionFFX2     = newGameVersion("ffx2")
	GameVersionLastMiss = newGameVersion("lastmiss")
)



// ParseGameVersionStrict é o único ponto de validação de entrada externa.
// Retorna GameVersion válido ou erro em inglês para versões desconhecidas.
func ParseGameVersionStrict(s string) (GameVersion, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ffx", "ffx-ps1", "ffx_ps1":
		return GameVersionFFX, nil
	case "ffx2", "ffx-2", "ffx_2":
		return GameVersionFFX2, nil
	case "lastmiss", "last_miss", "last-miss", "lastmission", "last_mission", "last-mission":
		return GameVersionLastMiss, nil
	default:
		return GameVersion{}, fmt.Errorf("unknown game version: %q", s)
	}
}

// String retorna o rótulo canônico: "ffx", "ffx2" ou "lastmiss".
func (g GameVersion) String() string {
	return g.s
}

// MarshalJSON serializa a versão como string.
func (g GameVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.s)
}

// UnmarshalJSON desserializa a versão, validando via ParseGameVersionStrict.
func (g *GameVersion) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	v, err := ParseGameVersionStrict(s)
	if err != nil {
		return err
	}
	*g = v
	return nil
}

// Suffix retorna o sufixo de arquivo: _ffx, _ffx2 ou _lastmiss.
func (g GameVersion) Suffix() string {
	return "_" + g.s
}

// CharsetVersion normaliza a versão para lookup nos mapas de charset
// (bytes <-> string), que só têm buckets ffx e ffx2. LastMiss reaproveita
// os mapas de ffx2. Qualquer outra versão (inclui zero value) causa panic:
// versão desconhecida no encode/decode é bug, não dado ruim.
// Uso restrito às funções de bytes para string e string para bytes.
func CharsetVersion(gv GameVersion) GameVersion {
	switch gv {
	case GameVersionFFX:
		return GameVersionFFX
	case GameVersionFFX2, GameVersionLastMiss:
		return GameVersionFFX2
	default:
		panic(fmt.Sprintf("unknown game version for charset lookup: %q", gv.String()))
	}
}

// currentGameVersion é a versão ativa do processo.
// Inicializada com FFX; SetCurrentGameVersion assume GameVersion válido.
var currentGameVersion = GameVersionFFX

// CurrentGameVersion retorna a versão ativa (sempre válida).
func CurrentGameVersion() GameVersion {
	return currentGameVersion
}

// SetCurrentGameVersion define a versão ativa. Assume GameVersion válida.
func SetCurrentGameVersion(g GameVersion) {
	currentGameVersion = g
}