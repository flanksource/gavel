import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import type { PRItem, PRDetail, PRInfo, SearchConfig, Project } from './types';
import { PRList } from './components/PRList';
import { PRDetailPanel } from './components/PRDetail';
import { FilterBar, emptyFilters, type Filters } from './components/FilterBar';
import { AppShell, Button } from '@flanksource/clicky-ui/components';
import { TaskManager, TaskManagerButton } from '@flanksource/clicky-ui/data';
import { ChatButton } from '@flanksource/clicky-ui/ai';
import { ActivityView } from './components/ActivityView';
import { TodoNewButton, TodoNavbarDensityPicker, TodoNavbarLayoutPicker, TodoWorkspaceList, TodoDetailPane, TodoFullPane } from './components/TodoView';
import { useWorkspaceTodos } from './components/todos/useWorkspaceTodos';
import { PlanReviewBar, TodoReviewButton, useReviewMode } from './components/todos/PlanReview';
import { CreateTodoDialog } from './components/todos/CreateTodoDialog';
import { TodoNewPage } from './components/todos/TodoNewPage';
import { StatusIndicator } from './components/StatusIndicator';
import { OrgChooser } from './components/OrgChooser';
import { AddProjectDialog } from './components/AddProjectDialog';
import { SettingsPage, type SettingsScope } from './components/settings/SettingsPage';
import { ProjectDetailPane, ProjectsSidebar } from './components/ProjectsView';
import { HostedProjectPane, HostedProjectsSidebar } from './components/HostedProjectsView';
import { ProcessManager } from './components/ProcessManager';
import { RunsView } from './components/RunsView';
import { ThemeToggle } from './components/ThemeToggle';
import { UserMenu } from './components/UserMenu';
import { githubScopeVisible, hostControlsVisible, tabFor, visibleTabs } from './hostedUI';
import { currentAuthMode } from './query';
import { ReactGrabHelp } from './components/ReactGrabHelp';
import { CommandPalette, SearchTrigger } from './components/CommandPalette';
import { flattenTodos, type TodoEntry } from './components/todos/todoGroup';
import { aggregateGavelShards, computeCounts, collectRepos, collectAuthors, filterPRs, prKey } from './utils';
import { useCopyFeedback } from './useCopyFeedback';
import {
  annotateRoutePaths,
  buildRoute,
  findPRByRoutePath,
  parseRoute,
  type RouteState,
  type Tab,
  type TodoDetailView,
} from './routes';
import { copyCurrentViewForAgent, downloadCurrentView } from './export';
import { loadUIState, saveUIState, filtersFromStored } from './storage';
import { useDocumentVisible } from './useDocumentVisible';
import { useProjectGitSummary } from './projectGitQueries';
import { useProjectCatalog } from './useProjectCatalog';
import { useAppQueries } from './useAppQueries';
import { useAppMutations } from './useAppMutations';
import { usePRDetailStream } from './usePRDetailStream';
import { useIsMobile } from './useIsMobile';
import { UiActivity, UiCheck, UiCog, UiCopy, UiFolderGit, UiGitPr, UiJson, UiListChecks, UiMarkdown, UiPlay, UiRobotAi, UiTerminal } from '@flanksource/clicky-ui/icons';
import { SessionsDetailPane, SessionsSidebar } from './components/sessions/SessionsView';
import { MenubarView } from './components/mobile/MenubarView';
import { ProcessesPage } from './components/ProcessesPage';
import { mobileTabFromRoute, routeTabFromMobile, type MobileTab } from './mobileTabs';
import { useAgentSessions } from './components/sessions/sessionQueries';
import { worktreeRef } from './projectRef';
import { PromptsView } from './components/prompts/PromptsView';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import type { ComponentType } from 'react';
import { Spinner } from './icons/Spinner';
import { applyProjectScope } from './projectScope';

const defaultConfig: SearchConfig = { repos: [] };

// Percentage width of the AppShell body sidebar per tab. A PR row carries title,
// repo, checks and badges so it wants half the body; a project is just a name and
// a run list, so its list is narrow.
const bodySplitByTab: Partial<Record<Tab, number>> = { prs: 50, projects: 22, todos: 38, sessions: 32 };

function prFromRoutePath(path: string): PRItem | null {
  const parts = path.split('/').filter(Boolean);
  if (parts.length !== 3) return null;
  const number = Number(parts[2]);
  if (!Number.isInteger(number) || number <= 0) return null;
  const repo = `${parts[0]}/${parts[1]}`;
  return {
    number,
    repo,
    title: `${repo}#${number}`,
    author: 'unknown',
    source: '...',
    target: '...',
    state: 'OPEN',
    isDraft: false,
    url: `https://github.com/${repo}/pull/${number}`,
    updatedAt: new Date().toISOString(),
    route_path: path,
  };
}

function mergePRItemFromDetail(pr: PRItem, info: PRInfo): PRItem {
  const author = info.author?.login || pr.author;
  return {
    ...pr,
    title: info.title || pr.title,
    author,
    authorAvatarUrl: info.author?.avatarUrl || pr.authorAvatarUrl,
    source: info.headRefName || pr.source,
    target: info.baseRefName || pr.target,
    state: info.state || pr.state,
    isDraft: info.isDraft,
    reviewDecision: info.reviewDecision || pr.reviewDecision,
    mergeable: info.mergeable || pr.mergeable,
    url: info.url || pr.url,
  };
}

export function App() {
  const initialRoute: RouteState = typeof window !== 'undefined'
    ? parseRoute(window.location)
    : { tab: 'prs', selectedPath: '', scopeProject: '', projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, projectRef: '', todoView: {}, filters: emptyFilters() };

  // Hydrate org/search config and filters from localStorage. URL query params
  // (if present) win for filters so deep links still work.
  const stored = typeof window !== 'undefined' ? loadUIState() : {};
  const hasUrlFilters = typeof window !== 'undefined'
    && ['state', 'checks', 'repos', 'authors'].some(name => new URLSearchParams(window.location.search).has(name));
  // First-run default hides bots (the daemon now fetches them so the @bots
  // author chip can toggle them back on). URL params and stored filters win.
  const defaultFilters: Filters = { ...emptyFilters(), authors: { '@bots': 'exclude' } };
  const initialFilters = hasUrlFilters ? initialRoute.filters : (filtersFromStored(stored.filters) ?? defaultFilters);
  const initialConfig: SearchConfig = { ...defaultConfig, ...stored.config };

  const [selected, setSelected] = useState<PRItem | null>(null);
  const [filters, setFilters] = useState<Filters>(initialFilters);
  const [selectedPath, setSelectedPath] = useState(initialRoute.selectedPath);
  const [scopeProject, setScopeProject] = useState(initialRoute.scopeProject || initialConfig.project || '');
  const [projectDiffPath, setProjectDiffPath] = useState(initialRoute.projectDiffPath);
  const [projectRunId, setProjectRunId] = useState(initialRoute.projectRunId);
  const [projectHistory, setProjectHistory] = useState(initialRoute.projectHistory);
  const [projectResults, setProjectResults] = useState(initialRoute.projectResults);
  const [projectRef, setProjectRef] = useState(initialRoute.projectRef);
  const [todoView, setTodoView] = useState(initialRoute.todoView);
  const authMode = currentAuthMode();
  const hostControls = hostControlsVisible(authMode);
  const [activeTab, setActiveTab] = useState<Tab>(() => tabFor(authMode, initialRoute.tab));
  const [menubarTab, setMenubarTab] = useState<MobileTab>('prs');
  const [menubarSessionId, setMenubarSessionId] = useState('');
  const [paletteOpen, setPaletteOpen] = useState(false);
  const { copyState, copyError, beginCopy, resetCopyFeedback } = useCopyFeedback({ copiedMs: 2500, errorMs: 2500 });
  const [addOpen, setAddOpen] = useState(false);
  const [settingsScope, setSettingsScope] = useState<SettingsScope | null>(null);
  const syncedScopeProject = useRef(initialConfig.project ?? '');
  const visible = useDocumentVisible();
  const isMobile = useIsMobile();

  const pathname = typeof window !== 'undefined' ? window.location.pathname : '';
  const isMenubar = pathname === '/menubar';
  // The processes and new-todo pages act on the host, so a hosted org gets the dashboard.
  const isProcessesPage = hostControls && pathname === '/processes';
  // The focused /todos/new form is a standalone page (like /menubar, /processes)
  // rather than a view inside the Todos tab, so it can be linked to and returns
  // to its referer. parseRoute reads it as the todos tab with selectedPath="new";
  // the early return below renders the form before any of that matters.
  const isTodoNewPage = hostControls && pathname === '/todos/new';
  // On mobile the desktop AppShell (split panes, sidebar, toolbars) has no room,
  // so the dashboard falls back to the same compact, single-column menubar
  // dropdown layout the native menubar webview uses. The standalone /processes
  // and /todos/new pages already lay out fine on narrow screens, so they keep
  // their own views.
  const isLifecyclePage = activeTab === 'prompts' && (selectedPath === 'lifecycles' || selectedPath.startsWith('lifecycles/'));
  const useMenubarLayout = isMenubar || (isMobile && !isProcessesPage && !isTodoNewPage && !isLifecyclePage);

  // The native menu-bar webview stays resident when the popover is dismissed and
  // reports its page hidden even while it is on screen (see useDocumentVisible),
  // so the Page Visibility API is the wrong signal there: gating the live streams
  // on `visible` leaves the popover stuck on stale or empty process/PR data. Treat
  // the menu-bar as always active — the open SSE stream is what keeps the backend
  // sampling — while the desktop tab keeps the pause-on-hide optimization.
  const streamsActive = visible || isMenubar;

  const {
    snapshot,
    projects: allProjects,
    projectsLoaded,
    projectError,
    processError,
    updateSnapshot,
    refreshProjects,
    refreshProjectsAndProcesses,
  } = useAppQueries({ enabled: streamsActive, hostData: hostControls, initialConfig });
  const {
    error: mutationError,
    markSeen,
    refresh: refreshPRs,
    togglePause,
    saveConfig,
    setIncludeBots,
    setShowClosed,
    setProjectHidden,
  } = useAppMutations();
  const { detail, loading: detailLoading, refresh: refreshDetail } = usePRDetailStream(selected);
  const rawPrs = snapshot.prs;
  const viewer = snapshot.viewer ?? '';
  const botsAvailable = !!snapshot.botsAvailable;
  const includeBotsServer = !!snapshot.includeBots;
  const showClosedServer = !!snapshot.showClosed;
  const unread = snapshot.unread ?? {};
  const fetchedAt = snapshot.fetchedAt;
  const nextFetchIn = snapshot.nextFetchIn;
  const error = mutationError || snapshot.error || processError;
  const config = snapshot.config;
  const paused = snapshot.paused;
  const rateLimit = snapshot.rateLimit;
  const syncStatus = snapshot.syncStatus ?? {};
  const gavelResultsMap = snapshot.gavelResults ?? {};

  // The native menu-bar webview is a small always-on-top popover, so shrink the
  // root font-size a notch: Tailwind's text and spacing are rem-based, so the
  // whole compact window scales down together and fits more without restyling
  // every element. Scoped to the webview — the desktop app keeps 16px.
  useEffect(() => {
    if (!isMenubar) return;
    const root = document.documentElement;
    const previous = root.style.fontSize;
    root.style.fontSize = '13px';
    return () => { root.style.fontSize = previous; };
  }, [isMenubar]);

  const prs = useMemo(() => annotateRoutePaths(rawPrs), [rawPrs]);
  // /api/projects also returns hidden projects; only the project picker lists
  // them (so they can be shown again). Everything else sees visible ones.
  const visibleProjects = useMemo(() => allProjects.filter(project => !project.hidden), [allProjects]);
  const hiddenDirs = useMemo(() => allProjects.filter(project => project.hidden).map(project => project.dir), [allProjects]);
  const scopeProjectHidden = allProjects.some(project => project.hidden && project.name === scopeProject);
  const projectScope = useMemo(
    () => applyProjectScope(visibleProjects, prs, scopeProject),
    [visibleProjects, prs, scopeProject],
  );
  const scopedProjects = projectScope.projects;
  const scopedPRs = projectScope.prs;

  const routeState: RouteState = useMemo(
    () => ({ tab: activeTab, selectedPath, scopeProject, projectDiffPath, projectRunId, projectHistory, projectResults, projectRef, todoView, filters }),
    [activeTab, selectedPath, scopeProject, projectDiffPath, projectRunId, projectHistory, projectResults, projectRef, todoView, filters],
  );

  const commitRoute = useCallback((next: RouteState, mode: 'push' | 'replace' = 'push') => {
    setActiveTab(tabFor(currentAuthMode(), next.tab));
    setSelectedPath(next.selectedPath);
    setScopeProject(next.scopeProject);
    setProjectDiffPath(next.projectDiffPath);
    setProjectRunId(next.projectRunId);
    setProjectHistory(next.projectHistory);
    setProjectResults(next.projectResults);
    setProjectRef(next.projectRef);
    setTodoView(next.todoView);
    setFilters(next.filters);
    const url = buildRoute(next);
    const current = `${window.location.pathname}${window.location.search}`;
    if (url !== current) {
      if (mode === 'replace') window.history.replaceState({}, '', url);
      else window.history.pushState({}, '', url);
    }
  }, []);

  // Switching the top-level tab navigates (so /todos, /activity are linkable and
  // back/forward works); the PR selection is dropped when leaving the prs tab.
  const changeTab = useCallback((next: Tab) => {
    commitRoute({ tab: next, selectedPath: '', scopeProject, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, projectRef: '', todoView: {}, filters });
  }, [commitRoute, filters, scopeProject]);

  // Selecting a todo encodes its ref in the path (/todos/{guid}) so a todo is
  // deep-linkable and back/forward works, mirroring PR selection. An empty id
  // clears the selection back to /todos.
  const navigateTodo = useCallback((id: string) => {
    commitRoute({ tab: 'todos', selectedPath: id, scopeProject, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, projectRef: '', todoView: {}, filters });
  }, [commitRoute, filters, scopeProject]);

  // The selected todo's detail tab, session inspector tab, and selected
  // attempts ride in the query (/todos/{ref}?tab=session&sessionTab=costs), so
  // a view within a todo is linkable too. Tab clicks push; attempt checkbox
  // toggles replace, so they don't flood the history.
  const navigateTodoView = useCallback((view: TodoDetailView, mode: 'push' | 'replace' = 'push') => {
    commitRoute({ ...routeState, todoView: view }, mode);
  }, [commitRoute, routeState]);

  const navigateSession = useCallback((id: string) => {
    commitRoute({ tab: 'sessions', selectedPath: id, scopeProject, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, projectRef: '', todoView: {}, filters });
  }, [commitRoute, filters, scopeProject]);

  // A session's worktree opens in its project's detail pane on that ref.
  const navigateSessionWorktree = useCallback((project: string, path: string) => {
    commitRoute({ tab: 'projects', selectedPath: project, scopeProject, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, projectRef: worktreeRef(path), todoView: {}, filters });
  }, [commitRoute, filters, scopeProject]);

  const navigateTask = useCallback((id: string | null) => {
    commitRoute({ tab: 'tasks', selectedPath: id ?? '', scopeProject, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, projectRef: '', todoView: {}, filters });
  }, [commitRoute, filters, scopeProject]);

  const navigateRun = useCallback((id: string) => {
    commitRoute({ tab: 'runs', selectedPath: id, scopeProject, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, projectRef: '', todoView: {}, filters });
  }, [commitRoute, filters, scopeProject]);

  const navigateProject = useCallback((name: string) => {
    commitRoute({ tab: 'projects', selectedPath: name, scopeProject, projectDiffPath: '', projectRunId: '', projectHistory, projectResults, projectRef: '', todoView: {}, filters });
  }, [commitRoute, filters, projectHistory, projectResults, scopeProject]);

  const navigateProjectRun = useCallback((project: string, runId: string) => {
    commitRoute({ tab: 'projects', selectedPath: project, scopeProject, projectDiffPath: '', projectRunId: runId, projectHistory: true, projectResults, projectRef: '', todoView: {}, filters });
  }, [commitRoute, filters, projectResults, scopeProject]);

  const navigateProjectDiff = useCallback((path: string) => {
    commitRoute({ tab: 'projects', selectedPath, scopeProject, projectDiffPath: path, projectRunId: '', projectHistory, projectResults, projectRef, todoView: {}, filters });
  }, [commitRoute, filters, projectHistory, projectRef, projectResults, scopeProject, selectedPath]);

  // Switching the worktree/branch drops the diff selection: a path from one ref
  // means nothing in another.
  const navigateProjectRef = useCallback((ref: string) => {
    commitRoute({ tab: 'projects', selectedPath, scopeProject, projectDiffPath: '', projectRunId: '', projectHistory, projectResults, projectRef: ref, todoView: {}, filters });
  }, [commitRoute, filters, projectHistory, projectResults, scopeProject, selectedPath]);

  const setProjectHistoryEnabled = useCallback((enabled: boolean) => {
    commitRoute({
      tab: 'projects',
      selectedPath,
      scopeProject,
      projectDiffPath: projectRunId ? '' : projectDiffPath,
      projectRunId: enabled ? projectRunId : '',
      projectHistory: enabled,
      projectResults,
      projectRef,
      todoView: {},
      filters,
    });
  }, [commitRoute, filters, projectDiffPath, projectRef, projectResults, projectRunId, scopeProject, selectedPath]);

  // The Todos data layer is mounted permanently so its chrome can live in the
  // AppShell's body slots, but only fetches while the Todos tab is active — or
  // while the ⌘K palette is open, so its global search can span todos from any
  // tab. The selectedPath is the todo ref on that tab (a PR route path otherwise).
  const onTodosTab = activeTab === 'todos' && !isTodoNewPage;
  const todos = useWorkspaceTodos(scopedProjects, {
    selectedId: onTodosTab ? selectedPath : '',
    onNavigate: navigateTodo,
    view: todoView,
    onViewChange: navigateTodoView,
    enabled: onTodosTab || paletteOpen,
  });
  // Selecting a prompt encodes its id in the path (/prompts/{id}) and keeps the
  // dashboard project scope in the query so the page remains deep-linkable.
  const navigatePrompt = useCallback((id: string, scope: string) => {
    commitRoute({ tab: 'prompts', selectedPath: id, scopeProject: scope, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, projectRef: '', todoView: {}, filters: scope === scopeProject ? filters : { ...filters, repos: {} } });
    if (scope !== scopeProject) todos.setFilters({ ...todos.filters, workspaces: {} });
  }, [commitRoute, filters, scopeProject, todos]);
  const review = useReviewMode(todos);
  // The full-width todos layout is a body-sidebar toggle, not a bodySplit tweak:
  // SplitPane seeds its width from defaultSplit once and never re-reads the
  // prop, so a percentage change would do nothing to a mounted pane. Dropping
  // the sidebar is also the only way to reach contentWidth — AppShell forces
  // "full" whenever a body sidebar is present.
  const todosFullWidth = onTodosTab && todos.layout === 'full';
  // ⌘K / Ctrl+K toggles the global search palette from anywhere (even while a
  // field is focused — it isn't a text-editing shortcut).
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && !e.altKey && (e.key === 'k' || e.key === 'K')) {
        e.preventDefault();
        setPaletteOpen(o => !o);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  useEffect(() => {
    const onPopState = () => {
      const next = parseRoute(window.location);
      setActiveTab(tabFor(currentAuthMode(), next.tab));
      setSelectedPath(next.selectedPath);
      setScopeProject(next.scopeProject);
      setProjectDiffPath(next.projectDiffPath);
      setProjectRunId(next.projectRunId);
      setProjectHistory(next.projectHistory);
      setProjectResults(next.projectResults);
      setProjectRef(next.projectRef);
      setTodoView(next.todoView);
      setFilters(next.filters);
    };
    window.addEventListener('popstate', onPopState);
    return () => window.removeEventListener('popstate', onPopState);
  }, []);

  useEffect(() => {
    if (isMenubar || !scopeProject || new URLSearchParams(window.location.search).has('project')) return;
    window.history.replaceState({}, '', buildRoute(routeState));
  }, [isMenubar, routeState, scopeProject]);

  useEffect(() => {
    if (!projectsLoaded || syncedScopeProject.current === scopeProject) return;
    if (!scopeProject) {
      syncedScopeProject.current = '';
      saveConfig({ ...config, project: '', repos: [], org: '', all: true });
      return;
    }
    const project = visibleProjects.find(candidate => candidate.name === scopeProject);
    if (!project) return;
    syncedScopeProject.current = project.name;
    saveConfig({ ...config, project: project.name, repos: project.repos, org: '', all: false });
  }, [config, visibleProjects, projectsLoaded, saveConfig, scopeProject]);

  // A project hidden while it is the scope (from the picker, the Projects
  // sidebar, or a deep link to one) drops back to All projects, like hideOrg.
  useEffect(() => {
    if (!scopeProjectHidden) return;
    updateScopeConfig({ project: '', repos: Array.from(new Set(visibleProjects.flatMap(project => project.repos))), org: '', all: false });
  }, [scopeProjectHidden]);

  // When PRs arrive (or the URL selection changes), reconcile `selected` with
  // the route's selectedPath. Fetches detail automatically for deep-linked PRs.
  useEffect(() => {
    // The menubar layout (native webview or mobile) drives PR selection through
    // its own local state, not the route, so skip route→selection reconciliation
    // there and let onSelect/onBack own it.
    if (useMenubarLayout || isProcessesPage || activeTab !== 'prs') return;
    if (!selectedPath) {
      if (selected) setSelected(null);
      return;
    }
    if (selected && selected.route_path === selectedPath) return;
    const target = findPRByRoutePath(prs, selectedPath);
    if (target) {
      loadPR(target);
      return;
    }
    const direct = prFromRoutePath(selectedPath);
    if (direct) {
      loadPR(direct);
    }
  }, [activeTab, selectedPath, prs]);

  const onProcChanged = useCallback(() => {
    void refreshProjectsAndProcesses();
  }, [refreshProjectsAndProcesses]);

  const projectsByRepo = useMemo(() => {
    const m: Record<string, Project> = {};
    for (const p of scopedProjects) for (const r of p.repos || []) m[r] = p;
    return m;
  }, [scopedProjects]);

  const openAdd = useCallback(() => setAddOpen(true), []);
  const openGlobalSettings = useCallback(() => setSettingsScope({ kind: 'global' }), []);
  const openProjectSettings = useCallback(
    (p: Project) => setSettingsScope({ kind: 'project', project: p }),
    [],
  );

  useEffect(() => { saveUIState(config, filters); }, [config, filters]);

  useEffect(() => {
    if (!detail?.pr) return;
    setSelected(current => current && current.repo === selected?.repo && current.number === selected.number
      ? mergePRItemFromDetail(current, detail.pr!)
      : current);
  }, [detail?.pr, selected?.number, selected?.repo]);

  useEffect(() => {
    if (!selected || !detail?.gavelResults) return;
    const aggregate = aggregateGavelShards(detail.gavelResults);
    if (!aggregate) return;
    const key = prKey(selected);
    updateSnapshot(current => ({
      ...current,
      gavelResults: { ...current.gavelResults, [key]: aggregate },
    }));
  }, [detail?.gavelResults, selected?.number, selected?.repo, updateSnapshot]);

  function loadPR(pr: PRItem) {
    if (selected?.repo === pr.repo && selected.number === pr.number) refreshDetail();
    setSelected(pr);
    markSeen(pr);
  }

  function handleSelect(pr: PRItem) {
    commitRoute({ ...routeState, selectedPath: pr.route_path || `${pr.repo}/${pr.number}` });
    loadPR(pr);
  }

  function clearSelectedPR() {
    setSelected(null);
  }

  // Closing the desktop detail pane tears down the stream + state and drops the
  // selection from the URL, so the pane returns to its empty state and browser
  // back reopens the PR (selecting a PR pushes history the same way).
  function closeSelectedPR() {
    clearSelectedPR();
    commitRoute({ ...routeState, selectedPath: '' });
  }

  function handleFiltersChange(next: Filters) {
    commitRoute({ ...routeState, filters: next });
  }

  function handleRefresh() {
    refreshPRs();
  }

  function handlePause() {
    togglePause();
  }

  function updateConfig(partial: Partial<SearchConfig>) {
    saveConfig({ ...config, ...partial });
  }

  function hideProject(name: string, hidden: boolean) {
    setProjectHidden({ name, hidden });
  }

  function updateScopeConfig(partial: Partial<SearchConfig>) {
    if (partial.project !== undefined && partial.project !== scopeProject) {
      syncedScopeProject.current = partial.project;
      if (isMenubar) {
        setScopeProject(partial.project);
        setSelectedPath('');
        setFilters({ ...filters, repos: {} });
      } else {
        commitRoute({
          ...routeState,
          scopeProject: partial.project,
          selectedPath: activeTab === 'prs' || activeTab === 'projects' || activeTab === 'todos' ? '' : routeState.selectedPath,
          projectDiffPath: activeTab === 'projects' ? '' : routeState.projectDiffPath,
          projectRunId: activeTab === 'projects' ? '' : routeState.projectRunId,
          projectRef: activeTab === 'projects' ? '' : routeState.projectRef,
          filters: { ...filters, repos: {} },
        });
      }
      if (activeTab === 'prs') setSelected(null);
      todos.setFilters({ ...todos.filters, workspaces: {} });
    }
    updateConfig(partial);
  }

  const onDownloadJSON = useCallback(() => downloadCurrentView(routeState, 'json'), [routeState]);
  const onDownloadMarkdown = useCallback(() => downloadCurrentView(routeState, 'md'), [routeState]);
  const onCopyForAgent = useCallback(async () => {
    if (copyState === 'copying') return;
    beginCopy();
    try {
      await copyCurrentViewForAgent(routeState);
      resetCopyFeedback('copied');
    } catch (e: any) {
      resetCopyFeedback('error', e?.message || 'Copy failed');
    }
  }, [copyState, routeState, beginCopy, resetCopyFeedback]);

  const counts = useMemo(() => computeCounts(scopedPRs), [scopedPRs]);
  const reposList = useMemo(() => collectRepos(scopedPRs), [scopedPRs]);
  const authors = useMemo(() => collectAuthors(scopedPRs, viewer, botsAvailable), [scopedPRs, viewer, botsAvailable]);

  // The @bots chip drives whether the daemon fetches bot PRs at all: when it's
  // not excluding bots, ask the server to include them (and refetch). Excluding
  // bots lets the server drop them at the source. Converges since the snapshot
  // echoes the server's includeBots back into includeBotsServer.
  useEffect(() => {
    const wantBots = (filters.authors['@bots'] ?? '') !== 'exclude';
    if (wantBots === includeBotsServer) return;
    setIncludeBots(wantBots);
  }, [filters.authors, includeBotsServer, setIncludeBots]);

  // Selecting the Closed or Merged State chip opts into fetching closed PRs (the
  // daemon syncs open-only by default). Asking the server to widen the fetch
  // triggers a refetch; deselecting both narrows it back to open. Converges since
  // the snapshot echoes the server's showClosed back into showClosedServer.
  useEffect(() => {
    const wantClosed = filters.state['closed'] === 'include' || filters.state['merged'] === 'include';
    if (wantClosed === showClosedServer) return;
    setShowClosed(wantClosed);
  }, [filters.state, setShowClosed, showClosedServer]);
  const filtered = useMemo(
    () => filterPRs(scopedPRs, filters.state, filters.checks, filters.repos, filters.authors, viewer),
    [scopedPRs, filters, viewer],
  );

  // The ⌘K palette searches PRs and todos in the current dashboard scope,
  // independent of the structured facet filters, and jumps to the chosen item.
  const todoEntries = useMemo(() => flattenTodos(todos.workspaces, todos.byDir), [todos.workspaces, todos.byDir]);

  // Both halves of the projects tab (the AppShell body sidebar and the detail
  // pane) read one catalog, so it is loaded here rather than inside either one.
  // A hosted org's projects tab lists its repositories instead, so neither the
  // catalog nor the git summary loads there.
  const projectCatalog = useProjectCatalog({ configured: scopedProjects, hiddenDirs, selectedName: selectedPath, enabled: hostControls && activeTab === 'projects' && projectHistory });
  // The git work summary shown on every project row refreshes slowly and only
  // matters while the Projects tab is on screen.
  const projectGitSummary = useProjectGitSummary({ enabled: hostControls && activeTab === 'projects' && visible });
  // The compact layout (mobile or the native menubar) keeps its tab and open
  // session in the URL on mobile, and in local state in the menubar webview,
  // which must stay on /menubar.
  const mobileTab: MobileTab = isMenubar ? menubarTab : mobileTabFromRoute(activeTab);
  const mobileSessionId = isMenubar ? menubarSessionId : activeTab === 'sessions' ? selectedPath : '';
  const changeMobileTab = useCallback((next: MobileTab) => {
    if (isMenubar) {
      setMenubarTab(next);
      setMenubarSessionId('');
    } else changeTab(routeTabFromMobile(next));
  }, [changeTab, isMenubar]);
  const openMobileSession = useCallback((id: string) => {
    if (isMenubar) setMenubarSessionId(id);
    else navigateSession(id);
  }, [isMenubar, navigateSession]);
  const agentSessions = useAgentSessions({
    enabled: (useMenubarLayout ? mobileTab === 'sessions' : activeTab === 'sessions') && streamsActive,
  });
  function selectPRFromPalette(pr: PRItem) {
    commitRoute({ tab: 'prs', selectedPath: pr.route_path || `${pr.repo}/${pr.number}`, scopeProject, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, projectRef: '', todoView: {}, filters });
    loadPR(pr);
  }
  function selectTodoFromPalette(entry: TodoEntry) {
    commitRoute({ tab: 'todos', selectedPath: entry.todo.ref, scopeProject, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, projectRef: '', todoView: {}, filters });
  }
  function openUUIDFromPalette(uuid: string) {
    // Keep the pasted identity in the URL. The global detail endpoint resolves
    // Todo UUIDs directly and Captain/provider session UUIDs through durable
    // prompt-run links, so reload/back navigation preserves the same lookup.
    commitRoute({ tab: 'todos', selectedPath: uuid, scopeProject, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, projectRef: '', todoView: {}, filters });
  }

  if (isMobile && isLifecyclePage) {
    return <div className="flex h-dvh min-w-0 flex-col bg-background text-foreground">
      <a className="shrink-0 border-b border-border px-4 py-3 text-sm text-primary" href="/prs">Back to dashboard</a>
      <PromptsView projects={scopedProjects} scopeProject={scopeProject} selectedId={selectedPath} onNavigate={navigatePrompt} />
    </div>;
  }

  if (useMenubarLayout) {
    return (
      <MenubarView
        prs={filtered}
        selected={selected}
        detail={detail}
        detailLoading={detailLoading}
        unread={unread}
        config={{ ...config, project: scopeProject }}
        projects={scopedProjects}
        availableProjects={allProjects}
        onSetProjectHidden={hideProject}
        projectsLoaded={projectsLoaded}
        projectError={projectError}
        projectsByRepo={projectsByRepo}
        syncStatus={syncStatus}
        gavelResults={gavelResultsMap}
        onSelect={loadPR}
        onBack={clearSelectedPR}
        onScopeChange={updateScopeConfig}
        onProcChanged={onProcChanged}
        fetchedAt={fetchedAt}
        error={error}
        tab={mobileTab}
        onTabChange={changeMobileTab}
        sessions={agentSessions}
        sessionId={mobileSessionId}
        onSessionChange={openMobileSession}
        chatEnabled={!isMenubar}
      />
    );
  }

  if (isProcessesPage) {
    return (
      <ProcessesPage
        projects={scopedProjects}
        projectsLoaded={projectsLoaded}
        projectError={projectError}
        onProcChanged={onProcChanged}
      />
    );
  }

  if (isTodoNewPage) {
    return <TodoNewPage projects={visibleProjects} projectError={projectError} />;
  }

  return (
    <>
      <AppShell
        brand={<img src="/brand/gavel-logo.svg" alt="gavel" className="h-7" />}
        nav={<TabBar active={activeTab} onChange={changeTab} />}
        search={<SearchTrigger onOpen={() => setPaletteOpen(true)} />}
        actions={
          <>
            {activeTab === 'todos' && <TodoReviewButton review={review} />}
            {activeTab === 'todos' && <TodoNavbarLayoutPicker todos={todos} />}
            {activeTab === 'todos' && <TodoNavbarDensityPicker todos={todos} />}
            {activeTab === 'todos' && <ReactGrabHelp />}
            {activeTab === 'todos' && <TodoNewButton todos={todos} />}
            {hostControls && <ChatButton label="Open Gavel assistant" />}
            {hostControls && <TaskManagerButton basePath="/api/v1" />}
            {hostControls && <ProcessManager projects={scopedProjects} onProcChanged={onProcChanged} />}
            {githubScopeVisible(authMode) && (
              <OrgChooser config={{ ...config, project: scopeProject }} projects={allProjects} onChange={updateScopeConfig} onSetProjectHidden={hideProject} />
            )}
            <StatusIndicator
              fetchedAt={fetchedAt}
              error={error}
              nextFetchIn={nextFetchIn}
              paused={paused}
              rateLimit={rateLimit}
              onRefresh={handleRefresh}
              onPause={handlePause}
              networkBusy={detailLoading}
            />
            {hostControls && (
              <Button
                variant="ghost"
                size="icon"
                type="button"
                onClick={openGlobalSettings}
                title="Global settings (~/.gavel.yaml)"
                aria-label="Global settings"
                className="text-muted-foreground hover:text-foreground"
              >
                <UiCog />
              </Button>
            )}
            <UserMenu />
            <ThemeToggle />
          </>
        }
        toolbar={
          activeTab === 'prs' ? (
            <FilterBar filters={filters} onChange={handleFiltersChange} counts={counts} repos={reposList} authors={authors} />
          ) : undefined
        }
        bodyHeader={
          activeTab === 'prs' ? (
            <span className="text-xs text-muted-foreground">
              {filtered.length} pull request{filtered.length !== 1 ? 's' : ''}
            </span>
          ) : undefined
        }
        bodyActions={
          activeTab === 'prs' ? (
            <ExportButtons
              onJSON={onDownloadJSON}
              onMarkdown={onDownloadMarkdown}
              onCopy={onCopyForAgent}
              copyState={copyState}
              copyError={copyError}
            />
          ) : undefined
        }
        bodySidebar={
          activeTab === 'prs' ? (
            <PRList prs={filtered} selected={selected} onSelect={handleSelect} unread={unread} syncStatus={syncStatus} gavelResults={gavelResultsMap} projectsByRepo={projectsByRepo} onProcChanged={onProcChanged} />
          ) : activeTab === 'projects' && !hostControls ? (
            <HostedProjectsSidebar projects={visibleProjects} selectedName={selectedPath} onSelect={navigateProject} />
          ) : activeTab === 'projects' ? (
            <ProjectsSidebar
              catalog={projectCatalog}
              gitSummary={projectGitSummary}
              selectedName={selectedPath}
              selectedRunId={projectRunId}
              historyEnabled={projectHistory}
              onHistoryChange={setProjectHistoryEnabled}
              onSelect={navigateProject}
              onSelectRun={navigateProjectRun}
              onChanged={onProcChanged}
              onAdd={openAdd}
              onSettings={openProjectSettings}
              onHide={project => hideProject(project.name, true)}
            />
          ) : activeTab === 'todos' && !todosFullWidth ? (
            <TodoWorkspaceList todos={todos} projectsLoaded={projectsLoaded} projectError={projectError} />
          ) : activeTab === 'sessions' ? (
            <SessionsSidebar query={agentSessions} selectedId={selectedPath} scopeProject={scopeProject} onSelect={navigateSession} />
          ) : undefined
        }
        bodySplit={bodySplitByTab[activeTab] ?? 38}
        // Only the full-width todos layout opts out of AppShell's default: the
        // table wants the whole viewport, but the detail wants the contained
        // measure so markdown and session logs stay readable. Every other tab
        // passes a body sidebar, which AppShell already forces to full width.
        contentWidth={todosFullWidth ? (todos.selected ? 'contained' : 'full') : undefined}
        contentClassName="overflow-hidden"
      >
        {activeTab === 'prs' ? (
          selected ? (
            <PRDetailPanel pr={selected} detail={detail} loading={detailLoading} projects={scopedProjects} onTodoCreated={refreshProjects} onActionDone={() => { if (selected) loadPR(selected); }} onClose={closeSelectedPR} />
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
              <div className="text-center">
                <UiGitPr className="text-4xl mb-2" />
                <p>Select a PR to view details</p>
              </div>
            </div>
          )
        ) : activeTab === 'projects' && !hostControls ? (
          <HostedProjectPane project={visibleProjects.find(project => project.name === selectedPath)} onSelectRun={navigateRun} />
        ) : activeTab === 'projects' ? (
          <ProjectDetailPane
            catalog={projectCatalog}
            selectedName={selectedPath}
            selectedRunId={projectRunId}
            diffPath={projectDiffPath}
            projectRef={projectRef}
            resultsEnabled={projectResults}
            onRefChange={navigateProjectRef}
            onDiffPathChange={navigateProjectDiff}
            onChanged={onProcChanged}
          />
        ) : activeTab === 'todos' ? (
          <div className="flex h-full min-h-0 flex-col">
            {projectError && (
              <div role="alert" className="shrink-0 border-b border-red-200 bg-red-50 px-4 py-2 text-xs text-red-700 dark:border-red-900 dark:bg-red-950/30 dark:text-red-300">
                {projectError}
              </div>
            )}
            <PlanReviewBar review={review} todos={todos} />
            <div className="min-h-0 flex-1">
              {todosFullWidth ? (
                <TodoFullPane todos={todos} projectsLoaded={projectsLoaded} scopeProject={scopeProject} navigationEnabled={!review.active} />
              ) : (
                <TodoDetailPane todos={todos} navigationEnabled={!review.active} />
              )}
            </div>
          </div>
        ) : activeTab === 'sessions' ? (
          <SessionsDetailPane query={agentSessions} selectedId={selectedPath} onOpenTodo={navigateTodo} onOpenWorktree={navigateSessionWorktree} />
        ) : activeTab === 'tasks' ? (
          <div className="h-full overflow-y-auto p-4">
            <TaskManager basePath="/api/v1" selectedId={selectedPath || undefined} onSelectRun={navigateTask} />
          </div>
        ) : activeTab === 'runs' ? (
          <RunsView selectedId={selectedPath} onSelect={navigateRun} />
        ) : activeTab === 'prompts' ? (
          <PromptsView
            projects={scopedProjects}
            scopeProject={scopeProject}
            selectedId={selectedPath}
            onNavigate={navigatePrompt}
          />
        ) : (
          <ActivityView />
        )}
      </AppShell>

      <CommandPalette
        open={paletteOpen}
        onClose={() => setPaletteOpen(false)}
        prs={prs}
        todos={todoEntries}
        todosLoading={todos.loadingList}
        onSelectPR={selectPRFromPalette}
        onSelectTodo={selectTodoFromPalette}
        onOpenUUID={openUUIDFromPalette}
      />

      <CreateTodoDialog open={todos.showCreate} onClose={() => todos.setShowCreate(false)} workspaces={visibleProjects.filter(project => !!project.dir)} onCreated={todos.created} defaultDir={todos.selected?.dir ?? scopedProjects[0]?.dir} />
      <AddProjectDialog open={addOpen} onClose={() => setAddOpen(false)} onSaved={onProcChanged} repoOptions={reposList} />
      {settingsScope && (
        <SettingsPage
          scope={settingsScope}
          repoOptions={reposList}
          onClose={() => setSettingsScope(null)}
          onSaved={onProcChanged}
        />
      )}
    </>
  );
}

function TabBar({ active, onChange }: { active: Tab; onChange: (t: Tab) => void }) {
  const allTabs: { id: Tab; label: string; icon: ComponentType<IconProps> }[] = [
    { id: 'prs', label: 'PRs', icon: UiGitPr },
    { id: 'projects', label: 'Projects', icon: UiFolderGit },
    { id: 'todos', label: 'Todos', icon: UiCheck },
    { id: 'runs', label: 'Runs', icon: UiPlay },
    { id: 'sessions', label: 'Sessions', icon: UiTerminal },
    { id: 'tasks', label: 'Tasks', icon: UiListChecks },
    { id: 'prompts', label: 'Prompts', icon: UiRobotAi },
    { id: 'activity', label: 'Activity', icon: UiActivity },
  ];
  const shown = new Set(visibleTabs(currentAuthMode(), allTabs.map(t => t.id)));
  const tabs = allTabs.filter(t => shown.has(t.id));
  return (
    <div className="flex gap-1 border-b border-transparent">
      {tabs.map(t => {
        const Icon = t.icon;
        return (
          <Button
            variant="ghost"
            key={t.id}
            onClick={() => onChange(t.id)}
            className={`px-3 py-1.5 text-sm rounded-md transition h-auto justify-start ${
              active === t.id
                ? 'bg-primary/10 text-primary font-medium'
                : 'text-muted-foreground hover:bg-muted'
            }`}
          >
            <Icon className="mr-1" />
            {t.label}
          </Button>
        );
      })}
    </div>
  );
}

function ExportButtons({ onJSON, onMarkdown, onCopy, copyState, copyError }: {
  onJSON: () => void;
  onMarkdown: () => void;
  onCopy: () => void;
  copyState: 'idle' | 'copying' | 'copied' | 'error';
  copyError: string;
}) {
  const CopyIcon = copyState === 'copied' ? UiCheck : copyState === 'copying' ? Spinner : UiCopy;
  return (
    <div className="flex items-center gap-1">
      <Button
        variant="ghost"
        className="text-xs px-2 py-1 rounded border border-border text-muted-foreground hover:bg-muted transition-colors h-auto justify-start"
        onClick={onJSON}
        title="Download current view as JSON"
      >
        <UiJson className="mr-0.5" />
        JSON
      </Button>
      <Button
        variant="ghost"
        className="text-xs px-2 py-1 rounded border border-border text-muted-foreground hover:bg-muted transition-colors h-auto justify-start"
        onClick={onMarkdown}
        title="Download current view as Markdown"
      >
        <UiMarkdown className="mr-0.5" />
        Markdown
      </Button>
      <Button
        variant="ghost"
        className={`text-xs px-2 py-1 rounded border transition-colors h-auto justify-start ${
          copyState === 'copied'
            ? 'border-green-300 bg-green-50 text-green-700'
            : copyState === 'error'
              ? 'border-red-300 bg-red-50 text-red-700'
              : 'border-gray-300 text-gray-600 hover:bg-gray-200'
        }`}
        onClick={onCopy}
        title={copyError || 'Copy Markdown export for agent'}
      >
        <CopyIcon className="mr-0.5" />
        {copyState === 'copying' ? 'Copying...' : copyState === 'copied' ? 'Copied' : copyState === 'error' ? 'Copy failed' : 'Copy for Agent'}
      </Button>
    </div>
  );
}
