"use client";

import { useEffect, useMemo, useState } from "react";
import { Save } from "lucide-react";
import { useSelector } from "@tanstack/react-store";
import { SetUnsavedEdits } from "@/wailsjs/go/main/App";
import { KIND_LABELS } from "@/lib/ffx/display-names";
import type { GameVersionId } from "@/lib/ffx/game-version";
import { useEditDraft } from "@/lib/ffx/edit-draft";
import { useWailsEvent } from "@/lib/ffx/use-wails-event";
import { saveAllDrafts } from "@/lib/ffx/save-all";
import { setActiveFile } from "@/lib/ffx/active-file-store";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { LAYOUT_DEFAULTS_PX } from "@/lib/ffx/layout-storage";
import { usePersistedLayout } from "@/lib/ffx/use-persisted-layout";
import { SHORTCUTS, shortcutTip } from "@/lib/ffx/shortcuts";
import { invalidateTextSearchResults } from "@/lib/ffx/search-store";
import { warmTextSearch } from "@/lib/ffx/tree-data";
import {
  createEntryView,
  type EntryView,
} from "@/components/entry-view/entry-view-store";
import { ContentTree } from "@/components/entry-view/content-tree";
import { EntryTable } from "@/components/entry-view/entry-table";
import { ImageActionDialogs } from "@/components/dialogs/image-action-dialogs";
import { ImagePanel } from "@/components/entry-view/image-panel";
import { TranslationDialog } from "@/components/dialogs/translation-dialog";
import { PreloadVersions } from "@/wailsjs/go/main/App";
import { GAME_VERSIONS } from "@/lib/ffx/game-version";

/**
 * Aba de uma versão do jogo: só orquestra. O estado compartilhado entre
 * árvore, tabela e diálogo vive no store da view (createEntryView) — um por
 * aba — e cada filho se assina com useSelector.
 */
export function GameVersionTab({ version }: { version: GameVersionId }) {
  const view: EntryView = useMemo(() => createEntryView(version), [version]);
  const { store: drafts, snapshot } = useEditDraft();
  const activeKind = useSelector(view.store, (s) => s.activeKind);
  const selectedEntry = useSelector(view.store, (s) => s.selectedEntry);
  const loading = useSelector(view.store, (s) => s.loading);
  const progress = useSelector(view.store, (s) => s.progress);
  const [saving, setSaving] = useState(false);
  // Largura da sidebar persistida (percentual do grupo por id de painel).
  const { defaultLayout, onLayoutChanged } = usePersistedLayout("sidebar");

  const hasDirty = snapshot.hasDirty;

  useEffect(() => {
    void SetUnsavedEdits(hasDirty);
  }, [hasDirty]);

  // Publica o arquivo ativo (progresso por row) para a barra de status.
  // O progresso é fixo na abertura; troca de arquivo/aba atualiza aqui.
  useEffect(() => {
    setActiveFile({
      version,
      kind: activeKind,
      entryId: selectedEntry?.id ?? null,
      entryLabel: selectedEntry?.label ?? null,
      progress,
    });
  }, [version, activeKind, selectedEntry, progress]);

  useEffect(() => {
    // Sincroniza com o backend ao montar/trocar de versão. Ao concluir a
    // carga da aba VISUALIZADA, pré-carrega as demais em background —
    // trocar de aba depois cai no caminho rápido (cache do backend).
    void view.actions.reload().then(() => {
      // Índice de busca da aba aquecido em background: o primeiro Ctrl+K
      // não paga a construção síncrona do índice.
      warmTextSearch(version);
      void PreloadVersions(
        GAME_VERSIONS.filter((t) => t.id !== version).map((t) => t.id),
      );
    });
  }, [view, version]);

  useEffect(
    () =>
      drafts.onSaved(() => {
        // Salvou = os textos mudaram: resultados do modal ficaram velhos.
        invalidateTextSearchResults(version);
        void view.actions.reload();
      }),
    [drafts, view, version],
  );

  // Após importar, o backend emite ImportDone → recarrega árvore e tabela.
  useWailsEvent("ImportDone", () => {
    invalidateTextSearchResults(version);
    void view.actions.reload();
  });

  const onSaveAll = () => {
    void saveAllDrafts(() => saving, setSaving);
  };

  return (
    <div className="flex h-full min-h-0 border-t">
      <ResizablePanelGroup
        orientation="horizontal"
        className="min-w-0"
        defaultLayout={defaultLayout}
        onLayoutChanged={onLayoutChanged}
      >
        <ResizablePanel
          id="tree"
          defaultSize={LAYOUT_DEFAULTS_PX.sidebar}
          minSize={200}
          maxSize={600}
        >
          <ContentTree view={view} />
        </ResizablePanel>
        <ResizableHandle withHandle />

        <ResizablePanel id="main" minSize={300}>
          <main className="h-full min-w-0 p-3 px-4 overflow-auto">
            <div className="flex items-center justify-between gap-4">
              <h3 className="text-lg font-semibold flex items-center gap-3">
                {KIND_LABELS[activeKind]}
                {selectedEntry ? (
                  <span className="font-normal opacity-70">
                    {" "}
                    · {selectedEntry.label}
                  </span>
                ) : null}
                {progress ? (
                  <Badge
                    variant="outline"
                    className={
                      progress.translated >= progress.total &&
                      progress.total > 0
                        ? "text-emerald-700 border-emerald-300 dark:text-emerald-400 dark:border-emerald-800"
                        : "text-muted-foreground"
                    }
                  >
                    {progress.translated}/{progress.total} linhas traduzidas (
                    {progress.pct}%)
                  </Badge>
                ) : null}
              </h3>
              <Tooltip>
                <TooltipTrigger asChild>
                  {/* O botão desabilitado não dispara pointer events, então o
                      trigger é o span envoltório (padrão Radix p/ disabled). */}
                  <span className="inline-flex">
                    <Button
                      disabled={!hasDirty || saving}
                      onClick={() => void onSaveAll()}
                    >
                      <Save size={18} />
                      {saving ? "Salvando…" : "Salvar"}
                    </Button>
                  </span>
                </TooltipTrigger>
                <TooltipContent>
                  {shortcutTip(SHORTCUTS.save.label, SHORTCUTS.save.keys)}
                </TooltipContent>
              </Tooltip>
            </div>

            {selectedEntry ? (
              // images não tem rows: ocupa o lugar da tabela a pré-visualização
              // + ações (extrair/salvar/importar) da textura.
              activeKind === "images" ? (
                <ImagePanel view={view} />
              ) : (
                <EntryTable view={view} />
              )
            ) : loading ? (
              <p className="mt-8 opacity-70">Carregando…</p>
            ) : (
              <p className="mt-8 opacity-70">
                Selecione um arquivo no sidebar.
              </p>
            )}
          </main>
        </ResizablePanel>
      </ResizablePanelGroup>

      <TranslationDialog view={view} />
      {/* Ações de imagem (extrair/replicar/deletar): um diálogo só, aberto
          pelo menu da árvore ou pelos botões do painel. */}
      <ImageActionDialogs view={view} />
    </div>
  );
}
