import { interactions } from "../../wailsjs/go/models"

export function extractFileInfo(node: any): interactions.GameDataInfo | null {
    if (node?.data) {
        return node.data as interactions.GameDataInfo
    }
    return null
}

export function wailsLog(func: Function, message: string) {
    func(message)
}