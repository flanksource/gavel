import { CopyAIPromptButton, DownloadMenu } from "../components/ExportControls.js";
import { useState } from "react";
import { sum } from "@flanksource/clicky-ui/data";
import { Fragment, jsx, jsxs } from "react/jsx-runtime";
import { ConfirmDialog, DropdownMenu } from "@flanksource/clicky-ui/components";
import { UiFile, UiSkipForward, UiTrash } from "@flanksource/clicky-ui/icons";
// src/adapters/NodeActions.tsx
var EDITABLE_FRAMEWORKS = /* @__PURE__ */ new Set([
	"go test",
	"ginkgo",
	"vitest"
]);
function canEditSource(node, supported) {
	return supported && node.framework !== "task" && EDITABLE_FRAMEWORKS.has(node.framework ?? "") && !!node.file;
}
function hasFailingContent(node) {
	if (node.failed || node.timed_out) return true;
	if ((node.children?.length ?? 0) === 0) return false;
	const counts = sum(node);
	return counts.failed > 0 || counts.timedout > 0;
}
function SourceActions({ node, busy, onEdit }) {
	const [pending, setPending] = useState(null);
	const target = pending?.scope === "file" ? node.file || "file" : node.name || "test";
	const verb = pending?.action === "skip" ? "Skip" : "Delete";
	const scopeLabel = pending?.scope === "file" ? "file" : "test";
	const choose = (action, scope) => () => setPending({
		action,
		scope
	});
	return /* @__PURE__ */ jsxs(Fragment, { children: [/* @__PURE__ */ jsx(DropdownMenu, {
		label: "Source",
		title: "Edit the test source",
		items: [
			{
				label: "Skip Test",
				icon: UiSkipForward,
				disabled: busy,
				onSelect: choose("skip", "test")
			},
			{
				label: "Delete Test",
				icon: UiTrash,
				disabled: busy,
				onSelect: choose("delete", "test")
			},
			{
				label: "Skip File",
				icon: UiFile,
				disabled: busy,
				onSelect: choose("skip", "file")
			},
			{
				label: "Delete File",
				icon: UiTrash,
				disabled: busy,
				onSelect: choose("delete", "file")
			}
		]
	}), /* @__PURE__ */ jsx(ConfirmDialog, {
		open: pending !== null,
		title: `${verb} ${scopeLabel} ${target}?`,
		message: "This edits the source file on disk.",
		confirmLabel: verb,
		variant: pending?.action === "delete" ? "destructive" : "default",
		onCancel: () => setPending(null),
		onConfirm: () => {
			const edit = pending;
			setPending(null);
			if (edit) onEdit(node, edit.action, edit.scope);
		}
	})] });
}
/**
* The header actions of a test node: download and AI-prompt copy of its export,
* and, for a Go, Ginkgo or vitest test with a source file, the Skip / Delete
* source edits.
*/
function NodeActions({ node, config }) {
	const nodeRouteState = node.framework === "task" ? void 0 : config.nodeRouteState(node);
	const failingOnly = config.failingOnlyRouteState(node);
	const failing = hasFailingContent(node);
	return /* @__PURE__ */ jsxs(Fragment, { children: [
		nodeRouteState && /* @__PURE__ */ jsx(DownloadMenu, {
			routeState: nodeRouteState,
			title: "Download this node as JSON or Markdown"
		}),
		nodeRouteState && failingOnly && /* @__PURE__ */ jsx(CopyAIPromptButton, {
			routeState: failingOnly,
			disabled: !failing,
			disabledTitle: "No failures to copy",
			title: "Copy failing output and repro steps as AI prompt"
		}),
		canEditSource(node, config.testEditSupported) && /* @__PURE__ */ jsx(SourceActions, {
			node,
			busy: config.testEditBusy,
			onEdit: config.onTestEdit
		})
	] });
}
export { NodeActions, canEditSource };

//# sourceMappingURL=NodeActions.js.map