import { MetricTile } from "@flanksource/clicky-ui/data";
import { jsx, jsxs } from "react/jsx-runtime";
import { Section } from "@flanksource/clicky-ui/components";
import { formatBytes } from "@flanksource/clicky-ui/utils";
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
	return /* @__PURE__ */ jsx(Section, {
		title: `${statements.length} ${statements.length === 1 ? "query" : "queries"}`,
		defaultOpen: false,
		children: /* @__PURE__ */ jsx("div", {
			className: "max-h-80 space-y-2 overflow-auto",
			children: statements.map((statement, index) => /* @__PURE__ */ jsxs("div", {
				className: "rounded border border-border bg-muted/50 p-2 text-xs",
				children: [
					/* @__PURE__ */ jsxs("div", {
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
					/* @__PURE__ */ jsx("pre", {
						className: "whitespace-pre-wrap break-all font-mono",
						children: statement.sql
					}),
					statement.params.length > 0 && /* @__PURE__ */ jsxs("div", {
						className: "mt-1",
						children: [/* @__PURE__ */ jsx("span", {
							className: "text-muted-foreground",
							children: "Params"
						}), /* @__PURE__ */ jsx("ol", {
							className: "list-decimal pl-5 font-mono",
							children: statement.params.map((param, position) => /* @__PURE__ */ jsx("li", {
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
	return /* @__PURE__ */ jsxs("div", {
		className: "space-y-2",
		children: [
			profile.sample_count > 0 ? /* @__PURE__ */ jsxs("div", {
				className: "grid grid-cols-3 gap-density-1",
				children: [
					/* @__PURE__ */ jsx(MetricTile, {
						label: "Peak CPU",
						value: `${profile.peak_cpu_percent.toFixed(1)}%`
					}),
					/* @__PURE__ */ jsx(MetricTile, {
						label: "Peak RSS",
						value: formatBytes(profile.peak_rss_bytes)
					}),
					/* @__PURE__ */ jsx(MetricTile, {
						label: "Peak memory",
						value: `${profile.peak_memory_percent.toFixed(1)}%`
					})
				]
			}) : /* @__PURE__ */ jsx("p", {
				className: "text-sm text-muted-foreground",
				children: "No process samples were captured during this test."
			}),
			profile.disk_io && /* @__PURE__ */ jsxs("div", {
				className: "grid grid-cols-2 gap-density-1",
				children: [/* @__PURE__ */ jsx(MetricTile, {
					label: "Observed disk read",
					value: formatBytes(profile.disk_io.disk_read_bytes)
				}), /* @__PURE__ */ jsx(MetricTile, {
					label: "Observed disk write",
					value: formatBytes(profile.disk_io.disk_write_bytes)
				})]
			}),
			/* @__PURE__ */ jsxs("p", {
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
export { ProfileTiles, SqlStatements, formatMiB, formatMs, formatSignedBytes };

//# sourceMappingURL=ResourceParts.js.map