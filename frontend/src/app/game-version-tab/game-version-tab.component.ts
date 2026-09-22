import {
  ChangeDetectionStrategy,
  Component,
  OnInit,
  computed,
  effect,
  inject,
  input,
  signal,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatButtonModule } from '@angular/material/button';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { MatTreeModule, MatTreeNestedDataSource } from '@angular/material/tree';
import { NestedTreeControl } from '@angular/cdk/tree';
import {
  ChevronDown,
  ChevronRight,
  FileText,
  LUCIDE_ICONS,
  LucideAngularModule,
  LucideIconProvider,
  RefreshCw,
  Save,
} from 'lucide-angular';
import { firstValueFrom } from 'rxjs';
import {
  ApplyTextCollection,
  ListLanguages,
  QuitApp,
  SetUnsavedEdits,
} from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { dto } from '../../../wailsjs/go/models';
import { ErrorHandlerService } from '../../service/error-handler.service';
import { EditDraftService } from '../core/edit-draft.service';
import { DisplayLabelService } from '../core/display-label.service';
import { ENTRY_KINDS, EntryRow, TreeDataService } from '../core/tree-data.service';
import { GameVersionId } from '../core/game-version.service';
import {
  EntryKind,
  shortenedLabel,
  shortenedOf,
} from '../core/display-names';
import {
  SOURCE_LANG,
  TranslationCellDialogComponent,
} from '../editor/translation-cell-dialog/translation-cell-dialog.component';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';

interface SideNode {
  label: string;
  kind?: EntryKind;
  entry?: EntryRow;
  children?: SideNode[];
}

const KIND_COLUMNS: Record<EntryKind, string[]> = {
  events: ['index', 'original', 'translated'],
  objects: ['index', 'name', 'original', 'translated'],
  macro: ['index', 'original', 'translated'],
};

@Component({
  selector: 'app-game-version-tab',
  standalone: true,
  imports: [
    MatButtonModule,
    MatDialogModule,
    LucideAngularModule,
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
        FileText,
        RefreshCw,
        Save,
      }),
    },
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class GameVersionTabComponent implements OnInit {
  readonly version = input.required<GameVersionId>();

  private readonly treeData = inject(TreeDataService);
  private readonly labels = inject(DisplayLabelService);
  protected readonly drafts = inject(EditDraftService);
  private readonly errors = inject(ErrorHandlerService);
  private readonly snackBar = inject(MatSnackBar);
  private readonly dialog = inject(MatDialog);

  protected readonly treeControl = new NestedTreeControl<SideNode>(
    (node) => node.children ?? []
  );
  protected readonly treeSource = new MatTreeNestedDataSource<SideNode>();
  protected readonly tableSource = new MatTableDataSource<dto.TextRow>([]);

  protected readonly activeKind = signal<EntryKind>('events');
  protected readonly selectedEntry = signal<EntryRow | null>(null);
  protected readonly loading = signal(false);
  protected readonly saving = signal(false);

  private currentRows: dto.TextRow[] = [];

  protected readonly kindLabel = computed(() =>
    this.labels.kindLabel(this.activeKind())
  );

  protected readonly displayedColumns = computed(() =>
    KIND_COLUMNS[this.activeKind()]
  );

  hasChild = (_: number, node: SideNode): boolean =>
    !!node.children && node.children.length > 0;

  constructor() {
    effect(() => {
      void SetUnsavedEdits(this.drafts.hasDirty());
    });
  }

  async ngOnInit(): Promise<void> {
    void this.loadLanguages();
    await this.reload();
    EventsOn('SaveRequested', async () => {
      await this.saveAll();
      await QuitApp();
    });
  }

  protected async reload(): Promise<void> {
    this.loading.set(true);
    try {
      const version = this.version();
      const roots: SideNode[] = [];
      for (const kind of ENTRY_KINDS) {
        const entries = await this.treeData.loadKindEntries(kind, version);
        roots.push({
          label: `${this.labels.kindLabel(kind)} (${entries.length})`,
          kind,
          children: kind === 'events' ? this.eventGroups(entries) : this.leaves(entries),
        });
      }
      this.treeSource.data = roots;
      this.treeControl.expandAll();
      await this.reselect();
    } catch (error) {
      this.errors.sendErrorNotification(error);
    } finally {
      this.loading.set(false);
    }
  }

  private eventGroups(entries: EntryRow[]): SideNode[] {
    const groups = new Map<string, EntryRow[]>();
    for (const entry of entries) {
      const short = shortenedOf(entry.id);
      const list = groups.get(short) ?? [];
      list.push(entry);
      groups.set(short, list);
    }
    const nodes: SideNode[] = [];
    for (const short of [...groups.keys()].sort()) {
      const files = groups.get(short) ?? [];
      nodes.push({
        label: `${shortenedLabel(short)} (${files.length})`,
        kind: 'events',
        children: this.leaves(files),
      });
    }
    return nodes;
  }

  private leaves(entries: EntryRow[]): SideNode[] {
    return entries.map((entry) => ({
      label: entry.label,
      kind: entry.kind,
      entry,
    }));
  }

  /** Mantém a seleção atual após um reload, se o arquivo ainda existir. */
  private async reselect(): Promise<void> {
    const current = this.selectedEntry();
    if (!current) return;
    const entries = await this.treeData.loadKindEntries(current.kind, this.version());
    const found = entries.find((e) => e.id === current.id);
    if (!found) {
      this.selectedEntry.set(null);
      this.tableSource.data = [];
      this.currentRows = [];
      return;
    }
    await this.selectEntry(found);
  }

  protected async selectNode(node: SideNode): Promise<void> {
    if (node.entry && node.kind) {
      await this.selectEntry(node.entry);
    } else if (node.kind) {
      this.activeKind.set(node.kind);
    }
  }

  protected async selectEntry(entry: EntryRow): Promise<void> {
    this.loading.set(true);
    try {
      this.activeKind.set(entry.kind);
      this.selectedEntry.set(entry);
      const full = await this.treeData.loadEntry(entry.kind, entry.id, this.version());
      this.drafts.setBase(this.version(), entry.kind, entry.id, full);
      this.currentRows = full.rows ?? [];
      this.tableSource.data = [...this.currentRows];
    } catch (error) {
      this.errors.sendErrorNotification(error);
      this.selectedEntry.set(null);
      this.currentRows = [];
      this.tableSource.data = [];
    } finally {
      this.loading.set(false);
    }
  }

  protected original(row: dto.TextRow): string {
    return row.text?.[SOURCE_LANG] ?? '';
  }

  protected translatedOf(row: dto.TextRow): string {
    const entry = this.selectedEntry();
    if (!entry) return this.original(row);
    const edited = this.drafts.editOf(
      this.version(),
      entry.kind,
      entry.id,
      row,
      SOURCE_LANG
    );
    return edited ?? this.original(row);
  }

  protected isEdited(row: dto.TextRow): boolean {
    const entry = this.selectedEntry();
    if (!entry) return false;
    return (
      this.drafts.editOf(this.version(), entry.kind, entry.id, row, SOURCE_LANG) !==
      undefined
    );
  }

  protected async openTranslation(row: dto.TextRow): Promise<void> {
    const entry = this.selectedEntry();
    if (!entry) return;
    const ref = this.dialog.open(TranslationCellDialogComponent, {
      width: '720px',
      maxWidth: '90vw',
      data: { row, languages: this.languages },
    });
    const value: string | undefined = await firstValueFrom(ref.afterClosed());
    if (value === undefined) return;
    this.drafts.setCell(
      this.version(),
      entry.kind,
      entry.id,
      row,
      SOURCE_LANG,
      value
    );
    this.tableSource.data = [...this.currentRows];
  }

  protected async saveAll(): Promise<void> {
    if (!this.drafts.hasDirty() || this.saving()) return;
    this.saving.set(true);
    try {
      for (const [version, byKind] of this.drafts.dirtyBatches()) {
        for (const kind of byKind.keys()) {
          const collection = this.drafts.buildCollection(version, kind);
          if (Object.keys(collection).length === 0) continue;
          await ApplyTextCollection(
            kind,
            version as unknown as Parameters<typeof ApplyTextCollection>[1],
            collection as unknown as Parameters<typeof ApplyTextCollection>[2]
          );
        }
      }
      this.drafts.clearAll();
      this.snackBar.open('Alterações salvas no jogo.', 'Fechar', { duration: 3000 });
      await this.reselect();
    } catch (error) {
      this.errors.sendErrorNotification(error);
    } finally {
      this.saving.set(false);
    }
  }

  private async loadLanguages(): Promise<void> {
    try {
      this.languages = (await ListLanguages()) ?? [];
    } catch {
      this.languages = [{ code: SOURCE_LANG, name: 'English' }];
    }
  }

  private languages: Array<{ code: string; name: string }> = [];
}
