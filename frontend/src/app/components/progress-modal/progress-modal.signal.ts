import { signal } from "@angular/core";

export type progress = { File: string, Total: number, Processed: number, Percentage: number }
export const showProgress = signal<boolean>(false)
export const progress = signal<progress>({ File: "", Total: 0, Processed: 0, Percentage: 0 })