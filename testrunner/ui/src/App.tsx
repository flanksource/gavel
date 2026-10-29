import { Suspense, lazy, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Icon,
  LintResults,
  TestRunner,
  aggregateStatusCounts,
  annotateRoutePaths,
  buildLintTree,
  collapseSingleChildChains,
  filterTests,
  findLintNode,
  findNodeByRoutePath,
  isProcessExited,
  totalLintViolations,
  violationIgnoreAction,
  type LintAction,
  type LintFilters,
  type LintGrouping,
  type LinterResult,
  type OpenFileTarget,
  type Test as ClickyTest,
  type TestFilters,
  type Violation,
} from '@flanksource/clicky-ui/data';
import { lintFindings, testFindings, type FileFinding, type FileRef, type FileReviewView } from '@flanksource/clicky-ui/file-findings';
import { Button, Tabs } from '@flanksource/clicky-ui/components';
import {
  UiArrowLeft,
  UiBeaker,
  UiChartLineUp,
  UiFileCode,
  UiHistory,
  UiListChecks,
  UiLoader,
  UiRefresh,
  UiServerProcess,
} from '@flanksource/clicky-ui/icons';
import { gavelRegistry } from './adapters';
import { BenchView } from './components/BenchView';
import { FixtureBenchView } from './components/FixtureBenchView';
import { LintTasks } from './components/LintTasks';
import { RerunDialog } from './components/RerunDialog';
import { RunActions } from './components/RunActions';
import { RunHistory } from './components/RunHistory';
import { RunProcesses } from './components/RunProcesses';
import { RunSwitcher } from './components/RunSwitcher';
import { apiBase, apiUrl } from './config';
import { fetchProcessMetricsFromSamples } from './processMetrics';
import { useProcessTrace } from './hooks/use-process-trace';
import { useRunHistory } from './hooks/use-run-history';
import { useTestRun } from './hooks/use-test-run';
import { buildLintRerunBody } from './lintRerun';
import {

  buildRoute,
  parseRoute,
  type RouteState,
  type TabKey,
} from './routes';
import { rebaseTraces, processMount } from './traces';
import type { RerunRequest, Test, TestEditAction, TestEditScope } from './types';
import { withExportPaths } from './exportPaths';
import { isLintOnlyPhase, processLinksByPid, stripTestTaskGroup, type ProcessStepLink } from './utils';

function countFixtureWork(tests: Test[]): number {
  return tests.reduce((count, test) => count + (test.fixture ? 1 : 0) + countFixtureWork(test.children || []), 0);
}

const asGavel = (tests: ClickyTest[]): Test[] => tests as Test[];

// The Files tab (and with it the file viewer and its highlighter) is a separate chunk, loaded
// the first time the tab is shown.
const FilesTab = lazy(() => import('./FilesTab'));
// A prompt lint result's live session pulls in the AI bundle, so it too loads
// on first use.
const PromptRunView = lazy(() => import('./PromptRunView'));

function renderPromptRun(result: LinterResult) {
  if (!result.run_id) return null;
  return (
    <Suspense fallback={<p role="status" className="text-sm text-muted-foreground">Loading the prompt run…</p>}>
      <PromptRunView result={result} baseUrl={apiBase()} />
    </Suspense>
  );
}

function FilesLoading() {
  return (
    <div className="flex h-full items-center justify-center text-sm text-muted-foreground" role="status">
      <Icon icon={UiLoader} className="mr-2 animate-spin text-blue-500" />
      Loading files...
    </div>
  );
}

// withWorkDir fills a test's missing work dir with the run's git root, so a test
// that recorded only an absolute file still resolves to a path under a root the
// file endpoint accepts.
function withWorkDir(tests: ClickyTest[], root: string | undefined): ClickyTest[] {
  if (!root) return tests;
  return tests.map(test => ({
    ...test,
    work_dir: test.work_dir || root,
    ...(test.children ? { children: withWorkDir(test.children, root) } : {}),
  }));
}

function findingKey(finding: FileFinding): string {
  return [finding.workDir, finding.file, finding.line, finding.rule, finding.message].join('\0');
}

// sourceUrl is where a mapped source is browsed at a finding's line: the
// source's URL with line=<n> added, unless it already names a line.
export function sourceUrl(url: string | undefined, line: number | undefined): string | undefined {
  if (!url) return undefined;
  if (!line) return url;
  const parsed = new URL(url);
  if (!parsed.searchParams.has('line')) parsed.searchParams.set('line', String(line));
  return parsed.toString();
}

async function fetchFile(ref: FileRef, scope: { runId?: string }): Promise<string> {
  const params = new URLSearchParams();
  if (ref.workDir) params.set('work_dir', ref.workDir);
  params.set('path', ref.file);
  const res = await fetch(apiUrl(`/api/file?${params}`, scope));
  const body = await res.text();
  if (!res.ok) throw new Error(`${res.status} ${body.trim() || res.statusText}`);
  return body;
}

// withOwnNames records every node's own name, which collapsing a single-child
// chain into it ("pkg/ > TestFoo") would otherwise lose and which a rerun of
// that test is addressed by.
function withOwnNames(tests: Test[]): Test[] {
  return tests.map(test => ({ ...test, own_name: test.name, ...(test.children ? { children: withOwnNames(test.children) } : {}) }));
}

// rerunBody is the rerun request for a test node: every package beneath it, its
// module root, and — for a leaf — the test itself. `error` explains a node whose
// packages span several module roots, which the server cannot rerun as one.
function rerunBody(node: Test): { body: RerunRequest } | { error: string } {
  const paths = new Set<string>();
  const workDirs = new Set<string>();
  const collect = (n: Test) => {
    if (n.package_path) paths.add(n.package_path);
    if (n.work_dir) workDirs.add(n.work_dir);
    (n.children || []).forEach(collect);
  };
  collect(node);
  if (workDirs.size > 1) return { error: 'Rerun across multiple module roots is not supported' };
  const isLeaf = !node.children || node.children.length === 0;
  return {
    body: {
      package_paths: Array.from(paths),
      work_dir: workDirs.size === 1 ? Array.from(workDirs)[0] : '',
      test_name: isLeaf ? node.own_name ?? node.name : '',
      suite: node.suite || [],
      framework: node.framework || '',
    },
  };
}

function PageHeader({ icon, title, status, actions }: { icon: typeof UiBeaker; title: string; status?: string; actions?: React.ReactNode }) {
  return (
    <div className="flex min-w-0 items-center gap-3 border-b border-border bg-card px-density-4 py-density-2" data-page-header="">
      <h1 className="inline-flex shrink-0 items-center gap-1.5 text-base font-bold">
        <Icon icon={icon} className="text-primary" />
        {title}
      </h1>
      {status && <span className="truncate text-sm text-muted-foreground">{status}</span>}
      {actions && <div className="ml-auto flex shrink-0 items-center gap-1">{actions}</div>}
    </div>
  );
}

function Waiting({ done }: { done: boolean }) {
  return (
    <div className="flex h-full items-center justify-center p-8 text-center text-sm text-muted-foreground">
      {done ? 'No tests were run' : (
        <div>
          <Icon icon={UiLoader} className="animate-spin text-3xl text-blue-500" />
          <p className="mt-2">Waiting for test results...</p>
        </div>
      )}
    </div>
  );
}

const NOTICE_MS = 8000;
const GLOBAL_STOP_KEY = '__global__';

export interface AppProps {
  /** Scopes every request to one run of a MultiServer (`${base}/${runId}/api/...`). */
  runId?: string;
  /** Opens another run's page; each run is served under its own base path, so this is a full page load. */
  navigate?: (href: string) => void;
}

const loadPage = (href: string) => window.location.assign(href);

export function App({ runId, navigate = loadPage }: AppProps = {}) {
  const scope = useMemo(() => (runId ? { runId } : {}), [runId]);
  const base = apiBase(scope);
  const run = useTestRun(scope.runId ? { runId: scope.runId } : {});
  const { snapshot, tests, lint, status: snapshotStatus, runMeta, done, statusText, error: streamError } = run;
  const performance = snapshot?.performance;
  const runError = snapshot?.error;
  const lintRun = !!snapshotStatus.lint_run;
  const history = useRunHistory();

  const [route, setRoute] = useState<RouteState>(() => parseRoute(window.location));
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  const [rerunBusy, setRerunBusy] = useState(false);
  const [rerunDialogOpen, setRerunDialogOpen] = useState(false);
  const [ignoreBusy, setIgnoreBusy] = useState(false);
  const [testEditBusy, setTestEditBusy] = useState(false);
  const [stopBusyKey, setStopBusyKey] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | undefined>(undefined);
  const noticeTimer = useRef<number | null>(null);
  const firstSeen = useRef(Date.now());
  const [now, setNow] = useState(Date.now());

  const say = useCallback((message: string) => {
    setNotice(message);
    if (noticeTimer.current !== null) window.clearTimeout(noticeTimer.current);
    noticeTimer.current = window.setTimeout(() => setNotice(undefined), NOTICE_MS);
  }, []);
  useEffect(() => () => { if (noticeTimer.current !== null) window.clearTimeout(noticeTimer.current); }, []);

  useEffect(() => {
    if (done) return;
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, [done]);

  const startTime = runMeta?.started ? Date.parse(runMeta.started) : firstSeen.current;
  const endTime = runMeta?.ended ? Date.parse(runMeta.ended) : null;

  const commitRoute = useCallback((next: RouteState, mode: 'push' | 'replace' = 'push') => {
    setRoute(next);
    const url = buildRoute(next);
    const current = `${window.location.pathname}${window.location.search}`;
    if (url === current) return;
    if (mode === 'replace') window.history.replaceState({}, '', url);
    else window.history.pushState({}, '', url);
  }, []);

  useEffect(() => {
    const onPopState = () => setRoute(parseRoute(window.location));
    window.addEventListener('popstate', onPopState);
    return () => window.removeEventListener('popstate', onPopState);
  }, []);

  // A fixture benchmark or profile run opens on its Bench tab.
  useEffect(() => {
    if (performance && route.tab === 'tests' && window.location.pathname === `${apiBase()}/`) {
      commitRoute({ ...route, tab: 'bench', selectedPath: '' }, 'replace');
    }
  }, [performance, route, commitRoute]);

  // While linters are still running but the test phase has settled, drop the
  // "Running tests across packages" virtual task group so the only pending row
  // is "Running linters". Real parsed test results stay visible.
  const displayedTests = useMemo(
    () => (isLintOnlyPhase(tests, !!snapshotStatus.running) ? stripTestTaskGroup(tests) : tests),
    [tests, snapshotStatus.running],
  );
  const totals = useMemo(() => aggregateStatusCounts(displayedTests), [displayedTests]);
  const filteredTests = useMemo(
    () => annotateRoutePaths(collapseSingleChildChains(withExportPaths(filterTests(withOwnNames(displayedTests), route.filters.status, route.filters.framework)))),
    [displayedTests, route.filters],
  );
  const selectedTest = useMemo(
    () => (route.tab === 'tests' && route.selectedPath ? findNodeByRoutePath(filteredTests, route.selectedPath) : null),
    [route.tab, route.selectedPath, filteredTests],
  );

  const lintTree = useMemo(
    () => buildLintTree({ results: lint ?? [], grouping: route.lintGrouping, filters: route.lintFilters }),
    [lint, route.lintGrouping, route.lintFilters],
  );
  const selectedLint = route.tab === 'lint' && route.selectedPath ? findLintNode(lintTree, route.selectedPath) : null;
  const lintTotal = useMemo(() => totalLintViolations(lint ?? []), [lint]);

  // The Files tab's findings: every lint violation and failed test that names a
  // file, rooted at the work dir it ran in (the git root when it recorded none).
  const gitRoot = snapshot?.git?.root;
  const { findings, linterByFinding, sourceUrlByFinding } = useMemo(() => {
    const linterByFinding = new Map<string, string>();
    const sourceUrlByFinding = new Map<string, string>();
    const lintEntries = (lint ?? []).flatMap(result => {
      const workDir = result.work_dir || gitRoot;
      return lintFindings([{ ...result, ...(workDir ? { work_dir: workDir } : {}) }]).map(finding => {
        linterByFinding.set(findingKey(finding), result.linter);
        const url = sourceUrl(result.sources?.[finding.file]?.url, finding.line);
        if (url) sourceUrlByFinding.set(findingKey(finding), url);
        return finding;
      });
    });
    return { findings: [...lintEntries, ...testFindings(withWorkDir(tests, gitRoot))], linterByFinding, sourceUrlByFinding };
  }, [lint, tests, gitRoot]);
  const selectedFile = useMemo<FileRef | null>(
    () => (route.tab === 'files' && route.selectedPath ? { file: route.selectedPath, ...(route.workDir ? { workDir: route.workDir } : {}) } : null),
    [route.tab, route.selectedPath, route.workDir],
  );

  const traces = useMemo(() => rebaseTraces(snapshotStatus.traces, base), [snapshotStatus.traces, base]);
  const processTraceAvailable = !!processMount(snapshotStatus.traces);
  const trace = useProcessTrace({
    traces: snapshotStatus.traces,
    done,
    base,
    ...(route.tab === 'processes' && route.selectedPath ? { selectedKey: route.selectedPath } : {}),
  });
  const processCount = useMemo(() => trace.nodes.filter(node => !isProcessExited(node)).length, [trace.nodes]);
  const rootProcess = useMemo(() => trace.nodes.find(node => node.isRoot) ?? null, [trace.nodes]);
  const selectedProcess = route.tab === 'processes' ? trace.nodes.find(node => node.key === route.selectedPath) ?? null : null;

  // Route paths for step links come from the unfiltered tree so a link never
  // lands on a test the active status filter would hide.
  const processLinks = useMemo(
    () => (processTraceAvailable ? processLinksByPid(asGavel(annotateRoutePaths(collapseSingleChildChains(displayedTests)))) : new Map<number, ProcessStepLink>()),
    [processTraceAvailable, displayedTests],
  );

  useEffect(() => {
    if (route.tab !== 'processes' || route.selectedPath || !rootProcess) return;
    commitRoute({ ...route, selectedPath: rootProcess.key }, 'replace');
  }, [route, rootProcess, commitRoute]);

  // A selection the data no longer has (a stale URL, a filtered-out or deleted
  // node) is cleared once the data it would be found in has arrived.
  useEffect(() => {
    if (!route.selectedPath) return;
    const ready = route.tab === 'tests' ? tests.length > 0 || done
      : route.tab === 'lint' ? lintRun || done || !!lint
        : route.tab === 'processes' ? trace.nodes.length > 0
          : true;
    const found = route.tab === 'tests' ? selectedTest : route.tab === 'lint' ? selectedLint : route.tab === 'processes' ? selectedProcess : true;
    if (!ready || found) return;
    commitRoute({ ...route, selectedPath: route.tab === 'processes' && rootProcess ? rootProcess.key : '' }, 'replace');
  }, [route, selectedTest, selectedLint, selectedProcess, tests.length, done, lintRun, lint, trace.nodes.length, rootProcess, commitRoute]);

  const onTabChange = useCallback((tab: string) => {
    const { workDir: _workDir, line: _line, ...rest } = route;
    commitRoute({ ...rest, tab: tab as TabKey, selectedPath: '' });
  }, [route, commitRoute]);
  const onSelectTest = useCallback((node: ClickyTest | null) => commitRoute({ ...route, selectedPath: node?.route_path ?? '' }), [route, commitRoute]);
  const onFiltersChange = useCallback((filters: TestFilters) => commitRoute({ ...route, tab: 'tests', filters, selectedPath: '' }), [route, commitRoute]);
  const onLintFiltersChange = useCallback((lintFilters: LintFilters) => commitRoute({ ...route, tab: 'lint', lintFilters, selectedPath: '' }), [route, commitRoute]);
  const onLintGroupingChange = useCallback((lintGrouping: LintGrouping) => commitRoute({ ...route, tab: 'lint', lintGrouping, selectedPath: '' }), [route, commitRoute]);
  const onSelectLint = useCallback((selectedPath: string | null) => commitRoute({ ...route, selectedPath: selectedPath ?? '' }), [route, commitRoute]);
  const onSelectProcess = useCallback((key: string) => commitRoute({ ...route, tab: 'processes', selectedPath: key }), [route, commitRoute]);
  const onHideExitedChange = useCallback((hideExited: boolean) => commitRoute({ ...route, tab: 'processes', hideExited }, 'replace'), [route, commitRoute]);
  const openFile = useCallback((ref: FileRef, line?: number) => {
    const { workDir: _workDir, line: _line, ...rest } = route;
    commitRoute({ ...rest, tab: 'files', selectedPath: ref.file, ...(ref.workDir ? { workDir: ref.workDir } : {}), ...(line ? { line } : {}) });
  }, [route, commitRoute]);
  const onOpenFile = useCallback(
    (target: OpenFileTarget) => openFile({ file: target.file, ...(target.workDir || gitRoot ? { workDir: target.workDir || gitRoot } : {}) }, target.line),
    [openFile, gitRoot],
  );
  const onSelectFile = useCallback((ref: FileRef) => openFile(ref), [openFile]);
  const onLineChange = useCallback((line: number | undefined) => {
    const { line: _line, ...rest } = route;
    commitRoute({ ...rest, ...(line ? { line } : {}) }, 'replace');
  }, [route, commitRoute]);
  const onViewChange = useCallback((view: FileReviewView) => commitRoute({ ...route, view }, 'replace'), [route, commitRoute]);
  const loadFile = useCallback((ref: FileRef) => fetchFile(ref, scope), [scope]);
  const onSampleChange = useCallback((sampleIndex: number) => commitRoute({ ...route, sampleIndex }), [route, commitRoute]);
  const onOpenStep = useCallback((link: ProcessStepLink) => {
    commitRoute({ ...route, tab: 'tests', filters: { status: new Map(), framework: new Map() }, selectedPath: link.routePath });
  }, [route, commitRoute]);

  const startRerun = useCallback(async (body: RerunRequest, label: string) => {
    if (rerunBusy) return;
    setRerunBusy(true);
    setRerunDialogOpen(true);
    say(`Rerunning ${label}...`);
    try {
      const result = await run.rerun(body);
      if (result.conflict) say('Another rerun is in progress');
      else if (!result.ok) say(`Rerun failed: ${result.message ?? result.status}`);
    } catch (e: unknown) {
      say(`Rerun error: ${e instanceof Error ? e.message : String(e)}`);
    } finally {
      setRerunBusy(false);
    }
  }, [rerunBusy, run, say]);

  const onRerun = useCallback((node: ClickyTest) => {
    if (node.framework === 'task') return;
    const request = rerunBody(node as Test);
    if ('error' in request) say(request.error);
    else void startRerun(request.body, node.name);
  }, [startRerun, say]);

  const onRerunLint = useCallback(() => {
    const body = buildLintRerunBody(selectedLint, lint ?? []);
    if (!body.work_dir) say('Rerun across multiple lint roots is not supported');
    else void startRerun(body, selectedLint?.name ?? 'lint');
  }, [selectedLint, lint, startRerun, say]);

  // stop stops one task, or the whole run when no task is named.
  const stop = useCallback(async ({ taskId, label }: { taskId?: string; label: string }) => {
    if (stopBusyKey) return;
    setStopBusyKey(taskId ?? GLOBAL_STOP_KEY);
    say(`Stopping ${label}...`);
    try {
      const result = await run.stop(taskId);
      if (!result.ok) say(`Stop failed: ${result.message ?? result.status}`);
    } catch (e: unknown) {
      say(`Stop error: ${e instanceof Error ? e.message : String(e)}`);
    } finally {
      setStopBusyKey(null);
    }
  }, [stopBusyKey, run, say]);
  const onStop = useCallback((node: ClickyTest) => { if (node.task_id) void stop({ taskId: node.task_id, label: node.name }); }, [stop]);
  const onStopAll = useCallback(() => { void stop({ label: 'the run' }); }, [stop]);

  const onIgnore = useCallback(async (action: LintAction) => {
    if (ignoreBusy) return;
    setIgnoreBusy(true);
    try {
      const res = await fetch(apiUrl('/api/lint/ignore', scope), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(action.req),
      });
      if (!res.ok) {
        say(`Ignore failed: ${(await res.text()).trim()}`);
        return;
      }
      say('Ignore rule saved to .gavel.yaml');
      if (route.tab !== 'files') commitRoute({ ...route, selectedPath: '' }, 'replace');
    } catch (e: unknown) {
      say(`Ignore error: ${e instanceof Error ? e.message : String(e)}`);
    } finally {
      setIgnoreBusy(false);
    }
  }, [ignoreBusy, scope, route, commitRoute, say]);

  const renderFindingActions = useCallback((finding: FileFinding) => {
    if (finding.kind !== 'lint') return null;
    const linter = linterByFinding.get(findingKey(finding));
    if (!linter) throw new Error(`no linter recorded for the finding at ${finding.file}:${finding.line}`);
    const violation: Violation = finding.rule ? { rule: { method: finding.rule } } : {};
    const action = violationIgnoreAction({ violation, linter, file: finding.file, workDir: finding.workDir ?? '' });
    const source = sourceUrlByFinding.get(findingKey(finding));
    if (!action && !source) return null;
    return (
      <>
        {source && (
          <Button asChild size="sm" variant="outline" className="h-6 gap-1 px-2 text-xs text-muted-foreground">
            <a href={source} target="_blank" rel="noreferrer" title={`Open ${finding.file}:${finding.line ?? ''} at its source`}>Open source</a>
          </Button>
        )}
        {action && (
          <Button type="button" size="sm" variant="outline" className="h-6 gap-1 px-2 text-xs text-muted-foreground" title={action.title} disabled={ignoreBusy} onClick={() => void onIgnore(action)}>
            {action.label}
          </Button>
        )}
      </>
    );
  }, [linterByFinding, sourceUrlByFinding, ignoreBusy, onIgnore]);

  const onTestEdit = useCallback(async (node: Test, action: TestEditAction, editScope: TestEditScope) => {
    if (testEditBusy || snapshotStatus.test_edit_supported !== true) return;
    const target = editScope === 'file' ? (node.file || 'file') : (node.name || 'test');
    setTestEditBusy(true);
    say(action === 'skip' ? `Skipping ${target}...` : `Deleting ${target}...`);
    try {
      const result = await run.editTest({
        action,
        scope: editScope,
        framework: node.framework || '',
        work_dir: node.work_dir || '',
        package_path: node.package_path || '',
        file: node.file || '',
        line: node.line || 0,
        test_name: node.name || '',
        suite: node.suite || [],
      });
      if (!result.ok) {
        say(`Test edit failed: ${result.message ?? result.status}`);
        return;
      }
      say(result.message || `${action === 'skip' ? 'Skip' : 'Delete'} ${editScope === 'file' ? 'file' : 'test'} saved`);
      if (action === 'delete') commitRoute({ ...route, selectedPath: '' }, 'replace');
    } catch (e: unknown) {
      say(`Test edit error: ${e instanceof Error ? e.message : String(e)}`);
    } finally {
      setTestEditBusy(false);
    }
  }, [testEditBusy, snapshotStatus.test_edit_supported, run, route, commitRoute, say]);

  const wholeResultRoute = useMemo<RouteState>(() => ({ ...route, selectedPath: '' }), [route]);
  const actions = useMemo(() => ({
    nodeRouteState: (node: Test) => (node.export_path ? { ...route, tab: 'tests' as const, selectedPath: node.export_path } : undefined),
    failingOnlyRouteState: (node: Test) => (node.export_path
      ? { ...route, tab: 'tests' as const, selectedPath: node.export_path, filters: { ...route.filters, status: new Map([['failed', 'include' as const], ['timedout', 'include' as const]]) } }
      : undefined),
    testEditSupported: snapshotStatus.test_edit_supported === true,
    onTestEdit,
    testEditBusy,
  }), [route, snapshotStatus.test_edit_supported, onTestEdit, testEditBusy]);

  const registry = useMemo(
    () => gavelRegistry({ traces, processes: trace.nodes, onOpenProcess: onSelectProcess, actions }),
    [traces, trace.nodes, onSelectProcess, actions],
  );
  const fetchProcessMetrics = useMemo(
    () => fetchProcessMetricsFromSamples({ samples: trace.samples, nodes: trace.nodes }),
    [trace.samples, trace.nodes],
  );

  const showLintTab = lintRun;
  const showFilesTab = findings.length > 0 || route.tab === 'files';
  const showBenchTab = !!run.bench || !!performance;
  const showProcessesTab = processTraceAvailable;
  const benchRegressions = run.bench?.deltas?.filter(d => d.significant && d.delta_pct > run.bench!.threshold).length || 0;
  const canExport = (route.tab === 'tests' && displayedTests.length > 0) || (route.tab === 'lint' && lintRun) || (route.tab === 'bench' && showBenchTab);
  const canGlobalStop = snapshotStatus.running && !!snapshotStatus.stop_supported;
  const runActions = (
    <RunActions
      runMeta={runMeta}
      canStop={canGlobalStop}
      stopping={stopBusyKey === GLOBAL_STOP_KEY}
      onStop={onStopAll}
      exportState={canExport ? wholeResultRoute : undefined}
    />
  );
  const status = notice ?? streamError ?? statusText;
  const backTo = (window as unknown as { __gavelBackTo?: string }).__gavelBackTo;

  const tabs = [
    { id: 'tests', label: 'Tests', icon: UiBeaker, count: totals.failed > 0 ? totals.failed : totals.total, countColor: totals.failed > 0 ? 'bg-red-500' : 'bg-muted-foreground' },
    ...(showLintTab ? [{ id: 'lint', label: 'Lint', icon: UiListChecks, count: lintTotal, countColor: lintTotal > 0 ? 'bg-yellow-500' : 'bg-muted-foreground' }] : []),
    ...(showFilesTab ? [{ id: 'files', label: 'Files', icon: UiFileCode, count: findings.length, countColor: findings.length > 0 ? 'bg-yellow-500' : 'bg-muted-foreground' }] : []),
    ...(showBenchTab ? [{
      id: 'bench', label: 'Bench', icon: UiChartLineUp,
      count: performance ? countFixtureWork(tests) : benchRegressions > 0 ? benchRegressions : (run.bench?.deltas?.length || 0),
      countColor: benchRegressions > 0 ? 'bg-red-500' : 'bg-muted-foreground',
    }] : []),
    ...(showProcessesTab ? [{ id: 'processes', label: 'Processes', icon: UiServerProcess, count: processCount, countColor: 'bg-blue-500' }] : []),
    ...(history.available ? [{ id: 'history', label: 'History', icon: UiHistory }] : []),
  ];

  return (
    <div className="flex h-screen flex-col bg-background text-foreground">
      {backTo && (
        <div className="flex shrink-0 items-center gap-2 border-b border-border bg-card px-4 py-1.5 text-sm">
          <a href={backTo} className="inline-flex items-center gap-1 text-primary no-underline hover:underline">
            <UiArrowLeft />
            Back to PR
          </a>
        </div>
      )}
      {(tabs.length > 1 || history.available) && (
        <div className="flex shrink-0 items-center gap-2 pr-density-4">
          <Tabs tabs={tabs} value={route.tab} onChange={onTabChange} className="min-w-0 flex-1 px-density-4" />
          {history.available && <RunSwitcher history={history} navigate={navigate} />}
        </div>
      )}

      <div className="min-h-0 flex-1">
        {route.tab === 'tests' && (displayedTests.length === 0 ? (
          <div className="flex h-full flex-col">
            <PageHeader icon={UiBeaker} title="Test Results" status={status} actions={runActions} />
            <div className="min-h-0 flex-1"><Waiting done={done} /></div>
          </div>
        ) : (
          <TestRunner
            tests={filteredTests}
            allTests={displayedTests}
            selected={selectedTest}
            filters={route.filters}
            expandAll={expandAll}
            done={done}
            {...(runMeta ? { runMeta } : {})}
            status={snapshotStatus}
            statusText={filteredTests.length === 0 ? `${status} · No tests match the current filters` : status}
            now={now}
            startTime={startTime}
            endTime={endTime}
            busy={{ rerun: rerunBusy, stop: stopBusyKey !== null }}
            onSelect={onSelectTest}
            onFiltersChange={onFiltersChange}
            onExpandAllChange={setExpandAll}
            onRerun={onRerun}
            onStop={onStop}
            fetchProcessMetrics={fetchProcessMetrics}
            adapters={registry}
            onOpenFile={onOpenFile}
            headerSlot={runActions}
          />
        ))}

        {route.tab === 'lint' && (
          <div className="flex h-full flex-col">
            <LintTasks tests={tests} />
            <div className="min-h-0 flex-1">
              <LintResults
                results={lint ?? []}
                grouping={route.lintGrouping}
                onGroupingChange={onLintGroupingChange}
                filters={route.lintFilters}
                onFiltersChange={onLintFiltersChange}
                selected={selectedLint ? route.selectedPath : null}
                onSelect={onSelectLint}
                expandAll={expandAll}
                onExpandAllChange={setExpandAll}
                onIgnore={onIgnore}
                ignoreBusy={ignoreBusy}
                onOpenFile={onOpenFile}
                renderRun={renderPromptRun}
                headerSlot={(
                  <>
                    <span className="truncate text-sm text-muted-foreground">{status}</span>
                    <Button type="button" size="sm" disabled={rerunBusy} onClick={onRerunLint} title={selectedLint ? 'Rerun the linters for this selection' : 'Rerun the linters'}>
                      <UiRefresh className="mr-1" />
                      {rerunBusy ? 'Running...' : 'Rerun'}
                    </Button>
                    {runActions}
                  </>
                )}
              />
            </div>
          </div>
        )}

        {route.tab === 'files' && (
          <div className="flex h-full flex-col">
            <PageHeader icon={UiFileCode} title="Files" status={status} actions={runActions} />
            <div className="min-h-0 flex-1">
              <Suspense fallback={<FilesLoading />}>
                <FilesTab
                  findings={findings}
                  selectedFile={selectedFile}
                  onSelectFile={onSelectFile}
                  line={route.line}
                  onLineChange={onLineChange}
                  view={route.view ?? 'preview'}
                  onViewChange={onViewChange}
                  loadFile={loadFile}
                  renderFindingActions={renderFindingActions}
                />
              </Suspense>
            </div>
          </div>
        )}

        {route.tab === 'bench' && (
          <div className="flex h-full flex-col">
            <PageHeader icon={UiChartLineUp} title={performance ? 'Fixture Benchmark' : 'Benchmark Comparison'} status={status} actions={runActions} />
            <div className="min-h-0 flex-1 overflow-auto">
              {performance
                ? <FixtureBenchView performance={performance} tests={tests} error={runError} sampleIndex={route.sampleIndex} onSampleChange={onSampleChange} />
                : <BenchView bench={run.bench} />}
            </div>
          </div>
        )}

        {route.tab === 'history' && (
          <div className="flex h-full flex-col">
            <PageHeader icon={UiHistory} title="Run History" status={history.current ? `Viewing ${history.current}` : 'Viewing the latest results'} />
            <div className="flex min-h-0 flex-1 flex-col p-density-2">
              <RunHistory history={history} navigate={navigate} />
            </div>
          </div>
        )}

        {route.tab === 'processes' && (
          <div className="flex h-full flex-col">
            <PageHeader icon={UiServerProcess} title="Processes" status={status} actions={runActions} />
            <div className="min-h-0 flex-1">
              <RunProcesses
                trace={trace}
                selectedKey={selectedProcess?.key ?? null}
                onSelect={onSelectProcess}
                hideExited={!!route.hideExited}
                onHideExitedChange={onHideExitedChange}
                expandAll={expandAll}
                onExpandAllChange={setExpandAll}
                stepLinks={processLinks}
                onOpenStep={onOpenStep}
              />
            </div>
          </div>
        )}
      </div>

      <RerunDialog open={rerunDialogOpen} onClose={() => setRerunDialogOpen(false)} streamUrl={apiUrl('/api/rerun/stream', scope)} />
    </div>
  );
}



