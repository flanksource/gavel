import type { Project, ProcProcess, ProcStatus } from '../types';
import { aggregateDotClass, crashedSummary } from '../utils';
import { ProcessPortLink } from './ProcessTable';
import { Spinner } from '../icons/Spinner';

interface Props {
  repo: string;
  project?: Project;
  status?: ProcStatus;
  onChanged: () => void;
}

// ProcControl is the compact inline indicator shown next to a repo/project: a
// health dot plus listening ports. Editing the workspace lives in the row's
// settings gear; starting, stopping, restarting, viewing logs and resource
// metrics all live in the top-right ProcessManager dropdown.
export function ProcControl({ project, status }: Props) {
  // Repos without a configured project (or without a Procfile) show nothing;
  // the header "+ Add dir" button is the single entry point for adding one.
  if (!project || !status?.hasProcfile) return null;
  const procs = status.processes ?? [];
  // Show the running/total count only when process states differ ("mixed").
  const mixed = new Set(procs.map(p => p.status)).size > 1;
  return (
    <span className="inline-flex items-center gap-0.5 shrink-0" onClick={(e) => e.stopPropagation()}>
      <ProcStatusIcon status={status} />
      {mixed && !transitioning(procs) && <span className="text-[10px] tabular-nums text-muted-foreground mr-0.5">{procs.filter(p => p.status === 'running').length}/{procs.length}</span>}
      <ProcPorts project={project} status={status} />
    </span>
  );
}

// A process is mid-(re)start: "restarting" (stopping) then "starting" (booting,
// waiting for its port to come up).
function transitioning(procs: ProcProcess[]): boolean {
  return procs.some(p => p.status === 'restarting' || p.status === 'starting');
}

// ProcStatusIcon is a Procfile project's aggregate process health: a spinner
// with its label while a process (re)starts, otherwise a coloured dot whose
// title carries the running count, crashes and ports.
export function ProcStatusIcon({ status, compact }: { status?: ProcStatus; compact?: boolean }) {
  if (!status?.hasProcfile) return null;
  const procs = status.processes ?? [];
  const running = procs.filter(p => p.status === 'running').length;
  const ports = listeningPorts(status);
  const crashed = crashedSummary(procs);
  if (transitioning(procs)) {
    const label = `${procs.some(p => p.status === 'restarting') ? 'restarting' : 'starting'}…`;
    return (
      <span className="inline-flex items-center gap-0.5 mr-0.5" title={`${running}/${procs.length} running · ${label}`}>
        <Spinner className="text-yellow-500 text-xs" />
        {!compact && <span className="text-[10px] text-muted-foreground">{label}</span>}
      </span>
    );
  }
  return (
    <span
      role="img"
      aria-label={`${running}/${procs.length} processes running${crashed ? `, ${crashed}` : ''}`}
      className={`inline-block w-2 h-2 rounded-full ${aggregateDotClass(procs)}`}
      title={`${running}/${procs.length} running${crashed ? ` · ${crashed}` : ''}${ports.length ? ` · ${ports.map(p => `:${p}`).join(' ')}` : ''}`}
    />
  );
}

// ProcPorts links every distinct port the project's processes listen on.
export function ProcPorts({ project, status }: { project: Project; status?: ProcStatus }) {
  if (!status?.hasProcfile) return null;
  return <>{listeningPorts(status).map(port => <ProcessPortLink key={port} project={project.name} port={port} />)}</>;
}

function listeningPorts(status: ProcStatus): number[] {
  return Array.from(new Set((status.processes ?? []).flatMap(p => p.ports ?? []))).sort((a, b) => a - b);
}
