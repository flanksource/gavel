import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { GoProfileRecord, StackRecord } from '@flanksource/clicky-ui/data';
import { apiBase } from '../config';
import {
  checkpointSession,
  fetchLatestSeq,
  fetchTraceRecords,
  processMount,
  rebaseTraces,
  TraceStreamNotFound,
  traceExportUrl,
} from '../traces';
import type { ProcessSample, ProcessTraceNode, TraceRef, TracesStatus } from '../types';

const POLL_INTERVAL_MS = 1000;

// SAMPLE_WINDOW bounds how many of the latest samples (of every process) a
// read of the samples stream fetches, so a long run's history is not
// re-downloaded on every poll.
export const SAMPLE_WINDOW = 2000;

export const PROCESS_TRACE_ENDED = 'The run has ended, so its process trace can no longer be checkpointed';

export interface UseProcessTraceOptions {
  // traces is status.traces: the process capture is read through the store's
  // query mount, and nothing is fetched until both are there.
  traces?: TracesStatus;
  // done stops polling: a finished or replayed run is fetched once, and has no
  // live session to checkpoint.
  done: boolean;
  // selectedKey is the process whose stack dumps and Go profiles are read.
  selectedKey?: string;
  // base is the URL prefix of the run's server: the page's base path by default,
  // plus the run id when a MultiServer hosts many runs (or pr/ui's
  // /api/project-runs/<id>). Mounts in traces are relative to the server and
  // are rebased onto it here, so pass status.traces as the server reported it.
  base?: string;
}

export interface ProcessTraceState {
  // nodes are the process_tree records, each with the stack transports the
  // store offers for it.
  nodes: ProcessTraceNode[];
  // samples are the latest process_samples records of every process: resources,
  // disk IO and the facet counts and lists the session armed.
  samples: ProcessSample[];
  // stacks are the stacks records, read while a process is selected.
  stacks: StackRecord[];
  // goprof are the goprof records, read while a process is selected and when the run ingests Go profiles.
  goprof: GoProfileRecord[];
  // goprofExportUrl downloads the goprof window as a pprof profile (`go tool pprof`); absent when the run has no goprof.
  goprofExportUrl?: string;
  // loading is true until the first read of the tree has answered.
  loading: boolean;
  error?: string;
  // canCollect is true while the run's capture is live: only then can a stack
  // be collected or a facet reading forced.
  canCollect: boolean;
  collecting: boolean;
  collectError?: string;
  collectStack: (key: string, transport: string) => Promise<void>;
  refreshing: boolean;
  refreshError?: string;
  refreshFacets: (key: string) => Promise<void>;
}

// readStream is ref's records, none while its stream has committed none yet.
async function readStream<T>(mount: string, ref: TraceRef): Promise<T[]> {
  try {
    return await fetchTraceRecords(mount, ref) as unknown as T[];
  } catch (e) {
    if (e instanceof TraceStreamNotFound) return [];
    throw e;
  }
}

// readRecentSamples is the last SAMPLE_WINDOW samples of the stream.
async function readRecentSamples(mount: string, ref: TraceRef): Promise<ProcessSample[]> {
  const latest = await fetchLatestSeq(mount, ref);
  if (latest === 0) return [];
  return readStream<ProcessSample>(mount, { ...ref, afterSeq: Math.max(0, latest - SAMPLE_WINDOW) });
}

function message(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

function usePolled<T>(enabled: boolean, poll: boolean, load: () => Promise<T>, key: string) {
  const [value, setValue] = useState<T | null>(null);
  const [error, setError] = useState<string | undefined>(undefined);
  const loadRef = useRef(load);
  useEffect(() => {
    loadRef.current = load;
  });

  const reload = useCallback(async () => {
    try {
      setValue(await loadRef.current());
      setError(undefined);
    } catch (e: unknown) {
      setError(message(e));
    }
  }, []);

  useEffect(() => {
    if (!enabled) return;
    let cancelled = false;
    const run = () => {
      loadRef.current().then(next => {
        if (cancelled) return;
        setValue(next);
        setError(undefined);
      }).catch((e: unknown) => {
        if (!cancelled) setError(message(e));
      });
    };
    run();
    const timer = poll ? window.setInterval(run, POLL_INTERVAL_MS) : undefined;
    return () => {
      cancelled = true;
      if (timer !== undefined) window.clearInterval(timer);
    };
  }, [enabled, poll, key]);

  return { value, error, reload };
}

// useCheckpoint posts checkpoint params to a running session, then re-reads the
// window the checkpoint wrote to. Without a session (the run has ended) it
// fails loudly instead of posting.
function useCheckpoint(mount: string | undefined, session: string | undefined, reload: () => Promise<void>) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | undefined>(undefined);
  const checkpoint = useCallback(async (params: Record<string, string>) => {
    setBusy(true);
    setError(undefined);
    try {
      if (!mount || !session) {
        setError(PROCESS_TRACE_ENDED);
        return;
      }
      await checkpointSession({ mount, session, params });
      await reload();
    } catch (e: unknown) {
      setError(message(e));
    } finally {
      setBusy(false);
    }
  }, [mount, session, reload]);
  return { busy, error, checkpoint };
}

export function useProcessTrace({ traces: reported, done, selectedKey, base }: UseProcessTraceOptions): ProcessTraceState {
  const scope = base ?? apiBase();
  const traces = useMemo(() => rebaseTraces(reported, scope), [reported, scope]);
  const mount = processMount(traces);
  const refs = traces?.process;
  const live = !done;
  const enabled = !!mount && !!refs;
  const sessions = live ? reported?.sessions : undefined;

  const tree = usePolled<ProcessTraceNode[]>(
    enabled, live, () => readStream<ProcessTraceNode>(mount!, refs!.tree), `${mount}|${refs?.tree.stream}`,
  );
  const sampled = usePolled<ProcessSample[]>(
    enabled, live, () => readRecentSamples(mount!, refs!.samples), `${mount}|${refs?.samples.stream}`,
  );
  const stacked = usePolled<StackRecord[]>(
    enabled && !!selectedKey, live, () => readStream<StackRecord>(mount!, refs!.stacks), `${mount}|${refs?.stacks.stream}`,
  );
  const profiled = usePolled<GoProfileRecord[]>(
    enabled && !!selectedKey && !!refs?.goprof, live, () => readStream<GoProfileRecord>(mount!, refs!.goprof!), `${mount}|${refs?.goprof?.stream}`,
  );

  const stackCheckpoint = useCheckpoint(mount, sessions?.stacks, stacked.reload);
  const sampleCheckpoint = useCheckpoint(mount, sessions?.samples, sampled.reload);
  const collectStack = useCallback(
    (key: string, transport: string) => stackCheckpoint.checkpoint({ key, transport, reason: 'manual' }),
    [stackCheckpoint.checkpoint],
  );
  const refreshFacets = useCallback((key: string) => sampleCheckpoint.checkpoint({ key }), [sampleCheckpoint.checkpoint]);

  const nodes = useMemo(() => tree.value ?? [], [tree.value]);
  const samples = useMemo(() => sampled.value ?? [], [sampled.value]);
  const stacks = useMemo(() => stacked.value ?? [], [stacked.value]);
  const goprof = useMemo(() => profiled.value ?? [], [profiled.value]);

  return {
    nodes,
    samples,
    stacks,
    goprof,
    ...(mount && refs?.goprof ? { goprofExportUrl: traceExportUrl(mount, refs.goprof, 'pprof') } : {}),
    loading: enabled && tree.value === null && !tree.error,
    error: tree.error ?? sampled.error ?? stacked.error ?? profiled.error,
    canCollect: enabled && !!sessions,
    collecting: stackCheckpoint.busy,
    ...(stackCheckpoint.error ? { collectError: stackCheckpoint.error } : {}),
    collectStack,
    refreshing: sampleCheckpoint.busy,
    ...(sampleCheckpoint.error ? { refreshError: sampleCheckpoint.error } : {}),
    refreshFacets,
  };
}
