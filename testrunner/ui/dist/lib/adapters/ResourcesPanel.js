import { Field } from "../components/Field.js";
import { ProfileTiles, SqlStatements } from "../components/ResourceParts.js";
import { MetricTile } from "@flanksource/clicky-ui/data";
import { jsx, jsxs } from "react/jsx-runtime";
// src/adapters/ResourcesPanel.tsx
/**
* What a fixture step's resource profiling measured: the process tree's peak
* CPU/RSS/memory and observed disk IO, and the SQL it ran. Go profiles are
* `goprof` windows on the Traces tab.
*/
function ResourcesPanel({ node }) {
	const sql = node.sql_profile;
	return /* @__PURE__ */ jsxs("div", {
		className: "space-y-4 p-density-4",
		children: [node.fixture_profile && /* @__PURE__ */ jsx(Field, {
			title: "Resources",
			children: /* @__PURE__ */ jsx(ProfileTiles, { profile: node.fixture_profile })
		}), sql && /* @__PURE__ */ jsxs(Field, {
			title: "SQL profile",
			children: [/* @__PURE__ */ jsxs("div", {
				className: "grid grid-cols-3 gap-density-1",
				children: [
					/* @__PURE__ */ jsx(MetricTile, {
						label: "Queries",
						value: String(sql.query_count)
					}),
					/* @__PURE__ */ jsx(MetricTile, {
						label: "Total SQL",
						value: `${sql.total_duration_ms.toFixed(1)} ms`
					}),
					/* @__PURE__ */ jsx(MetricTile, {
						label: "Slow queries",
						value: String(sql.slow_query_count)
					})
				]
			}), /* @__PURE__ */ jsx(SqlStatements, { profile: sql })]
		})]
	});
}
export { ResourcesPanel };

//# sourceMappingURL=ResourcesPanel.js.map