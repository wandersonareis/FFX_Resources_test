import {Component, OnInit, Signal, signal, WritableSignal} from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { DialogModule } from 'primeng/dialog';
import { InputTextModule } from 'primeng/inputtext';
import { SelectDirectory } from '../../../../wailsjs/go/main/App';
import { EventsEmit, EventsOn } from '../../../../wailsjs/runtime';
import { gameDirectory } from '../signals/signals.signal';
import {FloatLabelModule} from "primeng/floatlabel";

@Component({
    selector: 'app-config-modal',
    imports: [
        ButtonModule,
        DialogModule,
        InputTextModule,
        ReactiveFormsModule,
        FloatLabelModule,
    ],
    templateUrl: './config-modal.component.html'
})
export class ConfigModalComponent implements OnInit {
  visible: WritableSignal<boolean> = signal<boolean>(false);

  extractedDirectory: WritableSignal<string> = signal<string>("");
  translatedDirectory: WritableSignal<string> = signal<string>("");
  importedDirectory: WritableSignal<string> = signal<string>("");

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
    EventsOn("GameFilesLocation", (data: string) => {
      gameDirectory.set(data)
    })

    EventsOn("ExtractLocation", (data: string) => {
      this.extractedDirectory.set(data)
    })

    EventsOn("TranslateLocation", (data: string) => {
      this.translatedDirectory.set(data)
    })

    EventsOn("ReimportLocation", (data: string) => {
      this.importedDirectory.set(data)
    })

    this.inputs = [
      {
        eventName: 'GameLocationChanged',
        label: 'Original files:',
        dialogTitle: 'Select game original files folder',
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
        dialogTitle: 'Select translated files folder',
        value: this.translatedDirectory,
      },
      {
        eventName: 'ReimportLocationChanged',
        label: 'Output files:',
        dialogTitle: 'Select reimported output folder',
        value: this.importedDirectory,
      }
    ];
  }


  showConfigModal() {
    this.visible.set(!this.visible());
  }

  saveWorkingDirectories() {
    this.showConfigModal()
    EventsEmit("Refresh_Tree")
    EventsEmit("SaveConfig")
  }

}
