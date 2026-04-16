import type { Test } from './types';
import type { Filters } from './components/FilterBar';
import type { LintFilters, LintGrouping } from './components/LintFilterBar';

export type TabKey = 'tests' | 'lint' | 'bench' | 'diagnostics';

export interface RouteState {
  tab: TabKey;
  selectedPath: string;
  filters: Filters;
  lintGrouping: LintGrouping;
  lintFilters: LintFilters;
}

export type ExportFormat = 'json' | 'md' | 'pdf';

function splitCSV(value: string | null): string[] {
  if (!value) return [];
  return value.split(',').map(v => v.trim()).filter(Boolean);
}

export function parseRoute(location: Location): RouteState {
  const trimmed = location.pathname.replace(/^\/+|\/+$/g, '');
  const segments = trimmed ? trimmed.split('/').map(decodeURIComponent) : [];
  let tab: TabKey = 'tests';
  let selectedPath = '';

  if (segments[0] === 'tests' || segments[0] === 'lint' || segments[0] === 'bench' || segments[0] === 'diagnostics') {
    tab = segments[0];
    selectedPath = segments.slice(1).join('/');
  }

  const params = new URLSearchParams(location.search);
  const grouping = params.get('grouping');

  return {
    tab,
    selectedPath,
    filters: {
      status: new Set(splitCSV(params.get('status'))),
      framework: new Set(splitCSV(params.get('framework'))),
    },
    lintGrouping: grouping === 'file-linter-rule' ? 'file-linter-rule' : 'linter-file',
    lintFilters: {
      severity: new Set(splitCSV(params.get('severity')) as any[]),
      linter: new Set(splitCSV(params.get('linter'))),
    },
  };
}

export function buildRoute(state: RouteState): string {
  const segments = [state.tab];
  if (state.selectedPath) segments.push(...state.selectedPath.split('/').map(encodeURIComponent));

  const params = new URLSearchParams();
  if (state.tab === 'tests') {
    if (state.filters.status.size > 0) params.set('status', Array.from(state.filters.status).join(','));
    if (state.filters.framework.size > 0) params.set('framework', Array.from(state.filters.framework).join(','));
  }
  if (state.tab === 'lint') {
    if (state.lintGrouping !== 'linter-file') params.set('grouping', state.lintGrouping);
    if (state.lintFilters.severity.size > 0) params.set('severity', Array.from(state.lintFilters.severity).join(','));
    if (state.lintFilters.linter.size > 0) params.set('linter', Array.from(state.lintFilters.linter).join(','));
  }

  const query = params.toString();
  return `/${segments.join('/')}${query ? `?${query}` : ''}`;
}

export function buildExportRoute(state: RouteState, format: ExportFormat): string {
  const route = buildRoute(state);
  const [path, query = ''] = route.split('?', 2);
  return `${path}.${format}${query ? `?${query}` : ''}`;
}

function slugify(value: string): string {
  const slug = value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
  return slug || 'node';
}

function displayLabel(test: Test): string {
  if (test.framework !== 'go test') return test.name;
  if (test.name.endsWith('/')) return test.name;
  const parts = test.name.split('/');
  return parts.map((part, index) => {
    if (index === 0) {
      return part
        .replace(/^Test/, '')
        .replace(/([a-z0-9])([A-Z])/g, '$1 $2');
    }
    return part.replace(/_/g, ' ');
  }).join(' / ');
}

export function annotateRoutePaths(nodes: Test[], parentSegments: string[] = []): Test[] {
  const counts = new Map<string, number>();
  const slugs = nodes.map(node => {
    const slug = slugify(displayLabel(node));
    counts.set(slug, (counts.get(slug) || 0) + 1);
    return slug;
  });
  const seen = new Map<string, number>();

  return nodes.map((node, index) => {
    const slug = slugs[index];
    const ordinal = (seen.get(slug) || 0) + 1;
    seen.set(slug, ordinal);
    const finalSlug = (counts.get(slug) || 0) > 1 ? `${slug}~${ordinal}` : slug;
    const routePath = [...parentSegments, finalSlug].join('/');
    return {
      ...node,
      route_path: routePath,
      children: node.children ? annotateRoutePaths(node.children, [...parentSegments, finalSlug]) : undefined,
    };
  });
}

export function findNodeByRoutePath(nodes: Test[], target: string): Test | null {
  if (!target) return null;
  for (const node of nodes) {
    if (node.route_path === target) return node;
    if (node.children) {
      const child = findNodeByRoutePath(node.children, target);
      if (child) return child;
    }
  }
  return null;
}
