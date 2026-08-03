import { describe, expect, it } from 'vitest';
import { buildRoute, emptyRouteState, parseRoute } from './routes';

describe('project routes', () => {
  it('round-trips a selected project as a dedicated top-level tab', () => {
    const location = new URL('http://localhost:9092/projects/Clicky%20UI');

    const parsed = parseRoute(location as unknown as Location);

    expect(parsed).toEqual({
      ...emptyRouteState(),
      tab: 'projects',
      selectedPath: 'Clicky UI',
    });
    expect(buildRoute(parsed)).toBe('/projects/Clicky%20UI');
  });

  it('round-trips the selected project diff as a query parameter', () => {
    const location = new URL('http://localhost:9092/projects/gavel?diff=pr%2Fui%2Fsrc');

    const parsed = parseRoute(location as unknown as Location);

    expect(parsed.projectDiffPath).toBe('pr/ui/src');
    expect(buildRoute(parsed)).toBe('/projects/gavel?diff=pr%2Fui%2Fsrc');
  });

  it('round-trips independent project history and result options with a diff', () => {
    const location = new URL('http://localhost:9092/projects/gavel?diff=pr%2Fui%2Fsrc&history=true&results=true');

    const parsed = parseRoute(location as unknown as Location);

    expect(parsed.projectHistory).toBe(true);
    expect(parsed.projectResults).toBe(true);
    expect(buildRoute(parsed)).toBe('/projects/gavel?diff=pr%2Fui%2Fsrc&history=true&results=true');
  });

  it('round-trips a historical project run and drops the diff selection', () => {
    const location = new URL('http://localhost:9092/projects/Clicky%20UI/runs/run-2026-07-21T10-59-33Z?diff=src&results=true');

    const parsed = parseRoute(location as unknown as Location);

    expect(parsed).toEqual({
      ...emptyRouteState(),
      tab: 'projects',
      selectedPath: 'Clicky UI',
      projectRunId: 'run-2026-07-21T10-59-33Z',
      projectHistory: true,
      projectResults: true,
    });
    expect(buildRoute(parsed)).toBe('/projects/Clicky%20UI/runs/run-2026-07-21T10-59-33Z?results=true');
  });
});

describe('task routes', () => {
  it('round-trips a selected task generation', () => {
    const location = new URL('http://localhost:9092/tasks/run-123');

    const parsed = parseRoute(location as unknown as Location);

    expect(parsed).toEqual({
      ...emptyRouteState(),
      tab: 'tasks',
      selectedPath: 'run-123',
    });
    expect(buildRoute(parsed)).toBe('/tasks/run-123');
  });
});
