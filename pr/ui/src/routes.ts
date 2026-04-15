import type { PRItem } from './types';
import { emptyFilters, type Filters } from './components/FilterBar';

export type ExportFormat = 'json' | 'md';

export interface RouteState {
  selectedPath: string;
  filters: Filters;
}

function splitCSV(value: string | null): string[] {
  if (!value) return [];
  return value.split(',').map(v => v.trim()).filter(Boolean);
}

export function parseRoute(location: Location): RouteState {
  const trimmed = location.pathname.replace(/^\/+|\/+$/g, '');
  const segments = trimmed ? trimmed.split('/').map(decodeURIComponent) : [];
  let selectedPath = '';
  if (segments[0] === 'prs' && segments.length > 1) {
    selectedPath = segments.slice(1).join('/');
  }

  const params = new URLSearchParams(location.search);
  return {
    selectedPath,
    filters: {
      state: new Set(splitCSV(params.get('state'))),
      checks: new Set(splitCSV(params.get('checks'))),
      repos: new Set(splitCSV(params.get('repos'))),
      authors: new Set(splitCSV(params.get('authors'))),
    },
  };
}

export function buildRoute(state: RouteState): string {
  const segments: string[] = ['prs'];
  if (state.selectedPath) {
    segments.push(...state.selectedPath.split('/').map(encodeURIComponent));
  }

  const params = new URLSearchParams();
  if (state.filters.state.size > 0) params.set('state', Array.from(state.filters.state).join(','));
  if (state.filters.checks.size > 0) params.set('checks', Array.from(state.filters.checks).join(','));
  if (state.filters.repos.size > 0) params.set('repos', Array.from(state.filters.repos).join(','));
  if (state.filters.authors.size > 0) params.set('authors', Array.from(state.filters.authors).join(','));

  const query = params.toString();
  return `/${segments.join('/')}${query ? `?${query}` : ''}`;
}

export function buildExportRoute(state: RouteState, format: ExportFormat): string {
  const route = buildRoute(state);
  const [path, query = ''] = route.split('?', 2);
  return `${path}.${format}${query ? `?${query}` : ''}`;
}

export function annotateRoutePaths(prs: PRItem[]): PRItem[] {
  return prs.map(pr => ({ ...pr, route_path: `${pr.repo}/${pr.number}` }));
}

export function findPRByRoutePath(prs: PRItem[], target: string): PRItem | null {
  if (!target) return null;
  for (const pr of prs) {
    if (pr.route_path === target) return pr;
  }
  return null;
}

export function emptyRouteState(): RouteState {
  return { selectedPath: '', filters: emptyFilters() };
}
