import { useEffect } from 'react';
import { createStore } from '@tanstack/store';
import { useSelector } from '@tanstack/react-store';
import type { EntryKind } from './display-names';
import type { GameVersionId } from './game-version';

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
  /** ids selecionados por `${version}|${kind}` (imutável entre notificações). */
  byKind: ReadonlyMap<string, readonly string[]>;
  /** Formato ativo para todas as exportações feitas pelo frontend. */
  format: ExportFormat;
}

const stateKey = (version: GameVersionId, kind: EntryKind): string =>
  `${version}|${kind}`;

/**
 * Store reativo da seleção (TanStack Store): o estado é o snapshot imutável
 * {revision, byKind, format} consumido pelo hook. Os dados de trabalho ficam
 * imperativos no singleton abaixo.
 */
const exportStore = createStore<SelectionSnapshot>({
  revision: 0,
  byKind: new Map(),
  // 'json' é o default de hidratação; o formato persistido no localStorage é
  // carregado no cliente dentro do useExportSelection (mesmo padrão do
  // tabStorage no app-shell).
  format: 'json',
});

/**
 * Seleção de entradas para exportação (singleton reativo, mesmo padrão do
 * edit-draft). Marcar na sidebar alimenta dois destinos:
 *   - menu de contexto: exporta só o que está marcado na árvore invocada;
 *   - Exportar do topo: ids por kind, com `[]` (nada marcado) = tudo do kind.
 */
class ExportSelectionStore {
  private format: ExportFormat = 'json';
  private readonly selected = new Map<string, Set<string>>();

  /** Snapshot imutável atual (leitura imperativa). */
  getSnapshot(): SelectionSnapshot {
    return exportStore.state;
  }

  private notify(): void {
    const copied = new Map<string, readonly string[]>();
    for (const [key, ids] of this.selected) copied.set(key, [...ids]);
    exportStore.setState((prev) => ({
      revision: prev.revision + 1,
      byKind: copied,
      format: this.format,
    }));
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
  const snapshot = useSelector(exportStore);

  // Hydration-safe: localStorage só existe no cliente. setFormat cedo-retorna
  // quando o formato já é o mesmo, então chamadas repetidas (remount da
  // aba) são no-op.
  useEffect(() => {
    exportSelection.setFormat(loadStoredFormat());
  }, []);

  return snapshot;
}
