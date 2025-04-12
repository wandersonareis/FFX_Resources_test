import { core, fileFormats } from "../../wailsjs/go/models"

export function getFileInfoFromNode(node: any): fileFormats.TreeNodeData | null {
    if (node?.data) {
        return node.data as fileFormats.TreeNodeData
    }
    return null
}

export function wailsLog(func: Function, message: string) {
    func(message)
}
