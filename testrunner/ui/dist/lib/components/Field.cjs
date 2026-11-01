let react_jsx_runtime = require("react/jsx-runtime");
// src/components/Field.tsx
/** A labelled block of a detail pane, in the label style clicky-ui's own detail bodies use. */
function Field({ title, children }) {
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
		className: "space-y-1",
		children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
			className: "text-[10px] uppercase tracking-wide text-muted-foreground",
			children: title
		}), children]
	});
}
exports.Field = Field;

//# sourceMappingURL=Field.cjs.map