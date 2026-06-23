import { useState, useEffect, useMemo, type ComponentProps } from 'react';
import { Button, Modal } from '@flanksource/clicky-ui/components';
import { AnsiHtml, TimeseriesCoreBars } from '@flanksource/clicky-ui/data';
import { UiActivity, UiDatabase } from '@flanksource/clicky-ui/icons';
import type { FlatProc } from '../utils';
import { humanizeBytes, statusDotClass, aggregateDotClass, statusLabel } from '../utils';
import type { ProcNode, ProcProcess, Project, ProcStatus } from '../types';
import { GavelIcon } from './GavelIcon';
import { TodoBadge } from './TodoBadge';
import { useNow } from '../useNow';

// MetricIcon is the gauge's own icon prop type, derived from the component so it
// matches clicky-ui's icon typing (avoids a React 18/19 @types/react mismatch).
type MetricIcon = ComponentProps<typeof TimeseriesCoreBars>['icon'];

// MEMORY_UNIT renders RSS as cores-style bars at one bar per gigabyte; bytes are
// fed raw and TimeseriesCoreBars divides by perBar, so captions read "x / y GB".
const GIB = 1024 ** 3;
const MEMORY_UNIT = { perBar: GIB, label: 'GB', barLabel: 'GB' };

function cpuLabel(p: { cpuPercent?: number }): string {
  return p.cpuPercent && p.cpuPercent > 0 ? `${p.cpuPercent.toFixed(0)}%` : '—';
}

function uptimeLabel(p: { started?: string; status: string }): string {
  if (!p.started || !isActiveStatus(p.status)) return '—';
  const started = new Date(p.started).getTime();
  if (!Number.isFinite(started)) return '—';
  const seconds = Math.max(0, Math.floor((Date.now() - started) / 1000));
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  const remMinutes = minutes % 60;
  if (hours < 24) return remMinutes > 0 ? `${hours}h ${remMinutes}m` : `${hours}h`;
  const days = Math.floor(hours / 24);
  const remHours = hours % 24;
  return remHours > 0 ? `${days}d ${remHours}h` : `${days}d`;
}

// Uptime is a leaf that re-renders itself each second via the shared useNow()
// clock, so a ticking uptime doesn't reconcile its parent process row.
function Uptime({ proc }: { proc: { started?: string; status: string } }) {
  useNow();
  return <>{uptimeLabel(proc)}</>;
}

function isActiveStatus(status: string): boolean {
  return status === 'running' || status === 'starting' || status === 'restarting';
}

// filesLabel renders the open-file count; -1 means the platform can't report it.
function filesLabel(p: { openFiles?: number }): string {
  if (p.openFiles === undefined || p.openFiles < 0) return '—';
  return String(p.openFiles);
}

// Process gauges poll their recorded series from the backend; the cell only
// shows the latest point, so a short window keeps each payload tiny.
const PROC_REFRESH_MS = 2000;
const PROC_RANGE = '10m';

// runKey MUST match procRunKey in pr/ui/proc_metrics.go — the backend records
// each series under this id and the gauges below request it back.
function runKey(project: string, proc: ProcProcess): string {
  return `${project}/${proc.name}/${proc.started || 'not-started'}/${proc.pid || 0}`;
}

// metricId is the single URL-encoded path segment the clicky metrics handler
// (mounted at /api/proc) reads via its {id} wildcard.
function metricId(key: string, metric: string): string {
  return `/api/proc/metrics/${encodeURIComponent(`${key}/${metric}`)}`;
}

function firstPort(proc: ProcProcess): number | undefined {
  const ports = proc.ports ?? [];
  if (ports.length === 0) return undefined;
  return [...ports].sort((a, b) => a - b)[0];
}

// CpuBars renders one CPU reading — a single process or a workspace-summed
// aggregate — as unlabelled cell bars fed live from the recorded metric series
// (/api/proc/metrics). With no `max` the gauge sizes its bar count from the
// usage itself (one bar per core), so a busy multi-core process renders without
// clipping and the scale needs no client-side peak tracking. metricKey scopes
// the query cache; icon labels the reading in the workspace headers, where no
// column header names it (table rows are labelled by their column header).
function CpuBars({ metricKey, icon }: { metricKey: string; icon?: MetricIcon }) {
  return (
    <div className="flex justify-end">
      <TimeseriesCoreBars
        variant="cell"
        showLabel={false}
        {...(icon ? { icon } : {})}
        title="CPU"
        value={{ id: metricId(metricKey, 'cpu'), transform: v => v * 10 }}
        range={PROC_RANGE}
        refreshMs={PROC_REFRESH_MS}
      />
    </div>
  );
}

// MemoryBars mirrors CpuBars for RSS: live cell bars at one bar per gigabyte
// (MEMORY_UNIT), self-sizing from the reading, with an optional header icon.
function MemoryBars({ metricKey, icon }: { metricKey: string; icon?: MetricIcon }) {
  return (
    <div className="flex justify-end">
      <TimeseriesCoreBars
        variant="cell"
        showLabel={false}
        {...(icon ? { icon } : {})}
        title="Mem"
        value={{ id: metricId(metricKey, 'memory') }}
        unit={MEMORY_UNIT}
        range={PROC_RANGE}
        refreshMs={PROC_REFRESH_MS}
      />
    </div>
  );
}

function ProcessFavicon({ project, port }: { project: string; port?: number }) {
  const [faviconFailed, setFaviconFailed] = useState(false);
  useEffect(() => setFaviconFailed(false), [port]);
  if (!port || faviconFailed) return null;
  const favicon = `/api/proc/favicon?project=${encodeURIComponent(project)}&port=${port}`;
  return (
    <img
      src={favicon}
      alt=""
      className="h-4 w-4 shrink-0 rounded-sm"
      onError={() => setFaviconFailed(true)}
    />
  );
}

export function ProcessPortLink({ port }: { project: string; port: number }) {
  const href = `http://localhost:${port}`;
  return (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      title={`Open localhost:${port}`}
      className="inline-flex items-center gap-0.5 text-[10px] tabular-nums text-blue-500 hover:underline mr-1"
      onClick={e => e.stopPropagation()}
    >
      :{port}
    </a>
  );
}

async function control(project: string, name: string, action: 'start' | 'stop' | 'restart') {
  try {
    await fetch(`/api/proc/${action}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ project, names: [name] }),
    });
  } catch { /* surfaced on the next status poll */ }
}

// ProcLogPreview tails the last few lines of one process, refreshing on its own
// cadence while the row is expanded (it unmounts when the row collapses). Each
// fetch is abortable so an unmount or a superseding poll never leaks a pending
// request (which would otherwise pile up against the browser's connection cap).
function ProcLogPreview({ project, name }: { project: string; name: string }) {
  const [text, setText] = useState('loading…');
  useEffect(() => {
    let alive = true;
    let inflight: AbortController | null = null;
    const load = () => {
      inflight?.abort();
      inflight = new AbortController();
      fetch(`/api/proc/logs?project=${encodeURIComponent(project)}&name=${encodeURIComponent(name)}&lines=5`, { signal: inflight.signal })
        .then(r => r.text())
        .then(t => { if (alive) setText(t.trimEnd() || '(no output)'); })
        .catch((e) => { if (alive && e?.name !== 'AbortError') setText('failed to load logs'); });
    };
    load();
    const id = setInterval(() => { if (document.visibilityState === 'visible') load(); }, 3000);
    return () => { alive = false; inflight?.abort(); clearInterval(id); };
  }, [project, name]);

  return (
    <AnsiHtml
      as="pre"
      text={text}
      className="text-[10px] leading-snug bg-black text-gray-100 rounded p-2 overflow-x-auto whitespace-pre-wrap max-h-32"
    />
  );
}

// useProcTree fetches one process's live process-group breakdown while its row
// is expanded. The SSE status stream omits the per-process tree (its per-node
// cpu/mem churn would defeat the stream's change-detection), so the expanded
// view pulls the full single-project status on its own cadence — abortably and
// visibility-gated, mirroring ProcLogPreview — and the row unmounts the poll on
// collapse.
function useProcTree(project: string, name: string): ProcNode[] {
  const [tree, setTree] = useState<ProcNode[]>([]);
  useEffect(() => {
    let alive = true;
    let inflight: AbortController | null = null;
    const load = () => {
      inflight?.abort();
      inflight = new AbortController();
      fetch(`/api/proc/status?project=${encodeURIComponent(project)}`, { signal: inflight.signal })
        .then(r => r.json())
        .then((st: ProcStatus) => {
          if (!alive) return;
          setTree((st.processes ?? []).find(p => p.name === name)?.tree ?? []);
        })
        .catch(() => { /* keep the last tree on a failed poll */ });
    };
    load();
    const id = setInterval(() => { if (document.visibilityState === 'visible') load(); }, 3000);
    return () => { alive = false; inflight?.abort(); clearInterval(id); };
  }, [project, name]);
  return tree;
}

// flattenTree returns the group's processes in depth-first order with a depth
// for indentation. Roots are nodes whose parent isn't in the group (the group
// leader); a visited set guards against a malformed parent cycle.
function flattenTree(nodes: ProcNode[]): { node: ProcNode; depth: number }[] {
  const pids = new Set(nodes.map(n => n.pid));
  const children = new Map<number, ProcNode[]>();
  for (const n of nodes) {
    const arr = children.get(n.ppid) ?? [];
    arr.push(n);
    children.set(n.ppid, arr);
  }
  const byCpu = (a: ProcNode, b: ProcNode) => (b.cpuPercent ?? 0) - (a.cpuPercent ?? 0);
  const out: { node: ProcNode; depth: number }[] = [];
  const seen = new Set<number>();
  const visit = (n: ProcNode, depth: number) => {
    if (seen.has(n.pid)) return;
    seen.add(n.pid);
    out.push({ node: n, depth });
    for (const c of (children.get(n.pid) ?? []).sort(byCpu)) visit(c, depth + 1);
  };
  for (const r of nodes.filter(n => !pids.has(n.ppid)).sort(byCpu)) visit(r, 0);
  // Any node not reached (orphaned parent reference) still gets listed flat.
  for (const n of nodes) if (!seen.has(n.pid)) visit(n, 0);
  return out;
}

// ProcTree renders the per-process breakdown of a group as an indented table
// with each process's own CPU / memory / open-file metrics.
function ProcTree({ nodes }: { nodes: ProcNode[] }) {
  const rows = useMemo(() => flattenTree(nodes), [nodes]);
  return (
    <table className="w-full text-[10px] tabular-nums">
      <tbody>
        {rows.map(({ node, depth }) => (
          <tr key={node.pid} className="text-gray-600">
            <td className="py-0.5 pr-2 truncate" style={{ paddingLeft: `${depth * 14}px` }}>
              {depth > 0 && <span className="text-gray-300">└ </span>}
              <span className="text-gray-700">{node.command || '?'}</span>
              <span className="text-gray-400 ml-1">{node.pid}</span>
            </td>
            <td className="px-2 text-right w-12">{cpuLabel(node)}</td>
            <td className="px-2 text-right w-16">{humanizeBytes(node.memoryRss)}</td>
            <td className="px-2 text-right w-10">{filesLabel(node)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

// ProcLogsDialog opens the process's full logs in a large AnsiHtml terminal.
function ProcLogsDialog({ project, name, onClose }: { project: string; name: string; onClose: () => void }) {
  const [text, setText] = useState('loading…');
  useEffect(() => {
    let alive = true;
    const ac = new AbortController();
    fetch(`/api/proc/logs?project=${encodeURIComponent(project)}&name=${encodeURIComponent(name)}&lines=500`, { signal: ac.signal })
      .then(r => r.text())
      .then(t => { if (alive) setText(t.trimEnd() || '(no output)'); })
      .catch(e => { if (alive && e?.name !== 'AbortError') setText('failed to load logs'); });
    return () => { alive = false; ac.abort(); };
  }, [project, name]);

  return (
    <Modal open onClose={onClose} title={`${project} · ${name} · logs`} size="xl">
      <AnsiHtml
        as="pre"
        text={text}
        className="text-xs leading-snug bg-black text-gray-100 rounded p-3 overflow-auto max-h-[70vh] whitespace-pre-wrap"
      />
    </Modal>
  );
}

// ProcExpanded is the body shown when a process row is expanded: its process
// tree with per-process metrics, plus a log preview that can pop out to a dialog.
function ProcExpanded({ project, proc }: { project: string; proc: ProcProcess }) {
  const [logsOpen, setLogsOpen] = useState(false);
  const tree = useProcTree(project, proc.name);
  return (
    <div className="space-y-2 py-1">
      {tree.length > 0 && (
        <div>
          <div className="mb-0.5 flex items-center justify-between gap-2">
            <div className="text-[10px] uppercase tracking-wide text-gray-400">Process tree</div>
            <div className="text-[10px] tabular-nums text-gray-400">
              up <Uptime proc={proc} /> · pid {proc.pid || '—'}
            </div>
          </div>
          <ProcTree nodes={tree} />
        </div>
      )}
      <div>
        <div className="flex items-center justify-between mb-0.5">
          <span className="text-[10px] uppercase tracking-wide text-gray-400">Logs</span>
          <IconBtn icon="codicon:screen-full" title="Open logs in dialog" onClick={() => setLogsOpen(true)} />
        </div>
        <ProcLogPreview project={project} name={proc.name} />
      </div>
      {logsOpen && <ProcLogsDialog project={project} name={proc.name} onClose={() => setLogsOpen(false)} />}
    </div>
  );
}

function ProcessRow({ row, onChanged, showWorkspace }: { row: FlatProc; onChanged: () => void; showWorkspace: boolean }) {
  const { project, proc } = row;
  const [busy, setBusy] = useState(false);
  const [open, setOpen] = useState(false);

  const transitioning = proc.status === 'starting' || proc.status === 'restarting';
  const active = proc.status === 'running' || transitioning;
  const ports = (proc.ports ?? []).slice().sort((a, b) => a - b);
  const faviconPort = firstPort(proc);

  async function act(action: 'start' | 'stop' | 'restart') {
    setBusy(true);
    await control(project.name, proc.name, action);
    setBusy(false);
    onChanged();
  }

  return (
    <>
      <tr className="border-b border-gray-100 hover:bg-gray-50">
        <td className="py-1 pl-1 pr-2">
          <button className="flex items-center gap-1.5" onClick={() => setOpen(o => !o)} title="Toggle logs">
            <GavelIcon name={open ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="text-gray-400 text-xs" />
            <span className={`inline-block w-2 h-2 rounded-full ${statusDotClass(proc.status)}`} />
            <ProcessFavicon project={project.name} port={faviconPort} />
            <span className="font-medium truncate max-w-[180px]">{showWorkspace ? project.name : proc.name}</span>
          </button>
          {showWorkspace && <div className="text-[10px] text-gray-400 pl-5">{proc.name}</div>}
        </td>
        <td className={`px-2 ${proc.status === 'crashed' ? 'text-red-600' : 'text-gray-500'}`}>{statusLabel(proc)}</td>
        <td className="px-2 text-right tabular-nums"><CpuBars metricKey={runKey(project.name, proc)} /></td>
        <td className="px-2 text-right tabular-nums"><MemoryBars metricKey={runKey(project.name, proc)} /></td>
        <td className="px-2 text-right tabular-nums">{filesLabel(proc)}</td>
        <td className="px-2">
          {ports.map(port => <ProcessPortLink key={port} project={project.name} port={port} />)}
        </td>
        <td className="px-1 text-right whitespace-nowrap">
          {!active && <IconBtn icon="codicon:play" title="Start" disabled={busy} onClick={() => act('start')} />}
          {active && <IconBtn icon="codicon:debug-restart" title="Restart" disabled={busy} onClick={() => act('restart')} />}
          {active && <IconBtn icon="codicon:debug-stop" title="Stop" disabled={busy} onClick={() => act('stop')} />}
        </td>
      </tr>
      {open && (
        <tr className="bg-gray-50">
          <td colSpan={7} className="px-2 pb-2">
            <ProcExpanded project={project.name} proc={proc} />
          </td>
        </tr>
      )}
    </>
  );
}

function IconBtn({ icon, title, onClick, disabled }: { icon: string; title: string; onClick: () => void; disabled?: boolean }) {
  return (
    <Button variant="ghost" size="icon" title={title} aria-label={title} disabled={disabled}
      onClick={(e) => { e.stopPropagation(); onClick(); }}>
      <GavelIcon name={icon} className="text-sm" />
    </Button>
  );
}

export function ProcessTable({ procs, onChanged, showWorkspace = true }: { procs: FlatProc[]; onChanged: () => void; showWorkspace?: boolean }) {
  if (procs.length === 0) {
    return <div className="px-3 py-4 text-center text-xs text-gray-400">No processes</div>;
  }
  // Live CPU now lives in the gauges (their own metric series), so the listing
  // can't sort by it. Keep a stable task-manager-ish order: active processes
  // first, then alphabetical.
  const activeFirst = (s: string) => (isActiveStatus(s) ? 0 : 1);
  const sorted = [...procs].sort((a, b) =>
    activeFirst(a.proc.status) - activeFirst(b.proc.status) || a.proc.name.localeCompare(b.proc.name));
  return (
    <table className="w-full text-xs">
      <thead>
        <tr className="text-[10px] uppercase tracking-wide text-gray-400 border-b border-gray-200">
          <th className="py-1 pl-1 pr-2 text-left font-medium">{showWorkspace ? 'Workspace' : 'Process'}</th>
          <th className="px-2 text-left font-medium">Status</th>
          <th className="px-2 text-right font-medium">
            <span className="inline-flex items-center justify-end gap-1"><UiActivity size={12} aria-hidden /> CPU</span>
          </th>
          <th className="px-2 text-right font-medium">
            <span className="inline-flex items-center justify-end gap-1"><UiDatabase size={12} aria-hidden /> Mem</span>
          </th>
          <th className="px-2 text-right font-medium">Files</th>
          <th className="px-2 text-left font-medium">Ports</th>
          <th className="px-1 text-right font-medium" />
        </tr>
      </thead>
      <tbody>
        {sorted.map(row => (
          <ProcessRow key={`${row.project.name}/${row.proc.name}`} row={row} onChanged={onChanged} showWorkspace={showWorkspace} />
        ))}
      </tbody>
    </table>
  );
}

// WorkspaceGroup renders one project: a header with its profile selector and
// start/restart/stop-all controls, above its process rows. Profiles choose which
// processes auto-start; the selector is editable only while stopped (switching a
// running daemon's profile means stop → start).
export function WorkspaceGroup({ project, status, onChanged }: { project: Project; status: ProcStatus; onChanged: () => void }) {
  // A configured workspace without a Procfile has nothing to supervise: render a
  // compact, control-free row (a Start button here would always 400 server-side)
  // so the workspace still appears in every process listing.
  if (!status.hasProcfile) {
    return (
      <div className="py-1.5">
        <div className="flex items-center gap-2 px-1">
          <span className="inline-block w-2 h-2 rounded-full bg-gray-300" />
          <span className="text-sm font-medium truncate max-w-[200px]" title={project.dir}>{project.name}</span>
          <TodoBadge counts={project.todoCounts} />
          <div className="flex-1" />
          <span className="text-[10px] text-gray-400">No Procfile</span>
        </div>
      </div>
    );
  }

  const procs = status.processes ?? [];
  const profiles = status.profiles ?? [];
  const single = procs.length === 1;
  const running = procs.filter(p => p.status === 'running').length;
  const anyActive = procs.some(p => p.status === 'running' || p.status === 'starting' || p.status === 'restarting');

  const [busy, setBusy] = useState(false);
  const [open, setOpen] = useState(false);
  const [profile, setProfile] = useState(status.profile ?? '');
  // Track the active profile when the daemon (re)starts so the selector reflects it.
  useEffect(() => { setProfile(status.profile ?? ''); }, [status.profile]);

  async function control(action: 'start' | 'stop' | 'restart', withProfile: boolean) {
    setBusy(true);
    try {
      await fetch(`/api/proc/${action}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ project: project.name, ...(withProfile ? { profile } : {}) }),
      });
    } catch { /* surfaced on the next status poll */ }
    setBusy(false);
    onChanged();
  }

  // Workspace-level controls; shared by the single-process row and the multi
  // header. With one process, start/restart/stop-all act on that process.
  const controls = (
    <>
      {profiles.length > 0 && (
        <label className="flex items-center gap-1 text-[10px] text-gray-500" title="Profile to start">
          <GavelIcon name="codicon:layers" className="text-gray-400" />
          <select
            value={profile}
            disabled={busy || anyActive}
            onChange={e => setProfile(e.target.value)}
            className="text-[10px] border border-gray-200 rounded px-1 py-0.5 bg-white disabled:opacity-60"
          >
            <option value="">(default)</option>
            {profiles.map(pr => <option key={pr} value={pr}>{pr}</option>)}
          </select>
        </label>
      )}
      {!anyActive && <IconBtn icon="codicon:play" title="Start" disabled={busy} onClick={() => control('start', true)} />}
      {anyActive && <IconBtn icon="codicon:debug-restart" title={single ? 'Restart' : 'Restart all'} disabled={busy} onClick={() => control('restart', false)} />}
      {anyActive && <IconBtn icon="codicon:debug-stop" title={single ? 'Stop' : 'Stop all'} disabled={busy} onClick={() => control('stop', false)} />}
    </>
  );

  // A single-process workspace collapses to one compact row: the workspace and
  // its lone process share a line (status · cpu · mem · ports) with an expander
  // for the process tree + log preview, instead of a full table.
  if (single) {
    const proc = procs[0];
    const ports = (proc.ports ?? []).slice().sort((a, b) => a - b);
    const faviconPort = firstPort(proc);
    return (
      <div className="py-1.5">
        <div className="flex items-center gap-2 px-1">
          <button className="flex min-w-0 items-center gap-1.5" onClick={() => setOpen(o => !o)} title="Toggle logs">
            <GavelIcon name={open ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="text-gray-400 text-xs" />
            <span className={`inline-block w-2 h-2 rounded-full ${statusDotClass(proc.status)}`} />
            <ProcessFavicon project={project.name} port={faviconPort} />
            <span className="text-sm font-medium truncate max-w-[200px]" title={project.dir}>{project.name}</span>
          </button>
          <span className={`text-[10px] tabular-nums truncate ${proc.status === 'crashed' ? 'text-red-600' : 'text-gray-400'}`}>
            {statusLabel(proc)}
          </span>
          <TodoBadge counts={project.todoCounts} />
          <CpuBars metricKey={runKey(project.name, proc)} icon={UiActivity} />
          <MemoryBars metricKey={runKey(project.name, proc)} icon={UiDatabase} />
          {ports.map(port => <ProcessPortLink key={port} project={project.name} port={port} />)}
          <div className="flex-1" />
          {controls}
        </div>
        {open && (
          <div className="px-2 pb-1">
            <ProcExpanded project={project.name} proc={proc} />
          </div>
        )}
      </div>
    );
  }

  // The header rolls the workspace's processes up into one CPU/memory reading via
  // the per-project "__total__" metric series the sampler records, shown as
  // unlabelled cell gauges that the icons identify since there's no column header
  // to name them.
  return (
    <div className="py-1.5">
      <div className="flex items-center gap-2 px-1">
        <span className={`inline-block w-2 h-2 rounded-full ${aggregateDotClass(procs)}`} />
        <span className="text-sm font-medium truncate max-w-[200px]" title={project.dir}>{project.name}</span>
        <span className="text-[10px] tabular-nums text-gray-400">{running}/{procs.length}</span>
        <TodoBadge counts={project.todoCounts} />
        <div className="flex-1" />
        <CpuBars metricKey={`${project.name}/__total__`} icon={UiActivity} />
        <MemoryBars metricKey={`${project.name}/__total__`} icon={UiDatabase} />
        {controls}
      </div>
      <ProcessTable procs={procs.map(proc => ({ project, proc }))} onChanged={onChanged} showWorkspace={false} />
    </div>
  );
}
