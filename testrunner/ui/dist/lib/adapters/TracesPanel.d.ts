import { TraceResult, TracesStatus } from '../types';
/**
 * One fixture's trace windows, opened in the stores of the runs they name.
 * `traces` is status.traces with its mounts as the browser reaches them
 * (rebaseTraces), so every link and read goes through the run's own server.
 */
export declare function TracesPanel({ windows, traces }: {
    windows: TraceResult[];
    traces: TracesStatus | undefined;
}): import("react").JSX.Element;
//# sourceMappingURL=TracesPanel.d.ts.map