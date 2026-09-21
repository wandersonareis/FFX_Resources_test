import { Injectable, signal } from '@angular/core';
import { EventsEmit, EventsOn } from '../../../wailsjs/runtime/runtime';

export const GAME_VERSIONS = [
  { id: 'ffx', label: 'FFX' },
  { id: 'ffx2', label: 'FFX-2' },
  { id: 'lastmiss', label: 'Last Mission' },
] as const;

export type GameVersionId = (typeof GAME_VERSIONS)[number]['id'];

function isGameVersionId(value: unknown): value is GameVersionId {
  return (
    typeof value === 'string' &&
    (GAME_VERSIONS as readonly { id: string }[]).some((v) => v.id === value)
  );
}

@Injectable({ providedIn: 'root' })
export class GameVersionService {
  readonly activeVersion = signal<GameVersionId>('ffx');

  constructor() {
    EventsOn('GameVersion', (data: unknown) => {
      if (isGameVersionId(data)) {
        this.activeVersion.set(data);
      }
    });
  }

  setVersion(version: GameVersionId): void {
    if (this.activeVersion() === version) return;
    this.activeVersion.set(version);
    EventsEmit('GameVersionChanged', version);
    EventsEmit('Refresh_Tree');
  }
}
