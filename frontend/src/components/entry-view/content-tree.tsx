'use client';

import { useCallback, useMemo, useRef, useState } from 'react';
import { useHotkey, useHotkeys } from '@tanstack/react-hotkeys';
import { loggedToast as toast } from '@/lib/ffx/toast-logged';
import { RefreshCw } from 'lucide-react';
import { useSelector } from '@tanstack/react-store';
import type { EntryKind } from '@/lib/ffx/display-names';
import {
  entryKindsFor,
  imageKindsFor,
  exportDestination,
  exportJSON,
  exportStrings,
  extractImageSelection,
} from '@/lib/ffx/tree-data';
import {
  exportVbfSelection,
  extractVbfImagesSelection,
  invalidateVbfImage,
  selectVbfImageExtractDir,
} from '@/lib/ffx/vbf';
import { parseError } from '@/lib/ffx/error-handler';
import {
  EXPORT_FORMAT_LABELS,
  exportSelection,
  useExportSelection,
} from '@/lib/ffx/export-selection';
import { vbfSelection, useVbfSelection } from '@/lib/ffx/vbf-selection';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { EntryActionsMenu, menuTargetOf } from './entry-actions-menu';
import { VbfExtractDialog, type VbfExtractRequest } from './vbf-extract-dialog';
import type { EntryView } from './entry-view-store';
import { entryIdsInNode, entryNodeId, type SideNode } from './types';
import { TreeItem } from './tree-item';

/**
 * Sidebar da aba: árvore de kinds/grupos/arquivos com expandir, seleção de
 * exportação (checkbox tri-state), menu de contexto (Exportar / Abrir até o
 * arquivo / Deletar) e navegação por teclado. Estado da árvore vive no store
 * da view; ctxNode é local.
 */
export function ContentTree({ view }: { view: EntryView }) {
  const { version, store, actions } = view;
  const roots = useSelector(store, (s) => s.roots);
  const vbfRoots = useSelector(store, (s) => s.vbfRoots);
  const expanded = useSelector(store, (s) => s.expanded);
  const selectedEntry = useSelector(store, (s) => s.selectedEntry);
  const image = useSelector(store, (s) => s.image);
  const imageAction = useSelector(store, (s) => s.imageAction);
  const loading = useSelector(store, (s) => s.loading);
  const [ctxNode, setCtxNode] = useState<{
    id: string;
    kind?: EntryKind;
    label: string;
    vbfRoot?: string;
    vbfPath?: string;
    selectedVbfPaths?: string[];
  } | null>(null);
  const [extractRequest, setExtractRequest] =
    useState<VbfExtractRequest | null>(null);
  const extractRequestCounter = useRef(0);
  const closeExtractDialog = useCallback(() => setExtractRequest(null), []);
  // Container da árvore: consulta os nós visíveis para o foco por setas e é o
  // ALVO dos hotkeys — o listener só vê eventos que borbulham por aqui (foco
  // dentro da árvore), sem gate por `enabled` que ficaria obsoleto no foco.
  const treeRef = useRef<HTMLDivElement>(null);

  const selection = useExportSelection();
  const vbfSelectionSnapshot = useVbfSelection();
  // Seleção de exportação por kind (checkbox tri-state + menu de contexto).
  const selectedByKind = useMemo(() => {
    const out = new Map<EntryKind, ReadonlySet<string>>();
    for (const kind of [...entryKindsFor(version), ...imageKindsFor(version)]) {
      out.set(kind, new Set(selection.byKind.get(`${version}|${kind}`) ?? []));
    }
    return out;
  }, [selection, version]);
  const selectedImageIds = useMemo(
    () => [...(selection.byKind.get(`${version}|images`) ?? [])],
    [selection, version]
  );
  // Alvo do menu: o nó clicado (folha leva o id; grupo/raiz, não).
  const ctxTarget = useMemo(() => menuTargetOf(ctxNode), [ctxNode]);

  /**
   * Nó realçado para a seleção. Em imagens a árvore só guarda o
   * REPRESENTANTE do grupo de cópias — quando a seleção é uma cópia oculta
   * (a lista "Repetidas" navega até ela), o realce salta para o
   * representante em vez de sumir da árvore.
   */
  const selectedId = useMemo(() => {
    if (!selectedEntry) return null;
    // Folha de .vbf é identificada pelo caminho no container: o mesmo id de
    // data/ pode existir dos dois lados (e em dois containers).
    if (selectedEntry.vbf) return entryNodeId(selectedEntry);
    const base = `leaf:${selectedEntry.kind}:${selectedEntry.id}`;
    if (selectedEntry.kind !== 'images' || !image) return base;
    if (image.metadata.id !== selectedEntry.id) return base;
    const group = [
      selectedEntry.id,
      ...(image.duplicates ?? []).map((d) => d.id),
    ].sort();
    return `leaf:images:${group[0]}`;
  }, [selectedEntry, image]);

  // ---- Teclado: sidebar — ↑/↓ movem o foco entre nós visíveis; Enter
  // alterna expandir/fechar (grupo/raiz) ou abre o arquivo na tabela (folha);
  // → folha leva o foco para a tabela (↑/↓ direto).

  // Índice id → nó: o callback do hotkey recebe só o evento, então o nó é
  // resolvido pelo DOM ([data-node-id]) e devolvido por este mapa.
  const nodeById = useMemo(() => {
    const map = new Map<string, SideNode>();
    const walk = (node: SideNode) => {
      map.set(node.id, node);
      node.children?.forEach(walk);
    };
    vbfRoots.forEach(walk);
    roots.forEach(walk);
    return map;
  }, [roots, vbfRoots]);

  /**
   * Botão de nó sob o foco. `null` = chevron/checkbox/viewport: esses casos
   * ficam com o comportamento nativo do botão (Enter alterna de verdade).
   */
  const focusedNodeButton = (event: KeyboardEvent) =>
    event.target instanceof Element
      ? event.target.closest<HTMLElement>('[data-node-button]')
      : null;

  const moveFocus = (event: KeyboardEvent, delta: number) => {
    const button = focusedNodeButton(event);
    const buttons = Array.from(
      treeRef.current?.querySelectorAll<HTMLElement>('[data-node-button]') ?? []
    );
    const idx = button ? buttons.indexOf(button) : -1;
    if (idx < 0) return;
    event.preventDefault();
    buttons[idx + delta]?.focus();
  };

  const activate = (event: KeyboardEvent, mode: 'enter' | 'right') => {
    const button = focusedNodeButton(event);
    if (!button) return;
    const id = button.closest('[data-node-id]')?.getAttribute('data-node-id');
    const node = id ? nodeById.get(id) : undefined;
    if (!node) return;
    // Folha abre; arquivo fora do escopo avisa; grupo/raiz alterna.
    // `preventDefault: false` no registro: o cancelamento é manual e SÓ aqui,
    // para matar o click nativo do <button> sem matar Enter no checkbox/chevron.
    event.preventDefault();
    if (mode === 'right') {
      if (node.entry) {
        // Folha: abre o arquivo e já leva o foco para a tabela (↑/↓ direto).
        actions.requestTableFocus();
        void actions.selectNode(node);
      } else if (node.unsupported) {
        void actions.selectNode(node);
      } else if (!expanded.has(node.id)) {
        // Grupo/raiz: expande se colapsado (convenção de treeview).
        actions.toggleNode(node);
      }
      return;
    }
    if (node.entry || node.unsupported) void actions.selectNode(node);
    else actions.toggleNode(node);
  };

  // Callbacks e options são sincronizados a cada render pelo hook, então não
  // há closure velha de `actions`/`expanded` (dispensa useCallback).
  useHotkeys(
    [
      {
        hotkey: 'ArrowDown',
        callback: (event) => moveFocus(event, 1),
        options: { meta: { name: 'Próximo nó', group: 'Árvore' } },
      },
      {
        hotkey: 'ArrowUp',
        callback: (event) => moveFocus(event, -1),
        options: { meta: { name: 'Nó anterior', group: 'Árvore' } },
      },
      {
        hotkey: 'Enter',
        callback: (event) => activate(event, 'enter'),
        options: { meta: { name: 'Abrir/alternar nó', group: 'Árvore' } },
      },
      {
        hotkey: 'ArrowRight',
        callback: (event) => activate(event, 'right'),
        options: { meta: { name: 'Expandir ou ir para a tabela', group: 'Árvore' } },
      },
    ],
    // `preventDefault`/`stopPropagation` desligados: senão Enter morre no
    // checkbox/chevron e o evento deixa de subir (cortando o Delete em
    // document e os handlers sintéticos dos ancestres).
    { target: treeRef, preventDefault: false, stopPropagation: false, ignoreInputs: true }
  );

  // Checkbox de data/: folha = 1 id; grupo = todos os filhos.
  // Checkbox do .vbf: guarda o caminho (diretório = subárvore, expandida
  // pelo backend); inclui imagens e formatos fora do escopo de texto para
  // permitir a extração dos binários originais.
  const checkNode = useCallback(
    (node: SideNode, checked: boolean) => {
      if (node.vbf) {
        const root = node.vbfRoot ?? node.entry?.vbf?.root;
        const path = node.vbfPath ?? node.entry?.vbf?.path ?? '';
        if (root) vbfSelection.set(root, path, checked);
        return;
      }
      if (!node.kind) return;
      const ids = entryIdsInNode(node);
      exportSelection.setMany(version, node.kind, ids, checked);
    },
    [version]
  );

  const onExtractSelectedImages = useCallback(async () => {
    if (selectedImageIds.length === 0) {
      toast.message('Nenhuma imagem marcada para extrair.');
      return;
    }
    try {
      const result = await extractImageSelection(selectedImageIds, version);
      if (result.failed.length > 0) {
        toast.warning(`Extraídas ${result.done.length} de ${result.total} imagens.`, {
          description: result.failed.slice(0, 3).join('; '),
        });
      } else {
        toast.success(`Extraídas ${result.done.length} imagens selecionadas.`);
      }
      await actions.reload();
    } catch (error) {
      toast.error(parseError(error));
    }
  }, [actions, selectedImageIds, version]);

  const onDeleteSelectedImages = useCallback(() => {
    if (selectedImageIds.length === 0) return;
    actions.openImageAction({
      type: 'delete-selection',
      id: selectedImageIds[0],
      ids: selectedImageIds,
      label: `${selectedImageIds.length} imagens selecionadas`,
      duplicates: null,
    });
  }, [actions, selectedImageIds]);

  useHotkey(
    'Delete',
    () => {
      if (imageAction) return;
      if (selectedImageIds.length > 0) {
        onDeleteSelectedImages();
        return;
      }
      if (selectedEntry?.kind !== 'images' || selectedEntry.vbf) return;
      actions.openImageAction({
        type: 'delete',
        id: selectedEntry.id,
        label: selectedEntry.label,
        duplicates: image?.duplicates ?? [],
      });
    },
    {
      enabled: !imageAction &&
        (selectedImageIds.length > 0 || (selectedEntry?.kind === 'images' && !selectedEntry.vbf)),
      ignoreInputs: true,
    }
  );

  // Exporta só o que está marcado na árvore de data/, no FORMATO SELECIONADO
  // no topo (JSON | Strings) — leitura imperativa do store compartilhado.
  const onCtxExport = useCallback(async () => {
    const kind = ctxNode?.kind;
    if (!kind) return;
    const ids = exportSelection.idsOf(version, kind);
    if (ids.length === 0) {
      toast.message(`Nenhum arquivo marcado para exportar (${kind}).`);
      return;
    }
    const format = exportSelection.formatOf();
    const label = EXPORT_FORMAT_LABELS[format];
    const toastId = 'export';
    toast.loading(`Exportando (${label})…`, { id: toastId });
    try {
      const paths =
        format === 'json'
          ? await exportJSON(kind, version, ids)
          : await exportStrings(kind, version, ids);
      const dest = exportDestination(paths);
      toast.success(
        `Exportado (${label}): ${dest || `${(paths ?? []).length} arquivo(s)`}.`,
        { id: toastId }
      );
    } catch (error) {
      toast.error(parseError(error), { id: toastId });
    }
  }, [ctxNode, version]);

  // O menu do .vbf opera somente sobre os caminhos marcados no container
  // clicado — o nó sob o cursor apenas determina qual container recebe o
  // menu, não troca a seleção por esse nó.
  const vbfPathsForContext = ctxNode?.vbfRoot
    ? (ctxNode.selectedVbfPaths ?? vbfSelection.pathsOf(ctxNode.vbfRoot))
    : [];

  return (
    <aside className="w-70 shrink-0 border-r p-2 flex flex-col min-h-0">
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
        onExport={() => void onCtxExport()}
        selectedImageCount={selectedImageIds.length}
        onExtractImageSelection={() => void onExtractSelectedImages()}
        onDeleteImageSelection={onDeleteSelectedImages}
        onVbfExtract={(mode) => {
          if (!ctxNode?.vbfRoot || vbfPathsForContext.length === 0) return;
          setExtractRequest({
            requestId: ++extractRequestCounter.current,
            root: ctxNode.vbfRoot,
            paths: vbfPathsForContext,
            mode,
          });
          setCtxNode(null);
        }}
        onVbfExtractImages={(mode) => {
          if (!ctxNode?.vbfRoot || vbfPathsForContext.length === 0) return;
          const root = ctxNode.vbfRoot;
          const paths = [...vbfPathsForContext];
          setCtxNode(null);
          void (async () => {
            let destRoot = '';
            if (mode === 'choose') {
              destRoot = await selectVbfImageExtractDir();
              if (!destRoot) return;
            }
            try {
              const result = await extractVbfImagesSelection(root, paths, destRoot);
              if (result.total === 0) {
                toast.info('A seleção não contém imagens compatíveis para extrair.');
              } else if (result.failed.length > 0) {
                toast.warning(
                  `Extraídas ${result.done.length} de ${result.total} imagens.`,
                  { description: result.failed.slice(0, 3).join('; ') }
                );
              } else {
                toast.success(`Extraídas ${result.done.length} imagens do .vbf.`);
              }
              if (
                selectedEntry?.vbf &&
                result.done?.includes(selectedEntry.vbf.path)
              ) {
                invalidateVbfImage(selectedEntry.vbf.root, selectedEntry.vbf.path);
                await actions.selectEntry(selectedEntry);
              }
            } catch (error) {
              toast.error(parseError(error));
            }
          })();
        }}
        onVbfExport={() => {
          if (!ctxNode?.vbfRoot || vbfPathsForContext.length === 0) return;
          void (async () => {
            const format = exportSelection.formatOf();
            const label = EXPORT_FORMAT_LABELS[format];
            toast.loading(`Exportando (${label})…`, { id: 'export' });
            try {
              const paths = await exportVbfSelection(
                ctxNode.vbfRoot!,
                format,
                vbfPathsForContext
              );
              if (paths.length === 0) {
                toast.info('A seleção não contém textos compatíveis para extrair.', {
                  id: 'export',
                });
                return;
              }
              const dest = exportDestination(paths);
              toast.success(
                `Exportado (${label}): ${dest || `${paths.length} arquivo(s)`}.`,
                { id: 'export' }
              );
            } catch (error) {
              toast.error(parseError(error), { id: 'export' });
            }
          })();
          setCtxNode(null);
        }}
      >
        <ScrollArea
          id="entry-tree"
          ref={treeRef}
          className="flex-1 min-h-0"
          onContextMenuCapture={(event) => {
            const row = (event.target as HTMLElement).closest(
              '[data-node-id]'
            );
            if (!(row instanceof HTMLElement)) {
              setCtxNode(null);
              return;
            }
            const kind = row.getAttribute(
              'data-node-kind'
            ) as EntryKind | null;
            const isVbf = row.getAttribute('data-node-vbf') === 'true';
            if (isVbf) {
              const vbfRoot = row.getAttribute('data-node-vbf-root') ?? '';
              const selectedVbfPaths = vbfSelection.pathsOf(vbfRoot);
              const clickedPath = row.getAttribute('data-node-vbf-path') ?? '';
              const isFile = row.getAttribute('data-node-vbf-file') === 'true';
              // Checkboxes têm prioridade e podem conter arquivos ou pastas.
              // Sem checkbox, o menu opera apenas no arquivo clicado; uma
              // pasta precisa ser marcada para expandir sua subárvore.
              const paths = vbfSelection.countOf(vbfRoot) > 0
                ? selectedVbfPaths
                : isFile && clickedPath
                  ? [clickedPath]
                  : [];
              setCtxNode(
                vbfRoot && paths.length > 0
                  ? {
                      id: row.getAttribute('data-node-id') ?? '',
                      label: row.getAttribute('data-node-label') ?? '',
                      vbfRoot,
                      vbfPath: clickedPath,
                      selectedVbfPaths: paths,
                    }
                  : null
              );
              return;
            }
            setCtxNode(
              kind
                ? {
                    id: row.getAttribute('data-node-id') ?? '',
                    kind,
                    label: row.getAttribute('data-node-label') ?? '',
                  }
                : null
            );
          }}
        >
          {/* Os containers .vbf vêm PRIMEIRO: são a fonte da verdade e a
              árvore de data/ (imutável) é o espelho já extraído. */}
          {vbfRoots.length > 0 ? (
            <div className="px-2 pt-1 pb-1 text-xs font-medium text-muted-foreground">
              Containers .vbf · somente leitura
            </div>
          ) : null}
          {vbfRoots.map((node) => (
            <TreeItem
              key={node.id}
              node={node}
              depth={0}
              expanded={expanded}
              selectedId={selectedId}
              selectedByKind={selectedByKind}
              onToggle={actions.toggleNode}
              onSelect={(n) => void actions.selectNode(n)}
              onCheck={checkNode}
              vbfSelectionByRoot={vbfSelectionSnapshot.byRoot}
            />
          ))}
          {roots.length > 0 && vbfRoots.length > 0 ? (
            <div className="px-2 pt-3 pb-1 text-xs font-medium text-muted-foreground border-t mt-1">
              Arquivos de data/
            </div>
          ) : null}
          {roots.map((node) => (
            <TreeItem
              key={node.id}
              node={node}
              depth={0}
              expanded={expanded}
              selectedId={selectedId}
              selectedByKind={selectedByKind}
              onToggle={actions.toggleNode}
              onSelect={(n) => void actions.selectNode(n)}
              onCheck={checkNode}
              vbfSelectionByRoot={vbfSelectionSnapshot.byRoot}
            />
          ))}
        </ScrollArea>
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
