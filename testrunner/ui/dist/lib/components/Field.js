import { jsx, jsxs } from "react/jsx-runtime";
// src/components/Field.tsx
/** A labelled block of a detail pane, in the label style clicky-ui's own detail bodies use. */
function Field({ title, children }) {
	return /* @__PURE__ */ jsxs("div", {
		className: "space-y-1",
		children: [/* @__PURE__ */ jsx("div", {
			className: "text-[10px] uppercase tracking-wide text-muted-foreground",
			children: title
		}), children]
	});
}
export { Field };

//# sourceMappingURL=Field.js.map