'use client';

import { createContext, useContext, type ReactNode } from 'react';
import type { EntryKind } from '@/lib/ffx/display-names';
import type { SideNode } from '@/lib/ffx/content-tree/model';

/**
 * Estado e ações compartilhados por TODA a árvore. Existe só para a
 * recursão do TreeItem não arrastar 8 props nível a nível — o dono dos
 * dados continua sendo o store da view (index.tsx apenas o projeta aqui).
 *
 * O valor é remontado a cada render do ContentTree: mesmo custo de re-render
 * de hoje, sem subscriber extra e sem risco de closure velha.
 */
export interface ContentTreeValue {
  expanded: ReadonlySet<string>;
  selectedId: string | null;
  selectedByKind: ReadonlyMap<EntryKind, ReadonlySet<string>>;
  vbfSelectionByRoot: ReadonlyMap<string, readonly string[]>;
  onToggle: (node: SideNode) => void;
  onSelect: (node: SideNode) => void;
  onCheck: (node: SideNode, checked: boolean) => void;
}

const ContentTreeContext = createContext<ContentTreeValue | null>(null);

export function ContentTreeProvider({
  value,
  children,
}: {
  value: ContentTreeValue;
  children: ReactNode;
}) {
  return (
    <ContentTreeContext.Provider value={value}>
      {children}
    </ContentTreeContext.Provider>
  );
}

export function useContentTree(): ContentTreeValue {
  const value = useContext(ContentTreeContext);
  if (!value) {
    throw new Error('useContentTree precisa estar dentro de ContentTreeProvider');
  }
  return value;
}
