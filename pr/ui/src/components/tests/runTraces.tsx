import { useEffect, useMemo, useState } from 'react';
import { createTestRunnerRegistry, verificationAdapters, type TestNodeAdapterRegistry } from '@flanksource/clicky-ui/data';
import {
  detailAdapter,
  processMount,
  processesAdapter,
  rebaseTraces,
  resourcesAdapter,
  tracesAdapter,
  useProcessTrace,
} from '@flanksource/gavel/testrunner';
import type { ProcessTraceNode, ProcessTraceState, TracesStatus } from '@flanksource/gavel/testrunner';

export interface RunProcessesView {
  /** The run's process capture, as the Processes view and the step Processes tabs read it. */
  trace: ProcessTraceState;
  /** True when the run kept a process capture that the dashboard serves. */
  available: boolean;
  /** status.traces rebased onto the server's base once, for the Traces tabs. */
  traces: TracesStatus | undefined;
  selectedKey: string | null;
  setSelectedKey: (key: string) => void;
  hideExited: boolean;
  setHideExited: (hide: boolean) => void;
  expandAll: boolean | null;
  setExpandAll: (next: boolean | null) => void;
}

/**
 * The state behind a run's Processes view. `traces` is status.traces exactly as
 * the server that returned the snapshot reported it and `base` the URL prefix of
 * that server ('' for the dashboard itself, `/api/project-runs/<id>` for a run
 * server): the process trace rebases them itself, so the only rebase here is
 * the one the Traces tabs need.
 */
export function useRunProcesses({ traces: reported, base, done, active }: {
  traces: TracesStatus | undefined;
  base: string;
  done: boolean;
  /** Whether the Processes view is showing: the root process is selected on arrival. */
  active: boolean;
}): RunProcessesView {
  const [selectedKey, setSelectedKey] = useState<string | null>(null);
  const [hideExited, setHideExited] = useState(false);
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  const trace = useProcessTrace({ traces: reported, done, base, ...(selectedKey ? { selectedKey } : {}) });
  const traces = useMemo(() => rebaseTraces(reported, base), [reported, base]);
  const rootKey = trace.nodes.find(node => node.isRoot)?.key;

  useEffect(() => {
    if (active && selectedKey === null && rootKey) setSelectedKey(rootKey);
  }, [active, selectedKey, rootKey]);

  return {
    trace,
    available: Boolean(processMount(reported)),
    traces,
    selectedKey,
    setSelectedKey,
    hideExited,
    setHideExited,
    expandAll,
    setExpandAll,
  };
}

/**
 * The registry the Tests tab renders a run with: gavel's detail, Traces,
 * Processes and Resources adapters, then clicky's verification adapters, which
 * own fixture steps. `traces` is status.traces already rebased onto `base`.
 * Source editing and export downloads are not wired: the dashboard serves
 * neither for a stored run.
 */
export function runRegistry({ traces, processes, onOpenProcess }: {
  traces: TracesStatus | undefined;
  processes: ProcessTraceNode[];
  onOpenProcess: (key: string) => void;
}): TestNodeAdapterRegistry {
  return createTestRunnerRegistry([
    detailAdapter(),
    tracesAdapter(traces),
    processesAdapter({ processes, onOpenProcess }),
    resourcesAdapter(),
    ...verificationAdapters().adapters,
  ]);
}
