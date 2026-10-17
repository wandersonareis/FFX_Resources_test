// ExportText: binário de texto APONTADO → artefato (.json/.strings/.csv/.xlsx) ao
// lado da entrada, com o NOME CANÔNICO do frontend (ExportBaseName). O
// binário pode estar em qualquer pasta — o decode é todo avulso (avulso.go).
//
// CLI futuro: `export <game> <binário> <formato> <saída>` parseia argv para
// este mesmo request — nenhuma lógica de export vive fora daqui.
package appcore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters"
	_ "ffxresources/backend/formatters/csv"
	_ "ffxresources/backend/formatters/json"
	_ "ffxresources/backend/formatters/strings"
	_ "ffxresources/backend/formatters/xlsx"
	"ffxresources/backend/services"
)

// ExportRequest é o pedido de export de texto (um lote de binários).
type ExportRequest struct {
	// Inputs são os binários apontados (multi-seleção do dialog).
	Inputs []string
	// Kind é o kind esperado (o submenu escolhido); "" aceita o que a
	// detecção disser (a raça do CLI sem submenu).
	Kind string
	// Format é o formato do artefato de saída (json/strings).
	Format ArtifactFormat
	// Version: Known=false pede a detecção (caminho/header).
	Version VersionRequest
	// Langs filtra os idiomas do artefato (nil/vazio = todos os carregados).
	Langs []string
	// OutDir muda o destino ("" = ao lado de cada entrada).
	OutDir string
	// Collision é a política de sobrescrita do destino.
	Collision CollisionPolicy
	// MacrodicPath aponta o dicionário sob demanda (ErrMacrodicUnavailable).
	MacrodicPath string
}

// ExportText decodifica cada input e grava o artefato. Erros individuais
// NÃO abortam o lote: viram FileOutcome (StatusMismatch/StatusFailed/
// StatusCollision); o retorno com error é só para falhas do pedido inteiro
// (nenhum input, runtime de texto indisponível).
func ExportText(ctx context.Context, req ExportRequest) (ExportResult, error) {
	if len(req.Inputs) == 0 {
		return ExportResult{}, fmt.Errorf("nenhum binário selecionado")
	}

	result := ExportResult{Outcomes: make([]FileOutcome, 0, len(req.Inputs))}
	for _, input := range req.Inputs {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		result.Outcomes = append(result.Outcomes, exportOne(req, input))
	}
	return result, nil
}

// exportOne processa UM binário do lote.
func exportOne(req ExportRequest, input string) FileOutcome {
	outcome := FileOutcome{Input: input, Status: StatusFailed}

	d, err := DetectFile(input)
	if err != nil {
		outcome.Reason = err.Error()
		if inc, ok := err.(*IncompatibleError); ok {
			outcome.Kind, _ = kindOfDetection(inc.Got)
			outcome.Status = StatusMismatch
		}
		return outcome
	}
	outcome.Kind = d.Kind

	// O submenu declara o kind esperado: divergência é o erro de formato
	// que nomeia os dois lados ("esperado Sistema, detectado events").
	if req.Kind != "" && !equalKind(req.Kind, d.Kind) {
		inc := Incompatible(input, kindLabel(req.Kind), kindLabel(d.Kind),
			"binário não corresponde ao submenu escolhido")
		outcome.Status = StatusMismatch
		outcome.Reason = inc.Error()
		return outcome
	}
	if d.Kind == "" || d.Kind == services.KindImages {
		outcome.Status = StatusMismatch
		outcome.Reason = Incompatible(input, kindLabel(req.Kind), d.Format.String(),
			"não é um binário de texto").Error()
		return outcome
	}

	version, known := req.Version.VersionOf(d)
	if !known {
		outcome.Reason = "versão do jogo não detectada (caminho sem pistas) — informe FFX ou FFX-2"
		return outcome
	}
	outcome.Version = version.String()

	if _, err := prepareTextRuntime(version, d.Kind, req.MacrodicPath); err != nil {
		outcome.Reason = err.Error()
		return outcome
	}

	data, err := os.ReadFile(input)
	if err != nil {
		outcome.Reason = fmt.Sprintf("lendo binário: %v", err)
		return outcome
	}

	lang := detectFileLanguage(input)
	c, err := decodeAvulso(d, version, input, data, lang)
	if err != nil {
		outcome.Status = StatusMismatch
		outcome.Reason = err.Error()
		return outcome
	}

	raw, err := marshalArtifact(req.Format, req.Langs, c)
	if err != nil {
		outcome.Reason = fmt.Sprintf("serializando artefato: %v", err)
		return outcome
	}

	key := artifactKeyOf(d.Kind, c)
	baseName := ExportBaseName(d.Kind, version, key, req.Format)
	target, existed := ExportOutputPath(input, baseName, req.Collision)
	if req.OutDir != "" {
		target = filepath.Join(req.OutDir, filepath.Base(target))
	}
	if existed && req.Collision == CollisionFail {
		outcome.Status = StatusCollision
		outcome.Reason = target
		return outcome
	}
	if err := writeChecked(target, raw); err != nil {
		outcome.Reason = err.Error()
		return outcome
	}

	outcome.Output = target
	if existed {
		outcome.Status = StatusOverwritten
	} else {
		outcome.Status = StatusWritten
	}
	return outcome
}

// marshalArtifact serializa a collection no formato pedido, com o MESMO
// formatter do registry para todo kind (os bytes JSON/strings são
// idênticos entre kinds) — a gravação em si é do chamador (writeChecked:
// política de colisão é do appcore, não do formato).
func marshalArtifact(format ArtifactFormat, langs []string, c dto.Collection) ([]byte, error) {
	if len(c) == 0 {
		return nil, fmt.Errorf("nada para exportar")
	}
	f, ok := formatters.Get(format.String())
	if !ok {
		return nil, fmt.Errorf("formatter %s não registrado", format)
	}
	return f.MarshalLangs(c, langs)
}

// artifactKeyOf devolve a chave da (primeira) entrada — o componente do
// nome canônico do artefato.
func artifactKeyOf(kind string, c dto.Collection) string {
	keys := c.SortedKeys()
	if len(keys) == 0 {
		return "artefato"
	}
	return keys[0]
}

// kindLabel devolve o rótulo humano do kind (o mesmo do frontend).
func kindLabel(kind string) string {
	switch kind {
	case services.KindEvents:
		return "Eventos"
	case services.KindObjects:
		return "Sistema"
	case services.KindMacro:
		return "Dicionário"
	case services.KindLockit:
		return "Loc Kit"
	case services.KindHelp:
		return "Painel de ajuda"
	case services.KindBattleText:
		return "Textos de batalha"
	case services.KindCloud:
		return "Cloud"
	case services.KindTutorial:
		return "Tutorial"
	case services.KindMenuMain:
		return "Texto menu principal"
	case services.KindImages:
		return "Imagens"
	default:
		return kind
	}
}

// equalKind compara kinds (case-insensitive, trim).
func equalKind(want, got string) bool {
	w := kindCanonical(want)
	g := kindCanonical(got)
	if w == "" || g == "" {
		return false
	}
	// A família de tabelas é uma só no codec: um .bin de battletext serve
	// o submenu de qualquer kind de tabela com mesmo conteúdo? Não — o
	// submenu é específico; a igualdade é estrita.
	return w == g
}

// kindCanonical normaliza o kind para comparação.
func kindCanonical(kind string) string {
	k := strings.ToLower(strings.TrimSpace(kind))
	switch k {
	case "events", "objects", "macro", "lockit", "help", "battletext", "cloud", "tutorial", "menumain", "images":
		return k
	default:
		return ""
	}
}

// kindOfDetection mapeia o rótulo de formato/kind de um IncompatibleError
// para o kind (quando dá).
func kindOfDetection(got string) (string, bool) {
	if k := kindCanonical(got); k != "" {
		return k, true
	}
	return "", false
}

// writeChecked grava bytes no destino SEM a trava de árvore do frontend:
// o acesso externo (menu/CLI) opera avulso — a saída vai ao lado do
// arquivo apontado, mesmo que isso signifique gravar DENTRO da árvore do
// jogo (data/). Quem mantém a trava é common.WriteBytesToFile, o caminho
// dos escritores do frontend.
func writeChecked(path string, data []byte) error {
	return common.WriteBytesToFileAnywhere(path, data)
}
