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
import { SOURCE_LANG } from '@/lib/ffx/save-all';
import type { SideNode } from './types';

export interface EntryViewState {
  roots: SideNode[];
  expanded: Set<string>;
  activeKind: EntryKind;
  selectedEntry: EntryRow | null;
  rows: dto.TextRow[];
  loading: boolean;
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

function createEntryStore() {
  return createStore<EntryViewState>({
    roots: [],
    expanded: new Set<string>(),
    activeKind: 'events',
    selectedEntry: null,
    rows: [],
    loading: false,
    translationRow: null,
    dialogOpen: false,
    pendingTableFocus: false,
  });
}

/**
 * Store por aba de versão (TanStack Store): concentra o estado compartilhado
 * entre árvore, tabela e diálogo de tradução, eliminando prop drilling entre
 * os componentes de entry-view. Uma instância por GameVersionTab — trocar de
 * aba não preserva seleção (como no original com useState).
 */
export function createEntryView(version: GameVersionId): EntryView {
  const store = createEntryStore();

  const patch = (partial: Partial<EntryViewState>): void => {
    store.setState((prev) => ({ ...prev, ...partial }));
  };

  const selectEntry = async (entry: EntryRow): Promise<void> => {
    patch({ loading: true });
    try {
      patch({ activeKind: entry.kind, selectedEntry: entry });
      const full = await loadEntry(entry.kind, entry.id, version);
      editDraft.setBase(version, entry.kind, entry.id, full);
      patch({ rows: [...(full.rows ?? [])] });
    } catch (error) {
      sendErrorNotification(error);
      patch({ selectedEntry: null, rows: [] });
    } finally {
      patch({ loading: false });
    }
  };

  const reload = async (): Promise<void> => {
    patch({ loading: true });
    try {
      const nextRoots: SideNode[] = [];
      const nextExpanded = new Set<string>();
      for (const kind of entryKindsFor(version)) {
        const entries = await loadKindEntries(kind, version);
        nextRoots.push({
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
      }
      for (const node of nextRoots) {
        nextExpanded.add(node.id);
        for (const child of node.children ?? []) nextExpanded.add(child.id);
      }
      patch({ roots: nextRoots, expanded: nextExpanded });
      // Revalida a seleção atual: entrada removida por import limpa a tabela.
      const current = store.state.selectedEntry;
      if (current) {
        const entries = await loadKindEntries(current.kind, version);
        const found = entries.find((e) => e.id === current.id);
        if (!found) {
          patch({ selectedEntry: null, rows: [] });
        } else {
          await selectEntry(found);
        }
      }
    } catch (error) {
      sendErrorNotification(error);
    } finally {
      patch({ loading: false });
    }
  };

  const actions: EntryActions = {
    reload,
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
