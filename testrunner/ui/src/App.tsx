import { useState, useEffect, useRef, useMemo, useCallback } from 'preact/hooks';
import type { Test, Snapshot, LinterResult, BenchComparison } from './types';
import { Summary } from './components/Summary';
import { TestNode } from './components/TestNode';
import { DetailPanel, type IgnoreRequest } from './components/DetailPanel';
import { FilterBar, type Filters } from './components/FilterBar';
import { LintFilterBar, type LintGrouping, type LintFilters } from './components/LintFilterBar';
import { LintView } from './components/LintView';
import { BenchView } from './components/BenchView';
import { SplitPane } from './components/SplitPane';
import { copyCurrentViewForAgent, downloadCurrentView } from './export';
import {
  sumNonTaskTests,
  collectFrameworks,
  filterTests,
  totalLintViolations,
  groupLintByLinterFile,
  groupLintByFileLinterRule,
} from './utils';
import { annotateRoutePaths, buildExportRoute, buildRoute, findNodeByRoutePath, parseRoute, type RouteState, type TabKey } from './routes';

function applySnapshot(
  snap: Snapshot,
  startTime: React.MutableRefObject<number | null>,
  endTime: React.MutableRefObject<number | null>,
  doneRef: React.MutableRefObject<boolean>,
  setTests: (t: Test[]) => void,
  setLint: (l: LinterResult[] | undefined) => void,
  setLintRun: (r: boolean) => void,
  setBench: (b: BenchComparison | undefined) => void,
  setDone: (d: boolean) => void,
  setStatus: (s: string) => void,
) {
  if (!startTime.current) startTime.current = Date.now();
  setTests(snap.tests || []);
  setLint(snap.lint);
  setLintRun(!!snap.lint_run);
  setBench(snap.bench);
  if (snap.done) {
    endTime.current = Date.now();
    doneRef.current = true;
    setDone(true);
    setStatus('Test run complete');
  } else {
    setStatus('Running tests...');
  }
}

function currentRouteState(
  tab: TabKey,
  selectedPath: string,
  filters: Filters,
  lintGrouping: LintGrouping,
  lintFilters: LintFilters,
): RouteState {
  return { tab, selectedPath, filters, lintGrouping, lintFilters };
}

export function App() {
  const initialRoute = typeof window !== 'undefined' ? parseRoute(window.location) : {
    tab: 'tests' as TabKey,
    selectedPath: '',
    filters: { status: new Set<string>(), framework: new Set<string>() },
    lintGrouping: 'linter-file' as LintGrouping,
    lintFilters: { severity: new Set(), linter: new Set() },
  };

  const [tests, setTests] = useState<Test[]>([]);
  const [lint, setLint] = useState<LinterResult[] | undefined>(undefined);
  const [lintRun, setLintRun] = useState(false);
  const [bench, setBench] = useState<BenchComparison | undefined>(undefined);
  const [done, setDone] = useState(false);
  const [status, setStatus] = useState('Loading...');
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  const [filters, setFilters] = useState<Filters>(initialRoute.filters);
  const [activeTab, setActiveTab] = useState<TabKey>(initialRoute.tab);
  const [lintGrouping, setLintGrouping] = useState<LintGrouping>(initialRoute.lintGrouping);
  const [lintFilters, setLintFilters] = useState<LintFilters>(initialRoute.lintFilters);
  const [selectedPath, setSelectedPath] = useState(initialRoute.selectedPath);
  const [rerunBusy, setRerunBusy] = useState(false);
  const [ignoreBusy, setIgnoreBusy] = useState(false);
  const [copyState, setCopyState] = useState<'idle' | 'copying' | 'copied' | 'error'>('idle');
  const [copyError, setCopyError] = useState('');
  const startTime = useRef<number | null>(null);
  const endTime = useRef<number | null>(null);
  const [, tick] = useState(0);
  const doneRef = useRef(false);
  const copyResetTimer = useRef<number | null>(null);

  const routeState = useMemo(
    () => currentRouteState(activeTab, selectedPath, filters, lintGrouping, lintFilters),
    [activeTab, selectedPath, filters, lintGrouping, lintFilters],
  );

  const commitRoute = useCallback((next: RouteState, mode: 'push' | 'replace' = 'push') => {
    setActiveTab(next.tab);
    setSelectedPath(next.selectedPath);
    setFilters(next.filters);
    setLintGrouping(next.lintGrouping);
    setLintFilters(next.lintFilters);
    const url = buildRoute(next);
    const current = `${window.location.pathname}${window.location.search}`;
    if (url !== current) {
      if (mode === 'replace') window.history.replaceState({}, '', url);
      else window.history.pushState({}, '', url);
    }
  }, []);

  useEffect(() => {
    const onPopState = () => {
      const next = parseRoute(window.location);
      setActiveTab(next.tab);
      setSelectedPath(next.selectedPath);
      setFilters(next.filters);
      setLintGrouping(next.lintGrouping);
      setLintFilters(next.lintFilters);
    };

    window.addEventListener('popstate', onPopState);
    return () => {
      window.removeEventListener('popstate', onPopState);
    };
  }, []);

  useEffect(() => {
    return () => {
      if (copyResetTimer.current) window.clearTimeout(copyResetTimer.current);
    };
  }, []);

  useEffect(() => {
    fetch('/api/tests')
      .then(r => r.json())
      .then((snap: Snapshot) => {
        applySnapshot(snap, startTime, endTime, doneRef, setTests, setLint, setLintRun, setBench, setDone, setStatus);
      })
      .catch(() => {});

    const es = new EventSource('/api/tests/stream');

    es.addEventListener('message', (e: MessageEvent) => {
      const snap: Snapshot = JSON.parse(e.data);
      applySnapshot(snap, startTime, endTime, doneRef, setTests, setLint, setLintRun, setBench, setDone, setStatus);
      if (snap.done) es.close();
    });

    es.addEventListener('done', () => {
      endTime.current = Date.now();
      doneRef.current = true;
      setDone(true);
      setStatus('Test run complete');
      es.close();
    });

    es.onerror = () => {
      if (!doneRef.current) setStatus('Connection lost — retrying...');
    };

    const timer = setInterval(() => {
      if (startTime.current && !doneRef.current) tick(n => n + 1);
    }, 1000);

    return () => { es.close(); clearInterval(timer); };
  }, []);

  const totals = useMemo(() => {
    const t = { total: 0, passed: 0, failed: 0, skipped: 0, pending: 0 };
    for (const test of tests) {
      const s = sumNonTaskTests(test);
      t.total += s.total;
      t.passed += s.passed;
      t.failed += s.failed;
      t.skipped += s.skipped;
      t.pending += s.pending;
    }
    return t;
  }, [tests]);

  const frameworks = useMemo(() => collectFrameworks(tests), [tests]);
  const filteredTests = useMemo(
    () => annotateRoutePaths(filterTests(tests, filters.status, filters.framework)),
    [tests, filters],
  );
  const lintTree = useMemo(() => {
    const tree = lintGrouping === 'linter-file'
      ? groupLintByLinterFile(lint, lintFilters)
      : groupLintByFileLinterRule(lint, lintFilters);
    return annotateRoutePaths(tree);
  }, [lint, lintGrouping, lintFilters]);
  const lintTotal = useMemo(() => totalLintViolations(lint), [lint]);

  const selected = useMemo(() => {
    if (!selectedPath) return null;
    if (activeTab === 'tests') return findNodeByRoutePath(filteredTests, selectedPath);
    if (activeTab === 'lint') return findNodeByRoutePath(lintTree, selectedPath);
    return null;
  }, [activeTab, filteredTests, lintTree, selectedPath]);

  useEffect(() => {
    if (!selectedPath) return;
    const ready = activeTab === 'tests'
      ? (tests.length > 0 || done)
      : activeTab === 'lint'
        ? (lintRun || done || !!lint)
        : true;
    if (!ready || selected) return;
    commitRoute({ ...routeState, selectedPath: '' }, 'replace');
  }, [selectedPath, selected, activeTab, tests.length, done, lintRun, lint, routeState, commitRoute]);

  const onTabChange = useCallback((tab: TabKey) => {
    commitRoute({
      ...routeState,
      tab,
      selectedPath: '',
    });
  }, [routeState, commitRoute]);

  const onSelect = useCallback((test: Test) => {
    const nextPath = test.route_path === selectedPath ? '' : (test.route_path || '');
    commitRoute({
      ...routeState,
      selectedPath: nextPath,
    });
  }, [routeState, selectedPath, commitRoute]);

  const onTestFiltersChange = useCallback((nextFilters: Filters) => {
    commitRoute({
      ...routeState,
      tab: 'tests',
      filters: nextFilters,
      selectedPath: '',
    });
  }, [routeState, commitRoute]);

  const onLintGroupingChange = useCallback((grouping: LintGrouping) => {
    commitRoute({
      ...routeState,
      tab: 'lint',
      lintGrouping: grouping,
      selectedPath: '',
    });
  }, [routeState, commitRoute]);

  const onLintFiltersChange = useCallback((nextFilters: LintFilters) => {
    commitRoute({
      ...routeState,
      tab: 'lint',
      lintFilters: nextFilters,
      selectedPath: '',
    });
  }, [routeState, commitRoute]);

  const onRerun = useCallback(async (t: Test) => {
    if (t.kind === 'violation' || t.kind === 'lint-root' || t.kind === 'lint-folder' || t.kind === 'linter' || t.kind === 'lint-file' || t.kind === 'lint-rule' || t.framework === 'task') return;
    if (rerunBusy) return;
    const collectPaths = (n: Test, out: Set<string>) => {
      if (n.package_path) out.add(n.package_path);
      (n.children || []).forEach(c => collectPaths(c, out));
    };
    const paths = new Set<string>();
    collectPaths(t, paths);
    const isLeaf = !t.children || t.children.length === 0;
    const body = {
      package_paths: Array.from(paths),
      test_name: isLeaf ? t.name : '',
      suite: t.suite || [],
      framework: t.framework || '',
    };
    setRerunBusy(true);
    setStatus(`Rerunning ${t.name}...`);
    doneRef.current = false;
    setDone(false);
    try {
      const res = await fetch('/api/rerun', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (res.status === 409) {
        setStatus('Another rerun is in progress');
      } else if (!res.ok) {
        const text = await res.text();
        setStatus(`Rerun failed: ${text.trim()}`);
      } else {
        setStatus('Rerun complete');
      }
    } catch (e: any) {
      setStatus(`Rerun error: ${e?.message || e}`);
    } finally {
      setRerunBusy(false);
    }
  }, [rerunBusy]);

  const onIgnore = useCallback(async (req: IgnoreRequest) => {
    if (ignoreBusy) return;
    setIgnoreBusy(true);
    try {
      const res = await fetch('/api/lint/ignore', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(req),
      });
      if (!res.ok) {
        const text = await res.text();
        setStatus(`Ignore failed: ${text.trim()}`);
      } else {
        setStatus('Ignore rule saved to .gavel.yaml');
        commitRoute({ ...routeState, selectedPath: '' }, 'replace');
      }
    } catch (e: any) {
      setStatus(`Ignore error: ${e?.message || e}`);
    } finally {
      setIgnoreBusy(false);
    }
  }, [ignoreBusy, routeState, commitRoute]);

  const showLintTab = lintRun;
  const showBenchTab = !!bench;
  const showTabs = showLintTab || showBenchTab;
  const benchRegressions = bench?.deltas?.filter(d => d.significant && d.delta_pct > bench.threshold).length || 0;
  const hasContent = tests.length > 0 || (lintRun && lintTotal > 0) || showBenchTab;
  const canExportCurrentView = (activeTab === 'tests' && tests.length > 0)
    || (activeTab === 'lint' && lintRun)
    || (activeTab === 'bench' && !!bench);
  const jsonExportURL = useMemo(() => buildExportRoute(routeState, 'json'), [routeState]);
  const markdownExportURL = useMemo(() => buildExportRoute(routeState, 'md'), [routeState]);

  const resetCopyFeedback = useCallback((nextState: 'copied' | 'error', error: string = '') => {
    setCopyState(nextState);
    setCopyError(error);
    if (copyResetTimer.current) window.clearTimeout(copyResetTimer.current);
    copyResetTimer.current = window.setTimeout(() => {
      setCopyState('idle');
      setCopyError('');
      copyResetTimer.current = null;
    }, nextState === 'copied' ? 2000 : 3000);
  }, []);

  const onDownloadJSON = useCallback(() => {
    downloadCurrentView(routeState, 'json');
  }, [routeState]);

  const onDownloadMarkdown = useCallback(() => {
    downloadCurrentView(routeState, 'md');
  }, [routeState]);

  const onCopyForAgent = useCallback(async () => {
    if (copyState === 'copying') return;
    setCopyState('copying');
    setCopyError('');
    if (copyResetTimer.current) {
      window.clearTimeout(copyResetTimer.current);
      copyResetTimer.current = null;
    }
    try {
      await copyCurrentViewForAgent(routeState);
      resetCopyFeedback('copied');
    } catch (e: any) {
      resetCopyFeedback('error', e?.message || 'Copy failed');
    }
  }, [copyState, routeState, resetCopyFeedback]);

  return (
    <div class="bg-gray-100 h-screen flex flex-col">
      <div class="border-b bg-white px-6 py-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <h1 class="text-xl font-bold text-gray-900">
              <iconify-icon icon="codicon:beaker" class="mr-1.5 text-blue-600" />
              {activeTab === 'lint' ? 'Lint Results' : activeTab === 'bench' ? 'Benchmark Comparison' : 'Test Results'}
            </h1>
            {hasContent && (
              <div class="flex gap-1">
                <button
                  class="text-xs px-2 py-1 rounded border border-gray-300 text-gray-600 hover:bg-gray-200 transition-colors"
                  onClick={() => setExpandAll(true)}
                  title="Expand all"
                >
                  <iconify-icon icon="codicon:expand-all" class="mr-0.5" />
                  Expand
                </button>
                <button
                  class="text-xs px-2 py-1 rounded border border-gray-300 text-gray-600 hover:bg-gray-200 transition-colors"
                  onClick={() => setExpandAll(false)}
                  title="Collapse all"
                >
                  <iconify-icon icon="codicon:collapse-all" class="mr-0.5" />
                  Collapse
                </button>
              </div>
            )}
            {canExportCurrentView && (
              <div class="flex gap-1">
                <button
                  class="text-xs px-2 py-1 rounded border border-gray-300 text-gray-600 hover:bg-gray-200 transition-colors"
                  onClick={onDownloadJSON}
                  title={jsonExportURL}
                >
                  <iconify-icon icon="codicon:json" class="mr-0.5" />
                  JSON
                </button>
                <button
                  class="text-xs px-2 py-1 rounded border border-gray-300 text-gray-600 hover:bg-gray-200 transition-colors"
                  onClick={onDownloadMarkdown}
                  title={markdownExportURL}
                >
                  <iconify-icon icon="codicon:markdown" class="mr-0.5" />
                  Markdown
                </button>
                <button
                  class={`text-xs px-2 py-1 rounded border transition-colors ${
                    copyState === 'error'
                      ? 'border-red-300 text-red-700 bg-red-50 hover:bg-red-100'
                      : copyState === 'copied'
                        ? 'border-green-300 text-green-700 bg-green-50 hover:bg-green-100'
                        : 'border-gray-300 text-gray-600 hover:bg-gray-200'
                  }`}
                  onClick={onCopyForAgent}
                  title={copyError || markdownExportURL}
                >
                  <iconify-icon icon={copyState === 'copied' ? 'codicon:check' : copyState === 'copying' ? 'svg-spinners:ring-resize' : 'codicon:copy'} class="mr-0.5" />
                  {copyState === 'copying' ? 'Copying...' : copyState === 'copied' ? 'Copied' : copyState === 'error' ? 'Copy failed' : 'Copy for Agent'}
                </button>
              </div>
            )}
            <span class="text-sm text-gray-400">{status}</span>
          </div>
          <Summary tests={tests} startTime={startTime.current} endTime={endTime.current} done={done} />
        </div>

        {showTabs && (
          <div class="mt-2 flex items-center gap-1 border-b border-gray-200 -mx-6 px-6">
            <TabButton
              active={activeTab === 'tests'}
              onClick={() => onTabChange('tests')}
              icon="codicon:beaker"
              label="Tests"
              count={totals.failed > 0 ? totals.failed : totals.total}
              countColor={totals.failed > 0 ? 'bg-red-500' : 'bg-gray-400'}
            />
            {showLintTab && (
              <TabButton
                active={activeTab === 'lint'}
                onClick={() => onTabChange('lint')}
                icon="codicon:lightbulb"
                label="Lint"
                count={lintTotal}
                countColor={lintTotal > 0 ? 'bg-yellow-500' : 'bg-gray-400'}
              />
            )}
            {showBenchTab && (
              <TabButton
                active={activeTab === 'bench'}
                onClick={() => onTabChange('bench')}
                icon="codicon:graph"
                label="Bench"
                count={benchRegressions > 0 ? benchRegressions : (bench?.deltas?.length || 0)}
                countColor={benchRegressions > 0 ? 'bg-red-500' : 'bg-gray-400'}
              />
            )}
          </div>
        )}

        {activeTab === 'tests' && tests.length > 0 && (
          <div class="mt-2">
            <FilterBar filters={filters} onChange={onTestFiltersChange} counts={totals} frameworks={frameworks} />
          </div>
        )}
        {activeTab === 'lint' && lintRun && (
          <div class="mt-2">
            <LintFilterBar
              lint={lint}
              grouping={lintGrouping}
              onGroupingChange={onLintGroupingChange}
              filters={lintFilters}
              onFiltersChange={onLintFiltersChange}
            />
          </div>
        )}
      </div>

      <SplitPane
        defaultSplit={50}
        left={
          <>
            {activeTab === 'tests' && (
              <>
                {filteredTests.map((t, i) => (
                  <TestNode key={i} test={t} depth={0} expandAll={expandAll} selected={selected} onSelect={onSelect} onRerun={onRerun} rerunBusy={rerunBusy} />
                ))}
                {tests.length === 0 && !done && (
                  <div class="p-8 text-center text-gray-400">
                    <iconify-icon icon="svg-spinners:ring-resize" class="text-3xl text-blue-500" />
                    <p class="mt-2">Waiting for test results...</p>
                  </div>
                )}
                {filteredTests.length === 0 && tests.length > 0 && (
                  <div class="p-8 text-center text-gray-400 text-sm">
                    No tests match the current filters
                  </div>
                )}
              </>
            )}
            {activeTab === 'lint' && (
              <LintView
                lint={lint}
                tree={lintTree}
                expandAll={expandAll}
                selected={selected}
                onSelect={onSelect}
              />
            )}
            {activeTab === 'bench' && <BenchView bench={bench} />}
          </>
        }
        right={<DetailPanel test={selected} onRerun={onRerun} rerunBusy={rerunBusy} onIgnore={onIgnore} ignoreBusy={ignoreBusy} />}
      />
    </div>
  );
}

function TabButton({ active, onClick, icon, label, count, countColor }: {
  active: boolean;
  onClick: () => void;
  icon: string;
  label: string;
  count: number;
  countColor: string;
}) {
  return (
    <button
      class={`inline-flex items-center gap-1.5 px-3 py-1.5 text-sm border-b-2 -mb-px transition-colors ${
        active
          ? 'border-blue-500 text-blue-700 font-semibold'
          : 'border-transparent text-gray-500 hover:text-gray-800'
      }`}
      onClick={onClick}
    >
      <iconify-icon icon={icon} class="text-base" />
      {label}
      {count > 0 && (
        <span class={`inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded-full text-[10px] font-bold text-white ${countColor}`}>
          {count}
        </span>
      )}
    </button>
  );
}
