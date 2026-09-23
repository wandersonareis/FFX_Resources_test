import { useSyncExternalStore } from 'react';
import type { EntryKind } from './display-names';
import type { GameVersionId } from './game-version';

type Listener = () => void;

export type ExportFormat = 'json' | 'strings';

export const EXPORT_FORMAT_LABELS: Record<ExportFormat, string> = {
  json: 'JSON',
  strings: 'Strings',
};

const FORMAT_STORAGE_KEY = 'ffx.export-format';

function loadStoredFormat(): ExportFormat {
  try {
    return localStorage.getItem(FORMAT_STORAGE_KEY) === 'strings'
      ? 'strings'
      : 'json';
  } catch {
    return 'json';
  }
}

interface SelectionSnapshot {
  revision: number;
  /** ids selecionados por `${version}|${kind}` (imutável p/ useSyncExternalStore). */
  byKind: ReadonlyMap<string, readonly string[]>;
  /** Formato ativo para todas as exportações feitas pelo frontend. */
  format: ExportFormat;
}

const stateKey = (version: GameVersionId, kind: EntryKind): string =>
  `${version}|${kind}`;

/**
 * Seleção de entradas para exportação (singleton reativo, mesmo padrão do
 * edit-draft). Marcar na sidebar alimenta dois destinos:
 *   - menu de contexto: exporta só o que está marcado na árvore invocada;
 *   - Exportar do topo: ids por kind, com `[]` (nada marcado) = tudo do kind.
 */
class ExportSelectionStore {
  private format: ExportFormat = loadStoredFormat();
  private readonly selected = new Map<string, Set<string>>();
  private readonly listeners = new Set<Listener>();
  private snapshot: SelectionSnapshot = {
    revision: 0,
    byKind: new Map(),
    format: this.format,
  };
  private readonly serverSnapshot: SelectionSnapshot = {
    revision: 0,
    byKind: new Map(),
    format: 'json',
  };

  subscribe = (listener: Listener): (() => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };

  getSnapshot = (): SelectionSnapshot => this.snapshot;

  getServerSnapshot = (): SelectionSnapshot => this.serverSnapshot;

  private notify(): void {
    const copied = new Map<string, readonly string[]>();
    for (const [key, ids] of this.selected) copied.set(key, [...ids]);
    this.snapshot = {
      revision: this.snapshot.revision + 1,
      byKind: copied,
      format: this.format,
    };
    for (const listener of this.listeners) listener();
  }

  /**
   * Define o formato de exportação usado por TODAS as exportações do
   * frontend (Exportar do topo e menu de contexto). Persistido em
   * localStorage, como o tabStorage faz com a aba ativa.
   */
  setFormat(format: ExportFormat): void {
    if (this.format === format) return;
    this.format = format;
    try {
      localStorage.setItem(FORMAT_STORAGE_KEY, format);
    } catch {
      // localStorage indisponível — segue só em memória
    }
    this.notify();
  }

  formatOf(): ExportFormat {
    return this.format;
  }

  /**
   * Marca/desmarca várias entradas de uma vez (grupo = filhos; clique em
   * indeterminado chega como checked=true via radix). Uma única notificação.
   */
  setMany(
    version: GameVersionId,
    kind: EntryKind,
    ids: string[],
    checked: boolean
  ): void {
    if (ids.length === 0) return;
    const key = stateKey(version, kind);
    let set = this.selected.get(key);
    if (!set) {
      if (!checked) return;
      set = new Set();
      this.selected.set(key, set);
    }
    if (checked) {
      for (const id of ids) set.add(id);
    } else {
      for (const id of ids) set.delete(id);
    }
    if (set.size === 0) this.selected.delete(key);
    this.notify();
  }

  has(version: GameVersionId, kind: EntryKind, id: string): boolean {
    return this.selected.get(stateKey(version, kind))?.has(id) ?? false;
  }

  /** ids marcados do kind. Vazio = nenhum → exportação considera "tudo". */
  idsOf(version: GameVersionId, kind: EntryKind): string[] {
    return [...(this.selected.get(stateKey(version, kind)) ?? [])];
  }

  countOf(version: GameVersionId, kind: EntryKind): number {
    return this.selected.get(stateKey(version, kind))?.size ?? 0;
  }
}

export const exportSelection = new ExportSelectionStore();

/** Snapshot reativo da seleção (rerender ao marcar/desmarcar). */
export function useExportSelection(): SelectionSnapshot {
  return useSyncExternalStore(
    exportSelection.subscribe,
    exportSelection.getSnapshot,
    exportSelection.getServerSnapshot
  );
}
