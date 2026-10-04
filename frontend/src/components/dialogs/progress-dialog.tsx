'use client';

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Progress } from '@/components/ui/progress';

export interface ProgressDialogProps {
  open: boolean;
  value: number;
  /** O que está processando (o rótulo vem do evento Progress do backend). */
  label?: string;
}

export function ProgressDialog({ open, value, label }: ProgressDialogProps) {
  return (
    <Dialog open={open}>
      <DialogContent
        className="sm:max-w-[320px]"
        onInteractOutside={(e) => e.preventDefault()}
        // Sem botão de fechar: o ciclo de progresso é controlado pelo
        // backend (ShowProgress) — fechar aqui não cancelaria nada, só
        // esconderia a barra com o processo ainda rodando.
        showCloseButton={false}
      >
        <DialogHeader>
          <DialogTitle>{label?.trim() ? label.trim() : 'Processando…'}</DialogTitle>
          <DialogDescription className="sr-only">Aguarde</DialogDescription>
        </DialogHeader>
        <Progress value={value} />
      </DialogContent>
    </Dialog>
  );
}
