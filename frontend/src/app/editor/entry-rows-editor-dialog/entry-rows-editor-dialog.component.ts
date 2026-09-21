import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatListModule } from '@angular/material/list';
import { MatSelectModule } from '@angular/material/select';
import { ListLanguages } from '../../../../wailsjs/go/main/App';
import { common, dto } from '../../../../wailsjs/go/models';
import { GameTextEditorComponent } from '../game-text/game-text-editor.component';

export interface EntryRowsEditorDialogData {
  title: string;
  entry: dto.FileEntry;
}

const DEFAULT_LANG = 'us';

@Component({
  selector: 'app-entry-rows-editor-dialog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatListModule,
    MatSelectModule,
    GameTextEditorComponent,
  ],
  template: `
    <h2 mat-dialog-title>{{ data.title }}</h2>
    <mat-dialog-content class="rows-editor">
      <div class="rows-pane">
        <mat-form-field appearance="outline" subscriptSizing="dynamic" class="lang-field">
          <mat-label>Idioma</mat-label>
          <mat-select [ngModel]="lang()" (ngModelChange)="lang.set($event)">
            @for (l of languages(); track l.code) {
              <mat-option [value]="l.code">{{ l.name }} ({{ l.code }})</mat-option>
            }
          </mat-select>
        </mat-form-field>
        <mat-selection-list [multiple]="false" class="rows-list">
          @for (row of rows(); track rowKey(row)) {
            <mat-list-option
              [selected]="selectedKey() === rowKey(row)"
              (click)="selectedKey.set(rowKey(row))"
            >
              <span class="row-label">#{{ row.index }}</span>
              @if (row.name) {
                <span class="row-name">{{ row.name }}</span>
              }
              <span class="row-snippet">{{ snippet(row) }}</span>
            </mat-list-option>
          }
        </mat-selection-list>
      </div>
      <div class="editor-pane">
        @if (selectedRow(); as row) {
          <app-game-text-editor
            [value]="row.text[lang()] ?? ''"
            (valueChange)="updateText(row, $event)"
          />
        } @else {
          <p class="empty">Selecione uma linha para editar.</p>
        }
      </div>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="cancel()">Cancelar</button>
      <button mat-flat-button color="primary" (click)="save()">Salvar</button>
    </mat-dialog-actions>
  `,
  styles: [
    `
      .rows-editor {
        display: grid;
        grid-template-columns: 280px 1fr;
        gap: 16px;
        height: 60vh;
      }
      .rows-pane {
        display: flex;
        flex-direction: column;
        gap: 8px;
        min-height: 0;
      }
      .rows-list {
        overflow: auto;
        border: 1px solid var(--mat-sys-outline-variant, #ddd);
        border-radius: 8px;
      }
      .editor-pane {
        overflow: auto;
        min-height: 0;
      }
      .row-label {
        font-weight: 600;
        margin-right: 6px;
      }
      .row-name {
        color: var(--mat-sys-primary, #3f51b5);
        margin-right: 6px;
      }
      .row-snippet {
        opacity: 0.75;
      }
      .empty {
        opacity: 0.7;
      }
    `,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class EntryRowsEditorDialogComponent {
  readonly data = inject<EntryRowsEditorDialogData>(MAT_DIALOG_DATA);
  private readonly ref = inject(MatDialogRef<EntryRowsEditorDialogComponent>);

  protected readonly rows = signal<dto.TextRow[]>(
    (this.data.entry.rows ?? []).map((r) => new dto.TextRow(r))
  );
  protected readonly languages = signal<common.Language[]>([]);
  protected readonly lang = signal<string>(DEFAULT_LANG);
  protected readonly selectedKey = signal<string | null>(null);

  protected readonly selectedRow = computed(() => {
    const key = this.selectedKey();
    if (key === null) return null;
    return this.rows().find((r) => this.rowKey(r) === key) ?? null;
  });

  constructor() {
    void this.loadLanguages();
  }

  private async loadLanguages(): Promise<void> {
    try {
      const langs = (await ListLanguages()) ?? [];
      if (langs.length > 0) {
        this.languages.set(langs);
        if (!langs.some((l) => l.code === this.lang())) {
          this.lang.set(langs[0].code);
        }
      }
    } catch {
      this.languages.set([{ code: DEFAULT_LANG, name: 'English' } as common.Language]);
    }
  }

  protected rowKey(row: dto.TextRow): string {
    return `${row.index}:${row.name ?? ''}`;
  }

  protected snippet(row: dto.TextRow): string {
    const text = row.text?.[this.lang()] ?? '';
    return text.length > 60 ? `${text.slice(0, 60)}…` : text;
  }

  protected updateText(row: dto.TextRow, value: string): void {
    const key = this.rowKey(row);
    this.rows.set(
      this.rows().map((r) =>
        this.rowKey(r) === key
          ? new dto.TextRow({ ...r, text: { ...(r.text ?? {}), [this.lang()]: value } })
          : r
      )
    );
  }

  protected cancel(): void {
    this.ref.close();
  }

  protected save(): void {
    this.ref.close(
      dto.FileEntry.createFrom({ metadata: this.data.entry.metadata, rows: this.rows() })
    );
  }
}
