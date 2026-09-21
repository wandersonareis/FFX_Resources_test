import { ChangeDetectionStrategy, Component, input } from '@angular/core';
import { MatDialogModule } from '@angular/material/dialog';
import { MatProgressBarModule } from '@angular/material/progress-bar';

@Component({
  selector: 'app-progress-dialog',
  standalone: true,
  imports: [MatDialogModule, MatProgressBarModule],
  template: `
    <h2 mat-dialog-title>Processando…</h2>
    <mat-dialog-content>
      <mat-progress-bar mode="determinate" [value]="value()" />
    </mat-dialog-content>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ProgressDialogComponent {
  readonly value = input<number>(0);
}
