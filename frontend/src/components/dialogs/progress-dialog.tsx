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
}

export function ProgressDialog({ open, value }: ProgressDialogProps) {
  return (
    <Dialog open={open}>
      <DialogContent className="sm:max-w-[320px]" onInteractOutside={(e) => e.preventDefault()}>
        <DialogHeader>
          <DialogTitle>Processando…</DialogTitle>
          <DialogDescription className="sr-only">Aguarde</DialogDescription>
        </DialogHeader>
        <Progress value={value} />
      </DialogContent>
    </Dialog>
  );
}
