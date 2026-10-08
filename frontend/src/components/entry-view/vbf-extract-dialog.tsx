'use client';

import { useEffect, useState } from 'react';
import { loggedToast as toast } from '@/lib/ffx/toast-logged';
import { AlertTriangle, FolderOpen, HardDriveDownload, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { parseError } from '@/lib/ffx/error-handler';
import {
  extractVbfSelection,
  previewVbfExtraction,
  selectVbfExtractDir,
} from '@/lib/ffx/vbf';
import type { dto } from '@/wailsjs/go/models';

export interface VbfExtractRequest {
  requestId: number;
  root: string;
  paths: string[];
  /** data = destino padrão do jogo; choose = seletor nativo de pasta. */
  mode: 'data' | 'choose';
}

function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = value;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit++;
  }
  return `${size >= 10 || unit === 0 ? size.toFixed(0) : size.toFixed(1)} ${units[unit]}`;
}

/** Confirmação + extração da seleção VBF, preservando os caminhos internos. */
export function VbfExtractDialog({
  request,
  onClose,
  onDone,
}: {
  request: VbfExtractRequest | null;
  onClose: () => void;
  onDone: () => void;
}) {
  const [destRoot, setDestRoot] = useState<string | null>(
    request?.mode === 'data' ? '' : null
  );
  const [preview, setPreview] = useState<dto.VbfExtractPreview | null>(null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [extracting, setExtracting] = useState(false);

  // "Extrair para…" abre o diálogo nativo imediatamente; o diálogo de
  // confirmação só aparece depois de a pasta ser escolhida.
  useEffect(() => {
    if (!request || request.mode !== 'choose') return;
    let cancelled = false;
    void selectVbfExtractDir()
      .then((path) => {
        if (cancelled) return;
        if (!path) {
          onClose();
          return;
        }
        setDestRoot(path);
      })
      .catch((error) => {
        if (cancelled) return;
        toast.error(parseError(error));
        onClose();
      });
    return () => {
      cancelled = true;
    };
  }, [request, onClose]);

  useEffect(() => {
    if (!request || destRoot === null) return;
    let cancelled = false;
    void previewVbfExtraction(request.root, request.paths, destRoot)
      .then((result) => {
        if (!cancelled) setPreview(result);
      })
      .catch((error) => {
        if (!cancelled) setPreviewError(parseError(error));
      });
    return () => {
      cancelled = true;
    };
  }, [request, destRoot]);

  const runExtraction = async () => {
    if (!request || !preview || preview.files === 0 || extracting) return;
    setExtracting(true);
    const toastId = 'vbf-extract';
    toast.loading(`Extraindo ${preview.files} arquivo(s)…`, { id: toastId });
    try {
      const result = await extractVbfSelection(
        request.root,
        request.paths,
        destRoot ?? ''
      );
      const done = result.done ?? [];
      const failed = result.failed ?? [];
      if (failed.length > 0) {
        toast.warning(
          `Extraídos ${done.length} de ${result.total}. ${failed.length} arquivo(s) falharam.`,
          { id: toastId }
        );
      } else {
        toast.success(`Extraídos ${done.length} arquivo(s) para ${preview.destRoot}.`, {
          id: toastId,
        });
      }
      onDone();
    } catch (error) {
      toast.error(parseError(error), { id: toastId });
      setExtracting(false);
    }
  };

  const isOpen = request !== null && (request.mode === 'data' || destRoot !== null);
  const loadingPreview = Boolean(
    request && destRoot !== null && preview === null && previewError === null
  );
  const destination = preview?.destRoot ?? (destRoot || 'data/ do jogo');

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && !extracting && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <HardDriveDownload size={18} />
            Confirmar extração
          </DialogTitle>
          <DialogDescription>
            Os binários serão extraídos mantendo os caminhos relativos do .vbf.
          </DialogDescription>
        </DialogHeader>

        {loadingPreview ? (
          <div className="flex items-center gap-2 py-3 text-muted-foreground" role="status">
            <Loader2 size={16} className="animate-spin" />
            Calculando arquivos e possíveis sobrescritas…
          </div>
        ) : previewError ? (
          <p role="alert" className="text-destructive">{previewError}</p>
        ) : preview ? (
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
              <span className="text-muted-foreground">Arquivos</span>
              <span className="font-medium">{preview.files.toLocaleString('pt-BR')}</span>
              <span className="text-muted-foreground">Tamanho total</span>
              <span className="font-medium">{formatBytes(preview.bytes)}</span>
              <span className="text-muted-foreground">Destino</span>
              <span className="flex min-w-0 items-center gap-1 font-medium">
                <FolderOpen size={14} className="shrink-0" />
                <span className="truncate" title={destination}>{destination}</span>
              </span>
            </div>
            {preview.existing > 0 ? (
              <div
                role="alert"
                className="flex gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 p-3 text-sm"
              >
                <AlertTriangle size={17} className="mt-0.5 shrink-0 text-amber-600" />
                <span>
                  <strong>{preview.existing.toLocaleString('pt-BR')} arquivo(s)</strong> já
                  existem no destino e serão sobrescritos.
                </span>
              </div>
            ) : null}
            {preview.files === 0 ? (
              <p role="status" className="text-sm text-muted-foreground">
                A seleção não contém arquivos para extrair.
              </p>
            ) : null}
          </div>
        ) : null}

        <DialogFooter>
          <Button variant="outline" disabled={extracting} onClick={onClose}>
            Cancelar
          </Button>
          <Button
            disabled={loadingPreview || Boolean(previewError) || !preview || preview.files === 0 || extracting}
            onClick={() => void runExtraction()}
          >
            {extracting ? <Loader2 size={16} className="animate-spin" /> : <HardDriveDownload size={16} />}
            {extracting ? 'Extraindo…' : 'Confirmar extração'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
