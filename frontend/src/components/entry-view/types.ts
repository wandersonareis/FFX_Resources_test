import { createColumnHelper, tableFeatures } from '@tanstack/react-table';
import { dto } from '@/wailsjs/go/models';
import type { EntryKind } from '@/lib/ffx/display-names';
import type { EntryRow } from '@/lib/ffx/tree-data';

/** Nó da árvore da sidebar (raiz de kind, grupo de eventos ou folha). */
export interface SideNode {
  id: string;
  label: string;
  kind?: EntryKind;
  entry?: EntryRow;
  children?: SideNode[];
}

export const EMPTY_IDS: ReadonlySet<string> = new Set<string>();

// Infra compartilhada da tabela (TanStack v9: features + columnHelper).
export const features = tableFeatures({});

export type F = typeof features;

export const columnHelper = createColumnHelper<F, dto.TextRow>();
