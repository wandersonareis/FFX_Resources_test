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
  /** Raiz de kind em carga: spinner no lugar do chevron, filhos a caminho. */
  loading?: boolean;

  // ---- Navegador do container .vbf (fonte somente leitura) ----
  /** true = nó que pertence ao container .vbf. */
  vbf?: boolean;
  /** Caminho absoluto do container — só nas raízes. */
  vbfRoot?: string;
  /** Caminho interno dentro do container ("" = raiz do .vbf). */
  vbfPath?: string;
  /** É macrodic.dcp: os filhos saem de loadVbfMacroChunks (chunk_XX). */
  macro?: boolean;
  /**
   * Pode ser expandido mesmo sem filhos carregados (diretório do .vbf ainda
   * não buscado). Vira false quando a listagem devolve vazio.
   */
  expandable?: boolean;
  /** Arquivo que o app não serve (áudio, vídeo, .exe…): clique avisa. */
  unsupported?: boolean;
}

/** IDs das folhas de entrada contidas no nó; grupos de imagens são aninhados. */
export function entryIdsInNode(node: SideNode): string[] {
  if (node.entry) return [node.entry.id];
  return (node.children ?? []).flatMap(entryIdsInNode);
}

/** Estado tri-state do checkbox de um nó data/ sobre todas as folhas abaixo. */
export function entryCheckState(
  node: SideNode,
  selected: ReadonlySet<string>
): boolean | 'indeterminate' {
  const ids = entryIdsInNode(node);
  const count = ids.filter((id) => selected.has(id)).length;
  return count === 0 ? false : count === ids.length ? true : 'indeterminate';
}

/**
 * Id estável de uma folha da árvore. Folhas de .vbf são identificadas pelo
 * CAMINHO no container (o mesmo id de data/ pode existir nos dois lados e
 * em dois containers diferentes).
 */
export function entryNodeId(entry: EntryRow): string {
  return entry.vbf
    ? `vbfleaf:${entry.vbf.root}|${entry.vbf.path}`
    : `leaf:${entry.kind}:${entry.id}`;
}

export const EMPTY_IDS: ReadonlySet<string> = new Set<string>();

// Infra compartilhada da tabela (TanStack v9: features + columnHelper).
export const features = tableFeatures({});

export type F = typeof features;

export const columnHelper = createColumnHelper<F, dto.TextRow>();
