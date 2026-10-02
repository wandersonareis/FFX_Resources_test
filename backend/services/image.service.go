package services

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/ddsphyre"
)

// IMAGENS (.dds.phyre)
//
// O kind KindImages não é texto: não tem rows, não participa de
// ExportStrings/ExportJSON/ImportFile e não entra no preload de coleções.
// A leitura é só o scan da árvore + um decode por textura, então cada
// operação é independente e cacheless — o "store" do frontend guarda um
// único ImageEntry por vez (a textura visível).
//
// A imagem servida vem SEMPRE nesta ordem de preferência (determinada pelo
// pacote ddsphyre.Resolve, não aqui):
//
//	1. .dds em disco (mods/edits/images/... — extração anterior);
//	2. .png em disco (idem);
//	3. decode do próprio .dds.phyre EM MEMÓRIA (nada é gravado).
//
// Exportar (.dds + .png) é uma ação separada (ExtractImage) — abrir a
// textura não suja mods/.

// GetImage carrega a textura e devolve a pré-visualização pronta (data URL
// PNG) junto dos metadados do container e do DDS da fonte escolhida.
func (s *MetadataService) GetImage(kind, id string, version common.GameVersion) (dto.ImageEntry, error) {
	if !isImageKind(kind) {
		return dto.ImageEntry{}, fmt.Errorf("kind desconhecido: %s", kind)
	}
	if err := ensureVersionReady(version); err != nil {
		return dto.ImageEntry{}, err
	}
	if !ddsphyre.ValidID(id) {
		return dto.ImageEntry{}, fmt.Errorf("textura desconhecida: %s", id)
	}
	r, err := ddsphyre.Resolve(version, id)
	if err != nil {
		return dto.ImageEntry{}, err
	}
	entry := dto.ImageEntry{
		Metadata:       dto.NewImageMetadata(id, version),
		Source:         r.Source,
		Format:         r.Format,
		Width:          r.Width,
		Height:         r.Height,
		MipmapCount:    r.MipmapCount,
		MaxMipmapLevel: r.MaxMipmapLvl,
		PNGData:        dataURL("image/png", r.PNG),
		Flipped:        true,
		DDSData:        dataURL("application/octet-stream", r.DDS),
		DDSPath:        r.DDSPath,
		PNGPath:        r.PNGPath,
		Modded:         r.InMods,
	}
	s.fillDuplicates(&entry, id, version)
	return entry, nil
}

// fillDuplicates anexa o grupo de cópias de mesma imagem (payload original
// igual — as cópias da otimização do DVD, em qualquer diretório).
//
// O índice é memoizado no pacote ddsphyre; a primeira consulta da versão
// varre a árvore (~1-2 s) e as seguintes são O(1). Falha aqui não derruba a
// textura: duplicatas são informação suplementar, loga e segue.
func (s *MetadataService) fillDuplicates(entry *dto.ImageEntry, id string, version common.GameVersion) {
	entry.Duplicates = []dto.ImageDuplicate{}
	ix, err := ddsphyre.IndexFor(version)
	if err != nil {
		common.LogWarning("images %s: índice de duplicatas indisponível: %v", version, err)
		return
	}
	copies, payload := ix.Copies(id)
	entry.DupPayload = payload
	for _, c := range copies {
		entry.Duplicates = append(entry.Duplicates, dto.ImageDuplicate{
			ID:        c.ID,
			Key:       dto.NewImageMetadata(c.ID, version).Key,
			Modded:    c.InMods,
			Identical: c.Identical,
		})
	}
}

// ExtractImage grava .dds e .png em mods/edits/images (a cópia de trabalho
// que o translator edita) e devolve os caminhos escritos.
func (s *MetadataService) ExtractImage(kind, id string, version common.GameVersion) ([]string, error) {
	if !isImageKind(kind) {
		return nil, fmt.Errorf("kind desconhecido: %s", kind)
	}
	if err := ensureVersionReady(version); err != nil {
		return nil, err
	}
	if !ddsphyre.ValidID(id) {
		return nil, fmt.Errorf("textura desconhecida: %s", id)
	}
	paths, err := ddsphyre.Extract(version, id)
	if err != nil {
		return nil, err
	}
	// Atualiza o view (fonte passa a ser o .dds em disco).
	if s != nil && s.notifier != nil {
		s.notifier.NotifyInfo(fmt.Sprintf("Imagem exportada: ddsphyre/%s", id))
	}
	return paths, nil
}

// ImportImage reempacota o .dds escolhido sobre o container pristine e grava
// o resultado em mods/ (mesmo formato, mesma dimensão — realocar fixups é
// iteração futura, ver o TODO do pacote ddsphyre).
func (s *MetadataService) ImportImage(kind, id, ddsPath string, version common.GameVersion) error {
	if !isImageKind(kind) {
		return fmt.Errorf("kind desconhecido: %s", kind)
	}
	if err := ensureVersionReady(version); err != nil {
		return err
	}
	if !ddsphyre.ValidID(id) {
		return fmt.Errorf("textura desconhecida: %s", id)
	}
	if err := ddsphyre.Import(version, id, ddsPath); err != nil {
		return err
	}
	// O payload efetivo desta textura mudou: o índice de duplicatas em
	// cache está desatualizado (as cópias deixaram de ser idênticas).
	ddsphyre.InvalidateIndex(version)
	// Mantém os artefatos de trabalho em sincronia com o container recém-
	// gravado: sem isso a próxima abertura serviria o .dds extraído ANTIGO
	// (Resolve prefere .dds em disco), mascarando a textura importada.
	if _, err := ddsphyre.Extract(version, id); err != nil {
		common.LogWarning("images %s/%s: import ok, mas não atualizei os .dds/.png extraídos: %v", version, id, err)
	}
	if s != nil && s.notifier != nil {
		s.notifier.NotifyInfo(fmt.Sprintf("Imagem importada: ddsphyre/%s", id))
	}
	return nil
}

// ImportImageGroup reempacota o MESMO .dds sobre vários containers de uma
// vez — a propagação para as cópias idênticas (otimização do DVD): edita-se
// uma textura e as N iguais voltam a ficar sincronizadas.
//
// A validação é ANTES de qualquer gravação: todos os alvos precisam estar
// no grupo de mesmo payload ORIGINAL de id (alvo de outro grupo é recusado
// com erro, nada é escrito). Falhas individuais entram no resultado sem
// abortar o lote.
func (s *MetadataService) ImportImageGroup(kind, id, ddsPath string, targets []string, version common.GameVersion) (dto.ImageImportResult, error) {
	if !isImageKind(kind) {
		return dto.ImageImportResult{}, fmt.Errorf("kind desconhecido: %s", kind)
	}
	if err := ensureVersionReady(version); err != nil {
		return dto.ImageImportResult{}, err
	}
	if !ddsphyre.ValidID(id) {
		return dto.ImageImportResult{}, fmt.Errorf("textura desconhecida: %s", id)
	}
	ix, err := ddsphyre.IndexFor(version)
	if err != nil {
		return dto.ImageImportResult{}, fmt.Errorf("índice de duplicatas: %w", err)
	}
	group := ix.OriginalGroup(id)
	if len(group) == 0 {
		return dto.ImageImportResult{}, fmt.Errorf(
			"textura %s fora do índice de duplicatas — clique em Reanalisar", id)
	}

	list, err := groupTargets(id, group, targets)
	if err != nil {
		return dto.ImageImportResult{}, err
	}

	res := dto.ImageImportResult{
		Updated: []string{},
		Failed:  []string{},
		Total:   len(list),
	}
	for _, t := range list {
		if ierr := ddsphyre.Import(version, t, ddsPath); ierr != nil {
			res.Failed = append(res.Failed, t+": "+ierr.Error())
			continue
		}
		if _, eerr := ddsphyre.Extract(version, t); eerr != nil {
			common.LogWarning("images %s/%s: import ok, mas não atualizei os .dds/.png extraídos: %v", version, t, eerr)
		}
		res.Updated = append(res.Updated, t)
	}
	if len(res.Updated) > 0 {
		// Vários payloads efetivos mudaram: o índice é reconstruído na
		// próxima consulta.
		ddsphyre.InvalidateIndex(version)
		if s != nil && s.notifier != nil {
			s.notifier.NotifyInfo(fmt.Sprintf(
				"Imagem importada em %d de %d textura(s): %s",
				len(res.Updated), res.Total, filepath.Base(ddsPath),
			))
		}
	}
	return res, nil
}

// RefreshImageDuplicates descarta o índice de duplicatas em cache — a
// próxima consulta de textura reconstrói lendo a árvore.
//
// Existe por causa de edição EXTERNA (hex editor): o cache só detecta
// mudança que passa por aqui (import).
func (s *MetadataService) RefreshImageDuplicates(kind string, version common.GameVersion) error {
	if !isImageKind(kind) {
		return fmt.Errorf("kind desconhecido: %s", kind)
	}
	ddsphyre.InvalidateIndex(version)
	return nil
}

// groupTargets monta a lista final do lote e valida o grupo ANTES de
// qualquer gravação: id primeiro, sem repetições, preservando a ordem
// escolhida pelo frontend — e recusando qualquer alvo que não seja cópia de
// mesmo payload original de id (grupo vindo do índice de duplicatas).
func groupTargets(id string, group []string, targets []string) ([]string, error) {
	inGroup := make(map[string]bool, len(group))
	for _, g := range group {
		inGroup[g] = true
	}
	if !inGroup[id] {
		return nil, fmt.Errorf("textura %s fora do seu próprio grupo", id)
	}
	seen := map[string]bool{id: true}
	out := []string{id}
	for _, t := range targets {
		t = strings.TrimSpace(t)
		if t == "" || seen[t] {
			continue
		}
		if !inGroup[t] {
			return nil, fmt.Errorf(
				"importação em lote recusada: %s não é cópia idêntica de %s", t, id)
		}
		seen[t] = true
		out = append(out, t)
	}
	return out, nil
}

// SaveImage grava .dds ou .png no caminho escolhido pelo usuário, sem
// obrigar a ter extraído antes (é o "salvar em disco" do painel).
func (s *MetadataService) SaveImage(kind, id, format, destPath string, version common.GameVersion) error {
	if !isImageKind(kind) {
		return fmt.Errorf("kind desconhecido: %s", kind)
	}
	if err := ensureVersionReady(version); err != nil {
		return err
	}
	if !ddsphyre.ValidID(id) {
		return fmt.Errorf("textura desconhecida: %s", id)
	}
	return ddsphyre.Save(version, id, format, destPath)
}

func isImageKind(kind string) bool {
	return strings.EqualFold(strings.TrimSpace(kind), KindImages)
}

// dataURL empacota bytes como data URI para o <img> do frontend; vazio
// permanece vazio (o componente trata "sem cópia extraída").
func dataURL(media string, payload []byte) string {
	if len(payload) == 0 {
		return ""
	}
	return "data:" + media + ";base64," + base64.StdEncoding.EncodeToString(payload)
}
