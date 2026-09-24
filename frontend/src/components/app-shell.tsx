'use client';

import { useCallback, useEffect, useState } from 'react';
import { useHotkey } from '@tanstack/react-hotkeys';
import { toast } from 'sonner';
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
import { ConfigDialog } from '@/components/dialogs/config-dialog';
import { ImportSummaryDialog } from '@/components/dialogs/import-summary-dialog';
import { ProgressDialog } from '@/components/dialogs/progress-dialog';

export function AppShell() {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [saving, setSaving] = useState(false);
  const [configOpen, setConfigOpen] = useState(false);
  const [progress, setProgress] = useState({ open: false, value: 0 });
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

  useEffect(() => {
    // Hydration-safe: localStorage só existe no cliente.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setSelectedIndex(tabStorage.load());
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

  // Ctrl+1/2/3: trocar aba de versão.
  useHotkey('Mod+1', () => switchTab(0));
  useHotkey('Mod+2', () => switchTab(1));
  useHotkey('Mod+3', () => switchTab(2));
  // Ctrl+, : abrir Configurações (convenção de settings).
  useHotkey('Mod+,', () => setConfigOpen(true));
  // Ctrl+Alt+F: alterna formato JSON ↔ Strings direto (popover não abre).
  useHotkey('Mod+Alt+F', () => {
    const next =
      exportSelection.formatOf() === 'json' ? 'strings' : 'json';
    exportSelection.setFormat(next);
    toast.info(`Formato de exportação: ${EXPORT_FORMAT_LABELS[next]}`);
  });
  // Ctrl+S: salvar rascunhos (mesma ação do botão Salvar; sem fechar o app).
  useHotkey('Mod+S', () => {
    void saveAllDrafts(() => saving, setSaving);
  });

  useWailsEvent('Notify', (data) => {
    const payload = data as { severity?: string; message?: string };
    const sticky = payload?.severity === 'error';
    if (sticky) {
      toast.error(payload?.message ?? 'Notificação', { duration: Infinity });
    } else {
      toast(payload?.message ?? 'Notificação', { duration: 4000 });
    }
  });

  useWailsEvent('ShowProgress', (visible) => {
    setProgress((p) => ({ ...p, open: Boolean(visible) }));
  });

  useWailsEvent('Progress', (data) => {
    const payload = data as { percentage?: number };
    setProgress((p) => ({ ...p, value: payload?.percentage ?? 0 }));
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
      toast.loading(`Exportando (${label})…`, { id: toastId });
      try {
        let written = 0;
        for (const kind of kinds) {
          const ids = useCheckedIds ? exportSelection.idsOf(version, kind) : [];
          // Exportar seleção = atalho do Exportar do contexto.
          if (useCheckedIds && ids.length === 0) continue;
          const paths =
            format === 'json'
              ? await exportJSON(kind, version, ids)
              : await exportStrings(kind, version, ids);
          written += (paths ?? []).length;
        }
        if (written === 0) {
          toast.info('Nada para exportar.', { id: toastId });
        } else {
          toast.success(`Exportado (${label}): ${written} arquivo(s).`, {
            id: toastId,
          });
        }
      } catch (error) {
        toast.error(parseError(error), { id: toastId });
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
        `Importação concluída — ${changed} texto(s) atualizado(s).`,
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
        <span className="text-lg font-semibold">FFX Resources</span>
        <span className="flex-1" />
        <Button variant="ghost" size="icon" onClick={() => setConfigOpen(true)} aria-label="Configurações">
          <Settings size={22} />
        </Button>
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
          <TabsList className="w-fit">
            {GAME_VERSIONS.map((tab, index) => (
              <TabsTrigger key={tab.id} value={String(index)}>
                {tab.label}
              </TabsTrigger>
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

      <ConfigDialog open={configOpen} onOpenChange={setConfigOpen} />
      <ProgressDialog open={progress.open} value={progress.value} />
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
