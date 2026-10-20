import {
  DEFAULT_LINT_GROUPING,
  LINT_GROUPINGS,
  decodeFilterState,
  encodeFilterState,
  type FilterState,
  type LintFilters,
  type LintGrouping,
  type LintSeverity,
  type TestFilters,
} from '@flanksource/clicky-ui/data';
import { basePath } from './config';

const DEFAULT_STATUS_FILTER_TOKENS = ['!passed', '!skipped'];
const STATUS_FILTER_ALL_SENTINEL = 'all';

export function defaultStatusFilter(): FilterState<string> {
  return decodeFilterState(DEFAULT_STATUS_FILTER_TOKENS);
}

function isDefaultStatusFilter(state: FilterState<string>): boolean {
  const tokens = encodeFilterState(state);
  if (tokens.length !== DEFAULT_STATUS_FILTER_TOKENS.length) return false;
  const want = [...DEFAULT_STATUS_FILTER_TOKENS].sort();
  const got = [...tokens].sort();
  return want.every((token, i) => token === got[i]);
}

export type TabKey = 'tests' | 'lint' | 'bench' | 'processes';

const TAB_KEYS: readonly TabKey[] = ['tests', 'lint', 'bench', 'processes'];

export interface RouteState {
  tab: TabKey;
  selectedPath: string;
  filters: TestFilters;
  lintGrouping: LintGrouping;
  lintFilters: LintFilters;
  sampleIndex?: number;
  hideExited?: boolean;
}

export type ExportFormat = 'json' | 'md' | 'pdf';

function splitCSV(value: string | null): string[] {
  if (!value) return [];
  return value.split(',').map(v => v.trim()).filter(Boolean);
}

export function parseRoute(location: Pick<Location, 'pathname' | 'search'>): RouteState {
  let pathname = location.pathname;
  if (basePath && pathname.startsWith(basePath)) {
    pathname = pathname.slice(basePath.length);
  }
  const trimmed = pathname.replace(/^\/+|\/+$/g, '');
  const segments = trimmed ? trimmed.split('/').map(decodeURIComponent) : [];
  let tab: TabKey = 'tests';
  let selectedPath = '';

  const first = segments[0] as TabKey | undefined;
  if (first && TAB_KEYS.includes(first)) {
    tab = first;
    selectedPath = segments.slice(1).join('/');
  }

  const params = new URLSearchParams(location.search);
  const sample = params.get('sample');
  const rawStatus = params.get('status');
  let status: FilterState<string>;
  if (rawStatus === null) {
    status = defaultStatusFilter();
  } else if (rawStatus === STATUS_FILTER_ALL_SENTINEL) {
    status = new Map();
  } else {
    status = decodeFilterState(splitCSV(rawStatus));
  }
  return {
    tab,
    selectedPath,
    filters: {
      status,
      framework: decodeFilterState(splitCSV(params.get('framework'))),
    },
    lintGrouping: parseLintGrouping(params.get('grouping')),
    lintFilters: {
      severity: decodeFilterState<LintSeverity>(splitCSV(params.get('severity'))),
      linter: decodeFilterState(splitCSV(params.get('linter'))),
    },
    sampleIndex: sample !== null && /^\d+$/.test(sample) ? Number(sample) : undefined,
    hideExited: params.get('hideExited') === '1',
  };
}

export function buildRoute(state: RouteState): string {
  const segments: string[] = [state.tab];
  if (state.selectedPath) segments.push(...state.selectedPath.split('/').map(encodeURIComponent));

  const params = new URLSearchParams();
  if (state.tab === 'tests') {
    if (isDefaultStatusFilter(state.filters.status)) {
      // Omit the param entirely — a URL without `status=` already means
      // "apply defaults" under parseRoute.
    } else if (state.filters.status.size === 0) {
      params.set('status', STATUS_FILTER_ALL_SENTINEL);
    } else {
      params.set('status', encodeFilterState(state.filters.status).join(','));
    }
    if (state.filters.framework.size > 0) params.set('framework', encodeFilterState(state.filters.framework).join(','));
  }
  if (state.tab === 'lint') {
    if (state.lintGrouping !== DEFAULT_LINT_GROUPING) params.set('grouping', state.lintGrouping);
    if (state.lintFilters.severity.size > 0) params.set('severity', encodeFilterState(state.lintFilters.severity).join(','));
    if (state.lintFilters.linter.size > 0) params.set('linter', encodeFilterState(state.lintFilters.linter).join(','));
  }
  if (state.tab === 'bench' && state.sampleIndex !== undefined) params.set('sample', String(state.sampleIndex));
  if (state.tab === 'processes' && state.hideExited) params.set('hideExited', '1');

  const query = params.toString();
  return `${basePath}/${segments.join('/')}${query ? `?${query}` : ''}`;
}

export function buildExportRoute(state: RouteState, format: ExportFormat): string {
  const route = buildRoute(state);
  const [path, query = ''] = route.split('?', 2);
  return `${path}.${format}${query ? `?${query}` : ''}`;
}

function parseLintGrouping(value: string | null): LintGrouping {
  const grouping = LINT_GROUPINGS.find(candidate => candidate === value);
  return grouping ?? DEFAULT_LINT_GROUPING;
}
