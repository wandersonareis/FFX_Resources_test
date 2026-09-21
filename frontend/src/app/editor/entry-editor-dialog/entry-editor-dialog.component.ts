import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { GameTextEditorComponent } from '../game-text/game-text-editor.component';

export interface EntryEditorDialogData {
  title: string;
  content: string;
}

@Component({
  selector: 'app-entry-editor-dialog',
  standalone: true,
  imports: [MatDialogModule, MatButtonModule, GameTextEditorComponent],
  template: `
    <h2 mat-dialog-title>{{ data.title }}</h2>
    <mat-dialog-content>
      <app-game-text-editor [value]="content()" (valueChange)="content.set($event)" />
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="cancel()">Cancelar</button>
      <button mat-flat-button color="primary" (click)="save()">Salvar</button>
    </mat-dialog-actions>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class EntryEditorDialogComponent {
  readonly data = inject<EntryEditorDialogData>(MAT_DIALOG_DATA);
  private readonly ref = inject(MatDialogRef<EntryEditorDialogComponent>);
  protected readonly content = signal(this.data.content);

  protected cancel(): void {
    this.ref.close();
  }

  protected save(): void {
    this.ref.close(this.content());
  }
}
