import {
  DEFAULT_LINT_GROUPING,
  LINT_GROUPINGS,
  decodeFilterState,
  encodeFilterState,
  type LintFilters,
  type LintGrouping,
  type LintSeverity,
  type TestFilters,
} from '@flanksource/clicky-ui/data';
import type { FileReviewView, ProblemFilters } from '@flanksource/clicky-ui/file-findings';
import { basePath } from './config';

export type TabKey = 'tests' | 'lint' | 'files' | 'bench' | 'processes' | 'history';

const TAB_KEYS: readonly TabKey[] = ['tests', 'lint', 'files', 'bench', 'processes', 'history'];

export interface RouteState {
  tab: TabKey;
  // On the Files tab this is the open file's or folder's path relative to `workDir` (empty for a work dir's root folder).
  selectedPath: string;
  // Files tab: the open node is a folder (`?kind=folder`), not a file.
  folder?: boolean;
  // Files tab: the root `selectedPath` is relative to (`?dir=`), the same relative path can exist under several.
  workDir?: string;
  // Files tab: the open node's tab (`?view=`): a built-in or host tab id; preview (Source for code, Problems for a folder) when absent.
  view?: FileReviewView;
  // Files tab: 1-based line to reveal (`?line=`).
  line?: number;
  // Files tab: tri-state Problems filters (`?problemSeverity=`, `?problemType=`).
  problemFilters: ProblemFilters;
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

// Any tab id is kept: FileFindingsView shows a node's first tab for a view it does not have.
function parseView(value: string | null): FileReviewView {
  return value || 'preview';
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
  const line = params.get('line');
  const dir = params.get('dir');
  return {
    tab,
    selectedPath,
    filters: {
      status: decodeFilterState(splitCSV(params.get('status'))),
      framework: decodeFilterState(splitCSV(params.get('framework'))),
    },
    lintGrouping: parseLintGrouping(params.get('grouping')),
    lintFilters: {
      severity: decodeFilterState<LintSeverity>(splitCSV(params.get('severity'))),
      linter: decodeFilterState(splitCSV(params.get('linter'))),
    },
    problemFilters: {
      severity: decodeFilterState<LintSeverity>(splitCSV(params.get('problemSeverity'))),
      type: decodeFilterState(splitCSV(params.get('problemType'))),
    },
    sampleIndex: sample !== null && /^\d+$/.test(sample) ? Number(sample) : undefined,
    hideExited: params.get('hideExited') === '1',
    ...(dir ? { workDir: dir } : {}),
    ...(tab === 'files' && params.get('kind') === 'folder' ? { folder: true } : {}),
    view: parseView(params.get('view')),
    line: line !== null && /^\d+$/.test(line) && Number(line) > 0 ? Number(line) : undefined,
  };
}

export function buildRoute(state: RouteState): string {
  const segments: string[] = [state.tab];
  if (state.selectedPath) segments.push(...state.selectedPath.split('/').map(encodeURIComponent));

  const params = new URLSearchParams();
  if (state.tab === 'tests') {
    if (state.filters.status.size > 0) params.set('status', encodeFilterState(state.filters.status).join(','));
    if (state.filters.framework.size > 0) params.set('framework', encodeFilterState(state.filters.framework).join(','));
  }
  if (state.tab === 'lint') {
    if (state.lintGrouping !== DEFAULT_LINT_GROUPING) params.set('grouping', state.lintGrouping);
    if (state.lintFilters.severity.size > 0) params.set('severity', encodeFilterState(state.lintFilters.severity).join(','));
    if (state.lintFilters.linter.size > 0) params.set('linter', encodeFilterState(state.lintFilters.linter).join(','));
  }
  if (state.tab === 'files') {
    if (state.workDir) params.set('dir', state.workDir);
    if (state.folder) params.set('kind', 'folder');
    if (state.view && state.view !== 'preview') params.set('view', state.view);
    if (state.line !== undefined) params.set('line', String(state.line));
    if (state.problemFilters.severity.size > 0) params.set('problemSeverity', encodeFilterState(state.problemFilters.severity).join(','));
    if (state.problemFilters.type.size > 0) params.set('problemType', encodeFilterState(state.problemFilters.type).join(','));
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
