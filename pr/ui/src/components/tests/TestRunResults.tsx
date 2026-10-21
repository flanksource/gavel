import { useCallback, useEffect, useMemo, useState, type ComponentType } from 'react';
import {
  DEFAULT_LINT_GROUPING,
  LintResults,
  TestRunner,
  annotateRoutePaths,
  emptyLintFilters,
  emptyTestFilters,
  filterTests,
  findNodeByRoutePath,
  totalLintViolations,
  type LintFilters,
  type LintGrouping,
  type Test,
  type TestFilters,
} from '@flanksource/clicky-ui/data';
import { Button } from '@flanksource/clicky-ui/components';
import { UiAdd, UiBeaker, UiTerminal, UiWarningTriangle, type IconProps } from '@flanksource/clicky-ui/icons';
import { CreateTodoFromRunDialog } from './CreateTodoFromRunDialog';
import { buildRunFailureCandidates } from './RunFailureCandidates';
import { RunProcesses, fetchProcessMetricsFromSamples, processLinksByPid } from '@flanksource/gavel/testrunner';
import type { ProcessStepLink, Test as GavelTest, TracesStatus } from '@flanksource/gavel/testrunner';
import { runRegistry, useRunProcesses } from './runTraces';
import type { RunSnapshot } from './types';

type View = 'tests' | 'lint' | 'processes';

/** The results of one run, remounted per run so no selection outlives the run it was made in. */
export function TestRunResults(props: RunResultsProps) {
  return <RunResults key={props.runKey} {...props} />;
}

interface RunResultsProps {
  snapshot: RunSnapshot;
  done: boolean;
  runKey: string;
  projectName: string;
  projectDir: string;
  onTodoCreated?: () => void;
  emptyMessage?: string;
  /** The run's trace stores exactly as the server that returned the snapshot reported them. */
  traces?: TracesStatus;
  /**
   * URL prefix of that server: '' when `traces` already carries the mounts as
   * the dashboard serves them (the Tests tab), `/api/project-runs/<id>` for a
   * run server whose mounts are relative to it.
   */
  base?: string;
}

function RunResults({
  traces: reported,
  base = '',
  snapshot,
  done,
  runKey,
  projectName,
  projectDir,
  onTodoCreated,
  emptyMessage = 'This run has no tests or lint findings.',
}: RunResultsProps) {
  const [selectedPath, setSelectedPath] = useState<string | null>(null);
  const [filters, setFilters] = useState<TestFilters>(emptyTestFilters());
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  const [view, setView] = useState<View>('tests');
  const [lintGrouping, setLintGrouping] = useState<LintGrouping>(DEFAULT_LINT_GROUPING);
  const [lintFilters, setLintFilters] = useState<LintFilters>(emptyLintFilters());
  const [lintSelected, setLintSelected] = useState<string | null>(null);
  const [lintExpandAll, setLintExpandAll] = useState<boolean | null>(null);
  const [todoOpen, setTodoOpen] = useState(false);
  const processes = useRunProcesses({ traces: reported, base, done, active: view === 'processes' });
  const tests = snapshot.tests ?? [];
  const lint = useMemo(() => (snapshot.lint ?? []).filter(result => !result.skipped), [snapshot.lint]);
  const hasTests = tests.length > 0;
  const hasLint = Boolean(snapshot.status?.lint_run) || lint.length > 0;
  const lintCount = useMemo(() => totalLintViolations(lint), [lint]);
  const routedTests = useMemo(() => annotateRoutePaths(tests), [tests]);
  const visible = useMemo(
    () => annotateRoutePaths(filterTests(tests, filters.status, filters.framework)),
    [tests, filters],
  );
  const selected = useMemo(() => findNodeByRoutePath(visible, selectedPath ?? ''), [visible, selectedPath]);
  const stepLinks = useMemo(
    () => (processes.available ? processLinksByPid(routedTests as GavelTest[]) : new Map<number, ProcessStepLink>()),
    [processes.available, routedTests],
  );
  const { setSelectedKey } = processes;
  const onOpenProcess = useCallback((key: string) => {
    setSelectedKey(key);
    setView('processes');
  }, [setSelectedKey]);
  const onOpenStep = useCallback((link: ProcessStepLink) => {
    setFilters(emptyTestFilters());
    setSelectedPath(link.routePath);
    setView('tests');
  }, []);
  const registry = useMemo(
    () => runRegistry({ traces: processes.traces, processes: processes.trace.nodes, onOpenProcess }),
    [processes.traces, processes.trace.nodes, onOpenProcess],
  );
  const fetchProcessMetrics = useMemo(
    () => fetchProcessMetricsFromSamples({ samples: processes.trace.samples, nodes: processes.trace.nodes }),
    [processes.trace.samples, processes.trace.nodes],
  );
  const failures = useMemo(() => buildRunFailureCandidates(snapshot), [snapshot]);
  const canCreateTodo = done && failures.length > 0;

  useEffect(() => {
    setView(hasTests ? 'tests' : 'lint');
  }, [hasTests]);

  if (!hasTests && !hasLint) return <Centered>{emptyMessage}</Centered>;

  return (
    <>
      <div className="flex h-full flex-col">
        {(hasTests && hasLint || processes.available || canCreateTodo) && (
          <div className="flex items-center gap-1 border-b border-border px-2 py-1.5">
            {(hasTests && hasLint || processes.available) && (
              <>
                {hasTests && (
                  <ToggleButton active={view === 'tests'} onClick={() => setView('tests')} icon={UiBeaker}>
                    Tests <Count>{summaryTotal(snapshot)}</Count>
                  </ToggleButton>
                )}
                {hasLint && (
                  <ToggleButton active={view === 'lint'} onClick={() => setView('lint')} icon={UiWarningTriangle}>
                    Lint <Count>{lintCount}</Count>
                  </ToggleButton>
                )}
                {processes.available && (
                  <ToggleButton active={view === 'processes'} onClick={() => setView('processes')} icon={UiTerminal}>
                    Processes
                  </ToggleButton>
                )}
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
          {view === 'processes' && processes.available ? (
            <RunProcesses
              trace={processes.trace}
              selectedKey={processes.selectedKey}
              onSelect={processes.setSelectedKey}
              hideExited={processes.hideExited}
              onHideExitedChange={processes.setHideExited}
              expandAll={processes.expandAll}
              onExpandAllChange={processes.setExpandAll}
              stepLinks={stepLinks}
              onOpenStep={onOpenStep}
            />
          ) : view === 'tests' && hasTests ? (
            <TestRunner
              adapters={registry}
              tests={visible}
              allTests={routedTests}
              selected={selected}
              filters={filters}
              expandAll={expandAll}
              done={done}
              runMeta={snapshot.metadata}
              status={snapshot.status}
              title={null}
              fetchProcessMetrics={fetchProcessMetrics}
              onSelect={node => setSelectedPath(node?.route_path ?? null)}
              onFiltersChange={setFilters}
              onExpandAllChange={setExpandAll}
            />
          ) : (
            <LintResults
              results={lint}
              title={null}
              grouping={lintGrouping}
              onGroupingChange={setLintGrouping}
              filters={lintFilters}
              onFiltersChange={setLintFilters}
              selected={lintSelected}
              onSelect={setLintSelected}
              expandAll={lintExpandAll}
              onExpandAllChange={setLintExpandAll}
            />
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
