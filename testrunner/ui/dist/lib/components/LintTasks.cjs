const require_utils = require("../utils.cjs");
let _flanksource_clicky_ui_data = require("@flanksource/clicky-ui/data");
let react_jsx_runtime = require("react/jsx-runtime");
let _flanksource_clicky_ui_icons = require("@flanksource/clicky-ui/icons");
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
	const groups = require_utils.liveTaskGroups(tests);
	if (groups.length === 0) return null;
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("section", {
		"aria-label": "Tasks",
		className: "flex flex-col gap-1 border-b border-border px-density-3 py-density-2 text-xs",
		children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)("h3", {
			className: "text-xs font-medium uppercase tracking-wide text-muted-foreground",
			children: "Tasks"
		}), groups.map((group) => {
			const rows = group.children?.length ? group.children : [group];
			return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
				className: "flex flex-col gap-0.5",
				children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
					className: "text-muted-foreground",
					children: group.name
				}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)("ul", {
					className: "flex flex-col gap-0.5",
					children: rows.map((row) => {
						const running = statusOf(row) === "running";
						return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("li", {
							className: "flex items-center gap-2",
							children: [
								/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.Icon, {
									icon: running ? _flanksource_clicky_ui_icons.UiLoader : _flanksource_clicky_ui_icons.UiCircleOutline,
									className: running ? "animate-spin text-blue-500" : "text-muted-foreground"
								}),
								/* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
									className: "truncate font-medium",
									children: row.name
								}),
								/* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
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
exports.LintTasks = LintTasks;

//# sourceMappingURL=LintTasks.cjs.map