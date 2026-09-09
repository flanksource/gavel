import type { SearchConfig } from './types';
import type { Filters, FilterMode } from './components/FilterBar';

const KEY = 'gavel.pr-ui.v1';

// Each facet persists as a map of option key -> include/exclude (tri-state),
// matching the in-memory Filters shape.
type StoredFacet = Record<string, FilterMode>;

interface StoredFilters {
  state: StoredFacet;
  checks: StoredFacet;
  repos: StoredFacet;
  authors: StoredFacet;
}

interface Stored {
  config?: Partial<SearchConfig>;
  filters?: StoredFilters;
}

// loadUIState reads the persisted org/search config and filters. localStorage
// can throw (private mode / disabled storage); persistence is best-effort, so a
// failure falls back to empty rather than breaking startup.
export function loadUIState(): Stored {
  try {
    return JSON.parse(localStorage.getItem(KEY) || '{}') as Stored;
  } catch {
    return {};
  }
}

export function saveUIState(config: SearchConfig, filters: Filters): void {
  const stored: Stored = {
    config,
    filters: {
      state: filters.state,
      checks: filters.checks,
      repos: filters.repos,
      authors: filters.authors,
    },
  };
  try {
    localStorage.setItem(KEY, JSON.stringify(stored));
  } catch {
    // best-effort: storage unavailable (private mode) — skip persisting.
  }
}

export function filtersFromStored(s: Stored['filters']): Filters | null {
  if (!s) return null;
  return {
    state: s.state || {},
    checks: s.checks || {},
    repos: s.repos || {},
    authors: s.authors || {},
  };
}
