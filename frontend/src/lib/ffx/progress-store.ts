'use client';

import { useSyncExternalStore } from 'react';

/** Estado de UMA ação (step) do batch — espelha backend/progress. */
export type ProgressStepState = 'pending' | 'running' | 'done' | 'error';

export interface ProgressStep {
  label: string;
  total: number;
  processed: number;
  issueCount: number;
  state: ProgressStepState;
  error?: string;
}

/**
 * Snapshot completo do batch — payload dos três eventos do wails
 * (ProgressStart/Update/End). Snapshots completos (nada de delta
 * acumulado): um evento perdido nunca corrompe a exibição.
 */
export interface ProgressBatch {
  title: string;
  steps: ProgressStep[];
  current: number;
  done: boolean;
  failed?: boolean;
  issueCount: number;
  written?: number;
  destination?: string;
}

/**
 * Store do modal de progresso (padrão exportSelection: singleton de módulo,
 * sem props atravessando componentes). O ProgressModal escreve aqui os
 * eventos do wails; quem precisa saber se um batch passou (o catch do
 * export decidindo entre toast e silêncio) lê o snapshot.
 *
 * É lido em callback de evento e via useSyncExternalStore — nunca em
 * render direto — então um objeto mutável de módulo basta.
 */
let snapshot: ProgressBatch | null = null;
const listeners = new Set<() => void>();

function publish(next: ProgressBatch | null): void {
  snapshot = next;
  for (const listener of listeners) listener();
}

export function subscribeProgress(listener: () => void): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function getProgressBatch(): ProgressBatch | null {
  return snapshot;
}

/** Hook reativo sobre o snapshot (re-render só quem assina). */
export function useProgressBatch(): ProgressBatch | null {
  return useSyncExternalStore(subscribeProgress, getProgressBatch);
}

/** O usuário fechou o modal de resultado: limpa o ciclo. */
export function clearProgressBatch(): void {
  publish(null);
}

/** Um batch já foi exibido nesta operação? (erro depois do início fica no
 * modal — o catch do caller não repete em toast). */
export function progressWasShown(): boolean {
  return snapshot !== null;
}

// ---- escrita: os eventos do wails (chamados pelo ProgressModal) ----------

export function progressStarted(batch: ProgressBatch): void {
  publish({ ...batch, done: false });
}

export function progressUpdated(batch: ProgressBatch): void {
  publish(batch);
}

export function progressEnded(batch: ProgressBatch): void {
  publish({ ...batch, done: true });
}
