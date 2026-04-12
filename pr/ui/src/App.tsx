import { useState, useEffect, useMemo, useCallback } from 'preact/hooks';
import type { PRItem, PRDetail, Snapshot, SearchConfig, RateLimit } from './types';
import { Summary } from './components/Summary';
import { PRList } from './components/PRList';
import { PRDetailPanel } from './components/PRDetail';
import { FilterBar, emptyFilters, type Filters } from './components/FilterBar';
import { SplitPane } from './components/SplitPane';
import { RepoSelector } from './components/RepoSelector';
import { SearchControls } from './components/SearchControls';
import { ActivityView } from './components/ActivityView';
import { computeCounts, collectRepos, collectAuthors, filterPRs, prKey } from './utils';

type Tab = 'prs' | 'activity';

const defaultConfig: SearchConfig = { repos: [], author: '@me' };

export function App() {
  const [prs, setPrs] = useState<PRItem[]>([]);
  const [unread, setUnread] = useState<Record<string, boolean>>({});
  const [fetchedAt, setFetchedAt] = useState('');
  const [nextFetchIn, setNextFetchIn] = useState(60);
  const [error, setError] = useState<string | undefined>();
  const [selected, setSelected] = useState<PRItem | null>(null);
  const [detail, setDetail] = useState<PRDetail | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [filters, setFilters] = useState<Filters>(emptyFilters());
  const [config, setConfig] = useState<SearchConfig>(defaultConfig);
  const [paused, setPaused] = useState(false);
  const [rateLimit, setRateLimit] = useState<RateLimit | undefined>();
  const [activeTab, setActiveTab] = useState<Tab>('prs');
  const [, tick] = useState(0);

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
    setPrs(snap.prs || []);
    setUnread(snap.unread || {});
    setFetchedAt(snap.fetchedAt);
    setNextFetchIn(snap.nextFetchIn);
    setError(snap.error);
    setPaused(snap.paused);
    if (snap.rateLimit) setRateLimit(snap.rateLimit);
    if (snap.config) setConfig(snap.config);
  }

  // markSeen clears the unread flag for a PR locally (optimistic update) and
  // POSTs to the server so the state is persisted and the menubar updates.
  // Safe to call repeatedly — the server handler is idempotent.
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

  function handleSelect(pr: PRItem) {
    setSelected(pr);
    setDetail(null);
    setDetailLoading(true);
    markSeen(pr);
    fetch(`/api/prs/detail?repo=${encodeURIComponent(pr.repo)}&number=${pr.number}`)
      .then(r => r.json())
      .then((d: PRDetail) => { setDetail(d); setDetailLoading(false); })
      .catch(err => { setDetail({ error: String(err) }); setDetailLoading(false); });
  }

  const counts = useMemo(() => computeCounts(prs), [prs]);
  const repos = useMemo(() => collectRepos(prs), [prs]);
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
                <FilterBar filters={filters} onChange={setFilters} counts={counts} repos={repos} authors={authors} />
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
          />
        </div>
      </div>

      {activeTab === 'prs' ? (
        <SplitPane
          left={<PRList prs={filtered} selected={selected} onSelect={handleSelect} unread={unread} />}
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
