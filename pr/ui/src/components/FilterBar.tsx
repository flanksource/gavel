import { FilterBar as ClickyFilterBar } from '@flanksource/clicky-ui/components';
import type { FilterBarFilter, MultiSelectOption } from '@flanksource/clicky-ui/components';

export type FilterMode = 'include' | 'exclude';

// Facets are tri-state: a key maps to 'include' or 'exclude'; absent = neutral.
// This is exactly clicky-ui's FilterBar `kind:"multi"` value shape.
export interface Filters {
  state: Record<string, FilterMode>;
  checks: Record<string, FilterMode>;
  repos: Record<string, FilterMode>;
  authors: Record<string, FilterMode>;
}

export const emptyFilters = (): Filters => ({
  state: {},
  checks: {},
  repos: {},
  authors: {},
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

const STATE_DEFS: { key: string; label: string }[] = [
  { key: 'open', label: 'Open' },
  { key: 'merged', label: 'Merged' },
  { key: 'closed', label: 'Closed' },
  { key: 'draft', label: 'Draft' },
];

const CHECK_DEFS: { key: string; label: string }[] = [
  { key: 'failing', label: 'Failing' },
  { key: 'passing', label: 'Passing' },
  { key: 'running', label: 'Running' },
];

function shortName(repo: string): string {
  return repo.includes('/') ? repo.split('/')[1] : repo;
}

// authorLabel renders the synthetic @me / @bots keys verbatim and prefixes real
// logins with "@".
function authorLabel(author: string): string {
  if (author === '@me') return '@me';
  if (author === '@bots') return 'bots';
  return `@${author}`;
}

// FilterBar wraps clicky-ui's FilterBar with tri-state (include/exclude) facets.
// The project's record-based Filters contract flows straight through to each
// `kind:"multi"` filter's `value`/`onChange`, and on to filterPRs / routes /
// storage unchanged.
export function FilterBar({ filters, onChange, counts, repos, authors }: Props) {
  const setFacet = (key: keyof Filters, value: Record<string, FilterMode>) =>
    onChange({ ...filters, [key]: value });

  const c = counts as Record<string, number>;
  const stateOpts: MultiSelectOption[] = STATE_DEFS
    .filter(d => c[d.key] > 0)
    .map(d => ({ value: d.key, label: `${d.label} (${c[d.key]})` }));
  const checkOpts: MultiSelectOption[] = CHECK_DEFS
    .filter(d => c[d.key] > 0)
    .map(d => ({ value: d.key, label: `${d.label} (${c[d.key]})` }));
  const repoOpts: MultiSelectOption[] = repos.map(r => ({ value: r, label: shortName(r) }));
  const authorOpts: MultiSelectOption[] = authors.map(a => ({ value: a, label: authorLabel(a) }));

  const fb: FilterBarFilter[] = [];
  if (stateOpts.length) fb.push({ key: 'state', kind: 'multi', label: 'State', options: stateOpts, value: filters.state, onChange: (v) => setFacet('state', v) });
  if (checkOpts.length) fb.push({ key: 'checks', kind: 'multi', label: 'Checks', options: checkOpts, value: filters.checks, onChange: (v) => setFacet('checks', v) });
  if (repos.length > 1) fb.push({ key: 'repos', kind: 'multi', label: 'Repos', icon: 'codicon:repo', options: repoOpts, value: filters.repos, onChange: (v) => setFacet('repos', v) });
  if (authors.length > 1) fb.push({ key: 'authors', kind: 'multi', label: 'Authors', icon: 'codicon:person', options: authorOpts, value: filters.authors, onChange: (v) => setFacet('authors', v) });

  if (fb.length === 0) return null;
  return <ClickyFilterBar overflowMode="wrap" filters={fb} />;
}
