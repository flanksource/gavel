import { useMemo } from 'react';
import type { PRItem, PRDetail, SearchConfig, PRSyncStatus, GavelResultsSummary, Project } from '../../types';
import { Button } from '@flanksource/clicky-ui/components';
import { ChatButton } from '@flanksource/clicky-ui/ai';
import { UiArrowLeft, UiCheck, UiGitPr, UiTerminal } from '@flanksource/clicky-ui/icons';
import { PRList } from '../PRList';
import { PRDetailPanel } from '../PRDetail';
import { MenubarTodos } from '../MenubarTodos';
import { OrgChooser } from '../OrgChooser';
import { ProjectsPlaceholder } from '../ProjectsPlaceholder';
import { WorkspaceGroup } from '../ProcessTable';
import { SessionsSidebar, type SessionsQueryView } from '../sessions/SessionsView';
import { MobileSessionScreen } from '../sessions/MobileSessionScreen';
import { aggregateDotClass, flattenProcesses, prKey, emptyProcStatus } from '../../utils';
import { useProcStatus } from '../../procStatusQuery';
import type { MobileTab } from '../../mobileTabs';
import { CopyLinkButton } from './CopyLinkButton';
import { MenubarTab } from './MenubarTab';
import {
  menubarPointerEnterMessage,
  menubarPointerLeaveMessage,
  postMenubarMessage,
  useMenubarExternalLinks,
} from './menubarBridge';

export function MenubarView({
  prs,
  selected,
  detail,
  detailLoading,
  unread,
  config,
  projects,
  availableProjects,
  onSetProjectHidden,
  projectsLoaded,
  projectError,
  projectsByRepo,
  syncStatus,
  gavelResults,
  onSelect,
  onBack,
  onScopeChange,
  onProcChanged,
  fetchedAt,
  error,
  tab: menubarTab,
  onTabChange,
  sessions,
  sessionId,
  onSessionChange,
  chatEnabled,
}: {
  prs: PRItem[];
  selected: PRItem | null;
  detail: PRDetail | null;
  detailLoading: boolean;
  unread: Record<string, boolean>;
  config: SearchConfig;
  projects: Project[];
  availableProjects: Project[];
  onSetProjectHidden: (name: string, hidden: boolean) => void;
  projectsLoaded: boolean;
  projectError?: string;
  projectsByRepo: Record<string, Project>;
  syncStatus: Record<string, PRSyncStatus>;
  gavelResults: Record<string, GavelResultsSummary>;
  onSelect: (pr: PRItem) => void;
  onBack: () => void;
  onScopeChange: (partial: Partial<SearchConfig>) => void;
  onProcChanged: () => void;
  fetchedAt: string;
  error?: string;
  tab: MobileTab;
  onTabChange: (tab: MobileTab) => void;
  sessions: SessionsQueryView;
  // The open session's id, or '' for the list; shown full-screen like a PR.
  sessionId: string;
  onSessionChange: (id: string) => void;
  // False in the native menubar webview, which has no assistant layer.
  chatEnabled: boolean;
}) {
  useMenubarExternalLinks();
  const procStatus = useProcStatus();

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

  if (menubarTab === 'sessions' && sessionId) {
    return <MobileSessionScreen query={sessions} sessionId={sessionId} onBack={() => onSessionChange('')} />;
  }

  if (selected) {
    return (
      <div
        className="h-dvh overflow-hidden bg-background text-foreground"
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
        <div className="h-[calc(100dvh-2.75rem)] overflow-hidden">
          <PRDetailPanel pr={selected} detail={detail} loading={detailLoading} projects={projects} onActionDone={() => { if (selected) onSelect(selected); }} />
        </div>
      </div>
    );
  }

  return (
    <div
      className="flex h-dvh flex-col overflow-hidden bg-background text-foreground"
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
        <div className="flex min-w-0 items-center justify-end gap-2">
          <div className="min-w-0 max-w-40">
            <OrgChooser config={config} projects={availableProjects} onChange={onScopeChange} onSetProjectHidden={onSetProjectHidden} />
          </div>
          {projectError ? (
            <div role="alert" className="max-w-32 truncate text-[11px] text-destructive">{projectError}</div>
          ) : (
            <div className="shrink-0 text-[11px] text-muted-foreground tabular-nums">{error || fetched}</div>
          )}
          {chatEnabled && <ChatButton label="Open Gavel assistant" />}
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-1 overflow-x-auto border-b border-border px-2 py-1">
        <MenubarTab
          label="Processes"
          dot={dot}
          count={`${running}/${procs.length}`}
          active={menubarTab === 'processes'}
          onClick={() => onTabChange('processes')}
        />
        <MenubarTab
          label="PRs"
          icon={UiGitPr}
          count={prs.length}
          badge={failed > 0 ? failed : undefined}
          active={menubarTab === 'prs'}
          onClick={() => onTabChange('prs')}
        />
        <MenubarTab
          label="Todos"
          icon={UiCheck}
          count={openTodos}
          badge={failedTodos > 0 ? failedTodos : undefined}
          active={menubarTab === 'todos'}
          onClick={() => onTabChange('todos')}
        />
        <MenubarTab
          label="Sessions"
          icon={UiTerminal}
          active={menubarTab === 'sessions'}
          onClick={() => onTabChange('sessions')}
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
        ) : menubarTab === 'sessions' ? (
          <SessionsSidebar query={sessions} selectedId="" scopeProject={config.project ?? ''} onSelect={onSessionChange} />
        ) : (
          <PRList
            prs={prs}
            selected={selected}
            onSelect={onSelect}
            unread={unread}
            syncStatus={syncStatus}
            gavelResults={gavelResults}
            projectsByRepo={projectsByRepo}
            onProcChanged={onProcChanged}
          />
        )}
      </div>
    </div>
  );
}
