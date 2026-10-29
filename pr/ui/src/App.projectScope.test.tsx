import { useEffect } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ChatWindowManagerProvider } from '@flanksource/clicky-ui/ai';
import type { AgentSession, Project } from './types';
import { App } from './App';

const { isMobileMock, saveConfigMock, setProjectHiddenMock, agentSessionsMock, inspectorProps } = vi.hoisted(() => ({
  isMobileMock: vi.fn(() => false),
  saveConfigMock: vi.fn(),
  setProjectHiddenMock: vi.fn(),
  agentSessionsMock: vi.fn(),
  inspectorProps: { current: null as null | Record<string, unknown> },
}));

const projects: Project[] = [
  { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] },
  { name: 'clicky', dir: '/work/clicky', repos: ['acme/clicky'] },
  { name: 'archive', dir: '/work/archive', repos: ['acme/archive'], hidden: true },
];

vi.mock('./useAppQueries', () => ({
  useAppQueries: () => ({
    snapshot: {
      prs: [],
      config: { repos: ['acme/gavel'], project: 'gavel' },
      paused: false,
    },
    projects,
    projectsLoaded: true,
    projectError: '',
    processError: '',
    updateSnapshot: vi.fn(),
    refreshProjects: vi.fn(),
    refreshProjectsAndProcesses: vi.fn(),
  }),
}));

vi.mock('./useAppMutations', () => ({
  useAppMutations: () => ({
    error: '',
    markSeen: vi.fn(),
    refresh: vi.fn(),
    togglePause: vi.fn(),
    saveConfig: saveConfigMock,
    setIncludeBots: vi.fn(),
    setShowClosed: vi.fn(),
    setProjectHidden: setProjectHiddenMock,
  }),
}));

vi.mock('./usePRDetailStream', () => ({
  usePRDetailStream: () => ({ detail: null, loading: false, refresh: vi.fn() }),
}));

vi.mock('./components/todos/useWorkspaceTodos', () => ({
  useWorkspaceTodos: (choices: Project[]) => ({
    workspaces: choices,
    byDir: {},
    filters: { workspaces: {} },
    setFilters: vi.fn(),
    layout: 'split',
    selected: null,
    showCreate: true,
    setShowCreate: vi.fn(),
  }),
}));

vi.mock('./components/todos/PlanReview', () => ({
  useReviewMode: () => ({ active: false }),
  TodoReviewButton: () => null,
  PlanReviewBar: () => null,
}));

vi.mock('./useProjectCatalog', () => ({
  useProjectCatalog: ({ configured, selectedName }: { configured: Project[]; selectedName: string }) => ({
    projects: configured, selected: configured.find(project => project.name === selectedName), runs: [], loading: false, error: '',
  }),
}));
vi.mock('./components/ProjectRefPane', () => ({
  ProjectRefPane: ({ project, projectRef, diffPath }: { project: Project; projectRef: string; diffPath: string }) => (
    <div data-testid="project-review" data-ref={projectRef} data-diff={diffPath}>{project.name} changes</div>
  ),
}));

vi.mock('./useCopyFeedback', () => ({
  useCopyFeedback: () => ({
    copyState: 'idle',
    copyError: '',
    beginCopy: vi.fn(),
    resetCopyFeedback: vi.fn(),
  }),
}));

vi.mock('./components/sessions/sessionQueries', () => ({
  useAgentSessions: (options: { enabled: boolean }) => agentSessionsMock(options),
}));
vi.mock('@flanksource/clicky-ui/ai', async importOriginal => ({
  ...await importOriginal<typeof import('@flanksource/clicky-ui/ai')>(),
  SessionInspector: (props: Record<string, unknown>) => {
    inspectorProps.current = props;
    const onLoadStateChange = props.onLoadStateChange as ((state: { status: 'ready' }) => void) | undefined;
    // A session that loads fine: the inspector reports ready, then the screen offers its composer.
    useEffect(() => onLoadStateChange?.({ status: 'ready' }), [onLoadStateChange]);
    return <div data-testid="inspector">{props.composer as never}</div>;
  },
}));
vi.mock('./components/MenubarTodos', () => ({ MenubarTodos: () => <div data-testid="menubar-todos" /> }));
vi.mock('./useDocumentVisible', () => ({ useDocumentVisible: () => true }));
vi.mock('./useIsMobile', () => ({ useIsMobile: () => isMobileMock() }));
vi.mock('./components/StatusIndicator', () => ({ StatusIndicator: () => null }));
vi.mock('@flanksource/clicky-ui/components', async importOriginal => ({
  ...await importOriginal<typeof import('@flanksource/clicky-ui/components')>(),
  AppShell: () => null,
}));
vi.mock('./storage', () => ({
  loadUIState: () => ({}),
  saveUIState: vi.fn(),
  filtersFromStored: () => undefined,
}));

vi.mock('./components/todos/TodoNewPage', () => ({
  TodoNewPage: ({ projects: choices }: { projects: Project[] }) => (
    <div data-testid="todo-new-projects">
      {choices.map(project => <span key={project.name}>{project.name}</span>)}
    </div>
  ),
}));

vi.mock('./components/todos/CreateTodoDialog', () => ({
  CreateTodoDialog: ({ open, workspaces, defaultDir }: { open: boolean; workspaces: Project[]; defaultDir?: string }) => open ? (
    <div data-testid="create-dialog-projects" data-default-dir={defaultDir}>
      {workspaces.map(project => <span key={project.name}>{project.name}</span>)}
    </div>
  ) : null,
}));

vi.mock('./components/AddProjectDialog', () => ({ AddProjectDialog: () => null }));
vi.mock('./components/CommandPalette', () => ({
  CommandPalette: () => null,
  SearchTrigger: () => null,
}));

const agentSession: AgentSession = {
  id: 'sess-1',
  title: 'Refactor parser',
  source: 'claude',
  project: 'gavel',
  lifecycleStatus: 'running',
  activityState: 'thinking',
  processActive: true,
  lastActivityAt: new Date().toISOString(),
  tokens: { input: 0, output: 0, cacheRead: 0, total: 0 },
  costUsd: 0,
};

beforeEach(() => {
  isMobileMock.mockReturnValue(false);
  saveConfigMock.mockReset();
  setProjectHiddenMock.mockReset();
  agentSessionsMock.mockReset();
  agentSessionsMock.mockReturnValue({ data: { sessions: [agentSession], errors: [] }, error: null });
  inspectorProps.current = null;
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => [] }));
  window.history.replaceState({}, '', '/todos/new?project=gavel');
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

function renderApp() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <ChatWindowManagerProvider storageId="gavel-chat-test">
        <App />
      </ChatWindowManagerProvider>
    </QueryClientProvider>,
  );
}

describe('todo creation project choices', () => {
  it('keeps every project available when the dashboard route is scoped', () => {
    renderApp();

    expect(screen.getByTestId('todo-new-projects').textContent).toBe('gavelclicky');
  });

  it('keeps every project available in the dashboard new-todo dialog', () => {
    window.history.replaceState({}, '', '/todos?project=gavel');

    renderApp();

    const dialog = screen.getByTestId('create-dialog-projects');
    expect(dialog.textContent).toBe('gavelclicky');
    expect(dialog.dataset.defaultDir).toBe('/work/gavel');
  });
});

describe('hidden projects', () => {
  it('offers a hidden project only under Manage hidden and unhides it through the mutation', () => {
    isMobileMock.mockReturnValue(true);
    window.history.replaceState({}, '', '/todos?project=gavel');
    renderApp();

    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    expect(screen.queryByRole('button', { name: 'Filter by archive project' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: /Manage hidden \(1\)/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Unhide archive project' }));

    expect(setProjectHiddenMock).toHaveBeenCalledExactlyOnceWith({ name: 'archive', hidden: false });
  });

  it('falls back to all visible projects when the scoped project is hidden', () => {
    isMobileMock.mockReturnValue(true);
    window.history.replaceState({}, '', '/todos?project=archive');
    renderApp();

    expect(saveConfigMock).toHaveBeenCalledWith({
      repos: ['acme/gavel', 'acme/clicky'],
      project: '',
      org: '',
      all: false,
    });
    expect(`${window.location.pathname}${window.location.search}`).toBe('/todos');
  });
});

describe('compact project scope', () => {
  it('keeps the lifecycle editor route visible on a compact screen', () => {
    isMobileMock.mockReturnValue(true);
    window.history.replaceState({}, '', '/prompts/lifecycles/todos?project=gavel');
    renderApp();
    expect(screen.getByRole('link', { name: 'Back to dashboard' })).toBeTruthy();
    expect(screen.getByRole('tab', { name: 'Lifecycles' })).toBeTruthy();
    expect(window.location.pathname).toBe('/prompts/lifecycles/todos');
  });

  it('shows the active project picker in the mobile layout', () => {
    isMobileMock.mockReturnValue(true);
    window.history.replaceState({}, '', '/prs?project=gavel');

    renderApp();

    expect(screen.getByTitle('Switch project / GitHub scope').textContent).toContain('gavel');
  });

  it('updates the mobile route when a project is selected', () => {
    isMobileMock.mockReturnValue(true);
    window.history.replaceState({}, '', '/todos?project=gavel');
    renderApp();

    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    fireEvent.click(screen.getByRole('button', { name: 'Filter by clicky project' }));

    expect(`${window.location.pathname}${window.location.search}`).toBe('/todos?project=clicky');
    expect(saveConfigMock).toHaveBeenCalledWith({
      repos: ['acme/clicky'],
      project: 'clicky',
      org: '',
      all: false,
    });
  });

  it('keeps the native webview on the menubar route when a project is selected', () => {
    window.history.replaceState({}, '', '/menubar');
    renderApp();

    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    fireEvent.click(screen.getByRole('button', { name: 'Filter by clicky project' }));

    expect(`${window.location.pathname}${window.location.search}`).toBe('/menubar');
    expect(screen.getByTitle('Switch project / GitHub scope').textContent).toContain('clicky');
  });
});

describe('mobile sessions', () => {
  beforeEach(() => {
    isMobileMock.mockReturnValue(true);
  });

  const url = () => `${window.location.pathname}${window.location.search}`;

  it('puts the Sessions tab in the URL and lists sessions, polling only while it is open', () => {
    window.history.replaceState({}, '', '/prs');
    renderApp();
    expect(agentSessionsMock).toHaveBeenLastCalledWith({ enabled: false });

    fireEvent.click(screen.getByRole('button', { name: /^Sessions/ }));

    expect(url()).toBe('/sessions');
    expect(agentSessionsMock).toHaveBeenLastCalledWith({ enabled: true });
    expect(screen.getByText('Refactor parser')).toBeTruthy();
  });

  it('opens a session full-screen at /sessions/{id} and goes back to the list', async () => {
    window.history.replaceState({}, '', '/sessions');
    renderApp();

    fireEvent.click(screen.getByText('Refactor parser'));

    expect(url()).toBe('/sessions/sess-1');
    expect(screen.getByTestId('inspector')).toBeTruthy();
    expect(inspectorProps.current).toMatchObject({ src: '/api/captain/sessions/sess-1', follow: true, layout: 'compact' });
    expect(await screen.findByPlaceholderText('Add a follow-up…')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Back to sessions' }));

    expect(url()).toBe('/sessions');
    expect(screen.queryByTestId('inspector')).toBeNull();
  });

  it('restores the open session from a deep link', () => {
    window.history.replaceState({}, '', '/sessions/sess-1?project=gavel');
    renderApp();

    expect(inspectorProps.current).toMatchObject({ src: '/api/captain/sessions/sess-1' });
    expect(screen.getAllByText('Refactor parser').length).toBeGreaterThan(0);
  });

  it('keeps the mobile tab in the URL when switching back to PRs', () => {
    window.history.replaceState({}, '', '/sessions');
    renderApp();

    fireEvent.click(screen.getByRole('button', { name: /^PRs/ }));

    expect(window.location.pathname).toBe('/prs');
  });

  it('shows the Todos tab for /todos', () => {
    window.history.replaceState({}, '', '/todos');
    renderApp();

    expect(screen.getByTestId('menubar-todos')).toBeTruthy();
  });

  it('offers the Gavel assistant launcher', () => {
    window.history.replaceState({}, '', '/sessions');
    renderApp();

    expect(screen.getByRole('button', { name: 'Open Gavel assistant' })).toBeTruthy();
  });

  it('keeps the native menubar on /menubar with local tab state and no assistant launcher', () => {
    window.history.replaceState({}, '', '/menubar');
    isMobileMock.mockReturnValue(false);
    renderApp();
    expect(screen.queryByRole('button', { name: 'Open Gavel assistant' })).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: /^Sessions/ }));

    expect(url()).toBe('/menubar');
    expect(agentSessionsMock).toHaveBeenLastCalledWith({ enabled: true });
    fireEvent.click(screen.getByText('Refactor parser'));
    expect(url()).toBe('/menubar');
    expect(inspectorProps.current).toMatchObject({ src: '/api/captain/sessions/sess-1' });
    fireEvent.click(screen.getByRole('button', { name: 'Back to sessions' }));
    expect(screen.getByText('Refactor parser')).toBeTruthy();
  });
});

describe('mobile projects', () => {
  beforeEach(() => isMobileMock.mockReturnValue(true));

  it('opens the project list, reviews a project, and returns to the list', () => {
    window.history.replaceState({}, '', '/prs');
    renderApp();
    fireEvent.click(screen.getByRole('button', { name: 'Projects' }));
    expect(window.location.pathname).toBe('/projects');
    fireEvent.click(screen.getByRole('button', { name: 'Open gavel project' }));
    expect(window.location.pathname).toBe('/projects/gavel');
    expect(screen.getByTestId('project-review').textContent).toBe('gavel changes');
    fireEvent.click(screen.getByRole('button', { name: 'Back to projects' }));
    expect(window.location.pathname).toBe('/projects');
    expect(screen.getByRole('button', { name: 'Open gavel project' })).toBeTruthy();
  });

  it('restores a file diff and worktree from the existing project URL', () => {
    window.history.replaceState({}, '', '/projects/gavel?diff=src%2Fmain.go&ref=wt%3A%2Fwork%2Ffeature');
    renderApp();
    expect(screen.getByTestId('project-review').dataset).toMatchObject({ ref: 'wt:/work/feature', diff: 'src/main.go' });
    expect(screen.queryByRole('button', { name: 'Projects' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Back to projects' })).toBeNull();
  });

  it('keeps Processes and Projects on separate mobile routes', () => {
    window.history.replaceState({}, '', '/projects');
    renderApp();
    fireEvent.click(screen.getByRole('button', { name: /^Processes/ }));
    expect(window.location.pathname).toBe('/processes');
    fireEvent.click(screen.getByRole('button', { name: 'Projects' }));
    expect(window.location.pathname).toBe('/projects');
  });

  it('keeps a way back to Projects for an unknown project diff link', () => {
    window.history.replaceState({}, '', '/projects/missing?diff=src%2Fmain.go');
    renderApp();
    expect(screen.getByText('Project missing was not found')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Back to projects' }));
    expect(window.location.pathname).toBe('/projects');
  });
});
