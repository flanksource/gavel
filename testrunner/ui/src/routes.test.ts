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
