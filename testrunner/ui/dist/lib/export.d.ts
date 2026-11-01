import { ExportFormat, RouteState } from './routes';
export declare function downloadCurrentView(state: RouteState, format: Exclude<ExportFormat, 'pdf'>): void;
/** Copies the view's Markdown export, framed as an AI prompt; rejects when the export or the clipboard fails. */
export declare function copyCurrentViewForAgent(state: RouteState): Promise<void>;
//# sourceMappingURL=export.d.ts.map