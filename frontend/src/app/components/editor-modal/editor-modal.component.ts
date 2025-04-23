import {Component, inject, signal, WritableSignal} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { EditorModule } from 'primeng/editor';
import { DialogModule } from 'primeng/dialog';
import { getFileInfoFromNode } from '../../../utils/utils';
import { WriteTextFile } from '../../../../wailsjs/go/main/App';
import { extractedEditorText, selectedFile, showEditorModal } from '../signals/signals.signal';
import { ButtonModule } from 'primeng/button';
import { CompressService } from '../../../service/compress.service';
import {core, fileFormats, spira} from "../../../../wailsjs/go/models";
import {TreeNode} from "primeng/api";

@Component({
    selector: 'app-editor-modal',
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
    const fileInfo: fileFormats.TreeNodeData | null = getFileInfoFromNode(this.file())
    if (!fileInfo || !fileInfo.source) return;

    const filePath = fileInfo.source.path
    await WriteTextFile(filePath, event.textValue);
  }

  async saveToSpiraFile() {
    const fileInfo: fileFormats.TreeNodeData | null = getFileInfoFromNode(this.file())
    if (!fileInfo || !fileInfo.source) return;

    const filePath = fileInfo.source.path
    await this._compressService.compress(filePath);
  }
}
