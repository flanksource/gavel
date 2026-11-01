const require_Field = require("../components/Field.cjs");
const require_ResourceParts = require("../components/ResourceParts.cjs");
let _flanksource_clicky_ui_data = require("@flanksource/clicky-ui/data");
let react_jsx_runtime = require("react/jsx-runtime");
// src/adapters/ResourcesPanel.tsx
/**
* What a fixture step's resource profiling measured: the process tree's peak
* CPU/RSS/memory and observed disk IO, and the SQL it ran. Go profiles are
* `goprof` windows on the Traces tab.
*/
function ResourcesPanel({ node }) {
	const sql = node.sql_profile;
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
		className: "space-y-4 p-density-4",
		children: [node.fixture_profile && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
			title: "Resources",
			children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_ResourceParts.ProfileTiles, { profile: node.fixture_profile })
		}), sql && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(require_Field.Field, {
			title: "SQL profile",
			children: [/* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
				className: "grid grid-cols-3 gap-density-1",
				children: [
					/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.MetricTile, {
						label: "Queries",
						value: String(sql.query_count)
					}),
					/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.MetricTile, {
						label: "Total SQL",
						value: `${sql.total_duration_ms.toFixed(1)} ms`
					}),
					/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.MetricTile, {
						label: "Slow queries",
						value: String(sql.slow_query_count)
					})
				]
			}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_ResourceParts.SqlStatements, { profile: sql })]
		})]
	});
}
exports.ResourcesPanel = ResourcesPanel;

//# sourceMappingURL=ResourcesPanel.cjs.map