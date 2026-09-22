import { useSyncExternalStore } from 'react';
import { dto } from '@/wailsjs/go/models';
import { EntryKind } from './display-names';
import type { GameVersionId } from './game-version';

/** Chave de linha no rascunho: mesma que o backend usa (Index + Name). */
export function rowKey(row: Pick<dto.TextRow, 'index' | 'name'>): string {
  return `${row.index}:${row.name ?? ''}`;
}

interface DraftState {
  version: GameVersionId;
  /** Entrada original carregada (base para reconstruir o lote no salvar). */
  entry: dto.FileEntry;
  /** Edições por rowKey → idioma → texto (só o que difere do original). */
  edits: Map<string, Map<string, string>>;
}

const draftKey = (version: GameVersionId, kind: EntryKind, id: string): string =>
  `${version}|${kind}|${id}`;

interface DraftSnapshot {
  hasDirty: boolean;
  revision: number;
}

type Listener = () => void;
type ReloadListener = () => void;

/**
 * Rascunhos de edição em memória (singleton reativo). Trocar de arquivo não
 * perde nada: o rascunho fica por (version|kind|id) até o usuário salvar ou
 * descartar. O "traduzido" é o texto editado que substitui o slot 'us' no
 * binário. O snapshot é imutável para uso com useSyncExternalStore.
 */
class EditDraftStore {
  private readonly states = new Map<string, DraftState>();
  private readonly dirtyFiles = new Set<string>();
  private readonly listeners = new Set<Listener>();
  private readonly reloadListeners = new Set<ReloadListener>();
  private snapshot: DraftSnapshot = { hasDirty: false, revision: 0 };

  subscribe = (listener: Listener): (() => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };

  getSnapshot = (): DraftSnapshot => this.snapshot;

  private readonly serverSnapshot: DraftSnapshot = { hasDirty: false, revision: 0 };

  getServerSnapshot = (): DraftSnapshot => this.serverSnapshot;

  /** Callback chamado após salvar tudo (reselect da aba ativa). */
  onSaved(listener: ReloadListener): () => void {
    this.reloadListeners.add(listener);
    return () => this.reloadListeners.delete(listener);
  }

  private notify(): void {
    this.snapshot = {
      hasDirty: this.dirtyFiles.size > 0,
      revision: this.snapshot.revision + 1,
    };
    for (const listener of this.listeners) listener();
  }

  /** Registra a base (entrada original) ao abrir o arquivo na tabela. */
  setBase(
    version: GameVersionId,
    kind: EntryKind,
    id: string,
    entry: dto.FileEntry
  ): void {
    const key = draftKey(version, kind, id);
    const existing = this.states.get(key);
    if (existing) {
      existing.entry = entry;
      return;
    }
    this.states.set(key, { version, entry, edits: new Map() });
    this.touch(key, false);
  }

  /** Descarta rascunho sem edição (base fora de uso). */
  dropBase(version: GameVersionId, kind: EntryKind, id: string): void {
    const key = draftKey(version, kind, id);
    if (this.states.get(key)?.edits.size === 0) {
      this.states.delete(key);
    }
  }

  /** Registra/edita o texto de uma célula. Texto igual ao original limpa a edição. */
  setCell(
    version: GameVersionId,
    kind: EntryKind,
    id: string,
    row: dto.TextRow,
    lang: string,
    text: string
  ): void {
    const key = draftKey(version, kind, id);
    const state = this.states.get(key);
    if (!state) return;

    const rKey = rowKey(row);
    const original = row.text?.[lang] ?? '';
    let byLang = state.edits.get(rKey);
    if (text === original) {
      if (byLang) {
        byLang.delete(lang);
        if (byLang.size === 0) state.edits.delete(rKey);
      }
    } else {
      if (!byLang) {
        byLang = new Map();
        state.edits.set(rKey, byLang);
      }
      byLang.set(lang, text);
    }
    this.touch(key, state.edits.size > 0);
  }

  /** Valor editado da célula (undefined = sem edição). */
  editOf(
    version: GameVersionId,
    kind: EntryKind,
    id: string,
    row: dto.TextRow,
    lang: string
  ): string | undefined {
    return this.states
      .get(draftKey(version, kind, id))
      ?.edits.get(rowKey(row))
      ?.get(lang);
  }

  isDirty(version: GameVersionId, kind: EntryKind, id: string): boolean {
    const state = this.states.get(draftKey(version, kind, id));
    return !!state && state.edits.size > 0;
  }

  /** Limpa o rascunho (após salvar a entrada). */
  clear(version: GameVersionId, kind: EntryKind, id: string): void {
    const key = draftKey(version, kind, id);
    this.states.delete(key);
    this.touch(key, false);
  }

  /** Limpa todos os rascunhos (após salvar tudo). */
  clearAll(): void {
    this.states.clear();
    this.touchAll();
  }

  /**
   * Monta a Collection do kind/version com as entradas editadas
   * (base + edições mescladas), pronta para ApplyTextCollection.
   */
  buildCollection(
    version: GameVersionId,
    kind: EntryKind
  ): Record<string, dto.FileEntry> {
    const out: Record<string, dto.FileEntry> = {};
    for (const [key, state] of this.states) {
      if (state.edits.size === 0 || state.version !== version) continue;
      const [k, id] = splitDraftKey(key);
      if (k !== kind) continue;
      out[id] = this.merge(state);
    }
    return out;
  }

  /** Lotes por kind para os arquivos sujos (por versão). */
  dirtyBatches(): Map<GameVersionId, Map<EntryKind, string[]>> {
    const out = new Map<GameVersionId, Map<EntryKind, string[]>>();
    for (const [key, state] of this.states) {
      if (state.edits.size === 0) continue;
      const [kind, id] = splitDraftKey(key);
      const byKind = out.get(state.version) ?? new Map<EntryKind, string[]>();
      const list = byKind.get(kind) ?? [];
      list.push(id);
      byKind.set(kind, list);
      out.set(state.version, byKind);
    }
    return out;
  }

  private merge(state: DraftState): dto.FileEntry {
    const rows = state.entry.rows.map((row) => {
      const edits = state.edits.get(rowKey(row));
      if (!edits || edits.size === 0) return new dto.TextRow(row);
      const text = { ...(row.text ?? {}) };
      for (const [lang, value] of edits) {
        text[lang] = value;
      }
      return new dto.TextRow({ ...row, text });
    });
    return dto.FileEntry.createFrom({ metadata: state.entry.metadata, rows });
  }

  private touch(key: string, dirty: boolean): void {
    if (dirty) {
      this.dirtyFiles.add(key);
    } else {
      this.dirtyFiles.delete(key);
    }
    this.notify();
  }

  private touchAll(): void {
    this.dirtyFiles.clear();
    this.notify();
  }

  private fireSaved(): void {
    for (const listener of this.reloadListeners) listener();
  }

  /** Chamado pelo saveAll após persistir tudo. */
  markSaved(): void {
    this.clearAll();
    this.fireSaved();
  }
}

function splitDraftKey(key: string): [EntryKind, string] {
  const parts = key.split('|');
  return [parts[1] as EntryKind, parts[2]];
}

export const editDraft = new EditDraftStore();

export function useEditDraft(): { store: EditDraftStore; snapshot: DraftSnapshot } {
  const snapshot = useSyncExternalStore(
    editDraft.subscribe,
    editDraft.getSnapshot,
    editDraft.getServerSnapshot
  );
  return { store: editDraft, snapshot };
}
