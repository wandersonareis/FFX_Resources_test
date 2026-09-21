import { Injectable } from '@angular/core';
import { EntryKind, KIND_LABELS, resolveEntryLabel } from './display-names';

@Injectable({ providedIn: 'root' })
export class DisplayLabelService {
  kindLabel(kind: EntryKind): string {
    return KIND_LABELS[kind];
  }

  entryLabel(kind: EntryKind, id: string): string {
    return resolveEntryLabel(kind, id);
  }
}
