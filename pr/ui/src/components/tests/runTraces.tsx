import { useState } from 'react';
import type { TestNodeAdapter } from '@flanksource/clicky-ui/data';
import { ProcessesView, TracesSection, useProcessTrace } from '@flanksource/gavel/testrunner';
import type { TraceResult, TracesStatus } from '@flanksource/gavel/testrunner';

// fixtureTraces are the trace windows a fixture node armed, none for any other node.
function fixtureTraces(node: unknown): TraceResult[] {
  const fixture = node as { framework?: string; context?: { traces?: TraceResult[] } };
  if (fixture.framework !== 'fixture') return [];
  return fixture.context?.traces ?? [];
}

/**
 * The Traces tab of a fixture node: its windows, opened in the run stores the
 * dashboard serves (traces carries their mounts as the dashboard serves them).
 */
export function tracesAdapter(traces: TracesStatus | undefined): TestNodeAdapter {
  return {
    id: 'gavel-traces',
    match: node => fixtureTraces(node).length > 0,
    detailTabs: ({ node }) => [{
      id: 'traces',
      label: 'Traces',
      render: () => <div className="p-3"><TracesSection windows={fixtureTraces(node)} traces={traces} /></div>,
    }],
  };
}

/** The run's process capture as a tree, read through its store's query mount. */
export function RunProcesses({ traces, done }: { traces: TracesStatus; done: boolean }) {
  const [selected, setSelected] = useState<string | null>(null);
  const [hideExited, setHideExited] = useState(false);
  const trace = useProcessTrace({ traces, done, selectedKey: selected ?? undefined });
  return (
    <div className="h-full overflow-auto">
      <ProcessesView nodes={trace.nodes} selectedKey={selected} expandAll={null} hideExited={hideExited}
        error={trace.error} onHideExitedChange={setHideExited} onSelect={setSelected} />
    </div>
  );
}
