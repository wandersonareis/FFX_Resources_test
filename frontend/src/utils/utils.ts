import { core } from "../../wailsjs/go/models"

export function extractFileInfo(node: any): core.SpiraFileInfo | null {
    if (node?.data) {
        return node.data as core.SpiraFileInfo
    }
    return null
}

export function wailsLog(func: Function, message: string) {
    func(message)
}
