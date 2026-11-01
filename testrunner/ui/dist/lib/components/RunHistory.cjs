const require_runHistory = require("../runHistory.cjs");
let react = require("react");
let _flanksource_clicky_ui_data = require("@flanksource/clicky-ui/data");
let react_jsx_runtime = require("react/jsx-runtime");
let _flanksource_clicky_ui_utils = require("@flanksource/clicky-ui/utils");
// src/components/RunHistory.tsx
function ResultBadges({ run }) {
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", {
		className: "inline-flex flex-wrap items-center gap-1",
		children: [
			run.total > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.Badge, {
				tone: run.failed > 0 ? "danger" : "success",
				variant: "outline",
				size: "xs",
				children: run.failed > 0 ? `${run.failed} failed` : `${run.passed} passed`
			}),
			run.total > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", {
				className: "text-xs text-muted-foreground",
				children: ["of ", run.total]
			}),
			run.kind !== "test" && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(_flanksource_clicky_ui_data.Badge, {
				tone: run.lintViolations > 0 ? "warning" : "neutral",
				variant: "outline",
				size: "xs",
				children: [run.lintViolations, " lint"]
			})
		]
	});
}
/** Every recorded run, newest first; picking one opens it. */
function RunHistory({ history, navigate }) {
	const columns = (0, react.useMemo)(() => [
		{
			key: "started",
			label: "Started",
			kind: "timestamp",
			sortable: true,
			render: (_value, run) => /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", {
				className: "inline-flex items-center gap-1.5",
				children: [
					/* @__PURE__ */ (0, react_jsx_runtime.jsx)("a", {
						href: require_runHistory.runViewHref(run),
						className: "text-primary no-underline hover:underline",
						onClick: (event) => event.stopPropagation(),
						children: run.started ? new Date(run.started).toLocaleString() : run.runId
					}),
					history.current === run.runId && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.Badge, {
						tone: "info",
						variant: "soft",
						size: "xs",
						children: "viewing"
					}),
					!history.current && history.latest.includes(run.runId) && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.Badge, {
						tone: "info",
						variant: "soft",
						size: "xs",
						children: "latest"
					})
				]
			})
		},
		{
			key: "kind",
			label: "Kind",
			filterable: true,
			shrink: true
		},
		{
			key: "results",
			label: "Results",
			accessor: (run) => run.failed + run.lintViolations,
			sortable: true,
			render: (_value, run) => /* @__PURE__ */ (0, react_jsx_runtime.jsx)(ResultBadges, { run })
		},
		{
			key: "duration",
			label: "Duration",
			accessor: (run) => require_runHistory.runDurationMs(run) ?? null,
			sortable: true,
			align: "right",
			shrink: true,
			render: (value) => typeof value === "number" ? (0, _flanksource_clicky_ui_utils.formatDuration)(value) : ""
		},
		{
			key: "frameworks",
			label: "Frameworks",
			kind: "tags",
			accessor: (run) => run.frameworks ?? []
		},
		{
			key: "sha",
			label: "Commit",
			shrink: true,
			render: (value) => /* @__PURE__ */ (0, react_jsx_runtime.jsx)("code", {
				className: "text-xs",
				children: String(value ?? "").slice(0, 8)
			})
		}
	], [history.current, history.latest]);
	return /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.DataTable, {
		data: history.runs,
		columns,
		getRowId: (run) => run.runId,
		onRowClick: (run) => navigate(require_runHistory.runViewHref(run)),
		getRowClassName: (run) => history.current === run.runId ? "bg-primary/5" : void 0,
		autoFilter: true,
		infinite: {
			hasMore: history.hasMore,
			loading: history.loading,
			onLoadMore: history.loadMore
		},
		loading: history.loading && history.runs.length === 0,
		loadingMessage: "Loading run history",
		error: history.error,
		emptyMessage: "No runs recorded in .gavel yet"
	});
}
exports.RunHistory = RunHistory;

//# sourceMappingURL=RunHistory.cjs.map