const require_traces = require("../traces.cjs");
let react = require("react");
let _flanksource_clicky_ui_data = require("@flanksource/clicky-ui/data");
let react_jsx_runtime = require("react/jsx-runtime");
// src/adapters/TracesPanel.tsx
function useExportFormats(mount) {
	const [byKind, setByKind] = (0, react.useState)({});
	const [error, setError] = (0, react.useState)(void 0);
	(0, react.useEffect)(() => {
		if (!mount) return;
		require_traces.fetchTraceKinds(mount).then((kinds) => setByKind(Object.fromEntries(kinds.map((kind) => [kind.name, kind.exportFormats])))).catch((e) => setError(e instanceof Error ? e.message : String(e)));
	}, [mount]);
	return {
		byKind,
		error
	};
}
/**
* One fixture's trace windows, opened in the stores of the runs they name.
* `traces` is status.traces with its mounts as the browser reaches them
* (rebaseTraces), so every link and read goes through the run's own server.
*/
function TracesPanel({ windows, traces }) {
	const formats = useExportFormats(windows.map((window) => require_traces.runTraces(traces, window.run_id).mount).find(Boolean));
	const resolve = (0, react.useCallback)((window) => require_traces.runTraces(traces, window.run_id), [traces]);
	const mountOf = (0, react.useCallback)((window) => {
		const mount = require_traces.runTraces(traces, window.run_id).mount;
		if (!mount) throw new Error(`trace store of run ${window.run_id} is not served`);
		return mount;
	}, [traces]);
	const props = (0, react.useMemo)(() => ({
		formats: (kind) => formats.byKind[kind] ?? [],
		fetchRecords: (window, signal) => require_traces.fetchTraceRecords(mountOf(window), window, signal),
		exportUrl: (window, format) => require_traces.traceExportUrl(mountOf(window), window, format),
		openUrl: (window) => require_traces.openInQueryUrl(mountOf(window), window)
	}), [formats.byKind, mountOf]);
	return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
		className: "space-y-density-3 p-density-4",
		children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.TraceWindowList, {
			windows,
			resolve,
			...props,
			...formats.error ? { error: formats.error } : {}
		})
	});
}
exports.TracesPanel = TracesPanel;

//# sourceMappingURL=TracesPanel.cjs.map