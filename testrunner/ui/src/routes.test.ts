import { describe, expect, it } from 'vitest';

// vitest runs in node by default; routes.ts -> config.ts touches `window`
// at import time. Stub it before importing.
(globalThis as any).window = (globalThis as any).window || {};

const { buildRoute, defaultStatusFilter, parseRoute } = await import('./routes');

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
      view: 'run' as const,
      runName: '',
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

describe('run index routes', () => {
  it('parses / as the index view', () => {
    const route = parseRoute(new URL('http://example.test/') as unknown as Location);
    expect(route.view).toBe('index');
    expect(route.runName).toBe('');
  });

  it('parses /run/<name> as the run view with default tab', () => {
    const route = parseRoute(new URL('http://example.test/run/sha-abc123') as unknown as Location);
    expect(route.view).toBe('run');
    expect(route.runName).toBe('sha-abc123');
    expect(route.tab).toBe('tests');
  });

  it('parses /run/<name>/lint/<selected>?status=failed', () => {
    const route = parseRoute(new URL('http://example.test/run/last/lint/golangci?status=failed') as unknown as Location);
    expect(route.view).toBe('run');
    expect(route.runName).toBe('last');
    expect(route.tab).toBe('lint');
    expect(route.selectedPath).toBe('golangci');
  });

  it('builds /run/<name>/<tab> for run-detail URLs', () => {
    expect(buildRoute({
      view: 'run',
      runName: 'sha-abc123',
      tab: 'tests',
      selectedPath: '',
      filters: { status: defaultStatusFilter(), framework: new Map() },
      lintGrouping: 'linter-rule-file',
      lintFilters: { severity: new Map(), linter: new Map() },
    })).toBe('/run/sha-abc123/tests');
  });

  it('builds / for the index view regardless of other state', () => {
    expect(buildRoute({
      view: 'index',
      runName: '',
      tab: 'tests',
      selectedPath: '',
      filters: { status: defaultStatusFilter(), framework: new Map() },
      lintGrouping: 'linter-rule-file',
      lintFilters: { severity: new Map(), linter: new Map() },
    })).toBe('/');
  });

  it('keeps backward-compat: /tests still parses as a run view with empty runName', () => {
    const route = parseRoute(new URL('http://example.test/tests/foo') as unknown as Location);
    expect(route.view).toBe('run');
    expect(route.runName).toBe('');
    expect(route.tab).toBe('tests');
    expect(route.selectedPath).toBe('foo');
  });
});
