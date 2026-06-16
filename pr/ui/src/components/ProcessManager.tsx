import { useMemo } from 'react';
import { DropdownMenu } from '@flanksource/clicky-ui/components';
import type { Project, ProcStatus } from '../types';
import { flattenProcesses, aggregateDotClass } from '../utils';
import { WorkspaceGroup } from './ProcessTable';

interface Props {
  projects: Project[];
  procStatus: Record<string, ProcStatus>;
  onProcChanged: () => void;
}

// ProcessManager is the top-right task-manager dropdown: a profile-aware view of
// every supervised workspace. Each workspace shows its processes (with live CPU /
// memory / open-file metrics and a per-process log preview) plus a profile
// selector and start/restart/stop-all controls.
export function ProcessManager({ projects, procStatus, onProcChanged }: Props) {
  // Workspaces are the configured projects that actually have a Procfile.
  const workspaces = useMemo(
    () => projects
      .map(p => ({ project: p, status: procStatus[p.name] }))
      .filter((w): w is { project: Project; status: ProcStatus } => !!w.status?.hasProcfile),
    [projects, procStatus],
  );

  // Aggregate counts (across every process) drive the trigger badge.
  const procs = useMemo(() => flattenProcesses(projects, procStatus), [projects, procStatus]);
  const { running, crashed } = useMemo(() => {
    let running = 0, crashed = 0;
    for (const { proc } of procs) {
      if (proc.status === 'running') running++;
      else if (proc.status === 'crashed') crashed++;
    }
    return { running, crashed };
  }, [procs]);

  // No workspaces with a Procfile → no button (keeps the header clean).
  if (workspaces.length === 0) return null;

  const dot = aggregateDotClass(procs.map(p => p.proc));

  const trigger = (
    <button
      className="flex items-center gap-1.5 rounded-md border border-gray-200 bg-white px-2 py-1 text-xs hover:bg-gray-50"
      title="Processes"
      aria-label="Processes"
    >
      <iconify-icon icon="codicon:server-process" className="text-gray-500 text-sm" />
      <span className={`inline-block w-2 h-2 rounded-full ${dot}`} />
      <span className="tabular-nums text-gray-700">{running}/{procs.length}</span>
      {crashed > 0 && <span className="tabular-nums text-red-600">⚠{crashed}</span>}
    </button>
  );

  return (
    <DropdownMenu trigger={trigger} align="right" menuLabel="Processes" menuClassName="w-[720px] max-w-[90vw]">
      {() => (
        <div className="p-2">
          <div className="px-1 pb-1 text-xs font-semibold text-gray-600">
            Processes <span className="text-gray-400 font-normal">· {running} running of {procs.length}</span>
          </div>
          <div className="max-h-[60vh] overflow-y-auto divide-y divide-gray-100">
            {workspaces.map(w => (
              <WorkspaceGroup key={w.project.name} project={w.project} status={w.status} onChanged={onProcChanged} />
            ))}
          </div>
        </div>
      )}
    </DropdownMenu>
  );
}
