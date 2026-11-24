import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-a1ZmknZu.js";
import { o as r } from "./button-Dbm88adW.js";
// ../../../clicky-ui/packages/ui/src/hooks/use-density.ts
var i = /* @__PURE__ */ e(t(), 1), a = (0, i.createContext)(null);
function o() {
	let e = (0, i.useContext)(a);
	if (e) return e.density;
	if (typeof document > "u") return "comfortable";
	let t = document.documentElement.getAttribute("data-density");
	return t === "compact" || t === "spacious" ? t : "comfortable";
}
// ../../../clicky-ui/packages/ui/src/lib/size.ts
var s = {
	compact: {
		xs: 14,
		sm: 18,
		md: 24,
		lg: 36,
		xl: 48
	},
	comfortable: {
		xs: 16,
		sm: 20,
		md: 28,
		lg: 40,
		xl: 56
	},
	spacious: {
		xs: 20,
		sm: 24,
		md: 32,
		lg: 48,
		xl: 64
	}
};
function c(e, t = "comfortable") {
	return s[t][e];
}
// ../../../clicky-ui/packages/ui/src/data/icon-fallback.ts
var l = null;
function u(e) {
	l = e;
}
function d() {
	return l;
}
// ../../../clicky-ui/packages/ui/src/data/Icon.tsx
var f = /* @__PURE__ */ e(n(), 1), p = {
	emerald: "text-emerald-700 ring-emerald-200/80 dark:text-emerald-400 dark:ring-emerald-500/30",
	amber: "text-amber-700 ring-amber-200/80 dark:text-amber-400 dark:ring-amber-500/30",
	rose: "text-rose-700 ring-rose-200/80 dark:text-rose-400 dark:ring-rose-500/30",
	slate: "text-slate-700 ring-slate-200/80 dark:text-slate-300 dark:ring-slate-500/30",
	sky: "text-sky-700 ring-sky-200/80 dark:text-sky-400 dark:ring-sky-500/30",
	violet: "text-violet-700 ring-violet-200/80 dark:text-violet-400 dark:ring-violet-500/30",
	neutral: "text-foreground ring-border"
}, m = "1em";
function h(e, t, n, r, i, a, o, s) {
	let c = [
		typeof r == "number" ? `rotate(${r}deg)` : typeof r == "string" ? r : "",
		i === "horizontal" || i === "both" ? "scaleX(-1)" : "",
		i === "vertical" || i === "both" ? "scaleY(-1)" : ""
	].filter(Boolean).join(" ");
	return /* @__PURE__ */ (0, f.jsx)(e, {
		...o ? { className: o } : {},
		...t == null ? {} : { width: t },
		...n == null ? {} : { height: n },
		...a == null ? {} : { "aria-hidden": !s },
		...s ? {
			title: s,
			role: "img"
		} : { "aria-hidden": !0 },
		...c ? { style: { transform: c } } : {}
	});
}
function g(e, t, n, i, a) {
	let o = d();
	if (o) {
		let r = o, s = typeof t == "number" ? t : n;
		return /* @__PURE__ */ (0, f.jsx)(r, {
			name: e,
			...i ? { className: i } : {},
			...s == null ? {} : { size: s },
			...a ? { alt: a } : {}
		});
	}
	return /* @__PURE__ */ (0, f.jsx)("span", {
		className: r("inline-flex shrink-0 items-center justify-center rounded-sm border border-dashed border-muted-foreground/50 text-[0.65em] leading-none text-muted-foreground", i),
		title: a ?? e,
		role: a ? "img" : void 0,
		"aria-label": a,
		"aria-hidden": !a || void 0,
		style: {
			width: t,
			height: n
		},
		children: "?"
	});
}
function _({ icon: e, name: t, className: n, width: i, height: a, rotate: s, flip: l, inline: u, title: d, style: _ = "plain", size: v, tone: y = "neutral" }) {
	let b = o(), x = v == null ? void 0 : c(v, b), S = x == null ? void 0 : Math.max(8, Math.round(x * .5)), C = i ?? a ?? (_ === "badge" ? S : x ?? m), w = a ?? C, T = r("inline-block shrink-0 align-[-0.125em]", _ === "plain" ? n : void 0), E = _ === "plain" ? d : void 0, D = e ? h(e, C, w, s, l, u, T, E) : g(t, C, w, T, E);
	if (_ === "plain") return D;
	let O = x == null ? void 0 : {
		width: x,
		height: x
	};
	return /* @__PURE__ */ (0, f.jsx)("span", {
		title: d,
		"aria-label": d,
		role: d ? "img" : void 0,
		className: r("inline-flex items-center justify-center rounded-full bg-background shadow-sm ring-1 ring-inset", p[y], n),
		style: O,
		children: D
	});
}
function v({ icon: e, className: t }) {
	return e == null || e === "" || e === !1 ? null : typeof e == "string" ? /* @__PURE__ */ (0, f.jsx)(_, {
		name: e,
		className: r("shrink-0", t)
	}) : /* @__PURE__ */ (0, f.jsx)(f.Fragment, { children: e });
}
export { c as a, u as i, v as n, a as o, d as r, o as s, _ as t };
