import { BenchComparison, LinterResult, RerunRequest, RunMeta, Snapshot, SnapshotStatus, Test, TestEditRequest } from '../types';
export interface EventSourceLike {
    addEventListener(type: string, listener: EventListenerOrEventListenerObject): void;
    removeEventListener(type: string, listener: EventListenerOrEventListenerObject): void;
    close(): void;
    onerror: ((event: Event) => void) | null;
}
export interface UseTestRunOptions {
    /** API prefix, e.g. "" or "https://host" or "/results/{repo}/{id}". */
    baseUrl?: string;
    /** Alias for baseUrl kept for parity with useTaskRun. Defaults to
     *  window.__gavelBasePath ?? "". */
    basePath?: string;
    /** Scope every request to a specific run id, hitting
     *  `${base}/${runId}/api/...` instead of `${base}/api/...`. Set when the host
     *  serves many concurrent runs (gavel MultiServer); omit for a single-run
     *  server mounted at the root. */
    runId?: string;
    /** Disable the subscription entirely (e.g. before a target is known). */
    enabled?: boolean;
    /** Poll interval (ms) for the fallback transport. */
    pollMs?: number;
    /** Force the polling transport even when EventSource exists (mainly tests). */
    forcePoll?: boolean;
    /** Builds the SSE connection for a URL. Defaults to `(url) => new
     *  EventSource(url)`. A host embedding this hook behind a connection
     *  multiplexer (e.g. pr/ui's eventHub) passes its own factory here instead
     *  of letting this hook open a raw `EventSource`. */
    createEventSource?: (url: string) => EventSourceLike;
}
export interface UseTestRunResult {
    snapshot: Snapshot | undefined;
    tests: Test[];
    lint: LinterResult[] | undefined;
    bench: BenchComparison | undefined;
    status: SnapshotStatus;
    /** Human-readable run state, e.g. "Running tests...", "Test run complete". */
    statusText: string;
    runMeta: RunMeta | undefined;
    done: boolean;
    error: string | undefined;
    rerun: (body?: RerunRequest) => Promise<ActionResult>;
    stop: (taskId?: string) => Promise<ActionResult>;
    editTest: (body: TestEditRequest) => Promise<ActionResult>;
    /** Reconnect the stream / restart polling. */
    refetch: () => void;
}
/**
 * How a POSTed action ended. A failed one is also reported in `error`; the
 * result lets the caller tell a rerun refused because another is in progress
 * (`conflict`) from one that failed.
 */
export interface ActionResult {
    ok: boolean;
    /** The server's HTTP status. */
    status: number;
    /** Another rerun is in progress (409). */
    conflict: boolean;
    /** The server's own message: the failure text, or a test edit's summary. */
    message?: string;
}
export declare function deriveStatusText(snap: Snapshot): string;
export declare function useTestRun(options?: UseTestRunOptions): UseTestRunResult;
//# sourceMappingURL=use-test-run.d.ts.map