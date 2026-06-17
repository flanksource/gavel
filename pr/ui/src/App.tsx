import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import type { PRItem, PRDetail, Snapshot, SearchConfig, RateLimit, PRSyncStatus, GavelResultsSummary, Project, ProcStatus } from './types';
import { PRList } from './components/PRList';
import { PRDetailPanel } from './components/PRDetail';
import { FilterBar, emptyFilters, type Filters } from './components/FilterBar';
import { SplitPane, AppShell } from '@flanksource/clicky-ui/components';
import { ActivityView } from './components/ActivityView';
import { StatusIndicator } from './components/StatusIndicator';
import { OrgChooser } from './components/OrgChooser';
import { AddProjectDialog } from './components/AddProjectDialog';
import { ProjectsBar } from './components/ProjectsBar';
import { ProcessManager } from './components/ProcessManager';
import { ThemeToggle } from './components/ThemeToggle';
import { computeCounts, collectRepos, collectAuthors, filterPRs, prKey } from './utils';
import {
  annotateRoutePaths,
  buildRoute,
  findPRByRoutePath,
  parseRoute,
  type RouteState,
} from './routes';
import { copyCurrentViewForAgent, downloadCurrentView } from './export';
import { loadUIState, saveUIState, filtersFromStored } from './storage';

type Tab = 'prs' | 'activity';

const defaultConfig: SearchConfig = { repos: [] };

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
    : { selectedPath: '', filters: emptyFilters() };

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
  const [activeTab, setActiveTab] = useState<Tab>('prs');
  const [query, setQuery] = useState('');
  const [copyState, setCopyState] = useState<'idle' | 'copying' | 'copied' | 'error'>('idle');
  const [syncStatus, setSyncStatus] = useState<Record<string, PRSyncStatus>>({});
  const [gavelResultsMap, setGavelResultsMap] = useState<Record<string, GavelResultsSummary>>({});
  const [projects, setProjects] = useState<Project[]>([]);
  const [procStatus, setProcStatus] = useState<Record<string, ProcStatus>>({});
  const [addOpen, setAddOpen] = useState(false);
  const [editProject, setEditProject] = useState<Project | null>(null);
  const [copyError, setCopyError] = useState('');
  const copyResetTimer = useRef<number | null>(null);
  const [, tick] = useState(0);

  const prs = useMemo(() => annotateRoutePaths(rawPrs), [rawPrs]);

  const routeState: RouteState = useMemo(
    () => ({ selectedPath, filters }),
    [selectedPath, filters],
  );

  const commitRoute = useCallback((next: RouteState, mode: 'push' | 'replace' = 'push') => {
    setSelectedPath(next.selectedPath);
    setFilters(next.filters);
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
      setSelectedPath(next.selectedPath);
      setFilters(next.filters);
    };
    window.addEventListener('popstate', onPopState);
    return () => window.removeEventListener('popstate', onPopState);
  }, []);

  useEffect(() => {
    fetch('/api/prs')
      .then(r => r.json())
      .then((snap: Snapshot) => applySnapshot(snap))
      .catch(() => {});

    const es = new EventSource('/api/prs/stream');
    es.addEventListener('message', (e: MessageEvent) => {
      applySnapshot(JSON.parse(e.data));
    });
    es.onerror = () => setError('Connection lost — retrying...');

    const timer = setInterval(() => tick(n => n + 1), 1000);
    return () => { es.close(); clearInterval(timer); };
  }, []);

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

  // True while any process is mid-transition (starting/restarting) so the poller
  // can refresh faster and the start/restart progress shows promptly.
  const anyProcTransitioning = useMemo(
    () => Object.values(procStatus).some(
      s => s.processes?.some(p => p.status === 'starting' || p.status === 'restarting')),
    [procStatus],
  );

  // Poll process status only while projects are configured and the tab is
  // visible — this is the sole driver of the sidebar's per-repo proc badges
  // and is intentionally decoupled from the GitHub PR poller. While a process is
  // transitioning, poll every 1s so the UI tracks the start/restart progress;
  // otherwise the steady-state 3s cadence.
  useEffect(() => {
    if (projects.length === 0) { setProcStatus({}); return; }
    fetchProcStatus();
    const id = setInterval(() => {
      if (document.visibilityState === 'visible') fetchProcStatus();
    }, anyProcTransitioning ? 1000 : 3000);
    return () => clearInterval(id);
  }, [projects.length, fetchProcStatus, anyProcTransitioning]);

  const projectsByRepo = useMemo(() => {
    const m: Record<string, Project> = {};
    for (const p of projects) for (const r of p.repos || []) m[r] = p;
    return m;
  }, [projects]);

  // Projects whose repos have no PRs in the current list get pinned above it so
  // their controls stay reachable (a local dir with no open PRs still shows).
  const standaloneProjects = useMemo(() => {
    const reposWithPRs = new Set(prs.map(p => p.repo));
    return projects.filter(p => !(p.repos || []).some(r => reposWithPRs.has(r)));
  }, [projects, prs]);

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

  const resetCopyFeedback = useCallback((kind: 'copied' | 'error', message?: string) => {
    setCopyState(kind);
    setCopyError(message || '');
    if (copyResetTimer.current) window.clearTimeout(copyResetTimer.current);
    copyResetTimer.current = window.setTimeout(() => {
      setCopyState('idle');
      setCopyError('');
      copyResetTimer.current = null;
    }, 2500);
  }, []);

  const onDownloadJSON = useCallback(() => downloadCurrentView(routeState, 'json'), [routeState]);
  const onDownloadMarkdown = useCallback(() => downloadCurrentView(routeState, 'md'), [routeState]);
  const onCopyForAgent = useCallback(async () => {
    if (copyState === 'copying') return;
    setCopyState('copying');
    setCopyError('');
    try {
      await copyCurrentViewForAgent(routeState);
      resetCopyFeedback('copied');
    } catch (e: any) {
      resetCopyFeedback('error', e?.message || 'Copy failed');
    }
  }, [copyState, routeState, resetCopyFeedback]);

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

  return (
    <>
      <AppShell
        brand={<img src="/brand/gavel-logo.svg" alt="gavel" className="h-7" />}
        nav={<TabBar active={activeTab} onChange={setActiveTab} />}
        search={
          activeTab === 'prs' ? (
            <div className="flex w-full items-center gap-2 rounded-md border border-border bg-muted px-3 py-1.5">
              <iconify-icon icon="codicon:search" className="text-muted-foreground text-sm shrink-0" />
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
                  <iconify-icon icon="codicon:close" className="text-xs" />
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
          ) : undefined
        }
        bodyHeader={
          activeTab === 'prs' ? (
            <span className="text-xs text-muted-foreground">
              {searched.length} pull request{searched.length !== 1 ? 's' : ''}
            </span>
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
          ) : undefined
        }
        contentClassName="overflow-hidden"
      >
        {activeTab === 'prs' ? (
          <SplitPane
            left={
              <>
                <ProjectsBar projects={standaloneProjects} procStatus={procStatus} onChanged={onProcChanged} onEdit={openEdit} onAdd={openAdd} />
                <PRList prs={searched} selected={selected} onSelect={handleSelect} unread={unread} syncStatus={syncStatus} gavelResults={gavelResultsMap} projectsByRepo={projectsByRepo} procStatus={procStatus} onProcChanged={onProcChanged} onProcEdit={openEdit} />
              </>
            }
            right={
              selected ? (
                <PRDetailPanel pr={selected} detail={detail} loading={detailLoading} />
              ) : (
                <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
                  <div className="text-center">
                    <iconify-icon icon="codicon:git-pull-request" className="text-4xl mb-2" />
                    <p>Select a PR to view details</p>
                  </div>
                </div>
              )
            }
          />
        ) : (
          <ActivityView />
        )}
      </AppShell>

      <AddProjectDialog open={addOpen} onClose={() => setAddOpen(false)} onSaved={onProcChanged} repoOptions={reposList} edit={editProject} />
    </>
  );
}

function TabBar({ active, onChange }: { active: Tab; onChange: (t: Tab) => void }) {
  const tabs: { id: Tab; label: string; icon: string }[] = [
    { id: 'prs', label: 'PRs', icon: 'codicon:git-pull-request' },
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
          <iconify-icon icon={t.icon} className="mr-1" />
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
        <iconify-icon icon="codicon:json" className="mr-0.5" />
        JSON
      </button>
      <button
        className="text-xs px-2 py-1 rounded border border-border text-muted-foreground hover:bg-muted transition-colors"
        onClick={onMarkdown}
        title="Download current view as Markdown"
      >
        <iconify-icon icon="codicon:markdown" className="mr-0.5" />
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
        <iconify-icon
          icon={copyState === 'copied' ? 'codicon:check' : copyState === 'copying' ? 'svg-spinners:ring-resize' : 'codicon:copy'}
          className="mr-0.5"
        />
        {copyState === 'copying' ? 'Copying...' : copyState === 'copied' ? 'Copied' : copyState === 'error' ? 'Copy failed' : 'Copy for Agent'}
      </button>
    </div>
  );
}
