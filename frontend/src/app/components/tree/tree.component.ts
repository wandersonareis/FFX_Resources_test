import { ChangeDetectionStrategy, Component, effect, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TreeNode } from 'primeng/api';
import { TreeModule } from 'primeng/tree';
import { ContextMenuModule } from 'primeng/contextmenu';
import { ToastModule } from 'primeng/toast';
import { CommonModule } from '@angular/common';
import { ButtonModule } from 'primeng/button';
import { BuildTree } from '../../../../wailsjs/go/services/CollectionService';
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
    private readonly _ffxContextMenuService: FfxContextMenuService = inject(FfxContextMenuService);

    files = signal<TreeNode[]>([]);
    value = signal<number>(0);

    file = selectedFile
    items = this._ffxContextMenuService.items();

    async buildTree() {
        const treeNodes = await BuildTree();
        this.files.set(treeNodes)
    }

    async ngOnInit() {
        await this.buildTree();

        EventsOn("Refresh_Tree", async () => await this.buildTree())
        EventsOn("Progress", data => {
            progress.set(data)
            this.value.set(data.percentage)
        })
        EventsOn("ShowProgress", data => {
            showProgress.set(data)
            console.log("ShowProgress event", data);
        })
    }

    onNodeExpand(event: any) {
        findAndModifyNode(this.files(), event.node);
        console.log(event.node);

    }
}
