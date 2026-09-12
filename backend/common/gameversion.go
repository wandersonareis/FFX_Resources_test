package common

import "strings"

// GameVersion é a identidade única da versão do jogo.
// Valores: "ffx", "ffx2", "lastmiss". É a única fonte de versão;
// não existem mais variantes int (1/2) nem sufixos v1/v2.
type GameVersion string

const (
	GameVersionFFX      GameVersion = "ffx"
	GameVersionFFX2     GameVersion = "ffx2"
	GameVersionLastMiss GameVersion = "lastmiss"
)

// FFX, FFX2 e LastMiss são atalhos legíveis para as constantes acima.
const (
	FFX      = GameVersionFFX
	FFX2     = GameVersionFFX2
	LastMiss = GameVersionLastMiss
)

// ParseGameVersion normaliza qualquer entrada para ffx/ffx2/lastmiss.
// Desconhecidos caem para ffx.
func ParseGameVersion(s string) GameVersion {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ffx2", "ffx-2", "ffx_2":
		return GameVersionFFX2
	case "lastmiss", "last_miss", "last-miss", "lastmission", "last_mission", "last-mission", "lm", "lm_accesary":
		return GameVersionLastMiss
	default:
		return GameVersionFFX
	}
}

// IsValid informa se é um dos três valores suportados.
func (g GameVersion) IsValid() bool {
	switch g {
	case GameVersionFFX, GameVersionFFX2, GameVersionLastMiss:
		return true
	default:
		return false
	}
}

// Normalize retorna a forma canônica, caindo para ffx se inválida.
func (g GameVersion) Normalize() GameVersion {
	if g.IsValid() {
		return g
	}
	return ParseGameVersion(string(g))
}

// String retorna o rótulo de exibição: FFX, FFX-2 ou LastMiss.
func (g GameVersion) String() string {
	switch g.Normalize() {
	case GameVersionFFX2:
		return "FFX-2"
	case GameVersionLastMiss:
		return "LastMiss"
	default:
		return "FFX"
	}
}

// Suffix retorna o sufixo de arquivo unificado: _ffx, _ffx2, _lastmiss.
func (g GameVersion) Suffix() string {
	return "_" + string(g.Normalize())
}

// currentGameVersion é a versão ativa do processo. É definida
// explicitamente via SetCurrentGameVersion (AppConfig/testes) —
// não via variável de ambiente.
var currentGameVersion GameVersion = GameVersionFFX

// CurrentGameVersion retorna a versão ativa.
func CurrentGameVersion() GameVersion {
	return currentGameVersion.Normalize()
}

// SetCurrentGameVersion define a versão ativa.
func SetCurrentGameVersion(g GameVersion) {
	currentGameVersion = g.Normalize()
}
