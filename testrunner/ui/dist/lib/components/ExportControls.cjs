const require_routes = require("../routes.cjs");
const require_export = require("../export.cjs");
const require_use_copy_feedback = require("../hooks/use-copy-feedback.cjs");
let react_jsx_runtime = require("react/jsx-runtime");
let _flanksource_clicky_ui_components = require("@flanksource/clicky-ui/components");
let _flanksource_clicky_ui_icons = require("@flanksource/clicky-ui/icons");
// src/components/ExportControls.tsx
/** Download of the view (or one node of it) as JSON or Markdown, served by the export routes. */
function DownloadMenu({ routeState, title }) {
	return /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_components.DropdownMenu, {
		label: "Download",
		icon: _flanksource_clicky_ui_icons.UiDownload,
		variant: "outline",
		size: "sm",
		items: [{
			label: "JSON",
			icon: _flanksource_clicky_ui_icons.UiJson,
			title: require_routes.buildExportRoute(routeState, "json"),
			onSelect: () => require_export.downloadCurrentView(routeState, "json")
		}, {
			label: "Markdown",
			icon: _flanksource_clicky_ui_icons.UiMarkdown,
			title: require_routes.buildExportRoute(routeState, "md"),
			onSelect: () => require_export.downloadCurrentView(routeState, "md")
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
	idle: _flanksource_clicky_ui_icons.UiCopy,
	copying: _flanksource_clicky_ui_icons.UiLoader,
	copied: _flanksource_clicky_ui_icons.UiCheck,
	error: _flanksource_clicky_ui_icons.UiWarningCircle
};
/** Copies the Markdown export of `routeState` framed as an AI prompt. */
function CopyAIPromptButton({ routeState, disabled, disabledTitle, title }) {
	const { copyState, copyError, run } = require_use_copy_feedback.useCopyAction(() => require_export.copyCurrentViewForAgent(routeState));
	const Icon = COPY_ICON[copyState];
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(_flanksource_clicky_ui_components.Button, {
		type: "button",
		variant: "outline",
		size: "sm",
		disabled: disabled || copyState === "copying",
		onClick: () => {
			run();
		},
		title: disabled ? disabledTitle : copyError || title || require_routes.buildExportRoute(routeState, "md"),
		children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(Icon, { className: copyState === "copying" ? "mr-1 animate-spin" : "mr-1" }), COPY_LABEL[copyState]]
	});
}
exports.CopyAIPromptButton = CopyAIPromptButton;
exports.DownloadMenu = DownloadMenu;

//# sourceMappingURL=ExportControls.cjs.map