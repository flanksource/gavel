import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./jsx-runtime-lVm_Myqd.js";
import { t as r } from "./react-dom-BBUmKE87.js";
import { T as i, _ as a, l as o, o as s, u as c, v as l, x as u } from "./Callout-DmCAGTzW.js";
import { $ as d, A as f, B as p, C as m, D as h, E as g, H as _, I as v, M as y, N as b, O as x, S, T as C, Y as w, _ as T, _t as E, at as D, b as O, bt as k, c as A, ct as j, d as M, dt as N, et as P, f as ee, ft as F, g as I, gt as L, h as te, ht as ne, it as R, j as z, k as re, lt as B, mt as V, n as ie, nt as ae, ot as oe, p as se, pt as ce, rt as le, st as ue, tt as de, u as fe, ut as H, v as pe, vt as U, w as me, x as W, y as he, yt as ge } from "./clipboard-C0PFseuO.js";
import { C as _e, D as ve, E as ye, H as be, I as xe, J as Se, M as Ce, S as we, T as Te, X as Ee, Y as De, b as Oe, c as ke, d as Ae, f as je, g as Me, i as Ne, m as Pe, p as Fe, s as G, u as Ie, w as Le, x as Re, y as ze } from "./paths-RfM6xXIw.js";
import { a as Be, l as Ve, n as He, o as Ue, s as We } from "./floating-ui.react-dom-BkwMuSvQ.js";
// ../../../clicky-ui/packages/ui/src/data/AnsiHtml.tsx
var K = /* @__PURE__ */ e(t(), 1), Ge = /* @__PURE__ */ e(r(), 1), q = n(), Ke = {
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
function qe(e) {
	let t = [], n = /\x1b\[([0-9;]*)m/g, r = 0, i = [], a;
	for (; (a = n.exec(e)) !== null;) {
		a.index > r && t.push({
			text: e.slice(r, a.index),
			style: i.join(";")
		});
		let n = (a[1] ?? "").split(";").filter(Boolean);
		for (let e of n) if (e === "0" || e === "") i = [];
		else {
			let t = Ke[e];
			t && i.push(t);
		}
		r = a.index + a[0].length;
	}
	return r < e.length && t.push({
		text: e.slice(r),
		style: i.join(";")
	}), t;
}
function Je(e) {
	let t = {};
	for (let n of e.split(";")) {
		let [e, r] = n.split(":");
		if (!e || !r) continue;
		let i = e.replace(/-([a-z])/g, (e, t) => t.toUpperCase());
		t[i] = r;
	}
	return t;
}
function Ye({ text: e, className: t, as: n = "pre" }) {
	let r = qe(e).map((e, t) => e.style ? /* @__PURE__ */ (0, q.jsx)("span", {
		style: Je(e.style),
		children: e.text
	}, t) : /* @__PURE__ */ (0, q.jsx)("span", { children: e.text }, t));
	return n === "span" ? /* @__PURE__ */ (0, q.jsx)("span", {
		className: t,
		children: r
	}) : /* @__PURE__ */ (0, q.jsx)("pre", {
		className: t,
		children: r
	});
}
// ../../../clicky-ui/packages/ui/src/icons/components/UiEllipsis.tsx
var Xe = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
		fill: "currentColor",
		d: "M138 128a10 10 0 1 1-10-10a10 10 0 0 1 10 10m-78-10a10 10 0 1 0 10 10a10 10 0 0 0-10-10m136 0a10 10 0 1 0 10 10a10 10 0 0 0-10-10"
	})
}), {
	__source: "ph:dots-three-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "ellipsis",
	displayName: "UiEllipsis"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
		fill: "currentColor",
		d: "M224 80H32a16 16 0 0 0-16 16v64a16 16 0 0 0 16 16h192a16 16 0 0 0 16-16V96a16 16 0 0 0-16-16M60 140a12 12 0 1 1 12-12a12 12 0 0 1-12 12m68 0a12 12 0 1 1 12-12a12 12 0 0 1-12 12m68 0a12 12 0 1 1 12-12a12 12 0 0 1-12 12"
	})
}), {
	__source: "ph:dots-three-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "ellipsis",
	displayName: "UiEllipsisFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiFilter.tsx
var Ze = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
		fill: "currentColor",
		d: "M228.77 50.34A13.8 13.8 0 0 0 216 42H40a14 14 0 0 0-10.33 23.42l.06.07L98 138.38V216a14 14 0 0 0 21.77 11.64l32-21.33a14 14 0 0 0 6.23-11.65v-56.28l68.33-73a13.82 13.82 0 0 0 2.44-15.04m-11.26 6.94l-69.89 74.62A6 6 0 0 0 146 136v58.66a2 2 0 0 1-.89 1.67l-32 21.33A2 2 0 0 1 110 216v-80a6 6 0 0 0-1.62-4.1L38.53 57.32A2 2 0 0 1 40 54h176a1.9 1.9 0 0 1 1.83 1.19a1.86 1.86 0 0 1-.32 2.09"
	})
}), {
	__source: "ph:funnel-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "filter",
	displayName: "UiFilter"
}), Qe = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
		fill: "currentColor",
		d: "m227.81 66.76l-.08.09L160 139.17v55.49a16 16 0 0 1-7.13 13.34l-32 21.34A16 16 0 0 1 96 216v-76.83L28.27 66.85l-.08-.09A16 16 0 0 1 40 40h176a16 16 0 0 1 11.84 26.76Z"
	})
}), {
	__source: "ph:funnel-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "filter",
	displayName: "UiFilterFilled"
}), $e = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
		fill: "currentColor",
		d: "M174 128a6 6 0 0 1-6 6H88a6 6 0 0 1 0-12h80a6 6 0 0 1 6 6m56 0A102 102 0 1 1 128 26a102.12 102.12 0 0 1 102 102m-12 0a90 90 0 1 0-90 90a90.1 90.1 0 0 0 90-90"
	})
}), {
	__source: "ph:minus-circle-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "remove",
	displayName: "UiRemove"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
		fill: "currentColor",
		d: "M128 24a104 104 0 1 0 104 104A104.11 104.11 0 0 0 128 24m40 112H88a8 8 0 0 1 0-16h80a8 8 0 0 1 0 16"
	})
}), {
	__source: "ph:minus-circle-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "remove",
	displayName: "UiRemoveFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiEyeClosed.tsx
var et = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiZoomIn.tsx
var tt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiZoomOut.tsx
var nt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiDesktop.tsx
var rt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiStackFrameDot.tsx
var it = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
		fill: "currentColor",
		d: "M10 8a2 2 0 1 1-4 0a2 2 0 0 1 4 0"
	})
}), {
	__source: "codicon:debug-stackframe-dot",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "stack-frame-dot",
	displayName: "UiStackFrameDot"
}), at = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
		fill: "currentColor",
		d: "M10 13c0 1.103-.897 2-2 2s-2-.897-2-2s.897-2 2-2s2 .897 2 2m3.25-11a.75.75 0 0 0-.75.75v1.277A5.98 5.98 0 0 0 8 2C5.033 2 2.48 4.211 2.06 7.144a.75.75 0 0 0 1.487.213C3.86 5.159 5.776 3.501 8 3.501c1.529 0 2.919.763 3.743 2H9.75a.75.75 0 0 0 0 1.5h3.5a.75.75 0 0 0 .75-.75v-3.5a.75.75 0 0 0-.75-.75z"
	})
}), {
	__source: "codicon:debug-step-over",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "debug-step-over",
	displayName: "UiDebugStepOver"
}), ot = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiResizeVertical.tsx
var st = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiListDashes.tsx
var ct = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiSync.tsx
var lt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
		fill: "currentColor",
		d: "M14 3.5v3c0 .28-.22.5-.5.5h-3c-.28 0-.5-.22-.5-.5s.22-.5.5-.5h2.08c-.8-1.83-2.57-3-4.58-3c-2.22 0-4.2 1.5-4.81 3.64c-.06.22-.26.36-.48.36c-.05 0-.09 0-.14-.02a.493.493 0 0 1-.34-.62C2.96 3.79 5.33 2 8 2c2.05 0 3.91 1.02 5 2.69V3.5c0-.28.22-.5.5-.5s.5.22.5.5m-.58 5.52a.51.51 0 0 0-.62.35a5.02 5.02 0 0 1-4.81 3.64c-2.01 0-3.78-1.17-4.58-3h2.08c.28 0 .5-.22.5-.5s-.22-.5-.5-.5h-3c-.28 0-.5.22-.5.5v3c0 .28.22.5.5.5s.5-.22.5-.5v-1.19a5.97 5.97 0 0 0 5 2.69c2.67 0 5.04-1.79 5.77-4.36a.5.5 0 0 0-.35-.62z"
	})
}), {
	__source: "codicon:sync",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "sync",
	displayName: "UiSync"
}), ut = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsxs)("g", {
		fill: "currentColor",
		children: [
			/* @__PURE__ */ (0, q.jsx)("path", { d: "m220 169.09l-92 53.65l-92-53.65a8 8 0 0 0-8 13.82l96 56a8 8 0 0 0 8.06 0l96-56a8 8 0 1 0-8.06-13.82" }),
			/* @__PURE__ */ (0, q.jsx)("path", { d: "m220 121.09l-92 53.65l-92-53.65a8 8 0 0 0-8 13.82l96 56a8 8 0 0 0 8.06 0l96-56a8 8 0 1 0-8.06-13.82" }),
			/* @__PURE__ */ (0, q.jsx)("path", { d: "m28 86.91l96 56a8 8 0 0 0 8.06 0l96-56a8 8 0 0 0 0-13.82l-96-56a8 8 0 0 0-8.06 0l-96 56a8 8 0 0 0 0 13.82" })
		]
	})
}), {
	__source: "ph:stack-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "stack",
	displayName: "UiStackFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiArrowRight.tsx
var dt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiExternalLink.tsx
var ft = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
		fill: "currentColor",
		d: "M222 104a6 6 0 0 1-12 0V54.49l-69.75 69.75a6 6 0 0 1-8.48-8.48L201.51 46H152a6 6 0 0 1 0-12h64a6 6 0 0 1 6 6Zm-38 26a6 6 0 0 0-6 6v72a2 2 0 0 1-2 2H48a2 2 0 0 1-2-2V80a2 2 0 0 1 2-2h72a6 6 0 0 0 0-12H48a14 14 0 0 0-14 14v128a14 14 0 0 0 14 14h128a14 14 0 0 0 14-14v-72a6 6 0 0 0-6-6"
	})
}), {
	__source: "ph:arrow-square-out-light",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "external-link",
	displayName: "UiExternalLink"
}), pt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiWarningCircle.tsx
var mt = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#d97706",
		...r
	};
	return /* @__PURE__ */ (0, q.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, q.jsx)("path", {
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
	return /* @__PURE__ */ (0, q.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiListFlat.tsx
var ht = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiClass.tsx
var gt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, q.jsx)("circle", {
			cx: "8",
			cy: "8",
			r: "6.5",
			fill: "#E7EFFD",
			stroke: "#3574F0"
		}), /* @__PURE__ */ (0, q.jsx)("path", {
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
}), _t = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, q.jsx)("circle", {
			cx: "8",
			cy: "8",
			r: "6.5",
			fill: "#FFF7F7",
			stroke: "#DB3B4B"
		}), /* @__PURE__ */ (0, q.jsx)("path", {
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
}), vt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 32 32",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
		d: "M4.014 14.976a2.5 2.5 0 0 0 1.567-.518a2.38 2.38 0 0 0 .805-1.358a15.3 15.3 0 0 0 .214-2.944q.012-2.085.075-2.747a5.2 5.2 0 0 1 .418-1.686a3 3 0 0 1 .755-1.018A3.05 3.05 0 0 1 9 4.125A6.8 6.8 0 0 1 10.544 4h.7v1.96h-.387a2.34 2.34 0 0 0-1.723.468a3.4 3.4 0 0 0-.425 2.092a36 36 0 0 1-.137 4.133a4.7 4.7 0 0 1-.768 2.06A4.6 4.6 0 0 1 6.1 16a3.8 3.8 0 0 1 1.992 1.754a8.9 8.9 0 0 1 .618 3.865q0 2.435.05 2.9a1.76 1.76 0 0 0 .504 1.181a2.64 2.64 0 0 0 1.592.337h.387V28h-.7a5.7 5.7 0 0 1-1.773-.2a2.97 2.97 0 0 1-1.324-.93a3.35 3.35 0 0 1-.681-1.63a24 24 0 0 1-.165-3.234a16.5 16.5 0 0 0-.214-3.106a2.4 2.4 0 0 0-.805-1.361a2.5 2.5 0 0 0-1.567-.524Zm23.972 2.035a2.5 2.5 0 0 0-1.567.524a2.4 2.4 0 0 0-.805 1.361a16.5 16.5 0 0 0-.212 3.109a24 24 0 0 1-.169 3.234a3.35 3.35 0 0 1-.681 1.63a2.97 2.97 0 0 1-1.324.93a5.7 5.7 0 0 1-1.773.2h-.7V26.04h.387a2.64 2.64 0 0 0 1.592-.337a1.76 1.76 0 0 0 .506-1.186q.05-.462.05-2.9a8.9 8.9 0 0 1 .618-3.865A3.8 3.8 0 0 1 25.9 16a4.6 4.6 0 0 1-1.7-1.286a4.7 4.7 0 0 1-.768-2.06a36 36 0 0 1-.137-4.133a3.4 3.4 0 0 0-.425-2.092a2.34 2.34 0 0 0-1.723-.468h-.387V4h.7a6.8 6.8 0 0 1 1.54.125a3.05 3.05 0 0 1 1.149.581a3 3 0 0 1 .755 1.018a5.2 5.2 0 0 1 .418 1.686q.062.662.075 2.747a15.3 15.3 0 0 0 .212 2.947a2.38 2.38 0 0 0 .805 1.355a2.5 2.5 0 0 0 1.567.518Z",
		fill: "#E5A505"
	})
}), {
	__source: "incumbent",
	__viewBox: "0 0 32 32",
	__group: "files-code",
	__consumerName: "json",
	displayName: "UiJson"
}), yt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 32 32",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, q.jsx)("path", {
		d: "M2.5 7.955h27v16.091h-27z",
		fill: "none",
		stroke: "#755838"
	}), /* @__PURE__ */ (0, q.jsx)("path", {
		d: "M5.909 20.636v-9.272h2.727l2.728 3.409l2.727-3.409h2.727v9.272h-2.727v-5.318l-2.727 3.409l-2.728-3.409v5.318zm17.046 0l-4.091-4.5h2.727v-4.772h2.727v4.772h2.727z",
		fill: "#755838"
	})]
}), {
	__source: "incumbent",
	__viewBox: "0 0 32 32",
	__group: "files-code",
	__consumerName: "markdown",
	displayName: "UiMarkdown"
}), bt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiFileSpreadsheet.tsx
var xt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 24 24",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsxs)("g", {
		fill: "none",
		stroke: "currentColor",
		strokeLinecap: "round",
		strokeLinejoin: "round",
		strokeWidth: "2",
		children: [/* @__PURE__ */ (0, q.jsx)("path", { d: "M6 22a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h8a2.4 2.4 0 0 1 1.704.706l3.588 3.588A2.4 2.4 0 0 1 20 8v12a2 2 0 0 1-2 2z" }), /* @__PURE__ */ (0, q.jsx)("path", { d: "M14 2v5a1 1 0 0 0 1 1h5M8 13h2m4 0h2m-8 4h2m4 0h2" })]
	})
}), {
	__source: "lucide:file-spreadsheet",
	__viewBox: "0 0 24 24",
	__group: "files-code",
	__consumerName: "file-spreadsheet",
	displayName: "UiFileSpreadsheet"
}), J = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiCloudDownload.tsx
var St = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiChip.tsx
var Ct = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiRefresh.tsx
var wt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiClock.tsx
var Tt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiWatch.tsx
var Et = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiHourglass.tsx
var Dt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, q.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, q.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/data/data-table-utils.ts
function Ot(e, t) {
	return t.split(".").reduce((e, t) => {
		if (e && typeof e == "object") return e[t];
	}, e);
}
function kt(e) {
	if (Array.isArray(e)) return e.flatMap((e) => kt(e)).filter(Boolean);
	if (e == null) return [];
	if (typeof e == "object") return [JSON.stringify(e)];
	let t = String(e).trim();
	return t ? [t] : [];
}
function At(e, t) {
	return t.accessor ? t.accessor(e) : Ot(e, t.key);
}
function jt(e, t) {
	let n = At(e, t);
	return t.filterValue ? t.filterValue(n, e) : n;
}
function Mt(e, t) {
	return kt(jt(e, t));
}
function Nt(e) {
	return e.replace(/([a-z])([A-Z])/g, "$1 $2").replace(/[_-]/g, " ").replace(/\b\w/g, (e) => e.toUpperCase()).trim();
}
function Pt(e) {
	return typeof e.label == "string" ? e.label : Nt(e.key.split(".").at(-1) ?? e.key);
}
function Ft(e) {
	return e.hideable !== !1;
}
function It(e, t) {
	if (e.length === 0) return [];
	let n = t?.literalKeys ?? !1, r = /* @__PURE__ */ new Set();
	for (let t of e.slice(0, 20)) Object.keys(t).filter((e) => !e.startsWith("_")).forEach((e) => r.add(e));
	return Array.from(r).map((e, t) => ({
		key: e || `__col_${t}`,
		label: e ? Nt(e) : "(no column name)",
		sortable: !0,
		filterable: !0,
		...n ? { accessor: (t) => t[e] } : {}
	}));
}
// ../../../clicky-ui/packages/ui/src/components/filter-bar-utils.ts
function Lt(e) {
	let { key: t, kind: n, label: r, icon: i, description: a, ...o } = e;
	return o;
}
function Rt(e) {
	e.kind === "text" || e.kind === "lookup" || e.kind === "workload" || e.kind === "enum" ? e.onChange("") : e.kind === "lookup-multi" || e.kind === "select-multi" ? e.onChange([]) : e.kind === "number" || e.kind === "duration" ? e.onChange({}) : e.kind === "boolean" ? e.onChange(!1) : e.kind === "tristate" ? e.onChange(void 0) : e.kind === "date-range" ? e.onApply("", "") : e.onChange({});
}
function zt(e) {
	return e.kind === "date-range" ? String(e.from ?? "").trim() !== "" || String(e.to ?? "").trim() !== "" : e.kind === "text" || e.kind === "lookup" || e.kind === "workload" || e.kind === "enum" ? String(e.value ?? "").trim() !== "" : e.kind === "lookup-multi" || e.kind === "select-multi" ? e.value.length > 0 : e.kind === "number" || e.kind === "duration" ? String(e.value.min ?? "").trim() !== "" || String(e.value.max ?? "").trim() !== "" : e.kind === "boolean" ? e.value : e.kind === "tristate" ? e.value !== void 0 : Object.keys(e.value).length > 0;
}
// ../../../clicky-ui/packages/ui/src/data/data-table-server-filters.ts
function Bt(e) {
	if (e.autoFilter && e.manualFilter) throw Error("DataTable: autoFilter generates filters that narrow rows locally, which manualFilter then suppresses — the controls would render and do nothing. Pass externalFilters instead.");
	if (e.manualFilter && e.showGlobalFilter && !e.globalFilterControlled && !e.hasExternalSearch) throw Error("DataTable: manualFilter suppresses the built-in search's row narrowing, so showGlobalFilter needs globalFilter/onGlobalFilterChange or externalSearch to reach the source.");
}
// ../../../clicky-ui/packages/ui/src/hooks/use-sort.ts
function Vt(e, t) {
	return t.split(".").reduce((e, t) => {
		if (e && typeof e == "object") return e[t];
	}, e);
}
function Ht(e, t = {}) {
	let { defaultKey: n, defaultDir: r = "asc", sort: i, onSortChange: a, manual: o = !1, resolvers: s } = t, [c, l] = (0, K.useState)(n ? {
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
		sorted: (0, K.useMemo)(() => {
			if (!e) return [];
			if (o || !d) return e;
			let { key: t, dir: n } = d, r = s?.[t], i = (e) => r ? r(e) : Vt(e, t);
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
// ../../../clicky-ui/packages/ui/src/hooks/density-provider.tsx
function Ut({ density: e, children: t }) {
	let n = (0, K.useMemo)(() => ({
		density: e,
		setDensity: () => {}
	}), [e]);
	return /* @__PURE__ */ (0, q.jsx)(ge.Provider, {
		value: n,
		children: t
	});
}
// ../../../clicky-ui/packages/ui/src/data/FilterPill.tsx
function Y(e) {
	switch (e) {
		case "active": return "bg-primary/10 border-primary/40 text-primary font-medium";
		case "include": return "bg-green-500/10 border-green-500/50 text-green-700 dark:text-green-400 font-medium";
		case "exclude": return "bg-red-500/10 border-red-500/50 text-red-700 dark:text-red-400 font-medium";
		default: return "border-border text-muted-foreground hover:bg-accent";
	}
}
var Wt = {
	exclude: "translate-x-0",
	neutral: "translate-x-4",
	include: "translate-x-8"
};
function Gt({ mode: e, onChange: t, ariaLabel: n, interactive: r = !0 }) {
	let a = e === "exclude" ? "exclude" : e === "include" ? "include" : "neutral", o = a === "exclude" ? "bg-red-500/80" : a === "include" ? "bg-green-500/80" : "bg-muted", s = (e, n) => /* @__PURE__ */ (0, q.jsx)("span", {
		"aria-hidden": !0,
		"data-tristate-region": n,
		onClick: (n) => {
			n.preventDefault(), n.stopPropagation(), t(e);
		},
		className: "relative z-10 h-full flex-1 cursor-pointer"
	}), c = /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
		s(e === "exclude" ? "neutral" : "exclude", "exclude"),
		s(Jt(e), "neutral"),
		s(e === "include" ? "neutral" : "include", "include"),
		/* @__PURE__ */ (0, q.jsx)("span", {
			"aria-hidden": !0,
			className: i("pointer-events-none absolute left-0.5 top-0.5 z-20 flex size-4 items-center justify-center rounded-full bg-white shadow-sm transition-transform duration-150", Wt[a]),
			children: a !== "neutral" && /* @__PURE__ */ (0, q.jsx)(E, {
				icon: a === "exclude" ? B : de,
				className: i("text-[10px]", a === "exclude" && "text-red-600", a === "include" && "text-green-600")
			})
		})
	] }), l = i("relative inline-flex h-5 w-[52px] shrink-0 overflow-hidden rounded-full transition-colors duration-200", o);
	return r ? /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		"aria-label": n,
		title: n,
		onClick: (n) => {
			n.preventDefault(), n.stopPropagation(), t(Jt(e));
		},
		className: l,
		children: c
	}) : /* @__PURE__ */ (0, q.jsx)("span", {
		"aria-hidden": !0,
		className: l,
		children: c
	});
}
function Kt({ mode: e = "neutral", label: t, count: n, icon: r, badge: a, onModeChange: o, onClick: s, title: c, togglePosition: l = "left", interactive: u = !0, className: d }) {
	let f = !!o, p = r && /* @__PURE__ */ (0, q.jsx)(E, {
		...typeof r == "string" ? { name: r } : { icon: r },
		className: "text-sm"
	}), m = /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
		n !== void 0 && /* @__PURE__ */ (0, q.jsx)("span", {
			className: i("inline-flex items-center justify-center min-w-[16px] h-[16px] px-1 rounded-full text-[10px] font-bold text-white", a ?? "bg-muted-foreground"),
			children: n
		}),
		p,
		/* @__PURE__ */ (0, q.jsx)("span", {
			className: "truncate",
			children: t
		})
	] });
	if (f) {
		let r = /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
			p,
			/* @__PURE__ */ (0, q.jsx)("span", {
				className: "truncate",
				children: t
			}),
			n !== void 0 && /* @__PURE__ */ (0, q.jsx)("span", {
				className: "shrink-0 text-[10px] tabular-nums text-muted-foreground",
				"data-pill-count": "",
				children: n
			})
		] }), a = /* @__PURE__ */ (0, q.jsx)(Gt, {
			mode: e,
			onChange: o,
			interactive: u,
			...c ? { ariaLabel: c } : {}
		}), s = /* @__PURE__ */ (0, q.jsx)("span", {
			className: "inline-flex min-w-0 items-center gap-1.5 text-xs text-foreground",
			children: r
		});
		return /* @__PURE__ */ (0, q.jsx)("span", {
			className: i("inline-flex items-center gap-1.5 select-none", d),
			title: c,
			children: l === "right" ? /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [s, a] }) : /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [a, s] })
		});
	}
	return /* @__PURE__ */ (0, q.jsxs)("button", {
		type: "button",
		onClick: s,
		title: c,
		className: i("inline-flex items-center gap-1.5 text-xs px-2 py-0.5 rounded-full border transition-colors", Y(e), d),
		children: [/* @__PURE__ */ (0, q.jsx)(qt, { mode: e }), m]
	});
}
function qt({ mode: e }) {
	return e === "include" ? /* @__PURE__ */ (0, q.jsx)(E, {
		icon: j,
		className: "text-xs"
	}) : e === "exclude" ? /* @__PURE__ */ (0, q.jsx)(E, {
		icon: $e,
		className: "text-xs"
	}) : e === "active" ? /* @__PURE__ */ (0, q.jsx)("span", { className: "w-2 h-2 rounded-full bg-current" }) : /* @__PURE__ */ (0, q.jsx)("span", { className: "w-2 h-2 rounded-full bg-current opacity-30" });
}
function Jt(e) {
	return e === "include" ? "exclude" : e === "exclude" ? "neutral" : "include";
}
// ../../../clicky-ui/packages/ui/src/components/filter-bar-context.ts
var Yt = (0, K.createContext)({
	autoSubmit: !0,
	compact: !1
});
// ../../../clicky-ui/packages/ui/src/components/filter-bar-field-utils.tsx
function Xt(e, t) {
	if (t.length === 0) return e;
	let n = new Set(e.map((e) => e.value)), r = [...e];
	for (let e of t) n.has(e.value) || (n.add(e.value), r.push(e));
	return r;
}
function Zt(e) {
	let t = typeof e.label == "string" ? e.label : "";
	return [
		e.value,
		t,
		e.title ?? ""
	].filter(Boolean).join(" ");
}
function Qt(e, t) {
	let n = Object.values(t).filter((e) => e === "include").length, r = Object.values(t).filter((e) => e === "exclude").length;
	return n === 0 && r === 0 ? e : `${e} ${[n > 0 ? `+${n}` : null, r > 0 ? `-${r}` : null].filter(Boolean).join(" ")}`;
}
function $t(e, t) {
	let { autoSubmit: n } = (0, K.useContext)(Yt), [r, i] = (0, K.useState)(e), a = (0, K.useRef)(t), o = (0, K.useRef)(e), s = en(e);
	return (0, K.useEffect)(() => {
		a.current = t;
	}, [t]), (0, K.useEffect)(() => {
		tn(o.current, e) ? o.current = e : (o.current = e, i(e));
	}, [s, e]), (0, K.useEffect)(() => {
		if (tn(r, e)) return;
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
		s
	]), [r, i];
}
function en(e) {
	return Object.keys(e).sort().map((t) => `${t}:${e[t]}`).join("\0");
}
function tn(e, t) {
	let n = Object.keys(e), r = Object.keys(t);
	if (n.length !== r.length) return !1;
	for (let r of n) if (e[r] !== t[r]) return !1;
	return !0;
}
function nn(e, t, n) {
	let r = { ...e };
	return n === "neutral" ? (delete r[t], r) : ((n === "include" || n === "exclude") && (r[t] = n), r);
}
function rn(e) {
	return e === "include" ? "exclude" : e === "exclude" ? "neutral" : "include";
}
function an(e, t) {
	if (!(0, K.isValidElement)(e)) return e;
	let n = e;
	return (0, K.cloneElement)(n, { style: {
		width: `${t}px`,
		height: `${t}px`,
		...n.props.style
	} });
}
function on(e) {
	return e.icon == null ? { label: e.label } : {
		label: /* @__PURE__ */ (0, q.jsx)("span", {
			title: e.label,
			className: "inline-flex items-center text-[10px]",
			children: /* @__PURE__ */ (0, q.jsx)(U, {
				icon: an(e.icon, 10),
				className: "normal-case"
			})
		}),
		ariaLabel: e.label
	};
}
function sn(e) {
	return e ? "min-w-[12rem] max-w-[18rem] flex-1" : "min-w-[11rem] max-w-[15rem] shrink-0";
}
// ../../../clicky-ui/packages/ui/src/lib/combobox-values.ts
function cn(e, t) {
	return t > 0 ? { paddingLeft: `calc(${t}px + 0.75rem)` } : typeof e == "string" ? { paddingLeft: `${Math.min(e.length * .48 + .75, 7)}rem` } : { paddingLeft: "1.75rem" };
}
function ln(e, t) {
	let n = new Set(e.map((e) => e.value)), r = t.filter((e) => !n.has(e)).map((e) => ({
		value: e,
		label: e
	}));
	return r.length === 0 ? e : [...r, ...e];
}
function un(e, t) {
	let n = t.map((t) => {
		let n = e.find((e) => e.value === t);
		return n?.selectedLabel ?? n?.label ?? t;
	});
	return n.length <= 2 ? n.join(", ") : `${n.length} selected`;
}
function dn(e, t) {
	let n = /* @__PURE__ */ new Set([
		...t,
		"\n",
		"\r"
	]), r = [], i = "";
	for (let t of e) n.has(t) ? (r.push(i), i = "") : i += t;
	return r.push(i), r;
}
function fn(e) {
	let { allowCustomValue: t, query: n, multiple: r, tristate: i, onNew: a, choices: o, selectedValues: s } = e;
	return !t || !n || s.includes(n) || !i && !r && !a || o.some((e) => e.value === n) ? null : a ? a(n) : {
		value: n,
		label: `Add "${n}"`
	};
}
// ../../../clicky-ui/packages/ui/src/components/ComboboxActions.tsx
function pn({ disabled: e, loading: t, onClear: n, onToggle: r, showClear: i }) {
	return /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [i && /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		tabIndex: -1,
		"aria-label": "Clear",
		onClick: n,
		className: "absolute right-7 flex h-full items-center px-1 text-muted-foreground hover:text-foreground",
		children: /* @__PURE__ */ (0, q.jsx)(E, {
			icon: B,
			className: "text-xs"
		})
	}), t ? /* @__PURE__ */ (0, q.jsx)("span", {
		className: "pointer-events-none absolute right-2 top-1/2 -translate-y-1/2",
		children: /* @__PURE__ */ (0, q.jsxs)("svg", {
			className: "h-3.5 w-3.5 animate-spin text-muted-foreground",
			viewBox: "0 0 16 16",
			fill: "none",
			children: [/* @__PURE__ */ (0, q.jsx)("circle", {
				cx: "8",
				cy: "8",
				r: "6.5",
				stroke: "currentColor",
				strokeWidth: "2",
				opacity: "0.25"
			}), /* @__PURE__ */ (0, q.jsx)("path", {
				d: "M14.5 8a6.5 6.5 0 0 0-6.5-6.5",
				stroke: "currentColor",
				strokeWidth: "2",
				strokeLinecap: "round"
			})]
		})
	}) : /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		tabIndex: -1,
		disabled: e,
		"aria-label": "Toggle options",
		onClick: r,
		className: "pointer-events-auto absolute right-0 flex h-full items-center px-2 text-muted-foreground",
		children: /* @__PURE__ */ (0, q.jsx)(E, {
			icon: F,
			className: "text-xs"
		})
	})] });
}
// ../../../clicky-ui/packages/ui/src/components/ComboboxTags.tsx
function mn({ disabled: e, modes: t, onRemove: n, onSetMode: r, options: i, values: a }) {
	return a.map((a, o) => {
		let s = i.find((e) => e.value === a), c = s?.selectedLabel ?? s?.label ?? a, l = t?.[a], u = !e && /* @__PURE__ */ (0, q.jsx)("button", {
			type: "button",
			"aria-label": `Remove ${c}`,
			className: "shrink-0 text-muted-foreground hover:text-foreground",
			onMouseDown: (e) => e.preventDefault(),
			onClick: () => n(a),
			children: /* @__PURE__ */ (0, q.jsx)(E, {
				icon: B,
				className: "size-3"
			})
		});
		return l ? /* @__PURE__ */ (0, q.jsxs)("span", {
			"data-combobox-tag": a,
			"data-combobox-tag-mode": l,
			className: "inline-flex h-6 max-w-full items-center gap-1",
			children: [/* @__PURE__ */ (0, q.jsx)(Kt, {
				mode: l,
				label: /* @__PURE__ */ (0, q.jsxs)("span", {
					className: "inline-flex min-w-0 items-center gap-1.5",
					children: [/* @__PURE__ */ (0, q.jsx)(U, {
						icon: s?.icon,
						className: "shrink-0 text-muted-foreground"
					}), /* @__PURE__ */ (0, q.jsx)("span", {
						className: "truncate",
						children: c
					})]
				}),
				title: l === "include" ? `${c} included` : `${c} excluded`,
				className: "max-w-full",
				...e ? {} : { onClick: () => r(a, l === "include" ? "exclude" : "include") }
			}), u]
		}, `${a}-${o}`) : /* @__PURE__ */ (0, q.jsx)(hn, {
			value: a,
			...s ? { option: s } : {},
			children: u
		}, `${a}-${o}`);
	});
}
function hn({ value: e, option: t, children: n }) {
	return /* @__PURE__ */ (0, q.jsxs)("span", {
		"data-combobox-tag": e,
		className: "inline-flex h-6 max-w-full items-center gap-1 rounded-md bg-muted px-2 text-xs",
		children: [
			/* @__PURE__ */ (0, q.jsx)(U, {
				icon: t?.icon,
				className: "shrink-0 text-muted-foreground"
			}),
			/* @__PURE__ */ (0, q.jsx)("span", {
				className: "truncate",
				children: t?.selectedLabel ?? t?.label ?? e
			}),
			n
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/ComboboxControl.tsx
function gn({ anchorRef: e, ariaLabel: t, ariaRequired: n, closedLabel: r, describedBy: a, disabled: l, displayValue: u, effectivePrefix: d, highlighted: f, id: p, inputRef: m, invalid: h, label: g, labelRef: _, labelWidth: v, listId: y, loading: b, onClear: x, onInput: S, onKeyDown: C, onOpen: w, onPaste: T, onRemoveTag: E, onSetTagMode: D, onToggle: O, open: k, options: A, placeholder: j, showClear: M, size: N, suffix: P, tagModes: ee, tagValues: F, tags: I }) {
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		ref: e,
		"data-jsf-control": !0,
		className: i("relative flex items-center", I && "flex-wrap gap-1 rounded-md border border-input bg-background px-2 py-1 text-foreground focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-1", I && (N ? s[N] : "min-h-control-h"), I && (P ? M ? "pr-[5.5rem]" : "pr-14" : M ? "pr-14" : "pr-8"), I && h && "border-destructive focus-within:ring-destructive", I && l && "cursor-not-allowed opacity-50"),
		onMouseDown: (e) => {
			I && !e.target.closest("button, input") && (e.preventDefault(), m.current?.focus());
		},
		children: [
			g != null && /* @__PURE__ */ (0, q.jsx)("span", {
				ref: _,
				className: i("pointer-events-none whitespace-nowrap font-medium uppercase tracking-wide text-muted-foreground text-[10px]", I ? "shrink-0" : "absolute left-2 z-10"),
				children: g
			}),
			I && d,
			I && /* @__PURE__ */ (0, q.jsx)(mn, {
				disabled: l,
				modes: ee,
				onRemove: E,
				onSetMode: D,
				options: A,
				values: F
			}),
			/* @__PURE__ */ (0, q.jsx)("input", {
				ref: m,
				id: p,
				type: "text",
				role: "combobox",
				"aria-expanded": k,
				"aria-controls": y,
				"aria-activedescendant": f >= 0 ? `${y}-${f}` : void 0,
				"aria-label": t,
				"aria-invalid": h || void 0,
				"aria-required": n || void 0,
				"aria-describedby": a,
				autoComplete: "off",
				disabled: l,
				placeholder: I && F.length > 0 ? void 0 : j,
				value: u,
				...!I && !k && r ? { title: r } : {},
				onChange: (e) => S(e.target.value),
				onFocus: w,
				onClick: w,
				onKeyDown: C,
				onPaste: T,
				className: i(I ? "min-w-28 flex-1 bg-transparent px-1 py-1 text-foreground outline-none" : "w-full rounded-md border border-input bg-background text-foreground", I ? N ? c[N] : "text-sm" : N ? o[N] : "h-control-h px-control-px text-sm", "max-sm:text-base", !I && d && "pl-8", !I && (P ? M ? "pr-[5.5rem]" : "pr-14" : M ? "pr-14" : "pr-8"), !I && "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1", !I && "disabled:cursor-not-allowed disabled:opacity-50", !I && h && "border-destructive focus-visible:ring-destructive"),
				style: !I && g != null ? cn(g, v) : void 0
			}),
			!I && d && /* @__PURE__ */ (0, q.jsx)("div", {
				className: "absolute inset-y-0 left-1.5 flex items-center",
				children: d
			}),
			P && /* @__PURE__ */ (0, q.jsx)("div", {
				className: i("absolute flex h-full items-center", M ? "right-[3.75rem]" : "right-7"),
				children: P
			}),
			/* @__PURE__ */ (0, q.jsx)(pn, {
				disabled: l,
				loading: b,
				onClear: x,
				onToggle: O,
				showClear: M
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/ComboboxMenuTrigger.tsx
function _n({ anchorRef: e, triggerRef: t, content: n, className: r, disabled: a, open: o, listId: s, ariaLabel: c, title: l, onToggle: u }) {
	return /* @__PURE__ */ (0, q.jsx)("div", {
		ref: e,
		className: "h-full min-w-0",
		children: /* @__PURE__ */ (0, q.jsxs)("button", {
			ref: t,
			type: "button",
			role: "combobox",
			"aria-label": c,
			"aria-haspopup": "listbox",
			"aria-expanded": o,
			"aria-controls": o ? s : void 0,
			title: l,
			disabled: a,
			onClick: u,
			onKeyDown: (e) => {
				(e.key === "ArrowDown" || e.key === "ArrowUp") && (e.preventDefault(), o || u());
			},
			className: i("inline-flex h-control-h w-full min-w-0 items-center gap-1.5 rounded-md border border-input bg-background px-control-px text-sm hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50", r),
			children: [n, /* @__PURE__ */ (0, q.jsx)(E, {
				icon: F,
				className: "ml-auto size-3 shrink-0 text-muted-foreground"
			})]
		})
	});
}
function vn({ inputRef: e, value: t, ariaLabel: n, listId: r, highlighted: i, onChange: a, onKeyDown: o }) {
	return (0, K.useEffect)(() => {
		e.current?.focus();
	}, [e]), /* @__PURE__ */ (0, q.jsx)("div", {
		className: "sticky top-0 z-10 border-b border-border bg-popover p-1",
		children: /* @__PURE__ */ (0, q.jsx)("input", {
			ref: e,
			type: "text",
			"aria-label": `Search ${n ?? "options"}`,
			"aria-controls": r,
			"aria-activedescendant": i >= 0 ? `${r}-${i}` : void 0,
			value: t,
			onChange: (e) => a(e.target.value),
			onKeyDown: o,
			placeholder: "Search…",
			className: "h-control-h w-full rounded-sm bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
		})
	});
}
var yn = z("sm"), bn = 8, xn = 4, X = 16;
function Sn({ anchor: e, viewportWidth: t, viewportHeight: n, mobile: r, naturalHeight: i, preferredWidth: a }) {
	if (a !== void 0 && (!Number.isFinite(a) || a <= 0)) throw Error("Combobox menu width must be a positive finite number.");
	if (r) {
		let r = Math.max(0, n - e.bottom - xn - bn), a = Math.max(0, e.top - xn - bn), o = i != null && i > r && a > r, s = Math.max(0, t - 32);
		return Cn({
			strategy: "mobile",
			anchor: e,
			viewportHeight: n,
			openUp: o,
			left: X,
			width: s,
			maxWidth: s,
			maxHeight: o ? a : r
		});
	}
	let o = Math.max(0, n - e.bottom - xn - bn), s = Math.max(0, e.top - xn - bn), c = i != null && i > o && s > o, l = a === void 0 ? void 0 : Math.min(Math.max(e.width, a), Math.max(0, t - 16));
	return Cn({
		strategy: "desktop",
		anchor: e,
		viewportHeight: n,
		openUp: c,
		left: l === void 0 ? e.left : Math.max(bn, Math.min(e.left, t - l - bn)),
		width: l ?? e.width,
		maxWidth: l ?? Math.max(e.width, Math.min(400, t - e.left - bn)),
		maxHeight: c ? s : o
	});
}
function Cn({ strategy: e, anchor: t, viewportHeight: n, openUp: r, left: i, width: a, maxWidth: o, maxHeight: s }) {
	return {
		strategy: e,
		...r ? { bottom: n - t.top + xn } : { top: t.bottom + xn },
		left: i,
		width: a,
		maxWidth: o,
		maxHeight: s
	};
}
// ../../../clicky-ui/packages/ui/src/components/useComboboxLayout.ts
function wn(e, t) {
	let [n, r] = (0, K.useState)(0);
	return (0, K.useLayoutEffect)(() => {
		e == null ? r(0) : t.current && r(t.current.offsetWidth);
	}, [e, t]), n;
}
function Tn({ open: e, anchorRef: t, listRef: n, menuWidth: r }) {
	let [i, a] = (0, K.useState)(null), o = (0, K.useCallback)(() => {
		let i = t.current;
		if (!e || !i) return;
		let o = Sn({
			anchor: i.getBoundingClientRect(),
			viewportWidth: window.innerWidth,
			viewportHeight: window.innerHeight,
			mobile: V(yn),
			...r === void 0 ? {} : { preferredWidth: r },
			...n.current ? { naturalHeight: n.current.scrollHeight } : {}
		});
		a((e) => Z(e, o) ? e : o);
	}, [
		t,
		n,
		e,
		r
	]);
	return (0, K.useLayoutEffect)(() => {
		if (!e) a(null);
		else return o(), window.addEventListener("scroll", o, !0), window.addEventListener("resize", o), () => {
			window.removeEventListener("scroll", o, !0), window.removeEventListener("resize", o);
		};
	}, [e, o]), (0, K.useLayoutEffect)(() => {
		let t = n.current;
		if (!e || !i || !t) return;
		o();
		let r = typeof ResizeObserver > "u" ? null : new ResizeObserver(o), a = new MutationObserver(o);
		return r?.observe(t), a.observe(t, {
			childList: !0,
			characterData: !0,
			subtree: !0
		}), () => {
			r?.disconnect(), a.disconnect();
		};
	}, [
		n,
		e,
		i,
		o
	]), i;
}
function Z(e, t) {
	return e?.strategy === t.strategy && e.top === t.top && e.bottom === t.bottom && e.left === t.left && e.width === t.width && e.maxWidth === t.maxWidth && e.maxHeight === t.maxHeight;
}
// ../../../clicky-ui/packages/ui/src/components/ComboboxMenu.tsx
function En({ customEntry: e, filtered: t, floatingZ: n, footer: r, header: a, ariaLabel: o, hasOptions: s, highlighted: c, isSelected: l, listId: u, listRef: d, loading: f, menuPos: p, modes: m, multiple: h, onHighlight: g, onSelect: _, onSetMode: v, tristate: y }) {
	let b = p.strategy === "mobile", x = /* @__PURE__ */ (0, q.jsxs)("div", {
		id: a ? void 0 : u,
		ref: d,
		role: a ? void 0 : "listbox",
		"aria-label": a ? void 0 : o,
		"aria-multiselectable": !a && h ? !0 : void 0,
		style: {
			position: b ? "absolute" : "fixed",
			...p.top == null ? { bottom: p.bottom } : { top: p.top },
			left: p.left,
			...b ? { width: p.width } : { minWidth: p.width },
			maxWidth: p.maxWidth,
			maxHeight: p.maxHeight,
			zIndex: b ? void 0 : n,
			pointerEvents: b ? "auto" : void 0
		},
		className: i("overflow-auto rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-lg shadow-black/5", !b && "w-max"),
		children: [a, /* @__PURE__ */ (0, q.jsxs)("div", {
			id: a ? u : void 0,
			role: a ? "listbox" : void 0,
			"aria-label": a ? o : void 0,
			"aria-multiselectable": a && h ? !0 : void 0,
			children: [
				f && t.length === 0 && /* @__PURE__ */ (0, q.jsx)("div", {
					className: "px-2 py-4 text-center text-sm text-muted-foreground",
					children: "Loading…"
				}),
				!f && t.length === 0 && !e && s && /* @__PURE__ */ (0, q.jsx)("div", {
					className: "px-2 py-4 text-center text-sm text-muted-foreground",
					children: "No results"
				}),
				t.map((e, n) => {
					let r = l(e.value), a = y ? m[e.value] : void 0, o = e.group != null && e.group !== t[n - 1]?.group, s = /* @__PURE__ */ (0, q.jsx)("span", {
						"data-slot": "combobox-selection-indicator",
						"aria-hidden": !0,
						className: i("shrink-0 text-xs", e.description && e.descriptionLayout !== "inline" && "mt-1", r ? "opacity-100" : "opacity-0"),
						children: /* @__PURE__ */ (0, q.jsx)(E, { icon: de })
					});
					return /* @__PURE__ */ (0, q.jsxs)(K.Fragment, { children: [o && /* @__PURE__ */ (0, q.jsx)("div", {
						role: "presentation",
						className: "select-none px-2 pb-1 pt-2 text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
						children: e.group
					}), /* @__PURE__ */ (0, q.jsx)("div", {
						id: u ? `${u}-${n}` : void 0,
						role: "option",
						"aria-selected": y ? a != null : r,
						"aria-disabled": e.disabled,
						...a == null ? {} : { "aria-label": `${e.label}, ${a === "include" ? "included" : "excluded"}` },
						...e.title === void 0 ? {} : { title: e.title },
						...y ? { "data-filter-option": e.value } : {},
						onMouseDown: (t) => {
							t.preventDefault(), !y && !e.disabled && _(e);
						},
						...y ? { onClick: () => {
							e.disabled || _(e);
						} } : {},
						onMouseEnter: () => g(n),
						className: i("flex cursor-pointer rounded-sm py-1.5 text-sm", h ? "gap-2 px-2" : "gap-1.5 px-1.5", e.description && e.descriptionLayout !== "inline" ? "items-start" : "items-center", n === c && "bg-accent", r && !y && "font-medium", e.disabled && "cursor-not-allowed opacity-50"),
						children: y ? /* @__PURE__ */ (0, q.jsxs)("div", {
							className: "min-w-0 flex-1",
							children: [/* @__PURE__ */ (0, q.jsx)(Kt, {
								mode: a ?? "neutral",
								togglePosition: "right",
								interactive: !1,
								label: /* @__PURE__ */ (0, q.jsxs)("span", {
									className: "inline-flex min-w-0 items-center gap-1.5",
									children: [/* @__PURE__ */ (0, q.jsx)(U, {
										icon: e.icon,
										className: "shrink-0 text-sm text-muted-foreground"
									}), /* @__PURE__ */ (0, q.jsx)("span", {
										className: "truncate",
										children: e.label
									})]
								}),
								onModeChange: (t) => v(e.value, t),
								className: "w-full justify-between",
								...e.count === void 0 ? {} : { count: e.count }
							}), /* @__PURE__ */ (0, q.jsx)(On, { description: e.description })]
						}) : /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
							h && s,
							/* @__PURE__ */ (0, q.jsx)(U, {
								icon: e.icon,
								className: i("text-sm text-muted-foreground", e.description && e.descriptionLayout !== "inline" && "mt-1")
							}),
							/* @__PURE__ */ (0, q.jsx)(Dn, { option: e }),
							e.trailing != null && /* @__PURE__ */ (0, q.jsx)("span", {
								className: "ml-auto min-w-0 max-w-[55%] truncate pl-2 text-right text-xs text-muted-foreground",
								children: e.trailing
							}),
							!h && s
						] })
					})] }, e.value);
				}),
				e && /* @__PURE__ */ (0, q.jsxs)("div", {
					id: u ? `${u}-${t.length}` : void 0,
					role: "option",
					"aria-selected": !1,
					"data-filter-add-custom": e.value,
					title: e.title,
					onMouseDown: (e) => e.preventDefault(),
					onClick: () => _(e),
					onMouseEnter: () => g(t.length),
					className: i("flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-primary", c === t.length && "bg-accent"),
					children: [/* @__PURE__ */ (0, q.jsx)(U, {
						icon: e.icon,
						className: "text-sm text-muted-foreground"
					}), /* @__PURE__ */ (0, q.jsx)("span", {
						className: "min-w-0 truncate",
						children: e.label
					})]
				}),
				r != null && /* @__PURE__ */ (0, q.jsx)("div", {
					className: "select-none px-2 py-1.5 text-[11px] text-muted-foreground",
					children: r
				})
			]
		})]
	});
	return (0, Ge.createPortal)(b ? /* @__PURE__ */ (0, q.jsx)("div", {
		"data-slot": "combobox-viewport-layer",
		style: {
			position: "fixed",
			inset: 0,
			zIndex: n,
			pointerEvents: "none"
		},
		children: x
	}) : x, document.body);
}
function Dn({ option: e }) {
	return /* @__PURE__ */ (0, q.jsxs)("span", {
		"data-slot": "combobox-option-text",
		className: i("min-w-0 flex-1", e.descriptionLayout === "inline" && "flex items-center gap-2"),
		children: [/* @__PURE__ */ (0, q.jsx)("span", {
			className: "block min-w-0 truncate",
			children: e.label
		}), /* @__PURE__ */ (0, q.jsx)(On, {
			description: e.description,
			inline: e.descriptionLayout === "inline"
		})]
	});
}
function On({ description: e, inline: t = !1 }) {
	return e ? /* @__PURE__ */ (0, q.jsx)("span", {
		className: i("text-xs font-normal leading-4 text-muted-foreground", t ? "min-w-0 max-w-[55%] shrink-0 truncate" : "mt-0.5 block whitespace-pre-line"),
		children: e
	}) : null;
}
// ../../../clicky-ui/packages/ui/src/components/Combobox.tsx
function kn(e) {
	let { options: t, placeholder: n, label: r, ariaLabel: a, disabled: o, required: s, invalid: c, ariaRequired: l, describedBy: u, allowCustomValue: d = !0, onNew: f, onCreate: p, id: m, size: h, className: g, loading: _, onSearch: v, onKeyDown: y, suffix: b, prefix: x, footer: S } = e, C = e.multiple === !0, w = !C && e.searchPlacement === "menu", T = (0, K.useId)(), E = (0, K.useRef)(null), D = e.multiple === !0 && e.tristate === !0, O = e.multiple === !0 && e.variant === "tags", k = e.multiple === !0 && e.tristate !== !0 ? e.separators ?? [] : [], A = (0, K.useMemo)(() => e.multiple === !0 && e.tristate === !0 ? e.value : {}, [
		e.multiple,
		e.tristate,
		e.value
	]), j = (0, K.useMemo)(() => e.multiple === !0 && e.tristate === !0 ? Object.keys(e.value) : e.multiple === !0 ? e.value : e.value ? [e.value] : [], [
		e.multiple,
		e.tristate,
		e.value
	]), M = (0, K.useRef)(null), N = (0, K.useRef)(null), P = (0, K.useRef)(null), ee = (0, K.useRef)(null), F = (0, K.useRef)(null), [I, L] = (0, K.useState)(!1), te = re(), ne = Tn({
		open: I,
		anchorRef: N,
		listRef: ee,
		...e.menuWidth === void 0 ? {} : { menuWidth: e.menuWidth }
	}), [R, z] = (0, K.useState)(""), [B, V] = (0, K.useState)(-1), ie = wn(r, F), ae = (e) => j.includes(e), oe = (0, K.useMemo)(() => {
		if (C) return;
		let e = j[0];
		if (e) return t.find((t) => t.value === e) ?? (d ? f?.(e) ?? void 0 : void 0);
	}, [
		d,
		C,
		f,
		t,
		j
	]), se = (0, K.useMemo)(() => {
		if (D) {
			let e = Object.values(A), t = e.filter((e) => e === "include").length, n = e.length - t;
			return [t > 0 && `+${t}`, n > 0 && `-${n}`].filter(Boolean).join(" ");
		}
		if (C) return un(t, j);
		let e = j[0] ?? "";
		return oe?.selectedLabel ?? oe?.label ?? e;
	}, [
		D,
		A,
		C,
		t,
		oe,
		j
	]), ce = x ?? (oe?.icon == null ? null : /* @__PURE__ */ (0, q.jsx)(U, {
		icon: oe.icon,
		className: "size-4"
	})), le = O || I ? R : se, ue = (0, K.useMemo)(() => {
		if (v) return ln(t, j);
		let e = R.toLowerCase().trim();
		return e ? t.filter((t) => t.label.toLowerCase().includes(e) || t.description?.toLowerCase().includes(e) || t.group?.toLowerCase().includes(e) || t.value.toLowerCase().includes(e)) : t;
	}, [
		v,
		t,
		R,
		j
	]);
	(0, K.useEffect)(() => {
		if (!v || !I) return;
		let e = setTimeout(() => v(R.trim()), 250);
		return () => clearTimeout(e);
	}, [
		v,
		I,
		R
	]);
	let de = R.trim(), fe = (0, K.useMemo)(() => w && ue.length > 0 ? null : fn({
		allowCustomValue: d,
		query: de,
		multiple: C,
		tristate: D,
		onNew: f,
		choices: t,
		selectedValues: j
	}), [
		d,
		w,
		ue.length,
		C,
		f,
		t,
		j,
		de,
		D
	]), H = (0, K.useMemo)(() => fe ? [...ue, fe] : ue, [ue, fe]);
	(0, K.useEffect)(() => {
		V(-1);
	}, [R]), (0, K.useEffect)(() => {
		B >= H.length && V(-1);
	}, [B, H.length]), (0, K.useEffect)(() => {
		if (!I) return;
		let e = (e) => {
			let t = e.target;
			!M.current?.contains(t) && !ee.current?.contains(t) && xe();
		};
		return document.addEventListener("mousedown", e), document.addEventListener("focusin", e), () => {
			document.removeEventListener("mousedown", e), document.removeEventListener("focusin", e);
		};
	});
	function pe(t) {
		e.multiple === !0 && e.tristate !== !0 ? e.onChange(t) : e.multiple !== !0 && e.onChange(t[0] ?? "");
	}
	function me(t) {
		e.multiple === !0 && e.tristate === !0 && e.onChange(t);
	}
	function W(e, t) {
		let n = { ...A };
		t === "include" || t === "exclude" ? n[e] = t : delete n[e], me(n);
	}
	function he(e) {
		let t = A[e];
		W(e, t === void 0 ? "include" : t === "include" ? "exclude" : "neutral");
	}
	function ge() {
		I || (z(""), V(-1), L(!0));
	}
	function _e(e) {
		return p ? p(e) : e.value;
	}
	function ve(e) {
		D ? W(e, "include") : C ? j.includes(e) || pe([...j, e]) : pe([e]);
	}
	function ye(e) {
		let n = [...j];
		for (let r of e) {
			let e = r.trim();
			if (!e || n.includes(e)) continue;
			let i = t.find((t) => t.value === e), a = i ?? f?.(e) ?? {
				value: e,
				label: e
			}, o = i ? i.value : _e(a);
			o !== null && !n.includes(o) && n.push(o);
		}
		n.length !== j.length && pe(n), z(""), V(-1);
	}
	function be(e) {
		if (!O || !d || k.length === 0) return;
		let t = dn(e.clipboardData.getData("text"), k);
		t.length < 2 || (e.preventDefault(), ye(t));
	}
	function xe() {
		if (w) {
			z(""), L(!1), V(-1);
			return;
		}
		let e = R.trim();
		if (fe) {
			let e = _e(fe);
			e !== null && ve(e);
		} else if (!C && d && e && e !== j[0]) {
			let n = t.find((t) => t.value === e), r = n ?? (f ? f(e) : {
				value: e,
				label: e
			}), i = r ? n ? n.value : _e(r) : null;
			i !== null && pe([i]);
		}
		z(""), L(!1), V(-1);
	}
	function Se(e) {
		if (e === fe) {
			let t = _e(e);
			t !== null && ve(t), z(""), V(-1), C || L(!1), (w ? E : P).current?.focus();
		} else D ? (he(e.value), P.current?.focus()) : C ? (pe(ae(e.value) ? j.filter((t) => t !== e.value) : [...j, e.value]), V(-1), P.current?.focus()) : (pe([e.value]), z(""), L(!1), V(-1), (w ? E : P).current?.focus());
	}
	function Ce() {
		j.length > 0 && (D ? me({}) : pe([])), z(""), V(-1), P.current?.focus(), L(!1);
	}
	function we(e) {
		D ? W(e, "neutral") : pe(j.filter((t) => t !== e)), P.current?.focus();
	}
	function Te(e) {
		let t = ee.current;
		t && t.querySelectorAll("[role=\"option\"]")[e]?.scrollIntoView?.({ block: "nearest" });
	}
	function Ee(e) {
		if (O && d && k.includes(e.key)) e.preventDefault(), ye([R]);
		else if (O && e.key === "Backspace" && R === "" && j.length > 0) {
			e.preventDefault();
			let t = j[j.length - 1];
			D ? t && W(t, "neutral") : pe(j.slice(0, -1));
		} else if (e.key === "ArrowDown") {
			if (e.preventDefault(), !I) {
				ge();
				return;
			}
			let t = B < H.length - 1 ? B + 1 : 0;
			V(t), Te(t);
		} else if (e.key === "ArrowUp") {
			if (e.preventDefault(), !I) return;
			let t = B > 0 ? B - 1 : H.length - 1;
			V(t), Te(t);
		} else if (e.key === "Enter") {
			if (w && e.preventDefault(), I && B >= 0 && H[B] && !H[B].disabled) {
				e.preventDefault(), Se(H[B]);
				return;
			}
			if (I && (C || w) && fe) {
				e.preventDefault(), Se(fe);
				return;
			}
			xe(), y?.(e);
		} else e.key === "Escape" ? I && (e.preventDefault(), w && e.stopPropagation(), z(""), L(!1), V(-1), w && E.current?.focus()) : y?.(e);
	}
	let De = `${m ?? T}-listbox`, Oe = a ?? (typeof r == "string" ? r : void 0), ke = !s && !_ && !o && j.length > 0;
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		ref: M,
		className: i("relative", g),
		children: [w ? /* @__PURE__ */ (0, q.jsx)(_n, {
			anchorRef: N,
			triggerRef: E,
			content: e.triggerContent ?? /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [ce, /* @__PURE__ */ (0, q.jsx)("span", {
				className: "min-w-0 truncate",
				children: se || n
			})] }),
			className: e.triggerClassName,
			disabled: o,
			open: I,
			listId: De,
			ariaLabel: Oe,
			title: e.title,
			onToggle: () => I ? xe() : ge()
		}) : /* @__PURE__ */ (0, q.jsx)(gn, {
			anchorRef: N,
			ariaLabel: Oe,
			ariaRequired: l,
			closedLabel: se,
			describedBy: u,
			disabled: o,
			displayValue: le,
			effectivePrefix: ce,
			highlighted: B,
			id: m,
			inputRef: P,
			invalid: c,
			label: r,
			labelRef: F,
			labelWidth: ie,
			listId: De,
			loading: _,
			onClear: Ce,
			onInput: (e) => {
				z(e), I || L(!0);
			},
			onKeyDown: Ee,
			onOpen: ge,
			onPaste: be,
			onRemoveTag: we,
			onToggle: () => {
				I ? xe() : (ge(), P.current?.focus());
			},
			open: I,
			options: t,
			placeholder: n,
			showClear: ke,
			size: h,
			suffix: b,
			tagModes: D ? A : void 0,
			onSetTagMode: W,
			tagValues: j,
			tags: O
		}), I && ne && typeof document < "u" && /* @__PURE__ */ (0, q.jsx)(En, {
			customEntry: fe,
			filtered: ue,
			floatingZ: te,
			footer: S,
			ariaLabel: Oe,
			header: w ? /* @__PURE__ */ (0, q.jsx)(vn, {
				inputRef: P,
				value: R,
				ariaLabel: Oe,
				listId: De,
				highlighted: B,
				onChange: z,
				onKeyDown: Ee
			}) : void 0,
			hasOptions: t.length > 0,
			highlighted: B,
			isSelected: ae,
			listId: De,
			listRef: ee,
			loading: _,
			menuPos: ne,
			modes: A,
			multiple: C,
			onHighlight: V,
			onSelect: Se,
			onSetMode: W,
			tristate: D
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/components/filter-bar-multi.tsx
function An({ filter: e, grow: t }) {
	let [n, r] = $t(e.value, e.onChange), [a, o] = (0, K.useState)([]), [s, c] = (0, K.useState)(!1), [l, u] = (0, K.useState)(""), d = (0, K.useRef)(""), f = e.onSearch, p = (0, K.useCallback)((e) => {
		d.current = e, u(e), e ? (c(!0), Promise.resolve(f?.(e)).then((t) => {
			d.current === e && (c(!1), o(Array.isArray(t) ? t : []));
		}).catch(() => {
			d.current === e && (c(!1), o([]));
		})) : (o([]), c(!1));
	}, [f]), m = !!e.allowCustomValue, h = (0, K.useMemo)(() => {
		if (!m) return e.options;
		let t = new Set(e.options.map((e) => e.value));
		return Xt(Object.keys(n).filter((e) => !t.has(e)).map((e) => ({
			value: e,
			label: e
		})), e.options);
	}, [
		m,
		e.options,
		n
	]), g = (0, K.useMemo)(() => !f || !l ? h : Xt(h.filter((e) => n[e.value] !== void 0), a), [
		f,
		l,
		h,
		a,
		n
	]), _ = e.truncated && e.total ? Math.max(e.total - e.options.length, 0) : 0, v = _ > 0 && !l ? /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
		"… and ",
		_.toLocaleString(),
		" more",
		f ? " — type to search all" : ""
	] }) : void 0;
	return /* @__PURE__ */ (0, q.jsx)(kn, {
		multiple: !0,
		tristate: !0,
		...on(e),
		options: jn(g),
		value: n,
		onChange: r,
		allowCustomValue: m,
		size: "sm",
		className: i(sn(t), e.className),
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...f === void 0 ? {} : {
			onSearch: p,
			loading: s
		},
		...e.disabled === void 0 ? {} : { disabled: e.disabled },
		...v === void 0 ? {} : { footer: v }
	});
}
function jn(e) {
	return e.map((e) => {
		let t = e.title ?? Zt(e);
		return {
			value: e.value,
			label: typeof e.label == "string" ? e.label : t,
			title: t,
			...e.description === void 0 ? {} : { description: e.description },
			...e.selectedLabel === void 0 ? {} : { selectedLabel: e.selectedLabel },
			...e.icon === void 0 ? {} : { icon: e.icon },
			...e.disabled === void 0 ? {} : { disabled: e.disabled },
			...e.count === void 0 ? {} : { count: e.count }
		};
	});
}
function Mn({ filter: e, chrome: t = "full" }) {
	let [n, r] = (0, K.useState)(""), [a, o] = $t(e.value, e.onChange), s = e.options.length > 7, c = t === "embedded", l = (0, K.useMemo)(() => {
		let t = n.trim().toLowerCase();
		return t ? e.options.filter((e) => Zt(e).toLowerCase().includes(t)) : e.options;
	}, [e.options, n]);
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		"data-filter-panel-chrome": t,
		className: i("min-w-[18rem] max-w-[22rem] text-popover-foreground", c ? "p-0" : "rounded-md border border-border bg-popover p-2 shadow-sm shadow-black/5"),
		children: [
			!c && /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "mb-2 flex items-center justify-between gap-2",
				children: [/* @__PURE__ */ (0, q.jsx)("div", {
					className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
					children: e.label
				}), /* @__PURE__ */ (0, q.jsx)("button", {
					type: "button",
					className: "text-[10px] text-primary disabled:text-muted-foreground",
					onClick: () => o({}),
					disabled: Object.keys(a).length === 0,
					children: "Clear all"
				})]
			}),
			s && /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "mb-2 flex items-center gap-2 rounded-md border border-input bg-background px-2",
				children: [/* @__PURE__ */ (0, q.jsx)(E, {
					icon: Se,
					className: "shrink-0 text-muted-foreground"
				}), /* @__PURE__ */ (0, q.jsx)("input", {
					type: "search",
					"aria-label": `Filter ${e.label} options`,
					className: "h-8 min-w-0 flex-1 bg-transparent text-sm outline-none placeholder:text-placeholder",
					placeholder: `Filter ${e.label.toLowerCase()}`,
					value: n,
					onChange: (e) => r(e.target.value)
				})]
			}),
			/* @__PURE__ */ (0, q.jsxs)("div", {
				className: "max-h-72 space-y-0.5 overflow-auto",
				children: [l.map((e) => {
					let t = a[e.value] ?? "neutral", n = e.title ?? Zt(e);
					return /* @__PURE__ */ (0, q.jsx)("div", {
						role: "button",
						tabIndex: 0,
						"data-filter-option": e.value,
						className: "rounded-md px-1.5 py-0.5 hover:bg-accent/50 focus-visible:bg-accent/50 focus-visible:outline-none",
						onClick: () => o(nn(a, e.value, rn(t))),
						onKeyDown: (n) => {
							(n.key === "Enter" || n.key === " ") && (n.preventDefault(), o(nn(a, e.value, rn(t))));
						},
						children: /* @__PURE__ */ (0, q.jsx)(Kt, {
							className: "w-full justify-between",
							label: /* @__PURE__ */ (0, q.jsxs)("span", {
								className: "inline-flex min-w-0 items-start gap-1.5",
								children: [/* @__PURE__ */ (0, q.jsx)(U, {
									icon: e.icon,
									className: "mt-0.5 shrink-0 text-sm text-muted-foreground"
								}), /* @__PURE__ */ (0, q.jsxs)("span", {
									className: "min-w-0 text-left",
									children: [/* @__PURE__ */ (0, q.jsx)("span", {
										className: "block truncate",
										children: e.label
									}), e.description && /* @__PURE__ */ (0, q.jsx)("span", {
										className: "mt-0.5 block truncate text-xs font-normal text-muted-foreground",
										children: e.description
									})]
								})]
							}),
							mode: t,
							title: n,
							togglePosition: "right",
							...e.count === void 0 ? {} : { count: e.count },
							onModeChange: (t) => o(nn(a, e.value, t))
						})
					}, e.value);
				}), l.length === 0 && /* @__PURE__ */ (0, q.jsx)("div", {
					className: "px-2 py-3 text-sm text-muted-foreground",
					children: "No options found"
				})]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/filter-bar-tree-model.ts
function Nn(e) {
	let t = /* @__PURE__ */ new Map(), n = (e) => {
		t.set(e.key, e.label), e.children?.forEach(n);
	};
	return e.forEach(n), t;
}
function Pn(e, t) {
	let n = Nn(e), r = (e) => Object.keys(t).filter((n) => t[n] === e).map((e) => n.get(e) ?? e);
	return [...r("include"), ...r("exclude").map((e) => `−${e}`)].join(", ");
}
function Fn(e, t) {
	return (e.children ?? []).some((e) => t[e.key] !== void 0 || Fn(e, t));
}
// ../../../clicky-ui/packages/ui/src/components/TreePickerField.tsx
var In = 480, Ln = 480, Rn = 200, zn = 8;
function Bn({ roots: e, getKey: t, getChildren: n, getAriaLabel: r, renderRow: a, getSearchText: s, defaultOpen: c, hasMoreChildren: l, loadChildren: u, isSelectable: d, onSelect: f, closeOnSelect: p = !0, selected: m, revealSelected: h, showControls: g, ariaLabel: _, empty: v, label: y, placeholder: b = "Select…", disabled: S, size: C, className: w, triggerClassName: T, panelClassName: D, renderTrigger: O, renderFooter: k, triggerAriaLabel: A }) {
	let [j, M] = (0, K.useState)(!1), N = re(), [P, ee] = (0, K.useState)(null), I = (0, K.useRef)(null), L = (0, K.useRef)(null), te = (0, K.useRef)(null);
	return x(j, () => {
		M(!1), L.current?.focus();
	}), (0, K.useLayoutEffect)(() => {
		if (!j) {
			ee(null);
			return;
		}
		let e = () => {
			let e = I.current;
			if (!e) return;
			let t = e.getBoundingClientRect(), n = window.innerWidth - t.left - zn, r = Math.max(t.width, Math.min(In, n)), i = window.innerHeight - t.bottom - zn - 4, a = t.top - zn - 4, o = i < Rn && a > i, s = Math.min(Ln, Math.max(o ? a : i, Rn));
			ee({
				top: o ? Math.max(zn, t.top - 4 - s) : t.bottom + 4,
				left: t.left,
				width: t.width,
				maxWidth: r,
				maxHeight: s
			});
		};
		return e(), window.addEventListener("scroll", e, !0), window.addEventListener("resize", e), () => {
			window.removeEventListener("scroll", e, !0), window.removeEventListener("resize", e);
		};
	}, [j]), (0, K.useEffect)(() => {
		if (!j) return;
		let e = (e) => {
			let t = e.target;
			!I.current?.contains(t) && !te.current?.contains(t) && M(!1);
		};
		return document.addEventListener("mousedown", e), () => document.removeEventListener("mousedown", e);
	}, [j]), /* @__PURE__ */ (0, q.jsxs)("div", {
		ref: I,
		className: i("relative", w),
		children: [O ? O({
			open: j,
			disabled: !!S,
			triggerRef: L,
			toggle: () => {
				S || M((e) => !e);
			}
		}) : /* @__PURE__ */ (0, q.jsxs)("button", {
			ref: L,
			type: "button",
			disabled: S,
			"aria-label": A,
			"aria-haspopup": "tree",
			"aria-expanded": j,
			onClick: () => M((e) => !e),
			className: i("relative flex w-full items-center rounded-md border border-input bg-background text-left text-foreground", C ? o[C] : "h-control-h px-control-px text-sm", "pr-8", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1", "disabled:cursor-not-allowed disabled:opacity-50", T),
			children: [/* @__PURE__ */ (0, q.jsx)("span", {
				className: i("min-w-0 flex-1 truncate", y == null && "text-muted-foreground"),
				children: y ?? b
			}), /* @__PURE__ */ (0, q.jsx)("span", {
				className: "pointer-events-none absolute right-0 flex h-full items-center px-2 text-muted-foreground",
				children: /* @__PURE__ */ (0, q.jsx)(E, {
					icon: F,
					className: "text-xs"
				})
			})]
		}), j && P && (0, Ge.createPortal)(/* @__PURE__ */ (0, q.jsxs)("div", {
			ref: te,
			"data-slot": "tree-picker-popup",
			style: {
				position: "fixed",
				top: P.top,
				left: P.left,
				zIndex: N,
				minWidth: P.width,
				width: "max-content",
				maxWidth: P.maxWidth,
				maxHeight: P.maxHeight
			},
			className: i("flex flex-col overflow-hidden rounded-md border border-border bg-popover shadow-md", D),
			children: [/* @__PURE__ */ (0, q.jsx)("div", {
				className: "min-h-0 flex-1 overflow-auto",
				children: /* @__PURE__ */ (0, q.jsx)(ve, {
					roots: e,
					getKey: t,
					getChildren: n,
					renderRow: a,
					...r ? { getAriaLabel: r } : {},
					...s ? { getSearchText: s } : {},
					...c ? { defaultOpen: c } : {},
					...l ? { hasMoreChildren: l } : {},
					...u ? { loadChildren: u } : {},
					...m === void 0 ? {} : { selected: m },
					...h === void 0 ? {} : { revealSelected: h },
					...g === void 0 ? {} : { showControls: g },
					..._ === void 0 ? {} : { ariaLabel: _ },
					...v === void 0 ? {} : { empty: v },
					onSelect: (e) => {
						(!d || d(e)) && (f(e), p && M(!1));
					}
				})
			}), k && /* @__PURE__ */ (0, q.jsx)("div", {
				className: "shrink-0 border-t border-border bg-popover",
				children: k({ close: () => M(!1) })
			})]
		}), document.body)]
	});
}
// ../../../clicky-ui/packages/ui/src/components/filter-bar-tree.tsx
var Vn = (e) => (e.children?.length ?? 0) > 0;
function Hn({ filter: e, grow: t }) {
	let [n, r] = $t(e.value, e.onChange), a = (0, K.useMemo)(() => Pn(e.nodes, n), [e.nodes, n]), { label: s, ariaLabel: c } = on(e), l = (e, t) => r(nn(n, e, t));
	return /* @__PURE__ */ (0, q.jsx)(Bn, {
		roots: e.nodes,
		getKey: (e) => e.key,
		getChildren: (e) => e.children,
		getSearchText: (e) => [
			e.label,
			e.key,
			e.description
		].filter(Boolean).join(" "),
		defaultOpen: (e) => Fn(e, n),
		isSelectable: (e) => !Vn(e),
		onSelect: (e) => l(e.key, rn(n[e.key] ?? "neutral")),
		closeOnSelect: !1,
		ariaLabel: e.label,
		size: "sm",
		...e.disabled === void 0 ? {} : { disabled: e.disabled },
		className: i(sn(t), e.className),
		renderTrigger: ({ open: t, disabled: n, triggerRef: r, toggle: l }) => /* @__PURE__ */ (0, q.jsxs)("button", {
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
				/* @__PURE__ */ (0, q.jsx)("span", {
					className: "shrink-0 whitespace-nowrap text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
					children: s
				}),
				/* @__PURE__ */ (0, q.jsx)("span", {
					className: i("min-w-0 flex-1 truncate", !a && "text-placeholder"),
					children: a || e.placeholder
				}),
				/* @__PURE__ */ (0, q.jsx)("span", {
					className: "pointer-events-none absolute right-0 flex h-full items-center px-2 text-muted-foreground",
					children: /* @__PURE__ */ (0, q.jsx)(E, {
						icon: F,
						className: "text-xs"
					})
				})
			]
		}),
		renderRow: ({ node: e }) => /* @__PURE__ */ (0, q.jsx)(Kt, {
			className: "min-w-0 flex-1 justify-between",
			togglePosition: "right",
			mode: n[e.key] ?? "neutral",
			title: e.label,
			label: /* @__PURE__ */ (0, q.jsxs)("span", {
				className: "block min-w-0 text-left",
				children: [/* @__PURE__ */ (0, q.jsx)("span", {
					className: "block truncate",
					children: e.label
				}), e.description && /* @__PURE__ */ (0, q.jsx)("span", {
					className: "mt-0.5 block truncate text-xs font-normal text-muted-foreground",
					children: e.description
				})]
			}),
			onModeChange: (t) => l(e.key, t)
		})
	});
}
// ../../../clicky-ui/packages/ui/src/hooks/use-container-width.ts
function Un(e) {
	let [t, n] = (0, K.useState)(null), [r, i] = (0, K.useState)(!0);
	return (0, K.useLayoutEffect)(() => {
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
// ../../../clicky-ui/packages/ui/src/components/DateTimePicker.tsx
var Wn = (0, K.forwardRef)(({ value: e = "", onChange: t, className: n, inputClassName: r, buttonClassName: a, openButtonLabel: o = "Open time picker", suffix: s, prefix: c, ...l }, u) => {
	let d = (0, K.useRef)(null), f = (0, K.useRef)(null);
	function p(e) {
		d.current = e, typeof u == "function" ? u(e) : u && (u.current = e);
	}
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		"data-jsf-control": !0,
		className: i("relative", n),
		children: [
			c && /* @__PURE__ */ (0, q.jsx)("div", {
				className: "absolute inset-y-0 left-1.5 flex items-center",
				children: c
			}),
			/* @__PURE__ */ (0, q.jsx)("input", {
				...l,
				ref: p,
				type: "text",
				value: e,
				className: i("h-control-h w-full rounded-md border border-input bg-background px-control-px pr-8 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50", c && "pl-8", r),
				onChange: (e) => t?.(e.target.value)
			}),
			/* @__PURE__ */ (0, q.jsx)("input", {
				ref: f,
				type: "datetime-local",
				tabIndex: -1,
				"aria-hidden": "true",
				className: "pointer-events-none absolute left-0 top-0 h-0 w-0 opacity-0",
				value: Gn(e),
				onChange: (e) => t?.(e.target.value)
			}),
			s && /* @__PURE__ */ (0, q.jsx)("div", {
				className: "absolute inset-y-0 right-7 flex items-center",
				children: s
			}),
			/* @__PURE__ */ (0, q.jsx)("button", {
				type: "button",
				"aria-label": o,
				disabled: l.disabled,
				className: i("absolute inset-y-0 right-1 inline-flex items-center text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50", a),
				onClick: () => {
					d.current?.focus(), f.current?.showPicker?.();
				},
				children: /* @__PURE__ */ (0, q.jsx)(E, {
					icon: pt,
					className: "text-sm"
				})
			})
		]
	});
});
Wn.displayName = "DateTimePicker";
function Gn(e) {
	let t = e.trim();
	if (!t || t === "now" || t.startsWith("now")) return "";
	let n = /^\d{4}-\d{2}-\d{2}$/.test(t) ? `${t}T00:00` : t;
	return /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}/.test(n) ? n.slice(0, 16) : "";
}
// ../../../clicky-ui/packages/ui/src/components/select.tsx
var Kn = (0, K.forwardRef)(({ className: e, options: t, placeholder: n, children: r, ...a }, o) => /* @__PURE__ */ (0, q.jsxs)("div", {
	className: "relative inline-flex w-full items-center",
	children: [/* @__PURE__ */ (0, q.jsxs)("select", {
		ref: o,
		className: i("h-control-h w-full appearance-none rounded-md border border-input bg-background px-control-px pr-8 text-sm text-foreground", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1", "disabled:cursor-not-allowed disabled:opacity-50", e),
		...a,
		children: [n !== void 0 && /* @__PURE__ */ (0, q.jsx)("option", {
			value: "",
			disabled: !0,
			children: n
		}), t ? t.map((e) => /* @__PURE__ */ (0, q.jsx)("option", {
			value: e.value,
			disabled: e.disabled,
			children: e.label
		}, e.value)) : r]
	}), /* @__PURE__ */ (0, q.jsx)(E, {
		icon: F,
		className: "pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground"
	})]
}));
Kn.displayName = "Select";
// ../../../clicky-ui/packages/ui/src/components/MultiSelect.tsx
function qn({ options: e, value: t, onChange: n, placeholder: r = "Select options", ariaLabel: o, disabled: s, className: c, triggerClassName: l, menuClassName: u, facets: d = [] }) {
	let [f, p] = (0, K.useState)(!1), [h, _] = (0, K.useState)({}), v = e.filter((e) => d.every((t) => !h[t.key] || e.facets?.[t.key] === h[t.key])), y = re(), { refs: b, floatingStyles: C, context: w } = m({
		open: f,
		onOpenChange: p,
		placement: "bottom-start",
		whileElementsMounted: Ve,
		middleware: [
			Be(6),
			He({ padding: 8 }),
			Ue({ padding: 8 })
		]
	}), { getReferenceProps: T, getFloatingProps: D } = me([
		W(w),
		S(w, { escapeKey: !1 }),
		g(w, { role: "menu" })
	]);
	x(f, () => {
		p(!1), b.domReference.current instanceof HTMLElement && b.domReference.current.focus();
	});
	let k = (0, K.useMemo)(() => {
		let n = e.filter((e) => t.includes(e.value)), i = n.map((e) => e.label).filter((e) => typeof e == "string");
		return n.length === 0 ? r : i.length === 0 ? `${n.length} selected` : i.length <= 2 ? i.join(", ") : `${i.length} selected`;
	}, [
		e,
		r,
		t
	]);
	function A(e) {
		t.includes(e) ? n(t.filter((t) => t !== e)) : n([...t, e]);
	}
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i("inline-flex min-w-0", c),
		children: [/* @__PURE__ */ (0, q.jsxs)(a, {
			ref: b.setReference,
			type: "button",
			variant: "outline",
			disabled: s,
			"aria-label": o ?? `${r} filter`,
			className: i("w-fit max-w-[15rem] min-w-0 justify-between gap-3 text-left font-normal", l, t.length === 0 && "text-placeholder"),
			...T(),
			children: [/* @__PURE__ */ (0, q.jsx)("span", {
				className: "truncate",
				children: k
			}), /* @__PURE__ */ (0, q.jsx)(E, {
				icon: f ? H : F,
				className: "shrink-0 text-muted-foreground"
			})]
		}), f && /* @__PURE__ */ (0, q.jsx)(O, { children: /* @__PURE__ */ (0, q.jsx)(he, {
			context: w,
			modal: !1,
			children: /* @__PURE__ */ (0, q.jsxs)("div", {
				ref: b.setFloating,
				role: "menu",
				"aria-label": r,
				style: {
					...C,
					zIndex: y
				},
				className: i("min-w-[14rem] max-w-[20rem] rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-lg shadow-black/5", u),
				...D(),
				children: [
					/* @__PURE__ */ (0, q.jsxs)("div", {
						className: "mb-1 flex items-center justify-between gap-2 px-2 py-1 text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
						children: [/* @__PURE__ */ (0, q.jsx)("span", { children: r }), /* @__PURE__ */ (0, q.jsx)("button", {
							type: "button",
							className: "text-[10px] text-primary disabled:text-muted-foreground",
							onClick: () => n([]),
							disabled: t.length === 0,
							children: "Clear all"
						})]
					}),
					d.length > 0 && /* @__PURE__ */ (0, q.jsxs)("div", {
						className: "space-y-2 border-b border-border px-2 pb-2",
						children: [d.map((t) => {
							let n = /* @__PURE__ */ new Map();
							for (let r of e) {
								let e = r.facets?.[t.key];
								e !== void 0 && n.set(e, (n.get(e) ?? 0) + 1);
							}
							return /* @__PURE__ */ (0, q.jsxs)("div", {
								className: "space-y-1",
								children: [/* @__PURE__ */ (0, q.jsx)("span", {
									className: "text-xs font-medium",
									children: t.label
								}), /* @__PURE__ */ (0, q.jsx)(Kn, {
									"aria-label": t.label,
									value: h[t.key] ?? "",
									onChange: (e) => _((n) => ({
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
						}), /* @__PURE__ */ (0, q.jsx)(a, {
							type: "button",
							variant: "ghost",
							size: "sm",
							onClick: () => n([.../* @__PURE__ */ new Set([...t, ...v.filter((e) => !e.disabled).map((e) => e.value)])]),
							children: "Select all shown"
						})]
					}),
					/* @__PURE__ */ (0, q.jsxs)("div", {
						className: "max-h-64 overflow-auto",
						children: [v.length === 0 && /* @__PURE__ */ (0, q.jsx)("p", {
							role: "status",
							className: "px-2 py-3 text-sm text-muted-foreground",
							children: "No matching options"
						}), v.map((e) => {
							let n = t.includes(e.value);
							return /* @__PURE__ */ (0, q.jsxs)("label", {
								className: i("flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm hover:bg-accent", e.disabled && "cursor-not-allowed opacity-50"),
								children: [/* @__PURE__ */ (0, q.jsx)("input", {
									type: "checkbox",
									role: "menuitemcheckbox",
									className: "size-4 rounded border border-input",
									checked: n,
									disabled: e.disabled,
									onChange: () => A(e.value)
								}), /* @__PURE__ */ (0, q.jsx)("span", {
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
// ../../../clicky-ui/packages/ui/src/components/slider-shared.ts
var Jn = "pointer-events-none absolute inset-0 h-6 w-full appearance-none bg-transparent [&::-webkit-slider-runnable-track]:h-1 [&::-webkit-slider-runnable-track]:rounded-full [&::-webkit-slider-runnable-track]:bg-transparent [&::-moz-range-track]:h-1 [&::-moz-range-track]:rounded-full [&::-moz-range-track]:bg-transparent", Yn = "[&::-webkit-slider-thumb]:pointer-events-auto [&::-webkit-slider-thumb]:mt-[-5px] [&::-webkit-slider-thumb]:size-3.5 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:border [&::-webkit-slider-thumb]:border-background/80 [&::-webkit-slider-thumb]:bg-muted-foreground [&::-webkit-slider-thumb]:shadow-sm [&::-moz-range-thumb]:pointer-events-auto [&::-moz-range-thumb]:size-3.5 [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:border [&::-moz-range-thumb]:border-background/80 [&::-moz-range-thumb]:bg-muted-foreground [&::-moz-range-thumb]:shadow-sm", Xn = "h-7 min-w-0 rounded-md border border-input bg-background px-2 text-right font-mono text-xs tabular-nums text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring";
function Zn(e, t, n) {
	return Math.min(n, Math.max(t, e));
}
function Qn(e, { base: t, step: n, min: r, max: i }) {
	if (!(n > 0)) return Zn(e, r, i);
	let a = t + Math.round((e - t) / n) * n;
	return a < r ? Zn(t + Math.ceil((r - t) / n) * n, r, i) : a > i ? Zn(t + Math.floor((i - t) / n) * n, r, i) : a;
}
function $n(e, t, n) {
	return n <= t ? 0 : (e - t) / (n - t) * 100;
}
// ../../../clicky-ui/packages/ui/src/components/RangeSlider.tsx
function er({ min: e, max: t, value: n, onChange: r, step: a = 1, ariaLabelMin: o = "Minimum value", ariaLabelMax: s = "Maximum value", editable: c = !1, className: l, trackClassName: u, rangeClassName: d, thumbClassName: f }) {
	let [p, m] = nr(n, e, t);
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i(c && "space-y-2", l),
		children: [/* @__PURE__ */ (0, q.jsxs)("div", {
			className: "relative h-6",
			children: [
				/* @__PURE__ */ (0, q.jsx)("div", { className: i("absolute left-0 right-0 top-1/2 h-1 -translate-y-1/2 rounded-full bg-border", u) }),
				/* @__PURE__ */ (0, q.jsx)("div", {
					className: i("absolute top-1/2 h-1 -translate-y-1/2 rounded-full bg-primary/70", d),
					style: {
						left: `${$n(p, e, t)}%`,
						right: `${100 - $n(m, e, t)}%`
					}
				}),
				/* @__PURE__ */ (0, q.jsx)("input", {
					type: "range",
					"aria-label": o,
					min: e,
					max: t,
					step: a,
					value: p,
					className: i(Jn, Yn, f),
					style: { zIndex: p >= t ? 30 : 20 },
					onChange: (n) => {
						let i = Zn(Number(n.target.value), e, t);
						r([Math.min(i, m), m]);
					}
				}),
				/* @__PURE__ */ (0, q.jsx)("input", {
					type: "range",
					"aria-label": s,
					min: e,
					max: t,
					step: a,
					value: m,
					className: i(Jn, Yn, f),
					onChange: (n) => {
						let i = Zn(Number(n.target.value), e, t);
						r([p, Math.max(i, p)]);
					}
				})
			]
		}), c ? /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "grid grid-cols-2 gap-2",
			children: [/* @__PURE__ */ (0, q.jsx)(tr, {
				label: "Min",
				ariaLabel: `Edit ${rr(o)}`,
				min: e,
				max: m,
				step: a,
				stepBase: e,
				value: p,
				onChange: (e) => r([e, m])
			}), /* @__PURE__ */ (0, q.jsx)(tr, {
				label: "Max",
				ariaLabel: `Edit ${rr(s)}`,
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
function tr({ label: e, ariaLabel: t, min: n, max: r, step: i, stepBase: a, value: o, onChange: s }) {
	return /* @__PURE__ */ (0, q.jsxs)("label", {
		className: "grid grid-cols-[auto_minmax(0,1fr)] items-center gap-1.5 text-[11px] text-muted-foreground",
		children: [/* @__PURE__ */ (0, q.jsx)("span", { children: e }), /* @__PURE__ */ (0, q.jsx)("input", {
			"aria-label": t,
			className: Xn,
			max: r,
			min: n,
			step: i,
			type: "number",
			value: o,
			onChange: (e) => {
				if (e.currentTarget.value === "") return;
				let o = Number(e.currentTarget.value);
				if (!Number.isFinite(o)) throw Error(`${t} must be a finite number`);
				s(Qn(o, {
					base: a,
					step: i,
					min: n,
					max: r
				}));
			}
		})]
	});
}
function nr(e, t, n) {
	return [Zn(Math.min(e[0], e[1]), t, n), Zn(Math.max(e[0], e[1]), t, n)];
}
function rr(e) {
	return e.charAt(0).toLowerCase() + e.slice(1);
}
// ../../../clicky-ui/packages/ui/src/components/TimeRange.tsx
var Q = [
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
], ir = {
	min: Q[0],
	hr: Q[1],
	day: Q[2],
	"wk+": Q[3]
}, ar = [
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
], or = {
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
function sr({ kind: e = "time", label: t = "Time range", from: n = "", to: r = "", onApply: o, presets: s, chipRows: c, timeEnabled: l = !1, timeZone: u, timeZones: d, fromPlaceholder: f, toPlaceholder: p, emptyLabel: h, align: _ = "right", disabled: v = !1, className: y, triggerClassName: b, panelClassName: C, compact: w = !1 }) {
	let [T, D] = (0, K.useState)(!1), [k, A] = (0, K.useState)(n), [j, M] = (0, K.useState)(r), N = u ?? yr(), [P, ee] = (0, K.useState)(N), F = br(d, P), { rows: I, presets: L } = lr(e, s, c), te = re(), { refs: ne, floatingStyles: R, context: z } = m({
		open: T,
		onOpenChange: D,
		placement: _ === "left" ? "bottom-start" : "bottom-end",
		whileElementsMounted: Ve,
		middleware: [
			Be(6),
			He({ padding: 8 }),
			Ue({ padding: 8 })
		]
	}), { getReferenceProps: B, getFloatingProps: V } = me([
		W(z),
		S(z, { escapeKey: !1 }),
		g(z, { role: "dialog" })
	]);
	x(T, () => {
		D(!1), ne.domReference.current instanceof HTMLElement && ne.domReference.current.focus();
	}), (0, K.useEffect)(() => {
		T && (A(n), M(r), ee(N));
	}, [
		n,
		T,
		N,
		r
	]);
	let ie = hr(n, r, I, L) ?? vr(e, n, r, h), ae = _r(e, n, {
		timeEnabled: l,
		timeZone: P
	});
	function oe(t, n) {
		o(_r(e, t, {
			timeEnabled: l,
			timeZone: P
		}), _r(e, n, {
			timeEnabled: l,
			timeZone: P
		})), D(!1);
	}
	function se() {
		A(""), M(""), o("", ""), D(!1);
	}
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i("inline-flex", y),
		children: [/* @__PURE__ */ (0, q.jsxs)(a, {
			ref: ne.setReference,
			type: "button",
			variant: "outline",
			size: "sm",
			"aria-label": w ? `${t} filter: ${ie}` : `${t} filter`,
			title: w ? ie : void 0,
			disabled: v,
			className: i("h-7 min-w-0 text-xs font-normal", w ? "w-8 justify-center px-0 max-md:h-8" : "w-fit max-w-[12rem] gap-2 px-2", b),
			...B(),
			children: [/* @__PURE__ */ (0, q.jsx)(E, {
				icon: e === "time" ? Et : pt,
				className: "text-muted-foreground text-[14px]"
			}), !w && /* @__PURE__ */ (0, q.jsx)("span", {
				className: "truncate font-normal tabular-nums",
				children: ie
			})]
		}), T && /* @__PURE__ */ (0, q.jsx)(O, { children: /* @__PURE__ */ (0, q.jsx)(he, {
			context: z,
			modal: !1,
			children: /* @__PURE__ */ (0, q.jsxs)("div", {
				ref: ne.setFloating,
				role: "dialog",
				"aria-label": t,
				style: {
					...R,
					zIndex: te
				},
				className: i("flex max-h-[min(32rem,calc(100vh-2rem))] w-[22.5rem] max-w-[calc(100vw-1rem)] flex-col overflow-hidden rounded-lg border border-border bg-popover text-popover-foreground shadow-lg shadow-black/10 outline-none", C),
				...V(),
				children: [
					/* @__PURE__ */ (0, q.jsxs)("div", {
						className: "flex items-center justify-between border-b border-border/60 px-3.5 py-2.5",
						children: [/* @__PURE__ */ (0, q.jsx)("span", {
							className: "text-[10px] font-semibold uppercase tracking-wider text-muted-foreground",
							children: t
						}), /* @__PURE__ */ (0, q.jsx)("span", {
							className: "font-mono text-[10px] text-muted-foreground",
							children: l ? P : "Relative"
						})]
					}),
					/* @__PURE__ */ (0, q.jsxs)("div", {
						className: "min-h-0 flex-1 overflow-y-auto",
						children: [
							I.length > 0 && /* @__PURE__ */ (0, q.jsx)(pr, {
								rows: I,
								activeToken: ae,
								onApply: oe
							}),
							L.length > 0 && /* @__PURE__ */ (0, q.jsx)("div", {
								className: I.length > 0 ? "px-3.5 pb-3" : "px-3.5 pt-3",
								children: /* @__PURE__ */ (0, q.jsx)(mr, {
									presets: L,
									onApply: oe
								})
							}),
							/* @__PURE__ */ (0, q.jsxs)("div", {
								className: "border-t border-border/60 bg-muted/30 px-3.5 py-2.5",
								children: [/* @__PURE__ */ (0, q.jsxs)("div", {
									className: "mb-1.5 flex flex-wrap items-center justify-between gap-2",
									children: [/* @__PURE__ */ (0, q.jsx)("div", {
										className: "text-[10px] font-semibold uppercase tracking-wider text-muted-foreground",
										children: "Custom"
									}), l && /* @__PURE__ */ (0, q.jsxs)("div", {
										className: "flex min-w-0 items-center gap-1.5",
										children: [/* @__PURE__ */ (0, q.jsx)("span", {
											className: "text-[10px] font-semibold uppercase tracking-wider text-muted-foreground",
											children: "TZ"
										}), /* @__PURE__ */ (0, q.jsx)(Kn, {
											"aria-label": `${t} timezone`,
											value: P,
											options: F.map((e) => ({
												value: e,
												label: e
											})),
											onChange: (e) => ee(e.target.value),
											className: "h-7 w-40 truncate px-2 pr-7 font-mono text-xs"
										})]
									})]
								}), /* @__PURE__ */ (0, q.jsxs)("div", {
									className: "grid gap-1.5 md:grid-cols-[1fr_auto_1fr] md:items-center",
									children: [
										/* @__PURE__ */ (0, q.jsx)(cr, {
											"aria-label": `${t} from`,
											placeholder: f ?? (e === "date" ? "YYYY-MM-DD or now-1d" : "now-24h"),
											value: k,
											onChange: A,
											onClear: () => A(""),
											pickerKind: e,
											timeEnabled: l
										}),
										/* @__PURE__ */ (0, q.jsx)(E, {
											icon: dt,
											className: "hidden text-muted-foreground text-[12px] md:block"
										}),
										/* @__PURE__ */ (0, q.jsx)(cr, {
											"aria-label": `${t} to`,
											placeholder: p ?? (e === "date" ? "YYYY-MM-DD or now" : "now"),
											value: j,
											onChange: M,
											onClear: () => M(""),
											pickerKind: e,
											timeEnabled: l
										})
									]
								})]
							})
						]
					}),
					/* @__PURE__ */ (0, q.jsxs)("div", {
						className: "flex items-center justify-between border-t border-border/60 px-3 py-2",
						children: [/* @__PURE__ */ (0, q.jsx)(a, {
							type: "button",
							variant: "ghost",
							size: "sm",
							className: "h-7 px-2.5 text-xs",
							disabled: !n && !r,
							onClick: se,
							children: "Clear"
						}), /* @__PURE__ */ (0, q.jsxs)("div", {
							className: "flex gap-1.5",
							children: [/* @__PURE__ */ (0, q.jsx)(a, {
								type: "button",
								variant: "ghost",
								size: "sm",
								className: "h-7 px-2.5 text-xs",
								onClick: () => D(!1),
								children: "Cancel"
							}), /* @__PURE__ */ (0, q.jsx)(a, {
								type: "button",
								variant: "default",
								size: "sm",
								className: "h-7 px-3.5 text-xs",
								onClick: () => oe(k, j),
								children: "Apply"
							})]
						})]
					})
				]
			})
		}) })]
	});
}
function cr({ value: e, onChange: t, onClear: n, placeholder: r, pickerKind: i, timeEnabled: a, ...o }) {
	let s = o["aria-label"], c = i !== void 0, l = i === "time" && a, { date: u, time: d } = fr(e);
	function f(e) {
		t(l ? e ? d ? `${e}T${d}` : e : d ? `T${d}` : "" : e);
	}
	function p(e) {
		t(e ? u ? `${u}T${e}` : `T${e}` : u);
	}
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "space-y-1",
		children: [/* @__PURE__ */ (0, q.jsx)(ur, {
			ariaLabel: s,
			placeholder: r,
			value: l ? u : e,
			onChange: l ? f : t,
			onClear: n,
			showCalendar: c
		}), l && /* @__PURE__ */ (0, q.jsx)(dr, {
			ariaLabel: `${s} time`,
			value: d,
			onChange: p
		})]
	});
}
function lr(e, t, n) {
	if (t === void 0) return {
		rows: n ?? (e === "date" ? ar : Q),
		presets: []
	};
	let r = n ? [...n] : [], i = [];
	for (let e of t) typeof e == "string" ? e === "this" || e === "last" ? r.push(or[e]) : r.push(ir[e]) : i.push(e);
	return {
		rows: r,
		presets: i
	};
}
function ur({ ariaLabel: e, value: t, onChange: n, onClear: r, placeholder: a, showCalendar: o }) {
	let s = (0, K.useRef)(null), c = t.length > 0;
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "relative",
		children: [
			/* @__PURE__ */ (0, q.jsx)("input", {
				type: "text",
				"aria-label": e,
				placeholder: a,
				value: t,
				onChange: (e) => n(e.target.value),
				className: i("h-8 w-full rounded-md border border-input bg-background px-2 font-mono text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring", o && c ? "pr-12" : "pr-7")
			}),
			c && /* @__PURE__ */ (0, q.jsx)("button", {
				type: "button",
				"aria-label": `Clear ${e}`,
				onClick: r,
				className: i("absolute inset-y-0 inline-flex w-5 items-center justify-center text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", o ? "right-6" : "right-1"),
				children: /* @__PURE__ */ (0, q.jsx)(E, {
					icon: B,
					className: "text-[12px]"
				})
			}),
			o && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [/* @__PURE__ */ (0, q.jsx)("input", {
				ref: s,
				type: "date",
				tabIndex: -1,
				"aria-hidden": "true",
				value: /^\d{4}-\d{2}-\d{2}$/.test(t.trim()) ? t.trim() : "",
				onChange: (e) => n(e.target.value),
				className: "pointer-events-none absolute left-0 top-0 h-0 w-0 opacity-0"
			}), /* @__PURE__ */ (0, q.jsx)("button", {
				type: "button",
				"aria-label": `Pick ${e}`,
				onClick: () => s.current?.showPicker?.(),
				className: "absolute inset-y-0 right-1 inline-flex w-5 items-center justify-center text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
				children: /* @__PURE__ */ (0, q.jsx)(E, {
					icon: pt,
					className: "text-[12px]"
				})
			})] })
		]
	});
}
function dr({ ariaLabel: e, value: t, onChange: n }) {
	return /* @__PURE__ */ (0, q.jsx)("input", {
		type: "time",
		"aria-label": e,
		value: t,
		onChange: (e) => n(e.target.value),
		className: "h-8 w-full rounded-md border border-input bg-background px-2 font-mono text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring"
	});
}
function fr(e) {
	let t = e.trim(), n = /^(\d{4}-\d{2}-\d{2})?T?(\d{2}:\d{2})?/.exec(t);
	return !n || t.startsWith("now") || /^[+-]/.test(t) ? {
		date: t,
		time: ""
	} : {
		date: n[1] ?? "",
		time: n[2] ?? ""
	};
}
function pr({ rows: e, activeToken: t, onApply: n }) {
	return e.length === 0 ? null : /* @__PURE__ */ (0, q.jsx)("div", {
		className: "px-3.5 py-3",
		role: "list",
		"aria-label": "Time filters",
		children: e.map((r, a) => /* @__PURE__ */ (0, q.jsxs)("div", {
			className: i("grid grid-cols-[2rem_1fr] items-center gap-2", a === e.length - 1 ? "" : "mb-1.5"),
			role: "listitem",
			children: [/* @__PURE__ */ (0, q.jsx)("span", {
				className: "text-[9.5px] font-bold uppercase tracking-wider text-muted-foreground",
				children: r.label
			}), /* @__PURE__ */ (0, q.jsx)("div", {
				className: "flex divide-x divide-border overflow-hidden rounded-[5px] border border-border",
				children: r.chips.map((e) => {
					let r = e.token === t;
					return /* @__PURE__ */ (0, q.jsx)("button", {
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
function mr({ presets: e, onApply: t }) {
	return e.length === 0 ? null : /* @__PURE__ */ (0, q.jsx)("div", {
		className: "mb-3 space-y-1",
		role: "list",
		"aria-label": "Time filters",
		children: e.map((e) => /* @__PURE__ */ (0, q.jsxs)("button", {
			type: "button",
			className: "flex h-8 w-full items-center justify-between gap-3 rounded-md border border-border bg-background px-2 text-left text-xs hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
			onClick: () => t(e.from, e.to),
			children: [/* @__PURE__ */ (0, q.jsx)("span", {
				className: "min-w-0 truncate font-medium",
				children: e.label
			}), /* @__PURE__ */ (0, q.jsx)("span", {
				className: "shrink-0 font-mono text-[11px] text-muted-foreground",
				children: gr(e)
			})]
		}, `${e.label}-${e.from}-${e.to}`))
	});
}
function hr(e, t, n, r) {
	let i = e.trim(), a = t.trim();
	if (i || a) {
		for (let e of n) for (let t of e.chips) if (t.token === i && (t.to ?? "now") === a) return t.ariaLabel ?? t.title ?? t.label;
		return r.find((e) => e.from === i && e.to === a)?.label;
	}
}
function gr({ from: e, to: t }) {
	return e ? !t || t === "now" ? e : `${e} -> ${t}` : t;
}
function _r(e, t, n = {}) {
	let r = t.trim();
	return r ? r === "now" || r.startsWith("now") ? r : e === "time" && /^[+-]\d/.test(r) ? `now${r}` : e === "time" && n.timeEnabled ? xr(r, n.timeZone ?? yr()) : r : "";
}
function vr(e, t, n, r) {
	let i = t.trim(), a = n.trim();
	if (!i && !a) return r ?? (e === "date" ? "Any date" : "now-24h");
	if (e === "time") {
		let t = _r(e, i), n = _r(e, a || "now");
		return t ? n === "now" ? t : `${t} -> ${n}` : n;
	}
	return i && a && i === a ? i : i ? a ? `${i} -> ${a}` : i : a;
}
function yr() {
	try {
		return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
	} catch {
		return "UTC";
	}
}
function br(e, t) {
	let n = e && e.length > 0 ? e : [t, "UTC"];
	return Array.from(new Set([
		t,
		...n,
		"UTC"
	].filter(Boolean)));
}
function xr(e, t) {
	if (!Sr(e) || Cr(e)) return e;
	let n = e.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?(?:\.\d+)?$/);
	if (!n) return e;
	let [, r, i, a, o, s, c = "00"] = n;
	return `${r}-${i}-${a}T${o}:${s}:${c}${Tr(wr(t, Date.UTC(Number(r), Number(i) - 1, Number(a), Number(o), Number(s), Number(c))))}`;
}
function Sr(e) {
	return /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(?::\d{2})?(?:\.\d+)?$/.test(e.trim());
}
function Cr(e) {
	return /(?:Z|[+-]\d{2}:?\d{2})$/i.test(e.trim());
}
function wr(e, t) {
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
function Tr(e) {
	let t = e >= 0 ? "+" : "-", n = Math.abs(e);
	return `${t}${Math.floor(n / 60).toString().padStart(2, "0")}:${Math.floor(n % 60).toString().padStart(2, "0")}`;
}
// ../../../clicky-ui/packages/ui/src/components/tri-state.ts
function Er(e) {
	return e === void 0 || !e && void 0;
}
// ../../../clicky-ui/packages/ui/src/components/TriStateToggle.tsx
var Dr = {
	unset: "Any",
	on: "Yes",
	off: "No"
};
function Or({ value: e, onChange: t, label: n, labels: r, id: a, size: o = "sm", disabled: s, className: c }) {
	let l = {
		...Dr,
		...r
	}, u = e === void 0 ? "unset" : e ? "on" : "off", d = u === "on" ? de : u === "off" ? B : $e, f = u === "on" ? "border-green-500/50 bg-green-500/10 text-green-700 dark:text-green-400" : u === "off" ? "border-red-500/50 bg-red-500/10 text-red-700 dark:text-red-400" : "border-input bg-muted/30 text-muted-foreground";
	return /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		...a === void 0 ? {} : { id: a },
		role: "checkbox",
		"aria-checked": e === void 0 ? "mixed" : e,
		"aria-label": n,
		title: `${n}: ${l[u]}. Click to cycle.`,
		disabled: s,
		onClick: () => t(Er(e)),
		className: i("inline-flex shrink-0 items-center justify-center rounded-md border transition-colors", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", o === "sm" ? "h-6 w-6" : "h-8 w-8", f, s ? "opacity-60" : "hover:brightness-95", c),
		children: /* @__PURE__ */ (0, q.jsx)(E, {
			icon: d,
			className: o === "sm" ? "h-3.5 w-3.5" : "h-4 w-4"
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/NamespacePicker.tsx
function kr({ value: e, onChange: t, loadNamespaces: n, strict: r = !1, placeholder: i = "Select namespace…", className: a }) {
	let [o, s] = (0, K.useState)([]), [c, l] = (0, K.useState)(!1);
	(0, K.useEffect)(() => {
		let e = !1;
		return l(!0), n().then((t) => {
			e || s(t);
		}).finally(() => {
			e || l(!1);
		}), () => {
			e = !0;
		};
	}, [n]);
	let u = (0, K.useMemo)(() => {
		let t = o.map((e) => ({
			value: e,
			label: e
		}));
		return e && !o.includes(e) ? [{
			value: e,
			label: e
		}, ...t] : t;
	}, [o, e]), d = (0, K.useMemo)(() => r && !!e && !c && !o.includes(e), [
		r,
		e,
		c,
		o
	]);
	return /* @__PURE__ */ (0, q.jsx)(kn, {
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
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@flanksource/icons/1.0.63/b00fc4bb064c1d1ae674af83cb73c1f20c4dd4dfd0aa56010e6ee736bf3e3c7d/node_modules/@flanksource/icons/lib/iconContext.mjs
var Ar = {
	color: void 0,
	size: void 0,
	className: void 0,
	style: void 0,
	attr: void 0
}, jr = K.createContext && /*#__PURE__*/ K.createContext(Ar), Mr = [
	"attr",
	"size",
	"title",
	"square"
];
function Nr(e, t) {
	if (e == null) return {};
	var n = Pr(e, t), r, i;
	if (Object.getOwnPropertySymbols) {
		var a = Object.getOwnPropertySymbols(e);
		for (i = 0; i < a.length; i++) r = a[i], !(t.indexOf(r) >= 0) && Object.prototype.propertyIsEnumerable.call(e, r) && (n[r] = e[r]);
	}
	return n;
}
function Pr(e, t) {
	if (e == null) return {};
	for (var n = {}, r = Object.keys(e), i, a = 0; a < r.length; a++) i = r[a], !(t.indexOf(i) >= 0) && (n[i] = e[i]);
	return n;
}
function Fr() {
	return Fr = Object.assign ? Object.assign.bind() : function(e) {
		for (var t = 1; t < arguments.length; t++) {
			var n = arguments[t];
			for (var r in n) Object.prototype.hasOwnProperty.call(n, r) && (e[r] = n[r]);
		}
		return e;
	}, Fr.apply(this, arguments);
}
function Ir(e, t) {
	var n = Object.keys(e);
	if (Object.getOwnPropertySymbols) {
		var r = Object.getOwnPropertySymbols(e);
		t && (r = r.filter(function(t) {
			return Object.getOwnPropertyDescriptor(e, t).enumerable;
		})), n.push.apply(n, r);
	}
	return n;
}
function Lr(e) {
	for (var t = 1; t < arguments.length; t++) {
		var n = arguments[t] == null ? {} : arguments[t];
		t % 2 ? Ir(Object(n), !0).forEach(function(t) {
			Rr(e, t, n[t]);
		}) : Object.getOwnPropertyDescriptors ? Object.defineProperties(e, Object.getOwnPropertyDescriptors(n)) : Ir(Object(n)).forEach(function(t) {
			Object.defineProperty(e, t, Object.getOwnPropertyDescriptor(n, t));
		});
	}
	return e;
}
function Rr(e, t, n) {
	return t = zr(t), t in e ? Object.defineProperty(e, t, {
		value: n,
		enumerable: !0,
		configurable: !0,
		writable: !0
	}) : e[t] = n, e;
}
function zr(e) {
	var t = Br(e, "string");
	return typeof t == "symbol" ? t : String(t);
}
function Br(e, t) {
	if (typeof e != "object" || !e) return e;
	var n = e[Symbol.toPrimitive];
	if (n !== void 0) {
		var r = n.call(e, t || "default");
		if (typeof r != "object") return r;
		throw TypeError("@@toPrimitive must return a primitive value.");
	}
	return (t === "string" ? String : Number)(e);
}
function Vr(e) {
	return e && e.map((e, t) => /*#__PURE__*/ K.createElement(e.tag, Lr({ key: t }, e.attr), Vr(e.child)));
}
function Hr(e) {
	var t = (t) => /*#__PURE__*/ K.createElement(Wr, Fr({ attr: Lr({}, e.attr) }, t), Vr(e.child));
	return t.viewBox = e.attr?.viewBox, t;
}
function Ur(e) {
	if (!e) return !1;
	var t = e.trim().split(/[\s,]+/);
	if (t.length !== 4) return !1;
	var n = parseFloat(t[2]), r = parseFloat(t[3]);
	return !isFinite(n) || !isFinite(r) || r <= 0 ? !1 : n / r >= 2;
}
function Wr(e) {
	var t = (t) => {
		var { attr: n, size: r, title: i, square: a } = e, o = Nr(e, Mr), s = r || t.size || "1em", c = a ?? !Ur(n?.viewBox), l;
		return t.className && (l = t.className), e.className && (l = (l ? l + " " : "") + e.className), /*#__PURE__*/ K.createElement("svg", Fr({
			stroke: "currentColor",
			fill: "currentColor",
			strokeWidth: "0"
		}, t.attr, n, o, {
			className: l,
			style: Lr(Lr({ color: e.color || t.color }, t.style), e.style),
			height: s,
			width: c ? s : void 0,
			xmlns: "http://www.w3.org/2000/svg"
		}), i && /*#__PURE__*/ K.createElement("title", null, i), e.children);
	};
	return jr === void 0 ? t(Ar) : /*#__PURE__*/ K.createElement(jr.Consumer, null, (e) => t(e));
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@flanksource/icons/1.0.63/b00fc4bb064c1d1ae674af83cb73c1f20c4dd4dfd0aa56010e6ee736bf3e3c7d/node_modules/@flanksource/icons/mi/index.mjs
function Gr(e) {
	return Hr({
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
function Kr(e) {
	return Hr({
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
function qr(e) {
	return Hr({
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
function Jr(e) {
	return Hr({
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
function Yr(e) {
	return Hr({
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
// ../../../clicky-ui/packages/ui/src/components/DaemonSetIcon.tsx
function Xr({ className: e, title: t, "aria-label": n }) {
	return /* @__PURE__ */ (0, q.jsx)(ut, {
		...e === void 0 ? {} : { className: e },
		...t === void 0 ? {} : { title: t },
		...n === void 0 ? {} : { "aria-label": n }
	});
}
// ../../../clicky-ui/packages/ui/src/components/workload-picker-utils.tsx
var Zr = {
	service: {
		label: "Service",
		Icon: Jr
	},
	ingress: {
		label: "Ingress",
		Icon: Kr
	},
	pod: {
		label: "Pod",
		Icon: qr
	},
	deployment: {
		label: "Deployment",
		Icon: Gr
	},
	statefulset: {
		label: "StatefulSet",
		Icon: Yr
	},
	daemonset: {
		label: "DaemonSet",
		Icon: Xr
	}
}, Qr = [
	"service",
	"ingress",
	"pod",
	"deployment",
	"statefulset",
	"daemonset"
];
function $r(e, t) {
	return e === "ingress" ? t.hosts?.[0] : void 0;
}
function ei(e, t) {
	return $r(e, t) ?? t.name;
}
var ti = new Set(Qr);
function ni(e) {
	return e ? `${e}/` : "";
}
function ri(e, t, n) {
	let r = ei(t, n), i = n.namespace ?? e;
	return i ? `${i}/${t}/${r}` : `${t}/${r}`;
}
function ii(e) {
	let t = e.split("/"), n = t.at(-1), r = t.at(-2);
	if (n != null && r != null && ti.has(r)) {
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
function ai(e, t, n) {
	let r = /* @__PURE__ */ new Set(), i = /* @__PURE__ */ new Set();
	for (let a of n) for (let n of t[a] ?? []) r.add(ri(e, a, n)), i.add(ei(a, n));
	return {
		keys: r,
		names: i
	};
}
function oi(e, t) {
	return (t ? ii(t).kind : void 0) ?? e[0] ?? "service";
}
function si(e, t, n, r) {
	let i = [];
	for (let r of n) {
		let n = Zr[r], a = n.Icon;
		for (let o of t[r] ?? []) {
			let t = $r(r, o);
			i.push({
				value: ri(e, r, o),
				label: t ? `${t} (${o.name})` : o.name,
				group: n.label,
				icon: /* @__PURE__ */ (0, q.jsx)(a, { className: "h-4 w-4" })
			});
		}
	}
	if (r && !i.some((e) => e.value === r)) {
		let e = ii(r), t = Zr[e.kind ?? n[0] ?? "service"].Icon;
		return [{
			value: r,
			label: e.name,
			icon: /* @__PURE__ */ (0, q.jsx)(t, { className: "h-4 w-4" })
		}, ...i];
	}
	return i;
}
// ../../../clicky-ui/packages/ui/src/components/WorkloadPicker.tsx
function ci({ value: e, onChange: t, loadWorkloads: n, namespace: r, allowNamespaceSelection: a = !1, loadNamespaces: o, onNamespaceChange: s, kinds: c = Qr, strict: l = !1, allowCustomValue: u = !0, collapseSingleOption: d = !1, disabled: f = !1, placeholder: p = "Select workload / service…", className: m }) {
	let h = e ? ii(e).namespace : void 0, [g, _] = (0, K.useState)(h ?? r ?? ""), [v, y] = (0, K.useState)({}), [b, x] = (0, K.useState)(!1), S = (0, K.useRef)(r), C = a ? g || void 0 : r;
	(0, K.useEffect)(() => {
		if (!a || !e) return;
		let t = ii(e).namespace;
		t && _(t);
	}, [a, e]), (0, K.useEffect)(() => {
		a && S.current !== r && (S.current = r, e || _(r ?? ""));
	}, [
		a,
		r,
		e
	]);
	let w = c.join(",");
	(0, K.useEffect)(() => {
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
	let T = e && ii(e).name ? e : "", E = (0, K.useMemo)(() => si(C, v, c, T), [
		C,
		v,
		c,
		T
	]), D = (0, K.useMemo)(() => {
		if (!l || !T || b) return !1;
		let { keys: e, names: t } = ai(C, v, c);
		return !e.has(T) && !t.has(T);
	}, [
		l,
		T,
		b,
		C,
		v,
		c
	]), O = Zr[oi(c, T)], k = O.Icon, A = d && !b && E.length === 1 && (!T || T === E[0].value) ? E[0] : void 0, j = /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i("flex items-center gap-2", a ? "min-w-0 flex-1" : m),
		children: [/* @__PURE__ */ (0, q.jsx)(k, {
			className: "h-4 w-4 shrink-0 text-muted-foreground",
			title: O.label,
			"aria-label": O.label
		}), /* @__PURE__ */ (0, q.jsx)("div", {
			className: "min-w-0 flex-1",
			children: A ? /* @__PURE__ */ (0, q.jsx)("span", {
				className: "block truncate text-sm text-foreground",
				title: A.label,
				"aria-label": "Workload",
				children: A.label
			}) : /* @__PURE__ */ (0, q.jsx)(kn, {
				options: E,
				value: T,
				onChange: (e) => t(e || (a ? ni(C) : "")),
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
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i("flex w-full min-w-0 flex-nowrap items-center gap-density-2", m),
		children: [/* @__PURE__ */ (0, q.jsx)(kr, {
			className: "min-w-0 shrink grow-0 basis-28",
			value: g,
			onChange: (e) => {
				_(e), y({}), s?.(e || void 0), t(ni(e));
			},
			loadNamespaces: o,
			strict: !0
		}), j]
	});
}
// ../../../clicky-ui/packages/ui/src/components/FilterBar.tsx
var li = 8, ui = 44, di = "flex items-center gap-2 overflow-visible bg-background py-1.5";
function fi({ search: e, filters: t, timeRange: n, dateRange: r, children: o, leading: s, trailing: c, className: l, autoSubmit: u = !0, onApply: d, applyLabel: p = "Apply", isPending: m = !1, overflowMode: h = "responsive", overlay: g }) {
	let _ = !!(n || r), v = !u && !!d, y = (0, K.useMemo)(() => {
		let e = t ?? [];
		return [...e.filter((e) => e.placement !== "overflow"), ...e.filter((e) => e.placement === "overflow")];
	}, [t]), b = y.findIndex((e) => e.placement === "overflow"), x = b === -1 ? y.length : b, S = x < y.length, C = ne(z("md")), w = ne(z("sm")), { ref: T, wider: E } = Un(f.sm), D = C || !E, O = (0, K.useMemo)(() => ({
		autoSubmit: u,
		compact: D
	}), [u, D]), k = h === "responsive" ? C : w, A = y.length > 0 && (h === "responsive" || k || S), j = (0, K.useRef)(null), M = (0, K.useRef)(null), N = (0, K.useRef)(/* @__PURE__ */ new Map()), P = (0, K.useRef)(/* @__PURE__ */ new Map()), ee = (0, K.useMemo)(() => y.map((e) => `${e.key}:${e.placement ?? "auto"}`).join("\0"), [y]), [F, I] = (0, K.useState)(x), L = (0, K.useRef)(y), te = (0, K.useRef)(A), R = (0, K.useRef)(k), re = (0, K.useRef)(x);
	L.current = y, te.current = A, R.current = k, re.current = x, (0, K.useLayoutEffect)(() => {
		I(te.current && R.current ? 0 : re.current);
	}, [ee, k]);
	let B = (0, K.useCallback)(() => {
		let e = L.current, t = re.current;
		if (!te.current) {
			I(e.length);
			return;
		}
		if (R.current) {
			I(0);
			return;
		}
		let n = j.current;
		if (!n || (n.parentElement?.getBoundingClientRect().width ?? 0) <= 0) return;
		let r = Math.floor(n.getBoundingClientRect().width), i = e.slice(0, t).map((e) => {
			let t = N.current.get(e.key)?.getBoundingClientRect().width ?? 0;
			if (t > 0) {
				let n = Math.ceil(t);
				return P.current.set(e.key, n), n;
			}
			return P.current.get(e.key) ?? Ca(e);
		}), a = Math.ceil(M.current?.getBoundingClientRect().width ?? ui), o = a > 0 ? li : 0, s = Sa(i), c = !(t < e.length) && s <= r ? t : xa(i, Math.max(0, r - a - o));
		I((e) => e === c ? e : c);
	}, []);
	(0, K.useLayoutEffect)(() => {
		B();
	}, [B]), (0, K.useEffect)(() => {
		if (!A) return;
		let e = j.current, t = M.current, n = typeof ResizeObserver > "u" ? null : ResizeObserver;
		if (!n) return window.addEventListener("resize", B), B(), () => window.removeEventListener("resize", B);
		let r = new n(() => B());
		return e && r.observe(e), t && r.observe(t), window.addEventListener("resize", B), B(), () => {
			r.disconnect(), window.removeEventListener("resize", B);
		};
	}, [
		B,
		A,
		k
	]);
	let V = Math.min(F, x), ie = A ? y.slice(0, V) : y, ae = A ? y.slice(V) : [], oe = ae.filter(zt).length, se = (0, K.useCallback)((e, t) => {
		t ? N.current.set(e, t) : N.current.delete(e);
	}, []), ce = /* @__PURE__ */ (0, q.jsx)(Yt.Provider, {
		value: O,
		children: /* @__PURE__ */ (0, q.jsxs)("div", {
			ref: T,
			"data-slot": "filter-bar",
			className: i(di, h === "wrap" ? "flex-wrap max-sm:flex-nowrap" : "flex-wrap max-sm:flex-nowrap md:flex-nowrap", l, g && "invisible pointer-events-none"),
			...g ? { "aria-hidden": !0 } : {},
			children: [
				s && /* @__PURE__ */ (0, q.jsx)("div", {
					className: "flex shrink-0 items-center gap-2",
					children: s
				}),
				e && /* @__PURE__ */ (0, q.jsx)(Ii, { search: e }),
				o,
				y.length > 0 && /* @__PURE__ */ (0, q.jsx)("div", {
					ref: j,
					"data-filter-bar-list": !0,
					className: i("flex min-w-0 items-center gap-2", h === "wrap" ? "flex-wrap" : "flex-1"),
					children: ie.map((t, n) => /* @__PURE__ */ (0, q.jsx)("div", {
						ref: (e) => se(t.key, e),
						"data-filter-bar-item": t.key,
						className: "min-w-0 shrink-0",
						children: hi(t, !e && n === 0)
					}, t.key))
				}),
				A && ae.length > 0 && /* @__PURE__ */ (0, q.jsx)(gi, {
					triggerRef: M,
					filters: ae,
					activeHidden: oe,
					allHidden: ie.length === 0,
					mobilePage: w,
					...d ? { onApply: d } : {}
				}),
				(_ || c || v) && /* @__PURE__ */ (0, q.jsxs)("div", {
					className: "ml-auto flex min-w-0 items-center justify-end gap-2 max-md:flex-nowrap md:shrink-0 md:flex-wrap",
					children: [
						r && /* @__PURE__ */ (0, q.jsx)(Zi, {
							kind: "date",
							label: "Date range",
							...r
						}),
						n && /* @__PURE__ */ (0, q.jsx)(Zi, {
							kind: n.timeEnabled === !1 ? "date" : "time",
							label: n.timeEnabled === !1 ? "Date range" : "Time range",
							...n
						}),
						c,
						v && /* @__PURE__ */ (0, q.jsx)(a, {
							type: "button",
							variant: "default",
							size: "sm",
							disabled: m,
							onClick: d,
							children: m ? "Loading…" : p
						})
					]
				})
			]
		})
	});
	return g ? /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "grid",
		children: [/* @__PURE__ */ (0, q.jsx)("div", {
			className: "col-start-1 row-start-1 min-w-0",
			children: ce
		}), /* @__PURE__ */ (0, q.jsx)("div", {
			"data-slot": "filter-bar-overlay",
			className: i(di, "col-start-1 row-start-1 min-w-0 flex-wrap"),
			children: g
		})]
	}) : ce;
}
function pi({ filter: e, chrome: t = "full", autoSubmit: n = !0 }) {
	let r = (0, K.useMemo)(() => ({
		autoSubmit: n,
		compact: !1
	}), [n]);
	return /* @__PURE__ */ (0, q.jsx)(Yt.Provider, {
		value: r,
		children: /* @__PURE__ */ (0, q.jsx)(mi, {
			filter: e,
			chrome: t
		})
	});
}
function mi({ filter: e, chrome: t = "full" }) {
	return /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
		e.kind === "lookup" && /* @__PURE__ */ (0, q.jsx)(zi, {
			filter: e,
			grow: !0
		}),
		e.kind === "lookup-multi" && /* @__PURE__ */ (0, q.jsx)(Bi, {
			filter: e,
			grow: !0
		}),
		e.kind === "multi" && /* @__PURE__ */ (0, q.jsx)(Mn, {
			filter: e,
			chrome: t
		}),
		e.kind === "tree" && /* @__PURE__ */ (0, q.jsx)(Hn, {
			filter: e,
			grow: !0
		}),
		e.kind === "nested-multi" && /* @__PURE__ */ (0, q.jsx)(qi, {
			filter: e,
			chrome: t
		}),
		e.kind === "workload" && /* @__PURE__ */ (0, q.jsx)(Si, {
			filter: e,
			grow: !0
		}),
		e.kind === "select-multi" && /* @__PURE__ */ (0, q.jsx)(Xi, {
			filter: e,
			grow: !0
		}),
		(e.kind === "number" || e.kind === "duration") && /* @__PURE__ */ (0, q.jsx)(Ui, {
			filter: e,
			chrome: t
		}),
		e.kind === "enum" && /* @__PURE__ */ (0, q.jsx)(Ni, {
			filter: e,
			grow: !0
		}),
		e.kind === "boolean" && /* @__PURE__ */ (0, q.jsx)(Fi, { filter: e }),
		e.kind === "tristate" && /* @__PURE__ */ (0, q.jsx)(Pi, { filter: e }),
		e.kind === "date-range" && /* @__PURE__ */ (0, q.jsx)(Mi, {
			kind: e.timeEnabled === !1 ? "date" : "time",
			label: e.label,
			...Lt(e)
		}),
		e.kind === "text" && /* @__PURE__ */ (0, q.jsx)(Li, {
			filter: e,
			grow: !0
		})
	] });
}
function hi(e, t) {
	return e.kind === "lookup" ? /* @__PURE__ */ (0, q.jsx)(zi, {
		filter: e,
		grow: t
	}) : e.kind === "lookup-multi" ? /* @__PURE__ */ (0, q.jsx)(Bi, {
		filter: e,
		grow: t
	}) : e.kind === "multi" ? /* @__PURE__ */ (0, q.jsx)(An, {
		filter: e,
		grow: t
	}) : e.kind === "tree" ? /* @__PURE__ */ (0, q.jsx)(Hn, {
		filter: e,
		grow: t
	}) : e.kind === "nested-multi" ? /* @__PURE__ */ (0, q.jsx)(Ji, {
		filter: e,
		grow: t
	}) : e.kind === "workload" ? /* @__PURE__ */ (0, q.jsx)(Si, {
		filter: e,
		grow: t
	}) : e.kind === "select-multi" ? /* @__PURE__ */ (0, q.jsx)(Xi, {
		filter: e,
		grow: t
	}) : e.kind === "number" || e.kind === "duration" ? /* @__PURE__ */ (0, q.jsx)(Hi, {
		filter: e,
		grow: t
	}) : e.kind === "enum" ? /* @__PURE__ */ (0, q.jsx)(Ni, {
		filter: e,
		grow: t
	}) : e.kind === "boolean" ? /* @__PURE__ */ (0, q.jsx)(Fi, { filter: e }) : e.kind === "tristate" ? /* @__PURE__ */ (0, q.jsx)(Pi, { filter: e }) : e.kind === "date-range" ? /* @__PURE__ */ (0, q.jsx)(Zi, {
		kind: e.timeEnabled === !1 ? "date" : "time",
		label: e.label,
		...Lt(e)
	}) : /* @__PURE__ */ (0, q.jsx)(Li, {
		filter: e,
		grow: t
	});
}
function gi({ triggerRef: e, filters: t, activeHidden: n, allHidden: r, mobilePage: o, onApply: s }) {
	let [c, l] = (0, K.useState)(!1), [u, d] = (0, K.useState)({}), f = t.length > 0, p = (0, K.useMemo)(() => t.map((e) => e.key).join("\0"), [t]), { refs: m, floatingStyles: g, context: _, floatingZ: v, getReferenceProps: y, getFloatingProps: b } = Qi(c && !o, (e) => {
		d(fa(t)), l(e);
	}, "bottom-end"), x = (0, K.useCallback)(() => {
		d(fa(t)), l(!1);
	}, [t]), S = C([m.setReference, e]);
	(0, K.useEffect)(() => {
		f || x();
	}, [x, f]), (0, K.useEffect)(() => {
		c && d(fa(t));
	}, [p, c]);
	let w = t.map((e) => ha(e, Object.hasOwn(u, e.key) ? u[e.key] : ma(e), (t) => d((n) => ({
		...n,
		[e.key]: t
	})))), T = () => {
		ga(t, u), s?.(), l(!1);
	}, D = /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		className: i("rounded text-primary transition-colors hover:bg-accent focus:bg-accent focus:outline-none disabled:text-muted-foreground", o ? "h-10 px-3 text-sm" : "px-1.5 py-0.5 text-xs"),
		onClick: () => w.forEach(Rt),
		disabled: w.every((e) => !zt(e)),
		children: "Clear all"
	}), k = /* @__PURE__ */ (0, q.jsx)(vi, {
		onClose: x,
		onApply: T,
		mobilePage: o
	}), A = r ? "Filters" : "More filters";
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i("inline-flex shrink-0", !f && "invisible pointer-events-none"),
		"aria-hidden": !f || void 0,
		children: [
			/* @__PURE__ */ (0, q.jsxs)(a, {
				ref: S,
				type: "button",
				variant: "outline",
				size: "sm",
				"aria-label": A,
				tabIndex: f ? 0 : -1,
				title: f ? A : void 0,
				className: i("h-8 min-w-0 gap-1.5 px-2 text-xs font-normal max-md:h-9", n > 0 && "border-primary/40 text-primary"),
				...y(),
				"aria-expanded": c,
				children: [
					/* @__PURE__ */ (0, q.jsx)(E, {
						icon: Ze,
						className: "text-[14px]"
					}),
					r && /* @__PURE__ */ (0, q.jsx)("span", { children: "Filters" }),
					n > 0 && /* @__PURE__ */ (0, q.jsx)("span", {
						className: "rounded-full bg-muted px-1.5 text-[10px] font-medium text-muted-foreground",
						children: n
					})
				]
			}),
			o && /* @__PURE__ */ (0, q.jsx)(h, {
				open: f && c,
				onClose: x,
				title: "Filters",
				headerSlot: D,
				footer: k,
				scrollBody: !1,
				children: /* @__PURE__ */ (0, q.jsx)(_i, {
					filters: w,
					mobilePage: !0
				})
			}),
			!o && f && c && /* @__PURE__ */ (0, q.jsx)(O, { children: /* @__PURE__ */ (0, q.jsx)(he, {
				context: _,
				modal: !1,
				children: /* @__PURE__ */ (0, q.jsxs)("div", {
					ref: m.setFloating,
					role: "dialog",
					"aria-label": "Overflow filters",
					style: {
						...g,
						zIndex: v
					},
					className: "flex max-h-[min(34rem,calc(100vh-2rem))] w-[min(34rem,calc(100vw-2rem))] flex-col overflow-hidden rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-lg shadow-black/10",
					...b(),
					children: [
						/* @__PURE__ */ (0, q.jsxs)("div", {
							className: "mb-2 flex items-center justify-between gap-2 px-1",
							children: [/* @__PURE__ */ (0, q.jsx)("div", {
								className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
								children: "Filters"
							}), /* @__PURE__ */ (0, q.jsxs)("div", {
								className: "flex items-center gap-1",
								children: [D, /* @__PURE__ */ (0, q.jsx)("button", {
									type: "button",
									"aria-label": "Close overflow filters",
									title: "Close",
									className: "inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none",
									onClick: x,
									children: /* @__PURE__ */ (0, q.jsx)(E, {
										icon: B,
										className: "text-sm"
									})
								})]
							})]
						}),
						/* @__PURE__ */ (0, q.jsx)(_i, { filters: w }),
						/* @__PURE__ */ (0, q.jsx)("div", {
							className: "mt-3",
							children: k
						})
					]
				})
			}) })
		]
	});
}
function _i({ filters: e, mobilePage: t = !1 }) {
	return /* @__PURE__ */ (0, q.jsx)("div", {
		className: i("min-h-0 flex-1 divide-y divide-border/70 overflow-y-auto", t ? "bg-background" : "rounded-md border border-border/70"),
		children: e.map((e) => {
			let n = zt(e);
			return /* @__PURE__ */ (0, q.jsxs)("div", {
				"data-overflow-filter-row": e.key,
				className: "grid grid-cols-[minmax(0,1fr)_auto] items-start gap-2 overflow-visible p-4 md:h-12 md:grid-cols-[minmax(7rem,10rem)_auto_minmax(0,1fr)_auto] md:items-center md:p-2",
				children: [
					/* @__PURE__ */ (0, q.jsxs)("label", {
						htmlFor: bi(e),
						className: "col-span-2 flex min-w-0 items-center gap-1 truncate text-[10px] font-medium uppercase tracking-wide text-muted-foreground md:col-span-1",
						title: e.description ?? e.label,
						children: [/* @__PURE__ */ (0, q.jsx)(U, {
							icon: an(e.icon, 12),
							className: "text-[12px] normal-case"
						}), /* @__PURE__ */ (0, q.jsx)("span", {
							className: "truncate",
							children: e.label
						})]
					}),
					/* @__PURE__ */ (0, q.jsx)("span", {
						className: "hidden text-sm text-muted-foreground md:block",
						children: "="
					}),
					/* @__PURE__ */ (0, q.jsx)("div", {
						className: "min-w-0 overflow-visible",
						children: /* @__PURE__ */ (0, q.jsx)(Yt.Provider, {
							value: {
								autoSubmit: !1,
								compact: !1
							},
							children: /* @__PURE__ */ (0, q.jsx)(yi, { filter: e })
						})
					}),
					/* @__PURE__ */ (0, q.jsx)("button", {
						type: "button",
						"aria-label": `Clear ${e.label}`,
						title: `Clear ${e.label}`,
						className: i("inline-flex self-center items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none disabled:text-muted-foreground/40", t ? "h-10 w-10" : "h-6 w-6"),
						onClick: () => Rt(e),
						disabled: !n,
						children: /* @__PURE__ */ (0, q.jsx)(E, {
							icon: B,
							className: "text-sm"
						})
					})
				]
			}, e.key);
		})
	});
}
function vi({ onClose: e, onApply: t, mobilePage: n = !1 }) {
	let r = n ? "h-10 px-4 text-sm" : "h-8 px-3 text-xs";
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "flex justify-end gap-2",
		children: [/* @__PURE__ */ (0, q.jsx)(a, {
			type: "button",
			variant: "ghost",
			size: "sm",
			className: r,
			onClick: e,
			children: "Close"
		}), /* @__PURE__ */ (0, q.jsx)(a, {
			type: "button",
			variant: "default",
			size: "sm",
			className: r,
			onClick: t,
			children: "Apply"
		})]
	});
}
function yi({ filter: e }) {
	return e.kind === "date-range" ? /* @__PURE__ */ (0, q.jsx)(Zi, {
		kind: e.timeEnabled === !1 ? "date" : "time",
		label: e.label,
		...Lt(e)
	}) : e.kind === "text" ? /* @__PURE__ */ (0, q.jsx)(wi, { filter: e }) : e.kind === "lookup" ? /* @__PURE__ */ (0, q.jsx)(Ti, { filter: e }) : e.kind === "lookup-multi" ? /* @__PURE__ */ (0, q.jsx)(Ei, { filter: e }) : e.kind === "enum" ? /* @__PURE__ */ (0, q.jsx)(Oi, { filter: e }) : e.kind === "boolean" ? /* @__PURE__ */ (0, q.jsx)(ki, { filter: e }) : e.kind === "select-multi" ? /* @__PURE__ */ (0, q.jsx)(ji, { filter: e }) : e.kind === "multi" ? /* @__PURE__ */ (0, q.jsx)(An, {
		filter: e,
		grow: !0
	}) : e.kind === "tree" ? /* @__PURE__ */ (0, q.jsx)(Hn, {
		filter: e,
		grow: !0
	}) : e.kind === "nested-multi" ? /* @__PURE__ */ (0, q.jsx)(Ji, {
		filter: e,
		grow: !0
	}) : e.kind === "workload" ? /* @__PURE__ */ (0, q.jsx)(Si, {
		filter: e,
		grow: !0
	}) : e.kind === "tristate" ? /* @__PURE__ */ (0, q.jsx)(Ai, { filter: e }) : /* @__PURE__ */ (0, q.jsx)(Hi, {
		filter: e,
		grow: !0
	});
}
function bi(e) {
	return `filterbar-overflow-${e.key}`;
}
function xi(e) {
	return e.map((e) => ({
		value: e.value,
		label: e.label ?? e.value,
		...e.disabled === void 0 ? {} : { disabled: e.disabled },
		...e.count === void 0 ? {} : { trailing: /* @__PURE__ */ (0, q.jsx)("span", { children: e.count.toLocaleString() }) }
	}));
}
function Si({ filter: e, grow: t }) {
	return /* @__PURE__ */ (0, q.jsx)("div", {
		className: i("min-w-[14rem]", t && "flex-1", e.className),
		children: /* @__PURE__ */ (0, q.jsx)(ci, {
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
function Ci(e) {
	return i("h-8 w-full rounded-md border border-input bg-background px-2 text-sm text-foreground outline-none placeholder:text-placeholder focus-visible:ring-2 focus-visible:ring-ring", e && "cursor-not-allowed opacity-60");
}
function wi({ filter: e }) {
	let [t, n] = $i(e.value, e.onChange);
	return /* @__PURE__ */ (0, q.jsx)("input", {
		id: bi(e),
		type: "text",
		"aria-label": e.label,
		className: Ci(e.disabled),
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		value: t,
		disabled: e.disabled,
		onChange: (e) => n(e.target.value)
	});
}
function Ti({ filter: e }) {
	let [t, n] = $i(e.value, e.onChange);
	return e.inputType === "date" ? /* @__PURE__ */ (0, q.jsx)(Wn, {
		id: bi(e),
		"aria-label": e.label,
		className: "w-full",
		inputClassName: Ci(e.disabled),
		buttonClassName: "right-1",
		placeholder: e.placeholder,
		value: t,
		disabled: e.disabled,
		onChange: n
	}) : e.inputType === "number" ? /* @__PURE__ */ (0, q.jsx)("input", {
		id: bi(e),
		type: "number",
		"aria-label": e.label,
		className: Ci(e.disabled),
		placeholder: e.placeholder,
		value: t,
		disabled: e.disabled,
		onChange: (e) => n(e.target.value)
	}) : /* @__PURE__ */ (0, q.jsx)(kn, {
		id: bi(e),
		options: xi(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		size: "sm",
		className: "w-full",
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function Ei({ filter: e }) {
	return /* @__PURE__ */ (0, q.jsx)(kn, {
		multiple: !0,
		id: bi(e),
		options: xi(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		footer: Vi(e),
		size: "sm",
		className: "w-full",
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function Di(e) {
	return e.map((e) => ({
		value: e.value,
		label: e.label ?? e.value
	}));
}
function Oi({ filter: e }) {
	return /* @__PURE__ */ (0, q.jsx)(kn, {
		id: bi(e),
		options: Di(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		size: "sm",
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		className: "w-full",
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function ki({ filter: e }) {
	return /* @__PURE__ */ (0, q.jsx)("div", {
		className: "flex h-8 items-center",
		children: /* @__PURE__ */ (0, q.jsx)("input", {
			id: bi(e),
			type: "checkbox",
			"aria-label": e.label,
			className: "h-4 w-4 accent-primary",
			checked: e.value,
			disabled: e.disabled,
			onChange: (t) => e.onChange(t.target.checked)
		})
	});
}
function Ai({ filter: e }) {
	return /* @__PURE__ */ (0, q.jsx)("div", {
		className: "flex h-8 items-center",
		children: /* @__PURE__ */ (0, q.jsx)(Or, {
			id: bi(e),
			value: e.value,
			onChange: e.onChange,
			label: e.label,
			size: "md",
			...e.labels ? { labels: e.labels } : {},
			...e.disabled ? { disabled: e.disabled } : {}
		})
	});
}
function ji({ filter: e }) {
	return /* @__PURE__ */ (0, q.jsx)(qn, {
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
function Mi({ kind: e, label: t, from: n = "", to: r = "", onApply: i, presets: a, timeEnabled: o, timeZone: s, timeZones: c, fromPlaceholder: l, toPlaceholder: u, emptyLabel: d, disabled: f }) {
	return /* @__PURE__ */ (0, q.jsx)("div", {
		className: "w-72 p-3 text-popover-foreground",
		children: /* @__PURE__ */ (0, q.jsx)(sr, {
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
function Ni({ filter: e, grow: t }) {
	return /* @__PURE__ */ (0, q.jsx)(kn, {
		...on(e),
		options: Di(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		size: "sm",
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		className: i(sn(t), e.className),
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function Pi({ filter: e }) {
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		title: e.description,
		className: i("flex h-8 shrink-0 items-center gap-2 rounded-md border border-input bg-muted/30 px-2 text-xs", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, q.jsx)(Or, {
			value: e.value,
			onChange: e.onChange,
			label: e.label,
			...e.labels ? { labels: e.labels } : {},
			...e.disabled ? { disabled: e.disabled } : {}
		}), /* @__PURE__ */ (0, q.jsx)(Ri, {
			icon: e.icon,
			label: e.label
		})]
	});
}
function Fi({ filter: e }) {
	return /* @__PURE__ */ (0, q.jsxs)("label", {
		title: e.description,
		className: i("flex h-8 shrink-0 items-center gap-2 rounded-md border border-input bg-muted/30 px-2 text-xs", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, q.jsx)("input", {
			type: "checkbox",
			"aria-label": e.label,
			className: "h-3.5 w-3.5 accent-primary",
			checked: e.value,
			disabled: e.disabled,
			onChange: (t) => e.onChange(t.target.checked)
		}), /* @__PURE__ */ (0, q.jsx)(Ri, {
			icon: e.icon,
			label: e.label
		})]
	});
}
function Ii({ search: e }) {
	let [t, n] = $i(e.value, e.onChange);
	return /* @__PURE__ */ (0, q.jsx)("div", {
		className: "flex min-w-0 flex-1 items-center gap-2 md:flex-[1_1_14rem] md:min-w-[14rem] md:max-w-[24rem]",
		children: /* @__PURE__ */ (0, q.jsxs)("label", {
			className: i("flex h-8 min-w-0 flex-1 items-center rounded-md border border-input bg-background px-3 text-sm", e.className),
			children: [t.trim() ? /* @__PURE__ */ (0, q.jsx)("span", {
				className: "mr-2 whitespace-nowrap text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
				children: e.ariaLabel ?? "Search"
			}) : /* @__PURE__ */ (0, q.jsx)(E, {
				icon: Se,
				className: "mr-2 shrink-0 text-muted-foreground"
			}), /* @__PURE__ */ (0, q.jsx)("input", {
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
function Li({ filter: e, grow: t }) {
	let [n, r] = $i(e.value, e.onChange);
	return /* @__PURE__ */ (0, q.jsxs)("label", {
		title: e.description,
		className: i("flex h-8 items-center gap-2 rounded-md border border-input bg-muted/30 pl-2 pr-2 text-xs", t ? "min-w-[12rem] max-w-[18rem] flex-1" : "min-w-[11rem] max-w-[15rem] shrink-0", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, q.jsx)(Ri, {
			icon: e.icon,
			label: e.label
		}), /* @__PURE__ */ (0, q.jsx)("input", {
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
function Ri({ icon: e, label: t }) {
	return /* @__PURE__ */ (0, q.jsxs)("span", {
		className: "flex items-center gap-1 whitespace-nowrap font-medium uppercase tracking-wide text-muted-foreground",
		children: [/* @__PURE__ */ (0, q.jsx)(U, {
			icon: e,
			className: "text-[13px] normal-case"
		}), t]
	});
}
function zi({ filter: e, grow: t }) {
	let [n, r] = $i(e.value, e.onChange);
	if (e.inputType === "date" || e.inputType === "number") {
		let a = e.inputType === "date" && n ? se(n) : void 0;
		return /* @__PURE__ */ (0, q.jsxs)("label", {
			title: a ?? e.description,
			className: i("flex h-8 items-center gap-2 rounded-md border border-input bg-muted/30 pl-2 pr-2 text-xs", sn(t), e.disabled && "opacity-60", e.className),
			children: [/* @__PURE__ */ (0, q.jsx)(Ri, {
				icon: e.icon,
				label: e.label
			}), e.inputType === "date" ? /* @__PURE__ */ (0, q.jsx)(Wn, {
				"aria-label": e.label,
				className: "w-full",
				inputClassName: "w-full min-w-0 border-0 bg-transparent px-0 pr-6 text-sm text-foreground shadow-none focus-visible:ring-0",
				buttonClassName: "right-0",
				placeholder: e.placeholder,
				value: n,
				disabled: e.disabled,
				onChange: r
			}) : /* @__PURE__ */ (0, q.jsx)("input", {
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
	return /* @__PURE__ */ (0, q.jsx)(kn, {
		...on(e),
		options: xi(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		size: "sm",
		className: i(sn(t), e.className),
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function Bi({ filter: e, grow: t }) {
	return /* @__PURE__ */ (0, q.jsx)(kn, {
		multiple: !0,
		...on(e),
		options: xi(e.options),
		value: e.value,
		onChange: e.onChange,
		allowCustomValue: !1,
		footer: Vi(e),
		size: "sm",
		className: i(sn(t), e.className),
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...e.onSearch === void 0 ? {} : { onSearch: e.onSearch },
		...e.loading === void 0 ? {} : { loading: e.loading },
		...e.disabled === void 0 ? {} : { disabled: e.disabled }
	});
}
function Vi(e) {
	let t = e.truncated && e.total ? Math.max(e.total - e.options.length, 0) : 0;
	if (t !== 0) return /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
		"… and ",
		t.toLocaleString(),
		" more",
		e.onSearch ? " — type to search all" : ""
	] });
}
function Hi({ filter: e, grow: t }) {
	let [n, r] = (0, K.useState)(!1), o = Qi(n, r), s = na(e), [c, l] = ea(e.value, (t) => ua(e, t)), u = sa(oa(c.min) ?? s.min, s.min, s.max), d = sa(oa(c.max) ?? s.max, s.min, s.max), f = Math.min(u, d), p = Math.max(u, d), m = ra(e, s, c), h = e.domainMin !== void 0 && e.domainMax !== void 0;
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		title: e.description,
		className: i("inline-flex min-w-0", t ? "min-w-[8rem] max-w-[12rem] flex-1" : "shrink-0", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, q.jsxs)(a, {
			ref: o.refs.setReference,
			type: "button",
			variant: "outline",
			size: "sm",
			"aria-label": `${e.label} filter`,
			disabled: e.disabled,
			className: i("min-w-0 gap-2 font-normal", t ? "w-full max-w-[12rem] justify-between" : "w-auto max-w-[9.5rem] px-2.5", m === e.label && "text-muted-foreground"),
			...o.getReferenceProps(),
			children: [
				/* @__PURE__ */ (0, q.jsx)(U, {
					icon: e.icon,
					className: "text-[14px] text-muted-foreground"
				}),
				/* @__PURE__ */ (0, q.jsx)("span", {
					className: "truncate",
					children: m
				}),
				/* @__PURE__ */ (0, q.jsx)(E, {
					icon: n ? H : F,
					className: "text-muted-foreground"
				})
			]
		}), n && /* @__PURE__ */ (0, q.jsx)(O, { children: /* @__PURE__ */ (0, q.jsx)(he, {
			context: o.context,
			modal: !1,
			children: /* @__PURE__ */ (0, q.jsxs)("div", {
				ref: o.refs.setFloating,
				role: "dialog",
				"aria-label": e.label,
				style: {
					...o.floatingStyles,
					zIndex: o.floatingZ
				},
				className: "min-w-[18rem] max-w-[22rem] rounded-md border border-border bg-popover p-3 text-popover-foreground shadow-lg shadow-black/5",
				...o.getFloatingProps(),
				children: [/* @__PURE__ */ (0, q.jsxs)("div", {
					className: "mb-3 flex items-center justify-between gap-2",
					children: [/* @__PURE__ */ (0, q.jsx)("div", {
						className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
						children: e.label
					}), /* @__PURE__ */ (0, q.jsx)("button", {
						type: "button",
						className: "text-[10px] text-primary disabled:text-muted-foreground",
						onClick: () => {
							l({}), ua(e, {});
						},
						disabled: !String(c.min ?? "").trim() && !String(c.max ?? "").trim(),
						children: "Clear all"
					})]
				}), /* @__PURE__ */ (0, q.jsxs)("div", {
					className: "space-y-3",
					children: [h && /* @__PURE__ */ (0, q.jsxs)("div", {
						className: "space-y-2",
						children: [/* @__PURE__ */ (0, q.jsxs)("div", {
							className: "flex items-center justify-between text-[11px] text-muted-foreground",
							children: [/* @__PURE__ */ (0, q.jsx)("span", { children: ca(s.min, e) }), /* @__PURE__ */ (0, q.jsx)("span", { children: ca(s.max, e) })]
						}), /* @__PURE__ */ (0, q.jsx)("div", {
							className: "relative h-6",
							children: /* @__PURE__ */ (0, q.jsx)(er, {
								min: s.min,
								max: s.max,
								step: s.step,
								value: [f, p],
								ariaLabelMin: `${e.label} minimum slider`,
								ariaLabelMax: `${e.label} maximum slider`,
								onChange: ([t, n]) => l(aa([t, n], s, e))
							})
						})]
					}), /* @__PURE__ */ (0, q.jsx)(Wi, {
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
function Ui({ filter: e, chrome: t = "full" }) {
	let n = na(e), [r, a] = ea(e.value, (t) => ua(e, t)), o = sa(oa(r.min) ?? n.min, n.min, n.max), s = sa(oa(r.max) ?? n.max, n.min, n.max), c = Math.min(o, s), l = Math.max(o, s), u = t === "embedded", d = e.domainMin !== void 0 && e.domainMax !== void 0;
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		"data-filter-panel-chrome": t,
		className: i("min-w-[18rem] max-w-[22rem] text-popover-foreground", u ? "p-0" : "rounded-md border border-border bg-popover p-3 shadow-sm shadow-black/5"),
		children: [!u && /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "mb-3 flex items-center justify-between gap-2",
			children: [/* @__PURE__ */ (0, q.jsx)("div", {
				className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
				children: e.label
			}), /* @__PURE__ */ (0, q.jsx)("button", {
				type: "button",
				className: "text-[10px] text-primary disabled:text-muted-foreground",
				onClick: () => {
					a({}), ua(e, {});
				},
				disabled: !String(r.min ?? "").trim() && !String(r.max ?? "").trim(),
				children: "Clear all"
			})]
		}), /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "space-y-3",
			children: [d && /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "space-y-2",
				children: [/* @__PURE__ */ (0, q.jsxs)("div", {
					className: "flex items-center justify-between text-[11px] text-muted-foreground",
					children: [/* @__PURE__ */ (0, q.jsx)("span", { children: ca(n.min, e) }), /* @__PURE__ */ (0, q.jsx)("span", { children: ca(n.max, e) })]
				}), /* @__PURE__ */ (0, q.jsx)("div", {
					className: "relative h-6",
					children: /* @__PURE__ */ (0, q.jsx)(er, {
						min: n.min,
						max: n.max,
						step: n.step,
						value: [c, l],
						ariaLabelMin: `${e.label} minimum slider`,
						ariaLabelMax: `${e.label} maximum slider`,
						onChange: ([t, r]) => a(aa([t, r], n, e))
					})
				})]
			}), /* @__PURE__ */ (0, q.jsx)(Wi, {
				filter: e,
				value: r,
				onChange: a,
				bounds: n
			})]
		})]
	});
}
function Wi({ filter: e, value: t, onChange: n, bounds: r }) {
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "space-y-2",
		children: [/* @__PURE__ */ (0, q.jsx)(Gi, {
			filter: e,
			value: t,
			onChange: n,
			bounds: r,
			edge: "min"
		}), /* @__PURE__ */ (0, q.jsx)(Gi, {
			filter: e,
			value: t,
			onChange: n,
			bounds: r,
			edge: "max"
		})]
	});
}
function Gi({ filter: e, value: t, onChange: n, bounds: r, edge: i }) {
	let a = i === "min", o = a ? "minimum" : "maximum", s = a ? t.minOperator ?? ">=" : t.maxOperator ?? "<=", c = a ? t.min ?? "" : t.max ?? "", l = e.kind === "duration" ? t : void 0, u = a ? l?.minUnit : l?.maxUnit, d = (r) => {
		let i = a ? {
			...t,
			min: r
		} : {
			...t,
			max: r
		};
		n(e.kind === "duration" ? i : ia(i, a ? "min-input" : "max-input"));
	}, f = (e) => n(a ? {
		...t,
		minOperator: e
	} : {
		...t,
		maxOperator: e
	});
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "grid grid-cols-[2.5rem_3.5rem_minmax(0,1fr)_auto] items-center gap-1.5",
		children: [
			/* @__PURE__ */ (0, q.jsx)("span", {
				className: "text-[10px] text-muted-foreground",
				children: a ? "Min" : "Max"
			}),
			/* @__PURE__ */ (0, q.jsx)("select", {
				"aria-label": `${e.label} ${o} operator`,
				className: "h-8 rounded-md border border-input bg-background px-1 text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring",
				value: s,
				disabled: e.disabled,
				onChange: (e) => f(e.target.value),
				children: (a ? [">=", ">"] : ["<=", "<"]).map((e) => /* @__PURE__ */ (0, q.jsx)("option", {
					value: e,
					children: e
				}, e))
			}),
			/* @__PURE__ */ (0, q.jsx)("input", {
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
			/* @__PURE__ */ (0, q.jsx)(Ki, {
				filter: e,
				value: t,
				edge: i,
				...u ? { unit: u } : {},
				onChange: n
			})
		]
	});
}
function Ki({ filter: e, value: t, edge: n, unit: r, onChange: i }) {
	if (e.kind === "number") return e.unit ? /* @__PURE__ */ (0, q.jsx)("span", {
		className: "min-w-8 text-xs text-muted-foreground",
		children: e.unit
	}) : null;
	let a = r ?? e.unit ?? "ms";
	return /* @__PURE__ */ (0, q.jsx)("select", {
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
		].map((e) => /* @__PURE__ */ (0, q.jsx)("option", {
			value: e,
			children: e
		}, e))
	});
}
function qi({ filter: e, chrome: t = "full" }) {
	let [n, r] = (0, K.useState)(e.groups[0]?.groupKey ?? null), [a, o] = $t(e.value, e.onChange), s = e.groups, c = t === "embedded", l = (0, K.useMemo)(() => s.find((e) => e.groupKey === n) ?? s[0] ?? null, [s, n]), u = (0, K.useMemo)(() => {
		let e = {};
		for (let t of s) {
			let n = 0;
			for (let e of t.options) a[e.value] && (n += 1);
			e[t.groupKey] = n;
		}
		return e;
	}, [s, a]), d = (0, K.useMemo)(() => {
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
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		role: "dialog",
		className: "flex",
		children: [/* @__PURE__ */ (0, q.jsxs)("div", {
			"data-filter-panel-chrome": t,
			className: i("min-w-[14rem] max-w-[16rem] text-popover-foreground", c ? "p-0" : "rounded-md border border-border bg-popover p-2 shadow-sm shadow-black/5"),
			children: [!c && /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "mb-2 flex items-center justify-between gap-2",
				children: [/* @__PURE__ */ (0, q.jsx)("div", {
					className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
					children: e.label
				}), /* @__PURE__ */ (0, q.jsx)("button", {
					type: "button",
					className: "text-[10px] text-primary disabled:text-muted-foreground",
					onClick: () => o({}),
					disabled: Object.keys(a).length === 0,
					children: "Clear all"
				})]
			}), /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "max-h-72 space-y-0.5 overflow-auto",
				children: [d.map((e) => {
					let t = u[e.groupKey] ?? 0, a = e.groupKey === n;
					return /* @__PURE__ */ (0, q.jsxs)("div", {
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
							/* @__PURE__ */ (0, q.jsx)("span", {
								className: "min-w-0 flex-1 truncate",
								children: e.label ?? e.groupKey
							}),
							t > 0 && /* @__PURE__ */ (0, q.jsxs)("span", {
								className: "rounded-full bg-primary/15 px-1.5 text-[10px] font-medium text-primary",
								children: [
									t,
									"/",
									e.options.length
								]
							}),
							/* @__PURE__ */ (0, q.jsx)(E, {
								icon: N,
								className: "shrink-0 text-muted-foreground"
							})
						]
					}, e.groupKey);
				}), s.length === 0 && /* @__PURE__ */ (0, q.jsx)("div", {
					className: "px-2 py-3 text-sm text-muted-foreground",
					children: "No groups"
				})]
			})]
		}), l && /* @__PURE__ */ (0, q.jsxs)("div", {
			onMouseEnter: () => r(l.groupKey),
			className: "ml-1.5 min-w-[16rem] max-w-[20rem] rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-sm shadow-black/5",
			children: [/* @__PURE__ */ (0, q.jsxs)("div", {
				className: "mb-2 flex items-center justify-between gap-2",
				children: [/* @__PURE__ */ (0, q.jsx)("div", {
					className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
					children: l.label ?? l.groupKey
				}), /* @__PURE__ */ (0, q.jsx)("button", {
					type: "button",
					className: "text-[10px] text-primary disabled:text-muted-foreground",
					onClick: () => f(l.groupKey),
					disabled: (u[l.groupKey] ?? 0) === 0,
					children: "Clear"
				})]
			}), /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "max-h-72 space-y-0.5 overflow-auto",
				children: [l.options.map((e) => {
					let t = a[e.value] ?? "neutral", n = e.title ?? Zt(e);
					return /* @__PURE__ */ (0, q.jsx)("div", {
						role: "button",
						tabIndex: 0,
						"data-filter-option": e.value,
						className: "rounded-md px-1.5 py-0.5 hover:bg-accent/50 focus-visible:bg-accent/50 focus-visible:outline-none",
						onClick: () => o(nn(a, e.value, rn(t))),
						onKeyDown: (n) => {
							n.key === "Enter" || n.key === " " ? (n.preventDefault(), o(nn(a, e.value, rn(t)))) : n.key === "ArrowLeft" && (n.preventDefault(), r(null));
						},
						children: /* @__PURE__ */ (0, q.jsx)(Kt, {
							className: "w-full justify-between",
							label: Yi(e, l.groupKey),
							mode: t,
							title: n,
							togglePosition: "right",
							onModeChange: (t) => o(nn(a, e.value, t))
						})
					}, e.value);
				}), l.options.length === 0 && /* @__PURE__ */ (0, q.jsx)("div", {
					className: "px-2 py-3 text-sm text-muted-foreground",
					children: "No values"
				})]
			})]
		})]
	});
}
function Ji({ filter: e, grow: t }) {
	let [n, r] = (0, K.useState)(!1), [o, s] = (0, K.useState)(null), [c, l] = $t(e.value, e.onChange), u = Qi(n, (e) => {
		r(e), e || s(null);
	}), d = Qt(e.label, c), f = e.groups, p = (0, K.useMemo)(() => f.find((e) => e.groupKey === o) ?? null, [f, o]), m = (0, K.useMemo)(() => {
		let e = {};
		for (let t of f) {
			let n = 0;
			for (let e of t.options) c[e.value] && (n += 1);
			e[t.groupKey] = n;
		}
		return e;
	}, [f, c]), h = (0, K.useMemo)(() => {
		let e = [], t = [];
		for (let n of f) (m[n.groupKey] ?? 0) > 0 ? e.push(n) : t.push(n);
		return [...e, ...t];
	}, [f, m]), g = (e) => {
		let t = f.find((t) => t.groupKey === e);
		if (!t) return;
		let n = { ...c };
		for (let e of t.options) delete n[e.value];
		l(n);
	};
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		title: e.description,
		className: i("inline-flex min-w-0", t ? "min-w-[8rem] max-w-[12rem] flex-1" : "shrink-0", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, q.jsxs)(a, {
			ref: u.refs.setReference,
			type: "button",
			variant: "outline",
			size: "sm",
			"aria-label": `${e.label} filter`,
			disabled: e.disabled,
			className: i("min-w-0 gap-2 font-normal", t ? "w-full max-w-[12rem] justify-between" : "w-auto max-w-[8.5rem] px-2.5", d === e.label && "text-muted-foreground"),
			...u.getReferenceProps(),
			children: [/* @__PURE__ */ (0, q.jsx)("span", {
				className: "truncate",
				children: d
			}), /* @__PURE__ */ (0, q.jsx)(E, {
				icon: n ? H : F,
				className: "text-muted-foreground"
			})]
		}), n && /* @__PURE__ */ (0, q.jsx)(O, { children: /* @__PURE__ */ (0, q.jsx)(he, {
			context: u.context,
			modal: !1,
			children: /* @__PURE__ */ (0, q.jsxs)("div", {
				ref: u.refs.setFloating,
				role: "dialog",
				"aria-label": e.label,
				style: {
					...u.floatingStyles,
					zIndex: u.floatingZ
				},
				className: "flex",
				...u.getFloatingProps(),
				children: [/* @__PURE__ */ (0, q.jsxs)("div", {
					className: "min-w-[14rem] max-w-[16rem] rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-lg shadow-black/5",
					children: [/* @__PURE__ */ (0, q.jsxs)("div", {
						className: "mb-2 flex items-center justify-between gap-2",
						children: [/* @__PURE__ */ (0, q.jsx)("div", {
							className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
							children: e.label
						}), /* @__PURE__ */ (0, q.jsx)("button", {
							type: "button",
							className: "text-[10px] text-primary disabled:text-muted-foreground",
							onClick: () => l({}),
							disabled: Object.keys(c).length === 0,
							children: "Clear all"
						})]
					}), /* @__PURE__ */ (0, q.jsxs)("div", {
						className: "max-h-72 space-y-0.5 overflow-auto",
						children: [h.map((e) => {
							let t = m[e.groupKey] ?? 0, n = e.groupKey === o;
							return /* @__PURE__ */ (0, q.jsxs)("div", {
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
									/* @__PURE__ */ (0, q.jsx)("span", {
										className: "min-w-0 flex-1 truncate",
										children: e.label ?? e.groupKey
									}),
									t > 0 && /* @__PURE__ */ (0, q.jsxs)("span", {
										className: "rounded-full bg-primary/15 px-1.5 text-[10px] font-medium text-primary",
										children: [
											t,
											"/",
											e.options.length
										]
									}),
									/* @__PURE__ */ (0, q.jsx)(E, {
										icon: N,
										className: "shrink-0 text-muted-foreground"
									})
								]
							}, e.groupKey);
						}), f.length === 0 && /* @__PURE__ */ (0, q.jsx)("div", {
							className: "px-2 py-3 text-sm text-muted-foreground",
							children: "No groups"
						})]
					})]
				}), p && /* @__PURE__ */ (0, q.jsxs)("div", {
					onMouseEnter: () => s(p.groupKey),
					className: "ml-1.5 min-w-[16rem] max-w-[20rem] rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-lg shadow-black/5",
					children: [/* @__PURE__ */ (0, q.jsxs)("div", {
						className: "mb-2 flex items-center justify-between gap-2",
						children: [/* @__PURE__ */ (0, q.jsx)("div", {
							className: "text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
							children: p.label ?? p.groupKey
						}), /* @__PURE__ */ (0, q.jsx)("button", {
							type: "button",
							className: "text-[10px] text-primary disabled:text-muted-foreground",
							onClick: () => g(p.groupKey),
							disabled: (m[p.groupKey] ?? 0) === 0,
							children: "Clear"
						})]
					}), /* @__PURE__ */ (0, q.jsxs)("div", {
						className: "max-h-72 space-y-0.5 overflow-auto",
						children: [p.options.map((e) => {
							let t = c[e.value] ?? "neutral", n = e.title ?? Zt(e);
							return /* @__PURE__ */ (0, q.jsx)("div", {
								role: "button",
								tabIndex: 0,
								"data-filter-option": e.value,
								className: "rounded-md px-1.5 py-0.5 hover:bg-accent/50 focus-visible:bg-accent/50 focus-visible:outline-none",
								onClick: () => l(nn(c, e.value, rn(t))),
								onKeyDown: (n) => {
									n.key === "Enter" || n.key === " " ? (n.preventDefault(), l(nn(c, e.value, rn(t)))) : n.key === "ArrowLeft" && (n.preventDefault(), s(null));
								},
								children: /* @__PURE__ */ (0, q.jsx)(Kt, {
									className: "w-full justify-between",
									label: Yi(e, p.groupKey),
									mode: t,
									title: n,
									togglePosition: "right",
									onModeChange: (t) => l(nn(c, e.value, t))
								})
							}, e.value);
						}), p.options.length === 0 && /* @__PURE__ */ (0, q.jsx)("div", {
							className: "px-2 py-3 text-sm text-muted-foreground",
							children: "No values"
						})]
					})]
				})]
			})
		}) })]
	});
}
function Yi(e, t) {
	if (typeof e.label == "string") {
		let n = `${t}=`;
		return e.label.startsWith(n) ? e.label.slice(n.length) : e.label;
	}
	return e.label ?? e.value;
}
function Xi({ filter: e, grow: t }) {
	return /* @__PURE__ */ (0, q.jsxs)("label", {
		title: e.description,
		className: i("flex h-8 items-center gap-2 rounded-md border border-input bg-muted/30 pl-2 pr-1 text-xs", t ? "min-w-[12rem] max-w-[18rem] flex-1" : "min-w-[11rem] max-w-[15rem] shrink-0", e.disabled && "opacity-60", e.className),
		children: [/* @__PURE__ */ (0, q.jsx)("span", {
			className: "whitespace-nowrap font-medium uppercase tracking-wide text-muted-foreground",
			children: e.label
		}), /* @__PURE__ */ (0, q.jsx)(qn, {
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
function Zi({ kind: e, label: t, from: n = "", to: r = "", onApply: i, presets: a, timeEnabled: o, timeZone: s, timeZones: c, fromPlaceholder: l, toPlaceholder: u, emptyLabel: d, className: f, disabled: p }) {
	let { compact: m } = (0, K.useContext)(Yt);
	return /* @__PURE__ */ (0, q.jsx)(sr, {
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
function Qi(e, t, n = "bottom-start") {
	let r = re(), { refs: i, floatingStyles: a, context: o } = m({
		open: e,
		onOpenChange: t,
		placement: n,
		whileElementsMounted: Ve,
		middleware: [
			Be(6),
			He({ padding: 8 }),
			Ue({ padding: 8 })
		]
	}), { getReferenceProps: s, getFloatingProps: c } = me([
		W(o),
		S(o, { escapeKey: !1 }),
		g(o, { role: "dialog" })
	]);
	return x(e, () => {
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
function $i(e, t) {
	let { autoSubmit: n } = (0, K.useContext)(Yt), [r, i] = (0, K.useState)(e), a = (0, K.useRef)(t), o = (0, K.useRef)(e), s = (0, K.useRef)(!1);
	return (0, K.useEffect)(() => {
		a.current = t;
	}, [t]), (0, K.useEffect)(() => {
		o.current !== e && (o.current = e, s.current = !0, i(e));
	}, [e]), (0, K.useEffect)(() => {
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
function ea(e, t) {
	let { autoSubmit: n } = (0, K.useContext)(Yt), [r, i] = (0, K.useState)(e), a = (0, K.useRef)(t);
	return (0, K.useEffect)(() => {
		a.current = t;
	}, [t]), (0, K.useEffect)(() => {
		i(e);
	}, [
		e.max,
		e.maxOperator,
		e.maxUnit,
		e.min,
		e.minOperator,
		e.minUnit
	]), (0, K.useEffect)(() => {
		if (ta(r, e)) return;
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
function ta(e, t) {
	let n = e, r = t;
	return (e.min ?? "") === (t.min ?? "") && (e.minOperator ?? ">=") === (t.minOperator ?? ">=") && (n.minUnit ?? "") === (r.minUnit ?? "") && (e.max ?? "") === (t.max ?? "") && (e.maxOperator ?? "<=") === (t.maxOperator ?? "<=") && (n.maxUnit ?? "") === (r.maxUnit ?? "");
}
function na(e) {
	let t = oa(e.value.min), n = oa(e.value.max), r = t ?? 0, i = n ?? Math.max(r + 100, 100), a = e.domainMin ?? Math.min(r, i), o = e.domainMax ?? Math.max(i, a + 100), s = e.step && e.step > 0 ? e.step : 1;
	return {
		min: a,
		max: o <= a ? a + s : o,
		step: s
	};
}
function ra(e, t, n) {
	let r = oa(n.min), i = oa(n.max), a = String(n.min ?? "").trim() !== "", o = String(n.max ?? "").trim() !== "";
	if (!a && !o) return e.label;
	let s = la(r ?? t.min, e, n, "min"), c = la(i ?? t.max, e, n, "max");
	return a && o ? `${e.label} ${n.minOperator ?? ">="}${s}, ${n.maxOperator ?? "<="}${c}` : a ? `${e.label} ${n.minOperator ?? ">="}${s}` : `${e.label} ${n.maxOperator ?? "<="}${c}`;
}
function ia(e, t) {
	let n = e.min ?? "", r = e.max ?? "", i = oa(n), a = oa(r);
	return i != null && a != null && i > a && (t.startsWith("min") ? r = n : n = r), {
		...e,
		min: n,
		max: r
	};
}
function aa(e, t, n) {
	let [r, i] = e, a = {
		min: r <= t.min ? "" : da(sa(r, t.min, t.max)),
		minOperator: ">=",
		max: i >= t.max ? "" : da(sa(i, t.min, t.max)),
		maxOperator: "<="
	};
	return n.kind === "duration" && (a.minUnit = n.unit ?? "ms", a.maxUnit = n.unit ?? "ms"), a;
}
function oa(e) {
	if (!e?.trim()) return null;
	let t = Number(e);
	return Number.isFinite(t) ? t : null;
}
function sa(e, t, n) {
	return Math.min(n, Math.max(t, e));
}
function ca(e, t) {
	return t.kind === "number" && t.formatValue ? t.formatValue(e) : t.kind === "duration" ? `${da(e)} ${t.unit ?? "ms"}` : t.unit ? `${da(e)} ${t.unit}` : da(e);
}
function la(e, t, n, r) {
	if (t.kind !== "duration") return ca(e, t);
	let i = n, a = r === "min" ? i.minUnit : i.maxUnit;
	return `${da(e)} ${a ?? t.unit ?? "ms"}`;
}
function ua(e, t) {
	e.onChange(t);
}
function da(e) {
	return String(Number.isInteger(e) ? e : Number(e.toFixed(6)));
}
function fa(e) {
	return Object.fromEntries(e.map((e) => [e.key, ma(e)]));
}
function pa(e) {
	return e.kind !== "date-range";
}
function ma(e) {
	if (pa(e)) return e.value;
}
function ha(e, t, n) {
	return pa(e) ? e.kind === "text" || e.kind === "lookup" || e.kind === "workload" ? {
		...e,
		value: String(t ?? ""),
		onChange: (e) => n(e)
	} : e.kind === "lookup-multi" ? {
		...e,
		value: Array.isArray(t) ? t.map(String) : [],
		onChange: (e) => n(e)
	} : e.kind === "multi" || e.kind === "tree" || e.kind === "nested-multi" ? {
		...e,
		value: ba(t) ? t : {},
		onChange: (e) => n(e)
	} : e.kind === "select-multi" ? {
		...e,
		value: Array.isArray(t) ? t.map(String) : [],
		onChange: (e) => n(e)
	} : e.kind === "number" || e.kind === "duration" ? {
		...e,
		value: ya(t) ? t : {},
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
function ga(e, t) {
	for (let n of e) {
		if (!pa(n)) continue;
		let e = Object.hasOwn(t, n.key) ? t[n.key] : ma(n);
		va(ma(n), e) || _a(n, e);
	}
}
function _a(e, t) {
	pa(e) && (e.kind === "text" || e.kind === "lookup" || e.kind === "workload" || e.kind === "enum" ? e.onChange(String(t ?? "")) : e.kind === "lookup-multi" || e.kind === "select-multi" ? e.onChange(Array.isArray(t) ? t.map(String) : []) : e.kind === "number" || e.kind === "duration" ? ua(e, ya(t) ? t : {}) : e.kind === "boolean" ? e.onChange(!!t) : e.kind === "tristate" ? e.onChange(t === void 0 ? void 0 : !!t) : e.onChange(ba(t) ? t : {}));
}
function va(e, t) {
	return JSON.stringify(e) === JSON.stringify(t);
}
function ya(e) {
	return typeof e == "object" && !!e && !Array.isArray(e);
}
function ba(e) {
	return typeof e == "object" && !!e && !Array.isArray(e);
}
function xa(e, t) {
	if (e.length === 0) return 0;
	let n = 0, r = 0;
	for (let i of e) {
		let e = n + (r > 0 ? li : 0) + i;
		if (e > t) break;
		n = e, r += 1;
	}
	return r;
}
function Sa(e) {
	return e.length === 0 ? 0 : e.reduce((e, t) => e + t, 0) + (e.length - 1) * li;
}
function Ca(e) {
	return e.kind === "boolean" ? Math.max(88, e.label.length * 8 + 40) : e.kind === "multi" || e.kind === "tree" ? 176 : e.kind === "nested-multi" ? 136 : e.kind === "workload" ? 224 : e.kind === "number" || e.kind === "duration" ? 152 : Math.max(144, e.label.length * 8 + 96);
}
// ../../../clicky-ui/packages/ui/src/data/use-infinite-rows.ts
function wa(e, t) {
	let n = (0, K.useRef)(e);
	n.current = e;
	let r = (0, K.useRef)(!1), i = e?.loading ?? !1;
	return (0, K.useEffect)(() => {
		r.current = !1;
	}, [i, t]), (0, K.useCallback)(() => {
		let e = n.current;
		e && !e.loading && e.hasMore && !r.current && (r.current = !0, e.onLoadMore());
	}, []);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@tanstack/virtual-core/3.17.8/7c4c8a2dd3f012a209ac831b3d885a8aa02d549fe4515126bf4c5e475ccaaec2/node_modules/@tanstack/virtual-core/dist/esm/lazy-measurements.js
function Ta(e, t, n) {
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
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@tanstack/virtual-core/3.17.8/7c4c8a2dd3f012a209ac831b3d885a8aa02d549fe4515126bf4c5e475ccaaec2/node_modules/@tanstack/virtual-core/dist/esm/utils.js
function Ea(e, t, n) {
	let r = n.initialDeps ?? [], i, a = !0;
	function o() {
		let o = e();
		return (o.length !== r.length || o.some((e, t) => r[t] !== e)) && (r = o, i = t(...o), n?.onChange && !(a && n.skipInitialOnChange) && n.onChange(i), a = !1), i;
	}
	return o.updateDeps = (e) => {
		r = e;
	}, o;
}
function Da(e, t) {
	if (e === void 0) throw Error(`Unexpected undefined${t ? `: ${t}` : ""}`);
	return e;
}
var Oa = (e, t) => Math.abs(e - t) < 1.01, ka = (e, t, n) => {
	let r;
	return Object.assign(function(...i) {
		e.clearTimeout(r), r = e.setTimeout(() => t.apply(this, i), n);
	}, { cancel: () => {
		e.clearTimeout(r);
	} });
}, Aa, ja = () => {
	if (Aa !== void 0) return Aa;
	if (typeof navigator > "u") return Aa = !1;
	if (/iP(hone|od|ad)/.test(navigator.userAgent)) return Aa = !0;
	let e = navigator.maxTouchPoints;
	return Aa = navigator.platform === "MacIntel" && e !== void 0 && e > 0;
}, Ma = (e) => {
	let { offsetWidth: t, offsetHeight: n } = e;
	return {
		width: t,
		height: n
	};
}, Na = (e) => e, Pa = (e) => {
	let t = Math.max(e.startIndex - e.overscan, 0), n = Math.min(e.endIndex + e.overscan, e.count - 1) - t + 1, r = Array(n);
	for (let e = 0; e < n; e++) r[e] = t + e;
	return r;
}, Fa = (e, t) => {
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
	if (i(Ma(n)), !r.ResizeObserver) return () => {};
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
			i(Ma(n));
		};
		e.options.useAnimationFrameWithResizeObserver ? requestAnimationFrame(r) : r();
	});
	return a.observe(n, { box: "border-box" }), () => {
		a.unobserve(n);
	};
}, Ia = { passive: !0 }, La = typeof window > "u" || "onscrollend" in window, Ra = (e, t, n) => {
	let r = e.scrollElement;
	if (!r) return;
	let i = e.targetWindow;
	if (!i) return;
	let a = e.options.useScrollendEvent && La, o = 0, s = a ? null : ka(i, () => t(o, !1), e.options.isScrollingResetDelay), c = (e) => () => {
		o = n(r), s?.(), t(o, e);
	}, l = c(!0), u = c(!1);
	return r.addEventListener("scroll", l, Ia), a && r.addEventListener("scrollend", u, Ia), () => {
		r.removeEventListener("scroll", l), a && r.removeEventListener("scrollend", u), s?.cancel();
	};
}, za = (e, t) => Ra(e, t, (t) => {
	let { horizontal: n, isRtl: r } = e.options;
	return n ? t.scrollLeft * (r && -1 || 1) : t.scrollTop;
}), Ba = (e, t, n) => {
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
}, Va = (e, { adjustments: t = 0, behavior: n }, r) => {
	var i, a;
	(a = (i = r.scrollElement)?.scrollTo) == null || a.call(i, {
		[r.options.horizontal ? "left" : "top"]: e + t,
		behavior: n
	});
}, Ha = class {
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
				getItemKey: Na,
				rangeExtractor: Pa,
				onChange: () => {},
				measureElement: Ba,
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
		}, this.maybeNotify = Ea(() => (this.calculateRange(), [
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
						this._iosTouching = !1, ja() && this.targetWindow != null && (this._iosJustTouchEnded = !0, this._iosTouchEndTimerId = this.targetWindow.setTimeout(() => {
							this._iosJustTouchEnded = !1, this._iosTouchEndTimerId = null, this._flushIosDeferredIfReady();
						}, 150));
					};
					e.addEventListener("touchstart", t, Ia), e.addEventListener("touchend", n, Ia), this.unsubs.push(() => {
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
				e !== null && !r && (ja() && (this.isScrolling || this._iosTouching || this._iosJustTouchEnded) ? i !== 0 && (this._iosDeferredAdjustment += i) : this._scrollToOffset(this.getScrollOffset(), {
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
		}, this.rafId = null, this.getSize = () => this.options.enabled ? (this.scrollRect = this.scrollRect ?? this.options.initialRect, this.scrollRect[this.options.horizontal ? "width" : "height"]) : (this.scrollRect = null, 0), this.getScrollOffset = () => this.options.enabled ? (this.scrollOffset = this.scrollOffset ?? (typeof this.options.initialOffset == "function" ? this.options.initialOffset() : this.options.initialOffset), this.scrollOffset) : (this.scrollOffset = null, 0), this.getMeasurementOptions = Ea(() => [
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
		}), { key: !1 }), this.isIndexInRange = (e) => e >= 0 && e < this.options.count, this.getMeasurements = Ea(() => [this.getMeasurementOptions(), this.itemSizeCacheVersion], ({ count: e, paddingStart: t, scrollMargin: n, getItemKey: r, enabled: i, lanes: a, laneAssignmentMode: o, gap: s }, c) => {
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
				let c = Ta(e, a, r);
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
		}), this.calculateRange = Ea(() => [
			this.getMeasurements(),
			this.getSize(),
			this.getScrollOffset(),
			this.options.lanes
		], (e, t, n, r) => e.length === 0 || t === 0 ? (this.range = null, null) : (this.range = Ga(e, t, n, r, r === 1 && this._flatMeasurements != null ? this._flatMeasurements : null), this.range), {
			key: !1,
			debug: () => this.options.debug
		}), this.getVirtualIndexes = Ea(() => {
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
		}, this.getVirtualItems = Ea(() => [this.getVirtualIndexes(), this.getMeasurements()], (e, t) => {
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
			return Da(t[Ua(0, t.length - 1, r ? (e) => n[e * 2] : (e) => Da(t[e]).start, e)]);
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
		return e === 0 ? !1 : ja() && (this.isScrolling || this._iosTouching || this._iosJustTouchEnded) ? (this._iosDeferredAdjustment += e, !1) : (this._scrollToOffset(this.getScrollOffset(), {
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
		if (!n && Oa(t, this.getScrollOffset())) {
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
}, Ua = (e, t, n, r) => {
	for (; e <= t;) {
		let i = (e + t) / 2 | 0, a = n(i);
		if (a < r) e = i + 1;
		else if (a > r) t = i - 1;
		else return i;
	}
	return e > 0 ? e - 1 : 0;
};
function Wa(e, t, n) {
	let r = 0;
	for (; r <= t;) {
		let i = (r + t) / 2 | 0, a = e[i * 2];
		if (a < n) r = i + 1;
		else if (a > n) t = i - 1;
		else return i;
	}
	return r > 0 ? r - 1 : 0;
}
function Ga(e, t, n, r, i) {
	let a = e.length - 1;
	if (e.length <= r) return {
		startIndex: 0,
		endIndex: a
	};
	if (r === 1 && i !== null) {
		let e = Wa(i, a, n), r = e, o = n + t;
		for (; r < a && i[r * 2] + i[r * 2 + 1] < o;) r++;
		return {
			startIndex: e,
			endIndex: r
		};
	}
	let o = Ua(0, a, (t) => e[t].start, n), s = o;
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
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@tanstack/react-virtual/3.14.10/c23a74998075603a2e1ad8151ff0cc247ee45246fe4ea1dc2a791ffedde063b8/node_modules/@tanstack/react-virtual/dist/esm/index.js
var Ka = typeof document < "u" ? K.useLayoutEffect : K.useEffect;
function qa({ useFlushSync: e = !0, directDomUpdates: t = !1, directDomUpdatesMode: n = "transform", ...r }) {
	let i = K.useReducer((e) => e + 1, 0)[1], a = K.useRef({
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
			l && (e && n ? (0, Ge.flushSync)(i) : i()), (o = r.onChange) == null || o.call(r, t, n);
		}
	}, [l] = K.useState(() => {
		let e = new Ha(c);
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
	return l.setOptions(c), Ka(() => l._didMount(), []), Ka(() => (o(l), l._willUpdate())), Ka(() => {
		s(l);
	}), l;
}
function Ja(e) {
	return qa({
		observeElementRect: Fa,
		observeElementOffset: za,
		scrollToFn: Va,
		...e
	});
}
// ../../../clicky-ui/packages/ui/src/data/use-data-table-virtual.ts
var Ya = 8, Xa = 40;
function Za(e) {
	if (!e) return null;
	let t = typeof e == "object" ? e : {};
	return {
		overscan: t.overscan ?? Ya,
		estimateRowHeight: t.estimateRowHeight ?? Xa
	};
}
function Qa(e, t, n) {
	let r = e[0], i = e[e.length - 1];
	return !r || !i ? {
		top: 0,
		bottom: 0
	} : {
		top: Math.max(0, r.start - n),
		bottom: Math.max(0, t - (i.end - n))
	};
}
function $a({ count: e, options: t, getItemKey: n }) {
	let [r, i] = (0, K.useState)(null), [a, o] = (0, K.useState)(null), [s, c] = (0, K.useState)(0);
	(0, K.useEffect)(() => {
		if (!t || !a) return;
		let e = () => c(a.offsetTop);
		if (e(), typeof ResizeObserver > "u") return;
		let n = new ResizeObserver(e);
		return a.parentElement && n.observe(a.parentElement), () => n.disconnect();
	}, [t, a]);
	let l = (0, K.useCallback)(() => t?.estimateRowHeight ?? Xa, [t?.estimateRowHeight]), u = Ja({
		count: t ? e : 0,
		getScrollElement: () => r,
		estimateSize: l,
		overscan: t?.overscan ?? Ya,
		getItemKey: n,
		scrollMargin: s
	}), d = t ? u.getVirtualItems() : [], f = (0, K.useMemo)(() => Qa(d, u.getTotalSize(), s), [d, s]);
	return {
		enabled: !!t,
		setScrollEl: i,
		setTbodyEl: o,
		virtualizer: u,
		virtualItems: d,
		spacers: f
	};
}
// ../../../clicky-ui/packages/ui/src/data/DataTableMenuSections.tsx
var eo = [
	{
		value: "compact",
		icon: d,
		label: "Compact"
	},
	{
		value: "comfortable",
		icon: ht,
		label: "Comfortable"
	},
	{
		value: "spacious",
		icon: ct,
		label: "Spacious"
	}
];
function to(e) {
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
function no({ actions: e, separated: t, onClose: n, size: r = "menu" }) {
	let a = to(e), [o, s] = (0, K.useState)(null);
	return /* @__PURE__ */ (0, q.jsx)(q.Fragment, { children: a.map((e, a) => /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i((t || a > 0) && "mt-1 border-t border-border pt-1"),
		children: [e.section && /* @__PURE__ */ (0, q.jsx)("div", {
			className: "px-2 py-1.5 text-xs font-medium text-muted-foreground",
			children: e.section
		}), e.actions.map((e) => /* @__PURE__ */ (0, q.jsx)(ro, {
			action: e,
			submenu: o?.id === e.id ? o : null,
			onOpenSubmenu: s,
			onClose: n,
			size: r
		}, e.id))]
	}, e.section)) });
}
function ro({ action: e, submenu: t, onOpenSubmenu: n, onClose: r, size: a }) {
	let o = !!e.description, s = e.children ?? [], c = s.length > 0, l = (t) => {
		let r = t.getBoundingClientRect(), i = r.right + 224 > window.innerWidth ? r.left - 224 : r.right;
		n({
			id: e.id,
			x: i,
			y: r.top
		});
	};
	return /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [/* @__PURE__ */ (0, q.jsxs)("button", {
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
			e.icon && /* @__PURE__ */ (0, q.jsx)(E, {
				icon: e.icon,
				className: i("shrink-0 text-sm", o && "mt-0.5", e.iconClassName ?? "text-muted-foreground")
			}),
			/* @__PURE__ */ (0, q.jsxs)("span", {
				className: "min-w-0 flex-1",
				children: [/* @__PURE__ */ (0, q.jsx)("span", {
					className: i(o && "font-medium"),
					children: e.label
				}), e.description && /* @__PURE__ */ (0, q.jsx)("span", {
					className: "mt-0.5 block text-xs text-muted-foreground",
					children: e.description
				})]
			}),
			c && /* @__PURE__ */ (0, q.jsx)(E, {
				icon: N,
				className: "shrink-0 text-sm text-muted-foreground"
			})
		]
	}), c && t && /* @__PURE__ */ (0, q.jsx)("div", {
		role: "menu",
		"aria-label": typeof e.label == "string" ? e.label : "Submenu",
		className: "fixed z-50 max-h-[calc(100vh-1rem)] min-w-[14rem] max-w-[calc(100vw-1rem)] overflow-auto rounded-md border border-border bg-popover p-1.5 text-popover-foreground shadow-lg shadow-black/5",
		style: {
			left: t.x,
			top: t.y
		},
		children: s.map((e) => /* @__PURE__ */ (0, q.jsx)(ro, {
			action: e,
			submenu: null,
			onOpenSubmenu: n,
			onClose: r,
			size: a
		}, e.id))
	})] });
}
function io({ densityOverride: e, separated: t, onDensityChange: n, size: r = "menu" }) {
	let a = e ?? "inherit";
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i(t && "mt-1 border-t border-border pt-1"),
		children: [
			/* @__PURE__ */ (0, q.jsx)("div", {
				className: "px-2 py-1.5 text-xs font-medium text-muted-foreground",
				children: "Density"
			}),
			/* @__PURE__ */ (0, q.jsxs)("button", {
				type: "button",
				role: "menuitemradio",
				"aria-checked": a === "inherit",
				className: ao(a === "inherit", r),
				onClick: () => n(void 0),
				children: [
					/* @__PURE__ */ (0, q.jsx)(E, {
						icon: st,
						className: "text-sm text-muted-foreground"
					}),
					/* @__PURE__ */ (0, q.jsx)("span", {
						className: "min-w-0 flex-1 truncate",
						children: "Use page density"
					}),
					a === "inherit" ? /* @__PURE__ */ (0, q.jsx)(E, {
						icon: de,
						className: "text-sm text-foreground"
					}) : /* @__PURE__ */ (0, q.jsx)("span", {
						className: "inline-block h-4 w-4",
						"aria-hidden": !0
					})
				]
			}),
			eo.map((e) => {
				let t = a === e.value;
				return /* @__PURE__ */ (0, q.jsxs)("button", {
					type: "button",
					role: "menuitemradio",
					"aria-checked": t,
					className: ao(t, r),
					onClick: () => n(e.value),
					children: [
						/* @__PURE__ */ (0, q.jsx)(E, {
							icon: e.icon,
							className: "text-sm text-muted-foreground"
						}),
						/* @__PURE__ */ (0, q.jsx)("span", {
							className: "min-w-0 flex-1 truncate",
							children: e.label
						}),
						t ? /* @__PURE__ */ (0, q.jsx)(E, {
							icon: de,
							className: "text-sm text-foreground"
						}) : /* @__PURE__ */ (0, q.jsx)("span", {
							className: "inline-block h-4 w-4",
							"aria-hidden": !0
						})
					]
				}, e.value);
			})
		]
	});
}
function ao(e, t) {
	return i("flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", t === "sheet" && "min-h-10 py-2 text-sm", e && "text-foreground");
}
var oo = [
	{
		value: "system",
		icon: rt,
		label: "Use system theme"
	},
	{
		value: "light",
		icon: D,
		label: "Light"
	},
	{
		value: "dark",
		icon: oe,
		label: "Dark"
	}
];
function so({ value: e, separated: t, onChange: n, size: r = "menu" }) {
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i(t && "mt-1 border-t border-border pt-1"),
		children: [/* @__PURE__ */ (0, q.jsx)("div", {
			className: "px-2 py-1.5 text-xs font-medium text-muted-foreground",
			children: "Theme"
		}), oo.map((t) => {
			let i = e === t.value;
			return /* @__PURE__ */ (0, q.jsxs)("button", {
				type: "button",
				role: "menuitemradio",
				"aria-checked": i,
				className: ao(i, r),
				onClick: () => n(t.value),
				children: [
					/* @__PURE__ */ (0, q.jsx)(E, {
						icon: t.icon,
						className: "text-sm text-muted-foreground"
					}),
					/* @__PURE__ */ (0, q.jsx)("span", {
						className: "min-w-0 flex-1 truncate",
						children: t.label
					}),
					i ? /* @__PURE__ */ (0, q.jsx)(E, {
						icon: de,
						className: "text-sm text-foreground"
					}) : /* @__PURE__ */ (0, q.jsx)("span", {
						className: "inline-block h-4 w-4",
						"aria-hidden": !0
					})
				]
			}, t.value);
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/dataTableMenuPosition.ts
function co(e, t) {
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
function lo(e, t) {
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
var uo = 8;
function fo(e) {
	let { x: t, y: n, width: r, height: i } = e.reference, a = (0, K.useMemo)(() => ({ getBoundingClientRect: () => ({
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
	]), { refs: o, floatingStyles: s } = m({
		strategy: "fixed",
		placement: e.placement,
		transform: !1,
		whileElementsMounted: Ve,
		middleware: [
			Be(6),
			He({ padding: uo }),
			Ue({ padding: uo }),
			We({
				padding: uo,
				apply({ availableHeight: e, elements: t }) {
					t.floating.style.maxHeight = `${Math.max(0, Math.floor(e))}px`;
				}
			})
		]
	});
	return (0, K.useLayoutEffect)(() => {
		o.setPositionReference(a);
	}, [o, a]), {
		ref: o.setFloating,
		style: s
	};
}
// ../../../clicky-ui/packages/ui/src/data/DataTablePreferencesSheet.tsx
function po({ open: e, columns: t, hiddenColumns: n, activeColumnKey: r, actions: a, showColumnVisibilityControl: o, showDensityControl: s, showThemeControl: c, themeMenuValue: l, densityOverride: u, visibleHideableColumnCount: d, onToggle: f, onShowAll: p, onDensityChange: m, onThemeChange: g, onClose: _ }) {
	let v = r ? t.find((e) => e.key === r) : void 0, y = v && Ft(v) && d > 1, b = a.length > 0;
	return /* @__PURE__ */ (0, q.jsxs)(h, {
		open: e,
		onClose: _,
		title: "Table options",
		size: "sm",
		scrollBody: !0,
		children: [
			o && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
				/* @__PURE__ */ (0, q.jsxs)("div", {
					className: "flex items-center justify-between gap-3 px-2 py-1.5 text-xs font-medium text-muted-foreground",
					children: [/* @__PURE__ */ (0, q.jsx)("span", { children: "Columns" }), /* @__PURE__ */ (0, q.jsx)("button", {
						type: "button",
						className: "min-h-10 rounded px-1.5 text-sm text-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none",
						onClick: p,
						children: "Show all"
					})]
				}),
				v && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [/* @__PURE__ */ (0, q.jsxs)("button", {
					type: "button",
					disabled: !y,
					className: i("flex min-h-10 w-full items-center gap-2 rounded-md px-2 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", !y && "cursor-not-allowed opacity-50"),
					onClick: () => {
						y && (f(v), _());
					},
					children: [/* @__PURE__ */ (0, q.jsx)(E, {
						icon: et,
						className: "text-sm text-muted-foreground"
					}), /* @__PURE__ */ (0, q.jsxs)("span", { children: ["Hide ", Pt(v)] })]
				}), /* @__PURE__ */ (0, q.jsx)("div", { className: "my-1 h-px bg-border" })] }),
				/* @__PURE__ */ (0, q.jsx)("div", { children: t.map((e) => {
					let t = Ft(e), r = n[e.key] !== !0, a = !t || r && d <= 1;
					return /* @__PURE__ */ (0, q.jsxs)("label", {
						className: i("flex min-h-10 cursor-pointer items-center gap-2 rounded-md px-3 text-sm transition-colors hover:bg-accent hover:text-accent-foreground", a && "cursor-not-allowed opacity-50"),
						children: [/* @__PURE__ */ (0, q.jsx)("input", {
							type: "checkbox",
							className: "h-5 w-5 rounded border-border",
							checked: r,
							disabled: a,
							onChange: () => f(e)
						}), /* @__PURE__ */ (0, q.jsx)("span", {
							className: "truncate",
							children: Pt(e)
						})]
					}, e.key);
				}) })
			] }),
			s && /* @__PURE__ */ (0, q.jsx)(io, {
				densityOverride: u,
				separated: o,
				onDensityChange: m,
				size: "sheet"
			}),
			b && /* @__PURE__ */ (0, q.jsx)(no, {
				actions: a,
				separated: o || s,
				onClose: _,
				size: "sheet"
			}),
			c && g && /* @__PURE__ */ (0, q.jsx)(so, {
				value: l ?? "system",
				separated: o || s || b,
				onChange: g,
				size: "sheet"
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/SortableHeader.tsx
function mo({ active: e, dir: t, onClick: n, align: r = "left", className: a, children: o }) {
	return /* @__PURE__ */ (0, q.jsxs)("button", {
		type: "button",
		onClick: n,
		className: i("inline-flex items-center gap-1 cursor-pointer select-none hover:text-foreground", e ? "text-primary" : "text-muted-foreground", r === "right" ? "justify-end" : r === "center" ? "justify-center" : "justify-start", a),
		children: [o, e ? /* @__PURE__ */ (0, q.jsx)("span", {
			"aria-hidden": !0,
			children: t === "asc" ? "↑" : "↓"
		}) : /* @__PURE__ */ (0, q.jsx)("span", {
			"aria-hidden": !0,
			className: "opacity-40",
			children: "↕"
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/cells/Timestamp.tsx
function ho({ value: e, format: t, showTitleOnHover: n = !0, className: r }) {
	let a = T(e);
	if (!a) return /* @__PURE__ */ (0, q.jsx)("span", {
		className: i("text-muted-foreground", r),
		children: "—"
	});
	let o = te(a, t), s = n ? a.toISOString() : void 0;
	return /* @__PURE__ */ (0, q.jsx)("span", {
		className: i("tabular-nums whitespace-nowrap text-muted-foreground", r),
		title: s,
		children: o
	});
}
// ../../../clicky-ui/packages/ui/src/components/IconButton.tsx
var go = (0, K.forwardRef)(({ icon: e, label: t, iconClassName: n, className: r, type: a = "button", ...o }, s) => /* @__PURE__ */ (0, q.jsx)("button", {
	ref: s,
	type: a,
	"aria-label": t,
	title: t,
	className: i("inline-flex shrink-0 items-center justify-center rounded-sm border-0 bg-transparent p-0 text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50", r),
	...o,
	children: /* @__PURE__ */ (0, q.jsx)(E, {
		icon: e,
		className: i("text-sm", n)
	})
}));
go.displayName = "IconButton";
// ../../../clicky-ui/packages/ui/src/data/properties-utils.ts
function _o(e) {
	return e.replace(/[_-]+/g, " ").replace(/\.+/g, " ").replace(/\b\w/g, (e) => e.toUpperCase());
}
// ../../../clicky-ui/packages/ui/src/data/properties-context.ts
var vo = (0, K.createContext)(null), yo = "en-US";
function bo(e, t = 1) {
	return new Intl.NumberFormat(yo, { maximumFractionDigits: t }).format(Number.isFinite(e) ? e : 0);
}
function xo(e) {
	return e.endsWith(".0") ? e.slice(0, -2) : e;
}
var So = [
	"B",
	"KB",
	"MB",
	"GB",
	"TB",
	"PB"
];
function Co(e, t) {
	let n = t?.suffix ?? "";
	if (e === void 0 || !Number.isFinite(e) || e <= 0) return `0 B${n}`;
	let r = t?.base ?? 1024, i = e, a = 0;
	for (; i >= r && a < So.length - 1;) i /= r, a++;
	return `${i >= 10 || a === 0 ? i.toFixed(0) : xo(i.toFixed(1))} ${So[a]}${n}`;
}
function wo(e) {
	return Co(e, { suffix: "/s" });
}
var To = [
	"",
	"K",
	"M",
	"B",
	"T"
];
function Eo(e) {
	if (e === void 0 || !Number.isFinite(e)) return "0";
	let t = e < 0 ? "-" : "", n = Math.abs(e);
	if (n < 1e3) return `${t}${bo(n)}`;
	let r = 0;
	for (; n >= 1e3 && r < To.length - 1;) n /= 1e3, r++;
	return `${t}${n >= 10 ? n.toFixed(0) : xo(n.toFixed(1))}${To[r]}`;
}
function Do(e, t) {
	if (e === void 0 || !Number.isFinite(e)) return "0 ms";
	let n = t?.from === "s" ? e * 1e3 : e, r = Math.abs(n);
	return r < 1e3 ? `${bo(n, 0)} ms` : r < 6e4 ? `${bo(n / 1e3)} s` : r < 36e5 ? `${bo(n / 6e4)} min` : `${bo(n / 36e5)} h`;
}
var Oo = {
	none: (e) => Eo(e),
	short: (e) => Eo(e),
	percent: (e) => `${bo(e)}%`,
	percentunit: (e) => `${bo(e * 100)}%`,
	"%": (e) => `${bo(e)}%`,
	bytes: (e) => Co(e, { base: 1024 }),
	decbytes: (e) => Co(e, { base: 1e3 }),
	Bps: (e) => wo(e),
	binBps: (e) => wo(e),
	ms: (e) => Do(e, { from: "ms" }),
	s: (e) => Do(e, { from: "s" })
};
function ko(e, t) {
	let n = Number.isFinite(e) ? e : 0;
	if (!t) return bo(n);
	let r = Oo[t];
	return r ? r(n) : `${bo(n)}${t}`;
}
// ../../../clicky-ui/packages/ui/src/data/PropertyValue.tsx
function Ao({ item: e, density: t }) {
	let { value: n, type: r, unit: i } = e;
	if (n == null) return /* @__PURE__ */ (0, q.jsx)("span", {
		className: "font-mono italic text-muted-foreground",
		children: "null"
	});
	if (r === "json" || r === void 0 && typeof n == "object") return typeof n == "string" ? /* @__PURE__ */ (0, q.jsx)(v, {
		source: n,
		format: "json",
		density: t
	}) : /* @__PURE__ */ (0, q.jsx)(v, {
		data: n,
		format: "json",
		density: t
	});
	if (r === "bytes" || r === "duration" || r === "number" || i !== void 0) {
		if (typeof n != "number" || !Number.isFinite(n)) throw Error(`Property ${e.key}: expected a finite numeric value for ${r ?? i}`);
		let t = i ?? (r === "bytes" ? "bytes" : r === "duration" ? "ms" : void 0), a = t === "bytes" && n < 0 ? `-${ko(Math.abs(n), t)}` : ko(n, t);
		return /* @__PURE__ */ (0, q.jsx)("span", {
			className: "font-mono",
			title: String(n),
			children: a
		});
	}
	if (r === "date") {
		if (typeof n != "number" && typeof n != "string" && !(n instanceof Date)) throw Error(`Property ${e.key}: expected a date or timestamp`);
		return /* @__PURE__ */ (0, q.jsx)("time", {
			dateTime: new Date(n).toISOString(),
			title: String(n),
			children: new Date(n).toISOString()
		});
	}
	if (r === "boolean" && typeof n != "boolean") throw Error(`Property ${e.key}: expected a boolean`);
	return typeof n == "boolean" ? /* @__PURE__ */ (0, q.jsx)("input", {
		type: "checkbox",
		checked: n,
		disabled: !0,
		"aria-label": `${e.key} value`,
		title: String(n),
		className: "h-4 w-4 accent-primary"
	}) : /* @__PURE__ */ (0, q.jsx)("span", {
		className: "whitespace-pre-wrap break-words font-mono",
		children: String(n)
	});
}
// ../../../clicky-ui/packages/ui/src/data/Properties.tsx
var jo = "clicky-ui-properties", Mo = [
	{
		value: "compact",
		label: "Compact",
		icon: d
	},
	{
		value: "comfortable",
		label: "Comfortable",
		icon: ht
	},
	{
		value: "spacious",
		label: "Spacious",
		icon: ct
	}
], No = {
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
function Po({ items: e, renderLabel: t, renderValue: n, labelIcon: r, prefixActions: a, suffixActions: o, className: s, rowClassName: c, gridTemplateColumns: l, labelClassName: u, valueClassName: d, density: f = "comfortable", keyRef: p, showDensityMenu: m = !0, emptyMessage: h = "No properties" }) {
	let g = (0, K.useContext)(vo), [_, v] = (0, K.useState)();
	(0, K.useEffect)(() => v(Lo(p)), [p]);
	let y = e.filter((e) => !e.hidden);
	if (y.length === 0) return /* @__PURE__ */ (0, q.jsx)("div", {
		className: "text-sm text-muted-foreground",
		children: h
	});
	let b = g?.density ?? _ ?? f, x = No[b], S = g?.columns ?? l ?? x.gridTemplateColumns, C = g?.depth ?? 0;
	return /* @__PURE__ */ (0, q.jsxs)("dl", {
		className: i("grid gap-x-density-3 divide-y divide-border rounded-md border border-border bg-muted/20", g && "col-span-full rounded-none border-0 bg-transparent", s),
		style: { gridTemplateColumns: g ? "subgrid" : S },
		children: [m && !g ? /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "col-span-full flex items-center justify-end px-density-2 py-density-1",
			children: [/* @__PURE__ */ (0, q.jsx)("dt", {
				className: "sr-only",
				children: "Properties display options"
			}), /* @__PURE__ */ (0, q.jsx)("dd", { children: /* @__PURE__ */ (0, q.jsx)(Fo, {
				density: b,
				onChange: (e) => {
					v(e), Ro(p, e);
				}
			}) })]
		}) : null, y.map((e) => {
			let s = typeof r == "function" ? r(e.key, e.value, e) : r, l = t ? t(e.key, e.value, e) : _o(e.key), f = typeof l == "string" ? l : e.key;
			return /* @__PURE__ */ (0, q.jsxs)(K.Fragment, { children: [/* @__PURE__ */ (0, q.jsxs)("div", {
				className: i("col-span-full grid min-w-0 grid-cols-subgrid gap-density-3", x.row, c),
				children: [/* @__PURE__ */ (0, q.jsxs)("dt", {
					"aria-label": f,
					className: i("min-w-0 truncate", x.label, u),
					style: { paddingInlineStart: C * 16 },
					children: [/* @__PURE__ */ (0, q.jsxs)("div", {
						className: "flex min-w-0 items-center gap-density-1",
						children: [s ? /* @__PURE__ */ (0, q.jsx)(E, {
							name: s,
							className: "text-xs shrink-0"
						}) : null, /* @__PURE__ */ (0, q.jsx)("span", {
							className: "min-w-0 truncate",
							children: l
						})]
					}), e.subtitle ? /* @__PURE__ */ (0, q.jsx)("div", {
						className: "mt-0.5 text-[10px] font-normal text-muted-foreground/80",
						children: e.subtitle
					}) : null]
				}), /* @__PURE__ */ (0, q.jsx)("dd", {
					className: i("min-w-0 space-y-density-1", d),
					children: /* @__PURE__ */ (0, q.jsxs)("div", {
						className: "flex min-w-0 items-start gap-density-1",
						children: [
							/* @__PURE__ */ (0, q.jsx)(Vo, {
								actions: a,
								item: e,
								className: "shrink-0 pt-0.5"
							}),
							/* @__PURE__ */ (0, q.jsx)("div", {
								className: i("min-w-0 max-w-full flex-1", x.value),
								children: n ? n(e.key, e.value, e) : /* @__PURE__ */ (0, q.jsx)(Ao, {
									item: e,
									density: b
								})
							}),
							/* @__PURE__ */ (0, q.jsx)(Vo, {
								actions: o,
								item: e,
								className: "shrink-0 pt-0.5"
							})
						]
					})
				})]
			}), e.expandable && e.expanded && e.renderChildren ? /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "col-span-full grid grid-cols-subgrid",
				children: [/* @__PURE__ */ (0, q.jsxs)("dt", {
					className: "sr-only",
					children: [f, " properties"]
				}), /* @__PURE__ */ (0, q.jsx)("dd", {
					className: "col-span-full grid min-w-0 grid-cols-subgrid",
					children: /* @__PURE__ */ (0, q.jsx)(vo.Provider, {
						value: {
							depth: C + 1,
							columns: S,
							density: b
						},
						children: e.renderChildren()
					})
				})]
			}) : null] }, e.key);
		})]
	});
}
function Fo({ density: e, onChange: t }) {
	return /* @__PURE__ */ (0, q.jsx)(fe, {
		align: "right",
		menuLabel: "Properties options",
		trigger: /* @__PURE__ */ (0, q.jsx)(go, {
			icon: le,
			label: "Properties options",
			className: "h-6 w-6"
		}),
		children: (n) => /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "min-w-44 py-1 text-popover-foreground",
			children: [/* @__PURE__ */ (0, q.jsx)("div", {
				className: "px-2 py-1 text-xs font-medium text-muted-foreground",
				children: "Density"
			}), Mo.map((r) => {
				let i = e === r.value;
				return /* @__PURE__ */ (0, q.jsxs)("button", {
					type: "button",
					role: "menuitemradio",
					"aria-checked": i,
					className: Io(i),
					onClick: () => {
						t(r.value), n();
					},
					children: [
						/* @__PURE__ */ (0, q.jsx)(E, {
							icon: r.icon,
							className: "text-sm text-muted-foreground"
						}),
						/* @__PURE__ */ (0, q.jsx)("span", {
							className: "min-w-0 flex-1 truncate",
							children: r.label
						}),
						i ? /* @__PURE__ */ (0, q.jsx)(E, {
							icon: de,
							className: "text-sm text-foreground"
						}) : /* @__PURE__ */ (0, q.jsx)("span", {
							className: "inline-block h-4 w-4",
							"aria-hidden": !0
						})
					]
				}, r.value);
			})]
		})
	});
}
function Io(e) {
	return i("flex w-full items-center gap-2 px-2 py-1.5 text-left text-sm", "hover:bg-accent focus:bg-accent focus:outline-none", e && "bg-accent/60");
}
function Lo(e) {
	let t = zo(e);
	if (t && typeof window < "u") try {
		let e = window.localStorage.getItem(t);
		return Bo(e) ? e : void 0;
	} catch {
		return;
	}
}
function Ro(e, t) {
	let n = zo(e);
	if (n && typeof window < "u") try {
		window.localStorage.setItem(n, t);
	} catch {}
}
function zo(e) {
	let t = e?.trim();
	return t ? `${jo}:${t}:density` : void 0;
}
function Bo(e) {
	return e === "compact" || e === "comfortable" || e === "spacious";
}
function Vo({ actions: e, item: t, className: n }) {
	if (!e || e.length === 0) return null;
	let r = e.filter((e) => !e.visible || e.visible(t.key, t.value, t));
	return r.length === 0 ? null : /* @__PURE__ */ (0, q.jsx)("span", {
		className: i("inline-flex items-center", n),
		children: r.map((e) => /* @__PURE__ */ (0, q.jsx)(Ho, {
			action: e,
			item: t
		}, e.id))
	});
}
function Ho({ action: e, item: t }) {
	let n = e.label(t.key, t.value, t), r = e.disabled?.(t.key, t.value, t) ?? !1;
	return /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		"aria-label": n,
		title: n,
		disabled: r,
		className: i("inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground", "hover:bg-accent hover:text-foreground", "disabled:opacity-35 disabled:hover:bg-transparent disabled:hover:text-muted-foreground"),
		onClick: (n) => {
			n.stopPropagation(), e.onClick(t.key, t.value, t);
		},
		children: /* @__PURE__ */ (0, q.jsx)(E, {
			...typeof e.icon == "string" ? { name: e.icon } : { icon: e.icon },
			className: "text-xs"
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/cells/tag-utils.ts
function Uo(e, t = "=") {
	return e == null ? [] : Array.isArray(e) ? e.map((e) => Wo(e, t)).filter((e) => e.value !== "") : typeof e == "object" ? Object.entries(e).map(([e, n]) => {
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
function Wo(e, t) {
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
function Go(e, t = "=") {
	return Uo(e, t).map((e) => e.token);
}
function Ko(e, t = "=") {
	let n = e.indexOf(t);
	return n <= 0 ? {
		key: "",
		value: e
	} : {
		key: e.slice(0, n),
		value: e.slice(n + t.length)
	};
}
var qo = (0, K.createContext)({
	getMode: () => "neutral",
	toggleInclude: () => {},
	toggleExclude: () => {}
});
function Jo() {
	return (0, K.useContext)(qo);
}
function Yo(e, t) {
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
// ../../../clicky-ui/packages/ui/src/data/cells/TagList.tsx
function Xo({ value: e, children: t }) {
	return /* @__PURE__ */ (0, q.jsx)(qo.Provider, {
		value: e,
		children: t
	});
}
function Zo(e) {
	typeof navigator < "u" && navigator.clipboard?.writeText && navigator.clipboard.writeText(e).catch(() => void 0);
}
function Qo({ tag: e, actions: t, compact: n = !1 }) {
	let r = Jo(), a = r.getMode(e.token), o = (0, K.useCallback)((t) => {
		t.stopPropagation(), r.toggleInclude(e.token);
	}, [r, e.token]), s = (0, K.useCallback)((t) => {
		t.stopPropagation(), r.toggleExclude(e.token);
	}, [r, e.token]), c = (0, K.useCallback)((t) => {
		t.stopPropagation(), Zo(e.value);
	}, [e.value]), l = a === "include", u = a === "exclude", d = n && !!e.key, f = /* @__PURE__ */ (0, q.jsx)("span", {
		...d ? { title: e.display } : {},
		className: i("inline-flex items-center rounded-md border bg-background align-middle", l ? "border-green-500/50 bg-green-500/5" : u ? "border-red-500/50 bg-red-500/5" : "border-border"),
		children: e.key && !n ? /* @__PURE__ */ (0, q.jsx)(L, {
			size: "xs",
			variant: "label",
			tone: "neutral",
			label: e.key,
			value: e.value,
			maxWidth: 32,
			truncate: "auto",
			clickToCopy: !1,
			className: "border-transparent shadow-none"
		}) : /* @__PURE__ */ (0, q.jsx)(L, {
			size: "xs",
			variant: "soft",
			tone: "neutral",
			maxWidth: 32,
			truncate: "auto",
			clickToCopy: !1,
			children: e.value
		})
	}), p = /* @__PURE__ */ (0, q.jsxs)("span", {
		className: "inline-flex items-center gap-0.5",
		children: [
			/* @__PURE__ */ (0, q.jsx)($o, {
				ariaLabel: `Include ${e.display}`,
				onClick: o,
				active: l,
				activeClassName: "bg-green-500/20 text-green-700 dark:text-green-400",
				icon: tt
			}),
			/* @__PURE__ */ (0, q.jsx)($o, {
				ariaLabel: `Exclude ${e.display}`,
				onClick: s,
				active: u,
				activeClassName: "bg-red-500/20 text-red-700 dark:text-red-400",
				icon: nt
			}),
			/* @__PURE__ */ (0, q.jsx)($o, {
				ariaLabel: `Copy ${e.display}`,
				onClick: c,
				icon: w
			})
		]
	});
	return t === "inline" ? /* @__PURE__ */ (0, q.jsxs)("span", {
		className: "inline-flex items-center gap-0.5 rounded-md border border-border bg-background pr-0.5",
		children: [f, p]
	}) : /* @__PURE__ */ (0, q.jsx)(M, {
		placement: "top",
		delay: 120,
		arrow: !1,
		trigger: f,
		cardClassName: "!p-1",
		children: p
	});
}
function $o({ ariaLabel: e, onClick: t, active: n, activeClassName: r, icon: a }) {
	return /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		"aria-label": e,
		title: e,
		onClick: t,
		className: i("inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground", "hover:bg-accent hover:text-foreground", n && r),
		children: /* @__PURE__ */ (0, q.jsx)(E, {
			icon: a,
			className: "text-xs"
		})
	});
}
function es({ tags: e, maxVisible: t = 3, className: n, actions: r = "hover", compact: a, wrap: o = !1 }) {
	let s = k(), c = a ?? s === "compact";
	if (e.length === 0) return /* @__PURE__ */ (0, q.jsx)("span", {
		className: "text-muted-foreground",
		children: "—"
	});
	if (o) return /* @__PURE__ */ (0, q.jsx)("span", {
		className: i("flex min-w-0 flex-wrap items-center gap-1", n),
		children: e.map((e, t) => /* @__PURE__ */ (0, q.jsx)(Qo, {
			tag: e,
			actions: r,
			compact: c
		}, `${e.display}-${t}`))
	});
	let l = e.slice(0, t), u = e.slice(t);
	return /* @__PURE__ */ (0, q.jsxs)("span", {
		className: i("flex min-w-0 items-center gap-1", n),
		children: [/* @__PURE__ */ (0, q.jsx)("span", {
			className: "flex min-w-0 flex-1 flex-nowrap items-center gap-1 overflow-hidden",
			children: l.map((e, t) => /* @__PURE__ */ (0, q.jsx)(Qo, {
				tag: e,
				actions: r,
				compact: c
			}, `${e.display}-${t}`))
		}), u.length > 0 && /* @__PURE__ */ (0, q.jsx)(M, {
			placement: "top",
			trigger: /* @__PURE__ */ (0, q.jsxs)(L, {
				size: "xs",
				variant: "outlined",
				tone: "neutral",
				clickToCopy: !1,
				className: "shrink-0",
				children: ["+", u.length]
			}),
			cardClassName: "min-w-72 max-w-[90vw] whitespace-normal !p-0",
			children: /* @__PURE__ */ (0, q.jsx)(ts, { tags: u })
		})]
	});
}
function ts({ tags: e, rowClassName: t, gridTemplateColumns: n }) {
	let r = Jo(), i = (0, K.useMemo)(() => e.map((e, t) => ({
		key: `${e.token}-${t}`,
		value: e
	})), [e]), a = e.some((e) => !!e.key), o = n ?? (a ? "minmax(4rem, 8rem) minmax(0, 1fr)" : "0 minmax(0, 1fr)");
	return /* @__PURE__ */ (0, q.jsx)(Po, {
		items: i,
		density: "compact",
		className: "border-0 bg-transparent",
		...t ? { rowClassName: t } : {},
		gridTemplateColumns: o,
		renderLabel: (e, t) => t.key ? /* @__PURE__ */ (0, q.jsx)("span", {
			className: "font-mono",
			children: t.key
		}) : null,
		renderValue: (e, t) => /* @__PURE__ */ (0, q.jsx)("span", {
			className: "block truncate font-mono",
			title: t.value,
			children: t.value
		}),
		suffixActions: [
			{
				id: "include",
				icon: tt,
				label: (e, t) => `Include ${t.display}`,
				onClick: (e, t) => r.toggleInclude(t.token)
			},
			{
				id: "exclude",
				icon: nt,
				label: (e, t) => `Exclude ${t.display}`,
				onClick: (e, t) => r.toggleExclude(t.token)
			},
			{
				id: "copy",
				icon: w,
				label: (e, t) => `Copy ${t.display}`,
				onClick: (e, t) => Zo(t.value)
			}
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/cells/CellFilterActions.tsx
function ns({ children: e, value: t, displayValue: n = t, mode: r, onChange: i }) {
	return /* @__PURE__ */ (0, q.jsx)(M, {
		placement: "top",
		delay: 120,
		arrow: !1,
		trigger: /* @__PURE__ */ (0, q.jsx)("span", {
			className: "relative inline-flex min-w-0",
			children: e
		}),
		cardClassName: "!p-1",
		children: /* @__PURE__ */ (0, q.jsxs)("span", {
			className: "inline-flex items-center gap-0.5",
			children: [/* @__PURE__ */ (0, q.jsx)(rs, {
				label: `Include ${n}`,
				icon: tt,
				active: r === "include",
				activeClassName: "bg-green-500/20 text-green-700 dark:text-green-400",
				onClick: () => i(r === "include" ? void 0 : "include")
			}), /* @__PURE__ */ (0, q.jsx)(rs, {
				label: `Exclude ${n}`,
				icon: nt,
				active: r === "exclude",
				activeClassName: "bg-red-500/20 text-red-700 dark:text-red-400",
				onClick: () => i(r === "exclude" ? void 0 : "exclude")
			})]
		})
	});
}
function rs({ label: e, icon: t, active: n, activeClassName: r, onClick: a }) {
	return /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		"aria-label": e,
		"aria-pressed": n,
		title: e,
		onClick: (e) => {
			e.stopPropagation(), a();
		},
		className: i("inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground", "hover:bg-accent hover:text-foreground", n && r),
		children: /* @__PURE__ */ (0, q.jsx)(E, {
			icon: t,
			className: "text-xs"
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/cells/StatusDot.tsx
var is = {
	success: "bg-emerald-500",
	error: "bg-rose-500",
	warning: "bg-amber-500",
	info: "bg-sky-500"
}, as = {
	xs: "h-1.5 w-1.5",
	sm: "h-2 w-2",
	md: "h-2.5 w-2.5"
};
function os({ status: e, size: t = "sm", label: n, title: r, className: a }) {
	let o = r ?? n ?? e;
	return /* @__PURE__ */ (0, q.jsxs)("span", {
		className: i("inline-flex items-center gap-1.5 align-middle", a),
		title: o,
		children: [/* @__PURE__ */ (0, q.jsx)("span", {
			role: "img",
			"aria-label": n ?? e,
			className: i("inline-block shrink-0 rounded-full", is[e], as[t])
		}), n && /* @__PURE__ */ (0, q.jsx)("span", {
			className: "truncate text-xs",
			children: n
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/cells/status-mapping.ts
var ss = /* @__PURE__ */ new Set([
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
]), cs = /* @__PURE__ */ new Set([
	"warn",
	"warning",
	"degraded",
	"slow",
	"stale"
]), ls = /* @__PURE__ */ new Set([
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
function us(e) {
	if (e == null) return null;
	if (typeof e == "boolean") return e ? "success" : "error";
	let t = String(e).toLowerCase().trim();
	return t ? ss.has(t) ? "error" : cs.has(t) ? "warning" : ls.has(t) ? "success" : null : null;
}
// ../../../clicky-ui/packages/ui/src/data/DataTable.grouping.ts
var ds = "\0empty";
function fs(e, t) {
	return e === "\0empty" ? t ?? "(empty)" : e;
}
function ps(e) {
	if (e == null || e === "") return ds;
	switch (typeof e) {
		case "string":
		case "number":
		case "boolean":
		case "bigint": return String(e);
		default: throw Error(`DataTable automatic grouping requires scalar values; received ${Array.isArray(e) ? "array" : typeof e}`);
	}
}
function ms(e, t) {
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
function hs(e, t) {
	return typeof t == "function" ? t(e.key, e.rows) : t === !0;
}
// ../../../clicky-ui/packages/ui/src/data/DataTableGroupingControls.tsx
function gs({ modes: e, value: t, hasGroups: n, allExpanded: r, allCollapsed: i, onModeChange: a, onExpandAll: o, onCollapseAll: s }) {
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "flex shrink-0 items-center gap-1",
		role: "group",
		"aria-label": "Grouping controls",
		children: [
			/* @__PURE__ */ (0, q.jsx)("div", {
				className: "w-44 shrink-0",
				children: /* @__PURE__ */ (0, q.jsx)(Kn, {
					"aria-label": "Group rows by",
					value: t,
					options: e,
					onChange: (e) => a(e.target.value)
				})
			}),
			/* @__PURE__ */ (0, q.jsx)(_s, {
				label: "Expand all groups",
				pressed: r,
				disabled: !n,
				onClick: o,
				icon: De
			}),
			/* @__PURE__ */ (0, q.jsx)(_s, {
				label: "Collapse all groups",
				pressed: i,
				disabled: !n,
				onClick: s,
				icon: Ee
			})
		]
	});
}
function _s({ label: e, pressed: t, disabled: n, onClick: r, icon: a }) {
	return /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		"aria-label": e,
		title: e,
		"aria-pressed": t,
		disabled: n,
		onClick: r,
		className: i("inline-flex h-control-h w-control-h items-center justify-center rounded-md border border-border bg-background text-muted-foreground hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50", t && "bg-accent text-accent-foreground"),
		children: /* @__PURE__ */ (0, q.jsx)(E, {
			icon: a,
			className: "text-sm"
		})
	});
}
function vs(e) {
	return e.display ? e.display : e.menu ? "menu" : e.primary === !1 || e.section !== void 0 || e.children?.length ? "overflow" : "button";
}
function ys(e, t) {
	let n = [], r = [];
	for (let t of e) {
		let e = vs(t);
		e === "menu" ? n.push(t) : e === "button" && r.push(t);
	}
	let i = r.filter((e) => e.primary), a = i.length ? i : r.slice(0, t?.maxButtons ?? 3), o = new Set([...n, ...a].map((e) => e.id));
	return {
		menus: n,
		buttons: a,
		overflow: e.filter((e) => !o.has(e.id))
	};
}
// ../../../clicky-ui/packages/ui/src/data/SelectionActionBar.tsx
function bs({ count: e, onClear: t }) {
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		role: "status",
		"data-testid": "data-table-selection-count",
		className: "flex flex-wrap items-center gap-density-2 text-xs text-muted-foreground",
		children: [/* @__PURE__ */ (0, q.jsxs)("span", {
			className: "font-medium text-foreground",
			children: [e.toLocaleString(), " selected"]
		}), /* @__PURE__ */ (0, q.jsx)(a, {
			type: "button",
			variant: "outline",
			size: "sm",
			onClick: t,
			children: "Clear selection"
		})]
	});
}
var xs = "Cancel";
function Ss({ action: e, context: t, onCancel: n, onConfirm: r }) {
	let i = typeof e.confirm == "object" ? e.confirm : {}, o = t.selectedScope?.total ?? t.selectedRowIds.length, s = typeof i.message == "function" ? i.message(t) : i.message ?? `This will apply to ${o.toLocaleString()} selected ${o === 1 ? "row" : "rows"}.`, c = i.confirmLabel ?? (typeof e.label == "string" ? e.label : "Confirm");
	return /* @__PURE__ */ (0, q.jsx)(h, {
		open: !0,
		size: "sm",
		expandable: !1,
		onClose: n,
		title: i.title ?? e.label,
		footer: /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "flex justify-end gap-density-2",
			children: [/* @__PURE__ */ (0, q.jsx)(a, {
				type: "button",
				variant: "ghost",
				size: "sm",
				onClick: n,
				children: i.cancelLabel ?? xs
			}), /* @__PURE__ */ (0, q.jsx)(a, {
				type: "button",
				size: "sm",
				variant: e.variant === "destructive" ? "destructive" : "default",
				onClick: r,
				children: c
			})]
		}),
		children: /* @__PURE__ */ (0, q.jsx)("p", {
			className: "text-sm text-muted-foreground",
			children: s
		})
	});
}
function Cs({ actions: e, context: t, showCount: n = !0, maxButtons: r, className: i }) {
	let [o, s] = (0, K.useState)(null), [c, l] = (0, K.useState)(null), u = (0, K.useRef)(!1), d = (0, K.useCallback)((e) => {
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
	}, [t]), f = (0, K.useCallback)((e) => {
		e.confirm ? l(e) : d(e);
	}, [d]), { menus: p, buttons: m, overflow: h } = ys(e, { maxButtons: r }), g = [...p, ...m], _ = o !== null, v = h.find((e) => e.id === o), y = (e) => ({
		label: e.label,
		onSelect: () => f(e),
		disabled: e.disabled || _,
		...e.icon ? { icon: e.icon } : {},
		...e.section === void 0 ? {} : { group: e.section },
		...e.children?.length ? { children: e.children.map(y) } : {}
	}), b = /* @__PURE__ */ new Map();
	for (let e of h) {
		let t = e.section ?? "", n = b.get(t);
		n ? n.push(e) : b.set(t, [e]);
	}
	let x = [...b.values()].flat().map(y);
	return /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
		n ? /* @__PURE__ */ (0, q.jsx)(bs, {
			count: t.selectedScope?.total ?? t.selectedRowIds.length,
			onClear: t.clearSelection
		}) : null,
		/* @__PURE__ */ (0, q.jsxs)("span", {
			className: i,
			children: [g.map((e) => {
				let n = o === e.id, r = p.includes(e), i = /* @__PURE__ */ (0, q.jsxs)(a, {
					type: "button",
					size: "sm",
					variant: e.variant ?? "outline",
					disabled: e.disabled || _ && !n,
					loading: n,
					...r ? { "aria-haspopup": "menu" } : { onClick: () => f(e) },
					children: [
						e.icon && !n ? /* @__PURE__ */ (0, q.jsx)(E, {
							icon: e.icon,
							...e.iconClassName ? { className: e.iconClassName } : {}
						}) : null,
						n && e.pendingLabel !== void 0 ? e.pendingLabel : e.label,
						r ? /* @__PURE__ */ (0, q.jsx)(E, { icon: F }) : null
					]
				});
				if (!r) return /* @__PURE__ */ (0, q.jsx)("span", { children: i }, e.id);
				let s = e.menu ? { children: (n) => e.menu({
					context: t,
					close: n,
					busy: _,
					run: (t) => f({
						...e,
						...t.pendingLabel === void 0 ? {} : { pendingLabel: t.pendingLabel },
						...t.confirm === void 0 ? {} : { confirm: t.confirm },
						onSelect: t.onSelect
					})
				}) } : { items: (e.children ?? []).map(y) };
				return /* @__PURE__ */ (0, q.jsx)(fe, {
					align: "left",
					menuLabel: typeof e.label == "string" ? e.label : e.id,
					trigger: i,
					...s
				}, e.id);
			}), h.length > 0 ? /* @__PURE__ */ (0, q.jsx)(fe, {
				align: "right",
				menuLabel: "Selection actions",
				items: x,
				trigger: /* @__PURE__ */ (0, q.jsxs)(a, {
					type: "button",
					size: "sm",
					variant: "outline",
					"aria-haspopup": "menu",
					"aria-label": "More selection actions",
					disabled: _ && !v,
					loading: !!v,
					children: [
						v ? null : /* @__PURE__ */ (0, q.jsx)(E, { icon: Xe }),
						"More",
						/* @__PURE__ */ (0, q.jsx)(E, { icon: F })
					]
				})
			}) : null]
		}),
		c ? /* @__PURE__ */ (0, q.jsx)(Ss, {
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
// ../../../clicky-ui/packages/ui/src/data/DataTableOverflowCue.tsx
var ws = {
	overflow: !1,
	atStart: !0,
	atEnd: !0,
	interacted: !1
};
function Ts({ scroll: e }) {
	let [t, n] = (0, K.useState)(ws);
	return (0, K.useEffect)(() => {
		if (!e) {
			n(ws);
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
	}, [e]), t.overflow ? /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
		t.atStart ? null : /* @__PURE__ */ (0, q.jsx)("span", {
			"aria-hidden": !0,
			"data-testid": "data-table-overflow-left",
			className: "pointer-events-none absolute inset-y-0 left-0 z-20 w-6 bg-gradient-to-r from-background to-transparent"
		}),
		t.atEnd ? null : /* @__PURE__ */ (0, q.jsx)("span", {
			"aria-hidden": !0,
			"data-testid": "data-table-overflow-right",
			className: "pointer-events-none absolute inset-y-0 right-0 z-20 w-8 bg-gradient-to-l from-background to-transparent"
		}),
		t.atStart && !t.interacted ? /* @__PURE__ */ (0, q.jsxs)("span", {
			className: i("pointer-events-none absolute right-2 top-10 z-30 inline-flex items-center gap-1 rounded-full", "border border-border bg-background/95 px-2 py-1 text-[10px] font-medium text-muted-foreground shadow-sm"),
			children: ["More columns", /* @__PURE__ */ (0, q.jsx)(dt, {
				"aria-hidden": !0,
				className: "size-3"
			})]
		}) : null
	] }) : null;
}
// ../../../clicky-ui/packages/ui/src/data/DataTable.tsx
var Es = 64, Ds = 160, Os = 224, ks = 96, As = 288, js = 256, Ms = "clicky-ui-data-table-column-widths", Ns = "clicky-ui-data-table-column-visibility", Ps = "clicky-ui-data-table-density", Fs = "px-2.5 py-1.5 density-compact:px-2 density-compact:py-1 density-comfortable:px-2.5 density-comfortable:py-1.5 density-spacious:px-4 density-spacious:py-3", Is = "px-2.5 py-1.5 density-compact:px-2 density-compact:py-0.5 density-comfortable:px-2.5 density-comfortable:py-1.5 density-spacious:px-4 density-spacious:py-3", Ls = [
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
function Rs(e, t) {
	return e ? e.cursor ? e.cursor.current === void 0 && e.cursor.next === void 0 : e.page === 0 ? e.hasMore == null ? e.total != null && e.totalRelation !== "gte" ? e.total <= e.pageSize : t < e.pageSize : !e.hasMore : !1 : !0;
}
function zs({ grouping: e, groupingModes: t, groupingMode: n, defaultGroupingMode: r, onGroupingModeChange: i }) {
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
function Bs({ data: e, error: t, columns: n, loading: r = !1, loadingMessage: a = "Loading results…", loadingRowCount: o = 8, emptyMessage: s = "No data", virtualize: c, className: u, scrollContainerClassName: d, autoFilter: f = !1, boundedFilters: p, onBoundedFiltersChange: m, showGlobalFilter: g, manualFilter: _ = !1, globalFilter: v, onGlobalFilterChange: b, globalFilterPlaceholder: S = "Search all columns…", defaultSort: C, sort: w, onSortChange: E, manualSort: D = !1, filterBarProps: O, externalSearch: k, externalTimeRange: A, externalFilters: j, cellFilters: M, onCellFilterChange: N, pagination: P, infinite: F, scrollContainerRef: L, getRowId: te, rowSelection: R, selectionActions: re, selectionBar: B = "toolbar", getRowClassName: V, footer: ie, grouping: ae, groupingModes: oe, groupingMode: se, defaultGroupingMode: ce, onGroupingModeChange: le, onRowClick: ue, isRowClickable: de, getRowHref: fe, renderExpandedRow: H, expandedRowIds: U, onExpandedRowIdsChange: me, renderCard: W, detailStyle: he = "row", detailDialogTitle: ge, detailDialogSize: _e = "lg", resizableColumns: ve = !0, persistColumnWidths: ye = !0, columnResizeStorageKey: be, hideableColumns: xe = !0, persistColumnVisibility: Se = !0, columnVisibilityStorageKey: Ce, defaultHiddenColumns: we, density: Te, defaultDensity: Ee, onDensityChange: De, persistDensity: Oe = !0, densityStorageKey: ke, showDensityControl: Ae, themeMenuValue: je, onThemeChange: Me, showThemeControl: Ne = !1, showHeaderFilters: Pe = !0, menuActions: Fe }) {
	zs({
		grouping: ae,
		groupingModes: oe,
		groupingMode: se,
		defaultGroupingMode: ce,
		onGroupingModeChange: le
	});
	let G = (0, K.useMemo)(() => n.map((e) => typeof e == "string" ? {
		key: e,
		label: e
	} : e), [n]), Ie = (0, K.useMemo)(() => G.map((e) => e.key).join("|"), [G]), Le = be ?? `${Ms}:${Ie}`, Re = Ce ?? `${Ns}:${Ie}`, ze = ke ?? `${Ps}:${Ie}`, [Be, Ve] = (0, K.useState)(() => ye ? bc(Le, G) : {}), He = (we ?? []).join("|"), Ue = (0, K.useMemo)(() => new Set(we ?? []), [He]), [We, Ge] = (0, K.useState)(() => Se ? Sc(Re, G, Ue) : Oc({}, G, Ue)), Ke = Te !== void 0, [qe, Je] = (0, K.useState)(() => Ke ? Te : Oe ? wc(ze) ?? Ee : Ee), [Ye, Xe] = (0, K.useState)(null), [Ze, Qe] = (0, K.useState)(null), $e = ne(z("sm")), [et, tt] = (0, K.useState)({}), [nt, rt] = (0, K.useState)({}), [it, at] = (0, K.useState)({}), ot = p ?? it, st = (e, t) => {
		let n = oc(ot, e, t);
		p ? m?.(n) : at(n);
	}, [ct, lt] = (0, K.useState)({}), [ut, dt] = (0, K.useState)(null), [ft, pt] = (0, K.useState)(ce), [mt, ht] = (0, K.useState)(""), [gt, _t] = (0, K.useState)(() => {
		let e = G.find((e) => e.kind === "timestamp" && e.timestamp?.defaultRange);
		return {
			from: e?.timestamp?.defaultRange?.from ?? "",
			to: e?.timestamp?.defaultRange?.to ?? ""
		};
	});
	(0, K.useEffect)(() => {
		Ve((e) => {
			let t = Dc({
				...ye ? bc(Le, G) : {},
				...e
			}, G);
			return kc(e, t) ? e : t;
		});
	}, [
		Ie,
		ye,
		Le
	]), (0, K.useEffect)(() => {
		ye && xc(Le, Be);
	}, [
		Be,
		ye,
		Le
	]), (0, K.useEffect)(() => {
		Ge((e) => {
			let t = Oc({
				...Se ? Sc(Re, G, Ue) : {},
				...e
			}, G, Ue);
			return Ac(e, t) ? e : t;
		});
	}, [
		Ie,
		Ue,
		Se,
		Re
	]), (0, K.useEffect)(() => {
		Se && Cc(Re, We, Ue);
	}, [
		Ue,
		We,
		Se,
		Re
	]), (0, K.useEffect)(() => {
		Ke || Je(Oe ? wc(ze) ?? Ee : Ee);
	}, [
		Ee,
		Ke,
		Oe,
		ze
	]), (0, K.useEffect)(() => {
		!Ke && Oe && Tc(ze, qe);
	}, [
		Ke,
		qe,
		Oe,
		ze
	]);
	let vt = (0, K.useCallback)(() => {
		Xe(null), Qe(null);
	}, []), yt = !!Ye && $e;
	x(!!(Ye || Ze), vt, !yt), (0, K.useEffect)(() => {
		if ((Ye || Ze) && !yt) return document.addEventListener("click", vt), () => {
			document.removeEventListener("click", vt);
		};
	}, [
		vt,
		Ye,
		yt,
		Ze
	]);
	let bt = (0, K.useMemo)(() => e.map((e, t) => ({
		id: te?.(e, t) ?? String(t),
		row: e
	})), [e, te]), xt = (0, K.useMemo)(() => {
		let t = {};
		for (let n of G) {
			if (n.kind !== "timestamp") continue;
			let r = [];
			for (let t of e) {
				let e = At(t, n), i = T(e);
				i && r.push(i);
			}
			let i = ee(r);
			t[n.key] = I(n.timestamp?.mode ?? "auto", i);
		}
		return t;
	}, [G, e]), J = (0, K.useMemo)(() => G.map((e) => hc(e, xt[e.key])), [G, xt]), St = se ?? ft ?? ce ?? oe?.[0]?.value, Ct = oe?.find((e) => e.value === St), wt = (0, K.useMemo)(() => {
		if (!Ct) return ae;
		if (Ct.type === "none") return;
		if (Ct.type === "custom") return Ct;
		let e = J.find((e) => e.key === Ct.columnKey);
		if (!e) throw Error(`DataTable grouping mode "${Ct.value}" references missing column "${Ct.columnKey}"`);
		return {
			...Ct,
			getGroupKey: (t) => ps(At(t, e)),
			getGroupLabel: (e, t) => {
				let n = Ct.getGroupLabel?.(e, t);
				return n === void 0 ? fs(e, Ct.emptyGroupLabel) : n;
			}
		};
	}, [
		Ct,
		J,
		ae
	]), Tt = (0, K.useMemo)(() => {
		if (!xe) return J;
		let e = J.filter((e) => !We[e.key]);
		return e.length > 0 ? e : J;
	}, [
		J,
		We,
		xe
	]), Et = (0, K.useMemo)(() => J.filter(Ft).length, [J]), Dt = (0, K.useMemo)(() => Tt.filter(Ft).length, [Tt]), Ot = !W && xe && Et > 1, kt = Ae ?? Ot, Nt = !!(Fe && Fe.length > 0), It = Ot || kt || Ne || Nt, Rt = Ke ? Te : qe, Vt = (0, K.useMemo)(() => J.find((e) => e.kind === "timestamp" && e.filterable !== !1 && e.timestamp?.autoRangeFilter !== !1) ?? null, [J]), Y = (0, K.useMemo)(() => !f || _ || A || O?.timeRange ? null : Vt, [
		f,
		A,
		O?.timeRange,
		_,
		Vt
	]), Wt = (0, K.useMemo)(() => J.filter((e) => e.filterable !== !1 && e.kind !== "timestamp"), [J]), Gt = (0, K.useMemo)(() => f ? Wt.map((e) => {
		let t = Nc(bt, e);
		if (e.filterKind || t) return {
			column: e,
			kind: e.filterKind ?? "number",
			options: [],
			...t ? { numberBounds: t } : {}
		};
		let n = /* @__PURE__ */ new Set();
		for (let t of bt) for (let r of Mt(t.row, e)) r && n.add(r);
		let r = Array.from(n).sort((e, t) => e.localeCompare(t, void 0, { numeric: !0 })).map((e) => ({
			value: e,
			label: e
		})), i = e.kind === "tags" ? 50 : 20, a = r.length >= 2 && r.length <= i;
		if (e.kind === "tags" && a) {
			let t = mc(r, e.tags?.separator ?? "=");
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
		f,
		Wt,
		bt
	]), Kt = (0, K.useMemo)(() => {
		let e = {};
		for (let t of Gt) {
			if (t.kind !== "multi" && t.kind !== "nested-multi") continue;
			let n = t.column.key;
			e[n] = Yo(nt[n] ?? {}, (e) => rt((t) => ac(t, n, e)));
		}
		return e;
	}, [Gt, nt]);
	(0, K.useEffect)(() => {
		tt((e) => sc(e, Gt)), rt((e) => cc(e, Gt)), at((e) => lc(e, Gt));
	}, [Gt]);
	let qt = v !== void 0, Jt = qt ? v : mt, Yt = qt ? b ?? (() => {}) : ht, Xt = (0, K.useMemo)(() => Gt.map((e) => {
		let t = e.column.key;
		if (e.kind === "multi") return {
			key: t,
			kind: "multi",
			label: Pt(e.column),
			value: nt[t] ?? {},
			onChange: (e) => rt((n) => ac(n, t, e)),
			options: e.options
		};
		if (e.kind === "nested-multi") return {
			key: t,
			kind: "nested-multi",
			label: Pt(e.column),
			value: nt[t] ?? {},
			onChange: (e) => rt((n) => ac(n, t, e)),
			groups: e.groups ?? []
		};
		if (e.kind === "number" || e.kind === "duration") {
			if (e.kind === "duration") return {
				key: t,
				kind: "duration",
				label: Pt(e.column),
				unit: e.column.filterUnit ?? "ms",
				value: ot[t] ?? {},
				onChange: (e) => st(t, e)
			};
			let n = {
				key: t,
				kind: "number",
				label: Pt(e.column),
				value: ot[t] ?? {},
				onChange: (e) => st(t, e),
				...e.column.numberUnit ? { unit: e.column.numberUnit } : {}
			};
			return e.numberBounds?.min !== void 0 && (n.domainMin = e.numberBounds.min), e.numberBounds?.max !== void 0 && (n.domainMax = e.numberBounds.max), e.numberBounds?.step !== void 0 && (n.step = e.numberBounds.step), n;
		}
		return {
			key: t,
			kind: "text",
			label: Pt(e.column),
			value: et[t] ?? "",
			onChange: (e) => tt((n) => ic(n, t, e))
		};
	}), [
		Gt,
		nt,
		ot,
		et
	]), Zt = (0, K.useMemo)(() => new Map((j ?? []).map((e) => [e.key, e])), [j]), Qt = (0, K.useMemo)(() => {
		let e = Vt?.filterKey, t = e ? Zt.get(e) : void 0;
		return t?.kind === "date-range" ? t : null;
	}, [Vt, Zt]), $t = (0, K.useMemo)(() => {
		let e = /* @__PURE__ */ new Map();
		for (let t of J) {
			let n = Xt.find((e) => e.key === t.key);
			if (n) {
				e.set(t.key, n);
				continue;
			}
			let r = t.filterKey ? Zt.get(t.filterKey) : void 0;
			r && r !== Qt && e.set(t.key, r);
		}
		return e;
	}, [
		J,
		Xt,
		Zt,
		Qt
	]), en = !!(O?.leading || O?.children || O?.trailing || O?.timeRange || O?.dateRange), tn = (0, K.useMemo)(() => {
		if (!Y) return null;
		let e = {
			from: gt.from,
			to: gt.to,
			onApply: (e, t) => _t({
				from: e,
				to: t
			}),
			presets: Y.timestamp?.rangePresets ?? Ls,
			fromPlaceholder: "now-24h",
			toPlaceholder: "now",
			emptyLabel: "Any time"
		};
		return Y.timestamp?.timeEnabled !== void 0 && (e.timeEnabled = Y.timestamp.timeEnabled), Y.timestamp?.timeZone && (e.timeZone = Y.timestamp.timeZone), Y.timestamp?.timeZones && (e.timeZones = Y.timestamp.timeZones), e;
	}, [
		Y,
		gt.from,
		gt.to
	]), nn = (0, K.useMemo)(() => Qt && Vt ? {
		key: Vt.key,
		range: Lt(Qt)
	} : tn && Y ? {
		key: Y.key,
		range: tn
	} : null, [
		tn,
		Vt,
		Qt,
		Y
	]), rn = A ?? nn?.range ?? null, an = (0, K.useMemo)(() => {
		let e = (j ?? []).filter((e) => e !== Qt);
		return e.length > 0 ? [...e, ...Xt] : Xt;
	}, [
		j,
		Xt,
		Qt
	]), on = g ?? f;
	Bt({
		autoFilter: f,
		manualFilter: _,
		showGlobalFilter: on,
		globalFilterControlled: v !== void 0 || b !== void 0,
		hasExternalSearch: k !== void 0
	});
	let sn = on || Xt.length > 0 || !!rn || !!k || j && j.length > 0 || en || !!oe || It, cn = (e) => {
		Ke || Je(e), De?.(e);
	}, ln = Pe && ($t.size > 0 || !!nn), un = (0, K.useMemo)(() => {
		if (_) return bt;
		let e = Jt.trim().toLowerCase(), t = /* @__PURE__ */ new Date(), n = Y ? pe(gt.from, t) : null, r = Y ? pe(gt.to, t) : null;
		return bt.filter(({ row: t }) => {
			if (Y && (n || r)) {
				let e = T(jt(t, Y));
				if (!e || n && e.getTime() < n.getTime() || r && e.getTime() > r.getTime()) return !1;
			}
			if (e && !Wt.flatMap((e) => Mt(t, e)).join(" ").toLowerCase().includes(e)) return !1;
			for (let e of Gt) {
				let n = Mt(t, e.column);
				if (e.kind === "text") {
					let t = (et[e.column.key] ?? "").trim().toLowerCase();
					if (t && !n.some((e) => e.toLowerCase().includes(t))) return !1;
				} else if (e.kind === "number" || e.kind === "duration") {
					let n = ot[e.column.key] ?? {}, r = dc(n.min, e.kind === "duration" ? n.minUnit : void 0, e.column.filterUnit), i = dc(n.max, e.kind === "duration" ? n.maxUnit : void 0, e.column.filterUnit), a = String(n.min ?? "").trim() !== "", o = String(n.max ?? "").trim() !== "";
					if ((a || o) && !nc(t, e.column).some((e) => (r == null || (n.minOperator === ">" ? e > r : e >= r)) && (i == null || (n.maxOperator === "<" ? e < i : e <= i)))) return !1;
				} else {
					let t = nt[e.column.key] ?? {}, r = Object.entries(t).filter(([, e]) => e === "include").map(([e]) => e), i = Object.entries(t).filter(([, e]) => e === "exclude").map(([e]) => e);
					if (r.length > 0 && !n.some((e) => r.includes(e)) || i.length > 0 && n.some((e) => i.includes(e))) return !1;
				}
			}
			return !0;
		});
	}, [
		Jt,
		Wt,
		Gt,
		_,
		nt,
		ot,
		bt,
		et,
		Y,
		gt.from,
		gt.to
	]), dn = (0, K.useMemo)(() => Object.fromEntries(J.map((e) => [e.key, (t) => tc(t.row, e)])), [J]), { renderLink: fn } = y(), { sorted: pn, sort: mn, toggle: hn } = Ht(un, {
		defaultDir: C?.dir ?? "asc",
		manual: D,
		resolvers: dn,
		...w === void 0 ? {} : { sort: w },
		...E ? { onSortChange: E } : {},
		...C?.key ? { defaultKey: C.key } : {}
	}), gn = !!P && !E && !(Rs(P, e.length) && !F?.hasMore), _n = pn, vn = wa(F, e.length), yn = (0, K.useRef)(null), bn = !!F && F.hasMore, xn = F?.loading ? "Loading more…" : "Scroll to load more…", X = (0, K.useMemo)(() => new Set(R?.selectedRowIds ?? []), [R?.selectedRowIds]), Sn = (0, K.useMemo)(() => _n.filter((e) => R?.isRowSelectable?.(e.row) ?? !0), [R, _n]), Cn = R?.selectAll !== !1, wn = R?.selectAllPages;
	if (!Cn && wn) throw Error("DataTable rowSelection sets selectAll: false and selectAllPages; selecting across pages needs select-all");
	let Tn = (0, K.useMemo)(() => (wn?.scopes ?? []).flatMap((e) => {
		let t = e.total ?? P?.total;
		return t == null ? [] : [{
			...e,
			total: t
		}];
	}), [P?.total, wn?.scopes]), Z = wn?.selectedScopeId ? Tn.find((e) => e.id === wn.selectedScopeId) : void 0;
	if (wn?.selectedScopeId && !Z) throw Error(`Selected table scope ${wn.selectedScopeId} is not present in selectAllPages.scopes`);
	let En = Z ? Sn.length : Sn.filter((e) => X.has(e.id)).length, Dn = Sn.length > 0 && En === Sn.length, On = !Z && En > 0 && !Dn, kn = (0, K.useRef)(null), An = Tn[Tn.length - 1], jn = Dn && !Z ? Tn.find((e) => e.total > X.size) : void 0, Mn = !!An && (!!Z || X.size > 0);
	(0, K.useEffect)(() => {
		kn.current && (kn.current.indeterminate = On);
	}, [On]);
	let Nn = (0, K.useCallback)((e) => {
		R && R.onSelectionChange(Array.from(e), bt.filter((t) => e.has(t.id)).map((e) => e.row));
	}, [R, bt]), Pn = (0, K.useCallback)((e) => {
		if (!R || Z || R.isRowSelectable?.(e.row) === !1) return;
		let t = new Set(X);
		t.has(e.id) ? t.delete(e.id) : t.add(e.id), Nn(t);
	}, [
		Nn,
		R,
		X,
		Z
	]), Fn = (0, K.useCallback)(() => {
		if (!R || Z) return;
		let e = new Set(X);
		if (Dn) for (let t of Sn) e.delete(t.id);
		else for (let t of Sn) e.add(t.id);
		Nn(e);
	}, [
		Dn,
		Nn,
		R,
		Sn,
		X,
		Z
	]), In = (0, K.useMemo)(() => !R || !Z && X.size === 0 ? null : {
		selectedRowIds: Array.from(X),
		selectedRows: Z ? Sn.map((e) => e.row) : bt.filter((e) => X.has(e.id)).map((e) => e.row),
		...Z ? { selectedScope: {
			id: Z.id,
			total: Z.total
		} } : {},
		clearSelection: () => Nn(/* @__PURE__ */ new Set())
	}, [
		Nn,
		R,
		bt,
		Sn,
		X,
		Z
	]), Ln = (0, K.useMemo)(() => !In || typeof re != "function" ? null : re(In), [re, In]), Rn = Array.isArray(re) ? re : null, zn = t == null && B === "takeover" && !!In && !!Rn, Bn = In && Rn ? /* @__PURE__ */ (0, q.jsx)(Cs, {
		actions: Rn,
		context: In,
		showCount: !Mn,
		className: "flex flex-wrap items-center gap-density-2"
	}) : Ln, Vn = t == null ? ie === void 0 ? P ? null : r && pn.length === 0 ? a : `${pn.length} of ${e.length} row${e.length === 1 ? "" : "s"}` : typeof ie == "function" ? ie({
		visibleRowCount: pn.length,
		totalRowCount: e.length,
		loading: r
	}) : ie : null, [Hn, Un] = (0, K.useState)({}), Wn = Ct ? `mode:${Ct.value}` : "fixed", Gn = Hn[Wn], Kn = (0, K.useMemo)(() => {
		if (!wt) return null;
		let e = ms(_n, (e) => wt.getGroupKey(e.row)).map((e) => ({
			key: e.key,
			records: e.rows,
			rows: e.rows.map((e) => e.row)
		}));
		return wt.compareGroups && e.sort((e, t) => wt.compareGroups(e, t)), e.map((e) => ({
			...e,
			label: wt.getGroupLabel?.(e.key, e.rows) ?? e.key,
			meta: wt.getGroupMeta?.(e.key, e.rows),
			collapsed: Gn?.groups[e.key] ?? Gn?.all ?? hs(e, wt.defaultCollapsed)
		}));
	}, [
		Gn,
		wt,
		_n
	]), qn = !!Kn?.length, Jn = qn && Kn.every((e) => !e.collapsed), Yn = qn && Kn.every((e) => e.collapsed), Xn = (e) => {
		Un((t) => ({
			...t,
			[Wn]: {
				all: e,
				groups: {}
			}
		}));
	}, Zn = !zn && t == null && (Mn || Bn) ? /* @__PURE__ */ (0, q.jsxs)("div", {
		"data-testid": "data-table-selection-actions",
		className: "flex flex-wrap items-center gap-density-2",
		children: [Mn && An ? /* @__PURE__ */ (0, q.jsx)(Gs, {
			noun: wn?.noun ?? "rows",
			...wn?.loading ? { loading: !0 } : {},
			total: An.total,
			selectedCount: Z?.total ?? X.size,
			...jn ? { nextScope: jn } : {},
			onClear: () => Nn(/* @__PURE__ */ new Set())
		}) : null, Bn]
	}) : null, Qn = O?.trailing || oe || It ? /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
		O?.trailing,
		oe ? /* @__PURE__ */ (0, q.jsx)(gs, {
			modes: oe,
			value: St,
			hasGroups: qn,
			allExpanded: Jn,
			allCollapsed: Yn,
			onModeChange: (e) => {
				se === void 0 && pt(e), le?.(e);
			},
			onExpandAll: () => Xn(!1),
			onCollapseAll: () => Xn(!0)
		}) : null,
		It ? /* @__PURE__ */ (0, q.jsx)(Qs, {
			opensSheet: $e,
			onOpen: (e) => Xe(lo(e))
		}) : null
	] }) : void 0, $n = Zn || Qn ? /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [Zn, Qn] }) : void 0, er = zn ? /* @__PURE__ */ (0, q.jsxs)("div", {
		"data-testid": "data-table-selection-actions",
		className: "flex min-w-0 flex-1 flex-wrap items-center gap-density-2",
		children: [
			Mn && An ? /* @__PURE__ */ (0, q.jsx)(Gs, {
				noun: wn?.noun ?? "rows",
				...wn?.loading ? { loading: !0 } : {},
				total: An.total,
				selectedCount: Z?.total ?? X.size,
				...jn ? { nextScope: jn } : {},
				onClear: () => Nn(/* @__PURE__ */ new Set())
			}) : null,
			/* @__PURE__ */ (0, q.jsx)(Cs, {
				actions: Rn,
				context: In,
				showCount: !Mn,
				className: "flex flex-wrap items-center gap-density-2"
			}),
			Qn ? /* @__PURE__ */ (0, q.jsx)("div", {
				className: "ml-auto flex shrink-0 flex-wrap items-center gap-2",
				children: Qn
			}) : null
		]
	}) : null, tr = (0, K.useMemo)(() => {
		let e = (e) => {
			let t = {
				kind: "row",
				record: e
			}, n = U ? U.includes(e.id) : ct[e.id];
			if (he !== "row" || !H || !n) return [t];
			let r = H(e.row, {
				columns: J,
				visibleColumns: Tt,
				filterActionsByColumn: Kt
			});
			return r ? [t, {
				kind: "expansion",
				record: e,
				content: r
			}] : [t];
		};
		return Kn ? Kn.flatMap((t) => [{
			kind: "group",
			group: t
		}, ...t.collapsed ? [] : t.records.flatMap(e)]) : _n.flatMap(e);
	}, [
		Kn,
		_n,
		he,
		H,
		U,
		ct,
		J,
		Tt,
		Kt
	]), nr = (0, K.useCallback)((e) => {
		let t = tr[e];
		return t ? t.kind === "group" ? `group:${t.group.key}` : t.kind === "expansion" ? `expand:${t.record.id}` : `row:${t.record.id}` : `row:${e}`;
	}, [tr]), rr = (0, K.useMemo)(() => Za(c), [c]), Q = $a({
		count: tr.length,
		options: rr,
		getItemKey: nr
	}), [ir, ar] = (0, K.useState)({}), or = (0, K.useRef)(null), [sr, cr] = (0, K.useState)(null), lr = W ? 1 : Tt.length + +!!R, ur = Tt.map((e) => e.key).join(" ");
	(0, K.useEffect)(() => {
		if (!rr || W) {
			ar({});
			return;
		}
		let e = () => {
			let e = or.current;
			if (!e) return;
			let t = +!!R, n = {};
			Tt.forEach((r, i) => {
				let a = e.cells[i + t]?.getBoundingClientRect().width ?? 0;
				a > 0 && (n[r.key] = a);
			}), Object.keys(n).length > 0 && ar(n);
		}, t = requestAnimationFrame(e);
		if (typeof ResizeObserver > "u") return () => cancelAnimationFrame(t);
		let n = new ResizeObserver(e);
		return or.current?.parentElement && n.observe(or.current.parentElement), () => {
			cancelAnimationFrame(t), n.disconnect();
		};
	}, [
		rr,
		ur,
		R,
		Te,
		W
	]), (0, K.useEffect)(() => {
		if (!sr) return;
		let e = () => sr.style.setProperty("--dt-viewport-w", `${sr.clientWidth}px`);
		if (e(), typeof ResizeObserver > "u") return;
		let t = new ResizeObserver(e);
		return t.observe(sr), () => t.disconnect();
	}, [sr]);
	let dr = (0, K.useCallback)((e) => {
		Q.setScrollEl(e), cr(e), L && (L.current = e);
	}, [Q.setScrollEl, L]), fr = (0, K.useMemo)(() => Q.enabled ? Q.virtualItems.flatMap((e) => {
		let t = tr[e.index];
		return t ? [{
			item: t,
			index: e.index
		}] : [];
	}) : tr.map((e, t) => ({
		item: e,
		index: t
	})), [
		Q.enabled,
		Q.virtualItems,
		tr
	]), pr = (0, K.useCallback)((e) => {
		if (!R || Z) return;
		let t = e.filter((e) => R.isRowSelectable?.(e.row) ?? !0), n = new Set(X), r = t.length > 0 && t.every((e) => n.has(e.id));
		for (let e of t) r ? n.delete(e.id) : n.add(e.id);
		Nn(n);
	}, [
		Nn,
		R,
		X,
		Z
	]), mr = (0, K.useCallback)((e) => {
		yn.current && yn.current.disconnect(), e && bn && (yn.current = new IntersectionObserver((e) => {
			e[0]?.isIntersecting && vn();
		}, { rootMargin: "200px" }), yn.current.observe(e));
	}, [vn, bn]);
	if (W && (H || R)) throw Error("DataTable card rendering cannot be combined with row detail or row selection");
	let hr = (e, t) => {
		e.preventDefault(), e.stopPropagation();
		let n = e.clientX, r = (e.currentTarget.closest("th")?.getBoundingClientRect().width ?? 0) || Be[t.key] || jc(t), i = (e) => {
			let i = Mc(t, r + e.clientX - n);
			Ve((e) => ({
				...e,
				[t.key]: i
			}));
		}, a = () => {
			document.removeEventListener("mousemove", i), document.removeEventListener("mouseup", a), document.body.style.cursor = "", document.body.style.userSelect = "";
		};
		document.addEventListener("mousemove", i), document.addEventListener("mouseup", a), document.body.style.cursor = "col-resize", document.body.style.userSelect = "none";
	}, gr = (e, t) => {
		e.preventDefault(), e.stopPropagation();
		let n = e.currentTarget.closest("th"), r = e.currentTarget.closest("table");
		if (!n || !r) return;
		let i = Array.from(n.parentElement?.children ?? []).indexOf(n);
		if (i < 0) return;
		let a = _c(r, i, t);
		Ve((e) => ({
			...e,
			[t.key]: a
		}));
	}, _r = (e) => {
		Ft(e) && Ge((t) => {
			let n = t[e.key] === !0;
			if (!n && Dt <= 1) return t;
			let r = { ...t };
			return n ? Ue.has(e.key) ? r[e.key] = !1 : delete r[e.key] : r[e.key] = !0, Oc(r, G, Ue);
		});
	}, vr = () => Ge(Oc(Object.fromEntries([...Ue].map((e) => [e, !1])), G, Ue)), yr = (e, t) => {
		Ot && (e.preventDefault(), Xe(co(e, t.key)));
	}, br = (e, t) => {
		e.preventDefault(), e.stopPropagation(), Qe(lo(e, t));
	};
	return /* @__PURE__ */ (0, q.jsx)(Ut, {
		density: Rt ?? "comfortable",
		children: /* @__PURE__ */ (0, q.jsxs)("div", {
			className: i("flex min-h-0 min-w-0 flex-1 flex-col gap-1", u),
			"data-density": Rt,
			children: [
				(sn || Zn || er) && /* @__PURE__ */ (0, q.jsx)(fi, {
					...O,
					...er ? { overlay: er } : {},
					...k ? { search: k } : on ? { search: {
						value: Jt,
						onChange: Yt,
						placeholder: S
					} } : {},
					...rn ? { timeRange: rn } : {},
					trailing: $n,
					filters: an
				}),
				W && !gn && J.some((e) => e.sortable !== !1) && /* @__PURE__ */ (0, q.jsxs)("div", {
					className: "flex shrink-0 items-center gap-2 px-1 text-xs text-muted-foreground",
					children: [/* @__PURE__ */ (0, q.jsxs)("label", {
						className: "flex items-center gap-1.5",
						children: ["Sort by", /* @__PURE__ */ (0, q.jsxs)("select", {
							"aria-label": "Sort cards by",
							className: "h-7 rounded-md border border-input bg-background px-2 text-foreground",
							value: mn?.key ?? "",
							onChange: (e) => {
								let t = e.target.value;
								t ? E ? E({
									key: t,
									dir: "asc"
								}) : hn(t) : E?.(null);
							},
							children: [/* @__PURE__ */ (0, q.jsx)("option", {
								value: "",
								children: "Default order"
							}), J.filter((e) => e.sortable !== !1).map((e) => /* @__PURE__ */ (0, q.jsx)("option", {
								value: e.key,
								children: Pt(e)
							}, e.key))]
						})]
					}), mn && /* @__PURE__ */ (0, q.jsx)("button", {
						type: "button",
						className: "h-7 rounded-md border border-input px-2 text-foreground hover:bg-accent",
						onClick: () => {
							E ? E({
								key: mn.key,
								dir: mn.dir === "asc" ? "desc" : "asc"
							}) : hn(mn.key);
						},
						children: mn.dir === "asc" ? "Ascending" : "Descending"
					})]
				}),
				/* @__PURE__ */ (0, q.jsxs)("div", {
					className: "relative flex min-h-0 max-w-full flex-1 flex-col",
					children: [
						r && t == null ? /* @__PURE__ */ (0, q.jsx)(l, {
							"data-testid": "data-table-loading-bar",
							className: "rounded-t-md"
						}) : null,
						/* @__PURE__ */ (0, q.jsx)("div", {
							ref: dr,
							"data-testid": "data-table-scroll",
							className: i("min-h-0 max-w-full flex-1 overflow-auto overscroll-x-contain rounded-md border border-border bg-background", d),
							"aria-busy": r && t == null || void 0,
							children: /* @__PURE__ */ (0, q.jsxs)("table", {
								className: i(W ? "w-full text-left text-sm" : "w-max min-w-full text-left text-sm", W || Object.keys(ir).length > 0 ? "table-fixed" : "table-auto"),
								children: [
									W ? /* @__PURE__ */ (0, q.jsx)("colgroup", { children: /* @__PURE__ */ (0, q.jsx)("col", {}) }) : /* @__PURE__ */ (0, q.jsxs)("colgroup", { children: [R && t == null ? /* @__PURE__ */ (0, q.jsx)("col", { className: "w-10" }) : null, Tt.map((e) => /* @__PURE__ */ (0, q.jsx)("col", {
										style: t == null ? gc(e, Be, ir) : void 0,
										className: e.shrink && !e.grow && t == null ? "w-px" : void 0
									}, e.key))] }),
									W ? /* @__PURE__ */ (0, q.jsx)("thead", {
										className: "sr-only",
										children: /* @__PURE__ */ (0, q.jsx)("tr", { children: /* @__PURE__ */ (0, q.jsx)("th", {
											scope: "col",
											children: "Results"
										}) })
									}) : /* @__PURE__ */ (0, q.jsx)("thead", {
										className: "sticky top-0 z-10 bg-muted shadow-[0_1px_0_0_var(--tw-shadow-color)] shadow-border",
										children: /* @__PURE__ */ (0, q.jsxs)("tr", {
											ref: or,
											className: "border-b border-border text-xs text-muted-foreground",
											children: [R && t == null ? /* @__PURE__ */ (0, q.jsx)("th", {
												className: i("w-10 text-center", Fs),
												children: Cn ? /* @__PURE__ */ (0, q.jsx)("input", {
													ref: kn,
													type: "checkbox",
													"aria-label": "Select all visible rows",
													checked: Dn,
													disabled: !!Z || Sn.length === 0,
													onChange: Fn,
													className: "size-3.5 rounded border-border accent-primary"
												}) : null
											}) : null, Tt.map((e) => /* @__PURE__ */ (0, q.jsxs)("th", {
												className: i("group/header relative whitespace-nowrap font-medium", Fs, Rc(e.align), ve && e.resizable !== !1 && "select-none", e.headerClassName),
												onContextMenu: (t) => yr(t, e),
												children: [/* @__PURE__ */ (0, q.jsxs)("div", {
													className: i("flex min-w-0 items-center gap-1", zc(e.align), ve && e.resizable !== !1 && "pr-2"),
													children: [/* @__PURE__ */ (0, q.jsx)("span", {
														className: "min-w-0",
														children: e.sortable === !1 || gn ? /* @__PURE__ */ (0, q.jsx)("span", {
															...gn && e.sortable !== !1 ? { title: "Sorting is served by the query while this table is paged by the server" } : {},
															children: e.label
														}) : /* @__PURE__ */ (0, q.jsx)(mo, {
															active: mn?.key === e.key,
															...mn?.key === e.key ? { dir: mn.dir } : {},
															...e.align ? { align: e.align } : {},
															onClick: () => hn(e.key),
															children: e.label
														})
													}), ln && ($t.has(e.key) || nn?.key === e.key) && /* @__PURE__ */ (0, q.jsx)(Xs, {
														column: e,
														active: $t.has(e.key) ? zt($t.get(e.key)) : !!(nn?.range.from || nn?.range.to),
														onOpen: (t) => br(t, e.key)
													})]
												}), ve && e.resizable !== !1 && /* @__PURE__ */ (0, q.jsx)("span", {
													role: "separator",
													"aria-label": `Resize ${Pt(e)} column`,
													"aria-orientation": "vertical",
													className: "absolute right-0 top-0 hidden h-full w-3 cursor-col-resize touch-none items-center justify-center border-r border-border/70 bg-gradient-to-l from-border/30 to-transparent transition-colors hover:border-primary hover:from-primary/20 md:flex",
													onClick: (e) => {
														e.preventDefault(), e.stopPropagation();
													},
													onDoubleClick: (t) => gr(t, e),
													onMouseDown: (t) => hr(t, e),
													children: /* @__PURE__ */ (0, q.jsx)("span", {
														"aria-hidden": !0,
														className: "h-4 w-0.5 rounded-full bg-border transition-colors group-hover/header:bg-primary/70"
													})
												})]
											}, e.key))]
										})
									}),
									/* @__PURE__ */ (0, q.jsxs)("tbody", {
										ref: Q.setTbodyEl,
										children: [t == null ? r && _n.length === 0 ? /* @__PURE__ */ (0, q.jsx)(Js, {
											columns: Tt,
											rowCount: o,
											message: a,
											selection: !!R
										}) : un.length === 0 ? /* @__PURE__ */ (0, q.jsx)("tr", { children: /* @__PURE__ */ (0, q.jsx)("td", {
											colSpan: lr,
											className: "px-4 py-10 text-center text-sm text-muted-foreground",
											children: s
										}) }) : /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
											Q.enabled && Q.spacers.top > 0 && /* @__PURE__ */ (0, q.jsx)("tr", {
												"aria-hidden": !0,
												"data-virtual-spacer": "top",
												children: /* @__PURE__ */ (0, q.jsx)("td", {
													colSpan: lr,
													style: {
														height: Q.spacers.top,
														padding: 0,
														border: 0
													}
												})
											}),
											fr.map(({ item: e, index: t }) => {
												let n = Q.enabled ? {
													ref: Q.virtualizer.measureElement,
													"data-index": t
												} : {};
												if (e.kind === "expansion") return /* @__PURE__ */ (0, q.jsx)("tr", {
													...n,
													children: /* @__PURE__ */ (0, q.jsx)("td", {
														colSpan: lr,
														className: "bg-muted/40 p-density-3",
														children: /* @__PURE__ */ (0, q.jsx)("div", {
															"data-testid": "data-table-expansion",
															className: "sticky left-density-3 w-[calc(var(--dt-viewport-w)-2*var(--spacing-density-3))] overflow-x-auto rounded-md border border-border bg-background p-density-3",
															children: e.content
														})
													})
												}, `expand:${e.record.id}`);
												if (e.kind === "group") return /* @__PURE__ */ (0, q.jsx)(qs, {
													...Q.enabled ? {
														rowRef: Q.virtualizer.measureElement,
														dataIndex: t
													} : {},
													label: e.group.label,
													meta: e.group.meta,
													metaAlign: wt?.metaAlign ?? "end",
													...wt?.metaClassName ? { metaClassName: wt.metaClassName } : {},
													count: e.group.records.length,
													colSpan: lr,
													collapsed: e.group.collapsed,
													onToggleCollapsed: () => Un((t) => ({
														...t,
														[Wn]: {
															...t[Wn],
															groups: {
																...t[Wn]?.groups,
																[e.group.key]: !e.group.collapsed
															}
														}
													})),
													...R && Cn ? { selection: {
														selectableCount: e.group.records.filter((e) => R.isRowSelectable?.(e.row) ?? !0).length,
														selectedCount: Z ? e.group.records.filter((e) => R.isRowSelectable?.(e.row) ?? !0).length : e.group.records.filter((e) => (R.isRowSelectable?.(e.row) ?? !0) && X.has(e.id)).length,
														disabled: !!Z,
														onToggle: () => pr(e.group.records)
													} } : {}
												}, `group:${e.group.key}`);
												let r = e.record;
												if (W) return /* @__PURE__ */ (0, q.jsx)("tr", {
													...n,
													className: "align-top",
													children: /* @__PURE__ */ (0, q.jsx)("td", {
														className: "border-b border-border/60 p-density-2",
														children: W(r.row)
													})
												}, r.id);
												let a = fe?.(r.row), o = U ? U.includes(r.id) : ct[r.id] ?? !1, s = (H?.(r.row, {
													columns: J,
													visibleColumns: Tt,
													filterActionsByColumn: Kt
												}) ?? null) !== null, c = s && he === "row", l = s && he === "dialog", u = de?.(r.row) ?? !!ue, d = R?.isRowSelectable?.(r.row) ?? !0, f = !!R?.toggleOnRowClick && d && !Z, p = !!a || u || s || f, m = () => {
													f && Pn(r), c && (U ? me?.(o ? U.filter((e) => e !== r.id) : [...U, r.id]) : lt((e) => ({
														...e,
														[r.id]: !e[r.id]
													}))), l && dt(r), u && ue?.(r.row);
												};
												return /* @__PURE__ */ (0, q.jsx)(K.Fragment, { children: /* @__PURE__ */ (0, q.jsxs)("tr", {
													...n,
													"aria-expanded": c ? o : void 0,
													className: i("relative border-b border-border/60 align-top", p && "cursor-pointer hover:bg-accent/40", p && !a && "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring", (Z && d || X.has(r.id)) && "bg-accent/50", V?.(r.row)),
													tabIndex: p && !a ? 0 : void 0,
													onClick: m,
													onKeyDown: (e) => {
														a || e.target !== e.currentTarget || (e.key === "Enter" || e.key === " ") && (e.preventDefault(), m());
													},
													children: [R ? /* @__PURE__ */ (0, q.jsx)("td", {
														className: i("relative z-10 w-10 text-center", Is),
														onClick: (e) => e.stopPropagation(),
														children: /* @__PURE__ */ (0, q.jsx)("input", {
															type: "checkbox",
															"aria-label": `Select row ${r.id}`,
															checked: !!Z && d || X.has(r.id),
															disabled: !!Z || !d,
															onChange: () => Pn(r),
															className: "size-3.5 rounded border-border accent-primary"
														})
													}) : null, Tt.map((e, t) => {
														let n = At(r.row, e), o = e.render ? e.render(n, r.row) : pc(n), s = Vs(n, r.row, e), c = Hs(n, r.row, e), l = !!e.filterKey && !!N && s !== void 0;
														if (e.filterKey && N && s !== void 0) {
															let t = e.filterKey;
															o = /* @__PURE__ */ (0, q.jsx)(ns, {
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
														return e.kind === "tags" && Kt[e.key] && (o = /* @__PURE__ */ (0, q.jsx)(Xo, {
															value: Kt[e.key],
															children: o
														})), /* @__PURE__ */ (0, q.jsx)("td", {
															className: i(Is, Rc(e.align), e.cellClassName),
															style: Vc(e, Be[e.key]),
															children: /* @__PURE__ */ (0, q.jsx)(ec, {
																column: e,
																children: a && t === 0 ? fn({
																	to: a,
																	draggable: !1,
																	onDragStart: (e) => e.preventDefault(),
																	className: "hover:underline after:absolute after:inset-0 after:content-[''] [-webkit-user-drag:none]",
																	children: l ? /* @__PURE__ */ (0, q.jsx)("span", {
																		className: "relative z-10 inline-flex min-w-0 items-center",
																		children: o
																	}) : o
																}) : o
															})
														}, e.key);
													})]
												}) }, r.id);
											}),
											Q.enabled && Q.spacers.bottom > 0 && /* @__PURE__ */ (0, q.jsx)("tr", {
												"aria-hidden": !0,
												"data-virtual-spacer": "bottom",
												children: /* @__PURE__ */ (0, q.jsx)("td", {
													colSpan: lr,
													style: {
														height: Q.spacers.bottom,
														padding: 0,
														border: 0
													}
												})
											})
										] }) : /* @__PURE__ */ (0, q.jsx)(Ys, {
											colSpan: lr,
											error: t
										}), t == null && bn && /* @__PURE__ */ (0, q.jsx)("tr", {
											ref: mr,
											"aria-hidden": !0,
											children: /* @__PURE__ */ (0, q.jsx)("td", {
												colSpan: lr,
												className: "p-density-2 text-center text-xs text-muted-foreground",
												children: xn
											})
										})]
									})
								]
							})
						}),
						!W && /* @__PURE__ */ (0, q.jsx)(Ts, { scroll: sr })
					]
				}),
				Vn !== null && /* @__PURE__ */ (0, q.jsx)("div", {
					className: "shrink-0 px-1 text-xs text-muted-foreground",
					children: Vn
				}),
				t == null && P ? /* @__PURE__ */ (0, q.jsx)(Ks, {
					pagination: P,
					visibleRowCount: pn.length,
					infinite: !!F
				}) : null,
				he === "dialog" && H && /* @__PURE__ */ (0, q.jsx)(h, {
					open: ut !== null,
					onClose: () => dt(null),
					size: _e,
					...ut ? { title: typeof ge == "function" ? ge(ut.row) : ge } : {},
					children: ut ? H(ut.row, {
						columns: J,
						visibleColumns: Tt,
						filterActionsByColumn: Kt
					}) : null
				}),
				Ye && It && $e && /* @__PURE__ */ (0, q.jsx)(po, {
					open: !0,
					columns: J,
					hiddenColumns: We,
					activeColumnKey: Ye.columnKey,
					actions: Fe ?? [],
					showColumnVisibilityControl: Ot,
					showDensityControl: kt,
					showThemeControl: Ne,
					themeMenuValue: je,
					densityOverride: Rt,
					visibleHideableColumnCount: Dt,
					onToggle: _r,
					onShowAll: vr,
					onDensityChange: cn,
					...Me ? { onThemeChange: Me } : {},
					onClose: () => Xe(null)
				}),
				Ye && It && !$e && /* @__PURE__ */ (0, q.jsx)($s, {
					columns: J,
					hiddenColumns: We,
					anchor: Ye,
					actions: Fe ?? [],
					showColumnVisibilityControl: Ot,
					showDensityControl: kt,
					showThemeControl: Ne,
					themeMenuValue: je,
					densityOverride: Rt,
					visibleHideableColumnCount: Dt,
					onToggle: _r,
					onShowAll: vr,
					onDensityChange: cn,
					...Me ? { onThemeChange: Me } : {},
					onClose: () => Xe(null)
				}),
				Ze && /* @__PURE__ */ (0, q.jsx)(Zs, {
					filter: $t.get(Ze.columnKey ?? ""),
					...nn && nn.key === Ze.columnKey ? { timeRange: nn.range } : {},
					anchor: Ze,
					onClose: () => Qe(null)
				})
			]
		})
	});
}
function Vs(e, t, n) {
	let r = n.serverFilterValue ? n.serverFilterValue(e, t) : e;
	switch (typeof r) {
		case "string": return r || void 0;
		case "number":
		case "boolean": return String(r);
		default: return;
	}
}
function Hs(e, t, n) {
	let r = n.serverFilterLabel ? n.serverFilterLabel(e, t) : e;
	switch (typeof r) {
		case "string": return r || void 0;
		case "number":
		case "boolean": return String(r);
		default: return;
	}
}
var Us = [
	5,
	10,
	25,
	50,
	100,
	200
];
function Ws(e) {
	let [t, n] = (0, K.useState)([]), r = e?.current, i = r == null ? 0 : Math.max(t.indexOf(r) + 1, 1);
	return {
		page: i,
		push: (e) => n((t) => [...t.slice(0, i), e]),
		previous: () => i <= 1 ? void 0 : t[i - 2]
	};
}
function Gs({ noun: e, loading: t, total: n, selectedCount: r, nextScope: i, onClear: a }) {
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		role: "status",
		"data-testid": "data-table-selection-scope",
		className: "flex flex-wrap items-center gap-density-2 text-xs text-muted-foreground",
		children: [
			/* @__PURE__ */ (0, q.jsx)("span", { children: r >= n ? `All ${n.toLocaleString()} ${e} selected.` : `${r.toLocaleString()} of ${n.toLocaleString()} ${e} selected.` }),
			i ? /* @__PURE__ */ (0, q.jsx)("button", {
				type: "button",
				className: "rounded-md border border-border bg-background px-2 py-0.5 font-medium text-foreground hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50",
				disabled: t,
				onClick: i.onSelectAll,
				children: t ? "Selecting…" : i.label ?? `Select all ${i.total.toLocaleString()} ${e}`
			}) : null,
			/* @__PURE__ */ (0, q.jsx)("button", {
				type: "button",
				className: "rounded-md border border-border bg-background px-2 py-0.5 font-medium text-foreground hover:bg-accent",
				onClick: a,
				children: "Clear selection"
			})
		]
	});
}
function Ks({ pagination: e, visibleRowCount: t, infinite: n = !1 }) {
	let { page: r, pageSize: a, total: o, totalRelation: s, cursor: c, onPageChange: l, onPageSizeChange: u } = e, d = Array.from(/* @__PURE__ */ new Set([...e.pageSizeOptions ?? Us, a])).sort((e, t) => e - t), f = Ws(c), p = c ? f.page : Math.max(0, r), m = s === "gte", h = o != null && a > 0 && !m && !c ? Math.max(1, Math.ceil(o / a)) : void 0, g = p === 0, _ = c ? c.next == null : e.hasMore == null ? h == null ? t < a : p >= h - 1 : !e.hasMore, v = t > 0, y = o === 0 || !v ? 0 : n ? 1 : p * a + 1, b = o != null && v ? Math.min(m ? Infinity : o, y + Math.max(t - 1, 0)) : t, x = o == null ? `${t} row${t === 1 ? "" : "s"}` : `${y}-${b} of ${m ? `~${o}+` : o}`, S = !n && (c != null || l != null), C = () => {
		c ? c.onCursorChange(f.previous()) : l?.(Math.max(0, p - 1));
	}, w = () => {
		c ? c.next != null && (f.push(c.next), c.onCursorChange(c.next)) : l?.(p + 1);
	}, T = ne(z("sm"));
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "flex min-h-9 shrink-0 flex-wrap items-center justify-between gap-x-2 gap-y-1 border-t border-border/70 px-1 pt-2 text-xs text-muted-foreground max-sm:flex-nowrap sm:gap-3",
		children: [/* @__PURE__ */ (0, q.jsx)("div", {
			"aria-live": "polite",
			className: "min-w-0 truncate whitespace-nowrap",
			children: x
		}), /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "flex items-center gap-2 max-sm:flex-nowrap sm:flex-wrap sm:gap-3",
			children: [/* @__PURE__ */ (0, q.jsxs)("label", {
				className: "flex items-center gap-1.5",
				children: [/* @__PURE__ */ (0, q.jsx)("span", {
					className: "max-sm:sr-only",
					children: "Rows per page"
				}), /* @__PURE__ */ (0, q.jsx)("select", {
					className: "h-9 rounded-md border border-border bg-background px-2 text-xs text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:h-7",
					value: a,
					onChange: (e) => u(Number(e.target.value)),
					children: d.map((e) => /* @__PURE__ */ (0, q.jsx)("option", {
						value: e,
						children: e
					}, e))
				})]
			}), S ? /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
				/* @__PURE__ */ (0, q.jsx)("span", {
					className: i("text-center", !T && "min-w-20"),
					title: `Page ${p + 1}${h == null ? "" : ` of ${h}`}`,
					children: T ? `${p + 1}${h == null ? "" : `/${h}`}` : `Page ${p + 1}${h == null ? "" : ` of ${h}`}`
				}),
				/* @__PURE__ */ (0, q.jsx)("button", {
					type: "button",
					className: "inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-border bg-background text-foreground hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50 sm:h-7 sm:w-7",
					"aria-label": "Previous page",
					title: "Previous page",
					disabled: g,
					onClick: C,
					children: /* @__PURE__ */ (0, q.jsx)(E, { icon: ae })
				}),
				/* @__PURE__ */ (0, q.jsx)("button", {
					type: "button",
					className: "inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-border bg-background text-foreground hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50 sm:h-7 sm:w-7",
					"aria-label": "Next page",
					title: "Next page",
					disabled: _,
					onClick: w,
					children: /* @__PURE__ */ (0, q.jsx)(E, { icon: dt })
				})
			] }) : null]
		})]
	});
}
function qs({ label: e, meta: t, metaAlign: n, metaClassName: r, count: a, colSpan: o, collapsed: s, onToggleCollapsed: c, selection: l, rowRef: u, dataIndex: d }) {
	let f = !!l && l.selectableCount > 0 && l.selectedCount === l.selectableCount, p = !!l && l.selectedCount > 0 && !f, m = (0, K.useRef)(null);
	return (0, K.useEffect)(() => {
		m.current && (m.current.indeterminate = p);
	}, [p]), /* @__PURE__ */ (0, q.jsx)("tr", {
		ref: u,
		"data-index": d,
		className: "border-b border-border/60 bg-muted/40",
		children: /* @__PURE__ */ (0, q.jsx)("td", {
			colSpan: o,
			className: "px-density-2 py-density-1",
			children: /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "flex items-center gap-2",
				children: [
					l ? /* @__PURE__ */ (0, q.jsx)("input", {
						ref: m,
						type: "checkbox",
						"aria-label": `Select group ${typeof e == "string" ? e : a}`,
						checked: f,
						disabled: l.disabled || l.selectableCount === 0,
						onChange: l.onToggle,
						className: "size-3.5 rounded border-border accent-primary"
					}) : null,
					/* @__PURE__ */ (0, q.jsxs)("button", {
						type: "button",
						"aria-expanded": !s,
						onClick: c,
						className: i("flex min-w-0 items-center gap-1.5 text-left text-sm font-semibold text-foreground hover:text-primary", n === "end" && "flex-1"),
						children: [
							/* @__PURE__ */ (0, q.jsx)(E, {
								icon: s ? N : F,
								className: "shrink-0 text-muted-foreground"
							}),
							/* @__PURE__ */ (0, q.jsx)("span", {
								className: "truncate",
								children: e
							}),
							/* @__PURE__ */ (0, q.jsx)("span", {
								className: "shrink-0 font-normal text-muted-foreground",
								children: a
							})
						]
					}),
					t ? /* @__PURE__ */ (0, q.jsx)("div", {
						className: i("shrink-0 text-sm text-muted-foreground", r),
						children: t
					}) : null
				]
			})
		})
	});
}
function Js({ columns: e, rowCount: t, message: n, selection: r }) {
	let a = Math.max(1, t), o = Math.max(1, e.length + +!!r);
	return /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [/* @__PURE__ */ (0, q.jsx)("tr", {
		className: "border-b border-border/60 bg-muted/20",
		children: /* @__PURE__ */ (0, q.jsxs)("td", {
			colSpan: o,
			className: i(Is, "p-0"),
			children: [/* @__PURE__ */ (0, q.jsx)("div", {
				className: "h-0.5 w-full overflow-hidden bg-muted",
				children: /* @__PURE__ */ (0, q.jsx)("div", { className: "h-full w-1/2 animate-pulse rounded-full bg-primary/70" })
			}), /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground",
				children: [/* @__PURE__ */ (0, q.jsx)("span", {
					"aria-hidden": !0,
					className: "h-3 w-3 animate-spin rounded-full border-2 border-primary/30 border-t-primary"
				}), /* @__PURE__ */ (0, q.jsx)("span", { children: n })]
			})]
		})
	}), Array.from({ length: a }).map((t, n) => /* @__PURE__ */ (0, q.jsxs)("tr", {
		className: "border-b border-border/60",
		children: [r ? /* @__PURE__ */ (0, q.jsx)("td", {
			className: i("w-10", Is),
			children: /* @__PURE__ */ (0, q.jsx)("span", { className: "mx-auto block size-3.5 animate-pulse rounded bg-muted" })
		}) : null, e.map((e, t) => /* @__PURE__ */ (0, q.jsx)("td", {
			className: i(Is, Rc(e.align), e.cellClassName),
			style: Vc(e, void 0),
			children: /* @__PURE__ */ (0, q.jsx)(ec, {
				column: e,
				children: /* @__PURE__ */ (0, q.jsx)("span", {
					"aria-hidden": !0,
					className: i("block h-3 animate-pulse rounded-full bg-muted", Bc(t))
				})
			})
		}, e.key))]
	}, n))] });
}
function Ys({ colSpan: e, error: t }) {
	return /* @__PURE__ */ (0, q.jsx)("tr", {
		className: "border-b border-border/60",
		children: /* @__PURE__ */ (0, q.jsx)("td", {
			colSpan: Math.max(1, e),
			className: i(Is, "p-density-3"),
			children: /* @__PURE__ */ (0, q.jsx)("div", {
				role: "alert",
				className: "w-0 min-w-full overflow-hidden",
				children: t
			})
		})
	});
}
function Xs({ column: e, active: t, onOpen: n }) {
	return /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		"aria-label": `Open ${Pt(e)} column filter`,
		"aria-haspopup": "dialog",
		"aria-pressed": t,
		className: i("inline-flex h-5 w-5 shrink-0 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", t && "bg-accent text-foreground"),
		onClick: n,
		children: /* @__PURE__ */ (0, q.jsx)("span", {
			"data-filter-icon-state": t ? "filled" : "outline",
			className: "inline-flex",
			children: /* @__PURE__ */ (0, q.jsx)(E, {
				icon: t ? Qe : Ze,
				className: "text-xs"
			})
		})
	});
}
function Zs({ filter: e, timeRange: t, anchor: n, onClose: r }) {
	let i = fo(n);
	if (!e && !t) return null;
	let a = e?.label ?? "Time range", o = !!(t?.from || t?.to);
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		ref: i.ref,
		role: "dialog",
		"aria-label": `${a} column filter`,
		className: "z-50 max-w-[calc(100vw-1rem)] overflow-auto rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-lg shadow-black/5",
		style: i.style,
		onClick: (e) => e.stopPropagation(),
		onContextMenu: (e) => e.preventDefault(),
		children: [/* @__PURE__ */ (0, q.jsxs)("div", {
			className: "mb-2 flex items-center justify-between gap-3",
			children: [/* @__PURE__ */ (0, q.jsx)("div", {
				className: "truncate text-xs font-semibold uppercase tracking-wide text-muted-foreground",
				children: a
			}), /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "flex items-center gap-1",
				children: [
					e && /* @__PURE__ */ (0, q.jsx)("button", {
						type: "button",
						className: "rounded px-1.5 py-0.5 text-xs text-primary transition-colors hover:bg-accent focus:bg-accent focus:outline-none disabled:text-muted-foreground",
						onClick: () => Rt(e),
						disabled: !zt(e),
						children: "Clear all"
					}),
					t && /* @__PURE__ */ (0, q.jsx)("button", {
						type: "button",
						className: "rounded px-1.5 py-0.5 text-xs text-primary transition-colors hover:bg-accent focus:bg-accent focus:outline-none disabled:text-muted-foreground",
						onClick: () => t.onApply("", ""),
						disabled: !o,
						children: "Clear all"
					}),
					/* @__PURE__ */ (0, q.jsx)("button", {
						type: "button",
						"aria-label": "Close column filter",
						title: "Close",
						className: "inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none",
						onClick: r,
						children: /* @__PURE__ */ (0, q.jsx)(E, {
							icon: B,
							className: "text-sm"
						})
					})
				]
			})]
		}), /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "flex flex-col gap-2",
			children: [e && /* @__PURE__ */ (0, q.jsx)(pi, {
				filter: e,
				chrome: "embedded"
			}), t && /* @__PURE__ */ (0, q.jsx)(Mi, {
				kind: "time",
				label: "Time range",
				...t
			})]
		})]
	});
}
function Qs({ opensSheet: e, onOpen: t }) {
	return /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		"aria-label": "Open column menu",
		"aria-haspopup": e ? "dialog" : "menu",
		className: "inline-flex h-8 w-8 max-sm:h-9 max-sm:w-9 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
		onClick: (e) => {
			e.stopPropagation(), t(e);
		},
		children: /* @__PURE__ */ (0, q.jsx)(E, {
			icon: Xe,
			className: "text-sm"
		})
	});
}
function $s({ columns: e, hiddenColumns: t, anchor: n, actions: r, showColumnVisibilityControl: a, showDensityControl: o, showThemeControl: s, themeMenuValue: c, densityOverride: l, visibleHideableColumnCount: u, onToggle: d, onShowAll: f, onDensityChange: p, onThemeChange: m, onClose: h }) {
	let g = n.columnKey ? e.find((e) => e.key === n.columnKey) : void 0, _ = g && Ft(g) && u > 1, v = r.length > 0, y = fo(n);
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		ref: y.ref,
		role: "menu",
		"aria-label": "Column menu",
		className: "z-50 min-w-[16rem] max-w-[calc(100vw-1rem)] overflow-auto rounded-md border border-border bg-popover p-1.5 text-popover-foreground shadow-lg shadow-black/5",
		style: y.style,
		onClick: (e) => e.stopPropagation(),
		onContextMenu: (e) => e.preventDefault(),
		children: [
			a && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
				/* @__PURE__ */ (0, q.jsxs)("div", {
					className: "flex items-center justify-between gap-3 px-2 py-1.5 text-xs font-medium text-muted-foreground",
					children: [/* @__PURE__ */ (0, q.jsx)("span", { children: "Columns" }), /* @__PURE__ */ (0, q.jsx)("button", {
						type: "button",
						className: "rounded px-1.5 py-0.5 text-xs text-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none",
						onClick: f,
						children: "Show all"
					})]
				}),
				g && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [/* @__PURE__ */ (0, q.jsxs)("button", {
					type: "button",
					role: "menuitem",
					disabled: !_,
					className: i("flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", !_ && "cursor-not-allowed opacity-50"),
					onClick: () => {
						_ && (d(g), h());
					},
					children: [/* @__PURE__ */ (0, q.jsx)(E, {
						icon: et,
						className: "text-sm text-muted-foreground"
					}), /* @__PURE__ */ (0, q.jsxs)("span", { children: ["Hide ", Pt(g)] })]
				}), /* @__PURE__ */ (0, q.jsx)("div", { className: "my-1 h-px bg-border" })] }),
				/* @__PURE__ */ (0, q.jsx)("div", {
					className: "max-h-72 overflow-auto",
					children: e.map((e) => {
						let n = Ft(e), r = t[e.key] !== !0, a = !n || r && u <= 1;
						return /* @__PURE__ */ (0, q.jsxs)("label", {
							className: i("flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors hover:bg-accent hover:text-accent-foreground", a && "cursor-not-allowed opacity-50"),
							children: [/* @__PURE__ */ (0, q.jsx)("input", {
								type: "checkbox",
								className: "h-4 w-4 rounded border-border",
								checked: r,
								disabled: a,
								onChange: () => d(e)
							}), /* @__PURE__ */ (0, q.jsx)("span", {
								className: "truncate",
								children: Pt(e)
							})]
						}, e.key);
					})
				})
			] }),
			o && /* @__PURE__ */ (0, q.jsx)(io, {
				densityOverride: l,
				separated: a,
				onDensityChange: p
			}),
			v && /* @__PURE__ */ (0, q.jsx)(no, {
				actions: r,
				separated: a || o,
				onClose: h
			}),
			s && m && /* @__PURE__ */ (0, q.jsx)(so, {
				value: c ?? "system",
				separated: a || o || v,
				onChange: m
			})
		]
	});
}
function ec({ column: e, children: t }) {
	return /* @__PURE__ */ (0, q.jsx)("div", {
		className: i(Hc(e), Rc(e.align)),
		children: t
	});
}
function tc(e, t) {
	let n = At(e, t);
	return t.sortValue ? t.sortValue(n, e) : n;
}
function nc(e, t) {
	return rc(jt(e, t));
}
function rc(e) {
	if (Array.isArray(e)) return e.flatMap((e) => rc(e));
	let t = Ic(e);
	return t == null ? [] : [t];
}
function ic(e, t, n) {
	let r = { ...e };
	return n.trim() ? r[t] = n : delete r[t], r;
}
function ac(e, t, n) {
	let r = { ...e };
	return Object.keys(n).length > 0 ? r[t] = n : delete r[t], r;
}
function oc(e, t, n) {
	let r = { ...e }, i = String(n.min ?? "").trim() !== "", a = String(n.max ?? "").trim() !== "";
	return i || a ? r[t] = n : delete r[t], r;
}
function sc(e, t) {
	return fc(e, t, "text", (e) => !String(e ?? "").trim());
}
function cc(e, t) {
	let n = new Set(t.filter((e) => e.kind === "multi" || e.kind === "nested-multi").map((e) => e.column.key));
	return Object.fromEntries(Object.entries(e).filter(([e, t]) => n.has(e) && Object.keys(t).length > 0));
}
function lc(e, t) {
	let n = new Set(t.filter((e) => e.kind === "number" || e.kind === "duration").map((e) => e.column.key));
	return Object.fromEntries(Object.entries(e).filter(([e, t]) => n.has(e) && (String(t.min ?? "").trim() || String(t.max ?? "").trim())));
}
var uc = {
	ms: 1,
	s: 1e3,
	m: 6e4,
	h: 36e5
};
function dc(e, t, n) {
	let r = Ic(e);
	return r == null ? null : r * uc[t ?? n ?? "ms"] / uc[n ?? "ms"];
}
function fc(e, t, n, r) {
	let i = new Set(t.filter((e) => e.kind === n).map((e) => e.column.key));
	return Object.fromEntries(Object.entries(e).filter(([e, t]) => i.has(e) && !r(t)));
}
function pc(e) {
	return e == null || e === "" ? /* @__PURE__ */ (0, q.jsx)("span", {
		className: "text-muted-foreground",
		children: "—"
	}) : typeof e == "boolean" ? e ? "True" : "False" : typeof e == "object" ? JSON.stringify(e) : String(e);
}
function mc(e, t) {
	let n = /* @__PURE__ */ new Map();
	for (let r of e) {
		let { key: e } = Ko(r.value, t), i = n.get(e) ?? [];
		i.push(r), n.set(e, i);
	}
	return Array.from(n.entries()).sort(([e], [t]) => e === "" && t !== "" ? 1 : t === "" && e !== "" ? -1 : e.localeCompare(t)).map(([e, t]) => ({
		groupKey: e,
		label: e === "" ? "Other" : e,
		options: t
	}));
}
function hc(e, t) {
	if (!e.kind) return e;
	if (e.kind === "timestamp") {
		let n = t ?? "iso";
		return {
			...e,
			render: e.render ?? ((t) => /* @__PURE__ */ (0, q.jsx)(ho, {
				value: t,
				format: n,
				showTitleOnHover: e.timestamp?.alwaysShowFullOnHover !== !1
			})),
			sortValue: e.sortValue ?? ((e) => T(e)?.getTime() ?? null),
			filterValue: e.filterValue ?? ((e) => {
				let t = T(e);
				return t ? t.toISOString() : "";
			})
		};
	}
	if (e.kind === "tags") {
		let t = e.tags, n = t?.separator ?? "=";
		return {
			...e,
			render: e.render ?? ((e) => /* @__PURE__ */ (0, q.jsx)(es, {
				tags: Uo(e, n),
				maxVisible: t?.maxVisible ?? 3
			})),
			filterValue: e.filterValue ?? ((e) => Go(e, n)),
			sortValue: e.sortValue ?? ((e) => Go(e, n).length)
		};
	}
	if (e.kind === "status") {
		let t = e.status, n = t?.map ?? ((e) => us(e));
		return {
			...e,
			render: e.render ?? ((e, r) => {
				let i = n(e, r);
				if (!i) return /* @__PURE__ */ (0, q.jsx)("span", {
					className: "text-muted-foreground",
					children: "—"
				});
				let a = t?.showLabel ? typeof e == "string" ? e : i : void 0, o = t?.title?.(e, r) ?? (typeof e == "string" ? e : void 0);
				return /* @__PURE__ */ (0, q.jsx)(os, {
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
function gc(e, t, n = {}) {
	let r = t[e.key] ?? n[e.key];
	return r ? { width: `${r}px` } : void 0;
}
function _c(e, t, n) {
	return Mc(n, Array.from(e.querySelectorAll("th, td")).filter((e) => e instanceof HTMLTableCellElement && e.cellIndex === t && e.colSpan === 1).reduce((e, t) => {
		let n = vc(t), r = Array.from(t.children).map((e) => e instanceof HTMLElement ? e.scrollWidth + n : 0);
		return Math.max(e, t.scrollWidth, ...r);
	}, 0) || jc(n));
}
function vc(e) {
	if (typeof window > "u") return 0;
	let t = window.getComputedStyle(e);
	return yc(t.paddingLeft) + yc(t.paddingRight);
}
function yc(e) {
	let t = Number.parseFloat(e);
	return Number.isFinite(t) ? t : 0;
}
function bc(e, t) {
	if (typeof window > "u") return {};
	try {
		let n = window.localStorage.getItem(e);
		if (!n) return {};
		let r = JSON.parse(n);
		return !r || typeof r != "object" || Array.isArray(r) ? {} : Dc(r, t);
	} catch {
		return {};
	}
}
function xc(e, t) {
	if (typeof window < "u") try {
		Object.keys(t).length === 0 ? window.localStorage.removeItem(e) : window.localStorage.setItem(e, JSON.stringify(t));
	} catch {}
}
function Sc(e, t, n) {
	let r = Oc({}, t, n);
	if (typeof window > "u") return r;
	try {
		let i = window.localStorage.getItem(e);
		if (!i) return r;
		let a = JSON.parse(i);
		return !a || typeof a != "object" || Array.isArray(a) ? r : Oc(a, t, n);
	} catch {
		return r;
	}
}
function Cc(e, t, n) {
	if (typeof window > "u") return;
	let r = Object.fromEntries(Object.entries(t).filter(([e, t]) => t !== n.has(e)));
	try {
		Object.keys(r).length === 0 ? window.localStorage.removeItem(e) : window.localStorage.setItem(e, JSON.stringify(r));
	} catch {}
}
function wc(e) {
	if (typeof window < "u") try {
		let t = window.localStorage.getItem(e);
		return Ec(t) ? t : void 0;
	} catch {
		return;
	}
}
function Tc(e, t) {
	if (typeof window < "u") try {
		t ? window.localStorage.setItem(e, t) : window.localStorage.removeItem(e);
	} catch {}
}
function Ec(e) {
	return e === "compact" || e === "comfortable" || e === "spacious";
}
function Dc(e, t) {
	let n = new Map(t.map((e) => [e.key, e])), r = {};
	for (let [t, i] of Object.entries(e)) {
		let e = n.get(t);
		e && typeof i == "number" && Number.isFinite(i) && (r[t] = Mc(e, i));
	}
	return r;
}
function Oc(e, t, n) {
	let r = t.filter(Ft), i = new Set(r.map((e) => e.key)), a = {};
	for (let e of n) i.has(e) && (a[e] = !0);
	for (let [t, r] of Object.entries(e)) i.has(t) && (r === !0 ? a[t] = !0 : r === !1 && n.has(t) && (a[t] = !1));
	let o = r.filter((e) => a[e.key] !== !0).length;
	if (r.length > 0 && o === 0) {
		let e = r[0].key;
		n.has(e) ? a[e] = !1 : delete a[e];
	}
	return a;
}
function kc(e, t) {
	let n = Object.keys(e), r = Object.keys(t);
	return n.length === r.length && n.every((n) => Object.prototype.hasOwnProperty.call(t, n) && e[n] === t[n]);
}
function Ac(e, t) {
	let n = Object.keys(e), r = Object.keys(t);
	return n.length === r.length && n.every((n) => Object.prototype.hasOwnProperty.call(t, n) && e[n] === t[n]);
}
function jc(e) {
	return e.grow ? Os : e.shrink ? ks : Ds;
}
function Mc(e, t) {
	let n = e.minWidth ?? Es, r = e.maxWidth, i = Number.isFinite(n) ? Math.max(1, n) : Es;
	return Math.round(Math.max(i, Math.min(r != null && Number.isFinite(r) ? Math.max(i, r) : Infinity, t)));
}
function Nc(e, t) {
	let n = e.flatMap((e) => Pc(jt(e.row, t))).filter(Fc), r = n.map((e) => Ic(e)).filter((e) => e != null);
	return n.length === 0 || r.length !== n.length ? null : {
		min: Math.min(...r),
		max: Math.max(...r),
		step: Lc(r)
	};
}
function Pc(e) {
	return Array.isArray(e) ? e.flatMap((e) => Pc(e)) : [e];
}
function Fc(e) {
	return e != null && (typeof e != "string" || e.trim() !== "");
}
function Ic(e) {
	if (typeof e == "number") return Number.isFinite(e) ? e : null;
	if (typeof e != "string") return null;
	let t = e.trim();
	if (!t) return null;
	let n = Number(t);
	return Number.isFinite(n) ? n : null;
}
function Lc(e) {
	if (e.every(Number.isInteger)) return 1;
	let t = e.reduce((e, t) => {
		let [, n = ""] = String(t).split(".");
		return Math.max(e, n.length);
	}, 0);
	return 10 ** -Math.min(t || 2, 6);
}
function Rc(e) {
	return e === "right" ? "text-right" : e === "center" ? "text-center" : "text-left";
}
function zc(e) {
	return e === "right" ? "justify-end" : e === "center" ? "justify-center" : "justify-start";
}
function Bc(e) {
	return [
		"w-24",
		"w-40",
		"w-28",
		"w-56",
		"w-32",
		"w-48"
	][e % 6];
}
function Vc(e, t) {
	return t === void 0 ? e.grow ? {
		maxWidth: "0px",
		minWidth: e.minWidth ?? Os
	} : { maxWidth: e.maxWidth ?? (e.shrink ? js : As) } : { maxWidth: `${t}px` };
}
function Hc(e) {
	return i("w-full min-w-0 truncate", e.shrink && "whitespace-nowrap");
}
function Uc({ theme: e = "system", onThemeChange: t, showThemeControl: n = !1, showFullscreenControl: r = !1, fullscreenTitle: a = "Table", fullscreenButtonLabel: o = "Open table full screen", className: s, filterBarProps: c, scrollContainerRef: l, ...u }) {
	let d = t !== void 0, [f, p] = (0, K.useState)(e), m = d ? e : f, g = ce(m), _ = (e) => {
		d || p(e), t?.(e);
	}, [v, y] = (0, K.useState)(!1), b = ({ inFullscreen: e }) => {
		let t = r && !e, i = c?.trailing, a = t ? {
			...c,
			trailing: /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [i, /* @__PURE__ */ (0, q.jsx)(Wc, {
				label: o,
				onClick: () => y(!0)
			})] })
		} : c, s = e === v;
		return /* @__PURE__ */ (0, q.jsx)(Bs, {
			...u,
			...s && l ? { scrollContainerRef: l } : {},
			...a ? { filterBarProps: a } : {},
			themeMenuValue: m,
			showThemeControl: n,
			onThemeChange: _
		});
	};
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		"data-theme": g,
		className: i("flex min-h-0 flex-1 flex-col text-foreground", s),
		children: [b({ inFullscreen: !1 }), r && /* @__PURE__ */ (0, q.jsx)(h, {
			open: v,
			onClose: () => y(!1),
			title: a,
			size: "full",
			className: "h-[95vh]",
			children: /* @__PURE__ */ (0, q.jsx)("div", {
				"data-theme": g,
				className: "flex h-full min-h-0 flex-col bg-background text-foreground",
				children: b({ inFullscreen: !0 })
			})
		})]
	});
}
function Wc({ label: e, onClick: t }) {
	return /* @__PURE__ */ (0, q.jsx)("button", {
		type: "button",
		"aria-label": e,
		title: e,
		className: "inline-flex h-8 w-8 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
		onClick: t,
		children: /* @__PURE__ */ (0, q.jsx)(E, {
			icon: ue,
			className: "text-sm"
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/LogViewer.tsx
var Gc = {
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
function Kc({ logs: e, collapsedLines: t = 5, maxHeight: n, maxExpandedVh: r = 70, variant: a = "muted", bgClass: o, borderClass: s, className: c }) {
	let [l, u] = (0, K.useState)(!1), d = (0, K.useMemo)(() => e.split("\n"), [e]), f = d.length > t, p = Gc[a], m = n ?? `${t * 1.4}em`, h = l || n ? e : d.slice(0, t).join("\n");
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i("relative", c),
		children: [/* @__PURE__ */ (0, q.jsx)("pre", {
			className: i("mt-0.5 text-[11px] rounded p-1.5 whitespace-pre-wrap overflow-y-auto border transition-all duration-200", p.text, o ?? p.bg, s ?? p.border),
			style: { maxHeight: l ? `${r}vh` : m },
			"data-log-variant": a,
			children: /* @__PURE__ */ (0, q.jsx)(Ye, {
				as: "span",
				text: h
			})
		}), f && /* @__PURE__ */ (0, q.jsx)("button", {
			type: "button",
			className: i("text-[10px] mt-0.5", l ? "text-muted-foreground" : "text-primary hover:opacity-80"),
			onClick: () => u((e) => !e),
			children: l ? `▲ Collapse (${d.length} lines)` : `▼ Show more (${d.length} lines)`
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/lib/string.ts
function qc(e) {
	let t = e.length;
	for (; t > 0 && e[t - 1] === "/";) t--;
	return e.slice(0, t);
}
function Jc(e, t) {
	let n = 0, r = e.length;
	for (; n < r && t.includes(e[n]);) n++;
	for (; r > n && t.includes(e[r - 1]);) r--;
	return e.slice(n, r);
}
function Yc(e, t = " ") {
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
// ../../../clicky-ui/packages/ui/src/data/diagnostics/frame-heuristics.ts
var Xc = [
	"java.",
	"javax.",
	"sun.",
	"jdk.",
	"com.sun.",
	"oracle.jrockit."
];
function Zc(e, t = []) {
	return Xc.some((t) => e.startsWith(t)) || t.some((t) => e.startsWith(t));
}
function Qc(e, t) {
	return [e.split(".").filter(Boolean).pop() ?? e, t].filter(Boolean).join(".");
}
function $c(e, t) {
	if (e) return t === void 0 ? e : `${e}:${t}`;
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/FrameSourceWindow.tsx
function el({ frame: e, className: t, focalTone: n = "error", markers: r }) {
	let a = (0, K.useMemo)(() => e.sourceLines ?? [], [e.sourceLines]), o = e.sourceStartLine ?? 0, s = e.sourceLineNumbers, c = e.line ?? -1, l = e.sourceLanguage, [u, d] = (0, K.useState)(null);
	if ((0, K.useEffect)(() => {
		if (a.length === 0) {
			d(null);
			return;
		}
		let e = !1;
		return _(a.join("\n"), { lang: l }).then((t) => {
			e || d(t);
		}), () => {
			e = !0;
		};
	}, [a, l]), a.length === 0) return null;
	let f = "bg-primary/10 font-semibold text-foreground", m = n === "neutral" ? f : "bg-red-500/10 font-semibold text-red-800 dark:text-red-300", h = new Map(r?.map((e) => [e.line, e.content]));
	return /* @__PURE__ */ (0, q.jsx)("div", {
		"data-scroll-x": !0,
		className: i("mx-3 mb-2 ml-12 overflow-x-auto max-sm:ml-6 rounded border border-border/60 bg-muted/30 py-1 font-mono text-[11px] leading-relaxed", t),
		children: a.map((e, t) => {
			let n = s?.[t] ?? o + t, r = s === void 0 || n > 0, i = r && n === c, a = r ? h.get(n) : void 0;
			return /* @__PURE__ */ (0, q.jsxs)("div", {
				"data-source-line": r ? n : void 0,
				"data-marked": a === void 0 ? void 0 : "true",
				className: ["grid min-w-max grid-cols-[3.5rem_minmax(24rem,1fr)] gap-3 px-2", i ? m : a === void 0 ? "text-foreground" : f].join(" "),
				children: [
					/* @__PURE__ */ (0, q.jsxs)("span", {
						className: "select-none text-right text-muted-foreground",
						children: [i || a !== void 0 ? ">" : "", r ? n : ""]
					}),
					/* @__PURE__ */ (0, q.jsx)("code", {
						className: "whitespace-pre",
						children: e ? /* @__PURE__ */ (0, q.jsx)(p, {
							tokens: u?.[t],
							content: e
						}) : " "
					}),
					a === void 0 ? null : /* @__PURE__ */ (0, q.jsx)("span", {
						"data-source-marker": !0,
						className: "col-start-2 -mt-0.5 mb-0.5 font-sans font-normal",
						children: a
					})
				]
			}, `${n}-${t}`);
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/FrameSourceWindow.utils.ts
function tl(e) {
	return !!e.sourceLines && e.sourceLines.length > 0;
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/StackFrameRow.tsx
function nl({ frame: e, index: t, showIndex: n = !1, frameActions: r, renderFramePart: i, renderFrameDetail: a, frameActionsVisibility: o = "hover" }) {
	let s = e.kind !== "frame", c = e.displayName || e.method || e.functionName, l = r?.(e, t), u = s ? null : a?.(e, t);
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: [
			"group grid gap-x-2 gap-y-0.5 px-3 py-0.5 text-xs",
			n ? "grid-cols-[2rem_minmax(0,1fr)_auto]" : "grid-cols-[1.25rem_minmax(0,1fr)_auto]",
			e.runtime ? "text-muted-foreground" : "text-foreground"
		].join(" "),
		children: [
			/* @__PURE__ */ (0, q.jsxs)("div", {
				className: "flex items-start justify-end gap-1 pt-0.5 font-mono text-[10px] text-muted-foreground",
				children: [n && /* @__PURE__ */ (0, q.jsx)("span", {
					"data-frame-index": !0,
					children: t + 1
				}), /* @__PURE__ */ (0, q.jsx)(E, {
					icon: rl(e),
					className: "mt-px shrink-0 text-[11px]"
				})]
			}),
			/* @__PURE__ */ (0, q.jsx)("div", {
				className: "min-w-0",
				children: /* @__PURE__ */ (0, q.jsx)("div", {
					className: "flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-0.5",
					children: s ? /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [/* @__PURE__ */ (0, q.jsx)("span", {
						className: "min-w-0 break-all font-mono font-semibold leading-4 opacity-70",
						children: e.functionName
					}), e.annotationText && /* @__PURE__ */ (0, q.jsx)("span", {
						className: "min-w-0 break-all font-mono text-[11px] font-normal opacity-80",
						children: e.annotationText
					})] }) : /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
						i ? i("method", e, t) : /* @__PURE__ */ (0, q.jsx)("span", {
							className: "min-w-0 break-all font-mono font-semibold leading-4",
							children: c
						}),
						e.class && (i ? i("class", e, t) : /* @__PURE__ */ (0, q.jsx)("span", {
							className: "min-w-0 break-all font-mono text-[11px] text-muted-foreground",
							children: e.class
						})),
						e.location && (i ? i("location", e, t) : /* @__PURE__ */ (0, q.jsx)("span", {
							className: "rounded border border-border bg-muted/40 px-1.5 py-px font-mono text-[10px] text-muted-foreground",
							children: e.location
						}))
					] })
				})
			}),
			l ? /* @__PURE__ */ (0, q.jsx)("div", {
				className: ["flex shrink-0 items-start gap-1", o === "hover" ? "opacity-0 transition-opacity focus-within:opacity-100 group-hover:opacity-100 [@media(hover:none)]:opacity-100" : "opacity-100"].join(" "),
				children: l
			}) : /* @__PURE__ */ (0, q.jsx)("div", {}),
			!s && tl(e) ? /* @__PURE__ */ (0, q.jsx)("div", {
				className: "col-span-3",
				children: /* @__PURE__ */ (0, q.jsx)(el, { frame: e })
			}) : null,
			u ? /* @__PURE__ */ (0, q.jsx)("div", {
				className: "col-start-2 col-span-2 min-w-0",
				children: u
			}) : null
		]
	});
}
function rl(e) {
	switch (e.kind) {
		case "locked": return P;
		case "waiting_to_lock": return lt;
		case "waiting_on":
		case "parking": return Et;
	}
	return e.nativeMethod ? Ct : e.runtime ? at : e.method || e.class ? _t : it;
}
// ../../../clicky-ui/packages/ui/src/lib/color.ts
var il = /^(?:bg|text|border|fill|stroke)-/;
function al(e) {
	return il.test(e.trim());
}
// ../../../clicky-ui/packages/ui/src/data/ProgressBar.tsx
function ol({ segments: e, total: t, label: n, formatValue: r = String, height: a = "h-2", className: o }) {
	if (t === 0) return null;
	let s = e.filter((e) => e.count > 0).map((e) => `${r(e.count)} ${e.label}`).join(", ");
	return /* @__PURE__ */ (0, q.jsx)("div", {
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
			let o = e.count / t * 100, s = al(e.color);
			return /* @__PURE__ */ (0, q.jsx)("div", {
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
// ../../../clicky-ui/packages/ui/src/data/diagnostics/MetricTile.tsx
var sl = {
	neutral: "bg-muted-foreground/40",
	success: "bg-green-500",
	warning: "bg-yellow-500",
	danger: "bg-red-500",
	info: "bg-blue-500"
}, cl = {
	neutral: "text-foreground",
	success: "text-green-600 dark:text-green-400",
	warning: "text-yellow-600 dark:text-yellow-400",
	danger: "text-red-600 dark:text-red-400",
	info: "text-blue-600 dark:text-blue-400"
};
function ll({ label: e, value: t, barPct: n, tone: r = "neutral", className: a }) {
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i("flex min-w-0 flex-col gap-0.5 rounded-md border border-border bg-background px-density-2 py-1", a),
		"data-metric": e,
		children: [
			/* @__PURE__ */ (0, q.jsx)("span", {
				className: "text-[9px] uppercase leading-none tracking-wide text-muted-foreground",
				children: e
			}),
			/* @__PURE__ */ (0, q.jsx)("span", {
				className: i("truncate text-xs font-semibold tabular-nums", cl[r]),
				children: t
			}),
			n !== void 0 && /* @__PURE__ */ (0, q.jsx)("div", {
				className: "h-0.5 w-full overflow-hidden rounded-full bg-muted",
				children: /* @__PURE__ */ (0, q.jsx)("div", {
					className: i("h-full transition-all duration-300", sl[r]),
					style: { width: `${n}%` },
					"data-metric-bar": ""
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/OpenInFilesLink.tsx
function ul({ target: e, onOpenFile: t, className: n }) {
	return /* @__PURE__ */ (0, q.jsxs)(a, {
		type: "button",
		variant: "link",
		size: "sm",
		className: i("h-auto gap-1 p-0 text-xs", n),
		title: `Open ${e.file}${e.line ? `:${e.line}` : ""} in Files`,
		onClick: (n) => {
			n.stopPropagation(), t(e);
		},
		children: [/* @__PURE__ */ (0, q.jsx)(E, {
			icon: ft,
			className: "text-xs"
		}), "Open in Files"]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/copy.ts
var dl = /* @__PURE__ */ new Set(["stdout", "stderr"]);
function fl(e, t) {
	return t ? [{
		label: e,
		value: t
	}] : [];
}
function pl(e, t, n) {
	return t ? [n ? {
		title: e,
		lang: n,
		body: t
	} : {
		title: e,
		body: t
	}] : [];
}
function ml(e) {
	if (e.file) return e.line ? `${e.file}:${e.line}` : e.file;
}
function hl(e) {
	let t = e.failure_detail?.summary;
	return [e.message, t === e.message ? void 0 : t].filter(Boolean).join("\n");
}
function gl(e) {
	let t = e.failure_detail;
	return t ? [
		t.matcher && `matcher: ${t.matcher}`,
		t.expected && `expected: ${t.expected}`,
		t.actual && `actual: ${t.actual}`,
		t.location && `at: ${t.location}`
	].filter(Boolean).join("\n") : "";
}
function _l(e) {
	let t = e.duration ? ze(e.duration) : void 0;
	return {
		facts: [
			...fl("Result", we(e) ?? void 0),
			...fl("Tool", e.framework),
			...fl("Location", ml(e)),
			...fl("Duration", t)
		],
		brief: [
			...pl("Error", hl(e)),
			...pl("Expectation", gl(e)),
			...pl("Stack", e.failure_detail?.stack),
			...pl("stdout", e.stdout),
			...pl("stderr", e.stderr)
		],
		items: []
	};
}
function vl(e, t) {
	let n = /* @__PURE__ */ new Set();
	return e.filter((e) => {
		let r = t(e);
		return !n.has(r) && (n.add(r), !0);
	});
}
function yl(e) {
	return {
		facts: vl(e.flatMap((e) => e.facts ?? []), (e) => e.label),
		brief: vl(e.flatMap((e) => e.brief ?? []), (e) => e.title),
		items: vl(e.flatMap((e) => e.items ?? []), (e) => e.label)
	};
}
function bl(e) {
	let t = Math.max(0, ...(e.match(/`+/g) ?? []).map((e) => e.length));
	return "`".repeat(Math.max(3, t + 1));
}
function xl(e, t, n) {
	let r = [`# ${e.name}`];
	n.length > 0 && r.push("", ...n.map((e) => `- **${e.label}:** ${e.value}`));
	for (let e of t) {
		if (!e.body) continue;
		if (r.push("", `## ${e.title}`, ""), e.markdown) {
			r.push(e.body);
			continue;
		}
		let t = bl(e.body);
		r.push(`${t}${e.lang ?? ""}`, e.body, t);
	}
	return r.join("\n");
}
function Sl(e, t) {
	let n = _l(e), r = yl([
		{
			facts: n.facts,
			brief: n.brief.filter((e) => !dl.has(e.title))
		},
		t,
		{ brief: n.brief.filter((e) => dl.has(e.title)) }
	]);
	return xl(e, r.brief, r.facts);
}
function Cl(e) {
	let { children: t, ...n } = e;
	return JSON.stringify(n, null, 2);
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/attempts.ts
function wl(e) {
	return e.timed_out ? "timedout" : e.failed ? "failed" : e.skipped ? "skipped" : e.passed ? "passed" : e.pending ? "pending" : "running";
}
function Tl(e) {
	let t = wl(e);
	return t === "running" || t === "pending";
}
function El(e, t, n) {
	let { passed: r, failed: i, skipped: a, warned: o, aborted: s, pending: c, running: l, timed_out: u, failure_detail: d, message: f, stdout: p, stderr: m, duration: h, ...g } = e, _ = wl(t);
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
function Dl(e) {
	let t = e.attempts ?? [];
	return t.map((n, r) => ({
		id: r + 1,
		status: wl(n),
		...n.duration !== void 0 && n.duration > 0 ? { durationMs: n.duration / 1e6 } : {},
		node: El(e, n, r === t.length - 1)
	}));
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/adapter.ts
var Ol = class {
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
		return n && n.length > 0 ? n : Dl(e);
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
		return yl(this.adapters.flatMap((r) => r.copy && r.match(e, t) ? [r.copy(n)] : []));
	}
	ownsScroll(e, t) {
		let n = this.resolveDetail(e, t);
		return n?.ownsScroll ? n.ownsScroll({
			node: e,
			runner: t
		}) : !1;
	}
};
function kl(e = []) {
	return new Ol(e);
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/context.ts
var Al = (0, K.createContext)(null), jl = Al.Provider;
function Ml() {
	let e = (0, K.useContext)(Al);
	if (!e) throw Error("useTestRunner must be used within a <TestRunner> / TestRunnerProvider");
	return e;
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestTreeNode.tsx
var Nl = {
	failed: "text-red-600 dark:text-red-400",
	timedout: "text-amber-600 dark:text-amber-400",
	warned: "text-amber-600 dark:text-amber-400",
	aborted: "text-red-500 dark:text-red-300",
	running: "text-blue-600 dark:text-blue-400",
	skipped: "text-muted-foreground",
	pending: "text-muted-foreground",
	passed: "text-foreground"
}, Pl = /* @__PURE__ */ new Set([
	"failed",
	"timedout",
	"warned",
	"aborted",
	"skipped"
]);
function Fl({ node: e }) {
	let t = _e(e);
	return t.total === 0 ? null : /* @__PURE__ */ (0, q.jsxs)("span", {
		className: "flex shrink-0 items-center gap-1",
		children: [
			t.passed > 0 && /* @__PURE__ */ (0, q.jsx)(L, {
				tone: "success",
				variant: "solid",
				size: "xxs",
				count: t.passed
			}),
			t.failed > 0 && /* @__PURE__ */ (0, q.jsx)(L, {
				tone: "danger",
				variant: "solid",
				size: "xxs",
				count: t.failed
			}),
			t.warned > 0 && /* @__PURE__ */ (0, q.jsx)(L, {
				tone: "warning",
				variant: "outline",
				size: "xxs",
				count: t.warned
			}),
			t.timedout > 0 && /* @__PURE__ */ (0, q.jsx)(L, {
				tone: "warning",
				variant: "solid",
				size: "xxs",
				count: t.timedout
			}),
			t.aborted > 0 && /* @__PURE__ */ (0, q.jsx)(L, {
				tone: "danger",
				variant: "soft",
				size: "xxs",
				count: t.aborted
			}),
			t.skipped > 0 && /* @__PURE__ */ (0, q.jsx)(L, {
				tone: "neutral",
				variant: "soft",
				size: "xxs",
				count: t.skipped
			}),
			t.running > 0 && /* @__PURE__ */ (0, q.jsx)(L, {
				tone: "info",
				variant: "solid",
				size: "xxs",
				count: t.running
			}),
			t.pending > 0 && /* @__PURE__ */ (0, q.jsx)(L, {
				tone: "neutral",
				variant: "soft",
				size: "xxs",
				count: t.pending
			})
		]
	});
}
function Il({ node: e, selected: t }) {
	let n = Ml(), r = (e.children?.length ?? 0) > 0, a = Le(e), o = we(e), s = n.adapters.resolveFrameworkIcon(e, n)?.renderFrameworkIcon?.({
		node: e,
		runner: n
	}), c = n.adapters.resolveRowLeading(e, n), l = e.duration || (r ? Te(e) : 0), d = !!n.onRerun && e.framework !== "task", f = !!n.onStop && !!e.can_stop && !!e.task_id, p = a && (Pl.has(a) || e.framework === "task" && (a === "running" || a === "pending")) ? (e.message ?? "").trim() : "";
	return /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
		o ? /* @__PURE__ */ (0, q.jsx)(G, { outcome: o }) : /* @__PURE__ */ (0, q.jsx)(E, {
			icon: gt,
			className: "shrink-0 text-muted-foreground"
		}),
		s == null ? /* @__PURE__ */ (0, q.jsx)(ke, {
			tool: Ae(e),
			className: "text-sm"
		}) : /* @__PURE__ */ (0, q.jsx)("span", {
			className: "flex shrink-0 items-center text-sm",
			children: s
		}),
		c?.renderRowLeading?.({
			node: e,
			runner: n
		}),
		/* @__PURE__ */ (0, q.jsx)("span", {
			className: i("min-w-0 truncate", t ? "font-semibold" : "font-medium", a && Nl[a] || "text-foreground"),
			title: p || void 0,
			children: Re(e.name, e.framework)
		}),
		p ? /* @__PURE__ */ (0, q.jsx)("span", {
			className: "min-w-0 flex-1 shrink-[100] truncate text-xs text-muted-foreground",
			title: p,
			"data-row-message": "",
			children: p
		}) : /* @__PURE__ */ (0, q.jsx)("span", { className: "flex-1" }),
		e.progress && e.progress.total > 0 && !e.passed && !e.failed && /* @__PURE__ */ (0, q.jsxs)("span", {
			className: "flex shrink-0 items-center gap-1 text-xs text-blue-600 dark:text-blue-400",
			children: [e.progress.phase && /* @__PURE__ */ (0, q.jsx)("span", { children: e.progress.phase }), /* @__PURE__ */ (0, q.jsxs)("span", {
				className: "tabular-nums",
				children: [
					e.progress.done,
					"/",
					e.progress.total
				]
			})]
		}),
		l > 0 && /* @__PURE__ */ (0, q.jsx)("span", {
			className: "shrink-0 text-xs text-muted-foreground",
			children: ze(l)
		}),
		r && /* @__PURE__ */ (0, q.jsx)(Fl, { node: e }),
		f && /* @__PURE__ */ (0, q.jsx)(go, {
			icon: u,
			label: "Stop",
			disabled: n.busy.stop,
			onClick: (t) => {
				t.stopPropagation(), n.onStop(e);
			}
		}),
		d && /* @__PURE__ */ (0, q.jsx)(go, {
			icon: wt,
			label: "Rerun",
			disabled: n.busy.rerun,
			onClick: (t) => {
				t.stopPropagation(), n.onRerun(e);
			}
		})
	] });
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestTree.tsx
var Ll = (() => {
	let e = /* @__PURE__ */ new WeakMap(), t = 0;
	return (n) => {
		if (n.route_path) return n.route_path;
		let r = e.get(n);
		return r || (r = `tn-${t++}-${n.name}`, e.set(n, r)), r;
	};
})();
function Rl({ className: e, actions: t }) {
	let n = Ml();
	return /* @__PURE__ */ (0, q.jsx)(ve, {
		className: i("min-h-0", e),
		roots: n.tests,
		getKey: Ll,
		getChildren: (e) => e.children,
		selected: n.selected,
		onSelect: (e) => n.onSelect(e === n.selected ? null : e),
		expandAll: n.expandAll,
		onExpandAllChange: n.onExpandAllChange,
		defaultOpen: (e, t) => t < 1 || Oe(e),
		showSearch: !0,
		searchLabel: "Filter tests",
		getSearchText: (e) => `${e.name} ${e.framework ?? ""}`,
		renderRow: ({ node: e, selected: t }) => /* @__PURE__ */ (0, q.jsx)(Il, {
			node: e,
			selected: t
		}),
		rowClass: (e, t) => i("hover:bg-accent", Le(e) === "running" && "border-l-2 border-primary bg-primary/10 font-semibold hover:bg-primary/15", t && "border-l-2 border-primary bg-primary/20"),
		actions: t
	});
}
// ../../../clicky-ui/packages/ui/src/components/SplitButton.tsx
function zl({ label: e, onClick: t, items: n, icon: r, variant: o = "default", size: s = "default", loading: c = !1, disabled: l = !1, primaryDisabled: u = !1, align: d = "right", title: f, className: p, menuClassName: m }) {
	let [h, g] = (0, K.useState)(!1);
	return n.length === 0 ? /* @__PURE__ */ (0, q.jsxs)(a, {
		variant: o,
		size: s,
		loading: c,
		disabled: l || u,
		onClick: t,
		className: p,
		children: [r && /* @__PURE__ */ (0, q.jsx)(E, { ...typeof r == "string" ? { name: r } : { icon: r } }), e]
	}) : /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i("inline-flex items-stretch", p),
		children: [/* @__PURE__ */ (0, q.jsxs)(a, {
			variant: o,
			size: s,
			loading: c,
			disabled: l || u,
			onClick: t,
			className: "rounded-r-none",
			children: [r && /* @__PURE__ */ (0, q.jsx)(E, { ...typeof r == "string" ? { name: r } : { icon: r } }), e]
		}), /* @__PURE__ */ (0, q.jsx)(fe, {
			items: n,
			align: d,
			className: "items-stretch",
			...m ? { menuClassName: m } : {},
			onOpenChange: g,
			trigger: /* @__PURE__ */ (0, q.jsx)(a, {
				variant: o,
				size: s === "default" ? "icon" : s,
				disabled: l,
				title: f,
				"aria-label": f ?? "Open menu",
				className: i("h-auto self-stretch rounded-l-none border-l border-l-current/25 px-2", s !== "icon" && "w-8"),
				children: /* @__PURE__ */ (0, q.jsx)(E, { icon: h ? H : F })
			})
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/detailTabContext.ts
var Bl = (0, K.createContext)(null), Vl = Bl.Provider;
function Hl() {
	return (0, K.useContext)(Bl);
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/jvm-stacktrace.ts
var Ul = /^"(?<name>[^"]*)"(?<rest>.*)$/, Wl = /^\s*java\.lang\.Thread\.State:\s+(?<state>[A-Z_]+)(?:\s+\((?<sub>[^)]+)\))?/, Gl = /^\s*at\s+(?<fn>[^\s(]+)\((?<src>[^)]+)\)\s*$/, Kl = /^(?<file>[^:]+):(?<line>\d+)$/, ql = /^\s*-\s+(?<kind>locked|waiting on|waiting to lock|parking to wait for)\b(?<rest>.*)$/;
function Jl(e) {
	let t = e.trim();
	if (!t) return [];
	let n = Yl(t), r = [];
	for (let e of n) {
		let t = Xl(e);
		t && r.push(t);
	}
	return r;
}
function Yl(e) {
	let t = [], n = [], r = e.split("\n");
	for (let e of r) {
		let r = e.replace(/\r$/, "");
		r.startsWith("\"") && n.length > 0 ? (t.push(n.join("\n")), n = [r]) : n.push(r);
	}
	return n.length > 0 && t.push(n.join("\n")), t;
}
function Xl(e) {
	let t = e.split("\n"), n = t[0]?.trim() ?? "", r = Ul.exec(n);
	if (!r?.groups) return null;
	let i = r.groups.name;
	if (!i) return null;
	let a = r.groups.rest ?? "", o = /#(\d+)\b/.exec(a), s = /\bprio=(\d+)/.exec(a), c = /\bnid=(0x[0-9a-f]+)/i.exec(a), l = /\bdaemon\b/.test(a), u = Zl(a) ?? "", d = iu(u), f = [];
	for (let e = 1; e < t.length; e++) {
		let n = t[e];
		if (n == null || !n.trim()) continue;
		let r = Wl.exec(n);
		if (r?.groups) {
			let e = r.groups.state;
			if (!e) continue;
			let t = r.groups.sub;
			u = t ? `${e} (${t})` : e, d = iu(e);
			continue;
		}
		let i = Gl.exec(n);
		if (i?.groups) {
			let e = i.groups.fn, t = i.groups.src;
			if (!e || !t) continue;
			let n = {
				functionName: e,
				displayName: su(e),
				kind: "frame",
				runtime: au(e),
				nativeMethod: t === "Native Method",
				...ou(e)
			}, r = Kl.exec(t);
			r?.groups?.file && r.groups.line ? (n.file = r.groups.file, n.line = Number(r.groups.line), n.location = `${n.file}:${n.line}`) : n.location = t === "Native Method" ? "Native Method" : t, f.push(n);
			continue;
		}
		let a = ql.exec(n);
		if (a?.groups?.kind) {
			let e = a.groups.kind, t = nu(e);
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
		id: o ? Number(o[1]) : tu(i, a),
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
function Zl(e) {
	let t = e.toLowerCase(), n = t.indexOf("nid=0x");
	if (n === -1) return;
	let r = n + 6, i = r;
	for (; Ql(e[r]);) r++;
	if (r === i || !eu(e[r])) return;
	for (; eu(e[r]);) r++;
	let a = e.length;
	for (; a > r && eu(e[a - 1]);) a--;
	if (e[a - 1] === "]") {
		let n = t.lastIndexOf("[0x", a - 1);
		if (n > r && $l(e, n + 3, a - 1) && eu(e[n - 1])) for (a = n; a > r && eu(e[a - 1]);) a--;
	}
	return e.slice(r, a) || void 0;
}
function Ql(e) {
	if (e === void 0) return !1;
	let t = e.toLowerCase().charCodeAt(0);
	return t >= 48 && t <= 57 || t >= 97 && t <= 102;
}
function $l(e, t, n) {
	if (t >= n) return !1;
	for (let r = t; r < n; r++) if (!Ql(e[r])) return !1;
	return !0;
}
function eu(e) {
	return e !== void 0 && e.trim() === "";
}
function tu(e, t) {
	let n = /\btid=(0x[0-9a-f]+)/i.exec(t);
	if (n?.[1]) {
		let e = n[1].slice(2), t = Number.parseInt(e.slice(-8), 16);
		if (Number.isFinite(t)) return t;
	}
	let r = 0;
	for (let t = 0; t < e.length; t++) r = r * 31 + e.charCodeAt(t) | 0;
	return Math.abs(r);
}
function nu(e) {
	switch (e) {
		case "locked": return "locked";
		case "waiting on": return "waiting_on";
		case "waiting to lock": return "waiting_to_lock";
		case "parking to wait for": return "parking";
		default: return "frame";
	}
}
function ru(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.state, (t.get(n.state) ?? 0) + 1);
	return t;
}
function iu(e) {
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
function au(e) {
	return Zc(e);
}
function ou(e) {
	let t = e.lastIndexOf(".");
	return t <= 0 ? {} : {
		class: e.slice(0, t),
		method: e.slice(t + 1)
	};
}
function su(e) {
	let t = e.lastIndexOf(".");
	return t <= 0 ? e : Qc(e.slice(0, t), e.slice(t + 1));
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/stacktrace.ts
function cu(e) {
	let t = e.trim();
	return t ? /^"[^"]*"[^\n]*\b(prio=|tid=|nid=)/m.test(t) ? "jvm" : /^goroutine\s+\d+\s+\[[^\]]+\]:/m.test(t) ? "go" : "unknown" : "unknown";
}
function lu(e) {
	let t = cu(e);
	return t === "jvm" ? {
		format: "jvm",
		threads: Jl(e)
	} : t === "go" ? {
		format: "go",
		goroutines: hu(e)
	} : {
		format: "unknown",
		goroutines: [],
		threads: []
	};
}
function uu(e) {
	return e.format === "jvm" ? ru(e.threads) : e.format === "go" ? vu(e.goroutines) : /* @__PURE__ */ new Map();
}
var du = /^goroutine\s+(\d+)\s+\[(.+?)\]:$/, fu = /^\/usr\/local\/go[\d.]+\/src\//, pu = /^.*?\/go\/src\//, mu = /^.*?\/go\/pkg\/mod\//;
function hu(e) {
	let t = e.trim();
	if (!t) return [];
	let n = wu(t), r = [];
	for (let e of n) {
		let t = e.split("\n").map((e) => e.replace(/\r$/, "")), n = t[0]?.trim(), i = du.exec(n || "");
		if (!i) continue;
		let a = Number(i[1] ?? 0), o = i[2] ?? "", s = yu(o), c = [];
		for (let e = 1; e < t.length; e++) {
			let n = t[e];
			if (n == null) continue;
			let r = n.trim();
			if (!r) continue;
			let i = xu(r), a = c[c.length - 1];
			if (i && a && !a.file) {
				a.file = i.file, a.line = i.line;
				continue;
			}
			let o = r.startsWith("created by ") ? "created_by" : "frame", s = o === "created_by" ? r.slice(11) : r;
			c.push({
				functionName: s,
				displayName: Tu(s),
				kind: o,
				runtime: bu(s)
			});
		}
		for (let e of c) e.file && (e.file = Du(e.file), e.location = `${e.file}${e.line ? `:${e.line}` : ""}`);
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
function gu(e, t) {
	let n = e.file ? Du(e.file) : void 0;
	return {
		functionName: e.func,
		displayName: Eu(e.func),
		kind: t,
		runtime: bu(e.func),
		...n ? {
			file: n,
			line: e.line,
			location: `${n}${e.line ? `:${e.line}` : ""}`
		} : {}
	};
}
function _u(e) {
	return e.map((e) => {
		if (e.format !== "go" || e.error || e.goroutineId === void 0) throw Error(`stack record ${e.checkpointId}/${e.key} is not a Go goroutine record (format "${e.format}"${e.error ? `, error "${e.error}"` : ""})`);
		let t = [...(e.frames ?? []).map((e) => gu(e, "frame")), ...e.createdBy ? [gu(e.createdBy, "created_by")] : []], n = e.state ?? "", r = [
			n,
			...e.waitMinutes ? [`${e.waitMinutes} minutes`] : [],
			...e.lockedToThread ? ["locked to thread"] : []
		].join(", "), i = t.find((e) => e.kind === "frame")?.functionName;
		return {
			id: e.goroutineId,
			state: yu(n),
			rawState: r,
			frames: t,
			raw: e.raw,
			userFrameCount: t.filter((e) => !e.runtime && e.kind === "frame").length,
			searchText: `goroutine ${e.goroutineId} [${r}]\n${t.map((e) => `${e.functionName} ${e.file || ""}`).join("\n")}`.toLowerCase(),
			...i === void 0 ? {} : { topFunction: i }
		};
	});
}
function vu(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.state, (t.get(n.state) || 0) + 1);
	return t;
}
function yu(e) {
	return (e.split(",")[0] ?? "").trim().toLowerCase();
}
function bu(e) {
	return e.startsWith("runtime.") || e.startsWith("runtime/") || e.startsWith("internal/") || e.startsWith("runtime/internal/") || e.startsWith("syscall.") || e.startsWith("reflect.");
}
function xu(e) {
	let t = e.lastIndexOf(":");
	if (t <= 0) return;
	let n = t + 1, r = n;
	for (; n < e.length && Su(e[n] ?? "");) n++;
	if (n === r) return;
	let i = n;
	if (n < e.length) {
		let t = n;
		for (; n < e.length && (e[n] ?? "").trim() === "";) n++;
		if (n === t || e.slice(n, n + 3).toLowerCase() !== "+0x") return;
		n += 3;
		let r = n;
		for (; n < e.length && Cu(e[n] ?? "");) n++;
		if (n === r || n !== e.length) return;
	}
	return {
		file: e.slice(0, t),
		line: Number(e.slice(r, i))
	};
}
function Su(e) {
	return e >= "0" && e <= "9";
}
function Cu(e) {
	let t = e.toLowerCase();
	return Su(e) || t >= "a" && t <= "f";
}
function wu(e) {
	let t = [], n = [];
	for (let r of e.split("\n")) r.trim() ? n.push(r) : n.length > 0 && (t.push(n.join("\n")), n = []);
	return n.length > 0 && t.push(n.join("\n")), t;
}
function Tu(e) {
	let t = e.trim(), n = t.indexOf("(");
	return n !== -1 && (t = t.slice(0, n)), Eu(t);
}
function Eu(e) {
	return e.replace(/^((?:[^./\s]+\/)+)([^./\s]+)\./, "$2.");
}
function Du(e) {
	return e.replace(fu, "").replace(pu, "").replace(mu, "");
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/state-styles.ts
function Ou(e) {
	return e.includes("running") ? "bg-green-50 text-green-700 dark:bg-green-500/20 dark:text-green-300" : e.includes("chan") || e.includes("wait") ? "bg-blue-50 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300" : e.includes("sleep") ? "bg-amber-50 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300" : e.includes("select") ? "bg-violet-50 text-violet-700 dark:bg-violet-500/20 dark:text-violet-300" : "bg-muted text-muted-foreground";
}
function ku(e) {
	return e.includes("running") ? "bg-green-500" : e.includes("chan") || e.includes("wait") ? "bg-blue-500" : e.includes("sleep") ? "bg-amber-500" : e.includes("select") ? "bg-violet-500" : "bg-muted-foreground/40";
}
function Au(e) {
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
function ju(e) {
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
// ../../../clicky-ui/packages/ui/src/data/diagnostics/GoroutineCard.tsx
function Mu({ goroutine: e, search: t, hideRuntimeOnly: n }) {
	let r = n ? e.frames.filter((e) => !e.runtime || e.kind === "created_by") : e.frames, i = e.state === "running" || !!t;
	return /* @__PURE__ */ (0, q.jsxs)("details", {
		className: "border-0 bg-transparent",
		open: i,
		children: [/* @__PURE__ */ (0, q.jsx)("summary", {
			className: "cursor-pointer list-none px-0 py-1",
			children: /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "flex items-center gap-2 flex-wrap",
				children: [
					/* @__PURE__ */ (0, q.jsxs)("span", {
						className: "font-mono text-xs font-semibold text-foreground",
						children: ["g", e.id]
					}),
					/* @__PURE__ */ (0, q.jsxs)("span", {
						className: `inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] ${Ou(e.state)}`,
						children: [/* @__PURE__ */ (0, q.jsx)("span", { className: `h-2 w-2 rounded-full ${ku(e.state)}` }), e.rawState]
					}),
					/* @__PURE__ */ (0, q.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [r.length, "f"]
					}),
					e.userFrameCount > 0 && /* @__PURE__ */ (0, q.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [e.userFrameCount, "u"]
					}),
					e.topFunction && /* @__PURE__ */ (0, q.jsx)("span", {
						className: "truncate text-[11px] text-muted-foreground",
						children: e.topFunction
					})
				]
			})
		}), /* @__PURE__ */ (0, q.jsx)("div", {
			className: "pl-3 py-1 space-y-0.5",
			children: r.map((t, n) => /* @__PURE__ */ (0, q.jsx)(Nu, { frame: t }, `${e.id}-${n}`))
		})]
	});
}
function Nu({ frame: e }) {
	return /* @__PURE__ */ (0, q.jsx)("div", {
		className: e.runtime ? "text-muted-foreground" : "text-foreground",
		children: /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "flex items-start gap-1.5",
			children: [/* @__PURE__ */ (0, q.jsx)(E, {
				icon: e.kind === "created_by" ? R : e.runtime ? at : _t,
				className: "shrink-0 mt-0.5 text-[11px]"
			}), /* @__PURE__ */ (0, q.jsx)("div", {
				className: "min-w-0",
				children: /* @__PURE__ */ (0, q.jsxs)("div", {
					className: "break-all font-mono text-[11px] font-semibold leading-4",
					children: [e.displayName, e.location && /* @__PURE__ */ (0, q.jsx)("span", {
						className: "ml-2 text-[10px] font-normal opacity-80",
						children: e.location
					})]
				})
			})]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/JvmStackTrace.tsx
function Pu({ frames: e, hideRuntimeOnly: t = !1, className: n, resolveSource: r, frameActions: i }) {
	let a = t ? e.filter((e) => e.kind !== "frame" || !e.runtime) : e;
	return a.length === 0 ? null : /* @__PURE__ */ (0, q.jsx)("div", {
		className: n,
		children: a.map((e, t) => {
			let n = r?.(e), { line: a, ...o } = n ?? {}, s = n ? {
				...e,
				...o
			} : e;
			return /* @__PURE__ */ (0, q.jsx)(nl, {
				frame: s,
				index: t,
				...i ? { frameActions: i } : {}
			}, `${e.functionName}-${t}`);
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/ThreadCard.tsx
function Fu({ thread: e, search: t, hideRuntimeOnly: n, resolveSource: r, frameActions: i }) {
	let a = n ? e.frames.filter((e) => e.kind !== "frame" || !e.runtime) : e.frames, o = e.state === "runnable" || e.state === "blocked" || !!t;
	return /* @__PURE__ */ (0, q.jsxs)("details", {
		className: "border-0 bg-transparent",
		open: o,
		children: [/* @__PURE__ */ (0, q.jsx)("summary", {
			className: "cursor-pointer list-none px-0 py-1",
			children: /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "flex items-center gap-2 flex-wrap",
				children: [
					/* @__PURE__ */ (0, q.jsxs)("span", {
						className: "font-mono text-xs font-semibold text-foreground",
						children: ["#", e.id]
					}),
					/* @__PURE__ */ (0, q.jsx)("span", {
						className: "truncate font-mono text-xs text-foreground",
						children: e.name
					}),
					/* @__PURE__ */ (0, q.jsxs)("span", {
						className: `inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] ${Au(e.state)}`,
						children: [/* @__PURE__ */ (0, q.jsx)("span", { className: `h-2 w-2 rounded-full ${ju(e.state)}` }), e.rawState || e.state]
					}),
					e.daemon && /* @__PURE__ */ (0, q.jsx)("span", {
						className: "rounded-full bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground",
						children: "daemon"
					}),
					/* @__PURE__ */ (0, q.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [a.length, "f"]
					}),
					e.userFrameCount > 0 && /* @__PURE__ */ (0, q.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [e.userFrameCount, "u"]
					}),
					e.topFunction && /* @__PURE__ */ (0, q.jsx)("span", {
						className: "truncate text-[11px] text-muted-foreground",
						children: e.topFunction
					})
				]
			})
		}), /* @__PURE__ */ (0, q.jsx)(Pu, {
			frames: a,
			className: "space-y-0.5 py-1 pl-3",
			...r ? { resolveSource: r } : {},
			...i ? { frameActions: i } : {}
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/StackDumpView.tsx
var Iu = "10rem", Lu = {
	format: "unknown",
	goroutines: [],
	threads: []
};
function Ru(e) {
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
			goroutines: _u(r)
		} : Lu,
		text: i.map((e) => e.raw).join("\n"),
		...t.length > 0 ? { error: t.map((e) => e.error).join("\n") } : {},
		...a ? {
			collectedAt: a.takenAt,
			reason: a.reason
		} : {}
	};
}
function zu(e, t) {
	if (t) return Ru(t);
	if (e === void 0) throw Error("StackDumpView needs either text or records");
	return {
		parsed: lu(e),
		text: e
	};
}
function Bu(e) {
	return e.format === "jvm" ? e.threads : e.format === "go" ? e.goroutines : [];
}
function Vu(e) {
	return e.format === "jvm" ? "threads" : e.format === "go" ? "goroutines" : "frames";
}
function Hu(e, t) {
	return e.format === "jvm" ? ju(t) : ku(t);
}
function Uu({ status: e }) {
	return /* @__PURE__ */ (0, q.jsx)("span", {
		className: `rounded-full px-2 py-0.5 ${e === "ready" ? "bg-green-100 text-green-700 dark:bg-green-500/20 dark:text-green-300" : e === "unsupported" ? "bg-yellow-100 text-yellow-700 dark:bg-yellow-500/20 dark:text-yellow-300" : "bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300"}`,
		children: e
	});
}
function Wu({ text: e, records: t, error: n, status: r, collectedAt: i, reason: a, singleItem: o = !1, frameActions: s, resolveSource: c, headerActions: l }) {
	let [u, d] = (0, K.useState)(""), [f, p] = (0, K.useState)(/* @__PURE__ */ new Set()), [m, h] = (0, K.useState)(!0), g = (0, K.useMemo)(() => zu(e, t), [e, t]), { parsed: _, text: v } = g, y = n ?? g.error, b = i ?? g.collectedAt, x = a ?? g.reason, S = t ? `${t[0]?.checkpointId}|${t.length}` : e, C = (0, K.useMemo)(() => uu(_), [_]), w = (0, K.useMemo)(() => Bu(_), [_]), T = (0, K.useMemo)(() => {
		if (o) return w;
		let e = u.trim().toLowerCase();
		return w.filter((t) => !(f.size > 0 && !f.has(t.state) || m && t.userFrameCount === 0 || e && !t.searchText.includes(e)));
	}, [
		w,
		u,
		f,
		m,
		o
	]);
	(0, K.useEffect)(() => {
		d(""), p(/* @__PURE__ */ new Set()), h(!0);
	}, [S]);
	let D = w.length, O = Vu(_);
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		"data-stack-format": _.format,
		children: [
			!o && /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "space-y-density-2 px-1 py-1.5",
				children: [/* @__PURE__ */ (0, q.jsxs)("div", {
					className: "flex flex-wrap items-center justify-between gap-density-2 text-[11px] text-muted-foreground",
					children: [/* @__PURE__ */ (0, q.jsxs)("div", {
						className: "flex flex-wrap items-center gap-density-2",
						children: [
							r ? /* @__PURE__ */ (0, q.jsx)(Uu, { status: r }) : null,
							b ? /* @__PURE__ */ (0, q.jsx)("span", { children: new Date(b).toLocaleString() }) : null,
							x ? /* @__PURE__ */ (0, q.jsx)("span", {
								"data-stack-reason": "",
								children: x
							}) : null,
							D > 0 && /* @__PURE__ */ (0, q.jsxs)("span", { children: [
								T.length,
								" / ",
								D,
								" ",
								O
							] })
						]
					}), l]
				}), D > 0 && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [/* @__PURE__ */ (0, q.jsxs)("div", {
					className: "flex flex-wrap items-center gap-1.5",
					children: [
						/* @__PURE__ */ (0, q.jsxs)("div", {
							className: "relative min-w-[14rem] flex-1",
							children: [/* @__PURE__ */ (0, q.jsx)(E, {
								icon: Se,
								className: "absolute left-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground"
							}), /* @__PURE__ */ (0, q.jsx)("input", {
								className: "w-full rounded-md border border-border bg-muted/50 py-1 pl-7 pr-2 text-xs outline-none focus:border-primary focus:bg-background",
								"aria-label": `Filter ${O}`,
								placeholder: _.format === "jvm" ? "Filter by thread name, function, or file" : "Filter by goroutine id, function, or file",
								value: u,
								onChange: (e) => d(e.target.value)
							})]
						}),
						/* @__PURE__ */ (0, q.jsxs)("label", {
							className: "inline-flex items-center gap-1.5 rounded-md border border-border bg-muted/50 px-2 py-1 text-[11px] text-muted-foreground",
							children: [/* @__PURE__ */ (0, q.jsx)("input", {
								type: "checkbox",
								checked: m,
								onChange: (e) => h(e.target.checked)
							}), "Hide runtime-only"]
						}),
						(u || f.size > 0 || !m) && /* @__PURE__ */ (0, q.jsx)("button", {
							type: "button",
							className: "text-[11px] text-muted-foreground hover:text-foreground",
							onClick: () => {
								d(""), p(/* @__PURE__ */ new Set()), h(!0);
							},
							children: "Clear"
						})
					]
				}), /* @__PURE__ */ (0, q.jsx)("div", {
					className: "flex flex-wrap items-center gap-1",
					children: Array.from(C.entries()).sort((e, t) => t[1] - e[1]).map(([e, t]) => {
						let n = f.has(e);
						return /* @__PURE__ */ (0, q.jsxs)("button", {
							type: "button",
							"aria-pressed": n,
							className: `inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] transition-colors ${n ? "border-primary bg-primary/10 text-primary" : "border-border bg-muted/50 text-muted-foreground hover:bg-background"}`,
							onClick: () => {
								p((t) => {
									let n = new Set(t);
									return n.has(e) ? n.delete(e) : n.add(e), n;
								});
							},
							children: [
								/* @__PURE__ */ (0, q.jsx)("span", { className: `h-2 w-2 rounded-full ${Hu(_, e)}` }),
								e,
								/* @__PURE__ */ (0, q.jsx)("span", {
									className: "text-[10px] opacity-70",
									children: t
								})
							]
						}, e);
					})
				})] })]
			}),
			o && l ? /* @__PURE__ */ (0, q.jsx)("div", {
				className: "flex justify-end px-1 py-1.5",
				children: l
			}) : null,
			y && /* @__PURE__ */ (0, q.jsx)("div", {
				className: "mt-density-2 whitespace-pre-wrap text-xs text-red-600",
				children: y
			}),
			D === 0 ? v ? /* @__PURE__ */ (0, q.jsx)("pre", {
				className: "min-h-0 flex-1 overflow-auto whitespace-pre-wrap py-1 font-mono text-[11px] leading-4 text-foreground",
				style: { minHeight: Iu },
				children: v
			}) : null : /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "min-h-0 flex-1 space-y-1 overflow-auto py-1",
				style: { minHeight: Iu },
				children: [
					T.length === 0 && /* @__PURE__ */ (0, q.jsxs)("div", {
						className: "py-density-3 text-center text-xs text-muted-foreground",
						children: [
							"No ",
							O,
							" match the current filters."
						]
					}),
					_.format === "jvm" && _.threads.filter((e) => T.includes(e)).map((e) => /* @__PURE__ */ (0, q.jsx)(Fu, {
						thread: e,
						search: u,
						hideRuntimeOnly: m,
						...c ? { resolveSource: c } : {},
						...s ? { frameActions: s } : {}
					}, e.id)),
					_.format === "go" && _.goroutines.filter((e) => T.includes(e)).map((e) => /* @__PURE__ */ (0, q.jsx)(Mu, {
						goroutine: e,
						search: u,
						hideRuntimeOnly: m
					}, e.id))
				]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/Expectation.tsx
function Gu(e) {
	return typeof e == "string" ? e : JSON.stringify(e);
}
function Ku(e) {
	return Array.isArray(e) ? e.some((e) => typeof e == "object" && !!e) : typeof e == "object" && !!e;
}
function qu({ value: e, tree: t = !1, className: n }) {
	return t && Ku(e) ? /* @__PURE__ */ (0, q.jsx)(v, { data: e }) : /* @__PURE__ */ (0, q.jsx)("code", {
		className: i("whitespace-pre-wrap break-words font-mono", n),
		children: Gu(e)
	});
}
function Ju({ label: e, children: t }) {
	return /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [/* @__PURE__ */ (0, q.jsx)("dt", {
		className: "text-muted-foreground",
		children: e
	}), /* @__PURE__ */ (0, q.jsx)("dd", {
		className: "min-w-0",
		children: t
	})] });
}
function Yu({ check: e, checkLabel: t = "matcher", expected: n, actual: r, vars: a, location: o, className: s }) {
	let c = Object.entries(a ?? {});
	return /* @__PURE__ */ (0, q.jsxs)("dl", {
		className: i("grid grid-cols-[auto_minmax(0,1fr)] items-baseline gap-x-3 gap-y-0.5 rounded border border-border bg-muted/30 px-2 py-1.5 text-[11px]", s),
		"data-expectation": "",
		children: [
			e && /* @__PURE__ */ (0, q.jsx)(Ju, {
				label: t,
				children: /* @__PURE__ */ (0, q.jsx)(qu, { value: e })
			}),
			c.map(([e, t]) => /* @__PURE__ */ (0, q.jsx)(Ju, {
				label: e,
				children: /* @__PURE__ */ (0, q.jsx)(qu, {
					value: t,
					tree: !0
				})
			}, e)),
			n !== void 0 && /* @__PURE__ */ (0, q.jsx)(Ju, {
				label: "expected",
				children: /* @__PURE__ */ (0, q.jsx)(qu, {
					value: n,
					className: "text-emerald-700 dark:text-emerald-400"
				})
			}),
			r !== void 0 && /* @__PURE__ */ (0, q.jsx)(Ju, {
				label: "actual",
				children: /* @__PURE__ */ (0, q.jsx)(qu, {
					value: r,
					className: "text-red-700 dark:text-red-400"
				})
			}),
			o && /* @__PURE__ */ (0, q.jsx)(Ju, {
				label: "at",
				children: /* @__PURE__ */ (0, q.jsx)(qu, {
					value: o,
					className: "text-muted-foreground"
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestFailureDetail.tsx
function Xu({ stack: e }) {
	let t = (0, K.useMemo)(() => lu(e), [e]);
	return t.format === "go" ? /* @__PURE__ */ (0, q.jsx)("div", {
		className: "flex max-h-[32rem] min-h-0 flex-col",
		"data-failure-stack": "go",
		children: /* @__PURE__ */ (0, q.jsx)(Wu, {
			text: e,
			singleItem: t.goroutines.length === 1
		})
	}) : /* @__PURE__ */ (0, q.jsx)(Kc, {
		logs: e,
		collapsedLines: 6
	});
}
function Zu({ failure: e }) {
	let t = !!(e.matcher || e.expected || e.actual || e.location);
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "space-y-2",
		children: [
			e.summary && /* @__PURE__ */ (0, q.jsx)("p", {
				className: "text-sm font-medium text-red-700 dark:text-red-300",
				children: e.summary
			}),
			t && /* @__PURE__ */ (0, q.jsx)(Yu, {
				check: e.matcher,
				expected: e.expected || void 0,
				actual: e.actual || void 0,
				location: e.location
			}),
			e.stack && /* @__PURE__ */ (0, q.jsx)(Xu, { stack: e.stack })
		]
	});
}
var Qu = "25cqh";
function $u({ node: e }) {
	return !e.stdout && !e.stderr ? null : /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "space-y-3",
		children: [e.stdout && /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "space-y-1",
			children: [/* @__PURE__ */ (0, q.jsx)("div", {
				className: "text-[10px] uppercase tracking-wide text-muted-foreground",
				children: "stdout"
			}), /* @__PURE__ */ (0, q.jsx)(Kc, {
				logs: e.stdout,
				variant: "terminal",
				maxHeight: Qu
			})]
		}), e.stderr && /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "space-y-1",
			children: [/* @__PURE__ */ (0, q.jsx)("div", {
				className: "text-[10px] uppercase tracking-wide text-muted-foreground",
				children: "stderr"
			}), /* @__PURE__ */ (0, q.jsx)(Kc, {
				logs: e.stderr,
				variant: "terminal",
				maxHeight: Qu,
				borderClass: "border-red-500/70"
			})]
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/metric-tone.ts
function ed(e) {
	return e >= 90 ? "danger" : e >= 60 ? "warning" : e > 0 ? "success" : "neutral";
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/process-metrics.ts
var td = class extends Error {
	pid;
	constructor(e) {
		super(`process ${e} not found`), this.name = "ProcessNotFoundError", this.pid = e;
	}
}, nd = 2e3;
function rd(e, t, n) {
	let [r, i] = (0, K.useState)({}), [a, o] = (0, K.useState)(!1), s = (0, K.useRef)(n);
	s.current = n;
	let c = t && e !== void 0 && e > 0 && n !== void 0;
	return (0, K.useEffect)(() => {
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
				if (e instanceof td) {
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
			n = setTimeout(a, nd);
		};
		return a(), () => {
			t = !0, clearTimeout(n), o(!1);
		};
	}, [c, e]), {
		state: r,
		polling: a
	};
}
function id({ attempt: e, fetchProcessMetrics: t }) {
	let n = Tl(e), [r, i] = (0, K.useState)(Date.now);
	(0, K.useEffect)(() => {
		if (!n || !e.started) return;
		let t = setInterval(() => i(Date.now()), 1e3);
		return () => clearInterval(t);
	}, [n, e.started]);
	let a = n && e.started ? Math.max(0, r - Date.parse(e.started)) * 1e6 : e.duration, { state: o, polling: s } = rd(e.pid, n, t), c = {
		...e.pid === void 0 ? {} : { pid: e.pid },
		...e.cpu_percent === void 0 ? {} : { cpu_percent: e.cpu_percent },
		...e.rss === void 0 ? {} : { rss: e.rss }
	}, l = n && o.metrics ? o.metrics : c, u = l.pid ?? e.pid, d = n ? o.note === "process exited" ? "exited" : l.status ?? "running" : "exited", f = (u ?? 0) > 0 || l.cpu_percent !== void 0 || l.rss !== void 0;
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "space-y-3",
		"data-attempt-sequence": e.sequence,
		children: [/* @__PURE__ */ (0, q.jsxs)("section", { children: [
			/* @__PURE__ */ (0, q.jsxs)("div", {
				className: "mb-1 flex items-center gap-2",
				children: [
					/* @__PURE__ */ (0, q.jsx)("div", {
						className: "text-[10px] uppercase tracking-wide text-muted-foreground",
						children: "Process metrics"
					}),
					s && /* @__PURE__ */ (0, q.jsx)("span", {
						className: "size-1.5 animate-pulse rounded-full bg-blue-500",
						role: "status",
						"aria-label": "Reading live metrics"
					}),
					e.shared_process && /* @__PURE__ */ (0, q.jsx)("span", {
						className: "rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground",
						title: "The pid, CPU and RSS belong to the go test batch this package ran in, not to the package alone",
						"data-shared-process": "",
						children: "shared by batch"
					})
				]
			}),
			f ? /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "grid grid-cols-2 gap-density-1 md:grid-cols-4",
				children: [
					u !== void 0 && u > 0 && /* @__PURE__ */ (0, q.jsx)(ll, {
						label: "PID",
						value: String(u)
					}),
					/* @__PURE__ */ (0, q.jsx)(ll, {
						label: "Status",
						value: d
					}),
					/* @__PURE__ */ (0, q.jsx)(ll, {
						label: "CPU",
						value: l.cpu_percent === void 0 ? "—" : `${l.cpu_percent.toFixed(1)}%`,
						...l.cpu_percent === void 0 ? {} : {
							barPct: Math.min(100, l.cpu_percent),
							tone: ed(l.cpu_percent)
						}
					}),
					/* @__PURE__ */ (0, q.jsx)(ll, {
						label: "RSS",
						value: l.rss !== void 0 && l.rss > 0 ? Co(l.rss) : "—"
					}),
					l.open_files !== void 0 && /* @__PURE__ */ (0, q.jsx)(ll, {
						label: "FDs",
						value: String(l.open_files)
					}),
					a !== void 0 && a > 0 && /* @__PURE__ */ (0, q.jsx)(ll, {
						label: "Duration",
						value: ze(a)
					}),
					e.exit_code !== void 0 && /* @__PURE__ */ (0, q.jsx)(ll, {
						label: "Exit",
						value: String(e.exit_code),
						tone: e.exit_code === 0 ? "success" : "danger"
					}),
					e.goroutine_count !== void 0 && /* @__PURE__ */ (0, q.jsx)(ll, {
						label: "Goroutines",
						value: String(e.goroutine_count)
					})
				]
			}) : /* @__PURE__ */ (0, q.jsx)("div", {
				className: "text-xs italic text-muted-foreground",
				children: "No metrics captured for this attempt."
			}),
			o.note && n && /* @__PURE__ */ (0, q.jsx)("div", {
				className: "mt-1 text-[11px] italic text-muted-foreground",
				children: o.note
			})
		] }), e.stack_trace && /* @__PURE__ */ (0, q.jsxs)("section", { children: [/* @__PURE__ */ (0, q.jsx)("div", {
			className: "mb-1 text-[10px] uppercase tracking-wide text-muted-foreground",
			children: "Stack trace"
		}), /* @__PURE__ */ (0, q.jsx)("div", {
			className: "flex max-h-[32rem] min-h-0 flex-col",
			children: /* @__PURE__ */ (0, q.jsx)(Wu, { text: e.stack_trace })
		})] })]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestDetailPanel.tsx
function ad({ node: e }) {
	let t = Ml(), n = e.attempts?.at(-1), r = (e.children?.length ?? 0) > 0 ? _e(e) : null;
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "space-y-4 p-density-4",
		children: [
			e.message && !e.failure_detail && /* @__PURE__ */ (0, q.jsx)("p", {
				className: "whitespace-pre-wrap break-words text-sm",
				children: e.message
			}),
			e.failure_detail && /* @__PURE__ */ (0, q.jsx)(Zu, { failure: e.failure_detail }),
			/* @__PURE__ */ (0, q.jsx)($u, { node: e }),
			n && /* @__PURE__ */ (0, q.jsx)(id, {
				attempt: n,
				...t.fetchProcessMetrics ? { fetchProcessMetrics: t.fetchProcessMetrics } : {}
			}, `${n.run_kind ?? ""}:${n.sequence}`),
			r && r.total > 0 && /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "flex flex-wrap gap-3 text-xs text-muted-foreground",
				children: [
					/* @__PURE__ */ (0, q.jsxs)("span", { children: [r.total, " total"] }),
					r.passed > 0 && /* @__PURE__ */ (0, q.jsxs)("span", { children: [r.passed, " passed"] }),
					r.failed > 0 && /* @__PURE__ */ (0, q.jsxs)("span", { children: [r.failed, " failed"] }),
					r.skipped > 0 && /* @__PURE__ */ (0, q.jsxs)("span", { children: [r.skipped, " skipped"] })
				]
			}),
			e.detail != null && /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "space-y-1",
				children: [/* @__PURE__ */ (0, q.jsx)("div", {
					className: "text-[10px] uppercase tracking-wide text-muted-foreground",
					children: "Detail"
				}), /* @__PURE__ */ (0, q.jsx)(v, { data: e.detail })]
			})
		]
	});
}
function od(e, t) {
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
var sd = {
	idle: "Copy details",
	pending: "Copying…",
	copied: "Copied",
	error: "Copy failed"
}, cd = {
	idle: w,
	pending: Dt,
	copied: de,
	error: mt
};
function ld({ node: e }) {
	let t = Ml(), { state: n, copy: r } = ie(), i = (0, K.useMemo)(() => t.adapters.resolveCopy(e, t), [e, t]), a = [{
		label: "Copy as JSON",
		icon: vt,
		onSelect: () => r(Cl(e))
	}, ...i.items.map((e) => od(e, r))];
	return /* @__PURE__ */ (0, q.jsx)(zl, {
		label: sd[n],
		icon: cd[n],
		variant: "outline",
		size: "sm",
		align: "right",
		title: "More copy options",
		onClick: () => r(Sl(e, i)),
		items: a
	});
}
function ud({ node: e, actionNode: t, actions: n, attemptChooser: r }) {
	let i = Ml(), o = we(e), s = !!i.onRerun && t.framework !== "task", c = !!i.onStop && !!t.can_stop && !!t.task_id, l = e.file ? `${e.file}${e.line ? `:${e.line}` : ""}` : void 0;
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "flex min-w-0 items-center gap-2 border-b border-border px-density-4 py-density-2",
		"data-detail-header": "",
		children: [
			o ? /* @__PURE__ */ (0, q.jsx)(G, {
				outcome: o,
				className: "text-lg"
			}) : /* @__PURE__ */ (0, q.jsx)(E, {
				icon: gt,
				className: "shrink-0 text-lg text-muted-foreground"
			}),
			/* @__PURE__ */ (0, q.jsx)("h2", {
				className: "min-w-0 shrink truncate text-base font-semibold",
				title: e.name,
				children: e.name
			}),
			e.framework && /* @__PURE__ */ (0, q.jsxs)("span", {
				className: "inline-flex shrink-0 items-center gap-1 rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground",
				children: [/* @__PURE__ */ (0, q.jsx)(ke, {
					tool: Ae(e),
					className: "text-sm"
				}), e.framework]
			}),
			e.duration ? /* @__PURE__ */ (0, q.jsxs)("span", {
				className: "inline-flex shrink-0 items-center gap-1 text-xs text-muted-foreground",
				children: [/* @__PURE__ */ (0, q.jsx)(E, {
					icon: Tt,
					className: "text-xs"
				}), ze(e.duration)]
			}) : null,
			l && /* @__PURE__ */ (0, q.jsx)("span", {
				className: "min-w-0 shrink-[100] truncate font-mono text-xs text-muted-foreground",
				title: l,
				children: l
			}),
			e.file && i.onOpenFile && /* @__PURE__ */ (0, q.jsx)(ul, {
				className: "shrink-0",
				onOpenFile: i.onOpenFile,
				target: {
					file: Ne(e.file, e.work_dir),
					...e.work_dir ? { workDir: e.work_dir } : {},
					...e.line ? { line: e.line } : {}
				}
			}),
			/* @__PURE__ */ (0, q.jsxs)("div", {
				className: "ml-auto flex shrink-0 items-center gap-1",
				children: [
					r,
					n,
					/* @__PURE__ */ (0, q.jsx)(ld, { node: e }),
					s && /* @__PURE__ */ (0, q.jsxs)(a, {
						size: "sm",
						disabled: i.busy.rerun,
						onClick: () => i.onRerun(t),
						title: "Rerun this test or subtree",
						children: [/* @__PURE__ */ (0, q.jsx)(E, {
							icon: wt,
							className: "mr-1 text-sm"
						}), i.busy.rerun ? "Running..." : "Rerun"]
					}),
					c && /* @__PURE__ */ (0, q.jsxs)(a, {
						size: "sm",
						variant: "destructive",
						disabled: i.busy.stop,
						onClick: () => i.onStop(t),
						title: "Stop this running task",
						children: [/* @__PURE__ */ (0, q.jsx)(E, {
							icon: u,
							className: "mr-1 text-sm"
						}), i.busy.stop ? "Stopping..." : "Stop"]
					})
				]
			})
		]
	});
}
function dd() {
	let e = Ml(), t = e.selected, [n, r] = (0, K.useState)(null), a = e.activeTab !== void 0, [o, s] = (0, K.useState)("detail"), c = a ? e.activeTab : o, l = (t) => {
		a ? e.onTabChange?.(t) : s(t);
	};
	if (!t) return /* @__PURE__ */ (0, q.jsx)("div", {
		className: "flex h-full items-center justify-center text-sm text-muted-foreground",
		children: /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "text-center",
			children: [/* @__PURE__ */ (0, q.jsx)(E, {
				icon: gt,
				className: "mb-2 block text-4xl"
			}), "Select a test to view its details."]
		})
	});
	let u = e.adapters.resolveAttempts(t, e), d = t.task_id ?? t.name, f = e.activeAttempt ?? (n?.key === d ? n.id : void 0), p = u.find((e) => e.id === f) ?? u.at(-1), m = u.length > 1 && p ? p.node : t, h = u.length > 1 && p ? /* @__PURE__ */ (0, q.jsx)("div", {
		className: "w-44 shrink-0",
		children: /* @__PURE__ */ (0, q.jsx)(Kn, {
			"aria-label": "Attempt",
			value: String(p.id),
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
				label: `Attempt ${e.id}/${u.length} · ${e.status}${e.durationMs === void 0 ? "" : ` · ${ze(e.durationMs * 1e6)}`}`
			}))
		})
	}) : null, g = e.adapters.resolveDetail(m, e), _ = e.adapters.resolveActions(t, e), v = e.adapters.resolveTabs(m, e), y = e.adapters.ownsScroll(m, e), b = /* @__PURE__ */ (0, q.jsx)(Vl, {
		value: l,
		children: g?.renderDetail ? g.renderDetail({
			node: m,
			runner: e
		}) : /* @__PURE__ */ (0, q.jsx)(ad, { node: m })
	}), x = _?.nodeActions?.({
		node: t,
		runner: e
	}), S = i("min-h-0 flex-1 [container-type:size]", !y && "overflow-y-auto");
	if (v.length === 0) return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		children: [/* @__PURE__ */ (0, q.jsx)(ud, {
			node: m,
			actionNode: t,
			actions: x,
			attemptChooser: h
		}), /* @__PURE__ */ (0, q.jsx)("div", {
			className: S,
			children: b
		})]
	});
	let C = {
		id: "detail",
		label: "Detail",
		render: () => b
	}, w = [C, ...v], T = w.find((e) => e.id === c) ?? C;
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		children: [
			/* @__PURE__ */ (0, q.jsx)(ud, {
				node: m,
				actionNode: t,
				actions: x,
				attemptChooser: h
			}),
			/* @__PURE__ */ (0, q.jsx)("div", {
				className: "flex shrink-0 items-center gap-1 border-b border-border px-density-2 py-density-1",
				children: w.map((e) => /* @__PURE__ */ (0, q.jsx)(A, {
					active: T.id === e.id,
					onClick: () => l(e.id),
					label: e.label
				}, e.id))
			}),
			/* @__PURE__ */ (0, q.jsx)("div", {
				className: S,
				children: T.render()
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestRunSummary.tsx
function fd(e, t) {
	return t > 0 ? Math.round(e / t * 100) : 0;
}
function pd({ tests: e, startTime: t, endTime: n, done: r, now: a, runMeta: o, compact: s = !1, className: c }) {
	let l = Pe(e), u = r && n ? n : a ?? 0, d = t && u >= t ? `${((u - t) / 1e3).toFixed(1)}s` : "", f = [
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
	], p = [
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
	].filter((e) => e.n > 0).map((e) => /* @__PURE__ */ (0, q.jsxs)("span", {
		className: i("inline-flex items-center gap-0.5 tabular-nums", Ie[e.outcome].className),
		"aria-label": `${e.n} ${e.label}`,
		title: `${e.n} ${e.label}`,
		"data-summary-outcome": e.outcome,
		children: [/* @__PURE__ */ (0, q.jsx)(G, { outcome: e.outcome }), e.n]
	}, e.label)), m = /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [o?.timed_out && /* @__PURE__ */ (0, q.jsxs)("span", {
		className: i("inline-flex items-center gap-0.5", Ie.timedout.className),
		title: "The run hit its timeout",
		"data-summary-run": "timed-out",
		children: [/* @__PURE__ */ (0, q.jsx)(G, { outcome: "timedout" }), "Timed out"]
	}), o?.exit_code !== void 0 && /* @__PURE__ */ (0, q.jsxs)("span", {
		className: i("inline-flex items-center tabular-nums", o.exit_code === 0 ? "text-muted-foreground" : Ie.failed.className),
		title: `Exit code ${o.exit_code}`,
		"data-summary-run": "exit-code",
		children: ["exit ", o.exit_code]
	})] });
	return s ? /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i("flex flex-wrap items-center gap-x-3 gap-y-0.5 text-sm text-muted-foreground", c),
		children: [
			/* @__PURE__ */ (0, q.jsxs)("span", {
				className: "font-medium text-foreground",
				children: [
					l.total,
					" ",
					l.total === 1 ? "test" : "tests",
					o?.kind === "rerun" ? ` · rerun #${o.sequence}` : ""
				]
			}),
			p,
			m,
			d && /* @__PURE__ */ (0, q.jsxs)("span", {
				className: "inline-flex items-center gap-1",
				children: [/* @__PURE__ */ (0, q.jsx)(E, {
					icon: Tt,
					className: "text-xs"
				}), d]
			}),
			l.total > 0 && /* @__PURE__ */ (0, q.jsxs)("span", {
				className: "inline-flex items-center gap-1.5",
				children: [/* @__PURE__ */ (0, q.jsx)(ol, {
					className: "w-24",
					segments: f,
					total: l.total,
					height: "h-1.5"
				}), /* @__PURE__ */ (0, q.jsxs)("span", {
					className: "text-xs font-medium text-green-600 dark:text-green-400",
					children: [fd(l.passed, l.total), "%"]
				})]
			}),
			!r && /* @__PURE__ */ (0, q.jsx)(E, {
				icon: be,
				className: "animate-spin text-blue-500"
			})
		]
	}) : /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i("flex min-w-[18rem] flex-col items-end gap-1.5", c),
		children: [/* @__PURE__ */ (0, q.jsxs)("div", {
			className: "flex flex-wrap items-center justify-end gap-x-3 gap-y-0.5 text-sm text-muted-foreground",
			children: [
				/* @__PURE__ */ (0, q.jsxs)("span", {
					className: "font-medium text-foreground",
					children: [
						l.total,
						" ",
						l.total === 1 ? "test" : "tests",
						o?.kind === "rerun" ? ` · rerun #${o.sequence}` : ""
					]
				}),
				p,
				m,
				d && /* @__PURE__ */ (0, q.jsxs)("span", {
					className: "inline-flex items-center gap-1",
					children: [/* @__PURE__ */ (0, q.jsx)(E, {
						icon: Tt,
						className: "text-xs"
					}), d]
				}),
				!r && /* @__PURE__ */ (0, q.jsx)(E, {
					icon: be,
					className: "animate-spin text-blue-500"
				})
			]
		}), l.total > 0 && /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "flex w-72 items-center gap-2",
			children: [/* @__PURE__ */ (0, q.jsx)(ol, {
				className: "flex-1",
				segments: f,
				total: l.total,
				height: "h-2"
			}), /* @__PURE__ */ (0, q.jsxs)("span", {
				className: "shrink-0 text-xs font-medium text-green-600 dark:text-green-400",
				children: [fd(l.passed, l.total), "%"]
			})]
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestFilterBar.tsx
var md = {
	autoSubmit: !1,
	compact: !1
};
function hd({ label: e, options: t, state: n, onChange: r }) {
	let a = Object.fromEntries(n), o = n.size > 0;
	return /* @__PURE__ */ (0, q.jsx)(fe, {
		align: "right",
		menuLabel: e,
		menuClassName: "p-2",
		trigger: /* @__PURE__ */ (0, q.jsxs)("span", {
			role: "button",
			tabIndex: 0,
			className: i("inline-flex h-6 items-center gap-0.5 rounded px-1.5 text-xs transition-colors", o ? "bg-muted font-medium text-foreground" : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"),
			children: [Qt(e, a), /* @__PURE__ */ (0, q.jsx)(E, {
				icon: F,
				className: "text-[10px] opacity-60"
			})]
		}),
		children: () => /* @__PURE__ */ (0, q.jsx)(Yt.Provider, {
			value: md,
			children: /* @__PURE__ */ (0, q.jsx)(Mn, {
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
function gd({ filters: e, onChange: t, counts: n, frameworks: r, className: a }) {
	let o = Fe.filter((e) => n[e.key] > 0).map((e) => ({
		value: e.key,
		label: e.label,
		icon: /* @__PURE__ */ (0, q.jsx)(G, { outcome: e.key }),
		count: n[e.key]
	})), s = r.map((e) => ({
		value: e,
		label: je(e)?.label ?? e,
		title: e,
		icon: /* @__PURE__ */ (0, q.jsx)(ke, { tool: e })
	}));
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: i("flex items-center gap-0.5", a),
		children: [/* @__PURE__ */ (0, q.jsx)(hd, {
			label: "Status",
			options: o,
			state: e.status,
			onChange: (n) => t({
				...e,
				status: n
			})
		}), r.length > 1 && /* @__PURE__ */ (0, q.jsx)(hd, {
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
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestRunner.tsx
function _d(e) {
	return e instanceof Ol ? e : kl(e ?? []);
}
var vd = 640;
function yd({ tests: e, allTests: t, selected: n, filters: r, expandAll: a, done: o, runMeta: s, status: c, statusText: l, now: u, busy: d, activeTab: f, onTabChange: p, activeAttempt: m, onAttemptChange: h, onSelect: g, onFiltersChange: _, onExpandAllChange: v, onRerun: y, onStop: b, onOpenFile: x, fetchProcessMetrics: S, adapters: C, title: w, showSummary: T = !0, headerSlot: D, hideHeader: O = !1, defaultSplit: k = 45, startTime: A, endTime: j, className: M }) {
	let N = (0, K.useMemo)(() => _d(C), [C]), P = (0, K.useMemo)(() => ({
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
		onTabChange: p,
		activeAttempt: m,
		onAttemptChange: h,
		onSelect: g,
		onFiltersChange: _,
		onExpandAllChange: v,
		onRerun: y,
		onStop: b,
		onOpenFile: x,
		fetchProcessMetrics: S,
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
		p,
		m,
		h,
		g,
		_,
		v,
		y,
		b,
		x,
		S,
		N
	]), ee = t ?? e, F = (0, K.useMemo)(() => Me(ee), [ee]), I = (0, K.useMemo)(() => Pe(ee), [ee]), L = !O && ee.length > 0 ? /* @__PURE__ */ (0, q.jsx)(gd, {
		filters: r,
		onChange: _,
		counts: I,
		frameworks: F
	}) : null, { ref: te, wider: ne } = Un(vd);
	return /* @__PURE__ */ (0, q.jsx)(jl, {
		value: P,
		children: /* @__PURE__ */ (0, q.jsxs)("div", {
			ref: te,
			className: i("flex h-full min-h-0 flex-col bg-background", M),
			children: [!O && /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "flex min-w-0 items-center gap-3 border-b border-border bg-card px-density-4 py-density-2",
				"data-runner-header": "",
				children: [
					/* @__PURE__ */ (0, q.jsxs)("div", {
						className: "flex min-w-0 items-center gap-3",
						children: [w === void 0 ? /* @__PURE__ */ (0, q.jsxs)("h1", {
							className: "inline-flex shrink-0 items-center gap-1.5 text-base font-bold",
							children: [/* @__PURE__ */ (0, q.jsx)(E, {
								icon: Ce,
								className: "text-primary"
							}), "Test Results"]
						}) : w, l && /* @__PURE__ */ (0, q.jsx)("span", {
							className: "truncate text-sm text-muted-foreground",
							children: l
						})]
					}),
					T && /* @__PURE__ */ (0, q.jsx)(pd, {
						tests: ee,
						startTime: A,
						endTime: j,
						done: o,
						now: u,
						runMeta: s,
						compact: !0,
						className: "ml-auto flex-nowrap"
					}),
					D && /* @__PURE__ */ (0, q.jsx)("div", {
						className: i("flex shrink-0 items-center gap-1", !T && "ml-auto"),
						children: D
					})
				]
			}), ne ? /* @__PURE__ */ (0, q.jsx)(ye, {
				className: "min-h-0 flex-1",
				defaultSplit: k,
				minLeft: 25,
				minRight: 30,
				leftClass: "border-r-2 border-border bg-muted/30",
				left: /* @__PURE__ */ (0, q.jsx)(Rl, { actions: L }),
				right: /* @__PURE__ */ (0, q.jsx)(dd, {})
			}) : n ? /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "flex min-h-0 flex-1 flex-col",
				children: [/* @__PURE__ */ (0, q.jsxs)("button", {
					type: "button",
					onClick: () => g(null),
					className: "inline-flex items-center gap-1.5 border-b border-border px-density-3 py-density-2 text-left text-sm text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
					children: [/* @__PURE__ */ (0, q.jsx)(E, { icon: ae }), "Back to results"]
				}), /* @__PURE__ */ (0, q.jsx)("div", {
					className: "min-h-0 flex-1 overflow-hidden",
					children: /* @__PURE__ */ (0, q.jsx)(dd, {})
				})]
			}) : /* @__PURE__ */ (0, q.jsx)("div", {
				className: "min-h-0 flex-1 overflow-y-auto",
				children: /* @__PURE__ */ (0, q.jsx)(Rl, { actions: L })
			})]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/CopyButton.tsx
function bd({ value: e, label: t = "Copy", className: n, iconClassName: r }) {
	let { state: a, copy: o } = ie(), s = (0, K.useCallback)((t) => {
		t.stopPropagation(), o(typeof e == "function" ? e() : e);
	}, [o, e]);
	return /* @__PURE__ */ (0, q.jsx)(go, {
		icon: a === "copied" ? de : w,
		label: a === "copied" ? "Copied" : a === "error" ? "Copy failed" : t,
		onClick: s,
		className: n,
		iconClassName: i(a === "copied" && "text-emerald-600", a === "error" && "text-red-600", r)
	});
}
// ../../../clicky-ui/packages/ui/src/data/verification/verifyReportTests.ts
var xd = "acceptance-criteria";
function Sd(e) {
	return typeof e == "number" && Number.isFinite(e) ? e : void 0;
}
function Cd(e) {
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
function wd(e) {
	let t = { name: e.name };
	e.framework !== void 0 && (t.framework = e.framework), e.task_id !== void 0 && (t.task_id = e.task_id), e.file !== void 0 && (t.file = e.file), e.line !== void 0 && (t.line = e.line), e.message !== void 0 && (t.message = e.message), e.command !== void 0 && (t.command = e.command), e.work_dir !== void 0 && (t.work_dir = e.work_dir), e.stdout !== void 0 && (t.stdout = e.stdout), e.stderr !== void 0 && (t.stderr = e.stderr);
	let n = Sd(e.duration);
	return n !== void 0 && (t.duration = n), e.passed !== void 0 && (t.passed = e.passed), e.failed !== void 0 && (t.failed = e.failed), e.warned !== void 0 && (t.warned = e.warned), e.skipped !== void 0 && (t.skipped = e.skipped), e.pending !== void 0 && (t.pending = e.pending), e.running !== void 0 && (t.running = e.running), e.timed_out !== void 0 && (t.timed_out = e.timed_out), e.progress !== void 0 && (t.progress = e.progress), e.context !== void 0 && (t.context = e.context), e.detail !== void 0 && (t.detail = e.detail), e.summary === void 0 ? e.children !== void 0 && (t.children = e.children.map(wd)) : t.summary = Cd(e.summary), t;
}
function Td(e, t) {
	let n = {
		name: e.item || `Criterion ${t + 1}`,
		framework: "checklist",
		task_id: `${xd}:${t}`,
		passed: e.passed === !0,
		failed: e.passed === !1,
		pending: e.passed !== !0 && e.passed !== !1,
		detail: e
	};
	return e.message !== void 0 && (n.message = e.message), n;
}
function Ed(e) {
	return e.length === 0 ? null : {
		name: "Acceptance criteria",
		framework: "checklist",
		task_id: xd,
		children: e.map(Td)
	};
}
function Dd(e) {
	let t = (e.tests ?? []).map(wd), n = Ed(e.checklist ?? []);
	return n ? [...t, n] : t;
}
// ../../../clicky-ui/packages/ui/src/data/verification/stepKind.ts
var Od = [
	"satisfied",
	"failed",
	"timeout"
], kd = [
	"alert",
	"console",
	"pageerror",
	"response"
], Ad = /* @__PURE__ */ new Set([
	"step",
	"outputs",
	"exports",
	"wait",
	"attempts",
	"events",
	"background"
]), jd = class extends Error {};
function Md(e, t) {
	throw new jd(`${e} must be ${t}`);
}
var Nd = (e) => typeof e == "object" && !!e && !Array.isArray(e);
function Pd(e, t) {
	return Nd(e) ? e : Md(t, "an object");
}
function Fd(e, t) {
	return Array.isArray(e) ? e : Md(t, "an array");
}
function Id(e, t) {
	return typeof e == "string" ? e : Md(t, "a string");
}
function Ld(e, t) {
	return typeof e == "number" && Number.isFinite(e) ? e : Md(t, "a number");
}
function Rd(e, t) {
	return typeof e == "boolean" ? e : Md(t, "a boolean");
}
function zd(e, t, n) {
	return t.includes(e) ? e : Md(n, `one of ${t.join(", ")} (got ${JSON.stringify(e)})`);
}
function $(e, t, n) {
	return e[t] === void 0 || e[t] === null ? {} : { [t]: n(e[t]) };
}
function Bd(e, t) {
	return typeof e == "string" ? e : typeof e == "number" ? ze(e) : Md(t, "a duration string or nanosecond number");
}
function Vd(e, t) {
	return typeof e == "string" ? e : typeof e == "number" ? String(e) : Md(t, "a string or number");
}
function Hd(e) {
	let t = "detail.wait.policy", n = Pd(e, t), r = (e) => {
		let n = Pd(e, `${t}.backoff`);
		return {
			...$(n, "factor", (e) => Ld(e, `${t}.backoff.factor`)),
			...$(n, "max", (e) => Bd(e, `${t}.backoff.max`))
		};
	};
	return {
		...$(n, "until", (e) => Id(e, `${t}.until`)),
		...$(n, "failWhen", (e) => Id(e, `${t}.failWhen`)),
		...$(n, "interval", (e) => Bd(e, `${t}.interval`)),
		...$(n, "backoff", r),
		...$(n, "jitter", (e) => Vd(e, `${t}.jitter`)),
		...$(n, "timeout", (e) => Bd(e, `${t}.timeout`))
	};
}
function Ud(e) {
	let t = Pd(e, "detail.wait");
	return {
		policy: Hd(t.policy ?? {}),
		outcome: zd(t.outcome, Od, "detail.wait.outcome"),
		durationMs: Wd(t.durationMs, "detail.wait.durationMs"),
		...t.outputs === void 0 ? {} : { outputs: t.outputs },
		...$(t, "expression", (e) => Id(e, "detail.wait.expression")),
		...$(t, "cause", (e) => Id(e, "detail.wait.cause")),
		...$(t, "error", (e) => Id(e, "detail.wait.error")),
		...$(t, "trace", (e) => Id(e, "detail.wait.trace"))
	};
}
function Wd(e, t) {
	let n = Ld(e, t);
	return n >= 0 ? n : Md(t, "non-negative");
}
function Gd(e, t) {
	let n = Pd(e, t);
	Id(n.name, `${t}.name`);
	for (let e of [
		"framework",
		"message",
		"command",
		"work_dir",
		"stdout",
		"stderr"
	]) n[e] !== void 0 && Id(n[e], `${t}.${e}`);
	for (let e of [
		"passed",
		"failed",
		"timed_out",
		"warned",
		"skipped",
		"pending",
		"running"
	]) n[e] !== void 0 && Rd(n[e], `${t}.${e}`);
	return n.duration !== void 0 && Wd(n.duration, `${t}.duration`), n.children !== void 0 && Fd(n.children, `${t}.children`).forEach((e, n) => Gd(e, `${t}.children[${n}]`)), {
		...n,
		name: Id(n.name, `${t}.name`)
	};
}
function Kd(e, t) {
	let n = `detail.attempts[${t}]`, r = Pd(e, n);
	return {
		attempt: Ld(r.attempt, `${n}.attempt`),
		status: Id(r.status, `${n}.status`),
		...$(r, "error", (e) => Id(e, `${n}.error`)),
		...$(r, "durationMs", (e) => Wd(e, `${n}.durationMs`)),
		...$(r, "result", (e) => Gd(e, `${n}.result`))
	};
}
function qd(e) {
	let t = Fd(e, "detail.attempts").map(Kd);
	return t.forEach((e, t) => {
		e.attempt !== t + 1 && Md(`detail.attempts[${t}].attempt`, `the sequence number ${t + 1}`);
	}), t;
}
function Jd(e) {
	return Object.entries(Pd(e, "detail.exports")).map(([e, t]) => {
		let n = `detail.exports.${e}`, r = Pd(t, n);
		return {
			name: e,
			value: r.value,
			step: Id(r.step, `${n}.step`)
		};
	});
}
function Yd(e, t) {
	let n = `detail.events[${t}]`, r = Pd(e, n);
	return {
		type: zd(r.type, kd, `${n}.type`),
		...$(r, "text", (e) => Id(e, `${n}.text`)),
		...$(r, "url", (e) => Id(e, `${n}.url`)),
		...$(r, "status", (e) => Ld(e, `${n}.status`)),
		...$(r, "detail", (e) => Id(e, `${n}.detail`))
	};
}
function Xd(e) {
	let t = Pd(e.step, "detail.step"), n = Object.fromEntries(Object.entries(e).filter(([e]) => !Ad.has(e)));
	return {
		step: {
			id: Id(t.id, "detail.step.id"),
			kind: Id(t.kind, "detail.step.kind")
		},
		...e.outputs === void 0 ? {} : { outputs: e.outputs },
		exports: e.exports == null ? [] : Jd(e.exports),
		...$(e, "wait", Ud),
		attempts: qd(e.attempts ?? []),
		events: Fd(e.events ?? [], "detail.events").map(Yd),
		background: Rd(e.background ?? !1, "detail.background"),
		rest: n
	};
}
function Zd(e) {
	if (!Nd(e) || e.step === void 0) return { kind: "none" };
	try {
		return {
			kind: "step",
			step: Xd(e)
		};
	} catch (e) {
		if (e instanceof jd) return {
			kind: "invalid",
			error: e.message
		};
		throw e;
	}
}
function Qd(e) {
	return Zd(e).kind !== "none";
}
var $d = (e) => typeof e == "string", ef = (e) => typeof e == "number" && Number.isFinite(e);
function tf(e) {
	if (!Nd(e)) return {
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
	n("name", $d), i("kind", n("kind", $d)), i("state", n("state", $d)), i("startedAt", n("started_at", $d)), i("finishedAt", n("finished_at", $d)), i("durationNs", n("duration", ef)), i("error", n("error", $d));
	let a = n("origin", Nd);
	return a && (i("file", $d(a.file) ? a.file : void 0), i("sectionPath", $d(a.section_path) ? a.section_path : void 0), i("rowIndex", ef(a.row_index) ? a.row_index : void 0), i("line", ef(a.line) ? a.line : void 0)), {
		header: r,
		rest: Object.keys(t).length > 0 ? t : void 0
	};
}
function nf(e) {
	return e.expression === void 0 ? e.outcome === "failed" && e.policy.failWhen ? "failWhen" : "until" : e.expression === e.policy.failWhen ? "failWhen" : "until";
}
function rf(e) {
	return /T(\d{2}:\d{2}:\d{2})/.exec(e)?.[1] ?? e;
}
function af(e) {
	return e <= 0 ? "0ms" : ze(e * 1e6);
}
function of(e) {
	let t = [e.backoff?.factor !== void 0 && `×${e.backoff.factor}`, e.backoff?.max && `up to ${e.backoff.max}`].filter(Boolean);
	return [
		e.interval && `every ${e.interval}`,
		t.length > 0 && `backoff ${t.join(" ")}`,
		e.jitter && `jitter ±${e.jitter}`,
		e.timeout && `timeout ${e.timeout}`
	].filter(Boolean).join(" · ");
}
function sf(e) {
	return e === void 0 ? "" : typeof e == "string" ? e : JSON.stringify(e);
}
function cf(e) {
	let { policy: t } = e;
	return {
		title: "Wait",
		body: [
			`outcome: ${e.outcome}`,
			`elapsed: ${af(e.durationMs)}`,
			`policy: ${of(t)}`,
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
function lf(e) {
	return `${[
		`#${e.attempt}`,
		e.status,
		e.durationMs !== void 0 && af(e.durationMs)
	].filter(Boolean).join(" ")}${e.error ? ` — ${e.error}` : ""}`;
}
function uf(e) {
	let t = [
		e.status,
		e.text,
		e.url
	].filter((e) => e !== void 0).join(" ");
	return `[${e.type}]${t ? ` ${t}` : ""}${e.detail ? ` — ${e.detail}` : ""}`;
}
function df(e) {
	return [
		...e.outputs === void 0 ? [] : [{
			title: "Outputs",
			body: JSON.stringify(e.outputs, null, 2),
			lang: "json"
		}],
		...e.wait ? [cf(e.wait)] : [],
		...e.attempts.length > 0 ? [{
			title: "Attempts",
			body: e.attempts.map(lf).join("\n")
		}] : [],
		...e.exports.length > 0 ? [{
			title: "Exports",
			body: e.exports.map((e) => `${e.name}=${sf(e.value)} (from ${e.step})`).join("\n")
		}] : [],
		...e.events.length > 0 ? [{
			title: "Events",
			body: e.events.map(uf).join("\n")
		}] : []
	];
}
// ../../../clicky-ui/packages/ui/src/data/verification/stepKindViews.tsx
var ff = "border-red-500/60 bg-red-500/10 text-red-700 dark:text-red-400";
function pf({ name: e, title: t, aside: n, children: r }) {
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "min-w-0 space-y-1.5",
		"data-step-section": e,
		children: [/* @__PURE__ */ (0, q.jsxs)("div", {
			className: "flex items-center gap-2",
			children: [/* @__PURE__ */ (0, q.jsx)("div", {
				className: "text-[10px] uppercase tracking-wide text-muted-foreground",
				children: t
			}), n]
		}), r]
	});
}
var mf = {
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
function hf({ condition: e, expression: t, decisive: n, wait: r }) {
	let { label: i, tone: a } = mf[r.outcome];
	return /* @__PURE__ */ (0, q.jsxs)("li", {
		className: "flex min-w-0 items-center gap-2",
		"data-wait-condition": e,
		...n ? { "data-wait-decisive": "" } : {},
		children: [
			/* @__PURE__ */ (0, q.jsx)("span", {
				className: "w-14 shrink-0 text-muted-foreground",
				children: e
			}),
			/* @__PURE__ */ (0, q.jsx)("code", {
				className: "min-w-0 shrink break-words font-mono",
				children: t
			}),
			n && /* @__PURE__ */ (0, q.jsx)(L, {
				size: "xs",
				clickToCopy: !1,
				tone: a,
				children: i
			})
		]
	});
}
function gf({ wait: e, onViewOutputs: t }) {
	let { policy: n } = e, r = of(n), i = nf(e), o = e.outputs !== void 0 && !t;
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "min-w-0 text-[11px]",
		"data-step-section": "wait",
		children: [
			/* @__PURE__ */ (0, q.jsxs)("div", {
				className: "flex min-w-0 items-center gap-2",
				"data-wait-summary": "",
				children: [
					/* @__PURE__ */ (0, q.jsx)("span", {
						className: "shrink-0 text-muted-foreground",
						children: "Wait"
					}),
					/* @__PURE__ */ (0, q.jsx)(L, {
						size: "xs",
						clickToCopy: !1,
						children: af(e.durationMs)
					}),
					e.outputs !== void 0 && t && /* @__PURE__ */ (0, q.jsx)(a, {
						type: "button",
						variant: "link",
						size: "sm",
						className: "h-auto px-0 text-[11px]",
						onClick: t,
						children: "View outputs"
					})
				]
			}),
			/* @__PURE__ */ (0, q.jsxs)("ul", {
				className: "mt-1 space-y-0.5",
				"data-wait-conditions": "",
				children: [/* @__PURE__ */ (0, q.jsx)(hf, {
					condition: "until",
					expression: n.until ?? "Service readiness",
					decisive: i === "until",
					wait: e
				}), n.failWhen && /* @__PURE__ */ (0, q.jsx)(hf, {
					condition: "failWhen",
					expression: n.failWhen,
					decisive: i === "failWhen",
					wait: e
				})]
			}),
			(r || e.error || e.trace || o) && /* @__PURE__ */ (0, q.jsxs)("details", {
				className: "mt-1 text-muted-foreground",
				open: e.outcome !== "satisfied",
				children: [/* @__PURE__ */ (0, q.jsx)("summary", {
					className: "cursor-pointer",
					children: "Wait details"
				}), /* @__PURE__ */ (0, q.jsxs)("div", {
					className: "mt-1 space-y-1 break-words",
					children: [
						r && /* @__PURE__ */ (0, q.jsx)("p", { children: r }),
						e.error && /* @__PURE__ */ (0, q.jsx)("p", {
							className: "text-destructive",
							children: e.error
						}),
						e.trace && /* @__PURE__ */ (0, q.jsx)("pre", {
							className: "whitespace-pre-wrap",
							children: e.trace
						}),
						o && /* @__PURE__ */ (0, q.jsxs)("div", {
							className: "min-w-0",
							"data-wait-outputs": "",
							children: [/* @__PURE__ */ (0, q.jsx)("div", {
								className: "text-[10px] uppercase tracking-wide",
								children: "Outputs"
							}), /* @__PURE__ */ (0, q.jsx)(v, {
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
var _f = {
	passed: "success",
	failed: "danger",
	warned: "warning",
	timedout: "warning",
	running: "info"
};
function vf(e) {
	return e.split("/").filter(Boolean).at(-1) ?? e;
}
function yf({ rowIndex: e, line: t }) {
	return e === void 0 ? t === void 0 ? void 0 : `line ${t}` : `step #${e}`;
}
function bf({ header: e, stepKind: t }) {
	let n = [yf(e), e.sectionPath].filter(Boolean).join(" · "), r = [e.kind, t].filter((e, t, n) => !!e && n.indexOf(e) === t), i = e.startedAt && e.finishedAt ? `${rf(e.startedAt)} → ${rf(e.finishedAt)}` : e.startedAt ? `started ${rf(e.startedAt)}` : void 0;
	return !e.file && !n && r.length === 0 && !e.state && !i && e.durationNs === void 0 ? null : /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "min-w-0 space-y-1",
		"data-execution-header": "",
		children: [(e.file || n) && /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "flex min-w-0 flex-wrap items-baseline gap-x-1.5 text-xs",
			children: [
				e.file && /* @__PURE__ */ (0, q.jsx)("span", {
					className: "font-medium",
					title: e.file,
					children: vf(e.file)
				}),
				e.file && n && /* @__PURE__ */ (0, q.jsx)("span", {
					className: "text-muted-foreground",
					children: "·"
				}),
				n && /* @__PURE__ */ (0, q.jsx)("span", {
					className: "min-w-0 break-words text-muted-foreground",
					children: n
				})
			]
		}), /* @__PURE__ */ (0, q.jsxs)("div", {
			className: "flex min-w-0 flex-wrap items-center gap-2 text-[11px] text-muted-foreground",
			children: [
				r.map((e) => /* @__PURE__ */ (0, q.jsx)(L, {
					size: "xs",
					clickToCopy: !1,
					children: e
				}, e)),
				e.state && /* @__PURE__ */ (0, q.jsx)(L, {
					size: "xs",
					clickToCopy: !1,
					tone: _f[e.state] ?? "neutral",
					children: e.state
				}),
				i && /* @__PURE__ */ (0, q.jsx)("span", {
					className: "font-mono",
					children: i
				}),
				e.durationNs !== void 0 && /* @__PURE__ */ (0, q.jsx)("span", { children: ze(e.durationNs) })
			]
		})]
	});
}
function xf({ children: e }) {
	return /* @__PURE__ */ (0, q.jsx)("div", {
		className: i("whitespace-pre-wrap break-words rounded border px-2 py-1.5 text-sm", ff),
		"data-failure-callout": "",
		children: e
	});
}
function Sf({ exports: e }) {
	return /* @__PURE__ */ (0, q.jsxs)("table", {
		className: "w-full table-fixed border-collapse text-left text-[11px]",
		children: [/* @__PURE__ */ (0, q.jsx)("thead", {
			className: "text-muted-foreground",
			children: /* @__PURE__ */ (0, q.jsxs)("tr", {
				className: "border-b border-border",
				children: [
					/* @__PURE__ */ (0, q.jsx)("th", {
						className: "w-[28%] py-1 pr-2 font-medium",
						children: "Name"
					}),
					/* @__PURE__ */ (0, q.jsx)("th", {
						className: "py-1 pr-2 font-medium",
						children: "Value"
					}),
					/* @__PURE__ */ (0, q.jsx)("th", {
						className: "w-[18%] py-1 font-medium",
						children: "Step"
					})
				]
			})
		}), /* @__PURE__ */ (0, q.jsx)("tbody", {
			className: "font-mono",
			children: e.map((e) => /* @__PURE__ */ (0, q.jsxs)("tr", {
				className: "border-b border-border/60 align-top last:border-0",
				children: [
					/* @__PURE__ */ (0, q.jsx)("td", {
						className: "break-words py-1 pr-2",
						children: e.name
					}),
					/* @__PURE__ */ (0, q.jsx)("td", {
						className: "break-all py-1 pr-2",
						children: sf(e.value)
					}),
					/* @__PURE__ */ (0, q.jsx)("td", {
						className: "break-words py-1 text-muted-foreground",
						children: e.step
					})
				]
			}, e.name))
		})]
	});
}
function Cf(e) {
	return e.type === "alert" || e.type === "pageerror" || e.type === "response" && (e.status ?? 0) >= 400;
}
function wf({ event: e }) {
	let t = Cf(e);
	return /* @__PURE__ */ (0, q.jsxs)("li", {
		className: i("space-y-0.5 rounded border px-2 py-1 text-[11px]", t ? ff : "border-border"),
		"data-event-type": e.type,
		children: [
			/* @__PURE__ */ (0, q.jsxs)("div", {
				className: "flex min-w-0 items-center gap-2",
				children: [
					/* @__PURE__ */ (0, q.jsx)(L, {
						size: "xs",
						clickToCopy: !1,
						tone: t ? "danger" : "neutral",
						children: e.type
					}),
					e.status !== void 0 && /* @__PURE__ */ (0, q.jsx)("span", {
						className: "font-mono",
						children: e.status
					}),
					e.text && /* @__PURE__ */ (0, q.jsx)("span", {
						className: i("min-w-0 break-words", e.type === "alert" && "font-medium"),
						children: e.text
					})
				]
			}),
			e.url && /* @__PURE__ */ (0, q.jsx)("div", {
				className: "break-all font-mono text-[10px] opacity-80",
				children: e.url
			}),
			e.detail && /* @__PURE__ */ (0, q.jsx)("p", {
				className: "whitespace-pre-wrap break-words",
				children: e.detail
			})
		]
	});
}
function Tf({ step: e, renderWait: t }) {
	let n = Hl(), r = n && e.outputs !== void 0 ? () => n("outputs") : void 0;
	return /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
		/* @__PURE__ */ (0, q.jsxs)("div", {
			className: "text-[11px] text-muted-foreground",
			"data-step-ref": "",
			children: [
				e.step.kind,
				" · ",
				e.step.id,
				e.background && " · background"
			]
		}),
		e.wait && (t ? t(e.wait) : /* @__PURE__ */ (0, q.jsx)(gf, {
			wait: e.wait,
			...r ? { onViewOutputs: r } : {}
		})),
		e.events.length > 0 && /* @__PURE__ */ (0, q.jsx)(pf, {
			name: "events",
			title: "Events",
			children: /* @__PURE__ */ (0, q.jsx)("ul", {
				className: "space-y-1",
				children: e.events.map((e, t) => /* @__PURE__ */ (0, q.jsx)(wf, { event: e }, t))
			})
		}),
		e.exports.length > 0 && /* @__PURE__ */ (0, q.jsx)(pf, {
			name: "exports",
			title: "Exports",
			children: /* @__PURE__ */ (0, q.jsx)(Sf, { exports: e.exports })
		})
	] });
}
function Ef({ outputs: e }) {
	return /* @__PURE__ */ (0, q.jsx)("div", {
		className: "min-w-0 p-density-4",
		"data-step-section": "outputs",
		children: /* @__PURE__ */ (0, q.jsx)(v, { data: e })
	});
}
function Df({ result: e }) {
	if (e.kind === "none") return null;
	if (e.kind === "invalid") return /* @__PURE__ */ (0, q.jsx)("div", {
		"data-step-source": "",
		children: /* @__PURE__ */ (0, q.jsx)(pf, {
			name: "source",
			title: "Source",
			children: /* @__PURE__ */ (0, q.jsxs)(Of, { children: ["Malformed source detail: ", e.error] })
		})
	});
	let { source: t } = e, n = /* @__PURE__ */ (0, q.jsxs)("span", {
		className: "font-mono text-[11px] text-muted-foreground",
		title: t.file,
		children: [
			vf(t.file),
			":",
			t.line,
			"-",
			t.endLine
		]
	});
	return /* @__PURE__ */ (0, q.jsx)("div", {
		"data-step-source": "",
		children: /* @__PURE__ */ (0, q.jsxs)(pf, {
			name: "source",
			title: "Source",
			aside: n,
			children: [t.error ? /* @__PURE__ */ (0, q.jsxs)("div", {
				className: i("rounded border px-2 py-1 text-xs", ff),
				children: ["Source unavailable: ", t.error]
			}) : /* @__PURE__ */ (0, q.jsx)(b, {
				language: t.language,
				source: t.text ?? "",
				copyable: !0
			}), t.moreLines ? /* @__PURE__ */ (0, q.jsxs)("div", {
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
function Of({ children: e }) {
	return /* @__PURE__ */ (0, q.jsx)("div", {
		role: "alert",
		className: "rounded border border-red-500/60 bg-red-500/10 px-2 py-1 text-xs text-red-700 dark:text-red-400",
		children: e
	});
}
function kf({ error: e }) {
	return /* @__PURE__ */ (0, q.jsxs)(Of, { children: ["Malformed step detail: ", e] });
}
// ../../../clicky-ui/packages/ui/src/data/verification/stepSource.ts
var Af = (e) => typeof e == "object" && !!e && !Array.isArray(e), jf = class extends Error {};
function Mf(e, t, n, r) {
	let i = e[t];
	if (!n(i)) throw new jf(`detail.source.${t} must be ${r}`);
	return i;
}
function Nf(e, t, n, r) {
	return e[t] === void 0 ? void 0 : Mf(e, t, n, r);
}
var Pf = (e) => typeof e == "string", Ff = (e) => typeof e == "number" && Number.isFinite(e);
function If(e) {
	if (!Af(e) || e.source === void 0) return { kind: "none" };
	let t = e.source;
	if (!Af(t)) return {
		kind: "invalid",
		error: "detail.source must be an object"
	};
	try {
		let e = {
			file: Mf(t, "file", Pf, "a string"),
			line: Mf(t, "line", Ff, "a number"),
			endLine: Mf(t, "end_line", Ff, "a number")
		}, n = {
			language: Nf(t, "language", Pf, "a string"),
			text: Nf(t, "text", Pf, "a string"),
			moreLines: Nf(t, "more_lines", Ff, "a number"),
			error: Nf(t, "error", Pf, "a string")
		};
		for (let [t, r] of Object.entries(n)) r !== void 0 && (e[t] = r);
		return {
			kind: "source",
			source: e
		};
	} catch (e) {
		if (e instanceof jf) return {
			kind: "invalid",
			error: e.message
		};
		throw e;
	}
}
function Lf(e) {
	if (!Af(e) || !("source" in e)) return e;
	let { source: t, ...n } = e;
	return Object.keys(n).length > 0 ? n : void 0;
}
function Rf({ file: e, line: t, endLine: n }) {
	return `${e}:${t}-${n}`;
}
function zf(e) {
	let t = e.error ? `source unavailable: ${e.error}` : e.text ?? "", n = e.moreLines ? `\n… ${e.moreLines} more lines` : "";
	return [{
		title: `Source ${Rf(e)}`,
		...e.language && !e.error ? { lang: e.language } : {},
		body: t + n
	}];
}
// ../../../clicky-ui/packages/ui/src/data/verification/adapterViews.tsx
function Bf(e) {
	return e.failed ? "failed" : e.passed ? "passed" : "pending";
}
function Vf({ item: e }) {
	return /* @__PURE__ */ (0, q.jsxs)("li", {
		className: "flex items-start gap-1.5 text-xs",
		children: [/* @__PURE__ */ (0, q.jsx)(G, {
			outcome: Bf(e),
			className: "mt-0.5 text-sm"
		}), /* @__PURE__ */ (0, q.jsxs)("span", {
			className: "min-w-0",
			children: [e.name, e.message ? /* @__PURE__ */ (0, q.jsx)("span", {
				className: "block text-[11px] text-muted-foreground",
				children: e.message
			}) : null]
		})]
	});
}
function Hf({ node: e }) {
	let t = e.children && e.children.length > 0 ? e.children : [e];
	return /* @__PURE__ */ (0, q.jsx)("div", {
		className: "space-y-2 p-density-4",
		children: /* @__PURE__ */ (0, q.jsx)("ul", {
			className: "space-y-1",
			children: t.map((e, t) => /* @__PURE__ */ (0, q.jsx)(Vf, { item: e }, e.task_id ?? t))
		})
	});
}
function Uf(e) {
	if (e && typeof e == "object" && !Array.isArray(e)) {
		let t = e.cel_trace;
		if (typeof t == "string" && t.length > 0) return t;
	}
}
function Wf(e, t) {
	if (t !== void 0 && (e || typeof t != "object" || !t || Array.isArray(t))) return t;
}
function Gf({ node: e }) {
	if (e.passed) return null;
	let t = e.context, n = t?.cel_vars ?? {}, r = Uf(e.detail) || t?.cel_expression, i = !!r, a = Wf(i, t?.expected), o = Wf(i, t?.actual), s = a !== void 0 || o !== void 0;
	if (!r && Object.keys(n).length === 0 && !s) return null;
	let c = Zd(e.detail), l = c.kind === "step" ? c.step.wait : void 0, u = l && l.outcome !== "satisfied" ? nf(l) : "expression";
	return /* @__PURE__ */ (0, q.jsx)(Yu, {
		check: r,
		checkLabel: u,
		vars: n,
		expected: a,
		actual: o
	});
}
function Kf(...e) {
	let t = e.filter((e) => !!e);
	return t.filter((e, n) => !t.some((t, r) => r !== n && t.includes(e) && (t.length > e.length || r < n)));
}
function qf({ node: e, cause: t, error: n }) {
	let r = Kf(t, e.message, n);
	return r.length === 0 ? null : e.failed || e.timed_out || t ? /* @__PURE__ */ (0, q.jsx)(xf, { children: r.map((e, t) => /* @__PURE__ */ (0, q.jsx)("div", {
		className: t > 0 ? "mt-1 text-xs opacity-90" : void 0,
		children: e
	}, t)) }) : /* @__PURE__ */ (0, q.jsx)("p", {
		className: "whitespace-pre-wrap break-words text-sm",
		children: r.join("\n")
	});
}
function Jf({ node: e, renderWait: t }) {
	let n = e.context, r = n?.command ?? e.command, i = n?.cwd ?? e.work_dir, a = Zd(e.detail), { header: o, rest: s } = tf(e.detail), c = Lf(a.kind === "step" ? tf(a.step.rest).rest : s);
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "space-y-4 p-density-4",
		children: [
			/* @__PURE__ */ (0, q.jsx)(bf, {
				header: o,
				...a.kind === "step" ? { stepKind: a.step.step.kind } : {}
			}),
			/* @__PURE__ */ (0, q.jsx)(qf, {
				node: e,
				cause: a.kind === "step" ? a.step.wait?.cause : void 0,
				error: o.error
			}),
			r && /* @__PURE__ */ (0, q.jsxs)("div", {
				className: "min-w-0 space-y-1",
				children: [
					/* @__PURE__ */ (0, q.jsx)("div", {
						className: "text-[10px] uppercase tracking-wide text-muted-foreground",
						children: "Command"
					}),
					/* @__PURE__ */ (0, q.jsx)("code", {
						className: "block overflow-x-auto whitespace-pre rounded bg-muted px-2 py-1.5 text-xs",
						children: r
					}),
					/* @__PURE__ */ (0, q.jsxs)("div", {
						className: "flex flex-wrap items-center gap-2 text-[10px] text-muted-foreground",
						children: [i && /* @__PURE__ */ (0, q.jsxs)("span", { children: ["in ", i] }), typeof n?.exit_code == "number" && /* @__PURE__ */ (0, q.jsxs)("span", {
							className: n.exit_code === 0 ? void 0 : "text-red-600",
							children: ["exit ", n.exit_code]
						})]
					})
				]
			}),
			/* @__PURE__ */ (0, q.jsx)($u, { node: e }),
			/* @__PURE__ */ (0, q.jsx)(Gf, { node: e }),
			/* @__PURE__ */ (0, q.jsx)(Df, { result: If(e.detail) }),
			a.kind === "step" && /* @__PURE__ */ (0, q.jsx)(Tf, {
				step: a.step,
				...t ? { renderWait: t } : {}
			}),
			a.kind === "invalid" && /* @__PURE__ */ (0, q.jsx)(kf, { error: a.error }),
			c != null && /* @__PURE__ */ (0, q.jsxs)("details", {
				className: "text-muted-foreground",
				"data-raw-detail": "",
				children: [/* @__PURE__ */ (0, q.jsx)("summary", {
					className: "cursor-pointer text-[10px] uppercase tracking-wide",
					children: "Raw detail"
				}), /* @__PURE__ */ (0, q.jsx)("div", {
					className: "mt-1 text-foreground",
					children: /* @__PURE__ */ (0, q.jsx)(v, { data: c })
				})]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/verification/adapters.tsx
function Yf(e) {
	return e.framework === "checklist";
}
function Xf(e) {
	return e.failed ? "[ ]" : e.passed ? "[x]" : "[?]";
}
function Zf(e) {
	return { brief: [{
		title: "Criteria",
		body: (e.children && e.children.length > 0 ? e.children : [e]).map((e) => `- ${Xf(e)} ${e.name}${e.message ? ` — ${e.message}` : ""}`).join("\n")
	}] };
}
var Qf = {
	id: "verification-checklist",
	match: Yf,
	renderDetail: ({ node: e }) => /* @__PURE__ */ (0, q.jsx)(Hf, { node: e }),
	copy: ({ node: e }) => Zf(e)
};
function $f(e) {
	return e.context !== void 0 || Qd(e.detail) ? !0 : e.framework === "fixture" && (e.children?.length ?? 0) === 0;
}
function ep(e) {
	if (e && typeof e == "object" && !Array.isArray(e)) {
		let t = e.cel_trace;
		if (typeof t == "string" && t.length > 0) return t;
	}
}
function tp(e, t) {
	if (t !== void 0 && (e || typeof t != "object" || !t || Array.isArray(t))) return t;
}
function np(e) {
	return typeof e == "string" ? e : JSON.stringify(e, null, 2);
}
function rp(e, t) {
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
function ip(e, t) {
	let n = t?.cel_expression, r = ep(e.detail), i = !!(n || r), a = t?.cel_vars ?? {}, o = tp(i, t?.expected), s = tp(i, t?.actual), c = [
		n && `expression: ${n}`,
		r && `trace:\n${r}`,
		Object.keys(a).length > 0 && `vars: ${np(a)}`,
		o !== void 0 && `expected: ${np(o)}`,
		s !== void 0 && `actual: ${np(s)}`
	].filter(Boolean);
	return c.length > 0 ? [{
		title: "Assertion",
		body: c.join("\n")
	}] : [];
}
var ap = {
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
function op(e) {
	let t = we(e), n = e.context?.exit_code, r = e.message ?? (n ? `exit ${n}` : void 0);
	return `${t ? ap[t] : "·"} ${e.name}${r ? ` — ${r}` : ""}`;
}
function sp(e) {
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
function cp(e) {
	let t = [...e.summary ? [sp(e.summary)] : [], ...(e.children ?? []).map(op)];
	return t.length > 0 ? [{
		title: "Results",
		body: t.join("\n")
	}] : [];
}
function lp(e) {
	let t = e.context, n = t?.command ?? e.command, r = Zd(e.detail), i = If(e.detail);
	return {
		brief: [
			...rp(e, t),
			...ip(e, t),
			...i.kind === "source" ? zf(i.source) : [],
			...r.kind === "step" ? df(r.step) : [],
			...cp(e)
		],
		items: n ? [{
			label: "Copy command",
			icon: xe,
			value: n
		}] : []
	};
}
var up = /* @__PURE__ */ new Set([
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
function dp(e) {
	let t = e.framework ?? "";
	return t === "playwright" ? Qd(e.detail) : up.has(t);
}
var fp = {
	id: "verification-step-copy",
	match: dp,
	copy: ({ node: e }) => lp(e)
};
function pp(e = {}) {
	return kl([
		Qf,
		{
			id: "verification-fixture",
			match: $f,
			renderDetail: ({ node: t }) => /* @__PURE__ */ (0, q.jsx)(Jf, {
				node: t,
				...e
			}),
			detailAttempts: ({ node: e }) => {
				let t = Zd(e.detail);
				return t.kind === "step" ? t.step.attempts.map((n) => ({
					id: n.attempt,
					status: n.status,
					...n.durationMs === void 0 ? {} : { durationMs: n.durationMs },
					node: n.result ? wd(n.result) : {
						name: e.name,
						message: `Attempt ${n.attempt}: execution evidence was not stored in this report.`,
						detail: { step: t.step.step }
					}
				})) : [];
			},
			detailTabs: ({ node: e }) => {
				let t = Zd(e.detail);
				return t.kind !== "step" || t.step.outputs === void 0 ? [] : [{
					id: "outputs",
					label: "Outputs",
					render: () => /* @__PURE__ */ (0, q.jsx)(Ef, { outputs: t.step.outputs })
				}];
			},
			copy: ({ node: e }) => lp(e)
		},
		fp
	]);
}
export { yt as $, vo as A, hn as B, os as C, Co as D, Po as E, Kn as F, Dt as G, Kt as H, Wn as I, wt as J, Et as K, Un as L, fi as M, pi as N, Do as O, sr as P, bt as Q, Bn as R, Uc as S, Uo as T, Ut as U, ln as V, It as W, J as X, St as Y, xt as Z, Zc as _, id as a, ft as at, Jc as b, Zu as c, st as ct, Ml as d, it as dt, vt as et, kl as f, rt as ft, $c as g, nl as h, Ye as ht, yd as i, pt as it, go as j, Eo as k, $u as l, ot as lt, ll as m, Xe as mt, Dd as n, ht as nt, td as o, ut as ot, ul as p, et as pt, Tt as q, bd as r, mt as rt, ed as s, ct as st, pp as t, _t as tt, Wu as u, at as ut, Qc as v, es as w, qc as x, Yc as y, kn as z };
