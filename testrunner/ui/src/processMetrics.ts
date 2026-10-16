import { ProcessNotFoundError, type FetchProcessMetrics, type ProcessMetrics } from '@flanksource/clicky-ui/data';
import type { ProcessSample, ProcessTraceNode } from './types';

const isExited = (node: ProcessTraceNode) => node.status === 'exited' || !!node.exitedAt;

/**
 * The live metrics of an attempt's process, read from the process_samples the
 * run's trace already holds rather than from a server poll. clicky's
 * FetchProcessMetrics asks by pid, so the pid is matched to its latest sample
 * (the attempt's `process_key` is `<pid>@<start>`, which a pid reused within
 * one run would make ambiguous). A pid whose process has exited rejects with
 * ProcessNotFoundError, which stops the caller's poll; a live pid with no
 * sample yet rejects with a plain error and the poll goes on. An unknown
 * cpuPercent stays absent: it is unknown, not zero.
 */
export function fetchProcessMetricsFromSamples({ samples, nodes }: { samples: ProcessSample[]; nodes: ProcessTraceNode[] }): FetchProcessMetrics {
  return async (pid: number): Promise<ProcessMetrics> => {
    const owners = nodes.filter(node => node.pid === pid);
    if (owners.length > 0 && owners.every(isExited)) throw new ProcessNotFoundError(pid);
    const latest = samples.filter(sample => sample.pid === pid).sort((a, b) => b.tick - a.tick)[0];
    if (!latest) throw new Error(`no process sample for pid ${pid} yet`);
    return {
      pid,
      rss: latest.rssBytes,
      ...(latest.cpuPercent != null ? { cpu_percent: latest.cpuPercent } : {}),
      ...(latest.openFiles !== undefined ? { open_files: latest.openFiles } : {}),
    };
  };
}
