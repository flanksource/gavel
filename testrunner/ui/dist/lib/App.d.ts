export declare function sourceUrl(url: string | undefined, line: number | undefined): string | undefined;
export interface AppProps {
    /** Scopes every request to one run of a MultiServer (`${base}/${runId}/api/...`). */
    runId?: string;
    /** Opens another run's page; each run is served under its own base path, so this is a full page load. */
    navigate?: (href: string) => void;
}
export declare function App({ runId, navigate }?: AppProps): import("react").JSX.Element;
//# sourceMappingURL=App.d.ts.map