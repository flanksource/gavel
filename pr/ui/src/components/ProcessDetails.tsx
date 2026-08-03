import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Button, Modal } from "@flanksource/clicky-ui/components";
import { AnsiHtml } from "@flanksource/clicky-ui/data";
import { UiFullscreen } from "@flanksource/clicky-ui/icons";
import type { ProcNode, ProcProcess, ProcStatus } from "../types";
import { fetchJSON, fetchText, queryKeys } from "../query";
import { useDocumentVisible } from "../useDocumentVisible";
import { humanizeBytes } from "../utils";
import { useNow } from "../useNow";

const PROC_POLL_INTERVAL_MS = 3_000;

function cpuLabel(process: { cpuPercent?: number }): string {
  return process.cpuPercent && process.cpuPercent > 0 ? `${process.cpuPercent.toFixed(0)}%` : "—";
}

function uptimeLabel(process: { started?: string; status: string }): string {
  if (!process.started || !["running", "starting", "restarting"].includes(process.status)) return "—";
  const started = new Date(process.started).getTime();
  if (!Number.isFinite(started)) return "—";
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

function Uptime({ process }: { process: { started?: string; status: string } }) {
  useNow();
  return <>{uptimeLabel(process)}</>;
}

export function filesLabel(process: { openFiles?: number }): string {
  if (process.openFiles === undefined || process.openFiles < 0) return "—";
  return String(process.openFiles);
}

function ProcLogPreview({ project, name }: { project: string; name: string }) {
  const logs = useProcLogs({ project, name, lines: 5, poll: true });
  const text = logs.isPending
    ? "loading…"
    : logs.error instanceof Error
      ? logs.error.message
      : logs.data?.trimEnd() || "(no output)";

  return <AnsiHtml as="pre" text={text} className="text-[10px] leading-snug bg-black text-gray-100 rounded p-2 overflow-x-auto whitespace-pre-wrap max-h-32" />;
}

function useProcLogs({ project, name, lines, poll }: { project: string; name: string; lines: number; poll: boolean }) {
  const visible = useDocumentVisible();
  return useQuery({
    queryKey: queryKeys.processLogs(project, name, lines),
    queryFn: ({ signal }) => fetchText({
      url: `/api/proc/logs?project=${encodeURIComponent(project)}&name=${encodeURIComponent(name)}&lines=${lines}`,
      signal,
      context: `Failed to load logs for ${project}/${name}`,
    }),
    enabled: visible,
    staleTime: poll ? PROC_POLL_INTERVAL_MS : Infinity,
    refetchInterval: visible && poll ? PROC_POLL_INTERVAL_MS : false,
    refetchIntervalInBackground: false,
  });
}

function useProcTree(project: string, name: string) {
  const visible = useDocumentVisible();
  const status = useQuery({
    queryKey: queryKeys.processStatus(project),
    queryFn: ({ signal }) => fetchJSON<ProcStatus>({
      url: `/api/proc/status?project=${encodeURIComponent(project)}`,
      signal,
      context: `Failed to load process status for ${project}`,
    }),
    enabled: visible,
    staleTime: PROC_POLL_INTERVAL_MS,
    refetchInterval: visible ? PROC_POLL_INTERVAL_MS : false,
    refetchIntervalInBackground: false,
    select: (snapshot) => (snapshot.processes ?? []).find((process) => process.name === name)?.tree ?? [],
  });
  return { nodes: status.data ?? [], error: status.error };
}

function flattenTree(nodes: ProcNode[]): { node: ProcNode; depth: number }[] {
  const pids = new Set(nodes.map((node) => node.pid));
  const children = new Map<number, ProcNode[]>();
  for (const node of nodes) children.set(node.ppid, [...(children.get(node.ppid) ?? []), node]);
  const byCpu = (left: ProcNode, right: ProcNode) => (right.cpuPercent ?? 0) - (left.cpuPercent ?? 0);
  const rows: { node: ProcNode; depth: number }[] = [];
  const seen = new Set<number>();
  const visit = (node: ProcNode, depth: number) => {
    if (seen.has(node.pid)) return;
    seen.add(node.pid);
    rows.push({ node, depth });
    for (const child of (children.get(node.pid) ?? []).sort(byCpu)) visit(child, depth + 1);
  };
  for (const root of nodes.filter((node) => !pids.has(node.ppid)).sort(byCpu)) visit(root, 0);
  for (const node of nodes) if (!seen.has(node.pid)) visit(node, 0);
  return rows;
}

function ProcTree({ nodes }: { nodes: ProcNode[] }) {
  const rows = useMemo(() => flattenTree(nodes), [nodes]);
  return (
    <table className="w-full text-[10px] tabular-nums">
      <thead>
        <tr className="text-left uppercase tracking-wide text-gray-400">
          <th className="pb-1 font-medium">Process</th>
          <th className="px-2 pb-1 text-right font-medium">CPU</th>
          <th className="px-2 pb-1 text-right font-medium">RSS</th>
          <th className="px-2 pb-1 text-right font-medium">VMS</th>
          <th className="px-2 pb-1 text-right font-medium">Files</th>
        </tr>
      </thead>
      <tbody>
        {rows.map(({ node, depth }) => (
          <tr key={node.pid} className="text-gray-600">
            <td className="py-0.5 pr-2 truncate" style={{ paddingLeft: `${depth * 14}px` }}>
              {depth > 0 && <span className="text-gray-300">└ </span>}
              <span className="text-gray-700">{node.command || "?"}</span>
              <span className="text-gray-400 ml-1">{node.pid}</span>
              {node.root && <span className="ml-1 text-primary">root</span>}
              {node.status && <span className="ml-1 text-gray-400">{node.status}</span>}
            </td>
            <td className="px-2 text-right w-12">{cpuLabel(node)}</td>
            <td className="px-2 text-right w-16">{humanizeBytes(node.memoryRss)}</td>
            <td className="px-2 text-right w-16">{humanizeBytes(node.memoryVms)}</td>
            <td className="px-2 text-right w-10">{filesLabel(node)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function ProcLogsDialog({ project, name, onClose }: { project: string; name: string; onClose: () => void }) {
  const logs = useProcLogs({ project, name, lines: 500, poll: false });
  const text = logs.isPending
    ? "loading…"
    : logs.error instanceof Error
      ? logs.error.message
      : logs.data?.trimEnd() || "(no output)";

  return (
    <Modal open onClose={onClose} title={`${project} · ${name} · logs`} size="xl">
      <AnsiHtml as="pre" text={text} className="text-xs leading-snug bg-black text-gray-100 rounded p-3 overflow-auto max-h-[70vh] whitespace-pre-wrap" />
    </Modal>
  );
}

export function ProcExpanded({ project, proc }: { project: string; proc: ProcProcess }) {
  const [logsOpen, setLogsOpen] = useState(false);
  const tree = useProcTree(project, proc.name);
  return (
    <div className="space-y-2 py-1">
      {tree.error instanceof Error && <div role="alert" className="text-[10px] text-red-500">{tree.error.message}</div>}
      {tree.nodes.length > 0 && (
        <div>
          <div className="mb-0.5 flex items-center justify-between gap-2">
            <div className="text-[10px] uppercase tracking-wide text-gray-400">Process tree</div>
            <div className="text-[10px] tabular-nums text-gray-400">
              up <Uptime process={proc} /> · pid {proc.pid || "—"}
              {proc.taskRunId && <a href={`/tasks/${encodeURIComponent(proc.taskRunId)}`} className="ml-2 text-primary hover:underline">Task details</a>}
            </div>
          </div>
          <ProcTree nodes={tree.nodes} />
        </div>
      )}
      <div>
        <div className="flex items-center justify-between mb-0.5">
          <span className="text-[10px] uppercase tracking-wide text-gray-400">Logs</span>
          <Button variant="ghost" size="icon" title="Open logs in dialog" aria-label="Open logs in dialog" onClick={() => setLogsOpen(true)}>
            <UiFullscreen className="text-sm" />
          </Button>
        </div>
        <ProcLogPreview project={project} name={proc.name} />
      </div>
      {logsOpen && <ProcLogsDialog project={project} name={proc.name} onClose={() => setLogsOpen(false)} />}
    </div>
  );
}
