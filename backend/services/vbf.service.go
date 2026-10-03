package services

// NAVEGADOR DE .vbf — SOMENTE LEITURA.
//
// O container (~20 GB) nunca é lido inteiro nem pré-extraído: a árvore da
// sidebar sai do ÍNDICE (cabeçalho, ~10-18 MB) e só o arquivo clicado é
// decodificado, sob demanda, em memória. Nada aqui escreve no .vbf — nem o
// export, que lê do container (ou de um temp no dir de execução) e grava
// sempre em mods/.
//
// Fluxo por clique:
//
//	ListVbfRoots   → descobre os .vbf perto do executável do jogo;
//	ListVbfDir     → filhos imediatos de um diretório (só paths/contagens);
//	ListVbfMacroChunks → filhos virtuais de macrodic.dcp (chunk_XX);
//	GetVbfTextEntry    → decodifica o arquivo e devolve a tabela;
//	GetVbfImageEntry   → decodifica a textura e devolve a pré-visualização.
//
// Os dois Get* rodam DENTRO de common.WithVbfSourceReader: o overlay fica
// vivo só durante a chamada (escopo por chamada, serializado) e some no fim,
// quaisquer que sejam retorno ou pânico.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/vbf"
	"ffxresources/backend/interactions"
	"ffxresources/backend/loggingService"
)

// ---------------------------------------------------------------------------
// descoberta das raízes
// ---------------------------------------------------------------------------

// gameExePath devolve o executável do jogo configurado ("" quando ausente).
// É dele que sai o diretório onde os .vbf moram — um campo de ARQUIVO no
// config, fora do map Locations (que é de diretórios e faz MkdirAll).
func gameExePath() string {
	svc := interactions.NewInteractionService()
	if svc == nil {
		return ""
	}
	cfg := svc.FFXAppConfig()
	if cfg == nil {
		return ""
	}
	return strings.TrimSpace(cfg.GetGameExeLocation())
}

// vbfSearchDirs devolve os diretórios varridos por *.vbf: a pasta do
// executável e a pasta data/ ao lado dele.
func vbfSearchDirs() []string {
	exe := gameExePath()
	if exe == "" {
		return nil
	}
	dir := filepath.Dir(exe)
	return []string{dir, filepath.Join(dir, common.DirData)}
}

// discoverVbfRoots varre os diretórios de busca. Sem executável configurado
// devolve vazio (não é erro: ainda não há o que navegar).
func discoverVbfRoots() []dto.VbfRoot {
	seen := map[string]bool{}
	roots := make([]dto.VbfRoot, 0, 2)

	for _, dir := range vbfSearchDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // diretório inexistente/ilegível: segue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".vbf") {
				continue
			}
			abs, aerr := filepath.Abs(filepath.Join(dir, e.Name()))
			if aerr != nil {
				continue
			}
			key := strings.ToLower(abs)
			if seen[key] {
				continue
			}
			info, ierr := e.Info()
			if ierr != nil {
				continue
			}
			seen[key] = true
			roots = append(roots, dto.VbfRoot{
				Name:    e.Name(),
				Path:    abs,
				Version: vbfVersionOf(e.Name(), "").String(),
				Size:    info.Size(),
			})
		}
	}

	sort.Slice(roots, func(i, j int) bool { return roots[i].Name < roots[j].Name })
	return roots
}

// ---------------------------------------------------------------------------
// registro de containers abertos
// ---------------------------------------------------------------------------

var (
	vbfArchMu   sync.Mutex
	vbfArchives = map[string]*vbf.Archive{}
)

// vbfArchiveFor abre (uma única vez) e devolve o .vbf pedido. Só aceita
// caminhos que a descoberta encontrou perto do executável: o argumento vem
// da UI e não pode virar leitura arbitrária de qualquer arquivo do disco.
//
// Manter o handle aberto evita reler o cabeçalho (10-18 MB) a cada clique;
// a descoberta é barata, então é refeita a cada chamada — o exe pode ter
// mudado de lugar.
func vbfArchiveFor(p string) (*vbf.Archive, error) {
	abs, err := filepath.Abs(strings.TrimSpace(p))
	if err != nil {
		return nil, err
	}
	key := strings.ToLower(abs)

	vbfArchMu.Lock()
	defer vbfArchMu.Unlock()

	if a, ok := vbfArchives[key]; ok {
		return a, nil
	}
	known := false
	for _, r := range discoverVbfRoots() {
		if strings.EqualFold(r.Path, abs) {
			known = true
			break
		}
	}
	if !known {
		return nil, fmt.Errorf(".vbf não está na pasta do jogo: %s", filepath.Base(abs))
	}
	a, err := vbf.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("abrindo %s: %w", filepath.Base(abs), err)
	}
	vbfArchives[key] = a
	return a, nil
}

// CloseVbfArchives fecha os containers abertos (chamado no shutdown do app).
func CloseVbfArchives() {
	vbfArchMu.Lock()
	defer vbfArchMu.Unlock()
	for key, a := range vbfArchives {
		_ = a.Close()
		delete(vbfArchives, key)
	}
}

// ---------------------------------------------------------------------------
// bindings de árvore
// ---------------------------------------------------------------------------

// ListVbfRoots devolve as raízes .vbf da sidebar, com o total de arquivos do
// índice (informação de cabeçalho — a árvore em si nunca é pré-extraída).
func (s *MetadataService) ListVbfRoots() ([]dto.VbfRoot, error) {
	if gameExePath() == "" {
		return []dto.VbfRoot{}, nil
	}
	found := discoverVbfRoots()
	out := make([]dto.VbfRoot, 0, len(found))
	for _, r := range found {
		if !isContentVbf(r.Name) {
			// A pasta do jogo tem outros .vbf (ex.: metamenu.vbf) que não
			// são árvore de conteúdo traduzível: a sidebar mostra UM
			// container por versão, e este não é um deles.
			continue
		}
		a, err := vbfArchiveFor(r.Path)
		if err != nil {
			// Não é um container legível: some da árvore. Só paths e
			// contagens no diagnóstico — nunca conteúdo de arquivo.
			loggingService.DiagInfo("vbf", "ignorando "+r.Name,
				map[string]any{"path": r.Path, "erro": err.Error()})
			continue
		}
		r.Entries = a.Len()
		out = append(out, r)
	}
	// Uma vez com a lista completa: mods/ é compartilhado entre FFX e
	// FFX-2, então um arquivo só é órfão quando não existe em NENHUM
	// container (checar por container marcaria como erro o arquivo que só
	// existe no outro jogo).
	s.checkVbfSync(out)
	return out, nil
}

// ---------------------------------------------------------------------------
// sincronismo mods/ ↔ container
// ---------------------------------------------------------------------------

// isContentVbf diz se o container é um dos de CONTEÚDO do jogo: um por
// versão (FFX_Data.vbf na aba FFX, FFX2_Data.vbf na FFX-2, nenhum na Last
// Mission, que reusa a árvore do FFX-2 sem container próprio). Os nomes
// casam com vbfVersionOf, então a filtragem por versão no frontend cai
// sempre em exatamente um.
func isContentVbf(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "ffx_data.vbf", "ffx2_data.vbf":
		return true
	default:
		return false
	}
}

// modsTreeIgnored são subdiretórios de mods/ que NÃO são arquivos do jogo:
// artefatos de export (edits/) e a árvore de tradução reimportável
// (translated/). Tudo o resto em mods/ é o espelho de um caminho do jogo.
var modsTreeIgnored = map[string]bool{
	"edits":      true,
	"translated": true,
	"extracted":  true,
	"reimported": true,
}

// vbfSyncSeen guarda a assinatura (lista ordenada dos órfãos) já avisada
// por container: a listagem de raízes roda a cada refresh da sidebar e o
// toast não pode repetir para o mesmo estado.
var (
	vbfSyncMu   sync.Mutex
	vbfSyncSeen = map[string]string{}
)

// checkVbfSync compara o que existe em mods/ com o ÍNDICE DE TODOS os
// containers descobertos JUNTOS e avisa quando há arquivo em mods/ sem
// contraparte em nenhum deles: o .vbf é a fonte da verdade, então mods/
// órfão é erro — aquele arquivo nunca pode ser aplicado ao jogo e passaria
// despercebido no export/import.
//
// mods/ é compartilhado entre FFX e FFX-2, por isso a comparação é com a
// união dos índices: existir no outro container já resolve.
//
// Arquivo ausente em mods/ NÃO é problema (mods/ guarda só o que já foi
// tocado). E o aviso não exibe conteúdo de arquivo: só caminhos e contagens
// (toast curto no topo direito; lista completa no diagnóstico).
func (s *MetadataService) checkVbfSync(roots []dto.VbfRoot) {
	if len(roots) == 0 {
		return
	}
	archives := make([]*vbf.Archive, 0, len(roots))
	names := make([]string, 0, len(roots))
	keys := make([]string, 0, len(roots))
	for _, r := range roots {
		a, err := vbfArchiveFor(r.Path)
		if err != nil {
			continue
		}
		archives = append(archives, a)
		names = append(names, r.Name)
		keys = append(keys, r.Path)
	}
	if len(archives) == 0 {
		return // nenhum container legível: nada a comparar
	}
	scope := strings.Join(keys, "\n")

	modsRoot := filepath.Join(common.GameFilesRoot, common.ModsFolder)
	if info, err := os.Stat(modsRoot); err != nil || !info.IsDir() {
		return // mods/ ainda não existe: nada a comparar
	}

	var missing []string
	_ = filepath.WalkDir(modsRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // um item ilegível não aborta a varredura
		}
		rel, rerr := filepath.Rel(modsRoot, p)
		if rerr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		seg := rel
		if i := strings.IndexByte(rel, '/'); i >= 0 {
			seg = rel[:i]
		}
		if d.IsDir() {
			if modsTreeIgnored[seg] {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil // arquivos de sistema, não do jogo
		}
		if !hasInAnyArchive(archives, rel) {
			missing = append(missing, rel)
		}
		return nil
	})
	if len(missing) > 0 {
		sort.Strings(missing)
	}
	signature := strings.Join(missing, "\n")

	vbfSyncMu.Lock()
	changed := vbfSyncSeen[scope] != signature
	vbfSyncSeen[scope] = signature
	vbfSyncMu.Unlock()
	if !changed {
		return // mesmo estado já comunicado
	}

	if len(missing) == 0 {
		loggingService.DiagInfo("vbf", "mods/ em dia com os containers",
			map[string]any{"containers": names})
		return
	}

	// Toast: contagem + poucos caminhos (a lista completa é do log).
	preview := missing
	const maxShown = 5
	if len(preview) > maxShown {
		preview = preview[:maxShown]
	}
	msg := fmt.Sprintf(
		"%d arquivo(s) em mods/ sem contraparte em nenhum .vbf (os .vbf são a fonte da verdade): %s",
		len(missing), strings.Join(preview, ", "),
	)
	if len(missing) > len(preview) {
		msg += fmt.Sprintf(" e mais %d", len(missing)-len(preview))
	}
	loggingService.DiagWarn("vbf", msg, map[string]any{
		"containers": names,
		"contagem":   len(missing),
		"orfãos":     missing,
	})
	if s != nil && s.notifier != nil {
		s.notifier.NotifyWarn(msg)
	}
}

// hasInAnyArchive reporta se o caminho (relativo a um container) existe no
// índice de algum dos containers carregados.
func hasInAnyArchive(archives []*vbf.Archive, rel string) bool {
	for _, a := range archives {
		if a.Has(rel) {
			return true
		}
	}
	return false
}

// ListVbfDir devolve os filhos IMEDIATOS de um diretório do .vbf.
//
// Diretórios vêm primeiro, depois arquivos, ambos em ordem alfabética — e
// cada arquivo carrega kind/id/versão QUANDO o app sabe servi-lo (o caminho
// bate com o que o próprio app resolveria em data/). Vazio = formato fora do
// escopo (áudio, vídeo, executável…): o nó continua na árvore, mas abrir avisa.
func (s *MetadataService) ListVbfDir(vbfPath, dir string) ([]dto.VbfNode, error) {
	a, err := vbfArchiveFor(vbfPath)
	if err != nil {
		return nil, err
	}
	base := filepath.Base(vbfPath)

	nodes := a.List(dir)
	out := make([]dto.VbfNode, 0, len(nodes))
	for _, n := range nodes {
		dn := dto.VbfNode{Name: n.Name, Path: n.Path, IsDir: n.IsDir, Size: n.Size}
		if !n.IsDir {
			if t, ok := matchVbfPath(base, n.Path); ok {
				dn.Kind = t.Kind
				dn.ID = t.ID
				dn.Version = t.Version.String()
				dn.Macro = t.Kind == KindMacro
				dn.Image = t.Kind == KindImages
			}
		}
		out = append(out, dn)
	}
	return out, nil
}

// ListVbfMacroChunks devolve os filhos VIRTUAIS de macrodic.dcp: o
// dicionário é UM arquivo por localização e o app o trata como um grupo com
// um filho por chunk (chunk_00, chunk_01, …) — o mesmo id da árvore data/.
func (s *MetadataService) ListVbfMacroChunks(vbfPath, macroPath string) ([]dto.VbfNode, error) {
	a, t, err := s.vbfTargetOf(vbfPath, macroPath)
	if err != nil {
		return nil, err
	}
	if t.Kind != KindMacro {
		return nil, fmt.Errorf("não é macrodic.dcp: %s", cleanVbfPath(macroPath))
	}
	if err := ensureVersionReady(t.Version); err != nil {
		return nil, err
	}

	keys, err := vbfRun(a, t, macroPath, func() ([]string, error) {
		c, cerr := builders.BuildMacroDTOFromSource(t.Version, common.SourceVbf)
		if cerr != nil {
			return nil, cerr
		}
		out := make([]string, 0, len(c))
		for k := range c {
			out = append(out, k)
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(keys)

	out := make([]dto.VbfNode, 0, len(keys))
	for _, k := range keys {
		out = append(out, dto.VbfNode{
			Name:    k,
			Path:    macroPath,
			Kind:    KindMacro,
			ID:      k,
			Version: t.Version.String(),
		})
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// bindings de conteúdo
// ---------------------------------------------------------------------------

// GetVbfTextEntry devolve a tabela (rows + coluna Original) de um arquivo de
// TEXTO aberto pelo navegador de .vbf. id é exigido só para macro (o chunk);
// nos demais kinds o id sai do próprio caminho.
//
// Diferente de GetEntry, nada aqui é data-driven: o estado ATUAL vem do
// container (mods-first) e o ORIGINAL do container puro — os dois do mesmo
// .vbf. Sem dedup de display: ele é construído sobre a árvore data/ inteira,
// que é justamente o que pode não existir por completo. Divergência de
// estrutura continua sendo reportada (paths/contagens, sem conteúdo).
func (s *MetadataService) GetVbfTextEntry(vbfPath, innerPath, id string) (dto.FileEntry, error) {
	a, t, err := s.vbfTargetOf(vbfPath, innerPath)
	if err != nil {
		return dto.FileEntry{}, err
	}
	if t.Kind == KindMacro {
		if strings.TrimSpace(id) == "" {
			return dto.FileEntry{}, fmt.Errorf("macro: chunk não informado em %s", cleanVbfPath(innerPath))
		}
		t.ID = id
	}
	if err := ensureVersionReady(t.Version); err != nil {
		return dto.FileEntry{}, err
	}

	return vbfRun(a, t, innerPath, func() (dto.FileEntry, error) {
		current, exists, err := s.loadOriginalFrom(t.Kind, t.ID, t.Version, common.SourceVbfPreferred)
		if err != nil {
			return dto.FileEntry{}, err
		}
		if !exists {
			return dto.FileEntry{}, fmt.Errorf("%s/%s indisponível no .vbf e em mods/", t.Kind, t.ID)
		}
		orig, okOrig, oerr := s.loadOriginalFrom(t.Kind, t.ID, t.Version, common.SourceVbf)
		if oerr != nil || !okOrig {
			common.LogWarning(
				"original de %s/%s/%s ausente no .vbf — entregando sem a coluna Original",
				t.Version, t.Kind, t.ID,
			)
			return current, nil
		}
		merged, diverged := withOriginal(entryRef{kind: t.Kind, id: t.ID, version: t.Version}, current, orig)
		logDivergence(diverged)
		return merged, nil
	})
}

// GetVbfImageEntry devolve a textura (.dds.phyre) pedida a partir do .vbf,
// com pré-visualização PNG em data URL. Sem índice de duplicatas: ele é
// montado varrendo data/, que pode não ter sido extraída por completo.
func (s *MetadataService) GetVbfImageEntry(vbfPath, innerPath string) (dto.ImageEntry, error) {
	a, t, err := s.vbfTargetOf(vbfPath, innerPath)
	if err != nil {
		return dto.ImageEntry{}, err
	}
	if t.Kind != KindImages {
		return dto.ImageEntry{}, fmt.Errorf("não é uma textura do app: %s", cleanVbfPath(innerPath))
	}

	return vbfRun(a, t, innerPath, func() (dto.ImageEntry, error) {
		return s.GetImageFrom(t.Kind, t.ID, t.Version, common.SourceVbf)
	})
}

// ---------------------------------------------------------------------------
// internos
// ---------------------------------------------------------------------------

// vbfTargetOf resolve o caminho interno para (kind, id, versão) e devolve o
// container aberto. ok=false = formato que o app não serve.
func (s *MetadataService) vbfTargetOf(vbfPath, innerPath string) (*vbf.Archive, vbfTarget, error) {
	a, err := vbfArchiveFor(vbfPath)
	if err != nil {
		return nil, vbfTarget{}, err
	}
	t, ok := matchVbfPath(filepath.Base(vbfPath), innerPath)
	if !ok {
		return nil, vbfTarget{}, fmt.Errorf("arquivo fora do escopo do app: %s", cleanVbfPath(innerPath))
	}
	return a, t, nil
}

// vbfRun monta o overlay do alvo (todas as localizações da entrada — eventos,
// help, macro e objects têm UM ARQUIVO POR IDIOMA) e roda fn dentro do
// escopo de leitura do .vbf. O overlay cobre o conjunto candidato (decodifica
// uma vez); o decodificador de reserva atende o que faltar, sem custo.
func vbfRun[T any](a *vbf.Archive, t vbfTarget, clicked string, fn func() (T, error)) (T, error) {
	var zero T
	overlay := make(map[string][]byte)
	for _, p := range vbfOverlayPaths(t, clicked) {
		if data, err := a.Read(p); err == nil {
			overlay[p] = data
		}
	}
	decode := func(p string) ([]byte, bool) {
		data, err := a.Read(p)
		if err != nil {
			return nil, false
		}
		return data, true
	}

	var out T
	err := common.WithVbfSourceReader(overlay, decode, func() error {
		v, err := fn()
		if err != nil {
			return err
		}
		out = v
		return nil
	})
	if err != nil {
		return zero, err
	}
	return out, nil
}
