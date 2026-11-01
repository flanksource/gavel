import { isLintOnlyPhase, isLintOnlyRun, lintStatusCounts } from "../utils.js";
import { useCallback, useEffect, useRef, useState } from "react";
// src/hooks/use-test-run.ts
var DEFAULT_POLL_MS = 2e3;
var IDLE_STATUS = { running: false };
function sleep(ms, signal) {
	if (signal.aborted) return Promise.resolve();
	return new Promise((resolve) => {
		const onAbort = () => {
			clearTimeout(timer);
			resolve();
		};
		const timer = setTimeout(() => {
			signal.removeEventListener("abort", onAbort);
			resolve();
		}, ms);
		signal.addEventListener("abort", onAbort, { once: true });
	});
}
function defaultCreateEventSource(url) {
	return new EventSource(url);
}
function defaultBase() {
	if (typeof window !== "undefined") return window.__gavelBasePath || "";
	return "";
}
function hasEventSource() {
	return typeof globalThis !== "undefined" && typeof globalThis.EventSource !== "undefined";
}
function deriveStatusText(snap) {
	const status = snap.status || IDLE_STATUS;
	const tests = snap.tests || [];
	if (!status.running) {
		if (status.stopped) return status.stop_message || "Stopped by user";
		if (snap.metadata?.kind === "rerun") return "Rerun complete";
		if (isLintOnlyRun(tests, !!status.lint_run)) {
			const { done, failed } = lintStatusCounts(snap.lint);
			return `Lint run complete (${done} done · ${failed} failed)`;
		}
		return "Test run complete";
	}
	if (snap.metadata?.kind === "rerun") return `Running rerun #${snap.metadata.sequence || 1}...`;
	if (snap.performance) return "Running fixtures...";
	if (isLintOnlyPhase(tests, status.running)) {
		if (!snap.lint?.length) return "Running linters...";
		const { running, done, failed } = lintStatusCounts(snap.lint);
		return `Running linters... (${running} running · ${done} done · ${failed} failed)`;
	}
	return "Running tests...";
}
function useTestRun(options = {}) {
	const { baseUrl, basePath, runId, enabled = true, pollMs = DEFAULT_POLL_MS, forcePoll = false, createEventSource } = options;
	const root = baseUrl ?? basePath ?? defaultBase();
	const base = runId ? `${root}/${encodeURIComponent(runId)}` : root;
	const [snapshot, setSnapshot] = useState(void 0);
	const [statusText, setStatusText] = useState("Loading...");
	const [done, setDone] = useState(false);
	const [error, setError] = useState(void 0);
	const [streamToken, setStreamToken] = useState(0);
	const doneRef = useRef(false);
	const createEventSourceRef = useRef(createEventSource ?? defaultCreateEventSource);
	useEffect(() => {
		createEventSourceRef.current = createEventSource ?? defaultCreateEventSource;
	});
	const apply = useCallback((snap) => {
		const running = !!(snap.status && snap.status.running);
		setSnapshot(snap);
		setStatusText(deriveStatusText(snap));
		setError(void 0);
		if (!running) {
			doneRef.current = true;
			setDone(true);
		} else {
			doneRef.current = false;
			setDone(false);
		}
	}, []);
	useEffect(() => {
		if (!enabled) return;
		doneRef.current = false;
		setDone(false);
		if (forcePoll || !hasEventSource()) {
			const controller = new AbortController();
			(async () => {
				while (!controller.signal.aborted) {
					try {
						const res = await fetch(`${base}/api/tests`, {
							headers: { Accept: "application/json" },
							signal: controller.signal
						});
						if (controller.signal.aborted) return;
						if (res.ok) {
							const snap = await res.json();
							if (controller.signal.aborted) return;
							apply(snap);
							if (!snap.status?.running) return;
						}
					} catch {
						if (!controller.signal.aborted && !doneRef.current) setError("Connection lost — retrying...");
					}
					await sleep(pollMs, controller.signal);
				}
			})();
			return () => controller.abort();
		}
		fetch(`${base}/api/tests`, { headers: { Accept: "application/json" } }).then((r) => r.json()).then((snap) => apply(snap)).catch(() => {});
		const es = createEventSourceRef.current(`${base}/api/tests/stream`);
		const onMessage = (e) => {
			const snap = JSON.parse(e.data);
			apply(snap);
			if (!snap.status?.running) es.close();
		};
		const onDone = () => {
			doneRef.current = true;
			setDone(true);
			setSnapshot((prev) => prev ? {
				...prev,
				status: {
					...prev.status,
					running: false
				}
			} : prev);
			es.close();
		};
		es.addEventListener("message", onMessage);
		es.addEventListener("done", onDone);
		es.onerror = () => {
			if (!doneRef.current) setError("Connection lost — retrying...");
		};
		return () => {
			es.removeEventListener("message", onMessage);
			es.removeEventListener("done", onDone);
			es.close();
		};
	}, [
		base,
		enabled,
		pollMs,
		forcePoll,
		streamToken,
		apply
	]);
	const rerun = useCallback(async (body = {}) => {
		doneRef.current = false;
		setDone(false);
		setStreamToken((n) => n + 1);
		const res = await fetch(`${base}/api/rerun`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(body)
		});
		if (res.status === 409) return {
			ok: false,
			status: 409,
			conflict: true,
			message: "Another rerun is in progress"
		};
		if (!res.ok) {
			const message = (await res.text()).trim();
			setError(`Rerun failed: ${message}`);
			return {
				ok: false,
				status: res.status,
				conflict: false,
				message
			};
		}
		return {
			ok: true,
			status: res.status,
			conflict: false
		};
	}, [base]);
	const stop = useCallback(async (taskId) => {
		const payload = taskId ? {
			scope: "task",
			task_id: taskId
		} : { scope: "global" };
		const res = await fetch(`${base}/api/stop`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(payload)
		});
		if (!res.ok) {
			const message = (await res.text()).trim();
			setError(`Stop failed: ${message}`);
			return {
				ok: false,
				status: res.status,
				conflict: false,
				message
			};
		}
		return {
			ok: true,
			status: res.status,
			conflict: false
		};
	}, [base]);
	const editTest = useCallback(async (body) => {
		const res = await fetch(`${base}/api/tests/edit`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(body)
		});
		if (!res.ok) {
			const message = (await res.text()).trim();
			setError(`Test edit failed: ${message}`);
			return {
				ok: false,
				status: res.status,
				conflict: false,
				message
			};
		}
		const edited = await res.json();
		setStreamToken((n) => n + 1);
		return {
			ok: true,
			status: res.status,
			conflict: false,
			...edited.message ? { message: edited.message } : {}
		};
	}, [base]);
	const refetch = useCallback(() => setStreamToken((n) => n + 1), []);
	return {
		snapshot,
		tests: snapshot?.tests ?? [],
		lint: snapshot?.lint,
		bench: snapshot?.bench,
		status: snapshot?.status ?? IDLE_STATUS,
		statusText,
		runMeta: snapshot?.metadata,
		done,
		error,
		rerun,
		stop,
		editTest,
		refetch
	};
}
export { deriveStatusText, useTestRun };

//# sourceMappingURL=use-test-run.js.map