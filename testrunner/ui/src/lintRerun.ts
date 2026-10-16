import type { LintNode, LinterResult } from '@flanksource/clicky-ui/data';
import type { RerunRequest } from './types';

function normalize(path: string): string {
  return path.replace(/\\/g, '/').replace(/\/+/g, '/').replace(/\/$/, '');
}

// relPath is file relative to the linter's work dir, the form lint_files uses.
function relPath(file: string | undefined, workDir: string | undefined): string | undefined {
  if (!file) return undefined;
  const normalized = normalize(file);
  if (!workDir) return normalized;
  const prefix = `${normalize(workDir)}/`;
  return normalized.startsWith(prefix) ? normalized.slice(prefix.length) : normalized;
}

function collectWorkDirs(node: LintNode, out: Set<string>) {
  if (node.work_dir) out.add(node.work_dir);
  (node.children ?? []).forEach(child => collectWorkDirs(child, out));
}

/**
 * The rerun request that re-lints what a lint tree node covers: its linter, the
 * files beneath it that still have findings, and the single work dir they share.
 * Without a node it reruns every linter. `work_dir` is empty when the scope
 * spans several lint roots, which the server cannot rerun as one.
 */
export function buildLintRerunBody(node: LintNode | null, lint: readonly LinterResult[]): RerunRequest {
  const workDirs = new Set<string>();
  const linters = new Set<string>();
  const files = new Set<string>();

  const addMatches = (targetPath: string | undefined, source?: string, rule?: string) => {
    for (const result of lint) {
      if (node?.work_dir && result.work_dir && result.work_dir !== node.work_dir) continue;
      if (source && result.linter !== source) continue;
      let matched = false;
      for (const violation of result.violations ?? []) {
        const rawFile = relPath(violation.file, result.work_dir);
        if (rule && (violation.rule?.method || '(no rule)') !== rule) continue;
        if (!rawFile) {
          matched = rule !== undefined && targetPath === undefined;
          continue;
        }
        const inScope = !targetPath || rawFile === targetPath || rawFile.startsWith(`${targetPath}/`);
        if (!inScope) continue;
        matched = true;
        files.add(rawFile);
      }
      if (matched) {
        if (result.work_dir) workDirs.add(result.work_dir);
        linters.add(result.linter);
      }
    }
  };

  if (!node) {
    for (const result of lint) {
      linters.add(result.linter);
      if (result.work_dir) workDirs.add(result.work_dir);
    }
  } else if (node.kind === 'linter' || node.kind === 'linter-error') {
    if (node.linterName) linters.add(node.linterName);
    collectWorkDirs(node, workDirs);
  } else if (node.kind === 'lint-rule' || node.kind === 'lint-rule-group') {
    if (node.linterName) linters.add(node.linterName);
    addMatches(node.target_path, node.linterName, node.ruleName);
  } else {
    addMatches(node.target_path, node.linterName, node.ruleName);
  }

  if (node?.work_dir) workDirs.add(node.work_dir);
  return {
    lint: true,
    work_dir: workDirs.size === 1 ? Array.from(workDirs)[0]! : '',
    lint_linters: Array.from(linters),
    lint_files: Array.from(files),
  };
}
