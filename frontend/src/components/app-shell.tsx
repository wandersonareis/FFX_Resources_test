'use client';

import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';
import { Settings } from 'lucide-react';
import { QuitApp } from '@/wailsjs/go/main/App';
import { EventsEmit } from '@/wailsjs/runtime/runtime';
import { GAME_VERSIONS, GameVersionId, isGameVersionId } from '@/lib/ffx/game-version';
import { tabStorage } from '@/lib/ffx/tab-storage';
import { useWailsEvent } from '@/lib/ffx/use-wails-event';
import { saveAllDrafts } from '@/lib/ffx/save-all';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { GameVersionTab } from '@/components/game-version-tab';
import { ConfigDialog } from '@/components/dialogs/config-dialog';
import { ProgressDialog } from '@/components/dialogs/progress-dialog';

export function AppShell() {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [saving, setSaving] = useState(false);
  const [configOpen, setConfigOpen] = useState(false);
  const [progress, setProgress] = useState({ open: false, value: 0 });

  useEffect(() => {
    // Hydration-safe: localStorage só existe no cliente.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setSelectedIndex(tabStorage.load());
  }, []);

  const setVersion = useCallback((version: GameVersionId) => {
    EventsEmit('GameVersionChanged', version);
    EventsEmit('Refresh_Tree');
  }, []);

  useWailsEvent('Notify', (data) => {
    const payload = data as { severity?: string; message?: string };
    const sticky = payload?.severity === 'error';
    if (sticky) {
      toast.error(payload?.message ?? 'Notificação', { duration: Infinity });
    } else {
      toast(payload?.message ?? 'Notificação', { duration: 4000 });
    }
  });

  useWailsEvent('ShowProgress', (visible) => {
    setProgress((p) => ({ ...p, open: Boolean(visible) }));
  });

  useWailsEvent('Progress', (data) => {
    const payload = data as { percentage?: number };
    setProgress((p) => ({ ...p, value: payload?.percentage ?? 0 }));
  });

  useWailsEvent('GameVersion', (data) => {
    if (isGameVersionId(data)) {
      const index = GAME_VERSIONS.findIndex((t) => t.id === data);
      if (index >= 0 && index !== selectedIndex) {
        setSelectedIndex(index);
        tabStorage.save(index);
      }
    }
  });

  useWailsEvent('SaveRequested', () => {
    void saveAllDrafts(() => saving, setSaving).then(() => QuitApp());
  });

  const onTabChange = (index: string) => {
    const idx = Number.parseInt(index, 10);
    setSelectedIndex(idx);
    tabStorage.save(idx);
    setVersion(GAME_VERSIONS[idx].id);
  };

  return (
    <div className="flex h-screen flex-col">
      <header className="sticky top-0 z-10 flex items-center gap-2 bg-background px-4 py-2 border-b">
        <span className="text-lg font-semibold">FFX Resources</span>
        <span className="flex-1" />
        <Button variant="ghost" size="icon" onClick={() => setConfigOpen(true)} aria-label="Configurações">
          <Settings size={22} />
        </Button>
      </header>

      <Tabs value={String(selectedIndex)} onValueChange={onTabChange} className="flex-1 min-h-0 flex flex-col">
        <TabsList className="mx-4 mt-2 w-fit">
          {GAME_VERSIONS.map((tab, index) => (
            <TabsTrigger key={tab.id} value={String(index)}>
              {tab.label}
            </TabsTrigger>
          ))}
        </TabsList>
        {GAME_VERSIONS.map((tab, index) => (
          <TabsContent key={tab.id} value={String(index)} className="flex-1 min-h-0 mt-0">
            <GameVersionTab version={tab.id} />
          </TabsContent>
        ))}
      </Tabs>

      <ConfigDialog open={configOpen} onOpenChange={setConfigOpen} />
      <ProgressDialog open={progress.open} value={progress.value} />
    </div>
  );
}
