'use client';

import {
  ChevronDown,
  ChevronRight,
  File as FileIcon,
  FileText,
  Folder,
  FolderOpen,
  Image as ImageIcon,
  Loader2,
} from 'lucide-react';
import type { EntryKind } from '@/lib/ffx/display-names';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  vbfNodeCheckState,
  type VbfSelectionTreeNode,
} from '@/lib/ffx/vbf-selection';
import { EMPTY_IDS, entryNodeId, type SideNode } from './types';

export function TreeItem({
  node,
  depth,
  expanded,
  selectedId,
  selectedByKind,
  onToggle,
  onSelect,
  onCheck,
  onNodeKeyDown,
  vbfSelectionByRoot,
}: {
  node: SideNode;
  depth: number;
  expanded: Set<string>;
  selectedId: string | null;
  selectedByKind: ReadonlyMap<EntryKind, ReadonlySet<string>>;
  onToggle: (node: SideNode) => void;
  onSelect: (node: SideNode) => void;
  onCheck: (node: SideNode, checked: boolean) => void;
  onNodeKeyDown: (event: React.KeyboardEvent<HTMLElement>, node: SideNode) => void;
  vbfSelectionByRoot: ReadonlyMap<string, readonly string[]>;
}) {
  // Diretório do .vbf ainda sem filhos (ou grupo de macrodic) também é
  // expansível: o chevron aparece ANTES da listagem chegar.
  const hasChildren =
    (!!node.children && node.children.length > 0) || (node.expandable ?? false);
  const isExpanded = expanded.has(node.id);
  const isSelected = node.entry ? selectedId === entryNodeId(node.entry) : false;
  // Raiz de kind (Eventos/Sistema/Dicionário) não tem checkbox: "extrair o
  // kind inteiro" é papel do botão Exportar na linha das abas.
  const isKindRoot = node.id.startsWith('kind:');
  // Imagem não é texto: não entra na seleção de exportação (JSON/.strings).
  const isImage = node.kind === 'images';
  // O container .vbf é somente leitura, mas seus caminhos podem ser
  // selecionados para extração/exportação sem alterar o arquivo.
  const isVbf = node.vbf === true;
  // Raiz do container .vbf não tem checkbox (igual à raiz de kind):
  // extrair o container inteiro não faz sentido — marque diretórios/arquivos.
  const isVbfContainerRoot = isVbf && !node.entry && (node.vbfPath ?? '') === '';
  const vbfRoot = node.vbfRoot ?? node.entry?.vbf?.root;
  const vbfPath = node.vbfPath ?? node.entry?.vbf?.path ?? '';
  const selected = node.kind
    ? (selectedByKind.get(node.kind) ?? EMPTY_IDS)
    : EMPTY_IDS;

  let checked: boolean | 'indeterminate' = false;
  if (!isKindRoot && !isImage && !isVbf) {
    if (node.entry) {
      checked = selected.has(node.entry.id);
    } else {
      // Grupo: tri-state sobre os ids dos filhos (raiz de evento tem folhas).
      const childIds = (node.children ?? [])
        .filter((child) => child.entry)
        .map((child) => child.entry!.id);
      const count = childIds.filter((id) => selected.has(id)).length;
      checked =
        count === 0 ? false : count === childIds.length ? true : 'indeterminate';
    }
  }

  if (isVbf && !isVbfContainerRoot) {
    const paths = vbfRoot ? (vbfSelectionByRoot.get(vbfRoot) ?? []) : [];
    const asSelectionNode = (child: SideNode): VbfSelectionTreeNode => ({
      path: child.vbfPath ?? child.entry?.vbf?.path ?? '',
      children: child.children?.map(asSelectionNode),
    });
    checked = vbfNodeCheckState(
      paths,
      vbfPath,
      node.children?.map(asSelectionNode)
    );
  }

  return (
    <div>
      <div
        style={{ paddingLeft: depth * 16 }}
        className="flex items-center pr-1 pb-1"
        data-node-id={node.id}
        data-node-kind={node.kind}
        data-node-label={node.label}
        data-node-vbf={isVbf ? 'true' : undefined}
        data-node-vbf-root={vbfRoot}
        data-node-vbf-path={vbfPath}
      >
        {node.loading ? (
          <span
            className="w-8 shrink-0 grid place-items-center opacity-60"
            aria-label={`Carregando ${node.label}`}
          >
            <Loader2 size={16} className="animate-spin" />
          </span>
        ) : hasChildren ? (
          <Button
            variant="ghost"
            size="icon"
            className="size-8 shrink-0"
            onClick={() => onToggle(node)}
            aria-label={`Alternar ${node.label}`}
          >
            {isExpanded ? <ChevronDown size={20} /> : <ChevronRight size={20} />}
          </Button>
        ) : (
          <span className="w-8 shrink-0" />
        )}
        {isKindRoot || isVbfContainerRoot || (!isVbf && isImage) ? null : (
          <Checkbox
            className="mr-1"
            checked={checked}
            onCheckedChange={(value) => onCheck(node, value === true)}
            aria-label={`Selecionar ${node.label}`}
          />
        )}
        <Button
          variant="ghost"
          size="sm"
          data-node-button
          className={`w-full justify-start text-left min-w-0 focus-visible:bg-sky-200 focus-visible:ring-1 focus-visible:ring-sky-400 focus-visible:outline-none ${
            isSelected ? 'bg-sky-100 ring-1 ring-sky-400' : ''
          }`}
          onClick={() => onSelect(node)}
          onKeyDown={(event) => onNodeKeyDown(event, node)}
        >
          {/* Ícone do nó: diretório FECHADO quando colapsado e ABERTO quando
              expandido; arquivo de texto para binário com texto; imagem para
              .dds.phyre; arquivo genérico para formato fora do escopo. */}
          {node.entry ? (
            node.entry.kind === 'images' ? (
              <ImageIcon size={18} className="mr-2 shrink-0" />
            ) : (
              <FileText size={18} className="mr-2 shrink-0" />
            )
          ) : node.unsupported ? (
            <FileIcon size={18} className="mr-2 shrink-0 opacity-70" />
          ) : isExpanded ? (
            <FolderOpen size={18} className="mr-2 shrink-0" />
          ) : (
            <Folder size={18} className="mr-2 shrink-0" />
          )}
          <span className="truncate">{node.label}</span>
        </Button>
      </div>
      {hasChildren && isExpanded
        ? (node.children ?? []).map((child) => (
            <TreeItem
              key={child.id}
              node={child}
              depth={depth + 1}
              expanded={expanded}
              selectedId={selectedId}
              selectedByKind={selectedByKind}
              onToggle={onToggle}
              onSelect={onSelect}
              onCheck={onCheck}
              onNodeKeyDown={onNodeKeyDown}
              vbfSelectionByRoot={vbfSelectionByRoot}
            />
          ))
        : null}
    </div>
  );
}
