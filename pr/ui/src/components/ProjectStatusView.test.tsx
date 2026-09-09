import type React from 'react';
import type { ReactElement, ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { queryKeys } from '../query';
import { projectDiffQueryKey } from './projectMutations';
import { projectStatusResponse as statusResponse } from './ProjectStatusView.fixture';
import { ProjectStatusView } from './ProjectStatusView';

const { useTestRunMock } = vi.hoisted(() => ({ useTestRunMock: vi.fn() }));
let queryClient: QueryClient;

vi.mock('@flanksource/gavel/testrunner/hooks', () => ({
  useTestRun: useTestRunMock,
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, onClick, loading: _loading, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { loading?: boolean }) => (
    <button onClick={onClick} {...props}>{children}</button>
  ),
  SplitButton: ({ label, onClick, items, title, disabled }: {
    label: ReactNode;
    onClick: () => void;
    items: { label: ReactNode; onSelect?: () => void }[];
    title?: string;
    disabled?: boolean;
  }) => (
    <div>
      <button type="button" onClick={onClick} disabled={disabled}>{label}</button>
      {items.map((item, index) => (
        <button key={index} type="button" aria-label={`${title}: ${items.length > 1 && index === 0 ? 'open PR' : 'advanced'}`} onClick={item.onSelect} disabled={disabled}>
          {item.label}
        </button>
      ))}
    </div>
  ),
  Modal: ({ open, title, children, footer }: { open: boolean; title?: ReactNode; children: ReactNode; footer?: ReactNode }) => open ? (
    <section aria-label={typeof title === 'string' ? title : undefined}>{children}{footer}</section>
  ) : null,
  JsonSchemaForm: ({ value, onChange }: { value: Record<string, unknown>; onChange: (value: Record<string, unknown>) => void }) => (
    <div>
      <output data-testid="advanced-options">{JSON.stringify(value)}</output>
      <button type="button" onClick={() => onChange({ ...value, changed: true, linters: ['golangci-lint'] })}>
        Apply sample options
      </button>
    </div>
  ),
  SplitPane: ({ left, right }: { left: ReactNode; right: ReactNode }) => <div>{left}{right}</div>,
}));

vi.mock('./ProjectCommitTasks', () => ({
  ProjectCommitTasks: ({ preferredRunId, onLockedFilesChange, onComplete }: {
    preferredRunId?: string;
    onLockedFilesChange: (files: Map<string, number>) => void;
    onComplete: () => void;
  }) => preferredRunId ? (
    <section aria-label="Commit tasks">
      <span>{preferredRunId}</span>
      <button type="button" onClick={() => onLockedFilesChange(new Map([['one.go', 1]]))}>Report one.go locked</button>
      <button type="button" onClick={onComplete}>Complete commit tasks</button>
    </section>
  ) : null,
}));

beforeEach(() => {
  queryClient = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: Infinity } } });
  const values = new Map<string, string>();
  vi.stubGlobal('localStorage', {
    clear: () => values.clear(),
    getItem: (key: string) => values.get(key) ?? null,
    removeItem: (key: string) => values.delete(key),
    setItem: (key: string, value: string) => values.set(key, value),
  });
  useTestRunMock.mockReset();
  useTestRunMock.mockReturnValue({
    snapshot: { status: { running: true }, tests: [] },
    tests: [],
    lint: [],
    status: { running: true },
    statusText: 'Running tests...',
    done: false,
    error: undefined,
  });
});

const renderStatus = (view: ReactElement) => render(<QueryClientProvider client={queryClient}>{view}</QueryClientProvider>);
function deferredResponse() {
  let resolve!: (value: Response) => void;
  const promise = new Promise<Response>(complete => { resolve = complete; });
  return { promise, resolve };
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), { status, headers: { 'Content-Type': 'application/json' } });
}

afterEach(() => {
  vi.useRealTimers();
  window.history.replaceState({}, '', '/');
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'visible' });
});

describe('ProjectStatusView', () => {
  it('opens a test-runner detail dialog from a task href', async () => {
    window.history.replaceState({}, '', '/projects/gavel?action=test&run=test-run-123');
    expect(window.location.search).toBe('?action=test&run=test-run-123');
    vi.stubGlobal('fetch', vi.fn(async () => (
      { ok: true, json: async () => statusResponse }
    ) as Response));

    renderStatus(<ProjectStatusView project={statusResponse.project} />);

    expect(await screen.findByLabelText('Test details · gavel')).toBeTruthy();
    expect(useTestRunMock).toHaveBeenCalledWith(expect.objectContaining({ runId: 'test-run-123' }));
  });

  it('selects status files and queues a file-scoped gavel commit', async () => {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return jsonResponse({ runId: 'commit-run-1' }, 202);
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    renderStatus(<ProjectStatusView project={statusResponse.project} />);

    fireEvent.click(await screen.findByRole('checkbox', { name: 'Select one.go' }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected (1)' }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    expect(fetchMock.mock.calls[1]?.[0]).toBe('/api/projects/gavel/commit-queue');
    expect(JSON.parse(String(fetchMock.mock.calls[1]?.[1]?.body))).toEqual({
      action: 'commit',
      files: ['one.go'],
    });
    const tasks = await screen.findByRole('region', { name: 'Commit tasks' });
    expect(within(tasks).getByText('commit-run-1')).toBeTruthy();
  });

  it('keeps queueing enabled while a commit runs and locks its files out of the next group', async () => {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return jsonResponse({ runId: 'commit-run-1' }, 202);
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    renderStatus(<ProjectStatusView project={statusResponse.project} />);

    fireEvent.click(await screen.findByRole('checkbox', { name: 'Select one.go' }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected (1)' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Report one.go locked' }));

    const lockedRow = screen.getByRole('treeitem', { name: 'one.go' });
    expect(within(lockedRow).getByText('queued in group 1')).toBeTruthy();
    expect(within(lockedRow).getByRole('checkbox', { name: 'Select one.go' }).hasAttribute('disabled')).toBe(true);

    fireEvent.click(screen.getByRole('checkbox', { name: 'Select generated/output.log' }));
    const commit = screen.getByRole('button', { name: 'Commit selected (1)' });
    expect(commit.hasAttribute('disabled')).toBe(false);
    fireEvent.click(commit);

    await waitFor(() => expect(fetchMock.mock.calls.filter(([, init]) => init?.method === 'POST')).toHaveLength(2));
    expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'DELETE')).toBe(false);
  });

  it('keeps conflicted files out of the selectable lifecycle', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => (
      { ok: true, json: async () => statusResponse }
    ) as Response));

    renderStatus(<ProjectStatusView project={statusResponse.project} />);

    expect((await screen.findByRole('checkbox', { name: 'Select conflicted.go' })).hasAttribute('disabled')).toBe(true);
    expect(screen.getByRole('button', { name: 'Commit selected (0)' }).hasAttribute('disabled')).toBe(true);
  });

  it('renders changed paths as a file tree and ignores an untracked directory', async () => {
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return jsonResponse({ path: 'generated', directory: true, rule: '/generated/', added: true });
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    renderStatus(<ProjectStatusView project={statusResponse.project} />);

    expect(await screen.findByRole('tree', { name: 'Project files' })).toBeTruthy();
    expect(screen.getByRole('treeitem', { name: 'generated' })).toBeTruthy();
    expect(screen.getByRole('treeitem', { name: 'generated/cache' })).toBeTruthy();
    expect(screen.getByRole('treeitem', { name: 'generated/cache/data.json' })).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Ignore directory generated' }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(3));
    expect(fetchMock.mock.calls[1]?.[0]).toBe('/api/projects/gavel/ignore');
    expect(JSON.parse(String(fetchMock.mock.calls[1]?.[1]?.body))).toEqual({
      path: 'generated',
      directory: true,
    });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: queryKeys.projectStatusScope('gavel') });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: projectDiffQueryKey('gavel') });
  });

  it('surfaces contextual project action errors', async () => {
    vi.stubGlobal('fetch', vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return new Response(JSON.stringify({ error: 'runner unavailable' }), {
          status: 503,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      return new Response(JSON.stringify(statusResponse), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      });
    }));

    renderStatus(<ProjectStatusView project={statusResponse.project} />);

    fireEvent.click(await screen.findByRole('button', { name: 'Lint project' }));
    expect((await screen.findByRole('alert')).textContent).toContain('Failed to start lint for gavel: runner unavailable');
  });

  it('invalidates only project status and diffs when queued commit work completes', async () => {
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    vi.stubGlobal('fetch', vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return new Response(JSON.stringify({ runId: 'commit-run-1' }), {
          status: 202,
          headers: { 'Content-Type': 'application/json' },
        });
      }
      return new Response(JSON.stringify(statusResponse), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      });
    }));

    renderStatus(<ProjectStatusView project={statusResponse.project} />);
    fireEvent.click(await screen.findByRole('checkbox', { name: 'Select one.go' }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected (1)' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Complete commit tasks' }));

    await waitFor(() => {
      expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: queryKeys.projectStatusScope('gavel') });
      expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: projectDiffQueryKey('gavel') });
    });
    expect(invalidateQueries).toHaveBeenCalledTimes(2);
  });

  it('keeps diff counts on the file heading and uses extension-specific file icons', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => (
      { ok: true, json: async () => statusResponse }
    ) as Response));

    renderStatus(<ProjectStatusView project={statusResponse.project} />);

    const goRow = await screen.findByRole('treeitem', { name: 'one.go' });
    const goHeading = goRow.querySelector('[data-file-primary]');
    expect(goHeading).not.toBeNull();
    expect(within(goHeading as HTMLElement).getByText('+12')).toBeTruthy();
    expect(within(goHeading as HTMLElement).getByText('−3')).toBeTruthy();
    expect(goRow.querySelector('[data-file-icon="go"]')).not.toBeNull();

    const jsonRow = screen.getByRole('treeitem', { name: 'generated/cache/data.json' });
    expect(jsonRow.querySelector('[data-file-icon="json"]')).not.toBeNull();
  });

  it('loads and renders snapshot enrichment only for the hidden result option', async () => {
    const enrichedStatus = {
      ...statusResponse,
      resultsStale: true,
      files: statusResponse.files.map((file, index) => index === 0 ? { ...file, resultsStale: true } : file),
    };
    const fetchMock = vi.fn(async () => jsonResponse(enrichedStatus));
    vi.stubGlobal('fetch', fetchMock);

    const base = renderStatus(<ProjectStatusView project={statusResponse.project} showResults={false} />);
    expect(await screen.findByRole('tree', { name: 'Project files' })).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledWith('/api/projects/gavel/status', expect.anything());
    expect(screen.queryByText('4 passed')).toBeNull();
    expect(screen.queryByText('1 lint')).toBeNull();
    expect(screen.queryByText('stale results')).toBeNull();
    expect(screen.queryByText('Test or lint results are from an earlier commit.')).toBeNull();
    base.unmount();

    renderStatus(<ProjectStatusView project={statusResponse.project} showResults />);
    expect(await screen.findByText('4 passed')).toBeTruthy();
    expect(screen.getAllByText('1 lint')).toHaveLength(2);
    expect(screen.getByText('stale results')).toBeTruthy();
    expect(screen.getByText('Test or lint results are from an earlier commit.')).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledWith('/api/projects/gavel/status?includeResults=true', expect.anything());
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('opens schema-driven lint options and remembers the accepted configuration per project', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url.endsWith('/actions/schema?action=lint')) {
        return {
          ok: true,
          json: async () => ({
            schemaVersion: 1,
            action: 'lint',
            schema: {
              type: 'object',
              properties: {
                changed: { type: 'boolean', default: false },
                linters: { type: 'array', items: { type: 'string' } },
                files: { type: 'array', items: { type: 'string' } },
              },
            },
            defaults: { changed: false },
          }),
        } as Response;
      }
      if (init?.method === 'POST') {
        return jsonResponse({ action: 'lint', runId: 'advanced-lint-run', running: false }, 202);
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    renderStatus(<ProjectStatusView project={statusResponse.project} />);

    fireEvent.click(await screen.findByRole('checkbox', { name: 'Select one.go' }));
    fireEvent.click(screen.getByRole('button', { name: 'Lint options: advanced' }));
    expect(await screen.findByLabelText('Advanced lint options')).toBeTruthy();
    expect((await screen.findByTestId('advanced-options')).textContent).toContain('"files":["one.go"]');
    fireEvent.click(screen.getByRole('button', { name: 'Apply sample options' }));
    fireEvent.click(screen.getByRole('button', { name: 'Run lint' }));

    await waitFor(() => expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(true));
    const actionCall = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST');
    expect(JSON.parse(String(actionCall?.[1]?.body))).toEqual({
      action: 'lint',
      options: {
        changed: true,
        files: ['one.go'],
        linters: ['golangci-lint'],
      },
    });
    expect(localStorage.getItem('gavel.project-action-options.v1')).toContain('golangci-lint');

    fireEvent.click(screen.getByRole('button', { name: 'Lint options: advanced' }));
    expect((await screen.findByTestId('advanced-options')).textContent).toContain('golangci-lint');
  });

  it('opens schema-driven commit options with the current selection', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url.endsWith('/actions/schema?action=commit')) {
        return {
          ok: true,
          json: async () => ({
            schemaVersion: 1,
            action: 'commit',
            schema: {
              type: 'object',
              properties: {
                files: { type: 'array', items: { type: 'string' } },
                message: { type: 'string' },
                precommit: { type: 'string' },
              },
            },
            defaults: { precommit: 'fail' },
          }),
        } as Response;
      }
      if (init?.method === 'POST') {
        return jsonResponse({ runId: 'commit-run-1' }, 202);
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    renderStatus(<ProjectStatusView project={statusResponse.project} />);

    fireEvent.click(await screen.findByRole('checkbox', { name: 'Select one.go' }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit options: advanced' }));
    expect(await screen.findByLabelText('Advanced commit options')).toBeTruthy();
    expect((await screen.findByTestId('advanced-options')).textContent).toContain('"files":["one.go"]');
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected' }));

    await waitFor(() => expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(true));
    const actionCall = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST');
    expect(String(actionCall?.[0])).toBe('/api/projects/gavel/commit-queue');
    expect(JSON.parse(String(actionCall?.[1]?.body))).toEqual({
      action: 'commit',
      options: {
        files: ['one.go'],
        precommit: 'fail',
      },
    });
    expect(await screen.findByRole('region', { name: 'Commit tasks' })).toBeTruthy();
  });

  it.each([
    ['test', 'Test changed', 'Test details · gavel'],
    ['lint', 'Lint project', 'Lint details · gavel'],
  ] as const)('opens %s in the test-runner dialog backed by its SSE run id', async (action, button, title) => {
    const runId = `${action}-run-id`;
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return jsonResponse({ action, runId, running: true }, 202);
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    renderStatus(<ProjectStatusView project={statusResponse.project} />);

    fireEvent.click(await screen.findByRole('button', { name: button }));

    expect(await screen.findByLabelText(title)).toBeTruthy();
    expect(useTestRunMock).toHaveBeenCalledWith({
      baseUrl: '/api/project-runs',
      enabled: true,
      runId,
    });
  });

  it('selects file and folder rows for diff without coupling checkbox clicks', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => (
      { ok: true, json: async () => statusResponse }
    ) as Response));
    const onDiffPathChange = vi.fn();

    renderStatus(<ProjectStatusView project={statusResponse.project} onDiffPathChange={onDiffPathChange} />);

    const fileRow = await screen.findByRole('treeitem', { name: 'one.go' });
    fireEvent.click(within(fileRow).getByText('one.go'));
    expect(onDiffPathChange).toHaveBeenLastCalledWith('one.go');

    onDiffPathChange.mockClear();
    fireEvent.click(within(fileRow).getByRole('checkbox', { name: 'Select one.go' }));
    expect(onDiffPathChange).not.toHaveBeenCalled();

    const folderRow = screen.getByRole('treeitem', { name: 'generated' });
    fireEvent.click(within(folderRow).getByText('generated'));
    expect(onDiffPathChange).toHaveBeenLastCalledWith('generated');
  });

  it('loads the selected working-tree diff into the right pane', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/diff?path=one.go')) {
        return { ok: true, json: async () => ({ path: 'one.go', diff: 'diff --git a/one.go b/one.go\n-old\n+new\n' }) } as Response;
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    renderStatus(<ProjectStatusView project={statusResponse.project} diffPath="one.go" />);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith('/api/projects/gavel/diff?path=one.go', expect.anything()));
    expect(await screen.findByRole('heading', { name: 'one.go' })).toBeTruthy();
  });

  it('reuses cached project status after remount', async () => {
    const fetchMock = vi.fn(async () => ({ ok: true, json: async () => statusResponse }) as Response);
    vi.stubGlobal('fetch', fetchMock);

    const first = renderStatus(<ProjectStatusView project={statusResponse.project} />);
    expect(await screen.findByText('feature/projects')).toBeTruthy();
    first.unmount();

    renderStatus(<ProjectStatusView project={statusResponse.project} />);
    expect(await screen.findByText('feature/projects')).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('does not overlap status polls and stops when the action becomes terminal', async () => {
    vi.useFakeTimers();
    const terminal = deferredResponse();
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: async () => ({ ...statusResponse, action: { running: true } }) } as Response)
      .mockImplementation(() => terminal.promise);
    vi.stubGlobal('fetch', fetchMock);

    renderStatus(<ProjectStatusView project={statusResponse.project} />);
    await act(async () => undefined);
    expect(fetchMock).toHaveBeenCalledTimes(1);

    await act(async () => vi.advanceTimersByTimeAsync(3_000));
    expect(fetchMock).toHaveBeenCalledTimes(2);

    terminal.resolve({ ok: true, json: async () => ({ ...statusResponse, action: { running: false } }) } as Response);
    await act(async () => undefined);
    await act(async () => vi.advanceTimersByTimeAsync(3_000));
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('pauses running-action status polling while hidden and refreshes when visible', async () => {
    vi.useFakeTimers();
    const fetchMock = vi.fn(async () => ({
      ok: true,
      json: async () => ({ ...statusResponse, action: { running: true } }),
    }) as Response);
    vi.stubGlobal('fetch', fetchMock);

    renderStatus(<ProjectStatusView project={statusResponse.project} />);
    await act(async () => undefined);
    expect(fetchMock).toHaveBeenCalledTimes(1);

    act(() => {
      Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'hidden' });
      document.dispatchEvent(new Event('visibilitychange'));
    });
    await act(async () => vi.advanceTimersByTimeAsync(3_000));
    expect(fetchMock).toHaveBeenCalledTimes(1);

    act(() => {
      Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'visible' });
      document.dispatchEvent(new Event('visibilitychange'));
    });
    await act(async () => undefined);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});
