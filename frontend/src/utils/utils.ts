import {spira} from "../../wailsjs/go/models";

export function extractFileInfo(node: any): spira.GameDataInfo | null {
    if (node?.data) {
        return node.data as spira.GameDataInfo
    }
    return null
}

export function wailsLog(func: Function, message: string) {
    func(message)
}
