import type { Project, ProcStatus } from '../types';
import { ProcControl } from './ProcControl';

interface Props {
  /** Projects whose repos have no PRs in the current list (pinned above it). */
  projects: Project[];
  /** Status keyed by project name (and repo) from /api/proc/status. */
  procStatus: Record<string, ProcStatus>;
  onChanged: () => void;
  onEdit: (project: Project) => void;
}

// ProjectsBar pins projects that aren't represented by a repo group in the PR
// list (e.g. a local directory with no open PRs) so their Procfile controls are
// still reachable.
export function ProjectsBar({ projects, procStatus, onChanged, onEdit }: Props) {
  if (projects.length === 0) return null;
  return (
    <div className="border-b border-gray-200">
      <div className="px-3 py-1 bg-gray-100 text-[11px] font-semibold text-gray-500 uppercase tracking-wide">
        Projects
      </div>
      {projects.map(p => {
        const key = p.repos[0] || p.name;
        return (
          <div key={p.name} className="pl-6 pr-3 py-1.5 flex items-center gap-2 hover:bg-gray-50">
            <iconify-icon icon="codicon:folder" className="text-gray-400 shrink-0" />
            <span className="text-sm font-medium text-gray-700 truncate flex-1" title={p.dir}>{p.name}</span>
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
    </div>
  );
}
