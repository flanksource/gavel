import { useEffect, useMemo, useState, type ComponentType } from 'react';
import {
  TestRunner,
  emptyTestFilters,
  filterTests,
  type Test,
  type TestFilters,
} from '@flanksource/clicky-ui/data';
import { Button } from '@flanksource/clicky-ui/components';
import { UiAdd, UiBeaker, UiWarningTriangle, type IconProps } from '@flanksource/clicky-ui/icons';
import { CreateTodoFromRunDialog } from './CreateTodoFromRunDialog';
import { LintResults } from './LintResults';
import { buildRunFailureCandidates } from './RunFailureCandidates';
import type { RunSnapshot } from './types';

export function TestRunResults({
  snapshot,
  done,
  runKey,
  projectName,
  projectDir,
  onTodoCreated,
  emptyMessage = 'This run has no tests or lint findings.',
}: {
  snapshot: RunSnapshot;
  done: boolean;
  runKey: string;
  projectName: string;
  projectDir: string;
  onTodoCreated?: () => void;
  emptyMessage?: string;
}) {
  const [selected, setSelected] = useState<Test | null>(null);
  const [filters, setFilters] = useState<TestFilters>(emptyTestFilters());
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  const [view, setView] = useState<'tests' | 'lint'>('tests');
  const [todoOpen, setTodoOpen] = useState(false);
  const tests = snapshot.tests ?? [];
  const lint = useMemo(() => (snapshot.lint ?? []).filter(result => !result.skipped), [snapshot.lint]);
  const hasTests = tests.length > 0;
  const hasLint = Boolean(snapshot.status?.lint_run) || lint.length > 0;
  const lintCount = useMemo(() => lint.reduce((count, result) => count + (result.violations?.length ?? 0), 0), [lint]);
  const visible = useMemo(() => filterTests(tests, filters.status, filters.framework), [tests, filters]);
  const failures = useMemo(() => buildRunFailureCandidates(snapshot), [snapshot]);
  const canCreateTodo = done && failures.length > 0;

  useEffect(() => {
    setSelected(null);
    setFilters(emptyTestFilters());
    setExpandAll(null);
    setTodoOpen(false);
    setView(hasTests ? 'tests' : 'lint');
  }, [hasTests, runKey]);

  if (!hasTests && !hasLint) return <Centered>{emptyMessage}</Centered>;

  return (
    <>
      <div className="flex h-full flex-col">
        {(hasTests && hasLint || canCreateTodo) && (
          <div className="flex items-center gap-1 border-b border-border px-2 py-1.5">
            {hasTests && hasLint && (
              <>
                <ToggleButton active={view === 'tests'} onClick={() => setView('tests')} icon={UiBeaker}>
                  Tests <Count>{summaryTotal(snapshot)}</Count>
                </ToggleButton>
                <ToggleButton active={view === 'lint'} onClick={() => setView('lint')} icon={UiWarningTriangle}>
                  Lint <Count>{lintCount}</Count>
                </ToggleButton>
              </>
            )}
            {canCreateTodo && (
              <Button type="button" variant="outline" size="sm" className="ml-auto" onClick={() => setTodoOpen(true)}>
                <UiAdd /> Create todo
              </Button>
            )}
          </div>
        )}
        <div className="min-h-0 flex-1">
          {view === 'tests' && hasTests ? (
            <TestRunner
              tests={visible}
              selected={selected}
              filters={filters}
              expandAll={expandAll}
              done={done}
              runMeta={snapshot.metadata}
              status={snapshot.status}
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
      <CreateTodoFromRunDialog
        open={todoOpen}
        projectName={projectName}
        projectDir={projectDir}
        runId={runKey}
        candidates={failures}
        onClose={() => setTodoOpen(false)}
        onCreated={onTodoCreated}
      />
    </>
  );
}

function summaryTotal(snapshot: RunSnapshot): number {
  let total = 0;
  const walk = (test: Test) => {
    if (test.children && test.children.length > 0) test.children.forEach(walk);
    else if (test.passed || test.failed || test.skipped || test.warned) total += 1;
  };
  (snapshot.tests ?? []).forEach(walk);
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
  icon: ComponentType<IconProps>;
  children: React.ReactNode;
}) {
  const Icon = icon;
  return (
    <Button
      type="button"
      variant="ghost"
      size="sm"
      onClick={onClick}
      className={`flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs transition-colors ${
        active ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground hover:bg-muted'
      }`}
    >
      <Icon />
      {children}
    </Button>
  );
}

function Count({ children }: { children: React.ReactNode }) {
  return <span className="tabular-nums text-[10px] text-muted-foreground">{children}</span>;
}

function Centered({ children }: { children: React.ReactNode }) {
  return <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">{children}</div>;
}
