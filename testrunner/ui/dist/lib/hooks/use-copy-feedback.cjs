let react = require("react");
// src/hooks/use-copy-feedback.ts
function useCopyFeedback({ copiedMs = 2e3, errorMs = 3e3 } = {}) {
	const [copyState, setCopyState] = (0, react.useState)("idle");
	const [copyError, setCopyError] = (0, react.useState)("");
	const timer = (0, react.useRef)(null);
	const clearTimer = (0, react.useCallback)(() => {
		if (timer.current !== null) {
			window.clearTimeout(timer.current);
			timer.current = null;
		}
	}, []);
	(0, react.useEffect)(() => clearTimer, [clearTimer]);
	return {
		copyState,
		copyError,
		beginCopy: (0, react.useCallback)(() => {
			setCopyState("copying");
			setCopyError("");
			clearTimer();
		}, [clearTimer]),
		resetCopyFeedback: (0, react.useCallback)((next, error = "") => {
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
		run: (0, react.useCallback)(async () => {
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
exports.useCopyAction = useCopyAction;
exports.useCopyFeedback = useCopyFeedback;

//# sourceMappingURL=use-copy-feedback.cjs.map