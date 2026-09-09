import { useState, useEffect, useMemo, useCallback } from 'react';
import type { PRItem, PRDetail, PRInfo, SearchConfig, PRSyncStatus, GavelResultsSummary, Project, ProcStatus } from './types';
import { PRList } from './components/PRList';
import { PRDetailPanel } from './components/PRDetail';
import { FilterBar, emptyFilters, type Filters } from './components/FilterBar';
import { AppShell, Button } from '@flanksource/clicky-ui/components';
import { TaskManager, TaskManagerButton } from '@flanksource/clicky-ui/data';
import { ActivityView } from './components/ActivityView';
import { TodoNewButton, TodoNavbarDensityPicker, TodoNavbarLayoutPicker, TodoWorkspaceList, TodoDetailPane, TodoFullPane } from './components/TodoView';
import { useWorkspaceTodos } from './components/todos/useWorkspaceTodos';
import { PlanReviewBar, TodoReviewButton, useReviewMode } from './components/todos/PlanReview';
import { CreateTodoDialog } from './components/todos/CreateTodoDialog';
import { TodoNewPage } from './components/todos/TodoNewPage';
import { MenubarTodos } from './components/MenubarTodos';
import { StatusIndicator } from './components/StatusIndicator';
import { OrgChooser } from './components/OrgChooser';
import { AddProjectDialog } from './components/AddProjectDialog';
import { SettingsPage, type SettingsScope } from './components/settings/SettingsPage';
import { ProjectDetailPane, ProjectsSidebar } from './components/ProjectsView';
import { ProcessManager } from './components/ProcessManager';
import { ProjectsPlaceholder } from './components/ProjectsPlaceholder';
import { ThemeToggle } from './components/ThemeToggle';
import { ReactGrabHelp } from './components/ReactGrabHelp';
import { CommandPalette, SearchTrigger } from './components/CommandPalette';
import { flattenTodos, type TodoEntry } from './components/todos/todoGroup';
import { WorkspaceGroup } from './components/ProcessTable';
import { aggregateDotClass, aggregateGavelShards, computeCounts, collectRepos, collectAuthors, filterPRs, flattenProcesses, prKey, emptyProcStatus } from './utils';
import { useCopyFeedback } from './useCopyFeedback';
import {
  annotateRoutePaths,
  buildRoute,
  findPRByRoutePath,
  parseRoute,
  type RouteState,
  type Tab,
} from './routes';
import { copyCurrentViewForAgent, downloadCurrentView } from './export';
import { copyText } from './clipboard';
import { loadUIState, saveUIState, filtersFromStored } from './storage';
import { useDocumentVisible } from './useDocumentVisible';
import { useProjectCatalog } from './useProjectCatalog';
import { useAppQueries } from './useAppQueries';
import { useAppMutations } from './useAppMutations';
import { usePRDetailStream } from './usePRDetailStream';
import { useIsMobile } from './useIsMobile';
import { UiActivity, UiArrowLeft, UiCheck, UiClose, UiCog, UiCopy, UiFolderGit, UiGitPr, UiJson, UiLink, UiListChecks, UiMarkdown, UiRobotAi } from '@flanksource/clicky-ui/icons';
import { PromptsView } from './components/prompts/PromptsView';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import type { ComponentType } from 'react';
import { Spinner } from './icons/Spinner';

const defaultConfig: SearchConfig = { repos: [] };

// Percentage width of the AppShell body sidebar per tab. A PR row carries title,
// repo, checks and badges so it wants half the body; a project is just a name and
// a run list, so its list is narrow.
const bodySplitByTab: Partial<Record<Tab, number>> = { prs: 50, projects: 22, todos: 38 };

type WebKitExternalBridge = {
  webkit?: {
    messageHandlers?: {
      external?: {
        postMessage: (message: string) => void;
      };
    };
  };
};

const menubarOpenExternalMessage = 'gavel:open-external';
const menubarPointerEnterMessage = 'gavel:pointer-enter';
const menubarPointerLeaveMessage = 'gavel:pointer-leave';

function postMenubarMessage(type: string, payload: Record<string, unknown> = {}) {
  const bridge = (window as WebKitExternalBridge).webkit?.messageHandlers?.external;
  if (!bridge) return false;
  bridge.postMessage(JSON.stringify({ type, ...payload }));
  return true;
}

function useMenubarExternalLinks() {
  useEffect(() => {
    const onClick = (event: MouseEvent) => {
      if (event.defaultPrevented || event.button !== 0) return;
      const target = event.target;
      if (!(target instanceof Element)) return;
      const anchor = target.closest('a[href]');
      if (!(anchor instanceof HTMLAnchorElement) || !anchor.href) return;

      if (!postMenubarMessage(menubarOpenExternalMessage, { url: anchor.href })) return;

      event.preventDefault();
      event.stopPropagation();
      event.stopImmediatePropagation();
    };

    document.addEventListener('click', onClick, true);
    return () => document.removeEventListener('click', onClick, true);
  }, []);
}

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
    : { tab: 'prs', selectedPath: '', projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, promptScope: '', filters: emptyFilters() };

  // Hydrate org/search config and filters from localStorage. URL query params
  // (if present) win for filters so deep links still work.
  const stored = typeof window !== 'undefined' ? loadUIState() : {};
  const hasUrlFilters = typeof window !== 'undefined' && window.location.search.length > 1;
  // First-run default hides bots (the daemon now fetches them so the @bots
  // author chip can toggle them back on). URL params and stored filters win.
  const defaultFilters: Filters = { ...emptyFilters(), authors: { '@bots': 'exclude' } };
  const initialFilters = hasUrlFilters ? initialRoute.filters : (filtersFromStored(stored.filters) ?? defaultFilters);
  const initialConfig: SearchConfig = { ...defaultConfig, ...stored.config };

  const [selected, setSelected] = useState<PRItem | null>(null);
  const [filters, setFilters] = useState<Filters>(initialFilters);
  const [selectedPath, setSelectedPath] = useState(initialRoute.selectedPath);
  const [projectDiffPath, setProjectDiffPath] = useState(initialRoute.projectDiffPath);
  const [projectRunId, setProjectRunId] = useState(initialRoute.projectRunId);
  const [projectHistory, setProjectHistory] = useState(initialRoute.projectHistory);
  const [projectResults, setProjectResults] = useState(initialRoute.projectResults);
  const [promptScope, setPromptScope] = useState(initialRoute.promptScope);
  const [activeTab, setActiveTab] = useState<Tab>(initialRoute.tab);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const { copyState, copyError, beginCopy, resetCopyFeedback } = useCopyFeedback({ copiedMs: 2500, errorMs: 2500 });
  const [addOpen, setAddOpen] = useState(false);
  const [settingsScope, setSettingsScope] = useState<SettingsScope | null>(null);
  const visible = useDocumentVisible();
  const isMobile = useIsMobile();

  const pathname = typeof window !== 'undefined' ? window.location.pathname : '';
  const isMenubar = pathname === '/menubar';
  const isProcessesPage = pathname === '/processes';
  // The focused /todos/new form is a standalone page (like /menubar, /processes)
  // rather than a view inside the Todos tab, so it can be linked to and returns
  // to its referer. parseRoute reads it as the todos tab with selectedPath="new";
  // the early return below renders the form before any of that matters.
  const isTodoNewPage = pathname === '/todos/new';
  // On mobile the desktop AppShell (split panes, sidebar, toolbars) has no room,
  // so the dashboard falls back to the same compact, single-column menubar
  // dropdown layout the native menubar webview uses. The standalone /processes
  // and /todos/new pages already lay out fine on narrow screens, so they keep
  // their own views.
  const useMenubarLayout = isMenubar || (isMobile && !isProcessesPage && !isTodoNewPage);

  // The native menu-bar webview stays resident when the popover is dismissed and
  // reports its page hidden even while it is on screen (see useDocumentVisible),
  // so the Page Visibility API is the wrong signal there: gating the live streams
  // on `visible` leaves the popover stuck on stale or empty process/PR data. Treat
  // the menu-bar as always active — the open SSE stream is what keeps the backend
  // sampling — while the desktop tab keeps the pause-on-hide optimization.
  const streamsActive = visible || isMenubar;

  const {
    snapshot,
    projects,
    projectsLoaded,
    projectError,
    procStatus,
    processError,
    updateSnapshot,
    refreshProjects,
    refreshProjectsAndProcesses,
  } = useAppQueries({ enabled: streamsActive, initialConfig });
  const {
    error: mutationError,
    markSeen,
    refresh: refreshPRs,
    togglePause,
    saveConfig,
    setIncludeBots,
    setShowClosed,
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

  const routeState: RouteState = useMemo(
    () => ({ tab: activeTab, selectedPath, projectDiffPath, projectRunId, projectHistory, projectResults, promptScope, filters }),
    [activeTab, selectedPath, projectDiffPath, projectRunId, projectHistory, projectResults, promptScope, filters],
  );

  const commitRoute = useCallback((next: RouteState, mode: 'push' | 'replace' = 'push') => {
    setActiveTab(next.tab);
    setSelectedPath(next.selectedPath);
    setProjectDiffPath(next.projectDiffPath);
    setProjectRunId(next.projectRunId);
    setProjectHistory(next.projectHistory);
    setProjectResults(next.projectResults);
    setPromptScope(next.promptScope);
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
    commitRoute({ tab: next, selectedPath: '', projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, promptScope: '', filters });
  }, [commitRoute, filters]);

  // Selecting a todo encodes its ref in the path (/todos/{guid}) so a todo is
  // deep-linkable and back/forward works, mirroring PR selection. An empty id
  // clears the selection back to /todos.
  const navigateTodo = useCallback((id: string) => {
    commitRoute({ tab: 'todos', selectedPath: id, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, promptScope: '', filters });
  }, [commitRoute, filters]);

  const navigateTask = useCallback((id: string | null) => {
    commitRoute({ tab: 'tasks', selectedPath: id ?? '', projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, promptScope: '', filters });
  }, [commitRoute, filters]);

  const navigateProject = useCallback((name: string) => {
    commitRoute({ tab: 'projects', selectedPath: name, projectDiffPath: '', projectRunId: '', projectHistory, projectResults, promptScope: '', filters });
  }, [commitRoute, filters, projectHistory, projectResults]);

  const navigateProjectRun = useCallback((project: string, runId: string) => {
    commitRoute({ tab: 'projects', selectedPath: project, projectDiffPath: '', projectRunId: runId, projectHistory: true, projectResults, promptScope: '', filters });
  }, [commitRoute, filters, projectResults]);

  const navigateProjectDiff = useCallback((path: string) => {
    commitRoute({ tab: 'projects', selectedPath, projectDiffPath: path, projectRunId: '', projectHistory, projectResults, promptScope: '', filters });
  }, [commitRoute, filters, projectHistory, projectResults, selectedPath]);

  const setProjectHistoryEnabled = useCallback((enabled: boolean) => {
    commitRoute({
      tab: 'projects',
      selectedPath,
      projectDiffPath: projectRunId ? '' : projectDiffPath,
      projectRunId: enabled ? projectRunId : '',
      projectHistory: enabled,
      projectResults,
      promptScope: '',
      filters,
    });
  }, [commitRoute, filters, projectDiffPath, projectResults, projectRunId, selectedPath]);

  // Selecting a prompt encodes its id in the path (/prompts/{id}) and the scope
  // project as a query so a prompt page is deep-linkable per scope.
  const navigatePrompt = useCallback((id: string, scope: string) => {
    commitRoute({ tab: 'prompts', selectedPath: id, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, promptScope: scope, filters });
  }, [commitRoute, filters]);

  // The Todos data layer is mounted permanently so its chrome can live in the
  // AppShell's body slots, but only fetches while the Todos tab is active — or
  // while the ⌘K palette is open, so its global search can span todos from any
  // tab. The selectedPath is the todo ref on that tab (a PR route path otherwise).
  const onTodosTab = activeTab === 'todos' && !isTodoNewPage;
  const todos = useWorkspaceTodos(projects, onTodosTab ? selectedPath : '', navigateTodo, onTodosTab || paletteOpen);
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
      setActiveTab(next.tab);
      setSelectedPath(next.selectedPath);
      setProjectDiffPath(next.projectDiffPath);
      setProjectRunId(next.projectRunId);
      setProjectHistory(next.projectHistory);
      setProjectResults(next.projectResults);
      setPromptScope(next.promptScope);
      setFilters(next.filters);
    };
    window.addEventListener('popstate', onPopState);
    return () => window.removeEventListener('popstate', onPopState);
  }, []);

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
    for (const p of projects) for (const r of p.repos || []) m[r] = p;
    return m;
  }, [projects]);

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

  const counts = useMemo(() => computeCounts(prs), [prs]);
  const reposList = useMemo(() => collectRepos(prs), [prs]);
  const authors = useMemo(() => collectAuthors(prs, viewer, botsAvailable), [prs, viewer, botsAvailable]);

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
    () => filterPRs(prs, filters.state, filters.checks, filters.repos, filters.authors, viewer),
    [prs, filters, viewer],
  );

  // The ⌘K palette searches the full PR list and every workspace's todos (flattened
  // across workspaces), independent of the structured facet filters, and jumps to
  // the chosen item — switching tabs as needed.
  const todoEntries = useMemo(() => flattenTodos(todos.workspaces, todos.byDir), [todos.workspaces, todos.byDir]);

  // Both halves of the projects tab (the AppShell body sidebar and the detail
  // pane) read one catalog, so it is loaded here rather than inside either one.
  const projectCatalog = useProjectCatalog({ configured: projects, selectedName: selectedPath, enabled: activeTab === 'projects' && projectHistory });
  function selectPRFromPalette(pr: PRItem) {
    commitRoute({ tab: 'prs', selectedPath: pr.route_path || `${pr.repo}/${pr.number}`, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, promptScope: '', filters });
    loadPR(pr);
  }
  function selectTodoFromPalette(entry: TodoEntry) {
    commitRoute({ tab: 'todos', selectedPath: entry.todo.ref, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, promptScope: '', filters });
  }
  function openUUIDFromPalette(uuid: string) {
    // Keep the pasted identity in the URL. The global detail endpoint resolves
    // Todo UUIDs directly and Captain/provider session UUIDs through durable
    // prompt-run links, so reload/back navigation preserves the same lookup.
    commitRoute({ tab: 'todos', selectedPath: uuid, projectDiffPath: '', projectRunId: '', projectHistory: false, projectResults: false, promptScope: '', filters });
  }

  if (useMenubarLayout) {
    return (
      <MenubarView
        prs={filtered}
        selected={selected}
        detail={detail}
        detailLoading={detailLoading}
        unread={unread}
        projects={projects}
        projectsLoaded={projectsLoaded}
        projectError={projectError}
        projectsByRepo={projectsByRepo}
        procStatus={procStatus}
        syncStatus={syncStatus}
        gavelResults={gavelResultsMap}
        onSelect={loadPR}
        onBack={clearSelectedPR}
        onProcChanged={onProcChanged}
        fetchedAt={fetchedAt}
        error={error}
      />
    );
  }

  if (isProcessesPage) {
    return (
      <ProcessesPage
        projects={projects}
        projectsLoaded={projectsLoaded}
        projectError={projectError}
        procStatus={procStatus}
        onProcChanged={onProcChanged}
      />
    );
  }

  if (isTodoNewPage) {
    return <TodoNewPage projects={projects} procStatus={procStatus} projectError={projectError} />;
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
            <TaskManagerButton basePath="/api/v1" />
            <ProcessManager projects={projects} procStatus={procStatus} onProcChanged={onProcChanged} />
            <OrgChooser config={config} onChange={updateConfig} />
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
            <PRList prs={filtered} selected={selected} onSelect={handleSelect} unread={unread} syncStatus={syncStatus} gavelResults={gavelResultsMap} projectsByRepo={projectsByRepo} procStatus={procStatus} onProcChanged={onProcChanged} />
          ) : activeTab === 'projects' ? (
            <ProjectsSidebar
              catalog={projectCatalog}
              procStatus={procStatus}
              selectedName={selectedPath}
              selectedRunId={projectRunId}
              historyEnabled={projectHistory}
              onHistoryChange={setProjectHistoryEnabled}
              onSelect={navigateProject}
              onSelectRun={navigateProjectRun}
              onChanged={onProcChanged}
              onAdd={openAdd}
              onSettings={openProjectSettings}
            />
          ) : activeTab === 'todos' && !todosFullWidth ? (
            <TodoWorkspaceList todos={todos} projectsLoaded={projectsLoaded} projectError={projectError} />
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
            <PRDetailPanel pr={selected} detail={detail} loading={detailLoading} projects={projects} onTodoCreated={refreshProjects} onActionDone={() => { if (selected) loadPR(selected); }} onClose={closeSelectedPR} />
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
              <div className="text-center">
                <UiGitPr className="text-4xl mb-2" />
                <p>Select a PR to view details</p>
              </div>
            </div>
          )
        ) : activeTab === 'projects' ? (
          <ProjectDetailPane
            catalog={projectCatalog}
            selectedName={selectedPath}
            selectedRunId={projectRunId}
            diffPath={projectDiffPath}
            resultsEnabled={projectResults}
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
                <TodoFullPane todos={todos} projectsLoaded={projectsLoaded} navigationEnabled={!review.active} />
              ) : (
                <TodoDetailPane todos={todos} navigationEnabled={!review.active} />
              )}
            </div>
          </div>
        ) : activeTab === 'tasks' ? (
          <div className="h-full overflow-y-auto p-4">
            <TaskManager basePath="/api/v1" selectedId={selectedPath || undefined} onSelectRun={navigateTask} />
          </div>
        ) : activeTab === 'prompts' ? (
          <PromptsView
            projects={projects}
            scopeProject={promptScope}
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

      <CreateTodoDialog open={todos.showCreate} onClose={() => todos.setShowCreate(false)} workspaces={todos.workspaces} onCreated={todos.created} defaultDir={todos.selected?.dir} />
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

function ProcessesPage({
  projects,
  projectsLoaded,
  projectError,
  procStatus,
  onProcChanged,
}: {
  projects: Project[];
  projectsLoaded: boolean;
  projectError?: string;
  procStatus: Record<string, ProcStatus>;
  onProcChanged: () => void;
}) {
  const workspaces = useMemo(
    () => projects.map(p => ({ project: p, status: procStatus[p.name] ?? emptyProcStatus })),
    [projects, procStatus],
  );
  const procs = useMemo(() => flattenProcesses(projects, procStatus), [projects, procStatus]);
  const running = procs.filter(p => p.proc.status === 'running').length;
  const dot = aggregateDotClass(procs.map(p => p.proc));

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <div className="flex min-w-0 items-center gap-3">
          <img src="/brand/gavel-logo.svg" alt="gavel" className="h-7 shrink-0" />
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <span className={`inline-block h-2.5 w-2.5 rounded-full ${dot}`} />
              <span className="text-sm font-semibold">Processes</span>
            </div>
            <div className="text-xs text-muted-foreground">{running} running of {procs.length}</div>
          </div>
        </div>
        <a
          href="/prs"
          className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          title="Back to PR dashboard"
          aria-label="Back to PR dashboard"
        >
          <UiClose className="text-base" />
        </a>
      </header>

      <main className="mx-auto w-full max-w-6xl px-4 py-4">
        {workspaces.length > 0 ? (
          <div className="divide-y divide-border border-y border-border">
            {workspaces.map(w => (
              <WorkspaceGroup key={w.project.name} project={w.project} status={w.status} onChanged={onProcChanged} />
            ))}
          </div>
        ) : (
          <ProjectsPlaceholder
            loaded={projectsLoaded}
            error={projectError}
            emptyText="No projects configured"
            className="py-10 text-center text-sm text-muted-foreground"
          />
        )}
      </main>
    </div>
  );
}

// CopyLinkButton copies a PR's URL to the clipboard. It uses copyText so it
// works inside the native menu-bar WKWebView, where navigator.clipboard is
// denied and the async Clipboard API silently fails.
function CopyLinkButton({ url }: { url: string }) {
  const { copyState, beginCopy, resetCopyFeedback } = useCopyFeedback();
  const copied = copyState === 'copied';
  const errored = copyState === 'error';

  async function onCopy() {
    beginCopy();
    try {
      await copyText(url);
      resetCopyFeedback('copied');
    } catch {
      resetCopyFeedback('error');
    }
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      type="button"
      onClick={onCopy}
      className={`h-8 w-8 rounded-md hover:bg-muted hover:text-foreground ${copied ? 'text-green-600' : errored ? 'text-red-600' : 'text-muted-foreground'}`}
      title={copied ? 'Link copied' : errored ? 'Copy failed' : 'Copy link to this PR'}
      aria-label="Copy link to this PR"
    >
      {copied ? <UiCheck className="text-base" /> : <UiLink className="text-base" />}
    </Button>
  );
}

function MenubarView({
  prs,
  selected,
  detail,
  detailLoading,
  unread,
  projects,
  projectsLoaded,
  projectError,
  projectsByRepo,
  procStatus,
  syncStatus,
  gavelResults,
  onSelect,
  onBack,
  onProcChanged,
  fetchedAt,
  error,
}: {
  prs: PRItem[];
  selected: PRItem | null;
  detail: PRDetail | null;
  detailLoading: boolean;
  unread: Record<string, boolean>;
  projects: Project[];
  projectsLoaded: boolean;
  projectError?: string;
  projectsByRepo: Record<string, Project>;
  procStatus: Record<string, ProcStatus>;
  syncStatus: Record<string, PRSyncStatus>;
  gavelResults: Record<string, GavelResultsSummary>;
  onSelect: (pr: PRItem) => void;
  onBack: () => void;
  onProcChanged: () => void;
  fetchedAt: string;
  error?: string;
}) {
  useMenubarExternalLinks();
  const [menubarTab, setMenubarTab] = useState<'processes' | 'prs' | 'todos'>('prs');

  const workspaces = useMemo(
    () => projects.map(p => ({ project: p, status: procStatus[p.name] ?? emptyProcStatus })),
    [projects, procStatus],
  );
  const procs = useMemo(() => flattenProcesses(projects, procStatus), [projects, procStatus]);
  const running = procs.filter(p => p.proc.status === 'running').length;
  const dot = aggregateDotClass(procs.map(p => p.proc));
  const failed = prs.filter(pr => pr.checkStatus?.failed && pr.checkStatus.failed > 0).length;
  // Tab count/badge come from the projects poll's per-workspace todoCounts, so
  // the strip reflects todos without fetching the lists until the tab is opened.
  const openTodos = projects.reduce((n, p) => n + (p.todoCounts?.open ?? 0), 0);
  const failedTodos = projects.reduce((n, p) => n + (p.todoCounts?.failed ?? 0), 0);
  const unreadCount = prs.filter(pr => unread[prKey(pr)]).length;
  const fetched = fetchedAt ? new Date(fetchedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '';

  if (selected) {
    return (
      <div
        className="h-screen overflow-hidden bg-background text-foreground"
        onPointerEnter={() => postMenubarMessage(menubarPointerEnterMessage)}
        onPointerLeave={() => postMenubarMessage(menubarPointerLeaveMessage)}
      >
        <div className="flex h-11 items-center justify-between border-b border-border px-2">
          <div className="flex min-w-0 items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              type="button"
              onClick={onBack}
              className="h-8 w-8 rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
              title="Back to pull requests"
              aria-label="Back to pull requests"
            >
              <UiArrowLeft className="text-base" />
            </Button>
            <div className="min-w-0">
              <div className="truncate text-sm font-semibold">{selected.repo}#{selected.number}</div>
              <div className="truncate text-[11px] text-muted-foreground">{selected.title}</div>
            </div>
          </div>
          <div className="flex shrink-0 items-center gap-1">
            <CopyLinkButton url={selected.url} />
            <div className="text-[11px] text-muted-foreground tabular-nums">{error || fetched}</div>
          </div>
        </div>
        <div className="h-[calc(100vh-2.75rem)] overflow-hidden">
          <PRDetailPanel pr={selected} detail={detail} loading={detailLoading} projects={projects} onActionDone={() => { if (selected) onSelect(selected); }} />
        </div>
      </div>
    );
  }

  return (
    <div
      className="flex h-screen flex-col overflow-hidden bg-background text-foreground"
      onPointerEnter={() => postMenubarMessage(menubarPointerEnterMessage)}
      onPointerLeave={() => postMenubarMessage(menubarPointerLeaveMessage)}
    >
      <div className="flex shrink-0 items-center justify-between border-b border-border px-3 py-2">
        <div className="flex items-center gap-2 min-w-0">
          <span className={`inline-block h-2.5 w-2.5 rounded-full ${dot}`} />
          <div className="min-w-0">
            <div className="text-sm font-semibold leading-tight">Gavel</div>
            <div className="truncate text-[11px] text-muted-foreground">
              {running}/{procs.length} processes
              {failed > 0 ? ` · ${failed} failing` : ''}
              {unreadCount > 0 ? ` · ${unreadCount} unread` : ''}
            </div>
          </div>
        </div>
        {/* A failed project load is otherwise invisible in the menubar: every tab
            just renders empty. Surface it here, ahead of the PR poll's clock. */}
        {projectError ? (
          <div role="alert" className="truncate text-[11px] text-destructive">{projectError}</div>
        ) : (
          <div className="text-[11px] text-muted-foreground tabular-nums">{error || fetched}</div>
        )}
      </div>

      <div className="flex shrink-0 items-center gap-1 border-b border-border px-2 py-1">
        <MenubarTab
          label="Processes"
          dot={dot}
          count={`${running}/${procs.length}`}
          active={menubarTab === 'processes'}
          onClick={() => setMenubarTab('processes')}
        />
        <MenubarTab
          label="PRs"
          icon={UiGitPr}
          count={prs.length}
          badge={failed > 0 ? failed : undefined}
          active={menubarTab === 'prs'}
          onClick={() => setMenubarTab('prs')}
        />
        <MenubarTab
          label="Todos"
          icon={UiCheck}
          count={openTodos}
          badge={failedTodos > 0 ? failedTodos : undefined}
          active={menubarTab === 'todos'}
          onClick={() => setMenubarTab('todos')}
        />
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto">
        {menubarTab === 'processes' ? (
          workspaces.length > 0 ? (
            <div className="divide-y divide-border p-2">
              {workspaces.map(w => (
                <WorkspaceGroup key={w.project.name} project={w.project} status={w.status} onChanged={onProcChanged} />
              ))}
            </div>
          ) : (
            <ProjectsPlaceholder
              loaded={projectsLoaded}
              error={projectError}
              emptyText="No projects configured"
              className="px-3 py-6 text-center text-xs text-muted-foreground"
            />
          )
        ) : menubarTab === 'todos' ? (
          <MenubarTodos projects={projects} projectsLoaded={projectsLoaded} projectError={projectError} />
        ) : (
          <PRList
            prs={prs}
            selected={selected}
            onSelect={onSelect}
            unread={unread}
            syncStatus={syncStatus}
            gavelResults={gavelResults}
            projectsByRepo={projectsByRepo}
            procStatus={procStatus}
            onProcChanged={onProcChanged}
          />
        )}
      </div>
    </div>
  );
}

// MenubarTab is one segment of the menubar's Processes/PRs switcher. A status
// dot (process health) or an icon leads the label; count is the inline subtotal
// and badge is an attention-grabbing count (e.g. failing PRs).
function MenubarTab({ label, icon: Icon, dot, count, badge, active, onClick }: {
  label: string;
  icon?: ComponentType<IconProps>;
  dot?: string;
  count?: number | string;
  badge?: number;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <Button
      variant="ghost"
      type="button"
      onClick={onClick}
      className={`flex items-center gap-1.5 rounded-md px-2 py-1 text-xs transition h-auto justify-start ${
        active ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted'
      }`}
    >
      {dot ? <span className={`inline-block h-2 w-2 rounded-full ${dot}`} /> : Icon ? <Icon /> : null}
      <span>{label}</span>
      {count !== undefined && count !== '' && (
        <span className="tabular-nums text-[10px] text-muted-foreground">{count}</span>
      )}
      {badge !== undefined && badge > 0 && (
        <span className="rounded-full bg-red-500/15 px-1 text-[10px] font-medium tabular-nums text-red-600 dark:text-red-400">{badge}</span>
      )}
    </Button>
  );
}

function TabBar({ active, onChange }: { active: Tab; onChange: (t: Tab) => void }) {
  const tabs: { id: Tab; label: string; icon: ComponentType<IconProps> }[] = [
    { id: 'prs', label: 'PRs', icon: UiGitPr },
    { id: 'projects', label: 'Projects', icon: UiFolderGit },
    { id: 'todos', label: 'Todos', icon: UiCheck },
    { id: 'tasks', label: 'Tasks', icon: UiListChecks },
    { id: 'prompts', label: 'Prompts', icon: UiRobotAi },
    { id: 'activity', label: 'Activity', icon: UiActivity },
  ];
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
