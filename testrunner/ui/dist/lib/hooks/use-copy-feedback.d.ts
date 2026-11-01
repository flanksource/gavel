export type CopyState = 'idle' | 'copying' | 'copied' | 'error';
interface CopyFeedbackOptions {
    copiedMs?: number;
    errorMs?: number;
}
export declare function useCopyFeedback({ copiedMs, errorMs }?: CopyFeedbackOptions): {
    copyState: CopyState;
    copyError: string;
    beginCopy: () => void;
    resetCopyFeedback: (next: "copied" | "error", error?: string) => void;
};
/** Runs an async copy with the idle → copying → copied/error feedback; ignores a click while one is in flight. */
export declare function useCopyAction(action: () => Promise<void>): {
    copyState: CopyState;
    copyError: string;
    run: () => Promise<void>;
};
export {};
//# sourceMappingURL=use-copy-feedback.d.ts.map