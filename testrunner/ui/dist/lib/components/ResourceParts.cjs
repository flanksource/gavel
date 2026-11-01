let _flanksource_clicky_ui_data = require("@flanksource/clicky-ui/data");
let react_jsx_runtime = require("react/jsx-runtime");
let _flanksource_clicky_ui_components = require("@flanksource/clicky-ui/components");
let _flanksource_clicky_ui_utils = require("@flanksource/clicky-ui/utils");
// src/components/ResourceParts.tsx
/** Milliseconds as "125.0 ms" or, from a second, "1.25 s". */
function formatMs(ms) {
	return ms >= 1e3 ? `${(ms / 1e3).toFixed(2)} s` : `${ms.toFixed(1)} ms`;
}
/** Bytes with a binary unit, as the fixture benchmark tables report disk and RSS. */
function formatMiB(value) {
	return value >= 1048576 ? `${(value / 1048576).toFixed(1)} MiB` : value >= 1024 ? `${(value / 1024).toFixed(1)} KiB` : `${value} B`;
}
function formatSignedBytes(value) {
	return `${value >= 0 ? "+" : "-"}${formatMiB(Math.abs(value))}`;
}
/** The statements a fixture's SQL profile recorded, behind a "N queries" disclosure. */
function SqlStatements({ profile }) {
	const statements = profile.statements;
	if (!statements?.length) return null;
	return /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_components.Section, {
		title: `${statements.length} ${statements.length === 1 ? "query" : "queries"}`,
		defaultOpen: false,
		children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
			className: "max-h-80 space-y-2 overflow-auto",
			children: statements.map((statement, index) => /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
				className: "rounded border border-border bg-muted/50 p-2 text-xs",
				children: [
					/* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
						className: "mb-1 text-muted-foreground",
						children: [
							statement.duration_ms.toFixed(1),
							" ms · ",
							statement.rows,
							" ",
							statement.rows === 1 ? "row" : "rows",
							statement.slow ? " · slow" : "",
							statement.error ? " · error" : ""
						]
					}),
					/* @__PURE__ */ (0, react_jsx_runtime.jsx)("pre", {
						className: "whitespace-pre-wrap break-all font-mono",
						children: statement.sql
					}),
					statement.params.length > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
						className: "mt-1",
						children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
							className: "text-muted-foreground",
							children: "Params"
						}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)("ol", {
							className: "list-decimal pl-5 font-mono",
							children: statement.params.map((param, position) => /* @__PURE__ */ (0, react_jsx_runtime.jsx)("li", {
								className: "break-all",
								children: param
							}, position))
						})]
					})
				]
			}, index))
		})
	});
}
/** Peak CPU/RSS/memory and observed disk of a fixture step's process tree. */
function ProfileTiles({ profile }) {
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
		className: "space-y-2",
		children: [
			profile.sample_count > 0 ? /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
				className: "grid grid-cols-3 gap-density-1",
				children: [
					/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.MetricTile, {
						label: "Peak CPU",
						value: `${profile.peak_cpu_percent.toFixed(1)}%`
					}),
					/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.MetricTile, {
						label: "Peak RSS",
						value: (0, _flanksource_clicky_ui_utils.formatBytes)(profile.peak_rss_bytes)
					}),
					/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.MetricTile, {
						label: "Peak memory",
						value: `${profile.peak_memory_percent.toFixed(1)}%`
					})
				]
			}) : /* @__PURE__ */ (0, react_jsx_runtime.jsx)("p", {
				className: "text-sm text-muted-foreground",
				children: "No process samples were captured during this test."
			}),
			profile.disk_io && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
				className: "grid grid-cols-2 gap-density-1",
				children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.MetricTile, {
					label: "Observed disk read",
					value: (0, _flanksource_clicky_ui_utils.formatBytes)(profile.disk_io.disk_read_bytes)
				}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.MetricTile, {
					label: "Observed disk write",
					value: (0, _flanksource_clicky_ui_utils.formatBytes)(profile.disk_io.disk_write_bytes)
				})]
			}),
			/* @__PURE__ */ (0, react_jsx_runtime.jsxs)("p", {
				className: "text-xs text-muted-foreground",
				children: [
					profile.sample_count,
					" samples · ",
					profile.scope === "process_tree" ? `Command process tree · PID ${profile.pid}` : "Run process tree during this test"
				]
			})
		]
	});
}
exports.ProfileTiles = ProfileTiles;
exports.SqlStatements = SqlStatements;
exports.formatMiB = formatMiB;
exports.formatMs = formatMs;
exports.formatSignedBytes = formatSignedBytes;

//# sourceMappingURL=ResourceParts.cjs.map