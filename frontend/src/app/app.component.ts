import {
  ChangeDetectionStrategy,
  Component,
  OnInit,
  effect,
  inject,
  signal,
} from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTabsModule } from '@angular/material/tabs';
import { MatToolbarModule } from '@angular/material/toolbar';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { ConfigDialogComponent } from './config-dialog/config-dialog.component';
import { TabStorageService } from './core/tab-storage.service';
import { GAME_VERSIONS, GameVersionService } from './core/game-version.service';
import { GameVersionTabComponent } from './game-version-tab/game-version-tab.component';
import { ProgressDialogComponent } from './progress-dialog/progress-dialog.component';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
  imports: [
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
    MatSnackBarModule,
    MatTabsModule,
    MatToolbarModule,
    GameVersionTabComponent,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class AppComponent implements OnInit {
  private readonly versions = inject(GameVersionService);
  private readonly tabStorage = inject(TabStorageService);
  private readonly snackBar = inject(MatSnackBar);
  private readonly dialog = inject(MatDialog);

  protected readonly tabs = GAME_VERSIONS;
  protected readonly selectedIndex = signal(this.tabStorage.load());
  private progressRef: MatDialogRef<ProgressDialogComponent> | null = null;

  constructor() {
    // Mantém a aba sincronizada quando a versão muda (evento do backend).
    effect(() => {
      const index = this.tabs.findIndex((t) => t.id === this.versions.activeVersion());
      if (index >= 0 && index !== this.selectedIndex()) {
        this.selectedIndex.set(index);
        this.tabStorage.save(index);
      }
    });
  }

  ngOnInit(): void {
    EventsOn('Notify', (data: { severity?: string; message?: string }) => {
      const sticky = data?.severity === 'error';
      this.snackBar.open(data?.message ?? 'Notificação', 'Fechar', {
        duration: sticky ? undefined : 4000,
      });
    });

    EventsOn('ShowProgress', (visible: unknown) => {
      if (visible && !this.progressRef) {
        this.progressRef = this.dialog.open(ProgressDialogComponent, {
          width: '320px',
          disableClose: true,
        });
      } else if (!visible && this.progressRef) {
        this.progressRef.close();
        this.progressRef = null;
      }
    });

    EventsOn('Progress', (data: { percentage?: number }) => {
      this.progressRef?.componentRef?.setInput('value', data?.percentage ?? 0);
    });
  }

  protected onTabChange(index: number): void {
    this.selectedIndex.set(index);
    this.tabStorage.save(index);
    this.versions.setVersion(this.tabs[index].id);
  }

  protected openConfig(): void {
    this.dialog.open(ConfigDialogComponent, { width: '640px', maxWidth: '90vw' });
  }
}
