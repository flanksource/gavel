import { runResultLabel, runStartLabel, runViewHref } from "../runHistory.js";
import { jsx, jsxs } from "react/jsx-runtime";
import { Select } from "@flanksource/clicky-ui/components";
import { UiHistory } from "@flanksource/clicky-ui/icons";
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
		label: `${runStartLabel(run)} · ${run.kind} · ${runResultLabel(run)}${history.latest.includes(run.runId) ? " (latest)" : ""}`
	}))];
	if (current !== LATEST && !history.runs.some((run) => run.runId === current)) options.splice(1, 0, {
		value: current,
		label: current
	});
	return /* @__PURE__ */ jsxs("label", {
		className: "inline-flex min-w-0 items-center gap-1 text-sm",
		"data-run-switcher": "",
		children: [
			/* @__PURE__ */ jsx(UiHistory, {
				className: "shrink-0 text-muted-foreground",
				"aria-hidden": true
			}),
			/* @__PURE__ */ jsx("span", {
				className: "sr-only",
				children: "Run"
			}),
			/* @__PURE__ */ jsx(Select, {
				className: "h-7 max-w-[22rem] truncate py-0 text-xs",
				value: current,
				options,
				onChange: (event) => {
					const value = event.target.value;
					if (value === current) return;
					const run = history.runs.find((r) => r.runId === value);
					navigate(value === LATEST ? history.latestHref : run ? runViewHref(run) : history.latestHref);
				}
			})
		]
	});
}
export { RunSwitcher };

//# sourceMappingURL=RunSwitcher.js.map