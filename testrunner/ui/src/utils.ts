import type { Test, LinterResult, Violation, Severity, ProcessNode } from './types';
import type { FilterState } from './filterState';
import { matchesFilterState } from './filterState';

export interface LintFilters {
  severity: FilterState<Severity>;
  linter: FilterState<string>;
}

interface FlatViolation {
  linter: string;
  linterResult: LinterResult;
  v: Violation;
  file?: string;
  rawFile?: string;
  workDir?: string;
}

interface RuleBucket {
  file?: string;
  violations: Violation[];
}

interface FileBucket {
  rawFile?: string;
  workDir?: string;
  violations: Violation[];
}

interface LinterBucket {
  files: Map<string, FileBucket>;
  noFileRules: Map<string, RuleBucket>;
}

interface FileTreeNode {
  name: string;
  path?: string;
  rawPath?: string;
  children: Map<string, FileTreeNode>;
  files: Map<string, Map<string, Map<string, Violation[]>>>;
}

interface RuleScopedBucket {
  files: Map<string, FileBucket>;
  noFileViolations: Violation[];
}

interface RuleFileTreeNode {
  name: string;
  path?: string;
  rawPath?: string;
  children: Map<string, RuleFileTreeNode>;
  files: Map<string, FileBucket>;
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
  return { name: '', rawPath: '', children: new Map(), files: new Map() };
}

function ruleFileTreeRoot(): RuleFileTreeNode {
  return { name: '', rawPath: '', children: new Map(), files: new Map() };
}

function insertFileNode(root: FileTreeNode, file: string, rawFile: string, linter: string, violation: Violation): void {
  const segments = collapsedPathSegments(file);
  const rawSegments = collapsedPathSegments(rawFile);
  if (segments.length === 0 || rawSegments.length === 0) return;
  let current = root;
  const folderSegments = segments.slice(0, -1);
  const rawFolderSegments = rawSegments.slice(0, -1);
  const rawOffset = folderSegments.length - rawFolderSegments.length;
  for (let i = 0; i < folderSegments.length; i += 1) {
    const segment = folderSegments[i];
    let child = current.children.get(segment);
    if (!child) {
      const rawCount = Math.max(0, Math.min(rawFolderSegments.length, i + 1 - rawOffset));
      child = {
        name: segment,
        path: current.path ? `${current.path}/${segment}` : segment,
        rawPath: rawCount > 0 ? rawFolderSegments.slice(0, rawCount).join('/') : '',
        children: new Map(),
        files: new Map(),
      };
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

function insertRuleFileNode(root: RuleFileTreeNode, file: string, rawFile: string, workDir: string | undefined, violation: Violation): void {
  const segments = collapsedPathSegments(file);
  const rawSegments = collapsedPathSegments(rawFile);
  if (segments.length === 0 || rawSegments.length === 0) return;
  let current = root;
  const folderSegments = segments.slice(0, -1);
  const rawFolderSegments = rawSegments.slice(0, -1);
  const rawOffset = folderSegments.length - rawFolderSegments.length;
  for (let i = 0; i < folderSegments.length; i += 1) {
    const segment = folderSegments[i];
    let child = current.children.get(segment);
    if (!child) {
      const rawCount = Math.max(0, Math.min(rawFolderSegments.length, i + 1 - rawOffset));
      child = {
        name: segment,
        path: current.path ? `${current.path}/${segment}` : segment,
        rawPath: rawCount > 0 ? rawFolderSegments.slice(0, rawCount).join('/') : '',
        children: new Map(),
        files: new Map(),
      };
      current.children.set(segment, child);
    }
    current = child;
  }

  const basename = segments[segments.length - 1];
  let bucket = current.files.get(basename);
  if (!bucket) {
    bucket = { rawFile, workDir, violations: [] };
    current.files.set(basename, bucket);
  }
  bucket.violations.push(violation);
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
  const workDirs = new Set<string>();
  for (const lr of lint || []) {
    if (!matchesFilterState(lr.linter, filters.linter)) continue;
    if (lr.work_dir) workDirs.add(normalizeLintPath(lr.work_dir));
  }
  const multiRoot = workDirs.size > 1;
  for (const lr of lint || []) {
    if (!matchesFilterState(lr.linter, filters.linter)) continue;
    for (const v of lr.violations || []) {
      const sev = (v.severity || 'error') as Severity;
      if (!matchesFilterState(sev, filters.severity)) continue;
      const rawFile = relPath(v.file, lr.work_dir);
      const file = decorateLintFile(rawFile, lr.work_dir, multiRoot);
      out.push({
        linter: lr.linter,
        linterResult: lr,
        v: { ...v, raw_file: rawFile, file },
        file,
        rawFile,
        workDir: lr.work_dir,
      });
    }
  }
  return out;
}

export function decorateLintFile(file: string | undefined, workDir: string | undefined, multiRoot: boolean): string | undefined {
  if (!file) return file;
  if (!multiRoot || !workDir) return file;
  return `${workDirLabel(workDir)}/${file}`;
}

export function workDirLabel(workDir: string): string {
  const parts = normalizeLintPath(workDir).split('/').filter(Boolean);
  if (parts.length >= 2) return `${parts[parts.length - 2]}/${parts[parts.length - 1]}`;
  if (parts.length === 1) return parts[0];
  return normalizeLintPath(workDir);
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

function fileLeafNode(name: string, file: string, rawFile: string, linterName: string, violations: Violation[], workDir?: string): Test {
  const sev = worstSeverity(violations);
  return {
    name,
    framework: 'lint',
    kind: 'lint-file',
    file,
    target_path: rawFile,
    work_dir: workDir,
    linterName,
    violations: violations
      .slice()
      .sort((a, b) => (a.line || 0) - (b.line || 0)),
    failed: sev !== 'info',
    passed: sev === 'info',
    skipped: false,
  };
}

function ruleLeafNode(rule: string, linterName: string, file: string, rawFile: string, violations: Violation[], workDir?: string): Test {
  const sev = worstSeverity(violations);
  return {
    name: rule,
    framework: 'lint',
    kind: 'lint-rule',
    file,
    target_path: rawFile,
    work_dir: workDir,
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

function ruleGroupNode(rule: string, linterName: string, children: Test[], workDir?: string, noFileViolations?: Violation[]): Test {
  const node = folderNode(rule, 'lint-rule-group', children);
  node.linterName = linterName;
  node.ruleName = rule;
  node.work_dir = workDir;
  if (noFileViolations && noFileViolations.length > 0) {
    node.noFileViolations = noFileViolations.slice().sort((a, b) => (a.line || 0) - (b.line || 0));
    node.failed = worstSeverity(noFileViolations) !== 'info' || node.failed;
    node.passed = !node.failed;
  }
  return node;
}

function linterNode(
  linterName: string,
  linterResult: LinterResult | undefined,
  children: Test[],
  noFileViolations?: Violation[],
): Test {
  const node = folderNode(linterName, 'linter', children);
  node.linter = linterResult;
  node.linterName = linterName;
  if (noFileViolations && noFileViolations.length > 0) {
    node.noFileViolations = noFileViolations.slice().sort((a, b) => (a.line || 0) - (b.line || 0));
    node.failed = worstSeverity(noFileViolations) !== 'info' || node.failed;
    node.passed = !node.failed;
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
      const ruleLeaves = ruleNames.map(ruleName => {
        const violations = rules.get(ruleName)!;
        const rawFile = violations[0]?.raw_file || fullPath;
        return ruleLeafNode(ruleName, linterName, fullPath, rawFile, violations);
      });
      const node = linterNode(linterName, linterMeta.get(linterName), ruleLeaves);
      node.file = fullPath;
      node.target_path = ruleLeaves[0]?.target_path;
      return node;
    });
    const fileNode = folderNode(fileName, 'lint-file', linterNodes);
    fileNode.file = fullPath;
    fileNode.target_path = linterNodes[0]?.target_path;
    children.push(fileNode);
  }

  return children;
}

function buildFolderNode(folder: FileTreeNode, linterMeta: Map<string, LinterResult>): Test {
  const children = buildFileTree(folder, linterMeta);
  const node = folderNode(folder.name, 'lint-folder', children);
  node.file = folder.path;
  node.target_path = folder.rawPath;
  return node;
}

export function groupLintByLinterRuleFile(lint: LinterResult[] | undefined, filters: LintFilters): Test[] {
  const flat = flattenLint(lint, filters);
  const byLinter = new Map<string, Map<string, RuleScopedBucket>>();
  const linterMeta = new Map<string, LinterResult>();
  for (const f of flat) {
    linterMeta.set(f.linter, f.linterResult);
    let rules = byLinter.get(f.linter);
    if (!rules) {
      rules = new Map();
      byLinter.set(f.linter, rules);
    }
    const ruleName = ruleKeyFor(f.v);
    let bucket = rules.get(ruleName);
    if (!bucket) {
      bucket = { files: new Map(), noFileViolations: [] };
      rules.set(ruleName, bucket);
    }
    if (!f.file) {
      bucket.noFileViolations.push(f.v);
      continue;
    }
    let fileBucket = bucket.files.get(f.file);
    if (!fileBucket) {
      fileBucket = { rawFile: f.rawFile, workDir: f.workDir, violations: [] };
      bucket.files.set(f.file, fileBucket);
    }
    fileBucket.violations.push(f.v);
  }

  const linterNames = Array.from(byLinter.keys()).sort();
  return linterNames.map(linterName => {
    const rules = byLinter.get(linterName)!;
    const ruleNames = Array.from(rules.keys()).sort();
    const ruleNodes = ruleNames.map(ruleName => buildRuleGroupNode(
      ruleName,
      linterName,
      rules.get(ruleName)!,
      linterMeta.get(linterName),
    ));
    return linterNode(linterName, linterMeta.get(linterName), ruleNodes);
  });
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
    let fileBucket = bucket.files.get(f.file);
    if (!fileBucket) {
      fileBucket = { rawFile: f.rawFile, workDir: f.workDir, violations: [] };
      bucket.files.set(f.file, fileBucket);
    }
    fileBucket.violations.push(f.v);
  }

  const linterNames = Array.from(byLinter.keys()).sort();
  return linterNames.map(linter => {
    const bucket = byLinter.get(linter)!;
    const root = fileTreeRoot();
    for (const [file, fileBucket] of bucket.files.entries()) {
      for (const violation of fileBucket.violations) {
        insertFileNode(root, file, fileBucket.rawFile || file, linter, violation);
      }
    }
    const fileLeaves = buildLinterFileTree(root, linter, bucket);
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
    insertFileNode(root, f.file, f.rawFile || f.file, f.linter, f.v);
  }

  const fileTreeNodes = buildFileTree(root, linterMeta);
  const noFileLinterNodes = Array.from(byNoFileLinter.keys()).sort().map(linterName => {
    const rules = byNoFileLinter.get(linterName)!;
    const ruleNames = Array.from(rules.keys()).sort();
    const ruleLeaves = ruleNames.map(ruleName => {
      const violations = rules.get(ruleName)!.violations;
      return ruleLeafNode(ruleName, linterName, '', violations[0]?.raw_file || '', violations, linterMeta.get(linterName)?.work_dir);
    });
    const node = linterNode(linterName, linterMeta.get(linterName), ruleLeaves);
    node.file = undefined;
    return node;
  });

  return [...fileTreeNodes, ...noFileLinterNodes];
}

const SUMMARY_FILE_LIMIT = 5;

// groupLintBySummary collapses raw lint results into linter -> rule -> per-file
// nodes, mirroring the CLI `--summary` view. Skipped linters and linters that
// errored (even with zero violations) are shown as leaves so failures aren't
// silently hidden.
export function groupLintBySummary(lint: LinterResult[] | undefined, filters: LintFilters): Test[] {
  if (!lint || lint.length === 0) return [];

  const byLinter = new Map<string, {
    meta: LinterResult;
    violations: FlatViolation[];
    skipReason?: string;
    errorMsg?: string;
  }>();

  const workDirs = new Set<string>();
  for (const lr of lint) {
    if (lr.work_dir) workDirs.add(normalizeLintPath(lr.work_dir));
  }
  const multiRoot = workDirs.size > 1;

  for (const lr of lint) {
    if (!matchesFilterState(lr.linter, filters.linter)) continue;
    let bucket = byLinter.get(lr.linter);
    if (!bucket) {
      bucket = { meta: lr, violations: [] };
      byLinter.set(lr.linter, bucket);
    }
    if (lr.skipped) {
      bucket.skipReason = lr.error || 'skipped';
      continue;
    }
    if (!lr.success && lr.error && !bucket.errorMsg) {
      bucket.errorMsg = lr.error;
    }
    for (const v of lr.violations || []) {
      const sev = (v.severity || 'error') as Severity;
      if (!matchesFilterState(sev, filters.severity)) continue;
      const rawFile = relPath(v.file, lr.work_dir);
      const file = decorateLintFile(rawFile, lr.work_dir, multiRoot);
      bucket.violations.push({
        linter: lr.linter,
        linterResult: lr,
        v: { ...v, raw_file: rawFile, file },
        file,
        rawFile,
        workDir: lr.work_dir,
      });
    }
  }

  const names = Array.from(byLinter.keys()).sort();
  return names.map(name => {
    const b = byLinter.get(name)!;
    return summaryLinterNode(name, b.meta, b.violations, b.skipReason, b.errorMsg);
  });
}

function summaryLinterNode(
  linterName: string,
  meta: LinterResult,
  violations: FlatViolation[],
  skipReason: string | undefined,
  errorMsg: string | undefined,
): Test {
  // Skipped (no error, no violations).
  if (skipReason && violations.length === 0 && !errorMsg) {
    return {
      name: linterName,
      framework: 'lint',
      kind: 'linter',
      linter: meta,
      linterName,
      skipped: true,
      passed: false,
      failed: false,
      message: `skipped: ${skipReason}`,
    };
  }

  // Group violations by rule, then per-file.
  const byRule = new Map<string, FlatViolation[]>();
  for (const fv of violations) {
    const key = ruleKeyFor(fv.v);
    let list = byRule.get(key);
    if (!list) {
      list = [];
      byRule.set(key, list);
    }
    list.push(fv);
  }
  const ruleOrder = Array.from(byRule.keys()).sort((a, b) => {
    const diff = byRule.get(b)!.length - byRule.get(a)!.length;
    return diff !== 0 ? diff : a.localeCompare(b);
  });
  const ruleChildren = ruleOrder.map(rule => summaryRuleNode(rule, linterName, byRule.get(rule)!));

  const children: Test[] = [];
  if (errorMsg) {
    children.push(summaryErrorNode(linterName, errorMsg));
  }
  children.push(...ruleChildren);

  const failed = Boolean(errorMsg) || ruleChildren.some(r => r.failed);
  const passed = !failed && violations.length === 0;
  return {
    name: linterName,
    framework: 'lint',
    kind: 'linter',
    linter: meta,
    linterName,
    children,
    failed,
    passed,
    skipped: false,
  };
}

function summaryRuleNode(rule: string, linterName: string, flats: FlatViolation[]): Test {
  // Aggregate per file.
  const byFile = new Map<string, { file?: string; rawFile?: string; workDir?: string; violations: Violation[] }>();
  const fileOrder: string[] = [];
  for (const f of flats) {
    const key = f.file ?? '';
    let bucket = byFile.get(key);
    if (!bucket) {
      bucket = { file: f.file, rawFile: f.rawFile, workDir: f.workDir, violations: [] };
      byFile.set(key, bucket);
      fileOrder.push(key);
    }
    bucket.violations.push(f.v);
  }

  const shownKeys = fileOrder.slice(0, SUMMARY_FILE_LIMIT);
  const children: Test[] = shownKeys.map(key => {
    const b = byFile.get(key)!;
    const displayName = summaryFileLabel(b.file, b.violations);
    return {
      name: displayName,
      framework: 'lint',
      kind: 'lint-file',
      file: b.file,
      target_path: b.rawFile,
      work_dir: b.workDir,
      linterName,
      ruleName: rule,
      violations: b.violations.slice().sort((a, b) => (a.line || 0) - (b.line || 0)),
      failed: worstSeverity(b.violations) !== 'info',
      passed: worstSeverity(b.violations) === 'info',
      skipped: false,
    };
  });
  const remaining = fileOrder.length - shownKeys.length;
  if (remaining > 0) {
    children.push({
      name: `… ${remaining} more file${remaining === 1 ? '' : 's'}`,
      framework: 'lint',
      kind: 'lint-file',
      linterName,
      ruleName: rule,
      passed: true,
      failed: false,
      skipped: false,
    });
  }

  const firstMessage = flats.find(f => f.v.message)?.v.message || '';
  const displayName = firstMessage
    ? `${rule} (${flats.length}) — ${firstMessage}`
    : `${rule} (${flats.length})`;
  return {
    name: displayName,
    framework: 'lint',
    kind: 'lint-rule-group',
    linterName,
    ruleName: rule,
    children,
    failed: children.some(c => c.failed),
    passed: children.every(c => c.passed),
    skipped: false,
  };
}

function summaryFileLabel(file: string | undefined, violations: Violation[]): string {
  if (!file) return '(no file)';
  if (violations.length > 1) return `${file} (${violations.length})`;
  const v = violations[0];
  if (v && v.line) {
    return v.column ? `${file}:${v.line}:${v.column}` : `${file}:${v.line}`;
  }
  return file;
}

function summaryErrorNode(linterName: string, message: string): Test {
  return {
    name: `❌ ${firstLine(message)}`,
    framework: 'lint',
    kind: 'violation',
    linterName,
    failed: true,
    passed: false,
    skipped: false,
    message,
  };
}

function firstLine(s: string): string {
  const idx = s.indexOf('\n');
  return idx === -1 ? s : s.slice(0, idx);
}

function buildLinterFileTree(root: FileTreeNode, linterName: string, bucket: LinterBucket): Test[] {
  const folderNames = Array.from(root.children.keys()).sort();
  const fileNames = Array.from(root.files.keys()).sort();
  const children: Test[] = [];

  for (const folderName of folderNames) {
    const folder = root.children.get(folderName)!;
    children.push(buildLinterFolderNode(folder, linterName, bucket));
  }

  for (const fileName of fileNames) {
    const linters = root.files.get(fileName);
    const violations = linters?.get(linterName);
    if (!violations) continue;
    const fullPath = root.path ? `${root.path}/${fileName}` : fileName;
    const fileViolations = Array.from(violations.values()).flatMap(group => group);
    const rawFile = bucket.files.get(fullPath)?.rawFile || fileViolations[0]?.raw_file || fullPath;
    children.push(fileLeafNode(fileName, fullPath, rawFile, linterName, fileViolations, bucket.files.get(fullPath)?.workDir));
  }

  return children;
}

function buildLinterFolderNode(folder: FileTreeNode, linterName: string, bucket: LinterBucket): Test {
  const children = buildLinterFileTree(folder, linterName, bucket);
  const node = folderNode(folder.name, 'lint-folder', children);
  node.file = folder.path;
  node.target_path = folder.rawPath;
  node.work_dir = uniqueWorkDir(children);
  return node;
}

function buildRuleGroupNode(ruleName: string, linterName: string, bucket: RuleScopedBucket, linterResult: LinterResult | undefined): Test {
  const root = ruleFileTreeRoot();
  for (const [file, fileBucket] of bucket.files.entries()) {
    for (const violation of fileBucket.violations) {
      insertRuleFileNode(root, file, fileBucket.rawFile || file, fileBucket.workDir, violation);
    }
  }
  const children = buildRuleFileTree(root, linterName, ruleName);
  return ruleGroupNode(ruleName, linterName, children, uniqueWorkDir(children) || linterResult?.work_dir, bucket.noFileViolations);
}

function buildRuleFileTree(root: RuleFileTreeNode, linterName: string, ruleName: string): Test[] {
  const folderNames = Array.from(root.children.keys()).sort();
  const fileNames = Array.from(root.files.keys()).sort();
  const children: Test[] = [];

  for (const folderName of folderNames) {
    const folder = root.children.get(folderName)!;
    children.push(buildRuleFolderNode(folder, linterName, ruleName));
  }

  for (const fileName of fileNames) {
    const fileBucket = root.files.get(fileName)!;
    const fullPath = root.path ? `${root.path}/${fileName}` : fileName;
    const rawFile = fileBucket.rawFile || fileBucket.violations[0]?.raw_file || fullPath;
    const node = fileLeafNode(fileName, fullPath, rawFile, linterName, fileBucket.violations, fileBucket.workDir);
    node.ruleName = ruleName;
    children.push(node);
  }

  return children;
}

function buildRuleFolderNode(folder: RuleFileTreeNode, linterName: string, ruleName: string): Test {
  const children = buildRuleFileTree(folder, linterName, ruleName);
  const node = folderNode(folder.name, 'lint-folder', children);
  node.file = folder.path;
  node.target_path = folder.rawPath;
  node.linterName = linterName;
  node.ruleName = ruleName;
  node.work_dir = uniqueWorkDir(children);
  return node;
}

function uniqueWorkDir(children: Test[]): string | undefined {
  const dirs = new Set<string>();
  for (const child of children) {
    if (child.work_dir) dirs.add(child.work_dir);
  }
  if (dirs.size === 1) return Array.from(dirs)[0];
  return undefined;
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
  linterFilter: FilterState<string>,
): Record<Severity, number> {
  const counts: Record<Severity, number> = { error: 0, warning: 0, info: 0 };
  for (const lr of lint || []) {
    if (!matchesFilterState(lr.linter, linterFilter)) continue;
    for (const v of lr.violations || []) {
      const sev = (v.severity || 'error') as Severity;
      counts[sev] = (counts[sev] || 0) + 1;
    }
  }
  return counts;
}

export function countLintByLinter(
  lint: LinterResult[] | undefined,
  severityFilter: FilterState<Severity>,
): Record<string, number> {
  const counts: Record<string, number> = {};
  for (const lr of lint || []) {
    for (const v of lr.violations || []) {
      const sev = (v.severity || 'error') as Severity;
      if (!matchesFilterState(sev, severityFilter)) continue;
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

export function isLintNode(t: Test): boolean {
  return t.kind === 'lint-root'
    || t.kind === 'lint-folder'
    || t.kind === 'linter'
    || t.kind === 'violation'
    || t.kind === 'lint-file'
    || t.kind === 'lint-rule'
    || t.kind === 'lint-rule-group';
}

export function lintBadgeColor(t: Test): string {
  const sev = lintNodeSeverity(t);
  if (sev === 'error') return 'bg-red-500';
  if (sev === 'warning') return 'bg-amber-500';
  return 'bg-blue-500';
}

function lintNodeSeverity(t: Test): Severity {
  if (t.kind === 'violation') {
    return (t.violation?.severity || 'error') as Severity;
  }
  if (t.violations && t.violations.length > 0) {
    return worstSeverity(t.violations);
  }
  if (t.noFileViolations && t.noFileViolations.length > 0) {
    const own = worstSeverity(t.noFileViolations);
    if (own === 'error') return own;
    const child = worstSeverityInChildren(t.children);
    return severityRank(child) > severityRank(own) ? child : own;
  }
  return worstSeverityInChildren(t.children);
}

function worstSeverityInChildren(children: Test[] | undefined): Severity {
  let worst: Severity = 'info';
  for (const child of children || []) {
    const sev = lintNodeSeverity(child);
    if (sev === 'error') return 'error';
    if (sev === 'warning') worst = 'warning';
  }
  return worst;
}

function severityRank(sev: Severity): number {
  if (sev === 'error') return 3;
  if (sev === 'warning') return 2;
  return 1;
}

export function lintNodeCount(t: Test): number {
  if (t.kind === 'violation') return 0;
  if (t.violations) return t.violations.length;
  return (t.noFileViolations?.length || 0) + (t.children || []).reduce((n, child) => n + lintNodeCount(child), 0);
}

function taskStatus(t: Test): string {
  if (t.framework !== 'task' || !t.context || typeof t.context !== 'object') return '';
  const value = (t.context as Record<string, unknown>).status;
  return typeof value === 'string' ? value.toLowerCase() : '';
}

export function statusIcon(t: Test): string {
  if (taskStatus(t) === 'canceled') return 'codicon:debug-stop';
  if (t.kind === 'lint-folder') {
    return 'codicon:folder';
  }
  if (t.kind === 'lint-file') {
    return fileTypeIcon(t.file || t.name);
  }
  if (t.kind === 'linter') return lintToolIcon(t.linterName || t.linter?.linter);
  if (t.kind === 'lint-rule' || t.kind === 'lint-rule-group') {
    return 'codicon:symbol-ruler';
  }
  if (t.kind === 'violation') {
    const sev = t.violation?.severity;
    if (sev === 'error') return 'codicon:error';
    if (sev === 'warning') return 'codicon:warning';
    return 'codicon:circle-small';
  }
  if (t.timed_out) return 'ion:hourglass-outline';
  if (t.pending) return 'svg-spinners:ring-resize';
  if (t.failed) return 'codicon:error';
  if (t.skipped) return 'codicon:circle-slash';
  if (t.passed) return 'codicon:pass-filled';
  if (t.children && t.children.length > 0) {
    const s = sum(t);
    if (hasTimedOutDescendant(t)) return 'ion:hourglass-outline';
    if (s.failed > 0 && s.passed > 0) return 'codicon:warning';
    if (s.failed > 0) return 'codicon:error';
    if (s.pending > 0) return 'svg-spinners:ring-resize';
    return 'codicon:pass-filled';
  }
  return 'codicon:circle-outline';
}

function hasTimedOutDescendant(t: Test): boolean {
  if (t.timed_out) return true;
  if (!t.children) return false;
  for (const c of t.children) {
    if (hasTimedOutDescendant(c)) return true;
  }
  return false;
}

export function statusColor(t: Test): string {
  if (taskStatus(t) === 'canceled') return 'text-orange-600';
  if (t.kind === 'lint-folder') {
    return 'text-gray-500';
  }
  if (t.kind === 'lint-file' || t.kind === 'linter' || t.kind === 'lint-rule' || t.kind === 'lint-rule-group') {
    return 'text-gray-500';
  }
  if (t.kind === 'violation') {
    const sev = t.violation?.severity;
    if (sev === 'error') return 'text-red-600';
    if (sev === 'warning') return 'text-yellow-600';
    return 'text-gray-500';
  }
  if (t.timed_out) return 'text-amber-600';
  if (t.pending) return 'text-blue-500';
  if (t.failed) return 'text-red-600';
  if (t.skipped) return 'text-yellow-600';
  if (t.passed) return 'text-green-600';
  if (t.children && t.children.length > 0) {
    const s = sum(t);
    if (hasTimedOutDescendant(t)) return 'text-amber-600';
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
    case 'jest': return 'devicon:jest';
    case 'vitest': return 'logos:vitest';
    case 'playwright': return 'devicon:playwright';
    case 'fixture': return 'vscode-icons:file-type-markdown';
    case 'lint': return 'codicon:lightbulb';
    case 'task': return 'codicon:tools';
    default: return null;
  }
}

export function fileTypeIcon(file?: string): string {
  const normalized = (file || '').toLowerCase();
  if (normalized.endsWith('.go')) return 'devicon:go';
  if (normalized.endsWith('.ts')) return 'devicon:typescript';
  if (normalized.endsWith('.tsx')) return 'devicon:typescript';
  if (normalized.endsWith('.js')) return 'devicon:javascript';
  if (normalized.endsWith('.jsx')) return 'devicon:javascript';
  if (normalized.endsWith('.json')) return 'devicon:json';
  if (normalized.endsWith('.yml') || normalized.endsWith('.yaml')) return 'devicon:yaml';
  if (normalized.endsWith('.md')) return 'devicon:markdown';
  if (normalized.endsWith('.py')) return 'devicon:python';
  if (normalized.endsWith('.sh')) return 'devicon:bash';
  if (normalized.endsWith('.html')) return 'devicon:html5';
  if (normalized.endsWith('.css')) return 'devicon:css3';
  if (normalized.endsWith('.xml')) return 'devicon:xml';
  return 'vscode-icons:default-file';
}

export function lintToolIcon(linter?: string): string {
  const value = (linter || '').toLowerCase();
  if (value.includes('eslint')) return 'devicon:eslint';
  if (value.includes('markdownlint')) return 'devicon:markdown';
  if (value.includes('golangci') || value.includes('gosec') || value.includes('govet') || value.includes('gofmt') || value.includes('betterleaks')) return 'devicon:go';
  if (value.includes('pyright') || value.includes('ruff')) return 'devicon:python';
  if (value.includes('vale')) return 'devicon:markdown';
  if (value.includes('jscpd')) return 'devicon:javascript';
  return 'codicon:tools';
}

export function formatDuration(ns: number): string {
  if (ns <= 0) return '';
  const ms = ns / 1e6;
  if (ms < 1000) return `${ms.toFixed(0)}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

// formatRunTimestamp renders an RFC3339 timestamp with an explicit year so
// the UI no longer hides the year via toLocaleString() defaults.
export function formatRunTimestamp(iso: string | undefined): string {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const pad = (n: number) => n.toString().padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

// formatRunDuration reports a human duration between two RFC3339 timestamps.
// Returns a live "… so far" tick when ended is undefined.
export function formatRunDuration(started: string | undefined, ended: string | undefined): string {
  if (!started) return '—';
  const start = Date.parse(started);
  if (Number.isNaN(start)) return '—';
  const end = ended ? Date.parse(ended) : Date.now();
  if (Number.isNaN(end) || end < start) return '—';
  const ms = end - start;
  const sec = Math.floor(ms / 1000);
  if (sec < 60) return `${sec}s${ended ? '' : ' so far'}`;
  const min = Math.floor(sec / 60);
  const rem = sec % 60;
  if (min < 60) return `${min}m ${rem}s${ended ? '' : ' so far'}`;
  const hr = Math.floor(min / 60);
  const remMin = min % 60;
  return `${hr}h ${remMin}m${ended ? '' : ' so far'}`;
}

// hasTimeoutArgs reports whether SnapshotMetadata.Args carries any of the
// three timeout knobs so the DetailPanel can conditionally render the row.
export function hasTimeoutArgs(args: Record<string, unknown> | undefined): boolean {
  if (!args) return false;
  return !!(
    timeoutArgValue(args, 'timeout') ||
    timeoutArgValue(args, 'test_timeout') ||
    timeoutArgValue(args, 'lint_timeout')
  );
}

export function timeoutArgValue(args: Record<string, unknown> | undefined, key: string): string | null {
  if (!args) return null;
  const raw = args[key];
  if (typeof raw !== 'string') return null;
  const trimmed = raw.trim();
  return trimmed === '' ? null : trimmed;
}

// formatCount renders integer counts compactly so large totals don't crowd
// row layouts. <1000 is shown raw; 1000-999999 uses k with one decimal below
// 10k; 1000000+ uses M the same way.
export function formatCount(n: number): string {
  const abs = Math.abs(n);
  if (abs < 1000) return String(n);
  if (abs < 1_000_000) return compactWithUnit(n / 1000, 'k');
  return compactWithUnit(n / 1_000_000, 'M');
}

function compactWithUnit(value: number, unit: string): string {
  const rounded = Math.abs(value) >= 10 ? Math.round(value).toString() : value.toFixed(1).replace(/\.0$/, '');
  return `${rounded}${unit}`;
}

export function sum(t: Test): { total: number; passed: number; failed: number; skipped: number; pending: number; timedout: number } {
  if (t.summary) {
    return { total: t.summary.Total, passed: t.summary.Passed, failed: t.summary.Failed, skipped: t.summary.Skipped, pending: t.summary.Pending || 0, timedout: 0 };
  }
  if (!t.children || t.children.length === 0) {
    const isTimedOut = !!t.timed_out;
    const counted = isTimedOut || t.passed || t.failed || t.skipped || t.pending;
    return {
      total: counted ? 1 : 0,
      passed: !isTimedOut && t.passed ? 1 : 0,
      failed: !isTimedOut && t.failed ? 1 : 0,
      skipped: !isTimedOut && t.skipped ? 1 : 0,
      pending: !isTimedOut && t.pending ? 1 : 0,
      timedout: isTimedOut ? 1 : 0,
    };
  }
  const r = { total: 0, passed: 0, failed: 0, skipped: 0, pending: 0, timedout: 0 };
  for (const c of t.children) {
    const s = sum(c);
    r.total += s.total;
    r.passed += s.passed;
    r.failed += s.failed;
    r.skipped += s.skipped;
    r.pending += s.pending;
    r.timedout += s.timedout;
  }
  return r;
}

export function sumNonTaskTests(t: Test): { total: number; passed: number; failed: number; skipped: number; pending: number; timedout: number } {
  if (t.framework === 'task') {
    if (!t.children || t.children.length === 0) {
      return { total: 0, passed: 0, failed: 0, skipped: 0, pending: 0, timedout: 0 };
    }
    const r = { total: 0, passed: 0, failed: 0, skipped: 0, pending: 0, timedout: 0 };
    for (const c of t.children) {
      const s = sumNonTaskTests(c);
      r.total += s.total;
      r.passed += s.passed;
      r.failed += s.failed;
      r.skipped += s.skipped;
      r.pending += s.pending;
      r.timedout += s.timedout;
    }
    return r;
  }

  if (t.summary) {
    return { total: t.summary.Total, passed: t.summary.Passed, failed: t.summary.Failed, skipped: t.summary.Skipped, pending: t.summary.Pending || 0, timedout: 0 };
  }
  if (!t.children || t.children.length === 0) {
    const isTimedOut = !!t.timed_out;
    const counted = isTimedOut || t.passed || t.failed || t.skipped || t.pending;
    return {
      total: counted ? 1 : 0,
      passed: !isTimedOut && t.passed ? 1 : 0,
      failed: !isTimedOut && t.failed ? 1 : 0,
      skipped: !isTimedOut && t.skipped ? 1 : 0,
      pending: !isTimedOut && t.pending ? 1 : 0,
      timedout: isTimedOut ? 1 : 0,
    };
  }
  const r = { total: 0, passed: 0, failed: 0, skipped: 0, pending: 0, timedout: 0 };
  for (const c of t.children) {
    const s = sumNonTaskTests(c);
    r.total += s.total;
    r.passed += s.passed;
    r.failed += s.failed;
    r.skipped += s.skipped;
    r.pending += s.pending;
    r.timedout += s.timedout;
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
  if (t.timed_out) return 'timedout';
  if (t.pending) return 'pending';
  if (t.failed) return 'failed';
  if (t.skipped) return 'skipped';
  if (t.passed) return 'passed';
  return null;
}

// collapseSingleChildChains compresses any container node that has exactly
// one child and no status of its own into its child, joining the names with
// " > ". The child's kind, status flags, duration, violations, and
// navigation fields are preserved. The transform is recursive so chains of
// arbitrary depth collapse into a single row.
export function collapseSingleChildChains(tests: Test[]): Test[] {
  return tests.map(collapseNode);
}

export function collapseLintSingleChildChains(tests: Test[]): Test[] {
  return tests.map(collapseLintNode);
}

function collapseNode(t: Test): Test {
  const children = t.children ?? [];
  if (children.length === 0) return t;

  const collapsedChildren = children.map(collapseNode);

  if (collapsedChildren.length === 1 && isCollapsibleContainer(t)) {
    const child = collapsedChildren[0];
    return {
      ...child,
      name: joinChainNames(t.name, child.name),
      summary: undefined,
    };
  }

  return { ...t, children: collapsedChildren };
}

function collapseLintNode(t: Test): Test {
  const children = t.children ?? [];
  if (children.length === 0) return t;

  const collapsedChildren = children.map(collapseLintNode);

  if (collapsedChildren.length === 1 && isCollapsibleLintContainer(t)) {
    const child = collapsedChildren[0];
    return {
      ...child,
      name: joinChainNames(t.name, child.name),
      summary: undefined,
    };
  }

  return { ...t, children: collapsedChildren };
}

function isCollapsibleContainer(t: Test): boolean {
  if (t.passed || t.failed || t.skipped || t.pending) return false;
  if (t.kind === 'violation') return false;
  if (t.violations && t.violations.length > 0) return false;
  return true;
}

function isCollapsibleLintContainer(t: Test): boolean {
  return t.kind === 'lint-folder' && isCollapsibleContainer(t);
}

function joinChainNames(parent: string | undefined, child: string | undefined): string {
  const p = (parent ?? '').trim();
  const c = (child ?? '').trim();
  if (!p) return c;
  if (!c) return p;
  return `${p} > ${c}`;
}

export function filterTests(
  tests: Test[],
  statusFilter: FilterState<string>,
  frameworkFilter: FilterState<string>,
): Test[] {
  if (statusFilter.size === 0 && frameworkFilter.size === 0) return tests;
  return tests.map(t => filterNode(t, statusFilter, frameworkFilter)).filter(Boolean) as Test[];
}

function filterNode(t: Test, statusFilter: FilterState<string>, frameworkFilter: FilterState<string>): Test | null {
  const hasChildren = !!t.children && t.children.length > 0;

  if (hasChildren) {
    const filtered = t.children!.map(c => filterNode(c, statusFilter, frameworkFilter)).filter(Boolean) as Test[];
    if (filtered.length === 0) return null;
    return { ...t, children: filtered, summary: undefined };
  }

  return matchesLeaf(t, statusFilter, frameworkFilter) ? t : null;
}

function matchesLeaf(t: Test, statusFilter: FilterState<string>, frameworkFilter: FilterState<string>): boolean {
  const s = testStatus(t);
  if (s === null) return false;
  return matchesFilterState(s, statusFilter) && matchesFilterState(t.framework, frameworkFilter);
}

export function humanizeName(name: string, framework?: string): string {
  if (!name) return '';
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

export function countProcesses(node?: ProcessNode): number {
  if (!node) return 0;
  return 1 + (node.children || []).reduce((sum, child) => sum + countProcesses(child), 0);
}

export function findProcessByPID(node: ProcessNode | undefined, pid: number): ProcessNode | null {
  if (!node) return null;
  if (node.pid === pid) return node;
  for (const child of node.children || []) {
    const found = findProcessByPID(child, pid);
    if (found) return found;
  }
  return null;
}

export function processStateIcon(status?: string): string {
  const value = (status || '').toLowerCase();
  if (value.includes('run')) return 'codicon:play-circle';
  if (value.includes('sleep') || value.includes('idle')) return 'codicon:clock';
  if (value.includes('stop') || value.includes('halt')) return 'codicon:debug-pause';
  if (value.includes('zombie') || value.includes('dead')) return 'codicon:error';
  if (value.includes('wait') || value.includes('block')) return 'codicon:debug-step-over';
  return 'codicon:circle-filled';
}

export function processStateColor(status?: string): string {
  const value = (status || '').toLowerCase();
  if (value.includes('run')) return 'text-green-600';
  if (value.includes('sleep') || value.includes('idle')) return 'text-amber-500';
  if (value.includes('stop') || value.includes('halt')) return 'text-orange-600';
  if (value.includes('zombie') || value.includes('dead')) return 'text-red-600';
  if (value.includes('wait') || value.includes('block')) return 'text-blue-600';
  return 'text-gray-400';
}

export function formatBytes(value?: number): string {
  if (!value || value <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = value;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit++;
  }
  return `${size >= 10 || unit === 0 ? size.toFixed(0) : size.toFixed(1)} ${units[unit]}`;
}

export function processLabel(node: ProcessNode): string {
  if (node.name) return node.name;
  if (node.command) {
    const [first] = node.command.split(/\s+/, 1);
    return first || `pid ${node.pid}`;
  }
  return `pid ${node.pid}`;
}
