import { useCallback, useEffect, useRef, useState } from "react";
// src/hooks/use-copy-feedback.ts
function useCopyFeedback({ copiedMs = 2e3, errorMs = 3e3 } = {}) {
	const [copyState, setCopyState] = useState("idle");
	const [copyError, setCopyError] = useState("");
	const timer = useRef(null);
	const clearTimer = useCallback(() => {
		if (timer.current !== null) {
			window.clearTimeout(timer.current);
			timer.current = null;
		}
	}, []);
	useEffect(() => clearTimer, [clearTimer]);
	return {
		copyState,
		copyError,
		beginCopy: useCallback(() => {
			setCopyState("copying");
			setCopyError("");
			clearTimer();
		}, [clearTimer]),
		resetCopyFeedback: useCallback((next, error = "") => {
			setCopyState(next);
			setCopyError(error);
			clearTimer();
			timer.current = window.setTimeout(() => {
				timer.current = null;
				setCopyState("idle");
				setCopyError("");
			}, next === "copied" ? copiedMs : errorMs);
		}, [
			clearTimer,
			copiedMs,
			errorMs
		])
	};
}
/** Runs an async copy with the idle → copying → copied/error feedback; ignores a click while one is in flight. */
function useCopyAction(action) {
	const { copyState, copyError, beginCopy, resetCopyFeedback } = useCopyFeedback();
	return {
		copyState,
		copyError,
		run: useCallback(async () => {
			if (copyState === "copying") return;
			beginCopy();
			try {
				await action();
				resetCopyFeedback("copied");
			} catch (e) {
				resetCopyFeedback("error", e instanceof Error ? e.message : "Copy failed");
			}
		}, [
			action,
			copyState,
			beginCopy,
			resetCopyFeedback
		])
	};
}
export { useCopyAction, useCopyFeedback };

//# sourceMappingURL=use-copy-feedback.js.map