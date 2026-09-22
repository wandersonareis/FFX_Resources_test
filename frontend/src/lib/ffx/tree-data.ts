import {
  ExportEntry,
  GetTextEntry,
  ImportEntry,
  ListTextEntries,
} from '@/wailsjs/go/main/App';
import { dto, services } from '@/wailsjs/go/models';
import { resolveEntryLabel, EntryKind } from './display-names';
import type { GameVersionId } from './game-version';

export interface EntryRow {
  kind: EntryKind;
  id: string;
  key: string;
  label: string;
}

// Kinds servidos por versão na sidebar. lastmiss é expansão do ffx2 e não tem
// dicionário próprio → sem 'macro' (a aba não mostra "Dicionário").
export function entryKindsFor(version: GameVersionId): EntryKind[] {
  if (version === 'lastmiss') return ['events', 'objects'];
  return ['events', 'objects', 'macro'];
}

/**
 * Fluxo novo (DTO): o backend envia o sumário canônico (id + key) e o DTO
 * estruturado sob demanda. Os artefatos de export/import são JSON + .strings
 * em mods/edits; não há mais árvore de diretórios nem .txt.
 */
export async function loadKindEntries(
  kind: EntryKind,
  version: GameVersionId
): Promise<EntryRow[]> {
  const summaries: services.EntrySummary[] = await ListTextEntries(
    kind,
    version as Parameters<typeof ListTextEntries>[1]
  );
  return (summaries ?? [])
    .map((s) => ({
      kind,
      id: s.id,
      key: s.key,
      label: resolveEntryLabel(kind, s.id),
    }))
    .sort((a, b) => a.label.localeCompare(b.label));
}

export async function loadEntry(
  kind: EntryKind,
  id: string,
  version: GameVersionId
): Promise<dto.FileEntry> {
  return GetTextEntry(
    kind,
    id,
    version as Parameters<typeof GetTextEntry>[2]
  );
}

/** Exporta a entrada em JSON + .strings (mods/edits). */
export async function exportEntry(
  kind: EntryKind,
  id: string,
  version: GameVersionId,
  langs: string[] = []
): Promise<string[]> {
  return ExportEntry(
    kind,
    id,
    version as Parameters<typeof ExportEntry>[2],
    langs
  );
}

/** Lê o artefato JSON padrão em mods/edits e aplica no binário. */
export async function importEntry(
  kind: EntryKind,
  id: string,
  version: GameVersionId
): Promise<string[]> {
  return ImportEntry(
    kind,
    id,
    version as Parameters<typeof ImportEntry>[2]
  );
}
