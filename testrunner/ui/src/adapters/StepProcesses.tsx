import { ProcessTraceTree } from '@flanksource/clicky-ui/data';
import type { ProcessTraceNode } from '../types';

/**
 * The processes a fixture step started, as the slice of the run's process tree
 * whose pids it lists. Selecting one opens it on the Processes tab.
 */
export function StepProcesses({ processes, onOpenProcess }: { processes: ProcessTraceNode[]; onOpenProcess: (key: string) => void }) {
  return (
    <div className="flex h-80 min-h-0 flex-col px-3 pt-2" data-step-processes="">
      <ProcessTraceTree
        nodes={processes}
        selectedKey={null}
        expandAll={true}
        onSelect={onOpenProcess}
        emptyLabel="This step started no processes the trace recorded"
      />
    </div>
  );
}
