import { describe, expect, it } from 'vitest';
import { buildRoute, emptyRouteState, parseRoute } from './routes';

describe('dashboard project scope routes', () => {
  it('round-trips the selected project on the todos route', () => {
    const parsed = parseRoute(new URL('http://localhost:9092/todos?project=Clicky%20UI') as unknown as Location);

    expect(parsed.scopeProject).toBe('Clicky UI');
    expect(buildRoute(parsed)).toBe('/todos?project=Clicky+UI');
  });

  it('keeps the selected project alongside pull request filters', () => {
    const parsed = parseRoute(new URL('http://localhost:9092/prs?project=gavel&state=open') as unknown as Location);

    expect(parsed.scopeProject).toBe('gavel');
    expect(buildRoute(parsed)).toBe('/prs?project=gavel&state=open');
  });
});

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

describe('prompt routes', () => {
  it('round-trips a selected prompt with its scope project', () => {
    const location = new URL('http://localhost:9092/prompts/commit.message?project=Clicky%20UI');

    const parsed = parseRoute(location as unknown as Location);

    expect(parsed).toEqual({
      ...emptyRouteState(),
      tab: 'prompts',
      selectedPath: 'commit.message',
      scopeProject: 'Clicky UI',
    });
    expect(buildRoute(parsed)).toBe('/prompts/commit.message?project=Clicky+UI');
  });

  it('keeps the prompts table at the global scope without a query', () => {
    const parsed = parseRoute(new URL('http://localhost:9092/prompts') as unknown as Location);

    expect(parsed).toEqual({ ...emptyRouteState(), tab: 'prompts' });
    expect(buildRoute(parsed)).toBe('/prompts');
  });

  it('treats the project query as dashboard scope on every tab', () => {
    const parsed = parseRoute(new URL('http://localhost:9092/todos?project=x') as unknown as Location);

    expect(parsed.scopeProject).toBe('x');
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

describe('todo detail view routes', () => {
  const parse = (url: string) => parseRoute(new URL(`http://localhost:9092${url}`) as unknown as Location);

  it('round-trips the detail tab, inspector tab, and selected attempts of a todo', () => {
    const url = '/todos/todo-7?tab=session&sessionTab=costs&sessions=run-a%2Crun-b';

    const parsed = parse(url);

    expect(parsed).toEqual({
      ...emptyRouteState(),
      tab: 'todos',
      selectedPath: 'todo-7',
      todoView: { tab: 'session', sessionTab: 'costs', sessionIds: ['run-a', 'run-b'] },
    });
    expect(buildRoute(parsed)).toBe(url);
  });

  it('keeps a slash-containing ref in the path and the view in the query', () => {
    const parsed = parse('/todos/pkg/file.go/todo-3?tab=plan');

    expect(parsed.selectedPath).toBe('pkg/file.go/todo-3');
    expect(parsed.todoView).toEqual({ tab: 'plan' });
    expect(buildRoute(parsed)).toBe('/todos/pkg/file.go/todo-3?tab=plan');
  });

  it('leaves the default view out of the URL', () => {
    expect(buildRoute({ ...emptyRouteState(), tab: 'todos', selectedPath: 'todo-7' })).toBe('/todos/todo-7');
  });

  it('ignores unknown tab names', () => {
    expect(parse('/todos/todo-7?tab=bogus&sessionTab=nope').todoView).toEqual({});
  });

  it('drops the view when no todo is selected', () => {
    const parsed = parse('/todos?tab=session');

    expect(parsed.todoView).toEqual({});
    expect(buildRoute({ ...parsed, todoView: { tab: 'session' } })).toBe('/todos');
  });
});
