import { RunTraces, TraceRef, TracesStatus } from './types';
export type TraceRecord = Record<string, unknown>;
export interface TraceExportFormat {
    name: string;
    ext: string;
    mime: string;
}
export interface TraceKindInfo {
    name: string;
    title: string;
    exportFormats: TraceExportFormat[];
}
export declare class TraceStreamNotFound extends Error {
}
/** Run runId's store as status reaches it: its mount, or why it has none. */
export declare function runTraces(status: TracesStatus | undefined, runId: string): RunTraces;
/** The mount of the store holding the Processes tab's capture; undefined when not served. */
export declare function processMount(status: TracesStatus | undefined): string | undefined;
/**
 * status with every mount under base: a host that serves the runner's server
 * under a prefix of its own (the page's base path, a MultiServer run id, or
 * pr/ui's /api/project-runs/<id>) rebases the mounts the server reports
 * relative to itself.
 */
export declare function rebaseTraces(status: TracesStatus | undefined, base: string): TracesStatus | undefined;
export declare function queryUrl(mount: string, path: string): string;
/** The download of ref's window in one of its kind's export formats. */
export declare function traceExportUrl(mount: string, ref: TraceRef, format: string): string;
/** The query app's page for ref's window: its kind's result profile, filtered to the window. */
export declare function openInQueryUrl(mount: string, ref: TraceRef): string;
export declare function failure(res: Response, what: string): Promise<Error>;
/** Parses an NDJSON body: one record per non-empty line. */
export declare function parseNdjson(body: string): TraceRecord[];
/**
 * The records of ref's window as its stream holds them: the NDJSON export.
 * The result profile's JSON rows leave out the columns a kind hides (a
 * process's key and stack, a terminal event's data), which the viewers need.
 */
export declare function fetchTraceRecords(mount: string, ref: TraceRef, signal?: AbortSignal): Promise<TraceRecord[]>;
/**
 * The id the query server surfaces a kind's result profile (`traces/<kind>`)
 * under: commons-db `profiles.SurfaceKey`, `profile-` plus the slugged name.
 */
export declare function traceProfileId(kind: string): string;
/** The highest seq ref's stream has committed, read from its result profile; 0 when it has none. */
export declare function fetchLatestSeq(mount: string, ref: TraceRef): Promise<number>;
/** The trace kinds the store serves, with the formats each exports in. */
export declare function fetchTraceKinds(mount: string): Promise<TraceKindInfo[]>;
export interface CheckpointRequest {
    mount: string;
    session: string;
    /**
     * The kind's checkpoint params: `{key, transport, reason}` takes a stack dump
     * of one process on a `stacks` session, `{key}` forces a full facet reading
     * of one process on a `process_samples` session.
     */
    params: Record<string, string>;
}
/** Has the running session take a reading now, as its kind's checkpoint params describe it. */
export declare function checkpointSession({ mount, session, params }: CheckpointRequest): Promise<void>;
//# sourceMappingURL=traces.d.ts.map