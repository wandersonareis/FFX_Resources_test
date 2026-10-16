// Dedupe/Dupe: operações de artefato (.json/.strings) arquivo→arquivo.
//
//	DEDUPE — ler o arquivo, expandir as refs que já existem e REESCREVER
//	colapsando os textos repetidos em referência "$hash" (o marshal do
//	app): saída dedupe_<stem>.<ext>.
//
//	DUPE — ler o arquivo (as refs "$hash" saem expandidas pelo Unmarshal)
//	e reescrever com TODO o texto literal, sem colapsar nada: saída
//	dupe_<stem>.<ext>.
//
// Ambos SOBRESCRVEM o destino por design (o nome já é prefixado) — o
// chamador (menu) pergunta a certeza antes; o appcore apenas grava.
package appcore

import (
	"context"
	"fmt"
	"os"

	"ffxresources/backend/dto"
	jsonfmt "ffxresources/backend/formatters/json"
	strfmt "ffxresources/backend/formatters/strings"
)

// ArtifactRequest é o pedido dedupe/dupe (um artefato por chamada — o menu
// itera a seleção).
type ArtifactRequest struct {
	// Path é o artefato .json/.strings apontado.
	Path string
	// Langs filtra os idiomas (nil/vazio = todos).
	Langs []string
}

// DedupeArtifact colapsa textos repetidos do artefato em refs "$hash".
// A saída é dedupe_<stem>.<ext> ao lado da entrada.
func DedupeArtifact(ctx context.Context, req ArtifactRequest) (FileOutcome, error) {
	return artifactOp(ctx, req, PrefixDedupe, "dedupe")
}

// DupeArtifact resolve os hashes de conteúdo do artefato de volta ao texto
// completo (marshal sem dedup). A saída é dupe_<stem>.<ext>.
func DupeArtifact(ctx context.Context, req ArtifactRequest) (FileOutcome, error) {
	return artifactOp(ctx, req, PrefixDupe, "dupe")
}

// artifactOp é o motor comum: read (expande refs) → marshal (com ou sem o
// colapso) → grava com prefixo, sobrescrevendo.
func artifactOp(ctx context.Context, req ArtifactRequest, prefix, label string) (FileOutcome, error) {
	outcome := FileOutcome{Input: req.Path, Status: StatusFailed}

	select {
	case <-ctx.Done():
		return outcome, ctx.Err()
	default:
	}

	d, err := DetectFile(req.Path)
	if err != nil {
		outcome.Status = StatusMismatch
		outcome.Reason = err.Error()
		return outcome, nil
	}
	if d.Format != FormatArtifactJSON && d.Format != FormatArtifactStrings {
		outcome.Status = StatusMismatch
		outcome.Reason = Incompatible(req.Path, "artefato .json ou .strings", d.Format.String(),
			"operação "+label+" só roda em artefato de texto").Error()
		return outcome, nil
	}

	raw, err := os.ReadFile(req.Path)
	if err != nil {
		return outcome, fmt.Errorf("lendo artefato: %w", err)
	}

	// O Unmarshal EXPANDE as refs "$hash" atuais pelo valor da tabela
	// hash→texto do próprio arquivo — o merge de entrada é sempre texto
	// completo, não importa o estado do artefato escolhido.
	var c dto.Collection
	if d.Format == FormatArtifactJSON {
		c, err = jsonfmt.NewJSONEventsFormatter().Unmarshal(raw)
	} else {
		c, err = strfmt.NewStringsFormatter().Unmarshal(raw)
	}
	if err != nil {
		outcome.Status = StatusMismatch
		outcome.Reason = fmt.Sprintf("falha ao analisar artefato: %v", err)
		return outcome, nil
	}
	if len(c) == 0 {
		outcome.Status = StatusMismatch
		outcome.Reason = "arquivo sem entradas"
		return outcome, nil
	}

	// dedupe: marshal padrão (colapsa repetidos em $hash);
	// dupe: MarshalNoDedup (texto literal — resolve os hashes de conteúdo).
	var out []byte
	if prefix == PrefixDedupe {
		if d.Format == FormatArtifactJSON {
			out, err = jsonfmt.NewJSONObjectFormatter().MarshalLangs(c, req.Langs)
		} else {
			out, err = strfmt.MarshalLangs(c, req.Langs)
		}
	} else {
		if d.Format == FormatArtifactJSON {
			out, err = jsonfmt.MarshalNoDedupLangs(c, req.Langs)
		} else {
			out, err = strfmt.MarshalNoDedupLangs(c, req.Langs)
		}
	}
	if err != nil {
		outcome.Reason = fmt.Sprintf("serializando %s: %v", label, err)
		return outcome, nil
	}

	target := ArtifactPrefixedPath(req.Path, prefix)
	existed := fileExists(target)
	if err := writeChecked(target, out); err != nil {
		outcome.Reason = err.Error()
		return outcome, nil
	}

	outcome.Output = target
	if existed {
		outcome.Status = StatusOverwritten
	} else {
		outcome.Status = StatusWritten
	}
	return outcome, nil
}

// ArtifactTargetExists informa se o destino prefixado já existe — o
// chamador pergunta a certeza ANTES de chamar a operação.
func ArtifactTargetExists(path, prefix string) bool {
	return fileExists(ArtifactPrefixedPath(path, prefix))
}
