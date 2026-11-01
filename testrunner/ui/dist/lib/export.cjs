const require_routes = require("./routes.cjs");
let _flanksource_clicky_ui_components = require("@flanksource/clicky-ui/components");
// src/export.ts
function exportFilename(state, format) {
	return `${(state.selectedPath ? state.selectedPath.replace(/\//g, "-") : state.tab) || state.tab}.${format}`;
}
function downloadCurrentView(state, format) {
	const link = document.createElement("a");
	link.href = require_routes.buildExportRoute(state, format);
	link.download = exportFilename(state, format);
	document.body.appendChild(link);
	link.click();
	link.remove();
}
function promptHeader(state) {
	if (state.tab === "lint") return "Use this gavel lint report as context for your analysis or code changes.";
	if (state.tab === "bench") return "Use this gavel benchmark report as context for your analysis or code changes.";
	return "Use this gavel test report as context for your analysis or code changes.";
}
/** Copies the view's Markdown export, framed as an AI prompt; rejects when the export or the clipboard fails. */
async function copyCurrentViewForAgent(state) {
	const exportPath = require_routes.buildExportRoute(state, "md");
	const exportURL = new URL(exportPath, window.location.origin).toString();
	const response = await fetch(exportPath, { headers: { Accept: "text/markdown" } });
	if (!response.ok) throw new Error(`Markdown export failed (${response.status})`);
	const markdown = (await response.text()).trimEnd();
	const prompt = [
		promptHeader(state),
		"",
		`Source URL: ${exportURL}`,
		"",
		"```md",
		markdown,
		"```"
	].join("\n");
	await (0, _flanksource_clicky_ui_components.copyText)(prompt);
}
exports.copyCurrentViewForAgent = copyCurrentViewForAgent;
exports.downloadCurrentView = downloadCurrentView;

//# sourceMappingURL=export.cjs.map