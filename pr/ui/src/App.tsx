import { useState, useEffect, useMemo } from 'preact/hooks';
import type { PRItem, PRDetail, Snapshot, SearchConfig } from './types';
import { Summary } from './components/Summary';
import { PRList } from './components/PRList';
import { PRDetailPanel } from './components/PRDetail';
import { FilterBar, emptyFilters, type Filters } from './components/FilterBar';
import { SplitPane } from './components/SplitPane';
import { RepoSelector } from './components/RepoSelector';
import { SearchControls } from './components/SearchControls';
import { computeCounts, collectRepos, collectAuthors, filterPRs } from './utils';

const defaultConfig: SearchConfig = { repos: [], author: '@me' };

export function App() {
  const [prs, setPrs] = useState<PRItem[]>([]);
  const [fetchedAt, setFetchedAt] = useState('');
  const [nextFetchIn, setNextFetchIn] = useState(60);
  const [error, setError] = useState<string | undefined>();
  const [selected, setSelected] = useState<PRItem | null>(null);
  const [detail, setDetail] = useState<PRDetail | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [filters, setFilters] = useState<Filters>(emptyFilters());
  const [config, setConfig] = useState<SearchConfig>(defaultConfig);
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

    const timer = setInterval(() => tick(n => n + 1), 5000);
    return () => { es.close(); clearInterval(timer); };
  }, []);

  function applySnapshot(snap: Snapshot) {
    setPrs(snap.prs || []);
    setFetchedAt(snap.fetchedAt);
    setNextFetchIn(snap.nextFetchIn);
    setError(snap.error);
    if (snap.config) setConfig(snap.config);
  }

  function handleRefresh() {
    fetch('/api/prs/refresh', { method: 'POST' }).catch(() => {});
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
          <div class="flex items-center gap-3">
            <h1 class="text-xl font-bold text-gray-900 shrink-0">
              <iconify-icon icon="codicon:git-pull-request" class="mr-1.5 text-blue-600" />
              PR Dashboard
            </h1>
            <RepoSelector
              repos={config.repos || []}
              allOrg={config.all}
              org={config.org}
              onChange={(repos) => updateConfig({ repos })}
            />
            <SearchControls config={config} onChange={updateConfig} />
            <FilterBar filters={filters} onChange={setFilters} counts={counts} repos={repos} authors={authors} />
          </div>
          <Summary
            prs={prs}
            fetchedAt={fetchedAt}
            error={error}
            nextFetchIn={nextFetchIn}
            onRefresh={handleRefresh}
          />
        </div>
      </div>

      <SplitPane
        left={<PRList prs={filtered} selected={selected} onSelect={handleSelect} />}
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
    </div>
  );
}
