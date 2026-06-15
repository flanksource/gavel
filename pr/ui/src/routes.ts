import type { PRItem } from './types';
import { emptyFilters, type Filters, type FilterMode } from './components/FilterBar';

export type ExportFormat = 'json' | 'md';

export interface RouteState {
  selectedPath: string;
  filters: Filters;
}

function splitCSV(value: string | null): string[] {
  if (!value) return [];
  return value.split(',').map(v => v.trim()).filter(Boolean);
}

// Tri-state facets are encoded in the URL as a CSV where excluded keys carry a
// leading "-" (e.g. repos=foo,-bar means include foo, exclude bar).
function parseFacet(value: string | null): Record<string, FilterMode> {
  const out: Record<string, FilterMode> = {};
  for (const raw of splitCSV(value)) {
    if (raw.startsWith('-')) out[raw.slice(1)] = 'exclude';
    else out[raw] = 'include';
  }
  return out;
}

function buildFacet(modes: Record<string, FilterMode>): string {
  return Object.entries(modes)
    .map(([k, m]) => (m === 'exclude' ? `-${k}` : k))
    .join(',');
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
      state: parseFacet(params.get('state')),
      checks: parseFacet(params.get('checks')),
      repos: parseFacet(params.get('repos')),
      authors: parseFacet(params.get('authors')),
    },
  };
}

export function buildRoute(state: RouteState): string {
  const segments: string[] = ['prs'];
  if (state.selectedPath) {
    segments.push(...state.selectedPath.split('/').map(encodeURIComponent));
  }

  const params = new URLSearchParams();
  const { state: st, checks, repos, authors } = state.filters;
  if (Object.keys(st).length) params.set('state', buildFacet(st));
  if (Object.keys(checks).length) params.set('checks', buildFacet(checks));
  if (Object.keys(repos).length) params.set('repos', buildFacet(repos));
  if (Object.keys(authors).length) params.set('authors', buildFacet(authors));

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
