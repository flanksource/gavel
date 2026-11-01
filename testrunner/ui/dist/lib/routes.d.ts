import { FilterState, LintFilters, LintGrouping, TestFilters } from '@flanksource/clicky-ui/data';
import { FileReviewView } from '@flanksource/clicky-ui/file-findings';
export declare function defaultStatusFilter(): FilterState<string>;
export type TabKey = 'tests' | 'lint' | 'files' | 'bench' | 'processes' | 'history';
export interface RouteState {
    tab: TabKey;
    selectedPath: string;
    workDir?: string;
    view?: FileReviewView;
    line?: number;
    filters: TestFilters;
    lintGrouping: LintGrouping;
    lintFilters: LintFilters;
    sampleIndex?: number;
    hideExited?: boolean;
}
export type ExportFormat = 'json' | 'md' | 'pdf';
export declare function parseRoute(location: Pick<Location, 'pathname' | 'search'>): RouteState;
export declare function buildRoute(state: RouteState): string;
export declare function buildExportRoute(state: RouteState, format: ExportFormat): string;
//# sourceMappingURL=routes.d.ts.map