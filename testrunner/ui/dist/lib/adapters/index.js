import { GavelDetailBody } from "./GavelDetailBody.js";
import { NodeActions } from "./NodeActions.js";
import { ResourcesPanel } from "./ResourcesPanel.js";
import { StepProcesses } from "./StepProcesses.js";
import { TracesPanel } from "./TracesPanel.js";
import { createTestRunnerRegistry, verificationAdapters } from "@flanksource/clicky-ui/data";
import { jsx } from "react/jsx-runtime";
// src/adapters/index.tsx
var asGavel = (node) => node;
function fixtureTraces(node) {
	const gavel = asGavel(node);
	if (gavel.framework !== "fixture") return [];
	return gavel.context?.traces ?? [];
}
function stepProcesses(node, processes) {
	const pids = new Set(asGavel(node).fixture?.pids ?? []);
	if (pids.size === 0) return [];
	return processes.filter((process) => pids.has(process.pid));
}
/**
* The detail body of every node that is not a fixture step: suite, package,
* command, error, Go and Ginkgo context, attempt metrics and output. Fixture
* steps keep clicky's verification body (command, CEL, typed step sections).
*/
function detailAdapter() {
	return {
		id: "gavel-detail",
		match: (node) => asGavel(node).framework !== "fixture",
		renderDetail: ({ node }) => /* @__PURE__ */ jsx(GavelDetailBody, { node })
	};
}
/** The Traces tab of a fixture node: its windows, opened in the stores of the runs they name (traces rebased onto the page). */
function tracesAdapter(traces) {
	return {
		id: "gavel-traces",
		match: (node) => fixtureTraces(node).length > 0,
		detailTabs: ({ node }) => [{
			id: "traces",
			label: "Traces",
			render: () => /* @__PURE__ */ jsx(TracesPanel, {
				windows: fixtureTraces(node),
				traces
			})
		}]
	};
}
/** The Processes tab of a fixture step: the process-tree nodes whose pids the step started. */
function processesAdapter({ processes, onOpenProcess }) {
	return {
		id: "gavel-processes",
		match: (node) => stepProcesses(node, processes).length > 0,
		detailTabs: ({ node }) => [{
			id: "processes",
			label: "Processes",
			render: () => /* @__PURE__ */ jsx(StepProcesses, {
				processes: stepProcesses(node, processes),
				onOpenProcess
			})
		}]
	};
}
/** The Resources tab of a fixture node with resource profiling: process-tree peaks and SQL. */
function resourcesAdapter() {
	return {
		id: "gavel-resources",
		match: (node) => {
			const gavel = asGavel(node);
			return !!gavel.fixture_profile || !!gavel.sql_profile;
		},
		detailTabs: ({ node }) => [{
			id: "resources",
			label: "Resources",
			render: () => /* @__PURE__ */ jsx(ResourcesPanel, { node: asGavel(node) })
		}]
	};
}
/** Download, AI-prompt copy and Skip/Delete source actions in the node's detail header. */
function nodeActionsAdapter(config) {
	return {
		id: "gavel-actions",
		match: (node) => asGavel(node).framework !== "task",
		nodeActions: ({ node }) => /* @__PURE__ */ jsx(NodeActions, {
			node: asGavel(node),
			config
		})
	};
}
/** Gavel's node adapters, in the order they take precedence. */
function gavelAdapters({ traces, processes, onOpenProcess, actions }) {
	return [
		detailAdapter(),
		nodeActionsAdapter(actions),
		tracesAdapter(traces),
		processesAdapter({
			processes,
			onOpenProcess
		}),
		resourcesAdapter()
	];
}
/**
* The registry the test runner renders gavel runs with: gavel's adapters, then
* clicky's verification adapters, which own fixture steps (attempt chooser,
* typed step sections, copy brief).
*/
function gavelRegistry(options) {
	return createTestRunnerRegistry([...gavelAdapters(options), ...verificationAdapters().adapters]);
}
export { detailAdapter, gavelAdapters, gavelRegistry, nodeActionsAdapter, processesAdapter, resourcesAdapter, tracesAdapter };

//# sourceMappingURL=index.js.map