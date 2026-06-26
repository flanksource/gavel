import { useMemo } from 'react';
import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import type { Project, ProcStatus } from '../types';
import { flattenProcesses, aggregateDotClass, emptyProcStatus } from '../utils';
import { WorkspaceGroup } from './ProcessTable';
import { GavelIcon } from './GavelIcon';

interface Props {
  projects: Project[];
  procStatus: Record<string, ProcStatus>;
  onProcChanged: () => void;
}

// ProcessManager is the top-right task-manager dropdown: a profile-aware view of
// every configured workspace. Each workspace with a Procfile shows its processes
// (with live CPU / memory / open-file metrics and a per-process log preview) plus
// a profile selector and start/restart/stop-all controls; one without a Procfile
// shows as a compact "No Procfile" row.
export function ProcessManager({ projects, procStatus, onProcChanged }: Props) {
  // Every configured project is a workspace, listed straight from projects.json;
  // those without a Procfile render as a compact "No Procfile" row.
  const workspaces = useMemo(
    () => projects.map(p => ({ project: p, status: procStatus[p.name] ?? emptyProcStatus })),
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

  // No projects configured → no button (keeps the header clean).
  if (workspaces.length === 0) return null;

  const dot = aggregateDotClass(procs.map(p => p.proc));

  const trigger = (
    <Button
      variant="ghost"
      className="flex items-center justify-start gap-1.5 h-auto rounded-md border border-gray-200 bg-white px-2 py-1 text-xs hover:bg-gray-50"
      title="Processes"
      aria-label="Processes"
    >
      <GavelIcon name="codicon:server-process" className="text-gray-500 text-sm" />
      <span className={`inline-block w-2 h-2 rounded-full ${dot}`} />
      <span className="tabular-nums text-gray-700">{running}/{procs.length}</span>
      {crashed > 0 && (
        <span className="inline-flex items-center gap-0.5 tabular-nums text-red-600">
          <GavelIcon name="codicon:warning" />
          {crashed}
        </span>
      )}
    </Button>
  );

  return (
    <DropdownMenu trigger={trigger} align="right" menuLabel="Processes" menuClassName="w-[720px] max-w-[90vw]">
      {() => (
        <div className="p-2">
          <div className="flex items-center justify-between gap-2 px-1 pb-1 text-xs font-semibold text-gray-600">
            <span>
              Processes <span className="text-gray-400 font-normal">· {running} running of {procs.length}</span>
            </span>
            <a
              href="/processes"
              target="_blank"
              rel="noreferrer"
              title="Open processes full page"
              aria-label="Open processes full page"
              className="inline-flex h-6 w-6 items-center justify-center rounded text-gray-400 hover:bg-gray-100 hover:text-gray-700"
              onClick={e => e.stopPropagation()}
            >
              <GavelIcon name="codicon:link-external" className="text-sm" />
            </a>
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
