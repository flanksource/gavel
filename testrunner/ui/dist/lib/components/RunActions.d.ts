import { RouteState } from '../routes';
import { RunMeta } from '../types';
/** The run's timeout knobs (--timeout, --test-timeout, --lint-timeout) that were set, as pills. */
export declare function TimeoutPills({ runMeta }: {
    runMeta: RunMeta | undefined;
}): import("react").JSX.Element | null;
export interface RunActionsProps {
    runMeta: RunMeta | undefined;
    /** The run can be stopped and is running. */
    canStop: boolean;
    stopping: boolean;
    onStop: () => void;
    /** The route state that exports the whole current view; absent when there is nothing to export. */
    exportState: RouteState | undefined;
}
/**
 * The run-level actions of the page header: the timeout pills, the global
 * Stop, and the whole view's Download and AI-prompt copy.
 */
export declare function RunActions({ runMeta, canStop, stopping, onStop, exportState }: RunActionsProps): import("react").JSX.Element;
//# sourceMappingURL=RunActions.d.ts.map