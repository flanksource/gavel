import { useState } from 'react';
import { Button, Modal } from '@flanksource/clicky-ui/components';
import { LogsTable } from '@flanksource/clicky-ui/data';
import type { Project, ProcStatus, ProcProcess } from '../types';

interface Props {
  repo: string;
  project?: Project;
  status?: ProcStatus;
  onChanged: () => void;
  onEdit?: (project: Project) => void;
}

function dotColor(procs: ProcProcess[]): string {
  const total = procs.length;
  const running = procs.filter(p => p.status === 'running').length;
  const crashed = procs.filter(p => p.status === 'crashed').length;
  const transitioning = procs.some(p => p.status === 'starting' || p.status === 'restarting');
  if (crashed > 0) return 'bg-red-500';
  if (total > 0 && running === total) return 'bg-green-500';
  if (running > 0 || transitioning) return 'bg-yellow-400';
  return 'bg-gray-300';
}

// IconBtn is a clicky ghost icon button wrapping an iconify glyph, so we keep
// the codicon vocabulary while using clicky-ui's button chrome.
function IconBtn({ icon, title, onClick, disabled }: { icon: string; title: string; onClick: () => void; disabled?: boolean }) {
  return (
    <Button
      variant="ghost"
      size="icon"
      title={title}
      aria-label={title}
      disabled={disabled}
      onClick={(e) => { e.stopPropagation(); onClick(); }}
    >
      <iconify-icon icon={icon} className="text-sm" />
    </Button>
  );
}

export function ProcControl({ repo, project, status, onChanged, onEdit }: Props) {
  const [busy, setBusy] = useState(false);
  const [logsOpen, setLogsOpen] = useState(false);
  const [logs, setLogs] = useState('');

  // Repos without a configured project (or without a Procfile) show nothing;
  // the header "+ Add dir" button is the single entry point for adding one.
  if (!project || !status?.hasProcfile) return null;

  const procs = status.processes ?? [];
  const total = procs.length;
  const running = procs.filter(p => p.status === 'running').length;
  const allRunning = total > 0 && running === total;
  // Show the running/total count only when process states differ ("mixed").
  const mixed = new Set(procs.map(p => p.status)).size > 1;

  async function control(action: 'start' | 'stop' | 'restart') {
    if (!project) return;
    setBusy(true);
    try {
      await fetch(`/api/proc/${action}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ project: project.name }),
      });
    } catch { /* surfaced on next status poll */ }
    setBusy(false);
    onChanged();
  }

  async function openLogs() {
    if (!project) return;
    try {
      const r = await fetch(`/api/proc/logs?project=${encodeURIComponent(project.name)}&lines=500`);
      setLogs(await r.text());
    } catch {
      setLogs('failed to load logs');
    }
    setLogsOpen(true);
  }

  return (
    <span className="inline-flex items-center gap-0.5 shrink-0" onClick={(e) => e.stopPropagation()}>
      <span className={`inline-block w-2 h-2 rounded-full ${dotColor(procs)}`} title={`${running}/${total} running`} />
      {mixed && <span className="text-[10px] tabular-nums text-muted-foreground mr-0.5">{running}/{total}</span>}

      {!allRunning && <IconBtn icon="codicon:play" title="Start" disabled={busy} onClick={() => control('start')} />}
      {running > 0 && <IconBtn icon="codicon:debug-restart" title="Restart" disabled={busy} onClick={() => control('restart')} />}
      {running > 0 && <IconBtn icon="codicon:debug-stop" title="Stop" disabled={busy} onClick={() => control('stop')} />}
      <IconBtn icon="codicon:output" title="Logs" onClick={openLogs} />
      {onEdit && <IconBtn icon="codicon:gear" title="Edit directory" onClick={() => onEdit(project)} />}

      {logsOpen && (
        <Modal open onClose={() => setLogsOpen(false)} title={`${project.name} · logs`} size="xl">
          <LogsTable logs={logs} showRawDetails />
        </Modal>
      )}
    </span>
  );
}
