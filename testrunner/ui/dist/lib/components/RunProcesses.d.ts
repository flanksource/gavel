import { ProcessTraceState } from '../hooks/use-process-trace';
import { ProcessStepLink } from '../utils';
export interface RunProcessesProps {
    /** The run's process capture, as useProcessTrace reads it. */
    trace: ProcessTraceState;
    /** The key of the selected process, or null. */
    selectedKey: string | null;
    onSelect: (key: string) => void;
    hideExited: boolean;
    onHideExitedChange: (hide: boolean) => void;
    expandAll: boolean | null;
    onExpandAllChange: (next: boolean | null) => void;
    /** The fixture step that started each pid, linked from the process that has one. */
    stepLinks?: Map<number, ProcessStepLink>;
    onOpenStep?: (link: ProcessStepLink) => void;
    className?: string;
}
/**
 * The Processes view: the run's process tree beside the selected process's
 * detail (command, lifetime, usage, stack dumps with their collect buttons,
 * environment and the facet tabs the samples carry) and, when the run ingested
 * Go profiles for it, its profiles. Collect and refresh need the run's live
 * capture (`trace.canCollect`): a stored run shows what was recorded and offers
 * neither.
 */
export declare function RunProcesses({ trace, selectedKey, onSelect, hideExited, onHideExitedChange, expandAll, onExpandAllChange, stepLinks, onOpenStep, className, }: RunProcessesProps): import("react").JSX.Element;
//# sourceMappingURL=RunProcesses.d.ts.map