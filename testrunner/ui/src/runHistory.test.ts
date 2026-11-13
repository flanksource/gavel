import { describe, expect, it } from 'vitest';
import { runResultLabel, runViewHref, type RunInfo } from './runHistory';

describe('fixture history labels', () => {
  const fixture: RunInfo = {
    runId: 'run-example', path: 'run-example.json', href: '/runs/run-example/', kind: 'fixture',
    passed: 3, failed: 1, total: 4, skipped: 0, warned: 0, lintViolations: 0, lintLinters: 0,
  };

  it('reports fixture failures without adding a lint result', () => {
    expect(runResultLabel(fixture)).toBe('1 failed / 4 tests');
  });

  it('opens fixture results on the test view', () => {
    expect(runViewHref(fixture)).toBe('/runs/run-example/');
  });

  it('keeps lint results on combined runs', () => {
    expect(runResultLabel({ ...fixture, kind: 'test+lint', lintViolations: 2 })).toBe('1 failed / 4 tests · 2 lint');
  });
});
