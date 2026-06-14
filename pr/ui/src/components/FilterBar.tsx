import { FilterBar as ClickyFilterBar } from '@flanksource/clicky-ui/components';
import type { FilterBarFilter, MultiSelectOption } from '@flanksource/clicky-ui/components';

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

// FilterBar wraps clicky-ui's FilterBar but keeps the project's Set-based
// Filters contract so App.tsx, routes.ts and utils (filterPRs/computeCounts +
// URL sync) are untouched. Each facet is a `select-multi` whose option labels
// carry the live result count.
export function FilterBar({ filters, onChange, counts, repos, authors }: Props) {
  const setFacet = (key: keyof Filters, values: string[]) =>
    onChange({ ...filters, [key]: new Set(values) });

  const c = counts as Record<string, number>;
  const stateOpts: MultiSelectOption[] = STATE_DEFS
    .filter(d => c[d.key] > 0)
    .map(d => ({ value: d.key, label: `${d.label} (${c[d.key]})` }));
  const checkOpts: MultiSelectOption[] = CHECK_DEFS
    .filter(d => c[d.key] > 0)
    .map(d => ({ value: d.key, label: `${d.label} (${c[d.key]})` }));
  const repoOpts: MultiSelectOption[] = repos.map(r => ({ value: r, label: shortName(r) }));
  const authorOpts: MultiSelectOption[] = authors.map(a => ({ value: a, label: `@${a}` }));

  const fb: FilterBarFilter[] = [];
  if (stateOpts.length) fb.push({ key: 'state', kind: 'select-multi', label: 'State', options: stateOpts, value: [...filters.state], onChange: (v) => setFacet('state', v) });
  if (checkOpts.length) fb.push({ key: 'checks', kind: 'select-multi', label: 'Checks', options: checkOpts, value: [...filters.checks], onChange: (v) => setFacet('checks', v) });
  if (repos.length > 1) fb.push({ key: 'repos', kind: 'select-multi', label: 'Repos', options: repoOpts, value: [...filters.repos], onChange: (v) => setFacet('repos', v) });
  if (authors.length > 1) fb.push({ key: 'authors', kind: 'select-multi', label: 'Authors', options: authorOpts, value: [...filters.authors], onChange: (v) => setFacet('authors', v) });

  if (fb.length === 0) return null;
  return <ClickyFilterBar overflowMode="wrap" filters={fb} />;
}
