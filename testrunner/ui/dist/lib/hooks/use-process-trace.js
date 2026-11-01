import { apiBase } from "../config.js";
import { TraceStreamNotFound, checkpointSession, fetchLatestSeq, fetchTraceRecords, processMount, rebaseTraces, traceExportUrl } from "../traces.js";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
// src/hooks/use-process-trace.ts
var POLL_INTERVAL_MS = 1e3;
var SAMPLE_WINDOW = 2e3;
var PROCESS_TRACE_ENDED = "The run has ended, so its process trace can no longer be checkpointed";
async function readStream(mount, ref) {
	try {
		return await fetchTraceRecords(mount, ref);
	} catch (e) {
		if (e instanceof TraceStreamNotFound) return [];
		throw e;
	}
}
async function readRecentSamples(mount, ref) {
	const latest = await fetchLatestSeq(mount, ref);
	if (latest === 0) return [];
	return readStream(mount, {
		...ref,
		afterSeq: Math.max(0, latest - SAMPLE_WINDOW)
	});
}
function message(e) {
	return e instanceof Error ? e.message : String(e);
}
function usePolled(enabled, poll, load, key) {
	const [value, setValue] = useState(null);
	const [error, setError] = useState(void 0);
	const loadRef = useRef(load);
	useEffect(() => {
		loadRef.current = load;
	});
	const reload = useCallback(async () => {
		try {
			setValue(await loadRef.current());
			setError(void 0);
		} catch (e) {
			setError(message(e));
		}
	}, []);
	useEffect(() => {
		if (!enabled) return;
		let cancelled = false;
		const run = () => {
			loadRef.current().then((next) => {
				if (cancelled) return;
				setValue(next);
				setError(void 0);
			}).catch((e) => {
				if (!cancelled) setError(message(e));
			});
		};
		run();
		const timer = poll ? window.setInterval(run, POLL_INTERVAL_MS) : void 0;
		return () => {
			cancelled = true;
			if (timer !== void 0) window.clearInterval(timer);
		};
	}, [
		enabled,
		poll,
		key
	]);
	return {
		value,
		error,
		reload
	};
}
function useCheckpoint(mount, session, reload) {
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState(void 0);
	return {
		busy,
		error,
		checkpoint: useCallback(async (params) => {
			setBusy(true);
			setError(void 0);
			try {
				if (!mount || !session) {
					setError(PROCESS_TRACE_ENDED);
					return;
				}
				await checkpointSession({
					mount,
					session,
					params
				});
				await reload();
			} catch (e) {
				setError(message(e));
			} finally {
				setBusy(false);
			}
		}, [
			mount,
			session,
			reload
		])
	};
}
function useProcessTrace({ traces: reported, done, selectedKey, base }) {
	const scope = base ?? apiBase();
	const traces = useMemo(() => rebaseTraces(reported, scope), [reported, scope]);
	const mount = processMount(traces);
	const refs = traces?.process;
	const live = !done;
	const enabled = !!mount && !!refs;
	const sessions = live ? reported?.sessions : void 0;
	const tree = usePolled(enabled, live, () => readStream(mount, refs.tree), `${mount}|${refs?.tree.stream}`);
	const sampled = usePolled(enabled, live, () => readRecentSamples(mount, refs.samples), `${mount}|${refs?.samples.stream}`);
	const stacked = usePolled(enabled && !!selectedKey, live, () => readStream(mount, refs.stacks), `${mount}|${refs?.stacks.stream}`);
	const profiled = usePolled(enabled && !!selectedKey && !!refs?.goprof, live, () => readStream(mount, refs.goprof), `${mount}|${refs?.goprof?.stream}`);
	const stackCheckpoint = useCheckpoint(mount, sessions?.stacks, stacked.reload);
	const sampleCheckpoint = useCheckpoint(mount, sessions?.samples, sampled.reload);
	const collectStack = useCallback((key, transport) => stackCheckpoint.checkpoint({
		key,
		transport,
		reason: "manual"
	}), [stackCheckpoint.checkpoint]);
	const refreshFacets = useCallback((key) => sampleCheckpoint.checkpoint({ key }), [sampleCheckpoint.checkpoint]);
	return {
		nodes: useMemo(() => tree.value ?? [], [tree.value]),
		samples: useMemo(() => sampled.value ?? [], [sampled.value]),
		stacks: useMemo(() => stacked.value ?? [], [stacked.value]),
		goprof: useMemo(() => profiled.value ?? [], [profiled.value]),
		...mount && refs?.goprof ? { goprofExportUrl: traceExportUrl(mount, refs.goprof, "pprof") } : {},
		loading: enabled && tree.value === null && !tree.error,
		error: tree.error ?? sampled.error ?? stacked.error ?? profiled.error,
		canCollect: enabled && !!sessions,
		collecting: stackCheckpoint.busy,
		...stackCheckpoint.error ? { collectError: stackCheckpoint.error } : {},
		collectStack,
		refreshing: sampleCheckpoint.busy,
		...sampleCheckpoint.error ? { refreshError: sampleCheckpoint.error } : {},
		refreshFacets
	};
}
export { PROCESS_TRACE_ENDED, SAMPLE_WINDOW, useProcessTrace };

//# sourceMappingURL=use-process-trace.js.map