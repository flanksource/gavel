import { ProcessTraceNode } from '../types';
/**
 * The processes a fixture step started, as the slice of the run's process tree
 * whose pids it lists. Selecting one opens it on the Processes tab.
 */
export declare function StepProcesses({ processes, onOpenProcess }: {
    processes: ProcessTraceNode[];
    onOpenProcess: (key: string) => void;
}): import("react").JSX.Element;
//# sourceMappingURL=StepProcesses.d.ts.map