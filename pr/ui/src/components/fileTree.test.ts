import { describe, expect, it } from 'vitest';
import { buildFileTree, compareNodes, filesBelow, findFileTreeNode, firstFilePath, sortedChildren } from './fileTree';

interface Entry {
  path: string;
  size: number;
}

const entries: Entry[] = [
  { path: 'zeta.go', size: 1 },
  { path: 'cmd/run/main.go', size: 2 },
  { path: 'cmd/run/flags.go', size: 3 },
  { path: 'alpha.md', size: 4 },
  { path: 'cmd/root.go', size: 5 },
];

describe('buildFileTree', () => {
  it('nests files under their directories, directories before files, each alphabetical', () => {
    const roots = buildFileTree(entries);
    expect(roots.map(node => node.path)).toEqual(['cmd', 'alpha.md', 'zeta.go']);
    const cmd = roots[0];
    expect(sortedChildren(cmd).map(node => node.path)).toEqual(['cmd/run', 'cmd/root.go']);
    expect(sortedChildren(cmd.children.get('run')!).map(node => node.path)).toEqual(['cmd/run/flags.go', 'cmd/run/main.go']);
  });

  it('keeps the original entry on the file node and none on directories', () => {
    const roots = buildFileTree(entries);
    expect(roots[0].file).toBeUndefined();
    expect(roots[2].file).toBe(entries[0]);
  });

  it('returns no roots for no files', () => {
    expect(buildFileTree<Entry>([])).toEqual([]);
  });
});

describe('compareNodes', () => {
  it('orders a directory before a file regardless of name', () => {
    const roots = buildFileTree([{ path: 'a.txt' }, { path: 'z/inner.txt' }]);
    expect([...roots].reverse().sort(compareNodes).map(node => node.path)).toEqual(['z', 'a.txt']);
  });
});

describe('filesBelow', () => {
  it('collects every file under a directory', () => {
    const roots = buildFileTree(entries);
    expect(filesBelow(roots[0]).map(file => file.path).sort()).toEqual(['cmd/root.go', 'cmd/run/flags.go', 'cmd/run/main.go']);
  });

  it('returns the file itself for a file node', () => {
    const roots = buildFileTree(entries);
    expect(filesBelow(roots[1])).toEqual([entries[3]]);
  });
});

describe('findFileTreeNode', () => {
  it('finds nested files and directories by full path', () => {
    const roots = buildFileTree(entries);
    expect(findFileTreeNode(roots, 'cmd/run/main.go')?.file).toBe(entries[1]);
    expect(findFileTreeNode(roots, 'cmd/run')?.children.size).toBe(2);
  });

  it('returns null for an empty or unknown path', () => {
    const roots = buildFileTree(entries);
    expect(findFileTreeNode(roots, '')).toBeNull();
    expect(findFileTreeNode(roots, 'cmd/missing.go')).toBeNull();
  });
});

describe('firstFilePath', () => {
  it('is the first file in display order, descending into the first directory', () => {
    expect(firstFilePath(buildFileTree(entries))).toBe('cmd/run/flags.go');
  });

  it('is empty for an empty tree', () => {
    expect(firstFilePath(buildFileTree<Entry>([]))).toBe('');
  });
});
