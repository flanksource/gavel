import { useEffect, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import type { MouseEvent as ReactMouseEvent } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import type { Org, Project, SearchConfig } from '../types';
import { UiCheck, UiChevronDown, UiChevronRight, UiEye, UiEyeClosed, UiFolder, UiGlobe, UiOrganization, UiUser } from '@flanksource/clicky-ui/icons';
import { orgsQuery } from './oneShotQueries';

interface Props {
  config: SearchConfig;
  projects: Project[];
  onChange: (partial: Partial<SearchConfig>) => void;
}

// OrgChooser is the dashboard-wide scope switcher. Projects bind their repos
// to PR results and their workspace to local views such as Todos; GitHub scopes
// retain the existing org-wide and personal browsing modes.
//
// Orgs are fetched lazily on first open via /api/orgs?include-ignored=1 —
// the server caches the underlying list for 5 minutes, so the dropdown
// stays snappy. Requesting the un-filtered list here lets the chooser
// render an inline "Manage hidden" section so users can unhide without
// juggling a second endpoint.

export function OrgChooser({ config, projects, onChange }: Props) {
  const [open, setOpen] = useState(false);
  const [showHidden, setShowHidden] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const orgsResult = useQuery({ ...orgsQuery(), enabled: open });
  const orgs: Org[] = orgsResult.data ?? [];
  const loading = orgsResult.isPending && open;
  const err = orgsResult.error instanceof Error ? orgsResult.error.message : null;

  // Close on outside click — the dropdown is anchored to the button so a
  // stray click should dismiss rather than trap the user.
  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', onDown);
    return () => document.removeEventListener('mousedown', onDown);
  }, [open]);

  const ignoredOrgs = config.ignoredOrgs || [];
  const ignoredSet = new Set(ignoredOrgs);
  const visibleOrgs = orgs.filter(o => !ignoredSet.has(o.login));
  const hiddenOrgs = orgs.filter(o => ignoredSet.has(o.login));

  const activeProject = config.project || '';
  const projectRepos = Array.from(new Set(projects.flatMap(project => project.repos)));
  const allProjectsSelected = projects.length > 0
    && !activeProject
    && !config.all
    && config.repos.length === projectRepos.length
    && projectRepos.every(repo => config.repos.includes(repo));
  const activeOrg = !activeProject && config.all ? (config.org || '') : '';
  // Five display modes: a project, all projects, a specific org, org-wide with
  // no explicit org, or personal scope.
  let label: string;
  if (activeProject) {
    label = activeProject;
  } else if (allProjectsSelected) {
    label = 'All projects';
  } else if (!config.all) {
    label = '@me';
  } else if (activeOrg) {
    label = activeOrg;
  } else {
    label = 'All orgs';
  }
  const activeOrgMeta = activeOrg ? orgs.find(o => o.login === activeOrg) : undefined;

  function chooseProject(project: Project) {
    onChange({ project: project.name, repos: project.repos, org: '', all: false });
    setOpen(false);
  }

  function chooseAllProjects() {
    onChange({ project: '', repos: projectRepos, org: '', all: false });
    setOpen(false);
  }

  function chooseOrg(login: string) {
    // Selecting an org implies org-wide mode. Clear the repo list because
    // old repo filters are almost always from a different org.
    onChange({ project: '', org: login, all: true, repos: [] });
    setOpen(false);
  }

  function chooseMe() {
    onChange({ project: '', org: '', all: false, repos: [] });
    setOpen(false);
  }

  // Org-wide browsing without pinning a specific org — daemon's
  // ResolveDefaultOrg picks one (skipping ignored orgs) each search.
  function chooseAllOrgs() {
    onChange({ project: '', org: '', all: true, repos: [] });
    setOpen(false);
  }

  function hideOrg(login: string, e: ReactMouseEvent) {
    e.stopPropagation(); // don't also select it
    const next = Array.from(new Set([...ignoredOrgs, login]));
    const patch: Partial<SearchConfig> = { ignoredOrgs: next };
    // If the user just hid the currently-selected org, drop back to @me so the
    // next poll doesn't keep fetching PRs from a hidden org.
    if (config.org === login) {
      patch.org = '';
      patch.all = false;
      patch.repos = [];
    }
    onChange(patch);
  }

  function unhideOrg(login: string, e: ReactMouseEvent) {
    e.stopPropagation();
    onChange({ ignoredOrgs: ignoredOrgs.filter(o => o !== login) });
  }

  const HiddenChevron = showHidden ? UiChevronDown : UiChevronRight;

  return (
    <div className="relative min-w-0" ref={rootRef}>
      <Button
        variant="ghost"
        className="inline-flex max-w-full items-center justify-start gap-1.5 text-xs h-auto px-2 py-1 rounded border border-border text-muted-foreground hover:bg-muted transition-colors"
        onClick={() => setOpen(!open)}
        title="Switch project / GitHub scope"
      >
        {activeProject || allProjectsSelected ? (
          <UiFolder className="shrink-0 text-sm" />
        ) : activeOrgMeta?.avatarUrl ? (
          <img src={activeOrgMeta.avatarUrl} alt={activeOrg} className="w-4 h-4 rounded-sm" />
        ) : (
          <UiOrganization className="shrink-0 text-sm" />
        )}
        <span className="truncate font-medium">{label}</span>
        <UiChevronDown className="shrink-0 text-[10px]" />
      </Button>

      {open && (
        <div className="absolute top-full right-0 mt-1 max-h-[min(32rem,calc(100vh-4rem))] w-72 overflow-y-auto bg-popover rounded-lg shadow-lg border border-border z-50 py-1 text-sm">
          {projects.length > 0 && (
            <>
              <div className="px-3 pb-1 pt-1.5 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
                Projects
              </div>
              <Button
                variant="ghost"
                className={`w-full flex items-center justify-start gap-2 h-auto px-3 py-1.5 text-left transition-colors ${
                  allProjectsSelected ? 'bg-primary/10 text-primary' : 'hover:bg-muted text-foreground'
                }`}
                aria-label="Filter by all projects"
                onClick={chooseAllProjects}
              >
                <UiFolder className="shrink-0 text-base" />
                <span className="min-w-0 flex-1 truncate">All projects</span>
                {allProjectsSelected && <UiCheck className="text-xs" />}
              </Button>
              {projects.map(project => {
                const selected = activeProject === project.name;
                return (
                  <Button
                    key={project.name}
                    variant="ghost"
                    className={`w-full flex items-center justify-start gap-2 h-auto px-3 py-1.5 text-left transition-colors ${
                      selected ? 'bg-primary/10 text-primary' : 'hover:bg-muted text-foreground'
                    }`}
                    aria-label={`Filter by ${project.name} project`}
                    onClick={() => chooseProject(project)}
                  >
                    <UiFolder className="shrink-0 text-base" />
                    <span className="min-w-0 flex-1 truncate">{project.name}</span>
                    {selected && <UiCheck className="text-xs" />}
                  </Button>
                );
              })}
              <div className="border-t border-border my-1" />
            </>
          )}

          <div className="px-3 pb-1 pt-1.5 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
            GitHub scope
          </div>
          <Button
            variant="ghost"
            className={`w-full flex items-center justify-start gap-2 h-auto px-3 py-1.5 text-left transition-colors ${
              !activeProject && !allProjectsSelected && !config.all
                ? 'bg-primary/10 text-primary'
                : 'hover:bg-muted text-foreground'
            }`}
            onClick={chooseMe}
          >
            <UiUser className="text-base" />
            <span className="flex-1">@me (my PRs)</span>
            {!activeProject && !allProjectsSelected && !config.all && <UiCheck className="text-xs" />}
          </Button>

          <Button
            variant="ghost"
            className={`w-full flex items-center justify-start gap-2 h-auto px-3 py-1.5 text-left transition-colors ${
              config.all && !activeOrg ? 'bg-primary/10 text-primary' : 'hover:bg-muted text-foreground'
            }`}
            onClick={chooseAllOrgs}
          >
            <UiGlobe className="text-base" />
            <span className="flex-1">All orgs (default)</span>
            {config.all && !activeOrg && <UiCheck className="text-xs" />}
          </Button>

          <div className="border-t border-border my-1" />

          {loading && <div className="px-3 py-2 text-xs text-muted-foreground">Loading orgs…</div>}
          {err && <div className="px-3 py-2 text-xs text-red-500">{err}</div>}
          {!loading && !err && visibleOrgs.length === 0 && hiddenOrgs.length === 0 && (
            <div className="px-3 py-2 text-xs text-muted-foreground">No orgs — token has no org memberships</div>
          )}
          {visibleOrgs.map(o => {
            const selected = config.all && config.org === o.login;
            return (
              <div
                key={o.login}
                className={`group flex items-center gap-2 px-3 py-1.5 transition-colors ${
                  selected ? 'bg-primary/10 text-primary' : 'hover:bg-muted text-foreground'
                }`}
              >
                <Button
                  variant="ghost"
                  className="flex-1 flex items-center justify-start gap-2 h-auto p-0 text-left"
                  aria-label={`Select ${o.login} organization`}
                  onClick={() => chooseOrg(o.login)}
                >
                  {o.avatarUrl
                    ? <img src={o.avatarUrl} alt={o.login} className="w-4 h-4 rounded-sm shrink-0" />
                    : <UiOrganization className="text-base" />}
                  <span className="flex-1 truncate">{o.login}</span>
                  {selected && <UiCheck className="text-xs" />}
                </Button>
                <Button
                  variant="ghost"
                  className="h-auto p-0 opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-red-500 transition-opacity"
                  title={`Hide ${o.login} from this list`}
                  onClick={(e: ReactMouseEvent) => hideOrg(o.login, e)}
                >
                  <UiEyeClosed className="text-xs" />
                </Button>
              </div>
            );
          })}

          {hiddenOrgs.length > 0 && (
            <>
              <div className="border-t border-border my-1" />
              <Button
                variant="ghost"
                className="w-full flex items-center justify-start gap-2 h-auto px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted"
                onClick={() => setShowHidden(v => !v)}
              >
                <HiddenChevron className="text-[10px]" />
                <span className="flex-1 text-left">Manage hidden ({hiddenOrgs.length})</span>
              </Button>
              {showHidden && hiddenOrgs.map(o => (
                <div
                  key={o.login}
                  className="group flex items-center gap-2 px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted"
                >
                  {o.avatarUrl
                    ? <img src={o.avatarUrl} alt={o.login} className="w-4 h-4 rounded-sm shrink-0 opacity-60" />
                    : <UiOrganization className="text-base" />}
                  <span className="flex-1 truncate">{o.login}</span>
                  <Button
                    variant="ghost"
                    className="h-auto p-0 text-muted-foreground hover:text-primary transition-colors"
                    title={`Unhide ${o.login}`}
                    onClick={(e: ReactMouseEvent) => unhideOrg(o.login, e)}
                  >
                    <UiEye className="text-xs" />
                  </Button>
                </div>
              ))}
            </>
          )}
        </div>
      )}
    </div>
  );
}
