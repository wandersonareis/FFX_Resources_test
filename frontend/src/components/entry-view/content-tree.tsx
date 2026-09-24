'use client';

import { useCallback, useMemo, useRef, useState } from 'react';
import { toast } from 'sonner';
import { RefreshCw } from 'lucide-react';
import { useSelector } from '@tanstack/react-store';
import type { EntryKind } from '@/lib/ffx/display-names';
import {
  entryKindsFor,
  exportJSON,
  exportStrings,
} from '@/lib/ffx/tree-data';
import { parseError } from '@/lib/ffx/error-handler';
import {
  EXPORT_FORMAT_LABELS,
  exportSelection,
  useExportSelection,
} from '@/lib/ffx/export-selection';
import { Button } from '@/components/ui/button';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import { ScrollArea } from '@/components/ui/scroll-area';
import type { EntryView } from './entry-view-store';
import type { SideNode } from './types';
import { TreeItem } from './tree-item';

/**
 * Sidebar da aba: árvore de kinds/grupos/arquivos com expandir, seleção de
 * exportação (checkbox tri-state), menu de contexto Exportar e navegação por
 * teclado. Estado da árvore vive no store da view; ctxNode é local.
 */
export function ContentTree({ view }: { view: EntryView }) {
  const { version, store, actions } = view;
  const roots = useSelector(store, (s) => s.roots);
  const expanded = useSelector(store, (s) => s.expanded);
  const selectedEntry = useSelector(store, (s) => s.selectedEntry);
  const [ctxNode, setCtxNode] = useState<{ id: string; kind: EntryKind } | null>(
    null
  );
  // Container da árvore, para ordenar os nós visíveis no foco por setas.
  const treeRef = useRef<HTMLDivElement>(null);

  const selection = useExportSelection();
  // Seleção de exportação por kind (checkbox tri-state + menu de contexto).
  const selectedByKind = useMemo(() => {
    const out = new Map<EntryKind, ReadonlySet<string>>();
    for (const kind of entryKindsFor(version)) {
      out.set(kind, new Set(selection.byKind.get(`${version}|${kind}`) ?? []));
    }
    return out;
  }, [selection, version]);
  const ctxCount = ctxNode ? (selectedByKind.get(ctxNode.kind)?.size ?? 0) : 0;

  // ---- Teclado: sidebar — ↑/↓ movem o foco entre nós visíveis; Enter
  // alterna expandir/fechar (grupo/raiz) ou abre o arquivo na tabela (folha).
  const onNodeKeyDown = useCallback(
    (event: React.KeyboardEvent<HTMLElement>, node: SideNode) => {
      const buttons = Array.from(
        treeRef.current?.querySelectorAll<HTMLElement>('[data-node-button]') ??
          []
      );
      const idx = buttons.indexOf(event.currentTarget);
      if (event.key === 'ArrowDown') {
        event.preventDefault();
        buttons[idx + 1]?.focus();
      } else if (event.key === 'ArrowUp') {
        event.preventDefault();
        buttons[idx - 1]?.focus();
      } else if (event.key === 'Enter') {
        event.preventDefault();
        if (node.entry) void actions.selectNode(node);
        else actions.toggleNode(node);
      } else if (event.key === 'ArrowRight') {
        event.preventDefault();
        if (node.entry) {
          // Folha: abre o arquivo e já leva o foco para a tabela (↑/↓ direto).
          actions.requestTableFocus();
          void actions.selectNode(node);
        } else if (!expanded.has(node.id)) {
          // Grupo/raiz: expande se colapsado (convenção de treeview).
          actions.toggleNode(node);
        }
      }
    },
    [actions, expanded]
  );

  // Checkbox: folha = 1 id; grupo = todos os filhos (uma notificação só).
  const checkNode = useCallback(
    (node: SideNode, checked: boolean) => {
      if (!node.kind) return;
      const ids = node.entry
        ? [node.entry.id]
        : (node.children ?? [])
            .filter((child) => child.entry)
            .map((child) => child.entry!.id);
      exportSelection.setMany(version, node.kind, ids, checked);
    },
    [version]
  );

  // Exporta só o que está marcado na árvore invocada, no FORMATO SELECIONADO
  // no topo (JSON | Strings) — leitura imperativa do store compartilhado.
  const onCtxExport = useCallback(async () => {
    const kind = ctxNode?.kind;
    if (!kind) return;
    const ids = exportSelection.idsOf(version, kind);
    if (ids.length === 0) return;
    const format = exportSelection.formatOf();
    const label = EXPORT_FORMAT_LABELS[format];
    const toastId = 'export';
    toast.loading(`Exportando (${label})…`, { id: toastId });
    try {
      const paths =
        format === 'json'
          ? await exportJSON(kind, version, ids)
          : await exportStrings(kind, version, ids);
      toast.success(
        `Exportado (${label}): ${(paths ?? []).length} arquivo(s).`,
        { id: toastId }
      );
    } catch (error) {
      toast.error(parseError(error), { id: toastId });
    }
  }, [ctxNode, version]);

  return (
    <aside className="w-70 shrink-0 border-r p-2 flex flex-col min-h-0">
      <div className="flex items-center justify-between font-semibold px-2 py-1">
        <span>Conteúdo</span>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => void actions.reload()}
          aria-label="Recarregar"
        >
          <RefreshCw size={20} />
        </Button>
      </div>
      <ContextMenu
        open={ctxNode !== null && ctxCount > 0}
        onOpenChange={(open) => {
          if (!open) setCtxNode(null);
        }}
      >
        <ContextMenuTrigger asChild>
          <ScrollArea
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
              setCtxNode(
                kind
                  ? { id: row.getAttribute('data-node-id') ?? '', kind }
                  : null
              );
            }}
          >
            {roots.map((node) => (
              <TreeItem
                key={node.id}
                node={node}
                depth={0}
                expanded={expanded}
                selectedId={selectedEntry ? `leaf:${selectedEntry.kind}:${selectedEntry.id}` : null}
                selectedByKind={selectedByKind}
                onToggle={actions.toggleNode}
                onSelect={(n) => void actions.selectNode(n)}
                onCheck={checkNode}
                onNodeKeyDown={onNodeKeyDown}
              />
            ))}
          </ScrollArea>
        </ContextMenuTrigger>
        <ContextMenuContent>
          <ContextMenuItem onSelect={() => void onCtxExport()}>
            Exportar
          </ContextMenuItem>
        </ContextMenuContent>
      </ContextMenu>
    </aside>
  );
}
