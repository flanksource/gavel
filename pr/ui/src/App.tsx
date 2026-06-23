import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import type { PRItem, PRDetail, Snapshot, SearchConfig, RateLimit, PRSyncStatus, GavelResultsSummary, Project, ProcStatus } from './types';
import { PRList } from './components/PRList';
import { PRDetailPanel } from './components/PRDetail';
import { FilterBar, emptyFilters, type Filters } from './components/FilterBar';
import { SplitPane, AppShell } from '@flanksource/clicky-ui/components';
import { ActivityView } from './components/ActivityView';
import { TodoBodyHeader, TodoBodyActions, TodoFilterToolbar, TodoWorkspaceList, TodoDetailPane } from './components/TodoView';
import { useWorkspaceTodos } from './components/todos/useWorkspaceTodos';
import { CreateTodoDialog } from './components/todos/CreateTodoDialog';
import { TodoNewPage } from './components/todos/TodoNewPage';
import { MenubarTodos } from './components/MenubarTodos';
import { StatusIndicator } from './components/StatusIndicator';
import { OrgChooser } from './components/OrgChooser';
import { AddProjectDialog } from './components/AddProjectDialog';
import { ProjectsBar } from './components/ProjectsBar';
import { ProcessManager } from './components/ProcessManager';
import { ThemeToggle } from './components/ThemeToggle';
import { WorkspaceGroup } from './components/ProcessTable';
import { aggregateDotClass, computeCounts, collectRepos, collectAuthors, filterPRs, flattenProcesses, prKey, emptyProcStatus } from './utils';
import { useCopyFeedback } from './useCopyFeedback';
import {
  annotateRoutePaths,
  buildRoute,
  findPRByRoutePath,
  parseRoute,
  type RouteState,
  type Tab,
} from './routes';
import { copyCurrentViewForAgent, downloadCurrentView } from './export';
import { loadUIState, saveUIState, filtersFromStored } from './storage';
import { useDocumentVisible } from './useDocumentVisible';
import { useIsMobile } from './useIsMobile';
import { GavelIcon } from './components/GavelIcon';

const defaultConfig: SearchConfig = { repos: [] };

type WebKitExternalBridge = {
  webkit?: {
    messageHandlers?: {
      external?: {
        postMessage: (message: string) => void;
      };
    };
  };
};

const menubarOpenExternalMessage = 'gavel:open-external';
const menubarPointerEnterMessage = 'gavel:pointer-enter';
const menubarPointerLeaveMessage = 'gavel:pointer-leave';

function postMenubarMessage(type: string, payload: Record<string, unknown> = {}) {
  const bridge = (window as WebKitExternalBridge).webkit?.messageHandlers?.external;
  if (!bridge) return false;
  bridge.postMessage(JSON.stringify({ type, ...payload }));
  return true;
}

function useMenubarExternalLinks() {
  useEffect(() => {
    const onClick = (event: MouseEvent) => {
      if (event.defaultPrevented || event.button !== 0) return;
      const target = event.target;
      if (!(target instanceof Element)) return;
      const anchor = target.closest('a[href]');
      if (!(anchor instanceof HTMLAnchorElement) || !anchor.href) return;

      if (!postMenubarMessage(menubarOpenExternalMessage, { url: anchor.href })) return;

      event.preventDefault();
      event.stopPropagation();
      event.stopImmediatePropagation();
    };

    document.addEventListener('click', onClick, true);
    return () => document.removeEventListener('click', onClick, true);
  }, []);
}

// aggregateShards rolls per-shard summaries into one badge for the sidebar,
// where there is only room for a single number per PR. Returns null if the
// list is empty.
function aggregateShards(shards: GavelResultsSummary[]): GavelResultsSummary | null {
  if (!shards || shards.length === 0) return null;
  if (shards.length === 1) return shards[0];
  const agg: GavelResultsSummary = {
    artifactId: 0,
    artifactUrl: '',
    testsPassed: 0,
    testsFailed: 0,
    testsSkipped: 0,
    testsTotal: 0,
    lintViolations: 0,
    lintLinters: 0,
    hasBench: false,
    benchRegressions: 0,
    topFailures: [],
    topLintViolations: [],
  };
  for (const s of shards) {
    agg.testsPassed += s.testsPassed;
    agg.testsFailed += s.testsFailed;
    agg.testsSkipped += s.testsSkipped;
    agg.testsTotal += s.testsTotal;
    agg.lintViolations += s.lintViolations;
    agg.lintLinters += s.lintLinters;
    agg.benchRegressions = (agg.benchRegressions ?? 0) + (s.benchRegressions ?? 0);
    if (s.hasBench) agg.hasBench = true;
    for (const f of s.topFailures ?? []) {
      if ((agg.topFailures?.length ?? 0) >= 5) break;
      agg.topFailures!.push(f);
    }
    for (const v of s.topLintViolations ?? []) {
      if ((agg.topLintViolations?.length ?? 0) >= 5) break;
      agg.topLintViolations!.push(v);
    }
  }
  return agg;
}

export function App() {
  const initialRoute: RouteState = typeof window !== 'undefined'
    ? parseRoute(window.location)
    : { tab: 'prs', selectedPath: '', filters: emptyFilters() };

  // Hydrate org/search config and filters from localStorage. URL query params
  // (if present) win for filters so deep links still work.
  const stored = typeof window !== 'undefined' ? loadUIState() : {};
  const hasUrlFilters = typeof window !== 'undefined' && window.location.search.length > 1;
  // First-run default hides bots (the daemon now fetches them so the @bots
  // author chip can toggle them back on). URL params and stored filters win.
  const defaultFilters: Filters = { ...emptyFilters(), authors: { '@bots': 'exclude' } };
  const initialFilters = hasUrlFilters ? initialRoute.filters : (filtersFromStored(stored.filters) ?? defaultFilters);
  const initialConfig: SearchConfig = { ...defaultConfig, ...(stored.config || {}) };

  const [rawPrs, setRawPrs] = useState<PRItem[]>([]);
  const [viewer, setViewer] = useState('');
  const [botsAvailable, setBotsAvailable] = useState(false);
  const [includeBotsServer, setIncludeBotsServer] = useState(false);
  const [unread, setUnread] = useState<Record<string, boolean>>({});
  const [fetchedAt, setFetchedAt] = useState('');
  const [nextFetchIn, setNextFetchIn] = useState(60);
  const [error, setError] = useState<string | undefined>();
  const [selected, setSelected] = useState<PRItem | null>(null);
  const [detail, setDetail] = useState<PRDetail | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [filters, setFilters] = useState<Filters>(initialFilters);
  const [selectedPath, setSelectedPath] = useState(initialRoute.selectedPath);
  const [config, setConfig] = useState<SearchConfig>(initialConfig);
  const [paused, setPaused] = useState(false);
  const [rateLimit, setRateLimit] = useState<RateLimit | undefined>();
  const [activeTab, setActiveTab] = useState<Tab>(initialRoute.tab);
  const [query, setQuery] = useState('');
  const { copyState, copyError, beginCopy, resetCopyFeedback } = useCopyFeedback({ copiedMs: 2500, errorMs: 2500 });
  const [syncStatus, setSyncStatus] = useState<Record<string, PRSyncStatus>>({});
  const [gavelResultsMap, setGavelResultsMap] = useState<Record<string, GavelResultsSummary>>({});
  const [projects, setProjects] = useState<Project[]>([]);
  const [procStatus, setProcStatus] = useState<Record<string, ProcStatus>>({});
  const [addOpen, setAddOpen] = useState(false);
  const [editProject, setEditProject] = useState<Project | null>(null);
  const visible = useDocumentVisible();
  const isMobile = useIsMobile();

  const pathname = typeof window !== 'undefined' ? window.location.pathname : '';
  const isMenubar = pathname === '/menubar';
  const isProcessesPage = pathname === '/processes';
  // The focused /todos/new form is a standalone page (like /menubar, /processes)
  // rather than a view inside the Todos tab, so it can be linked to and returns
  // to its referer. parseRoute reads it as the todos tab with selectedPath="new";
  // the early return below renders the form before any of that matters.
  const isTodoNewPage = pathname === '/todos/new';
  // On mobile the desktop AppShell (split panes, sidebar, toolbars) has no room,
  // so the dashboard falls back to the same compact, single-column menubar
  // dropdown layout the native menubar webview uses. The standalone /processes
  // and /todos/new pages already lay out fine on narrow screens, so they keep
  // their own views.
  const useMenubarLayout = isMenubar || (isMobile && !isProcessesPage && !isTodoNewPage);

  const prs = useMemo(() => annotateRoutePaths(rawPrs), [rawPrs]);

  const routeState: RouteState = useMemo(
    () => ({ tab: activeTab, selectedPath, filters }),
    [activeTab, selectedPath, filters],
  );

  const commitRoute = useCallback((next: RouteState, mode: 'push' | 'replace' = 'push') => {
    setActiveTab(next.tab);
    setSelectedPath(next.selectedPath);
    setFilters(next.filters);
    const url = buildRoute(next);
    const current = `${window.location.pathname}${window.location.search}`;
    if (url !== current) {
      if (mode === 'replace') window.history.replaceState({}, '', url);
      else window.history.pushState({}, '', url);
    }
  }, []);

  // Switching the top-level tab navigates (so /todos, /activity are linkable and
  // back/forward works); the PR selection is dropped when leaving the prs tab.
  const changeTab = useCallback((next: Tab) => {
    commitRoute({ tab: next, selectedPath: '', filters });
  }, [commitRoute, filters]);

  // Selecting a todo encodes its ref in the path (/todos/{guid}) so a todo is
  // deep-linkable and back/forward works, mirroring PR selection. An empty id
  // clears the selection back to /todos.
  const navigateTodo = useCallback((id: string) => {
    commitRoute({ tab: 'todos', selectedPath: id, filters });
  }, [commitRoute, filters]);

  // The Todos data layer is mounted permanently so its chrome can live in the
  // AppShell's body slots, but only fetches while the Todos tab is active. The
  // selectedPath is the todo ref on that tab (a PR route path otherwise).
  const onTodosTab = activeTab === 'todos' && !isTodoNewPage;
  const todos = useWorkspaceTodos(projects, onTodosTab ? selectedPath : '', navigateTodo, onTodosTab);

  useEffect(() => {
    const onPopState = () => {
      const next = parseRoute(window.location);
      setActiveTab(next.tab);
      setSelectedPath(next.selectedPath);
      setFilters(next.filters);
    };
    window.addEventListener('popstate', onPopState);
    return () => window.removeEventListener('popstate', onPopState);
  }, []);

  // Pull the PR snapshot and keep it live over SSE — but only while visible. The
  // menubar webview stays resident when dismissed; holding the stream open (and
  // re-rendering on every push) there is pure waste, so close it on hide and
  // reopen — refetching fresh — on show.
  useEffect(() => {
    if (!visible) return;
    fetch('/api/prs')
      .then(r => r.json())
      .then((snap: Snapshot) => applySnapshot(snap))
      .catch(() => {});

    const es = new EventSource('/api/prs/stream');
    es.addEventListener('message', (e: MessageEvent) => {
      applySnapshot(JSON.parse(e.data));
    });
    es.onerror = () => setError('Connection lost — retrying...');

    return () => { es.close(); };
  }, [visible]);

  // Relative timestamps refresh themselves: each <RelativeTime/> (and the process
  // uptime / sync countdown leaves) subscribes to the shared useNow() clock, so
  // the per-second tick re-renders only those leaves instead of forcing a
  // full-app reconcile here. The clock is visibility-gated in useNow itself.
  function applySnapshot(snap: Snapshot) {
    setRawPrs(snap.prs || []);
    if (snap.viewer) setViewer(snap.viewer);
    setBotsAvailable(!!snap.botsAvailable);
    setIncludeBotsServer(!!snap.includeBots);
    setUnread(snap.unread || {});
    setFetchedAt(snap.fetchedAt);
    setNextFetchIn(snap.nextFetchIn);
    setError(snap.error);
    setPaused(snap.paused);
    if (snap.rateLimit) setRateLimit(snap.rateLimit);
    // The server doesn't persist org/all (only repos/author/etc), so after a
    // server restart its config carries empty org/all — keep the locally
    // restored values in that case rather than clobbering them.
    if (snap.config) setConfig(prev => ({ ...snap.config, org: snap.config.org || prev.org, all: snap.config.all || prev.all }));
    if (snap.syncStatus) setSyncStatus(snap.syncStatus);
    if (snap.gavelResults) {
      setGavelResultsMap(prev => ({ ...prev, ...snap.gavelResults }));
    }
  }

  // When PRs arrive (or the URL selection changes), reconcile `selected` with
  // the route's selectedPath. Fetches detail automatically for deep-linked PRs.
  useEffect(() => {
    // The menubar layout (native webview or mobile) drives PR selection through
    // its own local state, not the route, so skip route→selection reconciliation
    // there and let onSelect/onBack own it.
    if (useMenubarLayout || isProcessesPage) return;
    if (!selectedPath) {
      if (selected) { setSelected(null); setDetail(null); }
      return;
    }
    if (selected && selected.route_path === selectedPath) return;
    const target = findPRByRoutePath(prs, selectedPath);
    if (target) {
      loadPR(target);
    }
  }, [selectedPath, prs]);

  const fetchProjects = useCallback(() => {
    fetch('/api/projects')
      .then(r => r.json())
      .then((ps: Project[]) => setProjects(ps || []))
      .catch(() => {});
  }, []);

  const fetchProcStatus = useCallback(() => {
    fetch('/api/proc/status')
      .then(r => r.json())
      .then((m: Record<string, ProcStatus>) => setProcStatus(m || {}))
      .catch(() => {});
  }, []);

  const onProcChanged = useCallback(() => {
    fetchProjects();
    fetchProcStatus();
  }, [fetchProjects, fetchProcStatus]);

  useEffect(() => { fetchProjects(); }, [fetchProjects]);

  // Stream process status over SSE while projects are configured and the tab is
  // visible — this is the sole driver of the sidebar's per-repo proc badges and
  // is intentionally decoupled from the GitHub PR poller. The server pushes a
  // fresh frame on connect and then drives the cadence (faster while a process
  // is starting/restarting), replacing the old client-side poll. Closing on hide
  // lets the backend stop sampling; reopening on show pushes an immediate frame.
  useEffect(() => {
    if (projects.length === 0) { setProcStatus({}); return; }
    if (!visible) return;
    const es = new EventSource('/api/proc/status/stream');
    es.addEventListener('message', (e: MessageEvent) => {
      try { setProcStatus(JSON.parse(e.data)); } catch { /* ignore malformed frame */ }
    });
    // EventSource auto-reconnects; proc status is best-effort, so swallow errors.
    return () => { es.close(); };
  }, [projects.length, visible]);

  const projectsByRepo = useMemo(() => {
    const m: Record<string, Project> = {};
    for (const p of projects) for (const r of p.repos || []) m[r] = p;
    return m;
  }, [projects]);

  const openAdd = useCallback(() => { setEditProject(null); setAddOpen(true); }, []);
  const openEdit = useCallback((p: Project) => { setEditProject(p); setAddOpen(true); }, []);

  useEffect(() => { saveUIState(config, filters); }, [config, filters]);

  const markSeen = useCallback((pr: PRItem) => {
    const key = prKey(pr);
    setUnread(prev => {
      if (!prev[key]) return prev;
      const next = { ...prev };
      delete next[key];
      return next;
    });
    fetch('/api/prs/seen', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ repo: pr.repo, number: pr.number }),
    }).catch(() => {});
  }, []);

  const detailESRef = useRef<EventSource | null>(null);

  function loadPR(pr: PRItem) {
    const isNewPR = !selected || selected.repo !== pr.repo || selected.number !== pr.number;
    setSelected(pr);
    // Only clear detail when switching to a different PR; keep stale data during refresh
    if (isNewPR) setDetail(null);
    setDetailLoading(true);
    markSeen(pr);

    // Close any previous detail stream
    if (detailESRef.current) {
      detailESRef.current.close();
      detailESRef.current = null;
    }

    const url = `/api/prs/detail?repo=${encodeURIComponent(pr.repo)}&number=${pr.number}`;
    const es = new EventSource(url);
    detailESRef.current = es;

    es.addEventListener('pr', (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      setDetail(prev => ({ ...prev, pr: data.pr, comments: data.comments }));
      setDetailLoading(false);
    });

    es.addEventListener('runs', (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      setDetail(prev => prev ? { ...prev, runs: data.runs } : prev);
    });

    es.addEventListener('gavel', (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      const shards: GavelResultsSummary[] = data.gavelResults ?? [];
      setDetail(prev => prev ? { ...prev, gavelResults: shards } : prev);
      const agg = aggregateShards(shards);
      if (agg) {
        setGavelResultsMap(prev => ({ ...prev, [prKey(pr)]: agg }));
      }
    });

    es.addEventListener('error', (e: MessageEvent) => {
      if (e.data) {
        const data = JSON.parse(e.data);
        setDetail(prev => prev ?? { error: data.error });
      }
      setDetailLoading(false);
    });

    es.addEventListener('done', () => {
      setDetailLoading(false);
      es.close();
      detailESRef.current = null;
    });

    es.onerror = () => {
      if (!detailESRef.current) return; // already closed normally
      setDetail(prev => prev ?? { error: 'Connection lost' });
      setDetailLoading(false);
      es.close();
      detailESRef.current = null;
    };
  }

  function handleSelect(pr: PRItem) {
    commitRoute({ ...routeState, selectedPath: pr.route_path || `${pr.repo}/${pr.number}` });
    loadPR(pr);
  }

  function clearSelectedPR() {
    if (detailESRef.current) {
      detailESRef.current.close();
      detailESRef.current = null;
    }
    setSelected(null);
    setDetail(null);
    setDetailLoading(false);
  }

  function handleFiltersChange(next: Filters) {
    commitRoute({ ...routeState, filters: next });
  }

  function handleRefresh() {
    fetch('/api/prs/refresh', { method: 'POST' }).catch(() => {});
  }

  function handlePause() {
    fetch('/api/prs/pause', { method: 'POST' }).catch(() => {});
  }

  function updateConfig(partial: Partial<SearchConfig>) {
    const next = { ...config, ...partial };
    setConfig(next);
    fetch('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(next),
    }).catch(() => {});
  }

  const onDownloadJSON = useCallback(() => downloadCurrentView(routeState, 'json'), [routeState]);
  const onDownloadMarkdown = useCallback(() => downloadCurrentView(routeState, 'md'), [routeState]);
  const onCopyForAgent = useCallback(async () => {
    if (copyState === 'copying') return;
    beginCopy();
    try {
      await copyCurrentViewForAgent(routeState);
      resetCopyFeedback('copied');
    } catch (e: any) {
      resetCopyFeedback('error', e?.message || 'Copy failed');
    }
  }, [copyState, routeState, beginCopy, resetCopyFeedback]);

  const counts = useMemo(() => computeCounts(prs), [prs]);
  const reposList = useMemo(() => collectRepos(prs), [prs]);
  const authors = useMemo(() => collectAuthors(prs, viewer, botsAvailable), [prs, viewer, botsAvailable]);

  // The @bots chip drives whether the daemon fetches bot PRs at all: when it's
  // not excluding bots, ask the server to include them (and refetch). Excluding
  // bots lets the server drop them at the source. Converges since the snapshot
  // echoes the server's includeBots back into includeBotsServer.
  useEffect(() => {
    const wantBots = (filters.authors['@bots'] ?? '') !== 'exclude';
    if (wantBots === includeBotsServer) return;
    fetch('/api/prs/bots', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ include: wantBots }),
    }).catch(() => {});
  }, [filters.authors, includeBotsServer]);
  const filtered = useMemo(
    () => filterPRs(prs, filters.state, filters.checks, filters.repos, filters.authors, viewer),
    [prs, filters, viewer],
  );
  // Free-text search over title / branches / #number / repo, applied on top of
  // the structured (tri-state) facet filters.
  const searched = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return filtered;
    return filtered.filter(pr =>
      pr.title.toLowerCase().includes(q) ||
      String(pr.number).includes(q) ||
      pr.repo.toLowerCase().includes(q) ||
      (pr.source || '').toLowerCase().includes(q) ||
      (pr.target || '').toLowerCase().includes(q));
  }, [filtered, query]);

  if (useMenubarLayout) {
    return (
      <MenubarView
        prs={searched}
        selected={selected}
        detail={detail}
        detailLoading={detailLoading}
        unread={unread}
        projects={projects}
        projectsByRepo={projectsByRepo}
        procStatus={procStatus}
        syncStatus={syncStatus}
        gavelResults={gavelResultsMap}
        onSelect={loadPR}
        onBack={clearSelectedPR}
        onProcChanged={onProcChanged}
        fetchedAt={fetchedAt}
        error={error}
      />
    );
  }

  if (isProcessesPage) {
    return (
      <ProcessesPage
        projects={projects}
        procStatus={procStatus}
        onProcChanged={onProcChanged}
      />
    );
  }

  if (isTodoNewPage) {
    return <TodoNewPage projects={projects} />;
  }

  return (
    <>
      <AppShell
        brand={<img src="/brand/gavel-logo.svg" alt="gavel" className="h-7" />}
        nav={<TabBar active={activeTab} onChange={changeTab} />}
        search={
          activeTab === 'prs' ? (
            <div className="flex w-full items-center gap-2 rounded-md border border-border bg-muted px-3 py-1.5">
              <GavelIcon name="codicon:search" className="text-muted-foreground text-sm shrink-0" />
              <input
                value={query}
                onChange={(e) => setQuery((e.target as HTMLInputElement).value)}
                placeholder="Search pull requests, branches, #id…"
                aria-label="Search pull requests"
                className="flex-1 min-w-0 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
              />
              {query && (
                <button
                  type="button"
                  onClick={() => setQuery('')}
                  className="text-muted-foreground hover:text-foreground shrink-0"
                  aria-label="Clear search"
                >
                  <GavelIcon name="codicon:close" className="text-xs" />
                </button>
              )}
            </div>
          ) : undefined
        }
        actions={
          <>
            <ProcessManager projects={projects} procStatus={procStatus} onProcChanged={onProcChanged} />
            <OrgChooser config={config} onChange={updateConfig} />
            <StatusIndicator
              fetchedAt={fetchedAt}
              error={error}
              nextFetchIn={nextFetchIn}
              paused={paused}
              rateLimit={rateLimit}
              onRefresh={handleRefresh}
              onPause={handlePause}
              networkBusy={detailLoading}
            />
            <ThemeToggle />
          </>
        }
        toolbar={
          activeTab === 'prs' ? (
            <FilterBar filters={filters} onChange={handleFiltersChange} counts={counts} repos={reposList} authors={authors} />
          ) : activeTab === 'todos' && todos.aggregate.total > 0 ? (
            <TodoFilterToolbar todos={todos} />
          ) : undefined
        }
        bodyHeader={
          activeTab === 'prs' ? (
            <span className="text-xs text-muted-foreground">
              {searched.length} pull request{searched.length !== 1 ? 's' : ''}
            </span>
          ) : activeTab === 'todos' ? (
            <TodoBodyHeader todos={todos} />
          ) : undefined
        }
        bodyActions={
          activeTab === 'prs' ? (
            <ExportButtons
              onJSON={onDownloadJSON}
              onMarkdown={onDownloadMarkdown}
              onCopy={onCopyForAgent}
              copyState={copyState}
              copyError={copyError}
            />
          ) : activeTab === 'todos' ? (
            <TodoBodyActions todos={todos} />
          ) : undefined
        }
        bodySidebar={activeTab === 'todos' ? <TodoWorkspaceList todos={todos} /> : undefined}
        bodySplit={38}
        contentClassName="overflow-hidden"
      >
        {activeTab === 'prs' ? (
          <SplitPane
            left={
              <>
                <ProjectsBar projects={projects} procStatus={procStatus} onChanged={onProcChanged} onEdit={openEdit} onAdd={openAdd} />
                <PRList prs={searched} selected={selected} onSelect={handleSelect} unread={unread} syncStatus={syncStatus} gavelResults={gavelResultsMap} projectsByRepo={projectsByRepo} procStatus={procStatus} onProcChanged={onProcChanged} onProcEdit={openEdit} />
              </>
            }
            right={
              selected ? (
                <PRDetailPanel pr={selected} detail={detail} loading={detailLoading} />
              ) : (
                <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
                  <div className="text-center">
                    <GavelIcon name="codicon:git-pull-request" className="text-4xl mb-2" />
                    <p>Select a PR to view details</p>
                  </div>
                </div>
              )
            }
          />
        ) : activeTab === 'todos' ? (
          <TodoDetailPane todos={todos} />
        ) : (
          <ActivityView />
        )}
      </AppShell>

      <CreateTodoDialog open={todos.showCreate} onClose={() => todos.setShowCreate(false)} workspaces={todos.workspaces} onCreated={todos.created} />
      <AddProjectDialog open={addOpen} onClose={() => setAddOpen(false)} onSaved={onProcChanged} repoOptions={reposList} edit={editProject} />
    </>
  );
}

function ProcessesPage({
  projects,
  procStatus,
  onProcChanged,
}: {
  projects: Project[];
  procStatus: Record<string, ProcStatus>;
  onProcChanged: () => void;
}) {
  const workspaces = useMemo(
    () => projects.map(p => ({ project: p, status: procStatus[p.name] ?? emptyProcStatus })),
    [projects, procStatus],
  );
  const procs = useMemo(() => flattenProcesses(projects, procStatus), [projects, procStatus]);
  const running = procs.filter(p => p.proc.status === 'running').length;
  const dot = aggregateDotClass(procs.map(p => p.proc));

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <div className="flex min-w-0 items-center gap-3">
          <img src="/brand/gavel-logo.svg" alt="gavel" className="h-7 shrink-0" />
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <span className={`inline-block h-2.5 w-2.5 rounded-full ${dot}`} />
              <span className="text-sm font-semibold">Processes</span>
            </div>
            <div className="text-xs text-muted-foreground">{running} running of {procs.length}</div>
          </div>
        </div>
        <a
          href="/prs"
          className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          title="Back to PR dashboard"
          aria-label="Back to PR dashboard"
        >
          <GavelIcon name="codicon:close" className="text-base" />
        </a>
      </header>

      <main className="mx-auto w-full max-w-6xl px-4 py-4">
        {workspaces.length > 0 ? (
          <div className="divide-y divide-border border-y border-border">
            {workspaces.map(w => (
              <WorkspaceGroup key={w.project.name} project={w.project} status={w.status} onChanged={onProcChanged} />
            ))}
          </div>
        ) : (
          <div className="py-10 text-center text-sm text-muted-foreground">No projects configured</div>
        )}
      </main>
    </div>
  );
}

function MenubarView({
  prs,
  selected,
  detail,
  detailLoading,
  unread,
  projects,
  projectsByRepo,
  procStatus,
  syncStatus,
  gavelResults,
  onSelect,
  onBack,
  onProcChanged,
  fetchedAt,
  error,
}: {
  prs: PRItem[];
  selected: PRItem | null;
  detail: PRDetail | null;
  detailLoading: boolean;
  unread: Record<string, boolean>;
  projects: Project[];
  projectsByRepo: Record<string, Project>;
  procStatus: Record<string, ProcStatus>;
  syncStatus: Record<string, PRSyncStatus>;
  gavelResults: Record<string, GavelResultsSummary>;
  onSelect: (pr: PRItem) => void;
  onBack: () => void;
  onProcChanged: () => void;
  fetchedAt: string;
  error?: string;
}) {
  useMenubarExternalLinks();
  const [menubarTab, setMenubarTab] = useState<'processes' | 'prs' | 'todos'>('prs');

  const workspaces = useMemo(
    () => projects.map(p => ({ project: p, status: procStatus[p.name] ?? emptyProcStatus })),
    [projects, procStatus],
  );
  const procs = useMemo(() => flattenProcesses(projects, procStatus), [projects, procStatus]);
  const running = procs.filter(p => p.proc.status === 'running').length;
  const dot = aggregateDotClass(procs.map(p => p.proc));
  const failed = prs.filter(pr => pr.checkStatus?.failed && pr.checkStatus.failed > 0).length;
  // Tab count/badge come from the projects poll's per-workspace todoCounts, so
  // the strip reflects todos without fetching the lists until the tab is opened.
  const openTodos = projects.reduce((n, p) => n + (p.todoCounts?.open ?? 0), 0);
  const failedTodos = projects.reduce((n, p) => n + (p.todoCounts?.failed ?? 0), 0);
  const unreadCount = prs.filter(pr => unread[prKey(pr)]).length;
  const fetched = fetchedAt ? new Date(fetchedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '';

  if (selected) {
    return (
      <div
        className="h-screen overflow-hidden bg-background text-foreground"
        onPointerEnter={() => postMenubarMessage(menubarPointerEnterMessage)}
        onPointerLeave={() => postMenubarMessage(menubarPointerLeaveMessage)}
      >
        <div className="flex h-11 items-center justify-between border-b border-border px-2">
          <div className="flex min-w-0 items-center gap-2">
            <button
              type="button"
              onClick={onBack}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
              title="Back to pull requests"
              aria-label="Back to pull requests"
            >
              <GavelIcon name="codicon:arrow-left" className="text-base" />
            </button>
            <div className="min-w-0">
              <div className="truncate text-sm font-semibold">{selected.repo}#{selected.number}</div>
              <div className="truncate text-[11px] text-muted-foreground">{selected.title}</div>
            </div>
          </div>
          <div className="shrink-0 text-[11px] text-muted-foreground tabular-nums">{error || fetched}</div>
        </div>
        <div className="h-[calc(100vh-44px)] overflow-hidden">
          <PRDetailPanel pr={selected} detail={detail} loading={detailLoading} />
        </div>
      </div>
    );
  }

  return (
    <div
      className="flex h-screen flex-col overflow-hidden bg-background text-foreground"
      onPointerEnter={() => postMenubarMessage(menubarPointerEnterMessage)}
      onPointerLeave={() => postMenubarMessage(menubarPointerLeaveMessage)}
    >
      <div className="flex shrink-0 items-center justify-between border-b border-border px-3 py-2">
        <div className="flex items-center gap-2 min-w-0">
          <span className={`inline-block h-2.5 w-2.5 rounded-full ${dot}`} />
          <div className="min-w-0">
            <div className="text-sm font-semibold leading-tight">Gavel</div>
            <div className="truncate text-[11px] text-muted-foreground">
              {running}/{procs.length} processes
              {failed > 0 ? ` · ${failed} failing` : ''}
              {unreadCount > 0 ? ` · ${unreadCount} unread` : ''}
            </div>
          </div>
        </div>
        <div className="text-[11px] text-muted-foreground tabular-nums">{error || fetched}</div>
      </div>

      <div className="flex shrink-0 items-center gap-1 border-b border-border px-2 py-1">
        <MenubarTab
          label="Processes"
          dot={dot}
          count={`${running}/${procs.length}`}
          active={menubarTab === 'processes'}
          onClick={() => setMenubarTab('processes')}
        />
        <MenubarTab
          label="PRs"
          icon="codicon:git-pull-request"
          count={prs.length}
          badge={failed > 0 ? failed : undefined}
          active={menubarTab === 'prs'}
          onClick={() => setMenubarTab('prs')}
        />
        <MenubarTab
          label="Todos"
          icon="codicon:check"
          count={openTodos}
          badge={failedTodos > 0 ? failedTodos : undefined}
          active={menubarTab === 'todos'}
          onClick={() => setMenubarTab('todos')}
        />
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto">
        {menubarTab === 'processes' ? (
          workspaces.length > 0 ? (
            <div className="divide-y divide-border p-2">
              {workspaces.map(w => (
                <WorkspaceGroup key={w.project.name} project={w.project} status={w.status} onChanged={onProcChanged} />
              ))}
            </div>
          ) : (
            <div className="px-3 py-6 text-center text-xs text-muted-foreground">No projects configured</div>
          )
        ) : menubarTab === 'todos' ? (
          <MenubarTodos projects={projects} />
        ) : (
          <PRList
            prs={prs}
            selected={selected}
            onSelect={onSelect}
            unread={unread}
            syncStatus={syncStatus}
            gavelResults={gavelResults}
            projectsByRepo={projectsByRepo}
            procStatus={procStatus}
            onProcChanged={onProcChanged}
          />
        )}
      </div>
    </div>
  );
}

// MenubarTab is one segment of the menubar's Processes/PRs switcher. A status
// dot (process health) or an icon leads the label; count is the inline subtotal
// and badge is an attention-grabbing count (e.g. failing PRs).
function MenubarTab({ label, icon, dot, count, badge, active, onClick }: {
  label: string;
  icon?: string;
  dot?: string;
  count?: number | string;
  badge?: number;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex items-center gap-1.5 rounded-md px-2 py-1 text-xs transition ${
        active ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted'
      }`}
    >
      {dot ? <span className={`inline-block h-2 w-2 rounded-full ${dot}`} /> : icon ? <GavelIcon name={icon} /> : null}
      <span>{label}</span>
      {count !== undefined && count !== '' && (
        <span className="tabular-nums text-[10px] text-muted-foreground">{count}</span>
      )}
      {badge !== undefined && badge > 0 && (
        <span className="rounded-full bg-red-500/15 px-1 text-[10px] font-medium tabular-nums text-red-600 dark:text-red-400">{badge}</span>
      )}
    </button>
  );
}

function TabBar({ active, onChange }: { active: Tab; onChange: (t: Tab) => void }) {
  const tabs: { id: Tab; label: string; icon: string }[] = [
    { id: 'prs', label: 'PRs', icon: 'codicon:git-pull-request' },
    { id: 'todos', label: 'Todos', icon: 'codicon:check' },
    { id: 'activity', label: 'Activity', icon: 'codicon:pulse' },
  ];
  return (
    <div className="flex gap-1 border-b border-transparent">
      {tabs.map(t => (
        <button
          key={t.id}
          onClick={() => onChange(t.id)}
          className={`px-3 py-1.5 text-sm rounded-md transition ${
            active === t.id
              ? 'bg-primary/10 text-primary font-medium'
              : 'text-muted-foreground hover:bg-muted'
          }`}
        >
          <GavelIcon name={t.icon} className="mr-1" />
          {t.label}
        </button>
      ))}
    </div>
  );
}

function ExportButtons({ onJSON, onMarkdown, onCopy, copyState, copyError }: {
  onJSON: () => void;
  onMarkdown: () => void;
  onCopy: () => void;
  copyState: 'idle' | 'copying' | 'copied' | 'error';
  copyError: string;
}) {
  return (
    <div className="flex items-center gap-1">
      <button
        className="text-xs px-2 py-1 rounded border border-border text-muted-foreground hover:bg-muted transition-colors"
        onClick={onJSON}
        title="Download current view as JSON"
      >
        <GavelIcon name="codicon:json" className="mr-0.5" />
        JSON
      </button>
      <button
        className="text-xs px-2 py-1 rounded border border-border text-muted-foreground hover:bg-muted transition-colors"
        onClick={onMarkdown}
        title="Download current view as Markdown"
      >
        <GavelIcon name="codicon:markdown" className="mr-0.5" />
        Markdown
      </button>
      <button
        className={`text-xs px-2 py-1 rounded border transition-colors ${
          copyState === 'copied'
            ? 'border-green-300 bg-green-50 text-green-700'
            : copyState === 'error'
              ? 'border-red-300 bg-red-50 text-red-700'
              : 'border-gray-300 text-gray-600 hover:bg-gray-200'
        }`}
        onClick={onCopy}
        title={copyError || 'Copy Markdown export for agent'}
      >
        <GavelIcon
          name={copyState === 'copied' ? 'codicon:check' : copyState === 'copying' ? 'svg-spinners:ring-resize' : 'codicon:copy'}
          className="mr-0.5"
        />
        {copyState === 'copying' ? 'Copying...' : copyState === 'copied' ? 'Copied' : copyState === 'error' ? 'Copy failed' : 'Copy for Agent'}
      </button>
    </div>
  );
}
