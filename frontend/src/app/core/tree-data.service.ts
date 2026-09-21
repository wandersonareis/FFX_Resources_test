import { Injectable, inject } from '@angular/core';
import {
  ApplyEntry,
  ExportEntry,
  GetTextEntry,
  ImportEntry,
  ListTextEntries,
} from '../../../wailsjs/go/main/App';
import { dto, services } from '../../../wailsjs/go/models';
import { DisplayLabelService } from './display-label.service';
import { EntryKind } from './display-names';
import { GameVersionId } from './game-version.service';

export interface EntryRow {
  kind: EntryKind;
  id: string;
  key: string;
  label: string;
}

export const ENTRY_KINDS: EntryKind[] = ['events', 'objects', 'macro'];

/**
 * Fluxo novo (DTO): o backend envia o sumário canônico (id + key) e o DTO
 * estruturado sob demanda. Os artefatos de export/import são JSON + .strings
 * em mods/edits; não há mais árvore de diretórios nem .txt.
 */
@Injectable({ providedIn: 'root' })
export class TreeDataService {
  private readonly labels = inject(DisplayLabelService);

  async loadKindEntries(kind: EntryKind, version: GameVersionId): Promise<EntryRow[]> {
    const summaries: services.EntrySummary[] = await ListTextEntries(
      kind,
      version as unknown as Parameters<typeof ListTextEntries>[1]
    );
    return (summaries ?? [])
      .map((s) => ({
        kind,
        id: s.id,
        key: s.key,
        label: this.labels.entryLabel(kind, s.id),
      }))
      .sort((a, b) => a.label.localeCompare(b.label));
  }

  async loadEntry(
    kind: EntryKind,
    id: string,
    version: GameVersionId
  ): Promise<dto.FileEntry> {
    return GetTextEntry(
      kind,
      id,
      version as unknown as Parameters<typeof GetTextEntry>[2]
    );
  }

  /** Exporta a entrada em JSON + .strings (mods/edits). */
  async exportEntry(
    kind: EntryKind,
    id: string,
    version: GameVersionId,
    langs: string[] = []
  ): Promise<string[]> {
    return ExportEntry(
      kind,
      id,
      version as unknown as Parameters<typeof ExportEntry>[2],
      langs
    );
  }

  /** Lê o artefato JSON padrão em mods/edits e aplica no binário. */
  async importEntry(
    kind: EntryKind,
    id: string,
    version: GameVersionId
  ): Promise<string[]> {
    return ImportEntry(
      kind,
      id,
      version as unknown as Parameters<typeof ImportEntry>[2]
    );
  }

  /** Aplica uma entrada editada (DTO) de volta no binário. */
  async applyEntry(
    kind: EntryKind,
    id: string,
    version: GameVersionId,
    entry: dto.FileEntry
  ): Promise<void> {
    return ApplyEntry(
      kind,
      id,
      version as unknown as Parameters<typeof ApplyEntry>[2],
      entry
    );
  }
}
