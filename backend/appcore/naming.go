// Nomes de saída das operações arquivo→arquivo — as convenções do frontend
// (services/export_paths.go e formatters), aplicadas AO LADO do arquivo
// apontado:
//
//   - export de texto: o NOME CANÔNICO que o frontend usaria para o mesmo
//     escopo (event_<id>_all_localizations[_<v>].json, <key>_all_…,
//     macro_dictionary_all_…, help_<painel>_all_…), com o sufixo de versão
//     de common.WithVersionSuffixFor;
//   - extração de imagem: <stem>.dds / <stem>.png;
//   - import: imported/<basename real> ao lado de CADA binário;
//   - dedupe/dupe: dedupe_<stem>.<ext> / dupe_<stem>.<ext>.
//
// Colisão: a política padrão NÃO sobrescreve (CollisionFail devolve
// StatusCollision e o chamador decide); CollisionPrefix prefixa
// (export_/extract_); CollisionOverwrite grava por cima.
package appcore

import (
	"os"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/services"
)

// Prefixos por operação — a identidade de origem num directory listing.
const (
	PrefixExport  = "export_"
	PrefixExtract = "extract_"
	PrefixDedupe  = "dedupe_"
	PrefixDupe    = "dupe_"

	// ImportDir é a pasta criada ao lado do binário no import.
	ImportDir = "imported"
)

// ExportBaseName devolve o nome canônico do artefato de export para o
// escopo de UMA entrada (a régua de nomes do frontend, ver
// services/export_paths.go): events usa event_<id>_…, o bulk de macro é o
// dicionário inteiro (o container é um artefato só), painéis de help
// ganham o prefixo help_ e o resto usa a própria chave da Collection.
func ExportBaseName(kind string, version common.GameVersion, key string, format ArtifactFormat) string {
	var base string
	switch kind {
	case services.KindEvents:
		base = "event_" + key + "_all_localizations"
	case services.KindMacro:
		// O container é um artefato só por localização (todos os chunks).
		base = "macro_dictionary_all_localizations"
	case services.KindHelp:
		base = "help_" + key + "_all_localizations"
	default:
		// objects/lockit/battletext/cloud/tutorial/menumain: a chave da
		// Collection é o basename do arquivo (ou o id do layout).
		base = key + "_all_localizations"
	}
	return common.WithVersionSuffixFor(base+format.Ext(), version)
}

// ExportOutputPath resolve o destino do export ao lado da entrada, com a
// política de colisão aplicada. existed devolve se o destino escolhido já
// existia ANTES da gravação (o chamador marca StatusOverwritten).
func ExportOutputPath(input, baseName string, policy CollisionPolicy) (out string, existed bool) {
	dir := filepath.Dir(input)
	name := baseName
	if fileExists(filepath.Join(dir, name)) {
		existed = true
		switch policy {
		case CollisionPrefix:
			name = PrefixExport + name
		case CollisionOverwrite:
			// mantém o nome
		case CollisionFail:
			return filepath.Join(dir, name), true
		}
	}
	return filepath.Join(dir, name), existed && policy != CollisionPrefix
}

// ExtractImageOutputPath resolve o destino da extração de imagem
// (<stem>.dds/.png ao lado do .dds.phyre), com a mesma régua de colisão do
// export (prefixo extract_).
func ExtractImageOutputPath(phyrePath string, artifact ImageArtifact, policy CollisionPolicy) (out string, existed bool) {
	dir := filepath.Dir(phyrePath)
	stem := strings.TrimSuffix(filepath.Base(phyrePath), ".dds.phyre")
	if !strings.EqualFold(filepath.Ext(phyrePath), ".phyre") {
		stem = strings.TrimSuffix(filepath.Base(phyrePath), filepath.Ext(phyrePath))
	}
	name := stem + artifact.Ext()
	if fileExists(filepath.Join(dir, name)) {
		existed = true
		switch policy {
		case CollisionPrefix:
			name = PrefixExtract + name
		case CollisionOverwrite:
		case CollisionFail:
			return filepath.Join(dir, name), true
		}
	}
	return filepath.Join(dir, name), existed && policy != CollisionPrefix
}

// ImportOutputPath resolve o destino do import: imported/<basename real>
// ao lado do próprio binário apontado (cada binário do lote grava na sua
// pasta — a origem preservada, sem colisão entre homônimos de pastas
// diferentes).
func ImportOutputPath(binaryPath string, policy CollisionPolicy) (out string, existed bool) {
	dir := filepath.Join(filepath.Dir(binaryPath), ImportDir)
	name := filepath.Base(binaryPath)
	target := filepath.Join(dir, name)
	if fileExists(target) {
		existed = true
		if policy == CollisionFail {
			return target, true
		}
	}
	return target, existed
}

// ArtifactPrefixedPath resolve o destino do dedupe/dupe: prefixo fixo
// (PrefixDedupe/PrefixDupe) e SOBRESCRITA por design — a política é sempre
// CollisionOverwrite; existia é só para o chamador confirmar antes.
func ArtifactPrefixedPath(artifactPath, prefix string) string {
	base := filepath.Base(artifactPath)
	return filepath.Join(filepath.Dir(artifactPath), prefix+base)
}

// PrefixedPath é AliasArtifactPrefixedPath para os chamadores do menu/CLI.
func PrefixedPath(artifactPath, prefix string) string {
	return ArtifactPrefixedPath(artifactPath, prefix)
}

// fileExists informa se o caminho existe (qualquer erro conta como não).
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
