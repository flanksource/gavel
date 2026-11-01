import { ProcessTraceTree } from "@flanksource/clicky-ui/data";
import { jsx } from "react/jsx-runtime";
// src/adapters/StepProcesses.tsx
/**
* The processes a fixture step started, as the slice of the run's process tree
* whose pids it lists. Selecting one opens it on the Processes tab.
*/
function StepProcesses({ processes, onOpenProcess }) {
	return /* @__PURE__ */ jsx("div", {
		className: "flex h-80 min-h-0 flex-col px-3 pt-2",
		"data-step-processes": "",
		children: /* @__PURE__ */ jsx(ProcessTraceTree, {
			nodes: processes,
			selectedKey: null,
			expandAll: true,
			onSelect: onOpenProcess,
			emptyLabel: "This step started no processes the trace recorded"
		})
	});
}
export { StepProcesses };

//# sourceMappingURL=StepProcesses.js.map