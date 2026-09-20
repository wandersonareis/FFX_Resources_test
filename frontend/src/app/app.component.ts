import {
  ChangeDetectionStrategy,
  Component,
  inject,
  NgZone,
  OnInit,
  signal,
} from '@angular/core';
import { MessageService } from 'primeng/api';
import { CommonModule } from '@angular/common';
import { FfxTreeComponent } from './components/tree/tree.component';
import { ConfigModalComponent } from './components/config-modal/config-modal.component';
import { EventsEmit, EventsOn } from '../../wailsjs/runtime/runtime';
import { ToggleButton, type ToggleButtonChangeEvent } from 'primeng/togglebutton';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

const imports = [
  CommonModule,
  FormsModule,
  ReactiveFormsModule,
  FfxTreeComponent,
  ConfigModalComponent,
  ToggleButton,
];
@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
  imports: imports,
  providers: [MessageService],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class AppComponent implements OnInit {
  private readonly _messageService: MessageService = inject(MessageService);

  versionFFX = signal<boolean>(false);
  versionFFX2 = signal<boolean>(false);
  versionLastMiss = signal<boolean>(false);

  ngOnInit() {
    EventsOn('Notify', (data) => {
      let sticky: boolean = false;
      if (data.severity === 'error') {
        sticky = true;
      }

      this._messageService.add({
        severity: data.severity,
        summary: data.severity,
        detail: data.message,
        sticky: sticky,
      });
    });

    EventsOn('GameVersion', (data) => {
      console.log('GameVersion on init', data);
      let version: string = String(data);

      this.versionFFX.set(version === 'ffx');
      this.versionFFX2.set(version === 'ffx2');
      this.versionLastMiss.set(version === 'lastmiss');
    });
  }

  private resetVersions() {
    this.versionFFX.set(false);
    this.versionFFX2.set(false);
    this.versionLastMiss.set(false);
  }

  versionFFXChange(event: ToggleButtonChangeEvent) {
    this.resetVersions();
    this.versionFFX.set(true);
    console.log('versionFFXChange', event);
    EventsEmit('GameVersionChanged', 'ffx');
    EventsEmit('Refresh_Tree');
  }

  versionFFX2Change(event: ToggleButtonChangeEvent) {
    this.resetVersions();
    this.versionFFX2.set(true);
    console.log('versionFFX2Change', event);
    EventsEmit('GameVersionChanged', 'ffx2');
    EventsEmit('Refresh_Tree');
  }

  versionLastMissChange(event: ToggleButtonChangeEvent) {
    this.resetVersions();
    this.versionLastMiss.set(true);
    console.log('versionLastMissChange', event);
    EventsEmit('GameVersionChanged', 'lastmiss');
    EventsEmit('Refresh_Tree');
  }
}
