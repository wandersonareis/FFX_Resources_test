package common

import "path/filepath"

func GetEncodingDir() string {
	return GetEncodingDirForVersion(CurrentGameVersion())
}

// GetEncodingDirForVersion resolve o diretório de encoding sem depender
// de estado global ("ffx" -> ffx_encoding, "ffx2" -> ffx2_encoding,
// "lastmiss" reaproveita ffx2_encoding).
func GetEncodingDirForVersion(gv GameVersion) string {
	switch gv {
	case GameVersionFFX2, GameVersionLastMiss:
		return "ffx2_encoding"
	default:
		return "ffx_encoding"
	}
}

// GetEncodingDirForVersionString mantém compatibilidade com chamadores legados.
func GetEncodingDirForVersionString(gameVersionString string) string {
	gv, err := ParseGameVersionStrict(gameVersionString)
	if err != nil {
		gv = GameVersionFFX
	}
	return GetEncodingDirForVersion(gv)
}

func GetEncodingPath(charset string) string {
	return GetEncodingPathForVersion(CurrentGameVersion(), charset)
}

// GetEncodingPathForVersion monta o caminho do charset para uma versão
// explícita, sem ler o estado global.
func GetEncodingPathForVersion(gv GameVersion, charset string) string {
	encodingDir := GetEncodingDirForVersion(gv)
	base := gv.String()
	if gv == GameVersionLastMiss {
		base = GameVersionFFX2.String()
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
	switch gv {
	case GameVersionFFX2, GameVersionLastMiss:
		return filepath.Join("ffx_ps2", GameVersionFFX2.String(), "master")
	default:
		return filepath.Join("ffx_ps2", gv.String(), "master")
	}
}

// GetPathRootForVersionString mantém compatibilidade com chamadores legados.
func GetPathRootForVersionString(gameVersionString string) string {
	gv, err := ParseGameVersionStrict(gameVersionString)
	if err != nil {
		gv = GameVersionFFX
	}
	return GetPathRootForVersion(gv)
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

// VersionPathName retorna o nome do diretório de versão usado em caminhos de
// arquivo: "ffx" ou "ffx2". LastMiss divide a árvore com FFX2 (expansão),
// então normaliza para "ffx2". Só existem esses dois nomes, sem variação.
func VersionPathName(version GameVersion) string {
	if version == GameVersionFFX {
		return GameVersionFFX.String()
	}
	return GameVersionFFX2.String()
}

// GetLocalizationRootForVersion monta a raiz de localização para uma versão
// explícita, sem ler a versão global. LastMiss resolve sob a árvore ffx2.
func GetLocalizationRootForVersion(version GameVersion, localization string) string {
	return filepath.Join(GetPathRootForVersion(version), "new_"+localization+"pc")
}

// PackRootForVersion é o prefixo constante de build usado para montar
// event_file_path = packRoot + localization_pattern. Varia por versão
// (ffx vs ffx2/lastmiss) e localização, por isso é função e não const.
func PackRootForVersion(version GameVersion, localization string) string {
	if localization == "" {
		localization = DefaultLocalization
	}
	return GetLocalizationRootForVersion(version, localization)
}