import type React from 'react';
import type { ReactNode } from 'react';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ProjectStatusView } from './ProjectStatusView';

const { useTestRunMock } = vi.hoisted(() => ({ useTestRunMock: vi.fn() }));

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

beforeEach(() => {
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

afterEach(() => {
  window.history.replaceState({}, '', '/');
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

const statusResponse = {
  project: { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] },
  workDir: '/work/gavel',
  branch: 'feature/projects',
  files: [
    {
      path: 'one.go',
      state: 'unstaged',
      workKind: 'M',
      adds: 12,
      dels: 3,
      testStatus: { passed: 4, failed: 0, skipped: 0 },
      lintStatus: { errors: 0, warnings: 1, infos: 0 },
      problems: [],
      resultsStale: false,
    },
    {
      path: 'conflicted.go',
      state: 'conflict',
      workKind: 'M',
      adds: 1,
      dels: 1,
      testStatus: { passed: 0, failed: 1, skipped: 0 },
      lintStatus: { errors: 1, warnings: 0, infos: 0 },
      problems: [],
      resultsStale: false,
      conflictReason: 'unmerged',
    },
    {
      path: 'generated/output.log',
      state: 'untracked',
      workKind: '?',
      adds: 8,
      dels: 0,
      testStatus: { passed: 0, failed: 0, skipped: 0 },
      lintStatus: { errors: 0, warnings: 0, infos: 0 },
      problems: [],
      resultsStale: false,
    },
    {
      path: 'generated/cache/data.json',
      state: 'untracked',
      workKind: '?',
      adds: 2,
      dels: 0,
      testStatus: { passed: 0, failed: 0, skipped: 0 },
      lintStatus: { errors: 0, warnings: 0, infos: 0 },
      problems: [],
      resultsStale: false,
    },
  ],
  resultsStale: false,
  action: { running: false },
  commitQueue: { running: false },
};

const queuedCommit = (files: string[], status: string, extra: Record<string, unknown> = {}) => ({
  id: `task-${files.join('-')}`,
  action: 'commit',
  files,
  status,
  ...extra,
});

describe('ProjectStatusView', () => {
  it('opens a test-runner detail dialog from a task href', async () => {
    window.history.replaceState({}, '', '/projects/gavel?action=test&run=test-run-123');
    expect(window.location.search).toBe('?action=test&run=test-run-123');
    vi.stubGlobal('fetch', vi.fn(async () => (
      { ok: true, json: async () => statusResponse }
    ) as Response));

    render(<ProjectStatusView project={statusResponse.project} />);

    expect(await screen.findByLabelText('Test details · gavel')).toBeTruthy();
    expect(useTestRunMock).toHaveBeenCalledWith(expect.objectContaining({ runId: 'test-run-123' }));
  });

  it('selects status files and queues a file-scoped gavel commit', async () => {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return {
          ok: true,
          json: async () => ({ runId: 'commit-run-1', href: '/tasks/commit-run-1', running: true, entries: [queuedCommit(['one.go'], 'running')] }),
        } as Response;
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    render(<ProjectStatusView project={statusResponse.project} />);

    fireEvent.click(await screen.findByRole('checkbox', { name: 'Select one.go' }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit selected (1)' }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    expect(fetchMock.mock.calls[1]?.[0]).toBe('/api/projects/gavel/commit-queue');
    expect(JSON.parse(String(fetchMock.mock.calls[1]?.[1]?.body))).toEqual({
      action: 'commit',
      files: ['one.go'],
    });
    const queue = await screen.findByRole('region', { name: 'Commit queue' });
    expect(within(queue).getByRole('button', { name: 'one.go' })).toBeTruthy();
    expect(within(queue).getByText('running')).toBeTruthy();
  });

  it('keeps queueing enabled while a commit runs and locks its files out of the next group', async () => {
    const entries = [queuedCommit(['one.go'], 'running', { startedAt: '2026-07-20T10:00:00Z' })];
    const commitQueue = () => ({ runId: 'commit-run-1', href: '/tasks/commit-run-1', running: true, entries: [...entries] });
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        entries.push(queuedCommit(['generated/output.log'], 'pending'));
        return { ok: true, json: async () => commitQueue() } as Response;
      }
      return { ok: true, json: async () => ({ ...statusResponse, commitQueue: commitQueue() }) } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    render(<ProjectStatusView project={statusResponse.project} />);

    const lockedRow = await screen.findByRole('treeitem', { name: 'one.go' });
    expect(within(lockedRow).getByText('queued in group 1')).toBeTruthy();
    expect(within(lockedRow).getByRole('checkbox', { name: 'Select one.go' }).hasAttribute('disabled')).toBe(true);

    fireEvent.click(screen.getByRole('checkbox', { name: 'Select generated/output.log' }));
    const commit = screen.getByRole('button', { name: 'Commit selected (1)' });
    expect(commit.hasAttribute('disabled')).toBe(false);
    fireEvent.click(commit);

    await waitFor(() => expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(true));
    const queue = await screen.findByRole('region', { name: 'Commit queue' });
    expect(within(queue).getByText('2 groups, 1 waiting')).toBeTruthy();
  });

  it('cancels a queued commit group', async () => {
    const entries = [queuedCommit(['one.go'], 'running'), queuedCommit(['generated/output.log'], 'pending')];
    const commitQueue = () => ({ runId: 'commit-run-1', running: true, entries: [...entries] });
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'DELETE') {
        entries.splice(1, 1);
        return { ok: true, json: async () => commitQueue() } as Response;
      }
      return { ok: true, json: async () => ({ ...statusResponse, commitQueue: commitQueue() }) } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    render(<ProjectStatusView project={statusResponse.project} />);

    fireEvent.click(await screen.findByRole('button', { name: 'Cancel commit group 2' }));

    await waitFor(() => expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'DELETE')).toBe(true));
    const cancelCall = fetchMock.mock.calls.find(([, init]) => init?.method === 'DELETE');
    expect(cancelCall?.[0]).toBe('/api/projects/gavel/commit-queue/task-generated%2Foutput.log');
    await waitFor(() => expect(screen.queryByRole('button', { name: 'Cancel commit group 2' })).toBeNull());
  });

  it('keeps conflicted files out of the selectable lifecycle', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => (
      { ok: true, json: async () => statusResponse }
    ) as Response));

    render(<ProjectStatusView project={statusResponse.project} />);

    expect((await screen.findByRole('checkbox', { name: 'Select conflicted.go' })).hasAttribute('disabled')).toBe(true);
    expect(screen.getByRole('button', { name: 'Commit selected (0)' }).hasAttribute('disabled')).toBe(true);
  });

  it('renders changed paths as a file tree and ignores an untracked directory', async () => {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return { ok: true, json: async () => ({ path: 'generated', directory: true, rule: '/generated/', added: true }) } as Response;
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    render(<ProjectStatusView project={statusResponse.project} />);

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
  });

  it('keeps diff counts on the file heading and uses extension-specific file icons', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => (
      { ok: true, json: async () => statusResponse }
    ) as Response));

    render(<ProjectStatusView project={statusResponse.project} />);

    const goRow = await screen.findByRole('treeitem', { name: 'one.go' });
    const goHeading = goRow.querySelector('[data-file-primary]');
    expect(goHeading).not.toBeNull();
    expect(within(goHeading as HTMLElement).getByText('+12')).toBeTruthy();
    expect(within(goHeading as HTMLElement).getByText('−3')).toBeTruthy();
    expect(goRow.querySelector('[data-file-icon="go"]')).not.toBeNull();

    const jsonRow = screen.getByRole('treeitem', { name: 'generated/cache/data.json' });
    expect(jsonRow.querySelector('[data-file-icon="json"]')).not.toBeNull();
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
        return { ok: true, json: async () => ({ action: 'lint', runId: 'advanced-lint-run', running: false }) } as Response;
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    render(<ProjectStatusView project={statusResponse.project} />);

    fireEvent.click(await screen.findByRole('checkbox', { name: 'Select one.go' }));
    fireEvent.click(screen.getByRole('button', { name: 'Lint options: advanced' }));
    expect(await screen.findByLabelText('Advanced lint options')).toBeTruthy();
    expect(screen.getByTestId('advanced-options').textContent).toContain('"files":["one.go"]');
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
        return {
          ok: true,
          json: async () => ({ runId: 'commit-run-1', running: true, entries: [queuedCommit(['one.go'], 'running')] }),
        } as Response;
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    render(<ProjectStatusView project={statusResponse.project} />);

    fireEvent.click(await screen.findByRole('checkbox', { name: 'Select one.go' }));
    fireEvent.click(screen.getByRole('button', { name: 'Commit options: advanced' }));
    expect(await screen.findByLabelText('Advanced commit options')).toBeTruthy();
    expect(screen.getByTestId('advanced-options').textContent).toContain('"files":["one.go"]');
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
    expect(await screen.findByRole('region', { name: 'Commit queue' })).toBeTruthy();
  });

  it('shows a failed commit group with its output and exit status', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({
      ok: true,
      json: async () => ({
        ...statusResponse,
        commitQueue: {
          runId: 'commit-run-123',
          href: '/tasks/commit-run-123',
          running: false,
          entries: [queuedCommit(['one.go'], 'failed', {
            startedAt: '2026-07-20T10:00:00Z',
            endedAt: '2026-07-20T10:00:03Z',
            exitCode: 1,
            output: 'commit validation failed\n',
            error: 'exit status 1',
          })],
        },
      }),
    }) as Response));

    render(<ProjectStatusView project={statusResponse.project} />);

    const queue = await screen.findByRole('region', { name: 'Commit queue' });
    expect(within(queue).getByText('failed')).toBeTruthy();
    expect(within(queue).getByText('Exit 1')).toBeTruthy();
    expect(within(queue).getByText('exit status 1')).toBeTruthy();
    expect(within(queue).getByRole('link', { name: 'Details' }).getAttribute('href')).toBe('/tasks/commit-run-123');

    fireEvent.click(within(queue).getByRole('button', { name: 'one.go' }));
    expect(within(queue).getByText('commit validation failed')).toBeTruthy();
  });

  it.each([
    ['test', 'Test changed', 'Test details · gavel'],
    ['lint', 'Lint project', 'Lint details · gavel'],
  ] as const)('opens %s in the test-runner dialog backed by its SSE run id', async (action, button, title) => {
    const runId = `${action}-run-id`;
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        return { ok: true, json: async () => ({ action, runId, running: true }) } as Response;
      }
      return { ok: true, json: async () => statusResponse } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);

    render(<ProjectStatusView project={statusResponse.project} />);

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

    render(<ProjectStatusView project={statusResponse.project} onDiffPathChange={onDiffPathChange} />);

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

    render(<ProjectStatusView project={statusResponse.project} diffPath="one.go" />);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith('/api/projects/gavel/diff?path=one.go', expect.anything()));
    expect(await screen.findByRole('heading', { name: 'one.go' })).toBeTruthy();
  });
});
