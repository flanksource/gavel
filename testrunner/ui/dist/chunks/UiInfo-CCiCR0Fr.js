import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-B4NxTvNR.js";
import { a as r } from "./button-BLoR2pgR.js";
// ../../../clicky-ui/packages/ui/src/lib/size.ts
var i = {
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
function a(e, t = "comfortable") {
	return i[t][e];
}
// ../../../clicky-ui/packages/ui/src/hooks/use-density.ts
var o = /* @__PURE__ */ e(t(), 1), s = (0, o.createContext)(null);
function c() {
	let e = (0, o.useContext)(s);
	if (e) return e.density;
	if (typeof document > "u") return "comfortable";
	let t = document.documentElement.getAttribute("data-density");
	return t === "compact" || t === "spacious" ? t : "comfortable";
}
// ../../../clicky-ui/packages/ui/src/data/icon-fallback.ts
var l = null;
function u() {
	return l;
}
// ../../../clicky-ui/packages/ui/src/data/Icon.tsx
var d = n(), f = {
	emerald: "text-emerald-700 ring-emerald-200/80 dark:text-emerald-400 dark:ring-emerald-500/30",
	amber: "text-amber-700 ring-amber-200/80 dark:text-amber-400 dark:ring-amber-500/30",
	rose: "text-rose-700 ring-rose-200/80 dark:text-rose-400 dark:ring-rose-500/30",
	slate: "text-slate-700 ring-slate-200/80 dark:text-slate-300 dark:ring-slate-500/30",
	sky: "text-sky-700 ring-sky-200/80 dark:text-sky-400 dark:ring-sky-500/30",
	violet: "text-violet-700 ring-violet-200/80 dark:text-violet-400 dark:ring-violet-500/30",
	neutral: "text-foreground ring-border"
}, p = "1em";
function m(e, t, n, r, i, a, o, s) {
	let c = [
		typeof r == "number" ? `rotate(${r}deg)` : typeof r == "string" ? r : "",
		i === "horizontal" || i === "both" ? "scaleX(-1)" : "",
		i === "vertical" || i === "both" ? "scaleY(-1)" : ""
	].filter(Boolean).join(" ");
	return /* @__PURE__ */ (0, d.jsx)(e, {
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
function h(e, t, n, i, a) {
	let o = u();
	if (o) {
		let r = o, s = typeof t == "number" ? t : n;
		return /* @__PURE__ */ (0, d.jsx)(r, {
			name: e,
			...i ? { className: i } : {},
			...s == null ? {} : { size: s },
			...a ? { alt: a } : {}
		});
	}
	return /* @__PURE__ */ (0, d.jsx)("span", {
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
function g({ icon: e, name: t, className: n, width: i, height: o, rotate: s, flip: l, inline: u, title: g, style: _ = "plain", size: v, tone: y = "neutral" }) {
	let b = c(), x = v == null ? void 0 : a(v, b), S = x == null ? void 0 : Math.max(8, Math.round(x * .5)), C = i ?? o ?? (_ === "badge" ? S : x ?? p), w = o ?? C, T = r("inline-block shrink-0 align-[-0.125em]", _ === "plain" ? n : void 0), E = _ === "plain" ? g : void 0, D = e ? m(e, C, w, s, l, u, T, E) : h(t, C, w, T, E);
	if (_ === "plain") return D;
	let O = x == null ? void 0 : {
		width: x,
		height: x
	};
	return /* @__PURE__ */ (0, d.jsx)("span", {
		title: g,
		"aria-label": g,
		role: g ? "img" : void 0,
		className: r("inline-flex items-center justify-center rounded-full bg-background shadow-sm ring-1 ring-inset", f[y], n),
		style: O,
		children: D
	});
}
function _({ icon: e, className: t }) {
	return e == null || e === "" || e === !1 ? null : typeof e == "string" ? /* @__PURE__ */ (0, d.jsx)(g, {
		name: e,
		className: r("shrink-0", t)
	}) : /* @__PURE__ */ (0, d.jsx)(d.Fragment, { children: e });
}
// ../../../clicky-ui/packages/ui/src/icons/components/UiClose.tsx
var v = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, d.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, d.jsx)("path", {
		fill: "currentColor",
		d: "M204.24 195.76a6 6 0 1 1-8.48 8.48L128 136.49l-67.76 67.75a6 6 0 0 1-8.48-8.48L119.51 128L51.76 60.24a6 6 0 0 1 8.48-8.48L128 119.51l67.76-67.75a6 6 0 0 1 8.48 8.48L136.49 128Z"
	})
}), {
	__source: "ph:x-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "close",
	displayName: "UiClose"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, d.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, d.jsx)("path", {
		fill: "currentColor",
		d: "M208 32H48a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16m-26.34 138.34a8 8 0 0 1-11.32 11.32L128 139.31l-42.34 42.35a8 8 0 0 1-11.32-11.32L116.69 128L74.34 85.66a8 8 0 0 1 11.32-11.32L128 116.69l42.34-42.35a8 8 0 0 1 11.32 11.32L139.31 128Z"
	})
}), {
	__source: "ph:x-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "close",
	displayName: "UiCloseFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiWarningTriangle.tsx
var y = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#d97706",
		...r
	};
	return /* @__PURE__ */ (0, d.jsx)("svg", {
		xmlns: "http://www.w3.org/2000/svg",
		width: e,
		height: e,
		viewBox: "0 0 256 256",
		role: n ? "img" : "presentation",
		"aria-label": n,
		"aria-hidden": !n || void 0,
		className: t,
		style: a,
		...i,
		children: /* @__PURE__ */ (0, d.jsx)("path", {
			fill: "currentColor",
			d: "M235.07 189.09L147.61 37.22a22.75 22.75 0 0 0-39.22 0L20.93 189.09a21.53 21.53 0 0 0 0 21.72A22.35 22.35 0 0 0 40.55 222h174.9a22.35 22.35 0 0 0 19.6-11.19a21.53 21.53 0 0 0 .02-21.72m-10.41 15.71a10.46 10.46 0 0 1-9.21 5.2H40.55a10.46 10.46 0 0 1-9.21-5.2a9.51 9.51 0 0 1 0-9.72l87.45-151.87a10.75 10.75 0 0 1 18.42 0l87.46 151.87a9.51 9.51 0 0 1-.01 9.72M122 144v-40a6 6 0 0 1 12 0v40a6 6 0 0 1-12 0m16 36a10 10 0 1 1-10-10a10 10 0 0 1 10 10"
		})
	});
}, {
	__source: "ph:warning-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "warning-triangle",
	__defaultColor: "#d97706",
	displayName: "UiWarningTriangle"
});
Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#d97706",
		...r
	};
	return /* @__PURE__ */ (0, d.jsx)("svg", {
		xmlns: "http://www.w3.org/2000/svg",
		width: e,
		height: e,
		viewBox: "0 0 256 256",
		role: n ? "img" : "presentation",
		"aria-label": n,
		"aria-hidden": !n || void 0,
		className: t,
		style: a,
		...i,
		children: /* @__PURE__ */ (0, d.jsx)("path", {
			fill: "currentColor",
			d: "M236.8 188.09L149.35 36.22a24.76 24.76 0 0 0-42.7 0L19.2 188.09a23.51 23.51 0 0 0 0 23.72A24.35 24.35 0 0 0 40.55 224h174.9a24.35 24.35 0 0 0 21.33-12.19a23.51 23.51 0 0 0 .02-23.72M120 104a8 8 0 0 1 16 0v40a8 8 0 0 1-16 0Zm8 88a12 12 0 1 1 12-12a12 12 0 0 1-12 12"
		})
	});
}, {
	__source: "ph:warning-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "warning-triangle",
	__defaultColor: "#d97706",
	displayName: "UiWarningTriangleFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiInfo.tsx
var b = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#0ea5e9",
		...r
	};
	return /* @__PURE__ */ (0, d.jsx)("svg", {
		xmlns: "http://www.w3.org/2000/svg",
		width: e,
		height: e,
		viewBox: "0 0 256 256",
		role: n ? "img" : "presentation",
		"aria-label": n,
		"aria-hidden": !n || void 0,
		className: t,
		style: a,
		...i,
		children: /* @__PURE__ */ (0, d.jsx)("path", {
			fill: "currentColor",
			d: "M142 176a6 6 0 0 1-6 6a14 14 0 0 1-14-14v-40a2 2 0 0 0-2-2a6 6 0 0 1 0-12a14 14 0 0 1 14 14v40a2 2 0 0 0 2 2a6 6 0 0 1 6 6m-18-82a10 10 0 1 0-10-10a10 10 0 0 0 10 10m106 34A102 102 0 1 1 128 26a102.12 102.12 0 0 1 102 102m-12 0a90 90 0 1 0-90 90a90.1 90.1 0 0 0 90-90"
		})
	});
}, {
	__source: "ph:info-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "info",
	__defaultColor: "#0ea5e9",
	displayName: "UiInfo"
});
Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#0ea5e9",
		...r
	};
	return /* @__PURE__ */ (0, d.jsx)("svg", {
		xmlns: "http://www.w3.org/2000/svg",
		width: e,
		height: e,
		viewBox: "0 0 256 256",
		role: n ? "img" : "presentation",
		"aria-label": n,
		"aria-hidden": !n || void 0,
		className: t,
		style: a,
		...i,
		children: /* @__PURE__ */ (0, d.jsx)("path", {
			fill: "currentColor",
			d: "M128 24a104 104 0 1 0 104 104A104.11 104.11 0 0 0 128 24m-4 48a12 12 0 1 1-12 12a12 12 0 0 1 12-12m12 112a16 16 0 0 1-16-16v-40a8 8 0 0 1 0-16a16 16 0 0 1 16 16v40a8 8 0 0 1 0 16"
		})
	});
}, {
	__source: "ph:info-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "info",
	__defaultColor: "#0ea5e9",
	displayName: "UiInfoFilled"
});
export { _ as a, c, g as i, a as l, y as n, u as o, v as r, s, b as t };
