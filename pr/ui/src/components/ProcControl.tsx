import { Button } from '@flanksource/clicky-ui/components';
import type { Project, ProcStatus } from '../types';
import { aggregateDotClass, crashedSummary } from '../utils';
import { ProcessPortLink } from './ProcessTable';
import { GavelIcon } from './GavelIcon';

interface Props {
  repo: string;
  project?: Project;
  status?: ProcStatus;
  onChanged: () => void;
  onEdit?: (project: Project) => void;
}

// IconBtn is a clicky ghost icon button wrapping an offline SVG glyph, so we
// keep the codicon vocabulary while using clicky-ui's button chrome.
function IconBtn({ icon, title, onClick }: { icon: string; title: string; onClick: () => void }) {
  return (
    <Button
      variant="ghost"
      size="icon"
      title={title}
      aria-label={title}
      onClick={(e) => { e.stopPropagation(); onClick(); }}
    >
      <GavelIcon name={icon} className="text-sm" />
    </Button>
  );
}

// ProcControl is the compact inline indicator shown next to a repo/project: a
// health dot plus listening ports. Starting, stopping, restarting, viewing
// logs and resource metrics all live in the top-right ProcessManager dropdown.
export function ProcControl({ project, status, onEdit }: Props) {
  // Repos without a configured project (or without a Procfile) show nothing;
  // the header "+ Add dir" button is the single entry point for adding one.
  if (!project || !status?.hasProcfile) return null;

  const procs = status.processes ?? [];
  const total = procs.length;
  const running = procs.filter(p => p.status === 'running').length;
  // A process is mid-(re)start: "restarting" (stopping) then "starting" (booting,
  // waiting for its port to come up). Drives the spinner + label below.
  const restarting = procs.some(p => p.status === 'restarting');
  const transitioning = restarting || procs.some(p => p.status === 'starting');
  // Show the running/total count only when process states differ ("mixed").
  const mixed = new Set(procs.map(p => p.status)).size > 1;
  // Distinct listening ports across all processes, surfaced as quick-open links.
  const ports = Array.from(new Set(procs.flatMap(p => p.ports ?? []))).sort((a, b) => a - b);
  const crashed = crashedSummary(procs);

  return (
    <span className="inline-flex items-center gap-0.5 shrink-0" onClick={(e) => e.stopPropagation()}>
      {transitioning ? (
        <span className="inline-flex items-center gap-0.5 mr-0.5" title={`${running}/${total} running`}>
          <GavelIcon name="svg-spinners:ring-resize" className="text-yellow-500 text-xs" />
          <span className="text-[10px] text-muted-foreground">{restarting ? 'restarting' : 'starting'}…</span>
        </span>
      ) : (
        <>
          <span
            className={`inline-block w-2 h-2 rounded-full ${aggregateDotClass(procs)}`}
            title={`${running}/${total} running${crashed ? ` · ${crashed}` : ''}${ports.length ? ` · ${ports.map(p => `:${p}`).join(' ')}` : ''}`}
          />
          {mixed && <span className="text-[10px] tabular-nums text-muted-foreground mr-0.5">{running}/{total}</span>}
        </>
      )}

      {ports.map(port => <ProcessPortLink key={port} project={project.name} port={port} />)}

      {onEdit && <IconBtn icon="codicon:gear" title="Edit directory" onClick={() => onEdit(project)} />}
    </span>
  );
}
