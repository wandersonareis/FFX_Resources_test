import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { dto } from '../../../../wailsjs/go/models';
import { GameTextEditorComponent } from '../game-text/game-text-editor.component';

/** Idioma editado/salvo no "traduzido" (substitui o inglês no binário). */
export const SOURCE_LANG = 'us';

export interface TranslationCellDialogData {
  row: dto.TextRow;
  /** Idiomas de consulta, sem o idioma de origem (us). */
  languages: Array<{ code: string; name: string }>;
}

@Component({
  selector: 'app-translation-cell-dialog',
  standalone: true,
  imports: [CommonModule, MatButtonModule, MatDialogModule, GameTextEditorComponent],
  template: `
    <h2 mat-dialog-title>
      Linha #{{ data.row.index }}@if (data.row.name) {
        <span class="row-name"> · {{ data.row.name }}</span>
      }
    </h2>
    <mat-dialog-content>
      <div class="references">
        @for (lang of referenceLangs; track lang.code) {
          <div class="reference">
            <span class="ref-lang">{{ lang.name }} ({{ lang.code }})</span>
            <p class="ref-text">{{ refText(lang.code) }}</p>
          </div>
        } @empty {
          <p class="ref-empty">Sem outros idiomas nesta linha.</p>
        }
      </div>
      <div class="editor-block">
        <span class="editor-label">Tradução ({{ sourceLang }}) — gravada no jogo ao salvar</span>
        <app-game-text-editor
          [value]="text()"
          (valueChange)="onText($event)"
        />
      </div>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="cancel()">Cancelar</button>
      <button mat-flat-button color="primary" [disabled]="!changed()" (click)="save()">
        Salvar
      </button>
    </mat-dialog-actions>
  `,
  styles: [
    `
      mat-dialog-content {
        display: flex;
        flex-direction: column;
        overflow: hidden;
      }
      .references {
        display: grid;
        gap: 10px;
        margin-bottom: 16px;
        overflow-y: auto;
        flex: 1 1 auto;
        min-height: 120px;
        max-height: 42vh;
        padding-right: 4px;
      }
      .ref-lang {
        font-size: 12px;
        font-weight: 600;
        opacity: 0.75;
      }
      .ref-text {
        margin: 2px 0 0;
        white-space: pre-wrap;
      }
      .ref-empty {
        opacity: 0.7;
      }
      .editor-block {
        display: grid;
        gap: 6px;
        flex: 0 0 auto;
        overflow: visible;
      }
      .editor-label {
        font-size: 12px;
        font-weight: 600;
        opacity: 0.75;
      }
      .row-name {
        opacity: 0.7;
      }
    `,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class TranslationCellDialogComponent {
  readonly data = inject<TranslationCellDialogData>(MAT_DIALOG_DATA);
  private readonly ref = inject(MatDialogRef<TranslationCellDialogComponent>);

  protected readonly sourceLang = SOURCE_LANG;

  protected readonly text = signal(this.data.row.text?.[SOURCE_LANG] ?? '');

  protected readonly changed = signal(
    this.text() !== (this.data.row.text?.[SOURCE_LANG] ?? '')
  );

  protected readonly referenceLangs: ReadonlyArray<{ code: string; name: string }> =
    this.data.languages.filter((l) => l.code !== SOURCE_LANG);

  protected refText(code: string): string {
    return this.data.row.text?.[code] ?? '—';
  }

  protected onText(value: string): void {
    this.text.set(value);
    this.changed.set(value !== (this.data.row.text?.[SOURCE_LANG] ?? ''));
  }

  protected cancel(): void {
    this.ref.close();
  }

  protected save(): void {
    if (!this.changed()) return;
    this.ref.close(this.text());
  }
}
