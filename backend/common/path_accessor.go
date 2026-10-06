package common

import "path/filepath"

// GetPathRoot monta a raiz ffx_ps2/<versão>/master da versão ativa.
func GetPathRoot() string {
	return GetPathRootForVersion(CurrentGameVersion())
}

// GetPathRootForVersion monta ffx_ps2/<versão>/master para uma versão explícita.
// EternalCalm vive sob a árvore ffx; LastMiss, sob a ffx2.
func GetPathRootForVersion(gv GameVersion) string {
	switch gv {
	case GameVersionFFX, GameVersionEternalCalm:
		return filepath.Join("ffx_ps2", GameVersionFFX.String(), "master")
	case GameVersionFFX2, GameVersionLastMiss:
		return filepath.Join("ffx_ps2", GameVersionFFX2.String(), "master")
	default:
		return filepath.Join("ffx_ps2", gv.String(), "master")
	}
}

func GetPathOriginalsRoot() string {
	return filepath.Join(GetPathRoot(), OriginalsFolder)
}

func GetLocalizationRoot(localization string) string {
	return filepath.Join(GetPathRoot(), "new_"+localization+"pc")
}

// VersionPathName retorna o nome do diretório de versão usado em caminhos de
// arquivo: "ffx" ou "ffx2". EternalCalm divide a árvore com o FFX e LastMiss
// com o FFX-2 (expansões), então normalizam para o jogo-pai. Só existem
// esses dois nomes, sem variação.
func VersionPathName(version GameVersion) string {
	switch version {
	case GameVersionFFX, GameVersionEternalCalm:
		return GameVersionFFX.String()
	case GameVersionFFX2, GameVersionLastMiss:
		return GameVersionFFX2.String()
	default:
		return version.String()
	}
}

// GetLocalizationRootForVersion monta a raiz de localização para uma versão
// explícita, sem ler a versão global. EternalCalm resolve sob a árvore ffx;
// LastMiss, sob a ffx2.
func GetLocalizationRootForVersion(version GameVersion, localization string) string {
	return filepath.Join(GetPathRootForVersion(version), "new_"+localization+"pc")
}

// PackRootForVersion é o prefixo constante de build usado para montar
// event_file_path = packRoot + localization_pattern. Varia por versão
// (ffx vs eternalcalm/ffx2/lastmiss) e localização, por isso é função e
// não const.
func PackRootForVersion(version GameVersion, localization string) string {
	if localization == "" {
		localization = DefaultLocalization
	}
	return GetLocalizationRootForVersion(version, localization)
}
