'use client';

import { Copy, FlipVertical, FolderInput, Images, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { SHORTCUTS, shortcutTip } from '@/lib/ffx/shortcuts';
import { ImageExportMenu } from './export-menu';
import type { ImageBusy } from './use-image-actions';

/**
 * Barra de ações do painel: exportar (salvar como .dds/.png em pasta livre,
 * ou extrair os dois formatos para mods/images — a extração mora no menu
 * Exportar), importar um .dds editado (repack só do mesmo formato/dimensão
 * nesta etapa), propagar o import para as cópias idênticas, espelhar a
 * pré-visualização e reanalisar o índice.
 */
export function ImageToolbar({
  busy,
  readOnly,
  flipped,
  duplicateCount,
  onExtract,
  onSave,
  onImport,
  onImportTranslated,
  onReplicate,
  onToggleFlip,
  onRefresh,
}: {
  busy: ImageBusy;
  readOnly: boolean;
  flipped: boolean;
  duplicateCount: number;
  onExtract: () => void;
  onSave: (format: 'dds' | 'png') => void;
  onImport: () => void;
  onImportTranslated: () => void;
  onReplicate: () => void;
  onToggleFlip: () => void;
  onRefresh: () => void;
}) {
  const copyTotal = duplicateCount + 1;

  return (
    <div className="flex flex-wrap items-center gap-2">
      <ImageExportMenu
        disabled={busy !== null}
        readOnly={readOnly}
        onSave={onSave}
        onExtract={onExtract}
      />
      <Button size="sm" disabled={busy !== null} onClick={onImport}>
        <FolderInput size={16} />
        {busy === 'import'
          ? 'Importando…'
          : !readOnly && duplicateCount > 0
            ? `Importar .dds… (${copyTotal} idênticas)`
            : 'Importar .dds…'}
      </Button>
      <Button
        size="sm"
        variant="outline"
        disabled={busy !== null}
        title="Importa em lote os .dds.phyre de mods/translated/images — só os diferentes dos já em mods/"
        onClick={onImportTranslated}
      >
        <Images size={16} />
        {busy === 'importBatch' ? 'Importando…' : 'Importar traduzidas…'}
      </Button>
      {duplicateCount > 0 ? (
        <Button
          size="sm"
          variant="outline"
          disabled={busy !== null}
          title={
            readOnly
              ? 'Reempacota a imagem aberta em mods/ das cópias do .vbf já identificadas'
              : 'Reempacota a imagem aberta em mods/ de cada cópia idêntica — sem escolher arquivo'
          }
          onClick={onReplicate}
        >
          <Copy size={16} />
          Replicar para {duplicateCount} cópia
          {duplicateCount > 1 ? 's' : ''}
        </Button>
      ) : null}
      <Button
        size="sm"
        variant={flipped ? 'default' : 'outline'}
        aria-pressed={flipped}
        title={`${shortcutTip(SHORTCUTS.imageFlip.label, SHORTCUTS.imageFlip.keys)} — só na tela: não altera nem reenvia a imagem`}
        onClick={onToggleFlip}
      >
        <FlipVertical size={16} />
        {flipped ? 'Flip (on)' : 'Flip'}
      </Button>
      {!readOnly ? (
        <Button
          size="sm"
          variant="ghost"
          disabled={busy !== null}
          title="Reconstrói o índice de duplicatas — para textura editada fora do app (hex editor)"
          onClick={onRefresh}
        >
          <RefreshCw size={16} />
          {busy === 'refresh' ? 'Reanalisando…' : 'Reanalisar'}
        </Button>
      ) : null}
    </div>
  );
}
