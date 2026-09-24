'use client';

import { ChevronDown, ChevronRight, FileText } from 'lucide-react';
import type { EntryKind } from '@/lib/ffx/display-names';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { EMPTY_IDS, type SideNode } from './types';

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
}) {
  const hasChildren = !!node.children && node.children.length > 0;
  const isExpanded = expanded.has(node.id);
  const isSelected = node.entry ? selectedId === `leaf:${node.entry.kind}:${node.entry.id}` : false;
  // Raiz de kind (Eventos/Sistema/Dicionário) não tem checkbox: "extrair o
  // kind inteiro" é papel do botão Exportar na linha das abas.
  const isKindRoot = node.id.startsWith('kind:');
  const selected = node.kind
    ? (selectedByKind.get(node.kind) ?? EMPTY_IDS)
    : EMPTY_IDS;

  let checked: boolean | 'indeterminate' = false;
  if (!isKindRoot) {
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

  return (
    <div>
      <div
        style={{ paddingLeft: depth * 16 }}
        className="flex items-center pr-1 pb-1"
        data-node-id={node.id}
        data-node-kind={node.kind}
      >
        {hasChildren ? (
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
        {isKindRoot ? null : (
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
          {node.entry ? <FileText size={18} className="mr-2 shrink-0" /> : null}
          <span className="truncate">{node.label}</span>
        </Button>
      </div>
      {hasChildren && isExpanded
        ? node.children!.map((child) => (
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
            />
          ))
        : null}
    </div>
  );
}
