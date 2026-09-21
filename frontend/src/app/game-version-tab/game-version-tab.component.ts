import {
  ChangeDetectionStrategy,
  Component,
  OnInit,
  computed,
  inject,
  input,
  signal,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatMenuModule } from '@angular/material/menu';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { MatTreeModule, MatTreeNestedDataSource } from '@angular/material/tree';
import { NestedTreeControl } from '@angular/cdk/tree';
import {
  ChevronDown,
  ChevronRight,
  Download,
  EllipsisVertical,
  Eye,
  FileText,
  LUCIDE_ICONS,
  LucideAngularModule,
  LucideIconProvider,
  RefreshCw,
  Search,
  Upload,
} from 'lucide-angular';
import { firstValueFrom } from 'rxjs';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { dto } from '../../../wailsjs/go/models';
import { ErrorHandlerService } from '../../service/error-handler.service';
import { DisplayLabelService } from '../core/display-label.service';
import { ENTRY_KINDS, EntryRow, TreeDataService } from '../core/tree-data.service';
import { GameVersionId } from '../core/game-version.service';
import { EntryKind } from '../core/display-names';
import { EntryRowsEditorDialogComponent } from '../editor/entry-rows-editor-dialog/entry-rows-editor-dialog.component';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';

interface TreeListNode {
  label: string;
  kind?: EntryKind;
  entry?: EntryRow;
  children?: TreeListNode[];
}

@Component({
  selector: 'app-game-version-tab',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    LucideAngularModule,
    MatInputModule,
    MatMenuModule,
    MatSidenavModule,
    MatSnackBarModule,
    MatTableModule,
    MatTreeModule,
  ],
  templateUrl: './game-version-tab.component.html',
  styleUrl: './game-version-tab.component.css',
  providers: [
    {
      provide: LUCIDE_ICONS,
      multi: true,
      useValue: new LucideIconProvider({
        ChevronDown,
        ChevronRight,
        Download,
        EllipsisVertical,
        Eye,
        FileText,
        RefreshCw,
        Search,
        Upload,
      }),
    },
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class GameVersionTabComponent implements OnInit {
  readonly version = input.required<GameVersionId>();

  private readonly treeData = inject(TreeDataService);
  private readonly labels = inject(DisplayLabelService);
  private readonly errors = inject(ErrorHandlerService);
  private readonly snackBar = inject(MatSnackBar);
  private readonly dialog = inject(MatDialog);

  protected readonly treeControl = new NestedTreeControl<TreeListNode>(
    (node) => node.children ?? []
  );
  protected readonly treeSource = new MatTreeNestedDataSource<TreeListNode>();
  protected readonly tableSource = new MatTableDataSource<EntryRow>([]);
  protected readonly tableColumns = ['label', 'id', 'actions'];

  protected readonly activeKind = signal<EntryKind>('events');
  protected readonly selectedEntry = signal<EntryRow | null>(null);
  protected readonly fileEntry = signal<dto.FileEntry | null>(null);
  protected readonly loading = signal(false);

  protected readonly kindLabel = computed(() =>
    this.labels.kindLabel(this.activeKind())
  );

  hasChild = (_: number, node: TreeListNode): boolean =>
    !!node.children && node.children.length > 0;

  async ngOnInit(): Promise<void> {
    this.tableSource.filterPredicate = (row, filter) =>
      row.label.toLowerCase().includes(filter) ||
      row.id.toLowerCase().includes(filter);
    await this.reload();
    EventsOn('Refresh_Tree', async () => {
      await this.reload();
    });
  }

  protected applyFilter(event: Event): void {
    const value = (event.target as HTMLInputElement).value ?? '';
    this.tableSource.filter = value.trim().toLowerCase();
  }

  protected async reload(): Promise<void> {
    this.loading.set(true);
    try {
      const version = this.version();
      const roots: TreeListNode[] = [];
      const active: EntryRow[] = [];
      for (const kind of ENTRY_KINDS) {
        const entries = await this.treeData.loadKindEntries(kind, version);
        roots.push({
          label: `${this.labels.kindLabel(kind)} (${entries.length})`,
          kind,
          children: entries.map((entry) => ({
            label: entry.label,
            kind,
            entry,
          })),
        });
        if (kind === this.activeKind()) active.push(...entries);
      }
      this.treeSource.data = roots;
      this.tableSource.data = active;
      this.treeControl.expandAll();
    } catch (error) {
      this.errors.sendErrorNotification(error);
    } finally {
      this.loading.set(false);
    }
  }

  protected async selectNode(node: TreeListNode): Promise<void> {
    if (node.entry && node.kind) {
      this.activeKind.set(node.kind);
      await this.selectEntry(node.entry);
    } else if (node.kind) {
      this.activeKind.set(node.kind);
      this.selectedEntry.set(null);
      this.fileEntry.set(null);
      await this.reloadTableOnly();
    }
  }

  protected async selectEntry(entry: EntryRow): Promise<void> {
    this.selectedEntry.set(entry);
    this.loading.set(true);
    try {
      this.fileEntry.set(await this.treeData.loadEntry(entry.kind, entry.id, this.version()));
    } catch (error) {
      this.errors.sendErrorNotification(error);
      this.fileEntry.set(null);
    } finally {
      this.loading.set(false);
    }
  }

  private async reloadTableOnly(): Promise<void> {
    try {
      this.tableSource.data = await this.treeData.loadKindEntries(
        this.activeKind(),
        this.version()
      );
    } catch (error) {
      this.errors.sendErrorNotification(error);
    }
  }

  protected firstText(row: dto.TextRow): string {
    const values = Object.values(row.text ?? {});
    return values.length > 0 ? String(values[0]) : '';
  }

  protected async viewEntry(entry: EntryRow): Promise<void> {
    this.loading.set(true);
    try {
      const full = await this.treeData.loadEntry(entry.kind, entry.id, this.version());
      const ref = this.dialog.open(EntryRowsEditorDialogComponent, {
        width: '90vw',
        maxWidth: '1100px',
        data: { title: entry.label, entry: full },
      });
      const saved: dto.FileEntry | undefined = await firstValueFrom(ref.afterClosed());
      if (saved !== undefined) {
        await this.treeData.applyEntry(entry.kind, entry.id, this.version(), saved);
        this.snackBar.open('Alterações aplicadas.', 'Fechar', { duration: 3000 });
        await this.selectEntry(entry);
      }
    } catch (error) {
      this.errors.sendErrorNotification(error);
    } finally {
      this.loading.set(false);
    }
  }

  protected async extractEntry(entry: EntryRow): Promise<void> {
    try {
      const paths = await this.treeData.exportEntry(entry.kind, entry.id, this.version());
      this.snackBar.open(
        `Exportado (JSON + strings): ${paths.length} arquivo(s).`,
        'Fechar',
        { duration: 5000 }
      );
    } catch (error) {
      this.errors.sendErrorNotification(error);
    }
  }

  protected async importEntry(entry: EntryRow): Promise<void> {
    try {
      const paths = await this.treeData.importEntry(entry.kind, entry.id, this.version());
      this.snackBar.open(
        `Importado de: ${paths.join(', ')}`,
        'Fechar',
        { duration: 5000 }
      );
      await this.selectEntry(entry);
    } catch (error) {
      this.errors.sendErrorNotification(error);
    }
  }
}
