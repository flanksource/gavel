export type RerunStatus = 'running' | 'success' | 'failed' | 'canceled';
interface Props {
    open: boolean;
    onClose: () => void;
    /** URL of the rerun's SSE stream (/api/rerun/stream under the run's base). */
    streamUrl: string;
}
/**
 * The output of the rerun in progress, streamed from the server. The server
 * sends each write once, and replays them all to a client that reconnects, so
 * every (re)connection starts the output over.
 */
export declare function RerunDialog({ open, onClose, streamUrl }: Props): import("react").JSX.Element;
export {};
//# sourceMappingURL=RerunDialog.d.ts.map