import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { checkpointSession, fetchLatestSeq, fetchTraceRecords, processMount, TraceStreamNotFound } from '../traces';
import type { ProcessSample, ProcessTraceNode, TraceRef, TracesStatus } from '../types';

const POLL_INTERVAL_MS = 1000;

// SAMPLE_WINDOW bounds how many of the latest samples (of every process) a
// read of the samples stream fetches, so a long run's history is not
// re-downloaded on every poll.
export const SAMPLE_WINDOW = 2000;

export const PROCESS_TRACE_ENDED = 'The run has ended, so its process trace can no longer be checkpointed';

export interface ProcessStack {
  text?: string;
  error?: string;
  collectedAt?: string;
}

export interface UseProcessTraceOptions {
  // traces is status.traces: the process capture is read through the store's
  // query mount, and nothing is fetched until both are there.
  traces?: TracesStatus;
  // done stops polling: a finished or replayed run is fetched once.
  done: boolean;
  selectedKey?: string;
}

export interface ProcessTraceState {
  nodes: ProcessTraceNode[];
  samples: ProcessSample[];
  stack?: ProcessStack;
  error?: string;
  collecting: boolean;
  collectMessage?: string;
  collectStack: () => Promise<void>;
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

function stackFromNode(node: ProcessTraceNode | undefined): ProcessStack | undefined {
  if (!node) return undefined;
  const stack: ProcessStack = {
    text: node.stack || undefined,
    error: node.stackError || undefined,
    collectedAt: node.stackCollectedAt || undefined,
  };
  return stack.text || stack.error ? stack : undefined;
}

function message(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

function usePolled<T>(enabled: boolean, poll: boolean, load: () => Promise<T>, key: string) {
  const [value, setValue] = useState<T | null>(null);
  const [error, setError] = useState<string | undefined>(undefined);
  const loadRef = useRef(load);
  loadRef.current = load;

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

export function useProcessTrace({ traces, done, selectedKey }: UseProcessTraceOptions): ProcessTraceState {
  const mount = processMount(traces);
  const refs = traces?.process;
  const live = !done;

  const tree = usePolled<ProcessTraceNode[]>(
    !!mount && !!refs, live, () => readStream<ProcessTraceNode>(mount!, refs!.tree), `${mount}|${refs?.tree.stream}`,
  );
  const sampled = usePolled<ProcessSample[]>(
    !!mount && !!refs && !!selectedKey, live, () => readRecentSamples(mount!, refs!.samples),
    `${mount}|${refs?.samples.stream}|${selectedKey}`,
  );

  const [collecting, setCollecting] = useState(false);
  const [collectMessage, setCollectMessage] = useState<string | undefined>(undefined);

  const collectStack = useCallback(async () => {
    if (!mount || !traces?.process_session) {
      setCollectMessage(PROCESS_TRACE_ENDED);
      return;
    }
    setCollecting(true);
    setCollectMessage(undefined);
    try {
      await checkpointSession(mount, traces.process_session);
      await tree.reload();
    } catch (e: unknown) {
      setCollectMessage(message(e));
    } finally {
      setCollecting(false);
    }
  }, [mount, traces?.process_session, tree.reload]);

  const nodes = useMemo(() => tree.value ?? [], [tree.value]);
  const samples = useMemo(
    () => (sampled.value ?? []).filter(sample => sample.key === selectedKey),
    [sampled.value, selectedKey],
  );
  const stack = useMemo(() => stackFromNode(nodes.find(node => node.isRoot)), [nodes]);

  return { nodes, samples, stack, error: tree.error ?? sampled.error, collecting, collectMessage, collectStack };
}
