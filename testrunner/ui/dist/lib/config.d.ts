export declare const basePath: string;
/** The URL prefix of one run's server: the page's base path, then the run id when a MultiServer hosts many runs. */
export declare function apiBase({ baseUrl, runId }?: {
    baseUrl?: string;
    runId?: string;
}): string;
export declare function apiUrl(path: string, scope?: {
    baseUrl?: string;
    runId?: string;
}): string;
//# sourceMappingURL=config.d.ts.map