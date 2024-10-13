import { inject, Injectable, signal } from '@angular/core';
import { MenuItem, TreeNode } from 'primeng/api';
import { ReadFileAsString } from '../../wailsjs/go/main/App';
import { extractedEditorText, selectedFile, showEditorModal } from '../app/components/signals/signals.signal';
import { extractFileInfo } from '../utils/utils';
import { ExtractService } from './extract.service';
import { CompressService } from './compress.service';
import { showProgress } from '../app/components/progress-modal/progress-modal.signal';
import { EventsEmit } from '../../wailsjs/runtime/runtime';

@Injectable({
  providedIn: 'root'
})
export class FfxContextMenuService {
  private readonly _extractService: ExtractService = inject(ExtractService);
  private readonly _compressService: CompressService = inject(CompressService);

  items = signal<MenuItem[]>([]);
  file = selectedFile;
  extractedText = extractedEditorText;

  constructor() {
    this.items.set([
      { label: 'View', icon: 'pi pi-file', command: (event: any) => this.view() },
      { label: 'Extract', icon: 'pi pi-download', command: async (event: any) => await this.extract() },
      { label: 'Import', icon: 'pi pi-upload', command: (event: any) => this.compress() },
    ]);
  }

  async view() {
    if (!this.file()) return;

    const fileInfo = extractFileInfo(this.file());
    if (!fileInfo) return;

    showEditorModal.set(true);

    const textContent = await ReadFileAsString(fileInfo);
    this.extractedText.set(textContent.replace(/(\r\n|\n|\r)/g, '<br>'));
  }


  async extract() {
    //TODO: Review try catch
    try {
      const data = extractFileInfo(this.file());
      if (!data) return;

      await this._extractService.extraction(data);
    } catch (error) {
      EventsEmit("ApplicationError", error);
    }
  }

  async compress() {
    const data = extractFileInfo(this.file());
    if (!data) return;

    await this._compressService.compress(data);
  }
}
