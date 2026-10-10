'use client';

import { Copy, Download, FlipVertical, FolderInput, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { SHORTCUTS, shortcutTip } from '@/lib/ffx/shortcuts';
import { ImageExportMenu } from './export-menu';
import type { ImageBusy } from './use-image-actions';

/**
 * Barra de ações do painel: extrair cópias de trabalho (.dds + .png em
 * mods/images), salvar em qualquer lugar do disco, importar um .dds editado
 * (repack só do mesmo formato/dimensão nesta etapa), propagar o import para
 * as cópias idênticas, espelhar a pré-visualização e reanalisar o índice.
 */
export function ImageToolbar({
  busy,
  readOnly,
  flipped,
  duplicateCount,
  onExtract,
  onSave,
  onImport,
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
  onReplicate: () => void;
  onToggleFlip: () => void;
  onRefresh: () => void;
}) {
  const copyTotal = duplicateCount + 1;

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Button
        size="sm"
        variant="outline"
        disabled={busy !== null}
        title={
          readOnly
            ? 'Extrai esta imagem do .vbf como .dds e .png, preservando o caminho interno'
            : 'Gera .dds e .png na pasta de imagens de trabalho — o diálogo pergunta se é só esta ou todas as cópias'
        }
        onClick={onExtract}
      >
        <Download size={16} />
        Extrair .dds + .png
      </Button>
      <ImageExportMenu disabled={busy !== null} onSave={onSave} />
      <Button size="sm" disabled={busy !== null} onClick={onImport}>
        <FolderInput size={16} />
        {busy === 'import'
          ? 'Importando…'
          : !readOnly && duplicateCount > 0
            ? `Importar .dds… (${copyTotal} idênticas)`
            : 'Importar .dds…'}
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
