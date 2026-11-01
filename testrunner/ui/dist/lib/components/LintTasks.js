import { liveTaskGroups } from "../utils.js";
import { Icon } from "@flanksource/clicky-ui/data";
import { jsx, jsxs } from "react/jsx-runtime";
import { UiCircleOutline, UiLoader } from "@flanksource/clicky-ui/icons";
// src/components/LintTasks.tsx
function statusOf(test) {
	return String(test.context?.status ?? "");
}
/**
* The live clicky tasks behind a lint run, such as the per-review prompt tasks
* under the `lint prompt` group, with each task's own status text (thinking,
* tool: read_file). Renders nothing while no task is running or pending.
*/
function LintTasks({ tests }) {
	const groups = liveTaskGroups(tests);
	if (groups.length === 0) return null;
	return /* @__PURE__ */ jsxs("section", {
		"aria-label": "Tasks",
		className: "flex flex-col gap-1 border-b border-border px-density-3 py-density-2 text-xs",
		children: [/* @__PURE__ */ jsx("h3", {
			className: "text-xs font-medium uppercase tracking-wide text-muted-foreground",
			children: "Tasks"
		}), groups.map((group) => {
			const rows = group.children?.length ? group.children : [group];
			return /* @__PURE__ */ jsxs("div", {
				className: "flex flex-col gap-0.5",
				children: [/* @__PURE__ */ jsx("span", {
					className: "text-muted-foreground",
					children: group.name
				}), /* @__PURE__ */ jsx("ul", {
					className: "flex flex-col gap-0.5",
					children: rows.map((row) => {
						const running = statusOf(row) === "running";
						return /* @__PURE__ */ jsxs("li", {
							className: "flex items-center gap-2",
							children: [
								/* @__PURE__ */ jsx(Icon, {
									icon: running ? UiLoader : UiCircleOutline,
									className: running ? "animate-spin text-blue-500" : "text-muted-foreground"
								}),
								/* @__PURE__ */ jsx("span", {
									className: "truncate font-medium",
									children: row.name
								}),
								/* @__PURE__ */ jsx("span", {
									className: "truncate text-muted-foreground",
									children: row.message || statusOf(row)
								})
							]
						}, row.task_id ?? row.name);
					})
				})]
			}, group.name);
		})]
	});
}
export { LintTasks };

//# sourceMappingURL=LintTasks.js.map