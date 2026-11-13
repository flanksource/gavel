import {
  sameSelection,
  type FileFinding,
  type FileRef,
  type FileTreeNode,
  type FileTreeSelection,
} from "@flanksource/clicky-ui/file-findings";

/** The tree rows from a root down to `selection`, or none when it is not in the tree. */
export function nodePath(nodes: readonly FileTreeNode[], selection: FileTreeSelection): FileTreeNode[] {
  for (const node of nodes) {
    if (sameSelection({ kind: node.kind, ref: node.ref }, selection)) return [node];
    const below = nodePath(node.children ?? [], selection);
    if (below.length) return [node, ...below];
  }
  return [];
}

const dirname = (path: string) => path.slice(0, path.lastIndexOf("/") + 1);

function normalize(path: string): string {
  const parts: string[] = [];
  for (const part of path.split("/")) {
    if (part === "" || part === ".") continue;
    if (part === ".." && parts.length && parts[parts.length - 1] !== "..")
      parts.pop();
    else parts.push(part);
  }
  return (path.startsWith("/") ? "/" : "") + parts.join("/");
}

/** The deepest directory, with a trailing slash, that holds every path. */
export function commonRoot(paths: readonly string[]): string {
  if (!paths.length) return "";
  let prefix = dirname(paths[0]!);
  for (const path of paths) {
    while (!path.startsWith(prefix)) prefix = dirname(prefix.slice(0, -1));
  }
  return prefix;
}

/** A lint finding's file as a tree ref relative to `root`; a file outside it keeps its absolute path. */
export function editorFindingRef(finding: FileFinding, root: string): FileRef {
  const absolute = normalize(
    finding.file.startsWith("/") || !finding.workDir
      ? finding.file
      : `${finding.workDir}/${finding.file}`,
  );
  return {
    file: absolute.startsWith(root) ? absolute.slice(root.length) : absolute,
  };
}

/** `target` as a posix path relative to the directory holding `from`. */
export function relativeToFile(from: string, target: string): string {
  const base = dirname(from).split("/").filter(Boolean);
  const parts = normalize(target).split("/").filter(Boolean);
  let shared = 0;
  while (
    shared < base.length &&
    shared < parts.length - 1 &&
    base[shared] === parts[shared]
  )
    shared++;
  return [
    ...base.slice(shared).map(() => ".."),
    ...parts.slice(shared),
  ].join("/");
}
