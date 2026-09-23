'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';
import {
  createColumnHelper,
  tableFeatures,
  useTable,
} from '@tanstack/react-table';
import { ChevronDown, ChevronRight, FileText, RefreshCw, Save } from 'lucide-react';
import { dto } from '@/wailsjs/go/models';
import { ListLanguages, SetUnsavedEdits } from '@/wailsjs/go/main/App';
import { KIND_LABELS, EntryKind } from '@/lib/ffx/display-names';
import { eventGroupLabel, shortenedOf } from '@/lib/ffx/event-group-names';
import type { GameVersionId } from '@/lib/ffx/game-version';
import {
  EntryRow,
  entryKindsFor,
  exportJSON,
  exportStrings,
  loadEntry,
  loadKindEntries,
} from '@/lib/ffx/tree-data';
import { useEditDraft } from '@/lib/ffx/edit-draft';
import { parseError, sendErrorNotification } from '@/lib/ffx/error-handler';
import {
  EXPORT_FORMAT_LABELS,
  exportSelection,
  useExportSelection,
} from '@/lib/ffx/export-selection';
import { useWailsEvent } from '@/lib/ffx/use-wails-event';
import { SOURCE_LANG, saveAllDrafts } from '@/lib/ffx/save-all';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import { GameTextView } from '@/components/game-text-view';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { TranslationCellDialog } from '@/components/dialogs/translation-cell-dialog';

interface SideNode {
  id: string;
  label: string;
  kind?: EntryKind;
  entry?: EntryRow;
  children?: SideNode[];
}

const features = tableFeatures({});

type F = typeof features;

const EMPTY_IDS: ReadonlySet<string> = new Set<string>();

const columnHelper = createColumnHelper<F, dto.TextRow>();

export function GameVersionTab({ version }: { version: GameVersionId }) {
  const { store: drafts, snapshot } = useEditDraft();

  const [activeKind, setActiveKind] = useState<EntryKind>('events');
  const [selectedEntry, setSelectedEntry] = useState<EntryRow | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [roots, setRoots] = useState<SideNode[]>([]);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [rows, setRows] = useState<dto.TextRow[]>([]);
  const [languages, setLanguages] = useState<Array<{ code: string; name: string }>>([]);
  const [translationRow, setTranslationRow] = useState<dto.TextRow | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [ctxNode, setCtxNode] = useState<{ id: string; kind: EntryKind } | null>(
    null
  );

  const hasDirty = snapshot.hasDirty;

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

  useEffect(() => {
    void SetUnsavedEdits(hasDirty);
  }, [hasDirty]);

  useEffect(() => {
    ListLanguages()
      .then((v) => setLanguages(v ?? [{ code: SOURCE_LANG, name: 'English' }]))
      .catch(() => setLanguages([{ code: SOURCE_LANG, name: 'English' }]));
  }, []);

  const selectEntry = useCallback(
    async (entry: EntryRow) => {
      setLoading(true);
      try {
        setActiveKind(entry.kind);
        setSelectedEntry(entry);
        const full = await loadEntry(entry.kind, entry.id, version);
        drafts.setBase(version, entry.kind, entry.id, full);
        setRows([...(full.rows ?? [])]);
      } catch (error) {
        sendErrorNotification(error);
        setSelectedEntry(null);
        setRows([]);
      } finally {
        setLoading(false);
      }
    },
    [version, drafts]
  );

  const eventGroups = useCallback(
    (entries: EntryRow[]): SideNode[] => {
      const groups = new Map<string, EntryRow[]>();
      for (const entry of entries) {
        const short = shortenedOf(entry.id);
        const list = groups.get(short) ?? [];
        list.push(entry);
        groups.set(short, list);
      }
      const nodes: SideNode[] = [];
      for (const short of [...groups.keys()].sort()) {
        const files = groups.get(short) ?? [];
        const shortKind = 'events' as EntryKind;
        nodes.push({
          id: `group:events:${short}`,
          label: `${eventGroupLabel(version, short)} (${files.length})`,
          kind: shortKind,
          children: files.map((entry) => ({
            id: `leaf:events:${entry.id}`,
            label: entry.label,
            kind: entry.kind,
            entry,
          })),
        });
      }
      return nodes;
    },
    [version]
  );

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      const nextRoots: SideNode[] = [];
      const nextExpanded = new Set<string>();
      for (const kind of entryKindsFor(version)) {
        const entries = await loadKindEntries(kind, version);
        nextRoots.push({
          id: `kind:${kind}`,
          label: `${KIND_LABELS[kind]} (${entries.length})`,
          kind,
          children:
            kind === 'events'
              ? eventGroups(entries)
              : entries.map((entry) => ({
                  id: `leaf:${kind}:${entry.id}`,
                  label: entry.label,
                  kind: entry.kind,
                  entry,
                })),
        });
      }
      for (const node of nextRoots) {
        nextExpanded.add(node.id);
        for (const child of node.children ?? []) nextExpanded.add(child.id);
      }
      setRoots(nextRoots);
      setExpanded(nextExpanded);
      const current = selectedEntry;
      if (current) {
        const entries = await loadKindEntries(current.kind, version);
        const found = entries.find((e) => e.id === current.id);
        if (!found) {
          setSelectedEntry(null);
          setRows([]);
        } else {
          await selectEntry(found);
        }
      }
    } catch (error) {
      sendErrorNotification(error);
    } finally {
      setLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [version]);

  useEffect(() => {
    // Sincroniza com o backend ao montar/trocar de versão.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void reload();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [version]);
  useEffect(
    () => drafts.onSaved(() => void reload()),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [version]
  );

  // Após importar, o backend emite ImportDone → recarrega árvore e tabela.
  useWailsEvent('ImportDone', () => void reload());

  const selectNode = useCallback(
    async (node: SideNode) => {
      if (node.entry && node.kind) {
        await selectEntry(node.entry);
      } else if (node.kind) {
        setActiveKind(node.kind);
      }
    },
    [selectEntry]
  );

  const toggleNode = (node: SideNode) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(node.id)) next.delete(node.id);
      else next.add(node.id);
      return next;
    });
  };

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
  const columns = useMemo(
    () =>
      columnHelper.columns([
        columnHelper.accessor('index', {
          header: '#',
          cell: (info) => <span className="text-muted-foreground">{info.getValue()}</span>,
        }),
        ...(activeKind === 'objects'
          ? [columnHelper.accessor('name', { header: 'Nome' })]
          : []),
        columnHelper.accessor((row) => row.text?.[SOURCE_LANG] ?? '', {
          id: 'original',
          header: 'Original',
          cell: (info) => <GameTextView text={info.getValue()} />,
        }),
        columnHelper.accessor(
          (row) => {
            const entry = selectedEntry;
            const edited = entry
              ? drafts.editOf(version, entry.kind, entry.id, row, SOURCE_LANG)
              : undefined;
            return edited ?? row.text?.[SOURCE_LANG] ?? '';
          },
          {
            id: 'translated',
            header: 'Traduzido',
            cell: (info) => {
              const row = info.row.original;
              const entry = selectedEntry;
              const edited = entry
                ? drafts.editOf(version, entry.kind, entry.id, row, SOURCE_LANG) !==
                  undefined
                : false;
              return (
                <div
                  className={`text-cell translated-cell${edited ? ' edited' : ''}`}
                  onClick={() => {
                    setTranslationRow(row);
                    setDialogOpen(true);
                  }}
                >
                  <GameTextView text={info.getValue()} />
                </div>
              );
            },
          }
        ),
      ]),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [activeKind, selectedEntry, version, snapshot.revision]
  );

  const table = useTable({
    features,
    columns,
    data: rows,
  });

  const onSaveAll = () => {
    void saveAllDrafts(() => saving, setSaving);
  };

  return (
    <div className="flex h-full min-h-0 border-t">
      <aside className="w-80 shrink-0 border-r p-2 flex flex-col min-h-0">
        <div className="flex items-center justify-between font-semibold px-2 py-1">
          <span>Conteúdo</span>
          <Button variant="ghost" size="icon" onClick={() => void reload()} aria-label="Recarregar">
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
                  onToggle={toggleNode}
                  onSelect={(n) => void selectNode(n)}
                  onCheck={checkNode}
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

      <main className="flex-1 min-w-0 p-3 px-4 overflow-auto">
        <div className="flex items-center justify-between gap-4">
          <h3 className="text-lg font-semibold">
            {KIND_LABELS[activeKind]}
            {selectedEntry ? (
              <span className="font-normal opacity-70"> · {selectedEntry.label}</span>
            ) : null}
          </h3>
          <Button disabled={!hasDirty || saving} onClick={() => void onSaveAll()}>
            <Save size={18} />
            {saving ? 'Salvando…' : 'Salvar'}
          </Button>
        </div>

        {selectedEntry ? (
          <div className="mt-2">
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead
                        key={header.id}
                        className={header.column.id === 'index' ? 'w-14' : undefined}
                      >
                        {header.isPlaceholder ? null : <table.FlexRender header={header} />}
                      </TableHead>
                    ))}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows.map((row) => (
                  <TableRow key={row.id}>
                    {row.getAllCells().map((cell) => (
                      <TableCell
                        key={cell.id}
                        className={cell.column.id === 'original' ? 'text-cell' : undefined}
                      >
                        <table.FlexRender cell={cell} />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : loading ? (
          <p className="mt-8 opacity-70">Carregando…</p>
        ) : (
          <p className="mt-8 opacity-70">Selecione um arquivo no sidebar.</p>
        )}
      </main>

      {/* Sem key: a instância persiste e a navegação troca só o conteúdo */}
      <TranslationCellDialog
        open={dialogOpen}
        row={translationRow}
        languages={languages}
        // Navegação limitada ao arquivo atual (rows da entrada carregada).
        hasPrevious={rows.findIndex((r) => r.index === translationRow?.index) > 0}
        hasNext={
          rows.findIndex((r) => r.index === translationRow?.index) <
          rows.length - 1
        }
        onNavigate={(direction, value) => {
          const entry = selectedEntry;
          if (value !== undefined && entry && translationRow) {
            drafts.setCell(version, entry.kind, entry.id, translationRow, SOURCE_LANG, value);
            setRows((prev) => [...prev]);
          }
          const idx = rows.findIndex((r) => r.index === translationRow?.index);
          const next = rows[idx + (direction === 'next' ? 1 : -1)];
          if (next) setTranslationRow(next); // sem fechar: a key remonta o editor
        }}
        onClosed={(value) => {
          const entry = selectedEntry;
          if (value !== undefined && entry) {
            drafts.setCell(version, entry.kind, entry.id, translationRow!, SOURCE_LANG, value);
            setRows((prev) => [...prev]);
          }
          setDialogOpen(false);
          setTranslationRow(null);
        }}
      />
    </div>
  );
}

function TreeItem({
  node,
  depth,
  expanded,
  selectedId,
  selectedByKind,
  onToggle,
  onSelect,
  onCheck,
}: {
  node: SideNode;
  depth: number;
  expanded: Set<string>;
  selectedId: string | null;
  selectedByKind: ReadonlyMap<EntryKind, ReadonlySet<string>>;
  onToggle: (node: SideNode) => void;
  onSelect: (node: SideNode) => void;
  onCheck: (node: SideNode, checked: boolean) => void;
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
        className="flex items-center"
        data-node-id={node.id}
        data-node-kind={node.kind}
      >
        {hasChildren ? (
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 shrink-0"
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
          className={`w-full justify-start text-left min-w-0${isSelected ? ' bg-accent' : ''}`}
          onClick={() => onSelect(node)}
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
            />
          ))
        : null}
    </div>
  );
}
