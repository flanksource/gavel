import { t as e } from "./jsx-runtime-B4NxTvNR.js";
import { a as t } from "./button-BLoR2pgR.js";
import { n } from "./SplitButton-w2X8sfgS.js";
// ../../../clicky-ui/packages/ui/src/layout/Tabs.tsx
var r = e();
function i({ tabs: e, value: i, onChange: a, variant: o = "underline", className: s }) {
	return /* @__PURE__ */ (0, r.jsx)("div", {
		role: "tablist",
		className: t("flex items-center gap-density-1", o === "underline" && "border-b border-border", s),
		children: e.map((e) => /* @__PURE__ */ (0, r.jsx)(n, {
			active: e.id === i,
			onClick: () => {
				e.disabled || a(e.id);
			},
			label: e.label,
			variant: o,
			...e.icon === void 0 ? {} : { icon: e.icon },
			...e.count === void 0 ? {} : { count: e.count },
			...e.countColor === void 0 ? {} : { countColor: e.countColor },
			...e.disabled ? { className: "opacity-50 pointer-events-none" } : {}
		}, e.id))
	});
}
export { i as t };
