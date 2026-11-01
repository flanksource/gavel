let react = require("react");
let _flanksource_clicky_ui_data = require("@flanksource/clicky-ui/data");
let react_jsx_runtime = require("react/jsx-runtime");
let _flanksource_clicky_ui_utils = require("@flanksource/clicky-ui/utils");
// src/components/BenchView.tsx
function formatNs(ns) {
	if (ns >= 1e9) return `${(ns / 1e9).toFixed(2)}s`;
	if (ns >= 1e6) return `${(ns / 1e6).toFixed(2)}ms`;
	if (ns >= 1e3) return `${(ns / 1e3).toFixed(2)}µs`;
	return `${ns.toFixed(2)}ns`;
}
function categorize(d, threshold) {
	if (d.only_in) return "only";
	if (d.significant && d.delta_pct > threshold) return "regression";
	if (d.significant && d.delta_pct < -threshold) return "improvement";
	return "neutral";
}
function toRow(delta, threshold) {
	const only = !!delta.only_in;
	return {
		name: delta.name,
		base: only ? null : delta.base_mean,
		head: only ? null : delta.head_mean,
		baseSpread: delta.base_stddev ? `±${delta.base_stddev.toFixed(1)}%` : "",
		headSpread: delta.head_stddev ? `±${delta.head_stddev.toFixed(1)}%` : "",
		delta: only ? null : delta.delta_pct,
		pValue: delta.p_value !== void 0 && delta.p_value > 0 ? delta.p_value : null,
		samples: delta.samples ?? null,
		category: categorize(delta, threshold),
		onlyIn: delta.only_in ?? null,
		threshold,
		significant: !!delta.significant
	};
}
var CATEGORY_TEXT = {
	regression: "text-red-600 dark:text-red-400",
	improvement: "text-green-600 dark:text-green-400",
	neutral: "text-muted-foreground",
	only: "text-muted-foreground"
};
var CATEGORY_BAR = {
	regression: "bg-red-500",
	improvement: "bg-green-500",
	neutral: "bg-muted-foreground/40",
	only: "bg-muted-foreground/40"
};
function DeltaBar({ row }) {
	if (row.delta === null) return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", {
		className: "text-xs italic text-muted-foreground",
		children: ["only in ", row.onlyIn]
	});
	const clamped = Math.max(-50, Math.min(50, row.delta));
	const width = Math.abs(clamped);
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
		className: "flex w-full min-w-48 items-center gap-2",
		children: [/* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
			className: "relative h-4 min-w-20 flex-1 overflow-hidden rounded bg-muted",
			children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", { className: "absolute inset-y-0 left-1/2 w-px bg-muted-foreground/60" }), /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
				className: (0, _flanksource_clicky_ui_utils.cn)("absolute inset-y-0", CATEGORY_BAR[row.category]),
				style: row.delta >= 0 ? {
					left: "50%",
					width: `${width}%`
				} : {
					right: "50%",
					width: `${width}%`
				}
			})]
		}), /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", {
			className: (0, _flanksource_clicky_ui_utils.cn)("w-16 text-right text-xs tabular-nums", CATEGORY_TEXT[row.category], row.significant && "font-semibold"),
			children: [
				row.delta >= 0 ? "+" : "",
				row.delta.toFixed(2),
				"%"
			]
		})]
	});
}
var COLUMNS = [
	{
		key: "name",
		label: "Benchmark",
		sortable: true,
		grow: true,
		cellClassName: "font-mono text-xs"
	},
	{
		key: "base",
		label: "Base",
		sortable: true,
		align: "right",
		render: (_value, row) => row.base === null ? "" : /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
			className: "text-xs tabular-nums",
			children: [formatNs(row.base), row.baseSpread && /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
				className: "text-[10px] text-muted-foreground",
				children: row.baseSpread
			})]
		})
	},
	{
		key: "head",
		label: "Head",
		sortable: true,
		align: "right",
		render: (_value, row) => row.head === null ? "" : /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
			className: "text-xs tabular-nums",
			children: [formatNs(row.head), row.headSpread && /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
				className: "text-[10px] text-muted-foreground",
				children: row.headSpread
			})]
		})
	},
	{
		key: "delta",
		label: "Δ",
		sortable: true,
		render: (_value, row) => /* @__PURE__ */ (0, react_jsx_runtime.jsx)(DeltaBar, { row })
	},
	{
		key: "pValue",
		label: "p-value",
		sortable: true,
		align: "right",
		render: (_value, row) => /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", {
			className: "text-xs tabular-nums text-muted-foreground",
			children: [row.pValue !== null ? row.pValue.toPrecision(2) : "—", row.samples !== null && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", {
				className: "ml-1",
				children: ["n=", row.samples]
			})]
		})
	}
];
/** A benchmark comparison (`gavel bench compare --ui`): per-benchmark base and head means, delta bar and significance. */
function BenchView({ bench }) {
	const rows = (0, react.useMemo)(() => (bench?.deltas ?? []).map((delta) => toRow(delta, bench.threshold)), [bench]);
	const summary = (0, react.useMemo)(() => {
		const counts = {
			regression: 0,
			improvement: 0,
			neutral: 0,
			only: 0
		};
		rows.forEach((row) => {
			counts[row.category] += 1;
		});
		return counts;
	}, [rows]);
	if (!bench) return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
		className: "p-8 text-center text-sm text-muted-foreground",
		children: [
			"No benchmark comparison loaded. Run ",
			/* @__PURE__ */ (0, react_jsx_runtime.jsx)("code", {
				className: "rounded bg-muted px-1",
				children: "gavel bench compare --ui"
			}),
			"."
		]
	});
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col gap-density-2 p-density-4",
		children: [/* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
			className: "flex flex-wrap items-center gap-4 text-sm",
			children: [
				/* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", { children: [
					/* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
						className: "font-mono",
						children: bench.base_label || "base"
					}),
					/* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
						className: "mx-2 text-muted-foreground",
						children: "→"
					}),
					/* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
						className: "font-mono",
						children: bench.head_label || "head"
					})
				] }),
				/* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
					className: "text-xs text-muted-foreground",
					children: [
						"threshold ±",
						bench.threshold.toFixed(1),
						"%"
					]
				}),
				summary.regression > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(_flanksource_clicky_ui_data.Badge, {
					tone: "danger",
					variant: "soft",
					size: "xs",
					children: [
						summary.regression,
						" regression",
						summary.regression > 1 ? "s" : ""
					]
				}),
				summary.improvement > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(_flanksource_clicky_ui_data.Badge, {
					tone: "success",
					variant: "soft",
					size: "xs",
					children: [
						summary.improvement,
						" improvement",
						summary.improvement > 1 ? "s" : ""
					]
				}),
				summary.neutral > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(_flanksource_clicky_ui_data.Badge, {
					tone: "neutral",
					variant: "soft",
					size: "xs",
					children: [summary.neutral, " unchanged"]
				}),
				/* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
					className: (0, _flanksource_clicky_ui_utils.cn)("ml-auto font-mono text-sm", bench.geomean_delta > bench.threshold ? "font-bold text-red-600 dark:text-red-400" : bench.geomean_delta < -bench.threshold ? "font-bold text-green-600 dark:text-green-400" : "text-muted-foreground"),
					children: [
						"geomean: ",
						bench.geomean_delta >= 0 ? "+" : "",
						bench.geomean_delta.toFixed(2),
						"%"
					]
				})
			]
		}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.DataTable, {
			data: rows,
			columns: COLUMNS,
			defaultSort: {
				key: "delta",
				dir: "desc"
			},
			emptyMessage: "No benchmarks were compared",
			className: "min-h-0 flex-1"
		})]
	});
}
exports.BenchView = BenchView;

//# sourceMappingURL=BenchView.cjs.map