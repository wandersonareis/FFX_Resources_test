import { createStore } from '@tanstack/store';
import { dto } from '@/wailsjs/go/models';
import { KIND_LABELS, type EntryKind } from '@/lib/ffx/display-names';
import { eventGroupLabel, shortenedOf } from '@/lib/ffx/event-group-names';
import type { GameVersionId } from '@/lib/ffx/game-version';
import {
  type EntryRow,
  entryKindsFor,
  loadEntry,
  loadKindEntries,
} from '@/lib/ffx/tree-data';
import { editDraft } from '@/lib/ffx/edit-draft';
import { sendErrorNotification } from '@/lib/ffx/error-handler';
import { isRefText } from '@/lib/ffx/hash-ref';
import { SOURCE_LANG } from '@/lib/ffx/save-all';
import type { SideNode } from './types';

/** Estado de carga de um item principal da árvore. */
export type KindLoadStatus = 'loading' | 'ready' | 'error';

export interface EntryViewState {
  roots: SideNode[];
  expanded: Set<string>;
  activeKind: EntryKind;
  selectedEntry: EntryRow | null;
  rows: dto.TextRow[];
  loading: boolean;
  /** Estado de carga por item principal (árvore parcial). */
  kindStatus: Record<EntryKind, KindLoadStatus>;
  /** Geração do reload — patches de geração passada são descartados. */
  generation: number;
  /** Linha aberta no modal do editor. */
  translationRow: dto.TextRow | null;
  dialogOpen: boolean;
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
}

export interface EntryView {
  readonly version: GameVersionId;
  readonly store: ReturnType<typeof createEntryStore>;
  readonly actions: EntryActions;
}

/** Agrupa os arquivos de events por shortened (grupo nomeado + contagem). */
function eventGroups(version: GameVersionId, entries: EntryRow[]): SideNode[] {
  const groups = new Map<string, EntryRow[]>();
  for (const entry of entries) {
    const short = shortenedOf(entry.id);
    const list = groups.get(short) ?? [];
    list.push(entry);
    groups.set(short, list);
  }
  const nodes: SideNode[] = [];
  for (const short of [...groups.keys()].sort()) {
    const files = groups.get(short) ?? [];
    nodes.push({
      id: `group:events:${short}`,
      label: `${eventGroupLabel(version, short)} (${files.length})`,
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
  const kinds = entryKindsFor(version);
  return createStore<EntryViewState>({
    roots: kinds.map((kind) => loadingRootNode(kind)),
    expanded: new Set<string>(),
    activeKind: 'events',
    selectedEntry: null,
    rows: [],
    loading: true,
    kindStatus: Object.fromEntries(kinds.map((k) => [k, 'loading' as const])) as Record<
      EntryKind,
      KindLoadStatus
    >,
    generation: 0,
    translationRow: null,
    dialogOpen: false,
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

  const selectEntry = async (entry: EntryRow): Promise<void> => {
    patch({ loading: true });
    try {
      patch({ activeKind: entry.kind, selectedEntry: entry });
      const full = await loadEntry(entry.kind, entry.id, version);
      editDraft.setBase(version, entry.kind, entry.id, full);
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
      patch({ selectedEntry: null, rows: [] });
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
    const kinds = entryKindsFor(version);
    await Promise.allSettled(
      kinds.map(async (kind) => {
        try {
          const entries = await loadKindEntries(kind, version);
          if (gen !== generationCounter.current) return;
          store.setState((prev) => {
            const byKind = new Map(prev.roots.map((r) => [r.kind!, r] as const));
            byKind.set(kind, buildKindNode(kind, entries));
            const roots = kinds
              .map((k) => byKind.get(k))
              .filter((n): n is SideNode => Boolean(n));
            // Expansão em UNION: o que já está aberto não colapsa.
            const expanded = new Set(prev.expanded);
            expanded.add(`kind:${kind}`);
            for (const child of byKind.get(kind)?.children ?? []) {
              expanded.add(child.id);
            }
            return {
              ...prev,
              roots,
              expanded,
              kindStatus: { ...prev.kindStatus, [kind]: 'ready' as const },
            };
          });
          // Revalida a seleção quando o KIND dela completa (não no fim de
          // tudo): entrada removida por import limpa a tabela.
          const current = store.state.selectedEntry;
          if (current && current.kind === kind) {
            const found = entries.find((e) => e.id === current.id);
            if (!found) {
              patch({ selectedEntry: null, rows: [] });
            } else {
              await selectEntry(found);
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

  /** Carga da aba: marca todos os principais como loading e carrega. */
  const reload = async (): Promise<void> => {
    const gen = ++generationCounter.current;
    const kinds = entryKindsFor(version);
    patch({ loading: true, kindStatus: loadingStatuses(kinds) });
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
    const kinds = entryKindsFor(version);
    patch({
      roots: kinds.map((kind) => loadingRootNode(kind)),
      expanded: new Set<string>(),
      selectedEntry: null,
      rows: [],
      kindStatus: loadingStatuses(kinds),
      generation: gen,
      loading: true,
    });
    await loadAllKinds(gen);
    if (gen === generationCounter.current) {
      patch({ loading: false });
    }
  };

  const actions: EntryActions = {
    reload,
    coldReload,
    selectEntry,
    selectNode: async (node) => {
      if (node.entry && node.kind) {
        await selectEntry(node.entry);
      } else if (node.kind) {
        patch({ activeKind: node.kind });
      }
    },
    toggleNode: (node) => {
      store.setState((prev) => {
        const next = new Set(prev.expanded);
        if (next.has(node.id)) next.delete(node.id);
        else next.add(node.id);
        return { ...prev, expanded: next };
      });
    },
    requestTableFocus: () => patch({ pendingTableFocus: true }),
    consumeTableFocus: () => patch({ pendingTableFocus: false }),
    openDialog: (row) => patch({ translationRow: row, dialogOpen: true }),
    navigateRow: (direction, value) => {
      const { selectedEntry: entry, translationRow: row } = store.state;
      if (value !== undefined && entry && row) {
        editDraft.setCell(version, entry.kind, entry.id, row, SOURCE_LANG, value);
        patch({ rows: [...store.state.rows] });
      }
      const idx = store.state.rows.findIndex((r) => r.index === row?.index);
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
  };

  return { version, store, actions };
}
