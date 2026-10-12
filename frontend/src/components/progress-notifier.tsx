'use client';

import { useCallback, useRef, useState } from 'react';
import { loggedToast as toast } from '@/lib/ffx/toast-logged';
import { useWailsEvent } from '@/lib/ffx/use-wails-event';
import { exportProgress } from '@/lib/ffx/export-progress';
import { Progress } from '@/components/ui/progress';
import { ProgressDialog } from '@/components/dialogs/progress-dialog';

/** Cap da fila de toasts de ciclo aberto (descarta o mais antigo). */
const NOTIFY_QUEUE_CAP = 50;

type NotifyPayload = { severity?: string; message?: string };

type ProgressState = {
  open: boolean;
  value: number;
  label: string;
  processed: number;
  total: number;
  issueCount: number;
  complete: boolean;
};

const IDLE_PROGRESS: ProgressState = {
  open: false,
  value: 0,
  label: '',
  processed: 0,
  total: 0,
  issueCount: 0,
  complete: false,
};

/** Rende uma notificação por severidade (error: sticky; demais: 4 s). */
function emitNotify(payload: NotifyPayload): void {
  if (payload?.severity === 'error') {
    toast.error(payload?.message ?? 'Notificação');
  } else {
    toast(payload?.message ?? 'Notificação', { duration: 4000 });
  }
}

/** Conteúdo do toast de export: rótulo + barra embutida + contagem. */
export function ExportToastContent({
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

/**
 * Notificações Wails (Notify) e barra de progresso (ShowProgress/Progress) num
 * componente sem props: o estado do progresso vive aqui, então os ticks de
 * Progress re-renderizam apenas este overlay — não o shell inteiro (que monta a
 * aba GameVersionTab pesada). O guard do toast de export vem da store módulo
 * `exportProgress` (escrita por runExport), sem acoplar via props.
 */
export function ProgressNotifier() {
  const [progress, setProgress] = useState<ProgressState>(IDLE_PROGRESS);
  // Modal de progresso aberto bloqueia interação fora do diálogo (o radix
  // modal põe pointer-events: none no body): notificações que chegarem
  // durante um ciclo vão para a fila e aparecem quando o usuário fecha o
  // modal de resultado.
  const progressOpen = useRef(false);
  const notifyQueue = useRef<NotifyPayload[]>([]);

  const closeProgress = useCallback(() => {
    progressOpen.current = false;
    setProgress(IDLE_PROGRESS);
    if (notifyQueue.current.length > 0) {
      const pending = notifyQueue.current;
      notifyQueue.current = [];
      for (const q of pending) emitNotify(q);
    }
  }, []);

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
      setProgress({ ...IDLE_PROGRESS, open: true });
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
    if (exportProgress.isActive()) {
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

  return (
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
  );
}
