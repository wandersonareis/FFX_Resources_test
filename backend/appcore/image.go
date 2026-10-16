// Operações de IMAGEM arquivo→arquivo (.dds.phyre):
//
//	EXTRAIR — container apontado → .dds ou .png AO LADO dele (mesma régua de
//	colisão do export: sem colisão grava <stem>.<ext>; com colisão o
//	chamador decide prefixo extract_/sobrescrever);
//
//	IMPORTAR — .dds apontado + containers .dds.phyre apontados (multi): o
//	casamento é por DIMENSÕES e FORMATO DX (o nome pode ser ignorado,
// desde que batam as informações do header), o pack é sobre o container
// APONTADO (ddsphyre.PackDDS) e a saída é imported/<stem>.dds.phyre ao
// lado de cada container.
//
// Imagem não precisa de encoding/macrodic (ver TODO do runtime.go).
package appcore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ffxresources/backend/fileFormats/ddsphyre"
)

// ExtractImageRequest é o pedido de extração de textura.
type ExtractImageRequest struct {
	// Phyres são os containers .dds.phyre apontados (multi-seleção).
	Phyres []string
	// Artifact é o que extrair (.dds ou .png).
	Artifact ImageArtifact
	// OutDir muda o destino ("" = ao lado de cada container).
	OutDir string
	// Collision é a política de sobrescrita (régua do export).
	Collision CollisionPolicy
}

// ImportImageRequest é o pedido de import de textura.
type ImportImageRequest struct {
	// DDS é o .dds apontado (a imagem nova).
	DDS string
	// Phyres são os containers .dds.phyre apontados (multi-seleção).
	Phyres []string
	// OutDir muda o destino ("" = imported/ ao lado de cada container).
	OutDir string
	// Collision é a política de sobrescrita do binário de saída.
	Collision CollisionPolicy
}

// ExtractImage extrai .dds ou .png de cada container apontado.
func ExtractImage(ctx context.Context, req ExtractImageRequest) (ExportResult, error) {
	if len(req.Phyres) == 0 {
		return ExportResult{}, fmt.Errorf("nenhuma textura selecionada")
	}
	result := ExportResult{Outcomes: make([]FileOutcome, 0, len(req.Phyres))}
	for _, phyre := range req.Phyres {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		result.Outcomes = append(result.Outcomes, extractOneImage(req, phyre))
	}
	return result, nil
}

// extractOneImage processa UM container.
func extractOneImage(req ExtractImageRequest, phyre string) FileOutcome {
	outcome := FileOutcome{Input: phyre, Kind: "images", Status: StatusFailed}

	d, err := DetectFile(phyre)
	if err != nil {
		outcome.Status = StatusMismatch
		outcome.Reason = err.Error()
		return outcome
	}
	if d.Format != FormatPhyre {
		outcome.Status = StatusMismatch
		outcome.Reason = Incompatible(phyre, "textura .dds.phyre", d.Format.String(), "não é uma textura").Error()
		return outcome
	}

	texture, err := ddsphyre.ParseFile(phyre)
	if err != nil {
		outcome.Status = StatusMismatch
		outcome.Reason = fmt.Sprintf("decodificando textura: %v", err)
		return outcome
	}

	var raw []byte
	switch req.Artifact {
	case ImageDDS:
		raw, err = texture.ExtractToDDS()
	default:
		var dds []byte
		dds, err = texture.ExtractToDDS()
		if err == nil {
			raw, err = ddsphyre.DDSToPNG(dds)
		}
	}
	if err != nil {
		outcome.Reason = fmt.Sprintf("convertendo para %s: %v", req.Artifact, err)
		return outcome
	}

	target, existed := ExtractImageOutputPath(phyre, req.Artifact, req.Collision)
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

// ImportImageFile empacota o .dds apontado em cada container compatível.
// O casamento IGNORA o nome: o que precisa bater são as dimensões e o
// formato DX do header do DDS contra o do container (mesmo formato, mesma
// dimensão — realocar fixups é iteração futura do ddsphyre).
func ImportImageFile(ctx context.Context, req ImportImageRequest) (ImportResult, error) {
	if strings.TrimSpace(req.DDS) == "" {
		return ImportResult{}, fmt.Errorf("nenhum .dds selecionado")
	}
	if len(req.Phyres) == 0 {
		return ImportResult{}, fmt.Errorf("nenhum .dds.phyre selecionado")
	}

	dds, err := os.ReadFile(req.DDS)
	if err != nil {
		return ImportResult{}, fmt.Errorf("lendo %s: %w", filepath.Base(req.DDS), err)
	}

	result := ImportResult{Outcomes: make([]FileOutcome, 0, len(req.Phyres))}
	for _, phyre := range req.Phyres {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		result.Outcomes = append(result.Outcomes, importOneImage(req, dds, phyre))
	}
	return result, nil
}

// importOneImage processa UM container.
func importOneImage(req ImportImageRequest, dds []byte, phyre string) FileOutcome {
	outcome := FileOutcome{Input: phyre, Kind: "images", Status: StatusFailed}

	d, err := DetectFile(phyre)
	if err != nil {
		outcome.Status = StatusMismatch
		outcome.Reason = err.Error()
		return outcome
	}
	if d.Format != FormatPhyre {
		outcome.Status = StatusMismatch
		outcome.Reason = Incompatible(phyre, "textura .dds.phyre", d.Format.String(), "não é uma textura").Error()
		return outcome
	}

	container, err := os.ReadFile(phyre)
	if err != nil {
		outcome.Reason = fmt.Errorf("lendo textura: %w", err).Error()
		return outcome
	}

	// PackDDS valida o par (formato/dimensões) — divergência é erro de
	// formato nomeando os DOIS arquivos, como no import de texto.
	packed, err := ddsphyre.PackDDS(container, dds)
	if err != nil {
		outcome.Status = StatusMismatch
		outcome.Reason = Incompatible(phyre,
			fmt.Sprintf("textura compatível com %s", filepath.Base(req.DDS)),
			"textura .dds.phyre", err.Error()).Error()
		return outcome
	}

	target, existed := ImportOutputPath(phyre, req.Collision)
	if req.OutDir != "" {
		target = filepath.Join(req.OutDir, filepath.Base(phyre))
	}
	if existed && req.Collision == CollisionFail {
		outcome.Status = StatusCollision
		outcome.Reason = target
		return outcome
	}
	if err := writeChecked(target, packed); err != nil {
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
