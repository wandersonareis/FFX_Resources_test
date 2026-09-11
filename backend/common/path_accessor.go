package common

import "path/filepath"

func GetEncodingDir() string {
	return GetEncodingDirForVersionString(GetGameVersionString())
}

// GetEncodingDirForVersionString resolve o diretório de encoding sem depender
// do GAME_VERSION global ("ffx" -> ffx_encoding, "ffx2" -> ffx2_encoding).
func GetEncodingDirForVersionString(gameVersionString string) string {
	switch gameVersionString {
	case "ffx2":
		return "ffx2_encoding"
	default:
		return "ffx_encoding"
	}
}

func GetEncodingPath(charset string) string {
	return GetEncodingPathForVersion(GetGameVersionString(), charset)
}

// GetEncodingPathForVersion monta o caminho do charset para uma versão
// explícita, sem ler o estado global.
func GetEncodingPathForVersion(gameVersionString, charset string) string {
	encodingDir := GetEncodingDirForVersionString(gameVersionString)
	return filepath.Join(encodingDir, gameVersionString+"sjistbl_"+charset+".bin")
}

func GetPathRoot() string {
	return GetPathRootForVersion(GetGameVersionString())
}

// GetPathRootForVersion monta ffx_ps2/<versão>/master para uma versão explícita.
func GetPathRootForVersion(gameVersionString string) string {
	return filepath.Join("ffx_ps2", gameVersionString, "master")
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
