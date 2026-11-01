import { buildExportRoute } from "../routes.js";
import { copyCurrentViewForAgent, downloadCurrentView } from "../export.js";
import { useCopyAction } from "../hooks/use-copy-feedback.js";
import { jsx, jsxs } from "react/jsx-runtime";
import { Button, DropdownMenu } from "@flanksource/clicky-ui/components";
import { UiCheck, UiCopy, UiDownload, UiJson, UiLoader, UiMarkdown, UiWarningCircle } from "@flanksource/clicky-ui/icons";
// src/components/ExportControls.tsx
/** Download of the view (or one node of it) as JSON or Markdown, served by the export routes. */
function DownloadMenu({ routeState, title }) {
	return /* @__PURE__ */ jsx(DropdownMenu, {
		label: "Download",
		icon: UiDownload,
		variant: "outline",
		size: "sm",
		items: [{
			label: "JSON",
			icon: UiJson,
			title: buildExportRoute(routeState, "json"),
			onSelect: () => downloadCurrentView(routeState, "json")
		}, {
			label: "Markdown",
			icon: UiMarkdown,
			title: buildExportRoute(routeState, "md"),
			onSelect: () => downloadCurrentView(routeState, "md")
		}],
		...title ? { title } : {}
	});
}
var COPY_LABEL = {
	idle: "Copy AI Prompt",
	copying: "Copying...",
	copied: "Copied",
	error: "Copy failed"
};
var COPY_ICON = {
	idle: UiCopy,
	copying: UiLoader,
	copied: UiCheck,
	error: UiWarningCircle
};
/** Copies the Markdown export of `routeState` framed as an AI prompt. */
function CopyAIPromptButton({ routeState, disabled, disabledTitle, title }) {
	const { copyState, copyError, run } = useCopyAction(() => copyCurrentViewForAgent(routeState));
	const Icon = COPY_ICON[copyState];
	return /* @__PURE__ */ jsxs(Button, {
		type: "button",
		variant: "outline",
		size: "sm",
		disabled: disabled || copyState === "copying",
		onClick: () => {
			run();
		},
		title: disabled ? disabledTitle : copyError || title || buildExportRoute(routeState, "md"),
		children: [/* @__PURE__ */ jsx(Icon, { className: copyState === "copying" ? "mr-1 animate-spin" : "mr-1" }), COPY_LABEL[copyState]]
	});
}
export { CopyAIPromptButton, DownloadMenu };

//# sourceMappingURL=ExportControls.js.map