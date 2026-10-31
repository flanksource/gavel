import { t as e } from "./jsx-runtime-lVm_Myqd.js";
import { T as t } from "./Callout-DmCAGTzW.js";
import { c as n, s as r } from "./clipboard-C0PFseuO.js";
// ../../../clicky-ui/packages/ui/src/data/lint/stats.ts
var i = {
	info: 1,
	warning: 2,
	error: 3
};
function a(e) {
	return i[e];
}
function o(e) {
	return e.severity ?? "error";
}
function s(e) {
	let t = "info";
	for (let n of e ?? []) {
		let e = o(n);
		i[e] > i[t] && (t = e);
	}
	return t;
}
function c(e) {
	return e.status !== void 0;
}
function l(e) {
	return !e.skipped && !c(e) && (!!e.error || !!e.timed_out);
}
function u(e) {
	return e.error || "timed out";
}
function d(e) {
	let t = /* @__PURE__ */ new Set();
	for (let n of e) ((n.violations ?? []).length > 0 || l(n) || c(n)) && t.add(n.linter);
	return Array.from(t).sort();
}
function f(e, t) {
	let n = {
		error: 0,
		warning: 0,
		info: 0
	};
	for (let i of e) if (r(i.linter, t)) for (let e of i.violations ?? []) n[o(e)] += 1;
	return n;
}
function p(e, t) {
	let n = {};
	for (let i of e) for (let e of i.violations ?? []) r(o(e), t) && (n[i.linter] = (n[i.linter] ?? 0) + 1);
	return n;
}
function m(e) {
	return e.reduce((e, t) => e + (t.violations ?? []).length, 0);
}
function h(e) {
	return e.nestedUnder && e.linterName ? e.linterName.slice(e.nestedUnder.length + 1) : e.linterName;
}
function g(e) {
	return e.violations ? e.violations.length : (e.noFileViolations?.length ?? 0) + (e.children ?? []).reduce((e, t) => e + g(t), 0);
}
function _(e) {
	if (e.kind === "linter-error") return "error";
	let t = "info", n = (e) => {
		i[e] > i[t] && (t = e);
	};
	e.violations && n(s(e.violations)), e.noFileViolations && n(s(e.noFileViolations));
	for (let t of e.children ?? []) n(_(t));
	return t;
}
// ../../../clicky-ui/packages/ui/src/layout/Tabs.tsx
var v = e();
function y({ tabs: e, value: r, onChange: i, variant: a = "underline", className: o }) {
	return /* @__PURE__ */ (0, v.jsx)("div", {
		role: "tablist",
		className: t("flex items-center gap-density-1", a === "underline" && "border-b border-border", o),
		children: e.map((e) => /* @__PURE__ */ (0, v.jsx)(n, {
			active: e.id === r,
			onClick: () => {
				e.disabled || i(e.id);
			},
			label: e.label,
			variant: a,
			...e.icon === void 0 ? {} : { icon: e.icon },
			...e.count === void 0 ? {} : { count: e.count },
			...e.countColor === void 0 ? {} : { countColor: e.countColor },
			...e.disabled ? { className: "opacity-50 pointer-events-none" } : {}
		}, e.id))
	});
}
export { l as a, _ as c, o as d, a as f, f as i, u as l, s as m, d as n, c as o, m as p, p as r, g as s, y as t, h as u };
