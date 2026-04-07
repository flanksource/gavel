export interface Filters {
  state: Set<string>;
  checks: Set<string>;
  repos: Set<string>;
  authors: Set<string>;
}

export const emptyFilters = (): Filters => ({
  state: new Set(),
  checks: new Set(),
  repos: new Set(),
  authors: new Set(),
});

interface Props {
  filters: Filters;
  onChange: (f: Filters) => void;
  counts: {
    open: number; merged: number; closed: number; draft: number;
    failing: number; passing: number; running: number;
  };
  repos: string[];
  authors: string[];
}

function toggleSet(set: Set<string>, key: string): Set<string> {
  const next = new Set(set);
  if (next.has(key)) next.delete(key);
  else next.add(key);
  return next;
}

const STATE_DEFS: { key: string; label: string; badge: string; activeBg: string; activeBorder: string }[] = [
  { key: 'open', label: 'Open', badge: 'bg-green-500', activeBg: 'bg-green-50', activeBorder: 'border-green-300' },
  { key: 'merged', label: 'Merged', badge: 'bg-purple-500', activeBg: 'bg-purple-50', activeBorder: 'border-purple-300' },
  { key: 'closed', label: 'Closed', badge: 'bg-red-500', activeBg: 'bg-red-50', activeBorder: 'border-red-300' },
  { key: 'draft', label: 'Draft', badge: 'bg-gray-400', activeBg: 'bg-gray-50', activeBorder: 'border-gray-300' },
];

const CHECK_DEFS: { key: string; label: string; badge: string; activeBg: string; activeBorder: string }[] = [
  { key: 'failing', label: 'Failing', badge: 'bg-red-500', activeBg: 'bg-red-50', activeBorder: 'border-red-300' },
  { key: 'passing', label: 'Passing', badge: 'bg-green-500', activeBg: 'bg-green-50', activeBorder: 'border-green-300' },
  { key: 'running', label: 'Running', badge: 'bg-yellow-400', activeBg: 'bg-yellow-50', activeBorder: 'border-yellow-300' },
];

export function FilterBar({ filters, onChange, counts, repos, authors }: Props) {
  const hasActiveFilters = filters.state.size > 0 || filters.checks.size > 0
    || filters.repos.size > 0 || filters.authors.size > 0;

  return (
    <div class="flex items-center gap-1.5 flex-wrap">
      {STATE_DEFS.map(sf => {
        const count = (counts as any)[sf.key] as number;
        if (count === 0) return null;
        const active = filters.state.has(sf.key);
        return (
          <Pill key={sf.key} active={active} badge={sf.badge} activeBg={sf.activeBg} activeBorder={sf.activeBorder}
            count={count} label={sf.label}
            onClick={() => onChange({ ...filters, state: toggleSet(filters.state, sf.key) })} />
        );
      })}

      <Sep />

      {CHECK_DEFS.map(cf => {
        const count = (counts as any)[cf.key] as number;
        if (count === 0) return null;
        const active = filters.checks.has(cf.key);
        return (
          <Pill key={cf.key} active={active} badge={cf.badge} activeBg={cf.activeBg} activeBorder={cf.activeBorder}
            count={count} label={cf.label}
            onClick={() => onChange({ ...filters, checks: toggleSet(filters.checks, cf.key) })} />
        );
      })}

      {repos.length > 1 && (
        <>
          <Sep />
          {repos.map(repo => {
            const short = repo.includes('/') ? repo.split('/')[1] : repo;
            const active = filters.repos.has(repo);
            return (
              <button key={repo}
                class={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full border transition-colors ${
                  active ? 'bg-cyan-50 border-cyan-300 font-medium text-cyan-700' : 'border-gray-200 text-gray-500 hover:bg-gray-50'
                }`}
                onClick={() => onChange({ ...filters, repos: toggleSet(filters.repos, repo) })}
              >
                <iconify-icon icon="codicon:repo" class="text-[10px]" />
                {short}
              </button>
            );
          })}
        </>
      )}

      {authors.length > 1 && (
        <>
          <Sep />
          {authors.map(author => {
            const active = filters.authors.has(author);
            return (
              <button key={author}
                class={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full border transition-colors ${
                  active ? 'bg-blue-50 border-blue-300 font-medium text-blue-700' : 'border-gray-200 text-gray-500 hover:bg-gray-50'
                }`}
                onClick={() => onChange({ ...filters, authors: toggleSet(filters.authors, author) })}
              >
                @{author}
              </button>
            );
          })}
        </>
      )}

      {hasActiveFilters && (
        <button class="text-xs text-gray-400 hover:text-gray-600 ml-1"
          onClick={() => onChange(emptyFilters())}>
          Clear
        </button>
      )}
    </div>
  );
}

function Sep() {
  return <span class="text-gray-300 mx-0.5">|</span>;
}

function Pill({ active, badge, activeBg, activeBorder, count, label, onClick }: {
  active: boolean; badge: string; activeBg: string; activeBorder: string;
  count: number; label: string; onClick: () => void;
}) {
  return (
    <button
      class={`inline-flex items-center gap-1.5 text-xs px-2 py-0.5 rounded-full border transition-colors ${
        active ? `${activeBg} ${activeBorder} font-medium` : 'border-gray-200 text-gray-500 hover:bg-gray-50'
      }`}
      onClick={onClick}
    >
      <span class={`inline-flex items-center justify-center min-w-[16px] h-[16px] px-1 rounded-full text-[10px] font-bold text-white ${badge}`}>
        {count}
      </span>
      {label}
    </button>
  );
}
