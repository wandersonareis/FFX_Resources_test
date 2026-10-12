'use client';

import { useCallback, useState } from 'react';
import { Download, Upload } from 'lucide-react';
import { loggedToast as toast } from '@/lib/ffx/toast-logged';
import { EventsEmit } from '@/wailsjs/runtime/runtime';
import { dto } from '@/wailsjs/go/models';
import { GAME_VERSIONS, GameVersionId } from '@/lib/ffx/game-version';
import { EntryKind, KIND_LABELS } from '@/lib/ffx/display-names';
import { exportSelection, EXPORT_FORMAT_LABELS } from '@/lib/ffx/export-selection';
import { exportProgress } from '@/lib/ffx/export-progress';
import { parseError } from '@/lib/ffx/error-handler';
import {
  entryKindsFor,
  exportDestination,
  exportJSON,
  exportStrings,
  importFile,
  previewImport,
  selectImportFile,
} from '@/lib/ffx/tree-data';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ImportSummaryDialog } from '@/components/dialogs/import-summary-dialog';
import { ExportToastContent } from '@/components/progress-notifier';

/**
 * Controles Exportar/Importar da barra de abas + diálogo de resumo de import.
 * Única prop: a versão ativa. Tudo o mais é estado interno (escopo por versão,
 * resumo/importando) ou store módulo (`exportSelection`) — zero prop drilling.
 */
export function ExportImportControls({ version }: { version: GameVersionId }) {
  // Escopo de exportação por versão (kinds marcados no dropdown Exportar).
  const [scopes, setScopes] = useState<Record<string, EntryKind[]>>(() =>
    Object.fromEntries(GAME_VERSIONS.map((tab) => [tab.id, entryKindsFor(tab.id)]))
  );
  const [importSummary, setImportSummary] = useState<dto.ImportSummary | null>(
    null
  );
  const [importing, setImporting] = useState(false);

  const servedKinds = entryKindsFor(version);
  const activeScope = scopes[version] ?? servedKinds;

  const toggleScope = (kind: EntryKind) => {
    setScopes((prev) => {
      const current = prev[version] ?? entryKindsFor(version);
      const next = current.includes(kind)
        ? current.filter((k) => k !== kind)
        : [...current, kind];
      return { ...prev, [version]: next };
    });
  };

  // Exporta os kinds no FORMATO SELECIONADO (JSON | Strings), em sequência
  // (evita GetCollection concorrente). useCheckedIds=false força ids=[] →
  // tudo do kind; com seleção, exigirá ≥1 marcado.
  const runExport = useCallback(
    async (kinds: EntryKind[], useCheckedIds: boolean) => {
      const format = exportSelection.formatOf();
      if (useCheckedIds && kinds.every((kind) => exportSelection.countOf(version, kind) === 0)) {
        toast.info('Nada para exportar — selecione itens na árvore.');
        return;
      }
      const label = EXPORT_FORMAT_LABELS[format];
      const toastId = 'export';
      exportProgress.begin();
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
        exportProgress.end();
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
        exportProgress.end();
      }
    },
    [version]
  );

  const onImportClick = useCallback(async () => {
    try {
      const path = await selectImportFile();
      if (!path) return;
      setImportSummary(await previewImport(path, version));
    } catch (error) {
      toast.error(parseError(error));
    }
  }, [version]);

  const onConfirmImport = useCallback(async () => {
    const current = importSummary;
    if (!current) return;
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
  }, [importSummary, version]);

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline">
            <Download size={16} />
            Exportar
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onSelect={() => void runExport(servedKinds, false)}>
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
          <DropdownMenuItem onSelect={() => void runExport(activeScope, true)}>
            Exportar seleção
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <Button variant="outline" onClick={() => void onImportClick()}>
        <Upload size={16} />
        Importar
      </Button>
      <ImportSummaryDialog
        summary={importSummary}
        open={importSummary !== null}
        importing={importing}
        onOpenChange={(open) => {
          if (!open && !importing) setImportSummary(null);
        }}
        onConfirm={() => void onConfirmImport()}
      />
    </>
  );
}
