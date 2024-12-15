import {inject, Injectable, signal, WritableSignal} from '@angular/core';
import {MenuItem, TreeNode} from 'primeng/api';
import { ReadFileAsString } from '../../wailsjs/go/main/App';
import { extractedEditorText, selectedFile, showEditorModal } from '../app/components/signals/signals.signal';
import { extractFileInfo } from '../utils/utils';
import { ExtractService } from './extract.service';
import { CompressService } from './compress.service';
import { EventsEmit } from '../../wailsjs/runtime';
import {spira} from "../../wailsjs/go/models";
import GameDataInfo = spira.GameDataInfo;

@Injectable({
  providedIn: 'root'
})
export class FfxContextMenuService {
  private readonly _extractService: ExtractService = inject(ExtractService);
  private readonly _compressService: CompressService = inject(CompressService);

  items: WritableSignal<MenuItem[]> = signal<MenuItem[]>([]);
  file: WritableSignal<TreeNode | undefined> = selectedFile;
  extractedText: WritableSignal<string> = extractedEditorText;

  constructor() {
    this.items.set([
      { label: 'View', icon: 'pi pi-file', command: (event: any) => this.view() },
      { label: 'Extract', icon: 'pi pi-download', command: async (event: any) => await this.extract() },
      { label: 'Import', icon: 'pi pi-upload', command: (event: any) => this.compress() },
    ]);
  }

  async view(): Promise<void> {
    if (!this.file()) return;

    const fileInfo: spira.GameDataInfo | null = extractFileInfo(this.file());
    if (!fileInfo) return;

    showEditorModal.set(true);

    const textContent: string = await ReadFileAsString(fileInfo.file_path);
    this.extractedText.set(textContent.replace(/(\r\n|\n|\r)/g, '<br>'));
  }


  async extract(): Promise<void> {
    //TODO: Review try catch
    try {
      const fileInfo: GameDataInfo | null = extractFileInfo(this.file());
      if (!fileInfo) return;

      await this._extractService.extraction(fileInfo.file_path);
    } catch (error) {
      EventsEmit("Notify", error);
    }
  }

  async compress(): Promise<void> {
    const fileInfo: spira.GameDataInfo | null = extractFileInfo(this.file());
    if (!fileInfo) return;

    await this._compressService.compress(fileInfo.file_path);
  }
}
