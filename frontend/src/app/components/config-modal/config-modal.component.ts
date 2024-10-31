import { Component, OnInit, Signal, signal } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { DialogModule } from 'primeng/dialog';
import { InputTextModule } from 'primeng/inputtext';
import { SelectDirectory } from '../../../../wailsjs/go/main/App';
import { EventsEmit, EventsOnce } from '../../../../wailsjs/runtime/runtime';
import { gameDirectory } from '../signals/signals.signal';

@Component({
  selector: 'app-config-modal',
  standalone: true,
  imports: [
    ButtonModule,
    DialogModule,
    InputTextModule,
    ReactiveFormsModule,
  ],
  templateUrl: './config-modal.component.html',
})
export class ConfigModalComponent implements OnInit {
  visible = signal<boolean>(false);

  extractedDirectory = signal<string>("");
  translatedDirectory = signal<string>("");
  inputs: { eventName: string, label: string, dialogTitle: string, value: Signal<string> }[] = [];

  async selectDirectory(eventName: string, dialogTitle: string) {
    try {
      const path = await SelectDirectory(dialogTitle);
      if (path) {
        EventsEmit(eventName, path);
      }
    } catch (error) {
      EventsEmit("Notify", error);
    }
  }

  async ngOnInit() {
    EventsOnce("GameFilesLocation", (data: string) => {
      gameDirectory.set(data)
    })

    EventsOnce("ExtractLocation", (data: string) => {
      this.extractedDirectory.set(data)
    })

    EventsOnce("TranslateLocation", (data: string) => {
      this.translatedDirectory.set(data)
    })

    this.inputs = [
      {
        eventName: 'GameLocationChanged',
        label: 'Original files:',
        dialogTitle: 'Select original output folder',
        value: gameDirectory
      },
      {
        eventName: 'ExtractLocationChanged',
        label: 'Extracted files:',
        dialogTitle: 'Select extracted output folder',
        value: this.extractedDirectory,
      },
      {
        eventName: 'TranslateLocationChanged',
        label: 'Translated files:',
        dialogTitle: 'Select translated output folder',
        value: this.translatedDirectory,
      },
    ];
  }


  showConfigModal() {
    this.visible.set(!this.visible());
  }

  saveWorkingDirectories() {
    this.showConfigModal()
    EventsEmit("Refresh_Tree")
  }

}
