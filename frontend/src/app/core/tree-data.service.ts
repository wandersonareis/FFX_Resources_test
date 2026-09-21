import { Injectable, inject } from '@angular/core';
import {
  BuildTree,
  GetTextEntry,
  ListTextEntries,
} from '../../../wailsjs/go/main/App';
import { dto, fileFormats, services, spira } from '../../../wailsjs/go/models';
import { DisplayLabelService } from './display-label.service';
import { EntryKind } from './display-names';
import { GameVersionId } from './game-version.service';

export interface EntryRow {
  kind: EntryKind;
  id: string;
  key: string;
  label: string;
  /** Dados de arquivo (paths de origem/extração) quando resolvidos via BuildTree. */
  fileInfo?: fileFormats.TreeNodeData;
}

export const ENTRY_KINDS: EntryKind[] = ['events', 'objects', 'macro'];

/**
 * Backend envia sumário canônico (id + key); labels vivem no frontend.
 * Mantém um índice id -> TreeNodeData (via BuildTree) para as operações
 * de arquivo (Extract/Compress/View), que exigem caminho em disco.
 */
@Injectable({ providedIn: 'root' })
export class TreeDataService {
  private readonly labels = inject(DisplayLabelService);
  private pathIndex = new Map<string, fileFormats.TreeNodeData>();

  async loadKindEntries(kind: EntryKind, version: GameVersionId): Promise<EntryRow[]> {
    const summaries: services.EntrySummary[] = await ListTextEntries(
      kind,
      version as unknown as Parameters<typeof ListTextEntries>[1]
    );
    await this.ensurePathIndex();
    return (summaries ?? [])
      .map((s) => ({
        kind,
        id: s.id,
        key: s.key,
        label: this.labels.entryLabel(kind, s.id),
        fileInfo: this.pathIndex.get(s.id),
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

  private async ensurePathIndex(): Promise<void> {
    if (this.pathIndex.size > 0) return;
    const roots: spira.TreeNode[] = (await BuildTree()) ?? [];
    const walk = (nodes: spira.TreeNode[] | undefined): void => {
      for (const node of nodes ?? []) {
        const info = node.data;
        if (info?.source) {
          const { name, name_prefix } = info.source;
          if (name_prefix) this.pathIndex.set(name_prefix, info);
          if (name) {
            this.pathIndex.set(name, info);
            const withoutExt = name.replace(/\.[^.]+$/, '');
            if (!this.pathIndex.has(withoutExt)) this.pathIndex.set(withoutExt, info);
          }
        }
        walk(node.children);
      }
    };
    walk(roots);
  }

  refreshPathIndex(): void {
    this.pathIndex.clear();
  }
}
