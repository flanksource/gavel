import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-B4NxTvNR.js";
import { t as r } from "./react-dom-BBpWcFtU.js";
import { a as i, r as a, t as o } from "./button-BLoR2pgR.js";
import { c as s, i as c, n as l, o as u, r as d, s as f, t as p } from "./UiInfo-CCiCR0Fr.js";
import { B as m, C as h, F as g, G as _, J as v, L as y, M as b, P as x, Q as S, U as C, X as w, Y as T, a as E, at as D, c as O, d as ee, et as te, f as ne, i as re, it as k, j as A, l as ie, n as ae, r as oe, rt as j, t as se, u as ce } from "./SplitButton-w2X8sfgS.js";
import { r as le, t as ue } from "./use-theme-sEqHIo2i.js";
import { c as de, d as M, f as fe, i as pe, l as me, m as he, n as N, o as P, p as ge, t as _e, u as ve } from "./ErrorDetails-CbwPqfQb.js";
import { C as ye, D as be, E as xe, H as Se, I as Ce, L as we, M as F, O as Te, S as Ee, T as De, U as Oe, _ as ke, b as Ae, g as je, h as Me, j as Ne, k as Pe, t as Fe, v as Ie, w as Le, z as Re } from "./JsonSchemaForm-C40jV51n.js";
import { C as ze, D as Be, E as Ve, H as He, J as Ue, K as We, O as Ge, R as Ke, S as qe, T as Je, U as Ye, W as Xe, b as Ze, f as Qe, g as I, h as $e, k as et, q as tt, x as nt } from "./Markdown-s_A3Eic8.js";
import { $ as rt, A as it, B as at, D as ot, E as L, G as st, H as ct, J as lt, K as ut, M as dt, N as ft, O as pt, Q as mt, R as ht, S as gt, U as _t, V as vt, W as yt, Y as bt, b as xt, ct as St, et as Ct, f as wt, g as Tt, h as Et, i as Dt, j as Ot, k as kt, mt as At, n as jt, o as Mt, pt as Nt, q as Pt, r as Ft, t as It, tt as Lt, ut as Rt, v as zt, w as Bt, x as R, y as Vt, z as Ht } from "./locations-CUxJY9f0.js";
import { i as Ut, y as Wt } from "./Callout-BqeseRSA.js";
import { a as z, l as Gt, n as Kt, o as qt, s as Jt } from "./floating-ui.react-dom-CJoIjj1G.js";
// ../../../clicky-ui/packages/ui/src/data/AnsiHtml.tsx
var B = /* @__PURE__ */ e(t(), 1), V = n(), Yt = {
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
function Xt(e) {
	if (!Number.isInteger(e) || e < 0 || e > 255) return;
	if (e < 16) return Yt[String(e < 8 ? 30 + e : 90 + e - 8)]?.slice(6);
	if (e >= 232) {
		let t = 8 + (e - 232) * 10;
		return `rgb(${t},${t},${t})`;
	}
	let t = e - 16, n = (e) => e === 0 ? 0 : 55 + e * 40;
	return `rgb(${n(Math.floor(t / 36))},${n(Math.floor(t / 6) % 6)},${n(t % 6)})`;
}
function Zt(e, t) {
	for (let n = 0; n < t.length; n++) {
		let r = t[n];
		if (r === 0) e = {};
		else if (r === 38 || r === 48) {
			let i = r === 38 ? "color" : "background-color", a = t[++n], o;
			if (a === 2) {
				let e = t.slice(n + 1, n + 4);
				n += 3, e.length === 3 && e.every((e) => Number.isInteger(e) && e >= 0 && e <= 255) && (o = `rgb(${e.join(",")})`);
			} else if (a === 5) {
				let e = t[++n];
				e !== void 0 && (o = Xt(e));
			}
			o !== void 0 && (e[i] = o);
		} else if (r === 22) delete e["font-weight"], delete e.opacity;
		else if (r === 23) delete e["font-style"];
		else if (r === 24) delete e["text-decoration"];
		else if (r === 39) delete e.color;
		else if (r === 49) delete e["background-color"];
		else if (r !== void 0) {
			let t = r >= 40 && r <= 47 || r >= 100 && r <= 107, n = Yt[String(t ? r - 10 : r)];
			if (n) {
				let [r, i] = n.split(":");
				r && i && (e[t ? "background-color" : r] = i);
			}
		}
	}
	return e;
}
function Qt(e) {
	let t = [], n = /\x1b\[([0-9;]*)m/g, r = 0, i = {}, a = () => Object.entries(i).map(([e, t]) => `${e}:${t}`).join(";"), o;
	for (; (o = n.exec(e)) !== null;) o.index > r && t.push({
		text: e.slice(r, o.index),
		style: a()
	}), i = Zt(i, (o[1] ?? "").split(";").map(Number)), r = o.index + o[0].length;
	return r < e.length && t.push({
		text: e.slice(r),
		style: a()
	}), t;
}
function $t(e) {
	let t = {};
	for (let n of e.split(";")) {
		let [e, r] = n.split(":");
		if (!e || !r) continue;
		let i = e.replace(/-([a-z])/g, (e, t) => t.toUpperCase());
		t[i] = r;
	}
	return t;
}
function H({ text: e, className: t, as: n = "pre" }) {
	let r = Qt(e).map((e, t) => e.style ? /* @__PURE__ */ (0, V.jsx)("span", {
		style: $t(e.style),
		children: e.text
	}, t) : /* @__PURE__ */ (0, V.jsx)("span", { children: e.text }, t));
	return n === "span" ? /* @__PURE__ */ (0, V.jsx)("span", {
		className: t,
		children: r
	}) : /* @__PURE__ */ (0, V.jsx)("pre", {
		className: t,
		children: r
	});
}
// ../../../clicky-ui/packages/ui/src/icons/components/UiZoomIn.tsx
var en = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
var tn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
var nn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
var rn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M10 8a2 2 0 1 1-4 0a2 2 0 0 1 4 0"
	})
}), {
	__source: "codicon:debug-stackframe-dot",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "stack-frame-dot",
	displayName: "UiStackFrameDot"
}), an = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
var on = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiSync.tsx
var sn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M14 3.5v3c0 .28-.22.5-.5.5h-3c-.28 0-.5-.22-.5-.5s.22-.5.5-.5h2.08c-.8-1.83-2.57-3-4.58-3c-2.22 0-4.2 1.5-4.81 3.64c-.06.22-.26.36-.48.36c-.05 0-.09 0-.14-.02a.493.493 0 0 1-.34-.62C2.96 3.79 5.33 2 8 2c2.05 0 3.91 1.02 5 2.69V3.5c0-.28.22-.5.5-.5s.5.22.5.5m-.58 5.52a.51.51 0 0 0-.62.35a5.02 5.02 0 0 1-4.81 3.64c-2.01 0-3.78-1.17-4.58-3h2.08c.28 0 .5-.22.5-.5s-.22-.5-.5-.5h-3c-.28 0-.5.22-.5.5v3c0 .28.22.5.5.5s.5-.22.5-.5v-1.19a5.97 5.97 0 0 0 5 2.69c2.67 0 5.04-1.79 5.77-4.36a.5.5 0 0 0-.35-.62z"
	})
}), {
	__source: "codicon:sync",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "sync",
	displayName: "UiSync"
}), cn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M222 104a6 6 0 0 1-12 0V54.49l-69.75 69.75a6 6 0 0 1-8.48-8.48L201.51 46H152a6 6 0 0 1 0-12h64a6 6 0 0 1 6 6Zm-38 26a6 6 0 0 0-6 6v72a2 2 0 0 1-2 2H48a2 2 0 0 1-2-2V80a2 2 0 0 1 2-2h72a6 6 0 0 0 0-12H48a14 14 0 0 0-14 14v128a14 14 0 0 0 14 14h128a14 14 0 0 0 14-14v-72a6 6 0 0 0-6-6"
	})
}), {
	__source: "ph:arrow-square-out-light",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "link-external",
	displayName: "UiLinkExternal"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M192 136v72a16 16 0 0 1-16 16H48a16 16 0 0 1-16-16V80a16 16 0 0 1 16-16h72a8 8 0 0 1 0 16H48v128h128v-72a8 8 0 0 1 16 0m32-96a8 8 0 0 0-8-8h-64a8 8 0 0 0-5.66 13.66L172.69 72l-42.35 42.34a8 8 0 0 0 11.32 11.32L184 83.31l26.34 26.35A8 8 0 0 0 224 104Z"
	})
}), {
	__source: "ph:arrow-square-out-fill",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "link-external",
	displayName: "UiLinkExternalFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiArrowUp.tsx
var ln = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M204.24 116.24a6 6 0 0 1-8.48 0L134 54.49V216a6 6 0 0 1-12 0V54.49l-61.76 61.75a6 6 0 0 1-8.48-8.48l72-72a6 6 0 0 1 8.48 0l72 72a6 6 0 0 1 0 8.48"
	})
}), {
	__source: "ph:arrow-up-light",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "arrow-up",
	displayName: "UiArrowUp"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M207.39 115.06A8 8 0 0 1 200 120h-64v96a8 8 0 0 1-16 0v-96H56a8 8 0 0 1-5.66-13.66l72-72a8 8 0 0 1 11.32 0l72 72a8 8 0 0 1 1.73 8.72"
	})
}), {
	__source: "ph:arrow-up-fill",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "arrow-up",
	displayName: "UiArrowUpFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiArrowDown.tsx
var un = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "m204.24 148.24l-72 72a6 6 0 0 1-8.48 0l-72-72a6 6 0 0 1 8.48-8.48L122 201.51V40a6 6 0 0 1 12 0v161.51l61.76-61.75a6 6 0 0 1 8.48 8.48"
	})
}), {
	__source: "ph:arrow-down-light",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "arrow-down",
	displayName: "UiArrowDown"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "m205.66 149.66l-72 72a8 8 0 0 1-11.32 0l-72-72A8 8 0 0 1 56 136h64V40a8 8 0 0 1 16 0v96h64a8 8 0 0 1 5.66 13.66"
	})
}), {
	__source: "ph:arrow-down-fill",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "arrow-down",
	displayName: "UiArrowDownFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiExternalLink.tsx
var dn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M222 104a6 6 0 0 1-12 0V54.49l-69.75 69.75a6 6 0 0 1-8.48-8.48L201.51 46H152a6 6 0 0 1 0-12h64a6 6 0 0 1 6 6Zm-38 26a6 6 0 0 0-6 6v72a2 2 0 0 1-2 2H48a2 2 0 0 1-2-2V80a2 2 0 0 1 2-2h72a6 6 0 0 0 0-12H48a14 14 0 0 0-14 14v128a14 14 0 0 0 14 14h128a14 14 0 0 0 14-14v-72a6 6 0 0 0-6-6"
	})
}), {
	__source: "ph:arrow-square-out-light",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "external-link",
	displayName: "UiExternalLink"
}), fn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "m222.72 67.9l-88-48.17a13.9 13.9 0 0 0-13.44 0l-88 48.18A14 14 0 0 0 26 80.18v95.64a14 14 0 0 0 7.28 12.27l88 48.18a13.92 13.92 0 0 0 13.44 0l88-48.18a14 14 0 0 0 7.28-12.27V80.18a14 14 0 0 0-7.28-12.28M127 30.25a2 2 0 0 1 1.92 0L212.51 76L128 122.24L43.49 76ZM39 177.57a2 2 0 0 1-1-1.75V86.66l84 46V223Zm177.92 0L134 223v-90.36l84-46v89.16a2 2 0 0 1-1 1.77Z"
	})
}), {
	__source: "ph:cube-light",
	__viewBox: "0 0 256 256",
	__group: "layout-dashboard",
	__consumerName: "box",
	displayName: "UiBox"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "m223.68 66.15l-88-48.15a15.88 15.88 0 0 0-15.36 0l-88 48.17a16 16 0 0 0-8.32 14v95.64a16 16 0 0 0 8.32 14l88 48.17a15.88 15.88 0 0 0 15.36 0l88-48.17a16 16 0 0 0 8.32-14V80.18a16 16 0 0 0-8.32-14.03M128 120L47.65 76L128 32l80.35 44Zm8 99.64v-85.81l80-43.78v85.76Z"
	})
}), {
	__source: "ph:cube-fill",
	__viewBox: "0 0 256 256",
	__group: "layout-dashboard",
	__consumerName: "box",
	displayName: "UiBoxFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiDebug.tsx
var pn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M146 92a10 10 0 1 1 10 10a10 10 0 0 1-10-10m-46-10a10 10 0 1 0 10 10a10 10 0 0 0-10-10m146 46a6 6 0 0 1-6 6h-26v10a85.9 85.9 0 0 1-3.45 24.08l23.85 10.42a6 6 0 0 1-4.8 11l-23.23-10.15a86 86 0 0 1-156.74 0L26.4 189.5a6 6 0 1 1-4.8-11l23.85-10.42A85.9 85.9 0 0 1 42 144v-10H16a6 6 0 0 1 0-12h26v-10a85.9 85.9 0 0 1 3.45-24.08L21.6 77.5a6 6 0 0 1 4.8-11l23.23 10.15a86 86 0 0 1 156.74 0L229.6 66.5a6 6 0 1 1 4.8 11l-23.85 10.42A85.9 85.9 0 0 1 214 112v10h26a6 6 0 0 1 6 6m-192-6h148v-10a74 74 0 0 0-148 0Zm68 95.74V134H54v10a74.09 74.09 0 0 0 68 73.74M202 134h-68v83.74A74.09 74.09 0 0 0 202 144Z"
	})
}), {
	__source: "ph:bug-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "debug",
	displayName: "UiDebug"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M168 92a12 12 0 1 1-12-12a12 12 0 0 1 12 12m-68-12a12 12 0 1 0 12 12a12 12 0 0 0-12-12m116 64a87.8 87.8 0 0 1-3 23l22.24 9.72A8 8 0 0 1 232 192a7.9 7.9 0 0 1-3.2-.67L207.38 182a88 88 0 0 1-158.76 0l-21.42 9.33a7.9 7.9 0 0 1-3.2.67a8 8 0 0 1-3.2-15.33L43 167a87.8 87.8 0 0 1-3-23v-8H16a8 8 0 0 1 0-16h24v-8a87.8 87.8 0 0 1 3-23l-22.2-9.67a8 8 0 1 1 6.4-14.66L48.62 74a88 88 0 0 1 158.76 0l21.42-9.36a8 8 0 0 1 6.4 14.66L213 89.05a87.8 87.8 0 0 1 3 23v8h24a8 8 0 0 1 0 16h-24Zm-80 0a8 8 0 0 0-16 0v64a8 8 0 0 0 16 0Zm64-32a72 72 0 0 0-144 0v8h144Z"
	})
}), {
	__source: "ph:bug-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "debug",
	displayName: "UiDebugFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiTable.tsx
var mn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M224 50H32a6 6 0 0 0-6 6v136a14 14 0 0 0 14 14h176a14 14 0 0 0 14-14V56a6 6 0 0 0-6-6M38 110h44v36H38Zm56 0h124v36H94Zm124-48v36H38V62ZM38 192v-34h44v36H40a2 2 0 0 1-2-2m178 2H94v-36h124v34a2 2 0 0 1-2 2"
	})
}), {
	__source: "ph:table-light",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "table",
	displayName: "UiTable"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M224 48H32a8 8 0 0 0-8 8v136a16 16 0 0 0 16 16h176a16 16 0 0 0 16-16V56a8 8 0 0 0-8-8M40 112h40v32H40Zm56 0h120v32H96Zm-56 48h40v32H40Zm176 32H96v-32h120z"
	})
}), {
	__source: "ph:table-fill",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "table",
	displayName: "UiTableFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiClass.tsx
var hn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, V.jsx)("circle", {
			cx: "8",
			cy: "8",
			r: "6.5",
			fill: "#E7EFFD",
			stroke: "#3574F0"
		}), /* @__PURE__ */ (0, V.jsx)("path", {
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
}), gn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 32 32",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		d: "M4.014 14.976a2.5 2.5 0 0 0 1.567-.518a2.38 2.38 0 0 0 .805-1.358a15.3 15.3 0 0 0 .214-2.944q.012-2.085.075-2.747a5.2 5.2 0 0 1 .418-1.686a3 3 0 0 1 .755-1.018A3.05 3.05 0 0 1 9 4.125A6.8 6.8 0 0 1 10.544 4h.7v1.96h-.387a2.34 2.34 0 0 0-1.723.468a3.4 3.4 0 0 0-.425 2.092a36 36 0 0 1-.137 4.133a4.7 4.7 0 0 1-.768 2.06A4.6 4.6 0 0 1 6.1 16a3.8 3.8 0 0 1 1.992 1.754a8.9 8.9 0 0 1 .618 3.865q0 2.435.05 2.9a1.76 1.76 0 0 0 .504 1.181a2.64 2.64 0 0 0 1.592.337h.387V28h-.7a5.7 5.7 0 0 1-1.773-.2a2.97 2.97 0 0 1-1.324-.93a3.35 3.35 0 0 1-.681-1.63a24 24 0 0 1-.165-3.234a16.5 16.5 0 0 0-.214-3.106a2.4 2.4 0 0 0-.805-1.361a2.5 2.5 0 0 0-1.567-.524Zm23.972 2.035a2.5 2.5 0 0 0-1.567.524a2.4 2.4 0 0 0-.805 1.361a16.5 16.5 0 0 0-.212 3.109a24 24 0 0 1-.169 3.234a3.35 3.35 0 0 1-.681 1.63a2.97 2.97 0 0 1-1.324.93a5.7 5.7 0 0 1-1.773.2h-.7V26.04h.387a2.64 2.64 0 0 0 1.592-.337a1.76 1.76 0 0 0 .506-1.186q.05-.462.05-2.9a8.9 8.9 0 0 1 .618-3.865A3.8 3.8 0 0 1 25.9 16a4.6 4.6 0 0 1-1.7-1.286a4.7 4.7 0 0 1-.768-2.06a36 36 0 0 1-.137-4.133a3.4 3.4 0 0 0-.425-2.092a2.34 2.34 0 0 0-1.723-.468h-.387V4h.7a6.8 6.8 0 0 1 1.54.125a3.05 3.05 0 0 1 1.149.581a3 3 0 0 1 .755 1.018a5.2 5.2 0 0 1 .418 1.686q.062.662.075 2.747a15.3 15.3 0 0 0 .212 2.947a2.38 2.38 0 0 0 .805 1.355a2.5 2.5 0 0 0 1.567.518Z",
		fill: "#E5A505"
	})
}), {
	__source: "incumbent",
	__viewBox: "0 0 32 32",
	__group: "files-code",
	__consumerName: "json",
	displayName: "UiJson"
}), _n = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 32 32",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, V.jsx)("path", {
		d: "M2.5 7.955h27v16.091h-27z",
		fill: "none",
		stroke: "#755838"
	}), /* @__PURE__ */ (0, V.jsx)("path", {
		d: "M5.909 20.636v-9.272h2.727l2.728 3.409l2.727-3.409h2.727v9.272h-2.727v-5.318l-2.727 3.409l-2.728-3.409v5.318zm17.046 0l-4.091-4.5h2.727v-4.772h2.727v4.772h2.727z",
		fill: "#755838"
	})]
}), {
	__source: "incumbent",
	__viewBox: "0 0 32 32",
	__group: "files-code",
	__consumerName: "markdown",
	displayName: "UiMarkdown"
}), vn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
var yn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 24 24",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsxs)("g", {
		fill: "none",
		stroke: "currentColor",
		strokeLinecap: "round",
		strokeLinejoin: "round",
		strokeWidth: "2",
		children: [/* @__PURE__ */ (0, V.jsx)("path", { d: "M6 22a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h8a2.4 2.4 0 0 1 1.704.706l3.588 3.588A2.4 2.4 0 0 1 20 8v12a2 2 0 0 1-2 2z" }), /* @__PURE__ */ (0, V.jsx)("path", { d: "M14 2v5a1 1 0 0 0 1 1h5M8 13h2m4 0h2m-8 4h2m4 0h2" })]
	})
}), {
	__source: "lucide:file-spreadsheet",
	__viewBox: "0 0 24 24",
	__group: "files-code",
	__consumerName: "file-spreadsheet",
	displayName: "UiFileSpreadsheet"
}), bn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M230 64a30 30 0 1 0-36 29.4V112a10 10 0 0 1-10 10H96a21.84 21.84 0 0 0-10 2.42v-31a30 30 0 1 0-12 0v69.2a30 30 0 1 0 12 0V144a10 10 0 0 1 10-10h88a22 22 0 0 0 22-22V93.4A30.05 30.05 0 0 0 230 64M62 64a18 18 0 1 1 18 18a18 18 0 0 1-18-18m36 128a18 18 0 1 1-18-18a18 18 0 0 1 18 18M200 82a18 18 0 1 1 18-18a18 18 0 0 1-18 18"
	})
}), {
	__source: "ph:git-branch-light",
	__viewBox: "0 0 256 256",
	__group: "git-source-control",
	__consumerName: "git-branch",
	displayName: "UiGitBranch"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
		fill: "currentColor",
		d: "M232 64a32 32 0 1 0-40 31v17a8 8 0 0 1-8 8H96a23.8 23.8 0 0 0-8 1.38V95a32 32 0 1 0-16 0v66a32 32 0 1 0 16 0v-17a8 8 0 0 1 8-8h88a24 24 0 0 0 24-24V95a32.06 32.06 0 0 0 24-31M64 64a16 16 0 1 1 16 16a16 16 0 0 1-16-16m32 128a16 16 0 1 1-16-16a16 16 0 0 1 16 16"
	})
}), {
	__source: "ph:git-branch-fill",
	__viewBox: "0 0 256 256",
	__group: "git-source-control",
	__consumerName: "git-branch",
	displayName: "UiGitBranchFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiWrench.tsx
var xn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
var Sn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
var Cn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
var wn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiHourglass.tsx
var Tn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, V.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, V.jsx)("path", {
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
function En(e, t) {
	return t.split(".").reduce((e, t) => {
		if (e && typeof e == "object") return e[t];
	}, e);
}
function Dn(e) {
	if (Array.isArray(e)) return e.flatMap((e) => Dn(e)).filter(Boolean);
	if (e == null) return [];
	if (typeof e == "object") return [JSON.stringify(e)];
	let t = String(e).trim();
	return t ? [t] : [];
}
function On(e, t) {
	return t.accessor ? t.accessor(e) : En(e, t.key);
}
function kn(e, t) {
	let n = On(e, t);
	return t.filterValue ? t.filterValue(n, e) : n;
}
function An(e, t) {
	return Dn(kn(e, t));
}
function jn(e) {
	return e.replace(/([a-z])([A-Z])/g, "$1 $2").replace(/[_-]/g, " ").replace(/\b\w/g, (e) => e.toUpperCase()).trim();
}
function Mn(e) {
	return typeof e.label == "string" ? e.label : jn(e.key.split(".").at(-1) ?? e.key);
}
function Nn(e) {
	return e.hideable !== !1;
}
function Pn(e, t) {
	if (e.length === 0) return [];
	let n = t?.literalKeys ?? !1, r = /* @__PURE__ */ new Set();
	for (let t of e.slice(0, 20)) Object.keys(t).filter((e) => !e.startsWith("_")).forEach((e) => r.add(e));
	return Array.from(r).map((e, t) => ({
		key: e || `__col_${t}`,
		label: e ? jn(e) : "(no column name)",
		sortable: !0,
		filterable: !0,
		...n ? { accessor: (t) => t[e] } : {}
	}));
}
// ../../../clicky-ui/packages/ui/src/data/data-table-server-filters.ts
function Fn(e) {
	if (e.autoFilter && e.manualFilter) throw Error("DataTable: autoFilter generates filters that narrow rows locally, which manualFilter then suppresses — the controls would render and do nothing. Pass externalFilters instead.");
	if (e.manualFilter && e.showGlobalFilter && !e.globalFilterControlled && !e.hasExternalSearch) throw Error("DataTable: manualFilter suppresses the built-in search's row narrowing, so showGlobalFilter needs globalFilter/onGlobalFilterChange or externalSearch to reach the source.");
}
// ../../../clicky-ui/packages/ui/src/hooks/use-sort.ts
function In(e, t) {
	return t.split(".").reduce((e, t) => {
		if (e && typeof e == "object") return e[t];
	}, e);
}
function Ln(e, t = {}) {
	let { defaultKey: n, defaultDir: r = "asc", sort: i, onSortChange: a, manual: o = !1, resolvers: s } = t, [c, l] = (0, B.useState)(n ? {
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
		sorted: (0, B.useMemo)(() => {
			if (!e) return [];
			if (o || !d) return e;
			let { key: t, dir: n } = d, r = s?.[t], i = (e) => r ? r(e) : In(e, t);
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
function Rn({ density: e, children: t }) {
	let n = (0, B.useMemo)(() => ({
		density: e,
		setDensity: () => {}
	}), [e]);
	return /* @__PURE__ */ (0, V.jsx)(f.Provider, {
		value: n,
		children: t
	});
}
// ../../../clicky-ui/packages/ui/src/data/use-infinite-rows.ts
function zn(e, t) {
	let n = (0, B.useRef)(e);
	n.current = e;
	let r = (0, B.useRef)(!1), i = e?.loading ?? !1;
	return (0, B.useEffect)(() => {
		r.current = !1;
	}, [i, t]), (0, B.useCallback)(() => {
		let e = n.current;
		e && !e.loading && e.hasMore && !r.current && (r.current = !0, e.onLoadMore());
	}, []);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@tanstack/virtual-core/3.19.11/0511dfc93966ba3eb16991599b7c8f20894853d5c07be6f71ccfef3cbe6c89c4/node_modules/@tanstack/virtual-core/dist/esm/lazy-measurements.js
var Bn = /* @__PURE__ */ e(r(), 1);
function Vn(e) {
	return typeof e == "object" ? e.key : e;
}
function Hn(e, t) {
	let n = e.length;
	return new Proxy(e, { get(e, r, i) {
		if (typeof r == "string") {
			let i = r.charCodeAt(0);
			if (i >= 48 && i <= 57) {
				let i = +r;
				if (Number.isInteger(i) && i >= 0 && i < n) {
					let n = e[i];
					if (typeof n != "object") {
						let r = t[i * 2];
						n = e[i] = {
							index: i,
							key: n,
							start: r,
							size: t[i * 2 + 1],
							end: r + t[i * 2 + 1],
							lane: 0
						};
					}
					return n;
				}
			}
			if (r === "length") return n;
		}
		return Reflect.get(e, r, i);
	} });
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@tanstack/virtual-core/3.19.11/0511dfc93966ba3eb16991599b7c8f20894853d5c07be6f71ccfef3cbe6c89c4/node_modules/@tanstack/virtual-core/dist/esm/utils.js
function Un(e, t, n) {
	let r = n.initialDeps ?? [], i, a = !0;
	function o() {
		let o = e();
		return (o.length !== r.length || o.some((e, t) => r[t] !== e)) && (r = o, i = t(...o), n?.onChange && !(a && n.skipInitialOnChange) && n.onChange(i), a = !1), i;
	}
	return o.updateDeps = (e) => {
		r = e;
	}, o;
}
function U(e, t) {
	if (e === void 0) throw Error(`Unexpected undefined${t ? `: ${t}` : ""}`);
	return e;
}
var W = (e, t) => Math.abs(e - t) < 1.01, Wn = (e, t, n) => {
	let r;
	return Object.assign(function(...i) {
		e.clearTimeout(r), r = e.setTimeout(() => t.apply(this, i), n);
	}, { cancel: () => {
		e.clearTimeout(r);
	} });
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@tanstack/virtual-core/3.19.11/0511dfc93966ba3eb16991599b7c8f20894853d5c07be6f71ccfef3cbe6c89c4/node_modules/@tanstack/virtual-core/dist/esm/scroll-classifier.js
function Gn({ offset: e, isScrolling: t, elementMax: n, holdOffset: r, scrollOffset: i, pendingWrite: a, clampedMaxAtWrite: o }) {
	return t && a === null && e === i ? "spurious" : a !== null && Math.abs(e - a.target) < 1.5 ? "readBack" : a !== null && !a.smooth && n !== null && a.target > n + r + .5 && Math.abs(e - (n + r)) < 1.5 || a !== null && !a.smooth && a.lowMaxes?.some((t) => Math.abs(e - t) < 1.5) ? "clampedReadBack" : o !== null && Math.abs(e - o) < 1.5 ? "clampedEcho" : "external";
}
var Kn = (e) => e === "smooth" || e === "smoothFinishing";
function G({ kind: e, offset: t, command: n, target: r }) {
	return e !== "spurious" && e !== "readBack" && e !== "clampedReadBack" && n?.index != null && n.phase !== "queued" && !Kn(n.phase) && !W(t, n.lastTargetOffset) && !W(t, r()) && e !== "clampedEcho";
}
var qn = 40;
function Jn(e) {
	let t = e[e.length - 1];
	if (t === void 0) return !1;
	let n = e.length - 2;
	for (; n >= 0 && t.t - e[n].t < qn;) n--;
	if (n < 0) return !1;
	let r = e[n], i = n - 1;
	for (; i >= 0 && r.t - e[i].t < qn;) i--;
	if (i < 0) return !1;
	let a = e[i], o = (t.offset - r.offset) / (t.t - r.t), s = (r.offset - a.offset) / (r.t - a.t);
	return o * s >= 0 && Math.abs(o) >= 2 * Math.abs(s) && Math.abs(o) - Math.abs(s) >= .8;
}
function Yn(e) {
	let t = e[e.length - 1];
	if (t === void 0) return !1;
	let n = e.length - 2;
	for (; n >= 0 && t.t - e[n].t < qn;) n--;
	if (n < 0) return !1;
	let r = e[n];
	return Math.abs(t.offset - r.offset) / (t.t - r.t) >= .05;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@tanstack/virtual-core/3.19.11/0511dfc93966ba3eb16991599b7c8f20894853d5c07be6f71ccfef3cbe6c89c4/node_modules/@tanstack/virtual-core/dist/esm/ios-gesture.js
var Xn, Zn = () => {
	if (Xn !== void 0) return Xn;
	if (typeof navigator > "u") return Xn = !1;
	if (/iP(hone|od|ad)/.test(navigator.userAgent)) return Xn = !0;
	let e = navigator.maxTouchPoints;
	return Xn = navigator.platform === "MacIntel" && e !== void 0 && e > 0;
}, Qn, $n = () => Qn === void 0 ? Qn = Zn() && typeof CSS < "u" && typeof CSS.supports == "function" && CSS.supports("overflow-anchor: auto") : Qn, er = () => !Zn() || $n(), tr = {
	capture: !0,
	passive: !0
}, nr = class {
	constructor(e) {
		this.host = e, this._touch = null, this.inMotion = () => this._touch !== null && this._touch.phase !== "lateFling", this.dragging = () => this._touch?.phase === "down" && this._touch.moved, this.tail = () => this._touch?.phase === "tail" ? this._touch.moved ? "fling" : "tap" : null, this.onScroll = ({ isScrolling: e, ownReadBack: t, moved: n, external: r }) => (e && this._touch?.phase === "tail" ? this._openTail(this._touch.moved || !t, !1) : this._touch?.phase === "lateFling" && (this.host.now() > this._touch.until ? this._touch = null : r && e && n && this._openTail(!0, !1)), !t && e && n && this._touch?.phase === "down" ? (this._touch.moved = !0, !0) : !1), this.touchStart = () => {
			this.end(), this._touch = {
				phase: "down",
				moved: !1
			};
		}, this.touchEnd = () => {
			let e = this._touch?.phase === "down" && this._touch.moved;
			this.end(), this._openTail(e, e);
		}, this.end = () => {
			let e = this.host.getWindow();
			this._touch?.phase === "tail" && e != null && e.clearTimeout(this._touch.timerId), this._touch = null;
		}, this.listen = (e) => {
			let t = () => this.touchStart(), n = () => this.touchEnd();
			return e.addEventListener("touchstart", t, tr), e.addEventListener("touchend", n, tr), e.addEventListener("touchcancel", n, tr), () => {
				e.removeEventListener("touchstart", t, tr), e.removeEventListener("touchend", n, tr), e.removeEventListener("touchcancel", n, tr), this.end();
			};
		}, this._openTail = (e, t) => {
			let n = this.host.getWindow();
			if (n == null) return;
			this.end();
			let r = {
				phase: "tail",
				moved: e,
				awaitingFling: t,
				timerId: n.setTimeout(() => {
					this.host.tailStillMoving() ? this._openTail(r.moved, r.awaitingFling) : (this._touch = r.awaitingFling ? {
						phase: "lateFling",
						until: this.host.now() + 1e3
					} : null, this.host.onTailClosed());
				}, 150)
			};
			this._touch = r;
		};
	}
	get touch() {
		return this._touch;
	}
}, rr = (e) => {
	let { clientWidth: t, clientHeight: n } = e;
	return {
		width: t,
		height: n
	};
}, ir = (e) => {
	let t = /* @__PURE__ */ new Set();
	return {
		schedule: (n) => {
			let r = e.requestAnimationFrame(() => {
				t.delete(r), n();
			});
			t.add(r);
		},
		cancel: () => {
			t.forEach((t) => e.cancelAnimationFrame(t)), t.clear();
		}
	};
}, ar = (e) => e, or = (e) => {
	let t = e.overscan * Math.max(e.lanes ?? 1, 1), n = Math.max(e.startIndex - t, 0), r = Math.min(e.endIndex + t, e.count - 1) - n + 1, i = Array(r);
	for (let e = 0; e < r; e++) i[e] = n + e;
	return i;
}, sr = (e, t) => {
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
	if (i(rr(n)), !r.ResizeObserver) return () => {};
	let a = ir(r), o = new r.ResizeObserver(() => {
		let t = () => {
			i(rr(n));
		};
		e.options.useAnimationFrameWithResizeObserver ? a.schedule(t) : t();
	});
	return o.observe(n, { box: "content-box" }), () => {
		o.unobserve(n), a.cancel();
	};
}, cr = { passive: !0 }, lr = typeof window > "u" || "onscrollend" in window, ur = (e, t, n) => {
	let r = e.scrollElement;
	if (!r) return;
	let i = e.targetWindow;
	if (!i) return;
	let a = e.options.useScrollendEvent && lr, o = 0, s = a ? null : Wn(i, () => t(n(r), !1), e.options.isScrollingResetDelay), c = (e) => () => {
		o = n(r), s?.(), t(o, e);
	}, l = c(!0), u = c(!1);
	return r.addEventListener("scroll", l, cr), a && r.addEventListener("scrollend", u, cr), () => {
		r.removeEventListener("scroll", l), a && r.removeEventListener("scrollend", u), s?.cancel();
	};
}, dr = (e, t) => ur(e, t, (t) => {
	let { horizontal: n, isRtl: r } = e.options;
	return n ? t.scrollLeft * (r && -1 || 1) : t.scrollTop;
}), fr = (e, t, n) => {
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
}, pr = (e, { adjustments: t = 0, behavior: n }, r) => {
	var i;
	let { scrollElement: a, options: o } = r, s = e + t, c = o.horizontal && o.isRtl && a !== null && "ownerDocument" in a;
	(i = a?.scrollTo) == null || i.call(a, {
		[o.horizontal ? "left" : "top"]: c ? -s : s,
		behavior: n
	});
}, mr = (e, t) => e === "smooth" ? t ? "smoothLanded" : "smooth" : "instant", hr = (e) => e?.phase === "smooth" || e?.phase === "smoothLanded";
function gr(e, t, n, r) {
	if (t === 0) return !1;
	let i = r(0), a = /* @__PURE__ */ new Set(), o = 0;
	for (; o < e;) {
		let e = n(o);
		if (e === i) break;
		a.add(e), o++;
	}
	let s = e - o;
	if (s === 0 || s >= t) return !1;
	for (let e = 0; e < s; e++) if (r(e) !== n(o + e)) return !1;
	for (let e = s; e < t; e++) if (a.has(r(e))) return !1;
	return !0;
}
var _r = class {
	constructor(e) {
		this.unsubs = [], this.scrollElement = null, this.targetWindow = null, this.isScrolling = !1, this.scrollState = null, this.measurementsCache = [], this._singleLaneMeasurements = null, this.itemSizeCache = /* @__PURE__ */ new Map(), this.itemSizeCacheVersion = 0, this.laneAssignments = /* @__PURE__ */ new Map(), this.pendingMin = null, this.prevLanes = void 0, this.lanesChangedFlag = !1, this.lanesSettling = !1, this.pendingScrollAnchor = null, this.scrollRect = null, this.scrollOffset = null, this.scrollDirection = null, this._holdOffset = 0, this._reportedOverscroll = !1, this._iosGesture = new nr({
			getWindow: () => this.targetWindow,
			now: () => this.now(),
			tailStillMoving: () => {
				let e = this.scrollState;
				if (e?.phase !== "queued" || this.pendingWrite !== null) return !1;
				let t = this._readElementOffset();
				return t === null || this._lastElementOffset === null || Math.abs(t - this._lastElementOffset) < 1 ? !1 : (this._lastElementOffset = t, e.samples = [...e.samples ?? [], {
					t: this.now(),
					offset: t
				}], !0);
			},
			onTailClosed: () => {
				let e = this.scrollState;
				e?.phase === "queued" && Yn(e.samples ?? []) && (this.scrollState = null), this._flushHoldIfReady(!0), this._writeAwaitingCommand(), this._retryClampedAdjustment();
			}
		}), this._lastElementOffset = null, this.pendingWrite = null, this._clampedAdjustment = null, this.elementsCache = /* @__PURE__ */ new Map(), this._resizeBatch = null, this.now = () => {
			var e;
			return ((e = this.targetWindow?.performance)?.now)?.call(e) ?? Date.now();
		}, this.observer = /* @__PURE__ */ (() => {
			let e = null, t = null, n = !1, r = !1, i = /* @__PURE__ */ new Set(), a = () => {
				if (e) return e;
				let r = this.targetWindow;
				if (!r || !r.ResizeObserver) return null;
				let i = t = ir(r);
				return e = new r.ResizeObserver((e) => {
					let t = () => {
						let t = { sync: !1 };
						this._resizeBatch = t;
						try {
							e.forEach(o);
						} finally {
							this._resizeBatch = null, t.notify && (this.notify(t.sync), this._noteClampMax(), this._flushHoldIfReady(!0), this._retryClampedAdjustment(), this.measureElement(null));
						}
					};
					if (this.options.useAnimationFrameWithResizeObserver) i.schedule(t);
					else {
						n = !0;
						try {
							t();
						} finally {
							n = !1;
						}
					}
				});
			}, o = (e) => {
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
			return {
				disconnect: () => {
					var n;
					(n = a()) == null || n.disconnect(), t?.cancel(), i.clear(), r = !1, e = null, t = null;
				},
				observe: (o) => {
					let s = a();
					s && (n ? (i.add(o), r || (r = !0, t?.schedule(() => {
						r = !1, i.forEach((t) => {
							t.isConnected && e?.observe(t, { box: "border-box" });
						}), i.clear();
					}))) : s.observe(o, { box: "border-box" }));
				},
				unobserve: (e) => {
					var t;
					i.delete(e), (t = a()) == null || t.unobserve(e);
				}
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
				getItemKey: ar,
				rangeExtractor: or,
				onChange: () => {},
				measureElement: fr,
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
				let e = n.count, o = t.count, s = this.getMeasurements(), c = this._singleLaneMeasurements?.items ?? s, l = (e) => Vn(c[e]), u = e > 0 ? l(0) : null, d = e > 0 ? l(e - 1) : null;
				if (o !== e || e > 0 && o > 0 && (t.getItemKey(0) !== u || t.getItemKey(o - 1) !== d)) {
					a = !0;
					let c = e > 0 ? this.getVirtualItemForOffset(this.getScrollOffset()) ?? s[0] : null;
					c && (r = [c.key, this.getScrollOffset() - c.start]);
					let u = t.followOnAppend === !0 ? "auto" : t.followOnAppend || null;
					u && o > 0 && this.isAtEnd(n.scrollEndThreshold) && (e === 0 || t.getItemKey(o - 1) !== d) && (o > e || gr(e, o, l, t.getItemKey)) && (i = u);
				}
			}
			this.options = t, a && (this.pendingMin = 0, this.itemSizeCacheVersion++);
			let o = null;
			if (r && this.scrollOffset !== null && !this._awaitingCommandReadBack()) {
				let [e, t] = r, n = this.getMeasurements(), { count: a, getItemKey: s } = this.options, c = 0;
				for (; c < a && s(c) !== e;) c++;
				if (c < a) {
					let e = n[c];
					if (e) {
						let n = Math.max(0, e.start + t);
						!i && n !== this.scrollOffset && (o = n - this.scrollOffset, this.scrollOffset = n);
					}
				}
			}
			i ? this.pendingScrollAnchor = {
				kind: "follow",
				behavior: i
			} : o !== null && (this._canWriteScroll() ? this.pendingScrollAnchor = {
				kind: "anchor",
				delta: o
			} : (this._holdOffset += o, this._followOffsetTarget(o)));
		}, this.notify = (e) => {
			var t, n;
			(n = (t = this.options).onChange) == null || n.call(t, this, e);
		}, this.maybeNotify = Un(() => (this.calculateRange(), [
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
			var e;
			this.unsubs.filter(Boolean).forEach((e) => e()), this.unsubs = [], this.observer.disconnect(), this.rafId != null && this.targetWindow && (this.targetWindow.cancelAnimationFrame(this.rafId), this.rafId = null), this.scrollState = null, (e = this._momentumStop) == null || e.restore(), this._momentumStop = null, this.isScrolling = !1, this.scrollDirection = null, this._holdOffset = 0, this._reportedOverscroll = !1, this._clampedAdjustment = null, this.pendingWrite = null, this.scrollElement = null, this.targetWindow = null;
		}, this._didMount = () => () => {
			this.cleanup();
		}, this._willUpdate = () => {
			let e = this.options.enabled ? this.options.getScrollElement() : null;
			if (this.scrollElement !== e) {
				if (this.cleanup(), !e) {
					this.maybeNotify();
					return;
				}
				this.scrollElement = e, this.targetWindow = this.scrollElement && "ownerDocument" in this.scrollElement ? this.scrollElement.ownerDocument.defaultView : this.scrollElement?.window ?? null, this.elementsCache.forEach((e) => {
					this.observer.observe(e);
				}), this.unsubs.push(this.options.observeElementRect(this, (e) => {
					this.scrollRect = e, this.maybeNotify();
				})), this.unsubs.push(this.options.observeElementOffset(this, (e, t) => {
					let n = t ? this._elementMaxScrollOffset() : null, r = this._reportedOverscroll;
					this._reportedOverscroll = e < 0 || n !== null && e > n + 1;
					let i = e;
					this._lastElementOffset = i, e = this._toLayoutOffset(e);
					let a = this._classifyScrollEvent(e, t, n);
					if (a === "spurious") return;
					a === "readBack" ? e = this.pendingWrite.target : a === "clampedReadBack" && this.scrollState?.index != null && e < this.getMaxScrollOffset() - 1.5 ? this.scrollState.lastTargetOffset = e : G({
						kind: a,
						offset: e,
						command: this.scrollState,
						target: () => this.getScrollStateTarget(this.scrollState)
					}) && (this.scrollState = null), this.pendingWrite = null, this._clampedAdjustment !== null && Math.abs(e - this._clampedAdjustment.maxAtWrite) >= 1.5 && (this._clampedAdjustment = null);
					let o = this.getScrollOffset(), s = this.scrollDirection;
					this.scrollDirection = t ? o === e ? this.scrollDirection : o < e ? "forward" : "backward" : null, this.scrollOffset = e, this.isScrolling = t;
					let c = this.scrollState;
					if (c?.phase === "queued" && a === "external" && t) {
						if (r || this._reportedOverscroll) c.samples = [];
						else {
							let e = this.now();
							c.samples = [...(c.samples ?? []).filter((t) => e - t.t <= 200), {
								t: e,
								offset: i
							}], (s !== null && this.scrollDirection !== s || Jn(c.samples)) && (this.scrollState = null);
						}
					}
					this._iosGesture.onScroll({
						isScrolling: t,
						ownReadBack: a === "readBack",
						moved: e !== o,
						external: a === "external" && this.scrollState?.phase !== "smooth" && this.scrollState?.phase !== "smoothFinishing"
					}) && this.scrollState?.phase === "queued" && (this.scrollState = null), this._flushHoldIfReady(!0), this.scrollState && this.scheduleScrollReconcile(), this.maybeNotify();
				})), "addEventListener" in this.scrollElement && this.unsubs.push(this._iosGesture.listen(this.scrollElement)), this._scrollToOffset(this.getScrollOffset());
			}
			this._noteClampMax();
			let t = this.pendingScrollAnchor;
			this.pendingScrollAnchor = null, t && this.scrollElement && this.options.enabled && (t.kind === "anchor" ? this._canWriteScroll() ? hr(this.scrollState) && !W(this.getScrollOffset() - t.delta, this.scrollState.lastTargetOffset) || (this._followOffsetTarget(t.delta), this._scrollToOffset(this.getScrollOffset())) : (this._holdOffset += t.delta, this._followOffsetTarget(t.delta), this.notify(!1)) : this.scrollToEnd({ behavior: t.behavior })), this._flushHoldIfReady(!1), this._retryClampedAdjustment();
		}, this._classifyScrollEvent = (e, t, n) => Gn({
			offset: e,
			isScrolling: t,
			elementMax: n,
			holdOffset: this._holdOffset,
			scrollOffset: this.scrollOffset,
			pendingWrite: this.pendingWrite,
			clampedMaxAtWrite: this._clampedAdjustment?.maxAtWrite ?? null
		}), this._retryClampedAdjustment = () => {
			if (this._clampedAdjustment === null || !this.scrollElement || !this.options.enabled || !this._canWriteScroll()) return;
			let { target: e, maxAtWrite: t } = this._clampedAdjustment, n = this.getMaxScrollOffset();
			n > t + .5 && (this._clampedAdjustment = e > n + .5 ? {
				target: e,
				maxAtWrite: n
			} : null, this._scrollToOffset(e));
		}, this._writeAwaitingCommand = () => {
			let e = this.scrollState;
			if (e === null || e.phase !== "queued") return;
			let { behavior: t } = e;
			this.getMeasurements();
			let n = this.getScrollStateTarget(e);
			e.lastTargetOffset = n, this._clampedAdjustment = null, e.phase = mr(t, e.index != null && W(n, this.getScrollOffset())), this._scrollToOffset(n, { behavior: t }), this.scheduleScrollReconcile();
		}, this._readElementOffset = () => {
			let e = this.scrollElement;
			if (e == null) return null;
			let { horizontal: t, isRtl: n } = this.options;
			return typeof e.scrollTop == "number" ? t ? e.scrollLeft * (n && -1 || 1) : e.scrollTop : typeof e.scrollY == "number" ? t ? e.scrollX : e.scrollY : null;
		}, this._canWriteScroll = () => !(Zn() && this._iosGesture.inMotion()) && !this._reportedOverscroll, this._flushHoldIfReady = (e) => {
			if (this._holdOffset === 0 || !this.scrollElement || !this.options.enabled || !this._canWriteScroll() || this.scrollState?.phase === "smooth") return;
			this._holdOffset = 0, this.scrollOffset !== null && this.scrollOffset < 0 && (this.scrollOffset = 0), this.notify(e);
			let t = this.getCommandWrittenOffset() ?? this.getScrollOffset();
			this._clampedAdjustment = this._clampedBy(t), this._scrollToOffset(t);
		}, this.rafId = null, this.getSize = () => this.options.enabled ? (this.scrollRect = this.scrollRect ?? this.options.initialRect, this.scrollRect[this.options.horizontal ? "width" : "height"]) : (this.scrollRect = null, 0), this.getScrollOffset = () => this.options.enabled ? (this.scrollOffset = this.scrollOffset ?? (typeof this.options.initialOffset == "function" ? this.options.initialOffset() : this.options.initialOffset), this.scrollOffset) : (this.scrollOffset = null, 0), this.getMeasurementOptions = Un(() => [
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
		}), { key: !1 }), this.isIndexInRange = (e) => e >= 0 && e < this.options.count, this.getMeasurements = Un(() => [this.getMeasurementOptions(), this.itemSizeCacheVersion], ({ count: e, paddingStart: t, scrollMargin: n, getItemKey: r, enabled: i, lanes: a, laneAssignmentMode: o, gap: s }, c) => {
			let l = this.itemSizeCache;
			if (!i) return this.measurementsCache = [], this._singleLaneMeasurements = null, this.itemSizeCache.clear(), this.laneAssignments.clear(), [];
			if (this.laneAssignments.size > e) for (let t of this.laneAssignments.keys()) t >= e && this.laneAssignments.delete(t);
			this.lanesChangedFlag && (this.lanesChangedFlag = !1, this.lanesSettling = !0, this.measurementsCache = [], this._singleLaneMeasurements = null, this.itemSizeCache.clear(), this.laneAssignments.clear(), this.pendingMin = null), this.measurementsCache.length === 0 && !this.lanesSettling && (this.measurementsCache = this.options.initialMeasurementsCache, this.measurementsCache.forEach((e) => {
				this.itemSizeCache.set(e.key, e.size);
			}));
			let u = this.lanesSettling ? 0 : this.pendingMin ?? 0;
			if (this.pendingMin = null, this.lanesSettling && this.measurementsCache.length === e && (this.lanesSettling = !1), a === 1) {
				let i = e * 2, a = this._singleLaneMeasurements?.flat;
				if (!a || a.length < i) {
					let e = new Float64Array(i);
					a && u > 0 && e.set(a.subarray(0, u * 2)), a = e;
				}
				let o = u === 0 ? Array(e) : this._singleLaneMeasurements.items.slice(), c;
				if (u === 0) c = t + n;
				else {
					let e = u - 1;
					c = a[e * 2] + a[e * 2 + 1] + s;
				}
				for (let t = u; t < e; t++) {
					let e = r(t);
					o[t] = e;
					let n = l.get(e), i = typeof n == "number" ? n : this.options.estimateSize(t);
					a[t * 2] = c, a[t * 2 + 1] = i, c += i + s;
				}
				this._singleLaneMeasurements = {
					flat: a,
					items: o
				};
				let d = Hn(o, a);
				return this.measurementsCache = d, d;
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
		}), this.calculateRange = Un(() => [
			this.getMeasurements(),
			this.getSize(),
			this.getScrollOffset(),
			this.options.lanes,
			this._getLayoutTotalSize() + this.options.scrollMargin
		], (e, t, n, r, i) => {
			if (e.length === 0 || t === 0) return this.range = null, null;
			let a = Math.max(i - t, 0);
			return this.range = br(e, t, Math.min(n, a), r, r === 1 && this._singleLaneMeasurements !== null ? this._singleLaneMeasurements.flat : null), this.range;
		}, {
			key: !1,
			debug: () => this.options.debug
		}), this.getVirtualIndexes = Un(() => {
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
				t,
				this.options.lanes
			];
		}, (e, t, n, r, i, a) => r === null || i === null ? [] : e({
			startIndex: r,
			endIndex: i,
			overscan: t,
			count: n,
			lanes: a
		}), {
			key: !1,
			debug: () => this.options.debug
		}), this.indexFromElement = (e) => {
			let t = this.options.indexAttribute, n = e.getAttribute(t);
			return n ? parseInt(n, 10) : (console.warn(`Missing attribute name '${t}={index}' on measured element.`), -1);
		}, this.shouldMeasureDuringScroll = (e) => {
			if (!hr(this.scrollState)) return !0;
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
			r !== e && (r && this.observer.unobserve(r), this.observer.observe(e), this.elementsCache.set(n, e)), e.isConnected && (!this.isScrolling || this.scrollState) && this.shouldMeasureDuringScroll(t) && this.resizeItem(t, this.options.measureElement(e, void 0, this));
		}, this.resizeItem = (e, t) => {
			var n;
			if (!this.isIndexInRange(e)) return;
			this._resizeBatch?.sync && this.getMeasurements();
			let r, i, a, o = this._singleLaneMeasurements?.flat;
			if (this.options.lanes === 1 && o != null) a = this.options.getItemKey(e), i = o[e * 2], r = o[e * 2 + 1];
			else {
				let t = this.measurementsCache[e];
				if (!t) return;
				a = t.key, i = t.start, r = t.size;
			}
			let s = this.itemSizeCache.get(a) ?? r, c = t - s;
			if (c !== 0) {
				let o = this.options.anchorTo === "end" && !hr(this.scrollState) && this.getVirtualDistanceFromEnd() <= this.options.scrollEndThreshold, l = o ? this._getLayoutTotalSize() : 0, u = this.getScrollOffset(), d = this.itemSizeCache.has(a) ? i + s <= u && this.scrollDirection !== "backward" : i < u, f = !hr(this.scrollState) && (this.shouldAdjustScrollPositionOnItemSizeChange === void 0 ? d : this.shouldAdjustScrollPositionOnItemSizeChange(this.measurementsCache[e] ?? {
					index: e,
					key: a,
					start: i,
					size: r,
					end: i + r,
					lane: 0
				}, c, this));
				(this.pendingMin === null || e < this.pendingMin) && (this.pendingMin = e), this.itemSizeCache.set(a, t), this.itemSizeCacheVersion++;
				let p = !1;
				if (this._awaitingCommandReadBack() || (o ? p = this.applyScrollAdjustment(this._getLayoutTotalSize() - l) : f && (p = this.applyScrollAdjustment(c))), this._resizeBatch) {
					this._resizeBatch.notify = !0, (n = this._resizeBatch).sync || (n.sync = p);
					return;
				}
				this.notify(p), this._noteClampMax(), this._flushHoldIfReady(!0), this._retryClampedAdjustment();
			}
		}, this.getVirtualItems = Un(() => [
			this.getVirtualIndexes(),
			this.getMeasurements(),
			this._holdOffset
		], (e, t, n) => {
			let r = [];
			for (let i = 0, a = e.length; i < a; i++) {
				let a = t[e[i]];
				r.push(n === 0 ? a : {
					...a,
					start: a.start - n,
					end: a.end - n
				});
			}
			return r;
		}, {
			key: !1,
			debug: () => this.options.debug
		}), this.getVirtualItemForOffset = (e) => {
			let t = this.getMeasurements();
			if (t.length === 0) return;
			let n = this._singleLaneMeasurements?.flat, r = this.options.lanes === 1 && n != null;
			return U(t[vr(0, t.length - 1, r ? (e) => n[e * 2] : (e) => U(t[e]).start, e)]);
		}, this._noteClampMax = () => {
			if (!this._awaitingCommandReadBack()) return;
			let e = this._elementMaxScrollOffset();
			if (e === null) return;
			let t = this.pendingWrite, n = Math.max(e, 0) + this._holdOffset;
			if (n >= t.target) return;
			let r = t.lowMaxes ??= [];
			r[r.length - 1] !== n && r.push(n);
		}, this.getMaxScrollOffset = () => (this._elementMaxScrollOffset() ?? 0) + this._holdOffset, this._toLayoutOffset = (e) => {
			let t = e + this._holdOffset;
			return this.pendingScrollAnchor?.kind === "anchor" && (t += this.pendingScrollAnchor.delta), t;
		}, this._elementMaxScrollOffset = () => {
			let e = this.scrollElement;
			if (e === null) return null;
			if ("scrollHeight" in e) return this.options.horizontal ? e.scrollWidth - e.clientWidth : e.scrollHeight - e.clientHeight;
			if ("document" in e) {
				let t = e.document.documentElement;
				return this.options.horizontal ? t.scrollWidth - e.innerWidth : t.scrollHeight - e.innerHeight;
			}
			return null;
		}, this._momentumStop = null, this._stopMomentum = (e) => {
			let t = this.scrollElement, n = this.targetWindow;
			if (t === null || n === null) return;
			let r = "document" in t ? t.document.scrollingElement : t;
			if (!r?.style) return;
			if (this._momentumStop) {
				this._momentumStop.done = e;
				return;
			}
			let { style: i } = r, a = this.options.horizontal ? "overflow-x" : "overflow-y", o = i.getPropertyValue(a), s = i.getPropertyPriority(a), c = () => this.options.horizontal ? r.clientHeight : r.clientWidth, l = c();
			if (i.setProperty(a, "hidden", "important"), c() !== l) {
				i.setProperty(a, o, s);
				return;
			}
			let u = {
				restore: () => i.getPropertyValue(a) !== "hidden" || i.getPropertyPriority(a) !== "important" ? !1 : (i.setProperty(a, o, s), !0),
				done: e
			};
			this._momentumStop = u, n.requestAnimationFrame(() => n.requestAnimationFrame(() => {
				this._momentumStop === u && (this._momentumStop = null, u.restore() && u.done());
			}));
		}, this._clampedBy = (e) => {
			let t = this._elementMaxScrollOffset();
			if (t === null) return null;
			let n = t + this._holdOffset;
			return e > n + .5 ? {
				target: e,
				maxAtWrite: n
			} : null;
		}, this.getVirtualDistanceFromEnd = () => Math.max(this._getLayoutTotalSize() + this.options.scrollMargin - this.getSize() - this.getScrollOffset(), 0), this.getDistanceFromEnd = () => Math.max(this.getMaxScrollOffset() - this.getScrollOffset(), 0), this.isAtEnd = (e = this.options.scrollEndThreshold) => this.getDistanceFromEnd() <= e, this.getOffsetForAlignment = (e, t, n = 0) => {
			if (!this.scrollElement) return 0;
			let r = this.getSize(), i = this.getScrollOffset();
			t === "auto" && (t = e >= i + r ? "end" : "start"), t === "center" ? e += (n - r) / 2 : t === "end" && (e -= r);
			let a = this.getMaxScrollOffset();
			return Math.max(Math.min(a, e), 0);
		}, this.getOffsetForIndex = (e, t = "auto") => {
			e = Math.max(0, Math.min(e, this.options.count - 1));
			let n = this.getSize(), r = this.getScrollOffset(), i = this.getMeasurements()[e];
			if (!i) return;
			if (t === "auto") {
				if (i.end >= r + n - this.options.scrollPaddingEnd) t = "end";
				else if (i.start <= r + this.options.scrollPaddingStart) t = "start";
				else return [r, t];
			}
			if (t === "end" && e === this.options.count - 1) {
				let e = this.getMaxScrollOffset(), n = e - this.options.paddingEnd + this.options.scrollPaddingEnd;
				return [Math.max(Math.min(n, e), 0), t];
			}
			let a = t === "end" ? i.end + this.options.scrollPaddingEnd : i.start - this.options.scrollPaddingStart;
			return [this.getOffsetForAlignment(a, t, i.size), t];
		}, this.scrollToOffset = (e, { align: t = "start", behavior: n = "auto" } = {}) => {
			this._startScroll({
				index: null,
				align: t
			}, this.getOffsetForAlignment(e, t), n, "absolute");
		}, this.scrollToIndex = (e, { align: t = "auto", behavior: n = "auto" } = {}) => {
			e = Math.max(0, Math.min(e, this.options.count - 1));
			let r = this.getOffsetForIndex(e, t);
			if (!r) return;
			let [i, a] = r;
			this._startScroll({
				index: e,
				align: a
			}, i, n, "absolute");
		}, this.scrollBy = (e, { behavior: t = "auto" } = {}) => {
			let n = this.getOffsetForAlignment((this.getCommandWrittenOffset() ?? this.getScrollOffset()) + e, "start");
			this._startScroll({
				index: null,
				align: "start"
			}, n, t, "relative");
		}, this.scrollToEnd = ({ behavior: e = "auto" } = {}) => {
			this.options.count > 0 ? this._startScroll({
				index: this.options.count - 1,
				align: "end",
				toEnd: !0
			}, Math.max(this.getMaxScrollOffset(), 0), e, "absolute") : this.scrollToOffset(Math.max(this.getTotalSize() - this.getSize(), 0), { behavior: e });
		}, this.cancelScroll = () => {
			this.rafId != null && this.targetWindow && (this.targetWindow.cancelAnimationFrame(this.rafId), this.rafId = null), this.scrollState = null;
		}, this.getTotalSize = () => Math.max(this._getLayoutTotalSize() - this._holdOffset, 0), this._getLayoutTotalSize = () => {
			let e = this.getMeasurements(), t;
			if (e.length === 0) t = this.options.paddingStart;
			else if (this.options.lanes === 1) {
				let n = e.length - 1, r = this._singleLaneMeasurements?.flat;
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
		}, this._scrollToOffset = (e, { adjustments: t, behavior: n } = {}) => {
			let r = {
				target: e + (t ?? 0),
				smooth: n === "smooth"
			};
			return this.pendingWrite = r, this.options.scrollToFn(e - this._holdOffset, {
				behavior: n,
				adjustments: t
			}, this), r;
		}, this.getCommandWrittenOffset = () => this.scrollState !== null && this.pendingWrite !== null && !this.pendingWrite.smooth ? this.pendingWrite.target : null, this._awaitingCommandReadBack = () => {
			let e = this.getCommandWrittenOffset();
			return !(Zn() && this._iosGesture.dragging()) && e !== null && !W(e, this.getScrollOffset());
		}, this._followOffsetTarget = (e) => {
			this.scrollState !== null && this.scrollState.index == null && (this.scrollState.lastTargetOffset += e);
		}, this.measure = () => {
			this.pendingMin = null, this.itemSizeCache.clear(), this.laneAssignments.clear(), this.itemSizeCacheVersion++, this.notify(!1);
		}, this.setOptions(e);
	}
	get scrollAdjustments() {
		return 0;
	}
	applyScrollAdjustment(e, t) {
		if (e === 0) return !1;
		if (this._followOffsetTarget(e), this._canWriteScroll()) {
			let n = this.getScrollOffset();
			this._clampedAdjustment = this._clampedBy(n + e);
			let r = this._scrollToOffset(n, {
				adjustments: e,
				behavior: t
			});
			return this.scrollOffset !== null && (this.pendingWrite === r && (this.scrollOffset += e), this.scrollOffset < 0 && (this.scrollOffset = 0)), !0;
		}
		return this._holdOffset += e, this.scrollOffset !== null && (this.scrollOffset += e), !0;
	}
	scheduleScrollReconcile() {
		this.targetWindow ? this.rafId ??= this.targetWindow.requestAnimationFrame(() => {
			this.rafId = null, this.reconcileScroll();
		}) : this.scrollState = null;
	}
	getScrollStateTarget(e) {
		let t = e.toEnd ? [Math.max(this.getMaxScrollOffset(), 0)] : e.index == null ? void 0 : this.getOffsetForIndex(e.index, e.align);
		return t ? t[0] : e.lastTargetOffset;
	}
	reconcileScroll() {
		if (!this.scrollState || !this.scrollElement) return;
		if (this.now() - this.scrollState.startedAt > 5e3) {
			this.scrollState = null;
			return;
		}
		if (this.scrollState.phase === "queued") {
			this.scheduleScrollReconcile();
			return;
		}
		this.getMeasurements();
		let e = this.getScrollStateTarget(this.scrollState), t = this.scrollState.index == null ? 1 : this.options.useAnimationFrameWithResizeObserver ? 3 : 2, n = e !== this.scrollState.lastTargetOffset;
		if (!n && W(e, this.getScrollOffset())) {
			if (this.scrollState.phase === "smooth" ? this.scrollState.phase = "smoothLanded" : this.scrollState.phase === "smoothFinishing" && (this.scrollState.phase = "instant"), this.scrollState.stableFrames++, this.scrollState.stableFrames >= t) {
				this.getScrollOffset() !== e && this._scrollToOffset(e, { behavior: "auto" }), this.scrollState = null;
				return;
			}
		} else if (this.scrollState.stableFrames = 0, n) {
			let t = this.getSize() || 600, n = Math.abs(e - this.getScrollOffset()), r = hr(this.scrollState) && n > t;
			this.scrollState.lastTargetOffset = e, r ? this.scrollState.phase = "smooth" : this.scrollState.phase === "smooth" ? this.scrollState.phase = "smoothFinishing" : this.scrollState.phase === "smoothLanded" && (this.scrollState.phase = "instant"), this._scrollToOffset(e, { behavior: r ? "smooth" : "auto" });
		}
		this.scheduleScrollReconcile();
	}
	_startScroll(e, t, n, r) {
		this.pendingScrollAnchor = null, this._clampedAdjustment = null;
		let i = this._iosGesture.tail();
		if (r === "absolute" && i !== null) {
			if (i === "fling" && er()) {
				this._stopMomentum(() => {
					this.scrollState?.phase !== "queued" || this._iosGesture.tail() !== "fling" || this._reportedOverscroll || (this._iosGesture.end(), this._flushHoldIfReady(!0), this._writeAwaitingCommand());
				}), this.scrollState = {
					...e,
					startedAt: this.now(),
					lastTargetOffset: t,
					stableFrames: 0,
					phase: "queued",
					behavior: n,
					samples: []
				}, this.scheduleScrollReconcile();
				return;
			}
			this._iosGesture.end();
		}
		this.scrollState = {
			...e,
			startedAt: this.now(),
			lastTargetOffset: t,
			stableFrames: 0,
			behavior: n,
			phase: mr(n, e.index != null && W(t, this.getScrollOffset()))
		}, this._scrollToOffset(t, { behavior: n }), this.scheduleScrollReconcile();
	}
}, vr = (e, t, n, r) => {
	for (; e <= t;) {
		let i = (e + t) / 2 | 0, a = n(i);
		if (a < r) e = i + 1;
		else if (a > r) t = i - 1;
		else return i;
	}
	return e > 0 ? e - 1 : 0;
};
function yr(e, t, n) {
	let r = 0;
	for (; r <= t;) {
		let i = (r + t) / 2 | 0, a = e[i * 2];
		if (a < n) r = i + 1;
		else if (a > n) t = i - 1;
		else return i;
	}
	return r > 0 ? r - 1 : 0;
}
function br(e, t, n, r, i) {
	let a = e.length - 1;
	if (e.length <= r) return {
		startIndex: 0,
		endIndex: a
	};
	if (r === 1 && i !== null) {
		let e = yr(i, a, n), r = e, o = n + t;
		for (; r < a && i[r * 2] + i[r * 2 + 1] < o;) r++;
		return {
			startIndex: e,
			endIndex: r
		};
	}
	let o = vr(0, a, (t) => e[t].start, n), s = o;
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@tanstack/react-virtual/3.14.26/4c5b11c63540ab9342321fae6df54996727cb567588e246471cd5984756e5971/node_modules/@tanstack/react-virtual/dist/esm/index.js
var xr = typeof document < "u" ? B.useLayoutEffect : B.useEffect;
function Sr({ useFlushSync: e = !0, directDomUpdates: t = !1, directDomUpdatesMode: n = "transform", ...r }) {
	let i = B.useReducer((e) => e + 1, 0)[1], a = B.useRef({
		enabled: t,
		mode: n,
		container: null,
		lastSize: null,
		lastPositions: /* @__PURE__ */ new WeakMap(),
		prevRange: null
	});
	a.current.enabled = t, a.current.mode = n;
	let o = B.useRef(!1), s = (e) => {
		let t = a.current;
		if (!t.enabled || !t.container) return;
		let n = e.getTotalSize();
		if (n !== t.lastSize) {
			t.lastSize = n;
			let r = e.options.horizontal ? "width" : "height";
			t.container.style[r] = `${n}px`;
		}
	}, c = (e, t, n) => {
		let r = a.current, i = !!e.options.horizontal, o = t.start - e.options.scrollMargin;
		r.lastPositions.get(n) !== o && (r.lastPositions.set(n, o), r.mode === "transform" ? n.style.transform = i ? `translate3d(${o}px, 0, 0)` : `translate3d(0, ${o}px, 0)` : n.style[i ? "left" : "top"] = `${o}px`);
	}, l = (e) => {
		let t = a.current;
		if (t.enabled && t.container) {
			s(e);
			for (let t of e.getVirtualItems()) {
				let n = e.elementsCache.get(t.key);
				n && c(e, t, n);
			}
		}
	}, u = {
		...r,
		onChange: (t, n) => {
			var s;
			let c = a.current, u = !0;
			if (c.enabled) {
				l(t);
				let e = t.range, n = c.prevRange;
				u = !n || n.isScrolling !== t.isScrolling || n.startIndex !== e?.startIndex || n.endIndex !== e?.endIndex, u && (c.prevRange = e ? {
					startIndex: e.startIndex,
					endIndex: e.endIndex,
					isScrolling: t.isScrolling
				} : null);
			}
			u && (e && n && !o.current ? (0, Bn.flushSync)(i) : i()), (s = r.onChange) == null || s.call(r, t, n);
		}
	}, [d] = B.useState(() => {
		let e = new _r(u), t = e.measureElement;
		return e.measureElement = (n) => {
			o.current = !0;
			try {
				t(n);
			} finally {
				o.current = !1;
			}
			let r = a.current;
			if (n !== null && r.enabled && r.container) {
				let t = e.measurementsCache[e.indexFromElement(n)];
				t && c(e, t, n);
			}
		}, Object.assign(e, { containerRef: (t) => {
			let n = a.current;
			n.container = t, n.lastSize = null, t !== null && l(e);
		} });
	});
	return d.setOptions(u), xr(() => d._didMount(), []), xr(() => (s(d), d._willUpdate())), xr(() => {
		l(d);
	}), d;
}
function Cr(e) {
	return Sr({
		observeElementRect: sr,
		observeElementOffset: dr,
		scrollToFn: pr,
		...e
	});
}
// ../../../clicky-ui/packages/ui/src/data/use-data-table-virtual.ts
var wr = 8, K = 40;
function Tr(e) {
	if (!e) return null;
	let t = typeof e == "object" ? e : {};
	return {
		overscan: t.overscan ?? wr,
		estimateRowHeight: t.estimateRowHeight ?? K
	};
}
function Er(e, t, n) {
	let r = e[0], i = e[e.length - 1];
	return !r || !i ? {
		top: 0,
		bottom: 0
	} : {
		top: Math.max(0, r.start - n),
		bottom: Math.max(0, t - (i.end - n))
	};
}
function Dr({ count: e, options: t, getItemKey: n }) {
	let [r, i] = (0, B.useState)(null), [a, o] = (0, B.useState)(null), [s, c] = (0, B.useState)(0);
	(0, B.useEffect)(() => {
		if (!t || !a) return;
		let e = () => c(a.offsetTop);
		if (e(), typeof ResizeObserver > "u") return;
		let n = new ResizeObserver(e);
		return a.parentElement && n.observe(a.parentElement), () => n.disconnect();
	}, [t, a]);
	let l = (0, B.useCallback)(() => t?.estimateRowHeight ?? K, [t?.estimateRowHeight]), u = Cr({
		count: t ? e : 0,
		getScrollElement: () => r,
		estimateSize: l,
		overscan: t?.overscan ?? wr,
		getItemKey: n,
		scrollMargin: s
	}), d = t ? u.getVirtualItems() : [], f = (0, B.useMemo)(() => Er(d, u.getTotalSize(), s), [d, s]);
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
var Or = [
	{
		value: "compact",
		icon: v,
		label: "Compact"
	},
	{
		value: "comfortable",
		icon: we,
		label: "Comfortable"
	},
	{
		value: "spacious",
		icon: Se,
		label: "Spacious"
	}
];
function kr(e) {
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
function Ar({ actions: e, separated: t, onClose: n, size: r = "menu" }) {
	let a = kr(e), [o, s] = (0, B.useState)(null);
	return /* @__PURE__ */ (0, V.jsx)(V.Fragment, { children: a.map((e, a) => /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i((t || a > 0) && "mt-1 border-t border-border pt-1"),
		children: [e.section && /* @__PURE__ */ (0, V.jsx)("div", {
			className: "px-2 py-1.5 text-xs font-medium text-muted-foreground",
			children: e.section
		}), e.actions.map((e) => /* @__PURE__ */ (0, V.jsx)(jr, {
			action: e,
			submenu: o?.id === e.id ? o : null,
			onOpenSubmenu: s,
			onClose: n,
			size: r
		}, e.id))]
	}, e.section)) });
}
function jr({ action: e, submenu: t, onOpenSubmenu: n, onClose: r, size: a }) {
	let o = !!e.description, s = e.children ?? [], l = s.length > 0, u = (t) => {
		let r = t.getBoundingClientRect(), i = r.right + 224 > window.innerWidth ? r.left - 224 : r.right;
		n({
			id: e.id,
			x: i,
			y: r.top
		});
	};
	return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsxs)("button", {
		type: "button",
		role: "menuitem",
		"aria-haspopup": l ? "menu" : void 0,
		"aria-expanded": l ? t != null : void 0,
		disabled: e.disabled,
		className: i("flex w-full gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", o ? "items-start" : "items-center", a === "sheet" && "min-h-10 py-2 text-sm", e.disabled && "cursor-not-allowed opacity-50", t && "bg-accent text-accent-foreground"),
		onClick: (i) => {
			e.disabled || (l ? t ? n(null) : u(i.currentTarget) : (e.onSelect(), r()));
		},
		children: [
			e.icon && /* @__PURE__ */ (0, V.jsx)(c, {
				icon: e.icon,
				className: i("shrink-0 text-sm", o && "mt-0.5", e.iconClassName ?? "text-muted-foreground")
			}),
			/* @__PURE__ */ (0, V.jsxs)("span", {
				className: "min-w-0 flex-1",
				children: [/* @__PURE__ */ (0, V.jsx)("span", {
					className: i(o && "font-medium"),
					children: e.label
				}), e.description && /* @__PURE__ */ (0, V.jsx)("span", {
					className: "mt-0.5 block text-xs text-muted-foreground",
					children: e.description
				})]
			}),
			l && /* @__PURE__ */ (0, V.jsx)(c, {
				icon: j,
				className: "shrink-0 text-sm text-muted-foreground"
			})
		]
	}), l && t && /* @__PURE__ */ (0, V.jsx)("div", {
		role: "menu",
		"aria-label": typeof e.label == "string" ? e.label : "Submenu",
		className: "fixed z-50 max-h-[calc(100vh-1rem)] min-w-[14rem] max-w-[calc(100vw-1rem)] overflow-auto rounded-md border border-border bg-popover p-1.5 text-popover-foreground shadow-lg shadow-black/5",
		style: {
			left: t.x,
			top: t.y
		},
		children: s.map((e) => /* @__PURE__ */ (0, V.jsx)(jr, {
			action: e,
			submenu: null,
			onOpenSubmenu: n,
			onClose: r,
			size: a
		}, e.id))
	})] });
}
function Mr({ densityOverride: e, separated: t, onDensityChange: n, size: r = "menu" }) {
	let a = e ?? "inherit";
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i(t && "mt-1 border-t border-border pt-1"),
		children: [
			/* @__PURE__ */ (0, V.jsx)("div", {
				className: "px-2 py-1.5 text-xs font-medium text-muted-foreground",
				children: "Density"
			}),
			/* @__PURE__ */ (0, V.jsxs)("button", {
				type: "button",
				role: "menuitemradio",
				"aria-checked": a === "inherit",
				className: Nr(a === "inherit", r),
				onClick: () => n(void 0),
				children: [
					/* @__PURE__ */ (0, V.jsx)(c, {
						icon: on,
						className: "text-sm text-muted-foreground"
					}),
					/* @__PURE__ */ (0, V.jsx)("span", {
						className: "min-w-0 flex-1 truncate",
						children: "Use page density"
					}),
					a === "inherit" ? /* @__PURE__ */ (0, V.jsx)(c, {
						icon: w,
						className: "text-sm text-foreground"
					}) : /* @__PURE__ */ (0, V.jsx)("span", {
						className: "inline-block h-4 w-4",
						"aria-hidden": !0
					})
				]
			}),
			Or.map((e) => {
				let t = a === e.value;
				return /* @__PURE__ */ (0, V.jsxs)("button", {
					type: "button",
					role: "menuitemradio",
					"aria-checked": t,
					className: Nr(t, r),
					onClick: () => n(e.value),
					children: [
						/* @__PURE__ */ (0, V.jsx)(c, {
							icon: e.icon,
							className: "text-sm text-muted-foreground"
						}),
						/* @__PURE__ */ (0, V.jsx)("span", {
							className: "min-w-0 flex-1 truncate",
							children: e.label
						}),
						t ? /* @__PURE__ */ (0, V.jsx)(c, {
							icon: w,
							className: "text-sm text-foreground"
						}) : /* @__PURE__ */ (0, V.jsx)("span", {
							className: "inline-block h-4 w-4",
							"aria-hidden": !0
						})
					]
				}, e.value);
			})
		]
	});
}
function Nr(e, t) {
	return i("flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", t === "sheet" && "min-h-10 py-2 text-sm", e && "text-foreground");
}
var Pr = [
	{
		value: "system",
		icon: nn,
		label: "Use system theme"
	},
	{
		value: "light",
		icon: He,
		label: "Light"
	},
	{
		value: "dark",
		icon: Ye,
		label: "Dark"
	}
];
function Fr({ value: e, separated: t, onChange: n, size: r = "menu" }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i(t && "mt-1 border-t border-border pt-1"),
		children: [/* @__PURE__ */ (0, V.jsx)("div", {
			className: "px-2 py-1.5 text-xs font-medium text-muted-foreground",
			children: "Theme"
		}), Pr.map((t) => {
			let i = e === t.value;
			return /* @__PURE__ */ (0, V.jsxs)("button", {
				type: "button",
				role: "menuitemradio",
				"aria-checked": i,
				className: Nr(i, r),
				onClick: () => n(t.value),
				children: [
					/* @__PURE__ */ (0, V.jsx)(c, {
						icon: t.icon,
						className: "text-sm text-muted-foreground"
					}),
					/* @__PURE__ */ (0, V.jsx)("span", {
						className: "min-w-0 flex-1 truncate",
						children: t.label
					}),
					i ? /* @__PURE__ */ (0, V.jsx)(c, {
						icon: w,
						className: "text-sm text-foreground"
					}) : /* @__PURE__ */ (0, V.jsx)("span", {
						className: "inline-block h-4 w-4",
						"aria-hidden": !0
					})
				]
			}, t.value);
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/dataTableMenuPosition.ts
function Ir(e, t) {
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
function Lr(e, t) {
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
var Rr = 8;
function zr(e) {
	let { x: t, y: n, width: r, height: i } = e.reference, a = (0, B.useMemo)(() => ({ getBoundingClientRect: () => ({
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
	]), { refs: o, floatingStyles: s } = h({
		strategy: "fixed",
		placement: e.placement,
		transform: !1,
		whileElementsMounted: Gt,
		middleware: [
			z(6),
			Kt({ padding: Rr }),
			qt({ padding: Rr }),
			Jt({
				padding: Rr,
				apply({ availableHeight: e, elements: t }) {
					t.floating.style.maxHeight = `${Math.max(0, Math.floor(e))}px`;
				}
			})
		]
	});
	return (0, B.useLayoutEffect)(() => {
		o.setPositionReference(a);
	}, [o, a]), {
		ref: o.setFloating,
		style: s
	};
}
// ../../../clicky-ui/packages/ui/src/data/DataTablePreferencesSheet.tsx
function Br({ open: e, columns: t, hiddenColumns: n, activeColumnKey: r, actions: a, showColumnVisibilityControl: o, showDensityControl: s, showThemeControl: l, themeMenuValue: u, densityOverride: d, visibleHideableColumnCount: f, onToggle: p, onShowAll: m, onDensityChange: h, onThemeChange: g, onClose: _ }) {
	let v = r ? t.find((e) => e.key === r) : void 0, y = v && Nn(v) && f > 1, b = a.length > 0;
	return /* @__PURE__ */ (0, V.jsxs)(A, {
		open: e,
		onClose: _,
		title: "Table options",
		size: "sm",
		scrollBody: !0,
		children: [
			o && /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
				/* @__PURE__ */ (0, V.jsxs)("div", {
					className: "flex items-center justify-between gap-3 px-2 py-1.5 text-xs font-medium text-muted-foreground",
					children: [/* @__PURE__ */ (0, V.jsx)("span", { children: "Columns" }), /* @__PURE__ */ (0, V.jsx)("button", {
						type: "button",
						className: "min-h-10 rounded px-1.5 text-sm text-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none",
						onClick: m,
						children: "Show all"
					})]
				}),
				v && /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsxs)("button", {
					type: "button",
					disabled: !y,
					className: i("flex min-h-10 w-full items-center gap-2 rounded-md px-2 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", !y && "cursor-not-allowed opacity-50"),
					onClick: () => {
						y && (p(v), _());
					},
					children: [/* @__PURE__ */ (0, V.jsx)(c, {
						icon: Oe,
						className: "text-sm text-muted-foreground"
					}), /* @__PURE__ */ (0, V.jsxs)("span", { children: ["Hide ", Mn(v)] })]
				}), /* @__PURE__ */ (0, V.jsx)("div", { className: "my-1 h-px bg-border" })] }),
				/* @__PURE__ */ (0, V.jsx)("div", { children: t.map((e) => {
					let t = Nn(e), r = n[e.key] !== !0, a = !t || r && f <= 1;
					return /* @__PURE__ */ (0, V.jsxs)("label", {
						className: i("flex min-h-10 cursor-pointer items-center gap-2 rounded-md px-3 text-sm transition-colors hover:bg-accent hover:text-accent-foreground", a && "cursor-not-allowed opacity-50"),
						children: [/* @__PURE__ */ (0, V.jsx)("input", {
							type: "checkbox",
							className: "h-5 w-5 rounded border-border",
							checked: r,
							disabled: a,
							onChange: () => p(e)
						}), /* @__PURE__ */ (0, V.jsx)("span", {
							className: "truncate",
							children: Mn(e)
						})]
					}, e.key);
				}) })
			] }),
			s && /* @__PURE__ */ (0, V.jsx)(Mr, {
				densityOverride: d,
				separated: o,
				onDensityChange: h,
				size: "sheet"
			}),
			b && /* @__PURE__ */ (0, V.jsx)(Ar, {
				actions: a,
				separated: o || s,
				onClose: _,
				size: "sheet"
			}),
			l && g && /* @__PURE__ */ (0, V.jsx)(Fr, {
				value: u ?? "system",
				separated: o || s || b,
				onChange: g,
				size: "sheet"
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/SortableHeader.tsx
function Vr({ active: e, dir: t, onClick: n, align: r = "left", className: a, children: o }) {
	return /* @__PURE__ */ (0, V.jsxs)("button", {
		type: "button",
		onClick: n,
		className: i("inline-flex items-center gap-1 cursor-pointer select-none hover:text-foreground", e ? "text-primary" : "text-muted-foreground", r === "right" ? "justify-end" : r === "center" ? "justify-center" : "justify-start", a),
		children: [o, e ? /* @__PURE__ */ (0, V.jsx)("span", {
			"aria-hidden": !0,
			children: t === "asc" ? "↑" : "↓"
		}) : /* @__PURE__ */ (0, V.jsx)("span", {
			"aria-hidden": !0,
			className: "opacity-40",
			children: "↕"
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/cells/Timestamp.tsx
function Hr({ value: e, format: t, showTitleOnHover: n = !0, className: r }) {
	let a = ce(e);
	if (!a) return /* @__PURE__ */ (0, V.jsx)("span", {
		className: i("text-muted-foreground", r),
		children: "—"
	});
	let o = O(a, t), s = n ? a.toISOString() : void 0;
	return /* @__PURE__ */ (0, V.jsx)("span", {
		className: i("tabular-nums whitespace-nowrap text-muted-foreground", r),
		title: s,
		children: o
	});
}
// ../../../clicky-ui/packages/ui/src/data/cells/tag-utils.ts
function Ur(e, t = "=") {
	return e == null ? [] : Array.isArray(e) ? e.map((e) => Wr(e, t)).filter((e) => e.value !== "") : typeof e == "object" ? Object.entries(e).map(([e, n]) => {
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
function Wr(e, t) {
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
function Gr(e, t = "=") {
	return Ur(e, t).map((e) => e.token);
}
function Kr(e, t = "=") {
	let n = e.indexOf(t);
	return n <= 0 ? {
		key: "",
		value: e
	} : {
		key: e.slice(0, n),
		value: e.slice(n + t.length)
	};
}
var qr = (0, B.createContext)({
	getMode: () => "neutral",
	toggleInclude: () => {},
	toggleExclude: () => {}
});
function Jr() {
	return (0, B.useContext)(qr);
}
function Yr(e, t) {
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
function Xr({ value: e, children: t }) {
	return /* @__PURE__ */ (0, V.jsx)(qr.Provider, {
		value: e,
		children: t
	});
}
function Zr(e) {
	typeof navigator < "u" && navigator.clipboard?.writeText && navigator.clipboard.writeText(e).catch(() => void 0);
}
function Qr({ tag: e, actions: t, compact: n = !1 }) {
	let r = Jr(), a = r.getMode(e.token), o = (0, B.useCallback)((t) => {
		t.stopPropagation(), r.toggleInclude(e.token);
	}, [r, e.token]), s = (0, B.useCallback)((t) => {
		t.stopPropagation(), r.toggleExclude(e.token);
	}, [r, e.token]), c = (0, B.useCallback)((t) => {
		t.stopPropagation(), Zr(e.value);
	}, [e.value]), l = a === "include", u = a === "exclude", d = n && !!e.key, f = /* @__PURE__ */ (0, V.jsx)("span", {
		...d ? { title: e.display } : {},
		className: i("inline-flex items-center rounded-md border bg-background align-middle", l ? "border-green-500/50 bg-green-500/5" : u ? "border-red-500/50 bg-red-500/5" : "border-border"),
		children: e.key && !n ? /* @__PURE__ */ (0, V.jsx)(D, {
			size: "xs",
			variant: "label",
			tone: "neutral",
			label: e.key,
			value: e.value,
			maxWidth: 32,
			truncate: "auto",
			clickToCopy: !1,
			className: "border-transparent shadow-none"
		}) : /* @__PURE__ */ (0, V.jsx)(D, {
			size: "xs",
			variant: "soft",
			tone: "neutral",
			maxWidth: 32,
			truncate: "auto",
			clickToCopy: !1,
			children: e.value
		})
	}), p = /* @__PURE__ */ (0, V.jsxs)("span", {
		className: "inline-flex items-center gap-0.5",
		children: [
			/* @__PURE__ */ (0, V.jsx)($r, {
				ariaLabel: `Include ${e.display}`,
				onClick: o,
				active: l,
				activeClassName: "bg-green-500/20 text-green-700 dark:text-green-400",
				icon: en
			}),
			/* @__PURE__ */ (0, V.jsx)($r, {
				ariaLabel: `Exclude ${e.display}`,
				onClick: s,
				active: u,
				activeClassName: "bg-red-500/20 text-red-700 dark:text-red-400",
				icon: tn
			}),
			/* @__PURE__ */ (0, V.jsx)($r, {
				ariaLabel: `Copy ${e.display}`,
				onClick: c,
				icon: _
			})
		]
	});
	return t === "inline" ? /* @__PURE__ */ (0, V.jsxs)("span", {
		className: "inline-flex items-center gap-0.5 rounded-md border border-border bg-background pr-0.5",
		children: [f, p]
	}) : /* @__PURE__ */ (0, V.jsx)(oe, {
		placement: "top",
		delay: 120,
		arrow: !1,
		trigger: f,
		cardClassName: "!p-1",
		children: p
	});
}
function $r({ ariaLabel: e, onClick: t, active: n, activeClassName: r, icon: a }) {
	return /* @__PURE__ */ (0, V.jsx)("button", {
		type: "button",
		"aria-label": e,
		title: e,
		onClick: t,
		className: i("inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground", "hover:bg-accent hover:text-foreground", n && r),
		children: /* @__PURE__ */ (0, V.jsx)(c, {
			icon: a,
			className: "text-xs"
		})
	});
}
function ei({ tags: e, maxVisible: t = 3, className: n, actions: r = "hover", compact: a, wrap: o = !1 }) {
	let c = s(), l = a ?? c === "compact";
	if (e.length === 0) return /* @__PURE__ */ (0, V.jsx)("span", {
		className: "text-muted-foreground",
		children: "—"
	});
	if (o) return /* @__PURE__ */ (0, V.jsx)("span", {
		className: i("flex min-w-0 flex-wrap items-center gap-1", n),
		children: e.map((e, t) => /* @__PURE__ */ (0, V.jsx)(Qr, {
			tag: e,
			actions: r,
			compact: l
		}, `${e.display}-${t}`))
	});
	let u = e.slice(0, t), d = e.slice(t);
	return /* @__PURE__ */ (0, V.jsxs)("span", {
		className: i("flex min-w-0 items-center gap-1", n),
		children: [/* @__PURE__ */ (0, V.jsx)("span", {
			className: "flex min-w-0 flex-1 flex-nowrap items-center gap-1 overflow-hidden",
			children: u.map((e, t) => /* @__PURE__ */ (0, V.jsx)(Qr, {
				tag: e,
				actions: r,
				compact: l
			}, `${e.display}-${t}`))
		}), d.length > 0 && /* @__PURE__ */ (0, V.jsx)(oe, {
			placement: "top",
			trigger: /* @__PURE__ */ (0, V.jsxs)(D, {
				size: "xs",
				variant: "outlined",
				tone: "neutral",
				clickToCopy: !1,
				className: "shrink-0",
				children: ["+", d.length]
			}),
			cardClassName: "min-w-72 max-w-[90vw] whitespace-normal !p-0",
			children: /* @__PURE__ */ (0, V.jsx)(ti, { tags: d })
		})]
	});
}
function ti({ tags: e, rowClassName: t, gridTemplateColumns: n }) {
	let r = Jr(), i = (0, B.useMemo)(() => e.map((e, t) => ({
		key: `${e.token}-${t}`,
		value: e
	})), [e]), a = e.some((e) => !!e.key), o = n ?? (a ? "minmax(4rem, 8rem) minmax(0, 1fr)" : "0 minmax(0, 1fr)");
	return /* @__PURE__ */ (0, V.jsx)(Ne, {
		items: i,
		density: "compact",
		className: "border-0 bg-transparent",
		...t ? { rowClassName: t } : {},
		gridTemplateColumns: o,
		renderLabel: (e, t) => t.key ? /* @__PURE__ */ (0, V.jsx)("span", {
			className: "font-mono",
			children: t.key
		}) : null,
		renderValue: (e, t) => /* @__PURE__ */ (0, V.jsx)("span", {
			className: "block truncate font-mono",
			title: t.value,
			children: t.value
		}),
		suffixActions: [
			{
				id: "include",
				icon: en,
				label: (e, t) => `Include ${t.display}`,
				onClick: (e, t) => r.toggleInclude(t.token)
			},
			{
				id: "exclude",
				icon: tn,
				label: (e, t) => `Exclude ${t.display}`,
				onClick: (e, t) => r.toggleExclude(t.token)
			},
			{
				id: "copy",
				icon: _,
				label: (e, t) => `Copy ${t.display}`,
				onClick: (e, t) => Zr(t.value)
			}
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/cells/CellFilterActions.tsx
function ni({ children: e, value: t, displayValue: n = t, mode: r, onChange: i }) {
	return /* @__PURE__ */ (0, V.jsx)(oe, {
		placement: "top",
		delay: 120,
		arrow: !1,
		trigger: /* @__PURE__ */ (0, V.jsx)("span", {
			className: "relative inline-flex min-w-0",
			children: e
		}),
		cardClassName: "!p-1",
		children: /* @__PURE__ */ (0, V.jsxs)("span", {
			className: "inline-flex items-center gap-0.5",
			children: [/* @__PURE__ */ (0, V.jsx)(ri, {
				label: `Include ${n}`,
				icon: en,
				active: r === "include",
				activeClassName: "bg-green-500/20 text-green-700 dark:text-green-400",
				onClick: () => i(r === "include" ? void 0 : "include")
			}), /* @__PURE__ */ (0, V.jsx)(ri, {
				label: `Exclude ${n}`,
				icon: tn,
				active: r === "exclude",
				activeClassName: "bg-red-500/20 text-red-700 dark:text-red-400",
				onClick: () => i(r === "exclude" ? void 0 : "exclude")
			})]
		})
	});
}
function ri({ label: e, icon: t, active: n, activeClassName: r, onClick: a }) {
	return /* @__PURE__ */ (0, V.jsx)("button", {
		type: "button",
		"aria-label": e,
		"aria-pressed": n,
		title: e,
		onClick: (e) => {
			e.stopPropagation(), a();
		},
		className: i("inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground", "hover:bg-accent hover:text-foreground", n && r),
		children: /* @__PURE__ */ (0, V.jsx)(c, {
			icon: t,
			className: "text-xs"
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/cells/StatusDot.tsx
var ii = {
	success: "bg-emerald-500",
	error: "bg-rose-500",
	warning: "bg-amber-500",
	info: "bg-sky-500"
}, ai = {
	xs: "h-1.5 w-1.5",
	sm: "h-2 w-2",
	md: "h-2.5 w-2.5"
};
function oi({ status: e, size: t = "sm", label: n, title: r, className: a }) {
	let o = r ?? n ?? e;
	return /* @__PURE__ */ (0, V.jsxs)("span", {
		className: i("inline-flex items-center gap-1.5 align-middle", a),
		title: o,
		children: [/* @__PURE__ */ (0, V.jsx)("span", {
			role: "img",
			"aria-label": n ?? e,
			className: i("inline-block shrink-0 rounded-full", ii[e], ai[t])
		}), n && /* @__PURE__ */ (0, V.jsx)("span", {
			className: "truncate text-xs",
			children: n
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/cells/status-mapping.ts
var si = /* @__PURE__ */ new Set([
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
]), ci = /* @__PURE__ */ new Set([
	"warn",
	"warning",
	"degraded",
	"slow",
	"stale"
]), li = /* @__PURE__ */ new Set([
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
function ui(e) {
	if (e == null) return null;
	if (typeof e == "boolean") return e ? "success" : "error";
	let t = String(e).toLowerCase().trim();
	return t ? si.has(t) ? "error" : ci.has(t) ? "warning" : li.has(t) ? "success" : null : null;
}
// ../../../clicky-ui/packages/ui/src/data/DataTable.grouping.ts
var di = "\0empty";
function fi(e, t) {
	return e === "\0empty" ? t ?? "(empty)" : e;
}
function pi(e) {
	if (e == null || e === "") return di;
	switch (typeof e) {
		case "string":
		case "number":
		case "boolean":
		case "bigint": return String(e);
		default: throw Error(`DataTable automatic grouping requires scalar values; received ${Array.isArray(e) ? "array" : typeof e}`);
	}
}
function mi(e, t) {
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
function hi(e, t) {
	return typeof t == "function" ? t(e.key, e.rows) : t === !0;
}
// ../../../clicky-ui/packages/ui/src/data/DataTableGroupingControls.tsx
function gi({ modes: e, value: t, hasGroups: n, allExpanded: r, allCollapsed: i, onModeChange: a, onExpandAll: o, onCollapseAll: s }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "flex shrink-0 items-center gap-1",
		role: "group",
		"aria-label": "Grouping controls",
		children: [
			/* @__PURE__ */ (0, V.jsx)("div", {
				className: "w-44 shrink-0",
				children: /* @__PURE__ */ (0, V.jsx)(Qe, {
					"aria-label": "Group rows by",
					value: t,
					options: e,
					onChange: (e) => a(e.target.value)
				})
			}),
			/* @__PURE__ */ (0, V.jsx)(_i, {
				label: "Expand all groups",
				pressed: r,
				disabled: !n,
				onClick: o,
				icon: tt
			}),
			/* @__PURE__ */ (0, V.jsx)(_i, {
				label: "Collapse all groups",
				pressed: i,
				disabled: !n,
				onClick: s,
				icon: Ue
			})
		]
	});
}
function _i({ label: e, pressed: t, disabled: n, onClick: r, icon: a }) {
	return /* @__PURE__ */ (0, V.jsx)("button", {
		type: "button",
		"aria-label": e,
		title: e,
		"aria-pressed": t,
		disabled: n,
		onClick: r,
		className: i("inline-flex h-control-h w-control-h items-center justify-center rounded-md border border-border bg-background text-muted-foreground hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50", t && "bg-accent text-accent-foreground"),
		children: /* @__PURE__ */ (0, V.jsx)(c, {
			icon: a,
			className: "text-sm"
		})
	});
}
function vi(e) {
	return e.display ? e.display : e.menu ? "menu" : e.primary === !1 || e.section !== void 0 || e.children?.length ? "overflow" : "button";
}
function yi(e, t) {
	let n = [], r = [];
	for (let t of e) {
		let e = vi(t);
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
function bi({ count: e, onClear: t, compact: n }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		role: "status",
		"data-testid": "data-table-selection-count",
		className: i("flex items-center gap-density-2 text-xs text-muted-foreground", n ? "shrink-0 whitespace-nowrap" : "flex-wrap"),
		children: [/* @__PURE__ */ (0, V.jsxs)("span", {
			className: "font-medium text-foreground",
			children: [e.toLocaleString(), " selected"]
		}), n ? /* @__PURE__ */ (0, V.jsx)(o, {
			type: "button",
			variant: "ghost",
			size: "icon",
			className: "size-7",
			"aria-label": "Clear selection",
			title: "Clear selection",
			onClick: t,
			children: /* @__PURE__ */ (0, V.jsx)(c, { icon: d })
		}) : /* @__PURE__ */ (0, V.jsx)(o, {
			type: "button",
			variant: "outline",
			size: "sm",
			onClick: t,
			children: "Clear selection"
		})]
	});
}
var xi = "Cancel";
function Si({ action: e, context: t, onCancel: n, onConfirm: r }) {
	let i = typeof e.confirm == "object" ? e.confirm : {}, a = t.selectedScope?.total ?? t.selectedRowIds.length, s = typeof i.message == "function" ? i.message(t) : i.message ?? `This will apply to ${a.toLocaleString()} selected ${a === 1 ? "row" : "rows"}.`, c = i.confirmLabel ?? (typeof e.label == "string" ? e.label : "Confirm");
	return /* @__PURE__ */ (0, V.jsx)(A, {
		open: !0,
		size: "sm",
		expandable: !1,
		onClose: n,
		title: i.title ?? e.label,
		footer: /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex justify-end gap-density-2",
			children: [/* @__PURE__ */ (0, V.jsx)(o, {
				type: "button",
				variant: "ghost",
				size: "sm",
				onClick: n,
				children: i.cancelLabel ?? xi
			}), /* @__PURE__ */ (0, V.jsx)(o, {
				type: "button",
				size: "sm",
				variant: e.variant === "destructive" ? "destructive" : "default",
				onClick: r,
				children: c
			})]
		}),
		children: /* @__PURE__ */ (0, V.jsx)("p", {
			className: "text-sm text-muted-foreground",
			children: s
		})
	});
}
function Ci({ actions: e, context: t, showCount: n = !0, maxButtons: r, compact: i, className: a }) {
	let s = le(g("sm")), [l, u] = (0, B.useState)(null), [d, f] = (0, B.useState)(null), p = (0, B.useRef)(!1), m = (0, B.useCallback)((e) => {
		if (p.current || e.disabled) return;
		p.current = !0, u(e.id);
		let n = () => {
			p.current = !1, u(null);
		}, r;
		try {
			r = e.onSelect(t);
		} catch (e) {
			throw n(), e;
		}
		r instanceof Promise ? r.finally(n) : n();
	}, [t]), h = (0, B.useCallback)((e) => {
		e.confirm ? f(e) : m(e);
	}, [m]), { menus: _, buttons: v, overflow: y } = yi(e, { maxButtons: r }), b = [..._, ...v], x = l !== null, S = y.find((e) => e.id === l), C = (e) => ({
		label: e.label,
		onSelect: () => h(e),
		disabled: e.disabled || x,
		...e.icon ? { icon: e.icon } : {},
		...e.section === void 0 ? {} : { group: e.section },
		...e.children?.length ? { children: e.children.map(C) } : {}
	}), w = /* @__PURE__ */ new Map();
	for (let e of y) {
		let t = e.section ?? "", n = w.get(t);
		n ? n.push(e) : w.set(t, [e]);
	}
	let T = [...w.values()].flat().map(C);
	return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
		n ? /* @__PURE__ */ (0, V.jsx)(bi, {
			count: t.selectedScope?.total ?? t.selectedRowIds.length,
			onClear: t.clearSelection,
			compact: i ?? s
		}) : null,
		/* @__PURE__ */ (0, V.jsxs)("span", {
			className: a,
			children: [b.map((e) => {
				let n = l === e.id, r = _.includes(e), i = /* @__PURE__ */ (0, V.jsxs)(o, {
					type: "button",
					size: "sm",
					variant: e.variant ?? "outline",
					disabled: e.disabled || x && !n,
					loading: n,
					...r ? { "aria-haspopup": "menu" } : { onClick: () => h(e) },
					children: [
						e.icon && !n ? /* @__PURE__ */ (0, V.jsx)(c, {
							icon: e.icon,
							...e.iconClassName ? { className: e.iconClassName } : {}
						}) : null,
						n && e.pendingLabel !== void 0 ? e.pendingLabel : e.label,
						r ? /* @__PURE__ */ (0, V.jsx)(c, { icon: k }) : null
					]
				});
				if (!r) return /* @__PURE__ */ (0, V.jsx)("span", { children: i }, e.id);
				let a = e.menu ? { children: (n) => e.menu({
					context: t,
					close: n,
					busy: x,
					run: (t) => h({
						...e,
						...t.pendingLabel === void 0 ? {} : { pendingLabel: t.pendingLabel },
						...t.confirm === void 0 ? {} : { confirm: t.confirm },
						onSelect: t.onSelect
					})
				}) } : { items: (e.children ?? []).map(C) };
				return /* @__PURE__ */ (0, V.jsx)(ne, {
					align: "left",
					menuLabel: typeof e.label == "string" ? e.label : e.id,
					trigger: i,
					...a
				}, e.id);
			}), y.length > 0 ? /* @__PURE__ */ (0, V.jsx)(ne, {
				align: "right",
				menuLabel: "Selection actions",
				items: T,
				trigger: /* @__PURE__ */ (0, V.jsxs)(o, {
					type: "button",
					size: "sm",
					variant: "outline",
					"aria-haspopup": "menu",
					"aria-label": "More selection actions",
					disabled: x && !S,
					loading: !!S,
					children: [
						S ? null : /* @__PURE__ */ (0, V.jsx)(c, { icon: he }),
						"More",
						/* @__PURE__ */ (0, V.jsx)(c, { icon: k })
					]
				})
			}) : null]
		}),
		d ? /* @__PURE__ */ (0, V.jsx)(Si, {
			action: d,
			context: t,
			onCancel: () => f(null),
			onConfirm: () => {
				let e = d;
				f(null), m(e);
			}
		}) : null
	] });
}
// ../../../clicky-ui/packages/ui/src/data/DataTableOverflowCue.tsx
var wi = {
	overflow: !1,
	atStart: !0,
	atEnd: !0,
	interacted: !1,
	headerHeight: 0
};
function Ti({ scroll: e }) {
	let [t, n] = (0, B.useState)(wi);
	return (0, B.useEffect)(() => {
		if (!e) {
			n(wi);
			return;
		}
		let t = () => {
			let t = e.scrollWidth - e.clientWidth, r = e.scrollLeft <= 1;
			n((n) => ({
				overflow: t > 1,
				atStart: r,
				atEnd: t <= 1 || e.scrollLeft >= t - 1,
				interacted: n.interacted || !r,
				headerHeight: e.querySelector("thead")?.offsetHeight ?? 0
			}));
		};
		t(), e.addEventListener("scroll", t, { passive: !0 }), window.addEventListener("resize", t);
		let r = typeof ResizeObserver > "u" ? void 0 : new ResizeObserver(t);
		r?.observe(e);
		let i = e.querySelector("table");
		return i && r?.observe(i), () => {
			e.removeEventListener("scroll", t), window.removeEventListener("resize", t), r?.disconnect();
		};
	}, [e]), t.overflow ? /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
		t.atStart ? null : /* @__PURE__ */ (0, V.jsx)("span", {
			"aria-hidden": !0,
			"data-testid": "data-table-overflow-left",
			className: "pointer-events-none absolute inset-y-0 left-0 z-20 w-6 bg-gradient-to-r from-background to-transparent"
		}),
		t.atEnd ? null : /* @__PURE__ */ (0, V.jsx)("span", {
			"aria-hidden": !0,
			"data-testid": "data-table-overflow-right",
			className: "pointer-events-none absolute inset-y-0 right-0 z-20 w-8 bg-gradient-to-l from-background to-transparent"
		}),
		t.atStart && !t.interacted && t.headerHeight > 0 ? /* @__PURE__ */ (0, V.jsx)("span", {
			"data-testid": "data-table-overflow-prompt",
			className: "pointer-events-none absolute right-2 z-30 flex items-center",
			style: {
				top: 0,
				height: t.headerHeight
			},
			children: /* @__PURE__ */ (0, V.jsxs)("span", {
				className: i("inline-flex max-h-full items-center gap-1 rounded-full leading-none", "border border-border bg-background/95 px-2 py-0.5 text-[10px] font-medium text-muted-foreground shadow-sm"),
				children: ["More columns", /* @__PURE__ */ (0, V.jsx)(Rt, {
					"aria-hidden": !0,
					className: "size-3"
				})]
			})
		}) : null
	] }) : null;
}
// ../../../clicky-ui/packages/ui/src/data/DataTableList.tsx
function Ei({ title: e, status: t, meta: n }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		"data-list-item": "",
		className: "flex min-w-0 flex-col gap-0.5",
		children: [/* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex min-w-0 items-start gap-2",
			children: [/* @__PURE__ */ (0, V.jsx)("div", {
				"data-list-title": "",
				className: "min-w-0 flex-1 break-words font-medium",
				children: e
			}), t != null && t !== !1 ? /* @__PURE__ */ (0, V.jsx)("div", {
				"data-list-status": "",
				className: "flex shrink-0 items-center self-center",
				children: t
			}) : null]
		}), n.length > 0 ? /* @__PURE__ */ (0, V.jsx)("div", {
			className: "flex min-w-0 flex-wrap items-center gap-x-1.5 gap-y-0.5 text-xs text-muted-foreground",
			children: n.map((e, t) => /* @__PURE__ */ (0, V.jsxs)(B.Fragment, { children: [t > 0 ? /* @__PURE__ */ (0, V.jsx)("span", {
				"aria-hidden": !0,
				children: "·"
			}) : null, /* @__PURE__ */ (0, V.jsxs)("span", {
				"data-list-meta": e.key,
				title: e.label,
				className: "inline-flex min-w-0 items-center",
				children: [/* @__PURE__ */ (0, V.jsxs)("span", {
					className: "sr-only",
					children: [e.label, ": "]
				}), /* @__PURE__ */ (0, V.jsx)("span", {
					"data-list-meta-value": "",
					className: "inline-flex min-w-0 items-center gap-1",
					children: e.content
				})]
			})] }, e.key))
		}) : null]
	});
}
// ../../../clicky-ui/packages/ui/src/data/data-table-list.ts
function Di(e, t) {
	let n = e.filter((e) => e.listRole === t);
	if (n.length > 1) throw Error(`DataTable list layout: columns ${n.map((e) => `"${e.key}"`).join(", ")} all declare listRole "${t}"; only one may`);
	return n[0];
}
function Oi(e) {
	let t = Di(e, "title") ?? e.find((e) => e.listRole === void 0) ?? e.find((e) => e.listRole === "meta");
	if (!t) throw Error(`DataTable list layout needs a title column; visible columns ${e.map((e) => `"${e.key}"`).join(", ")} are all status or hidden`);
	let n = Di(e, "status") ?? e.find((e) => e !== t && e.listRole === void 0 && e.kind === "status");
	return {
		title: t,
		status: n,
		meta: e.filter((e) => e !== t && e !== n && (e.listRole === "meta" || e.listRole === void 0))
	};
}
// ../../../clicky-ui/packages/ui/src/data/DataTable.tsx
var ki = 64, Ai = 160, ji = 224, Mi = 96, Ni = 288, Pi = 256, Fi = "clicky-ui-data-table-column-widths", Ii = "clicky-ui-data-table-column-visibility", Li = "clicky-ui-data-table-density", Ri = "px-2.5 py-1.5 density-compact:px-2 density-compact:py-1 density-comfortable:px-2.5 density-comfortable:py-1.5 density-spacious:px-4 density-spacious:py-3", zi = "px-2.5 py-1.5 density-compact:px-2 density-compact:py-0.5 density-comfortable:px-2.5 density-comfortable:py-1.5 density-spacious:px-4 density-spacious:py-3";
function Bi(e, t) {
	return e ? e.cursor ? e.cursor.current === void 0 && e.cursor.next === void 0 : e.page === 0 ? e.hasMore == null ? e.total != null && e.totalRelation !== "gte" ? e.total <= e.pageSize : t < e.pageSize : !e.hasMore : !1 : !0;
}
function Vi({ grouping: e, groupingModes: t, groupingMode: n, defaultGroupingMode: r, onGroupingModeChange: i }) {
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
function Hi({ data: e, error: t, columns: n, loading: r = !1, loadingMessage: o = "Loading results…", loadingRowCount: s = 8, emptyMessage: c = "No data", virtualize: l, className: u, scrollContainerClassName: d, autoFilter: f = !1, boundedFilters: p, onBoundedFiltersChange: m, showGlobalFilter: h, manualFilter: _ = !1, globalFilter: v, onGlobalFilterChange: y, globalFilterPlaceholder: S = "Search all columns…", defaultSort: C, sort: w, onSortChange: T, manualSort: D = !1, filterBarProps: O, externalSearch: te, externalTimeRange: ne, externalFilters: k, cellFilters: ae, onCellFilterChange: oe, pagination: j, infinite: se, scrollContainerRef: ue, getRowId: de, rowSelection: M, selectionActions: fe, selectionBar: pe = "toolbar", getRowClassName: me, footer: he, grouping: N, groupingModes: P, groupingMode: ge, defaultGroupingMode: _e, onGroupingModeChange: ve, onRowClick: ye, isRowClickable: be, getRowHref: xe, renderExpandedRow: Se, expandedRowIds: Ce, onExpandedRowIdsChange: we, renderCard: F, layout: Te = "table", listBreakpoint: Ee = x.sm, toolbarPanel: De, detailStyle: Oe = "row", detailDialogTitle: ke, detailDialogSize: Ae = "lg", resizableColumns: je = !0, persistColumnWidths: Me = !0, columnResizeStorageKey: Ne, hideableColumns: Pe = !0, persistColumnVisibility: Fe = !0, columnVisibilityStorageKey: Ie, defaultHiddenColumns: Le, density: Re, defaultDensity: ze, onDensityChange: Be, persistDensity: Ve = !0, densityStorageKey: He, showDensityControl: Ue, themeMenuValue: We, onThemeChange: Ge, showThemeControl: Ke = !1, showHeaderFilters: qe = !0, menuActions: Ye }) {
	if (Vi({
		grouping: N,
		groupingModes: P,
		groupingMode: ge,
		defaultGroupingMode: _e,
		onGroupingModeChange: ve
	}), F && Te !== "table") throw Error("DataTable renderCard cannot be combined with layout=\"list\" or \"auto\"");
	let { ref: Xe, wider: Ze } = _t(Ee), Qe = Te === "list" || Te === "auto" && !Ze, I = (0, B.useMemo)(() => n.map((e) => typeof e == "string" ? {
		key: e,
		label: e
	} : e), [n]), $e = (0, B.useMemo)(() => I.map((e) => e.key).join("|"), [I]), et = Ne ?? `${Fi}:${$e}`, tt = Ie ?? `${Ii}:${$e}`, nt = He ?? `${Li}:${$e}`, [rt, it] = (0, B.useState)(() => Me ? Sa(et, I) : {}), at = (Le ?? []).join("|"), ot = (0, B.useMemo)(() => new Set(Le ?? []), [at]), [L, st] = (0, B.useState)(() => Fe ? wa(tt, I, ot) : Aa({}, I, ot)), ct = Re !== void 0, [ut, dt] = (0, B.useState)(() => ct ? Re : Ve ? Ea(nt) ?? ze : ze), [ft, pt] = (0, B.useState)(null), [mt, ht] = (0, B.useState)(null), gt = le(g("sm")), [vt, yt] = (0, B.useState)({}), [xt, St] = (0, B.useState)({}), [Ct, wt] = (0, B.useState)({}), Tt = p ?? Ct, Et = (e, t) => {
		let n = ca(Tt, e, t);
		p ? m?.(n) : wt(n);
	}, [Dt, Ot] = (0, B.useState)({}), [kt, At] = (0, B.useState)(null), [jt, Mt] = (0, B.useState)(_e), [Nt, Pt] = (0, B.useState)(""), [Ft, It] = (0, B.useState)(() => {
		let e = I.find((e) => e.kind === "timestamp" && e.timestamp?.defaultRange);
		return {
			from: e?.timestamp?.defaultRange?.from ?? "",
			to: e?.timestamp?.defaultRange?.to ?? ""
		};
	});
	(0, B.useEffect)(() => {
		it((e) => {
			let t = ka({
				...Me ? Sa(et, I) : {},
				...e
			}, I);
			return ja(e, t) ? e : t;
		});
	}, [
		$e,
		Me,
		et
	]), (0, B.useEffect)(() => {
		Me && Ca(et, rt);
	}, [
		rt,
		Me,
		et
	]), (0, B.useEffect)(() => {
		st((e) => {
			let t = Aa({
				...Fe ? wa(tt, I, ot) : {},
				...e
			}, I, ot);
			return Ma(e, t) ? e : t;
		});
	}, [
		$e,
		ot,
		Fe,
		tt
	]), (0, B.useEffect)(() => {
		Fe && Ta(tt, L, ot);
	}, [
		ot,
		L,
		Fe,
		tt
	]), (0, B.useEffect)(() => {
		ct || dt(Ve ? Ea(nt) ?? ze : ze);
	}, [
		ze,
		ct,
		Ve,
		nt
	]), (0, B.useEffect)(() => {
		!ct && Ve && Da(nt, ut);
	}, [
		ct,
		ut,
		Ve,
		nt
	]);
	let Lt = (0, B.useCallback)(() => {
		pt(null), ht(null);
	}, []), Rt = !!ft && gt;
	b(!!(ft || mt), Lt, !Rt), (0, B.useEffect)(() => {
		if ((ft || mt) && !Rt) return document.addEventListener("click", Lt), () => {
			document.removeEventListener("click", Lt);
		};
	}, [
		Lt,
		ft,
		Rt,
		mt
	]);
	let zt = (0, B.useMemo)(() => e.map((e, t) => ({
		id: de?.(e, t) ?? String(t),
		row: e
	})), [e, de]), Bt = (0, B.useMemo)(() => {
		let t = {};
		for (let n of I) {
			if (n.kind !== "timestamp") continue;
			let r = [];
			for (let t of e) {
				let e = On(t, n), i = ce(e);
				i && r.push(i);
			}
			let i = E(r);
			t[n.key] = ie(n.timestamp?.mode ?? "auto", i);
		}
		return t;
	}, [I, e]), R = (0, B.useMemo)(() => I.map((e) => _a(e, Bt[e.key])), [I, Bt]), Vt = ge ?? jt ?? _e ?? P?.[0]?.value, Ut = P?.find((e) => e.value === Vt), Wt = (0, B.useMemo)(() => {
		if (!Ut) return N;
		if (Ut.type === "none") return;
		if (Ut.type === "custom") return Ut;
		let e = R.find((e) => e.key === Ut.columnKey);
		if (!e) throw Error(`DataTable grouping mode "${Ut.value}" references missing column "${Ut.columnKey}"`);
		return {
			...Ut,
			getGroupKey: (t) => pi(On(t, e)),
			getGroupLabel: (e, t) => {
				let n = Ut.getGroupLabel?.(e, t);
				return n === void 0 ? fi(e, Ut.emptyGroupLabel) : n;
			}
		};
	}, [
		Ut,
		R,
		N
	]), z = (0, B.useMemo)(() => {
		if (!Pe) return R;
		let e = R.filter((e) => !L[e.key]);
		return e.length > 0 ? e : R;
	}, [
		R,
		L,
		Pe
	]), Gt = (0, B.useMemo)(() => R.filter(Nn).length, [R]), Kt = (0, B.useMemo)(() => z.filter(Nn).length, [z]), qt = !F && Pe && Gt > 1, Jt = Ue ?? qt, Yt = Qe ? R.filter((e) => e.sortable !== !1) : [], Xt = !!(Ye && Ye.length > 0) || Yt.length > 0, Zt = qt || Jt || Ke || Xt, Qt = ct ? Re : ut, $t = (0, B.useMemo)(() => R.find((e) => e.kind === "timestamp" && e.filterable !== !1 && e.timestamp?.autoRangeFilter !== !1) ?? null, [R]), H = (0, B.useMemo)(() => !f || _ || ne || O?.timeRange ? null : $t, [
		f,
		ne,
		O?.timeRange,
		_,
		$t
	]), en = (0, B.useMemo)(() => R.filter((e) => e.filterable !== !1 && e.kind !== "timestamp"), [R]), tn = (0, B.useMemo)(() => f ? en.map((e) => {
		let t = Fa(zt, e);
		if (e.filterKind || t) return {
			column: e,
			kind: e.filterKind ?? "number",
			options: [],
			...t ? { numberBounds: t } : {}
		};
		let n = /* @__PURE__ */ new Set();
		for (let t of zt) for (let r of An(t.row, e)) r && n.add(r);
		let r = Array.from(n).sort((e, t) => e.localeCompare(t, void 0, { numeric: !0 })).map((e) => ({
			value: e,
			label: e
		})), i = e.kind === "tags" ? 50 : 20, a = r.length >= 2 && r.length <= i;
		if (e.kind === "tags" && a) {
			let t = ga(r, e.tags?.separator ?? "=");
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
		en,
		zt
	]), nn = (0, B.useMemo)(() => {
		let e = {};
		for (let t of tn) {
			if (t.kind !== "multi" && t.kind !== "nested-multi") continue;
			let n = t.column.key;
			e[n] = Yr(xt[n] ?? {}, (e) => St((t) => sa(t, n, e)));
		}
		return e;
	}, [tn, xt]);
	(0, B.useEffect)(() => {
		yt((e) => la(e, tn)), St((e) => ua(e, tn)), wt((e) => da(e, tn));
	}, [tn]);
	let rn = v !== void 0, an = rn ? v : Nt, on = rn ? y ?? (() => {}) : Pt, sn = (0, B.useMemo)(() => tn.map((e) => {
		let t = e.column.key;
		if (e.kind === "multi") return {
			key: t,
			kind: "multi",
			label: Mn(e.column),
			value: xt[t] ?? {},
			onChange: (e) => St((n) => sa(n, t, e)),
			options: e.options
		};
		if (e.kind === "nested-multi") return {
			key: t,
			kind: "nested-multi",
			label: Mn(e.column),
			value: xt[t] ?? {},
			onChange: (e) => St((n) => sa(n, t, e)),
			groups: e.groups ?? []
		};
		if (e.kind === "number" || e.kind === "duration") {
			if (e.kind === "duration") return {
				key: t,
				kind: "duration",
				label: Mn(e.column),
				unit: e.column.filterUnit ?? "ms",
				value: Tt[t] ?? {},
				onChange: (e) => Et(t, e)
			};
			let n = {
				key: t,
				kind: "number",
				label: Mn(e.column),
				value: Tt[t] ?? {},
				onChange: (e) => Et(t, e),
				...e.column.numberUnit ? { unit: e.column.numberUnit } : {}
			};
			return e.numberBounds?.min !== void 0 && (n.domainMin = e.numberBounds.min), e.numberBounds?.max !== void 0 && (n.domainMax = e.numberBounds.max), e.numberBounds?.step !== void 0 && (n.step = e.numberBounds.step), n;
		}
		return {
			key: t,
			kind: "text",
			label: Mn(e.column),
			value: vt[t] ?? "",
			onChange: (e) => yt((n) => oa(n, t, e))
		};
	}), [
		tn,
		xt,
		Tt,
		vt
	]), cn = (0, B.useMemo)(() => new Map((k ?? []).map((e) => [e.key, e])), [k]), dn = (0, B.useMemo)(() => {
		let e = $t?.filterKey, t = e ? cn.get(e) : void 0;
		return t?.kind === "date-range" ? t : null;
	}, [$t, cn]), fn = (0, B.useMemo)(() => {
		let e = /* @__PURE__ */ new Map();
		for (let t of R) {
			let n = sn.find((e) => e.key === t.key);
			if (n) {
				e.set(t.key, n);
				continue;
			}
			let r = t.filterKey ? cn.get(t.filterKey) : void 0;
			r && r !== dn && e.set(t.key, r);
		}
		return e;
	}, [
		R,
		sn,
		cn,
		dn
	]), pn = !!(O?.leading || O?.children || O?.trailing || O?.timeRange || O?.dateRange), mn = (0, B.useMemo)(() => {
		if (!H) return null;
		let e = {
			from: Ft.from,
			to: Ft.to,
			onApply: (e, t) => It({
				from: e,
				to: t
			}),
			presets: H.timestamp?.rangePresets ?? re,
			fromPlaceholder: "now-24h",
			toPlaceholder: "now",
			emptyLabel: "Any time"
		};
		return H.timestamp?.timeEnabled !== void 0 && (e.timeEnabled = H.timestamp.timeEnabled), H.timestamp?.timeZone && (e.timeZone = H.timestamp.timeZone), H.timestamp?.timeZones && (e.timeZones = H.timestamp.timeZones), e;
	}, [
		H,
		Ft.from,
		Ft.to
	]), hn = (0, B.useMemo)(() => dn && $t ? {
		key: $t.key,
		range: lt(dn)
	} : mn && H ? {
		key: H.key,
		range: mn
	} : null, [
		mn,
		$t,
		dn,
		H
	]), gn = ne ?? hn?.range ?? null, _n = (0, B.useMemo)(() => {
		let e = (k ?? []).filter((e) => e !== dn);
		return e.length > 0 ? [...e, ...sn] : sn;
	}, [
		k,
		sn,
		dn
	]), vn = h ?? f;
	Fn({
		autoFilter: f,
		manualFilter: _,
		showGlobalFilter: vn,
		globalFilterControlled: v !== void 0 || y !== void 0,
		hasExternalSearch: te !== void 0
	});
	let yn = vn || sn.length > 0 || !!gn || !!te || k && k.length > 0 || pn || !!P || Zt, bn = (e) => {
		ct || dt(e), Be?.(e);
	}, xn = qe && (fn.size > 0 || !!hn), Sn = (0, B.useMemo)(() => {
		if (_) return zt;
		let e = an.trim().toLowerCase(), t = /* @__PURE__ */ new Date(), n = H ? ee(Ft.from, t) : null, r = H ? ee(Ft.to, t) : null;
		return zt.filter(({ row: t }) => {
			if (H && (n || r)) {
				let e = ce(kn(t, H));
				if (!e || n && e.getTime() < n.getTime() || r && e.getTime() > r.getTime()) return !1;
			}
			if (e && !en.flatMap((e) => An(t, e)).join(" ").toLowerCase().includes(e)) return !1;
			for (let e of tn) {
				let n = An(t, e.column);
				if (e.kind === "text") {
					let t = (vt[e.column.key] ?? "").trim().toLowerCase();
					if (t && !n.some((e) => e.toLowerCase().includes(t))) return !1;
				} else if (e.kind === "number" || e.kind === "duration") {
					let n = Tt[e.column.key] ?? {}, r = pa(n.min, e.kind === "duration" ? n.minUnit : void 0, e.column.filterUnit), i = pa(n.max, e.kind === "duration" ? n.maxUnit : void 0, e.column.filterUnit), a = String(n.min ?? "").trim() !== "", o = String(n.max ?? "").trim() !== "";
					if ((a || o) && !ia(t, e.column).some((e) => (r == null || (n.minOperator === ">" ? e > r : e >= r)) && (i == null || (n.maxOperator === "<" ? e < i : e <= i)))) return !1;
				} else {
					let t = xt[e.column.key] ?? {}, r = Object.entries(t).filter(([, e]) => e === "include").map(([e]) => e), i = Object.entries(t).filter(([, e]) => e === "exclude").map(([e]) => e);
					if (r.length > 0 && !n.some((e) => r.includes(e)) || i.length > 0 && n.some((e) => i.includes(e))) return !1;
				}
			}
			return !0;
		});
	}, [
		an,
		en,
		tn,
		_,
		xt,
		Tt,
		zt,
		vt,
		H,
		Ft.from,
		Ft.to
	]), Cn = (0, B.useMemo)(() => Object.fromEntries(R.map((e) => [e.key, (t) => ra(t.row, e)])), [R]), { renderLink: wn } = Je(), { sorted: Tn, sort: En, toggle: Dn } = Ln(Sn, {
		defaultDir: C?.dir ?? "asc",
		manual: D,
		resolvers: Cn,
		...w === void 0 ? {} : { sort: w },
		...T ? { onSortChange: T } : {},
		...C?.key ? { defaultKey: C.key } : {}
	}), jn = !!j && !T && !(Bi(j, e.length) && !se?.hasMore), Pn = [...Yt.map((e) => {
		let t = En?.key === e.key;
		return {
			id: `sort:${e.key}`,
			label: Mn(e),
			section: "Sort by",
			...t ? {
				icon: En.dir === "asc" ? ln : un,
				iconClassName: "text-primary",
				description: En.dir === "asc" ? "Ascending" : "Descending"
			} : {},
			...jn ? { disabled: !0 } : {},
			onSelect: () => Dn(e.key)
		};
	}), ...Ye ?? []], In = Tn, Bn = zn(se, e.length), Vn = (0, B.useRef)(null), Hn = !!se && se.hasMore, Un = se?.loading ? "Loading more…" : "Scroll to load more…", U = (0, B.useMemo)(() => new Set(M?.selectedRowIds ?? []), [M?.selectedRowIds]), W = (0, B.useMemo)(() => In.filter((e) => M?.isRowSelectable?.(e.row) ?? !0), [M, In]), Wn = M?.selectAll !== !1, Gn = M?.selectAllPages;
	if (!Wn && Gn) throw Error("DataTable rowSelection sets selectAll: false and selectAllPages; selecting across pages needs select-all");
	let Kn = (0, B.useMemo)(() => (Gn?.scopes ?? []).flatMap((e) => {
		let t = e.total ?? j?.total;
		return t == null ? [] : [{
			...e,
			total: t
		}];
	}), [j?.total, Gn?.scopes]), G = Gn?.selectedScopeId ? Kn.find((e) => e.id === Gn.selectedScopeId) : void 0;
	if (Gn?.selectedScopeId && !G) throw Error(`Selected table scope ${Gn.selectedScopeId} is not present in selectAllPages.scopes`);
	let qn = G ? W.length : W.filter((e) => U.has(e.id)).length, Jn = W.length > 0 && qn === W.length, Yn = !G && qn > 0 && !Jn, Xn = (0, B.useRef)(null), Zn = Kn[Kn.length - 1], Qn = Jn && !G ? Kn.find((e) => e.total > U.size) : void 0, $n = !!Zn && (!!G || U.size > 0);
	(0, B.useEffect)(() => {
		Xn.current && (Xn.current.indeterminate = Yn);
	}, [Yn]);
	let er = (0, B.useCallback)((e) => {
		M && M.onSelectionChange(Array.from(e), zt.filter((t) => e.has(t.id)).map((e) => e.row));
	}, [M, zt]), tr = (0, B.useCallback)((e) => {
		if (!M || G || M.isRowSelectable?.(e.row) === !1) return;
		let t = new Set(U);
		t.has(e.id) ? t.delete(e.id) : t.add(e.id), er(t);
	}, [
		er,
		M,
		U,
		G
	]), nr = (0, B.useCallback)(() => {
		if (!M || G) return;
		let e = new Set(U);
		if (Jn) for (let t of W) e.delete(t.id);
		else for (let t of W) e.add(t.id);
		er(e);
	}, [
		Jn,
		er,
		M,
		W,
		U,
		G
	]), rr = (0, B.useMemo)(() => !M || !G && U.size === 0 ? null : {
		selectedRowIds: Array.from(U),
		selectedRows: G ? W.map((e) => e.row) : zt.filter((e) => U.has(e.id)).map((e) => e.row),
		...G ? { selectedScope: {
			id: G.id,
			total: G.total
		} } : {},
		clearSelection: () => er(/* @__PURE__ */ new Set())
	}, [
		er,
		M,
		zt,
		W,
		U,
		G
	]), ir = (0, B.useMemo)(() => !rr || typeof fe != "function" ? null : fe(rr), [fe, rr]), ar = Array.isArray(fe) ? fe : null, or = t == null && pe === "takeover" && !!rr && !!ar, sr = rr && ar ? /* @__PURE__ */ (0, V.jsx)(Ci, {
		actions: ar,
		context: rr,
		showCount: !$n,
		className: "flex flex-wrap items-center gap-density-2"
	}) : ir, cr = t == null ? he === void 0 ? j ? null : r && Tn.length === 0 ? o : `${Tn.length} of ${e.length} row${e.length === 1 ? "" : "s"}` : typeof he == "function" ? he({
		visibleRowCount: Tn.length,
		totalRowCount: e.length,
		loading: r
	}) : he : null, [lr, ur] = (0, B.useState)({}), dr = Ut ? `mode:${Ut.value}` : "fixed", fr = lr[dr], pr = (0, B.useMemo)(() => {
		if (!Wt) return null;
		let e = mi(In, (e) => Wt.getGroupKey(e.row)).map((e) => ({
			key: e.key,
			records: e.rows,
			rows: e.rows.map((e) => e.row)
		}));
		return Wt.compareGroups && e.sort((e, t) => Wt.compareGroups(e, t)), e.map((e) => ({
			...e,
			label: Wt.getGroupLabel?.(e.key, e.rows) ?? e.key,
			meta: Wt.getGroupMeta?.(e.key, e.rows),
			collapsed: fr?.groups[e.key] ?? fr?.all ?? hi(e, Wt.defaultCollapsed)
		}));
	}, [
		fr,
		Wt,
		In
	]), mr = !!pr?.length, hr = mr && pr.every((e) => !e.collapsed), gr = mr && pr.every((e) => e.collapsed), _r = (e) => {
		ur((t) => ({
			...t,
			[dr]: {
				all: e,
				groups: {}
			}
		}));
	}, vr = !or && t == null && ($n || sr) ? /* @__PURE__ */ (0, V.jsxs)("div", {
		"data-testid": "data-table-selection-actions",
		className: "flex flex-wrap items-center gap-density-2",
		children: [$n && Zn ? /* @__PURE__ */ (0, V.jsx)(qi, {
			noun: Gn?.noun ?? "rows",
			...Gn?.loading ? { loading: !0 } : {},
			total: Zn.total,
			selectedCount: G?.total ?? U.size,
			...Qn ? { nextScope: Qn } : {},
			onClear: () => er(/* @__PURE__ */ new Set())
		}) : null, sr]
	}) : null, yr = O?.trailing || P || Zt ? /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
		O?.trailing,
		P ? /* @__PURE__ */ (0, V.jsx)(gi, {
			modes: P,
			value: Vt,
			hasGroups: mr,
			allExpanded: hr,
			allCollapsed: gr,
			onModeChange: (e) => {
				ge === void 0 && Mt(e), ve?.(e);
			},
			onExpandAll: () => _r(!1),
			onCollapseAll: () => _r(!0)
		}) : null,
		Zt ? /* @__PURE__ */ (0, V.jsx)(ea, {
			opensSheet: gt,
			onOpen: (e) => pt(Lr(e))
		}) : null
	] }) : void 0, br = vr || yr ? /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [vr, yr] }) : void 0, xr = or ? /* @__PURE__ */ (0, V.jsxs)("div", {
		"data-testid": "data-table-selection-actions",
		className: "flex min-w-0 flex-1 flex-wrap items-center gap-density-2",
		children: [
			$n && Zn ? /* @__PURE__ */ (0, V.jsx)(qi, {
				noun: Gn?.noun ?? "rows",
				...Gn?.loading ? { loading: !0 } : {},
				total: Zn.total,
				selectedCount: G?.total ?? U.size,
				...Qn ? { nextScope: Qn } : {},
				onClear: () => er(/* @__PURE__ */ new Set())
			}) : null,
			/* @__PURE__ */ (0, V.jsx)(Ci, {
				actions: ar,
				context: rr,
				showCount: !$n,
				className: "flex flex-wrap items-center gap-density-2"
			}),
			yr ? /* @__PURE__ */ (0, V.jsx)("div", {
				className: "ml-auto flex shrink-0 flex-wrap items-center gap-2",
				children: yr
			}) : null
		]
	}) : null, Sr = (0, B.useMemo)(() => {
		let e = (e) => {
			let t = {
				kind: "row",
				record: e
			}, n = Ce ? Ce.includes(e.id) : Dt[e.id];
			if (Oe !== "row" || !Se || !n) return [t];
			let r = Se(e.row, {
				columns: R,
				visibleColumns: z,
				filterActionsByColumn: nn
			});
			return r ? [t, {
				kind: "expansion",
				record: e,
				content: r
			}] : [t];
		};
		return pr ? pr.flatMap((t) => [{
			kind: "group",
			group: t
		}, ...t.collapsed ? [] : t.records.flatMap(e)]) : In.flatMap(e);
	}, [
		pr,
		In,
		Oe,
		Se,
		Ce,
		Dt,
		R,
		z,
		nn
	]), Cr = (0, B.useCallback)((e) => {
		let t = Sr[e];
		return t ? t.kind === "group" ? `group:${t.group.key}` : t.kind === "expansion" ? `expand:${t.record.id}` : `row:${t.record.id}` : `row:${e}`;
	}, [Sr]), wr = (0, B.useMemo)(() => Tr(l), [l]), K = Dr({
		count: Sr.length,
		options: wr,
		getItemKey: Cr
	}), [Er, Or] = (0, B.useState)({}), kr = (0, B.useRef)(null), [Ar, jr] = (0, B.useState)(null), Mr = F ? 1 : (Qe ? 1 : z.length) + +!!M, Nr = (0, B.useMemo)(() => Qe ? Oi(z) : null, [Qe, z]), Pr = z.map((e) => e.key).join(" ");
	(0, B.useEffect)(() => {
		if (!wr || F || Qe) {
			Or({});
			return;
		}
		let e = () => {
			let e = kr.current;
			if (!e) return;
			let t = +!!M, n = {};
			z.forEach((r, i) => {
				let a = e.cells[i + t]?.getBoundingClientRect().width ?? 0;
				a > 0 && (n[r.key] = a);
			}), Object.keys(n).length > 0 && Or(n);
		}, t = requestAnimationFrame(e);
		if (typeof ResizeObserver > "u") return () => cancelAnimationFrame(t);
		let n = new ResizeObserver(e);
		return kr.current?.parentElement && n.observe(kr.current.parentElement), () => {
			cancelAnimationFrame(t), n.disconnect();
		};
	}, [
		wr,
		Pr,
		M,
		Re,
		F,
		Qe
	]), (0, B.useEffect)(() => {
		if (!Ar) return;
		let e = () => Ar.style.setProperty("--dt-viewport-w", `${Ar.clientWidth}px`);
		if (e(), typeof ResizeObserver > "u") return;
		let t = new ResizeObserver(e);
		return t.observe(Ar), () => t.disconnect();
	}, [Ar]);
	let Fr = (0, B.useCallback)((e) => {
		K.setScrollEl(e), jr(e), ue && (ue.current = e);
	}, [K.setScrollEl, ue]), Rr = (0, B.useMemo)(() => K.enabled ? K.virtualItems.flatMap((e) => {
		let t = Sr[e.index];
		return t ? [{
			item: t,
			index: e.index
		}] : [];
	}) : Sr.map((e, t) => ({
		item: e,
		index: t
	})), [
		K.enabled,
		K.virtualItems,
		Sr
	]), zr = (0, B.useCallback)((e) => {
		if (!M || G) return;
		let t = e.filter((e) => M.isRowSelectable?.(e.row) ?? !0), n = new Set(U), r = t.length > 0 && t.every((e) => n.has(e.id));
		for (let e of t) r ? n.delete(e.id) : n.add(e.id);
		er(n);
	}, [
		er,
		M,
		U,
		G
	]), Hr = (0, B.useCallback)((e) => {
		Vn.current && Vn.current.disconnect(), e && Hn && (Vn.current = new IntersectionObserver((e) => {
			e[0]?.isIntersecting && Bn();
		}, { rootMargin: "200px" }), Vn.current.observe(e));
	}, [Bn, Hn]);
	if (F && (Se || M)) throw Error("DataTable card rendering cannot be combined with row detail or row selection");
	let Ur = (e, t) => {
		e.preventDefault(), e.stopPropagation();
		let n = e.clientX, r = (e.currentTarget.closest("th")?.getBoundingClientRect().width ?? 0) || rt[t.key] || Na(t), i = (e) => {
			let i = Pa(t, r + e.clientX - n);
			it((e) => ({
				...e,
				[t.key]: i
			}));
		}, a = () => {
			document.removeEventListener("mousemove", i), document.removeEventListener("mouseup", a), document.body.style.cursor = "", document.body.style.userSelect = "";
		};
		document.addEventListener("mousemove", i), document.addEventListener("mouseup", a), document.body.style.cursor = "col-resize", document.body.style.userSelect = "none";
	}, Wr = (e, t) => {
		e.preventDefault(), e.stopPropagation();
		let n = e.currentTarget.closest("th"), r = e.currentTarget.closest("table");
		if (!n || !r) return;
		let i = Array.from(n.parentElement?.children ?? []).indexOf(n);
		if (i < 0) return;
		let a = ya(r, i, t);
		it((e) => ({
			...e,
			[t.key]: a
		}));
	}, Gr = (e) => {
		Nn(e) && st((t) => {
			let n = t[e.key] === !0;
			if (!n && Kt <= 1) return t;
			let r = { ...t };
			return n ? ot.has(e.key) ? r[e.key] = !1 : delete r[e.key] : r[e.key] = !0, Aa(r, I, ot);
		});
	}, Kr = () => st(Aa(Object.fromEntries([...ot].map((e) => [e, !1])), I, ot)), qr = (e, t) => {
		qt && (e.preventDefault(), pt(Ir(e, t.key)));
	}, Jr = (e, t) => {
		e.preventDefault(), e.stopPropagation(), ht(Lr(e, t));
	}, Zr = (e, t, n, r) => {
		let i = r, a = Ui(n, e, t), o = Wi(n, e, t), s = t.filterKey, c = !!s && !!oe && a !== void 0;
		return s && oe && a !== void 0 && (i = /* @__PURE__ */ (0, V.jsx)(ni, {
			value: a,
			...o === void 0 ? {} : { displayValue: o },
			mode: ae?.[s]?.[a],
			onChange: (e) => oe({
				key: s,
				value: a,
				mode: e
			}),
			children: i
		})), t.kind === "tags" && nn[t.key] && (i = /* @__PURE__ */ (0, V.jsx)(Xr, {
			value: nn[t.key],
			children: i
		})), {
			content: i,
			hasCellFilterActions: c
		};
	}, Qr = (e, t, n) => wn({
		to: e,
		draggable: !1,
		onDragStart: (e) => e.preventDefault(),
		className: "hover:underline after:absolute after:inset-0 after:content-[''] [-webkit-user-drag:none]",
		children: n ? /* @__PURE__ */ (0, V.jsx)("span", {
			className: "relative z-10 inline-flex min-w-0 items-center",
			children: t
		}) : t
	}), $r = (e, t) => {
		let n = On(e, t), r = t.listRender ?? t.render, i = r ? r(n, e) : ha(n);
		return {
			empty: r ? i == null || i === !1 || i === "" : n == null || n === "" || Array.isArray(n) && n.length === 0,
			...Zr(e, t, n, i)
		};
	}, ei = (e, t, n) => {
		let r = $r(e, t.title), i = t.status ? $r(e, t.status).content : null, a = t.meta.flatMap((t) => {
			let n = $r(e, t);
			return n.empty ? [] : [{
				key: t.key,
				label: Mn(t),
				content: n.content
			}];
		});
		return /* @__PURE__ */ (0, V.jsx)(Ei, {
			title: n ? Qr(n, r.content, r.hasCellFilterActions) : r.content,
			status: i,
			meta: a
		});
	};
	return /* @__PURE__ */ (0, V.jsx)(Rn, {
		density: Qt ?? "comfortable",
		children: /* @__PURE__ */ (0, V.jsxs)("div", {
			ref: Xe,
			className: i("flex min-h-0 min-w-0 flex-1 flex-col gap-1", u),
			"data-density": Qt,
			"data-layout": Qe ? "list" : F ? "cards" : "table",
			children: [
				(yn || vr || xr) && /* @__PURE__ */ (0, V.jsx)(Ht, {
					...O,
					...xr ? { overlay: xr } : {},
					...te ? { search: te } : vn ? { search: {
						value: an,
						onChange: on,
						placeholder: S
					} } : {},
					...gn ? { timeRange: gn } : {},
					trailing: br,
					filters: _n
				}),
				De,
				F && !jn && R.some((e) => e.sortable !== !1) && /* @__PURE__ */ (0, V.jsxs)("div", {
					className: "flex shrink-0 items-center gap-2 px-1 text-xs text-muted-foreground",
					children: [/* @__PURE__ */ (0, V.jsxs)("label", {
						className: "flex items-center gap-1.5",
						children: ["Sort by", /* @__PURE__ */ (0, V.jsxs)("select", {
							"aria-label": "Sort cards by",
							className: "h-7 rounded-md border border-input bg-background px-2 text-foreground",
							value: En?.key ?? "",
							onChange: (e) => {
								let t = e.target.value;
								t ? T ? T({
									key: t,
									dir: "asc"
								}) : Dn(t) : T?.(null);
							},
							children: [/* @__PURE__ */ (0, V.jsx)("option", {
								value: "",
								children: "Default order"
							}), R.filter((e) => e.sortable !== !1).map((e) => /* @__PURE__ */ (0, V.jsx)("option", {
								value: e.key,
								children: Mn(e)
							}, e.key))]
						})]
					}), En && /* @__PURE__ */ (0, V.jsx)("button", {
						type: "button",
						className: "h-7 rounded-md border border-input px-2 text-foreground hover:bg-accent",
						onClick: () => {
							T ? T({
								key: En.key,
								dir: En.dir === "asc" ? "desc" : "asc"
							}) : Dn(En.key);
						},
						children: En.dir === "asc" ? "Ascending" : "Descending"
					})]
				}),
				/* @__PURE__ */ (0, V.jsxs)("div", {
					className: "relative flex min-h-0 max-w-full flex-1 flex-col",
					children: [
						r && t == null ? /* @__PURE__ */ (0, V.jsx)(a, {
							"data-testid": "data-table-loading-bar",
							className: "rounded-t-md"
						}) : null,
						/* @__PURE__ */ (0, V.jsx)("div", {
							ref: Fr,
							"data-testid": "data-table-scroll",
							className: i("min-h-0 max-w-full flex-1 overflow-auto overscroll-x-contain rounded-md border border-border bg-background", d),
							"aria-busy": r && t == null || void 0,
							children: /* @__PURE__ */ (0, V.jsxs)("table", {
								className: i(F || Qe ? "w-full text-left text-sm" : "w-max min-w-full text-left text-sm", F || Qe || Object.keys(Er).length > 0 ? "table-fixed" : "table-auto"),
								children: [
									F ? /* @__PURE__ */ (0, V.jsx)("colgroup", { children: /* @__PURE__ */ (0, V.jsx)("col", {}) }) : Qe ? /* @__PURE__ */ (0, V.jsxs)("colgroup", { children: [M ? /* @__PURE__ */ (0, V.jsx)("col", { className: "w-10" }) : null, /* @__PURE__ */ (0, V.jsx)("col", {})] }) : /* @__PURE__ */ (0, V.jsxs)("colgroup", { children: [M && t == null ? /* @__PURE__ */ (0, V.jsx)("col", { className: "w-10" }) : null, z.map((e) => /* @__PURE__ */ (0, V.jsx)("col", {
										style: t == null ? va(e, rt, Er) : void 0,
										className: e.shrink && !e.grow && t == null ? "w-px" : void 0
									}, e.key))] }),
									F ? /* @__PURE__ */ (0, V.jsx)("thead", {
										className: "sr-only",
										children: /* @__PURE__ */ (0, V.jsx)("tr", { children: /* @__PURE__ */ (0, V.jsx)("th", {
											scope: "col",
											children: "Results"
										}) })
									}) : Qe ? /* @__PURE__ */ (0, V.jsx)("thead", {
										className: M && Wn && t == null ? "border-b border-border bg-muted text-xs text-muted-foreground" : "sr-only",
										children: /* @__PURE__ */ (0, V.jsxs)("tr", { children: [M ? /* @__PURE__ */ (0, V.jsx)("th", {
											className: i("w-10 text-center", Ri),
											children: Wn ? /* @__PURE__ */ (0, V.jsx)("input", {
												ref: Xn,
												type: "checkbox",
												"aria-label": "Select all visible rows",
												checked: Jn,
												disabled: !!G || W.length === 0,
												onChange: nr,
												className: "size-3.5 rounded border-border accent-primary"
											}) : null
										}) : null, /* @__PURE__ */ (0, V.jsx)("th", {
											scope: "col",
											className: i("font-medium", Ri),
											children: M && Wn ? "Select all" : "Results"
										})] })
									}) : /* @__PURE__ */ (0, V.jsx)("thead", {
										className: "sticky top-0 z-10 bg-muted shadow-[0_1px_0_0_var(--tw-shadow-color)] shadow-border",
										children: /* @__PURE__ */ (0, V.jsxs)("tr", {
											ref: kr,
											className: "border-b border-border text-xs text-muted-foreground",
											children: [M && t == null ? /* @__PURE__ */ (0, V.jsx)("th", {
												className: i("w-10 text-center", Ri),
												children: Wn ? /* @__PURE__ */ (0, V.jsx)("input", {
													ref: Xn,
													type: "checkbox",
													"aria-label": "Select all visible rows",
													checked: Jn,
													disabled: !!G || W.length === 0,
													onChange: nr,
													className: "size-3.5 rounded border-border accent-primary"
												}) : null
											}) : null, z.map((e) => /* @__PURE__ */ (0, V.jsxs)("th", {
												className: i("group/header relative whitespace-nowrap font-medium", Ri, Ba(e.align), je && e.resizable !== !1 && "select-none", e.headerClassName),
												onContextMenu: (t) => qr(t, e),
												children: [/* @__PURE__ */ (0, V.jsxs)("div", {
													className: i("flex min-w-0 items-center gap-1", Va(e.align), je && e.resizable !== !1 && "pr-2"),
													children: [/* @__PURE__ */ (0, V.jsx)("span", {
														className: "min-w-0",
														children: e.sortable === !1 || jn ? /* @__PURE__ */ (0, V.jsx)("span", {
															...jn && e.sortable !== !1 ? { title: "Sorting is served by the query while this table is paged by the server" } : {},
															children: e.label
														}) : /* @__PURE__ */ (0, V.jsx)(Vr, {
															active: En?.key === e.key,
															...En?.key === e.key ? { dir: En.dir } : {},
															...e.align ? { align: e.align } : {},
															onClick: () => Dn(e.key),
															children: e.label
														})
													}), xn && (fn.has(e.key) || hn?.key === e.key) && /* @__PURE__ */ (0, V.jsx)(Qi, {
														column: e,
														active: fn.has(e.key) ? bt(fn.get(e.key)) : !!(hn?.range.from || hn?.range.to),
														onOpen: (t) => Jr(t, e.key)
													})]
												}), je && e.resizable !== !1 && /* @__PURE__ */ (0, V.jsx)("span", {
													role: "separator",
													"aria-label": `Resize ${Mn(e)} column`,
													"aria-orientation": "vertical",
													className: "absolute right-0 top-0 hidden h-full w-3 cursor-col-resize touch-none items-center justify-center border-r border-border/70 bg-gradient-to-l from-border/30 to-transparent transition-colors hover:border-primary hover:from-primary/20 md:flex",
													onClick: (e) => {
														e.preventDefault(), e.stopPropagation();
													},
													onDoubleClick: (t) => Wr(t, e),
													onMouseDown: (t) => Ur(t, e),
													children: /* @__PURE__ */ (0, V.jsx)("span", {
														"aria-hidden": !0,
														className: "h-4 w-0.5 rounded-full bg-border transition-colors group-hover/header:bg-primary/70"
													})
												})]
											}, e.key))]
										})
									}),
									/* @__PURE__ */ (0, V.jsxs)("tbody", {
										ref: K.setTbodyEl,
										children: [t == null ? r && In.length === 0 ? /* @__PURE__ */ (0, V.jsx)(Xi, {
											columns: Qe ? z.slice(0, 1) : z,
											rowCount: s,
											message: o,
											selection: !!M
										}) : Sn.length === 0 ? /* @__PURE__ */ (0, V.jsx)("tr", { children: /* @__PURE__ */ (0, V.jsx)("td", {
											colSpan: Mr,
											className: "px-4 py-10 text-center text-sm text-muted-foreground",
											children: c
										}) }) : /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
											K.enabled && K.spacers.top > 0 && /* @__PURE__ */ (0, V.jsx)("tr", {
												"aria-hidden": !0,
												"data-virtual-spacer": "top",
												children: /* @__PURE__ */ (0, V.jsx)("td", {
													colSpan: Mr,
													style: {
														height: K.spacers.top,
														padding: 0,
														border: 0
													}
												})
											}),
											Rr.map(({ item: e, index: t }) => {
												let n = K.enabled ? {
													ref: K.virtualizer.measureElement,
													"data-index": t
												} : {};
												if (e.kind === "expansion") return /* @__PURE__ */ (0, V.jsx)("tr", {
													...n,
													children: /* @__PURE__ */ (0, V.jsx)("td", {
														colSpan: Mr,
														className: "bg-muted/40 p-density-3",
														children: /* @__PURE__ */ (0, V.jsx)("div", {
															"data-testid": "data-table-expansion",
															className: "sticky left-density-3 w-[calc(var(--dt-viewport-w)-2*var(--spacing-density-3))] overflow-x-auto rounded-md border border-border bg-background p-density-3",
															children: e.content
														})
													})
												}, `expand:${e.record.id}`);
												if (e.kind === "group") return /* @__PURE__ */ (0, V.jsx)(Yi, {
													...K.enabled ? {
														rowRef: K.virtualizer.measureElement,
														dataIndex: t
													} : {},
													label: e.group.label,
													meta: e.group.meta,
													metaAlign: Wt?.metaAlign ?? "end",
													...Wt?.metaClassName ? { metaClassName: Wt.metaClassName } : {},
													count: e.group.records.length,
													colSpan: Mr,
													collapsed: e.group.collapsed,
													onToggleCollapsed: () => ur((t) => ({
														...t,
														[dr]: {
															...t[dr],
															groups: {
																...t[dr]?.groups,
																[e.group.key]: !e.group.collapsed
															}
														}
													})),
													...M && Wn ? { selection: {
														selectableCount: e.group.records.filter((e) => M.isRowSelectable?.(e.row) ?? !0).length,
														selectedCount: G ? e.group.records.filter((e) => M.isRowSelectable?.(e.row) ?? !0).length : e.group.records.filter((e) => (M.isRowSelectable?.(e.row) ?? !0) && U.has(e.id)).length,
														disabled: !!G,
														onToggle: () => zr(e.group.records)
													} } : {}
												}, `group:${e.group.key}`);
												let r = e.record;
												if (F) return /* @__PURE__ */ (0, V.jsx)("tr", {
													...n,
													className: "align-top",
													children: /* @__PURE__ */ (0, V.jsx)("td", {
														className: "border-b border-border/60 p-density-2",
														children: F(r.row)
													})
												}, r.id);
												let a = xe?.(r.row), o = Ce ? Ce.includes(r.id) : Dt[r.id] ?? !1, s = (Se?.(r.row, {
													columns: R,
													visibleColumns: z,
													filterActionsByColumn: nn
												}) ?? null) !== null, c = s && Oe === "row", l = s && Oe === "dialog", u = be?.(r.row) ?? !!ye, d = M?.isRowSelectable?.(r.row) ?? !0, f = !!M?.toggleOnRowClick && d && !G, p = !!a || u || s || f, m = () => {
													f && tr(r), c && (Ce ? we?.(o ? Ce.filter((e) => e !== r.id) : [...Ce, r.id]) : Ot((e) => ({
														...e,
														[r.id]: !e[r.id]
													}))), l && At(r), u && ye?.(r.row);
												};
												return /* @__PURE__ */ (0, V.jsx)(B.Fragment, { children: /* @__PURE__ */ (0, V.jsxs)("tr", {
													...n,
													"aria-expanded": c ? o : void 0,
													className: i("relative border-b border-border/60 align-top", p && "cursor-pointer hover:bg-accent/40", p && !a && "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring", (G && d || U.has(r.id)) && "bg-accent/50", me?.(r.row)),
													tabIndex: p && !a ? 0 : void 0,
													onClick: m,
													onKeyDown: (e) => {
														a || e.target !== e.currentTarget || (e.key === "Enter" || e.key === " ") && (e.preventDefault(), m());
													},
													children: [M ? /* @__PURE__ */ (0, V.jsx)("td", {
														className: i("relative z-10 w-10 text-center", zi),
														onClick: (e) => e.stopPropagation(),
														children: /* @__PURE__ */ (0, V.jsx)("input", {
															type: "checkbox",
															"aria-label": `Select row ${r.id}`,
															checked: !!G && d || U.has(r.id),
															disabled: !!G || !d,
															onChange: () => tr(r),
															className: "size-3.5 rounded border-border accent-primary"
														})
													}) : null, Nr ? /* @__PURE__ */ (0, V.jsx)("td", {
														className: "min-w-0 border-0 px-density-3 py-density-2",
														children: ei(r.row, Nr, a)
													}) : z.map((e, t) => {
														let n = On(r.row, e), { content: o, hasCellFilterActions: s } = Zr(r.row, e, n, e.render ? e.render(n, r.row) : ha(n));
														return /* @__PURE__ */ (0, V.jsx)("td", {
															className: i(zi, Ba(e.align), e.cellClassName),
															style: Ua(e, rt[e.key]),
															children: /* @__PURE__ */ (0, V.jsx)(na, {
																column: e,
																children: a && t === 0 ? Qr(a, o, s) : o
															})
														}, e.key);
													})]
												}) }, r.id);
											}),
											K.enabled && K.spacers.bottom > 0 && /* @__PURE__ */ (0, V.jsx)("tr", {
												"aria-hidden": !0,
												"data-virtual-spacer": "bottom",
												children: /* @__PURE__ */ (0, V.jsx)("td", {
													colSpan: Mr,
													style: {
														height: K.spacers.bottom,
														padding: 0,
														border: 0
													}
												})
											})
										] }) : /* @__PURE__ */ (0, V.jsx)(Zi, {
											colSpan: Mr,
											error: t
										}), t == null && Hn && /* @__PURE__ */ (0, V.jsx)("tr", {
											ref: Hr,
											"aria-hidden": !0,
											children: /* @__PURE__ */ (0, V.jsx)("td", {
												colSpan: Mr,
												className: "p-density-2 text-center text-xs text-muted-foreground",
												children: Un
											})
										})]
									})
								]
							})
						}),
						!F && !Qe && /* @__PURE__ */ (0, V.jsx)(Ti, { scroll: Ar })
					]
				}),
				cr !== null && /* @__PURE__ */ (0, V.jsx)("div", {
					className: "shrink-0 px-1 text-xs text-muted-foreground",
					children: cr
				}),
				t == null && j ? /* @__PURE__ */ (0, V.jsx)(Ji, {
					pagination: j,
					visibleRowCount: Tn.length,
					infinite: !!se
				}) : null,
				Oe === "dialog" && Se && /* @__PURE__ */ (0, V.jsx)(A, {
					open: kt !== null,
					onClose: () => At(null),
					size: Ae,
					...kt ? { title: typeof ke == "function" ? ke(kt.row) : ke } : {},
					children: kt ? Se(kt.row, {
						columns: R,
						visibleColumns: z,
						filterActionsByColumn: nn
					}) : null
				}),
				ft && Zt && gt && /* @__PURE__ */ (0, V.jsx)(Br, {
					open: !0,
					columns: R,
					hiddenColumns: L,
					activeColumnKey: ft.columnKey,
					actions: Pn,
					showColumnVisibilityControl: qt,
					showDensityControl: Jt,
					showThemeControl: Ke,
					themeMenuValue: We,
					densityOverride: Qt,
					visibleHideableColumnCount: Kt,
					onToggle: Gr,
					onShowAll: Kr,
					onDensityChange: bn,
					...Ge ? { onThemeChange: Ge } : {},
					onClose: () => pt(null)
				}),
				ft && Zt && !gt && /* @__PURE__ */ (0, V.jsx)(ta, {
					columns: R,
					hiddenColumns: L,
					anchor: ft,
					actions: Pn,
					showColumnVisibilityControl: qt,
					showDensityControl: Jt,
					showThemeControl: Ke,
					themeMenuValue: We,
					densityOverride: Qt,
					visibleHideableColumnCount: Kt,
					onToggle: Gr,
					onShowAll: Kr,
					onDensityChange: bn,
					...Ge ? { onThemeChange: Ge } : {},
					onClose: () => pt(null)
				}),
				mt && /* @__PURE__ */ (0, V.jsx)($i, {
					filter: fn.get(mt.columnKey ?? ""),
					...hn && hn.key === mt.columnKey ? { timeRange: hn.range } : {},
					anchor: mt,
					onClose: () => ht(null)
				})
			]
		})
	});
}
function Ui(e, t, n) {
	let r = n.serverFilterValue ? n.serverFilterValue(e, t) : e;
	switch (typeof r) {
		case "string": return r || void 0;
		case "number":
		case "boolean": return String(r);
		default: return;
	}
}
function Wi(e, t, n) {
	let r = n.serverFilterLabel ? n.serverFilterLabel(e, t) : e;
	switch (typeof r) {
		case "string": return r || void 0;
		case "number":
		case "boolean": return String(r);
		default: return;
	}
}
var Gi = [
	5,
	10,
	25,
	50,
	100,
	200
];
function Ki(e) {
	let [t, n] = (0, B.useState)([]), r = e?.current, i = r == null ? 0 : Math.max(t.indexOf(r) + 1, 1);
	return {
		page: i,
		push: (e) => n((t) => [...t.slice(0, i), e]),
		previous: () => i <= 1 ? void 0 : t[i - 2]
	};
}
function qi({ noun: e, loading: t, total: n, selectedCount: r, nextScope: i, onClear: a }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		role: "status",
		"data-testid": "data-table-selection-scope",
		className: "flex flex-wrap items-center gap-density-2 text-xs text-muted-foreground",
		children: [
			/* @__PURE__ */ (0, V.jsx)("span", { children: r >= n ? `All ${n.toLocaleString()} ${e} selected.` : `${r.toLocaleString()} of ${n.toLocaleString()} ${e} selected.` }),
			i ? /* @__PURE__ */ (0, V.jsx)("button", {
				type: "button",
				className: "rounded-md border border-border bg-background px-2 py-0.5 font-medium text-foreground hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50",
				disabled: t,
				onClick: i.onSelectAll,
				children: t ? "Selecting…" : i.label ?? `Select all ${i.total.toLocaleString()} ${e}`
			}) : null,
			/* @__PURE__ */ (0, V.jsx)("button", {
				type: "button",
				className: "rounded-md border border-border bg-background px-2 py-0.5 font-medium text-foreground hover:bg-accent",
				onClick: a,
				children: "Clear selection"
			})
		]
	});
}
function Ji({ pagination: e, visibleRowCount: t, infinite: n = !1 }) {
	let { page: r, pageSize: a, total: o, totalRelation: s, cursor: l, onPageChange: u, onPageSizeChange: d } = e, f = Array.from(/* @__PURE__ */ new Set([...e.pageSizeOptions ?? Gi, a])).sort((e, t) => e - t), p = Ki(l), m = l ? p.page : Math.max(0, r), h = s === "gte", _ = o != null && a > 0 && !h && !l ? Math.max(1, Math.ceil(o / a)) : void 0, v = m === 0, y = l ? l.next == null : e.hasMore == null ? _ == null ? t < a : m >= _ - 1 : !e.hasMore, b = t > 0, x = o === 0 || !b ? 0 : n ? 1 : m * a + 1, C = o != null && b ? Math.min(h ? Infinity : o, x + Math.max(t - 1, 0)) : t, w = o == null ? `${t} row${t === 1 ? "" : "s"}` : `${x}-${C} of ${h ? `~${o}+` : o}`, T = !n && (l != null || u != null), E = () => {
		l ? l.onCursorChange(p.previous()) : u?.(Math.max(0, m - 1));
	}, D = () => {
		l ? l.next != null && (p.push(l.next), l.onCursorChange(l.next)) : u?.(m + 1);
	}, O = le(g("sm"));
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "flex min-h-9 shrink-0 flex-wrap items-center justify-between gap-x-2 gap-y-1 border-t border-border/70 px-1 pt-2 text-xs text-muted-foreground max-sm:flex-nowrap sm:gap-3",
		children: [/* @__PURE__ */ (0, V.jsx)("div", {
			"aria-live": "polite",
			className: "min-w-0 truncate whitespace-nowrap",
			children: w
		}), /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex items-center gap-2 max-sm:flex-nowrap sm:flex-wrap sm:gap-3",
			children: [/* @__PURE__ */ (0, V.jsxs)("label", {
				className: "flex items-center gap-1.5",
				children: [/* @__PURE__ */ (0, V.jsx)("span", {
					className: "max-sm:sr-only",
					children: "Rows per page"
				}), /* @__PURE__ */ (0, V.jsx)("select", {
					className: "h-9 rounded-md border border-border bg-background px-2 text-xs text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:h-7",
					value: a,
					onChange: (e) => d(Number(e.target.value)),
					children: f.map((e) => /* @__PURE__ */ (0, V.jsx)("option", {
						value: e,
						children: e
					}, e))
				})]
			}), T ? /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
				/* @__PURE__ */ (0, V.jsx)("span", {
					className: i("text-center", !O && "min-w-20"),
					title: `Page ${m + 1}${_ == null ? "" : ` of ${_}`}`,
					children: O ? `${m + 1}${_ == null ? "" : `/${_}`}` : `Page ${m + 1}${_ == null ? "" : ` of ${_}`}`
				}),
				/* @__PURE__ */ (0, V.jsx)("button", {
					type: "button",
					className: "inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-border bg-background text-foreground hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50 sm:h-7 sm:w-7",
					"aria-label": "Previous page",
					title: "Previous page",
					disabled: v,
					onClick: E,
					children: /* @__PURE__ */ (0, V.jsx)(c, { icon: S })
				}),
				/* @__PURE__ */ (0, V.jsx)("button", {
					type: "button",
					className: "inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-border bg-background text-foreground hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50 sm:h-7 sm:w-7",
					"aria-label": "Next page",
					title: "Next page",
					disabled: y,
					onClick: D,
					children: /* @__PURE__ */ (0, V.jsx)(c, { icon: Rt })
				})
			] }) : null]
		})]
	});
}
function Yi({ label: e, meta: t, metaAlign: n, metaClassName: r, count: a, colSpan: o, collapsed: s, onToggleCollapsed: l, selection: u, rowRef: d, dataIndex: f }) {
	let p = !!u && u.selectableCount > 0 && u.selectedCount === u.selectableCount, m = !!u && u.selectedCount > 0 && !p, h = (0, B.useRef)(null);
	return (0, B.useEffect)(() => {
		h.current && (h.current.indeterminate = m);
	}, [m]), /* @__PURE__ */ (0, V.jsx)("tr", {
		ref: d,
		"data-index": f,
		className: "border-b border-border/60 bg-muted/40",
		children: /* @__PURE__ */ (0, V.jsx)("td", {
			colSpan: o,
			className: "px-density-2 py-density-1",
			children: /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex items-center gap-2",
				children: [
					u ? /* @__PURE__ */ (0, V.jsx)("input", {
						ref: h,
						type: "checkbox",
						"aria-label": `Select group ${typeof e == "string" ? e : a}`,
						checked: p,
						disabled: u.disabled || u.selectableCount === 0,
						onChange: u.onToggle,
						className: "size-3.5 rounded border-border accent-primary"
					}) : null,
					/* @__PURE__ */ (0, V.jsxs)("button", {
						type: "button",
						"aria-expanded": !s,
						onClick: l,
						className: i("flex min-w-0 items-center gap-1.5 text-left text-sm font-semibold text-foreground hover:text-primary", n === "end" && "flex-1"),
						children: [
							/* @__PURE__ */ (0, V.jsx)(c, {
								icon: s ? j : k,
								className: "shrink-0 text-muted-foreground"
							}),
							/* @__PURE__ */ (0, V.jsx)("span", {
								className: "truncate",
								children: e
							}),
							/* @__PURE__ */ (0, V.jsx)("span", {
								className: "shrink-0 font-normal text-muted-foreground",
								children: a
							})
						]
					}),
					t ? /* @__PURE__ */ (0, V.jsx)("div", {
						className: i("shrink-0 text-sm text-muted-foreground", r),
						children: t
					}) : null
				]
			})
		})
	});
}
function Xi({ columns: e, rowCount: t, message: n, selection: r }) {
	let a = Math.max(1, t), o = Math.max(1, e.length + +!!r);
	return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsx)("tr", {
		className: "border-b border-border/60 bg-muted/20",
		children: /* @__PURE__ */ (0, V.jsxs)("td", {
			colSpan: o,
			className: i(zi, "p-0"),
			children: [/* @__PURE__ */ (0, V.jsx)("div", {
				className: "h-0.5 w-full overflow-hidden bg-muted",
				children: /* @__PURE__ */ (0, V.jsx)("div", { className: "h-full w-1/2 animate-pulse rounded-full bg-primary/70" })
			}), /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground",
				children: [/* @__PURE__ */ (0, V.jsx)("span", {
					"aria-hidden": !0,
					className: "h-3 w-3 animate-spin rounded-full border-2 border-primary/30 border-t-primary"
				}), /* @__PURE__ */ (0, V.jsx)("span", { children: n })]
			})]
		})
	}), Array.from({ length: a }).map((t, n) => /* @__PURE__ */ (0, V.jsxs)("tr", {
		className: "border-b border-border/60",
		children: [r ? /* @__PURE__ */ (0, V.jsx)("td", {
			className: i("w-10", zi),
			children: /* @__PURE__ */ (0, V.jsx)("span", { className: "mx-auto block size-3.5 animate-pulse rounded bg-muted" })
		}) : null, e.map((e, t) => /* @__PURE__ */ (0, V.jsx)("td", {
			className: i(zi, Ba(e.align), e.cellClassName),
			style: Ua(e, void 0),
			children: /* @__PURE__ */ (0, V.jsx)(na, {
				column: e,
				children: /* @__PURE__ */ (0, V.jsx)("span", {
					"aria-hidden": !0,
					className: i("block h-3 animate-pulse rounded-full bg-muted", Ha(t))
				})
			})
		}, e.key))]
	}, n))] });
}
function Zi({ colSpan: e, error: t }) {
	return /* @__PURE__ */ (0, V.jsx)("tr", {
		className: "border-b border-border/60",
		children: /* @__PURE__ */ (0, V.jsx)("td", {
			colSpan: Math.max(1, e),
			className: i(zi, "p-density-3"),
			children: /* @__PURE__ */ (0, V.jsx)("div", {
				role: "alert",
				className: "w-0 min-w-full overflow-hidden",
				children: t
			})
		})
	});
}
function Qi({ column: e, active: t, onOpen: n }) {
	return /* @__PURE__ */ (0, V.jsx)("button", {
		type: "button",
		"aria-label": `Open ${Mn(e)} column filter`,
		"aria-haspopup": "dialog",
		"aria-pressed": t,
		className: i("inline-flex h-5 w-5 shrink-0 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", t && "bg-accent text-foreground"),
		onClick: n,
		children: /* @__PURE__ */ (0, V.jsx)("span", {
			"data-filter-icon-state": t ? "filled" : "outline",
			className: "inline-flex",
			children: /* @__PURE__ */ (0, V.jsx)(c, {
				icon: t ? At : Nt,
				className: "text-xs"
			})
		})
	});
}
function $i({ filter: e, timeRange: t, anchor: n, onClose: r }) {
	let i = zr(n);
	if (!e && !t) return null;
	let a = e?.label ?? "Time range", o = !!(t?.from || t?.to);
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		ref: i.ref,
		role: "dialog",
		"aria-label": `${a} column filter`,
		className: "z-50 max-w-[calc(100vw-1rem)] overflow-auto rounded-md border border-border bg-popover p-2 text-popover-foreground shadow-lg shadow-black/5",
		style: i.style,
		onClick: (e) => e.stopPropagation(),
		onContextMenu: (e) => e.preventDefault(),
		children: [/* @__PURE__ */ (0, V.jsxs)("div", {
			className: "mb-2 flex items-center justify-between gap-3",
			children: [/* @__PURE__ */ (0, V.jsx)("div", {
				className: "truncate text-xs font-semibold uppercase tracking-wide text-muted-foreground",
				children: a
			}), /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex items-center gap-1",
				children: [
					e && /* @__PURE__ */ (0, V.jsx)("button", {
						type: "button",
						className: "rounded px-1.5 py-0.5 text-xs text-primary transition-colors hover:bg-accent focus:bg-accent focus:outline-none disabled:text-muted-foreground",
						onClick: () => Pt(e),
						disabled: !bt(e),
						children: "Clear all"
					}),
					t && /* @__PURE__ */ (0, V.jsx)("button", {
						type: "button",
						className: "rounded px-1.5 py-0.5 text-xs text-primary transition-colors hover:bg-accent focus:bg-accent focus:outline-none disabled:text-muted-foreground",
						onClick: () => t.onApply("", ""),
						disabled: !o,
						children: "Clear all"
					}),
					/* @__PURE__ */ (0, V.jsx)("button", {
						type: "button",
						"aria-label": "Close column filter",
						title: "Close",
						className: "inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none",
						onClick: r,
						children: /* @__PURE__ */ (0, V.jsx)(c, {
							icon: d,
							className: "text-sm"
						})
					})
				]
			})]
		}), /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex flex-col gap-2",
			children: [e && /* @__PURE__ */ (0, V.jsx)(at, {
				filter: e,
				chrome: "embedded"
			}), t && /* @__PURE__ */ (0, V.jsx)(vt, {
				kind: "time",
				label: "Time range",
				...t
			})]
		})]
	});
}
function ea({ opensSheet: e, onOpen: t }) {
	return /* @__PURE__ */ (0, V.jsx)("button", {
		type: "button",
		"aria-label": "Open column menu",
		"aria-haspopup": e ? "dialog" : "menu",
		className: "inline-flex h-8 w-8 max-sm:h-9 max-sm:w-9 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
		onClick: (e) => {
			e.stopPropagation(), t(e);
		},
		children: /* @__PURE__ */ (0, V.jsx)(c, {
			icon: he,
			className: "text-sm"
		})
	});
}
function ta({ columns: e, hiddenColumns: t, anchor: n, actions: r, showColumnVisibilityControl: a, showDensityControl: o, showThemeControl: s, themeMenuValue: l, densityOverride: u, visibleHideableColumnCount: d, onToggle: f, onShowAll: p, onDensityChange: m, onThemeChange: h, onClose: g }) {
	let _ = n.columnKey ? e.find((e) => e.key === n.columnKey) : void 0, v = _ && Nn(_) && d > 1, y = r.length > 0, b = zr(n);
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		ref: b.ref,
		role: "menu",
		"aria-label": "Column menu",
		className: "z-50 min-w-[16rem] max-w-[calc(100vw-1rem)] overflow-auto rounded-md border border-border bg-popover p-1.5 text-popover-foreground shadow-lg shadow-black/5",
		style: b.style,
		onClick: (e) => e.stopPropagation(),
		onContextMenu: (e) => e.preventDefault(),
		children: [
			a && /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
				/* @__PURE__ */ (0, V.jsxs)("div", {
					className: "flex items-center justify-between gap-3 px-2 py-1.5 text-xs font-medium text-muted-foreground",
					children: [/* @__PURE__ */ (0, V.jsx)("span", { children: "Columns" }), /* @__PURE__ */ (0, V.jsx)("button", {
						type: "button",
						className: "rounded px-1.5 py-0.5 text-xs text-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none",
						onClick: p,
						children: "Show all"
					})]
				}),
				_ && /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsxs)("button", {
					type: "button",
					role: "menuitem",
					disabled: !v,
					className: i("flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", !v && "cursor-not-allowed opacity-50"),
					onClick: () => {
						v && (f(_), g());
					},
					children: [/* @__PURE__ */ (0, V.jsx)(c, {
						icon: Oe,
						className: "text-sm text-muted-foreground"
					}), /* @__PURE__ */ (0, V.jsxs)("span", { children: ["Hide ", Mn(_)] })]
				}), /* @__PURE__ */ (0, V.jsx)("div", { className: "my-1 h-px bg-border" })] }),
				/* @__PURE__ */ (0, V.jsx)("div", {
					className: "max-h-72 overflow-auto",
					children: e.map((e) => {
						let n = Nn(e), r = t[e.key] !== !0, a = !n || r && d <= 1;
						return /* @__PURE__ */ (0, V.jsxs)("label", {
							className: i("flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors hover:bg-accent hover:text-accent-foreground", a && "cursor-not-allowed opacity-50"),
							children: [/* @__PURE__ */ (0, V.jsx)("input", {
								type: "checkbox",
								className: "h-4 w-4 rounded border-border",
								checked: r,
								disabled: a,
								onChange: () => f(e)
							}), /* @__PURE__ */ (0, V.jsx)("span", {
								className: "truncate",
								children: Mn(e)
							})]
						}, e.key);
					})
				})
			] }),
			o && /* @__PURE__ */ (0, V.jsx)(Mr, {
				densityOverride: u,
				separated: a,
				onDensityChange: m
			}),
			y && /* @__PURE__ */ (0, V.jsx)(Ar, {
				actions: r,
				separated: a || o,
				onClose: g
			}),
			s && h && /* @__PURE__ */ (0, V.jsx)(Fr, {
				value: l ?? "system",
				separated: a || o || y,
				onChange: h
			})
		]
	});
}
function na({ column: e, children: t }) {
	return /* @__PURE__ */ (0, V.jsx)("div", {
		className: i(Wa(e), Ba(e.align)),
		children: t
	});
}
function ra(e, t) {
	let n = On(e, t);
	return t.sortValue ? t.sortValue(n, e) : n;
}
function ia(e, t) {
	return aa(kn(e, t));
}
function aa(e) {
	if (Array.isArray(e)) return e.flatMap((e) => aa(e));
	let t = Ra(e);
	return t == null ? [] : [t];
}
function oa(e, t, n) {
	let r = { ...e };
	return n.trim() ? r[t] = n : delete r[t], r;
}
function sa(e, t, n) {
	let r = { ...e };
	return Object.keys(n).length > 0 ? r[t] = n : delete r[t], r;
}
function ca(e, t, n) {
	let r = { ...e }, i = String(n.min ?? "").trim() !== "", a = String(n.max ?? "").trim() !== "";
	return i || a ? r[t] = n : delete r[t], r;
}
function la(e, t) {
	return ma(e, t, "text", (e) => !String(e ?? "").trim());
}
function ua(e, t) {
	let n = new Set(t.filter((e) => e.kind === "multi" || e.kind === "nested-multi").map((e) => e.column.key));
	return Object.fromEntries(Object.entries(e).filter(([e, t]) => n.has(e) && Object.keys(t).length > 0));
}
function da(e, t) {
	let n = new Set(t.filter((e) => e.kind === "number" || e.kind === "duration").map((e) => e.column.key));
	return Object.fromEntries(Object.entries(e).filter(([e, t]) => n.has(e) && (String(t.min ?? "").trim() || String(t.max ?? "").trim())));
}
var fa = {
	ms: 1,
	s: 1e3,
	m: 6e4,
	h: 36e5
};
function pa(e, t, n) {
	let r = Ra(e);
	return r == null ? null : r * fa[t ?? n ?? "ms"] / fa[n ?? "ms"];
}
function ma(e, t, n, r) {
	let i = new Set(t.filter((e) => e.kind === n).map((e) => e.column.key));
	return Object.fromEntries(Object.entries(e).filter(([e, t]) => i.has(e) && !r(t)));
}
function ha(e) {
	return e == null || e === "" ? /* @__PURE__ */ (0, V.jsx)("span", {
		className: "text-muted-foreground",
		children: "—"
	}) : typeof e == "boolean" ? e ? "True" : "False" : typeof e == "object" ? JSON.stringify(e) : String(e);
}
function ga(e, t) {
	let n = /* @__PURE__ */ new Map();
	for (let r of e) {
		let { key: e } = Kr(r.value, t), i = n.get(e) ?? [];
		i.push(r), n.set(e, i);
	}
	return Array.from(n.entries()).sort(([e], [t]) => e === "" && t !== "" ? 1 : t === "" && e !== "" ? -1 : e.localeCompare(t)).map(([e, t]) => ({
		groupKey: e,
		label: e === "" ? "Other" : e,
		options: t
	}));
}
function _a(e, t) {
	if (!e.kind) return e;
	if (e.kind === "timestamp") {
		let n = t ?? "iso";
		return {
			...e,
			render: e.render ?? ((t) => /* @__PURE__ */ (0, V.jsx)(Hr, {
				value: t,
				format: n,
				showTitleOnHover: e.timestamp?.alwaysShowFullOnHover !== !1
			})),
			sortValue: e.sortValue ?? ((e) => ce(e)?.getTime() ?? null),
			filterValue: e.filterValue ?? ((e) => {
				let t = ce(e);
				return t ? t.toISOString() : "";
			})
		};
	}
	if (e.kind === "tags") {
		let t = e.tags, n = t?.separator ?? "=";
		return {
			...e,
			render: e.render ?? ((e) => /* @__PURE__ */ (0, V.jsx)(ei, {
				tags: Ur(e, n),
				maxVisible: t?.maxVisible ?? 3
			})),
			filterValue: e.filterValue ?? ((e) => Gr(e, n)),
			sortValue: e.sortValue ?? ((e) => Gr(e, n).length)
		};
	}
	if (e.kind === "status") {
		let t = e.status, n = t?.map ?? ((e) => ui(e));
		return {
			...e,
			render: e.render ?? ((e, r) => {
				let i = n(e, r);
				if (!i) return /* @__PURE__ */ (0, V.jsx)("span", {
					className: "text-muted-foreground",
					children: "—"
				});
				let a = t?.showLabel ? typeof e == "string" ? e : i : void 0, o = t?.title?.(e, r) ?? (typeof e == "string" ? e : void 0);
				return /* @__PURE__ */ (0, V.jsx)(oi, {
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
function va(e, t, n = {}) {
	let r = t[e.key] ?? n[e.key];
	return r ? { width: `${r}px` } : void 0;
}
function ya(e, t, n) {
	return Pa(n, Array.from(e.querySelectorAll("th, td")).filter((e) => e instanceof HTMLTableCellElement && e.cellIndex === t && e.colSpan === 1).reduce((e, t) => {
		let n = ba(t), r = Array.from(t.children).map((e) => e instanceof HTMLElement ? e.scrollWidth + n : 0);
		return Math.max(e, t.scrollWidth, ...r);
	}, 0) || Na(n));
}
function ba(e) {
	if (typeof window > "u") return 0;
	let t = window.getComputedStyle(e);
	return xa(t.paddingLeft) + xa(t.paddingRight);
}
function xa(e) {
	let t = Number.parseFloat(e);
	return Number.isFinite(t) ? t : 0;
}
function Sa(e, t) {
	if (typeof window > "u") return {};
	try {
		let n = window.localStorage.getItem(e);
		if (!n) return {};
		let r = JSON.parse(n);
		return !r || typeof r != "object" || Array.isArray(r) ? {} : ka(r, t);
	} catch {
		return {};
	}
}
function Ca(e, t) {
	if (typeof window < "u") try {
		Object.keys(t).length === 0 ? window.localStorage.removeItem(e) : window.localStorage.setItem(e, JSON.stringify(t));
	} catch {}
}
function wa(e, t, n) {
	let r = Aa({}, t, n);
	if (typeof window > "u") return r;
	try {
		let i = window.localStorage.getItem(e);
		if (!i) return r;
		let a = JSON.parse(i);
		return !a || typeof a != "object" || Array.isArray(a) ? r : Aa(a, t, n);
	} catch {
		return r;
	}
}
function Ta(e, t, n) {
	if (typeof window > "u") return;
	let r = Object.fromEntries(Object.entries(t).filter(([e, t]) => t !== n.has(e)));
	try {
		Object.keys(r).length === 0 ? window.localStorage.removeItem(e) : window.localStorage.setItem(e, JSON.stringify(r));
	} catch {}
}
function Ea(e) {
	if (typeof window < "u") try {
		let t = window.localStorage.getItem(e);
		return Oa(t) ? t : void 0;
	} catch {
		return;
	}
}
function Da(e, t) {
	if (typeof window < "u") try {
		t ? window.localStorage.setItem(e, t) : window.localStorage.removeItem(e);
	} catch {}
}
function Oa(e) {
	return e === "compact" || e === "comfortable" || e === "spacious";
}
function ka(e, t) {
	let n = new Map(t.map((e) => [e.key, e])), r = {};
	for (let [t, i] of Object.entries(e)) {
		let e = n.get(t);
		e && typeof i == "number" && Number.isFinite(i) && (r[t] = Pa(e, i));
	}
	return r;
}
function Aa(e, t, n) {
	let r = t.filter(Nn), i = new Set(r.map((e) => e.key)), a = {};
	for (let e of n) i.has(e) && (a[e] = !0);
	for (let [t, r] of Object.entries(e)) i.has(t) && (r === !0 ? a[t] = !0 : r === !1 && n.has(t) && (a[t] = !1));
	let o = r.filter((e) => a[e.key] !== !0).length;
	if (r.length > 0 && o === 0) {
		let e = r[0].key;
		n.has(e) ? a[e] = !1 : delete a[e];
	}
	return a;
}
function ja(e, t) {
	let n = Object.keys(e), r = Object.keys(t);
	return n.length === r.length && n.every((n) => Object.prototype.hasOwnProperty.call(t, n) && e[n] === t[n]);
}
function Ma(e, t) {
	let n = Object.keys(e), r = Object.keys(t);
	return n.length === r.length && n.every((n) => Object.prototype.hasOwnProperty.call(t, n) && e[n] === t[n]);
}
function Na(e) {
	return e.grow ? ji : e.shrink ? Mi : Ai;
}
function Pa(e, t) {
	let n = e.minWidth ?? ki, r = e.maxWidth, i = Number.isFinite(n) ? Math.max(1, n) : ki;
	return Math.round(Math.max(i, Math.min(r != null && Number.isFinite(r) ? Math.max(i, r) : Infinity, t)));
}
function Fa(e, t) {
	let n = e.flatMap((e) => Ia(kn(e.row, t))).filter(La), r = n.map((e) => Ra(e)).filter((e) => e != null);
	return n.length === 0 || r.length !== n.length ? null : {
		min: Math.min(...r),
		max: Math.max(...r),
		step: za(r)
	};
}
function Ia(e) {
	return Array.isArray(e) ? e.flatMap((e) => Ia(e)) : [e];
}
function La(e) {
	return e != null && (typeof e != "string" || e.trim() !== "");
}
function Ra(e) {
	if (typeof e == "number") return Number.isFinite(e) ? e : null;
	if (typeof e != "string") return null;
	let t = e.trim();
	if (!t) return null;
	let n = Number(t);
	return Number.isFinite(n) ? n : null;
}
function za(e) {
	if (e.every(Number.isInteger)) return 1;
	let t = e.reduce((e, t) => {
		let [, n = ""] = String(t).split(".");
		return Math.max(e, n.length);
	}, 0);
	return 10 ** -Math.min(t || 2, 6);
}
function Ba(e) {
	return e === "right" ? "text-right" : e === "center" ? "text-center" : "text-left";
}
function Va(e) {
	return e === "right" ? "justify-end" : e === "center" ? "justify-center" : "justify-start";
}
function Ha(e) {
	return [
		"w-24",
		"w-40",
		"w-28",
		"w-56",
		"w-32",
		"w-48"
	][e % 6];
}
function Ua(e, t) {
	return t === void 0 ? e.grow ? {
		maxWidth: "0px",
		minWidth: e.minWidth ?? ji
	} : { maxWidth: e.maxWidth ?? (e.shrink ? Pi : Ni) } : { maxWidth: `${t}px` };
}
function Wa(e) {
	return i("w-full min-w-0 truncate", e.shrink && "whitespace-nowrap");
}
function Ga({ theme: e = "system", onThemeChange: t, showThemeControl: n = !1, showFullscreenControl: r = !1, fullscreenTitle: a = "Table", fullscreenButtonLabel: o = "Open table full screen", className: s, filterBarProps: c, scrollContainerRef: l, ...u }) {
	let d = t !== void 0, [f, p] = (0, B.useState)(e), m = d ? e : f, h = ue(m), g = (e) => {
		d || p(e), t?.(e);
	}, [_, v] = (0, B.useState)(!1), y = ({ inFullscreen: e }) => {
		let t = r && !e, i = c?.trailing, a = t ? {
			...c,
			trailing: /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [i, /* @__PURE__ */ (0, V.jsx)(Ka, {
				label: o,
				onClick: () => v(!0)
			})] })
		} : c, s = e === _;
		return /* @__PURE__ */ (0, V.jsx)(Hi, {
			...u,
			...s && l ? { scrollContainerRef: l } : {},
			...a ? { filterBarProps: a } : {},
			themeMenuValue: m,
			showThemeControl: n,
			onThemeChange: g
		});
	};
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		"data-theme": h,
		className: i("flex min-h-0 flex-1 flex-col text-foreground", s),
		children: [y({ inFullscreen: !1 }), r && /* @__PURE__ */ (0, V.jsx)(A, {
			open: _,
			onClose: () => v(!1),
			title: a,
			size: "full",
			className: "h-[95vh]",
			children: /* @__PURE__ */ (0, V.jsx)("div", {
				"data-theme": h,
				className: "flex h-full min-h-0 flex-col bg-background text-foreground",
				children: y({ inFullscreen: !0 })
			})
		})]
	});
}
function Ka({ label: e, onClick: t }) {
	return /* @__PURE__ */ (0, V.jsx)("button", {
		type: "button",
		"aria-label": e,
		title: e,
		className: "inline-flex h-8 w-8 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
		onClick: t,
		children: /* @__PURE__ */ (0, V.jsx)(c, {
			icon: te,
			className: "text-sm"
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/log-sections.ts
var qa = /* @__PURE__ */ new Set([
	"group",
	"endgroup",
	"notice",
	"warning",
	"error",
	"debug",
	"progress",
	"description",
	"status",
	"stop-commands"
]), Ja = {
	error: "Error",
	warning: "Warning",
	notice: "Notice"
}, Ya = {
	"%25": "%",
	"%0D": "\r",
	"%0A": "\n"
}, Xa = /^##\[(group|endgroup)\](.*)$/, Za = /^\s*(\d+)\s*$/;
function Qa(e) {
	let t = Xa.exec(e);
	if (t) return {
		name: t[1],
		properties: {},
		message: t[2]
	};
	if (!e.startsWith("::")) return;
	let n = e.indexOf("::", 2);
	if (n < 0) return;
	let r = e.slice(2, n), i = r.indexOf(" "), a = i < 0 ? r : r.slice(0, i);
	if (!qa.has(a)) return;
	let o = {}, s = i < 0 ? "" : r.slice(i + 1).trim();
	for (let e of s ? s.split(",") : []) {
		let [t, ...n] = e.split("=");
		if (!t?.trim() || n.length === 0) return;
		o[t.trim()] = n.join("=");
	}
	return {
		name: a,
		properties: o,
		message: e.slice(n + 2).replace(/%25|%0D|%0A/g, (e) => Ya[e])
	};
}
function $a({ properties: e, message: t }) {
	let n = t.indexOf("/"), r = "done" in e || "total" in e, i = Za.exec(r ? e.done ?? "" : t.slice(0, Math.max(n, 0))), a = Za.exec(r ? e.total ?? "" : t.slice(n + 1));
	return i && a ? {
		done: Number(i[1]),
		total: Number(a[1])
	} : void 0;
}
function eo(e) {
	let t = [], n = {
		sections: t,
		hasCommands: !1
	}, r = [], i = [], a = "", o, s = () => {
		i.length > 0 && (r[r.length - 1]?.children ?? t).push({
			kind: "text",
			text: i.join("\n")
		}), i = [];
	}, c = (e) => {
		let t = e.split("\n").length;
		for (let e of r) e.lines += t;
		i.push(e), o = void 0;
	}, l = (e, i) => {
		let { name: l, message: u } = e;
		if (l === "group") {
			s();
			let e = {
				kind: "group",
				name: u.trim(),
				lines: 0,
				closed: !1,
				children: []
			};
			(r[r.length - 1]?.children ?? t).push(e), r.push(e), o = void 0;
		} else if (l === "endgroup") {
			s();
			let e = r.pop();
			e && (e.closed = !0), o = void 0;
		} else if (l === "status") n.status = u, u !== o && c(u), o = u;
		else if (l === "description") n.description = u;
		else if (l === "progress") {
			let t = $a(e);
			t ? n.progress = t : c(i);
		} else l === "stop-commands" ? (a = u, u || c(i)) : l !== "debug" && c(`${Ja[l]}: ${u}`);
	};
	for (let t of e.split("\n")) {
		let e = t.replace(/\r$/, "");
		if (a) {
			e === `::${a}::` ? a = "" : c(t);
			continue;
		}
		let r = Qa(e);
		r ? (n.hasCommands = !0, l(r, t)) : c(t);
	}
	return s(), n;
}
function to(e) {
	return e.kind === "text" ? e.text.split("\n").length : e.closed ? 1 : 1 + no(e.children);
}
function no(e) {
	return e.reduce((e, t) => e + to(t), 0);
}
function ro(e, t, n) {
	let r = [], i = t;
	for (let t of n === "head" ? e : [...e].reverse()) {
		if (i <= 0) break;
		let e = to(t);
		if (e <= i) r.push(t);
		else if (t.kind === "text") {
			let e = t.text.split("\n");
			r.push({
				kind: "text",
				text: (n === "head" ? e.slice(0, i) : e.slice(-i)).join("\n")
			});
		} else r.push({
			...t,
			children: ro(t.children, i - 1, n)
		});
		i -= e;
	}
	return n === "head" ? r : r.reverse();
}
// ../../../clicky-ui/packages/ui/src/lib/color.ts
var io = /^(?:bg|text|border|fill|stroke)-/;
function ao(e) {
	return io.test(e.trim());
}
// ../../../clicky-ui/packages/ui/src/data/ProgressBar.tsx
function oo({ segments: e, total: t, label: n, formatValue: r = String, height: a = "h-2", className: o }) {
	if (t === 0) return null;
	let s = e.filter((e) => e.count > 0).map((e) => `${r(e.count)} ${e.label}`).join(", ");
	return /* @__PURE__ */ (0, V.jsx)("div", {
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
			let o = e.count / t * 100, s = ao(e.color);
			return /* @__PURE__ */ (0, V.jsx)("div", {
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
// ../../../clicky-ui/packages/ui/src/data/LogViewer.tsx
function so(e) {
	return e.map((e, t) => e.kind === "text" ? /* @__PURE__ */ (0, V.jsx)(H, {
		as: "span",
		text: e.text
	}, t) : /* @__PURE__ */ (0, V.jsxs)("details", {
		open: !e.closed,
		className: "block",
		"data-log-group": e.name,
		children: [/* @__PURE__ */ (0, V.jsxs)("summary", {
			className: "cursor-pointer select-none opacity-80 hover:opacity-100",
			children: [e.name || "group", /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "ml-1 opacity-60",
				children: [
					"(",
					e.lines,
					" ",
					e.lines === 1 ? "line" : "lines",
					")"
				]
			})]
		}), /* @__PURE__ */ (0, V.jsx)("div", {
			className: "border-l border-current/20 pl-2",
			children: so(e.children)
		})]
	}, t));
}
var co = {
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
function lo({ log: e, className: t }) {
	let { description: n, status: r, progress: a } = e;
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i("flex items-center gap-2 rounded-b border border-t-0 px-1.5 py-1 text-[11px]", t),
		"data-log-status": "",
		children: [
			n && /* @__PURE__ */ (0, V.jsx)("span", {
				className: "shrink-0 font-medium",
				children: n
			}),
			r && /* @__PURE__ */ (0, V.jsx)("span", {
				className: "min-w-0 flex-1 truncate",
				title: r,
				children: r
			}),
			a && /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "ml-auto flex shrink-0 items-center gap-1.5 tabular-nums",
				children: [
					/* @__PURE__ */ (0, V.jsx)(oo, {
						label: "Progress",
						segments: [{
							count: a.done,
							color: "bg-blue-500",
							label: "done"
						}],
						total: a.total,
						height: "h-1.5",
						className: "w-24"
					}),
					a.done,
					"/",
					a.total
				]
			})
		]
	});
}
function uo({ logs: e, collapsedLines: t = 5, tailLines: n, maxHeight: r, maxExpandedVh: a = 70, variant: o = "muted", bgClass: s, borderClass: c, className: l }) {
	let [u, d] = (0, B.useState)(!1), [f, p] = (0, B.useState)(!1), m = (0, B.useMemo)(() => eo(e), [e]), h = (0, B.useMemo)(() => f ? [{
		kind: "text",
		text: e
	}] : m.sections, [
		f,
		e,
		m
	]), g = (0, B.useMemo)(() => n === void 0 ? u || r ? h : ro(h, t, "head") : ro(h, n, "tail"), [
		h,
		n,
		u,
		r,
		t
	]), _ = (0, B.useMemo)(() => no(h), [h]), v = n === void 0 && _ > t, y = co[o], b = !f && !!(m.description || m.status || m.progress), x = r ?? `${t * 1.4}em`, S = s ?? y.bg, C = c ?? y.border;
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i("relative", l),
		children: [
			/* @__PURE__ */ (0, V.jsx)("pre", {
				className: i("mt-0.5 text-[11px] rounded p-1.5 whitespace-pre-wrap overflow-y-auto border transition-all duration-200", b && "rounded-b-none", y.text, S, C),
				style: { maxHeight: u ? `${a}vh` : x },
				"data-log-variant": o,
				children: so(g)
			}),
			m.hasCommands && /* @__PURE__ */ (0, V.jsx)("button", {
				type: "button",
				"aria-pressed": f,
				title: "Show the log exactly as it was printed",
				className: i("absolute right-1 top-1.5 rounded border px-1 text-[10px] hover:opacity-100", f ? "opacity-100" : "opacity-60", y.text, S, C),
				onClick: () => p((e) => !e),
				children: "Raw"
			}),
			b && /* @__PURE__ */ (0, V.jsx)(lo, {
				log: m,
				className: i(y.text, S, C)
			}),
			v && /* @__PURE__ */ (0, V.jsx)("button", {
				type: "button",
				className: i("text-[10px] mt-0.5", u ? "text-muted-foreground" : "text-primary hover:opacity-80"),
				onClick: () => d((e) => !e),
				children: u ? `▲ Collapse (${_} lines)` : `▼ Show more (${_} lines)`
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/rpc/useOperations.ts
var fo = Pe();
function po(e) {
	return (0, fo.useQuery)({
		queryKey: ["openapi-spec"],
		queryFn: () => e.getOpenAPISpec(),
		staleTime: 3e5
	});
}
function mo(e) {
	let { data: t, ...n } = po(e);
	return {
		operations: (0, B.useMemo)(() => {
			if (!t) return [];
			let e = [];
			for (let [n, r] of Object.entries(t.paths)) for (let [t, i] of Object.entries(r)) e.push({
				path: n,
				method: t,
				operation: i
			});
			return e;
		}, [t]),
		spec: t,
		...n
	};
}
// ../../../clicky-ui/packages/ui/src/rpc/FilterFormFields.tsx
function ho({ filters: e, timeRange: t, autoSubmit: n, isSubmitting: r, submitLabel: a, submittingLabel: s, className: c, showSubmit: l = !0 }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i("space-y-3", c),
		children: [/* @__PURE__ */ (0, V.jsxs)("div", {
			className: "divide-y divide-border border-y border-border",
			children: [e.map((e) => /* @__PURE__ */ (0, V.jsx)(go, {
				filter: e,
				autoSubmit: n
			}, e.key)), t && /* @__PURE__ */ (0, V.jsx)(yo, { timeRange: t })]
		}), !n && l && /* @__PURE__ */ (0, V.jsx)("div", {
			className: "flex justify-end",
			children: /* @__PURE__ */ (0, V.jsx)(o, {
				type: "submit",
				disabled: r,
				children: r ? s : a
			})
		})]
	});
}
function go({ filter: e, autoSubmit: t }) {
	let n = `clicky-param-${e.key.replace(/[^a-zA-Z0-9_-]/g, "-")}`;
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		title: e.description,
		className: i("grid grid-cols-1 gap-2 py-2 sm:grid-cols-[minmax(8rem,14rem)_minmax(0,1fr)] sm:items-center", e.disabled && "opacity-60"),
		children: [/* @__PURE__ */ (0, V.jsx)("label", {
			htmlFor: n,
			className: "text-sm font-medium text-muted-foreground",
			children: e.label
		}), /* @__PURE__ */ (0, V.jsx)("div", {
			className: "min-w-0",
			children: _o(e, n, t)
		})]
	});
}
function _o(e, t, n) {
	if (e.kind === "date-range") return /* @__PURE__ */ (0, V.jsx)(ct, {
		kind: e.timeEnabled === !1 ? "date" : "time",
		label: e.label,
		align: "left",
		from: e.from ?? "",
		to: e.to ?? "",
		onApply: e.onApply,
		...e.disabled === void 0 ? {} : { disabled: e.disabled },
		...e.presets ? { presets: e.presets } : {},
		...e.timeEnabled === void 0 ? {} : { timeEnabled: e.timeEnabled },
		...e.timeZone ? { timeZone: e.timeZone } : {},
		...e.timeZones ? { timeZones: e.timeZones } : {},
		...e.fromPlaceholder ? { fromPlaceholder: e.fromPlaceholder } : {},
		...e.toPlaceholder ? { toPlaceholder: e.toPlaceholder } : {}
	});
	if (e.kind === "enum") return /* @__PURE__ */ (0, V.jsxs)("select", {
		id: t,
		"aria-label": e.label,
		className: xo,
		value: e.value,
		disabled: e.disabled,
		onChange: (t) => e.onChange(t.target.value),
		children: [/* @__PURE__ */ (0, V.jsx)("option", {
			value: "",
			children: e.placeholder ?? ""
		}), e.options.map((e) => /* @__PURE__ */ (0, V.jsx)("option", {
			value: e.value,
			children: e.label ?? e.value
		}, e.value))]
	});
	if (e.kind === "boolean") return /* @__PURE__ */ (0, V.jsx)("div", {
		className: "flex h-8 items-center",
		children: /* @__PURE__ */ (0, V.jsx)("input", {
			id: t,
			type: "checkbox",
			"aria-label": e.label,
			className: "h-4 w-4 accent-primary disabled:cursor-not-allowed",
			checked: e.value,
			disabled: e.disabled,
			onChange: (t) => e.onChange(t.target.checked)
		})
	});
	if (e.kind === "lookup") {
		let n = `${t}-options`;
		return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsx)("input", {
			id: t,
			type: e.inputType === "number" ? "number" : "text",
			"aria-label": e.label,
			className: xo,
			...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
			value: e.value,
			list: n,
			disabled: e.disabled,
			onChange: (t) => e.onChange(t.target.value)
		}), /* @__PURE__ */ (0, V.jsx)("datalist", {
			id: n,
			children: e.options.map((e) => /* @__PURE__ */ (0, V.jsx)("option", {
				value: e.value,
				label: e.label ?? e.value,
				disabled: e.disabled
			}, e.value))
		})] });
	}
	if (e.kind === "lookup-multi") {
		let n = `${t}-options`;
		return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsx)("input", {
			id: t,
			type: "text",
			"aria-label": e.label,
			className: xo,
			...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
			value: e.value.join(", "),
			list: n,
			disabled: e.disabled,
			onChange: (t) => e.onChange(bo(t.target.value))
		}), /* @__PURE__ */ (0, V.jsx)("datalist", {
			id: n,
			children: e.options.map((e) => /* @__PURE__ */ (0, V.jsx)("option", {
				value: e.value,
				label: e.label ?? e.value,
				disabled: e.disabled
			}, e.value))
		})] });
	}
	return e.kind === "multi" ? /* @__PURE__ */ (0, V.jsx)(vo, {
		filter: e,
		id: t
	}) : e.kind === "number" || e.kind === "duration" ? /* @__PURE__ */ (0, V.jsx)(at, {
		filter: e,
		chrome: "embedded",
		autoSubmit: n
	}) : e.kind === "text" ? /* @__PURE__ */ (0, V.jsx)("input", {
		id: t,
		type: "text",
		"aria-label": e.label,
		className: xo,
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		value: e.value,
		disabled: e.disabled,
		onChange: (t) => e.onChange(t.target.value)
	}) : /* @__PURE__ */ (0, V.jsx)("input", {
		id: t,
		type: "text",
		"aria-label": e.label,
		className: xo,
		value: JSON.stringify(e.value),
		disabled: !0,
		readOnly: !0
	});
}
function vo({ filter: e, id: t }) {
	let n = e.truncated && e.total ? Math.max(e.total - e.options.length, 0) : 0;
	return /* @__PURE__ */ (0, V.jsx)(I, {
		multiple: !0,
		id: t,
		ariaLabel: e.label,
		value: Object.keys(e.value),
		options: e.options.map((e) => ({
			value: e.value,
			label: typeof e.label == "string" ? e.label : e.title ?? e.value,
			...e.disabled === void 0 ? {} : { disabled: e.disabled },
			...e.title === void 0 ? {} : { title: e.title }
		})),
		onChange: (t) => {
			let n = {};
			for (let r of t) n[r] = e.value[r] ?? "include";
			e.onChange(n);
		},
		allowCustomValue: e.allowCustomValue ?? !1,
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...e.disabled === void 0 ? {} : { disabled: e.disabled },
		...e.className === void 0 ? {} : { className: e.className },
		...n > 0 ? { footer: `… and ${n.toLocaleString()} more` } : {}
	});
}
function yo({ timeRange: e }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "grid grid-cols-1 gap-2 py-2 sm:grid-cols-[minmax(8rem,14rem)_minmax(0,1fr)] sm:items-center",
		children: [/* @__PURE__ */ (0, V.jsx)("div", {
			className: "text-sm font-medium text-muted-foreground",
			children: "Time range"
		}), /* @__PURE__ */ (0, V.jsx)("div", {
			className: "min-w-0",
			children: /* @__PURE__ */ (0, V.jsx)(ct, {
				kind: e.timeEnabled ? "time" : "date",
				label: "Time range",
				align: "left",
				from: e.from ?? "",
				to: e.to ?? "",
				onApply: e.onApply,
				...e.presets ? { presets: e.presets } : {},
				...e.timeEnabled === void 0 ? {} : { timeEnabled: e.timeEnabled },
				...e.timeZone ? { timeZone: e.timeZone } : {},
				...e.timeZones ? { timeZones: e.timeZones } : {},
				...e.fromPlaceholder ? { fromPlaceholder: e.fromPlaceholder } : {},
				...e.toPlaceholder ? { toPlaceholder: e.toPlaceholder } : {}
			})
		})]
	});
}
function bo(e) {
	return e.split(",").map((e) => e.trim()).filter(Boolean);
}
var xo = "h-8 w-full min-w-0 rounded-md border border-input bg-background px-2 text-sm text-foreground outline-none placeholder:text-placeholder disabled:cursor-not-allowed disabled:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring";
// ../../../clicky-ui/packages/ui/src/rpc/FilterForm.tsx
function So({ client: e, path: t, method: n, parameters: r = [], initialValues: i = {}, lockedValues: a = {}, hideLocked: s = !1, enableLookup: c = n.toUpperCase() === "GET", autoSubmit: l = !1, submitLabel: u = "Execute request", submittingLabel: d = "Executing…", emptyMessage: f = "This operation does not require input.", isSubmitting: p = !1, className: m, formId: h, showSubmit: g = !0, excludedParameterNames: _ = [], presentation: v = "form", onChange: y, onSubmit: b }) {
	let x = (0, B.useMemo)(() => `${n}:${t}:${JSON.stringify(i)}:${JSON.stringify(a)}`, [
		i,
		a,
		n,
		t
	]), S = (0, B.useMemo)(() => r.filter((e) => !_.includes(e.name)), [_, r]), [C, w] = (0, B.useState)(() => Ae(r, n, a, i)), [T, E] = (0, B.useState)(""), D = be(C, 250), O = (0, B.useRef)(null), ee = (0, B.useRef)(y);
	(0, B.useEffect)(() => {
		ee.current = y;
	}), (0, B.useEffect)(() => {
		ee.current?.(C);
	}, [C]);
	let te = (0, B.useMemo)(() => Ee(D, r), [D, r]), ne = (0, fo.useQuery)({
		queryKey: [
			"filter-form-lookup",
			n,
			t,
			te
		],
		queryFn: async () => await e.lookupFilters?.(t, n, te, { Accept: "application/json+clicky" }) ?? { filters: {} },
		enabled: c && !!e.lookupFilters && r.some((e) => e.in === "query"),
		placeholderData: fo.keepPreviousData,
		staleTime: 3e4,
		retry: 0
	}), { data: re } = po(e), k = re?.components?.["x-clicky-filters"], A = (0, B.useMemo)(() => Le(S, C, w, {
		lookup: ne.data,
		components: k,
		lockedValues: a,
		hideLocked: s
	}), [
		k,
		s,
		a,
		ne.data,
		S,
		C
	]), ie = A.filters.length > 0 || A.search != null || A.timeRange != null;
	async function ae(e) {
		e?.preventDefault();
		let t = r.filter((e) => e.required ? (a[e.name] ?? C[e.name] ?? "").trim() === "" : !1);
		t.length > 0 ? E(`Missing required fields: ${t.map((e) => xe(e.name)).join(", ")}`) : (E(""), await b(De(C)));
	}
	return (0, B.useEffect)(() => {
		w(Ae(r, n, a, i)), E(""), O.current = null;
	}, [x]), (0, B.useEffect)(() => {
		if (!l || r.filter((e) => e.required ? (a[e.name] ?? D[e.name] ?? "").trim() === "" : !1).length > 0) return;
		let e = De(D), t = JSON.stringify(e);
		O.current !== t && (O.current = t, E(""), b(e));
	}, [
		l,
		D,
		a,
		b,
		r
	]), /* @__PURE__ */ (0, V.jsxs)("form", {
		id: h,
		className: "space-y-3",
		onSubmit: ae,
		children: [ie ? v === "filter-bar" ? /* @__PURE__ */ (0, V.jsx)(Ht, {
			autoSubmit: l,
			filters: A.filters,
			overflowMode: "wrap",
			...A.search ? { search: A.search } : {},
			...A.timeRange ? { timeRange: A.timeRange } : {},
			...l ? {} : {
				onApply: () => void ae(),
				applyLabel: u,
				isPending: p
			},
			...m ? { className: m } : {}
		}) : /* @__PURE__ */ (0, V.jsx)(ho, {
			autoSubmit: l,
			filters: A.filters,
			isSubmitting: p,
			submitLabel: u,
			submittingLabel: d,
			showSubmit: g,
			...A.timeRange ? { timeRange: A.timeRange } : {},
			...m ? { className: m } : {}
		}) : /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex items-center justify-between gap-4",
			children: [/* @__PURE__ */ (0, V.jsx)("p", {
				className: "text-sm text-muted-foreground",
				children: f
			}), g ? /* @__PURE__ */ (0, V.jsx)(o, {
				type: "button",
				onClick: ae,
				disabled: p,
				children: p ? d : u
			}) : null]
		}), T && /* @__PURE__ */ (0, V.jsx)("div", {
			className: "rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive",
			children: T
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/rpc/operationErrorDiagnostics.tsx
function Co(e, t) {
	return /* @__PURE__ */ (0, V.jsx)(_e, {
		title: t,
		diagnostics: wo(e)
	});
}
function wo(e) {
	let t = jo(e), n = e instanceof Error ? e.message : String(e ?? ""), r = t?.responseData, i = Mo(t, "responseBody"), a = N(r ?? e, n) ?? {
		message: n || "Request failed",
		context: []
	}, o = To(r), s = [
		...Eo(t),
		...Do(r),
		...Oo(o),
		...a.context
	], c = [
		...ko(o),
		...Ao(i, a.message),
		...a.details ?? []
	], l = a.stacktrace ?? (e instanceof Error ? e.stack : void 0);
	return {
		...a,
		message: a.message || n || "Request failed",
		context: s,
		raw: e,
		...c.length > 0 ? { details: c } : {},
		...l === void 0 ? {} : { stacktrace: l }
	};
}
function To(e) {
	return jo(jo(e)?.diagnostics) || null;
}
function Eo(e) {
	if (!e) return [];
	let t = [], n = Mo(e, "method"), r = Mo(e, "url"), i = e.status;
	return n && t.push(["Method", n]), r && t.push(["URL", r]), typeof i == "number" && t.push(["Status", String(i)]), t;
}
function Do(e) {
	let t = Mo(jo(e), "code");
	return t ? [["Code", t]] : [];
}
function Oo(e) {
	if (!e) return [];
	let t = [];
	e.provider && t.push(["Provider", e.provider]);
	let n = e.response?.durationMs;
	typeof n == "number" && t.push(["Duration", `${n}ms`]);
	let r = e.response?.returnedRows;
	return typeof r == "number" && t.push(["Rows", String(r)]), t;
}
function ko(e) {
	if (!e?.request) return [];
	let t = [], n = e.request.query;
	typeof n == "string" && n.trim() && t.push({
		label: "Query",
		value: n
	});
	let r = e.request.arguments;
	Array.isArray(r) && r.length > 0 && t.push({
		label: "Arguments",
		value: JSON.stringify(r, null, 2)
	});
	let i = jo(e.request.options);
	return i && Object.keys(i).length > 0 && t.push({
		label: "Options",
		value: JSON.stringify(i, null, 2)
	}), t;
}
function Ao(e, t) {
	return !e || e === t ? [] : [{
		label: "Response",
		value: e
	}];
}
function jo(e) {
	return !e || typeof e != "object" || Array.isArray(e) ? null : e;
}
function Mo(e, t) {
	let n = e?.[t];
	return typeof n == "string" && n.trim() ? n : void 0;
}
// ../../../clicky-ui/packages/ui/src/rpc/classify.ts
function No(e) {
	if (!e) return null;
	let t = e.stdout || e.output || "";
	if (!t.trim()) return null;
	try {
		return JSON.parse(t);
	} catch {
		return null;
	}
}
// ../../../clicky-ui/packages/ui/src/data/object-graph-nodes.ts
function Po(e) {
	if (e.raw) return !1;
	if (e.value == null) return !0;
	if (typeof e.value != "string") return !1;
	let t = e.value.trim();
	return t === "" || t === "null";
}
function Fo(e) {
	return e.flatMap((e) => {
		let t = e.children ? Fo(e.children) : void 0;
		return t && t.length > 0 ? [{
			...e,
			children: t
		}] : e.expandable ? [t ? {
			...e,
			children: t
		} : e] : e.children && e.children.length > 0 || Po(e) ? [] : [e];
	});
}
function Io(e) {
	return e.reduce((e, t) => e + 1 + Io(t.children ?? []), 0);
}
// ../../../clicky-ui/packages/ui/src/data/ObjectGraph.tsx
function Lo({ roots: e, className: t, empty: n, showControls: r, defaultOpenDepth: i = 2, renderLabel: a, renderValue: o, renderType: s, loadChildren: l, onNodeSelect: u, selectedId: d, showOptionsMenu: f = !1, defaultHideNulls: p = !1, menuActions: m, showFullscreenControl: h = !1, fullscreenTitle: g = "Object", fullscreenButtonLabel: _ = "Open object browser full screen" }) {
	let [v, y] = (0, B.useState)(p), [b, x] = (0, B.useState)(!1), S = (0, B.useMemo)(() => v ? Fo(e) : e, [v, e]), C = (0, B.useMemo)(() => v ? Io(e) - Io(S) : 0, [
		v,
		e,
		S
	]), T = [{
		label: v && C > 0 ? `Hide null values (${C})` : "Hide null values",
		onSelect: () => y((e) => !e),
		...v ? { icon: w } : {},
		title: "Drop fields whose value is null or blank"
	}, ...m ?? []], E = ({ inFullscreen: p }) => {
		let m = f || h && !p ? /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [h && !p && /* @__PURE__ */ (0, V.jsx)(Ro, {
			label: _,
			onClick: () => x(!0),
			children: /* @__PURE__ */ (0, V.jsx)(c, {
				icon: te,
				className: "text-sm"
			})
		}), f && /* @__PURE__ */ (0, V.jsx)(ne, {
			items: T,
			align: "right",
			menuLabel: "Display options",
			trigger: /* @__PURE__ */ (0, V.jsx)(Ro, {
				label: "Display options",
				children: /* @__PURE__ */ (0, V.jsx)(c, {
					icon: he,
					className: "text-sm"
				})
			})
		})] }) : void 0;
		return /* @__PURE__ */ (0, V.jsx)($e, {
			roots: S,
			getKey: (e) => e.id,
			getChildren: (e) => e.children,
			defaultOpen: (e, t) => t < i,
			...t !== void 0 && !p ? { className: t } : {},
			...m === void 0 ? {} : { actions: m },
			empty: v && e.length > 0 ? /* @__PURE__ */ (0, V.jsx)("div", {
				className: "px-3 py-4 text-sm text-muted-foreground",
				children: "All values are null."
			}) : n,
			...r === void 0 ? {} : { showControls: r },
			...l ? {
				loadChildren: l,
				hasMoreChildren: (e) => e.expandable === !0
			} : {},
			renderRow: ({ node: e, loading: t, error: n }) => /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex min-w-0 flex-1 flex-wrap items-center gap-2 font-mono text-xs",
				children: [
					a ? a(e) : /* @__PURE__ */ (0, V.jsx)(zo, {
						node: e,
						selected: d != null && e.id === d,
						...o ? { renderValue: o } : {},
						...s ? { renderType: s } : {},
						...u ? { onSelect: u } : {}
					}),
					t && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "shrink-0 text-muted-foreground",
						children: "…"
					}),
					n != null && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "break-all text-red-600",
						children: n instanceof Error ? n.message : String(n)
					})
				]
			})
		});
	};
	return h ? /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [E({ inFullscreen: !1 }), /* @__PURE__ */ (0, V.jsx)(A, {
		open: b,
		onClose: () => x(!1),
		title: g,
		size: "full",
		className: "h-[95vh]",
		scrollBody: !1,
		children: /* @__PURE__ */ (0, V.jsx)("div", {
			className: "flex h-full min-h-0 flex-col",
			children: E({ inFullscreen: !0 })
		})
	})] }) : E({ inFullscreen: !1 });
}
var Ro = (0, B.forwardRef)(function({ label: e, className: t, children: n, ...r }, a) {
	return /* @__PURE__ */ (0, V.jsx)("button", {
		ref: a,
		type: "button",
		"aria-label": e,
		title: e,
		className: i("inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-accent-foreground", t),
		...r,
		children: n
	});
});
function zo({ node: e, onSelect: t, renderValue: n, renderType: r, selected: a = !1 }) {
	let o = t ? /* @__PURE__ */ (0, V.jsx)("button", {
		type: "button",
		onClick: (n) => {
			n.stopPropagation(), t(e);
		},
		title: "Select this path",
		className: i("rounded px-0.5 text-left text-muted-foreground hover:text-foreground hover:underline", a && "bg-primary/15 text-foreground"),
		children: e.label
	}) : /* @__PURE__ */ (0, V.jsx)("span", {
		className: "text-muted-foreground",
		children: e.label
	}), s = n?.(e);
	return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
		o,
		e.kind && [
			"opaque",
			"identity",
			"truncated",
			"error"
		].includes(e.kind) && /* @__PURE__ */ (0, V.jsx)("span", {
			className: i("rounded border border-border px-1 text-[10px]", e.kind === "error" ? "text-destructive" : "text-muted-foreground"),
			children: e.kind
		}),
		e.type && (r ? r(e) : /* @__PURE__ */ (0, V.jsxs)("span", {
			className: "text-muted-foreground/60",
			children: ["@", e.type]
		})),
		s ?? (e.value == null ? null : /* @__PURE__ */ (0, V.jsx)("span", {
			className: "break-all text-foreground",
			children: String(e.value)
		})),
		s == null && e.raw && !e.children?.length && e.raw !== String(e.value) && /* @__PURE__ */ (0, V.jsx)("span", {
			className: "break-all italic text-muted-foreground",
			children: e.raw
		})
	] });
}
// ../../../clicky-ui/packages/ui/src/data/ExecutionTree.tsx
function Bo({ roots: e, className: t, empty: n, showControls: r, defaultOpenDepth: i = 1, costThreshold: a, renderRow: o, renderDetail: s, renderAfterChildren: c, loadChildren: l }) {
	return /* @__PURE__ */ (0, V.jsx)($e, {
		roots: e,
		getKey: (e) => e.id,
		getChildren: (e) => e.children,
		defaultOpen: (e, t) => t < i,
		...t === void 0 ? {} : { className: t },
		...n === void 0 ? {} : { empty: n },
		...r === void 0 ? {} : { showControls: r },
		...s ? { renderDetail: s } : {},
		...c ? { renderAfterChildren: c } : {},
		...l ? {
			loadChildren: l,
			hasMoreChildren: (e) => e.expandable === !0
		} : {},
		renderRow: ({ node: e }) => o ? o(e) : /* @__PURE__ */ (0, V.jsx)(Vo, {
			node: e,
			costThreshold: a
		})
	});
}
function Vo({ node: e, costThreshold: t }) {
	let n = t != null && e.cost != null && e.cost >= t;
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "flex min-w-0 flex-1 flex-wrap items-center gap-2 font-mono text-xs",
		children: [
			e.status && /* @__PURE__ */ (0, V.jsx)(Wo, { status: e.status }),
			/* @__PURE__ */ (0, V.jsx)("span", {
				className: "font-semibold text-foreground",
				children: e.label
			}),
			e.cost != null && /* @__PURE__ */ (0, V.jsx)("span", {
				className: n ? "text-red-600" : "text-muted-foreground",
				children: Ho(e)
			}),
			e.times != null && e.times > 1 && /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "text-muted-foreground/60",
				children: ["×", e.times]
			}),
			e.className && /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "break-all text-muted-foreground/60",
				children: [e.className, e.lineNumber != null && e.lineNumber > 0 ? `:${e.lineNumber}` : ""]
			}),
			e.detail && Object.entries(e.detail).map(([e, t]) => /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "text-muted-foreground/60",
				children: [
					e,
					"=",
					t
				]
			}, e))
		]
	});
}
function Ho(e) {
	let t = e.unit || "ms", n = e.cost ?? 0;
	return t === "ms" ? `${n.toFixed(3)}ms` : `${n} ${t}`;
}
var Uo = {
	ok: "text-green-600",
	success: "text-green-600",
	error: "text-red-600",
	warning: "text-amber-500",
	running: "text-blue-600",
	pending: "text-muted-foreground"
};
function Wo({ status: e }) {
	return /* @__PURE__ */ (0, V.jsx)("span", {
		className: `shrink-0 ${Uo[e] ?? "text-muted-foreground"}`,
		title: e,
		children: "●"
	});
}
// ../../../clicky-ui/packages/ui/src/data/clicky-icon-glyphs.ts
var Go = {
	"ion:warning": l,
	"ion:information-circle": p,
	"ion:information": p,
	"ion:checkmark": w,
	"ion:close": d
}, Ko = [
	"java.",
	"javax.",
	"sun.",
	"jdk.",
	"com.sun.",
	"oracle.jrockit."
];
function qo(e, t = []) {
	return Ko.some((t) => e.startsWith(t)) || t.some((t) => e.startsWith(t));
}
function Jo(e, t) {
	return [e.split(".").filter(Boolean).pop() ?? e, t].filter(Boolean).join(".");
}
function Yo(e, t) {
	if (e) return t === void 0 ? e : `${e}:${t}`;
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/stacktrace-parse.ts
var Xo = /^\s*at\s+([\w$.]+)\.([\w$<>]+)\(([^)]+)\)/, Zo = /^\.\.\.\s+\d+\s+more$/;
function Qo(e) {
	let t = {
		causedBy: [],
		frames: [],
		language: "java"
	};
	if (!e || !e.trim()) return t;
	let n = [];
	for (let r of e.split("\n")) {
		let e = r.trim();
		if (!e || Zo.test(e)) continue;
		let i = Xo.exec(r);
		if (i && i[1] && i[2] && i[3] !== void 0) t.frames.push(es(i[1], i[2], i[3]));
		else if (e.startsWith("Caused by:")) t.causedBy.push(e.slice(10).trim());
		else if (e.startsWith("Internal Exception:")) t.causedBy.push(e.slice(19).trim());
		else {
			if (!t.exceptionClass) {
				let n = $o(e);
				if (n) {
					t.exceptionClass = n.exceptionClass, n.message && (t.message = n.message);
					continue;
				}
			}
			n.push(e);
		}
	}
	return !t.message && n.length > 0 && (t.message = n.join(" ")), t;
}
function $o(e) {
	let t = e.indexOf(":"), n = t === -1 ? e : e.slice(0, t);
	if (!n || ![
		"Exception",
		"Error",
		"Throwable"
	].some((e) => n.endsWith(e))) return null;
	for (let e of n) {
		let t = e.charCodeAt(0);
		if (!(t >= 48 && t <= 57 || t >= 65 && t <= 90 || t >= 97 && t <= 122 || e === "_" || e === "." || e === "$")) return null;
	}
	if (t === -1) return { exceptionClass: n };
	let r = e.slice(t + 1).trimStart();
	return r ? {
		exceptionClass: n,
		message: r
	} : { exceptionClass: n };
}
function es(e, t, n) {
	let r = n.trim(), i, a, o = !1;
	if (r === "Native Method") o = !0;
	else if (r !== "Unknown Source") {
		let e = (r.split(" ~[")[0] ?? r).trim(), t = e.lastIndexOf(":");
		if (t >= 0) {
			i = e.slice(0, t);
			let n = Number(e.slice(t + 1));
			Number.isFinite(n) && (a = n);
		} else i = e;
	}
	let s = `${e}.${t}`, c = Yo(i, a) ?? (o ? "Native Method" : void 0), l = {
		functionName: s,
		displayName: Jo(e, t),
		kind: "frame",
		runtime: qo(e),
		nativeMethod: o,
		class: e,
		method: t
	};
	return i !== void 0 && (l.file = i), a !== void 0 && (l.line = a), c !== void 0 && (l.location = c), l;
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/FrameSourceWindow.tsx
function ts({ frame: e, className: t, focalTone: n = "error", markers: r }) {
	let a = (0, B.useMemo)(() => e.sourceLines ?? [], [e.sourceLines]), o = e.sourceStartLine ?? 0, s = e.sourceLineNumbers, c = e.line ?? -1, l = e.sourceLanguage, [u, d] = (0, B.useState)(null);
	if ((0, B.useEffect)(() => {
		if (a.length === 0) {
			d(null);
			return;
		}
		let e = !1;
		return m(a.join("\n"), { lang: l }).then((t) => {
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
	}[n], h = new Map(r?.map((e) => [e.line, e.content]));
	return /* @__PURE__ */ (0, V.jsx)("div", {
		"data-scroll-x": !0,
		className: i("mx-3 mb-2 ml-12 overflow-x-auto max-sm:ml-6 rounded border border-border/60 bg-muted/30 py-1 font-mono text-[11px] leading-relaxed", t),
		children: a.map((e, t) => {
			let r = s?.[t] ?? o + t, i = s === void 0 || r > 0, a = i && r === c, l = i ? h.get(r) : void 0;
			return /* @__PURE__ */ (0, V.jsxs)("div", {
				"data-source-line": i ? r : void 0,
				"data-source-tone": a ? n : void 0,
				"data-marked": l === void 0 ? void 0 : "true",
				className: ["grid min-w-max grid-cols-[3.5rem_minmax(24rem,1fr)] gap-3 px-2", a ? p : l === void 0 ? "text-foreground" : f].join(" "),
				children: [
					/* @__PURE__ */ (0, V.jsxs)("span", {
						className: "select-none text-right text-muted-foreground",
						children: [a || l !== void 0 ? ">" : "", i ? r : ""]
					}),
					/* @__PURE__ */ (0, V.jsx)("code", {
						className: "whitespace-pre",
						children: e ? /* @__PURE__ */ (0, V.jsx)(y, {
							tokens: u?.[t],
							content: e
						}) : " "
					}),
					l === void 0 ? null : /* @__PURE__ */ (0, V.jsx)("span", {
						"data-source-marker": !0,
						className: "col-start-2 -mt-0.5 mb-0.5 font-sans font-normal",
						children: l
					})
				]
			}, `${r}-${t}`);
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/FrameSourceWindow.utils.ts
function ns(e) {
	return !!e.sourceLines && e.sourceLines.length > 0;
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/StackFrameRow.tsx
function rs({ frame: e, index: t, showIndex: n = !1, frameActions: r, renderFramePart: i, renderFrameDetail: a, frameActionsVisibility: o = "hover" }) {
	let s = e.kind !== "frame", l = e.displayName || e.method || e.functionName, u = r?.(e, t), d = s ? null : a?.(e, t);
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: [
			"group grid gap-x-2 gap-y-0.5 px-3 py-0.5 text-xs",
			n ? "grid-cols-[2rem_minmax(0,1fr)_auto]" : "grid-cols-[1.25rem_minmax(0,1fr)_auto]",
			e.runtime ? "text-muted-foreground" : "text-foreground"
		].join(" "),
		children: [
			/* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex items-start justify-end gap-1 pt-0.5 font-mono text-[10px] text-muted-foreground",
				children: [n && /* @__PURE__ */ (0, V.jsx)("span", {
					"data-frame-index": !0,
					children: t + 1
				}), /* @__PURE__ */ (0, V.jsx)(c, {
					icon: is(e),
					className: "mt-px shrink-0 text-[11px]"
				})]
			}),
			/* @__PURE__ */ (0, V.jsx)("div", {
				className: "min-w-0",
				children: /* @__PURE__ */ (0, V.jsx)("div", {
					className: "flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-0.5",
					children: s ? /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsx)("span", {
						className: "min-w-0 break-all font-mono font-semibold leading-4 opacity-70",
						children: e.functionName
					}), e.annotationText && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "min-w-0 break-all font-mono text-[11px] font-normal opacity-80",
						children: e.annotationText
					})] }) : /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
						i ? i("method", e, t) : /* @__PURE__ */ (0, V.jsx)("span", {
							className: "min-w-0 break-all font-mono font-semibold leading-4",
							children: l
						}),
						e.class && (i ? i("class", e, t) : /* @__PURE__ */ (0, V.jsx)("span", {
							className: "min-w-0 break-all font-mono text-[11px] text-muted-foreground",
							children: e.class
						})),
						e.location && (i ? i("location", e, t) : /* @__PURE__ */ (0, V.jsx)("span", {
							className: "rounded border border-border bg-muted/40 px-1.5 py-px font-mono text-[10px] text-muted-foreground",
							children: e.location
						}))
					] })
				})
			}),
			u ? /* @__PURE__ */ (0, V.jsx)("div", {
				className: ["flex shrink-0 items-start gap-1", o === "hover" ? "opacity-0 transition-opacity focus-within:opacity-100 group-hover:opacity-100 [@media(hover:none)]:opacity-100" : "opacity-100"].join(" "),
				children: u
			}) : /* @__PURE__ */ (0, V.jsx)("div", {}),
			!s && ns(e) ? /* @__PURE__ */ (0, V.jsx)("div", {
				className: "col-span-3",
				children: /* @__PURE__ */ (0, V.jsx)(ts, { frame: e })
			}) : null,
			d ? /* @__PURE__ */ (0, V.jsx)("div", {
				className: "col-start-2 col-span-2 min-w-0",
				children: d
			}) : null
		]
	});
}
function is(e) {
	switch (e.kind) {
		case "locked": return T;
		case "waiting_to_lock": return sn;
		case "waiting_on":
		case "parking": return rt;
	}
	return e.nativeMethod ? Cn : e.runtime ? fe : e.method || e.class ? ve : rn;
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/RenderedStackTrace.tsx
function as({ input: e, language: t = "java", resolver: n, contextLines: r = 3, hideRuntimeOnly: i = !1, include: a, exclude: o, frameActions: s, renderFramePart: l, renderFrameDetail: u, frameActionsVisibility: d = "hover", showFrameIndex: f = !0, frameFilter: p, renderCauseDetail: m, className: h }) {
	let g = (0, B.useMemo)(() => typeof e == "string" ? Qo(e) : e, [e, t]), _ = (0, B.useMemo)(() => n ? g.frames.map((e) => {
		if (e.sourceLines && e.sourceLines.length > 0) return e;
		let t = n(e, r);
		if (!t) return e;
		let { line: i, ...a } = t;
		return {
			...e,
			...a
		};
	}) : g.frames, [
		g.frames,
		n,
		r
	]), v = (0, B.useMemo)(() => _.filter((e) => i && e.runtime || p && !p(e) || o && o.some((t) => e.class?.startsWith(t)) ? !1 : a && a.length > 0 ? a.some((t) => e.class?.startsWith(t)) : !0), [
		_,
		i,
		a,
		o,
		p
	]);
	if (v.length === 0 && !g.exceptionClass) return null;
	let y = v.length !== _.length;
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: ["overflow-hidden rounded-md border border-border bg-background text-foreground", h].filter(Boolean).join(" "),
		children: [
			(g.exceptionClass || g.message) && /* @__PURE__ */ (0, V.jsx)("div", {
				className: "border-b border-border bg-red-500/5 px-3 py-2",
				children: /* @__PURE__ */ (0, V.jsxs)("div", {
					className: "flex min-w-0 items-start gap-2",
					children: [/* @__PURE__ */ (0, V.jsx)(c, {
						icon: M,
						className: "mt-0.5 shrink-0 text-sm text-red-600 dark:text-red-400"
					}), /* @__PURE__ */ (0, V.jsxs)("div", {
						className: "min-w-0 flex-1",
						children: [g.exceptionClass && /* @__PURE__ */ (0, V.jsx)("div", {
							className: "break-all font-mono text-xs font-semibold text-red-700 dark:text-red-300",
							children: g.exceptionClass
						}), g.message && /* @__PURE__ */ (0, V.jsx)("div", {
							className: "mt-0.5 whitespace-pre-wrap break-words text-xs text-foreground",
							children: g.message
						})]
					})]
				})
			}),
			g.causedBy.length > 0 && /* @__PURE__ */ (0, V.jsx)("div", {
				className: "space-y-1 border-b border-border bg-orange-500/5 px-3 py-2",
				children: g.causedBy.map((e, t) => {
					let n = m?.(e, t);
					return /* @__PURE__ */ (0, V.jsxs)("div", {
						"data-cause-index": t,
						className: "min-w-0",
						children: [/* @__PURE__ */ (0, V.jsxs)("div", {
							className: "flex min-w-0 items-start gap-2 font-mono text-[11px] text-orange-700 dark:text-orange-300",
							children: [
								/* @__PURE__ */ (0, V.jsx)(c, {
									icon: rn,
									className: "mt-0.5 shrink-0 text-xs"
								}),
								/* @__PURE__ */ (0, V.jsx)("span", {
									className: "shrink-0 opacity-75",
									children: "Caused by"
								}),
								/* @__PURE__ */ (0, V.jsx)("span", {
									className: "min-w-0 break-all",
									children: e
								})
							]
						}), n ? /* @__PURE__ */ (0, V.jsx)("div", {
							className: "mt-1 pl-5",
							children: n
						}) : null]
					}, t);
				})
			}),
			/* @__PURE__ */ (0, V.jsx)("div", {
				className: "divide-y divide-border/60",
				children: v.map((e, t) => /* @__PURE__ */ (0, V.jsx)(rs, {
					frame: e,
					index: t,
					showIndex: f,
					...s ? { frameActions: s } : {},
					...l ? { renderFramePart: l } : {},
					...u ? { renderFrameDetail: u } : {},
					frameActionsVisibility: d
				}, `${e.functionName}-${t}`))
			}),
			y && /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "border-t border-border bg-muted/30 px-3 py-1.5 text-[11px] text-muted-foreground",
				children: [
					"Showing ",
					v.length,
					" of ",
					_.length,
					" frames"
				]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/clicky-parse.ts
function os(e) {
	if (typeof e == "string") try {
		return ss(JSON.parse(e));
	} catch (t) {
		return {
			ok: !1,
			message: t instanceof Error ? t.message : "Failed to parse JSON",
			raw: e
		};
	}
	return ss(e);
}
function ss(e) {
	if (!e || typeof e != "object") return {
		ok: !1,
		message: "Payload must be an object",
		raw: String(e ?? "")
	};
	let t = e;
	return "version" in t && t.version === 1 && t.node && cs(t.node) ? {
		ok: !0,
		document: t
	} : cs(t) ? {
		ok: !0,
		document: {
			version: 1,
			node: t
		}
	} : {
		ok: !1,
		message: "Payload is neither a Clicky document nor a Clicky node",
		raw: JSON.stringify(e, null, 2)
	};
}
function cs(e) {
	return !!e && typeof e == "object" && typeof e.kind == "string";
}
// ../../../clicky-ui/packages/ui/src/data/clicky-export/ClickyExportDialog.tsx
function ls({ open: e, onClose: t, title: n = "Export", formats: r, scopes: a, onDownload: s, pendingContent: l, allowCloseWhilePending: u = !1 }) {
	let [d, f] = (0, B.useState)(r[0]?.format), [p, m] = (0, B.useState)(a[0]?.scope), [h, g] = (0, B.useState)(!1), [_, v] = (0, B.useState)("");
	(0, B.useEffect)(() => {
		e && v("");
	}, [e]);
	let y = a.find((e) => e.scope === p), b = async () => {
		if (d) {
			g(!0), v("");
			try {
				await s({
					format: d,
					scope: p
				}), t();
			} catch (e) {
				v(e instanceof Error ? e.message : "The export failed");
			} finally {
				g(!1);
			}
		}
	};
	return /* @__PURE__ */ (0, V.jsx)(A, {
		open: e,
		onClose: t,
		title: n,
		size: "lg",
		closeOnEsc: !0,
		footer: /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex items-center justify-end gap-2",
			children: [/* @__PURE__ */ (0, V.jsx)(o, {
				variant: "outline",
				onClick: t,
				disabled: h && !u,
				children: h && u ? "Continue in background" : "Cancel"
			}), /* @__PURE__ */ (0, V.jsx)(o, {
				onClick: () => void b(),
				loading: h,
				disabled: !d,
				children: "Download"
			})]
		}),
		children: /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "space-y-4",
			children: [
				h && l,
				_ && /* @__PURE__ */ (0, V.jsx)("div", {
					role: "alert",
					className: "rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive",
					children: _
				}),
				a.length > 1 ? /* @__PURE__ */ (0, V.jsxs)("fieldset", { children: [/* @__PURE__ */ (0, V.jsx)("legend", {
					className: "mb-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground",
					children: "Rows"
				}), /* @__PURE__ */ (0, V.jsx)("div", {
					className: "flex flex-wrap gap-2",
					children: a.map((e) => /* @__PURE__ */ (0, V.jsxs)("label", {
						className: i("cursor-pointer rounded-md border px-3 py-1.5 text-sm transition-colors", e.scope === p ? "border-foreground/20 bg-accent text-accent-foreground" : "border-input hover:bg-accent hover:text-accent-foreground"),
						children: [/* @__PURE__ */ (0, V.jsx)("input", {
							type: "radio",
							name: "clicky-export-scope",
							"aria-label": e.label,
							className: "sr-only",
							checked: e.scope === p,
							onChange: () => m(e.scope)
						}), e.label]
					}, e.scope ?? "default"))
				})] }) : a[0]?.label && /* @__PURE__ */ (0, V.jsx)("p", {
					className: "text-sm text-muted-foreground",
					children: a[0].label
				}),
				/* @__PURE__ */ (0, V.jsxs)("fieldset", { children: [/* @__PURE__ */ (0, V.jsx)("legend", {
					className: "mb-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground",
					children: "Format"
				}), /* @__PURE__ */ (0, V.jsx)("div", {
					className: "grid gap-1 sm:grid-cols-2",
					children: r.map((e) => {
						let t = y?.note?.(e.format);
						return /* @__PURE__ */ (0, V.jsxs)("label", {
							className: i("flex cursor-pointer items-start gap-2 rounded-md border px-3 py-2 text-left transition-colors", e.format === d ? "border-foreground/20 bg-accent text-accent-foreground" : "border-transparent hover:bg-accent hover:text-accent-foreground"),
							children: [
								/* @__PURE__ */ (0, V.jsx)("input", {
									type: "radio",
									name: "clicky-export-format",
									"aria-label": e.label,
									className: "sr-only",
									checked: e.format === d,
									onChange: () => f(e.format)
								}),
								e.icon && /* @__PURE__ */ (0, V.jsx)(c, {
									icon: e.icon,
									className: i("mt-0.5 shrink-0 text-sm", e.iconClassName ?? "text-muted-foreground")
								}),
								/* @__PURE__ */ (0, V.jsxs)("span", {
									className: "min-w-0",
									children: [
										/* @__PURE__ */ (0, V.jsx)("span", {
											className: "block text-sm font-medium",
											children: e.label
										}),
										e.description && /* @__PURE__ */ (0, V.jsx)("span", {
											className: "mt-0.5 block text-xs text-muted-foreground",
											children: e.description
										}),
										t && /* @__PURE__ */ (0, V.jsx)("span", {
											className: "mt-0.5 block text-xs text-amber-700 [[data-theme=dark]_&]:text-amber-400",
											children: t
										})
									]
								})
							]
						}, e.format);
					})
				})] })
			]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/clicky-export/useClickyExport.tsx
function us({ formats: e, scopes: t, onDownload: n, label: r = "Export…", title: i }) {
	let [a, o] = (0, B.useState)(!1), s = (0, B.useCallback)(() => o(!1), []), c = e.length > 0, l = (0, B.useMemo)(() => c ? {
		id: "export",
		label: r,
		icon: Sn,
		section: "",
		onSelect: () => o(!0)
	} : void 0, [c, r]);
	return c ? {
		action: l,
		dialog: /* @__PURE__ */ (0, V.jsx)(ls, {
			open: a,
			onClose: s,
			formats: e,
			scopes: t,
			onDownload: n,
			...i ? { title: i } : {}
		})
	} : {
		action: void 0,
		dialog: null
	};
}
// ../../../clicky-ui/packages/ui/src/data/clicky-download.ts
var ds = ["clicky", "json"], fs = [
	"pdf",
	"html",
	"markdown",
	"yaml",
	"csv",
	"pretty",
	"excel",
	"slack"
], ps = [
	"yaml",
	"json",
	"csv",
	"pdf",
	"markdown"
];
function ms({ url: e, download: t }) {
	return !(t ? t.all ?? !0 : e) || !e ? [] : t?.formats?.length ? [...t.formats] : [...ps];
}
function hs(e) {
	return e?.scopes?.length ? e.scopes : [void 0];
}
function gs(e, t, n) {
	let r = n?.formatMaxRows?.[e];
	if (r) return `Limited to ${r.toLocaleString()} rows`;
	if (t === "all" && n?.allRowsMode === "buffered") return "Compatibility mode buffers the complete result";
	if (t === "all" && n?.allRowsMode === "streaming") return "Streams rows as they are read";
}
function _s({ url: e, download: t }) {
	return e ? ms({
		url: e,
		download: t
	}).map((e) => {
		let t = Ss(e);
		return {
			format: e,
			label: t.label,
			description: t.description,
			icon: t.icon,
			...t.iconClassName ? { iconClassName: t.iconClassName } : {}
		};
	}) : [];
}
function vs(e, t) {
	return hs(e).map((n) => ({
		scope: n,
		label: n === "all" || n === void 0 ? ys(t) : "Current page",
		note: (t) => gs(t, n, e)
	}));
}
function ys(e) {
	let t = e?.total;
	if (t === void 0) return "All rows";
	let n = t.toLocaleString();
	return e?.totalRelation === "gte" ? `All ${n}+ rows` : `All ${n} rows`;
}
async function bs({ url: e, format: t, scope: n, download: r }) {
	let i = "";
	if (await xs({
		url: e,
		format: t,
		...n ? { scope: n } : {},
		...r ? { download: r } : {},
		onError: (e) => {
			i = e;
		}
	}), i) throw Error(i);
}
async function xs({ url: e, format: t, scope: n, download: r, onPending: i, onError: a }) {
	a?.(""), i?.(!0);
	try {
		let i = r?.prepare ? await r.prepare({
			format: t,
			...n ? { scope: n } : {}
		}) : {
			url: e,
			...r?.label ? { label: r.label } : {}
		};
		if (!i.url.trim()) throw Error("The prepared download URL is empty");
		As(ws(i.url, t, i.label ?? r?.label, n));
	} catch (e) {
		a?.(e instanceof Error ? e.message : "Preparing the download failed");
	} finally {
		i?.(!1);
	}
}
function Ss(e) {
	switch (e) {
		case "clicky": return {
			label: "Clicky",
			description: "Rendered Clicky JSON with the rich Clicky viewer",
			icon: Xe,
			accept: "application/json+clicky, application/clicky+json;q=0.9, application/json;q=0.8,*/*;q=0.7"
		};
		case "json": return {
			label: "JSON",
			description: "Plain JSON for inspecting the raw response body",
			icon: gn,
			accept: "application/json, text/plain;q=0.8,*/*;q=0.7"
		};
		case "ndjson": return {
			label: "NDJSON",
			description: "Newline-delimited JSON for streaming data pipelines",
			icon: gn,
			accept: "application/x-ndjson, application/ndjson;q=0.9,*/*;q=0.7"
		};
		case "pdf": return {
			label: "PDF",
			description: "Portable document for sharing and printing",
			icon: vn,
			iconClassName: "text-rose-600 dark:text-rose-400",
			accept: "application/pdf, */*;q=0.7"
		};
		case "html": return {
			label: "HTML",
			description: "Browser-ready HTML preview of the formatted output",
			icon: Ke,
			accept: "text/html, */*;q=0.7"
		};
		case "markdown": return {
			label: "Markdown",
			description: "Markdown formatted for docs, comments, and chat",
			icon: _n,
			accept: "text/markdown, text/plain;q=0.8,*/*;q=0.7"
		};
		case "yaml": return {
			label: "YAML",
			description: "YAML for config-friendly inspection and export",
			icon: Ke,
			iconClassName: "text-violet-600 dark:text-violet-400",
			accept: "application/yaml, text/yaml;q=0.9, text/plain;q=0.8,*/*;q=0.7"
		};
		case "csv": return {
			label: "CSV",
			description: "Comma-separated values for spreadsheets and imports",
			icon: yn,
			iconClassName: "text-emerald-600 dark:text-emerald-400",
			accept: "text/csv, text/plain;q=0.8,*/*;q=0.7"
		};
		case "pretty": return {
			label: "Pretty",
			description: "Human-readable plain text output from the formatter",
			icon: Ke,
			accept: "text/plain, */*;q=0.7"
		};
		case "excel": return {
			label: "Excel",
			description: "Spreadsheet workbook for offline analysis",
			icon: mn,
			accept: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet, */*;q=0.7"
		};
		case "slack": return {
			label: "Slack",
			description: "Slack Block Kit JSON for chat-native output",
			icon: me,
			accept: "application/vnd.slack.block-kit+json, application/json;q=0.8,*/*;q=0.7"
		};
	}
}
function Cs(e, t) {
	let n = typeof window > "u" ? "http://localhost" : window.location.origin, r = new URL(e, n);
	return r.searchParams.set("format", t === "clicky" ? "clicky-json" : t), js(e) ? r.toString() : `${r.pathname}${r.search}${r.hash}`;
}
function ws(e, t, n, r) {
	let i = Cs(e, t), a = typeof window > "u" ? "http://localhost" : window.location.origin, o = new URL(i, a);
	return o.searchParams.set("filename", Ts(e, t, n)), o.searchParams.set("_download", Date.now().toString()), r && o.searchParams.set("scope", r), r === "all" && (o.searchParams.delete("limit"), o.searchParams.delete("offset"), o.searchParams.delete("page")), js(i) ? o.toString() : `${o.pathname}${o.search}${o.hash}`;
}
function Ts(e, t, n) {
	return `${Ds(n || Es(e) || "clicky-output")}${Os(t)}`;
}
function Es(e) {
	let t = typeof window > "u" ? "http://localhost" : window.location.origin;
	return new URL(e, t).pathname.split("/").filter(Boolean).at(-1) ?? "";
}
function Ds(e) {
	return Ie(e.trim().replace(/\.[a-z0-9]+$/i, "").replace(/[^a-z0-9._-]+/gi, "-").replace(/-+/g, "-"), "._-") || "clicky-output";
}
function Os(e) {
	switch (e) {
		case "clicky": return ".clicky.json";
		case "json": return ".json";
		case "ndjson": return ".ndjson";
		case "pdf": return ".pdf";
		case "html": return ".html";
		case "markdown": return ".md";
		case "yaml": return ".yaml";
		case "csv": return ".csv";
		case "pretty": return ".txt";
		case "excel": return ".xlsx";
		case "slack": return ".slack.json";
	}
}
var ks = "clicky-download-frame";
function As(e) {
	if (typeof document > "u") return;
	let t = document.getElementById(ks);
	t || (t = document.createElement("iframe"), t.id = ks, t.name = ks, t.title = "Clicky download target", t.setAttribute("aria-hidden", "true"), t.tabIndex = -1, t.style.display = "none", document.body.appendChild(t)), t.src = e;
}
function js(e) {
	return /^[a-z][a-z\d+\-.]*:/i.test(e) || e.startsWith("//");
}
// ../../../clicky-ui/packages/ui/src/data/debugConsoleSignal.ts
var Ms = /* @__PURE__ */ new Set(), Ns = /* @__PURE__ */ new Set();
function Ps(e) {
	return Ns.add(e), () => {
		Ns.delete(e);
	};
}
function Fs() {
	return Ms.size > 0;
}
function Is(e = {}) {
	for (let t of Ms) t(e);
}
// ../../../clicky-ui/packages/ui/src/data/useDebugAction.tsx
function Ls() {
	let e = (0, B.useSyncExternalStore)(Ps, Fs, () => !1);
	return (0, B.useMemo)(() => e ? {
		id: "show-query",
		label: "Debug",
		icon: pn,
		section: "",
		onSelect: () => Is({ tab: "queries" })
	} : void 0, [e]);
}
// ../../../clicky-ui/packages/ui/src/data/useClickyTableMenu.ts
function Rs({ url: e, download: t, pagination: n }) {
	let r = Ls(), i = us({
		formats: (0, B.useMemo)(() => _s({
			url: e,
			download: t
		}), [t, e]),
		scopes: (0, B.useMemo)(() => vs(t, n), [t, n]),
		onDownload: (0, B.useCallback)(({ format: n, scope: r }) => bs({
			url: e ?? "",
			format: n,
			scope: r,
			download: t
		}), [t, e])
	});
	return {
		actions: (0, B.useMemo)(() => [i.action, r].filter((e) => e !== void 0), [i.action, r]),
		dialog: i.dialog
	};
}
// ../../../clicky-ui/packages/ui/src/data/Clicky.tsx
var zs = {
	operations: [],
	operationsLoading: !1
}, Bs = (0, B.createContext)(zs);
function Vs(e) {
	let t = (0, B.useContext)(fo.QueryClientContext), [n] = (0, B.useState)(() => new fo.QueryClient({ defaultOptions: { queries: {
		staleTime: 3e4,
		retry: 1
	} } })), { actions: r, dialog: i } = Rs({
		url: e.url,
		download: e.download,
		pagination: e.pagination
	}), a = (0, B.useMemo)(() => [...e.menuActions ?? [], ...r], [e.menuActions, r]), o = /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [i, /* @__PURE__ */ (0, V.jsx)(Hs, {
		...e.commandRuntime ? { commandRuntime: e.commandRuntime } : {},
		...e.onTableRowClick ? { onTableRowClick: e.onTableRowClick } : {},
		...e.getTableRowHref ? { getTableRowHref: e.getTableRowHref } : {},
		...e.isTableRowClickable ? { isTableRowClickable: e.isTableRowClickable } : {},
		...e.search ? { tableSearch: e.search } : {},
		...e.timeRange ? { tableTimeRange: e.timeRange } : {},
		...e.externalFilters ? { tableExternalFilters: e.externalFilters } : {},
		...e.cellFilters ? { tableCellFilters: e.cellFilters } : {},
		...e.onCellFilterChange ? { onTableCellFilterChange: e.onCellFilterChange } : {},
		...e.pagination ? { tablePagination: e.pagination } : {},
		...e.sort === void 0 ? {} : { tableSort: e.sort },
		...e.onSortChange ? { onTableSortChange: e.onSortChange } : {},
		...e.infinite ? { tableInfinite: e.infinite } : {},
		...e.rowSelection ? { tableRowSelection: e.rowSelection } : {},
		...e.hiddenColumns ? { tableHiddenColumns: e.hiddenColumns } : {},
		...e.loading === void 0 ? {} : { tableLoading: e.loading },
		...e.renderRowDetail ? { tableRenderRowDetail: e.renderRowDetail } : {},
		...e.cellRenderers ? { tableCellRenderers: e.cellRenderers } : {},
		...e.renderRowCard ? { tableRenderRowCard: e.renderRowCard } : {},
		...e.detailStyle ? { tableDetailStyle: e.detailStyle } : {},
		...e.detailDialogSize ? { tableDetailDialogSize: e.detailDialogSize } : {},
		...e.detailDialogTitle ? { tableDetailDialogTitle: e.detailDialogTitle } : {},
		...a.length > 0 ? { tableMenuActions: a } : {},
		children: e.url ? /* @__PURE__ */ (0, V.jsx)(Gs, {
			...e,
			url: e.url
		}) : /* @__PURE__ */ (0, V.jsx)(Ws, {
			data: e.data,
			...e.className ? { className: e.className } : {}
		})
	})] });
	return !t && (e.url || e.commandRuntime) ? /* @__PURE__ */ (0, V.jsx)(fo.QueryClientProvider, {
		client: n,
		children: o
	}) : o;
}
function Hs({ commandRuntime: e, onTableRowClick: t, getTableRowHref: n, isTableRowClickable: r, tableSearch: i, tableTimeRange: a, tableExternalFilters: o, tableCellFilters: s, onTableCellFilterChange: c, tablePagination: l, tableSort: u, onTableSortChange: d, tableInfinite: f, tableRowSelection: p, tableHiddenColumns: m, tableMenuActions: h, tableLoading: g, tableRenderRowDetail: _, tableCellRenderers: v, tableRenderRowCard: y, tableDetailStyle: b, tableDetailDialogSize: x, tableDetailDialogTitle: S, children: C }) {
	if (!e) {
		let e = t || n || r || i || a || o || s || c || l || u !== void 0 || d || f || p || m || h || g || _ || v || y || b || x || S;
		return /* @__PURE__ */ (0, V.jsx)(Bs.Provider, {
			value: e ? {
				...zs,
				onTableRowClick: t,
				getTableRowHref: n,
				isTableRowClickable: r,
				tableSearch: i,
				tableTimeRange: a,
				tableExternalFilters: o,
				tableCellFilters: s,
				onTableCellFilterChange: c,
				tablePagination: l,
				tableSort: u,
				onTableSortChange: d,
				tableInfinite: f,
				tableRowSelection: p,
				tableHiddenColumns: m,
				tableMenuActions: h,
				tableLoading: g,
				tableRenderRowDetail: _,
				tableCellRenderers: v,
				tableRenderRowCard: y,
				tableDetailStyle: b,
				tableDetailDialogSize: x,
				tableDetailDialogTitle: S
			} : zs,
			children: C
		});
	}
	return /* @__PURE__ */ (0, V.jsx)(Us, {
		commandRuntime: e,
		...t ? { onTableRowClick: t } : {},
		...n ? { getTableRowHref: n } : {},
		...r ? { isTableRowClickable: r } : {},
		...i ? { tableSearch: i } : {},
		...a ? { tableTimeRange: a } : {},
		...o ? { tableExternalFilters: o } : {},
		...s ? { tableCellFilters: s } : {},
		...c ? { onTableCellFilterChange: c } : {},
		...l ? { tablePagination: l } : {},
		...u === void 0 ? {} : { tableSort: u },
		...d ? { onTableSortChange: d } : {},
		...f ? { tableInfinite: f } : {},
		...p ? { tableRowSelection: p } : {},
		...m ? { tableHiddenColumns: m } : {},
		...h ? { tableMenuActions: h } : {},
		...g === void 0 ? {} : { tableLoading: g },
		..._ ? { tableRenderRowDetail: _ } : {},
		...v ? { tableCellRenderers: v } : {},
		...y ? { tableRenderRowCard: y } : {},
		...b ? { tableDetailStyle: b } : {},
		...x ? { tableDetailDialogSize: x } : {},
		...S ? { tableDetailDialogTitle: S } : {},
		children: C
	});
}
function Us({ commandRuntime: e, onTableRowClick: t, getTableRowHref: n, isTableRowClickable: r, tableSearch: i, tableTimeRange: a, tableExternalFilters: o, tableCellFilters: s, onTableCellFilterChange: c, tablePagination: l, tableSort: u, onTableSortChange: d, tableInfinite: f, tableRowSelection: p, tableHiddenColumns: m, tableMenuActions: h, tableLoading: g, tableRenderRowDetail: _, tableCellRenderers: v, tableRenderRowCard: y, tableDetailStyle: b, tableDetailDialogSize: x, tableDetailDialogTitle: S, children: C }) {
	let { operations: w, isLoading: T } = mo(e.client), E = (0, B.useMemo)(() => ({
		commandRuntime: e,
		onTableRowClick: t,
		getTableRowHref: n,
		isTableRowClickable: r,
		tableSearch: i,
		tableTimeRange: a,
		tableExternalFilters: o,
		tableCellFilters: s,
		onTableCellFilterChange: c,
		tablePagination: l,
		tableSort: u,
		onTableSortChange: d,
		tableInfinite: f,
		tableRowSelection: p,
		tableHiddenColumns: m,
		tableMenuActions: h,
		tableLoading: g,
		tableRenderRowDetail: _,
		tableCellRenderers: v,
		tableRenderRowCard: y,
		tableDetailStyle: b,
		tableDetailDialogSize: x,
		tableDetailDialogTitle: S,
		operations: w,
		operationsLoading: T
	}), [
		e,
		n,
		T,
		r,
		t,
		w,
		i,
		a,
		o,
		s,
		c,
		l,
		u,
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
		S
	]);
	return /* @__PURE__ */ (0, V.jsx)(Bs.Provider, {
		value: E,
		children: C
	});
}
function Ws({ data: e, className: t }) {
	if (e === void 0) return /* @__PURE__ */ (0, V.jsx)(q, {
		className: t,
		title: "No Clicky data",
		message: "Provide either a local data payload or a URL."
	});
	let n = os(e);
	return n.ok ? /* @__PURE__ */ (0, V.jsx)("div", {
		className: t,
		children: /* @__PURE__ */ (0, V.jsx)(J, { node: n.document.node })
	}) : /* @__PURE__ */ (0, V.jsx)(Xs, {
		parsed: n,
		className: t
	});
}
function Gs({ data: e, dataFormat: t, url: n, view: r, download: a, className: o }) {
	let s = (0, B.useMemo)(() => ic({
		data: e,
		url: n,
		view: r
	}), [
		e,
		n,
		r
	]), c = (0, B.useMemo)(() => s.filter(sc), [s]), l = (0, B.useMemo)(() => s.filter(cc), [s]), u = (0, B.useMemo)(() => ms({
		url: n,
		download: a
	}), [a, n]), [d, f] = (0, B.useState)(() => s[0] ?? "clicky");
	(0, B.useEffect)(() => {
		s.includes(d) || f(s[0] ?? "clicky");
	}, [d, s]);
	let p = Cs(n, d), m = e === void 0 ? void 0 : t, h = (0, fo.useQuery)({
		queryKey: [
			"clicky",
			d,
			p
		],
		enabled: oc(d) && d !== m,
		queryFn: async () => Ys(p, d)
	}), g = d !== m && h.data?.kind === "text" ? h.data.text : void 0, _ = d === "clicky" && g !== void 0 ? g : e, v = (0, B.useMemo)(() => uc(e), [e]), y = d === "json" && g !== void 0 ? uc(g) : v, b = cc(d) ? d : null, x = dc(_), S = u.length > 0 && !x, C = `Fetching ${p}`, w = (0, B.useContext)(Bs), T = x && d === "clicky", E = (0, B.useMemo)(() => {
		if (!T || s.length <= 1) return [];
		let e = Ss(d);
		return [{
			id: "view",
			label: `View: ${ac(d)}`,
			icon: e.icon,
			...e.iconClassName ? { iconClassName: e.iconClassName } : {},
			section: "",
			children: s.map((e) => {
				let t = Ss(e);
				return {
					id: `view-${e}`,
					label: ac(e),
					description: t.description,
					icon: t.icon,
					...t.iconClassName ? { iconClassName: t.iconClassName } : {},
					disabled: e === d,
					onSelect: () => f(e)
				};
			}),
			onSelect: () => void 0
		}];
	}, [
		T,
		s,
		d
	]), D = (0, B.useMemo)(() => [...E, ...w.tableMenuActions ?? []], [E, w.tableMenuActions]);
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i("space-y-density-3", o),
		children: [!T && /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex flex-wrap items-center gap-2 rounded-md border border-border bg-muted/20 px-density-3 py-density-2",
			children: [
				c.length > 1 && /* @__PURE__ */ (0, V.jsx)("div", {
					role: "radiogroup",
					"aria-label": "Clicky view mode",
					className: "flex flex-wrap gap-1",
					children: c.map((e) => /* @__PURE__ */ (0, V.jsx)("button", {
						type: "button",
						role: "radio",
						"aria-checked": d === e,
						onClick: () => f(e),
						className: i("rounded-md border px-2.5 py-1 text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2", d === e ? "border-foreground/20 bg-accent text-foreground" : "border-transparent text-muted-foreground hover:border-foreground/20 hover:text-foreground"),
						children: ac(e)
					}, e))
				}),
				l.length > 0 && /* @__PURE__ */ (0, V.jsx)(Ks, {
					activeFormat: b,
					formats: l,
					onSelect: f
				}),
				/* @__PURE__ */ (0, V.jsxs)("div", {
					className: "ml-auto flex items-center gap-2",
					children: [oc(d) && h.isFetching && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "text-xs text-muted-foreground",
						children: "Refreshing…"
					}), S && /* @__PURE__ */ (0, V.jsx)(qs, {
						url: n,
						...a ? { download: a } : {}
					})]
				})
			]
		}), d === "pdf" ? /* @__PURE__ */ (0, V.jsx)(Zs, { src: p }) : d === "html" ? /* @__PURE__ */ (0, V.jsx)(Qs, { src: p }) : d === "excel" ? /* @__PURE__ */ (0, V.jsx)($s, {
			title: "Excel preview",
			message: "Excel output is available for download, but not inline preview.",
			href: p
		}) : d === "clicky" ? h.isPending && _ === void 0 ? /* @__PURE__ */ (0, V.jsx)(q, {
			title: "Loading Clicky",
			message: C
		}) : h.isError && _ === void 0 ? Co(h.error, `Failed to load ${p}`) : /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [h.isError && e !== void 0 && Co(h.error, `Refresh failed: ${p}`), /* @__PURE__ */ (0, V.jsx)(Bs.Provider, {
			value: {
				...w,
				tableMenuActions: D
			},
			children: /* @__PURE__ */ (0, V.jsx)(Ws, {
				data: _,
				className: "flex min-h-0 flex-1 flex-col"
			})
		})] }) : d === "json" ? h.isPending && y === void 0 ? /* @__PURE__ */ (0, V.jsx)(q, {
			title: "Loading JSON",
			message: C
		}) : h.isError && y === void 0 ? Co(h.error, `Failed to load ${p}`) : /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [h.isError && v !== void 0 && Co(h.error, `Refresh failed: ${p}`), /* @__PURE__ */ (0, V.jsx)(nc, { value: y })] }) : h.isPending ? /* @__PURE__ */ (0, V.jsx)(q, {
			title: `Loading ${ac(d)}`,
			message: C
		}) : h.isError ? Co(h.error, `Failed to load ${ac(d)} from ${p}`) : /* @__PURE__ */ (0, V.jsx)(ec, {
			format: d,
			response: h.data
		})]
	});
}
function Ks({ activeFormat: e, formats: t, onSelect: n }) {
	let r = (0, B.useRef)(null), a = (0, B.useRef)(null), [o, s] = (0, B.useState)(!1);
	return Js(o, r, a, () => s(!1)), /* @__PURE__ */ (0, V.jsxs)("div", {
		ref: r,
		className: "relative",
		children: [/* @__PURE__ */ (0, V.jsx)("button", {
			ref: a,
			type: "button",
			"aria-label": "Open additional view menu",
			"aria-haspopup": "menu",
			"aria-expanded": o,
			"aria-pressed": e != null,
			className: i("inline-flex h-[34px] w-[34px] items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2", e && "border-foreground/20 bg-accent text-foreground"),
			onClick: () => s((e) => !e),
			children: /* @__PURE__ */ (0, V.jsx)(c, {
				icon: he,
				className: "text-sm"
			})
		}), o && /* @__PURE__ */ (0, V.jsx)("div", {
			role: "menu",
			className: "absolute left-0 top-[calc(100%+0.375rem)] z-50 min-w-[18rem] rounded-md border border-border bg-popover p-1.5 text-popover-foreground shadow-lg shadow-black/5",
			children: t.map((t) => {
				let r = t === e, a = Ss(t);
				return /* @__PURE__ */ (0, V.jsxs)("button", {
					type: "button",
					role: "menuitemradio",
					"aria-checked": r,
					className: i("flex w-full items-start gap-3 rounded-md px-3 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", r && "bg-accent/70"),
					onClick: () => {
						n(t), s(!1);
					},
					children: [/* @__PURE__ */ (0, V.jsx)(c, {
						icon: a.icon,
						className: i("mt-0.5 shrink-0 text-base", a.iconClassName ?? "text-muted-foreground", r && !a.iconClassName && "text-foreground")
					}), /* @__PURE__ */ (0, V.jsxs)("span", {
						className: "min-w-0 flex-1",
						children: [/* @__PURE__ */ (0, V.jsxs)("span", {
							className: "flex items-center justify-between gap-3",
							children: [/* @__PURE__ */ (0, V.jsx)("span", {
								className: "font-medium",
								children: a.label
							}), r && /* @__PURE__ */ (0, V.jsx)(c, {
								icon: w,
								className: "text-xs text-muted-foreground"
							})]
						}), /* @__PURE__ */ (0, V.jsx)("span", {
							className: "mt-0.5 block text-xs text-muted-foreground",
							children: a.description
						})]
					})]
				}, t);
			})
		})]
	});
}
function qs({ url: e, download: t }) {
	let [n, r] = (0, B.useState)(!1), i = (0, B.useMemo)(() => _s({
		url: e,
		download: t
	}), [t, e]), a = (0, B.useMemo)(() => vs(t), [t]);
	return i.length === 0 ? null : /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsxs)("button", {
		type: "button",
		"aria-label": t?.label ? `Export ${t.label}` : "Export",
		className: "inline-flex items-center gap-2 rounded-md border border-input bg-background px-3 py-1.5 text-xs font-medium transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
		onClick: () => r(!0),
		children: [/* @__PURE__ */ (0, V.jsx)(c, {
			icon: Sn,
			className: "text-sm"
		}), /* @__PURE__ */ (0, V.jsx)("span", { children: "Export" })]
	}), /* @__PURE__ */ (0, V.jsx)(ls, {
		open: n,
		onClose: () => r(!1),
		formats: i,
		scopes: a,
		onDownload: ({ format: n, scope: r }) => bs({
			url: e,
			format: n,
			scope: r,
			...t ? { download: t } : {}
		})
	})] });
}
function Js(e, t, n, r) {
	b(e, () => {
		r(), n.current?.focus();
	}), (0, B.useEffect)(() => {
		if (!e) return;
		let n = (e) => {
			t.current?.contains(e.target) || r();
		};
		return document.addEventListener("mousedown", n), () => {
			document.removeEventListener("mousedown", n);
		};
	}, [
		r,
		e,
		t,
		n
	]);
}
function Ys(e, t) {
	return fetch(e, { headers: { Accept: Ss(t).accept } }).then(async (n) => {
		if (!n.ok) throw Me(n, "GET", e, await je(n, "json"));
		let r = n.headers.get("Content-Type") ?? "";
		return t === "excel" ? {
			kind: "blob",
			blob: await n.blob(),
			contentType: r
		} : {
			kind: "text",
			text: await n.text(),
			contentType: r
		};
	});
}
function Xs({ parsed: e, className: t }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i("rounded-md border border-destructive/30 bg-destructive/5 p-density-3", t),
		children: [/* @__PURE__ */ (0, V.jsx)("div", {
			className: "text-sm font-medium text-destructive",
			children: "Invalid Clicky payload"
		}), /* @__PURE__ */ (0, V.jsxs)("pre", {
			className: "mt-2 whitespace-pre-wrap break-all text-xs text-muted-foreground",
			children: [e.message, e.raw ? `\n\n${e.raw}` : ""]
		})]
	});
}
function q({ title: e, message: t, tone: n = "default", className: r }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i("rounded-md border p-density-3", n === "destructive" && "border-destructive/30 bg-destructive/5", n === "warning" && "border-yellow-500/30 bg-yellow-500/5", n === "default" && "border-border bg-muted/30", r),
		children: [/* @__PURE__ */ (0, V.jsx)("div", {
			className: i("text-sm font-medium", n === "destructive" && "text-destructive", n === "warning" && "text-yellow-700 dark:text-yellow-400"),
			children: e
		}), /* @__PURE__ */ (0, V.jsx)("div", {
			className: "mt-1 text-xs text-muted-foreground",
			children: t
		})]
	});
}
function Zs({ src: e }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "overflow-hidden rounded-md border border-border bg-background",
		children: [/* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex items-center justify-between border-b border-border px-3 py-2 text-xs text-muted-foreground",
			children: [/* @__PURE__ */ (0, V.jsx)("span", { children: "PDF preview" }), /* @__PURE__ */ (0, V.jsx)("a", {
				href: e,
				target: "_blank",
				rel: "noreferrer",
				className: "text-primary hover:underline",
				children: "Open in new tab"
			})]
		}), /* @__PURE__ */ (0, V.jsx)("iframe", {
			title: "Clicky PDF preview",
			src: e,
			className: "h-[720px] w-full bg-background"
		})]
	});
}
function Qs({ src: e }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "overflow-hidden rounded-md border border-border bg-background",
		children: [/* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex items-center justify-between border-b border-border px-3 py-2 text-xs text-muted-foreground",
			children: [/* @__PURE__ */ (0, V.jsx)("span", { children: "HTML preview" }), /* @__PURE__ */ (0, V.jsx)("a", {
				href: e,
				target: "_blank",
				rel: "noreferrer",
				className: "text-primary hover:underline",
				children: "Open in new tab"
			})]
		}), /* @__PURE__ */ (0, V.jsx)("iframe", {
			title: "Clicky HTML preview",
			src: e,
			className: "h-[720px] w-full bg-background",
			sandbox: "allow-same-origin"
		})]
	});
}
function $s({ title: e, message: t, href: n }) {
	return /* @__PURE__ */ (0, V.jsx)(q, {
		title: e,
		message: /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsxs)("span", { children: [t, " "] }), /* @__PURE__ */ (0, V.jsx)("a", {
			href: n,
			target: "_blank",
			rel: "noreferrer",
			className: "text-primary hover:underline",
			children: "Open format"
		})] })
	});
}
function ec({ format: e, response: t }) {
	let n = Ss(e);
	if (e === "slack") {
		let e = t?.kind === "text" ? uc(t.text) : void 0;
		return /* @__PURE__ */ (0, V.jsx)(nc, {
			value: e,
			emptyLabel: n.label
		});
	}
	return /* @__PURE__ */ (0, V.jsx)(tc, {
		title: `${n.label} output`,
		content: t?.kind === "text" ? t.text : ""
	});
}
function tc({ title: e, content: t }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "overflow-hidden rounded-md border border-border bg-background",
		children: [/* @__PURE__ */ (0, V.jsx)("div", {
			className: "border-b border-border px-3 py-2 text-xs text-muted-foreground",
			children: e
		}), /* @__PURE__ */ (0, V.jsx)("pre", {
			"aria-label": "Clicky text preview",
			className: "overflow-auto whitespace-pre-wrap break-words px-3 py-3 font-mono text-xs text-foreground",
			children: t || "No content"
		})]
	});
}
function nc({ value: e, emptyLabel: t = "JSON" }) {
	let n = (0, B.useMemo)(() => pc(e), [e]);
	return e === void 0 ? /* @__PURE__ */ (0, V.jsx)(q, {
		title: `No ${t} payload`,
		message: "The response did not include a JSON body to inspect."
	}) : /* @__PURE__ */ (0, V.jsx)("div", {
		"aria-label": "JSON tree",
		className: "overflow-hidden rounded-md border border-border bg-background",
		children: /* @__PURE__ */ (0, V.jsx)($e, {
			roots: n,
			className: "min-h-[12rem]",
			showControls: n.some((e) => (e.children?.length ?? 0) > 0),
			getKey: (e) => e.id,
			getChildren: (e) => e.children,
			renderRow: ({ node: e }) => /* @__PURE__ */ (0, V.jsx)(rc, { node: e }),
			rowClass: () => "hover:bg-accent",
			empty: /* @__PURE__ */ (0, V.jsx)("div", {
				className: "px-3 py-4 text-sm text-muted-foreground",
				children: "No JSON fields."
			}),
			toolbarClassName: "pl-3"
		})
	});
}
function rc({ node: e }) {
	let t = typeof e.value == "string" ? "text-emerald-700 dark:text-emerald-400" : typeof e.value == "number" ? "text-sky-700 dark:text-sky-400" : typeof e.value == "boolean" || e.value === null ? "text-violet-700 dark:text-violet-400" : "text-muted-foreground";
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "flex min-w-0 items-start gap-2 font-mono text-xs",
		children: [
			/* @__PURE__ */ (0, V.jsx)("span", {
				className: "shrink-0 text-foreground",
				children: e.key
			}),
			/* @__PURE__ */ (0, V.jsx)("span", {
				className: "shrink-0 text-muted-foreground",
				children: ":"
			}),
			/* @__PURE__ */ (0, V.jsx)("span", {
				className: i("min-w-0 break-words", t),
				children: e.preview
			})
		]
	});
}
function ic({ data: e, url: t, view: n }) {
	let r = lc(n, "clicky", e, t), i = lc(n, "json", e, t), a = [];
	r && (t || e !== void 0) && a.push("clicky"), i && (t || e !== void 0) && a.push("json");
	for (let r of fs) lc(n, r, e, t) && t && a.push(r);
	return a.length === 0 && (t || e !== void 0) && a.push("clicky"), a;
}
function ac(e) {
	return Ss(e).label;
}
function oc(e) {
	return e !== "pdf" && e !== "html" && e !== "excel";
}
function sc(e) {
	return ds.includes(e);
}
function cc(e) {
	return fs.includes(e);
}
function lc(e, t, n, r) {
	let i = t === "clicky" || t === "json" ? r != null || n !== void 0 : !!r;
	return e === void 0 ? i : Array.isArray(e) ? e.length === 0 ? !1 : t === "clicky" ? e.includes("clicky") || e.includes("json") : e.includes(t) : t in e ? !!e[t] : t === "clicky" && e.json === !0 || i;
}
function uc(e) {
	if (e === void 0) return;
	if (typeof e != "string") return e;
	let t = e.trim();
	if (!t) return "";
	try {
		return JSON.parse(t);
	} catch {
		return e;
	}
}
function dc(e) {
	if (e === void 0) return !1;
	let t = os(e);
	return t.ok ? fc(t.document.node) : !1;
}
function fc(e) {
	return e ? !!(e.kind === "table" || e.children?.some(fc) || e.items?.some(fc) || e.fields?.some((e) => fc(e.value)) || e.label && fc(e.label) || e.content && fc(e.content) || e.tooltip && fc(e.tooltip) || e.columns?.some((e) => e.header && fc(e.header)) || e.rows?.some((e) => Object.values(e.cells).some((e) => fc(e))) || e.rows?.some((e) => e.detail && fc(e.detail))) : !1;
}
function pc(e) {
	return Array.isArray(e) ? e.map((e, t) => mc(String(t), e, `$[${t}]`)) : _c(e) ? Object.entries(e).map(([e, t]) => mc(e, t, `$.${e}`)) : e === void 0 ? [] : [mc("$", e, "$")];
}
function mc(e, t, n) {
	let r = hc(t, n);
	return {
		id: n,
		key: e,
		value: t,
		preview: gc(t),
		...r ? { children: r } : {}
	};
}
function hc(e, t) {
	if (Array.isArray(e)) return e.map((e, n) => mc(String(n), e, `${t}[${n}]`));
	if (_c(e)) return Object.entries(e).map(([e, n]) => mc(e, n, `${t}.${e}`));
}
function gc(e) {
	if (Array.isArray(e)) return `[${e.length} item${e.length === 1 ? "" : "s"}]`;
	if (_c(e)) {
		let t = Object.keys(e).length;
		return `{${t} key${t === 1 ? "" : "s"}}`;
	}
	return typeof e == "string" ? JSON.stringify(e) : e === void 0 ? "undefined" : String(e);
}
function _c(e) {
	return typeof e == "object" && !!e && !Array.isArray(e);
}
var vc = { "text-muted": "text-muted-foreground" };
function yc(e) {
	return e.split(/\s+/).map((e) => vc[e] ?? e).join(" ");
}
function bc(e) {
	let t = e.style?.className;
	if (!t) return e;
	let n = yc(t);
	return n === t ? e : {
		...e,
		style: {
			...e.style,
			className: n
		}
	};
}
function J({ node: e }) {
	if (!e) return null;
	let t = bc(e);
	switch (t.kind) {
		case "text": return /* @__PURE__ */ (0, V.jsx)(Sc, { node: t });
		case "link": return /* @__PURE__ */ (0, V.jsx)(wc, { node: t });
		case "link-command": return /* @__PURE__ */ (0, V.jsx)(Tc, { node: t });
		case "icon": return /* @__PURE__ */ (0, V.jsx)(Ac, { node: t });
		case "list": return /* @__PURE__ */ (0, V.jsx)(Kc, { node: t });
		case "map": return /* @__PURE__ */ (0, V.jsx)(qc, { node: t });
		case "table": return /* @__PURE__ */ (0, V.jsx)(Yc, { node: t });
		case "tree": return /* @__PURE__ */ (0, V.jsx)(wl, { node: t });
		case "object-graph": return /* @__PURE__ */ (0, V.jsx)(Lo, { roots: t.objects ?? [] });
		case "execution-tree": return /* @__PURE__ */ (0, V.jsx)(Bo, { roots: t.executionRoots ?? [] });
		case "code": return /* @__PURE__ */ (0, V.jsx)(Ve, {
			language: t.language,
			source: t.source ?? t.plain ?? "",
			highlightedHtml: t.highlightedHtml
		});
		case "heading": return /* @__PURE__ */ (0, V.jsx)(jc, { node: t });
		case "blockquote": return /* @__PURE__ */ (0, V.jsx)(Pc, { node: t });
		case "admonition": return /* @__PURE__ */ (0, V.jsx)(Fc, { node: t });
		case "footnote-ref": return /* @__PURE__ */ (0, V.jsx)(zc, { node: t });
		case "footnote": return /* @__PURE__ */ (0, V.jsx)(Bc, { node: t });
		case "footnotes": return /* @__PURE__ */ (0, V.jsx)(Vc, { node: t });
		case "collapsed": return /* @__PURE__ */ (0, V.jsx)(Ol, { node: t });
		case "stacktrace": return /* @__PURE__ */ (0, V.jsx)(Dl, { node: t });
		case "button": return /* @__PURE__ */ (0, V.jsx)(kl, { node: t });
		case "button-group": return /* @__PURE__ */ (0, V.jsx)(Al, { node: t });
		case "html": return /* @__PURE__ */ (0, V.jsx)(jl, { node: t });
		case "comment": return /* @__PURE__ */ (0, V.jsx)(xc, { node: t });
		case "badge": return /* @__PURE__ */ (0, V.jsx)(El, { node: t });
		default: return /* @__PURE__ */ (0, V.jsx)("pre", {
			className: "rounded-md border border-border bg-muted p-density-3 text-xs",
			children: JSON.stringify(t, null, 2)
		});
	}
}
function xc({ node: e }) {
	let t = e.text ?? e.plain, n = /* @__PURE__ */ (0, V.jsx)(Cc, { node: e }), r = e.style?.maxWidth && e.style.maxWidth > 0 ? Math.min(e.style.maxWidth, 50) : 50, a = e.style?.maxLines && e.style.maxLines > 0 ? Math.min(e.style.maxLines, 3) : 3, o = Ml({
		...e.style,
		maxWidth: r,
		maxLines: a
	}, t);
	return !t && !e.children?.length ? null : /* @__PURE__ */ (0, V.jsx)("div", {
		className: i("mt-1 text-xs leading-snug text-muted-foreground", e.style?.className),
		style: o,
		title: e.tooltip?.plain ?? t,
		children: n
	});
}
function Sc({ node: e }) {
	let t = Ml(e.style, e.text ?? e.plain), n = /* @__PURE__ */ (0, V.jsx)(Cc, { node: e });
	return !e.style && !e.tooltip ? /* @__PURE__ */ (0, V.jsx)(V.Fragment, { children: n }) : /* @__PURE__ */ (0, V.jsx)("span", {
		style: t,
		title: e.tooltip?.plain,
		className: i(e.style?.className, (e.text ?? "").includes("\n") && "whitespace-pre-wrap"),
		children: n
	});
}
function Cc({ node: e }) {
	return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [e.text, e.children?.map((e, t) => /* @__PURE__ */ (0, V.jsx)(B.Fragment, { children: /* @__PURE__ */ (0, V.jsx)(J, { node: e }) }, t))] });
}
function wc({ node: e }) {
	let t = /* @__PURE__ */ (0, V.jsx)(Cc, { node: e }), n = Ml(e.style, e.text ?? e.plain), r = Ul(e.target), a = r === "_blank" ? "noopener noreferrer" : void 0;
	return e.href ? /* @__PURE__ */ (0, V.jsx)("a", {
		href: e.href,
		target: r,
		rel: a,
		title: e.tooltip?.plain,
		style: n,
		className: i(e.style?.className, "inline cursor-pointer"),
		children: t
	}) : /* @__PURE__ */ (0, V.jsx)("span", {
		style: n,
		children: t
	});
}
function Tc({ node: e }) {
	let t = (0, B.useContext)(Bs), n = (0, B.useMemo)(() => Pl(e), [e]), r = (0, B.useMemo)(() => Fl(n, t.commandRuntime, t.operations), [
		n,
		t.commandRuntime,
		t.operations
	]), a = n.target ?? "Dialog", o = t.commandRuntime != null, [s, c] = (0, B.useState)(!1), [l, u] = (0, B.useState)(!1);
	if (!o) return /* @__PURE__ */ (0, V.jsx)("span", {
		title: e.tooltip?.plain,
		className: i(e.style?.className),
		children: /* @__PURE__ */ (0, V.jsx)(Cc, { node: e })
	});
	if (a === "_self" || a === "_window" || a === "_tab") {
		let n = t.commandRuntime?.hrefForCommand?.(r) ?? Vl(r);
		return n ? /* @__PURE__ */ (0, V.jsx)("a", {
			href: n,
			target: Ul(a),
			rel: a === "_self" ? void 0 : "noopener noreferrer",
			title: e.tooltip?.plain,
			style: Ml(e.style, e.text ?? e.plain),
			className: i(e.style?.className, "inline cursor-pointer"),
			children: /* @__PURE__ */ (0, V.jsx)(Cc, { node: e })
		}) : /* @__PURE__ */ (0, V.jsx)("span", {
			title: e.tooltip?.plain,
			className: i(e.style?.className),
			children: /* @__PURE__ */ (0, V.jsx)(Cc, { node: e })
		});
	}
	return a === "_clicky" ? /* @__PURE__ */ (0, V.jsx)(Ec, {
		node: e,
		onClick: () => t.commandRuntime?.onNavigate?.(r)
	}) : a === "Hover" ? /* @__PURE__ */ (0, V.jsx)(oe, {
		delay: 75,
		cardClassName: "w-[28rem] max-w-[calc(100vw-2rem)] whitespace-normal p-0",
		trigger: /* @__PURE__ */ (0, V.jsx)(Ec, { node: e }),
		children: /* @__PURE__ */ (0, V.jsx)("div", {
			className: "max-h-[26rem] overflow-auto p-density-3",
			children: /* @__PURE__ */ (0, V.jsx)(Oc, { resolved: r })
		})
	}) : a === "Expand" ? /* @__PURE__ */ (0, V.jsxs)("span", {
		className: "inline",
		children: [/* @__PURE__ */ (0, V.jsx)(Ec, {
			node: e,
			ariaExpanded: l,
			onClick: () => u((e) => !e)
		}), l && /* @__PURE__ */ (0, V.jsx)("div", {
			className: "mt-2 rounded-md border border-border bg-background p-density-3",
			children: /* @__PURE__ */ (0, V.jsx)(Oc, { resolved: r })
		})]
	}) : /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsx)(Ec, {
		node: e,
		ariaExpanded: s,
		onClick: () => c(!0)
	}), /* @__PURE__ */ (0, V.jsx)(Dc, {
		open: s,
		onClose: () => c(!1),
		resolved: r
	})] });
}
function Ec({ node: e, onClick: t, ariaExpanded: n }) {
	let r = Ml(e.style, e.text ?? e.plain);
	return /* @__PURE__ */ (0, V.jsx)("button", {
		type: "button",
		title: e.tooltip?.plain,
		"aria-expanded": n,
		onClick: (e) => {
			e.preventDefault(), t?.();
		},
		style: r,
		className: i("inline cursor-pointer appearance-none border-0 bg-transparent p-0 text-left align-baseline text-inherit", e.style?.className),
		children: /* @__PURE__ */ (0, V.jsx)(Cc, { node: e })
	});
}
function Dc({ open: e, onClose: t, resolved: n }) {
	let r = (0, B.useContext)(Bs), i = n.operation, a = i?.operation.parameters ?? [], o = (0, B.useMemo)(() => Rl(a, n.request), [a, n.request]), [s, c] = (0, B.useState)(null), [l, u] = (0, B.useState)(""), [d, f] = (0, B.useState)(!1), [p, m] = (0, B.useState)(!1), h = e && n.request.autoRun === !0 && i != null && zl(a, o).length === 0;
	(0, B.useEffect)(() => {
		e || (m(!1), c(null), u(""));
	}, [e, n.request.command]);
	async function g(e) {
		if (r.commandRuntime && i) {
			f(!0), u("");
			try {
				let t = await Bl(r.commandRuntime.client, i, e);
				c(t);
			} catch (e) {
				c(null), u(e instanceof Error ? e.message : String(e ?? "Unknown error"));
			} finally {
				f(!1);
			}
		}
	}
	return (0, B.useEffect)(() => {
		h && !p && (m(!0), g(o));
	}, [
		p,
		o,
		h
	]), /* @__PURE__ */ (0, V.jsx)(A, {
		open: e,
		onClose: t,
		title: i?.operation.summary || n.request.command,
		size: "xl",
		children: /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "space-y-4",
			children: [
				i ? /* @__PURE__ */ (0, V.jsxs)("div", {
					className: "flex flex-wrap items-center gap-3 text-xs text-muted-foreground",
					children: [/* @__PURE__ */ (0, V.jsx)("span", {
						className: "rounded-md border border-border bg-muted px-2 py-1 font-medium uppercase",
						children: i.method
					}), /* @__PURE__ */ (0, V.jsx)("code", { children: i.path })]
				}) : r.operationsLoading ? /* @__PURE__ */ (0, V.jsx)(q, {
					title: "Loading commands",
					message: "Resolving the command definition."
				}) : /* @__PURE__ */ (0, V.jsx)(q, {
					title: "Unknown command",
					message: `No operation matched "${n.request.command}".`,
					tone: "destructive"
				}),
				i && /* @__PURE__ */ (0, V.jsx)("div", {
					className: "rounded-md border border-border bg-muted/20 p-density-3",
					children: /* @__PURE__ */ (0, V.jsx)(So, {
						client: r.commandRuntime.client,
						path: i.path,
						method: i.method,
						parameters: a,
						initialValues: o,
						isSubmitting: d,
						submitLabel: "Run command",
						submittingLabel: "Running…",
						onSubmit: g
					})
				}),
				l ? /* @__PURE__ */ (0, V.jsx)(q, {
					title: "Command failed",
					message: l,
					tone: "destructive"
				}) : /* @__PURE__ */ (0, V.jsx)(kc, {
					response: s,
					pending: d,
					emptyMessage: "Run the command to load a Clicky response."
				})
			]
		})
	});
}
function Oc({ resolved: e }) {
	let t = (0, B.useContext)(Bs), n = e.operation, r = n?.operation.parameters ?? [], i = (0, B.useMemo)(() => Rl(r, e.request), [r, e.request]), a = (0, B.useMemo)(() => zl(r, i), [i, r]), o = (0, fo.useQuery)({
		queryKey: [
			"clicky-command",
			e.request.command,
			n?.method,
			n?.path,
			i
		],
		enabled: t.commandRuntime != null && n != null && a.length === 0,
		retry: 0,
		queryFn: async () => Bl(t.commandRuntime.client, n, i)
	});
	return t.commandRuntime ? n ? a.length > 0 ? /* @__PURE__ */ (0, V.jsx)(q, {
		title: "Missing required parameters",
		message: `This link is missing: ${a.map((e) => Nl(e.name)).join(", ")}.`,
		tone: "warning"
	}) : o.isPending ? /* @__PURE__ */ (0, V.jsx)(q, {
		title: "Running command",
		message: "Loading Clicky response…"
	}) : o.isError ? /* @__PURE__ */ (0, V.jsx)(q, {
		title: "Command failed",
		message: o.error instanceof Error ? o.error.message : "Request failed",
		tone: "destructive"
	}) : /* @__PURE__ */ (0, V.jsx)(kc, {
		response: o.data,
		emptyMessage: "No response body returned."
	}) : t.operationsLoading ? /* @__PURE__ */ (0, V.jsx)(q, {
		title: "Loading commands",
		message: "Resolving the command definition."
	}) : /* @__PURE__ */ (0, V.jsx)(q, {
		title: "Unknown command",
		message: `No operation matched "${e.request.command}".`,
		tone: "destructive"
	}) : /* @__PURE__ */ (0, V.jsx)(q, {
		title: "Command runtime unavailable",
		message: "This Clicky view was rendered without a command runtime.",
		tone: "warning"
	});
}
function kc({ response: e, pending: t = !1, emptyMessage: n }) {
	if (t) return /* @__PURE__ */ (0, V.jsx)(q, {
		title: "Running command",
		message: "Loading Clicky response…"
	});
	if (!e) return /* @__PURE__ */ (0, V.jsx)(q, {
		title: "No response",
		message: n
	});
	let r = Wl(e) ? e.parsed ?? No(e) : e, i = Wl(e) ? e.stdout || e.output || e.message || "" : typeof e == "string" ? e : JSON.stringify(e, null, 2), a = typeof r == "string" || typeof r == "object" && r ? r : i, o = a === "" ? null : os(a);
	return o?.ok ? /* @__PURE__ */ (0, V.jsx)("div", {
		className: "rounded-md border border-border bg-background p-density-3",
		children: /* @__PURE__ */ (0, V.jsx)(J, { node: o.document.node })
	}) : typeof r == "object" && r ? /* @__PURE__ */ (0, V.jsx)(nc, { value: r }) : /* @__PURE__ */ (0, V.jsx)(tc, {
		title: "Response body",
		content: i
	});
}
function Ac({ node: e }) {
	let t = Ml(e.style, e.plain ?? e.unicode), n = e.style?.className, r = e.iconify ? Go[e.iconify] : void 0;
	return /* @__PURE__ */ (0, V.jsx)("span", {
		style: t,
		title: e.tooltip?.plain,
		className: "inline-flex items-center",
		"data-clicky-icon": e.iconify ?? "",
		children: r ? /* @__PURE__ */ (0, V.jsx)(c, {
			icon: r,
			...n ? { className: n } : {}
		}) : e.iconify && u() ? /* @__PURE__ */ (0, V.jsx)(c, {
			name: e.iconify,
			...n ? { className: n } : {}
		}) : /* @__PURE__ */ (0, V.jsx)("span", {
			className: n,
			children: e.unicode ?? e.plain
		})
	});
}
function jc({ node: e }) {
	let t = Mc(e.level), n = e.content ?? Gc(e), r = Ml(e.style, Y(n)), a = i(Nc(t), "text-foreground", e.style?.className);
	switch (t) {
		case 1: return /* @__PURE__ */ (0, V.jsx)("h1", {
			className: a,
			style: r,
			children: n && /* @__PURE__ */ (0, V.jsx)(J, { node: n })
		});
		case 2: return /* @__PURE__ */ (0, V.jsx)("h2", {
			className: a,
			style: r,
			children: n && /* @__PURE__ */ (0, V.jsx)(J, { node: n })
		});
		case 3: return /* @__PURE__ */ (0, V.jsx)("h3", {
			className: a,
			style: r,
			children: n && /* @__PURE__ */ (0, V.jsx)(J, { node: n })
		});
		case 4: return /* @__PURE__ */ (0, V.jsx)("h4", {
			className: a,
			style: r,
			children: n && /* @__PURE__ */ (0, V.jsx)(J, { node: n })
		});
		case 5: return /* @__PURE__ */ (0, V.jsx)("h5", {
			className: a,
			style: r,
			children: n && /* @__PURE__ */ (0, V.jsx)(J, { node: n })
		});
		default: return /* @__PURE__ */ (0, V.jsx)("h6", {
			className: a,
			style: r,
			children: n && /* @__PURE__ */ (0, V.jsx)(J, { node: n })
		});
	}
}
function Mc(e) {
	return Number.isFinite(e) ? Math.min(6, Math.max(1, Math.trunc(e ?? 1))) : 1;
}
function Nc(e) {
	switch (e) {
		case 1: return "text-2xl font-semibold leading-tight";
		case 2: return "text-xl font-semibold leading-tight";
		case 3: return "text-lg font-semibold leading-snug";
		case 4: return "text-base font-semibold leading-snug";
		case 5: return "text-sm font-semibold leading-snug";
		default: return "text-xs font-semibold uppercase leading-snug text-muted-foreground";
	}
}
function Pc({ node: e }) {
	let t = e.content ?? Gc(e);
	return t ? /* @__PURE__ */ (0, V.jsx)("blockquote", {
		className: "border-l-4 border-border pl-density-3 text-sm leading-relaxed text-muted-foreground",
		children: /* @__PURE__ */ (0, V.jsx)(J, { node: t })
	}) : null;
}
function Fc({ node: e }) {
	let t = Ic(e.severity), n = e.label ?? {
		kind: "text",
		text: Nl(t),
		plain: Nl(t)
	}, r = Rc(t);
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		role: t === "danger" || t === "warning" ? "alert" : "note",
		"data-severity": t,
		className: i("rounded-md border px-density-3 py-density-2", r.container),
		children: [/* @__PURE__ */ (0, V.jsx)("div", {
			className: i("text-sm font-semibold", r.title),
			children: /* @__PURE__ */ (0, V.jsx)(J, { node: n })
		}), e.content && /* @__PURE__ */ (0, V.jsx)("div", {
			className: i("mt-1 text-sm leading-relaxed", r.body),
			children: /* @__PURE__ */ (0, V.jsx)(J, { node: e.content })
		})]
	});
}
function Ic(e) {
	let t = (e ?? "note").trim().toLowerCase();
	return t === "info" || t === "tip" || t === "warning" || t === "danger" || t === "note" ? t : "note";
}
var Lc = {
	info: "note",
	tip: "tip",
	warning: "warning",
	danger: "caution",
	note: "default"
};
function Rc(e) {
	let t = Ut(Lc[e] ?? "default");
	return {
		container: t.subtle,
		title: t.title,
		body: t.content
	};
}
function zc({ node: e }) {
	let t = Uc(e.id);
	return t ? /* @__PURE__ */ (0, V.jsx)("sup", {
		id: Wc("fnref", t),
		className: "align-super text-[0.75em]",
		children: /* @__PURE__ */ (0, V.jsx)("a", {
			href: `#${Wc("fn", t)}`,
			"aria-label": `Footnote ${t}`,
			className: "text-primary hover:underline",
			children: `[^${t}]`
		})
	}) : null;
}
function Bc({ node: e }) {
	let t = Uc(e.id);
	return t ? /* @__PURE__ */ (0, V.jsxs)("div", {
		id: Wc("fn", t),
		className: "text-sm text-muted-foreground",
		children: [
			/* @__PURE__ */ (0, V.jsx)("span", {
				className: "mr-2 font-mono text-xs text-foreground",
				children: `[^${t}]`
			}),
			e.content && /* @__PURE__ */ (0, V.jsx)(J, { node: e.content }),
			/* @__PURE__ */ (0, V.jsx)(Hc, { id: t })
		]
	}) : null;
}
function Vc({ node: e }) {
	let t = (e.items ?? []).filter((e) => Uc(e.id));
	return t.length === 0 ? null : /* @__PURE__ */ (0, V.jsx)("section", {
		"aria-label": "Footnotes",
		className: "border-t border-border pt-density-2 text-xs text-muted-foreground",
		children: /* @__PURE__ */ (0, V.jsx)("ol", {
			className: "ml-5 list-decimal space-y-1",
			children: t.map((e, t) => /* @__PURE__ */ (0, V.jsxs)("li", {
				id: Wc("fn", e.id ?? ""),
				children: [e.content && /* @__PURE__ */ (0, V.jsx)(J, { node: e.content }), /* @__PURE__ */ (0, V.jsx)(Hc, { id: e.id ?? "" })]
			}, `${e.id}-${t}`))
		})
	});
}
function Hc({ id: e }) {
	return /* @__PURE__ */ (0, V.jsx)("a", {
		href: `#${Wc("fnref", e)}`,
		"aria-label": `Back to footnote reference ${e}`,
		className: "ml-2 text-primary hover:underline",
		children: "back"
	});
}
function Uc(e) {
	return (e ?? "").trim();
}
function Wc(e, t) {
	return `${e}-${encodeURIComponent(t)}`;
}
function Gc(e) {
	let t = e.text ?? e.plain;
	return t ? {
		kind: "text",
		text: t,
		plain: t
	} : void 0;
}
function Kc({ node: e }) {
	let t = e.items ?? [];
	if (t.length === 0) return null;
	if (e.inline) return /* @__PURE__ */ (0, V.jsx)("span", {
		className: "inline-flex flex-wrap items-start gap-1",
		children: t.map((t, n) => /* @__PURE__ */ (0, V.jsxs)(B.Fragment, { children: [n > 0 && /* @__PURE__ */ (0, V.jsx)("span", {
			className: "text-muted-foreground",
			children: ","
		}), /* @__PURE__ */ (0, V.jsxs)("span", {
			className: "inline-flex items-center gap-1",
			children: [
				!e.ordered && e.bullet && /* @__PURE__ */ (0, V.jsx)(J, { node: e.bullet }),
				e.ordered && /* @__PURE__ */ (0, V.jsxs)("span", {
					className: "text-muted-foreground",
					children: [n + 1, "."]
				}),
				/* @__PURE__ */ (0, V.jsx)(J, { node: t })
			]
		})] }, n))
	});
	if (e.unstyled) return /* @__PURE__ */ (0, V.jsx)("div", {
		className: "space-y-1",
		children: t.map((t, n) => /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex items-start gap-2",
			children: [
				!e.ordered && e.bullet && /* @__PURE__ */ (0, V.jsx)(J, { node: e.bullet }),
				e.ordered && /* @__PURE__ */ (0, V.jsxs)("span", {
					className: "text-muted-foreground",
					children: [n + 1, "."]
				}),
				/* @__PURE__ */ (0, V.jsx)("div", {
					className: "min-w-0 flex-1",
					children: /* @__PURE__ */ (0, V.jsx)(J, { node: t })
				})
			]
		}, n))
	});
	let n = e.ordered ? "ol" : "ul", r = e.ordered ? "list-decimal" : "list-disc";
	return /* @__PURE__ */ (0, V.jsx)(n, {
		className: i("ml-5 space-y-1 text-sm", r),
		children: t.map((e, t) => /* @__PURE__ */ (0, V.jsx)("li", { children: /* @__PURE__ */ (0, V.jsx)(J, { node: e }) }, t))
	});
}
function qc({ node: e }) {
	let t = (0, B.useContext)(Bs), n = (e.fields ?? []).filter((e) => !(t.tablePagination && Jc(e)));
	if (n.length === 0) return null;
	let r = n.filter((e) => Yl(e.value)), i = n.filter((e) => !Yl(e.value));
	return r.length === 0 && i.length === 1 && i[0]?.value.kind === "table" ? /* @__PURE__ */ (0, V.jsx)(J, { node: i[0].value }) : /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "space-y-density-4",
		children: [r.length > 0 && /* @__PURE__ */ (0, V.jsx)("dl", {
			className: "grid gap-density-3 md:grid-cols-2",
			children: r.map((e) => /* @__PURE__ */ (0, V.jsxs)("div", { children: [/* @__PURE__ */ (0, V.jsx)("dt", {
				className: "text-xs font-semibold uppercase tracking-wide text-muted-foreground",
				children: e.label || Nl(e.name)
			}), /* @__PURE__ */ (0, V.jsx)("dd", {
				className: "mt-1 text-sm text-foreground",
				children: /* @__PURE__ */ (0, V.jsx)(J, { node: e.value })
			})] }, e.name))
		}), i.map((e) => /* @__PURE__ */ (0, V.jsxs)("section", {
			className: "space-y-2",
			children: [/* @__PURE__ */ (0, V.jsx)("div", {
				className: "text-xs font-semibold uppercase tracking-wide text-muted-foreground",
				children: e.label || Nl(e.name)
			}), /* @__PURE__ */ (0, V.jsx)("div", {
				className: "rounded-md border border-border bg-background p-density-3",
				children: /* @__PURE__ */ (0, V.jsx)(J, { node: e.value })
			})]
		}, e.name))]
	});
}
function Jc(e) {
	if ((e.label || e.name || "").toLowerCase() !== "page" || e.value.kind !== "map") return !1;
	let t = new Set((e.value.fields ?? []).map((e) => e.name.toLowerCase()));
	return t.has("limit") && t.has("offset") && t.has("total");
}
function Yc({ node: e }) {
	return /* @__PURE__ */ (0, V.jsx)(Xc, {
		columns: e.columns ?? [],
		rows: e.rows ?? [],
		...e.autoFilter === void 0 ? {} : { autoFilter: e.autoFilter }
	});
}
function Xc({ columns: e, rows: t, autoFilter: n, onTableRowClick: r, getTableRowHref: i, isTableRowClickable: a, search: o, timeRange: s, externalFilters: c, cellFilters: l, onCellFilterChange: u, pagination: d, sort: f, onSortChange: p, infinite: m, rowSelection: h, hiddenColumns: g, menuActions: _, loading: v, renderRowDetail: y, cellRenderers: b, renderRowCard: x, detailStyle: S, detailDialogSize: C, detailDialogTitle: w }) {
	let T = (0, B.useContext)(Bs), E = r ?? T.onTableRowClick, D = i ?? T.getTableRowHref, O = a ?? T.isTableRowClickable, ee = o ?? T.tableSearch, te = s ?? T.tableTimeRange, ne = c ?? T.tableExternalFilters, re = l ?? T.tableCellFilters, k = u ?? T.onTableCellFilterChange, A = d ?? T.tablePagination, ie = f === void 0 ? T.tableSort : f, ae = p ?? T.onTableSortChange, oe = m ?? T.tableInfinite, j = h ?? T.tableRowSelection, se = g ?? T.tableHiddenColumns, ce = _ ?? T.tableMenuActions, le = v ?? T.tableLoading, ue = y ?? T.tableRenderRowDetail, de = b ?? T.tableCellRenderers, M = x ?? T.tableRenderRowCard, fe = S ?? T.tableDetailStyle, pe = C ?? T.tableDetailDialogSize, me = w ?? T.tableDetailDialogTitle, he = new Set(se), N = e.filter((e) => !he.has(e.name));
	if (N.length === 0) return /* @__PURE__ */ (0, V.jsx)("div", {
		className: "text-sm text-muted-foreground",
		children: "No data"
	});
	if ($c(N, t)) return /* @__PURE__ */ (0, V.jsx)(Zc, {
		columns: N,
		rows: t
	});
	let P = new Map(e.map((e) => [e.name, e])), ge = N.map((e) => {
		let t = el(e), n = tl(e), r = e.type === "json", i = {
			key: `cells.${e.name}`,
			accessor: (t) => t.cells[e.name],
			label: e.header ? /* @__PURE__ */ (0, V.jsx)(J, { node: e.header }) : e.label || Nl(e.name),
			...e.align ? { align: e.align } : {},
			...ae ? { sortable: !!e.sortKey } : e.sortable === void 0 ? {} : { sortable: e.sortable },
			...e.filterable === void 0 ? {} : { filterable: e.filterable },
			...e.filterKey ? { filterKey: e.filterKey } : {},
			serverFilterValue: (e) => {
				let t = e;
				return t.filterValue === void 0 ? Y(t) : t.filterValue;
			},
			serverFilterLabel: (e) => Y(e),
			...e.grow === void 0 ? {} : { grow: e.grow },
			...e.shrink === void 0 ? {} : { shrink: e.shrink },
			...e.minWidth === void 0 ? {} : { minWidth: e.minWidth },
			...e.maxWidth === void 0 ? {} : { maxWidth: e.maxWidth }
		};
		return e.kind === "timestamp" && !t && !n && !r ? {
			...i,
			kind: "timestamp",
			accessor: (t) => ql(t.cells[e.name], e),
			serverFilterValue: (e) => typeof e == "string" || typeof e == "number" ? String(e) : void 0,
			serverFilterLabel: (t, n) => typeof t == "string" || typeof t == "number" ? String(t) : Y(n.cells[e.name])
		} : n ? {
			...i,
			kind: "tags",
			render: (e) => /* @__PURE__ */ (0, V.jsx)(nl, { node: e }),
			sortValue: (e) => ll(e),
			filterValue: (e) => cl(e)
		} : r ? {
			...i,
			render: (e) => /* @__PURE__ */ (0, V.jsx)(rl, { node: e }),
			sortValue: (e) => fl(e),
			filterValue: (e) => fl(e)
		} : t ? {
			...i,
			kind: "tags",
			render: (t) => {
				let n = pl(t, e);
				return /* @__PURE__ */ (0, V.jsx)(ei, {
					tags: n,
					maxVisible: 3
				});
			},
			sortValue: (t) => pl(t, e).length,
			filterValue: (t) => pl(t, e).map((e) => e.token)
		} : {
			...i,
			...e.kind ? { kind: e.kind } : {},
			render: (e) => {
				let t = e;
				return t.kind === "code" ? /* @__PURE__ */ (0, V.jsx)(Qc, { node: t }) : /* @__PURE__ */ (0, V.jsx)(J, { node: t });
			},
			sortValue: (e) => Gl(e),
			filterValue: (e) => Y(e)
		};
	}).map((e, t) => {
		let n = de?.[N[t].name];
		return n ? {
			...e,
			render: (e, r) => n(ql(e, N[t]), Jl(r, P))
		} : e;
	}), _e = A || ae ? void 0 : N.find((e) => e.sortable !== !1) ?? N[0], ve = ie ? (() => {
		let e = N.find((e) => e.sortKey === ie.key);
		if (!e) throw Error(`Server sort key ${ie.key} does not match a Clicky column`);
		return {
			key: `cells.${e.name}`,
			dir: ie.dir
		};
	})() : ie === null ? null : void 0, ye = ae ? (e) => {
		if (!e) {
			ae(null);
			return;
		}
		let t = e.key.startsWith("cells.") ? e.key.slice(6) : e.key, n = N.find((e) => e.name === t);
		if (!n?.sortKey) throw Error(`Clicky column ${t} does not declare a server sort key`);
		ae({
			key: n.sortKey,
			dir: e.dir
		});
	} : void 0;
	return /* @__PURE__ */ (0, V.jsx)(Ga, {
		className: "min-h-0 flex-1",
		data: t,
		columns: ge,
		defaultHiddenColumns: N.filter((e) => e.defaultHidden).map((e) => `cells.${e.name}`),
		...n === void 0 ? {} : { autoFilter: n },
		..._e ? { defaultSort: {
			key: `cells.${_e.name}`,
			dir: "asc"
		} } : {},
		...ve === void 0 ? {} : { sort: ve },
		...ye ? {
			onSortChange: ye,
			manualSort: !0
		} : {},
		getRowId: (e, t) => j?.getRowId(e, t) ?? (M && typeof Jl(e, P).id == "string" ? String(Jl(e, P).id) : void 0) ?? `${t}-${N.map((t) => Y(e.cells[t.name])).filter(Boolean).join("|")}`,
		...j ? { rowSelection: j } : {},
		...E ? { onRowClick: E } : {},
		...D ? { getRowHref: D } : {},
		...E && (O || D) ? { isRowClickable: O ?? ((e) => !!D?.(e)) } : {},
		...ee ? { externalSearch: ee } : {},
		...te ? { externalTimeRange: te } : {},
		...ne ? { externalFilters: ne } : {},
		...re ? { cellFilters: re } : {},
		...k ? { onCellFilterChange: k } : {},
		...A ? { pagination: A } : {},
		...oe ? { infinite: oe } : {},
		...ce ? { menuActions: ce } : {},
		...le === void 0 ? {} : { loading: le },
		...ue ? { renderExpandedRow: (e) => ue(Jl(e, P)) } : {},
		...M ? { renderCard: (e) => M(Jl(e, P)) } : {},
		...fe ? { detailStyle: fe } : {},
		...pe ? { detailDialogSize: pe } : {},
		...me ? { detailDialogTitle: (e) => me(Jl(e, P)) } : {}
	});
}
function Zc({ columns: e, rows: t }) {
	return /* @__PURE__ */ (0, V.jsx)("div", {
		className: "space-y-2",
		children: t.map((t, n) => {
			let r = _l(t, e, n), i = gl(t, e);
			return /* @__PURE__ */ (0, V.jsxs)("details", {
				className: "rounded-md border border-border bg-background",
				children: [/* @__PURE__ */ (0, V.jsx)("summary", {
					className: "cursor-pointer px-3 py-2 text-sm font-medium text-foreground",
					children: r
				}), /* @__PURE__ */ (0, V.jsx)("div", {
					className: "border-t border-border p-3",
					children: /* @__PURE__ */ (0, V.jsx)(qc, { node: i })
				})]
			}, vl(t, e, n));
		})
	});
}
function Qc({ node: e }) {
	return /* @__PURE__ */ (0, V.jsx)(Ve, {
		language: e.language,
		source: e.source ?? e.plain ?? "",
		highlightedHtml: e.highlightedHtml,
		inline: !0
	});
}
function $c(e, t) {
	return t.length === 0 || e.some((e) => tl(e) || e.type === "json") ? !1 : t.every((t) => {
		let n = e.map((e) => t.cells[e.name]).filter(Boolean);
		if (n.length === 0) return !1;
		let r = n.filter(hl);
		return r.length > 0 && r.length === n.length;
	});
}
function el(e) {
	if (e.kind === "tags") return !0;
	let t = e.name.toLowerCase();
	return /(^|[._-])(tags?|labels?|annotations?)$/.test(t);
}
function tl(e) {
	return e.type === "key_value" || e.type === "key_values";
}
function nl({ node: e }) {
	let t = il(e);
	if (t.length === 0) {
		let t = Y(e).trim();
		return t ? /* @__PURE__ */ (0, V.jsx)("span", {
			className: "text-sm text-foreground",
			children: t
		}) : /* @__PURE__ */ (0, V.jsx)("span", {
			className: "text-muted-foreground",
			children: "—"
		});
	}
	return /* @__PURE__ */ (0, V.jsx)("dl", {
		className: "grid grid-cols-[max-content_minmax(0,1fr)] gap-x-2 gap-y-1 text-xs",
		children: t.map((e, t) => /* @__PURE__ */ (0, V.jsxs)(B.Fragment, { children: [/* @__PURE__ */ (0, V.jsx)("dt", {
			className: "font-medium text-muted-foreground",
			children: e.key
		}), /* @__PURE__ */ (0, V.jsx)("dd", {
			className: "min-w-0 break-words text-foreground",
			children: /* @__PURE__ */ (0, V.jsx)(J, { node: e.value })
		})] }, `${e.key}-${t}`))
	});
}
function rl({ node: e }) {
	let t = dl(e);
	return /* @__PURE__ */ (0, V.jsx)("div", {
		className: "max-h-40 min-w-40 max-w-xl overflow-auto text-xs",
		onClick: (e) => e.stopPropagation(),
		children: /* @__PURE__ */ (0, V.jsx)(et, {
			data: t,
			defaultOpenDepth: 0
		})
	});
}
function il(e) {
	if (!e) return [];
	if (e.kind === "map") return al(e.fields ?? []);
	if (e.kind === "list") return (e.items ?? []).flatMap((e) => il(e));
	let t = e.kind === "code" ? e.source : e.text ?? e.plain;
	if (!t) return [];
	try {
		return ol(JSON.parse(t));
	} catch {
		return [];
	}
}
function al(e) {
	let t = e.find((e) => /^(key|name)$/i.test(e.name)), n = e.find((e) => /^value$/i.test(e.name));
	if (t && n) {
		let e = Y(t.value).trim();
		if (e) return [{
			key: e,
			value: n.value
		}];
	}
	return e.map((e) => ({
		key: e.label || e.name,
		value: e.value
	}));
}
function ol(e) {
	if (Array.isArray(e)) return e.flatMap(ol);
	if (!e || typeof e != "object") return [];
	let t = e, n = Object.keys(t).find((e) => /^(key|name)$/i.test(e)), r = Object.keys(t).find((e) => /^value$/i.test(e));
	return n && r && typeof t[n] == "string" ? [{
		key: t[n],
		value: sl(t[r])
	}] : Object.keys(t).sort((e, t) => e.localeCompare(t)).map((e) => ({
		key: e,
		value: sl(t[e])
	}));
}
function sl(e) {
	if (Array.isArray(e)) return {
		kind: "list",
		plain: JSON.stringify(e),
		items: e.map(sl)
	};
	if (e && typeof e == "object") {
		let t = e;
		return {
			kind: "map",
			plain: JSON.stringify(e),
			fields: Object.keys(t).sort((e, t) => e.localeCompare(t)).map((e) => ({
				name: e,
				value: sl(t[e])
			}))
		};
	}
	let t = e === null ? "null" : String(e ?? "");
	return {
		kind: "text",
		text: t,
		plain: t
	};
}
function cl(e) {
	return il(e).map((e) => `${e.key}=${ul(e.value)}`).filter((e) => !e.endsWith("="));
}
function ll(e) {
	return cl(e).join(", ");
}
function ul(e) {
	let t = Y(e).trim();
	if (t) return t;
	let n = dl(e);
	return n === void 0 ? "" : JSON.stringify(n);
}
function dl(e) {
	if (!e) return;
	if (e.kind === "map") return Object.fromEntries((e.fields ?? []).map((e) => [e.name, dl(e.value)]));
	if (e.kind === "list") return (e.items ?? []).map(dl);
	let t = e.kind === "code" ? e.source ?? e.plain : e.text ?? e.plain;
	if (t === void 0) return Y(e);
	try {
		return JSON.parse(t);
	} catch {
		return t;
	}
}
function fl(e) {
	let t = dl(e);
	return t === void 0 ? "" : JSON.stringify(t);
}
function pl(e, t) {
	return el(t) ? Ur(ml(e)) : [];
}
function ml(e) {
	if (!e) return [];
	if (e.kind === "map" && e.fields?.length) return Object.fromEntries(e.fields.map((e) => [e.name, Y(e.value)]).filter(([, e]) => e.trim() !== ""));
	if (e.kind === "list" && e.items?.length) return e.items.map((e) => Y(e)).filter(Boolean);
	let t = Y(e).trim();
	return t ? t.split(/[,\s]+/).map((e) => e.trim()).filter(Boolean) : [];
}
function hl(e) {
	if (!e) return !1;
	if (e.kind === "map") return (e.fields ?? []).length > 0;
	if (e.kind === "table") return (e.rows ?? []).length > 0;
	if (e.kind !== "list") return !1;
	let t = e.items ?? [];
	return t.length > 0 && t.every(hl);
}
function gl(e, t) {
	return {
		kind: "map",
		fields: t.flatMap((t) => {
			let n = e.cells[t.name];
			return n ? [{
				name: t.name,
				label: Y(t.header) || t.label || Nl(t.name),
				value: n
			}] : [];
		})
	};
}
function _l(e, t, n) {
	let r = Cl(t.find((t) => xl(e.cells[t.name]))?.label ?? t[0]?.label ?? t[0]?.name ?? "Item"), i = yl(e, t);
	return i ? `${r} ${n + 1}: ${i}` : `${r} ${n + 1}`;
}
function vl(e, t, n) {
	return `${n}-${yl(e, t) || t.map((t) => Y(e.cells[t.name])).join("|")}`;
}
function yl(e, t) {
	for (let n of [
		"name",
		"title",
		"agreementname",
		"companyname",
		"customernumber",
		"id",
		"guid"
	]) for (let r of t) {
		let t = bl(e.cells[r.name], n);
		if (t) return t;
	}
	for (let n of t) {
		let t = Y(e.cells[n.name]).trim();
		if (t) return t;
	}
	return "";
}
function bl(e, t) {
	if (!e) return "";
	if (e.kind === "map") {
		for (let n of e.fields ?? []) if (Sl(n.name) === t) {
			let e = Y(n.value).trim();
			if (e) return e;
		}
	}
	for (let n of e.fields ?? []) {
		let e = bl(n.value, t);
		if (e) return e;
	}
	for (let n of e.items ?? []) {
		let e = bl(n, t);
		if (e) return e;
	}
	return "";
}
function xl(e) {
	return !!(Y(e).trim() || (e?.fields ?? []).length > 0);
}
function Sl(e) {
	return e.toLowerCase().replace(/[^a-z0-9]/g, "");
}
function Cl(e) {
	let t = Nl(e).trim();
	return t.endsWith("s") ? t.slice(0, -1) : t;
}
function wl({ node: e }) {
	let t = e.roots ?? [];
	return t.length === 0 ? null : /* @__PURE__ */ (0, V.jsx)($e, {
		roots: t,
		getChildren: (e) => e.children,
		getKey: (e) => e.id,
		defaultOpen: (e, t) => t < 1,
		renderRow: ({ node: e }) => /* @__PURE__ */ (0, V.jsx)("div", {
			className: "min-w-0 flex-1 truncate",
			children: /* @__PURE__ */ (0, V.jsx)(J, { node: e.label })
		})
	});
}
function Tl(e) {
	return e === "pill" || e === "rounded" || e === "square";
}
function El({ node: e }) {
	let t = { variant: "label" };
	return e.badgeLabel !== void 0 && (t.label = e.badgeLabel), e.badgeValue !== void 0 && (t.value = e.badgeValue), e.badgeColor !== void 0 && (t.color = e.badgeColor), e.badgeText !== void 0 && (t.textColor = e.badgeText), e.badgeIcon !== void 0 && (t.icon = e.badgeIcon), Tl(e.badgeShape) && (t.shape = e.badgeShape), /* @__PURE__ */ (0, V.jsx)(D, { ...t });
}
function Dl({ node: e }) {
	let t = {
		causedBy: e.causedBy ?? [],
		frames: e.frames ?? [],
		language: "java",
		...e.exceptionClass === void 0 ? {} : { exceptionClass: e.exceptionClass },
		...e.message === void 0 ? {} : { message: e.message }
	};
	return /* @__PURE__ */ (0, V.jsx)(as, {
		input: t,
		className: "space-y-1"
	});
}
function Ol({ node: e }) {
	let [t, n] = (0, B.useState)(!1);
	return /* @__PURE__ */ (0, V.jsxs)("details", {
		className: "rounded-md border border-border bg-background",
		onToggle: (e) => n(e.currentTarget.open),
		children: [/* @__PURE__ */ (0, V.jsxs)("summary", {
			className: "cursor-pointer break-words px-3 py-2 text-sm font-medium text-foreground",
			children: [/* @__PURE__ */ (0, V.jsx)(J, { node: e.label }), /* @__PURE__ */ (0, V.jsx)("span", {
				className: "ml-2 inline-block text-xs font-normal text-muted-foreground",
				children: t ? "Show less" : "Show more"
			})]
		}), e.content && /* @__PURE__ */ (0, V.jsx)("div", {
			className: "border-t border-border px-3 py-3",
			children: /* @__PURE__ */ (0, V.jsx)(J, { node: e.content })
		})]
	});
}
function kl({ node: e }) {
	let t = [e.id, e.payload].filter(Boolean).join("\n") || void 0, n = e.label ? /* @__PURE__ */ (0, V.jsx)(J, { node: e.label }) : e.text;
	return e.href ? /* @__PURE__ */ (0, V.jsx)("a", {
		href: e.href,
		target: "_blank",
		rel: "noopener noreferrer",
		title: t,
		className: "inline-flex items-center rounded-md border border-border bg-background px-3 py-1.5 text-sm font-medium text-foreground hover:bg-accent",
		children: n
	}) : /* @__PURE__ */ (0, V.jsx)("button", {
		type: "button",
		title: t,
		className: "inline-flex items-center rounded-md border border-border bg-muted px-3 py-1.5 text-sm font-medium text-foreground",
		children: n
	});
}
function Al({ node: e }) {
	let t = e.items ?? [];
	return t.length === 0 ? null : /* @__PURE__ */ (0, V.jsx)("div", {
		className: "flex flex-wrap gap-2",
		children: t.map((e, t) => /* @__PURE__ */ (0, V.jsx)(J, { node: e }, t))
	});
}
function jl({ node: e }) {
	let t = Ge(e.html ?? "");
	if (!t) return null;
	let n = Be(t) ? "div" : "span";
	return /* @__PURE__ */ (0, V.jsx)(n, { dangerouslySetInnerHTML: { __html: t } });
}
function Ml(e, t) {
	if (!e) return t?.includes("\n") ? { whiteSpace: "pre-wrap" } : void 0;
	let n = {};
	e.color && (n.color = e.color), e.backgroundColor && (n.backgroundColor = e.backgroundColor), e.bold && (n.fontWeight = 700), e.faint && (n.opacity = .7), e.italic && (n.fontStyle = "italic");
	let r = [];
	return e.underline && r.push("underline"), e.strikethrough && r.push("line-through"), r.length > 0 && (n.textDecoration = r.join(" ")), (e.textTransform === "uppercase" || e.textTransform === "lowercase" || e.textTransform === "capitalize") && (n.textTransform = e.textTransform), e.maxWidth && e.maxWidth > 0 && (n.maxWidth = `${e.maxWidth}ch`), e.maxLines && e.maxLines > 0 && (n.display = "-webkit-box", n.overflow = "hidden", n.WebkitLineClamp = e.maxLines, n.WebkitBoxOrient = "vertical", n.whiteSpace = "pre-wrap"), e.truncateMode === "suffix" && e.maxWidth && (!e.maxLines || e.maxLines <= 1) && (n.overflow = "hidden", n.textOverflow = "ellipsis", n.whiteSpace = "nowrap", n.display = "inline-block", n.verticalAlign = "bottom"), e.monospace && (n.fontFamily = "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, \"Liberation Mono\", \"Courier New\", monospace"), !n.whiteSpace && t?.includes("\n") && (n.whiteSpace = "pre-wrap"), Object.keys(n).length > 0 ? n : void 0;
}
function Nl(e) {
	return e.replace(/([a-z])([A-Z])/g, "$1 $2").replace(/_/g, " ").replace(/\b\w/g, (e) => e.toUpperCase());
}
function Pl(e) {
	return {
		command: e.command ?? "",
		args: e.args ?? [],
		flags: e.flags ?? {},
		target: e.target ?? "Dialog",
		autoRun: e.autoRun ?? !1
	};
}
function Fl(e, t, n) {
	return {
		request: e,
		operation: t?.resolveCommand?.(e, n) ?? Il(e, n)
	};
}
function Il(e, t) {
	let n = Ll(e.command);
	if (n) return t.find((e) => {
		let t = Ll(e.operation["x-clicky"]?.command);
		return t && t === n ? !0 : Ll(e.operation.operationId?.replaceAll("_", "/")) === n;
	});
}
function Ll(e) {
	return (e ?? "").trim().replace(/^\//, "").replace(/\s+/g, "/").replace(/\/+/g, "/").toLowerCase();
}
function Rl(e, t) {
	let n = { ...t.flags }, r = [...t.args ?? []], i = e.filter((e) => Te(e)), a = 0;
	for (let e of i) if (e.name === "args") {
		let e = r.slice(a);
		e.length > 0 && (n.args = e.join(",")), a = r.length;
	} else a < r.length && (n[e.name] = r[a] ?? "", a += 1);
	return n;
}
function zl(e, t) {
	return e.filter((e) => e.required && (t[e.name] ?? "").trim() === "");
}
async function Bl(e, t, n) {
	return e.executeCommand(t.path, t.method, ye(De(n), t.operation.parameters ?? []), { Accept: "application/json+clicky" });
}
function Vl(e) {
	let t = e.operation;
	if (!t) return;
	let n = Rl(t.operation.parameters ?? [], e.request), r = ye(De(n), t.operation.parameters ?? []), i = Hl(t.path, r);
	return t.method.toUpperCase(), i;
}
function Hl(e, t) {
	let n = new URL(e, "http://localhost"), r = { ...t };
	for (let [e, i] of Object.entries(t)) if (n.pathname.includes(`{${e}}`)) {
		if (Array.isArray(i)) throw Error(`Path parameter ${e} must be a scalar value`);
		n.pathname = n.pathname.replace(`{${e}}`, encodeURIComponent(i)), delete r[e];
	}
	for (let [e, t] of Object.entries(r)) n.searchParams.set(e, Array.isArray(t) ? t.join(",") : t);
	return `${n.pathname}${n.search}`;
}
function Ul(e) {
	if (e === "_self") return "_self";
	if (e === "_window" || e === "_tab") return "_blank";
}
function Wl(e) {
	return typeof e == "object" && !!e && ("stdout" in e || "output" in e || "message" in e);
}
function Gl(e) {
	return e?.filterValue === void 0 ? Y(e) : e.filterValue;
}
var Kl = /* @__PURE__ */ new Set([
	"number",
	"int",
	"integer",
	"float",
	"decimal",
	"currency",
	"bytes",
	"duration",
	"boolean",
	"bool",
	"date",
	"datetime",
	"time",
	"timestamp"
]);
function ql(e, t) {
	if (e == null) return;
	if (e.kind === "execution-tree") return e.executionRoots ?? [];
	if (e.kind === "object-graph") return e.objects ?? [];
	if (e.kind === "map" || e.kind === "list") return dl(e);
	if (e.kind === "code" && e.language === "json") return JSON.parse(e.source ?? "null");
	if (e.filterValue !== void 0) return e.filterValue;
	let n = Y(e);
	return n === "" && t?.type !== void 0 && Kl.has(t.type) ? null : n;
}
function Jl(e, t) {
	return Object.fromEntries(Object.entries(e.cells).map(([e, n]) => [e, ql(n, t.get(e))]));
}
function Y(e) {
	if (!e) return "";
	if (e.kind === "heading" || e.kind === "blockquote") return Y(e.content) || e.plain || "";
	if (e.kind === "admonition") return [Y(e.label), Y(e.content)].filter(Boolean).join(" ") || e.plain || "";
	if (e.kind === "footnote-ref") return e.id ? `[^${e.id}]` : e.plain || "";
	if (e.kind === "footnote") return Y(e.content) || e.plain || "";
	if (e.kind === "footnotes") return (e.items ?? []).map((e) => Y(e)).join(" ") || e.plain || "";
	if (e.plain) return e.plain;
	if (e.text) return e.text;
	if (e.kind === "html") return ke(Ge(e.html ?? "")).trim();
	if (e.kind === "code") return e.source ?? "";
	if ((e.kind === "link" || e.kind === "link-command") && e.children?.length) return [e.text, ...e.children.map((e) => Y(e))].join("");
	if (e.kind === "button" && e.label) return Y(e.label);
	if (e.kind === "button-group") return (e.items ?? []).map((e) => Y(e)).join(" ");
	if (e.kind === "text" && e.children?.length) return [e.text, ...e.children.map((e) => Y(e))].join("");
	if (e.kind === "list" && e.items?.length) return e.items.map((e) => Y(e)).join(" ");
	if (e.kind === "map" && e.fields?.length) return e.fields.map((e) => {
		let t = Y(e.value), n = e.label ?? e.name;
		return n ? `${n}: ${t}` : t;
	}).join(" ");
	if (e.kind === "tree" && e.roots?.length) {
		let t = (e) => [Y(e.label), ...(e.children ?? []).map((e) => t(e))].filter(Boolean).join(" ");
		return e.roots.map((e) => t(e)).join(" ");
	}
	return e.kind === "table" && e.rows?.length ? e.rows.map((e) => Object.values(e.cells ?? {}).map((e) => Y(e)).filter(Boolean).join(" ")).filter(Boolean).join(" ") : "";
}
function Yl(e) {
	return e.kind === "text" || e.kind === "link" || e.kind === "link-command" || e.kind === "icon" || e.kind === "html" || e.kind === "footnote-ref" || e.kind === "badge" || e.kind === "button" || e.kind === "button-group";
}
var Xl = (0, B.createContext)((e) => new EventSource(e));
function Zl() {
	return (0, B.useContext)(Xl);
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/MetricTile.tsx
var Ql = {
	neutral: "bg-muted-foreground/40",
	success: "bg-green-500",
	warning: "bg-yellow-500",
	danger: "bg-red-500",
	info: "bg-blue-500"
}, $l = {
	neutral: "text-foreground",
	success: "text-green-600 dark:text-green-400",
	warning: "text-yellow-600 dark:text-yellow-400",
	danger: "text-red-600 dark:text-red-400",
	info: "text-blue-600 dark:text-blue-400"
};
function eu({ label: e, value: t, barPct: n, tone: r = "neutral", className: a }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i("flex min-w-0 flex-col gap-0.5 rounded-md border border-border bg-background px-density-2 py-1", a),
		"data-metric": e,
		children: [
			/* @__PURE__ */ (0, V.jsx)("span", {
				className: "text-[9px] uppercase leading-none tracking-wide text-muted-foreground",
				children: e
			}),
			/* @__PURE__ */ (0, V.jsx)("span", {
				className: i("truncate text-xs font-semibold tabular-nums", $l[r]),
				children: t
			}),
			n !== void 0 && /* @__PURE__ */ (0, V.jsx)("div", {
				className: "h-0.5 w-full overflow-hidden rounded-full bg-muted",
				children: /* @__PURE__ */ (0, V.jsx)("div", {
					className: i("h-full transition-all duration-300", Ql[r]),
					style: { width: `${n}%` },
					"data-metric-bar": ""
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/OpenInFilesLink.tsx
function tu({ target: e, onOpenFile: t, className: n }) {
	return /* @__PURE__ */ (0, V.jsxs)(o, {
		type: "button",
		variant: "link",
		size: "sm",
		className: i("h-auto gap-1 p-0 text-xs", n),
		title: `Open ${e.file}${e.line ? `:${e.line}` : ""} in Files`,
		onClick: (n) => {
			n.stopPropagation(), t(e);
		},
		children: [/* @__PURE__ */ (0, V.jsx)(c, {
			icon: dn,
			className: "text-xs"
		}), "Open in Files"]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/copy.ts
var nu = /* @__PURE__ */ new Set(["stdout", "stderr"]);
function ru(e, t) {
	return t ? [{
		label: e,
		value: t
	}] : [];
}
function iu(e, t, n) {
	return t ? [n ? {
		title: e,
		lang: n,
		body: t
	} : {
		title: e,
		body: t
	}] : [];
}
function au(e) {
	if (e.file) return e.line ? `${e.file}:${e.line}` : e.file;
}
function ou(e) {
	let t = e.failure_detail?.summary;
	return [e.message, t === e.message ? void 0 : t].filter(Boolean).join("\n");
}
function su(e) {
	let t = e.failure_detail;
	return t ? [
		t.matcher && `matcher: ${t.matcher}`,
		t.expected && `expected: ${t.expected}`,
		t.actual && `actual: ${t.actual}`,
		t.location && `at: ${t.location}`
	].filter(Boolean).join("\n") : "";
}
function cu(e) {
	let t = e.duration ? L(e.duration) : void 0;
	return {
		facts: [
			...ru("Result", kt(e) ?? void 0),
			...ru("Tool", e.framework),
			...ru("Location", au(e)),
			...ru("Duration", t)
		],
		brief: [
			...iu("Error", ou(e)),
			...iu("Expectation", su(e)),
			...iu("Stack", e.failure_detail?.stack),
			...e.log ? iu("output", e.log) : [...iu("stdout", e.stdout), ...iu("stderr", e.stderr)]
		],
		items: []
	};
}
function lu(e, t) {
	let n = /* @__PURE__ */ new Set();
	return e.filter((e) => {
		let r = t(e);
		return !n.has(r) && (n.add(r), !0);
	});
}
function uu(e) {
	return {
		facts: lu(e.flatMap((e) => e.facts ?? []), (e) => e.label),
		brief: lu(e.flatMap((e) => e.brief ?? []), (e) => e.title),
		items: lu(e.flatMap((e) => e.items ?? []), (e) => e.label)
	};
}
function du(e) {
	let t = Math.max(0, ...(e.match(/`+/g) ?? []).map((e) => e.length));
	return "`".repeat(Math.max(3, t + 1));
}
function fu(e, t, n) {
	let r = [`# ${e.name}`];
	n.length > 0 && r.push("", ...n.map((e) => `- **${e.label}:** ${e.value}`));
	for (let e of t) {
		if (!e.body) continue;
		if (r.push("", `## ${e.title}`, ""), e.markdown) {
			r.push(e.body);
			continue;
		}
		let t = du(e.body);
		r.push(`${t}${e.lang ?? ""}`, e.body, t);
	}
	return r.join("\n");
}
function pu(e, t) {
	let n = cu(e), r = uu([
		{
			facts: n.facts,
			brief: n.brief.filter((e) => !nu.has(e.title))
		},
		t,
		{ brief: n.brief.filter((e) => nu.has(e.title)) }
	]);
	return fu(e, r.brief, r.facts);
}
function mu(e) {
	let { children: t, ...n } = e;
	return JSON.stringify(n, null, 2);
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/attempts.ts
function hu(e) {
	return e.timed_out ? "timedout" : e.failed ? "failed" : e.skipped ? "skipped" : e.passed ? "passed" : e.pending ? "pending" : "running";
}
function gu(e) {
	let t = hu(e);
	return t === "running" || t === "pending";
}
function _u(e, t, n) {
	let { passed: r, failed: i, skipped: a, warned: o, aborted: s, pending: c, running: l, timed_out: u, failure_detail: d, message: f, stdout: p, stderr: m, log: h, duration: g, ..._ } = e, v = hu(t);
	return {
		..._,
		...v === "timedout" ? { timed_out: !0 } : {},
		...v === "failed" ? { failed: !0 } : {},
		...v === "skipped" ? { skipped: !0 } : {},
		...v === "passed" ? { passed: !0 } : {},
		...v === "pending" ? { pending: !0 } : {},
		...v === "running" ? { running: !0 } : {},
		...t.message === void 0 ? {} : { message: t.message },
		...t.stdout === void 0 ? {} : { stdout: t.stdout },
		...t.stderr === void 0 ? {} : { stderr: t.stderr },
		...t.log === void 0 ? {} : { log: t.log },
		...t.duration === void 0 ? {} : { duration: t.duration },
		...t.command === void 0 ? {} : { command: t.command },
		...n && d ? { failure_detail: d } : {},
		attempts: [t]
	};
}
function vu(e) {
	let t = e.attempts ?? [];
	return t.map((n, r) => ({
		id: r + 1,
		status: hu(n),
		...n.duration !== void 0 && n.duration > 0 ? { durationMs: n.duration / 1e6 } : {},
		node: _u(e, n, r === t.length - 1)
	}));
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/adapter.ts
var yu = class {
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
	resolveRowLabel(e, t) {
		return this.firstWith(e, t, (e) => !!e.rowLabel)?.rowLabel?.({
			node: e,
			runner: t
		});
	}
	resolveFrameworkBadge(e, t) {
		return this.firstWith(e, t, (e) => !!e.frameworkBadge)?.frameworkBadge?.({
			node: e,
			runner: t
		}) ?? null;
	}
	resolveFrameworkIcon(e, t) {
		return this.firstWith(e, t, (e) => !!e.renderFrameworkIcon);
	}
	resolveActions(e, t) {
		return this.firstWith(e, t, (e) => !!e.nodeActions);
	}
	resolveTreeChildren(e, t) {
		let n = this.firstWith(e, t, (e) => !!e.treeChildren);
		return n?.treeChildren ? n.treeChildren({
			node: e,
			runner: t
		}) : e.children;
	}
	resolveAttempts(e, t) {
		let n = this.firstWith(e, t, (e) => !!e.detailAttempts)?.detailAttempts?.({
			node: e,
			runner: t
		});
		return n && n.length > 0 ? n : vu(e);
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
		return uu(this.adapters.flatMap((r) => r.copy && r.match(e, t) ? [r.copy(n)] : []));
	}
	ownsScroll(e, t) {
		let n = this.resolveDetail(e, t);
		return n?.ownsScroll ? n.ownsScroll({
			node: e,
			runner: t
		}) : !1;
	}
};
function bu(e = []) {
	return new yu(e);
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/context.ts
var xu = (0, B.createContext)(null), Su = xu.Provider;
function Cu() {
	let e = (0, B.useContext)(xu);
	if (!e) throw Error("useTestRunner must be used within a <TestRunner> / TestRunnerProvider");
	return e;
}
// ../../../clicky-ui/packages/ui/src/hooks/use-resource-clock.ts
function wu(e) {
	let [t, n] = (0, B.useState)(Date.now);
	return (0, B.useEffect)(() => {
		if (!e) return;
		let t = setInterval(() => n(Date.now()), 1e3);
		return () => clearInterval(t);
	}, [e]), t;
}
// ../../../clicky-ui/packages/ui/src/data/ElapsedDuration.tsx
function Tu({ startedAt: e, finishedAt: t, format: n = qe, fallback: r = null, className: a }) {
	let o = wu(!!e && !t), s = Ze(e, t, o);
	return s === void 0 ? /* @__PURE__ */ (0, V.jsx)(V.Fragment, { children: r }) : /* @__PURE__ */ (0, V.jsx)("span", {
		className: i("tabular-nums", a),
		children: n(s)
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestTreeNode.tsx
var Eu = {
	failed: "text-red-600 dark:text-red-400",
	timedout: "text-amber-600 dark:text-amber-400",
	warned: "text-amber-600 dark:text-amber-400",
	aborted: "text-muted-foreground",
	running: "text-blue-600 dark:text-blue-400",
	skipped: "text-muted-foreground",
	pending: "text-muted-foreground",
	passed: "text-foreground"
}, Du = /* @__PURE__ */ new Set([
	"failed",
	"timedout",
	"warned",
	"aborted",
	"skipped"
]);
function Ou({ node: e }) {
	let t = it(e);
	return t.total === 0 || t.passed === t.total ? null : /* @__PURE__ */ (0, V.jsxs)("span", {
		className: "flex shrink-0 items-center gap-1",
		children: [
			t.passed > 0 && /* @__PURE__ */ (0, V.jsx)(D, {
				tone: "success",
				variant: "outline",
				className: "rounded-sm",
				size: "xxs",
				count: t.passed
			}),
			t.failed > 0 && /* @__PURE__ */ (0, V.jsx)(D, {
				tone: "danger",
				variant: "outline",
				className: "rounded-sm",
				size: "xxs",
				count: t.failed
			}),
			t.warned > 0 && /* @__PURE__ */ (0, V.jsx)(D, {
				tone: "warning",
				variant: "outline",
				className: "rounded-sm",
				size: "xxs",
				count: t.warned
			}),
			t.timedout > 0 && /* @__PURE__ */ (0, V.jsx)(D, {
				tone: "danger",
				variant: "outline",
				className: "rounded-sm",
				size: "xxs",
				count: t.timedout
			}),
			t.aborted > 0 && /* @__PURE__ */ (0, V.jsx)(D, {
				tone: "neutral",
				variant: "outline",
				className: "rounded-sm",
				size: "xxs",
				count: t.aborted
			}),
			t.skipped > 0 && /* @__PURE__ */ (0, V.jsx)(D, {
				tone: "neutral",
				variant: "outline",
				className: "rounded-sm",
				size: "xxs",
				count: t.skipped
			}),
			t.running > 0 && /* @__PURE__ */ (0, V.jsx)(D, {
				tone: "info",
				variant: "outline",
				className: "rounded-sm",
				size: "xxs",
				count: t.running
			}),
			t.pending > 0 && /* @__PURE__ */ (0, V.jsx)(D, {
				tone: "neutral",
				variant: "outline",
				className: "rounded-sm",
				size: "xxs",
				count: t.pending
			})
		]
	});
}
function ku({ node: e, selected: t }) {
	let n = Cu(), r = (e.children?.length ?? 0) > 0, a = dt(e), o = kt(e), s = n.adapters.resolveFrameworkIcon(e, n)?.renderFrameworkIcon?.({
		node: e,
		runner: n
	}), l = n.adapters.resolveRowLeading(e, n), u = e.duration || (r ? ft(e) : 0), d = !!e.running && !!e.started && !e.duration, f = !!n.onRerun && e.framework !== "task", p = !!n.onStop && !!e.can_stop && !!e.task_id, m = a && (Du.has(a) || e.framework === "task" && (a === "running" || a === "pending")) ? (e.message ?? "").trim() : "";
	return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
		o ? /* @__PURE__ */ (0, V.jsx)(Et, { outcome: o }) : /* @__PURE__ */ (0, V.jsx)(c, {
			icon: hn,
			className: "shrink-0 text-muted-foreground"
		}),
		s == null ? /* @__PURE__ */ (0, V.jsx)(Tt, {
			tool: Vt(e),
			className: "text-sm"
		}) : /* @__PURE__ */ (0, V.jsx)("span", {
			className: "flex shrink-0 items-center text-sm",
			children: s
		}),
		l?.renderRowLeading?.({
			node: e,
			runner: n
		}),
		/* @__PURE__ */ (0, V.jsx)("span", {
			className: i("min-w-0 truncate", t ? "font-semibold" : "font-medium", a && Eu[a] || "text-foreground"),
			title: m || void 0,
			children: n.adapters.resolveRowLabel(e, n) ?? pt(e.name, e.framework)
		}),
		m ? /* @__PURE__ */ (0, V.jsx)("span", {
			className: "min-w-0 flex-1 shrink-[100] truncate text-xs text-muted-foreground",
			title: m,
			"data-row-message": "",
			children: m
		}) : /* @__PURE__ */ (0, V.jsx)("span", { className: "flex-1" }),
		e.progress && (e.progress.total > 0 || e.progress.phase) && !e.passed && !e.failed && /* @__PURE__ */ (0, V.jsxs)("span", {
			className: "flex min-w-0 items-center gap-1 text-xs text-blue-600 dark:text-blue-400",
			"data-row-progress": "",
			children: [e.progress.phase && /* @__PURE__ */ (0, V.jsx)("span", {
				className: "truncate",
				title: e.progress.phase,
				children: e.progress.phase
			}), e.progress.total > 0 && /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "shrink-0 tabular-nums",
				children: [
					e.progress.done,
					"/",
					e.progress.total
				]
			})]
		}),
		d ? /* @__PURE__ */ (0, V.jsx)(Tu, {
			startedAt: e.started,
			finishedAt: void 0,
			format: (e) => L(e * 1e6),
			className: "shrink-0 text-xs text-muted-foreground"
		}) : u > 0 && /* @__PURE__ */ (0, V.jsx)("span", {
			className: "shrink-0 text-xs text-muted-foreground",
			children: L(u)
		}),
		r && /* @__PURE__ */ (0, V.jsx)(Ou, { node: e }),
		p && /* @__PURE__ */ (0, V.jsx)(F, {
			icon: Wt,
			label: "Stop",
			disabled: n.busy.stop,
			onClick: (t) => {
				t.stopPropagation(), n.onStop(e);
			}
		}),
		f && /* @__PURE__ */ (0, V.jsx)(F, {
			icon: wn,
			label: "Rerun",
			disabled: n.busy.rerun,
			onClick: (t) => {
				t.stopPropagation(), n.onRerun(e);
			}
		})
	] });
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestTree.tsx
var Au = (() => {
	let e = /* @__PURE__ */ new WeakMap(), t = 0;
	return (n) => {
		if (n.route_path) return n.route_path;
		let r = e.get(n);
		return r || (r = `tn-${t++}-${n.name}`, e.set(n, r)), r;
	};
})();
function ju({ className: e, actions: t }) {
	let n = Cu();
	return /* @__PURE__ */ (0, V.jsx)($e, {
		className: i("min-h-0", e),
		roots: n.tests,
		getKey: Au,
		getChildren: (e) => n.adapters.resolveTreeChildren(e, n),
		selected: n.selected,
		onSelect: (e) => n.onSelect(e === n.selected ? null : e),
		expandAll: n.expandAll,
		onExpandAllChange: n.onExpandAllChange,
		defaultOpen: (e, t) => t < 1 || ot(e),
		showSearch: !0,
		searchLabel: "Filter tests",
		getSearchText: (e) => `${e.name} ${e.framework ?? ""}`,
		renderRow: ({ node: e, selected: t }) => /* @__PURE__ */ (0, V.jsx)(ku, {
			node: e,
			selected: t
		}),
		rowClass: (e, t) => i("hover:bg-accent", dt(e) === "running" && "border-l-2 border-primary bg-primary/10 font-semibold hover:bg-primary/15", t && "border-l-2 border-primary bg-primary/20"),
		actions: t
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/detailTabContext.ts
var Mu = (0, B.createContext)(null), Nu = Mu.Provider;
function Pu() {
	return (0, B.useContext)(Mu);
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/jvm-stacktrace.ts
var Fu = /^"(?<name>[^"]*)"(?<rest>.*)$/, Iu = /^\s*java\.lang\.Thread\.State:\s+(?<state>[A-Z_]+)(?:\s+\((?<sub>[^)]+)\))?/, Lu = /^\s*at\s+(?<fn>[^\s(]+)\((?<src>[^)]+)\)\s*$/, Ru = /^(?<file>[^:]+):(?<line>\d+)$/, zu = /^\s*-\s+(?<kind>locked|waiting on|waiting to lock|parking to wait for)\b(?<rest>.*)$/;
function Bu(e) {
	let t = e.trim();
	if (!t) return [];
	let n = Vu(t), r = [];
	for (let e of n) {
		let t = Hu(e);
		t && r.push(t);
	}
	return r;
}
function Vu(e) {
	let t = [], n = [], r = e.split("\n");
	for (let e of r) {
		let r = e.replace(/\r$/, "");
		r.startsWith("\"") && n.length > 0 ? (t.push(n.join("\n")), n = [r]) : n.push(r);
	}
	return n.length > 0 && t.push(n.join("\n")), t;
}
function Hu(e) {
	let t = e.split("\n"), n = t[0]?.trim() ?? "", r = Fu.exec(n);
	if (!r?.groups) return null;
	let i = r.groups.name;
	if (!i) return null;
	let a = r.groups.rest ?? "", o = /#(\d+)\b/.exec(a), s = /\bprio=(\d+)/.exec(a), c = /\bnid=(0x[0-9a-f]+)/i.exec(a), l = /\bdaemon\b/.test(a), u = Uu(a) ?? "", d = Xu(u), f = [];
	for (let e = 1; e < t.length; e++) {
		let n = t[e];
		if (n == null || !n.trim()) continue;
		let r = Iu.exec(n);
		if (r?.groups) {
			let e = r.groups.state;
			if (!e) continue;
			let t = r.groups.sub;
			u = t ? `${e} (${t})` : e, d = Xu(e);
			continue;
		}
		let i = Lu.exec(n);
		if (i?.groups) {
			let e = i.groups.fn, t = i.groups.src;
			if (!e || !t) continue;
			let n = {
				functionName: e,
				displayName: $u(e),
				kind: "frame",
				runtime: Zu(e),
				nativeMethod: t === "Native Method",
				...Qu(e)
			}, r = Ru.exec(t);
			r?.groups?.file && r.groups.line ? (n.file = r.groups.file, n.line = Number(r.groups.line), n.location = `${n.file}:${n.line}`) : n.location = t === "Native Method" ? "Native Method" : t, f.push(n);
			continue;
		}
		let a = zu.exec(n);
		if (a?.groups?.kind) {
			let e = a.groups.kind, t = Ju(e);
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
		id: o ? Number(o[1]) : qu(i, a),
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
function Uu(e) {
	let t = e.toLowerCase(), n = t.indexOf("nid=0x");
	if (n === -1) return;
	let r = n + 6, i = r;
	for (; Wu(e[r]);) r++;
	if (r === i || !Ku(e[r])) return;
	for (; Ku(e[r]);) r++;
	let a = e.length;
	for (; a > r && Ku(e[a - 1]);) a--;
	if (e[a - 1] === "]") {
		let n = t.lastIndexOf("[0x", a - 1);
		if (n > r && Gu(e, n + 3, a - 1) && Ku(e[n - 1])) for (a = n; a > r && Ku(e[a - 1]);) a--;
	}
	return e.slice(r, a) || void 0;
}
function Wu(e) {
	if (e === void 0) return !1;
	let t = e.toLowerCase().charCodeAt(0);
	return t >= 48 && t <= 57 || t >= 97 && t <= 102;
}
function Gu(e, t, n) {
	if (t >= n) return !1;
	for (let r = t; r < n; r++) if (!Wu(e[r])) return !1;
	return !0;
}
function Ku(e) {
	return e !== void 0 && e.trim() === "";
}
function qu(e, t) {
	let n = /\btid=(0x[0-9a-f]+)/i.exec(t);
	if (n?.[1]) {
		let e = n[1].slice(2), t = Number.parseInt(e.slice(-8), 16);
		if (Number.isFinite(t)) return t;
	}
	let r = 0;
	for (let t = 0; t < e.length; t++) r = r * 31 + e.charCodeAt(t) | 0;
	return Math.abs(r);
}
function Ju(e) {
	switch (e) {
		case "locked": return "locked";
		case "waiting on": return "waiting_on";
		case "waiting to lock": return "waiting_to_lock";
		case "parking to wait for": return "parking";
		default: return "frame";
	}
}
function Yu(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.state, (t.get(n.state) ?? 0) + 1);
	return t;
}
function Xu(e) {
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
function Zu(e) {
	return qo(e);
}
function Qu(e) {
	let t = e.lastIndexOf(".");
	return t <= 0 ? {} : {
		class: e.slice(0, t),
		method: e.slice(t + 1)
	};
}
function $u(e) {
	let t = e.lastIndexOf(".");
	return t <= 0 ? e : Jo(e.slice(0, t), e.slice(t + 1));
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/stacktrace.ts
function ed(e) {
	let t = e.trim();
	return t ? /^"[^"]*"[^\n]*\b(prio=|tid=|nid=)/m.test(t) ? "jvm" : /^goroutine\s+\d+\s+\[[^\]]+\]:/m.test(t) ? "go" : "unknown" : "unknown";
}
function td(e) {
	let t = ed(e);
	return t === "jvm" ? {
		format: "jvm",
		threads: Bu(e)
	} : t === "go" ? {
		format: "go",
		goroutines: sd(e)
	} : {
		format: "unknown",
		goroutines: [],
		threads: []
	};
}
function nd(e) {
	return e.format === "jvm" ? Yu(e.threads) : e.format === "go" ? ud(e.goroutines) : /* @__PURE__ */ new Map();
}
var rd = /^goroutine\s+(\d+)\s+\[(.+?)\]:$/, id = /^\/usr\/local\/go[\d.]+\/src\//, ad = /^.*?\/go\/src\//, od = /^.*?\/go\/pkg\/mod\//;
function sd(e) {
	let t = e.trim();
	if (!t) return [];
	let n = gd(t), r = [];
	for (let e of n) {
		let t = e.split("\n").map((e) => e.replace(/\r$/, "")), n = t[0]?.trim(), i = rd.exec(n || "");
		if (!i) continue;
		let a = Number(i[1] ?? 0), o = i[2] ?? "", s = dd(o), c = [];
		for (let e = 1; e < t.length; e++) {
			let n = t[e];
			if (n == null) continue;
			let r = n.trim();
			if (!r) continue;
			let i = pd(r), a = c[c.length - 1];
			if (i && a && !a.file) {
				a.file = i.file, a.line = i.line;
				continue;
			}
			let o = r.startsWith("created by ") ? "created_by" : "frame", s = o === "created_by" ? r.slice(11) : r;
			c.push({
				functionName: s,
				displayName: _d(s),
				kind: o,
				runtime: fd(s)
			});
		}
		for (let e of c) e.file && (e.file = yd(e.file), e.location = `${e.file}${e.line ? `:${e.line}` : ""}`);
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
function cd(e, t) {
	let n = e.file ? yd(e.file) : void 0;
	return {
		functionName: e.func,
		displayName: vd(e.func),
		kind: t,
		runtime: fd(e.func),
		...n ? {
			file: n,
			line: e.line,
			location: `${n}${e.line ? `:${e.line}` : ""}`
		} : {}
	};
}
function ld(e) {
	return e.map((e) => {
		if (e.format !== "go" || e.error || e.goroutineId === void 0) throw Error(`stack record ${e.checkpointId}/${e.key} is not a Go goroutine record (format "${e.format}"${e.error ? `, error "${e.error}"` : ""})`);
		let t = [...(e.frames ?? []).map((e) => cd(e, "frame")), ...e.createdBy ? [cd(e.createdBy, "created_by")] : []], n = e.state ?? "", r = [
			n,
			...e.waitMinutes ? [`${e.waitMinutes} minutes`] : [],
			...e.lockedToThread ? ["locked to thread"] : []
		].join(", "), i = t.find((e) => e.kind === "frame")?.functionName;
		return {
			id: e.goroutineId,
			state: dd(n),
			rawState: r,
			frames: t,
			raw: e.raw,
			userFrameCount: t.filter((e) => !e.runtime && e.kind === "frame").length,
			searchText: `goroutine ${e.goroutineId} [${r}]\n${t.map((e) => `${e.functionName} ${e.file || ""}`).join("\n")}`.toLowerCase(),
			...i === void 0 ? {} : { topFunction: i }
		};
	});
}
function ud(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.state, (t.get(n.state) || 0) + 1);
	return t;
}
function dd(e) {
	return (e.split(",")[0] ?? "").trim().toLowerCase();
}
function fd(e) {
	return e.startsWith("runtime.") || e.startsWith("runtime/") || e.startsWith("internal/") || e.startsWith("runtime/internal/") || e.startsWith("syscall.") || e.startsWith("reflect.");
}
function pd(e) {
	let t = e.lastIndexOf(":");
	if (t <= 0) return;
	let n = t + 1, r = n;
	for (; n < e.length && md(e[n] ?? "");) n++;
	if (n === r) return;
	let i = n;
	if (n < e.length) {
		let t = n;
		for (; n < e.length && (e[n] ?? "").trim() === "";) n++;
		if (n === t || e.slice(n, n + 3).toLowerCase() !== "+0x") return;
		n += 3;
		let r = n;
		for (; n < e.length && hd(e[n] ?? "");) n++;
		if (n === r || n !== e.length) return;
	}
	return {
		file: e.slice(0, t),
		line: Number(e.slice(r, i))
	};
}
function md(e) {
	return e >= "0" && e <= "9";
}
function hd(e) {
	let t = e.toLowerCase();
	return md(e) || t >= "a" && t <= "f";
}
function gd(e) {
	let t = [], n = [];
	for (let r of e.split("\n")) r.trim() ? n.push(r) : n.length > 0 && (t.push(n.join("\n")), n = []);
	return n.length > 0 && t.push(n.join("\n")), t;
}
function _d(e) {
	let t = e.trim(), n = t.indexOf("(");
	return n !== -1 && (t = t.slice(0, n)), vd(t);
}
function vd(e) {
	return e.replace(/^((?:[^./\s]+\/)+)([^./\s]+)\./, "$2.");
}
function yd(e) {
	return e.replace(id, "").replace(ad, "").replace(od, "");
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/state-styles.ts
function bd(e) {
	return e.includes("running") ? "bg-green-50 text-green-700 dark:bg-green-500/20 dark:text-green-300" : e.includes("chan") || e.includes("wait") ? "bg-blue-50 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300" : e.includes("sleep") ? "bg-amber-50 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300" : e.includes("select") ? "bg-violet-50 text-violet-700 dark:bg-violet-500/20 dark:text-violet-300" : "bg-muted text-muted-foreground";
}
function xd(e) {
	return e.includes("running") ? "bg-green-500" : e.includes("chan") || e.includes("wait") ? "bg-blue-500" : e.includes("sleep") ? "bg-amber-500" : e.includes("select") ? "bg-violet-500" : "bg-muted-foreground/40";
}
function Sd(e) {
	switch (e.toLowerCase()) {
		case "runnable": return "bg-green-50 text-green-700 dark:bg-green-500/20 dark:text-green-300";
		case "blocked": return "bg-red-50 text-red-700 dark:bg-red-500/20 dark:text-red-300";
		case "waiting":
		case "timed_waiting": return "bg-blue-50 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300";
		case "new": return "bg-amber-50 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300";
		case "terminated": return "bg-muted text-muted-foreground";
		default: return "bg-muted text-muted-foreground";
	}
}
function Cd(e) {
	switch (e.toLowerCase()) {
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
function wd({ goroutine: e, search: t, hideRuntimeOnly: n }) {
	let r = n ? e.frames.filter((e) => !e.runtime || e.kind === "created_by") : e.frames, i = e.state === "running" || !!t;
	return /* @__PURE__ */ (0, V.jsxs)("details", {
		className: "border-0 bg-transparent",
		open: i,
		children: [/* @__PURE__ */ (0, V.jsx)("summary", {
			className: "cursor-pointer list-none px-0 py-1",
			children: /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex items-center gap-2 flex-wrap",
				children: [
					/* @__PURE__ */ (0, V.jsxs)("span", {
						className: "font-mono text-xs font-semibold text-foreground",
						children: ["g", e.id]
					}),
					/* @__PURE__ */ (0, V.jsxs)("span", {
						className: `inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] ${bd(e.state)}`,
						children: [/* @__PURE__ */ (0, V.jsx)("span", { className: `h-2 w-2 rounded-full ${xd(e.state)}` }), e.rawState]
					}),
					/* @__PURE__ */ (0, V.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [r.length, "f"]
					}),
					e.userFrameCount > 0 && /* @__PURE__ */ (0, V.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [e.userFrameCount, "u"]
					}),
					e.topFunction && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "truncate text-[11px] text-muted-foreground",
						children: e.topFunction
					})
				]
			})
		}), /* @__PURE__ */ (0, V.jsx)("div", {
			className: "pl-3 py-1 space-y-0.5",
			children: r.map((t, n) => /* @__PURE__ */ (0, V.jsx)(Td, { frame: t }, `${e.id}-${n}`))
		})]
	});
}
function Td({ frame: e }) {
	return /* @__PURE__ */ (0, V.jsx)("div", {
		className: e.runtime ? "text-muted-foreground" : "text-foreground",
		children: /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex items-start gap-1.5",
			children: [/* @__PURE__ */ (0, V.jsx)(c, {
				icon: e.kind === "created_by" ? ge : e.runtime ? fe : ve,
				className: "shrink-0 mt-0.5 text-[11px]"
			}), /* @__PURE__ */ (0, V.jsx)("div", {
				className: "min-w-0",
				children: /* @__PURE__ */ (0, V.jsxs)("div", {
					className: "break-all font-mono text-[11px] font-semibold leading-4",
					children: [e.displayName, e.location && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "ml-2 text-[10px] font-normal opacity-80",
						children: e.location
					})]
				})
			})]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/JvmStackTrace.tsx
function Ed({ frames: e, hideRuntimeOnly: t = !1, className: n, resolveSource: r, frameActions: i }) {
	let a = t ? e.filter((e) => e.kind !== "frame" || !e.runtime) : e;
	return a.length === 0 ? null : /* @__PURE__ */ (0, V.jsx)("div", {
		className: n,
		children: a.map((e, t) => {
			let n = r?.(e), { line: a, ...o } = n ?? {}, s = n ? {
				...e,
				...o
			} : e;
			return /* @__PURE__ */ (0, V.jsx)(rs, {
				frame: s,
				index: t,
				...i ? { frameActions: i } : {}
			}, `${e.functionName}-${t}`);
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/ThreadCard.tsx
function Dd({ thread: e, search: t, hideRuntimeOnly: n, resolveSource: r, frameActions: i }) {
	let a = n ? e.frames.filter((e) => e.kind !== "frame" || !e.runtime) : e.frames, o = e.state === "runnable" || e.state === "blocked" || !!t;
	return /* @__PURE__ */ (0, V.jsxs)("details", {
		className: "border-0 bg-transparent",
		open: o,
		children: [/* @__PURE__ */ (0, V.jsx)("summary", {
			className: "cursor-pointer list-none px-0 py-1",
			children: /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex items-center gap-2 flex-wrap",
				children: [
					/* @__PURE__ */ (0, V.jsxs)("span", {
						className: "font-mono text-xs font-semibold text-foreground",
						children: ["#", e.id]
					}),
					/* @__PURE__ */ (0, V.jsx)("span", {
						className: "truncate font-mono text-xs text-foreground",
						children: e.name
					}),
					/* @__PURE__ */ (0, V.jsxs)("span", {
						className: `inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] ${Sd(e.state)}`,
						children: [/* @__PURE__ */ (0, V.jsx)("span", { className: `h-2 w-2 rounded-full ${Cd(e.state)}` }), e.rawState || e.state]
					}),
					e.daemon && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "rounded-full bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground",
						children: "daemon"
					}),
					/* @__PURE__ */ (0, V.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [a.length, "f"]
					}),
					e.userFrameCount > 0 && /* @__PURE__ */ (0, V.jsxs)("span", {
						className: "text-[11px] text-muted-foreground",
						children: [e.userFrameCount, "u"]
					}),
					e.topFunction && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "truncate text-[11px] text-muted-foreground",
						children: e.topFunction
					})
				]
			})
		}), /* @__PURE__ */ (0, V.jsx)(Ed, {
			frames: a,
			className: "space-y-0.5 py-1 pl-3",
			...r ? { resolveSource: r } : {},
			...i ? { frameActions: i } : {}
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/ThreadStateFilterChips.tsx
function Od({ counts: e, selected: t, onToggle: n, dot: r, className: a }) {
	return /* @__PURE__ */ (0, V.jsx)("div", {
		className: i("flex flex-wrap items-center gap-1", a),
		children: Array.from(e.entries()).sort((e, t) => t[1] - e[1]).map(([e, i]) => {
			let a = t.has(e);
			return /* @__PURE__ */ (0, V.jsxs)("button", {
				type: "button",
				"aria-pressed": a,
				className: `inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] transition-colors ${a ? "border-primary bg-primary/10 text-primary" : "border-border bg-muted/50 text-muted-foreground hover:bg-background"}`,
				onClick: () => n(e),
				children: [
					/* @__PURE__ */ (0, V.jsx)("span", { className: `h-2 w-2 rounded-full ${r(e)}` }),
					e,
					/* @__PURE__ */ (0, V.jsx)("span", {
						className: "text-[10px] opacity-70",
						children: i
					})
				]
			}, e);
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/StackDumpView.tsx
var kd = "10rem", Ad = {
	format: "unknown",
	goroutines: [],
	threads: []
};
function jd(e) {
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
			goroutines: ld(r)
		} : Ad,
		text: i.map((e) => e.raw).join("\n"),
		...t.length > 0 ? { error: t.map((e) => e.error).join("\n") } : {},
		...a ? {
			collectedAt: a.takenAt,
			reason: a.reason
		} : {}
	};
}
function Md(e, t) {
	if (t) return jd(t);
	if (e === void 0) throw Error("StackDumpView needs either text or records");
	return {
		parsed: td(e),
		text: e
	};
}
function Nd(e) {
	return e.format === "jvm" ? e.threads : e.format === "go" ? e.goroutines : [];
}
function Pd(e) {
	return e.format === "jvm" ? "threads" : e.format === "go" ? "goroutines" : "frames";
}
function Fd(e, t) {
	return e.format === "jvm" ? Cd(t) : xd(t);
}
function Id({ status: e }) {
	return /* @__PURE__ */ (0, V.jsx)("span", {
		className: `rounded-full px-2 py-0.5 ${e === "ready" ? "bg-green-100 text-green-700 dark:bg-green-500/20 dark:text-green-300" : e === "unsupported" ? "bg-yellow-100 text-yellow-700 dark:bg-yellow-500/20 dark:text-yellow-300" : "bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300"}`,
		children: e
	});
}
function Ld({ text: e, records: t, error: n, status: r, collectedAt: i, reason: a, singleItem: o = !1, frameActions: s, resolveSource: l, headerActions: u }) {
	let [d, f] = (0, B.useState)(""), [p, m] = (0, B.useState)(/* @__PURE__ */ new Set()), [h, g] = (0, B.useState)(!0), _ = (0, B.useMemo)(() => Md(e, t), [e, t]), { parsed: v, text: y } = _, b = n ?? _.error, x = i ?? _.collectedAt, S = a ?? _.reason, C = t ? `${t[0]?.checkpointId}|${t.length}` : e, w = (0, B.useMemo)(() => nd(v), [v]), T = (0, B.useMemo)(() => Nd(v), [v]), E = (0, B.useMemo)(() => {
		if (o) return T;
		let e = d.trim().toLowerCase();
		return T.filter((t) => !(p.size > 0 && !p.has(t.state) || h && t.userFrameCount === 0 || e && !t.searchText.includes(e)));
	}, [
		T,
		d,
		p,
		h,
		o
	]);
	(0, B.useEffect)(() => {
		f(""), m(/* @__PURE__ */ new Set()), g(!0);
	}, [C]);
	let D = T.length, O = Pd(v);
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		"data-stack-format": v.format,
		children: [
			!o && /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "space-y-density-2 px-1 py-1.5",
				children: [/* @__PURE__ */ (0, V.jsxs)("div", {
					className: "flex flex-wrap items-center justify-between gap-density-2 text-[11px] text-muted-foreground",
					children: [/* @__PURE__ */ (0, V.jsxs)("div", {
						className: "flex flex-wrap items-center gap-density-2",
						children: [
							r ? /* @__PURE__ */ (0, V.jsx)(Id, { status: r }) : null,
							x ? /* @__PURE__ */ (0, V.jsx)("span", { children: new Date(x).toLocaleString() }) : null,
							S ? /* @__PURE__ */ (0, V.jsx)("span", {
								"data-stack-reason": "",
								children: S
							}) : null,
							D > 0 && /* @__PURE__ */ (0, V.jsxs)("span", { children: [
								E.length,
								" / ",
								D,
								" ",
								O
							] })
						]
					}), u]
				}), D > 0 && /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsxs)("div", {
					className: "flex flex-wrap items-center gap-1.5",
					children: [
						/* @__PURE__ */ (0, V.jsxs)("div", {
							className: "relative min-w-[14rem] flex-1",
							children: [/* @__PURE__ */ (0, V.jsx)(c, {
								icon: We,
								className: "absolute left-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground"
							}), /* @__PURE__ */ (0, V.jsx)("input", {
								className: "w-full rounded-md border border-border bg-muted/50 py-1 pl-7 pr-2 text-xs outline-none focus:border-primary focus:bg-background",
								"aria-label": `Filter ${O}`,
								placeholder: v.format === "jvm" ? "Filter by thread name, function, or file" : "Filter by goroutine id, function, or file",
								value: d,
								onChange: (e) => f(e.target.value)
							})]
						}),
						/* @__PURE__ */ (0, V.jsxs)("label", {
							className: "inline-flex items-center gap-1.5 rounded-md border border-border bg-muted/50 px-2 py-1 text-[11px] text-muted-foreground",
							children: [/* @__PURE__ */ (0, V.jsx)("input", {
								type: "checkbox",
								checked: h,
								onChange: (e) => g(e.target.checked)
							}), "Hide runtime-only"]
						}),
						(d || p.size > 0 || !h) && /* @__PURE__ */ (0, V.jsx)("button", {
							type: "button",
							className: "text-[11px] text-muted-foreground hover:text-foreground",
							onClick: () => {
								f(""), m(/* @__PURE__ */ new Set()), g(!0);
							},
							children: "Clear"
						})
					]
				}), /* @__PURE__ */ (0, V.jsx)(Od, {
					counts: w,
					selected: p,
					dot: (e) => Fd(v, e),
					onToggle: (e) => {
						m((t) => {
							let n = new Set(t);
							return n.has(e) ? n.delete(e) : n.add(e), n;
						});
					}
				})] })]
			}),
			o && u ? /* @__PURE__ */ (0, V.jsx)("div", {
				className: "flex justify-end px-1 py-1.5",
				children: u
			}) : null,
			b && /* @__PURE__ */ (0, V.jsx)("div", {
				className: "mt-density-2 whitespace-pre-wrap text-xs text-red-600",
				children: b
			}),
			D === 0 ? y ? /* @__PURE__ */ (0, V.jsx)("pre", {
				className: "min-h-0 flex-1 overflow-auto whitespace-pre-wrap py-1 font-mono text-[11px] leading-4 text-foreground",
				style: { minHeight: kd },
				children: y
			}) : null : /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "min-h-0 flex-1 space-y-1 overflow-auto py-1",
				style: { minHeight: kd },
				children: [
					E.length === 0 && /* @__PURE__ */ (0, V.jsxs)("div", {
						className: "py-density-3 text-center text-xs text-muted-foreground",
						children: [
							"No ",
							O,
							" match the current filters."
						]
					}),
					v.format === "jvm" && v.threads.filter((e) => E.includes(e)).map((e) => /* @__PURE__ */ (0, V.jsx)(Dd, {
						thread: e,
						search: d,
						hideRuntimeOnly: h,
						...l ? { resolveSource: l } : {},
						...s ? { frameActions: s } : {}
					}, e.id)),
					v.format === "go" && v.goroutines.filter((e) => E.includes(e)).map((e) => /* @__PURE__ */ (0, V.jsx)(wd, {
						goroutine: e,
						search: d,
						hideRuntimeOnly: h
					}, e.id))
				]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/Expectation.tsx
function Rd(e) {
	return typeof e == "string" ? e : JSON.stringify(e);
}
function zd(e) {
	return Array.isArray(e) ? e.some((e) => typeof e == "object" && !!e) : typeof e == "object" && !!e;
}
function Bd({ value: e, tree: t = !1, className: n }) {
	return t && zd(e) ? /* @__PURE__ */ (0, V.jsx)(et, { data: e }) : /* @__PURE__ */ (0, V.jsx)("code", {
		className: i("whitespace-pre-wrap break-words font-mono", n),
		children: Rd(e)
	});
}
function Vd({ label: e, children: t }) {
	return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsx)("dt", {
		className: "text-muted-foreground",
		children: e
	}), /* @__PURE__ */ (0, V.jsx)("dd", {
		className: "min-w-0",
		children: t
	})] });
}
function Hd({ check: e, checkLabel: t = "matcher", expected: n, actual: r, vars: a, location: o, className: s }) {
	let c = Object.entries(a ?? {});
	return /* @__PURE__ */ (0, V.jsxs)("dl", {
		className: i("grid grid-cols-[auto_minmax(0,1fr)] items-baseline gap-x-3 gap-y-0.5 rounded border border-border bg-muted/30 px-2 py-1.5 text-[11px]", s),
		"data-expectation": "",
		children: [
			e && /* @__PURE__ */ (0, V.jsx)(Vd, {
				label: t,
				children: /* @__PURE__ */ (0, V.jsx)(Bd, {
					value: e,
					className: e.includes("\n") ? "block overflow-x-auto whitespace-pre break-normal" : ""
				})
			}),
			c.map(([e, t]) => /* @__PURE__ */ (0, V.jsx)(Vd, {
				label: e,
				children: /* @__PURE__ */ (0, V.jsx)(Bd, {
					value: t,
					tree: !0
				})
			}, e)),
			n !== void 0 && /* @__PURE__ */ (0, V.jsx)(Vd, {
				label: "expected",
				children: /* @__PURE__ */ (0, V.jsx)(Bd, {
					value: n,
					className: "text-emerald-700 dark:text-emerald-400"
				})
			}),
			r !== void 0 && /* @__PURE__ */ (0, V.jsx)(Vd, {
				label: "actual",
				children: /* @__PURE__ */ (0, V.jsx)(Bd, {
					value: r,
					className: "text-red-700 dark:text-red-400"
				})
			}),
			o && /* @__PURE__ */ (0, V.jsx)(Vd, {
				label: "at",
				children: /* @__PURE__ */ (0, V.jsx)(Bd, {
					value: o,
					className: "text-muted-foreground"
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestFailureDetail.tsx
function Ud({ stack: e }) {
	let t = (0, B.useMemo)(() => td(e), [e]);
	return t.format === "go" ? /* @__PURE__ */ (0, V.jsx)("div", {
		className: "flex max-h-[32rem] min-h-0 flex-col",
		"data-failure-stack": "go",
		children: /* @__PURE__ */ (0, V.jsx)(Ld, {
			text: e,
			singleItem: t.goroutines.length === 1
		})
	}) : /* @__PURE__ */ (0, V.jsx)(uo, {
		logs: e,
		collapsedLines: 6
	});
}
function Wd({ failure: e }) {
	let t = !!(e.matcher || e.expected || e.actual || e.location);
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "space-y-2",
		children: [
			e.summary && /* @__PURE__ */ (0, V.jsx)("p", {
				className: "text-sm font-medium text-red-700 dark:text-red-300",
				children: e.summary
			}),
			t && /* @__PURE__ */ (0, V.jsx)(Hd, {
				check: e.matcher,
				expected: e.expected || void 0,
				actual: e.actual || void 0,
				location: e.location
			}),
			e.stack && /* @__PURE__ */ (0, V.jsx)(Ud, { stack: e.stack })
		]
	});
}
var Gd = 10;
function Kd({ node: e, view: t = "preview" }) {
	let n = Pu(), r = (e.log ? [e.log] : [e.stdout, e.stderr]).flatMap((e) => e ? [e.replace(/\r?\n$/, "")] : []).join("\n");
	if (!r) return null;
	let i = t === "preview";
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "space-y-1",
		children: [
			i && /* @__PURE__ */ (0, V.jsx)("div", {
				className: "text-[10px] uppercase tracking-wide text-muted-foreground",
				children: "Output"
			}),
			/* @__PURE__ */ (0, V.jsx)(uo, {
				logs: r,
				variant: "terminal",
				maxHeight: "none",
				collapsedLines: Infinity,
				...i ? { tailLines: Gd } : {}
			}),
			i && r.split("\n").length > Gd && n && /* @__PURE__ */ (0, V.jsx)(o, {
				type: "button",
				variant: "link",
				size: "sm",
				className: "px-0",
				onClick: () => n("stdout"),
				children: "Show more in Stdout"
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/metric-tone.ts
function qd(e) {
	return e >= 90 ? "danger" : e >= 60 ? "warning" : e > 0 ? "success" : "neutral";
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/process-metrics.ts
var Jd = class extends Error {
	pid;
	constructor(e) {
		super(`process ${e} not found`), this.name = "ProcessNotFoundError", this.pid = e;
	}
}, Yd = 2e3;
function Xd(e, t, n) {
	let [r, i] = (0, B.useState)({}), [a, o] = (0, B.useState)(!1), s = (0, B.useRef)(n);
	s.current = n;
	let c = t && e !== void 0 && e > 0 && n !== void 0;
	return (0, B.useEffect)(() => {
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
				if (e instanceof Jd) {
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
			n = setTimeout(a, Yd);
		};
		return a(), () => {
			t = !0, clearTimeout(n), o(!1);
		};
	}, [c, e]), {
		state: r,
		polling: a
	};
}
function Zd({ attempt: e, fetchProcessMetrics: t }) {
	let n = gu(e), [r, i] = (0, B.useState)(Date.now);
	(0, B.useEffect)(() => {
		if (!n || !e.started) return;
		let t = setInterval(() => i(Date.now()), 1e3);
		return () => clearInterval(t);
	}, [n, e.started]);
	let a = n && e.started ? Math.max(0, r - Date.parse(e.started)) * 1e6 : e.duration, { state: o, polling: s } = Xd(e.pid, n, t), c = {
		...e.pid === void 0 ? {} : { pid: e.pid },
		...e.cpu_percent === void 0 ? {} : { cpu_percent: e.cpu_percent },
		...e.rss === void 0 ? {} : { rss: e.rss }
	}, l = n && o.metrics ? o.metrics : c, u = l.pid ?? e.pid, d = n ? o.note === "process exited" ? "exited" : l.status ?? "running" : "exited", f = (u ?? 0) > 0 || l.cpu_percent !== void 0 || l.rss !== void 0;
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "space-y-3",
		"data-attempt-sequence": e.sequence,
		children: [/* @__PURE__ */ (0, V.jsxs)("section", { children: [
			/* @__PURE__ */ (0, V.jsxs)("div", {
				className: "mb-1 flex items-center gap-2",
				children: [
					/* @__PURE__ */ (0, V.jsx)("div", {
						className: "text-[10px] uppercase tracking-wide text-muted-foreground",
						children: "Process metrics"
					}),
					s && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "size-1.5 animate-pulse rounded-full bg-blue-500",
						role: "status",
						"aria-label": "Reading live metrics"
					}),
					e.shared_process && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground",
						title: "The pid, CPU and RSS belong to the go test batch this package ran in, not to the package alone",
						"data-shared-process": "",
						children: "shared by batch"
					})
				]
			}),
			f ? /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "grid grid-cols-2 gap-density-1 md:grid-cols-4",
				children: [
					u !== void 0 && u > 0 && /* @__PURE__ */ (0, V.jsx)(eu, {
						label: "PID",
						value: String(u)
					}),
					/* @__PURE__ */ (0, V.jsx)(eu, {
						label: "Status",
						value: d
					}),
					/* @__PURE__ */ (0, V.jsx)(eu, {
						label: "CPU",
						value: l.cpu_percent === void 0 ? "—" : `${l.cpu_percent.toFixed(1)}%`,
						...l.cpu_percent === void 0 ? {} : {
							barPct: Math.min(100, l.cpu_percent),
							tone: qd(l.cpu_percent)
						}
					}),
					/* @__PURE__ */ (0, V.jsx)(eu, {
						label: "RSS",
						value: l.rss !== void 0 && l.rss > 0 ? nt(l.rss) : "—"
					}),
					l.open_files !== void 0 && /* @__PURE__ */ (0, V.jsx)(eu, {
						label: "FDs",
						value: String(l.open_files)
					}),
					a !== void 0 && a > 0 && /* @__PURE__ */ (0, V.jsx)(eu, {
						label: "Duration",
						value: L(a)
					}),
					e.exit_code !== void 0 && /* @__PURE__ */ (0, V.jsx)(eu, {
						label: "Exit",
						value: String(e.exit_code),
						tone: e.exit_code === 0 ? "success" : "danger"
					}),
					e.goroutine_count !== void 0 && /* @__PURE__ */ (0, V.jsx)(eu, {
						label: "Goroutines",
						value: String(e.goroutine_count)
					})
				]
			}) : /* @__PURE__ */ (0, V.jsx)("div", {
				className: "text-xs italic text-muted-foreground",
				children: "No metrics captured for this attempt."
			}),
			o.note && n && /* @__PURE__ */ (0, V.jsx)("div", {
				className: "mt-1 text-[11px] italic text-muted-foreground",
				children: o.note
			})
		] }), e.stack_trace && /* @__PURE__ */ (0, V.jsxs)("section", { children: [/* @__PURE__ */ (0, V.jsx)("div", {
			className: "mb-1 text-[10px] uppercase tracking-wide text-muted-foreground",
			children: "Stack trace"
		}), /* @__PURE__ */ (0, V.jsx)("div", {
			className: "flex max-h-[32rem] min-h-0 flex-col",
			children: /* @__PURE__ */ (0, V.jsx)(Ld, { text: e.stack_trace })
		})] })]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestViewToggle.tsx
var Qd = [{
	id: "tree",
	label: "Tree",
	icon: Ce,
	title: "Group tests in a tree beside their detail"
}, {
	id: "list",
	label: "List",
	icon: Se,
	title: "One row per test, its detail opening beneath it"
}];
function $d() {
	let e = Cu();
	return e.onViewChange ? /* @__PURE__ */ (0, V.jsx)(C, {
		size: "sm",
		"aria-label": "View",
		value: e.view ?? "tree",
		options: Qd,
		onChange: e.onViewChange
	}) : null;
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestDetailPanel.tsx
function ef({ node: e }) {
	let t = Cu(), n = e.attempts?.at(-1), r = (e.children?.length ?? 0) > 0 ? it(e) : null;
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "space-y-4 p-density-4",
		children: [
			e.message && !e.failure_detail && /* @__PURE__ */ (0, V.jsx)("p", {
				className: "whitespace-pre-wrap break-words text-sm",
				children: e.message
			}),
			e.failure_detail && /* @__PURE__ */ (0, V.jsx)(Wd, { failure: e.failure_detail }),
			/* @__PURE__ */ (0, V.jsx)(Kd, { node: e }),
			n && /* @__PURE__ */ (0, V.jsx)(Zd, {
				attempt: n,
				...t.fetchProcessMetrics ? { fetchProcessMetrics: t.fetchProcessMetrics } : {}
			}, `${n.run_kind ?? ""}:${n.sequence}`),
			r && r.total > 0 && /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex flex-wrap gap-3 text-xs text-muted-foreground",
				children: [
					/* @__PURE__ */ (0, V.jsxs)("span", { children: [r.total, " total"] }),
					r.passed > 0 && /* @__PURE__ */ (0, V.jsxs)("span", { children: [r.passed, " passed"] }),
					r.failed > 0 && /* @__PURE__ */ (0, V.jsxs)("span", { children: [r.failed, " failed"] }),
					r.skipped > 0 && /* @__PURE__ */ (0, V.jsxs)("span", { children: [r.skipped, " skipped"] })
				]
			}),
			e.detail != null && /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "space-y-1",
				children: [/* @__PURE__ */ (0, V.jsx)("div", {
					className: "text-[10px] uppercase tracking-wide text-muted-foreground",
					children: "Detail"
				}), /* @__PURE__ */ (0, V.jsx)(et, { data: e.detail })]
			})
		]
	});
}
function tf(e, t) {
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
var nf = {
	idle: "Copy details",
	pending: "Copying…",
	copied: "Copied",
	error: "Copy failed"
}, rf = {
	idle: _,
	pending: Tn,
	copied: w,
	error: Re
};
function af({ node: e }) {
	let t = Cu(), { state: n, copy: r } = pe(), i = (0, B.useMemo)(() => t.adapters.resolveCopy(e, t), [e, t]), a = [{
		label: "Copy as JSON",
		icon: gn,
		onSelect: () => r(mu(e))
	}, ...i.items.map((e) => tf(e, r))];
	return /* @__PURE__ */ (0, V.jsx)(se, {
		label: nf[n],
		icon: rf[n],
		variant: "outline",
		size: "sm",
		align: "right",
		title: "More copy options",
		onClick: () => r(pu(e, i)),
		items: a
	});
}
function of({ node: e, actionNode: t, actions: n, attemptChooser: r }) {
	let i = Cu(), a = !!i.onRerun && t.framework !== "task", s = !!i.onStop && !!t.can_stop && !!t.task_id;
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "ml-auto flex shrink-0 items-center gap-1",
		"data-detail-actions": "",
		children: [
			r,
			n,
			i.view !== "list" && /* @__PURE__ */ (0, V.jsx)($d, {}),
			/* @__PURE__ */ (0, V.jsx)(af, { node: e }),
			a && /* @__PURE__ */ (0, V.jsxs)(o, {
				size: "sm",
				disabled: i.busy.rerun,
				onClick: () => i.onRerun(t),
				title: "Rerun this test or subtree",
				children: [/* @__PURE__ */ (0, V.jsx)(c, {
					icon: wn,
					className: "mr-1 text-sm"
				}), i.busy.rerun ? "Running..." : "Rerun"]
			}),
			s && /* @__PURE__ */ (0, V.jsxs)(o, {
				size: "sm",
				variant: "destructive",
				disabled: i.busy.stop,
				onClick: () => i.onStop(t),
				title: "Stop this running task",
				children: [/* @__PURE__ */ (0, V.jsx)(c, {
					icon: Wt,
					className: "mr-1 text-sm"
				}), i.busy.stop ? "Stopping..." : "Stop"]
			})
		]
	});
}
function sf(e) {
	let { node: t } = e, n = Cu(), r = kt(t), i = t.file ? `${t.file}${t.line ? `:${t.line}` : ""}` : void 0, a = n.adapters.resolveFrameworkBadge(t, n);
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "flex min-w-0 items-center gap-2 border-b border-border px-density-4 py-density-2",
		"data-detail-header": "",
		children: [
			r ? /* @__PURE__ */ (0, V.jsx)(Et, {
				outcome: r,
				className: "text-lg"
			}) : /* @__PURE__ */ (0, V.jsx)(c, {
				icon: hn,
				className: "shrink-0 text-lg text-muted-foreground"
			}),
			/* @__PURE__ */ (0, V.jsx)("h2", {
				className: "min-w-0 shrink truncate text-base font-semibold",
				title: t.name,
				children: t.name
			}),
			t.framework && /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "inline-flex shrink-0 items-center gap-1 rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground",
				"data-framework-badge": "",
				children: [
					/* @__PURE__ */ (0, V.jsx)(Tt, {
						tool: Vt(t),
						className: "text-sm"
					}),
					t.framework,
					a != null && /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsx)("span", {
						"aria-hidden": "true",
						children: "·"
					}), a] })
				]
			}),
			t.duration ? /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "inline-flex shrink-0 items-center gap-1 text-xs text-muted-foreground",
				children: [/* @__PURE__ */ (0, V.jsx)(c, {
					icon: P,
					className: "text-xs"
				}), L(t.duration)]
			}) : null,
			i && /* @__PURE__ */ (0, V.jsx)("span", {
				className: "min-w-0 shrink-[100] truncate font-mono text-xs text-muted-foreground",
				title: i,
				children: i
			}),
			t.file && n.onOpenFile && /* @__PURE__ */ (0, V.jsx)(tu, {
				className: "shrink-0",
				onOpenFile: n.onOpenFile,
				target: {
					file: wt(t.file, t.work_dir),
					...t.work_dir ? { workDir: t.work_dir } : {},
					...t.line ? { line: t.line } : {}
				}
			}),
			/* @__PURE__ */ (0, V.jsx)(of, { ...e })
		]
	});
}
function cf({ header: e = "full", tabVariant: t = "pill" } = {}) {
	let n = Cu(), r = n.selected, [a, o] = (0, B.useState)(null), s = n.activeTab !== void 0, [l, u] = (0, B.useState)("detail"), d = s ? n.activeTab : l, f = (e) => {
		s ? n.onTabChange?.(e) : u(e);
	};
	if (!r) return /* @__PURE__ */ (0, V.jsx)("div", {
		className: "flex h-full items-center justify-center text-sm text-muted-foreground",
		children: /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "text-center",
			children: [/* @__PURE__ */ (0, V.jsx)(c, {
				icon: hn,
				className: "mb-2 block text-4xl"
			}), "Select a test to view its details."]
		})
	});
	let p = n.adapters.resolveAttempts(r, n), m = r.task_id ?? r.name, h = n.activeAttempt ?? (a?.key === m ? a.id : void 0), g = p.find((e) => e.id === h) ?? p.at(-1), _ = p.length > 1 && g ? g.node : r, v = p.length > 1 && g ? /* @__PURE__ */ (0, V.jsx)("div", {
		className: "w-44 shrink-0",
		children: /* @__PURE__ */ (0, V.jsx)(Qe, {
			"aria-label": "Attempt",
			value: String(g.id),
			className: "h-7 text-xs",
			onChange: (e) => {
				let t = Number(e.target.value);
				n.onAttemptChange ? n.onAttemptChange(t) : o({
					key: m,
					id: t
				});
			},
			options: p.map((e) => ({
				value: String(e.id),
				label: `Attempt ${e.id}/${p.length} · ${e.status}${e.durationMs === void 0 ? "" : ` · ${L(e.durationMs * 1e6)}`}`
			}))
		})
	}) : null, y = n.adapters.resolveDetail(_, n), b = n.adapters.resolveActions(r, n), x = n.adapters.resolveTabs(_, n);
	(_.log || _.stdout || _.stderr) && !x.some((e) => e.id === "stdout") && x.unshift({
		id: "stdout",
		label: "Stdout",
		render: () => /* @__PURE__ */ (0, V.jsx)("div", {
			className: "p-density-4",
			children: /* @__PURE__ */ (0, V.jsx)(Kd, {
				node: _,
				view: "full"
			})
		})
	});
	let S = n.adapters.ownsScroll(_, n), C = /* @__PURE__ */ (0, V.jsx)(Nu, {
		value: f,
		children: y?.renderDetail ? y.renderDetail({
			node: _,
			runner: n
		}) : /* @__PURE__ */ (0, V.jsx)(ef, { node: _ })
	}), w = b?.nodeActions?.({
		node: r,
		runner: n
	}), T = i("min-h-0 flex-1 [container-type:size]", !S && "overflow-y-auto"), E = {
		node: _,
		actionNode: r,
		actions: w,
		attemptChooser: v
	}, D = e === "full";
	if (x.length === 0) return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		children: [D ? /* @__PURE__ */ (0, V.jsx)(sf, { ...E }) : /* @__PURE__ */ (0, V.jsx)("div", {
			className: "flex shrink-0 border-b border-border px-density-2 py-density-1",
			children: /* @__PURE__ */ (0, V.jsx)(of, { ...E })
		}), /* @__PURE__ */ (0, V.jsx)("div", {
			className: T,
			children: C
		})]
	});
	let O = {
		id: "detail",
		label: "Detail",
		render: () => C
	}, ee = [O, ...x], te = ee.find((e) => e.id === d) ?? O;
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		children: [
			D && /* @__PURE__ */ (0, V.jsx)(sf, { ...E }),
			/* @__PURE__ */ (0, V.jsxs)("div", {
				className: i("flex shrink-0 items-center gap-1 border-b border-border px-density-2", t === "underline" ? "items-stretch" : "py-density-1"),
				children: [ee.map((e) => /* @__PURE__ */ (0, V.jsx)(ae, {
					active: te.id === e.id,
					onClick: () => f(e.id),
					label: e.label,
					variant: t
				}, e.id)), !D && /* @__PURE__ */ (0, V.jsx)(of, { ...E })]
			}),
			/* @__PURE__ */ (0, V.jsx)("div", {
				className: T,
				children: te.render()
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestList.tsx
function lf(e, t) {
	return t ? e.route_path && t.route_path ? e.route_path === t.route_path : e === t : !1;
}
function uf({ test: e }) {
	let t = Cu(), n = lf(e, t.selected), r = kt(e), a = [e.package, ...e.suite ?? []].filter(Boolean).join(" › "), o = e.message?.split("\n").find((e) => e.trim()) ?? "";
	return /* @__PURE__ */ (0, V.jsxs)("li", {
		className: "border-b border-border last:border-b-0",
		"data-test-list-row": "",
		children: [/* @__PURE__ */ (0, V.jsxs)("button", {
			type: "button",
			"aria-expanded": n,
			onClick: () => t.onSelect(n ? null : e),
			className: i("flex w-full min-w-0 items-start gap-2 px-density-3 py-density-2 text-left hover:bg-accent/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", n && "bg-accent/40"),
			children: [
				/* @__PURE__ */ (0, V.jsx)(c, {
					icon: n ? k : j,
					className: "mt-1 shrink-0 text-xs text-muted-foreground"
				}),
				r && /* @__PURE__ */ (0, V.jsx)(Et, {
					outcome: r,
					className: "mt-0.5 shrink-0"
				}),
				/* @__PURE__ */ (0, V.jsxs)("span", {
					className: "min-w-0 flex-1",
					children: [/* @__PURE__ */ (0, V.jsxs)("span", {
						className: "flex min-w-0 items-baseline gap-2",
						children: [/* @__PURE__ */ (0, V.jsx)("span", {
							className: "min-w-0 truncate text-sm font-medium",
							title: e.name,
							children: e.name
						}), a && /* @__PURE__ */ (0, V.jsx)("span", {
							className: "min-w-0 shrink-[100] truncate text-xs text-muted-foreground",
							title: a,
							children: a
						})]
					}), o && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "block truncate font-mono text-xs text-muted-foreground",
						title: o,
						children: o
					})]
				})
			]
		}), n && /* @__PURE__ */ (0, V.jsx)("div", {
			className: "h-96 border-t border-border",
			"data-test-list-detail": "",
			children: /* @__PURE__ */ (0, V.jsx)(cf, { header: "actions" })
		})]
	});
}
function df({ className: e }) {
	let t = Cu(), n = (0, B.useMemo)(() => Ot(t.tests), [t.tests]);
	return n.length === 0 ? /* @__PURE__ */ (0, V.jsx)("div", {
		className: "p-8 text-center text-sm text-muted-foreground",
		children: "No tests match the current filters"
	}) : /* @__PURE__ */ (0, V.jsx)("ul", {
		className: i("m-0 list-none p-0", e),
		"data-test-list": "",
		children: n.map((e, t) => /* @__PURE__ */ (0, V.jsx)(uf, { test: e }, e.route_path ?? `${t}:${e.name}`))
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestRunSummary.tsx
function ff(e, t) {
	return t > 0 ? Math.round(e / t * 100) : 0;
}
function pf({ tests: e, startTime: t, endTime: n, done: r, now: a, runMeta: o, compact: s = !1, className: l }) {
	let u = gt(e), d = r && n ? n : a ?? 0, f = t && d >= t ? `${((d - t) / 1e3).toFixed(1)}s` : "", p = [
		{
			count: u.passed,
			color: "bg-green-500",
			label: "passed"
		},
		{
			count: u.skipped,
			color: "bg-slate-400",
			label: "skipped"
		},
		{
			count: u.failed,
			color: "bg-red-500",
			label: "failed"
		},
		{
			count: u.warned,
			color: "bg-amber-400",
			label: "warned"
		},
		{
			count: u.aborted,
			color: "bg-slate-400",
			label: "aborted"
		},
		{
			count: u.timedout,
			color: "bg-amber-500",
			label: "timed out"
		},
		{
			count: u.running,
			color: "bg-blue-400",
			label: "running"
		},
		{
			count: u.pending,
			color: "bg-muted-foreground/40",
			label: "queued"
		}
	], m = [
		{
			n: u.passed,
			label: "passed",
			outcome: "passed"
		},
		{
			n: u.failed,
			label: "failed",
			outcome: "failed"
		},
		{
			n: u.warned,
			label: "warned",
			outcome: "warned"
		},
		{
			n: u.aborted,
			label: "aborted",
			outcome: "aborted"
		},
		{
			n: u.timedout,
			label: "timed out",
			outcome: "timedout"
		},
		{
			n: u.skipped,
			label: "skipped",
			outcome: "skipped"
		},
		{
			n: u.running,
			label: "running",
			outcome: "running"
		},
		{
			n: u.pending,
			label: "queued",
			outcome: "pending"
		}
	].filter((e) => e.n > 0).map((e) => /* @__PURE__ */ (0, V.jsxs)("span", {
		className: i("inline-flex items-center gap-0.5 tabular-nums", zt[e.outcome].className),
		"aria-label": `${e.n} ${e.label}`,
		title: `${e.n} ${e.label}`,
		"data-summary-outcome": e.outcome,
		children: [/* @__PURE__ */ (0, V.jsx)(Et, { outcome: e.outcome }), e.n]
	}, e.label)), h = /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [o?.timed_out && /* @__PURE__ */ (0, V.jsxs)("span", {
		className: i("inline-flex items-center gap-0.5", zt.timedout.className),
		title: "The run hit its timeout",
		"data-summary-run": "timed-out",
		children: [/* @__PURE__ */ (0, V.jsx)(Et, { outcome: "timedout" }), "Timed out"]
	}), o?.exit_code !== void 0 && /* @__PURE__ */ (0, V.jsxs)("span", {
		className: i("inline-flex items-center tabular-nums", o.exit_code === 0 ? "text-muted-foreground" : zt.failed.className),
		title: `Exit code ${o.exit_code}`,
		"data-summary-run": "exit-code",
		children: ["exit ", o.exit_code]
	})] });
	return s ? /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i("flex flex-wrap items-center gap-x-3 gap-y-0.5 text-sm text-muted-foreground", l),
		children: [
			/* @__PURE__ */ (0, V.jsxs)("span", {
				className: "font-medium text-foreground",
				children: [
					u.total,
					" ",
					u.total === 1 ? "test" : "tests",
					o?.kind === "rerun" ? ` · rerun #${o.sequence}` : ""
				]
			}),
			m,
			h,
			f && /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "inline-flex items-center gap-1",
				children: [/* @__PURE__ */ (0, V.jsx)(c, {
					icon: P,
					className: "text-xs"
				}), f]
			}),
			u.total > 0 && /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "inline-flex items-center gap-1.5",
				children: [/* @__PURE__ */ (0, V.jsx)(oo, {
					className: "w-24",
					segments: p,
					total: u.total,
					height: "h-1.5"
				}), /* @__PURE__ */ (0, V.jsxs)("span", {
					className: "text-xs font-medium text-green-600 dark:text-green-400",
					children: [ff(u.passed, u.total), "%"]
				})]
			}),
			!r && /* @__PURE__ */ (0, V.jsx)(c, {
				icon: St,
				className: "animate-spin text-blue-500"
			})
		]
	}) : /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i("flex min-w-[18rem] flex-col items-end gap-1.5", l),
		children: [/* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex flex-wrap items-center justify-end gap-x-3 gap-y-0.5 text-sm text-muted-foreground",
			children: [
				/* @__PURE__ */ (0, V.jsxs)("span", {
					className: "font-medium text-foreground",
					children: [
						u.total,
						" ",
						u.total === 1 ? "test" : "tests",
						o?.kind === "rerun" ? ` · rerun #${o.sequence}` : ""
					]
				}),
				m,
				h,
				f && /* @__PURE__ */ (0, V.jsxs)("span", {
					className: "inline-flex items-center gap-1",
					children: [/* @__PURE__ */ (0, V.jsx)(c, {
						icon: P,
						className: "text-xs"
					}), f]
				}),
				!r && /* @__PURE__ */ (0, V.jsx)(c, {
					icon: St,
					className: "animate-spin text-blue-500"
				})
			]
		}), u.total > 0 && /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex w-72 items-center gap-2",
			children: [/* @__PURE__ */ (0, V.jsx)(oo, {
				className: "flex-1",
				segments: p,
				total: u.total,
				height: "h-2"
			}), /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "shrink-0 text-xs font-medium text-green-600 dark:text-green-400",
				children: [ff(u.passed, u.total), "%"]
			})]
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/FilterFacet.tsx
var mf = {
	autoSubmit: !1,
	compact: !1
};
function hf({ label: e, options: t, state: n, onChange: r }) {
	let a = Object.fromEntries(n), o = n.size > 0;
	return /* @__PURE__ */ (0, V.jsx)(ne, {
		align: "right",
		menuLabel: e,
		menuClassName: "p-2",
		trigger: /* @__PURE__ */ (0, V.jsxs)("span", {
			role: "button",
			tabIndex: 0,
			className: i("inline-flex h-6 items-center gap-0.5 rounded px-1.5 text-xs transition-colors", o ? "bg-muted font-medium text-foreground" : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"),
			children: [st(e, a), /* @__PURE__ */ (0, V.jsx)(c, {
				icon: k,
				className: "text-[10px] opacity-60"
			})]
		}),
		children: () => /* @__PURE__ */ (0, V.jsx)(ut.Provider, {
			value: mf,
			children: /* @__PURE__ */ (0, V.jsx)(yt, {
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
// ../../../clicky-ui/packages/ui/src/data/test-runner/TestFilterBar.tsx
function gf({ filters: e, onChange: t, counts: n, frameworks: r, className: a }) {
	let o = R.filter((e) => n[e.key] > 0).map((e) => ({
		value: e.key,
		label: e.label,
		icon: /* @__PURE__ */ (0, V.jsx)(Et, { outcome: e.key }),
		count: n[e.key]
	})), s = r.map((e) => ({
		value: e,
		label: xt(e)?.label ?? e,
		title: e,
		icon: /* @__PURE__ */ (0, V.jsx)(Tt, { tool: e })
	}));
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i("flex items-center gap-0.5", a),
		children: [/* @__PURE__ */ (0, V.jsx)(hf, {
			label: "Status",
			options: o,
			state: e.status,
			onChange: (n) => t({
				...e,
				status: n
			})
		}), r.length > 1 && /* @__PURE__ */ (0, V.jsx)(hf, {
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
function _f(e) {
	return e instanceof yu ? e : bu(e ?? []);
}
var vf = 640;
function yf({ tests: e, allTests: t, selected: n, filters: r, expandAll: a, done: o, runMeta: s, status: l, statusText: u, now: d, busy: f, activeTab: p, onTabChange: m, activeAttempt: h, onAttemptChange: g, onSelect: _, onFiltersChange: v, onExpandAllChange: y, onRerun: b, onStop: x, onOpenFile: C, fetchProcessMetrics: w, adapters: T, title: E, showSummary: D = !0, headerSlot: O, hideHeader: ee = !1, view: te = "tree", onViewChange: ne, defaultSplit: re = 45, startTime: k, endTime: A, className: ie }) {
	let ae = (0, B.useMemo)(() => _f(T), [T]), oe = (0, B.useMemo)(() => ({
		tests: e,
		selected: n,
		runMeta: s,
		status: l,
		done: o,
		now: d,
		filters: r,
		expandAll: a,
		busy: f ?? {},
		view: te,
		onViewChange: ne,
		activeTab: p,
		onTabChange: m,
		activeAttempt: h,
		onAttemptChange: g,
		onSelect: _,
		onFiltersChange: v,
		onExpandAllChange: y,
		onRerun: b,
		onStop: x,
		onOpenFile: C,
		fetchProcessMetrics: w,
		adapters: ae
	}), [
		e,
		n,
		s,
		l,
		o,
		d,
		r,
		a,
		f,
		te,
		ne,
		p,
		m,
		h,
		g,
		_,
		v,
		y,
		b,
		x,
		C,
		w,
		ae
	]), j = t ?? e, se = (0, B.useMemo)(() => Bt(j), [j]), ce = (0, B.useMemo)(() => gt(j), [j]), le = !ee && j.length > 0 ? /* @__PURE__ */ (0, V.jsx)(gf, {
		filters: r,
		onChange: v,
		counts: ce,
		frameworks: se
	}) : null, { ref: ue, wider: de } = _t(vf);
	return /* @__PURE__ */ (0, V.jsx)(Su, {
		value: oe,
		children: /* @__PURE__ */ (0, V.jsxs)("div", {
			ref: ue,
			className: i("flex h-full min-h-0 flex-col bg-background", ie),
			children: [!ee && /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex min-w-0 items-center gap-3 border-b border-border bg-card px-density-4 py-density-2",
				"data-runner-header": "",
				children: [
					/* @__PURE__ */ (0, V.jsxs)("div", {
						className: "flex min-w-0 items-center gap-3",
						children: [E === void 0 ? /* @__PURE__ */ (0, V.jsxs)("h1", {
							className: "inline-flex shrink-0 items-center gap-1.5 text-base font-bold",
							children: [/* @__PURE__ */ (0, V.jsx)(c, {
								icon: Ct,
								className: "text-primary"
							}), "Test Results"]
						}) : E, u && /* @__PURE__ */ (0, V.jsx)("span", {
							className: "truncate text-sm text-muted-foreground",
							children: u
						})]
					}),
					D && /* @__PURE__ */ (0, V.jsx)(pf, {
						tests: j,
						startTime: k,
						endTime: A,
						done: o,
						now: d,
						runMeta: s,
						compact: !0,
						className: "ml-auto flex-nowrap"
					}),
					O && /* @__PURE__ */ (0, V.jsx)("div", {
						className: i("flex shrink-0 items-center gap-1", !D && "ml-auto"),
						children: O
					})
				]
			}), te === "list" ? /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex min-h-0 flex-1 flex-col",
				children: [(le || ne) && /* @__PURE__ */ (0, V.jsxs)("div", {
					className: "flex min-w-0 shrink-0 flex-wrap items-center gap-2 border-b border-border bg-card px-density-3 py-density-1",
					"data-test-list-toolbar": "",
					children: [le, /* @__PURE__ */ (0, V.jsx)("div", {
						className: "ml-auto",
						children: /* @__PURE__ */ (0, V.jsx)($d, {})
					})]
				}), /* @__PURE__ */ (0, V.jsx)("div", {
					className: "min-h-0 flex-1 overflow-y-auto",
					children: /* @__PURE__ */ (0, V.jsx)(df, {})
				})]
			}) : de ? /* @__PURE__ */ (0, V.jsx)(ht, {
				className: "min-h-0 flex-1",
				defaultSplit: re,
				minLeft: 25,
				minRight: 30,
				leftClass: "border-r-2 border-border bg-muted/30",
				left: /* @__PURE__ */ (0, V.jsx)(ju, { actions: le }),
				right: /* @__PURE__ */ (0, V.jsx)(cf, {})
			}) : n ? /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex min-h-0 flex-1 flex-col",
				children: [/* @__PURE__ */ (0, V.jsxs)("button", {
					type: "button",
					onClick: () => _(null),
					className: "inline-flex items-center gap-1.5 border-b border-border px-density-3 py-density-2 text-left text-sm text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
					children: [/* @__PURE__ */ (0, V.jsx)(c, { icon: S }), "Back to results"]
				}), /* @__PURE__ */ (0, V.jsx)("div", {
					className: "min-h-0 flex-1 overflow-hidden",
					children: /* @__PURE__ */ (0, V.jsx)(cf, {})
				})]
			}) : /* @__PURE__ */ (0, V.jsx)("div", {
				className: "min-h-0 flex-1 overflow-y-auto",
				children: /* @__PURE__ */ (0, V.jsx)(ju, { actions: le })
			})]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/CopyButton.tsx
function bf({ value: e, label: t = "Copy", className: n, iconClassName: r }) {
	let { state: a, copy: o } = pe(), s = (0, B.useCallback)((t) => {
		t.stopPropagation(), o(typeof e == "function" ? e() : e);
	}, [o, e]);
	return /* @__PURE__ */ (0, V.jsx)(F, {
		icon: a === "copied" ? w : _,
		label: a === "copied" ? "Copied" : a === "error" ? "Copy failed" : t,
		onClick: s,
		className: n,
		iconClassName: i(a === "copied" && "text-emerald-600", a === "error" && "text-red-600", r)
	});
}
// ../../../clicky-ui/packages/ui/src/data/lint/ViolationLocations.tsx
var xf = {
	width: 240,
	height: 96
};
function Sf({ location: e, workDir: t, primary: n }) {
	return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [e.ref && /* @__PURE__ */ (0, V.jsxs)("span", {
		className: "inline-flex items-center gap-0.5 rounded bg-muted px-1 font-mono text-[10px] text-muted-foreground",
		title: `git ref ${e.ref}`,
		children: [/* @__PURE__ */ (0, V.jsx)(c, {
			icon: bn,
			className: "shrink-0"
		}), e.ref]
	}), /* @__PURE__ */ (0, V.jsx)("span", {
		className: i("break-all font-mono", n ? "text-foreground" : "text-muted-foreground"),
		children: jt(e, t)
	})] });
}
var Cf = "inline-flex max-w-full items-center gap-1 rounded border border-border bg-muted/40 px-1.5 py-0.5";
function wf({ entity: e, entityHref: t }) {
	let n = t?.(e), r = `${e.kind} ${e.id}`, a = /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsx)(c, {
		icon: fn,
		className: "shrink-0 text-muted-foreground"
	}), It(e)] });
	return n ? /* @__PURE__ */ (0, V.jsx)("a", {
		href: n,
		title: r,
		className: i(Cf, "text-primary hover:underline"),
		children: a
	}) : /* @__PURE__ */ (0, V.jsx)("span", {
		title: r,
		className: Cf,
		children: a
	});
}
function Tf({ target: e }) {
	let t = Mt(e);
	return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsxs)("a", {
		href: e.url,
		target: "_blank",
		rel: "noreferrer",
		title: t ? `${e.url}\n${t}` : e.url,
		className: "inline-flex min-w-0 items-center gap-1 break-all text-primary hover:underline",
		children: [/* @__PURE__ */ (0, V.jsx)(c, {
			icon: cn,
			className: "shrink-0"
		}), e.url]
	}), t && /* @__PURE__ */ (0, V.jsx)("code", {
		className: "break-all rounded bg-muted px-1 font-mono text-[11px]",
		children: t
	})] });
}
function Ef({ image: e }) {
	let t = Ft(e, xf), n = Dt(e);
	return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [/* @__PURE__ */ (0, V.jsx)("a", {
		href: e.src,
		target: "_blank",
		rel: "noreferrer",
		title: `Open ${e.src}`,
		className: "inline-block rounded border border-border bg-muted hover:border-primary",
		children: /* @__PURE__ */ (0, V.jsx)("div", {
			className: "relative overflow-hidden rounded",
			style: t.frame,
			children: /* @__PURE__ */ (0, V.jsx)("img", {
				src: e.src,
				alt: `Region ${n} of ${e.src}`,
				className: "max-w-none",
				style: t.image
			})
		})
	}), /* @__PURE__ */ (0, V.jsxs)("span", {
		className: "inline-flex items-center gap-1 font-mono text-muted-foreground",
		children: [/* @__PURE__ */ (0, V.jsx)(c, {
			icon: mt,
			className: "shrink-0"
		}), /* @__PURE__ */ (0, V.jsx)("span", { children: n })]
	})] });
}
function Df({ locations: e, workDir: t, onOpenFile: n, entityHref: r, className: a }) {
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: i("space-y-0.5", a),
		"data-lint-violation-locations": "",
		children: [/* @__PURE__ */ (0, V.jsx)("div", {
			className: "text-[10px] uppercase tracking-wide text-muted-foreground",
			children: e.length === 1 ? "Location" : `${e.length} locations`
		}), /* @__PURE__ */ (0, V.jsx)("ul", {
			className: "space-y-1",
			children: e.map((e, i) => /* @__PURE__ */ (0, V.jsxs)("li", {
				className: "flex flex-wrap items-center gap-x-2 gap-y-1 text-xs",
				"data-testid": "lint-violation-location",
				...i === 0 ? { "data-primary": "" } : {},
				children: [
					e.file && /* @__PURE__ */ (0, V.jsx)(Sf, {
						location: e,
						workDir: t,
						primary: i === 0
					}),
					e.entity && /* @__PURE__ */ (0, V.jsx)(wf, {
						entity: e.entity,
						entityHref: r
					}),
					e.url && /* @__PURE__ */ (0, V.jsx)(Tf, { target: e.url }),
					e.image && /* @__PURE__ */ (0, V.jsx)(Ef, { image: e.image }),
					e.message && /* @__PURE__ */ (0, V.jsx)("span", { children: e.message }),
					e.file && n && /* @__PURE__ */ (0, V.jsx)(tu, {
						target: {
							file: wt(e.file, t) ?? e.file,
							...t ? { workDir: t } : {},
							...e.line ? { line: e.line } : {}
						},
						onOpenFile: n
					})
				]
			}, i))
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/verification/verifyReportTests.ts
var Of = "acceptance-criteria";
function kf(e) {
	return typeof e == "number" && Number.isFinite(e) ? e : void 0;
}
function Af(e) {
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
function jf(e) {
	let t = { name: e.name };
	e.framework !== void 0 && (t.framework = e.framework), e.task_id !== void 0 && (t.task_id = e.task_id), e.file !== void 0 && (t.file = e.file), e.line !== void 0 && (t.line = e.line), e.message !== void 0 && (t.message = e.message), e.command !== void 0 && (t.command = e.command), e.work_dir !== void 0 && (t.work_dir = e.work_dir), e.stdout !== void 0 && (t.stdout = e.stdout), e.stderr !== void 0 && (t.stderr = e.stderr);
	let n = kf(e.duration);
	return n !== void 0 && (t.duration = n), e.passed !== void 0 && (t.passed = e.passed), e.failed !== void 0 && (t.failed = e.failed), e.warned !== void 0 && (t.warned = e.warned), e.skipped !== void 0 && (t.skipped = e.skipped), e.pending !== void 0 && (t.pending = e.pending), e.running !== void 0 && (t.running = e.running), e.timed_out !== void 0 && (t.timed_out = e.timed_out), e.progress !== void 0 && (t.progress = e.progress), e.context !== void 0 && (t.context = e.context), e.detail !== void 0 && (t.detail = e.detail), e.summary === void 0 ? e.children !== void 0 && (t.children = e.children.map(jf)) : t.summary = Af(e.summary), t;
}
function Mf(e, t) {
	let n = {
		name: e.item || `Criterion ${t + 1}`,
		framework: "checklist",
		task_id: `${Of}:${t}`,
		passed: e.passed === !0,
		failed: e.passed === !1,
		pending: e.passed !== !0 && e.passed !== !1,
		detail: e
	};
	return e.message !== void 0 && (n.message = e.message), n;
}
function Nf(e) {
	return e.length === 0 ? null : {
		name: "Acceptance criteria",
		framework: "checklist",
		task_id: Of,
		children: e.map(Mf)
	};
}
function Pf(e) {
	let t = (e.tests ?? []).map(jf), n = Nf(e.checklist ?? []);
	return n ? [...t, n] : t;
}
// ../../../clicky-ui/packages/ui/src/data/charts/chart-palette.ts
var Ff = [
	{
		index: 1,
		hue: "sky",
		hex: "#0284c7"
	},
	{
		index: 2,
		hue: "amber",
		hex: "#d97706"
	},
	{
		index: 3,
		hue: "rose",
		hex: "#e11d48"
	},
	{
		index: 4,
		hue: "violet",
		hex: "#7c3aed"
	},
	{
		index: 5,
		hue: "emerald",
		hex: "#059669"
	},
	{
		index: 6,
		hue: "indigo",
		hex: "#6366f1"
	},
	{
		index: 7,
		hue: "teal",
		hex: "#0d9488"
	}
];
function If(e) {
	if (!Number.isInteger(e) || e < 0) throw Error(`chartColor: index must be a non-negative integer, got ${e}`);
	let t = Ff[e % Ff.length];
	return `var(--chart-${t.index}, ${t.hex})`;
}
var Lf = {
	text: "var(--foreground, #111827)",
	axis: "var(--chart-axis, var(--muted-foreground, #6b7280))",
	grid: "var(--chart-grid, var(--border, #e5e7eb))",
	surface: "var(--background, #ffffff)",
	other: "var(--muted-foreground, #6b7280)"
};
// ../../../clicky-ui/packages/ui/src/data/charts/chart-scale.ts
function Rf(e) {
	return Math.round(e * 100) / 100;
}
var zf = new Intl.NumberFormat("en-US", { maximumFractionDigits: 2 });
function Bf(e) {
	return zf.format(e);
}
function Vf(e, t, n) {
	if (!Number.isFinite(t) || t <= 0) throw Error(`${e}: width must be a positive number, got ${t}`);
	if (n !== void 0 && (!Number.isFinite(n) || n <= 0)) throw Error(`${e}: height must be a positive number, got ${n}`);
}
function Hf(e, t, n) {
	if (!Number.isFinite(n)) throw Error(`${e}: value for "${t}" must be finite, got ${n}`);
}
// ../../../clicky-ui/packages/ui/src/data/charts/chart-marks.tsx
var Uf = {
	display: "block",
	maxWidth: "100%",
	height: "auto",
	overflow: "visible",
	fontVariantNumeric: "tabular-nums"
};
function Wf({ width: e, height: t, label: n, title: r, className: i, children: a }) {
	return /* @__PURE__ */ (0, V.jsxs)("svg", {
		"aria-label": n,
		className: i,
		fontSize: 11,
		height: t,
		role: "img",
		style: Uf,
		viewBox: `0 0 ${e} ${t}`,
		width: e,
		xmlns: "http://www.w3.org/2000/svg",
		children: [r !== void 0 && /* @__PURE__ */ (0, V.jsx)("title", { children: r }), a]
	});
}
function Gf({ width: e, height: t, className: n }) {
	return /* @__PURE__ */ (0, V.jsxs)(Wf, {
		className: n,
		height: t,
		label: "No data",
		width: e,
		children: [/* @__PURE__ */ (0, V.jsx)("rect", {
			fill: "none",
			height: Math.max(0, t - 2),
			rx: 4,
			stroke: Lf.grid,
			strokeDasharray: "4 3",
			width: Math.max(0, e - 2),
			x: 1,
			y: 1
		}), /* @__PURE__ */ (0, V.jsx)("text", {
			dominantBaseline: "central",
			fill: Lf.axis,
			fontSize: Math.min(11, t * .6),
			textAnchor: "middle",
			x: e / 2,
			y: t / 2,
			children: "No data"
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/charts/Sparkline.tsx
function Kf({ values: e, width: t, height: n, color: r = If(0), className: i }) {
	if (Vf("Sparkline", t, n), e.forEach((e, t) => Hf("Sparkline", `#${t}`, e)), e.length === 0) return /* @__PURE__ */ (0, V.jsx)(Gf, {
		className: i,
		height: n,
		width: t
	});
	let a = Math.min(4, Math.max(1.5, n / 8)), o = a + 1, s = Math.min(...e), c = Math.max(...e), l = (r) => {
		let i = e.length === 1 ? t / 2 : o + r / (e.length - 1) * (t - o * 2), a = c === s ? n / 2 : n - o - (e[r] - s) / (c - s) * (n - o * 2);
		return [Rf(i), Rf(a)];
	}, u = e.map((e, t) => l(t)), [d, f] = u[u.length - 1], p = `${Bf(e[0])} → ${Bf(e[e.length - 1])} (min ${Bf(s)}, max ${Bf(c)})`;
	return /* @__PURE__ */ (0, V.jsxs)(Wf, {
		className: i,
		height: n,
		label: p,
		title: p,
		width: t,
		children: [u.length > 1 && /* @__PURE__ */ (0, V.jsx)("polyline", {
			fill: "none",
			points: u.map((e) => e.join(",")).join(" "),
			stroke: r,
			strokeLinecap: "round",
			strokeLinejoin: "round",
			strokeWidth: 2
		}), /* @__PURE__ */ (0, V.jsx)("circle", {
			cx: d,
			cy: f,
			fill: r,
			r: Rf(a),
			stroke: Lf.surface,
			strokeWidth: 2
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/verification/stepKind.ts
var qf = [
	"satisfied",
	"failed",
	"timeout"
], Jf = [
	"alert",
	"console",
	"pageerror",
	"response"
], Yf = [
	"answered",
	"rejected",
	"supplied",
	"unanswered"
], Xf = /* @__PURE__ */ new Set([
	"step",
	"outputs",
	"view",
	"form",
	"exports",
	"wait",
	"attempts",
	"events",
	"background"
]), Zf = class extends Error {};
function Qf(e, t) {
	throw new Zf(`${e} must be ${t}`);
}
var $f = (e) => typeof e == "object" && !!e && !Array.isArray(e);
function X(e, t) {
	return $f(e) ? e : Qf(t, "an object");
}
function ep(e, t) {
	return Array.isArray(e) ? e : Qf(t, "an array");
}
function Z(e, t) {
	return typeof e == "string" ? e : Qf(t, "a string");
}
function Q(e, t) {
	return typeof e == "number" && Number.isFinite(e) ? e : Qf(t, "a number");
}
function tp(e, t) {
	return typeof e == "boolean" ? e : Qf(t, "a boolean");
}
function np(e, t, n) {
	return t.includes(e) ? e : Qf(n, `one of ${t.join(", ")} (got ${JSON.stringify(e)})`);
}
function $(e, t, n) {
	return e[t] === void 0 || e[t] === null ? {} : { [t]: n(e[t]) };
}
function rp(e) {
	let t = X(e, "detail.view");
	return Q(t.version, "detail.view.version"), X(t.node, "detail.view.node"), t;
}
function ip(e) {
	let t = X(e, "detail.form"), n = X(t.schema, "detail.form.schema");
	return {
		title: Z(t.title, "detail.form.title"),
		schema: n,
		values: X(t.values, "detail.form.values"),
		state: np(t.state, Yf, "detail.form.state"),
		...$(t, "description", (e) => Z(e, "detail.form.description")),
		...$(t, "answeredBy", (e) => Z(e, "detail.form.answeredBy")),
		...$(t, "answeredAt", (e) => Z(e, "detail.form.answeredAt"))
	};
}
function ap(e, t) {
	return typeof e == "string" ? e : typeof e == "number" ? L(e) : Qf(t, "a duration string or nanosecond number");
}
function op(e, t) {
	return typeof e == "string" ? e : typeof e == "number" ? String(e) : Qf(t, "a string or number");
}
function sp(e) {
	let t = "detail.wait.policy", n = X(e, t), r = (e) => {
		let n = X(e, `${t}.backoff`);
		return {
			...$(n, "factor", (e) => Q(e, `${t}.backoff.factor`)),
			...$(n, "max", (e) => ap(e, `${t}.backoff.max`))
		};
	};
	return {
		...$(n, "until", (e) => Z(e, `${t}.until`)),
		...$(n, "failWhen", (e) => Z(e, `${t}.failWhen`)),
		...$(n, "interval", (e) => ap(e, `${t}.interval`)),
		...$(n, "backoff", r),
		...$(n, "jitter", (e) => op(e, `${t}.jitter`)),
		...$(n, "timeout", (e) => ap(e, `${t}.timeout`))
	};
}
function cp(e) {
	let t = X(e, "detail.wait");
	return {
		policy: sp(t.policy ?? {}),
		outcome: np(t.outcome, qf, "detail.wait.outcome"),
		durationMs: lp(t.durationMs, "detail.wait.durationMs"),
		...t.outputs === void 0 ? {} : { outputs: t.outputs },
		...$(t, "expression", (e) => Z(e, "detail.wait.expression")),
		...$(t, "cause", (e) => Z(e, "detail.wait.cause")),
		...$(t, "error", (e) => Z(e, "detail.wait.error")),
		...$(t, "trace", (e) => Z(e, "detail.wait.trace"))
	};
}
function lp(e, t) {
	let n = Q(e, t);
	return n >= 0 ? n : Qf(t, "non-negative");
}
function up(e, t) {
	let n = X(e, t);
	Z(n.name, `${t}.name`);
	for (let e of [
		"framework",
		"message",
		"command",
		"work_dir",
		"stdout",
		"stderr"
	]) n[e] !== void 0 && Z(n[e], `${t}.${e}`);
	for (let e of [
		"passed",
		"failed",
		"timed_out",
		"warned",
		"skipped",
		"pending",
		"running"
	]) n[e] !== void 0 && tp(n[e], `${t}.${e}`);
	return n.duration !== void 0 && lp(n.duration, `${t}.duration`), n.detail !== void 0 && X(n.detail, `${t}.detail`), n.children !== void 0 && ep(n.children, `${t}.children`).forEach((e, n) => up(e, `${t}.children[${n}]`)), {
		...n,
		name: Z(n.name, `${t}.name`)
	};
}
function dp(e, t) {
	let n = `detail.attempts[${t}]`, r = X(e, n);
	return {
		attempt: Q(r.attempt, `${n}.attempt`),
		status: Z(r.status, `${n}.status`),
		...$(r, "error", (e) => Z(e, `${n}.error`)),
		...$(r, "durationMs", (e) => lp(e, `${n}.durationMs`)),
		...$(r, "result", (e) => up(e, `${n}.result`))
	};
}
function fp(e) {
	let t = ep(e, "detail.attempts").map(dp);
	return t.forEach((e, t) => {
		e.attempt !== t + 1 && Qf(`detail.attempts[${t}].attempt`, `the sequence number ${t + 1}`);
	}), t;
}
function pp(e) {
	return Object.entries(X(e, "detail.exports")).map(([e, t]) => {
		let n = `detail.exports.${e}`, r = X(t, n);
		return {
			name: e,
			value: r.value,
			step: Z(r.step, `${n}.step`)
		};
	});
}
function mp(e, t) {
	let n = `detail.events[${t}]`, r = X(e, n);
	return {
		type: np(r.type, Jf, `${n}.type`),
		...$(r, "text", (e) => Z(e, `${n}.text`)),
		...$(r, "url", (e) => Z(e, `${n}.url`)),
		...$(r, "status", (e) => Q(e, `${n}.status`)),
		...$(r, "detail", (e) => Z(e, `${n}.detail`))
	};
}
function hp(e) {
	let t = X(e.step, "detail.step"), n = Object.fromEntries(Object.entries(e).filter(([e]) => !Xf.has(e)));
	return {
		step: {
			id: Z(t.id, "detail.step.id"),
			kind: Z(t.kind, "detail.step.kind")
		},
		...e.outputs === void 0 ? {} : { outputs: e.outputs },
		...$(e, "view", rp),
		...$(e, "form", ip),
		exports: e.exports == null ? [] : pp(e.exports),
		...$(e, "wait", cp),
		attempts: fp(e.attempts ?? []),
		events: ep(e.events ?? [], "detail.events").map(mp),
		background: tp(e.background ?? !1, "detail.background"),
		rest: n
	};
}
function gp(e) {
	if (!$f(e) || e.step === void 0) return { kind: "none" };
	try {
		return {
			kind: "step",
			step: hp(e)
		};
	} catch (e) {
		if (e instanceof Zf) return {
			kind: "invalid",
			error: e.message
		};
		throw e;
	}
}
var _p = [
	"error",
	"warning",
	"info"
];
function vp(e, t) {
	let n = X(e, t);
	return {
		kind: Z(n.kind, `${t}.kind`),
		id: Z(n.id, `${t}.id`),
		...$(n, "name", (e) => Z(e, `${t}.name`))
	};
}
var yp = [
	"css",
	"xpath",
	"testId",
	"role",
	"text"
];
function bp(e, t) {
	let n = X(e, t), r = {
		url: Z(n.url, `${t}.url`),
		...Object.assign({}, ...[...yp, "name"].map((e) => $(n, e, (n) => Z(n, `${t}.${e}`))))
	}, i = yp.filter((e) => r[e] !== void 0);
	return i.length > 1 && Qf(t, `a target with at most one selector strategy (got ${i.join(", ")})`), r.name !== void 0 && r.role === void 0 && Qf(`${t}.name`, "paired with a role"), r;
}
function xp(e, t) {
	let n = X(e, t);
	return {
		src: Z(n.src, `${t}.src`),
		x: Q(n.x, `${t}.x`),
		y: Q(n.y, `${t}.y`),
		width: Q(n.width, `${t}.width`),
		height: Q(n.height, `${t}.height`)
	};
}
function Sp(e, t) {
	let n = `detail.violation.locations[${t}]`, r = X(e, n), i = {
		...$(r, "file", (e) => Z(e, `${n}.file`)),
		...$(r, "line", (e) => Q(e, `${n}.line`)),
		...$(r, "column", (e) => Q(e, `${n}.column`)),
		...$(r, "end_line", (e) => Q(e, `${n}.end_line`)),
		...$(r, "ref", (e) => Z(e, `${n}.ref`)),
		...$(r, "entity", (e) => vp(e, `${n}.entity`)),
		...$(r, "url", (e) => bp(e, `${n}.url`)),
		...$(r, "image", (e) => xp(e, `${n}.image`)),
		...$(r, "message", (e) => Z(e, `${n}.message`))
	};
	return i.file || i.entity || i.url || i.image || i.message || Qf(n, "a location naming a file, entity, url, image or message"), i;
}
function Cp(e) {
	let t = "detail.violation.rule", n = X(e, t);
	return {
		...$(n, "method", (e) => Z(e, `${t}.method`)),
		...$(n, "description", (e) => Z(e, `${t}.description`)),
		...$(n, "pattern", (e) => Z(e, `${t}.pattern`)),
		...$(n, "package", (e) => Z(e, `${t}.package`))
	};
}
function wp(e) {
	let t = "detail.violation", n = X(e, t);
	return {
		...$(n, "severity", (e) => np(e, _p, `${t}.severity`)),
		...$(n, "message", (e) => Z(e, `${t}.message`)),
		...$(n, "rule", Cp),
		...$(n, "source", (e) => Z(e, `${t}.source`)),
		...$(n, "file", (e) => Z(e, `${t}.file`)),
		...$(n, "line", (e) => Q(e, `${t}.line`)),
		locations: ep(n.locations ?? [], `${t}.locations`).map(Sp)
	};
}
function Tp(e) {
	if (!$f(e) || e.violation === void 0) return { kind: "none" };
	try {
		return {
			kind: "violation",
			violation: wp(e.violation)
		};
	} catch (e) {
		if (e instanceof Zf) return {
			kind: "invalid",
			error: e.message
		};
		throw e;
	}
}
function Ep(e, t) {
	if (!$f(e) || !(t in e)) return e;
	let n = Object.fromEntries(Object.entries(e).filter(([e]) => e !== t));
	return Object.keys(n).length > 0 ? n : void 0;
}
function Dp(e) {
	return Ep(e, "violation");
}
function Op(e) {
	let t = "detail.query.connection", n = X(e, t);
	return {
		name: Z(n.name, `${t}.name`),
		type: Z(n.type, `${t}.type`),
		...$(n, "host", (e) => Z(e, `${t}.host`)),
		...$(n, "port", (e) => Z(e, `${t}.port`)),
		...$(n, "user", (e) => Z(e, `${t}.user`)),
		...$(n, "database", (e) => Z(e, `${t}.database`))
	};
}
function kp(e, t) {
	let n = `detail.query.batches[${t}]`, r = X(e, n);
	return {
		batch: Q(r.batch, `${n}.batch`),
		run: Q(r.run, `${n}.run`),
		...$(r, "rowsAffected", (e) => Q(e, `${n}.rowsAffected`))
	};
}
function Ap(e) {
	let t = "detail.query.session", n = X(e, t);
	return {
		engine: Z(n.engine, `${t}.engine`),
		sessionId: Q(n.sessionId, `${t}.sessionId`),
		...$(n, "queryId", (e) => Z(e, `${t}.queryId`)),
		...$(n, "server", (e) => Z(e, `${t}.server`)),
		...$(n, "database", (e) => Z(e, `${t}.database`)),
		...$(n, "user", (e) => Z(e, `${t}.user`)),
		...$(n, "page", (e) => Q(e, `${t}.page`))
	};
}
var jp = [
	"database",
	"user",
	"application",
	"clientHost",
	"state",
	"command",
	"waitType",
	"waitResource",
	"startedAt",
	"transactionStartedAt",
	"queryId",
	"sql"
], Mp = [
	"requestId",
	"blockedBy",
	"waitMs",
	"elapsedMs",
	"transactionMs",
	"cpuMs",
	"logicalReads",
	"reads",
	"writes",
	"memoryBytes",
	"openTransactions"
];
function Np(e, t) {
	let n = `detail.query.diagnose[${t}]`, r = X(e, n), i = {
		sampledAt: Z(r.sampledAt, `${n}.sampledAt`),
		engine: Z(r.engine, `${n}.engine`),
		sessionId: Q(r.sessionId, `${n}.sessionId`),
		...$(r, "blocked", (e) => tp(e, `${n}.blocked`))
	};
	for (let e of jp) Object.assign(i, $(r, e, (t) => Z(t, `${n}.${e}`)));
	for (let e of Mp) Object.assign(i, $(r, e, (t) => Q(t, `${n}.${e}`)));
	return i;
}
function Pp(e) {
	let t = "detail.query", n = X(e, t);
	return {
		template: Z(n.template, `${t}.template`),
		...$(n, "sql", (e) => Z(e, `${t}.sql`)),
		...$(n, "arguments", (e) => ep(e, `${t}.arguments`)),
		...$(n, "params", (e) => X(e, `${t}.params`)),
		...$(n, "queryFile", (e) => Z(e, `${t}.queryFile`)),
		...$(n, "connection", Op),
		...$(n, "batches", (e) => ep(e, `${t}.batches`).map(kp)),
		...$(n, "session", Ap),
		...$(n, "diagnose", (e) => ep(e, `${t}.diagnose`).map(Np)),
		...$(n, "diagnoseDropped", (e) => Q(e, `${t}.diagnoseDropped`)),
		...$(n, "diagnoseError", (e) => Z(e, `${t}.diagnoseError`))
	};
}
function Fp(e) {
	if (!$f(e) || e.query === void 0) return { kind: "none" };
	try {
		return {
			kind: "query",
			query: Pp(e.query)
		};
	} catch (e) {
		if (e instanceof Zf) return {
			kind: "invalid",
			error: e.message
		};
		throw e;
	}
}
function Ip(e) {
	return Ep(e, "query");
}
function Lp(e) {
	let t = Object.entries(e.params ?? {}), n = e.connection;
	return [
		...n ? [{
			title: "Connection",
			body: Rp(n)
		}] : [],
		{
			title: e.queryFile ? `Template (${e.queryFile})` : "Template",
			lang: "sql",
			body: e.template
		},
		...e.sql ? [{
			title: "SQL",
			lang: "sql",
			body: e.sql
		}] : [],
		...t.length > 0 ? [{
			title: "Params",
			body: t.map(([e, t]) => `${e}=${Jp(t)}`).join("\n")
		}] : [],
		...e.arguments && e.arguments.length > 0 ? [{
			title: "Arguments",
			body: e.arguments.map((e, t) => `${t + 1}: ${Jp(e)}`).join("\n")
		}] : []
	];
}
function Rp(e) {
	let t = e.host && (e.port ? `${e.host}:${e.port}` : e.host);
	return [
		`${e.name} (${e.type})`,
		t && `host: ${t}`,
		e.user && `user: ${e.user}`,
		e.database && `db: ${e.database}`
	].filter(Boolean).join("\n");
}
function zp(e) {
	let t = gp(e.detail);
	return t.kind === "step" ? t.step.form?.state : void 0;
}
function Bp(e) {
	return gp(e).kind !== "none";
}
var Vp = (e) => typeof e == "string", Hp = (e) => typeof e == "number" && Number.isFinite(e);
function Up(e) {
	if (!$f(e)) return {
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
	n("name", Vp), i("kind", n("kind", Vp)), i("state", n("state", Vp)), i("startedAt", n("started_at", Vp)), i("finishedAt", n("finished_at", Vp)), i("durationNs", n("duration", Hp)), i("error", n("error", Vp));
	let a = n("origin", $f);
	return a && (i("file", Vp(a.file) ? a.file : void 0), i("sectionPath", Vp(a.section_path) ? a.section_path : void 0), i("rowIndex", Hp(a.row_index) ? a.row_index : void 0), i("line", Hp(a.line) ? a.line : void 0)), {
		header: r,
		rest: Object.keys(t).length > 0 ? t : void 0
	};
}
function Wp(e) {
	return e.expression === void 0 ? e.outcome === "failed" && e.policy.failWhen ? "failWhen" : "until" : e.expression === e.policy.failWhen ? "failWhen" : "until";
}
function Gp(e) {
	return /T(\d{2}:\d{2}:\d{2})/.exec(e)?.[1] ?? e;
}
function Kp(e) {
	return e <= 0 ? "0ms" : L(e * 1e6);
}
function qp(e) {
	let t = [e.backoff?.factor !== void 0 && `×${e.backoff.factor}`, e.backoff?.max && `up to ${e.backoff.max}`].filter(Boolean);
	return [
		e.interval && `every ${e.interval}`,
		t.length > 0 && `backoff ${t.join(" ")}`,
		e.jitter && `jitter ±${e.jitter}`,
		e.timeout && `timeout ${e.timeout}`
	].filter(Boolean).join(" · ");
}
function Jp(e) {
	return e === void 0 ? "" : typeof e == "string" ? e : JSON.stringify(e);
}
function Yp(e) {
	let { policy: t } = e;
	return {
		title: "Wait",
		body: [
			`outcome: ${e.outcome}`,
			`elapsed: ${Kp(e.durationMs)}`,
			`policy: ${qp(t)}`,
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
function Xp(e) {
	return `${[
		`#${e.attempt}`,
		e.status,
		e.durationMs !== void 0 && Kp(e.durationMs)
	].filter(Boolean).join(" ")}${e.error ? ` — ${e.error}` : ""}`;
}
function Zp(e) {
	let t = [
		e.status,
		e.text,
		e.url
	].filter((e) => e !== void 0).join(" ");
	return `[${e.type}]${t ? ` ${t}` : ""}${e.detail ? ` — ${e.detail}` : ""}`;
}
function Qp(e) {
	return [
		...e.outputs === void 0 ? [] : [{
			title: "Outputs",
			body: JSON.stringify(e.outputs, null, 2),
			lang: "json"
		}],
		...e.wait ? [Yp(e.wait)] : [],
		...e.attempts.length > 0 ? [{
			title: "Attempts",
			body: e.attempts.map(Xp).join("\n")
		}] : [],
		...e.exports.length > 0 ? [{
			title: "Exports",
			body: e.exports.map((e) => `${e.name}=${Jp(e.value)} (from ${e.step})`).join("\n")
		}] : [],
		...e.events.length > 0 ? [{
			title: "Events",
			body: e.events.map(Zp).join("\n")
		}] : []
	];
}
// ../../../clicky-ui/packages/ui/src/data/verification/stepKindViews.tsx
var $p = "border-red-500/60 bg-red-500/10 text-red-700 dark:text-red-400";
function em({ name: e, title: t, aside: n, children: r }) {
	return /* @__PURE__ */ (0, V.jsxs)("section", {
		className: "min-w-0 space-y-1.5",
		"data-step-section": e,
		children: [/* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex items-center gap-2",
			children: [/* @__PURE__ */ (0, V.jsx)("div", {
				className: "text-[10px] uppercase tracking-wide text-muted-foreground",
				children: t
			}), n]
		}), r]
	});
}
var tm = {
	answered: {
		label: "Answered",
		tone: "success"
	},
	rejected: {
		label: "Rejected",
		tone: "warning"
	},
	supplied: {
		label: "Supplied",
		tone: "info"
	},
	unanswered: {
		label: "Not answered",
		tone: "danger"
	}
}, nm = () => {};
function rm({ state: e, size: t }) {
	let { label: n, tone: r } = tm[e];
	return /* @__PURE__ */ (0, V.jsx)(D, {
		tone: r,
		clickToCopy: !1,
		...t ? { size: t } : {},
		children: n
	});
}
function im({ node: e }) {
	if (e.waiting) return /* @__PURE__ */ (0, V.jsx)("span", {
		className: "flex shrink-0 items-center",
		"data-step-badge": "",
		children: /* @__PURE__ */ (0, V.jsx)(D, {
			tone: "warning",
			size: "xs",
			icon: Tn,
			clickToCopy: !1,
			children: "Awaiting answer"
		})
	});
	let t = zp(e);
	return t ? /* @__PURE__ */ (0, V.jsx)("span", {
		className: "flex shrink-0 items-center",
		"data-step-badge": "",
		children: /* @__PURE__ */ (0, V.jsx)(rm, {
			state: t,
			size: "xs"
		})
	}) : null;
}
function am({ form: e }) {
	let t = [e.answeredBy && `by ${e.answeredBy}`, e.answeredAt && `at ${Gp(e.answeredAt)}`].filter(Boolean).join(" ");
	return /* @__PURE__ */ (0, V.jsxs)(em, {
		name: "form",
		title: "Form",
		aside: /* @__PURE__ */ (0, V.jsx)(rm, { state: e.state }),
		children: [
			/* @__PURE__ */ (0, V.jsx)("div", {
				className: "text-sm font-medium",
				children: e.title
			}),
			e.description ? /* @__PURE__ */ (0, V.jsx)("div", {
				className: "text-xs text-muted-foreground",
				children: e.description
			}) : null,
			t ? /* @__PURE__ */ (0, V.jsx)("div", {
				className: "text-[11px] text-muted-foreground",
				children: t
			}) : null,
			/* @__PURE__ */ (0, V.jsx)(Fe, {
				schema: e.schema,
				value: e.values,
				onChange: nm,
				readOnly: !0
			})
		]
	});
}
function om({ children: e }) {
	return /* @__PURE__ */ (0, V.jsx)(em, {
		name: "pending-form",
		title: "Form",
		aside: /* @__PURE__ */ (0, V.jsx)(D, {
			tone: "warning",
			size: "xs",
			icon: Tn,
			clickToCopy: !1,
			children: "Awaiting answer"
		}),
		children: e
	});
}
function sm({ name: e, keyLabel: t, rows: n }) {
	return /* @__PURE__ */ (0, V.jsxs)("table", {
		className: "w-full table-fixed border-collapse text-left text-[11px]",
		[`data-query-${e}`]: "",
		children: [/* @__PURE__ */ (0, V.jsx)("thead", {
			className: "text-muted-foreground",
			children: /* @__PURE__ */ (0, V.jsxs)("tr", {
				className: "border-b border-border",
				children: [/* @__PURE__ */ (0, V.jsx)("th", {
					className: "w-[28%] py-1 pr-2 font-medium",
					children: t
				}), /* @__PURE__ */ (0, V.jsx)("th", {
					className: "py-1 font-medium",
					children: "Value"
				})]
			})
		}), /* @__PURE__ */ (0, V.jsx)("tbody", {
			className: "font-mono",
			children: n.map(([e, t]) => /* @__PURE__ */ (0, V.jsxs)("tr", {
				className: "border-b border-border/60 align-top last:border-0",
				children: [/* @__PURE__ */ (0, V.jsx)("td", {
					className: "break-words py-1 pr-2",
					children: e
				}), /* @__PURE__ */ (0, V.jsx)("td", {
					className: "break-all py-1",
					children: Jp(t)
				})]
			}, e))
		})]
	});
}
function cm({ query: e }) {
	let [t, n] = (0, B.useState)("template"), r = Object.entries(e.params ?? {}), i = e.arguments ?? [], a = [{
		id: "template",
		label: "Template",
		title: "The query as authored, with its params references"
	}, e.sql ? {
		id: "sql",
		label: "SQL",
		title: "The statement issued, params bound to driver placeholders"
	} : {
		id: "sql",
		label: "SQL",
		title: "No issued statement was recorded",
		disabled: !0
	}];
	return /* @__PURE__ */ (0, V.jsxs)(em, {
		name: "query",
		title: "Query",
		children: [
			(e.connection || e.session) && /* @__PURE__ */ (0, V.jsx)(um, {
				connection: e.connection,
				session: e.session
			}),
			e.queryFile ? /* @__PURE__ */ (0, V.jsx)("div", {
				className: "break-all font-mono text-[11px] text-muted-foreground",
				title: e.queryFile,
				children: e.queryFile
			}) : null,
			/* @__PURE__ */ (0, V.jsx)("div", {
				className: "min-w-0",
				"data-query-code": t,
				children: /* @__PURE__ */ (0, V.jsx)(Ve, {
					language: "sql",
					source: t === "sql" && e.sql ? e.sql : e.template,
					copyable: !0,
					toolbar: /* @__PURE__ */ (0, V.jsx)(C, {
						size: "sm",
						value: t,
						options: a,
						onChange: n,
						"aria-label": "Query view"
					})
				})
			}),
			r.length > 0 && /* @__PURE__ */ (0, V.jsx)(sm, {
				name: "params",
				keyLabel: "Param",
				rows: r
			}),
			i.length > 0 && /* @__PURE__ */ (0, V.jsx)(sm, {
				name: "arguments",
				keyLabel: "Argument",
				rows: i.map((e, t) => [String(t + 1), e])
			}),
			(e.batches?.length ?? 0) > 0 && /* @__PURE__ */ (0, V.jsx)(fm, { batches: e.batches }),
			(e.diagnose || e.diagnoseError) && /* @__PURE__ */ (0, V.jsx)(gm, {
				samples: e.diagnose ?? [],
				dropped: e.diagnoseDropped,
				error: e.diagnoseError,
				engine: e.session?.engine
			})
		]
	});
}
function lm({ engine: e, sessionId: t, queryId: n }) {
	switch (e) {
		case "sqlserver": return `spid ${t}`;
		case "postgres": return `pid ${t}`;
		case "mysql": return `conn ${t}`;
		case "clickhouse": return n ? `query ${n}` : void 0;
		default: return `${e} ${t}`;
	}
}
function um({ connection: e, session: t }) {
	let n = [
		["host", e?.host && e.port ? `${e.host}:${e.port}` : e?.host],
		["user", e?.user],
		["db", e?.database],
		["server", t?.server]
	], r = t && lm(t);
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "flex min-w-0 flex-wrap items-center gap-1.5 text-[11px]",
		"data-query-connection": "",
		children: [
			/* @__PURE__ */ (0, V.jsx)(Lt, {
				className: "shrink-0 text-muted-foreground",
				"aria-hidden": !0
			}),
			e && /* @__PURE__ */ (0, V.jsx)("span", {
				className: "font-medium",
				children: e.name
			}),
			e && /* @__PURE__ */ (0, V.jsx)("span", {
				className: "text-muted-foreground",
				children: e.type
			}),
			r && /* @__PURE__ */ (0, V.jsx)("span", {
				className: "inline-flex min-w-0",
				"data-query-session": "",
				children: /* @__PURE__ */ (0, V.jsx)(D, {
					size: "xs",
					tone: "info",
					maxWidth: 40,
					children: r
				})
			}),
			n.map(([e, t]) => t ? /* @__PURE__ */ (0, V.jsx)("span", {
				className: "inline-flex min-w-0",
				"data-query-target": e,
				children: /* @__PURE__ */ (0, V.jsx)(D, {
					variant: "label",
					size: "xs",
					label: e,
					value: t,
					maxWidth: 48
				})
			}, e) : null)
		]
	});
}
function dm({ batch: e, run: t }) {
	return t > 1 ? `batch ${e} run ${t}` : `batch ${e}`;
}
function fm({ batches: e }) {
	return /* @__PURE__ */ (0, V.jsx)("div", {
		className: "flex min-w-0 flex-wrap gap-1",
		"data-query-batches": "",
		children: e.map((e) => /* @__PURE__ */ (0, V.jsxs)(D, {
			size: "xs",
			clickToCopy: !1,
			tone: "success",
			children: [
				dm(e),
				" done",
				e.rowsAffected === void 0 ? "" : ` · ${e.rowsAffected} row(s)`
			]
		}, `${e.batch}:${e.run}`))
	});
}
var pm = ["postgres"], mm = [
	{
		label: "State",
		value: (e) => e.state || e.command || "unknown"
	},
	{
		label: "Elapsed",
		value: (e) => Kp(e.elapsedMs ?? 0)
	},
	{
		label: "CPU",
		hiddenFor: pm,
		value: (e) => Kp(e.cpuMs ?? 0)
	},
	{
		label: "Logical reads",
		hiddenFor: pm,
		value: (e) => ze(e.logicalReads ?? 0)
	},
	{
		label: "Reads",
		hiddenFor: pm,
		value: (e) => ze(e.reads ?? 0)
	},
	{
		label: "Writes",
		hiddenFor: pm,
		value: (e) => ze(e.writes ?? 0)
	},
	{
		label: "Memory",
		hiddenFor: pm,
		value: (e) => nt(e.memoryBytes ?? 0)
	},
	{
		label: "Wait",
		value: (e) => e.waitType ? `${e.waitType}${e.waitMs ? ` ${Kp(e.waitMs)}` : ""}` : "none",
		tone: (e) => e.waitType ? "warning" : "neutral"
	},
	{
		label: "Blocked by",
		value: (e) => String(e.blockedBy),
		tone: () => "danger",
		when: (e) => (e.blockedBy ?? 0) > 0
	}
], hm = [{
	key: "cpu",
	label: "CPU",
	hiddenFor: pm,
	read: (e) => e.cpuMs ?? 0
}, {
	key: "logicalReads",
	label: "Logical reads",
	hiddenFor: pm,
	read: (e) => e.logicalReads ?? 0
}];
function gm({ samples: e, dropped: t, error: n, engine: r }) {
	let i = e[e.length - 1], a = i?.engine ?? r, o = (e) => !(a && e.hiddenFor?.includes(a)), s = `${e.length} sample${e.length === 1 ? "" : "s"}${t ? ` · ${t} older dropped` : ""}`;
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "min-w-0 space-y-1.5",
		"data-query-diagnose": "",
		children: [n && /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex items-start gap-1.5 text-[11px] text-yellow-700 dark:text-yellow-400",
			"data-query-diagnose-error": "",
			children: [/* @__PURE__ */ (0, V.jsx)(l, {
				className: "mt-px shrink-0",
				"aria-hidden": !0
			}), /* @__PURE__ */ (0, V.jsx)("span", {
				className: "min-w-0 break-words",
				children: n
			})]
		}), i && /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
			/* @__PURE__ */ (0, V.jsx)("div", {
				className: "text-[10px] uppercase tracking-wide text-muted-foreground",
				"data-query-diagnose-count": "",
				children: s
			}),
			/* @__PURE__ */ (0, V.jsx)("div", {
				className: "grid grid-cols-[repeat(auto-fill,minmax(5.5rem,1fr))] gap-1",
				children: mm.filter((e) => o(e) && (e.when?.(i) ?? !0)).map((e) => /* @__PURE__ */ (0, V.jsx)(eu, {
					label: e.label,
					value: e.value(i),
					tone: e.tone?.(i) ?? "neutral"
				}, e.label))
			}),
			i.waitResource && /* @__PURE__ */ (0, V.jsx)("div", {
				className: "break-all font-mono text-[11px] text-muted-foreground",
				"data-query-diagnose-wait-resource": "",
				children: i.waitResource
			}),
			e.length > 1 && /* @__PURE__ */ (0, V.jsx)("div", {
				className: "flex flex-wrap gap-3",
				children: hm.filter(o).map((t) => /* @__PURE__ */ (0, V.jsxs)("div", {
					className: "flex items-center gap-1.5 text-[10px] text-muted-foreground",
					"data-query-spark": t.key,
					children: [/* @__PURE__ */ (0, V.jsx)("span", { children: t.label }), /* @__PURE__ */ (0, V.jsx)(Kf, {
						values: e.map(t.read),
						width: 96,
						height: 22
					})]
				}, t.key))
			})
		] })]
	});
}
var _m = {
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
function vm({ condition: e, expression: t, decisive: n, wait: r }) {
	let { label: i, tone: a } = _m[r.outcome];
	return /* @__PURE__ */ (0, V.jsxs)("li", {
		className: "flex min-w-0 items-center gap-2",
		"data-wait-condition": e,
		...n ? { "data-wait-decisive": "" } : {},
		children: [
			/* @__PURE__ */ (0, V.jsx)("span", {
				className: "w-14 shrink-0 text-muted-foreground",
				children: e
			}),
			/* @__PURE__ */ (0, V.jsx)("code", {
				className: "min-w-0 shrink break-words font-mono",
				children: t
			}),
			n && /* @__PURE__ */ (0, V.jsx)(D, {
				size: "xs",
				clickToCopy: !1,
				tone: a,
				children: i
			})
		]
	});
}
function ym({ wait: e, onViewOutputs: t }) {
	let { policy: n } = e, r = qp(n), i = Wp(e), a = e.outputs !== void 0 && !t;
	return /* @__PURE__ */ (0, V.jsxs)("section", {
		className: "min-w-0 text-[11px]",
		"data-step-section": "wait",
		children: [
			/* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex min-w-0 items-center gap-2",
				"data-wait-summary": "",
				children: [
					/* @__PURE__ */ (0, V.jsx)("span", {
						className: "shrink-0 text-muted-foreground",
						children: "Wait"
					}),
					/* @__PURE__ */ (0, V.jsx)(D, {
						size: "xs",
						clickToCopy: !1,
						children: Kp(e.durationMs)
					}),
					e.outputs !== void 0 && t && /* @__PURE__ */ (0, V.jsx)(o, {
						type: "button",
						variant: "link",
						size: "sm",
						className: "h-auto px-0 text-[11px]",
						onClick: t,
						children: "View outputs"
					})
				]
			}),
			/* @__PURE__ */ (0, V.jsxs)("ul", {
				className: "mt-1 space-y-0.5",
				"data-wait-conditions": "",
				children: [/* @__PURE__ */ (0, V.jsx)(vm, {
					condition: "until",
					expression: n.until ?? "Service readiness",
					decisive: i === "until",
					wait: e
				}), n.failWhen && /* @__PURE__ */ (0, V.jsx)(vm, {
					condition: "failWhen",
					expression: n.failWhen,
					decisive: i === "failWhen",
					wait: e
				})]
			}),
			(r || e.error || e.trace || a) && /* @__PURE__ */ (0, V.jsxs)("details", {
				className: "mt-1 text-muted-foreground",
				open: e.outcome !== "satisfied",
				children: [/* @__PURE__ */ (0, V.jsx)("summary", {
					className: "cursor-pointer",
					children: "Wait details"
				}), /* @__PURE__ */ (0, V.jsxs)("div", {
					className: "mt-1 space-y-1 break-words",
					children: [
						r && /* @__PURE__ */ (0, V.jsx)("p", { children: r }),
						e.error && /* @__PURE__ */ (0, V.jsx)("p", {
							className: "text-destructive",
							children: e.error
						}),
						e.trace && /* @__PURE__ */ (0, V.jsx)("pre", {
							className: "whitespace-pre-wrap",
							children: e.trace
						}),
						a && /* @__PURE__ */ (0, V.jsxs)("div", {
							className: "min-w-0",
							"data-wait-outputs": "",
							children: [/* @__PURE__ */ (0, V.jsx)("div", {
								className: "text-[10px] uppercase tracking-wide",
								children: "Outputs"
							}), /* @__PURE__ */ (0, V.jsx)(et, {
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
var bm = {
	passed: "success",
	failed: "danger",
	warned: "warning",
	timedout: "warning",
	running: "info"
};
function xm(e) {
	return e.split("/").filter(Boolean).at(-1) ?? e;
}
function Sm({ rowIndex: e, line: t }) {
	return e === void 0 ? t === void 0 ? void 0 : `line ${t}` : `step #${e}`;
}
function Cm({ header: e, stepKind: t }) {
	let n = [Sm(e), e.sectionPath].filter(Boolean).join(" · "), r = e.kind === t ? void 0 : e.kind, i = e.startedAt && e.finishedAt ? `${Gp(e.startedAt)} → ${Gp(e.finishedAt)}` : e.startedAt ? `started ${Gp(e.startedAt)}` : void 0;
	return !e.file && !n && !r && !e.state && !i && e.durationNs === void 0 ? null : /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "min-w-0 space-y-1",
		"data-execution-header": "",
		children: [(e.file || n) && /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex min-w-0 flex-wrap items-baseline gap-x-1.5 text-xs",
			children: [
				e.file && /* @__PURE__ */ (0, V.jsx)("span", {
					className: "font-medium",
					title: e.file,
					children: xm(e.file)
				}),
				e.file && n && /* @__PURE__ */ (0, V.jsx)("span", {
					className: "text-muted-foreground",
					children: "·"
				}),
				n && /* @__PURE__ */ (0, V.jsx)("span", {
					className: "min-w-0 break-words text-muted-foreground",
					children: n
				})
			]
		}), /* @__PURE__ */ (0, V.jsxs)("div", {
			className: "flex min-w-0 flex-wrap items-center gap-2 text-[11px] text-muted-foreground",
			children: [
				r && /* @__PURE__ */ (0, V.jsx)(D, {
					size: "xs",
					clickToCopy: !1,
					children: r
				}),
				e.state && /* @__PURE__ */ (0, V.jsx)(D, {
					size: "xs",
					clickToCopy: !1,
					tone: bm[e.state] ?? "neutral",
					children: e.state
				}),
				i && /* @__PURE__ */ (0, V.jsx)("span", {
					className: "font-mono",
					children: i
				}),
				e.durationNs !== void 0 && /* @__PURE__ */ (0, V.jsx)("span", { children: L(e.durationNs) })
			]
		})]
	});
}
function wm({ children: e }) {
	return /* @__PURE__ */ (0, V.jsx)("div", {
		className: i("whitespace-pre-wrap break-words rounded border px-2 py-1.5 text-sm", $p),
		"data-failure-callout": "",
		children: e
	});
}
function Tm(e) {
	return e.children?.length ? e.children.flatMap(Tm) : e.failed || e.timed_out ? [e] : [];
}
function Em({ node: e }) {
	let t = e.children?.length ? Tm(e) : [];
	return t.length === 0 ? null : /* @__PURE__ */ (0, V.jsx)(em, {
		name: "failed-tests",
		title: `Failed tests (${t.length})`,
		children: /* @__PURE__ */ (0, V.jsx)("ul", {
			className: "space-y-1.5",
			children: t.map((e, t) => /* @__PURE__ */ (0, V.jsxs)("li", {
				className: i("space-y-0.5 rounded border px-2 py-1.5", $p),
				"data-failed-test": e.name,
				children: [/* @__PURE__ */ (0, V.jsxs)("div", {
					className: "flex min-w-0 items-baseline gap-2 text-sm",
					children: [
						/* @__PURE__ */ (0, V.jsx)("span", {
							className: "min-w-0 break-words font-medium",
							children: e.name
						}),
						e.file && /* @__PURE__ */ (0, V.jsx)("span", {
							className: "shrink-0 font-mono text-[11px] opacity-80",
							children: e.line ? `${e.file}:${e.line}` : e.file
						}),
						e.duration !== void 0 && /* @__PURE__ */ (0, V.jsx)("span", {
							className: "ml-auto shrink-0 text-[11px] opacity-80",
							children: L(e.duration)
						})
					]
				}), e.message && /* @__PURE__ */ (0, V.jsx)("pre", {
					className: "whitespace-pre-wrap break-words font-mono text-xs",
					children: e.message
				})]
			}, e.task_id ?? `${e.name}-${t}`))
		})
	});
}
function Dm({ exports: e }) {
	return /* @__PURE__ */ (0, V.jsxs)("table", {
		className: "w-full table-fixed border-collapse text-left text-[11px]",
		children: [/* @__PURE__ */ (0, V.jsx)("thead", {
			className: "text-muted-foreground",
			children: /* @__PURE__ */ (0, V.jsxs)("tr", {
				className: "border-b border-border",
				children: [
					/* @__PURE__ */ (0, V.jsx)("th", {
						className: "w-[28%] py-1 pr-2 font-medium",
						children: "Name"
					}),
					/* @__PURE__ */ (0, V.jsx)("th", {
						className: "py-1 pr-2 font-medium",
						children: "Value"
					}),
					/* @__PURE__ */ (0, V.jsx)("th", {
						className: "w-[18%] py-1 font-medium",
						children: "Step"
					})
				]
			})
		}), /* @__PURE__ */ (0, V.jsx)("tbody", {
			className: "font-mono",
			children: e.map((e) => /* @__PURE__ */ (0, V.jsxs)("tr", {
				className: "border-b border-border/60 align-top last:border-0",
				children: [
					/* @__PURE__ */ (0, V.jsx)("td", {
						className: "break-words py-1 pr-2",
						children: e.name
					}),
					/* @__PURE__ */ (0, V.jsx)("td", {
						className: "break-all py-1 pr-2",
						children: Jp(e.value)
					}),
					/* @__PURE__ */ (0, V.jsx)("td", {
						className: "break-words py-1 text-muted-foreground",
						children: e.step
					})
				]
			}, e.name))
		})]
	});
}
function Om(e) {
	return e.type === "alert" || e.type === "pageerror" || e.type === "response" && (e.status ?? 0) >= 400;
}
function km({ event: e }) {
	let t = Om(e);
	return /* @__PURE__ */ (0, V.jsxs)("li", {
		className: i("space-y-0.5 rounded border px-2 py-1 text-[11px]", t ? $p : "border-border"),
		"data-event-type": e.type,
		children: [
			/* @__PURE__ */ (0, V.jsxs)("div", {
				className: "flex min-w-0 items-center gap-2",
				children: [
					/* @__PURE__ */ (0, V.jsx)(D, {
						size: "xs",
						clickToCopy: !1,
						tone: t ? "danger" : "neutral",
						children: e.type
					}),
					e.status !== void 0 && /* @__PURE__ */ (0, V.jsx)("span", {
						className: "font-mono",
						children: e.status
					}),
					e.text && /* @__PURE__ */ (0, V.jsx)("span", {
						className: i("min-w-0 break-words", e.type === "alert" && "font-medium"),
						children: e.text
					})
				]
			}),
			e.url && /* @__PURE__ */ (0, V.jsx)("div", {
				className: "break-all font-mono text-[10px] opacity-80",
				children: e.url
			}),
			e.detail && /* @__PURE__ */ (0, V.jsx)("p", {
				className: "whitespace-pre-wrap break-words",
				children: e.detail
			})
		]
	});
}
function Am({ step: e, renderWait: t }) {
	let n = Pu(), r = n && e.outputs !== void 0 ? () => n("outputs") : void 0;
	return /* @__PURE__ */ (0, V.jsxs)(V.Fragment, { children: [
		e.wait && (t ? t(e.wait) : /* @__PURE__ */ (0, V.jsx)(ym, {
			wait: e.wait,
			...r ? { onViewOutputs: r } : {}
		})),
		e.events.length > 0 && /* @__PURE__ */ (0, V.jsx)(em, {
			name: "events",
			title: "Events",
			children: /* @__PURE__ */ (0, V.jsx)("ul", {
				className: "space-y-1",
				children: e.events.map((e, t) => /* @__PURE__ */ (0, V.jsx)(km, { event: e }, t))
			})
		}),
		e.exports.length > 0 && /* @__PURE__ */ (0, V.jsx)(em, {
			name: "exports",
			title: "Exports",
			children: /* @__PURE__ */ (0, V.jsx)(Dm, { exports: e.exports })
		})
	] });
}
function jm({ outputs: e }) {
	return /* @__PURE__ */ (0, V.jsx)("div", {
		className: "min-w-0 p-density-4",
		"data-step-section": "outputs",
		children: /* @__PURE__ */ (0, V.jsx)(et, { data: e })
	});
}
function Mm({ result: e }) {
	if (e.kind === "none") return null;
	if (e.kind === "invalid") return /* @__PURE__ */ (0, V.jsx)("div", {
		"data-step-source": "",
		children: /* @__PURE__ */ (0, V.jsx)(em, {
			name: "source",
			title: "Source",
			children: /* @__PURE__ */ (0, V.jsxs)(Nm, { children: ["Malformed source detail: ", e.error] })
		})
	});
	let { source: t } = e, n = /* @__PURE__ */ (0, V.jsxs)("span", {
		className: "font-mono text-[11px] text-muted-foreground",
		title: t.file,
		children: [
			xm(t.file),
			":",
			t.line,
			"-",
			t.endLine
		]
	});
	return /* @__PURE__ */ (0, V.jsx)("div", {
		"data-step-source": "",
		children: /* @__PURE__ */ (0, V.jsxs)(em, {
			name: "source",
			title: "Source",
			aside: n,
			children: [t.error ? /* @__PURE__ */ (0, V.jsxs)("div", {
				className: i("rounded border px-2 py-1 text-xs", $p),
				children: ["Source unavailable: ", t.error]
			}) : /* @__PURE__ */ (0, V.jsx)(Ve, {
				language: t.language,
				source: t.text ?? "",
				copyable: !0
			}), t.moreLines ? /* @__PURE__ */ (0, V.jsxs)("div", {
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
function Nm({ children: e }) {
	return /* @__PURE__ */ (0, V.jsx)("div", {
		role: "alert",
		className: "rounded border border-red-500/60 bg-red-500/10 px-2 py-1 text-xs text-red-700 dark:text-red-400",
		children: e
	});
}
function Pm({ error: e }) {
	return /* @__PURE__ */ (0, V.jsxs)(Nm, { children: ["Malformed step detail: ", e] });
}
// ../../../clicky-ui/packages/ui/src/data/verification/stepSource.ts
var Fm = (e) => typeof e == "object" && !!e && !Array.isArray(e), Im = class extends Error {};
function Lm(e, t, n, r) {
	let i = e[t];
	if (!n(i)) throw new Im(`detail.source.${t} must be ${r}`);
	return i;
}
function Rm(e, t, n, r) {
	return e[t] === void 0 ? void 0 : Lm(e, t, n, r);
}
var zm = (e) => typeof e == "string", Bm = (e) => typeof e == "number" && Number.isFinite(e);
function Vm(e) {
	if (!Fm(e) || e.source === void 0) return { kind: "none" };
	let t = e.source;
	if (!Fm(t)) return {
		kind: "invalid",
		error: "detail.source must be an object"
	};
	try {
		let e = {
			file: Lm(t, "file", zm, "a string"),
			line: Lm(t, "line", Bm, "a number"),
			endLine: Lm(t, "end_line", Bm, "a number")
		}, n = {
			language: Rm(t, "language", zm, "a string"),
			text: Rm(t, "text", zm, "a string"),
			moreLines: Rm(t, "more_lines", Bm, "a number"),
			error: Rm(t, "error", zm, "a string")
		};
		for (let [t, r] of Object.entries(n)) r !== void 0 && (e[t] = r);
		return {
			kind: "source",
			source: e
		};
	} catch (e) {
		if (e instanceof Im) return {
			kind: "invalid",
			error: e.message
		};
		throw e;
	}
}
function Hm(e) {
	if (!Fm(e) || !("source" in e)) return e;
	let { source: t, ...n } = e;
	return Object.keys(n).length > 0 ? n : void 0;
}
var Um = {
	yaml: "YAML",
	markdown: "Markdown",
	json: "JSON"
};
function Wm(e) {
	let t = e.kind === "source" ? e.source.language : void 0;
	return t && Um[t] || "Source";
}
function Gm({ file: e, line: t, endLine: n }) {
	return `${e}:${t}-${n}`;
}
function Km(e) {
	let t = e.error ? `source unavailable: ${e.error}` : e.text ?? "", n = e.moreLines ? `\n… ${e.moreLines} more lines` : "";
	return [{
		title: `Source ${Gm(e)}`,
		...e.language && !e.error ? { lang: e.language } : {},
		body: t + n
	}];
}
// ../../../clicky-ui/packages/ui/src/data/verification/adapterViews.tsx
function qm(e) {
	return e.failed ? "failed" : e.passed ? "passed" : "pending";
}
function Jm({ item: e }) {
	return /* @__PURE__ */ (0, V.jsxs)("tr", {
		className: "align-top",
		"data-checklist-item": "",
		children: [/* @__PURE__ */ (0, V.jsx)("td", {
			className: "px-2 py-1.5",
			children: /* @__PURE__ */ (0, V.jsxs)("span", {
				className: "flex items-start gap-1.5",
				children: [/* @__PURE__ */ (0, V.jsx)(Et, {
					outcome: qm(e),
					className: "mt-0.5 text-sm"
				}), /* @__PURE__ */ (0, V.jsx)("span", {
					className: "min-w-0",
					children: e.name
				})]
			})
		}), /* @__PURE__ */ (0, V.jsx)("td", {
			className: "px-2 py-1.5 text-muted-foreground",
			children: e.message
		})]
	});
}
function Ym({ node: e }) {
	let t = e.children && e.children.length > 0 ? e.children : [e];
	return /* @__PURE__ */ (0, V.jsx)("div", {
		className: "p-density-4",
		children: /* @__PURE__ */ (0, V.jsx)("div", {
			className: "overflow-hidden rounded-md border border-border/60",
			children: /* @__PURE__ */ (0, V.jsx)("table", {
				className: "w-full border-collapse text-xs",
				children: /* @__PURE__ */ (0, V.jsx)("tbody", {
					className: "divide-y divide-border/60",
					children: t.map((e, t) => /* @__PURE__ */ (0, V.jsx)(Jm, { item: e }, e.task_id ?? t))
				})
			})
		})
	});
}
function Xm(e) {
	if (e && typeof e == "object" && !Array.isArray(e)) {
		let t = e.cel_trace;
		if (typeof t == "string" && t.length > 0) return t;
	}
}
function Zm(e, t) {
	if (t !== void 0 && (e || typeof t != "object" || !t || Array.isArray(t))) return t;
}
function Qm({ node: e }) {
	if (e.passed) return null;
	let t = e.context, n = t?.cel_vars ?? {}, r = Xm(e.detail) || t?.cel_expression, i = !!r, a = Zm(i, t?.expected), o = Zm(i, t?.actual), s = a !== void 0 || o !== void 0;
	if (!r && Object.keys(n).length === 0 && !s) return null;
	let c = gp(e.detail), l = c.kind === "step" ? c.step.wait : void 0, u = l && l.outcome !== "satisfied" ? Wp(l) : "expression";
	return /* @__PURE__ */ (0, V.jsx)(Hd, {
		check: r,
		checkLabel: u,
		vars: n,
		expected: a,
		actual: o
	});
}
function $m(...e) {
	let t = e.filter((e) => !!e);
	return t.filter((e, n) => !t.some((t, r) => r !== n && t.includes(e) && (t.length > e.length || r < n)));
}
function eh({ node: e, cause: t, error: n }) {
	let r = $m(t, e.message, n);
	return r.length === 0 ? null : e.failed || e.timed_out || t ? /* @__PURE__ */ (0, V.jsx)(wm, { children: r.map((e, t) => /* @__PURE__ */ (0, V.jsx)("div", {
		className: t > 0 ? "mt-1 text-xs opacity-90" : void 0,
		children: e
	}, t)) }) : /* @__PURE__ */ (0, V.jsx)("p", {
		className: "whitespace-pre-wrap break-words text-sm",
		children: r.join("\n")
	});
}
function th({ node: e, renderWait: t, renderPendingPrompt: n }) {
	let r = e.context, i = r?.command ?? e.command, a = r?.cwd ?? e.work_dir, o = gp(e.detail), s = Tp(e.detail), c = Fp(e.detail), { header: l, rest: u } = Up(e.detail), d = Ip(Dp(Hm(o.kind === "step" ? Up(o.step.rest).rest : u))), f = e.waiting === !0 && n ? n(e) : null;
	return /* @__PURE__ */ (0, V.jsxs)("div", {
		className: "space-y-4 p-density-4",
		"data-fixture-detail": "",
		children: [
			/* @__PURE__ */ (0, V.jsx)(Cm, {
				header: l,
				...o.kind === "step" ? { stepKind: o.step.step.kind } : {}
			}),
			f != null && /* @__PURE__ */ (0, V.jsx)(om, { children: f }),
			/* @__PURE__ */ (0, V.jsx)(eh, {
				node: e,
				cause: o.kind === "step" ? o.step.wait?.cause : void 0,
				error: l.error
			}),
			o.kind === "step" && o.step.step.kind === "test" && /* @__PURE__ */ (0, V.jsx)(Em, { node: e }),
			c.kind === "query" && /* @__PURE__ */ (0, V.jsx)(cm, { query: c.query }),
			c.kind === "invalid" && /* @__PURE__ */ (0, V.jsx)(Pm, { error: c.error }),
			i && c.kind !== "query" && /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "min-w-0 space-y-1",
				"data-step-section": "command",
				children: [
					/* @__PURE__ */ (0, V.jsx)("div", {
						className: "text-[10px] uppercase tracking-wide text-muted-foreground",
						children: "Command"
					}),
					/* @__PURE__ */ (0, V.jsx)("code", {
						className: "block overflow-x-auto whitespace-pre rounded bg-muted px-2 py-1.5 text-xs",
						children: i
					}),
					/* @__PURE__ */ (0, V.jsxs)("div", {
						className: "flex flex-wrap items-center gap-2 text-[10px] text-muted-foreground",
						children: [a && /* @__PURE__ */ (0, V.jsxs)("span", { children: ["in ", a] }), typeof r?.exit_code == "number" && /* @__PURE__ */ (0, V.jsxs)("span", {
							className: r.exit_code === 0 ? void 0 : "text-red-600",
							children: ["exit ", r.exit_code]
						})]
					})
				]
			}),
			o.kind === "step" && o.step.view && /* @__PURE__ */ (0, V.jsxs)("div", {
				className: "min-w-0 space-y-1",
				"data-step-section": "view",
				children: [/* @__PURE__ */ (0, V.jsx)("div", {
					className: "text-[10px] uppercase tracking-wide text-muted-foreground",
					children: "Response"
				}), /* @__PURE__ */ (0, V.jsx)(Vs, { data: o.step.view })]
			}),
			o.kind === "step" && o.step.form && /* @__PURE__ */ (0, V.jsx)(am, { form: o.step.form }),
			s.kind === "violation" && s.violation.locations.length > 0 && /* @__PURE__ */ (0, V.jsx)("div", {
				className: "min-w-0",
				"data-step-section": "locations",
				children: /* @__PURE__ */ (0, V.jsx)(Df, {
					locations: s.violation.locations,
					workDir: e.work_dir
				})
			}),
			s.kind === "invalid" && /* @__PURE__ */ (0, V.jsx)(Pm, { error: s.error }),
			/* @__PURE__ */ (0, V.jsx)(Kd, { node: e }),
			/* @__PURE__ */ (0, V.jsx)(Qm, { node: e }),
			(e.failed || e.pending) && /* @__PURE__ */ (0, V.jsx)(Mm, { result: Vm(e.detail) }),
			o.kind === "step" && /* @__PURE__ */ (0, V.jsx)(Am, {
				step: o.step,
				...t ? { renderWait: t } : {}
			}),
			o.kind === "invalid" && /* @__PURE__ */ (0, V.jsx)(Pm, { error: o.error }),
			d != null && /* @__PURE__ */ (0, V.jsxs)("details", {
				className: "text-muted-foreground",
				"data-raw-detail": "",
				children: [/* @__PURE__ */ (0, V.jsx)("summary", {
					className: "cursor-pointer text-[10px] uppercase tracking-wide",
					children: "Raw detail"
				}), /* @__PURE__ */ (0, V.jsx)("div", {
					className: "mt-1 text-foreground",
					children: /* @__PURE__ */ (0, V.jsx)(et, { data: d })
				})]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/verification/adapters.tsx
function nh(e) {
	return e.framework === "checklist";
}
function rh(e) {
	return e.failed ? "[ ]" : e.passed ? "[x]" : "[?]";
}
function ih(e) {
	return { brief: [{
		title: "Criteria",
		body: (e.children && e.children.length > 0 ? e.children : [e]).map((e) => `- ${rh(e)} ${e.name}${e.message ? ` — ${e.message}` : ""}`).join("\n")
	}] };
}
var ah = {
	id: "verification-checklist",
	match: nh,
	renderDetail: ({ node: e }) => /* @__PURE__ */ (0, V.jsx)(Ym, { node: e }),
	copy: ({ node: e }) => ih(e)
};
function oh(e) {
	return e.context !== void 0 || Bp(e.detail) ? !0 : e.framework === "fixture" && (e.children?.length ?? 0) === 0;
}
function sh(e) {
	if (e && typeof e == "object" && !Array.isArray(e)) {
		let t = e.cel_trace;
		if (typeof t == "string" && t.length > 0) return t;
	}
}
function ch(e, t) {
	if (t !== void 0 && (e || typeof t != "object" || !t || Array.isArray(t))) return t;
}
function lh(e) {
	return typeof e == "string" ? e : JSON.stringify(e, null, 2);
}
function uh(e, t) {
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
function dh(e, t) {
	let n = t?.cel_expression, r = sh(e.detail), i = !!(n || r), a = t?.cel_vars ?? {}, o = ch(i, t?.expected), s = ch(i, t?.actual), c = [
		n && `expression: ${n}`,
		r && `trace:\n${r}`,
		Object.keys(a).length > 0 && `vars: ${lh(a)}`,
		o !== void 0 && `expected: ${lh(o)}`,
		s !== void 0 && `actual: ${lh(s)}`
	].filter(Boolean);
	return c.length > 0 ? [{
		title: "Assertion",
		body: c.join("\n")
	}] : [];
}
var fh = {
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
function ph(e) {
	let t = kt(e), n = e.context?.exit_code, r = e.message ?? (n ? `exit ${n}` : void 0);
	return `${t ? fh[t] : "·"} ${e.name}${r ? ` — ${r}` : ""}`;
}
function mh(e) {
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
function hh(e) {
	let t = [...e.summary ? [mh(e.summary)] : [], ...(e.children ?? []).map(ph)];
	return t.length > 0 ? [{
		title: "Results",
		body: t.join("\n")
	}] : [];
}
function gh(e) {
	let t = e.context, n = t?.command ?? e.command, r = gp(e.detail), i = Vm(e.detail), a = Fp(e.detail);
	return {
		brief: [
			...a.kind === "query" ? Lp(a.query) : uh(e, t),
			...dh(e, t),
			...i.kind === "source" ? Km(i.source) : [],
			...r.kind === "step" ? Qp(r.step) : [],
			...hh(e)
		],
		items: n ? [{
			label: "Copy command",
			icon: de,
			value: n
		}] : []
	};
}
var _h = /* @__PURE__ */ new Set([
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
function vh(e) {
	let t = e.framework ?? "";
	return t === "playwright" ? Bp(e.detail) : _h.has(t);
}
var yh = {
	id: "verification-step-copy",
	match: vh,
	copy: ({ node: e }) => gh(e)
}, bh = {
	id: "verification-step-badge",
	match: (e) => e.waiting === !0 || zp(e) !== void 0,
	renderRowLeading: ({ node: e }) => /* @__PURE__ */ (0, V.jsx)(im, { node: e })
};
function xh(e, t, n) {
	let r = {
		...t,
		name: e.name,
		...e.task_id === void 0 ? {} : { task_id: e.task_id },
		waiting: !0
	};
	return oh(r) ? r : {
		...r,
		detail: {
			...t.detail,
			step: n
		}
	};
}
function Sh(e = {}) {
	return bu([
		ah,
		{
			id: "verification-fixture",
			match: oh,
			renderDetail: ({ node: t }) => /* @__PURE__ */ (0, V.jsx)(th, {
				node: t,
				...e
			}),
			detailAttempts: ({ node: e }) => {
				let t = gp(e.detail);
				if (t.kind !== "step") return [];
				let n = t.step.attempts.length - 1;
				return t.step.attempts.map((r, i) => {
					let a = r.result ? jf(r.result) : {
						name: e.name,
						message: `Attempt ${r.attempt}: execution evidence was not stored in this report.`,
						detail: { step: t.step.step }
					}, o = e.waiting === !0 && i === n;
					return i === n && e.log !== void 0 && !e.running && (a.log = e.log), {
						id: r.attempt,
						status: r.status,
						...r.durationMs === void 0 ? {} : { durationMs: r.durationMs },
						node: o ? xh(e, a, t.step.step) : a
					};
				});
			},
			detailTabs: ({ node: e }) => {
				let t = [], n = gp(e.detail);
				n.kind === "step" && n.step.outputs !== void 0 && t.push({
					id: "outputs",
					label: "Outputs",
					render: () => /* @__PURE__ */ (0, V.jsx)(jm, { outputs: n.step.outputs })
				});
				let r = Vm(e.detail);
				return r.kind !== "none" && t.push({
					id: "source",
					label: Wm(r),
					render: () => /* @__PURE__ */ (0, V.jsx)(Mm, { result: r })
				}), t;
			},
			copy: ({ node: e }) => gh(e)
		},
		yh,
		bh
	]);
}
export { vn as A, an as B, Pn as C, xn as D, Sn as E, fn as F, H, dn as I, un as L, gn as M, mn as N, bn as O, pn as P, ln as R, Rn as S, wn as T, nn as V, Vs as _, yf as a, Ga as b, qd as c, Ld as d, wu as f, Zl as g, eu as h, bf as i, _n as j, yn as k, Wd as l, bu as m, Pf as n, Zd as o, Cu as p, Df as r, Jd as s, Sh as t, Kd as u, ts as v, Tn as w, oi as x, oo as y, on as z };
