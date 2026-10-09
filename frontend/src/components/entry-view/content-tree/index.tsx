'use client';

import { useCallback, useMemo, useRef, useState } from 'react';
import { RefreshCw } from 'lucide-react';
import { useSelector } from '@tanstack/react-store';
import type { EntryKind } from '@/lib/ffx/display-names';
import { entryKindsFor, imageKindsFor } from '@/lib/ffx/tree-data';
import { useExportSelection } from '@/lib/ffx/export-selection';
import { useVbfSelection } from '@/lib/ffx/vbf-selection';
import {
  buildContentRoots,
  highlightedNodeId,
} from '@/lib/ffx/content-tree/model';
import {
  menuTargetOf,
  type TreeCtxNode,
} from '@/lib/ffx/content-tree/menu-target';
import { Button } from '@/components/ui/button';
import type { EntryView } from '../entry-view-store';
import { EntryActionsMenu } from '../entry-actions-menu';
import { VbfExtractDialog, type VbfExtractRequest } from './vbf-extract-dialog';
import { ContentTreeProvider, type ContentTreeValue } from './tree-context';
import { TreeList } from './tree-list';
import { useContentTreeActions } from './use-content-tree-actions';
import { useContentTreeHotkeys } from './use-content-tree-hotkeys';

/**
 * Sidebar da aba: árvore de kinds/grupos/arquivos com expandir, seleção de
 * exportação (checkbox tri-state), menu de contexto (Exportar / Abrir até o
 * arquivo / Deletar) e navegação por teclado.
 *
 * O estado da árvore vive no store da view; aqui ficam só o que é local
 * (nó do menu e requisição de extração) e a projeção desses dados para a
 * árvore. Operações => use-content-tree-actions, teclado =>
 * use-content-tree-hotkeys, render => tree-list.
 */
export function ContentTree({ view }: { view: EntryView }) {
  const { version, store, actions } = view;
  const roots = useSelector(store, (s) => s.roots);
  const contentRoots = useMemo(() => buildContentRoots(roots), [roots]);
  const vbfRoots = useSelector(store, (s) => s.vbfRoots);
  const expanded = useSelector(store, (s) => s.expanded);
  const selectedEntry = useSelector(store, (s) => s.selectedEntry);
  const image = useSelector(store, (s) => s.image);
  const loading = useSelector(store, (s) => s.loading);

  const [ctxNode, setCtxNode] = useState<TreeCtxNode | null>(null);
  const [extractRequest, setExtractRequest] =
    useState<VbfExtractRequest | null>(null);
  const closeExtractDialog = useCallback(() => setExtractRequest(null), []);
  // Container da árvore: consulta os nós visíveis para o foco por setas e é o
  // ALVO dos hotkeys — o listener só vê eventos que borbulham por aqui (foco
  // dentro da árvore), sem gate por `enabled` que ficaria obsoleto no foco.
  const treeRef = useRef<HTMLDivElement>(null);

  const selection = useExportSelection();
  const vbfSelectionSnapshot = useVbfSelection();
  // Seleção de exportação por kind (checkbox tri-state + menu de contexto).
  // Depende de `byKind`, não do snapshot inteiro: o store preserva a
  // identidade do mapa quando só o formato muda, então trocar JSON/Strings
  // não reconstrói os Set por kind nem a lista de imagens marcadas.
  const selectedByKind = useMemo(() => {
    const out = new Map<EntryKind, ReadonlySet<string>>();
    for (const kind of [...entryKindsFor(version), ...imageKindsFor(version)]) {
      out.set(kind, new Set(selection.byKind.get(`${version}|${kind}`) ?? []));
    }
    return out;
  }, [selection.byKind, version]);
  const selectedImageIds = useMemo(
    () => [...(selection.byKind.get(`${version}|images`) ?? [])],
    [selection.byKind, version]
  );
  // Alvo do menu: o nó clicado (folha leva o id; grupo/raiz, não).
  const ctxTarget = useMemo(() => menuTargetOf(ctxNode), [ctxNode]);
  const selectedId = useMemo(
    () => highlightedNodeId(selectedEntry, image),
    [selectedEntry, image]
  );

  const treeActions = useContentTreeActions({
    view,
    ctxNode,
    selectedEntry,
    selectedImageIds,
    setCtxNode,
    openExtractRequest: setExtractRequest,
  });

  useContentTreeHotkeys({
    view,
    treeRef,
    selectedImageIds,
    deleteSelectedImages: treeActions.onDeleteSelectedImages,
  });

  const treeValue: ContentTreeValue = {
    expanded,
    selectedId,
    selectedByKind,
    vbfSelectionByRoot: vbfSelectionSnapshot.byRoot,
    onToggle: actions.toggleNode,
    onSelect: (node) => void actions.selectNode(node),
    onCheck: treeActions.checkNode,
  };

  return (
    <aside className="w-[290px] shrink-0 border-r p-2 flex flex-col min-h-0">
      <div className="flex items-center justify-between font-semibold px-2 py-1">
        <span>Conteúdo</span>
        <Button
          variant="ghost"
          size="icon"
          disabled={loading}
          onClick={() => void actions.coldReload()}
          aria-label="Recarregar"
        >
          <RefreshCw size={20} className={loading ? 'animate-spin' : undefined} />
        </Button>
      </div>
      <EntryActionsMenu
        view={view}
        target={ctxTarget}
        onOpenChange={(open) => {
          if (!open) setCtxNode(null);
        }}
        onExport={() => void treeActions.onCtxExport()}
        selectedImageCount={selectedImageIds.length}
        onExtractImageSelection={() => void treeActions.onExtractSelectedImages()}
        onDeleteImageSelection={treeActions.onDeleteSelectedImages}
        onVbfExtract={treeActions.onVbfExtract}
        onVbfExtractImages={treeActions.onVbfExtractImages}
        onVbfExport={treeActions.onVbfExport}
      >
        <ContentTreeProvider value={treeValue}>
          <TreeList
            treeRef={treeRef}
            vbfRoots={vbfRoots}
            roots={contentRoots}
            onNodeContextMenu={treeActions.onNodeContextMenu}
          />
        </ContentTreeProvider>
      </EntryActionsMenu>
      <VbfExtractDialog
        key={extractRequest?.requestId ?? 'closed'}
        request={extractRequest}
        onClose={closeExtractDialog}
        onDone={() => {
          setExtractRequest(null);
          void actions.coldReload();
        }}
      />
    </aside>
  );
}
