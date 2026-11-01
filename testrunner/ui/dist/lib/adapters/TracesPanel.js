import { fetchTraceKinds, fetchTraceRecords, openInQueryUrl, runTraces, traceExportUrl } from "../traces.js";
import { useCallback, useEffect, useMemo, useState } from "react";
import { TraceWindowList } from "@flanksource/clicky-ui/data";
import { jsx } from "react/jsx-runtime";
// src/adapters/TracesPanel.tsx
function useExportFormats(mount) {
	const [byKind, setByKind] = useState({});
	const [error, setError] = useState(void 0);
	useEffect(() => {
		if (!mount) return;
		fetchTraceKinds(mount).then((kinds) => setByKind(Object.fromEntries(kinds.map((kind) => [kind.name, kind.exportFormats])))).catch((e) => setError(e instanceof Error ? e.message : String(e)));
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
	const formats = useExportFormats(windows.map((window) => runTraces(traces, window.run_id).mount).find(Boolean));
	const resolve = useCallback((window) => runTraces(traces, window.run_id), [traces]);
	const mountOf = useCallback((window) => {
		const mount = runTraces(traces, window.run_id).mount;
		if (!mount) throw new Error(`trace store of run ${window.run_id} is not served`);
		return mount;
	}, [traces]);
	const props = useMemo(() => ({
		formats: (kind) => formats.byKind[kind] ?? [],
		fetchRecords: (window, signal) => fetchTraceRecords(mountOf(window), window, signal),
		exportUrl: (window, format) => traceExportUrl(mountOf(window), window, format),
		openUrl: (window) => openInQueryUrl(mountOf(window), window)
	}), [formats.byKind, mountOf]);
	return /* @__PURE__ */ jsx("div", {
		className: "space-y-density-3 p-density-4",
		children: /* @__PURE__ */ jsx(TraceWindowList, {
			windows,
			resolve,
			...props,
			...formats.error ? { error: formats.error } : {}
		})
	});
}
export { TracesPanel };

//# sourceMappingURL=TracesPanel.js.map