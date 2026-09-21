import {
  ChangeDetectionStrategy,
  Component,
  OnDestroy,
  OnInit,
  WritableSignal,
  signal,
} from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { FolderOpen, LUCIDE_ICONS, LucideAngularModule, LucideIconProvider, X } from 'lucide-angular';
import {
  GetGameFilesLocation,
  GetTranslateLocation,
  SelectDirectory,
} from '../../../wailsjs/go/main/App';
import { EventsEmit, EventsOn } from '../../../wailsjs/runtime';

type LocationKey = 'GameFilesLocation' | 'TranslateLocation';

interface DirectoryInput {
  key: LocationKey;
  eventName: string;
  label: string;
  hint: string;
  dialogTitle: string;
  value: WritableSignal<string>;
}

/**
 * Diálogo de diretórios: apenas gamefiles e translated (fonte do
 * reimport, <game>/mods/translated por padrão). Extract/reimport são
 * internos do backend e não aparecem aqui.
 * Toda mudança é aplicada imediatamente (ao confirmar o campo ou
 * escolher no seletor): o backend persiste no config.json e re-emite
 * o valor + Refresh_Tree.
 */
@Component({
  selector: 'app-config-dialog',
  standalone: true,
  imports: [
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatSnackBarModule,
    LucideAngularModule,
  ],
  providers: [
    { provide: LUCIDE_ICONS, multi: true, useValue: new LucideIconProvider({ FolderOpen, X }) },
  ],
  template: `
    <h2 mat-dialog-title>Configurações</h2>
    <mat-dialog-content class="config-content">
      @for (item of inputs; track item.key) {
        <div class="config-row">
          <mat-form-field appearance="outline" class="config-field">
            <mat-label>{{ item.label }}</mat-label>
            <input
              matInput
              [value]="item.value()"
              (change)="applyFromInput(item, $event)"
              (keydown.enter)="applyFromInput(item, $event)"
            />
            <mat-hint>{{ item.hint }}</mat-hint>
          </mat-form-field>
          <button
            mat-icon-button
            (click)="selectDirectory(item)"
            [attr.aria-label]="'Selecionar ' + item.label"
            title="Procurar pasta"
          >
            <i-lucide name="folder-open" [size]="20"></i-lucide>
          </button>
        </div>
      }
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-flat-button color="primary" (click)="close()">
        <i-lucide name="x" [size]="18"></i-lucide> Fechar
      </button>
    </mat-dialog-actions>
  `,
  styles: [
    `
      .config-content {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        min-width: min(560px, 70vw);
        padding-top: 0.5rem;
      }
      .config-row {
        display: flex;
        align-items: center;
        gap: 0.5rem;
      }
      .config-field {
        flex: 1;
      }
    `,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ConfigDialogComponent implements OnInit, OnDestroy {
  private readonly ref: MatDialogRef<ConfigDialogComponent>;
  private readonly snackBar: MatSnackBar;
  private readonly unsubscribes: Array<() => void> = [];

  protected readonly gameDirectory = signal('');
  protected readonly translatedDirectory = signal('');

  protected inputs: DirectoryInput[] = [];

  constructor(ref: MatDialogRef<ConfigDialogComponent>, snackBar: MatSnackBar) {
    this.ref = ref;
    this.snackBar = snackBar;
  }

  ngOnInit(): void {
    this.unsubscribes.push(
      EventsOn('GameFilesLocation', (data: string) => this.gameDirectory.set(data ?? '')),
      EventsOn('TranslateLocation', (data: string) => this.translatedDirectory.set(data ?? ''))
    );

    this.inputs = [
      {
        key: 'GameFilesLocation',
        eventName: 'GameLocationChanged',
        label: 'Arquivos originais do jogo',
        hint: 'Padrão: <execução>/data',
        dialogTitle: 'Selecione a pasta dos arquivos originais do jogo',
        value: this.gameDirectory,
      },
      {
        key: 'TranslateLocation',
        eventName: 'TranslateLocationChanged',
        label: 'Arquivos traduzidos (reimport)',
        hint: 'Padrão: <jogo>/mods/translated',
        dialogTitle: 'Selecione a pasta dos arquivos traduzidos',
        value: this.translatedDirectory,
      },
    ];

    // Carga sob demanda (não depende do evento de startup do backend).
    GetGameFilesLocation()
      .then((v) => this.gameDirectory.set(v ?? ''))
      .catch((e) => this.notify(e));
    GetTranslateLocation()
      .then((v) => this.translatedDirectory.set(v ?? ''))
      .catch((e) => this.notify(e));
  }

  ngOnDestroy(): void {
    for (const off of this.unsubscribes) {
      try {
        off();
      } catch {
        // ignora falhas de unsubscribe
      }
    }
  }

  protected applyFromInput(item: DirectoryInput, event: Event): void {
    const value = (event.target as HTMLInputElement)?.value ?? '';
    this.apply(item, value);
  }

  protected async selectDirectory(item: DirectoryInput): Promise<void> {
    try {
      const path = await SelectDirectory(item.dialogTitle);
      if (path) this.apply(item, path);
    } catch (error) {
      this.notify(error);
    }
  }

  protected close(): void {
    this.ref.close();
  }

  private apply(item: DirectoryInput, raw: string): void {
    const path = (raw ?? '').trim();
    if (!path) {
      this.snackBar.open('Informe um diretório válido.', 'Fechar', { duration: 3000 });
      return;
    }
    if (path === item.value()) return;
    item.value.set(path);
    try {
      EventsEmit(item.eventName, path);
    } catch (error) {
      this.notify(error);
    }
  }

  private notify(error: unknown): void {
    const message = error instanceof Error ? error.message : String(error ?? 'Erro');
    this.snackBar.open(message, 'Fechar', { duration: 4000 });
    EventsEmit('Notify', { severity: 'error', message });
  }
}
