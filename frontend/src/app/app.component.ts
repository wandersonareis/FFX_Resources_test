import { Component, inject, OnInit } from '@angular/core';
import { MessageService } from 'primeng/api';
import { ButtonModule } from 'primeng/button';
import { TreeTableModule } from 'primeng/treetable';
import { ToastModule } from 'primeng/toast';
import { CommonModule } from '@angular/common';
import { FfxTreeComponent } from './components/tree/tree.component';
import { ConfigModalComponent } from './components/config-modal/config-modal.component';
import { ProgressModalComponent } from './components/progress-modal/progress-modal.component';
import { EventsOn } from '../../wailsjs/runtime/runtime';

const imports = [
  CommonModule,
  TreeTableModule,
  ToastModule,
  ButtonModule,
  FfxTreeComponent,
  ConfigModalComponent,
  ProgressModalComponent,
]
@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  standalone: true,
  imports: imports,
  providers: [MessageService]
})
export class AppComponent implements OnInit {
  private readonly _messageService: MessageService = inject(MessageService)
  ngOnInit() {
    EventsOn("Notify", data => {
      this._messageService.add({ severity: data.severity, summary: data.severity, detail: data.message })
    }
    )
  }

}