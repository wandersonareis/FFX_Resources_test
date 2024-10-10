import { signal } from "@angular/core";

export const showProgress = signal<boolean>(false)
export const progress = signal<{ Total: number, Processed: number, Percentage: number }>({ Total: 0, Processed: 0, Percentage: 0 })