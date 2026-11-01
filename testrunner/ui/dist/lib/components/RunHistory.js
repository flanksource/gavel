import { runDurationMs, runViewHref } from "../runHistory.js";
import { useMemo } from "react";
import { Badge, DataTable } from "@flanksource/clicky-ui/data";
import { jsx, jsxs } from "react/jsx-runtime";
import { formatDuration } from "@flanksource/clicky-ui/utils";
// src/components/RunHistory.tsx
function ResultBadges({ run }) {
	return /* @__PURE__ */ jsxs("span", {
		className: "inline-flex flex-wrap items-center gap-1",
		children: [
			run.total > 0 && /* @__PURE__ */ jsx(Badge, {
				tone: run.failed > 0 ? "danger" : "success",
				variant: "outline",
				size: "xs",
				children: run.failed > 0 ? `${run.failed} failed` : `${run.passed} passed`
			}),
			run.total > 0 && /* @__PURE__ */ jsxs("span", {
				className: "text-xs text-muted-foreground",
				children: ["of ", run.total]
			}),
			run.kind !== "test" && /* @__PURE__ */ jsxs(Badge, {
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
	const columns = useMemo(() => [
		{
			key: "started",
			label: "Started",
			kind: "timestamp",
			sortable: true,
			render: (_value, run) => /* @__PURE__ */ jsxs("span", {
				className: "inline-flex items-center gap-1.5",
				children: [
					/* @__PURE__ */ jsx("a", {
						href: runViewHref(run),
						className: "text-primary no-underline hover:underline",
						onClick: (event) => event.stopPropagation(),
						children: run.started ? new Date(run.started).toLocaleString() : run.runId
					}),
					history.current === run.runId && /* @__PURE__ */ jsx(Badge, {
						tone: "info",
						variant: "soft",
						size: "xs",
						children: "viewing"
					}),
					!history.current && history.latest.includes(run.runId) && /* @__PURE__ */ jsx(Badge, {
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
			render: (_value, run) => /* @__PURE__ */ jsx(ResultBadges, { run })
		},
		{
			key: "duration",
			label: "Duration",
			accessor: (run) => runDurationMs(run) ?? null,
			sortable: true,
			align: "right",
			shrink: true,
			render: (value) => typeof value === "number" ? formatDuration(value) : ""
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
			render: (value) => /* @__PURE__ */ jsx("code", {
				className: "text-xs",
				children: String(value ?? "").slice(0, 8)
			})
		}
	], [history.current, history.latest]);
	return /* @__PURE__ */ jsx(DataTable, {
		data: history.runs,
		columns,
		getRowId: (run) => run.runId,
		onRowClick: (run) => navigate(runViewHref(run)),
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
export { RunHistory };

//# sourceMappingURL=RunHistory.js.map