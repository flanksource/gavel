import {
  createTestRunnerRegistry,
  verificationAdapters,
  type Test as ClickyTest,
  type TestNodeAdapter,
  type TestNodeAdapterRegistry,
} from '@flanksource/clicky-ui/data';
import type { FixtureContext, ProcessTraceNode, Test, TraceResult, TracesStatus } from '../types';
import { GavelDetailBody } from './GavelDetailBody';
import { NodeActions, type NodeActionsConfig } from './NodeActions';
import { ResourcesPanel } from './ResourcesPanel';
import { StepProcesses } from './StepProcesses';
import { TracesPanel } from './TracesPanel';

export type { NodeActionsConfig } from './NodeActions';

const asGavel = (node: ClickyTest): Test => node as Test;

// fixtureTraces are the trace windows a fixture node armed, none for any other node.
function fixtureTraces(node: ClickyTest): TraceResult[] {
  const gavel = asGavel(node);
  if (gavel.framework !== 'fixture') return [];
  return (gavel.context as FixtureContext | undefined)?.traces ?? [];
}

// stepProcesses are the nodes of the run's process tree whose pid the fixture step started.
function stepProcesses(node: ClickyTest, processes: ProcessTraceNode[]): ProcessTraceNode[] {
  const pids = new Set(asGavel(node).fixture?.pids ?? []);
  if (pids.size === 0) return [];
  return processes.filter(process => pids.has(process.pid));
}

/**
 * The detail body of every node that is not a fixture step: suite, package,
 * command, error, Go and Ginkgo context, attempt metrics and output. Fixture
 * steps keep clicky's verification body (command, CEL, typed step sections).
 */
export function detailAdapter(): TestNodeAdapter {
  return {
    id: 'gavel-detail',
    match: node => asGavel(node).framework !== 'fixture',
    renderDetail: ({ node }) => <GavelDetailBody node={node} />,
  };
}

/** The Traces tab of a fixture node: its windows, opened in the stores of the runs they name (traces rebased onto the page). */
export function tracesAdapter(traces: TracesStatus | undefined): TestNodeAdapter {
  return {
    id: 'gavel-traces',
    match: node => fixtureTraces(node).length > 0,
    detailTabs: ({ node }) => [{
      id: 'traces',
      label: 'Traces',
      render: () => <TracesPanel windows={fixtureTraces(node)} traces={traces} />,
    }],
  };
}

/** The Processes tab of a fixture step: the process-tree nodes whose pids the step started. */
export function processesAdapter({ processes, onOpenProcess }: { processes: ProcessTraceNode[]; onOpenProcess: (key: string) => void }): TestNodeAdapter {
  return {
    id: 'gavel-processes',
    match: node => stepProcesses(node, processes).length > 0,
    detailTabs: ({ node }) => [{
      id: 'processes',
      label: 'Processes',
      render: () => <StepProcesses processes={stepProcesses(node, processes)} onOpenProcess={onOpenProcess} />,
    }],
  };
}

/** The Resources tab of a fixture node with resource profiling: process-tree peaks and SQL. */
export function resourcesAdapter(): TestNodeAdapter {
  return {
    id: 'gavel-resources',
    match: node => {
      const gavel = asGavel(node);
      return !!gavel.fixture_profile || !!gavel.sql_profile;
    },
    detailTabs: ({ node }) => [{
      id: 'resources',
      label: 'Resources',
      render: () => <ResourcesPanel node={asGavel(node)} />,
    }],
  };
}

/** Download, AI-prompt copy and Skip/Delete source actions in the node's detail header. */
export function nodeActionsAdapter(config: NodeActionsConfig): TestNodeAdapter {
  return {
    id: 'gavel-actions',
    match: node => asGavel(node).framework !== 'task',
    nodeActions: ({ node }) => <NodeActions node={asGavel(node)} config={config} />,
  };
}

export interface GavelAdapterOptions {
  /** status.traces with its mounts rebased onto the page (rebaseTraces). */
  traces: TracesStatus | undefined;
  /** The run's process tree, for the Processes tab of a fixture step. */
  processes: ProcessTraceNode[];
  onOpenProcess: (key: string) => void;
  actions: NodeActionsConfig;
}

/** Gavel's node adapters, in the order they take precedence. */
export function gavelAdapters({ traces, processes, onOpenProcess, actions }: GavelAdapterOptions): TestNodeAdapter[] {
  return [
    detailAdapter(),
    nodeActionsAdapter(actions),
    tracesAdapter(traces),
    processesAdapter({ processes, onOpenProcess }),
    resourcesAdapter(),
  ];
}

/**
 * The registry the test runner renders gavel runs with: gavel's adapters, then
 * clicky's verification adapters, which own fixture steps (attempt chooser,
 * typed step sections, copy brief).
 */
export function gavelRegistry(options: GavelAdapterOptions): TestNodeAdapterRegistry {
  return createTestRunnerRegistry([...gavelAdapters(options), ...verificationAdapters().adapters]);
}
