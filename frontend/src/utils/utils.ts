import { lib } from "../../wailsjs/go/models"

export function extractFileInfo(node: any): lib.FileInfo | null {
    if (!node) {
        return null
    }

    if (node?.data) {
        return node.data as lib.FileInfo
    }
    return null
}

export function wailsLog(func: Function, message: string) {
    func(message)
}