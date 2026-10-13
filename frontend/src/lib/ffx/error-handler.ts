import { Events } from '@wailsio/runtime';

export function sendErrorNotificationWithMessage(message: string): void {
  Events.Emit('Notify', { severity: 'error', message });
}

export function sendErrorNotification(err: unknown): void {
  Events.Emit('Notify', { severity: 'error', message: parseError(err) });
}

export function parseError(err: unknown): string {
  if (err instanceof Error) {
    return err.message;
  }
  return 'Erro desconhecido';
}
