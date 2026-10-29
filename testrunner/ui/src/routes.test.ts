import { describe, expect, it } from 'vitest';
import { buildExportRoute, buildRoute, defaultStatusFilter, parseRoute, type RouteState } from './routes';

const at = (url: string) => parseRoute(new URL(`http://example.test${url}`));

describe('lint route grouping', () => {
  it('defaults lint routes to linter-rule-file grouping', () => {
    const route = at('/lint');
    expect(route.tab).toBe('lint');
    expect(route.lintGrouping).toBe('linter-rule-file');
  });

  it('preserves legacy grouping values from the query string', () => {
    expect(at('/lint?grouping=file-linter-rule').lintGrouping).toBe('file-linter-rule');
  });

  it('falls back to the default grouping for a value that is not one', () => {
    expect(at('/lint?grouping=nonsense').lintGrouping).toBe('linter-rule-file');
  });

  it('omits the default grouping from lint URLs and keeps the others', () => {
    const baseState: Omit<RouteState, 'lintGrouping'> = {
      tab: 'lint',
      selectedPath: '',
      filters: { status: defaultStatusFilter(), framework: new Map() },
      lintFilters: { severity: new Map(), linter: new Map() },
    };

    expect(buildRoute({ ...baseState, lintGrouping: 'linter-rule-file' })).toBe('/lint');
    expect(buildRoute({ ...baseState, lintGrouping: 'linter-file' })).toBe('/lint?grouping=linter-file');
  });

  it('round trips severity and linter filters with their include/exclude marks', () => {
    const route = at('/lint/golangci?severity=error,!info&linter=eslint');
    expect(route.selectedPath).toBe('golangci');
    expect(route.lintFilters.severity.get('error')).toBe('include');
    expect(route.lintFilters.severity.get('info')).toBe('exclude');
    expect(route.lintFilters.linter.get('eslint')).toBe('include');
    expect(buildRoute(route)).toBe('/lint/golangci?severity=error%2C%21info&linter=eslint');
  });
});

describe('test status route', () => {
  it('defaults to the failing-only filter and omits it from the URL', () => {
    const route = at('/tests/pkg/test-foo');
    expect(route.selectedPath).toBe('pkg/test-foo');
    expect(route.filters.status.get('passed')).toBe('exclude');
    expect(route.filters.status.get('skipped')).toBe('exclude');
    expect(buildRoute(route)).toBe('/tests/pkg/test-foo');
  });

  it('writes status=all for no filter and reads it back as no filter', () => {
    const route = at('/tests?status=all');
    expect(route.filters.status.size).toBe(0);
    expect(buildRoute(route)).toBe('/tests?status=all');
  });

  it('round trips an explicit status and framework filter', () => {
    const route = at('/tests?status=failed,timedout&framework=ginkgo');
    expect(buildRoute(route)).toBe('/tests?status=failed%2Ctimedout&framework=ginkgo');
  });

  it('builds the export URL for the same view with the format as the extension', () => {
    expect(buildExportRoute(at('/tests/pkg?status=all'), 'md')).toBe('/tests/pkg.md?status=all');
  });
});

describe('fixture profile sample route', () => {
  it('round trips a selected sample index on the Bench tab', () => {
    const route = at('/bench?sample=3');
    expect(route.sampleIndex).toBe(3);
    expect(buildRoute(route)).toBe('/bench?sample=3');
  });
});

describe('files route', () => {
  const workDir = '/Users/dev/my repo';
  const file = 'docs/getting started/intro.md';
  // URLSearchParams form-encodes: "/" as %2F and a space as "+".
  const DIR_PARAM = '%2FUsers%2Fdev%2Fmy+repo';

  it('round trips the work dir, a relative file with spaces, the view and the line', () => {
    const url = `/files/docs/getting%20started/intro.md?dir=${DIR_PARAM}&view=source&line=42`;
    const route = at(url);
    expect(route.tab).toBe('files');
    expect(route.selectedPath).toBe(file);
    expect(route.workDir).toBe(workDir);
    expect(route.view).toBe('source');
    expect(route.line).toBe(42);
    expect(buildRoute(route)).toBe(url);
  });

  it('defaults to the preview and omits it and an absent line from the URL', () => {
    const route = at(`/files/a.go?dir=${DIR_PARAM}`);
    expect(route.view).toBe('preview');
    expect(route.line).toBeUndefined();
    expect(buildRoute(route)).toBe(`/files/a.go?dir=${DIR_PARAM}`);
  });

  it('ignores a line that is not a positive integer and a view that is not one', () => {
    expect(at('/files/a.go?line=0').line).toBeUndefined();
    expect(at('/files/a.go?line=x').line).toBeUndefined();
    expect(at('/files/a.go?view=nonsense').view).toBe('preview');
    expect(at('/files/a.go?view=problems').view).toBe('problems');
  });

  it('does not carry the file query onto other tabs', () => {
    const route = { ...at(`/files/a.go?dir=${DIR_PARAM}&line=3`), tab: 'lint' as const, selectedPath: '' };
    expect(buildRoute(route)).toBe('/lint');
  });
});

describe('processes route', () => {
  it('round trips the processes tab, an encoded node key and the hide-exited filter', () => {
    const route = at('/processes/4242%401?hideExited=1');
    expect(route.tab).toBe('processes');
    expect(route.selectedPath).toBe('4242@1');
    expect(route.hideExited).toBe(true);
    expect(buildRoute(route)).toBe('/processes/4242%401?hideExited=1');
  });

  it('omits hideExited from the URL by default', () => {
    const route = at('/processes');
    expect(route.hideExited).toBe(false);
    expect(buildRoute(route)).toBe('/processes');
  });
});
