import { Badge } from "@flanksource/clicky-ui/data";
import { jsx, jsxs } from "react/jsx-runtime";
import { CopyButton } from "@flanksource/clicky-ui/components";
import { SessionInspector, usePromptRunStream } from "@flanksource/clicky-ui/ai";
// src/PromptRunView.tsx
var STATUS_TONE = {
	running: "info",
	succeeded: "success",
	failed: "danger",
	idle: "neutral"
};
function CopyableId({ label, value }) {
	return /* @__PURE__ */ jsxs("span", {
		className: "inline-flex items-center gap-1",
		children: [
			/* @__PURE__ */ jsx("span", {
				className: "text-muted-foreground",
				children: label
			}),
			/* @__PURE__ */ jsx("code", {
				className: "select-all font-mono",
				children: value
			}),
			/* @__PURE__ */ jsx(CopyButton, {
				value,
				label: `Copy ${label.toLowerCase()} id`
			})
		]
	});
}
function RunHeader({ run, result }) {
	const engine = [
		run.run?.model ?? run.summary?.model,
		run.run?.provider ?? run.summary?.provider,
		run.run?.mode ?? run.summary?.mode
	].filter(Boolean).join(" · ");
	const sessionId = run.sessionId ?? result.session_id;
	return /* @__PURE__ */ jsxs("div", {
		role: "group",
		"aria-label": "Prompt run",
		className: "flex flex-wrap items-center gap-x-3 gap-y-1 text-xs",
		children: [
			/* @__PURE__ */ jsx(Badge, {
				tone: STATUS_TONE[run.status],
				variant: "soft",
				size: "sm",
				children: run.status
			}),
			engine && /* @__PURE__ */ jsx("span", {
				className: "text-muted-foreground",
				children: engine
			}),
			result.run_id && /* @__PURE__ */ jsx(CopyableId, {
				label: "Run",
				value: result.run_id
			}),
			sessionId && /* @__PURE__ */ jsx(CopyableId, {
				label: "Session",
				value: sessionId
			}),
			run.summary?.duration && /* @__PURE__ */ jsxs("span", {
				className: "text-muted-foreground",
				children: ["Elapsed ", run.summary.duration]
			})
		]
	});
}
/**
* Follows the Captain prompt run behind a prompt lint result and shows its
* transcript as it streams, under a header with the run's status, model,
* ids and, once it finishes, its duration. The transcript is the run's own
* `entry` frames, so it shows whether or not Captain records a session. The
* run's findings reach the lint tree through the run snapshot.
*/
function PromptRunView({ result, baseUrl }) {
	const run = usePromptRunStream(result.run_id, { baseUrl });
	const lifecycleStatus = run.status === "running" ? "running" : run.status === "succeeded" ? "succeeded" : "failed";
	return /* @__PURE__ */ jsxs("div", {
		className: "flex h-[32rem] min-h-0 flex-col gap-2",
		children: [
			/* @__PURE__ */ jsx(RunHeader, {
				run,
				result
			}),
			run.status === "failed" && /* @__PURE__ */ jsx("p", {
				role: "alert",
				className: "text-sm text-red-600",
				children: run.error
			}),
			run.entries.length === 0 ? /* @__PURE__ */ jsx("p", {
				role: "status",
				className: "text-sm text-muted-foreground",
				children: run.status === "running" ? "Waiting for the prompt run to start…" : "The prompt run produced no transcript."
			}) : /* @__PURE__ */ jsx(SessionInspector, {
				session: {
					...run.sessionId ? { id: run.sessionId } : {},
					lifecycleStatus,
					messages: run.entries
				},
				className: "min-h-0 flex-1 text-xs",
				layout: "compact",
				follow: false
			})
		]
	});
}
export { PromptRunView as default };

//# sourceMappingURL=PromptRunView.js.map