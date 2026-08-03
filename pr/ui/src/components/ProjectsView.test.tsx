import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen, waitFor } from '@testing-library/react';
import type { PropsWithChildren, ReactElement } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ProjectDetailPane, ProjectsSidebar } from './ProjectsView';
import type { Project } from '../types';
import { useProjectCatalog } from '../useProjectCatalog';

const streamListeners = new Map<string, EventListener>();

vi.mock('./ProcControl', () => ({ ProcControl: () => null }));
vi.mock('./TodoBadge', () => ({ TodoBadge: () => null }));
vi.mock('./GitChangesBadge', () => ({ GitChangesBadge: () => null }));
vi.mock('./RelativeTime', () => ({ RelativeTime: ({ iso }: { iso: string }) => <span>{iso}</span> }));
vi.mock('./ProjectStatusView', () => ({ ProjectStatusView: ({ project, showResults }: { project: Project; showResults: boolean }) => <div>Status {project.name}, results {showResults ? 'on' : 'off'}</div> }));
vi.mock('./tests/TestRunDetail', () => ({ TestRunDetail: ({ project, runId }: { project: string; runId: string }) => <div>Run {project}/{runId}</div> }));

const project: Project = { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] };
const response = {
  projects: [{
    name: 'gavel',
    dir: '/work/gavel',
    runs: [{
      runId: 'run-1',
      kind: 'lint',
      started: '2026-07-21T10:59:33Z',
      passed: 0,
      failed: 0,
      skipped: 0,
      warned: 0,
      total: 0,
      lintViolations: 3,
      lintLinters: 2,
    }],
  }],
};

// Stands in for App's AppShell, which renders the sidebar into `bodySidebar` and
// the detail pane into `children` off the one catalog the hook returns.
function ProjectsTab({ configured, selectedName, selectedRunId, enabled = true, historyEnabled = true, resultsEnabled = false }: {
  configured: Project[];
  selectedName: string;
  selectedRunId: string;
  enabled?: boolean;
  historyEnabled?: boolean;
  resultsEnabled?: boolean;
}) {
  const catalog = useProjectCatalog({ configured, selectedName, enabled: enabled && historyEnabled });
  return (
    <div>
      <ProjectsSidebar
        catalog={catalog}
        procStatus={{}}
        selectedName={selectedName}
        selectedRunId={selectedRunId}
        historyEnabled={historyEnabled}
        onSelect={() => {}}
        onSelectRun={() => {}}
        onHistoryChange={() => {}}
        onChanged={() => {}}
        onAdd={() => {}}
        onSettings={() => {}}
      />
      <ProjectDetailPane
        catalog={catalog}
        selectedName={selectedName}
        selectedRunId={selectedRunId}
        diffPath=""
        resultsEnabled={resultsEnabled}
        onDiffPathChange={() => {}}
        onChanged={() => {}}
      />
    </div>
  );
}

function CatalogProbe({ name, enabled = true }: { name: string; enabled?: boolean }) {
  const catalog = useProjectCatalog({ configured: [project], selectedName: 'gavel', enabled });
  return <div data-testid={name}>{catalog.runs.length}</div>;
}

function createClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false } } });
}

function Provider({ client, children }: PropsWithChildren<{ client: QueryClient }>) {
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

function renderWithClient(ui: ReactElement, client = createClient()) {
  return { client, ...render(<Provider client={client}>{ui}</Provider>) };
}

function stubRunHistory(payload: unknown) {
  vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => payload }) as Response));
  vi.stubGlobal('EventSource', vi.fn(() => ({
    addEventListener: (type: string, listener: EventListener) => streamListeners.set(type, listener),
    close: vi.fn(),
    onerror: null,
  })));
}

afterEach(() => {
  streamListeners.clear();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('projects tab', () => {
  it('lists history-only projects when the configured project catalog is empty', async () => {
    stubRunHistory(response);

    renderWithClient(<ProjectsTab configured={[]} selectedName="gavel" selectedRunId="run-1" />);

    expect(await screen.findByRole('button', { name: 'Open lint run run-1 for gavel' })).toBeTruthy();
    expect(screen.getByText('Run gavel/run-1')).toBeTruthy();
  });

  it('loads project run history and renders the selected run in the detail pane', async () => {
    stubRunHistory(response);

    renderWithClient(<ProjectsTab configured={[project]} selectedName="gavel" selectedRunId="run-1" />);

    expect(await screen.findByRole('button', { name: 'Open lint run run-1 for gavel' })).toBeTruthy();
    expect(screen.getByText('Run gavel/run-1')).toBeTruthy();
  });

  it('applies valid SSE updates and surfaces malformed frames', async () => {
    stubRunHistory({ projects: [{ name: 'gavel', dir: '/work/gavel', runs: [] }] });

    renderWithClient(<ProjectsTab configured={[project]} selectedName="gavel" selectedRunId="" />);

    await screen.findByText('No test or lint runs');
    act(() => streamListeners.get('message')?.(new MessageEvent('message', { data: JSON.stringify(response) })));
    expect(await screen.findByRole('button', { name: 'Open lint run run-1 for gavel' })).toBeTruthy();

    act(() => streamListeners.get('message')?.(new MessageEvent('message', { data: '{' })));
    expect((await screen.findByRole('alert')).textContent).toContain('invalid update');
  });

  it('does not open the run-history stream while another tab is active', () => {
    stubRunHistory(response);

    renderWithClient(<ProjectsTab configured={[project]} selectedName="gavel" selectedRunId="run-1" enabled={false} />);

    expect(fetch).not.toHaveBeenCalled();
    expect(EventSource).not.toHaveBeenCalled();
  });

  it('does not load run history before opt-in and hides cached history after opt-out', async () => {
    stubRunHistory(response);
    const client = createClient();
    const view = renderWithClient(
      <ProjectsTab configured={[project]} selectedName="gavel" selectedRunId="" historyEnabled={false} />,
      client,
    );

    expect(fetch).not.toHaveBeenCalled();
    expect(EventSource).not.toHaveBeenCalled();
    expect(screen.queryByRole('button', { name: 'Open lint run run-1 for gavel' })).toBeNull();

    view.rerender(
      <Provider client={client}>
        <ProjectsTab configured={[project]} selectedName="gavel" selectedRunId="" historyEnabled />
      </Provider>,
    );
    expect(await screen.findByRole('button', { name: 'Open lint run run-1 for gavel' })).toBeTruthy();

    view.rerender(
      <Provider client={client}>
        <ProjectsTab configured={[project]} selectedName="gavel" selectedRunId="" historyEnabled={false} />
      </Provider>,
    );
    expect(screen.queryByRole('button', { name: 'Open lint run run-1 for gavel' })).toBeNull();
  });

  it('forwards the hidden result option independently of run history', () => {
    stubRunHistory(response);

    renderWithClient(<ProjectsTab configured={[project]} selectedName="gavel" selectedRunId="" historyEnabled={false} resultsEnabled />);

    expect(screen.getByText('Status gavel, results on')).toBeTruthy();
    expect(fetch).not.toHaveBeenCalled();
  });

  it('deduplicates the run-history bootstrap request across consumers', async () => {
    stubRunHistory(response);

    renderWithClient(
      <>
        <CatalogProbe name="first" />
        <CatalogProbe name="second" />
      </>,
    );

    await waitFor(() => expect(screen.getByTestId('first').textContent).toBe('1'));
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it('aborts the run-history request when the tab is disabled', async () => {
    let signal: AbortSignal | undefined;
    vi.stubGlobal('fetch', vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
      signal = init?.signal ?? undefined;
      return new Promise<Response>(() => {});
    }));
    vi.stubGlobal('EventSource', vi.fn(() => ({ addEventListener: vi.fn(), close: vi.fn(), onerror: null })));
    const client = createClient();
    const view = renderWithClient(<CatalogProbe name="probe" />, client);
    await waitFor(() => expect(signal).toBeDefined());

    view.rerender(
      <Provider client={client}>
        <CatalogProbe name="probe" enabled={false} />
      </Provider>,
    );

    await waitFor(() => expect(signal?.aborted).toBe(true));
  });
});
