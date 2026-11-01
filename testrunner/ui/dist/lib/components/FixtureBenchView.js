import { SqlStatements, formatMiB, formatMs, formatSignedBytes } from "./ResourceParts.js";
import { useMemo } from "react";
import { Callout, DataTable, LabeledSparkline, MetricTile, ProcessTraceTree } from "@flanksource/clicky-ui/data";
import { Fragment, jsx, jsxs } from "react/jsx-runtime";
import { Button } from "@flanksource/clicky-ui/components";
import { cn } from "@flanksource/clicky-ui/utils";
// src/components/FixtureBenchView.tsx
var LIFECYCLE_KINDS = [
	"setup",
	"build",
	"daemon",
	"daemon-stop",
	"cleanup"
];
function benchmarkEntries(tests) {
	const phases = [];
	const fixtures = [];
	const visit = (test) => {
		if (test.fixture) {
			const entry = {
				key: test.fixture.key,
				name: test.name,
				kind: test.fixture.kind,
				status: test.fixture.state,
				duration_ms: (test.duration || 0) / 1e6,
				command_ms: test.fixture.command_ms,
				profile: test.fixture_profile,
				sql_profile: test.sql_profile,
				violations: test.fixture.violations
			};
			if (LIFECYCLE_KINDS.includes(entry.kind)) phases.push(entry);
			else fixtures.push(entry);
		}
		test.children?.forEach(visit);
	};
	tests.forEach(visit);
	return {
		phases,
		fixtures
	};
}
function toRows(entries, deltas) {
	const byKey = new Map(deltas?.map((delta) => [delta.key, delta]));
	return entries.map((entry) => {
		const delta = byKey.get(entry.key) ?? null;
		return {
			key: entry.key,
			name: entry.name,
			kind: entry.kind,
			status: entry.status,
			duration: entry.duration_ms,
			command: entry.command_ms ?? null,
			diskRead: entry.profile?.disk_io?.disk_read_bytes ?? null,
			diskWrite: entry.profile?.disk_io?.disk_write_bytes ?? null,
			sql: entry.sql_profile?.total_duration_ms ?? null,
			baseline: delta?.baseline_ms ?? null,
			change: delta?.delta_pct ?? null,
			entry,
			delta
		};
	});
}
function entryColumns(compared) {
	const mono = "font-mono tabular-nums";
	const columns = [
		{
			key: "name",
			label: "Step",
			sortable: true,
			grow: true,
			render: (_value, row) => /* @__PURE__ */ jsxs("div", { children: [
				/* @__PURE__ */ jsx("span", {
					className: "font-medium",
					children: row.name
				}),
				/* @__PURE__ */ jsx("span", {
					className: "ml-2 text-xs text-muted-foreground",
					children: row.kind
				}),
				row.entry.violations?.map((reason) => /* @__PURE__ */ jsx("div", {
					className: "text-xs text-red-700 dark:text-red-400",
					children: reason
				}, reason))
			] })
		},
		{
			key: "status",
			label: "Status",
			sortable: true,
			filterable: true
		},
		{
			key: "duration",
			label: "Duration",
			sortable: true,
			align: "right",
			cellClassName: mono,
			render: (_value, row) => formatMs(row.duration)
		},
		{
			key: "command",
			label: "Command",
			sortable: true,
			align: "right",
			cellClassName: `${mono} text-muted-foreground`,
			render: (_value, row) => row.command === null ? "—" : formatMs(row.command)
		},
		{
			key: "diskRead",
			label: "Disk read",
			sortable: true,
			align: "right",
			cellClassName: mono,
			render: (_value, row) => row.diskRead === null ? "—" : formatMiB(row.diskRead)
		},
		{
			key: "diskWrite",
			label: "Disk write",
			sortable: true,
			align: "right",
			cellClassName: mono,
			render: (_value, row) => row.diskWrite === null ? "—" : formatMiB(row.diskWrite)
		},
		{
			key: "sql",
			label: "SQL",
			sortable: true,
			align: "right",
			cellClassName: mono,
			render: (_value, row) => {
				const sql = row.entry.sql_profile;
				if (!sql) return "—";
				return /* @__PURE__ */ jsxs("div", { children: [
					formatMs(sql.total_duration_ms),
					/* @__PURE__ */ jsxs("div", {
						className: "text-xs text-muted-foreground",
						children: [
							sql.query_count,
							" ",
							sql.query_count === 1 ? "query" : "queries",
							" · ",
							sql.slow_query_count,
							" slow"
						]
					}),
					/* @__PURE__ */ jsxs("div", {
						className: "text-xs text-muted-foreground",
						children: ["Max ", formatMs(sql.max_query_ms)]
					}),
					/* @__PURE__ */ jsx(SqlStatements, { profile: sql })
				] });
			}
		}
	];
	if (compared) columns.push({
		key: "baseline",
		label: "Baseline",
		sortable: true,
		align: "right",
		cellClassName: mono,
		render: (_value, row) => row.baseline === null ? "—" : formatMs(row.baseline)
	}, {
		key: "change",
		label: "Change",
		sortable: true,
		align: "right",
		cellClassName: mono,
		render: (_value, row) => {
			const delta = row.delta;
			if (!delta) return "—";
			return /* @__PURE__ */ jsxs(Fragment, { children: [/* @__PURE__ */ jsx("span", {
				className: cn(delta.delta_ms > 0 && "text-red-700 dark:text-red-400", delta.delta_ms < 0 && "text-green-700 dark:text-green-400"),
				children: delta.delta_pct === void 0 ? "—" : `${delta.delta_pct >= 0 ? "+" : ""}${delta.delta_pct.toFixed(1)}%`
			}), delta.disk_read_bytes_delta !== void 0 && delta.disk_write_bytes_delta !== void 0 && /* @__PURE__ */ jsxs("div", {
				className: "text-xs text-muted-foreground",
				children: [
					"Disk R ",
					formatSignedBytes(delta.disk_read_bytes_delta),
					" · W ",
					formatSignedBytes(delta.disk_write_bytes_delta)
				]
			})] });
		}
	});
	return columns;
}
function EntryTable({ title, entries, deltas }) {
	const rows = useMemo(() => toRows(entries, deltas), [entries, deltas]);
	const columns = useMemo(() => entryColumns(!!deltas), [deltas]);
	if (!entries.length) return null;
	return /* @__PURE__ */ jsxs("section", {
		className: "space-y-1",
		children: [/* @__PURE__ */ jsx("h2", {
			className: "text-sm font-semibold",
			children: title
		}), /* @__PURE__ */ jsx(DataTable, {
			data: rows,
			columns,
			emptyMessage: `No ${title.toLowerCase()}`
		})]
	});
}
function processRows(sample) {
	const pids = new Set(sample.processes.map((process) => process.pid));
	return sample.processes.map((process) => ({
		key: String(process.pid),
		...process.ppid !== void 0 && pids.has(process.ppid) ? { parentKey: String(process.ppid) } : {},
		pid: process.pid,
		ppid: process.ppid ?? 0,
		name: process.command?.split(/\s+/, 1)[0] ?? "",
		command: process.command || `PID ${process.pid}`,
		status: process.status || "active",
		startedAt: sample.sampled_at,
		lastSeenAt: sample.sampled_at,
		cpuPercent: process.cpuPercent ?? 0,
		rssBytes: process.rssBytes ?? 0,
		peakCpuPercent: process.cpuPercent ?? 0,
		peakRssBytes: process.rssBytes ?? 0
	}));
}
function ProfileSection({ performance, sampleIndex, onSampleChange }) {
	const profile = performance.profile;
	const samples = profile?.samples ?? [];
	const selected = sampleIndex !== void 0 && sampleIndex < samples.length ? sampleIndex : 0;
	const sample = samples[selected];
	const nodes = useMemo(() => sample ? processRows(sample) : [], [sample]);
	return /* @__PURE__ */ jsxs("section", {
		className: "space-y-3",
		children: [/* @__PURE__ */ jsx("h2", {
			className: "text-sm font-semibold",
			children: "Process profile"
		}), !profile ? /* @__PURE__ */ jsx("p", {
			className: "text-sm text-muted-foreground",
			children: "Sampling in progress. Profile appears when the run finishes."
		}) : samples.length === 0 ? /* @__PURE__ */ jsx("p", {
			className: "text-sm text-muted-foreground",
			children: "No process samples were observed."
		}) : /* @__PURE__ */ jsxs(Fragment, { children: [
			/* @__PURE__ */ jsxs("div", {
				className: "grid grid-cols-2 gap-density-1 sm:grid-cols-5",
				children: [
					/* @__PURE__ */ jsx(MetricTile, {
						label: "Peak CPU",
						value: `${profile.peak_cpu_percent.toFixed(1)}%`
					}),
					/* @__PURE__ */ jsx(MetricTile, {
						label: "Peak RSS",
						value: formatMiB(profile.peak_rss_bytes)
					}),
					profile.disk_io && /* @__PURE__ */ jsx(MetricTile, {
						label: "Observed disk read",
						value: formatMiB(profile.disk_io.disk_read_bytes)
					}),
					profile.disk_io && /* @__PURE__ */ jsx(MetricTile, {
						label: "Observed disk write",
						value: formatMiB(profile.disk_io.disk_write_bytes)
					}),
					/* @__PURE__ */ jsx(MetricTile, {
						label: "Samples",
						value: `${samples.length} at ${profile.interval_ms} ms`
					})
				]
			}),
			/* @__PURE__ */ jsxs("div", {
				className: "grid grid-cols-2 gap-density-2",
				children: [/* @__PURE__ */ jsx(LabeledSparkline, {
					label: "CPU %",
					values: samples.map((item) => item.cpu_percent),
					format: (value) => `${value.toFixed(1)}%`
				}), /* @__PURE__ */ jsx(LabeledSparkline, {
					label: "RSS",
					values: samples.map((item) => item.rss_bytes),
					format: formatMiB
				})]
			}),
			/* @__PURE__ */ jsx("div", {
				className: "flex flex-wrap gap-1",
				"aria-label": "Profile samples",
				children: samples.map((item, index) => /* @__PURE__ */ jsx(Button, {
					type: "button",
					size: "sm",
					variant: index === selected ? "default" : "outline",
					"aria-label": `Sample ${index + 1}`,
					onClick: () => onSampleChange(index),
					children: new Date(item.sampled_at).toLocaleTimeString()
				}, item.sampled_at))
			}),
			sample && /* @__PURE__ */ jsxs("div", {
				className: "space-y-1",
				children: [/* @__PURE__ */ jsxs("div", {
					className: "text-sm text-muted-foreground",
					children: [
						"Sample ",
						selected + 1,
						" · CPU ",
						sample.cpu_percent.toFixed(1),
						"% · RSS ",
						formatMiB(sample.rss_bytes)
					]
				}), /* @__PURE__ */ jsx("div", {
					className: "h-64 rounded border border-border",
					children: /* @__PURE__ */ jsx(ProcessTraceTree, {
						className: "h-full",
						nodes,
						selectedKey: null,
						expandAll: true,
						onSelect: () => {},
						emptyLabel: "No processes were sampled"
					})
				})]
			})
		] })]
	});
}
/** A fixture benchmark or profile run: per-step timings against a baseline, and the sampled process tree. */
function FixtureBenchView({ performance, tests, error, sampleIndex, onSampleChange }) {
	const { phases, fixtures } = useMemo(() => benchmarkEntries(tests), [tests]);
	const comparison = performance.comparison;
	const profile = performance.profile;
	return /* @__PURE__ */ jsxs("div", {
		className: "mx-auto max-w-7xl space-y-5 p-density-4",
		children: [
			/* @__PURE__ */ jsxs("div", {
				className: "flex flex-wrap items-baseline gap-3",
				children: [
					/* @__PURE__ */ jsxs("h2", {
						className: "text-lg font-semibold",
						children: ["Fixture ", performance.mode === "profile" ? "profile" : "benchmark"]
					}),
					/* @__PURE__ */ jsx("span", {
						className: "text-sm text-muted-foreground",
						children: performance.status === "running" ? "Running" : `${performance.status} · ${formatMs(performance.duration_ms)}`
					}),
					performance.artifact_path && /* @__PURE__ */ jsx("span", {
						className: "break-all font-mono text-xs text-muted-foreground",
						children: performance.artifact_path
					})
				]
			}),
			error && /* @__PURE__ */ jsx("div", {
				role: "alert",
				children: /* @__PURE__ */ jsx(Callout, {
					variant: "caution",
					children: error
				})
			}),
			comparison && /* @__PURE__ */ jsx(Callout, {
				variant: "note",
				children: /* @__PURE__ */ jsxs("div", {
					className: "space-y-1 text-sm",
					children: [
						/* @__PURE__ */ jsxs("div", { children: [
							"Compared with ",
							/* @__PURE__ */ jsx("span", {
								className: "font-mono",
								children: comparison.baseline
							}),
							". Changes describe these two runs only."
						] }),
						!!comparison.added?.length && /* @__PURE__ */ jsxs("div", { children: ["Added: ", comparison.added.join(", ")] }),
						!!comparison.missing?.length && /* @__PURE__ */ jsxs("div", { children: ["Missing: ", comparison.missing.join(", ")] }),
						!!comparison.unmeasured?.length && /* @__PURE__ */ jsxs("div", { children: ["Unmeasured: ", comparison.unmeasured.join(", ")] }),
						(comparison.peak_cpu_percent_delta !== void 0 || comparison.peak_rss_bytes_delta !== void 0) && /* @__PURE__ */ jsxs("div", { children: [
							"Peak change: CPU ",
							comparison.peak_cpu_percent_delta === void 0 ? "—" : `${comparison.peak_cpu_percent_delta >= 0 ? "+" : ""}${comparison.peak_cpu_percent_delta.toFixed(1)} points`,
							"; RSS ",
							comparison.peak_rss_bytes_delta === void 0 ? "—" : formatSignedBytes(comparison.peak_rss_bytes_delta)
						] })
					]
				})
			}),
			/* @__PURE__ */ jsx(EntryTable, {
				title: "Run phases",
				entries: phases,
				deltas: comparison?.deltas
			}),
			/* @__PURE__ */ jsx(EntryTable, {
				title: "Fixture steps",
				entries: fixtures,
				deltas: comparison?.deltas
			}),
			performance.mode === "profile" || profile ? /* @__PURE__ */ jsx(ProfileSection, {
				performance,
				sampleIndex,
				onSampleChange
			}) : null,
			!phases.length && !fixtures.length && /* @__PURE__ */ jsx("p", {
				className: "text-sm text-muted-foreground",
				children: "Waiting for setup or fixture steps…"
			})
		]
	});
}
export { FixtureBenchView };

//# sourceMappingURL=FixtureBenchView.js.map