import { useState, useEffect, useMemo, useCallback, useRef } from 'preact/hooks';
import type { PRItem, PRDetail, Snapshot, SearchConfig, RateLimit, PRSyncStatus } from './types';
import { Summary } from './components/Summary';
import { PRList } from './components/PRList';
import { PRDetailPanel } from './components/PRDetail';
import { FilterBar, type Filters } from './components/FilterBar';
import { SplitPane } from './components/SplitPane';
import { RepoSelector } from './components/RepoSelector';
import { SearchControls } from './components/SearchControls';
import { ActivityView } from './components/ActivityView';
import { computeCounts, collectRepos, collectAuthors, filterPRs, prKey } from './utils';
import {
  annotateRoutePaths,
  buildRoute,
  findPRByRoutePath,
  parseRoute,
  type RouteState,
} from './routes';
import { copyCurrentViewForAgent, downloadCurrentView } from './export';

type Tab = 'prs' | 'activity';

const defaultConfig: SearchConfig = { repos: [], author: '@me' };

export function App() {
  const initialRoute: RouteState = typeof window !== 'undefined'
    ? parseRoute(window.location)
    : { selectedPath: '', filters: { state: new Set(), checks: new Set(), repos: new Set(), authors: new Set() } };

  const [rawPrs, setRawPrs] = useState<PRItem[]>([]);
  const [unread, setUnread] = useState<Record<string, boolean>>({});
  const [fetchedAt, setFetchedAt] = useState('');
  const [nextFetchIn, setNextFetchIn] = useState(60);
  const [error, setError] = useState<string | undefined>();
  const [selected, setSelected] = useState<PRItem | null>(null);
  const [detail, setDetail] = useState<PRDetail | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [filters, setFilters] = useState<Filters>(initialRoute.filters);
  const [selectedPath, setSelectedPath] = useState(initialRoute.selectedPath);
  const [config, setConfig] = useState<SearchConfig>(defaultConfig);
  const [paused, setPaused] = useState(false);
  const [rateLimit, setRateLimit] = useState<RateLimit | undefined>();
  const [activeTab, setActiveTab] = useState<Tab>('prs');
  const [copyState, setCopyState] = useState<'idle' | 'copying' | 'copied' | 'error'>('idle');
  const [syncStatus, setSyncStatus] = useState<Record<string, PRSyncStatus>>({});
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
    setUnread(snap.unread || {});
    setFetchedAt(snap.fetchedAt);
    setNextFetchIn(snap.nextFetchIn);
    setError(snap.error);
    setPaused(snap.paused);
    if (snap.rateLimit) setRateLimit(snap.rateLimit);
    if (snap.config) setConfig(snap.config);
    if (snap.syncStatus) setSyncStatus(snap.syncStatus);
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
      setDetail(prev => prev ? { ...prev, gavelResults: data.gavelResults } : prev);
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
  const authors = useMemo(() => collectAuthors(prs), [prs]);
  const filtered = useMemo(
    () => filterPRs(prs, filters.state, filters.checks, filters.repos, filters.authors),
    [prs, filters],
  );

  return (
    <div class="bg-gray-100 h-screen flex flex-col">
      <div class="border-b bg-white px-6 py-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3 flex-wrap">
            <h1 class="shrink-0 flex items-center">
              <img src="/brand/gavel-logo.svg" alt="gavel" class="h-7" />
            </h1>
            <TabBar active={activeTab} onChange={setActiveTab} />
            {activeTab === 'prs' && (
              <>
                <RepoSelector
                  repos={config.repos || []}
                  allOrg={config.all}
                  org={config.org}
                  onChange={(repos) => updateConfig({ repos })}
                />
                <SearchControls config={config} onChange={updateConfig} />
                <FilterBar filters={filters} onChange={handleFiltersChange} counts={counts} repos={reposList} authors={authors} />
                <ExportButtons
                  onJSON={onDownloadJSON}
                  onMarkdown={onDownloadMarkdown}
                  onCopy={onCopyForAgent}
                  copyState={copyState}
                  copyError={copyError}
                />
              </>
            )}
          </div>
          <Summary
            prs={prs}
            fetchedAt={fetchedAt}
            error={error}
            nextFetchIn={nextFetchIn}
            paused={paused}
            rateLimit={rateLimit}
            onRefresh={handleRefresh}
            onPause={handlePause}
            networkBusy={detailLoading}
          />
        </div>
      </div>

      {activeTab === 'prs' ? (
        <SplitPane
          left={<PRList prs={filtered} selected={selected} onSelect={handleSelect} unread={unread} syncStatus={syncStatus} />}
          right={
            selected ? (
              <PRDetailPanel pr={selected} detail={detail} loading={detailLoading} />
            ) : (
              <div class="flex items-center justify-center h-full text-gray-400 text-sm">
                <div class="text-center">
                  <iconify-icon icon="codicon:git-pull-request" class="text-4xl mb-2" />
                  <p>Select a PR to view details</p>
                </div>
              </div>
            )
          }
        />
      ) : (
        <ActivityView />
      )}
    </div>
  );
}

function TabBar({ active, onChange }: { active: Tab; onChange: (t: Tab) => void }) {
  const tabs: { id: Tab; label: string; icon: string }[] = [
    { id: 'prs', label: 'PRs', icon: 'codicon:git-pull-request' },
    { id: 'activity', label: 'Activity', icon: 'codicon:pulse' },
  ];
  return (
    <div class="flex gap-1 border-b border-transparent">
      {tabs.map(t => (
        <button
          key={t.id}
          onClick={() => onChange(t.id)}
          class={`px-3 py-1.5 text-sm rounded-md transition ${
            active === t.id
              ? 'bg-blue-50 text-blue-700 font-medium'
              : 'text-gray-600 hover:bg-gray-100'
          }`}
        >
          <iconify-icon icon={t.icon} class="mr-1" />
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
    <div class="flex items-center gap-1">
      <button
        class="text-xs px-2 py-1 rounded border border-gray-300 text-gray-600 hover:bg-gray-200 transition-colors"
        onClick={onJSON}
        title="Download current view as JSON"
      >
        <iconify-icon icon="codicon:json" class="mr-0.5" />
        JSON
      </button>
      <button
        class="text-xs px-2 py-1 rounded border border-gray-300 text-gray-600 hover:bg-gray-200 transition-colors"
        onClick={onMarkdown}
        title="Download current view as Markdown"
      >
        <iconify-icon icon="codicon:markdown" class="mr-0.5" />
        Markdown
      </button>
      <button
        class={`text-xs px-2 py-1 rounded border transition-colors ${
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
          class="mr-0.5"
        />
        {copyState === 'copying' ? 'Copying...' : copyState === 'copied' ? 'Copied' : copyState === 'error' ? 'Copy failed' : 'Copy for Agent'}
      </button>
    </div>
  );
}
