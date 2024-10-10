import { Component, OnInit } from '@angular/core';
import { MessageService } from 'primeng/api';
import { ButtonModule } from 'primeng/button';
import { TreeTableModule } from 'primeng/treetable';
import { ToastModule } from 'primeng/toast';
import { CommonModule } from '@angular/common';
import { FfxTreeComponent } from './components/tree/tree.component';
import { ConfigModalComponent } from './components/config-modal/config-modal.component';
import { ProgressModalComponent } from './components/progress-modal/progress-modal.component';

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
export class AppComponent {}