/** One recorded run from `/api/runs` (snapshots.RunInfo plus the URL of its view). */
export interface RunInfo {
    runId: string;
    path: string;
    started?: string;
    ended?: string;
    kind: 'test' | 'lint' | 'test+lint';
    repo?: string;
    sha?: string;
    frameworks?: string[];
    passed: number;
    failed: number;
    skipped: number;
    warned: number;
    total: number;
    lintViolations: number;
    lintLinters: number;
    href: string;
}
/** One page of `/api/runs`. */
export interface RunsPage {
    /** The run this page is viewing; absent for the default (latest) view. */
    current?: string;
    /** The runs merged into the default view at latestHref. */
    latest?: string[];
    latestHref: string;
    runs: RunInfo[];
    /** Cursor for the following page; absent when none follows. */
    next?: string;
}
/**
 * Fetches a page of the run history. Resolves null when the host serves no
 * history (a live `gavel test --ui` run, or a page embedded in another app).
 */
export declare function fetchRunsPage({ before, signal }?: {
    before?: string;
    signal?: AbortSignal;
}): Promise<RunsPage | null>;
/** The URL that opens a run on the tab holding its results: Lint for a lint-only run, Tests otherwise. */
export declare function runViewHref(run: Pick<RunInfo, 'href' | 'kind'>): string;
/** "Oct 22, 15:31" — the run's start in the viewer's locale, or its id when unknown. */
export declare function runStartLabel(run: Pick<RunInfo, 'runId' | 'started'>): string;
/** A run's results in a few words: "2 failed / 40 tests · 3 lint". */
export declare function runResultLabel(run: RunInfo): string;
export declare function runDurationMs(run: Pick<RunInfo, 'started' | 'ended'>): number | undefined;
//# sourceMappingURL=runHistory.d.ts.map