import { useEffect, useMemo, useState } from 'react';
import {
  TestRunner,
  emptyTestFilters,
  filterTests,
  type Test,
  type TestFilters,
} from '@flanksource/clicky-ui/data';
import { GavelIcon } from '../GavelIcon';
import { LintResults } from './LintResults';
import type { RunSnapshot } from './types';

export function TestRunDetail({ project, runId }: { project: string; runId: string }) {
  const [snap, setSnap] = useState<RunSnapshot | null>(null);
  const [error, setError] = useState('');
  const [selected, setSelected] = useState<Test | null>(null);
  const [filters, setFilters] = useState<TestFilters>(emptyTestFilters());
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  const [view, setView] = useState<'tests' | 'lint'>('tests');

  useEffect(() => {
    setSnap(null);
    setError('');
    setSelected(null);
    setFilters(emptyTestFilters());
    const url = `/api/tests/run?project=${encodeURIComponent(project)}&runId=${encodeURIComponent(runId)}`;
    fetch(url)
      .then(r => (r.ok ? r.json() : r.json().then(e => Promise.reject(e.error || 'failed to load run'))))
      .then((s: RunSnapshot) => setSnap(s))
      .catch(e => setError(typeof e === 'string' ? e : 'Failed to load run'));
  }, [project, runId]);

  const tests = snap?.tests ?? [];
  const lint = useMemo(() => (snap?.lint ?? []).filter(l => !l.skipped), [snap]);
  const hasTests = tests.length > 0;
  const hasLint = lint.length > 0;
  const lintCount = useMemo(() => lint.reduce((n, l) => n + l.violations.length, 0), [lint]);
  const visible = useMemo(() => filterTests(tests, filters.status, filters.framework), [tests, filters]);

  // Default to whichever section the run actually has, re-evaluated per run.
  useEffect(() => {
    setView(hasTests ? 'tests' : 'lint');
  }, [hasTests, runId]);

  if (error) return <Centered>{error}</Centered>;
  if (!snap) return <Centered>Loading…</Centered>;
  if (!hasTests && !hasLint) return <Centered>This run has no tests or lint findings.</Centered>;

  return (
    <div className="flex h-full flex-col">
      {hasTests && hasLint && (
        <div className="flex gap-1 border-b border-border px-2 py-1.5">
          <ToggleButton active={view === 'tests'} onClick={() => setView('tests')} icon="codicon:beaker">
            Tests {tests.length > 0 && <Count>{summaryTotal(snap)}</Count>}
          </ToggleButton>
          <ToggleButton active={view === 'lint'} onClick={() => setView('lint')} icon="codicon:warning">
            Lint <Count>{lintCount}</Count>
          </ToggleButton>
        </div>
      )}
      <div className="min-h-0 flex-1">
        {view === 'tests' && hasTests ? (
          <TestRunner
            tests={visible}
            selected={selected}
            filters={filters}
            expandAll={expandAll}
            done
            runMeta={snap.metadata}
            status={snap.status}
            title={null}
            onSelect={setSelected}
            onFiltersChange={setFilters}
            onExpandAllChange={setExpandAll}
          />
        ) : (
          <LintResults lint={lint} />
        )}
      </div>
    </div>
  );
}

// summaryTotal is the leaf-test count, used only for the toggle badge.
function summaryTotal(snap: RunSnapshot): number {
  let total = 0;
  const walk = (t: Test) => {
    if (t.children && t.children.length > 0) t.children.forEach(walk);
    else if (t.passed || t.failed || t.skipped || t.warned) total += 1;
  };
  snap.tests.forEach(walk);
  return total;
}

function ToggleButton({
  active,
  onClick,
  icon,
  children,
}: {
  active: boolean;
  onClick: () => void;
  icon: string;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs transition-colors ${
        active ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground hover:bg-muted'
      }`}
    >
      <GavelIcon name={icon} />
      {children}
    </button>
  );
}

function Count({ children }: { children: React.ReactNode }) {
  return <span className="tabular-nums text-[10px] text-muted-foreground">{children}</span>;
}

function Centered({ children }: { children: React.ReactNode }) {
  return <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">{children}</div>;
}
