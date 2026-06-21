import type { Project, ProcStatus } from '../types';
import { ProcControl } from './ProcControl';
import { GavelIcon } from './GavelIcon';

interface Props {
  /** Projects whose repos have no PRs in the current list (pinned above it). */
  projects: Project[];
  /** Status keyed by project name (and repo) from /api/proc/status. */
  procStatus: Record<string, ProcStatus>;
  onChanged: () => void;
  onEdit: (project: Project) => void;
  /** Opens the add-workspace-directory dialog. */
  onAdd: () => void;
}

// ProjectsBar pins projects that aren't represented by a repo group in the PR
// list (e.g. a local directory with no open PRs) so their Procfile controls are
// still reachable. The section is always shown so its "Add directory" action
// stays available even when there are no standalone projects yet.
export function ProjectsBar({ projects, procStatus, onChanged, onEdit, onAdd }: Props) {
  return (
    <div className="border-b border-border">
      <div className="px-3 py-1 bg-muted text-[11px] font-semibold text-muted-foreground uppercase tracking-wide">
        Projects
      </div>
      {projects.map(p => {
        const key = p.repos[0] || p.name;
        return (
          <div key={p.name} className="pl-6 pr-3 py-1.5 flex items-center gap-2 hover:bg-muted">
            <GavelIcon name="codicon:folder" className="text-muted-foreground shrink-0" />
            <span className="text-sm font-medium text-foreground truncate flex-1" title={p.dir}>{p.name}</span>
            <ProcControl
              repo={key}
              project={p}
              status={procStatus[p.name]}
              onChanged={onChanged}
              onEdit={onEdit}
            />
          </div>
        );
      })}
      <button
        type="button"
        onClick={onAdd}
        title="Add a local workspace directory"
        className="w-full pl-6 pr-3 py-1.5 flex items-center gap-2 text-xs text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
      >
        <GavelIcon name="codicon:add" className="shrink-0" />
        Add directory
      </button>
    </div>
  );
}
