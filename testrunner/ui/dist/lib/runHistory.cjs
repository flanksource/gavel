const require_config = require("./config.cjs");
// src/runHistory.ts
/**
* Fetches a page of the run history. Resolves null when the host serves no
* history (a live `gavel test --ui` run, or a page embedded in another app).
*/
async function fetchRunsPage({ before, signal } = {}) {
	const query = before ? `?before=${encodeURIComponent(before)}` : "";
	const res = await fetch(require_config.apiUrl(`/api/runs${query}`), signal ? { signal } : {});
	if (res.status === 404) return null;
	if (!res.ok) throw new Error(`GET /api/runs: ${res.status} ${(await res.text()).trim()}`);
	return await res.json();
}
/** The URL that opens a run on the tab holding its results: Lint for a lint-only run, Tests otherwise. */
function runViewHref(run) {
	return run.kind === "lint" ? `${run.href}lint` : run.href;
}
/** "Oct 22, 15:31" — the run's start in the viewer's locale, or its id when unknown. */
function runStartLabel(run) {
	if (!run.started) return run.runId;
	return new Date(run.started).toLocaleString(void 0, {
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit"
	});
}
/** A run's results in a few words: "2 failed / 40 tests · 3 lint". */
function runResultLabel(run) {
	const parts = [];
	if (run.total > 0) parts.push(run.failed > 0 ? `${run.failed} failed / ${run.total} tests` : `${run.total} tests`);
	if (run.kind !== "test") parts.push(`${run.lintViolations} lint`);
	return parts.join(" · ") || "no results";
}
function runDurationMs(run) {
	if (!run.started || !run.ended) return void 0;
	const ms = Date.parse(run.ended) - Date.parse(run.started);
	return ms >= 0 ? ms : void 0;
}
exports.fetchRunsPage = fetchRunsPage;
exports.runDurationMs = runDurationMs;
exports.runResultLabel = runResultLabel;
exports.runStartLabel = runStartLabel;
exports.runViewHref = runViewHref;

//# sourceMappingURL=runHistory.cjs.map