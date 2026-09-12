package common

import "path/filepath"

func GetEncodingDir() string {
	return GetEncodingDirForVersion(CurrentGameVersion())
}

// GetEncodingDirForVersion resolve o diretório de encoding sem depender
// de estado global ("ffx" -> ffx_encoding, "ffx2" -> ffx2_encoding,
// "lastmiss" reaproveita ffx2_encoding).
func GetEncodingDirForVersion(gv GameVersion) string {
	switch gv.Normalize() {
	case GameVersionFFX2, GameVersionLastMiss:
		return "ffx2_encoding"
	default:
		return "ffx_encoding"
	}
}

// GetEncodingDirForVersionString mantém compatibilidade com chamadores legados.
func GetEncodingDirForVersionString(gameVersionString string) string {
	return GetEncodingDirForVersion(ParseGameVersion(gameVersionString))
}

func GetEncodingPath(charset string) string {
	return GetEncodingPathForVersion(CurrentGameVersion(), charset)
}

// GetEncodingPathForVersion monta o caminho do charset para uma versão
// explícita, sem ler o estado global.
func GetEncodingPathForVersion(gv GameVersion, charset string) string {
	encodingDir := GetEncodingDirForVersion(gv)
	base := string(gv.Normalize())
	if gv.Normalize() == GameVersionLastMiss {
		base = string(GameVersionFFX2)
	}
	return filepath.Join(encodingDir, base+"sjistbl_"+charset+".bin")
}

// GetPathRoot monta a raiz ffx_ps2/<versão>/master da versão ativa.
func GetPathRoot() string {
	return GetPathRootForVersion(CurrentGameVersion())
}

// GetPathRootForVersion monta ffx_ps2/<versão>/master para uma versão explícita.
// LastMiss vive sob a árvore ffx2.
func GetPathRootForVersion(gv GameVersion) string {
	switch gv.Normalize() {
	case GameVersionFFX2, GameVersionLastMiss:
		return filepath.Join("ffx_ps2", string(GameVersionFFX2), "master")
	default:
		return filepath.Join("ffx_ps2", string(gv.Normalize()), "master")
	}
}

// GetPathRootForVersionString mantém compatibilidade com chamadores legados.
func GetPathRootForVersionString(gameVersionString string) string {
	return GetPathRootForVersion(ParseGameVersion(gameVersionString))
}

func GetPathOriginalsRoot() string {
	return filepath.Join(GetPathRoot(), OriginalsFolder)
}

func GetPathOriginalsKernel() string {
	return filepath.Join(GetPathOriginalsRoot(), "battle", "kernel")
}

func GetPathMonsterFolder() string {
	return filepath.Join(GetPathOriginalsRoot(), "battle", "mon")
}

func GetPathOriginalsEncounter() string {
	return filepath.Join(GetPathOriginalsRoot(), "battle", "btl")
}

func GetInternationalEncounterPath() string {
	return filepath.Join(GetPathRoot(), "inpc", "battle", "btl")
}

func GetPathOriginalsEvent() string {
	return filepath.Join(GetPathOriginalsRoot(), "event", "obj")
}

func GetAbmapPath() string {
	return filepath.Join(GetPathOriginalsRoot(), "menu", "abmap")
}

func GetLocalizationRoot(localization string) string {
	return filepath.Join(GetPathRoot(), "new_"+localization+"pc")
}
