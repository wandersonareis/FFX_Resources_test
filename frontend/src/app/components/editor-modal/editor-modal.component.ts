import {Component, inject, signal, WritableSignal} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { EditorModule } from 'primeng/editor';
import { DialogModule } from 'primeng/dialog';
import { extractFileInfo } from '../../../utils/utils';
import { WriteTextFile } from '../../../../wailsjs/go/main/App';
import { extractedEditorText, selectedFile, showEditorModal } from '../signals/signals.signal';
import { ButtonModule } from 'primeng/button';
import { CompressService } from '../../../service/compress.service';
import {spira} from "../../../../wailsjs/go/models";
import {TreeNode} from "primeng/api";

@Component({
  selector: 'app-editor-modal',
  standalone: true,
  imports: [
    CommonModule,
    ButtonModule,
    FormsModule,
    DialogModule,
    EditorModule
  ],
  templateUrl: './editor-modal.component.html',
  styleUrl: './editor-modal.component.css'
})
export class EditorModalComponent {
  private readonly _compressService: CompressService = inject(CompressService)
  visible: WritableSignal<boolean> = showEditorModal

  file: WritableSignal<TreeNode | undefined> = selectedFile
  text: WritableSignal<string> = extractedEditorText;

  showEditor() {
    this.visible.set(!this.visible());
  }

  async onTextChange(event: any) {
    const fileInfo: spira.GameDataInfo | null = extractFileInfo(this.file())
    if (!fileInfo) return;

    await WriteTextFile(fileInfo.file_path, event.textValue);
  }

  async saveToSpiraFile() {
    const fileInfo: spira.GameDataInfo | null = extractFileInfo(this.file())
    if (!fileInfo) return;

    await this._compressService.compress(fileInfo);
  }
}
