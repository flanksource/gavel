import { RouteState } from '../routes';
/** Download of the view (or one node of it) as JSON or Markdown, served by the export routes. */
export declare function DownloadMenu({ routeState, title }: {
    routeState: RouteState;
    title?: string;
}): import("react").JSX.Element;
/** Copies the Markdown export of `routeState` framed as an AI prompt. */
export declare function CopyAIPromptButton({ routeState, disabled, disabledTitle, title }: {
    routeState: RouteState;
    disabled?: boolean;
    disabledTitle?: string;
    title?: string;
}): import("react").JSX.Element;
//# sourceMappingURL=ExportControls.d.ts.map