'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useHotkey, useHotkeys } from '@tanstack/react-hotkeys';
import { loggedToast as toast } from '@/lib/ffx/toast-logged';
import { SHORTCUTS, shortcutTip } from '@/lib/ffx/shortcuts';
import { openSearchDialog } from '@/lib/ffx/search-store';
import { ChevronDown, Download, Settings, Upload } from 'lucide-react';
import { QuitApp } from '@/wailsjs/go/main/App';
import { EventsEmit } from '@/wailsjs/runtime/runtime';
import { dto } from '@/wailsjs/go/models';
import { GAME_VERSIONS, GameVersionId, isGameVersionId } from '@/lib/ffx/game-version';
import { tabStorage } from '@/lib/ffx/tab-storage';
import { useWailsEvent } from '@/lib/ffx/use-wails-event';
import { saveAllDrafts } from '@/lib/ffx/save-all';
import { EntryKind, KIND_LABELS } from '@/lib/ffx/display-names';
import {
  entryKindsFor,
  exportDestination,
  exportJSON,
  exportStrings,
  importFile,
  previewImport,
  selectImportFile,
} from '@/lib/ffx/tree-data';
import {
  EXPORT_FORMAT_LABELS,
  exportSelection,
  ExportFormat,
  useExportSelection,
} from '@/lib/ffx/export-selection';
import { parseError } from '@/lib/ffx/error-handler';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { GameVersionTab } from '@/components/game-version-tab';
import { StatusBar } from '@/components/status-bar/status-bar';
import { ConfigDialog } from '@/components/dialogs/config-dialog';
import { ImportSummaryDialog } from '@/components/dialogs/import-summary-dialog';
import { ProgressDialog } from '@/components/dialogs/progress-dialog';

/** Cap da fila de toasts de ciclo aberto (descarta o mais antigo). */
const NOTIFY_QUEUE_CAP = 50;

type NotifyPayload = { severity?: string; message?: string };

/** Rende uma notificação por severidade (error: sticky; demais: 4 s). */
function emitNotify(payload: NotifyPayload): void {
  if (payload?.severity === 'error') {
    toast.error(payload?.message ?? 'Notificação');
  } else {
    toast(payload?.message ?? 'Notificação', { duration: 4000 });
  }
}

/** Conteúdo do toast de export: rótulo + barra embutida + contagem. */
function ExportToastContent({
  label,
  processed,
  total,
  value,
}: {
  label: string;
  processed: number;
  total: number;
  value: number;
}) {
  const count = total > 0 ? ` ${processed}/${total}` : '';
  return (
    <div className="w-56">
      <p className="text-sm">{label}</p>
      <Progress value={value} className="mt-1.5 h-1.5" />
      <p className="mt-1 text-xs opacity-70">{count.trim()}</p>
    </div>
  );
}

export function AppShell() {  const [selectedIndex, setSelectedIndex] = useState(0);
  const [saving, setSaving] = useState(false);
  const [configOpen, setConfigOpen] = useState(false);
  const [progress, setProgress] = useState({
    open: false,
    value: 0,
    label: '',
    processed: 0,
    total: 0,
    issueCount: 0,
    complete: false,
  });
  // Guard do toast de export: só consome Progress enquanto um export roda
  // (evita pegar progresso de carga de árvore alheia).
  const exportProgress = useRef({ active: false, processed: 0, total: 0 });
  // Modal de progresso aberto bloqueia interação fora do diálogo (o radix
  // modal põe pointer-events: none no body): notificações que chegarem
  // durante um ciclo vão para a fila e aparecem quando o usuário fecha o
  // modal de resultado.
  const progressOpen = useRef(false);
  const notifyQueue = useRef<NotifyPayload[]>([]);
  // Escopo de exportação por versão (kinds marcados no dropdown Exportar).
  const [scopes, setScopes] = useState<Record<string, EntryKind[]>>(() =>
    Object.fromEntries(GAME_VERSIONS.map((tab) => [tab.id, entryKindsFor(tab.id)]))
  );
  const [importSummary, setImportSummary] = useState<dto.ImportSummary | null>(
    null
  );
  const [importing, setImporting] = useState(false);

  const selection = useExportSelection();
  const exportFormat = selection.format;

  const closeProgress = useCallback(() => {
    progressOpen.current = false;
    setProgress({
      open: false,
      value: 0,
      label: '',
      processed: 0,
      total: 0,
      issueCount: 0,
      complete: false,
    });
    if (notifyQueue.current.length > 0) {
      const pending = notifyQueue.current;
      notifyQueue.current = [];
      for (const q of pending) emitNotify(q);
    }
  }, []);

  useEffect(() => {
    // Hydration-safe: localStorage só existe no cliente.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setSelectedIndex(tabStorage.load(GAME_VERSIONS.length));
  }, []);

  const setVersion = useCallback((version: GameVersionId) => {
    EventsEmit('GameVersionChanged', version);
    EventsEmit('Refresh_Tree');
  }, []);

  // ---- Atalhos globais (TanStack Hotkeys) ----
  const switchTab = (index: number) => {
    if (index < 0 || index >= GAME_VERSIONS.length || index === selectedIndex)
      return;
    setSelectedIndex(index);
    tabStorage.save(index);
    setVersion(GAME_VERSIONS[index].id);
  };

  // A tecla vem do catálogo lib/ffx/shortcuts.ts, que é a mesma fonte do
  // tooltip de cada controle (nenhuma tecla é escrita em dois lugares).
  // Ctrl+1..4: trocar aba de versão — uma tecla por GAME_VERSIONS (a lista
  // deriva do catálogo: acrescentar versão é acrescentar uma tecla ali).
  useHotkeys(
    SHORTCUTS.versionTabs.keys.map((hotkey, index) => ({
      hotkey,
      callback: () => switchTab(index),
      options: {
        meta: {
          name: `${SHORTCUTS.versionTabs.label}: ${GAME_VERSIONS[index].label}`,
          group: 'Abas',
        },
      },
    }))
  );
  // Ctrl+K: abrir o modal de busca da aba ATIVA (convenção de palette; o
  // store da busca é de módulo por versão, então o atalho só aponta o id).
  useHotkey(
    SHORTCUTS.search.keys,
    () => openSearchDialog(GAME_VERSIONS[selectedIndex].id),
    { meta: { name: SHORTCUTS.search.label, group: 'Global' } }
  );
  // Ctrl+, : abrir Configurações (convenção de settings).
  useHotkey(SHORTCUTS.config.keys, () => setConfigOpen(true), {
    meta: { name: SHORTCUTS.config.label, group: 'Global' },
  });
  // Ctrl+Alt+F: alterna formato JSON ↔ Strings direto (popover não abre).
  // Fora do catálogo: não é exibido em texto nenhum.
  useHotkey('Mod+Alt+F', () => {
    const next =
      exportSelection.formatOf() === 'json' ? 'strings' : 'json';
    exportSelection.setFormat(next);
    toast.info(`Formato de exportação: ${EXPORT_FORMAT_LABELS[next]}`);
  });
  // Ctrl+S: salvar rascunhos (mesma ação do botão Salvar; sem fechar o app).
  useHotkey(
    SHORTCUTS.save.keys,
    () => {
      void saveAllDrafts(() => saving, setSaving);
    },
    { meta: { name: SHORTCUTS.save.label, group: 'Global' } }
  );

  useWailsEvent('Notify', (data) => {
    const payload = data as NotifyPayload;
    // Processando em andamento: acumula (o overlay bloqueia interação) e
    // aparece sobreposto quando o ciclo termina.
    if (progressOpen.current) {
      if (notifyQueue.current.length >= NOTIFY_QUEUE_CAP) {
        notifyQueue.current.shift();
      }
      notifyQueue.current.push(payload ?? { message: 'Notificação' });
      return;
    }
    emitNotify(payload);
  });

  useWailsEvent('ShowProgress', (visible) => {
    if (Boolean(visible)) {
      progressOpen.current = true;
      setProgress({
        open: true,
        value: 0,
        label: '',
        processed: 0,
        total: 0,
        issueCount: 0,
        complete: false,
      });
      return;
    }
    // O modal permanece visível após o fim para mostrar o resultado e só
    // libera o app quando o usuário o fecha.
    setProgress((p) => ({ ...p, complete: true }));
  });

  useWailsEvent('Progress', (data) => {
    const payload = data as {
      label?: string;
      percentage?: number;
      processed?: number;
      total?: number;
      issueCount?: number;
      done?: boolean;
    };
    setProgress((p) => ({
      ...p,
      value: payload?.percentage ?? p.value,
      label: payload?.label ?? p.label,
      processed: payload?.processed ?? p.processed,
      total: payload?.total ?? p.total,
      issueCount: payload?.issueCount ?? p.issueCount,
      complete: payload?.done ?? p.complete,
    }));
    // Toast de export com barra embutida: re-render a cada Progress com a
    // contagem corrente (guard: só quando um export está ativo).
    if (exportProgress.current.active) {
      toast.loading(
        <ExportToastContent
          label={payload?.label ?? ''}
          processed={payload?.processed ?? 0}
          total={payload?.total ?? 0}
          value={payload?.percentage ?? 0}
        />,
        { id: 'export' }
      );
    }
  });

  useWailsEvent('GameVersion', (data) => {
    if (isGameVersionId(data)) {
      const index = GAME_VERSIONS.findIndex((t) => t.id === data);
      if (index >= 0 && index !== selectedIndex) {
        setSelectedIndex(index);
        tabStorage.save(index);
      }
    }
  });

  useWailsEvent('SaveRequested', () => {
    void saveAllDrafts(() => saving, setSaving).then(() => QuitApp());
  });

  const activeVersion = GAME_VERSIONS[selectedIndex].id;
  const servedKinds = entryKindsFor(activeVersion);
  const activeScope = scopes[activeVersion] ?? servedKinds;

  const toggleScope = (kind: EntryKind) => {
    setScopes((prev) => {
      const current = prev[activeVersion] ?? entryKindsFor(activeVersion);
      const next = current.includes(kind)
        ? current.filter((k) => k !== kind)
        : [...current, kind];
      return { ...prev, [activeVersion]: next };
    });
  };

  // Exporta os kinds no FORMATO SELECIONADO (JSON | Strings), em sequência
  // (evita GetCollection concorrente). useCheckedIds=false força ids=[] →
  // tudo do kind; com seleção, exigirá ≥1 marcado.
  const runExport = useCallback(
    async (kinds: EntryKind[], useCheckedIds: boolean) => {
      const version = GAME_VERSIONS[selectedIndex].id;
      const format = exportSelection.formatOf();
      if (useCheckedIds && kinds.every((kind) => exportSelection.countOf(version, kind) === 0)) {
        toast.info('Nada para exportar — selecione itens na árvore.');
        return;
      }
      const label = EXPORT_FORMAT_LABELS[format];
      const toastId = 'export';
      exportProgress.current = { active: true, processed: 0, total: 0 };
      toast.loading(
        <ExportToastContent label={`Exportando (${label})…`} processed={0} total={0} value={0} />,
        { id: toastId }
      );
      try {
        let written = 0;
        const allPaths: string[] = [];
        for (const kind of kinds) {
          const ids = useCheckedIds ? exportSelection.idsOf(version, kind) : [];
          // Exportar seleção = atalho do Exportar do contexto.
          if (useCheckedIds && ids.length === 0) continue;
          const paths =
            format === 'json'
              ? await exportJSON(kind, version, ids)
              : await exportStrings(kind, version, ids);
          allPaths.push(...(paths ?? []));
          written += (paths ?? []).length;
        }
        exportProgress.current.active = false;
        if (written === 0) {
          toast.info('Nada para exportar.', { id: toastId });
        } else {
          const dest = exportDestination(allPaths);
          toast.success(
            `Exportado (${label}): ${dest || `${written} arquivo(s)`}.`,
            { id: toastId }
          );
        }
      } catch (error) {
        toast.error(parseError(error), { id: toastId });
      } finally {
        exportProgress.current.active = false;
      }
    },
    [selectedIndex]
  );

  const onImportClick = useCallback(async () => {
    const version = GAME_VERSIONS[selectedIndex].id;
    try {
      const path = await selectImportFile();
      if (!path) return;
      setImportSummary(await previewImport(path, version));
    } catch (error) {
      toast.error(parseError(error));
    }
  }, [selectedIndex]);

  const onConfirmImport = useCallback(async () => {
    const current = importSummary;
    if (!current) return;
    const version = GAME_VERSIONS[selectedIndex].id;
    setImporting(true);
    const toastId = 'import';
    toast.loading('Importando…', { id: toastId });
    try {
      const changed = await importFile(current.path, version);
      setImportSummary(null);
      toast.success(
        `Importação concluída — ${changed} texto(s) atualizado(s) e salvos em binário (mods/).`,
        { id: toastId }
      );
      // Avisa a aba ativa para recarregar árvore/tabela (rascunhos preservados).
      try {
        EventsEmit('ImportDone');
      } catch {
        // sem runtime Wails (dev no browser) — segue sem recarga automática
      }
    } catch (error) {
      toast.error(parseError(error), { id: toastId });
    } finally {
      setImporting(false);
    }
  }, [importSummary, selectedIndex]);

  const onTabChange = (index: string) => {
    const idx = Number.parseInt(index, 10);
    setSelectedIndex(idx);
    tabStorage.save(idx);
    setVersion(GAME_VERSIONS[idx].id);
  };

  return (
    <div className="flex h-screen flex-col">
      <header className="sticky top-0 z-10 flex items-center gap-2 bg-background px-4 py-2 border-b">
        <span className="text-xl font-malva-medium text-gradient">FINAL FANTASY X and X-2 HD Remaster Resources Editor</span>
        <span className="flex-1" />
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setConfigOpen(true)}
              aria-label="Configurações"
            >
              <Settings size={22} />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            {shortcutTip(SHORTCUTS.config.label, SHORTCUTS.config.keys)}
          </TooltipContent>
        </Tooltip>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="min-w-19 h-9 gap-1 px-2 text-sm"
            >
              {EXPORT_FORMAT_LABELS[exportFormat]}
              <ChevronDown size={14} />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel className="text-muted-foreground">
              Formato
            </DropdownMenuLabel>
            <DropdownMenuRadioGroup
              value={exportFormat}
              onValueChange={(value) => exportSelection.setFormat(value as ExportFormat)}
            >
              {(Object.keys(EXPORT_FORMAT_LABELS) as ExportFormat[]).map(
                (format) => (
                  <DropdownMenuRadioItem key={format} value={format}>
                    {EXPORT_FORMAT_LABELS[format]}
                  </DropdownMenuRadioItem>
                )
              )}
            </DropdownMenuRadioGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </header>

      <Tabs value={String(selectedIndex)} onValueChange={onTabChange} className="flex-1 min-h-0 flex flex-col">
        <div className="mx-4 mt-2 flex items-center justify-between gap-2">
          <TabsList variant="gradient" className="w-fit">
            {GAME_VERSIONS.map((tab, index) => (
              <Tooltip key={tab.id}>
                <TooltipTrigger asChild>
                  <TabsTrigger variant="gradient" value={String(index)}>
                    {tab.label}
                  </TabsTrigger>
                </TooltipTrigger>
                <TooltipContent>
                  {shortcutTip(
                    SHORTCUTS.versionTabs.label,
                    SHORTCUTS.versionTabs.keys[index]
                  )}
                </TooltipContent>
              </Tooltip>
            ))}
          </TabsList>
          <div className="flex items-center gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline">
                  <Download size={16} />
                  Exportar
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem
                  onSelect={() => void runExport(servedKinds, false)}
                >
                  Exportar tudo
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuLabel className="text-muted-foreground">
                  Escopo
                </DropdownMenuLabel>
                {servedKinds.map((kind) => (
                  <DropdownMenuCheckboxItem
                    key={kind}
                    checked={activeScope.includes(kind)}
                    onCheckedChange={() => toggleScope(kind)}
                    onSelect={(event) => event.preventDefault()}
                  >
                    {KIND_LABELS[kind]}
                  </DropdownMenuCheckboxItem>
                ))}
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onSelect={() => void runExport(activeScope, true)}
                >
                  Exportar seleção
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <Button variant="outline" onClick={() => void onImportClick()}>
              <Upload size={16} />
              Importar
            </Button>
          </div>
        </div>
        {GAME_VERSIONS.map((tab, index) => (
          <TabsContent key={tab.id} value={String(index)} className="flex-1 min-h-0 mt-0">
            <GameVersionTab version={tab.id} />
          </TabsContent>
        ))}
      </Tabs>

      <StatusBar />

      <ConfigDialog open={configOpen} onOpenChange={setConfigOpen} />
      <ProgressDialog
        open={progress.open}
        value={progress.value}
        label={progress.label}
        processed={progress.processed}
        total={progress.total}
        issueCount={progress.issueCount}
        complete={progress.complete}
        onClose={closeProgress}
      />
      <ImportSummaryDialog
        summary={importSummary}
        open={importSummary !== null}
        importing={importing}
        onOpenChange={(open) => {
          if (!open && !importing) setImportSummary(null);
        }}
        onConfirm={() => void onConfirmImport()}
      />
    </div>
  );
}
