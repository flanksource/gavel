import { RunInfo } from '../runHistory';
export interface RunHistory {
    /** The host serves a run history; false until the first page answers, and for hosts without one. */
    available: boolean;
    /** The run being viewed; undefined for the default (latest) view. */
    current?: string;
    latest: string[];
    latestHref: string;
    /** Every page fetched so far, newest first. */
    runs: RunInfo[];
    hasMore: boolean;
    loading: boolean;
    error?: string;
    loadMore: () => void;
}
/** Loads the run history page by page from `/api/runs` under the page's base path. */
export declare function useRunHistory(): RunHistory;
//# sourceMappingURL=use-run-history.d.ts.map