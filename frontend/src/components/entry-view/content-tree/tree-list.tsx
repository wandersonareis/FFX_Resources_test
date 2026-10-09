'use client';

import type { MouseEvent, RefObject } from 'react';
import { ScrollArea } from '@/components/ui/scroll-area';
import type { SideNode } from '@/lib/ffx/content-tree/model';
import { TreeItem } from './tree-item';

/**
 * ScrollArea da árvore com as duas seções e seus separadores. Os containers
 * .vbf vêm PRIMEIRO: são a fonte da verdade e a árvore de data/ (imutável) é
 * o espelho já extraído.
 */
export function TreeList({
  treeRef,
  vbfRoots,
  roots,
  onNodeContextMenu,
}: {
  treeRef: RefObject<HTMLDivElement | null>;
  vbfRoots: readonly SideNode[];
  roots: readonly SideNode[];
  onNodeContextMenu: (event: MouseEvent) => void;
}) {
  return (
    <ScrollArea
      id="entry-tree"
      ref={treeRef}
      className="flex-1 min-h-0"
      onContextMenuCapture={onNodeContextMenu}
    >
      {vbfRoots.length > 0 ? (
        <div className="px-2 pt-1 pb-1 text-xs font-medium text-muted-foreground">
          Containers .vbf · somente leitura
        </div>
      ) : null}
      {vbfRoots.map((node) => (
        <TreeItem key={node.id} node={node} depth={0} />
      ))}
      {roots.length > 0 && vbfRoots.length > 0 ? (
        <div className="px-2 pt-3 pb-1 text-xs font-medium text-muted-foreground border-t mt-1">
          Arquivos de data/
        </div>
      ) : null}
      {roots.map((node) => (
        <TreeItem key={node.id} node={node} depth={0} />
      ))}
    </ScrollArea>
  );
}
