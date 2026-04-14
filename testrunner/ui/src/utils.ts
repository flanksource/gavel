import type { Test, LinterResult, Violation, Severity } from './types';

export interface LintFilters {
  severity: Set<Severity>;
  linter: Set<string>;
}

interface FlatViolation {
  linter: string;
  linterResult: LinterResult;
  v: Violation;
  file: string;
}

export function relPath(file: string | undefined, workDir: string | undefined): string {
  if (!file) return '(no file)';
  if (!workDir) return file;
  const prefix = workDir.endsWith('/') ? workDir : workDir + '/';
  if (file.startsWith(prefix)) return file.slice(prefix.length);
  return file;
}

function worstSeverity(vs: Violation[] | undefined): Severity {
  let worst: Severity = 'info';
  for (const v of vs || []) {
    const s = (v.severity || 'error') as Severity;
    if (s === 'error') return 'error';
    if (s === 'warning') worst = 'warning';
  }
  return worst;
}

function flattenLint(lint: LinterResult[] | undefined, filters: LintFilters): FlatViolation[] {
  const out: FlatViolation[] = [];
  for (const lr of lint || []) {
    if (filters.linter.size > 0 && !filters.linter.has(lr.linter)) continue;
    for (const v of lr.violations || []) {
      const sev = (v.severity || 'error') as Severity;
      if (filters.severity.size > 0 && !filters.severity.has(sev)) continue;
      const file = relPath(v.file, lr.work_dir);
      out.push({ linter: lr.linter, linterResult: lr, v: { ...v, file }, file });
    }
  }
  return out;
}

function folderNode(name: string, kind: Test['kind'], children: Test[]): Test {
  const failed = children.some(c => c.failed || (c.children?.length || 0) > 0);
  return {
    name,
    framework: 'lint',
    kind,
    children,
    failed,
    passed: !failed,
    skipped: false,
  };
}

function fileLeafNode(file: string, linterName: string, violations: Violation[]): Test {
  const sev = worstSeverity(violations);
  return {
    name: `${file} (${violations.length})`,
    framework: 'lint',
    kind: 'lint-file',
    file,
    linterName,
    violations: violations
      .slice()
      .sort((a, b) => (a.line || 0) - (b.line || 0)),
    failed: sev !== 'info',
    passed: sev === 'info',
    skipped: false,
  };
}

function ruleLeafNode(rule: string, linterName: string, file: string, violations: Violation[]): Test {
  const sev = worstSeverity(violations);
  return {
    name: `${rule} (${violations.length})`,
    framework: 'lint',
    kind: 'lint-rule',
    file,
    linterName,
    ruleName: rule,
    violations: violations
      .slice()
      .sort((a, b) => (a.line || 0) - (b.line || 0)),
    failed: sev !== 'info',
    passed: sev === 'info',
    skipped: false,
  };
}

export function groupLintByLinterFile(lint: LinterResult[] | undefined, filters: LintFilters): Test[] {
  const flat = flattenLint(lint, filters);
  const byLinter = new Map<string, Map<string, Violation[]>>();
  const linterMeta = new Map<string, LinterResult>();
  for (const f of flat) {
    linterMeta.set(f.linter, f.linterResult);
    let files = byLinter.get(f.linter);
    if (!files) {
      files = new Map();
      byLinter.set(f.linter, files);
    }
    let arr = files.get(f.file);
    if (!arr) {
      arr = [];
      files.set(f.file, arr);
    }
    arr.push(f.v);
  }

  const linterNames = Array.from(byLinter.keys()).sort();
  return linterNames.map(linter => {
    const files = byLinter.get(linter)!;
    const fileNames = Array.from(files.keys()).sort();
    const fileLeaves = fileNames.map(file => fileLeafNode(file, linter, files.get(file)!));
    const total = fileLeaves.reduce((n, f) => n + (f.violations?.length || 0), 0);
    const node = folderNode(`${linter} (${total})`, 'linter', fileLeaves);
    node.linter = linterMeta.get(linter);
    node.linterName = linter;
    return node;
  });
}

export function groupLintByFileLinterRule(lint: LinterResult[] | undefined, filters: LintFilters): Test[] {
  const flat = flattenLint(lint, filters);
  const byFile = new Map<string, Map<string, Map<string, Violation[]>>>();
  for (const f of flat) {
    let linters = byFile.get(f.file);
    if (!linters) {
      linters = new Map();
      byFile.set(f.file, linters);
    }
    let rules = linters.get(f.linter);
    if (!rules) {
      rules = new Map();
      linters.set(f.linter, rules);
    }
    const ruleKey = f.v.rule?.method || '(no rule)';
    let arr = rules.get(ruleKey);
    if (!arr) {
      arr = [];
      rules.set(ruleKey, arr);
    }
    arr.push(f.v);
  }

  const fileNames = Array.from(byFile.keys()).sort();
  return fileNames.map(file => {
    const linters = byFile.get(file)!;
    const linterNames = Array.from(linters.keys()).sort();
    const linterNodes = linterNames.map(linter => {
      const rules = linters.get(linter)!;
      const ruleNames = Array.from(rules.keys()).sort();
      const ruleLeaves = ruleNames.map(rule => ruleLeafNode(rule, linter, file, rules.get(rule)!));
      const total = ruleLeaves.reduce((n, r) => n + (r.violations?.length || 0), 0);
      const node = folderNode(`${linter} (${total})`, 'linter', ruleLeaves);
      node.linterName = linter;
      node.file = file;
      return node;
    });
    const total = linterNodes.reduce((n, l) => {
      return n + (l.children || []).reduce((m, r) => m + (r.violations?.length || 0), 0);
    }, 0);
    const node = folderNode(`${file} (${total})`, 'lint-root', linterNodes);
    node.file = file;
    return node;
  });
}

export function collectLintLinters(lint: LinterResult[] | undefined): string[] {
  const set = new Set<string>();
  for (const lr of lint || []) {
    if ((lr.violations || []).length > 0) set.add(lr.linter);
  }
  return Array.from(set).sort();
}

export function countLintBySeverity(
  lint: LinterResult[] | undefined,
  linterFilter: Set<string>,
): Record<Severity, number> {
  const counts: Record<Severity, number> = { error: 0, warning: 0, info: 0 };
  for (const lr of lint || []) {
    if (linterFilter.size > 0 && !linterFilter.has(lr.linter)) continue;
    for (const v of lr.violations || []) {
      const sev = (v.severity || 'error') as Severity;
      counts[sev] = (counts[sev] || 0) + 1;
    }
  }
  return counts;
}

export function countLintByLinter(
  lint: LinterResult[] | undefined,
  severityFilter: Set<Severity>,
): Record<string, number> {
  const counts: Record<string, number> = {};
  for (const lr of lint || []) {
    for (const v of lr.violations || []) {
      const sev = (v.severity || 'error') as Severity;
      if (severityFilter.size > 0 && !severityFilter.has(sev)) continue;
      counts[lr.linter] = (counts[lr.linter] || 0) + 1;
    }
  }
  return counts;
}

export function totalLintViolations(lint: LinterResult[] | undefined): number {
  let n = 0;
  for (const lr of lint || []) n += (lr.violations || []).length;
  return n;
}

export function statusIcon(t: Test): string {
  if (t.kind === 'lint-file') {
    return 'codicon:file';
  }
  if (t.kind === 'lint-rule') {
    return 'codicon:symbol-ruler';
  }
  if (t.kind === 'violation') {
    const sev = t.violation?.severity;
    if (sev === 'error') return 'codicon:error';
    if (sev === 'warning') return 'codicon:warning';
    return 'codicon:info';
  }
  if (t.pending) return 'svg-spinners:ring-resize';
  if (t.failed) return 'codicon:error';
  if (t.skipped) return 'codicon:circle-slash';
  if (t.passed) return 'codicon:pass-filled';
  if (t.children && t.children.length > 0) {
    const s = sum(t);
    if (s.failed > 0 && s.passed > 0) return 'codicon:warning';
    if (s.failed > 0) return 'codicon:error';
    if (s.pending > 0) return 'svg-spinners:ring-resize';
    return 'codicon:pass-filled';
  }
  return 'codicon:circle-outline';
}

export function statusColor(t: Test): string {
  if (t.kind === 'lint-file' || t.kind === 'lint-rule') {
    const sev = worstSeverity(t.violations);
    if (sev === 'error') return 'text-red-600';
    if (sev === 'warning') return 'text-yellow-600';
    return 'text-blue-500';
  }
  if (t.kind === 'violation') {
    const sev = t.violation?.severity;
    if (sev === 'error') return 'text-red-600';
    if (sev === 'warning') return 'text-yellow-600';
    return 'text-blue-500';
  }
  if (t.pending) return 'text-blue-500';
  if (t.failed) return 'text-red-600';
  if (t.skipped) return 'text-yellow-600';
  if (t.passed) return 'text-green-600';
  if (t.children && t.children.length > 0) {
    const s = sum(t);
    if (s.failed > 0 && s.passed > 0) return 'text-orange-500';
    if (s.failed > 0) return 'text-red-600';
    if (s.pending > 0) return 'text-blue-500';
    return 'text-green-600';
  }
  return 'text-gray-500';
}

export function statusBg(t: Test): string {
  if (t.failed) return 'bg-red-50';
  if (t.skipped) return 'bg-yellow-50';
  return '';
}

export function frameworkIcon(framework?: string): string | null {
  switch (framework) {
    case 'go test': return 'devicon:go';
    case 'ginkgo': return 'devicon:go';
    case 'fixture': return 'vscode-icons:file-type-markdown';
    case 'lint': return 'codicon:lightbulb';
    default: return null;
  }
}

export function formatDuration(ns: number): string {
  if (ns <= 0) return '';
  const ms = ns / 1e6;
  if (ms < 1000) return `${ms.toFixed(0)}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

export function sum(t: Test): { total: number; passed: number; failed: number; skipped: number; pending: number } {
  if (t.summary) {
    return { total: t.summary.Total, passed: t.summary.Passed, failed: t.summary.Failed, skipped: t.summary.Skipped, pending: t.summary.Pending || 0 };
  }
  if (!t.children || t.children.length === 0) {
    return {
      total: (t.passed || t.failed || t.skipped || t.pending) ? 1 : 0,
      passed: t.passed ? 1 : 0,
      failed: t.failed ? 1 : 0,
      skipped: t.skipped ? 1 : 0,
      pending: t.pending ? 1 : 0,
    };
  }
  const r = { total: 0, passed: 0, failed: 0, skipped: 0, pending: 0 };
  for (const c of t.children) {
    const s = sum(c);
    r.total += s.total;
    r.passed += s.passed;
    r.failed += s.failed;
    r.skipped += s.skipped;
    r.pending += s.pending;
  }
  return r;
}

export function isFolder(t: Test): boolean {
  return !t.passed && !t.failed && !t.skipped && !t.pending;
}

export function hasFailed(t: Test): boolean {
  if (t.failed) return true;
  return (t.children || []).some(hasFailed);
}

export function hasPending(t: Test): boolean {
  if (t.pending) return true;
  return (t.children || []).some(hasPending);
}

export function collectFrameworks(tests: Test[]): string[] {
  const set = new Set<string>();
  function walk(t: Test) {
    if (t.framework) set.add(t.framework);
    (t.children || []).forEach(walk);
  }
  tests.forEach(walk);
  return Array.from(set).sort();
}

export function testStatus(t: Test): string | null {
  if (t.kind === 'violation') return 'failed';
  if (t.pending) return 'pending';
  if (t.failed) return 'failed';
  if (t.skipped) return 'skipped';
  if (t.passed) return 'passed';
  return null;
}

export function filterTests(
  tests: Test[],
  statusFilter: Set<string>,
  frameworkFilter: Set<string>,
): Test[] {
  if (statusFilter.size === 0 && frameworkFilter.size === 0) return tests;
  return tests.map(t => filterNode(t, statusFilter, frameworkFilter)).filter(Boolean) as Test[];
}

function filterNode(t: Test, statusFilter: Set<string>, frameworkFilter: Set<string>): Test | null {
  const hasChildren = !!t.children && t.children.length > 0;

  if (hasChildren) {
    const filtered = t.children!.map(c => filterNode(c, statusFilter, frameworkFilter)).filter(Boolean) as Test[];
    if (filtered.length === 0) return null;
    return { ...t, children: filtered, summary: undefined };
  }

  return matchesLeaf(t, statusFilter, frameworkFilter) ? t : null;
}

function matchesLeaf(t: Test, statusFilter: Set<string>, frameworkFilter: Set<string>): boolean {
  const s = testStatus(t);
  if (s === null) return false;
  if (statusFilter.size > 0 && !statusFilter.has(s)) return false;
  if (frameworkFilter.size > 0) {
    if (!t.framework || !frameworkFilter.has(t.framework)) return false;
  }
  return true;
}

export function humanizeName(name: string, framework?: string): string {
  if (framework !== 'go test') return name;
  if (name.endsWith('/')) return name;
  const parts = name.split('/');
  return parts.map((p, i) => {
    if (i === 0) {
      // Strip Test prefix, insert spaces before capitals
      let s = p.replace(/^Test/, '');
      s = s.replace(/([a-z0-9])([A-Z])/g, '$1 $2');
      return s;
    }
    // Subtests: replace underscores with spaces
    return p.replace(/_/g, ' ');
  }).join(' / ');
}

export function totalDuration(t: Test): number {
  if (t.duration && t.duration > 0) return t.duration;
  let d = 0;
  for (const c of t.children || []) d += totalDuration(c);
  return d;
}
