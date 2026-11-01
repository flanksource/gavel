import { timeoutArgValue } from "../utils.js";
import { CopyAIPromptButton, DownloadMenu } from "./ExportControls.js";
import { Badge } from "@flanksource/clicky-ui/data";
import { Fragment, jsx, jsxs } from "react/jsx-runtime";
import { Button } from "@flanksource/clicky-ui/components";
import { UiStop, UiWatch } from "@flanksource/clicky-ui/icons";
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
		const value = timeoutArgValue(runMeta?.args, arg.key);
		return value ? [{
			...arg,
			value
		}] : [];
	});
	if (set.length === 0) return null;
	return /* @__PURE__ */ jsxs("span", {
		className: "inline-flex items-center gap-1",
		"data-timeout-pills": "",
		children: [/* @__PURE__ */ jsx(UiWatch, { className: "text-muted-foreground" }), set.map((arg) => /* @__PURE__ */ jsx("span", {
			title: arg.title,
			children: /* @__PURE__ */ jsxs(Badge, {
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
	return /* @__PURE__ */ jsxs(Fragment, { children: [
		info && /* @__PURE__ */ jsx("span", {
			className: "hidden text-xs text-muted-foreground lg:inline",
			title: info,
			children: "run info"
		}),
		/* @__PURE__ */ jsx(TimeoutPills, { runMeta }),
		canStop && /* @__PURE__ */ jsxs(Button, {
			type: "button",
			size: "sm",
			variant: "destructive",
			disabled: stopping,
			onClick: onStop,
			title: "Stop the active run",
			children: [/* @__PURE__ */ jsx(UiStop, { className: "mr-1" }), stopping ? "Stopping..." : "Stop"]
		}),
		exportState && /* @__PURE__ */ jsx(DownloadMenu, {
			routeState: exportState,
			title: "Download the whole result as JSON or Markdown"
		}),
		exportState && /* @__PURE__ */ jsx(CopyAIPromptButton, { routeState: exportState })
	] });
}
export { RunActions, TimeoutPills };

//# sourceMappingURL=RunActions.js.map