import { ChangeDetectionStrategy, Component, inject, NgZone, OnInit, signal } from '@angular/core';
import { MessageService } from 'primeng/api';
import { CommonModule } from '@angular/common';
import { FfxTreeComponent } from './components/tree/tree.component';
import { ConfigModalComponent } from './components/config-modal/config-modal.component';
import { EventsEmit, EventsOn } from '../../wailsjs/runtime/runtime';
import { ToggleButton } from 'primeng/togglebutton';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { type ToggleButtonChangeEvent } from 'primeng/togglebutton/togglebutton.interface';

const imports = [
  CommonModule,
  FormsModule,
  ReactiveFormsModule,
  FfxTreeComponent,
  ConfigModalComponent,
  ToggleButton
]
@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
  imports: imports,
  providers: [MessageService],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AppComponent implements OnInit {
  private readonly _messageService: MessageService = inject(MessageService)

  versionFFX = signal<boolean>(false);
  versionFFX2 = signal<boolean>(false);

  ngOnInit() {
    EventsOn("Notify", data => {
      this._messageService.add({ severity: data.severity, summary: data.severity, detail: data.message })
    });

    EventsOn("GameVersion", data => {
      console.log("GameVersion on init", data);
      let version: number = parseInt(data);

      this.versionFFX.set(version === 1);
      this.versionFFX2.set(version === 2);
    });
  };

  versionFFXChange(event: ToggleButtonChangeEvent) {
    this.versionFFX2.set(false)
    console.log("versionFFXChange", event);
    EventsEmit("GameVersionChanged", 1);
    EventsEmit("Refresh_Tree");
  }

  versionFFX2Change(event: ToggleButtonChangeEvent) {
    this.versionFFX.set(false)
    console.log("versionFFX2Change", event);
    EventsEmit("GameVersionChanged", 2);
    EventsEmit("Refresh_Tree");
  }
}