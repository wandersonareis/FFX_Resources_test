import { signal } from "@angular/core";
import { TreeNode } from "primeng/api";
export const selectedFile = signal<TreeNode | undefined>(undefined);
export const extractedEditorText = signal<string>("");

export const gameDirectory = signal<string>("")

export const showEditorModal = signal<boolean>(false)
