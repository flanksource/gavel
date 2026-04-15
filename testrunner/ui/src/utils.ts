import type { Test, LinterResult, Violation, Severity } from './types';

export interface LintFilters {
  severity: Set<Severity>;
  linter: Set<string>;
}

interface FlatViolation {
  linter: string;
  linterResult: LinterResult;
  v: Violation;
  file?: string;
}

interface RuleBucket {
  file?: string;
  violations: Violation[];
}

interface LinterBucket {
  files: Map<string, Violation[]>;
  noFileRules: Map<string, RuleBucket>;
}

interface FileTreeNode {
  name: string;
  path?: string;
  children: Map<string, FileTreeNode>;
  files: Map<string, Map<string, Map<string, Violation[]>>>;
}

export function relPath(file: string | undefined, workDir: string | undefined): string | undefined {
  if (!file) return undefined;
  const normalizedFile = normalizeLintPath(file);
  if (!workDir) return normalizedFile;
  const prefix = normalizeLintPath(workDir).replace(/\/?$/, '/');
  if (normalizedFile.startsWith(prefix)) return normalizedFile.slice(prefix.length);
  return normalizedFile;
}

function normalizeLintPath(path: string): string {
  return path.replace(/\\/g, '/').replace(/\/+/g, '/').replace(/\/$/, '');
}

function ruleKeyFor(v: Violation): string {
  return v.rule?.method || '(no rule)';
}

function ruleBuckets(violations: Violation[]): Map<string, RuleBucket> {
  const buckets = new Map<string, RuleBucket>();
  for (const v of violations) {
    const key = ruleKeyFor(v);
    let bucket = buckets.get(key);
    if (!bucket) {
      bucket = { file: v.file, violations: [] };
      buckets.set(key, bucket);
    }
    bucket.violations.push(v);
  }
  return buckets;
}

function collapsedPathSegments(path: string): string[] {
  const parts = normalizeLintPath(path).split('/').filter(Boolean);
  if (parts.length >= 3 && parts[0] === '.shell' && parts[1] === 'checkout') {
    return [`${parts[0]}/${parts[1]}/${parts[2]}`, ...parts.slice(3)];
  }
  return parts;
}

function fileTreeRoot(): FileTreeNode {
  return { name: '', children: new Map(), files: new Map() };
}

function insertFileNode(root: FileTreeNode, file: string, linter: string, violation: Violation): void {
  const segments = collapsedPathSegments(file);
  if (segments.length === 0) return;
  let current = root;
  for (const segment of segments.slice(0, -1)) {
    let child = current.children.get(segment);
    if (!child) {
      child = { name: segment, path: current.path ? `${current.path}/${segment}` : segment, children: new Map(), files: new Map() };
      current.children.set(segment, child);
    }
    current = child;
  }

  const basename = segments[segments.length - 1];
  let linters = current.files.get(basename);
  if (!linters) {
    linters = new Map();
    current.files.set(basename, linters);
  }
  let rules = linters.get(linter);
  if (!rules) {
    rules = new Map();
    linters.set(linter, rules);
  }
  const key = ruleKeyFor(violation);
  let arr = rules.get(key);
  if (!arr) {
    arr = [];
    rules.set(key, arr);
  }
  arr.push(violation);
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
  const failed = children.some(hasFailed);
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

function fileLeafNode(name: string, file: string, linterName: string, violations: Violation[]): Test {
  const sev = worstSeverity(violations);
  return {
    name: `${name} (${violations.length})`,
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

function linterNode(
  linterName: string,
  linterResult: LinterResult | undefined,
  children: Test[],
  noFileViolations?: Violation[],
): Test {
  const total = children.reduce((n, child) => n + lintViolationCount(child), 0) + (noFileViolations?.length || 0);
  const node = folderNode(`${linterName} (${total})`, 'linter', children);
  node.linter = linterResult;
  node.linterName = linterName;
  if (noFileViolations && noFileViolations.length > 0) {
    node.noFileViolations = noFileViolations.slice().sort((a, b) => (a.line || 0) - (b.line || 0));
  }
  return node;
}

function lintViolationCount(t: Test): number {
  if (t.violations) return t.violations.length;
  return (t.children || []).reduce((n, child) => n + lintViolationCount(child), 0);
}

function buildFileTree(root: FileTreeNode, linterMeta: Map<string, LinterResult>): Test[] {
  const folderNames = Array.from(root.children.keys()).sort();
  const fileNames = Array.from(root.files.keys()).sort();
  const children: Test[] = [];

  for (const folderName of folderNames) {
    const folder = root.children.get(folderName)!;
    children.push(buildFolderNode(folder, linterMeta));
  }

  for (const fileName of fileNames) {
    const linters = root.files.get(fileName)!;
    const linterNames = Array.from(linters.keys()).sort();
    const fullPath = root.path ? `${root.path}/${fileName}` : fileName;
    const linterNodes = linterNames.map(linterName => {
      const rules = linters.get(linterName)!;
      const ruleNames = Array.from(rules.keys()).sort();
      const ruleLeaves = ruleNames.map(ruleName => ruleLeafNode(ruleName, linterName, fullPath, rules.get(ruleName)!));
      const node = linterNode(linterName, linterMeta.get(linterName), ruleLeaves);
      node.file = fullPath;
      return node;
    });
    const fileTotal = linterNodes.reduce((n, node) => n + lintViolationCount(node), 0);
    const fileNode = folderNode(`${fileName} (${fileTotal})`, 'lint-file', linterNodes);
    fileNode.file = fullPath;
    children.push(fileNode);
  }

  return children;
}

function buildFolderNode(folder: FileTreeNode, linterMeta: Map<string, LinterResult>): Test {
  const children = buildFileTree(folder, linterMeta);
  return folderNode(`${folder.name} (${children.reduce((n, child) => n + lintViolationCount(child), 0)})`, 'lint-folder', children);
}

export function groupLintByLinterFile(lint: LinterResult[] | undefined, filters: LintFilters): Test[] {
  const flat = flattenLint(lint, filters);
  const byLinter = new Map<string, LinterBucket>();
  const linterMeta = new Map<string, LinterResult>();
  for (const f of flat) {
    linterMeta.set(f.linter, f.linterResult);
    let bucket = byLinter.get(f.linter);
    if (!bucket) {
      bucket = { files: new Map(), noFileRules: new Map() };
      byLinter.set(f.linter, bucket);
    }
    if (!f.file) {
      const key = ruleKeyFor(f.v);
      let rule = bucket.noFileRules.get(key);
      if (!rule) {
        rule = { violations: [] };
        bucket.noFileRules.set(key, rule);
      }
      rule.violations.push(f.v);
      continue;
    }
    let arr = bucket.files.get(f.file);
    if (!arr) {
      arr = [];
      bucket.files.set(f.file, arr);
    }
    arr.push(f.v);
  }

  const linterNames = Array.from(byLinter.keys()).sort();
  return linterNames.map(linter => {
    const bucket = byLinter.get(linter)!;
    const fileNames = Array.from(bucket.files.keys()).sort();
    const fileLeaves = fileNames.map(file => {
      return fileLeafNode(file, file, linter, bucket.files.get(file)!);
    });
    const noFileViolations = Array.from(bucket.noFileRules.values()).flatMap(rule => rule.violations);
    return linterNode(linter, linterMeta.get(linter), fileLeaves, noFileViolations);
  });
}

export function groupLintByFileLinterRule(lint: LinterResult[] | undefined, filters: LintFilters): Test[] {
  const flat = flattenLint(lint, filters);
  const root = fileTreeRoot();
  const byNoFileLinter = new Map<string, Map<string, RuleBucket>>();
  const linterMeta = new Map<string, LinterResult>();
  for (const f of flat) {
    linterMeta.set(f.linter, f.linterResult);
    if (!f.file) {
      let rules = byNoFileLinter.get(f.linter);
      if (!rules) {
        rules = new Map();
        byNoFileLinter.set(f.linter, rules);
      }
      const key = ruleKeyFor(f.v);
      let bucket = rules.get(key);
      if (!bucket) {
        bucket = { violations: [] };
        rules.set(key, bucket);
      }
      bucket.violations.push(f.v);
      continue;
    }
    insertFileNode(root, f.file, f.linter, f.v);
  }

  const fileTreeNodes = buildFileTree(root, linterMeta);
  const noFileLinterNodes = Array.from(byNoFileLinter.keys()).sort().map(linterName => {
    const rules = byNoFileLinter.get(linterName)!;
    const ruleNames = Array.from(rules.keys()).sort();
    const ruleLeaves = ruleNames.map(ruleName => ruleLeafNode(ruleName, linterName, '', rules.get(ruleName)!.violations));
    const node = linterNode(linterName, linterMeta.get(linterName), ruleLeaves);
    node.file = undefined;
    return node;
  });

  return [...fileTreeNodes, ...noFileLinterNodes];
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
  if (t.kind === 'lint-folder') {
    return 'codicon:folder';
  }
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
  if (t.kind === 'lint-folder') {
    return 'text-gray-500';
  }
  if (t.kind === 'lint-file' || t.kind === 'lint-rule') {
    const sev = worstSeverity(t.violations);
    if (t.children && t.children.length > 0) {
      const s = sum(t);
      if (s.failed > 0) return 'text-red-600';
      if (s.passed > 0) return 'text-green-600';
      return 'text-gray-500';
    }
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
    case 'task': return 'codicon:tools';
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

export function sumNonTaskTests(t: Test): { total: number; passed: number; failed: number; skipped: number; pending: number } {
  if (t.framework === 'task') {
    if (!t.children || t.children.length === 0) {
      return { total: 0, passed: 0, failed: 0, skipped: 0, pending: 0 };
    }
    const r = { total: 0, passed: 0, failed: 0, skipped: 0, pending: 0 };
    for (const c of t.children) {
      const s = sumNonTaskTests(c);
      r.total += s.total;
      r.passed += s.passed;
      r.failed += s.failed;
      r.skipped += s.skipped;
      r.pending += s.pending;
    }
    return r;
  }

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
    const s = sumNonTaskTests(c);
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
    if (t.framework && t.framework !== 'task') set.add(t.framework);
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
