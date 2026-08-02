import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TaskProgressProps, TaskSnapshot } from '@flanksource/clicky-ui/data';
import { Button } from '@flanksource/clicky-ui/components';
import { queryKeys } from '../query';
import { ProjectCommitTasks } from './ProjectCommitTasks';
import { projectCommitTaskKeys } from './project-commit-tasks';

vi.mock('@flanksource/clicky-ui/data', async importOriginal => {
  const actual = await importOriginal<typeof import('@flanksource/clicky-ui/data')>();

  return {
    ...actual,
    TaskProgress: ({ snapshots: taskSnapshots, onControl, onTaskControl }: TaskProgressProps) => {
      const group = taskSnapshots.find(snapshot => snapshot.type === 'group');
      const task = taskSnapshots.find(snapshot => snapshot.type === 'task');

      return (
        <div>
          {group && <Button type="button" onClick={() => onControl?.('stop', group)}>Stop group</Button>}
          {group && task && <Button type="button" onClick={() => onTaskControl?.('stop', task, group)}>Stop task</Button>}
        </div>
      );
    },
  };
});

const snapshots: TaskSnapshot[] = [
  {
    id: 'commit-run-1',
    groupId: 'commit-run-1',
    name: 'Commit gavel',
    type: 'group',
    status: 'running',
    controls: ['stop'],
    details: { entries: [{ taskId: 'commit-one', action: 'commit', files: ['one.go'] }] },
  },
  {
    id: 'commit-one',
    groupId: 'commit-run-1',
    name: 'Commit one.go',
    type: 'task',
    status: 'running',
    controls: ['stop'],
  },
];

beforeEach(() => {
  vi.stubGlobal('EventSource', undefined);
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('ProjectCommitTasks', () => {
  it('reports locked files and routes group and child controls through the generic task API', async () => {
    const client = createClient();
    const invalidateQueries = vi.spyOn(client, 'invalidateQueries');
    const onLockedFilesChange = vi.fn();
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (init?.method === 'POST') return { ok: true } as Response;
      if (url.includes('/tasks?')) {
        return {
          ok: true,
          json: async () => [{ id: 'commit-run-1', name: 'Commit gavel', status: 'running', total: 1, completed: 0, failed: 0, running: 1 }],
        } as Response;
      }
      return { ok: true, json: async () => snapshots } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    renderWithClient(client,
      <ProjectCommitTasks
        projectName="gavel"
        onLockedFilesChange={onLockedFilesChange}
        onErrorChange={vi.fn()}
        onComplete={vi.fn()}
      />,
    );

    await waitFor(() => expect(onLockedFilesChange).toHaveBeenCalledWith(new Map([['one.go', 1]])));
    fireEvent.click(screen.getByRole('button', { name: 'Stop group' }));
    fireEvent.click(screen.getByRole('button', { name: 'Stop task' }));

    await waitFor(() => expect(fetchMock.mock.calls.filter(([, init]) => init?.method === 'POST')).toHaveLength(2));
    expect(fetchMock.mock.calls.filter(([, init]) => init?.method === 'POST').map(call => call[0])).toEqual([
      '/api/v1/tasks/commit-run-1/control',
      '/api/v1/tasks/commit-run-1/tasks/commit-one/control',
    ]);
    expect(invalidateQueries).toHaveBeenCalledTimes(2);
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: queryKeys.projectStatus('gavel'), exact: true });
    expect(client.getQueryData(projectCommitTaskKeys.run('gavel', 'commit-run-1'))).toEqual(snapshots);
  });

  it('refreshes the project once when a run becomes complete', async () => {
    const client = createClient();
    const onComplete = vi.fn();
    const terminal = snapshots.map(snapshot => ({ ...snapshot, status: 'success', controls: [] }));
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => ({
      ok: true,
      json: async () => String(input).includes('/tasks?')
        ? [{ id: 'commit-run-1', name: 'Commit gavel', status: 'success', total: 1, completed: 1, failed: 0, running: 0 }]
        : terminal,
    }) as Response));

    const { rerender } = renderWithClient(client,
      <ProjectCommitTasks
        projectName="gavel"
        preferredRunId="commit-run-1"
        onLockedFilesChange={vi.fn()}
        onErrorChange={vi.fn()}
        onComplete={onComplete}
      />,
    );
    rerender(withClient(client,
      <ProjectCommitTasks
        projectName="gavel"
        preferredRunId="commit-run-1"
        onLockedFilesChange={vi.fn()}
        onErrorChange={vi.fn()}
        onComplete={onComplete}
      />,
    ));

    await waitFor(() => expect(onComplete).toHaveBeenCalledTimes(1));
    expect(client.getQueryData(projectCommitTaskKeys.run('gavel', 'commit-run-1'))).toEqual(terminal);
  });

  it('does not refresh project status for a historical completed run on mount', async () => {
    const onComplete = vi.fn();
    const terminal = snapshots.map(snapshot => ({ ...snapshot, status: 'success', controls: [] }));
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => ({
      ok: true,
      json: async () => String(input).includes('/tasks?')
        ? [{ id: 'commit-run-1', name: 'Commit gavel', status: 'success', total: 1, completed: 1, failed: 0, running: 0 }]
        : terminal,
    }) as Response));

    renderWithClient(createClient(),
      <ProjectCommitTasks
        projectName="gavel"
        onLockedFilesChange={vi.fn()}
        onErrorChange={vi.fn()}
        onComplete={onComplete}
      />,
    );

    await waitFor(() => expect(screen.getByRole('region', { name: 'Commit tasks' })).toBeTruthy());
    expect(onComplete).not.toHaveBeenCalled();
  });

  it('coalesces duplicate controls while a matching mutation is pending', async () => {
    const client = createClient();
    const invalidateQueries = vi.spyOn(client, 'invalidateQueries');
    let completeControl: ((response: Response) => void) | undefined;
    const pendingControl = new Promise<Response>(resolve => {
      completeControl = resolve;
    });
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') return pendingControl;
      return {
        ok: true,
        json: async () => String(input).includes('/tasks?')
          ? [{ id: 'commit-run-1', name: 'Commit gavel', status: 'running', total: 1, completed: 0, failed: 0, running: 1 }]
          : snapshots,
      } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    renderWithClient(client,
      <ProjectCommitTasks
        projectName="gavel"
        onLockedFilesChange={vi.fn()}
        onErrorChange={vi.fn()}
        onComplete={vi.fn()}
      />,
    );

    const stopGroup = await screen.findByRole('button', { name: 'Stop group' });
    fireEvent.click(stopGroup);
    fireEvent.click(stopGroup);

    await waitFor(() => expect(fetchMock.mock.calls.filter(([, init]) => init?.method === 'POST')).toHaveLength(1));
    completeControl?.({ ok: true } as Response);
    await waitFor(() => expect(invalidateQueries).toHaveBeenCalledTimes(1));
  });

  it('surfaces task-control failures in the task panel and to its owner', async () => {
    const onErrorChange = vi.fn();
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return { ok: false, status: 409, text: async () => 'task is already stopped' } as Response;
      }
      return {
        ok: true,
        json: async () => String(input).includes('/tasks?')
          ? [{ id: 'commit-run-1', name: 'Commit gavel', status: 'running', total: 1, completed: 0, failed: 0, running: 1 }]
          : snapshots,
      } as Response;
    }));

    renderWithClient(createClient(),
      <ProjectCommitTasks
        projectName="gavel"
        onLockedFilesChange={vi.fn()}
        onErrorChange={onErrorChange}
        onComplete={vi.fn()}
      />,
    );

    fireEvent.click(await screen.findByRole('button', { name: 'Stop group' }));

    expect((await screen.findByRole('alert')).textContent).toContain('Failed to stop commit task: task is already stopped');
    await waitFor(() => expect(onErrorChange).toHaveBeenLastCalledWith('Failed to stop commit task: task is already stopped'));
  });

  it('owns one preferred-run stream, mirrors frames into the query cache, and closes it on unmount', async () => {
    const client = createClient();
    const eventSources: TestEventSource[] = [];
    class TestEventSource {
      readonly listeners = new Map<string, (event: MessageEvent<string>) => void>();
      readonly close = vi.fn();
      onerror: (() => void) | null = null;

      constructor(readonly url: string) {
        eventSources.push(this);
      }

      addEventListener(type: string, listener: EventListenerOrEventListenerObject) {
        this.listeners.set(type, listener as (event: MessageEvent<string>) => void);
      }

      emit(type: string, data: unknown) {
        this.listeners.get(type)?.({ data: JSON.stringify(data) } as MessageEvent<string>);
      }
    }
    vi.stubGlobal('EventSource', TestEventSource);

    const { rerender, unmount } = renderWithClient(client,
      <ProjectCommitTasks
        projectName="gavel"
        preferredRunId="commit-run-1"
        onLockedFilesChange={vi.fn()}
        onErrorChange={vi.fn()}
        onComplete={vi.fn()}
      />,
    );

    expect(eventSources.map(source => source.url)).toEqual(['/api/v1/tasks/stream?tasks=commit-run-1']);
    act(() => {
      eventSources[0]?.emit('task', snapshots[0]);
      eventSources[0]?.emit('task', snapshots[1]);
    });
    await waitFor(() => expect(client.getQueryData(projectCommitTaskKeys.run('gavel', 'commit-run-1'))).toEqual(snapshots));

    rerender(withClient(client,
      <ProjectCommitTasks
        projectName="gavel"
        preferredRunId="commit-run-1"
        onLockedFilesChange={vi.fn()}
        onErrorChange={vi.fn()}
        onComplete={vi.fn()}
      />,
    ));
    expect(eventSources).toHaveLength(1);

    unmount();
    expect(eventSources[0]?.close).toHaveBeenCalledTimes(1);
  });
});

function createClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: Infinity },
      mutations: { retry: false },
    },
  });
}

function withClient(client: QueryClient, children: React.ReactNode) {
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

function renderWithClient(client: QueryClient, children: React.ReactNode) {
  return render(withClient(client, children));
}
