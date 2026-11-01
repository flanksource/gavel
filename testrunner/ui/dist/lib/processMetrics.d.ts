import { FetchProcessMetrics } from '@flanksource/clicky-ui/data';
import { ProcessSample, ProcessTraceNode } from './types';
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
export declare function fetchProcessMetricsFromSamples({ samples, nodes }: {
    samples: ProcessSample[];
    nodes: ProcessTraceNode[];
}): FetchProcessMetrics;
//# sourceMappingURL=processMetrics.d.ts.map