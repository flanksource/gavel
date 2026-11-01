// src/traces.ts
var TraceStreamNotFound = class extends Error {};
/** Run runId's store as status reaches it: its mount, or why it has none. */
function runTraces(status, runId) {
	const run = status?.runs[runId];
	if (run) return run;
	return { unavailable: `The trace store of run ${runId || "(unnamed)"} is not served by this view.` };
}
/** The mount of the store holding the Processes tab's capture; undefined when not served. */
function processMount(status) {
	if (!status?.process_run || !status.process) return void 0;
	return status.runs[status.process_run]?.mount;
}
/**
* status with every mount under base: a host that serves the runner's server
* under a prefix of its own (the page's base path, a MultiServer run id, or
* pr/ui's /api/project-runs/<id>) rebases the mounts the server reports
* relative to itself.
*/
function rebaseTraces(status, base) {
	if (!status) return void 0;
	if (!base) return status;
	const runs = Object.fromEntries(Object.entries(status.runs).map(([id, run]) => [id, run.mount ? {
		...run,
		mount: `${base}${run.mount}`
	} : run]));
	return {
		...status,
		runs
	};
}
function queryUrl(mount, path) {
	return `${mount}${path}`;
}
function windowQuery(ref, extra = {}) {
	const params = new URLSearchParams({ stream: ref.stream });
	if (ref.afterSeq) params.set("afterSeq", String(ref.afterSeq));
	if (ref.toSeq) params.set("toSeq", String(ref.toSeq));
	for (const [key, value] of Object.entries(extra)) params.set(key, value);
	return params.toString();
}
/** The download of ref's window in one of its kind's export formats. */
function traceExportUrl(mount, ref, format) {
	return queryUrl(mount, `/api/v1/traces/${encodeURIComponent(ref.kind)}/export?${windowQuery(ref, { format })}`);
}
/** The query app's page for ref's window: its kind's result profile, filtered to the window. */
function openInQueryUrl(mount, ref) {
	return queryUrl(mount, `/traces/${encodeURIComponent(ref.kind)}?${windowQuery(ref)}`);
}
async function failure(res, what) {
	const text = (await res.text()).trim();
	return /* @__PURE__ */ new Error(`${what} failed (${res.status})${text ? `: ${text}` : ""}`);
}
/** Parses an NDJSON body: one record per non-empty line. */
function parseNdjson(body) {
	return body.split("\n").filter((line) => line.trim() !== "").map((line) => JSON.parse(line));
}
/**
* The records of ref's window as its stream holds them: the NDJSON export.
* The result profile's JSON rows leave out the columns a kind hides (a
* process's key and stack, a terminal event's data), which the viewers need.
*/
async function fetchTraceRecords(mount, ref, signal) {
	const res = await fetch(traceExportUrl(mount, ref, "ndjson"), signal ? { signal } : void 0);
	if (res.status === 404) throw new TraceStreamNotFound(`${ref.kind} stream ${ref.stream} has no records`);
	if (!res.ok) throw await failure(res, `export ${ref.kind} window`);
	return parseNdjson(await res.text());
}
/**
* The id the query server surfaces a kind's result profile (`traces/<kind>`)
* under: commons-db `profiles.SurfaceKey`, `profile-` plus the slugged name.
*/
function traceProfileId(kind) {
	return `profile-${`traces/${kind}`.toLowerCase().replace(/[ _/.-]/g, "-").replace(/[^a-z0-9-]/g, "").replace(/^-+|-+$/g, "")}`;
}
/** The highest seq ref's stream has committed, read from its result profile; 0 when it has none. */
async function fetchLatestSeq(mount, ref) {
	const query = new URLSearchParams({
		stream: ref.stream,
		sort: "seq",
		order: "desc",
		limit: "1"
	});
	const res = await fetch(queryUrl(mount, `/api/v1/profile/${traceProfileId(ref.kind)}?${query}`), { headers: { Accept: "application/json" } });
	if (!res.ok) throw await failure(res, `read ${ref.kind} stream`);
	return (await res.json())[0]?.seq ?? 0;
}
/** The trace kinds the store serves, with the formats each exports in. */
async function fetchTraceKinds(mount) {
	const res = await fetch(queryUrl(mount, "/api/v1/traces/kinds"), { headers: { Accept: "application/json" } });
	if (!res.ok) throw await failure(res, "list trace kinds");
	return res.json();
}
/** Has the running session take a reading now, as its kind's checkpoint params describe it. */
async function checkpointSession({ mount, session, params }) {
	const res = await fetch(queryUrl(mount, `/api/v1/sessions/${encodeURIComponent(session)}/checkpoint`), {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(params)
	});
	if (!res.ok) throw await failure(res, "checkpoint");
}
exports.TraceStreamNotFound = TraceStreamNotFound;
exports.checkpointSession = checkpointSession;
exports.failure = failure;
exports.fetchLatestSeq = fetchLatestSeq;
exports.fetchTraceKinds = fetchTraceKinds;
exports.fetchTraceRecords = fetchTraceRecords;
exports.openInQueryUrl = openInQueryUrl;
exports.parseNdjson = parseNdjson;
exports.processMount = processMount;
exports.queryUrl = queryUrl;
exports.rebaseTraces = rebaseTraces;
exports.runTraces = runTraces;
exports.traceExportUrl = traceExportUrl;
exports.traceProfileId = traceProfileId;

//# sourceMappingURL=traces.cjs.map