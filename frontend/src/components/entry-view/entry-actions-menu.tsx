'use client';

import { useEffect, useState, type ReactNode } from 'react';
import { loggedToast as toast } from '@/lib/ffx/toast-logged';
import { Copy, Download, FolderOpen, Trash2 } from 'lucide-react';
import type { dto } from '@/wailsjs/go/models';
import { resolveEntryLabel, type EntryKind } from '@/lib/ffx/display-names';
import { imageDuplicates, revealEntry } from '@/lib/ffx/tree-data';
import { parseError } from '@/lib/ffx/error-handler';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuLabel,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import type { EntryView } from './entry-view-store';

/** O nó sob o cursor: folha tem id (arquivo); grupo/raiz, não. */
export type EntryMenuTarget = {
  kind: EntryKind;
  id: string | null;
  label: string;
};

/**
 * Menu de contexto compartilhado da árvore e do painel de imagem — MESMO
 * menu nos dois lugares.
 *
 * Exportar (imagem: o diálogo de extração; texto: o que o chamador marcou),
 * Abrir até o arquivo e Deletar vêm em todo kind; Replicar só aparece para
 * imagem e Deletar só HABILITA para imagem (apagar binário de texto não tem
 * escopo). O alvo é o NÓ CLICADO, não a entry selecionada.
 *
 * Controlado pelo chamador: `target === null` = fechado. As cópias do alvo
 * vêm do painel quando ele as tem (`initialDuplicates`), senão são buscadas
 * aqui num binding leve (índice memoizado, sem decode de imagem) — é o que
 * habilita Replicar e põe a contagem no rótulo.
 */
export function EntryActionsMenu({
  view,
  target,
  onOpenChange,
  initialId,
  initialDuplicates,
  onExport,
  children,
}: {
  view: EntryView;
  target: EntryMenuTarget | null;
  onOpenChange: (open: boolean) => void;
  /** Cópias já conhecidas pelo chamador (id a que pertencem). */
  initialId?: string | null;
  initialDuplicates?: dto.ImageDuplicate[] | null;
  /** Exportar de TEXTO — a lógica de seleção/formato fica no chamador. */
  onExport?: () => void;
  children: ReactNode;
}) {
  const { version, actions } = view;
  const [fetched, setFetched] = useState<{
    id: string;
    list: dto.ImageDuplicate[];
  } | null>(null);

  const targetId = target?.id ?? null;
  const isImage = target?.kind === 'images';

  const duplicates =
    initialId != null && initialId === targetId && initialDuplicates
      ? initialDuplicates
      : fetched && fetched.id === targetId
        ? fetched.list
        : null;

  useEffect(() => {
    if (!targetId || !isImage) return;
    if (initialId != null && initialId === targetId && initialDuplicates) return;
    if (fetched?.id === targetId) return;
    let dead = false;
    imageDuplicates(targetId, version)
      .then((res) => {
        if (!dead) setFetched({ id: targetId, list: res.duplicates ?? [] });
      })
      .catch((error) => {
        if (dead) return;
        // Sem índice o menu não some: segue com "sem cópias" (Replicar
        // desabilitado) e o erro aparece.
        setFetched({ id: targetId, list: [] });
        toast.error(parseError(error));
      });
    return () => {
      dead = true;
    };
  }, [
    targetId,
    isImage,
    initialId,
    initialDuplicates,
    fetched,
    version,
  ]);

  const canFile = targetId !== null;
  const canExtract = isImage && canFile;
  const loading = canExtract && duplicates === null;
  const copies = duplicates?.length ?? 0;
  const canReplicate = canExtract && !loading && copies > 0;
  const canDelete = isImage && canFile;

  const openAction = (type: 'extract' | 'replicate' | 'delete') => {
    if (!targetId || !target) return;
    actions.openImageAction({
      type,
      id: targetId,
      label: target.label,
      duplicates,
    });
  };

  const onReveal = () => {
    if (!targetId || !target) return;
    void revealEntry(target.kind, targetId, version)
      .then(() => toast.success(`Exibido no Explorer: ${target.label}`))
      .catch((error) => toast.error(parseError(error)));
  };

  const replicateLabel = loading
    ? 'Carregando cópias…'
    : copies > 0
      ? `Replicar para ${copies} cópia${copies > 1 ? 's' : ''}`
      : 'Replicar';

  return (
    <ContextMenu open={target !== null} onOpenChange={onOpenChange}>
      <ContextMenuTrigger asChild>{children}</ContextMenuTrigger>
      <ContextMenuContent>
        {target?.label ? (
          <ContextMenuLabel>{target.label}</ContextMenuLabel>
        ) : null}
        <ContextMenuItem
          disabled={isImage && !canExtract}
          onSelect={() => (isImage ? openAction('extract') : onExport?.())}
        >
          <Download size={16} />
          Exportar
        </ContextMenuItem>
        <ContextMenuItem disabled={!canFile} onSelect={onReveal}>
          <FolderOpen size={16} />
          Abrir até o arquivo
        </ContextMenuItem>
        {isImage ? (
          <>
            <ContextMenuItem
              disabled={!canReplicate}
              onSelect={() => openAction('replicate')}
            >
              <Copy size={16} />
              {replicateLabel}
            </ContextMenuItem>
            <ContextMenuSeparator />
          </>
        ) : null}
        <ContextMenuItem
          variant="destructive"
          disabled={!canDelete}
          onSelect={() => openAction('delete')}
        >
          <Trash2 size={16} />
          Deletar
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
}

/** Nó da árvore → alvo do menu (folha carrega o id, grupo não). */
export function menuTargetOf(
  node: { id: string; kind: EntryKind; label?: string } | null
): EntryMenuTarget | null {
  if (!node) return null;
  const prefix = `leaf:${node.kind}:`;
  const isLeaf = node.id.startsWith(prefix);
  const id = isLeaf ? node.id.slice(prefix.length) : null;
  return {
    kind: node.kind,
    id,
    label: id ? resolveEntryLabel(node.kind, id) : (node.label ?? ''),
  };
}
