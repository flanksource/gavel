const require_config = require("../config.cjs");
const require_traces = require("../traces.cjs");
let react = require("react");
// src/hooks/use-process-trace.ts
var POLL_INTERVAL_MS = 1e3;
var SAMPLE_WINDOW = 2e3;
var PROCESS_TRACE_ENDED = "The run has ended, so its process trace can no longer be checkpointed";
async function readStream(mount, ref) {
	try {
		return await require_traces.fetchTraceRecords(mount, ref);
	} catch (e) {
		if (e instanceof require_traces.TraceStreamNotFound) return [];
		throw e;
	}
}
async function readRecentSamples(mount, ref) {
	const latest = await require_traces.fetchLatestSeq(mount, ref);
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
	const [value, setValue] = (0, react.useState)(null);
	const [error, setError] = (0, react.useState)(void 0);
	const loadRef = (0, react.useRef)(load);
	(0, react.useEffect)(() => {
		loadRef.current = load;
	});
	const reload = (0, react.useCallback)(async () => {
		try {
			setValue(await loadRef.current());
			setError(void 0);
		} catch (e) {
			setError(message(e));
		}
	}, []);
	(0, react.useEffect)(() => {
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
	const [busy, setBusy] = (0, react.useState)(false);
	const [error, setError] = (0, react.useState)(void 0);
	return {
		busy,
		error,
		checkpoint: (0, react.useCallback)(async (params) => {
			setBusy(true);
			setError(void 0);
			try {
				if (!mount || !session) {
					setError(PROCESS_TRACE_ENDED);
					return;
				}
				await require_traces.checkpointSession({
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
	const scope = base ?? require_config.apiBase();
	const traces = (0, react.useMemo)(() => require_traces.rebaseTraces(reported, scope), [reported, scope]);
	const mount = require_traces.processMount(traces);
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
	const collectStack = (0, react.useCallback)((key, transport) => stackCheckpoint.checkpoint({
		key,
		transport,
		reason: "manual"
	}), [stackCheckpoint.checkpoint]);
	const refreshFacets = (0, react.useCallback)((key) => sampleCheckpoint.checkpoint({ key }), [sampleCheckpoint.checkpoint]);
	return {
		nodes: (0, react.useMemo)(() => tree.value ?? [], [tree.value]),
		samples: (0, react.useMemo)(() => sampled.value ?? [], [sampled.value]),
		stacks: (0, react.useMemo)(() => stacked.value ?? [], [stacked.value]),
		goprof: (0, react.useMemo)(() => profiled.value ?? [], [profiled.value]),
		...mount && refs?.goprof ? { goprofExportUrl: require_traces.traceExportUrl(mount, refs.goprof, "pprof") } : {},
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
exports.PROCESS_TRACE_ENDED = PROCESS_TRACE_ENDED;
exports.SAMPLE_WINDOW = SAMPLE_WINDOW;
exports.useProcessTrace = useProcessTrace;

//# sourceMappingURL=use-process-trace.cjs.map