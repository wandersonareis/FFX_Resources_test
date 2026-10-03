package services

import (
	"path"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/ddsphyre"
	"ffxresources/backend/fileFormats/helpfile"
	"ffxresources/backend/fileFormats/lockit"
)

// MAPEAMENTO DE CAMINHO DO .vbf PARA ENTRADA DO APP.
//
// Um caminho dentro do container vale como entrada SÓ quando os readers já
// sabem servir aquele arquivo — ou seja, quando ele é exatamente o caminho
// que o próprio app resolveria para (kind, id, versão) em alguma
// localização. Nada de "adivinhar pelo sufixo": a confirmação é a igualdade
// com originalRelPathLoc, a mesma função usada para localizar o pristine em
// data/.

// vbfTarget é o alvo canônico de um caminho dentro de um .vbf.
type vbfTarget struct {
	// Kind/ID são os mesmos que a árvore data/ usa (events, objects, macro,
	// lockit, help, images).
	Kind string
	ID   string
	// Version é a árvore do arquivo: FFX_Data.vbf → ffx, FFX2_Data.vbf →
	// ffx2. Vem do PRÓPRIO caminho, não da aba aberta — um .vbf só tem
	// conteúdo de uma das duas árvores.
	Version common.GameVersion
}

// vbfVersionOf devolve a versão a que pertence um .vbf (ou um caminho
// interno). Marcadores específicos primeiro: "FFX_Data.vbf" contém "ffx"
// mas não "ffx2".
func vbfVersionOf(vbfName, inner string) common.GameVersion {
	n := strings.ToLower(vbfName + " " + inner)
	switch {
	case strings.Contains(n, "ffx2"),
		strings.Contains(n, "ffx-2"),
		strings.Contains(n, "ffx_ps2/ffx2/"):
		return common.GameVersionFFX2
	default:
		return common.GameVersionFFX
	}
}

// normPath normaliza um caminho para comparação: "/" como separador, sem
// barra inicial e em minúsculas.
func normPath(p string) string {
	s := strings.ReplaceAll(p, "\\", "/")
	return strings.ToLower(path.Clean("/" + s))
}

// cleanVbfPath devolve o caminho interno normalizado SEM barra inicial
// (forma como os caminhos do .vbf são indexados) e em minúsculas.
func cleanVbfPath(p string) string {
	return strings.TrimPrefix(normPath(p), "/")
}

// cleanVbfPathCase é cleanVbfPath preservando a caixa original: é dele que
// se extrai o id (as chaves do app preservam o caso do arquivo).
func cleanVbfPathCase(p string) string {
	s := strings.ReplaceAll(p, "\\", "/")
	return strings.TrimPrefix(path.Clean("/"+s), "/")
}

// eqPath compara dois caminhos de arquivo (separador e caixa ignorados).
func eqPath(a, b string) bool { return normPath(a) == normPath(b) }

// matchesAnyLoc confere se o caminho é EXATAMENTE o binário da entrada em
// ALGUMA localização — a confirmação de que o app sabe servir o arquivo, e
// não um "adivinhar pelo sufixo": é a mesma função (originalRelPathLoc) que
// localiza o pristine em data/.
//
// Qualquer pasta de idioma vale como clique (new_uspc, new_jppc, …) porque a
// leitura é sempre por TODAS as variantes — ver vbfOverlayPaths.
func matchesAnyLoc(kind, id string, version common.GameVersion, slash string) bool {
	for _, loc := range common.SupportedLanguageCodes() {
		if rel, ok := originalRelPathLoc(kind, id, version, loc); ok && eqPath(rel, slash) {
			return true
		}
	}
	return false
}

// imageIDFor extrai o id da textura de <raiz da versão>/<id>.dds.phyre
// ("" = caminho fora da raiz da versão — é o .vbf de outra árvore).
func imageIDFor(version common.GameVersion, slash string) string {
	root := normPath(ddsphyre.RootForVersion(version)) // "/ffx_data"
	if !strings.HasPrefix(normPath(slash), root+"/") {
		return ""
	}
	// A normalização não muda o comprimento de um caminho limpo: o prefixo
	// medido em `root` (com a barra inicial) vale para `slash` (sem ela).
	rel := slash[len(root):]
	rel = strings.TrimSuffix(rel, ddsphyre.Suffix)
	if !ddsphyre.ValidID(rel) {
		return ""
	}
	return rel
}

// lockitIDFor devolve o id (stem) do layout cujo arquivo por idioma é o
// caminho dado ("" = não é lockit desta versão).
func lockitIDFor(version common.GameVersion, slash string) string {
	for _, l := range lockit.LayoutsForVersion(version) {
		for _, lang := range lockitLangs(l) {
			if eqPath(l.RelPath(lang), slash) {
				return l.ID()
			}
		}
	}
	return ""
}

// lockitLangs devolve os idiomas do layout (com fallback para a lista
// suportada quando o layout não declara nenhum).
func lockitLangs(l lockit.Layout) []string {
	if len(l.Languages) > 0 {
		return l.Languages
	}
	return common.SupportedLanguageCodes()
}

// isKnownObject informa se o stem tem layout de objects registrado para a
// versão (é o que separa um binário de objetos de qualquer outro .bin).
func isKnownObject(version common.GameVersion, stem string) bool {
	_, ok := objectKeyForID(version, stem)
	return ok
}

// matchVbfPath resolve um caminho interno de .vbf para (kind, id, versão).
// ok=false = formato que o app não serve (áudio, vídeo, .exe, painéis
// secundários…) — o navegador de árvore mostra o arquivo, mas abri-lo avisa
// em vez de quebrar.
func matchVbfPath(vbfName, inner string) (vbfTarget, bool) {
	version := vbfVersionOf(vbfName, inner)
	slash := cleanVbfPathCase(inner)
	if slash == "" {
		return vbfTarget{}, false
	}
	lower := normPath(slash)
	base := path.Base(lower)
	stem := strings.TrimSuffix(base, path.Ext(base))

	// Heurística de FORMA → (kind, id) candidato. A confirmação é a
	// igualdade com o caminho que o próprio app resolveria, adiante.
	var kind, id string
	switch {
	case strings.HasSuffix(lower, ddsphyre.Suffix):
		// Textura: <raiz da versão>/<id>.dds.phyre.
		kind = KindImages
		if id = imageIDFor(version, slash); id == "" {
			return vbfTarget{}, false
		}

	case lockit.IsLockitKey(lower):
		// lockit: <raiz da versão>/gamedata/ps3data/lockit/<stem>_<lang>.bin
		// — vive FORA da pasta de localização, um arquivo por idioma.
		kind = KindLockit
		if id = lockitIDFor(version, slash); id == "" {
			return vbfTarget{}, false
		}

	case strings.EqualFold(base, "macrodic.dcp"):
		// macro: um dicionário por localização; o id é o CHUNK, que só
		// existe depois de decodificar (o chamador passa chunk_XX).
		kind, id = KindMacro, ""

	case strings.HasSuffix(lower, helpfile.HelpFileExt) &&
		strings.Contains(lower, "/"+helpfile.HelpDirPrefix+"/") &&
		helpfile.IsHelpEntry(stem):
		// help: <raiz de localização>/help/<dir>/<id>.sps2 (só FFX). Os
		// painéis secundários (…_page em subpasta) não são entradas.
		kind, id = KindHelp, stem

	case isKnownObject(version, stem):
		// objects: <raiz de localização>/<pattern>, id = stem do arquivo.
		kind, id = KindObjects, stem

	default:
		// events: <raiz de localização>/event/obj_ps3/<xx>/<id>/<id>.bin
		kind, id = KindEvents, stem
	}

	if !matchesAnyLoc(kind, id, version, slash) {
		return vbfTarget{}, false
	}
	return vbfTarget{Kind: kind, ID: id, Version: version}, true
}

// vbfOverlayPaths devolve os caminhos RELATIVOS que o decodificador de
// (kind, id) vai tocar — todos eles, não só o clicado.
//
// Isso existe porque eventos, help e macro têm UM ARQUIVO POR LOCALIZAÇÃO:
// o reader varre new_uspc, new_jppc, new_krpc… e o resultado vira uma row
// por idioma. Montar o overlay só com o arquivo clicado faria a coluna
// Original aparecer em um único idioma.
func vbfOverlayPaths(t vbfTarget, clicked string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(common.SupportedLanguageCodes())+1)
	add := func(p string) {
		if p == "" {
			return
		}
		k := normPath(p)
		if seen[k] {
			return
		}
		seen[k] = true
		out = append(out, p)
	}

	add(clicked)

	switch t.Kind {
	case KindImages:
		// Imagem é um arquivo único (o próprio caminho).
	case KindLockit:
		// lockit não tem raiz de localização: tem um arquivo por idioma.
		if l, ok := lockit.LayoutForID(t.Version, t.ID); ok {
			for _, lang := range lockitLangs(l) {
				add(l.RelPath(lang))
			}
		}
	default:
		for _, loc := range common.SupportedLanguageCodes() {
			if rel, ok := originalRelPathLoc(t.Kind, t.ID, t.Version, loc); ok {
				add(rel)
			}
		}
	}
	return out
}
