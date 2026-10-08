'use client';

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Progress } from '@/components/ui/progress';
import { AlertTriangle, CheckCircle2 } from 'lucide-react';
import { Button } from '@/components/ui/button';

export interface ProgressDialogProps {
  open: boolean;
  value: number;
  /** O que está processando (o rótulo vem do evento Progress do backend). */
  label?: string;
  processed: number;
  total: number;
  issueCount: number;
  complete: boolean;
  onClose: () => void;
}

export function ProgressDialog({
  open,
  value,
  label,
  processed,
  total,
  issueCount,
  complete,
  onClose,
}: ProgressDialogProps) {
  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen && complete) onClose();
      }}
    >
      <DialogContent
        className="sm:max-w-sm"
        onInteractOutside={(e) => e.preventDefault()}
        showCloseButton={complete}
      >
        <DialogHeader>
          <DialogTitle>{label?.trim() ? label.trim() : 'Processando…'}</DialogTitle>
          <DialogDescription>
            {complete
              ? issueCount > 0
                ? `Concluído com ${issueCount} falha(s).`
                : 'Processamento concluído.'
              : 'Aguarde enquanto os arquivos são processados.'}
          </DialogDescription>
        </DialogHeader>
        {complete ? (
          <div className="space-y-2 py-2" role="status">
            <div className="flex justify-center">
              {issueCount > 0 ? (
                <AlertTriangle className="size-14 text-amber-600" aria-label="Concluído com falhas" />
              ) : (
                <CheckCircle2 className="size-14 text-emerald-600" aria-label="Concluído" />
              )}
            </div>
            <p className="text-center text-sm tabular-nums text-muted-foreground">
              {processed.toLocaleString('pt-BR')} / {total.toLocaleString('pt-BR')} processados
            </p>
          </div>
        ) : (
          <>
            <Progress value={value} />
            <p className="text-center text-sm tabular-nums text-muted-foreground">
              {processed.toLocaleString('pt-BR')} / {total.toLocaleString('pt-BR')}
            </p>
          </>
        )}
        {complete ? (
          <div className="flex justify-end">
            <Button onClick={onClose}>Fechar</Button>
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}
