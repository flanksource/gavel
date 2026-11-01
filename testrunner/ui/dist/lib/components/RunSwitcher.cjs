const require_runHistory = require("../runHistory.cjs");
let react_jsx_runtime = require("react/jsx-runtime");
let _flanksource_clicky_ui_components = require("@flanksource/clicky-ui/components");
let _flanksource_clicky_ui_icons = require("@flanksource/clicky-ui/icons");
// src/components/RunSwitcher.tsx
var LATEST = "";
/** Switches the page to another recorded run, or back to the latest results. */
function RunSwitcher({ history, navigate }) {
	const current = history.current ?? LATEST;
	const options = [{
		value: LATEST,
		label: "Latest results"
	}, ...history.runs.map((run) => ({
		value: run.runId,
		label: `${require_runHistory.runStartLabel(run)} · ${run.kind} · ${require_runHistory.runResultLabel(run)}${history.latest.includes(run.runId) ? " (latest)" : ""}`
	}))];
	if (current !== LATEST && !history.runs.some((run) => run.runId === current)) options.splice(1, 0, {
		value: current,
		label: current
	});
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("label", {
		className: "inline-flex min-w-0 items-center gap-1 text-sm",
		"data-run-switcher": "",
		children: [
			/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_icons.UiHistory, {
				className: "shrink-0 text-muted-foreground",
				"aria-hidden": true
			}),
			/* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
				className: "sr-only",
				children: "Run"
			}),
			/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_components.Select, {
				className: "h-7 max-w-[22rem] truncate py-0 text-xs",
				value: current,
				options,
				onChange: (event) => {
					const value = event.target.value;
					if (value === current) return;
					const run = history.runs.find((r) => r.runId === value);
					navigate(value === LATEST ? history.latestHref : run ? require_runHistory.runViewHref(run) : history.latestHref);
				}
			})
		]
	});
}
exports.RunSwitcher = RunSwitcher;

//# sourceMappingURL=RunSwitcher.cjs.map