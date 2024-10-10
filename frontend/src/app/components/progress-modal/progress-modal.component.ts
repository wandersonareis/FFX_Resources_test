import { ChangeDetectionStrategy, Component, signal } from '@angular/core';
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
  templateUrl: './progress-modal.component.html',
  styleUrl: './progress-modal.component.css'
})
export class ProgressModalComponent {
  visible = showProgress;
  currentProgress = progress().Percentage;

}
