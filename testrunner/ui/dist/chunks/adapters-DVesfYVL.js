import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./jsx-runtime-f3E3oPuc.js";
import { t as r } from "./react-dom-D8HDTZ2P.js";
import { T as i, _ as a, l as o, v as s, x as c } from "./Callout-DALwzckz.js";
import { B as l, D as u, Dt as d, E as f, Et as p, F as m, I as h, J as g, L as _, N as v, O as y, Ot as b, P as x, R as S, St as C, Tt as w, V as T, W as E, X as D, _ as O, _t as k, b as A, bt as j, ct as M, dt as N, ft as ee, h as P, ht as te, j as ne, kt as F, l as I, lt as L, m as re, mt as R, n as ie, p as ae, pt as oe, rt as se, s as ce, st as le, ut as ue, v as de, vt as fe, w as pe, wt as me, x as he, xt as z, y as ge, yt as _e, z as ve } from "./clipboard-CMVj3cUp.js";
import { B, C as ye, D as be, E as xe, F as Se, I as Ce, K as we, L as Te, M as Ee, N as De, O as Oe, P as ke, R as Ae, S as je, T as Me, V as Ne, X as Pe, b as Fe, c as Ie, ct as Le, d as Re, dt as ze, f as V, g as Be, i as Ve, k as He, lt as Ue, m as We, nt as Ge, p as Ke, s as H, u as qe, ut as Je, v as Ye, w as Xe, x as Ze, y as Qe, z as $e } from "./paths-D_tC_cYe.js";
import { a as et, l as tt, n as nt, o as rt, s as it } from "./floating-ui.react-dom-Dh-Ge66-.js";
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/AnsiHtml.tsx
var U = /* @__PURE__ */ e(t(), 1), at = /* @__PURE__ */ e(r(), 1), W = n(), ot = {
	30: "color:#1e1e1e",
	31: "color:#cd3131",
	32: "color:#0dbc79",
	33: "color:#e5e510",
	34: "color:#2472c8",
	35: "color:#bc3fbc",
	36: "color:#11a8cd",
	37: "color:#e5e5e5",
	90: "color:#666",
	91: "color:#f14c4c",
	92: "color:#23d18b",
	93: "color:#f5f543",
	94: "color:#3b8eea",
	95: "color:#d670d6",
	96: "color:#29b8db",
	97: "color:#fff",
	1: "font-weight:bold",
	2: "opacity:0.7",
	3: "font-style:italic",
	4: "text-decoration:underline"
};
function st(e) {
	let t = [], n = /\x1b\[([0-9;]*)m/g, r = 0, i = [], a;
	for (; (a = n.exec(e)) !== null;) {
		a.index > r && t.push({
			text: e.slice(r, a.index),
			style: i.join(";")
		});
		let n = (a[1] ?? "").split(";").filter(Boolean);
		for (let e of n) if (e === "0" || e === "") i = [];
		else {
			let t = ot[e];
			t && i.push(t);
		}
		r = a.index + a[0].length;
	}
	return r < e.length && t.push({
		text: e.slice(r),
		style: i.join(";")
	}), t;
}
function ct(e) {
	let t = {};
	for (let n of e.split(";")) {
		let [e, r] = n.split(":");
		if (!e || !r) continue;
		let i = e.replace(/-([a-z])/g, (e, t) => t.toUpperCase());
		t[i] = r;
	}
	return t;
}
function lt({ text: e, className: t, as: n = "pre" }) {
	let r = st(e).map((e, t) => e.style ? /* @__PURE__ */ (0, W.jsx)("span", {
		style: ct(e.style),
		children: e.text
	}, t) : /* @__PURE__ */ (0, W.jsx)("span", { children: e.text }, t));
	return n === "span" ? /* @__PURE__ */ (0, W.jsx)("span", {
		className: t,
		children: r
	}) : /* @__PURE__ */ (0, W.jsx)("pre", {
		className: t,
		children: r
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiFilter.tsx
var ut = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M228.77 50.34A13.8 13.8 0 0 0 216 42H40a14 14 0 0 0-10.33 23.42l.06.07L98 138.38V216a14 14 0 0 0 21.77 11.64l32-21.33a14 14 0 0 0 6.23-11.65v-56.28l68.33-73a13.82 13.82 0 0 0 2.44-15.04m-11.26 6.94l-69.89 74.62A6 6 0 0 0 146 136v58.66a2 2 0 0 1-.89 1.67l-32 21.33A2 2 0 0 1 110 216v-80a6 6 0 0 0-1.62-4.1L38.53 57.32A2 2 0 0 1 40 54h176a1.9 1.9 0 0 1 1.83 1.19a1.86 1.86 0 0 1-.32 2.09"
	})
}), {
	__source: "ph:funnel-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "filter",
	displayName: "UiFilter"
}), dt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "m227.81 66.76l-.08.09L160 139.17v55.49a16 16 0 0 1-7.13 13.34l-32 21.34A16 16 0 0 1 96 216v-76.83L28.27 66.85l-.08-.09A16 16 0 0 1 40 40h176a16 16 0 0 1 11.84 26.76Z"
	})
}), {
	__source: "ph:funnel-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "filter",
	displayName: "UiFilterFilled"
}), ft = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M52.44 36a6 6 0 0 0-8.88 8l20.88 23c-37.28 21.9-53.23 57-53.92 58.57a6 6 0 0 0 0 4.88c.34.77 8.66 19.22 27.24 37.8C55 185.47 84.62 206 128 206a124.9 124.9 0 0 0 52.57-11.25l23 25.29a6 6 0 0 0 8.88-8.08Zm48.62 71.32l45 49.52a34 34 0 0 1-45-49.52M128 194c-31.38 0-58.78-11.42-81.45-33.93A134.6 134.6 0 0 1 22.69 128c4.29-8.2 20.1-35.18 50-51.91l20.2 22.21a46 46 0 0 0 61.35 67.48l17.81 19.6A113.5 113.5 0 0 1 128 194m6.4-99.4a6 6 0 0 1 2.25-11.79a46.17 46.17 0 0 1 37.15 40.87a6 6 0 0 1-5.42 6.53h-.56a6 6 0 0 1-6-5.45A34.1 34.1 0 0 0 134.4 94.6m111.08 35.85c-.41.92-10.37 23-32.86 43.12a6 6 0 1 1-8-8.94A134.1 134.1 0 0 0 233.31 128a134.7 134.7 0 0 0-23.86-32.07C186.78 73.42 159.38 62 128 62a120 120 0 0 0-19.69 1.6a6 6 0 1 1-2-11.83A131 131 0 0 1 128 50c43.38 0 73 20.54 90.24 37.76c18.58 18.58 26.9 37 27.24 37.81a6 6 0 0 1 0 4.88"
	})
}), {
	__source: "ph:eye-slash-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "eye-closed",
	displayName: "UiEyeClosed"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M96.68 57.87a4 4 0 0 1 2.08-6.6A130 130 0 0 1 128 48c34.88 0 66.57 13.26 91.66 38.35c18.83 18.83 27.3 37.62 27.65 38.41a8 8 0 0 1 0 6.5c-.35.79-8.82 19.57-27.65 38.4q-4.28 4.26-8.79 8.07a4 4 0 0 1-5.55-.36Zm117.24 152.75a8 8 0 1 1-11.84 10.76L180 197.13A127.2 127.2 0 0 1 128 208c-34.88 0-66.57-13.26-91.66-38.34C17.51 150.83 9 132.05 8.69 131.26a8 8 0 0 1 0-6.5c.31-.76 8.82-19.58 27.65-38.41a135 135 0 0 1 25-19.78L42.08 45.38a8 8 0 1 1 11.84-10.76Zm-65.49-48.25l-52.69-58a40 40 0 0 0 52.69 58"
	})
}), {
	__source: "ph:eye-slash-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "eye-closed",
	displayName: "UiEyeClosedFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiZoomIn.tsx
var pt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M150 112a6 6 0 0 1-6 6h-26v26a6 6 0 0 1-12 0v-26H80a6 6 0 0 1 0-12h26V80a6 6 0 0 1 12 0v26h26a6 6 0 0 1 6 6m78.24 116.24a6 6 0 0 1-8.48 0l-51.38-51.38a86.15 86.15 0 1 1 8.48-8.48l51.38 51.38a6 6 0 0 1 0 8.48M112 186a74 74 0 1 0-74-74a74.09 74.09 0 0 0 74 74"
	})
}), {
	__source: "ph:magnifying-glass-plus-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "zoom-in",
	displayName: "UiZoomIn"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "m229.66 218.34l-50.06-50.06a88.21 88.21 0 1 0-11.32 11.31l50.06 50.07a8 8 0 0 0 11.32-11.32M144 120h-24v24a8 8 0 0 1-16 0v-24H80a8 8 0 0 1 0-16h24V80a8 8 0 0 1 16 0v24h24a8 8 0 0 1 0 16"
	})
}), {
	__source: "ph:magnifying-glass-plus-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "zoom-in",
	displayName: "UiZoomInFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiZoomOut.tsx
var mt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M150 112a6 6 0 0 1-6 6H80a6 6 0 0 1 0-12h64a6 6 0 0 1 6 6m78.24 116.24a6 6 0 0 1-8.48 0l-51.38-51.38a86.15 86.15 0 1 1 8.48-8.48l51.38 51.38a6 6 0 0 1 0 8.48M112 186a74 74 0 1 0-74-74a74.09 74.09 0 0 0 74 74"
	})
}), {
	__source: "ph:magnifying-glass-minus-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "zoom-out",
	displayName: "UiZoomOut"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "m229.66 218.34l-50.06-50.06a88.21 88.21 0 1 0-11.32 11.31l50.06 50.07a8 8 0 0 0 11.32-11.32M144 120H80a8 8 0 0 1 0-16h64a8 8 0 0 1 0 16"
	})
}), {
	__source: "ph:magnifying-glass-minus-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "zoom-out",
	displayName: "UiZoomOutFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiDesktop.tsx
var ht = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M208 42H48a22 22 0 0 0-22 22v112a22 22 0 0 0 22 22h74v20H96a6 6 0 0 0 0 12h64a6 6 0 0 0 0-12h-26v-20h74a22 22 0 0 0 22-22V64a22 22 0 0 0-22-22M48 54h160a10 10 0 0 1 10 10v82H38V64a10 10 0 0 1 10-10m160 132H48a10 10 0 0 1-10-10v-18h180v18a10 10 0 0 1-10 10"
	})
}), {
	__source: "ph:desktop-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "desktop",
	displayName: "UiDesktop"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M208 40H48a24 24 0 0 0-24 24v112a24 24 0 0 0 24 24h72v16H96a8 8 0 0 0 0 16h64a8 8 0 0 0 0-16h-24v-16h72a24 24 0 0 0 24-24V64a24 24 0 0 0-24-24m0 144H48a8 8 0 0 1-8-8v-16h176v16a8 8 0 0 1-8 8"
	})
}), {
	__source: "ph:desktop-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "desktop",
	displayName: "UiDesktopFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiStackFrameDot.tsx
var gt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M10 8a2 2 0 1 1-4 0a2 2 0 0 1 4 0"
	})
}), {
	__source: "codicon:debug-stackframe-dot",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "stack-frame-dot",
	displayName: "UiStackFrameDot"
}), _t = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M10 13c0 1.103-.897 2-2 2s-2-.897-2-2s.897-2 2-2s2 .897 2 2m3.25-11a.75.75 0 0 0-.75.75v1.277A5.98 5.98 0 0 0 8 2C5.033 2 2.48 4.211 2.06 7.144a.75.75 0 0 0 1.487.213C3.86 5.159 5.776 3.501 8 3.501c1.529 0 2.919.763 3.743 2H9.75a.75.75 0 0 0 0 1.5h3.5a.75.75 0 0 0 .75-.75v-3.5a.75.75 0 0 0-.75-.75z"
	})
}), {
	__source: "codicon:debug-step-over",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "debug-step-over",
	displayName: "UiDebugStepOver"
}), vt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M128 26a102 102 0 1 0 102 102A102.12 102.12 0 0 0 128 26m0 192a90 90 0 1 1 90-90a90.1 90.1 0 0 1-90 90m47.18-95.09l-64-40A6 6 0 0 0 102 88v80a6 6 0 0 0 9.18 5.09l64-40a6 6 0 0 0 0-10.18M114 157.17V98.83L160.68 128Z"
	})
}), {
	__source: "ph:play-circle-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "play-circle",
	displayName: "UiPlayCircle"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M128 24a104 104 0 1 0 104 104A104.11 104.11 0 0 0 128 24m40.55 110.58l-52 36A8 8 0 0 1 104 164V92a8 8 0 0 1 12.55-6.58l52 36a8 8 0 0 1 0 13.16"
	})
}), {
	__source: "ph:play-circle-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "play-circle",
	displayName: "UiPlayCircleFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiResizeVertical.tsx
var yt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M222 128a6 6 0 0 1-6 6H40a6 6 0 0 1 0-12h176a6 6 0 0 1 6 6m-98.24-27.76a6 6 0 0 0 8.48 0l32-32a6 6 0 0 0-8.48-8.48L134 81.51V16a6 6 0 0 0-12 0v65.51l-21.76-21.75a6 6 0 0 0-8.48 8.48Zm8.48 55.52a6 6 0 0 0-8.48 0l-32 32a6 6 0 0 0 8.48 8.48L122 174.49V240a6 6 0 0 0 12 0v-65.51l21.76 21.75a6 6 0 0 0 8.48-8.48Z"
	})
}), {
	__source: "ph:arrows-in-line-vertical-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "resize-vertical",
	displayName: "UiResizeVertical"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M90.34 69.66A8 8 0 0 1 96 56h24V16a8 8 0 0 1 16 0v40h24a8 8 0 0 1 5.66 13.66l-32 32a8 8 0 0 1-11.32 0Zm43.32 84.68a8 8 0 0 0-11.32 0l-32 32A8 8 0 0 0 96 200h24v40a8 8 0 0 0 16 0v-40h24a8 8 0 0 0 5.66-13.66ZM216 120H40a8 8 0 0 0 0 16h176a8 8 0 0 0 0-16"
	})
}), {
	__source: "ph:arrows-in-line-vertical-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "resize-vertical",
	displayName: "UiResizeVerticalFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiListDashes.tsx
var bt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M90 64a6 6 0 0 1 6-6h120a6 6 0 0 1 0 12H96a6 6 0 0 1-6-6m126 58H96a6 6 0 0 0 0 12h120a6 6 0 0 0 0-12m0 64H96a6 6 0 0 0 0 12h120a6 6 0 0 0 0-12M56 58H40a6 6 0 0 0 0 12h16a6 6 0 0 0 0-12m0 64H40a6 6 0 0 0 0 12h16a6 6 0 0 0 0-12m0 64H40a6 6 0 0 0 0 12h16a6 6 0 0 0 0-12"
	})
}), {
	__source: "ph:list-dashes-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "list-dashes",
	displayName: "UiListDashes"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M208 32H48a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16M72 184h-8a8 8 0 0 1 0-16h8a8 8 0 0 1 0 16m0-48h-8a8 8 0 0 1 0-16h8a8 8 0 0 1 0 16m0-48h-8a8 8 0 0 1 0-16h8a8 8 0 0 1 0 16m120 96h-88a8 8 0 0 1 0-16h88a8 8 0 0 1 0 16m0-48h-88a8 8 0 0 1 0-16h88a8 8 0 0 1 0 16m0-48h-88a8 8 0 0 1 0-16h88a8 8 0 0 1 0 16"
	})
}), {
	__source: "ph:list-dashes-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "list-dashes",
	displayName: "UiListDashesFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiSync.tsx
var xt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M14 3.5v3c0 .28-.22.5-.5.5h-3c-.28 0-.5-.22-.5-.5s.22-.5.5-.5h2.08c-.8-1.83-2.57-3-4.58-3c-2.22 0-4.2 1.5-4.81 3.64c-.06.22-.26.36-.48.36c-.05 0-.09 0-.14-.02a.493.493 0 0 1-.34-.62C2.96 3.79 5.33 2 8 2c2.05 0 3.91 1.02 5 2.69V3.5c0-.28.22-.5.5-.5s.5.22.5.5m-.58 5.52a.51.51 0 0 0-.62.35a5.02 5.02 0 0 1-4.81 3.64c-2.01 0-3.78-1.17-4.58-3h2.08c.28 0 .5-.22.5-.5s-.22-.5-.5-.5h-3c-.28 0-.5.22-.5.5v3c0 .28.22.5.5.5s.5-.22.5-.5v-1.19a5.97 5.97 0 0 0 5 2.69c2.67 0 5.04-1.79 5.77-4.36a.5.5 0 0 0-.35-.62z"
	})
}), {
	__source: "codicon:sync",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "sync",
	displayName: "UiSync"
}), St = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M229.18 173a6 6 0 0 1-2.16 8.2l-96 56a6 6 0 0 1-6 0l-96-56a6 6 0 0 1 6-10.36l93 54.23l93-54.23a6 6 0 0 1 8.16 2.16M221 122.82l-93 54.23l-93-54.23a6 6 0 0 0-6 10.36l96 56a6 6 0 0 0 6 0l96-56a6 6 0 0 0-6-10.36M26 80a6 6 0 0 1 3-5.18l96-56a6 6 0 0 1 6 0l96 56a6 6 0 0 1 0 10.36l-96 56a6 6 0 0 1-6 0l-96-56A6 6 0 0 1 26 80m17.91 0L128 129.05L212.09 80L128 31Z"
	})
}), {
	__source: "ph:stack-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "stack",
	displayName: "UiStack"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsxs)("g", {
		fill: "currentColor",
		children: [
			/* @__PURE__ */ (0, W.jsx)("path", { d: "m220 169.09l-92 53.65l-92-53.65a8 8 0 0 0-8 13.82l96 56a8 8 0 0 0 8.06 0l96-56a8 8 0 1 0-8.06-13.82" }),
			/* @__PURE__ */ (0, W.jsx)("path", { d: "m220 121.09l-92 53.65l-92-53.65a8 8 0 0 0-8 13.82l96 56a8 8 0 0 0 8.06 0l96-56a8 8 0 1 0-8.06-13.82" }),
			/* @__PURE__ */ (0, W.jsx)("path", { d: "m28 86.91l96 56a8 8 0 0 0 8.06 0l96-56a8 8 0 0 0 0-13.82l-96-56a8 8 0 0 0-8.06 0l-96 56a8 8 0 0 0 0 13.82" })
		]
	})
}), {
	__source: "ph:stack-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "stack",
	displayName: "UiStackFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiArrowRight.tsx
var Ct = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "m220.24 132.24l-72 72a6 6 0 0 1-8.48-8.48L201.51 134H40a6 6 0 0 1 0-12h161.51l-61.75-61.76a6 6 0 0 1 8.48-8.48l72 72a6 6 0 0 1 0 8.48"
	})
}), {
	__source: "ph:arrow-right-light",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "arrow-right",
	displayName: "UiArrowRight"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "m221.66 133.66l-72 72A8 8 0 0 1 136 200v-64H40a8 8 0 0 1 0-16h96V56a8 8 0 0 1 13.66-5.66l72 72a8 8 0 0 1 0 11.32"
	})
}), {
	__source: "ph:arrow-right-fill",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "arrow-right",
	displayName: "UiArrowRightFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiExternalLink.tsx
var wt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M222 104a6 6 0 0 1-12 0V54.49l-69.75 69.75a6 6 0 0 1-8.48-8.48L201.51 46H152a6 6 0 0 1 0-12h64a6 6 0 0 1 6 6Zm-38 26a6 6 0 0 0-6 6v72a2 2 0 0 1-2 2H48a2 2 0 0 1-2-2V80a2 2 0 0 1 2-2h72a6 6 0 0 0 0-12H48a14 14 0 0 0-14 14v128a14 14 0 0 0 14 14h128a14 14 0 0 0 14-14v-72a6 6 0 0 0-6-6"
	})
}), {
	__source: "ph:arrow-square-out-light",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "external-link",
	displayName: "UiExternalLink"
}), G = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M208 34h-26V24a6 6 0 0 0-12 0v10H86V24a6 6 0 0 0-12 0v10H48a14 14 0 0 0-14 14v160a14 14 0 0 0 14 14h160a14 14 0 0 0 14-14V48a14 14 0 0 0-14-14M48 46h26v10a6 6 0 0 0 12 0V46h84v10a6 6 0 0 0 12 0V46h26a2 2 0 0 1 2 2v34H46V48a2 2 0 0 1 2-2m160 164H48a2 2 0 0 1-2-2V94h164v114a2 2 0 0 1-2 2m-98-90v64a6 6 0 0 1-12 0v-54.29l-7.32 3.66a6 6 0 1 1-5.36-10.74l16-8A6 6 0 0 1 110 120m59.57 29.25L148 178h20a6 6 0 0 1 0 12h-32a6 6 0 0 1-4.8-9.6L160 142a10 10 0 1 0-16.65-11a6 6 0 1 1-10.35-6a22 22 0 1 1 36.62 24.26Z"
	})
}), {
	__source: "ph:calendar-light",
	__viewBox: "0 0 256 256",
	__group: "forms-editing",
	__consumerName: "calendar",
	displayName: "UiCalendar"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M208 32h-24v-8a8 8 0 0 0-16 0v8H88v-8a8 8 0 0 0-16 0v8H48a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16m-96 152a8 8 0 0 1-16 0v-51.06l-4.42 2.22a8 8 0 0 1-7.16-14.32l16-8A8 8 0 0 1 112 120Zm56-8a8 8 0 0 1 0 16h-32a8 8 0 0 1-6.4-12.8l28.78-38.37a8 8 0 1 0-13.31-8.83a8 8 0 1 1-13.85-8A24 24 0 0 1 176 136a23.76 23.76 0 0 1-4.84 14.45L152 176ZM48 80V48h24v8a8 8 0 0 0 16 0v-8h80v8a8 8 0 0 0 16 0v-8h24v32Z"
	})
}), {
	__source: "ph:calendar-fill",
	__viewBox: "0 0 256 256",
	__group: "forms-editing",
	__consumerName: "calendar",
	displayName: "UiCalendarFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiWarningCircle.tsx
var Tt = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#d97706",
		...r
	};
	return /* @__PURE__ */ (0, W.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, W.jsx)("path", {
			fill: "currentColor",
			d: "M128 26a102 102 0 1 0 102 102A102.12 102.12 0 0 0 128 26m0 192a90 90 0 1 1 90-90a90.1 90.1 0 0 1-90 90m-6-82V80a6 6 0 0 1 12 0v56a6 6 0 0 1-12 0m16 36a10 10 0 1 1-10-10a10 10 0 0 1 10 10"
		})
	});
}, {
	__source: "ph:warning-circle-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "warning-circle",
	__defaultColor: "#d97706",
	displayName: "UiWarningCircle"
});
Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#d97706",
		...r
	};
	return /* @__PURE__ */ (0, W.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, W.jsx)("path", {
			fill: "currentColor",
			d: "M128 24a104 104 0 1 0 104 104A104.11 104.11 0 0 0 128 24m-8 56a8 8 0 0 1 16 0v56a8 8 0 0 1-16 0Zm8 104a12 12 0 1 1 12-12a12 12 0 0 1-12 12"
		})
	});
}, {
	__source: "ph:warning-circle-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "warning-circle",
	__defaultColor: "#d97706",
	displayName: "UiWarningCircleFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiListFlat.tsx
var K = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M222 128a6 6 0 0 1-6 6H40a6 6 0 0 1 0-12h176a6 6 0 0 1 6 6M40 70h176a6 6 0 0 0 0-12H40a6 6 0 0 0 0 12m176 116H40a6 6 0 0 0 0 12h176a6 6 0 0 0 0-12"
	})
}), {
	__source: "ph:list-light",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "list-flat",
	displayName: "UiListFlat"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M208 32H48a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16m-16 152H64a8 8 0 0 1 0-16h128a8 8 0 0 1 0 16m0-48H64a8 8 0 0 1 0-16h128a8 8 0 0 1 0 16m0-48H64a8 8 0 0 1 0-16h128a8 8 0 0 1 0 16"
	})
}), {
	__source: "ph:list-fill",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "list-flat",
	displayName: "UiListFlatFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiClass.tsx
var Et = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, W.jsx)("circle", {
			cx: "8",
			cy: "8",
			r: "6.5",
			fill: "#E7EFFD",
			stroke: "#3574F0"
		}), /* @__PURE__ */ (0, W.jsx)("path", {
			d: "M8.13295 11.5C9.61223 11.5 10.8836 10.6105 11.2075 9.33909H10.2213C9.90229 10.0739 9.11914 10.6057 8.13295 10.6057C6.77936 10.6057 5.80284 9.51796 5.80284 8C5.80284 6.48204 6.77936 5.39434 8.13295 5.39434C9.11914 5.39434 9.90229 5.92611 10.2213 6.66091H11.2075C10.8836 5.3895 9.61223 4.5 8.13295 4.5C6.21859 4.5 4.79248 5.99378 4.79248 8C4.79248 10.0062 6.21859 11.5 8.13295 11.5Z",
			fill: "#3574F0"
		})]
	})
}), {
	__source: "jb-expui-nodes:class",
	__viewBox: "0 0 16 16",
	__group: "uir-ast-code",
	__consumerName: "uir-class",
	displayName: "UiClass"
}), q = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, W.jsx)("circle", {
			cx: "8",
			cy: "8",
			r: "6.5",
			fill: "#FFF7F7",
			stroke: "#DB3B4B"
		}), /* @__PURE__ */ (0, W.jsx)("path", {
			d: "M10.0657 5.24573C9.25361 5.24573 8.636 5.6473 8.29593 6.31537C8.00368 5.64257 7.40552 5.24573 6.6071 5.24573C5.84915 5.24573 5.28147 5.60946 4.96189 6.24862V5.37188H4V10.7543H4.99869V7.59001C4.99869 6.68068 5.5138 6.10775 6.32852 6.10775C7.0749 6.10775 7.50591 6.61761 7.50591 7.43758V10.7543H8.49409V7.59001C8.49409 6.68068 9.01971 6.10775 9.82392 6.10775C10.5756 6.10775 11.0013 6.61761 11.0013 7.43758V10.7543H12V7.35874C12 6.07096 11.2168 5.24573 10.0657 5.24573Z",
			fill: "#DB3B4B"
		})]
	})
}), {
	__source: "jb-expui-nodes:method",
	__viewBox: "0 0 16 16",
	__group: "uir-ast-code",
	__consumerName: "uir-method",
	displayName: "UiMethod"
}), Dt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 32 32",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		d: "M4.014 14.976a2.5 2.5 0 0 0 1.567-.518a2.38 2.38 0 0 0 .805-1.358a15.3 15.3 0 0 0 .214-2.944q.012-2.085.075-2.747a5.2 5.2 0 0 1 .418-1.686a3 3 0 0 1 .755-1.018A3.05 3.05 0 0 1 9 4.125A6.8 6.8 0 0 1 10.544 4h.7v1.96h-.387a2.34 2.34 0 0 0-1.723.468a3.4 3.4 0 0 0-.425 2.092a36 36 0 0 1-.137 4.133a4.7 4.7 0 0 1-.768 2.06A4.6 4.6 0 0 1 6.1 16a3.8 3.8 0 0 1 1.992 1.754a8.9 8.9 0 0 1 .618 3.865q0 2.435.05 2.9a1.76 1.76 0 0 0 .504 1.181a2.64 2.64 0 0 0 1.592.337h.387V28h-.7a5.7 5.7 0 0 1-1.773-.2a2.97 2.97 0 0 1-1.324-.93a3.35 3.35 0 0 1-.681-1.63a24 24 0 0 1-.165-3.234a16.5 16.5 0 0 0-.214-3.106a2.4 2.4 0 0 0-.805-1.361a2.5 2.5 0 0 0-1.567-.524Zm23.972 2.035a2.5 2.5 0 0 0-1.567.524a2.4 2.4 0 0 0-.805 1.361a16.5 16.5 0 0 0-.212 3.109a24 24 0 0 1-.169 3.234a3.35 3.35 0 0 1-.681 1.63a2.97 2.97 0 0 1-1.324.93a5.7 5.7 0 0 1-1.773.2h-.7V26.04h.387a2.64 2.64 0 0 0 1.592-.337a1.76 1.76 0 0 0 .506-1.186q.05-.462.05-2.9a8.9 8.9 0 0 1 .618-3.865A3.8 3.8 0 0 1 25.9 16a4.6 4.6 0 0 1-1.7-1.286a4.7 4.7 0 0 1-.768-2.06a36 36 0 0 1-.137-4.133a3.4 3.4 0 0 0-.425-2.092a2.34 2.34 0 0 0-1.723-.468h-.387V4h.7a6.8 6.8 0 0 1 1.54.125a3.05 3.05 0 0 1 1.149.581a3 3 0 0 1 .755 1.018a5.2 5.2 0 0 1 .418 1.686q.062.662.075 2.747a15.3 15.3 0 0 0 .212 2.947a2.38 2.38 0 0 0 .805 1.355a2.5 2.5 0 0 0 1.567.518Z",
		fill: "#E5A505"
	})
}), {
	__source: "incumbent",
	__viewBox: "0 0 32 32",
	__group: "files-code",
	__consumerName: "json",
	displayName: "UiJson"
}), Ot = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 32 32",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, W.jsx)("path", {
		d: "M2.5 7.955h27v16.091h-27z",
		fill: "none",
		stroke: "#755838"
	}), /* @__PURE__ */ (0, W.jsx)("path", {
		d: "M5.909 20.636v-9.272h2.727l2.728 3.409l2.727-3.409h2.727v9.272h-2.727v-5.318l-2.727 3.409l-2.728-3.409v5.318zm17.046 0l-4.091-4.5h2.727v-4.772h2.727v4.772h2.727z",
		fill: "#755838"
	})]
}), {
	__source: "incumbent",
	__viewBox: "0 0 32 32",
	__group: "files-code",
	__consumerName: "markdown",
	displayName: "UiMarkdown"
}), kt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "m212.24 83.76l-56-56A6 6 0 0 0 152 26H56a14 14 0 0 0-14 14v176a14 14 0 0 0 14 14h144a14 14 0 0 0 14-14V88a6 6 0 0 0-1.76-4.24M158 46.48L193.52 82H158ZM200 218H56a2 2 0 0 1-2-2V40a2 2 0 0 1 2-2h90v50a6 6 0 0 0 6 6h50v122a2 2 0 0 1-2 2m-34-82a6 6 0 0 1-6 6H96a6 6 0 0 1 0-12h64a6 6 0 0 1 6 6m0 32a6 6 0 0 1-6 6H96a6 6 0 0 1 0-12h64a6 6 0 0 1 6 6"
	})
}), {
	__source: "ph:file-text-light",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "file-text",
	displayName: "UiFileText"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "m213.66 82.34l-56-56A8 8 0 0 0 152 24H56a16 16 0 0 0-16 16v176a16 16 0 0 0 16 16h144a16 16 0 0 0 16-16V88a8 8 0 0 0-2.34-5.66M160 176H96a8 8 0 0 1 0-16h64a8 8 0 0 1 0 16m0-32H96a8 8 0 0 1 0-16h64a8 8 0 0 1 0 16m-8-56V44l44 44Z"
	})
}), {
	__source: "ph:file-text-fill",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "file-text",
	displayName: "UiFileTextFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiFileSpreadsheet.tsx
var At = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 24 24",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsxs)("g", {
		fill: "none",
		stroke: "currentColor",
		strokeLinecap: "round",
		strokeLinejoin: "round",
		strokeWidth: "2",
		children: [/* @__PURE__ */ (0, W.jsx)("path", { d: "M6 22a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h8a2.4 2.4 0 0 1 1.704.706l3.588 3.588A2.4 2.4 0 0 1 20 8v12a2 2 0 0 1-2 2z" }), /* @__PURE__ */ (0, W.jsx)("path", { d: "M14 2v5a1 1 0 0 0 1 1h5M8 13h2m4 0h2m-8 4h2m4 0h2" })]
	})
}), {
	__source: "lucide:file-spreadsheet",
	__viewBox: "0 0 24 24",
	__group: "files-code",
	__consumerName: "file-spreadsheet",
	displayName: "UiFileSpreadsheet"
}), jt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M224.91 69.75a6 6 0 0 0-9.63-2.16l-41.07 37.9l-19.51-4.19l-4.19-19.51l37.9-41.07a6 6 0 0 0-2.16-9.63a70 70 0 0 0-89.77 94.39l-61.39 53c-.11.09-.21.19-.32.3a30 30 0 0 0 42.43 42.45c.11-.11.21-.21.3-.32l53-61.39a70 70 0 0 0 94.39-89.77ZM160 154a58 58 0 0 1-28-7.22a6 6 0 0 0-7.45 1.33l-55.98 64.77a18 18 0 0 1-25.45-25.45l64.76-55.94a6 6 0 0 0 1.32-7.49a58 58 0 0 1 64-84.53l-33.62 36.46a6 6 0 0 0-1.45 5.33l5.65 26.35a6 6 0 0 0 4.61 4.61l26.35 5.65a6 6 0 0 0 5.33-1.45l36.42-33.62A58.06 58.06 0 0 1 160 154"
	})
}), {
	__source: "ph:wrench-light",
	__viewBox: "0 0 256 256",
	__group: "dev-tools",
	__consumerName: "wrench",
	displayName: "UiWrench"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M232 96a72 72 0 0 1-100.94 66L79 222.22c-.12.14-.26.29-.39.42a32 32 0 0 1-45.26-45.26c.14-.13.28-.27.43-.39L94 124.94a72.07 72.07 0 0 1 83.54-98.78a8 8 0 0 1 3.93 13.19L144 80l5.66 26.35L176 112l40.65-37.52a8 8 0 0 1 13.19 3.93A72.6 72.6 0 0 1 232 96"
	})
}), {
	__source: "ph:wrench-fill",
	__viewBox: "0 0 256 256",
	__group: "dev-tools",
	__consumerName: "wrench",
	displayName: "UiWrenchFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiCloudDownload.tsx
var Mt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M246 128a85.27 85.27 0 0 1-17.2 51.6a6 6 0 1 1-9.6-7.2A74 74 0 1 0 86 128a6 6 0 0 1-12 0a85.5 85.5 0 0 1 3.91-25.64A51 51 0 0 0 72 102a50 50 0 0 0 0 100h24a6 6 0 0 1 0 12H72A62 62 0 1 1 82.43 90.88A86 86 0 0 1 246 128m-66.24 43.76L158 193.51V128a6 6 0 0 0-12 0v65.51l-21.76-21.75a6 6 0 0 0-8.48 8.48l32 32a6 6 0 0 0 8.48 0l32-32a6 6 0 0 0-8.48-8.48"
	})
}), {
	__source: "ph:cloud-arrow-down-light",
	__viewBox: "0 0 256 256",
	__group: "infrastructure",
	__consumerName: "cloud-download",
	displayName: "UiCloudDownload"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M247.93 124.52C246.11 77.54 207.07 40 160.06 40a88.1 88.1 0 0 0-78.77 48.67A87.5 87.5 0 0 0 72 127.73a8.18 8.18 0 0 1-7.43 8.27a8 8 0 0 1-8.57-8a103.7 103.7 0 0 1 5.34-32.92a4 4 0 0 0-4.75-5.18A64.09 64.09 0 0 0 8 152c0 35.19 29.75 64 65 64h87a88.09 88.09 0 0 0 87.93-91.48m-50.27 25.14l-32 32a8 8 0 0 1-11.32 0l-32-32a8 8 0 0 1 11.32-11.32L152 156.69V96a8 8 0 0 1 16 0v60.69l18.34-18.35a8 8 0 0 1 11.32 11.32"
	})
}), {
	__source: "ph:cloud-arrow-down-fill",
	__viewBox: "0 0 256 256",
	__group: "infrastructure",
	__consumerName: "cloud-download",
	displayName: "UiCloudDownloadFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiChip.tsx
var Nt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M152 98h-48a6 6 0 0 0-6 6v48a6 6 0 0 0 6 6h48a6 6 0 0 0 6-6v-48a6 6 0 0 0-6-6m-6 48h-36v-36h36Zm86 0h-18v-36h18a6 6 0 0 0 0-12h-18V56a14 14 0 0 0-14-14h-42V24a6 6 0 0 0-12 0v18h-36V24a6 6 0 0 0-12 0v18H56a14 14 0 0 0-14 14v42H24a6 6 0 0 0 0 12h18v36H24a6 6 0 0 0 0 12h18v42a14 14 0 0 0 14 14h42v18a6 6 0 0 0 12 0v-18h36v18a6 6 0 0 0 12 0v-18h42a14 14 0 0 0 14-14v-42h18a6 6 0 0 0 0-12m-30 54a2 2 0 0 1-2 2H56a2 2 0 0 1-2-2V56a2 2 0 0 1 2-2h144a2 2 0 0 1 2 2Z"
	})
}), {
	__source: "ph:cpu-light",
	__viewBox: "0 0 256 256",
	__group: "infrastructure",
	__consumerName: "chip",
	displayName: "UiChip"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M104 104h48v48h-48Zm136 48a8 8 0 0 1-8 8h-16v40a16 16 0 0 1-16 16h-40v16a8 8 0 0 1-16 0v-16h-32v16a8 8 0 0 1-16 0v-16H56a16 16 0 0 1-16-16v-40H24a8 8 0 0 1 0-16h16v-32H24a8 8 0 0 1 0-16h16V56a16 16 0 0 1 16-16h40V24a8 8 0 0 1 16 0v16h32V24a8 8 0 0 1 16 0v16h40a16 16 0 0 1 16 16v40h16a8 8 0 0 1 0 16h-16v32h16a8 8 0 0 1 8 8m-72-56a8 8 0 0 0-8-8H96a8 8 0 0 0-8 8v64a8 8 0 0 0 8 8h64a8 8 0 0 0 8-8Z"
	})
}), {
	__source: "ph:cpu-fill",
	__viewBox: "0 0 256 256",
	__group: "infrastructure",
	__consumerName: "chip",
	displayName: "UiChipFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiRefresh.tsx
var Pt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M222 48v48a6 6 0 0 1-6 6h-48a6 6 0 0 1 0-12h33.52l-18.05-18a81.5 81.5 0 0 0-57.53-24h-.46a81.5 81.5 0 0 0-57.29 23.28a6 6 0 1 1-8.38-8.58a93.38 93.38 0 0 1 65.67-26.76h.52a93.45 93.45 0 0 1 66 27.53l18 18V48a6 6 0 0 1 12 0m-34.19 136.72a81.5 81.5 0 0 1-57.29 23.34h-.46a81.5 81.5 0 0 1-57.53-24L54.48 166H88a6 6 0 0 0 0-12H40a6 6 0 0 0-6 6v48a6 6 0 0 0 12 0v-33.52l18 18.05a93.45 93.45 0 0 0 66 27.53h.52a93.38 93.38 0 0 0 65.67-26.76a6 6 0 1 0-8.38-8.58"
	})
}), {
	__source: "ph:arrows-clockwise-light",
	__viewBox: "0 0 256 256",
	__group: "actions-tools",
	__consumerName: "refresh",
	displayName: "UiRefresh"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M224 48v48a8 8 0 0 1-8 8h-48a8 8 0 0 1-5.66-13.66L180.65 72a79.48 79.48 0 0 0-54.72-22.09h-.45a79.52 79.52 0 0 0-55.89 22.8a8 8 0 0 1-11.18-11.44A96 96 0 0 1 192 60.7l18.36-18.36A8 8 0 0 1 224 48m-37.59 135.29a80 80 0 0 1-111.06.71l18.31-18.31A8 8 0 0 0 88 152H40a8 8 0 0 0-8 8v48a8 8 0 0 0 13.66 5.66L64 195.3a95.42 95.42 0 0 0 66 26.76h.53a95.36 95.36 0 0 0 67.07-27.33a8 8 0 0 0-11.18-11.44Z"
	})
}), {
	__source: "ph:arrows-clockwise-fill",
	__viewBox: "0 0 256 256",
	__group: "actions-tools",
	__consumerName: "refresh",
	displayName: "UiRefreshFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiClock.tsx
var J = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M128 26a102 102 0 1 0 102 102A102.12 102.12 0 0 0 128 26m0 192a90 90 0 1 1 90-90a90.1 90.1 0 0 1-90 90m62-90a6 6 0 0 1-6 6h-56a6 6 0 0 1-6-6V72a6 6 0 0 1 12 0v50h50a6 6 0 0 1 6 6"
	})
}), {
	__source: "ph:clock-light",
	__viewBox: "0 0 256 256",
	__group: "time",
	__consumerName: "clock",
	displayName: "UiClock"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M128 24a104 104 0 1 0 104 104A104.11 104.11 0 0 0 128 24m56 112h-56a8 8 0 0 1-8-8V72a8 8 0 0 1 16 0v48h48a8 8 0 0 1 0 16"
	})
}), {
	__source: "ph:clock-fill",
	__viewBox: "0 0 256 256",
	__group: "time",
	__consumerName: "clock",
	displayName: "UiClockFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiWatch.tsx
var Ft = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M128 42a94 94 0 1 0 94 94a94.11 94.11 0 0 0-94-94m0 176a82 82 0 1 1 82-82a82.1 82.1 0 0 1-82 82m44.24-126.24a6 6 0 0 1 0 8.48l-40 40a6 6 0 1 1-8.48-8.48l40-40a6 6 0 0 1 8.48 0M98 16a6 6 0 0 1 6-6h48a6 6 0 0 1 0 12h-48a6 6 0 0 1-6-6"
	})
}), {
	__source: "ph:timer-light",
	__viewBox: "0 0 256 256",
	__group: "time",
	__consumerName: "watch",
	displayName: "UiWatch"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M128 40a96 96 0 1 0 96 96a96.11 96.11 0 0 0-96-96m45.66 61.66l-40 40a8 8 0 0 1-11.32-11.32l40-40a8 8 0 0 1 11.32 11.32M96 16a8 8 0 0 1 8-8h48a8 8 0 0 1 0 16h-48a8 8 0 0 1-8-8"
	})
}), {
	__source: "ph:timer-fill",
	__viewBox: "0 0 256 256",
	__group: "time",
	__consumerName: "watch",
	displayName: "UiWatchFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiHourglass.tsx
var It = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M198 75.64V40a14 14 0 0 0-14-14H72a14 14 0 0 0-14 14v36a14.06 14.06 0 0 0 5.6 11.2L118 128l-54.4 40.8A14.06 14.06 0 0 0 58 180v36a14 14 0 0 0 14 14h112a14 14 0 0 0 14-14v-35.64a14.08 14.08 0 0 0-5.56-11.17L138 128l54.49-41.19A14.08 14.08 0 0 0 198 75.64M70 40a2 2 0 0 1 2-2h112a2 2 0 0 1 2 2v35.64a2 2 0 0 1-.79 1.6L178.9 82H76.67l-5.87-4.4A2 2 0 0 1 70 76Zm58 80.49L92.67 94H163Zm58 59.87V216a2 2 0 0 1-2 2H72a2 2 0 0 1-2-2v-36a2 2 0 0 1 .8-1.6L122 140v28a6 6 0 0 0 12 0v-27.94l51.21 38.7a2 2 0 0 1 .79 1.6"
	})
}), {
	__source: "ph:hourglass-medium-light",
	__viewBox: "0 0 256 256",
	__group: "time",
	__consumerName: "hourglass",
	displayName: "UiHourglass"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, W.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, W.jsx)("path", {
		fill: "currentColor",
		d: "M200 75.64V40a16 16 0 0 0-16-16H72a16 16 0 0 0-16 16v36a16.07 16.07 0 0 0 6.4 12.8l52.27 39.2l-52.27 39.2A16.07 16.07 0 0 0 56 180v36a16 16 0 0 0 16 16h112a16 16 0 0 0 16-16v-35.64a16.09 16.09 0 0 0-6.35-12.77L141.27 128l52.38-39.59A16.09 16.09 0 0 0 200 75.64m-16 104.72V216H72v-36l48-36v24a8 8 0 0 0 16 0v-23.92Zm0-104.72L178.23 80H77.33L72 76V40h112Z"
	})
}), {
	__source: "ph:hourglass-medium-fill",
	__viewBox: "0 0 256 256",
	__group: "time",
	__consumerName: "hourglass",
	displayName: "UiHourglassFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/data-table-utils.ts
function Lt(e, t) {
	return t.split(".").reduce((e, t) => {
		if (e && typeof e == "object") return e[t];
	}, e);
}
function Rt(e) {
	if (Array.isArray(e)) return e.flatMap((e) => Rt(e)).filter(Boolean);
	if (e == null) return [];
	if (typeof e == "object") return [JSON.stringify(e)];
	let t = String(e).trim();
	return t ? [t] : [];
}
function zt(e, t) {
	return t.accessor ? t.accessor(e) : Lt(e, t.key);
}
function Bt(e, t) {
	let n = zt(e, t);
	return t.filterValue ? t.filterValue(n, e) : n;
}
function Vt(e, t) {
	return Rt(Bt(e, t));
}
function Ht(e) {
	return e.replace(/([a-z])([A-Z])/g, "$1 $2").replace(/[_-]/g, " ").replace(/\b\w/g, (e) => e.toUpperCase()).trim();
}
function Ut(e) {
	return typeof e.label == "string" ? e.label : Ht(e.key.split(".").at(-1) ?? e.key);
}
function Wt(e) {
	return e.hideable !== !1;
}
function Gt(e, t) {
	if (e.length === 0) return [];
	let n = t?.literalKeys ?? !1, r = /* @__PURE__ */ new Set();
	for (let t of e.slice(0, 20)) Object.keys(t).filter((e) => !e.startsWith("_")).forEach((e) => r.add(e));
	return Array.from(r).map((e, t) => ({
		key: e || `__col_${t}`,
		label: e ? Ht(e) : "(no column name)",
		sortable: !0,
		filterable: !0,
		...n ? { accessor: (t) => t[e] } : {}
	}));
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/filter-bar-utils.ts
function Kt(e) {
	let { key: t, kind: n, label: r, icon: i, description: a, ...o } = e;
	return o;
}
function qt(e) {
	e.kind === "text" || e.kind === "lookup" || e.kind === "workload" || e.kind === "enum" ? e.onChange("") : e.kind === "lookup-multi" || e.kind === "select-multi" ? e.onChange([]) : e.kind === "number" || e.kind === "duration" ? e.onChange({}) : e.kind === "boolean" ? e.onChange(!1) : e.kind === "tristate" ? e.onChange(void 0) : e.kind === "date-range" ? e.onApply("", "") : e.onChange({});
}
function Jt(e) {
	return e.kind === "date-range" ? String(e.from ?? "").trim() !== "" || String(e.to ?? "").trim() !== "" : e.kind === "text" || e.kind === "lookup" || e.kind === "workload" || e.kind === "enum" ? String(e.value ?? "").trim() !== "" : e.kind === "lookup-multi" || e.kind === "select-multi" ? e.value.length > 0 : e.kind === "number" || e.kind === "duration" ? String(e.value.min ?? "").trim() !== "" || String(e.value.max ?? "").trim() !== "" : e.kind === "boolean" ? e.value : e.kind === "tristate" ? e.value !== void 0 : Object.keys(e.value).length > 0;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/data-table-server-filters.ts
function Yt(e) {
	if (e.autoFilter && e.manualFilter) throw Error("DataTable: autoFilter generates filters that narrow rows locally, which manualFilter then suppresses — the controls would render and do nothing. Pass externalFilters instead.");
	if (e.manualFilter && e.showGlobalFilter && !e.globalFilterControlled && !e.hasExternalSearch) throw Error("DataTable: manualFilter suppresses the built-in search's row narrowing, so showGlobalFilter needs globalFilter/onGlobalFilterChange or externalSearch to reach the source.");
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/hooks/use-sort.ts
function Xt(e, t) {
	return t.split(".").reduce((e, t) => {
		if (e && typeof e == "object") return e[t];
	}, e);
}
function Zt(e, t = {}) {
	let { defaultKey: n, defaultDir: r = "asc", sort: i, onSortChange: a, manual: o = !1, resolvers: s } = t, [c, l] = (0, U.useState)(n ? {
		key: n,
		dir: r
	} : null), u = i !== void 0, d = u ? i : c;
	function f(e) {
		let t = d?.key === e ? d.dir === "asc" ? {
			key: e,
			dir: "desc"
		} : null : {
			key: e,
			dir: "asc"
		};
		u ? a?.(t) : l(t);
	}
	return {
		sorted: (0, U.useMemo)(() => {
			if (!e) return [];
			if (o || !d) return e;
			let { key: t, dir: n } = d, r = s?.[t], i = (e) => r ? r(e) : Xt(e, t);
			return [...e].sort((e, t) => {
				let r = i(e), a = i(t);
				if (r == null && a == null) return 0;
				if (r == null) return 1;
				if (a == null) return -1;
				let o = typeof r == "number" && typeof a == "number" ? r - a : String(r).localeCompare(String(a));
				return n === "asc" ? o : -o;
			});
		}, [
			e,
			o,
			d,
			s
		]),
		sort: d,
		toggle: f
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/hooks/density-provider.tsx
function Qt({ density: e, children: t }) {
	let n = (0, U.useMemo)(() => ({
		density: e,
		setDensity: () => {}
	}), [e]);
	return /* @__PURE__ */ (0, W.jsx)(b.Provider, {
		value: n,
		children: t
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/filter-bar-tree-model.ts
function $t(e) {
	let t = /* @__PURE__ */ new Map(), n = (e) => {
		t.set(e.key, e.label), e.children?.forEach(n);
	};
	return e.forEach(n), t;
}
function en(e, t) {
	let n = $t(e), r = (e) => Object.keys(t).filter((n) => t[n] === e).map((e) => n.get(e) ?? e);
	return [...r("include"), ...r("exclude").map((e) => `−${e}`)].join(", ");
}
function tn(e, t) {
	return (e.children ?? []).some((e) => t[e.key] !== void 0 || tn(e, t));
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/TreePickerField.tsx
var nn = 480, rn = 480, an = 200, on = 8;
function sn({ roots: e, getKey: t, getChildren: n, getAriaLabel: r, renderRow: a, getSearchText: s, defaultOpen: c, hasMoreChildren: l, loadChildren: u, isSelectable: d, onSelect: f, closeOnSelect: m = !0, selected: g, revealSelected: v, showControls: y, ariaLabel: b, empty: x, label: S, placeholder: C = "Select…", disabled: w, size: T, className: E, triggerClassName: D, panelClassName: O, renderTrigger: k, renderFooter: A, triggerAriaLabel: j }) {
	let [M, N] = (0, U.useState)(!1), ee = _(), [P, te] = (0, U.useState)(null), ne = (0, U.useRef)(null), F = (0, U.useRef)(null), I = (0, U.useRef)(null);
	return h(M, () => {
		N(!1), F.current?.focus();
	}), (0, U.useLayoutEffect)(() => {
		if (!M) {
			te(null);
			return;
		}
		let e = () => {
			let e = ne.current;
			if (!e) return;
			let t = e.getBoundingClientRect(), n = window.innerWidth - t.left - on, r = Math.max(t.width, Math.min(nn, n)), i = window.innerHeight - t.bottom - on - 4, a = t.top - on - 4, o = i < an && a > i, s = Math.min(rn, Math.max(o ? a : i, an));
			te({
				top: o ? Math.max(on, t.top - 4 - s) : t.bottom + 4,
				left: t.left,
				width: t.width,
				maxWidth: r,
				maxHeight: s
			});
		};
		return e(), window.addEventListener("scroll", e, !0), window.addEventListener("resize", e), () => {
			window.removeEventListener("scroll", e, !0), window.removeEventListener("resize", e);
		};
	}, [M]), (0, U.useEffect)(() => {
		if (!M) return;
		let e = (e) => {
			let t = e.target;
			!ne.current?.contains(t) && !I.current?.contains(t) && N(!1);
		};
		return document.addEventListener("mousedown", e), () => document.removeEventListener("mousedown", e);
	}, [M]), /* @__PURE__ */ (0, W.jsxs)("div", {
		ref: ne,
		className: i("relative", E),
		children: [k ? k({
			open: M,
			disabled: !!w,
			triggerRef: F,
			toggle: () => {
				w || N((e) => !e);
			}
		}) : /* @__PURE__ */ (0, W.jsxs)("button", {
			ref: F,
			type: "button",
			disabled: w,
			"aria-label": j,
			"aria-haspopup": "tree",
			"aria-expanded": M,
			onClick: () => N((e) => !e),
			className: i("relative flex w-full items-center rounded-md border border-input bg-background text-left text-foreground", T ? o[T] : "h-control-h px-control-px text-sm", "pr-8", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1", "disabled:cursor-not-allowed disabled:opacity-50", D),
			children: [/* @__PURE__ */ (0, W.jsx)("span", {
				className: i("min-w-0 flex-1 truncate", S == null && "text-muted-foreground"),
				children: S ?? C
			}), /* @__PURE__ */ (0, W.jsx)("span", {
				className: "pointer-events-none absolute right-0 flex h-full items-center px-2 text-muted-foreground",
				children: /* @__PURE__ */ (0, W.jsx)(p, {
					icon: z,
					className: "text-xs"
				})
			})]
		}), M && P && (0, at.createPortal)(/* @__PURE__ */ (0, W.jsxs)("div", {
			ref: I,
			"data-slot": "tree-picker-popup",
			style: {
				position: "fixed",
				top: P.top,
				left: P.left,
				zIndex: ee,
				minWidth: P.width,
				width: "max-content",
				maxWidth: P.maxWidth,
				maxHeight: P.maxHeight
			},
			className: i("flex flex-col overflow-hidden rounded-md border border-border bg-popover shadow-md", O),
			children: [/* @__PURE__ */ (0, W.jsx)("div", {
				className: "min-h-0 flex-1 overflow-auto",
				children: /* @__PURE__ */ (0, W.jsx)(xe, {
					roots: e,
					getKey: t,
					getChildren: n,
					renderRow: a,
					...r ? { getAriaLabel: r } : {},
					...s ? { getSearchText: s } : {},
					...c ? { defaultOpen: c } : {},
					...l ? { hasMoreChildren: l } : {},
					...u ? { loadChildren: u } : {},
					...g === void 0 ? {} : { selected: g },
					...v === void 0 ? {} : { revealSelected: v },
					...y === void 0 ? {} : { showControls: y },
					...b === void 0 ? {} : { ariaLabel: b },
					...x === void 0 ? {} : { empty: x },
					onSelect: (e) => {
						(!d || d(e)) && (f(e), m && N(!1));
					}
				})
			}), A && /* @__PURE__ */ (0, W.jsx)("div", {
				className: "shrink-0 border-t border-border bg-popover",
				children: A({ close: () => N(!1) })
			})]
		}), document.body)]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/filter-bar-tree.tsx
var cn = (e) => (e.children?.length ?? 0) > 0;
function ln({ filter: e, grow: t }) {
	let [n, r] = $e(e.value, e.onChange), a = (0, U.useMemo)(() => en(e.nodes, n), [e.nodes, n]), { label: s, ariaLabel: c } = Ee(e), l = (e, t) => r(Ae(n, e, t));
	return /* @__PURE__ */ (0, W.jsx)(sn, {
		roots: e.nodes,
		getKey: (e) => e.key,
		getChildren: (e) => e.children,
		getSearchText: (e) => [
			e.label,
			e.key,
			e.description
		].filter(Boolean).join(" "),
		defaultOpen: (e) => tn(e, n),
		isSelectable: (e) => !cn(e),
		onSelect: (e) => l(e.key, Se(n[e.key] ?? "neutral")),
		closeOnSelect: !1,
		ariaLabel: e.label,
		size: "sm",
		...e.disabled === void 0 ? {} : { disabled: e.disabled },
		className: i(De(t), e.className),
		renderTrigger: ({ open: t, disabled: n, triggerRef: r, toggle: l }) => /* @__PURE__ */ (0, W.jsxs)("button", {
			ref: r,
			type: "button",
			disabled: n,
			"aria-label": c ?? e.label,
			"aria-haspopup": "tree",
			"aria-expanded": t,
			title: a || e.description,
			onClick: l,
			className: i("relative flex w-full items-center gap-1.5 rounded-md border border-input bg-background pr-8 text-left text-foreground", o.sm, "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1", "disabled:cursor-not-allowed disabled:opacity-50"),
			children: [
				/* @__PURE__ */ (0, W.jsx)("span", {
					className: "shrink-0 whitespace-nowrap text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
					children: s
				}),
				/* @__PURE__ */ (0, W.jsx)("span", {
					className: i("min-w-0 flex-1 truncate", !a && "text-placeholder"),
					children: a || e.placeholder
				}),
				/* @__PURE__ */ (0, W.jsx)("span", {
					className: "pointer-events-none absolute right-0 flex h-full items-center px-2 text-muted-foreground",
					children: /* @__PURE__ */ (0, W.jsx)(p, {
						icon: z,
						className: "text-xs"
					})
				})
			]
		}),
		renderRow: ({ node: e }) => /* @__PURE__ */ (0, W.jsx)(Ne, {
			className: "min-w-0 flex-1 justify-between",
			togglePosition: "right",
			mode: n[e.key] ?? "neutral",
			title: e.label,
			label: /* @__PURE__ */ (0, W.jsxs)("span", {
				className: "block min-w-0 text-left",
				children: [/* @__PURE__ */ (0, W.jsx)("span", {
					className: "block truncate",
					children: e.label
				}), e.description && /* @__PURE__ */ (0, W.jsx)("span", {
					className: "mt-0.5 block truncate text-xs font-normal text-muted-foreground",
					children: e.description
				})]
			}),
			onModeChange: (t) => l(e.key, t)
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/hooks/use-container-width.ts
function un(e) {
	let [t, n] = (0, U.useState)(null), [r, i] = (0, U.useState)(!0);
	return (0, U.useLayoutEffect)(() => {
		if (!t) return;
		let n = () => {
			let n = t.getBoundingClientRect().width, r = n <= 0 || n >= e;
			i((e) => e === r ? e : r);
		};
		n();
		let r = typeof ResizeObserver > "u" ? null : ResizeObserver;
		if (!r) return window.addEventListener("resize", n), () => window.removeEventListener("resize", n);
		let a = new r(n);
		return a.observe(t), window.addEventListener("resize", n), () => {
			a.disconnect(), window.removeEventListener("resize", n);
		};
	}, [t, e]), {
		ref: n,
		wider: r
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/DateTimePicker.tsx
var dn = (0, U.forwardRef)(({ value: e = "", onChange: t, className: n, inputClassName: r, buttonClassName: a, openButtonLabel: o = "Open time picker", suffix: s, prefix: c, ...l }, u) => {
	let d = (0, U.useRef)(null), f = (0, U.useRef)(null);
	function m(e) {
		d.current = e, typeof u == "function" ? u(e) : u && (u.current = e);
	}
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		"data-jsf-control": !0,
		className: i("relative", n),
		children: [
			c && /* @__PURE__ */ (0, W.jsx)("div", {
				className: "absolute inset-y-0 left-1.5 flex items-center",
				children: c
			}),
			/* @__PURE__ */ (0, W.jsx)("input", {
				...l,
				ref: m,
				type: "text",
				value: e,
				className: i("h-control-h w-full rounded-md border border-input bg-background px-control-px pr-8 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50", c && "pl-8", r),
				onChange: (e) => t?.(e.target.value)
			}),
			/* @__PURE__ */ (0, W.jsx)("input", {
				ref: f,
				type: "datetime-local",
				tabIndex: -1,
				"aria-hidden": "true",
				className: "pointer-events-none absolute left-0 top-0 h-0 w-0 opacity-0",
				value: fn(e),
				onChange: (e) => t?.(e.target.value)
			}),
			s && /* @__PURE__ */ (0, W.jsx)("div", {
				className: "absolute inset-y-0 right-7 flex items-center",
				children: s
			}),
			/* @__PURE__ */ (0, W.jsx)("button", {
				type: "button",
				"aria-label": o,
				disabled: l.disabled,
				className: i("absolute inset-y-0 right-1 inline-flex items-center text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50", a),
				onClick: () => {
					d.current?.focus(), f.current?.showPicker?.();
				},
				children: /* @__PURE__ */ (0, W.jsx)(p, {
					icon: G,
					className: "text-sm"
				})
			})
		]
	});
});
dn.displayName = "DateTimePicker";
function fn(e) {
	let t = e.trim();
	if (!t || t === "now" || t.startsWith("now")) return "";
	let n = /^\d{4}-\d{2}-\d{2}$/.test(t) ? `${t}T00:00` : t;
	return /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}/.test(n) ? n.slice(0, 16) : "";
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/select.tsx
var pn = (0, U.forwardRef)(({ className: e, options: t, placeholder: n, children: r, ...a }, o) => /* @__PURE__ */ (0, W.jsxs)("div", {
	className: "relative inline-flex w-full items-center",
	children: [/* @__PURE__ */ (0, W.jsxs)("select", {
		ref: o,
		className: i("h-control-h w-full appearance-none rounded-md border border-input bg-background px-control-px pr-8 text-sm text-foreground", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1", "disabled:cursor-not-allowed disabled:opacity-50", e),
		...a,
		children: [n !== void 0 && /* @__PURE__ */ (0, W.jsx)("option", {
			value: "",
			disabled: !0,
			children: n
		}), t ? t.map((e) => /* @__PURE__ */ (0, W.jsx)("option", {
			value: e.value,
			disabled: e.disabled,
			children: e.label
		}, e.value)) : r]
	}), /* @__PURE__ */ (0, W.jsx)(p, {
		icon: z,
		className: "pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground"
	})]
}));
pn.displayName = "Select";
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/MultiSelect.tsx
function mn({ options: e, value: t, onChange: n, placeholder: r = "Select options", ariaLabel: o, disabled: s, className: c, triggerClassName: l, menuClassName: d, facets: m = [] }) {
	let [g, v] = (0, U.useState)(!1), [b, S] = (0, U.useState)({}), C = e.filter((e) => m.every((t) => !b[t.key] || e.facets?.[t.key] === b[t.key])), w = _(), { refs: T, floatingStyles: E, context: D } = y({
		open: g,
		onOpenChange: v,
		placement: "bottom-start",
		whileElementsMounted: tt,
		middleware: [
			et(6),
			nt({ padding: 8 }),
			rt({ padding: 8 })
		]
	}), { getReferenceProps: O, getFloatingProps: k } = ne([
		f(D),
		u(D, { escapeKey: !1 }),
		x(D, { role: "menu" })
	]);
	h(g, () => {
		v(!1), T.domReference.current instanceof HTMLElement && T.domReference.current.focus();
	});
	let A = (0, U.useMemo)(() => {
		let n = e.filter((e) => t.includes(e.value)), i = n.map((e) => e.label).filter((e) => typeof e == "string");
		return n.length === 0 ? r : i.length === 0 ? `${n.length} selected` : i.length <= 2 ? i.join(", ") : `${i.length} selected`;
	}, [
		e,
		r,
		t
	]);
	function j(e) {
		t.includes(e) ? n(t.filter((t) => t !== e)) : n([...t, e]);
	}
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i("inline-flex min-w-0", c),
		children: [/* @__PURE__ */ (0, W.jsxs)(a, {
			ref: T.setReference,
			type: "button",
			variant: "outline",
			disabled: s,
			"aria-label": o ?? `${r} filter`,
			className: i("w-fit max-w-[15rem] min-w-0 justify-between gap-3 text-left font-normal", l, t.length === 0 && "text-placeholder"),
			...O(),
			children: [/* @__PURE__ */ (0, W.jsx)("span", {
				className: "truncate",
				children: A
			}), /* @__PURE__ */ (0, W.jsx)(p, {
				icon: g ? _e : z,
				className: "shrink-0 text-muted-foreground"
			})]
		}), g && /* @__PURE__ */ (0, W.jsx)(pe, { children: /* @__PURE__ */ (0, W.jsx)(he, {
			context: D,
			modal: !1,
			children: /* @__PURE__ */ (0, W.jsxs)("div", {
				ref: T.setFloating,
				role: "menu",
				"aria-label": r,
				style: {
					...E,
					zIndex: w
				},
				className: i("min-w-[14rem] max-w-[20rem] rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-lg shadow-black/5", d),
				...k(),
				children: [
					/* @__PURE__ */ (0, W.jsxs)("div", {
						className: "mb-1 flex items-center justify-between gap-2 px-2 py-1 text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
						children: [/* @__PURE__ */ (0, W.jsx)("span", { children: r }), /* @__PURE__ */ (0, W.jsx)("button", {
							type: "button",
							className: "text-[10px] text-primary disabled:text-muted-foreground",
							onClick: () => n([]),
							disabled: t.length === 0,
							children: "Clear all"
						})]
					}),
					m.length > 0 && /* @__PURE__ */ (0, W.jsxs)("div", {
						className: "space-y-2 border-b border-border px-2 pb-2",
						children: [m.map((t) => {
							let n = /* @__PURE__ */ new Map();
							for (let r of e) {
								let e = r.facets?.[t.key];
								e !== void 0 && n.set(e, (n.get(e) ?? 0) + 1);
							}
							return /* @__PURE__ */ (0, W.jsxs)("div", {
								className: "space-y-1",
								children: [/* @__PURE__ */ (0, W.jsx)("span", {
									className: "text-xs font-medium",
									children: t.label
								}), /* @__PURE__ */ (0, W.jsx)(pn, {
									"aria-label": t.label,
									value: b[t.key] ?? "",
									onChange: (e) => S((n) => ({
										...n,
										[t.key]: e.target.value
									})),
									options: [{
										value: "",
										label: `All (${e.length})`
									}, ...[...n].map(([e, n]) => ({
										value: e,
										label: `${t.labels?.[e] ?? e} (${n})`
									}))]
								})]
							}, t.key);
						}), /* @__PURE__ */ (0, W.jsx)(a, {
							type: "button",
							variant: "ghost",
							size: "sm",
							onClick: () => n([.../* @__PURE__ */ new Set([...t, ...C.filter((e) => !e.disabled).map((e) => e.value)])]),
							children: "Select all shown"
						})]
					}),
					/* @__PURE__ */ (0, W.jsxs)("div", {
						className: "max-h-64 overflow-auto",
						children: [C.length === 0 && /* @__PURE__ */ (0, W.jsx)("p", {
							role: "status",
							className: "px-2 py-3 text-sm text-muted-foreground",
							children: "No matching options"
						}), C.map((e) => {
							let n = t.includes(e.value);
							return /* @__PURE__ */ (0, W.jsxs)("label", {
								className: i("flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm hover:bg-accent", e.disabled && "cursor-not-allowed opacity-50"),
								children: [/* @__PURE__ */ (0, W.jsx)("input", {
									type: "checkbox",
									role: "menuitemcheckbox",
									className: "size-4 rounded border border-input",
									checked: n,
									disabled: e.disabled,
									onChange: () => j(e.value)
								}), /* @__PURE__ */ (0, W.jsx)("span", {
									className: "min-w-0 truncate",
									children: e.label
								})]
							}, e.value);
						})]
					})
				]
			})
		}) })]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/slider-shared.ts
var hn = "pointer-events-none absolute inset-0 h-6 w-full appearance-none bg-transparent [&::-webkit-slider-runnable-track]:h-1 [&::-webkit-slider-runnable-track]:rounded-full [&::-webkit-slider-runnable-track]:bg-transparent [&::-moz-range-track]:h-1 [&::-moz-range-track]:rounded-full [&::-moz-range-track]:bg-transparent", gn = "[&::-webkit-slider-thumb]:pointer-events-auto [&::-webkit-slider-thumb]:mt-[-5px] [&::-webkit-slider-thumb]:size-3.5 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:border [&::-webkit-slider-thumb]:border-background/80 [&::-webkit-slider-thumb]:bg-muted-foreground [&::-webkit-slider-thumb]:shadow-sm [&::-moz-range-thumb]:pointer-events-auto [&::-moz-range-thumb]:size-3.5 [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:border [&::-moz-range-thumb]:border-background/80 [&::-moz-range-thumb]:bg-muted-foreground [&::-moz-range-thumb]:shadow-sm", _n = "h-7 min-w-0 rounded-md border border-input bg-background px-2 text-right font-mono text-xs tabular-nums text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring";
function vn(e, t, n) {
	return Math.min(n, Math.max(t, e));
}
function yn(e, { base: t, step: n, min: r, max: i }) {
	if (!(n > 0)) return vn(e, r, i);
	let a = t + Math.round((e - t) / n) * n;
	return a < r ? vn(t + Math.ceil((r - t) / n) * n, r, i) : a > i ? vn(t + Math.floor((i - t) / n) * n, r, i) : a;
}
function bn(e, t, n) {
	return n <= t ? 0 : (e - t) / (n - t) * 100;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/RangeSlider.tsx
function xn({ min: e, max: t, value: n, onChange: r, step: a = 1, ariaLabelMin: o = "Minimum value", ariaLabelMax: s = "Maximum value", editable: c = !1, className: l, trackClassName: u, rangeClassName: d, thumbClassName: f }) {
	let [p, m] = Y(n, e, t);
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i(c && "space-y-2", l),
		children: [/* @__PURE__ */ (0, W.jsxs)("div", {
			className: "relative h-6",
			children: [
				/* @__PURE__ */ (0, W.jsx)("div", { className: i("absolute left-0 right-0 top-1/2 h-1 -translate-y-1/2 rounded-full bg-border", u) }),
				/* @__PURE__ */ (0, W.jsx)("div", {
					className: i("absolute top-1/2 h-1 -translate-y-1/2 rounded-full bg-primary/70", d),
					style: {
						left: `${bn(p, e, t)}%`,
						right: `${100 - bn(m, e, t)}%`
					}
				}),
				/* @__PURE__ */ (0, W.jsx)("input", {
					type: "range",
					"aria-label": o,
					min: e,
					max: t,
					step: a,
					value: p,
					className: i(hn, gn, f),
					style: { zIndex: p >= t ? 30 : 20 },
					onChange: (n) => {
						let i = vn(Number(n.target.value), e, t);
						r([Math.min(i, m), m]);
					}
				}),
				/* @__PURE__ */ (0, W.jsx)("input", {
					type: "range",
					"aria-label": s,
					min: e,
					max: t,
					step: a,
					value: m,
					className: i(hn, gn, f),
					onChange: (n) => {
						let i = vn(Number(n.target.value), e, t);
						r([p, Math.max(i, p)]);
					}
				})
			]
		}), c ? /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "grid grid-cols-2 gap-2",
			children: [/* @__PURE__ */ (0, W.jsx)(Sn, {
				label: "Min",
				ariaLabel: `Edit ${Cn(o)}`,
				min: e,
				max: m,
				step: a,
				stepBase: e,
				value: p,
				onChange: (e) => r([e, m])
			}), /* @__PURE__ */ (0, W.jsx)(Sn, {
				label: "Max",
				ariaLabel: `Edit ${Cn(s)}`,
				min: p,
				max: t,
				step: a,
				stepBase: e,
				value: m,
				onChange: (e) => r([p, e])
			})]
		}) : null]
	});
}
function Sn({ label: e, ariaLabel: t, min: n, max: r, step: i, stepBase: a, value: o, onChange: s }) {
	return /* @__PURE__ */ (0, W.jsxs)("label", {
		className: "grid grid-cols-[auto_minmax(0,1fr)] items-center gap-1.5 text-[11px] text-muted-foreground",
		children: [/* @__PURE__ */ (0, W.jsx)("span", { children: e }), /* @__PURE__ */ (0, W.jsx)("input", {
			"aria-label": t,
			className: _n,
			max: r,
			min: n,
			step: i,
			type: "number",
			value: o,
			onChange: (e) => {
				if (e.currentTarget.value === "") return;
				let o = Number(e.currentTarget.value);
				if (!Number.isFinite(o)) throw Error(`${t} must be a finite number`);
				s(yn(o, {
					base: a,
					step: i,
					min: n,
					max: r
				}));
			}
		})]
	});
}
function Y(e, t, n) {
	return [vn(Math.min(e[0], e[1]), t, n), vn(Math.max(e[0], e[1]), t, n)];
}
function Cn(e) {
	return e.charAt(0).toLowerCase() + e.slice(1);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/TimeRange.tsx
var wn = [
	{
		id: "minutes",
		label: "Min",
		chips: [
			{
				label: "5m",
				token: "now-5m"
			},
			{
				label: "15m",
				token: "now-15m"
			},
			{
				label: "30m",
				token: "now-30m"
			}
		]
	},
	{
		id: "hours",
		label: "Hr",
		chips: [
			{
				label: "1h",
				token: "now-1h"
			},
			{
				label: "2h",
				token: "now-2h"
			},
			{
				label: "4h",
				token: "now-4h"
			},
			{
				label: "8h",
				token: "now-8h"
			},
			{
				label: "12h",
				token: "now-12h"
			}
		]
	},
	{
		id: "days",
		label: "Day",
		chips: [
			{
				label: "1d",
				token: "now-1d"
			},
			{
				label: "2d",
				token: "now-2d"
			},
			{
				label: "3d",
				token: "now-3d"
			},
			{
				label: "1w",
				token: "now-1w"
			}
		]
	},
	{
		id: "weeks",
		label: "Wk+",
		chips: [
			{
				label: "2w",
				token: "now-2w"
			},
			{
				label: "1M",
				token: "now-1M"
			},
			{
				label: "3M",
				token: "now-3M"
			},
			{
				label: "6M",
				token: "now-6M"
			},
			{
				label: "1y",
				token: "now-1y"
			}
		]
	}
], Tn = {
	min: wn[0],
	hr: wn[1],
	day: wn[2],
	"wk+": wn[3]
}, En = [
	{
		id: "days",
		label: "Day",
		chips: [
			{
				label: "1d",
				token: "now-1d"
			},
			{
				label: "2d",
				token: "now-2d"
			},
			{
				label: "3d",
				token: "now-3d"
			},
			{
				label: "7d",
				token: "now-7d"
			}
		]
	},
	{
		id: "weeks",
		label: "Wk",
		chips: [
			{
				label: "1w",
				token: "now-1w"
			},
			{
				label: "2w",
				token: "now-2w"
			},
			{
				label: "4w",
				token: "now-4w"
			}
		]
	},
	{
		id: "months",
		label: "Mo",
		chips: [
			{
				label: "1M",
				token: "now-1M"
			},
			{
				label: "3M",
				token: "now-3M"
			},
			{
				label: "6M",
				token: "now-6M"
			},
			{
				label: "1y",
				token: "now-1y"
			}
		]
	}
], X = {
	this: {
		id: "this",
		label: "This",
		chips: [
			{
				label: "Week",
				token: "now/w",
				to: "now/w+1w",
				ariaLabel: "this week"
			},
			{
				label: "Month",
				token: "now/M",
				to: "now/M+1M",
				ariaLabel: "this month"
			},
			{
				label: "Quarter",
				token: "now/q",
				to: "now/q+1q",
				ariaLabel: "this quarter"
			},
			{
				label: "Year",
				token: "now/y",
				to: "now/y+1y",
				ariaLabel: "this year"
			}
		]
	},
	last: {
		id: "last",
		label: "Last",
		chips: [
			{
				label: "Week",
				token: "now/w-1w",
				to: "now/w",
				ariaLabel: "last week"
			},
			{
				label: "Month",
				token: "now/M-1M",
				to: "now/M",
				ariaLabel: "last month"
			},
			{
				label: "Quarter",
				token: "now/q-1q",
				to: "now/q",
				ariaLabel: "last quarter"
			},
			{
				label: "Year",
				token: "now/y-1y",
				to: "now/y",
				ariaLabel: "last year"
			}
		]
	}
};
function Dn({ kind: e = "time", label: t = "Time range", from: n = "", to: r = "", onApply: o, presets: s, chipRows: c, timeEnabled: l = !1, timeZone: d, timeZones: m, fromPlaceholder: g, toPlaceholder: v, emptyLabel: b, align: S = "right", disabled: C = !1, className: w, triggerClassName: T, panelClassName: E, compact: D = !1 }) {
	let [O, k] = (0, U.useState)(!1), [A, j] = (0, U.useState)(n), [M, N] = (0, U.useState)(r), ee = d ?? zn(), [P, te] = (0, U.useState)(ee), F = Bn(m, P), { rows: I, presets: L } = kn(e, s, c), re = _(), { refs: R, floatingStyles: ie, context: ae } = y({
		open: O,
		onOpenChange: k,
		placement: S === "left" ? "bottom-start" : "bottom-end",
		whileElementsMounted: tt,
		middleware: [
			et(6),
			nt({ padding: 8 }),
			rt({ padding: 8 })
		]
	}), { getReferenceProps: oe, getFloatingProps: se } = ne([
		f(ae),
		u(ae, { escapeKey: !1 }),
		x(ae, { role: "dialog" })
	]);
	h(O, () => {
		k(!1), R.domReference.current instanceof HTMLElement && R.domReference.current.focus();
	}), (0, U.useEffect)(() => {
		O && (j(n), N(r), te(ee));
	}, [
		n,
		O,
		ee,
		r
	]);
	let ce = Fn(n, r, I, L) ?? Rn(e, n, r, b), le = Ln(e, n, {
		timeEnabled: l,
		timeZone: P
	});
	function ue(t, n) {
		o(Ln(e, t, {
			timeEnabled: l,
			timeZone: P
		}), Ln(e, n, {
			timeEnabled: l,
			timeZone: P
		})), k(!1);
	}
	function de() {
		j(""), N(""), o("", ""), k(!1);
	}
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i("inline-flex", w),
		children: [/* @__PURE__ */ (0, W.jsxs)(a, {
			ref: R.setReference,
			type: "button",
			variant: "outline",
			size: "sm",
			"aria-label": D ? `${t} filter: ${ce}` : `${t} filter`,
			title: D ? ce : void 0,
			disabled: C,
			className: i("h-7 min-w-0 text-xs font-normal", D ? "w-8 justify-center px-0 max-md:h-8" : "w-fit max-w-[12rem] gap-2 px-2", T),
			...oe(),
			children: [/* @__PURE__ */ (0, W.jsx)(p, {
				icon: e === "time" ? Ft : G,
				className: "text-muted-foreground text-[14px]"
			}), !D && /* @__PURE__ */ (0, W.jsx)("span", {
				className: "truncate font-normal tabular-nums",
				children: ce
			})]
		}), O && /* @__PURE__ */ (0, W.jsx)(pe, { children: /* @__PURE__ */ (0, W.jsx)(he, {
			context: ae,
			modal: !1,
			children: /* @__PURE__ */ (0, W.jsxs)("div", {
				ref: R.setFloating,
				role: "dialog",
				"aria-label": t,
				style: {
					...ie,
					zIndex: re
				},
				className: i("flex max-h-[min(32rem,calc(100vh-2rem))] w-[22.5rem] max-w-[calc(100vw-1rem)] flex-col overflow-hidden rounded-lg border border-border bg-popover text-popover-foreground shadow-lg shadow-black/10 outline-none", E),
				...se(),
				children: [
					/* @__PURE__ */ (0, W.jsxs)("div", {
						className: "flex items-center justify-between border-b border-border/60 px-3.5 py-2.5",
						children: [/* @__PURE__ */ (0, W.jsx)("span", {
							className: "text-[10px] font-semibold uppercase tracking-wider text-muted-foreground",
							children: t
						}), /* @__PURE__ */ (0, W.jsx)("span", {
							className: "font-mono text-[10px] text-muted-foreground",
							children: l ? P : "Relative"
						})]
					}),
					/* @__PURE__ */ (0, W.jsxs)("div", {
						className: "min-h-0 flex-1 overflow-y-auto",
						children: [
							I.length > 0 && /* @__PURE__ */ (0, W.jsx)(Nn, {
								rows: I,
								activeToken: le,
								onApply: ue
							}),
							L.length > 0 && /* @__PURE__ */ (0, W.jsx)("div", {
								className: I.length > 0 ? "px-3.5 pb-3" : "px-3.5 pt-3",
								children: /* @__PURE__ */ (0, W.jsx)(Pn, {
									presets: L,
									onApply: ue
								})
							}),
							/* @__PURE__ */ (0, W.jsxs)("div", {
								className: "border-t border-border/60 bg-muted/30 px-3.5 py-2.5",
								children: [/* @__PURE__ */ (0, W.jsxs)("div", {
									className: "mb-1.5 flex flex-wrap items-center justify-between gap-2",
									children: [/* @__PURE__ */ (0, W.jsx)("div", {
										className: "text-[10px] font-semibold uppercase tracking-wider text-muted-foreground",
										children: "Custom"
									}), l && /* @__PURE__ */ (0, W.jsxs)("div", {
										className: "flex min-w-0 items-center gap-1.5",
										children: [/* @__PURE__ */ (0, W.jsx)("span", {
											className: "text-[10px] font-semibold uppercase tracking-wider text-muted-foreground",
											children: "TZ"
										}), /* @__PURE__ */ (0, W.jsx)(pn, {
											"aria-label": `${t} timezone`,
											value: P,
											options: F.map((e) => ({
												value: e,
												label: e
											})),
											onChange: (e) => te(e.target.value),
											className: "h-7 w-40 truncate px-2 pr-7 font-mono text-xs"
										})]
									})]
								}), /* @__PURE__ */ (0, W.jsxs)("div", {
									className: "grid gap-1.5 md:grid-cols-[1fr_auto_1fr] md:items-center",
									children: [
										/* @__PURE__ */ (0, W.jsx)(On, {
											"aria-label": `${t} from`,
											placeholder: g ?? (e === "date" ? "YYYY-MM-DD or now-1d" : "now-24h"),
											value: A,
											onChange: j,
											onClear: () => j(""),
											pickerKind: e,
											timeEnabled: l
										}),
										/* @__PURE__ */ (0, W.jsx)(p, {
											icon: Ct,
											className: "hidden text-muted-foreground text-[12px] md:block"
										}),
										/* @__PURE__ */ (0, W.jsx)(On, {
											"aria-label": `${t} to`,
											placeholder: v ?? (e === "date" ? "YYYY-MM-DD or now" : "now"),
											value: M,
											onChange: N,
											onClear: () => N(""),
											pickerKind: e,
											timeEnabled: l
										})
									]
								})]
							})
						]
					}),
					/* @__PURE__ */ (0, W.jsxs)("div", {
						className: "flex items-center justify-between border-t border-border/60 px-3 py-2",
						children: [/* @__PURE__ */ (0, W.jsx)(a, {
							type: "button",
							variant: "ghost",
							size: "sm",
							className: "h-7 px-2.5 text-xs",
							disabled: !n && !r,
							onClick: de,
							children: "Clear"
						}), /* @__PURE__ */ (0, W.jsxs)("div", {
							className: "flex gap-1.5",
							children: [/* @__PURE__ */ (0, W.jsx)(a, {
								type: "button",
								variant: "ghost",
								size: "sm",
								className: "h-7 px-2.5 text-xs",
								onClick: () => k(!1),
								children: "Cancel"
							}), /* @__PURE__ */ (0, W.jsx)(a, {
								type: "button",
								variant: "default",
								size: "sm",
								className: "h-7 px-3.5 text-xs",
								onClick: () => ue(A, M),
								children: "Apply"
							})]
						})]
					})
				]
			})
		}) })]
	});
}
function On({ value: e, onChange: t, onClear: n, placeholder: r, pickerKind: i, timeEnabled: a, ...o }) {
	let s = o["aria-label"], c = i !== void 0, l = i === "time" && a, { date: u, time: d } = Mn(e);
	function f(e) {
		t(l ? e ? d ? `${e}T${d}` : e : d ? `T${d}` : "" : e);
	}
	function p(e) {
		t(e ? u ? `${u}T${e}` : `T${e}` : u);
	}
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "space-y-1",
		children: [/* @__PURE__ */ (0, W.jsx)(An, {
			ariaLabel: s,
			placeholder: r,
			value: l ? u : e,
			onChange: l ? f : t,
			onClear: n,
			showCalendar: c
		}), l && /* @__PURE__ */ (0, W.jsx)(jn, {
			ariaLabel: `${s} time`,
			value: d,
			onChange: p
		})]
	});
}
function kn(e, t, n) {
	if (t === void 0) return {
		rows: n ?? (e === "date" ? En : wn),
		presets: []
	};
	let r = n ? [...n] : [], i = [];
	for (let e of t) typeof e == "string" ? e === "this" || e === "last" ? r.push(X[e]) : r.push(Tn[e]) : i.push(e);
	return {
		rows: r,
		presets: i
	};
}
function An({ ariaLabel: e, value: t, onChange: n, onClear: r, placeholder: a, showCalendar: o }) {
	let s = (0, U.useRef)(null), c = t.length > 0;
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "relative",
		children: [
			/* @__PURE__ */ (0, W.jsx)("input", {
				type: "text",
				"aria-label": e,
				placeholder: a,
				value: t,
				onChange: (e) => n(e.target.value),
				className: i("h-8 w-full rounded-md border border-input bg-background px-2 font-mono text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring", o && c ? "pr-12" : "pr-7")
			}),
			c && /* @__PURE__ */ (0, W.jsx)("button", {
				type: "button",
				"aria-label": `Clear ${e}`,
				onClick: r,
				className: i("absolute inset-y-0 inline-flex w-5 items-center justify-center text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", o ? "right-6" : "right-1"),
				children: /* @__PURE__ */ (0, W.jsx)(p, {
					icon: fe,
					className: "text-[12px]"
				})
			}),
			o && /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [/* @__PURE__ */ (0, W.jsx)("input", {
				ref: s,
				type: "date",
				tabIndex: -1,
				"aria-hidden": "true",
				value: /^\d{4}-\d{2}-\d{2}$/.test(t.trim()) ? t.trim() : "",
				onChange: (e) => n(e.target.value),
				className: "pointer-events-none absolute left-0 top-0 h-0 w-0 opacity-0"
			}), /* @__PURE__ */ (0, W.jsx)("button", {
				type: "button",
				"aria-label": `Pick ${e}`,
				onClick: () => s.current?.showPicker?.(),
				className: "absolute inset-y-0 right-1 inline-flex w-5 items-center justify-center text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
				children: /* @__PURE__ */ (0, W.jsx)(p, {
					icon: G,
					className: "text-[12px]"
				})
			})] })
		]
	});
}
function jn({ ariaLabel: e, value: t, onChange: n }) {
	return /* @__PURE__ */ (0, W.jsx)("input", {
		type: "time",
		"aria-label": e,
		value: t,
		onChange: (e) => n(e.target.value),
		className: "h-8 w-full rounded-md border border-input bg-background px-2 font-mono text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring"
	});
}
function Mn(e) {
	let t = e.trim(), n = /^(\d{4}-\d{2}-\d{2})?T?(\d{2}:\d{2})?/.exec(t);
	return !n || t.startsWith("now") || /^[+-]/.test(t) ? {
		date: t,
		time: ""
	} : {
		date: n[1] ?? "",
		time: n[2] ?? ""
	};
}
function Nn({ rows: e, activeToken: t, onApply: n }) {
	return e.length === 0 ? null : /* @__PURE__ */ (0, W.jsx)("div", {
		className: "px-3.5 py-3",
		role: "list",
		"aria-label": "Time filters",
		children: e.map((r, a) => /* @__PURE__ */ (0, W.jsxs)("div", {
			className: i("grid grid-cols-[2rem_1fr] items-center gap-2", a === e.length - 1 ? "" : "mb-1.5"),
			role: "listitem",
			children: [/* @__PURE__ */ (0, W.jsx)("span", {
				className: "text-[9.5px] font-bold uppercase tracking-wider text-muted-foreground",
				children: r.label
			}), /* @__PURE__ */ (0, W.jsx)("div", {
				className: "flex divide-x divide-border overflow-hidden rounded-[5px] border border-border",
				children: r.chips.map((e) => {
					let r = e.token === t;
					return /* @__PURE__ */ (0, W.jsx)("button", {
						type: "button",
						title: e.title ?? e.ariaLabel ?? e.token,
						"aria-label": e.ariaLabel ?? e.token,
						"aria-pressed": r,
						onClick: () => n(e.token, e.to ?? "now"),
						className: i("flex-1 min-w-0 px-0 py-1.5 text-center font-mono text-[11.5px] leading-none transition-colors focus-visible:relative focus-visible:z-10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", r ? "bg-foreground font-semibold text-background" : "bg-background font-medium text-muted-foreground hover:bg-muted"),
						children: e.label
					}, e.token);
				})
			})]
		}, r.id))
	});
}
function Pn({ presets: e, onApply: t }) {
	return e.length === 0 ? null : /* @__PURE__ */ (0, W.jsx)("div", {
		className: "mb-3 space-y-1",
		role: "list",
		"aria-label": "Time filters",
		children: e.map((e) => /* @__PURE__ */ (0, W.jsxs)("button", {
			type: "button",
			className: "flex h-8 w-full items-center justify-between gap-3 rounded-md border border-border bg-background px-2 text-left text-xs hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
			onClick: () => t(e.from, e.to),
			children: [/* @__PURE__ */ (0, W.jsx)("span", {
				className: "min-w-0 truncate font-medium",
				children: e.label
			}), /* @__PURE__ */ (0, W.jsx)("span", {
				className: "shrink-0 font-mono text-[11px] text-muted-foreground",
				children: In(e)
			})]
		}, `${e.label}-${e.from}-${e.to}`))
	});
}
function Fn(e, t, n, r) {
	let i = e.trim(), a = t.trim();
	if (i || a) {
		for (let e of n) for (let t of e.chips) if (t.token === i && (t.to ?? "now") === a) return t.ariaLabel ?? t.title ?? t.label;
		return r.find((e) => e.from === i && e.to === a)?.label;
	}
}
function In({ from: e, to: t }) {
	return e ? !t || t === "now" ? e : `${e} -> ${t}` : t;
}
function Ln(e, t, n = {}) {
	let r = t.trim();
	return r ? r === "now" || r.startsWith("now") ? r : e === "time" && /^[+-]\d/.test(r) ? `now${r}` : e === "time" && n.timeEnabled ? Vn(r, n.timeZone ?? zn()) : r : "";
}
function Rn(e, t, n, r) {
	let i = t.trim(), a = n.trim();
	if (!i && !a) return r ?? (e === "date" ? "Any date" : "now-24h");
	if (e === "time") {
		let t = Ln(e, i), n = Ln(e, a || "now");
		return t ? n === "now" ? t : `${t} -> ${n}` : n;
	}
	return i && a && i === a ? i : i ? a ? `${i} -> ${a}` : i : a;
}
function zn() {
	try {
		return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
	} catch {
		return "UTC";
	}
}
function Bn(e, t) {
	let n = e && e.length > 0 ? e : [t, "UTC"];
	return Array.from(new Set([
		t,
		...n,
		"UTC"
	].filter(Boolean)));
}
function Vn(e, t) {
	if (!Hn(e) || Un(e)) return e;
	let n = e.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?(?:\.\d+)?$/);
	if (!n) return e;
	let [, r, i, a, o, s, c = "00"] = n;
	return `${r}-${i}-${a}T${o}:${s}:${c}${Gn(Wn(t, Date.UTC(Number(r), Number(i) - 1, Number(a), Number(o), Number(s), Number(c))))}`;
}
function Hn(e) {
	return /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(?::\d{2})?(?:\.\d+)?$/.test(e.trim());
}
function Un(e) {
	return /(?:Z|[+-]\d{2}:?\d{2})$/i.test(e.trim());
}
function Wn(e, t) {
	try {
		let n = new Intl.DateTimeFormat("en-US", {
			timeZone: e,
			year: "numeric",
			month: "2-digit",
			day: "2-digit",
			hour: "2-digit",
			minute: "2-digit",
			second: "2-digit",
			hour12: !1
		}), r = Object.fromEntries(n.formatToParts(new Date(t)).filter((e) => e.type !== "literal").map((e) => [e.type, e.value]));
		return (Date.UTC(Number(r.year), Number(r.month) - 1, Number(r.day), Number(r.hour), Number(r.minute), Number(r.second)) - t) / 6e4;
	} catch {
		return 0;
	}
}
function Gn(e) {
	let t = e >= 0 ? "+" : "-", n = Math.abs(e);
	return `${t}${Math.floor(n / 60).toString().padStart(2, "0")}:${Math.floor(n % 60).toString().padStart(2, "0")}`;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/tri-state.ts
function Kn(e) {
	return e === void 0 || !e && void 0;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/TriStateToggle.tsx
var qn = {
	unset: "Any",
	on: "Yes",
	off: "No"
};
function Jn({ value: e, onChange: t, label: n, labels: r, id: a, size: o = "sm", disabled: s, className: c }) {
	let l = {
		...qn,
		...r
	}, u = e === void 0 ? "unset" : e ? "on" : "off", d = u === "on" ? L : u === "off" ? fe : Le, f = u === "on" ? "border-green-500/50 bg-green-500/10 text-green-700 dark:text-green-400" : u === "off" ? "border-red-500/50 bg-red-500/10 text-red-700 dark:text-red-400" : "border-input bg-muted/30 text-muted-foreground";
	return /* @__PURE__ */ (0, W.jsx)("button", {
		type: "button",
		...a === void 0 ? {} : { id: a },
		role: "checkbox",
		"aria-checked": e === void 0 ? "mixed" : e,
		"aria-label": n,
		title: `${n}: ${l[u]}. Click to cycle.`,
		disabled: s,
		onClick: () => t(Kn(e)),
		className: i("inline-flex shrink-0 items-center justify-center rounded-md border transition-colors", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", o === "sm" ? "h-6 w-6" : "h-8 w-8", f, s ? "opacity-60" : "hover:brightness-95", c),
		children: /* @__PURE__ */ (0, W.jsx)(p, {
			icon: d,
			className: o === "sm" ? "h-3.5 w-3.5" : "h-4 w-4"
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/NamespacePicker.tsx
function Yn({ value: e, onChange: t, loadNamespaces: n, strict: r = !1, placeholder: i = "Select namespace…", className: a }) {
	let [o, s] = (0, U.useState)([]), [c, l] = (0, U.useState)(!1);
	(0, U.useEffect)(() => {
		let e = !1;
		return l(!0), n().then((t) => {
			e || s(t);
		}).finally(() => {
			e || l(!1);
		}), () => {
			e = !0;
		};
	}, [n]);
	let u = (0, U.useMemo)(() => {
		let t = o.map((e) => ({
			value: e,
			label: e
		}));
		return e && !o.includes(e) ? [{
			value: e,
			label: e
		}, ...t] : t;
	}, [o, e]), d = (0, U.useMemo)(() => r && !!e && !c && !o.includes(e), [
		r,
		e,
		c,
		o
	]);
	return /* @__PURE__ */ (0, W.jsx)(He, {
		options: u,
		value: e,
		onChange: t,
		ariaLabel: "Namespace",
		allowCustomValue: !0,
		loading: c,
		invalid: d,
		placeholder: i,
		...a ? { className: a } : {}
	});
}
// ../../../../../../pnpm/store/v11/links/@flanksource/icons/1.0.63/b00fc4bb064c1d1ae674af83cb73c1f20c4dd4dfd0aa56010e6ee736bf3e3c7d/node_modules/@flanksource/icons/lib/iconContext.mjs
var Xn = {
	color: void 0,
	size: void 0,
	className: void 0,
	style: void 0,
	attr: void 0
}, Zn = U.createContext && /*#__PURE__*/ U.createContext(Xn), Qn = [
	"attr",
	"size",
	"title",
	"square"
];
function $n(e, t) {
	if (e == null) return {};
	var n = er(e, t), r, i;
	if (Object.getOwnPropertySymbols) {
		var a = Object.getOwnPropertySymbols(e);
		for (i = 0; i < a.length; i++) r = a[i], !(t.indexOf(r) >= 0) && Object.prototype.propertyIsEnumerable.call(e, r) && (n[r] = e[r]);
	}
	return n;
}
function er(e, t) {
	if (e == null) return {};
	for (var n = {}, r = Object.keys(e), i, a = 0; a < r.length; a++) i = r[a], !(t.indexOf(i) >= 0) && (n[i] = e[i]);
	return n;
}
function tr() {
	return tr = Object.assign ? Object.assign.bind() : function(e) {
		for (var t = 1; t < arguments.length; t++) {
			var n = arguments[t];
			for (var r in n) Object.prototype.hasOwnProperty.call(n, r) && (e[r] = n[r]);
		}
		return e;
	}, tr.apply(this, arguments);
}
function nr(e, t) {
	var n = Object.keys(e);
	if (Object.getOwnPropertySymbols) {
		var r = Object.getOwnPropertySymbols(e);
		t && (r = r.filter(function(t) {
			return Object.getOwnPropertyDescriptor(e, t).enumerable;
		})), n.push.apply(n, r);
	}
	return n;
}
function rr(e) {
	for (var t = 1; t < arguments.length; t++) {
		var n = arguments[t] == null ? {} : arguments[t];
		t % 2 ? nr(Object(n), !0).forEach(function(t) {
			ir(e, t, n[t]);
		}) : Object.getOwnPropertyDescriptors ? Object.defineProperties(e, Object.getOwnPropertyDescriptors(n)) : nr(Object(n)).forEach(function(t) {
			Object.defineProperty(e, t, Object.getOwnPropertyDescriptor(n, t));
		});
	}
	return e;
}
function ir(e, t, n) {
	return t = Z(t), t in e ? Object.defineProperty(e, t, {
		value: n,
		enumerable: !0,
		configurable: !0,
		writable: !0
	}) : e[t] = n, e;
}
function Z(e) {
	var t = ar(e, "string");
	return typeof t == "symbol" ? t : String(t);
}
function ar(e, t) {
	if (typeof e != "object" || !e) return e;
	var n = e[Symbol.toPrimitive];
	if (n !== void 0) {
		var r = n.call(e, t || "default");
		if (typeof r != "object") return r;
		throw TypeError("@@toPrimitive must return a primitive value.");
	}
	return (t === "string" ? String : Number)(e);
}
function or(e) {
	return e && e.map((e, t) => /*#__PURE__*/ U.createElement(e.tag, rr({ key: t }, e.attr), or(e.child)));
}
function sr(e) {
	var t = (t) => /*#__PURE__*/ U.createElement(lr, tr({ attr: rr({}, e.attr) }, t), or(e.child));
	return t.viewBox = e.attr?.viewBox, t;
}
function cr(e) {
	if (!e) return !1;
	var t = e.trim().split(/[\s,]+/);
	if (t.length !== 4) return !1;
	var n = parseFloat(t[2]), r = parseFloat(t[3]);
	return !isFinite(n) || !isFinite(r) || r <= 0 ? !1 : n / r >= 2;
}
function lr(e) {
	var t = (t) => {
		var { attr: n, size: r, title: i, square: a } = e, o = $n(e, Qn), s = r || t.size || "1em", c = a ?? !cr(n?.viewBox), l;
		return t.className && (l = t.className), e.className && (l = (l ? l + " " : "") + e.className), /*#__PURE__*/ U.createElement("svg", tr({
			stroke: "currentColor",
			fill: "currentColor",
			strokeWidth: "0"
		}, t.attr, n, o, {
			className: l,
			style: rr(rr({ color: e.color || t.color }, t.style), e.style),
			height: s,
			width: c ? s : void 0,
			xmlns: "http://www.w3.org/2000/svg"
		}), i && /*#__PURE__*/ U.createElement("title", null, i), e.children);
	};
	return Zn === void 0 ? t(Xn) : /*#__PURE__*/ U.createElement(Zn.Consumer, null, (e) => t(e));
}
// ../../../../../../pnpm/store/v11/links/@flanksource/icons/1.0.63/b00fc4bb064c1d1ae674af83cb73c1f20c4dd4dfd0aa56010e6ee736bf3e3c7d/node_modules/@flanksource/icons/mi/index.mjs
function ur(e) {
	return sr({
		tag: "svg",
		attr: {
			role: "img",
			viewBox: "0 0 32.978 32"
		},
		child: [
			{
				tag: "path",
				attr: {
					d: "M16.379.848a2.238 2.238 0 0 0-.793.199L4.723 6.238a2.06 2.06 0 0 0-1.121 1.399L.922 19.293c-.14.61.004 1.254.398 1.742l7.52 9.348c.394.488.992.773 1.62.773l12.056-.004c.632 0 1.23-.28 1.625-.773l7.515-9.348a2.042 2.042 0 0 0 .403-1.742L29.375 7.633a2.055 2.055 0 0 0-1.125-1.399L17.387 1.047a2.082 2.082 0 0 0-1.008-.2m0 0",
					fill: "rgb(19.607843%, 42.352941%, 89.803922%)"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M16.375.004a2.22 2.22 0 0 0-.84.21L4.07 5.696a2.171 2.171 0 0 0-1.187 1.473L.055 19.473a2.164 2.164 0 0 0 .297 1.668c.039.062.082.117.125.172l7.937 9.87a2.21 2.21 0 0 0 1.715.817l12.723-.004c.668 0 1.3-.3 1.714-.816l7.934-9.867c.414-.516.57-1.196.422-1.84L30.09 7.164a2.187 2.187 0 0 0-1.184-1.473L17.438.215a2.216 2.216 0 0 0-1.063-.211m.004.844c.348-.02.695.05 1.008.199L28.25 6.234c.566.274.98.786 1.125 1.399l2.684 11.656a2.042 2.042 0 0 1-.403 1.742l-7.515 9.348a2.065 2.065 0 0 1-1.625.773l-12.055.004c-.633 0-1.23-.285-1.621-.773l-7.52-9.348a2.043 2.043 0 0 1-.398-1.742l2.68-11.656a2.06 2.06 0 0 1 1.12-1.399l10.864-5.191c.246-.117.52-.188.793-.2m0 0",
					fill: "rgb(100%, 100%, 100%)"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M15.688 24.121c-4.465.207-8.266-3.226-8.504-7.691-.192-4.434 3.214-8.203 7.644-8.465a8.135 8.135 0 0 1 8.586 7.512l-3.21.207a4.91 4.91 0 0 0-5.188-4.586 4.907 4.907 0 0 0-4.618 5.16 4.91 4.91 0 0 0 5.137 4.648Zm1.664-8.637 2.554 6.254 5.899-6.254Zm0 0",
					fill: "rgb(100%, 100%, 100%)",
					fillRule: "evenodd"
				},
				child: []
			}
		]
	})(e);
}
function dr(e) {
	return sr({
		tag: "svg",
		attr: {
			role: "img",
			viewBox: "0 0 32.978 32"
		},
		child: [{
			tag: "path",
			attr: {
				d: "M16.379.848a2.238 2.238 0 0 0-.793.199L4.723 6.238a2.06 2.06 0 0 0-1.121 1.399L.922 19.293c-.14.61.004 1.254.398 1.742l7.52 9.348c.394.488.992.773 1.62.773l12.056-.004c.632 0 1.23-.28 1.625-.773l7.515-9.348a2.042 2.042 0 0 0 .403-1.742L29.375 7.633a2.055 2.055 0 0 0-1.125-1.399L17.387 1.047a2.082 2.082 0 0 0-1.008-.2m0 0",
				fill: "rgb(19.607843%, 42.352941%, 89.803922%)"
			},
			child: []
		}, {
			tag: "path",
			attr: {
				d: "M16.375.004a2.22 2.22 0 0 0-.84.21L4.07 5.696a2.171 2.171 0 0 0-1.187 1.473L.055 19.473a2.164 2.164 0 0 0 .297 1.668c.039.062.082.117.125.172l7.937 9.87a2.21 2.21 0 0 0 1.715.817l12.723-.004c.668 0 1.3-.3 1.714-.816l7.934-9.867c.414-.516.57-1.196.422-1.84L30.09 7.164a2.187 2.187 0 0 0-1.184-1.473L17.438.215a2.216 2.216 0 0 0-1.063-.211m.004.844c.348-.02.695.05 1.008.199L28.25 6.234c.566.274.98.786 1.125 1.399l2.684 11.656a2.042 2.042 0 0 1-.403 1.742l-7.515 9.348a2.065 2.065 0 0 1-1.625.773l-12.055.004c-.633 0-1.23-.285-1.621-.773l-7.52-9.348a2.043 2.043 0 0 1-.398-1.742l2.68-11.656a2.06 2.06 0 0 1 1.12-1.399l10.864-5.191c.246-.117.52-.188.793-.2m5.133 22.598h-4.153L8.363 12.191H5.762V8.508h4.23l9.004 11.254h2.516v-2.914l5.707 4.754-5.707 4.757Zm-4.528-8.691 2.012-2.516h2.516v2.914l5.707-4.754-5.707-4.757v2.914h-4.153l-2.664 3.332Zm-6.578 2.492L8.363 19.81H5.762v3.683h4.23l2.703-3.379Zm0 0",
				fill: "rgb(100%, 100%, 100%)"
			},
			child: []
		}]
	})(e);
}
function fr(e) {
	return sr({
		tag: "svg",
		attr: {
			role: "img",
			viewBox: "0 0 32.978 32"
		},
		child: [
			{
				tag: "path",
				attr: {
					d: "M16.379.848a2.238 2.238 0 0 0-.793.199L4.723 6.238a2.06 2.06 0 0 0-1.121 1.399L.922 19.293c-.14.61.004 1.254.398 1.742l7.52 9.348c.394.488.992.773 1.62.773l12.056-.004c.632 0 1.23-.28 1.625-.773l7.515-9.348a2.042 2.042 0 0 0 .403-1.742L29.375 7.633a2.055 2.055 0 0 0-1.125-1.399L17.387 1.047a2.082 2.082 0 0 0-1.008-.2m0 0",
					fill: "rgb(19.607843%, 42.352941%, 89.803922%)"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M16.375.004a2.22 2.22 0 0 0-.84.21L4.07 5.696a2.171 2.171 0 0 0-1.187 1.473L.055 19.473a2.164 2.164 0 0 0 .297 1.668c.039.062.082.117.125.172l7.937 9.87a2.21 2.21 0 0 0 1.715.817l12.723-.004c.668 0 1.3-.3 1.714-.816l7.934-9.867c.414-.516.57-1.196.422-1.84L30.09 7.164a2.187 2.187 0 0 0-1.184-1.473L17.438.215a2.216 2.216 0 0 0-1.063-.211m.004.844c.348-.02.695.05 1.008.199L28.25 6.234c.566.274.98.786 1.125 1.399l2.684 11.656a2.042 2.042 0 0 1-.403 1.742l-7.515 9.348a2.065 2.065 0 0 1-1.625.773l-12.055.004c-.633 0-1.23-.285-1.621-.773l-7.52-9.348a2.043 2.043 0 0 1-.398-1.742l2.68-11.656a2.06 2.06 0 0 1 1.12-1.399l10.864-5.191c.246-.117.52-.188.793-.2m0 0",
					fill: "rgb(100%, 100%, 100%)"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "m9.867 11.363 6.621-1.922 6.621 1.922-6.62 1.918Zm0 .735v7.043l6.168 3.418.031-8.618Zm13.242 0v7.043l-6.168 3.418-.03-8.618Zm0 0",
					fill: "rgb(100%, 100%, 100%)",
					fillRule: "evenodd"
				},
				child: []
			}
		]
	})(e);
}
function pr(e) {
	return sr({
		tag: "svg",
		attr: {
			role: "img",
			viewBox: "0 0 32.978 32"
		},
		child: [
			{
				tag: "path",
				attr: {
					d: "M16.379.848a2.238 2.238 0 0 0-.793.199L4.723 6.238a2.06 2.06 0 0 0-1.121 1.399L.922 19.293c-.14.61.004 1.254.398 1.742l7.52 9.348c.394.488.992.773 1.62.773l12.056-.004c.632 0 1.23-.28 1.625-.773l7.515-9.348a2.042 2.042 0 0 0 .403-1.742L29.375 7.633a2.055 2.055 0 0 0-1.125-1.399L17.387 1.047a2.082 2.082 0 0 0-1.008-.2m0 0",
					fill: "rgb(19.607843%, 42.352941%, 89.803922%)"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M16.375.004a2.22 2.22 0 0 0-.84.21L4.07 5.696a2.171 2.171 0 0 0-1.187 1.473L.055 19.473a2.164 2.164 0 0 0 .297 1.668c.039.062.082.117.125.172l7.937 9.87a2.21 2.21 0 0 0 1.715.817l12.723-.004c.668 0 1.3-.3 1.714-.816l7.934-9.867c.414-.516.57-1.196.422-1.84L30.09 7.164a2.187 2.187 0 0 0-1.184-1.473L17.438.215a2.216 2.216 0 0 0-1.063-.211m.004.844c.348-.02.695.05 1.008.199L28.25 6.234c.566.274.98.786 1.125 1.399l2.684 11.656a2.042 2.042 0 0 1-.403 1.742l-7.515 9.348a2.065 2.065 0 0 1-1.625.773l-12.055.004c-.633 0-1.23-.285-1.621-.773l-7.52-9.348a2.043 2.043 0 0 1-.398-1.742l2.68-11.656a2.06 2.06 0 0 1 1.12-1.399l10.864-5.191c.246-.117.52-.188.793-.2m0 0",
					fill: "rgb(100%, 100%, 100%)"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M6.574 19.668h5.317v3.73H6.574Zm7.258 0h5.316v3.73h-5.316Zm7.254 0h5.32v3.73h-5.32ZM12.277 8.602h8.426v3.73h-8.426Zm0 0",
					fill: "rgb(100%, 100%, 100%)",
					fillRule: "evenodd"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M16.488 12.332V16H9.234v3.664",
					fill: "none",
					stroke: "rgb(100%, 100%, 100%)",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: ".968"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M16.488 12.332V16h7.254v3.664",
					fill: "none",
					stroke: "rgb(100%, 100%, 100%)",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: ".967"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M16.473 12.336v3.668h.03v3.66",
					fill: "none",
					stroke: "rgb(100%, 100%, 100%)",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: ".968"
				},
				child: []
			}
		]
	})(e);
}
function mr(e) {
	return sr({
		tag: "svg",
		attr: {
			role: "img",
			viewBox: "0 0 32.978 32"
		},
		child: [
			{
				tag: "path",
				attr: {
					d: "M16.445.848a2.085 2.085 0 0 0-.797.199L4.79 6.238c-.57.27-.984.785-1.125 1.399L.984 19.293a2.043 2.043 0 0 0 .282 1.578c.039.055.078.11.12.164l7.516 9.348c.395.488.993.773 1.625.773l12.055-.004c.633 0 1.23-.28 1.625-.773l7.512-9.348a2.042 2.042 0 0 0 .402-1.742L29.437 7.633a2.055 2.055 0 0 0-1.125-1.399L17.454 1.047a2.082 2.082 0 0 0-1.008-.2m0 0",
					fill: "rgb(19.607843%, 42.352941%, 89.803922%)"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M16.438.004c-.29.016-.575.086-.836.21L4.132 5.696A2.176 2.176 0 0 0 2.95 7.168L.121 19.473a2.14 2.14 0 0 0 .422 1.84l7.934 9.87A2.206 2.206 0 0 0 10.19 32l12.727-.004c.668 0 1.297-.3 1.715-.816l7.93-9.867a2.168 2.168 0 0 0 .425-1.84L30.156 7.164a2.192 2.192 0 0 0-1.187-1.473L17.504.215c-.332-.156-.7-.23-1.067-.211m.008.844c.348-.02.692.05 1.008.199l10.86 5.187c.57.274.984.79 1.125 1.399l2.683 11.656a2.042 2.042 0 0 1-.402 1.742l-7.512 9.348a2.07 2.07 0 0 1-1.625.773l-12.055.004c-.632 0-1.23-.285-1.625-.773l-7.515-9.348a2.042 2.042 0 0 1-.403-1.742l2.68-11.656A2.055 2.055 0 0 1 4.79 6.238l10.86-5.191c.25-.117.519-.188.796-.2m0 0",
					fill: "rgb(100%, 100%, 100%)"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M13.402 9.23h11.93v8.38h-11.93Zm0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0",
					fill: "none",
					stroke: "rgb(100%, 100%, 100%)",
					strokeDasharray: "2.903 2.903",
					strokeDashoffset: ".9",
					strokeLinecap: "square",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: ".968"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M10.59 11.809h11.93v8.382H10.59Zm0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0",
					fill: "rgb(19.607843%, 42.352941%, 89.803922%)",
					fillRule: "evenodd",
					stroke: "rgb(100%, 100%, 100%)",
					strokeDasharray: "2.903 2.903",
					strokeDashoffset: "1.287",
					strokeLinecap: "square",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: ".968"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M7.773 14.39h11.934v8.38H7.773Zm0 0",
					fill: "rgb(100%, 100%, 100%)",
					fillRule: "evenodd",
					stroke: "rgb(100%, 100%, 100%)",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: ".968"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M8.75 16.719c0-.785 2.234-1.422 4.988-1.422 2.758 0 4.992.637 4.992 1.422 0 .785-2.234 1.422-4.992 1.422-2.754 0-4.988-.637-4.988-1.422m0 0",
					fill: "rgb(100%, 100%, 100%)",
					fillRule: "evenodd"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M18.73 16.719c0 .785-2.234 1.422-4.992 1.422-2.754 0-4.988-.637-4.988-1.422 0-.785 2.234-1.422 4.988-1.422 2.758 0 4.992.637 4.992 1.422v3.847c0 .786-2.234 1.422-4.992 1.422-2.754 0-4.988-.636-4.988-1.422V16.72",
					fill: "none",
					stroke: "rgb(19.607843%, 42.352941%, 89.803922%)",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: ".968"
				},
				child: []
			}
		]
	})(e);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/DaemonSetIcon.tsx
function hr({ className: e, title: t, "aria-label": n }) {
	return /* @__PURE__ */ (0, W.jsx)(St, {
		...e === void 0 ? {} : { className: e },
		...t === void 0 ? {} : { title: t },
		...n === void 0 ? {} : { "aria-label": n }
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/workload-picker-utils.tsx
var gr = {
	service: {
		label: "Service",
		Icon: pr
	},
	ingress: {
		label: "Ingress",
		Icon: dr
	},
	pod: {
		label: "Pod",
		Icon: fr
	},
	deployment: {
		label: "Deployment",
		Icon: ur
	},
	statefulset: {
		label: "StatefulSet",
		Icon: mr
	},
	daemonset: {
		label: "DaemonSet",
		Icon: hr
	}
}, _r = [
	"service",
	"ingress",
	"pod",
	"deployment",
	"statefulset",
	"daemonset"
];
function vr(e, t) {
	return e === "ingress" ? t.hosts?.[0] : void 0;
}
function yr(e, t) {
	return vr(e, t) ?? t.name;
}
var br = new Set(_r);
function xr(e) {
	return e ? `${e}/` : "";
}
function Sr(e, t, n) {
	let r = yr(t, n), i = n.namespace ?? e;
	return i ? `${i}/${t}/${r}` : `${t}/${r}`;
}
function Cr(e) {
	let t = e.split("/"), n = t.at(-1), r = t.at(-2);
	if (n != null && r != null && br.has(r)) {
		let e = t.slice(0, -2).join("/"), i = {
			kind: r,
			name: n
		};
		return e && (i.namespace = e), i;
	}
	return n === "" && t.length === 2 && t[0] ? {
		namespace: t[0],
		name: ""
	} : { name: e };
}
function wr(e, t, n) {
	let r = /* @__PURE__ */ new Set(), i = /* @__PURE__ */ new Set();
	for (let a of n) for (let n of t[a] ?? []) r.add(Sr(e, a, n)), i.add(yr(a, n));
	return {
		keys: r,
		names: i
	};
}
function Tr(e, t) {
	return (t ? Cr(t).kind : void 0) ?? e[0] ?? "service";
}
function Er(e, t, n, r) {
	let i = [];
	for (let r of n) {
		let n = gr[r], a = n.Icon;
		for (let o of t[r] ?? []) {
			let t = vr(r, o);
			i.push({
				value: Sr(e, r, o),
				label: t ? `${t} (${o.name})` : o.name,
				group: n.label,
				icon: /* @__PURE__ */ (0, W.jsx)(a, { className: "h-4 w-4" })
			});
		}
	}
	if (r && !i.some((e) => e.value === r)) {
		let e = Cr(r), t = gr[e.kind ?? n[0] ?? "service"].Icon;
		return [{
			value: r,
			label: e.name,
			icon: /* @__PURE__ */ (0, W.jsx)(t, { className: "h-4 w-4" })
		}, ...i];
	}
	return i;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/WorkloadPicker.tsx
function Dr({ value: e, onChange: t, loadWorkloads: n, namespace: r, allowNamespaceSelection: a = !1, loadNamespaces: o, onNamespaceChange: s, kinds: c = _r, strict: l = !1, allowCustomValue: u = !0, collapseSingleOption: d = !1, disabled: f = !1, placeholder: p = "Select workload / service…", className: m }) {
	let h = e ? Cr(e).namespace : void 0, [g, _] = (0, U.useState)(h ?? r ?? ""), [v, y] = (0, U.useState)({}), [b, x] = (0, U.useState)(!1), S = (0, U.useRef)(r), C = a ? g || void 0 : r;
	(0, U.useEffect)(() => {
		if (!a || !e) return;
		let t = Cr(e).namespace;
		t && _(t);
	}, [a, e]), (0, U.useEffect)(() => {
		a && S.current !== r && (S.current = r, e || _(r ?? ""));
	}, [
		a,
		r,
		e
	]);
	let w = c.join(",");
	(0, U.useEffect)(() => {
		if (a && !C) {
			y({}), x(!1);
			return;
		}
		let e = !1;
		return x(!0), (C ? n(c, C) : n(c)).then((t) => {
			e || y(t);
		}).finally(() => {
			e || x(!1);
		}), () => {
			e = !0;
		};
	}, [
		a,
		C,
		n,
		w
	]);
	let T = e && Cr(e).name ? e : "", E = (0, U.useMemo)(() => Er(C, v, c, T), [
		C,
		v,
		c,
		T
	]), D = (0, U.useMemo)(() => {
		if (!l || !T || b) return !1;
		let { keys: e, names: t } = wr(C, v, c);
		return !e.has(T) && !t.has(T);
	}, [
		l,
		T,
		b,
		C,
		v,
		c
	]), O = gr[Tr(c, T)], k = O.Icon, A = d && !b && E.length === 1 && (!T || T === E[0].value) ? E[0] : void 0, j = /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i("flex items-center gap-2", a ? "min-w-0 flex-1" : m),
		children: [/* @__PURE__ */ (0, W.jsx)(k, {
			className: "h-4 w-4 shrink-0 text-muted-foreground",
			title: O.label,
			"aria-label": O.label
		}), /* @__PURE__ */ (0, W.jsx)("div", {
			className: "min-w-0 flex-1",
			children: A ? /* @__PURE__ */ (0, W.jsx)("span", {
				className: "block truncate text-sm text-foreground",
				title: A.label,
				"aria-label": "Workload",
				children: A.label
			}) : /* @__PURE__ */ (0, W.jsx)(He, {
				options: E,
				value: T,
				onChange: (e) => t(e || (a ? xr(C) : "")),
				ariaLabel: "Workload",
				allowCustomValue: u,
				loading: b,
				invalid: D,
				disabled: f,
				placeholder: p
			})
		})]
	});
	if (!a) return j;
	if (!o) throw Error("WorkloadPicker namespace selection requires loadNamespaces");
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i("flex w-full min-w-0 flex-nowrap items-center gap-density-2", m),
		children: [/* @__PURE__ */ (0, W.jsx)(Yn, {
			className: "min-w-0 shrink grow-0 basis-28",
			value: g,
			onChange: (e) => {
				_(e), y({}), s?.(e || void 0), t(xr(e));
			},
			loadNamespaces: o,
			strict: !0
		}), j]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/FilterBar.tsx
var Or = 8, kr = 44, Ar = "flex items-center gap-2 overflow-visible bg-background py-1.5";
function jr({ search: e, filters: t, timeRange: n, dateRange: r, children: o, leading: s, trailing: c, className: l, autoSubmit: u = !0, onApply: d, applyLabel: f = "Apply", isPending: p = !1, overflowMode: m = "responsive", overlay: h }) {
	let g = !!(n || r), _ = !u && !!d, v = (0, U.useMemo)(() => {
		let e = t ?? [];
		return [...e.filter((e) => e.placement !== "overflow"), ...e.filter((e) => e.placement === "overflow")];
	}, [t]), y = v.findIndex((e) => e.placement === "overflow"), b = y === -1 ? v.length : y, x = b < v.length, C = me(ve("md")), w = me(ve("sm")), { ref: T, wider: E } = un(S.sm), D = C || !E, O = (0, U.useMemo)(() => ({
		autoSubmit: u,
		compact: D
	}), [u, D]), k = m === "responsive" ? C : w, A = v.length > 0 && (m === "responsive" || k || x), j = (0, U.useRef)(null), M = (0, U.useRef)(null), N = (0, U.useRef)(/* @__PURE__ */ new Map()), ee = (0, U.useRef)(/* @__PURE__ */ new Map()), P = (0, U.useMemo)(() => v.map((e) => `${e.key}:${e.placement ?? "auto"}`).join("\0"), [v]), [te, ne] = (0, U.useState)(b), F = (0, U.useRef)(v), I = (0, U.useRef)(A), L = (0, U.useRef)(k), re = (0, U.useRef)(b);
	F.current = v, I.current = A, L.current = k, re.current = b, (0, U.useLayoutEffect)(() => {
		ne(I.current && L.current ? 0 : re.current);
	}, [P, k]);
	let R = (0, U.useCallback)(() => {
		let e = F.current, t = re.current;
		if (!I.current) {
			ne(e.length);
			return;
		}
		if (L.current) {
			ne(0);
			return;
		}
		let n = j.current;
		if (!n || (n.parentElement?.getBoundingClientRect().width ?? 0) <= 0) return;
		let r = Math.floor(n.getBoundingClientRect().width), i = e.slice(0, t).map((e) => {
			let t = N.current.get(e.key)?.getBoundingClientRect().width ?? 0;
			if (t > 0) {
				let n = Math.ceil(t);
				return ee.current.set(e.key, n), n;
			}
			return ee.current.get(e.key) ?? Hi(e);
		}), a = Math.ceil(M.current?.getBoundingClientRect().width ?? kr), o = a > 0 ? Or : 0, s = Vi(i), c = !(t < e.length) && s <= r ? t : Bi(i, Math.max(0, r - a - o));
		ne((e) => e === c ? e : c);
	}, []);
	(0, U.useLayoutEffect)(() => {
		R();
	}, [R]), (0, U.useEffect)(() => {
		if (!A) return;
		let e = j.current, t = M.current, n = typeof ResizeObserver > "u" ? null : ResizeObserver;
		if (!n) return window.addEventListener("resize", R), R(), () => window.removeEventListener("resize", R);
		let r = new n(() => R());
		return e && r.observe(e), t && r.observe(t), window.addEventListener("resize", R), R(), () => {
			r.disconnect(), window.removeEventListener("resize", R);
		};
	}, [
		R,
		A,
		k
	]);
	let ie = Math.min(te, b), ae = A ? v.slice(0, ie) : v, oe = A ? v.slice(ie) : [], se = oe.filter(Jt).length, ce = (0, U.useCallback)((e, t) => {
		t ? N.current.set(e, t) : N.current.delete(e);
	}, []), le = /* @__PURE__ */ (0, W.jsx)(B.Provider, {
		value: O,
		children: /* @__PURE__ */ (0, W.jsxs)("div", {
			ref: T,
			"data-slot": "filter-bar",
			className: i(Ar, m === "wrap" ? "flex-wrap max-sm:flex-nowrap" : "flex-wrap max-sm:flex-nowrap md:flex-nowrap", l, h && "invisible pointer-events-none"),
			...h ? { "aria-hidden": !0 } : {},
			children: [
				s && /* @__PURE__ */ (0, W.jsx)("div", {
					className: "flex shrink-0 items-center gap-2",
					children: s
				}),
				e && /* @__PURE__ */ (0, W.jsx)(ti, { search: e }),
				o,
				v.length > 0 && /* @__PURE__ */ (0, W.jsx)("div", {
					ref: j,
					"data-filter-bar-list": !0,
					className: i("flex min-w-0 items-center gap-2", m === "wrap" ? "flex-wrap" : "flex-1"),
					children: ae.map((t, n) => /* @__PURE__ */ (0, W.jsx)("div", {
						ref: (e) => ce(t.key, e),
						"data-filter-bar-item": t.key,
						className: "min-w-0 shrink-0",
						children: Pr(t, !e && n === 0)
					}, t.key))
				}),
				A && oe.length > 0 && /* @__PURE__ */ (0, W.jsx)(Fr, {
					triggerRef: M,
					filters: oe,
					activeHidden: se,
					allHidden: ae.length === 0,
					mobilePage: w,
					...d ? { onApply: d } : {}
				}),
				(g || c || _) && /* @__PURE__ */ (0, W.jsxs)("div", {
					className: "ml-auto flex min-w-0 items-center justify-end gap-2 max-md:flex-nowrap md:shrink-0 md:flex-wrap",
					children: [
						r && /* @__PURE__ */ (0, W.jsx)(gi, {
							kind: "date",
							label: "Date range",
							...r
						}),
						n && /* @__PURE__ */ (0, W.jsx)(gi, {
							kind: n.timeEnabled === !1 ? "date" : "time",
							label: n.timeEnabled === !1 ? "Date range" : "Time range",
							...n
						}),
						c,
						_ && /* @__PURE__ */ (0, W.jsx)(a, {
							type: "button",
							variant: "default",
							size: "sm",
							disabled: p,
							onClick: d,
							children: p ? "Loading…" : f
						})
					]
				})
			]
		})
	});
	return h ? /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "grid",
		children: [/* @__PURE__ */ (0, W.jsx)("div", {
			className: "col-start-1 row-start-1 min-w-0",
			children: le
		}), /* @__PURE__ */ (0, W.jsx)("div", {
			"data-slot": "filter-bar-overlay",
			className: i(Ar, "col-start-1 row-start-1 min-w-0 flex-wrap"),
			children: h
		})]
	}) : le;
}
function Mr({ filter: e, chrome: t = "full", autoSubmit: n = !0 }) {
	let r = (0, U.useMemo)(() => ({
		autoSubmit: n,
		compact: !1
	}), [n]);
	return /* @__PURE__ */ (0, W.jsx)(B.Provider, {
		value: r,
		children: /* @__PURE__ */ (0, W.jsx)(Nr, {
			filter: e,
			chrome: t
		})
	});
}
function Nr({ filter: e, chrome: t = "full" }) {
	return /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
		e.kind === "lookup" && /* @__PURE__ */ (0, W.jsx)(ii, {
			filter: e,
			grow: !0
		}),
		e.kind === "lookup-multi" && /* @__PURE__ */ (0, W.jsx)(ai, {
			filter: e,
			grow: !0
		}),
		e.kind === "multi" && /* @__PURE__ */ (0, W.jsx)(Oe, {
			filter: e,
			chrome: t
		}),
		e.kind === "tree" && /* @__PURE__ */ (0, W.jsx)(ln, {
			filter: e,
			grow: !0
		}),
		e.kind === "nested-multi" && /* @__PURE__ */ (0, W.jsx)(fi, {
			filter: e,
			chrome: t
		}),
		e.kind === "workload" && /* @__PURE__ */ (0, W.jsx)(Vr, {
			filter: e,
			grow: !0
		}),
		e.kind === "select-multi" && /* @__PURE__ */ (0, W.jsx)(hi, {
			filter: e,
			grow: !0
		}),
		(e.kind === "number" || e.kind === "duration") && /* @__PURE__ */ (0, W.jsx)(ci, {
			filter: e,
			chrome: t
		}),
		e.kind === "enum" && /* @__PURE__ */ (0, W.jsx)(Qr, {
			filter: e,
			grow: !0
		}),
		e.kind === "boolean" && /* @__PURE__ */ (0, W.jsx)(ei, { filter: e }),
		e.kind === "tristate" && /* @__PURE__ */ (0, W.jsx)($r, { filter: e }),
		e.kind === "date-range" && /* @__PURE__ */ (0, W.jsx)(Zr, {
			kind: e.timeEnabled === !1 ? "date" : "time",
			label: e.label,
			...Kt(e)
		}),
		e.kind === "text" && /* @__PURE__ */ (0, W.jsx)(ni, {
			filter: e,
			grow: !0
		})
	] });
}
function Pr(e, t) {
	return e.kind === "lookup" ? /* @__PURE__ */ (0, W.jsx)(ii, {
		filter: e,
		grow: t
	}) : e.kind === "lookup-multi" ? /* @__PURE__ */ (0, W.jsx)(ai, {
		filter: e,
		grow: t
	}) : e.kind === "multi" ? /* @__PURE__ */ (0, W.jsx)(be, {
		filter: e,
		grow: t
	}) : e.kind === "tree" ? /* @__PURE__ */ (0, W.jsx)(ln, {
		filter: e,
		grow: t
	}) : e.kind === "nested-multi" ? /* @__PURE__ */ (0, W.jsx)(pi, {
		filter: e,
		grow: t
	}) : e.kind === "workload" ? /* @__PURE__ */ (0, W.jsx)(Vr, {
		filter: e,
		grow: t
	}) : e.kind === "select-multi" ? /* @__PURE__ */ (0, W.jsx)(hi, {
		filter: e,
		grow: t
	}) : e.kind === "number" || e.kind === "duration" ? /* @__PURE__ */ (0, W.jsx)(si, {
		filter: e,
		grow: t
	}) : e.kind === "enum" ? /* @__PURE__ */ (0, W.jsx)(Qr, {
		filter: e,
		grow: t
	}) : e.kind === "boolean" ? /* @__PURE__ */ (0, W.jsx)(ei, { filter: e }) : e.kind === "tristate" ? /* @__PURE__ */ (0, W.jsx)($r, { filter: e }) : e.kind === "date-range" ? /* @__PURE__ */ (0, W.jsx)(gi, {
		kind: e.timeEnabled === !1 ? "date" : "time",
		label: e.label,
		...Kt(e)
	}) : /* @__PURE__ */ (0, W.jsx)(ni, {
		filter: e,
		grow: t
	});
}
function Fr({ triggerRef: e, filters: t, activeHidden: n, allHidden: r, mobilePage: o, onApply: s }) {
	let [c, l] = (0, U.useState)(!1), [u, d] = (0, U.useState)({}), f = t.length > 0, h = (0, U.useMemo)(() => t.map((e) => e.key).join("\0"), [t]), { refs: g, floatingStyles: _, context: y, floatingZ: b, getReferenceProps: x, getFloatingProps: S } = _i(c && !o, (e) => {
		d(ji(t)), l(e);
	}, "bottom-end"), C = (0, U.useCallback)(() => {
		d(ji(t)), l(!1);
	}, [t]), w = v([g.setReference, e]);
	(0, U.useEffect)(() => {
		f || C();
	}, [C, f]), (0, U.useEffect)(() => {
		c && d(ji(t));
	}, [h, c]);
	let T = t.map((e) => Pi(e, Object.hasOwn(u, e.key) ? u[e.key] : Ni(e), (t) => d((n) => ({
		...n,
		[e.key]: t
	})))), E = () => {
		Fi(t, u), s?.(), l(!1);
	}, D = /* @__PURE__ */ (0, W.jsx)("button", {
		type: "button",
		className: i("rounded text-primary transition-colors hover:bg-accent focus:bg-accent focus:outline-none disabled:text-muted-foreground", o ? "h-10 px-3 text-sm" : "px-1.5 py-0.5 text-xs"),
		onClick: () => T.forEach(qt),
		disabled: T.every((e) => !Jt(e)),
		children: "Clear all"
	}), O = /* @__PURE__ */ (0, W.jsx)(Lr, {
		onClose: C,
		onApply: E,
		mobilePage: o
	}), k = r ? "Filters" : "More filters";
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i("inline-flex shrink-0", !f && "invisible pointer-events-none"),
		"aria-hidden": !f || void 0,
		children: [
			/* @__PURE__ */ (0, W.jsxs)(a, {
				ref: w,
				type: "button",
				variant: "outline",
				size: "sm",
				"aria-label": k,
				tabIndex: f ? 0 : -1,
				title: f ? k : void 0,
				className: i("h-8 min-w-0 gap-1.5 px-2 text-xs font-normal max-md:h-9", n > 0 && "border-primary/40 text-primary"),
				...x(),
				"aria-expanded": c,
				children: [
					/* @__PURE__ */ (0, W.jsx)(p, {
						icon: ut,
						className: "text-[14px]"
					}),
					r && /* @__PURE__ */ (0, W.jsx)("span", { children: "Filters" }),
					n > 0 && /* @__PURE__ */ (0, W.jsx)("span", {
						className: "rounded-full bg-muted px-1.5 text-[10px] font-medium text-muted-foreground",
						children: n
					})
				]
			}),
			o && /* @__PURE__ */ (0, W.jsx)(m, {
				open: f && c,
				onClose: C,
				title: "Filters",
				headerSlot: D,
				footer: O,
				scrollBody: !1,
				children: /* @__PURE__ */ (0, W.jsx)(Ir, {
					filters: T,
					mobilePage: !0
				})
			}),
			!o && f && c && /* @__PURE__ */ (0, W.jsx)(pe, { children: /* @__PURE__ */ (0, W.jsx)(he, {
				context: y,
				modal: !1,
				children: /* @__PURE__ */ (0, W.jsxs)("div", {
					ref: g.setFloating,
					role: "dialog",
					"aria-label": "Overflow filters",
					style: {
						..._,
						zIndex: b
					},
					className: "flex max-h-[min(34rem,calc(100vh-2rem))] w-[min(34rem,calc(100vw-2rem))] flex-col overflow-hidden rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-lg shadow-black/10",
					...S(),
					children: [
						/* @__PURE__ */ (0, W.jsxs)("div", {
							className: "mb-2 flex items-center justify-between gap-2 px-1",
							children: [/* @__PURE__ */ (0, W.jsx)("div", {
								className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
								children: "Filters"
							}), /* @__PURE__ */ (0, W.jsxs)("div", {
								className: "flex items-center gap-1",
								children: [D, /* @__PURE__ */ (0, W.jsx)("button", {
									type: "button",
									"aria-label": "Close overflow filters",
									title: "Close",
									className: "inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none",
									onClick: C,
									children: /* @__PURE__ */ (0, W.jsx)(p, {
										icon: fe,
										className: "text-sm"
									})
								})]
							})]
						}),
						/* @__PURE__ */ (0, W.jsx)(Ir, { filters: T }),
						/* @__PURE__ */ (0, W.jsx)("div", {
							className: "mt-3",
							children: O
						})
					]
				})
			}) })
		]
	});
}
function Ir({ filters: e, mobilePage: t = !1 }) {
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: i("min-h-0 flex-1 divide-y divide-border/70 overflow-y-auto", t ? "bg-background" : "rounded-md border border-border/70"),
		children: e.map((e) => {
			let n = Jt(e);
			return /* @__PURE__ */ (0, W.jsxs)("div", {
				"data-overflow-filter-row": e.key,
				className: "grid grid-cols-[minmax(0,1fr)_auto] items-start gap-2 overflow-visible p-4 md:h-12 md:grid-cols-[minmax(7rem,10rem)_auto_minmax(0,1fr)_auto] md:items-center md:p-2",
				children: [
					/* @__PURE__ */ (0, W.jsxs)("label", {
						htmlFor: zr(e),
						className: "col-span-2 flex min-w-0 items-center gap-1 truncate text-[10px] font-medium uppercase tracking-wide text-muted-foreground md:col-span-1",
						title: e.description ?? e.label,
						children: [/* @__PURE__ */ (0, W.jsx)(d, {
							icon: Ce(e.icon, 12),
							className: "text-[12px] normal-case"
						}), /* @__PURE__ */ (0, W.jsx)("span", {
							className: "truncate",
							children: e.label
						})]
					}),
					/* @__PURE__ */ (0, W.jsx)("span", {
						className: "hidden text-sm text-muted-foreground md:block",
						children: "="
					}),
					/* @__PURE__ */ (0, W.jsx)("div", {
						className: "min-w-0 overflow-visible",
						children: /* @__PURE__ */ (0, W.jsx)(B.Provider, {
							value: {
								autoSubmit: !1,
								compact: !1
							},
							children: /* @__PURE__ */ (0, W.jsx)(Rr, { filter: e })
						})
					}),
					/* @__PURE__ */ (0, W.jsx)("button", {
						type: "button",
						"aria-label": `Clear ${e.label}`,
						title: `Clear ${e.label}`,
						className: i("inline-flex self-center items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none disabled:text-muted-foreground/40", t ? "h-10 w-10" : "h-6 w-6"),
						onClick: () => qt(e),
						disabled: !n,
						children: /* @__PURE__ */ (0, W.jsx)(p, {
							icon: fe,
							className: "text-sm"
						})
					})
				]
			}, e.key);
		})
	});
}
function Lr({ onClose: e, onApply: t, mobilePage: n = !1 }) {
	let r = n ? "h-10 px-4 text-sm" : "h-8 px-3 text-xs";
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "flex justify-end gap-2",
		children: [/* @__PURE__ */ (0, W.jsx)(a, {
			type: "button",
			variant: "ghost",
			size: "sm",
			className: r,
			onClick: e,
			children: "Close"
		}), /* @__PURE__ */ (0, W.jsx)(a, {
			type: "button",
			variant: "default",
			size: "sm",
			className: r,
			onClick: t,
			children: "Apply"
		})]
	});
}
function Rr({ filter: e }) {
	return e.kind === "date-range" ? /* @__PURE__ */ (0, W.jsx)(gi, {
		kind: e.timeEnabled === !1 ? "date" : "time",
		label: e.label,
		...Kt(e)
	}) : e.kind === "text" ? /* @__PURE__ */ (0, W.jsx)(Ur, { filter: e }) : e.kind === "lookup" ? /* @__PURE__ */ (0, W.jsx)(Wr, { filter: e }) : e.kind === "lookup-multi" ? /* @__PURE__ */ (0, W.jsx)(Gr, { filter: e }) : e.kind === "enum" ? /* @__PURE__ */ (0, W.jsx)(qr, { filter: e }) : e.kind === "boolean" ? /* @__PURE__ */ (0, W.jsx)(Jr, { filter: e }) : e.kind === "select-multi" ? /* @__PURE__ */ (0, W.jsx)(Xr, { filter: e }) : e.kind === "multi" ? /* @__PURE__ */ (0, W.jsx)(be, {
		filter: e,
		grow: !0
	}) : e.kind === "tree" ? /* @__PURE__ */ (0, W.jsx)(ln, {
		filter: e,
		grow: !0
	}) : e.kind === "nested-multi" ? /* @__PURE__ */ (0, W.jsx)(pi, {
		filter: e,
		grow: !0
	}) : e.kind === "workload" ? /* @__PURE__ */ (0, W.jsx)(Vr, {
		filter: e,
		grow: !0
	}) : e.kind === "tristate" ? /* @__PURE__ */ (0, W.jsx)(Yr, { filter: e }) : /* @__PURE__ */ (0, W.jsx)(si, {
		filter: e,
		grow: !0
	});
}
function zr(e) {
	return `filterbar-overflow-${e.key}`;
}
function Br(e) {
	return e.map((e) => ({
		value: e.value,
		label: e.label ?? e.value,
		...e.disabled === void 0 ? {} : { disabled: e.disabled },
		...e.count === void 0 ? {} : { trailing: /* @__PURE__ */ (0, W.jsx)("span", { children: e.count.toLocaleString() }) }
	}));
}
function Vr({ filter: e, grow: t }) {
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: i("min-w-[14rem]", t && "flex-1", e.className),
		children: /* @__PURE__ */ (0, W.jsx)(Dr, {
			value: e.value,
			onChange: e.onChange,
			loadWorkloads: e.loadWorkloads,
			...e.kinds ? { kinds: e.kinds } : {},
			...e.collapseSingleOption ? { collapseSingleOption: !0 } : {},
			allowCustomValue: !1,
			strict: !0,
			...e.disabled === void 0 ? {} : { disabled: e.disabled },
			placeholder: `Select ${e.label.toLowerCase()}…`
		})
	});
}
function Hr(e) {
	return i("h-8 w-full rounded-md border border-input bg-background px-2 text-sm text-foreground outline-none placeholder:text-placeholder focus-visible:ring-2 focus-visible:ring-ring", e && "cursor-not-allowed opacity-60");
}
function Ur({ filter: e }) {
	let [t, n] = vi(e.value, e.onChange);
	return /* @__PURE__ */ (0, W.jsx)("input", {
		id: zr(e),
		type: "text",
		"aria-label": e.label,
		className: Hr(e.disabled),
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		value: t,
		disabled: e.disabled,
		onChange: (e) => n(e.target.value)
	});
}
function Wr({ filter: e }) {
	let [t, n] = vi(e.value, e.onChange);
	return e.inputType === "date" ? /* @__PURE__ */ (0, W.jsx)(dn, {
		id: zr(e),
		"aria-label": e.label,
		className: "w-full",
		inputClassName: Hr(e.disabled),
		buttonClassName: "right-1",
		placeholder: e.placeholder,
		value: t,
		disabled: e.disabled,
		onChange: n
	}) : e.inputType === "number" ? /* @__PURE__ */ (0, W.jsx)("input", {
		id: zr(e),
		type: "number",
		"aria-label": e.label,
		className: Hr(e.disabled),
		placeholder: e.placeholder,
		value: t,
		disabled: e.disabled,
		onChange: (e) => n(e.target.value)
	}) : /* @__PURE__ */ (0, W.jsx)(He, {
		id: zr(e),
		options: Br(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		size: "sm",
		className: "w-full",
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function Gr({ filter: e }) {
	return /* @__PURE__ */ (0, W.jsx)(He, {
		multiple: !0,
		id: zr(e),
		options: Br(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		footer: oi(e),
		size: "sm",
		className: "w-full",
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function Kr(e) {
	return e.map((e) => ({
		value: e.value,
		label: e.label ?? e.value
	}));
}
function qr({ filter: e }) {
	return /* @__PURE__ */ (0, W.jsx)(He, {
		id: zr(e),
		options: Kr(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		size: "sm",
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		className: "w-full",
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function Jr({ filter: e }) {
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: "flex h-8 items-center",
		children: /* @__PURE__ */ (0, W.jsx)("input", {
			id: zr(e),
			type: "checkbox",
			"aria-label": e.label,
			className: "h-4 w-4 accent-primary",
			checked: e.value,
			disabled: e.disabled,
			onChange: (t) => e.onChange(t.target.checked)
		})
	});
}
function Yr({ filter: e }) {
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: "flex h-8 items-center",
		children: /* @__PURE__ */ (0, W.jsx)(Jn, {
			id: zr(e),
			value: e.value,
			onChange: e.onChange,
			label: e.label,
			size: "md",
			...e.labels ? { labels: e.labels } : {},
			...e.disabled ? { disabled: e.disabled } : {}
		})
	});
}
function Xr({ filter: e }) {
	return /* @__PURE__ */ (0, W.jsx)(mn, {
		options: e.options,
		value: e.value,
		onChange: e.onChange,
		ariaLabel: `${e.label} filter`,
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...e.disabled === void 0 ? {} : { disabled: e.disabled },
		triggerClassName: "h-8 w-full rounded-md border border-input bg-background px-2 text-sm shadow-none",
		menuClassName: "left-auto right-0"
	});
}
function Zr({ kind: e, label: t, from: n = "", to: r = "", onApply: i, presets: a, timeEnabled: o, timeZone: s, timeZones: c, fromPlaceholder: l, toPlaceholder: u, emptyLabel: d, disabled: f }) {
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: "w-72 p-3 text-popover-foreground",
		children: /* @__PURE__ */ (0, W.jsx)(Dn, {
			kind: e,
			label: t,
			align: "left",
			from: n,
			to: r,
			onApply: i,
			...f === void 0 ? {} : { disabled: f },
			...a ? { presets: a } : {},
			...o === void 0 ? {} : { timeEnabled: o },
			...s ? { timeZone: s } : {},
			...c ? { timeZones: c } : {},
			...l ? { fromPlaceholder: l } : {},
			...u ? { toPlaceholder: u } : {},
			...d ? { emptyLabel: d } : {},
			panelClassName: "left-0 right-auto"
		})
	});
}
function Qr({ filter: e, grow: t }) {
	return /* @__PURE__ */ (0, W.jsx)(He, {
		...Ee(e),
		options: Kr(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		size: "sm",
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		className: i(De(t), e.className),
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function $r({ filter: e }) {
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		title: e.description,
		className: i("flex h-8 shrink-0 items-center gap-2 rounded-md border border-input bg-muted/30 px-2 text-xs", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, W.jsx)(Jn, {
			value: e.value,
			onChange: e.onChange,
			label: e.label,
			...e.labels ? { labels: e.labels } : {},
			...e.disabled ? { disabled: e.disabled } : {}
		}), /* @__PURE__ */ (0, W.jsx)(ri, {
			icon: e.icon,
			label: e.label
		})]
	});
}
function ei({ filter: e }) {
	return /* @__PURE__ */ (0, W.jsxs)("label", {
		title: e.description,
		className: i("flex h-8 shrink-0 items-center gap-2 rounded-md border border-input bg-muted/30 px-2 text-xs", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, W.jsx)("input", {
			type: "checkbox",
			"aria-label": e.label,
			className: "h-3.5 w-3.5 accent-primary",
			checked: e.value,
			disabled: e.disabled,
			onChange: (t) => e.onChange(t.target.checked)
		}), /* @__PURE__ */ (0, W.jsx)(ri, {
			icon: e.icon,
			label: e.label
		})]
	});
}
function ti({ search: e }) {
	let [t, n] = vi(e.value, e.onChange);
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: "flex min-w-0 flex-1 items-center gap-2 md:flex-[1_1_14rem] md:min-w-[14rem] md:max-w-[24rem]",
		children: /* @__PURE__ */ (0, W.jsxs)("label", {
			className: i("flex h-8 min-w-0 flex-1 items-center rounded-md border border-input bg-background px-3 text-sm", e.className),
			children: [t.trim() ? /* @__PURE__ */ (0, W.jsx)("span", {
				className: "mr-2 whitespace-nowrap text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
				children: e.ariaLabel ?? "Search"
			}) : /* @__PURE__ */ (0, W.jsx)(p, {
				icon: Ue,
				className: "mr-2 shrink-0 text-muted-foreground"
			}), /* @__PURE__ */ (0, W.jsx)("input", {
				type: "search",
				"aria-label": e.ariaLabel ?? e.placeholder ?? "Search",
				className: "w-full min-w-0 bg-transparent outline-none placeholder:text-placeholder",
				placeholder: e.placeholder ?? "Search…",
				value: t,
				onChange: (e) => n(e.target.value)
			})]
		})
	});
}
function ni({ filter: e, grow: t }) {
	let [n, r] = vi(e.value, e.onChange);
	return /* @__PURE__ */ (0, W.jsxs)("label", {
		title: e.description,
		className: i("flex h-8 items-center gap-2 rounded-md border border-input bg-muted/30 pl-2 pr-2 text-xs", t ? "min-w-[12rem] max-w-[18rem] flex-1" : "min-w-[11rem] max-w-[15rem] shrink-0", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, W.jsx)(ri, {
			icon: e.icon,
			label: e.label
		}), /* @__PURE__ */ (0, W.jsx)("input", {
			type: "text",
			"aria-label": e.label,
			className: "w-full min-w-0 bg-transparent text-sm text-foreground outline-none placeholder:text-placeholder disabled:cursor-not-allowed",
			...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
			value: n,
			disabled: e.disabled,
			onChange: (e) => r(e.target.value)
		})]
	});
}
function ri({ icon: e, label: t }) {
	return /* @__PURE__ */ (0, W.jsxs)("span", {
		className: "flex items-center gap-1 whitespace-nowrap font-medium uppercase tracking-wide text-muted-foreground",
		children: [/* @__PURE__ */ (0, W.jsx)(d, {
			icon: e,
			className: "text-[13px] normal-case"
		}), t]
	});
}
function ii({ filter: e, grow: t }) {
	let [n, r] = vi(e.value, e.onChange);
	if (e.inputType === "date" || e.inputType === "number") {
		let a = e.inputType === "date" && n ? P(n) : void 0;
		return /* @__PURE__ */ (0, W.jsxs)("label", {
			title: a ?? e.description,
			className: i("flex h-8 items-center gap-2 rounded-md border border-input bg-muted/30 pl-2 pr-2 text-xs", De(t), e.disabled && "opacity-60", e.className),
			children: [/* @__PURE__ */ (0, W.jsx)(ri, {
				icon: e.icon,
				label: e.label
			}), e.inputType === "date" ? /* @__PURE__ */ (0, W.jsx)(dn, {
				"aria-label": e.label,
				className: "w-full",
				inputClassName: "w-full min-w-0 border-0 bg-transparent px-0 pr-6 text-sm text-foreground shadow-none focus-visible:ring-0",
				buttonClassName: "right-0",
				placeholder: e.placeholder,
				value: n,
				disabled: e.disabled,
				onChange: r
			}) : /* @__PURE__ */ (0, W.jsx)("input", {
				type: "number",
				"aria-label": e.label,
				className: "w-full min-w-0 bg-transparent text-sm text-foreground outline-none placeholder:text-placeholder disabled:cursor-not-allowed",
				placeholder: e.placeholder,
				value: n,
				disabled: e.disabled,
				onChange: (e) => r(e.target.value)
			})]
		});
	}
	return /* @__PURE__ */ (0, W.jsx)(He, {
		...Ee(e),
		options: Br(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		size: "sm",
		className: i(De(t), e.className),
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function ai({ filter: e, grow: t }) {
	return /* @__PURE__ */ (0, W.jsx)(He, {
		multiple: !0,
		...Ee(e),
		options: Br(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		footer: oi(e),
		size: "sm",
		className: i(De(t), e.className),
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...e.onSearch === void 0 ? {} : { onSearch: e.onSearch },
		...e.loading === void 0 ? {} : { loading: e.loading },
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function oi(e) {
	let t = e.truncated && e.total ? Math.max(e.total - e.options.length, 0) : 0;
	if (t !== 0) return /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
		"… and ",
		t.toLocaleString(),
		" more",
		e.onSearch ? " — type to search all" : ""
	] });
}
function si({ filter: e, grow: t }) {
	let [n, r] = (0, U.useState)(!1), o = _i(n, r), s = xi(e), [c, l] = yi(e.value, (t) => ki(e, t)), u = Ei(Ti(c.min) ?? s.min, s.min, s.max), f = Ei(Ti(c.max) ?? s.max, s.min, s.max), m = Math.min(u, f), h = Math.max(u, f), g = Si(e, s, c), _ = e.domainMin !== void 0 && e.domainMax !== void 0;
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		title: e.description,
		className: i("inline-flex min-w-0", t ? "min-w-[8rem] max-w-[12rem] flex-1" : "shrink-0", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, W.jsxs)(a, {
			ref: o.refs.setReference,
			type: "button",
			variant: "outline",
			size: "sm",
			"aria-label": `${e.label} filter`,
			disabled: e.disabled,
			className: i("min-w-0 gap-2 font-normal", t ? "w-full max-w-[12rem] justify-between" : "w-auto max-w-[9.5rem] px-2.5", g === e.label && "text-muted-foreground"),
			...o.getReferenceProps(),
			children: [
				/* @__PURE__ */ (0, W.jsx)(d, {
					icon: e.icon,
					className: "text-[14px] text-muted-foreground"
				}),
				/* @__PURE__ */ (0, W.jsx)("span", {
					className: "truncate",
					children: g
				}),
				/* @__PURE__ */ (0, W.jsx)(p, {
					icon: n ? _e : z,
					className: "text-muted-foreground"
				})
			]
		}), n && /* @__PURE__ */ (0, W.jsx)(pe, { children: /* @__PURE__ */ (0, W.jsx)(he, {
			context: o.context,
			modal: !1,
			children: /* @__PURE__ */ (0, W.jsxs)("div", {
				ref: o.refs.setFloating,
				role: "dialog",
				"aria-label": e.label,
				style: {
					...o.floatingStyles,
					zIndex: o.floatingZ
				},
				className: "min-w-[18rem] max-w-[22rem] rounded-md border border-border bg-popover p-3 text-popover-foreground shadow-lg shadow-black/5",
				...o.getFloatingProps(),
				children: [/* @__PURE__ */ (0, W.jsxs)("div", {
					className: "mb-3 flex items-center justify-between gap-2",
					children: [/* @__PURE__ */ (0, W.jsx)("div", {
						className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
						children: e.label
					}), /* @__PURE__ */ (0, W.jsx)("button", {
						type: "button",
						className: "text-[10px] text-primary disabled:text-muted-foreground",
						onClick: () => {
							l({}), ki(e, {});
						},
						disabled: !String(c.min ?? "").trim() && !String(c.max ?? "").trim(),
						children: "Clear all"
					})]
				}), /* @__PURE__ */ (0, W.jsxs)("div", {
					className: "space-y-3",
					children: [_ && /* @__PURE__ */ (0, W.jsxs)("div", {
						className: "space-y-2",
						children: [/* @__PURE__ */ (0, W.jsxs)("div", {
							className: "flex items-center justify-between text-[11px] text-muted-foreground",
							children: [/* @__PURE__ */ (0, W.jsx)("span", { children: Di(s.min, e) }), /* @__PURE__ */ (0, W.jsx)("span", { children: Di(s.max, e) })]
						}), /* @__PURE__ */ (0, W.jsx)("div", {
							className: "relative h-6",
							children: /* @__PURE__ */ (0, W.jsx)(xn, {
								min: s.min,
								max: s.max,
								step: s.step,
								value: [m, h],
								ariaLabelMin: `${e.label} minimum slider`,
								ariaLabelMax: `${e.label} maximum slider`,
								onChange: ([t, n]) => l(wi([t, n], s, e))
							})
						})]
					}), /* @__PURE__ */ (0, W.jsx)(li, {
						filter: e,
						value: c,
						onChange: l,
						bounds: s
					})]
				})]
			})
		}) })]
	});
}
function ci({ filter: e, chrome: t = "full" }) {
	let n = xi(e), [r, a] = yi(e.value, (t) => ki(e, t)), o = Ei(Ti(r.min) ?? n.min, n.min, n.max), s = Ei(Ti(r.max) ?? n.max, n.min, n.max), c = Math.min(o, s), l = Math.max(o, s), u = t === "embedded", d = e.domainMin !== void 0 && e.domainMax !== void 0;
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		"data-filter-panel-chrome": t,
		className: i("min-w-[18rem] max-w-[22rem] text-popover-foreground", u ? "p-0" : "rounded-md border border-border bg-popover p-3 shadow-sm shadow-black/5"),
		children: [!u && /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "mb-3 flex items-center justify-between gap-2",
			children: [/* @__PURE__ */ (0, W.jsx)("div", {
				className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
				children: e.label
			}), /* @__PURE__ */ (0, W.jsx)("button", {
				type: "button",
				className: "text-[10px] text-primary disabled:text-muted-foreground",
				onClick: () => {
					a({}), ki(e, {});
				},
				disabled: !String(r.min ?? "").trim() && !String(r.max ?? "").trim(),
				children: "Clear all"
			})]
		}), /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "space-y-3",
			children: [d && /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "space-y-2",
				children: [/* @__PURE__ */ (0, W.jsxs)("div", {
					className: "flex items-center justify-between text-[11px] text-muted-foreground",
					children: [/* @__PURE__ */ (0, W.jsx)("span", { children: Di(n.min, e) }), /* @__PURE__ */ (0, W.jsx)("span", { children: Di(n.max, e) })]
				}), /* @__PURE__ */ (0, W.jsx)("div", {
					className: "relative h-6",
					children: /* @__PURE__ */ (0, W.jsx)(xn, {
						min: n.min,
						max: n.max,
						step: n.step,
						value: [c, l],
						ariaLabelMin: `${e.label} minimum slider`,
						ariaLabelMax: `${e.label} maximum slider`,
						onChange: ([t, r]) => a(wi([t, r], n, e))
					})
				})]
			}), /* @__PURE__ */ (0, W.jsx)(li, {
				filter: e,
				value: r,
				onChange: a,
				bounds: n
			})]
		})]
	});
}
function li({ filter: e, value: t, onChange: n, bounds: r }) {
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "space-y-2",
		children: [/* @__PURE__ */ (0, W.jsx)(ui, {
			filter: e,
			value: t,
			onChange: n,
			bounds: r,
			edge: "min"
		}), /* @__PURE__ */ (0, W.jsx)(ui, {
			filter: e,
			value: t,
			onChange: n,
			bounds: r,
			edge: "max"
		})]
	});
}
function ui({ filter: e, value: t, onChange: n, bounds: r, edge: i }) {
	let a = i === "min", o = a ? "minimum" : "maximum", s = a ? t.minOperator ?? ">=" : t.maxOperator ?? "<=", c = a ? t.min ?? "" : t.max ?? "", l = e.kind === "duration" ? t : void 0, u = a ? l?.minUnit : l?.maxUnit, d = (r) => {
		let i = a ? {
			...t,
			min: r
		} : {
			...t,
			max: r
		};
		n(e.kind === "duration" ? i : Ci(i, a ? "min-input" : "max-input"));
	}, f = (e) => n(a ? {
		...t,
		minOperator: e
	} : {
		...t,
		maxOperator: e
	});
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "grid grid-cols-[2.5rem_3.5rem_minmax(0,1fr)_auto] items-center gap-1.5",
		children: [
			/* @__PURE__ */ (0, W.jsx)("span", {
				className: "text-[10px] text-muted-foreground",
				children: a ? "Min" : "Max"
			}),
			/* @__PURE__ */ (0, W.jsx)("select", {
				"aria-label": `${e.label} ${o} operator`,
				className: "h-8 rounded-md border border-input bg-background px-1 text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring",
				value: s,
				disabled: e.disabled,
				onChange: (e) => f(e.target.value),
				children: (a ? [">=", ">"] : ["<=", "<"]).map((e) => /* @__PURE__ */ (0, W.jsx)("option", {
					value: e,
					children: e
				}, e))
			}),
			/* @__PURE__ */ (0, W.jsx)("input", {
				type: "number",
				inputMode: "decimal",
				step: r.step,
				"aria-label": `${e.label} ${o}`,
				className: "h-8 min-w-0 rounded-md border border-input bg-background px-2 text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring",
				...(a ? e.minPlaceholder : e.maxPlaceholder) === void 0 ? {} : { placeholder: a ? e.minPlaceholder : e.maxPlaceholder },
				value: c,
				disabled: e.disabled,
				onChange: (e) => d(e.target.value)
			}),
			/* @__PURE__ */ (0, W.jsx)(di, {
				filter: e,
				value: t,
				edge: i,
				...u ? { unit: u } : {},
				onChange: n
			})
		]
	});
}
function di({ filter: e, value: t, edge: n, unit: r, onChange: i }) {
	if (e.kind === "number") return e.unit ? /* @__PURE__ */ (0, W.jsx)("span", {
		className: "min-w-8 text-xs text-muted-foreground",
		children: e.unit
	}) : null;
	let a = r ?? e.unit ?? "ms";
	return /* @__PURE__ */ (0, W.jsx)("select", {
		"aria-label": `${e.label} ${n === "min" ? "minimum" : "maximum"} unit`,
		className: "h-8 rounded-md border border-input bg-background px-1 text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring",
		value: a,
		disabled: e.disabled,
		onChange: (e) => {
			i(n === "min" ? {
				...t,
				minUnit: e.target.value
			} : {
				...t,
				maxUnit: e.target.value
			});
		},
		children: [
			"ms",
			"s",
			"m",
			"h"
		].map((e) => /* @__PURE__ */ (0, W.jsx)("option", {
			value: e,
			children: e
		}, e))
	});
}
function fi({ filter: e, chrome: t = "full" }) {
	let [n, r] = (0, U.useState)(e.groups[0]?.groupKey ?? null), [a, o] = $e(e.value, e.onChange), s = e.groups, c = t === "embedded", l = (0, U.useMemo)(() => s.find((e) => e.groupKey === n) ?? s[0] ?? null, [s, n]), u = (0, U.useMemo)(() => {
		let e = {};
		for (let t of s) {
			let n = 0;
			for (let e of t.options) a[e.value] && (n += 1);
			e[t.groupKey] = n;
		}
		return e;
	}, [s, a]), d = (0, U.useMemo)(() => {
		let e = [], t = [];
		for (let n of s) (u[n.groupKey] ?? 0) > 0 ? e.push(n) : t.push(n);
		return [...e, ...t];
	}, [s, u]), f = (e) => {
		let t = s.find((t) => t.groupKey === e);
		if (!t) return;
		let n = { ...a };
		for (let e of t.options) delete n[e.value];
		o(n);
	};
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		role: "dialog",
		className: "flex",
		children: [/* @__PURE__ */ (0, W.jsxs)("div", {
			"data-filter-panel-chrome": t,
			className: i("min-w-[14rem] max-w-[16rem] text-popover-foreground", c ? "p-0" : "rounded-md border border-border bg-popover p-2 shadow-sm shadow-black/5"),
			children: [!c && /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "mb-2 flex items-center justify-between gap-2",
				children: [/* @__PURE__ */ (0, W.jsx)("div", {
					className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
					children: e.label
				}), /* @__PURE__ */ (0, W.jsx)("button", {
					type: "button",
					className: "text-[10px] text-primary disabled:text-muted-foreground",
					onClick: () => o({}),
					disabled: Object.keys(a).length === 0,
					children: "Clear all"
				})]
			}), /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "max-h-72 space-y-0.5 overflow-auto",
				children: [d.map((e) => {
					let t = u[e.groupKey] ?? 0, a = e.groupKey === n;
					return /* @__PURE__ */ (0, W.jsxs)("div", {
						role: "button",
						tabIndex: 0,
						onMouseEnter: () => r(e.groupKey),
						onFocus: () => r(e.groupKey),
						onClick: () => r(e.groupKey),
						onKeyDown: (t) => {
							(t.key === "Enter" || t.key === " " || t.key === "ArrowRight") && (t.preventDefault(), r(e.groupKey));
						},
						className: i("flex cursor-pointer items-center justify-between gap-2 rounded-md px-1.5 py-0.5 text-sm", "hover:bg-accent/50 focus-visible:bg-accent/50 focus-visible:outline-none", a && "bg-accent/60"),
						children: [
							/* @__PURE__ */ (0, W.jsx)("span", {
								className: "min-w-0 flex-1 truncate",
								children: e.label ?? e.groupKey
							}),
							t > 0 && /* @__PURE__ */ (0, W.jsxs)("span", {
								className: "rounded-full bg-primary/15 px-1.5 text-[10px] font-medium text-primary",
								children: [
									t,
									"/",
									e.options.length
								]
							}),
							/* @__PURE__ */ (0, W.jsx)(p, {
								icon: j,
								className: "shrink-0 text-muted-foreground"
							})
						]
					}, e.groupKey);
				}), s.length === 0 && /* @__PURE__ */ (0, W.jsx)("div", {
					className: "px-2 py-3 text-sm text-muted-foreground",
					children: "No groups"
				})]
			})]
		}), l && /* @__PURE__ */ (0, W.jsxs)("div", {
			onMouseEnter: () => r(l.groupKey),
			className: "ml-1.5 min-w-[16rem] max-w-[20rem] rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-sm shadow-black/5",
			children: [/* @__PURE__ */ (0, W.jsxs)("div", {
				className: "mb-2 flex items-center justify-between gap-2",
				children: [/* @__PURE__ */ (0, W.jsx)("div", {
					className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
					children: l.label ?? l.groupKey
				}), /* @__PURE__ */ (0, W.jsx)("button", {
					type: "button",
					className: "text-[10px] text-primary disabled:text-muted-foreground",
					onClick: () => f(l.groupKey),
					disabled: (u[l.groupKey] ?? 0) === 0,
					children: "Clear"
				})]
			}), /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "max-h-72 space-y-0.5 overflow-auto",
				children: [l.options.map((e) => {
					let t = a[e.value] ?? "neutral", n = e.title ?? ke(e);
					return /* @__PURE__ */ (0, W.jsx)("div", {
						role: "button",
						tabIndex: 0,
						"data-filter-option": e.value,
						className: "rounded-md px-1.5 py-0.5 hover:bg-accent/50 focus-visible:bg-accent/50 focus-visible:outline-none",
						onClick: () => o(Ae(a, e.value, Se(t))),
						onKeyDown: (n) => {
							n.key === "Enter" || n.key === " " ? (n.preventDefault(), o(Ae(a, e.value, Se(t)))) : n.key === "ArrowLeft" && (n.preventDefault(), r(null));
						},
						children: /* @__PURE__ */ (0, W.jsx)(Ne, {
							className: "w-full justify-between",
							label: mi(e, l.groupKey),
							mode: t,
							title: n,
							togglePosition: "right",
							onModeChange: (t) => o(Ae(a, e.value, t))
						})
					}, e.value);
				}), l.options.length === 0 && /* @__PURE__ */ (0, W.jsx)("div", {
					className: "px-2 py-3 text-sm text-muted-foreground",
					children: "No values"
				})]
			})]
		})]
	});
}
function pi({ filter: e, grow: t }) {
	let [n, r] = (0, U.useState)(!1), [o, s] = (0, U.useState)(null), [c, l] = $e(e.value, e.onChange), u = _i(n, (e) => {
		r(e), e || s(null);
	}), d = Te(e.label, c), f = e.groups, m = (0, U.useMemo)(() => f.find((e) => e.groupKey === o) ?? null, [f, o]), h = (0, U.useMemo)(() => {
		let e = {};
		for (let t of f) {
			let n = 0;
			for (let e of t.options) c[e.value] && (n += 1);
			e[t.groupKey] = n;
		}
		return e;
	}, [f, c]), g = (0, U.useMemo)(() => {
		let e = [], t = [];
		for (let n of f) (h[n.groupKey] ?? 0) > 0 ? e.push(n) : t.push(n);
		return [...e, ...t];
	}, [f, h]), _ = (e) => {
		let t = f.find((t) => t.groupKey === e);
		if (!t) return;
		let n = { ...c };
		for (let e of t.options) delete n[e.value];
		l(n);
	};
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		title: e.description,
		className: i("inline-flex min-w-0", t ? "min-w-[8rem] max-w-[12rem] flex-1" : "shrink-0", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, W.jsxs)(a, {
			ref: u.refs.setReference,
			type: "button",
			variant: "outline",
			size: "sm",
			"aria-label": `${e.label} filter`,
			disabled: e.disabled,
			className: i("min-w-0 gap-2 font-normal", t ? "w-full max-w-[12rem] justify-between" : "w-auto max-w-[8.5rem] px-2.5", d === e.label && "text-muted-foreground"),
			...u.getReferenceProps(),
			children: [/* @__PURE__ */ (0, W.jsx)("span", {
				className: "truncate",
				children: d
			}), /* @__PURE__ */ (0, W.jsx)(p, {
				icon: n ? _e : z,
				className: "text-muted-foreground"
			})]
		}), n && /* @__PURE__ */ (0, W.jsx)(pe, { children: /* @__PURE__ */ (0, W.jsx)(he, {
			context: u.context,
			modal: !1,
			children: /* @__PURE__ */ (0, W.jsxs)("div", {
				ref: u.refs.setFloating,
				role: "dialog",
				"aria-label": e.label,
				style: {
					...u.floatingStyles,
					zIndex: u.floatingZ
				},
				className: "flex",
				...u.getFloatingProps(),
				children: [/* @__PURE__ */ (0, W.jsxs)("div", {
					className: "min-w-[14rem] max-w-[16rem] rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-lg shadow-black/5",
					children: [/* @__PURE__ */ (0, W.jsxs)("div", {
						className: "mb-2 flex items-center justify-between gap-2",
						children: [/* @__PURE__ */ (0, W.jsx)("div", {
							className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
							children: e.label
						}), /* @__PURE__ */ (0, W.jsx)("button", {
							type: "button",
							className: "text-[10px] text-primary disabled:text-muted-foreground",
							onClick: () => l({}),
							disabled: Object.keys(c).length === 0,
							children: "Clear all"
						})]
					}), /* @__PURE__ */ (0, W.jsxs)("div", {
						className: "max-h-72 space-y-0.5 overflow-auto",
						children: [g.map((e) => {
							let t = h[e.groupKey] ?? 0, n = e.groupKey === o;
							return /* @__PURE__ */ (0, W.jsxs)("div", {
								role: "button",
								tabIndex: 0,
								onMouseEnter: () => s(e.groupKey),
								onFocus: () => s(e.groupKey),
								onClick: () => s(e.groupKey),
								onKeyDown: (t) => {
									(t.key === "Enter" || t.key === " " || t.key === "ArrowRight") && (t.preventDefault(), s(e.groupKey));
								},
								className: i("flex cursor-pointer items-center justify-between gap-2 rounded-md px-1.5 py-0.5 text-sm", "hover:bg-accent/50 focus-visible:bg-accent/50 focus-visible:outline-none", n && "bg-accent/60"),
								children: [
									/* @__PURE__ */ (0, W.jsx)("span", {
										className: "min-w-0 flex-1 truncate",
										children: e.label ?? e.groupKey
									}),
									t > 0 && /* @__PURE__ */ (0, W.jsxs)("span", {
										className: "rounded-full bg-primary/15 px-1.5 text-[10px] font-medium text-primary",
										children: [
											t,
											"/",
											e.options.length
										]
									}),
									/* @__PURE__ */ (0, W.jsx)(p, {
										icon: j,
										className: "shrink-0 text-muted-foreground"
									})
								]
							}, e.groupKey);
						}), f.length === 0 && /* @__PURE__ */ (0, W.jsx)("div", {
							className: "px-2 py-3 text-sm text-muted-foreground",
							children: "No groups"
						})]
					})]
				}), m && /* @__PURE__ */ (0, W.jsxs)("div", {
					onMouseEnter: () => s(m.groupKey),
					className: "ml-1.5 min-w-[16rem] max-w-[20rem] rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-lg shadow-black/5",
					children: [/* @__PURE__ */ (0, W.jsxs)("div", {
						className: "mb-2 flex items-center justify-between gap-2",
						children: [/* @__PURE__ */ (0, W.jsx)("div", {
							className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
							children: m.label ?? m.groupKey
						}), /* @__PURE__ */ (0, W.jsx)("button", {
							type: "button",
							className: "text-[10px] text-primary disabled:text-muted-foreground",
							onClick: () => _(m.groupKey),
							disabled: (h[m.groupKey] ?? 0) === 0,
							children: "Clear"
						})]
					}), /* @__PURE__ */ (0, W.jsxs)("div", {
						className: "max-h-72 space-y-0.5 overflow-auto",
						children: [m.options.map((e) => {
							let t = c[e.value] ?? "neutral", n = e.title ?? ke(e);
							return /* @__PURE__ */ (0, W.jsx)("div", {
								role: "button",
								tabIndex: 0,
								"data-filter-option": e.value,
								className: "rounded-md px-1.5 py-0.5 hover:bg-accent/50 focus-visible:bg-accent/50 focus-visible:outline-none",
								onClick: () => l(Ae(c, e.value, Se(t))),
								onKeyDown: (n) => {
									n.key === "Enter" || n.key === " " ? (n.preventDefault(), l(Ae(c, e.value, Se(t)))) : n.key === "ArrowLeft" && (n.preventDefault(), s(null));
								},
								children: /* @__PURE__ */ (0, W.jsx)(Ne, {
									className: "w-full justify-between",
									label: mi(e, m.groupKey),
									mode: t,
									title: n,
									togglePosition: "right",
									onModeChange: (t) => l(Ae(c, e.value, t))
								})
							}, e.value);
						}), m.options.length === 0 && /* @__PURE__ */ (0, W.jsx)("div", {
							className: "px-2 py-3 text-sm text-muted-foreground",
							children: "No values"
						})]
					})]
				})]
			})
		}) })]
	});
}
function mi(e, t) {
	if (typeof e.label == "string") {
		let n = `${t}=`;
		return e.label.startsWith(n) ? e.label.slice(n.length) : e.label;
	}
	return e.label ?? e.value;
}
function hi({ filter: e, grow: t }) {
	return /* @__PURE__ */ (0, W.jsxs)("label", {
		title: e.description,
		className: i("flex h-8 items-center gap-2 rounded-md border border-input bg-muted/30 pl-2 pr-1 text-xs", t ? "min-w-[12rem] max-w-[18rem] flex-1" : "min-w-[11rem] max-w-[15rem] shrink-0", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, W.jsx)("span", {
			className: "whitespace-nowrap font-medium uppercase tracking-wide text-muted-foreground",
			children: e.label
		}), /* @__PURE__ */ (0, W.jsx)(mn, {
			options: e.options,
			value: e.value,
			onChange: e.onChange,
			ariaLabel: `${e.label} filter`,
			...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
			...e.disabled === void 0 ? {} : { disabled: e.disabled },
			triggerClassName: "h-6 min-w-0 border-0 bg-transparent px-1 text-xs shadow-none focus-visible:ring-0",
			menuClassName: "left-auto right-0"
		})]
	});
}
function gi({ kind: e, label: t, from: n = "", to: r = "", onApply: i, presets: a, timeEnabled: o, timeZone: s, timeZones: c, fromPlaceholder: l, toPlaceholder: u, emptyLabel: d, className: f, disabled: p }) {
	let { compact: m } = (0, U.useContext)(B);
	return /* @__PURE__ */ (0, W.jsx)(Dn, {
		kind: e,
		label: t,
		from: n,
		to: r,
		onApply: i,
		compact: m,
		...p === void 0 ? {} : { disabled: p },
		...a ? { presets: a } : {},
		...o === void 0 ? {} : { timeEnabled: o },
		...s ? { timeZone: s } : {},
		...c ? { timeZones: c } : {},
		...l ? { fromPlaceholder: l } : {},
		...u ? { toPlaceholder: u } : {},
		...d ? { emptyLabel: d } : {},
		...f ? { className: f } : {}
	});
}
function _i(e, t, n = "bottom-start") {
	let r = _(), { refs: i, floatingStyles: a, context: o } = y({
		open: e,
		onOpenChange: t,
		placement: n,
		whileElementsMounted: tt,
		middleware: [
			et(6),
			nt({ padding: 8 }),
			rt({ padding: 8 })
		]
	}), { getReferenceProps: s, getFloatingProps: c } = ne([
		f(o),
		u(o, { escapeKey: !1 }),
		x(o, { role: "dialog" })
	]);
	return h(e, () => {
		t(!1), i.domReference.current instanceof HTMLElement && i.domReference.current.focus();
	}), {
		refs: i,
		floatingStyles: a,
		context: o,
		floatingZ: r,
		getReferenceProps: s,
		getFloatingProps: c
	};
}
function vi(e, t) {
	let { autoSubmit: n } = (0, U.useContext)(B), [r, i] = (0, U.useState)(e), a = (0, U.useRef)(t), o = (0, U.useRef)(e), s = (0, U.useRef)(!1);
	return (0, U.useEffect)(() => {
		a.current = t;
	}, [t]), (0, U.useEffect)(() => {
		o.current !== e && (o.current = e, s.current = !0, i(e));
	}, [e]), (0, U.useEffect)(() => {
		if (s.current) {
			s.current = !1;
			return;
		}
		if (r === e) return;
		if (!n || r === "") {
			a.current(r);
			return;
		}
		let t = window.setTimeout(() => {
			a.current(r);
		}, 500);
		return () => window.clearTimeout(t);
	}, [
		n,
		r,
		e
	]), [r, i];
}
function yi(e, t) {
	let { autoSubmit: n } = (0, U.useContext)(B), [r, i] = (0, U.useState)(e), a = (0, U.useRef)(t);
	return (0, U.useEffect)(() => {
		a.current = t;
	}, [t]), (0, U.useEffect)(() => {
		i(e);
	}, [
		e.max,
		e.maxOperator,
		e.maxUnit,
		e.min,
		e.minOperator,
		e.minUnit
	]), (0, U.useEffect)(() => {
		if (bi(r, e)) return;
		if (!n) {
			a.current(r);
			return;
		}
		let t = window.setTimeout(() => {
			a.current(r);
		}, 500);
		return () => window.clearTimeout(t);
	}, [
		n,
		r,
		e
	]), [r, i];
}
function bi(e, t) {
	let n = e, r = t;
	return (e.min ?? "") === (t.min ?? "") && (e.minOperator ?? ">=") === (t.minOperator ?? ">=") && (n.minUnit ?? "") === (r.minUnit ?? "") && (e.max ?? "") === (t.max ?? "") && (e.maxOperator ?? "<=") === (t.maxOperator ?? "<=") && (n.maxUnit ?? "") === (r.maxUnit ?? "");
}
function xi(e) {
	let t = Ti(e.value.min), n = Ti(e.value.max), r = t ?? 0, i = n ?? Math.max(r + 100, 100), a = e.domainMin ?? Math.min(r, i), o = e.domainMax ?? Math.max(i, a + 100), s = e.step && e.step > 0 ? e.step : 1;
	return {
		min: a,
		max: o <= a ? a + s : o,
		step: s
	};
}
function Si(e, t, n) {
	let r = Ti(n.min), i = Ti(n.max), a = String(n.min ?? "").trim() !== "", o = String(n.max ?? "").trim() !== "";
	if (!a && !o) return e.label;
	let s = Oi(r ?? t.min, e, n, "min"), c = Oi(i ?? t.max, e, n, "max");
	return a && o ? `${e.label} ${n.minOperator ?? ">="}${s}, ${n.maxOperator ?? "<="}${c}` : a ? `${e.label} ${n.minOperator ?? ">="}${s}` : `${e.label} ${n.maxOperator ?? "<="}${c}`;
}
function Ci(e, t) {
	let n = e.min ?? "", r = e.max ?? "", i = Ti(n), a = Ti(r);
	return i != null && a != null && i > a && (t.startsWith("min") ? r = n : n = r), {
		...e,
		min: n,
		max: r
	};
}
function wi(e, t, n) {
	let [r, i] = e, a = {
		min: r <= t.min ? "" : Ai(Ei(r, t.min, t.max)),
		minOperator: ">=",
		max: i >= t.max ? "" : Ai(Ei(i, t.min, t.max)),
		maxOperator: "<="
	};
	return n.kind === "duration" && (a.minUnit = n.unit ?? "ms", a.maxUnit = n.unit ?? "ms"), a;
}
function Ti(e) {
	if (!e?.trim()) return null;
	let t = Number(e);
	return Number.isFinite(t) ? t : null;
}
function Ei(e, t, n) {
	return Math.min(n, Math.max(t, e));
}
function Di(e, t) {
	return t.kind === "number" && t.formatValue ? t.formatValue(e) : t.kind === "duration" ? `${Ai(e)} ${t.unit ?? "ms"}` : t.unit ? `${Ai(e)} ${t.unit}` : Ai(e);
}
function Oi(e, t, n, r) {
	if (t.kind !== "duration") return Di(e, t);
	let i = n, a = r === "min" ? i.minUnit : i.maxUnit;
	return `${Ai(e)} ${a ?? t.unit ?? "ms"}`;
}
function ki(e, t) {
	e.onChange(t);
}
function Ai(e) {
	return String(Number.isInteger(e) ? e : Number(e.toFixed(6)));
}
function ji(e) {
	return Object.fromEntries(e.map((e) => [e.key, Ni(e)]));
}
function Mi(e) {
	return e.kind !== "date-range";
}
function Ni(e) {
	if (Mi(e)) return e.value;
}
function Pi(e, t, n) {
	return Mi(e) ? e.kind === "text" || e.kind === "lookup" || e.kind === "workload" ? {
		...e,
		value: String(t ?? ""),
		onChange: (e) => n(e)
	} : e.kind === "lookup-multi" ? {
		...e,
		value: Array.isArray(t) ? t.map(String) : [],
		onChange: (e) => n(e)
	} : e.kind === "multi" || e.kind === "tree" || e.kind === "nested-multi" ? {
		...e,
		value: zi(t) ? t : {},
		onChange: (e) => n(e)
	} : e.kind === "select-multi" ? {
		...e,
		value: Array.isArray(t) ? t.map(String) : [],
		onChange: (e) => n(e)
	} : e.kind === "number" || e.kind === "duration" ? {
		...e,
		value: Ri(t) ? t : {},
		onChange: (e) => n(e)
	} : e.kind === "enum" ? {
		...e,
		value: String(t ?? ""),
		onChange: (e) => n(e)
	} : e.kind === "tristate" ? {
		...e,
		value: t === void 0 ? void 0 : !!t,
		onChange: (e) => n(e)
	} : {
		...e,
		value: !!t,
		onChange: (e) => n(e)
	} : e;
}
function Fi(e, t) {
	for (let n of e) {
		if (!Mi(n)) continue;
		let e = Object.hasOwn(t, n.key) ? t[n.key] : Ni(n);
		Li(Ni(n), e) || Ii(n, e);
	}
}
function Ii(e, t) {
	Mi(e) && (e.kind === "text" || e.kind === "lookup" || e.kind === "workload" || e.kind === "enum" ? e.onChange(String(t ?? "")) : e.kind === "lookup-multi" || e.kind === "select-multi" ? e.onChange(Array.isArray(t) ? t.map(String) : []) : e.kind === "number" || e.kind === "duration" ? ki(e, Ri(t) ? t : {}) : e.kind === "boolean" ? e.onChange(!!t) : e.kind === "tristate" ? e.onChange(t === void 0 ? void 0 : !!t) : e.onChange(zi(t) ? t : {}));
}
function Li(e, t) {
	return JSON.stringify(e) === JSON.stringify(t);
}
function Ri(e) {
	return typeof e == "object" && !!e && !Array.isArray(e);
}
function zi(e) {
	return typeof e == "object" && !!e && !Array.isArray(e);
}
function Bi(e, t) {
	if (e.length === 0) return 0;
	let n = 0, r = 0;
	for (let i of e) {
		let e = n + (r > 0 ? Or : 0) + i;
		if (e > t) break;
		n = e, r += 1;
	}
	return r;
}
function Vi(e) {
	return e.length === 0 ? 0 : e.reduce((e, t) => e + t, 0) + (e.length - 1) * Or;
}
function Hi(e) {
	return e.kind === "boolean" ? Math.max(88, e.label.length * 8 + 40) : e.kind === "multi" || e.kind === "tree" ? 176 : e.kind === "nested-multi" ? 136 : e.kind === "workload" ? 224 : e.kind === "number" || e.kind === "duration" ? 152 : Math.max(144, e.label.length * 8 + 96);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/use-infinite-rows.ts
function Ui(e, t) {
	let n = (0, U.useRef)(e);
	n.current = e;
	let r = (0, U.useRef)(!1), i = e?.loading ?? !1;
	return (0, U.useEffect)(() => {
		r.current = !1;
	}, [i, t]), (0, U.useCallback)(() => {
		let e = n.current;
		e && !e.loading && e.hasMore && !r.current && (r.current = !0, e.onLoadMore());
	}, []);
}
// ../../../../../../pnpm/store/v11/links/@tanstack/virtual-core/3.17.8/7c4c8a2dd3f012a209ac831b3d885a8aa02d549fe4515126bf4c5e475ccaaec2/node_modules/@tanstack/virtual-core/dist/esm/lazy-measurements.js
function Wi(e, t, n) {
	let r = Array(e);
	return new Proxy(r, { get(r, i, a) {
		if (typeof i == "string") {
			let a = i.charCodeAt(0);
			if (a >= 48 && a <= 57) {
				let a = +i;
				if (Number.isInteger(a) && a >= 0 && a < e) {
					let e = r[a];
					if (!e) {
						let i = t[a * 2];
						e = r[a] = {
							index: a,
							key: n(a),
							start: i,
							size: t[a * 2 + 1],
							end: i + t[a * 2 + 1],
							lane: 0
						};
					}
					return e;
				}
			}
			if (i === "length") return e;
		}
		return Reflect.get(r, i, a);
	} });
}
// ../../../../../../pnpm/store/v11/links/@tanstack/virtual-core/3.17.8/7c4c8a2dd3f012a209ac831b3d885a8aa02d549fe4515126bf4c5e475ccaaec2/node_modules/@tanstack/virtual-core/dist/esm/utils.js
function Gi(e, t, n) {
	let r = n.initialDeps ?? [], i, a = !0;
	function o() {
		let o = e();
		return (o.length !== r.length || o.some((e, t) => r[t] !== e)) && (r = o, i = t(...o), n?.onChange && !(a && n.skipInitialOnChange) && n.onChange(i), a = !1), i;
	}
	return o.updateDeps = (e) => {
		r = e;
	}, o;
}
function Ki(e, t) {
	if (e === void 0) throw Error(`Unexpected undefined${t ? `: ${t}` : ""}`);
	return e;
}
var qi = (e, t) => Math.abs(e - t) < 1.01, Ji = (e, t, n) => {
	let r;
	return Object.assign(function(...i) {
		e.clearTimeout(r), r = e.setTimeout(() => t.apply(this, i), n);
	}, { cancel: () => {
		e.clearTimeout(r);
	} });
}, Yi, Xi = () => {
	if (Yi !== void 0) return Yi;
	if (typeof navigator > "u") return Yi = !1;
	if (/iP(hone|od|ad)/.test(navigator.userAgent)) return Yi = !0;
	let e = navigator.maxTouchPoints;
	return Yi = navigator.platform === "MacIntel" && e !== void 0 && e > 0;
}, Zi = (e) => {
	let { offsetWidth: t, offsetHeight: n } = e;
	return {
		width: t,
		height: n
	};
}, Qi = (e) => e, $i = (e) => {
	let t = Math.max(e.startIndex - e.overscan, 0), n = Math.min(e.endIndex + e.overscan, e.count - 1) - t + 1, r = Array(n);
	for (let e = 0; e < n; e++) r[e] = t + e;
	return r;
}, ea = (e, t) => {
	let n = e.scrollElement;
	if (!n) return;
	let r = e.targetWindow;
	if (!r) return;
	let i = (e) => {
		let { width: n, height: r } = e;
		t({
			width: Math.round(n),
			height: Math.round(r)
		});
	};
	if (i(Zi(n)), !r.ResizeObserver) return () => {};
	let a = new r.ResizeObserver((t) => {
		let r = () => {
			let e = t[0];
			if (e?.borderBoxSize) {
				let t = e.borderBoxSize[0];
				if (t) {
					i({
						width: t.inlineSize,
						height: t.blockSize
					});
					return;
				}
			}
			i(Zi(n));
		};
		e.options.useAnimationFrameWithResizeObserver ? requestAnimationFrame(r) : r();
	});
	return a.observe(n, { box: "border-box" }), () => {
		a.unobserve(n);
	};
}, ta = { passive: !0 }, na = typeof window > "u" || "onscrollend" in window, ra = (e, t, n) => {
	let r = e.scrollElement;
	if (!r) return;
	let i = e.targetWindow;
	if (!i) return;
	let a = e.options.useScrollendEvent && na, o = 0, s = a ? null : Ji(i, () => t(o, !1), e.options.isScrollingResetDelay), c = (e) => () => {
		o = n(r), s?.(), t(o, e);
	}, l = c(!0), u = c(!1);
	return r.addEventListener("scroll", l, ta), a && r.addEventListener("scrollend", u, ta), () => {
		r.removeEventListener("scroll", l), a && r.removeEventListener("scrollend", u), s?.cancel();
	};
}, ia = (e, t) => ra(e, t, (t) => {
	let { horizontal: n, isRtl: r } = e.options;
	return n ? t.scrollLeft * (r && -1 || 1) : t.scrollTop;
}), aa = (e, t, n) => {
	if (n.options.useCachedMeasurements) {
		let t = n.indexFromElement(e), r = n.options.getItemKey(t);
		return n.itemSizeCache.get(r) ?? n.options.estimateSize(t);
	}
	if (t?.borderBoxSize) {
		let e = t.borderBoxSize[0];
		if (e) return Math.round(e[n.options.horizontal ? "inlineSize" : "blockSize"]);
	}
	if (!t) {
		let t = n.indexFromElement(e), r = n.options.getItemKey(t), i = n.itemSizeCache.get(r);
		if (i !== void 0) return i;
	}
	return e[n.options.horizontal ? "offsetWidth" : "offsetHeight"];
}, oa = (e, { adjustments: t = 0, behavior: n }, r) => {
	var i, a;
	(a = (i = r.scrollElement)?.scrollTo) == null || a.call(i, {
		[r.options.horizontal ? "left" : "top"]: e + t,
		behavior: n
	});
}, sa = class {
	constructor(e) {
		this.unsubs = [], this.scrollElement = null, this.targetWindow = null, this.isScrolling = !1, this.scrollState = null, this.measurementsCache = [], this._flatMeasurements = null, this.itemSizeCache = /* @__PURE__ */ new Map(), this.itemSizeCacheVersion = 0, this.laneAssignments = /* @__PURE__ */ new Map(), this.pendingMin = null, this.prevLanes = void 0, this.lanesChangedFlag = !1, this.lanesSettling = !1, this.pendingScrollAnchor = null, this.scrollRect = null, this.scrollOffset = null, this.scrollDirection = null, this.scrollAdjustments = 0, this._iosDeferredAdjustment = 0, this._iosTouching = !1, this._iosJustTouchEnded = !1, this._iosTouchEndTimerId = null, this._intendedScrollOffset = null, this.elementsCache = /* @__PURE__ */ new Map(), this.now = () => {
			var e;
			return ((e = this.targetWindow?.performance)?.now)?.call(e) ?? Date.now();
		}, this.observer = /* @__PURE__ */ (() => {
			let e = null, t = () => e || (!this.targetWindow || !this.targetWindow.ResizeObserver ? null : e = new this.targetWindow.ResizeObserver((e) => {
				e.forEach((e) => {
					let t = () => {
						let t = e.target, n = this.indexFromElement(t);
						if (t.isConnected) this.isIndexInRange(n) && this.shouldMeasureDuringScroll(n) && this.resizeItem(n, this.options.measureElement(t, e, this));
						else {
							this.observer.unobserve(t);
							for (let [e, n] of this.elementsCache) if (n === t) {
								this.elementsCache.delete(e);
								break;
							}
						}
					};
					this.options.useAnimationFrameWithResizeObserver ? requestAnimationFrame(t) : t();
				});
			}));
			return {
				disconnect: () => {
					var n;
					(n = t()) == null || n.disconnect(), e = null;
				},
				observe: (e) => t()?.observe(e, { box: "border-box" }),
				unobserve: (e) => t()?.unobserve(e)
			};
		})(), this.range = null, this.setOptions = (e) => {
			let t = {
				debug: !1,
				initialOffset: 0,
				overscan: 1,
				paddingStart: 0,
				paddingEnd: 0,
				scrollPaddingStart: 0,
				scrollPaddingEnd: 0,
				horizontal: !1,
				getItemKey: Qi,
				rangeExtractor: $i,
				onChange: () => {},
				measureElement: aa,
				initialRect: {
					width: 0,
					height: 0
				},
				scrollMargin: 0,
				gap: 0,
				indexAttribute: "data-index",
				initialMeasurementsCache: [],
				lanes: 1,
				anchorTo: "start",
				followOnAppend: !1,
				scrollEndThreshold: 1,
				isScrollingResetDelay: 150,
				enabled: !0,
				isRtl: !1,
				useScrollendEvent: !1,
				useAnimationFrameWithResizeObserver: !1,
				laneAssignmentMode: "estimate",
				useCachedMeasurements: !1
			};
			for (let n in e) {
				let r = e[n];
				r !== void 0 && (t[n] = r);
			}
			let n = this.options, r = null, i = null, a = !1;
			if (n !== void 0 && n.enabled && t.enabled && t.anchorTo === "end" && this.scrollElement !== null) {
				let e = n.count, o = t.count, s = this.getMeasurements(), c = e > 0 ? s[0]?.key ?? n.getItemKey(0) : null, l = e > 0 ? s[e - 1]?.key ?? n.getItemKey(e - 1) : null;
				if (o !== e || e > 0 && o > 0 && (t.getItemKey(0) !== c || t.getItemKey(o - 1) !== l)) {
					a = !0;
					let c = e > 0 ? this.getVirtualItemForOffset(this.getScrollOffset()) ?? s[0] : null;
					c && (r = [c.key, this.getScrollOffset() - c.start]);
					let u = t.followOnAppend === !0 ? "auto" : t.followOnAppend || null;
					u && o > e && this.isAtEnd(n.scrollEndThreshold) && (e === 0 || t.getItemKey(o - 1) !== l) && (i = u);
				}
			}
			this.options = t, a && (this.pendingMin = 0, this.itemSizeCacheVersion++);
			let o = !1, s = 0;
			if (r && this.scrollOffset !== null) {
				let [e, t] = r, n = this.getMeasurements(), { count: i, getItemKey: a } = this.options, c = 0;
				for (; c < i && a(c) !== e;) c++;
				if (c < i) {
					let e = n[c];
					if (e) {
						let n = Math.max(0, e.start + t);
						n !== this.scrollOffset && (s = n - this.scrollOffset, this.scrollOffset = n, o = !0);
					}
				}
			}
			(o || i) && (this.pendingScrollAnchor = [
				o ? r[0] : null,
				o ? r[1] : 0,
				i,
				s
			]);
		}, this.notify = (e) => {
			var t, n;
			(n = (t = this.options).onChange) == null || n.call(t, this, e);
		}, this.maybeNotify = Gi(() => (this.calculateRange(), [
			this.isScrolling,
			this.range ? this.range.startIndex : null,
			this.range ? this.range.endIndex : null
		]), (e) => {
			this.notify(e);
		}, {
			key: !1,
			debug: () => this.options.debug,
			initialDeps: [
				this.isScrolling,
				this.range ? this.range.startIndex : null,
				this.range ? this.range.endIndex : null
			]
		}), this.cleanup = () => {
			this.unsubs.filter(Boolean).forEach((e) => e()), this.unsubs = [], this.observer.disconnect(), this.rafId != null && this.targetWindow && (this.targetWindow.cancelAnimationFrame(this.rafId), this.rafId = null), this.scrollState = null, this.isScrolling = !1, this.scrollDirection = null, this._iosDeferredAdjustment = 0, this._iosTouching = !1, this._iosJustTouchEnded = !1, this.scrollElement = null, this.targetWindow = null;
		}, this._didMount = () => () => {
			this.cleanup();
		}, this._willUpdate = () => {
			let e = this.options.enabled ? this.options.getScrollElement() : null;
			if (this.scrollElement !== e) {
				if (this.cleanup(), !e) {
					this.maybeNotify();
					return;
				}
				if (this.scrollElement = e, this.targetWindow = this.scrollElement && "ownerDocument" in this.scrollElement ? this.scrollElement.ownerDocument.defaultView : this.scrollElement?.window ?? null, this.elementsCache.forEach((e) => {
					this.observer.observe(e);
				}), this.unsubs.push(this.options.observeElementRect(this, (e) => {
					this.scrollRect = e, this.maybeNotify();
				})), this.unsubs.push(this.options.observeElementOffset(this, (e, t) => {
					if (t && this._intendedScrollOffset === null && e === this.scrollOffset) return;
					this._intendedScrollOffset !== null && Math.abs(e - this._intendedScrollOffset) < 1.5 && (e = this._intendedScrollOffset), this._intendedScrollOffset = null, this.scrollAdjustments = 0;
					let n = this.getScrollOffset();
					this.scrollDirection = t ? n === e ? this.scrollDirection : n < e ? "forward" : "backward" : null, this.scrollOffset = e, this.isScrolling = t, this._flushIosDeferredIfReady(), this.scrollState && this.scheduleScrollReconcile(), this.maybeNotify();
				})), "addEventListener" in this.scrollElement) {
					let e = this.scrollElement, t = () => {
						this._iosTouching = !0, this._iosJustTouchEnded = !1, this._iosTouchEndTimerId !== null && this.targetWindow != null && (this.targetWindow.clearTimeout(this._iosTouchEndTimerId), this._iosTouchEndTimerId = null);
					}, n = () => {
						this._iosTouching = !1, Xi() && this.targetWindow != null && (this._iosJustTouchEnded = !0, this._iosTouchEndTimerId = this.targetWindow.setTimeout(() => {
							this._iosJustTouchEnded = !1, this._iosTouchEndTimerId = null, this._flushIosDeferredIfReady();
						}, 150));
					};
					e.addEventListener("touchstart", t, ta), e.addEventListener("touchend", n, ta), this.unsubs.push(() => {
						e.removeEventListener("touchstart", t), e.removeEventListener("touchend", n), this._iosTouchEndTimerId !== null && this.targetWindow != null && (this.targetWindow.clearTimeout(this._iosTouchEndTimerId), this._iosTouchEndTimerId = null);
					});
				}
				this._scrollToOffset(this.getScrollOffset(), {
					adjustments: void 0,
					behavior: void 0
				});
			}
			let t = this.pendingScrollAnchor;
			if (this.pendingScrollAnchor = null, t && this.scrollElement && this.options.enabled) {
				let [e, n, r, i] = t;
				e !== null && !r && (Xi() && (this.isScrolling || this._iosTouching || this._iosJustTouchEnded) ? i !== 0 && (this._iosDeferredAdjustment += i) : this._scrollToOffset(this.getScrollOffset(), {
					adjustments: void 0,
					behavior: void 0
				})), r && this.scrollToEnd({ behavior: r });
			}
		}, this._flushIosDeferredIfReady = () => {
			if (this._iosDeferredAdjustment === 0 || this.isScrolling || this._iosTouching || this._iosJustTouchEnded) return;
			let e = this.getScrollOffset(), t = this.getMaxScrollOffset();
			if (e < 0 || e > t) return;
			if (this._iosDeferredAdjustment < 0 && e >= t - 1) {
				this._iosDeferredAdjustment = 0;
				return;
			}
			let n = this._iosDeferredAdjustment;
			this._iosDeferredAdjustment = 0, this._scrollToOffset(e, {
				adjustments: this.scrollAdjustments += n,
				behavior: void 0
			});
		}, this.rafId = null, this.getSize = () => this.options.enabled ? (this.scrollRect = this.scrollRect ?? this.options.initialRect, this.scrollRect[this.options.horizontal ? "width" : "height"]) : (this.scrollRect = null, 0), this.getScrollOffset = () => this.options.enabled ? (this.scrollOffset = this.scrollOffset ?? (typeof this.options.initialOffset == "function" ? this.options.initialOffset() : this.options.initialOffset), this.scrollOffset) : (this.scrollOffset = null, 0), this.getMeasurementOptions = Gi(() => [
			this.options.count,
			this.options.paddingStart,
			this.options.scrollMargin,
			this.options.getItemKey,
			this.options.enabled,
			this.options.lanes,
			this.options.laneAssignmentMode,
			this.options.gap
		], (e, t, n, r, i, a, o, s) => (this.prevLanes !== void 0 && this.prevLanes !== a && (this.lanesChangedFlag = !0), this.prevLanes = a, this.pendingMin = null, {
			count: e,
			paddingStart: t,
			scrollMargin: n,
			getItemKey: r,
			enabled: i,
			lanes: a,
			laneAssignmentMode: o,
			gap: s
		}), { key: !1 }), this.isIndexInRange = (e) => e >= 0 && e < this.options.count, this.getMeasurements = Gi(() => [this.getMeasurementOptions(), this.itemSizeCacheVersion], ({ count: e, paddingStart: t, scrollMargin: n, getItemKey: r, enabled: i, lanes: a, laneAssignmentMode: o, gap: s }, c) => {
			let l = this.itemSizeCache;
			if (!i) return this.measurementsCache = [], this.itemSizeCache.clear(), this.laneAssignments.clear(), [];
			if (this.laneAssignments.size > e) for (let t of this.laneAssignments.keys()) t >= e && this.laneAssignments.delete(t);
			this.lanesChangedFlag && (this.lanesChangedFlag = !1, this.lanesSettling = !0, this.measurementsCache = [], this.itemSizeCache.clear(), this.laneAssignments.clear(), this.pendingMin = null), this.measurementsCache.length === 0 && !this.lanesSettling && (this.measurementsCache = this.options.initialMeasurementsCache, this.measurementsCache.forEach((e) => {
				this.itemSizeCache.set(e.key, e.size);
			}));
			let u = this.lanesSettling ? 0 : this.pendingMin ?? 0;
			if (this.pendingMin = null, this.lanesSettling && this.measurementsCache.length === e && (this.lanesSettling = !1), a === 1) {
				let i = e * 2, a = this._flatMeasurements;
				if (!a || a.length < i) {
					let e = new Float64Array(i);
					a && u > 0 && e.set(a.subarray(0, u * 2)), a = e, this._flatMeasurements = a;
				}
				let o;
				if (u === 0) o = t + n;
				else {
					let e = u - 1;
					o = a[e * 2] + a[e * 2 + 1] + s;
				}
				for (let t = u; t < e; t++) {
					let e = r(t), n = l.get(e), i = typeof n == "number" ? n : this.options.estimateSize(t);
					a[t * 2] = o, a[t * 2 + 1] = i, o += i + s;
				}
				let c = Wi(e, a, r);
				return this.measurementsCache = c, c;
			}
			let d = this.measurementsCache.slice(0, u), f = Array(a).fill(void 0), p = new Float64Array(a), m = 0;
			for (let e = 0; e < u; e++) {
				let t = d[e];
				t && (f[t.lane] === void 0 && m++, f[t.lane] = e, p[t.lane] = t.end);
			}
			for (let i = u; i < e; i++) {
				let e = r(i), c = this.laneAssignments.get(i), u, h, g = o === "estimate" || l.has(e);
				if (c !== void 0 && this.options.lanes > 1) {
					u = c;
					let e = f[u], r = e === void 0 ? void 0 : d[e];
					h = r ? r.end + s : t + n;
				} else if (m === a) {
					let e = 0, t = p[0], n = f[0];
					for (let r = 1; r < a; r++) {
						let i = p[r];
						(i < t || i === t && f[r] < n) && (e = r, t = i, n = f[r]);
					}
					u = e, h = t + s, g && this.laneAssignments.set(i, u);
				} else u = i % this.options.lanes, h = t + n, g && this.laneAssignments.set(i, u);
				let _ = l.get(e), v = typeof _ == "number" ? _ : this.options.estimateSize(i), y = h + v;
				d[i] = {
					index: i,
					start: h,
					size: v,
					end: y,
					key: e,
					lane: u
				}, f[u] === void 0 && m++, f[u] = i, p[u] = y;
			}
			return this.measurementsCache = d, d;
		}, {
			key: !1,
			debug: () => this.options.debug
		}), this.calculateRange = Gi(() => [
			this.getMeasurements(),
			this.getSize(),
			this.getScrollOffset(),
			this.options.lanes
		], (e, t, n, r) => e.length === 0 || t === 0 ? (this.range = null, null) : (this.range = ua(e, t, n, r, r === 1 && this._flatMeasurements != null ? this._flatMeasurements : null), this.range), {
			key: !1,
			debug: () => this.options.debug
		}), this.getVirtualIndexes = Gi(() => {
			let e = null, t = null, n = this.calculateRange();
			return n && (e = n.startIndex, t = n.endIndex), this.maybeNotify.updateDeps([
				this.isScrolling,
				e,
				t
			]), [
				this.options.rangeExtractor,
				this.options.overscan,
				this.options.count,
				e,
				t
			];
		}, (e, t, n, r, i) => r === null || i === null ? [] : e({
			startIndex: r,
			endIndex: i,
			overscan: t,
			count: n
		}), {
			key: !1,
			debug: () => this.options.debug
		}), this.indexFromElement = (e) => {
			let t = this.options.indexAttribute, n = e.getAttribute(t);
			return n ? parseInt(n, 10) : (console.warn(`Missing attribute name '${t}={index}' on measured element.`), -1);
		}, this.shouldMeasureDuringScroll = (e) => {
			if (!this.scrollState || this.scrollState.behavior !== "smooth") return !0;
			let t = this.scrollState.index ?? this.getVirtualItemForOffset(this.scrollState.lastTargetOffset)?.index;
			if (t !== void 0 && this.range) {
				let n = Math.max(this.options.overscan, Math.ceil((this.range.endIndex - this.range.startIndex) / 2)), r = Math.max(0, t - n), i = Math.min(this.options.count - 1, t + n);
				return e >= r && e <= i;
			}
			return !0;
		}, this.measureElement = (e) => {
			if (!e) {
				this.elementsCache.forEach((e, t) => {
					e.isConnected || (this.observer.unobserve(e), this.elementsCache.delete(t));
				});
				return;
			}
			let t = this.indexFromElement(e);
			if (!this.isIndexInRange(t)) return;
			let n = this.options.getItemKey(t), r = this.elementsCache.get(n);
			r !== e && (r && this.observer.unobserve(r), this.observer.observe(e), this.elementsCache.set(n, e)), (!this.isScrolling || this.scrollState) && this.shouldMeasureDuringScroll(t) && this.resizeItem(t, this.options.measureElement(e, void 0, this));
		}, this.resizeItem = (e, t) => {
			if (!this.isIndexInRange(e)) return;
			let n, r, i, a = this._flatMeasurements;
			if (this.options.lanes === 1 && a !== null) i = this.options.getItemKey(e), r = a[e * 2], n = a[e * 2 + 1];
			else {
				let t = this.measurementsCache[e];
				if (!t) return;
				i = t.key, r = t.start, n = t.size;
			}
			let o = this.itemSizeCache.get(i) ?? n, s = t - o;
			if (s !== 0) {
				let a = this.options.anchorTo === "end" && this.scrollState?.behavior !== "smooth" && this.getVirtualDistanceFromEnd() <= this.options.scrollEndThreshold, c = a ? this.getTotalSize() : 0, l = this.getScrollOffset() + this.scrollAdjustments, u = this.itemSizeCache.has(i) ? r + o <= l && this.scrollDirection !== "backward" : r < l, d = this.scrollState?.behavior !== "smooth" && (this.shouldAdjustScrollPositionOnItemSizeChange === void 0 ? u : this.shouldAdjustScrollPositionOnItemSizeChange(this.measurementsCache[e] ?? {
					index: e,
					key: i,
					start: r,
					size: n,
					end: r + n,
					lane: 0
				}, s, this));
				(this.pendingMin === null || e < this.pendingMin) && (this.pendingMin = e), this.itemSizeCache.set(i, t), this.itemSizeCacheVersion++;
				let f = !1;
				a ? f = this.applyScrollAdjustment(this.getTotalSize() - c) : d && (f = this.applyScrollAdjustment(s)), this.notify(f);
			}
		}, this.getVirtualItems = Gi(() => [this.getVirtualIndexes(), this.getMeasurements()], (e, t) => {
			let n = [];
			for (let r = 0, i = e.length; r < i; r++) {
				let i = t[e[r]];
				n.push(i);
			}
			return n;
		}, {
			key: !1,
			debug: () => this.options.debug
		}), this.getVirtualItemForOffset = (e) => {
			let t = this.getMeasurements();
			if (t.length === 0) return;
			let n = this._flatMeasurements, r = this.options.lanes === 1 && n != null;
			return Ki(t[ca(0, t.length - 1, r ? (e) => n[e * 2] : (e) => Ki(t[e]).start, e)]);
		}, this.getMaxScrollOffset = () => {
			if (!this.scrollElement) return 0;
			if ("scrollHeight" in this.scrollElement) return this.options.horizontal ? this.scrollElement.scrollWidth - this.scrollElement.clientWidth : this.scrollElement.scrollHeight - this.scrollElement.clientHeight;
			{
				let e = this.scrollElement.document.documentElement;
				return this.options.horizontal ? e.scrollWidth - this.scrollElement.innerWidth : e.scrollHeight - this.scrollElement.innerHeight;
			}
		}, this.getVirtualDistanceFromEnd = () => Math.max(this.getTotalSize() - this.getSize() - this.getScrollOffset(), 0), this.getDistanceFromEnd = () => Math.max(this.getMaxScrollOffset() - this.getScrollOffset(), 0), this.isAtEnd = (e = this.options.scrollEndThreshold) => this.getDistanceFromEnd() <= e, this.getOffsetForAlignment = (e, t, n = 0) => {
			if (!this.scrollElement) return 0;
			let r = this.getSize(), i = this.getScrollOffset();
			t === "auto" && (t = e >= i + r ? "end" : "start"), t === "center" ? e += (n - r) / 2 : t === "end" && (e -= r);
			let a = this.getMaxScrollOffset();
			return Math.max(Math.min(a, e), 0);
		}, this.getOffsetForIndex = (e, t = "auto") => {
			e = Math.max(0, Math.min(e, this.options.count - 1));
			let n = this.getSize(), r = this.getScrollOffset(), i = this.measurementsCache[e];
			if (!i) return;
			if (t === "auto") {
				if (i.end >= r + n - this.options.scrollPaddingEnd) t = "end";
				else if (i.start <= r + this.options.scrollPaddingStart) t = "start";
				else return [r, t];
			}
			if (t === "end" && e === this.options.count - 1) return [this.getMaxScrollOffset(), t];
			let a = t === "end" ? i.end + this.options.scrollPaddingEnd : i.start - this.options.scrollPaddingStart;
			return [this.getOffsetForAlignment(a, t, i.size), t];
		}, this.scrollToOffset = (e, { align: t = "start", behavior: n = "auto" } = {}) => {
			this._iosDeferredAdjustment = 0;
			let r = this.getOffsetForAlignment(e, t), i = this.now();
			this.scrollState = {
				index: null,
				align: t,
				behavior: n,
				startedAt: i,
				lastTargetOffset: r,
				stableFrames: 0
			}, this._scrollToOffset(r, {
				adjustments: void 0,
				behavior: n
			}), this.scheduleScrollReconcile();
		}, this.scrollToIndex = (e, { align: t = "auto", behavior: n = "auto" } = {}) => {
			this._iosDeferredAdjustment = 0, e = Math.max(0, Math.min(e, this.options.count - 1));
			let r = this.getOffsetForIndex(e, t);
			if (!r) return;
			let [i, a] = r, o = this.now();
			this.scrollState = {
				index: e,
				align: a,
				behavior: n,
				startedAt: o,
				lastTargetOffset: i,
				stableFrames: 0
			}, this._scrollToOffset(i, {
				adjustments: void 0,
				behavior: n
			}), this.scheduleScrollReconcile();
		}, this.scrollBy = (e, { behavior: t = "auto" } = {}) => {
			let n = this.getScrollOffset() + e, r = this.now();
			this.scrollState = {
				index: null,
				align: "start",
				behavior: t,
				startedAt: r,
				lastTargetOffset: n,
				stableFrames: 0
			}, this._scrollToOffset(n, {
				adjustments: void 0,
				behavior: t
			}), this.scheduleScrollReconcile();
		}, this.scrollToEnd = ({ behavior: e = "auto" } = {}) => {
			this.options.count > 0 ? this.scrollToIndex(this.options.count - 1, {
				align: "end",
				behavior: e
			}) : this.scrollToOffset(Math.max(this.getTotalSize() - this.getSize(), 0), { behavior: e });
		}, this.getTotalSize = () => {
			let e = this.getMeasurements(), t;
			if (e.length === 0) t = this.options.paddingStart;
			else if (this.options.lanes === 1) {
				let n = e.length - 1, r = this._flatMeasurements;
				t = r == null ? e[n]?.end ?? 0 : r[n * 2] + r[n * 2 + 1];
			} else {
				let n = Array(this.options.lanes).fill(null), r = e.length - 1;
				for (; r >= 0 && n.some((e) => e === null);) {
					let t = e[r];
					n[t.lane] === null && (n[t.lane] = t.end), r--;
				}
				t = Math.max(...n.filter((e) => e !== null));
			}
			return Math.max(t - this.options.scrollMargin + this.options.paddingEnd, 0);
		}, this.takeSnapshot = () => {
			let e = [];
			if (this.itemSizeCache.size === 0) return e;
			let t = this.getMeasurements();
			for (let n of t) n && this.itemSizeCache.has(n.key) && e.push({
				index: n.index,
				key: n.key,
				start: n.start,
				size: n.size,
				end: n.end,
				lane: n.lane
			});
			return e;
		}, this._scrollToOffset = (e, { adjustments: t, behavior: n }) => {
			this._intendedScrollOffset = e + (t ?? 0), this.options.scrollToFn(e, {
				behavior: n,
				adjustments: t
			}, this);
		}, this.measure = () => {
			this.pendingMin = null, this.itemSizeCache.clear(), this.laneAssignments.clear(), this.itemSizeCacheVersion++, this.notify(!1);
		}, this.setOptions(e);
	}
	applyScrollAdjustment(e, t) {
		return e === 0 ? !1 : Xi() && (this.isScrolling || this._iosTouching || this._iosJustTouchEnded) ? (this._iosDeferredAdjustment += e, !1) : (this._scrollToOffset(this.getScrollOffset(), {
			adjustments: this.scrollAdjustments += e,
			behavior: t
		}), this.scrollOffset !== null && (this.scrollOffset += this.scrollAdjustments, this.scrollOffset < 0 && (this.scrollOffset = 0), this.scrollAdjustments = 0), !0);
	}
	scheduleScrollReconcile() {
		this.targetWindow ? this.rafId ??= this.targetWindow.requestAnimationFrame(() => {
			this.rafId = null, this.reconcileScroll();
		}) : this.scrollState = null;
	}
	reconcileScroll() {
		if (!this.scrollState || !this.scrollElement) return;
		if (this.now() - this.scrollState.startedAt > 5e3) {
			this.scrollState = null;
			return;
		}
		let e = this.scrollState.index == null ? void 0 : this.getOffsetForIndex(this.scrollState.index, this.scrollState.align), t = e ? e[0] : this.scrollState.lastTargetOffset, n = t !== this.scrollState.lastTargetOffset;
		if (!n && qi(t, this.getScrollOffset())) {
			if (this.scrollState.stableFrames++, this.scrollState.stableFrames >= 1) {
				this.getScrollOffset() !== t && this._scrollToOffset(t, {
					adjustments: void 0,
					behavior: "auto"
				}), this.scrollState = null;
				return;
			}
		} else if (this.scrollState.stableFrames = 0, n) {
			let e = this.getSize() || 600, n = Math.abs(t - this.getScrollOffset()), r = this.scrollState.behavior === "smooth" && n > e;
			this.scrollState.lastTargetOffset = t, r || (this.scrollState.behavior = "auto"), this._scrollToOffset(t, {
				adjustments: void 0,
				behavior: r ? "smooth" : "auto"
			});
		}
		this.scheduleScrollReconcile();
	}
}, ca = (e, t, n, r) => {
	for (; e <= t;) {
		let i = (e + t) / 2 | 0, a = n(i);
		if (a < r) e = i + 1;
		else if (a > r) t = i - 1;
		else return i;
	}
	return e > 0 ? e - 1 : 0;
};
function la(e, t, n) {
	let r = 0;
	for (; r <= t;) {
		let i = (r + t) / 2 | 0, a = e[i * 2];
		if (a < n) r = i + 1;
		else if (a > n) t = i - 1;
		else return i;
	}
	return r > 0 ? r - 1 : 0;
}
function ua(e, t, n, r, i) {
	let a = e.length - 1;
	if (e.length <= r) return {
		startIndex: 0,
		endIndex: a
	};
	if (r === 1 && i !== null) {
		let e = la(i, a, n), r = e, o = n + t;
		for (; r < a && i[r * 2] + i[r * 2 + 1] < o;) r++;
		return {
			startIndex: e,
			endIndex: r
		};
	}
	let o = ca(0, a, (t) => e[t].start, n), s = o;
	if (r === 1) for (; s < a && e[s].end < n + t;) s++;
	else if (r > 1) {
		let i = Array(r).fill(0);
		for (; s < a && i.some((e) => e < n + t);) {
			let t = e[s];
			i[t.lane] = t.end, s++;
		}
		let c = Array(r).fill(n + t);
		for (; o >= 0 && c.some((e) => e >= n);) {
			let t = e[o];
			c[t.lane] = t.start, o--;
		}
		o = Math.max(0, o - o % r), s = Math.min(a, s + (r - 1 - s % r));
	}
	return {
		startIndex: o,
		endIndex: s
	};
}
// ../../../../../../pnpm/store/v11/links/@tanstack/react-virtual/3.14.10/c23a74998075603a2e1ad8151ff0cc247ee45246fe4ea1dc2a791ffedde063b8/node_modules/@tanstack/react-virtual/dist/esm/index.js
var da = typeof document < "u" ? U.useLayoutEffect : U.useEffect;
function fa({ useFlushSync: e = !0, directDomUpdates: t = !1, directDomUpdatesMode: n = "transform", ...r }) {
	let i = U.useReducer((e) => e + 1, 0)[1], a = U.useRef({
		enabled: t,
		mode: n,
		container: null,
		lastSize: null,
		lastPositions: /* @__PURE__ */ new WeakMap(),
		prevRange: null
	});
	a.current.enabled = t, a.current.mode = n;
	let o = (e) => {
		let t = a.current;
		if (!t.enabled || !t.container) return;
		let n = e.getTotalSize();
		if (n !== t.lastSize) {
			t.lastSize = n;
			let r = e.options.horizontal ? "width" : "height";
			t.container.style[r] = `${n}px`;
		}
	}, s = (e) => {
		let t = a.current;
		if (!t.enabled || !t.container) return;
		o(e);
		let n = !!e.options.horizontal, r = t.mode === "transform", i = n ? "left" : "top", s = e.options.scrollMargin, c = e.getVirtualItems();
		for (let a of c) {
			let o = a.start - s, c = e.elementsCache.get(a.key);
			c && t.lastPositions.get(c) !== o && (t.lastPositions.set(c, o), r ? c.style.transform = n ? `translate3d(${o}px, 0, 0)` : `translate3d(0, ${o}px, 0)` : c.style[i] = `${o}px`);
		}
	}, c = {
		...r,
		onChange: (t, n) => {
			var o;
			let c = a.current, l = !0;
			if (c.enabled) {
				s(t);
				let e = t.range, n = c.prevRange;
				l = !n || n.isScrolling !== t.isScrolling || n.startIndex !== e?.startIndex || n.endIndex !== e?.endIndex, l && (c.prevRange = e ? {
					startIndex: e.startIndex,
					endIndex: e.endIndex,
					isScrolling: t.isScrolling
				} : null);
			}
			l && (e && n ? (0, at.flushSync)(i) : i()), (o = r.onChange) == null || o.call(r, t, n);
		}
	}, [l] = U.useState(() => {
		let e = new sa(c);
		return Object.assign(e, { containerRef: (t) => {
			let n = a.current;
			if (n.container = t, n.lastSize = null, t && n.enabled) {
				let r = e.getTotalSize();
				n.lastSize = r;
				let i = e.options.horizontal ? "width" : "height";
				t.style[i] = `${r}px`;
			}
		} });
	});
	return l.setOptions(c), da(() => l._didMount(), []), da(() => (o(l), l._willUpdate())), da(() => {
		s(l);
	}), l;
}
function pa(e) {
	return fa({
		observeElementRect: ea,
		observeElementOffset: ia,
		scrollToFn: oa,
		...e
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/use-data-table-virtual.ts
var ma = 8, ha = 40;
function ga(e) {
	if (!e) return null;
	let t = typeof e == "object" ? e : {};
	return {
		overscan: t.overscan ?? ma,
		estimateRowHeight: t.estimateRowHeight ?? ha
	};
}
function _a(e, t, n) {
	let r = e[0], i = e[e.length - 1];
	return !r || !i ? {
		top: 0,
		bottom: 0
	} : {
		top: Math.max(0, r.start - n),
		bottom: Math.max(0, t - (i.end - n))
	};
}
function va({ count: e, options: t, getItemKey: n }) {
	let [r, i] = (0, U.useState)(null), [a, o] = (0, U.useState)(null), [s, c] = (0, U.useState)(0);
	(0, U.useEffect)(() => {
		if (!t || !a) return;
		let e = () => c(a.offsetTop);
		if (e(), typeof ResizeObserver > "u") return;
		let n = new ResizeObserver(e);
		return a.parentElement && n.observe(a.parentElement), () => n.disconnect();
	}, [t, a]);
	let l = (0, U.useCallback)(() => t?.estimateRowHeight ?? ha, [t?.estimateRowHeight]), u = pa({
		count: t ? e : 0,
		getScrollElement: () => r,
		estimateSize: l,
		overscan: t?.overscan ?? ma,
		getItemKey: n,
		scrollMargin: s
	}), d = t ? u.getVirtualItems() : [], f = (0, U.useMemo)(() => _a(d, u.getTotalSize(), s), [d, s]);
	return {
		enabled: !!t,
		setScrollEl: i,
		setTbodyEl: o,
		virtualizer: u,
		virtualItems: d,
		spacers: f
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/DataTableMenuSections.tsx
var ya = [
	{
		value: "compact",
		icon: le,
		label: "Compact"
	},
	{
		value: "comfortable",
		icon: K,
		label: "Comfortable"
	},
	{
		value: "spacious",
		icon: bt,
		label: "Spacious"
	}
];
function ba(e) {
	let t = [];
	for (let n of e) {
		let e = n.section ?? "Download", r = t.find((t) => t.section === e);
		r || (r = {
			section: e,
			actions: []
		}, t.push(r)), r.actions.push(n);
	}
	return t;
}
function xa({ actions: e, separated: t, onClose: n, size: r = "menu" }) {
	let a = ba(e), [o, s] = (0, U.useState)(null);
	return /* @__PURE__ */ (0, W.jsx)(W.Fragment, { children: a.map((e, a) => /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i((t || a > 0) && "mt-1 border-t border-border pt-1"),
		children: [e.section && /* @__PURE__ */ (0, W.jsx)("div", {
			className: "px-2 py-1.5 text-xs font-medium text-muted-foreground",
			children: e.section
		}), e.actions.map((e) => /* @__PURE__ */ (0, W.jsx)(Sa, {
			action: e,
			submenu: o?.id === e.id ? o : null,
			onOpenSubmenu: s,
			onClose: n,
			size: r
		}, e.id))]
	}, e.section)) });
}
function Sa({ action: e, submenu: t, onOpenSubmenu: n, onClose: r, size: a }) {
	let o = !!e.description, s = e.children ?? [], c = s.length > 0, l = (t) => {
		let r = t.getBoundingClientRect(), i = r.right + 224 > window.innerWidth ? r.left - 224 : r.right;
		n({
			id: e.id,
			x: i,
			y: r.top
		});
	};
	return /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [/* @__PURE__ */ (0, W.jsxs)("button", {
		type: "button",
		role: "menuitem",
		"aria-haspopup": c ? "menu" : void 0,
		"aria-expanded": c ? t != null : void 0,
		disabled: e.disabled,
		className: i("flex w-full gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", o ? "items-start" : "items-center", a === "sheet" && "min-h-10 py-2 text-sm", e.disabled && "cursor-not-allowed opacity-50", t && "bg-accent text-accent-foreground"),
		onClick: (i) => {
			e.disabled || (c ? t ? n(null) : l(i.currentTarget) : (e.onSelect(), r()));
		},
		children: [
			e.icon && /* @__PURE__ */ (0, W.jsx)(p, {
				icon: e.icon,
				className: i("shrink-0 text-sm", o && "mt-0.5", e.iconClassName ?? "text-muted-foreground")
			}),
			/* @__PURE__ */ (0, W.jsxs)("span", {
				className: "min-w-0 flex-1",
				children: [/* @__PURE__ */ (0, W.jsx)("span", {
					className: i(o && "font-medium"),
					children: e.label
				}), e.description && /* @__PURE__ */ (0, W.jsx)("span", {
					className: "mt-0.5 block text-xs text-muted-foreground",
					children: e.description
				})]
			}),
			c && /* @__PURE__ */ (0, W.jsx)(p, {
				icon: j,
				className: "shrink-0 text-sm text-muted-foreground"
			})
		]
	}), c && t && /* @__PURE__ */ (0, W.jsx)("div", {
		role: "menu",
		"aria-label": typeof e.label == "string" ? e.label : "Submenu",
		className: "fixed z-50 max-h-[calc(100vh-1rem)] min-w-[14rem] max-w-[calc(100vw-1rem)] overflow-auto rounded-md border border-border bg-popover p-1.5 text-popover-foreground shadow-lg shadow-black/5",
		style: {
			left: t.x,
			top: t.y
		},
		children: s.map((e) => /* @__PURE__ */ (0, W.jsx)(Sa, {
			action: e,
			submenu: null,
			onOpenSubmenu: n,
			onClose: r,
			size: a
		}, e.id))
	})] });
}
function Ca({ densityOverride: e, separated: t, onDensityChange: n, size: r = "menu" }) {
	let a = e ?? "inherit";
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i(t && "mt-1 border-t border-border pt-1"),
		children: [
			/* @__PURE__ */ (0, W.jsx)("div", {
				className: "px-2 py-1.5 text-xs font-medium text-muted-foreground",
				children: "Density"
			}),
			/* @__PURE__ */ (0, W.jsxs)("button", {
				type: "button",
				role: "menuitemradio",
				"aria-checked": a === "inherit",
				className: wa(a === "inherit", r),
				onClick: () => n(void 0),
				children: [
					/* @__PURE__ */ (0, W.jsx)(p, {
						icon: yt,
						className: "text-sm text-muted-foreground"
					}),
					/* @__PURE__ */ (0, W.jsx)("span", {
						className: "min-w-0 flex-1 truncate",
						children: "Use page density"
					}),
					a === "inherit" ? /* @__PURE__ */ (0, W.jsx)(p, {
						icon: L,
						className: "text-sm text-foreground"
					}) : /* @__PURE__ */ (0, W.jsx)("span", {
						className: "inline-block h-4 w-4",
						"aria-hidden": !0
					})
				]
			}),
			ya.map((e) => {
				let t = a === e.value;
				return /* @__PURE__ */ (0, W.jsxs)("button", {
					type: "button",
					role: "menuitemradio",
					"aria-checked": t,
					className: wa(t, r),
					onClick: () => n(e.value),
					children: [
						/* @__PURE__ */ (0, W.jsx)(p, {
							icon: e.icon,
							className: "text-sm text-muted-foreground"
						}),
						/* @__PURE__ */ (0, W.jsx)("span", {
							className: "min-w-0 flex-1 truncate",
							children: e.label
						}),
						t ? /* @__PURE__ */ (0, W.jsx)(p, {
							icon: L,
							className: "text-sm text-foreground"
						}) : /* @__PURE__ */ (0, W.jsx)("span", {
							className: "inline-block h-4 w-4",
							"aria-hidden": !0
						})
					]
				}, e.value);
			})
		]
	});
}
function wa(e, t) {
	return i("flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", t === "sheet" && "min-h-10 py-2 text-sm", e && "text-foreground");
}
var Ta = [
	{
		value: "system",
		icon: ht,
		label: "Use system theme"
	},
	{
		value: "light",
		icon: oe,
		label: "Light"
	},
	{
		value: "dark",
		icon: R,
		label: "Dark"
	}
];
function Ea({ value: e, separated: t, onChange: n, size: r = "menu" }) {
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i(t && "mt-1 border-t border-border pt-1"),
		children: [/* @__PURE__ */ (0, W.jsx)("div", {
			className: "px-2 py-1.5 text-xs font-medium text-muted-foreground",
			children: "Theme"
		}), Ta.map((t) => {
			let i = e === t.value;
			return /* @__PURE__ */ (0, W.jsxs)("button", {
				type: "button",
				role: "menuitemradio",
				"aria-checked": i,
				className: wa(i, r),
				onClick: () => n(t.value),
				children: [
					/* @__PURE__ */ (0, W.jsx)(p, {
						icon: t.icon,
						className: "text-sm text-muted-foreground"
					}),
					/* @__PURE__ */ (0, W.jsx)("span", {
						className: "min-w-0 flex-1 truncate",
						children: t.label
					}),
					i ? /* @__PURE__ */ (0, W.jsx)(p, {
						icon: L,
						className: "text-sm text-foreground"
					}) : /* @__PURE__ */ (0, W.jsx)("span", {
						className: "inline-block h-4 w-4",
						"aria-hidden": !0
					})
				]
			}, t.value);
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/dataTableMenuPosition.ts
function Da(e, t) {
	return {
		reference: {
			x: e.clientX,
			y: e.clientY,
			width: 0,
			height: 0
		},
		placement: "bottom-start",
		...t ? { columnKey: t } : {}
	};
}
function Oa(e, t) {
	let n = e.currentTarget.getBoundingClientRect();
	return {
		reference: {
			x: n.left,
			y: n.top,
			width: n.width,
			height: n.height
		},
		placement: "bottom-end",
		...t ? { columnKey: t } : {}
	};
}
var ka = 8;
function Aa(e) {
	let { x: t, y: n, width: r, height: i } = e.reference, a = (0, U.useMemo)(() => ({ getBoundingClientRect: () => ({
		x: t,
		y: n,
		top: n,
		left: t,
		width: r,
		height: i,
		right: t + r,
		bottom: n + i,
		toJSON: () => ({})
	}) }), [
		t,
		n,
		r,
		i
	]), { refs: o, floatingStyles: s } = y({
		strategy: "fixed",
		placement: e.placement,
		transform: !1,
		whileElementsMounted: tt,
		middleware: [
			et(6),
			nt({ padding: ka }),
			rt({ padding: ka }),
			it({
				padding: ka,
				apply({ availableHeight: e, elements: t }) {
					t.floating.style.maxHeight = `${Math.max(0, Math.floor(e))}px`;
				}
			})
		]
	});
	return (0, U.useLayoutEffect)(() => {
		o.setPositionReference(a);
	}, [o, a]), {
		ref: o.setFloating,
		style: s
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/DataTablePreferencesSheet.tsx
function ja({ open: e, columns: t, hiddenColumns: n, activeColumnKey: r, actions: a, showColumnVisibilityControl: o, showDensityControl: s, showThemeControl: c, themeMenuValue: l, densityOverride: u, visibleHideableColumnCount: d, onToggle: f, onShowAll: h, onDensityChange: g, onThemeChange: _, onClose: v }) {
	let y = r ? t.find((e) => e.key === r) : void 0, b = y && Wt(y) && d > 1, x = a.length > 0;
	return /* @__PURE__ */ (0, W.jsxs)(m, {
		open: e,
		onClose: v,
		title: "Table options",
		size: "sm",
		scrollBody: !0,
		children: [
			o && /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
				/* @__PURE__ */ (0, W.jsxs)("div", {
					className: "flex items-center justify-between gap-3 px-2 py-1.5 text-xs font-medium text-muted-foreground",
					children: [/* @__PURE__ */ (0, W.jsx)("span", { children: "Columns" }), /* @__PURE__ */ (0, W.jsx)("button", {
						type: "button",
						className: "min-h-10 rounded px-1.5 text-sm text-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none",
						onClick: h,
						children: "Show all"
					})]
				}),
				y && /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [/* @__PURE__ */ (0, W.jsxs)("button", {
					type: "button",
					disabled: !b,
					className: i("flex min-h-10 w-full items-center gap-2 rounded-md px-2 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", !b && "cursor-not-allowed opacity-50"),
					onClick: () => {
						b && (f(y), v());
					},
					children: [/* @__PURE__ */ (0, W.jsx)(p, {
						icon: ft,
						className: "text-sm text-muted-foreground"
					}), /* @__PURE__ */ (0, W.jsxs)("span", { children: ["Hide ", Ut(y)] })]
				}), /* @__PURE__ */ (0, W.jsx)("div", { className: "my-1 h-px bg-border" })] }),
				/* @__PURE__ */ (0, W.jsx)("div", { children: t.map((e) => {
					let t = Wt(e), r = n[e.key] !== !0, a = !t || r && d <= 1;
					return /* @__PURE__ */ (0, W.jsxs)("label", {
						className: i("flex min-h-10 cursor-pointer items-center gap-2 rounded-md px-3 text-sm transition-colors hover:bg-accent hover:text-accent-foreground", a && "cursor-not-allowed opacity-50"),
						children: [/* @__PURE__ */ (0, W.jsx)("input", {
							type: "checkbox",
							className: "h-5 w-5 rounded border-border",
							checked: r,
							disabled: a,
							onChange: () => f(e)
						}), /* @__PURE__ */ (0, W.jsx)("span", {
							className: "truncate",
							children: Ut(e)
						})]
					}, e.key);
				}) })
			] }),
			s && /* @__PURE__ */ (0, W.jsx)(Ca, {
				densityOverride: u,
				separated: o,
				onDensityChange: g,
				size: "sheet"
			}),
			x && /* @__PURE__ */ (0, W.jsx)(xa, {
				actions: a,
				separated: o || s,
				onClose: v,
				size: "sheet"
			}),
			c && _ && /* @__PURE__ */ (0, W.jsx)(Ea, {
				value: l ?? "system",
				separated: o || s || x,
				onChange: _,
				size: "sheet"
			})
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/SortableHeader.tsx
function Ma({ active: e, dir: t, onClick: n, align: r = "left", className: a, children: o }) {
	return /* @__PURE__ */ (0, W.jsxs)("button", {
		type: "button",
		onClick: n,
		className: i("inline-flex items-center gap-1 cursor-pointer select-none hover:text-foreground", e ? "text-primary" : "text-muted-foreground", r === "right" ? "justify-end" : r === "center" ? "justify-center" : "justify-start", a),
		children: [o, e ? /* @__PURE__ */ (0, W.jsx)("span", {
			"aria-hidden": !0,
			children: t === "asc" ? "↑" : "↓"
		}) : /* @__PURE__ */ (0, W.jsx)("span", {
			"aria-hidden": !0,
			className: "opacity-40",
			children: "↕"
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/cells/Timestamp.tsx
function Na({ value: e, format: t, showTitleOnHover: n = !0, className: r }) {
	let a = ge(e);
	if (!a) return /* @__PURE__ */ (0, W.jsx)("span", {
		className: i("text-muted-foreground", r),
		children: "—"
	});
	let o = O(a, t), s = n ? a.toISOString() : void 0;
	return /* @__PURE__ */ (0, W.jsx)("span", {
		className: i("tabular-nums whitespace-nowrap text-muted-foreground", r),
		title: s,
		children: o
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/IconButton.tsx
var Pa = (0, U.forwardRef)(({ icon: e, label: t, iconClassName: n, className: r, type: a = "button", ...o }, s) => /* @__PURE__ */ (0, W.jsx)("button", {
	ref: s,
	type: a,
	"aria-label": t,
	title: t,
	className: i("inline-flex shrink-0 items-center justify-center rounded-sm border-0 bg-transparent p-0 text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50", r),
	...o,
	children: /* @__PURE__ */ (0, W.jsx)(p, {
		icon: e,
		className: i("text-sm", n)
	})
}));
Pa.displayName = "IconButton";
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/properties-utils.ts
function Fa(e) {
	return e.replace(/[_-]+/g, " ").replace(/\.+/g, " ").replace(/\b\w/g, (e) => e.toUpperCase());
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/properties-context.ts
var Ia = (0, U.createContext)(null), La = "en-US";
function Ra(e, t = 1) {
	return new Intl.NumberFormat(La, { maximumFractionDigits: t }).format(Number.isFinite(e) ? e : 0);
}
function za(e) {
	return e.endsWith(".0") ? e.slice(0, -2) : e;
}
var Ba = [
	"B",
	"KB",
	"MB",
	"GB",
	"TB",
	"PB"
];
function Va(e, t) {
	let n = t?.suffix ?? "";
	if (e === void 0 || !Number.isFinite(e) || e <= 0) return `0 B${n}`;
	let r = t?.base ?? 1024, i = e, a = 0;
	for (; i >= r && a < Ba.length - 1;) i /= r, a++;
	return `${i >= 10 || a === 0 ? i.toFixed(0) : za(i.toFixed(1))} ${Ba[a]}${n}`;
}
function Ha(e) {
	return Va(e, { suffix: "/s" });
}
var Ua = [
	"",
	"K",
	"M",
	"B",
	"T"
];
function Wa(e) {
	if (e === void 0 || !Number.isFinite(e)) return "0";
	let t = e < 0 ? "-" : "", n = Math.abs(e);
	if (n < 1e3) return `${t}${Ra(n)}`;
	let r = 0;
	for (; n >= 1e3 && r < Ua.length - 1;) n /= 1e3, r++;
	return `${t}${n >= 10 ? n.toFixed(0) : za(n.toFixed(1))}${Ua[r]}`;
}
function Ga(e, t) {
	if (e === void 0 || !Number.isFinite(e)) return "0 ms";
	let n = t?.from === "s" ? e * 1e3 : e, r = Math.abs(n);
	return r < 1e3 ? `${Ra(n, 0)} ms` : r < 6e4 ? `${Ra(n / 1e3)} s` : r < 36e5 ? `${Ra(n / 6e4)} min` : `${Ra(n / 36e5)} h`;
}
var Ka = {
	none: (e) => Wa(e),
	short: (e) => Wa(e),
	percent: (e) => `${Ra(e)}%`,
	percentunit: (e) => `${Ra(e * 100)}%`,
	"%": (e) => `${Ra(e)}%`,
	bytes: (e) => Va(e, { base: 1024 }),
	decbytes: (e) => Va(e, { base: 1e3 }),
	Bps: (e) => Ha(e),
	binBps: (e) => Ha(e),
	ms: (e) => Ga(e, { from: "ms" }),
	s: (e) => Ga(e, { from: "s" })
};
function qa(e, t) {
	let n = Number.isFinite(e) ? e : 0;
	if (!t) return Ra(n);
	let r = Ka[t];
	return r ? r(n) : `${Ra(n)}${t}`;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/PropertyValue.tsx
function Ja({ item: e, density: t }) {
	let { value: n, type: r, unit: i } = e;
	if (n == null) return /* @__PURE__ */ (0, W.jsx)("span", {
		className: "font-mono italic text-muted-foreground",
		children: "null"
	});
	if (r === "json" || r === void 0 && typeof n == "object") return typeof n == "string" ? /* @__PURE__ */ (0, W.jsx)(E, {
		source: n,
		format: "json",
		density: t
	}) : /* @__PURE__ */ (0, W.jsx)(E, {
		data: n,
		format: "json",
		density: t
	});
	if (r === "bytes" || r === "duration" || r === "number" || i !== void 0) {
		if (typeof n != "number" || !Number.isFinite(n)) throw Error(`Property ${e.key}: expected a finite numeric value for ${r ?? i}`);
		let t = i ?? (r === "bytes" ? "bytes" : r === "duration" ? "ms" : void 0), a = t === "bytes" && n < 0 ? `-${qa(Math.abs(n), t)}` : qa(n, t);
		return /* @__PURE__ */ (0, W.jsx)("span", {
			className: "font-mono",
			title: String(n),
			children: a
		});
	}
	if (r === "date") {
		if (typeof n != "number" && typeof n != "string" && !(n instanceof Date)) throw Error(`Property ${e.key}: expected a date or timestamp`);
		return /* @__PURE__ */ (0, W.jsx)("time", {
			dateTime: new Date(n).toISOString(),
			title: String(n),
			children: new Date(n).toISOString()
		});
	}
	if (r === "boolean" && typeof n != "boolean") throw Error(`Property ${e.key}: expected a boolean`);
	return typeof n == "boolean" ? /* @__PURE__ */ (0, W.jsx)("input", {
		type: "checkbox",
		checked: n,
		disabled: !0,
		"aria-label": `${e.key} value`,
		title: String(n),
		className: "h-4 w-4 accent-primary"
	}) : /* @__PURE__ */ (0, W.jsx)("span", {
		className: "whitespace-pre-wrap break-words font-mono",
		children: String(n)
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/Properties.tsx
var Ya = "clicky-ui-properties", Xa = [
	{
		value: "compact",
		label: "Compact",
		icon: le
	},
	{
		value: "comfortable",
		label: "Comfortable",
		icon: K
	},
	{
		value: "spacious",
		label: "Spacious",
		icon: bt
	}
], Za = {
	spacious: {
		row: "px-density-4 py-density-3",
		label: "text-sm font-medium text-muted-foreground",
		value: "text-base text-foreground",
		gridTemplateColumns: "minmax(min(12rem, 40%), min(20rem, 50%)) minmax(0, 1fr)"
	},
	comfortable: {
		row: "px-density-3 py-density-2",
		label: "text-xs font-medium text-muted-foreground",
		value: "text-sm text-foreground",
		gridTemplateColumns: "minmax(min(7rem, 40%), min(12rem, 50%)) minmax(0, 1fr)"
	},
	compact: {
		row: "px-density-2 py-density-1.5",
		label: "font-mono text-[11px] text-muted-foreground",
		value: "text-xs text-foreground",
		gridTemplateColumns: "minmax(min(8rem, 40%), min(14rem, 50%)) minmax(0, 1fr)"
	}
};
function Qa({ items: e, renderLabel: t, renderValue: n, labelIcon: r, prefixActions: a, suffixActions: o, className: s, rowClassName: c, gridTemplateColumns: l, labelClassName: u, valueClassName: d, density: f = "comfortable", keyRef: m, showDensityMenu: h = !0, emptyMessage: g = "No properties" }) {
	let _ = (0, U.useContext)(Ia), [v, y] = (0, U.useState)();
	(0, U.useEffect)(() => y(to(m)), [m]);
	let b = e.filter((e) => !e.hidden);
	if (b.length === 0) return /* @__PURE__ */ (0, W.jsx)("div", {
		className: "text-sm text-muted-foreground",
		children: g
	});
	let x = _?.density ?? v ?? f, S = Za[x], C = _?.columns ?? l ?? S.gridTemplateColumns, w = _?.depth ?? 0;
	return /* @__PURE__ */ (0, W.jsxs)("dl", {
		className: i("grid gap-x-density-3 divide-y divide-border rounded-md border border-border bg-muted/20", _ && "col-span-full rounded-none border-0 bg-transparent", s),
		style: { gridTemplateColumns: _ ? "subgrid" : C },
		children: [h && !_ ? /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "col-span-full flex items-center justify-end px-density-2 py-density-1",
			children: [/* @__PURE__ */ (0, W.jsx)("dt", {
				className: "sr-only",
				children: "Properties display options"
			}), /* @__PURE__ */ (0, W.jsx)("dd", { children: /* @__PURE__ */ (0, W.jsx)($a, {
				density: x,
				onChange: (e) => {
					y(e), no(m, e);
				}
			}) })]
		}) : null, b.map((e) => {
			let s = typeof r == "function" ? r(e.key, e.value, e) : r, l = t ? t(e.key, e.value, e) : Fa(e.key), f = typeof l == "string" ? l : e.key;
			return /* @__PURE__ */ (0, W.jsxs)(U.Fragment, { children: [/* @__PURE__ */ (0, W.jsxs)("div", {
				className: i("col-span-full grid min-w-0 grid-cols-subgrid gap-density-3", S.row, c),
				children: [/* @__PURE__ */ (0, W.jsxs)("dt", {
					"aria-label": f,
					className: i("min-w-0 truncate", S.label, u),
					style: { paddingInlineStart: w * 16 },
					children: [/* @__PURE__ */ (0, W.jsxs)("div", {
						className: "flex min-w-0 items-center gap-density-1",
						children: [s ? /* @__PURE__ */ (0, W.jsx)(p, {
							name: s,
							className: "text-xs shrink-0"
						}) : null, /* @__PURE__ */ (0, W.jsx)("span", {
							className: "min-w-0 truncate",
							children: l
						})]
					}), e.subtitle ? /* @__PURE__ */ (0, W.jsx)("div", {
						className: "mt-0.5 text-[10px] font-normal text-muted-foreground/80",
						children: e.subtitle
					}) : null]
				}), /* @__PURE__ */ (0, W.jsx)("dd", {
					className: i("min-w-0 space-y-density-1", d),
					children: /* @__PURE__ */ (0, W.jsxs)("div", {
						className: "flex min-w-0 items-start gap-density-1",
						children: [
							/* @__PURE__ */ (0, W.jsx)(ao, {
								actions: a,
								item: e,
								className: "shrink-0 pt-0.5"
							}),
							/* @__PURE__ */ (0, W.jsx)("div", {
								className: i("min-w-0 max-w-full flex-1", S.value),
								children: n ? n(e.key, e.value, e) : /* @__PURE__ */ (0, W.jsx)(Ja, {
									item: e,
									density: x
								})
							}),
							/* @__PURE__ */ (0, W.jsx)(ao, {
								actions: o,
								item: e,
								className: "shrink-0 pt-0.5"
							})
						]
					})
				})]
			}), e.expandable && e.expanded && e.renderChildren ? /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "col-span-full grid grid-cols-subgrid",
				children: [/* @__PURE__ */ (0, W.jsxs)("dt", {
					className: "sr-only",
					children: [f, " properties"]
				}), /* @__PURE__ */ (0, W.jsx)("dd", {
					className: "col-span-full grid min-w-0 grid-cols-subgrid",
					children: /* @__PURE__ */ (0, W.jsx)(Ia.Provider, {
						value: {
							depth: w + 1,
							columns: C,
							density: x
						},
						children: e.renderChildren()
					})
				})]
			}) : null] }, e.key);
		})]
	});
}
function $a({ density: e, onChange: t }) {
	return /* @__PURE__ */ (0, W.jsx)(I, {
		align: "right",
		menuLabel: "Properties options",
		trigger: /* @__PURE__ */ (0, W.jsx)(Pa, {
			icon: N,
			label: "Properties options",
			className: "h-6 w-6"
		}),
		children: (n) => /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "min-w-44 py-1 text-popover-foreground",
			children: [/* @__PURE__ */ (0, W.jsx)("div", {
				className: "px-2 py-1 text-xs font-medium text-muted-foreground",
				children: "Density"
			}), Xa.map((r) => {
				let i = e === r.value;
				return /* @__PURE__ */ (0, W.jsxs)("button", {
					type: "button",
					role: "menuitemradio",
					"aria-checked": i,
					className: eo(i),
					onClick: () => {
						t(r.value), n();
					},
					children: [
						/* @__PURE__ */ (0, W.jsx)(p, {
							icon: r.icon,
							className: "text-sm text-muted-foreground"
						}),
						/* @__PURE__ */ (0, W.jsx)("span", {
							className: "min-w-0 flex-1 truncate",
							children: r.label
						}),
						i ? /* @__PURE__ */ (0, W.jsx)(p, {
							icon: L,
							className: "text-sm text-foreground"
						}) : /* @__PURE__ */ (0, W.jsx)("span", {
							className: "inline-block h-4 w-4",
							"aria-hidden": !0
						})
					]
				}, r.value);
			})]
		})
	});
}
function eo(e) {
	return i("flex w-full items-center gap-2 px-2 py-1.5 text-left text-sm", "hover:bg-accent focus:bg-accent focus:outline-none", e && "bg-accent/60");
}
function to(e) {
	let t = ro(e);
	if (t && typeof window < "u") try {
		let e = window.localStorage.getItem(t);
		return io(e) ? e : void 0;
	} catch {
		return;
	}
}
function no(e, t) {
	let n = ro(e);
	if (n && typeof window < "u") try {
		window.localStorage.setItem(n, t);
	} catch {}
}
function ro(e) {
	let t = e?.trim();
	return t ? `${Ya}:${t}:density` : void 0;
}
function io(e) {
	return e === "compact" || e === "comfortable" || e === "spacious";
}
function ao({ actions: e, item: t, className: n }) {
	if (!e || e.length === 0) return null;
	let r = e.filter((e) => !e.visible || e.visible(t.key, t.value, t));
	return r.length === 0 ? null : /* @__PURE__ */ (0, W.jsx)("span", {
		className: i("inline-flex items-center", n),
		children: r.map((e) => /* @__PURE__ */ (0, W.jsx)(oo, {
			action: e,
			item: t
		}, e.id))
	});
}
function oo({ action: e, item: t }) {
	let n = e.label(t.key, t.value, t), r = e.disabled?.(t.key, t.value, t) ?? !1;
	return /* @__PURE__ */ (0, W.jsx)("button", {
		type: "button",
		"aria-label": n,
		title: n,
		disabled: r,
		className: i("inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground", "hover:bg-accent hover:text-foreground", "disabled:opacity-35 disabled:hover:bg-transparent disabled:hover:text-muted-foreground"),
		onClick: (n) => {
			n.stopPropagation(), e.onClick(t.key, t.value, t);
		},
		children: /* @__PURE__ */ (0, W.jsx)(p, {
			...typeof e.icon == "string" ? { name: e.icon } : { icon: e.icon },
			className: "text-xs"
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/cells/tag-utils.ts
function so(e, t = "=") {
	return e == null ? [] : Array.isArray(e) ? e.map((e) => co(e, t)).filter((e) => e.value !== "") : typeof e == "object" ? Object.entries(e).map(([e, n]) => {
		if (n == null) return null;
		let r = String(n).trim();
		return r ? {
			key: e,
			value: r,
			display: `${e}${t}${r}`,
			token: `${e}${t}${r}`
		} : null;
	}).filter((e) => e !== null) : [];
}
function co(e, t) {
	if (typeof e == "string") {
		let n = e.trim(), r = n.indexOf("="), i = n.indexOf(":"), a = r >= 0 && (i < 0 || r < i) ? r : i >= 0 && i < n.length - 1 ? i : -1;
		if (a > 0) {
			let e = n.slice(0, a), r = n.slice(a + 1);
			return {
				key: e,
				value: r,
				display: `${e}${t}${r}`,
				token: `${e}${t}${r}`
			};
		}
		return {
			value: n,
			display: n,
			token: n
		};
	}
	let n = "key" in e && typeof e.key == "string" ? e.key : "name" in e && typeof e.name == "string" ? e.name : void 0, r = String(e.value ?? "").trim();
	return n ? {
		key: n,
		value: r,
		display: `${n}${t}${r}`,
		token: `${n}${t}${r}`
	} : {
		value: r,
		display: r,
		token: r
	};
}
function lo(e, t = "=") {
	return so(e, t).map((e) => e.token);
}
function uo(e, t = "=") {
	let n = e.indexOf(t);
	return n <= 0 ? {
		key: "",
		value: e
	} : {
		key: e.slice(0, n),
		value: e.slice(n + t.length)
	};
}
var fo = (0, U.createContext)({
	getMode: () => "neutral",
	toggleInclude: () => {},
	toggleExclude: () => {}
});
function po() {
	return (0, U.useContext)(fo);
}
function mo(e, t) {
	return {
		getMode: (t) => e[t] ?? "neutral",
		toggleInclude: (n) => {
			let r = { ...e };
			r[n] === "include" ? delete r[n] : r[n] = "include", t(r);
		},
		toggleExclude: (n) => {
			let r = { ...e };
			r[n] === "exclude" ? delete r[n] : r[n] = "exclude", t(r);
		}
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/cells/TagList.tsx
function ho({ value: e, children: t }) {
	return /* @__PURE__ */ (0, W.jsx)(fo.Provider, {
		value: e,
		children: t
	});
}
function go(e) {
	typeof navigator < "u" && navigator.clipboard?.writeText && navigator.clipboard.writeText(e).catch(() => void 0);
}
function _o({ tag: e, actions: t, compact: n = !1 }) {
	let r = po(), a = r.getMode(e.token), o = (0, U.useCallback)((t) => {
		t.stopPropagation(), r.toggleInclude(e.token);
	}, [r, e.token]), s = (0, U.useCallback)((t) => {
		t.stopPropagation(), r.toggleExclude(e.token);
	}, [r, e.token]), c = (0, U.useCallback)((t) => {
		t.stopPropagation(), go(e.value);
	}, [e.value]), l = a === "include", u = a === "exclude", d = n && !!e.key, f = /* @__PURE__ */ (0, W.jsx)("span", {
		...d ? { title: e.display } : {},
		className: i("inline-flex items-center rounded-md border bg-background align-middle", l ? "border-green-500/50 bg-green-500/5" : u ? "border-red-500/50 bg-red-500/5" : "border-border"),
		children: e.key && !n ? /* @__PURE__ */ (0, W.jsx)(w, {
			size: "xs",
			variant: "label",
			tone: "neutral",
			label: e.key,
			value: e.value,
			maxWidth: 32,
			truncate: "auto",
			clickToCopy: !1,
			className: "border-transparent shadow-none"
		}) : /* @__PURE__ */ (0, W.jsx)(w, {
			size: "xs",
			variant: "soft",
			tone: "neutral",
			maxWidth: 32,
			truncate: "auto",
			clickToCopy: !1,
			children: e.value
		})
	}), p = /* @__PURE__ */ (0, W.jsxs)("span", {
		className: "inline-flex items-center gap-0.5",
		children: [
			/* @__PURE__ */ (0, W.jsx)(vo, {
				ariaLabel: `Include ${e.display}`,
				onClick: o,
				active: l,
				activeClassName: "bg-green-500/20 text-green-700 dark:text-green-400",
				icon: pt
			}),
			/* @__PURE__ */ (0, W.jsx)(vo, {
				ariaLabel: `Exclude ${e.display}`,
				onClick: s,
				active: u,
				activeClassName: "bg-red-500/20 text-red-700 dark:text-red-400",
				icon: mt
			}),
			/* @__PURE__ */ (0, W.jsx)(vo, {
				ariaLabel: `Copy ${e.display}`,
				onClick: c,
				icon: se
			})
		]
	});
	return t === "inline" ? /* @__PURE__ */ (0, W.jsxs)("span", {
		className: "inline-flex items-center gap-0.5 rounded-md border border-border bg-background pr-0.5",
		children: [f, p]
	}) : /* @__PURE__ */ (0, W.jsx)(ae, {
		placement: "top",
		delay: 120,
		arrow: !1,
		trigger: f,
		cardClassName: "!p-1",
		children: p
	});
}
function vo({ ariaLabel: e, onClick: t, active: n, activeClassName: r, icon: a }) {
	return /* @__PURE__ */ (0, W.jsx)("button", {
		type: "button",
		"aria-label": e,
		title: e,
		onClick: t,
		className: i("inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground", "hover:bg-accent hover:text-foreground", n && r),
		children: /* @__PURE__ */ (0, W.jsx)(p, {
			icon: a,
			className: "text-xs"
		})
	});
}
function yo({ tags: e, maxVisible: t = 3, className: n, actions: r = "hover", compact: a, wrap: o = !1 }) {
	let s = F(), c = a ?? s === "compact";
	if (e.length === 0) return /* @__PURE__ */ (0, W.jsx)("span", {
		className: "text-muted-foreground",
		children: "—"
	});
	if (o) return /* @__PURE__ */ (0, W.jsx)("span", {
		className: i("flex min-w-0 flex-wrap items-center gap-1", n),
		children: e.map((e, t) => /* @__PURE__ */ (0, W.jsx)(_o, {
			tag: e,
			actions: r,
			compact: c
		}, `${e.display}-${t}`))
	});
	let l = e.slice(0, t), u = e.slice(t);
	return /* @__PURE__ */ (0, W.jsxs)("span", {
		className: i("flex min-w-0 items-center gap-1", n),
		children: [/* @__PURE__ */ (0, W.jsx)("span", {
			className: "flex min-w-0 flex-1 flex-nowrap items-center gap-1 overflow-hidden",
			children: l.map((e, t) => /* @__PURE__ */ (0, W.jsx)(_o, {
				tag: e,
				actions: r,
				compact: c
			}, `${e.display}-${t}`))
		}), u.length > 0 && /* @__PURE__ */ (0, W.jsx)(ae, {
			placement: "top",
			trigger: /* @__PURE__ */ (0, W.jsxs)(w, {
				size: "xs",
				variant: "outlined",
				tone: "neutral",
				clickToCopy: !1,
				className: "shrink-0",
				children: ["+", u.length]
			}),
			cardClassName: "min-w-72 max-w-[90vw] whitespace-normal !p-0",
			children: /* @__PURE__ */ (0, W.jsx)(bo, { tags: u })
		})]
	});
}
function bo({ tags: e, rowClassName: t, gridTemplateColumns: n }) {
	let r = po(), i = (0, U.useMemo)(() => e.map((e, t) => ({
		key: `${e.token}-${t}`,
		value: e
	})), [e]), a = e.some((e) => !!e.key), o = n ?? (a ? "minmax(4rem, 8rem) minmax(0, 1fr)" : "0 minmax(0, 1fr)");
	return /* @__PURE__ */ (0, W.jsx)(Qa, {
		items: i,
		density: "compact",
		className: "border-0 bg-transparent",
		...t ? { rowClassName: t } : {},
		gridTemplateColumns: o,
		renderLabel: (e, t) => t.key ? /* @__PURE__ */ (0, W.jsx)("span", {
			className: "font-mono",
			children: t.key
		}) : null,
		renderValue: (e, t) => /* @__PURE__ */ (0, W.jsx)("span", {
			className: "block truncate font-mono",
			title: t.value,
			children: t.value
		}),
		suffixActions: [
			{
				id: "include",
				icon: pt,
				label: (e, t) => `Include ${t.display}`,
				onClick: (e, t) => r.toggleInclude(t.token)
			},
			{
				id: "exclude",
				icon: mt,
				label: (e, t) => `Exclude ${t.display}`,
				onClick: (e, t) => r.toggleExclude(t.token)
			},
			{
				id: "copy",
				icon: se,
				label: (e, t) => `Copy ${t.display}`,
				onClick: (e, t) => go(t.value)
			}
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/cells/CellFilterActions.tsx
function xo({ children: e, value: t, displayValue: n = t, mode: r, onChange: i }) {
	return /* @__PURE__ */ (0, W.jsx)(ae, {
		placement: "top",
		delay: 120,
		arrow: !1,
		trigger: /* @__PURE__ */ (0, W.jsx)("span", {
			className: "relative inline-flex min-w-0",
			children: e
		}),
		cardClassName: "!p-1",
		children: /* @__PURE__ */ (0, W.jsxs)("span", {
			className: "inline-flex items-center gap-0.5",
			children: [/* @__PURE__ */ (0, W.jsx)(So, {
				label: `Include ${n}`,
				icon: pt,
				active: r === "include",
				activeClassName: "bg-green-500/20 text-green-700 dark:text-green-400",
				onClick: () => i(r === "include" ? void 0 : "include")
			}), /* @__PURE__ */ (0, W.jsx)(So, {
				label: `Exclude ${n}`,
				icon: mt,
				active: r === "exclude",
				activeClassName: "bg-red-500/20 text-red-700 dark:text-red-400",
				onClick: () => i(r === "exclude" ? void 0 : "exclude")
			})]
		})
	});
}
function So({ label: e, icon: t, active: n, activeClassName: r, onClick: a }) {
	return /* @__PURE__ */ (0, W.jsx)("button", {
		type: "button",
		"aria-label": e,
		"aria-pressed": n,
		title: e,
		onClick: (e) => {
			e.stopPropagation(), a();
		},
		className: i("inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground", "hover:bg-accent hover:text-foreground", n && r),
		children: /* @__PURE__ */ (0, W.jsx)(p, {
			icon: t,
			className: "text-xs"
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/cells/StatusDot.tsx
var Co = {
	success: "bg-emerald-500",
	error: "bg-rose-500",
	warning: "bg-amber-500",
	info: "bg-sky-500"
}, wo = {
	xs: "h-1.5 w-1.5",
	sm: "h-2 w-2",
	md: "h-2.5 w-2.5"
};
function To({ status: e, size: t = "sm", label: n, title: r, className: a }) {
	let o = r ?? n ?? e;
	return /* @__PURE__ */ (0, W.jsxs)("span", {
		className: i("inline-flex items-center gap-1.5 align-middle", a),
		title: o,
		children: [/* @__PURE__ */ (0, W.jsx)("span", {
			role: "img",
			"aria-label": n ?? e,
			className: i("inline-block shrink-0 rounded-full", Co[e], wo[t])
		}), n && /* @__PURE__ */ (0, W.jsx)("span", {
			className: "truncate text-xs",
			children: n
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/cells/status-mapping.ts
var Eo = /* @__PURE__ */ new Set([
	"error",
	"err",
	"fatal",
	"critical",
	"crit",
	"panic",
	"failed",
	"fail",
	"down",
	"unhealthy"
]), Do = /* @__PURE__ */ new Set([
	"warn",
	"warning",
	"degraded",
	"slow",
	"stale"
]), Oo = /* @__PURE__ */ new Set([
	"ok",
	"healthy",
	"success",
	"info",
	"debug",
	"trace",
	"up",
	"running",
	"ready"
]);
function ko(e) {
	if (e == null) return null;
	if (typeof e == "boolean") return e ? "success" : "error";
	let t = String(e).toLowerCase().trim();
	return t ? Eo.has(t) ? "error" : Do.has(t) ? "warning" : Oo.has(t) ? "success" : null : null;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/DataTable.grouping.ts
var Ao = "\0empty";
function jo(e, t) {
	return e === "\0empty" ? t ?? "(empty)" : e;
}
function Mo(e) {
	if (e == null || e === "") return Ao;
	switch (typeof e) {
		case "string":
		case "number":
		case "boolean":
		case "bigint": return String(e);
		default: throw Error(`DataTable automatic grouping requires scalar values; received ${Array.isArray(e) ? "array" : typeof e}`);
	}
}
function No(e, t) {
	let n = /* @__PURE__ */ new Map();
	for (let r of e) {
		let e = t(r), i = n.get(e);
		i ? i.push(r) : n.set(e, [r]);
	}
	return Array.from(n, ([e, t]) => ({
		key: e,
		rows: t
	}));
}
function Po(e, t) {
	return typeof t == "function" ? t(e.key, e.rows) : t === !0;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/DataTableGroupingControls.tsx
function Fo({ modes: e, value: t, hasGroups: n, allExpanded: r, allCollapsed: i, onModeChange: a, onExpandAll: o, onCollapseAll: s }) {
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "flex shrink-0 items-center gap-1",
		role: "group",
		"aria-label": "Grouping controls",
		children: [
			/* @__PURE__ */ (0, W.jsx)("div", {
				className: "w-44 shrink-0",
				children: /* @__PURE__ */ (0, W.jsx)(pn, {
					"aria-label": "Group rows by",
					value: t,
					options: e,
					onChange: (e) => a(e.target.value)
				})
			}),
			/* @__PURE__ */ (0, W.jsx)(Io, {
				label: "Expand all groups",
				pressed: r,
				disabled: !n,
				onClick: o,
				icon: Je
			}),
			/* @__PURE__ */ (0, W.jsx)(Io, {
				label: "Collapse all groups",
				pressed: i,
				disabled: !n,
				onClick: s,
				icon: ze
			})
		]
	});
}
function Io({ label: e, pressed: t, disabled: n, onClick: r, icon: a }) {
	return /* @__PURE__ */ (0, W.jsx)("button", {
		type: "button",
		"aria-label": e,
		title: e,
		"aria-pressed": t,
		disabled: n,
		onClick: r,
		className: i("inline-flex h-control-h w-control-h items-center justify-center rounded-md border border-border bg-background text-muted-foreground hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50", t && "bg-accent text-accent-foreground"),
		children: /* @__PURE__ */ (0, W.jsx)(p, {
			icon: a,
			className: "text-sm"
		})
	});
}
function Lo(e) {
	return e.display ? e.display : e.menu ? "menu" : e.primary === !1 || e.section !== void 0 || e.children?.length ? "overflow" : "button";
}
function Ro(e, t) {
	let n = [], r = [];
	for (let t of e) {
		let e = Lo(t);
		e === "menu" ? n.push(t) : e === "button" && r.push(t);
	}
	let i = r.filter((e) => e.primary), a = i.length ? i : r.slice(0, t?.maxButtons ?? 3), o = new Set([...n, ...a].map((e) => e.id));
	return {
		menus: n,
		buttons: a,
		overflow: e.filter((e) => !o.has(e.id))
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/SelectionActionBar.tsx
function zo({ count: e, onClear: t }) {
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		role: "status",
		"data-testid": "data-table-selection-count",
		className: "flex flex-wrap items-center gap-density-2 text-xs text-muted-foreground",
		children: [/* @__PURE__ */ (0, W.jsxs)("span", {
			className: "font-medium text-foreground",
			children: [e.toLocaleString(), " selected"]
		}), /* @__PURE__ */ (0, W.jsx)(a, {
			type: "button",
			variant: "outline",
			size: "sm",
			onClick: t,
			children: "Clear selection"
		})]
	});
}
var Bo = "Cancel";
function Vo({ action: e, context: t, onCancel: n, onConfirm: r }) {
	let i = typeof e.confirm == "object" ? e.confirm : {}, o = t.selectedScope?.total ?? t.selectedRowIds.length, s = typeof i.message == "function" ? i.message(t) : i.message ?? `This will apply to ${o.toLocaleString()} selected ${o === 1 ? "row" : "rows"}.`, c = i.confirmLabel ?? (typeof e.label == "string" ? e.label : "Confirm");
	return /* @__PURE__ */ (0, W.jsx)(m, {
		open: !0,
		size: "sm",
		expandable: !1,
		onClose: n,
		title: i.title ?? e.label,
		footer: /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "flex justify-end gap-density-2",
			children: [/* @__PURE__ */ (0, W.jsx)(a, {
				type: "button",
				variant: "ghost",
				size: "sm",
				onClick: n,
				children: i.cancelLabel ?? Bo
			}), /* @__PURE__ */ (0, W.jsx)(a, {
				type: "button",
				size: "sm",
				variant: e.variant === "destructive" ? "destructive" : "default",
				onClick: r,
				children: c
			})]
		}),
		children: /* @__PURE__ */ (0, W.jsx)("p", {
			className: "text-sm text-muted-foreground",
			children: s
		})
	});
}
function Ho({ actions: e, context: t, showCount: n = !0, maxButtons: r, className: i }) {
	let [o, s] = (0, U.useState)(null), [c, l] = (0, U.useState)(null), u = (0, U.useRef)(!1), d = (0, U.useCallback)((e) => {
		if (u.current || e.disabled) return;
		u.current = !0, s(e.id);
		let n = () => {
			u.current = !1, s(null);
		}, r;
		try {
			r = e.onSelect(t);
		} catch (e) {
			throw n(), e;
		}
		r instanceof Promise ? r.finally(n) : n();
	}, [t]), f = (0, U.useCallback)((e) => {
		e.confirm ? l(e) : d(e);
	}, [d]), { menus: m, buttons: h, overflow: g } = Ro(e, { maxButtons: r }), _ = [...m, ...h], v = o !== null, y = g.find((e) => e.id === o), b = (e) => ({
		label: e.label,
		onSelect: () => f(e),
		disabled: e.disabled || v,
		...e.icon ? { icon: e.icon } : {},
		...e.section === void 0 ? {} : { group: e.section },
		...e.children?.length ? { children: e.children.map(b) } : {}
	}), x = /* @__PURE__ */ new Map();
	for (let e of g) {
		let t = e.section ?? "", n = x.get(t);
		n ? n.push(e) : x.set(t, [e]);
	}
	let S = [...x.values()].flat().map(b);
	return /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
		n ? /* @__PURE__ */ (0, W.jsx)(zo, {
			count: t.selectedScope?.total ?? t.selectedRowIds.length,
			onClear: t.clearSelection
		}) : null,
		/* @__PURE__ */ (0, W.jsxs)("span", {
			className: i,
			children: [_.map((e) => {
				let n = o === e.id, r = m.includes(e), i = /* @__PURE__ */ (0, W.jsxs)(a, {
					type: "button",
					size: "sm",
					variant: e.variant ?? "outline",
					disabled: e.disabled || v && !n,
					loading: n,
					...r ? { "aria-haspopup": "menu" } : { onClick: () => f(e) },
					children: [
						e.icon && !n ? /* @__PURE__ */ (0, W.jsx)(p, {
							icon: e.icon,
							...e.iconClassName ? { className: e.iconClassName } : {}
						}) : null,
						n && e.pendingLabel !== void 0 ? e.pendingLabel : e.label,
						r ? /* @__PURE__ */ (0, W.jsx)(p, { icon: z }) : null
					]
				});
				if (!r) return /* @__PURE__ */ (0, W.jsx)("span", { children: i }, e.id);
				let s = e.menu ? { children: (n) => e.menu({
					context: t,
					close: n,
					busy: v,
					run: (t) => f({
						...e,
						...t.pendingLabel === void 0 ? {} : { pendingLabel: t.pendingLabel },
						...t.confirm === void 0 ? {} : { confirm: t.confirm },
						onSelect: t.onSelect
					})
				}) } : { items: (e.children ?? []).map(b) };
				return /* @__PURE__ */ (0, W.jsx)(I, {
					align: "left",
					menuLabel: typeof e.label == "string" ? e.label : e.id,
					trigger: i,
					...s
				}, e.id);
			}), g.length > 0 ? /* @__PURE__ */ (0, W.jsx)(I, {
				align: "right",
				menuLabel: "Selection actions",
				items: S,
				trigger: /* @__PURE__ */ (0, W.jsxs)(a, {
					type: "button",
					size: "sm",
					variant: "outline",
					"aria-haspopup": "menu",
					"aria-label": "More selection actions",
					disabled: v && !y,
					loading: !!y,
					children: [
						y ? null : /* @__PURE__ */ (0, W.jsx)(p, { icon: k }),
						"More",
						/* @__PURE__ */ (0, W.jsx)(p, { icon: z })
					]
				})
			}) : null]
		}),
		c ? /* @__PURE__ */ (0, W.jsx)(Vo, {
			action: c,
			context: t,
			onCancel: () => l(null),
			onConfirm: () => {
				let e = c;
				l(null), d(e);
			}
		}) : null
	] });
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/DataTableOverflowCue.tsx
var Uo = {
	overflow: !1,
	atStart: !0,
	atEnd: !0,
	interacted: !1
};
function Wo({ scroll: e }) {
	let [t, n] = (0, U.useState)(Uo);
	return (0, U.useEffect)(() => {
		if (!e) {
			n(Uo);
			return;
		}
		let t = () => {
			let t = e.scrollWidth - e.clientWidth, r = e.scrollLeft <= 1;
			n((n) => ({
				overflow: t > 1,
				atStart: r,
				atEnd: t <= 1 || e.scrollLeft >= t - 1,
				interacted: n.interacted || !r
			}));
		};
		t(), e.addEventListener("scroll", t, { passive: !0 }), window.addEventListener("resize", t);
		let r = typeof ResizeObserver > "u" ? void 0 : new ResizeObserver(t);
		r?.observe(e);
		let i = e.querySelector("table");
		return i && r?.observe(i), () => {
			e.removeEventListener("scroll", t), window.removeEventListener("resize", t), r?.disconnect();
		};
	}, [e]), t.overflow ? /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
		t.atStart ? null : /* @__PURE__ */ (0, W.jsx)("span", {
			"aria-hidden": !0,
			"data-testid": "data-table-overflow-left",
			className: "pointer-events-none absolute inset-y-0 left-0 z-20 w-6 bg-gradient-to-r from-background to-transparent"
		}),
		t.atEnd ? null : /* @__PURE__ */ (0, W.jsx)("span", {
			"aria-hidden": !0,
			"data-testid": "data-table-overflow-right",
			className: "pointer-events-none absolute inset-y-0 right-0 z-20 w-8 bg-gradient-to-l from-background to-transparent"
		}),
		t.atStart && !t.interacted ? /* @__PURE__ */ (0, W.jsxs)("span", {
			className: i("pointer-events-none absolute right-2 top-10 z-30 inline-flex items-center gap-1 rounded-full", "border border-border bg-background/95 px-2 py-1 text-[10px] font-medium text-muted-foreground shadow-sm"),
			children: ["More columns", /* @__PURE__ */ (0, W.jsx)(Ct, {
				"aria-hidden": !0,
				className: "size-3"
			})]
		}) : null
	] }) : null;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/DataTable.tsx
var Go = 64, Ko = 160, qo = 224, Jo = 96, Yo = 288, Xo = 256, Zo = "clicky-ui-data-table-column-widths", Qo = "clicky-ui-data-table-column-visibility", $o = "clicky-ui-data-table-density", es = "px-2.5 py-1.5 density-compact:px-2 density-compact:py-1 density-comfortable:px-2.5 density-comfortable:py-1.5 density-spacious:px-4 density-spacious:py-3", ts = "px-2.5 py-1.5 density-compact:px-2 density-compact:py-0.5 density-comfortable:px-2.5 density-comfortable:py-1.5 density-spacious:px-4 density-spacious:py-3", ns = [
	{
		label: "Last 5 minutes",
		from: "now-5m",
		to: "now"
	},
	{
		label: "Last 15 minutes",
		from: "now-15m",
		to: "now"
	},
	{
		label: "Last 1 hour",
		from: "now-1h",
		to: "now"
	},
	{
		label: "Last 6 hours",
		from: "now-6h",
		to: "now"
	},
	{
		label: "Last 24 hours",
		from: "now-24h",
		to: "now"
	},
	{
		label: "Last 7 days",
		from: "now-7d",
		to: "now"
	},
	{
		label: "Last 30 days",
		from: "now-30d",
		to: "now"
	}
];
function rs(e, t) {
	return e ? e.cursor ? e.cursor.current === void 0 && e.cursor.next === void 0 : e.page === 0 ? e.hasMore == null ? e.total != null && e.totalRelation !== "gte" ? e.total <= e.pageSize : t < e.pageSize : !e.hasMore : !1 : !0;
}
function is({ grouping: e, groupingModes: t, groupingMode: n, defaultGroupingMode: r, onGroupingModeChange: i }) {
	if (e && t) throw Error("DataTable accepts grouping or groupingModes, not both");
	if (!t) {
		if (n !== void 0 || r !== void 0) throw Error("DataTable groupingMode values require groupingModes");
		return;
	}
	if (t.length === 0) throw Error("DataTable groupingModes must not be empty");
	let a = new Set(t.map((e) => e.value));
	if (a.size !== t.length) throw Error("DataTable groupingModes values must be unique");
	if (n !== void 0 && !i) throw Error("DataTable controlled groupingMode requires onGroupingModeChange");
	if (n !== void 0 && r !== void 0) throw Error("DataTable cannot combine groupingMode with defaultGroupingMode");
	let o = n ?? r;
	if (o !== void 0 && !a.has(o)) throw Error(`DataTable grouping mode "${o}" does not exist`);
}
function as({ data: e, error: t, columns: n, loading: r = !1, loadingMessage: a = "Loading results…", loadingRowCount: o = 8, emptyMessage: c = "No data", virtualize: u, className: d, scrollContainerClassName: f, autoFilter: p = !1, boundedFilters: g, onBoundedFiltersChange: _, showGlobalFilter: v, manualFilter: y = !1, globalFilter: b, onGlobalFilterChange: x, globalFilterPlaceholder: S = "Search all columns…", defaultSort: C, sort: w, onSortChange: T, manualSort: E = !1, filterBarProps: D, externalSearch: O, externalTimeRange: k, externalFilters: j, cellFilters: M, onCellFilterChange: N, pagination: ee, infinite: P, scrollContainerRef: te, getRowId: ne, rowSelection: F, selectionActions: I, selectionBar: L = "toolbar", getRowClassName: R, footer: ie, grouping: ae, groupingModes: oe, groupingMode: se, defaultGroupingMode: ce, onGroupingModeChange: le, onRowClick: ue, isRowClickable: fe, getRowHref: pe, renderExpandedRow: he, expandedRowIds: z, onExpandedRowIdsChange: _e, renderCard: B, detailStyle: ye = "row", detailDialogTitle: be, detailDialogSize: xe = "lg", resizableColumns: Se = !0, persistColumnWidths: Ce = !0, columnResizeStorageKey: we, hideableColumns: Te = !0, persistColumnVisibility: Ee = !0, columnVisibilityStorageKey: De, defaultHiddenColumns: Oe, density: ke, defaultDensity: Ae, onDensityChange: je, persistDensity: Me = !0, densityStorageKey: Ne, showDensityControl: Pe, themeMenuValue: Fe, onThemeChange: Ie, showThemeControl: Le = !1, showHeaderFilters: Re = !0, menuActions: ze }) {
	is({
		grouping: ae,
		groupingModes: oe,
		groupingMode: se,
		defaultGroupingMode: ce,
		onGroupingModeChange: le
	});
	let V = (0, U.useMemo)(() => n.map((e) => typeof e == "string" ? {
		key: e,
		label: e
	} : e), [n]), Be = (0, U.useMemo)(() => V.map((e) => e.key).join("|"), [V]), Ve = we ?? `${Zo}:${Be}`, He = De ?? `${Qo}:${Be}`, Ue = Ne ?? `${$o}:${Be}`, [We, Ge] = (0, U.useState)(() => Ce ? zs(Ve, V) : {}), Ke = (Oe ?? []).join("|"), H = (0, U.useMemo)(() => new Set(Oe ?? []), [Ke]), [qe, Je] = (0, U.useState)(() => Ee ? Vs(He, V, H) : qs({}, V, H)), Ye = ke !== void 0, [Xe, Ze] = (0, U.useState)(() => Ye ? ke : Me ? Us(Ue) ?? Ae : Ae), [Qe, $e] = (0, U.useState)(null), [et, tt] = (0, U.useState)(null), nt = me(ve("sm")), [rt, it] = (0, U.useState)({}), [at, ot] = (0, U.useState)({}), [st, ct] = (0, U.useState)({}), lt = g ?? st, ut = (e, t) => {
		let n = Ts(lt, e, t);
		g ? _?.(n) : ct(n);
	}, [dt, ft] = (0, U.useState)({}), [pt, mt] = (0, U.useState)(null), [ht, gt] = (0, U.useState)(ce), [_t, vt] = (0, U.useState)(""), [yt, bt] = (0, U.useState)(() => {
		let e = V.find((e) => e.kind === "timestamp" && e.timestamp?.defaultRange);
		return {
			from: e?.timestamp?.defaultRange?.from ?? "",
			to: e?.timestamp?.defaultRange?.to ?? ""
		};
	});
	(0, U.useEffect)(() => {
		Ge((e) => {
			let t = Ks({
				...Ce ? zs(Ve, V) : {},
				...e
			}, V);
			return Js(e, t) ? e : t;
		});
	}, [
		Be,
		Ce,
		Ve
	]), (0, U.useEffect)(() => {
		Ce && Bs(Ve, We);
	}, [
		We,
		Ce,
		Ve
	]), (0, U.useEffect)(() => {
		Je((e) => {
			let t = qs({
				...Ee ? Vs(He, V, H) : {},
				...e
			}, V, H);
			return Ys(e, t) ? e : t;
		});
	}, [
		Be,
		H,
		Ee,
		He
	]), (0, U.useEffect)(() => {
		Ee && Hs(He, qe, H);
	}, [
		H,
		qe,
		Ee,
		He
	]), (0, U.useEffect)(() => {
		Ye || Ze(Me ? Us(Ue) ?? Ae : Ae);
	}, [
		Ae,
		Ye,
		Me,
		Ue
	]), (0, U.useEffect)(() => {
		!Ye && Me && Ws(Ue, Xe);
	}, [
		Ye,
		Xe,
		Me,
		Ue
	]);
	let xt = (0, U.useCallback)(() => {
		$e(null), tt(null);
	}, []), St = !!Qe && nt;
	h(!!(Qe || et), xt, !St), (0, U.useEffect)(() => {
		if ((Qe || et) && !St) return document.addEventListener("click", xt), () => {
			document.removeEventListener("click", xt);
		};
	}, [
		xt,
		Qe,
		St,
		et
	]);
	let Ct = (0, U.useMemo)(() => e.map((e, t) => ({
		id: ne?.(e, t) ?? String(t),
		row: e
	})), [e, ne]), wt = (0, U.useMemo)(() => {
		let t = {};
		for (let n of V) {
			if (n.kind !== "timestamp") continue;
			let r = [];
			for (let t of e) {
				let e = zt(t, n), i = ge(e);
				i && r.push(i);
			}
			let i = re(r);
			t[n.key] = de(n.timestamp?.mode ?? "auto", i);
		}
		return t;
	}, [V, e]), G = (0, U.useMemo)(() => V.map((e) => Ps(e, wt[e.key])), [V, wt]), Tt = se ?? ht ?? ce ?? oe?.[0]?.value, K = oe?.find((e) => e.value === Tt), Et = (0, U.useMemo)(() => {
		if (!K) return ae;
		if (K.type === "none") return;
		if (K.type === "custom") return K;
		let e = G.find((e) => e.key === K.columnKey);
		if (!e) throw Error(`DataTable grouping mode "${K.value}" references missing column "${K.columnKey}"`);
		return {
			...K,
			getGroupKey: (t) => Mo(zt(t, e)),
			getGroupLabel: (e, t) => {
				let n = K.getGroupLabel?.(e, t);
				return n === void 0 ? jo(e, K.emptyGroupLabel) : n;
			}
		};
	}, [
		K,
		G,
		ae
	]), q = (0, U.useMemo)(() => {
		if (!Te) return G;
		let e = G.filter((e) => !qe[e.key]);
		return e.length > 0 ? e : G;
	}, [
		G,
		qe,
		Te
	]), Dt = (0, U.useMemo)(() => G.filter(Wt).length, [G]), Ot = (0, U.useMemo)(() => q.filter(Wt).length, [q]), kt = !B && Te && Dt > 1, At = Pe ?? kt, jt = !!(ze && ze.length > 0), Mt = kt || At || Le || jt, Nt = Ye ? ke : Xe, Pt = (0, U.useMemo)(() => G.find((e) => e.kind === "timestamp" && e.filterable !== !1 && e.timestamp?.autoRangeFilter !== !1) ?? null, [G]), J = (0, U.useMemo)(() => !p || y || k || D?.timeRange ? null : Pt, [
		p,
		k,
		D?.timeRange,
		y,
		Pt
	]), Ft = (0, U.useMemo)(() => G.filter((e) => e.filterable !== !1 && e.kind !== "timestamp"), [G]), It = (0, U.useMemo)(() => p ? Ft.map((e) => {
		let t = Qs(Ct, e);
		if (e.filterKind || t) return {
			column: e,
			kind: e.filterKind ?? "number",
			options: [],
			...t ? { numberBounds: t } : {}
		};
		let n = /* @__PURE__ */ new Set();
		for (let t of Ct) for (let r of Vt(t.row, e)) r && n.add(r);
		let r = Array.from(n).sort((e, t) => e.localeCompare(t, void 0, { numeric: !0 })).map((e) => ({
			value: e,
			label: e
		})), i = e.kind === "tags" ? 50 : 20, a = r.length >= 2 && r.length <= i;
		if (e.kind === "tags" && a) {
			let t = Ns(r, e.tags?.separator ?? "=");
			if (t.length >= 2) return {
				column: e,
				kind: "nested-multi",
				options: r,
				groups: t
			};
		}
		return {
			column: e,
			kind: a ? "multi" : "text",
			options: r
		};
	}) : [], [
		p,
		Ft,
		Ct
	]), Lt = (0, U.useMemo)(() => {
		let e = {};
		for (let t of It) {
			if (t.kind !== "multi" && t.kind !== "nested-multi") continue;
			let n = t.column.key;
			e[n] = mo(at[n] ?? {}, (e) => ot((t) => ws(t, n, e)));
		}
		return e;
	}, [It, at]);
	(0, U.useEffect)(() => {
		it((e) => Es(e, It)), ot((e) => Ds(e, It)), ct((e) => Os(e, It));
	}, [It]);
	let Rt = b !== void 0, Ht = Rt ? b : _t, Gt = Rt ? x ?? (() => {}) : vt, qt = (0, U.useMemo)(() => It.map((e) => {
		let t = e.column.key;
		if (e.kind === "multi") return {
			key: t,
			kind: "multi",
			label: Ut(e.column),
			value: at[t] ?? {},
			onChange: (e) => ot((n) => ws(n, t, e)),
			options: e.options
		};
		if (e.kind === "nested-multi") return {
			key: t,
			kind: "nested-multi",
			label: Ut(e.column),
			value: at[t] ?? {},
			onChange: (e) => ot((n) => ws(n, t, e)),
			groups: e.groups ?? []
		};
		if (e.kind === "number" || e.kind === "duration") {
			if (e.kind === "duration") return {
				key: t,
				kind: "duration",
				label: Ut(e.column),
				unit: e.column.filterUnit ?? "ms",
				value: lt[t] ?? {},
				onChange: (e) => ut(t, e)
			};
			let n = {
				key: t,
				kind: "number",
				label: Ut(e.column),
				value: lt[t] ?? {},
				onChange: (e) => ut(t, e),
				...e.column.numberUnit ? { unit: e.column.numberUnit } : {}
			};
			return e.numberBounds?.min !== void 0 && (n.domainMin = e.numberBounds.min), e.numberBounds?.max !== void 0 && (n.domainMax = e.numberBounds.max), e.numberBounds?.step !== void 0 && (n.step = e.numberBounds.step), n;
		}
		return {
			key: t,
			kind: "text",
			label: Ut(e.column),
			value: rt[t] ?? "",
			onChange: (e) => it((n) => Cs(n, t, e))
		};
	}), [
		It,
		at,
		lt,
		rt
	]), Xt = (0, U.useMemo)(() => new Map((j ?? []).map((e) => [e.key, e])), [j]), $t = (0, U.useMemo)(() => {
		let e = Pt?.filterKey, t = e ? Xt.get(e) : void 0;
		return t?.kind === "date-range" ? t : null;
	}, [Pt, Xt]), en = (0, U.useMemo)(() => {
		let e = /* @__PURE__ */ new Map();
		for (let t of G) {
			let n = qt.find((e) => e.key === t.key);
			if (n) {
				e.set(t.key, n);
				continue;
			}
			let r = t.filterKey ? Xt.get(t.filterKey) : void 0;
			r && r !== $t && e.set(t.key, r);
		}
		return e;
	}, [
		G,
		qt,
		Xt,
		$t
	]), tn = !!(D?.leading || D?.children || D?.trailing || D?.timeRange || D?.dateRange), nn = (0, U.useMemo)(() => {
		if (!J) return null;
		let e = {
			from: yt.from,
			to: yt.to,
			onApply: (e, t) => bt({
				from: e,
				to: t
			}),
			presets: J.timestamp?.rangePresets ?? ns,
			fromPlaceholder: "now-24h",
			toPlaceholder: "now",
			emptyLabel: "Any time"
		};
		return J.timestamp?.timeEnabled !== void 0 && (e.timeEnabled = J.timestamp.timeEnabled), J.timestamp?.timeZone && (e.timeZone = J.timestamp.timeZone), J.timestamp?.timeZones && (e.timeZones = J.timestamp.timeZones), e;
	}, [
		J,
		yt.from,
		yt.to
	]), rn = (0, U.useMemo)(() => $t && Pt ? {
		key: Pt.key,
		range: Kt($t)
	} : nn && J ? {
		key: J.key,
		range: nn
	} : null, [
		nn,
		Pt,
		$t,
		J
	]), an = k ?? rn?.range ?? null, on = (0, U.useMemo)(() => {
		let e = (j ?? []).filter((e) => e !== $t);
		return e.length > 0 ? [...e, ...qt] : qt;
	}, [
		j,
		qt,
		$t
	]), sn = v ?? p;
	Yt({
		autoFilter: p,
		manualFilter: y,
		showGlobalFilter: sn,
		globalFilterControlled: b !== void 0 || x !== void 0,
		hasExternalSearch: O !== void 0
	});
	let cn = sn || qt.length > 0 || !!an || !!O || j && j.length > 0 || tn || !!oe || Mt, ln = (e) => {
		Ye || Ze(e), je?.(e);
	}, un = Re && (en.size > 0 || !!rn), dn = (0, U.useMemo)(() => {
		if (y) return Ct;
		let e = Ht.trim().toLowerCase(), t = /* @__PURE__ */ new Date(), n = J ? A(yt.from, t) : null, r = J ? A(yt.to, t) : null;
		return Ct.filter(({ row: t }) => {
			if (J && (n || r)) {
				let e = ge(Bt(t, J));
				if (!e || n && e.getTime() < n.getTime() || r && e.getTime() > r.getTime()) return !1;
			}
			if (e && !Ft.flatMap((e) => Vt(t, e)).join(" ").toLowerCase().includes(e)) return !1;
			for (let e of It) {
				let n = Vt(t, e.column);
				if (e.kind === "text") {
					let t = (rt[e.column.key] ?? "").trim().toLowerCase();
					if (t && !n.some((e) => e.toLowerCase().includes(t))) return !1;
				} else if (e.kind === "number" || e.kind === "duration") {
					let n = lt[e.column.key] ?? {}, r = As(n.min, e.kind === "duration" ? n.minUnit : void 0, e.column.filterUnit), i = As(n.max, e.kind === "duration" ? n.maxUnit : void 0, e.column.filterUnit), a = String(n.min ?? "").trim() !== "", o = String(n.max ?? "").trim() !== "";
					if ((a || o) && !xs(t, e.column).some((e) => (r == null || (n.minOperator === ">" ? e > r : e >= r)) && (i == null || (n.maxOperator === "<" ? e < i : e <= i)))) return !1;
				} else {
					let t = at[e.column.key] ?? {}, r = Object.entries(t).filter(([, e]) => e === "include").map(([e]) => e), i = Object.entries(t).filter(([, e]) => e === "exclude").map(([e]) => e);
					if (r.length > 0 && !n.some((e) => r.includes(e)) || i.length > 0 && n.some((e) => i.includes(e))) return !1;
				}
			}
			return !0;
		});
	}, [
		Ht,
		Ft,
		It,
		y,
		at,
		lt,
		Ct,
		rt,
		J,
		yt.from,
		yt.to
	]), fn = (0, U.useMemo)(() => Object.fromEntries(G.map((e) => [e.key, (t) => bs(t.row, e)])), [G]), { renderLink: pn } = l(), { sorted: mn, sort: hn, toggle: gn } = Zt(dn, {
		defaultDir: C?.dir ?? "asc",
		manual: E,
		resolvers: fn,
		...w === void 0 ? {} : { sort: w },
		...T ? { onSortChange: T } : {},
		...C?.key ? { defaultKey: C.key } : {}
	}), _n = !!ee && !T && !(rs(ee, e.length) && !P?.hasMore), vn = mn, yn = Ui(P, e.length), bn = (0, U.useRef)(null), xn = !!P && P.hasMore, Sn = P?.loading ? "Loading more…" : "Scroll to load more…", Y = (0, U.useMemo)(() => new Set(F?.selectedRowIds ?? []), [F?.selectedRowIds]), Cn = (0, U.useMemo)(() => vn.filter((e) => F?.isRowSelectable?.(e.row) ?? !0), [F, vn]), wn = F?.selectAll !== !1, Tn = F?.selectAllPages;
	if (!wn && Tn) throw Error("DataTable rowSelection sets selectAll: false and selectAllPages; selecting across pages needs select-all");
	let En = (0, U.useMemo)(() => (Tn?.scopes ?? []).flatMap((e) => {
		let t = e.total ?? ee?.total;
		return t == null ? [] : [{
			...e,
			total: t
		}];
	}), [ee?.total, Tn?.scopes]), X = Tn?.selectedScopeId ? En.find((e) => e.id === Tn.selectedScopeId) : void 0;
	if (Tn?.selectedScopeId && !X) throw Error(`Selected table scope ${Tn.selectedScopeId} is not present in selectAllPages.scopes`);
	let Dn = X ? Cn.length : Cn.filter((e) => Y.has(e.id)).length, On = Cn.length > 0 && Dn === Cn.length, kn = !X && Dn > 0 && !On, An = (0, U.useRef)(null), jn = En[En.length - 1], Mn = On && !X ? En.find((e) => e.total > Y.size) : void 0, Nn = !!jn && (!!X || Y.size > 0);
	(0, U.useEffect)(() => {
		An.current && (An.current.indeterminate = kn);
	}, [kn]);
	let Pn = (0, U.useCallback)((e) => {
		F && F.onSelectionChange(Array.from(e), Ct.filter((t) => e.has(t.id)).map((e) => e.row));
	}, [F, Ct]), Fn = (0, U.useCallback)((e) => {
		if (!F || X || F.isRowSelectable?.(e.row) === !1) return;
		let t = new Set(Y);
		t.has(e.id) ? t.delete(e.id) : t.add(e.id), Pn(t);
	}, [
		Pn,
		F,
		Y,
		X
	]), In = (0, U.useCallback)(() => {
		if (!F || X) return;
		let e = new Set(Y);
		if (On) for (let t of Cn) e.delete(t.id);
		else for (let t of Cn) e.add(t.id);
		Pn(e);
	}, [
		On,
		Pn,
		F,
		Cn,
		Y,
		X
	]), Ln = (0, U.useMemo)(() => !F || !X && Y.size === 0 ? null : {
		selectedRowIds: Array.from(Y),
		selectedRows: X ? Cn.map((e) => e.row) : Ct.filter((e) => Y.has(e.id)).map((e) => e.row),
		...X ? { selectedScope: {
			id: X.id,
			total: X.total
		} } : {},
		clearSelection: () => Pn(/* @__PURE__ */ new Set())
	}, [
		Pn,
		F,
		Ct,
		Cn,
		Y,
		X
	]), Rn = (0, U.useMemo)(() => !Ln || typeof I != "function" ? null : I(Ln), [I, Ln]), zn = Array.isArray(I) ? I : null, Bn = t == null && L === "takeover" && !!Ln && !!zn, Vn = Ln && zn ? /* @__PURE__ */ (0, W.jsx)(Ho, {
		actions: zn,
		context: Ln,
		showCount: !Nn,
		className: "flex flex-wrap items-center gap-density-2"
	}) : Rn, Hn = t == null ? ie === void 0 ? ee ? null : r && mn.length === 0 ? a : `${mn.length} of ${e.length} row${e.length === 1 ? "" : "s"}` : typeof ie == "function" ? ie({
		visibleRowCount: mn.length,
		totalRowCount: e.length,
		loading: r
	}) : ie : null, [Un, Wn] = (0, U.useState)({}), Gn = K ? `mode:${K.value}` : "fixed", Kn = Un[Gn], qn = (0, U.useMemo)(() => {
		if (!Et) return null;
		let e = No(vn, (e) => Et.getGroupKey(e.row)).map((e) => ({
			key: e.key,
			records: e.rows,
			rows: e.rows.map((e) => e.row)
		}));
		return Et.compareGroups && e.sort((e, t) => Et.compareGroups(e, t)), e.map((e) => ({
			...e,
			label: Et.getGroupLabel?.(e.key, e.rows) ?? e.key,
			meta: Et.getGroupMeta?.(e.key, e.rows),
			collapsed: Kn?.groups[e.key] ?? Kn?.all ?? Po(e, Et.defaultCollapsed)
		}));
	}, [
		Kn,
		Et,
		vn
	]), Jn = !!qn?.length, Yn = Jn && qn.every((e) => !e.collapsed), Xn = Jn && qn.every((e) => e.collapsed), Zn = (e) => {
		Wn((t) => ({
			...t,
			[Gn]: {
				all: e,
				groups: {}
			}
		}));
	}, Qn = !Bn && t == null && (Nn || Vn) ? /* @__PURE__ */ (0, W.jsxs)("div", {
		"data-testid": "data-table-selection-actions",
		className: "flex flex-wrap items-center gap-density-2",
		children: [Nn && jn ? /* @__PURE__ */ (0, W.jsx)(us, {
			noun: Tn?.noun ?? "rows",
			...Tn?.loading ? { loading: !0 } : {},
			total: jn.total,
			selectedCount: X?.total ?? Y.size,
			...Mn ? { nextScope: Mn } : {},
			onClear: () => Pn(/* @__PURE__ */ new Set())
		}) : null, Vn]
	}) : null, $n = D?.trailing || oe || Mt ? /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
		D?.trailing,
		oe ? /* @__PURE__ */ (0, W.jsx)(Fo, {
			modes: oe,
			value: Tt,
			hasGroups: Jn,
			allExpanded: Yn,
			allCollapsed: Xn,
			onModeChange: (e) => {
				se === void 0 && gt(e), le?.(e);
			},
			onExpandAll: () => Zn(!1),
			onCollapseAll: () => Zn(!0)
		}) : null,
		Mt ? /* @__PURE__ */ (0, W.jsx)(_s, {
			opensSheet: nt,
			onOpen: (e) => $e(Oa(e))
		}) : null
	] }) : void 0, er = Qn || $n ? /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [Qn, $n] }) : void 0, tr = Bn ? /* @__PURE__ */ (0, W.jsxs)("div", {
		"data-testid": "data-table-selection-actions",
		className: "flex min-w-0 flex-1 flex-wrap items-center gap-density-2",
		children: [
			Nn && jn ? /* @__PURE__ */ (0, W.jsx)(us, {
				noun: Tn?.noun ?? "rows",
				...Tn?.loading ? { loading: !0 } : {},
				total: jn.total,
				selectedCount: X?.total ?? Y.size,
				...Mn ? { nextScope: Mn } : {},
				onClear: () => Pn(/* @__PURE__ */ new Set())
			}) : null,
			/* @__PURE__ */ (0, W.jsx)(Ho, {
				actions: zn,
				context: Ln,
				showCount: !Nn,
				className: "flex flex-wrap items-center gap-density-2"
			}),
			$n ? /* @__PURE__ */ (0, W.jsx)("div", {
				className: "ml-auto flex shrink-0 flex-wrap items-center gap-2",
				children: $n
			}) : null
		]
	}) : null, nr = (0, U.useMemo)(() => {
		let e = (e) => {
			let t = {
				kind: "row",
				record: e
			}, n = z ? z.includes(e.id) : dt[e.id];
			if (ye !== "row" || !he || !n) return [t];
			let r = he(e.row, {
				columns: G,
				visibleColumns: q,
				filterActionsByColumn: Lt
			});
			return r ? [t, {
				kind: "expansion",
				record: e,
				content: r
			}] : [t];
		};
		return qn ? qn.flatMap((t) => [{
			kind: "group",
			group: t
		}, ...t.collapsed ? [] : t.records.flatMap(e)]) : vn.flatMap(e);
	}, [
		qn,
		vn,
		ye,
		he,
		z,
		dt,
		G,
		q,
		Lt
	]), rr = (0, U.useCallback)((e) => {
		let t = nr[e];
		return t ? t.kind === "group" ? `group:${t.group.key}` : t.kind === "expansion" ? `expand:${t.record.id}` : `row:${t.record.id}` : `row:${e}`;
	}, [nr]), ir = (0, U.useMemo)(() => ga(u), [u]), Z = va({
		count: nr.length,
		options: ir,
		getItemKey: rr
	}), [ar, or] = (0, U.useState)({}), sr = (0, U.useRef)(null), [cr, lr] = (0, U.useState)(null), ur = B ? 1 : q.length + +!!F, dr = q.map((e) => e.key).join(" ");
	(0, U.useEffect)(() => {
		if (!ir || B) {
			or({});
			return;
		}
		let e = () => {
			let e = sr.current;
			if (!e) return;
			let t = +!!F, n = {};
			q.forEach((r, i) => {
				let a = e.cells[i + t]?.getBoundingClientRect().width ?? 0;
				a > 0 && (n[r.key] = a);
			}), Object.keys(n).length > 0 && or(n);
		}, t = requestAnimationFrame(e);
		if (typeof ResizeObserver > "u") return () => cancelAnimationFrame(t);
		let n = new ResizeObserver(e);
		return sr.current?.parentElement && n.observe(sr.current.parentElement), () => {
			cancelAnimationFrame(t), n.disconnect();
		};
	}, [
		ir,
		dr,
		F,
		ke,
		B
	]), (0, U.useEffect)(() => {
		if (!cr) return;
		let e = () => cr.style.setProperty("--dt-viewport-w", `${cr.clientWidth}px`);
		if (e(), typeof ResizeObserver > "u") return;
		let t = new ResizeObserver(e);
		return t.observe(cr), () => t.disconnect();
	}, [cr]);
	let fr = (0, U.useCallback)((e) => {
		Z.setScrollEl(e), lr(e), te && (te.current = e);
	}, [Z.setScrollEl, te]), pr = (0, U.useMemo)(() => Z.enabled ? Z.virtualItems.flatMap((e) => {
		let t = nr[e.index];
		return t ? [{
			item: t,
			index: e.index
		}] : [];
	}) : nr.map((e, t) => ({
		item: e,
		index: t
	})), [
		Z.enabled,
		Z.virtualItems,
		nr
	]), mr = (0, U.useCallback)((e) => {
		if (!F || X) return;
		let t = e.filter((e) => F.isRowSelectable?.(e.row) ?? !0), n = new Set(Y), r = t.length > 0 && t.every((e) => n.has(e.id));
		for (let e of t) r ? n.delete(e.id) : n.add(e.id);
		Pn(n);
	}, [
		Pn,
		F,
		Y,
		X
	]), hr = (0, U.useCallback)((e) => {
		bn.current && bn.current.disconnect(), e && xn && (bn.current = new IntersectionObserver((e) => {
			e[0]?.isIntersecting && yn();
		}, { rootMargin: "200px" }), bn.current.observe(e));
	}, [yn, xn]);
	if (B && (he || F)) throw Error("DataTable card rendering cannot be combined with row detail or row selection");
	let gr = (e, t) => {
		e.preventDefault(), e.stopPropagation();
		let n = e.clientX, r = (e.currentTarget.closest("th")?.getBoundingClientRect().width ?? 0) || We[t.key] || Xs(t), i = (e) => {
			let i = Zs(t, r + e.clientX - n);
			Ge((e) => ({
				...e,
				[t.key]: i
			}));
		}, a = () => {
			document.removeEventListener("mousemove", i), document.removeEventListener("mouseup", a), document.body.style.cursor = "", document.body.style.userSelect = "";
		};
		document.addEventListener("mousemove", i), document.addEventListener("mouseup", a), document.body.style.cursor = "col-resize", document.body.style.userSelect = "none";
	}, _r = (e, t) => {
		e.preventDefault(), e.stopPropagation();
		let n = e.currentTarget.closest("th"), r = e.currentTarget.closest("table");
		if (!n || !r) return;
		let i = Array.from(n.parentElement?.children ?? []).indexOf(n);
		if (i < 0) return;
		let a = Is(r, i, t);
		Ge((e) => ({
			...e,
			[t.key]: a
		}));
	}, vr = (e) => {
		Wt(e) && Je((t) => {
			let n = t[e.key] === !0;
			if (!n && Ot <= 1) return t;
			let r = { ...t };
			return n ? H.has(e.key) ? r[e.key] = !1 : delete r[e.key] : r[e.key] = !0, qs(r, V, H);
		});
	}, yr = () => Je(qs(Object.fromEntries([...H].map((e) => [e, !1])), V, H)), br = (e, t) => {
		kt && (e.preventDefault(), $e(Da(e, t.key)));
	}, xr = (e, t) => {
		e.preventDefault(), e.stopPropagation(), tt(Oa(e, t));
	};
	return /* @__PURE__ */ (0, W.jsx)(Qt, {
		density: Nt ?? "comfortable",
		children: /* @__PURE__ */ (0, W.jsxs)("div", {
			className: i("flex min-h-0 min-w-0 flex-1 flex-col gap-1", d),
			"data-density": Nt,
			children: [
				(cn || Qn || tr) && /* @__PURE__ */ (0, W.jsx)(jr, {
					...D,
					...tr ? { overlay: tr } : {},
					...O ? { search: O } : sn ? { search: {
						value: Ht,
						onChange: Gt,
						placeholder: S
					} } : {},
					...an ? { timeRange: an } : {},
					trailing: er,
					filters: on
				}),
				B && !_n && G.some((e) => e.sortable !== !1) && /* @__PURE__ */ (0, W.jsxs)("div", {
					className: "flex shrink-0 items-center gap-2 px-1 text-xs text-muted-foreground",
					children: [/* @__PURE__ */ (0, W.jsxs)("label", {
						className: "flex items-center gap-1.5",
						children: ["Sort by", /* @__PURE__ */ (0, W.jsxs)("select", {
							"aria-label": "Sort cards by",
							className: "h-7 rounded-md border border-input bg-background px-2 text-foreground",
							value: hn?.key ?? "",
							onChange: (e) => {
								let t = e.target.value;
								t ? T ? T({
									key: t,
									dir: "asc"
								}) : gn(t) : T?.(null);
							},
							children: [/* @__PURE__ */ (0, W.jsx)("option", {
								value: "",
								children: "Default order"
							}), G.filter((e) => e.sortable !== !1).map((e) => /* @__PURE__ */ (0, W.jsx)("option", {
								value: e.key,
								children: Ut(e)
							}, e.key))]
						})]
					}), hn && /* @__PURE__ */ (0, W.jsx)("button", {
						type: "button",
						className: "h-7 rounded-md border border-input px-2 text-foreground hover:bg-accent",
						onClick: () => {
							T ? T({
								key: hn.key,
								dir: hn.dir === "asc" ? "desc" : "asc"
							}) : gn(hn.key);
						},
						children: hn.dir === "asc" ? "Ascending" : "Descending"
					})]
				}),
				/* @__PURE__ */ (0, W.jsxs)("div", {
					className: "relative flex min-h-0 max-w-full flex-1 flex-col",
					children: [
						r && t == null ? /* @__PURE__ */ (0, W.jsx)(s, {
							"data-testid": "data-table-loading-bar",
							className: "rounded-t-md"
						}) : null,
						/* @__PURE__ */ (0, W.jsx)("div", {
							ref: fr,
							"data-testid": "data-table-scroll",
							className: i("min-h-0 max-w-full flex-1 overflow-auto overscroll-x-contain rounded-md border border-border bg-background", f),
							"aria-busy": r && t == null || void 0,
							children: /* @__PURE__ */ (0, W.jsxs)("table", {
								className: i(B ? "w-full text-left text-sm" : "w-max min-w-full text-left text-sm", B || Object.keys(ar).length > 0 ? "table-fixed" : "table-auto"),
								children: [
									B ? /* @__PURE__ */ (0, W.jsx)("colgroup", { children: /* @__PURE__ */ (0, W.jsx)("col", {}) }) : /* @__PURE__ */ (0, W.jsxs)("colgroup", { children: [F && t == null ? /* @__PURE__ */ (0, W.jsx)("col", { className: "w-10" }) : null, q.map((e) => /* @__PURE__ */ (0, W.jsx)("col", {
										style: t == null ? Fs(e, We, ar) : void 0,
										className: e.shrink && !e.grow && t == null ? "w-px" : void 0
									}, e.key))] }),
									B ? /* @__PURE__ */ (0, W.jsx)("thead", {
										className: "sr-only",
										children: /* @__PURE__ */ (0, W.jsx)("tr", { children: /* @__PURE__ */ (0, W.jsx)("th", {
											scope: "col",
											children: "Results"
										}) })
									}) : /* @__PURE__ */ (0, W.jsx)("thead", {
										className: "sticky top-0 z-10 bg-muted shadow-[0_1px_0_0_var(--tw-shadow-color)] shadow-border",
										children: /* @__PURE__ */ (0, W.jsxs)("tr", {
											ref: sr,
											className: "border-b border-border text-xs text-muted-foreground",
											children: [F && t == null ? /* @__PURE__ */ (0, W.jsx)("th", {
												className: i("w-10 text-center", es),
												children: wn ? /* @__PURE__ */ (0, W.jsx)("input", {
													ref: An,
													type: "checkbox",
													"aria-label": "Select all visible rows",
													checked: On,
													disabled: !!X || Cn.length === 0,
													onChange: In,
													className: "size-3.5 rounded border-border accent-primary"
												}) : null
											}) : null, q.map((e) => /* @__PURE__ */ (0, W.jsxs)("th", {
												className: i("group/header relative whitespace-nowrap font-medium", es, rc(e.align), Se && e.resizable !== !1 && "select-none", e.headerClassName),
												onContextMenu: (t) => br(t, e),
												children: [/* @__PURE__ */ (0, W.jsxs)("div", {
													className: i("flex min-w-0 items-center gap-1", ic(e.align), Se && e.resizable !== !1 && "pr-2"),
													children: [/* @__PURE__ */ (0, W.jsx)("span", {
														className: "min-w-0",
														children: e.sortable === !1 || _n ? /* @__PURE__ */ (0, W.jsx)("span", {
															..._n && e.sortable !== !1 ? { title: "Sorting is served by the query while this table is paged by the server" } : {},
															children: e.label
														}) : /* @__PURE__ */ (0, W.jsx)(Ma, {
															active: hn?.key === e.key,
															...hn?.key === e.key ? { dir: hn.dir } : {},
															...e.align ? { align: e.align } : {},
															onClick: () => gn(e.key),
															children: e.label
														})
													}), un && (en.has(e.key) || rn?.key === e.key) && /* @__PURE__ */ (0, W.jsx)(hs, {
														column: e,
														active: en.has(e.key) ? Jt(en.get(e.key)) : !!(rn?.range.from || rn?.range.to),
														onOpen: (t) => xr(t, e.key)
													})]
												}), Se && e.resizable !== !1 && /* @__PURE__ */ (0, W.jsx)("span", {
													role: "separator",
													"aria-label": `Resize ${Ut(e)} column`,
													"aria-orientation": "vertical",
													className: "absolute right-0 top-0 hidden h-full w-3 cursor-col-resize touch-none items-center justify-center border-r border-border/70 bg-gradient-to-l from-border/30 to-transparent transition-colors hover:border-primary hover:from-primary/20 md:flex",
													onClick: (e) => {
														e.preventDefault(), e.stopPropagation();
													},
													onDoubleClick: (t) => _r(t, e),
													onMouseDown: (t) => gr(t, e),
													children: /* @__PURE__ */ (0, W.jsx)("span", {
														"aria-hidden": !0,
														className: "h-4 w-0.5 rounded-full bg-border transition-colors group-hover/header:bg-primary/70"
													})
												})]
											}, e.key))]
										})
									}),
									/* @__PURE__ */ (0, W.jsxs)("tbody", {
										ref: Z.setTbodyEl,
										children: [t == null ? r && vn.length === 0 ? /* @__PURE__ */ (0, W.jsx)(ps, {
											columns: q,
											rowCount: o,
											message: a,
											selection: !!F
										}) : dn.length === 0 ? /* @__PURE__ */ (0, W.jsx)("tr", { children: /* @__PURE__ */ (0, W.jsx)("td", {
											colSpan: ur,
											className: "px-4 py-10 text-center text-sm text-muted-foreground",
											children: c
										}) }) : /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
											Z.enabled && Z.spacers.top > 0 && /* @__PURE__ */ (0, W.jsx)("tr", {
												"aria-hidden": !0,
												"data-virtual-spacer": "top",
												children: /* @__PURE__ */ (0, W.jsx)("td", {
													colSpan: ur,
													style: {
														height: Z.spacers.top,
														padding: 0,
														border: 0
													}
												})
											}),
											pr.map(({ item: e, index: t }) => {
												let n = Z.enabled ? {
													ref: Z.virtualizer.measureElement,
													"data-index": t
												} : {};
												if (e.kind === "expansion") return /* @__PURE__ */ (0, W.jsx)("tr", {
													...n,
													children: /* @__PURE__ */ (0, W.jsx)("td", {
														colSpan: ur,
														className: "bg-muted/40 p-density-3",
														children: /* @__PURE__ */ (0, W.jsx)("div", {
															"data-testid": "data-table-expansion",
															className: "sticky left-density-3 w-[calc(var(--dt-viewport-w)-2*var(--spacing-density-3))] overflow-x-auto rounded-md border border-border bg-background p-density-3",
															children: e.content
														})
													})
												}, `expand:${e.record.id}`);
												if (e.kind === "group") return /* @__PURE__ */ (0, W.jsx)(fs, {
													...Z.enabled ? {
														rowRef: Z.virtualizer.measureElement,
														dataIndex: t
													} : {},
													label: e.group.label,
													meta: e.group.meta,
													metaAlign: Et?.metaAlign ?? "end",
													...Et?.metaClassName ? { metaClassName: Et.metaClassName } : {},
													count: e.group.records.length,
													colSpan: ur,
													collapsed: e.group.collapsed,
													onToggleCollapsed: () => Wn((t) => ({
														...t,
														[Gn]: {
															...t[Gn],
															groups: {
																...t[Gn]?.groups,
																[e.group.key]: !e.group.collapsed
															}
														}
													})),
													...F && wn ? { selection: {
														selectableCount: e.group.records.filter((e) => F.isRowSelectable?.(e.row) ?? !0).length,
														selectedCount: X ? e.group.records.filter((e) => F.isRowSelectable?.(e.row) ?? !0).length : e.group.records.filter((e) => (F.isRowSelectable?.(e.row) ?? !0) && Y.has(e.id)).length,
														disabled: !!X,
														onToggle: () => mr(e.group.records)
													} } : {}
												}, `group:${e.group.key}`);
												let r = e.record;
												if (B) return /* @__PURE__ */ (0, W.jsx)("tr", {
													...n,
													className: "align-top",
													children: /* @__PURE__ */ (0, W.jsx)("td", {
														className: "border-b border-border/60 p-density-2",
														children: B(r.row)
													})
												}, r.id);
												let a = pe?.(r.row), o = z ? z.includes(r.id) : dt[r.id] ?? !1, s = (he?.(r.row, {
													columns: G,
													visibleColumns: q,
													filterActionsByColumn: Lt
												}) ?? null) !== null, c = s && ye === "row", l = s && ye === "dialog", u = fe?.(r.row) ?? !!ue, d = F?.isRowSelectable?.(r.row) ?? !0, f = !!F?.toggleOnRowClick && d && !X, p = !!a || u || s || f, m = () => {
													f && Fn(r), c && (z ? _e?.(o ? z.filter((e) => e !== r.id) : [...z, r.id]) : ft((e) => ({
														...e,
														[r.id]: !e[r.id]
													}))), l && mt(r), u && ue?.(r.row);
												};
												return /* @__PURE__ */ (0, W.jsx)(U.Fragment, { children: /* @__PURE__ */ (0, W.jsxs)("tr", {
													...n,
													"aria-expanded": c ? o : void 0,
													className: i("relative border-b border-border/60 align-top", p && "cursor-pointer hover:bg-accent/40", p && !a && "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring", (X && d || Y.has(r.id)) && "bg-accent/50", R?.(r.row)),
													tabIndex: p && !a ? 0 : void 0,
													onClick: m,
													onKeyDown: (e) => {
														a || e.target !== e.currentTarget || (e.key === "Enter" || e.key === " ") && (e.preventDefault(), m());
													},
													children: [F ? /* @__PURE__ */ (0, W.jsx)("td", {
														className: i("relative z-10 w-10 text-center", ts),
														onClick: (e) => e.stopPropagation(),
														children: /* @__PURE__ */ (0, W.jsx)("input", {
															type: "checkbox",
															"aria-label": `Select row ${r.id}`,
															checked: !!X && d || Y.has(r.id),
															disabled: !!X || !d,
															onChange: () => Fn(r),
															className: "size-3.5 rounded border-border accent-primary"
														})
													}) : null, q.map((e, t) => {
														let n = zt(r.row, e), o = e.render ? e.render(n, r.row) : Ms(n), s = os(n, r.row, e), c = ss(n, r.row, e), l = !!e.filterKey && !!N && s !== void 0;
														if (e.filterKey && N && s !== void 0) {
															let t = e.filterKey;
															o = /* @__PURE__ */ (0, W.jsx)(xo, {
																value: s,
																...c === void 0 ? {} : { displayValue: c },
																mode: M?.[t]?.[s],
																onChange: (e) => N({
																	key: t,
																	value: s,
																	mode: e
																}),
																children: o
															});
														}
														return e.kind === "tags" && Lt[e.key] && (o = /* @__PURE__ */ (0, W.jsx)(ho, {
															value: Lt[e.key],
															children: o
														})), /* @__PURE__ */ (0, W.jsx)("td", {
															className: i(ts, rc(e.align), e.cellClassName),
															style: oc(e, We[e.key]),
															children: /* @__PURE__ */ (0, W.jsx)(ys, {
																column: e,
																children: a && t === 0 ? pn({
																	to: a,
																	draggable: !1,
																	onDragStart: (e) => e.preventDefault(),
																	className: "hover:underline after:absolute after:inset-0 after:content-[''] [-webkit-user-drag:none]",
																	children: l ? /* @__PURE__ */ (0, W.jsx)("span", {
																		className: "relative z-10 inline-flex min-w-0 items-center",
																		children: o
																	}) : o
																}) : o
															})
														}, e.key);
													})]
												}) }, r.id);
											}),
											Z.enabled && Z.spacers.bottom > 0 && /* @__PURE__ */ (0, W.jsx)("tr", {
												"aria-hidden": !0,
												"data-virtual-spacer": "bottom",
												children: /* @__PURE__ */ (0, W.jsx)("td", {
													colSpan: ur,
													style: {
														height: Z.spacers.bottom,
														padding: 0,
														border: 0
													}
												})
											})
										] }) : /* @__PURE__ */ (0, W.jsx)(ms, {
											colSpan: ur,
											error: t
										}), t == null && xn && /* @__PURE__ */ (0, W.jsx)("tr", {
											ref: hr,
											"aria-hidden": !0,
											children: /* @__PURE__ */ (0, W.jsx)("td", {
												colSpan: ur,
												className: "p-density-2 text-center text-xs text-muted-foreground",
												children: Sn
											})
										})]
									})
								]
							})
						}),
						!B && /* @__PURE__ */ (0, W.jsx)(Wo, { scroll: cr })
					]
				}),
				Hn !== null && /* @__PURE__ */ (0, W.jsx)("div", {
					className: "shrink-0 px-1 text-xs text-muted-foreground",
					children: Hn
				}),
				t == null && ee ? /* @__PURE__ */ (0, W.jsx)(ds, {
					pagination: ee,
					visibleRowCount: mn.length,
					infinite: !!P
				}) : null,
				ye === "dialog" && he && /* @__PURE__ */ (0, W.jsx)(m, {
					open: pt !== null,
					onClose: () => mt(null),
					size: xe,
					...pt ? { title: typeof be == "function" ? be(pt.row) : be } : {},
					children: pt ? he(pt.row, {
						columns: G,
						visibleColumns: q,
						filterActionsByColumn: Lt
					}) : null
				}),
				Qe && Mt && nt && /* @__PURE__ */ (0, W.jsx)(ja, {
					open: !0,
					columns: G,
					hiddenColumns: qe,
					activeColumnKey: Qe.columnKey,
					actions: ze ?? [],
					showColumnVisibilityControl: kt,
					showDensityControl: At,
					showThemeControl: Le,
					themeMenuValue: Fe,
					densityOverride: Nt,
					visibleHideableColumnCount: Ot,
					onToggle: vr,
					onShowAll: yr,
					onDensityChange: ln,
					...Ie ? { onThemeChange: Ie } : {},
					onClose: () => $e(null)
				}),
				Qe && Mt && !nt && /* @__PURE__ */ (0, W.jsx)(vs, {
					columns: G,
					hiddenColumns: qe,
					anchor: Qe,
					actions: ze ?? [],
					showColumnVisibilityControl: kt,
					showDensityControl: At,
					showThemeControl: Le,
					themeMenuValue: Fe,
					densityOverride: Nt,
					visibleHideableColumnCount: Ot,
					onToggle: vr,
					onShowAll: yr,
					onDensityChange: ln,
					...Ie ? { onThemeChange: Ie } : {},
					onClose: () => $e(null)
				}),
				et && /* @__PURE__ */ (0, W.jsx)(gs, {
					filter: en.get(et.columnKey ?? ""),
					...rn && rn.key === et.columnKey ? { timeRange: rn.range } : {},
					anchor: et,
					onClose: () => tt(null)
				})
			]
		})
	});
}
function os(e, t, n) {
	let r = n.serverFilterValue ? n.serverFilterValue(e, t) : e;
	switch (typeof r) {
		case "string": return r || void 0;
		case "number":
		case "boolean": return String(r);
		default: return;
	}
}
function ss(e, t, n) {
	let r = n.serverFilterLabel ? n.serverFilterLabel(e, t) : e;
	switch (typeof r) {
		case "string": return r || void 0;
		case "number":
		case "boolean": return String(r);
		default: return;
	}
}
var cs = [
	5,
	10,
	25,
	50,
	100,
	200
];
function ls(e) {
	let [t, n] = (0, U.useState)([]), r = e?.current, i = r == null ? 0 : Math.max(t.indexOf(r) + 1, 1);
	return {
		page: i,
		push: (e) => n((t) => [...t.slice(0, i), e]),
		previous: () => i <= 1 ? void 0 : t[i - 2]
	};
}
function us({ noun: e, loading: t, total: n, selectedCount: r, nextScope: i, onClear: a }) {
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		role: "status",
		"data-testid": "data-table-selection-scope",
		className: "flex flex-wrap items-center gap-density-2 text-xs text-muted-foreground",
		children: [
			/* @__PURE__ */ (0, W.jsx)("span", { children: r >= n ? `All ${n.toLocaleString()} ${e} selected.` : `${r.toLocaleString()} of ${n.toLocaleString()} ${e} selected.` }),
			i ? /* @__PURE__ */ (0, W.jsx)("button", {
				type: "button",
				className: "rounded-md border border-border bg-background px-2 py-0.5 font-medium text-foreground hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50",
				disabled: t,
				onClick: i.onSelectAll,
				children: t ? "Selecting…" : i.label ?? `Select all ${i.total.toLocaleString()} ${e}`
			}) : null,
			/* @__PURE__ */ (0, W.jsx)("button", {
				type: "button",
				className: "rounded-md border border-border bg-background px-2 py-0.5 font-medium text-foreground hover:bg-accent",
				onClick: a,
				children: "Clear selection"
			})
		]
	});
}
function ds({ pagination: e, visibleRowCount: t, infinite: n = !1 }) {
	let { page: r, pageSize: a, total: o, totalRelation: s, cursor: c, onPageChange: l, onPageSizeChange: u } = e, d = Array.from(/* @__PURE__ */ new Set([...e.pageSizeOptions ?? cs, a])).sort((e, t) => e - t), f = ls(c), m = c ? f.page : Math.max(0, r), h = s === "gte", g = o != null && a > 0 && !h && !c ? Math.max(1, Math.ceil(o / a)) : void 0, _ = m === 0, v = c ? c.next == null : e.hasMore == null ? g == null ? t < a : m >= g - 1 : !e.hasMore, y = t > 0, b = o === 0 || !y ? 0 : n ? 1 : m * a + 1, x = o != null && y ? Math.min(h ? Infinity : o, b + Math.max(t - 1, 0)) : t, S = o == null ? `${t} row${t === 1 ? "" : "s"}` : `${b}-${x} of ${h ? `~${o}+` : o}`, C = !n && (c != null || l != null), w = () => {
		c ? c.onCursorChange(f.previous()) : l?.(Math.max(0, m - 1));
	}, T = () => {
		c ? c.next != null && (f.push(c.next), c.onCursorChange(c.next)) : l?.(m + 1);
	}, E = me(ve("sm"));
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "flex min-h-9 shrink-0 flex-wrap items-center justify-between gap-x-2 gap-y-1 border-t border-border/70 px-1 pt-2 text-xs text-muted-foreground max-sm:flex-nowrap sm:gap-3",
		children: [/* @__PURE__ */ (0, W.jsx)("div", {
			"aria-live": "polite",
			className: "min-w-0 truncate whitespace-nowrap",
			children: S
		}), /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "flex items-center gap-2 max-sm:flex-nowrap sm:flex-wrap sm:gap-3",
			children: [/* @__PURE__ */ (0, W.jsxs)("label", {
				className: "flex items-center gap-1.5",
				children: [/* @__PURE__ */ (0, W.jsx)("span", {
					className: "max-sm:sr-only",
					children: "Rows per page"
				}), /* @__PURE__ */ (0, W.jsx)("select", {
					className: "h-9 rounded-md border border-border bg-background px-2 text-xs text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:h-7",
					value: a,
					onChange: (e) => u(Number(e.target.value)),
					children: d.map((e) => /* @__PURE__ */ (0, W.jsx)("option", {
						value: e,
						children: e
					}, e))
				})]
			}), C ? /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
				/* @__PURE__ */ (0, W.jsx)("span", {
					className: i("text-center", !E && "min-w-20"),
					title: `Page ${m + 1}${g == null ? "" : ` of ${g}`}`,
					children: E ? `${m + 1}${g == null ? "" : `/${g}`}` : `Page ${m + 1}${g == null ? "" : ` of ${g}`}`
				}),
				/* @__PURE__ */ (0, W.jsx)("button", {
					type: "button",
					className: "inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-border bg-background text-foreground hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50 sm:h-7 sm:w-7",
					"aria-label": "Previous page",
					title: "Previous page",
					disabled: _,
					onClick: w,
					children: /* @__PURE__ */ (0, W.jsx)(p, { icon: ue })
				}),
				/* @__PURE__ */ (0, W.jsx)("button", {
					type: "button",
					className: "inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-border bg-background text-foreground hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50 sm:h-7 sm:w-7",
					"aria-label": "Next page",
					title: "Next page",
					disabled: v,
					onClick: T,
					children: /* @__PURE__ */ (0, W.jsx)(p, { icon: Ct })
				})
			] }) : null]
		})]
	});
}
function fs({ label: e, meta: t, metaAlign: n, metaClassName: r, count: a, colSpan: o, collapsed: s, onToggleCollapsed: c, selection: l, rowRef: u, dataIndex: d }) {
	let f = !!l && l.selectableCount > 0 && l.selectedCount === l.selectableCount, m = !!l && l.selectedCount > 0 && !f, h = (0, U.useRef)(null);
	return (0, U.useEffect)(() => {
		h.current && (h.current.indeterminate = m);
	}, [m]), /* @__PURE__ */ (0, W.jsx)("tr", {
		ref: u,
		"data-index": d,
		className: "border-b border-border/60 bg-muted/40",
		children: /* @__PURE__ */ (0, W.jsx)("td", {
			colSpan: o,
			className: "px-density-2 py-density-1",
			children: /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex items-center gap-2",
				children: [
					l ? /* @__PURE__ */ (0, W.jsx)("input", {
						ref: h,
						type: "checkbox",
						"aria-label": `Select group ${typeof e == "string" ? e : a}`,
						checked: f,
						disabled: l.disabled || l.selectableCount === 0,
						onChange: l.onToggle,
						className: "size-3.5 rounded border-border accent-primary"
					}) : null,
					/* @__PURE__ */ (0, W.jsxs)("button", {
						type: "button",
						"aria-expanded": !s,
						onClick: c,
						className: i("flex min-w-0 items-center gap-1.5 text-left text-sm font-semibold text-foreground hover:text-primary", n === "end" && "flex-1"),
						children: [
							/* @__PURE__ */ (0, W.jsx)(p, {
								icon: s ? j : z,
								className: "shrink-0 text-muted-foreground"
							}),
							/* @__PURE__ */ (0, W.jsx)("span", {
								className: "truncate",
								children: e
							}),
							/* @__PURE__ */ (0, W.jsx)("span", {
								className: "shrink-0 font-normal text-muted-foreground",
								children: a
							})
						]
					}),
					t ? /* @__PURE__ */ (0, W.jsx)("div", {
						className: i("shrink-0 text-sm text-muted-foreground", r),
						children: t
					}) : null
				]
			})
		})
	});
}
function ps({ columns: e, rowCount: t, message: n, selection: r }) {
	let a = Math.max(1, t), o = Math.max(1, e.length + +!!r);
	return /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [/* @__PURE__ */ (0, W.jsx)("tr", {
		className: "border-b border-border/60 bg-muted/20",
		children: /* @__PURE__ */ (0, W.jsxs)("td", {
			colSpan: o,
			className: i(ts, "p-0"),
			children: [/* @__PURE__ */ (0, W.jsx)("div", {
				className: "h-0.5 w-full overflow-hidden bg-muted",
				children: /* @__PURE__ */ (0, W.jsx)("div", { className: "h-full w-1/2 animate-pulse rounded-full bg-primary/70" })
			}), /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground",
				children: [/* @__PURE__ */ (0, W.jsx)("span", {
					"aria-hidden": !0,
					className: "h-3 w-3 animate-spin rounded-full border-2 border-primary/30 border-t-primary"
				}), /* @__PURE__ */ (0, W.jsx)("span", { children: n })]
			})]
		})
	}), Array.from({ length: a }).map((t, n) => /* @__PURE__ */ (0, W.jsxs)("tr", {
		className: "border-b border-border/60",
		children: [r ? /* @__PURE__ */ (0, W.jsx)("td", {
			className: i("w-10", ts),
			children: /* @__PURE__ */ (0, W.jsx)("span", { className: "mx-auto block size-3.5 animate-pulse rounded bg-muted" })
		}) : null, e.map((e, t) => /* @__PURE__ */ (0, W.jsx)("td", {
			className: i(ts, rc(e.align), e.cellClassName),
			style: oc(e, void 0),
			children: /* @__PURE__ */ (0, W.jsx)(ys, {
				column: e,
				children: /* @__PURE__ */ (0, W.jsx)("span", {
					"aria-hidden": !0,
					className: i("block h-3 animate-pulse rounded-full bg-muted", ac(t))
				})
			})
		}, e.key))]
	}, n))] });
}
function ms({ colSpan: e, error: t }) {
	return /* @__PURE__ */ (0, W.jsx)("tr", {
		className: "border-b border-border/60",
		children: /* @__PURE__ */ (0, W.jsx)("td", {
			colSpan: Math.max(1, e),
			className: i(ts, "p-density-3"),
			children: /* @__PURE__ */ (0, W.jsx)("div", {
				role: "alert",
				className: "w-0 min-w-full overflow-hidden",
				children: t
			})
		})
	});
}
function hs({ column: e, active: t, onOpen: n }) {
	return /* @__PURE__ */ (0, W.jsx)("button", {
		type: "button",
		"aria-label": `Open ${Ut(e)} column filter`,
		"aria-haspopup": "dialog",
		"aria-pressed": t,
		className: i("inline-flex h-5 w-5 shrink-0 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", t && "bg-accent text-foreground"),
		onClick: n,
		children: /* @__PURE__ */ (0, W.jsx)("span", {
			"data-filter-icon-state": t ? "filled" : "outline",
			className: "inline-flex",
			children: /* @__PURE__ */ (0, W.jsx)(p, {
				icon: t ? dt : ut,
				className: "text-xs"
			})
		})
	});
}
function gs({ filter: e, timeRange: t, anchor: n, onClose: r }) {
	let i = Aa(n);
	if (!e && !t) return null;
	let a = e?.label ?? "Time range", o = !!(t?.from || t?.to);
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		ref: i.ref,
		role: "dialog",
		"aria-label": `${a} column filter`,
		className: "z-50 max-w-[calc(100vw-1rem)] overflow-auto rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-lg shadow-black/5",
		style: i.style,
		onClick: (e) => e.stopPropagation(),
		onContextMenu: (e) => e.preventDefault(),
		children: [/* @__PURE__ */ (0, W.jsxs)("div", {
			className: "mb-2 flex items-center justify-between gap-3",
			children: [/* @__PURE__ */ (0, W.jsx)("div", {
				className: "truncate text-xs font-semibold uppercase tracking-wide text-muted-foreground",
				children: a
			}), /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex items-center gap-1",
				children: [
					e && /* @__PURE__ */ (0, W.jsx)("button", {
						type: "button",
						className: "rounded px-1.5 py-0.5 text-xs text-primary transition-colors hover:bg-accent focus:bg-accent focus:outline-none disabled:text-muted-foreground",
						onClick: () => qt(e),
						disabled: !Jt(e),
						children: "Clear all"
					}),
					t && /* @__PURE__ */ (0, W.jsx)("button", {
						type: "button",
						className: "rounded px-1.5 py-0.5 text-xs text-primary transition-colors hover:bg-accent focus:bg-accent focus:outline-none disabled:text-muted-foreground",
						onClick: () => t.onApply("", ""),
						disabled: !o,
						children: "Clear all"
					}),
					/* @__PURE__ */ (0, W.jsx)("button", {
						type: "button",
						"aria-label": "Close column filter",
						title: "Close",
						className: "inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none",
						onClick: r,
						children: /* @__PURE__ */ (0, W.jsx)(p, {
							icon: fe,
							className: "text-sm"
						})
					})
				]
			})]
		}), /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "flex flex-col gap-2",
			children: [e && /* @__PURE__ */ (0, W.jsx)(Mr, {
				filter: e,
				chrome: "embedded"
			}), t && /* @__PURE__ */ (0, W.jsx)(Zr, {
				kind: "time",
				label: "Time range",
				...t
			})]
		})]
	});
}
function _s({ opensSheet: e, onOpen: t }) {
	return /* @__PURE__ */ (0, W.jsx)("button", {
		type: "button",
		"aria-label": "Open column menu",
		"aria-haspopup": e ? "dialog" : "menu",
		className: "inline-flex h-8 w-8 max-sm:h-9 max-sm:w-9 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
		onClick: (e) => {
			e.stopPropagation(), t(e);
		},
		children: /* @__PURE__ */ (0, W.jsx)(p, {
			icon: k,
			className: "text-sm"
		})
	});
}
function vs({ columns: e, hiddenColumns: t, anchor: n, actions: r, showColumnVisibilityControl: a, showDensityControl: o, showThemeControl: s, themeMenuValue: c, densityOverride: l, visibleHideableColumnCount: u, onToggle: d, onShowAll: f, onDensityChange: m, onThemeChange: h, onClose: g }) {
	let _ = n.columnKey ? e.find((e) => e.key === n.columnKey) : void 0, v = _ && Wt(_) && u > 1, y = r.length > 0, b = Aa(n);
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		ref: b.ref,
		role: "menu",
		"aria-label": "Column menu",
		className: "z-50 min-w-[16rem] max-w-[calc(100vw-1rem)] overflow-auto rounded-md border border-border bg-popover p-1.5 text-popover-foreground shadow-lg shadow-black/5",
		style: b.style,
		onClick: (e) => e.stopPropagation(),
		onContextMenu: (e) => e.preventDefault(),
		children: [
			a && /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
				/* @__PURE__ */ (0, W.jsxs)("div", {
					className: "flex items-center justify-between gap-3 px-2 py-1.5 text-xs font-medium text-muted-foreground",
					children: [/* @__PURE__ */ (0, W.jsx)("span", { children: "Columns" }), /* @__PURE__ */ (0, W.jsx)("button", {
						type: "button",
						className: "rounded px-1.5 py-0.5 text-xs text-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none",
						onClick: f,
						children: "Show all"
					})]
				}),
				_ && /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [/* @__PURE__ */ (0, W.jsxs)("button", {
					type: "button",
					role: "menuitem",
					disabled: !v,
					className: i("flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", !v && "cursor-not-allowed opacity-50"),
					onClick: () => {
						v && (d(_), g());
					},
					children: [/* @__PURE__ */ (0, W.jsx)(p, {
						icon: ft,
						className: "text-sm text-muted-foreground"
					}), /* @__PURE__ */ (0, W.jsxs)("span", { children: ["Hide ", Ut(_)] })]
				}), /* @__PURE__ */ (0, W.jsx)("div", { className: "my-1 h-px bg-border" })] }),
				/* @__PURE__ */ (0, W.jsx)("div", {
					className: "max-h-72 overflow-auto",
					children: e.map((e) => {
						let n = Wt(e), r = t[e.key] !== !0, a = !n || r && u <= 1;
						return /* @__PURE__ */ (0, W.jsxs)("label", {
							className: i("flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors hover:bg-accent hover:text-accent-foreground", a && "cursor-not-allowed opacity-50"),
							children: [/* @__PURE__ */ (0, W.jsx)("input", {
								type: "checkbox",
								className: "h-4 w-4 rounded border-border",
								checked: r,
								disabled: a,
								onChange: () => d(e)
							}), /* @__PURE__ */ (0, W.jsx)("span", {
								className: "truncate",
								children: Ut(e)
							})]
						}, e.key);
					})
				})
			] }),
			o && /* @__PURE__ */ (0, W.jsx)(Ca, {
				densityOverride: l,
				separated: a,
				onDensityChange: m
			}),
			y && /* @__PURE__ */ (0, W.jsx)(xa, {
				actions: r,
				separated: a || o,
				onClose: g
			}),
			s && h && /* @__PURE__ */ (0, W.jsx)(Ea, {
				value: c ?? "system",
				separated: a || o || y,
				onChange: h
			})
		]
	});
}
function ys({ column: e, children: t }) {
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: i(sc(e), rc(e.align)),
		children: t
	});
}
function bs(e, t) {
	let n = zt(e, t);
	return t.sortValue ? t.sortValue(n, e) : n;
}
function xs(e, t) {
	return Ss(Bt(e, t));
}
function Ss(e) {
	if (Array.isArray(e)) return e.flatMap((e) => Ss(e));
	let t = tc(e);
	return t == null ? [] : [t];
}
function Cs(e, t, n) {
	let r = { ...e };
	return n.trim() ? r[t] = n : delete r[t], r;
}
function ws(e, t, n) {
	let r = { ...e };
	return Object.keys(n).length > 0 ? r[t] = n : delete r[t], r;
}
function Ts(e, t, n) {
	let r = { ...e }, i = String(n.min ?? "").trim() !== "", a = String(n.max ?? "").trim() !== "";
	return i || a ? r[t] = n : delete r[t], r;
}
function Es(e, t) {
	return js(e, t, "text", (e) => !String(e ?? "").trim());
}
function Ds(e, t) {
	let n = new Set(t.filter((e) => e.kind === "multi" || e.kind === "nested-multi").map((e) => e.column.key));
	return Object.fromEntries(Object.entries(e).filter(([e, t]) => n.has(e) && Object.keys(t).length > 0));
}
function Os(e, t) {
	let n = new Set(t.filter((e) => e.kind === "number" || e.kind === "duration").map((e) => e.column.key));
	return Object.fromEntries(Object.entries(e).filter(([e, t]) => n.has(e) && (String(t.min ?? "").trim() || String(t.max ?? "").trim())));
}
var ks = {
	ms: 1,
	s: 1e3,
	m: 6e4,
	h: 36e5
};
function As(e, t, n) {
	let r = tc(e);
	return r == null ? null : r * ks[t ?? n ?? "ms"] / ks[n ?? "ms"];
}
function js(e, t, n, r) {
	let i = new Set(t.filter((e) => e.kind === n).map((e) => e.column.key));
	return Object.fromEntries(Object.entries(e).filter(([e, t]) => i.has(e) && !r(t)));
}
function Ms(e) {
	return e == null || e === "" ? /* @__PURE__ */ (0, W.jsx)("span", {
		className: "text-muted-foreground",
		children: "—"
	}) : typeof e == "boolean" ? e ? "True" : "False" : typeof e == "object" ? JSON.stringify(e) : String(e);
}
function Ns(e, t) {
	let n = /* @__PURE__ */ new Map();
	for (let r of e) {
		let { key: e } = uo(r.value, t), i = n.get(e) ?? [];
		i.push(r), n.set(e, i);
	}
	return Array.from(n.entries()).sort(([e], [t]) => e === "" && t !== "" ? 1 : t === "" && e !== "" ? -1 : e.localeCompare(t)).map(([e, t]) => ({
		groupKey: e,
		label: e === "" ? "Other" : e,
		options: t
	}));
}
function Ps(e, t) {
	if (!e.kind) return e;
	if (e.kind === "timestamp") {
		let n = t ?? "iso";
		return {
			...e,
			render: e.render ?? ((t) => /* @__PURE__ */ (0, W.jsx)(Na, {
				value: t,
				format: n,
				showTitleOnHover: e.timestamp?.alwaysShowFullOnHover !== !1
			})),
			sortValue: e.sortValue ?? ((e) => ge(e)?.getTime() ?? null),
			filterValue: e.filterValue ?? ((e) => {
				let t = ge(e);
				return t ? t.toISOString() : "";
			})
		};
	}
	if (e.kind === "tags") {
		let t = e.tags, n = t?.separator ?? "=";
		return {
			...e,
			render: e.render ?? ((e) => /* @__PURE__ */ (0, W.jsx)(yo, {
				tags: so(e, n),
				maxVisible: t?.maxVisible ?? 3
			})),
			filterValue: e.filterValue ?? ((e) => lo(e, n)),
			sortValue: e.sortValue ?? ((e) => lo(e, n).length)
		};
	}
	if (e.kind === "status") {
		let t = e.status, n = t?.map ?? ((e) => ko(e));
		return {
			...e,
			render: e.render ?? ((e, r) => {
				let i = n(e, r);
				if (!i) return /* @__PURE__ */ (0, W.jsx)("span", {
					className: "text-muted-foreground",
					children: "—"
				});
				let a = t?.showLabel ? typeof e == "string" ? e : i : void 0, o = t?.title?.(e, r) ?? (typeof e == "string" ? e : void 0);
				return /* @__PURE__ */ (0, W.jsx)(To, {
					status: i,
					...a ? { label: a } : {},
					...o ? { title: o } : {}
				});
			}),
			filterValue: e.filterValue ?? ((e, t) => n(e, t) ?? ""),
			sortValue: e.sortValue ?? ((e, t) => n(e, t) ?? "")
		};
	}
	return e;
}
function Fs(e, t, n = {}) {
	let r = t[e.key] ?? n[e.key];
	return r ? { width: `${r}px` } : void 0;
}
function Is(e, t, n) {
	return Zs(n, Array.from(e.querySelectorAll("th, td")).filter((e) => e instanceof HTMLTableCellElement && e.cellIndex === t && e.colSpan === 1).reduce((e, t) => {
		let n = Ls(t), r = Array.from(t.children).map((e) => e instanceof HTMLElement ? e.scrollWidth + n : 0);
		return Math.max(e, t.scrollWidth, ...r);
	}, 0) || Xs(n));
}
function Ls(e) {
	if (typeof window > "u") return 0;
	let t = window.getComputedStyle(e);
	return Rs(t.paddingLeft) + Rs(t.paddingRight);
}
function Rs(e) {
	let t = Number.parseFloat(e);
	return Number.isFinite(t) ? t : 0;
}
function zs(e, t) {
	if (typeof window > "u") return {};
	try {
		let n = window.localStorage.getItem(e);
		if (!n) return {};
		let r = JSON.parse(n);
		return !r || typeof r != "object" || Array.isArray(r) ? {} : Ks(r, t);
	} catch {
		return {};
	}
}
function Bs(e, t) {
	if (typeof window < "u") try {
		Object.keys(t).length === 0 ? window.localStorage.removeItem(e) : window.localStorage.setItem(e, JSON.stringify(t));
	} catch {}
}
function Vs(e, t, n) {
	let r = qs({}, t, n);
	if (typeof window > "u") return r;
	try {
		let i = window.localStorage.getItem(e);
		if (!i) return r;
		let a = JSON.parse(i);
		return !a || typeof a != "object" || Array.isArray(a) ? r : qs(a, t, n);
	} catch {
		return r;
	}
}
function Hs(e, t, n) {
	if (typeof window > "u") return;
	let r = Object.fromEntries(Object.entries(t).filter(([e, t]) => t !== n.has(e)));
	try {
		Object.keys(r).length === 0 ? window.localStorage.removeItem(e) : window.localStorage.setItem(e, JSON.stringify(r));
	} catch {}
}
function Us(e) {
	if (typeof window < "u") try {
		let t = window.localStorage.getItem(e);
		return Gs(t) ? t : void 0;
	} catch {
		return;
	}
}
function Ws(e, t) {
	if (typeof window < "u") try {
		t ? window.localStorage.setItem(e, t) : window.localStorage.removeItem(e);
	} catch {}
}
function Gs(e) {
	return e === "compact" || e === "comfortable" || e === "spacious";
}
function Ks(e, t) {
	let n = new Map(t.map((e) => [e.key, e])), r = {};
	for (let [t, i] of Object.entries(e)) {
		let e = n.get(t);
		e && typeof i == "number" && Number.isFinite(i) && (r[t] = Zs(e, i));
	}
	return r;
}
function qs(e, t, n) {
	let r = t.filter(Wt), i = new Set(r.map((e) => e.key)), a = {};
	for (let e of n) i.has(e) && (a[e] = !0);
	for (let [t, r] of Object.entries(e)) i.has(t) && (r === !0 ? a[t] = !0 : r === !1 && n.has(t) && (a[t] = !1));
	let o = r.filter((e) => a[e.key] !== !0).length;
	if (r.length > 0 && o === 0) {
		let e = r[0].key;
		n.has(e) ? a[e] = !1 : delete a[e];
	}
	return a;
}
function Js(e, t) {
	let n = Object.keys(e), r = Object.keys(t);
	return n.length === r.length && n.every((n) => Object.prototype.hasOwnProperty.call(t, n) && e[n] === t[n]);
}
function Ys(e, t) {
	let n = Object.keys(e), r = Object.keys(t);
	return n.length === r.length && n.every((n) => Object.prototype.hasOwnProperty.call(t, n) && e[n] === t[n]);
}
function Xs(e) {
	return e.grow ? qo : e.shrink ? Jo : Ko;
}
function Zs(e, t) {
	let n = e.minWidth ?? Go, r = e.maxWidth, i = Number.isFinite(n) ? Math.max(1, n) : Go;
	return Math.round(Math.max(i, Math.min(r != null && Number.isFinite(r) ? Math.max(i, r) : Infinity, t)));
}
function Qs(e, t) {
	let n = e.flatMap((e) => $s(Bt(e.row, t))).filter(ec), r = n.map((e) => tc(e)).filter((e) => e != null);
	return n.length === 0 || r.length !== n.length ? null : {
		min: Math.min(...r),
		max: Math.max(...r),
		step: nc(r)
	};
}
function $s(e) {
	return Array.isArray(e) ? e.flatMap((e) => $s(e)) : [e];
}
function ec(e) {
	return e != null && (typeof e != "string" || e.trim() !== "");
}
function tc(e) {
	if (typeof e == "number") return Number.isFinite(e) ? e : null;
	if (typeof e != "string") return null;
	let t = e.trim();
	if (!t) return null;
	let n = Number(t);
	return Number.isFinite(n) ? n : null;
}
function nc(e) {
	if (e.every(Number.isInteger)) return 1;
	let t = e.reduce((e, t) => {
		let [, n = ""] = String(t).split(".");
		return Math.max(e, n.length);
	}, 0);
	return 10 ** -Math.min(t || 2, 6);
}
function rc(e) {
	return e === "right" ? "text-right" : e === "center" ? "text-center" : "text-left";
}
function ic(e) {
	return e === "right" ? "justify-end" : e === "center" ? "justify-center" : "justify-start";
}
function ac(e) {
	return [
		"w-24",
		"w-40",
		"w-28",
		"w-56",
		"w-32",
		"w-48"
	][e % 6];
}
function oc(e, t) {
	return t === void 0 ? e.grow ? {
		maxWidth: "0px",
		minWidth: e.minWidth ?? qo
	} : { maxWidth: e.maxWidth ?? (e.shrink ? Xo : Yo) } : { maxWidth: `${t}px` };
}
function sc(e) {
	return i("w-full min-w-0 truncate", e.shrink && "whitespace-nowrap");
}
function cc({ theme: e = "system", onThemeChange: t, showThemeControl: n = !1, showFullscreenControl: r = !1, fullscreenTitle: a = "Table", fullscreenButtonLabel: o = "Open table full screen", className: s, filterBarProps: c, scrollContainerRef: l, ...u }) {
	let d = t !== void 0, [f, p] = (0, U.useState)(e), h = d ? e : f, g = C(h), _ = (e) => {
		d || p(e), t?.(e);
	}, [v, y] = (0, U.useState)(!1), b = ({ inFullscreen: e }) => {
		let t = r && !e, i = c?.trailing, a = t ? {
			...c,
			trailing: /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [i, /* @__PURE__ */ (0, W.jsx)(lc, {
				label: o,
				onClick: () => y(!0)
			})] })
		} : c, s = e === v;
		return /* @__PURE__ */ (0, W.jsx)(as, {
			...u,
			...s && l ? { scrollContainerRef: l } : {},
			...a ? { filterBarProps: a } : {},
			themeMenuValue: h,
			showThemeControl: n,
			onThemeChange: _
		});
	};
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		"data-theme": g,
		className: i("flex min-h-0 flex-1 flex-col text-foreground", s),
		children: [b({ inFullscreen: !1 }), r && /* @__PURE__ */ (0, W.jsx)(m, {
			open: v,
			onClose: () => y(!1),
			title: a,
			size: "full",
			className: "h-[95vh]",
			children: /* @__PURE__ */ (0, W.jsx)("div", {
				"data-theme": g,
				className: "flex h-full min-h-0 flex-col bg-background text-foreground",
				children: b({ inFullscreen: !0 })
			})
		})]
	});
}
function lc({ label: e, onClick: t }) {
	return /* @__PURE__ */ (0, W.jsx)("button", {
		type: "button",
		"aria-label": e,
		title: e,
		className: "inline-flex h-8 w-8 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
		onClick: t,
		children: /* @__PURE__ */ (0, W.jsx)(p, {
			icon: te,
			className: "text-sm"
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/LogViewer.tsx
var uc = {
	muted: {
		bg: "bg-muted",
		border: "border-border",
		text: "text-muted-foreground"
	},
	terminal: {
		bg: "bg-zinc-950",
		border: "border-zinc-800",
		text: "text-zinc-100"
	}
};
function dc({ logs: e, collapsedLines: t = 5, maxHeight: n, maxExpandedVh: r = 70, variant: a = "muted", bgClass: o, borderClass: s, className: c }) {
	let [l, u] = (0, U.useState)(!1), d = (0, U.useMemo)(() => e.split("\n"), [e]), f = d.length > t, p = uc[a], m = n ?? `${t * 1.4}em`, h = l || n ? e : d.slice(0, t).join("\n");
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i("relative", c),
		children: [/* @__PURE__ */ (0, W.jsx)("pre", {
			className: i("mt-0.5 text-[11px] rounded p-1.5 whitespace-pre-wrap overflow-y-auto border transition-all duration-200", p.text, o ?? p.bg, s ?? p.border),
			style: { maxHeight: l ? `${r}vh` : m },
			"data-log-variant": a,
			children: /* @__PURE__ */ (0, W.jsx)(lt, {
				as: "span",
				text: h
			})
		}), f && /* @__PURE__ */ (0, W.jsx)("button", {
			type: "button",
			className: i("text-[10px] mt-0.5", l ? "text-muted-foreground" : "text-primary hover:opacity-80"),
			onClick: () => u((e) => !e),
			children: l ? `▲ Collapse (${d.length} lines)` : `▼ Show more (${d.length} lines)`
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/lib/string.ts
function fc(e) {
	let t = e.length;
	for (; t > 0 && e[t - 1] === "/";) t--;
	return e.slice(0, t);
}
function pc(e, t) {
	let n = 0, r = e.length;
	for (; n < r && t.includes(e[n]);) n++;
	for (; r > n && t.includes(e[r - 1]);) r--;
	return e.slice(n, r);
}
function mc(e, t = " ") {
	let n = "", r = 0;
	for (; r < e.length;) {
		let i = e.indexOf("<", r);
		if (i === -1) return n + e.slice(r);
		let a = e.indexOf(">", i + 1);
		if (a === -1) return n + e.slice(r);
		a === i + 1 ? (n += e.slice(r, a + 1), r = a + 1) : (n += e.slice(r, i) + t, r = a + 1);
	}
	return n;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/frame-heuristics.ts
var hc = [
	"java.",
	"javax.",
	"sun.",
	"jdk.",
	"com.sun.",
	"oracle.jrockit."
];
function gc(e, t = []) {
	return hc.some((t) => e.startsWith(t)) || t.some((t) => e.startsWith(t));
}
function _c(e, t) {
	return [e.split(".").filter(Boolean).pop() ?? e, t].filter(Boolean).join(".");
}
function vc(e, t) {
	if (e) return t === void 0 ? e : `${e}:${t}`;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/FrameSourceWindow.tsx
function yc({ frame: e, className: t, focalTone: n = "error", markers: r }) {
	let a = (0, U.useMemo)(() => e.sourceLines ?? [], [e.sourceLines]), o = e.sourceStartLine ?? 0, s = e.sourceLineNumbers, c = e.line ?? -1, l = e.sourceLanguage, [u, d] = (0, U.useState)(null);
	if ((0, U.useEffect)(() => {
		if (a.length === 0) {
			d(null);
			return;
		}
		let e = !1;
		return D(a.join("\n"), { lang: l }).then((t) => {
			e || d(t);
		}), () => {
			e = !0;
		};
	}, [a, l]), a.length === 0) return null;
	let f = "bg-primary/10 font-semibold text-foreground", p = {
		neutral: f,
		error: "bg-red-500/10 font-semibold text-red-800 dark:text-red-300",
		warning: "bg-amber-500/10 font-semibold text-foreground",
		info: "bg-sky-500/10 font-semibold text-foreground"
	}[n], m = new Map(r?.map((e) => [e.line, e.content]));
	return /* @__PURE__ */ (0, W.jsx)("div", {
		"data-scroll-x": !0,
		className: i("mx-3 mb-2 ml-12 overflow-x-auto max-sm:ml-6 rounded border border-border/60 bg-muted/30 py-1 font-mono text-[11px] leading-relaxed", t),
		children: a.map((e, t) => {
			let r = s?.[t] ?? o + t, i = s === void 0 || r > 0, a = i && r === c, l = i ? m.get(r) : void 0;
			return /* @__PURE__ */ (0, W.jsxs)("div", {
				"data-source-line": i ? r : void 0,
				"data-source-tone": a ? n : void 0,
				"data-marked": l === void 0 ? void 0 : "true",
				className: ["grid min-w-max grid-cols-[3.5rem_minmax(24rem,1fr)] gap-3 px-2", a ? p : l === void 0 ? "text-foreground" : f].join(" "),
				children: [
					/* @__PURE__ */ (0, W.jsxs)("span", {
						className: "select-none text-right text-muted-foreground",
						children: [a || l !== void 0 ? ">" : "", i ? r : ""]
					}),
					/* @__PURE__ */ (0, W.jsx)("code", {
						className: "whitespace-pre",
						children: e ? /* @__PURE__ */ (0, W.jsx)(g, {
							tokens: u?.[t],
							content: e
						}) : " "
					}),
					l === void 0 ? null : /* @__PURE__ */ (0, W.jsx)("span", {
						"data-source-marker": !0,
						className: "col-start-2 -mt-0.5 mb-0.5 font-sans font-normal",
						children: l
					})
				]
			}, `${r}-${t}`);
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/FrameSourceWindow.utils.ts
function bc(e) {
	return !!e.sourceLines && e.sourceLines.length > 0;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/StackFrameRow.tsx
function xc({ frame: e, index: t, showIndex: n = !1, frameActions: r, renderFramePart: i, renderFrameDetail: a, frameActionsVisibility: o = "hover" }) {
	let s = e.kind !== "frame", c = e.displayName || e.method || e.functionName, l = r?.(e, t), u = s ? null : a?.(e, t);
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: [
			"group grid gap-x-2 gap-y-0.5 px-3 py-0.5 text-xs",
			n ? "grid-cols-[2rem_minmax(0,1fr)_auto]" : "grid-cols-[1.25rem_minmax(0,1fr)_auto]",
			e.runtime ? "text-muted-foreground" : "text-foreground"
		].join(" "),
		children: [
			/* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex items-start justify-end gap-1 pt-0.5 font-mono text-[10px] text-muted-foreground",
				children: [n && /* @__PURE__ */ (0, W.jsx)("span", {
					"data-frame-index": !0,
					children: t + 1
				}), /* @__PURE__ */ (0, W.jsx)(p, {
					icon: Sc(e),
					className: "mt-px shrink-0 text-[11px]"
				})]
			}),
			/* @__PURE__ */ (0, W.jsx)("div", {
				className: "min-w-0",
				children: /* @__PURE__ */ (0, W.jsx)("div", {
					className: "flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-0.5",
					children: s ? /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [/* @__PURE__ */ (0, W.jsx)("span", {
						className: "min-w-0 break-all font-mono font-semibold leading-4 opacity-70",
						children: e.functionName
					}), e.annotationText && /* @__PURE__ */ (0, W.jsx)("span", {
						className: "min-w-0 break-all font-mono text-[11px] font-normal opacity-80",
						children: e.annotationText
					})] }) : /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
						i ? i("method", e, t) : /* @__PURE__ */ (0, W.jsx)("span", {
							className: "min-w-0 break-all font-mono font-semibold leading-4",
							children: c
						}),
						e.class && (i ? i("class", e, t) : /* @__PURE__ */ (0, W.jsx)("span", {
							className: "min-w-0 break-all font-mono text-[11px] text-muted-foreground",
							children: e.class
						})),
						e.location && (i ? i("location", e, t) : /* @__PURE__ */ (0, W.jsx)("span", {
							className: "rounded border border-border bg-muted/40 px-1.5 py-px font-mono text-[10px] text-muted-foreground",
							children: e.location
						}))
					] })
				})
			}),
			l ? /* @__PURE__ */ (0, W.jsx)("div", {
				className: ["flex shrink-0 items-start gap-1", o === "hover" ? "opacity-0 transition-opacity focus-within:opacity-100 group-hover:opacity-100 [@media(hover:none)]:opacity-100" : "opacity-100"].join(" "),
				children: l
			}) : /* @__PURE__ */ (0, W.jsx)("div", {}),
			!s && bc(e) ? /* @__PURE__ */ (0, W.jsx)("div", {
				className: "col-span-3",
				children: /* @__PURE__ */ (0, W.jsx)(yc, { frame: e })
			}) : null,
			u ? /* @__PURE__ */ (0, W.jsx)("div", {
				className: "col-start-2 col-span-2 min-w-0",
				children: u
			}) : null
		]
	});
}
function Sc(e) {
	switch (e.kind) {
		case "locked": return M;
		case "waiting_to_lock": return xt;
		case "waiting_on":
		case "parking": return Ft;
	}
	return e.nativeMethod ? Nt : e.runtime ? _t : e.method || e.class ? q : gt;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/lib/color.ts
var Cc = /^(?:bg|text|border|fill|stroke)-/;
function wc(e) {
	return Cc.test(e.trim());
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/ProgressBar.tsx
function Tc({ segments: e, total: t, label: n, formatValue: r = String, height: a = "h-2", className: o }) {
	if (t === 0) return null;
	let s = e.filter((e) => e.count > 0).map((e) => `${r(e.count)} ${e.label}`).join(", ");
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: i("w-full bg-muted rounded-full flex overflow-hidden", a, o),
		title: s,
		role: "progressbar",
		"aria-label": n,
		"aria-valuetext": `${r(e.reduce((e, t) => e + t.count, 0))} of ${r(t)}. ${s}`,
		"aria-valuemin": 0,
		"aria-valuemax": t,
		"aria-valuenow": e.reduce((e, t) => e + t.count, 0),
		children: e.map((e, n) => {
			if (e.count === 0) return null;
			let o = e.count / t * 100, s = wc(e.color);
			return /* @__PURE__ */ (0, W.jsx)("div", {
				className: i(s ? e.color : void 0, a, "transition-all duration-300"),
				style: {
					width: `${o}%`,
					...s ? {} : { backgroundColor: e.color }
				},
				title: `${r(e.count)} ${e.label}`
			}, n);
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/MetricTile.tsx
var Ec = {
	neutral: "bg-muted-foreground/40",
	success: "bg-green-500",
	warning: "bg-yellow-500",
	danger: "bg-red-500",
	info: "bg-blue-500"
}, Dc = {
	neutral: "text-foreground",
	success: "text-green-600 dark:text-green-400",
	warning: "text-yellow-600 dark:text-yellow-400",
	danger: "text-red-600 dark:text-red-400",
	info: "text-blue-600 dark:text-blue-400"
};
function Oc({ label: e, value: t, barPct: n, tone: r = "neutral", className: a }) {
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i("flex min-w-0 flex-col gap-0.5 rounded-md border border-border bg-background px-density-2 py-1", a),
		"data-metric": e,
		children: [
			/* @__PURE__ */ (0, W.jsx)("span", {
				className: "text-[9px] uppercase leading-none tracking-wide text-muted-foreground",
				children: e
			}),
			/* @__PURE__ */ (0, W.jsx)("span", {
				className: i("truncate text-xs font-semibold tabular-nums", Dc[r]),
				children: t
			}),
			n !== void 0 && /* @__PURE__ */ (0, W.jsx)("div", {
				className: "h-0.5 w-full overflow-hidden rounded-full bg-muted",
				children: /* @__PURE__ */ (0, W.jsx)("div", {
					className: i("h-full transition-all duration-300", Ec[r]),
					style: { width: `${n}%` },
					"data-metric-bar": ""
				})
			})
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/OpenInFilesLink.tsx
function kc({ target: e, onOpenFile: t, className: n }) {
	return /* @__PURE__ */ (0, W.jsxs)(a, {
		type: "button",
		variant: "link",
		size: "sm",
		className: i("h-auto gap-1 p-0 text-xs", n),
		title: `Open ${e.file}${e.line ? `:${e.line}` : ""} in Files`,
		onClick: (n) => {
			n.stopPropagation(), t(e);
		},
		children: [/* @__PURE__ */ (0, W.jsx)(p, {
			icon: wt,
			className: "text-xs"
		}), "Open in Files"]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/copy.ts
var Ac = /* @__PURE__ */ new Set(["stdout", "stderr"]);
function jc(e, t) {
	return t ? [{
		label: e,
		value: t
	}] : [];
}
function Mc(e, t, n) {
	return t ? [n ? {
		title: e,
		lang: n,
		body: t
	} : {
		title: e,
		body: t
	}] : [];
}
function Nc(e) {
	if (e.file) return e.line ? `${e.file}:${e.line}` : e.file;
}
function Pc(e) {
	let t = e.failure_detail?.summary;
	return [e.message, t === e.message ? void 0 : t].filter(Boolean).join("\n");
}
function Fc(e) {
	let t = e.failure_detail;
	return t ? [
		t.matcher && `matcher: ${t.matcher}`,
		t.expected && `expected: ${t.expected}`,
		t.actual && `actual: ${t.actual}`,
		t.location && `at: ${t.location}`
	].filter(Boolean).join("\n") : "";
}
function Ic(e) {
	let t = e.duration ? Ye(e.duration) : void 0;
	return {
		facts: [
			...jc("Result", Ze(e) ?? void 0),
			...jc("Tool", e.framework),
			...jc("Location", Nc(e)),
			...jc("Duration", t)
		],
		brief: [
			...Mc("Error", Pc(e)),
			...Mc("Expectation", Fc(e)),
			...Mc("Stack", e.failure_detail?.stack),
			...Mc("stdout", e.stdout),
			...Mc("stderr", e.stderr)
		],
		items: []
	};
}
function Lc(e, t) {
	let n = /* @__PURE__ */ new Set();
	return e.filter((e) => {
		let r = t(e);
		return !n.has(r) && (n.add(r), !0);
	});
}
function Rc(e) {
	return {
		facts: Lc(e.flatMap((e) => e.facts ?? []), (e) => e.label),
		brief: Lc(e.flatMap((e) => e.brief ?? []), (e) => e.title),
		items: Lc(e.flatMap((e) => e.items ?? []), (e) => e.label)
	};
}
function zc(e) {
	let t = Math.max(0, ...(e.match(/`+/g) ?? []).map((e) => e.length));
	return "`".repeat(Math.max(3, t + 1));
}
function Bc(e, t, n) {
	let r = [`# ${e.name}`];
	n.length > 0 && r.push("", ...n.map((e) => `- **${e.label}:** ${e.value}`));
	for (let e of t) {
		if (!e.body) continue;
		if (r.push("", `## ${e.title}`, ""), e.markdown) {
			r.push(e.body);
			continue;
		}
		let t = zc(e.body);
		r.push(`${t}${e.lang ?? ""}`, e.body, t);
	}
	return r.join("\n");
}
function Vc(e, t) {
	let n = Ic(e), r = Rc([
		{
			facts: n.facts,
			brief: n.brief.filter((e) => !Ac.has(e.title))
		},
		t,
		{ brief: n.brief.filter((e) => Ac.has(e.title)) }
	]);
	return Bc(e, r.brief, r.facts);
}
function Hc(e) {
	let { children: t, ...n } = e;
	return JSON.stringify(n, null, 2);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/attempts.ts
function Uc(e) {
	return e.timed_out ? "timedout" : e.failed ? "failed" : e.skipped ? "skipped" : e.passed ? "passed" : e.pending ? "pending" : "running";
}
function Wc(e) {
	let t = Uc(e);
	return t === "running" || t === "pending";
}
function Gc(e, t, n) {
	let { passed: r, failed: i, skipped: a, warned: o, aborted: s, pending: c, running: l, timed_out: u, failure_detail: d, message: f, stdout: p, stderr: m, duration: h, ...g } = e, _ = Uc(t);
	return {
		...g,
		..._ === "timedout" ? { timed_out: !0 } : {},
		..._ === "failed" ? { failed: !0 } : {},
		..._ === "skipped" ? { skipped: !0 } : {},
		..._ === "passed" ? { passed: !0 } : {},
		..._ === "pending" ? { pending: !0 } : {},
		..._ === "running" ? { running: !0 } : {},
		...t.message === void 0 ? {} : { message: t.message },
		...t.stdout === void 0 ? {} : { stdout: t.stdout },
		...t.stderr === void 0 ? {} : { stderr: t.stderr },
		...t.duration === void 0 ? {} : { duration: t.duration },
		...t.command === void 0 ? {} : { command: t.command },
		...n && d ? { failure_detail: d } : {},
		attempts: [t]
	};
}
function Kc(e) {
	let t = e.attempts ?? [];
	return t.map((n, r) => ({
		id: r + 1,
		status: Uc(n),
		...n.duration !== void 0 && n.duration > 0 ? { durationMs: n.duration / 1e6 } : {},
		node: Gc(e, n, r === t.length - 1)
	}));
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/adapter.ts
var qc = class {
	adapters;
	constructor(e = []) {
		this.adapters = e;
	}
	firstWith(e, t, n) {
		for (let r of this.adapters) if (n(r) && r.match(e, t)) return r;
		return null;
	}
	resolveDetail(e, t) {
		return this.firstWith(e, t, (e) => !!e.renderDetail);
	}
	resolveRowLeading(e, t) {
		return this.firstWith(e, t, (e) => !!e.renderRowLeading);
	}
	resolveFrameworkIcon(e, t) {
		return this.firstWith(e, t, (e) => !!e.renderFrameworkIcon);
	}
	resolveActions(e, t) {
		return this.firstWith(e, t, (e) => !!e.nodeActions);
	}
	resolveAttempts(e, t) {
		let n = this.firstWith(e, t, (e) => !!e.detailAttempts)?.detailAttempts?.({
			node: e,
			runner: t
		});
		return n && n.length > 0 ? n : Kc(e);
	}
	resolveTabs(e, t) {
		let n = {
			node: e,
			runner: t
		}, r = /* @__PURE__ */ new Set(), i = [];
		for (let a of this.adapters) if (a.detailTabs && a.match(e, t)) for (let e of a.detailTabs(n)) r.has(e.id) || (r.add(e.id), i.push(e));
		return i;
	}
	resolveCopy(e, t) {
		let n = {
			node: e,
			runner: t
		};
		return Rc(this.adapters.flatMap((r) => r.copy && r.match(e, t) ? [r.copy(n)] : []));
	}
	ownsScroll(e, t) {
		let n = this.resolveDetail(e, t);
		return n?.ownsScroll ? n.ownsScroll({
			node: e,
			runner: t
		}) : !1;
	}
};
function Jc(e = []) {
	return new qc(e);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/context.ts
var Yc = (0, U.createContext)(null), Xc = Yc.Provider;
function Zc() {
	let e = (0, U.useContext)(Yc);
	if (!e) throw Error("useTestRunner must be used within a <TestRunner> / TestRunnerProvider");
	return e;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/TestTreeNode.tsx
var Qc = {
	failed: "text-red-600 dark:text-red-400",
	timedout: "text-amber-600 dark:text-amber-400",
	warned: "text-amber-600 dark:text-amber-400",
	aborted: "text-red-500 dark:text-red-300",
	running: "text-blue-600 dark:text-blue-400",
	skipped: "text-muted-foreground",
	pending: "text-muted-foreground",
	passed: "text-foreground"
}, $c = /* @__PURE__ */ new Set([
	"failed",
	"timedout",
	"warned",
	"aborted",
	"skipped"
]);
function el({ node: e }) {
	let t = je(e);
	return t.total === 0 ? null : /* @__PURE__ */ (0, W.jsxs)("span", {
		className: "flex shrink-0 items-center gap-1",
		children: [
			t.passed > 0 && /* @__PURE__ */ (0, W.jsx)(w, {
				tone: "success",
				variant: "solid",
				size: "xxs",
				count: t.passed
			}),
			t.failed > 0 && /* @__PURE__ */ (0, W.jsx)(w, {
				tone: "danger",
				variant: "solid",
				size: "xxs",
				count: t.failed
			}),
			t.warned > 0 && /* @__PURE__ */ (0, W.jsx)(w, {
				tone: "warning",
				variant: "outline",
				size: "xxs",
				count: t.warned
			}),
			t.timedout > 0 && /* @__PURE__ */ (0, W.jsx)(w, {
				tone: "warning",
				variant: "solid",
				size: "xxs",
				count: t.timedout
			}),
			t.aborted > 0 && /* @__PURE__ */ (0, W.jsx)(w, {
				tone: "danger",
				variant: "soft",
				size: "xxs",
				count: t.aborted
			}),
			t.skipped > 0 && /* @__PURE__ */ (0, W.jsx)(w, {
				tone: "neutral",
				variant: "soft",
				size: "xxs",
				count: t.skipped
			}),
			t.running > 0 && /* @__PURE__ */ (0, W.jsx)(w, {
				tone: "info",
				variant: "solid",
				size: "xxs",
				count: t.running
			}),
			t.pending > 0 && /* @__PURE__ */ (0, W.jsx)(w, {
				tone: "neutral",
				variant: "soft",
				size: "xxs",
				count: t.pending
			})
		]
	});
}
function tl({ node: e, selected: t }) {
	let n = Zc(), r = (e.children?.length ?? 0) > 0, a = ye(e), o = Ze(e), s = n.adapters.resolveFrameworkIcon(e, n)?.renderFrameworkIcon?.({
		node: e,
		runner: n
	}), l = n.adapters.resolveRowLeading(e, n), u = e.duration || (r ? Xe(e) : 0), d = !!n.onRerun && e.framework !== "task", f = !!n.onStop && !!e.can_stop && !!e.task_id, m = a && ($c.has(a) || e.framework === "task" && (a === "running" || a === "pending")) ? (e.message ?? "").trim() : "";
	return /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
		o ? /* @__PURE__ */ (0, W.jsx)(H, { outcome: o }) : /* @__PURE__ */ (0, W.jsx)(p, {
			icon: Et,
			className: "shrink-0 text-muted-foreground"
		}),
		s == null ? /* @__PURE__ */ (0, W.jsx)(Ie, {
			tool: Re(e),
			className: "text-sm"
		}) : /* @__PURE__ */ (0, W.jsx)("span", {
			className: "flex shrink-0 items-center text-sm",
			children: s
		}),
		l?.renderRowLeading?.({
			node: e,
			runner: n
		}),
		/* @__PURE__ */ (0, W.jsx)("span", {
			className: i("min-w-0 truncate", t ? "font-semibold" : "font-medium", a && Qc[a] || "text-foreground"),
			title: m || void 0,
			children: Fe(e.name, e.framework)
		}),
		m ? /* @__PURE__ */ (0, W.jsx)("span", {
			className: "min-w-0 flex-1 shrink-[100] truncate text-xs text-muted-foreground",
			title: m,
			"data-row-message": "",
			children: m
		}) : /* @__PURE__ */ (0, W.jsx)("span", { className: "flex-1" }),
		e.progress && e.progress.total > 0 && !e.passed && !e.failed && /* @__PURE__ */ (0, W.jsxs)("span", {
			className: "flex shrink-0 items-center gap-1 text-xs text-blue-600 dark:text-blue-400",
			children: [e.progress.phase && /* @__PURE__ */ (0, W.jsx)("span", { children: e.progress.phase }), /* @__PURE__ */ (0, W.jsxs)("span", {
				className: "tabular-nums",
				children: [
					e.progress.done,
					"/",
					e.progress.total
				]
			})]
		}),
		u > 0 && /* @__PURE__ */ (0, W.jsx)("span", {
			className: "shrink-0 text-xs text-muted-foreground",
			children: Ye(u)
		}),
		r && /* @__PURE__ */ (0, W.jsx)(el, { node: e }),
		f && /* @__PURE__ */ (0, W.jsx)(Pa, {
			icon: c,
			label: "Stop",
			disabled: n.busy.stop,
			onClick: (t) => {
				t.stopPropagation(), n.onStop(e);
			}
		}),
		d && /* @__PURE__ */ (0, W.jsx)(Pa, {
			icon: Pt,
			label: "Rerun",
			disabled: n.busy.rerun,
			onClick: (t) => {
				t.stopPropagation(), n.onRerun(e);
			}
		})
	] });
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/TestTree.tsx
var nl = (() => {
	let e = /* @__PURE__ */ new WeakMap(), t = 0;
	return (n) => {
		if (n.route_path) return n.route_path;
		let r = e.get(n);
		return r || (r = `tn-${t++}-${n.name}`, e.set(n, r)), r;
	};
})();
function rl({ className: e, actions: t }) {
	let n = Zc();
	return /* @__PURE__ */ (0, W.jsx)(xe, {
		className: i("min-h-0", e),
		roots: n.tests,
		getKey: nl,
		getChildren: (e) => e.children,
		selected: n.selected,
		onSelect: (e) => n.onSelect(e === n.selected ? null : e),
		expandAll: n.expandAll,
		onExpandAllChange: n.onExpandAllChange,
		defaultOpen: (e, t) => t < 1 || Qe(e),
		showSearch: !0,
		searchLabel: "Filter tests",
		getSearchText: (e) => `${e.name} ${e.framework ?? ""}`,
		renderRow: ({ node: e, selected: t }) => /* @__PURE__ */ (0, W.jsx)(tl, {
			node: e,
			selected: t
		}),
		rowClass: (e, t) => i("hover:bg-accent", ye(e) === "running" && "border-l-2 border-primary bg-primary/10 font-semibold hover:bg-primary/15", t && "border-l-2 border-primary bg-primary/20"),
		actions: t
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/SplitButton.tsx
function il({ label: e, onClick: t, items: n, icon: r, variant: o = "default", size: s = "default", loading: c = !1, disabled: l = !1, primaryDisabled: u = !1, align: d = "right", title: f, className: m, menuClassName: h }) {
	let [g, _] = (0, U.useState)(!1);
	return n.length === 0 ? /* @__PURE__ */ (0, W.jsxs)(a, {
		variant: o,
		size: s,
		loading: c,
		disabled: l || u,
		onClick: t,
		className: m,
		children: [r && /* @__PURE__ */ (0, W.jsx)(p, { ...typeof r == "string" ? { name: r } : { icon: r } }), e]
	}) : /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i("inline-flex items-stretch", m),
		children: [/* @__PURE__ */ (0, W.jsxs)(a, {
			variant: o,
			size: s,
			loading: c,
			disabled: l || u,
			onClick: t,
			className: "rounded-r-none",
			children: [r && /* @__PURE__ */ (0, W.jsx)(p, { ...typeof r == "string" ? { name: r } : { icon: r } }), e]
		}), /* @__PURE__ */ (0, W.jsx)(I, {
			items: n,
			align: d,
			className: "items-stretch",
			...h ? { menuClassName: h } : {},
			onOpenChange: _,
			trigger: /* @__PURE__ */ (0, W.jsx)(a, {
				variant: o,
				size: s === "default" ? "icon" : s,
				disabled: l,
				title: f,
				"aria-label": f ?? "Open menu",
				className: i("h-auto self-stretch rounded-l-none border-l border-l-current/25 px-2", s !== "icon" && "w-8"),
				children: /* @__PURE__ */ (0, W.jsx)(p, { icon: g ? _e : z })
			})
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/detailTabContext.ts
var al = (0, U.createContext)(null), ol = al.Provider;
function sl() {
	return (0, U.useContext)(al);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/jvm-stacktrace.ts
var cl = /^"(?<name>[^"]*)"(?<rest>.*)$/, ll = /^\s*java\.lang\.Thread\.State:\s+(?<state>[A-Z_]+)(?:\s+\((?<sub>[^)]+)\))?/, ul = /^\s*at\s+(?<fn>[^\s(]+)\((?<src>[^)]+)\)\s*$/, dl = /^(?<file>[^:]+):(?<line>\d+)$/, fl = /^\s*-\s+(?<kind>locked|waiting on|waiting to lock|parking to wait for)\b(?<rest>.*)$/;
function pl(e) {
	let t = e.trim();
	if (!t) return [];
	let n = ml(t), r = [];
	for (let e of n) {
		let t = hl(e);
		t && r.push(t);
	}
	return r;
}
function ml(e) {
	let t = [], n = [], r = e.split("\n");
	for (let e of r) {
		let r = e.replace(/\r$/, "");
		r.startsWith("\"") && n.length > 0 ? (t.push(n.join("\n")), n = [r]) : n.push(r);
	}
	return n.length > 0 && t.push(n.join("\n")), t;
}
function hl(e) {
	let t = e.split("\n"), n = t[0]?.trim() ?? "", r = cl.exec(n);
	if (!r?.groups) return null;
	let i = r.groups.name;
	if (!i) return null;
	let a = r.groups.rest ?? "", o = /#(\d+)\b/.exec(a), s = /\bprio=(\d+)/.exec(a), c = /\bnid=(0x[0-9a-f]+)/i.exec(a), l = /\bdaemon\b/.test(a), u = gl(a) ?? "", d = Cl(u), f = [];
	for (let e = 1; e < t.length; e++) {
		let n = t[e];
		if (n == null || !n.trim()) continue;
		let r = ll.exec(n);
		if (r?.groups) {
			let e = r.groups.state;
			if (!e) continue;
			let t = r.groups.sub;
			u = t ? `${e} (${t})` : e, d = Cl(e);
			continue;
		}
		let i = ul.exec(n);
		if (i?.groups) {
			let e = i.groups.fn, t = i.groups.src;
			if (!e || !t) continue;
			let n = {
				functionName: e,
				displayName: El(e),
				kind: "frame",
				runtime: wl(e),
				nativeMethod: t === "Native Method",
				...Tl(e)
			}, r = dl.exec(t);
			r?.groups?.file && r.groups.line ? (n.file = r.groups.file, n.line = Number(r.groups.line), n.location = `${n.file}:${n.line}`) : n.location = t === "Native Method" ? "Native Method" : t, f.push(n);
			continue;
		}
		let a = fl.exec(n);
		if (a?.groups?.kind) {
			let e = a.groups.kind, t = xl(e);
			f.push({
				functionName: e,
				displayName: e,
				kind: t,
				runtime: !1,
				nativeMethod: !1,
				annotationText: (a.groups.rest ?? "").trim()
			});
		}
	}
	let p = f.filter((e) => e.kind === "frame" && !e.runtime).length, m = f.find((e) => e.kind === "frame")?.functionName, h = [n, ...f.map((e) => `${e.functionName} ${e.location ?? ""} ${e.annotationText ?? ""}`)].join("\n").toLowerCase();
	return {
		id: o ? Number(o[1]) : bl(i, a),
		name: i,
		state: d || "unknown",
		rawState: u || "",
		daemon: l,
		frames: f,
		raw: e,
		userFrameCount: p,
		searchText: h,
		...c?.[1] === void 0 ? {} : { nid: c[1] },
		...s?.[1] === void 0 ? {} : { priority: Number(s[1]) },
		...m === void 0 ? {} : { topFunction: m }
	};
}
function gl(e) {
	let t = e.toLowerCase(), n = t.indexOf("nid=0x");
	if (n === -1) return;
	let r = n + 6, i = r;
	for (; _l(e[r]);) r++;
	if (r === i || !yl(e[r])) return;
	for (; yl(e[r]);) r++;
	let a = e.length;
	for (; a > r && yl(e[a - 1]);) a--;
	if (e[a - 1] === "]") {
		let n = t.lastIndexOf("[0x", a - 1);
		if (n > r && vl(e, n + 3, a - 1) && yl(e[n - 1])) for (a = n; a > r && yl(e[a - 1]);) a--;
	}
	return e.slice(r, a) || void 0;
}
function _l(e) {
	if (e === void 0) return !1;
	let t = e.toLowerCase().charCodeAt(0);
	return t >= 48 && t <= 57 || t >= 97 && t <= 102;
}
function vl(e, t, n) {
	if (t >= n) return !1;
	for (let r = t; r < n; r++) if (!_l(e[r])) return !1;
	return !0;
}
function yl(e) {
	return e !== void 0 && e.trim() === "";
}
function bl(e, t) {
	let n = /\btid=(0x[0-9a-f]+)/i.exec(t);
	if (n?.[1]) {
		let e = n[1].slice(2), t = Number.parseInt(e.slice(-8), 16);
		if (Number.isFinite(t)) return t;
	}
	let r = 0;
	for (let t = 0; t < e.length; t++) r = r * 31 + e.charCodeAt(t) | 0;
	return Math.abs(r);
}
function xl(e) {
	switch (e) {
		case "locked": return "locked";
		case "waiting on": return "waiting_on";
		case "waiting to lock": return "waiting_to_lock";
		case "parking to wait for": return "parking";
		default: return "frame";
	}
}
function Sl(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.state, (t.get(n.state) ?? 0) + 1);
	return t;
}
function Cl(e) {
	if (!e) return "";
	let t = e.trim().toUpperCase();
	if (t.startsWith("RUNNABLE")) return "runnable";
	if (t.startsWith("TIMED_WAITING")) return "timed_waiting";
	if (t.startsWith("WAITING")) return "waiting";
	if (t.startsWith("BLOCKED")) return "blocked";
	if (t.startsWith("NEW")) return "new";
	if (t.startsWith("TERMINATED")) return "terminated";
	let n = e.trim().toLowerCase();
	return n.includes("runnable") ? "runnable" : n.includes("waiting on condition") || n.includes("sleeping") ? "timed_waiting" : n.includes("waiting") ? "waiting" : n.includes("blocked") ? "blocked" : n.split(/\s+/)[0] ?? "";
}
function wl(e) {
	return gc(e);
}
function Tl(e) {
	let t = e.lastIndexOf(".");
	return t <= 0 ? {} : {
		class: e.slice(0, t),
		method: e.slice(t + 1)
	};
}
function El(e) {
	let t = e.lastIndexOf(".");
	return t <= 0 ? e : _c(e.slice(0, t), e.slice(t + 1));
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/stacktrace.ts
function Dl(e) {
	let t = e.trim();
	return t ? /^"[^"]*"[^\n]*\b(prio=|tid=|nid=)/m.test(t) ? "jvm" : /^goroutine\s+\d+\s+\[[^\]]+\]:/m.test(t) ? "go" : "unknown" : "unknown";
}
function Ol(e) {
	let t = Dl(e);
	return t === "jvm" ? {
		format: "jvm",
		threads: pl(e)
	} : t === "go" ? {
		format: "go",
		goroutines: Pl(e)
	} : {
		format: "unknown",
		goroutines: [],
		threads: []
	};
}
function kl(e) {
	return e.format === "jvm" ? Sl(e.threads) : e.format === "go" ? Ll(e.goroutines) : /* @__PURE__ */ new Map();
}
var Al = /^goroutine\s+(\d+)\s+\[(.+?)\]:$/, jl = /^\/usr\/local\/go[\d.]+\/src\//, Ml = /^.*?\/go\/src\//, Nl = /^.*?\/go\/pkg\/mod\//;
function Pl(e) {
	let t = e.trim();
	if (!t) return [];
	let n = Ul(t), r = [];
	for (let e of n) {
		let t = e.split("\n").map((e) => e.replace(/\r$/, "")), n = t[0]?.trim(), i = Al.exec(n || "");
		if (!i) continue;
		let a = Number(i[1] ?? 0), o = i[2] ?? "", s = Rl(o), c = [];
		for (let e = 1; e < t.length; e++) {
			let n = t[e];
			if (n == null) continue;
			let r = n.trim();
			if (!r) continue;
			let i = Bl(r), a = c[c.length - 1];
			if (i && a && !a.file) {
				a.file = i.file, a.line = i.line;
				continue;
			}
			let o = r.startsWith("created by ") ? "created_by" : "frame", s = o === "created_by" ? r.slice(11) : r;
			c.push({
				functionName: s,
				displayName: Wl(s),
				kind: o,
				runtime: zl(s)
			});
		}
		for (let e of c) e.file && (e.file = Kl(e.file), e.location = `${e.file}${e.line ? `:${e.line}` : ""}`);
		let l = c.find((e) => e.kind === "frame")?.functionName;
		r.push({
			id: a,
			state: s,
			rawState: o,
			frames: c,
			raw: e,
			userFrameCount: c.filter((e) => !e.runtime && e.kind === "frame").length,
			searchText: `${n}\n${c.map((e) => `${e.functionName} ${e.file || ""}`).join("\n")}`.toLowerCase(),
			...l === void 0 ? {} : { topFunction: l }
		});
	}
	return r;
}
function Fl(e, t) {
	let n = e.file ? Kl(e.file) : void 0;
	return {
		functionName: e.func,
		displayName: Gl(e.func),
		kind: t,
		runtime: zl(e.func),
		...n ? {
			file: n,
			line: e.line,
			location: `${n}${e.line ? `:${e.line}` : ""}`
		} : {}
	};
}
function Il(e) {
	return e.map((e) => {
		if (e.format !== "go" || e.error || e.goroutineId === void 0) throw Error(`stack record ${e.checkpointId}/${e.key} is not a Go goroutine record (format "${e.format}"${e.error ? `, error "${e.error}"` : ""})`);
		let t = [...(e.frames ?? []).map((e) => Fl(e, "frame")), ...e.createdBy ? [Fl(e.createdBy, "created_by")] : []], n = e.state ?? "", r = [
			n,
			...e.waitMinutes ? [`${e.waitMinutes} minutes`] : [],
			...e.lockedToThread ? ["locked to thread"] : []
		].join(", "), i = t.find((e) => e.kind === "frame")?.functionName;
		return {
			id: e.goroutineId,
			state: Rl(n),
			rawState: r,
			frames: t,
			raw: e.raw,
			userFrameCount: t.filter((e) => !e.runtime && e.kind === "frame").length,
			searchText: `goroutine ${e.goroutineId} [${r}]\n${t.map((e) => `${e.functionName} ${e.file || ""}`).join("\n")}`.toLowerCase(),
			...i === void 0 ? {} : { topFunction: i }
		};
	});
}
function Ll(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.state, (t.get(n.state) || 0) + 1);
	return t;
}
function Rl(e) {
	return (e.split(",")[0] ?? "").trim().toLowerCase();
}
function zl(e) {
	return e.startsWith("runtime.") || e.startsWith("runtime/") || e.startsWith("internal/") || e.startsWith("runtime/internal/") || e.startsWith("syscall.") || e.startsWith("reflect.");
}
function Bl(e) {
	let t = e.lastIndexOf(":");
	if (t <= 0) return;
	let n = t + 1, r = n;
	for (; n < e.length && Vl(e[n] ?? "");) n++;
	if (n === r) return;
	let i = n;
	if (n < e.length) {
		let t = n;
		for (; n < e.length && (e[n] ?? "").trim() === "";) n++;
		if (n === t || e.slice(n, n + 3).toLowerCase() !== "+0x") return;
		n += 3;
		let r = n;
		for (; n < e.length && Hl(e[n] ?? "");) n++;
		if (n === r || n !== e.length) return;
	}
	return {
		file: e.slice(0, t),
		line: Number(e.slice(r, i))
	};
}
function Vl(e) {
	return e >= "0" && e <= "9";
}
function Hl(e) {
	let t = e.toLowerCase();
	return Vl(e) || t >= "a" && t <= "f";
}
function Ul(e) {
	let t = [], n = [];
	for (let r of e.split("\n")) r.trim() ? n.push(r) : n.length > 0 && (t.push(n.join("\n")), n = []);
	return n.length > 0 && t.push(n.join("\n")), t;
}
function Wl(e) {
	let t = e.trim(), n = t.indexOf("(");
	return n !== -1 && (t = t.slice(0, n)), Gl(t);
}
function Gl(e) {
	return e.replace(/^((?:[^./\s]+\/)+)([^./\s]+)\./, "$2.");
}
function Kl(e) {
	return e.replace(jl, "").replace(Ml, "").replace(Nl, "");
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/state-styles.ts
function ql(e) {
	return e.includes("running") ? "bg-green-50 text-green-700 dark:bg-green-500/20 dark:text-green-300" : e.includes("chan") || e.includes("wait") ? "bg-blue-50 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300" : e.includes("sleep") ? "bg-amber-50 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300" : e.includes("select") ? "bg-violet-50 text-violet-700 dark:bg-violet-500/20 dark:text-violet-300" : "bg-muted text-muted-foreground";
}
function Jl(e) {
	return e.includes("running") ? "bg-green-500" : e.includes("chan") || e.includes("wait") ? "bg-blue-500" : e.includes("sleep") ? "bg-amber-500" : e.includes("select") ? "bg-violet-500" : "bg-muted-foreground/40";
}
function Yl(e) {
	switch (e) {
		case "runnable": return "bg-green-50 text-green-700 dark:bg-green-500/20 dark:text-green-300";
		case "blocked": return "bg-red-50 text-red-700 dark:bg-red-500/20 dark:text-red-300";
		case "waiting":
		case "timed_waiting": return "bg-blue-50 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300";
		case "new": return "bg-amber-50 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300";
		case "terminated": return "bg-muted text-muted-foreground";
		default: return "bg-muted text-muted-foreground";
	}
}
function Xl(e) {
	switch (e) {
		case "runnable": return "bg-green-500";
		case "blocked": return "bg-red-500";
		case "waiting":
		case "timed_waiting": return "bg-blue-500";
		case "new": return "bg-amber-500";
		case "terminated": return "bg-muted-foreground/40";
		default: return "bg-muted-foreground/40";
	}
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/GoroutineCard.tsx
function Zl({ goroutine: e, search: t, hideRuntimeOnly: n }) {
	let r = n ? e.frames.filter((e) => !e.runtime || e.kind === "created_by") : e.frames, i = e.state === "running" || !!t;
	return /* @__PURE__ */ (0, W.jsxs)("details", {
		className: "border-0 bg-transparent",
		open: i,
		children: [/* @__PURE__ */ (0, W.jsx)("summary", {
			className: "cursor-pointer list-none px-0 py-1",
			children: /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex items-center gap-2 flex-wrap",
				children: [
					/* @__PURE__ */ (0, W.jsxs)("span", {
						className: "font-mono text-xs font-semibold text-foreground",
						children: ["g", e.id]
					}),
					/* @__PURE__ */ (0, W.jsxs)("span", {
						className: `inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] ${ql(e.state)}`,
						children: [/* @__PURE__ */ (0, W.jsx)("span", { className: `h-2 w-2 rounded-full ${Jl(e.state)}` }), e.rawState]
					}),
					/* @__PURE__ */ (0, W.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [r.length, "f"]
					}),
					e.userFrameCount > 0 && /* @__PURE__ */ (0, W.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [e.userFrameCount, "u"]
					}),
					e.topFunction && /* @__PURE__ */ (0, W.jsx)("span", {
						className: "truncate text-[11px] text-muted-foreground",
						children: e.topFunction
					})
				]
			})
		}), /* @__PURE__ */ (0, W.jsx)("div", {
			className: "pl-3 py-1 space-y-0.5",
			children: r.map((t, n) => /* @__PURE__ */ (0, W.jsx)(Ql, { frame: t }, `${e.id}-${n}`))
		})]
	});
}
function Ql({ frame: e }) {
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: e.runtime ? "text-muted-foreground" : "text-foreground",
		children: /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "flex items-start gap-1.5",
			children: [/* @__PURE__ */ (0, W.jsx)(p, {
				icon: e.kind === "created_by" ? ee : e.runtime ? _t : q,
				className: "shrink-0 mt-0.5 text-[11px]"
			}), /* @__PURE__ */ (0, W.jsx)("div", {
				className: "min-w-0",
				children: /* @__PURE__ */ (0, W.jsxs)("div", {
					className: "break-all font-mono text-[11px] font-semibold leading-4",
					children: [e.displayName, e.location && /* @__PURE__ */ (0, W.jsx)("span", {
						className: "ml-2 text-[10px] font-normal opacity-80",
						children: e.location
					})]
				})
			})]
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/JvmStackTrace.tsx
function $l({ frames: e, hideRuntimeOnly: t = !1, className: n, resolveSource: r, frameActions: i }) {
	let a = t ? e.filter((e) => e.kind !== "frame" || !e.runtime) : e;
	return a.length === 0 ? null : /* @__PURE__ */ (0, W.jsx)("div", {
		className: n,
		children: a.map((e, t) => {
			let n = r?.(e), { line: a, ...o } = n ?? {}, s = n ? {
				...e,
				...o
			} : e;
			return /* @__PURE__ */ (0, W.jsx)(xc, {
				frame: s,
				index: t,
				...i ? { frameActions: i } : {}
			}, `${e.functionName}-${t}`);
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/ThreadCard.tsx
function eu({ thread: e, search: t, hideRuntimeOnly: n, resolveSource: r, frameActions: i }) {
	let a = n ? e.frames.filter((e) => e.kind !== "frame" || !e.runtime) : e.frames, o = e.state === "runnable" || e.state === "blocked" || !!t;
	return /* @__PURE__ */ (0, W.jsxs)("details", {
		className: "border-0 bg-transparent",
		open: o,
		children: [/* @__PURE__ */ (0, W.jsx)("summary", {
			className: "cursor-pointer list-none px-0 py-1",
			children: /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex items-center gap-2 flex-wrap",
				children: [
					/* @__PURE__ */ (0, W.jsxs)("span", {
						className: "font-mono text-xs font-semibold text-foreground",
						children: ["#", e.id]
					}),
					/* @__PURE__ */ (0, W.jsx)("span", {
						className: "truncate font-mono text-xs text-foreground",
						children: e.name
					}),
					/* @__PURE__ */ (0, W.jsxs)("span", {
						className: `inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] ${Yl(e.state)}`,
						children: [/* @__PURE__ */ (0, W.jsx)("span", { className: `h-2 w-2 rounded-full ${Xl(e.state)}` }), e.rawState || e.state]
					}),
					e.daemon && /* @__PURE__ */ (0, W.jsx)("span", {
						className: "rounded-full bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground",
						children: "daemon"
					}),
					/* @__PURE__ */ (0, W.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [a.length, "f"]
					}),
					e.userFrameCount > 0 && /* @__PURE__ */ (0, W.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [e.userFrameCount, "u"]
					}),
					e.topFunction && /* @__PURE__ */ (0, W.jsx)("span", {
						className: "truncate text-[11px] text-muted-foreground",
						children: e.topFunction
					})
				]
			})
		}), /* @__PURE__ */ (0, W.jsx)($l, {
			frames: a,
			className: "space-y-0.5 py-1 pl-3",
			...r ? { resolveSource: r } : {},
			...i ? { frameActions: i } : {}
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/StackDumpView.tsx
var tu = "10rem", nu = {
	format: "unknown",
	goroutines: [],
	threads: []
};
function ru(e) {
	let t = e.filter((e) => e.error), n = e.filter((e) => !e.error), r = n.filter((e) => e.format === "go"), i = n.filter((e) => e.format === "native");
	if (r.length + i.length !== n.length) {
		let e = n.find((e) => e.format !== "go" && e.format !== "native");
		throw Error(`stack record ${e?.checkpointId} has unknown format "${e?.format}"; expected "go" or "native"`);
	}
	if (r.length > 0 && i.length > 0) throw Error("a stack dump cannot mix Go and native records; pass the records of one checkpoint and one process");
	let a = e[0];
	return {
		parsed: r.length > 0 ? {
			format: "go",
			goroutines: Il(r)
		} : nu,
		text: i.map((e) => e.raw).join("\n"),
		...t.length > 0 ? { error: t.map((e) => e.error).join("\n") } : {},
		...a ? {
			collectedAt: a.takenAt,
			reason: a.reason
		} : {}
	};
}
function iu(e, t) {
	if (t) return ru(t);
	if (e === void 0) throw Error("StackDumpView needs either text or records");
	return {
		parsed: Ol(e),
		text: e
	};
}
function au(e) {
	return e.format === "jvm" ? e.threads : e.format === "go" ? e.goroutines : [];
}
function ou(e) {
	return e.format === "jvm" ? "threads" : e.format === "go" ? "goroutines" : "frames";
}
function su(e, t) {
	return e.format === "jvm" ? Xl(t) : Jl(t);
}
function cu({ status: e }) {
	return /* @__PURE__ */ (0, W.jsx)("span", {
		className: `rounded-full px-2 py-0.5 ${e === "ready" ? "bg-green-100 text-green-700 dark:bg-green-500/20 dark:text-green-300" : e === "unsupported" ? "bg-yellow-100 text-yellow-700 dark:bg-yellow-500/20 dark:text-yellow-300" : "bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300"}`,
		children: e
	});
}
function lu({ text: e, records: t, error: n, status: r, collectedAt: i, reason: a, singleItem: o = !1, frameActions: s, resolveSource: c, headerActions: l }) {
	let [u, d] = (0, U.useState)(""), [f, m] = (0, U.useState)(/* @__PURE__ */ new Set()), [h, g] = (0, U.useState)(!0), _ = (0, U.useMemo)(() => iu(e, t), [e, t]), { parsed: v, text: y } = _, b = n ?? _.error, x = i ?? _.collectedAt, S = a ?? _.reason, C = t ? `${t[0]?.checkpointId}|${t.length}` : e, w = (0, U.useMemo)(() => kl(v), [v]), T = (0, U.useMemo)(() => au(v), [v]), E = (0, U.useMemo)(() => {
		if (o) return T;
		let e = u.trim().toLowerCase();
		return T.filter((t) => !(f.size > 0 && !f.has(t.state) || h && t.userFrameCount === 0 || e && !t.searchText.includes(e)));
	}, [
		T,
		u,
		f,
		h,
		o
	]);
	(0, U.useEffect)(() => {
		d(""), m(/* @__PURE__ */ new Set()), g(!0);
	}, [C]);
	let D = T.length, O = ou(v);
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		"data-stack-format": v.format,
		children: [
			!o && /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "space-y-density-2 px-1 py-1.5",
				children: [/* @__PURE__ */ (0, W.jsxs)("div", {
					className: "flex flex-wrap items-center justify-between gap-density-2 text-[11px] text-muted-foreground",
					children: [/* @__PURE__ */ (0, W.jsxs)("div", {
						className: "flex flex-wrap items-center gap-density-2",
						children: [
							r ? /* @__PURE__ */ (0, W.jsx)(cu, { status: r }) : null,
							x ? /* @__PURE__ */ (0, W.jsx)("span", { children: new Date(x).toLocaleString() }) : null,
							S ? /* @__PURE__ */ (0, W.jsx)("span", {
								"data-stack-reason": "",
								children: S
							}) : null,
							D > 0 && /* @__PURE__ */ (0, W.jsxs)("span", { children: [
								E.length,
								" / ",
								D,
								" ",
								O
							] })
						]
					}), l]
				}), D > 0 && /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [/* @__PURE__ */ (0, W.jsxs)("div", {
					className: "flex flex-wrap items-center gap-1.5",
					children: [
						/* @__PURE__ */ (0, W.jsxs)("div", {
							className: "relative min-w-[14rem] flex-1",
							children: [/* @__PURE__ */ (0, W.jsx)(p, {
								icon: Ue,
								className: "absolute left-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground"
							}), /* @__PURE__ */ (0, W.jsx)("input", {
								className: "w-full rounded-md border border-border bg-muted/50 py-1 pl-7 pr-2 text-xs outline-none focus:border-primary focus:bg-background",
								"aria-label": `Filter ${O}`,
								placeholder: v.format === "jvm" ? "Filter by thread name, function, or file" : "Filter by goroutine id, function, or file",
								value: u,
								onChange: (e) => d(e.target.value)
							})]
						}),
						/* @__PURE__ */ (0, W.jsxs)("label", {
							className: "inline-flex items-center gap-1.5 rounded-md border border-border bg-muted/50 px-2 py-1 text-[11px] text-muted-foreground",
							children: [/* @__PURE__ */ (0, W.jsx)("input", {
								type: "checkbox",
								checked: h,
								onChange: (e) => g(e.target.checked)
							}), "Hide runtime-only"]
						}),
						(u || f.size > 0 || !h) && /* @__PURE__ */ (0, W.jsx)("button", {
							type: "button",
							className: "text-[11px] text-muted-foreground hover:text-foreground",
							onClick: () => {
								d(""), m(/* @__PURE__ */ new Set()), g(!0);
							},
							children: "Clear"
						})
					]
				}), /* @__PURE__ */ (0, W.jsx)("div", {
					className: "flex flex-wrap items-center gap-1",
					children: Array.from(w.entries()).sort((e, t) => t[1] - e[1]).map(([e, t]) => {
						let n = f.has(e);
						return /* @__PURE__ */ (0, W.jsxs)("button", {
							type: "button",
							"aria-pressed": n,
							className: `inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] transition-colors ${n ? "border-primary bg-primary/10 text-primary" : "border-border bg-muted/50 text-muted-foreground hover:bg-background"}`,
							onClick: () => {
								m((t) => {
									let n = new Set(t);
									return n.has(e) ? n.delete(e) : n.add(e), n;
								});
							},
							children: [
								/* @__PURE__ */ (0, W.jsx)("span", { className: `h-2 w-2 rounded-full ${su(v, e)}` }),
								e,
								/* @__PURE__ */ (0, W.jsx)("span", {
									className: "text-[10px] opacity-70",
									children: t
								})
							]
						}, e);
					})
				})] })]
			}),
			o && l ? /* @__PURE__ */ (0, W.jsx)("div", {
				className: "flex justify-end px-1 py-1.5",
				children: l
			}) : null,
			b && /* @__PURE__ */ (0, W.jsx)("div", {
				className: "mt-density-2 whitespace-pre-wrap text-xs text-red-600",
				children: b
			}),
			D === 0 ? y ? /* @__PURE__ */ (0, W.jsx)("pre", {
				className: "min-h-0 flex-1 overflow-auto whitespace-pre-wrap py-1 font-mono text-[11px] leading-4 text-foreground",
				style: { minHeight: tu },
				children: y
			}) : null : /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "min-h-0 flex-1 space-y-1 overflow-auto py-1",
				style: { minHeight: tu },
				children: [
					E.length === 0 && /* @__PURE__ */ (0, W.jsxs)("div", {
						className: "py-density-3 text-center text-xs text-muted-foreground",
						children: [
							"No ",
							O,
							" match the current filters."
						]
					}),
					v.format === "jvm" && v.threads.filter((e) => E.includes(e)).map((e) => /* @__PURE__ */ (0, W.jsx)(eu, {
						thread: e,
						search: u,
						hideRuntimeOnly: h,
						...c ? { resolveSource: c } : {},
						...s ? { frameActions: s } : {}
					}, e.id)),
					v.format === "go" && v.goroutines.filter((e) => E.includes(e)).map((e) => /* @__PURE__ */ (0, W.jsx)(Zl, {
						goroutine: e,
						search: u,
						hideRuntimeOnly: h
					}, e.id))
				]
			})
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/Expectation.tsx
function uu(e) {
	return typeof e == "string" ? e : JSON.stringify(e);
}
function du(e) {
	return Array.isArray(e) ? e.some((e) => typeof e == "object" && !!e) : typeof e == "object" && !!e;
}
function fu({ value: e, tree: t = !1, className: n }) {
	return t && du(e) ? /* @__PURE__ */ (0, W.jsx)(E, { data: e }) : /* @__PURE__ */ (0, W.jsx)("code", {
		className: i("whitespace-pre-wrap break-words font-mono", n),
		children: uu(e)
	});
}
function pu({ label: e, children: t }) {
	return /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [/* @__PURE__ */ (0, W.jsx)("dt", {
		className: "text-muted-foreground",
		children: e
	}), /* @__PURE__ */ (0, W.jsx)("dd", {
		className: "min-w-0",
		children: t
	})] });
}
function mu({ check: e, checkLabel: t = "matcher", expected: n, actual: r, vars: a, location: o, className: s }) {
	let c = Object.entries(a ?? {});
	return /* @__PURE__ */ (0, W.jsxs)("dl", {
		className: i("grid grid-cols-[auto_minmax(0,1fr)] items-baseline gap-x-3 gap-y-0.5 rounded border border-border bg-muted/30 px-2 py-1.5 text-[11px]", s),
		"data-expectation": "",
		children: [
			e && /* @__PURE__ */ (0, W.jsx)(pu, {
				label: t,
				children: /* @__PURE__ */ (0, W.jsx)(fu, { value: e })
			}),
			c.map(([e, t]) => /* @__PURE__ */ (0, W.jsx)(pu, {
				label: e,
				children: /* @__PURE__ */ (0, W.jsx)(fu, {
					value: t,
					tree: !0
				})
			}, e)),
			n !== void 0 && /* @__PURE__ */ (0, W.jsx)(pu, {
				label: "expected",
				children: /* @__PURE__ */ (0, W.jsx)(fu, {
					value: n,
					className: "text-emerald-700 dark:text-emerald-400"
				})
			}),
			r !== void 0 && /* @__PURE__ */ (0, W.jsx)(pu, {
				label: "actual",
				children: /* @__PURE__ */ (0, W.jsx)(fu, {
					value: r,
					className: "text-red-700 dark:text-red-400"
				})
			}),
			o && /* @__PURE__ */ (0, W.jsx)(pu, {
				label: "at",
				children: /* @__PURE__ */ (0, W.jsx)(fu, {
					value: o,
					className: "text-muted-foreground"
				})
			})
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/TestFailureDetail.tsx
function hu({ stack: e }) {
	let t = (0, U.useMemo)(() => Ol(e), [e]);
	return t.format === "go" ? /* @__PURE__ */ (0, W.jsx)("div", {
		className: "flex max-h-[32rem] min-h-0 flex-col",
		"data-failure-stack": "go",
		children: /* @__PURE__ */ (0, W.jsx)(lu, {
			text: e,
			singleItem: t.goroutines.length === 1
		})
	}) : /* @__PURE__ */ (0, W.jsx)(dc, {
		logs: e,
		collapsedLines: 6
	});
}
function gu({ failure: e }) {
	let t = !!(e.matcher || e.expected || e.actual || e.location);
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "space-y-2",
		children: [
			e.summary && /* @__PURE__ */ (0, W.jsx)("p", {
				className: "text-sm font-medium text-red-700 dark:text-red-300",
				children: e.summary
			}),
			t && /* @__PURE__ */ (0, W.jsx)(mu, {
				check: e.matcher,
				expected: e.expected || void 0,
				actual: e.actual || void 0,
				location: e.location
			}),
			e.stack && /* @__PURE__ */ (0, W.jsx)(hu, { stack: e.stack })
		]
	});
}
var _u = "25cqh";
function vu({ node: e }) {
	return !e.stdout && !e.stderr ? null : /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "space-y-3",
		children: [e.stdout && /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "space-y-1",
			children: [/* @__PURE__ */ (0, W.jsx)("div", {
				className: "text-[10px] uppercase tracking-wide text-muted-foreground",
				children: "stdout"
			}), /* @__PURE__ */ (0, W.jsx)(dc, {
				logs: e.stdout,
				variant: "terminal",
				maxHeight: _u
			})]
		}), e.stderr && /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "space-y-1",
			children: [/* @__PURE__ */ (0, W.jsx)("div", {
				className: "text-[10px] uppercase tracking-wide text-muted-foreground",
				children: "stderr"
			}), /* @__PURE__ */ (0, W.jsx)(dc, {
				logs: e.stderr,
				variant: "terminal",
				maxHeight: _u,
				borderClass: "border-red-500/70"
			})]
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/metric-tone.ts
function yu(e) {
	return e >= 90 ? "danger" : e >= 60 ? "warning" : e > 0 ? "success" : "neutral";
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/process-metrics.ts
var bu = class extends Error {
	pid;
	constructor(e) {
		super(`process ${e} not found`), this.name = "ProcessNotFoundError", this.pid = e;
	}
}, xu = 2e3;
function Su(e, t, n) {
	let [r, i] = (0, U.useState)({}), [a, o] = (0, U.useState)(!1), s = (0, U.useRef)(n);
	s.current = n;
	let c = t && e !== void 0 && e > 0 && n !== void 0;
	return (0, U.useEffect)(() => {
		if (!c || e === void 0) return;
		let t = !1, n;
		i({}), o(!0);
		let r = (e) => {
			let t = s.current;
			if (!t) throw Error("fetchProcessMetrics was removed while an attempt was being polled");
			return t(e);
		}, a = async () => {
			try {
				let n = await r(e);
				if (t) return;
				i({ metrics: n });
			} catch (e) {
				if (t) return;
				if (e instanceof bu) {
					i((e) => ({
						...e,
						note: "process exited"
					})), o(!1);
					return;
				}
				i((t) => ({
					...t,
					note: e instanceof Error ? e.message : String(e)
				}));
			}
			n = setTimeout(a, xu);
		};
		return a(), () => {
			t = !0, clearTimeout(n), o(!1);
		};
	}, [c, e]), {
		state: r,
		polling: a
	};
}
function Cu({ attempt: e, fetchProcessMetrics: t }) {
	let n = Wc(e), [r, i] = (0, U.useState)(Date.now);
	(0, U.useEffect)(() => {
		if (!n || !e.started) return;
		let t = setInterval(() => i(Date.now()), 1e3);
		return () => clearInterval(t);
	}, [n, e.started]);
	let a = n && e.started ? Math.max(0, r - Date.parse(e.started)) * 1e6 : e.duration, { state: o, polling: s } = Su(e.pid, n, t), c = {
		...e.pid === void 0 ? {} : { pid: e.pid },
		...e.cpu_percent === void 0 ? {} : { cpu_percent: e.cpu_percent },
		...e.rss === void 0 ? {} : { rss: e.rss }
	}, l = n && o.metrics ? o.metrics : c, u = l.pid ?? e.pid, d = n ? o.note === "process exited" ? "exited" : l.status ?? "running" : "exited", f = (u ?? 0) > 0 || l.cpu_percent !== void 0 || l.rss !== void 0;
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "space-y-3",
		"data-attempt-sequence": e.sequence,
		children: [/* @__PURE__ */ (0, W.jsxs)("section", { children: [
			/* @__PURE__ */ (0, W.jsxs)("div", {
				className: "mb-1 flex items-center gap-2",
				children: [
					/* @__PURE__ */ (0, W.jsx)("div", {
						className: "text-[10px] uppercase tracking-wide text-muted-foreground",
						children: "Process metrics"
					}),
					s && /* @__PURE__ */ (0, W.jsx)("span", {
						className: "size-1.5 animate-pulse rounded-full bg-blue-500",
						role: "status",
						"aria-label": "Reading live metrics"
					}),
					e.shared_process && /* @__PURE__ */ (0, W.jsx)("span", {
						className: "rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground",
						title: "The pid, CPU and RSS belong to the go test batch this package ran in, not to the package alone",
						"data-shared-process": "",
						children: "shared by batch"
					})
				]
			}),
			f ? /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "grid grid-cols-2 gap-density-1 md:grid-cols-4",
				children: [
					u !== void 0 && u > 0 && /* @__PURE__ */ (0, W.jsx)(Oc, {
						label: "PID",
						value: String(u)
					}),
					/* @__PURE__ */ (0, W.jsx)(Oc, {
						label: "Status",
						value: d
					}),
					/* @__PURE__ */ (0, W.jsx)(Oc, {
						label: "CPU",
						value: l.cpu_percent === void 0 ? "—" : `${l.cpu_percent.toFixed(1)}%`,
						...l.cpu_percent === void 0 ? {} : {
							barPct: Math.min(100, l.cpu_percent),
							tone: yu(l.cpu_percent)
						}
					}),
					/* @__PURE__ */ (0, W.jsx)(Oc, {
						label: "RSS",
						value: l.rss !== void 0 && l.rss > 0 ? Va(l.rss) : "—"
					}),
					l.open_files !== void 0 && /* @__PURE__ */ (0, W.jsx)(Oc, {
						label: "FDs",
						value: String(l.open_files)
					}),
					a !== void 0 && a > 0 && /* @__PURE__ */ (0, W.jsx)(Oc, {
						label: "Duration",
						value: Ye(a)
					}),
					e.exit_code !== void 0 && /* @__PURE__ */ (0, W.jsx)(Oc, {
						label: "Exit",
						value: String(e.exit_code),
						tone: e.exit_code === 0 ? "success" : "danger"
					}),
					e.goroutine_count !== void 0 && /* @__PURE__ */ (0, W.jsx)(Oc, {
						label: "Goroutines",
						value: String(e.goroutine_count)
					})
				]
			}) : /* @__PURE__ */ (0, W.jsx)("div", {
				className: "text-xs italic text-muted-foreground",
				children: "No metrics captured for this attempt."
			}),
			o.note && n && /* @__PURE__ */ (0, W.jsx)("div", {
				className: "mt-1 text-[11px] italic text-muted-foreground",
				children: o.note
			})
		] }), e.stack_trace && /* @__PURE__ */ (0, W.jsxs)("section", { children: [/* @__PURE__ */ (0, W.jsx)("div", {
			className: "mb-1 text-[10px] uppercase tracking-wide text-muted-foreground",
			children: "Stack trace"
		}), /* @__PURE__ */ (0, W.jsx)("div", {
			className: "flex max-h-[32rem] min-h-0 flex-col",
			children: /* @__PURE__ */ (0, W.jsx)(lu, { text: e.stack_trace })
		})] })]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/TestDetailPanel.tsx
function wu({ node: e }) {
	let t = Zc(), n = e.attempts?.at(-1), r = (e.children?.length ?? 0) > 0 ? je(e) : null;
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "space-y-4 p-density-4",
		children: [
			e.message && !e.failure_detail && /* @__PURE__ */ (0, W.jsx)("p", {
				className: "whitespace-pre-wrap break-words text-sm",
				children: e.message
			}),
			e.failure_detail && /* @__PURE__ */ (0, W.jsx)(gu, { failure: e.failure_detail }),
			/* @__PURE__ */ (0, W.jsx)(vu, { node: e }),
			n && /* @__PURE__ */ (0, W.jsx)(Cu, {
				attempt: n,
				...t.fetchProcessMetrics ? { fetchProcessMetrics: t.fetchProcessMetrics } : {}
			}, `${n.run_kind ?? ""}:${n.sequence}`),
			r && r.total > 0 && /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex flex-wrap gap-3 text-xs text-muted-foreground",
				children: [
					/* @__PURE__ */ (0, W.jsxs)("span", { children: [r.total, " total"] }),
					r.passed > 0 && /* @__PURE__ */ (0, W.jsxs)("span", { children: [r.passed, " passed"] }),
					r.failed > 0 && /* @__PURE__ */ (0, W.jsxs)("span", { children: [r.failed, " failed"] }),
					r.skipped > 0 && /* @__PURE__ */ (0, W.jsxs)("span", { children: [r.skipped, " skipped"] })
				]
			}),
			e.detail != null && /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "space-y-1",
				children: [/* @__PURE__ */ (0, W.jsx)("div", {
					className: "text-[10px] uppercase tracking-wide text-muted-foreground",
					children: "Detail"
				}), /* @__PURE__ */ (0, W.jsx)(E, { data: e.detail })]
			})
		]
	});
}
function Tu(e, t) {
	let n = e.icon ? { icon: e.icon } : {};
	return "reason" in e ? {
		label: e.label,
		...n,
		disabled: !0,
		title: e.reason,
		onSelect: () => {}
	} : {
		label: e.label,
		...n,
		onSelect: () => t(e.value)
	};
}
var Eu = {
	idle: "Copy details",
	pending: "Copying…",
	copied: "Copied",
	error: "Copy failed"
}, Du = {
	idle: se,
	pending: It,
	copied: L,
	error: Tt
};
function Ou({ node: e }) {
	let t = Zc(), { state: n, copy: r } = ie(), i = (0, U.useMemo)(() => t.adapters.resolveCopy(e, t), [e, t]), a = [{
		label: "Copy as JSON",
		icon: Dt,
		onSelect: () => r(Hc(e))
	}, ...i.items.map((e) => Tu(e, r))];
	return /* @__PURE__ */ (0, W.jsx)(il, {
		label: Eu[n],
		icon: Du[n],
		variant: "outline",
		size: "sm",
		align: "right",
		title: "More copy options",
		onClick: () => r(Vc(e, i)),
		items: a
	});
}
function ku({ node: e, actionNode: t, actions: n, attemptChooser: r }) {
	let i = Zc(), o = Ze(e), s = !!i.onRerun && t.framework !== "task", l = !!i.onStop && !!t.can_stop && !!t.task_id, u = e.file ? `${e.file}${e.line ? `:${e.line}` : ""}` : void 0;
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "flex min-w-0 items-center gap-2 border-b border-border px-density-4 py-density-2",
		"data-detail-header": "",
		children: [
			o ? /* @__PURE__ */ (0, W.jsx)(H, {
				outcome: o,
				className: "text-lg"
			}) : /* @__PURE__ */ (0, W.jsx)(p, {
				icon: Et,
				className: "shrink-0 text-lg text-muted-foreground"
			}),
			/* @__PURE__ */ (0, W.jsx)("h2", {
				className: "min-w-0 shrink truncate text-base font-semibold",
				title: e.name,
				children: e.name
			}),
			e.framework && /* @__PURE__ */ (0, W.jsxs)("span", {
				className: "inline-flex shrink-0 items-center gap-1 rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground",
				children: [/* @__PURE__ */ (0, W.jsx)(Ie, {
					tool: Re(e),
					className: "text-sm"
				}), e.framework]
			}),
			e.duration ? /* @__PURE__ */ (0, W.jsxs)("span", {
				className: "inline-flex shrink-0 items-center gap-1 text-xs text-muted-foreground",
				children: [/* @__PURE__ */ (0, W.jsx)(p, {
					icon: J,
					className: "text-xs"
				}), Ye(e.duration)]
			}) : null,
			u && /* @__PURE__ */ (0, W.jsx)("span", {
				className: "min-w-0 shrink-[100] truncate font-mono text-xs text-muted-foreground",
				title: u,
				children: u
			}),
			e.file && i.onOpenFile && /* @__PURE__ */ (0, W.jsx)(kc, {
				className: "shrink-0",
				onOpenFile: i.onOpenFile,
				target: {
					file: Ve(e.file, e.work_dir),
					...e.work_dir ? { workDir: e.work_dir } : {},
					...e.line ? { line: e.line } : {}
				}
			}),
			/* @__PURE__ */ (0, W.jsxs)("div", {
				className: "ml-auto flex shrink-0 items-center gap-1",
				children: [
					r,
					n,
					/* @__PURE__ */ (0, W.jsx)(Ou, { node: e }),
					s && /* @__PURE__ */ (0, W.jsxs)(a, {
						size: "sm",
						disabled: i.busy.rerun,
						onClick: () => i.onRerun(t),
						title: "Rerun this test or subtree",
						children: [/* @__PURE__ */ (0, W.jsx)(p, {
							icon: Pt,
							className: "mr-1 text-sm"
						}), i.busy.rerun ? "Running..." : "Rerun"]
					}),
					l && /* @__PURE__ */ (0, W.jsxs)(a, {
						size: "sm",
						variant: "destructive",
						disabled: i.busy.stop,
						onClick: () => i.onStop(t),
						title: "Stop this running task",
						children: [/* @__PURE__ */ (0, W.jsx)(p, {
							icon: c,
							className: "mr-1 text-sm"
						}), i.busy.stop ? "Stopping..." : "Stop"]
					})
				]
			})
		]
	});
}
function Au() {
	let e = Zc(), t = e.selected, [n, r] = (0, U.useState)(null), a = e.activeTab !== void 0, [o, s] = (0, U.useState)("detail"), c = a ? e.activeTab : o, l = (t) => {
		a ? e.onTabChange?.(t) : s(t);
	};
	if (!t) return /* @__PURE__ */ (0, W.jsx)("div", {
		className: "flex h-full items-center justify-center text-sm text-muted-foreground",
		children: /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "text-center",
			children: [/* @__PURE__ */ (0, W.jsx)(p, {
				icon: Et,
				className: "mb-2 block text-4xl"
			}), "Select a test to view its details."]
		})
	});
	let u = e.adapters.resolveAttempts(t, e), d = t.task_id ?? t.name, f = e.activeAttempt ?? (n?.key === d ? n.id : void 0), m = u.find((e) => e.id === f) ?? u.at(-1), h = u.length > 1 && m ? m.node : t, g = u.length > 1 && m ? /* @__PURE__ */ (0, W.jsx)("div", {
		className: "w-44 shrink-0",
		children: /* @__PURE__ */ (0, W.jsx)(pn, {
			"aria-label": "Attempt",
			value: String(m.id),
			className: "h-7 text-xs",
			onChange: (t) => {
				let n = Number(t.target.value);
				e.onAttemptChange ? e.onAttemptChange(n) : r({
					key: d,
					id: n
				});
			},
			options: u.map((e) => ({
				value: String(e.id),
				label: `Attempt ${e.id}/${u.length} · ${e.status}${e.durationMs === void 0 ? "" : ` · ${Ye(e.durationMs * 1e6)}`}`
			}))
		})
	}) : null, _ = e.adapters.resolveDetail(h, e), v = e.adapters.resolveActions(t, e), y = e.adapters.resolveTabs(h, e), b = e.adapters.ownsScroll(h, e), x = /* @__PURE__ */ (0, W.jsx)(ol, {
		value: l,
		children: _?.renderDetail ? _.renderDetail({
			node: h,
			runner: e
		}) : /* @__PURE__ */ (0, W.jsx)(wu, { node: h })
	}), S = v?.nodeActions?.({
		node: t,
		runner: e
	}), C = i("min-h-0 flex-1 [container-type:size]", !b && "overflow-y-auto");
	if (y.length === 0) return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		children: [/* @__PURE__ */ (0, W.jsx)(ku, {
			node: h,
			actionNode: t,
			actions: S,
			attemptChooser: g
		}), /* @__PURE__ */ (0, W.jsx)("div", {
			className: C,
			children: x
		})]
	});
	let w = {
		id: "detail",
		label: "Detail",
		render: () => x
	}, T = [w, ...y], E = T.find((e) => e.id === c) ?? w;
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		children: [
			/* @__PURE__ */ (0, W.jsx)(ku, {
				node: h,
				actionNode: t,
				actions: S,
				attemptChooser: g
			}),
			/* @__PURE__ */ (0, W.jsx)("div", {
				className: "flex shrink-0 items-center gap-1 border-b border-border px-density-2 py-density-1",
				children: T.map((e) => /* @__PURE__ */ (0, W.jsx)(ce, {
					active: E.id === e.id,
					onClick: () => l(e.id),
					label: e.label
				}, e.id))
			}),
			/* @__PURE__ */ (0, W.jsx)("div", {
				className: C,
				children: E.render()
			})
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/TestRunSummary.tsx
function ju(e, t) {
	return t > 0 ? Math.round(e / t * 100) : 0;
}
function Mu({ tests: e, startTime: t, endTime: n, done: r, now: a, runMeta: o, compact: s = !1, className: c }) {
	let l = We(e), u = r && n ? n : a ?? 0, d = t && u >= t ? `${((u - t) / 1e3).toFixed(1)}s` : "", f = [
		{
			count: l.passed,
			color: "bg-green-500",
			label: "passed"
		},
		{
			count: l.skipped,
			color: "bg-slate-400",
			label: "skipped"
		},
		{
			count: l.failed,
			color: "bg-red-500",
			label: "failed"
		},
		{
			count: l.warned,
			color: "bg-amber-400",
			label: "warned"
		},
		{
			count: l.aborted,
			color: "bg-red-300",
			label: "aborted"
		},
		{
			count: l.timedout,
			color: "bg-amber-500",
			label: "timed out"
		},
		{
			count: l.running,
			color: "bg-blue-400",
			label: "running"
		},
		{
			count: l.pending,
			color: "bg-muted-foreground/40",
			label: "queued"
		}
	], m = [
		{
			n: l.passed,
			label: "passed",
			outcome: "passed"
		},
		{
			n: l.failed,
			label: "failed",
			outcome: "failed"
		},
		{
			n: l.warned,
			label: "warned",
			outcome: "warned"
		},
		{
			n: l.aborted,
			label: "aborted",
			outcome: "aborted"
		},
		{
			n: l.timedout,
			label: "timed out",
			outcome: "timedout"
		},
		{
			n: l.skipped,
			label: "skipped",
			outcome: "skipped"
		},
		{
			n: l.running,
			label: "running",
			outcome: "running"
		},
		{
			n: l.pending,
			label: "queued",
			outcome: "pending"
		}
	].filter((e) => e.n > 0).map((e) => /* @__PURE__ */ (0, W.jsxs)("span", {
		className: i("inline-flex items-center gap-0.5 tabular-nums", qe[e.outcome].className),
		"aria-label": `${e.n} ${e.label}`,
		title: `${e.n} ${e.label}`,
		"data-summary-outcome": e.outcome,
		children: [/* @__PURE__ */ (0, W.jsx)(H, { outcome: e.outcome }), e.n]
	}, e.label)), h = /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [o?.timed_out && /* @__PURE__ */ (0, W.jsxs)("span", {
		className: i("inline-flex items-center gap-0.5", qe.timedout.className),
		title: "The run hit its timeout",
		"data-summary-run": "timed-out",
		children: [/* @__PURE__ */ (0, W.jsx)(H, { outcome: "timedout" }), "Timed out"]
	}), o?.exit_code !== void 0 && /* @__PURE__ */ (0, W.jsxs)("span", {
		className: i("inline-flex items-center tabular-nums", o.exit_code === 0 ? "text-muted-foreground" : qe.failed.className),
		title: `Exit code ${o.exit_code}`,
		"data-summary-run": "exit-code",
		children: ["exit ", o.exit_code]
	})] });
	return s ? /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i("flex flex-wrap items-center gap-x-3 gap-y-0.5 text-sm text-muted-foreground", c),
		children: [
			/* @__PURE__ */ (0, W.jsxs)("span", {
				className: "font-medium text-foreground",
				children: [
					l.total,
					" ",
					l.total === 1 ? "test" : "tests",
					o?.kind === "rerun" ? ` · rerun #${o.sequence}` : ""
				]
			}),
			m,
			h,
			d && /* @__PURE__ */ (0, W.jsxs)("span", {
				className: "inline-flex items-center gap-1",
				children: [/* @__PURE__ */ (0, W.jsx)(p, {
					icon: J,
					className: "text-xs"
				}), d]
			}),
			l.total > 0 && /* @__PURE__ */ (0, W.jsxs)("span", {
				className: "inline-flex items-center gap-1.5",
				children: [/* @__PURE__ */ (0, W.jsx)(Tc, {
					className: "w-24",
					segments: f,
					total: l.total,
					height: "h-1.5"
				}), /* @__PURE__ */ (0, W.jsxs)("span", {
					className: "text-xs font-medium text-green-600 dark:text-green-400",
					children: [ju(l.passed, l.total), "%"]
				})]
			}),
			!r && /* @__PURE__ */ (0, W.jsx)(p, {
				icon: Ge,
				className: "animate-spin text-blue-500"
			})
		]
	}) : /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i("flex min-w-[18rem] flex-col items-end gap-1.5", c),
		children: [/* @__PURE__ */ (0, W.jsxs)("div", {
			className: "flex flex-wrap items-center justify-end gap-x-3 gap-y-0.5 text-sm text-muted-foreground",
			children: [
				/* @__PURE__ */ (0, W.jsxs)("span", {
					className: "font-medium text-foreground",
					children: [
						l.total,
						" ",
						l.total === 1 ? "test" : "tests",
						o?.kind === "rerun" ? ` · rerun #${o.sequence}` : ""
					]
				}),
				m,
				h,
				d && /* @__PURE__ */ (0, W.jsxs)("span", {
					className: "inline-flex items-center gap-1",
					children: [/* @__PURE__ */ (0, W.jsx)(p, {
						icon: J,
						className: "text-xs"
					}), d]
				}),
				!r && /* @__PURE__ */ (0, W.jsx)(p, {
					icon: Ge,
					className: "animate-spin text-blue-500"
				})
			]
		}), l.total > 0 && /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "flex w-72 items-center gap-2",
			children: [/* @__PURE__ */ (0, W.jsx)(Tc, {
				className: "flex-1",
				segments: f,
				total: l.total,
				height: "h-2"
			}), /* @__PURE__ */ (0, W.jsxs)("span", {
				className: "shrink-0 text-xs font-medium text-green-600 dark:text-green-400",
				children: [ju(l.passed, l.total), "%"]
			})]
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/FilterFacet.tsx
var Nu = {
	autoSubmit: !1,
	compact: !1
};
function Pu({ label: e, options: t, state: n, onChange: r }) {
	let a = Object.fromEntries(n), o = n.size > 0;
	return /* @__PURE__ */ (0, W.jsx)(I, {
		align: "right",
		menuLabel: e,
		menuClassName: "p-2",
		trigger: /* @__PURE__ */ (0, W.jsxs)("span", {
			role: "button",
			tabIndex: 0,
			className: i("inline-flex h-6 items-center gap-0.5 rounded px-1.5 text-xs transition-colors", o ? "bg-muted font-medium text-foreground" : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"),
			children: [Te(e, a), /* @__PURE__ */ (0, W.jsx)(p, {
				icon: z,
				className: "text-[10px] opacity-60"
			})]
		}),
		children: () => /* @__PURE__ */ (0, W.jsx)(B.Provider, {
			value: Nu,
			children: /* @__PURE__ */ (0, W.jsx)(Oe, {
				chrome: "embedded",
				filter: {
					key: e,
					kind: "multi",
					label: e,
					options: t,
					value: a,
					onChange: (e) => r(new Map(Object.entries(e)))
				}
			})
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/TestFilterBar.tsx
function Fu({ filters: e, onChange: t, counts: n, frameworks: r, className: a }) {
	let o = Ke.filter((e) => n[e.key] > 0).map((e) => ({
		value: e.key,
		label: e.label,
		icon: /* @__PURE__ */ (0, W.jsx)(H, { outcome: e.key }),
		count: n[e.key]
	})), s = r.map((e) => ({
		value: e,
		label: V(e)?.label ?? e,
		title: e,
		icon: /* @__PURE__ */ (0, W.jsx)(Ie, { tool: e })
	}));
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: i("flex items-center gap-0.5", a),
		children: [/* @__PURE__ */ (0, W.jsx)(Pu, {
			label: "Status",
			options: o,
			state: e.status,
			onChange: (n) => t({
				...e,
				status: n
			})
		}), r.length > 1 && /* @__PURE__ */ (0, W.jsx)(Pu, {
			label: "Tool",
			options: s,
			state: e.framework,
			onChange: (n) => t({
				...e,
				framework: n
			})
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/TestRunner.tsx
function Iu(e) {
	return e instanceof qc ? e : Jc(e ?? []);
}
var Lu = 640;
function Ru({ tests: e, allTests: t, selected: n, filters: r, expandAll: a, done: o, runMeta: s, status: c, statusText: l, now: u, busy: d, activeTab: f, onTabChange: m, activeAttempt: h, onAttemptChange: g, onSelect: _, onFiltersChange: v, onExpandAllChange: y, onRerun: b, onStop: x, onOpenFile: S, fetchProcessMetrics: C, adapters: w, title: T, showSummary: E = !0, headerSlot: D, hideHeader: O = !1, defaultSplit: k = 45, startTime: A, endTime: j, className: M }) {
	let N = (0, U.useMemo)(() => Iu(w), [w]), ee = (0, U.useMemo)(() => ({
		tests: e,
		selected: n,
		runMeta: s,
		status: c,
		done: o,
		now: u,
		filters: r,
		expandAll: a,
		busy: d ?? {},
		activeTab: f,
		onTabChange: m,
		activeAttempt: h,
		onAttemptChange: g,
		onSelect: _,
		onFiltersChange: v,
		onExpandAllChange: y,
		onRerun: b,
		onStop: x,
		onOpenFile: S,
		fetchProcessMetrics: C,
		adapters: N
	}), [
		e,
		n,
		s,
		c,
		o,
		u,
		r,
		a,
		d,
		f,
		m,
		h,
		g,
		_,
		v,
		y,
		b,
		x,
		S,
		C,
		N
	]), P = t ?? e, te = (0, U.useMemo)(() => Be(P), [P]), ne = (0, U.useMemo)(() => We(P), [P]), F = !O && P.length > 0 ? /* @__PURE__ */ (0, W.jsx)(Fu, {
		filters: r,
		onChange: v,
		counts: ne,
		frameworks: te
	}) : null, { ref: I, wider: L } = un(Lu);
	return /* @__PURE__ */ (0, W.jsx)(Xc, {
		value: ee,
		children: /* @__PURE__ */ (0, W.jsxs)("div", {
			ref: I,
			className: i("flex h-full min-h-0 flex-col bg-background", M),
			children: [!O && /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex min-w-0 items-center gap-3 border-b border-border bg-card px-density-4 py-density-2",
				"data-runner-header": "",
				children: [
					/* @__PURE__ */ (0, W.jsxs)("div", {
						className: "flex min-w-0 items-center gap-3",
						children: [T === void 0 ? /* @__PURE__ */ (0, W.jsxs)("h1", {
							className: "inline-flex shrink-0 items-center gap-1.5 text-base font-bold",
							children: [/* @__PURE__ */ (0, W.jsx)(p, {
								icon: we,
								className: "text-primary"
							}), "Test Results"]
						}) : T, l && /* @__PURE__ */ (0, W.jsx)("span", {
							className: "truncate text-sm text-muted-foreground",
							children: l
						})]
					}),
					E && /* @__PURE__ */ (0, W.jsx)(Mu, {
						tests: P,
						startTime: A,
						endTime: j,
						done: o,
						now: u,
						runMeta: s,
						compact: !0,
						className: "ml-auto flex-nowrap"
					}),
					D && /* @__PURE__ */ (0, W.jsx)("div", {
						className: i("flex shrink-0 items-center gap-1", !E && "ml-auto"),
						children: D
					})
				]
			}), L ? /* @__PURE__ */ (0, W.jsx)(Me, {
				className: "min-h-0 flex-1",
				defaultSplit: k,
				minLeft: 25,
				minRight: 30,
				leftClass: "border-r-2 border-border bg-muted/30",
				left: /* @__PURE__ */ (0, W.jsx)(rl, { actions: F }),
				right: /* @__PURE__ */ (0, W.jsx)(Au, {})
			}) : n ? /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex min-h-0 flex-1 flex-col",
				children: [/* @__PURE__ */ (0, W.jsxs)("button", {
					type: "button",
					onClick: () => _(null),
					className: "inline-flex items-center gap-1.5 border-b border-border px-density-3 py-density-2 text-left text-sm text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
					children: [/* @__PURE__ */ (0, W.jsx)(p, { icon: ue }), "Back to results"]
				}), /* @__PURE__ */ (0, W.jsx)("div", {
					className: "min-h-0 flex-1 overflow-hidden",
					children: /* @__PURE__ */ (0, W.jsx)(Au, {})
				})]
			}) : /* @__PURE__ */ (0, W.jsx)("div", {
				className: "min-h-0 flex-1 overflow-y-auto",
				children: /* @__PURE__ */ (0, W.jsx)(rl, { actions: F })
			})]
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/CopyButton.tsx
function zu({ value: e, label: t = "Copy", className: n, iconClassName: r }) {
	let { state: a, copy: o } = ie(), s = (0, U.useCallback)((t) => {
		t.stopPropagation(), o(typeof e == "function" ? e() : e);
	}, [o, e]);
	return /* @__PURE__ */ (0, W.jsx)(Pa, {
		icon: a === "copied" ? L : se,
		label: a === "copied" ? "Copied" : a === "error" ? "Copy failed" : t,
		onClick: s,
		className: n,
		iconClassName: i(a === "copied" && "text-emerald-600", a === "error" && "text-red-600", r)
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/verification/verifyReportTests.ts
var Bu = "acceptance-criteria";
function Vu(e) {
	return typeof e == "number" && Number.isFinite(e) ? e : void 0;
}
function Hu(e) {
	return {
		Total: e.total,
		Passed: e.passed,
		Failed: e.failed,
		Warned: e.warned,
		Skipped: e.skipped,
		Pending: e.pending,
		Running: e.running,
		Timedout: e.timedout,
		Duration: 0
	};
}
function Uu(e) {
	let t = { name: e.name };
	e.framework !== void 0 && (t.framework = e.framework), e.task_id !== void 0 && (t.task_id = e.task_id), e.file !== void 0 && (t.file = e.file), e.line !== void 0 && (t.line = e.line), e.message !== void 0 && (t.message = e.message), e.command !== void 0 && (t.command = e.command), e.work_dir !== void 0 && (t.work_dir = e.work_dir), e.stdout !== void 0 && (t.stdout = e.stdout), e.stderr !== void 0 && (t.stderr = e.stderr);
	let n = Vu(e.duration);
	return n !== void 0 && (t.duration = n), e.passed !== void 0 && (t.passed = e.passed), e.failed !== void 0 && (t.failed = e.failed), e.warned !== void 0 && (t.warned = e.warned), e.skipped !== void 0 && (t.skipped = e.skipped), e.pending !== void 0 && (t.pending = e.pending), e.running !== void 0 && (t.running = e.running), e.timed_out !== void 0 && (t.timed_out = e.timed_out), e.progress !== void 0 && (t.progress = e.progress), e.context !== void 0 && (t.context = e.context), e.detail !== void 0 && (t.detail = e.detail), e.summary === void 0 ? e.children !== void 0 && (t.children = e.children.map(Uu)) : t.summary = Hu(e.summary), t;
}
function Wu(e, t) {
	let n = {
		name: e.item || `Criterion ${t + 1}`,
		framework: "checklist",
		task_id: `${Bu}:${t}`,
		passed: e.passed === !0,
		failed: e.passed === !1,
		pending: e.passed !== !0 && e.passed !== !1,
		detail: e
	};
	return e.message !== void 0 && (n.message = e.message), n;
}
function Gu(e) {
	return e.length === 0 ? null : {
		name: "Acceptance criteria",
		framework: "checklist",
		task_id: Bu,
		children: e.map(Wu)
	};
}
function Ku(e) {
	let t = (e.tests ?? []).map(Uu), n = Gu(e.checklist ?? []);
	return n ? [...t, n] : t;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/verification/stepKind.ts
var qu = [
	"satisfied",
	"failed",
	"timeout"
], Ju = [
	"alert",
	"console",
	"pageerror",
	"response"
], Yu = /* @__PURE__ */ new Set([
	"step",
	"outputs",
	"exports",
	"wait",
	"attempts",
	"events",
	"background"
]), Xu = class extends Error {};
function Zu(e, t) {
	throw new Xu(`${e} must be ${t}`);
}
var Qu = (e) => typeof e == "object" && !!e && !Array.isArray(e);
function $u(e, t) {
	return Qu(e) ? e : Zu(t, "an object");
}
function ed(e, t) {
	return Array.isArray(e) ? e : Zu(t, "an array");
}
function Q(e, t) {
	return typeof e == "string" ? e : Zu(t, "a string");
}
function td(e, t) {
	return typeof e == "number" && Number.isFinite(e) ? e : Zu(t, "a number");
}
function nd(e, t) {
	return typeof e == "boolean" ? e : Zu(t, "a boolean");
}
function rd(e, t, n) {
	return t.includes(e) ? e : Zu(n, `one of ${t.join(", ")} (got ${JSON.stringify(e)})`);
}
function $(e, t, n) {
	return e[t] === void 0 || e[t] === null ? {} : { [t]: n(e[t]) };
}
function id(e, t) {
	return typeof e == "string" ? e : typeof e == "number" ? Ye(e) : Zu(t, "a duration string or nanosecond number");
}
function ad(e, t) {
	return typeof e == "string" ? e : typeof e == "number" ? String(e) : Zu(t, "a string or number");
}
function od(e) {
	let t = "detail.wait.policy", n = $u(e, t), r = (e) => {
		let n = $u(e, `${t}.backoff`);
		return {
			...$(n, "factor", (e) => td(e, `${t}.backoff.factor`)),
			...$(n, "max", (e) => id(e, `${t}.backoff.max`))
		};
	};
	return {
		...$(n, "until", (e) => Q(e, `${t}.until`)),
		...$(n, "failWhen", (e) => Q(e, `${t}.failWhen`)),
		...$(n, "interval", (e) => id(e, `${t}.interval`)),
		...$(n, "backoff", r),
		...$(n, "jitter", (e) => ad(e, `${t}.jitter`)),
		...$(n, "timeout", (e) => id(e, `${t}.timeout`))
	};
}
function sd(e) {
	let t = $u(e, "detail.wait");
	return {
		policy: od(t.policy ?? {}),
		outcome: rd(t.outcome, qu, "detail.wait.outcome"),
		durationMs: cd(t.durationMs, "detail.wait.durationMs"),
		...t.outputs === void 0 ? {} : { outputs: t.outputs },
		...$(t, "expression", (e) => Q(e, "detail.wait.expression")),
		...$(t, "cause", (e) => Q(e, "detail.wait.cause")),
		...$(t, "error", (e) => Q(e, "detail.wait.error")),
		...$(t, "trace", (e) => Q(e, "detail.wait.trace"))
	};
}
function cd(e, t) {
	let n = td(e, t);
	return n >= 0 ? n : Zu(t, "non-negative");
}
function ld(e, t) {
	let n = $u(e, t);
	Q(n.name, `${t}.name`);
	for (let e of [
		"framework",
		"message",
		"command",
		"work_dir",
		"stdout",
		"stderr"
	]) n[e] !== void 0 && Q(n[e], `${t}.${e}`);
	for (let e of [
		"passed",
		"failed",
		"timed_out",
		"warned",
		"skipped",
		"pending",
		"running"
	]) n[e] !== void 0 && nd(n[e], `${t}.${e}`);
	return n.duration !== void 0 && cd(n.duration, `${t}.duration`), n.children !== void 0 && ed(n.children, `${t}.children`).forEach((e, n) => ld(e, `${t}.children[${n}]`)), {
		...n,
		name: Q(n.name, `${t}.name`)
	};
}
function ud(e, t) {
	let n = `detail.attempts[${t}]`, r = $u(e, n);
	return {
		attempt: td(r.attempt, `${n}.attempt`),
		status: Q(r.status, `${n}.status`),
		...$(r, "error", (e) => Q(e, `${n}.error`)),
		...$(r, "durationMs", (e) => cd(e, `${n}.durationMs`)),
		...$(r, "result", (e) => ld(e, `${n}.result`))
	};
}
function dd(e) {
	let t = ed(e, "detail.attempts").map(ud);
	return t.forEach((e, t) => {
		e.attempt !== t + 1 && Zu(`detail.attempts[${t}].attempt`, `the sequence number ${t + 1}`);
	}), t;
}
function fd(e) {
	return Object.entries($u(e, "detail.exports")).map(([e, t]) => {
		let n = `detail.exports.${e}`, r = $u(t, n);
		return {
			name: e,
			value: r.value,
			step: Q(r.step, `${n}.step`)
		};
	});
}
function pd(e, t) {
	let n = `detail.events[${t}]`, r = $u(e, n);
	return {
		type: rd(r.type, Ju, `${n}.type`),
		...$(r, "text", (e) => Q(e, `${n}.text`)),
		...$(r, "url", (e) => Q(e, `${n}.url`)),
		...$(r, "status", (e) => td(e, `${n}.status`)),
		...$(r, "detail", (e) => Q(e, `${n}.detail`))
	};
}
function md(e) {
	let t = $u(e.step, "detail.step"), n = Object.fromEntries(Object.entries(e).filter(([e]) => !Yu.has(e)));
	return {
		step: {
			id: Q(t.id, "detail.step.id"),
			kind: Q(t.kind, "detail.step.kind")
		},
		...e.outputs === void 0 ? {} : { outputs: e.outputs },
		exports: e.exports == null ? [] : fd(e.exports),
		...$(e, "wait", sd),
		attempts: dd(e.attempts ?? []),
		events: ed(e.events ?? [], "detail.events").map(pd),
		background: nd(e.background ?? !1, "detail.background"),
		rest: n
	};
}
function hd(e) {
	if (!Qu(e) || e.step === void 0) return { kind: "none" };
	try {
		return {
			kind: "step",
			step: md(e)
		};
	} catch (e) {
		if (e instanceof Xu) return {
			kind: "invalid",
			error: e.message
		};
		throw e;
	}
}
function gd(e) {
	return hd(e).kind !== "none";
}
var _d = (e) => typeof e == "string", vd = (e) => typeof e == "number" && Number.isFinite(e);
function yd(e) {
	if (!Qu(e)) return {
		header: {},
		rest: e
	};
	let t = { ...e };
	delete t.key;
	let n = (e, n) => {
		let r = t[e];
		if (n(r)) return delete t[e], r;
	}, r = {}, i = (e, t) => {
		t !== void 0 && (r[e] = t);
	};
	n("name", _d), i("kind", n("kind", _d)), i("state", n("state", _d)), i("startedAt", n("started_at", _d)), i("finishedAt", n("finished_at", _d)), i("durationNs", n("duration", vd)), i("error", n("error", _d));
	let a = n("origin", Qu);
	return a && (i("file", _d(a.file) ? a.file : void 0), i("sectionPath", _d(a.section_path) ? a.section_path : void 0), i("rowIndex", vd(a.row_index) ? a.row_index : void 0), i("line", vd(a.line) ? a.line : void 0)), {
		header: r,
		rest: Object.keys(t).length > 0 ? t : void 0
	};
}
function bd(e) {
	return e.expression === void 0 ? e.outcome === "failed" && e.policy.failWhen ? "failWhen" : "until" : e.expression === e.policy.failWhen ? "failWhen" : "until";
}
function xd(e) {
	return /T(\d{2}:\d{2}:\d{2})/.exec(e)?.[1] ?? e;
}
function Sd(e) {
	return e <= 0 ? "0ms" : Ye(e * 1e6);
}
function Cd(e) {
	let t = [e.backoff?.factor !== void 0 && `×${e.backoff.factor}`, e.backoff?.max && `up to ${e.backoff.max}`].filter(Boolean);
	return [
		e.interval && `every ${e.interval}`,
		t.length > 0 && `backoff ${t.join(" ")}`,
		e.jitter && `jitter ±${e.jitter}`,
		e.timeout && `timeout ${e.timeout}`
	].filter(Boolean).join(" · ");
}
function wd(e) {
	return e === void 0 ? "" : typeof e == "string" ? e : JSON.stringify(e);
}
function Td(e) {
	let { policy: t } = e;
	return {
		title: "Wait",
		body: [
			`outcome: ${e.outcome}`,
			`elapsed: ${Sd(e.durationMs)}`,
			`policy: ${Cd(t)}`,
			t.until && `until: ${t.until}`,
			t.failWhen && `failWhen: ${t.failWhen}`,
			e.expression && `decided by: ${e.expression}`,
			e.cause && `cause: ${e.cause}`,
			e.error && `error: ${e.error}`,
			e.trace && `trace: ${e.trace}`,
			e.outputs !== void 0 && `outputs: ${JSON.stringify(e.outputs)}`
		].filter(Boolean).join("\n")
	};
}
function Ed(e) {
	return `${[
		`#${e.attempt}`,
		e.status,
		e.durationMs !== void 0 && Sd(e.durationMs)
	].filter(Boolean).join(" ")}${e.error ? ` — ${e.error}` : ""}`;
}
function Dd(e) {
	let t = [
		e.status,
		e.text,
		e.url
	].filter((e) => e !== void 0).join(" ");
	return `[${e.type}]${t ? ` ${t}` : ""}${e.detail ? ` — ${e.detail}` : ""}`;
}
function Od(e) {
	return [
		...e.outputs === void 0 ? [] : [{
			title: "Outputs",
			body: JSON.stringify(e.outputs, null, 2),
			lang: "json"
		}],
		...e.wait ? [Td(e.wait)] : [],
		...e.attempts.length > 0 ? [{
			title: "Attempts",
			body: e.attempts.map(Ed).join("\n")
		}] : [],
		...e.exports.length > 0 ? [{
			title: "Exports",
			body: e.exports.map((e) => `${e.name}=${wd(e.value)} (from ${e.step})`).join("\n")
		}] : [],
		...e.events.length > 0 ? [{
			title: "Events",
			body: e.events.map(Dd).join("\n")
		}] : []
	];
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/verification/stepKindViews.tsx
var kd = "border-red-500/60 bg-red-500/10 text-red-700 dark:text-red-400";
function Ad({ name: e, title: t, aside: n, children: r }) {
	return /* @__PURE__ */ (0, W.jsxs)("section", {
		className: "min-w-0 space-y-1.5",
		"data-step-section": e,
		children: [/* @__PURE__ */ (0, W.jsxs)("div", {
			className: "flex items-center gap-2",
			children: [/* @__PURE__ */ (0, W.jsx)("div", {
				className: "text-[10px] uppercase tracking-wide text-muted-foreground",
				children: t
			}), n]
		}), r]
	});
}
var jd = {
	satisfied: {
		label: "Satisfied",
		tone: "success"
	},
	failed: {
		label: "Failed",
		tone: "danger"
	},
	timeout: {
		label: "Timed out",
		tone: "warning"
	}
};
function Md({ condition: e, expression: t, decisive: n, wait: r }) {
	let { label: i, tone: a } = jd[r.outcome];
	return /* @__PURE__ */ (0, W.jsxs)("li", {
		className: "flex min-w-0 items-center gap-2",
		"data-wait-condition": e,
		...n ? { "data-wait-decisive": "" } : {},
		children: [
			/* @__PURE__ */ (0, W.jsx)("span", {
				className: "w-14 shrink-0 text-muted-foreground",
				children: e
			}),
			/* @__PURE__ */ (0, W.jsx)("code", {
				className: "min-w-0 shrink break-words font-mono",
				children: t
			}),
			n && /* @__PURE__ */ (0, W.jsx)(w, {
				size: "xs",
				clickToCopy: !1,
				tone: a,
				children: i
			})
		]
	});
}
function Nd({ wait: e, onViewOutputs: t }) {
	let { policy: n } = e, r = Cd(n), i = bd(e), o = e.outputs !== void 0 && !t;
	return /* @__PURE__ */ (0, W.jsxs)("section", {
		className: "min-w-0 text-[11px]",
		"data-step-section": "wait",
		children: [
			/* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex min-w-0 items-center gap-2",
				"data-wait-summary": "",
				children: [
					/* @__PURE__ */ (0, W.jsx)("span", {
						className: "shrink-0 text-muted-foreground",
						children: "Wait"
					}),
					/* @__PURE__ */ (0, W.jsx)(w, {
						size: "xs",
						clickToCopy: !1,
						children: Sd(e.durationMs)
					}),
					e.outputs !== void 0 && t && /* @__PURE__ */ (0, W.jsx)(a, {
						type: "button",
						variant: "link",
						size: "sm",
						className: "h-auto px-0 text-[11px]",
						onClick: t,
						children: "View outputs"
					})
				]
			}),
			/* @__PURE__ */ (0, W.jsxs)("ul", {
				className: "mt-1 space-y-0.5",
				"data-wait-conditions": "",
				children: [/* @__PURE__ */ (0, W.jsx)(Md, {
					condition: "until",
					expression: n.until ?? "Service readiness",
					decisive: i === "until",
					wait: e
				}), n.failWhen && /* @__PURE__ */ (0, W.jsx)(Md, {
					condition: "failWhen",
					expression: n.failWhen,
					decisive: i === "failWhen",
					wait: e
				})]
			}),
			(r || e.error || e.trace || o) && /* @__PURE__ */ (0, W.jsxs)("details", {
				className: "mt-1 text-muted-foreground",
				open: e.outcome !== "satisfied",
				children: [/* @__PURE__ */ (0, W.jsx)("summary", {
					className: "cursor-pointer",
					children: "Wait details"
				}), /* @__PURE__ */ (0, W.jsxs)("div", {
					className: "mt-1 space-y-1 break-words",
					children: [
						r && /* @__PURE__ */ (0, W.jsx)("p", { children: r }),
						e.error && /* @__PURE__ */ (0, W.jsx)("p", {
							className: "text-destructive",
							children: e.error
						}),
						e.trace && /* @__PURE__ */ (0, W.jsx)("pre", {
							className: "whitespace-pre-wrap",
							children: e.trace
						}),
						o && /* @__PURE__ */ (0, W.jsxs)("div", {
							className: "min-w-0",
							"data-wait-outputs": "",
							children: [/* @__PURE__ */ (0, W.jsx)("div", {
								className: "text-[10px] uppercase tracking-wide",
								children: "Outputs"
							}), /* @__PURE__ */ (0, W.jsx)(E, {
								data: e.outputs,
								defaultOpenDepth: 1
							})]
						})
					]
				})]
			})
		]
	});
}
var Pd = {
	passed: "success",
	failed: "danger",
	warned: "warning",
	timedout: "warning",
	running: "info"
};
function Fd(e) {
	return e.split("/").filter(Boolean).at(-1) ?? e;
}
function Id({ rowIndex: e, line: t }) {
	return e === void 0 ? t === void 0 ? void 0 : `line ${t}` : `step #${e}`;
}
function Ld({ header: e, stepKind: t }) {
	let n = [Id(e), e.sectionPath].filter(Boolean).join(" · "), r = [e.kind, t].filter((e, t, n) => !!e && n.indexOf(e) === t), i = e.startedAt && e.finishedAt ? `${xd(e.startedAt)} → ${xd(e.finishedAt)}` : e.startedAt ? `started ${xd(e.startedAt)}` : void 0;
	return !e.file && !n && r.length === 0 && !e.state && !i && e.durationNs === void 0 ? null : /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "min-w-0 space-y-1",
		"data-execution-header": "",
		children: [(e.file || n) && /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "flex min-w-0 flex-wrap items-baseline gap-x-1.5 text-xs",
			children: [
				e.file && /* @__PURE__ */ (0, W.jsx)("span", {
					className: "font-medium",
					title: e.file,
					children: Fd(e.file)
				}),
				e.file && n && /* @__PURE__ */ (0, W.jsx)("span", {
					className: "text-muted-foreground",
					children: "·"
				}),
				n && /* @__PURE__ */ (0, W.jsx)("span", {
					className: "min-w-0 break-words text-muted-foreground",
					children: n
				})
			]
		}), /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "flex min-w-0 flex-wrap items-center gap-2 text-[11px] text-muted-foreground",
			children: [
				r.map((e) => /* @__PURE__ */ (0, W.jsx)(w, {
					size: "xs",
					clickToCopy: !1,
					children: e
				}, e)),
				e.state && /* @__PURE__ */ (0, W.jsx)(w, {
					size: "xs",
					clickToCopy: !1,
					tone: Pd[e.state] ?? "neutral",
					children: e.state
				}),
				i && /* @__PURE__ */ (0, W.jsx)("span", {
					className: "font-mono",
					children: i
				}),
				e.durationNs !== void 0 && /* @__PURE__ */ (0, W.jsx)("span", { children: Ye(e.durationNs) })
			]
		})]
	});
}
function Rd({ children: e }) {
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: i("whitespace-pre-wrap break-words rounded border px-2 py-1.5 text-sm", kd),
		"data-failure-callout": "",
		children: e
	});
}
function zd({ exports: e }) {
	return /* @__PURE__ */ (0, W.jsxs)("table", {
		className: "w-full table-fixed border-collapse text-left text-[11px]",
		children: [/* @__PURE__ */ (0, W.jsx)("thead", {
			className: "text-muted-foreground",
			children: /* @__PURE__ */ (0, W.jsxs)("tr", {
				className: "border-b border-border",
				children: [
					/* @__PURE__ */ (0, W.jsx)("th", {
						className: "w-[28%] py-1 pr-2 font-medium",
						children: "Name"
					}),
					/* @__PURE__ */ (0, W.jsx)("th", {
						className: "py-1 pr-2 font-medium",
						children: "Value"
					}),
					/* @__PURE__ */ (0, W.jsx)("th", {
						className: "w-[18%] py-1 font-medium",
						children: "Step"
					})
				]
			})
		}), /* @__PURE__ */ (0, W.jsx)("tbody", {
			className: "font-mono",
			children: e.map((e) => /* @__PURE__ */ (0, W.jsxs)("tr", {
				className: "border-b border-border/60 align-top last:border-0",
				children: [
					/* @__PURE__ */ (0, W.jsx)("td", {
						className: "break-words py-1 pr-2",
						children: e.name
					}),
					/* @__PURE__ */ (0, W.jsx)("td", {
						className: "break-all py-1 pr-2",
						children: wd(e.value)
					}),
					/* @__PURE__ */ (0, W.jsx)("td", {
						className: "break-words py-1 text-muted-foreground",
						children: e.step
					})
				]
			}, e.name))
		})]
	});
}
function Bd(e) {
	return e.type === "alert" || e.type === "pageerror" || e.type === "response" && (e.status ?? 0) >= 400;
}
function Vd({ event: e }) {
	let t = Bd(e);
	return /* @__PURE__ */ (0, W.jsxs)("li", {
		className: i("space-y-0.5 rounded border px-2 py-1 text-[11px]", t ? kd : "border-border"),
		"data-event-type": e.type,
		children: [
			/* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex min-w-0 items-center gap-2",
				children: [
					/* @__PURE__ */ (0, W.jsx)(w, {
						size: "xs",
						clickToCopy: !1,
						tone: t ? "danger" : "neutral",
						children: e.type
					}),
					e.status !== void 0 && /* @__PURE__ */ (0, W.jsx)("span", {
						className: "font-mono",
						children: e.status
					}),
					e.text && /* @__PURE__ */ (0, W.jsx)("span", {
						className: i("min-w-0 break-words", e.type === "alert" && "font-medium"),
						children: e.text
					})
				]
			}),
			e.url && /* @__PURE__ */ (0, W.jsx)("div", {
				className: "break-all font-mono text-[10px] opacity-80",
				children: e.url
			}),
			e.detail && /* @__PURE__ */ (0, W.jsx)("p", {
				className: "whitespace-pre-wrap break-words",
				children: e.detail
			})
		]
	});
}
function Hd({ step: e, renderWait: t }) {
	let n = sl(), r = n && e.outputs !== void 0 ? () => n("outputs") : void 0;
	return /* @__PURE__ */ (0, W.jsxs)(W.Fragment, { children: [
		/* @__PURE__ */ (0, W.jsxs)("div", {
			className: "text-[11px] text-muted-foreground",
			"data-step-ref": "",
			children: [
				e.step.kind,
				" · ",
				e.step.id,
				e.background && " · background"
			]
		}),
		e.wait && (t ? t(e.wait) : /* @__PURE__ */ (0, W.jsx)(Nd, {
			wait: e.wait,
			...r ? { onViewOutputs: r } : {}
		})),
		e.events.length > 0 && /* @__PURE__ */ (0, W.jsx)(Ad, {
			name: "events",
			title: "Events",
			children: /* @__PURE__ */ (0, W.jsx)("ul", {
				className: "space-y-1",
				children: e.events.map((e, t) => /* @__PURE__ */ (0, W.jsx)(Vd, { event: e }, t))
			})
		}),
		e.exports.length > 0 && /* @__PURE__ */ (0, W.jsx)(Ad, {
			name: "exports",
			title: "Exports",
			children: /* @__PURE__ */ (0, W.jsx)(zd, { exports: e.exports })
		})
	] });
}
function Ud({ outputs: e }) {
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: "min-w-0 p-density-4",
		"data-step-section": "outputs",
		children: /* @__PURE__ */ (0, W.jsx)(E, { data: e })
	});
}
function Wd({ result: e }) {
	if (e.kind === "none") return null;
	if (e.kind === "invalid") return /* @__PURE__ */ (0, W.jsx)("div", {
		"data-step-source": "",
		children: /* @__PURE__ */ (0, W.jsx)(Ad, {
			name: "source",
			title: "Source",
			children: /* @__PURE__ */ (0, W.jsxs)(Gd, { children: ["Malformed source detail: ", e.error] })
		})
	});
	let { source: t } = e, n = /* @__PURE__ */ (0, W.jsxs)("span", {
		className: "font-mono text-[11px] text-muted-foreground",
		title: t.file,
		children: [
			Fd(t.file),
			":",
			t.line,
			"-",
			t.endLine
		]
	});
	return /* @__PURE__ */ (0, W.jsx)("div", {
		"data-step-source": "",
		children: /* @__PURE__ */ (0, W.jsxs)(Ad, {
			name: "source",
			title: "Source",
			aside: n,
			children: [t.error ? /* @__PURE__ */ (0, W.jsxs)("div", {
				className: i("rounded border px-2 py-1 text-xs", kd),
				children: ["Source unavailable: ", t.error]
			}) : /* @__PURE__ */ (0, W.jsx)(T, {
				language: t.language,
				source: t.text ?? "",
				copyable: !0
			}), t.moreLines ? /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "text-[11px] text-muted-foreground",
				children: [
					"… ",
					t.moreLines,
					" more lines"
				]
			}) : null]
		})
	});
}
function Gd({ children: e }) {
	return /* @__PURE__ */ (0, W.jsx)("div", {
		role: "alert",
		className: "rounded border border-red-500/60 bg-red-500/10 px-2 py-1 text-xs text-red-700 dark:text-red-400",
		children: e
	});
}
function Kd({ error: e }) {
	return /* @__PURE__ */ (0, W.jsxs)(Gd, { children: ["Malformed step detail: ", e] });
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/verification/stepSource.ts
var qd = (e) => typeof e == "object" && !!e && !Array.isArray(e), Jd = class extends Error {};
function Yd(e, t, n, r) {
	let i = e[t];
	if (!n(i)) throw new Jd(`detail.source.${t} must be ${r}`);
	return i;
}
function Xd(e, t, n, r) {
	return e[t] === void 0 ? void 0 : Yd(e, t, n, r);
}
var Zd = (e) => typeof e == "string", Qd = (e) => typeof e == "number" && Number.isFinite(e);
function $d(e) {
	if (!qd(e) || e.source === void 0) return { kind: "none" };
	let t = e.source;
	if (!qd(t)) return {
		kind: "invalid",
		error: "detail.source must be an object"
	};
	try {
		let e = {
			file: Yd(t, "file", Zd, "a string"),
			line: Yd(t, "line", Qd, "a number"),
			endLine: Yd(t, "end_line", Qd, "a number")
		}, n = {
			language: Xd(t, "language", Zd, "a string"),
			text: Xd(t, "text", Zd, "a string"),
			moreLines: Xd(t, "more_lines", Qd, "a number"),
			error: Xd(t, "error", Zd, "a string")
		};
		for (let [t, r] of Object.entries(n)) r !== void 0 && (e[t] = r);
		return {
			kind: "source",
			source: e
		};
	} catch (e) {
		if (e instanceof Jd) return {
			kind: "invalid",
			error: e.message
		};
		throw e;
	}
}
function ef(e) {
	if (!qd(e) || !("source" in e)) return e;
	let { source: t, ...n } = e;
	return Object.keys(n).length > 0 ? n : void 0;
}
function tf({ file: e, line: t, endLine: n }) {
	return `${e}:${t}-${n}`;
}
function nf(e) {
	let t = e.error ? `source unavailable: ${e.error}` : e.text ?? "", n = e.moreLines ? `\n… ${e.moreLines} more lines` : "";
	return [{
		title: `Source ${tf(e)}`,
		...e.language && !e.error ? { lang: e.language } : {},
		body: t + n
	}];
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/verification/adapterViews.tsx
function rf(e) {
	return e.failed ? "failed" : e.passed ? "passed" : "pending";
}
function af({ item: e }) {
	return /* @__PURE__ */ (0, W.jsxs)("li", {
		className: "flex items-start gap-1.5 text-xs",
		children: [/* @__PURE__ */ (0, W.jsx)(H, {
			outcome: rf(e),
			className: "mt-0.5 text-sm"
		}), /* @__PURE__ */ (0, W.jsxs)("span", {
			className: "min-w-0",
			children: [e.name, e.message ? /* @__PURE__ */ (0, W.jsx)("span", {
				className: "block text-[11px] text-muted-foreground",
				children: e.message
			}) : null]
		})]
	});
}
function of({ node: e }) {
	let t = e.children && e.children.length > 0 ? e.children : [e];
	return /* @__PURE__ */ (0, W.jsx)("div", {
		className: "space-y-2 p-density-4",
		children: /* @__PURE__ */ (0, W.jsx)("ul", {
			className: "space-y-1",
			children: t.map((e, t) => /* @__PURE__ */ (0, W.jsx)(af, { item: e }, e.task_id ?? t))
		})
	});
}
function sf(e) {
	if (e && typeof e == "object" && !Array.isArray(e)) {
		let t = e.cel_trace;
		if (typeof t == "string" && t.length > 0) return t;
	}
}
function cf(e, t) {
	if (t !== void 0 && (e || typeof t != "object" || !t || Array.isArray(t))) return t;
}
function lf({ node: e }) {
	if (e.passed) return null;
	let t = e.context, n = t?.cel_vars ?? {}, r = sf(e.detail) || t?.cel_expression, i = !!r, a = cf(i, t?.expected), o = cf(i, t?.actual), s = a !== void 0 || o !== void 0;
	if (!r && Object.keys(n).length === 0 && !s) return null;
	let c = hd(e.detail), l = c.kind === "step" ? c.step.wait : void 0, u = l && l.outcome !== "satisfied" ? bd(l) : "expression";
	return /* @__PURE__ */ (0, W.jsx)(mu, {
		check: r,
		checkLabel: u,
		vars: n,
		expected: a,
		actual: o
	});
}
function uf(...e) {
	let t = e.filter((e) => !!e);
	return t.filter((e, n) => !t.some((t, r) => r !== n && t.includes(e) && (t.length > e.length || r < n)));
}
function df({ node: e, cause: t, error: n }) {
	let r = uf(t, e.message, n);
	return r.length === 0 ? null : e.failed || e.timed_out || t ? /* @__PURE__ */ (0, W.jsx)(Rd, { children: r.map((e, t) => /* @__PURE__ */ (0, W.jsx)("div", {
		className: t > 0 ? "mt-1 text-xs opacity-90" : void 0,
		children: e
	}, t)) }) : /* @__PURE__ */ (0, W.jsx)("p", {
		className: "whitespace-pre-wrap break-words text-sm",
		children: r.join("\n")
	});
}
function ff({ node: e, renderWait: t }) {
	let n = e.context, r = n?.command ?? e.command, i = n?.cwd ?? e.work_dir, a = hd(e.detail), { header: o, rest: s } = yd(e.detail), c = ef(a.kind === "step" ? yd(a.step.rest).rest : s);
	return /* @__PURE__ */ (0, W.jsxs)("div", {
		className: "space-y-4 p-density-4",
		children: [
			/* @__PURE__ */ (0, W.jsx)(Ld, {
				header: o,
				...a.kind === "step" ? { stepKind: a.step.step.kind } : {}
			}),
			/* @__PURE__ */ (0, W.jsx)(df, {
				node: e,
				cause: a.kind === "step" ? a.step.wait?.cause : void 0,
				error: o.error
			}),
			r && /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "min-w-0 space-y-1",
				children: [
					/* @__PURE__ */ (0, W.jsx)("div", {
						className: "text-[10px] uppercase tracking-wide text-muted-foreground",
						children: "Command"
					}),
					/* @__PURE__ */ (0, W.jsx)("code", {
						className: "block overflow-x-auto whitespace-pre rounded bg-muted px-2 py-1.5 text-xs",
						children: r
					}),
					/* @__PURE__ */ (0, W.jsxs)("div", {
						className: "flex flex-wrap items-center gap-2 text-[10px] text-muted-foreground",
						children: [i && /* @__PURE__ */ (0, W.jsxs)("span", { children: ["in ", i] }), typeof n?.exit_code == "number" && /* @__PURE__ */ (0, W.jsxs)("span", {
							className: n.exit_code === 0 ? void 0 : "text-red-600",
							children: ["exit ", n.exit_code]
						})]
					})
				]
			}),
			/* @__PURE__ */ (0, W.jsx)(vu, { node: e }),
			/* @__PURE__ */ (0, W.jsx)(lf, { node: e }),
			/* @__PURE__ */ (0, W.jsx)(Wd, { result: $d(e.detail) }),
			a.kind === "step" && /* @__PURE__ */ (0, W.jsx)(Hd, {
				step: a.step,
				...t ? { renderWait: t } : {}
			}),
			a.kind === "invalid" && /* @__PURE__ */ (0, W.jsx)(Kd, { error: a.error }),
			c != null && /* @__PURE__ */ (0, W.jsxs)("details", {
				className: "text-muted-foreground",
				"data-raw-detail": "",
				children: [/* @__PURE__ */ (0, W.jsx)("summary", {
					className: "cursor-pointer text-[10px] uppercase tracking-wide",
					children: "Raw detail"
				}), /* @__PURE__ */ (0, W.jsx)("div", {
					className: "mt-1 text-foreground",
					children: /* @__PURE__ */ (0, W.jsx)(E, { data: c })
				})]
			})
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/verification/adapters.tsx
function pf(e) {
	return e.framework === "checklist";
}
function mf(e) {
	return e.failed ? "[ ]" : e.passed ? "[x]" : "[?]";
}
function hf(e) {
	return { brief: [{
		title: "Criteria",
		body: (e.children && e.children.length > 0 ? e.children : [e]).map((e) => `- ${mf(e)} ${e.name}${e.message ? ` — ${e.message}` : ""}`).join("\n")
	}] };
}
var gf = {
	id: "verification-checklist",
	match: pf,
	renderDetail: ({ node: e }) => /* @__PURE__ */ (0, W.jsx)(of, { node: e }),
	copy: ({ node: e }) => hf(e)
};
function _f(e) {
	return e.context !== void 0 || gd(e.detail) ? !0 : e.framework === "fixture" && (e.children?.length ?? 0) === 0;
}
function vf(e) {
	if (e && typeof e == "object" && !Array.isArray(e)) {
		let t = e.cel_trace;
		if (typeof t == "string" && t.length > 0) return t;
	}
}
function yf(e, t) {
	if (t !== void 0 && (e || typeof t != "object" || !t || Array.isArray(t))) return t;
}
function bf(e) {
	return typeof e == "string" ? e : JSON.stringify(e, null, 2);
}
function xf(e, t) {
	let n = t?.command ?? e.command;
	if (!n) return [];
	let r = t?.cwd ?? e.work_dir, i = t?.exit_code;
	return [{
		title: "Command",
		lang: "sh",
		body: [
			`$ ${n}`,
			r && `cwd: ${r}`,
			i !== void 0 && `exit: ${i}`
		].filter(Boolean).join("\n")
	}];
}
function Sf(e, t) {
	let n = t?.cel_expression, r = vf(e.detail), i = !!(n || r), a = t?.cel_vars ?? {}, o = yf(i, t?.expected), s = yf(i, t?.actual), c = [
		n && `expression: ${n}`,
		r && `trace:\n${r}`,
		Object.keys(a).length > 0 && `vars: ${bf(a)}`,
		o !== void 0 && `expected: ${bf(o)}`,
		s !== void 0 && `actual: ${bf(s)}`
	].filter(Boolean);
	return c.length > 0 ? [{
		title: "Assertion",
		body: c.join("\n")
	}] : [];
}
var Cf = {
	passed: "✓",
	failed: "✗",
	errored: "✗",
	timedout: "⏱",
	warned: "!",
	aborted: "■",
	skipped: "-",
	pending: "○",
	running: "…"
};
function wf(e) {
	let t = Ze(e), n = e.context?.exit_code, r = e.message ?? (n ? `exit ${n}` : void 0);
	return `${t ? Cf[t] : "·"} ${e.name}${r ? ` — ${r}` : ""}`;
}
function Tf(e) {
	let t = [
		[e.Passed, "passed"],
		[e.Failed, "failed"],
		[e.Warned, "warned"],
		[e.Timedout, "timed out"],
		[e.Skipped, "skipped"],
		[e.Pending, "pending"],
		[e.Running, "running"]
	].filter(([e]) => !!e).map(([e, t]) => `${e} ${t}`);
	return [`${e.Total} total`, ...t].join(" · ");
}
function Ef(e) {
	let t = [...e.summary ? [Tf(e.summary)] : [], ...(e.children ?? []).map(wf)];
	return t.length > 0 ? [{
		title: "Results",
		body: t.join("\n")
	}] : [];
}
function Df(e) {
	let t = e.context, n = t?.command ?? e.command, r = hd(e.detail), i = $d(e.detail);
	return {
		brief: [
			...xf(e, t),
			...Sf(e, t),
			...i.kind === "source" ? nf(i.source) : [],
			...r.kind === "step" ? Od(r.step) : [],
			...Ef(e)
		],
		items: n ? [{
			label: "Copy command",
			icon: Pe,
			value: n
		}] : []
	};
}
var Of = /* @__PURE__ */ new Set([
	"exec",
	"query",
	"fixture",
	"test",
	"lint",
	"verify",
	"services",
	"todo",
	"todo_prompt",
	"ai",
	"browser",
	"playwright",
	"agent-browser"
]);
function kf(e) {
	let t = e.framework ?? "";
	return t === "playwright" ? gd(e.detail) : Of.has(t);
}
var Af = {
	id: "verification-step-copy",
	match: kf,
	copy: ({ node: e }) => Df(e)
};
function jf(e = {}) {
	return Jc([
		gf,
		{
			id: "verification-fixture",
			match: _f,
			renderDetail: ({ node: t }) => /* @__PURE__ */ (0, W.jsx)(ff, {
				node: t,
				...e
			}),
			detailAttempts: ({ node: e }) => {
				let t = hd(e.detail);
				return t.kind === "step" ? t.step.attempts.map((n) => ({
					id: n.attempt,
					status: n.status,
					...n.durationMs === void 0 ? {} : { durationMs: n.durationMs },
					node: n.result ? Uu(n.result) : {
						name: e.name,
						message: `Attempt ${n.attempt}: execution evidence was not stored in this report.`,
						detail: { step: t.step.step }
					}
				})) : [];
			},
			detailTabs: ({ node: e }) => {
				let t = hd(e.detail);
				return t.kind !== "step" || t.step.outputs === void 0 ? [] : [{
					id: "outputs",
					label: "Outputs",
					render: () => /* @__PURE__ */ (0, W.jsx)(Ud, { outputs: t.step.outputs })
				}];
			},
			copy: ({ node: e }) => Df(e)
		},
		Af
	]);
}
export { K as $, Wa as A, Qt as B, cc as C, Qa as D, so as E, Dn as F, Pt as G, It as H, pn as I, At as J, Mt as K, dn as L, Pa as M, jr as N, Va as O, Mr as P, q as Q, un as R, fc as S, yo as T, Ft as U, Gt as V, J as W, Ot as X, kt as Y, Dt as Z, vc as _, Cu as a, yt as at, mc as b, gu as c, gt as ct, Zc as d, lt as dt, Tt as et, Jc as f, yc as g, xc as h, Ru as i, bt as it, Ia as j, Ga as k, vu as l, ht as lt, Oc as m, Ku as n, wt as nt, bu as o, vt as ot, kc as p, jt as q, zu as r, St as rt, yu as s, _t as st, jf as t, G as tt, lu as u, ft as ut, gc as v, To as w, pc as x, _c as y, sn as z };
