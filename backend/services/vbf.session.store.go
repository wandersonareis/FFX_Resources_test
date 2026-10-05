package services

// STORE DE SESSÃO DO NAVEGADOR DE .vbf — O CONTAINER É SOMENTE LEITURA.
//
// Cada clique decodifica o arquivo do container e o REGISTRA aqui; o
// repetido vira ref "$hash" a cada inserção, contra o que já está na
// sessão. A âncora é a ordem de CLIQUE (HashOrder append-only): a 1ª carga
// de um ponteiro é a def — com o texto linkado na UI, editar "na cópia" é
// editar a def e replicar visualmente. Texto traduzido nunca colapsa
// (régua do DedupDisplayDTO), então a cópia traduzida em mods permanece
// literal, arquivo por arquivo.
//
// Frescura: cada entrada guarda o carimbo físico do arquivo de mods —
// binário alterado fora da sessão (tradução salva via data/, file touch)
// re-decodifica na próxima carga. Nada aqui escreve no container: o estado
// guardado é puro (pós-scope do overlay) e a mutação do apply acontece
// nele; o save vai para mods/ — o .vbf só é lido.

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/eventtable"
	"ffxresources/backend/fileFormats/helpfile"
	"ffxresources/backend/fileFormats/lockit"
	"ffxresources/backend/fileFormats/objectsfile"
)

var (
	vbfSessionMu sync.Mutex
	// vbfSessions: lower(abs caminho do container) → sessão daquele .vbf.
	vbfSessions = map[string]*vbfSession{}
)

// vbfSession é o estado acumulado de UM container na sessão: um estado por
// kind (a régua de ponteiros é por kind — a mesma que o dedupViewKinds usa
// em data/).
type vbfSession struct {
	kinds map[string]*vbfSessionKind
}

// vbfSessionKind guarda a coleção cruda acumulada do kind + a régua de
// ponteiros da sessão + o carimbo físico de cada arquivo carregado + o
// ESTADO decodificado (o objeto que o apply muta e que o save persiste em
// mods/).
type vbfSessionKind struct {
	kind    string
	version common.GameVersion
	order   *builders.HashOrder      // régua append-only + Replace (frescura)
	stored  map[string]dto.FileEntry // id → entrada merged pré-colapso (withOriginal)
	stamp   map[string]string        // id → carimbo físico do mods na carga
	estado  map[string]*vbfEstado    // id → objeto decodificado (escopo do apply)
}

// vbfEstado é o objeto decodificado do arquivo no .vbf. Qual campo existe
// depende do kind — o tipo do estado é o tipo do binário que o applier
// manipula. Nada aqui aponta para o container: o objeto veio da leitura
// (mods-first) e é mutado em memória.
type vbfEstado struct {
	// strings: events e as tabelas eventtable (o slice embutido no DTO).
	strings []*event.LocalizedFieldStringObject
	// painel: help (o painel com todas as localizações carregadas).
	painel *helpfile.HelpKeyedStringFile
	// lockitFile: o arquivo de lockit (uma entrada por id).
	lockitFile *lockit.LockitFile
	// binario: objects (binário + layout do arquivo, para o save).
	binario *vbfEstadoBinario
	// macro: o dicionário INTEIRO — base do rebuild (macro não tem
	// propagação por arquivo: o artefato é um só, por localização).
	macro dto.Collection
}

// vbfEstadoBinario é o binário de objects junto do layout: o save grava no
// pattern path do layout (→ mods/) e a key identifica o arquivo.
type vbfEstadoBinario struct {
	key    string
	layout objectsfile.FileLayout
	bin    datastore.IBinaryFile
}

// save grava o binário do arquivo tocado em mods/ pelo estado da sessão —
// o mesmo objeto decodificado no clique, já mutado pelo apply.
// Para arquivo que só existe no container, isso CRIA o arquivo em mods/
// (a extração efetiva acontece na gravação).
//
// Aqui entram os kinds cujo applier delega o save ao escopo (events,
// eventtable e help). Lockit e objects salvam dentro do próprio applier
// (o arquivo carregado é o da sessão) e macro salva pelo rebuild do
// dicionário inteiro — nenhum dos três passa por aqui.
func (k *vbfSessionKind) save(id string) error {
	e := k.estado[id]
	if e == nil {
		return fmt.Errorf("%s %s: sem estado na sessão", k.kind, id)
	}
	switch k.kind {
	case KindEvents:
		if len(e.strings) == 0 {
			return fmt.Errorf("%s %s: sem strings na sessão", k.kind, id)
		}
		return event.ExportLocalizedStringsToLocalizations(k.version, id, e.strings)
	case KindBattleText, KindCloud, KindTutorial, KindMenuMain:
		if len(e.strings) == 0 {
			return fmt.Errorf("%s %s: sem strings na sessão", k.kind, id)
		}
		bins := make([]eventtable.BinSpec, 0, 2)
		for _, rel := range eventtable.RelPaths(k.kind, id) {
			bins = append(bins, eventtable.BinSpec{Name: id, Rel: rel, Start: 0, Len: len(e.strings)})
		}
		f := &eventtable.File{
			Kind:    k.kind,
			ID:      id,
			Version: k.version,
			Strings: e.strings,
			Bins:    bins,
		}
		return f.Save()
	case KindHelp:
		if e.painel == nil {
			return fmt.Errorf("help %s: sem painel na sessão", id)
		}
		return e.painel.Save()
	}
	return fmt.Errorf("%s: kind sem fluxo de save na sessão", k.kind)
}

func vbfSessionFor(vbfPath string) *vbfSession {
	key := strings.ToLower(strings.TrimSpace(vbfPath))
	sess := vbfSessions[key]
	if sess == nil {
		sess = &vbfSession{kinds: map[string]*vbfSessionKind{}}
		vbfSessions[key] = sess
	}
	return sess
}

func (s *vbfSession) kind(kind string, version common.GameVersion) *vbfSessionKind {
	k := s.kinds[kind]
	if k == nil {
		k = &vbfSessionKind{
			kind:    kind,
			version: version,
			order:   builders.NewHashOrder(nil),
			stored:  map[string]dto.FileEntry{},
			stamp:   map[string]string{},
			estado:  map[string]*vbfEstado{},
		}
		s.kinds[kind] = k
	}
	return k
}

// vbfSessionResetAll descarta as sessões inteiras. Chamado quando o estado
// que a sessão REFLETE mudou além do carimbo por arquivo: escritas em mods/
// (apply/import invalidam clearDedupViewCache) e fechamento dos containers.
// O custo é re-decodificar na próxima carga — o comportamento original do
// fluxo por clique.
func vbfSessionResetAll() {
	vbfSessionMu.Lock()
	defer vbfSessionMu.Unlock()
	vbfSessions = map[string]*vbfSession{}
}

// vbfModsStamp mede o arquivo de mods da entrada (localização padrão): é o
// gatilho de frescura por arquivo. Vazio = sem arquivo em mods (estado puro
// do container).
func vbfModsStamp(t vbfTarget) string {
	rel, ok := originalRelPath(t.Kind, t.ID, t.Version)
	if !ok {
		return "none"
	}
	if st, ok := common.StampFileFrom(rel, common.SourceMods); ok {
		return fmt.Sprintf("%d:%d", st.ModTime.UnixNano(), st.Size)
	}
	return ""
}

// vbfSessionCached devolve a view atual da entrada se ela JÁ está na
// sessão e o arquivo de mods não mudou desde a carga — sem decodificar o
// container de novo. ok=false → o chamador decodifica (fluxo normal).
func vbfSessionCached(vbfPath string, t vbfTarget) (dto.FileEntry, bool) {
	vbfSessionMu.Lock()
	defer vbfSessionMu.Unlock()
	sess := vbfSessions[strings.ToLower(strings.TrimSpace(vbfPath))]
	if sess == nil {
		return dto.FileEntry{}, false
	}
	k, ok := sess.kinds[t.Kind]
	if !ok {
		return dto.FileEntry{}, false
	}
	entry, ok := k.stored[t.ID]
	if !ok {
		return dto.FileEntry{}, false
	}
	if k.stamp[t.ID] != vbfModsStamp(t) {
		return dto.FileEntry{}, false
	}
	base, _ := k.order.Offset(t.ID)
	// A montagem é O(rows) sobre o estado guardado — sempre fresca com a
	// régua da sessão; a entrada guardada não é mutada (cópia no dedup).
	return builders.DedupDisplayDTO(entry, base, k.order), true
}

// vbfSessionUpsert registra a entrada recém-decodificada na sessão e
// devolve a view dedupada com a régua do clique. strings é o estado
// decodificado do arquivo no motor de events/eventtable; nil = kind sem
// fluxo de save (mantém o contrato antigo dos testes).
func vbfSessionUpsert(vbfPath string, t vbfTarget, merged dto.FileEntry, strings []*event.LocalizedFieldStringObject) dto.FileEntry {
	var estado *vbfEstado
	if strings != nil {
		estado = &vbfEstado{strings: strings}
	}
	return vbfSessionUpsertEstado(vbfPath, t, merged, estado)
}

// vbfSessionUpsertEstado registra a entrada com o ESTADO completo do kind:
// o objeto decodificado que o apply vai mutar e que o save persiste em
// mods/. Entrada nova entra no FIM da ordem (Extend); entrada já registrada
// (carimbo mudou) é substituída no lugar (Replace) — o que re-registra as
// defs dela e refresca o texto dos links que apontam para cá. estado nil =
// só DTO (cache/visualização), sem fluxo de edição.
func vbfSessionUpsertEstado(vbfPath string, t vbfTarget, merged dto.FileEntry, estado *vbfEstado) dto.FileEntry {
	vbfSessionMu.Lock()
	defer vbfSessionMu.Unlock()
	k := vbfSessionFor(vbfPath).kind(t.Kind, t.Version)
	if _, had := k.stored[t.ID]; had {
		k.order.Replace(t.ID, merged.Rows)
	} else {
		k.order.Extend(t.ID, merged.Rows)
	}
	k.stored[t.ID] = merged
	k.stamp[t.ID] = vbfModsStamp(t)
	if estado != nil {
		k.estado[t.ID] = estado
	}
	base, _ := k.order.Offset(t.ID)
	view := builders.DedupDisplayDTO(merged, base, k.order)
	if collapsed := len(view.Refs); collapsed > 0 {
		common.LogVerbose("vbf %s/%s: %d repetição(ões) colapsada(s) contra a sessão", t.Kind, t.ID, collapsed)
	}
	return view
}

// vbfSessionForKind devolve a sessão/kind abertos no .vbf (lock segurado).
func vbfSessionForKind(vbfPath, kind string) (*vbfSession, *vbfSessionKind, bool) {
	sess := vbfSessions[strings.ToLower(strings.TrimSpace(vbfPath))]
	if sess == nil {
		return nil, nil, false
	}
	k, ok := sess.kinds[kind]
	if !ok {
		return nil, nil, false
	}
	return sess, k, true
}

// vbfSessionApplyScope monta o escopo do motor de apply com os arquivos
// ABERTOS na sessão do container: é por essa limitação que a propagação do
// .vbf alcança só as cópias que foram carregadas no clique.
func vbfSessionApplyScope(vbfPath, kind string, version common.GameVersion) (builders.TableApplyScope, bool) {
	vbfSessionMu.Lock()
	defer vbfSessionMu.Unlock()
	sess, k, ok := vbfSessionForKind(vbfPath, kind)
	if !ok {
		return builders.TableApplyScope{}, false
	}
	estado := k.estado
	ids := make([]string, 0, len(estado))
	for id, e := range estado {
		if e == nil || len(e.strings) == 0 {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return builders.TableApplyScope{}, false
	}
	sort.Strings(ids)
	return builders.TableApplyScope{
		Ids: ids,
		StringsFor: func(id string) []*event.LocalizedFieldStringObject {
			if e := estado[id]; e != nil {
				return e.strings
			}
			return nil
		},
		Save: func(id string) error {
			vbfSessionMu.Lock()
			defer vbfSessionMu.Unlock()
			return sess.kinds[kind].save(id)
		},
	}, true
}

// vbfSessionHelpScope é o escopo de help com os PAINÉIS abertos no clique:
// índice de propagação e save correm só sobre eles (o store de data/ não
// participa — a fonte é o container).
func vbfSessionHelpScope(vbfPath string) (builders.HelpApplyScope, bool) {
	vbfSessionMu.Lock()
	defer vbfSessionMu.Unlock()
	sess, k, ok := vbfSessionForKind(vbfPath, KindHelp)
	if !ok {
		return builders.HelpApplyScope{}, false
	}
	estado := k.estado
	ids := make([]string, 0, len(estado))
	for id, e := range estado {
		if e == nil || e.painel == nil {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return builders.HelpApplyScope{}, false
	}
	sort.Strings(ids)
	return builders.HelpApplyScope{
		Ids: ids,
		Panel: func(name string) *helpfile.HelpKeyedStringFile {
			if e := estado[name]; e != nil {
				return e.painel
			}
			return nil
		},
		Save: func(name string) error {
			vbfSessionMu.Lock()
			defer vbfSessionMu.Unlock()
			return sess.kinds[KindHelp].save(name)
		},
	}, true
}

// vbfSessionLockitScope devolve o arquivo de lockit aberto no clique.
func vbfSessionLockitScope(vbfPath string) (builders.LockitApplyScope, bool) {
	vbfSessionMu.Lock()
	defer vbfSessionMu.Unlock()
	_, k, ok := vbfSessionForKind(vbfPath, KindLockit)
	if !ok {
		return builders.LockitApplyScope{}, false
	}
	estado := k.estado
	return builders.LockitApplyScope{
		Load: func(id string) (*lockit.LockitFile, error) {
			if e := estado[id]; e != nil && e.lockitFile != nil {
				return e.lockitFile, nil
			}
			return nil, fmt.Errorf("lockit %s: arquivo não aberto na sessão do .vbf", id)
		},
	}, true
}

// vbfSessionObjectsLoader devolve o loader de binário de objects que atende
// pela sessão: o apply só alcança os arquivos abertos no clique, e o objeto
// é o MESMO que foi decodificado (mutação em memória → save em mods/).
func vbfSessionObjectsLoader(vbfPath string) (func(objectsfile.FileLayout) (datastore.IBinaryFile, error), bool) {
	vbfSessionMu.Lock()
	defer vbfSessionMu.Unlock()
	_, k, ok := vbfSessionForKind(vbfPath, KindObjects)
	if !ok {
		return nil, false
	}
	estado := k.estado
	return func(layout objectsfile.FileLayout) (datastore.IBinaryFile, error) {
		// id de objects = stem do arquivo (vbf_match), e o layout tem
		// FileName = <id>.bin — o mesmo identificador do DTO.
		id := strings.TrimSuffix(layout.FileName, ".bin")
		if e := estado[id]; e != nil && e.binario != nil {
			return e.binario.bin, nil
		}
		return nil, fmt.Errorf("objects %s: arquivo não aberto na sessão do .vbf", id)
	}, true
}

// vbfSessionMacro devolve o dicionário completo guardado no clique — a base
// do rebuild (macro não tem propagação por arquivo: o artefato é um só por
// localização, então o escopo é o dicionário inteiro).
func vbfSessionMacro(vbfPath string) (dto.Collection, bool) {
	vbfSessionMu.Lock()
	defer vbfSessionMu.Unlock()
	_, k, ok := vbfSessionForKind(vbfPath, KindMacro)
	if !ok {
		return nil, false
	}
	for _, e := range k.estado {
		if e != nil && e.macro != nil {
			return e.macro, true
		}
	}
	return nil, false
}
