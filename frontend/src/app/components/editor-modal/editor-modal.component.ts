import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { EditorModule } from 'primeng/editor';
import { DialogModule } from 'primeng/dialog';
import { extractFileInfo } from '../../../utils/utils';
import { WriteTextFile } from '../../../../wailsjs/go/main/App';
import { extractedEditorText, selectedFile, showEditorModal } from '../signals/signals.signal';
import { ButtonModule } from 'primeng/button';
import { CompressService } from '../../../service/compress.service';

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
  visible = showEditorModal

  file = selectedFile
  text = extractedEditorText;

  showEditor() {
    this.visible.set(!this.visible());
  }

  async onTextChange(event: any) {
    const data = extractFileInfo(this.file())
    if (!data) return;

    await WriteTextFile(data, event.textValue);
  }

  async saveToSpiraFile() {
    const data = extractFileInfo(this.file())
    if (!data) return;

    await this._compressService.compress(data);
  }
}
