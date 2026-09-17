import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Project } from './types';
import { App } from './App';

const { isMobileMock, saveConfigMock } = vi.hoisted(() => ({
  isMobileMock: vi.fn(() => false),
  saveConfigMock: vi.fn(),
}));

const projects: Project[] = [
  { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] },
  { name: 'clicky', dir: '/work/clicky', repos: ['acme/clicky'] },
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
    procStatus: {},
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
  useProjectCatalog: () => ({ projects: [] }),
}));

vi.mock('./useCopyFeedback', () => ({
  useCopyFeedback: () => ({
    copyState: 'idle',
    copyError: '',
    beginCopy: vi.fn(),
    resetCopyFeedback: vi.fn(),
  }),
}));

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

beforeEach(() => {
  isMobileMock.mockReturnValue(false);
  saveConfigMock.mockReset();
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
      <App />
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

describe('compact project scope', () => {
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
