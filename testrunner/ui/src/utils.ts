import type { Test } from './types';

export function statusIcon(t: Test): string {
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
