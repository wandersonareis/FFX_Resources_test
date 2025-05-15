import { Injectable } from '@angular/core';
import { EventsEmit } from '../../wailsjs/runtime/runtime';

@Injectable({
  providedIn: 'root',
})
export class ErrorHandlerService {
  constructor() {}

  sendErrorNotificationWithMessage(message: string): void {
    const notification = {
      severity: 'error',
      message: message,
    };

    EventsEmit('Notify', notification);
    return;
  }
  
  sendErrorNotification(err: unknown): void {
    const errorMessage = this.parseError(err);
    const notification = {
      severity: 'error',
      message: errorMessage,
    };

    EventsEmit('Notify', notification);
    return;
  }

  parseError(err: unknown): string {
    if (err instanceof Error) {
      return err.message;
    }
    return 'Erro desconhecido';
  }
}
