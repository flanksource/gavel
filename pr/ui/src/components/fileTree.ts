export interface FileTreeNode<T extends { path: string }> {
  name: string;
  path: string;
  children: Map<string, FileTreeNode<T>>;
  file?: T;
}

export function buildFileTree<T extends { path: string }>(files: T[]): FileTreeNode<T>[] {
  const roots = new Map<string, FileTreeNode<T>>();
  for (const file of files) {
    const parts = file.path.split('/').filter(Boolean);
    let nodes = roots;
    let currentPath = '';
    parts.forEach((part, index) => {
      currentPath = currentPath ? `${currentPath}/${part}` : part;
      const node: FileTreeNode<T> = nodes.get(part) ?? { name: part, path: currentPath, children: new Map() };
      nodes.set(part, node);
      if (index === parts.length - 1) node.file = file;
      nodes = node.children;
    });
  }
  return [...roots.values()].sort(compareNodes);
}

// compareNodes orders directories before files, then alphabetically by name.
export function compareNodes<T extends { path: string }>(left: FileTreeNode<T>, right: FileTreeNode<T>) {
  const leftDirectory = left.children.size > 0;
  const rightDirectory = right.children.size > 0;
  if (leftDirectory !== rightDirectory) return leftDirectory ? -1 : 1;
  return left.name.localeCompare(right.name);
}

export function sortedChildren<T extends { path: string }>(node: FileTreeNode<T>) {
  return [...node.children.values()].sort(compareNodes);
}

export function filesBelow<T extends { path: string }>(node: FileTreeNode<T>): T[] {
  if (node.file) return [node.file];
  return [...node.children.values()].flatMap(filesBelow);
}

export function findFileTreeNode<T extends { path: string }>(nodes: FileTreeNode<T>[], path: string): FileTreeNode<T> | null {
  if (!path) return null;
  for (const node of nodes) {
    if (node.path === path) return node;
    const child = findFileTreeNode([...node.children.values()], path);
    if (child) return child;
  }
  return null;
}

// firstFilePath is the path of the first file in the tree's display order
// (directories first, depth-first), or '' for an empty tree.
export function firstFilePath<T extends { path: string }>(roots: FileTreeNode<T>[]): string {
  for (const node of roots) {
    if (node.file) return node.path;
    const nested = firstFilePath(sortedChildren(node));
    if (nested) return nested;
  }
  return '';
}
