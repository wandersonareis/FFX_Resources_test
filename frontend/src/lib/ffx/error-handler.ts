import { EventsEmit } from '@/wailsjs/runtime/runtime';

export function sendErrorNotificationWithMessage(message: string): void {
  EventsEmit('Notify', { severity: 'error', message });
}

export function sendErrorNotification(err: unknown): void {
  EventsEmit('Notify', { severity: 'error', message: parseError(err) });
}

export function parseError(err: unknown): string {
  if (err instanceof Error) {
    return err.message;
  }
  return 'Erro desconhecido';
}
