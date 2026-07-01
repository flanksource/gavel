import { Button } from '@flanksource/clicky-ui/components';
import type { Project, ProcStatus } from '../types';
import { ProcControl } from './ProcControl';
import { GavelIcon } from './GavelIcon';
import { TodoBadge } from './TodoBadge';
import { GitChangesBadge } from './GitChangesBadge';

interface Props {
  /** Every configured project from projects.json (not filtered by PR state). */
  projects: Project[];
  /** Status keyed by project name (and repo) from /api/proc/status. */
  procStatus: Record<string, ProcStatus>;
  onChanged: () => void;
  onEdit: (project: Project) => void;
  /** Opens the add-workspace-directory dialog. */
  onAdd: () => void;
}

// ProjectsBar lists every configured workspace straight from projects.json,
// independent of the GitHub PR fetch, so a project never vanishes when one of
// its repos gains an open PR. The section is always shown so its "Add directory"
// action stays available even when no projects are configured yet.
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
            <TodoBadge counts={p.todoCounts} />
            <GitChangesBadge count={procStatus[p.name]?.gitChanges} />
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
      <Button
        variant="ghost"
        type="button"
        onClick={onAdd}
        title="Add a local workspace directory"
        className="w-full pl-6 pr-3 py-1.5 h-auto flex items-center justify-start gap-2 text-xs text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
      >
        <GavelIcon name="codicon:add" className="shrink-0" />
        Add directory
      </Button>
    </div>
  );
}
