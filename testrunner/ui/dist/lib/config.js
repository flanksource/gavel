// src/config.ts
var basePath = typeof window !== "undefined" && window.__gavelBasePath || "";
/** The URL prefix of one run's server: the page's base path, then the run id when a MultiServer hosts many runs. */
function apiBase({ baseUrl = basePath, runId } = {}) {
	return runId ? `${baseUrl}/${encodeURIComponent(runId)}` : baseUrl;
}
function apiUrl(path, scope) {
	return apiBase(scope) + path;
}
export { apiBase, apiUrl, basePath };

//# sourceMappingURL=config.js.map