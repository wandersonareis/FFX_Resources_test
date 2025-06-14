import { inject, Injectable, signal, WritableSignal } from '@angular/core';
import { MenuItem, TreeNode } from 'primeng/api';
import { ReadFileAsString } from '../../wailsjs/go/main/App';
import {
  extractedEditorText,
  selectedFile,
  showEditorModal,
} from '../app/components/signals/signals.signal';
import { getFileInfoFromNode } from '../utils/utils';
import { ExtractService } from './extract.service';
import { CompressService } from './compress.service';
import { fileFormats } from '../../wailsjs/go/models';
import { ErrorHandlerService } from './error-handler.service';

@Injectable({
  providedIn: 'root',
})
export class FfxContextMenuService {
  private readonly _extractService: ExtractService = inject(ExtractService);
  private readonly _compressService: CompressService = inject(CompressService);
  private readonly _errorHandler: ErrorHandlerService =
    inject(ErrorHandlerService);

  readonly items: WritableSignal<MenuItem[]> = signal<MenuItem[]>([]);
  readonly file: WritableSignal<TreeNode | undefined> = selectedFile;
  readonly extractedText: WritableSignal<string> = extractedEditorText;

  constructor() {
    this.items.set([
      {
        label: 'View',
        icon: 'pi pi-file',
        command: (event: any) => this.view(),
      },
      {
        label: 'Extract',
        icon: 'pi pi-download',
        command: async (event: any) => await this.extract(),
      },
      {
        label: 'Import',
        icon: 'pi pi-upload',
        command: (event: any) => this.compress(),
      },
    ]);
  }

  async view(): Promise<void> {
    if (!this.file()) return;

    const fileInfo: fileFormats.TreeNodeData | null = getFileInfoFromNode(
      this.file()
    );

    if (!fileInfo || !fileInfo.extract_location) return;

    if (!fileInfo.extract_location.IsExist) {
      this._errorHandler.sendErrorNotificationWithMessage(
        'File not found. Please extract the file first.'
      );

      return;
    }

    const filePath = fileInfo.extract_location.TargetFile;

    showEditorModal.set(true);

    const textContent: string = await ReadFileAsString(filePath);
    this.extractedText.set(textContent.replace(/(\r\n|\n|\r)/g, '<br>'));
  }

  async extract(): Promise<void> {
    try {
      const fileInfo: fileFormats.TreeNodeData | null = getFileInfoFromNode(
        this.file()
      );

      if (!fileInfo || !fileInfo.source) return;

      const filePath = fileInfo.source.path;

      await this._extractService.extraction(filePath);
    } catch (error) {
      this._errorHandler.sendErrorNotification(error);
    }
  }

  async compress(): Promise<void> {
    try {
      const fileInfo: fileFormats.TreeNodeData | null = getFileInfoFromNode(
        this.file()
      );

      if (!fileInfo || !fileInfo.source) return;

      const filePath = fileInfo.source.path;

      await this._compressService.compress(filePath);
    } catch (error) {
      this._errorHandler.sendErrorNotification(error);
    }
  }
}
