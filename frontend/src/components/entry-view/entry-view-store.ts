import { createStore } from '@tanstack/store';
import { dto } from '@/wailsjs/go/models';
import {
  KIND_LABELS,
  IMAGE_TREE_PREFIX,
  resolveEntryLabel,
  type EntryKind,
} from '@/lib/ffx/display-names';
import { resolveEventGroup, shortenedOf } from '@/lib/ffx/event-group-names';
import type { GameVersionId } from '@/lib/ffx/game-version';
import {
  type EntryRow,
  allKindsFor,
  imageExists,
  loadImage,
  loadEntry,
  loadKindEntries,
} from '@/lib/ffx/tree-data';
import { editDraft, rowKey } from '@/lib/ffx/edit-draft';
import { entryProgress, type EntryProgress } from '@/lib/ffx/entry-progress';
import {
  sendErrorNotification,
  sendErrorNotificationWithMessage,
} from '@/lib/ffx/error-handler';
import { isRefText } from '@/lib/ffx/hash-ref';
import { SOURCE_LANG } from '@/lib/ffx/save-all';
import {
  invalidateVbfCache,
  loadVbfDir,
  loadVbfEntry,
  loadVbfImage,
  loadVbfMacroChunks,
  loadVbfRoots,
  type VbfNode,
} from '@/lib/ffx/vbf';
import type { SideNode } from './types';

/** Estado de carga de um item principal da árvore. */
export type KindLoadStatus = 'loading' | 'ready' | 'error';

/**
 * Ação de imagem em curso — extrair / replicar / deletar sobre um id
 * CLICADO na árvore ou no painel (que pode não ser a entry selecionada).
 *
 * Vive no store porque o diálogo é renderizado uma vez só, na aba: menu e
 * painel só abrem. `duplicates` null = "buscar no backend" (o menu não sabe
 * as cópias do nó sob o cursor); `label` é o nome exibido no título.
 */
export type ImageActionState = {
  type: 'extract' | 'replicate' | 'delete';
  id: string;
  label: string;
  duplicates: dto.ImageDuplicate[] | null;
};

export interface EntryViewState {
  roots: SideNode[];
  /**
   * Raízes dos containers .vbf descobertos perto do executável do jogo
   * (somente leitura). Vindas do backend, não dos kinds de data/ — por isso
   * um campo próprio: os kinds são montados e unidos em `roots`, e o
   * carregamento deles é independente desta lista.
   */
  vbfRoots: SideNode[];
  expanded: Set<string>;
  activeKind: EntryKind;
  selectedEntry: EntryRow | null;
  rows: dto.TextRow[];
  /**
   * Textura aberta (kind=images). null = nenhum / kind de texto.
   * Guardada aqui junto da seleção para o painel de imagem não precisar de
   * estado próprio (e para sobreviver à troca de componente).
   */
  image: dto.ImageEntry | null;
  loading: boolean;
  /**
   * Progresso de tradução da entrada aberta (por row, todos os formatos):
   * FIXA na abertura do arquivo — não acompanha rascunho em edição.
   * null = sem arquivo aberto.
   */
  progress: EntryProgress | null;
  /** Estado de carga por item principal (árvore parcial). */
  kindStatus: Record<EntryKind, KindLoadStatus>;
  /** Geração do reload — patches de geração passada são descartados. */
  generation: number;
  /** Linha aberta no modal do editor. */
  translationRow: dto.TextRow | null;
  dialogOpen: boolean;
  /**
   * Diálogo de ação de imagem aberto (extrair/replicar/deletar) ou null.
   * Um só por aba — menu e painel apenas abrem.
   */
  imageAction: ImageActionState | null;
  /**
   * Pedido de foco na 1ª linha da tabela (ArrowRight na folha da árvore).
   * Vive no store (igual ao antigo pendingTableFocusRef) para sobreviver ao
   * unmount/remount da tabela e ser consumido quando as rows chegarem.
   */
  pendingTableFocus: boolean;
}

export interface EntryActions {
  /** Recarrega a árvore e revalida a entrada selecionada. */
  reload(): Promise<void>;
  /** Reload frio (botão): limpa a árvore para placeholders e recarrega tudo. */
  coldReload(): Promise<void>;
  /** Abre um arquivo na tabela (carrega rows e registra a base do rascunho). */
  selectEntry(entry: EntryRow): Promise<void>;
  /** Clique/Enter num nó da árvore: folha abre, raiz/kind só troca o kind. */
  selectNode(node: SideNode): Promise<void>;
  /** Alterna expandir/colapsar grupo ou raiz. */
  toggleNode(node: SideNode): void;
  /** ArrowRight na folha: pede foco na 1ª linha após o arquivo carregar. */
  requestTableFocus(): void;
  /** Tabela focou a 1ª linha: limpa o pedido pendente. */
  consumeTableFocus(): void;
  /** Abre o modal do editor para a linha clicada/teclada na tabela. */
  openDialog(row: dto.TextRow): void;
  /** Navegação ←/→ do modal sem fechar (value aplicado como rascunho). */
  navigateRow(direction: 'prev' | 'next', value?: string): void;
  /** Fechamento do modal (value != undefined aplica a edição). */
  commitRow(value?: string): void;
  /** Abre o diálogo de ação de imagem (extrair/replicar/deletar). */
  openImageAction(action: ImageActionState): void;
  /** Fecha o diálogo de ação de imagem (cancelar/ao concluir). */
  closeImageAction(): void;
}

export interface EntryView {
  readonly version: GameVersionId;
  readonly store: ReturnType<typeof createEntryStore>;
  readonly actions: EntryActions;
}

/**
 * Agrupa os arquivos de events por fragmento do eventID. Fragmentos com nome
 * no dicionário viram nós próprios; sem nome, fundem-se no primeiro fragmento
 * nomeado do mesmo shortened (azmm entra em azit = Al Bhed Home). Órfão sem
 * irmão nomeado exibe o próprio fragmento, igual ao nome da pasta no disco.
 */
function eventGroups(version: GameVersionId, entries: EntryRow[]): SideNode[] {
  const byFragment = new Map<string, EntryRow[]>();
  for (const entry of entries) {
    const fragment = shortenedOf(entry.id);
    const list = byFragment.get(fragment) ?? [];
    list.push(entry);
    byFragment.set(fragment, list);
  }
  const fragments = [...byFragment.keys()];

  // Fragmentos do mesmo shortened podem fundir-se em um alvo comum.
  const fragmentsByTarget = new Map<string, string[]>();
  const targetNames = new Map<string, string>();
  for (const fragment of fragments) {
    const { target, name } = resolveEventGroup(version, fragment, fragments);
    if (name) targetNames.set(target, name);
    const list = fragmentsByTarget.get(target) ?? [];
    list.push(fragment);
    fragmentsByTarget.set(target, list);
  }

  const nodes: SideNode[] = [];
  for (const [target, frags] of [...fragmentsByTarget.entries()].sort(([a], [b]) =>
    a.localeCompare(b),
  )) {
    const files = frags.flatMap((fragment) => byFragment.get(fragment) ?? []);
    const name = targetNames.get(target);
    nodes.push({
      id: `group:events:${target}`,
      label: name
        ? `${name} (${target.slice(0, 2)}) - ${files.length}`
        : `${target} - ${files.length}`,
      kind: 'events' as EntryKind,
      children: files.map((entry) => ({
        id: `leaf:events:${entry.id}`,
        label: entry.label,
        kind: entry.kind,
        entry,
      })),
    });
  }
  return nodes;
}

function createEntryStore(version: GameVersionId) {
  const kinds = allKindsFor(version);
  return createStore<EntryViewState>({
    roots: kinds.map((kind) => loadingRootNode(kind)),
    vbfRoots: [],
    expanded: new Set<string>(),
    activeKind: 'events',
    selectedEntry: null,
    rows: [],
    image: null,
    loading: true,
    progress: null,
    kindStatus: Object.fromEntries(kinds.map((k) => [k, 'loading' as const])) as Record<
      EntryKind,
      KindLoadStatus
    >,
    generation: 0,
    translationRow: null,
    dialogOpen: false,
    imageAction: null,
    pendingTableFocus: false,
  });
}

/** Raiz de kind em carga: label sem contagem + spinner no lugar dos filhos. */
function loadingRootNode(kind: EntryKind): SideNode {
  return { id: `kind:${kind}`, label: KIND_LABELS[kind], kind, loading: true };
}

function loadingStatuses(kinds: EntryKind[]): Record<EntryKind, KindLoadStatus> {
  return Object.fromEntries(
    kinds.map((k) => [k, 'loading' as const])
  ) as Record<EntryKind, KindLoadStatus>;
}

/**
 * Id de um nó do navegador de .vbf. Caminho interno identifica arquivo e
 * diretório dentro do container; `id` extra cobre os chunks virtuais do
 * macrodic (todos com o mesmo caminho do dicionário).
 */
function vbfNodeId(root: string, node: Pick<VbfNode, 'path' | 'id'>): string {
  return `vbf:${root}|${node.path}|${node.id ?? ''}`;
}

/**
 * Converte um nó do índice do .vbf em SideNode.
 *
 * - diretório → grupo expansível (filhos vão pro backend na expansão);
 * - macrodic.dcp → grupo com filhos chunk_XX (o app trata dicionário assim);
 * - arquivo com kind servível (evento, lockit, textura, …) → folha com
 *   `entry.vbf`, que é o que troca a carga para os bindings VBF;
 * - o resto (áudio, vídeo, .exe…) → folha sem entry: o clique avisa que o
 *   formato está fora do escopo do app, sem tentar decodificar.
 *
 * Nada aqui escreve no container: é só o espelho do índice.
 */
function vbfSideNode(root: string, node: VbfNode): SideNode {
  const id = vbfNodeId(root, node);
  if (node.isDir) {
    return {
      id,
      label: node.name,
      vbf: true,
      vbfRoot: root,
      vbfPath: node.path,
      expandable: true,
    };
  }
  if (node.macro) {
    return {
      id,
      label: node.name,
      vbf: true,
      vbfRoot: root,
      vbfPath: node.path,
      macro: true,
      expandable: true,
    };
  }
  if (!node.kind) {
    return { id, label: node.name, vbf: true, unsupported: true };
  }
  const kind = node.kind as EntryKind;
  const label = resolveEntryLabel(kind, node.id);
  return {
    id,
    label,
    kind,
    vbf: true,
    entry: {
      kind,
      id: node.id,
      key: node.id,
      label,
      vbf: { root, path: node.path },
    },
  };
}

/** Estado "sem arquivo aberto" — limpa tabela, progresso e imagem juntos. */
const NO_SELECTION: Partial<EntryViewState> = {
  selectedEntry: null,
  rows: [],
  progress: null,
  image: null,
};

/**
 * Árvore de imagens: categoria (1º segmento após gamedata/ps3data) →
 * diretório da textura (o trecho antes de /d3d11, que é só a variante de
 * renderizador) → folha com o nome do arquivo. Os ids são caminhos longos e
 * repetitivos; agrupar assim deixa a árvore navegável sem perder o caminho.
 */
function imageTreeNodes(entries: EntryRow[]): SideNode[] {
  const byCategory = new Map<string, Map<string, EntryRow[]>>();
  for (const entry of entries) {
    const rel = entry.id.startsWith(IMAGE_TREE_PREFIX)
      ? entry.id.slice(IMAGE_TREE_PREFIX.length)
      : entry.id;
    const firstSlash = rel.indexOf('/');
    const category = firstSlash < 0 ? rel : rel.slice(0, firstSlash);
    const rest = firstSlash < 0 ? '' : rel.slice(firstSlash + 1);
    const lastSlash = rest.lastIndexOf('/');
    const dir =
      lastSlash < 0
        ? ''
        : rest.slice(0, lastSlash).replace(/(^|\/)d3d11$/, '');
    const byDir = byCategory.get(category) ?? new Map<string, EntryRow[]>();
    const list = byDir.get(dir) ?? [];
    list.push(entry);
    byDir.set(dir, list);
    byCategory.set(category, byDir);
  }

  const nodes: SideNode[] = [];
  for (const [category, byDir] of [...byCategory.entries()].sort(([a], [b]) =>
    a.localeCompare(b)
  )) {
    const count = [...byDir.values()].reduce((n, list) => n + list.length, 0);
    nodes.push({
      id: `group:images:${category}`,
      label: `${category} (${count})`,
      kind: 'images',
      children: [...byDir.entries()]
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([dir, list]) => ({
          id: `group:images:${category}/${dir}`,
          label: dir ? `${dir} (${list.length})` : `raiz (${list.length})`,
          kind: 'images' as EntryKind,
          children: list.map((entry) => ({
            id: `leaf:images:${entry.id}`,
            label: entry.label,
            kind: entry.kind,
            entry,
          })),
        })),
    });
  }
  return nodes;
}

/**
 * Store por aba de versão (TanStack Store): concentra o estado compartilhado
 * entre árvore, tabela e diálogo de tradução, eliminando prop drilling entre
 * os componentes de entry-view. Uma instância por GameVersionTab — trocar de
 * aba não preserva seleção (como no original com useState).
 */
export function createEntryView(version: GameVersionId): EntryView {
  const store = createEntryStore(version);

  const patch = (partial: Partial<EntryViewState>): void => {
    store.setState((prev) => ({ ...prev, ...partial }));
  };

  /**
   * Atualiza UM nó da árvore em qualquer profundidade (raízes de kind e
   * raízes VBF são caminhos separados, por isso os dois no mesmo passe).
   */
  const patchNode = (id: string, partial: Partial<SideNode>): void => {
    const walk = (list: SideNode[]): SideNode[] =>
      list.map((node) =>
        node.id === id
          ? { ...node, ...partial }
          : node.children
            ? { ...node, children: walk(node.children) }
            : node
      );
    store.setState((prev) => ({
      ...prev,
      roots: walk(prev.roots),
      vbfRoots: walk(prev.vbfRoots),
    }));
  };

  /**
   * Expansão preguiçosa de um nó do .vbf: o container nunca é indexado de
   * uma vez — só o diretório clicado é consultado (e o macrodic quebra em
   * chunk_XX). Guardado por id para um duplo clique não disparar duas buscas.
   */
  const loadingVbf = new Set<string>();
  const loadVbfChildren = async (node: SideNode): Promise<void> => {
    const root = node.vbfRoot ?? node.entry?.vbf?.root;
    if (!root || loadingVbf.has(node.id)) return;
    loadingVbf.add(node.id);
    patchNode(node.id, { loading: true });
    try {
      const path = node.vbfPath ?? '';
      const list = node.macro
        ? await loadVbfMacroChunks(root, path)
        : await loadVbfDir(root, path);
      const children = list.map((child) => vbfSideNode(root, child));
      patchNode(node.id, {
        children,
        loading: false,
        expandable: children.length > 0,
      });
      if (children.length > 0) {
        store.setState((prev) => ({
          ...prev,
          expanded: new Set(prev.expanded).add(node.id),
        }));
      }
    } catch (error) {
      patchNode(node.id, { loading: false, expandable: false });
      sendErrorNotification(error);
    } finally {
      loadingVbf.delete(node.id);
    }
  };

  const selectEntry = async (entry: EntryRow): Promise<void> => {
    patch({ loading: true });
    try {
      patch({ activeKind: entry.kind, selectedEntry: entry });
      if (entry.kind === 'images') {
        // Textura: sem rows/progresso — só a imagem (data URL) + metadados.
        const image = entry.vbf
          ? await loadVbfImage(entry.vbf.root, entry.vbf.path)
          : await loadImage(entry.id, version);
        patch({ image, rows: [], progress: null });
        return;
      }
      patch({ image: null });
      const full = entry.vbf
        ? await loadVbfEntry(entry.vbf.root, entry.vbf.path, entry.id)
        : await loadEntry(entry.kind, entry.id, version);
      // Entrada aberta a partir do .vbf é VISUALIZAÇÃO: não alimenta o
      // rascunho de edição (salvar gravaria em mods/ contra um original que
      // pode nem estar extraído em data/) e nem registra a base do rascunho.
      if (!entry.vbf) {
        editDraft.setBase(version, entry.kind, entry.id, full);
      }
      // Progresso por row: FIXO na abertura (não acompanha rascunho).
      patch({ progress: entryProgress(full) });
      // Dedup: refs "$hash" (repetições idênticas) ficam fora da tabela —
      // o tradutor traduz cada texto uma vez. A base do rascunho guarda a
      // entry COMPLETA; o backend resolve refs e propaga o texto editado
      // para todas as cópias no salvar (UI e import).
      const rows = (full.rows ?? []).filter(
        (r) => !isRefText(r.text?.[SOURCE_LANG], r.hash?.[SOURCE_LANG])
      );
      patch({ rows });
    } catch (error) {
      sendErrorNotification(error);
      patch({ ...NO_SELECTION });
    } finally {
      patch({ loading: false });
    }
  };

  const buildKindNode = (kind: EntryKind, entries: EntryRow[]): SideNode => ({
    id: `kind:${kind}`,
    label: `${KIND_LABELS[kind]} (${entries.length})`,
    kind,
    children:
      kind === 'events'
        ? eventGroups(version, entries)
        : kind === 'images'
          ? imageTreeNodes(entries)
          : entries.map((entry) => ({
              id: `leaf:${kind}:${entry.id}`,
              label: entry.label,
              kind: entry.kind,
              entry,
            })),
  });

  /** Contador de geração: patches de geração passada são descartados. */
  const generationCounter = { current: 0 };

  /**
   * Carga da árvore em PARTES para uma geração: todos os kinds disparam em
   * paralelo e cada um entra na árvore assim que chega (upsert, na ordem
   * canônica dos kinds — nós não trocam de posição). A árvore NUNCA colapsa
   * o que já está aberto (expansão em union).
   */
  const loadAllKinds = async (gen: number): Promise<void> => {
    const kinds = allKindsFor(version);
    await Promise.allSettled(
      kinds.map(async (kind) => {
        try {
          const entries = await loadKindEntries(kind, version);
          if (gen !== generationCounter.current) return;
          store.setState((prev) => {
            const byKind = new Map(prev.roots.map((r) => [r.kind!, r] as const));
            // Kind sem nada (ex.: images numa versão sem texturas) não vira
            // raiz vazia "Imagens (0)" — a árvore só mostra o que existe.
            if (entries.length === 0) byKind.delete(kind);
            else byKind.set(kind, buildKindNode(kind, entries));
            const roots = kinds
              .map((k) => byKind.get(k))
              .filter((n): n is SideNode => Boolean(n));
            // A árvore nasce FECHADA, inclusive a de data/: nada é expandido
            // automaticamente. O expanded que já está aqui guarda só o que
            // o USUÁRIO abriu, e sobrevive ao reload (união) — quem quer ver
            // os filhos expande o nó.
            return {
              ...prev,
              roots,
              kindStatus: { ...prev.kindStatus, [kind]: 'ready' as const },
            };
          });
          // Revalida a seleção quando o KIND dela completa (não no fim de
          // tudo): entrada removida por import limpa a tabela. Seleção que
          // veio do .vbf fica de fora — a árvore de data/ não tem como
          // revalidá-la (o arquivo pode nem estar extraído ali); se o
          // container sumir, a própria carga da entrada avisa.
          const current = store.state.selectedEntry;
          if (current && !current.vbf && current.kind === kind) {
            const found = entries.find((e) => e.id === current.id);
            if (found) {
              await selectEntry(found);
            } else if (
              kind === 'images' &&
              entries.length > 0 &&
              (await imageExists(current.id, version))
            ) {
              // Cópia oculta do grupo: a árvore só guarda o representante,
              // mas o id continua servível (a lista "Repetidas" navega até
              // ele) — quem decide se ainda existe é a carga, não a lista.
              await selectEntry(current);
            } else {
              // Não existe mais (ex.: imagem + cópias apagadas com escopo
              // both): limpa em SILÊNCIO — o Resolve erroaria "não
              // encontrada em data/ nem em mods/" para o id que o delete
              // acabou de remover com sucesso.
              patch({ ...NO_SELECTION });
            }
          }
        } catch (error) {
          if (gen !== generationCounter.current) return;
          store.setState((prev) => ({
            ...prev,
            kindStatus: { ...prev.kindStatus, [kind]: 'error' as const },
          }));
          sendErrorNotification(error);
        }
      })
    );
  };

  /** Tamanho humano do container — só informativo na raiz da árvore. */
  const formatVbfSize = (bytes: number): string => {
    if (!Number.isFinite(bytes) || bytes <= 0) return '';
    const gb = bytes / 1024 ** 3;
    if (gb >= 1) return `${gb.toFixed(1)} GB`;
    const mb = bytes / 1024 ** 2;
    if (mb >= 1) return `${mb.toFixed(0)} MB`;
    return `${bytes} B`;
  };

  /**
   * Raízes dos containers .vbf descobertos perto do executável do jogo. O
   * backend varre a pasta uma vez (promessa compartilhada em vbf.ts) e cada
   * aba só espelha a lista na própria árvore. Sem executível configurado ele
   * devolve lista vazia — a árvore de data/ continua normal.
   */
  const loadVbfRootNodes = async (gen: number): Promise<void> => {
    try {
      const all = await loadVbfRoots();
      if (gen !== generationCounter.current) return;
      // UM container por versão: FFX_Data.vbf na aba FFX, FFX2_Data.vbf na
      // FFX-2 e NENHUM na Last Mission (que reusa a árvore do FFX-2 sem
      // container próprio). Nomes sem a extensão — a seção acima da árvore
      // já diz que é .vbf.
      const list = all.filter((root) => root.version === version);
      patch({
        vbfRoots: list.map((root) => ({
          // Mesmo formato de vbfNodeId com path/id vazios = a RAIZ do
          // container (é o que a expansão usa como diretório inicial).
          id: vbfNodeId(root.path, { path: '', id: '' }),
          label: `${root.name.replace(/\.vbf$/i, '')} · ${formatVbfSize(root.size)}`,
          vbf: true,
          vbfRoot: root.path,
          vbfPath: '',
          expandable: true,
        })),
      });
    } catch (error) {
      if (gen !== generationCounter.current) return;
      patch({ vbfRoots: [] });
      sendErrorNotification(error);
    }
  };

  /** Carga da aba: marca todos os principais como loading e carrega. */
  const reload = async (): Promise<void> => {
    const gen = ++generationCounter.current;
    const kinds = allKindsFor(version);
    patch({ loading: true, kindStatus: loadingStatuses(kinds) });
    void loadVbfRootNodes(gen);
    await loadAllKinds(gen);
    if (gen === generationCounter.current) {
      patch({ loading: false });
    }
  };

  /**
   * Reload frio (botão da sidebar): a aba perde os arquivos — raízes voltam
   * aos placeholders de carga e a carga recomeça do zero, com o mesmo
   * efeito da carga inicial (kinds em paralelo, expandidos ao completar).
   */
  const coldReload = async (): Promise<void> => {
    const gen = ++generationCounter.current;
    const kinds = allKindsFor(version);
    // Recomeça do zero também o lado do .vbf: cache de diretórios/entradas
    // descartado e as raízes re-descobertas (executável do jogo pode ter
    // mudado no diálogo de configuração).
    invalidateVbfCache();
    patch({
      roots: kinds.map((kind) => loadingRootNode(kind)),
      vbfRoots: [],
      expanded: new Set<string>(),
      selectedEntry: null,
      rows: [],
      image: null,
      kindStatus: loadingStatuses(kinds),
      generation: gen,
      loading: true,
    });
    void loadVbfRootNodes(gen);
    await loadAllKinds(gen);
    if (gen === generationCounter.current) {
      patch({ loading: false });
    }
  };

  /**
   * Alterna expandir/colapsar. Um nó do .vbf ainda sem filhos busca no
   * backend ANTES de abrir (expansão preguiçosa, com spinner no chevron) —
   * o container nunca é indexado de uma vez.
   */
  const toggleNode = (node: SideNode): void => {
    if (node.vbf && node.expandable && !node.children) {
      void loadVbfChildren(node);
      return;
    }
    store.setState((prev) => {
      const next = new Set(prev.expanded);
      if (next.has(node.id)) next.delete(node.id);
      else next.add(node.id);
      return { ...prev, expanded: next };
    });
  };

  const actions: EntryActions = {
    reload,
    coldReload,
    selectEntry,
    selectNode: async (node) => {
      if (node.vbf && node.unsupported) {
        // Fora do escopo do app (áudio, vídeo, .exe…): avisa SEM tentar
        // decodificar — só o caminho do arquivo aparece, nunca conteúdo.
        sendErrorNotificationWithMessage(
          `Formato fora do escopo do app: ${node.label}`
        );
        return;
      }
      if (node.vbf && !node.entry) {
        // Raiz/diretório do .vbf: o clique alterna a expansão.
        toggleNode(node);
        return;
      }
      if (node.entry && node.kind) {
        await selectEntry(node.entry);
      } else if (node.kind) {
        patch({ activeKind: node.kind });
      }
    },
    toggleNode,
    requestTableFocus: () => patch({ pendingTableFocus: true }),
    consumeTableFocus: () => patch({ pendingTableFocus: false }),
    openDialog: (row) => {
      // Entrada aberta pelo .vbf é somente leitura: não existe edição para
      // abrir (o salvar sairia de mods/ contra um original que pode nem estar
      // extraído em data/).
      if (store.state.selectedEntry?.vbf) return;
      patch({ translationRow: row, dialogOpen: true });
    },
    navigateRow: (direction, value) => {
      const { selectedEntry: entry, translationRow: row } = store.state;
      if (value !== undefined && entry && row) {
        editDraft.setCell(version, entry.kind, entry.id, row, SOURCE_LANG, value);
        patch({ rows: [...store.state.rows] });
      }
      const idx = store.state.rows.findIndex(
        (r) => rowKey(r) === (row ? rowKey(row) : null)
      );
      const next =
        store.state.rows[idx + (direction === 'next' ? 1 : -1)];
      if (next) patch({ translationRow: next });
    },
    commitRow: (value) => {
      const { selectedEntry: entry, translationRow: row } = store.state;
      if (value !== undefined && entry && row) {
        editDraft.setCell(version, entry.kind, entry.id, row, SOURCE_LANG, value);
        patch({ rows: [...store.state.rows] });
      }
      patch({ dialogOpen: false, translationRow: null });
    },
    openImageAction: (action) => {
      // Ações de imagem agem sobre data/ + mods/: a partir do .vbf o id
      // pode nem existir lá, e o container não é alvo de escrita.
      if (store.state.selectedEntry?.vbf) {
        sendErrorNotificationWithMessage(
          'Ações de imagem estão disponíveis na árvore de data/ — o .vbf é somente leitura.'
        );
        return;
      }
      patch({ imageAction: action });
    },
    closeImageAction: () => patch({ imageAction: null }),
  };

  return { version, store, actions };
}
