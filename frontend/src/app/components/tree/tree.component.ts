import { ChangeDetectionStrategy, Component, effect, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { MessageService, TreeNode } from 'primeng/api';
import { TreeModule } from 'primeng/tree';
import { ContextMenuModule } from 'primeng/contextmenu';
import { ToastModule } from 'primeng/toast';
import { CommonModule } from '@angular/common';
import { ButtonModule } from 'primeng/button';
import { PopulateTree } from '../../../../wailsjs/go/services/CollectionService';
import { EventsOn } from '../../../../wailsjs/runtime/runtime';
import { FfxContextMenuService } from '../../../service/ffx-context-menu.service';
import { selectedFile } from '../signals/signals.signal';
import { findAndModifyNode } from '../../../utils/expandingIconChange';
import { EditorModalComponent } from '../editor-modal/editor-modal.component';
import { progress, showProgress } from '../progress-modal/progress-modal.signal';
import { ProgressModalComponent } from '../progress-modal/progress-modal.component';

const imports = [
    CommonModule,
    TreeModule,
    EditorModalComponent,
    ToastModule,
    ContextMenuModule,
    ButtonModule,
    FormsModule,
    ProgressModalComponent,
]

@Component({
    selector: 'ffx-tree',
    exportAs: 'ffxTree',
    standalone: true,
    imports: imports,
    changeDetection: ChangeDetectionStrategy.OnPush,
    templateUrl: './tree.component.html',
})
export class FfxTreeComponent implements OnInit {
    private readonly _messageService: MessageService = inject(MessageService);
    private readonly _ffxContextMenuService: FfxContextMenuService = inject(FfxContextMenuService);

    files = signal<TreeNode[]>([]);
    value = signal<number>(0);

    file = selectedFile
    items = this._ffxContextMenuService.items();

    async treePolulation() {
        const treeNodes = await PopulateTree();
        this.files.set(treeNodes)
    }

    async ngOnInit() {
        //EventsOn("ApplicationError", data => this._messageService.add({ severity: 'error', summary: 'Error', detail: data }))
        EventsOn("Refresh_Tree", async () => await this.treePolulation())
        EventsOn("Progress", data => {
            //console.log(data);
            progress.set(data)
            this.value.set(data.percentage)
        })
        EventsOn("ShowProgress", data => {
            showProgress.set(data)
            console.log("ShowProgress event", data);

        })

        await this.treePolulation();
    }

    onNodeExpand(event: any) {
        findAndModifyNode(this.files(), event.node);
        //console.log(event.node);

    }

    constructor() {
        effect(() => {
            console.log("showProgress value on tree component", showProgress());

        })
    }
}
