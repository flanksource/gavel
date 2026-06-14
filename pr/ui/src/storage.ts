import type { SearchConfig } from './types';
import type { Filters } from './components/FilterBar';

const KEY = 'gavel.pr-ui.v1';

interface StoredFilters {
  state: string[];
  checks: string[];
  repos: string[];
  authors: string[];
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
      state: [...filters.state],
      checks: [...filters.checks],
      repos: [...filters.repos],
      authors: [...filters.authors],
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
    state: new Set(s.state || []),
    checks: new Set(s.checks || []),
    repos: new Set(s.repos || []),
    authors: new Set(s.authors || []),
  };
}
