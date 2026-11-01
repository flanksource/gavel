import { TestNodeAdapter, TestNodeAdapterRegistry } from '@flanksource/clicky-ui/data';
import { ProcessTraceNode, TracesStatus } from '../types';
import { NodeActionsConfig } from './NodeActions';
export type { NodeActionsConfig } from './NodeActions';
/**
 * The detail body of every node that is not a fixture step: suite, package,
 * command, error, Go and Ginkgo context, attempt metrics and output. Fixture
 * steps keep clicky's verification body (command, CEL, typed step sections).
 */
export declare function detailAdapter(): TestNodeAdapter;
/** The Traces tab of a fixture node: its windows, opened in the stores of the runs they name (traces rebased onto the page). */
export declare function tracesAdapter(traces: TracesStatus | undefined): TestNodeAdapter;
/** The Processes tab of a fixture step: the process-tree nodes whose pids the step started. */
export declare function processesAdapter({ processes, onOpenProcess }: {
    processes: ProcessTraceNode[];
    onOpenProcess: (key: string) => void;
}): TestNodeAdapter;
/** The Resources tab of a fixture node with resource profiling: process-tree peaks and SQL. */
export declare function resourcesAdapter(): TestNodeAdapter;
/** Download, AI-prompt copy and Skip/Delete source actions in the node's detail header. */
export declare function nodeActionsAdapter(config: NodeActionsConfig): TestNodeAdapter;
export interface GavelAdapterOptions {
    /** status.traces with its mounts rebased onto the page (rebaseTraces). */
    traces: TracesStatus | undefined;
    /** The run's process tree, for the Processes tab of a fixture step. */
    processes: ProcessTraceNode[];
    onOpenProcess: (key: string) => void;
    actions: NodeActionsConfig;
}
/** Gavel's node adapters, in the order they take precedence. */
export declare function gavelAdapters({ traces, processes, onOpenProcess, actions }: GavelAdapterOptions): TestNodeAdapter[];
/**
 * The registry the test runner renders gavel runs with: gavel's adapters, then
 * clicky's verification adapters, which own fixture steps (attempt chooser,
 * typed step sections, copy brief).
 */
export declare function gavelRegistry(options: GavelAdapterOptions): TestNodeAdapterRegistry;
//# sourceMappingURL=index.d.ts.map