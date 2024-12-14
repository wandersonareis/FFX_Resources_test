import {
  ChangeDetectionStrategy,
  Component,
  computed,
  effect,
  input,
  InputSignal,
  signal,
  WritableSignal
} from '@angular/core';
import { ButtonModule } from 'primeng/button';
import { DialogModule } from 'primeng/dialog';
import { ProgressBarModule } from 'primeng/progressbar';
import { progress, showProgress } from './progress-modal.signal';

const imports = [
  ButtonModule,
  DialogModule,
  ProgressBarModule
];

@Component({
  selector: 'app-progress-modal',
  standalone: true,
  imports: imports,
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './progress-modal.component.css',
  template: `
    <p-dialog [modal]="true" [(visible)]="visible" [style]="{ width: '20rem' }">
      <p-progressBar [value]="value()" />
    </p-dialog>
  `,
})

export class ProgressModalComponent {
  visible: WritableSignal<boolean> = showProgress;
  currentProgress: WritableSignal<progress> = progress;
  value: InputSignal<number> = input<number>(0);
}
