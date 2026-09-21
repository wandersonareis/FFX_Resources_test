import { ChangeDetectionStrategy, Component, OnInit, Signal, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { SelectDirectory } from '../../../wailsjs/go/main/App';
import { EventsEmit, EventsOn } from '../../../wailsjs/runtime';

interface DirectoryInput {
  eventName: string;
  label: string;
  dialogTitle: string;
  value: Signal<string>;
}

@Component({
  selector: 'app-config-dialog',
  standalone: true,
  imports: [MatButtonModule, MatDialogModule, MatFormFieldModule, MatIconModule, MatInputModule],
  template: `
    <h2 mat-dialog-title>Configurações</h2>
    <mat-dialog-content class="config-content">
      @for (item of inputs; track item.label) {
        <div class="config-row">
          <mat-form-field appearance="outline" class="config-field">
            <mat-label>{{ item.label }}</mat-label>
            <input matInput [value]="item.value()" readonly />
          </mat-form-field>
          <button
            mat-icon-button
            (click)="selectDirectory(item.eventName, item.dialogTitle)"
            [attr.aria-label]="'Selecionar ' + item.label"
          >
            <mat-icon>folder_open</mat-icon>
          </button>
        </div>
      }
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="close()">Cancelar</button>
      <button mat-flat-button color="primary" (click)="save()">Salvar</button>
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
export class ConfigDialogComponent implements OnInit {
  private readonly ref: MatDialogRef<ConfigDialogComponent>;

  protected readonly gameDirectory = signal('');
  protected readonly extractedDirectory = signal('');
  protected readonly translatedDirectory = signal('');
  protected readonly importedDirectory = signal('');

  protected inputs: DirectoryInput[] = [];

  constructor(ref: MatDialogRef<ConfigDialogComponent>) {
    this.ref = ref;
  }

  ngOnInit(): void {
    EventsOn('GameFilesLocation', (data: string) => this.gameDirectory.set(data));
    EventsOn('ExtractLocation', (data: string) => this.extractedDirectory.set(data));
    EventsOn('TranslateLocation', (data: string) => this.translatedDirectory.set(data));
    EventsOn('ReimportLocation', (data: string) => this.importedDirectory.set(data));

    this.inputs = [
      {
        eventName: 'GameLocationChanged',
        label: 'Arquivos originais',
        dialogTitle: 'Selecione a pasta dos arquivos originais do jogo',
        value: this.gameDirectory,
      },
      {
        eventName: 'ExtractLocationChanged',
        label: 'Arquivos extraídos',
        dialogTitle: 'Selecione a pasta de saída da extração',
        value: this.extractedDirectory,
      },
      {
        eventName: 'TranslateLocationChanged',
        label: 'Arquivos traduzidos',
        dialogTitle: 'Selecione a pasta dos arquivos traduzidos',
        value: this.translatedDirectory,
      },
      {
        eventName: 'ReimportLocationChanged',
        label: 'Arquivos de saída',
        dialogTitle: 'Selecione a pasta de saída da reimportação',
        value: this.importedDirectory,
      },
    ];
  }

  protected async selectDirectory(eventName: string, dialogTitle: string): Promise<void> {
    try {
      const path = await SelectDirectory(dialogTitle);
      if (path) EventsEmit(eventName, path);
    } catch (error) {
      EventsEmit('Notify', error);
    }
  }

  protected close(): void {
    this.ref.close();
  }

  protected save(): void {
    this.ref.close();
    EventsEmit('Refresh_Tree');
    EventsEmit('SaveConfig');
  }
}
