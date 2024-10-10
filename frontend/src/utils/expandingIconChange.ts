import { TreeNode } from "primeng/api/treenode";

export function findAndModifyNode(nodes: TreeNode[], targetNode: TreeNode) {
    const openedFolderIcon = "pi pi-folder-open";

    for (const node of nodes) {
        if (node === targetNode) {
            node.icon = openedFolderIcon;
            return;
        }
      
        if (node.children && node.children.length > 0) {
            findAndModifyNode(node.children, targetNode);
        }
    }
}