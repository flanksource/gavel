import { describe, expect, it } from 'vitest';
import { buildRoute, defaultStatusFilter, parseRoute } from './routes';

describe('lint route grouping', () => {
  it('defaults lint routes to linter-rule-file grouping', () => {
    const location = new URL('http://example.test/lint');
    const route = parseRoute(location as unknown as Location);
    expect(route.tab).toBe('lint');
    expect(route.lintGrouping).toBe('linter-rule-file');
  });

  it('preserves legacy grouping values from the query string', () => {
    const location = new URL('http://example.test/lint?grouping=file-linter-rule');
    const route = parseRoute(location as unknown as Location);
    expect(route.lintGrouping).toBe('file-linter-rule');
  });

  it('omits the default grouping from lint URLs and keeps legacy grouping params', () => {
    const baseState = {
      tab: 'lint' as const,
      selectedPath: '',
      filters: { status: defaultStatusFilter(), framework: new Map() },
      lintFilters: { severity: new Map(), linter: new Map() },
    };

    expect(buildRoute({
      ...baseState,
      lintGrouping: 'linter-rule-file',
    })).toBe('/lint');

    expect(buildRoute({
      ...baseState,
      lintGrouping: 'linter-file',
    })).toBe('/lint?grouping=linter-file');
  });
});

describe('fixture profile sample route', () => {
  it('round trips a selected sample index on the Bench tab', () => {
    const route = parseRoute(new URL('http://example.test/bench?sample=3') as unknown as Location);
    expect(route.sampleIndex).toBe(3);
    expect(buildRoute(route)).toBe('/bench?sample=3');
  });
});

describe('processes route', () => {
  it('round trips the processes tab, an encoded node key and the hide-exited filter', () => {
    const route = parseRoute(new URL('http://example.test/processes/4242%401?hideExited=1') as unknown as Location);
    expect(route.tab).toBe('processes');
    expect(route.selectedPath).toBe('4242@1');
    expect(route.hideExited).toBe(true);
    expect(buildRoute(route)).toBe('/processes/4242%401?hideExited=1');
  });

  it('omits hideExited from the URL by default', () => {
    const route = parseRoute(new URL('http://example.test/processes') as unknown as Location);
    expect(route.hideExited).toBe(false);
    expect(buildRoute(route)).toBe('/processes');
  });
});
