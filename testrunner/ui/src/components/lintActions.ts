import type { Test, LinterResult } from '../types';
import { relPath } from '../utils';

export interface IgnoreRequest {
  source?: string;
  rule?: string;
  file?: string;
  work_dir?: string;
}

export interface LintAction {
  key: string;
  label: string;
  title: string;
  req: IgnoreRequest;
  variant?: 'primary' | 'subtle';
  disabledWithoutWorkDir?: boolean;
}

export function folderPattern(path: string | undefined): string {
  if (!path) return '**';
  const trimmed = path.replace(/\/+$/, '');
  return trimmed ? `${trimmed}/**` : '**';
}

interface FolderLintStat {
  linter: string;
  count: number;
  workDir?: string;
}

export function collectFolderLintStats(folder: Test, lint: LinterResult[] | undefined): FolderLintStat[] {
  const targetPath = folder.target_path || '';
  const counts = new Map<string, FolderLintStat>();
  for (const lr of lint || []) {
    if (folder.work_dir && lr.work_dir && lr.work_dir !== folder.work_dir) continue;
    for (const violation of lr.violations || []) {
      const rawFile = relPath(violation.file, lr.work_dir);
      if (!rawFile) continue;
      const matches = targetPath === '' ? true : rawFile === targetPath || rawFile.startsWith(`${targetPath}/`);
      if (!matches) continue;
      const current = counts.get(lr.linter);
      if (current) {
        current.count += 1;
      } else {
        counts.set(lr.linter, { linter: lr.linter, count: 1, workDir: lr.work_dir });
      }
    }
  }
  return Array.from(counts.values()).sort((a, b) => b.count - a.count || a.linter.localeCompare(b.linter));
}

// lintFolderActions returns the ordered set of "Ignore …" buttons that should
// render in the FolderLintDetail Actions section. The result is driven by which
// of {linterName, ruleName} the node carries:
//
//   - linter + rule  → 3-button rule matrix (folder / repo / disable-linter)
//   - linter only    → 3-button linter matrix (folder / linter-in-folder / disable-linter)
//   - neither        → multi-linter folder: "everything in folder" + per-linter pair
export function lintFolderActions(node: Test, lint: LinterResult[] | undefined): LintAction[] {
  const pattern = folderPattern(node.target_path);
  const linter = node.linterName || '';
  const rule = node.ruleName || '';
  const stats = (linter && rule) || linter ? [] : collectFolderLintStats(node, lint);
  const workDir = node.work_dir || (stats.length === 1 ? stats[0].workDir : '') || '';

  if (linter && rule) {
    return [
      {
        key: 'rule-folder',
        label: `Ignore ${rule} in this folder`,
        title: 'Add {source, rule, file} to .gavel.yaml',
        req: { source: linter, rule, file: pattern, work_dir: workDir },
        disabledWithoutWorkDir: true,
      },
      {
        key: 'rule-repo',
        label: `Ignore rule ${rule} everywhere`,
        title: 'Add {source, rule} to .gavel.yaml',
        req: { source: linter, rule, work_dir: workDir },
      },
      {
        key: 'linter-disable',
        label: `Disable ${linter} entirely`,
        title: 'Add {source} to .gavel.yaml',
        req: { source: linter, work_dir: workDir },
        variant: 'subtle',
      },
    ];
  }

  if (linter) {
    return [
      {
        key: 'folder-everything',
        label: 'Ignore everything in this folder',
        title: 'Add {file} to .gavel.yaml',
        req: { file: pattern, work_dir: workDir },
        disabledWithoutWorkDir: true,
      },
      {
        key: 'linter-folder',
        label: `Ignore ${linter} in this folder`,
        title: 'Add {source, file} to .gavel.yaml',
        req: { source: linter, file: pattern, work_dir: workDir },
        disabledWithoutWorkDir: true,
      },
      {
        key: 'linter-disable',
        label: `Disable ${linter} entirely`,
        title: 'Add {source} to .gavel.yaml',
        req: { source: linter, work_dir: workDir },
        variant: 'subtle',
      },
    ];
  }

  const actions: LintAction[] = [
    {
      key: 'folder-everything',
      label: 'Ignore everything in this folder',
      title: 'Add {file} to .gavel.yaml',
      req: { file: pattern, work_dir: workDir },
      disabledWithoutWorkDir: true,
    },
  ];
  for (const stat of stats) {
    actions.push({
      key: `linter-folder:${stat.linter}`,
      label: `Ignore ${stat.linter} in this folder`,
      title: 'Add {source, file} to .gavel.yaml',
      req: { source: stat.linter, file: pattern, work_dir: workDir },
      variant: 'subtle',
      disabledWithoutWorkDir: true,
    });
    actions.push({
      key: `linter-disable:${stat.linter}`,
      label: `Disable ${stat.linter} entirely`,
      title: 'Add {source} to .gavel.yaml',
      req: { source: stat.linter, work_dir: workDir },
      variant: 'subtle',
    });
  }
  return actions;
}

// uniqueChildLinters walks node.children and returns the set of unique
// linterName values for kind='linter' children. Used for multi-linter file
// parents built by buildFileTree (a 'lint-file' node whose children are 'linter'
// nodes).
function uniqueChildLinters(node: Test): string[] {
  const out: string[] = [];
  const seen = new Set<string>();
  for (const child of node.children || []) {
    if (child.kind !== 'linter') continue;
    const name = child.linterName || '';
    if (!name || seen.has(name)) continue;
    seen.add(name);
    out.push(name);
  }
  return out;
}

// lintFileActions returns the buttons for a lint-file node. Two shapes:
//   - single-linter file (leaf-ish, has linterName) → ignore-in-file + disable-entirely
//   - multi-linter file parent (kind=lint-file with linter children, no
//     linterName) → "everything in file" + per-linter pair
export function lintFileActions(node: Test): LintAction[] {
  const linter = node.linterName || '';
  const targetPath = node.target_path || node.file || '';
  const workDir = node.work_dir || '';

  if (linter) {
    return [
      {
        key: 'file-everything-linter',
        label: `Ignore all ${linter} in this file`,
        title: 'Add {source, file} to .gavel.yaml',
        req: { source: linter, file: targetPath, work_dir: workDir },
      },
      {
        key: 'linter-disable',
        label: `Disable ${linter} entirely`,
        title: 'Add {source} to .gavel.yaml',
        req: { source: linter, work_dir: workDir },
        variant: 'subtle',
      },
    ];
  }

  const actions: LintAction[] = [
    {
      key: 'file-everything',
      label: 'Ignore everything in this file',
      title: 'Add {file} to .gavel.yaml',
      req: { file: targetPath, work_dir: workDir },
    },
  ];
  for (const childLinter of uniqueChildLinters(node)) {
    actions.push({
      key: `file-linter:${childLinter}`,
      label: `Ignore all ${childLinter} in this file`,
      title: 'Add {source, file} to .gavel.yaml',
      req: { source: childLinter, file: targetPath, work_dir: workDir },
      variant: 'subtle',
    });
    actions.push({
      key: `linter-disable:${childLinter}`,
      label: `Disable ${childLinter} entirely`,
      title: 'Add {source} to .gavel.yaml',
      req: { source: childLinter, work_dir: workDir },
      variant: 'subtle',
    });
  }
  return actions;
}
