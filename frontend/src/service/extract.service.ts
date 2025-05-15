import { inject, Injectable } from '@angular/core';
import { Extract } from '../../wailsjs/go/main/App';
import { ErrorHandlerService } from './error-handler.service';

@Injectable({
  providedIn: 'root',
})
export class ExtractService {
  private readonly _errorHandler: ErrorHandlerService =
    inject(ErrorHandlerService);

  async extraction(file: string) {
    try {
      await Extract(file);
    } catch (error) {
      this._errorHandler.sendErrorNotification(error);
    }
  }
}
