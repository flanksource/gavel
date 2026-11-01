const require_utils = require("../utils.cjs");
const require_ExportControls = require("./ExportControls.cjs");
let _flanksource_clicky_ui_data = require("@flanksource/clicky-ui/data");
let react_jsx_runtime = require("react/jsx-runtime");
let _flanksource_clicky_ui_components = require("@flanksource/clicky-ui/components");
let _flanksource_clicky_ui_icons = require("@flanksource/clicky-ui/icons");
// src/components/RunActions.tsx
var TIMEOUT_ARGS = [
	{
		key: "timeout",
		label: "global",
		title: "Global --timeout"
	},
	{
		key: "test_timeout",
		label: "per-test",
		title: "--test-timeout (per-package)"
	},
	{
		key: "lint_timeout",
		label: "per-lint",
		title: "--lint-timeout (per-linter)"
	}
];
/** The run's timeout knobs (--timeout, --test-timeout, --lint-timeout) that were set, as pills. */
function TimeoutPills({ runMeta }) {
	const set = TIMEOUT_ARGS.flatMap((arg) => {
		const value = require_utils.timeoutArgValue(runMeta?.args, arg.key);
		return value ? [{
			...arg,
			value
		}] : [];
	});
	if (set.length === 0) return null;
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", {
		className: "inline-flex items-center gap-1",
		"data-timeout-pills": "",
		children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_icons.UiWatch, { className: "text-muted-foreground" }), set.map((arg) => /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
			title: arg.title,
			children: /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(_flanksource_clicky_ui_data.Badge, {
				tone: "neutral",
				variant: "outline",
				size: "xs",
				children: [
					arg.label,
					" ",
					arg.value
				]
			})
		}, arg.key))]
	});
}
function runInfo(runMeta) {
	return [
		runMeta.pid ? `PID ${runMeta.pid}` : "",
		runMeta.frameworks?.length ? `Frameworks ${runMeta.frameworks.join(", ")}` : "",
		runMeta.command ? `Command ${runMeta.command}` : ""
	].filter(Boolean).join("\n");
}
/**
* The run-level actions of the page header: the timeout pills, the global
* Stop, and the whole view's Download and AI-prompt copy.
*/
function RunActions({ runMeta, canStop, stopping, onStop, exportState }) {
	const info = runMeta ? runInfo(runMeta) : "";
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(react_jsx_runtime.Fragment, { children: [
		info && /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
			className: "hidden text-xs text-muted-foreground lg:inline",
			title: info,
			children: "run info"
		}),
		/* @__PURE__ */ (0, react_jsx_runtime.jsx)(TimeoutPills, { runMeta }),
		canStop && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(_flanksource_clicky_ui_components.Button, {
			type: "button",
			size: "sm",
			variant: "destructive",
			disabled: stopping,
			onClick: onStop,
			title: "Stop the active run",
			children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_icons.UiStop, { className: "mr-1" }), stopping ? "Stopping..." : "Stop"]
		}),
		exportState && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_ExportControls.DownloadMenu, {
			routeState: exportState,
			title: "Download the whole result as JSON or Markdown"
		}),
		exportState && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_ExportControls.CopyAIPromptButton, { routeState: exportState })
	] });
}
exports.RunActions = RunActions;
exports.TimeoutPills = TimeoutPills;

//# sourceMappingURL=RunActions.cjs.map