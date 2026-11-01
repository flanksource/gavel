const require_config = require("./config.cjs");
let _flanksource_clicky_ui_data = require("@flanksource/clicky-ui/data");
// src/routes.ts
var DEFAULT_STATUS_FILTER_TOKENS = ["!passed", "!skipped"];
var STATUS_FILTER_ALL_SENTINEL = "all";
function defaultStatusFilter() {
	return (0, _flanksource_clicky_ui_data.decodeFilterState)(DEFAULT_STATUS_FILTER_TOKENS);
}
function isDefaultStatusFilter(state) {
	const tokens = (0, _flanksource_clicky_ui_data.encodeFilterState)(state);
	if (tokens.length !== DEFAULT_STATUS_FILTER_TOKENS.length) return false;
	const want = [...DEFAULT_STATUS_FILTER_TOKENS].sort();
	const got = [...tokens].sort();
	return want.every((token, i) => token === got[i]);
}
var TAB_KEYS = [
	"tests",
	"lint",
	"files",
	"bench",
	"processes",
	"history"
];
function splitCSV(value) {
	if (!value) return [];
	return value.split(",").map((v) => v.trim()).filter(Boolean);
}
function parseView(value) {
	return value === "source" || value === "problems" ? value : "preview";
}
function parseRoute(location) {
	let pathname = location.pathname;
	if (require_config.basePath && pathname.startsWith(require_config.basePath)) pathname = pathname.slice(require_config.basePath.length);
	const trimmed = pathname.replace(/^\/+|\/+$/g, "");
	const segments = trimmed ? trimmed.split("/").map(decodeURIComponent) : [];
	let tab = "tests";
	let selectedPath = "";
	const first = segments[0];
	if (first && TAB_KEYS.includes(first)) {
		tab = first;
		selectedPath = segments.slice(1).join("/");
	}
	const params = new URLSearchParams(location.search);
	const sample = params.get("sample");
	const line = params.get("line");
	const dir = params.get("dir");
	const rawStatus = params.get("status");
	let status;
	if (rawStatus === null) status = defaultStatusFilter();
	else if (rawStatus === STATUS_FILTER_ALL_SENTINEL) status = /* @__PURE__ */ new Map();
	else status = (0, _flanksource_clicky_ui_data.decodeFilterState)(splitCSV(rawStatus));
	return {
		tab,
		selectedPath,
		filters: {
			status,
			framework: (0, _flanksource_clicky_ui_data.decodeFilterState)(splitCSV(params.get("framework")))
		},
		lintGrouping: parseLintGrouping(params.get("grouping")),
		lintFilters: {
			severity: (0, _flanksource_clicky_ui_data.decodeFilterState)(splitCSV(params.get("severity"))),
			linter: (0, _flanksource_clicky_ui_data.decodeFilterState)(splitCSV(params.get("linter")))
		},
		sampleIndex: sample !== null && /^\d+$/.test(sample) ? Number(sample) : void 0,
		hideExited: params.get("hideExited") === "1",
		...dir ? { workDir: dir } : {},
		view: parseView(params.get("view")),
		line: line !== null && /^\d+$/.test(line) && Number(line) > 0 ? Number(line) : void 0
	};
}
function buildRoute(state) {
	const segments = [state.tab];
	if (state.selectedPath) segments.push(...state.selectedPath.split("/").map(encodeURIComponent));
	const params = new URLSearchParams();
	if (state.tab === "tests") {
		if (isDefaultStatusFilter(state.filters.status)) {} else if (state.filters.status.size === 0) params.set("status", STATUS_FILTER_ALL_SENTINEL);
		else params.set("status", (0, _flanksource_clicky_ui_data.encodeFilterState)(state.filters.status).join(","));
		if (state.filters.framework.size > 0) params.set("framework", (0, _flanksource_clicky_ui_data.encodeFilterState)(state.filters.framework).join(","));
	}
	if (state.tab === "lint") {
		if (state.lintGrouping !== _flanksource_clicky_ui_data.DEFAULT_LINT_GROUPING) params.set("grouping", state.lintGrouping);
		if (state.lintFilters.severity.size > 0) params.set("severity", (0, _flanksource_clicky_ui_data.encodeFilterState)(state.lintFilters.severity).join(","));
		if (state.lintFilters.linter.size > 0) params.set("linter", (0, _flanksource_clicky_ui_data.encodeFilterState)(state.lintFilters.linter).join(","));
	}
	if (state.tab === "files") {
		if (state.workDir) params.set("dir", state.workDir);
		if (state.view && state.view !== "preview") params.set("view", state.view);
		if (state.line !== void 0) params.set("line", String(state.line));
	}
	if (state.tab === "bench" && state.sampleIndex !== void 0) params.set("sample", String(state.sampleIndex));
	if (state.tab === "processes" && state.hideExited) params.set("hideExited", "1");
	const query = params.toString();
	return `${require_config.basePath}/${segments.join("/")}${query ? `?${query}` : ""}`;
}
function buildExportRoute(state, format) {
	const [path, query = ""] = buildRoute(state).split("?", 2);
	return `${path}.${format}${query ? `?${query}` : ""}`;
}
function parseLintGrouping(value) {
	return _flanksource_clicky_ui_data.LINT_GROUPINGS.find((candidate) => candidate === value) ?? _flanksource_clicky_ui_data.DEFAULT_LINT_GROUPING;
}
exports.buildExportRoute = buildExportRoute;
exports.buildRoute = buildRoute;
exports.defaultStatusFilter = defaultStatusFilter;
exports.parseRoute = parseRoute;

//# sourceMappingURL=routes.cjs.map