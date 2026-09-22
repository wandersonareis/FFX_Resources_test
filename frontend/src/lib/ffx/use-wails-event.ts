'use client';

import { useEffect, useRef } from 'react';
import { EventsOn } from '@/wailsjs/runtime/runtime';

/** Registra um listener Wails com cleanup automático. */
export function useWailsEvent(
  eventName: string,
  callback: (...data: unknown[]) => void
): void {
  const ref = useRef(callback);
  useEffect(() => {
    ref.current = callback;
  }, [callback]);
  useEffect(
    () => EventsOn(eventName, (...data) => ref.current(...data)),
    [eventName]
  );
}
