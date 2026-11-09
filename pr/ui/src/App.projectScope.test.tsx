import { useEffect, type ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ChatWindowManagerProvider } from '@flanksource/clicky-ui/ai';
import type { AgentSession, Project } from './types';
import { App } from './App';

const { isMobileMock, refreshMock, saveConfigMock, setProjectHiddenMock, agentSessionsMock, inspectorProps } = vi.hoisted(() => ({
  isMobileMock: vi.fn(() => false),
  refreshMock: vi.fn(),
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
    config: { repos: ['acme/gavel'], project: 'gavel' },
    prsLoading: false,
    prError: '',
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
    refresh: refreshMock,
    refreshing: false,
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
  // Keep the shell slots under test while leaving layout to the browser checks.
  AppShell: ({ navSections, mobileActions, toolbar, bodySidebar, bodyPane, onBodyBack, bodyBackLabel }: {
    navSections?: { items: { key: string; label: string; to: string; active?: boolean }[] }[];
    mobileActions?: ReactNode;
    toolbar?: ReactNode;
    bodySidebar?: ReactNode;
    bodyPane?: 'sidebar' | 'main';
    onBodyBack?: () => void;
    bodyBackLabel?: string;
  }) => navSections ? (
    <div data-testid="shell" data-body-pane={bodyPane}>
      {bodyPane === 'main' && <button type="button" onClick={onBodyBack}>{bodyBackLabel}</button>}
      <nav>
        {navSections.flatMap(section => section.items).map(item => (
          <a key={item.key} href={item.to} aria-current={item.active ? 'page' : undefined}>{item.label}</a>
        ))}
      </nav>
      {mobileActions}
      {toolbar}
      {bodyPane === 'sidebar' && navSections.some(section => section.items.some(item => item.key === 'prs' && item.active)) && bodySidebar}
    </div>
  ) : null,
}));
vi.mock('./components/UserMenu', () => ({ UserMenu: () => null }));
vi.mock('./components/ThemeToggle', () => ({ ThemeToggle: () => null }));
vi.mock('./components/ProcessManager', () => ({ ProcessManager: () => null }));
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
  refreshMock.mockReset();
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
    window.history.replaceState({}, '', '/menubar');
    renderApp();

    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    expect(screen.queryByRole('button', { name: 'Filter by archive project' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: /Manage hidden \(1\)/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Unhide archive project' }));

    expect(setProjectHiddenMock).toHaveBeenCalledExactlyOnceWith({ name: 'archive', hidden: false });
  });

  it('falls back to all visible projects when the scoped project is hidden', () => {
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
  it('keeps the project picker under More actions in the mobile layout', () => {
    isMobileMock.mockReturnValue(true);
    window.history.replaceState({}, '', '/sessions?project=gavel');

    renderApp();
    expect(screen.queryByTitle('Switch project / GitHub scope')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'More actions' }));

    expect(screen.getByRole('button', { name: 'More actions' }).getAttribute('aria-expanded')).toBe('true');
    expect(screen.getByTitle('Switch project / GitHub scope').textContent).toContain('gavel');
  });

  it('updates the mobile route when a project is selected', () => {
    isMobileMock.mockReturnValue(true);
    window.history.replaceState({}, '', '/sessions?project=gavel');
    renderApp();

    fireEvent.click(screen.getByRole('button', { name: 'More actions' }));
    fireEvent.click(screen.getByTitle('Switch project / GitHub scope'));
    fireEvent.click(screen.getByRole('button', { name: 'Filter by clicky project' }));

    expect(`${window.location.pathname}${window.location.search}`).toBe('/sessions?project=clicky');
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

describe('mobile shell', () => {
  beforeEach(() => {
    isMobileMock.mockReturnValue(true);
  });

  const url = () => `${window.location.pathname}${window.location.search}`;

  it('refreshes the mobile PR list through the existing mutation without navigating', () => {
    window.history.replaceState({}, '', '/prs?project=gavel');
    renderApp();
    const list = screen.getByRole('region', { name: 'Pull requests' });
    const touch = { identifier: 1, clientX: 100, clientY: 100 };
    fireEvent.touchStart(list, { touches: [touch] });
    fireEvent.touchMove(list, { touches: [{ ...touch, clientY: 220 }] });
    fireEvent.touchEnd(list, { touches: [] });

    expect(refreshMock).toHaveBeenCalledTimes(1);
    expect(window.location.pathname).toBe('/prs');
    expect(new URLSearchParams(window.location.search).get('project')).toBe('gavel');
  });

  it.each(['/sessions', '/prs/acme/gavel/7'])('leaves %s outside the pull-to-refresh region', path => {
    window.history.replaceState({}, '', path);
    renderApp();
    expect(screen.queryByRole('region', { name: 'Pull requests' })).toBeNull();
    expect(refreshMock).not.toHaveBeenCalled();
  });

  it('moves the dashboard tabs into the navigation drawer as scoped route links', () => {
    window.history.replaceState({}, '', '/sessions?project=gavel');
    renderApp();

    const links = screen.getAllByRole('link').map(link => [link.textContent, link.getAttribute('href'), link.getAttribute('aria-current')]);
    expect(links).toEqual([
      ['PRs', '/prs?project=gavel&authors=-%40bots', null],
      ['Projects', '/projects?project=gavel', null],
      ['Todos', '/todos?project=gavel', null],
      ['Sessions', '/sessions?project=gavel', 'page'],
      ['Tasks', '/tasks?project=gavel', null],
      ['Prompts', '/prompts?project=gavel', null],
      ['Activity', '/activity?project=gavel', null],
    ]);
  });

  it.each([
    ['/sessions/sess-1', 'Back to sessions', '/sessions'],
    ['/projects/gavel', 'Back to projects', '/projects'],
    ['/todos/abc', 'Back to todos', '/todos'],
  ])('pages %s to the detail pane and returns to the list with %s', (deepLink, back, list) => {
    window.history.replaceState({}, '', deepLink);
    renderApp();
    expect(screen.getByTestId('shell').dataset.bodyPane).toBe('main');

    fireEvent.click(screen.getByRole('button', { name: back }));

    expect(window.location.pathname).toBe(list);
    expect(screen.getByTestId('shell').dataset.bodyPane).toBe('sidebar');
  });

  it('polls sessions only while the Sessions route is open', () => {
    window.history.replaceState({}, '', '/prs');
    const { unmount } = renderApp();
    expect(agentSessionsMock).toHaveBeenLastCalledWith({ enabled: false });
    unmount();

    window.history.replaceState({}, '', '/sessions');
    renderApp();
    expect(agentSessionsMock).toHaveBeenLastCalledWith({ enabled: true });
  });

  it('keeps search and the Gavel assistant in the header and folds the rest under More actions', () => {
    window.history.replaceState({}, '', '/sessions');
    renderApp();

    expect(screen.getByRole('button', { name: 'Search pull requests, todos, and sessions' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Open Gavel assistant' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Global settings' })).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: 'More actions' }));

    expect(screen.getByRole('button', { name: 'Global settings' })).toBeTruthy();
    expect(screen.getAllByRole('button', { name: 'Open Gavel assistant' })).toHaveLength(1);
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

