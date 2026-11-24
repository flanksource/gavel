import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-a1ZmknZu.js";
import { t as r } from "./react-dom-CxaMoh3H.js";
import { l as i, s as a } from "./use-theme-DIoUqVj6.js";
import { n as o, s, t as c } from "./Icon-DPM7-YUX.js";
import { $ as l, A as u, B as d, C as f, D as p, E as m, F as h, H as g, I as _, J as v, L as y, M as b, N as x, O as S, P as C, Q as w, R as T, S as E, T as D, U as O, V as k, W as A, X as j, Y as M, Z as N, _ as P, a as F, b as I, c as L, d as R, et as z, f as B, g as V, h as H, j as U, k as ee, l as te, m as W, o as ne, p as re, s as ie, t as ae, tt as oe, u as G, v as se, w as K, x as ce, y as le, z as ue } from "./Callout-BvAoqcdf.js";
import { n as de, o as q, t as fe } from "./button-Dbm88adW.js";
import { i as pe, n as me, r as he, t as ge } from "./Badge-BcEdZ4Rn.js";
import { t as _e } from "./dist-CMVKLuPp.js";
import { _ as ve, a as ye, b as be, c as xe, d as Se, f as Ce, g as we, h as Te, l as Ee, m as De, n as Oe, o as ke, p as Ae, s as je, u as Me, v as Ne, y as Pe } from "./floating-ui.react-dom-BTsb3rDV.js";
// ../../../clicky-ui/packages/ui/src/hooks/use-hotkey.ts
var Fe = /* @__PURE__ */ e(r(), 1), J = /* @__PURE__ */ e(n(), 1), Y = /* @__PURE__ */ e(t(), 1), Ie = [], Le = null;
function Re(e) {
	let t = e.split("+").map((e) => e.trim().toLowerCase()).filter(Boolean), n = t[t.length - 1];
	if (!n) throw Error(`useHotkey: empty combo "${e}"`);
	return {
		key: n,
		mod: t.includes("mod") || t.includes("cmd") || t.includes("ctrl"),
		shift: t.includes("shift"),
		alt: t.includes("alt") || t.includes("option")
	};
}
function ze(e, t) {
	return e.key.toLowerCase() === t.key && t.mod === (e.metaKey || e.ctrlKey) && t.shift === e.shiftKey && t.alt === e.altKey;
}
function Be(e) {
	return e instanceof HTMLElement ? e.isContentEditable ? !0 : [
		"INPUT",
		"TEXTAREA",
		"SELECT"
	].includes(e.tagName) : !1;
}
function Ve(e) {
	return e.composedPath?.()[0] ?? e.target;
}
function He(e) {
	if (e.defaultPrevented || e.isComposing) return;
	let t = Ve(e), n;
	for (let r of Ie) ze(e, r.combo) && (r.enableOnFormElements || !Be(t)) && (!n || r.priority >= n.priority) && (n = r);
	n && (e.preventDefault(), n.run(e));
}
function Ue() {
	Le || typeof document > "u" || (Le = document, Le.addEventListener("keydown", He));
}
function We(e) {
	let t = Ie.findIndex((t) => t.id === e);
	t !== -1 && Ie.splice(t, 1), Ie.length === 0 && Le && (Le.removeEventListener("keydown", He), Le = null);
}
function Ge(e, t, n = {}) {
	let r = (0, Y.useId)(), i = (0, Y.useRef)(t);
	i.current = t;
	let { enabled: a = !0, priority: o = 0, enableOnFormElements: s } = n;
	(0, Y.useEffect)(() => {
		if (!e || !a) return;
		let t = Re(e);
		return Ue(), Ie.push({
			id: r,
			combo: t,
			priority: o,
			enableOnFormElements: s ?? t.mod,
			run: (e) => i.current(e)
		}), () => We(r);
	}, [
		e,
		a,
		o,
		s,
		r
	]);
}
// ../../../clicky-ui/packages/ui/src/icons/components/UiChevronUp.tsx
var Ke = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M212.24 164.24a6 6 0 0 1-8.48 0L128 88.49l-75.76 75.75a6 6 0 0 1-8.48-8.48l80-80a6 6 0 0 1 8.48 0l80 80a6 6 0 0 1 0 8.48"
	})
}), {
	__source: "ph:caret-up-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "chevron-up",
	displayName: "UiChevronUp"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M215.39 163.06A8 8 0 0 1 208 168H48a8 8 0 0 1-5.66-13.66l80-80a8 8 0 0 1 11.32 0l80 80a8 8 0 0 1 1.73 8.72"
	})
}), {
	__source: "ph:caret-up-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "chevron-up",
	displayName: "UiChevronUpFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiEllipsisBold.tsx
var qe = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M144 128a16 16 0 1 1-16-16a16 16 0 0 1 16 16m-84-16a16 16 0 1 0 16 16a16 16 0 0 0-16-16m136 0a16 16 0 1 0 16 16a16 16 0 0 0-16-16"
	})
}), {
	__source: "ph:dots-three-bold",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "ellipsis-bold",
	displayName: "UiEllipsisBold"
}), Je = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			d: "M4.5 2.5L8 6L11.5 2.5",
			stroke: "#6C707E",
			strokeLinecap: "round",
			strokeLinejoin: "round"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M4.5 13.5L8 10L11.5 13.5",
			stroke: "#6C707E",
			strokeLinecap: "round",
			strokeLinejoin: "round"
		})]
	})
}), {
	__source: "jb-expui-general:collapseAll",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "collapse-all",
	displayName: "UiCollapseAll"
}), Ye = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			d: "M4.5 5.5L8 2L11.5 5.5",
			stroke: "#6C707E",
			strokeLinecap: "round",
			strokeLinejoin: "round"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M4.5 10.5L8 14L11.5 10.5",
			stroke: "#6C707E",
			strokeLinecap: "round",
			strokeLinejoin: "round"
		})]
	})
}), {
	__source: "jb-expui-general:expandAll",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "expand-all",
	displayName: "UiExpandAll"
}), Xe = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "m228.24 219.76l-51.38-51.38a86.15 86.15 0 1 0-8.48 8.48l51.38 51.38a6 6 0 0 0 8.48-8.48M38 112a74 74 0 1 1 74 74a74.09 74.09 0 0 1-74-74"
	})
}), {
	__source: "ph:magnifying-glass-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "search",
	displayName: "UiSearch"
}), X = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M128 26a102 102 0 1 0 102 102A102.12 102.12 0 0 0 128 26m0 192a90 90 0 1 1 90-90a90.1 90.1 0 0 1-90 90m46-90a6 6 0 0 1-6 6h-34v34a6 6 0 0 1-12 0v-34H88a6 6 0 0 1 0-12h34V88a6 6 0 0 1 12 0v34h34a6 6 0 0 1 6 6"
	})
}), {
	__source: "ph:plus-circle-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "add",
	displayName: "UiAdd"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M128 24a104 104 0 1 0 104 104A104.13 104.13 0 0 0 128 24m40 112h-32v32a8 8 0 0 1-16 0v-32H88a8 8 0 0 1 0-16h32V88a8 8 0 0 1 16 0v32h32a8 8 0 0 1 0 16"
	})
}), {
	__source: "ph:plus-circle-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "add",
	displayName: "UiAddFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiRemove.tsx
var Ze = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiEye.tsx
var Qe = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M245.48 125.57c-.34-.78-8.66-19.23-27.24-37.81C201 70.54 171.38 50 128 50S55 70.54 37.76 87.76c-18.58 18.58-26.9 37-27.24 37.81a6 6 0 0 0 0 4.88c.34.77 8.66 19.22 27.24 37.8C55 185.47 84.62 206 128 206s73-20.53 90.24-37.75c18.58-18.58 26.9-37 27.24-37.8a6 6 0 0 0 0-4.88M128 194c-31.38 0-58.78-11.42-81.45-33.93A134.8 134.8 0 0 1 22.69 128a134.6 134.6 0 0 1 23.86-32.06C69.22 73.42 96.62 62 128 62s58.78 11.42 81.45 33.94A134.6 134.6 0 0 1 233.31 128C226.94 140.21 195 194 128 194m0-112a46 46 0 1 0 46 46a46.06 46.06 0 0 0-46-46m0 80a34 34 0 1 1 34-34a34 34 0 0 1-34 34"
	})
}), {
	__source: "ph:eye-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "eye",
	displayName: "UiEye"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M247.31 124.76c-.35-.79-8.82-19.58-27.65-38.41C194.57 61.26 162.88 48 128 48S61.43 61.26 36.34 86.35C17.51 105.18 9 124 8.69 124.76a8 8 0 0 0 0 6.5c.35.79 8.82 19.57 27.65 38.4C61.43 194.74 93.12 208 128 208s66.57-13.26 91.66-38.34c18.83-18.83 27.3-37.61 27.65-38.4a8 8 0 0 0 0-6.5M128 168a40 40 0 1 1 40-40a40 40 0 0 1-40 40"
	})
}), {
	__source: "ph:eye-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "eye",
	displayName: "UiEyeFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiEyeClosed.tsx
var $e = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiFullscreen.tsx
var et = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M214 48v40a6 6 0 0 1-12 0V54h-34a6 6 0 0 1 0-12h40a6 6 0 0 1 6 6M88 202H54v-34a6 6 0 0 0-12 0v40a6 6 0 0 0 6 6h40a6 6 0 0 0 0-12m120-40a6 6 0 0 0-6 6v34h-34a6 6 0 0 0 0 12h40a6 6 0 0 0 6-6v-40a6 6 0 0 0-6-6M88 42H48a6 6 0 0 0-6 6v40a6 6 0 0 0 12 0V54h34a6 6 0 0 0 0-12"
	})
}), {
	__source: "ph:corners-out-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "fullscreen",
	displayName: "UiFullscreen"
}), tt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M93.66 202.34A8 8 0 0 1 88 216H48a8 8 0 0 1-8-8v-40a8 8 0 0 1 13.66-5.66ZM88 40H48a8 8 0 0 0-8 8v40a8 8 0 0 0 13.66 5.66l40-40A8 8 0 0 0 88 40m123.06 120.61a8 8 0 0 0-8.72 1.73l-40 40A8 8 0 0 0 168 216h40a8 8 0 0 0 8-8v-40a8 8 0 0 0-4.94-7.39M208 40h-40a8 8 0 0 0-5.66 13.66l40 40A8 8 0 0 0 216 88V48a8 8 0 0 0-8-8"
	})
}), {
	__source: "ph:corners-out-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "fullscreen",
	displayName: "UiFullscreenFilled"
}), nt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M232.13 143.64a6 6 0 0 0-6-1.49a90.07 90.07 0 0 1-112.27-112.3a6 6 0 0 0-7.49-7.48a102.88 102.88 0 0 0-51.89 36.31a102 102 0 0 0 142.84 142.84a102.88 102.88 0 0 0 36.31-51.89a6 6 0 0 0-1.5-5.99m-42 48.29a90 90 0 0 1-126-126a90.9 90.9 0 0 1 35.52-28.27a102.06 102.06 0 0 0 118.69 118.69a90.9 90.9 0 0 1-28.24 35.58Z"
	})
}), {
	__source: "ph:moon-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "moon",
	displayName: "UiMoon"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M235.54 150.21a104.84 104.84 0 0 1-37 52.91A104 104 0 0 1 32 120a103.1 103.1 0 0 1 20.88-62.52a104.84 104.84 0 0 1 52.91-37a8 8 0 0 1 10 10a88.08 88.08 0 0 0 109.8 109.8a8 8 0 0 1 10 10Z"
	})
}), {
	__source: "ph:moon-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "moon",
	displayName: "UiMoonFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiSun.tsx
var rt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M122 40V16a6 6 0 0 1 12 0v24a6 6 0 0 1-12 0m68 88a62 62 0 1 1-62-62a62.07 62.07 0 0 1 62 62m-12 0a50 50 0 1 0-50 50a50.06 50.06 0 0 0 50-50M59.76 68.24a6 6 0 1 0 8.48-8.48l-16-16a6 6 0 0 0-8.48 8.48Zm0 119.52l-16 16a6 6 0 1 0 8.48 8.48l16-16a6 6 0 1 0-8.48-8.48M192 70a6 6 0 0 0 4.24-1.76l16-16a6 6 0 0 0-8.48-8.48l-16 16A6 6 0 0 0 192 70m4.24 117.76a6 6 0 0 0-8.48 8.48l16 16a6 6 0 0 0 8.48-8.48ZM46 128a6 6 0 0 0-6-6H16a6 6 0 0 0 0 12h24a6 6 0 0 0 6-6m82 82a6 6 0 0 0-6 6v24a6 6 0 0 0 12 0v-24a6 6 0 0 0-6-6m112-88h-24a6 6 0 0 0 0 12h24a6 6 0 0 0 0-12"
	})
}), {
	__source: "ph:sun-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "sun",
	displayName: "UiSun"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M120 40V16a8 8 0 0 1 16 0v24a8 8 0 0 1-16 0m8 24a64 64 0 1 0 64 64a64.07 64.07 0 0 0-64-64m-69.66 5.66a8 8 0 0 0 11.32-11.32l-16-16a8 8 0 0 0-11.32 11.32Zm0 116.68l-16 16a8 8 0 0 0 11.32 11.32l16-16a8 8 0 0 0-11.32-11.32M192 72a8 8 0 0 0 5.66-2.34l16-16a8 8 0 0 0-11.32-11.32l-16 16A8 8 0 0 0 192 72m5.66 114.34a8 8 0 0 0-11.32 11.32l16 16a8 8 0 0 0 11.32-11.32ZM48 128a8 8 0 0 0-8-8H16a8 8 0 0 0 0 16h24a8 8 0 0 0 8-8m80 80a8 8 0 0 0-8 8v24a8 8 0 0 0 16 0v-24a8 8 0 0 0-8-8m112-88h-24a8 8 0 0 0 0 16h24a8 8 0 0 0 0-16"
	})
}), {
	__source: "ph:sun-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "sun",
	displayName: "UiSunFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiListDashes.tsx
var it = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiHardDrive.tsx
var at = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 24 24",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "none",
		stroke: "currentColor",
		strokeLinecap: "round",
		strokeLinejoin: "round",
		strokeWidth: "2",
		d: "M10 16h.01m-7.798-4.423a2 2 0 0 0-.212.896V18a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-5.527a2 2 0 0 0-.212-.896L18.55 5.11A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11zm19.734.436H2.054M6 16h.01"
	})
}), {
	__source: "lucide:hard-drive",
	__viewBox: "0 0 24 24",
	__group: "ui-controls",
	__consumerName: "hard-drive",
	displayName: "UiHardDrive"
}), ot = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "currentColor",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", { d: "m220 169.09l-92 53.65l-92-53.65a8 8 0 0 0-8 13.82l96 56a8 8 0 0 0 8.06 0l96-56a8 8 0 1 0-8.06-13.82" }),
			/* @__PURE__ */ (0, J.jsx)("path", { d: "m220 121.09l-92 53.65l-92-53.65a8 8 0 0 0-8 13.82l96 56a8 8 0 0 0 8.06 0l96-56a8 8 0 1 0-8.06-13.82" }),
			/* @__PURE__ */ (0, J.jsx)("path", { d: "m28 86.91l96 56a8 8 0 0 0 8.06 0l96-56a8 8 0 0 0 0-13.82l-96-56a8 8 0 0 0-8.06 0l-96 56a8 8 0 0 0 0 13.82" })
		]
	})
}), {
	__source: "ph:stack-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "stack",
	displayName: "UiStackFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiListOrdered.tsx
var st = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 24 24",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "none",
		stroke: "currentColor",
		strokeLinecap: "round",
		strokeLinejoin: "round",
		strokeWidth: "2",
		d: "M11 5h10m-10 7h10m-10 7h10M4 4h1v5M4 9h2m.5 11H3.4c0-1 2.6-1.925 2.6-3.5a1.5 1.5 0 0 0-2.6-1.02"
	})
}), {
	__source: "lucide:list-ordered",
	__viewBox: "0 0 24 24",
	__group: "ui-controls",
	__consumerName: "list-ordered",
	displayName: "UiListOrdered"
}), ct = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M118 60a10 10 0 1 1 10 10a10 10 0 0 1-10-10m10 58a10 10 0 1 0 10 10a10 10 0 0 0-10-10m0 68a10 10 0 1 0 10 10a10 10 0 0 0-10-10"
	})
}), {
	__source: "ph:dots-three-vertical-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "dots-vertical",
	displayName: "UiDotsVertical"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M160 16H96a16 16 0 0 0-16 16v192a16 16 0 0 0 16 16h64a16 16 0 0 0 16-16V32a16 16 0 0 0-16-16m-32 192a12 12 0 1 1 12-12a12 12 0 0 1-12 12m0-68a12 12 0 1 1 12-12a12 12 0 0 1-12 12m0-68a12 12 0 1 1 12-12a12 12 0 0 1-12 12"
	})
}), {
	__source: "ph:dots-three-vertical-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "dots-vertical",
	displayName: "UiDotsVerticalFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiArrowLeft.tsx
var lt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M222 128a6 6 0 0 1-6 6H54.49l61.75 61.76a6 6 0 1 1-8.48 8.48l-72-72a6 6 0 0 1 0-8.48l72-72a6 6 0 0 1 8.48 8.48L54.49 122H216a6 6 0 0 1 6 6"
	})
}), {
	__source: "ph:arrow-left-light",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "arrow-left",
	displayName: "UiArrowLeft"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M224 128a8 8 0 0 1-8 8h-96v64a8 8 0 0 1-13.66 5.66l-72-72a8 8 0 0 1 0-11.32l72-72A8 8 0 0 1 120 56v64h96a8 8 0 0 1 8 8"
	})
}), {
	__source: "ph:arrow-left-fill",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "arrow-left",
	displayName: "UiArrowLeftFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiLink.tsx
var ut = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M238 88.18a52.42 52.42 0 0 1-15.4 35.66l-34.75 34.75A52.28 52.28 0 0 1 150.62 174h-.05A52.63 52.63 0 0 1 98 119.9a6 6 0 0 1 6-5.84h.17a6 6 0 0 1 5.83 6.16A40.62 40.62 0 0 0 150.58 162a40.4 40.4 0 0 0 28.73-11.9l34.75-34.74a40.63 40.63 0 0 0-57.43-57.46l-11 11a6 6 0 0 1-8.49-8.49l11-11a52.62 52.62 0 0 1 74.43 0A52.83 52.83 0 0 1 238 88.18m-127.62 98.9l-11 11A40.36 40.36 0 0 1 70.6 210a40.63 40.63 0 0 1-28.7-69.36l34.72-34.74A40.63 40.63 0 0 1 146 135.77a6 6 0 0 0 5.83 6.16h.17a6 6 0 0 0 6-5.84a52.63 52.63 0 0 0-89.86-38.67l-34.76 34.74A52.63 52.63 0 0 0 70.56 222a52.26 52.26 0 0 0 37.22-15.42l11-11a6 6 0 1 0-8.49-8.48Z"
	})
}), {
	__source: "ph:link-light",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "link",
	displayName: "UiLink"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M208 32H48a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16m-92.3 160.49a43.31 43.31 0 0 1-55-66.43l25.37-25.37a43.35 43.35 0 0 1 61.25 0a42.9 42.9 0 0 1 9.95 15.43a8 8 0 1 1-15 5.6a27.33 27.33 0 0 0-44.9-9.72L72 137.37a27.32 27.32 0 0 0 34.68 41.91a8 8 0 1 1 9 13.21Zm79.61-62.55l-25.37 25.37A43 43 0 0 1 139.32 168a43.35 43.35 0 0 1-40.53-28.12a8 8 0 1 1 15-5.6A27.35 27.35 0 0 0 139.28 152a27.14 27.14 0 0 0 19.32-8l25.4-25.37a27.32 27.32 0 0 0-34.68-41.91a8 8 0 1 1-9-13.21a43.32 43.32 0 0 1 55 66.43Z"
	})
}), {
	__source: "ph:link-fill",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "link",
	displayName: "UiLinkFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiLayoutDashboard.tsx
var dt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M104 42H56a14 14 0 0 0-14 14v48a14 14 0 0 0 14 14h48a14 14 0 0 0 14-14V56a14 14 0 0 0-14-14m2 62a2 2 0 0 1-2 2H56a2 2 0 0 1-2-2V56a2 2 0 0 1 2-2h48a2 2 0 0 1 2 2Zm94-62h-48a14 14 0 0 0-14 14v48a14 14 0 0 0 14 14h48a14 14 0 0 0 14-14V56a14 14 0 0 0-14-14m2 62a2 2 0 0 1-2 2h-48a2 2 0 0 1-2-2V56a2 2 0 0 1 2-2h48a2 2 0 0 1 2 2Zm-98 34H56a14 14 0 0 0-14 14v48a14 14 0 0 0 14 14h48a14 14 0 0 0 14-14v-48a14 14 0 0 0-14-14m2 62a2 2 0 0 1-2 2H56a2 2 0 0 1-2-2v-48a2 2 0 0 1 2-2h48a2 2 0 0 1 2 2Zm94-62h-48a14 14 0 0 0-14 14v48a14 14 0 0 0 14 14h48a14 14 0 0 0 14-14v-48a14 14 0 0 0-14-14m2 62a2 2 0 0 1-2 2h-48a2 2 0 0 1-2-2v-48a2 2 0 0 1 2-2h48a2 2 0 0 1 2 2Z"
	})
}), {
	__source: "ph:squares-four-light",
	__viewBox: "0 0 256 256",
	__group: "layout-dashboard",
	__consumerName: "layout-dashboard",
	displayName: "UiLayoutDashboard"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M120 56v48a16 16 0 0 1-16 16H56a16 16 0 0 1-16-16V56a16 16 0 0 1 16-16h48a16 16 0 0 1 16 16m80-16h-48a16 16 0 0 0-16 16v48a16 16 0 0 0 16 16h48a16 16 0 0 0 16-16V56a16 16 0 0 0-16-16m-96 96H56a16 16 0 0 0-16 16v48a16 16 0 0 0 16 16h48a16 16 0 0 0 16-16v-48a16 16 0 0 0-16-16m96 0h-48a16 16 0 0 0-16 16v48a16 16 0 0 0 16 16h48a16 16 0 0 0 16-16v-48a16 16 0 0 0-16-16"
	})
}), {
	__source: "ph:squares-four-fill",
	__viewBox: "0 0 256 256",
	__group: "layout-dashboard",
	__consumerName: "layout-dashboard",
	displayName: "UiLayoutDashboardFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiEdit.tsx
var ft = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "m225.9 74.78l-44.69-44.69a14 14 0 0 0-19.8 0L38.1 153.41a13.94 13.94 0 0 0-4.1 9.9V208a14 14 0 0 0 14 14h44.69a13.94 13.94 0 0 0 9.9-4.1L225.9 94.58a14 14 0 0 0 0-19.8M94.1 209.41a2 2 0 0 1-1.41.59H48a2 2 0 0 1-2-2v-44.69a2 2 0 0 1 .59-1.41L136 72.48L183.51 120ZM217.41 86.1L192 111.51L144.49 64l25.41-25.42a2 2 0 0 1 2.83 0l44.68 44.69a2 2 0 0 1 0 2.83"
	})
}), {
	__source: "ph:pencil-simple-light",
	__viewBox: "0 0 256 256",
	__group: "forms-editing",
	__consumerName: "edit",
	displayName: "UiEdit"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "m227.31 73.37l-44.68-44.69a16 16 0 0 0-22.63 0L36.69 152A15.86 15.86 0 0 0 32 163.31V208a16 16 0 0 0 16 16h44.69a15.86 15.86 0 0 0 11.31-4.69L227.31 96a16 16 0 0 0 0-22.63M192 108.68L147.31 64l24-24L216 84.68Z"
	})
}), {
	__source: "ph:pencil-simple-fill",
	__viewBox: "0 0 256 256",
	__group: "forms-editing",
	__consumerName: "edit",
	displayName: "UiEditFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiSelect.tsx
var pt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M150 40a6 6 0 0 1-6 6h-32a6 6 0 0 1 0-12h32a6 6 0 0 1 6 6m-6 170h-32a6 6 0 0 0 0 12h32a6 6 0 0 0 0-12m64-176h-24a6 6 0 0 0 0 12h24a2 2 0 0 1 2 2v24a6 6 0 0 0 12 0V48a14 14 0 0 0-14-14m8 72a6 6 0 0 0-6 6v32a6 6 0 0 0 12 0v-32a6 6 0 0 0-6-6m0 72a6 6 0 0 0-6 6v24a2 2 0 0 1-2 2h-24a6 6 0 0 0 0 12h24a14 14 0 0 0 14-14v-24a6 6 0 0 0-6-6M40 150a6 6 0 0 0 6-6v-32a6 6 0 0 0-12 0v32a6 6 0 0 0 6 6m32 60H48a2 2 0 0 1-2-2v-24a6 6 0 0 0-12 0v24a14 14 0 0 0 14 14h24a6 6 0 0 0 0-12m0-176H48a14 14 0 0 0-14 14v24a6 6 0 0 0 12 0V48a2 2 0 0 1 2-2h24a6 6 0 0 0 0-12"
	})
}), {
	__source: "ph:selection-light",
	__viewBox: "0 0 256 256",
	__group: "forms-editing",
	__consumerName: "select",
	displayName: "UiSelect"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M208 32H48a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16M104 200H72a16 16 0 0 1-16-16v-32a8 8 0 0 1 16 0v32h32a8 8 0 0 1 0 16m0-128H72v32a8 8 0 0 1-16 0V72a16 16 0 0 1 16-16h32a8 8 0 0 1 0 16m96 112a16 16 0 0 1-16 16h-32a8 8 0 0 1 0-16h32v-32a8 8 0 0 1 16 0Zm0-80a8 8 0 0 1-16 0V72h-32a8 8 0 0 1 0-16h32a16 16 0 0 1 16 16Z"
	})
}), {
	__source: "ph:selection-fill",
	__viewBox: "0 0 256 256",
	__group: "forms-editing",
	__consumerName: "select",
	displayName: "UiSelectFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiCalendar.tsx
var mt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiCalendarBlank.tsx
var ht = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M208 34h-26V24a6 6 0 0 0-12 0v10H86V24a6 6 0 0 0-12 0v10H48a14 14 0 0 0-14 14v160a14 14 0 0 0 14 14h160a14 14 0 0 0 14-14V48a14 14 0 0 0-14-14M48 46h26v10a6 6 0 0 0 12 0V46h84v10a6 6 0 0 0 12 0V46h26a2 2 0 0 1 2 2v34H46V48a2 2 0 0 1 2-2m160 164H48a2 2 0 0 1-2-2V94h164v114a2 2 0 0 1-2 2"
	})
}), {
	__source: "ph:calendar-blank-light",
	__viewBox: "0 0 256 256",
	__group: "forms-editing",
	__consumerName: "calendar-blank",
	displayName: "UiCalendarBlank"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M208 32h-24v-8a8 8 0 0 0-16 0v8H88v-8a8 8 0 0 0-16 0v8H48a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16m0 48H48V48h24v8a8 8 0 0 0 16 0v-8h80v8a8 8 0 0 0 16 0v-8h24Z"
	})
}), {
	__source: "ph:calendar-blank-fill",
	__viewBox: "0 0 256 256",
	__group: "forms-editing",
	__consumerName: "calendar-blank",
	displayName: "UiCalendarBlankFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiWarningCircle.tsx
var gt = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#d97706",
		...r
	};
	return /* @__PURE__ */ (0, J.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, J.jsx)("path", {
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
}), _t = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#d97706",
		...r
	};
	return /* @__PURE__ */ (0, J.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, J.jsx)("path", {
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
}), vt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M138 180a10 10 0 1 1-10-10a10 10 0 0 1 10 10M128 74c-21 0-38 15.25-38 34v4a6 6 0 0 0 12 0v-4c0-12.13 11.66-22 26-22s26 9.87 26 22s-11.66 22-26 22a6 6 0 0 0-6 6v8a6 6 0 0 0 12 0v-2.42c18.11-2.58 32-16.66 32-33.58c0-18.75-17-34-38-34m102 54A102 102 0 1 1 128 26a102.12 102.12 0 0 1 102 102m-12 0a90 90 0 1 0-90 90a90.1 90.1 0 0 0 90-90"
	})
}), {
	__source: "ph:question-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "question",
	displayName: "UiQuestion"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M128 24a104 104 0 1 0 104 104A104.11 104.11 0 0 0 128 24m0 168a12 12 0 1 1 12-12a12 12 0 0 1-12 12m8-48.72v.72a8 8 0 0 1-16 0v-8a8 8 0 0 1 8-8c13.23 0 24-9 24-20s-10.77-20-24-20s-24 9-24 20v4a8 8 0 0 1-16 0v-4c0-19.85 17.94-36 40-36s40 16.15 40 36c0 17.38-13.76 31.93-32 35.28"
	})
}), {
	__source: "ph:question-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "question",
	displayName: "UiQuestionFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiCircleOutline.tsx
var yt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M128 26a102 102 0 1 0 102 102A102.12 102.12 0 0 0 128 26m0 192a90 90 0 1 1 90-90a90.1 90.1 0 0 1-90 90"
	})
}), {
	__source: "ph:circle-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "circle-outline",
	displayName: "UiCircleOutline"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M232 128A104 104 0 1 1 128 24a104.13 104.13 0 0 1 104 104"
	})
}), {
	__source: "ph:circle-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "circle-outline",
	displayName: "UiCircleOutlineFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiCheck.tsx
var bt = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#059669",
		...r
	};
	return /* @__PURE__ */ (0, J.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, J.jsx)("path", {
			fill: "currentColor",
			d: "m228.24 76.24l-128 128a6 6 0 0 1-8.48 0l-56-56a6 6 0 0 1 8.48-8.48L96 191.51L219.76 67.76a6 6 0 0 1 8.48 8.48"
		})
	});
}, {
	__source: "ph:check-light",
	__viewBox: "0 0 256 256",
	__group: "approval-review",
	__consumerName: "check",
	__defaultColor: "#059669",
	displayName: "UiCheck"
}), xt = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#059669",
		...r
	};
	return /* @__PURE__ */ (0, J.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, J.jsx)("path", {
			fill: "currentColor",
			d: "M128 24a104 104 0 1 0 104 104A104.11 104.11 0 0 0 128 24m45.66 85.66l-56 56a8 8 0 0 1-11.32 0l-24-24a8 8 0 0 1 11.32-11.32L112 148.69l50.34-50.35a8 8 0 0 1 11.32 11.32"
		})
	});
}, {
	__source: "ph:check-circle-fill",
	__viewBox: "0 0 256 256",
	__group: "approval-review",
	__consumerName: "check",
	__defaultColor: "#059669",
	displayName: "UiCheckFilled"
}), St = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M208 82h-34V56a46 46 0 0 0-92 0v26H48a14 14 0 0 0-14 14v112a14 14 0 0 0 14 14h160a14 14 0 0 0 14-14V96a14 14 0 0 0-14-14M94 56a34 34 0 0 1 68 0v26H94Zm116 152a2 2 0 0 1-2 2H48a2 2 0 0 1-2-2V96a2 2 0 0 1 2-2h160a2 2 0 0 1 2 2Zm-72-56a10 10 0 1 1-10-10a10 10 0 0 1 10 10"
	})
}), {
	__source: "ph:lock-light",
	__viewBox: "0 0 256 256",
	__group: "approval-review",
	__consumerName: "lock",
	displayName: "UiLock"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M208 80h-32V56a48 48 0 0 0-96 0v24H48a16 16 0 0 0-16 16v112a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V96a16 16 0 0 0-16-16m-80 84a12 12 0 1 1 12-12a12 12 0 0 1-12 12m32-84H96V56a32 32 0 0 1 64 0Z"
	})
}), {
	__source: "ph:lock-fill",
	__viewBox: "0 0 256 256",
	__group: "approval-review",
	__consumerName: "lock",
	displayName: "UiLockFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiLockOpen.tsx
var Ct = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M208 82H94V56a34 34 0 0 1 34-34c16.3 0 31 11.69 34.12 27.19a6 6 0 0 0 11.76-2.38C169.55 25.48 150.26 10 128 10a46.06 46.06 0 0 0-46 46v26H48a14 14 0 0 0-14 14v112a14 14 0 0 0 14 14h160a14 14 0 0 0 14-14V96a14 14 0 0 0-14-14m2 126a2 2 0 0 1-2 2H48a2 2 0 0 1-2-2V96a2 2 0 0 1 2-2h160a2 2 0 0 1 2 2Zm-72-56a10 10 0 1 1-10-10a10 10 0 0 1 10 10"
	})
}), {
	__source: "ph:lock-open-light",
	__viewBox: "0 0 256 256",
	__group: "security-auth",
	__consumerName: "lock-open",
	displayName: "UiLockOpen"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M208 80H96V56a32 32 0 0 1 32-32c15.37 0 29.2 11 32.16 25.59a8 8 0 0 0 15.68-3.18C171.32 24.15 151.2 8 128 8a48.05 48.05 0 0 0-48 48v24H48a16 16 0 0 0-16 16v112a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V96a16 16 0 0 0-16-16m-80 84a12 12 0 1 1 12-12a12 12 0 0 1-12 12"
	})
}), {
	__source: "ph:lock-open-fill",
	__viewBox: "0 0 256 256",
	__group: "security-auth",
	__consumerName: "lock-open",
	displayName: "UiLockOpenFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiListFlat.tsx
var wt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiListTree.tsx
var Tt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M160 110h48a14 14 0 0 0 14-14V48a14 14 0 0 0-14-14h-48a14 14 0 0 0-14 14v18h-18a22 22 0 0 0-22 22v34H70v-10a14 14 0 0 0-14-14H24a14 14 0 0 0-14 14v32a14 14 0 0 0 14 14h32a14 14 0 0 0 14-14v-10h36v34a22 22 0 0 0 22 22h18v18a14 14 0 0 0 14 14h48a14 14 0 0 0 14-14v-48a14 14 0 0 0-14-14h-48a14 14 0 0 0-14 14v18h-18a10 10 0 0 1-10-10V88a10 10 0 0 1 10-10h18v18a14 14 0 0 0 14 14M58 144a2 2 0 0 1-2 2H24a2 2 0 0 1-2-2v-32a2 2 0 0 1 2-2h32a2 2 0 0 1 2 2Zm100 16a2 2 0 0 1 2-2h48a2 2 0 0 1 2 2v48a2 2 0 0 1-2 2h-48a2 2 0 0 1-2-2Zm0-112a2 2 0 0 1 2-2h48a2 2 0 0 1 2 2v48a2 2 0 0 1-2 2h-48a2 2 0 0 1-2-2Z"
	})
}), {
	__source: "ph:tree-structure-light",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "list-tree",
	displayName: "UiListTree"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M144 96V80h-16a8 8 0 0 0-8 8v80a8 8 0 0 0 8 8h16v-16a16 16 0 0 1 16-16h48a16 16 0 0 1 16 16v48a16 16 0 0 1-16 16h-48a16 16 0 0 1-16-16v-16h-16a24 24 0 0 1-24-24v-32H72v8a16 16 0 0 1-16 16H24a16 16 0 0 1-16-16v-32a16 16 0 0 1 16-16h32a16 16 0 0 1 16 16v8h32V88a24 24 0 0 1 24-24h16V48a16 16 0 0 1 16-16h48a16 16 0 0 1 16 16v48a16 16 0 0 1-16 16h-48a16 16 0 0 1-16-16"
	})
}), {
	__source: "ph:tree-structure-fill",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "list-tree",
	displayName: "UiListTreeFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiRows.tsx
var Et = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M208 138H48a14 14 0 0 0-14 14v40a14 14 0 0 0 14 14h160a14 14 0 0 0 14-14v-40a14 14 0 0 0-14-14m2 54a2 2 0 0 1-2 2H48a2 2 0 0 1-2-2v-40a2 2 0 0 1 2-2h160a2 2 0 0 1 2 2Zm-2-142H48a14 14 0 0 0-14 14v40a14 14 0 0 0 14 14h160a14 14 0 0 0 14-14V64a14 14 0 0 0-14-14m2 54a2 2 0 0 1-2 2H48a2 2 0 0 1-2-2V64a2 2 0 0 1 2-2h160a2 2 0 0 1 2 2Z"
	})
}), {
	__source: "ph:rows-light",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "rows",
	displayName: "UiRows"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M224 152v40a16 16 0 0 1-16 16H48a16 16 0 0 1-16-16v-40a16 16 0 0 1 16-16h160a16 16 0 0 1 16 16M208 48H48a16 16 0 0 0-16 16v40a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V64a16 16 0 0 0-16-16"
	})
}), {
	__source: "ph:rows-fill",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "rows",
	displayName: "UiRowsFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiToggleOff.tsx
var Dt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M176 58H80a70 70 0 0 0 0 140h96a70 70 0 0 0 0-140m0 128H80a58 58 0 0 1 0-116h96a58 58 0 0 1 0 116M80 90a38 38 0 1 0 38 38a38 38 0 0 0-38-38m0 64a26 26 0 1 1 26-26a26 26 0 0 1-26 26"
	})
}), {
	__source: "ph:toggle-left-light",
	__viewBox: "0 0 256 256",
	__group: "configs-metadata",
	__consumerName: "toggle-off",
	displayName: "UiToggleOff"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M176 56H80a72 72 0 0 0 0 144h96a72 72 0 0 0 0-144M80 168a40 40 0 1 1 40-40a40 40 0 0 1-40 40"
	})
}), {
	__source: "ph:toggle-left-fill",
	__viewBox: "0 0 256 256",
	__group: "configs-metadata",
	__consumerName: "toggle-off",
	displayName: "UiToggleOffFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiEnum.tsx
var Ot = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("circle", {
			cx: "8",
			cy: "8",
			r: "6.5",
			fill: "#FAF5FF",
			stroke: "#834DF0"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M5.49023 11.5H10.5102V10.585H6.48023V8.435H10.0702V7.515H6.48023V5.415H10.4102V4.5H5.49023V11.5Z",
			fill: "#834DF0"
		})]
	})
}), {
	__source: "jb-expui-nodes:enum",
	__viewBox: "0 0 16 16",
	__group: "uir-ast-code",
	__consumerName: "uir-enum",
	displayName: "UiEnum"
}), kt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("rect", {
			x: "2.5",
			y: "2.5",
			width: "11",
			height: "11",
			rx: "1.5",
			fill: "#EBECF0",
			stroke: "#6C707E"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			fillRule: "evenodd",
			clipRule: "evenodd",
			d: "M7.5 5.5H5V4.5H11V5.5H8.5V11.5H7.5V5.5Z",
			fill: "#6C707E"
		})]
	})
}), {
	__source: "jb-expui-nodes:textArea",
	__viewBox: "0 0 16 16",
	__group: "uir-ast-code",
	__consumerName: "uir-string",
	displayName: "UiString"
}), At = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M4 2.5H12C12.8284 2.5 13.5 3.17157 13.5 4V12C13.5 12.8284 12.8284 13.5 12 13.5H4C3.17157 13.5 2.5 12.8284 2.5 12V4C2.5 3.17157 3.17157 2.5 4 2.5Z",
				stroke: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M7 5.5L6 10.5M10 5.5L9 10.5",
				stroke: "#6C707E",
				strokeLinecap: "round",
				strokeLinejoin: "round"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M11 6.5H5.5M10.5 9.5H5",
				stroke: "#6C707E",
				strokeLinecap: "round"
			})
		]
	})
}), {
	__source: "jb-expui-general:hashtag",
	__viewBox: "0 0 16 16",
	__group: "uir-ast-code",
	__consumerName: "uir-numeric",
	displayName: "UiNumeric"
}), jt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M11.5 3.5H14.5",
				stroke: "#6C707E",
				strokeLinecap: "round",
				strokeLinejoin: "round"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M8.5 8H14.5",
				stroke: "#6C707E",
				strokeLinecap: "round",
				strokeLinejoin: "round"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M5.5 12.5H14.5",
				stroke: "#6C707E",
				strokeLinecap: "round",
				strokeLinejoin: "round"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M8.5 4.5L4 9L1.5 6.5",
				stroke: "#6C707E",
				strokeLinecap: "round"
			})
		]
	})
}), {
	__source: "jb-site:DatabaseIcons/icons/expui/checkConstraint",
	__viewBox: "0 0 16 16",
	__group: "uir-ast-code",
	__consumerName: "uir-boolean",
	displayName: "UiBoolean"
}), Mt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			d: "M5 2C4.44772 2 4 2.44772 4 3V13C4 13.5523 4.44772 14 5 14H7V13H5V3L7 3V2H5Z",
			fill: "#6C707E"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M9 3H11V13H9V14H11C11.5523 14 12 13.5523 12 13V3C12 2.44772 11.5523 2 11 2H9V3Z",
			fill: "#6C707E"
		})]
	})
}), {
	__source: "jb-expui-json:array",
	__viewBox: "0 0 16 16",
	__group: "uir-ast-code",
	__consumerName: "uir-array",
	displayName: "UiArray"
}), Nt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			d: "M3.98755 9.77751C3.99755 9.39251 3.87005 9.08501 3.60505 8.85501C3.34005 8.62501 2.98255 8.51001 2.53255 8.51001H2V7.43001H2.53255C2.98255 7.43001 3.34005 7.31501 3.60505 7.08501C3.87005 6.85501 3.99755 6.54751 3.98755 6.16251L3.95755 4.58C3.94755 4.06 4.05505 3.605 4.28005 3.215C4.51005 2.825 4.83755 2.525 5.26255 2.315C5.68755 2.105 6.18755 2 6.76255 2H7.5V3.005H6.77005C6.25005 3.005 5.83755 3.15 5.53255 3.44C5.23255 3.725 5.08505 4.1125 5.09005 4.6025L5.12005 6.17001C5.13005 6.68001 4.99255 7.09751 4.70755 7.42251C4.42255 7.74251 4.03755 7.92001 3.55255 7.95501C4.03755 8.04501 4.42255 8.25501 4.70755 8.58501C4.99255 8.91001 5.13005 9.30501 5.12005 9.77001L5.09005 11.5475C5.08005 11.9875 5.21255 12.3375 5.48755 12.5975C5.76755 12.8625 6.14505 12.995 6.62005 12.995H7.5V14H6.61255C6.07755 14 5.60755 13.9 5.20255 13.7C4.80255 13.505 4.49255 13.2225 4.27255 12.8525C4.05755 12.4875 3.95255 12.06 3.95755 11.57L3.98755 9.77751Z",
			fill: "#6C707E"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M12.0145 9.77751C12.0045 9.39251 12.132 9.08501 12.397 8.85501C12.662 8.62501 13.0195 8.51001 13.4695 8.51001H14V7.43001H13.4695C13.0195 7.43001 12.662 7.31501 12.397 7.08501C12.132 6.85501 12.0045 6.54751 12.0145 6.16251L12.0445 4.58C12.0545 4.06 11.9445 3.605 11.7145 3.215C11.4895 2.825 11.1645 2.525 10.7395 2.315C10.3145 2.105 9.81455 2 9.23955 2H8.5V3.005H9.23205C9.75205 3.005 10.162 3.15 10.462 3.44C10.767 3.725 10.917 4.1125 10.912 4.6025L10.882 6.17001C10.872 6.68001 11.0095 7.09751 11.2945 7.42251C11.5795 7.74251 11.9645 7.92001 12.4495 7.95501C11.9645 8.04501 11.5795 8.25501 11.2945 8.58501C11.0095 8.91001 10.872 9.30501 10.882 9.77001L10.912 11.5475C10.922 11.9875 10.787 12.3375 10.507 12.5975C10.232 12.8625 9.85705 12.995 9.38205 12.995H8.5V14H9.38955C9.92455 14 10.392 13.9 10.792 13.7C11.197 13.505 11.507 13.2225 11.722 12.8525C11.942 12.4875 12.0495 12.06 12.0445 11.57L12.0145 9.77751Z",
			fill: "#6C707E"
		})]
	})
}), {
	__source: "jb-expui-json:object",
	__viewBox: "0 0 16 16",
	__group: "uir-ast-code",
	__consumerName: "uir-object",
	displayName: "UiObject"
}), Pt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("circle", {
				cx: "8",
				cy: "8",
				r: "6.5",
				stroke: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M9 5C9 4.44772 8.55228 4 8 4C7.44772 4 7 4.44772 7 5V8C7 8.55229 7.44772 9 8 9C8.55229 9 9 8.55228 9 8L9 5Z",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M8 12C8.55228 12 9 11.5523 9 11C9 10.4477 8.55228 10 8 10C7.44772 10 7 10.4477 7 11C7 11.5523 7.44772 12 8 12Z",
				fill: "#6C707E"
			})
		]
	})
}), {
	__source: "jb-expui-status:errorOutline",
	__viewBox: "0 0 16 16",
	__group: "uir-ast-code",
	__consumerName: "uir-null",
	displayName: "UiNull"
}), Ft = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "m212.24 83.76l-56-56A6 6 0 0 0 152 26H56a14 14 0 0 0-14 14v176a14 14 0 0 0 14 14h144a14 14 0 0 0 14-14V88a6 6 0 0 0-1.76-4.24M158 46.48L193.52 82H158ZM200 218H56a2 2 0 0 1-2-2V40a2 2 0 0 1 2-2h90v50a6 6 0 0 0 6 6h50v122a2 2 0 0 1-2 2"
	})
}), {
	__source: "ph:file-light",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "file",
	displayName: "UiFile"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "m213.66 82.34l-56-56A8 8 0 0 0 152 24H56a16 16 0 0 0-16 16v176a16 16 0 0 0 16 16h144a16 16 0 0 0 16-16V88a8 8 0 0 0-2.34-5.66M152 88V44l44 44Z"
	})
}), {
	__source: "ph:file-fill",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "file",
	displayName: "UiFileFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiFileCode.tsx
var It = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M180.24 147.76a6 6 0 0 1 0 8.48l-24 24a6 6 0 0 1-8.48-8.48L167.51 152l-19.75-19.76a6 6 0 1 1 8.48-8.48Zm-72-24a6 6 0 0 0-8.48 0l-24 24a6 6 0 0 0 0 8.48l24 24a6 6 0 1 0 8.48-8.48L88.49 152l19.75-19.76a6 6 0 0 0 0-8.48M214 88v128a14 14 0 0 1-14 14H56a14 14 0 0 1-14-14V40a14 14 0 0 1 14-14h96a6 6 0 0 1 4.25 1.76l56 56A6 6 0 0 1 214 88m-56-6h35.52L158 46.48Zm44 134V94h-50a6 6 0 0 1-6-6V38H56a2 2 0 0 0-2 2v176a2 2 0 0 0 2 2h144a2 2 0 0 0 2-2"
	})
}), {
	__source: "ph:file-code-light",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "file-code",
	displayName: "UiFileCode"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "m213.66 82.34l-56-56A8 8 0 0 0 152 24H56a16 16 0 0 0-16 16v176a16 16 0 0 0 16 16h144a16 16 0 0 0 16-16V88a8 8 0 0 0-2.34-5.66m-104 88a8 8 0 0 1-11.32 11.32l-24-24a8 8 0 0 1 0-11.32l24-24a8 8 0 0 1 11.32 11.32L91.31 152Zm72-12.68l-24 24a8 8 0 0 1-11.32-11.32L164.69 152l-18.35-18.34a8 8 0 0 1 11.32-11.32l24 24a8 8 0 0 1 0 11.32M152 88V44l44 44Z"
	})
}), {
	__source: "ph:file-code-fill",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "file-code",
	displayName: "UiFileCodeFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiFolder.tsx
var Lt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M216 74h-85.51l-27.9-27.9a13.94 13.94 0 0 0-9.9-4.1H40a14 14 0 0 0-14 14v144.62A13.39 13.39 0 0 0 39.38 214h177.51A13.12 13.12 0 0 0 230 200.89V88a14 14 0 0 0-14-14M40 54h52.69a2 2 0 0 1 1.41.59L113.51 74H38V56a2 2 0 0 1 2-2m178 146.89a1.11 1.11 0 0 1-1.11 1.11H39.38a1.4 1.4 0 0 1-1.38-1.38V86h178a2 2 0 0 1 2 2Z"
	})
}), {
	__source: "ph:folder-light",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "folder",
	displayName: "UiFolder"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M216 72h-84.69L104 44.69A15.88 15.88 0 0 0 92.69 40H40a16 16 0 0 0-16 16v144.62A15.41 15.41 0 0 0 39.39 216h177.5A15.13 15.13 0 0 0 232 200.89V88a16 16 0 0 0-16-16M40 56h52.69l16 16H40Z"
	})
}), {
	__source: "ph:folder-fill",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "folder",
	displayName: "UiFolderFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiCopy.tsx
var Rt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M216 34H88a6 6 0 0 0-6 6v42H40a6 6 0 0 0-6 6v128a6 6 0 0 0 6 6h128a6 6 0 0 0 6-6v-42h42a6 6 0 0 0 6-6V40a6 6 0 0 0-6-6m-54 176H46V94h116Zm48-48h-36V88a6 6 0 0 0-6-6H94V46h116Z"
	})
}), {
	__source: "ph:copy-light",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "copy",
	displayName: "UiCopy"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M216 32H88a8 8 0 0 0-8 8v40H40a8 8 0 0 0-8 8v128a8 8 0 0 0 8 8h128a8 8 0 0 0 8-8v-40h40a8 8 0 0 0 8-8V40a8 8 0 0 0-8-8m-8 128h-32V88a8 8 0 0 0-8-8H96V48h112Z"
	})
}), {
	__source: "ph:copy-fill",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "copy",
	displayName: "UiCopyFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiCode2.tsx
var zt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M67.84 92.61L25.37 128l42.47 35.39a6 6 0 1 1-7.68 9.22l-48-40a6 6 0 0 1 0-9.22l48-40a6 6 0 0 1 7.68 9.22m176 30.78l-48-40a6 6 0 1 0-7.68 9.22L230.63 128l-42.47 35.39a6 6 0 1 0 7.68 9.22l48-40a6 6 0 0 0 0-9.22m-81.79-89a6 6 0 0 0-7.69 3.61l-64 176a6 6 0 0 0 3.64 7.64a6.2 6.2 0 0 0 2 .36a6 6 0 0 0 5.64-3.95l64-176a6 6 0 0 0-3.59-7.69Z"
	})
}), {
	__source: "ph:code-light",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "code-2",
	displayName: "UiCode2"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M216 40H40a16 16 0 0 0-16 16v144a16 16 0 0 0 16 16h176a16 16 0 0 0 16-16V56a16 16 0 0 0-16-16M92.8 145.6a8 8 0 1 1-9.6 12.8l-32-24a8 8 0 0 1 0-12.8l32-24a8 8 0 0 1 9.6 12.8L69.33 128Zm58.89-71.4l-32 112a8 8 0 1 1-15.38-4.4l32-112a8 8 0 0 1 15.38 4.4m53.11 60.2l-32 24a8 8 0 0 1-9.6-12.8l23.47-17.6l-23.47-17.6a8 8 0 1 1 9.6-12.8l32 24a8 8 0 0 1 0 12.8"
	})
}), {
	__source: "ph:code-fill",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "code-2",
	displayName: "UiCode2Filled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiDownload.tsx
var Bt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M222 144v64a6 6 0 0 1-6 6H40a6 6 0 0 1-6-6v-64a6 6 0 0 1 12 0v58h164v-58a6 6 0 0 1 12 0m-98.24 4.24a6 6 0 0 0 8.48 0l40-40a6 6 0 0 0-8.48-8.48L134 129.51V32a6 6 0 0 0-12 0v97.51L92.24 99.76a6 6 0 0 0-8.48 8.48Z"
	})
}), {
	__source: "ph:download-simple-light",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "download",
	displayName: "UiDownload"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M224 144v64a8 8 0 0 1-8 8H40a8 8 0 0 1-8-8v-64a8 8 0 0 1 16 0v56h160v-56a8 8 0 0 1 16 0m-101.66 5.66a8 8 0 0 0 11.32 0l40-40A8 8 0 0 0 168 96h-32V32a8 8 0 0 0-16 0v64H88a8 8 0 0 0-5.66 13.66Z"
	})
}), {
	__source: "ph:download-simple-fill",
	__viewBox: "0 0 256 256",
	__group: "files-code",
	__consumerName: "download",
	displayName: "UiDownloadFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiBug.tsx
var Vt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M146 92a10 10 0 1 1 10 10a10 10 0 0 1-10-10m-46-10a10 10 0 1 0 10 10a10 10 0 0 0-10-10m146 46a6 6 0 0 1-6 6h-26v10a85.9 85.9 0 0 1-3.45 24.08l23.85 10.42a6 6 0 0 1-4.8 11l-23.23-10.15a86 86 0 0 1-156.74 0L26.4 189.5a6 6 0 1 1-4.8-11l23.85-10.42A85.9 85.9 0 0 1 42 144v-10H16a6 6 0 0 1 0-12h26v-10a85.9 85.9 0 0 1 3.45-24.08L21.6 77.5a6 6 0 0 1 4.8-11l23.23 10.15a86 86 0 0 1 156.74 0L229.6 66.5a6 6 0 1 1 4.8 11l-23.85 10.42A85.9 85.9 0 0 1 214 112v10h26a6 6 0 0 1 6 6m-192-6h148v-10a74 74 0 0 0-148 0Zm68 95.74V134H54v10a74.09 74.09 0 0 0 68 73.74M202 134h-68v83.74A74.09 74.09 0 0 0 202 144Z"
	})
}), {
	__source: "ph:bug-light",
	__viewBox: "0 0 256 256",
	__group: "dev-tools",
	__consumerName: "bug",
	displayName: "UiBug"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M168 92a12 12 0 1 1-12-12a12 12 0 0 1 12 12m-68-12a12 12 0 1 0 12 12a12 12 0 0 0-12-12m116 64a87.8 87.8 0 0 1-3 23l22.24 9.72A8 8 0 0 1 232 192a7.9 7.9 0 0 1-3.2-.67L207.38 182a88 88 0 0 1-158.76 0l-21.42 9.33a7.9 7.9 0 0 1-3.2.67a8 8 0 0 1-3.2-15.33L43 167a87.8 87.8 0 0 1-3-23v-8H16a8 8 0 0 1 0-16h24v-8a87.8 87.8 0 0 1 3-23l-22.2-9.67a8 8 0 1 1 6.4-14.66L48.62 74a88 88 0 0 1 158.76 0l21.42-9.36a8 8 0 0 1 6.4 14.66L213 89.05a87.8 87.8 0 0 1 3 23v8h24a8 8 0 0 1 0 16h-24Zm-80 0a8 8 0 0 0-16 0v64a8 8 0 0 0 16 0Zm64-32a72 72 0 0 0-144 0v8h144Z"
	})
}), {
	__source: "ph:bug-fill",
	__viewBox: "0 0 256 256",
	__group: "dev-tools",
	__consumerName: "bug",
	displayName: "UiBugFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiTrash.tsx
var Ht = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M216 50h-42V40a22 22 0 0 0-22-22h-48a22 22 0 0 0-22 22v10H40a6 6 0 0 0 0 12h10v146a14 14 0 0 0 14 14h128a14 14 0 0 0 14-14V62h10a6 6 0 0 0 0-12M94 40a10 10 0 0 1 10-10h48a10 10 0 0 1 10 10v10H94Zm100 168a2 2 0 0 1-2 2H64a2 2 0 0 1-2-2V62h132Zm-84-104v64a6 6 0 0 1-12 0v-64a6 6 0 0 1 12 0m48 0v64a6 6 0 0 1-12 0v-64a6 6 0 0 1 12 0"
	})
}), {
	__source: "ph:trash-light",
	__viewBox: "0 0 256 256",
	__group: "actions-tools",
	__consumerName: "trash",
	displayName: "UiTrash"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M216 48h-40v-8a24 24 0 0 0-24-24h-48a24 24 0 0 0-24 24v8H40a8 8 0 0 0 0 16h8v144a16 16 0 0 0 16 16h128a16 16 0 0 0 16-16V64h8a8 8 0 0 0 0-16M112 168a8 8 0 0 1-16 0v-64a8 8 0 0 1 16 0Zm48 0a8 8 0 0 1-16 0v-64a8 8 0 0 1 16 0Zm0-120H96v-8a8 8 0 0 1 8-8h48a8 8 0 0 1 8 8Z"
	})
}), {
	__source: "ph:trash-fill",
	__viewBox: "0 0 256 256",
	__group: "actions-tools",
	__consumerName: "trash",
	displayName: "UiTrashFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiAsterisk.tsx
var Ut = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M213.14 179.09a6 6 0 0 1-8.23 2.06L134 138.6V216a6 6 0 0 1-12 0v-77.4l-70.91 42.55A6.07 6.07 0 0 1 48 182a6 6 0 0 1-3.1-11.15L116.34 128L44.91 85.15a6 6 0 0 1 6.18-10.3L122 117.4V40a6 6 0 0 1 12 0v77.4l70.91-42.55a6 6 0 0 1 6.18 10.3L139.66 128l71.43 42.85a6 6 0 0 1 2.05 8.24"
	})
}), {
	__source: "ph:asterisk-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "asterisk",
	displayName: "UiAsterisk"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M128 24a104 104 0 1 0 104 104A104.11 104.11 0 0 0 128 24m59.43 129.07a8 8 0 0 1-4 14.93a7.9 7.9 0 0 1-4-1.07L136 141.86V192a8 8 0 0 1-16 0v-50.14l-43.43 25.07A8 8 0 0 1 65.65 164a8 8 0 0 1 2.92-10.93L112 128l-43.43-25.07a8 8 0 0 1 8-13.86L120 114.14V64a8 8 0 0 1 16 0v50.14l43.43-25.07a8 8 0 0 1 8 13.86L144 128Z"
	})
}), {
	__source: "ph:asterisk-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "asterisk",
	displayName: "UiAsteriskFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiProhibit.tsx
var Wt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M128 26a102 102 0 1 0 102 102A102.12 102.12 0 0 0 128 26m90 102a89.6 89.6 0 0 1-22.29 59.22L68.78 60.29A89.95 89.95 0 0 1 218 128m-180 0a89.6 89.6 0 0 1 22.29-59.22l126.93 126.93A89.95 89.95 0 0 1 38 128"
	})
}), {
	__source: "ph:prohibit-light",
	__viewBox: "0 0 256 256",
	__group: "runtime-process",
	__consumerName: "prohibit",
	displayName: "UiProhibit"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M200 128a71.7 71.7 0 0 1-15.78 44.91L83.09 71.78A71.95 71.95 0 0 1 200 128m-144 0a71.95 71.95 0 0 0 116.91 56.22L71.78 83.09A71.7 71.7 0 0 0 56 128m180 0A108 108 0 1 1 128 20a108.12 108.12 0 0 1 108 108m-20 0a88 88 0 1 0-88 88a88.1 88.1 0 0 0 88-88"
	})
}), {
	__source: "ph:prohibit-fill",
	__viewBox: "0 0 256 256",
	__group: "runtime-process",
	__consumerName: "prohibit",
	displayName: "UiProhibitFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiPalette.tsx
var Gt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M199.37 55.31A101.32 101.32 0 0 0 128 26h-1A102 102 0 0 0 26 128c0 42.09 26.07 77.44 68 92.26a30.2 30.2 0 0 0 10.11 1.74A30.06 30.06 0 0 0 134 192a18 18 0 0 1 18-18h46.21a29.82 29.82 0 0 0 29.25-23.31a102.7 102.7 0 0 0 2.54-23.58a101.25 101.25 0 0 0-30.63-71.8M215.76 148a17.89 17.89 0 0 1-17.55 14H152a30 30 0 0 0-30 30a18 18 0 0 1-24 17c-37-13.14-60-44.15-60-81a90 90 0 0 1 89.07-90h.93a90.34 90.34 0 0 1 90 89.22a90.5 90.5 0 0 1-2.24 20.78M138 76a10 10 0 1 1-10-10a10 10 0 0 1 10 10m-44 24a10 10 0 1 1-10-10a10 10 0 0 1 10 10m0 56a10 10 0 1 1-10-10a10 10 0 0 1 10 10m88-56a10 10 0 1 1-10-10a10 10 0 0 1 10 10"
	})
}), {
	__source: "ph:palette-light",
	__viewBox: "0 0 256 256",
	__group: "media",
	__consumerName: "palette",
	displayName: "UiPalette"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M200.77 53.89A103.27 103.27 0 0 0 128 24h-1.07A104 104 0 0 0 24 128c0 43 26.58 79.06 69.36 94.17A32 32 0 0 0 136 192a16 16 0 0 1 16-16h46.21a31.81 31.81 0 0 0 31.2-24.88a104.4 104.4 0 0 0 2.59-24a103.28 103.28 0 0 0-31.23-73.23M84 168a12 12 0 1 1 12-12a12 12 0 0 1-12 12m0-56a12 12 0 1 1 12-12a12 12 0 0 1-12 12m44-24a12 12 0 1 1 12-12a12 12 0 0 1-12 12m44 24a12 12 0 1 1 12-12a12 12 0 0 1-12 12"
	})
}), {
	__source: "ph:palette-fill",
	__viewBox: "0 0 256 256",
	__group: "media",
	__consumerName: "palette",
	displayName: "UiPaletteFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiArrowsCounterClockwise.tsx
var Kt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M88 102H40a6 6 0 0 1-6-6V48a6 6 0 0 1 12 0v33.52l18-18a93.45 93.45 0 0 1 66-27.53h.52a93.4 93.4 0 0 1 65.67 26.71a6 6 0 0 1-8.38 8.58A82 82 0 0 0 72.53 72L54.48 90H88a6 6 0 0 1 0 12m128 52h-48a6 6 0 0 0 0 12h33.52l-18.05 18a81.5 81.5 0 0 1-57.53 24h-.46a81.5 81.5 0 0 1-57.29-23.34a6 6 0 0 0-8.38 8.58A93.4 93.4 0 0 0 125.48 220h.52a93.45 93.45 0 0 0 66-27.53l18-18.05V208a6 6 0 0 0 12 0v-48a6 6 0 0 0-6-6"
	})
}), {
	__source: "ph:arrows-counter-clockwise-light",
	__viewBox: "0 0 256 256",
	__group: "actions-tools",
	__consumerName: "arrows-counter-clockwise",
	displayName: "UiArrowsCounterClockwise"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M88 104H40a8 8 0 0 1-8-8V48a8 8 0 0 1 13.66-5.66L64 60.7a95.42 95.42 0 0 1 66-26.76h.53a95.36 95.36 0 0 1 67.07 27.33a8 8 0 0 1-11.18 11.44a79.52 79.52 0 0 0-55.89-22.77h-.45A79.48 79.48 0 0 0 75.35 72l18.31 18.34A8 8 0 0 1 88 104m128 48h-48a8 8 0 0 0-5.66 13.66L180.65 184a79.48 79.48 0 0 1-54.72 22.09h-.45a79.52 79.52 0 0 1-55.89-22.77a8 8 0 1 0-11.18 11.44a95.36 95.36 0 0 0 67.07 27.33h.52a95.42 95.42 0 0 0 66-26.76l18.36 18.36A8 8 0 0 0 224 208v-48a8 8 0 0 0-8-8"
	})
}), {
	__source: "ph:arrows-counter-clockwise-fill",
	__viewBox: "0 0 256 256",
	__group: "actions-tools",
	__consumerName: "arrows-counter-clockwise",
	displayName: "UiArrowsCounterClockwiseFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiListNumbers.tsx
var qt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M222 128a6 6 0 0 1-6 6H104a6 6 0 0 1 0-12h112a6 6 0 0 1 6 6M104 70h112a6 6 0 0 0 0-12H104a6 6 0 0 0 0 12m112 116H104a6 6 0 0 0 0 12h112a6 6 0 0 0 0-12M42.68 53.37L50 49.71V104a6 6 0 0 0 12 0V40a6 6 0 0 0-8.68-5.37l-16 8a6 6 0 0 0 5.36 10.74M72 202H52l21.48-28.74A21.5 21.5 0 0 0 77.79 157A21.75 21.75 0 0 0 69 142.38a22.86 22.86 0 0 0-31.35 4.31a22.2 22.2 0 0 0-3.28 5.92a6 6 0 0 0 11.28 4.11a9.9 9.9 0 0 1 1.48-2.67a10.78 10.78 0 0 1 14.78-2a9.9 9.9 0 0 1 4 6.61a9.64 9.64 0 0 1-2 7.28l-.06.09l-28.65 38.38A6 6 0 0 0 40 214h32a6 6 0 0 0 0-12"
	})
}), {
	__source: "ph:list-numbers-light",
	__viewBox: "0 0 256 256",
	__group: "playbooks-workflows",
	__consumerName: "list-numbers",
	displayName: "UiListNumbers"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("path", {
		fill: "currentColor",
		d: "M208 32H48a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16M56.84 75.58a8 8 0 0 1 3.58-10.74l16-8A8 8 0 0 1 88 64v48a8 8 0 0 1-16 0V76.94l-4.42 2.22a8 8 0 0 1-10.74-3.58M92 180a8 8 0 0 1 0 16H68a8 8 0 0 1-6.4-12.8l21.67-28.89A3.92 3.92 0 0 0 84 152a4 4 0 0 0-7.77-1.33a8 8 0 0 1-15.09-5.34a20 20 0 1 1 35 18.53L84 180Zm100 4h-72a8 8 0 0 1 0-16h72a8 8 0 0 1 0 16m0-48h-72a8 8 0 0 1 0-16h72a8 8 0 0 1 0 16m0-48h-72a8 8 0 0 1 0-16h72a8 8 0 0 1 0 16"
	})
}), {
	__source: "ph:list-numbers-fill",
	__viewBox: "0 0 256 256",
	__group: "playbooks-workflows",
	__consumerName: "list-numbers",
	displayName: "UiListNumbersFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiConstant1.tsx
var Jt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M0.586274 6.58573L6.58597 0.586031C7.36702 -0.195018 8.63335 -0.195019 9.4144 0.58603L15.4141 6.58573C16.1951 7.36678 16.1951 8.63311 15.4141 9.41416L9.4144 15.4139C8.63335 16.1949 7.36702 16.1949 6.58597 15.4139L0.586275 9.41416C-0.194773 8.63311 -0.194774 7.36678 0.586274 6.58573Z",
				fill: "#FFF4EB"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M7.29308 1.29314L1.29338 7.29284C0.902857 7.68336 0.902858 8.31653 1.29338 8.70705L7.29308 14.7067C7.68361 15.0973 8.31677 15.0973 8.70729 14.7067L14.707 8.70705C15.0975 8.31652 15.0975 7.68336 14.707 7.29284L8.70729 1.29314C8.31677 0.902613 7.6836 0.902613 7.29308 1.29314ZM6.58597 0.586031L0.586274 6.58573C-0.194774 7.36678 -0.194773 8.63311 0.586275 9.41416L6.58597 15.4139C7.36702 16.1949 8.63335 16.1949 9.4144 15.4139L15.4141 9.41416C16.1951 8.63311 16.1951 7.36678 15.4141 6.58573L9.4144 0.58603C8.63335 -0.195019 7.36702 -0.195018 6.58597 0.586031Z",
				fill: "#E66D17"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M8.1329 11.5C9.62185 11.5 10.8981 10.6008 11.2123 9.32459H10.1971C9.8829 10.0546 9.10459 10.5767 8.1329 10.5767C6.79381 10.5767 5.82696 9.50345 5.82696 8C5.82696 6.49655 6.79381 5.42334 8.1329 5.42334C9.10459 5.42334 9.8829 5.94544 10.1971 6.67541H11.2123C10.8981 5.39917 9.62185 4.5 8.1329 4.5C6.2137 4.5 4.7876 5.99378 4.7876 8C4.7876 10.0062 6.2137 11.5 8.1329 11.5Z",
				fill: "#E66D17"
			})
		]
	})
}), {
	__source: "jb-download:constant",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-constant",
	displayName: "UiConstant1"
}), Yt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M0.586274 6.58573L6.58597 0.586031C7.36702 -0.195018 8.63335 -0.195019 9.4144 0.58603L15.4141 6.58573C16.1951 7.36678 16.1951 8.63311 15.4141 9.41416L9.4144 15.4139C8.63335 16.1949 7.36702 16.1949 6.58597 15.4139L0.586275 9.41416C-0.194773 8.63311 -0.194774 7.36678 0.586274 6.58573Z",
				fill: "#45322B"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M7.29308 1.29314L1.29338 7.29284C0.902857 7.68336 0.902858 8.31653 1.29338 8.70705L7.29308 14.7067C7.68361 15.0973 8.31677 15.0973 8.70729 14.7067L14.707 8.70705C15.0975 8.31652 15.0975 7.68336 14.707 7.29284L8.70729 1.29314C8.31677 0.902613 7.6836 0.902613 7.29308 1.29314ZM6.58597 0.586031L0.586274 6.58573C-0.194774 7.36678 -0.194773 8.63311 0.586275 9.41416L6.58597 15.4139C7.36702 16.1949 8.63335 16.1949 9.4144 15.4139L15.4141 9.41416C16.1951 8.63311 16.1951 7.36678 15.4141 6.58573L9.4144 0.58603C8.63335 -0.195019 7.36702 -0.195018 6.58597 0.586031Z",
				fill: "#C77D55"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M8.1329 11.5C9.62185 11.5 10.8981 10.6008 11.2123 9.32459H10.1971C9.8829 10.0546 9.10459 10.5767 8.1329 10.5767C6.79381 10.5767 5.82696 9.50345 5.82696 8C5.82696 6.49655 6.79381 5.42334 8.1329 5.42334C9.10459 5.42334 9.8829 5.94544 10.1971 6.67541H11.2123C10.8981 5.39917 9.62185 4.5 8.1329 4.5C6.2137 4.5 4.7876 5.99378 4.7876 8C4.7876 10.0062 6.2137 11.5 8.1329 11.5Z",
				fill: "#C77D55"
			})
		]
	})
}), {
	__source: "jb-download:constant-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-constant",
	displayName: "UiConstant1Dark"
}), Xt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 14 14",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M0 3C0 1.89543 0.895431 1 2 1H12C13.1046 1 14 1.89543 14 3V10C14 8.98555 13.6224 8.05931 13 7.35418V5H10V6C9.6547 6 9.31962 6.04375 9 6.12602V5H5V12H6.53513C6.75287 12.3764 7.03018 12.714 7.35418 13H2C0.89543 13 0 12.1046 0 11V3ZM5 2H9V4H5V2ZM4 2H2C1.44772 2 1 2.44772 1 3V4H4V2ZM1 5V11C1 11.5523 1.44772 12 2 12H4V5H1ZM13 4V3C13 2.44772 12.5523 2 12 2H10V4H13Z",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("circle", {
				cx: "10",
				cy: "10",
				r: "2",
				fill: "#EDF3FF"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M11.7383 12.4454C11.2479 12.7946 10.6479 13 10 13C8.34315 13 7 11.6569 7 10C7 8.34315 8.34315 7 10 7C11.6569 7 13 8.34315 13 10C13 10.6479 12.7946 11.2479 12.4454 11.7383L13.8536 13.1464C14.0488 13.3417 14.0488 13.6583 13.8536 13.8536C13.6583 14.0488 13.3417 14.0488 13.1464 13.8536L11.7383 12.4454ZM12 10C12 11.1046 11.1046 12 10 12C8.89543 12 8 11.1046 8 10C8 8.89543 8.89543 8 10 8C11.1046 8 12 8.89543 12 10Z",
				fill: "#3574F0"
			})
		]
	})
}), {
	__source: "jb-download:data-schema",
	__viewBox: "0 0 14 14",
	__group: "programming",
	__consumerName: "programming-data-schema",
	displayName: "UiDataSchema"
}), Zt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 14 14",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M0 3C0 1.89543 0.895431 1 2 1H12C13.1046 1 14 1.89543 14 3V10C14 8.98555 13.6224 8.05931 13 7.35418V5H10V6C9.6547 6 9.31962 6.04375 9 6.12602V5H5V12H6.53513C6.75287 12.3764 7.03018 12.714 7.35418 13H2C0.89543 13 0 12.1046 0 11V3ZM5 2H9V4H5V2ZM4 2H2C1.44772 2 1 2.44772 1 3V4H4V2ZM1 5V11C1 11.5523 1.44772 12 2 12H4V5H1ZM13 4V3C13 2.44772 12.5523 2 12 2H10V4H13Z",
				fill: "#CED0D6"
			}),
			/* @__PURE__ */ (0, J.jsx)("circle", {
				cx: "10",
				cy: "10",
				r: "2",
				fill: "#25324D"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M11.7383 12.4454C11.2479 12.7946 10.6479 13 10 13C8.34315 13 7 11.6569 7 10C7 8.34315 8.34315 7 10 7C11.6569 7 13 8.34315 13 10C13 10.6479 12.7946 11.2479 12.4454 11.7383L13.8536 13.1464C14.0488 13.3417 14.0488 13.6583 13.8536 13.8536C13.6583 14.0488 13.3417 14.0488 13.1464 13.8536L11.7383 12.4454ZM12 10C12 11.1046 11.1046 12 10 12C8.89543 12 8 11.1046 8 10C8 8.89543 8.89543 8 10 8C11.1046 8 12 8.89543 12 10Z",
				fill: "#548AF7"
			})
		]
	})
}), {
	__source: "jb-download:data-schema-dark",
	__viewBox: "0 0 14 14",
	__group: "programming",
	__consumerName: "programming-data-schema",
	displayName: "UiDataSchemaDark"
}), Qt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			fillRule: "evenodd",
			clipRule: "evenodd",
			d: "M11 10.5C11 10.2239 11.2239 10 11.5 10H15.5C15.7761 10 16 10.2239 16 10.5V14.5C16 14.7761 15.7761 15 15.5 15C15.2239 15 15 14.7761 15 14.5V11.7071L10.8536 15.8536C10.6583 16.0488 10.3417 16.0488 10.1464 15.8536C9.95118 15.6583 9.95118 15.3417 10.1464 15.1464L14.2929 11H11.5C11.2239 11 11 10.7761 11 10.5Z",
			fill: "#3574F0"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			fillRule: "evenodd",
			clipRule: "evenodd",
			d: "M2 3.39246H2.01664C2.07865 3.00405 2.30765 2.67483 2.59694 2.41334C2.9371 2.10586 3.40129 1.85189 3.93236 1.64885C4.99792 1.24147 6.43619 1 8 1C9.56381 1 11.0021 1.24147 12.0676 1.64885C12.5987 1.85189 13.0629 2.10586 13.4031 2.41334C13.6924 2.67483 13.9213 3.00405 13.9834 3.39246H14V9H13V8.04112C12.7276 8.21609 12.4117 8.369 12.0676 8.50053C11.0021 8.90791 9.56381 9.14938 8 9.14938C6.43619 9.14938 4.99792 8.90791 3.93236 8.50053C3.58833 8.369 3.27236 8.21609 3 8.04112V9.49048C3 9.59029 3.05102 9.74236 3.2675 9.93804C3.48396 10.1337 3.82523 10.3328 4.28947 10.5103C5.21452 10.864 6.52624 11.0932 8 11.0932C8.73793 11.0932 9.43524 11.0357 10.0637 10.934C10.1783 11.3137 10.4393 11.6298 10.7811 11.8168C9.94846 11.9945 8.99989 12.0932 8 12.0932C6.43619 12.0932 4.99792 11.8518 3.93236 11.4444C3.58833 11.3128 3.27236 11.1599 3 10.985V12.4343C3 12.5341 3.05102 12.6862 3.2675 12.8819C3.48396 13.0775 3.82523 13.2767 4.28947 13.4542C5.21452 13.8078 6.52624 14.0371 8 14.0371C8.70883 14.0371 9.38017 13.984 9.98904 13.8896L9.43934 14.4393C9.27627 14.6024 9.1586 14.7937 9.08632 14.9975C8.73343 15.0235 8.37024 15.0371 8 15.0371C6.43619 15.0371 4.99792 14.7956 3.93236 14.3882C3.40129 14.1852 2.9371 13.9312 2.59694 13.6237C2.2768 13.3344 2.03051 12.9621 2.00263 12.5184H2V3.39246ZM3 3.60275C3 3.50293 3.05102 3.35087 3.2675 3.15519C3.48396 2.95953 3.82523 2.7604 4.28947 2.58291C5.21452 2.22925 6.52624 2 8 2C9.47376 2 10.7855 2.22925 11.7105 2.58291C12.1748 2.7604 12.516 2.95953 12.7325 3.15519C12.949 3.35087 13 3.50293 13 3.60275C13 3.70256 12.949 3.85463 12.7325 4.05031C12.516 4.24597 12.1748 4.4451 11.7105 4.62258C10.7855 4.97625 9.47376 5.2055 8 5.2055C6.52624 5.2055 5.21452 4.97625 4.28947 4.62258C3.82523 4.4451 3.48396 4.24597 3.2675 4.05031C3.05102 3.85463 3 3.70256 3 3.60275ZM3 6.54663V5.09723C3.27236 5.27221 3.58833 5.42512 3.93236 5.55665C4.99792 5.96403 6.43619 6.2055 8 6.2055C9.56381 6.2055 11.0021 5.96403 12.0676 5.55665C12.4117 5.42512 12.7276 5.27221 13 5.09723V6.54663C13 6.64645 12.949 6.79851 12.7325 6.99419C12.516 7.18985 12.1748 7.38898 11.7105 7.56647C10.7855 7.92013 9.47376 8.14938 8 8.14938C6.52624 8.14938 5.21452 7.92013 4.28947 7.56647C3.82523 7.38898 3.48396 7.18985 3.2675 6.99419C3.05102 6.79851 3 6.64645 3 6.54663Z",
			fill: "#6C707E"
		})]
	})
}), {
	__source: "jb-download:database-link",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-database-link",
	displayName: "UiDatabaseLink"
}), $t = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			fillRule: "evenodd",
			clipRule: "evenodd",
			d: "M11 10.5C11 10.2239 11.2239 10 11.5 10H15.5C15.7761 10 16 10.2239 16 10.5V14.5C16 14.7761 15.7761 15 15.5 15C15.2239 15 15 14.7761 15 14.5V11.7071L10.8536 15.8536C10.6583 16.0488 10.3417 16.0488 10.1464 15.8536C9.95118 15.6583 9.95118 15.3417 10.1464 15.1464L14.2929 11H11.5C11.2239 11 11 10.7761 11 10.5Z",
			fill: "#548AF7"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			fillRule: "evenodd",
			clipRule: "evenodd",
			d: "M2 3.39246H2.01664C2.07865 3.00405 2.30765 2.67483 2.59694 2.41334C2.9371 2.10586 3.40129 1.85189 3.93236 1.64885C4.99792 1.24147 6.43619 1 8 1C9.56381 1 11.0021 1.24147 12.0676 1.64885C12.5987 1.85189 13.0629 2.10586 13.4031 2.41334C13.6924 2.67483 13.9213 3.00405 13.9834 3.39246H14V9H13V8.04112C12.7276 8.21609 12.4117 8.369 12.0676 8.50053C11.0021 8.90791 9.56381 9.14938 8 9.14938C6.43619 9.14938 4.99792 8.90791 3.93236 8.50053C3.58833 8.369 3.27236 8.21609 3 8.04112V9.49048C3 9.59029 3.05102 9.74236 3.2675 9.93804C3.48396 10.1337 3.82523 10.3328 4.28947 10.5103C5.21452 10.864 6.52624 11.0932 8 11.0932C8.73793 11.0932 9.43524 11.0357 10.0637 10.934C10.1783 11.3137 10.4393 11.6298 10.7811 11.8168C9.94846 11.9945 8.99989 12.0932 8 12.0932C6.43619 12.0932 4.99792 11.8518 3.93236 11.4444C3.58833 11.3128 3.27236 11.1599 3 10.985V12.4343C3 12.5341 3.05102 12.6862 3.2675 12.8819C3.48396 13.0775 3.82523 13.2767 4.28947 13.4542C5.21452 13.8078 6.52624 14.0371 8 14.0371C8.70883 14.0371 9.38017 13.984 9.98904 13.8896L9.43934 14.4393C9.27627 14.6024 9.1586 14.7937 9.08632 14.9975C8.73343 15.0235 8.37024 15.0371 8 15.0371C6.43619 15.0371 4.99792 14.7956 3.93236 14.3882C3.40129 14.1852 2.9371 13.9312 2.59694 13.6237C2.2768 13.3344 2.03051 12.9621 2.00263 12.5184H2V3.39246ZM3 3.60275C3 3.50293 3.05102 3.35087 3.2675 3.15519C3.48396 2.95953 3.82523 2.7604 4.28947 2.58291C5.21452 2.22925 6.52624 2 8 2C9.47376 2 10.7855 2.22925 11.7105 2.58291C12.1748 2.7604 12.516 2.95953 12.7325 3.15519C12.949 3.35087 13 3.50293 13 3.60275C13 3.70256 12.949 3.85463 12.7325 4.05031C12.516 4.24597 12.1748 4.4451 11.7105 4.62258C10.7855 4.97625 9.47376 5.2055 8 5.2055C6.52624 5.2055 5.21452 4.97625 4.28947 4.62258C3.82523 4.4451 3.48396 4.24597 3.2675 4.05031C3.05102 3.85463 3 3.70256 3 3.60275ZM3 6.54663V5.09723C3.27236 5.27221 3.58833 5.42512 3.93236 5.55665C4.99792 5.96403 6.43619 6.2055 8 6.2055C9.56381 6.2055 11.0021 5.96403 12.0676 5.55665C12.4117 5.42512 12.7276 5.27221 13 5.09723V6.54663C13 6.64645 12.949 6.79851 12.7325 6.99419C12.516 7.18985 12.1748 7.38898 11.7105 7.56647C10.7855 7.92013 9.47376 8.14938 8 8.14938C6.52624 8.14938 5.21452 7.92013 4.28947 7.56647C3.82523 7.38898 3.48396 7.18985 3.2675 6.99419C3.05102 6.79851 3 6.64645 3 6.54663Z",
			fill: "#CED0D6"
		})]
	})
}), {
	__source: "jb-download:database-link-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-database-link",
	displayName: "UiDatabaseLinkDark"
}), en = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "1",
				y: "3",
				width: "13",
				height: "10",
				rx: "1",
				fill: "#FAF5FF"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M4 5H3V6H4V5Z",
				fill: "#834DF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M5 7H4V8H5V7Z",
				fill: "#834DF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M7 5H8V6H7V5Z",
				fill: "#834DF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M9 7H8V8H9V7Z",
				fill: "#834DF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M5 5H6V6H5V5Z",
				fill: "#834DF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M7 7H6V8H7V7Z",
				fill: "#834DF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M9 5H10V6H9V5Z",
				fill: "#834DF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M12 5H11V6H12V5Z",
				fill: "#834DF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M10 7H11V8H10V7Z",
				fill: "#834DF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M10 10H5V11H10V10Z",
				fill: "#834DF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M0 4C0 2.89543 0.895431 2 2 2H13C14.1046 2 15 2.89543 15 4V12C15 13.1046 14.1046 14 13 14H2C0.89543 14 0 13.1046 0 12V4ZM2 3H13C13.5523 3 14 3.44772 14 4V12C14 12.5523 13.5523 13 13 13H2C1.44772 13 1 12.5523 1 12V4C1 3.44772 1.44772 3 2 3Z",
				fill: "#834DF0"
			})
		]
	})
}), {
	__source: "jb-download-unverified:edit-mode",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-edit-mode",
	displayName: "UiEditMode"
}), tn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "1",
				y: "3",
				width: "13",
				height: "10",
				rx: "1",
				fill: "#2F2936"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M4 5H3V6H4V5Z",
				fill: "#A571E6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M5 7H4V8H5V7Z",
				fill: "#A571E6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M7 5H8V6H7V5Z",
				fill: "#A571E6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M9 7H8V8H9V7Z",
				fill: "#A571E6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M5 5H6V6H5V5Z",
				fill: "#A571E6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M7 7H6V8H7V7Z",
				fill: "#A571E6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M9 5H10V6H9V5Z",
				fill: "#A571E6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M12 5H11V6H12V5Z",
				fill: "#A571E6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M10 7H11V8H10V7Z",
				fill: "#A571E6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M10 10H5V11H10V10Z",
				fill: "#A571E6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M0 4C0 2.89543 0.895431 2 2 2H13C14.1046 2 15 2.89543 15 4V12C15 13.1046 14.1046 14 13 14H2C0.89543 14 0 13.1046 0 12V4ZM2 3H13C13.5523 3 14 3.44772 14 4V12C14 12.5523 13.5523 13 13 13H2C1.44772 13 1 12.5523 1 12V4C1 3.44772 1.44772 3 2 3Z",
				fill: "#A571E6"
			})
		]
	})
}), {
	__source: "jb-download-unverified:edit-mode-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-edit-mode",
	displayName: "UiEditModeDark"
}), nn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "2.5",
				y: "2.5",
				width: "11",
				height: "11",
				rx: "1.5",
				stroke: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "4.5",
				y: "4.5",
				width: "7",
				height: "2",
				rx: "0.5",
				stroke: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "4.29999",
				y: "8",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "7.29999",
				y: "8",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "10.3",
				y: "8",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "4.29999",
				y: "10.6001",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "7.29999",
				y: "10.6001",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "10.3",
				y: "10.6001",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#6C707E"
			})
		]
	})
}), {
	__source: "jb-download:evaluate-expression",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-evaluate-expression",
	displayName: "UiEvaluateExpression"
}), rn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "2.5",
				y: "2.5",
				width: "11",
				height: "11",
				rx: "1.5",
				stroke: "#CED0D6"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "4.5",
				y: "4.5",
				width: "7",
				height: "2",
				rx: "0.5",
				stroke: "#CED0D6"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "4.29999",
				y: "8",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#CED0D6"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "7.29999",
				y: "8",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#CED0D6"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "10.3",
				y: "8",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#CED0D6"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "4.29999",
				y: "10.6001",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#CED0D6"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "7.29999",
				y: "10.6001",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#CED0D6"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "10.3",
				y: "10.6001",
				width: "1.4",
				height: "1.4",
				rx: "0.7",
				fill: "#CED0D6"
			})
		]
	})
}), {
	__source: "jb-download:evaluate-expression-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-evaluate-expression",
	displayName: "UiEvaluateExpressionDark"
}), an = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("g", {
		fill: "none",
		children: /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M9.5 3.5H9C9 3.77614 9.22386 4 9.5 4V3.5ZM12.5 3.5H13C13 3.22386 12.7761 3 12.5 3V3.5ZM12.5 6.5H12C12 6.77614 12.2239 7 12.5 7V6.5ZM12.5 10.5V10C12.2239 10 12 10.2239 12 10.5H12.5ZM12.5 13.5V14C12.7761 14 13 13.7761 13 13.5H12.5ZM2.5 13.5H2C2 13.7761 2.22386 14 2.5 14V13.5ZM5.5 3.5V4C5.77614 4 6 3.77614 6 3.5H5.5ZM2.5 3.5V3C2.22386 3 2 3.22386 2 3.5H2.5ZM2.5 6.5H2C2 6.77614 2.22386 7 2.5 7V6.5ZM2.5 10.5V10C2.22386 10 2 10.2239 2 10.5H2.5ZM7.5 0C6.11929 0 5 1.11929 5 2.5H6C6 1.67157 6.67157 1 7.5 1V0ZM10 2.5C10 1.11929 8.88071 0 7.5 0V1C8.32843 1 9 1.67157 9 2.5H10ZM10 3.5V2.5H9V3.5H10ZM12.5 3H9.5V4H12.5V3ZM13 6.5V3.5H12V6.5H13ZM12.5 7H13.5V6H12.5V7ZM13.5 7C14.3284 7 15 7.67157 15 8.5H16C16 7.11929 14.8807 6 13.5 6V7ZM15 8.5C15 9.32843 14.3284 10 13.5 10V11C14.8807 11 16 9.88071 16 8.5H15ZM13.5 10H12.5V11H13.5V10ZM13 13.5V10.5H12V13.5H13ZM2.5 14H12.5V13H2.5V14ZM5.5 3H2.5V4H5.5V3ZM5 2.5V3.5H6V2.5H5ZM2 3.5V6.5H3V3.5H2ZM2 10.5V13.5H3V10.5H2ZM2.5 7H3.5V6H2.5V7ZM3.5 10H2.5V11H3.5V10ZM5 8.5C5 9.32843 4.32843 10 3.5 10V11C4.88071 11 6 9.88071 6 8.5H5ZM3.5 7C4.32843 7 5 7.67157 5 8.5H6C6 7.11929 4.88071 6 3.5 6V7Z",
			fill: "#3574F0"
		})
	})
}), {
	__source: "jb-download:extension",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-extension",
	displayName: "UiExtension"
}), on = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsx)("g", {
		fill: "none",
		children: /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M9.5 3.5H9C9 3.77614 9.22386 4 9.5 4V3.5ZM12.5 3.5H13C13 3.22386 12.7761 3 12.5 3V3.5ZM12.5 6.5H12C12 6.77614 12.2239 7 12.5 7V6.5ZM12.5 10.5V10C12.2239 10 12 10.2239 12 10.5H12.5ZM12.5 13.5V14C12.7761 14 13 13.7761 13 13.5H12.5ZM2.5 13.5H2C2 13.7761 2.22386 14 2.5 14V13.5ZM5.5 3.5V4C5.77614 4 6 3.77614 6 3.5H5.5ZM2.5 3.5V3C2.22386 3 2 3.22386 2 3.5H2.5ZM2.5 6.5H2C2 6.77614 2.22386 7 2.5 7V6.5ZM2.5 10.5V10C2.22386 10 2 10.2239 2 10.5H2.5ZM7.5 0C6.11929 0 5 1.11929 5 2.5H6C6 1.67157 6.67157 1 7.5 1V0ZM10 2.5C10 1.11929 8.88071 0 7.5 0V1C8.32843 1 9 1.67157 9 2.5H10ZM10 3.5V2.5H9V3.5H10ZM12.5 3H9.5V4H12.5V3ZM13 6.5V3.5H12V6.5H13ZM12.5 7H13.5V6H12.5V7ZM13.5 7C14.3284 7 15 7.67157 15 8.5H16C16 7.11929 14.8807 6 13.5 6V7ZM15 8.5C15 9.32843 14.3284 10 13.5 10V11C14.8807 11 16 9.88071 16 8.5H15ZM13.5 10H12.5V11H13.5V10ZM13 13.5V10.5H12V13.5H13ZM2.5 14H12.5V13H2.5V14ZM5.5 3H2.5V4H5.5V3ZM5 2.5V3.5H6V2.5H5ZM2 3.5V6.5H3V3.5H2ZM2 10.5V13.5H3V10.5H2ZM2.5 7H3.5V6H2.5V7ZM3.5 10H2.5V11H3.5V10ZM5 8.5C5 9.32843 4.32843 10 3.5 10V11C4.88071 11 6 9.88071 6 8.5H5ZM3.5 7C4.32843 7 5 7.67157 5 8.5H6C6 7.11929 4.88071 6 3.5 6V7Z",
			fill: "#548AF7"
		})
	})
}), {
	__source: "jb-download:extension-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-extension",
	displayName: "UiExtensionDark"
}), sn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M12.5 15.5V9.5M9.90193 14L15.0981 11M15.0981 14L9.90192 11",
				stroke: "#3574F0",
				strokeLinecap: "round"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M3 5.41421V13C3 14.1046 3.89543 15 5 15H8.75777C8.27914 14.285 8 13.4251 8 12.5C8 10.0147 10.0147 8 12.5 8C12.669 8 12.8358 8.00932 13 8.02746V3C13 1.89543 12.1046 1 11 1H7.41421C7.149 1 6.89464 1.10536 6.70711 1.29289L3.29289 4.70711C3.10536 4.89464 3 5.149 3 5.41421Z",
				fill: "#EBECF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M3 5.41421V13C3 14.1046 3.89543 15 5 15H8.75777C8.55127 14.6915 8.3819 14.356 8.25606 14H5C4.44772 14 4 13.5523 4 13L4 6H6C7.10457 6 8 5.10457 8 4V2L11 2C11.5523 2 12 2.44772 12 3V8.02746C12.1642 8.00932 12.331 8 12.5 8C12.669 8 12.8358 8.00932 13 8.02746V3C13 1.89543 12.1046 1 11 1H7.41421C7.149 1 6.89464 1.10536 6.70711 1.29289L3.29289 4.70711C3.10536 4.89464 3 5.149 3 5.41421ZM7 2.41421L4.41421 5H6C6.55228 5 7 4.55228 7 4V2.41421Z",
				fill: "#6C707E"
			})
		]
	})
}), {
	__source: "jb-download:file-format",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-file-format",
	displayName: "UiFileFormat"
}), cn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M12.5 15.5V9.5M9.90193 14L15.0981 11M15.0981 14L9.90192 11",
				stroke: "#548AF7",
				strokeLinecap: "round"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M3 5.41421V13C3 14.1046 3.89543 15 5 15H8.75777C8.27914 14.285 8 13.4251 8 12.5C8 10.0147 10.0147 8 12.5 8C12.669 8 12.8358 8.00932 13 8.02746V3C13 1.89543 12.1046 1 11 1H7.41421C7.149 1 6.89464 1.10536 6.70711 1.29289L3.29289 4.70711C3.10536 4.89464 3 5.149 3 5.41421Z",
				fill: "#43454A"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M3 5.41421V13C3 14.1046 3.89543 15 5 15H8.75777C8.55127 14.6915 8.3819 14.356 8.25606 14H5C4.44772 14 4 13.5523 4 13L4 6H6C7.10457 6 8 5.10457 8 4V2L11 2C11.5523 2 12 2.44772 12 3V8.02746C12.1642 8.00932 12.331 8 12.5 8C12.669 8 12.8358 8.00932 13 8.02746V3C13 1.89543 12.1046 1 11 1H7.41421C7.149 1 6.89464 1.10536 6.70711 1.29289L3.29289 4.70711C3.10536 4.89464 3 5.149 3 5.41421ZM7 2.41421L4.41421 5H6C6.55228 5 7 4.55228 7 4V2.41421Z",
				fill: "#CED0D6"
			})
		]
	})
}), {
	__source: "jb-download:file-format-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-file-format",
	displayName: "UiFileFormatDark"
}), ln = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			d: "M10 13L14 8L13.5 11H16L12 16L12.5 13H10Z",
			fill: "#4682FA"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			fillRule: "evenodd",
			clipRule: "evenodd",
			d: "M2 3.39246H2.01664C2.07865 3.00405 2.30765 2.67483 2.59694 2.41334C2.9371 2.10586 3.40129 1.85189 3.93236 1.64885C4.99792 1.24147 6.43619 1 8 1C9.56381 1 11.0021 1.24147 12.0676 1.64885C12.5987 1.85189 13.0629 2.10586 13.4031 2.41334C13.6924 2.67483 13.9213 3.00405 13.9834 3.39246H14V6.3984L12.445 8.34212C12.3238 8.398 12.1977 8.45079 12.0676 8.50053C11.0021 8.90791 9.56381 9.14938 8 9.14938C6.43619 9.14938 4.99792 8.90791 3.93236 8.50053C3.58833 8.369 3.27236 8.21609 3 8.04112V9.49048C3 9.59029 3.05102 9.74236 3.2675 9.93804C3.48396 10.1337 3.82523 10.3328 4.28947 10.5103C5.21452 10.864 6.52624 11.0932 8 11.0932C8.88175 11.0932 9.70549 11.0112 10.4231 10.8695L9.5057 12.0163C9.02362 12.0667 8.51872 12.0932 8 12.0932C6.43619 12.0932 4.99792 11.8518 3.93236 11.4444C3.58833 11.3128 3.27236 11.1599 3 10.985V12.4343C3 12.5341 3.05102 12.6862 3.2675 12.8819C3.48396 13.0775 3.82523 13.2767 4.28947 13.4542C5.21452 13.8078 6.52624 14.0371 8 14.0371C8.34306 14.0371 8.67734 14.0247 9 14.0011V15.0036C8.6743 15.0257 8.34012 15.0371 8 15.0371C6.43619 15.0371 4.99792 14.7956 3.93236 14.3882C3.40129 14.1852 2.9371 13.9312 2.59694 13.6237C2.2768 13.3344 2.03051 12.9621 2.00263 12.5184H2V3.39246ZM3 3.60275C3 3.50293 3.05102 3.35087 3.2675 3.15519C3.48396 2.95953 3.82523 2.7604 4.28947 2.58291C5.21452 2.22925 6.52624 2 8 2C9.47376 2 10.7855 2.22925 11.7105 2.58291C12.1748 2.7604 12.516 2.95953 12.7325 3.15519C12.949 3.35087 13 3.50293 13 3.60275C13 3.70256 12.949 3.85463 12.7325 4.05031C12.516 4.24597 12.1748 4.4451 11.7105 4.62258C10.7855 4.97625 9.47376 5.2055 8 5.2055C6.52624 5.2055 5.21452 4.97625 4.28947 4.62258C3.82523 4.4451 3.48396 4.24597 3.2675 4.05031C3.05102 3.85463 3 3.70256 3 3.60275ZM3 6.54663V5.09723C3.27236 5.27221 3.58833 5.42512 3.93236 5.55665C4.99792 5.96403 6.43619 6.2055 8 6.2055C9.56381 6.2055 11.0021 5.96403 12.0676 5.55665C12.4117 5.42512 12.7276 5.27221 13 5.09723V6.54663C13 6.64645 12.949 6.79851 12.7325 6.99419C12.516 7.18985 12.1748 7.38898 11.7105 7.56647C10.7855 7.92013 9.47376 8.14938 8 8.14938C6.52624 8.14938 5.21452 7.92013 4.28947 7.56647C3.82523 7.38898 3.48396 7.18985 3.2675 6.99419C3.05102 6.79851 3 6.64645 3 6.54663Z",
			fill: "#6C707E"
		})]
	})
}), {
	__source: "jb-download-unverified:hibernate-event",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-hibernate-event",
	displayName: "UiHibernateEvent"
}), un = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			d: "M10 13L14 8L13.5 11H16L12 16L12.5 13H10Z",
			fill: "#548AF7"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			fillRule: "evenodd",
			clipRule: "evenodd",
			d: "M2 3.39246H2.01664C2.07865 3.00405 2.30765 2.67483 2.59694 2.41334C2.9371 2.10586 3.40129 1.85189 3.93236 1.64885C4.99792 1.24147 6.43619 1 8 1C9.56381 1 11.0021 1.24147 12.0676 1.64885C12.5987 1.85189 13.0629 2.10586 13.4031 2.41334C13.6924 2.67483 13.9213 3.00405 13.9834 3.39246H14V6.3984L12.445 8.34212C12.3238 8.398 12.1977 8.45079 12.0676 8.50053C11.0021 8.90791 9.56381 9.14938 8 9.14938C6.43619 9.14938 4.99792 8.90791 3.93236 8.50053C3.58833 8.369 3.27236 8.21609 3 8.04112V9.49048C3 9.59029 3.05102 9.74236 3.2675 9.93804C3.48396 10.1337 3.82523 10.3328 4.28947 10.5103C5.21452 10.864 6.52624 11.0932 8 11.0932C8.88175 11.0932 9.70549 11.0112 10.4231 10.8695L9.5057 12.0163C9.02362 12.0667 8.51872 12.0932 8 12.0932C6.43619 12.0932 4.99792 11.8518 3.93236 11.4444C3.58833 11.3128 3.27236 11.1599 3 10.985V12.4343C3 12.5341 3.05102 12.6862 3.2675 12.8819C3.48396 13.0775 3.82523 13.2767 4.28947 13.4542C5.21452 13.8078 6.52624 14.0371 8 14.0371C8.34306 14.0371 8.67734 14.0247 9 14.0011V15.0036C8.6743 15.0257 8.34012 15.0371 8 15.0371C6.43619 15.0371 4.99792 14.7956 3.93236 14.3882C3.40129 14.1852 2.9371 13.9312 2.59694 13.6237C2.2768 13.3344 2.03051 12.9621 2.00263 12.5184H2V3.39246ZM3 3.60275C3 3.50293 3.05102 3.35087 3.2675 3.15519C3.48396 2.95953 3.82523 2.7604 4.28947 2.58291C5.21452 2.22925 6.52624 2 8 2C9.47376 2 10.7855 2.22925 11.7105 2.58291C12.1748 2.7604 12.516 2.95953 12.7325 3.15519C12.949 3.35087 13 3.50293 13 3.60275C13 3.70256 12.949 3.85463 12.7325 4.05031C12.516 4.24597 12.1748 4.4451 11.7105 4.62258C10.7855 4.97625 9.47376 5.2055 8 5.2055C6.52624 5.2055 5.21452 4.97625 4.28947 4.62258C3.82523 4.4451 3.48396 4.24597 3.2675 4.05031C3.05102 3.85463 3 3.70256 3 3.60275ZM3 6.54663V5.09723C3.27236 5.27221 3.58833 5.42512 3.93236 5.55665C4.99792 5.96403 6.43619 6.2055 8 6.2055C9.56381 6.2055 11.0021 5.96403 12.0676 5.55665C12.4117 5.42512 12.7276 5.27221 13 5.09723V6.54663C13 6.64645 12.949 6.79851 12.7325 6.99419C12.516 7.18985 12.1748 7.38898 11.7105 7.56647C10.7855 7.92013 9.47376 8.14938 8 8.14938C6.52624 8.14938 5.21452 7.92013 4.28947 7.56647C3.82523 7.38898 3.48396 7.18985 3.2675 6.99419C3.05102 6.79851 3 6.64645 3 6.54663Z",
			fill: "#CED0D6"
		})]
	})
}), {
	__source: "jb-download-unverified:hibernate-event-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-hibernate-event",
	displayName: "UiHibernateEventDark"
}), dn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M11 3C11.5523 3 12 3.44772 12 4V5H15.5C15.7761 5 16 5.22386 16 5.5C16 5.77614 15.7761 6 15.5 6H12V8H11V4H7C5.34315 4 4 5.34315 4 7V9C4 10.6569 5.34315 12 7 12H8V13H7C5.13616 13 3.57002 11.7252 3.12598 10H1C0.447715 10 0 9.55228 0 9V7C0 6.44772 0.447715 6 1 6H3.12598C3.57002 4.27477 5.13616 3 7 3H11ZM1 9H3V7H1V9Z",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "9.5",
				y: "9.5",
				width: "6",
				height: "6",
				rx: "1.5",
				fill: "#EDF3FF",
				stroke: "#3574F0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M11 8C9.34315 8 8 9.34315 8 11V12H7C5.34315 12 4 10.6569 4 9V7C4 5.34315 5.34315 4 7 4H11V8Z",
				fill: "#EBECF0"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M3 9H1V7H3V9Z",
				fill: "#EBECF0"
			})
		]
	})
}), {
	__source: "jb-download:plugin-module",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-plugin-module",
	displayName: "UiPluginModule"
}), fn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M11 3C11.5523 3 12 3.44772 12 4V5H15.5C15.7761 5 16 5.22386 16 5.5C16 5.77614 15.7761 6 15.5 6H12V8H11V4H7C5.34315 4 4 5.34315 4 7V9C4 10.6569 5.34315 12 7 12H8V13H7C5.13616 13 3.57002 11.7252 3.12598 10H1C0.447715 10 0 9.55228 0 9V7C0 6.44772 0.447715 6 1 6H3.12598C3.57002 4.27477 5.13616 3 7 3H11ZM1 9H3V7H1V9Z",
				fill: "#CED0D6"
			}),
			/* @__PURE__ */ (0, J.jsx)("rect", {
				x: "9.5",
				y: "9.5",
				width: "6",
				height: "6",
				rx: "1.5",
				fill: "#25324D",
				stroke: "#548AF7"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M11 8C9.34315 8 8 9.34315 8 11V12H7C5.34315 12 4 10.6569 4 9V7C4 5.34315 5.34315 4 7 4H11V8Z",
				fill: "#43454A"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M3 9H1V7H3V9Z",
				fill: "#43454A"
			})
		]
	})
}), {
	__source: "jb-download:plugin-module-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-plugin-module",
	displayName: "UiPluginModuleDark"
}), pn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 14 14",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			fillRule: "evenodd",
			clipRule: "evenodd",
			d: "M3.14635 3.64645C3.34162 3.45118 3.6582 3.45118 3.85347 3.64645C4.04873 3.84171 4.04873 4.15829 3.85347 4.35355L1.707 6.5L7.70698 6.5C7.98312 6.5 8.20698 6.72386 8.20698 7C8.20698 7.27614 7.98312 7.5 7.70698 7.5L1.707 7.5L3.85347 9.64645C4.04873 9.84171 4.04873 10.1583 3.85347 10.3536C3.6582 10.5488 3.34162 10.5488 3.14635 10.3536L0.146326 7.35355C0.0525565 7.25978 -0.000122082 7.13261 -0.00012207 7C-0.000122059 6.86739 0.0525565 6.74021 0.146326 6.64645L3.14635 3.64645Z",
			fill: "#6C707E"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M7.70591 5.5L9.56594 2.25218C9.75803 1.91678 10.2415 1.91678 10.4336 2.25218L12.8671 6.50138C13.0436 6.80948 13.0436 7.18808 12.8672 7.49622L10.4336 11.7477C10.2416 12.0833 9.75795 12.0833 9.56589 11.7477L7.70687 8.5H6.55446L8.69814 12.245C9.27433 13.2517 10.7252 13.2517 11.3014 12.245L13.735 7.99355C14.0877 7.37726 14.0876 6.62006 13.7347 6.00386L11.3012 1.75466C10.7249 0.748446 9.27457 0.748446 8.69831 1.75466L6.55334 5.5H7.70591Z",
			fill: "#369650"
		})]
	})
}), {
	__source: "jb-download:read-access",
	__viewBox: "0 0 14 14",
	__group: "programming",
	__consumerName: "programming-read-access",
	displayName: "UiReadAccess"
}), mn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 14 14",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			fillRule: "evenodd",
			clipRule: "evenodd",
			d: "M3.14635 3.64645C3.34162 3.45118 3.6582 3.45118 3.85347 3.64645C4.04873 3.84171 4.04873 4.15829 3.85347 4.35355L1.707 6.5L7.70698 6.5C7.98312 6.5 8.20698 6.72386 8.20698 7C8.20698 7.27614 7.98312 7.5 7.70698 7.5L1.707 7.5L3.85347 9.64645C4.04873 9.84171 4.04873 10.1583 3.85347 10.3536C3.6582 10.5488 3.34162 10.5488 3.14635 10.3536L0.146326 7.35355C0.0525565 7.25978 -0.000122082 7.13261 -0.00012207 7C-0.000122059 6.86739 0.0525565 6.74021 0.146326 6.64645L3.14635 3.64645Z",
			fill: "#CED0D6"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M7.70591 5.5L9.56594 2.25218C9.75803 1.91678 10.2415 1.91678 10.4336 2.25218L12.8671 6.50138C13.0436 6.80948 13.0436 7.18808 12.8672 7.49622L10.4336 11.7477C10.2416 12.0833 9.75795 12.0833 9.56589 11.7477L7.70687 8.5H6.55446L8.69814 12.245C9.27433 13.2517 10.7252 13.2517 11.3014 12.245L13.735 7.99355C14.0877 7.37726 14.0876 6.62006 13.7347 6.00386L11.3012 1.75466C10.7249 0.748446 9.27457 0.748446 8.69831 1.75466L6.55334 5.5H7.70591Z",
			fill: "#57965C"
		})]
	})
}), {
	__source: "jb-download:read-access-dark",
	__viewBox: "0 0 14 14",
	__group: "programming",
	__consumerName: "programming-read-access",
	displayName: "UiReadAccessDark"
}), hn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			d: "M2.9675 14V12.35H4.6175V14H2.9675Z",
			fill: "#6C707E"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M12.905 4.8255L13.1855 5.854L10.678 6.449L12.4205 8.404L11.613 9.0925L9.9725 7.0185L8.3235 9.0925L7.516 8.404L9.2585 6.449L6.751 5.854L7.0315 4.8255L9.505 5.565L9.437 3.1H10.4995L10.4315 5.565L12.905 4.8255Z",
			fill: "#6C707E"
		})]
	})
}), {
	__source: "jb-download:regex",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-regex",
	displayName: "UiRegex"
}), gn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			d: "M2.9675 12.35H4.6175V14H2.9675V12.35Z",
			fill: "#CED0D6"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M11.613 9.0925L9.709 6.687L10.389 6.126L12.4205 8.404L11.613 9.0925ZM8.3235 9.0925L7.516 8.404L9.5475 6.126L10.2275 6.687L8.3235 9.0925ZM9.7345 6.5595L6.751 5.854L7.0315 4.8255L9.964 5.701L9.7345 6.5595ZM9.9725 6.6105C9.84783 6.6105 9.74017 6.568 9.6495 6.483C9.5645 6.39233 9.522 6.28467 9.522 6.16C9.522 6.03533 9.5645 5.9305 9.6495 5.8455C9.74017 5.7605 9.84783 5.718 9.9725 5.718C10.0972 5.718 10.202 5.7605 10.287 5.8455C10.372 5.9305 10.4145 6.03533 10.4145 6.16C10.4145 6.28467 10.372 6.39233 10.287 6.483C10.202 6.568 10.0972 6.6105 9.9725 6.6105ZM9.522 6.16L9.437 3.1H10.4995L10.4145 6.16H9.522ZM10.202 6.5595L9.9725 5.701L12.905 4.8255L13.1855 5.854L10.202 6.5595Z",
			fill: "#CED0D6"
		})]
	})
}), {
	__source: "jb-download:regex-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-regex",
	displayName: "UiRegexDark"
}), _n = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M8.85355 10.8536C9.04882 10.6583 9.04882 10.3417 8.85355 10.1464C8.65829 9.95119 8.34171 9.95119 8.14645 10.1464L6 12.2929L5.99999 2.5C5.99999 2.22386 5.77614 2 5.49999 2C5.22385 2 4.99999 2.22386 4.99999 2.5L4.99999 12.2929L2.85355 10.1464C2.65829 9.95118 2.34171 9.95118 2.14645 10.1464C1.95118 10.3417 1.95118 10.6583 2.14645 10.8536L5.14645 13.8536C5.34171 14.0488 5.65829 14.0488 5.85355 13.8536L8.85355 10.8536Z",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M11.9729 1.89999C12.3738 1.89999 12.7262 1.97112 13.0302 2.11339C13.3373 2.25566 13.575 2.46259 13.7431 2.73419C13.9145 3.00256 14.0002 3.31942 14.0002 3.68479V6.98279H12.9429L12.9332 5.12524H13.1951C13.1951 5.53587 13.1207 5.88992 12.972 6.18739C12.8265 6.48162 12.6163 6.70796 12.3415 6.86639C12.0666 7.02159 11.7417 7.09919 11.3666 7.09919C11.0465 7.09919 10.7588 7.03776 10.5033 6.91489C10.2511 6.79202 10.0523 6.62227 9.90677 6.40564C9.7645 6.18901 9.69337 5.94327 9.69337 5.66844C9.69337 5.40331 9.76127 5.15919 9.89707 4.93609C10.0361 4.71299 10.2269 4.53031 10.4694 4.38804C10.7151 4.24577 10.9916 4.15847 11.2987 4.12614L12.5985 3.98549C12.702 3.97579 12.7812 3.93861 12.8362 3.87394C12.8944 3.80604 12.9235 3.70742 12.9235 3.57809L12.9817 4.65964L11.5461 4.81969C11.4006 4.83586 11.2729 4.87627 11.1629 4.94094C11.053 5.00237 10.9673 5.08644 10.9059 5.19314C10.8444 5.29661 10.8137 5.41462 10.8137 5.54719C10.8137 5.68946 10.8509 5.81556 10.9253 5.92549C10.9996 6.03219 11.1047 6.11626 11.2405 6.17769C11.3763 6.23912 11.5315 6.26984 11.7061 6.26984C11.9486 6.26984 12.162 6.22296 12.3463 6.12919C12.5306 6.03219 12.6729 5.89801 12.7731 5.72664C12.8734 5.55204 12.9235 5.35157 12.9235 5.12524V3.58779C12.9235 3.42289 12.8831 3.27901 12.8022 3.15614C12.7214 3.03004 12.6066 2.93304 12.4579 2.86514C12.3091 2.79724 12.1394 2.76329 11.9486 2.76329C11.7579 2.76329 11.5865 2.79562 11.4345 2.86029C11.2826 2.92496 11.1597 3.01226 11.0659 3.12219C10.9722 3.23212 10.9156 3.35337 10.8962 3.48594H9.81462C9.84695 3.18201 9.96012 2.91041 10.1541 2.67114C10.3514 2.42864 10.6068 2.23949 10.9204 2.10369C11.2373 1.96789 11.5881 1.89999 11.9729 1.89999Z",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M12.7049 9.9551H10.212V9.0336H14.0096V9.9357L11.3906 13.0785H13.9999V14H10.0859V13.0979L12.7049 9.9551Z",
				fill: "#6C707E"
			})
		]
	})
}), {
	__source: "jb-download:sort-alphabetically",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-sort-alphabetically",
	displayName: "UiSortAlphabetically"
}), vn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, J.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M8.85355 10.8536C9.04882 10.6583 9.04882 10.3417 8.85355 10.1464C8.65829 9.95119 8.34171 9.95119 8.14645 10.1464L6 12.2929L5.99999 2.5C5.99999 2.22386 5.77614 2 5.49999 2C5.22385 2 4.99999 2.22386 4.99999 2.5L4.99999 12.2929L2.85355 10.1464C2.65829 9.95118 2.34171 9.95118 2.14645 10.1464C1.95118 10.3417 1.95118 10.6583 2.14645 10.8536L5.14645 13.8536C5.34171 14.0488 5.65829 14.0488 5.85355 13.8536L8.85355 10.8536Z",
				fill: "#CED0D6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M11.9729 1.89999C12.3738 1.89999 12.7262 1.97112 13.0302 2.11339C13.3373 2.25566 13.575 2.46259 13.7431 2.73419C13.9145 3.00256 14.0002 3.31942 14.0002 3.68479V6.98279H12.9429L12.9332 5.12524H13.1951C13.1951 5.53587 13.1207 5.88992 12.972 6.18739C12.8265 6.48162 12.6163 6.70796 12.3415 6.86639C12.0666 7.02159 11.7417 7.09919 11.3666 7.09919C11.0465 7.09919 10.7588 7.03776 10.5033 6.91489C10.2511 6.79202 10.0523 6.62227 9.90677 6.40564C9.7645 6.18901 9.69337 5.94327 9.69337 5.66844C9.69337 5.40331 9.76127 5.15919 9.89707 4.93609C10.0361 4.71299 10.2269 4.53031 10.4694 4.38804C10.7151 4.24577 10.9916 4.15847 11.2987 4.12614L12.5985 3.98549C12.702 3.97579 12.7812 3.93861 12.8362 3.87394C12.8944 3.80604 12.9235 3.70742 12.9235 3.57809L12.9817 4.65964L11.5461 4.81969C11.4006 4.83586 11.2729 4.87627 11.1629 4.94094C11.053 5.00237 10.9673 5.08644 10.9059 5.19314C10.8444 5.29661 10.8137 5.41462 10.8137 5.54719C10.8137 5.68946 10.8509 5.81556 10.9253 5.92549C10.9996 6.03219 11.1047 6.11626 11.2405 6.17769C11.3763 6.23912 11.5315 6.26984 11.7061 6.26984C11.9486 6.26984 12.162 6.22296 12.3463 6.12919C12.5306 6.03219 12.6729 5.89801 12.7731 5.72664C12.8734 5.55204 12.9235 5.35157 12.9235 5.12524V3.58779C12.9235 3.42289 12.8831 3.27901 12.8022 3.15614C12.7214 3.03004 12.6066 2.93304 12.4579 2.86514C12.3091 2.79724 12.1394 2.76329 11.9486 2.76329C11.7579 2.76329 11.5865 2.79562 11.4345 2.86029C11.2826 2.92496 11.1597 3.01226 11.0659 3.12219C10.9722 3.23212 10.9156 3.35337 10.8962 3.48594H9.81462C9.84695 3.18201 9.96012 2.91041 10.1541 2.67114C10.3514 2.42864 10.6068 2.23949 10.9204 2.10369C11.2373 1.96789 11.5881 1.89999 11.9729 1.89999Z",
				fill: "#CED0D6"
			}),
			/* @__PURE__ */ (0, J.jsx)("path", {
				d: "M12.7049 9.9551H10.212V9.0336H14.0096V9.9357L11.3906 13.0785H13.9999V14H10.0859V13.0979L12.7049 9.9551Z",
				fill: "#CED0D6"
			})
		]
	})
}), {
	__source: "jb-download:sort-alphabetically-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-sort-alphabetically",
	displayName: "UiSortAlphabeticallyDark"
}), yn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			d: "M4 2.5H12C12.8284 2.5 13.5 3.17157 13.5 4V12C13.5 12.8284 12.8284 13.5 12 13.5H4C3.17157 13.5 2.5 12.8284 2.5 12V4C2.5 3.17157 3.17157 2.5 4 2.5Z",
			fill: "#FFF4EB",
			stroke: "#E66D17"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M5.5 5.5H10.5M5.5 8H10.5M5.5 10.5H8.5",
			stroke: "#E66D17",
			strokeLinecap: "round"
		})]
	})
}), {
	__source: "jb-download-unverified:text",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-text",
	displayName: "UiText"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			d: "M4 2.5H12C12.8284 2.5 13.5 3.17157 13.5 4V12C13.5 12.8284 12.8284 13.5 12 13.5H4C3.17157 13.5 2.5 12.8284 2.5 12V4C2.5 3.17157 3.17157 2.5 4 2.5Z",
			fill: "#45322B",
			stroke: "#C77D55"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M5.5 5.5H10.5M5.5 8H10.5M5.5 10.5H8.5",
			stroke: "#C77D55",
			strokeLinecap: "round"
		})]
	})
}), {
	__source: "jb-download-unverified:text-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-text",
	displayName: "UiTextDark"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiVariable5.tsx
var bn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("circle", {
			cx: "8",
			cy: "8",
			r: "7",
			fill: "#FFF4EB"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M9.76895 5.38506V6.35862H9.73918C9.39401 5.69 8.75515 5.3046 7.91033 5.3046C6.42182 5.3046 5.30343 6.47931 5.30343 8.04023C5.30343 9.60115 6.42182 10.7759 7.91033 10.7759C8.8501 10.7759 9.62734 10.339 10.0417 9.65425C10.481 10.3357 11.2993 10.7517 12.3276 10.7517C13.9609 10.7517 15.0391 9.60115 15.0391 7.91954C15.0391 4.11379 11.8931 1 8.03906 1C4.18504 1 1.03906 4.18621 1.03906 8.08046C1.03906 11.8862 4.18504 15 8.03906 15C9.88964 15 11.5552 14.2356 12.4724 13.4069L11.7483 12.6103C10.9356 13.3345 9.63217 13.9943 8.03906 13.9943C4.74021 13.9943 2.04481 11.331 2.04481 8.08046C2.04481 4.74138 4.74021 2.00575 8.03906 2.00575C11.3379 2.00575 14.0333 4.66897 14.0333 7.91954C14.0333 9.02989 13.3655 9.78621 12.3276 9.78621C11.3862 9.78621 10.6942 9.1908 10.6942 8.32184V5.38506H9.76895ZM7.99883 9.8908C7.03332 9.8908 6.30918 9.09425 6.30918 8.04023C6.30918 6.98621 7.03332 6.18966 7.99883 6.18966C8.96435 6.18966 9.68849 6.98621 9.68849 8.04023C9.68849 9.09425 8.96435 9.8908 7.99883 9.8908Z",
			fill: "#E66D17"
		})]
	})
}), {
	__source: "jb-download:variable-alt-1",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-variable-alt-1",
	displayName: "UiVariable5"
}), xn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("circle", {
			cx: "8",
			cy: "8",
			r: "7",
			fill: "#45322B"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M9.76895 5.38506V6.35862H9.73918C9.39401 5.69 8.75515 5.3046 7.91033 5.3046C6.42182 5.3046 5.30343 6.47931 5.30343 8.04023C5.30343 9.60115 6.42182 10.7759 7.91033 10.7759C8.8501 10.7759 9.62734 10.339 10.0417 9.65425C10.481 10.3357 11.2993 10.7517 12.3276 10.7517C13.9609 10.7517 15.0391 9.60115 15.0391 7.91954C15.0391 4.11379 11.8931 1 8.03906 1C4.18504 1 1.03906 4.18621 1.03906 8.08046C1.03906 11.8862 4.18504 15 8.03906 15C9.88964 15 11.5552 14.2356 12.4724 13.4069L11.7483 12.6103C10.9356 13.3345 9.63217 13.9943 8.03906 13.9943C4.74021 13.9943 2.04481 11.331 2.04481 8.08046C2.04481 4.74138 4.74021 2.00575 8.03906 2.00575C11.3379 2.00575 14.0333 4.66897 14.0333 7.91954C14.0333 9.02989 13.3655 9.78621 12.3276 9.78621C11.3862 9.78621 10.6942 9.1908 10.6942 8.32184V5.38506H9.76895ZM7.99883 9.8908C7.03332 9.8908 6.30918 9.09425 6.30918 8.04023C6.30918 6.98621 7.03332 6.18966 7.99883 6.18966C8.96435 6.18966 9.68849 6.98621 9.68849 8.04023C9.68849 9.09425 8.96435 9.8908 7.99883 9.8908Z",
			fill: "#C77D55"
		})]
	})
}), {
	__source: "jb-download:variable-alt-1-dark",
	__viewBox: "0 0 16 16",
	__group: "programming",
	__consumerName: "programming-variable-alt-1",
	displayName: "UiVariable5Dark"
}), Sn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 14 14",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			fillRule: "evenodd",
			clipRule: "evenodd",
			d: "M4.85343 10.3536C4.65817 10.5488 4.34159 10.5488 4.14632 10.3536C3.95106 10.1583 3.95106 9.84171 4.14632 9.64645L6.29277 7.5L0.499878 7.5C0.223736 7.5 -0.00012207 7.27614 -0.00012207 7C-0.00012207 6.72386 0.223736 6.5 0.499878 6.5L6.29277 6.5L4.14632 4.35355C3.95106 4.15829 3.95106 3.84171 4.14632 3.64645C4.34159 3.45118 4.65817 3.45118 4.85343 3.64645L7.85343 6.64645C8.04869 6.84171 8.04869 7.15829 7.85343 7.35355L4.85343 10.3536Z",
			fill: "#6C707E"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M7.85712 5.23583L9.56587 2.25218C9.75795 1.91678 10.2414 1.91678 10.4335 2.25218L12.867 6.50138C13.0435 6.80948 13.0435 7.18808 12.8671 7.49622L10.4336 11.7477C10.2415 12.0833 9.75787 12.0833 9.56581 11.7477L7.85769 8.76361L7.12479 9.4965L8.69807 12.245C9.27425 13.2517 10.7251 13.2517 11.3013 12.245L13.7349 7.99355C14.0876 7.37726 14.0876 6.62006 13.7347 6.00386L11.3011 1.75466C10.7249 0.748446 9.2745 0.748446 8.69823 1.75466L7.12427 4.50297L7.85712 5.23583Z",
			fill: "#DB3B4B"
		})]
	})
}), {
	__source: "jb-download:write-access",
	__viewBox: "0 0 14 14",
	__group: "programming",
	__consumerName: "programming-write-access",
	displayName: "UiWriteAccess"
}), Cn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, J.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 14 14",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, J.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, J.jsx)("path", {
			fillRule: "evenodd",
			clipRule: "evenodd",
			d: "M4.85343 10.3536C4.65817 10.5488 4.34159 10.5488 4.14632 10.3536C3.95106 10.1583 3.95106 9.84171 4.14632 9.64645L6.29277 7.5L0.499878 7.5C0.223736 7.5 -0.00012207 7.27614 -0.00012207 7C-0.00012207 6.72386 0.223736 6.5 0.499878 6.5L6.29277 6.5L4.14632 4.35355C3.95106 4.15829 3.95106 3.84171 4.14632 3.64645C4.34159 3.45118 4.65817 3.45118 4.85343 3.64645L7.85343 6.64645C8.04869 6.84171 8.04869 7.15829 7.85343 7.35355L4.85343 10.3536Z",
			fill: "#CED0D6"
		}), /* @__PURE__ */ (0, J.jsx)("path", {
			d: "M7.85712 5.23583L9.56587 2.25218C9.75795 1.91678 10.2414 1.91678 10.4335 2.25218L12.867 6.50138C13.0435 6.80948 13.0435 7.18808 12.8671 7.49622L10.4336 11.7477C10.2415 12.0833 9.75787 12.0833 9.56581 11.7477L7.85769 8.76361L7.12479 9.4965L8.69807 12.245C9.27425 13.2517 10.7251 13.2517 11.3013 12.245L13.7349 7.99355C14.0876 7.37726 14.0876 6.62006 13.7347 6.00386L11.3011 1.75466C10.7249 0.748446 9.2745 0.748446 8.69823 1.75466L7.12427 4.50297L7.85712 5.23583Z",
			fill: "#DB5C5C"
		})]
	})
}), {
	__source: "jb-download:write-access-dark",
	__viewBox: "0 0 14 14",
	__group: "programming",
	__consumerName: "programming-write-access",
	displayName: "UiWriteAccessDark"
});
// ../../../clicky-ui/packages/ui/src/components/ItemActions.tsx
function wn({ label: e, index: t, count: n, onMove: r, onDuplicate: i, onRemove: a, leading: o, reveal: s = !0, size: l = "md", className: u }) {
	let d = q("inline-flex aspect-square items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-foreground disabled:opacity-30", T[l]);
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: q("ml-auto flex shrink-0 items-center gap-0.5", s && "opacity-0 transition-opacity focus-within:opacity-100 group-hover:opacity-100 [@media(hover:none)]:opacity-100", u),
		children: [
			o,
			r && /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)("button", {
				type: "button",
				"aria-label": `Move ${e} up`,
				disabled: t === 0,
				className: d,
				onClick: () => r(t - 1),
				children: /* @__PURE__ */ (0, J.jsx)(c, {
					icon: Ke,
					className: "text-sm"
				})
			}), /* @__PURE__ */ (0, J.jsx)("button", {
				type: "button",
				"aria-label": `Move ${e} down`,
				disabled: t >= n - 1,
				className: d,
				onClick: () => r(t + 1),
				children: /* @__PURE__ */ (0, J.jsx)(c, {
					icon: pe,
					className: "text-sm"
				})
			})] }),
			i && /* @__PURE__ */ (0, J.jsx)("button", {
				type: "button",
				"aria-label": `Duplicate ${e}`,
				className: d,
				onClick: i,
				children: /* @__PURE__ */ (0, J.jsx)(c, {
					icon: Rt,
					className: "text-sm"
				})
			}),
			a && /* @__PURE__ */ (0, J.jsx)("button", {
				type: "button",
				"aria-label": `Remove ${e}`,
				className: q(d, "hover:bg-destructive/10 hover:text-destructive"),
				onClick: a,
				"data-item-index": t,
				children: /* @__PURE__ */ (0, J.jsx)(c, {
					icon: Ht,
					className: "text-sm"
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/AccordionList.tsx
var Tn = "[data-autofocus], input, select, textarea";
function En({ items: e, onChange: t, renderHeader: n, renderBody: r, renderActions: i, renderMeta: a, summary: o, toolbar: s, itemLabel: u, allowReorder: d = !1, allowDrag: f = !1, canDrag: p, onReorder: m, allowDuplicate: h = !1, allowRemove: g = !1, revealActions: _ = !0, cloneItem: v, onCreate: y, addLabel: b = "Add item", addDescription: x, expanded: S, defaultExpanded: C = null, onExpandedChange: w, readOnly: T = !1, size: E = "md", idPrefix: D = "accordion", itemId: k, focusSelector: A = Tn, className: j, listClassName: M, rowClassName: P, bodyClassName: F }) {
	let I = S !== void 0, [L, R] = (0, Y.useState)(C), B = I ? S : L, [V, H] = (0, Y.useState)(null), [U, ee] = (0, Y.useState)(null), te = (0, Y.useRef)([]), W = (0, Y.useRef)(null), ne = (0, Y.useRef)(null);
	(0, Y.useEffect)(() => {
		let e = ne.current;
		if (!e) return;
		if (ne.current = null, e.kind === "add") {
			W.current?.focus();
			return;
		}
		let t = te.current[e.index];
		e.kind === "row" ? t?.focus() : (document.getElementById(ie(e.index))?.querySelector(A) ?? t)?.focus();
	});
	function re(e) {
		return k ? k(e) : `${D}-${e}`;
	}
	function ie(e) {
		return `${re(e)}-panel`;
	}
	function ae(e) {
		return `${re(e)}-header`;
	}
	function G(e) {
		I || R(e), w?.(e);
	}
	function se(e) {
		let t = e(B);
		t !== B && G(t);
	}
	function K(e) {
		t?.(e);
	}
	function ce() {
		y && (ne.current = {
			kind: "panel",
			index: e.length
		}, G(e.length), K([...e, y()]));
	}
	function le(t) {
		se((e) => e === t ? null : e !== null && e > t ? e - 1 : e), ne.current = e.length === 1 ? {
			kind: "add",
			index: 0
		} : {
			kind: "row",
			index: Math.max(t - 1, 0)
		}, K(z(e, t));
	}
	function ue(t) {
		se((e) => e !== null && e > t ? e + 1 : e), ne.current = {
			kind: "row",
			index: t + 1
		}, K(N(e, t, v));
	}
	function de(e) {
		return V === null || U !== e || V === e ? null : V < e ? "bottom" : "top";
	}
	function fe() {
		H(null), ee(null);
	}
	function me(e, t) {
		e !== t && (m ? m(e, t) : ge(e, t));
	}
	function ge(t, n) {
		n < 0 || n >= e.length || (se((e) => e === t ? n : e === n ? t : e), ne.current = {
			kind: "row",
			index: n
		}, K(l(e, t, n)));
	}
	function _e(e, t) {
		let n = te.current.length, r = (t) => {
			e.preventDefault(), (t >= n ? W.current : te.current[t])?.focus();
		};
		e.key === "ArrowDown" ? r(Math.min(t + 1, n)) : e.key === "ArrowUp" && t > 0 ? r(t - 1) : e.key === "Home" ? r(0) : e.key === "End" && r(n);
	}
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: q("flex flex-col gap-2", j),
		children: [(o !== void 0 || s != null) && /* @__PURE__ */ (0, J.jsxs)("div", {
			"data-accordion-summary": !0,
			className: "flex items-center justify-between gap-2",
			children: [o !== void 0 && /* @__PURE__ */ (0, J.jsx)("p", {
				className: q("text-muted-foreground", O[E]),
				children: o
			}), s != null && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "ml-auto flex shrink-0 items-center",
				children: s
			})]
		}), /* @__PURE__ */ (0, J.jsxs)("div", {
			className: q("divide-y divide-border rounded-lg border border-border", M),
			children: [e.map((t, o) => {
				let s = B === o, l = {
					item: t,
					index: o,
					open: s,
					onChange: (t) => K(oe(e, o, t)),
					headerId: ae(o),
					panelId: ie(o)
				}, m = u?.({
					item: t,
					index: o
				}) ?? `Item ${o + 1}`, v = i?.(l), y = a?.(l), b = f && !T && (p?.({
					item: t,
					index: o
				}) ?? !0), x = b && V !== null && V !== o, S = x ? de(o) : null;
				return /* @__PURE__ */ (0, J.jsxs)("div", {
					className: "group",
					children: [/* @__PURE__ */ (0, J.jsxs)("div", {
						"data-accordion-row": !0,
						"data-drop-edge": S ?? void 0,
						onDragOver: (e) => {
							x && (e.preventDefault(), e.dataTransfer.dropEffect = "move", ee(o));
						},
						onDrop: (e) => {
							x && V !== null && (e.preventDefault(), me(V, o), fe());
						},
						className: q("flex w-full items-center gap-2 px-3 py-2 hover:bg-accent/40", o % 2 == 0 ? "bg-card" : "bg-muted/20", s && "bg-accent/30", V === o && "opacity-40", S === "top" && "shadow-[inset_0_2px_0_0_var(--color-primary)]", S === "bottom" && "shadow-[inset_0_-2px_0_0_var(--color-primary)]", P),
						children: [
							f && !T && /* @__PURE__ */ (0, J.jsx)("button", {
								type: "button",
								"aria-label": `Reorder ${m}`,
								title: `Reorder ${m}`,
								disabled: !b,
								draggable: b,
								onDragStart: (e) => {
									e.dataTransfer.effectAllowed = "move", e.dataTransfer.setData("text/plain", String(o));
									let t = e.currentTarget.closest("[data-accordion-row]");
									t instanceof HTMLElement && e.dataTransfer.setDragImage(t, 12, 12), H(o);
								},
								onDragEnd: fe,
								onKeyDown: (t) => {
									let n = t.key === "ArrowUp" ? o - 1 : t.key === "ArrowDown" ? o + 1 : o;
									n === o || n < 0 || n >= e.length || (t.preventDefault(), me(o, n));
								},
								className: "shrink-0 cursor-grab rounded text-muted-foreground hover:text-foreground active:cursor-grabbing disabled:cursor-default disabled:opacity-30",
								children: /* @__PURE__ */ (0, J.jsx)(c, {
									icon: ct,
									className: "text-sm"
								})
							}),
							/* @__PURE__ */ (0, J.jsxs)("button", {
								type: "button",
								id: ae(o),
								"aria-expanded": s,
								"aria-controls": ie(o),
								ref: (e) => {
									te.current[o] = e;
								},
								onKeyDown: (e) => _e(e, o),
								onClick: () => G(s ? null : o),
								className: "flex min-w-0 flex-1 items-center gap-2 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
								children: [/* @__PURE__ */ (0, J.jsx)(c, {
									icon: s ? pe : he,
									className: "shrink-0 text-sm text-muted-foreground"
								}), n(l)]
							}),
							y,
							!T && (d || h || g || v) && /* @__PURE__ */ (0, J.jsx)(wn, {
								label: m,
								index: o,
								count: e.length,
								size: E,
								reveal: _,
								...v ? { leading: v } : {},
								...d ? { onMove: (e) => ge(o, e) } : {},
								...h ? { onDuplicate: () => ue(o) } : {},
								...g ? { onRemove: () => le(o) } : {}
							})
						]
					}), s && /* @__PURE__ */ (0, J.jsx)("div", {
						id: ie(o),
						role: "region",
						"aria-labelledby": ae(o),
						className: q("border-t border-border bg-muted/20 p-3", F),
						children: r(l)
					})]
				}, o);
			}), !T && y && /* @__PURE__ */ (0, J.jsx)(Dn, {
				buttonRef: W,
				label: b,
				empty: e.length === 0,
				onAdd: ce,
				onKeyDown: (t) => _e(t, e.length),
				...x ? { copy: x } : {}
			})]
		})]
	});
}
function Dn({ label: e, copy: t, empty: n, onAdd: r, onKeyDown: i, buttonRef: a }) {
	return /* @__PURE__ */ (0, J.jsxs)("button", {
		type: "button",
		ref: (e) => {
			a.current = e;
		},
		onClick: r,
		onKeyDown: i,
		className: q("group/add flex w-full items-center gap-2 px-3 py-2 text-left text-muted-foreground hover:bg-accent/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring", n && "items-start py-5"),
		children: [
			/* @__PURE__ */ (0, J.jsx)(c, {
				icon: he,
				className: "shrink-0 text-sm opacity-0"
			}),
			/* @__PURE__ */ (0, J.jsx)("span", {
				className: "inline-flex size-6 shrink-0 items-center justify-center rounded-md border border-dashed border-border text-foreground/60 group-hover/add:border-primary group-hover/add:text-primary",
				children: /* @__PURE__ */ (0, J.jsx)(c, {
					icon: X,
					className: "text-[13px]"
				})
			}),
			/* @__PURE__ */ (0, J.jsxs)("span", {
				className: "flex min-w-0 flex-col gap-0.5",
				children: [/* @__PURE__ */ (0, J.jsx)("span", {
					className: "text-sm font-medium text-foreground",
					children: e
				}), t && /* @__PURE__ */ (0, J.jsx)("span", {
					className: q("text-xs", n ? "max-w-xl leading-snug" : "truncate"),
					children: t
				})]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/IconButton.tsx
var On = (0, Y.forwardRef)(({ icon: e, label: t, iconClassName: n, className: r, type: i = "button", ...a }, o) => /* @__PURE__ */ (0, J.jsx)("button", {
	ref: o,
	type: i,
	"aria-label": t,
	title: t,
	className: q("inline-flex shrink-0 items-center justify-center rounded-sm border-0 bg-transparent p-0 text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50", r),
	...a,
	children: /* @__PURE__ */ (0, J.jsx)(c, {
		icon: e,
		className: q("text-sm", n)
	})
}));
On.displayName = "IconButton";
var kn = "input:not([inert]):not([inert] *),select:not([inert]):not([inert] *),textarea:not([inert]):not([inert] *),a[href]:not([inert]):not([inert] *),area[href]:not([inert]):not([inert] *),button:not([inert]):not([inert] *),[tabindex]:not(slot):not([inert]):not([inert] *),audio[controls]:not([inert]):not([inert] *),video[controls]:not([inert]):not([inert] *),[contenteditable]:not([contenteditable=\"false\"]):not([inert]):not([inert] *),details>summary:first-of-type:not([inert]):not([inert] *),details:not([inert]):not([inert] *)", An = typeof Element > "u", jn = An ? function() {} : Element.prototype.matches || Element.prototype.msMatchesSelector || Element.prototype.webkitMatchesSelector, Mn = !An && Element.prototype.getRootNode ? function(e) {
	return e == null || e.getRootNode == null ? void 0 : e.getRootNode();
} : function(e) {
	return e?.ownerDocument;
}, Nn = function(e, t) {
	t === void 0 && (t = !0);
	var n = e == null || e.getAttribute == null ? void 0 : e.getAttribute("inert");
	return n === "" || n === "true" || t && e && (typeof e.closest == "function" ? e.closest("[inert]") : Nn(e.parentNode));
}, Pn = function(e) {
	var t, n = e == null || e.getAttribute == null || (t = e.getAttribute("contenteditable")) == null ? void 0 : t.toLowerCase();
	return n === "" || n === "true" || n === "plaintext-only";
}, Fn = function(e, t, n) {
	if (Nn(e)) return [];
	var r = Array.prototype.slice.apply(e.querySelectorAll(kn));
	return t && jn.call(e, kn) && r.unshift(e), r = r.filter(n), r;
}, In = function(e, t, n) {
	for (var r = [], i = Array.from(e); i.length;) {
		var a = i.shift();
		if (!Nn(a, !1)) {
			if (a.tagName === "SLOT") {
				var o = a.assignedElements(), s = In(o.length ? o : a.children, !0, n);
				n.flatten ? r.push.apply(r, s) : r.push({
					scopeParent: a,
					candidates: s
				});
			} else {
				jn.call(a, kn) && n.filter(a) && (t || !e.includes(a)) && r.push(a);
				var c = a.shadowRoot || typeof n.getShadowRoot == "function" && n.getShadowRoot(a), l = !Nn(c, !1) && (!n.shadowRootFilter || n.shadowRootFilter(a));
				if (c && l) {
					var u = In(c === !0 ? a.children : c.children, !0, n);
					n.flatten ? r.push.apply(r, u) : r.push({
						scopeParent: a,
						candidates: u
					});
				} else i.unshift.apply(i, a.children);
			}
		}
	}
	return r;
}, Ln = function(e) {
	return !isNaN(parseInt(e.getAttribute("tabindex"), 10));
}, Rn = function(e) {
	if (!e) throw Error("No node provided");
	return e.tabIndex < 0 && (/^(AUDIO|VIDEO|DETAILS)$/.test(e.tagName) || Pn(e)) && !Ln(e) ? 0 : e.tabIndex;
}, zn = function(e, t) {
	var n = Rn(e);
	return n < 0 && t && !Ln(e) ? 0 : n;
}, Bn = function(e, t) {
	return e.tabIndex === t.tabIndex ? e.documentOrder - t.documentOrder : e.tabIndex - t.tabIndex;
}, Vn = function(e) {
	return e.tagName === "INPUT";
}, Hn = function(e) {
	return Vn(e) && e.type === "hidden";
}, Un = function(e) {
	return e.tagName === "DETAILS" && Array.prototype.slice.apply(e.children).some(function(e) {
		return e.tagName === "SUMMARY";
	});
}, Wn = function(e, t) {
	for (var n = 0; n < e.length; n++) if (e[n].checked && e[n].form === t) return e[n];
}, Gn = function(e) {
	if (!e.name) return !0;
	var t = e.form || Mn(e), n = function(e) {
		return t.querySelectorAll("input[type=\"radio\"][name=\"" + e + "\"]");
	}, r;
	if (typeof window < "u" && window.CSS !== void 0 && typeof window.CSS.escape == "function") r = n(window.CSS.escape(e.name));
	else try {
		r = n(e.name);
	} catch (e) {
		return console.error("Looks like you have a radio button with a name attribute containing invalid CSS selector characters and need the CSS.escape polyfill: %s", e.message), !1;
	}
	var i = Wn(r, e.form);
	return !i || i === e;
}, Kn = function(e) {
	return Vn(e) && e.type === "radio";
}, qn = function(e) {
	return Kn(e) && !Gn(e);
}, Jn = function(e) {
	var t = e && Mn(e), n = t?.host, r = !1;
	if (t && t !== e) {
		var i, a;
		for (r = !!((i = n) != null && (i = i.ownerDocument) != null && i.contains(n) || e != null && (a = e.ownerDocument) != null && a.contains(e)); !r && n;) {
			var o;
			t = Mn(n), n = t?.host, r = !!((o = n) != null && (o = o.ownerDocument) != null && o.contains(n));
		}
	}
	return r;
}, Yn = function(e) {
	var t = e.getBoundingClientRect(), n = t.width, r = t.height;
	return n === 0 && r === 0;
}, Xn = function(e, t) {
	var n = t.displayCheck, r = t.getShadowRoot;
	if (n === "full-native" && "checkVisibility" in e) return !e.checkVisibility({
		checkOpacity: !1,
		opacityProperty: !1,
		contentVisibilityAuto: !0,
		visibilityProperty: !0,
		checkVisibilityCSS: !0
	});
	var i = getComputedStyle(e).visibility;
	if (i === "hidden" || i === "collapse") return !0;
	var a = jn.call(e, "details>summary:first-of-type") ? e.parentElement : e;
	if (jn.call(a, "details:not([open]) *")) return !0;
	if (!n || n === "full" || n === "full-native" || n === "legacy-full") {
		if (typeof r == "function") {
			for (var o = e; e;) {
				var s = e.parentElement, c = Mn(e);
				if (s && !s.shadowRoot && r(s) === !0) return Yn(e);
				e = e.assignedSlot ? e.assignedSlot : !s && c !== e.ownerDocument ? c.host : s;
			}
			e = o;
		}
		if (Jn(e)) return !e.getClientRects().length;
		if (n !== "legacy-full") return !0;
	} else if (n === "non-zero-area") return Yn(e);
	return !1;
}, Zn = function(e) {
	if (/^(INPUT|BUTTON|SELECT|TEXTAREA)$/.test(e.tagName)) for (var t = e.parentElement; t;) {
		if (t.tagName === "FIELDSET" && t.disabled) {
			for (var n = 0; n < t.children.length; n++) {
				var r = t.children.item(n);
				if (r.tagName === "LEGEND") return jn.call(t, "fieldset[disabled] *") ? !0 : !r.contains(e);
			}
			return !0;
		}
		t = t.parentElement;
	}
	return !1;
}, Qn = function(e, t) {
	return !(t.disabled || Hn(t) || Xn(t, e) || Un(t) || Zn(t));
}, $n = function(e, t) {
	return !(qn(t) || Rn(t) < 0 || !Qn(e, t));
}, er = function(e) {
	var t = parseInt(e.getAttribute("tabindex"), 10);
	return !!(isNaN(t) || t >= 0);
}, tr = function(e) {
	var t = [], n = [];
	return e.forEach(function(e, r) {
		var i = !!e.scopeParent, a = i ? e.scopeParent : e, o = zn(a, i), s = i ? tr(e.candidates) : a;
		o === 0 ? i ? t.push.apply(t, s) : t.push(a) : n.push({
			documentOrder: r,
			tabIndex: o,
			item: e,
			isScope: i,
			content: s
		});
	}), n.sort(Bn).reduce(function(e, t) {
		return t.isScope ? e.push.apply(e, t.content) : e.push(t.content), e;
	}, []).concat(t);
}, nr = function(e, t) {
	return t ||= {}, tr(t.getShadowRoot ? In([e], t.includeContainer, {
		filter: $n.bind(null, t),
		flatten: !1,
		getShadowRoot: t.getShadowRoot,
		shadowRootFilter: er
	}) : Fn(e, t.includeContainer, $n.bind(null, t)));
}, rr = function(e, t) {
	return t ||= {}, t.getShadowRoot ? In([e], t.includeContainer, {
		filter: Qn.bind(null, t),
		flatten: !0,
		getShadowRoot: t.getShadowRoot
	}) : Fn(e, t.includeContainer, Qn.bind(null, t));
}, ir = function(e, t) {
	if (t ||= {}, !e) throw Error("No node provided");
	return jn.call(e, kn) !== !1 && $n(t, e);
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@floating-ui/react/0.27.20/a5461da18b9b5ae2ae73018151621914115cfa402d046f06e6fc049e6b6c54d2/node_modules/@floating-ui/react/dist/floating-ui.react.utils.mjs
function ar() {
	let e = navigator.userAgentData;
	return e != null && e.platform ? e.platform : navigator.platform;
}
function or() {
	let e = navigator.userAgentData;
	return e && Array.isArray(e.brands) ? e.brands.map((e) => {
		let { brand: t, version: n } = e;
		return t + "/" + n;
	}).join(" ") : navigator.userAgent;
}
function sr() {
	return /apple/i.test(navigator.vendor);
}
function cr() {
	let e = /android/i;
	return e.test(ar()) || e.test(or());
}
function lr() {
	return or().includes("jsdom/");
}
var ur = "data-floating-ui-focusable", dr = "input:not([type='hidden']):not([disabled]),[contenteditable]:not([contenteditable='false']),textarea:not([disabled])", fr = "ArrowLeft", pr = "ArrowRight", mr = "ArrowUp", hr = "ArrowDown";
function gr(e) {
	let t = e.activeElement;
	for (; ((n = t) == null || (n = n.shadowRoot) == null ? void 0 : n.activeElement) != null;) {
		var n;
		t = t.shadowRoot.activeElement;
	}
	return t;
}
function _r(e, t) {
	if (!e || !t) return !1;
	let n = t.getRootNode == null ? void 0 : t.getRootNode();
	if (e.contains(t)) return !0;
	if (n && Pe(n)) {
		let n = t;
		for (; n;) {
			if (e === n) return !0;
			n = n.parentNode || n.host;
		}
	}
	return !1;
}
function vr(e) {
	return "composedPath" in e ? e.composedPath()[0] : e.target;
}
function yr(e, t) {
	if (t == null) return !1;
	if ("composedPath" in e) return e.composedPath().includes(t);
	let n = e;
	return n.target != null && t.contains(n.target);
}
function br(e) {
	return e.matches("html,body");
}
function xr(e) {
	return e?.ownerDocument || document;
}
function Sr(e) {
	return we(e) && e.matches(dr);
}
function Cr(e) {
	return e ? e.getAttribute("role") === "combobox" && Sr(e) : !1;
}
function wr(e) {
	return e ? e.hasAttribute(ur) ? e : e.querySelector("[" + ur + "]") || e : null;
}
function Tr(e, t, n) {
	return n === void 0 && (n = !0), e.filter((e) => e.parentId === t && (!n || e.context?.open)).flatMap((t) => [t, ...Tr(e, t.id, n)]);
}
function Er(e, t) {
	let n, r = -1;
	function i(t, a) {
		a > r && (n = t, r = a), Tr(e, t).forEach((e) => {
			i(e.id, a + 1);
		});
	}
	return i(t, 0), e.find((e) => e.id === n);
}
function Dr(e, t) {
	let n = [], r = e.find((e) => e.id === t)?.parentId;
	for (; r;) {
		let t = e.find((e) => e.id === r);
		r = t?.parentId, t && (n = n.concat(t));
	}
	return n;
}
function Or(e) {
	e.preventDefault(), e.stopPropagation();
}
function kr(e) {
	return "nativeEvent" in e;
}
function Ar(e) {
	return e.mozInputSource === 0 && e.isTrusted ? !0 : cr() && e.pointerType ? e.type === "click" && e.buttons === 1 : e.detail === 0 && !e.pointerType;
}
function jr(e) {
	return lr() ? !1 : !cr() && e.width === 0 && e.height === 0 || cr() && e.width === 1 && e.height === 1 && e.pressure === 0 && e.detail === 0 && e.pointerType === "mouse" || e.width < 1 && e.height < 1 && e.pressure === 0 && e.detail === 0 && e.pointerType === "touch";
}
function Mr(e, t) {
	let n = ["mouse", "pen"];
	return t || n.push("", void 0), n.includes(e);
}
var Nr = typeof document < "u" ? Y.useLayoutEffect : function() {}, Pr = { ...Y };
function Fr(e) {
	let t = Y.useRef(e);
	return Nr(() => {
		t.current = e;
	}), t;
}
var Ir = Pr.useInsertionEffect || ((e) => e());
function Lr(e) {
	let t = Y.useRef(() => {});
	return Ir(() => {
		t.current = e;
	}), Y.useCallback(function() {
		var e = [...arguments];
		return t.current == null ? void 0 : t.current(...e);
	}, []);
}
function Rr(e, t, n) {
	return Math.floor(e / t) !== n;
}
function zr(e, t) {
	return t < 0 || t >= e.current.length;
}
function Br(e, t) {
	return Hr(e, { disabledIndices: t });
}
function Vr(e, t) {
	return Hr(e, {
		decrement: !0,
		startingIndex: e.current.length,
		disabledIndices: t
	});
}
function Hr(e, t) {
	let { startingIndex: n = -1, decrement: r = !1, disabledIndices: i, amount: a = 1 } = t === void 0 ? {} : t, o = n;
	do
		o += r ? -a : a;
	while (o >= 0 && o <= e.current.length - 1 && qr(e, o, i));
	return o;
}
function Ur(e, t) {
	let { event: n, orientation: r, loop: i, rtl: a, cols: o, disabledIndices: s, minIndex: c, maxIndex: l, prevIndex: u, stopEvent: d = !1 } = t, f = u;
	if (n.key === mr) {
		if (d && Or(n), u === -1) f = l;
		else if (f = Hr(e, {
			startingIndex: f,
			amount: o,
			decrement: !0,
			disabledIndices: s
		}), i && (u - o < c || f < 0)) {
			let e = u % o, t = l % o, n = l - (t - e);
			f = t === e ? l : t > e ? n : n - o;
		}
		zr(e, f) && (f = u);
	}
	if (n.key === hr && (d && Or(n), u === -1 ? f = c : (f = Hr(e, {
		startingIndex: u,
		amount: o,
		disabledIndices: s
	}), i && u + o > l && (f = Hr(e, {
		startingIndex: u % o - o,
		amount: o,
		disabledIndices: s
	}))), zr(e, f) && (f = u)), r === "both") {
		let t = Me(u / o);
		n.key === (a ? fr : pr) && (d && Or(n), u % o === o - 1 ? i && (f = Hr(e, {
			startingIndex: u - u % o - 1,
			disabledIndices: s
		})) : (f = Hr(e, {
			startingIndex: u,
			disabledIndices: s
		}), i && Rr(f, o, t) && (f = Hr(e, {
			startingIndex: u - u % o - 1,
			disabledIndices: s
		}))), Rr(f, o, t) && (f = u)), n.key === (a ? pr : fr) && (d && Or(n), u % o === 0 ? i && (f = Hr(e, {
			startingIndex: u + (o - u % o),
			decrement: !0,
			disabledIndices: s
		})) : (f = Hr(e, {
			startingIndex: u,
			decrement: !0,
			disabledIndices: s
		}), i && Rr(f, o, t) && (f = Hr(e, {
			startingIndex: u + (o - u % o),
			decrement: !0,
			disabledIndices: s
		}))), Rr(f, o, t) && (f = u));
		let r = Me(l / o) === t;
		zr(e, f) && (f = i && r ? n.key === (a ? pr : fr) ? l : Hr(e, {
			startingIndex: u - u % o - 1,
			disabledIndices: s
		}) : u);
	}
	return f;
}
function Wr(e, t, n) {
	let r = [], i = 0;
	return e.forEach((e, a) => {
		let { width: o, height: s } = e, c = !1;
		for (n && (i = 0); !c;) {
			let e = [];
			for (let n = 0; n < o; n++) for (let r = 0; r < s; r++) e.push(i + n + r * t);
			i % t + o <= t && e.every((e) => r[e] == null) ? (e.forEach((e) => {
				r[e] = a;
			}), c = !0) : i++;
		}
	}), [...r];
}
function Gr(e, t, n, r, i) {
	if (e === -1) return -1;
	let a = n.indexOf(e), o = t[e];
	switch (i) {
		case "tl": return a;
		case "tr": return o ? a + o.width - 1 : a;
		case "bl": return o ? a + (o.height - 1) * r : a;
		case "br": return n.lastIndexOf(e);
	}
}
function Kr(e, t) {
	return t.flatMap((t, n) => e.includes(t) ? [n] : []);
}
function qr(e, t, n) {
	if (typeof n == "function") return n(t);
	if (n) return n.includes(t);
	let r = e.current[t];
	return r == null || r.hasAttribute("disabled") || r.getAttribute("aria-disabled") === "true";
}
var Jr = () => ({
	getShadowRoot: !0,
	displayCheck: typeof ResizeObserver == "function" && ResizeObserver.toString().includes("[native code]") ? "full" : "none"
});
function Yr(e, t) {
	let n = nr(e, Jr()), r = n.length;
	if (r === 0) return;
	let i = gr(xr(e)), a = n.indexOf(i);
	return n[a === -1 ? t === 1 ? 0 : r - 1 : a + t];
}
function Xr(e) {
	return Yr(xr(e).body, 1) || e;
}
function Zr(e) {
	return Yr(xr(e).body, -1) || e;
}
function Qr(e, t) {
	let n = t || e.currentTarget, r = e.relatedTarget;
	return !r || !_r(n, r);
}
function $r(e) {
	nr(e, Jr()).forEach((e) => {
		e.dataset.tabindex = e.getAttribute("tabindex") || "", e.setAttribute("tabindex", "-1");
	});
}
function ei(e) {
	e.querySelectorAll("[data-tabindex]").forEach((e) => {
		let t = e.dataset.tabindex;
		delete e.dataset.tabindex, t ? e.setAttribute("tabindex", t) : e.removeAttribute("tabindex");
	});
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@floating-ui/react/0.27.20/a5461da18b9b5ae2ae73018151621914115cfa402d046f06e6fc049e6b6c54d2/node_modules/@floating-ui/react/dist/floating-ui.react.mjs
function ti(e) {
	let t = Y.useRef(void 0), n = Y.useCallback((t) => {
		let n = e.map((e) => {
			if (e != null) {
				if (typeof e == "function") {
					let n = e, r = n(t);
					return typeof r == "function" ? r : () => {
						n(null);
					};
				}
				return e.current = t, () => {
					e.current = null;
				};
			}
		});
		return () => {
			n.forEach((e) => e?.());
		};
	}, e);
	return Y.useMemo(() => e.every((e) => e == null) ? null : (e) => {
		t.current &&= (t.current(), void 0), e != null && (t.current = n(e));
	}, e);
}
function ni(e, t) {
	let n = e.compareDocumentPosition(t);
	return n & Node.DOCUMENT_POSITION_FOLLOWING || n & Node.DOCUMENT_POSITION_CONTAINED_BY ? -1 : n & Node.DOCUMENT_POSITION_PRECEDING || n & Node.DOCUMENT_POSITION_CONTAINS ? 1 : 0;
}
var ri = /*#__PURE__*/ Y.createContext({
	register: () => {},
	unregister: () => {},
	map: /*#__PURE__*/ new Map(),
	elementsRef: { current: [] }
});
function ii(e) {
	let { children: t, elementsRef: n, labelsRef: r } = e, [i, a] = Y.useState(() => /* @__PURE__ */ new Set()), o = Y.useCallback((e) => {
		a((t) => new Set(t).add(e));
	}, []), s = Y.useCallback((e) => {
		a((t) => {
			let n = new Set(t);
			return n.delete(e), n;
		});
	}, []), c = Y.useMemo(() => {
		let e = /* @__PURE__ */ new Map();
		return Array.from(i.keys()).sort(ni).forEach((t, n) => {
			e.set(t, n);
		}), e;
	}, [i]);
	return /*#__PURE__*/ (0, J.jsx)(ri.Provider, {
		value: Y.useMemo(() => ({
			register: o,
			unregister: s,
			map: c,
			elementsRef: n,
			labelsRef: r
		}), [
			o,
			s,
			c,
			n,
			r
		]),
		children: t
	});
}
function ai(e) {
	e === void 0 && (e = {});
	let { label: t } = e, { register: n, unregister: r, map: i, elementsRef: a, labelsRef: o } = Y.useContext(ri), [s, c] = Y.useState(null), l = Y.useRef(null), u = Y.useCallback((e) => {
		if (l.current = e, s !== null && (a.current[s] = e, o)) {
			let n = t !== void 0;
			o.current[s] = n ? t : e?.textContent ?? null;
		}
	}, [
		s,
		a,
		o,
		t
	]);
	return Nr(() => {
		let e = l.current;
		if (e) return n(e), () => {
			r(e);
		};
	}, [n, r]), Nr(() => {
		let e = l.current ? i.get(l.current) : null;
		e != null && c(e);
	}, [i]), Y.useMemo(() => ({
		ref: u,
		index: s ?? -1
	}), [s, u]);
}
var oi = "data-floating-ui-focusable", si = "active", ci = "selected", li = "ArrowLeft", ui = "ArrowRight", di = "ArrowUp", fi = "ArrowDown", pi = [li, ui], mi = [di, fi];
[...pi, ...mi];
var hi = { ...Y }, gi = !1, _i = 0, vi = () => "floating-ui-" + Math.random().toString(36).slice(2, 6) + _i++;
function yi() {
	let [e, t] = Y.useState(() => gi ? vi() : void 0);
	return Nr(() => {
		e ?? t(vi());
	}, []), Y.useEffect(() => {
		gi = !0;
	}, []), e;
}
var bi = hi.useId || yi;
function xi() {
	let e = /* @__PURE__ */ new Map();
	return {
		emit(t, n) {
			var r;
			(r = e.get(t)) == null || r.forEach((e) => e(n));
		},
		on(t, n) {
			e.has(t) || e.set(t, /* @__PURE__ */ new Set()), e.get(t).add(n);
		},
		off(t, n) {
			var r;
			(r = e.get(t)) == null || r.delete(n);
		}
	};
}
var Si = /*#__PURE__*/ Y.createContext(null), Ci = /*#__PURE__*/ Y.createContext(null), wi = () => Y.useContext(Si)?.id || null, Ti = () => Y.useContext(Ci);
function Ei(e) {
	let t = bi(), n = Ti(), r = wi(), i = e || r;
	return Nr(() => {
		if (!t) return;
		let e = {
			id: t,
			parentId: i
		};
		return n?.addNode(e), () => {
			n?.removeNode(e);
		};
	}, [
		n,
		t,
		i
	]), t;
}
function Di(e) {
	let { children: t, id: n } = e, r = wi();
	return /*#__PURE__*/ (0, J.jsx)(Si.Provider, {
		value: Y.useMemo(() => ({
			id: n,
			parentId: r
		}), [n, r]),
		children: t
	});
}
function Oi(e) {
	let { children: t } = e, n = Y.useRef([]), r = Y.useCallback((e) => {
		n.current = [...n.current, e];
	}, []), i = Y.useCallback((e) => {
		n.current = n.current.filter((t) => t !== e);
	}, []), [a] = Y.useState(() => xi());
	return /*#__PURE__*/ (0, J.jsx)(Ci.Provider, {
		value: Y.useMemo(() => ({
			nodesRef: n,
			addNode: r,
			removeNode: i,
			events: a
		}), [
			r,
			i,
			a
		]),
		children: t
	});
}
function ki(e) {
	return "data-floating-ui-" + e;
}
function Ai(e) {
	e.current !== -1 && (clearTimeout(e.current), e.current = -1);
}
var ji = /*#__PURE__*/ ki("safe-polygon");
function Mi(e, t, n) {
	if (n && !Mr(n)) return 0;
	if (typeof e == "number") return e;
	if (typeof e == "function") {
		let n = e();
		return typeof n == "number" ? n : n?.[t];
	}
	return e?.[t];
}
function Ni(e) {
	return typeof e == "function" ? e() : e;
}
function Pi(e, t) {
	t === void 0 && (t = {});
	let { open: n, onOpenChange: r, dataRef: i, events: a, elements: o } = e, { enabled: s = !0, delay: c = 0, handleClose: l = null, mouseOnly: u = !1, restMs: d = 0, move: f = !0 } = t, p = Ti(), m = wi(), h = Fr(l), g = Fr(c), _ = Fr(n), v = Fr(d), y = Y.useRef(), b = Y.useRef(-1), x = Y.useRef(), S = Y.useRef(-1), C = Y.useRef(!0), w = Y.useRef(!1), T = Y.useRef(() => {}), E = Y.useRef(!1), D = Lr(() => {
		let e = i.current.openEvent?.type;
		return e?.includes("mouse") && e !== "mousedown";
	});
	Y.useEffect(() => {
		if (!s) return;
		function e(e) {
			let { open: t } = e;
			t || (Ai(b), Ai(S), C.current = !0, E.current = !1);
		}
		return a.on("openchange", e), () => {
			a.off("openchange", e);
		};
	}, [s, a]), Y.useEffect(() => {
		if (!s || !h.current || !n) return;
		function e(e) {
			D() && r(!1, e, "hover");
		}
		let t = xr(o.floating).documentElement;
		return t.addEventListener("mouseleave", e), () => {
			t.removeEventListener("mouseleave", e);
		};
	}, [
		o.floating,
		n,
		r,
		s,
		h,
		D
	]);
	let O = Y.useCallback(function(e, t, n) {
		t === void 0 && (t = !0), n === void 0 && (n = "hover");
		let i = Mi(g.current, "close", y.current);
		i && !x.current ? (Ai(b), b.current = window.setTimeout(() => r(!1, e, n), i)) : t && (Ai(b), r(!1, e, n));
	}, [g, r]), k = Lr(() => {
		T.current(), x.current = void 0;
	}), A = Lr(() => {
		if (w.current) {
			let e = xr(o.floating).body;
			e.style.pointerEvents = "", e.removeAttribute(ji), w.current = !1;
		}
	}), j = Lr(() => i.current.openEvent ? ["click", "mousedown"].includes(i.current.openEvent.type) : !1);
	Y.useEffect(() => {
		if (!s) return;
		function e(e) {
			if (Ai(b), C.current = !1, u && !Mr(y.current) || Ni(v.current) > 0 && !Mi(g.current, "open")) return;
			let t = Mi(g.current, "open", y.current);
			t ? b.current = window.setTimeout(() => {
				_.current || r(!0, e, "hover");
			}, t) : n || r(!0, e, "hover");
		}
		function t(e) {
			if (j()) {
				A();
				return;
			}
			T.current();
			let t = xr(o.floating);
			if (Ai(S), E.current = !1, h.current && i.current.floatingContext) {
				n || Ai(b), x.current = h.current({
					...i.current.floatingContext,
					tree: p,
					x: e.clientX,
					y: e.clientY,
					onClose() {
						A(), k(), j() || O(e, !0, "safe-polygon");
					}
				});
				let r = x.current;
				t.addEventListener("mousemove", r), T.current = () => {
					t.removeEventListener("mousemove", r);
				};
			} else y.current === "touch" && _r(o.floating, e.relatedTarget) || O(e);
		}
		function a(e) {
			j() || i.current.floatingContext && (h.current == null || h.current({
				...i.current.floatingContext,
				tree: p,
				x: e.clientX,
				y: e.clientY,
				onClose() {
					A(), k(), j() || O(e);
				}
			})(e));
		}
		function c() {
			Ai(b);
		}
		function l(e) {
			j() || O(e, !1);
		}
		if (Te(o.domReference)) {
			let r = o.domReference, i = o.floating;
			return n && r.addEventListener("mouseleave", a), f && r.addEventListener("mousemove", e, { once: !0 }), r.addEventListener("mouseenter", e), r.addEventListener("mouseleave", t), i && (i.addEventListener("mouseleave", a), i.addEventListener("mouseenter", c), i.addEventListener("mouseleave", l)), () => {
				n && r.removeEventListener("mouseleave", a), f && r.removeEventListener("mousemove", e), r.removeEventListener("mouseenter", e), r.removeEventListener("mouseleave", t), i && (i.removeEventListener("mouseleave", a), i.removeEventListener("mouseenter", c), i.removeEventListener("mouseleave", l));
			};
		}
	}, [
		o,
		s,
		e,
		u,
		f,
		O,
		k,
		A,
		r,
		n,
		_,
		p,
		g,
		h,
		i,
		j,
		v
	]), Nr(() => {
		var e;
		if (s && n && (e = h.current) != null && (e = e.__options) != null && e.blockPointerEvents && D()) {
			w.current = !0;
			let e = o.floating;
			if (Te(o.domReference) && e) {
				var t;
				let n = xr(o.floating).body;
				n.setAttribute(ji, "");
				let r = o.domReference, i = p == null || (t = p.nodesRef.current.find((e) => e.id === m)) == null || (t = t.context) == null ? void 0 : t.elements.floating;
				return i && (i.style.pointerEvents = ""), n.style.pointerEvents = "none", r.style.pointerEvents = "auto", e.style.pointerEvents = "auto", () => {
					n.style.pointerEvents = "", r.style.pointerEvents = "", e.style.pointerEvents = "";
				};
			}
		}
	}, [
		s,
		n,
		m,
		o,
		p,
		h,
		D
	]), Nr(() => {
		n || (y.current = void 0, E.current = !1, k(), A());
	}, [
		n,
		k,
		A
	]), Y.useEffect(() => () => {
		k(), Ai(b), Ai(S), A();
	}, [
		s,
		o.domReference,
		k,
		A
	]);
	let M = Y.useMemo(() => {
		function e(e) {
			y.current = e.pointerType;
		}
		return {
			onPointerDown: e,
			onPointerEnter: e,
			onMouseMove(e) {
				let { nativeEvent: t } = e;
				function i() {
					!C.current && !_.current && r(!0, t, "hover");
				}
				u && !Mr(y.current) || n || Ni(v.current) === 0 || E.current && e.movementX ** 2 + e.movementY ** 2 < 2 || (Ai(S), y.current === "touch" ? i() : (E.current = !0, S.current = window.setTimeout(i, Ni(v.current))));
			}
		};
	}, [
		u,
		r,
		n,
		_,
		v
	]);
	return Y.useMemo(() => s ? { reference: M } : {}, [s, M]);
}
var Fi = 0;
function Ii(e, t) {
	t === void 0 && (t = {});
	let { preventScroll: n = !1, cancelPrevious: r = !0, sync: i = !1 } = t;
	r && cancelAnimationFrame(Fi);
	let a = () => e?.focus({ preventScroll: n });
	i ? a() : Fi = requestAnimationFrame(a);
}
function Li(e, t) {
	if (!e || !t) return !1;
	let n = t.getRootNode == null ? void 0 : t.getRootNode();
	if (e.contains(t)) return !0;
	if (n && Pe(n)) {
		let n = t;
		for (; n;) {
			if (e === n) return !0;
			n = n.parentNode || n.host;
		}
	}
	return !1;
}
function Ri(e) {
	return "composedPath" in e ? e.composedPath()[0] : e.target;
}
function zi(e) {
	return e?.ownerDocument || document;
}
var Bi = {
	inert: /*#__PURE__*/ new WeakMap(),
	"aria-hidden": /*#__PURE__*/ new WeakMap(),
	none: /*#__PURE__*/ new WeakMap()
};
function Vi(e) {
	return e === "inert" ? Bi.inert : e === "aria-hidden" ? Bi["aria-hidden"] : Bi.none;
}
var Hi = /*#__PURE__*/ new WeakSet(), Ui = {}, Wi = 0, Gi = () => typeof HTMLElement < "u" && "inert" in HTMLElement.prototype;
function Ki(e) {
	return e ? Pe(e) ? e.host : Ki(e.parentNode) : null;
}
var qi = (e, t) => t.map((t) => {
	if (e.contains(t)) return t;
	let n = Ki(t);
	return e.contains(n) ? n : null;
}).filter((e) => e != null);
function Ji(e, t, n, r) {
	let i = "data-floating-ui-inert", a = r ? "inert" : n ? "aria-hidden" : null, o = qi(t, e), s = /* @__PURE__ */ new Set(), c = new Set(o), l = [];
	Ui[i] || (Ui[i] = /* @__PURE__ */ new WeakMap());
	let u = Ui[i];
	o.forEach(d), f(t), s.clear();
	function d(e) {
		e && !s.has(e) && (s.add(e), e.parentNode && d(e.parentNode));
	}
	function f(e) {
		e && !c.has(e) && [].forEach.call(e.children, (e) => {
			if (Ce(e) !== "script") {
				if (s.has(e)) f(e);
				else {
					let t = a ? e.getAttribute(a) : null, n = t !== null && t !== "false", r = Vi(a), o = (r.get(e) || 0) + 1, s = (u.get(e) || 0) + 1;
					r.set(e, o), u.set(e, s), l.push(e), o === 1 && n && Hi.add(e), s === 1 && e.setAttribute(i, ""), !n && a && e.setAttribute(a, a === "inert" ? "" : "true");
				}
			}
		});
	}
	return Wi++, () => {
		l.forEach((e) => {
			let t = Vi(a), n = (t.get(e) || 0) - 1, r = (u.get(e) || 0) - 1;
			t.set(e, n), u.set(e, r), n || (!Hi.has(e) && a && e.removeAttribute(a), Hi.delete(e)), r || e.removeAttribute(i);
		}), Wi--, Wi || (Bi.inert = /* @__PURE__ */ new WeakMap(), Bi["aria-hidden"] = /* @__PURE__ */ new WeakMap(), Bi.none = /* @__PURE__ */ new WeakMap(), Hi = /* @__PURE__ */ new WeakSet(), Ui = {});
	};
}
function Yi(e, t, n) {
	t === void 0 && (t = !1), n === void 0 && (n = !1);
	let r = zi(e[0]).body;
	return Ji(e.concat(Array.from(r.querySelectorAll("[aria-live],[role=\"status\"],output"))), r, t, n);
}
var Xi = {
	border: 0,
	clip: "rect(0 0 0 0)",
	height: "1px",
	margin: "-1px",
	overflow: "hidden",
	padding: 0,
	position: "fixed",
	whiteSpace: "nowrap",
	width: "1px",
	top: 0,
	left: 0
}, Zi = /*#__PURE__*/ Y.forwardRef(function(e, t) {
	let [n, r] = Y.useState();
	Nr(() => {
		sr() && r("button");
	}, []);
	let i = {
		ref: t,
		tabIndex: 0,
		role: n,
		"aria-hidden": !n || void 0,
		[ki("focus-guard")]: "",
		style: Xi
	};
	return /*#__PURE__*/ (0, J.jsx)("span", {
		...e,
		...i
	});
}), Qi = {
	clipPath: "inset(50%)",
	position: "fixed",
	top: 0,
	left: 0
}, $i = /*#__PURE__*/ Y.createContext(null), ea = /*#__PURE__*/ ki("portal");
function ta(e) {
	e === void 0 && (e = {});
	let { id: t, root: n } = e, r = bi(), i = ra(), [a, o] = Y.useState(null), s = Y.useRef(null);
	return Nr(() => () => {
		a?.remove(), queueMicrotask(() => {
			s.current = null;
		});
	}, [a]), Nr(() => {
		if (!r || s.current) return;
		let e = t ? document.getElementById(t) : null;
		if (!e) return;
		let n = document.createElement("div");
		n.id = r, n.setAttribute(ea, ""), e.appendChild(n), s.current = n, o(n);
	}, [t, r]), Nr(() => {
		if (n === null || !r || s.current) return;
		let e = n || i?.portalNode;
		e && !Ne(e) && (e = e.current), e ||= document.body;
		let a = null;
		t && (a = document.createElement("div"), a.id = t, e.appendChild(a));
		let c = document.createElement("div");
		c.id = r, c.setAttribute(ea, ""), e = a || e, e.appendChild(c), s.current = c, o(c);
	}, [
		t,
		n,
		r,
		i
	]), a;
}
function na(e) {
	let { children: t, id: n, root: r, preserveTabOrder: i = !0 } = e, a = ta({
		id: n,
		root: r
	}), [o, s] = Y.useState(null), c = Y.useRef(null), l = Y.useRef(null), u = Y.useRef(null), d = Y.useRef(null), f = o?.modal, p = o?.open, m = !!o && !o.modal && o.open && i && !!(r || a);
	return Y.useEffect(() => {
		if (!a || !i || f) return;
		function e(e) {
			a && Qr(e) && (e.type === "focusin" ? ei : $r)(a);
		}
		return a.addEventListener("focusin", e, !0), a.addEventListener("focusout", e, !0), () => {
			a.removeEventListener("focusin", e, !0), a.removeEventListener("focusout", e, !0);
		};
	}, [
		a,
		i,
		f
	]), Y.useEffect(() => {
		a && (p || ei(a));
	}, [p, a]), /*#__PURE__*/ (0, J.jsxs)($i.Provider, {
		value: Y.useMemo(() => ({
			preserveTabOrder: i,
			beforeOutsideRef: c,
			afterOutsideRef: l,
			beforeInsideRef: u,
			afterInsideRef: d,
			portalNode: a,
			setFocusManagerState: s
		}), [i, a]),
		children: [
			m && a && /*#__PURE__*/ (0, J.jsx)(Zi, {
				"data-type": "outside",
				ref: c,
				onFocus: (e) => {
					if (Qr(e, a)) {
						var t;
						(t = u.current) == null || t.focus();
					} else Zr(o ? o.domReference : null)?.focus();
				}
			}),
			m && a && /*#__PURE__*/ (0, J.jsx)("span", {
				"aria-owns": a.id,
				style: Qi
			}),
			a && /*#__PURE__*/ Fe.createPortal(t, a),
			m && a && /*#__PURE__*/ (0, J.jsx)(Zi, {
				"data-type": "outside",
				ref: l,
				onFocus: (e) => {
					if (Qr(e, a)) {
						var t;
						(t = d.current) == null || t.focus();
					} else Xr(o ? o.domReference : null)?.focus(), o != null && o.closeOnFocusOut && o?.onOpenChange(!1, e.nativeEvent, "focus-out");
				}
			})
		]
	});
}
var ra = () => Y.useContext($i);
function ia(e) {
	return Y.useMemo(() => (t) => {
		e.forEach((e) => {
			e && (e.current = t);
		});
	}, e);
}
var aa = 20, oa = [];
function sa() {
	oa = oa.filter((e) => e.deref()?.isConnected);
}
function ca(e) {
	sa(), e && Ce(e) !== "body" && (oa.push(new WeakRef(e)), oa.length > aa && (oa = oa.slice(-20)));
}
function la() {
	return sa(), oa[oa.length - 1]?.deref();
}
function ua(e) {
	let t = Jr();
	return ir(e, t) ? e : nr(e, t)[0] || e;
}
function da(e, t) {
	var n;
	if (!t.current.includes("floating") && !((n = e.getAttribute("role")) != null && n.includes("dialog"))) return;
	let r = Jr(), i = rr(e, r).filter((e) => {
		let t = e.getAttribute("data-tabindex") || "";
		return ir(e, r) || e.hasAttribute("data-tabindex") && !t.startsWith("-");
	}), a = e.getAttribute("tabindex");
	t.current.includes("floating") || i.length === 0 ? a !== "0" && e.setAttribute("tabindex", "0") : (a !== "-1" || e.hasAttribute("data-tabindex") && e.getAttribute("data-tabindex") !== "-1") && (e.setAttribute("tabindex", "-1"), e.setAttribute("data-tabindex", "-1"));
}
var fa = /*#__PURE__*/ Y.forwardRef(function(e, t) {
	return /*#__PURE__*/ (0, J.jsx)("button", {
		...e,
		type: "button",
		ref: t,
		tabIndex: -1,
		style: Xi
	});
});
function pa(e) {
	let { context: t, children: n, disabled: r = !1, order: i = ["content"], guards: a = !0, initialFocus: o = 0, returnFocus: s = !0, restoreFocus: c = !1, modal: l = !0, visuallyHiddenDismiss: u = !1, closeOnFocusOut: d = !0, outsideElementsInert: f = !1, getInsideElements: p = () => [] } = e, { open: m, onOpenChange: h, events: g, dataRef: _, elements: { domReference: v, floating: y } } = t, b = Lr(() => _.current.floatingContext?.nodeId), x = Lr(p), S = typeof o == "number" && o < 0, C = Cr(v) && S, w = Gi(), T = !w || a, E = !T || w && f, D = Fr(i), O = Fr(o), k = Fr(s), A = Ti(), j = ra(), M = Y.useRef(null), N = Y.useRef(null), P = Y.useRef(!1), F = Y.useRef(!1), I = Y.useRef(-1), L = Y.useRef(-1), R = j != null, z = wr(y), B = Lr(function(e) {
		return e === void 0 && (e = z), e ? nr(e, Jr()) : [];
	}), V = Lr((e) => {
		let t = B(e);
		return D.current.map((e) => v && e === "reference" ? v : z && e === "floating" ? z : t).filter(Boolean).flat();
	});
	Y.useEffect(() => {
		if (r || !l) return;
		function e(e) {
			if (e.key === "Tab") {
				_r(z, gr(xr(z))) && B().length === 0 && !C && Or(e);
				let t = V(), n = vr(e);
				D.current[0] === "reference" && n === v && (Or(e), e.shiftKey ? Ii(t[t.length - 1]) : Ii(t[1])), D.current[1] === "floating" && n === z && e.shiftKey && (Or(e), Ii(t[0]));
			}
		}
		let t = xr(z);
		return t.addEventListener("keydown", e), () => {
			t.removeEventListener("keydown", e);
		};
	}, [
		r,
		v,
		z,
		l,
		D,
		C,
		B,
		V
	]), Y.useEffect(() => {
		if (r || !y) return;
		function e(e) {
			let t = vr(e), n = B().indexOf(t);
			n !== -1 && (I.current = n);
		}
		return y.addEventListener("focusin", e), () => {
			y.removeEventListener("focusin", e);
		};
	}, [
		r,
		y,
		B
	]), Y.useEffect(() => {
		if (r || !d) return;
		function e() {
			F.current = !0, setTimeout(() => {
				F.current = !1;
			});
		}
		function t(e) {
			let t = e.relatedTarget, n = e.currentTarget, r = vr(e);
			queueMicrotask(() => {
				let i = b(), a = !(_r(v, t) || _r(y, t) || _r(t, y) || _r(j?.portalNode, t) || t != null && t.hasAttribute(ki("focus-guard")) || A && (Tr(A.nodesRef.current, i).find((e) => _r(e.context?.elements.floating, t) || _r(e.context?.elements.domReference, t)) || Dr(A.nodesRef.current, i).find((e) => [e.context?.elements.floating, wr(e.context?.elements.floating)].includes(t) || e.context?.elements.domReference === t)));
				if (n === v && z && da(z, D), c && n !== v && !(r != null && r.isConnected) && gr(xr(z)) === xr(z).body) {
					we(z) && z.focus();
					let e = I.current, t = B(), n = t[e] || t[t.length - 1] || z;
					we(n) && n.focus();
				}
				_.current.insideReactTree ? _.current.insideReactTree = !1 : (C || !l) && t && a && !F.current && t !== la() && (P.current = !0, h(!1, e, "focus-out"));
			});
		}
		let n = !(A || !j);
		function i() {
			Ai(L), _.current.insideReactTree = !0, L.current = window.setTimeout(() => {
				_.current.insideReactTree = !1;
			});
		}
		if (y && we(v)) return v.addEventListener("focusout", t), v.addEventListener("pointerdown", e), y.addEventListener("focusout", t), n && y.addEventListener("focusout", i, !0), () => {
			v.removeEventListener("focusout", t), v.removeEventListener("pointerdown", e), y.removeEventListener("focusout", t), n && y.removeEventListener("focusout", i, !0);
		};
	}, [
		r,
		v,
		y,
		z,
		l,
		A,
		j,
		h,
		d,
		c,
		B,
		C,
		b,
		D,
		_
	]);
	let H = Y.useRef(null), U = Y.useRef(null), ee = ia([H, j?.beforeInsideRef]), te = ia([U, j?.afterInsideRef]);
	Y.useEffect(() => {
		var e, t;
		if (r || !y) return;
		let n = Array.from((j == null || (e = j.portalNode) == null ? void 0 : e.querySelectorAll("[" + ki("portal") + "]")) || []), i = (t = (A ? Dr(A.nodesRef.current, b()) : []).find((e) => Cr(e.context?.elements.domReference || null))) == null || (t = t.context) == null ? void 0 : t.elements.domReference, a = [
			y,
			i,
			...n,
			...x(),
			M.current,
			N.current,
			H.current,
			U.current,
			j?.beforeOutsideRef.current,
			j?.afterOutsideRef.current,
			D.current.includes("reference") || C ? v : null
		].filter((e) => e != null), o = l || C ? Yi(a, !E, E) : Yi(a);
		return () => {
			o();
		};
	}, [
		r,
		v,
		y,
		l,
		D,
		j,
		C,
		T,
		E,
		A,
		b,
		x
	]), Nr(() => {
		if (r || !we(z)) return;
		let e = gr(xr(z));
		queueMicrotask(() => {
			let t = V(z), n = O.current, r = (typeof n == "number" ? t[n] : n.current) || z, i = _r(z, e);
			!S && !i && m && Ii(r, { preventScroll: r === z });
		});
	}, [
		r,
		m,
		z,
		S,
		V,
		O
	]), Nr(() => {
		if (r || !z) return;
		let e = xr(z);
		ca(gr(e));
		function t(e) {
			let { reason: t, event: n, nested: r } = e;
			if (["hover", "safe-polygon"].includes(t) && n.type === "mouseleave" && (P.current = !0), t === "outside-press") {
				if (r) P.current = !1;
				else if (Ar(n) || jr(n)) P.current = !1;
				else {
					let e = !1;
					document.createElement("div").focus({ get preventScroll() {
						return e = !0, !1;
					} }), e ? P.current = !1 : P.current = !0;
				}
			}
		}
		g.on("openchange", t);
		let n = e.createElement("span");
		n.setAttribute("tabindex", "-1"), n.setAttribute("aria-hidden", "true"), Object.assign(n.style, Xi), R && v && v.insertAdjacentElement("afterend", n);
		function i() {
			if (typeof k.current == "boolean") {
				let e = v || la();
				return e && e.isConnected ? e : n;
			}
			return k.current.current || n;
		}
		return () => {
			g.off("openchange", t);
			let r = gr(e), a = _r(y, r) || A && Tr(A.nodesRef.current, b(), !1).some((e) => _r(e.context?.elements.floating, r)), o = i();
			queueMicrotask(() => {
				let t = ua(o);
				k.current && !P.current && we(t) && (t === r || r === e.body || a) && t.focus({ preventScroll: !0 }), n.remove();
			});
		};
	}, [
		r,
		y,
		z,
		k,
		_,
		g,
		A,
		R,
		v,
		b
	]), Y.useEffect(() => (queueMicrotask(() => {
		P.current = !1;
	}), () => {
		queueMicrotask(sa);
	}), [r]), Nr(() => {
		if (!r && j) return j.setFocusManagerState({
			modal: l,
			closeOnFocusOut: d,
			open: m,
			onOpenChange: h,
			domReference: v
		}), () => {
			j.setFocusManagerState(null);
		};
	}, [
		r,
		j,
		l,
		m,
		h,
		d,
		v
	]), Nr(() => {
		r || z && da(z, D);
	}, [
		r,
		z,
		D
	]);
	function W(e) {
		return r || !u || !l ? null : /*#__PURE__*/ (0, J.jsx)(fa, {
			ref: e === "start" ? M : N,
			onClick: (e) => h(!1, e.nativeEvent),
			children: typeof u == "string" ? u : "Dismiss"
		});
	}
	let ne = !r && T && (!l || !C) && (R || l);
	return /*#__PURE__*/ (0, J.jsxs)(J.Fragment, { children: [
		ne && /*#__PURE__*/ (0, J.jsx)(Zi, {
			"data-type": "inside",
			ref: ee,
			onFocus: (e) => {
				if (l) {
					let e = V();
					Ii(i[0] === "reference" ? e[0] : e[e.length - 1]);
				} else if (j != null && j.preserveTabOrder && j.portalNode) {
					if (P.current = !1, Qr(e, j.portalNode)) Xr(v)?.focus();
					else {
						var t;
						(t = j.beforeOutsideRef.current) == null || t.focus();
					}
				}
			}
		}),
		!C && W("start"),
		n,
		W("end"),
		ne && /*#__PURE__*/ (0, J.jsx)(Zi, {
			"data-type": "inside",
			ref: te,
			onFocus: (e) => {
				if (l) Ii(V()[0]);
				else if (j != null && j.preserveTabOrder && j.portalNode) {
					if (d && (P.current = !0), Qr(e, j.portalNode)) Zr(v)?.focus();
					else {
						var t;
						(t = j.afterOutsideRef.current) == null || t.focus();
					}
				}
			}
		})
	] });
}
function ma(e) {
	return we(e.target) && e.target.tagName === "BUTTON";
}
function ha(e) {
	return we(e.target) && e.target.tagName === "A";
}
function ga(e) {
	return Sr(e);
}
function _a(e, t) {
	t === void 0 && (t = {});
	let { open: n, onOpenChange: r, dataRef: i, elements: { domReference: a } } = e, { enabled: o = !0, event: s = "click", toggle: c = !0, ignoreMouse: l = !1, keyboardHandlers: u = !0, stickIfOpen: d = !0 } = t, f = Y.useRef(), p = Y.useRef(!1), m = Y.useMemo(() => ({
		onPointerDown(e) {
			f.current = e.pointerType;
		},
		onMouseDown(e) {
			let t = f.current;
			e.button === 0 && s !== "click" && (Mr(t, !0) && l || (n && c && (!(i.current.openEvent && d) || i.current.openEvent.type === "mousedown") ? r(!1, e.nativeEvent, "click") : (e.preventDefault(), r(!0, e.nativeEvent, "click"))));
		},
		onClick(e) {
			let t = f.current;
			s === "mousedown" && f.current ? f.current = void 0 : Mr(t, !0) && l || (n && c && (!(i.current.openEvent && d) || i.current.openEvent.type === "click") ? r(!1, e.nativeEvent, "click") : r(!0, e.nativeEvent, "click"));
		},
		onKeyDown(e) {
			f.current = void 0, !(e.defaultPrevented || !u || ma(e)) && (e.key === " " && !ga(a) && (e.preventDefault(), p.current = !0), !ha(e) && e.key === "Enter" && r(!(n && c), e.nativeEvent, "click"));
		},
		onKeyUp(e) {
			e.defaultPrevented || !u || ma(e) || ga(a) || e.key === " " && p.current && (p.current = !1, r(!(n && c), e.nativeEvent, "click"));
		}
	}), [
		i,
		a,
		s,
		l,
		u,
		r,
		n,
		d,
		c
	]);
	return Y.useMemo(() => o ? { reference: m } : {}, [o, m]);
}
var va = {
	pointerdown: "onPointerDown",
	mousedown: "onMouseDown",
	click: "onClick"
}, ya = {
	pointerdown: "onPointerDownCapture",
	mousedown: "onMouseDownCapture",
	click: "onClickCapture"
}, ba = (e) => ({
	escapeKey: typeof e == "boolean" ? e : e?.escapeKey ?? !1,
	outsidePress: typeof e == "boolean" ? e : e?.outsidePress ?? !0
});
function xa(e, t) {
	t === void 0 && (t = {});
	let { open: n, onOpenChange: r, elements: i, dataRef: a } = e, { enabled: o = !0, escapeKey: s = !0, outsidePress: c = !0, outsidePressEvent: l = "pointerdown", referencePress: u = !1, referencePressEvent: d = "pointerdown", ancestorScroll: f = !1, bubbles: p, capture: m } = t, h = Ti(), g = Lr(typeof c == "function" ? c : () => !1), _ = typeof c == "function" ? g : c, v = Y.useRef(!1), { escapeKey: y, outsidePress: b } = ba(p), { escapeKey: x, outsidePress: S } = ba(m), C = Y.useRef(!1), w = Lr((e) => {
		if (!n || !o || !s || e.key !== "Escape" || C.current) return;
		let t = a.current.floatingContext?.nodeId, i = h ? Tr(h.nodesRef.current, t) : [];
		if (!y && (e.stopPropagation(), i.length > 0)) {
			let e = !0;
			if (i.forEach((t) => {
				var n;
				(n = t.context) != null && n.open && !t.context.dataRef.current.__escapeKeyBubbles && (e = !1);
			}), !e) return;
		}
		r(!1, kr(e) ? e.nativeEvent : e, "escape-key");
	}), T = Lr((e) => {
		var t;
		let n = () => {
			var t;
			w(e), (t = vr(e)) == null || t.removeEventListener("keydown", n);
		};
		(t = vr(e)) == null || t.addEventListener("keydown", n);
	}), E = Lr((e) => {
		let t = a.current.insideReactTree;
		a.current.insideReactTree = !1;
		let n = v.current;
		if (v.current = !1, l === "click" && n || t || typeof _ == "function" && !_(e)) return;
		let o = vr(e), s = "[" + ki("inert") + "]", c = xr(i.floating).querySelectorAll(s), u = Te(o) ? o : null;
		for (; u && !ve(u);) {
			let e = De(u);
			if (ve(e) || !Te(e)) break;
			u = e;
		}
		if (c.length && Te(o) && !br(o) && !_r(o, i.floating) && Array.from(c).every((e) => !_r(u, e))) return;
		if (we(o) && k) {
			let t = ve(o), n = Se(o), r = /auto|scroll/, i = t || r.test(n.overflowX), a = t || r.test(n.overflowY), s = i && o.clientWidth > 0 && o.scrollWidth > o.clientWidth, c = a && o.clientHeight > 0 && o.scrollHeight > o.clientHeight, l = n.direction === "rtl", u = c && (l ? e.offsetX <= o.offsetWidth - o.clientWidth : e.offsetX > o.clientWidth), d = s && e.offsetY > o.clientHeight;
			if (u || d) return;
		}
		let d = a.current.floatingContext?.nodeId, f = h && Tr(h.nodesRef.current, d).some((t) => yr(e, t.context?.elements.floating));
		if (yr(e, i.floating) || yr(e, i.domReference) || f) return;
		let p = h ? Tr(h.nodesRef.current, d) : [];
		if (p.length > 0) {
			let e = !0;
			if (p.forEach((t) => {
				var n;
				(n = t.context) != null && n.open && !t.context.dataRef.current.__outsidePressBubbles && (e = !1);
			}), !e) return;
		}
		r(!1, e, "outside-press");
	}), D = Lr((e) => {
		var t;
		let n = () => {
			var t;
			E(e), (t = vr(e)) == null || t.removeEventListener(l, n);
		};
		(t = vr(e)) == null || t.addEventListener(l, n);
	});
	Y.useEffect(() => {
		if (!n || !o) return;
		a.current.__escapeKeyBubbles = y, a.current.__outsidePressBubbles = b;
		let e = -1;
		function t(e) {
			r(!1, e, "ancestor-scroll");
		}
		function c() {
			window.clearTimeout(e), C.current = !0;
		}
		function u() {
			e = window.setTimeout(() => {
				C.current = !1;
			}, be() ? 5 : 0);
		}
		let d = xr(i.floating);
		s && (d.addEventListener("keydown", x ? T : w, x), d.addEventListener("compositionstart", c), d.addEventListener("compositionend", u)), _ && d.addEventListener(l, S ? D : E, S);
		let p = [];
		return f && (Te(i.domReference) && (p = Ae(i.domReference)), Te(i.floating) && (p = p.concat(Ae(i.floating))), !Te(i.reference) && i.reference && i.reference.contextElement && (p = p.concat(Ae(i.reference.contextElement)))), p = p.filter((e) => e !== d.defaultView?.visualViewport), p.forEach((e) => {
			e.addEventListener("scroll", t);
		}), () => {
			s && (d.removeEventListener("keydown", x ? T : w, x), d.removeEventListener("compositionstart", c), d.removeEventListener("compositionend", u)), _ && d.removeEventListener(l, S ? D : E, S), p.forEach((e) => {
				e.removeEventListener("scroll", t);
			}), window.clearTimeout(e);
		};
	}, [
		a,
		i,
		s,
		_,
		l,
		n,
		r,
		f,
		o,
		y,
		b,
		w,
		x,
		T,
		E,
		S,
		D
	]), Y.useEffect(() => {
		a.current.insideReactTree = !1;
	}, [
		a,
		_,
		l
	]);
	let O = Y.useMemo(() => ({
		onKeyDown: w,
		...u && {
			[va[d]]: (e) => {
				r(!1, e.nativeEvent, "reference-press");
			},
			...d !== "click" && { onClick(e) {
				r(!1, e.nativeEvent, "reference-press");
			} }
		}
	}), [
		w,
		r,
		u,
		d
	]), k = Y.useMemo(() => {
		function e(e) {
			e.button === 0 && (v.current = !0);
		}
		return {
			onKeyDown: w,
			onMouseDown: e,
			onMouseUp: e,
			[ya[l]]: () => {
				a.current.insideReactTree = !0;
			}
		};
	}, [
		w,
		l,
		a
	]);
	return Y.useMemo(() => o ? {
		reference: O,
		floating: k
	} : {}, [
		o,
		O,
		k
	]);
}
function Sa(e) {
	let { open: t = !1, onOpenChange: n, elements: r } = e, i = bi(), a = Y.useRef({}), [o] = Y.useState(() => xi()), s = wi() != null, [c, l] = Y.useState(r.reference), u = Lr((e, t, r) => {
		a.current.openEvent = e ? t : void 0, o.emit("openchange", {
			open: e,
			event: t,
			reason: r,
			nested: s
		}), n?.(e, t, r);
	}), d = Y.useMemo(() => ({ setPositionReference: l }), []), f = Y.useMemo(() => ({
		reference: c || r.reference || null,
		floating: r.floating || null,
		domReference: r.reference
	}), [
		c,
		r.reference,
		r.floating
	]);
	return Y.useMemo(() => ({
		dataRef: a,
		open: t,
		onOpenChange: u,
		elements: f,
		events: o,
		floatingId: i,
		refs: d
	}), [
		t,
		u,
		f,
		o,
		i,
		d
	]);
}
function Ca(e) {
	let { elements: t, ...n } = e === void 0 ? {} : e, { nodeId: r } = n, i = Sa({
		...n,
		elements: {
			reference: t?.reference ?? null,
			floating: t?.floating ?? null
		}
	}), a = n.rootContext || i, o = a.elements, [s, c] = Y.useState(null), [l, u] = Y.useState(null), d = o?.domReference || s, f = Y.useRef(null), p = Ti();
	Nr(() => {
		d && (f.current = d);
	}, [d]);
	let m = xe({
		...n,
		elements: {
			...o,
			...l && { reference: l }
		}
	}), h = Y.useCallback((e) => {
		let t = Te(e) ? {
			getBoundingClientRect: () => e.getBoundingClientRect(),
			getClientRects: () => e.getClientRects(),
			contextElement: e
		} : e;
		u(t), m.refs.setReference(t);
	}, [m.refs]), g = Y.useCallback((e) => {
		(Te(e) || e === null) && (f.current = e, c(e)), (Te(m.refs.reference.current) || m.refs.reference.current === null || e !== null && !Te(e)) && m.refs.setReference(e);
	}, [m.refs]), _ = Y.useMemo(() => ({
		...m.refs,
		setReference: g,
		setPositionReference: h,
		domReference: f
	}), [
		m.refs,
		g,
		h
	]), v = Y.useMemo(() => ({
		...m.elements,
		domReference: d
	}), [m.elements, d]), y = Y.useMemo(() => ({
		...m,
		...a,
		refs: _,
		elements: v,
		nodeId: r
	}), [
		m,
		_,
		v,
		r,
		a
	]);
	return Nr(() => {
		a.dataRef.current.floatingContext = y;
		let e = p?.nodesRef.current.find((e) => e.id === r);
		e && (e.context = y);
	}), Y.useMemo(() => ({
		...m,
		context: y,
		refs: _,
		elements: v
	}), [
		m,
		_,
		v,
		y
	]);
}
function wa(e, t, n) {
	let r = /* @__PURE__ */ new Map(), i = n === "item", a = e;
	if (i && e) {
		let { [si]: t, [ci]: n, ...r } = e;
		a = r;
	}
	return {
		...n === "floating" && {
			tabIndex: -1,
			[oi]: ""
		},
		...a,
		...t.map((t) => {
			let r = t ? t[n] : null;
			return typeof r == "function" ? e ? r(e) : null : r;
		}).concat(e).reduce((e, t) => (t && Object.entries(t).forEach((t) => {
			let [n, a] = t;
			if (!(i && [si, ci].includes(n))) {
				if (n.indexOf("on") === 0) {
					if (r.has(n) || r.set(n, []), typeof a == "function") {
						var o;
						(o = r.get(n)) == null || o.push(a), e[n] = function() {
							var e = [...arguments];
							return r.get(n)?.map((t) => t(...e)).find((e) => e !== void 0);
						};
					}
				} else e[n] = a;
			}
		}), e), {})
	};
}
function Ta(e) {
	e === void 0 && (e = []);
	let t = e.map((e) => e?.reference), n = e.map((e) => e?.floating), r = e.map((e) => e?.item), i = Y.useCallback((t) => wa(t, e, "reference"), t), a = Y.useCallback((t) => wa(t, e, "floating"), n), o = Y.useCallback((t) => wa(t, e, "item"), r);
	return Y.useMemo(() => ({
		getReferenceProps: i,
		getFloatingProps: a,
		getItemProps: o
	}), [
		i,
		a,
		o
	]);
}
var Ea = "Escape";
function Da(e, t, n) {
	switch (e) {
		case "vertical": return t;
		case "horizontal": return n;
		default: return t || n;
	}
}
function Oa(e, t) {
	return Da(t, e === di || e === fi, e === li || e === ui);
}
function ka(e, t, n) {
	return Da(t, e === fi, n ? e === li : e === ui) || e === "Enter" || e === " " || e === "";
}
function Aa(e, t, n) {
	return Da(t, n ? e === li : e === ui, e === fi);
}
function ja(e, t, n, r) {
	return t === "both" || t === "horizontal" && r && r > 1 ? e === Ea : Da(t, n ? e === ui : e === li, e === di);
}
function Ma(e, t) {
	let { open: n, onOpenChange: r, elements: i, floatingId: a } = e, { listRef: o, activeIndex: s, onNavigate: c = () => {}, enabled: l = !0, selectedIndex: u = null, allowEscape: d = !1, loop: f = !1, nested: p = !1, rtl: m = !1, virtual: h = !1, focusItemOnOpen: g = "auto", focusItemOnHover: _ = !0, openOnArrowKeyDown: v = !0, disabledIndices: y = void 0, orientation: b = "vertical", parentOrientation: x, cols: S = 1, scrollItemIntoView: C = !0, virtualItemRef: w, itemSizes: T, dense: E = !1 } = t, D = Fr(wr(i.floating)), O = wi(), k = Ti();
	Nr(() => {
		e.dataRef.current.orientation = b;
	}, [e, b]);
	let A = Lr(() => {
		c(N.current === -1 ? null : N.current);
	}), j = Cr(i.domReference), M = Y.useRef(g), N = Y.useRef(u ?? -1), P = Y.useRef(null), F = Y.useRef(!0), I = Y.useRef(A), L = Y.useRef(!!i.floating), R = Y.useRef(n), z = Y.useRef(!1), B = Y.useRef(!1), V = Fr(y), H = Fr(n), U = Fr(C), ee = Fr(u), [te, W] = Y.useState(), [ne, re] = Y.useState(), ie = Lr(() => {
		function e(e) {
			if (h) {
				var t;
				(t = e.id) != null && t.endsWith("-fui-option") && (e.id = a + "-" + Math.random().toString(16).slice(2, 10)), W(e.id), k?.events.emit("virtualfocus", e), w && (w.current = e);
			} else Ii(e, {
				sync: z.current,
				preventScroll: !0
			});
		}
		let t = o.current[N.current], n = B.current;
		t && e(t), (z.current ? (e) => e() : requestAnimationFrame)(() => {
			let r = o.current[N.current] || t;
			if (!r) return;
			t || e(r);
			let i = U.current;
			i && oe && (n || !F.current) && (r.scrollIntoView == null || r.scrollIntoView(typeof i == "boolean" ? {
				block: "nearest",
				inline: "nearest"
			} : i));
		});
	});
	Nr(() => {
		l && (n && i.floating ? M.current && u != null && (B.current = !0, N.current = u, A()) : L.current && (N.current = -1, I.current()));
	}, [
		l,
		n,
		i.floating,
		u,
		A
	]), Nr(() => {
		if (l && n && i.floating) {
			if (s == null) {
				if (z.current = !1, ee.current != null) return;
				if (L.current && (N.current = -1, ie()), (!R.current || !L.current) && M.current && (P.current != null || M.current === !0 && P.current == null)) {
					let e = 0, t = () => {
						o.current[0] == null ? (e < 2 && (e ? requestAnimationFrame : queueMicrotask)(t), e++) : (N.current = P.current == null || ka(P.current, b, m) || p ? Br(o, V.current) : Vr(o, V.current), P.current = null, A());
					};
					t();
				}
			} else zr(o, s) || (N.current = s, ie(), B.current = !1);
		}
	}, [
		l,
		n,
		i.floating,
		s,
		ee,
		p,
		o,
		b,
		m,
		A,
		ie,
		V
	]), Nr(() => {
		var e;
		if (!l || i.floating || !k || h || !L.current) return;
		let t = k.nodesRef.current, n = (e = t.find((e) => e.id === O)) == null || (e = e.context) == null ? void 0 : e.elements.floating, r = gr(xr(i.floating)), a = t.some((e) => e.context && _r(e.context.elements.floating, r));
		n && !a && F.current && n.focus({ preventScroll: !0 });
	}, [
		l,
		i.floating,
		k,
		O,
		h
	]), Nr(() => {
		if (!l || !k || !h || O) return;
		function e(e) {
			re(e.id), w && (w.current = e);
		}
		return k.events.on("virtualfocus", e), () => {
			k.events.off("virtualfocus", e);
		};
	}, [
		l,
		k,
		h,
		O,
		w
	]), Nr(() => {
		I.current = A, R.current = n, L.current = !!i.floating;
	}), Nr(() => {
		n || (P.current = null, M.current = g);
	}, [n, g]);
	let ae = s != null, oe = Y.useMemo(() => {
		function e(e) {
			if (!H.current) return;
			let t = o.current.indexOf(e);
			t !== -1 && N.current !== t && (N.current = t, A());
		}
		return {
			onFocus(t) {
				let { currentTarget: n } = t;
				z.current = !0, e(n);
			},
			onClick: (e) => {
				let { currentTarget: t } = e;
				return t.focus({ preventScroll: !0 });
			},
			onMouseMove(t) {
				let { currentTarget: n } = t;
				z.current = !0, B.current = !1, _ && e(n);
			},
			onPointerLeave(e) {
				let { pointerType: t } = e;
				if (F.current && t !== "touch" && (z.current = !0, _ && (N.current = -1, A(), !h))) {
					var n;
					(n = D.current) == null || n.focus({ preventScroll: !0 });
				}
			}
		};
	}, [
		H,
		D,
		_,
		o,
		A,
		h
	]), G = Y.useCallback(() => {
		var e;
		return x ?? (k == null || (e = k.nodesRef.current.find((e) => e.id === O)) == null || (e = e.context) == null || (e = e.dataRef) == null ? void 0 : e.current.orientation);
	}, [
		O,
		k,
		x
	]), se = Lr((e) => {
		if (F.current = !1, z.current = !0, e.which === 229 || !H.current && e.currentTarget === D.current) return;
		if (p && ja(e.key, b, m, S)) {
			Oa(e.key, G()) || Or(e), r(!1, e.nativeEvent, "list-navigation"), we(i.domReference) && (h ? k?.events.emit("virtualfocus", i.domReference) : i.domReference.focus());
			return;
		}
		let t = N.current, a = Br(o, y), s = Vr(o, y);
		if (j || (e.key === "Home" && (Or(e), N.current = a, A()), e.key === "End" && (Or(e), N.current = s, A())), S > 1) {
			let t = T || Array.from({ length: o.current.length }, () => ({
				width: 1,
				height: 1
			})), n = Wr(t, S, E), r = n.findIndex((e) => e != null && !qr(o, e, y)), i = n.reduce((e, t, n) => t != null && !qr(o, t, y) ? n : e, -1), c = n[Ur({ current: n.map((e) => e == null ? null : o.current[e]) }, {
				event: e,
				orientation: b,
				loop: f,
				rtl: m,
				cols: S,
				disabledIndices: Kr([...(typeof y == "function" ? null : y) || o.current.map((e, t) => qr(o, t, y) ? t : void 0), void 0], n),
				minIndex: r,
				maxIndex: i,
				prevIndex: Gr(N.current > s ? a : N.current, t, n, S, e.key === fi ? "bl" : e.key === (m ? li : ui) ? "tr" : "tl"),
				stopEvent: !0
			})];
			if (c != null && (N.current = c, A()), b === "both") return;
		}
		if (Oa(e.key, b)) {
			if (Or(e), n && !h && gr(e.currentTarget.ownerDocument) === e.currentTarget) {
				N.current = ka(e.key, b, m) ? a : s, A();
				return;
			}
			ka(e.key, b, m) ? f ? N.current = t >= s ? d && t !== o.current.length ? -1 : a : Hr(o, {
				startingIndex: t,
				disabledIndices: y
			}) : N.current = Math.min(s, Hr(o, {
				startingIndex: t,
				disabledIndices: y
			})) : f ? N.current = t <= a ? d && t !== -1 ? o.current.length : s : Hr(o, {
				startingIndex: t,
				decrement: !0,
				disabledIndices: y
			}) : N.current = Math.max(a, Hr(o, {
				startingIndex: t,
				decrement: !0,
				disabledIndices: y
			})), zr(o, N.current) && (N.current = -1), A();
		}
	}), K = Y.useMemo(() => h && n && ae && { "aria-activedescendant": ne || te }, [
		h,
		n,
		ae,
		ne,
		te
	]), ce = Y.useMemo(() => ({
		"aria-orientation": b === "both" ? void 0 : b,
		...j ? {} : K,
		onKeyDown: se,
		onPointerMove() {
			F.current = !0;
		}
	}), [
		K,
		se,
		b,
		j
	]), le = Y.useMemo(() => {
		function e(e) {
			g === "auto" && Ar(e.nativeEvent) && (M.current = !0);
		}
		function t(e) {
			M.current = g, g === "auto" && jr(e.nativeEvent) && (M.current = !0);
		}
		return {
			...K,
			onKeyDown(e) {
				F.current = !1;
				let t = e.key.startsWith("Arrow"), i = ["Home", "End"].includes(e.key), a = t || i, s = Aa(e.key, b, m), c = ja(e.key, b, m, S), l = Aa(e.key, G(), m), d = Oa(e.key, b), f = (p ? l : d) || e.key === "Enter" || e.key.trim() === "";
				if (h && n) {
					let t = k?.nodesRef.current.find((e) => e.parentId == null), n = k && t ? Er(k.nodesRef.current, t.id) : null;
					if (a && n && w) {
						let t = new KeyboardEvent("keydown", {
							key: e.key,
							bubbles: !0
						});
						if (s || c) {
							let r = n.context?.elements.domReference === e.currentTarget, i = c && !r ? n.context?.elements.domReference : s ? o.current.find((e) => e?.id === te) : null;
							i && (Or(e), i.dispatchEvent(t), re(void 0));
						}
						if ((d || i) && n.context && n.context.open && n.parentId && e.currentTarget !== n.context.elements.domReference) {
							var g;
							Or(e), (g = n.context.elements.domReference) == null || g.dispatchEvent(t);
							return;
						}
					}
					return se(e);
				}
				if (n || v || !t) {
					if (f) {
						let t = Oa(e.key, G());
						P.current = p && t ? null : e.key;
					}
					p ? l && (Or(e), n ? (N.current = Br(o, V.current), A()) : r(!0, e.nativeEvent, "list-navigation")) : d && (u != null && (N.current = u), Or(e), !n && v ? r(!0, e.nativeEvent, "list-navigation") : se(e), n && A());
				}
			},
			onFocus() {
				n && !h && (N.current = -1, A());
			},
			onPointerDown: t,
			onPointerEnter: t,
			onMouseDown: e,
			onClick: e
		};
	}, [
		te,
		K,
		S,
		se,
		V,
		g,
		o,
		p,
		A,
		r,
		n,
		v,
		b,
		G,
		m,
		u,
		k,
		h,
		w
	]);
	return Y.useMemo(() => l ? {
		reference: le,
		floating: ce,
		item: oe
	} : {}, [
		l,
		le,
		ce,
		oe
	]);
}
var Na = /*#__PURE__*/ new Map([
	["select", "listbox"],
	["combobox", "listbox"],
	["label", !1]
]);
function Pa(e, t) {
	t === void 0 && (t = {});
	let { open: n, elements: r, floatingId: i } = e, { enabled: a = !0, role: o = "dialog" } = t, s = bi(), c = r.domReference?.id || s, l = Y.useMemo(() => wr(r.floating)?.id || i, [r.floating, i]), u = Na.get(o) ?? o, d = wi() != null, f = Y.useMemo(() => u === "tooltip" || o === "label" ? { ["aria-" + (o === "label" ? "labelledby" : "describedby")]: n ? l : void 0 } : {
		"aria-expanded": n ? "true" : "false",
		"aria-haspopup": u === "alertdialog" ? "dialog" : u,
		"aria-controls": n ? l : void 0,
		...u === "listbox" && { role: "combobox" },
		...u === "menu" && { id: c },
		...u === "menu" && d && { role: "menuitem" },
		...o === "select" && { "aria-autocomplete": "none" },
		...o === "combobox" && { "aria-autocomplete": "list" }
	}, [
		u,
		l,
		d,
		n,
		c,
		o
	]), p = Y.useMemo(() => {
		let e = {
			id: l,
			...u && { role: u }
		};
		return u === "tooltip" || o === "label" ? e : {
			...e,
			...u === "menu" && { "aria-labelledby": c }
		};
	}, [
		u,
		l,
		c,
		o
	]), m = Y.useCallback((e) => {
		let { active: t, selected: n } = e, r = {
			role: "option",
			...t && { id: l + "-fui-option" }
		};
		switch (o) {
			case "select":
			case "combobox": return {
				...r,
				"aria-selected": n
			};
		}
		return {};
	}, [l, o]);
	return Y.useMemo(() => a ? {
		reference: f,
		floating: p,
		item: m
	} : {}, [
		a,
		f,
		p,
		m
	]);
}
function Fa(e, t) {
	let { open: n, dataRef: r } = e, { listRef: i, activeIndex: a, onMatch: o, onTypingChange: s, enabled: c = !0, findMatch: l = null, resetMs: u = 750, ignoreKeys: d = [], selectedIndex: f = null } = t, p = Y.useRef(-1), m = Y.useRef(""), h = Y.useRef(f ?? a ?? -1), g = Y.useRef(null), _ = Lr(o), v = Lr(s), y = Fr(l), b = Fr(d);
	Nr(() => {
		n && (Ai(p), g.current = null, m.current = "");
	}, [n]), Nr(() => {
		n && m.current === "" && (h.current = f ?? a ?? -1);
	}, [
		n,
		f,
		a
	]);
	let x = Lr((e) => {
		e ? r.current.typing || (r.current.typing = e, v(e)) : r.current.typing && (r.current.typing = e, v(e));
	}), S = Lr((e) => {
		function t(e, t, n) {
			let r = y.current ? y.current(t, n) : t.find((e) => e?.toLocaleLowerCase().indexOf(n.toLocaleLowerCase()) === 0);
			return r ? e.indexOf(r) : -1;
		}
		let r = i.current;
		if (m.current.length > 0 && m.current[0] !== " " && (t(r, r, m.current) === -1 ? x(!1) : e.key === " " && Or(e)), r == null || b.current.includes(e.key) || e.key.length !== 1 || e.ctrlKey || e.metaKey || e.altKey) return;
		n && e.key !== " " && (Or(e), x(!0)), r.every((e) => !e || e[0]?.toLocaleLowerCase() !== e[1]?.toLocaleLowerCase()) && m.current === e.key && (m.current = "", h.current = g.current), m.current += e.key, Ai(p), p.current = window.setTimeout(() => {
			m.current = "", h.current = g.current, x(!1);
		}, u);
		let a = h.current, o = t(r, [...r.slice((a || 0) + 1), ...r.slice(0, (a || 0) + 1)], m.current);
		o === -1 ? e.key !== " " && (m.current = "", x(!1)) : (_(o), g.current = o);
	}), C = Y.useMemo(() => ({ onKeyDown: S }), [S]), w = Y.useMemo(() => ({
		onKeyDown: S,
		onKeyUp(e) {
			e.key === " " && x(!1);
		}
	}), [S, x]);
	return Y.useMemo(() => c ? {
		reference: C,
		floating: w
	} : {}, [
		c,
		C,
		w
	]);
}
function Ia(e, t, n) {
	return n === void 0 && (n = !0), e.filter((e) => e.parentId === t && (!n || e.context?.open)).flatMap((t) => [t, ...Ia(e, t.id, n)]);
}
function La(e, t) {
	let [n, r] = e, i = !1, a = t.length;
	for (let e = 0, o = a - 1; e < a; o = e++) {
		let [a, s] = t[e] || [0, 0], [c, l] = t[o] || [0, 0];
		s >= r != l >= r && n <= (c - a) * (r - s) / (l - s) + a && (i = !i);
	}
	return i;
}
function Ra(e, t) {
	return e[0] >= t.x && e[0] <= t.x + t.width && e[1] >= t.y && e[1] <= t.y + t.height;
}
function za(e) {
	e === void 0 && (e = {});
	let { buffer: t = .5, blockPointerEvents: n = !1, requireIntent: r = !0 } = e, i = { current: -1 }, a = !1, o = null, s = null, c = typeof performance < "u" ? performance.now() : 0;
	function l(e, t) {
		let n = performance.now(), r = n - c;
		if (o === null || s === null || r === 0) return o = e, s = t, c = n, null;
		let i = e - o, a = t - s, l = Math.sqrt(i * i + a * a) / r;
		return o = e, s = t, c = n, l;
	}
	let u = (e) => {
		let { x: n, y: o, placement: s, elements: c, onClose: u, nodeId: d, tree: f } = e;
		return function(e) {
			function p() {
				Ai(i), u();
			}
			if (Ai(i), !c.domReference || !c.floating || s == null || n == null || o == null) return;
			let { clientX: m, clientY: h } = e, g = [m, h], _ = Ri(e), v = e.type === "mouseleave", y = Li(c.floating, _), b = Li(c.domReference, _), x = c.domReference.getBoundingClientRect(), S = c.floating.getBoundingClientRect(), C = s.split("-")[0], w = n > S.right - S.width / 2, T = o > S.bottom - S.height / 2, E = Ra(g, x), D = S.width > x.width, O = S.height > x.height, k = (D ? x : S).left, A = (D ? x : S).right, j = (O ? x : S).top, M = (O ? x : S).bottom;
			if (y && (a = !0, !v)) return;
			if (b && (a = !1), b && !v) {
				a = !0;
				return;
			}
			if (v && Te(e.relatedTarget) && Li(c.floating, e.relatedTarget) || f && Ia(f.nodesRef.current, d).length) return;
			if (C === "top" && o >= x.bottom - 1 || C === "bottom" && o <= x.top + 1 || C === "left" && n >= x.right - 1 || C === "right" && n <= x.left + 1) return p();
			let N = [];
			switch (C) {
				case "top":
					N = [
						[k, x.top + 1],
						[k, S.bottom - 1],
						[A, S.bottom - 1],
						[A, x.top + 1]
					];
					break;
				case "bottom":
					N = [
						[k, S.top + 1],
						[k, x.bottom - 1],
						[A, x.bottom - 1],
						[A, S.top + 1]
					];
					break;
				case "left":
					N = [
						[S.right - 1, M],
						[S.right - 1, j],
						[x.left + 1, j],
						[x.left + 1, M]
					];
					break;
				case "right": N = [
					[x.right - 1, M],
					[x.right - 1, j],
					[S.left + 1, j],
					[S.left + 1, M]
				];
			}
			function P(e) {
				let [n, r] = e;
				switch (C) {
					case "top": return [
						[D ? n + t / 2 : w ? n + t * 4 : n - t * 4, r + t + 1],
						[D ? n - t / 2 : w ? n + t * 4 : n - t * 4, r + t + 1],
						[S.left, w || D ? S.bottom - t : S.top],
						[S.right, w ? D ? S.bottom - t : S.top : S.bottom - t]
					];
					case "bottom": return [
						[D ? n + t / 2 : w ? n + t * 4 : n - t * 4, r - t],
						[D ? n - t / 2 : w ? n + t * 4 : n - t * 4, r - t],
						[S.left, w || D ? S.top + t : S.bottom],
						[S.right, w ? D ? S.top + t : S.bottom : S.top + t]
					];
					case "left": {
						let e = [n + t + 1, O ? r + t / 2 : T ? r + t * 4 : r - t * 4], i = [n + t + 1, O ? r - t / 2 : T ? r + t * 4 : r - t * 4];
						return [
							[T || O ? S.right - t : S.left, S.top],
							[T ? O ? S.right - t : S.left : S.right - t, S.bottom],
							e,
							i
						];
					}
					case "right": return [
						[n - t, O ? r + t / 2 : T ? r + t * 4 : r - t * 4],
						[n - t, O ? r - t / 2 : T ? r + t * 4 : r - t * 4],
						[T || O ? S.left + t : S.right, S.top],
						[T ? O ? S.left + t : S.right : S.left + t, S.bottom]
					];
				}
			}
			if (!La([m, h], N)) {
				if (a && !E) return p();
				if (!v && r) {
					let t = l(e.clientX, e.clientY);
					if (t !== null && t < .1) return p();
				}
				La([m, h], P([n, o])) ? !a && r && (i.current = window.setTimeout(p, 40)) : p();
			}
		};
	};
	return u.__options = { blockPointerEvents: n }, u;
}
// ../../../clicky-ui/packages/ui/src/overlay/dropdownMenuContext.ts
var Ba = "min-w-[8rem] overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-md", Va = je({
	padding: 8,
	apply({ availableHeight: e, elements: t }) {
		t.floating.style.maxHeight = `${Math.max(0, e)}px`;
	}
}), Ha = (0, Y.createContext)({
	getItemProps: () => ({}),
	activeIndex: null,
	setActiveIndex: () => {},
	setHasFocusInside: () => {},
	isOpen: !1
}), Ua = {
	raised: 30,
	chatFab: 60,
	chatWindow: 61,
	popover: 9e3,
	modal: 1e4,
	modalStep: 10,
	popoverOverModalOffset: 5,
	tour: 5e4,
	tourCardOffset: 10,
	toast: 1e5
}, Wa = [], Ga = /* @__PURE__ */ new Set(), Ka = [], qa = null, Ja = 0;
function Ya() {
	for (let e of Ga) e();
}
function Xa(e) {
	return Ga.add(e), () => {
		Ga.delete(e);
	};
}
function Za() {
	return `${Ja}:` + Wa.join("\0");
}
function Qa(e) {
	if (e.defaultPrevented || e.key !== "Escape" || e.isComposing || e.keyCode === 229) return;
	let t = Ka.findLast((e) => !e.dismissed);
	t && (e.preventDefault(), e.stopPropagation(), e.stopImmediatePropagation(), t.dismissed = !0, t.onEscape());
}
function $a() {
	qa || typeof document > "u" || (qa = document, qa.addEventListener("keydown", Qa));
}
function eo(e) {
	let t = Ka.findIndex((t) => t.id === e);
	t !== -1 && Ka.splice(t, 1), Ka.length === 0 && qa && (qa.removeEventListener("keydown", Qa), qa = null);
}
function to(e) {
	let t = (0, Y.useId)();
	(0, Y.useSyncExternalStore)(Xa, Za, Za), (0, Y.useEffect)(() => {
		if (e) return Wa.push(t), Ya(), () => {
			let e = Wa.indexOf(t);
			e !== -1 && Wa.splice(e, 1), Ya();
		};
	}, [e, t]);
	let n = Wa.indexOf(t);
	return {
		isTop: n !== -1 && n === Wa.length - 1,
		depth: n === -1 ? 0 : n
	};
}
function no(e, t, n = !0) {
	let r = (0, Y.useId)(), i = (0, Y.useRef)(t);
	i.current = t;
	let a = (0, Y.useRef)(null);
	(0, Y.useEffect)(() => {
		a.current && (a.current.dismissed = !1);
	}), (0, Y.useEffect)(() => {
		if (!e || !n) return;
		$a();
		let t = {
			id: r,
			dismissed: !1,
			onEscape: () => i.current()
		};
		return a.current = t, Ka.push(t), () => {
			a.current = null, eo(r);
		};
	}, [
		n,
		r,
		e
	]);
}
function ro() {
	(0, Y.useSyncExternalStore)(Xa, Za, Za);
	let e = Wa.length === 0 ? Ua.popover : Ua.modal + (Wa.length - 1) * Ua.modalStep + Ua.popoverOverModalOffset;
	return Ja === 0 ? e : Math.max(e, Ua.tour + Ua.tourCardOffset + Ua.popoverOverModalOffset);
}
// ../../../clicky-ui/packages/ui/src/overlay/DropdownMenuSubmenu.tsx
var io = "flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs text-popover-foreground hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50", ao = "select-none px-3 pb-1 pt-2 text-[10px] font-medium uppercase tracking-wide text-muted-foreground";
function oo({ icon: e, iconColor: t, iconClassName: n }) {
	return e ? /* @__PURE__ */ (0, J.jsx)("span", {
		className: q("inline-flex shrink-0 text-base", n),
		style: t ? { color: t } : void 0,
		children: /* @__PURE__ */ (0, J.jsx)(c, { ...typeof e == "string" ? { name: e } : { icon: e } })
	}) : null;
}
function so({ item: e }) {
	let t = (0, Y.useContext)(Ha), n = Ti(), r = ai(), i = r.index === t.activeIndex, a = () => {
		e.disabled || (e.onSelect(), n?.events.emit("click"));
	};
	return /* @__PURE__ */ (0, J.jsxs)("button", {
		ref: r.ref,
		type: "button",
		role: "menuitem",
		tabIndex: i ? 0 : -1,
		disabled: e.disabled,
		title: e.title,
		className: io,
		...t.getItemProps({
			onClick: a,
			onFocus: () => t.setHasFocusInside(!0),
			onKeyDown: (e) => {
				(e.key === "Enter" || e.key === " ") && (e.preventDefault(), a());
			}
		}),
		children: [/* @__PURE__ */ (0, J.jsx)(oo, {
			icon: e.icon,
			iconColor: e.iconColor,
			iconClassName: e.iconClassName
		}), e.label]
	});
}
function co({ item: e }) {
	let t = (0, Y.useContext)(Ha), n = Ti(), r = Ei(), i = wi(), a = ai(), [o, s] = (0, Y.useState)(!1), [l, u] = (0, Y.useState)(null), [d, f] = (0, Y.useState)(!1), p = (0, Y.useRef)([]), m = ro(), { refs: h, floatingStyles: g, context: _ } = Ca({
		...r ? { nodeId: r } : {},
		open: o,
		onOpenChange: s,
		placement: "right-start",
		middleware: [
			ye({
				mainAxis: 0,
				alignmentAxis: -4
			}),
			Oe({ padding: 8 }),
			ke({
				padding: 8,
				crossAxis: !0
			}),
			Va
		],
		whileElementsMounted: Ee
	}), { getReferenceProps: v, getFloatingProps: y, getItemProps: b } = Ta([
		Pi(_, {
			delay: { open: 75 },
			handleClose: za()
		}),
		_a(_, { toggle: !1 }),
		Pa(_, { role: "menu" }),
		xa(_, {
			escapeKey: !1,
			bubbles: !0
		}),
		Ma(_, {
			listRef: p,
			activeIndex: l,
			nested: !0,
			onNavigate: u,
			loop: !0
		})
	]);
	no(o, () => _.onOpenChange(!1)), (0, Y.useEffect)(() => {
		if (!n) return;
		let e = () => _.onOpenChange(!1), t = (e) => {
			e.nodeId !== r && e.parentId === i && _.onOpenChange(!1);
		};
		return n.events.on("click", e), n.events.on("menuopen", t), () => {
			n.events.off("click", e), n.events.off("menuopen", t);
		};
	}, [
		n,
		r,
		i,
		_.onOpenChange
	]), (0, Y.useEffect)(() => {
		o && n && n.events.emit("menuopen", {
			parentId: i,
			nodeId: r
		});
	}, [
		o,
		n,
		r,
		i
	]);
	let x = a.index === t.activeIndex, S = ti([h.setReference, a.ref]), C = typeof e.label == "string" ? e.label : void 0;
	return /* @__PURE__ */ (0, J.jsxs)(Di, {
		id: r,
		children: [/* @__PURE__ */ (0, J.jsxs)("button", {
			ref: S,
			type: "button",
			role: "menuitem",
			tabIndex: x ? 0 : -1,
			disabled: e.disabled,
			title: e.title,
			"aria-haspopup": "menu",
			"aria-expanded": o,
			"data-focus-inside": d ? "" : void 0,
			className: io,
			...t.getItemProps(v({ onFocus: () => {
				f(!1), t.setHasFocusInside(!0);
			} })),
			children: [
				/* @__PURE__ */ (0, J.jsx)(oo, {
					icon: e.icon,
					iconColor: e.iconColor,
					iconClassName: e.iconClassName
				}),
				/* @__PURE__ */ (0, J.jsx)("span", {
					className: "min-w-0 flex-1 truncate",
					children: e.label
				}),
				/* @__PURE__ */ (0, J.jsx)(c, {
					icon: he,
					className: "ml-auto shrink-0 opacity-60"
				})
			]
		}), /* @__PURE__ */ (0, J.jsx)(Ha.Provider, {
			value: {
				getItemProps: b,
				activeIndex: l,
				setActiveIndex: u,
				setHasFocusInside: f,
				isOpen: o
			},
			children: /* @__PURE__ */ (0, J.jsx)(ii, {
				elementsRef: p,
				children: o && /* @__PURE__ */ (0, J.jsx)(na, { children: /* @__PURE__ */ (0, J.jsx)(pa, {
					context: _,
					modal: !1,
					initialFocus: -1,
					returnFocus: !1,
					children: /* @__PURE__ */ (0, J.jsxs)("div", {
						ref: h.setFloating,
						role: "menu",
						style: {
							...g,
							zIndex: m
						},
						className: "min-w-[8rem] overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-md",
						...y(),
						"aria-label": C,
						...C ? { "aria-labelledby": void 0 } : {},
						children: [e.header != null && /* @__PURE__ */ (0, J.jsx)("div", {
							className: "border-b border-border px-3 py-1.5",
							children: e.header
						}), /* @__PURE__ */ (0, J.jsx)(lo, { items: e.children ?? [] })]
					})
				}) })
			})
		})]
	});
}
function lo({ items: e }) {
	return /* @__PURE__ */ (0, J.jsx)(J.Fragment, { children: e.map((t, n) => {
		let r = t.group != null && t.group !== e[n - 1]?.group, i = !!t.children && t.children.length > 0;
		return /* @__PURE__ */ (0, J.jsxs)(Y.Fragment, { children: [r && /* @__PURE__ */ (0, J.jsx)("div", {
			role: "presentation",
			className: ao,
			children: t.group
		}), i ? /* @__PURE__ */ (0, J.jsx)(co, { item: t }) : /* @__PURE__ */ (0, J.jsx)(so, { item: t })] }, n);
	}) });
}
// ../../../clicky-ui/packages/ui/src/overlay/DropdownMenu.tsx
function uo(e) {
	return /* @__PURE__ */ (0, J.jsx)(Oi, { children: /* @__PURE__ */ (0, J.jsx)(fo, { ...e }) });
}
function fo({ role: e = "menu", label: t, icon: n, hideChevron: r = !1, variant: i = "outline", size: a = "sm", trigger: o, items: s, children: l, header: u, footer: d, align: f = "right", onOpenChange: p, title: m, menuLabel: h, className: g, menuClassName: _ }) {
	let [v, y] = (0, Y.useState)(!1), [b, x] = (0, Y.useState)(null), S = ro(), C = Ti(), w = Ei(), T = (0, Y.useRef)([]), E = (e) => {
		e !== v && (y(e), p?.(e));
	}, { refs: D, floatingStyles: O, context: k } = Ca({
		...w ? { nodeId: w } : {},
		open: v,
		onOpenChange: E,
		placement: f === "right" ? "bottom-end" : "bottom-start",
		whileElementsMounted: Ee,
		middleware: [
			ye(4),
			Oe({ padding: 8 }),
			ke({ padding: 8 }),
			Va
		]
	}), A = _a(k), j = xa(k, { escapeKey: !1 }), M = Pa(k, { role: e }), N = Ma(k, {
		listRef: T,
		activeIndex: b,
		onNavigate: x,
		loop: !0
	}), { getReferenceProps: P, getFloatingProps: F, getItemProps: I } = Ta([
		A,
		j,
		M,
		...e === "menu" ? [N] : []
	]), L = () => {
		E(!1), D.domReference.current instanceof HTMLElement && D.domReference.current.focus();
	};
	no(v, L);
	let R = (0, Y.useRef)(L);
	return (0, Y.useLayoutEffect)(() => {
		R.current = L;
	}), (0, Y.useEffect)(() => {
		if (!C) return;
		let e = () => R.current();
		return C.events.on("click", e), () => C.events.off("click", e);
	}, [C]), /* @__PURE__ */ (0, J.jsx)(Di, {
		id: w,
		children: /* @__PURE__ */ (0, J.jsxs)("div", {
			className: q("relative inline-flex", g),
			children: [o ? /* @__PURE__ */ (0, J.jsx)(_e, {
				ref: D.setReference,
				className: "inline-flex",
				"aria-haspopup": e,
				"aria-expanded": v,
				...P(),
				children: o
			}) : /* @__PURE__ */ (0, J.jsxs)(fe, {
				ref: D.setReference,
				type: "button",
				variant: i,
				size: a,
				title: m,
				"aria-haspopup": e,
				"aria-expanded": v,
				...P(),
				children: [
					n && /* @__PURE__ */ (0, J.jsx)(c, { ...typeof n == "string" ? { name: n } : { icon: n } }),
					t,
					!r && /* @__PURE__ */ (0, J.jsx)(c, { icon: pe })
				]
			}), v && /* @__PURE__ */ (0, J.jsx)(na, { children: /* @__PURE__ */ (0, J.jsx)(pa, {
				context: k,
				modal: !1,
				closeOnFocusOut: !1,
				children: /* @__PURE__ */ (0, J.jsxs)("div", {
					ref: D.setFloating,
					...F(),
					role: e,
					"aria-modal": e !== "dialog" && void 0,
					"aria-label": h,
					...h ? { "aria-labelledby": void 0 } : {},
					style: {
						...O,
						zIndex: S
					},
					className: q("min-w-[8rem] overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-md", _),
					children: [
						u != null && /* @__PURE__ */ (0, J.jsx)("div", {
							className: "border-b border-border px-3 py-1.5",
							children: u
						}),
						l ? l(L) : /* @__PURE__ */ (0, J.jsx)(Ha.Provider, {
							value: {
								getItemProps: I,
								activeIndex: b,
								setActiveIndex: x,
								setHasFocusInside: () => {},
								isOpen: v
							},
							children: /* @__PURE__ */ (0, J.jsx)(ii, {
								elementsRef: T,
								children: /* @__PURE__ */ (0, J.jsx)(lo, { items: s ?? [] })
							})
						}),
						d != null && /* @__PURE__ */ (0, J.jsx)("div", {
							className: "border-t border-border px-3 py-1.5",
							children: d
						})
					]
				})
			}) })]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/InputField.tsx
var po = (0, Y.forwardRef)(function({ as: e = "input", value: t, onChange: n, prefix: r, suffix: i, shortcut: a, shortcutKey: o, hotkey: s, onShortcut: c, invalid: l, className: u, inputClassName: d, shortcutClassName: f, type: p = "text", clearable: m, onClear: h, clearLabel: g, ..._ }, v) {
	let b = (0, Y.useRef)(null);
	(0, Y.useImperativeHandle)(v, () => b.current, []);
	let x = o ?? (a ? a.slice(-1).toLowerCase() : void 0);
	return Ge(s ?? (x ? `mod+${x}` : null), () => {
		c?.(), b.current?.focus();
	}, { enabled: !!c }), /* @__PURE__ */ (0, J.jsxs)("div", {
		"data-clearable-field": !0,
		"data-jsf-control": !0,
		className: q("flex w-full gap-density-2 rounded-md border border-border bg-background px-density-2", "focus-within:ring-2 focus-within:ring-ring", e === "textarea" ? "min-h-control-h items-start py-density-1" : "h-control-h items-center", l && "border-destructive focus-within:ring-destructive", "has-[:disabled]:cursor-not-allowed has-[:disabled]:opacity-50", u),
		children: [
			r && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "inline-flex shrink-0 items-center",
				children: r
			}),
			e === "textarea" ? /* @__PURE__ */ (0, J.jsx)("textarea", {
				ref: b,
				value: t,
				onChange: (e) => n?.(e.target.value, e),
				className: q("min-w-0 flex-1 resize-y border-none bg-transparent text-sm text-foreground outline-none", "placeholder:text-placeholder", "disabled:cursor-not-allowed", d),
				...l ? { "aria-invalid": !0 } : {},
				..._
			}) : /* @__PURE__ */ (0, J.jsx)("input", {
				ref: b,
				type: p,
				value: t,
				onChange: (e) => n?.(e.target.value, e),
				className: q("min-w-0 flex-1 border-none bg-transparent text-sm text-foreground outline-none", "placeholder:text-placeholder", "disabled:cursor-not-allowed", p === "search" && "[&::-webkit-search-cancel-button]:appearance-none", d),
				...l ? { "aria-invalid": !0 } : {},
				..._
			}),
			i && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "inline-flex shrink-0 items-center",
				children: i
			}),
			m && t !== void 0 && t !== "" && !_.disabled && !_.readOnly && /* @__PURE__ */ (0, J.jsx)(y, {
				...h ? { onClear: h } : {},
				...g ? { clearLabel: g } : {}
			}),
			a && /* @__PURE__ */ (0, J.jsx)("kbd", {
				className: q("shrink-0 rounded border border-border px-1 font-mono text-[10px] text-muted-foreground", f),
				children: a
			})
		]
	});
});
// ../../../clicky-ui/packages/ui/src/lib/combobox-values.ts
function mo(e, t) {
	return t > 0 ? { paddingLeft: `calc(${t}px + 0.75rem)` } : typeof e == "string" ? { paddingLeft: `${Math.min(e.length * .48 + .75, 7)}rem` } : { paddingLeft: "1.75rem" };
}
function ho(e, t) {
	let n = new Set(e.map((e) => e.value)), r = t.filter((e) => !n.has(e)).map((e) => ({
		value: e,
		label: e
	}));
	return r.length === 0 ? e : [...r, ...e];
}
function go(e, t) {
	let n = t.map((t) => {
		let n = e.find((e) => e.value === t);
		return n?.selectedLabel ?? n?.label ?? t;
	});
	return n.length <= 2 ? n.join(", ") : `${n.length} selected`;
}
function _o(e, t) {
	let n = /* @__PURE__ */ new Set([
		...t,
		"\n",
		"\r"
	]), r = [], i = "";
	for (let t of e) n.has(t) ? (r.push(i), i = "") : i += t;
	return r.push(i), r;
}
function vo(e) {
	let { allowCustomValue: t, query: n, multiple: r, tristate: i, onNew: a, choices: o, selectedValues: s } = e;
	return !t || !n || s.includes(n) || !i && !r && !a || o.some((e) => e.value === n) ? null : a ? a(n) : {
		value: n,
		label: `Add "${n}"`
	};
}
// ../../../clicky-ui/packages/ui/src/components/ComboboxActions.tsx
function yo({ disabled: e, loading: t, onClear: n, clearLabel: r = "Clear", onToggle: i, showClear: a, showToggle: o }) {
	return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [a && /* @__PURE__ */ (0, J.jsx)("button", {
		type: "button",
		"aria-label": r,
		title: r,
		onClick: n,
		className: q("absolute flex h-full items-center px-1 text-muted-foreground hover:text-foreground", o ? "right-7" : "right-1"),
		children: /* @__PURE__ */ (0, J.jsx)(c, {
			icon: j,
			className: "text-xs"
		})
	}), t ? /* @__PURE__ */ (0, J.jsx)("span", {
		className: "pointer-events-none absolute right-2 top-1/2 -translate-y-1/2",
		children: /* @__PURE__ */ (0, J.jsxs)("svg", {
			className: "h-3.5 w-3.5 animate-spin text-muted-foreground",
			viewBox: "0 0 16 16",
			fill: "none",
			children: [/* @__PURE__ */ (0, J.jsx)("circle", {
				cx: "8",
				cy: "8",
				r: "6.5",
				stroke: "currentColor",
				strokeWidth: "2",
				opacity: "0.25"
			}), /* @__PURE__ */ (0, J.jsx)("path", {
				d: "M14.5 8a6.5 6.5 0 0 0-6.5-6.5",
				stroke: "currentColor",
				strokeWidth: "2",
				strokeLinecap: "round"
			})]
		})
	}) : o && /* @__PURE__ */ (0, J.jsx)("button", {
		type: "button",
		tabIndex: -1,
		disabled: e,
		"aria-label": "Toggle options",
		onClick: i,
		className: "pointer-events-auto absolute right-0 flex h-full items-center px-2 text-muted-foreground",
		children: /* @__PURE__ */ (0, J.jsx)(c, {
			icon: pe,
			className: "text-xs"
		})
	})] });
}
// ../../../clicky-ui/packages/ui/src/lib/format.ts
var bo = "en-US";
function xo(e, t = 1) {
	return new Intl.NumberFormat(bo, { maximumFractionDigits: t }).format(Number.isFinite(e) ? e : 0);
}
function So(e) {
	return e.endsWith(".0") ? e.slice(0, -2) : e;
}
var Co = [
	"B",
	"KB",
	"MB",
	"GB",
	"TB",
	"PB"
];
function wo(e, t) {
	let n = t?.suffix ?? "";
	if (e === void 0 || !Number.isFinite(e) || e <= 0) return `0 B${n}`;
	let r = t?.base ?? 1024, i = e, a = 0;
	for (; i >= r && a < Co.length - 1;) i /= r, a++;
	return `${i >= 10 || a === 0 ? i.toFixed(0) : So(i.toFixed(1))} ${Co[a]}${n}`;
}
function To(e) {
	return wo(e, { suffix: "/s" });
}
var Eo = [
	"",
	"K",
	"M",
	"B",
	"T"
];
function Do(e) {
	if (e === void 0 || !Number.isFinite(e)) return "0";
	let t = e < 0 ? "-" : "", n = Math.abs(e);
	if (n < 1e3) return `${t}${xo(n)}`;
	let r = 0;
	for (; n >= 1e3 && r < Eo.length - 1;) n /= 1e3, r++;
	return `${t}${n >= 10 ? n.toFixed(0) : So(n.toFixed(1))}${Eo[r]}`;
}
function Oo(e, t) {
	if (e === void 0 || !Number.isFinite(e)) return "0 ms";
	let n = t?.from === "s" ? e * 1e3 : e, r = Math.abs(n);
	return r < 1e3 ? `${xo(n, 0)} ms` : r < 6e4 ? `${xo(n / 1e3)} s` : r < 36e5 ? `${xo(n / 6e4)} min` : `${xo(n / 36e5)} h`;
}
function ko(e, t, n) {
	if (!e) return;
	let r = Ao("startedAt", e), i = t ? Ao("finishedAt", t) : n;
	return Math.max(0, i - r);
}
function Ao(e, t) {
	let n = Date.parse(t);
	if (Number.isNaN(n)) throw Error(`elapsedMs: malformed ${e} ${JSON.stringify(t)}`);
	return n;
}
var jo = {
	none: (e) => Do(e),
	short: (e) => Do(e),
	percent: (e) => `${xo(e)}%`,
	percentunit: (e) => `${xo(e * 100)}%`,
	"%": (e) => `${xo(e)}%`,
	bytes: (e) => wo(e, { base: 1024 }),
	decbytes: (e) => wo(e, { base: 1e3 }),
	Bps: (e) => To(e),
	binBps: (e) => To(e),
	ms: (e) => Oo(e, { from: "ms" }),
	s: (e) => Oo(e, { from: "s" })
};
function Mo(e, t) {
	let n = Number.isFinite(e) ? e : 0;
	if (!t) return xo(n);
	let r = jo[t];
	return r ? r(n) : `${xo(n)}${t}`;
}
// ../../../clicky-ui/packages/ui/src/data/FilterPill.tsx
function No(e) {
	switch (e) {
		case "active": return "bg-primary/10 border-primary/40 text-primary font-medium";
		case "include": return "bg-green-500/10 border-green-500/50 text-green-700 dark:text-green-400 font-medium";
		case "exclude": return "bg-red-500/10 border-red-500/50 text-red-700 dark:text-red-400 font-medium";
		default: return "border-border text-muted-foreground hover:bg-accent";
	}
}
var Po = {
	exclude: "translate-x-0",
	neutral: "translate-x-4",
	include: "translate-x-8"
};
function Fo({ mode: e, onChange: t, ariaLabel: n, interactive: r = !0 }) {
	let i = e === "exclude" ? "exclude" : e === "include" ? "include" : "neutral", a = i === "exclude" ? "bg-red-500/80" : i === "include" ? "bg-green-500/80" : "bg-muted", o = (e, n) => /* @__PURE__ */ (0, J.jsx)("span", {
		"aria-hidden": !0,
		"data-tristate-region": n,
		onClick: (n) => {
			n.preventDefault(), n.stopPropagation(), t(e);
		},
		className: "relative z-10 h-full flex-1 cursor-pointer"
	}), s = /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [
		o(e === "exclude" ? "neutral" : "exclude", "exclude"),
		o(Ro(e), "neutral"),
		o(e === "include" ? "neutral" : "include", "include"),
		/* @__PURE__ */ (0, J.jsx)("span", {
			"aria-hidden": !0,
			className: q("pointer-events-none absolute left-0.5 top-0.5 z-20 flex size-4 items-center justify-center rounded-full bg-white shadow-sm transition-transform duration-150", Po[i]),
			children: i !== "neutral" && /* @__PURE__ */ (0, J.jsx)(c, {
				icon: i === "exclude" ? j : bt,
				className: q("text-[10px]", i === "exclude" && "text-red-600", i === "include" && "text-green-600")
			})
		})
	] }), l = q("relative inline-flex h-5 w-[52px] shrink-0 overflow-hidden rounded-full transition-colors duration-200", a);
	return r ? /* @__PURE__ */ (0, J.jsx)("button", {
		type: "button",
		"aria-label": n,
		title: n,
		onClick: (n) => {
			n.preventDefault(), n.stopPropagation(), t(Ro(e));
		},
		className: l,
		children: s
	}) : /* @__PURE__ */ (0, J.jsx)("span", {
		"aria-hidden": !0,
		className: l,
		children: s
	});
}
function Io({ mode: e = "neutral", label: t, count: n, icon: r, badge: i, onModeChange: a, onClick: o, title: s, togglePosition: l = "left", interactive: u = !0, className: d }) {
	let f = !!a, p = r && /* @__PURE__ */ (0, J.jsx)(c, {
		...typeof r == "string" ? { name: r } : { icon: r },
		className: "text-sm"
	}), m = /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [
		n !== void 0 && /* @__PURE__ */ (0, J.jsx)("span", {
			className: q("inline-flex items-center justify-center min-w-[16px] h-[16px] px-1 rounded-full text-[10px] font-bold text-white", i ?? "bg-muted-foreground"),
			children: n
		}),
		p,
		/* @__PURE__ */ (0, J.jsx)("span", {
			className: "truncate",
			children: t
		})
	] });
	if (f) {
		let r = /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [
			p,
			/* @__PURE__ */ (0, J.jsx)("span", {
				className: "truncate",
				children: t
			}),
			n !== void 0 && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "shrink-0 rounded-full border border-border px-1.5 text-[10px] leading-4 tabular-nums text-muted-foreground",
				title: n.toLocaleString(),
				"data-pill-count": "",
				children: Do(n)
			})
		] }), i = /* @__PURE__ */ (0, J.jsx)(Fo, {
			mode: e,
			onChange: a,
			interactive: u,
			...s ? { ariaLabel: s } : {}
		}), o = /* @__PURE__ */ (0, J.jsx)("span", {
			className: "inline-flex min-w-0 items-center gap-1.5 text-xs text-foreground",
			children: r
		});
		return /* @__PURE__ */ (0, J.jsx)("span", {
			className: q("inline-flex items-center gap-1.5 select-none", d),
			title: s,
			children: l === "right" ? /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [o, i] }) : /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [i, o] })
		});
	}
	return /* @__PURE__ */ (0, J.jsxs)("button", {
		type: "button",
		onClick: o,
		title: s,
		className: q("inline-flex items-center gap-1.5 text-xs px-2 py-0.5 rounded-full border transition-colors", No(e), d),
		children: [/* @__PURE__ */ (0, J.jsx)(Lo, { mode: e }), m]
	});
}
function Lo({ mode: e }) {
	return e === "include" ? /* @__PURE__ */ (0, J.jsx)(c, {
		icon: X,
		className: "text-xs"
	}) : e === "exclude" ? /* @__PURE__ */ (0, J.jsx)(c, {
		icon: Ze,
		className: "text-xs"
	}) : e === "active" ? /* @__PURE__ */ (0, J.jsx)("span", { className: "w-2 h-2 rounded-full bg-current" }) : /* @__PURE__ */ (0, J.jsx)("span", { className: "w-2 h-2 rounded-full bg-current opacity-30" });
}
function Ro(e) {
	return e === "include" ? "exclude" : e === "exclude" ? "neutral" : "include";
}
// ../../../clicky-ui/packages/ui/src/components/ComboboxTags.tsx
function zo({ values: e, ...t }) {
	return /* @__PURE__ */ (0, J.jsx)("span", {
		"data-combobox-more": !0,
		className: "inline-flex h-6 shrink-0",
		children: /* @__PURE__ */ (0, J.jsx)(uo, {
			role: "dialog",
			align: "left",
			menuLabel: "More selected",
			trigger: /* @__PURE__ */ (0, J.jsxs)("button", {
				type: "button",
				"aria-label": `Show ${e.length} more selected`,
				className: "inline-flex h-6 items-center whitespace-nowrap rounded-md border border-dashed border-border px-2 text-xs text-muted-foreground hover:bg-muted hover:text-foreground",
				children: [
					"+",
					e.length,
					" more"
				]
			}),
			children: () => /* @__PURE__ */ (0, J.jsx)("div", {
				"data-combobox-more-list": !0,
				className: "flex max-h-60 w-80 max-w-[calc(100vw-2rem)] flex-col items-start gap-1 overflow-y-auto p-1",
				children: /* @__PURE__ */ (0, J.jsx)(Bo, {
					values: e,
					...t
				})
			})
		})
	});
}
function Bo({ disabled: e, modes: t, onRemove: n, onSetMode: r, options: i, values: a }) {
	return a.map((a, s) => {
		let l = i.find((e) => e.value === a), u = l?.selectedLabel ?? l?.label ?? a, d = t?.[a], f = !e && /* @__PURE__ */ (0, J.jsx)("button", {
			type: "button",
			"aria-label": `Remove ${u}`,
			className: "shrink-0 text-muted-foreground hover:text-foreground",
			onMouseDown: (e) => e.preventDefault(),
			onClick: () => n(a),
			children: /* @__PURE__ */ (0, J.jsx)(c, {
				icon: j,
				className: "size-3"
			})
		});
		return d ? /* @__PURE__ */ (0, J.jsxs)("span", {
			"data-combobox-tag": a,
			"data-combobox-tag-mode": d,
			className: "inline-flex h-6 max-w-full items-center gap-1",
			children: [/* @__PURE__ */ (0, J.jsx)(Io, {
				mode: d,
				label: /* @__PURE__ */ (0, J.jsxs)("span", {
					className: "inline-flex min-w-0 items-center gap-1.5",
					children: [/* @__PURE__ */ (0, J.jsx)(o, {
						icon: l?.icon,
						className: "shrink-0 text-muted-foreground"
					}), /* @__PURE__ */ (0, J.jsx)("span", {
						className: "truncate",
						children: u
					})]
				}),
				title: d === "include" ? `${u} included` : `${u} excluded`,
				className: "max-w-full",
				...e ? {} : { onClick: () => r(a, d === "include" ? "exclude" : "include") }
			}), f]
		}, `${a}-${s}`) : /* @__PURE__ */ (0, J.jsx)(Vo, {
			value: a,
			...l ? { option: l } : {},
			children: f
		}, `${a}-${s}`);
	});
}
function Vo({ value: e, option: t, children: n }) {
	return /* @__PURE__ */ (0, J.jsxs)("span", {
		"data-combobox-tag": e,
		className: "inline-flex h-6 max-w-full items-center gap-1 rounded-md bg-muted px-2 text-xs",
		children: [
			/* @__PURE__ */ (0, J.jsx)(o, {
				icon: t?.icon,
				className: "shrink-0 text-muted-foreground"
			}),
			/* @__PURE__ */ (0, J.jsx)("span", {
				className: "truncate",
				children: t?.selectedLabel ?? t?.label ?? e
			}),
			n
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/useSingleRowTagCount.ts
var Ho = "[data-combobox-tag], [data-combobox-more], input";
function Uo(e, t) {
	let [n, r] = (0, Y.useState)(), [i, a] = (0, Y.useState)(0), o = t.join("\0"), s = n?.key === o && n.width === i ? n : void 0, c = s?.count ?? t.length;
	return (0, Y.useEffect)(() => {
		let t = e.current;
		if (!t || typeof ResizeObserver > "u") return;
		let n = new ResizeObserver(([e]) => {
			e && a(Math.round(e.contentRect.width));
		});
		return n.observe(t), () => n.disconnect();
	}, [e]), (0, Y.useLayoutEffect)(() => {
		let n = e.current;
		if (!n || c === 0) return;
		let a = [...n.children].filter((e) => e instanceof HTMLElement && e.matches(Ho)), l = a[0];
		if (!l) return;
		let u = l.offsetTop + Math.max(l.offsetHeight / 2, 1), d = a.some((e) => e.offsetTop > u), f = d ? s && s.min < c ? s.min : 0 : c, p = d ? c - 1 : s?.max ?? t.length, m = Math.ceil((f + p) / 2);
		m !== c && r({
			key: o,
			width: i,
			min: f,
			max: p,
			count: m
		});
	}), c;
}
// ../../../clicky-ui/packages/ui/src/components/ComboboxControl.tsx
function Wo({ anchorRef: e, ariaLabel: t, ariaRequired: n, closedLabel: r, describedBy: i, disabled: a, displayValue: o, effectivePrefix: s, highlighted: c, id: l, inputRef: u, invalid: d, label: f, labelRef: p, labelWidth: m, listId: h, loading: _, onClear: v, clearLabel: y, onInput: b, onKeyDown: x, onOpen: S, onPaste: C, onRemoveTag: w, onSetTagMode: T, onToggle: E, open: D, options: k, placeholder: A, inputPlaceholder: j, showClear: M, showToggle: N, size: P, suffix: F, tagModes: I, tagValues: L, tags: R }) {
	let z = [
		"pr-2",
		"pr-8",
		"pr-14",
		"pr-[5.5rem]"
	][Number(N) + Number(M) + Number(!!F)], B = [
		"right-0",
		"right-7",
		"right-[3.75rem]"
	][Number(N) + Number(M)], V = Uo(e, R ? L : []), H = L.slice(0, V), U = L.slice(V);
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		ref: e,
		"data-jsf-control": !0,
		"data-clearable-field": !0,
		className: q("relative flex items-center", R && "flex-wrap gap-1 rounded-md border border-input bg-background px-2 py-0.5 text-foreground focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-1", R && (P ? ue[P] : "min-h-control-h"), R && z, R && d && "border-destructive focus-within:ring-destructive", R && a && "cursor-not-allowed opacity-50"),
		onMouseDown: (e) => {
			R && !e.target.closest("button, input") && (e.preventDefault(), u.current?.focus());
		},
		children: [
			f != null && /* @__PURE__ */ (0, J.jsx)("span", {
				ref: p,
				className: q("pointer-events-none whitespace-nowrap font-medium uppercase tracking-wide text-muted-foreground text-[10px]", R ? "shrink-0" : "absolute left-2 z-10"),
				children: f
			}),
			R && s,
			R && /* @__PURE__ */ (0, J.jsx)(Bo, {
				disabled: a,
				modes: I,
				onRemove: w,
				onSetMode: T,
				options: k,
				values: H
			}),
			R && U.length > 0 && /* @__PURE__ */ (0, J.jsx)(zo, {
				disabled: a,
				modes: I,
				onRemove: w,
				onSetMode: T,
				options: k,
				values: U
			}),
			/* @__PURE__ */ (0, J.jsx)("input", {
				ref: u,
				id: l,
				type: "text",
				role: "combobox",
				"aria-expanded": D,
				"aria-controls": h,
				"aria-activedescendant": c >= 0 ? `${h}-${c}` : void 0,
				"aria-label": t,
				"aria-invalid": d || void 0,
				"aria-required": n || void 0,
				"aria-describedby": i,
				autoComplete: "off",
				disabled: a,
				placeholder: j ?? (R && L.length > 0 ? void 0 : A),
				value: o,
				...!R && !D && r ? { title: r } : {},
				onChange: (e) => b(e.target.value),
				onFocus: S,
				onClick: S,
				onKeyDown: x,
				onPaste: C,
				className: q(R ? "min-w-28 flex-1 bg-transparent px-1 py-1 text-foreground outline-none" : "w-full rounded-md border border-input bg-background text-foreground", R ? P ? O[P] : "text-sm" : P ? g[P] : "h-control-h px-control-px text-sm", "max-sm:text-base", !R && s && "pl-8", !R && z, !R && "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1", !R && "disabled:cursor-not-allowed disabled:opacity-50", !R && d && "border-destructive focus-visible:ring-destructive"),
				style: !R && f != null ? mo(f, m) : void 0
			}),
			!R && s && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "absolute inset-y-0 left-1.5 flex items-center",
				children: s
			}),
			F && /* @__PURE__ */ (0, J.jsx)("div", {
				className: q("absolute flex h-full items-center", B),
				children: F
			}),
			/* @__PURE__ */ (0, J.jsx)(yo, {
				disabled: a,
				loading: _,
				onClear: v,
				...y ? { clearLabel: y } : {},
				onToggle: E,
				showClear: M,
				showToggle: N
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/ComboboxMenuTrigger.tsx
function Go({ anchorRef: e, triggerRef: t, content: n, className: r, disabled: i, open: a, listId: o, ariaLabel: s, title: l, onToggle: u, clearable: d, onClear: f, clearLabel: p }) {
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		ref: e,
		"data-clearable-field": !0,
		className: "relative h-full min-w-0",
		children: [/* @__PURE__ */ (0, J.jsxs)("button", {
			ref: t,
			type: "button",
			role: "combobox",
			"aria-label": s,
			"aria-haspopup": "listbox",
			"aria-expanded": a,
			"aria-controls": a ? o : void 0,
			title: l,
			disabled: i,
			onClick: u,
			onKeyDown: (e) => {
				(e.key === "ArrowDown" || e.key === "ArrowUp") && (e.preventDefault(), a || u());
			},
			className: q("inline-flex h-control-h w-full min-w-0 items-center gap-1.5 rounded-md border border-input bg-background px-control-px text-sm hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50", r, d && "pr-14"),
			children: [n, /* @__PURE__ */ (0, J.jsx)(c, {
				icon: pe,
				className: "pointer-events-none absolute right-2 top-1/2 size-3 shrink-0 -translate-y-1/2 text-muted-foreground"
			})]
		}), d && !i && /* @__PURE__ */ (0, J.jsx)("span", {
			className: "absolute inset-y-0 right-7 flex items-center",
			children: /* @__PURE__ */ (0, J.jsx)(y, {
				...f ? { onClear: f } : {},
				...p ? { clearLabel: p } : {}
			})
		})]
	});
}
function Ko({ inputRef: e, value: t, ariaLabel: n, listId: r, highlighted: i, onChange: a, onKeyDown: o }) {
	return (0, Y.useEffect)(() => {
		e.current?.focus();
	}, [e]), /* @__PURE__ */ (0, J.jsx)("div", {
		className: "sticky top-0 z-10 border-b border-border bg-popover p-1",
		children: /* @__PURE__ */ (0, J.jsx)("input", {
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
// ../../../clicky-ui/packages/ui/src/lib/breakpoints.ts
var qo = {
	sm: 640,
	md: 768,
	lg: 1024,
	xl: 1280,
	"2xl": 1536,
	"3xl": 1920
};
function Jo(e) {
	return `(max-width: ${qo[e] - 1}px)`;
}
var Yo = Jo("sm"), Xo = 8, Zo = 4, Qo = 16;
function $o({ anchor: e, viewportWidth: t, viewportHeight: n, mobile: r, naturalHeight: i, preferredWidth: a }) {
	if (a !== void 0 && (!Number.isFinite(a) || a <= 0)) throw Error("Combobox menu width must be a positive finite number.");
	if (r) {
		let r = Math.max(0, n - e.bottom - Zo - Xo), a = Math.max(0, e.top - Zo - Xo), o = i != null && i > r && a > r, s = Math.max(0, t - 32);
		return es({
			strategy: "mobile",
			anchor: e,
			viewportHeight: n,
			openUp: o,
			left: Qo,
			width: s,
			maxWidth: s,
			maxHeight: o ? a : r
		});
	}
	let o = Math.max(0, n - e.bottom - Zo - Xo), s = Math.max(0, e.top - Zo - Xo), c = i != null && i > o && s > o, l = a === void 0 ? void 0 : Math.min(Math.max(e.width, a), Math.max(0, t - 16));
	return es({
		strategy: "desktop",
		anchor: e,
		viewportHeight: n,
		openUp: c,
		left: l === void 0 ? e.left : Math.max(Xo, Math.min(e.left, t - l - Xo)),
		width: l ?? e.width,
		maxWidth: l ?? Math.max(e.width, Math.min(400, t - e.left - Xo)),
		maxHeight: c ? s : o
	});
}
function es({ strategy: e, anchor: t, viewportHeight: n, openUp: r, left: i, width: a, maxWidth: o, maxHeight: s }) {
	return {
		strategy: e,
		...r ? { bottom: n - t.top + Zo } : { top: t.bottom + Zo },
		left: i,
		width: a,
		maxWidth: o,
		maxHeight: s
	};
}
// ../../../clicky-ui/packages/ui/src/components/useComboboxLayout.ts
function ts(e, t) {
	let [n, r] = (0, Y.useState)(0);
	return (0, Y.useLayoutEffect)(() => {
		e == null ? r(0) : t.current && r(t.current.offsetWidth);
	}, [e, t]), n;
}
function ns({ open: e, anchorRef: t, listRef: n, menuWidth: r }) {
	let [a, o] = (0, Y.useState)(null), s = (0, Y.useCallback)(() => {
		let a = t.current;
		if (!e || !a) return;
		let s = $o({
			anchor: a.getBoundingClientRect(),
			viewportWidth: window.innerWidth,
			viewportHeight: window.innerHeight,
			mobile: i(Yo),
			...r === void 0 ? {} : { preferredWidth: r },
			...n.current ? { naturalHeight: n.current.scrollHeight } : {}
		});
		o((e) => rs(e, s) ? e : s);
	}, [
		t,
		n,
		e,
		r
	]);
	return (0, Y.useLayoutEffect)(() => {
		if (!e) o(null);
		else return s(), window.addEventListener("scroll", s, !0), window.addEventListener("resize", s), () => {
			window.removeEventListener("scroll", s, !0), window.removeEventListener("resize", s);
		};
	}, [e, s]), (0, Y.useLayoutEffect)(() => {
		let t = n.current;
		if (!e || !a || !t) return;
		s();
		let r = typeof ResizeObserver > "u" ? null : new ResizeObserver(s), i = new MutationObserver(s);
		return r?.observe(t), i.observe(t, {
			childList: !0,
			characterData: !0,
			subtree: !0
		}), () => {
			r?.disconnect(), i.disconnect();
		};
	}, [
		n,
		e,
		a,
		s
	]), a;
}
function rs(e, t) {
	return e?.strategy === t.strategy && e.top === t.top && e.bottom === t.bottom && e.left === t.left && e.width === t.width && e.maxWidth === t.maxWidth && e.maxHeight === t.maxHeight;
}
// ../../../clicky-ui/packages/ui/src/components/ComboboxMenu.tsx
function is({ customEntry: e, filtered: t, floatingZ: n, footer: r, header: i, ariaLabel: a, hasOptions: s, highlighted: l, isSelected: u, listId: d, listRef: f, loading: p, menuPos: m, modes: h, multiple: g, onHighlight: _, onSelect: v, onSetMode: y, tristate: b }) {
	let x = m.strategy === "mobile", S = /* @__PURE__ */ (0, J.jsxs)("div", {
		id: i ? void 0 : d,
		ref: f,
		role: i ? void 0 : "listbox",
		"aria-label": i ? void 0 : a,
		"aria-multiselectable": !i && g ? !0 : void 0,
		style: {
			position: x ? "absolute" : "fixed",
			...m.top == null ? { bottom: m.bottom } : { top: m.top },
			left: m.left,
			...x ? { width: m.width } : { minWidth: m.width },
			maxWidth: m.maxWidth,
			maxHeight: m.maxHeight,
			zIndex: x ? void 0 : n,
			pointerEvents: x ? "auto" : void 0
		},
		className: q("overflow-auto rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-lg shadow-black/5", !x && "w-max"),
		children: [i, /* @__PURE__ */ (0, J.jsxs)("div", {
			className: q(t.some((e) => e.layout === "responsive") && "@container/combobox-options"),
			id: i ? d : void 0,
			role: i ? "listbox" : void 0,
			"aria-label": i ? a : void 0,
			"aria-multiselectable": i && g ? !0 : void 0,
			children: [
				p && t.length === 0 && /* @__PURE__ */ (0, J.jsx)("div", {
					className: "px-2 py-4 text-center text-sm text-muted-foreground",
					children: "Loading…"
				}),
				!p && t.length === 0 && !e && s && /* @__PURE__ */ (0, J.jsx)("div", {
					className: "px-2 py-4 text-center text-sm text-muted-foreground",
					children: "No results"
				}),
				t.map((e, n) => {
					let r = u(e.value), i = b ? h[e.value] : void 0, a = e.group != null && e.group !== t[n - 1]?.group, s = /* @__PURE__ */ (0, J.jsx)("span", {
						"data-slot": "combobox-selection-indicator",
						"aria-hidden": !0,
						className: q("shrink-0 text-xs", e.description && e.descriptionLayout !== "inline" && "mt-1", r ? "opacity-100" : "opacity-0"),
						children: /* @__PURE__ */ (0, J.jsx)(c, { icon: bt })
					});
					return /* @__PURE__ */ (0, J.jsxs)(Y.Fragment, { children: [a && /* @__PURE__ */ (0, J.jsx)("div", {
						role: "presentation",
						className: "select-none px-2 pb-1 pt-2 text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
						children: e.group
					}), /* @__PURE__ */ (0, J.jsx)("div", {
						id: d ? `${d}-${n}` : void 0,
						role: "option",
						"aria-selected": b ? i != null : r,
						"aria-disabled": e.disabled,
						...i == null ? {} : { "aria-label": `${e.label}, ${i === "include" ? "included" : "excluded"}` },
						...e.title === void 0 ? {} : { title: e.title },
						...b ? { "data-filter-option": e.value } : {},
						onMouseDown: (t) => {
							t.preventDefault(), !b && !e.disabled && v(e);
						},
						...b ? { onClick: () => {
							e.disabled || v(e);
						} } : {},
						onMouseEnter: () => _(n),
						className: q("flex cursor-pointer rounded-sm py-1.5 text-sm", g ? "gap-2 px-2" : "gap-1.5 px-1.5", e.layout === "responsive" ? "items-start @min-[28rem]/combobox-options:items-center" : e.description && e.descriptionLayout !== "inline" ? "items-start" : "items-center", n === l && "bg-accent", r && !b && "font-medium", e.disabled && "cursor-not-allowed opacity-50"),
						children: b ? /* @__PURE__ */ (0, J.jsxs)("div", {
							className: "min-w-0 flex-1",
							children: [/* @__PURE__ */ (0, J.jsx)(Io, {
								mode: i ?? "neutral",
								togglePosition: "right",
								interactive: !1,
								label: /* @__PURE__ */ (0, J.jsxs)("span", {
									className: "inline-flex min-w-0 items-center gap-1.5",
									children: [/* @__PURE__ */ (0, J.jsx)(o, {
										icon: e.icon,
										className: "shrink-0 text-sm text-muted-foreground"
									}), /* @__PURE__ */ (0, J.jsx)("span", {
										className: "truncate",
										children: e.label
									})]
								}),
								onModeChange: (t) => y(e.value, t),
								className: "w-full justify-between",
								...e.count === void 0 ? {} : { count: e.count }
							}), /* @__PURE__ */ (0, J.jsx)(os, { description: e.description })]
						}) : /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [
							g && s,
							/* @__PURE__ */ (0, J.jsx)(o, {
								icon: e.icon,
								className: q("shrink-0 text-sm text-muted-foreground", e.description && e.descriptionLayout !== "inline" && "mt-1")
							}),
							/* @__PURE__ */ (0, J.jsx)(as, { option: e }),
							e.trailing != null && e.layout !== "responsive" && /* @__PURE__ */ (0, J.jsx)("span", {
								className: "ml-auto min-w-0 max-w-[55%] truncate pl-2 text-right text-xs text-muted-foreground",
								children: e.trailing
							}),
							!g && s
						] })
					})] }, e.value);
				}),
				e && /* @__PURE__ */ (0, J.jsxs)("div", {
					id: d ? `${d}-${t.length}` : void 0,
					role: "option",
					"aria-selected": !1,
					"data-filter-add-custom": e.value,
					title: e.title,
					onMouseDown: (e) => e.preventDefault(),
					onClick: () => v(e),
					onMouseEnter: () => _(t.length),
					className: q("flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-primary", l === t.length && "bg-accent"),
					children: [/* @__PURE__ */ (0, J.jsx)(o, {
						icon: e.icon,
						className: "text-sm text-muted-foreground"
					}), /* @__PURE__ */ (0, J.jsx)("span", {
						className: "min-w-0 truncate",
						children: e.label
					})]
				}),
				r != null && /* @__PURE__ */ (0, J.jsx)("div", {
					className: "select-none px-2 py-1.5 text-[11px] text-muted-foreground",
					children: r
				})
			]
		})]
	});
	return (0, Fe.createPortal)(x ? /* @__PURE__ */ (0, J.jsx)("div", {
		"data-slot": "combobox-viewport-layer",
		style: {
			position: "fixed",
			inset: 0,
			zIndex: n,
			pointerEvents: "none"
		},
		children: S
	}) : S, document.body);
}
function as({ option: e }) {
	let t = e.layout === "responsive";
	return /* @__PURE__ */ (0, J.jsxs)("span", {
		"data-slot": "combobox-option-text",
		className: q("min-w-0 flex-1", t ? "flex flex-col gap-1 @min-[28rem]/combobox-options:flex-row @min-[28rem]/combobox-options:items-center @min-[28rem]/combobox-options:gap-2" : e.descriptionLayout === "inline" && "flex items-center gap-2"),
		children: [
			/* @__PURE__ */ (0, J.jsx)("span", {
				className: q("block min-w-0", t ? "whitespace-normal [overflow-wrap:anywhere] @min-[28rem]/combobox-options:flex-1 @min-[28rem]/combobox-options:truncate" : "truncate"),
				children: e.label
			}),
			/* @__PURE__ */ (0, J.jsx)(os, {
				description: e.description,
				inline: e.descriptionLayout === "inline",
				responsive: t
			}),
			t && e.trailing != null && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "min-w-0 text-xs font-normal text-muted-foreground @min-[28rem]/combobox-options:ml-auto @min-[28rem]/combobox-options:max-w-[55%] @min-[28rem]/combobox-options:truncate @min-[28rem]/combobox-options:pl-2 @min-[28rem]/combobox-options:text-right",
				children: e.trailing
			})
		]
	});
}
function os({ description: e, inline: t = !1, responsive: n = !1 }) {
	return e ? /* @__PURE__ */ (0, J.jsx)("span", {
		className: q("text-xs font-normal leading-4 text-muted-foreground", n ? "min-w-0 whitespace-normal [overflow-wrap:anywhere] @min-[28rem]/combobox-options:max-w-[55%] @min-[28rem]/combobox-options:shrink-0 @min-[28rem]/combobox-options:truncate" : t ? "min-w-0 truncate" : "mt-0.5 block whitespace-pre-line"),
		children: e
	}) : null;
}
// ../../../clicky-ui/packages/ui/src/components/Combobox.tsx
function ss(e) {
	let { options: t, placeholder: n, inputPlaceholder: r, label: i, ariaLabel: a, disabled: s, required: c, invalid: l, ariaRequired: u, describedBy: d, allowCustomValue: f = !0, onNew: p, onCreate: m, id: h, size: g, className: _, loading: v, onSearch: y, onKeyDown: b, suffix: x, prefix: S, footer: C, menuHeader: w } = e, T = e.multiple === !0, E = !T && e.searchPlacement === "menu", D = (0, Y.useId)(), O = (0, Y.useRef)(null), k = e.multiple === !0 && e.tristate === !0, A = e.multiple === !0 && e.variant === "tags", j = e.multiple === !0 && e.tristate !== !0 ? e.separators ?? [] : [], M = (0, Y.useMemo)(() => e.multiple === !0 && e.tristate === !0 ? e.value : {}, [
		e.multiple,
		e.tristate,
		e.value
	]), N = (0, Y.useMemo)(() => e.multiple === !0 && e.tristate === !0 ? Object.keys(e.value) : e.multiple === !0 ? e.value : e.value ? [e.value] : [], [
		e.multiple,
		e.tristate,
		e.value
	]), P = (0, Y.useRef)(null), F = (0, Y.useRef)(null), I = (0, Y.useRef)(null), L = (0, Y.useRef)(null), R = (0, Y.useRef)(null), [z, B] = (0, Y.useState)(!1), V = ro(), H = ns({
		open: z,
		anchorRef: F,
		listRef: L,
		...e.menuWidth === void 0 ? {} : { menuWidth: e.menuWidth }
	}), [U, ee] = (0, Y.useState)(""), te = (0, Y.useRef)(e.value);
	(0, Y.useEffect)(() => {
		e.onInputValueChange && e.multiple !== !0 && te.current !== e.value && z && ee(e.value), te.current = e.value;
	}, [
		e.value,
		e.multiple,
		e.onInputValueChange,
		z
	]);
	let [W, ne] = (0, Y.useState)(-1), re = ts(i, R), ie = (e) => N.includes(e), ae = (0, Y.useMemo)(() => {
		if (T) return;
		let e = N[0];
		if (e) return t.find((t) => t.value === e) ?? (f ? p?.(e) ?? void 0 : void 0);
	}, [
		f,
		T,
		p,
		t,
		N
	]), oe = (0, Y.useMemo)(() => {
		if (k) {
			let e = Object.values(M), t = e.filter((e) => e === "include").length, n = e.length - t;
			return [t > 0 && `+${t}`, n > 0 && `-${n}`].filter(Boolean).join(" ");
		}
		if (T) return go(t, N);
		let e = N[0] ?? "";
		return ae?.selectedLabel ?? ae?.label ?? e;
	}, [
		k,
		M,
		T,
		t,
		ae,
		N
	]), G = S ?? (ae?.icon == null ? null : /* @__PURE__ */ (0, J.jsx)(o, {
		icon: ae.icon,
		className: "size-4"
	})), se = A || z ? U : oe, K = (0, Y.useMemo)(() => {
		if (y) return ho(t, N);
		let e = U.toLowerCase().trim();
		return e ? t.filter((t) => t.label.toLowerCase().includes(e) || t.description?.toLowerCase().includes(e) || t.group?.toLowerCase().includes(e) || t.value.toLowerCase().includes(e)) : t;
	}, [
		y,
		t,
		U,
		N
	]);
	(0, Y.useEffect)(() => {
		if (!y || !z) return;
		let e = setTimeout(() => y(U.trim()), 250);
		return () => clearTimeout(e);
	}, [
		y,
		z,
		U
	]);
	let ce = U.trim(), le = (0, Y.useMemo)(() => E && K.length > 0 ? null : vo({
		allowCustomValue: f,
		query: ce,
		multiple: T,
		tristate: k,
		onNew: p,
		choices: t,
		selectedValues: N
	}), [
		f,
		E,
		K.length,
		T,
		p,
		t,
		N,
		ce,
		k
	]), ue = (0, Y.useMemo)(() => le ? [...K, le] : K, [K, le]);
	(0, Y.useEffect)(() => {
		ne(-1);
	}, [U]), (0, Y.useEffect)(() => {
		W >= ue.length && ne(-1);
	}, [W, ue.length]), (0, Y.useEffect)(() => {
		if (!z) return;
		let e = (e) => {
			let t = e.target;
			!P.current?.contains(t) && !L.current?.contains(t) && be();
		};
		return document.addEventListener("mousedown", e), document.addEventListener("focusin", e), () => {
			document.removeEventListener("mousedown", e), document.removeEventListener("focusin", e);
		};
	});
	function de(t) {
		e.multiple === !0 && e.tristate !== !0 ? e.onChange(t) : e.multiple !== !0 && e.onChange(t[0] ?? "");
	}
	function fe(t) {
		e.multiple === !0 && e.tristate === !0 && e.onChange(t);
	}
	function pe(e, t) {
		let n = { ...M };
		t === "include" || t === "exclude" ? n[e] = t : delete n[e], fe(n);
	}
	function me(e) {
		let t = M[e];
		pe(e, t === void 0 ? "include" : t === "include" ? "exclude" : "neutral");
	}
	function he() {
		z || (ee(""), ne(-1), B(!0));
	}
	function ge(e) {
		return m ? m(e) : e.value;
	}
	function _e(e) {
		k ? pe(e, "include") : T ? N.includes(e) || de([...N, e]) : de([e]);
	}
	function ve(e) {
		let n = [...N];
		for (let r of e) {
			let e = r.trim();
			if (!e || n.includes(e)) continue;
			let i = t.find((t) => t.value === e), a = i ?? p?.(e) ?? {
				value: e,
				label: e
			}, o = i ? i.value : ge(a);
			o !== null && !n.includes(o) && n.push(o);
		}
		n.length !== N.length && de(n), ee(""), ne(-1);
	}
	function ye(e) {
		if (!A || !f || j.length === 0) return;
		let t = _o(e.clipboardData.getData("text"), j);
		t.length < 2 || (e.preventDefault(), ve(t));
	}
	function be() {
		if (E) {
			ee(""), B(!1), ne(-1);
			return;
		}
		let e = U.trim();
		if (le) {
			let e = ge(le);
			e !== null && _e(e);
		} else if (!T && f && e && e !== N[0]) {
			let n = t.find((t) => t.value === e), r = n ?? (p ? p(e) : {
				value: e,
				label: e
			}), i = r ? n ? n.value : ge(r) : null;
			i !== null && de([i]);
		}
		ee(""), B(!1), ne(-1);
	}
	function xe(e) {
		if (e === le) {
			let t = ge(e);
			t !== null && _e(t), ee(""), ne(-1), T || B(!1), (E ? O : I).current?.focus();
		} else k ? (me(e.value), I.current?.focus()) : T ? (de(ie(e.value) ? N.filter((t) => t !== e.value) : [...N, e.value]), ne(-1), I.current?.focus()) : (de([e.value]), ee(""), B(!1), ne(-1), (E ? O : I).current?.focus());
	}
	function Se() {
		N.length > 0 && (e.onClear ? e.onClear() : k ? fe({}) : de([])), ee(""), ne(-1), I.current?.focus(), B(!1);
	}
	function Ce(e) {
		k ? pe(e, "neutral") : de(N.filter((t) => t !== e)), I.current?.focus();
	}
	function we(e) {
		let t = L.current;
		t && t.querySelectorAll("[role=\"option\"]")[e]?.scrollIntoView?.({ block: "nearest" });
	}
	function Te(e) {
		if (A && f && j.includes(e.key)) e.preventDefault(), ve([U]);
		else if (A && e.key === "Backspace" && U === "" && N.length > 0) {
			e.preventDefault();
			let t = N[N.length - 1];
			k ? t && pe(t, "neutral") : de(N.slice(0, -1));
		} else if (e.key === "ArrowDown") {
			if (e.preventDefault(), !z) {
				he();
				return;
			}
			let t = W < ue.length - 1 ? W + 1 : 0;
			ne(t), we(t);
		} else if (e.key === "ArrowUp") {
			if (e.preventDefault(), !z) return;
			let t = W > 0 ? W - 1 : ue.length - 1;
			ne(t), we(t);
		} else if (e.key === "Enter") {
			if (E && e.preventDefault(), z && W >= 0 && ue[W] && !ue[W].disabled) {
				e.preventDefault(), xe(ue[W]);
				return;
			}
			if (z && (T || E) && le) {
				e.preventDefault(), xe(le);
				return;
			}
			be(), b?.(e);
		} else e.key === "Escape" ? z && (e.preventDefault(), E && e.stopPropagation(), ee(""), B(!1), ne(-1), E && O.current?.focus()) : b?.(e);
	}
	let Ee = `${h ?? D}-listbox`, De = a ?? (typeof i == "string" ? i : void 0), Oe = (e.clearable ?? !c) && !v && !s && N.length > 0, ke = !A || t.length > 0 || !!y || !!v, Ae = z && (ke || le !== null);
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		ref: P,
		className: q("relative", _),
		children: [E ? /* @__PURE__ */ (0, J.jsx)(Go, {
			anchorRef: F,
			triggerRef: O,
			content: e.triggerContent ?? /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [G, /* @__PURE__ */ (0, J.jsx)("span", {
				className: "min-w-0 truncate",
				children: oe || n
			})] }),
			className: e.triggerClassName,
			disabled: s,
			open: Ae,
			listId: Ee,
			ariaLabel: De,
			title: e.title,
			clearable: e.clearable === !0 && Oe,
			onClear: Se,
			clearLabel: e.clearLabel ?? "Clear",
			onToggle: () => z ? be() : he()
		}) : /* @__PURE__ */ (0, J.jsx)(Wo, {
			anchorRef: F,
			ariaLabel: De,
			ariaRequired: u,
			closedLabel: oe,
			describedBy: d,
			disabled: s,
			displayValue: se,
			effectivePrefix: G,
			highlighted: Ae ? W : -1,
			id: h,
			inputRef: I,
			invalid: l,
			label: i,
			labelRef: R,
			labelWidth: re,
			listId: Ee,
			loading: v,
			onClear: Se,
			...e.clearLabel ? { clearLabel: e.clearLabel } : {},
			onInput: (t) => {
				ee(t), e.onInputValueChange?.(t), z || B(!0);
			},
			onKeyDown: Te,
			onOpen: he,
			onPaste: ye,
			onRemoveTag: Ce,
			onToggle: () => {
				z ? be() : (he(), I.current?.focus());
			},
			open: Ae,
			options: t,
			placeholder: n,
			inputPlaceholder: r,
			showToggle: ke,
			showClear: Oe,
			size: g,
			suffix: x,
			tagModes: k ? M : void 0,
			onSetTagMode: pe,
			tagValues: N,
			tags: A
		}), Ae && H && typeof document < "u" && /* @__PURE__ */ (0, J.jsx)(is, {
			customEntry: le,
			filtered: K,
			floatingZ: V,
			footer: C,
			ariaLabel: De,
			header: E || w ? /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [w, E ? /* @__PURE__ */ (0, J.jsx)(Ko, {
				inputRef: I,
				value: U,
				ariaLabel: De,
				listId: Ee,
				highlighted: W,
				onChange: (t) => {
					ee(t), e.onInputValueChange?.(t);
				},
				onKeyDown: Te
			}) : null] }) : void 0,
			hasOptions: t.length > 0,
			highlighted: W,
			isSelected: ie,
			listId: Ee,
			listRef: L,
			loading: v,
			menuPos: H,
			modes: M,
			multiple: T,
			onHighlight: ne,
			onSelect: xe,
			onSetMode: pe,
			tristate: k
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/components/jsonPathTree.ts
var cs = /^[A-Za-z_$][A-Za-z0-9_$]*$/, ls = 1e3;
function us(e) {
	return Array.isArray(e) ? "array" : typeof e == "object" && e ? "object" : "scalar";
}
function ds(e, t) {
	return typeof t == "number" ? `${e}[${t}]` : cs.test(t) ? `${e}.${t}` : `${e}[${JSON.stringify(t)}]`;
}
function fs(e) {
	if (Array.isArray(e)) return `[${e.length} item${e.length === 1 ? "" : "s"}]`;
	if (typeof e == "object" && e) {
		let t = Object.keys(e).length;
		return `{${t} ${t === 1 ? "property" : "properties"}}`;
	}
	return typeof e == "string" ? JSON.stringify(e.length > 80 ? `${e.slice(0, 77)}…` : e) : String(e);
}
function ps(e) {
	return Array.isArray(e) ? e.length : typeof e == "object" && e ? Object.keys(e).length : 0;
}
function ms(e, t, n, r = {}) {
	let { maxDepth: i = 12, maxObjectProperties: a = 100, arrayMode: o = "first" } = r, s = us(e), c = {
		key: t,
		path: t,
		value: e,
		kind: s,
		summary: fs(e),
		childCount: ps(e)
	};
	return s === "scalar" || n >= i || (c.children = (Array.isArray(e) ? o === "all" ? e.map((e, t) => [t, e]) : e.length > 0 ? [[0, e[0]]] : [] : Object.entries(e).slice(0, a)).map(([e, i]) => ms(i, ds(t, e), n + 1, r))), c;
}
function hs(e, t = {}) {
	let { maxChildren: n = ls, parseEmbeddedJSON: r = !0, keyPrefix: i = "" } = t, a = /* @__PURE__ */ new Map(), o = /* @__PURE__ */ new Map();
	function s(e) {
		if (!r || typeof e.value != "string") return;
		if (o.has(e.key)) return o.get(e.key);
		let t = e.value.trim(), n;
		if (t.startsWith("{") || t.startsWith("[")) try {
			let e = JSON.parse(t);
			us(e) !== "scalar" && (n = e);
		} catch {
			n = void 0;
		}
		return o.set(e.key, n), n;
	}
	function c(e) {
		let t = s(e), n = t === void 0 ? {
			value: e.value,
			path: e.path
		} : {
			value: t,
			path: "$"
		}, r = t === void 0 ? e.root : e.root ?? gs(e.path);
		r !== void 0 && (n.root = r);
		let i = t === void 0 ? e.origin : {
			path: e.path,
			kind: us(t),
			...e.origin ? { outer: e.origin } : {}
		};
		return i !== void 0 && (n.origin = i), n;
	}
	function l(e, t, n, r, i) {
		let o = a.get(n);
		if (o) return o;
		let s = {
			key: n,
			path: t,
			value: e,
			kind: us(e),
			summary: fs(e),
			childCount: ps(e)
		};
		return r !== void 0 && (s.root = r), i !== void 0 && (s.origin = i), a.set(n, s), s;
	}
	function u(e) {
		return Array.isArray(e) ? e.map((e, t) => [t, e]) : typeof e == "object" && e ? Object.entries(e) : [];
	}
	function d(e) {
		if (e.children) return e.children;
		let t = c(e), r = u(t.value), i = Math.min(r.length, n), a = r.slice(0, i).map(([n, r]) => {
			let i = ds(t.path, n);
			return l(r, i, `${e.key}>${i}`, t.root, t.origin);
		});
		return i < r.length && a.push({
			key: `${e.key}>…`,
			path: t.path,
			value: void 0,
			kind: "more",
			summary: `${r.length - i} more not shown — address them by index`,
			childCount: r.length - i
		}), e.children = a, a;
	}
	function f(e) {
		if (e.kind === "more") return !1;
		if (e.kind !== "scalar") return e.childCount > 0;
		let t = s(e);
		return t !== void 0 && ps(t) > 0;
	}
	let p = l(e, "$", `${i}$`, void 0, void 0);
	return {
		roots: e === void 0 ? [] : [p],
		getChildren: (e) => e.children,
		hasMoreChildren: (e) => e.children === void 0 && f(e),
		loadChildren: (e) => Promise.resolve(d(e)),
		materializePath: (e) => {
			let t = _s(e);
			if (!t) return;
			let n = p;
			for (let e of t) {
				let t = ds(c(n).path, e), r = d(n).find((e) => e.path === t);
				if (!r) return;
				n = r;
			}
			return n;
		}
	};
}
function gs(e) {
	return _s(e)?.slice(0, 1).map(String)[0];
}
function _s(e) {
	let t = e.trim();
	if (t === "" || !t.startsWith("$")) return;
	t = t.slice(1);
	let n = [];
	for (; t.length > 0;) {
		if (t.startsWith("..")) return;
		if (t.startsWith(".")) {
			let e = /^\.([A-Za-z_$][A-Za-z0-9_$]*)/.exec(t);
			if (!e) return;
			n.push(e[1]), t = t.slice(e[0].length);
		} else if (t.startsWith("[")) {
			let e = t.indexOf("]");
			if (e < 0) return;
			let r = t.slice(1, e).trim();
			if (/^\d+$/.test(r)) n.push(Number(r));
			else if (/^"(?:[^"\\]|\\.)*"$/.test(r)) n.push(JSON.parse(r));
			else if (/^'[^']*'$/.test(r)) n.push(r.slice(1, -1));
			else return;
			t = t.slice(e + 1);
		} else return;
	}
	return n;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/nodes/identity.js
var vs = Symbol.for("yaml.alias"), ys = Symbol.for("yaml.document"), bs = Symbol.for("yaml.map"), xs = Symbol.for("yaml.pair"), Ss = Symbol.for("yaml.scalar"), Cs = Symbol.for("yaml.seq"), ws = Symbol.for("yaml.node.type"), Ts = (e) => !!e && typeof e == "object" && e[ws] === vs, Es = (e) => !!e && typeof e == "object" && e[ws] === ys, Ds = (e) => !!e && typeof e == "object" && e[ws] === bs, Os = (e) => !!e && typeof e == "object" && e[ws] === xs, ks = (e) => !!e && typeof e == "object" && e[ws] === Ss, As = (e) => !!e && typeof e == "object" && e[ws] === Cs;
function js(e) {
	if (e && typeof e == "object") switch (e[ws]) {
		case bs:
		case Cs: return !0;
	}
	return !1;
}
function Ms(e) {
	if (e && typeof e == "object") switch (e[ws]) {
		case vs:
		case bs:
		case Ss:
		case Cs: return !0;
	}
	return !1;
}
var Ns = (e) => (ks(e) || js(e)) && !!e.anchor, Ps = Symbol("break visit"), Fs = Symbol("skip children"), Is = Symbol("remove node");
function Ls(e, t) {
	let n = zs(t);
	Es(e) ? Rs(null, e.contents, n, Object.freeze([e])) === Is && (e.contents = null) : Rs(null, e, n, Object.freeze([]));
}
Ls.BREAK = Ps, Ls.SKIP = Fs, Ls.REMOVE = Is;
function Rs(e, t, n, r) {
	let i = Bs(e, t, n, r);
	if (Ms(i) || Os(i)) return Vs(e, r, i), Rs(e, i, n, r);
	if (typeof i != "symbol") {
		if (js(t)) {
			r = Object.freeze(r.concat(t));
			for (let e = 0; e < t.items.length; ++e) {
				let i = Rs(e, t.items[e], n, r);
				if (typeof i == "number") e = i - 1;
				else if (i === Ps) return Ps;
				else i === Is && (t.items.splice(e, 1), --e);
			}
		} else if (Os(t)) {
			r = Object.freeze(r.concat(t));
			let e = Rs("key", t.key, n, r);
			if (e === Ps) return Ps;
			e === Is && (t.key = null);
			let i = Rs("value", t.value, n, r);
			if (i === Ps) return Ps;
			i === Is && (t.value = null);
		}
	}
	return i;
}
function zs(e) {
	return typeof e == "object" && (e.Collection || e.Node || e.Value) ? Object.assign({
		Alias: e.Node,
		Map: e.Node,
		Scalar: e.Node,
		Seq: e.Node
	}, e.Value && {
		Map: e.Value,
		Scalar: e.Value,
		Seq: e.Value
	}, e.Collection && {
		Map: e.Collection,
		Seq: e.Collection
	}, e) : e;
}
function Bs(e, t, n, r) {
	if (typeof n == "function") return n(e, t, r);
	if (Ds(t)) return n.Map?.(e, t, r);
	if (As(t)) return n.Seq?.(e, t, r);
	if (Os(t)) return n.Pair?.(e, t, r);
	if (ks(t)) return n.Scalar?.(e, t, r);
	if (Ts(t)) return n.Alias?.(e, t, r);
}
function Vs(e, t, n) {
	let r = t[t.length - 1];
	if (js(r)) r.items[e] = n;
	else if (Os(r)) e === "key" ? r.key = n : r.value = n;
	else if (Es(r)) r.contents = n;
	else {
		let e = Ts(r) ? "alias" : "scalar";
		throw Error(`Cannot replace node with ${e} parent`);
	}
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/doc/directives.js
var Hs = {
	"!": "%21",
	",": "%2C",
	"[": "%5B",
	"]": "%5D",
	"{": "%7B",
	"}": "%7D"
}, Us = (e) => e.replace(/[!,[\]{}]/g, (e) => Hs[e]), Ws = class e {
	constructor(t, n) {
		this.docStart = null, this.docEnd = !1, this.yaml = Object.assign({}, e.defaultYaml, t), this.tags = Object.assign({}, e.defaultTags, n);
	}
	clone() {
		let t = new e(this.yaml, this.tags);
		return t.docStart = this.docStart, t;
	}
	atDocument() {
		let t = new e(this.yaml, this.tags);
		switch (this.yaml.version) {
			case "1.1":
				this.atNextDocument = !0;
				break;
			case "1.2": this.atNextDocument = !1, this.yaml = {
				explicit: e.defaultYaml.explicit,
				version: "1.2"
			}, this.tags = Object.assign({}, e.defaultTags);
		}
		return t;
	}
	add(t, n) {
		this.atNextDocument &&= (this.yaml = {
			explicit: e.defaultYaml.explicit,
			version: "1.1"
		}, this.tags = Object.assign({}, e.defaultTags), !1);
		let r = t.trim().split(/[ \t]+/), i = r.shift();
		switch (i) {
			case "%TAG": {
				if (r.length !== 2 && (n(0, "%TAG directive should contain exactly two parts"), r.length < 2)) return !1;
				let [e, t] = r;
				return this.tags[e] = t, !0;
			}
			case "%YAML": {
				if (this.yaml.explicit = !0, r.length !== 1) return n(0, "%YAML directive should contain exactly one part"), !1;
				let [e] = r;
				if (e === "1.1" || e === "1.2") return this.yaml.version = e, !0;
				{
					let t = /^\d+\.\d+$/.test(e);
					return n(6, `Unsupported YAML version ${e}`, t), !1;
				}
			}
			default: return n(0, `Unknown directive ${i}`, !0), !1;
		}
	}
	tagName(e, t) {
		if (e === "!") return "!";
		if (e[0] !== "!") return t(`Not a valid tag: ${e}`), null;
		if (e[1] === "<") {
			let n = e.slice(2, -1);
			return n === "!" || n === "!!" ? (t(`Verbatim tags aren't resolved, so ${e} is invalid.`), null) : (e[e.length - 1] !== ">" && t("Verbatim tags must end with a >"), n);
		}
		let [, n, r] = e.match(/^(.*!)([^!]*)$/s);
		r || t(`The ${e} tag has no suffix`);
		let i = this.tags[n];
		if (i) try {
			return i + decodeURIComponent(r);
		} catch (e) {
			return t(String(e)), null;
		}
		return n === "!" ? e : (t(`Could not resolve tag: ${e}`), null);
	}
	tagString(e) {
		for (let [t, n] of Object.entries(this.tags)) if (e.startsWith(n)) return t + Us(e.substring(n.length));
		return e[0] === "!" ? e : `!<${e}>`;
	}
	toString(e) {
		let t = this.yaml.explicit ? [`%YAML ${this.yaml.version || "1.2"}`] : [], n = Object.entries(this.tags), r;
		if (e && n.length > 0 && Ms(e.contents)) {
			let t = {};
			Ls(e.contents, (e, n) => {
				Ms(n) && n.tag && (t[n.tag] = !0);
			}), r = Object.keys(t);
		} else r = [];
		for (let [i, a] of n) i === "!!" && a === "tag:yaml.org,2002:" || e && !r.some((e) => e.startsWith(a)) || t.push(`%TAG ${i} ${a}`);
		return t.join("\n");
	}
};
Ws.defaultYaml = {
	explicit: !1,
	version: "1.2"
}, Ws.defaultTags = { "!!": "tag:yaml.org,2002:" };
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/doc/anchors.js
function Gs(e) {
	if (/[\x00-\x19\s,[\]{}]/.test(e)) {
		let t = `Anchor must not contain whitespace or control characters: ${JSON.stringify(e)}`;
		throw Error(t);
	}
	return !0;
}
function Ks(e) {
	let t = /* @__PURE__ */ new Set();
	return Ls(e, { Value(e, n) {
		n.anchor && t.add(n.anchor);
	} }), t;
}
function qs(e, t) {
	for (let n = 1;; ++n) {
		let r = `${e}${n}`;
		if (!t.has(r)) return r;
	}
}
function Js(e, t) {
	let n = [], r = /* @__PURE__ */ new Map(), i = null;
	return {
		onAnchor: (r) => {
			n.push(r), i ??= Ks(e);
			let a = qs(t, i);
			return i.add(a), a;
		},
		setAnchors: () => {
			for (let e of n) {
				let t = r.get(e);
				if (typeof t == "object" && t.anchor && (ks(t.node) || js(t.node))) t.node.anchor = t.anchor;
				else {
					let t = /* @__PURE__ */ Error("Failed to resolve repeated object (this should not happen)");
					throw t.source = e, t;
				}
			}
		},
		sourceObjects: r
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/doc/applyReviver.js
function Ys(e, t, n, r) {
	if (r && typeof r == "object") {
		if (Array.isArray(r)) for (let t = 0, n = r.length; t < n; ++t) {
			let n = r[t], i = Ys(e, r, String(t), n);
			i === void 0 ? delete r[t] : i !== n && (r[t] = i);
		}
		else if (r instanceof Map) for (let t of Array.from(r.keys())) {
			let n = r.get(t), i = Ys(e, r, t, n);
			i === void 0 ? r.delete(t) : i !== n && r.set(t, i);
		}
		else if (r instanceof Set) for (let t of Array.from(r)) {
			let n = Ys(e, r, t, t);
			n === void 0 ? r.delete(t) : n !== t && (r.delete(t), r.add(n));
		}
		else for (let [t, n] of Object.entries(r)) {
			let i = Ys(e, r, t, n);
			i === void 0 ? delete r[t] : i !== n && (r[t] = i);
		}
	}
	return e.call(t, n, r);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/nodes/toJS.js
function Xs(e, t, n) {
	if (Array.isArray(e)) return e.map((e, t) => Xs(e, String(t), n));
	if (e && typeof e.toJSON == "function") {
		if (!n || !Ns(e)) return e.toJSON(t, n);
		let r = {
			aliasCount: 0,
			count: 1,
			res: void 0
		};
		n.anchors.set(e, r), n.onCreate = (e) => {
			r.res = e, delete n.onCreate;
		};
		let i = e.toJSON(t, n);
		return n.onCreate && n.onCreate(i), i;
	}
	return typeof e == "bigint" && !n?.keep ? Number(e) : e;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/nodes/Node.js
var Zs = class {
	constructor(e) {
		Object.defineProperty(this, ws, { value: e });
	}
	clone() {
		let e = Object.create(Object.getPrototypeOf(this), Object.getOwnPropertyDescriptors(this));
		return this.range && (e.range = this.range.slice()), e;
	}
	toJS(e, { mapAsMap: t, maxAliasCount: n, onAnchor: r, reviver: i } = {}) {
		if (!Es(e)) throw TypeError("A document argument is required");
		let a = {
			anchors: /* @__PURE__ */ new Map(),
			doc: e,
			keep: !0,
			mapAsMap: t === !0,
			mapKeyWarned: !1,
			maxAliasCount: typeof n == "number" ? n : 100
		}, o = Xs(this, "", a);
		if (typeof r == "function") for (let { count: e, res: t } of a.anchors.values()) r(t, e);
		return typeof i == "function" ? Ys(i, { "": o }, "", o) : o;
	}
}, Qs = class extends Zs {
	constructor(e) {
		super(vs), this.source = e, Object.defineProperty(this, "tag", { set() {
			throw Error("Alias nodes cannot have tags");
		} });
	}
	resolve(e, t) {
		if (t?.maxAliasCount === 0) throw ReferenceError("Alias resolution is disabled");
		let n;
		t?.aliasResolveCache ? n = t.aliasResolveCache : (n = [], Ls(e, { Node: (e, t) => {
			(Ts(t) || Ns(t)) && n.push(t);
		} }), t && (t.aliasResolveCache = n));
		let r;
		for (let e of n) {
			if (e === this) break;
			e.anchor === this.source && (r = e);
		}
		if (r && t) {
			let { anchors: e, doc: n, maxAliasCount: i } = t, a = e.get(r);
			/* istanbul ignore if */
			if (a ||= (Xs(r, null, t), e.get(r)), a?.res === void 0) throw ReferenceError("This should not happen: Alias anchor was not resolved?");
			if (i >= 0 && (a.count += 1, a.aliasCount === 0 && (a.aliasCount = $s(n, r, e)), a.count * a.aliasCount > i)) throw ReferenceError("Excessive alias count indicates a resource exhaustion attack");
		}
		return r;
	}
	toJSON(e, t) {
		if (!t) return { source: this.source };
		let n = this.resolve(t.doc, t);
		if (!n) {
			let e = `Unresolved alias (the anchor must be set before the alias): ${this.source}`;
			throw ReferenceError(e);
		}
		return t.anchors.get(n).res;
	}
	toString(e, t, n) {
		let r = `*${this.source}`;
		if (e) {
			if (Gs(this.source), e.options.verifyAliasOrder && !e.anchors.has(this.source)) {
				let e = `Unresolved alias (the anchor must be set before the alias): ${this.source}`;
				throw Error(e);
			}
			if (e.implicitKey) return `${r} `;
		}
		return r;
	}
};
function $s(e, t, n) {
	if (Ts(t)) {
		let r = t.resolve(e), i = n && r && n.get(r);
		return i ? i.count * i.aliasCount : 0;
	}
	if (js(t)) {
		let r = 0;
		for (let i of t.items) {
			let t = $s(e, i, n);
			t > r && (r = t);
		}
		return r;
	}
	if (Os(t)) {
		let r = $s(e, t.key, n), i = $s(e, t.value, n);
		return Math.max(r, i);
	}
	return 1;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/nodes/Scalar.js
var ec = (e) => !e || typeof e != "function" && typeof e != "object", Z = class extends Zs {
	constructor(e) {
		super(Ss), this.value = e;
	}
	toJSON(e, t) {
		return t?.keep ? this.value : Xs(this.value, e, t);
	}
	toString() {
		return String(this.value);
	}
};
Z.BLOCK_FOLDED = "BLOCK_FOLDED", Z.BLOCK_LITERAL = "BLOCK_LITERAL", Z.PLAIN = "PLAIN", Z.QUOTE_DOUBLE = "QUOTE_DOUBLE", Z.QUOTE_SINGLE = "QUOTE_SINGLE";
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/doc/createNode.js
var tc = "tag:yaml.org,2002:";
function nc(e, t, n) {
	if (t) {
		let e = n.filter((e) => e.tag === t), r = e.find((e) => !e.format) ?? e[0];
		if (!r) throw Error(`Tag ${t} not found`);
		return r;
	}
	return n.find((t) => t.identify?.(e) && !t.format);
}
function rc(e, t, n) {
	if (Es(e) && (e = e.contents), Ms(e)) return e;
	if (Os(e)) {
		let t = n.schema[bs].createNode?.(n.schema, null, n);
		return t.items.push(e), t;
	}
	(e instanceof String || e instanceof Number || e instanceof Boolean || typeof BigInt < "u" && e instanceof BigInt) && (e = e.valueOf());
	let { aliasDuplicateObjects: r, onAnchor: i, onTagObj: a, schema: o, sourceObjects: s } = n, c;
	if (r && e && typeof e == "object") {
		if (c = s.get(e), c) return c.anchor ?? (c.anchor = i(e)), new Qs(c.anchor);
		c = {
			anchor: null,
			node: null
		}, s.set(e, c);
	}
	t?.startsWith("!!") && (t = tc + t.slice(2));
	let l = nc(e, t, o.tags);
	if (!l) {
		if (e && typeof e.toJSON == "function" && (e = e.toJSON()), !e || typeof e != "object") {
			let t = new Z(e);
			return c && (c.node = t), t;
		}
		l = e instanceof Map ? o[bs] : Symbol.iterator in Object(e) ? o[Cs] : o[bs];
	}
	a && (a(l), delete n.onTagObj);
	let u = l?.createNode ? l.createNode(n.schema, e, n) : typeof l?.nodeClass?.from == "function" ? l.nodeClass.from(n.schema, e, n) : new Z(e);
	return t ? u.tag = t : l.default || (u.tag = l.tag), c && (c.node = u), u;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/nodes/Collection.js
function ic(e, t, n) {
	let r = n;
	for (let e = t.length - 1; e >= 0; --e) {
		let n = t[e];
		if (typeof n == "number" && Number.isInteger(n) && n >= 0) {
			let e = [];
			e[n] = r, r = e;
		} else r = /* @__PURE__ */ new Map([[n, r]]);
	}
	return rc(r, void 0, {
		aliasDuplicateObjects: !1,
		keepUndefined: !1,
		onAnchor: () => {
			throw Error("This should not happen, please report a bug.");
		},
		schema: e,
		sourceObjects: /* @__PURE__ */ new Map()
	});
}
var ac = (e) => e == null || typeof e == "object" && !!e[Symbol.iterator]().next().done, oc = class extends Zs {
	constructor(e, t) {
		super(e), Object.defineProperty(this, "schema", {
			value: t,
			configurable: !0,
			enumerable: !1,
			writable: !0
		});
	}
	clone(e) {
		let t = Object.create(Object.getPrototypeOf(this), Object.getOwnPropertyDescriptors(this));
		return e && (t.schema = e), t.items = t.items.map((t) => Ms(t) || Os(t) ? t.clone(e) : t), this.range && (t.range = this.range.slice()), t;
	}
	addIn(e, t) {
		if (ac(e)) this.add(t);
		else {
			let [n, ...r] = e, i = this.get(n, !0);
			if (js(i)) i.addIn(r, t);
			else if (i === void 0 && this.schema) this.set(n, ic(this.schema, r, t));
			else throw Error(`Expected YAML collection at ${n}. Remaining path: ${r}`);
		}
	}
	deleteIn(e) {
		let [t, ...n] = e;
		if (n.length === 0) return this.delete(t);
		let r = this.get(t, !0);
		if (js(r)) return r.deleteIn(n);
		throw Error(`Expected YAML collection at ${t}. Remaining path: ${n}`);
	}
	getIn(e, t) {
		let [n, ...r] = e, i = this.get(n, !0);
		return r.length === 0 ? !t && ks(i) ? i.value : i : js(i) ? i.getIn(r, t) : void 0;
	}
	hasAllNullValues(e) {
		return this.items.every((t) => {
			if (!Os(t)) return !1;
			let n = t.value;
			return n == null || e && ks(n) && n.value == null && !n.commentBefore && !n.comment && !n.tag;
		});
	}
	hasIn(e) {
		let [t, ...n] = e;
		if (n.length === 0) return this.has(t);
		let r = this.get(t, !0);
		return js(r) ? r.hasIn(n) : !1;
	}
	setIn(e, t) {
		let [n, ...r] = e;
		if (r.length === 0) this.set(n, t);
		else {
			let e = this.get(n, !0);
			if (js(e)) e.setIn(r, t);
			else if (e === void 0 && this.schema) this.set(n, ic(this.schema, r, t));
			else throw Error(`Expected YAML collection at ${n}. Remaining path: ${r}`);
		}
	}
}, sc = (e) => e.replace(/^(?!$)(?: $)?/gm, "#");
function cc(e, t) {
	return /^\n+$/.test(e) ? e.substring(1) : t ? e.replace(/^(?! *$)/gm, t) : e;
}
var lc = (e, t, n) => e.endsWith("\n") ? cc(n, t) : n.includes("\n") ? "\n" + cc(n, t) : (e.endsWith(" ") ? "" : " ") + n, uc = "flow", dc = "block", fc = "quoted";
function pc(e, t, n = "flow", { indentAtStart: r, lineWidth: i = 80, minContentWidth: a = 20, onFold: o, onOverflow: s } = {}) {
	if (!i || i < 0) return e;
	i < a && (a = 0);
	let c = Math.max(1 + a, 1 + i - t.length);
	if (e.length <= c) return e;
	let l = [], u = {}, d = i - t.length;
	typeof r == "number" && (r > i - Math.max(2, a) ? l.push(0) : d = i - r);
	let f, p, m = !1, h = -1, g = -1, _ = -1;
	n === "block" && (h = mc(e, h, t.length), h !== -1 && (d = h + c));
	for (let r; r = e[h += 1];) {
		if (n === "quoted" && r === "\\") {
			switch (g = h, e[h + 1]) {
				case "x":
					h += 3;
					break;
				case "u":
					h += 5;
					break;
				case "U":
					h += 9;
					break;
				default: h += 1;
			}
			_ = h;
		}
		if (r === "\n") n === "block" && (h = mc(e, h, t.length)), d = h + t.length + c, f = void 0;
		else {
			if (r === " " && p && p !== " " && p !== "\n" && p !== "	") {
				let t = e[h + 1];
				t && t !== " " && t !== "\n" && t !== "	" && (f = h);
			}
			if (h >= d) {
				if (f) l.push(f), d = f + c, f = void 0;
				else if (n === "quoted") {
					for (; p === " " || p === "	";) p = r, r = e[h += 1], m = !0;
					let t = h > _ + 1 ? h - 2 : g - 1;
					if (u[t]) return e;
					l.push(t), u[t] = !0, d = t + c, f = void 0;
				} else m = !0;
			}
		}
		p = r;
	}
	if (m && s && s(), l.length === 0) return e;
	o && o();
	let v = e.slice(0, l[0]);
	for (let r = 0; r < l.length; ++r) {
		let i = l[r], a = l[r + 1] || e.length;
		i === 0 ? v = `\n${t}${e.slice(0, a)}` : (n === "quoted" && u[i] && (v += `${e[i]}\\`), v += `\n${t}${e.slice(i + 1, a)}`);
	}
	return v;
}
function mc(e, t, n) {
	let r = t, i = t + 1, a = e[i];
	for (; a === " " || a === "	";) if (t < i + n) a = e[++t];
	else {
		do
			a = e[++t];
		while (a && a !== "\n");
		r = t, i = t + 1, a = e[i];
	}
	return r;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/stringify/stringifyString.js
var hc = (e, t) => ({
	indentAtStart: t ? e.indent.length : e.indentAtStart,
	lineWidth: e.options.lineWidth,
	minContentWidth: e.options.minContentWidth
}), gc = (e) => /^(%|---|\.\.\.)/m.test(e);
function _c(e, t, n) {
	if (!t || t < 0) return !1;
	let r = t - n, i = e.length;
	if (i <= r) return !1;
	for (let t = 0, n = 0; t < i; ++t) if (e[t] === "\n") {
		if (t - n > r) return !0;
		if (n = t + 1, i - n <= r) return !1;
	}
	return !0;
}
function vc(e, t) {
	let n = JSON.stringify(e);
	if (t.options.doubleQuotedAsJSON) return n;
	let { implicitKey: r } = t, i = t.options.doubleQuotedMinMultiLineLength, a = t.indent || (gc(e) ? "  " : ""), o = "", s = 0;
	for (let e = 0, t = n[e]; t; t = n[++e]) if (t === " " && n[e + 1] === "\\" && n[e + 2] === "n" && (o += n.slice(s, e) + "\\ ", e += 1, s = e, t = "\\"), t === "\\") switch (n[e + 1]) {
		case "u":
			{
				o += n.slice(s, e);
				let t = n.substr(e + 2, 4);
				switch (t) {
					case "0000":
						o += "\\0";
						break;
					case "0007":
						o += "\\a";
						break;
					case "000b":
						o += "\\v";
						break;
					case "001b":
						o += "\\e";
						break;
					case "0085":
						o += "\\N";
						break;
					case "00a0":
						o += "\\_";
						break;
					case "2028":
						o += "\\L";
						break;
					case "2029":
						o += "\\P";
						break;
					default: t.substr(0, 2) === "00" ? o += "\\x" + t.substr(2) : o += n.substr(e, 6);
				}
				e += 5, s = e + 1;
			}
			break;
		case "n":
			if (r || n[e + 2] === "\"" || n.length < i) e += 1;
			else {
				for (o += n.slice(s, e) + "\n\n"; n[e + 2] === "\\" && n[e + 3] === "n" && n[e + 4] !== "\"";) o += "\n", e += 2;
				o += a, n[e + 2] === " " && (o += "\\"), e += 1, s = e + 1;
			}
			break;
		default: e += 1;
	}
	return o = s ? o + n.slice(s) : n, r ? o : pc(o, a, fc, hc(t, !1));
}
function yc(e, t) {
	if (t.options.singleQuote === !1 || t.implicitKey && e.includes("\n") || /[ \t]\n|\n[ \t]/.test(e)) return vc(e, t);
	let n = t.indent || (gc(e) ? "  " : ""), r = "'" + e.replace(/'/g, "''").replace(/\n+/g, `$&\n${n}`) + "'";
	return t.implicitKey ? r : pc(r, n, uc, hc(t, !1));
}
function bc(e, t) {
	let { singleQuote: n } = t.options, r;
	if (n === !1) r = vc;
	else {
		let t = e.includes("\""), i = e.includes("'");
		r = t && !i ? yc : i && !t ? vc : n ? yc : vc;
	}
	return r(e, t);
}
var xc;
try {
	xc = /* @__PURE__ */ RegExp("(^|(?<!\n))\n+(?!\n|$)", "g");
} catch {
	xc = /\n+(?!\n|$)/g;
}
function Sc({ comment: e, type: t, value: n }, r, i, a) {
	let { blockQuote: o, commentString: s, lineWidth: c } = r.options;
	if (!o || /\n[\t ]+$/.test(n)) return bc(n, r);
	let l = r.indent || (r.forceBlockIndent || gc(n) ? "  " : ""), u = o === "literal" ? !0 : o === "folded" || t === Z.BLOCK_FOLDED ? !1 : t === Z.BLOCK_LITERAL || !_c(n, c, l.length);
	if (!n) return u ? "|\n" : ">\n";
	let d, f;
	for (f = n.length; f > 0; --f) {
		let e = n[f - 1];
		if (e !== "\n" && e !== "	" && e !== " ") break;
	}
	let p = n.substring(f), m = p.indexOf("\n");
	m === -1 ? d = "-" : n === p || m !== p.length - 1 ? (d = "+", a && a()) : d = "", p &&= (n = n.slice(0, -p.length), p[p.length - 1] === "\n" && (p = p.slice(0, -1)), p.replace(xc, `$&${l}`));
	let h = !1, g, _ = -1;
	for (g = 0; g < n.length; ++g) {
		let e = n[g];
		if (e === " ") h = !0;
		else if (e === "\n") _ = g;
		else break;
	}
	let v = n.substring(0, _ < g ? _ + 1 : g);
	v &&= (n = n.substring(v.length), v.replace(/\n+/g, `$&${l}`));
	let y = (h ? l ? "2" : "1" : "") + d;
	if (e && (y += " " + s(e.replace(/ ?[\r\n]+/g, " ")), i && i()), !u) {
		let e = n.replace(/\n+/g, "\n$&").replace(/(?:^|\n)([\t ].*)(?:([\n\t ]*)\n(?![\n\t ]))?/g, "$1$2").replace(/\n+/g, `$&${l}`), i = !1, a = hc(r, !0);
		o !== "folded" && t !== Z.BLOCK_FOLDED && (a.onOverflow = () => {
			i = !0;
		});
		let s = pc(`${v}${e}${p}`, l, dc, a);
		if (!i) return `>${y}\n${l}${s}`;
	}
	return n = n.replace(/\n+/g, `$&${l}`), `|${y}\n${l}${v}${n}${p}`;
}
function Cc(e, t, n, r) {
	let { type: i, value: a } = e, { actualString: o, implicitKey: s, indent: c, indentStep: l, inFlow: u } = t;
	if (s && a.includes("\n") || u && /[[\]{},]/.test(a)) return bc(a, t);
	if (/^[\n\t ,[\]{}#&*!|>'"%@`]|^[?-]$|^[?-][ \t]|[\n:][ \t]|[ \t]\n|[\n\t ]#|[\n\t :]$/.test(a)) return s || u || !a.includes("\n") ? bc(a, t) : Sc(e, t, n, r);
	if (!s && !u && i !== Z.PLAIN && a.includes("\n")) return Sc(e, t, n, r);
	if (gc(a)) {
		if (c === "") return t.forceBlockIndent = !0, Sc(e, t, n, r);
		if (s && c === l) return bc(a, t);
	}
	let d = a.replace(/\n+/g, `$&\n${c}`);
	if (o) {
		let e = (e) => e.default && e.tag !== "tag:yaml.org,2002:str" && e.test?.test(d), { compat: n, tags: r } = t.doc.schema;
		if (r.some(e) || n?.some(e)) return bc(a, t);
	}
	return s ? d : pc(d, c, uc, hc(t, !1));
}
function wc(e, t, n, r) {
	let { implicitKey: i, inFlow: a } = t, o = typeof e.value == "string" ? e : Object.assign({}, e, { value: String(e.value) }), { type: s } = e;
	s !== Z.QUOTE_DOUBLE && /[\x00-\x08\x0b-\x1f\x7f-\x9f\u{D800}-\u{DFFF}]/u.test(o.value) && (s = Z.QUOTE_DOUBLE);
	let c = (e) => {
		switch (e) {
			case Z.BLOCK_FOLDED:
			case Z.BLOCK_LITERAL: return i || a ? bc(o.value, t) : Sc(o, t, n, r);
			case Z.QUOTE_DOUBLE: return vc(o.value, t);
			case Z.QUOTE_SINGLE: return yc(o.value, t);
			case Z.PLAIN: return Cc(o, t, n, r);
			default: return null;
		}
	}, l = c(s);
	if (l === null) {
		let { defaultKeyType: e, defaultStringType: n } = t.options, r = i && e || n;
		if (l = c(r), l === null) throw Error(`Unsupported default string type ${r}`);
	}
	return l;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/stringify/stringify.js
function Tc(e, t) {
	let n = Object.assign({
		blockQuote: !0,
		commentString: sc,
		defaultKeyType: null,
		defaultStringType: "PLAIN",
		directives: null,
		doubleQuotedAsJSON: !1,
		doubleQuotedMinMultiLineLength: 40,
		falseStr: "false",
		flowCollectionPadding: !0,
		indentSeq: !0,
		lineWidth: 80,
		minContentWidth: 20,
		nullStr: "null",
		simpleKeys: !1,
		singleQuote: null,
		trailingComma: !1,
		trueStr: "true",
		verifyAliasOrder: !0
	}, e.schema.toStringOptions, t), r;
	switch (n.collectionStyle) {
		case "block":
			r = !1;
			break;
		case "flow":
			r = !0;
			break;
		default: r = null;
	}
	return {
		anchors: /* @__PURE__ */ new Set(),
		doc: e,
		flowCollectionPadding: n.flowCollectionPadding ? " " : "",
		indent: "",
		indentStep: typeof n.indent == "number" ? " ".repeat(n.indent) : "  ",
		inFlow: r,
		options: n
	};
}
function Ec(e, t) {
	if (t.tag) {
		let n = e.filter((e) => e.tag === t.tag);
		if (n.length > 0) return n.find((e) => e.format === t.format) ?? n[0];
	}
	let n, r;
	if (ks(t)) {
		r = t.value;
		let i = e.filter((e) => e.identify?.(r));
		if (i.length > 1) {
			let e = i.filter((e) => e.test);
			e.length > 0 && (i = e);
		}
		n = i.find((e) => e.format === t.format) ?? i.find((e) => !e.format);
	} else r = t, n = e.find((e) => e.nodeClass && r instanceof e.nodeClass);
	if (!n) {
		let e = r?.constructor?.name ?? (r === null ? "null" : typeof r);
		throw Error(`Tag not resolved for ${e} value`);
	}
	return n;
}
function Dc(e, t, { anchors: n, doc: r }) {
	if (!r.directives) return "";
	let i = [], a = (ks(e) || js(e)) && e.anchor;
	a && Gs(a) && (n.add(a), i.push(`&${a}`));
	let o = e.tag ?? (t.default ? null : t.tag);
	return o && i.push(r.directives.tagString(o)), i.join(" ");
}
function Oc(e, t, n, r) {
	if (Os(e)) return e.toString(t, n, r);
	if (Ts(e)) {
		if (t.doc.directives) return e.toString(t);
		if (t.resolvedAliases?.has(e)) throw TypeError("Cannot stringify circular structure without alias nodes");
		t.resolvedAliases ? t.resolvedAliases.add(e) : t.resolvedAliases = /* @__PURE__ */ new Set([e]), e = e.resolve(t.doc);
	}
	let i, a = Ms(e) ? e : t.doc.createNode(e, { onTagObj: (e) => i = e });
	i ??= Ec(t.doc.schema.tags, a);
	let o = Dc(a, i, t);
	o.length > 0 && (t.indentAtStart = (t.indentAtStart ?? 0) + o.length + 1);
	let s = typeof i.stringify == "function" ? i.stringify(a, t, n, r) : ks(a) ? wc(a, t, n, r) : a.toString(t, n, r);
	return o ? ks(a) || s[0] === "{" || s[0] === "[" ? `${o} ${s}` : `${o}\n${t.indent}${s}` : s;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/stringify/stringifyPair.js
function kc({ key: e, value: t }, n, r, i) {
	let { allNullValues: a, doc: o, indent: s, indentStep: c, options: { commentString: l, indentSeq: u, simpleKeys: d } } = n, f = Ms(e) && e.comment || null;
	if (d) {
		if (f) throw Error("With simple keys, key nodes cannot have comments");
		if (js(e) || !Ms(e) && typeof e == "object") throw Error("With simple keys, collection cannot be used as a key value");
	}
	let p = !d && (!e || f && t == null && !n.inFlow || js(e) || (ks(e) ? e.type === Z.BLOCK_FOLDED || e.type === Z.BLOCK_LITERAL : typeof e == "object"));
	n = Object.assign({}, n, {
		allNullValues: !1,
		implicitKey: !p && (d || !a),
		indent: s + c
	});
	let m = !1, h = !1, g = Oc(e, n, () => m = !0, () => h = !0);
	if (!p && !n.inFlow && g.length > 1024) {
		if (d) throw Error("With simple keys, single line scalar must not span more than 1024 characters");
		p = !0;
	}
	if (n.inFlow) {
		if (a || t == null) return m && r && r(), g === "" ? "?" : p ? `? ${g}` : g;
	} else if (a && !d || t == null && p) return g = `? ${g}`, f && !m ? g += lc(g, n.indent, l(f)) : h && i && i(), g;
	m && (f = null), p ? (f && (g += lc(g, n.indent, l(f))), g = `? ${g}\n${s}:`) : (g = `${g}:`, f && (g += lc(g, n.indent, l(f))));
	let _, v, y;
	Ms(t) ? (_ = !!t.spaceBefore, v = t.commentBefore, y = t.comment) : (_ = !1, v = null, y = null, t && typeof t == "object" && (t = o.createNode(t))), n.implicitKey = !1, !p && !f && ks(t) && (n.indentAtStart = g.length + 1), h = !1, !u && c.length >= 2 && !n.inFlow && !p && As(t) && !t.flow && !t.tag && !t.anchor && (n.indent = n.indent.substring(2));
	let b = !1, x = Oc(t, n, () => b = !0, () => h = !0), S = " ";
	if (f || _ || v) {
		if (S = _ ? "\n" : "", v) {
			let e = l(v);
			S += `\n${cc(e, n.indent)}`;
		}
		x === "" && !n.inFlow ? S === "\n" && y && (S = "\n\n") : S += `\n${n.indent}`;
	} else if (!p && js(t)) {
		let e = x[0], r = x.indexOf("\n"), i = r !== -1, a = n.inFlow ?? t.flow ?? t.items.length === 0;
		if (i || !a) {
			let t = !1;
			if (i && (e === "&" || e === "!")) {
				let n = x.indexOf(" ");
				e === "&" && n !== -1 && n < r && x[n + 1] === "!" && (n = x.indexOf(" ", n + 1)), (n === -1 || r < n) && (t = !0);
			}
			t || (S = `\n${n.indent}`);
		}
	} else (x === "" || x[0] === "\n") && (S = "");
	return g += S + x, n.inFlow ? b && r && r() : y && !b ? g += lc(g, n.indent, l(y)) : h && i && i(), g;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/log.js
function Ac(e, t) {
	(e === "debug" || e === "warn") && console.warn(t);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/schema/yaml-1.1/merge.js
var jc = "<<", Mc = {
	identify: (e) => e === jc || typeof e == "symbol" && e.description === jc,
	default: "key",
	tag: "tag:yaml.org,2002:merge",
	test: /^<<$/,
	resolve: () => Object.assign(new Z(Symbol(jc)), { addToJSMap: Pc }),
	stringify: () => jc
}, Nc = (e, t) => (Mc.identify(t) || ks(t) && (!t.type || t.type === Z.PLAIN) && Mc.identify(t.value)) && e?.doc.schema.tags.some((e) => e.tag === Mc.tag && e.default);
function Pc(e, t, n) {
	let r = Ic(e, n);
	if (As(r)) for (let n of r.items) Fc(e, t, n);
	else if (Array.isArray(r)) for (let n of r) Fc(e, t, n);
	else Fc(e, t, r);
}
function Fc(e, t, n) {
	let r = Ic(e, n);
	if (!Ds(r)) throw Error("Merge sources must be maps or map aliases");
	let i = r.toJSON(null, e, Map);
	for (let [e, n] of i) t instanceof Map ? t.has(e) || t.set(e, n) : t instanceof Set ? t.add(e) : Object.prototype.hasOwnProperty.call(t, e) || Object.defineProperty(t, e, {
		value: n,
		writable: !0,
		enumerable: !0,
		configurable: !0
	});
	return t;
}
function Ic(e, t) {
	return e && Ts(t) ? t.resolve(e.doc, e) : t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/nodes/addPairToJSMap.js
function Lc(e, t, { key: n, value: r }) {
	if (Ms(n) && n.addToJSMap) n.addToJSMap(e, t, r);
	else if (Nc(e, n)) Pc(e, t, r);
	else {
		let i = Xs(n, "", e);
		if (t instanceof Map) t.set(i, Xs(r, i, e));
		else if (t instanceof Set) t.add(i);
		else {
			let a = Rc(n, i, e), o = Xs(r, a, e);
			a in t ? Object.defineProperty(t, a, {
				value: o,
				writable: !0,
				enumerable: !0,
				configurable: !0
			}) : t[a] = o;
		}
	}
	return t;
}
function Rc(e, t, n) {
	if (t === null) return "";
	if (typeof t != "object") return String(t);
	if (Ms(e) && n?.doc) {
		let t = Tc(n.doc, {});
		t.anchors = /* @__PURE__ */ new Set();
		for (let e of n.anchors.keys()) t.anchors.add(e.anchor);
		t.inFlow = !0, t.inStringifyKey = !0;
		let r = e.toString(t);
		if (!n.mapKeyWarned) {
			let e = JSON.stringify(r);
			e.length > 40 && (e = e.substring(0, 36) + "...\""), Ac(n.doc.options.logLevel, `Keys with collection values will be stringified due to JS Object restrictions: ${e}. Set mapAsMap: true to use object keys.`), n.mapKeyWarned = !0;
		}
		return r;
	}
	return JSON.stringify(t);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/nodes/Pair.js
function zc(e, t, n) {
	return new Bc(rc(e, void 0, n), rc(t, void 0, n));
}
var Bc = class e {
	constructor(e, t = null) {
		Object.defineProperty(this, ws, { value: xs }), this.key = e, this.value = t;
	}
	clone(t) {
		let { key: n, value: r } = this;
		return Ms(n) && (n = n.clone(t)), Ms(r) && (r = r.clone(t)), new e(n, r);
	}
	toJSON(e, t) {
		return Lc(t, t?.mapAsMap ? /* @__PURE__ */ new Map() : {}, this);
	}
	toString(e, t, n) {
		return e?.doc ? kc(this, e, t, n) : JSON.stringify(this);
	}
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/stringify/stringifyCollection.js
function Vc(e, t, n) {
	return (t.inFlow ?? e.flow ? Uc : Hc)(e, t, n);
}
function Hc({ comment: e, items: t }, n, { blockItemPrefix: r, flowChars: i, itemIndent: a, onChompKeep: o, onComment: s }) {
	let { indent: c, options: { commentString: l } } = n, u = Object.assign({}, n, {
		indent: a,
		type: null
	}), d = !1, f = [];
	for (let e = 0; e < t.length; ++e) {
		let i = t[e], o = null;
		if (Ms(i)) !d && i.spaceBefore && f.push(""), Wc(n, f, i.commentBefore, d), i.comment && (o = i.comment);
		else if (Os(i)) {
			let e = Ms(i.key) ? i.key : null;
			e && (!d && e.spaceBefore && f.push(""), Wc(n, f, e.commentBefore, d));
		}
		d = !1;
		let s = Oc(i, u, () => o = null, () => d = !0);
		o && (s += lc(s, a, l(o))), d && o && (d = !1), f.push(r + s);
	}
	let p;
	if (f.length === 0) p = i.start + i.end;
	else {
		p = f[0];
		for (let e = 1; e < f.length; ++e) {
			let t = f[e];
			p += t ? `\n${c}${t}` : "\n";
		}
	}
	return e ? (p += "\n" + cc(l(e), c), s && s()) : d && o && o(), p;
}
function Uc({ items: e }, t, { flowChars: n, itemIndent: r }) {
	let { indent: i, indentStep: a, flowCollectionPadding: o, options: { commentString: s } } = t;
	r += a;
	let c = Object.assign({}, t, {
		indent: r,
		inFlow: !0,
		type: null
	}), l = !1, u = 0, d = [];
	for (let n = 0; n < e.length; ++n) {
		let i = e[n], a = null;
		if (Ms(i)) i.spaceBefore && d.push(""), Wc(t, d, i.commentBefore, !1), i.comment && (a = i.comment);
		else if (Os(i)) {
			let e = Ms(i.key) ? i.key : null;
			e && (e.spaceBefore && d.push(""), Wc(t, d, e.commentBefore, !1), e.comment && (l = !0));
			let n = Ms(i.value) ? i.value : null;
			n ? (n.comment && (a = n.comment), n.commentBefore && (l = !0)) : i.value == null && e?.comment && (a = e.comment);
		}
		a && (l = !0);
		let o = Oc(i, c, () => a = null);
		l ||= d.length > u || o.includes("\n"), n < e.length - 1 ? o += "," : t.options.trailingComma && (t.options.lineWidth > 0 && (l ||= d.reduce((e, t) => e + t.length + 2, 2) + (o.length + 2) > t.options.lineWidth), l && (o += ",")), a && (o += lc(o, r, s(a))), d.push(o), u = d.length;
	}
	let { start: f, end: p } = n;
	if (d.length === 0) return f + p;
	if (!l) {
		let e = d.reduce((e, t) => e + t.length + 2, 2);
		l = t.options.lineWidth > 0 && e > t.options.lineWidth;
	}
	if (l) {
		let e = f;
		for (let t of d) e += t ? `\n${a}${i}${t}` : "\n";
		return `${e}\n${i}${p}`;
	}
	return `${f}${o}${d.join(" ")}${o}${p}`;
}
function Wc({ indent: e, options: { commentString: t } }, n, r, i) {
	if (r && i && (r = r.replace(/^\n+/, "")), r) {
		let i = cc(t(r), e);
		n.push(i.trimStart());
	}
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/nodes/YAMLMap.js
function Gc(e, t) {
	let n = ks(t) ? t.value : t;
	for (let r of e) if (Os(r) && (r.key === t || r.key === n || ks(r.key) && r.key.value === n)) return r;
}
var Kc = class extends oc {
	static get tagName() {
		return "tag:yaml.org,2002:map";
	}
	constructor(e) {
		super(bs, e), this.items = [];
	}
	static from(e, t, n) {
		let { keepUndefined: r, replacer: i } = n, a = new this(e), o = (e, o) => {
			if (typeof i == "function") o = i.call(t, e, o);
			else if (Array.isArray(i) && !i.includes(e)) return;
			(o !== void 0 || r) && a.items.push(zc(e, o, n));
		};
		if (t instanceof Map) for (let [e, n] of t) o(e, n);
		else if (t && typeof t == "object") for (let e of Object.keys(t)) o(e, t[e]);
		return typeof e.sortMapEntries == "function" && a.items.sort(e.sortMapEntries), a;
	}
	add(e, t) {
		let n;
		n = Os(e) ? e : !e || typeof e != "object" || !("key" in e) ? new Bc(e, e?.value) : new Bc(e.key, e.value);
		let r = Gc(this.items, n.key), i = this.schema?.sortMapEntries;
		if (r) {
			if (!t) throw Error(`Key ${n.key} already set`);
			ks(r.value) && ec(n.value) ? r.value.value = n.value : r.value = n.value;
		} else if (i) {
			let e = this.items.findIndex((e) => i(n, e) < 0);
			e === -1 ? this.items.push(n) : this.items.splice(e, 0, n);
		} else this.items.push(n);
	}
	delete(e) {
		let t = Gc(this.items, e);
		return t ? this.items.splice(this.items.indexOf(t), 1).length > 0 : !1;
	}
	get(e, t) {
		let n = Gc(this.items, e)?.value;
		return (!t && ks(n) ? n.value : n) ?? void 0;
	}
	has(e) {
		return !!Gc(this.items, e);
	}
	set(e, t) {
		this.add(new Bc(e, t), !0);
	}
	toJSON(e, t, n) {
		let r = n ? new n() : t?.mapAsMap ? /* @__PURE__ */ new Map() : {};
		t?.onCreate && t.onCreate(r);
		for (let e of this.items) Lc(t, r, e);
		return r;
	}
	toString(e, t, n) {
		if (!e) return JSON.stringify(this);
		for (let e of this.items) if (!Os(e)) throw Error(`Map items must all be pairs; found ${JSON.stringify(e)} instead`);
		return !e.allNullValues && this.hasAllNullValues(!1) && (e = Object.assign({}, e, { allNullValues: !0 })), Vc(this, e, {
			blockItemPrefix: "",
			flowChars: {
				start: "{",
				end: "}"
			},
			itemIndent: e.indent || "",
			onChompKeep: n,
			onComment: t
		});
	}
}, qc = {
	collection: "map",
	default: !0,
	nodeClass: Kc,
	tag: "tag:yaml.org,2002:map",
	resolve(e, t) {
		return Ds(e) || t("Expected a mapping for this tag"), e;
	},
	createNode: (e, t, n) => Kc.from(e, t, n)
}, Jc = class extends oc {
	static get tagName() {
		return "tag:yaml.org,2002:seq";
	}
	constructor(e) {
		super(Cs, e), this.items = [];
	}
	add(e) {
		this.items.push(e);
	}
	delete(e) {
		let t = Yc(e);
		return typeof t == "number" && this.items.splice(t, 1).length > 0;
	}
	get(e, t) {
		let n = Yc(e);
		if (typeof n != "number") return;
		let r = this.items[n];
		return !t && ks(r) ? r.value : r;
	}
	has(e) {
		let t = Yc(e);
		return typeof t == "number" && t < this.items.length;
	}
	set(e, t) {
		let n = Yc(e);
		if (typeof n != "number") throw Error(`Expected a valid index, not ${e}.`);
		let r = this.items[n];
		ks(r) && ec(t) ? r.value = t : this.items[n] = t;
	}
	toJSON(e, t) {
		let n = [];
		t?.onCreate && t.onCreate(n);
		let r = 0;
		for (let e of this.items) n.push(Xs(e, String(r++), t));
		return n;
	}
	toString(e, t, n) {
		return e ? Vc(this, e, {
			blockItemPrefix: "- ",
			flowChars: {
				start: "[",
				end: "]"
			},
			itemIndent: (e.indent || "") + "  ",
			onChompKeep: n,
			onComment: t
		}) : JSON.stringify(this);
	}
	static from(e, t, n) {
		let { replacer: r } = n, i = new this(e);
		if (t && Symbol.iterator in Object(t)) {
			let e = 0;
			for (let a of t) {
				if (typeof r == "function") {
					let n = t instanceof Set ? a : String(e++);
					a = r.call(t, n, a);
				}
				i.items.push(rc(a, void 0, n));
			}
		}
		return i;
	}
};
function Yc(e) {
	let t = ks(e) ? e.value : e;
	return t && typeof t == "string" && (t = Number(t)), typeof t == "number" && Number.isInteger(t) && t >= 0 ? t : null;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/schema/common/seq.js
var Xc = {
	collection: "seq",
	default: !0,
	nodeClass: Jc,
	tag: "tag:yaml.org,2002:seq",
	resolve(e, t) {
		return As(e) || t("Expected a sequence for this tag"), e;
	},
	createNode: (e, t, n) => Jc.from(e, t, n)
}, Zc = {
	identify: (e) => typeof e == "string",
	default: !0,
	tag: "tag:yaml.org,2002:str",
	resolve: (e) => e,
	stringify(e, t, n, r) {
		return t = Object.assign({ actualString: !0 }, t), wc(e, t, n, r);
	}
}, Qc = {
	identify: (e) => e == null,
	createNode: () => new Z(null),
	default: !0,
	tag: "tag:yaml.org,2002:null",
	test: /^(?:~|[Nn]ull|NULL)?$/,
	resolve: () => new Z(null),
	stringify: ({ source: e }, t) => typeof e == "string" && Qc.test.test(e) ? e : t.options.nullStr
}, $c = {
	identify: (e) => typeof e == "boolean",
	default: !0,
	tag: "tag:yaml.org,2002:bool",
	test: /^(?:[Tt]rue|TRUE|[Ff]alse|FALSE)$/,
	resolve: (e) => new Z(e[0] === "t" || e[0] === "T"),
	stringify({ source: e, value: t }, n) {
		return e && $c.test.test(e) && t === (e[0] === "t" || e[0] === "T") ? e : t ? n.options.trueStr : n.options.falseStr;
	}
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/stringify/stringifyNumber.js
function el({ format: e, minFractionDigits: t, tag: n, value: r }) {
	if (typeof r == "bigint") return String(r);
	let i = typeof r == "number" ? r : Number(r);
	if (!isFinite(i)) return isNaN(i) ? ".nan" : i < 0 ? "-.inf" : ".inf";
	let a = Object.is(r, -0) ? "-0" : JSON.stringify(r);
	if (!e && t && (!n || n === "tag:yaml.org,2002:float") && /^-?\d/.test(a) && !a.includes("e")) {
		let e = a.indexOf(".");
		e < 0 && (e = a.length, a += ".");
		let n = t - (a.length - e - 1);
		for (; n-- > 0;) a += "0";
	}
	return a;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/schema/core/float.js
var tl = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	test: /^(?:[-+]?\.(?:inf|Inf|INF)|\.nan|\.NaN|\.NAN)$/,
	resolve: (e) => e.slice(-3).toLowerCase() === "nan" ? NaN : e[0] === "-" ? -Infinity : Infinity,
	stringify: el
}, nl = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	format: "EXP",
	test: /^[-+]?(?:\.[0-9]+|[0-9]+(?:\.[0-9]*)?)[eE][-+]?[0-9]+$/,
	resolve: (e) => parseFloat(e),
	stringify(e) {
		let t = Number(e.value);
		return isFinite(t) ? t.toExponential() : el(e);
	}
}, rl = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	test: /^[-+]?(?:\.[0-9]+|[0-9]+\.[0-9]*)$/,
	resolve(e) {
		let t = new Z(parseFloat(e)), n = e.indexOf(".");
		return n !== -1 && e[e.length - 1] === "0" && (t.minFractionDigits = e.length - n - 1), t;
	},
	stringify: el
}, il = (e) => typeof e == "bigint" || Number.isInteger(e), al = (e, t, n, { intAsBigInt: r }) => r ? BigInt(e) : parseInt(e.substring(t), n);
function ol(e, t, n) {
	let { value: r } = e;
	return il(r) && r >= 0 ? n + r.toString(t) : el(e);
}
var sl = {
	identify: (e) => il(e) && e >= 0,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "OCT",
	test: /^0o[0-7]+$/,
	resolve: (e, t, n) => al(e, 2, 8, n),
	stringify: (e) => ol(e, 8, "0o")
}, cl = {
	identify: il,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	test: /^[-+]?[0-9]+$/,
	resolve: (e, t, n) => al(e, 0, 10, n),
	stringify: el
}, ll = {
	identify: (e) => il(e) && e >= 0,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "HEX",
	test: /^0x[0-9a-fA-F]+$/,
	resolve: (e, t, n) => al(e, 2, 16, n),
	stringify: (e) => ol(e, 16, "0x")
}, ul = [
	qc,
	Xc,
	Zc,
	Qc,
	$c,
	sl,
	cl,
	ll,
	tl,
	nl,
	rl
];
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/schema/json/schema.js
function dl(e) {
	return typeof e == "bigint" || Number.isInteger(e);
}
var fl = ({ value: e }) => JSON.stringify(e), pl = [
	{
		identify: (e) => typeof e == "string",
		default: !0,
		tag: "tag:yaml.org,2002:str",
		resolve: (e) => e,
		stringify: fl
	},
	{
		identify: (e) => e == null,
		createNode: () => new Z(null),
		default: !0,
		tag: "tag:yaml.org,2002:null",
		test: /^null$/,
		resolve: () => null,
		stringify: fl
	},
	{
		identify: (e) => typeof e == "boolean",
		default: !0,
		tag: "tag:yaml.org,2002:bool",
		test: /^true$|^false$/,
		resolve: (e) => e === "true",
		stringify: fl
	},
	{
		identify: dl,
		default: !0,
		tag: "tag:yaml.org,2002:int",
		test: /^-?(?:0|[1-9][0-9]*)$/,
		resolve: (e, t, { intAsBigInt: n }) => n ? BigInt(e) : parseInt(e, 10),
		stringify: ({ value: e }) => dl(e) ? e.toString() : JSON.stringify(e)
	},
	{
		identify: (e) => typeof e == "number",
		default: !0,
		tag: "tag:yaml.org,2002:float",
		test: /^-?(?:0|[1-9][0-9]*)(?:\.[0-9]*)?(?:[eE][-+]?[0-9]+)?$/,
		resolve: (e) => parseFloat(e),
		stringify: fl
	}
], ml = [qc, Xc].concat(pl, {
	default: !0,
	tag: "",
	test: /^/,
	resolve(e, t) {
		return t(`Unresolved plain scalar ${JSON.stringify(e)}`), e;
	}
}), hl = {
	identify: (e) => e instanceof Uint8Array,
	default: !1,
	tag: "tag:yaml.org,2002:binary",
	resolve(e, t) {
		if (typeof atob == "function") {
			let t = atob(e.replace(/[\n\r]/g, "")), n = new Uint8Array(t.length);
			for (let e = 0; e < t.length; ++e) n[e] = t.charCodeAt(e);
			return n;
		}
		return t("This environment does not support reading binary tags; either Buffer or atob is required"), e;
	},
	stringify({ comment: e, type: t, value: n }, r, i, a) {
		if (!n) return "";
		let o = n, s;
		if (typeof btoa == "function") {
			let e = "";
			for (let t = 0; t < o.length; ++t) e += String.fromCharCode(o[t]);
			s = btoa(e);
		} else throw Error("This environment does not support writing binary tags; either Buffer or btoa is required");
		if (t ??= Z.BLOCK_LITERAL, t !== Z.QUOTE_DOUBLE) {
			let e = Math.max(r.options.lineWidth - r.indent.length, r.options.minContentWidth), n = Math.ceil(s.length / e), i = Array(n);
			for (let t = 0, r = 0; t < n; ++t, r += e) i[t] = s.substr(r, e);
			s = i.join(t === Z.BLOCK_LITERAL ? "\n" : " ");
		}
		return wc({
			comment: e,
			type: t,
			value: s
		}, r, i, a);
	}
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/schema/yaml-1.1/pairs.js
function gl(e, t) {
	if (As(e)) for (let n = 0; n < e.items.length; ++n) {
		let r = e.items[n];
		if (!Os(r)) {
			if (Ds(r)) {
				r.items.length > 1 && t("Each pair must have its own sequence indicator");
				let e = r.items[0] || new Bc(new Z(null));
				if (r.commentBefore && (e.key.commentBefore = e.key.commentBefore ? `${r.commentBefore}\n${e.key.commentBefore}` : r.commentBefore), r.comment) {
					let t = e.value ?? e.key;
					t.comment = t.comment ? `${r.comment}\n${t.comment}` : r.comment;
				}
				r = e;
			}
			e.items[n] = Os(r) ? r : new Bc(r);
		}
	}
	else t("Expected a sequence for this tag");
	return e;
}
function _l(e, t, n) {
	let { replacer: r } = n, i = new Jc(e);
	i.tag = "tag:yaml.org,2002:pairs";
	let a = 0;
	if (t && Symbol.iterator in Object(t)) for (let e of t) {
		typeof r == "function" && (e = r.call(t, String(a++), e));
		let o, s;
		if (Array.isArray(e)) {
			if (e.length === 2) o = e[0], s = e[1];
			else throw TypeError(`Expected [key, value] tuple: ${e}`);
		} else if (e && e instanceof Object) {
			let t = Object.keys(e);
			if (t.length === 1) o = t[0], s = e[o];
			else throw TypeError(`Expected tuple with one key, not ${t.length} keys`);
		} else o = e;
		i.items.push(zc(o, s, n));
	}
	return i;
}
var vl = {
	collection: "seq",
	default: !1,
	tag: "tag:yaml.org,2002:pairs",
	resolve: gl,
	createNode: _l
}, yl = class e extends Jc {
	constructor() {
		super(), this.add = Kc.prototype.add.bind(this), this.delete = Kc.prototype.delete.bind(this), this.get = Kc.prototype.get.bind(this), this.has = Kc.prototype.has.bind(this), this.set = Kc.prototype.set.bind(this), this.tag = e.tag;
	}
	toJSON(e, t) {
		if (!t) return super.toJSON(e);
		let n = /* @__PURE__ */ new Map();
		t?.onCreate && t.onCreate(n);
		for (let e of this.items) {
			let r, i;
			if (Os(e) ? (r = Xs(e.key, "", t), i = Xs(e.value, r, t)) : r = Xs(e, "", t), n.has(r)) throw Error("Ordered maps must not include duplicate keys");
			n.set(r, i);
		}
		return n;
	}
	static from(e, t, n) {
		let r = _l(e, t, n), i = new this();
		return i.items = r.items, i;
	}
};
yl.tag = "tag:yaml.org,2002:omap";
var bl = {
	collection: "seq",
	identify: (e) => e instanceof Map,
	nodeClass: yl,
	default: !1,
	tag: "tag:yaml.org,2002:omap",
	resolve(e, t) {
		let n = gl(e, t), r = [];
		for (let { key: e } of n.items) ks(e) && (r.includes(e.value) ? t(`Ordered maps must not include duplicate keys: ${e.value}`) : r.push(e.value));
		return Object.assign(new yl(), n);
	},
	createNode: (e, t, n) => yl.from(e, t, n)
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/schema/yaml-1.1/bool.js
function xl({ value: e, source: t }, n) {
	return t && (e ? Sl : Cl).test.test(t) ? t : e ? n.options.trueStr : n.options.falseStr;
}
var Sl = {
	identify: (e) => e === !0,
	default: !0,
	tag: "tag:yaml.org,2002:bool",
	test: /^(?:Y|y|[Yy]es|YES|[Tt]rue|TRUE|[Oo]n|ON)$/,
	resolve: () => new Z(!0),
	stringify: xl
}, Cl = {
	identify: (e) => e === !1,
	default: !0,
	tag: "tag:yaml.org,2002:bool",
	test: /^(?:N|n|[Nn]o|NO|[Ff]alse|FALSE|[Oo]ff|OFF)$/,
	resolve: () => new Z(!1),
	stringify: xl
}, wl = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	test: /^(?:[-+]?\.(?:inf|Inf|INF)|\.nan|\.NaN|\.NAN)$/,
	resolve: (e) => e.slice(-3).toLowerCase() === "nan" ? NaN : e[0] === "-" ? -Infinity : Infinity,
	stringify: el
}, Tl = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	format: "EXP",
	test: /^[-+]?(?:[0-9][0-9_]*)?(?:\.[0-9_]*)?[eE][-+]?[0-9]+$/,
	resolve: (e) => parseFloat(e.replace(/_/g, "")),
	stringify(e) {
		let t = Number(e.value);
		return isFinite(t) ? t.toExponential() : el(e);
	}
}, El = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	test: /^[-+]?(?:[0-9][0-9_]*)?\.[0-9_]*$/,
	resolve(e) {
		let t = new Z(parseFloat(e.replace(/_/g, ""))), n = e.indexOf(".");
		if (n !== -1) {
			let r = e.substring(n + 1).replace(/_/g, "");
			r[r.length - 1] === "0" && (t.minFractionDigits = r.length);
		}
		return t;
	},
	stringify: el
}, Dl = (e) => typeof e == "bigint" || Number.isInteger(e);
function Ol(e, t, n, { intAsBigInt: r }) {
	let i = e[0];
	if ((i === "-" || i === "+") && (t += 1), e = e.substring(t).replace(/_/g, ""), r) {
		switch (n) {
			case 2:
				e = `0b${e}`;
				break;
			case 8:
				e = `0o${e}`;
				break;
			case 16: e = `0x${e}`;
		}
		let t = BigInt(e);
		return i === "-" ? BigInt(-1) * t : t;
	}
	let a = parseInt(e, n);
	return i === "-" ? -1 * a : a;
}
function kl(e, t, n) {
	let { value: r } = e;
	if (Dl(r)) {
		let e = r.toString(t);
		return r < 0 ? "-" + n + e.substr(1) : n + e;
	}
	return el(e);
}
var Al = {
	identify: Dl,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "BIN",
	test: /^[-+]?0b[0-1_]+$/,
	resolve: (e, t, n) => Ol(e, 2, 2, n),
	stringify: (e) => kl(e, 2, "0b")
}, jl = {
	identify: Dl,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "OCT",
	test: /^[-+]?0[0-7_]+$/,
	resolve: (e, t, n) => Ol(e, 1, 8, n),
	stringify: (e) => kl(e, 8, "0")
}, Ml = {
	identify: Dl,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	test: /^[-+]?[0-9][0-9_]*$/,
	resolve: (e, t, n) => Ol(e, 0, 10, n),
	stringify: el
}, Nl = {
	identify: Dl,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "HEX",
	test: /^[-+]?0x[0-9a-fA-F_]+$/,
	resolve: (e, t, n) => Ol(e, 2, 16, n),
	stringify: (e) => kl(e, 16, "0x")
}, Pl = class e extends Kc {
	constructor(t) {
		super(t), this.tag = e.tag;
	}
	add(e) {
		let t;
		t = Os(e) ? e : e && typeof e == "object" && "key" in e && "value" in e && e.value === null ? new Bc(e.key, null) : new Bc(e, null), Gc(this.items, t.key) || this.items.push(t);
	}
	get(e, t) {
		let n = Gc(this.items, e);
		return !t && Os(n) ? ks(n.key) ? n.key.value : n.key : n;
	}
	set(e, t) {
		if (typeof t != "boolean") throw Error(`Expected boolean value for set(key, value) in a YAML set, not ${typeof t}`);
		let n = Gc(this.items, e);
		n && !t ? this.items.splice(this.items.indexOf(n), 1) : !n && t && this.items.push(new Bc(e));
	}
	toJSON(e, t) {
		return super.toJSON(e, t, Set);
	}
	toString(e, t, n) {
		if (!e) return JSON.stringify(this);
		if (this.hasAllNullValues(!0)) return super.toString(Object.assign({}, e, { allNullValues: !0 }), t, n);
		throw Error("Set items must all have null values");
	}
	static from(e, t, n) {
		let { replacer: r } = n, i = new this(e);
		if (t && Symbol.iterator in Object(t)) for (let e of t) typeof r == "function" && (e = r.call(t, e, e)), i.items.push(zc(e, null, n));
		return i;
	}
};
Pl.tag = "tag:yaml.org,2002:set";
var Fl = {
	collection: "map",
	identify: (e) => e instanceof Set,
	nodeClass: Pl,
	default: !1,
	tag: "tag:yaml.org,2002:set",
	createNode: (e, t, n) => Pl.from(e, t, n),
	resolve(e, t) {
		if (Ds(e)) {
			if (e.hasAllNullValues(!0)) return Object.assign(new Pl(), e);
			t("Set items must all have null values");
		} else t("Expected a mapping for this tag");
		return e;
	}
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/schema/yaml-1.1/timestamp.js
function Il(e, t) {
	let n = e[0], r = n === "-" || n === "+" ? e.substring(1) : e, i = (e) => t ? BigInt(e) : Number(e), a = r.replace(/_/g, "").split(":").reduce((e, t) => e * i(60) + i(t), i(0));
	return n === "-" ? i(-1) * a : a;
}
function Ll(e) {
	let { value: t } = e, n = (e) => e;
	if (typeof t == "bigint") n = (e) => BigInt(e);
	else if (isNaN(t) || !isFinite(t)) return el(e);
	let r = "";
	t < 0 && (r = "-", t *= n(-1));
	let i = n(60), a = [t % i];
	return t < 60 ? a.unshift(0) : (t = (t - a[0]) / i, a.unshift(t % i), t >= 60 && (t = (t - a[0]) / i, a.unshift(t))), r + a.map((e) => String(e).padStart(2, "0")).join(":").replace(/000000\d*$/, "");
}
var Rl = {
	identify: (e) => typeof e == "bigint" || Number.isInteger(e),
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "TIME",
	test: /^[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+$/,
	resolve: (e, t, { intAsBigInt: n }) => Il(e, n),
	stringify: Ll
}, zl = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	format: "TIME",
	test: /^[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+\.[0-9_]*$/,
	resolve: (e) => Il(e, !1),
	stringify: Ll
}, Bl = {
	identify: (e) => e instanceof Date,
	default: !0,
	tag: "tag:yaml.org,2002:timestamp",
	test: RegExp("^([0-9]{4})-([0-9]{1,2})-([0-9]{1,2})(?:(?:t|T|[ \\t]+)([0-9]{1,2}):([0-9]{1,2}):([0-9]{1,2}(\\.[0-9]+)?)(?:[ \\t]*(Z|[-+][012]?[0-9](?::[0-9]{2})?))?)?$"),
	resolve(e) {
		let t = e.match(Bl.test);
		if (!t) throw Error("!!timestamp expects a date, starting with yyyy-mm-dd");
		let [, n, r, i, a, o, s] = t.map(Number), c = t[7] ? Number((t[7] + "00").substr(1, 3)) : 0, l = Date.UTC(n, r - 1, i, a || 0, o || 0, s || 0, c), u = t[8];
		if (u && u !== "Z") {
			let e = Il(u, !1);
			Math.abs(e) < 30 && (e *= 60), l -= 6e4 * e;
		}
		return new Date(l);
	},
	stringify: ({ value: e }) => e?.toISOString().replace(/(T00:00:00)?\.000Z$/, "") ?? ""
}, Vl = [
	qc,
	Xc,
	Zc,
	Qc,
	Sl,
	Cl,
	Al,
	jl,
	Ml,
	Nl,
	wl,
	Tl,
	El,
	hl,
	Mc,
	bl,
	vl,
	Fl,
	Rl,
	zl,
	Bl
], Hl = /* @__PURE__ */ new Map([
	["core", ul],
	["failsafe", [
		qc,
		Xc,
		Zc
	]],
	["json", ml],
	["yaml11", Vl],
	["yaml-1.1", Vl]
]), Ul = {
	binary: hl,
	bool: $c,
	float: rl,
	floatExp: nl,
	floatNaN: tl,
	floatTime: zl,
	int: cl,
	intHex: ll,
	intOct: sl,
	intTime: Rl,
	map: qc,
	merge: Mc,
	null: Qc,
	omap: bl,
	pairs: vl,
	seq: Xc,
	set: Fl,
	timestamp: Bl
}, Wl = {
	"tag:yaml.org,2002:binary": hl,
	"tag:yaml.org,2002:merge": Mc,
	"tag:yaml.org,2002:omap": bl,
	"tag:yaml.org,2002:pairs": vl,
	"tag:yaml.org,2002:set": Fl,
	"tag:yaml.org,2002:timestamp": Bl
};
function Gl(e, t, n) {
	let r = Hl.get(t);
	if (r && !e) return n && !r.includes(Mc) ? r.concat(Mc) : r.slice();
	let i = r;
	if (!i) {
		if (Array.isArray(e)) i = [];
		else {
			let e = Array.from(Hl.keys()).filter((e) => e !== "yaml11").map((e) => JSON.stringify(e)).join(", ");
			throw Error(`Unknown schema "${t}"; use one of ${e} or define customTags array`);
		}
	}
	if (Array.isArray(e)) for (let t of e) i = i.concat(t);
	else typeof e == "function" && (i = e(i.slice()));
	return n && (i = i.concat(Mc)), i.reduce((e, t) => {
		let n = typeof t == "string" ? Ul[t] : t;
		if (!n) {
			let e = JSON.stringify(t), n = Object.keys(Ul).map((e) => JSON.stringify(e)).join(", ");
			throw Error(`Unknown custom tag ${e}; use one of ${n}`);
		}
		return e.includes(n) || e.push(n), e;
	}, []);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/schema/Schema.js
var Kl = (e, t) => e.key < t.key ? -1 : +(e.key > t.key), ql = class e {
	constructor({ compat: e, customTags: t, merge: n, resolveKnownTags: r, schema: i, sortMapEntries: a, toStringDefaults: o }) {
		this.compat = Array.isArray(e) ? Gl(e, "compat") : e ? Gl(null, e) : null, this.name = typeof i == "string" && i || "core", this.knownTags = r ? Wl : {}, this.tags = Gl(t, this.name, n), this.toStringOptions = o ?? null, Object.defineProperty(this, bs, { value: qc }), Object.defineProperty(this, Ss, { value: Zc }), Object.defineProperty(this, Cs, { value: Xc }), this.sortMapEntries = typeof a == "function" ? a : a === !0 ? Kl : null;
	}
	clone() {
		let t = Object.create(e.prototype, Object.getOwnPropertyDescriptors(this));
		return t.tags = this.tags.slice(), t;
	}
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/stringify/stringifyDocument.js
function Jl(e, t) {
	let n = [], r = t.directives === !0;
	if (t.directives !== !1 && e.directives) {
		let t = e.directives.toString(e);
		t ? (n.push(t), r = !0) : e.directives.docStart && (r = !0);
	}
	r && n.push("---");
	let i = Tc(e, t), { commentString: a } = i.options;
	if (e.commentBefore) {
		n.length !== 1 && n.unshift("");
		let t = a(e.commentBefore);
		n.unshift(cc(t, ""));
	}
	let o = !1, s = null;
	if (e.contents) {
		if (Ms(e.contents)) {
			if (e.contents.spaceBefore && r && n.push(""), e.contents.commentBefore) {
				let t = a(e.contents.commentBefore);
				n.push(cc(t, ""));
			}
			i.forceBlockIndent = !!e.comment, s = e.contents.comment;
		}
		let t = s ? void 0 : () => o = !0, c = Oc(e.contents, i, () => s = null, t);
		s && (c += lc(c, "", a(s))), (c[0] === "|" || c[0] === ">") && n[n.length - 1] === "---" ? n[n.length - 1] = `--- ${c}` : n.push(c);
	} else n.push(Oc(e.contents, i));
	if (e.directives?.docEnd) {
		if (e.comment) {
			let t = a(e.comment);
			t.includes("\n") ? (n.push("..."), n.push(cc(t, ""))) : n.push(`... ${t}`);
		} else n.push("...");
	} else {
		let t = e.comment;
		t && o && (t = t.replace(/^\n+/, "")), t && ((!o || s) && n[n.length - 1] !== "" && n.push(""), n.push(cc(a(t), "")));
	}
	return n.join("\n") + "\n";
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/doc/Document.js
var Yl = class e {
	constructor(e, t, n) {
		this.commentBefore = null, this.comment = null, this.errors = [], this.warnings = [], Object.defineProperty(this, ws, { value: ys });
		let r = null;
		typeof t == "function" || Array.isArray(t) ? r = t : n === void 0 && t && (n = t, t = void 0);
		let i = Object.assign({
			intAsBigInt: !1,
			keepSourceTokens: !1,
			logLevel: "warn",
			prettyErrors: !0,
			strict: !0,
			stringKeys: !1,
			uniqueKeys: !0,
			version: "1.2"
		}, n);
		this.options = i;
		let { version: a } = i;
		n?._directives ? (this.directives = n._directives.atDocument(), this.directives.yaml.explicit && (a = this.directives.yaml.version)) : this.directives = new Ws({ version: a }), this.setSchema(a, n), this.contents = e === void 0 ? null : this.createNode(e, r, n);
	}
	clone() {
		let t = Object.create(e.prototype, { [ws]: { value: ys } });
		return t.commentBefore = this.commentBefore, t.comment = this.comment, t.errors = this.errors.slice(), t.warnings = this.warnings.slice(), t.options = Object.assign({}, this.options), this.directives && (t.directives = this.directives.clone()), t.schema = this.schema.clone(), t.contents = Ms(this.contents) ? this.contents.clone(t.schema) : this.contents, this.range && (t.range = this.range.slice()), t;
	}
	add(e) {
		Xl(this.contents) && this.contents.add(e);
	}
	addIn(e, t) {
		Xl(this.contents) && this.contents.addIn(e, t);
	}
	createAlias(e, t) {
		if (!e.anchor) {
			let n = Ks(this);
			e.anchor = !t || n.has(t) ? qs(t || "a", n) : t;
		}
		return new Qs(e.anchor);
	}
	createNode(e, t, n) {
		let r;
		if (typeof t == "function") e = t.call({ "": e }, "", e), r = t;
		else if (Array.isArray(t)) {
			let e = t.filter((e) => typeof e == "number" || e instanceof String || e instanceof Number).map(String);
			e.length > 0 && (t = t.concat(e)), r = t;
		} else n === void 0 && t && (n = t, t = void 0);
		let { aliasDuplicateObjects: i, anchorPrefix: a, flow: o, keepUndefined: s, onTagObj: c, tag: l } = n ?? {}, { onAnchor: u, setAnchors: d, sourceObjects: f } = Js(this, a || "a"), p = {
			aliasDuplicateObjects: i ?? !0,
			keepUndefined: s ?? !1,
			onAnchor: u,
			onTagObj: c,
			replacer: r,
			schema: this.schema,
			sourceObjects: f
		}, m = rc(e, l, p);
		return o && js(m) && (m.flow = !0), d(), m;
	}
	createPair(e, t, n = {}) {
		return new Bc(this.createNode(e, null, n), this.createNode(t, null, n));
	}
	delete(e) {
		return Xl(this.contents) ? this.contents.delete(e) : !1;
	}
	deleteIn(e) {
		return ac(e) ? this.contents != null && (this.contents = null, !0) : Xl(this.contents) ? this.contents.deleteIn(e) : !1;
	}
	get(e, t) {
		return js(this.contents) ? this.contents.get(e, t) : void 0;
	}
	getIn(e, t) {
		return ac(e) ? !t && ks(this.contents) ? this.contents.value : this.contents : js(this.contents) ? this.contents.getIn(e, t) : void 0;
	}
	has(e) {
		return js(this.contents) ? this.contents.has(e) : !1;
	}
	hasIn(e) {
		return ac(e) ? this.contents !== void 0 : js(this.contents) ? this.contents.hasIn(e) : !1;
	}
	set(e, t) {
		this.contents == null ? this.contents = ic(this.schema, [e], t) : Xl(this.contents) && this.contents.set(e, t);
	}
	setIn(e, t) {
		ac(e) ? this.contents = t : this.contents == null ? this.contents = ic(this.schema, Array.from(e), t) : Xl(this.contents) && this.contents.setIn(e, t);
	}
	setSchema(e, t = {}) {
		typeof e == "number" && (e = String(e));
		let n;
		switch (e) {
			case "1.1":
				this.directives ? this.directives.yaml.version = "1.1" : this.directives = new Ws({ version: "1.1" }), n = {
					resolveKnownTags: !1,
					schema: "yaml-1.1"
				};
				break;
			case "1.2":
			case "next":
				this.directives ? this.directives.yaml.version = e : this.directives = new Ws({ version: e }), n = {
					resolveKnownTags: !0,
					schema: "core"
				};
				break;
			case null:
				this.directives && delete this.directives, n = null;
				break;
			default: {
				let t = JSON.stringify(e);
				throw Error(`Expected '1.1', '1.2' or null as first argument, but found: ${t}`);
			}
		}
		if (t.schema instanceof Object) this.schema = t.schema;
		else if (n) this.schema = new ql(Object.assign(n, t));
		else throw Error("With a null YAML version, the { schema: Schema } option is required");
	}
	toJS({ json: e, jsonArg: t, mapAsMap: n, maxAliasCount: r, onAnchor: i, reviver: a } = {}) {
		let o = {
			anchors: /* @__PURE__ */ new Map(),
			doc: this,
			keep: !e,
			mapAsMap: n === !0,
			mapKeyWarned: !1,
			maxAliasCount: typeof r == "number" ? r : 100
		}, s = Xs(this.contents, t ?? "", o);
		if (typeof i == "function") for (let { count: e, res: t } of o.anchors.values()) i(t, e);
		return typeof a == "function" ? Ys(a, { "": s }, "", s) : s;
	}
	toJSON(e, t) {
		return this.toJS({
			json: !0,
			jsonArg: e,
			mapAsMap: !1,
			onAnchor: t
		});
	}
	toString(e = {}) {
		if (this.errors.length > 0) throw Error("Document with errors cannot be stringified");
		if ("indent" in e && (!Number.isInteger(e.indent) || Number(e.indent) <= 0)) {
			let t = JSON.stringify(e.indent);
			throw Error(`"indent" option must be a positive integer, not ${t}`);
		}
		return Jl(this, e);
	}
};
function Xl(e) {
	if (js(e)) return !0;
	throw Error("Expected a YAML collection as document contents");
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/errors.js
var Zl = class extends Error {
	constructor(e, t, n, r) {
		super(), this.name = e, this.code = n, this.message = r, this.pos = t;
	}
}, Ql = class extends Zl {
	constructor(e, t, n) {
		super("YAMLParseError", e, t, n);
	}
}, $l = class extends Zl {
	constructor(e, t, n) {
		super("YAMLWarning", e, t, n);
	}
}, eu = (e, t) => (n) => {
	if (n.pos[0] === -1) return;
	n.linePos = n.pos.map((e) => t.linePos(e));
	let { line: r, col: i } = n.linePos[0];
	n.message += ` at line ${r}, column ${i}`;
	let a = i - 1, o = e.substring(t.lineStarts[r - 1], t.lineStarts[r]).replace(/[\n\r]+$/, "");
	if (a >= 60 && o.length > 80) {
		let e = Math.min(a - 39, o.length - 79);
		o = "…" + o.substring(e), a -= e - 1;
	}
	if (o.length > 80 && (o = o.substring(0, 79) + "…"), r > 1 && /^ *$/.test(o.substring(0, a))) {
		let n = e.substring(t.lineStarts[r - 2], t.lineStarts[r - 1]);
		n.length > 80 && (n = n.substring(0, 79) + "…\n"), o = n + o;
	}
	if (/[^ ]/.test(o)) {
		let e = 1, t = n.linePos[1];
		t?.line === r && t.col > i && (e = Math.max(1, Math.min(t.col - i, 80 - a)));
		let s = " ".repeat(a) + "^".repeat(e);
		n.message += `:\n\n${o}\n${s}\n`;
	}
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/resolve-props.js
function tu(e, { flow: t, indicator: n, next: r, offset: i, onError: a, parentIndent: o, startOnNewline: s }) {
	let c = !1, l = s, u = s, d = "", f = "", p = !1, m = !1, h = null, g = null, _ = null, v = null, y = null, b = null, x = null;
	for (let i of e) switch (m &&= (i.type !== "space" && i.type !== "newline" && i.type !== "comma" && a(i.offset, "MISSING_CHAR", "Tags and anchors must be separated from the next token by white space"), !1), h &&= (l && i.type !== "comment" && i.type !== "newline" && a(h, "TAB_AS_INDENT", "Tabs are not allowed as indentation"), null), i.type) {
		case "space":
			!t && (n !== "doc-start" || r?.type !== "flow-collection") && i.source.includes("	") && (h = i), u = !0;
			break;
		case "comment": {
			u || a(i, "MISSING_CHAR", "Comments must be separated from other tokens by white space characters");
			let e = i.source.substring(1) || " ";
			d ? d += f + e : d = e, f = "", l = !1;
			break;
		}
		case "newline":
			l ? d ? d += i.source : b && n === "seq-item-ind" || (c = !0) : f += i.source, l = !0, p = !0, (g || _) && (v = i), u = !0;
			break;
		case "anchor":
			g && a(i, "MULTIPLE_ANCHORS", "A node can have at most one anchor"), i.source.endsWith(":") && a(i.offset + i.source.length - 1, "BAD_ALIAS", "Anchor ending in : is ambiguous", !0), g = i, x ??= i.offset, l = !1, u = !1, m = !0;
			break;
		case "tag":
			_ && a(i, "MULTIPLE_TAGS", "A node can have at most one tag"), _ = i, x ??= i.offset, l = !1, u = !1, m = !0;
			break;
		case n:
			(g || _) && a(i, "BAD_PROP_ORDER", `Anchors and tags must be after the ${i.source} indicator`), b && a(i, "UNEXPECTED_TOKEN", `Unexpected ${i.source} in ${t ?? "collection"}`), b = i, l = n === "seq-item-ind" || n === "explicit-key-ind", u = !1;
			break;
		case "comma": if (t) {
			y && a(i, "UNEXPECTED_TOKEN", `Unexpected , in ${t}`), y = i, l = !1, u = !1;
			break;
		}
		default: a(i, "UNEXPECTED_TOKEN", `Unexpected ${i.type} token`), l = !1, u = !1;
	}
	let S = e[e.length - 1], C = S ? S.offset + S.source.length : i;
	return m && r && r.type !== "space" && r.type !== "newline" && r.type !== "comma" && (r.type !== "scalar" || r.source !== "") && a(r.offset, "MISSING_CHAR", "Tags and anchors must be separated from the next token by white space"), h && (l && h.indent <= o || r?.type === "block-map" || r?.type === "block-seq") && a(h, "TAB_AS_INDENT", "Tabs are not allowed as indentation"), {
		comma: y,
		found: b,
		spaceBefore: c,
		comment: d,
		hasNewline: p,
		anchor: g,
		tag: _,
		newlineAfterProp: v,
		end: C,
		start: x ?? C
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/util-contains-newline.js
function nu(e) {
	if (!e) return null;
	switch (e.type) {
		case "alias":
		case "scalar":
		case "double-quoted-scalar":
		case "single-quoted-scalar":
			if (e.source.includes("\n")) return !0;
			if (e.end) {
				for (let t of e.end) if (t.type === "newline") return !0;
			}
			return !1;
		case "flow-collection":
			for (let t of e.items) {
				for (let e of t.start) if (e.type === "newline") return !0;
				if (t.sep) {
					for (let e of t.sep) if (e.type === "newline") return !0;
				}
				if (nu(t.key) || nu(t.value)) return !0;
			}
			return !1;
		default: return !0;
	}
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/util-flow-indent-check.js
function ru(e, t, n) {
	if (t?.type === "flow-collection") {
		let r = t.end[0];
		r.indent === e && (r.source === "]" || r.source === "}") && nu(t) && n(r, "BAD_INDENT", "Flow end indicator should be more indented than parent", !0);
	}
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/util-map-includes.js
function iu(e, t, n) {
	let { uniqueKeys: r } = e.options;
	if (r === !1) return !1;
	let i = typeof r == "function" ? r : (e, t) => e === t || ks(e) && ks(t) && e.value === t.value;
	return t.some((e) => i(e.key, n));
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/resolve-block-map.js
var au = "All mapping items must start at the same column";
function ou({ composeNode: e, composeEmptyNode: t }, n, r, i, a) {
	let o = new ((a?.nodeClass) ?? Kc)(n.schema);
	n.atRoot &&= !1;
	let s = r.offset, c = null;
	for (let a of r.items) {
		let { start: l, key: u, sep: d, value: f } = a, p = tu(l, {
			indicator: "explicit-key-ind",
			next: u ?? d?.[0],
			offset: s,
			onError: i,
			parentIndent: r.indent,
			startOnNewline: !0
		}), m = !p.found;
		if (m) {
			if (u && (u.type === "block-seq" ? i(s, "BLOCK_AS_IMPLICIT_KEY", "A block sequence may not be used as an implicit map key") : "indent" in u && u.indent !== r.indent && i(s, "BAD_INDENT", au)), !p.anchor && !p.tag && !d) {
				c = p.end, p.comment && (o.comment ? o.comment += "\n" + p.comment : o.comment = p.comment);
				continue;
			}
			(p.newlineAfterProp || nu(u)) && i(u ?? l[l.length - 1], "MULTILINE_IMPLICIT_KEY", "Implicit keys need to be on a single line");
		} else p.found?.indent !== r.indent && i(s, "BAD_INDENT", au);
		n.atKey = !0;
		let h = p.end, g = u ? e(n, u, p, i) : t(n, h, l, null, p, i);
		n.schema.compat && ru(r.indent, u, i), n.atKey = !1, iu(n, o.items, g) && i(h, "DUPLICATE_KEY", "Map keys must be unique");
		let _ = tu(d ?? [], {
			indicator: "map-value-ind",
			next: f,
			offset: g.range[2],
			onError: i,
			parentIndent: r.indent,
			startOnNewline: !u || u.type === "block-scalar"
		});
		if (s = _.end, _.found) {
			m && (f?.type === "block-map" && !_.hasNewline && i(s, "BLOCK_AS_IMPLICIT_KEY", "Nested mappings are not allowed in compact mappings"), n.options.strict && p.start < _.found.offset - 1024 && i(g.range, "KEY_OVER_1024_CHARS", "The : indicator must be at most 1024 chars after the start of an implicit block mapping key"));
			let c = f ? e(n, f, _, i) : t(n, s, d, null, _, i);
			n.schema.compat && ru(r.indent, f, i), s = c.range[2];
			let l = new Bc(g, c);
			n.options.keepSourceTokens && (l.srcToken = a), o.items.push(l);
		} else {
			m && i(g.range, "MISSING_CHAR", "Implicit map keys need to be followed by map values"), _.comment && (g.comment ? g.comment += "\n" + _.comment : g.comment = _.comment);
			let e = new Bc(g);
			n.options.keepSourceTokens && (e.srcToken = a), o.items.push(e);
		}
	}
	return c && c < s && i(c, "IMPOSSIBLE", "Map comment with trailing content"), o.range = [
		r.offset,
		s,
		c ?? s
	], o;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/resolve-block-seq.js
function su({ composeNode: e, composeEmptyNode: t }, n, r, i, a) {
	let o = new ((a?.nodeClass) ?? Jc)(n.schema);
	n.atRoot &&= !1, n.atKey &&= !1;
	let s = r.offset, c = null;
	for (let { start: a, value: l } of r.items) {
		let u = tu(a, {
			indicator: "seq-item-ind",
			next: l,
			offset: s,
			onError: i,
			parentIndent: r.indent,
			startOnNewline: !0
		});
		if (!u.found) {
			if (u.anchor || u.tag || l) l?.type === "block-seq" ? i(u.end, "BAD_INDENT", "All sequence items must start at the same column") : i(s, "MISSING_CHAR", "Sequence item without - indicator");
			else {
				c = u.end, u.comment && (o.comment = u.comment);
				continue;
			}
		}
		let d = l ? e(n, l, u, i) : t(n, u.end, a, null, u, i);
		n.schema.compat && ru(r.indent, l, i), s = d.range[2], o.items.push(d);
	}
	return o.range = [
		r.offset,
		s,
		c ?? s
	], o;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/resolve-end.js
function cu(e, t, n, r) {
	let i = "";
	if (e) {
		let a = !1, o = "";
		for (let s of e) {
			let { source: e, type: c } = s;
			switch (c) {
				case "space":
					a = !0;
					break;
				case "comment": {
					n && !a && r(s, "MISSING_CHAR", "Comments must be separated from other tokens by white space characters");
					let t = e.substring(1) || " ";
					i ? i += o + t : i = t, o = "";
					break;
				}
				case "newline":
					i && (o += e), a = !0;
					break;
				default: r(s, "UNEXPECTED_TOKEN", `Unexpected ${c} at node end`);
			}
			t += e.length;
		}
	}
	return {
		comment: i,
		offset: t
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/resolve-flow-collection.js
var lu = "Block collections are not allowed within flow collections", uu = (e) => e && (e.type === "block-map" || e.type === "block-seq");
function du({ composeNode: e, composeEmptyNode: t }, n, r, i, a) {
	let o = r.start.source === "{", s = o ? "flow map" : "flow sequence", c = new ((a?.nodeClass) ?? (o ? Kc : Jc))(n.schema);
	c.flow = !0;
	let l = n.atRoot;
	l && (n.atRoot = !1), n.atKey &&= !1;
	let u = r.offset + r.start.source.length;
	for (let a = 0; a < r.items.length; ++a) {
		let l = r.items[a], { start: d, key: f, sep: p, value: m } = l, h = tu(d, {
			flow: s,
			indicator: "explicit-key-ind",
			next: f ?? p?.[0],
			offset: u,
			onError: i,
			parentIndent: r.indent,
			startOnNewline: !1
		});
		if (!h.found) {
			if (!h.anchor && !h.tag && !p && !m) {
				a === 0 && h.comma ? i(h.comma, "UNEXPECTED_TOKEN", `Unexpected , in ${s}`) : a < r.items.length - 1 && i(h.start, "UNEXPECTED_TOKEN", `Unexpected empty item in ${s}`), h.comment && (c.comment ? c.comment += "\n" + h.comment : c.comment = h.comment), u = h.end;
				continue;
			}
			!o && n.options.strict && nu(f) && i(f, "MULTILINE_IMPLICIT_KEY", "Implicit keys of flow sequence pairs need to be on a single line");
		}
		if (a === 0) h.comma && i(h.comma, "UNEXPECTED_TOKEN", `Unexpected , in ${s}`);
		else if (h.comma || i(h.start, "MISSING_CHAR", `Missing , between ${s} items`), h.comment) {
			let e = "";
			loop: for (let t of d) switch (t.type) {
				case "comma":
				case "space": break;
				case "comment":
					e = t.source.substring(1);
					break loop;
				default: break loop;
			}
			if (e) {
				let t = c.items[c.items.length - 1];
				Os(t) && (t = t.value ?? t.key), t.comment ? t.comment += "\n" + e : t.comment = e, h.comment = h.comment.substring(e.length + 1);
			}
		}
		if (!o && !p && !h.found) {
			let r = m ? e(n, m, h, i) : t(n, h.end, p, null, h, i);
			c.items.push(r), u = r.range[2], uu(m) && i(r.range, "BLOCK_IN_FLOW", lu);
		} else {
			n.atKey = !0;
			let a = h.end, g = f ? e(n, f, h, i) : t(n, a, d, null, h, i);
			uu(f) && i(g.range, "BLOCK_IN_FLOW", lu), n.atKey = !1;
			let _ = tu(p ?? [], {
				flow: s,
				indicator: "map-value-ind",
				next: m,
				offset: g.range[2],
				onError: i,
				parentIndent: r.indent,
				startOnNewline: !1
			});
			if (_.found) {
				if (!o && !h.found && n.options.strict) {
					if (p) for (let e of p) {
						if (e === _.found) break;
						if (e.type === "newline") {
							i(e, "MULTILINE_IMPLICIT_KEY", "Implicit keys of flow sequence pairs need to be on a single line");
							break;
						}
					}
					h.start < _.found.offset - 1024 && i(_.found, "KEY_OVER_1024_CHARS", "The : indicator must be at most 1024 chars after the start of an implicit flow sequence key");
				}
			} else m && ("source" in m && m.source?.[0] === ":" ? i(m, "MISSING_CHAR", `Missing space after : in ${s}`) : i(_.start, "MISSING_CHAR", `Missing , or : between ${s} items`));
			let v = m ? e(n, m, _, i) : _.found ? t(n, _.end, p, null, _, i) : null;
			v ? uu(m) && i(v.range, "BLOCK_IN_FLOW", lu) : _.comment && (g.comment ? g.comment += "\n" + _.comment : g.comment = _.comment);
			let y = new Bc(g, v);
			if (n.options.keepSourceTokens && (y.srcToken = l), o) {
				let e = c;
				iu(n, e.items, g) && i(a, "DUPLICATE_KEY", "Map keys must be unique"), e.items.push(y);
			} else {
				let e = new Kc(n.schema);
				e.flow = !0, e.items.push(y);
				let t = (v ?? g).range;
				e.range = [
					g.range[0],
					t[1],
					t[2]
				], c.items.push(e);
			}
			u = v ? v.range[2] : _.end;
		}
	}
	let d = o ? "}" : "]", [f, ...p] = r.end, m = u;
	if (f?.source === d) m = f.offset + f.source.length;
	else {
		let e = s[0].toUpperCase() + s.substring(1), t = l ? `${e} must end with a ${d}` : `${e} in block collection must be sufficiently indented and end with a ${d}`;
		i(u, l ? "MISSING_CHAR" : "BAD_INDENT", t), f && f.source.length !== 1 && p.unshift(f);
	}
	if (p.length > 0) {
		let e = cu(p, m, n.options.strict, i);
		e.comment && (c.comment ? c.comment += "\n" + e.comment : c.comment = e.comment), c.range = [
			r.offset,
			m,
			e.offset
		];
	} else c.range = [
		r.offset,
		m,
		m
	];
	return c;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/compose-collection.js
function fu(e, t, n, r, i, a) {
	let o = n.type === "block-map" ? ou(e, t, n, r, a) : n.type === "block-seq" ? su(e, t, n, r, a) : du(e, t, n, r, a), s = o.constructor;
	return i === "!" || i === s.tagName ? (o.tag = s.tagName, o) : (i && (o.tag = i), o);
}
function pu(e, t, n, r, i) {
	let a = r.tag, o = a ? t.directives.tagName(a.source, (e) => i(a, "TAG_RESOLVE_FAILED", e)) : null;
	if (n.type === "block-seq") {
		let { anchor: e, newlineAfterProp: t } = r, n = e && a ? e.offset > a.offset ? e : a : e ?? a;
		n && (!t || t.offset < n.offset) && i(n, "MISSING_CHAR", "Missing newline after block sequence props");
	}
	let s = n.type === "block-map" ? "map" : n.type === "block-seq" ? "seq" : n.start.source === "{" ? "map" : "seq";
	if (!a || !o || o === "!" || o === Kc.tagName && s === "map" || o === Jc.tagName && s === "seq") return fu(e, t, n, i, o);
	let c = t.schema.tags.find((e) => e.tag === o && e.collection === s);
	if (!c) {
		let r = t.schema.knownTags[o];
		if (r?.collection === s) t.schema.tags.push(Object.assign({}, r, { default: !1 })), c = r;
		else return r ? i(a, "BAD_COLLECTION_TYPE", `${r.tag} used for ${s} collection, but expects ${r.collection ?? "scalar"}`, !0) : i(a, "TAG_RESOLVE_FAILED", `Unresolved tag: ${o}`, !0), fu(e, t, n, i, o);
	}
	let l = fu(e, t, n, i, o, c), u = c.resolve?.(l, (e) => i(a, "TAG_RESOLVE_FAILED", e), t.options) ?? l, d = Ms(u) ? u : new Z(u);
	return d.range = l.range, d.tag = o, c?.format && (d.format = c.format), d;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/resolve-block-scalar.js
function mu(e, t, n) {
	let r = t.offset, i = hu(t, e.options.strict, n);
	if (!i) return {
		value: "",
		type: null,
		comment: "",
		range: [
			r,
			r,
			r
		]
	};
	let a = i.mode === ">" ? Z.BLOCK_FOLDED : Z.BLOCK_LITERAL, o = t.source ? gu(t.source) : [], s = o.length;
	for (let e = o.length - 1; e >= 0; --e) {
		let t = o[e][1];
		if (t === "" || t === "\r") s = e;
		else break;
	}
	if (s === 0) {
		let e = i.chomp === "+" && o.length > 0 ? "\n".repeat(Math.max(1, o.length - 1)) : "", n = r + i.length;
		return t.source && (n += t.source.length), {
			value: e,
			type: a,
			comment: i.comment,
			range: [
				r,
				n,
				n
			]
		};
	}
	let c = t.indent + i.indent, l = t.offset + i.length, u = 0;
	for (let t = 0; t < s; ++t) {
		let [r, a] = o[t];
		if (a === "" || a === "\r") i.indent === 0 && r.length > c && (c = r.length);
		else {
			r.length < c && n(l + r.length, "MISSING_CHAR", "Block scalars with more-indented leading empty lines must use an explicit indentation indicator"), i.indent === 0 && (c = r.length), u = t, c === 0 && !e.atRoot && n(l, "BAD_INDENT", "Block scalar values in collections must be indented");
			break;
		}
		l += r.length + a.length + 1;
	}
	for (let e = o.length - 1; e >= s; --e) o[e][0].length > c && (s = e + 1);
	let d = "", f = "", p = !1;
	for (let e = 0; e < u; ++e) d += o[e][0].slice(c) + "\n";
	for (let e = u; e < s; ++e) {
		let [t, r] = o[e];
		l += t.length + r.length + 1;
		let s = r[r.length - 1] === "\r";
		/* istanbul ignore if already caught in lexer */
		if (s && (r = r.slice(0, -1)), r && t.length < c) {
			let e = `Block scalar lines must not be less indented than their ${i.indent ? "explicit indentation indicator" : "first line"}`;
			n(l - r.length - (s ? 2 : 1), "BAD_INDENT", e), t = "";
		}
		a === Z.BLOCK_LITERAL ? (d += f + t.slice(c) + r, f = "\n") : t.length > c || r[0] === "	" ? (f === " " ? f = "\n" : !p && f === "\n" && (f = "\n\n"), d += f + t.slice(c) + r, f = "\n", p = !0) : r === "" ? f === "\n" ? d += "\n" : f = "\n" : (d += f + r, f = " ", p = !1);
	}
	switch (i.chomp) {
		case "-": break;
		case "+":
			for (let e = s; e < o.length; ++e) d += "\n" + o[e][0].slice(c);
			d[d.length - 1] !== "\n" && (d += "\n");
			break;
		default: d += "\n";
	}
	let m = r + i.length + t.source.length;
	return {
		value: d,
		type: a,
		comment: i.comment,
		range: [
			r,
			m,
			m
		]
	};
}
function hu({ offset: e, props: t }, n, r) {
	/* istanbul ignore if should not happen */
	if (t[0].type !== "block-scalar-header") return r(t[0], "IMPOSSIBLE", "Block scalar header not found"), null;
	let { source: i } = t[0], a = i[0], o = 0, s = "", c = -1;
	for (let t = 1; t < i.length; ++t) {
		let n = i[t];
		if (!s && (n === "-" || n === "+")) s = n;
		else {
			let r = Number(n);
			!o && r ? o = r : c === -1 && (c = e + t);
		}
	}
	c !== -1 && r(c, "UNEXPECTED_TOKEN", `Block scalar header includes extra characters: ${i}`);
	let l = !1, u = "", d = i.length;
	for (let e = 1; e < t.length; ++e) {
		let i = t[e];
		switch (i.type) {
			case "space": l = !0;
			case "newline":
				d += i.source.length;
				break;
			case "comment":
				n && !l && r(i, "MISSING_CHAR", "Comments must be separated from other tokens by white space characters"), d += i.source.length, u = i.source.substring(1);
				break;
			case "error":
				r(i, "UNEXPECTED_TOKEN", i.message), d += i.source.length;
				break;
			/* istanbul ignore next should not happen */
			default: {
				r(i, "UNEXPECTED_TOKEN", `Unexpected token in block scalar header: ${i.type}`);
				let e = i.source;
				e && typeof e == "string" && (d += e.length);
			}
		}
	}
	return {
		mode: a,
		indent: o,
		chomp: s,
		comment: u,
		length: d
	};
}
function gu(e) {
	let t = e.split(/\n( *)/), n = t[0], r = n.match(/^( *)/), i = [r?.[1] ? [r[1], n.slice(r[1].length)] : ["", n]];
	for (let e = 1; e < t.length; e += 2) i.push([t[e], t[e + 1]]);
	return i;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/resolve-flow-scalar.js
function _u(e, t, n) {
	let { offset: r, type: i, source: a, end: o } = e, s, c, l = (e, t, i) => n(r + e, t, i);
	switch (i) {
		case "scalar":
			s = Z.PLAIN, c = vu(a, l);
			break;
		case "single-quoted-scalar":
			s = Z.QUOTE_SINGLE, c = yu(a, l);
			break;
		case "double-quoted-scalar":
			s = Z.QUOTE_DOUBLE, c = xu(a, l);
			break;
		/* istanbul ignore next should not happen */
		default: return n(e, "UNEXPECTED_TOKEN", `Expected a flow scalar value, but found: ${i}`), {
			value: "",
			type: null,
			comment: "",
			range: [
				r,
				r + a.length,
				r + a.length
			]
		};
	}
	let u = r + a.length, d = cu(o, u, t, n);
	return {
		value: c,
		type: s,
		comment: d.comment,
		range: [
			r,
			u,
			d.offset
		]
	};
}
function vu(e, t) {
	let n = "";
	switch (e[0]) {
		/* istanbul ignore next should not happen */
		case "	":
			n = "a tab character";
			break;
		case ",":
			n = "flow indicator character ,";
			break;
		case "%":
			n = "directive indicator character %";
			break;
		case "|":
		case ">":
			n = `block scalar indicator ${e[0]}`;
			break;
		case "@":
		case "`": n = `reserved character ${e[0]}`;
	}
	return n && t(0, "BAD_SCALAR_START", `Plain value cannot start with ${n}`), bu(e);
}
function yu(e, t) {
	return (e[e.length - 1] !== "'" || e.length === 1) && t(e.length, "MISSING_CHAR", "Missing closing 'quote"), bu(e.slice(1, -1)).replace(/''/g, "'");
}
function bu(e) {
	let t = /(.*?)\r?\n/sy, n = t.exec(e);
	if (!n) return e;
	let r, i;
	try {
		r = /* @__PURE__ */ RegExp("(?<![ 	])[ 	]+$"), i = /* @__PURE__ */ RegExp("^[ 	]+|(?<![ 	])[ 	]+$", "g");
	} catch {
		r = /[ \t]+$/, i = /^[ \t]+|[ \t]+$/g;
	}
	let a = n[1].replace(r, ""), o = " ", s = t.lastIndex;
	for (; n = t.exec(e);) {
		let e = n[1].replace(i, "");
		e === "" ? o === "\n" ? a += o : o = "\n" : (a += o + e, o = " "), s = t.lastIndex;
	}
	let c = /[ \t]*(.*)/sy;
	return c.lastIndex = s, n = c.exec(e), a + o + (n?.[1] ?? "");
}
function xu(e, t) {
	let n = "";
	for (let r = 1; r < e.length - 1; ++r) {
		let i = e[r];
		if (i !== "\r" || e[r + 1] !== "\n") {
			if (i === "\n") {
				let { fold: t, offset: i } = Su(e, r);
				n += t, r = i;
			} else if (i === "\\") {
				let i = e[++r], a = Cu[i];
				if (a) n += a;
				else if (i === "\n") for (i = e[r + 1]; i === " " || i === "	";) i = e[++r + 1];
				else if (i === "\r" && e[r + 1] === "\n") for (i = e[++r + 1]; i === " " || i === "	";) i = e[++r + 1];
				else if (i === "x" || i === "u" || i === "U") {
					let a = i === "x" ? 2 : i === "u" ? 4 : 8;
					n += wu(e, r + 1, a, t), r += a;
				} else {
					let i = e.substr(r - 1, 2);
					t(r - 1, "BAD_DQ_ESCAPE", `Invalid escape sequence ${i}`), n += i;
				}
			} else if (i === " " || i === "	") {
				let t = r, a = e[r + 1];
				for (; a === " " || a === "	";) a = e[++r + 1];
				a !== "\n" && (a !== "\r" || e[r + 2] !== "\n") && (n += r > t ? e.slice(t, r + 1) : i);
			} else n += i;
		}
	}
	return (e[e.length - 1] !== "\"" || e.length === 1) && t(e.length, "MISSING_CHAR", "Missing closing \"quote"), n;
}
function Su(e, t) {
	let n = "", r = e[t + 1];
	for (; (r === " " || r === "	" || r === "\n" || r === "\r") && (r !== "\r" || e[t + 2] === "\n");) r === "\n" && (n += "\n"), t += 1, r = e[t + 1];
	return n ||= " ", {
		fold: n,
		offset: t
	};
}
var Cu = {
	0: "\0",
	a: "\x07",
	b: "\b",
	e: "\x1B",
	f: "\f",
	n: "\n",
	r: "\r",
	t: "	",
	v: "\v",
	N: "",
	_: "\xA0",
	L: "\u2028",
	P: "\u2029",
	" ": " ",
	"\"": "\"",
	"/": "/",
	"\\": "\\",
	"	": "	"
};
function wu(e, t, n, r) {
	let i = e.substr(t, n), a = i.length === n && /^[0-9a-fA-F]+$/.test(i) ? parseInt(i, 16) : NaN;
	try {
		return String.fromCodePoint(a);
	} catch {
		let i = e.substr(t - 2, n + 2);
		return r(t - 2, "BAD_DQ_ESCAPE", `Invalid escape sequence ${i}`), i;
	}
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/compose-scalar.js
function Tu(e, t, n, r) {
	let { value: i, type: a, comment: o, range: s } = t.type === "block-scalar" ? mu(e, t, r) : _u(t, e.options.strict, r), c = n ? e.directives.tagName(n.source, (e) => r(n, "TAG_RESOLVE_FAILED", e)) : null, l;
	l = e.options.stringKeys && e.atKey ? e.schema[Ss] : c ? Eu(e.schema, i, c, n, r) : t.type === "scalar" ? Du(e, i, t, r) : e.schema[Ss];
	let u;
	try {
		let a = l.resolve(i, (e) => r(n ?? t, "TAG_RESOLVE_FAILED", e), e.options);
		u = ks(a) ? a : new Z(a);
	} catch (e) {
		let a = e instanceof Error ? e.message : String(e);
		r(n ?? t, "TAG_RESOLVE_FAILED", a), u = new Z(i);
	}
	return u.range = s, u.source = i, a && (u.type = a), c && (u.tag = c), l.format && (u.format = l.format), o && (u.comment = o), u;
}
function Eu(e, t, n, r, i) {
	if (n === "!") return e[Ss];
	let a = [];
	for (let t of e.tags) if (!t.collection && t.tag === n) {
		if (t.default && t.test) a.push(t);
		else return t;
	}
	for (let e of a) if (e.test?.test(t)) return e;
	let o = e.knownTags[n];
	return o && !o.collection ? (e.tags.push(Object.assign({}, o, {
		default: !1,
		test: void 0
	})), o) : (i(r, "TAG_RESOLVE_FAILED", `Unresolved tag: ${n}`, n !== "tag:yaml.org,2002:str"), e[Ss]);
}
function Du({ atKey: e, directives: t, schema: n }, r, i, a) {
	let o = n.tags.find((t) => (t.default === !0 || e && t.default === "key") && t.test?.test(r)) || n[Ss];
	if (n.compat) {
		let e = n.compat.find((e) => e.default && e.test?.test(r)) ?? n[Ss];
		o.tag !== e.tag && a(i, "TAG_RESOLVE_FAILED", `Value may be parsed as either ${t.tagString(o.tag)} or ${t.tagString(e.tag)}`, !0);
	}
	return o;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/util-empty-scalar-position.js
function Ou(e, t, n) {
	if (t) {
		n ??= t.length;
		for (let r = n - 1; r >= 0; --r) {
			let n = t[r];
			switch (n.type) {
				case "space":
				case "comment":
				case "newline":
					e -= n.source.length;
					continue;
			}
			for (n = t[++r]; n?.type === "space";) e += n.source.length, n = t[++r];
			break;
		}
	}
	return e;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/compose-node.js
var ku = {
	composeNode: Au,
	composeEmptyNode: ju
};
function Au(e, t, n, r) {
	let i = e.atKey, { spaceBefore: a, comment: o, anchor: s, tag: c } = n, l, u = !0;
	switch (t.type) {
		case "alias":
			l = Mu(e, t, r), (s || c) && r(t, "ALIAS_PROPS", "An alias node must not specify any properties");
			break;
		case "scalar":
		case "single-quoted-scalar":
		case "double-quoted-scalar":
		case "block-scalar":
			l = Tu(e, t, c, r), s && (l.anchor = s.source.substring(1));
			break;
		case "block-map":
		case "block-seq":
		case "flow-collection":
			try {
				l = pu(ku, e, t, n, r), s && (l.anchor = s.source.substring(1));
			} catch (e) {
				r(t, "RESOURCE_EXHAUSTION", e instanceof Error ? e.message : String(e));
			}
			break;
		default: r(t, "UNEXPECTED_TOKEN", t.type === "error" ? t.message : `Unsupported token (type: ${t.type})`), u = !1;
	}
	return l ??= ju(e, t.offset, void 0, null, n, r), s && l.anchor === "" && r(s, "BAD_ALIAS", "Anchor cannot be an empty string"), i && e.options.stringKeys && (!ks(l) || typeof l.value != "string" || l.tag && l.tag !== "tag:yaml.org,2002:str") && r(c ?? t, "NON_STRING_KEY", "With stringKeys, all keys must be strings"), a && (l.spaceBefore = !0), o && (t.type === "scalar" && t.source === "" ? l.comment = o : l.commentBefore = o), e.options.keepSourceTokens && u && (l.srcToken = t), l;
}
function ju(e, t, n, r, { spaceBefore: i, comment: a, anchor: o, tag: s, end: c }, l) {
	let u = Tu(e, {
		type: "scalar",
		offset: Ou(t, n, r),
		indent: -1,
		source: ""
	}, s, l);
	return o && (u.anchor = o.source.substring(1), u.anchor === "" && l(o, "BAD_ALIAS", "Anchor cannot be an empty string")), i && (u.spaceBefore = !0), a && (u.comment = a, u.range[2] = c), u;
}
function Mu({ options: e }, { offset: t, source: n, end: r }, i) {
	let a = new Qs(n.substring(1));
	a.source === "" && i(t, "BAD_ALIAS", "Alias cannot be an empty string"), a.source.endsWith(":") && i(t + n.length - 1, "BAD_ALIAS", "Alias ending in : is ambiguous", !0);
	let o = t + n.length, s = cu(r, o, e.strict, i);
	return a.range = [
		t,
		o,
		s.offset
	], s.comment && (a.comment = s.comment), a;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/compose-doc.js
function Nu(e, t, { offset: n, start: r, value: i, end: a }, o) {
	let s = new Yl(void 0, Object.assign({ _directives: t }, e)), c = {
		atKey: !1,
		atRoot: !0,
		directives: s.directives,
		options: s.options,
		schema: s.schema
	}, l = tu(r, {
		indicator: "doc-start",
		next: i ?? a?.[0],
		offset: n,
		onError: o,
		parentIndent: 0,
		startOnNewline: !0
	});
	l.found && (s.directives.docStart = !0, i && (i.type === "block-map" || i.type === "block-seq") && !l.hasNewline && o(l.end, "MISSING_CHAR", "Block collection cannot start on same line with directives-end marker")), s.contents = i ? Au(c, i, l, o) : ju(c, l.end, r, null, l, o);
	let u = s.contents.range[2], d = cu(a, u, !1, o);
	return d.comment && (s.comment = d.comment), s.range = [
		n,
		u,
		d.offset
	], s;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/compose/composer.js
function Pu(e) {
	if (typeof e == "number") return [e, e + 1];
	if (Array.isArray(e)) return e.length === 2 ? e : [e[0], e[1]];
	let { offset: t, source: n } = e;
	return [t, t + (typeof n == "string" ? n.length : 1)];
}
function Fu(e) {
	let t = "", n = !1, r = !1;
	for (let i = 0; i < e.length; ++i) {
		let a = e[i];
		switch (a[0]) {
			case "#":
				t += (t === "" ? "" : r ? "\n\n" : "\n") + (a.substring(1) || " "), n = !0, r = !1;
				break;
			case "%":
				e[i + 1]?.[0] !== "#" && (i += 1), n = !1;
				break;
			default: n || (r = !0), n = !1;
		}
	}
	return {
		comment: t,
		afterEmptyLine: r
	};
}
var Iu = class {
	constructor(e = {}) {
		this.doc = null, this.atDirectives = !1, this.prelude = [], this.errors = [], this.warnings = [], this.onError = (e, t, n, r) => {
			let i = Pu(e);
			r ? this.warnings.push(new $l(i, t, n)) : this.errors.push(new Ql(i, t, n));
		}, this.directives = new Ws({ version: e.version || "1.2" }), this.options = e;
	}
	decorate(e, t) {
		let { comment: n, afterEmptyLine: r } = Fu(this.prelude);
		if (n) {
			let i = e.contents;
			if (t) e.comment = e.comment ? `${e.comment}\n${n}` : n;
			else if (r || e.directives.docStart || !i) e.commentBefore = n;
			else if (js(i) && !i.flow && i.items.length > 0) {
				let e = i.items[0];
				Os(e) && (e = e.key);
				let t = e.commentBefore;
				e.commentBefore = t ? `${n}\n${t}` : n;
			} else {
				let e = i.commentBefore;
				i.commentBefore = e ? `${n}\n${e}` : n;
			}
		}
		if (t) {
			for (let t = 0; t < this.errors.length; ++t) e.errors.push(this.errors[t]);
			for (let t = 0; t < this.warnings.length; ++t) e.warnings.push(this.warnings[t]);
		} else e.errors = this.errors, e.warnings = this.warnings;
		this.prelude = [], this.errors = [], this.warnings = [];
	}
	streamInfo() {
		return {
			comment: Fu(this.prelude).comment,
			directives: this.directives,
			errors: this.errors,
			warnings: this.warnings
		};
	}
	*compose(e, t = !1, n = -1) {
		for (let t of e) yield* this.next(t);
		yield* this.end(t, n);
	}
	*next(e) {
		switch (e.type) {
			case "directive":
				this.directives.add(e.source, (t, n, r) => {
					let i = Pu(e);
					i[0] += t, this.onError(i, "BAD_DIRECTIVE", n, r);
				}), this.prelude.push(e.source), this.atDirectives = !0;
				break;
			case "document": {
				let t = Nu(this.options, this.directives, e, this.onError);
				this.atDirectives && !t.directives.docStart && this.onError(e, "MISSING_CHAR", "Missing directives-end/doc-start indicator line"), this.decorate(t, !1), this.doc && (yield this.doc), this.doc = t, this.atDirectives = !1;
				break;
			}
			case "byte-order-mark":
			case "space": break;
			case "comment":
			case "newline":
				this.prelude.push(e.source);
				break;
			case "error": {
				let t = e.source ? `${e.message}: ${JSON.stringify(e.source)}` : e.message, n = new Ql(Pu(e), "UNEXPECTED_TOKEN", t);
				this.atDirectives || !this.doc ? this.errors.push(n) : this.doc.errors.push(n);
				break;
			}
			case "doc-end": {
				if (!this.doc) {
					this.errors.push(new Ql(Pu(e), "UNEXPECTED_TOKEN", "Unexpected doc-end without preceding document"));
					break;
				}
				this.doc.directives.docEnd = !0;
				let t = cu(e.end, e.offset + e.source.length, this.doc.options.strict, this.onError);
				if (this.decorate(this.doc, !0), t.comment) {
					let e = this.doc.comment;
					this.doc.comment = e ? `${e}\n${t.comment}` : t.comment;
				}
				this.doc.range[2] = t.offset;
				break;
			}
			default: this.errors.push(new Ql(Pu(e), "UNEXPECTED_TOKEN", `Unsupported token ${e.type}`));
		}
	}
	*end(e = !1, t = -1) {
		if (this.doc) this.decorate(this.doc, !0), yield this.doc, this.doc = null;
		else if (e) {
			let e = new Yl(void 0, Object.assign({ _directives: this.directives }, this.options));
			this.atDirectives && this.onError(t, "MISSING_CHAR", "Missing directives-end indicator line"), e.range = [
				0,
				t,
				t
			], this.decorate(e, !1), yield e;
		}
	}
}, Lu = Symbol("break visit"), Ru = Symbol("skip children"), zu = Symbol("remove item");
function Bu(e, t) {
	"type" in e && e.type === "document" && (e = {
		start: e.start,
		value: e.value
	}), Vu(Object.freeze([]), e, t);
}
Bu.BREAK = Lu, Bu.SKIP = Ru, Bu.REMOVE = zu, Bu.itemAtPath = (e, t) => {
	let n = e;
	for (let [e, r] of t) {
		let t = n?.[e];
		if (t && "items" in t) n = t.items[r];
		else return;
	}
	return n;
}, Bu.parentCollection = (e, t) => {
	let n = Bu.itemAtPath(e, t.slice(0, -1)), r = t[t.length - 1][0], i = n?.[r];
	if (i && "items" in i) return i;
	throw Error("Parent collection not found");
};
function Vu(e, t, n) {
	let r = n(t, e);
	if (typeof r == "symbol") return r;
	for (let i of ["key", "value"]) {
		let a = t[i];
		if (a && "items" in a) {
			for (let t = 0; t < a.items.length; ++t) {
				let r = Vu(Object.freeze(e.concat([[i, t]])), a.items[t], n);
				if (typeof r == "number") t = r - 1;
				else if (r === Lu) return Lu;
				else r === zu && (a.items.splice(t, 1), --t);
			}
			typeof r == "function" && i === "key" && (r = r(t, e));
		}
	}
	return typeof r == "function" ? r(t, e) : r;
}
function Hu(e) {
	switch (e) {
		case "﻿": return "byte-order-mark";
		case "": return "doc-mode";
		case "": return "flow-error-end";
		case "": return "scalar";
		case "---": return "doc-start";
		case "...": return "doc-end";
		case "":
		case "\n":
		case "\r\n": return "newline";
		case "-": return "seq-item-ind";
		case "?": return "explicit-key-ind";
		case ":": return "map-value-ind";
		case "{": return "flow-map-start";
		case "}": return "flow-map-end";
		case "[": return "flow-seq-start";
		case "]": return "flow-seq-end";
		case ",": return "comma";
	}
	switch (e[0]) {
		case " ":
		case "	": return "space";
		case "#": return "comment";
		case "%": return "directive-line";
		case "*": return "alias";
		case "&": return "anchor";
		case "!": return "tag";
		case "'": return "single-quoted-scalar";
		case "\"": return "double-quoted-scalar";
		case "|":
		case ">": return "block-scalar-header";
	}
	return null;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/parse/lexer.js
function Uu(e) {
	switch (e) {
		case void 0:
		case " ":
		case "\n":
		case "\r":
		case "	": return !0;
		default: return !1;
	}
}
var Wu = /* @__PURE__ */ new Set("0123456789ABCDEFabcdef"), Gu = /* @__PURE__ */ new Set("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-#;/?:@&=+$_.!~*'()"), Ku = /* @__PURE__ */ new Set(",[]{}"), qu = /* @__PURE__ */ new Set(" ,[]{}\n\r	"), Ju = (e) => !e || qu.has(e), Yu = class {
	constructor() {
		this.atEnd = !1, this.blockScalarIndent = -1, this.blockScalarKeep = !1, this.buffer = "", this.flowKey = !1, this.flowLevel = 0, this.indentNext = 0, this.indentValue = 0, this.lineEndPos = null, this.next = null, this.pos = 0;
	}
	*lex(e, t = !1) {
		if (e) {
			if (typeof e != "string") throw TypeError("source is not a string");
			this.buffer = this.buffer ? this.buffer + e : e, this.lineEndPos = null;
		}
		this.atEnd = !t;
		let n = this.next ?? "stream";
		for (; n && (t || this.hasChars(1));) n = yield* this.parseNext(n);
	}
	atLineEnd() {
		let e = this.pos, t = this.buffer[e];
		for (; t === " " || t === "	";) t = this.buffer[++e];
		return !t || t === "#" || t === "\n" || t === "\r" && this.buffer[e + 1] === "\n";
	}
	charAt(e) {
		return this.buffer[this.pos + e];
	}
	continueScalar(e) {
		let t = this.buffer[e];
		if (this.indentNext > 0) {
			let n = 0;
			for (; t === " ";) t = this.buffer[++n + e];
			if (t === "\r") {
				let t = this.buffer[n + e + 1];
				if (t === "\n" || !t && !this.atEnd) return e + n + 1;
			}
			return t === "\n" || n >= this.indentNext || !t && !this.atEnd ? e + n : -1;
		}
		if (t === "-" || t === ".") {
			let t = this.buffer.substr(e, 3);
			if ((t === "---" || t === "...") && Uu(this.buffer[e + 3])) return -1;
		}
		return e;
	}
	getLine() {
		let e = this.lineEndPos;
		return (typeof e != "number" || e !== -1 && e < this.pos) && (e = this.buffer.indexOf("\n", this.pos), this.lineEndPos = e), e === -1 ? this.atEnd ? this.buffer.substring(this.pos) : null : (this.buffer[e - 1] === "\r" && --e, this.buffer.substring(this.pos, e));
	}
	hasChars(e) {
		return this.pos + e <= this.buffer.length;
	}
	setNext(e) {
		return this.buffer = this.buffer.substring(this.pos), this.pos = 0, this.lineEndPos = null, this.next = e, null;
	}
	peek(e) {
		return this.buffer.substr(this.pos, e);
	}
	*parseNext(e) {
		switch (e) {
			case "stream": return yield* this.parseStream();
			case "line-start": return yield* this.parseLineStart();
			case "block-start": return yield* this.parseBlockStart();
			case "doc": return yield* this.parseDocument();
			case "flow": return yield* this.parseFlowCollection();
			case "quoted-scalar": return yield* this.parseQuotedScalar();
			case "block-scalar": return yield* this.parseBlockScalar();
			case "plain-scalar": return yield* this.parsePlainScalar();
		}
	}
	*parseStream() {
		let e = this.getLine();
		if (e === null) return this.setNext("stream");
		if (e[0] === "﻿" && (yield* this.pushCount(1), e = e.substring(1)), e[0] === "%") {
			let t = e.length, n = e.indexOf("#");
			for (; n !== -1;) {
				let r = e[n - 1];
				if (r === " " || r === "	") {
					t = n - 1;
					break;
				}
				n = e.indexOf("#", n + 1);
			}
			for (;;) {
				let n = e[t - 1];
				if (n === " " || n === "	") --t;
				else break;
			}
			let r = (yield* this.pushCount(t)) + (yield* this.pushSpaces(!0));
			return yield* this.pushCount(e.length - r), this.pushNewline(), "stream";
		}
		if (this.atLineEnd()) {
			let t = yield* this.pushSpaces(!0);
			return yield* this.pushCount(e.length - t), yield* this.pushNewline(), "stream";
		}
		return yield "", yield* this.parseLineStart();
	}
	*parseLineStart() {
		let e = this.charAt(0);
		if (!e && !this.atEnd) return this.setNext("line-start");
		if (e === "-" || e === ".") {
			if (!this.atEnd && !this.hasChars(4)) return this.setNext("line-start");
			let e = this.peek(3);
			if ((e === "---" || e === "...") && Uu(this.charAt(3))) return yield* this.pushCount(3), this.indentValue = 0, this.indentNext = 0, e === "---" ? "doc" : "stream";
		}
		return this.indentValue = yield* this.pushSpaces(!1), this.indentNext > this.indentValue && !Uu(this.charAt(1)) && (this.indentNext = this.indentValue), yield* this.parseBlockStart();
	}
	*parseBlockStart() {
		let [e, t] = this.peek(2);
		if (!t && !this.atEnd) return this.setNext("block-start");
		if ((e === "-" || e === "?" || e === ":") && Uu(t)) {
			let e = (yield* this.pushCount(1)) + (yield* this.pushSpaces(!0));
			return this.indentNext = this.indentValue + 1, this.indentValue += e, "block-start";
		}
		return "doc";
	}
	*parseDocument() {
		yield* this.pushSpaces(!0);
		let e = this.getLine();
		if (e === null) return this.setNext("doc");
		let t = yield* this.pushIndicators();
		switch (e[t]) {
			case "#": yield* this.pushCount(e.length - t);
			case void 0: return yield* this.pushNewline(), yield* this.parseLineStart();
			case "{":
			case "[": return yield* this.pushCount(1), this.flowKey = !1, this.flowLevel = 1, "flow";
			case "}":
			case "]": return yield* this.pushCount(1), "doc";
			case "*": return yield* this.pushUntil(Ju), "doc";
			case "\"":
			case "'": return yield* this.parseQuotedScalar();
			case "|":
			case ">": return t += yield* this.parseBlockScalarHeader(), t += yield* this.pushSpaces(!0), yield* this.pushCount(e.length - t), yield* this.pushNewline(), yield* this.parseBlockScalar();
			default: return yield* this.parsePlainScalar();
		}
	}
	*parseFlowCollection() {
		let e, t, n = -1;
		do
			e = yield* this.pushNewline(), e > 0 ? (t = yield* this.pushSpaces(!1), this.indentValue = n = t) : t = 0, t += yield* this.pushSpaces(!0);
		while (e + t > 0);
		let r = this.getLine();
		if (r === null) return this.setNext("flow");
		if ((n !== -1 && n < this.indentNext && r[0] !== "#" || n === 0 && (r.startsWith("---") || r.startsWith("...")) && Uu(r[3])) && (n !== this.indentNext - 1 || this.flowLevel !== 1 || r[0] !== "]" && r[0] !== "}")) return this.flowLevel = 0, yield "", yield* this.parseLineStart();
		let i = 0;
		for (; r[i] === ",";) i += yield* this.pushCount(1), i += yield* this.pushSpaces(!0), this.flowKey = !1;
		switch (i += yield* this.pushIndicators(), r[i]) {
			case void 0: return "flow";
			case "#": return yield* this.pushCount(r.length - i), "flow";
			case "{":
			case "[": return yield* this.pushCount(1), this.flowKey = !1, this.flowLevel += 1, "flow";
			case "}":
			case "]": return yield* this.pushCount(1), this.flowKey = !0, --this.flowLevel, this.flowLevel ? "flow" : "doc";
			case "*": return yield* this.pushUntil(Ju), "flow";
			case "\"":
			case "'": return this.flowKey = !0, yield* this.parseQuotedScalar();
			case ":": {
				let e = this.charAt(1);
				if (this.flowKey || Uu(e) || e === ",") return this.flowKey = !1, yield* this.pushCount(1), yield* this.pushSpaces(!0), "flow";
			}
			default: return this.flowKey = !1, yield* this.parsePlainScalar();
		}
	}
	*parseQuotedScalar() {
		let e = this.charAt(0), t = this.buffer.indexOf(e, this.pos + 1);
		if (e === "'") for (; t !== -1 && this.buffer[t + 1] === "'";) t = this.buffer.indexOf("'", t + 2);
		else for (; t !== -1;) {
			let e = 0;
			for (; this.buffer[t - 1 - e] === "\\";) e += 1;
			if (e % 2 == 0) break;
			t = this.buffer.indexOf("\"", t + 1);
		}
		let n = this.buffer.substring(0, t), r = n.indexOf("\n", this.pos);
		if (r !== -1) {
			for (; r !== -1;) {
				let e = this.continueScalar(r + 1);
				if (e === -1) break;
				r = n.indexOf("\n", e);
			}
			r !== -1 && (t = r - (n[r - 1] === "\r" ? 2 : 1));
		}
		if (t === -1) {
			if (!this.atEnd) return this.setNext("quoted-scalar");
			t = this.buffer.length;
		}
		return yield* this.pushToIndex(t + 1, !1), this.flowLevel ? "flow" : "doc";
	}
	*parseBlockScalarHeader() {
		this.blockScalarIndent = -1, this.blockScalarKeep = !1;
		let e = this.pos;
		for (;;) {
			let t = this.buffer[++e];
			if (t === "+") this.blockScalarKeep = !0;
			else if (t > "0" && t <= "9") this.blockScalarIndent = Number(t) - 1;
			else if (t !== "-") break;
		}
		return yield* this.pushUntil((e) => Uu(e) || e === "#");
	}
	*parseBlockScalar() {
		let e = this.pos - 1, t = 0, n;
		loop: for (let r = this.pos; n = this.buffer[r]; ++r) switch (n) {
			case " ":
				t += 1;
				break;
			case "\n":
				e = r, t = 0;
				break;
			case "\r": {
				let e = this.buffer[r + 1];
				if (!e && !this.atEnd) return this.setNext("block-scalar");
				if (e === "\n") break;
			}
			default: break loop;
		}
		if (!n && !this.atEnd) return this.setNext("block-scalar");
		if (t >= this.indentNext) {
			this.indentNext = this.blockScalarIndent === -1 ? t : this.blockScalarIndent + (this.indentNext === 0 ? 1 : this.indentNext);
			do {
				let t = this.continueScalar(e + 1);
				if (t === -1) break;
				e = this.buffer.indexOf("\n", t);
			} while (e !== -1);
			if (e === -1) {
				if (!this.atEnd) return this.setNext("block-scalar");
				e = this.buffer.length;
			}
		}
		let r = e + 1;
		for (n = this.buffer[r]; n === " ";) n = this.buffer[++r];
		if (n === "	") {
			for (; n === "	" || n === " " || n === "\r" || n === "\n";) n = this.buffer[++r];
			e = r - 1;
		} else if (!this.blockScalarKeep) do {
			let n = e - 1, r = this.buffer[n];
			r === "\r" && (r = this.buffer[--n]);
			let i = n;
			for (; r === " ";) r = this.buffer[--n];
			if (r === "\n" && n >= this.pos && n + 1 + t > i) e = n;
			else break;
		} while (1);
		return yield "", yield* this.pushToIndex(e + 1, !0), yield* this.parseLineStart();
	}
	*parsePlainScalar() {
		let e = this.flowLevel > 0, t = this.pos - 1, n = this.pos - 1, r;
		for (; r = this.buffer[++n];) if (r === ":") {
			let r = this.buffer[n + 1];
			if (Uu(r) || e && Ku.has(r)) break;
			t = n;
		} else if (Uu(r)) {
			let i = this.buffer[n + 1];
			if (r === "\r" && (i === "\n" ? (n += 1, r = "\n", i = this.buffer[n + 1]) : t = n), i === "#" || e && Ku.has(i)) break;
			if (r === "\n") {
				let e = this.continueScalar(n + 1);
				if (e === -1) break;
				n = Math.max(n, e - 2);
			}
		} else {
			if (e && Ku.has(r)) break;
			t = n;
		}
		return !r && !this.atEnd ? this.setNext("plain-scalar") : (yield "", yield* this.pushToIndex(t + 1, !0), e ? "flow" : "doc");
	}
	*pushCount(e) {
		return e > 0 ? (yield this.buffer.substr(this.pos, e), this.pos += e, e) : 0;
	}
	*pushToIndex(e, t) {
		let n = this.buffer.slice(this.pos, e);
		return n ? (yield n, this.pos += n.length, n.length) : (t && (yield ""), 0);
	}
	*pushIndicators() {
		let e = 0;
		loop: for (;;) {
			switch (this.charAt(0)) {
				case "!":
					e += yield* this.pushTag(), e += yield* this.pushSpaces(!0);
					continue loop;
				case "&":
					e += yield* this.pushUntil(Ju), e += yield* this.pushSpaces(!0);
					continue loop;
				case "-":
				case "?":
				case ":": {
					let t = this.flowLevel > 0, n = this.charAt(1);
					if (Uu(n) || t && Ku.has(n)) {
						t ? this.flowKey &&= !1 : this.indentNext = this.indentValue + 1, e += yield* this.pushCount(1), e += yield* this.pushSpaces(!0);
						continue loop;
					}
				}
			}
			break loop;
		}
		return e;
	}
	*pushTag() {
		if (this.charAt(1) === "<") {
			let e = this.pos + 2, t = this.buffer[e];
			for (; !Uu(t) && t !== ">";) t = this.buffer[++e];
			return yield* this.pushToIndex(t === ">" ? e + 1 : e, !1);
		}
		{
			let e = this.pos + 1, t = this.buffer[e];
			for (; t;) if (Gu.has(t)) t = this.buffer[++e];
			else if (t === "%" && Wu.has(this.buffer[e + 1]) && Wu.has(this.buffer[e + 2])) t = this.buffer[e += 3];
			else break;
			return yield* this.pushToIndex(e, !1);
		}
	}
	*pushNewline() {
		let e = this.buffer[this.pos];
		return e === "\n" ? yield* this.pushCount(1) : e === "\r" && this.charAt(1) === "\n" ? yield* this.pushCount(2) : 0;
	}
	*pushSpaces(e) {
		let t = this.pos - 1, n;
		do
			n = this.buffer[++t];
		while (n === " " || e && n === "	");
		let r = t - this.pos;
		return r > 0 && (yield this.buffer.substr(this.pos, r), this.pos = t), r;
	}
	*pushUntil(e) {
		let t = this.pos, n = this.buffer[t];
		for (; !e(n);) n = this.buffer[++t];
		return yield* this.pushToIndex(t, !1);
	}
}, Xu = class {
	constructor() {
		this.lineStarts = [], this.addNewLine = (e) => this.lineStarts.push(e), this.linePos = (e) => {
			let t = 0, n = this.lineStarts.length;
			for (; t < n;) {
				let r = t + n >> 1;
				this.lineStarts[r] < e ? t = r + 1 : n = r;
			}
			if (this.lineStarts[t] === e) return {
				line: t + 1,
				col: 1
			};
			if (t === 0) return {
				line: 0,
				col: e
			};
			let r = this.lineStarts[t - 1];
			return {
				line: t,
				col: e - r + 1
			};
		};
	}
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/parse/parser.js
function Zu(e, t) {
	for (let n = 0; n < e.length; ++n) if (e[n].type === t) return !0;
	return !1;
}
function Qu(e) {
	for (let t = 0; t < e.length; ++t) switch (e[t].type) {
		case "space":
		case "comment":
		case "newline": break;
		default: return t;
	}
	return -1;
}
function $u(e) {
	switch (e?.type) {
		case "alias":
		case "scalar":
		case "single-quoted-scalar":
		case "double-quoted-scalar":
		case "flow-collection": return !0;
		default: return !1;
	}
}
function ed(e) {
	switch (e.type) {
		case "document": return e.start;
		case "block-map": {
			let t = e.items[e.items.length - 1];
			return t.sep ?? t.start;
		}
		case "block-seq": return e.items[e.items.length - 1].start;
		/* istanbul ignore next should not happen */
		default: return [];
	}
}
function td(e) {
	if (e.length === 0) return [];
	let t = e.length;
	loop: for (; --t >= 0;) switch (e[t].type) {
		case "doc-start":
		case "explicit-key-ind":
		case "map-value-ind":
		case "seq-item-ind":
		case "newline": break loop;
	}
	for (; e[++t]?.type === "space";);
	return e.splice(t, e.length);
}
function nd(e, t) {
	if (t.length < 1e5) Array.prototype.push.apply(e, t);
	else for (let n = 0; n < t.length; ++n) e.push(t[n]);
}
function rd(e) {
	if (e.start.type === "flow-seq-start") for (let t of e.items) t.sep && !t.value && !Zu(t.start, "explicit-key-ind") && !Zu(t.sep, "map-value-ind") && (t.key && (t.value = t.key), delete t.key, $u(t.value) ? t.value.end ? nd(t.value.end, t.sep) : t.value.end = t.sep : nd(t.start, t.sep), delete t.sep);
}
var id = class {
	constructor(e) {
		this.atNewLine = !0, this.atScalar = !1, this.indent = 0, this.offset = 0, this.onKeyLine = !1, this.stack = [], this.source = "", this.type = "", this.lexer = new Yu(), this.onNewLine = e;
	}
	*parse(e, t = !1) {
		this.onNewLine && this.offset === 0 && this.onNewLine(0);
		for (let n of this.lexer.lex(e, t)) yield* this.next(n);
		t || (yield* this.end());
	}
	*next(e) {
		if (this.source = e, this.atScalar) {
			this.atScalar = !1, yield* this.step(), this.offset += e.length;
			return;
		}
		let t = Hu(e);
		if (!t) {
			let t = `Not a YAML token: ${e}`;
			yield* this.pop({
				type: "error",
				offset: this.offset,
				message: t,
				source: e
			}), this.offset += e.length;
		} else if (t === "scalar") this.atNewLine = !1, this.atScalar = !0, this.type = "scalar";
		else {
			switch (this.type = t, yield* this.step(), t) {
				case "newline":
					this.atNewLine = !0, this.indent = 0, this.onNewLine && this.onNewLine(this.offset + e.length);
					break;
				case "space":
					this.atNewLine && e[0] === " " && (this.indent += e.length);
					break;
				case "explicit-key-ind":
				case "map-value-ind":
				case "seq-item-ind":
					this.atNewLine && (this.indent += e.length);
					break;
				case "doc-mode":
				case "flow-error-end": return;
				default: this.atNewLine = !1;
			}
			this.offset += e.length;
		}
	}
	*end() {
		for (; this.stack.length > 0;) yield* this.pop();
	}
	get sourceToken() {
		return {
			type: this.type,
			offset: this.offset,
			indent: this.indent,
			source: this.source
		};
	}
	*step() {
		let e = this.peek(1);
		if (this.type === "doc-end" && e?.type !== "doc-end") {
			for (; this.stack.length > 0;) yield* this.pop();
			this.stack.push({
				type: "doc-end",
				offset: this.offset,
				source: this.source
			});
		} else {
			if (!e) return yield* this.stream();
			switch (e.type) {
				case "document": return yield* this.document(e);
				case "alias":
				case "scalar":
				case "single-quoted-scalar":
				case "double-quoted-scalar": return yield* this.scalar(e);
				case "block-scalar": return yield* this.blockScalar(e);
				case "block-map": return yield* this.blockMap(e);
				case "block-seq": return yield* this.blockSequence(e);
				case "flow-collection": return yield* this.flowCollection(e);
				case "doc-end": return yield* this.documentEnd(e);
			}
			/* istanbul ignore next should not happen */
			yield* this.pop();
		}
	}
	peek(e) {
		return this.stack[this.stack.length - e];
	}
	*pop(e) {
		let t = e ?? this.stack.pop();
		/* istanbul ignore if should not happen */
		if (!t) yield {
			type: "error",
			offset: this.offset,
			source: "",
			message: "Tried to pop an empty stack"
		};
		else if (this.stack.length === 0) yield t;
		else {
			let e = this.peek(1);
			switch (t.type === "block-scalar" ? t.indent = "indent" in e ? e.indent : 0 : t.type === "flow-collection" && e.type === "document" && (t.indent = 0), t.type === "flow-collection" && rd(t), e.type) {
				case "document":
					e.value = t;
					break;
				case "block-scalar":
					e.props.push(t);
					break;
				case "block-map": {
					let n = e.items[e.items.length - 1];
					if (n.value) {
						e.items.push({
							start: [],
							key: t,
							sep: []
						}), this.onKeyLine = !0;
						return;
					}
					if (n.sep) n.value = t;
					else {
						Object.assign(n, {
							key: t,
							sep: []
						}), this.onKeyLine = !n.explicitKey;
						return;
					}
					break;
				}
				case "block-seq": {
					let n = e.items[e.items.length - 1];
					n.value ? e.items.push({
						start: [],
						value: t
					}) : n.value = t;
					break;
				}
				case "flow-collection": {
					let n = e.items[e.items.length - 1];
					!n || n.value ? e.items.push({
						start: [],
						key: t,
						sep: []
					}) : n.sep ? n.value = t : Object.assign(n, {
						key: t,
						sep: []
					});
					return;
				}
				/* istanbul ignore next should not happen */
				default: yield* this.pop(), yield* this.pop(t);
			}
			if ((e.type === "document" || e.type === "block-map" || e.type === "block-seq") && (t.type === "block-map" || t.type === "block-seq")) {
				let n = t.items[t.items.length - 1];
				n && !n.sep && !n.value && n.start.length > 0 && Qu(n.start) === -1 && (t.indent === 0 || n.start.every((e) => e.type !== "comment" || e.indent < t.indent)) && (e.type === "document" ? e.end = n.start : e.items.push({ start: n.start }), t.items.splice(-1, 1));
			}
		}
	}
	*stream() {
		switch (this.type) {
			case "directive-line":
				yield {
					type: "directive",
					offset: this.offset,
					source: this.source
				};
				return;
			case "byte-order-mark":
			case "space":
			case "comment":
			case "newline":
				yield this.sourceToken;
				return;
			case "doc-mode":
			case "doc-start": {
				let e = {
					type: "document",
					offset: this.offset,
					start: []
				};
				this.type === "doc-start" && e.start.push(this.sourceToken), this.stack.push(e);
				return;
			}
		}
		yield {
			type: "error",
			offset: this.offset,
			message: `Unexpected ${this.type} token in YAML stream`,
			source: this.source
		};
	}
	*document(e) {
		if (e.value) return yield* this.lineEnd(e);
		switch (this.type) {
			case "doc-start":
				Qu(e.start) === -1 ? e.start.push(this.sourceToken) : (yield* this.pop(), yield* this.step());
				return;
			case "anchor":
			case "tag":
			case "space":
			case "comment":
			case "newline":
				e.start.push(this.sourceToken);
				return;
		}
		let t = this.startBlockValue(e);
		t ? this.stack.push(t) : yield {
			type: "error",
			offset: this.offset,
			message: `Unexpected ${this.type} token in YAML document`,
			source: this.source
		};
	}
	*scalar(e) {
		if (this.type === "map-value-ind") {
			let t = td(ed(this.peek(2))), n;
			e.end ? (n = e.end, n.push(this.sourceToken), delete e.end) : n = [this.sourceToken];
			let r = {
				type: "block-map",
				offset: e.offset,
				indent: e.indent,
				items: [{
					start: t,
					key: e,
					sep: n
				}]
			};
			this.onKeyLine = !0, this.stack[this.stack.length - 1] = r;
		} else yield* this.lineEnd(e);
	}
	*blockScalar(e) {
		switch (this.type) {
			case "space":
			case "comment":
			case "newline":
				e.props.push(this.sourceToken);
				return;
			case "scalar":
				if (e.source = this.source, this.atNewLine = !0, this.indent = 0, this.onNewLine) {
					let e = this.source.indexOf("\n") + 1;
					for (; e !== 0;) this.onNewLine(this.offset + e), e = this.source.indexOf("\n", e) + 1;
				}
				yield* this.pop();
				break;
			/* istanbul ignore next should not happen */
			default: yield* this.pop(), yield* this.step();
		}
	}
	*blockMap(e) {
		let t = e.items[e.items.length - 1];
		switch (this.type) {
			case "newline":
				if (this.onKeyLine = !1, t.value) {
					let n = "end" in t.value ? t.value.end : void 0;
					(Array.isArray(n) ? n[n.length - 1] : void 0)?.type === "comment" ? n?.push(this.sourceToken) : e.items.push({ start: [this.sourceToken] });
				} else t.sep ? t.sep.push(this.sourceToken) : t.start.push(this.sourceToken);
				return;
			case "space":
			case "comment":
				if (t.value) e.items.push({ start: [this.sourceToken] });
				else if (t.sep) t.sep.push(this.sourceToken);
				else {
					if (this.atIndentedComment(t.start, e.indent)) {
						let n = e.items[e.items.length - 2]?.value?.end;
						if (Array.isArray(n)) {
							nd(n, t.start), n.push(this.sourceToken), e.items.pop();
							return;
						}
					}
					t.start.push(this.sourceToken);
				}
				return;
		}
		if (this.indent >= e.indent) {
			let n = !this.onKeyLine && this.indent === e.indent, r = n && (t.sep || t.explicitKey) && this.type !== "seq-item-ind", i = [];
			if (r && t.sep && !t.value) {
				let n = [];
				for (let r = 0; r < t.sep.length; ++r) {
					let i = t.sep[r];
					switch (i.type) {
						case "newline":
							n.push(r);
							break;
						case "space": break;
						case "comment":
							i.indent > e.indent && (n.length = 0);
							break;
						default: n.length = 0;
					}
				}
				n.length >= 2 && (i = t.sep.splice(n[1]));
			}
			switch (this.type) {
				case "anchor":
				case "tag":
					r || t.value ? (i.push(this.sourceToken), e.items.push({ start: i }), this.onKeyLine = !0) : t.sep ? t.sep.push(this.sourceToken) : t.start.push(this.sourceToken);
					return;
				case "explicit-key-ind":
					!t.sep && !t.explicitKey ? (t.start.push(this.sourceToken), t.explicitKey = !0) : r || t.value ? (i.push(this.sourceToken), e.items.push({
						start: i,
						explicitKey: !0
					})) : this.stack.push({
						type: "block-map",
						offset: this.offset,
						indent: this.indent,
						items: [{
							start: [this.sourceToken],
							explicitKey: !0
						}]
					}), this.onKeyLine = !0;
					return;
				case "map-value-ind":
					if (t.explicitKey) {
						if (!t.sep) {
							if (Zu(t.start, "newline")) Object.assign(t, {
								key: null,
								sep: [this.sourceToken]
							});
							else {
								let e = td(t.start);
								this.stack.push({
									type: "block-map",
									offset: this.offset,
									indent: this.indent,
									items: [{
										start: e,
										key: null,
										sep: [this.sourceToken]
									}]
								});
							}
						} else if (t.value) e.items.push({
							start: [],
							key: null,
							sep: [this.sourceToken]
						});
						else if (Zu(t.sep, "map-value-ind")) this.stack.push({
							type: "block-map",
							offset: this.offset,
							indent: this.indent,
							items: [{
								start: i,
								key: null,
								sep: [this.sourceToken]
							}]
						});
						else if ($u(t.key) && !Zu(t.sep, "newline")) {
							let e = td(t.start), n = t.key, r = t.sep;
							r.push(this.sourceToken), delete t.key, delete t.sep, this.stack.push({
								type: "block-map",
								offset: this.offset,
								indent: this.indent,
								items: [{
									start: e,
									key: n,
									sep: r
								}]
							});
						} else i.length > 0 ? t.sep = t.sep.concat(i, this.sourceToken) : t.sep.push(this.sourceToken);
					} else t.sep ? t.value || r ? e.items.push({
						start: i,
						key: null,
						sep: [this.sourceToken]
					}) : Zu(t.sep, "map-value-ind") ? this.stack.push({
						type: "block-map",
						offset: this.offset,
						indent: this.indent,
						items: [{
							start: [],
							key: null,
							sep: [this.sourceToken]
						}]
					}) : t.sep.push(this.sourceToken) : Object.assign(t, {
						key: null,
						sep: [this.sourceToken]
					});
					this.onKeyLine = !0;
					return;
				case "alias":
				case "scalar":
				case "single-quoted-scalar":
				case "double-quoted-scalar": {
					let n = this.flowScalar(this.type);
					r || t.value ? (e.items.push({
						start: i,
						key: n,
						sep: []
					}), this.onKeyLine = !0) : t.sep ? this.stack.push(n) : (Object.assign(t, {
						key: n,
						sep: []
					}), this.onKeyLine = !0);
					return;
				}
				default: {
					let r = this.startBlockValue(e);
					if (r) {
						if (r.type === "block-seq") {
							if (!t.explicitKey && t.sep && !Zu(t.sep, "newline")) {
								yield* this.pop({
									type: "error",
									offset: this.offset,
									message: "Unexpected block-seq-ind on same line with key",
									source: this.source
								});
								return;
							}
						} else n && e.items.push({ start: i });
						this.stack.push(r);
						return;
					}
				}
			}
		}
		yield* this.pop(), yield* this.step();
	}
	*blockSequence(e) {
		let t = e.items[e.items.length - 1];
		switch (this.type) {
			case "newline":
				if (t.value) {
					let n = "end" in t.value ? t.value.end : void 0;
					(Array.isArray(n) ? n[n.length - 1] : void 0)?.type === "comment" ? n?.push(this.sourceToken) : e.items.push({ start: [this.sourceToken] });
				} else t.start.push(this.sourceToken);
				return;
			case "space":
			case "comment":
				if (t.value) e.items.push({ start: [this.sourceToken] });
				else {
					if (this.atIndentedComment(t.start, e.indent)) {
						let n = e.items[e.items.length - 2]?.value?.end;
						if (Array.isArray(n)) {
							nd(n, t.start), n.push(this.sourceToken), e.items.pop();
							return;
						}
					}
					t.start.push(this.sourceToken);
				}
				return;
			case "anchor":
			case "tag":
				if (t.value || this.indent <= e.indent) break;
				t.start.push(this.sourceToken);
				return;
			case "seq-item-ind":
				if (this.indent !== e.indent) break;
				t.value || Zu(t.start, "seq-item-ind") ? e.items.push({ start: [this.sourceToken] }) : t.start.push(this.sourceToken);
				return;
		}
		if (this.indent > e.indent) {
			let t = this.startBlockValue(e);
			if (t) {
				this.stack.push(t);
				return;
			}
		}
		yield* this.pop(), yield* this.step();
	}
	*flowCollection(e) {
		let t = e.items[e.items.length - 1];
		if (this.type === "flow-error-end") {
			let e;
			do
				yield* this.pop(), e = this.peek(1);
			while (e?.type === "flow-collection");
		} else if (e.end.length === 0) {
			switch (this.type) {
				case "comma":
				case "explicit-key-ind":
					!t || t.sep ? e.items.push({ start: [this.sourceToken] }) : t.start.push(this.sourceToken);
					return;
				case "map-value-ind":
					!t || t.value ? e.items.push({
						start: [],
						key: null,
						sep: [this.sourceToken]
					}) : t.sep ? t.sep.push(this.sourceToken) : Object.assign(t, {
						key: null,
						sep: [this.sourceToken]
					});
					return;
				case "space":
				case "comment":
				case "newline":
				case "anchor":
				case "tag":
					!t || t.value ? e.items.push({ start: [this.sourceToken] }) : t.sep ? t.sep.push(this.sourceToken) : t.start.push(this.sourceToken);
					return;
				case "alias":
				case "scalar":
				case "single-quoted-scalar":
				case "double-quoted-scalar": {
					let n = this.flowScalar(this.type);
					!t || t.value ? e.items.push({
						start: [],
						key: n,
						sep: []
					}) : t.sep ? this.stack.push(n) : Object.assign(t, {
						key: n,
						sep: []
					});
					return;
				}
				case "flow-map-end":
				case "flow-seq-end":
					e.end.push(this.sourceToken);
					return;
			}
			let n = this.startBlockValue(e);
			/* istanbul ignore else should not happen */
			n ? this.stack.push(n) : (yield* this.pop(), yield* this.step());
		} else {
			let t = this.peek(2);
			if (t.type === "block-map" && (this.type === "map-value-ind" && t.indent === e.indent || this.type === "newline" && !t.items[t.items.length - 1].sep)) yield* this.pop(), yield* this.step();
			else if (this.type === "map-value-ind" && t.type !== "flow-collection") {
				let n = td(ed(t));
				rd(e);
				let r = e.end.splice(1, e.end.length);
				r.push(this.sourceToken);
				let i = {
					type: "block-map",
					offset: e.offset,
					indent: e.indent,
					items: [{
						start: n,
						key: e,
						sep: r
					}]
				};
				this.onKeyLine = !0, this.stack[this.stack.length - 1] = i;
			} else yield* this.lineEnd(e);
		}
	}
	flowScalar(e) {
		if (this.onNewLine) {
			let e = this.source.indexOf("\n") + 1;
			for (; e !== 0;) this.onNewLine(this.offset + e), e = this.source.indexOf("\n", e) + 1;
		}
		return {
			type: e,
			offset: this.offset,
			indent: this.indent,
			source: this.source
		};
	}
	startBlockValue(e) {
		switch (this.type) {
			case "alias":
			case "scalar":
			case "single-quoted-scalar":
			case "double-quoted-scalar": return this.flowScalar(this.type);
			case "block-scalar-header": return {
				type: "block-scalar",
				offset: this.offset,
				indent: this.indent,
				props: [this.sourceToken],
				source: ""
			};
			case "flow-map-start":
			case "flow-seq-start": return {
				type: "flow-collection",
				offset: this.offset,
				indent: this.indent,
				start: this.sourceToken,
				items: [],
				end: []
			};
			case "seq-item-ind": return {
				type: "block-seq",
				offset: this.offset,
				indent: this.indent,
				items: [{ start: [this.sourceToken] }]
			};
			case "explicit-key-ind": {
				this.onKeyLine = !0;
				let t = td(ed(e));
				return t.push(this.sourceToken), {
					type: "block-map",
					offset: this.offset,
					indent: this.indent,
					items: [{
						start: t,
						explicitKey: !0
					}]
				};
			}
			case "map-value-ind": {
				this.onKeyLine = !0;
				let t = td(ed(e));
				return {
					type: "block-map",
					offset: this.offset,
					indent: this.indent,
					items: [{
						start: t,
						key: null,
						sep: [this.sourceToken]
					}]
				};
			}
		}
		return null;
	}
	atIndentedComment(e, t) {
		return this.type !== "comment" || this.indent <= t ? !1 : e.every((e) => e.type === "newline" || e.type === "space");
	}
	*documentEnd(e) {
		this.type !== "doc-mode" && (e.end ? e.end.push(this.sourceToken) : e.end = [this.sourceToken], this.type === "newline" && (yield* this.pop()));
	}
	*lineEnd(e) {
		switch (this.type) {
			case "comma":
			case "doc-start":
			case "doc-end":
			case "flow-seq-end":
			case "flow-map-end":
			case "map-value-ind":
				yield* this.pop(), yield* this.step();
				break;
			case "newline": this.onKeyLine = !1;
			default: e.end ? e.end.push(this.sourceToken) : e.end = [this.sourceToken], this.type === "newline" && (yield* this.pop());
		}
	}
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.9.1/8588333d20225b54a202c1a74e7963de6cab6794692dd0e8a9e106c69aaa7d59/node_modules/yaml/browser/dist/public-api.js
function ad(e) {
	let t = e.prettyErrors !== !1;
	return {
		lineCounter: e.lineCounter || t && new Xu() || null,
		prettyErrors: t
	};
}
function od(e, t = {}) {
	let { lineCounter: n, prettyErrors: r } = ad(t), i = new id(n?.addNewLine), a = new Iu(t), o = null;
	for (let t of a.compose(i.parse(e), !0, e.length)) if (!o) o = t;
	else if (o.options.logLevel !== "silent") {
		o.errors.push(new Ql(t.range.slice(0, 2), "MULTIPLE_DOCS", "Source contains multiple documents; please use YAML.parseAllDocuments()"));
		break;
	}
	return r && n && (o.errors.forEach(eu(e, n)), o.warnings.forEach(eu(e, n))), o;
}
function sd(e, t, n) {
	let r;
	typeof t == "function" ? r = t : n === void 0 && t && typeof t == "object" && (n = t);
	let i = od(e, n);
	if (!i) return null;
	if (i.warnings.forEach((e) => Ac(i.options.logLevel, e)), i.errors.length > 0) {
		if (i.options.logLevel !== "silent") throw i.errors[0];
		i.errors = [];
	}
	return i.toJS(Object.assign({ reviver: r }, n));
}
function cd(e, t, n) {
	let r = null;
	if (typeof t == "function" || Array.isArray(t) ? r = t : n === void 0 && t && (n = t), typeof n == "string" && (n = n.length), typeof n == "number") {
		let e = Math.round(n);
		n = e < 1 ? void 0 : e > 8 ? { indent: 8 } : { indent: e };
	}
	if (e === void 0) {
		let { keepUndefined: e } = n ?? t ?? {};
		if (!e) return;
	}
	return Es(e) && !r ? e.toString(n) : new Yl(e, r, n).toString(n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/jsonc-parser/3.3.1/58003957d75a2d2251ecd31c04d4addb7d50a9ac64da586f7d3632efb5050498/node_modules/jsonc-parser/lib/esm/impl/scanner.js
function ld(e, t = !1) {
	let n = e.length, r = 0, i = "", a = 0, o = 16, s = 0, c = 0, l = 0, u = 0, d = 0;
	function f(t, n) {
		let i = 0, a = 0;
		for (; i < t || !n;) {
			let t = e.charCodeAt(r);
			if (t >= 48 && t <= 57) a = a * 16 + t - 48;
			else if (t >= 65 && t <= 70) a = a * 16 + t - 65 + 10;
			else if (t >= 97 && t <= 102) a = a * 16 + t - 97 + 10;
			else break;
			r++, i++;
		}
		return i < t && (a = -1), a;
	}
	function p(e) {
		r = e, i = "", a = 0, o = 16, d = 0;
	}
	function m() {
		let t = r;
		if (e.charCodeAt(r) === 48) r++;
		else for (r++; r < e.length && fd(e.charCodeAt(r));) r++;
		if (r < e.length && e.charCodeAt(r) === 46) {
			if (r++, r < e.length && fd(e.charCodeAt(r))) for (r++; r < e.length && fd(e.charCodeAt(r));) r++;
			else return d = 3, e.substring(t, r);
		}
		let n = r;
		if (r < e.length && (e.charCodeAt(r) === 69 || e.charCodeAt(r) === 101)) {
			if (r++, (r < e.length && e.charCodeAt(r) === 43 || e.charCodeAt(r) === 45) && r++, r < e.length && fd(e.charCodeAt(r))) {
				for (r++; r < e.length && fd(e.charCodeAt(r));) r++;
				n = r;
			} else d = 3;
		}
		return e.substring(t, n);
	}
	function h() {
		let t = "", i = r;
		for (;;) {
			if (r >= n) {
				t += e.substring(i, r), d = 2;
				break;
			}
			let a = e.charCodeAt(r);
			if (a === 34) {
				t += e.substring(i, r), r++;
				break;
			}
			if (a === 92) {
				if (t += e.substring(i, r), r++, r >= n) {
					d = 2;
					break;
				}
				switch (e.charCodeAt(r++)) {
					case 34:
						t += "\"";
						break;
					case 92:
						t += "\\";
						break;
					case 47:
						t += "/";
						break;
					case 98:
						t += "\b";
						break;
					case 102:
						t += "\f";
						break;
					case 110:
						t += "\n";
						break;
					case 114:
						t += "\r";
						break;
					case 116:
						t += "	";
						break;
					case 117:
						let e = f(4, !0);
						e >= 0 ? t += String.fromCharCode(e) : d = 4;
						break;
					default: d = 5;
				}
				i = r;
			} else {
				if (a >= 0 && a <= 31) {
					if (dd(a)) {
						t += e.substring(i, r), d = 2;
						break;
					}
					d = 6;
				}
				r++;
			}
		}
		return t;
	}
	function g() {
		if (i = "", d = 0, a = r, c = s, u = l, r >= n) return a = n, o = 17;
		let t = e.charCodeAt(r);
		if (ud(t)) {
			do
				r++, i += String.fromCharCode(t), t = e.charCodeAt(r);
			while (ud(t));
			return o = 15;
		}
		if (dd(t)) return r++, i += String.fromCharCode(t), t === 13 && e.charCodeAt(r) === 10 && (r++, i += "\n"), s++, l = r, o = 14;
		switch (t) {
			case 123: return r++, o = 1;
			case 125: return r++, o = 2;
			case 91: return r++, o = 3;
			case 93: return r++, o = 4;
			case 58: return r++, o = 6;
			case 44: return r++, o = 5;
			case 34: return r++, i = h(), o = 10;
			case 47:
				let c = r - 1;
				if (e.charCodeAt(r + 1) === 47) {
					for (r += 2; r < n && !dd(e.charCodeAt(r));) r++;
					return i = e.substring(c, r), o = 12;
				}
				if (e.charCodeAt(r + 1) === 42) {
					r += 2;
					let t = n - 1, a = !1;
					for (; r < t;) {
						let t = e.charCodeAt(r);
						if (t === 42 && e.charCodeAt(r + 1) === 47) {
							r += 2, a = !0;
							break;
						}
						r++, dd(t) && (t === 13 && e.charCodeAt(r) === 10 && r++, s++, l = r);
					}
					return a || (r++, d = 1), i = e.substring(c, r), o = 13;
				}
				return i += String.fromCharCode(t), r++, o = 16;
			case 45: if (i += String.fromCharCode(t), r++, r === n || !fd(e.charCodeAt(r))) return o = 16;
			case 48:
			case 49:
			case 50:
			case 51:
			case 52:
			case 53:
			case 54:
			case 55:
			case 56:
			case 57: return i += m(), o = 11;
			default:
				for (; r < n && _(t);) r++, t = e.charCodeAt(r);
				if (a !== r) {
					switch (i = e.substring(a, r), i) {
						case "true": return o = 8;
						case "false": return o = 9;
						case "null": return o = 7;
					}
					return o = 16;
				}
				return i += String.fromCharCode(t), r++, o = 16;
		}
	}
	function _(e) {
		if (ud(e) || dd(e)) return !1;
		switch (e) {
			case 125:
			case 93:
			case 123:
			case 91:
			case 34:
			case 58:
			case 44:
			case 47: return !1;
		}
		return !0;
	}
	function v() {
		let e;
		do
			e = g();
		while (e >= 12 && e <= 15);
		return e;
	}
	return {
		setPosition: p,
		getPosition: () => r,
		scan: t ? v : g,
		getToken: () => o,
		getTokenValue: () => i,
		getTokenOffset: () => a,
		getTokenLength: () => r - a,
		getTokenStartLine: () => c,
		getTokenStartCharacter: () => a - u,
		getTokenError: () => d
	};
}
function ud(e) {
	return e === 32 || e === 9;
}
function dd(e) {
	return e === 10 || e === 13;
}
function fd(e) {
	return e >= 48 && e <= 57;
}
var pd;
(function(e) {
	e[e.lineFeed = 10] = "lineFeed", e[e.carriageReturn = 13] = "carriageReturn", e[e.space = 32] = "space", e[e._0 = 48] = "_0", e[e._1 = 49] = "_1", e[e._2 = 50] = "_2", e[e._3 = 51] = "_3", e[e._4 = 52] = "_4", e[e._5 = 53] = "_5", e[e._6 = 54] = "_6", e[e._7 = 55] = "_7", e[e._8 = 56] = "_8", e[e._9 = 57] = "_9", e[e.a = 97] = "a", e[e.b = 98] = "b", e[e.c = 99] = "c", e[e.d = 100] = "d", e[e.e = 101] = "e", e[e.f = 102] = "f", e[e.g = 103] = "g", e[e.h = 104] = "h", e[e.i = 105] = "i", e[e.j = 106] = "j", e[e.k = 107] = "k", e[e.l = 108] = "l", e[e.m = 109] = "m", e[e.n = 110] = "n", e[e.o = 111] = "o", e[e.p = 112] = "p", e[e.q = 113] = "q", e[e.r = 114] = "r", e[e.s = 115] = "s", e[e.t = 116] = "t", e[e.u = 117] = "u", e[e.v = 118] = "v", e[e.w = 119] = "w", e[e.x = 120] = "x", e[e.y = 121] = "y", e[e.z = 122] = "z", e[e.A = 65] = "A", e[e.B = 66] = "B", e[e.C = 67] = "C", e[e.D = 68] = "D", e[e.E = 69] = "E", e[e.F = 70] = "F", e[e.G = 71] = "G", e[e.H = 72] = "H", e[e.I = 73] = "I", e[e.J = 74] = "J", e[e.K = 75] = "K", e[e.L = 76] = "L", e[e.M = 77] = "M", e[e.N = 78] = "N", e[e.O = 79] = "O", e[e.P = 80] = "P", e[e.Q = 81] = "Q", e[e.R = 82] = "R", e[e.S = 83] = "S", e[e.T = 84] = "T", e[e.U = 85] = "U", e[e.V = 86] = "V", e[e.W = 87] = "W", e[e.X = 88] = "X", e[e.Y = 89] = "Y", e[e.Z = 90] = "Z", e[e.asterisk = 42] = "asterisk", e[e.backslash = 92] = "backslash", e[e.closeBrace = 125] = "closeBrace", e[e.closeBracket = 93] = "closeBracket", e[e.colon = 58] = "colon", e[e.comma = 44] = "comma", e[e.dot = 46] = "dot", e[e.doubleQuote = 34] = "doubleQuote", e[e.minus = 45] = "minus", e[e.openBrace = 123] = "openBrace", e[e.openBracket = 91] = "openBracket", e[e.plus = 43] = "plus", e[e.slash = 47] = "slash", e[e.formFeed = 12] = "formFeed", e[e.tab = 9] = "tab";
})(pd ||= {}), Array(20).fill(0).map((e, t) => " ".repeat(t));
var md = 200;
Array(md).fill(0).map((e, t) => "\n" + " ".repeat(t)), Array(md).fill(0).map((e, t) => "\r" + " ".repeat(t)), Array(md).fill(0).map((e, t) => "\r\n" + " ".repeat(t)), Array(md).fill(0).map((e, t) => "\n" + "	".repeat(t)), Array(md).fill(0).map((e, t) => "\r" + "	".repeat(t)), Array(md).fill(0).map((e, t) => "\r\n" + "	".repeat(t));
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/jsonc-parser/3.3.1/58003957d75a2d2251ecd31c04d4addb7d50a9ac64da586f7d3632efb5050498/node_modules/jsonc-parser/lib/esm/impl/parser.js
var hd;
(function(e) {
	e.DEFAULT = { allowTrailingComma: !1 };
})(hd ||= {});
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/jsonc-parser/3.3.1/58003957d75a2d2251ecd31c04d4addb7d50a9ac64da586f7d3632efb5050498/node_modules/jsonc-parser/lib/esm/main.js
var gd = ld, _d;
(function(e) {
	e[e.None = 0] = "None", e[e.UnexpectedEndOfComment = 1] = "UnexpectedEndOfComment", e[e.UnexpectedEndOfString = 2] = "UnexpectedEndOfString", e[e.UnexpectedEndOfNumber = 3] = "UnexpectedEndOfNumber", e[e.InvalidUnicode = 4] = "InvalidUnicode", e[e.InvalidEscapeCharacter = 5] = "InvalidEscapeCharacter", e[e.InvalidCharacter = 6] = "InvalidCharacter";
})(_d ||= {});
var vd;
(function(e) {
	e[e.OpenBraceToken = 1] = "OpenBraceToken", e[e.CloseBraceToken = 2] = "CloseBraceToken", e[e.OpenBracketToken = 3] = "OpenBracketToken", e[e.CloseBracketToken = 4] = "CloseBracketToken", e[e.CommaToken = 5] = "CommaToken", e[e.ColonToken = 6] = "ColonToken", e[e.NullKeyword = 7] = "NullKeyword", e[e.TrueKeyword = 8] = "TrueKeyword", e[e.FalseKeyword = 9] = "FalseKeyword", e[e.StringLiteral = 10] = "StringLiteral", e[e.NumericLiteral = 11] = "NumericLiteral", e[e.LineCommentTrivia = 12] = "LineCommentTrivia", e[e.BlockCommentTrivia = 13] = "BlockCommentTrivia", e[e.LineBreakTrivia = 14] = "LineBreakTrivia", e[e.Trivia = 15] = "Trivia", e[e.Unknown = 16] = "Unknown", e[e.EOF = 17] = "EOF";
})(vd ||= {});
var yd;
(function(e) {
	e[e.InvalidSymbol = 1] = "InvalidSymbol", e[e.InvalidNumberFormat = 2] = "InvalidNumberFormat", e[e.PropertyNameExpected = 3] = "PropertyNameExpected", e[e.ValueExpected = 4] = "ValueExpected", e[e.ColonExpected = 5] = "ColonExpected", e[e.CommaExpected = 6] = "CommaExpected", e[e.CloseBraceExpected = 7] = "CloseBraceExpected", e[e.CloseBracketExpected = 8] = "CloseBracketExpected", e[e.EndOfFileExpected = 9] = "EndOfFileExpected", e[e.InvalidCommentToken = 10] = "InvalidCommentToken", e[e.UnexpectedEndOfComment = 11] = "UnexpectedEndOfComment", e[e.UnexpectedEndOfString = 12] = "UnexpectedEndOfString", e[e.UnexpectedEndOfNumber = 13] = "UnexpectedEndOfNumber", e[e.InvalidUnicode = 14] = "InvalidUnicode", e[e.InvalidEscapeCharacter = 15] = "InvalidEscapeCharacter", e[e.InvalidCharacter = 16] = "InvalidCharacter";
})(yd ||= {});
// ../../../clicky-ui/packages/ui/src/data/json-view/parse-prefix.ts
var bd = /^"(?:[^"\\]|\\["\\/bfnrt]|\\u[\da-fA-F]{4})*(?:\\(?:u[\da-fA-F]{0,3})?)?$/, xd = /^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d*)?$/, Sd = /^-?(?:0|[1-9]\d*)\.$/;
function Cd(e) {
	if (e.startsWith("\"")) {
		for (let t of e) if (t.charCodeAt(0) < 32) return !1;
		return bd.test(e);
	}
	return e === "-" || xd.test(e) || Sd.test(e) || [
		"true",
		"false",
		"null"
	].some((t) => t.startsWith(e));
}
function wd(e) {
	let t = gd(e), n = () => e.slice(t.getTokenOffset(), t.getTokenOffset() + t.getTokenLength()), r = (e, n) => ({
		status: e,
		message: n,
		offset: t.getTokenOffset()
	});
	function i() {
		do
			t.scan();
		while (/^[ \t\r\n]+$/.test(n()));
	}
	function a(e, t) {
		return e.status === "incomplete" ? {
			...e,
			...Object.keys(t).length > 0 ? { value: t } : {}
		} : e;
	}
	function o() {
		let n = t.getTokenOffset(), a = e.slice(n, n + t.getTokenLength()), o;
		try {
			o = JSON.parse(a);
		} catch (t) {
			if (!(t instanceof SyntaxError)) throw t;
			return r(n + a.length === e.length && Cd(a) ? "incomplete" : "invalid", "Expected a complete JSON string, number, boolean, or null");
		}
		return i(), {
			status: "complete",
			value: o
		};
	}
	function s() {
		return n() === "" ? r("incomplete", "Input ended before a JSON value") : n() === "{" ? c() : n() === "[" ? l() : o();
	}
	function c() {
		let e = {};
		if (i(), n() === "}") return i(), {
			status: "complete",
			value: e
		};
		for (;;) {
			if (n() === "") return a(r("incomplete", "Input ended before a property name or '}'"), e);
			if (!n().startsWith("\"")) return r("invalid", "Expected a quoted property name");
			let t = o();
			if (t.status !== "complete") return a(t, e);
			if (typeof t.value != "string") throw Error("JSON property scanner returned a non-string key");
			if (n() === "") return a(r("incomplete", "Input ended before ':'"), e);
			if (n() !== ":") return r("invalid", "Expected ':' after the property name");
			i();
			let c = s();
			if ("value" in c && Object.defineProperty(e, t.value, {
				value: c.value,
				enumerable: !0,
				configurable: !0,
				writable: !0
			}), c.status !== "complete") return a(c, e);
			if (n() === "}") return i(), {
				status: "complete",
				value: e
			};
			if (n() === "") return a(r("incomplete", "Input ended before '}'"), e);
			if (n() !== ",") return r("invalid", "Expected ',' or '}' after the property value");
			i();
		}
	}
	function l() {
		let e = [];
		if (i(), n() === "]") return i(), {
			status: "complete",
			value: e
		};
		for (;;) {
			let t = s();
			if ("value" in t && e.push(t.value), t.status !== "complete") return a(t, e);
			if (n() === "]") return i(), {
				status: "complete",
				value: e
			};
			if (n() === "") return a(r("incomplete", "Input ended before ']'"), e);
			if (n() !== ",") return r("invalid", "Expected ',' or ']' after the array item");
			i();
		}
	}
	i();
	let u = s();
	return u.status === "complete" && n() !== "" ? r("invalid", "Unexpected content after the JSON value") : u;
}
// ../../../clicky-ui/packages/ui/src/data/json-view/parse-source.ts
function Td(e, t = "json") {
	switch (t) {
		case "json": return Ed(e);
		case "ndjson": return Dd(e);
		default: throw Error(`Unsupported JSON input format: ${String(t)}`);
	}
}
function Ed(e) {
	try {
		return {
			status: "complete",
			value: JSON.parse(e)
		};
	} catch (e) {
		if (!(e instanceof SyntaxError)) throw e;
	}
	let t = wd(e);
	if (t.status === "complete") throw Error("JSON prefix parser accepted a document rejected by JSON.parse");
	let n = e.slice(0, t.offset).split(/\r\n|\r|\n/), r = n.at(-1);
	if (r === void 0) throw Error("JSON diagnostic has no source line");
	return {
		status: t.status,
		...t.status === "incomplete" && "value" in t ? { value: t.value } : {},
		diagnostic: {
			message: t.message,
			line: n.length,
			column: r.length + 1
		}
	};
}
function Dd(e) {
	let t = e.split(/\r\n|\n/), n = t.findLastIndex((e) => e.trim().length > 0), r = [];
	for (let [e, i] of t.entries()) {
		if (i.trim().length === 0) continue;
		let t = Ed(i);
		if (t.status === "complete") {
			r.push(t.value);
			continue;
		}
		if (!t.diagnostic) throw Error("Invalid NDJSON record has no diagnostic");
		let a = t.status === "incomplete" && e === n;
		return {
			status: a ? "incomplete" : "invalid",
			...r.length > 0 ? { value: r } : {},
			diagnostic: {
				...t.diagnostic,
				line: e + 1
			},
			...a && "value" in t ? { partialRecord: {
				line: e + 1,
				value: t.value
			} } : {}
		};
	}
	return {
		status: "complete",
		value: r
	};
}
// ../../../clicky-ui/packages/ui/src/data/json-view/JsonSource.tsx
function Od({ source: e, inputFormat: t, renderValue: n }) {
	let r = (0, Y.useId)(), i = (0, Y.useMemo)(() => Td(e, t), [e, t]), a = t === "ndjson" ? "NDJSON" : "JSON";
	if (i.status !== "complete" && !i.diagnostic) throw Error("JSON source status has no diagnostic");
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "space-y-density-2",
		"data-source-status": i.status,
		children: [
			"value" in i && n(i.value),
			i.partialRecord && /* @__PURE__ */ (0, J.jsxs)("div", {
				className: "space-y-density-1",
				children: [/* @__PURE__ */ (0, J.jsxs)("div", {
					className: "text-muted-foreground font-sans text-sm",
					children: [
						"Line ",
						i.partialRecord.line,
						" (incomplete record)"
					]
				}), n(i.partialRecord.value)]
			}),
			i.status === "incomplete" && !("value" in i) && !i.partialRecord && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "font-sans text-sm text-muted-foreground",
				children: "No completed values received."
			}),
			/* @__PURE__ */ (0, J.jsx)(En, {
				idPrefix: r,
				items: [e],
				readOnly: !0,
				renderHeader: () => /* @__PURE__ */ (0, J.jsxs)("span", {
					className: "flex min-w-0 flex-wrap items-center gap-density-2 font-sans text-sm",
					children: ["Original source", i.diagnostic && /* @__PURE__ */ (0, J.jsxs)("span", {
						role: i.status === "invalid" ? "alert" : "status",
						className: "inline-flex flex-wrap items-center gap-density-2",
						children: [/* @__PURE__ */ (0, J.jsxs)(ge, {
							variant: "status",
							status: i.status === "invalid" ? "error" : "warning",
							size: "xs",
							clickToCopy: !1,
							children: [
								i.status === "invalid" ? "Invalid" : "Incomplete",
								" ",
								a
							]
						}), /* @__PURE__ */ (0, J.jsxs)("span", {
							className: "text-xs text-muted-foreground",
							children: [
								"Line ",
								i.diagnostic.line,
								", column",
								" ",
								i.diagnostic.column
							]
						})]
					})]
				}),
				renderBody: ({ item: e }) => /* @__PURE__ */ (0, J.jsxs)("div", {
					className: "space-y-density-2",
					children: [i.diagnostic && /* @__PURE__ */ (0, J.jsxs)("p", {
						className: "font-sans text-sm text-muted-foreground whitespace-normal",
						children: [
							i.diagnostic.message,
							".",
							i.status === "incomplete" && " Showing completed values only."
						]
					}), /* @__PURE__ */ (0, J.jsx)("pre", {
						className: "max-h-80 overflow-auto whitespace-pre-wrap break-words font-mono text-xs",
						children: e
					})]
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/JsonView.tsx
var kd = {
	compact: "leading-4 [--json-view-indent:0.75rem] [--json-view-row-space:0px]",
	comfortable: "leading-5 [--json-view-indent:1rem] [--json-view-row-space:1px]",
	spacious: "leading-6 [--json-view-indent:1.25rem] [--json-view-row-space:2px]"
};
function Ad({ density: e, format: t = "yaml", ...n }) {
	let r = s(), i = e ?? r, a = Object.hasOwn(n, "source");
	if (a === Object.hasOwn(n, "data")) throw Error("JsonView requires exactly one of data or source");
	if (!a && Object.hasOwn(n, "inputFormat")) throw Error("JsonView inputFormat requires source");
	if (a && typeof n.source != "string") throw Error("JsonView source must be a string");
	let { data: o, source: c, inputFormat: l = "json", ...u } = n;
	return /* @__PURE__ */ (0, J.jsx)("div", {
		"data-format": t,
		"data-density": i,
		className: q("font-mono text-density-base whitespace-pre-wrap break-words", kd[i]),
		children: a && c !== void 0 ? /* @__PURE__ */ (0, J.jsx)(Od, {
			source: c,
			inputFormat: l,
			renderValue: (e) => /* @__PURE__ */ (0, J.jsx)(jd, {
				...u,
				data: e,
				format: t
			})
		}) : /* @__PURE__ */ (0, J.jsx)(jd, {
			...u,
			data: o,
			format: t
		})
	});
}
function jd({ data: e, name: t, depth: n = 0, defaultOpenDepth: r = 2, format: i, sequenceItem: a = !1 }) {
	let [o, s] = (0, Y.useState)(n < r), c = Array.isArray(e), l = typeof e == "object" && e ? Object.entries(e) : [], u = /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [a && /* @__PURE__ */ (0, J.jsx)("span", {
		className: "text-muted-foreground",
		children: "- "
	}), t !== void 0 && /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)("span", {
		className: "text-purple-600 [[data-theme=dark]_&]:text-purple-400",
		children: i === "yaml" ? Nd(t) : t
	}), /* @__PURE__ */ (0, J.jsx)("span", {
		className: "text-muted-foreground",
		children: ": "
	})] })] });
	if (l.length === 0) return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "py-[var(--json-view-row-space)]",
		children: [u, /* @__PURE__ */ (0, J.jsx)(Md, {
			data: e,
			format: i
		})]
	});
	let [d, f] = c ? ["[", "]"] : ["{", "}"], p = c ? `${l.length} ${l.length === 1 ? "item" : "items"}` : "Object";
	return /* @__PURE__ */ (0, J.jsxs)("div", { children: [/* @__PURE__ */ (0, J.jsxs)(fe, {
		variant: "ghost",
		type: "button",
		"aria-expanded": o,
		"aria-label": `${o ? "Collapse" : "Expand"} ${t ?? "value"}`,
		className: "h-auto max-w-full justify-start gap-0 rounded px-0 py-[var(--json-view-row-space)] text-left font-normal text-[length:inherit] leading-[inherit] whitespace-pre-wrap [&_svg]:size-[0.75em]",
		onClick: () => s(!o),
		children: [/* @__PURE__ */ (0, J.jsx)(he, {
			"aria-hidden": "true",
			className: q("mr-1 shrink-0 text-muted-foreground", o && "rotate-90")
		}), /* @__PURE__ */ (0, J.jsxs)("span", { children: [u, /* @__PURE__ */ (0, J.jsx)("span", {
			className: "text-muted-foreground",
			children: i === "json" ? o ? d : c ? `${d} ${p} ${f}` : `${d}…${f}` : !o || t === void 0 && !a ? p : null
		})] })]
	}), o && /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)("div", {
		className: "border-l border-border pl-[var(--json-view-indent)]",
		children: l.map(([e, t]) => /* @__PURE__ */ (0, J.jsx)(jd, {
			data: t,
			...c ? {} : { name: e },
			sequenceItem: c && i === "yaml",
			depth: n + 1,
			defaultOpenDepth: r,
			format: i
		}, e))
	}), i === "json" && /* @__PURE__ */ (0, J.jsx)("div", {
		className: "text-muted-foreground py-[var(--json-view-row-space)]",
		children: f
	})] })] });
}
function Md({ data: e, format: t }) {
	return e == null ? /* @__PURE__ */ (0, J.jsx)("span", {
		className: "text-muted-foreground italic",
		children: "null"
	}) : typeof e == "string" ? /* @__PURE__ */ (0, J.jsx)("span", {
		className: "text-green-700 [[data-theme=dark]_&]:text-green-400",
		children: t === "yaml" ? Nd(e) : JSON.stringify(e)
	}) : typeof e == "number" || typeof e == "boolean" ? /* @__PURE__ */ (0, J.jsx)("span", {
		className: "text-blue-700 [[data-theme=dark]_&]:text-blue-400",
		children: String(e)
	}) : /* @__PURE__ */ (0, J.jsx)("span", {
		className: "text-muted-foreground",
		children: typeof e == "object" ? Array.isArray(e) ? "[]" : "{}" : String(e)
	});
}
function Nd(e) {
	return cd(e, {
		lineWidth: 0,
		defaultStringType: e.includes("\n") ? "QUOTE_DOUBLE" : "PLAIN"
	}).trimEnd();
}
// ../../../clicky-ui/packages/ui/src/data/TreeNode.tsx
var Pd = 18;
function Fd(e, t, n) {
	if (!t?.length) return e;
	let r = new Map(t.map((e) => [n(e), e])), i = new Set(e.map(n));
	return [...e.map((e) => r.get(n(e)) ?? e), ...t.filter((e) => !i.has(n(e)))];
}
function Id({ node: e, depth: t = 0, expandAll: n = null, forcedOpenKeys: r = null, revealPath: i = null, revealOverrides: a = null, onRevealToggle: o, selected: s = null, defaultOpen: l, getExpanded: u, onExpandedChange: d, getChildren: f, getKey: p, getAriaLabel: m, onSelect: h, renderRow: g, renderDetail: _, renderAfterChildren: v, rowClass: y, indentPx: b = 16, basePaddingPx: x = 8, isSecondary: S, hasMoreChildren: C, loadChildren: w, reloadOnOpen: T = !1, getChildrenVersion: E }) {
	if (u && !d) throw Error("Controlled tree expansion requires onExpandedChange");
	let [D, O] = (0, Y.useState)(null), [k, A] = (0, Y.useState)(!1), [j, M] = (0, Y.useState)(null), N = f(e), P = D === null ? N : Fd(D, N, p), F = (P?.length ?? 0) > 0, I = (C?.(e) ?? !1) && D === null, L = F || I || !!(T && C?.(e)), R = S?.(e) ?? !1, z = F && S != null && P.every((e) => S(e)), B = R || z || n === !0 && I, V = n !== null && !B ? n : l ? l(e, t) : t < 1, [H, U] = (0, Y.useState)(V), ee = (0, Y.useRef)(u ? null : n), te = p(e), W = s != null && p(s) === te, ne = r?.has(te) ?? !1, re = i?.get(te) ?? null, ie = re !== null, ae = a?.get(te), oe = u ? u(e) : H, G = ie && ae === void 0 && !oe, se = ne || G || (ae ?? oe);
	(0, Y.useEffect)(() => {
		n !== null && n !== ee.current && !B && (u ? d(e, n) : U(n)), ee.current = n;
	}, [
		n,
		B,
		u,
		d,
		e
	]);
	let K = (0, Y.useRef)(!1), ce = p(e), le = (0, Y.useRef)(ce), ue = (0, Y.useRef)(0), de = E?.(e), fe = (0, Y.useRef)(de);
	function me() {
		if (!w || K.current) return;
		K.current = !0;
		let t = ++ue.current;
		le.current = ce, A(!0), M(null), w(e).then((e) => {
			le.current === ce && t === ue.current && (O(e), A(!1));
		}).catch((e) => {
			le.current === ce && t === ue.current && (K.current = !1, M(e), A(!1));
		});
	}
	(0, Y.useEffect)(() => {
		se && I && j === null && me();
	}, [
		se,
		I,
		j
	]), (0, Y.useEffect)(() => {
		E && fe.current !== de && (fe.current = de, K.current = !1, se && C?.(e) && me());
	}, [de]), (0, Y.useEffect)(() => () => {
		ue.current++, K.current = !1;
	}, []);
	function ge() {
		if (!L) return;
		if (ie) {
			if (!o) throw Error("A revealed tree path requires onRevealToggle");
			let e = G || !(ae ?? oe);
			e && I && me(), o(te, e);
			return;
		}
		let t = !oe;
		t && T && C?.(e) && (K.current = !1), t && (I || T) && me(), u ? d(e, t) : U(t);
	}
	let _e = y ? y(e, W) : W ? "bg-primary/10 border-l-2 border-primary" : "hover:bg-accent", ve = _?.(e), ye = v?.(e);
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		role: "treeitem",
		"aria-label": m?.(e),
		"aria-expanded": L ? se : void 0,
		"aria-selected": W,
		children: [
			/* @__PURE__ */ (0, J.jsxs)("div", {
				className: q("flex items-center gap-1.5 py-1 px-2 cursor-pointer text-sm", _e),
				style: { paddingLeft: `${t * b + x}px` },
				onClick: (t) => {
					t.stopPropagation(), h?.(e), ge();
				},
				children: [L ? /* @__PURE__ */ (0, J.jsx)("button", {
					type: "button",
					"aria-label": G ? "Show all children" : se ? "Collapse" : "Expand",
					onClick: (e) => {
						e.stopPropagation(), ge();
					},
					className: "flex shrink-0 items-center text-muted-foreground",
					children: /* @__PURE__ */ (0, J.jsx)(c, {
						icon: se && !G ? pe : he,
						className: "text-xs w-3"
					})
				}) : /* @__PURE__ */ (0, J.jsx)("span", {
					className: "w-3 shrink-0",
					"aria-hidden": !0
				}), g({
					node: e,
					depth: t,
					open: se,
					selected: W,
					hasChildren: F,
					toggle: ge,
					loading: k,
					error: j
				})]
			}),
			ve != null && ve !== !1 && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "pb-1 pr-2",
				style: { paddingLeft: `${t * b + x + Pd}px` },
				onClick: (e) => e.stopPropagation(),
				onKeyDown: (e) => e.stopPropagation(),
				children: ve
			}),
			se && F && /* @__PURE__ */ (0, J.jsx)("div", {
				role: "group",
				children: (G ? P.filter((e) => p(e) === re) : P).map((e) => /* @__PURE__ */ (0, J.jsx)(Id, {
					node: e,
					depth: t + 1,
					expandAll: n,
					forcedOpenKeys: r,
					revealPath: i,
					revealOverrides: a,
					...o ? { onRevealToggle: o } : {},
					selected: s,
					getChildren: f,
					getKey: p,
					renderRow: g,
					indentPx: b,
					basePaddingPx: x,
					...l ? { defaultOpen: l } : {},
					...u ? { getExpanded: u } : {},
					...d ? { onExpandedChange: d } : {},
					..._ ? { renderDetail: _ } : {},
					...v ? { renderAfterChildren: v } : {},
					...m ? { getAriaLabel: m } : {},
					...h ? { onSelect: h } : {},
					...y ? { rowClass: y } : {},
					...S ? { isSecondary: S } : {},
					...C ? { hasMoreChildren: C } : {},
					...w ? { loadChildren: w } : {},
					reloadOnOpen: T,
					...E ? { getChildrenVersion: E } : {}
				}, p(e)))
			}),
			ye != null && ye !== !1 && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "pb-1 pr-2",
				style: { paddingLeft: `${t * b + x + Pd}px` },
				onClick: (e) => e.stopPropagation(),
				onKeyDown: (e) => e.stopPropagation(),
				children: ye
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/Tree.tsx
function Ld(e, t, n, r) {
	let i = /* @__PURE__ */ new Map(), a = (e) => {
		if (n(e) === r) return i.set(r, null), !0;
		for (let r of t(e) ?? []) if (a(r)) return i.set(n(e), n(r)), !0;
		return !1;
	};
	for (let t of e) if (a(t)) return i;
	return null;
}
function Rd(e, t) {
	let n = 0, r = [...e];
	for (; r.length > 0;) {
		let e = t(r.pop()) ?? [];
		n += e.length, r.push(...e);
	}
	return n;
}
function zd(e, t = /* @__PURE__ */ new Set(), n = /* @__PURE__ */ new Set()) {
	if (e == null || t.has(e)) return "";
	if (typeof e == "string" || typeof e == "number" || typeof e == "boolean" || typeof e == "bigint") return String(e);
	if (typeof e != "object" || n.has(e)) return "";
	if (n.add(e), Array.isArray(e)) return e.map((e) => zd(e, t, n)).filter(Boolean).join(" ");
	let r = [
		"label",
		"name",
		"title",
		"text",
		"plain",
		"id",
		"content"
	], i = e, a = [];
	for (let e of r) {
		if (!(e in i)) continue;
		let r = zd(i[e], t, n);
		r && a.push(r);
	}
	for (let [e, o] of Object.entries(i)) {
		if (e === "children" || r.includes(e)) continue;
		let i = zd(o, t, n);
		i && a.push(i);
	}
	return a.join(" ");
}
function Bd(e, t) {
	let n = /* @__PURE__ */ new Set();
	return t && n.add(t), zd(e, n).trim().toLowerCase();
}
function Vd(e, t, n, r, i, a) {
	let o = r.trim().toLowerCase();
	if (!o) return {
		roots: e,
		filteredChildren: null,
		forcedOpenKeys: null
	};
	let s = /* @__PURE__ */ new Map(), c = /* @__PURE__ */ new Set();
	function l(e, t) {
		return a ? a(e).trim().toLowerCase() : Bd(e, t.filter((e) => !i?.(e)));
	}
	function u(e) {
		if (i?.(e)) return !1;
		let r = n(e), a = t(e) ?? [], d = a.length > 0 && a.every((e) => i?.(e)), f = l(e, a).includes(o), p = [];
		for (let e of a) u(e) && p.push(e);
		return f ? (a.length > 0 && (s.set(r, a), d || c.add(r)), !0) : p.length > 0 && (s.set(r, p), c.add(r), !0);
	}
	return {
		roots: e.filter((e) => u(e)),
		filteredChildren: s,
		forcedOpenKeys: c
	};
}
function Hd({ roots: e, empty: t, className: n, ariaLabel: r, showControls: i = !0, showSearch: a, searchLabel: o, searchQuery: s, onSearchQueryChange: l, expandAll: u, onExpandAllChange: d, toolbarClassName: f, actions: p, getSearchText: m, revealSelected: h = !1, ...g }) {
	let [_, v] = (0, Y.useState)(null), [y, b] = (0, Y.useState)(""), x = s ?? y, S = (e) => {
		s === void 0 ? b(e) : l?.(e);
	}, C = g.basePaddingPx ?? 8, w = d !== void 0, T = w ? u ?? null : _, E = h && g.selected != null ? g.getKey(g.selected) : null, [D, O] = (0, Y.useState)({
		selectedKey: E,
		revealKey: E,
		clickedKey: null
	}), [k, A] = (0, Y.useState)({
		selectedKey: null,
		overrides: /* @__PURE__ */ new Map(),
		bulk: !1
	});
	if (D.selectedKey !== E) {
		let e = E !== null && E === D.clickedKey;
		O({
			selectedKey: E,
			revealKey: e ? D.revealKey : E,
			clickedKey: null
		}), e || A({
			selectedKey: null,
			overrides: /* @__PURE__ */ new Map(),
			bulk: !1
		});
	}
	let M = D.revealKey, N = (0, Y.useRef)(T);
	(0, Y.useEffect)(() => {
		N.current !== T && (N.current = T, A({
			selectedKey: M,
			overrides: /* @__PURE__ */ new Map(),
			bulk: !0
		}));
	}, [T, M]);
	let P = (0, Y.useMemo)(() => Rd(e, g.getChildren), [e, g.getChildren]), F = a ?? (P > 20 || !!s), I = F || a === !1 ? x : "", L = (0, Y.useMemo)(() => Vd(e, g.getChildren, g.getKey, I, g.isSecondary, m), [
		e,
		g.getChildren,
		g.getKey,
		I,
		g.isSecondary,
		m
	]), R = L.roots, z = L.filteredChildren, B = (0, Y.useMemo)(() => h && M !== null ? Ld(e, g.getChildren, g.getKey, M) : null, [
		h,
		e,
		g.getChildren,
		g.getKey,
		M
	]), V = !I.trim() && !(k.selectedKey === M && k.bulk), H = k.selectedKey === M ? k.overrides : null, U = g.onSelect, ee = U ? (e) => {
		O((t) => ({
			...t,
			clickedKey: g.getKey(e)
		})), U(e);
	} : void 0, te = z ? (e) => z.get(g.getKey(e)) ?? [] : g.getChildren, W = T, ne = I.trim() ? /* @__PURE__ */ (0, J.jsx)("div", {
		className: "px-3 py-4 text-sm text-muted-foreground",
		children: "No matching tree nodes."
	}) : t, re = i && e.length > 0, ie = F || re || p != null, ae = (e) => {
		A({
			selectedKey: M,
			overrides: /* @__PURE__ */ new Map(),
			bulk: !0
		}), w ? d?.(e) : v(e);
	};
	return R.length === 0 && !ie ? /* @__PURE__ */ (0, J.jsx)(J.Fragment, { children: ne ?? null }) : /* @__PURE__ */ (0, J.jsx)("div", {
		className: q("flex flex-col min-h-0", n),
		children: /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "min-h-0 flex-1 overflow-auto",
			children: [ie && /* @__PURE__ */ (0, J.jsxs)("div", {
				role: "presentation",
				className: q("sticky top-0 z-10 flex items-center gap-1.5 border-b border-border/70 bg-background/95 py-1 pr-2 text-sm backdrop-blur supports-[backdrop-filter]:bg-background/80", f),
				style: { paddingLeft: `${C}px` },
				children: [
					F ? /* @__PURE__ */ (0, J.jsx)(c, {
						icon: Xe,
						className: "w-3 shrink-0 text-xs text-muted-foreground"
					}) : /* @__PURE__ */ (0, J.jsx)("span", {
						className: "w-3 shrink-0",
						"aria-hidden": !0
					}),
					F ? /* @__PURE__ */ (0, J.jsxs)("label", {
						className: "flex min-w-0 flex-1 items-center gap-2 rounded-md border border-border bg-background px-2 py-1 text-xs text-foreground shadow-sm",
						children: [/* @__PURE__ */ (0, J.jsx)("input", {
							type: "search",
							value: x,
							onChange: (e) => S(e.target.value),
							placeholder: o ?? "Filter tree",
							className: "h-5 w-full min-w-0 border-0 bg-transparent p-0 text-xs outline-none placeholder:text-placeholder [&::-webkit-search-cancel-button]:appearance-none",
							"aria-label": o ?? "Filter tree nodes"
						}), x && /* @__PURE__ */ (0, J.jsx)("button", {
							type: "button",
							onClick: () => S(""),
							className: "inline-flex items-center rounded p-0.5 text-muted-foreground hover:bg-accent hover:text-accent-foreground",
							"aria-label": "Clear tree filter",
							title: "Clear tree filter",
							children: /* @__PURE__ */ (0, J.jsx)(c, {
								icon: j,
								className: "text-xs"
							})
						})]
					}) : /* @__PURE__ */ (0, J.jsx)("span", { className: "flex-1" }),
					(re || p != null) && /* @__PURE__ */ (0, J.jsxs)("div", {
						className: "ml-auto flex items-center gap-1",
						children: [p, re && /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)("button", {
							type: "button",
							onClick: () => ae(!0),
							"aria-pressed": W === !0,
							"aria-label": "Expand all",
							title: "Expand all",
							disabled: I.trim().length > 0,
							className: q("inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50", W === !0 && "bg-accent text-accent-foreground"),
							children: /* @__PURE__ */ (0, J.jsx)(c, {
								icon: Ye,
								className: "text-sm"
							})
						}), /* @__PURE__ */ (0, J.jsx)("button", {
							type: "button",
							onClick: () => ae(!1),
							"aria-pressed": W === !1,
							"aria-label": "Collapse all",
							title: "Collapse all",
							disabled: I.trim().length > 0,
							className: q("inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50", W === !1 && "bg-accent text-accent-foreground"),
							children: /* @__PURE__ */ (0, J.jsx)(c, {
								icon: Je,
								className: "text-sm"
							})
						})] })]
					})
				]
			}), R.length === 0 ? ne ?? null : /* @__PURE__ */ (0, J.jsx)("div", {
				role: "tree",
				"aria-label": r,
				children: R.map((e) => /* @__PURE__ */ (0, Y.createElement)(Id, {
					...g,
					...ee ? { onSelect: ee } : {},
					key: g.getKey(e),
					node: e,
					expandAll: W,
					forcedOpenKeys: L.forcedOpenKeys,
					revealPath: V ? B : null,
					revealOverrides: H,
					onRevealToggle: (e, t) => A((n) => {
						let r = new Map(n.selectedKey === M ? n.overrides : []);
						return r.set(e, t), {
							selectedKey: M,
							overrides: r,
							bulk: !1
						};
					}),
					getChildren: te
				}))
			})]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/overlay/Modal.tsx
var Ud = {
	sm: "max-w-sm",
	md: "max-w-md",
	lg: "max-w-2xl",
	xl: "max-w-4xl",
	"2xl": "max-w-6xl",
	full: "max-w-[95vw]"
}, Wd = {
	title: "Discard changes?",
	message: "You have unsaved changes. Closing now will discard them.",
	confirmLabel: "Discard",
	cancelLabel: "Keep editing"
}, Gd = Jo("sm"), Kd = "(prefers-reduced-motion: reduce)", qd = 200, Jd = "a[href],button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),[tabindex]:not([tabindex='-1'])";
function Yd(e) {
	return Array.from(e.querySelectorAll(Jd)).filter((e) => {
		let t = window.getComputedStyle(e);
		return !e.hidden && t.display !== "none" && t.visibility !== "hidden";
	});
}
function Xd({ open: e, minimized: t = !1, onClose: n, confirmClose: r = !1, title: a, "aria-label": o, subtitle: s, size: l = "md", closeOnBackdrop: u = !1, closeOnEsc: d = !0, hideClose: f = !1, expandable: p = !0, className: m, headerSlot: h, footer: g, error: _, scrollBody: v = !0, children: y }) {
	let b = (0, Y.useRef)(null), x = (0, Y.useRef)(null), [S, C] = (0, Y.useState)();
	(0, Y.useLayoutEffect)(() => {
		let e = x.current?.textContent?.replace(/\s+/g, " ").trim() || void 0;
		e !== S && C(e);
	});
	let w = (0, Y.useRef)(void 0), T = (0, Y.useRef)(void 0), [E, D] = (0, Y.useState)(e), [O, k] = (0, Y.useState)(!1), A = E && !t, M = e && A, { depth: N } = to(A), [P, F] = (0, Y.useState)(!1), [I, L] = (0, Y.useState)(!1), R = () => {
		r ? L(!0) : n();
	};
	if ((0, Y.useEffect)(() => {
		window.clearTimeout(w.current), window.cancelAnimationFrame(T.current ?? 0);
		let t = i(Gd) && !i(Kd);
		return e ? (D(!0), t ? (k(!1), T.current = window.requestAnimationFrame(() => {
			k(!0);
		})) : k(!0)) : t ? (k(!1), w.current = window.setTimeout(() => D(!1), qd)) : (k(!1), D(!1)), () => {
			window.clearTimeout(w.current), window.cancelAnimationFrame(T.current ?? 0);
		};
	}, [e]), (0, Y.useEffect)(() => {
		e || (F(!1), L(!1));
	}, [e]), (0, Y.useEffect)(() => {
		r || L(!1);
	}, [r]), no(A, () => {
		e && d && (I ? L(!1) : R());
	}), (0, Y.useLayoutEffect)(() => {
		if (!M) return;
		let e = document.activeElement;
		return b.current?.focus({ preventScroll: !0 }), () => {
			e instanceof HTMLElement && e.isConnected && e.focus({ preventScroll: !0 });
		};
	}, [M]), !A) return null;
	let z = /* @__PURE__ */ (0, J.jsx)("div", {
		className: q("fixed inset-0 flex items-center justify-center p-density-2 sm:p-density-4 max-sm:items-stretch max-sm:justify-end max-sm:p-0 max-sm:!bg-transparent", N === 0 ? "bg-black/40" : "bg-black/20"),
		style: { zIndex: Ua.modal + N * Ua.modalStep },
		onClick: u ? R : void 0,
		role: "presentation",
		children: /* @__PURE__ */ (0, J.jsxs)("div", {
			ref: b,
			tabIndex: -1,
			role: "dialog",
			"aria-modal": "true",
			"aria-label": o ?? (typeof a == "string" ? a : S),
			"data-state": O ? "open" : "closed",
			className: q("relative flex w-full flex-col overflow-hidden rounded-lg border border-border bg-background shadow-xl", "max-sm:!h-dvh max-sm:!max-h-dvh max-sm:!max-w-none max-sm:!rounded-none max-sm:!border-0 max-sm:shadow-none", "max-sm:transition-transform max-sm:duration-200 max-sm:ease-out max-sm:motion-reduce:transition-none", "max-sm:data-[state=closed]:translate-x-full max-sm:data-[state=open]:translate-x-0", P ? Ud.full : Ud[l], m),
			style: {
				maxHeight: "calc(100dvh - 2rem)",
				...P ? { height: "calc(100dvh - 2rem)" } : {}
			},
			onClick: (e) => e.stopPropagation(),
			onKeyDown: (e) => {
				if (e.key !== "Tab") return;
				let t = I ? b.current?.querySelector("[role='alertdialog']") : b.current;
				if (!t) return;
				let n = Yd(t), r = n[0], i = n[n.length - 1];
				!r || !i ? (e.preventDefault(), t.focus()) : e.shiftKey && document.activeElement === r ? (e.preventDefault(), i.focus()) : !e.shiftKey && document.activeElement === i && (e.preventDefault(), r.focus());
			},
			children: [
				(a || s || h || p || !f) && /* @__PURE__ */ (0, J.jsxs)("div", {
					className: "shrink-0 px-density-4 py-density-3 border-b border-border max-sm:min-h-14 max-sm:bg-background",
					children: [/* @__PURE__ */ (0, J.jsxs)("div", {
						className: "flex items-center gap-density-2",
						children: [
							!f && /* @__PURE__ */ (0, J.jsx)(fe, {
								type: "button",
								onClick: R,
								variant: "ghost",
								size: "icon",
								"aria-label": "Back",
								className: "-ml-2 text-foreground sm:hidden",
								children: /* @__PURE__ */ (0, J.jsx)(c, { icon: lt })
							}),
							a ? /* @__PURE__ */ (0, J.jsx)("h2", {
								ref: x,
								className: "text-sm font-semibold flex-1 max-sm:text-base max-sm:leading-6",
								children: a
							}) : /* @__PURE__ */ (0, J.jsx)("span", { className: "flex-1" }),
							h,
							p && /* @__PURE__ */ (0, J.jsx)("button", {
								type: "button",
								onClick: () => F((e) => !e),
								"aria-label": P ? "Restore size" : "Expand to fullscreen",
								className: "text-muted-foreground hover:text-foreground max-sm:hidden",
								children: /* @__PURE__ */ (0, J.jsx)(c, { icon: P ? tt : et })
							}),
							!f && /* @__PURE__ */ (0, J.jsx)("button", {
								type: "button",
								onClick: R,
								"aria-label": "Close",
								className: "text-muted-foreground hover:text-foreground max-sm:hidden",
								children: /* @__PURE__ */ (0, J.jsx)(c, { icon: j })
							})
						]
					}), s ? /* @__PURE__ */ (0, J.jsx)("div", {
						className: "mt-density-2",
						children: s
					}) : null]
				}),
				/* @__PURE__ */ (0, J.jsx)("div", {
					"data-slot": "modal-body",
					className: q("min-h-0 flex-1 px-density-4 py-density-3 max-sm:py-density-4", v ? "overflow-auto" : "flex flex-col overflow-hidden"),
					children: y
				}),
				_ != null && _ !== !1 && _ !== "" && /* @__PURE__ */ (0, J.jsx)("div", {
					role: "alert",
					"data-slot": "modal-error",
					className: "mx-density-4 mb-density-3 max-h-32 shrink-0 overflow-auto rounded-md border border-destructive/40 bg-destructive/5 px-density-3 py-density-2 text-sm wrap-break-word text-destructive",
					children: _
				}),
				g && /* @__PURE__ */ (0, J.jsx)("div", {
					"data-slot": "modal-footer",
					className: "shrink-0 px-density-4 py-density-3 border-t border-border max-sm:py-density-4",
					children: g
				}),
				I && /* @__PURE__ */ (0, J.jsx)(Zd, {
					options: r === !0 ? Wd : {
						...Wd,
						...r
					},
					onConfirm: () => {
						L(!1), n();
					},
					onCancel: () => L(!1)
				})
			]
		})
	});
	return typeof document < "u" ? (0, Fe.createPortal)(z, document.body) : null;
}
function Zd({ options: e, onConfirm: t, onCancel: n }) {
	return /* @__PURE__ */ (0, J.jsx)("div", {
		className: "absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-background/80 p-density-4 max-sm:rounded-none",
		role: "alertdialog",
		"aria-modal": "true",
		"aria-label": typeof e.title == "string" ? e.title : "Confirm close",
		onClick: (e) => e.stopPropagation(),
		children: /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "w-full max-w-sm rounded-lg border border-border bg-background p-density-4 shadow-xl",
			children: [
				/* @__PURE__ */ (0, J.jsx)("h3", {
					className: "text-sm font-semibold",
					children: e.title
				}),
				/* @__PURE__ */ (0, J.jsx)("p", {
					className: "mt-density-2 text-sm text-muted-foreground",
					children: e.message
				}),
				/* @__PURE__ */ (0, J.jsxs)("div", {
					className: "mt-density-4 flex justify-end gap-density-2",
					children: [/* @__PURE__ */ (0, J.jsx)(fe, {
						variant: "ghost",
						size: "sm",
						onClick: n,
						children: e.cancelLabel
					}), /* @__PURE__ */ (0, J.jsx)(fe, {
						variant: "destructive",
						size: "sm",
						onClick: t,
						children: e.confirmLabel
					})]
				})
			]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/JSONPathPlayground.tsx
var Qd = 250;
function $d({ open: e, onClose: t, rows: n, value: r, onCommit: i, assignsRoot: a, source: o, evaluate: s, title: l = "JSONPath playground" }) {
	let [u, d] = (0, Y.useState)(r), [f, p] = (0, Y.useState)(0), [m, h] = (0, Y.useState)(null), [g, _] = (0, Y.useState)(null), [y, b] = (0, Y.useState)(!1);
	(0, Y.useEffect)(() => {
		e && (d(r), p((e) => e < n.length ? e : 0));
	}, [
		e,
		r,
		n.length
	]);
	let x = n[f], S = (0, Y.useMemo)(() => ef(x, o), [x, o]), C = (0, Y.useMemo)(() => hs(S, { keyPrefix: `row${f}:${o ?? ""}:` }), [
		S,
		f,
		o
	]);
	(0, Y.useEffect)(() => {
		e && h(C.materializePath(u) ?? null);
	}, [e, C]);
	let w = m?.root ?? o, T = w !== void 0 && w !== o && !a, E = (0, Y.useRef)(0);
	return (0, Y.useEffect)(() => {
		if (!e || !s) return;
		let t = E.current += 1;
		b(!0);
		let n = setTimeout(() => {
			let e = {
				jsonpath: u,
				row: x
			};
			w !== void 0 && (e.source = w), s(e).then((e) => {
				E.current === t && _(e);
			}).catch((e) => {
				E.current === t && _({
					matches: [],
					count: 0,
					error: String(e)
				});
			}).finally(() => {
				E.current === t && b(!1);
			});
		}, Qd);
		return () => clearTimeout(n);
	}, [
		e,
		s,
		u,
		x,
		w
	]), /* @__PURE__ */ (0, J.jsx)(Xd, {
		open: e,
		onClose: t,
		title: l,
		size: "2xl",
		scrollBody: !1,
		className: "h-[80vh]",
		footer: /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "flex items-center justify-end gap-2",
			children: [/* @__PURE__ */ (0, J.jsx)(fe, {
				variant: "outline",
				onClick: t,
				children: "Cancel"
			}), /* @__PURE__ */ (0, J.jsx)(fe, {
				onClick: () => {
					i(u, w === void 0 ? {} : { root: w }), t();
				},
				disabled: u.trim() === "",
				children: "Use path"
			})]
		}),
		children: /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "flex h-full min-h-0 flex-col gap-3",
			children: [n.length > 1 && /* @__PURE__ */ (0, J.jsxs)("label", {
				className: "flex shrink-0 items-center gap-2 text-xs text-muted-foreground",
				children: [
					"Sample row",
					/* @__PURE__ */ (0, J.jsx)("select", {
						"aria-label": "Sample row",
						value: f,
						onChange: (e) => p(Number(e.target.value)),
						className: "h-7 rounded-md border border-input bg-background px-2 text-xs text-foreground",
						children: n.map((e, t) => /* @__PURE__ */ (0, J.jsxs)("option", {
							value: t,
							children: [
								t + 1,
								" of ",
								n.length
							]
						}, t))
					}),
					/* @__PURE__ */ (0, J.jsx)("span", { children: "Fields absent from the first row are only visible in the one that carries them." })
				]
			}), /* @__PURE__ */ (0, J.jsxs)("div", {
				className: "grid min-h-0 flex-1 gap-3 md:grid-cols-2",
				children: [/* @__PURE__ */ (0, J.jsx)("div", {
					className: "flex min-h-0 flex-col overflow-hidden rounded-md border border-border",
					children: /* @__PURE__ */ (0, J.jsx)(Hd, {
						className: "min-h-0 flex-1",
						ariaLabel: "JSON paths",
						roots: C.roots,
						getKey: (e) => e.key,
						getChildren: C.getChildren,
						hasMoreChildren: C.hasMoreChildren,
						loadChildren: C.loadChildren,
						getSearchText: (e) => `${e.path} ${e.summary}`,
						defaultOpen: (e, t) => t < 2,
						selected: m,
						revealSelected: !0,
						onSelect: (e) => {
							e.kind !== "more" && (h(e), d(e.path));
						},
						renderRow: ({ node: e }) => /* @__PURE__ */ (0, J.jsxs)("span", {
							className: "flex min-w-0 flex-1 items-center gap-density-2 font-mono text-xs",
							children: [/* @__PURE__ */ (0, J.jsx)("span", {
								className: q("shrink-0", e.kind === "more" ? "text-muted-foreground" : "text-primary"),
								children: e.kind === "more" ? "…" : e.path
							}), /* @__PURE__ */ (0, J.jsx)("span", {
								className: "truncate text-muted-foreground",
								title: e.summary,
								children: e.summary
							})]
						})
					})
				}), /* @__PURE__ */ (0, J.jsxs)("div", {
					className: "flex min-h-0 flex-col gap-2",
					children: [
						/* @__PURE__ */ (0, J.jsx)(po, {
							"aria-label": "JSONPath expression",
							value: u,
							onChange: d,
							inputClassName: "font-mono text-xs",
							placeholder: "$.messages[0].payload"
						}),
						w !== void 0 && /* @__PURE__ */ (0, J.jsxs)("p", {
							className: q("flex shrink-0 items-start gap-1.5 rounded-md border px-2 py-1.5 text-xs text-foreground", T ? "border-amber-600/40 bg-amber-500/10" : "border-border bg-muted/40"),
							children: [/* @__PURE__ */ (0, J.jsx)(c, {
								icon: T ? M : v,
								className: q("mt-0.5 shrink-0", T ? "text-amber-600" : "text-muted-foreground")
							}), /* @__PURE__ */ (0, J.jsxs)("span", { children: [
								"This path is rooted inside ",
								/* @__PURE__ */ (0, J.jsx)("code", {
									className: "font-mono",
									children: w
								}),
								", a column holding JSON as text.",
								" ",
								w === o ? /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [
									"That is already the column's ",
									/* @__PURE__ */ (0, J.jsx)("code", {
										className: "font-mono",
										children: "Source"
									}),
									"."
								] }) : a ? /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [
									"Applying it sets the column's ",
									/* @__PURE__ */ (0, J.jsx)("code", {
										className: "font-mono",
										children: "Source"
									}),
									" to",
									" ",
									/* @__PURE__ */ (0, J.jsx)("code", {
										className: "font-mono",
										children: w
									}),
									"."
								] }) : /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [
									"The column's ",
									/* @__PURE__ */ (0, J.jsx)("code", {
										className: "font-mono",
										children: "Source"
									}),
									" must be set to",
									" ",
									/* @__PURE__ */ (0, J.jsx)("code", {
										className: "font-mono",
										children: w
									}),
									" for it to resolve."
								] })
							] })]
						}),
						/* @__PURE__ */ (0, J.jsx)(tf, {
							evaluating: y,
							result: g,
							selected: m,
							evaluated: !!s
						})
					]
				})]
			})]
		})
	});
}
function ef(e, t) {
	if (!t) return e;
	if (typeof e != "object" || !e) return;
	let n = e[t];
	if (typeof n != "string") return n;
	try {
		return JSON.parse(n);
	} catch {
		return n;
	}
}
function tf({ evaluating: e, result: t, selected: n, evaluated: r }) {
	return r ? /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex shrink-0 flex-wrap items-center gap-x-3 gap-y-1 text-xs",
		children: [t?.error ? /* @__PURE__ */ (0, J.jsx)("span", {
			className: "font-mono text-destructive",
			children: t.error
		}) : /* @__PURE__ */ (0, J.jsx)("span", {
			className: q(t && t.count === 0 && "text-muted-foreground"),
			children: e && t === null ? "Evaluating…" : `${t?.count ?? 0} match${t?.count === 1 ? "" : "es"}`
		}), t && !t.error && /* @__PURE__ */ (0, J.jsx)("span", {
			className: "text-muted-foreground",
			children: t.filterField ? /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: ["Filterable as ", /* @__PURE__ */ (0, J.jsx)("code", {
				className: "font-mono text-foreground",
				children: t.filterField
			})] }) : "Selects a set — not filterable unless the column declares filter.field"
		})]
	}), /* @__PURE__ */ (0, J.jsx)("div", {
		className: "min-h-0 flex-1 overflow-auto rounded-md border border-border bg-muted/20 p-2 text-xs",
		children: t && !t.error && t.count > 0 ? /* @__PURE__ */ (0, J.jsx)(Ad, {
			data: t.count === 1 ? t.matches[0] : t.matches,
			defaultOpenDepth: 1
		}) : /* @__PURE__ */ (0, J.jsx)("p", {
			className: "text-muted-foreground",
			children: t?.error ? "Fix the expression to see what it selects." : "This path selects nothing in this row."
		})
	})] }) : /* @__PURE__ */ (0, J.jsx)("div", {
		className: "min-h-0 flex-1 overflow-auto rounded-md border border-border bg-muted/20 p-2 text-xs",
		children: n ? /* @__PURE__ */ (0, J.jsx)(Ad, {
			data: n.value,
			defaultOpenDepth: 1
		}) : /* @__PURE__ */ (0, J.jsx)("p", {
			className: "text-muted-foreground",
			children: "Select a node to see its value."
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/TreePickerField.tsx
var nf = 480, rf = 480, af = 200, of = 8;
function sf({ roots: e, getKey: t, getChildren: n, getAriaLabel: r, renderRow: i, getSearchText: a, defaultOpen: o, hasMoreChildren: s, loadChildren: l, isSelectable: u, onSelect: d, closeOnSelect: f = !0, selected: p, revealSelected: m, showControls: h, ariaLabel: _, empty: v, label: b, placeholder: x = "Select…", disabled: S, size: C, className: w, triggerClassName: T, panelClassName: E, renderTrigger: D, renderFooter: O, triggerAriaLabel: k, clearable: A, onClear: j, clearLabel: M }) {
	let [N, P] = (0, Y.useState)(!1), F = A && !S && (p != null || b != null), I = ro(), [L, R] = (0, Y.useState)(null), z = (0, Y.useRef)(null), B = (0, Y.useRef)(null), V = (0, Y.useRef)(null);
	return no(N, () => {
		P(!1), B.current?.focus();
	}), (0, Y.useLayoutEffect)(() => {
		if (!N) {
			R(null);
			return;
		}
		let e = () => {
			let e = z.current;
			if (!e) return;
			let t = e.getBoundingClientRect(), n = window.innerWidth - t.left - of, r = Math.max(t.width, Math.min(nf, n)), i = window.innerHeight - t.bottom - of - 4, a = t.top - of - 4, o = i < af && a > i, s = Math.min(rf, Math.max(o ? a : i, af));
			R({
				top: o ? Math.max(of, t.top - 4 - s) : t.bottom + 4,
				left: t.left,
				width: t.width,
				maxWidth: r,
				maxHeight: s
			});
		};
		return e(), window.addEventListener("scroll", e, !0), window.addEventListener("resize", e), () => {
			window.removeEventListener("scroll", e, !0), window.removeEventListener("resize", e);
		};
	}, [N]), (0, Y.useEffect)(() => {
		if (!N) return;
		let e = (e) => {
			let t = e.target;
			!z.current?.contains(t) && !V.current?.contains(t) && P(!1);
		};
		return document.addEventListener("mousedown", e), () => document.removeEventListener("mousedown", e);
	}, [N]), /* @__PURE__ */ (0, J.jsxs)("div", {
		ref: z,
		"data-clearable-field": !0,
		className: q("relative", w),
		children: [
			D ? D({
				open: N,
				disabled: !!S,
				triggerRef: B,
				toggle: () => {
					S || P((e) => !e);
				}
			}) : /* @__PURE__ */ (0, J.jsxs)("button", {
				ref: B,
				type: "button",
				disabled: S,
				"aria-label": k,
				"aria-haspopup": "tree",
				"aria-expanded": N,
				onClick: () => P((e) => !e),
				className: q("relative flex w-full items-center rounded-md border border-input bg-background text-left text-foreground", C ? g[C] : "h-control-h px-control-px text-sm", "pr-8", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1", "disabled:cursor-not-allowed disabled:opacity-50", T, F && "pr-14"),
				children: [/* @__PURE__ */ (0, J.jsx)("span", {
					className: q("min-w-0 flex-1 truncate", b == null && "text-muted-foreground"),
					children: b ?? x
				}), /* @__PURE__ */ (0, J.jsx)("span", {
					className: "pointer-events-none absolute right-0 flex h-full items-center px-2 text-muted-foreground",
					children: /* @__PURE__ */ (0, J.jsx)(c, {
						icon: pe,
						className: "text-xs"
					})
				})]
			}),
			F && !D && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "absolute inset-y-0 right-7 flex items-center",
				children: /* @__PURE__ */ (0, J.jsx)(y, {
					...M ? { clearLabel: M } : {},
					onClear: () => {
						if (!j) throw Error("A clearable tree picker requires onClear");
						j(), P(!1), B.current?.focus();
					}
				})
			}),
			N && L && (0, Fe.createPortal)(/* @__PURE__ */ (0, J.jsxs)("div", {
				ref: V,
				"data-slot": "tree-picker-popup",
				style: {
					position: "fixed",
					top: L.top,
					left: L.left,
					zIndex: I,
					minWidth: L.width,
					width: "max-content",
					maxWidth: L.maxWidth,
					maxHeight: L.maxHeight
				},
				className: q("flex flex-col overflow-hidden rounded-md border border-border bg-popover shadow-md", E),
				children: [/* @__PURE__ */ (0, J.jsx)("div", {
					className: "min-h-0 flex-1 overflow-auto",
					children: /* @__PURE__ */ (0, J.jsx)(Hd, {
						roots: e,
						getKey: t,
						getChildren: n,
						renderRow: i,
						...r ? { getAriaLabel: r } : {},
						...a ? { getSearchText: a } : {},
						...o ? { defaultOpen: o } : {},
						...s ? { hasMoreChildren: s } : {},
						...l ? { loadChildren: l } : {},
						...p === void 0 ? {} : { selected: p },
						...m === void 0 ? {} : { revealSelected: m },
						...h === void 0 ? {} : { showControls: h },
						..._ === void 0 ? {} : { ariaLabel: _ },
						...v === void 0 ? {} : { empty: v },
						onSelect: (e) => {
							(!u || u(e)) && (d(e), f && P(!1));
						}
					})
				}), O && /* @__PURE__ */ (0, J.jsx)("div", {
					className: "shrink-0 border-t border-border bg-popover",
					children: O({ close: () => P(!1) })
				})]
			}), document.body)
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/JSONPathField.tsx
function cf({ json: e, value: t, onChange: n, isSelectable: r, getPath: i, pickerLabel: a, rows: o, evaluate: s, onSelectPath: l, source: u, disabled: d, className: f, "aria-label": p, ...m }) {
	let [h, g] = (0, Y.useState)(!1), _ = (0, Y.useMemo)(() => e === void 0 ? [] : [ms(e, "$", 0)], [e]), v = (0, Y.useMemo)(() => o ?? (e === void 0 ? [] : [e]), [o, e]), y = a ?? `Browse ${p ?? "JSONPath"} JSON paths`, b = !!d || e === void 0, x = l ?? ((e) => n(e));
	return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)(sf, {
		className: "w-full",
		roots: _,
		getKey: (e) => e.key,
		getChildren: (e) => e.children,
		getSearchText: (e) => `${e.path} ${e.summary}`,
		defaultOpen: (e, t) => t < 2,
		...r ? { isSelectable: r } : {},
		disabled: b,
		onSelect: (e) => n(i?.(e) ?? e.path),
		renderFooter: ({ close: e }) => /* @__PURE__ */ (0, J.jsxs)("button", {
			type: "button",
			onClick: () => {
				e(), g(!0);
			},
			className: "flex w-full items-center gap-2 px-3 py-2 text-left text-xs text-muted-foreground hover:bg-accent hover:text-accent-foreground",
			children: [/* @__PURE__ */ (0, J.jsx)(c, {
				icon: et,
				className: "text-xs"
			}), "Open playground…"]
		}),
		renderRow: ({ node: e }) => /* @__PURE__ */ (0, J.jsxs)("span", {
			className: "flex min-w-0 flex-1 items-center gap-density-2 font-mono text-xs",
			children: [/* @__PURE__ */ (0, J.jsx)("span", {
				className: "shrink-0 text-primary",
				children: e.path
			}), /* @__PURE__ */ (0, J.jsx)("span", {
				className: "truncate text-muted-foreground",
				title: e.summary,
				children: e.summary
			})]
		}),
		renderTrigger: ({ open: e, triggerRef: r, toggle: i }) => /* @__PURE__ */ (0, J.jsx)(po, {
			...m,
			...d === void 0 ? {} : { disabled: d },
			"aria-label": p,
			className: f,
			value: t,
			onChange: n,
			suffix: /* @__PURE__ */ (0, J.jsx)(On, {
				ref: r,
				icon: Tt,
				label: y,
				disabled: b,
				"aria-haspopup": "tree",
				"aria-expanded": e,
				onClick: i
			})
		})
	}), h && /* @__PURE__ */ (0, J.jsx)($d, {
		open: !0,
		onClose: () => g(!1),
		rows: v,
		value: t,
		onCommit: x,
		assignsRoot: l !== void 0,
		...u ? { source: u } : {},
		...s ? { evaluate: s } : {},
		title: `${p ?? "JSONPath"} playground`
	})] });
}
// ../../../clicky-ui/packages/ui/src/components/SegmentedControl.tsx
var lf = {
	sm: "text-xs px-density-2 py-density-1 gap-1",
	md: "h-full px-control-px text-sm gap-1.5",
	lg: "min-h-[4.25rem] min-w-44 flex-1 px-density-3 py-density-3 gap-density-2 text-sm"
};
function uf({ value: e, options: t, onChange: n, size: r = "md", wrap: i = !1, className: a, "aria-label": o }) {
	let s = (0, Y.useRef)([]), l = (0, Y.useId)(), u = t.some((t) => t.id === e), d = (e, r) => {
		let i = t.length, a = e;
		for (let e = 0; e < i; e++) {
			a = (a + r + i) % i;
			let e = t[a];
			if (e && !e.disabled) {
				n(e.id), s.current[a]?.focus();
				return;
			}
		}
	};
	return /* @__PURE__ */ (0, J.jsx)("div", {
		role: "radiogroup",
		"aria-label": o,
		className: q("inline-flex items-center rounded-md bg-muted p-0.5", r === "md" && (i ? "min-h-control-h" : "h-control-h"), r === "lg" && "gap-density-2 bg-transparent p-0", i && "flex-wrap", a),
		children: t.map((t, a) => {
			let o = t.id === e, f = !u && t.implied === !0, p = t.marker ? `${l}-marker-${a}` : void 0;
			return /* @__PURE__ */ (0, J.jsxs)("button", {
				ref: (e) => {
					s.current[a] = e;
				},
				type: "button",
				role: "radio",
				"aria-checked": o,
				"aria-describedby": p,
				"data-implied": f ? "true" : void 0,
				title: t.title,
				disabled: t.disabled,
				tabIndex: o || f ? 0 : -1,
				onClick: () => !t.disabled && n(t.id),
				onKeyDown: (e) => {
					e.key === "ArrowRight" || e.key === "ArrowDown" ? (e.preventDefault(), d(a, 1)) : (e.key === "ArrowLeft" || e.key === "ArrowUp") && (e.preventDefault(), d(a, -1));
				},
				className: q("inline-flex items-center rounded-[5px] border border-transparent font-medium transition-colors", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", "disabled:cursor-not-allowed disabled:opacity-50", r === "lg" && "rounded-lg text-left", t.description ? "whitespace-normal" : "whitespace-nowrap", lf[r], r === "md" && i && "h-auto py-1", o ? r === "lg" ? "border-primary/70 bg-primary/10 text-foreground shadow-sm ring-1 ring-primary/25" : "border-border bg-card text-foreground shadow-sm" : f ? r === "lg" ? "border-dashed border-primary/60 bg-card text-foreground hover:shadow-sm" : "border-dashed border-primary/60 text-foreground hover:bg-background/60" : r === "lg" ? "border-border bg-card text-muted-foreground hover:border-muted-foreground/40 hover:bg-card hover:text-foreground hover:shadow-sm" : "text-foreground/80 hover:border-border hover:bg-background/60 hover:text-foreground", o && t.activeClassName),
				children: [(0, Y.isValidElement)(t.icon) ? /* @__PURE__ */ (0, J.jsx)("span", {
					className: q("inline-flex shrink-0", t.iconClassName),
					children: t.icon
				}) : t.icon && /* @__PURE__ */ (0, J.jsx)(c, {
					...typeof t.icon == "string" ? { name: t.icon } : { icon: t.icon },
					className: q(r === "lg" ? "size-4 shrink-0" : "size-3.5 shrink-0", r === "lg" && (o ? "text-primary" : "text-muted-foreground"), t.iconClassName)
				}), /* @__PURE__ */ (0, J.jsxs)("span", {
					className: q("min-w-0", r === "lg" && "flex flex-col items-start text-left leading-tight"),
					children: [t.marker ? /* @__PURE__ */ (0, J.jsxs)("span", {
						className: "inline-flex min-w-0 max-w-full items-baseline gap-1",
						children: [/* @__PURE__ */ (0, J.jsx)("span", {
							className: "min-w-0 truncate",
							children: t.label
						}), /* @__PURE__ */ (0, J.jsx)("span", {
							id: p,
							"aria-hidden": "true",
							className: "shrink-0 text-[11px] font-normal opacity-70",
							children: t.marker
						})]
					}) : /* @__PURE__ */ (0, J.jsx)("span", {
						className: "min-w-0 truncate",
						children: t.label
					}), t.description && /* @__PURE__ */ (0, J.jsx)("span", {
						className: q("mt-0.5 max-w-56 whitespace-normal break-words text-[11px] font-normal leading-snug", o ? "text-muted-foreground" : "text-muted-foreground/80"),
						children: t.description
					})]
				})]
			}, t.id);
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/Switch.tsx
function df({ checked: e, onChange: t, disabled: n, label: r, className: i, "aria-label": a, ...o }) {
	let s = (0, Y.useId)(), c = /* @__PURE__ */ (0, J.jsx)("button", {
		...o,
		type: "button",
		role: "switch",
		"aria-checked": e,
		"aria-label": r ? void 0 : a,
		"aria-labelledby": r ? s : o["aria-labelledby"],
		disabled: n,
		onClick: () => t(!e),
		className: q("inline-flex shrink-0 items-center rounded-full px-0.5 transition-colors", "h-5 w-9 density-compact:h-4 density-compact:w-7 density-spacious:h-6 density-spacious:w-11", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2", "disabled:cursor-not-allowed disabled:opacity-50", e ? "justify-end bg-primary" : "justify-start bg-input", !r && i),
		children: /* @__PURE__ */ (0, J.jsx)("span", { className: q("pointer-events-none rounded-full bg-background shadow transition-all", "h-4 w-4 density-compact:h-3 density-compact:w-3 density-spacious:h-5 density-spacious:w-5") })
	});
	return r ? /* @__PURE__ */ (0, J.jsxs)("label", {
		className: q("inline-flex cursor-pointer items-center gap-density-2", i),
		children: [c, /* @__PURE__ */ (0, J.jsx)("span", {
			id: s,
			className: "text-sm text-foreground",
			children: r
		})]
	}) : c;
}
// ../../../clicky-ui/packages/ui/src/components/DatePicker.tsx
var ff = (0, Y.forwardRef)(({ value: e = "", onChange: t, className: n, inputClassName: r, buttonClassName: i, openButtonLabel: a = "Open date picker", suffix: o, prefix: s, clearable: l, onClear: u, clearLabel: d, ...f }, p) => {
	let m = (0, Y.useRef)(null), h = l && e !== "" && !f.disabled && !f.readOnly;
	function g(e) {
		m.current = e, typeof p == "function" ? p(e) : p && (p.current = e);
	}
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		"data-jsf-control": !0,
		"data-clearable-field": !0,
		className: q("relative", n),
		children: [
			s && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "absolute inset-y-0 left-1.5 flex items-center",
				children: s
			}),
			/* @__PURE__ */ (0, J.jsx)("input", {
				...f,
				ref: g,
				type: "date",
				value: e,
				className: q("h-control-h w-full rounded-md border border-input bg-background px-control-px pr-8 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring", "[&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none", s && "pl-8", r, (h || o) && "pr-16", h && o && "pr-24"),
				onChange: (e) => t?.(e.target.value)
			}),
			o && /* @__PURE__ */ (0, J.jsx)("div", {
				className: q("absolute inset-y-0 flex items-center", h ? "right-14" : "right-7"),
				children: o
			}),
			h && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "absolute inset-y-0 right-7 flex items-center",
				children: /* @__PURE__ */ (0, J.jsx)(y, {
					...d ? { clearLabel: d } : {},
					onClear: () => {
						u ? u() : t?.(""), m.current?.focus();
					}
				})
			}),
			/* @__PURE__ */ (0, J.jsx)("button", {
				type: "button",
				"aria-label": a,
				disabled: f.disabled || f.readOnly,
				className: q("absolute inset-y-0 right-1 inline-flex items-center text-muted-foreground", i),
				onClick: () => {
					m.current?.focus(), m.current?.showPicker?.();
				},
				children: /* @__PURE__ */ (0, J.jsx)(c, {
					icon: mt,
					className: "text-sm"
				})
			})
		]
	});
});
ff.displayName = "DatePicker";
// ../../../clicky-ui/packages/ui/src/components/DateTimePicker.tsx
var pf = (0, Y.forwardRef)(({ value: e = "", onChange: t, className: n, inputClassName: r, buttonClassName: i, openButtonLabel: a = "Open time picker", suffix: o, prefix: s, clearable: l, onClear: u, clearLabel: d, ...f }, p) => {
	let m = (0, Y.useRef)(null), h = (0, Y.useRef)(null), g = l && e !== "" && !f.disabled && !f.readOnly;
	function _(e) {
		m.current = e, typeof p == "function" ? p(e) : p && (p.current = e);
	}
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		"data-jsf-control": !0,
		"data-clearable-field": !0,
		className: q("relative", n),
		children: [
			s && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "absolute inset-y-0 left-1.5 flex items-center",
				children: s
			}),
			/* @__PURE__ */ (0, J.jsx)("input", {
				...f,
				ref: _,
				type: "text",
				value: e,
				className: q("h-control-h w-full rounded-md border border-input bg-background px-control-px pr-8 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50", s && "pl-8", r, (g || o) && "pr-16", g && o && "pr-24"),
				onChange: (e) => t?.(e.target.value)
			}),
			/* @__PURE__ */ (0, J.jsx)("input", {
				ref: h,
				type: "datetime-local",
				tabIndex: -1,
				"aria-hidden": "true",
				className: "pointer-events-none absolute left-0 top-0 h-0 w-0 opacity-0",
				value: mf(e),
				onChange: (e) => t?.(e.target.value)
			}),
			o && /* @__PURE__ */ (0, J.jsx)("div", {
				className: q("absolute inset-y-0 flex items-center", g ? "right-14" : "right-7"),
				children: o
			}),
			g && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "absolute inset-y-0 right-7 flex items-center",
				children: /* @__PURE__ */ (0, J.jsx)(y, {
					...d ? { clearLabel: d } : {},
					onClear: () => {
						u ? u() : t?.(""), m.current?.focus();
					}
				})
			}),
			/* @__PURE__ */ (0, J.jsx)("button", {
				type: "button",
				"aria-label": a,
				disabled: f.disabled || f.readOnly,
				className: q("absolute inset-y-0 right-1 inline-flex items-center text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50", i),
				onClick: () => {
					m.current?.focus(), h.current?.showPicker?.();
				},
				children: /* @__PURE__ */ (0, J.jsx)(c, {
					icon: mt,
					className: "text-sm"
				})
			})
		]
	});
});
pf.displayName = "DateTimePicker";
function mf(e) {
	let t = e.trim();
	if (!t || t === "now" || t.startsWith("now")) return "";
	let n = /^\d{4}-\d{2}-\d{2}$/.test(t) ? `${t}T00:00` : t;
	return /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}/.test(n) ? n.slice(0, 16) : "";
}
// ../../../clicky-ui/packages/ui/src/components/DateField.tsx
var hf = (0, Y.forwardRef)(({ mode: e = "date", ...t }, n) => e === "datetime" ? /* @__PURE__ */ (0, J.jsx)(pf, {
	ref: n,
	...t
}) : /* @__PURE__ */ (0, J.jsx)(ff, {
	ref: n,
	...t
}));
hf.displayName = "DateField";
// ../../../clicky-ui/packages/ui/src/components/select.tsx
var gf = (0, Y.forwardRef)(({ className: e, options: t, placeholder: n, children: r, clearable: i, onClear: a, clearLabel: o, ...s }, l) => {
	let u = i && s.value !== void 0 && s.value !== "" && !s.disabled;
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		"data-clearable-field": !0,
		className: "relative inline-flex w-full items-center",
		children: [
			/* @__PURE__ */ (0, J.jsxs)("select", {
				ref: l,
				className: q("h-control-h w-full appearance-none rounded-md border border-input bg-background px-control-px pr-8 text-sm text-foreground", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1", "disabled:cursor-not-allowed disabled:opacity-50", e, u && "pr-14"),
				...s,
				children: [n !== void 0 && /* @__PURE__ */ (0, J.jsx)("option", {
					value: "",
					disabled: !0,
					children: n
				}), t ? t.map((e) => /* @__PURE__ */ (0, J.jsx)("option", {
					value: e.value,
					disabled: e.disabled,
					children: e.label
				}, e.value)) : r]
			}),
			u && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "absolute inset-y-0 right-7 flex items-center",
				children: /* @__PURE__ */ (0, J.jsx)(y, {
					...a ? { onClear: a } : {},
					...o ? { clearLabel: o } : {}
				})
			}),
			/* @__PURE__ */ (0, J.jsx)(c, {
				icon: pe,
				className: "pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground"
			})
		]
	});
});
gf.displayName = "Select";
// ../../../clicky-ui/packages/ui/src/data/cells/timestamp-format.ts
var _f = {
	label: "Last 24 hours",
	from: "now-24h",
	to: "now"
}, vf = [
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
	_f,
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
function yf(e) {
	if (e == null) return null;
	if (e instanceof Date) return Number.isNaN(e.getTime()) ? null : e;
	if (typeof e == "number") {
		let t = e < 0xe8d4a51000 ? e * 1e3 : e, n = new Date(t);
		return Number.isNaN(n.getTime()) ? null : n;
	}
	if (typeof e == "string") {
		let t = e.trim();
		if (!t) return null;
		let n = new Date(t);
		return Number.isNaN(n.getTime()) ? null : n;
	}
	return null;
}
var bf = {
	s: 1e3,
	m: 6e4,
	h: 36e5,
	d: 864e5,
	w: 6048e5
}, xf = /^now\s*([+-])\s*(\d+)\s*([smhdwMqy])$/, Sf = /^now\/([wMqy])(?:\s*([+-])\s*(\d+)\s*([wMqy]))?$/;
function Cf(e, t = /* @__PURE__ */ new Date()) {
	if (e == null) return null;
	if (typeof e == "string") {
		let n = e.trim();
		if (!n) return null;
		if (n.toLowerCase() === "now") return new Date(t.getTime());
		let r = n.match(Sf);
		if (r) {
			let [, e, n, i, a] = r, o = Tf(t, e);
			return n && i && a && (o = Ef(o, a, Number(i) * (n === "-" ? -1 : 1))), o;
		}
		let i = n.match(xf);
		if (i) {
			let [, e, n, r] = i, a = r;
			if (wf(a)) return Ef(t, a, Number(n) * (e === "-" ? -1 : 1));
			let o = bf[a.toLowerCase()];
			if (!o) return null;
			let s = Number(n) * o * (e === "-" ? -1 : 1);
			return new Date(t.getTime() + s);
		}
	}
	return yf(e);
}
function wf(e) {
	return e === "M" || e === "q" || e === "y";
}
function Tf(e, t) {
	let n = e.getUTCFullYear(), r = e.getUTCMonth(), i = e.getUTCDate();
	if (t === "w") {
		let e = new Date(Date.UTC(n, r, i)), t = (e.getUTCDay() + 6) % 7;
		return e.setUTCDate(e.getUTCDate() - t), e;
	}
	return t === "M" ? new Date(Date.UTC(n, r, 1)) : t === "q" ? new Date(Date.UTC(n, Math.floor(r / 3) * 3, 1)) : t === "y" ? new Date(Date.UTC(n, 0, 1)) : new Date(e.getTime());
}
function Ef(e, t, n) {
	let r = new Date(e.getTime());
	if (t === "w") r.setUTCDate(r.getUTCDate() + n * 7);
	else if (t === "M") r.setUTCMonth(r.getUTCMonth() + n);
	else if (t === "q") r.setUTCMonth(r.getUTCMonth() + n * 3);
	else if (t === "y") r.setUTCFullYear(r.getUTCFullYear() + n);
	else {
		let e = bf[t.toLowerCase()];
		e && r.setTime(r.getTime() + n * e);
	}
	return r;
}
var Df = 6e4, Of = 864e5;
function kf(e) {
	if (e.length === 0) return "iso";
	let t = Infinity, n = -Infinity;
	for (let r of e) {
		let e = r.getTime();
		e < t && (t = e), e > n && (n = e);
	}
	let r = n - t;
	if (r <= Df) return "relative";
	let i = new Date(t), a = new Date(n);
	return i.getFullYear() === a.getFullYear() && i.getMonth() === a.getMonth() && i.getDate() === a.getDate() ? "time" : i.getFullYear() === a.getFullYear() && r <= Of * 365 ? "short" : "iso";
}
function Af(e, t, n = /* @__PURE__ */ new Date()) {
	switch (t) {
		case "relative": return Nf(e, n);
		case "time": return e.toLocaleTimeString(void 0, {
			hour: "2-digit",
			minute: "2-digit",
			second: "2-digit"
		});
		case "short": return e.toLocaleString(void 0, {
			month: "short",
			day: "numeric",
			hour: "2-digit",
			minute: "2-digit"
		});
		default: return jf(e);
	}
}
function jf(e) {
	return `${e.getFullYear()}-${String(e.getMonth() + 1).padStart(2, "0")}-${String(e.getDate()).padStart(2, "0")} ${String(e.getHours()).padStart(2, "0")}:${String(e.getMinutes()).padStart(2, "0")}`;
}
var Mf = [
	["y", 31536e3],
	["w", 604800],
	["d", 86400],
	["h", 3600],
	["m", 60]
];
function Nf(e, t = /* @__PURE__ */ new Date(), n = {}) {
	let r = t.getTime() - e.getTime(), i = r < 0, a = Math.abs(r);
	if (n.style === "token") {
		let e = Math.floor(a / 1e3), t = Mf.find(([, t]) => e >= t);
		if (!t) return "now";
		let n = `${Math.floor(e / t[1])}${t[0]}`;
		return i ? `in ${n}` : n;
	}
	let o = Math.round(a / 1e3), s;
	return s = o < 1 ? "just now" : o < 60 ? `${o}s` : o < 3600 ? `${Math.round(o / 60)}m` : o < 86400 ? `${Math.round(o / 3600)}h` : `${Math.round(o / 86400)}d`, s === "just now" ? s : i ? `in ${s}` : `${s} ago`;
}
function Pf(e, t = /* @__PURE__ */ new Date()) {
	let n = yf(e);
	if (!n) return typeof e == "string" ? e : "";
	let r = n.toLocaleString(void 0, {
		year: "numeric",
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit"
	}), i = Nf(n, t);
	return i === "just now" ? r : `${r} (${i})`;
}
function Ff(e, t) {
	switch (e) {
		case "absolute": return "iso";
		case "relative": return "relative";
		case "time": return "time";
		default: return t;
	}
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@flanksource/icons/1.0.65/7031cbad50e7a415a67e1d34190ddf1950507b712b763ed01cbc2ec536738903/node_modules/@flanksource/icons/lib/iconContext.mjs
var If = {
	color: void 0,
	size: void 0,
	className: void 0,
	style: void 0,
	attr: void 0
}, Lf = Y.createContext && /*#__PURE__*/ Y.createContext(If), Rf = [
	"attr",
	"size",
	"title",
	"square"
];
function zf(e, t) {
	if (e == null) return {};
	var n = Bf(e, t), r, i;
	if (Object.getOwnPropertySymbols) {
		var a = Object.getOwnPropertySymbols(e);
		for (i = 0; i < a.length; i++) r = a[i], !(t.indexOf(r) >= 0) && Object.prototype.propertyIsEnumerable.call(e, r) && (n[r] = e[r]);
	}
	return n;
}
function Bf(e, t) {
	if (e == null) return {};
	for (var n = {}, r = Object.keys(e), i, a = 0; a < r.length; a++) i = r[a], !(t.indexOf(i) >= 0) && (n[i] = e[i]);
	return n;
}
function Vf() {
	return Vf = Object.assign ? Object.assign.bind() : function(e) {
		for (var t = 1; t < arguments.length; t++) {
			var n = arguments[t];
			for (var r in n) Object.prototype.hasOwnProperty.call(n, r) && (e[r] = n[r]);
		}
		return e;
	}, Vf.apply(this, arguments);
}
function Hf(e, t) {
	var n = Object.keys(e);
	if (Object.getOwnPropertySymbols) {
		var r = Object.getOwnPropertySymbols(e);
		t && (r = r.filter(function(t) {
			return Object.getOwnPropertyDescriptor(e, t).enumerable;
		})), n.push.apply(n, r);
	}
	return n;
}
function Uf(e) {
	for (var t = 1; t < arguments.length; t++) {
		var n = arguments[t] == null ? {} : arguments[t];
		t % 2 ? Hf(Object(n), !0).forEach(function(t) {
			Wf(e, t, n[t]);
		}) : Object.getOwnPropertyDescriptors ? Object.defineProperties(e, Object.getOwnPropertyDescriptors(n)) : Hf(Object(n)).forEach(function(t) {
			Object.defineProperty(e, t, Object.getOwnPropertyDescriptor(n, t));
		});
	}
	return e;
}
function Wf(e, t, n) {
	return t = Gf(t), t in e ? Object.defineProperty(e, t, {
		value: n,
		enumerable: !0,
		configurable: !0,
		writable: !0
	}) : e[t] = n, e;
}
function Gf(e) {
	var t = Kf(e, "string");
	return typeof t == "symbol" ? t : String(t);
}
function Kf(e, t) {
	if (typeof e != "object" || !e) return e;
	var n = e[Symbol.toPrimitive];
	if (n !== void 0) {
		var r = n.call(e, t || "default");
		if (typeof r != "object") return r;
		throw TypeError("@@toPrimitive must return a primitive value.");
	}
	return (t === "string" ? String : Number)(e);
}
function qf(e) {
	return e && e.map((e, t) => /*#__PURE__*/ Y.createElement(e.tag, Uf({ key: t }, e.attr), qf(e.child)));
}
function Jf(e) {
	var t = (t) => /*#__PURE__*/ Y.createElement(Xf, Vf({ attr: Uf({}, e.attr) }, t), qf(e.child));
	return t.viewBox = e.attr?.viewBox, t;
}
function Yf(e) {
	if (!e) return !1;
	var t = e.trim().split(/[\s,]+/);
	if (t.length !== 4) return !1;
	var n = parseFloat(t[2]), r = parseFloat(t[3]);
	return !isFinite(n) || !isFinite(r) || r <= 0 ? !1 : n / r >= 2;
}
function Xf(e) {
	var t = (t) => {
		var { attr: n, size: r, title: i, square: a } = e, o = zf(e, Rf), s = r || t.size || "1em", c = a ?? !Yf(n?.viewBox), l;
		return t.className && (l = t.className), e.className && (l = (l ? l + " " : "") + e.className), /*#__PURE__*/ Y.createElement("svg", Vf({
			stroke: "currentColor",
			fill: "currentColor",
			strokeWidth: "0"
		}, t.attr, n, o, {
			className: l,
			style: Uf(Uf({ color: e.color || t.color }, t.style), e.style),
			height: s,
			width: c ? s : void 0,
			xmlns: "http://www.w3.org/2000/svg"
		}), i && /*#__PURE__*/ Y.createElement("title", null, i), e.children);
	};
	return Lf === void 0 ? t(If) : /*#__PURE__*/ Y.createElement(Lf.Consumer, null, (e) => t(e));
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@flanksource/icons/1.0.65/7031cbad50e7a415a67e1d34190ddf1950507b712b763ed01cbc2ec536738903/node_modules/@flanksource/icons/mi/index.mjs
function Zf(e) {
	return Jf({
		tag: "svg",
		attr: {
			role: "img",
			viewBox: "0 0 32 32"
		},
		child: [{
			tag: "path",
			attr: {
				d: "M8.738 7.754c-.039-.04-.078-.074-.117-.11-.805-.777-1.43-1.675-1.809-2.734-.105-.297-.187-.601-.171-.918 0-.031 0-.062.004-.094.027-.41.304-.62.707-.523.128.035.25.082.367.14.441.223.8.551 1.133.907a7.23 7.23 0 0 1 1.386 2.14c.008.02.016.04.028.055.011.012.023.028.035.035a11.277 11.277 0 0 1 5.183-1.593 1.4 1.4 0 0 0-.027-.141 7.179 7.179 0 0 1-.168-2.363c.031-.477.125-.946.281-1.395.078-.234.196-.453.352-.648.058-.07.129-.133.203-.184.156-.11.36-.11.516 0 .16.106.289.254.37.43.16.293.274.61.337.937a7.12 7.12 0 0 1 .136 2.114 5.824 5.824 0 0 1-.222 1.273c.457.086.91.148 1.355.254.89.207 1.754.52 2.57.934.407.203.793.445 1.2.675.011-.027.031-.062.047-.097a6.95 6.95 0 0 1 2.105-2.993c.223-.19.477-.34.754-.44a1.24 1.24 0 0 1 .238-.056c.399-.043.567.207.598.508a1.93 1.93 0 0 1-.05.664 5.59 5.59 0 0 1-.688 1.582c-.434.7-.95 1.332-1.606 1.844a.503.503 0 0 0-.066.063c.64.593 1.21 1.253 1.699 1.976a.97.97 0 0 1-.105.02c-.68 0-1.36 0-2.036.004a.26.26 0 0 1-.203-.11 9.507 9.507 0 0 0-5.644-2.98 9.178 9.178 0 0 0-1.938-.075 9.356 9.356 0 0 0-5.281 2.036 8.92 8.92 0 0 0-1.086 1.015.31.31 0 0 1-.25.114c-.645-.004-1.293-.004-1.941-.004h-.137c.039-.153.437-.7.887-1.188.336-.363.691-.707 1.054-1.074M25.25 22.266c-.45.632-.965 1.214-1.535 1.738l.12.101a6.95 6.95 0 0 1 2.157 3.079c.11.296.16.609.14.925a.966.966 0 0 1-.05.235.459.459 0 0 1-.52.324 1.39 1.39 0 0 1-.433-.129 3.584 3.584 0 0 1-.59-.375 6.83 6.83 0 0 1-2.105-2.988c-.012-.031-.028-.067-.047-.121-.512.34-1.047.64-1.602.898a11.5 11.5 0 0 1-3.52.965l.028.14c.195.762.262 1.551.191 2.333a4.993 4.993 0 0 1-.28 1.437 4.899 4.899 0 0 1-.27.559.787.787 0 0 1-.149.191c-.254.262-.558.266-.808-.004-.098-.11-.18-.23-.25-.36-.196-.37-.297-.773-.368-1.183a7.43 7.43 0 0 1-.09-1.691c.02-.414.083-.828.188-1.227l.023-.105-.011-.047a11.328 11.328 0 0 1-5.184-1.598c-.023.055-.047.106-.07.157a7.073 7.073 0 0 1-1.977 2.648 2.375 2.375 0 0 1-.77.445.766.766 0 0 1-.417.047.447.447 0 0 1-.348-.289c-.09-.219-.074-.445-.043-.672.05-.316.145-.62.277-.914a7.158 7.158 0 0 1 1.696-2.414c.031-.027.058-.055.09-.086l.023-.039a11.54 11.54 0 0 1-1.824-2.004.95.95 0 0 1 .148-.012c.672-.003 1.344 0 2.02-.003a.289.289 0 0 1 .234.105 9.549 9.549 0 0 0 2.524 1.848 8.91 8.91 0 0 0 3.18.945c3.015.293 5.609-.621 7.784-2.742a.488.488 0 0 1 .383-.156c.625.007 1.254 0 1.883 0h.16Zm-2.805-9.61c.18 0 .352.012.52-.004.191-.015.332.051.469.18.808.746 1.625 1.48 2.437 2.219l.137.117.137-.117c.835-.762 1.675-1.528 2.511-2.293.07-.07.168-.11.27-.102.207.008.414 0 .629 0v6.742c-.11.036-1.59.04-1.77.004v-3.414l-.031-.015c-.578.527-1.156 1.05-1.746 1.59-.59-.532-1.172-1.06-1.758-1.583l-.031.012c-.004.285 0 .57 0 .856v2.57H22.46c-.031-.113-.047-6.379-.016-6.762m-16.195.008h1.742c.035.113.043 6.59.008 6.754H6.254c-.012-.43-.004-.855-.004-1.285-.004-.422 0-.848 0-1.281H4.086v2.543c-.133.039-1.621.043-1.77.007v-6.738h1.762v2.383c.129.035 1.98.043 2.168.008l.004-.582Zm3.836 6.75v-6.73c.101-.032 3.937-.047 4.187-.012v1.433c-.054.004-.113.012-.171.012h-2.239v1.117h2.133v1.485h-2.113c-.035.125-.047 1.011-.016 1.258.051.003.11.011.168.011h2.238v1.426Zm6.187.004c-.03-.18-.023-6.629.008-6.754h1.739v4.977c.074.004.132.011.19.011h2.25v1.766Zm0 0",
				fill: "rgb(5.882353%, 8.627451%, 53.72549%)"
			},
			child: []
		}]
	})(e);
}
function Qf(e) {
	return Jf({
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
					d: "M13.46 10.195h10.653m-15.25 0h2.668m5.801 3.871h6.781m-11.379 0h2.668m1.93 3.868h6.781m-11.379 0h2.668M13.461 21.8h10.652m-15.25 0h2.668",
					fill: "none",
					stroke: "rgb(100%, 100%, 100%)",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: "1.451"
				},
				child: []
			}
		]
	})(e);
}
function $f(e) {
	return Jf({
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
function ep(e) {
	return Jf({
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
function tp(e) {
	return Jf({
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
function np(e) {
	return Jf({
		tag: "svg",
		attr: {
			role: "img",
			viewBox: "0 0 32.978 32"
		},
		child: [
			{
				tag: "path",
				attr: {
					d: "M16.379.848a2.083 2.083 0 0 0-.793.199L4.723 6.238a2.06 2.06 0 0 0-1.121 1.399L.922 19.293c-.14.61.004 1.254.398 1.742l7.52 9.348c.394.488.992.773 1.62.773l12.056-.004c.632 0 1.23-.28 1.625-.773l7.515-9.348a2.042 2.042 0 0 0 .403-1.742L29.375 7.633a2.055 2.055 0 0 0-1.125-1.399L17.387 1.047a2.082 2.082 0 0 0-1.008-.2m0 0",
					fill: "rgb(19.607843%, 42.352941%, 89.803922%)"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M16.375.004a2.22 2.22 0 0 0-.84.21L4.07 5.696a2.171 2.171 0 0 0-1.187 1.473L.055 19.473a2.164 2.164 0 0 0 .297 1.668c.039.062.082.117.125.172l7.937 9.87a2.21 2.21 0 0 0 1.715.817l12.723-.004c.668 0 1.3-.3 1.714-.816l7.934-9.867c.414-.516.57-1.196.422-1.84L30.09 7.164a2.187 2.187 0 0 0-1.184-1.473L17.438.215a2.216 2.216 0 0 0-1.063-.211m.004.844c.348-.02.695.05 1.008.199L28.25 6.234c.57.274.98.79 1.125 1.399l2.684 11.656a2.042 2.042 0 0 1-.403 1.742l-7.515 9.348a2.065 2.065 0 0 1-1.625.773l-12.055.004c-.633 0-1.23-.285-1.621-.773l-7.52-9.348a2.043 2.043 0 0 1-.398-1.742l2.68-11.656a2.068 2.068 0 0 1 1.12-1.399l10.864-5.191c.246-.117.52-.188.793-.2m0 0",
					fill: "rgb(100%, 100%, 100%)"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M17.332 14.066h6.781",
					fill: "none",
					stroke: "rgb(100%, 100%, 100%)",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: "1.451"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M13.46 10.195h10.653",
					fill: "none",
					stroke: "rgb(100%, 100%, 100%)",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: "1.452"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M8.863 10.195h2.668",
					fill: "none",
					stroke: "rgb(100%, 100%, 100%)",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: "1.451"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M12.402 14.688c0-2.133 1.848-3.864 4.13-3.864 2.28 0 4.128 1.73 4.128 3.864h-1.93c0-1.067-.984-1.93-2.199-1.93-1.215 0-2.195.863-2.195 1.93Zm0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0",
					fill: "rgb(100%, 100%, 100%)",
					fillRule: "evenodd",
					stroke: "rgb(19.607843%, 42.352941%, 89.803922%)",
					strokeLinecap: "square",
					strokeMiterlimit: "10",
					strokeWidth: ".484"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M17.332 17.934h6.781m-10.652 3.87h10.652m-15.25 0h2.668",
					fill: "none",
					stroke: "rgb(100%, 100%, 100%)",
					strokeLinejoin: "round",
					strokeMiterlimit: "10",
					strokeWidth: "1.451"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M11.387 14.574h10.27v6.688h-10.27Zm0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0h0m0 0",
					fill: "rgb(100%, 100%, 100%)",
					fillRule: "evenodd",
					stroke: "rgb(19.607843%, 42.352941%, 89.803922%)",
					strokeLinecap: "square",
					strokeMiterlimit: "10",
					strokeWidth: ".484"
				},
				child: []
			},
			{
				tag: "path",
				attr: {
					d: "M17.535 17.918a1.016 1.016 0 1 1-2.032-.001 1.016 1.016 0 0 1 2.032.001m0 0",
					fill: "rgb(19.607843%, 42.352941%, 89.803922%)"
				},
				child: []
			}
		]
	})(e);
}
function rp(e) {
	return Jf({
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
function ip(e) {
	return Jf({
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
				d: "M16.375.004a2.22 2.22 0 0 0-.84.21L4.07 5.696a2.171 2.171 0 0 0-1.187 1.473L.055 19.473a2.164 2.164 0 0 0 .297 1.668c.039.062.082.117.125.172l7.937 9.87a2.21 2.21 0 0 0 1.715.817l12.723-.004c.668 0 1.3-.3 1.714-.816l7.934-9.867c.414-.516.57-1.196.422-1.84L30.09 7.164a2.187 2.187 0 0 0-1.184-1.473L17.438.215a2.216 2.216 0 0 0-1.063-.211m.004.844c.348-.02.695.05 1.008.199L28.25 6.234c.566.274.98.786 1.125 1.399l2.684 11.656a2.042 2.042 0 0 1-.403 1.742l-7.515 9.348a2.065 2.065 0 0 1-1.625.773l-12.055.004c-.633 0-1.23-.285-1.621-.773l-7.52-9.348a2.043 2.043 0 0 1-.398-1.742l2.68-11.656a2.06 2.06 0 0 1 1.12-1.399l10.864-5.191c.246-.117.52-.188.793-.2m.113 6.122-7.39 3.281v4.93c0 4.554 3.152 8.816 7.39 9.851 4.235-1.035 7.387-5.297 7.387-9.851v-4.93Zm0 3.476c.957 0 1.817.578 2.18 1.461a2.365 2.365 0 0 1-.512 2.578 2.359 2.359 0 0 1-2.574.512 2.367 2.367 0 0 1 .906-4.55m4.43 9.78c0 .06-.02.114-.05.165-.056.082-.106.164-.165.238a8.226 8.226 0 0 1-4.215 3.027 8.205 8.205 0 0 1-4.387-3.265.301.301 0 0 1-.05-.164v-1.313c0-1.812 2.957-2.727 4.437-2.727 1.477 0 4.43.915 4.43 2.727Zm0 0",
				fill: "rgb(100%, 100%, 100%)"
			},
			child: []
		}]
	})(e);
}
function ap(e) {
	return Jf({
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
function op(e) {
	return Jf({
		tag: "svg",
		attr: {
			role: "img",
			viewBox: "0 0 32 32"
		},
		child: [{
			tag: "path",
			attr: { d: "m0 0 15.941 32L32 0Zm17.84 6.398h1.855v1.86H17.84Zm-3.695 7.47h-1.88v-1.872h1.856Zm0-2.786h-1.88V9.211h1.856Zm0-2.785h-1.88V6.426h1.856Zm2.785 8.348h-1.86v-1.872h1.86Zm0-2.786h-1.86v-1.863h1.86Zm0-2.78h-1.86V9.21h1.86Zm0-2.786h-1.86V6.426h1.86Zm.91.93h1.855v1.855H17.84Zm0 4.636v-1.863h1.894v1.871Zm0 0" },
			child: []
		}]
	})(e);
}
// ../../../clicky-ui/packages/ui/src/components/icon-menu-picker.tsx
function sp({ value: e, onChange: t, options: n, ariaLabel: r, triggerTitle: i, footer: a, showLabel: o, className: s, triggerClassName: l, menuClassName: u }, d) {
	let [f, p] = (0, Y.useState)(!1), m = ro(), [h, g] = (0, Y.useState)(null), _ = (0, Y.useRef)([]), v = (0, Y.useRef)(n.map((e) => e.label));
	v.current = n.map((e) => e.label);
	let y = (0, Y.useMemo)(() => n.findIndex((t) => t.value === e), [n, e]), { refs: b, floatingStyles: x, context: S } = Ca({
		open: f,
		onOpenChange: p,
		placement: "bottom-start",
		whileElementsMounted: Ee,
		middleware: [
			ye(6),
			Oe({ padding: 8 }),
			ke({ padding: 8 })
		]
	}), { getReferenceProps: C, getFloatingProps: w, getItemProps: T } = Ta([
		_a(S),
		xa(S, { escapeKey: !1 }),
		Pa(S, { role: "menu" }),
		Ma(S, {
			listRef: _,
			activeIndex: h,
			selectedIndex: y,
			onNavigate: g,
			loop: !0
		}),
		Fa(S, {
			listRef: v,
			activeIndex: h,
			onMatch: g
		})
	]);
	no(f, () => {
		p(!1), b.domReference.current instanceof HTMLElement && b.domReference.current.focus();
	});
	let E = ti([b.setReference, d]), D = (e) => {
		t(e), p(!1);
	}, O = n[y];
	if (!O) throw Error(`IconMenuPicker: value "${e}" not found in options for ${r}`);
	let k = i ?? `${r}: ${O.label}`;
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		ref: E,
		className: q("relative inline-flex", o && "w-full", s),
		children: [o ? /* @__PURE__ */ (0, J.jsxs)(fe, {
			type: "button",
			variant: "ghost",
			size: "default",
			"aria-label": r,
			"aria-haspopup": "menu",
			"aria-expanded": f,
			title: k,
			className: q("w-full justify-start text-muted-foreground hover:text-foreground", l),
			...C(),
			children: [
				/* @__PURE__ */ (0, J.jsx)(c, {
					icon: O.icon,
					className: "shrink-0 text-foreground"
				}),
				/* @__PURE__ */ (0, J.jsx)("span", {
					className: "min-w-0 flex-1 truncate text-left capitalize",
					children: `${r}: ${O.label}`
				}),
				/* @__PURE__ */ (0, J.jsx)(c, {
					icon: pe,
					className: "shrink-0 text-muted-foreground"
				})
			]
		}) : /* @__PURE__ */ (0, J.jsx)(On, {
			icon: O.icon,
			label: r,
			title: k,
			"aria-haspopup": "menu",
			"aria-expanded": f,
			className: l,
			...C()
		}), f && /* @__PURE__ */ (0, J.jsx)(na, { children: /* @__PURE__ */ (0, J.jsx)(pa, {
			context: S,
			modal: !1,
			initialFocus: y,
			children: /* @__PURE__ */ (0, J.jsxs)("div", {
				ref: b.setFloating,
				role: "menu",
				"aria-label": r,
				style: {
					...x,
					zIndex: m
				},
				className: q("min-w-[12rem] rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-lg shadow-black/5", u),
				...w(),
				children: [n.map((t, n) => {
					let r = t.value === e;
					return /* @__PURE__ */ (0, J.jsxs)("button", {
						ref: (e) => {
							_.current[n] = e;
						},
						type: "button",
						role: "menuitemradio",
						"aria-checked": r,
						tabIndex: n === h ? 0 : -1,
						className: q("flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-sm transition-colors", "hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground focus:outline-none", r && "text-foreground"),
						...T({
							onClick: () => D(t.value),
							onKeyDown: (e) => {
								(e.key === "Enter" || e.key === " ") && (e.preventDefault(), D(t.value));
							}
						}),
						children: [
							/* @__PURE__ */ (0, J.jsx)(c, {
								icon: t.icon,
								className: "shrink-0 text-muted-foreground"
							}),
							/* @__PURE__ */ (0, J.jsx)("span", {
								className: "min-w-0 flex-1 truncate capitalize",
								children: t.label
							}),
							r ? /* @__PURE__ */ (0, J.jsx)(c, {
								icon: bt,
								className: "shrink-0 text-foreground"
							}) : /* @__PURE__ */ (0, J.jsx)("span", {
								className: "inline-block size-4 shrink-0",
								"aria-hidden": !0
							})
						]
					}, t.value);
				}), a ? /* @__PURE__ */ (0, J.jsx)("div", {
					className: "mt-1 border-t border-border/60 px-2 py-1.5 text-[11px] text-muted-foreground",
					children: a
				}) : null]
			})
		}) })]
	});
}
var cp = (0, Y.forwardRef)(sp);
// ../../../clicky-ui/packages/ui/src/components/OnePasswordSelector.model.ts
function lp(e) {
	if (e === "") return;
	if (!e.startsWith("op://") || e.includes("#") || [
		0,
		10,
		13
	].some((t) => e.includes(String.fromCharCode(t)))) throw Error("invalid 1Password reference: expected op://<vault>/<item>/<field>");
	let t = e.slice(5).split("/");
	if (t.length < 3 || t.some((e) => e === "")) throw Error("invalid 1Password reference: expected op://<vault>/<item>/<field>");
	try {
		return {
			vault: decodeURIComponent(t[0]),
			item: decodeURIComponent(t[1]),
			field: t.slice(2).map(decodeURIComponent).join("/")
		};
	} catch (e) {
		throw Error(`invalid 1Password reference encoding: ${e instanceof Error ? e.message : String(e)}`);
	}
}
function up(e) {
	if (!e.vault || !e.item || !e.field) return "";
	let t = e.field.split("/").map(encodeURIComponent).join("/");
	return `op://${encodeURIComponent(e.vault)}/${encodeURIComponent(e.item)}/${t}`;
}
// ../../../clicky-ui/packages/ui/src/components/OnePasswordSelector.tsx
var dp = {
	vault: "",
	item: "",
	field: ""
};
function fp(e) {
	return e instanceof Error ? e.message : String(e);
}
function pp(e, t) {
	return e.find((e) => e.id === t || e.name === t)?.id;
}
function mp(e, t) {
	return e.find((e) => e.id === t || e.name === t)?.name ?? t;
}
function hp(e) {
	return e.map((e) => ({
		value: e.id,
		label: e.name
	}));
}
function gp({ value: e, onChange: t, loadVaults: n, loadItems: r, loadFields: i, allowCustomValue: a = !0 }) {
	let [o, s] = (0, Y.useState)(() => lp(e) ?? dp), [c, l] = (0, Y.useState)([]), [u, d] = (0, Y.useState)([]), [f, p] = (0, Y.useState)([]), [m, h] = (0, Y.useState)({
		vaults: !0,
		items: !1,
		fields: !1
	}), [g, _] = (0, Y.useState)("");
	(0, Y.useEffect)(() => {
		s(lp(e) ?? dp);
	}, [e]), (0, Y.useEffect)(() => {
		let e = !0;
		return h((e) => ({
			...e,
			vaults: !0
		})), n().then((t) => {
			e && l(t);
		}).catch((t) => {
			e && _(fp(t));
		}).finally(() => {
			e && h((e) => ({
				...e,
				vaults: !1
			}));
		}), () => {
			e = !1;
		};
	}, [n]);
	let v = (0, Y.useMemo)(() => pp(c, o.vault), [o.vault, c]);
	(0, Y.useEffect)(() => {
		if (!v) {
			d([]);
			return;
		}
		let e = !0;
		return h((e) => ({
			...e,
			items: !0
		})), r(v).then((t) => {
			e && d(t);
		}).catch((t) => {
			e && _(fp(t));
		}).finally(() => {
			e && h((e) => ({
				...e,
				items: !1
			}));
		}), () => {
			e = !1;
		};
	}, [r, v]);
	let y = (0, Y.useMemo)(() => pp(u, o.item), [u, o.item]);
	(0, Y.useEffect)(() => {
		if (!v || !y) {
			p([]);
			return;
		}
		let e = !0;
		return h((e) => ({
			...e,
			fields: !0
		})), i(v, y).then((t) => {
			e && p(t);
		}).catch((t) => {
			e && _(fp(t));
		}).finally(() => {
			e && h((e) => ({
				...e,
				fields: !1
			}));
		}), () => {
			e = !1;
		};
	}, [
		y,
		i,
		v
	]);
	let b = (0, Y.useMemo)(() => f.map((e) => ({
		value: e.id,
		label: e.label,
		...e.section ? { group: e.section } : {}
	})), [f]);
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "space-y-2",
		children: [/* @__PURE__ */ (0, J.jsxs)("div", {
			className: "grid gap-2 md:grid-cols-3",
			children: [
				/* @__PURE__ */ (0, J.jsx)(ss, {
					ariaLabel: "1Password vault",
					placeholder: "Vault",
					value: o.vault,
					onChange: (e) => {
						_(""), s({
							vault: e,
							item: "",
							field: ""
						}), t("");
					},
					options: hp(c),
					loading: m.vaults,
					allowCustomValue: a
				}),
				/* @__PURE__ */ (0, J.jsx)(ss, {
					ariaLabel: "1Password item",
					placeholder: "Item",
					value: o.item,
					onChange: (e) => {
						_(""), s((t) => ({
							...t,
							item: e,
							field: ""
						})), t("");
					},
					options: hp(u),
					loading: m.items,
					allowCustomValue: a
				}),
				/* @__PURE__ */ (0, J.jsx)(ss, {
					ariaLabel: "1Password field",
					placeholder: "Field",
					value: o.field,
					onChange: (e) => {
						_(""), s((t) => ({
							...t,
							field: e
						})), t(f.find((t) => t.id === e)?.reference ?? up({
							vault: mp(c, o.vault),
							item: mp(u, o.item),
							field: e
						}));
					},
					options: b,
					loading: m.fields,
					allowCustomValue: a
				})
			]
		}), g && /* @__PURE__ */ (0, J.jsx)("p", {
			role: "alert",
			className: "text-sm text-destructive",
			children: g
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/components/SecretKeySelector.tsx
var _p = [
	"secret",
	"configmap",
	"value"
], vp = "max-w-full min-w-0 shrink", yp = `${vp} basis-40 grow-0`, bp = `${vp} w-full basis-auto grow-0 @min-[22rem]/secret:basis-40`, xp = `${vp} w-full basis-auto grow @min-[22rem]/secret:basis-72`, Sp = `${vp} basis-64 grow`;
function Cp({ value: e, onChange: t, loadResources: n, loadKeyPreview: r, loadHelmValues: i, loadServiceAccounts: a, onePassword: o, sources: s, allowLiteral: c = !0, strict: l = !1, className: u }) {
	let d = e?.kind, f = (0, Y.useMemo)(() => {
		let e = s ?? (c ? _p : ["secret", "configmap"]);
		return d && !e.includes(d) ? [d, ...e] : e;
	}, [
		s,
		c,
		d
	]), p = e?.kind ?? f[0] ?? "secret", m = p === "secret" || p === "configmap" || p === "helm", h = p === "serviceaccount", g = p === "onepassword", _ = p === "value", v = m ? p : "secret", y = e && (e.kind === "serviceaccount" || Tp(e)) ? e.name : "", b = e && Tp(e) ? e.key : "", x = e && e.kind === "value" ? e.value : "", S = e && e.kind === "onepassword" ? e.ref : "", [C, w] = (0, Y.useState)([]), [T, E] = (0, Y.useState)(!1), [D, O] = (0, Y.useState)([]), [k, A] = (0, Y.useState)(!1);
	(0, Y.useEffect)(() => {
		let e = m && n ? () => n(v) : h && a ? a : null;
		if (!e) {
			w([]);
			return;
		}
		let t = !1;
		return E(!0), e().then((e) => !t && w(e)).finally(() => !t && E(!1)), () => {
			t = !0;
		};
	}, [
		n,
		a,
		v,
		m,
		h
	]), (0, Y.useEffect)(() => {
		if (!m || !y || !r) {
			O([]);
			return;
		}
		let e = !1;
		return A(!0), r(v, y).then((t) => !e && O(t)).catch(() => !e && O([])).finally(() => !e && A(!1)), () => {
			e = !0;
		};
	}, [
		r,
		v,
		y,
		m
	]);
	let [j, M] = (0, Y.useState)(), [N, P] = (0, Y.useState)();
	(0, Y.useEffect)(() => {
		if (M(void 0), P(void 0), p !== "helm" || !y || !i) return;
		let e = !1;
		return i(y).then((t) => !e && M(t)).catch((t) => !e && P(t instanceof Error ? t.message : String(t))), () => {
			e = !0;
		};
	}, [
		i,
		p,
		y
	]);
	let F = (0, Y.useMemo)(() => C.map((e) => ({
		value: e.name,
		label: e.name
	})), [C]), I = (0, Y.useMemo)(() => C.find((e) => e.name === y), [C, y]), L = (0, Y.useMemo)(() => kp(I?.keys ?? [], D, b), [
		I,
		b,
		D
	]), R = l && (m && !!n || h && !!a) && !!y && !T && !I, z = l && m && !!b && !!I && !T && !(I.keys ?? []).includes(b), B = (e) => {
		switch (e) {
			case "value": return t({
				kind: "value",
				value: ""
			});
			case "onepassword": return t({
				kind: "onepassword",
				ref: ""
			});
			case "serviceaccount": return t({
				kind: "serviceaccount",
				name: ""
			});
			default: return t({
				kind: e,
				name: "",
				key: ""
			});
		}
	}, V = (e) => t(h ? e ? {
		kind: "serviceaccount",
		name: e
	} : void 0 : e ? {
		kind: v,
		name: e,
		key: b
	} : void 0), H = (e) => y ? t({
		kind: v,
		name: y,
		key: e
	}) : void 0, U = (e) => t({
		kind: "value",
		value: e
	}), ee = (e) => t({
		kind: "onepassword",
		ref: e
	}), te = (0, Y.useMemo)(() => f.map((e) => ({
		value: e,
		label: Op[e],
		icon: (0, Y.createElement)(Dp[e], { className: "size-4" })
	})), [f]), W = (0, Y.useMemo)(() => f.map((e) => ({
		value: e,
		label: Op[e],
		icon: Dp[e]
	})), [f]);
	return /* @__PURE__ */ (0, J.jsx)("div", {
		"data-slot": "secret-key-selector",
		className: q("@container/secret w-full min-w-0", u),
		children: /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "flex w-full max-w-full min-w-0 flex-nowrap items-center gap-2",
			"data-slot": "secret-fields",
			children: [
				/* @__PURE__ */ (0, J.jsx)("div", {
					className: q(yp, "@max-md:hidden"),
					"data-slot": "secret-source-combobox",
					children: /* @__PURE__ */ (0, J.jsx)(ss, {
						ariaLabel: "Secret value source",
						options: te,
						value: p,
						onChange: (e) => B(e),
						allowCustomValue: !1,
						required: !0
					})
				}),
				/* @__PURE__ */ (0, J.jsx)("div", {
					className: "hidden shrink-0 @max-md:!block",
					"data-slot": "secret-source-menu",
					children: /* @__PURE__ */ (0, J.jsx)(cp, {
						value: p,
						onChange: B,
						options: W,
						ariaLabel: "Secret value source",
						triggerClassName: "size-control-h rounded-md border border-input bg-background"
					})
				}),
				_ ? /* @__PURE__ */ (0, J.jsx)(wp, {
					value: x,
					onChange: U,
					placeholder: "Static value…",
					ariaLabel: "Static value"
				}) : g ? o ? /* @__PURE__ */ (0, J.jsx)("div", {
					className: Sp,
					"data-slot": "secret-onepassword-field",
					children: /* @__PURE__ */ (0, J.jsx)(gp, {
						value: S,
						onChange: ee,
						...o,
						allowCustomValue: !0
					})
				}) : /* @__PURE__ */ (0, J.jsx)(wp, {
					value: S,
					onChange: ee,
					placeholder: "op://vault/item/field",
					ariaLabel: "1Password reference"
				}) : h ? /* @__PURE__ */ (0, J.jsx)("div", {
					className: Sp,
					"data-slot": "secret-serviceaccount-field",
					children: /* @__PURE__ */ (0, J.jsx)(ss, {
						options: F,
						value: y,
						onChange: V,
						allowCustomValue: !0,
						loading: T,
						invalid: R,
						placeholder: "Service account…"
					})
				}) : /* @__PURE__ */ (0, J.jsxs)("div", {
					className: q("flex flex-col items-center gap-2 @min-[22rem]/secret:flex-row", vp, "grow"),
					"data-slot": "secret-reference-fields",
					children: [/* @__PURE__ */ (0, J.jsx)("div", {
						className: bp,
						"data-slot": "secret-resource-field",
						children: /* @__PURE__ */ (0, J.jsx)(ss, {
							options: F,
							value: y,
							onChange: V,
							allowCustomValue: !0,
							loading: T,
							invalid: R,
							placeholder: `Select ${v}…`
						})
					}), /* @__PURE__ */ (0, J.jsx)("div", {
						className: xp,
						"data-slot": "secret-key-field",
						children: v === "helm" ? /* @__PURE__ */ (0, J.jsx)(cf, {
							"aria-label": "Helm value JSONPath",
							value: b,
							onChange: H,
							disabled: !y,
							invalid: z || !!N,
							title: N,
							placeholder: y ? Ep(v) : "—",
							...j === void 0 ? {} : { json: j }
						}) : /* @__PURE__ */ (0, J.jsx)(ss, {
							options: L,
							value: b,
							onChange: H,
							allowCustomValue: !0,
							loading: k,
							invalid: z,
							placeholder: y ? Ep(v) : "—"
						})
					})]
				})
			]
		})
	});
}
function wp({ value: e, onChange: t, placeholder: n, ariaLabel: r }) {
	return /* @__PURE__ */ (0, J.jsx)("div", {
		className: Sp,
		"data-slot": "secret-value-fields",
		children: /* @__PURE__ */ (0, J.jsx)("input", {
			type: "text",
			value: e,
			onChange: (e) => t(e.target.value),
			placeholder: n,
			"aria-label": r,
			className: q("h-control-h w-full min-w-0 rounded border border-input bg-background px-control-px text-sm", "outline-none focus-visible:ring-2 focus-visible:ring-ring")
		})
	});
}
function Tp(e) {
	return e.kind === "secret" || e.kind === "configmap" || e.kind === "helm";
}
function Ep(e) {
	return e === "helm" ? "jsonpath…" : "Key…";
}
var Dp = {
	secret: np,
	configmap: Qf,
	helm: Zf,
	serviceaccount: ip,
	onepassword: op,
	value: ft
}, Op = {
	secret: "Secret",
	configmap: "ConfigMap",
	helm: "Helm",
	serviceaccount: "Service Account",
	onepassword: "1Password",
	value: "Value"
};
function kp(e, t, n) {
	let r = new Map(t.map((e) => [e.key, e.value])), i = e.map((e) => ({
		value: e,
		label: r.has(e) ? `${e} — ${r.get(e)}` : e
	}));
	return n && !i.some((e) => e.value === n) ? [{
		value: n,
		label: n
	}, ...i] : i;
}
// ../../../clicky-ui/packages/ui/src/components/SecretKeySelector.model.ts
function Ap(e) {
	if (!e) return "";
	switch (e.kind) {
		case "value": return e.value;
		case "onepassword": return e.ref;
		case "serviceaccount": return e.name ? `serviceaccount://${e.name}` : "";
		default: return `${e.kind}://${e.name}/${e.key}`;
	}
}
function jp(e) {
	if (typeof e == "string" && e !== "") {
		if (e.startsWith("op://")) return {
			kind: "onepassword",
			ref: e
		};
		if (e.startsWith("serviceaccount://")) return {
			kind: "serviceaccount",
			name: e.slice(17)
		};
		for (let t of [
			"secret",
			"configmap",
			"helm"
		]) {
			let n = `${t}://`;
			if (e.startsWith(n)) {
				let [r = "", i = ""] = e.slice(n.length).split("/", 2);
				return {
					kind: t,
					name: r,
					key: i
				};
			}
		}
		return {
			kind: "value",
			value: e
		};
	}
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-filter-input.tsx
function Mp({ value: e, onChange: t }) {
	let n = (0, Y.useId)();
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex min-w-0 flex-col gap-1",
		children: [/* @__PURE__ */ (0, J.jsx)("label", {
			htmlFor: n,
			className: "text-xs text-muted-foreground",
			children: "Filter fields and values"
		}), /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "relative",
			children: [
				/* @__PURE__ */ (0, J.jsx)(c, {
					icon: Xe,
					className: "pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground"
				}),
				/* @__PURE__ */ (0, J.jsx)("input", {
					id: n,
					type: "text",
					"aria-label": "Filter fields",
					placeholder: "Search names, values, or tags…",
					value: e,
					onChange: (e) => t(e.target.value),
					onKeyDown: (e) => e.stopPropagation(),
					className: "h-7 w-full rounded-md border border-input bg-background pl-7 pr-7 text-xs text-foreground outline-none placeholder:text-placeholder focus-visible:ring-2 focus-visible:ring-ring"
				}),
				e !== "" && /* @__PURE__ */ (0, J.jsx)("button", {
					type: "button",
					"aria-label": "Clear filter",
					onClick: () => t(""),
					className: "absolute right-1.5 top-1/2 inline-flex -translate-y-1/2 items-center justify-center text-muted-foreground hover:text-foreground",
					children: /* @__PURE__ */ (0, J.jsx)(c, {
						icon: j,
						className: "text-xs"
					})
				})
			]
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-schema-extensions.ts
var Np = "schema declares x-clicky-extensions but no extensionLoader was provided; render it with PageForm or pass extensionLoader", Pp = [];
function Fp(e) {
	if (!("x-clicky-extensions" in e)) return e;
	let { "x-clicky-extensions": t, ...n } = e;
	return n;
}
function Ip(e, t) {
	let n = JSON.stringify(e ?? []), r = e !== void 0 && e.length > 0, i = (0, Y.useRef)(t);
	i.current = t;
	let [a, o] = (0, Y.useState)(), s = r && t !== void 0;
	if ((0, Y.useEffect)(() => {
		if (!s) return;
		let e = !1, t = JSON.parse(n);
		return Promise.all(t.map((e, t) => i.current(e, t))).then((t) => !e && o({
			key: n,
			extensions: t
		}), (t) => !e && o({
			key: n,
			error: t
		})), () => {
			e = !0;
		};
	}, [n, s]), !r) return Pp;
	if (!t) throw Error(Np);
	if (a?.key === n) {
		if ("error" in a) throw a.error;
		return a.extensions;
	}
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-section-actions.tsx
var Lp = (0, Y.createContext)(null);
function Rp(e) {
	let t = (0, Y.useContext)(Lp);
	return t?.owner === e ? t.element : void 0;
}
// ../../../clicky-ui/packages/ui/src/data/properties-context.ts
var zp = (0, Y.createContext)(null), Bp = 6, Vp = 120;
function Hp({ trigger: e, children: t, placement: n = "top", delay: r = 0, arrow: i = !0, className: a, cardClassName: o }) {
	let [s, c] = (0, Y.useState)(!1), [l, u] = (0, Y.useState)(null), d = ro(), f = (0, Y.useRef)(null), p = (0, Y.useRef)(null), m = (0, Y.useRef)(null), h = (0, Y.useRef)(null), g = () => {
		h.current !== null && (window.clearTimeout(h.current), h.current = null);
	};
	function _() {
		g(), r > 0 ? (m.current !== null && window.clearTimeout(m.current), m.current = window.setTimeout(() => c(!0), r)) : c(!0);
	}
	function v() {
		m.current !== null && (window.clearTimeout(m.current), m.current = null), g(), h.current = window.setTimeout(() => c(!1), Vp);
	}
	(0, Y.useLayoutEffect)(() => {
		if (!s) return;
		let e = () => {
			let e = f.current, t = p.current;
			if (!e || !t) return;
			let r = e.getBoundingClientRect(), i = t.getBoundingClientRect(), a = 0, o = 0;
			switch (n) {
				case "top":
					a = r.top - i.height - Bp, o = r.left + r.width / 2 - i.width / 2;
					break;
				case "bottom":
					a = r.bottom + Bp, o = r.left + r.width / 2 - i.width / 2;
					break;
				case "left":
					a = r.top + r.height / 2 - i.height / 2, o = r.left - i.width - Bp;
					break;
				case "right": a = r.top + r.height / 2 - i.height / 2, o = r.right + Bp;
			}
			o = Math.min(Math.max(4, o), window.innerWidth - i.width - 4), a = Math.min(Math.max(4, a), window.innerHeight - i.height - 4), u({
				top: a,
				left: o
			});
		};
		return e(), window.addEventListener("scroll", e, !0), window.addEventListener("resize", e), () => {
			window.removeEventListener("scroll", e, !0), window.removeEventListener("resize", e);
		};
	}, [
		s,
		n,
		t
	]), (0, Y.useEffect)(() => () => {
		m.current !== null && window.clearTimeout(m.current), h.current !== null && window.clearTimeout(h.current);
	}, []);
	let y = s && typeof document < "u" ? (0, Fe.createPortal)(/* @__PURE__ */ (0, J.jsxs)("div", {
		ref: p,
		role: "tooltip",
		onMouseEnter: g,
		onMouseLeave: v,
		style: {
			...l ? {
				position: "fixed",
				top: l.top,
				left: l.left
			} : {
				position: "fixed",
				top: -9999,
				left: -9999,
				visibility: "hidden"
			},
			zIndex: d
		},
		className: q("rounded-md border border-border bg-background px-2.5 py-1.5 text-[11px] shadow-lg", "whitespace-nowrap", o),
		children: [t, i && l && /* @__PURE__ */ (0, J.jsx)(Up, { placement: n })]
	}), document.body) : null;
	return /* @__PURE__ */ (0, J.jsxs)("span", {
		ref: f,
		className: q("relative inline-flex items-center", a),
		onMouseEnter: _,
		onMouseLeave: v,
		onFocus: _,
		onBlur: v,
		children: [e, y]
	});
}
function Up({ placement: e }) {
	let t = "absolute h-1.5 w-1.5 rotate-45 border bg-background border-border";
	switch (e) {
		case "top": return /* @__PURE__ */ (0, J.jsx)("span", {
			"aria-hidden": !0,
			className: q(t, "left-1/2 -translate-x-1/2 -bottom-0.5 border-l-0 border-t-0")
		});
		case "bottom": return /* @__PURE__ */ (0, J.jsx)("span", {
			"aria-hidden": !0,
			className: q(t, "left-1/2 -translate-x-1/2 -top-0.5 border-r-0 border-b-0")
		});
		case "left": return /* @__PURE__ */ (0, J.jsx)("span", {
			"aria-hidden": !0,
			className: q(t, "top-1/2 -translate-y-1/2 -right-0.5 border-l-0 border-b-0")
		});
		case "right": return /* @__PURE__ */ (0, J.jsx)("span", {
			"aria-hidden": !0,
			className: q(t, "top-1/2 -translate-y-1/2 -left-0.5 border-r-0 border-t-0")
		});
	}
}
// ../../../clicky-ui/packages/ui/src/components/form-layout-context.ts
var Wp = (0, Y.createContext)(/* @__PURE__ */ new Map()), Gp = [
	"json",
	"date",
	"date-time",
	"textarea",
	"md",
	"percent",
	"connection"
], Kp = [
	"stacked",
	"inline",
	"properties"
], qp = [
	"inline",
	"stack",
	"table"
], Jp = [
	"Title",
	"PreferencesMenu",
	"FieldFilter",
	"ErrorSummary",
	"NoMatches",
	"SectionHeading",
	"FieldError"
];
function Yp(e, t) {
	let n = e.get(t.name);
	if (n) throw Error(`form extension "${t.extension}" cannot register ${t.kind} "${t.name}": already registered by extension "${n.extension}"`);
	e.set(t.name, {
		extension: t.extension,
		value: t.value
	});
}
function Xp(e) {
	return new Map([...e].map(([e, t]) => [e, t.value]));
}
var Zp = (e, t) => e.includes(t);
function Qp(e) {
	let t = /* @__PURE__ */ new Map(), n = /* @__PURE__ */ new Map(), r = /* @__PURE__ */ new Map(), i = /* @__PURE__ */ new Map(), a = [], o = [];
	return e.forEach((e, s) => {
		if (typeof e?.name != "string" || e.name === "") throw Error(`form extension at index ${s} must have a name`);
		let c = e.name;
		a.push(...e.pre ?? []), o.push(...e.post ?? []);
		for (let [n, r] of Object.entries(e.formats ?? {})) {
			if (Zp(Gp, n)) throw Error(`form extension "${c}" cannot register reserved format "${n}"`);
			Yp(t, {
				kind: "format",
				name: n,
				extension: c,
				value: r
			});
		}
		for (let [t, r] of Object.entries(e.lookups ?? {})) Yp(n, {
			kind: "lookup",
			name: t,
			extension: c,
			value: r
		});
		for (let [t, n] of Object.entries(e.layouts ?? {})) {
			if (Zp(Kp, t) || Zp(qp, t)) throw Error(`form extension "${c}" cannot register reserved layout "${t}"`);
			Yp(r, {
				kind: "layout",
				name: t,
				extension: c,
				value: n
			});
		}
		for (let [t, n] of Object.entries(e.slots ?? {})) {
			if (!Zp(Jp, t)) throw Error(`form extension "${c}" registers unknown slot "${t}"`);
			Yp(i, {
				kind: "slot",
				name: t,
				extension: c,
				value: n
			});
		}
	}), {
		pre: a,
		post: o,
		formats: new Map([...t].map(([e, t]) => [e, {
			extension: t.extension,
			render: t.value
		}])),
		lookups: Xp(n),
		layouts: Xp(r),
		slots: Object.fromEntries([...i].map(([e, t]) => [e, t.value]))
	};
}
function $p(e) {
	return Zp(Kp, e);
}
function em(e) {
	let t = e.layouts.get(e.name);
	if (!t) throw Error(`${e.description} is not registered by any form extension (registered: ${[...e.layouts.keys()].join(", ") || "none"})`);
	return t;
}
function tm(e, t) {
	$p(e) || em({
		name: e,
		layouts: t,
		description: `form layout "${e}"`
	});
}
function nm(e) {
	let { fieldKey: t, fieldLayout: n, formMode: r, layouts: i } = e;
	if (n !== void 0 && !Zp(qp, n)) return em({
		name: n,
		layouts: i,
		description: `layout "${n}" on field "${t}"`
	});
	if (!$p(r)) return em({
		name: r,
		layouts: i,
		description: `form layout "${r}"`
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-help-links.tsx
function rm({ field: e }) {
	return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [e.helper, e.schema["x-help"]?.links?.map((e) => {
		let t = new URL(e.href);
		if (t.protocol !== "https:" && t.protocol !== "http:") throw Error(`Unsupported documentation URL: ${e.href}`);
		return /* @__PURE__ */ (0, J.jsx)("a", {
			className: "ml-1 text-primary underline underline-offset-2",
			href: t.href,
			target: "_blank",
			rel: "noopener noreferrer",
			children: e.label
		}, e.href);
	})] });
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-help-model.ts
function im(e) {
	return !!(e.helper || e.schema["x-help"]?.links?.length);
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-layout.tsx
function am(e) {
	let { layout: t, size: n, columns: r, columnMinWidth: i, columnsMaxWidth: a, className: o, children: s } = e, c = (0, Y.useContext)(zp), l = (0, Y.useContext)(Wp);
	if (!$p(t.mode)) {
		tm(t.mode, l);
		let { Grid: n } = l.get(t.mode);
		return /* @__PURE__ */ (0, J.jsx)(n, { ...e });
	}
	if (t.mode === "properties") return /* @__PURE__ */ (0, J.jsx)("div", {
		className: q("min-w-0 overflow-hidden rounded-md border border-border", c && "col-span-full grid grid-cols-subgrid rounded-none border-0", o),
		children: s
	});
	if (t.mode === "inline") {
		let e = t.labelMaxWidth ?? "40ch", r = t.valueMaxWidth ?? "600px";
		return /* @__PURE__ */ (0, J.jsx)("div", {
			className: q("grid", k[n], o),
			style: { gridTemplateColumns: `fit-content(${e}) minmax(0, ${r})` },
			children: s
		});
	}
	return r === "auto" ? /* @__PURE__ */ (0, J.jsx)("div", {
		className: q("grid gap-x-4", A[n], o),
		style: {
			gridTemplateColumns: `repeat(auto-fill, minmax(${i ?? "15rem"}, 1fr))`,
			...a ? { maxWidth: a } : {}
		},
		children: s
	}) : r && r > 1 ? /* @__PURE__ */ (0, J.jsx)("div", {
		className: q("grid gap-x-3", A[n], o),
		style: { gridTemplateColumns: `repeat(${r}, minmax(0, 1fr))` },
		children: s
	}) : /* @__PURE__ */ (0, J.jsx)("div", {
		className: q("grid", A[n], o),
		children: s
	});
}
function om({ label: e, value: t, helper: n, error: r, errorElement: i, layout: a, size: o, maxWidth: s }) {
	return a.mode === "inline" ? /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "col-span-2 grid grid-cols-subgrid items-start gap-x-3 gap-y-0.5",
		children: [
			/* @__PURE__ */ (0, J.jsx)("div", {
				className: q("flex min-w-0 items-center", ue[o]),
				children: e
			}),
			/* @__PURE__ */ (0, J.jsx)("div", {
				className: "min-w-0",
				children: t
			}),
			n && /* @__PURE__ */ (0, J.jsx)("p", {
				className: "col-start-2 text-xs text-muted-foreground",
				children: n
			}),
			r && /* @__PURE__ */ (0, J.jsx)("p", {
				className: "col-start-2 text-xs text-destructive",
				children: r
			}),
			i && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "col-start-2",
				children: i
			})
		]
	}) : /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex w-full flex-col gap-1",
		style: { maxWidth: s ?? a.valueMaxWidth ?? "600px" },
		children: [
			e,
			t,
			n && /* @__PURE__ */ (0, J.jsx)("p", {
				className: "text-xs text-muted-foreground",
				children: n
			}),
			r && /* @__PURE__ */ (0, J.jsx)("p", {
				className: "text-xs text-destructive",
				children: r
			}),
			i
		]
	});
}
function sm({ field: e, fieldId: t, size: n, helpDisplay: r }) {
	return /* @__PURE__ */ (0, J.jsxs)("label", {
		htmlFor: t,
		className: q("flex min-w-0 items-center gap-2 font-medium", O[n], e.labelClassName),
		children: [
			/* @__PURE__ */ (0, J.jsx)(o, {
				icon: e.labelIcon,
				className: "shrink-0 text-[15px] text-muted-foreground"
			}),
			/* @__PURE__ */ (0, J.jsx)("span", {
				className: "truncate",
				title: e.label === e.key ? void 0 : e.key,
				children: e.label
			}),
			e.required && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "shrink-0 text-destructive",
				children: "*"
			}),
			e.badge && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "shrink-0 rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
				children: e.badge
			}),
			r === "hover" && im(e) && /* @__PURE__ */ (0, J.jsx)(cm, {
				label: e.label,
				helper: /* @__PURE__ */ (0, J.jsx)(rm, { field: e })
			})
		]
	});
}
function cm({ label: e, helper: t }) {
	return /* @__PURE__ */ (0, J.jsx)(Hp, {
		placement: "top",
		trigger: /* @__PURE__ */ (0, J.jsx)("button", {
			type: "button",
			"aria-label": `About ${e}`,
			className: "shrink-0 cursor-help text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
			children: /* @__PURE__ */ (0, J.jsx)(c, {
				icon: vt,
				className: "text-sm"
			})
		}),
		cardClassName: "max-w-xs whitespace-normal text-left text-xs font-normal text-foreground",
		children: t
	});
}
function lm({ field: e, label: t = e.label, size: n, helpDisplay: r }) {
	let { required: i, badge: a, labelIcon: s } = e;
	return /* @__PURE__ */ (0, J.jsxs)("span", {
		className: q("flex min-w-0 items-center gap-2", O[n]),
		children: [
			/* @__PURE__ */ (0, J.jsx)(o, {
				icon: s,
				className: "shrink-0 text-[15px] text-muted-foreground"
			}),
			/* @__PURE__ */ (0, J.jsx)("span", {
				className: "min-w-0 truncate",
				children: t
			}),
			i && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "shrink-0 text-destructive",
				children: "*"
			}),
			a && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "shrink-0 rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
				children: a
			}),
			r === "hover" && im(e) && /* @__PURE__ */ (0, J.jsx)(cm, {
				label: t,
				helper: /* @__PURE__ */ (0, J.jsx)(rm, { field: e })
			})
		]
	});
}
function um({ heading: e, actions: t, actionsOwner: n, size: r, helper: i, helpDisplay: a, children: o }) {
	let s = a === "hover" && !!i, [c, l] = (0, Y.useState)(null), u = /* @__PURE__ */ (0, J.jsx)(Lp.Provider, {
		value: n !== void 0 && c ? {
			owner: n,
			element: c
		} : null,
		children: o
	});
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: q("col-span-full flex min-w-0 max-w-full flex-col", d[r]),
		children: [
			e != null && /* @__PURE__ */ (0, J.jsxs)("div", {
				"data-jsf-section-heading": !0,
				className: q("flex min-w-0 items-center gap-2 border-b border-border pb-1 font-semibold", O[r]),
				children: [e, (t != null || n !== void 0) && /* @__PURE__ */ (0, J.jsxs)("span", {
					className: "ml-auto flex shrink-0 items-center gap-1 font-normal",
					children: [t, n !== void 0 && /* @__PURE__ */ (0, J.jsx)("span", {
						ref: l,
						"data-jsf-section-actions": !0,
						className: "flex items-center empty:hidden"
					})]
				})]
			}),
			i && !s && /* @__PURE__ */ (0, J.jsx)("p", {
				className: "min-w-0 break-words text-xs text-muted-foreground",
				children: i
			}),
			u
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-error-display.tsx
function dm({ id: e, errors: t }) {
	return /* @__PURE__ */ (0, J.jsx)("span", {
		id: e,
		"data-jsf-errors": !0,
		children: t.map((e, t) => /* @__PURE__ */ (0, J.jsx)("span", {
			className: "block",
			children: e.message
		}, `${e.schemaPath ?? ""}:${e.keyword ?? ""}:${e.message}:${t}`))
	});
}
function fm({ id: e, errors: t, className: n }) {
	return t.length === 0 ? null : /* @__PURE__ */ (0, J.jsx)("p", {
		className: q("text-xs text-destructive", n),
		children: /* @__PURE__ */ (0, J.jsx)(dm, {
			...e ? { id: e } : {},
			errors: t
		})
	});
}
function pm({ errors: e }) {
	return e.length === 0 ? null : /* @__PURE__ */ (0, J.jsxs)("div", {
		role: "alert",
		"aria-label": "Form errors",
		className: "rounded-md border border-destructive/40 bg-destructive/5 px-3 py-2 text-xs text-destructive",
		children: [/* @__PURE__ */ (0, J.jsx)("p", {
			className: "font-medium",
			children: "Review the following errors:"
		}), /* @__PURE__ */ (0, J.jsx)("ul", {
			className: "mt-1 list-disc space-y-0.5 pl-4",
			children: e.map((e, t) => /* @__PURE__ */ (0, J.jsx)("li", { children: e.instancePath ? `${e.instancePath}: ${e.message}` : e.message }, `${e.instancePath}:${e.schemaPath ?? ""}:${e.message}:${t}`))
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-tone.ts
var mm = [
	"neutral",
	"slate",
	"violet",
	"amber",
	"sky",
	"teal",
	"indigo",
	"emerald",
	"rose"
], hm = {
	neutral: "bg-muted text-muted-foreground ring-border",
	slate: "bg-slate-100 text-slate-700 ring-slate-300/60 dark:bg-slate-400/10 dark:text-slate-300 dark:ring-slate-400/25",
	violet: "bg-violet-100 text-violet-700 ring-violet-300/60 dark:bg-violet-400/10 dark:text-violet-300 dark:ring-violet-400/25",
	amber: "bg-amber-100 text-amber-800 ring-amber-300/60 dark:bg-amber-400/10 dark:text-amber-300 dark:ring-amber-400/25",
	sky: "bg-sky-100 text-sky-700 ring-sky-300/60 dark:bg-sky-400/10 dark:text-sky-300 dark:ring-sky-400/25",
	teal: "bg-teal-100 text-teal-700 ring-teal-300/60 dark:bg-teal-400/10 dark:text-teal-300 dark:ring-teal-400/25",
	indigo: "bg-indigo-100 text-indigo-700 ring-indigo-300/60 dark:bg-indigo-400/10 dark:text-indigo-300 dark:ring-indigo-400/25",
	emerald: "bg-emerald-100 text-emerald-700 ring-emerald-300/60 dark:bg-emerald-400/10 dark:text-emerald-300 dark:ring-emerald-400/25",
	rose: "bg-rose-100 text-rose-700 ring-rose-300/60 dark:bg-rose-400/10 dark:text-rose-300 dark:ring-rose-400/25"
}, gm = {
	neutral: "border-l-border dark:border-l-border",
	slate: "border-l-slate-400 dark:border-l-slate-500",
	violet: "border-l-violet-400 dark:border-l-violet-500",
	amber: "border-l-amber-400 dark:border-l-amber-500",
	sky: "border-l-sky-400 dark:border-l-sky-500",
	teal: "border-l-teal-400 dark:border-l-teal-500",
	indigo: "border-l-indigo-400 dark:border-l-indigo-500",
	emerald: "border-l-emerald-400 dark:border-l-emerald-500",
	rose: "border-l-rose-400 dark:border-l-rose-500"
}, _m = {
	neutral: "bg-muted-foreground dark:bg-muted-foreground",
	slate: "bg-slate-400 dark:bg-slate-500",
	violet: "bg-violet-400 dark:bg-violet-500",
	amber: "bg-amber-400 dark:bg-amber-500",
	sky: "bg-sky-400 dark:bg-sky-500",
	teal: "bg-teal-400 dark:bg-teal-500",
	indigo: "bg-indigo-400 dark:bg-indigo-500",
	emerald: "bg-emerald-400 dark:bg-emerald-500",
	rose: "bg-rose-400 dark:bg-rose-500"
};
function vm(e) {
	return typeof e == "string" && mm.includes(e);
}
function ym(e) {
	return vm(e) ? e : "neutral";
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-item-summary.ts
var bm = [
	"title",
	"name",
	"label",
	"id",
	"key"
], xm = "  ·  ", Sm = ", ";
function Cm(e, t) {
	let n = w(e["x-item"]) ? e["x-item"] : {}, r = Rm(n.title), i = zm(n.summary), a = zm(n.badges), o = typeof n.glyph == "string" ? n.glyph : Om(t);
	return {
		title: r.length > 0 ? r : bm,
		...typeof n.fallback == "string" && n.fallback ? { fallback: n.fallback } : {},
		...i.length > 0 ? { summary: i } : {},
		...o ? { glyph: o } : {},
		...typeof n.badge == "string" && n.badge ? { badge: n.badge } : {},
		...a.length > 0 ? { badges: a } : {},
		...n.count === !1 ? { count: !1 } : {},
		...typeof n.flag == "string" && n.flag ? { flag: n.flag } : {},
		...Array.isArray(n.actions) ? { actions: Tm(n.actions) } : {},
		noun: Bm([n.noun, t?.title]) ?? "item",
		nounPlural: Bm([n.nounPlural, e.title]) ?? "items",
		...typeof n.empty == "string" && n.empty ? { empty: n.empty } : {}
	};
}
var wm = [
	"reorder",
	"duplicate",
	"remove"
];
function Tm(e) {
	return e.filter((e) => wm.includes(e));
}
function Em(e, t) {
	return !e.actions || e.actions.includes(t);
}
function Dm(e, t) {
	return e ? x(e, t).properties : {};
}
function Om(e) {
	for (let [t, n] of Object.entries(Dm(e, {}))) {
		let e = n["x-enum-icons"];
		if (Array.isArray(n.enum) && e && Object.keys(e).length > 0) return t;
	}
}
function km({ item: e, index: t, spec: n, itemSchema: r }) {
	let i = w(e) ? e : {}, a = Dm(r, i), o = Bm((n.title ?? bm).map((e) => {
		let t = Lm(i[e]);
		return t && Im(a[e], t);
	})) ?? n.fallback ?? `Item ${t + 1}`, s = (n.summary ?? []).map((e) => Fm(i, e)).filter((e) => !!e).join(xm), c = (n.badges ?? []).map((e) => Fm(i, e)).filter((e) => !!e), l = n.glyph ? Lm(i[n.glyph]) : "", u = n.glyph ? a[n.glyph] : void 0, d = n.badge ? Lm(i[n.badge]) : "", f = n.badge ? a[n.badge] : void 0, p = f?.["x-enum-icons"]?.[d];
	return {
		title: o,
		...s ? { summary: s } : {},
		...n.glyph ? { glyph: {
			...u?.["x-enum-icons"]?.[l] ? { icon: u["x-enum-icons"][l] } : {},
			tone: ym(u?.["x-enum-tones"]?.[l]),
			...l ? { label: Im(u, l) } : {}
		} } : {},
		...d ? { badge: {
			...p ? { icon: p } : {},
			label: Im(f, d)
		} } : {},
		...c.length > 0 ? { badges: c } : {},
		...n.flag && i[n.flag] === !0 ? { flagged: !0 } : {}
	};
}
function Am(e) {
	return `Add ${e.noun ?? "item"}`;
}
function jm(e, t) {
	return t === 1 ? `1 ${e.noun ?? "item"}` : `${t} ${e.nounPlural ?? "items"}`;
}
function Mm(e, t) {
	return t === 0 ? Pm(e) : e.count === !1 ? void 0 : jm(e, t);
}
function Nm(e, t) {
	return e.empty ? e.empty : (typeof t.description == "string" ? t.description.trim() : "") || void 0;
}
function Pm(e) {
	return `No ${e.nounPlural ?? "items"} yet`;
}
function Fm(e, t) {
	let { property: n, pattern: r } = typeof t == "string" ? {
		property: t,
		pattern: void 0
	} : t, i = Lm(e[n]);
	if (i) return r ? r.replace("{}", i) : i;
}
function Im(e, t) {
	let n = e?.["x-enum-labels"]?.[t];
	return typeof n == "string" && n ? n : t;
}
function Lm(e) {
	return e == null ? "" : Array.isArray(e) ? e.map(Lm).filter(Boolean).join(Sm) : typeof e == "object" ? "" : String(e).trim();
}
function Rm(e) {
	return typeof e == "string" ? [e] : Array.isArray(e) ? e.filter((e) => typeof e == "string" && !!e) : [];
}
function zm(e) {
	return Array.isArray(e) ? e.flatMap((e) => typeof e == "string" && e ? [e] : w(e) && typeof e.property == "string" && e.property ? [{
		property: e.property,
		...typeof e.pattern == "string" && e.pattern ? { pattern: e.pattern } : {}
	}] : []) : [];
}
function Bm(e) {
	for (let t of e) {
		let e = typeof t == "string" ? t.trim() : "";
		if (e) return e;
	}
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-union-model.ts
function Vm(e) {
	return e.flatMap((e) => {
		let t = e.oneOf ?? e.anyOf, n = e.type !== void 0 || e.properties !== void 0 || e.const !== void 0 || e.enum !== void 0;
		return t && !n ? Vm(t) : [e];
	});
}
function Hm(e) {
	let t = e.oneOf ?? e.anyOf, n = t && Vm(t);
	if (n && n.length > 1) {
		let e = n.map(qm), t = n.length === 2 && n.some((e) => e.type === "string" && e.pattern), r = new Set(n.map((e) => `${e.format ?? ""}:${e.title ?? ""}`)).size > 1;
		if (!t && (r || new Set(e).size > 1 || e.includes("object") || e.includes("array"))) return n;
	}
	return Array.isArray(e.type) && e.type.length > 1 ? e.type.map((e) => ({ type: e })) : [];
}
function Um(e) {
	return e.type === "string" && e.enum === void 0 && e.const === void 0 && e.pattern === void 0 && e.format === void 0;
}
function Wm(e) {
	if (e.length !== 2) return;
	let t = e.find((e) => e.type === "array"), n = e.find((e) => e !== t);
	if (!t || !n || !Um(n)) return;
	let r = t.items;
	return r === void 0 || Um(r) ? t : void 0;
}
function Gm(e) {
	return typeof e == "string" ? e === "" ? [] : [e] : e ?? [];
}
function Km(e) {
	if (!Array.isArray(e)) return e;
	if (e.length !== 0) return e.length === 1 && typeof e[0] == "string" ? e[0] : e;
}
function qm(e) {
	return e.const === null ? "null" : e.const === void 0 ? e.properties || e.additionalProperties || Object.keys(Ym(e).properties).length ? "object" : e.type?.toString() ?? "value" : typeof e.const;
}
function Jm(e) {
	return e.title ?? e.allOf?.find((e) => e.title)?.title;
}
function Ym(e) {
	let t = ("allOf" in e ? e.allOf ?? [] : []).filter((e) => e.if === void 0).map(Ym);
	return {
		properties: Object.assign({}, ...t.map((e) => e.properties), e.properties),
		closed: e.additionalProperties === !1 || e.unevaluatedProperties === !1 || t.some((e) => e.closed)
	};
}
function Xm(e) {
	return e.default === void 0 ? e.const === void 0 ? e.enum?.length ? e.enum[0] : qm(e) === "null" ? null : qm(e) === "object" ? Object.fromEntries(Object.entries(Ym(e).properties).filter(([, e]) => e.const !== void 0 || e.default !== void 0).map(([e, t]) => [e, Xm(t)])) : e.type === "array" ? [] : e.type === "boolean" ? !1 : e.type === "integer" || e.type === "number" ? 0 : "" : e.const : e.default;
}
function Zm(e, t) {
	if (typeof t == "string" && (typeof e.pattern == "string" && !new RegExp(e.pattern).test(t) || e.format === "date" && (!/^\d{4}-\d{2}-\d{2}$/.test(t) || Number.isNaN(Date.parse(t)) || new Date(t).toISOString().slice(0, 10) !== t) || e.format === "date-time" && (!/^\d{4}-\d{2}-\d{2}T/.test(t) || Number.isNaN(Date.parse(t))))) return !1;
	if (e.const !== void 0) return e.const === t;
	if (e.enum) return e.enum.includes(t);
	if (e.type === "null") return t === null;
	if (e.type === "array") return Array.isArray(t);
	let n = Vm([e]);
	if (n[0] !== e) return n.some((e) => Zm(e, t));
	if (e.type === "object" || qm(e) === "object") {
		if (!w(t)) return !1;
		let { properties: n, closed: r } = Ym(e);
		return r && Object.keys(t).some((e) => !(e in n)) ? !1 : Object.entries(n).every(([e, n]) => n.const === void 0 || t[e] === n.const);
	}
	return e.type === "integer" ? typeof t == "number" && Number.isInteger(t) : e.type === typeof t;
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-nullable-model.ts
function Qm(e) {
	if (Array.isArray(e.type) && e.type.includes("null") && e.type.length > 1) {
		let t = e.type.filter((e) => e !== "null");
		return {
			...e,
			type: t.length === 1 ? t[0] : t,
			...e.enum ? { enum: e.enum.filter((e) => e !== null) } : {}
		};
	}
	let t = e.oneOf ? "oneOf" : "anyOf", n = e[t];
	if (n?.some((e) => e.type === "null" || e.const === null)) {
		let r = n.filter((e) => e.type !== "null" && e.const !== null);
		if (!r.length) return;
		let { [t]: i, ...a } = e;
		return r.length === 1 ? {
			...a,
			...r[0],
			...a.title === void 0 ? {} : { title: a.title }
		} : {
			...a,
			[t]: r
		};
	}
	if (e.enum?.includes(null) && e.enum.length > 1 && e.type === void 0) return {
		...e,
		enum: e.enum.filter((e) => e !== null)
	};
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-resolve.ts
function $m(e) {
	return nh(e) ? typeof e.additionalProperties == "object" && e.additionalProperties !== null || !!e.patternProperties && Object.keys(e.patternProperties).length > 0 : !1;
}
function eh(e) {
	let t = e.patternProperties;
	if (t && Object.keys(t).length !== 0) return Object.entries(t).map(([e, t]) => ({
		pattern: e,
		schema: t
	}));
}
function th(e, t) {
	return Array.isArray(e.type) ? e.type.includes(t) : e.type === t;
}
function nh(e) {
	return th(e, "object") ? !0 : e.type === void 0 ? e.properties !== void 0 || e.required !== void 0 || e.additionalProperties !== void 0 || e.patternProperties !== void 0 || e.propertyNames !== void 0 || e.unevaluatedProperties !== void 0 ? !0 : e.allOf?.some((e) => e.if === void 0 && e.then === void 0 && nh(e)) ?? !1 : !1;
}
function rh(e) {
	let t = e["x-enum-labels"], n = e["x-enum-icons"], r = e["x-enum-descriptions"], i = e["x-enum-tones"];
	return (e.enum ?? []).map((e) => {
		let a = String(e), o = t?.[a], s = n?.[a], c = r?.[a], l = i?.[a], u = typeof o == "string" && o && o !== a;
		return {
			value: a,
			label: u ? `${o} (${a})` : a,
			...u ? { name: o } : {},
			...typeof s == "string" && s ? { icon: s } : {},
			...typeof c == "string" && c ? { description: c } : {},
			...vm(l) ? { tone: l } : {}
		};
	});
}
function ih(e) {
	let t = e["x-enum-display"];
	if (t === "combobox" || t === "radio" || t === "grid" || t === "segmented") return t;
	let n = e["x-enum-icons"];
	if (n && Object.keys(n).length > 0) return "grid";
}
function ah(e) {
	let t = e["x-array-display"];
	return t === "filter-pills" || t === "accordion" || t === "cards" || t === "stacked" || t === "list" ? t : void 0;
}
function oh(e) {
	let t = e["x-help-display"];
	return t === "inline" || t === "hover" ? t : void 0;
}
function sh(e) {
	let t = typeof e.description == "string" ? e.description.trim() : "", n = e["x-help"], r = w(n) && typeof n.body == "string" ? n.body.trim() : "";
	return t && r && r !== t ? `${t} ${r}` : t || r || void 0;
}
function ch(e, t) {
	if (typeof e == "string" && e) return (0, Y.createElement)(o, {
		icon: e,
		className: "size-4 text-muted-foreground/70"
	});
	if (typeof t == "string" && t) return (0, Y.createElement)("span", { className: "text-sm text-muted-foreground" }, t);
}
function lh(e) {
	let t = [...e.anyOf ?? [], ...e.oneOf ?? []], n = t.every((e) => e.enum?.length || typeof e.const == "string"), r = t.filter((e) => e.enum?.length || n && typeof e.const == "string"), i = r[0];
	if (i) return {
		...i,
		enum: [...new Set(r.flatMap((e) => e.enum ?? [e.const]))],
		"x-enum-labels": Object.assign({}, ...r.map((e) => e["x-enum-labels"])),
		"x-enum-icons": Object.assign({}, ...r.map((e) => e["x-enum-icons"])),
		"x-enum-descriptions": Object.assign({}, ...r.map((e) => e["x-enum-descriptions"]))
	};
}
function uh(e) {
	let t = e["x-clicky-lookup"];
	if (w(t) && typeof t.url == "string" && typeof t.filter == "string") return t;
}
function dh(e) {
	let t = Qm(e.prop);
	if (t) return {
		...dh({
			...e,
			prop: t
		}),
		nullable: !0
	};
	let { key: n, prop: r, required: i, value: a, onChange: o } = e, s = r["x-icon"], c = r["x-layout"], l = typeof c == "string" && c !== "" ? c : void 0, u = r["x-label-position"], d = l ?? (u === "top" ? "stack" : u === "left" ? "inline" : void 0), f = ch(r["x-input-prefix-icon"], r["x-input-prefix"]), p = ch(r["x-input-suffix-icon"], r["x-input-suffix"]), m = typeof r["x-label-classes"] == "string" ? r["x-label-classes"] : void 0, h = typeof r["x-input-classes"] == "string" ? r["x-input-classes"] : void 0, g = r["x-col-span"] === "full" ? "full" : typeof r["x-col-span"] == "number" && Number.isFinite(r["x-col-span"]) ? r["x-col-span"] : void 0, _ = fh(r), v = sh(r), y = oh(r), b = {
		key: n,
		kind: "string",
		label: typeof r.title == "string" ? r.title : n,
		required: i,
		schema: r,
		value: a,
		onChange: o,
		...r.readOnly === !0 ? { readOnly: !0 } : {},
		...r.writeOnly === !0 ? { writeOnly: !0 } : {},
		...r["x-disabled"] === !0 ? { disabled: !0 } : {},
		...typeof r.description == "string" ? { description: r.description } : {},
		...v ? { helper: v } : {},
		...y ? { helpDisplay: y } : {},
		...s != null && s !== "" ? { labelIcon: s } : {},
		...d ? { layout: d } : {},
		...f ? { prefix: f } : {},
		...p ? { suffix: p } : {},
		...m ? { labelClassName: m } : {},
		...h ? { inputClassName: h } : {},
		...g == null ? {} : { colSpan: g }
	};
	if (typeof r.format == "string" && e.extensions?.formats.has(r.format)) return {
		...b,
		kind: "custom",
		format: r.format
	};
	let S = uh(r);
	if (S) return {
		...b,
		kind: "lookup",
		lookup: S,
		options: [],
		allowCustomValue: S.multi !== !0
	};
	if (r.format === "json") return {
		...b,
		kind: "json"
	};
	if (Hm(r).length > 1) return {
		...b,
		kind: "union"
	};
	if (Array.isArray(r.enum) && r.enum.length > 0) {
		let e = ih(r);
		return {
			...b,
			kind: "enum",
			options: rh(r),
			...e ? { display: e } : {}
		};
	}
	let C = lh(r);
	if (C) return {
		...b,
		kind: "enum",
		options: rh(C)
	};
	let T = r.anyOf?.find((e) => e.type === "string" && typeof e.pattern == "string");
	if (r.anyOf?.length === 2 && T) {
		let t = r.anyOf.find((e) => e !== T);
		if (t && (t.type === "number" || t.type === "integer" || t.type === "boolean" || t.format === "date" || t.format === "date-time")) {
			let n = {
				...r,
				...t
			};
			return delete n.anyOf, dh({
				...e,
				prop: n
			});
		}
	}
	if (r.format === "date" || r.format === "date-time") return {
		...b,
		kind: "date",
		dateFormat: r.format
	};
	if (r.format === "textarea" && th(r, "string")) return {
		...b,
		kind: "textarea"
	};
	if (r.format === "md" && th(r, "string")) {
		let e = r["x-md-editor"];
		return {
			...b,
			kind: "markdown",
			...w(e) ? { markdownOptions: e } : {}
		};
	}
	if (th(r, "boolean")) return {
		...b,
		kind: "boolean"
	};
	if (th(r, "integer") || th(r, "number")) return {
		...b,
		kind: "number",
		coerceNumber: !0,
		...r.format === "percent" ? { unit: "%" } : {},
		...typeof r.minimum == "number" ? { minimum: r.minimum } : {}
	};
	if (th(r, "array")) {
		let e = ah(r), t = r.items;
		return {
			...b,
			kind: "array",
			...t ? { itemSchema: t } : {},
			...typeof r.minItems == "number" ? { minItems: r.minItems } : {},
			...typeof r.maxItems == "number" ? { maxItems: r.maxItems } : {},
			...e ? { arrayDisplay: e } : {},
			...e === "accordion" ? { itemSpec: Cm(r, t) } : {},
			...e === "accordion" && !y ? { helpDisplay: "hover" } : {},
			...t && Array.isArray(t.enum) ? { options: rh(t) } : {},
			...d ? { layout: d } : {}
		};
	}
	if ($m(r)) {
		let e = typeof r.additionalProperties == "object" && r.additionalProperties !== null ? r.additionalProperties : void 0, t = eh(r), n = _ || t ? !0 : r.additionalProperties !== !1;
		return {
			...b,
			kind: "string-map",
			...e ? { valueSchema: e } : {},
			...t ? { valuePatternSchemas: t } : {},
			...r.properties ? { knownProperties: r.properties } : {},
			allowExtraKeys: n,
			..._ ? { keyOptions: _ } : {},
			...d ? { layout: d } : {}
		};
	}
	if (nh(r)) {
		let e = w(a) ? a : {}, t = x(r, e);
		if (Object.keys(t.properties).length > 0 || r.properties) return {
			...b,
			kind: "object",
			objectProperties: t.properties,
			...t.required.length > 0 ? { objectRequired: t.required } : {},
			...d ? { layout: d } : {}
		};
	}
	return nh(r) ? {
		...b,
		kind: "string-map",
		allowExtraKeys: r.additionalProperties !== !1,
		..._ ? { keyOptions: _ } : {},
		...d ? { layout: d } : {}
	} : b;
}
function fh(e) {
	let t = e.propertyNames;
	if (t && Array.isArray(t.enum) && t.enum.length !== 0) return rh(t);
}
function ph(e) {
	if (!e) return "string";
	if (e.enum === void 0 && e.const === void 0 && e.properties === void 0 && e.items === void 0 && e.additionalProperties === void 0 && e.allOf === void 0 && e.oneOf === void 0 && e.anyOf === void 0) {
		if (e.type === void 0 || th(e, "string")) return "string";
		if (th(e, "integer")) return "integer";
		if (th(e, "number")) return "number";
	}
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-defaults.ts
function mh(e, t) {
	return hh(e, t, /* @__PURE__ */ new Set());
}
function hh(e, t, n) {
	if (n.has(e)) return t;
	let r = new Set(n).add(e), i = t, a = Math.min(100, Object.keys(e.properties ?? {}).length + (e.allOf?.length ?? 0) + 2);
	for (let t = 0; t < a; t += 1) {
		let { properties: t } = x(e, i), n = !1, a = i;
		for (let [e, o] of Object.entries(t)) {
			let t = a[e], s = t, c = !1;
			if (t === void 0 && o.default !== void 0 && (s = gh(o.default), c = !0), nh(o) && !$m(o)) {
				let e = w(s) ? s : {}, t = hh(o, e, r);
				t !== e && (s = t, c = !0);
			}
			c && (n || (a = { ...i }), a[e] = s, n = !0);
		}
		if (!n) break;
		i = a;
	}
	return i;
}
function gh(e) {
	return Array.isArray(e) ? e.map(gh) : w(e) ? Object.fromEntries(Object.entries(e).map(([e, t]) => [e, gh(t)])) : e;
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-listener-paths.ts
var _h = "/";
function vh(e) {
	return e.includes(_h);
}
function yh(e, t) {
	let n = t.split(_h);
	if (n.some((e) => e === "")) throw Error(`${e}: target "${t}" has an empty path segment`);
	return n;
}
function bh({ properties: e, value: t, target: n, apply: r }) {
	let [i = "", ...a] = n.split(_h), o = e[i];
	if (!o) return;
	let s = xh(o, t[i], a, r);
	s && (e[i] = s);
}
function xh(e, t, n, r) {
	let [i, ...a] = n;
	if (i === void 0) return r(e);
	if (Sh(e)) {
		if (!e.items) return;
		let t = xh(e.items, {}, n, r);
		return t && {
			...e,
			items: t
		};
	}
	let o = w(t) ? t : {}, s = x(e, o).properties[i];
	if (!s) return;
	let c = xh(s, o[i], a, r);
	return c && {
		...e,
		allOf: [...e.allOf ?? [], { properties: { [i]: c } }]
	};
}
function Sh(e) {
	return e.items === void 0 ? Array.isArray(e.type) ? e.type.includes("array") : e.type === "array" : !0;
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-listeners.ts
var Ch = [
	"require",
	"optional",
	"reset",
	"set"
], wh = [
	[
		"hide",
		"x-hidden",
		!0
	],
	[
		"show",
		"x-hidden",
		!1
	],
	[
		"disable",
		"x-disabled",
		!0
	],
	[
		"enable",
		"x-disabled",
		!1
	]
], Th = [
	["hide", "show"],
	["enable", "disable"],
	["require", "optional"]
], Eh = [
	"hide",
	"show",
	"enable",
	"disable",
	"require",
	"optional",
	"reset"
], Dh = {
	...Object.fromEntries([
		"type",
		"properties",
		"items",
		"additionalProperties",
		"patternProperties",
		"allOf",
		"anyOf",
		"oneOf",
		"if",
		"then",
		"else",
		"not",
		"$ref",
		"$defs"
	].map((e) => [e, "it changes the field's shape; use an allOf branch"])),
	required: "use require/optional",
	default: "a patch changes presentation, not values; use set/reset",
	const: "a patch changes presentation, not values; use set/reset",
	"x-on-change": "listeners cannot be patched in",
	"x-on-load": "listeners cannot be patched in"
};
function Oh(e, t, n) {
	let { properties: r, required: i } = x(e, t), a = new Set(i), o = Hh(e), s = [...Ah(e, t), ...Object.entries(r).flatMap(([e, t]) => Fh(e, t).map((t) => ({
		listener: t,
		source: Mh(e)
	})))];
	for (let { listener: e, source: i } of s) {
		let s = Ih({
			listener: e,
			source: i,
			self: t,
			known: o,
			options: n
		});
		for (let [e, n, a] of wh) for (let o of Vh(s, e)) bh({
			properties: r,
			value: t,
			target: o,
			apply: (e) => kh({
				...e,
				[n]: a
			}, n, a, i.label)
		});
		for (let e of Vh(s, "require")) r[e] && a.add(e);
		for (let e of Vh(s, "optional")) a.delete(e);
		for (let [e, n] of Object.entries(s.patch ?? {})) bh({
			properties: r,
			value: t,
			target: e,
			apply: (e) => ({
				...e,
				...n
			})
		});
	}
	return {
		properties: r,
		required: [...a]
	};
}
function kh(e, t, n, r) {
	if (t !== "x-hidden") return e;
	if (n) return {
		...e,
		"x-hidden-by": r
	};
	let { "x-hidden-by": i, ...a } = e;
	return a;
}
function Ah(e, t) {
	let n = jh(e["x-on-load"], "#/x-on-load");
	for (let { branch: r, pointer: i } of _(e, t)) n.push(...jh(r["x-on-load"], `#${i}/x-on-load`));
	for (let [r, i] of (e.allOf ?? []).entries()) {
		if (i.if === void 0 && i.then === void 0) {
			n.push(...jh(i["x-on-load"], `#/allOf/${r}/x-on-load`));
			continue;
		}
		let e = h(i.if, t) ? "then" : "else";
		n.push(...jh(i[e]?.["x-on-load"], `#/allOf/${r}/${e}/x-on-load`));
	}
	return n;
}
function jh(e, t) {
	if (e === void 0) return [];
	if (!Array.isArray(e)) throw Error(`x-on-load at ${t} must be an array of listeners, got ${typeof e}`);
	return e.map((e, n) => ({
		listener: e,
		source: {
			label: `x-on-load at ${t}/${n}`,
			key: "",
			load: !0
		}
	}));
}
function Mh(e) {
	return {
		label: `x-on-change on "${e}"`,
		key: e,
		load: !1
	};
}
function Nh(e, t) {
	let n = {
		...t.value,
		[t.key]: t.next
	}, r = Hh(e), i = /* @__PURE__ */ new Set(), a = [t.key];
	for (let o = a.shift(); o !== void 0; o = a.shift()) {
		if (i.has(o)) continue;
		i.add(o);
		let { properties: s } = x(e, n), c = {
			root: t.root === t.value ? n : t.root,
			...t.evaluate ? { evaluate: t.evaluate } : {}
		};
		for (let e of Fh(o, s[o])) {
			let t = Ih({
				listener: e,
				source: Mh(o),
				self: n,
				known: r,
				options: c
			});
			for (let [e, r] of Ph(t, s)) Object.is(n[e], r) || (n = { ...n }, r === void 0 ? delete n[e] : n[e] = r, a.push(e));
		}
	}
	return n;
}
function Ph(e, t) {
	return [...Vh(e, "reset").map((e) => {
		let n = t[e]?.default;
		return [e, n === void 0 ? void 0 : gh(n)];
	}), ...Object.entries(e.set ?? {}).map(([e, t]) => [e, gh(t)])];
}
function Fh(e, t) {
	let n = t?.["x-on-change"];
	if (n === void 0) return [];
	if (!Array.isArray(n)) throw Error(`"${e}": x-on-change must be an array of listeners, got ${typeof n}`);
	return n;
}
function Ih({ listener: e, source: t, self: n, known: r, options: i }) {
	return Rh(t, e, r), e.else && Rh(t, e.else, r), Lh(e, t, n, i) ? e : e.else ?? {};
}
function Lh(e, { label: t, key: n, load: r }, i, a) {
	if (e.when === void 0) return !0;
	let { expr: o, ...s } = e.when, c = Object.keys(s).length > 0, l = r ? i : i[n];
	if (!c && o === void 0) throw Error(`${t}: when states nothing it can evaluate`);
	if (o !== void 0 && typeof o != "string") throw Error(`${t}: when.expr must be a string, got ${typeof o}`);
	if (o !== void 0 && !a.evaluate) throw Error(`${t}: when.expr "${o}" needs the form's expressionEvaluator`);
	if (c) {
		let e = C(s, l);
		if (e === void 0) throw Error(`${t}: when states nothing it can evaluate (${JSON.stringify(s)})`);
		if (!e) return !1;
	}
	if (o === void 0 || !a.evaluate) return !0;
	let u = a.evaluate({
		expr: o,
		key: n,
		value: l,
		self: i,
		root: a.root
	});
	if (typeof u != "boolean") throw Error(`${t}: expressionEvaluator returned ${typeof u} for "${o}", expected boolean`);
	return u;
}
function Rh(e, t, n) {
	let { label: r } = e;
	if (t.set !== void 0 && !w(t.set)) throw Error(`${r}: set must be an object of target → value`);
	if (e.load) {
		let e = Ch.find((e) => t[e] !== void 0);
		if (e) throw Error(`${r}: ${e} is not allowed in x-on-load, which only derives state (hide/show/enable/disable/patch)`);
	}
	let i = Object.keys(t.set ?? {});
	for (let e of Ch) {
		let n = (e === "set" ? i : Vh(t, e)).find(vh);
		if (n !== void 0) throw Error(`${r}: ${e} "${n}" cannot take a path; value and required actions act on sibling keys only`);
	}
	for (let e of [...Eh.flatMap((e) => Vh(t, e)), ...i]) zh(r, e, n);
	let a = [...Th.map(([e, n]) => [
		e,
		Vh(t, e),
		n,
		Vh(t, n)
	]), [
		"reset",
		Vh(t, "reset"),
		"set",
		i
	]];
	for (let [e, t, n, i] of a) {
		let a = t.find((e) => i.includes(e));
		if (a !== void 0) throw Error(`${r}: both ${e} and ${n} "${a}"`);
	}
	Bh(r, t, n);
}
function zh(e, t, n) {
	if (vh(t) && n.has(t)) throw Error(`${e}: target "${t}" is both a declared key and a path`);
	let [r = ""] = yh(e, t);
	if (!n.has(r)) throw Error(`${e}: unknown target "${t}"`);
}
function Bh(e, t, n) {
	let r = t.patch;
	if (r !== void 0) {
		if (!w(r)) throw Error(`${e}: patch must be an object of target → keywords`);
		for (let [i, a] of Object.entries(r)) {
			if (zh(e, i, n), !w(a)) throw Error(`${e}: patch "${i}" must be an object of keywords`);
			for (let t of Object.keys(a)) {
				let n = Dh[t];
				if (n) throw Error(`${e}: patch "${i}" may not set ${t} (${n})`);
			}
			for (let [n, r] of wh) if (r in a && Vh(t, n).includes(i)) throw Error(`${e}: both ${n} and patch ${r} on "${i}"`);
		}
	}
}
function Vh(e, t) {
	let n = e[t];
	if (n === void 0) return [];
	if (!Array.isArray(n)) throw Error(`x-on-change: ${t} must be an array of property keys, got ${typeof n}`);
	return n;
}
function Hh(e) {
	let t = new Set(Object.keys(e.properties ?? {}));
	for (let n of b(e)) for (let e of Object.keys(n.properties ?? {})) t.add(e);
	for (let n of e.allOf ?? []) for (let e of [n, ...b(n)]) for (let n of Object.keys(e?.properties ?? {})) t.add(n);
	return t;
}
function Uh(e, t) {
	return {
		...Wh(e, t),
		layoutMode: t.layout.mode
	};
}
function Wh(e, t) {
	return {
		hideReadOnlyFields: t.hideReadOnlyFields,
		valueCells: t.viewOnly || e.readOnly === !0,
		...t.rootValue ? { rootValue: t.rootValue } : {},
		...t.expressionEvaluator ? { evaluate: t.expressionEvaluator } : {}
	};
}
function Gh(e, t, n) {
	let r = e, i = (t.length > 0 ? t : [{}]).map((e) => {
		let t = w(e) ? e : {};
		return Oh(r, t, {
			root: n.rootValue ?? t,
			...n.evaluate ? { evaluate: n.evaluate } : {}
		}).properties;
	}), a = [...new Set(i.flatMap((e) => Object.keys(e)))].flatMap((e) => {
		let t = i.map((t) => t[e]).find((e) => e !== void 0 && Kh(e, n));
		return t ? [[e, t]] : [];
	});
	return {
		columns: K(E(a), r["x-order"]),
		rows: t.length > 0 ? i : []
	};
}
function Kh(e, t) {
	return e["x-hidden"] !== !0 && !(t.hideReadOnlyFields && e.readOnly === !0) && (e.writeOnly !== !0 || !t.valueCells && e.readOnly !== !0);
}
function qh(e, t) {
	if (e.arrayDisplay === "accordion") return t.layoutMode === "inline" ? "inline" : "stack";
	if (e.layout === "table") return "grid";
	let n = e.schema["x-table-max-columns"] ?? 4;
	if (!Number.isInteger(n) || n < 1) throw Error(`${e.key}: x-table-max-columns must be a positive integer, got ${JSON.stringify(n)}`);
	let r = Array.isArray(e.value) ? e.value : [], { columns: i } = Gh(e.itemSchema ?? { type: "object" }, r, t);
	return i.some(([e, t]) => Jh(e, t)) ? "inline" : i.length <= n ? "grid" : "inline";
}
function Jh(e, t) {
	let n = dh({
		key: e,
		prop: t,
		required: !1,
		value: void 0,
		onChange: () => {}
	});
	return n.kind === "object" || n.kind === "string-map" || n.kind === "array" && H(n.itemSchema);
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-filter.ts
function Yh({ key: e, prop: t, filter: n, value: r }) {
	let i = n.trim().toLowerCase();
	if (!i || e.toLowerCase().includes(i) || t.title?.toLowerCase().includes(i)) return !0;
	if (t.format === "password") return !1;
	if (Array.isArray(r)) return r.some((e) => Yh({
		key: "",
		prop: t.items ?? {},
		filter: n,
		value: e
	}));
	if (w(r)) {
		let { properties: e } = x(t, r), i = $m(t) || nh(t) && !t.properties && Object.keys(e).length === 0;
		return [.../* @__PURE__ */ new Set([...Object.keys(e), ...Object.keys(r)])].some((a) => {
			let o = e[a] ?? Object.entries(t.patternProperties ?? {}).find(([e]) => new RegExp(e).test(a))?.[1] ?? (w(t.additionalProperties) ? t.additionalProperties : i ? { type: "string" } : void 0);
			return o !== void 0 && Yh({
				key: a,
				prop: o,
				filter: n,
				value: r[a]
			});
		});
	}
	if (r == null) return !1;
	let a = String(r).toLowerCase(), o = t["x-enum-labels"]?.[String(r)];
	return a.includes(i) || !!o?.toLowerCase().includes(i);
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-errors.ts
function Xh(e, t) {
	return `${e}/${String(t).replaceAll("~", "~0").replaceAll("/", "~1")}`;
}
function Zh(e, t) {
	return e.filter((e) => e.instancePath === t || e.instancePath.startsWith(`${t}/`)).length;
}
function Qh(e, t) {
	return e.filter((e) => e.instancePath === t);
}
function $h({ schema: e, value: t, errors: n, hiddenKeys: r, viewOnly: i, hideReadOnlyFields: a, hideEmpty: o, fieldFilter: s, pre: c, extensions: l, rootValue: u = t, instancePath: d = "", expressionEvaluator: f }) {
	let p = /* @__PURE__ */ new Set();
	return eg(e, t, d, p, {
		hiddenKeys: new Set(r ?? []),
		viewOnly: i,
		hideReadOnlyFields: a,
		hideEmpty: o,
		pre: c,
		...l ? { extensions: l } : {},
		rootValue: u,
		root: !0,
		...s ? { fieldFilter: s } : {},
		...f ? { expressionEvaluator: f } : {}
	}), n.filter((e) => !p.has(e.instancePath));
}
function eg(e, t, n, r, i) {
	let { properties: a, required: o } = Oh(e, t, {
		root: i.rootValue,
		...i.expressionEvaluator ? { evaluate: i.expressionEvaluator } : {}
	}), s = i.root && typeof e["x-discriminator"] == "string" ? e["x-discriminator"] : void 0, c = s != null && (t[s] == null || t[s] === "");
	for (let [e, l] of Object.entries(a)) {
		if (l["x-hidden"] || i.root && i.hiddenKeys.has(e) || i.root && i.fieldFilter && !Yh({
			key: e,
			prop: l,
			filter: i.fieldFilter,
			value: t[e]
		}) || s && (c ? e !== s : e === s)) continue;
		let a = rg(dh({
			key: e,
			prop: l,
			required: o.includes(e),
			value: t[e],
			onChange: () => {},
			...i.extensions ? { extensions: i.extensions } : {}
		}), l, t[e], i.rootValue, i.pre);
		if (!a || i.hideReadOnlyFields && a.readOnly || tg(a, i) || i.hideEmpty && P(a.value)) continue;
		let u = Xh(n, e);
		r.add(u), ig(a, u, r, {
			...i,
			root: !1,
			viewOnly: ng(a, i)
		});
	}
}
function tg(e, t) {
	return e.writeOnly === !0 && (t.viewOnly || e.readOnly === !0);
}
function ng(e, t) {
	return t.viewOnly || e.readOnly === !0;
}
function rg(e, t, n, r, i) {
	let a = e;
	for (let e of i) {
		if (!a) return null;
		a = e(a, {
			key: a.key,
			prop: t,
			value: n,
			rootValue: r
		});
	}
	return a;
}
function ig(e, t, n, r) {
	e.kind === "object" ? eg(e.schema, w(e.value) ? e.value : {}, t, n, r) : e.kind === "array" ? ag(e, t, n, r) : e.kind === "string-map" && sg(e, t, n, r);
}
function ag(e, t, n, r) {
	if (e.arrayDisplay === "filter-pills" || e.arrayDisplay !== "stacked" && (og(e) || ph(e.itemSchema))) return;
	let i = e.itemSchema ?? { type: "string" }, a = e.arrayDisplay !== "stacked" && H(i) && qh(e, {
		hideReadOnlyFields: r.hideReadOnlyFields,
		valueCells: ng(e, r),
		layoutMode: "stacked",
		rootValue: r.rootValue,
		...r.expressionEvaluator ? { evaluate: r.expressionEvaluator } : {}
	}) === "grid";
	for (let [o, s] of (Array.isArray(e.value) ? e.value : []).entries()) {
		let c = Xh(t, o);
		if (a) {
			eg(i, w(s) ? s : {}, c, n, r);
			continue;
		}
		let l = dh({
			key: `${e.key}[${o}]`,
			prop: i,
			required: !1,
			value: s,
			onChange: () => {},
			...r.extensions ? { extensions: r.extensions } : {}
		});
		tg(l, r) || (n.add(c), ig(l, c, n, {
			...r,
			viewOnly: ng(l, r)
		}));
	}
}
function og(e) {
	return e.options && e.options.length > 0 ? !0 : Array.isArray(e.itemSchema?.enum) && e.itemSchema.enum.length > 0;
}
function sg(e, t, n, r) {
	let i = w(e.value) ? e.value : {}, a = e.knownProperties ?? {}, o = [...Object.keys(a), ...Object.keys(i).filter((e) => !(e in a))];
	for (let s of o) {
		let o = a[s] ?? cg(e, s), c = Xh(t, s), l = dh({
			key: s,
			prop: o,
			required: !1,
			value: i[s],
			onChange: () => {},
			...r.extensions ? { extensions: r.extensions } : {}
		});
		tg(l, r) || (n.add(c), ig(l, c, n, {
			...r,
			viewOnly: ng(l, r)
		}));
	}
}
function cg(e, t) {
	for (let n of e.valuePatternSchemas ?? []) try {
		if (new RegExp(n.pattern).test(t)) return n.schema;
	} catch {
		continue;
	}
	return e.valueSchema ?? { type: "string" };
}
// ../../../clicky-ui/packages/ui/src/components/form-lookup-context.ts
var lg = (0, Y.createContext)({ handlers: /* @__PURE__ */ new Map() });
function ug(e) {
	let { fetcher: t, handlers: n } = (0, Y.useContext)(lg);
	return e && n.get(e.url) || t;
}
function dg(e, t) {
	if (!e || !t) return "";
	let n = e;
	for (let e of t.split(".")) {
		if (typeof n != "object" || !n) return "";
		n = n[e];
	}
	return typeof n == "string" || typeof n == "number" ? String(n) : "";
}
function fg(e, t) {
	let n = e.scope;
	if (!n) return {};
	let r = dg(t, n.from);
	if (!r) return {};
	let i = n.map ? n.map[r] ?? [] : [r];
	return i.length === 0 ? {} : { [n.param]: i.join(n.join ?? ",") };
}
// ../../../clicky-ui/packages/ui/src/components/FormLookupProvider.tsx
var pg = /* @__PURE__ */ new Map();
function mg({ fetcher: e, handlers: t = pg, children: n }) {
	let r = (0, Y.useMemo)(() => ({
		...e ? { fetcher: e } : {},
		handlers: t
	}), [e, t]);
	return /* @__PURE__ */ (0, J.jsx)(lg.Provider, {
		value: r,
		children: n
	});
}
// ../../../clicky-ui/packages/ui/src/components/form-connection-context.ts
var hg = (0, Y.createContext)(void 0);
function gg() {
	return (0, Y.useContext)(hg);
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-discriminator.tsx
function _g({ schema: e, value: t, onChange: n, ctx: r, discKey: i, hiddenKeys: a = [] }) {
	let s = e.properties?.[i];
	if (!s) return /* @__PURE__ */ (0, J.jsx)(J.Fragment, { children: r.render.renderObjectFields(e, t, n, r, { hiddenKeys: a }) });
	let c = t[i], l = (e.required ?? []).includes(i);
	if (c == null || c === "") {
		let e = r.render.renderFieldNodes({
			key: i,
			prop: s,
			required: l,
			value: c,
			onChange: (e) => n({
				...t,
				[i]: e
			}),
			instancePath: Xh(r.instancePath, i)
		}, r);
		return /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "col-span-full flex flex-col gap-2",
			children: [/* @__PURE__ */ (0, J.jsx)("h3", {
				className: q("font-semibold", O[r.size]),
				children: typeof s.title == "string" && s.title ? s.title : i
			}), e?.value]
		});
	}
	let u = r.render.renderObjectFields(e, t, n, r, { hiddenKeys: [...a, i] });
	if (a.includes(i)) return /* @__PURE__ */ (0, J.jsx)(J.Fragment, { children: u });
	let d = String(c);
	return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsxs)("div", {
		className: "col-span-full flex items-center justify-between rounded-md border border-input px-3 py-2",
		children: [/* @__PURE__ */ (0, J.jsxs)("span", {
			className: q("flex items-center gap-2 font-medium", O[r.size]),
			children: [/* @__PURE__ */ (0, J.jsx)(o, {
				icon: yg(s, d),
				className: "text-lg text-foreground"
			}), /* @__PURE__ */ (0, J.jsx)("span", { children: vg(s, d) })]
		}), !r.readOnly && /* @__PURE__ */ (0, J.jsxs)("button", {
			type: "button",
			onClick: () => {
				let e = { ...t };
				delete e[i], n(e);
			},
			className: "text-xs font-medium text-primary hover:underline",
			children: ["Change ", typeof s.title == "string" && s.title ? s.title.toLowerCase() : i]
		})]
	}), u] });
}
function vg(e, t) {
	let n = e["x-enum-labels"]?.[t];
	return typeof n == "string" && n ? n : t;
}
function yg(e, t) {
	let n = e["x-enum-icons"]?.[t];
	return typeof n == "string" && n ? n : void 0;
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-refs.ts
function bg(e) {
	let t = e, n = /* @__PURE__ */ new Map(), r = /* @__PURE__ */ new Set(), i = !1, a = (e) => {
		if (n.has(e)) return n.get(e);
		if (r.has(e)) return {
			format: "json",
			description: "↻ recursive schema"
		};
		let a = wg(t, e);
		if (!Dg(a)) {
			i = !0;
			return;
		}
		r.add(e);
		let s = o(a);
		return r.delete(e), n.set(e, s), s;
	}, o = (e) => {
		if (Array.isArray(e)) {
			let t = !1, n = e.map((e) => {
				let n = o(e);
				return n !== e && (t = !0), n;
			});
			return t ? n : e;
		}
		if (e && typeof e == "object") {
			let t = e, n = t.$ref;
			if (typeof n == "string" && Cg(n)) {
				let e = a(n);
				if (!e) return s(t);
				let r = {}, i = !1;
				for (let [e, n] of Object.entries(t)) e !== "$ref" && (r[e] = o(n), i = !0);
				return i ? xg(e, r) : e;
			}
			let r = !1, i = {};
			for (let [e, n] of Object.entries(t)) {
				let t = o(n);
				i[e] = t, t !== n && (r = !0);
			}
			return r ? i : e;
		}
		return e;
	}, s = (e) => {
		let t = !1, n = {};
		for (let [r, i] of Object.entries(e)) {
			if (r === "$ref") {
				n[r] = i;
				continue;
			}
			let e = o(i);
			n[r] = e, e !== i && (t = !0);
		}
		return t ? n : e;
	};
	if (!Object.hasOwn(t, "$defs")) return o(e);
	let { $defs: c, ...l } = t, u = o(l);
	return i ? {
		...u,
		$defs: o(c)
	} : u;
}
function xg(e, t) {
	let n = {
		...e,
		...t
	};
	Eg(e.properties) && Eg(t.properties) && (n.properties = {
		...e.properties,
		...t.properties
	}), Array.isArray(e.required) && Array.isArray(t.required) && (n.required = [.../* @__PURE__ */ new Set([...e.required, ...t.required])]);
	let r = [...Array.isArray(e.allOf) ? e.allOf : [], ...Array.isArray(t.allOf) ? t.allOf : []], i = e.if !== void 0 && t.if !== void 0;
	return i && (delete n.if, delete n.then, delete n.else, r.push(Sg(e), Sg(t))), (i || Array.isArray(e.allOf) && Array.isArray(t.allOf)) && (n.allOf = r), n;
}
function Sg(e) {
	return Object.fromEntries([
		"if",
		"then",
		"else"
	].filter((t) => e[t] !== void 0).map((t) => [t, e[t]]));
}
function Cg(e) {
	if (!e.startsWith("#")) return !1;
	if (e === "#") return !0;
	try {
		return decodeURIComponent(e.slice(1)).startsWith("/");
	} catch {
		return e.startsWith("#/");
	}
}
function wg(e, t) {
	let n;
	try {
		n = decodeURIComponent(t.slice(1));
	} catch {
		return;
	}
	if (n === "") return e;
	if (!n.startsWith("/")) return;
	let r = e;
	for (let e of n.slice(1).split("/")) {
		let t = Tg(e);
		if (t === void 0) return;
		if (Array.isArray(r)) {
			if (!/^(0|[1-9]\d*)$/.test(t)) return;
			let e = Number(t);
			if (!Number.isSafeInteger(e) || e >= r.length) return;
			r = r[e];
		} else {
			if (!Eg(r) || !Object.hasOwn(r, t)) return;
			r = r[t];
		}
	}
	return r;
}
function Tg(e) {
	if (!/~(?:[^01]|$)/.test(e)) return e.replaceAll("~1", "/").replaceAll("~0", "~");
}
function Eg(e) {
	return typeof e == "object" && !!e && !Array.isArray(e);
}
function Dg(e) {
	return Eg(e);
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-debug-glyph.tsx
function Og({ glyph: e, className: t }) {
	let n = a();
	if (!e) return /* @__PURE__ */ (0, J.jsx)("span", {
		"aria-hidden": !0,
		className: q("inline-block size-3.5 shrink-0", t)
	});
	let r = n === "dark" && e.dark ? e.dark : e.light;
	return /* @__PURE__ */ (0, J.jsx)(r, {
		"data-glyph": r.displayName,
		className: q("size-3.5 shrink-0", t)
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-debug-glyph-map.ts
var Q = (e, t) => t ? {
	light: e,
	dark: t
} : { light: e }, kg = {
	behaviour: Q(ln, un),
	presentation: Q(Gt),
	"opt-in": Q(dn, fn),
	consumer: Q(an, on),
	schema: Q(Xt, Zt)
}, Ag = {
	union: Q(pt),
	json: Q(It),
	string: Q(kt),
	textarea: Q(yn),
	markdown: Q(It),
	number: Q(At),
	boolean: Q(jt),
	enum: Q(Ot),
	lookup: Q(pt),
	date: Q(ht),
	"string-map": Q(Xt, Zt),
	array: Q(Mt),
	object: Q(Nt),
	display: Q(yn),
	link: Q(ut),
	custom: Q(an, on)
}, jg = {
	string: Ag.string,
	integer: Ag.number,
	number: Ag.number,
	boolean: Ag.boolean,
	array: Ag.array,
	object: Ag.object,
	null: Q(Pt)
}, Mg = {
	enum: Ag.enum,
	const: Q(Jt, Yt),
	default: Q(Jt, Yt),
	format: Q(sn, cn),
	pattern: Q(hn, gn),
	readOnly: Q(pn, mn),
	writeOnly: Q(Sn, Cn),
	"x-hidden": Q($e),
	"x-hidden-by": Q($e),
	"x-disabled": Q(St),
	"x-on-change": kg.behaviour,
	"x-on-load": kg.behaviour,
	"x-clicky-lookup": Q(Qt, $t),
	"x-discriminator": Ag.enum,
	"x-order": Q(_n, vn),
	"x-clicky-order": Q(st),
	"x-clicky-hoist": Q(st),
	"x-item": Q(Et),
	"x-help": Q(vt),
	"x-help-display": Q(vt),
	"x-md-editor": Ag.markdown,
	"x-columns": Q(me),
	"x-col-span": Q(me),
	"x-layout": Q(dt),
	"x-object-display": Q(dt),
	"x-label-position": Q(dt)
};
function Ng(e, t) {
	return e === "type" && typeof t == "string" ? jg[t] : Mg[e];
}
var Pg = {
	hide: Q($e),
	show: Q(Qe),
	enable: Q(Ct),
	disable: Q(St),
	require: Q(Ut),
	optional: Q(yt),
	reset: Q(Kt),
	set: Q(bn, xn),
	patch: Q(en, tn)
}, Fg = {
	const: Q(Jt, Yt),
	enum: Ag.enum,
	not: Q(Wt),
	expr: Q(nn, rn),
	schema: Q(Xt, Zt),
	always: Q(yt)
}, Ig = {
	"x-hidden": "behaviour",
	"x-hidden-by": "behaviour",
	"x-disabled": "behaviour",
	"x-on-change": "behaviour",
	"x-on-load": "behaviour",
	"x-clicky-lookup": "behaviour",
	"x-discriminator": "behaviour",
	"x-order": "presentation",
	"x-clicky-order": "presentation",
	"x-clicky-hoist": "presentation",
	"x-enum-labels": "presentation",
	"x-enum-icons": "presentation",
	"x-enum-descriptions": "presentation",
	"x-enum-tones": "presentation",
	"x-enum-display": "presentation",
	"x-array-display": "presentation",
	"x-help-display": "presentation",
	"x-help": "presentation",
	"x-item": "presentation",
	"x-number-display": "presentation",
	"x-md-editor": "presentation",
	"x-label-classes": "presentation",
	"x-input-classes": "presentation",
	"x-input-prefix": "presentation",
	"x-input-suffix": "presentation",
	"x-input-prefix-icon": "presentation",
	"x-input-suffix-icon": "presentation",
	"x-label-position": "presentation",
	"x-layout": "presentation",
	"x-object-display": "presentation",
	"x-col-span": "presentation",
	"x-columns": "presentation",
	"x-column-min-width": "presentation",
	"x-columns-max-width": "presentation",
	"x-classes": "presentation",
	"x-icon": "presentation",
	"x-table-max-columns": "presentation",
	"x-clicky-extensions": "opt-in",
	"x-clicky-component": "opt-in",
	"x-clicky-default-source": "opt-in",
	"x-clicky-unit": "opt-in"
}, Lg = [
	"behaviour",
	"presentation",
	"opt-in",
	"consumer",
	"schema"
], Rg = /* @__PURE__ */ new Set(["title", "description"]);
function zg(e) {
	let t = /* @__PURE__ */ new Map();
	for (let [n, r] of Object.entries(e)) {
		if (Rg.has(n)) continue;
		let e = Ig[n] ?? (n.startsWith("x-") ? "consumer" : "schema");
		t.set(e, [...t.get(e) ?? [], [n, r]]);
	}
	return Lg.filter((e) => t.has(e)).map((e) => ({
		group: e,
		entries: t.get(e)
	}));
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-debug-listeners.tsx
var Bg = [
	"hide",
	"show",
	"enable",
	"disable",
	"require",
	"optional",
	"reset",
	"set",
	"patch"
], Vg = {
	true: {
		label: "holds",
		className: "text-emerald-700 dark:text-emerald-300"
	},
	false: {
		label: "does not hold",
		className: "text-muted-foreground"
	},
	unknown: {
		label: "needs evaluator",
		className: "text-amber-700 dark:text-amber-300"
	}
};
function Hg({ listeners: e, subject: t }) {
	return /* @__PURE__ */ (0, J.jsx)("ol", {
		className: "mt-0.5 space-y-1",
		children: e.map((e, n) => {
			let r = qg(e.when, t);
			return /* @__PURE__ */ (0, J.jsxs)("li", {
				"data-debug-listener": "",
				"data-active": r,
				className: q("space-y-0.5 rounded border px-1.5 py-1", r === "true" ? "border-emerald-500/50 bg-emerald-500/5" : "border-border"),
				children: [
					/* @__PURE__ */ (0, J.jsxs)("div", {
						className: "flex flex-wrap items-center gap-1",
						children: [
							/* @__PURE__ */ (0, J.jsx)("span", {
								className: "w-10 shrink-0 text-sky-700 dark:text-sky-300",
								children: "when"
							}),
							Kg(e.when).map((e, t) => /* @__PURE__ */ (0, J.jsxs)("span", {
								"data-condition": "",
								className: "inline-flex items-center gap-1 rounded bg-muted px-1",
								children: [/* @__PURE__ */ (0, J.jsx)(Og, { glyph: e.glyph }), e.text]
							}, t)),
							/* @__PURE__ */ (0, J.jsx)("span", {
								className: q("ml-auto font-sans text-[10px]", Vg[r].className),
								children: Vg[r].label
							})
						]
					}),
					/* @__PURE__ */ (0, J.jsx)(Ug, {
						actions: e,
						branch: "then",
						dimmed: r === "false"
					}),
					e.else && /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)("div", {
						className: "w-10 text-sky-700 dark:text-sky-300",
						children: "else"
					}), /* @__PURE__ */ (0, J.jsx)(Ug, {
						actions: e.else,
						branch: "else",
						dimmed: r === "true"
					})] })
				]
			}, n);
		})
	});
}
function Ug({ actions: e, branch: t, dimmed: n }) {
	let r = Bg.flatMap((t) => Wg(t, e).map((e) => ({
		verb: t,
		...e
	})));
	return /* @__PURE__ */ (0, J.jsxs)("ul", {
		[`data-${t}`]: "",
		className: q("space-y-0.5 pl-2", n && "opacity-50"),
		children: [r.length === 0 && /* @__PURE__ */ (0, J.jsx)("li", {
			className: "text-muted-foreground",
			children: "nothing"
		}), r.map(({ verb: e, detail: t, targets: n }) => /* @__PURE__ */ (0, J.jsxs)("li", {
			"data-action": e,
			className: "flex items-start gap-1",
			children: [
				/* @__PURE__ */ (0, J.jsx)(Og, {
					glyph: Pg[e],
					className: "mt-px"
				}),
				/* @__PURE__ */ (0, J.jsx)("span", {
					className: "w-12 shrink-0",
					children: e
				}),
				/* @__PURE__ */ (0, J.jsxs)("div", {
					className: "flex min-w-0 flex-1 flex-wrap items-center gap-1",
					children: [t !== void 0 && /* @__PURE__ */ (0, J.jsx)("span", {
						"data-detail": "",
						className: "text-muted-foreground",
						children: t
					}), n.map((e) => /* @__PURE__ */ (0, J.jsx)("span", {
						"data-target": "",
						className: "rounded bg-muted px-1",
						children: e
					}, e))]
				})
			]
		}, `${e}${t ?? ""}`))]
	});
}
function Wg(e, t) {
	if (e === "set") return Gg(Object.entries(t.set ?? {}).map(([e, t]) => [e, `= ${JSON.stringify(t)}`]));
	if (e === "patch") return Gg(Object.entries(t.patch ?? {}).map(([e, t]) => [e, Object.entries(t).map(([e, t]) => `${e}: ${JSON.stringify(t)}`).join(", ")]));
	let n = t[e] ?? [];
	return n.length > 0 ? [{ targets: n }] : [];
}
function Gg(e) {
	let t = /* @__PURE__ */ new Map();
	for (let [n, r] of e) t.set(r, [...t.get(r) ?? [], n]);
	return [...t].map(([e, t]) => ({
		detail: e,
		targets: t
	}));
}
function Kg(e) {
	if (!e) return [{
		glyph: Fg.always,
		text: "always"
	}];
	let { const: t, enum: n, not: r, expr: i, ...a } = e, o = [];
	return t !== void 0 && o.push({
		glyph: Fg.const,
		text: `= ${JSON.stringify(t)}`
	}), n !== void 0 && o.push({
		glyph: Fg.enum,
		text: `∈ ${JSON.stringify(n)}`
	}), r?.const !== void 0 && o.push({
		glyph: Fg.not,
		text: `≠ ${JSON.stringify(r.const)}`
	}), r?.enum !== void 0 && o.push({
		glyph: Fg.not,
		text: `∉ ${JSON.stringify(r.enum)}`
	}), r && r.const === void 0 && r.enum === void 0 && o.push({
		glyph: Fg.not,
		text: `not ${JSON.stringify(r)}`
	}), i !== void 0 && o.push({
		glyph: Fg.expr,
		text: `expr ${i}`
	}), Object.keys(a).length > 0 && o.push({
		glyph: Fg.schema,
		text: JSON.stringify(a)
	}), o;
}
function qg(e, t) {
	if (!e) return "true";
	let { expr: n, ...r } = e;
	if (Object.keys(r).length > 0) {
		let e = C(r, t);
		if (e === !1) return "false";
		if (e === void 0) return "unknown";
	}
	return n === void 0 ? "true" : "unknown";
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-debug-keywords.tsx
var Jg = {
	behaviour: "Behaviour",
	presentation: "Presentation",
	"opt-in": "Opt-in extensions",
	consumer: "Consumer extensions",
	schema: "Schema"
}, Yg = {
	"opt-in": "read only when the host passes the matching create*FormExtensions()",
	consumer: "not read by the form"
}, Xg = /* @__PURE__ */ new Set([
	"properties",
	"items",
	"allOf",
	"anyOf",
	"oneOf",
	"$defs",
	"patternProperties",
	"additionalProperties"
]), Zg = /* @__PURE__ */ new Set([
	"x-icon",
	"x-input-prefix-icon",
	"x-input-suffix-icon"
]), Qg = [
	"x-enum-labels",
	"x-enum-icons",
	"x-enum-descriptions",
	"x-enum-tones"
];
function $g({ prop: e, value: t }) {
	return /* @__PURE__ */ (0, J.jsx)(J.Fragment, { children: zg(e).map(({ group: n, entries: r }) => {
		let i = r.filter(([e]) => Qg.includes(e)), a = r.filter(([e]) => !Qg.includes(e));
		return /* @__PURE__ */ (0, J.jsxs)("section", {
			role: "group",
			"aria-label": Jg[n],
			className: "space-y-0.5 border-t border-border pt-1.5",
			children: [
				/* @__PURE__ */ (0, J.jsxs)("div", {
					className: "flex items-center gap-1 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground",
					children: [
						/* @__PURE__ */ (0, J.jsx)(Og, { glyph: kg[n] }),
						Jg[n],
						Yg[n] && /* @__PURE__ */ (0, J.jsxs)("span", {
							className: "font-normal normal-case tracking-normal",
							children: ["· ", Yg[n]]
						})
					]
				}),
				i.length > 0 && /* @__PURE__ */ (0, J.jsx)(a_, { prop: e }),
				/* @__PURE__ */ (0, J.jsx)("ul", {
					className: "space-y-0.5 font-mono text-[11px]",
					children: a.map(([e, n]) => /* @__PURE__ */ (0, J.jsxs)("li", {
						className: "flex gap-1",
						children: [/* @__PURE__ */ (0, J.jsx)(Og, {
							glyph: Ng(e, n),
							className: "mt-px"
						}), /* @__PURE__ */ (0, J.jsxs)("div", {
							className: "min-w-0 flex-1 break-words",
							children: [
								/* @__PURE__ */ (0, J.jsx)("span", {
									className: "text-muted-foreground",
									children: e
								}),
								" ",
								e_(e, n, t)
							]
						})]
					}, e))
				})
			]
		}, n);
	}) });
}
function e_(e, t, n) {
	if (e === "x-on-change" || e === "x-on-load") return Array.isArray(t) ? /* @__PURE__ */ (0, J.jsx)(Hg, {
		listeners: t,
		subject: n
	}) : /* @__PURE__ */ (0, J.jsx)(t_, { value: t });
	if (e === "x-clicky-lookup") return /* @__PURE__ */ (0, J.jsx)(n_, { descriptor: t });
	if (e === "x-item") return /* @__PURE__ */ (0, J.jsx)(r_, { spec: t });
	if (e === "x-help") return /* @__PURE__ */ (0, J.jsx)(i_, { help: t });
	if (Zg.has(e) && typeof t == "string") return /* @__PURE__ */ (0, J.jsxs)("span", {
		className: "inline-flex items-center gap-1",
		children: [/* @__PURE__ */ (0, J.jsx)(o, {
			icon: t,
			className: "size-3.5"
		}), t]
	});
	if (Xg.has(e)) {
		if (Array.isArray(t)) return `[${t.length}]`;
		if (typeof t == "object" && t) return `{${Object.keys(t).join(", ")}}`;
	}
	return typeof t == "string" ? t : Array.isArray(t) && t.every((e) => typeof e == "string") ? t.join(", ") : /* @__PURE__ */ (0, J.jsx)(t_, { value: t });
}
function t_({ value: e }) {
	return typeof e != "object" || !e ? /* @__PURE__ */ (0, J.jsx)("span", { children: JSON.stringify(e) }) : /* @__PURE__ */ (0, J.jsx)("pre", {
		className: "mt-0.5 max-h-40 overflow-auto whitespace-pre-wrap rounded bg-muted/60 p-1",
		children: JSON.stringify(e, null, 2)
	});
}
function n_({ descriptor: e }) {
	let { scope: t, hierarchy: n } = e, r = [
		`filter ${e.filter}`,
		e.searchParam && `search ${e.searchParam}`,
		e.multi && "multi-select",
		t && `scope ${t.param} ← ${t.from}${t.map ? ` via ${JSON.stringify(t.map)}` : ""}`,
		n && `tree split on ${JSON.stringify(n.delimiters)}`
	].filter(Boolean);
	return /* @__PURE__ */ (0, J.jsxs)("span", { children: [
		"GET ",
		e.url,
		" ",
		r.join(" · ")
	] });
}
function r_({ spec: e }) {
	let t = e.summary?.map((e) => typeof e == "string" ? e : (e.pattern ?? "{}").replace("{}", e.property)), n = [
		e.title && `title ${e.title.join(" | ")}${e.fallback ? ` (else ${e.fallback})` : ""}`,
		t && `summary ${t.join(" · ")}`,
		e.glyph && `glyph ${e.glyph}`,
		e.badge && `badge ${e.badge}`,
		e.flag && `flag ${e.flag}`,
		e.actions && `actions ${e.actions.length > 0 ? e.actions.join(", ") : "none"}`,
		e.noun && `noun ${e.noun}${e.nounPlural ? `/${e.nounPlural}` : ""}`,
		e.empty && `empty "${e.empty}"`
	].filter(Boolean);
	return /* @__PURE__ */ (0, J.jsx)("span", { children: n.join(" · ") });
}
function i_({ help: e }) {
	let t = [e.source, e.section].filter(Boolean).join(" › ");
	return /* @__PURE__ */ (0, J.jsxs)("span", { children: [t && /* @__PURE__ */ (0, J.jsxs)("span", {
		className: "text-muted-foreground",
		children: [
			"[",
			t,
			"] "
		]
	}), /* @__PURE__ */ (0, J.jsx)("span", {
		className: "font-sans",
		children: e.body
	})] });
}
function a_({ prop: e }) {
	let t = e["x-enum-labels"] ?? {}, n = e["x-enum-icons"] ?? {}, r = e["x-enum-descriptions"] ?? {}, i = e["x-enum-tones"] ?? {}, a = (e.enum ?? []).map(String), s = [...new Set([
		t,
		n,
		r,
		i
	].flatMap((e) => Object.keys(e)))], c = [...a, ...s.filter((e) => !a.includes(e))], l = e.enum !== void 0, u = Object.keys(n).length > 0;
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex gap-1 font-mono text-[11px]",
		children: [/* @__PURE__ */ (0, J.jsx)(Og, {
			glyph: Ag.enum,
			className: "mt-px"
		}), /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "min-w-0 flex-1",
			children: [/* @__PURE__ */ (0, J.jsx)("div", {
				className: "text-muted-foreground",
				children: "x-enum-*"
			}), /* @__PURE__ */ (0, J.jsxs)("table", {
				className: "w-full",
				children: [/* @__PURE__ */ (0, J.jsx)("thead", {
					className: "text-left text-[10px] text-muted-foreground",
					children: /* @__PURE__ */ (0, J.jsxs)("tr", { children: [
						/* @__PURE__ */ (0, J.jsx)("th", {
							className: "pr-2 font-normal",
							children: "value"
						}),
						/* @__PURE__ */ (0, J.jsx)("th", {
							className: "pr-2 font-normal",
							children: "label"
						}),
						/* @__PURE__ */ (0, J.jsx)("th", {
							className: "pr-2 font-normal",
							children: "description"
						}),
						/* @__PURE__ */ (0, J.jsx)("th", {
							className: "font-normal",
							children: "tone"
						})
					] })
				}), /* @__PURE__ */ (0, J.jsx)("tbody", { children: c.map((e) => {
					let s = i[e];
					return /* @__PURE__ */ (0, J.jsxs)("tr", {
						className: "align-top",
						children: [
							/* @__PURE__ */ (0, J.jsx)("td", {
								className: "pr-2",
								children: /* @__PURE__ */ (0, J.jsxs)("span", {
									className: "inline-flex items-center gap-1",
									children: [
										u && /* @__PURE__ */ (0, J.jsx)(o, {
											icon: n[e],
											className: "size-3.5"
										}),
										e,
										l && !a.includes(e) && /* @__PURE__ */ (0, J.jsx)("span", {
											className: "text-amber-700 dark:text-amber-300",
											children: " ⚠ not in enum"
										})
									]
								})
							}),
							/* @__PURE__ */ (0, J.jsx)("td", {
								className: "pr-2",
								children: t[e] ?? ""
							}),
							/* @__PURE__ */ (0, J.jsx)("td", {
								className: "pr-2 font-sans",
								children: r[e] ?? ""
							}),
							/* @__PURE__ */ (0, J.jsx)("td", { children: s && /* @__PURE__ */ (0, J.jsxs)("span", {
								className: "inline-flex items-center gap-1",
								children: [/* @__PURE__ */ (0, J.jsx)("span", { className: q("size-2 rounded-full", vm(s) ? _m[s] : "outline outline-1 outline-rose-500") }), s]
							}) })
						]
					}, e);
				}) })]
			})]
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-debug-badge.tsx
var o_ = "inline-flex items-center gap-1 rounded px-1 text-[11px] text-sky-700 hover:bg-muted dark:text-sky-300";
function s_(e) {
	let { fieldKey: t, instancePath: n, field: r, prop: i, hiddenReason: a } = e, [o, s] = (0, Y.useState)(!1), [l, u] = (0, Y.useState)(!1), d = a !== void 0;
	return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)(Hp, {
		placement: "right",
		trigger: /* @__PURE__ */ (0, J.jsx)("button", {
			type: "button",
			"aria-label": `Debug ${t}`,
			className: q("inline-flex h-4 w-4 shrink-0 items-center justify-center rounded text-[11px]", d ? "text-amber-600 dark:text-amber-400" : "text-muted-foreground hover:text-foreground"),
			children: /* @__PURE__ */ (0, J.jsx)(c, { icon: d ? $e : Vt })
		}),
		children: /* @__PURE__ */ (0, J.jsx)(c_, {
			info: e,
			expanded: o,
			onToggle: () => s(!o),
			onOpenDialog: () => u(!0)
		})
	}), /* @__PURE__ */ (0, J.jsx)(Xd, {
		open: l,
		onClose: () => u(!1),
		title: i.title || t,
		subtitle: `${n || "/"} · ${r.kind}`,
		size: "lg",
		children: /* @__PURE__ */ (0, J.jsx)("div", {
			"data-testid": "json-schema-form-debug-card-body",
			"data-clamped": "false",
			className: "text-xs",
			children: /* @__PURE__ */ (0, J.jsx)(l_, {
				...e,
				withHeading: !1
			})
		})
	})] });
}
function c_({ info: e, expanded: t, onToggle: n, onOpenDialog: r }) {
	let i = (0, Y.useRef)(null), [a, o] = (0, Y.useState)(!1);
	(0, Y.useLayoutEffect)(() => {
		let e = i.current;
		e && !t && o(e.scrollHeight > e.clientHeight + 1);
	}, [
		t,
		e.prop,
		e.field.value
	]);
	let s = !t;
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		"data-testid": "json-schema-form-debug-card",
		className: "max-w-md space-y-1 whitespace-normal text-xs",
		children: [/* @__PURE__ */ (0, J.jsxs)("div", {
			className: "relative",
			children: [/* @__PURE__ */ (0, J.jsx)("div", {
				ref: i,
				"data-testid": "json-schema-form-debug-card-body",
				"data-clamped": String(s),
				className: s ? "max-h-80 overflow-hidden" : "max-h-[70vh] overflow-y-auto",
				children: /* @__PURE__ */ (0, J.jsx)(l_, { ...e })
			}), s && a && /* @__PURE__ */ (0, J.jsx)("div", {
				"aria-hidden": !0,
				className: "pointer-events-none absolute inset-x-0 bottom-0 h-10 bg-gradient-to-t from-background"
			})]
		}), /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "flex items-center gap-1 border-t border-border pt-1",
			children: [a && /* @__PURE__ */ (0, J.jsxs)("button", {
				type: "button",
				"aria-expanded": t,
				onClick: n,
				className: o_,
				children: [/* @__PURE__ */ (0, J.jsx)(c, { icon: t ? Ke : pe }), t ? "Show less" : "Show more"]
			}), /* @__PURE__ */ (0, J.jsxs)("button", {
				type: "button",
				onClick: r,
				className: q(o_, "ml-auto"),
				children: [/* @__PURE__ */ (0, J.jsx)(c, { icon: et }), "Open in dialog"]
			})]
		})]
	});
}
function l_({ fieldKey: e, instancePath: t, field: n, prop: r, hiddenReason: i, source: a, withHeading: o = !0 }) {
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "space-y-1.5",
		children: [
			o && /* @__PURE__ */ (0, J.jsxs)("div", {
				className: "flex items-start gap-1.5",
				children: [/* @__PURE__ */ (0, J.jsx)(Og, {
					glyph: Ag[n.kind],
					className: "mt-0.5 size-4"
				}), /* @__PURE__ */ (0, J.jsxs)("div", {
					className: "min-w-0",
					children: [/* @__PURE__ */ (0, J.jsx)("div", {
						className: "font-medium",
						children: r.title || e
					}), /* @__PURE__ */ (0, J.jsxs)("div", {
						className: "font-mono text-[11px] text-muted-foreground",
						children: [
							t || "/",
							" · ",
							n.kind
						]
					})]
				})]
			}),
			i !== void 0 && /* @__PURE__ */ (0, J.jsxs)("p", {
				className: "flex items-center gap-1 text-amber-700 dark:text-amber-300",
				children: [
					/* @__PURE__ */ (0, J.jsx)(c, { icon: $e }),
					" ",
					i
				]
			}),
			a !== void 0 && /* @__PURE__ */ (0, J.jsxs)("p", {
				"data-debug-extension": a,
				className: "text-muted-foreground",
				children: [
					"Rendered by form extension ",
					/* @__PURE__ */ (0, J.jsx)("span", {
						className: "font-mono",
						children: a
					}),
					" (format ",
					/* @__PURE__ */ (0, J.jsx)("span", {
						className: "font-mono",
						children: n.format
					}),
					")"
				]
			}),
			r.description && /* @__PURE__ */ (0, J.jsx)("p", {
				className: "text-muted-foreground",
				children: r.description
			}),
			/* @__PURE__ */ (0, J.jsx)($g, {
				prop: r,
				value: n.value
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-debug.tsx
function u_(e, t) {
	let n = t.hiddenReason !== void 0, r = /* @__PURE__ */ (0, J.jsx)(s_, { ...t }), i = /* @__PURE__ */ (0, J.jsxs)("div", {
		"data-debug-field": t.fieldKey,
		...n ? { "data-debug-hidden": t.hiddenReason } : {},
		className: q("min-w-0", n && "rounded-md opacity-60 outline-dashed outline-1 outline-offset-2 outline-muted-foreground/60"),
		children: [e.label === null ? r : null, e.value]
	});
	return e.label === null ? {
		label: null,
		value: i
	} : {
		label: /* @__PURE__ */ (0, J.jsxs)("span", {
			className: q("inline-flex min-w-0 items-center gap-1", n && "opacity-60"),
			children: [e.label, r]
		}),
		value: i
	};
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-color.tsx
var d_ = /^#[0-9a-f]{6}$/i, f_ = "#000000";
function p_({ label: e, value: t, disabled: n, onPick: r }) {
	let i = d_.test(t);
	return /* @__PURE__ */ (0, J.jsx)("input", {
		type: "color",
		"aria-label": `${e} color picker`,
		"data-jsf-color-swatch": !0,
		value: i ? t.toLowerCase() : f_,
		disabled: n,
		onChange: (e) => r(e.target.value),
		className: q("size-5 cursor-pointer rounded border border-border bg-transparent p-0 disabled:cursor-not-allowed disabled:opacity-60", !i && "opacity-50")
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-number-stepper.tsx
function m_({ field: e, fieldId: t, readOnly: n, size: r }) {
	let i = u(e.value), a = i.trim() === "" ? 0 : Number(i), o = Number.isFinite(a), s = e.schema.multipleOf ?? 1, c = e.schema.minimum ?? -Infinity, l = e.schema.maximum ?? Infinity, d = (t) => {
		if (n || !o) return;
		let r = Math.min(l, Math.max(c, a + t * s));
		e.onChange(e.coerceNumber === !1 ? String(r) : r);
	};
	return /* @__PURE__ */ (0, J.jsx)(po, {
		clearable: e.nullable === !0,
		clearLabel: `Clear ${e.label} to null`,
		onClear: () => e.onChange(null),
		id: t,
		type: "number",
		step: "any",
		inputMode: "decimal",
		"data-jsf-input": !0,
		...B(e, t),
		...e.schema.minimum === void 0 ? {} : { min: e.schema.minimum },
		...e.schema.maximum === void 0 ? {} : { max: e.schema.maximum },
		className: T[r],
		inputClassName: q(O[r], "[appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none", e.inputClassName),
		value: i,
		disabled: n,
		placeholder: R(e.schema),
		prefix: e.prefix,
		suffix: /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [
			e.suffix ?? (e.unit ? /* @__PURE__ */ (0, J.jsx)("span", {
				className: "text-muted-foreground",
				children: e.unit
			}) : null),
			/* @__PURE__ */ (0, J.jsx)(fe, {
				type: "button",
				variant: "ghost",
				size: "xs",
				"aria-label": `Decrease ${e.label}`,
				disabled: n || !o || a <= c,
				onClick: () => d(-1),
				children: "−"
			}),
			/* @__PURE__ */ (0, J.jsx)(fe, {
				type: "button",
				variant: "ghost",
				size: "xs",
				"aria-label": `Increase ${e.label}`,
				disabled: n || !o || a >= l,
				onClick: () => d(1),
				children: "+"
			})
		] }),
		onChange: (t) => e.onChange(e.coerceNumber !== !1 && t.trim() !== "" && String(Number(t)) === t.trim() ? Number(t) : t),
		onBlur: (t) => {
			let n = t.currentTarget.value;
			e.coerceNumber !== !1 && n.trim() !== "" && Number.isFinite(Number(n)) && e.onChange(Number(n));
		},
		onKeyDown: (e) => {
			(e.key === "ArrowUp" || e.key === "ArrowDown") && (e.preventDefault(), d(e.key === "ArrowUp" ? 1 : -1));
		}
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-enum-choice.tsx
var h_ = "default";
function g_(e, t) {
	let n = e.default;
	if (typeof n != "string" && typeof n != "number" && typeof n != "boolean") return;
	let r = String(n);
	return t.some((e) => e.value === r) ? r : void 0;
}
function __(e, t) {
	let n = [e.description, t ? `(${h_})` : void 0].filter(Boolean);
	return n.length > 0 ? n.join(" ") : void 0;
}
var v_ = {
	xs: "sm",
	sm: "sm",
	md: "md",
	lg: "lg",
	xl: "lg"
};
function y_({ field: e, fieldId: t, readOnly: n, options: r, value: i, size: a }) {
	let o = r.some((e) => e.description) ? "lg" : v_[a], s = g_(e.schema, r);
	return /* @__PURE__ */ (0, J.jsx)("div", {
		id: t,
		"data-jsf-input": !0,
		className: q(e.inputClassName),
		children: /* @__PURE__ */ (0, J.jsx)(uf, {
			"aria-label": e.label,
			value: i,
			onChange: (t) => e.onChange(t),
			size: o,
			wrap: !0,
			options: r.map((e) => {
				let t = e.value === s, r = __(e, t) ?? (e.name ? e.label : void 0);
				return {
					id: e.value,
					label: e.name ?? e.label,
					...typeof e.icon == "string" || (0, Y.isValidElement)(e.icon) ? { icon: e.icon } : {},
					...e.description ? { description: e.description } : {},
					...r ? { title: r } : {},
					...t ? {
						marker: h_,
						implied: !0
					} : {},
					...n ? { disabled: !0 } : {}
				};
			})
		})
	});
}
function b_({ field: e, fieldId: t, readOnly: n, options: r, value: i, size: a }) {
	let o = g_(e.schema, r), s = i === "" ? o : void 0;
	return /* @__PURE__ */ (0, J.jsx)("div", {
		role: "radiogroup",
		"aria-label": e.label,
		id: t,
		"data-jsf-input": !0,
		className: q("inline-flex flex-wrap items-center gap-1 rounded-md border border-input bg-background p-0.5", e.inputClassName),
		children: r.map((r) => {
			let c = r.value === i, l = r.value === s, u = r.value === o, d = u ? `${t}-default` : void 0;
			return /* @__PURE__ */ (0, J.jsxs)("label", {
				title: __(r, u),
				"data-implied": l ? "true" : void 0,
				className: q("inline-flex cursor-pointer select-none items-center rounded px-2.5 py-1", O[a], c ? "bg-primary text-primary-foreground" : l ? "text-foreground outline-dashed outline-1 -outline-offset-1 outline-primary/60 hover:bg-accent" : "text-muted-foreground hover:bg-accent hover:text-foreground", n && "cursor-not-allowed opacity-60"),
				children: [
					/* @__PURE__ */ (0, J.jsx)("input", {
						type: "radio",
						name: t,
						className: "sr-only",
						value: r.value,
						checked: c,
						disabled: n,
						"aria-describedby": d,
						onChange: () => e.onChange(r.value)
					}),
					r.label,
					u && /* @__PURE__ */ (0, J.jsx)("span", {
						id: d,
						"aria-hidden": "true",
						className: "ml-1.5 text-[11px] font-normal opacity-70",
						children: h_
					})
				]
			}, r.value);
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-grid.tsx
function x_({ field: e, fieldId: t, readOnly: n, size: r }) {
	let i = u(e.value), [a, o] = (0, Y.useState)(""), s = U(e.options ?? [], i);
	if (n) {
		let e = s.find((e) => e.value === i);
		return /* @__PURE__ */ (0, J.jsx)("div", {
			id: t,
			"data-jsf-input": !0,
			children: /* @__PURE__ */ (0, J.jsx)(S_, {
				option: e ?? {
					value: i,
					label: i
				},
				checked: !0,
				readOnly: !0,
				size: r,
				onSelect: () => {}
			})
		});
	}
	let c = a.trim().toLowerCase(), l = c ? s.filter((e) => e.label.toLowerCase().includes(c) || e.value.toLowerCase().includes(c)) : s;
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex w-full flex-col gap-2",
		children: [/* @__PURE__ */ (0, J.jsx)("input", {
			type: "search",
			"aria-label": `Filter ${e.label}`,
			placeholder: "Filter…",
			value: a,
			onChange: (e) => o(e.target.value),
			className: V(r)
		}), /* @__PURE__ */ (0, J.jsxs)("div", {
			role: "radiogroup",
			"aria-label": e.label,
			id: t,
			"data-jsf-input": !0,
			className: "grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4",
			children: [l.map((t) => /* @__PURE__ */ (0, J.jsx)(S_, {
				option: t,
				checked: t.value === i,
				readOnly: !1,
				size: r,
				onSelect: () => e.onChange(t.value)
			}, t.value)), l.length === 0 && /* @__PURE__ */ (0, J.jsx)("div", {
				className: q("col-span-full py-4 text-center text-muted-foreground", O[r]),
				children: "No matches"
			})]
		})]
	});
}
function S_({ option: e, checked: t, readOnly: n, size: r, onSelect: i }) {
	return /* @__PURE__ */ (0, J.jsxs)("button", {
		type: "button",
		role: "radio",
		"aria-checked": t,
		disabled: n,
		onClick: i,
		className: q("flex flex-col items-center justify-center gap-1.5 rounded-md border p-3 text-center transition-colors", O[r], t ? "border-primary bg-primary/5 ring-1 ring-primary" : "border-input hover:border-primary/50 hover:bg-accent", n && "cursor-default opacity-80"),
		children: [/* @__PURE__ */ (0, J.jsx)(o, {
			icon: e.icon,
			className: "text-2xl text-foreground"
		}), /* @__PURE__ */ (0, J.jsx)("span", {
			className: "break-words leading-tight",
			children: e.label
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/components/lookup-option-badges.tsx
function C_({ badges: e }) {
	return /* @__PURE__ */ (0, J.jsx)("span", {
		className: "flex shrink-0 items-center gap-1",
		children: e.map((e) => {
			let t = e.tone ? hm[e.tone] : "", n = typeof e.icon == "string" && e.icon !== "";
			return /* @__PURE__ */ (0, J.jsxs)(ge, {
				size: "xxs",
				variant: "soft",
				clickToCopy: !1,
				...n ? { icon: e.icon } : {},
				className: q("border-transparent", t),
				children: [!n && e.icon ? /* @__PURE__ */ (0, J.jsx)(o, {
					icon: e.icon,
					className: "mr-0.5"
				}) : null, e.label]
			}, e.label);
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/lookup-option-view.tsx
function w_(e) {
	let t = e.icon ? e.tone ? /* @__PURE__ */ (0, J.jsx)("span", {
		className: q("inline-flex items-center justify-center rounded p-0.5", hm[e.tone]),
		children: /* @__PURE__ */ (0, J.jsx)(o, { icon: e.icon })
	}) : e.icon : void 0;
	return {
		value: e.value,
		label: e.label,
		selectedLabel: e.label,
		...e.description ? {
			description: e.description,
			descriptionLayout: "inline"
		} : {},
		...t ? { icon: t } : {},
		...e.group ? { group: e.group } : {},
		...e.badges && e.badges.length > 0 ? { trailing: /* @__PURE__ */ (0, J.jsx)(C_, { badges: e.badges }) } : {},
		...e.description ? { title: `${e.label}\n${e.description}` } : {}
	};
}
// ../../../clicky-ui/packages/ui/src/connections/connectionTypes.ts
function T_(e) {
	return e.split("_").map((e) => e.charAt(0).toUpperCase() + e.slice(1)).join(" ");
}
function E_(e, t) {
	return t?.[e] ?? T_(e);
}
function D_(e, t) {
	if (typeof e != "string" || !t) return;
	let n = Object.keys(t).filter((n) => t[n] === e);
	return n.length === 1 ? n[0] : void 0;
}
function O_(e) {
	let t = e.properties?.type;
	return {
		icons: k_(t?.["x-enum-icons"], "x-enum-icons"),
		labels: k_(t?.["x-enum-labels"], "x-enum-labels")
	};
}
function k_(e, t) {
	if (e === void 0) return {};
	if (typeof e != "object" || !e || Array.isArray(e)) throw Error(`connection schema type.${t} must be an object of type -> string`);
	return Object.fromEntries(Object.entries(e).map(([e, n]) => {
		if (typeof n != "string") throw Error(`connection schema type.${t}.${e} must be a string, got ${JSON.stringify(n)}`);
		return [e, n];
	}));
}
function A_({ name: e, namespace: t }) {
	return `connection://${t ? `${t}/` : ""}${e}`;
}
// ../../../clicky-ui/packages/ui/src/connections/ConnectionPicker.tsx
function j_({ options: e, value: t, onChange: n, types: r, typeIcons: i, typeLabels: a, onCreate: o, ...s }) {
	let l = r?.length === 1 ? r[0] : void 0, u = (0, Y.useMemo)(() => {
		let t = e.map((e) => {
			let t = D_(e.icon, i);
			return {
				option: e,
				group: l ? void 0 : e.group ?? (t ? E_(t, a) : void 0)
			};
		});
		return [...new Set(t.map((e) => e.group))].flatMap((e) => t.filter((t) => t.group === e)).map(({ option: e, group: t }) => {
			let { group: n, ...r } = w_(e);
			return t ? {
				...r,
				group: t
			} : r;
		});
	}, [
		e,
		l,
		i,
		a
	]), d = t !== "" && !e.some((e) => e.value === t) ? [{
		value: t,
		label: t,
		selectedLabel: t
	}, ...u] : u, f = l ? `${E_(l, a)} connection` : "connection", p = o ? /* @__PURE__ */ (0, J.jsxs)("button", {
		type: "button",
		className: "flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-primary hover:bg-muted/60",
		onMouseDown: (e) => e.preventDefault(),
		onClick: () => o(l ? { type: l } : {}),
		children: [/* @__PURE__ */ (0, J.jsx)(X, { className: "size-4" }), `New ${f}…`]
	}) : void 0;
	return /* @__PURE__ */ (0, J.jsx)(ss, {
		...s,
		options: d,
		value: t,
		onChange: n,
		allowCustomValue: !1,
		placeholder: `Select a ${f}`,
		ariaLabel: l ? `${E_(l, a)} connection` : "Connection",
		menuWidth: 540,
		...l ? { prefix: /* @__PURE__ */ (0, J.jsx)(c, {
			name: i?.[l] ?? l,
			className: "text-base"
		}) } : {},
		...p ? { footer: p } : {}
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-connection.tsx
var M_ = (0, Y.lazy)(async () => ({ default: (await import("./CreateConnectionDialog-Cx3R71jU.js")).CreateConnectionDialog }));
function N_(e) {
	let t = e.schema["x-connection-types"];
	if (t !== void 0) {
		if (!Array.isArray(t) || t.length === 0 || t.some((e) => typeof e != "string" || e === "")) throw Error(`${e.key}: x-connection-types must be a non-empty list of connection type names, got ${JSON.stringify(t)}`);
		return t;
	}
}
function P_({ field: e, fieldId: t, readOnly: n, size: r, options: i, loading: a, onSearch: o }) {
	let s = gg(), [c, l] = (0, Y.useState)(), d = N_(e);
	return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)(j_, {
		id: t,
		options: i,
		value: u(e.value),
		onChange: (t) => e.onChange(t),
		disabled: n,
		size: r,
		loading: a,
		onSearch: o,
		...d ? { types: d } : {},
		...s ? { onCreate: l } : {},
		...e.nullable ? {
			clearable: !0,
			clearLabel: `Clear ${e.label} to null`,
			onClear: () => e.onChange(null)
		} : {},
		...te(e, t)
	}), c && s ? /* @__PURE__ */ (0, J.jsx)(Y.Suspense, {
		fallback: null,
		children: /* @__PURE__ */ (0, J.jsx)(M_, {
			apiBase: s,
			existing: i.map((e) => e.value),
			...c.type ? { type: c.type } : {},
			onClose: () => l(void 0),
			onCreated: (t) => {
				e.onChange(t), l(void 0), o("");
			}
		})
	}) : null] });
}
function F_(e, t) {
	if (!e) return [];
	if (!t) return [e];
	let n = [], r = "";
	for (let i of e) t.includes(i) ? (r && n.push(r), r = "") : r += i;
	return r && n.push(r), n;
}
function I_(e, t, n = {}) {
	let r = [], i = /* @__PURE__ */ new Map();
	for (let n of e) {
		let e = t(n);
		if (e.length === 0) continue;
		let a = r, o = "", s;
		e.forEach((t, n) => {
			o = o ? `${o}/${t}` : t;
			let r = i.get(o);
			r || (r = {
				key: o,
				label: t,
				path: e.slice(0, n + 1),
				items: [],
				children: []
			}, i.set(o, r), a.push(r)), a = r.children, s = r;
		}), s?.items.push(n);
	}
	return n.compare && L_(r, n.compare), r;
}
function L_(e, t) {
	e.sort(t);
	for (let n of e) L_(n.children, t);
}
function R_(e, t) {
	let n = e.children.length > 0;
	return n === t.children.length > 0 ? e.label.localeCompare(t.label) : n ? -1 : 1;
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-lookup-tree.tsx
function z_(e) {
	return e.items[0];
}
function B_(e, t) {
	return I_(e, (e) => F_(e.label || e.value, t.delimiters));
}
function V_({ node: e }) {
	let t = z_(e);
	return /* @__PURE__ */ (0, J.jsx)("span", {
		className: q("min-w-0 flex-1 truncate text-sm", t ? "text-foreground" : "text-muted-foreground"),
		children: e.label
	});
}
function H_({ field: e, fieldId: t, readOnly: n, size: r, options: i, loading: a, hierarchy: o }) {
	let s = (0, Y.useMemo)(() => B_(i, o), [i, o]), l = e.lookup?.multi === !0, u = a ? "Loading…" : "Select…";
	if (!l) {
		let i = typeof e.value == "string" ? e.value : "", a = U_(s, i) ?? null;
		return /* @__PURE__ */ (0, J.jsx)("div", {
			id: t,
			"data-jsf-input": !0,
			children: /* @__PURE__ */ (0, J.jsx)(sf, {
				...e.nullable ? {
					clearable: !0,
					clearLabel: `Clear ${e.label} to null`,
					onClear: () => e.onChange(null)
				} : {},
				roots: s,
				getKey: (e) => e.key,
				getChildren: (e) => e.children,
				getSearchText: (e) => e.key,
				defaultOpen: () => !1,
				renderRow: ({ node: e }) => /* @__PURE__ */ (0, J.jsx)(V_, { node: e }),
				isSelectable: (e) => z_(e) !== void 0,
				onSelect: (t) => e.onChange(z_(t)?.value),
				selected: a,
				revealSelected: !0,
				ariaLabel: e.label,
				disabled: n,
				size: r,
				placeholder: u,
				...i ? { label: i } : {}
			})
		});
	}
	let d = ee(e.value);
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		id: t,
		"data-jsf-input": !0,
		className: "flex flex-col gap-1.5",
		children: [d.length > 0 && /* @__PURE__ */ (0, J.jsx)("ul", {
			className: "flex flex-wrap gap-1",
			children: d.map((t) => /* @__PURE__ */ (0, J.jsxs)("li", {
				className: "inline-flex items-center gap-1 rounded-md border border-border bg-muted/40 py-0.5 pl-2 pr-1 text-xs",
				children: [/* @__PURE__ */ (0, J.jsx)("span", {
					className: "truncate",
					children: t
				}), !n && /* @__PURE__ */ (0, J.jsx)("button", {
					type: "button",
					"aria-label": `Remove ${t}`,
					onClick: () => e.onChange(d.filter((e) => e !== t)),
					className: "rounded p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground",
					children: /* @__PURE__ */ (0, J.jsx)(c, {
						icon: j,
						className: "text-[0.7rem]"
					})
				})]
			}, t))
		}), !n && /* @__PURE__ */ (0, J.jsx)(sf, {
			roots: s,
			getKey: (e) => e.key,
			getChildren: (e) => e.children,
			getSearchText: (e) => e.key,
			defaultOpen: () => !1,
			renderRow: ({ node: e }) => /* @__PURE__ */ (0, J.jsx)(V_, { node: e }),
			isSelectable: (e) => {
				let t = z_(e);
				return t !== void 0 && !d.includes(t.value);
			},
			onSelect: (t) => {
				let n = z_(t);
				n && e.onChange([...d, n.value]);
			},
			ariaLabel: e.label,
			size: r,
			renderTrigger: ({ triggerRef: t, toggle: n }) => /* @__PURE__ */ (0, J.jsxs)("button", {
				ref: t,
				type: "button",
				onClick: n,
				className: "inline-flex w-fit items-center gap-1.5 rounded-md border border-dashed border-input px-2 py-1 text-xs text-muted-foreground transition-colors hover:border-input hover:text-foreground",
				children: [/* @__PURE__ */ (0, J.jsx)(c, {
					icon: X,
					className: "text-[0.7rem]"
				}), a ? "Loading…" : `Add ${e.label ?? "value"}`]
			})
		})]
	});
}
function U_(e, t) {
	if (!t) return null;
	for (let n of e) {
		if (z_(n)?.value === t) return n;
		let e = U_(n.children, t);
		if (e) return e;
	}
	return null;
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-tags-combobox.tsx
function W_({ field: e, fieldId: t, readOnly: n, size: r, options: i, itemType: a, separators: o, loading: s, onSearch: c }) {
	let l = K_(e.value), [u, d] = (0, Y.useState)(), f = `${t}-refusal`, p = te(e, t), m = a === "string" && e.itemSchema ? e.itemSchema : void 0;
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "w-full",
		children: [/* @__PURE__ */ (0, J.jsx)(ss, {
			multiple: !0,
			variant: "tags",
			...e.nullable ? {
				clearable: !0,
				clearLabel: `Clear ${e.label} to null`,
				onClear: () => e.onChange(null)
			} : {},
			id: t,
			options: i.length === 0 ? [] : ho(i, l),
			value: l,
			onChange: (t) => e.onChange(t.map((e) => q_(e, a))),
			disabled: n,
			...n ? {} : { inputPlaceholder: R(e.schema) || (a === "integer" ? "Add integer…" : a === "number" ? "Add number…" : a === "string" ? "Add tag…" : "Choose values…") },
			size: r,
			allowCustomValue: e.allowCustomValue ?? a !== void 0,
			prefix: e.prefix,
			suffix: e.suffix,
			...o ? { separators: o } : {},
			...s === void 0 ? {} : { loading: s },
			...c ? { onSearch: c } : {},
			...p,
			...u ? {
				invalid: !0,
				describedBy: [p.describedBy, f].filter(Boolean).join(" ")
			} : {},
			...m ? {
				onCreate: (t) => {
					let n = G_(m, t.value, e.label);
					return d(n), n ? null : t.value;
				},
				onInputValueChange: () => d(void 0)
			} : {},
			...e.inputClassName ? { className: e.inputClassName } : {},
			...R(e.schema) ? { placeholder: R(e.schema) } : {}
		}), u && /* @__PURE__ */ (0, J.jsx)("p", {
			id: f,
			role: "alert",
			className: "mt-1 text-xs text-destructive",
			children: u
		})]
	});
}
function G_(e, t, n) {
	let { pattern: r, minLength: i, maxLength: a } = e;
	if (!((typeof r != "string" || new RegExp(r, "u").test(t)) && (typeof i != "number" || t.length >= i) && (typeof a != "number" || t.length <= a))) return `“${t}” is not a valid ${e.title || n}${e.description ? `: ${e.description}` : typeof r == "string" ? `; it must match ${r}` : ""}`;
}
function K_(e) {
	return Array.isArray(e) ? e.filter((e) => typeof e == "string" || typeof e == "number").map(String) : [];
}
function q_(e, t) {
	if (t !== "integer" && t !== "number") return e;
	let n = e.trim();
	return n === "" || !Number.isFinite(Number(n)) ? e : Number(n);
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-fields.tsx
var J_ = (0, Y.lazy)(async () => ({ default: (await import("./MdxEditorField-Cnm98CYm.js")).MdxEditorField }));
function Y_({ field: e, fieldId: t, size: n }) {
	let r = u(e.value), i = e.schema.format === "password", a = i && r ? "••••••••" : e.kind === "date" && r ? Pf(r) : r;
	return /* @__PURE__ */ (0, J.jsx)("span", {
		id: t,
		"data-jsf-input": !0,
		"data-jsf-readonly": !0,
		className: q("flex items-center text-foreground", T[n], O[n]),
		title: i ? void 0 : r || void 0,
		children: a || /* @__PURE__ */ (0, J.jsx)("span", {
			className: "text-muted-foreground",
			children: "—"
		})
	});
}
function X_({ field: e, fieldId: t, readOnly: n, size: r }) {
	let i = e.prefix ?? (e.schema.format === "color" ? /* @__PURE__ */ (0, J.jsx)(p_, {
		label: e.label,
		value: u(e.value),
		disabled: n,
		onPick: e.onChange
	}) : void 0);
	return /* @__PURE__ */ (0, J.jsx)(po, {
		prefix: i,
		suffix: e.suffix,
		clearable: e.nullable === !0,
		clearLabel: `Clear ${e.label} to null`,
		onClear: () => e.onChange(null),
		id: t,
		type: e.schema.format === "password" ? "password" : "text",
		autoComplete: e.schema.format === "password" ? "new-password" : void 0,
		"data-jsf-input": !0,
		...B(e, t),
		className: T[r],
		inputClassName: q(O[r], e.inputClassName),
		value: u(e.value),
		disabled: n,
		placeholder: R(e.schema),
		onChange: (t) => e.onChange(t)
	});
}
function Z_({ field: e, fieldId: t, readOnly: n, size: r }) {
	if (e.schema["x-number-display"] === "slider" && typeof e.schema.maximum == "number") return /* @__PURE__ */ (0, J.jsx)($_, {
		field: e,
		fieldId: t,
		readOnly: n,
		size: r
	});
	if (e.schema["x-number-display"] === "stepper") return /* @__PURE__ */ (0, J.jsx)(m_, {
		field: e,
		fieldId: t,
		readOnly: n,
		size: r
	});
	let i = e.coerceNumber !== !1, a = typeof e.schema.multipleOf == "number" ? e.schema.multipleOf : void 0, o = a !== void 0 && !(e.allowCustomValue && typeof e.value == "string" && e.value.includes("{{")), s = e.suffix ?? (e.unit ? /* @__PURE__ */ (0, J.jsx)("span", {
		className: "pointer-events-none select-none text-muted-foreground",
		children: e.unit
	}) : void 0);
	return /* @__PURE__ */ (0, J.jsx)(po, {
		prefix: e.prefix,
		suffix: s,
		clearable: e.nullable === !0,
		clearLabel: `Clear ${e.label} to null`,
		onClear: () => e.onChange(null),
		id: t,
		type: o ? "number" : "text",
		inputMode: "decimal",
		...o ? {
			step: a,
			...typeof e.schema.minimum == "number" ? { min: e.schema.minimum } : {},
			...typeof e.schema.maximum == "number" ? { max: e.schema.maximum } : {}
		} : {},
		"data-jsf-input": !0,
		...B(e, t),
		className: T[r],
		inputClassName: q(O[r], e.inputClassName),
		value: u(e.value),
		disabled: n,
		placeholder: R(e.schema),
		onChange: (t) => {
			i && t.trim() !== "" && String(Number(t)) === t.trim() ? e.onChange(Number(t)) : e.onChange(t);
		},
		onBlur: (t) => {
			let n = t.target.value;
			i && n.trim() !== "" && Number.isFinite(Number(n)) && e.onChange(Number(n));
		}
	});
}
var Q_ = "absolute inset-0 h-6 w-full cursor-pointer appearance-none bg-transparent focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 [&::-webkit-slider-runnable-track]:h-1 [&::-webkit-slider-runnable-track]:rounded-full [&::-webkit-slider-runnable-track]:bg-transparent [&::-webkit-slider-thumb]:size-3.5 [&::-webkit-slider-thumb]:-mt-[5px] [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:border [&::-webkit-slider-thumb]:border-background/80 [&::-webkit-slider-thumb]:bg-primary [&::-webkit-slider-thumb]:shadow-sm [&::-moz-range-track]:h-1 [&::-moz-range-track]:rounded-full [&::-moz-range-track]:bg-transparent [&::-moz-range-thumb]:size-3.5 [&::-moz-range-thumb]:appearance-none [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:border-0 [&::-moz-range-thumb]:bg-primary [&::-moz-range-thumb]:shadow-sm";
function $_({ field: e, fieldId: t, readOnly: n, size: r }) {
	let i = typeof e.schema.minimum == "number" ? e.schema.minimum : 0, a = typeof e.schema.maximum == "number" ? e.schema.maximum : 100, o = typeof e.schema.multipleOf == "number" ? e.schema.multipleOf : 1, s = typeof e.value == "number" ? e.value : Number(u(e.value)), c = Number.isFinite(s), l = c ? Math.min(a, Math.max(i, s)) : i, d = a > i ? (l - i) / (a - i) * 100 : 0;
	return /* @__PURE__ */ (0, J.jsx)(F, {
		clearable: e.nullable === !0 && !n && e.value != null,
		clearLabel: `Clear ${e.label} to null`,
		onClear: () => e.onChange(null),
		children: /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "flex items-center gap-3",
			children: [/* @__PURE__ */ (0, J.jsxs)("div", {
				className: "relative h-6 min-w-0 flex-1",
				children: [
					/* @__PURE__ */ (0, J.jsx)("div", { className: "absolute inset-x-0 top-1/2 h-1 -translate-y-1/2 rounded-full bg-border" }),
					/* @__PURE__ */ (0, J.jsx)("div", {
						className: "absolute left-0 top-1/2 h-1 -translate-y-1/2 rounded-full bg-primary/70",
						style: { width: `${d}%` }
					}),
					/* @__PURE__ */ (0, J.jsx)("input", {
						id: t,
						"data-jsf-input": !0,
						type: "range",
						"aria-label": e.label,
						...B(e, t),
						min: i,
						max: a,
						step: o,
						value: l,
						disabled: n,
						className: q(Q_, e.inputClassName),
						onChange: (t) => e.onChange(Number(t.target.value))
					})
				]
			}), /* @__PURE__ */ (0, J.jsx)("span", {
				className: q("w-16 shrink-0 text-right tabular-nums", O[r]),
				children: c ? l.toLocaleString() : "—"
			})]
		})
	});
}
function ev({ field: e, fieldId: t, readOnly: n, size: r }) {
	let i = u(e.value);
	return !n && e.allowCustomValue && i.includes("{{") ? /* @__PURE__ */ (0, J.jsx)(X_, {
		field: e,
		fieldId: t,
		readOnly: n,
		size: r
	}) : n ? /* @__PURE__ */ (0, J.jsx)("div", {
		id: t,
		"data-jsf-input": !0,
		className: q("flex items-center text-foreground", T[r], O[r]),
		title: i || void 0,
		children: i ? Pf(i) : /* @__PURE__ */ (0, J.jsx)("span", {
			className: "text-muted-foreground",
			children: "—"
		})
	}) : /* @__PURE__ */ (0, J.jsx)(hf, {
		clearable: e.nullable === !0,
		clearLabel: `Clear ${e.label} to null`,
		onClear: () => e.onChange(null),
		disabled: n,
		mode: e.dateFormat === "date-time" ? "datetime" : "date",
		id: t,
		"aria-label": e.label,
		"data-jsf-input": !0,
		...B(e, t),
		inputClassName: q(V(r), e.prefix && "pl-8", e.suffix ? "pr-14" : "pr-8"),
		value: i,
		onChange: (t) => e.onChange(t),
		placeholder: R(e.schema),
		prefix: e.prefix,
		suffix: e.suffix
	});
}
function tv({ field: e, fieldId: t, readOnly: n, size: r }) {
	return typeof e.value == "boolean" || e.value === void 0 || e.value === null ? /* @__PURE__ */ (0, J.jsxs)("div", {
		className: q("flex items-center gap-1", T[r]),
		children: [e.prefix, /* @__PURE__ */ (0, J.jsx)(df, {
			id: t,
			"data-jsf-input": !0,
			...B(e, t),
			checked: e.value === !0,
			disabled: n,
			onChange: e.onChange
		})]
	}) : /* @__PURE__ */ (0, J.jsx)(X_, {
		field: e,
		fieldId: t,
		readOnly: n,
		size: r
	});
}
function nv({ field: e, fieldId: t, readOnly: n, size: r }) {
	let i = u(e.value), a = U(e.options ?? [], i);
	return e.display === "grid" ? /* @__PURE__ */ (0, J.jsx)(F, {
		clearable: e.nullable === !0 && !n && e.value != null,
		clearLabel: `Clear ${e.label} to null`,
		onClear: () => e.onChange(null),
		children: /* @__PURE__ */ (0, J.jsx)(x_, {
			field: e,
			fieldId: t,
			readOnly: n,
			size: r
		})
	}) : e.display === "radio" ? /* @__PURE__ */ (0, J.jsx)(F, {
		clearable: e.nullable === !0 && !n && e.value != null,
		clearLabel: `Clear ${e.label} to null`,
		onClear: () => e.onChange(null),
		children: /* @__PURE__ */ (0, J.jsx)(b_, {
			field: e,
			fieldId: t,
			readOnly: n,
			options: a,
			value: i,
			size: r
		})
	}) : e.display === "segmented" ? /* @__PURE__ */ (0, J.jsx)(F, {
		clearable: e.nullable === !0 && !n && e.value != null,
		clearLabel: `Clear ${e.label} to null`,
		onClear: () => e.onChange(null),
		children: /* @__PURE__ */ (0, J.jsx)(y_, {
			field: e,
			fieldId: t,
			readOnly: n,
			options: a,
			value: i,
			size: r
		})
	}) : /* @__PURE__ */ (0, J.jsx)(ss, {
		id: t,
		options: a,
		value: i,
		disabled: n,
		size: r,
		allowCustomValue: e.allowCustomValue ?? !1,
		...e.nullable ? { clearable: !0 } : {},
		clearLabel: `Clear ${e.label} to null`,
		...e.nullable ? { onClear: () => e.onChange(null) } : {},
		onChange: (t) => e.onChange(t),
		prefix: e.prefix,
		suffix: e.suffix,
		...te(e, t),
		...e.inputClassName ? { className: e.inputClassName } : {},
		...R(e.schema) ? { placeholder: R(e.schema) } : {}
	});
}
function rv(e) {
	let [t, n] = (0, Y.useState)(), r = `${e.fieldId}-lookup-error`;
	return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)(iv, {
		...e,
		onError: n,
		...t ? { errorId: r } : {}
	}), t && /* @__PURE__ */ (0, J.jsx)("p", {
		id: r,
		role: "alert",
		"data-jsf-lookup-error": !0,
		className: "text-xs text-destructive",
		children: t
	})] });
}
function iv({ field: e, fieldId: t, readOnly: n, size: r, rootValue: i, onError: a, errorId: o }) {
	let s = e.lookup, c = ug(s), l = u(e.value), [d, f] = (0, Y.useState)([]), [p, m] = (0, Y.useState)(!1), h = (0, Y.useRef)(i);
	h.current = i;
	let g = (0, Y.useCallback)((e) => {
		if (!s || !c) return;
		m(!0);
		let t = h.current;
		c({
			descriptor: s,
			query: e,
			...t ? { rootValue: t } : {}
		}).then((e) => {
			f(e), a(void 0);
		}).catch((e) => a(e instanceof Error ? e.message : String(e))).finally(() => m(!1));
	}, [
		s,
		c,
		a
	]), _ = s ? JSON.stringify(fg(s, i)) : "";
	return (0, Y.useEffect)(() => {
		s && c ? g("") : f([]);
	}, [_, g]), s?.hierarchy ? /* @__PURE__ */ (0, J.jsx)(H_, {
		field: e,
		fieldId: t,
		readOnly: n,
		size: r,
		options: d,
		loading: p,
		hierarchy: s.hierarchy
	}) : s?.multi ? /* @__PURE__ */ (0, J.jsx)(W_, {
		field: e,
		fieldId: t,
		readOnly: n,
		size: r,
		options: d,
		loading: p,
		...c ? { onSearch: g } : {}
	}) : !s || !c ? /* @__PURE__ */ (0, J.jsx)(ss, {
		id: t,
		options: U([], l),
		...e.nullable ? {
			clearable: !0,
			clearLabel: `Clear ${e.label} to null`,
			onClear: () => e.onChange(null)
		} : {},
		value: l,
		disabled: n,
		size: r,
		allowCustomValue: !0,
		onChange: (t) => e.onChange(t),
		...te(e, t),
		...R(e.schema) ? { placeholder: R(e.schema) } : {}
	}) : e.schema.format === "connection" ? /* @__PURE__ */ (0, J.jsx)(P_, {
		field: e,
		fieldId: t,
		readOnly: n,
		size: r,
		options: d,
		loading: p,
		onSearch: g
	}) : /* @__PURE__ */ (0, J.jsx)(ss, {
		id: t,
		options: U(d, l).map(w_),
		...e.nullable ? {
			clearable: !0,
			clearLabel: `Clear ${e.label} to null`,
			onClear: () => e.onChange(null)
		} : {},
		value: l,
		disabled: n,
		size: r,
		loading: p,
		allowCustomValue: e.allowCustomValue ?? !1,
		onSearch: g,
		onChange: (t) => e.onChange(t),
		...te(e, t),
		...o ? {
			invalid: !0,
			describedBy: o
		} : {},
		...R(e.schema) ? { placeholder: R(e.schema) } : {}
	});
}
function av({ field: e, fieldId: t, readOnly: n, size: r }) {
	let i = u(e.value);
	return n ? /* @__PURE__ */ (0, J.jsx)("span", {
		id: t,
		"data-jsf-input": !0,
		"data-jsf-readonly": !0,
		className: q("block whitespace-pre-wrap break-words text-foreground", O[r]),
		title: i || void 0,
		children: i || /* @__PURE__ */ (0, J.jsx)("span", {
			className: "text-muted-foreground",
			children: "—"
		})
	}) : /* @__PURE__ */ (0, J.jsx)(po, {
		as: "textarea",
		prefix: e.prefix,
		suffix: e.suffix,
		clearable: e.nullable === !0,
		clearLabel: `Clear ${e.label} to null`,
		onClear: () => e.onChange(null),
		id: t,
		"data-jsf-input": !0,
		...B(e, t),
		rows: 4,
		className: "h-auto min-h-[5rem]",
		inputClassName: q(O[r], e.inputClassName),
		value: i,
		disabled: n,
		placeholder: R(e.schema),
		onChange: (t) => e.onChange(t)
	});
}
function ov({ field: e, fieldId: t, readOnly: n, size: r }) {
	let i = /* @__PURE__ */ (0, J.jsx)("textarea", {
		id: t,
		"data-jsf-input": !0,
		...B(e, t),
		rows: 6,
		className: q(V(r), "h-auto resize-y font-mono", e.prefix && "pl-8", e.suffix && "pr-8"),
		value: u(e.value),
		disabled: n,
		placeholder: R(e.schema),
		onChange: (t) => e.onChange(t.target.value)
	});
	return /* @__PURE__ */ (0, J.jsx)(Y.Suspense, {
		fallback: i,
		children: /* @__PURE__ */ (0, J.jsx)(J_, {
			clearable: e.nullable === !0,
			clearLabel: `Clear ${e.label} to null`,
			onClear: () => e.onChange(null),
			id: t,
			value: u(e.value),
			readOnly: n,
			size: r,
			placeholder: R(e.schema),
			onChange: (t) => e.onChange(t),
			...e.markdownOptions
		})
	});
}
function sv({ field: e, size: t }) {
	let n = e.displayVariant ?? "text";
	if (n === "divider") return /* @__PURE__ */ (0, J.jsx)("hr", {
		"data-jsf-input": !0,
		className: "my-1 w-full border-t border-border"
	});
	if (n === "spacer") return /* @__PURE__ */ (0, J.jsx)("div", {
		"data-jsf-input": !0,
		"aria-hidden": !0,
		className: "h-2"
	});
	if (n === "heading") return /* @__PURE__ */ (0, J.jsx)("div", {
		"data-jsf-input": !0,
		className: q("border-b border-border pb-1 font-semibold text-foreground", O[t]),
		children: e.label
	});
	let r = e.description ?? u(e.value) ?? e.label;
	return /* @__PURE__ */ (0, J.jsx)("p", {
		"data-jsf-input": !0,
		className: q("text-muted-foreground", O[t]),
		children: r || e.label
	});
}
function cv({ field: e, fieldId: t, size: n }) {
	let r = u(e.value), i = e.href ?? (/^https?:\/\//i.test(r) ? r : void 0);
	return i ? /* @__PURE__ */ (0, J.jsx)("a", {
		id: t,
		"data-jsf-input": !0,
		href: i,
		target: "_blank",
		rel: "noreferrer noopener",
		className: q("inline-flex items-center text-primary underline underline-offset-2", T[n], O[n]),
		children: r || e.label || i
	}) : /* @__PURE__ */ (0, J.jsx)(Y_, {
		field: e,
		fieldId: t,
		size: n
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-item-row.tsx
function lv({ glyph: e }) {
	return e ? /* @__PURE__ */ (0, J.jsx)("span", {
		title: e.label,
		className: q("inline-flex size-6 shrink-0 items-center justify-center rounded-md ring-1 ring-inset", hm[e.tone]),
		children: e.icon != null && /* @__PURE__ */ (0, J.jsx)(o, {
			icon: e.icon,
			className: "text-[13px]"
		})
	}) : null;
}
function uv({ badge: e }) {
	return e ? /* @__PURE__ */ (0, J.jsxs)("span", {
		className: "inline-flex shrink-0 items-center gap-1 rounded-full border border-border px-2 py-0.5 text-[11px] text-muted-foreground",
		children: [e.icon != null && /* @__PURE__ */ (0, J.jsx)(o, {
			icon: e.icon,
			className: "text-[13px]"
		}), e.label]
	}) : null;
}
function dv({ badges: e }) {
	return /* @__PURE__ */ (0, J.jsx)(J.Fragment, { children: e?.map((e) => /* @__PURE__ */ (0, J.jsx)(uv, { badge: { label: e } }, e)) });
}
function fv({ count: e }) {
	if (e <= 0) return null;
	let t = e === 1 ? "1 error" : `${e} errors`;
	return /* @__PURE__ */ (0, J.jsxs)("span", {
		title: t,
		className: "inline-flex shrink-0 items-center gap-1 text-xs text-destructive",
		children: [/* @__PURE__ */ (0, J.jsx)(c, {
			icon: gt,
			className: "text-sm"
		}), t]
	});
}
function pv() {
	return /* @__PURE__ */ (0, J.jsx)("span", {
		title: "Required",
		className: "inline-flex shrink-0 items-center text-destructive",
		children: /* @__PURE__ */ (0, J.jsx)(c, {
			icon: Ut,
			className: "text-[11px]"
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-cards-array.tsx
function mv({ field: e, ctx: t, readOnly: n }) {
	let r = Array.isArray(e.value) ? e.value : [], i = e.itemSchema ?? { type: "object" }, a = e.itemSpec ?? Cm(e.schema, i), o = Nm(a, e.schema), s = {
		...t,
		readOnly: n,
		depth: t.depth + 1,
		...e.helpDisplay ? { layout: {
			...t.layout,
			help: e.helpDisplay
		} } : {}
	};
	function u(t) {
		e.onChange(t);
	}
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex flex-col gap-2",
		children: [
			r.map((o, c) => {
				let d = e.itemSummary?.({
					item: o,
					index: c
				}) ?? km({
					item: o,
					index: c,
					spec: a,
					itemSchema: i
				});
				return /* @__PURE__ */ (0, J.jsxs)("article", {
					className: q("group rounded-lg border-l-4 border-border", t.presentation ? "bg-muted/20" : "border shadow-sm", d.glyph ? gm[d.glyph.tone] : void 0),
					children: [/* @__PURE__ */ (0, J.jsxs)("header", {
						className: q("flex flex-wrap items-center gap-2 rounded-tr-md border-b border-border bg-muted/40", t.presentation ? "px-2 py-1" : "px-3 py-2"),
						children: [
							/* @__PURE__ */ (0, J.jsx)("span", {
								className: "w-5 shrink-0 text-center font-mono text-xs text-muted-foreground",
								children: c + 1
							}),
							/* @__PURE__ */ (0, J.jsx)("span", {
								className: "shrink-0 text-sm font-semibold",
								children: d.title
							}),
							/* @__PURE__ */ (0, J.jsx)(uv, { badge: d.badge }),
							/* @__PURE__ */ (0, J.jsx)(dv, { badges: d.badges }),
							d.flagged && /* @__PURE__ */ (0, J.jsx)(pv, {}),
							d.summary && /* @__PURE__ */ (0, J.jsx)("code", {
								className: "truncate font-mono text-xs text-muted-foreground",
								children: d.summary
							}),
							!n && /* @__PURE__ */ (0, J.jsx)(wn, {
								label: d.title,
								index: c,
								count: r.length,
								size: t.size,
								...Em(a, "reorder") ? { onMove: (e) => u(l(r, c, e)) } : {},
								...Em(a, "duplicate") ? { onDuplicate: () => u(N(r, c)) } : {},
								...Em(a, "remove") && L(e, r.length) ? { onRemove: () => u(z(r, c)) } : {}
							})
						]
					}), /* @__PURE__ */ (0, J.jsx)("div", {
						className: t.presentation ? "p-2" : "p-3",
						children: t.render.renderFieldNodes({
							key: `${e.key}[${c}]`,
							prop: i,
							required: !1,
							value: o,
							onChange: (e) => u(oe(r, c, e)),
							instancePath: Xh(t.instancePath, c)
						}, s)?.value ?? null
					})]
				}, c);
			}),
			r.length === 0 && o && /* @__PURE__ */ (0, J.jsx)("p", {
				className: "text-xs leading-snug text-muted-foreground",
				children: o
			}),
			!n && ie(e, r.length) && /* @__PURE__ */ (0, J.jsxs)(fe, {
				type: "button",
				variant: "outline",
				className: q("gap-1.5 self-start", g[t.size]),
				onClick: () => u([...r, p(i)]),
				children: [/* @__PURE__ */ (0, J.jsx)(c, {
					icon: X,
					className: "text-sm"
				}), Am(a)]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-accordion-array.tsx
function hv({ field: e, ctx: t, readOnly: n, toolbar: r }) {
	let i = Array.isArray(e.value) ? e.value : [], a = e.itemSchema ?? { type: "object" }, o = e.itemSpec ?? Cm(e.schema, a), s = Nm(o, e.schema), c = Mm(o, i.length), l = {
		...t,
		readOnly: n,
		depth: t.depth + 1,
		...e.helpDisplay ? { layout: {
			...t.layout,
			help: e.helpDisplay
		} } : {}
	};
	function u(t, n) {
		return e.itemSummary?.({
			item: t,
			index: n
		}) ?? km({
			item: t,
			index: n,
			spec: o,
			itemSchema: a
		});
	}
	function d({ item: e, index: n, open: r }) {
		let i = u(e, n);
		return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [
			/* @__PURE__ */ (0, J.jsx)(lv, { glyph: i.glyph }),
			/* @__PURE__ */ (0, J.jsx)("span", {
				className: "shrink-0 text-sm font-medium",
				children: i.title
			}),
			/* @__PURE__ */ (0, J.jsx)(uv, { badge: i.badge }),
			/* @__PURE__ */ (0, J.jsx)(dv, { badges: i.badges }),
			i.flagged && /* @__PURE__ */ (0, J.jsx)(pv, {}),
			i.summary && /* @__PURE__ */ (0, J.jsx)("code", {
				className: "truncate font-mono text-xs text-muted-foreground",
				children: i.summary
			}),
			!r && /* @__PURE__ */ (0, J.jsx)(fv, { count: Zh(t.errors, Xh(t.instancePath, n)) })
		] });
	}
	return t.presentation ? /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "min-w-0 space-y-1",
		children: [
			c !== void 0 && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "text-xs text-muted-foreground",
				children: c
			}),
			i.map((e, n) => /* @__PURE__ */ (0, J.jsx)("div", {
				className: q("flex min-w-0 flex-wrap items-center gap-2 rounded px-2 py-1 even:bg-muted/40", ue[t.size]),
				children: d({
					item: e,
					index: n,
					open: !1
				})
			}, n)),
			i.length === 0 && s && /* @__PURE__ */ (0, J.jsx)("p", {
				className: "text-xs text-muted-foreground",
				children: s
			})
		]
	}) : /* @__PURE__ */ (0, J.jsx)(En, {
		items: i,
		onChange: (t) => e.onChange(t),
		...c === void 0 ? {} : { summary: c },
		...r === void 0 ? {} : { toolbar: r },
		size: t.size,
		readOnly: n,
		itemId: (e) => W(Xh(t.instancePath, e), t.idPrefix),
		focusSelector: "[data-jsf-input]",
		itemLabel: ({ item: e, index: t }) => u(e, t).title,
		allowReorder: Em(o, "reorder"),
		allowDuplicate: Em(o, "duplicate"),
		allowRemove: Em(o, "remove") && L(e, i.length),
		...ie(e, i.length) ? { onCreate: () => p(a) } : {},
		addLabel: Am(o),
		...s ? { addDescription: s } : {},
		renderHeader: d,
		renderBody: ({ item: n, index: r, onChange: i }) => t.render.renderFieldNodes({
			key: `${e.key}[${r}]`,
			prop: a,
			required: !1,
			value: n,
			onChange: i,
			instancePath: Xh(t.instancePath, r)
		}, l)?.value ?? null
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-extensions.ts
function gv(e, t, n, r) {
	return n.reduce((t, n) => n(e, t, r), t);
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-table-array.tsx
function _v({ field: e, ctx: t, readOnly: n, toolbar: r }) {
	let i = Array.isArray(e.value) ? e.value : [], a = e.itemSchema ?? { type: "object" }, o = Wh(e, t), s = o.valueCells, { columns: l, rows: u } = Gh(a, i, o), d = {
		...t,
		readOnly: n,
		depth: t.depth + 1,
		...s ? { hideReadOnlyFields: !1 } : {}
	}, f = !n && L(e, i.length);
	function m(n, r, c) {
		let l = w(n) ? n : {}, f = u[r]?.[c], p = f ? t.render.renderFieldNodes({
			key: `${e.key}[${r}].${c}`,
			prop: s ? {
				...f,
				readOnly: !0
			} : f,
			required: !1,
			value: l[c],
			onChange: (t) => e.onChange(oe(i, r, Nh(a, {
				root: o.rootValue ?? l,
				...o.evaluate ? { evaluate: o.evaluate } : {},
				value: l,
				key: c,
				next: t
			}))),
			instancePath: Xh(Xh(t.instancePath, r), c)
		}, d) : null;
		return p ? t.presentation ? /* @__PURE__ */ (0, J.jsx)("div", {
			className: q("flex min-w-0 items-center [overflow-wrap:anywhere] [&_span[data-jsf-readonly]]:h-auto [&_span[data-jsf-readonly]]:min-w-0", ue[t.size]),
			children: p.value
		}) : p.value : /* @__PURE__ */ (0, J.jsx)("span", {
			"data-jsf-cell-placeholder": !0,
			"aria-hidden": "true",
			className: q("block", ue[t.size])
		});
	}
	function h(e, n, r) {
		let i = {
			...t.rootValue ? { rootValue: t.rootValue } : {},
			...t.onRootChange ? { onRootChange: t.onRootChange } : {}
		}, a = dh({
			key: e,
			prop: n,
			required: !1,
			value: void 0,
			onChange: () => {}
		});
		for (let r of t.pre) {
			if (!a) break;
			a = r(a, {
				key: e,
				prop: n,
				value: void 0,
				...i
			});
		}
		return a ? gv(a, {
			label: r,
			value: null
		}, t.post, i).label : r;
	}
	let _ = /* @__PURE__ */ (0, J.jsxs)("div", {
		className: q("overflow-x-auto", !t.presentation && "rounded-md border border-input"),
		children: [/* @__PURE__ */ (0, J.jsxs)("table", {
			className: "w-full text-sm",
			children: [/* @__PURE__ */ (0, J.jsx)("thead", { children: /* @__PURE__ */ (0, J.jsxs)("tr", {
				className: q("border-b bg-muted/40 text-left", t.presentation ? "border-border" : "border-input"),
				children: [l.map(([e, n]) => /* @__PURE__ */ (0, J.jsx)("th", {
					className: "px-2 py-1 text-xs font-medium text-muted-foreground [overflow-wrap:anywhere]",
					children: h(e, n, t.presentation ? n.title || e : /* @__PURE__ */ (0, J.jsx)(vv, {
						name: e,
						schema: n
					}))
				}, e)), f && /* @__PURE__ */ (0, J.jsx)("th", { className: "w-10 px-2 py-1" })]
			}) }), /* @__PURE__ */ (0, J.jsx)("tbody", { children: i.map((r, a) => /* @__PURE__ */ (0, J.jsxs)("tr", {
				className: q("border-b last:border-b-0 align-top", t.presentation ? "border-border" : "border-input"),
				children: [l.map(([e]) => /* @__PURE__ */ (0, J.jsx)("td", {
					className: q("px-2 py-1", !t.presentation && !n && "min-w-40"),
					children: m(r, a, e)
				}, e)), f && /* @__PURE__ */ (0, J.jsx)("td", {
					className: "px-2 py-1",
					children: /* @__PURE__ */ (0, J.jsx)("button", {
						type: "button",
						"aria-label": `Remove item ${a + 1}`,
						className: q("inline-flex aspect-square items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-foreground", T[t.size]),
						onClick: () => e.onChange(z(i, a)),
						children: /* @__PURE__ */ (0, J.jsx)(c, {
							icon: Ht,
							className: "text-sm"
						})
					})
				})]
			}, a)) })]
		}), !n && ie(e, i.length) && /* @__PURE__ */ (0, J.jsx)("div", {
			className: "p-2",
			children: /* @__PURE__ */ (0, J.jsxs)(fe, {
				type: "button",
				variant: "outline",
				className: q("gap-1.5", g[t.size]),
				onClick: () => e.onChange([...i, p(a)]),
				children: [/* @__PURE__ */ (0, J.jsx)(c, {
					icon: X,
					className: "text-sm"
				}), "Add item"]
			})
		})]
	});
	if (r === void 0) return _;
	let v = Mm(e.itemSpec ?? Cm(e.schema, a), i.length);
	return r === null && v === void 0 ? _ : /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex flex-col gap-2",
		children: [/* @__PURE__ */ (0, J.jsxs)("div", {
			"data-jsf-array-summary": !0,
			className: "flex items-center justify-between gap-2",
			children: [/* @__PURE__ */ (0, J.jsx)("p", {
				className: q("text-muted-foreground", O[t.size]),
				children: v
			}), r != null && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "flex shrink-0 items-center",
				children: r
			})]
		}), _]
	});
}
function vv({ name: e, schema: t }) {
	let n = typeof t.title == "string" && t.title ? t.title : e, r = sh(t);
	return r ? /* @__PURE__ */ (0, J.jsx)(Hp, {
		placement: "bottom",
		trigger: /* @__PURE__ */ (0, J.jsx)("button", {
			type: "button",
			"aria-label": `${n} help`,
			className: "cursor-help border-b border-dotted border-current text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
			children: n
		}),
		cardClassName: "max-w-xs whitespace-normal text-left font-normal text-foreground",
		children: r
	}) : n;
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-object-array-view.tsx
var yv = {
	grid: "Grid",
	stack: "Stack form",
	inline: "Inline form"
}, bv = [
	"grid",
	"stack",
	"inline"
];
function xv({ field: e, ctx: t, readOnly: n }) {
	let [r, i] = (0, Y.useState)(() => qh(e, Uh(e, t))), a = Rp(t.instancePath ?? ""), o = t.presentation ? void 0 : /* @__PURE__ */ (0, J.jsx)(Sv, {
		label: e.label || e.key,
		view: r,
		onChange: i
	}), s = o && a ? null : o, c = o && a ? (0, Fe.createPortal)(o, a) : null, l = s === void 0 ? {} : { toolbar: s };
	if (r === "grid") return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [c, /* @__PURE__ */ (0, J.jsx)(_v, {
		field: e,
		ctx: t,
		readOnly: n,
		...l
	})] });
	let u = {
		...t,
		layout: {
			...t.layout,
			mode: r === "stack" ? "stacked" : "inline"
		}
	};
	return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [c, /* @__PURE__ */ (0, J.jsx)(hv, {
		field: e,
		ctx: u,
		readOnly: n,
		...l
	})] });
}
function Sv({ label: e, view: t, onChange: n }) {
	let r = `View options for ${e}`;
	return /* @__PURE__ */ (0, J.jsx)(uo, {
		icon: ct,
		hideChevron: !0,
		variant: "ghost",
		size: "icon",
		align: "right",
		title: r,
		menuLabel: r,
		items: bv.map((e) => ({
			label: e === t ? /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [yv[e], /* @__PURE__ */ (0, J.jsx)("span", {
				className: "sr-only",
				children: " (current view)"
			})] }) : yv[e],
			onSelect: () => n(e),
			...e === t ? { icon: bt } : {}
		}))
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-array.tsx
var Cv = [","];
function wv({ field: e, fieldId: t, ctx: n }) {
	let r = n.readOnly || e.readOnly === !0, i = Tv(e), a = ph(e.itemSchema), o = e.arrayDisplay === "stacked" || e.arrayDisplay === "list";
	if (e.arrayDisplay === "list") {
		let t = e.itemSchema?.enum !== void 0 && e.itemSchema.enum.every((e) => typeof e == "string");
		if (a !== "string" && !t) throw Error(`${e.key} uses x-array-display list without string items`);
		if (Array.isArray(e.value) && e.value.some((e) => typeof e != "string")) throw Error(`${e.key} compact list received a non-string value`);
	}
	if (e.arrayDisplay === "filter-pills" && i.length > 0) return /* @__PURE__ */ (0, J.jsx)(Ev, {
		field: e,
		fieldId: t,
		readOnly: r,
		size: n.size
	});
	if (!o && i.length > 0) return /* @__PURE__ */ (0, J.jsx)(W_, {
		field: e,
		fieldId: t,
		readOnly: r,
		size: n.size,
		options: i
	});
	if (!o && a) return /* @__PURE__ */ (0, J.jsx)(W_, {
		field: e,
		fieldId: t,
		readOnly: r,
		size: n.size,
		options: [],
		itemType: a,
		separators: Cv
	});
	if (H(e.itemSchema)) {
		if (e.arrayDisplay === "cards") return /* @__PURE__ */ (0, J.jsx)(mv, {
			field: e,
			ctx: n,
			readOnly: r
		});
		if (e.arrayDisplay !== "stacked") return /* @__PURE__ */ (0, J.jsx)(xv, {
			field: e,
			ctx: n,
			readOnly: r
		});
	}
	let s = Array.isArray(e.value) ? e.value : [], u = e.itemSchema ?? { type: "string" }, f = {
		...n,
		readOnly: r,
		depth: n.depth + 1
	};
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		id: t,
		role: "group",
		"aria-label": e.label,
		className: q("flex flex-col", !n.presentation && "rounded-md border border-input p-2", d[n.size]),
		children: [s.map((t, i) => /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "flex flex-wrap items-start gap-2",
			children: [/* @__PURE__ */ (0, J.jsx)(am, {
				layout: n.layout,
				size: n.size,
				columns: 1,
				className: "min-w-0 flex-1 basis-64",
				children: n.render.renderFieldRow({
					key: `${e.key}[${i}]`,
					prop: u,
					required: !1,
					value: t,
					onChange: (t) => e.onChange(oe(s, i, t)),
					instancePath: Xh(n.instancePath, i)
				}, f, { labelOverride: `Item ${i + 1}` })
			}), !r && /* @__PURE__ */ (0, J.jsx)(wn, {
				label: `item ${i + 1}`,
				index: i,
				count: s.length,
				size: n.size,
				reveal: !1,
				onMove: (t) => e.onChange(l(s, i, t)),
				...L(e, s.length) ? { onRemove: () => e.onChange(z(s, i)) } : {}
			})]
		}, i)), !r && ie(e, s.length) && /* @__PURE__ */ (0, J.jsxs)(fe, {
			type: "button",
			variant: "outline",
			className: q("self-start gap-1.5", g[n.size]),
			onClick: () => e.onChange([...s, p(u)]),
			children: [/* @__PURE__ */ (0, J.jsx)(c, {
				icon: X,
				className: "text-sm"
			}), "Add item"]
		})]
	});
}
function Tv(e) {
	let t = e.options ?? (Array.isArray(e.itemSchema?.enum) ? e.itemSchema.enum.map((e) => ({
		value: String(e),
		label: String(e)
	})) : []), n = /* @__PURE__ */ new Set();
	return t.filter((e) => !n.has(e.value) && (n.add(e.value), !0));
}
function Ev({ field: e, fieldId: t, readOnly: n, size: r }) {
	let i = Tv(e), a = new Set(i.map((e) => e.value)), o = ee(e.value).filter((e) => a.has(e)), s = o.length === 0, c = new Set(s ? i.map((e) => e.value) : o);
	function l(t) {
		let n = i.map((e) => e.value).filter((e) => t.has(e));
		e.onChange(n.length === i.length ? [] : n);
	}
	function u(e) {
		if (n) return;
		let t = new Set(c);
		t.has(e) ? t.delete(e) : t.add(e), l(t);
	}
	return /* @__PURE__ */ (0, J.jsx)("div", {
		id: t,
		"data-jsf-input": !0,
		role: "group",
		"aria-label": e.label,
		className: q("flex flex-wrap items-center gap-1.5 rounded-md border border-input bg-transparent px-2 py-1 shadow-sm", ue[r]),
		children: i.map((e) => /* @__PURE__ */ (0, J.jsx)(Io, {
			label: e.label,
			mode: c.has(e.value) ? "active" : "neutral",
			title: s ? `${e.label} enabled by default` : e.label,
			...n ? {} : { onClick: () => u(e.value) },
			className: "max-w-full"
		}, e.value))
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-exclusive-keys.ts
function Dv(e) {
	let t = e.oneOf;
	if (!t || t.length < 2) return;
	let n = t.map((e) => {
		let t = e.required;
		return Object.keys(e).length === 1 && Array.isArray(t) && t.length === 1 ? t[0] : void 0;
	});
	if (!n.some((t) => t === void 0 || !e.properties?.[t])) return n;
}
function Ov({ value: e, keys: t, next: n }) {
	let r = t.find((t) => e[t] !== void 0), i = Object.fromEntries(Object.entries(e).filter(([e]) => !t.includes(e)));
	return r === void 0 ? i : {
		...i,
		[n]: e[r]
	};
}
// ../../../clicky-ui/packages/ui/src/data/properties-utils.ts
function kv(e) {
	return e.replace(/[_-]+/g, " ").replace(/\.+/g, " ").replace(/\b\w/g, (e) => e.toUpperCase());
}
// ../../../clicky-ui/packages/ui/src/data/PropertyValue.tsx
function Av({ item: e, density: t }) {
	let { value: n, type: r, unit: i } = e;
	if (n == null) return /* @__PURE__ */ (0, J.jsx)("span", {
		className: "font-mono italic text-muted-foreground",
		children: "null"
	});
	if (r === "json" || r === void 0 && typeof n == "object") return typeof n == "string" ? /* @__PURE__ */ (0, J.jsx)(Ad, {
		source: n,
		format: "json",
		density: t
	}) : /* @__PURE__ */ (0, J.jsx)(Ad, {
		data: n,
		format: "json",
		density: t
	});
	if (r === "bytes" || r === "duration" || r === "number" || i !== void 0) {
		if (typeof n != "number" || !Number.isFinite(n)) throw Error(`Property ${e.key}: expected a finite numeric value for ${r ?? i}`);
		let t = i ?? (r === "bytes" ? "bytes" : r === "duration" ? "ms" : void 0), a = t === "bytes" && n < 0 ? `-${Mo(Math.abs(n), t)}` : Mo(n, t);
		return /* @__PURE__ */ (0, J.jsx)("span", {
			className: "font-mono",
			title: String(n),
			children: a
		});
	}
	if (r === "date") {
		if (typeof n != "number" && typeof n != "string" && !(n instanceof Date)) throw Error(`Property ${e.key}: expected a date or timestamp`);
		return /* @__PURE__ */ (0, J.jsx)("time", {
			dateTime: new Date(n).toISOString(),
			title: String(n),
			children: new Date(n).toISOString()
		});
	}
	if (r === "boolean" && typeof n != "boolean") throw Error(`Property ${e.key}: expected a boolean`);
	return typeof n == "boolean" ? /* @__PURE__ */ (0, J.jsx)("input", {
		type: "checkbox",
		checked: n,
		disabled: !0,
		"aria-label": `${e.key} value`,
		title: String(n),
		className: "h-4 w-4 accent-primary"
	}) : /* @__PURE__ */ (0, J.jsx)("span", {
		className: "whitespace-pre-wrap break-words font-mono",
		children: String(n)
	});
}
// ../../../clicky-ui/packages/ui/src/data/Properties.tsx
var jv = "clicky-ui-properties", Mv = [
	{
		value: "compact",
		label: "Compact",
		icon: Et
	},
	{
		value: "comfortable",
		label: "Comfortable",
		icon: wt
	},
	{
		value: "spacious",
		label: "Spacious",
		icon: it
	}
], Nv = {
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
function Pv({ items: e, renderLabel: t, renderValue: n, labelIcon: r, prefixActions: i, suffixActions: a, className: o, rowClassName: s, gridTemplateColumns: l, labelClassName: u, valueClassName: d, density: f = "comfortable", keyRef: p, showDensityMenu: m = !0, emptyMessage: h = "No properties" }) {
	let g = (0, Y.useContext)(zp), [_, v] = (0, Y.useState)();
	(0, Y.useEffect)(() => v(Lv(p)), [p]);
	let y = e.filter((e) => !e.hidden);
	if (y.length === 0) return /* @__PURE__ */ (0, J.jsx)("div", {
		className: "text-sm text-muted-foreground",
		children: h
	});
	let b = g?.density ?? _ ?? f, x = Nv[b], S = g?.columns ?? l ?? x.gridTemplateColumns, C = g?.depth ?? 0;
	return /* @__PURE__ */ (0, J.jsxs)("dl", {
		className: q("grid gap-x-density-3 divide-y divide-border rounded-md border border-border bg-muted/20", g && "col-span-full rounded-none border-0 bg-transparent", o),
		style: { gridTemplateColumns: g ? "subgrid" : S },
		children: [m && !g ? /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "col-span-full flex items-center justify-end px-density-2 py-density-1",
			children: [/* @__PURE__ */ (0, J.jsx)("dt", {
				className: "sr-only",
				children: "Properties display options"
			}), /* @__PURE__ */ (0, J.jsx)("dd", { children: /* @__PURE__ */ (0, J.jsx)(Fv, {
				density: b,
				onChange: (e) => {
					v(e), Rv(p, e);
				}
			}) })]
		}) : null, y.map((e) => {
			let o = typeof r == "function" ? r(e.key, e.value, e) : r, l = t ? t(e.key, e.value, e) : kv(e.key), f = typeof l == "string" ? l : e.key;
			return /* @__PURE__ */ (0, J.jsxs)(Y.Fragment, { children: [/* @__PURE__ */ (0, J.jsxs)("div", {
				className: q("col-span-full grid min-w-0 grid-cols-subgrid gap-density-3", x.row, s),
				children: [/* @__PURE__ */ (0, J.jsxs)("dt", {
					"aria-label": f,
					className: q("min-w-0 truncate", x.label, u),
					style: { paddingInlineStart: C * 16 },
					children: [/* @__PURE__ */ (0, J.jsxs)("div", {
						className: "flex min-w-0 items-center gap-density-1",
						children: [o ? /* @__PURE__ */ (0, J.jsx)(c, {
							name: o,
							className: "text-xs shrink-0"
						}) : null, /* @__PURE__ */ (0, J.jsx)("span", {
							className: "min-w-0 truncate",
							children: l
						})]
					}), e.subtitle ? /* @__PURE__ */ (0, J.jsx)("div", {
						className: "mt-0.5 text-[10px] font-normal text-muted-foreground/80",
						children: e.subtitle
					}) : null]
				}), /* @__PURE__ */ (0, J.jsx)("dd", {
					className: q("min-w-0 space-y-density-1", d),
					children: /* @__PURE__ */ (0, J.jsxs)("div", {
						className: "flex min-w-0 items-start gap-density-1",
						children: [
							/* @__PURE__ */ (0, J.jsx)(Vv, {
								actions: i,
								item: e,
								className: "shrink-0 pt-0.5"
							}),
							/* @__PURE__ */ (0, J.jsx)("div", {
								className: q("min-w-0 max-w-full flex-1", x.value),
								children: n ? n(e.key, e.value, e) : /* @__PURE__ */ (0, J.jsx)(Av, {
									item: e,
									density: b
								})
							}),
							/* @__PURE__ */ (0, J.jsx)(Vv, {
								actions: a,
								item: e,
								className: "shrink-0 pt-0.5"
							})
						]
					})
				})]
			}), e.expandable && e.expanded && e.renderChildren ? /* @__PURE__ */ (0, J.jsxs)("div", {
				className: "col-span-full grid grid-cols-subgrid",
				children: [/* @__PURE__ */ (0, J.jsxs)("dt", {
					className: "sr-only",
					children: [f, " properties"]
				}), /* @__PURE__ */ (0, J.jsx)("dd", {
					className: "col-span-full grid min-w-0 grid-cols-subgrid",
					children: /* @__PURE__ */ (0, J.jsx)(zp.Provider, {
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
function Fv({ density: e, onChange: t }) {
	return /* @__PURE__ */ (0, J.jsx)(uo, {
		align: "right",
		menuLabel: "Properties options",
		trigger: /* @__PURE__ */ (0, J.jsx)(On, {
			icon: ct,
			label: "Properties options",
			className: "h-6 w-6"
		}),
		children: (n) => /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "min-w-44 py-1 text-popover-foreground",
			children: [/* @__PURE__ */ (0, J.jsx)("div", {
				className: "px-2 py-1 text-xs font-medium text-muted-foreground",
				children: "Density"
			}), Mv.map((r) => {
				let i = e === r.value;
				return /* @__PURE__ */ (0, J.jsxs)("button", {
					type: "button",
					role: "menuitemradio",
					"aria-checked": i,
					className: Iv(i),
					onClick: () => {
						t(r.value), n();
					},
					children: [
						/* @__PURE__ */ (0, J.jsx)(c, {
							icon: r.icon,
							className: "text-sm text-muted-foreground"
						}),
						/* @__PURE__ */ (0, J.jsx)("span", {
							className: "min-w-0 flex-1 truncate",
							children: r.label
						}),
						i ? /* @__PURE__ */ (0, J.jsx)(c, {
							icon: bt,
							className: "text-sm text-foreground"
						}) : /* @__PURE__ */ (0, J.jsx)("span", {
							className: "inline-block h-4 w-4",
							"aria-hidden": !0
						})
					]
				}, r.value);
			})]
		})
	});
}
function Iv(e) {
	return q("flex w-full items-center gap-2 px-2 py-1.5 text-left text-sm", "hover:bg-accent focus:bg-accent focus:outline-none", e && "bg-accent/60");
}
function Lv(e) {
	let t = zv(e);
	if (t && typeof window < "u") try {
		let e = window.localStorage.getItem(t);
		return Bv(e) ? e : void 0;
	} catch {
		return;
	}
}
function Rv(e, t) {
	let n = zv(e);
	if (n && typeof window < "u") try {
		window.localStorage.setItem(n, t);
	} catch {}
}
function zv(e) {
	let t = e?.trim();
	return t ? `${jv}:${t}:density` : void 0;
}
function Bv(e) {
	return e === "compact" || e === "comfortable" || e === "spacious";
}
function Vv({ actions: e, item: t, className: n }) {
	if (!e || e.length === 0) return null;
	let r = e.filter((e) => !e.visible || e.visible(t.key, t.value, t));
	return r.length === 0 ? null : /* @__PURE__ */ (0, J.jsx)("span", {
		className: q("inline-flex items-center", n),
		children: r.map((e) => /* @__PURE__ */ (0, J.jsx)(Hv, {
			action: e,
			item: t
		}, e.id))
	});
}
function Hv({ action: e, item: t }) {
	let n = e.label(t.key, t.value, t), r = e.disabled?.(t.key, t.value, t) ?? !1;
	return /* @__PURE__ */ (0, J.jsx)("button", {
		type: "button",
		"aria-label": n,
		title: n,
		disabled: r,
		className: q("inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground", "hover:bg-accent hover:text-foreground", "disabled:opacity-35 disabled:hover:bg-transparent disabled:hover:text-muted-foreground"),
		onClick: (n) => {
			n.stopPropagation(), e.onClick(t.key, t.value, t);
		},
		children: /* @__PURE__ */ (0, J.jsx)(c, {
			...typeof e.icon == "string" ? { name: e.icon } : { icon: e.icon },
			className: "text-xs"
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-properties-size.ts
var Uv = {
	xs: "xs",
	sm: "xs",
	md: "sm",
	lg: "sm",
	xl: "md"
}, Wv = {
	xs: "px-1.5 py-0",
	sm: "px-2 py-0.5",
	md: "px-2 py-1",
	lg: "px-2.5 py-1.5",
	xl: "px-3 py-2"
}, Gv = (0, Y.createContext)(null);
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-property-draft.ts
function Kv(e, t) {
	if (Object.is(e, t)) return !0;
	if (Array.isArray(e) || Array.isArray(t)) return Array.isArray(e) && Array.isArray(t) && e.length === t.length && e.every((e, n) => Kv(e, t[n]));
	if (e === null || t === null || typeof e != "object" || typeof t != "object") return !1;
	let n = e, r = t;
	return Object.keys(n).length === Object.keys(r).length && Object.keys(n).every((e) => Object.prototype.hasOwnProperty.call(r, e) && Kv(n[e], r[e]));
}
function qv(e, t) {
	return t.slice(1).split("/").reduce((e, t) => {
		if (typeof e == "object" && e) return e[t.replace(/~1/g, "/").replace(/~0/g, "~")];
	}, e);
}
function Jv({ root: e, path: t, value: n }) {
	let r = t.slice(1).split("/").map((e) => e.replace(/~1/g, "/").replace(/~0/g, "~"));
	function i(e, a) {
		let o = r[a];
		if (o === void 0) return n;
		if (Array.isArray(e)) return e.map((e, t) => String(t) === o ? i(e, a + 1) : e);
		if (e != null && typeof e != "object") throw Error(`Cannot edit property at ${t}: expected an object`);
		let s = e;
		return {
			...s,
			[o]: i(s?.[o], a + 1)
		};
	}
	return i(e, 0);
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-properties.tsx
function Yv({ field: e, fieldId: t, ctx: n, renderEditor: r, renderPreview: i, actionsPlacement: a = "inline" }) {
	let o = (0, Y.useContext)(Gv), [s, c] = (0, Y.useState)(!1), [l, u] = (0, Y.useState)(e.value), [d, f] = (0, Y.useState)(), p = (0, Y.useRef)(null), m = (0, Y.useRef)(null), h = (0, Y.useRef)(!1), g = (0, Y.useRef)(!1), _ = n.readOnly || e.readOnly, v = d ? !Kv(d, n.rootValue) : !Kv(l, e.value), y = !s && v && !_ && !n.autoSave, b = (0, Y.useRef)({
		draft: l,
		rootDraft: d,
		dirty: v
	});
	b.current = {
		draft: l,
		rootDraft: d,
		dirty: v
	}, (0, Y.useEffect)(() => {
		!s && h.current && (m.current?.focus(), h.current = !1);
		let e = p.current;
		if (!s || !e) return;
		let t, n = () => {
			t = e.querySelector("[contenteditable=\"true\"]") ?? e.querySelector("input:not(:disabled):not([type=hidden]), textarea:not(:disabled), select:not(:disabled), [role=combobox]") ?? e.querySelector("button:not(:disabled)"), t?.focus();
		};
		n();
		let r = new MutationObserver(() => {
			t && !t.isConnected && document.activeElement === document.body && n();
		});
		return r.observe(e, {
			childList: !0,
			subtree: !0
		}), () => r.disconnect();
	}, [s]);
	let x = ({ restoreFocus: t }) => {
		u(e.value), f(void 0), h.current = t, c(!1);
	}, S = () => {
		h.current = !1, c(!1);
	}, C = ({ restoreFocus: t }) => {
		let { draft: r, rootDraft: i } = b.current;
		i ? n.onRootChange?.(i) : e.onChange(r), h.current = t, c(!1);
	}, w = () => {
		y || (u(e.value), f(void 0)), c(!0);
	}, E = () => {
		let e = p.current, t = e?.closest("[data-json-schema-form]");
		if (!e || !t) throw Error("Property editor must belong to a JsonSchemaForm");
		let n = () => Array.from(t.querySelectorAll("[data-property-editor]")).filter((e) => e.closest("[data-json-schema-form]") === t), r = n().indexOf(e) + 1, i = n()[r]?.dataset.propertyEditor;
		(0, Fe.flushSync)(() => C({ restoreFocus: !0 }));
		let a = n(), o = a.find((e) => e.dataset.propertyEditor === i) ?? a[r];
		o?.getAttribute("role") === "button" ? o.focus() : o?.querySelector("input:not(:disabled):not([type=hidden]), textarea:not(:disabled), select:not(:disabled), [contenteditable=\"true\"], [role=combobox]")?.focus();
	}, D = (e) => {
		e.defaultPrevented || e.nativeEvent.isComposing || e.keyCode === 229 || (e.key === "Escape" ? (e.preventDefault(), e.stopPropagation(), x({ restoreFocus: !0 })) : e.key !== "Enter" || e.shiftKey || e.altKey || e.ctrlKey || e.metaKey || e.target instanceof Element && e.target.closest("button, [role=button], a") || (e.preventDefault(), e.stopPropagation(), E()));
	}, O = v && !_ && !n.autoSave ? /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)(On, {
		icon: bt,
		label: `Save ${e.label}`,
		className: q("w-6 text-[var(--clicky-success)] hover:text-[var(--clicky-success-700)]", T[n.size]),
		onClick: () => C({ restoreFocus: !0 })
	}), /* @__PURE__ */ (0, J.jsx)(On, {
		icon: j,
		label: `Cancel editing ${e.label}`,
		className: q("w-6 text-destructive hover:text-destructive", T[n.size]),
		onClick: () => x({ restoreFocus: !0 })
	})] }) : null, k = a === "row" ? o && (0, Fe.createPortal)(O, o) : O;
	if (s) return /* @__PURE__ */ (0, J.jsxs)("div", {
		ref: p,
		"data-property-editor": _ ? void 0 : t,
		className: "flex min-w-0 items-start gap-1",
		onFocusCapture: () => {
			g.current = !0;
		},
		onBlurCapture: () => {
			g.current = !1, setTimeout(() => {
				!g.current && p.current?.isConnected && (n.autoSave && b.current.dirty ? C({ restoreFocus: !1 }) : S());
			}, 0);
		},
		onKeyDown: D,
		onKeyDownCapture: (e) => {
			e.key === "Enter" && e.target instanceof Element && e.target.closest("[contenteditable=\"true\"]") && D(e);
		},
		children: [/* @__PURE__ */ (0, J.jsx)("div", {
			className: "min-w-0 flex-1",
			children: r({
				...e,
				value: l,
				onChange: (e) => {
					u(e), d && f(Jv({
						root: d,
						path: n.instancePath,
						value: e
					}));
				}
			}, {
				...n,
				...n.rootValue ? { rootValue: d ?? Jv({
					root: n.rootValue,
					path: n.instancePath,
					value: l
				}) } : {},
				...n.onRootChange ? { onRootChange: (e) => {
					f(e), u(qv(e, n.instancePath));
				} } : {}
			})
		}), k]
	});
	let A = /* @__PURE__ */ (0, J.jsx)("div", {
		ref: m,
		id: t,
		"data-property-editor": _ ? void 0 : t,
		role: _ ? void 0 : "button",
		tabIndex: _ ? void 0 : 0,
		"aria-label": _ ? e.label : `Edit ${e.label}`,
		onFocus: (e) => {
			e.target === e.currentTarget && !_ && !h.current && w();
		},
		...B(e, t),
		className: q("block w-full min-w-0 rounded text-left", !_ && "cursor-pointer hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"),
		onClickCapture: (e) => {
			e.preventDefault(), e.stopPropagation(), _ || w();
		},
		onKeyDownCapture: (e) => {
			(e.key === "Enter" || e.key === " ") && (e.preventDefault(), e.stopPropagation(), _ || w());
		},
		onChangeCapture: (e) => {
			e.stopPropagation(), _ || w();
		},
		children: /* @__PURE__ */ (0, J.jsx)("div", {
			className: "pointer-events-none [&_*]:pointer-events-none",
			children: i(y ? {
				...e,
				value: l
			} : e)
		})
	});
	return y ? /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex min-w-0 items-start gap-1",
		children: [/* @__PURE__ */ (0, J.jsx)("div", {
			className: "min-w-0 flex-1",
			children: A
		}), k]
	}) : A;
}
function Xv({ field: e, fieldId: t, label: n, value: r, messages: i, ctx: a }) {
	let [o, s] = (0, Y.useState)(null), c = e.kind === "object" || e.kind === "string-map";
	return /* @__PURE__ */ (0, J.jsx)(Gv.Provider, {
		value: o,
		children: /* @__PURE__ */ (0, J.jsx)(Pv, {
			showDensityMenu: !1,
			density: "comfortable",
			className: "rounded-none border-0 border-b",
			rowClassName: q("items-center", Wv[a.size]),
			gridTemplateColumns: `minmax(0, ${a.layout.labelMaxWidth ?? "12rem"}) minmax(0, ${a.layout.valueMaxWidth ?? "1fr"})`,
			items: [{
				key: e.key,
				value: r,
				...e.helper && (e.helpDisplay ?? a.layout.help) !== "hover" ? { subtitle: e.helper } : {},
				...c ? {
					expandable: !0,
					expanded: !0,
					renderChildren: () => r
				} : {}
			}],
			renderLabel: () => n,
			renderValue: () => /* @__PURE__ */ (0, J.jsxs)("div", {
				className: "flex min-w-0 items-start gap-1",
				children: [/* @__PURE__ */ (0, J.jsxs)("div", {
					className: "min-w-0 flex-1",
					children: [!c && r, /* @__PURE__ */ (0, J.jsx)(fm, {
						id: re(t),
						errors: i
					})]
				}), /* @__PURE__ */ (0, J.jsx)("span", {
					ref: s,
					"data-property-row-actions": !0,
					className: "ml-auto inline-flex shrink-0 gap-1 empty:hidden"
				})]
			})
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-object.tsx
function Zv({ field: e, ctx: t }) {
	let n = w(e.value) ? e.value : {}, r = {
		type: "object",
		properties: e.objectProperties ?? {},
		...e.objectRequired ? { required: e.objectRequired } : {},
		...Array.isArray(e.schema.allOf) ? { allOf: e.schema.allOf } : {},
		...e.schema.oneOf && Dv({
			type: "object",
			properties: e.objectProperties ?? {},
			oneOf: e.schema.oneOf
		}) ? { oneOf: e.schema.oneOf } : {},
		...e.schema["x-on-load"] === void 0 ? {} : { "x-on-load": e.schema["x-on-load"] },
		...Array.isArray(e.schema["x-order"]) ? { "x-order": e.schema["x-order"] } : {},
		...typeof e.schema["x-columns"] == "number" || e.schema["x-columns"] === "auto" ? { "x-columns": e.schema["x-columns"] } : {},
		...typeof e.schema["x-column-min-width"] == "string" ? { "x-column-min-width": e.schema["x-column-min-width"] } : {},
		...typeof e.schema["x-columns-max-width"] == "string" ? { "x-columns-max-width": e.schema["x-columns-max-width"] } : {},
		...typeof e.schema["x-classes"] == "string" ? { "x-classes": e.schema["x-classes"] } : {},
		...e.schema["x-object-display"] === "row" ? { "x-object-display": "row" } : {}
	}, i = {
		...t,
		readOnly: t.readOnly || e.readOnly === !0,
		depth: t.depth + 1
	}, a = (t) => e.onChange(t);
	return r["x-object-display"] === "row" && t.layout.mode !== "properties" ? /* @__PURE__ */ (0, J.jsx)("div", {
		"data-jsf-object-row": "",
		className: q("flex w-full min-w-0 flex-nowrap items-start", d[t.size], r["x-classes"]),
		children: t.render.renderObjectFields(r, n, a, i)
	}) : /* @__PURE__ */ (0, J.jsx)(am, {
		layout: t.layout,
		size: t.size,
		columns: I(r["x-columns"]),
		columnMinWidth: G(r["x-column-min-width"], ne),
		...typeof r["x-columns-max-width"] == "string" ? { columnsMaxWidth: r["x-columns-max-width"] } : {},
		...typeof r["x-classes"] == "string" ? { className: r["x-classes"] } : {},
		children: t.render.renderObjectFields(r, n, a, i)
	});
}
function Qv({ field: e, ctx: t }) {
	if (t.layout.mode !== "properties" && e.valueSchema?.type === "string" && !e.valuePatternSchemas?.length) {
		let n = {
			...t.layout,
			mode: "properties"
		};
		return /* @__PURE__ */ (0, J.jsx)(am, {
			layout: n,
			size: t.size,
			children: /* @__PURE__ */ (0, J.jsx)($v, {
				field: e,
				ctx: {
					...t,
					layout: n
				}
			})
		});
	}
	return /* @__PURE__ */ (0, J.jsx)($v, {
		field: e,
		ctx: t
	});
}
function $v({ field: e, ctx: t }) {
	let n = t.readOnly || e.readOnly === !0, r = w(e.value) ? e.value : {}, i = e.knownProperties ?? {}, a = Object.keys(i), o = Object.keys(r).filter((e) => !(e in i)), s = {
		...t,
		readOnly: n,
		depth: t.depth + 1
	}, l = t.layout.mode === "properties", u = e.schema.propertyNames, f = typeof u?.title == "string" && u.title ? u.title : void 0;
	function p(t) {
		for (let { pattern: n, schema: r } of e.valuePatternSchemas ?? []) {
			let e;
			try {
				e = new RegExp(n);
			} catch {
				e = void 0;
			}
			if (e?.test(t)) return r;
		}
		return e.valueSchema ?? { type: "string" };
	}
	function m(e) {
		return p(e)["x-layout"] === "stack";
	}
	function h(t, n) {
		e.onChange({
			...r,
			[t]: n
		});
	}
	function _(t, n) {
		if (n === t) return;
		let i = {};
		for (let [e, a] of Object.entries(r)) i[e === t ? n : e] = a;
		e.onChange(i);
	}
	function v(t) {
		let n = { ...r };
		delete n[t], e.onChange(n);
	}
	function y() {
		"" in r || e.onChange({
			...r,
			"": ""
		});
	}
	function b(e, n) {
		return t.render.renderFieldNodes({
			key: e,
			prop: n,
			required: !1,
			value: r[e],
			onChange: (t) => h(e, t),
			instancePath: Xh(t.instancePath, e)
		}, s)?.value ?? null;
	}
	function x(e) {
		return b(e, p(e));
	}
	function S({ key: r, value: i, onChange: s }) {
		let c = l ? Uv[t.size] : t.size;
		return e.keyOptions ? /* @__PURE__ */ (0, J.jsx)(ss, {
			options: se(e.keyOptions, [...a, ...o], r),
			value: i,
			disabled: n,
			size: c,
			allowCustomValue: !1,
			onChange: s,
			placeholder: "Select…"
		}) : /* @__PURE__ */ (0, J.jsx)("input", {
			type: "text",
			"aria-label": "Field name",
			className: q(V(c), "font-mono"),
			value: i,
			disabled: n,
			onChange: (e) => s(e.target.value)
		});
	}
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: q(l ? "col-span-full grid grid-cols-subgrid" : "flex flex-col", !l && !t.presentation && "rounded-md border border-input p-2", !l && d[s.size]),
		children: [
			a.map((n) => /* @__PURE__ */ (0, J.jsx)("div", {
				className: "contents",
				children: t.render.renderFieldRow({
					key: n,
					prop: i[n],
					required: Array.isArray(e.schema.required) && e.schema.required.includes(n),
					value: r[n],
					onChange: (e) => h(n, e),
					instancePath: Xh(t.instancePath, n)
				}, s)
			}, `known-${n}`)),
			o.map((i) => {
				let a = t.presentation ? /* @__PURE__ */ (0, J.jsx)("span", {
					className: q("flex min-w-0 items-center break-words text-xs", T[s.size], e.keyOptions ? "text-foreground" : "font-mono text-muted-foreground"),
					children: e.keyOptions?.find((e) => e.value === i)?.label ?? i
				}) : S({
					key: i,
					value: i,
					onChange: (e) => _(i, e)
				}), o = n ? null : /* @__PURE__ */ (0, J.jsx)("button", {
					type: "button",
					"aria-label": `Remove ${i}`,
					className: q("inline-flex aspect-square shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-foreground", T[s.size]),
					onClick: () => v(i),
					children: /* @__PURE__ */ (0, J.jsx)(c, {
						icon: Ht,
						className: "text-sm"
					})
				});
				if (l) {
					let n = Xh(t.instancePath, i), a = e.keyOptions?.find((e) => e.value === i)?.label ?? (i || "New field");
					return /* @__PURE__ */ (0, J.jsx)("div", {
						className: "contents",
						children: t.render.renderFieldRow({
							key: i,
							prop: {
								...p(i),
								title: a
							},
							required: !1,
							value: r[i],
							onChange: (e) => h(i, e),
							instancePath: n
						}, {
							...s,
							post: [...s.post, (e, r, c) => c?.instancePath === n ? {
								...r,
								label: /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)("span", {
									className: "sr-only",
									children: r.label
								}), /* @__PURE__ */ (0, J.jsx)(Yv, {
									actionsPlacement: "row",
									field: {
										...e,
										kind: "string",
										label: `${a} key`,
										value: i,
										onChange: (e) => _(i, String(e))
									},
									fieldId: `${W(n, t.idPrefix)}-key`,
									ctx: {
										...s,
										size: Uv[t.size]
									},
									renderPreview: (e) => /* @__PURE__ */ (0, J.jsx)("span", { children: e.value === "" ? "New field" : String(e.value) }),
									renderEditor: (e) => /* @__PURE__ */ (0, J.jsxs)("div", {
										className: "flex min-w-0 items-center gap-1",
										children: [S({
											key: i,
											value: String(e.value),
											onChange: e.onChange
										}), o]
									})
								})] })
							} : r]
						})
					}, `extra-${i}`);
				}
				return m(i) ? /* @__PURE__ */ (0, J.jsxs)("div", {
					className: q("space-y-1.5", !t.presentation && "rounded-md border border-input p-2"),
					children: [
						f && /* @__PURE__ */ (0, J.jsx)("span", {
							className: "block text-xs font-medium text-muted-foreground",
							children: f
						}),
						/* @__PURE__ */ (0, J.jsxs)("div", {
							className: "flex items-center gap-2",
							children: [/* @__PURE__ */ (0, J.jsx)("div", {
								className: "min-w-0 flex-1",
								children: a
							}), o]
						}),
						/* @__PURE__ */ (0, J.jsx)("div", {
							className: "min-w-0",
							children: x(i)
						})
					]
				}, `extra-${i}`) : /* @__PURE__ */ (0, J.jsxs)("div", {
					className: q("grid items-center gap-2", t.presentation ? "grid-cols-[minmax(0,10rem)_minmax(0,1fr)]" : "grid-cols-[10rem_1fr_auto]"),
					children: [
						a,
						/* @__PURE__ */ (0, J.jsx)("div", {
							className: "min-w-0",
							children: x(i)
						}),
						o
					]
				}, `extra-${i}`);
			}),
			!n && e.allowExtraKeys !== !1 && /* @__PURE__ */ (0, J.jsxs)(fe, {
				type: "button",
				variant: l ? "ghost" : "outline",
				onClick: y,
				className: q("self-start gap-1.5", l && "col-span-full justify-self-start m-1", g[l ? Uv[t.size] : s.size]),
				children: [/* @__PURE__ */ (0, J.jsx)(c, {
					icon: X,
					className: "text-sm"
				}), "Add field"]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-custom-control.tsx
function ey({ field: e, fieldId: t, readOnly: n, ctx: r }) {
	let i = e.format === void 0 ? void 0 : r.extensions.formats.get(e.format);
	if (!i) throw Error(`field "${e.key}" resolved to a custom control but no form extension registers format "${e.format}"`);
	let a = i.render;
	return /* @__PURE__ */ (0, J.jsx)(a, {
		field: e,
		value: e.value,
		onChange: e.onChange,
		fieldId: t,
		readOnly: n,
		ctx: {
			fieldId: t,
			instancePath: r.instancePath,
			...r.rootValue ? { rootValue: r.rootValue } : {},
			...r.onRootChange ? { onRootChange: r.onRootChange } : {}
		}
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-hoist.ts
function ty(e) {
	return e["x-clicky-hoist"] === !0;
}
function ny(e) {
	let { properties: t, value: n, onChange: r, schema: i, listenerOptions: a, instancePath: o } = e, s = /* @__PURE__ */ new Map(), c = /* @__PURE__ */ new Set();
	for (let [e, l] of Object.entries(t)) {
		if (!nh(l)) continue;
		let t = n[e];
		if (t !== void 0 && !w(t)) throw Error(`cannot hoist fields out of ${e}: its value is not an object`);
		let u = t ?? {}, d = l, f = Oh(d, u, a), p = Object.entries(f.properties), m = p.filter(([, e]) => ty(e));
		if (m.length !== 0) {
			m.length === p.length && c.add(e);
			for (let [t, c] of m) s.set(`${e}.${t}`, {
				key: t,
				prop: c,
				required: f.required.includes(t),
				value: u[t],
				instancePath: Xh(Xh(o, e), t),
				onChange: (o) => {
					let s = Nh(d, {
						...a,
						value: u,
						key: t,
						next: o
					});
					r(Nh(i, {
						...a,
						value: n,
						key: e,
						next: s
					}));
				}
			});
		}
	}
	return {
		rows: s,
		emptied: c
	};
}
// ../../../clicky-ui/packages/ui/src/overlay/confirmMemory.ts
var ry = "clicky-ui:confirm:", iy = "confirmed", ay = /* @__PURE__ */ new Set();
function oy() {
	if (typeof window > "u") return null;
	try {
		let e = window.localStorage;
		return e.getItem(ry), e;
	} catch {
		return null;
	}
}
function sy() {
	return oy() !== null;
}
function cy(e) {
	let t = oy();
	if (!t) throw Error(`cannot remember confirmation "${e}": localStorage is unavailable`);
	t.setItem(ry + e, iy), ly();
}
function ly() {
	for (let e of ay) e();
}
// ../../../clicky-ui/packages/ui/src/overlay/ConfirmDialog.tsx
function uy({ open: e, title: t, message: n, warning: r, confirmLabel: i = "Confirm", cancelLabel: a = "Cancel", variant: o = "default", remember: s, onConfirm: c, onCancel: l }) {
	let [u, d] = (0, Y.useState)(!1);
	(0, Y.useEffect)(() => {
		e && d(!1);
	}, [e]);
	let f = s !== void 0 && sy();
	return /* @__PURE__ */ (0, J.jsx)(Xd, {
		open: e,
		onClose: l,
		title: t,
		size: "sm",
		expandable: !1,
		footer: /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "flex justify-end gap-density-2",
			children: [/* @__PURE__ */ (0, J.jsx)(fe, {
				type: "button",
				variant: "ghost",
				size: "sm",
				onClick: l,
				children: a
			}), /* @__PURE__ */ (0, J.jsx)(fe, {
				type: "button",
				variant: o,
				size: "sm",
				onClick: () => {
					f && u && cy(s.key), c();
				},
				children: i
			})]
		}),
		children: /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "space-y-density-3",
			children: [
				r ? /* @__PURE__ */ (0, J.jsx)(ae, {
					variant: "warning",
					className: "my-0",
					children: r
				}) : null,
				n ? /* @__PURE__ */ (0, J.jsx)("p", {
					className: "text-sm text-muted-foreground",
					children: n
				}) : null,
				f ? /* @__PURE__ */ (0, J.jsxs)("label", {
					className: "flex items-center gap-2 text-sm",
					children: [/* @__PURE__ */ (0, J.jsx)("input", {
						type: "checkbox",
						className: "size-4 shrink-0 rounded border-input",
						checked: u,
						onChange: (e) => d(e.target.checked)
					}), s.label ?? "Remember my choice"]
				}) : null
			]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/rpc/router.tsx
var dy = (0, Y.createContext)(null);
function fy(e) {
	return !e.defaultPrevented && e.button === 0 && !e.metaKey && !e.ctrlKey && !e.shiftKey && !e.altKey && e.currentTarget.target !== "_blank";
}
function py(e) {
	return window.addEventListener("popstate", e), () => window.removeEventListener("popstate", e);
}
function my() {
	let e = (0, Y.useSyncExternalStore)(py, () => `${window.location.pathname}${window.location.search}${window.location.hash}`, () => "/").split(/[?#]/, 1)[0] || "/", t = (0, Y.useCallback)((e, t) => {
		typeof window < "u" && (t?.replace ? window.history.replaceState(window.history.state, "", e) : window.history.pushState(window.history.state, "", e), window.dispatchEvent(new PopStateEvent("popstate")));
	}, []), n = (0, Y.useCallback)(({ to: e, className: n, children: r, title: i, key: a, draggable: o, onDragStart: s }) => /* @__PURE__ */ (0, J.jsx)("a", {
		href: e,
		className: n,
		title: i,
		draggable: o,
		onDragStart: s,
		onClick: (n) => {
			fy(n) && (n.preventDefault(), t(e));
		},
		children: r
	}, a), [t]);
	return (0, Y.useMemo)(() => ({
		pathname: e,
		renderLink: n,
		navigate: t
	}), [
		e,
		n,
		t
	]);
}
function hy() {
	let e = (0, Y.useContext)(dy), t = my();
	return e ?? t;
}
// ../../../clicky-ui/packages/ui/src/data/code-block-actions.tsx
var gy = {
	javascript: "js",
	typescript: "ts",
	tsx: "tsx",
	jsx: "jsx",
	python: "py",
	bash: "sh",
	shell: "sh",
	yaml: "yaml",
	json: "json",
	sql: "sql",
	java: "java",
	go: "go",
	html: "html",
	xml: "xml",
	css: "css"
};
function _y(e, t, n) {
	let r = t && gy[t.toLowerCase()] || "txt", i = new Blob([e], { type: "text/plain;charset=utf-8" }), a = URL.createObjectURL(i), o = document.createElement("a");
	o.href = a, o.download = n || `snippet.${r}`, o.click(), URL.revokeObjectURL(a);
}
function vy({ source: e, language: t, copyable: n = !1, downloadable: r = !1, downloadFilename: i, theme: a, onToggleTheme: o }) {
	let [s, l] = (0, Y.useState)(!1);
	return !n && !r && !o ? null : /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex items-center gap-1 text-muted-foreground",
		children: [
			o && /* @__PURE__ */ (0, J.jsx)("button", {
				type: "button",
				"aria-label": a === "dark" ? "Switch to light theme" : "Switch to dark theme",
				title: "Toggle theme",
				onClick: o,
				className: "rounded p-1 hover:bg-accent hover:text-accent-foreground",
				children: /* @__PURE__ */ (0, J.jsx)(c, {
					icon: a === "dark" ? rt : nt,
					className: "size-3.5"
				})
			}),
			r && /* @__PURE__ */ (0, J.jsx)("button", {
				type: "button",
				"aria-label": "Download code",
				title: "Download file",
				onClick: () => _y(e, t, i),
				className: "rounded p-1 hover:bg-accent hover:text-accent-foreground",
				children: /* @__PURE__ */ (0, J.jsx)(c, {
					icon: Bt,
					className: "size-3.5"
				})
			}),
			n && /* @__PURE__ */ (0, J.jsx)("button", {
				type: "button",
				"aria-label": s ? "Copied" : "Copy code",
				title: "Copy code",
				onClick: async () => {
					try {
						await navigator.clipboard.writeText(e), l(!0), setTimeout(() => l(!1), 1500);
					} catch (e) {
						console.warn("clicky-ui: copy to clipboard failed", e);
					}
				},
				className: "rounded p-1 hover:bg-accent hover:text-accent-foreground",
				children: /* @__PURE__ */ (0, J.jsx)(c, {
					icon: s ? bt : Rt,
					className: q("size-3.5", s && "text-emerald-600")
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/code-diff.ts
function yy(e) {
	return e === "" ? [] : (e.endsWith("\n") ? e.slice(0, -1) : e).split("\n");
}
function by(e, t) {
	let n = xy(yy(e), yy(t));
	return n.length > 0 ? [{ lines: n }] : [];
}
function xy(e, t) {
	let n = e.length, r = t.length, i = r + 1, a = new Int32Array((n + 1) * i), o = (e, t) => a[e * i + t];
	for (let s = n - 1; s >= 0; s--) for (let n = r - 1; n >= 0; n--) a[s * i + n] = e[s] === t[n] ? o(s + 1, n + 1) + 1 : Math.max(o(s + 1, n), o(s, n + 1));
	let s = [], c = 0, l = 0, u = 1, d = 1;
	for (; c < n && l < r;) e[c] === t[l] ? (s.push({
		type: "context",
		content: e[c],
		oldNumber: u++,
		newNumber: d++
	}), c++, l++) : o(c + 1, l) >= o(c, l + 1) ? (s.push({
		type: "remove",
		content: e[c],
		oldNumber: u++
	}), c++) : (s.push({
		type: "add",
		content: t[l],
		newNumber: d++
	}), l++);
	for (; c < n;) s.push({
		type: "remove",
		content: e[c++],
		oldNumber: u++
	});
	for (; l < r;) s.push({
		type: "add",
		content: t[l++],
		newNumber: d++
	});
	return s;
}
var Sy = /^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@/;
function Cy(e) {
	let t = e.split("	")[0].trim();
	if (t !== "" && t !== "/dev/null") return t.startsWith("a/") || t.startsWith("b/") ? t.slice(2) : t;
}
function wy(e) {
	let t = [], n = null, r = 0, i = 0, a = 0, o = 0, s, c, l = e.endsWith("\n") ? e.slice(0, -1) : e;
	for (let e of l.split("\n")) if (n) {
		if (e.startsWith("\\")) continue;
		if (e.startsWith("+")) n.lines.push({
			type: "add",
			content: e.slice(1),
			newNumber: i++
		}), o--;
		else if (e.startsWith("-")) n.lines.push({
			type: "remove",
			content: e.slice(1),
			oldNumber: r++
		}), a--;
		else {
			let t = e.startsWith(" ") ? e.slice(1) : e;
			n.lines.push({
				type: "context",
				content: t,
				oldNumber: r++,
				newNumber: i++
			}), a--, o--;
		}
		a <= 0 && o <= 0 && (n = null);
	} else if (e.startsWith("@@")) {
		let c = Sy.exec(e);
		if (!c) throw Error(`Malformed unified diff hunk header: ${e}`);
		r = Number(c[1]), i = Number(c[3]), a = c[2] ? Number(c[2]) : 1, o = c[4] ? Number(c[4]) : 1, n = {
			header: e,
			lines: [],
			...s ? { path: s } : {}
		}, t.push(n), a <= 0 && o <= 0 && (n = null);
	} else e.startsWith("diff --git ") ? (s = void 0, c = void 0) : e.startsWith("--- ") ? c = Cy(e.slice(4)) : e.startsWith("+++ ") && (s = Cy(e.slice(4)) ?? c);
	return t;
}
// ../../../clicky-ui/packages/ui/src/data/code-diff-selection.ts
function Ty(e) {
	if (e.type === "add" && e.newNumber !== void 0) return {
		side: "new",
		line: e.newNumber
	};
	if (e.type === "remove" && e.oldNumber !== void 0) return {
		side: "old",
		line: e.oldNumber
	};
}
function Ey(e) {
	return e.flatMap((e) => Ty(e) ?? []);
}
function Dy(e, t, n) {
	let [r, i] = t <= n ? [t, n] : [n, t];
	return Ey(e.slice(r, i + 1));
}
function Oy(e, t) {
	let n = e.filter(t).length;
	return n === 0 ? "none" : n === e.length ? "all" : "partial";
}
// ../../../clicky-ui/packages/ui/src/data/code-highlight.ts
var ky = [
	"java",
	"go",
	"python",
	"javascript",
	"typescript",
	"tsx",
	"jsx",
	"json",
	"yaml",
	"bash",
	"shell",
	"sql",
	"xml",
	"html",
	"css"
], Ay = ["github-light", "github-dark"], jy = "github-light", My = null;
function Ny() {
	return My ??= (async () => {
		let [{ createHighlighterCore: e }, { createOnigurumaEngine: t }] = await Promise.all([import("./core-C2kj-dEj.js"), import("./engine-oniguruma-auXKmDQ8.js")]);
		return e({
			themes: [import("./github-light-BlLy38f4.js"), import("./github-dark-BrXOGRCP.js")],
			langs: [
				import("./java-CVshGEK5.js"),
				import("./go-ovMYcIms.js"),
				import("./python-C3MlKT6U.js"),
				import("./javascript-QPgQlnkz.js"),
				import("./typescript-CjXaXMdk.js"),
				import("./tsx-CmA3HQIl.js"),
				import("./jsx-BG4FKGoD.js"),
				import("./json-DyEoy2Fw.js"),
				import("./yaml-BZQ6jzJl.js"),
				import("./bash-BlazeVbU.js"),
				import("./sql-0Hc_5b-g.js"),
				import("./xml-DoCFmnj4.js"),
				import("./html-Cu_xiovd.js"),
				import("./css-BXHpbd9g.js")
			],
			engine: t(import("./wasm-BaiNo56X.js"))
		});
	})(), My;
}
function Py(e) {
	if (!e) return;
	let t = e.slice(e.lastIndexOf("/") + 1), n = t.lastIndexOf(".");
	if (!(n <= 0)) return t.slice(n + 1).toLowerCase();
}
function Fy(e) {
	let t = e.toLowerCase();
	return ky.includes(t) ? t : t === "shell" || t === "sh" || t === "zsh" ? "bash" : t === "ts" ? "typescript" : t === "js" ? "javascript" : t === "py" ? "python" : t === "yml" ? "yaml" : null;
}
function Iy(e) {
	return e && Ay.includes(e) ? e : jy;
}
var Ly = 8, Ry = [], zy = !1;
function By(e) {
	return new Promise((t, n) => {
		Ry.push(() => {
			try {
				t(e());
			} catch (e) {
				n(e);
			}
		}), zy || (zy = !0, setTimeout(Vy, 0));
	});
}
function Vy() {
	let e = performance.now() + Ly;
	for (; Ry.length > 0 && performance.now() < e;) Ry.shift()();
	Ry.length > 0 ? setTimeout(Vy, 0) : zy = !1;
}
var Hy = 2e6, Uy = /* @__PURE__ */ new Map(), Wy = 0;
function Gy(e, t) {
	Uy.set(e, t), Wy += t.length;
	for (let [e, t] of Uy) {
		if (Wy <= Hy) break;
		Uy.delete(e), Wy -= t.length;
	}
}
async function Ky(e, t) {
	if (!e || !t.lang) return null;
	let n = Fy(t.lang);
	if (!n) return null;
	let r = Iy(t.theme), i = t.transformers ? void 0 : `${n} ${r} ${e}`, a = i === void 0 ? void 0 : Uy.get(i);
	if (i !== void 0 && a !== void 0) return Uy.delete(i), Uy.set(i, a), a;
	try {
		let a = await Ny(), o = await By(() => a.codeToHtml(e, {
			lang: n,
			theme: r,
			...t.transformers ? { transformers: t.transformers } : {}
		}));
		return i !== void 0 && Gy(i, o), o;
	} catch {
		return null;
	}
}
async function qy(e, t) {
	if (!e || !t.lang) return null;
	let n = Fy(t.lang);
	if (!n) return null;
	try {
		let r = await Ny(), { tokens: i } = await By(() => r.codeToTokens(e, {
			lang: n,
			theme: Iy(t.theme)
		}));
		return i.map((e) => e.map((e) => ({
			content: e.content,
			...e.color ? { color: e.color } : {},
			...e.fontStyle ? { fontStyle: e.fontStyle } : {}
		})));
	} catch {
		return null;
	}
}
// ../../../clicky-ui/packages/ui/src/data/DiffSelectionBox.tsx
var Jy = "flex w-7 shrink-0 select-none items-center justify-center self-stretch [@media(pointer:coarse)]:min-h-8 [@media(pointer:coarse)]:w-8";
function Yy() {
	return /* @__PURE__ */ (0, J.jsx)("span", {
		"aria-hidden": !0,
		className: Jy
	});
}
function Xy({ state: e, label: t, disabled: n, onToggle: r }) {
	return /* @__PURE__ */ (0, J.jsx)("button", {
		type: "button",
		role: "checkbox",
		"aria-checked": e === "partial" ? "mixed" : e === "all",
		"aria-label": t,
		title: t,
		disabled: n,
		onClick: r,
		className: q(Jy, "focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring disabled:cursor-not-allowed disabled:opacity-50"),
		children: /* @__PURE__ */ (0, J.jsxs)("span", {
			"aria-hidden": !0,
			className: q("inline-flex size-3.5 shrink-0 items-center justify-center rounded-[3px] border", e === "all" && "border-primary bg-primary text-primary-foreground", e === "partial" && "border-primary bg-background text-primary", e === "none" && "border-muted-foreground/60 bg-background"),
			children: [e === "all" && /* @__PURE__ */ (0, J.jsx)(bt, { className: "size-3 text-current" }), e === "partial" && /* @__PURE__ */ (0, J.jsx)("span", { className: "h-0.5 w-2 rounded bg-current" })]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/token-style.ts
function Zy(e) {
	let t = {};
	e.color && (t.color = e.color);
	let n = e.fontStyle ?? 0;
	return n & 1 && (t.fontStyle = "italic"), n & 2 && (t.fontWeight = 700), n & 4 && (t.textDecorationLine = "underline"), t;
}
// ../../../clicky-ui/packages/ui/src/data/HighlightedTokens.tsx
function Qy({ tokens: e, content: t }) {
	return !e || e.length === 0 ? /* @__PURE__ */ (0, J.jsx)(J.Fragment, { children: t }) : /* @__PURE__ */ (0, J.jsx)(J.Fragment, { children: e.map((e, t) => /* @__PURE__ */ (0, J.jsx)("span", {
		style: Zy(e),
		children: e.content
	}, t)) });
}
// ../../../clicky-ui/packages/ui/src/data/CodeDiff.tsx
var $y = [{
	id: "unified",
	label: "Unified",
	icon: Et,
	title: "Unified view"
}, {
	id: "split",
	label: "Split",
	icon: me,
	title: "Split view"
}], eb = {
	context: void 0,
	add: "bg-emerald-500/10",
	remove: "bg-rose-500/10"
}, tb = {
	context: "text-muted-foreground",
	add: "text-emerald-700 [[data-theme=dark]_&]:text-emerald-300",
	remove: "text-rose-700 [[data-theme=dark]_&]:text-rose-300"
}, nb = {
	context: "",
	add: "+",
	remove: "-"
};
function rb(e) {
	let { language: t, view: n = "unified", showLineNumbers: r = !0, bare: i = !1, className: a } = e, { unified: o, original: s, modified: c, originalLabel: l, modifiedLabel: u, onLineAction: d, lineWidgets: f, lineSelection: p } = e, [m, h] = (0, Y.useState)(null), [g, _] = (0, Y.useState)(n), v = (0, Y.useMemo)(() => o === void 0 ? by(s ?? "", c ?? "") : wy(o), [
		o,
		s,
		c
	]), { oldSource: y, newSource: b } = (0, Y.useMemo)(() => ib(v), [v]), [x, S] = (0, Y.useState)({
		old: null,
		mod: null
	});
	(0, Y.useEffect)(() => {
		if (!t) {
			S({
				old: null,
				mod: null
			});
			return;
		}
		let e = !1;
		return Promise.all([qy(y, { lang: t }), qy(b, { lang: t })]).then(([t, n]) => {
			e || S({
				old: t,
				mod: n
			});
		}), () => {
			e = !0;
		};
	}, [
		t,
		y,
		b
	]);
	let C = (0, Y.useMemo)(() => ab(v, x.old, x.mod), [v, x]), w = {
		onLineAction: d,
		lineWidgets: f,
		selection: (0, Y.useMemo)(() => {
			if (!p) return;
			let e = (e) => {
				let t = v[e];
				if (!t) throw Error(`CodeDiff selection: no hunk ${e}`);
				return t.lines;
			};
			return {
				selection: p,
				onLineClick: (t, n, r) => {
					let i = e(t), a = i[n], o = a && Ty(a);
					if (!o) throw Error(`CodeDiff selection: line ${t}:${n} is not a changed line`);
					let s = m?.hunks === v && m.hunk === t ? m.line : void 0, c = r && s !== void 0 ? Dy(i, s, n) : [o];
					p.onChange(c, !p.isSelected(o)), h({
						hunks: v,
						hunk: t,
						line: n
					});
				},
				onHunkClick: (t) => {
					let n = Ey(e(t));
					p.onChange(n, Oy(n, p.isSelected) !== "all");
				}
			};
		}, [
			p,
			v,
			m
		])
	}, T = new Set(C.map((e) => e.path).filter((e) => e !== void 0)).size > 1, E, D = C.map((e, t) => {
		let n = T && e.path !== void 0 && e.path !== E ? e.path : void 0;
		return E = e.path, /* @__PURE__ */ (0, J.jsxs)(Y.Fragment, { children: [n && /* @__PURE__ */ (0, J.jsx)(fb, { path: n }), g === "split" ? /* @__PURE__ */ (0, J.jsx)(lb, {
			hunk: e,
			index: t,
			showLineNumbers: r,
			annotations: w
		}) : /* @__PURE__ */ (0, J.jsx)(cb, {
			hunk: e,
			index: t,
			showLineNumbers: r,
			annotations: w
		})] }, t);
	});
	return i ? /* @__PURE__ */ (0, J.jsx)("div", {
		"data-scroll-x": !0,
		className: q("overflow-x-auto font-mono text-xs leading-relaxed", a),
		children: D
	}) : /* @__PURE__ */ (0, J.jsxs)("div", {
		className: q("overflow-hidden rounded-md border border-border bg-muted/40", a),
		children: [
			/* @__PURE__ */ (0, J.jsxs)("div", {
				className: "flex items-center justify-between gap-2 border-b border-border px-3 py-1.5",
				children: [/* @__PURE__ */ (0, J.jsx)("span", {
					className: "text-[11px] uppercase tracking-wide text-muted-foreground",
					children: t || "diff"
				}), /* @__PURE__ */ (0, J.jsx)(uf, {
					size: "sm",
					"aria-label": "Diff view",
					value: g,
					onChange: _,
					options: $y
				})]
			}),
			l != null || u != null ? /* @__PURE__ */ (0, J.jsxs)("div", {
				className: q("border-b border-border text-xs font-medium", g === "split" ? "grid grid-cols-2 divide-x divide-border" : "flex flex-wrap gap-x-4"),
				children: [/* @__PURE__ */ (0, J.jsxs)("div", {
					"data-diff-side": "original",
					className: "min-w-0 break-words px-3 py-1.5",
					children: [g === "unified" ? "− " : null, l]
				}), /* @__PURE__ */ (0, J.jsxs)("div", {
					"data-diff-side": "modified",
					className: "min-w-0 break-words px-3 py-1.5",
					children: [g === "unified" ? "+ " : null, u]
				})]
			}) : null,
			/* @__PURE__ */ (0, J.jsx)("div", {
				"data-scroll-x": !0,
				className: "overflow-x-auto font-mono text-xs leading-relaxed",
				children: D
			})
		]
	});
}
function ib(e) {
	let t = [], n = [];
	for (let r of e) for (let e of r.lines) e.type !== "add" && t.push(e.content), e.type !== "remove" && n.push(e.content);
	return {
		oldSource: t.join("\n"),
		newSource: n.join("\n")
	};
}
function ab(e, t, n) {
	let r = 0, i = 0;
	return e.map((e) => ({
		header: e.header,
		path: e.path,
		lines: e.lines.map((e, a) => {
			let o = {
				line: e,
				index: a,
				oldTokens: e.type === "add" ? void 0 : t?.[r],
				newTokens: e.type === "remove" ? void 0 : n?.[i]
			};
			return e.type !== "add" && (r += 1), e.type !== "remove" && (i += 1), o;
		})
	}));
}
function ob(e, t, n) {
	return t === "old" ? e.type !== "add" && e.oldNumber === n : e.type !== "remove" && e.newNumber === n;
}
function sb({ widgets: e }) {
	return /* @__PURE__ */ (0, J.jsx)(J.Fragment, { children: e.map((e) => /* @__PURE__ */ (0, J.jsx)("div", {
		"data-diff-widget": e.key,
		className: "whitespace-normal border-y border-border bg-background px-3 py-2 font-sans text-sm leading-normal text-foreground",
		children: e.node
	}, e.key)) });
}
function cb({ hunk: e, index: t, showLineNumbers: n, annotations: r }) {
	let { onLineAction: i, lineWidgets: a, selection: o } = r;
	return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)(gb, {
		hunk: e,
		index: t,
		selection: o
	}), e.lines.map((e, r) => {
		let { line: s } = e, c = s.type === "remove" ? e.oldTokens : e.newTokens, l = s.type === "remove" ? "old" : "new", u = a?.filter((e) => ob(s, e.side, e.line)), d = o ? mb(o, s) : !1;
		return /* @__PURE__ */ (0, J.jsxs)(Y.Fragment, { children: [/* @__PURE__ */ (0, J.jsxs)("div", {
			className: q("group flex", eb[s.type], d && pb),
			"data-diff-line": s.type,
			"data-selected": d || void 0,
			children: [
				o && /* @__PURE__ */ (0, J.jsx)(hb, {
					controls: o,
					hunk: t,
					resolved: e
				}),
				n && /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)(_b, {
					value: s.oldNumber,
					line: s,
					side: "old",
					actionSide: l,
					onLineAction: i
				}), /* @__PURE__ */ (0, J.jsx)(_b, {
					value: s.newNumber,
					line: s,
					side: "new",
					actionSide: l,
					onLineAction: i
				})] }),
				/* @__PURE__ */ (0, J.jsx)(vb, { type: s.type }),
				/* @__PURE__ */ (0, J.jsx)("code", {
					className: "whitespace-pre pl-1 pr-3",
					children: /* @__PURE__ */ (0, J.jsx)(Qy, {
						tokens: c,
						content: s.content
					})
				})
			]
		}), u && /* @__PURE__ */ (0, J.jsx)(sb, { widgets: u })] }, r);
	})] });
}
function lb({ hunk: e, index: t, showLineNumbers: n, annotations: r }) {
	let i = (0, Y.useMemo)(() => ub(e.lines), [e.lines]), { onLineAction: a, lineWidgets: o, selection: s } = r;
	return /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)(gb, {
		hunk: e,
		index: t,
		selection: s
	}), i.map((e, r) => {
		let i = o?.filter((t) => e.left !== void 0 && ob(e.left.line, t.side, t.line) && t.side === "old" || e.right !== void 0 && ob(e.right.line, t.side, t.line) && t.side === "new");
		return /* @__PURE__ */ (0, J.jsxs)(Y.Fragment, { children: [/* @__PURE__ */ (0, J.jsxs)("div", {
			className: "flex",
			children: [
				/* @__PURE__ */ (0, J.jsx)(db, {
					resolved: e.left,
					side: "old",
					hunk: t,
					showLineNumbers: n,
					onLineAction: a,
					selection: s
				}),
				/* @__PURE__ */ (0, J.jsx)("span", {
					"aria-hidden": !0,
					className: "w-px shrink-0 bg-border"
				}),
				/* @__PURE__ */ (0, J.jsx)(db, {
					resolved: e.right,
					side: "new",
					hunk: t,
					showLineNumbers: n,
					onLineAction: a,
					selection: s
				})
			]
		}), i && /* @__PURE__ */ (0, J.jsx)(sb, { widgets: i })] }, r);
	})] });
}
function ub(e) {
	let t = [], n = [], r = [], i = () => {
		let e = Math.max(n.length, r.length);
		for (let i = 0; i < e; i++) t.push({
			left: n[i],
			right: r[i]
		});
		n = [], r = [];
	};
	for (let a of e) a.line.type === "remove" ? n.push(a) : a.line.type === "add" ? r.push(a) : (i(), t.push({
		left: a,
		right: a
	}));
	return i(), t;
}
function db({ resolved: e, side: t, hunk: n, showLineNumbers: r, onLineAction: i, selection: a }) {
	if (!e) return /* @__PURE__ */ (0, J.jsx)("div", { className: "min-w-0 flex-1 basis-0" });
	let { line: o } = e, s = t === "old" && o.type === "remove" || t === "new" && o.type === "add", c = s ? o.type : "context", l = t === "old" ? o.oldNumber : o.newNumber, u = t === "old" ? e.oldTokens : e.newTokens, d = s && a ? mb(a, o) : !1;
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: q("group flex min-w-0 flex-1 basis-0", s && eb[o.type], d && pb),
		"data-diff-line": c,
		"data-selected": d || void 0,
		children: [
			a && (s ? /* @__PURE__ */ (0, J.jsx)(hb, {
				controls: a,
				hunk: n,
				resolved: e
			}) : /* @__PURE__ */ (0, J.jsx)(Yy, {})),
			r && /* @__PURE__ */ (0, J.jsx)(_b, {
				value: l,
				line: o,
				side: t,
				actionSide: t,
				onLineAction: i
			}),
			/* @__PURE__ */ (0, J.jsx)(vb, { type: c }),
			/* @__PURE__ */ (0, J.jsx)("code", {
				className: "min-w-0 flex-1 whitespace-pre-wrap pl-1 pr-3 [overflow-wrap:anywhere]",
				children: /* @__PURE__ */ (0, J.jsx)(Qy, {
					tokens: u,
					content: o.content
				})
			})
		]
	});
}
function fb({ path: e }) {
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		"data-diff-file": e,
		className: "flex items-center gap-1.5 border-b border-border bg-muted px-2 py-1 font-medium text-foreground",
		children: [/* @__PURE__ */ (0, J.jsx)(Ft, { className: "size-3.5 shrink-0 text-muted-foreground" }), /* @__PURE__ */ (0, J.jsx)("span", {
			className: "truncate",
			children: e
		})]
	});
}
var pb = "relative before:absolute before:inset-y-0 before:left-0 before:w-[3px] before:bg-emerald-500";
function mb(e, t) {
	let n = Ty(t);
	return n !== void 0 && e.selection.isSelected(n);
}
function hb({ controls: e, hunk: t, resolved: n }) {
	let r = Ty(n.line);
	if (!r) return /* @__PURE__ */ (0, J.jsx)(Yy, {});
	let { selection: i } = e, a = i.isSelected(r);
	return /* @__PURE__ */ (0, J.jsx)(Xy, {
		state: a ? "all" : "none",
		label: i.label?.(r, a) ?? `${a ? "Deselect" : "Select"} ${r.side === "new" ? "added" : "removed"} line ${r.line}`,
		disabled: i.disabled === !0,
		onToggle: (r) => e.onLineClick(t, n.index, r.shiftKey)
	});
}
function gb({ hunk: e, index: t, selection: n }) {
	if (!n) return e.header ? /* @__PURE__ */ (0, J.jsx)("div", {
		className: "select-none bg-muted/60 px-2 py-0.5 text-muted-foreground",
		children: e.header
	}) : null;
	let r = Ey(e.lines.map((e) => e.line)), i = Oy(r, n.selection.isSelected), a = i === "all", o = t + 1;
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		"data-diff-hunk": o,
		className: "flex select-none items-center bg-muted/60 pr-2 text-muted-foreground",
		children: [r.length === 0 ? /* @__PURE__ */ (0, J.jsx)(Yy, {}) : /* @__PURE__ */ (0, J.jsx)(Xy, {
			state: i,
			label: n.selection.hunkLabel?.(o, a) ?? `${a ? "Deselect" : "Select"} hunk ${o}`,
			disabled: n.selection.disabled === !0,
			onToggle: () => n.onHunkClick(t)
		}), /* @__PURE__ */ (0, J.jsx)("span", {
			className: "min-w-0 flex-1 truncate py-0.5",
			children: e.header ?? `Hunk ${o}`
		})]
	});
}
function _b({ value: e, line: t, side: n, actionSide: r, onLineAction: i }) {
	return /* @__PURE__ */ (0, J.jsxs)("span", {
		className: "relative w-10 shrink-0 select-none px-2 text-right tabular-nums text-muted-foreground",
		children: [e ?? "", i !== void 0 && e !== void 0 && n === r && /* @__PURE__ */ (0, J.jsx)("button", {
			type: "button",
			"aria-label": `Comment on line ${e}`,
			onClick: (r) => i({
				side: n,
				line: e,
				content: t.content,
				element: r.currentTarget
			}),
			className: "absolute inset-y-0 left-0.5 flex items-center rounded-sm text-primary opacity-0 hover:text-primary/80 focus-visible:opacity-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring group-hover:opacity-100",
			children: /* @__PURE__ */ (0, J.jsx)(X, { className: "size-3.5" })
		})]
	});
}
function vb({ type: e }) {
	return /* @__PURE__ */ (0, J.jsx)("span", {
		className: q("w-4 shrink-0 select-none text-center", tb[e]),
		children: nb[e]
	});
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/dompurify/3.4.17/01e4a3f71ba72cce59e84f85ac1722101fd2d9d7e0e33d37c75eb81d762bf897/node_modules/dompurify/dist/purify.es.mjs
function yb(e, t) {
	this.v = e, this.k = t;
}
function bb(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function xb(e) {
	if (Array.isArray(e)) return e;
}
function Sb(e, t) {
	var n = e == null ? null : typeof Symbol < "u" && e[Symbol.iterator] || e["@@iterator"];
	if (n != null) {
		var r, i, a, o, s = [], c = !0, l = !1;
		try {
			if (a = (n = n.call(e)).next, t === 0) {
				if (Object(n) !== n) return;
				c = !1;
			} else for (; !(c = (r = a.call(n)).done) && (s.push(r.value), s.length !== t); c = !0);
		} catch (e) {
			l = !0, i = e;
		} finally {
			try {
				if (!c && n.return != null && (o = n.return(), Object(o) !== o)) return;
			} finally {
				if (l) throw i;
			}
		}
		return s;
	}
}
function Cb() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function wb(e, t) {
	return xb(e) || Sb(e, t) || Tb(e, t) || Cb();
}
function Tb(e, t) {
	if (e) {
		if (typeof e == "string") return bb(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? bb(e, t) : void 0;
	}
}
function Eb(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof yb;
			Promise.resolve(s ? o.v : o).then(function(n) {
				if (s) {
					var c = t === "return" && o.k ? t : "next";
					if (!o.k || n.done) return r(c, n);
					n = e[c](n).value;
				}
				i(!!a.done, n);
			}, function(e) {
				r("throw", e);
			});
		} catch (e) {
			i(2, e);
		}
	}
	function i(e, i) {
		e === 2 ? t.reject(i) : t.resolve({
			value: i,
			done: e
		}), (t = t.next) ? r(t.key, t.arg) : n = null;
	}
	this._invoke = function(e, i) {
		return new Promise(function(a, o) {
			var s = {
				key: e,
				arg: i,
				resolve: a,
				reject: o,
				next: null
			};
			n ? n = n.next = s : (t = n = s, r(e, i));
		});
	}, typeof e.return != "function" && (this.return = void 0);
}
Eb.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, Eb.prototype.next = function(e) {
	return this._invoke("next", e);
}, Eb.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, Eb.prototype.return = function(e) {
	return this._invoke("return", e);
};
var Db = Object.entries, Ob = Object.setPrototypeOf, kb = Object.isFrozen, Ab = Object.getPrototypeOf, jb = Object.getOwnPropertyDescriptor, Mb = Object.freeze, Nb = Object.seal, Pb = Object.create, Fb = typeof Reflect < "u" && Reflect, Ib = Fb.apply, Lb = Fb.construct;
Mb ||= function(e) {
	return e;
}, Nb ||= function(e) {
	return e;
}, Ib ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, Lb ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var Rb = cx(Array.prototype.forEach);
Array.prototype.indexOf;
var zb = cx(Array.prototype.lastIndexOf), Bb = cx(Array.prototype.pop), Vb = cx(Array.prototype.push), Hb = cx(Array.prototype.slice), Ub = cx(Array.prototype.splice), Wb = Array.isArray, Gb = lx(Set), Kb = cx(Set.prototype.add), qb = cx(Set.prototype.has), Jb = cx(String.prototype.toLowerCase), Yb = cx(String.prototype.toString), Xb = cx(String.prototype.match), Zb = cx(String.prototype.replace), Qb = cx(String.prototype.indexOf), $b = cx(String.prototype.trim), ex = cx(Number.prototype.toString), tx = cx(Boolean.prototype.toString), nx = typeof BigInt > "u" ? null : cx(BigInt.prototype.toString), rx = typeof Symbol > "u" ? null : cx(Symbol.prototype.toString), ix = cx(Object.prototype.hasOwnProperty), ax = cx(Object.prototype.toString), ox = cx(RegExp.prototype.test), sx = lx(TypeError);
function cx(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return Ib(e, t, n);
	};
}
function lx(e) {
	return function() {
		return Lb(e, [...arguments]);
	};
}
function $(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : Jb;
	if (Ob && Ob(e, null), !Wb(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (kb(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function ux(e) {
	for (let t = 0; t < e.length; t++) ix(e, t) || (e[t] = null);
	return e;
}
function dx(e) {
	let t = Pb(null);
	for (let r of Db(e)) {
		var n = wb(r, 2);
		let i = n[0], a = n[1];
		ix(e, i) && (t[i] = Wb(a) ? ux(a) : a && typeof a == "object" && a.constructor === Object ? dx(a) : a);
	}
	return t;
}
function fx(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return ex(e);
		case "boolean": return tx(e);
		case "bigint": return nx ? nx(e) : "0";
		case "symbol": return rx ? rx(e) : "Symbol()";
		case "undefined": return ax(e);
		case "function":
		case "object": {
			if (e === null) return ax(e);
			let t = e, n = px(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : ax(e);
			}
			return ax(e);
		}
		default: return ax(e);
	}
}
function px(e, t) {
	for (; e !== null;) {
		let n = jb(e, t);
		if (n) {
			if (n.get) return cx(n.get);
			if (typeof n.value == "function") return cx(n.value);
		}
		e = Ab(e);
	}
	function n() {
		return null;
	}
	return n;
}
function mx(e, t) {
	for (; e !== null;) {
		let n = jb(e, t);
		if (n) return n.get ? cx(n.get) : null;
		e = Ab(e);
	}
	return null;
}
function hx(e) {
	try {
		return ox(e, ""), !0;
	} catch {
		return !1;
	}
}
var gx = Mb(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), _x = Mb(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), vx = Mb([
	"feBlend",
	"feColorMatrix",
	"feComponentTransfer",
	"feComposite",
	"feConvolveMatrix",
	"feDiffuseLighting",
	"feDisplacementMap",
	"feDistantLight",
	"feDropShadow",
	"feFlood",
	"feFuncA",
	"feFuncB",
	"feFuncG",
	"feFuncR",
	"feGaussianBlur",
	"feImage",
	"feMerge",
	"feMergeNode",
	"feMorphology",
	"feOffset",
	"fePointLight",
	"feSpecularLighting",
	"feSpotLight",
	"feTile",
	"feTurbulence"
]), yx = Mb([
	"animate",
	"color-profile",
	"cursor",
	"discard",
	"font-face",
	"font-face-format",
	"font-face-name",
	"font-face-src",
	"font-face-uri",
	"foreignobject",
	"hatch",
	"hatchpath",
	"mesh",
	"meshgradient",
	"meshpatch",
	"meshrow",
	"missing-glyph",
	"script",
	"set",
	"solidcolor",
	"unknown",
	"use"
]), bx = Mb(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), xx = Mb([
	"maction",
	"maligngroup",
	"malignmark",
	"mlongdiv",
	"mscarries",
	"mscarry",
	"msgroup",
	"mstack",
	"msline",
	"msrow",
	"semantics",
	"annotation",
	"annotation-xml",
	"mprescripts",
	"none"
]), Sx = Mb(["#text"]), Cx = Mb(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), wx = Mb(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), Tx = Mb(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), Ex = Mb([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), Dx = Nb(/{{[\w\W]*|^[\w\W]*}}/g), Ox = Nb(/<%[\w\W]*|^[\w\W]*%>/g), kx = Nb(/\${[\w\W]*/g), Ax = Nb(/^data-[\-\w.\u00B7-\uFFFF]+$/), jx = Nb(/^aria-[\-\w]+$/), Mx = Nb(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), Nx = Nb(/^(?:\w+script|data):/i), Px = Nb(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), Fx = Nb(/^html$/i), Ix = Nb(/^[a-z][.\w]*(-[.\w]+)+$/i), Lx = Nb(/<[/\w!]/g), Rx = Nb(/<[/\w]/g), zx = Nb(/<\/no(script|embed|frames)/i), Bx = Nb(/\/>/i), Vx = {
	element: 1,
	attribute: 2,
	text: 3,
	cdataSection: 4,
	entityReference: 5,
	entityNode: 6,
	processingInstruction: 7,
	comment: 8,
	document: 9,
	documentType: 10,
	documentFragment: 11,
	notation: 12
}, Hx = Mb($({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), Ux = function() {
	return typeof window > "u" ? null : window;
}, Wx = function(e, t) {
	if (typeof e != "object" || typeof e.createPolicy != "function") return null;
	let n = null, r = "data-tt-policy-suffix";
	t && t.hasAttribute(r) && (n = t.getAttribute(r));
	let i = "dompurify" + (n ? "#" + n : "");
	try {
		return e.createPolicy(i, {
			createHTML(e) {
				return e;
			},
			createScriptURL(e) {
				return e;
			}
		});
	} catch {
		return console.warn("TrustedTypes policy " + i + " could not be created."), null;
	}
}, Gx = function() {
	return {
		afterSanitizeAttributes: [],
		afterSanitizeElements: [],
		afterSanitizeShadowDOM: [],
		beforeSanitizeAttributes: [],
		beforeSanitizeElements: [],
		beforeSanitizeShadowDOM: [],
		uponSanitizeAttribute: [],
		uponSanitizeElement: [],
		uponSanitizeShadowNode: []
	};
}, Kx = function(e, t, n, r) {
	return ix(e, t) && Wb(e[t]) ? $(r.base ? dx(r.base) : {}, e[t], r.transform) : n;
}, qx = function(e, t, n) {
	let r = ix(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? dx(r) : n();
};
function Jx() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : Ux(), t = (e) => Jx(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== Vx.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = px(d, "cloneNode"), p = px(d, "remove"), m = px(d, "removeAttributeNode"), h = px(d, "nextSibling"), g = px(d, "childNodes"), _ = px(d, "parentNode"), v = px(d, "shadowRoot"), y = px(d, "attributes"), b = o && o.prototype ? px(o.prototype, "nodeType") : null, x = o && o.prototype ? px(o.prototype, "nodeName") : null, S = o && o.prototype ? px(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, w = function(e) {
		return x ? x(e) : e.nodeName;
	}, T = mx(d, "namespaceURI"), E = function(e) {
		return T ? T(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let D, O = "", k, A = !1, j = 0, M = function() {
		if (j > 0) throw sx("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
	}, N = function(e) {
		M(), j++;
		try {
			return D.createHTML(e);
		} finally {
			j--;
		}
	}, P = function(e) {
		M(), j++;
		try {
			return D.createScriptURL(e);
		} finally {
			j--;
		}
	}, F = function() {
		return A ||= (k = Wx(u, i), !0), k;
	}, I = n, L = I.implementation, R = I.createNodeIterator, z = I.createDocumentFragment, B = I.getElementsByTagName, V = r.importNode, H = Gx();
	t.isSupported = typeof Db == "function" && typeof _ == "function" && L && L.createHTMLDocument !== void 0;
	let U = Dx, ee = Ox, te = kx, W = Ax, ne = jx, re = Nx, ie = Px, ae = Ix, oe = Mx, G = null, se = $({}, [
		...gx,
		..._x,
		...vx,
		...bx,
		...Sx
	]), K = null, ce = $({}, [
		...Cx,
		...wx,
		...Tx,
		...Ex
	]), le = Object.seal(Pb(null, {
		tagNameCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		attributeNameCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		allowCustomizedBuiltInElements: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: !1
		}
	})), ue = null, de = null, q = Object.seal(Pb(null, {
		tagCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		attributeCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		}
	})), fe = !0, pe = !0, me = !1, he = !0, ge = !1, _e = !0, ve = !1, ye = !1, be = null, xe = null, Se = !1, Ce = !1, we = !1, Te = !1, Ee = !0, De = !1, Oe = "user-content-", ke = !0, Ae = !1, je = null, Me = {}, Ne = null, Pe = $({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Fe = null, J = $({}, [
		"audio",
		"video",
		"img",
		"source",
		"image",
		"track"
	]), Y = null, Ie = $({}, [
		"alt",
		"class",
		"for",
		"id",
		"label",
		"name",
		"pattern",
		"placeholder",
		"role",
		"summary",
		"title",
		"value",
		"style",
		"xmlns"
	]), Le = "http://www.w3.org/1998/Math/MathML", Re = "http://www.w3.org/2000/svg", ze = "http://www.w3.org/1999/xhtml", Be = ze, Ve = !1, He = null, Ue = $({}, [
		Le,
		Re,
		ze
	], Yb), We = Mb([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), Ge = $({}, We), Ke = Mb(["annotation-xml"]), qe = $({}, Ke), Je = $({}, [
		"title",
		"style",
		"font",
		"a",
		"script"
	]), Ye = null, Xe = ["application/xhtml+xml", "text/html"], X = null, Ze = null, Qe = n.createElement("form"), $e = function(e) {
		return e instanceof RegExp || e instanceof Function;
	}, et = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		if (Ze && Ze === e) return;
		e && typeof e == "object" || (e = {}), e = dx(e), Ye = Xe.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, X = Ye === "application/xhtml+xml" ? Yb : Jb, G = Kx(e, "ALLOWED_TAGS", se, { transform: X }), K = Kx(e, "ALLOWED_ATTR", ce, { transform: X }), He = Kx(e, "ALLOWED_NAMESPACES", Ue, { transform: Yb }), Y = Kx(e, "ADD_URI_SAFE_ATTR", Ie, {
			transform: X,
			base: Ie
		}), Fe = Kx(e, "ADD_DATA_URI_TAGS", J, {
			transform: X,
			base: J
		}), Ne = Kx(e, "FORBID_CONTENTS", Pe, { transform: X }), ue = Kx(e, "FORBID_TAGS", dx({}), { transform: X }), de = Kx(e, "FORBID_ATTR", dx({}), { transform: X }), Me = ix(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? dx(e.USE_PROFILES) : e.USE_PROFILES : !1, fe = e.ALLOW_ARIA_ATTR !== !1, pe = e.ALLOW_DATA_ATTR !== !1, me = e.ALLOW_UNKNOWN_PROTOCOLS || !1, he = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, ge = e.SAFE_FOR_TEMPLATES || !1, _e = e.SAFE_FOR_XML !== !1, ve = e.WHOLE_DOCUMENT || !1, Ce = e.RETURN_DOM || !1, we = e.RETURN_DOM_FRAGMENT || !1, Te = e.RETURN_TRUSTED_TYPE || !1, Se = e.FORCE_BODY || !1, Ee = e.SANITIZE_DOM !== !1, De = e.SANITIZE_NAMED_PROPS || !1, ke = e.KEEP_CONTENT !== !1, Ae = e.IN_PLACE || !1, oe = hx(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : Mx, Be = typeof e.NAMESPACE == "string" ? e.NAMESPACE : ze, Ge = qx(e, "MATHML_TEXT_INTEGRATION_POINTS", () => $({}, We)), qe = qx(e, "HTML_INTEGRATION_POINTS", () => $({}, Ke));
		let t = qx(e, "CUSTOM_ELEMENT_HANDLING", () => Pb(null));
		if (le = Pb(null), ix(t, "tagNameCheck") && $e(t.tagNameCheck) && (le.tagNameCheck = t.tagNameCheck), ix(t, "attributeNameCheck") && $e(t.attributeNameCheck) && (le.attributeNameCheck = t.attributeNameCheck), ix(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (le.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), Nb(le), ge && (pe = !1), we && (Ce = !0), Me && (G = $({}, Sx), K = Pb(null), Me.html === !0 && ($(G, gx), $(K, Cx)), Me.svg === !0 && ($(G, _x), $(K, wx), $(K, Ex)), Me.svgFilters === !0 && ($(G, vx), $(K, wx), $(K, Ex)), Me.mathMl === !0 && ($(G, bx), $(K, Tx), $(K, Ex))), q.tagCheck = null, q.attributeCheck = null, ix(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? q.tagCheck = e.ADD_TAGS : Wb(e.ADD_TAGS) && (G === se && (G = dx(G)), $(G, e.ADD_TAGS, X))), ix(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? q.attributeCheck = e.ADD_ATTR : Wb(e.ADD_ATTR) && (K === ce && (K = dx(K)), $(K, e.ADD_ATTR, X))), ix(e, "ADD_FORBID_CONTENTS") && Wb(e.ADD_FORBID_CONTENTS) && (Ne === Pe && (Ne = dx(Ne)), $(Ne, e.ADD_FORBID_CONTENTS, X)), ke && (G["#text"] = !0), ve && $(G, [
			"html",
			"head",
			"body"
		]), G.table && ($(G, ["tbody"]), delete ue.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw sx("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw sx("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = D;
			D = e.TRUSTED_TYPES_POLICY;
			try {
				O = N("");
			} catch (e) {
				throw D = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (D = void 0, O = "") : (D === void 0 && (D = F()), D && typeof O == "string" && (O = N("")));
		Mb && Mb(e), Ze = e;
	}, tt = $({}, [
		..._x,
		...vx,
		...yx
	]), nt = $({}, [...bx, ...xx]), rt = function(e, t, n) {
		return t.namespaceURI === ze ? e === "svg" : t.namespaceURI === Le ? e === "svg" && (n === "annotation-xml" || Ge[n]) : !!tt[e];
	}, it = function(e, t, n) {
		return t.namespaceURI === ze ? e === "math" : t.namespaceURI === Re ? e === "math" && qe[n] : !!nt[e];
	}, at = function(e, t, n) {
		return t.namespaceURI === Re && !qe[n] || t.namespaceURI === Le && !Ge[n] ? !1 : !nt[e] && (Je[e] || !tt[e]);
	}, ot = function(e) {
		let t = E(e), n = Jb(w(e)), r = _(e), i, a;
		return r && C(r) === Vx.element ? (i = { namespaceURI: E(r) }, a = Jb(w(r))) : (i = { namespaceURI: Be }, a = "template"), He[t] ? t === Re ? rt(n, i, a) : t === Le ? it(n, i, a) : t === ze ? at(n, i, a) : !!(Ye === "application/xhtml+xml" && He[t]) : !1;
	}, st = function(e) {
		Vb(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw sx("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
		}
	}, ct = function(e, t, n) {
		try {
			m(e, t);
		} catch {
			try {
				e.removeAttribute(n);
			} catch {}
		}
	}, lt = function(e) {
		pt(e);
		let t = g(e);
		t && Rb(Hb(t), (e) => {
			try {
				p(e);
			} catch {}
		});
		let n = C(e) === Vx.element ? y(e) : null;
		if (n) for (let t = n.length - 1; t >= 0; --t) {
			let r = n[t], i = r && r.name;
			typeof i == "string" && ct(e, r, i);
		}
	}, ut = function(e, n, r) {
		if (!r) try {
			r = n.getAttributeNode(e);
		} catch {
			r = null;
		}
		Vb(t.removed, {
			attribute: r || null,
			from: n
		});
		try {
			r ? m(n, r) : n.removeAttribute(e);
		} catch {
			try {
				n.removeAttribute(e);
			} catch {}
		}
		if (e === "is") {
			if (Ce || we) try {
				st(n);
			} catch {}
			else try {
				n.setAttribute(e, "");
			} catch {}
		}
	}, dt = function(e) {
		let t = y(e);
		if (t) for (let n = t.length - 1; n >= 0; --n) {
			let r = t[n], i = r && r.name;
			typeof i != "string" || K[X(i)] || ct(e, r, i);
		}
	}, ft = function(e, t) {
		let n = arguments.length > 2 && arguments[2] !== void 0 && arguments[2], r = [e];
		for (; r.length > 0;) {
			let e = r.pop();
			if (t(e) === !1) continue;
			if (n && C(e) === Vx.element) {
				let t = v(e);
				Ot(t) && r.push(t);
			}
			let i = g(e);
			if (i) for (let e = i.length - 1; e >= 0; --e) r.push(i[e]);
		}
	}, pt = function(e) {
		ft(e, (e) => {
			C(e) === Vx.element && dt(e);
		});
	}, mt = function(e) {
		let t = [];
		return ft(e, (e) => {
			t.push(e);
		}, !0), t;
	}, ht = function(e, t) {
		let n = Gb();
		Rb(mt(e), (e) => {
			C(e) === Vx.element && Kb(n, e);
		}), Rb(t, (e) => {
			C(e) === Vx.element && !qb(n, e) && dt(e);
		});
	}, gt = function(e, n) {
		let r = !1;
		return Rb(t.removed, (t) => {
			t.element && (r ||= t.element === e, pt(t.element));
		}), ht(e, n), r;
	}, _t = function(e, t) {
		let n = je, r = null, i = () => (r || (r = Gb(), Rb(t, (e) => {
			Kb(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = mt(e), a = [e], o = !1;
			if (Rb(r, (e) => {
				C(e) === Vx.documentFragment ? Vb(a, e) : !o && !qb(n, e) && (o = qb(i(), e));
			}), !o) return;
			if (t >= 8) throw sx("the tree kept changing during in-place sanitization; refusing to return it");
			Rb(a, (t, n) => {
				vt(t, n > 0, e, i());
			});
		}
	}, vt = function(e, t, n, r) {
		let i = Ct(e), a;
		for (; a = i.nextNode();) qb(r, a) && !qb(je, a) && (t && At(H.uponSanitizeShadowNode, a, null), It(a, n), Ut(a, n), Ot(a.content) && Wt(a.content));
	}, yt = function(e) {
		je && Kb(je, e);
	}, bt = function(e, t) {
		return _e ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, xt = function(e) {
		_e && ft(e, (e) => {
			let t = C(e);
			if (t === Vx.processingInstruction || t === Vx.comment && ox(Rx, e.data)) {
				try {
					p(e);
				} catch {}
				return !1;
			}
			if (t === Vx.element) {
				let t = e, n = X(w(e));
				try {
					t.hasAttribute && t.hasAttribute("patchsrc") && t.removeAttribute("patchsrc"), t.hasAttribute && t.hasAttribute("for") && bt("for", n) && t.removeAttribute("for");
				} catch {}
			}
			return !0;
		});
	}, St = function(e) {
		let t = null, r = null;
		if (Se) e = "<remove></remove>" + e;
		else {
			let t = Xb(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		Ye === "application/xhtml+xml" && Be === ze && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = D ? N(e) : e;
		if (Be === ze) try {
			t = new l().parseFromString(i, Ye);
		} catch {}
		if (!t || !t.documentElement) {
			t = L.createDocument(Be, "template", null);
			try {
				t.documentElement.innerHTML = Ve ? O : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), Be === ze ? B.call(t, ve ? "html" : "body")[0] : ve ? t.documentElement : a;
	}, Ct = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return R.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, wt = function(e) {
		return e = Zb(e, U, " "), e = Zb(e, ee, " "), e = Zb(e, te, " "), e;
	}, Tt = function(e) {
		Et(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && Rb(t, (e) => {
			Ot(e.content) && Tt(e.content);
		});
	}, Et = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = R.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
		for (; r;) r.data = wt(r.data), r = n.nextNode();
	}, Dt = function(e) {
		let t = x ? x(e) : null;
		return typeof t != "string" || X(t) !== "form" ? !1 : typeof e.nodeName != "string" || typeof e.textContent != "string" || typeof e.removeChild != "function" || e.attributes !== y(e) || typeof e.removeAttribute != "function" || typeof e.removeAttributeNode != "function" || typeof e.getAttributeNode != "function" || typeof e.setAttribute != "function" || typeof e.namespaceURI != "string" || typeof e.insertBefore != "function" || typeof e.hasChildNodes != "function" || e.nodeType !== b(e) || e.childNodes !== g(e);
	}, Ot = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return b(e) === Vx.documentFragment;
		} catch {
			return !1;
		}
	}, kt = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return typeof b(e) == "number";
		} catch {
			return !1;
		}
	};
	function At(e, n, r) {
		e.length !== 0 && Rb(e, (e) => {
			e.call(t, n, r, Ze);
		});
	}
	let jt = function(e, t) {
		return !!(_e && e.hasChildNodes() && !kt(e.firstElementChild) && ox(Lx, e.textContent) && ox(Lx, e.innerHTML) || _e && Hx[t] && C(e) === Vx.element && E(e) === ze && (kt(e.firstElementChild) || typeof e.textContent == "string" && ox(Lx, e.textContent)) || e.nodeType === Vx.processingInstruction || _e && e.nodeType === Vx.comment && ox(Rx, e.data));
	}, Mt = function(e, t) {
		return e instanceof RegExp ? ox(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, Nt = function(e, t, n) {
		if (!ue[t] && zt(t) && Mt(le.tagNameCheck, t)) return !1;
		if (ke && !Ne[t] && !Ne[Jb(t)]) {
			let t = _(e), r = g(e);
			if (r && t) {
				let i = r.length;
				for (let a = i - 1; a >= 0; --a) {
					let i = e === n ? f(r[a], !0) : r[a];
					t.insertBefore(i, h(e));
				}
			}
		}
		return st(e), !0;
	}, Pt = function(e, t, n, r) {
		return e.length === 0 ? t : t === n || t === r ? dx(t) : t;
	}, Ft = function(e, t) {
		return e === t || _(e) !== null ? !1 : (Ae && pt(e), !0);
	}, It = function(e, n) {
		if (yt(e), At(H.beforeSanitizeElements, e, null), Ft(e, n)) return !0;
		if (Dt(e)) return st(e), !0;
		let r = X(w(e));
		if (G = Pt(H.uponSanitizeElement, G, se, be), At(H.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: G
		}), Ft(e, n)) return !0;
		if (jt(e, r)) return st(e), !0;
		if (ue[r] || !(q.tagCheck instanceof Function && q.tagCheck(r)) && !G[r]) {
			let t = Nt(e, r, n);
			return t === !1 && (At(H.afterSanitizeElements, e, null), Ft(e, n)) ? !0 : t;
		}
		if (C(e) === Vx.element && !ot(e) || (r === "noscript" || r === "noembed" || r === "noframes") && ox(zx, e.innerHTML)) return st(e), !0;
		if (ge && e.nodeType === Vx.text) {
			let n = wt(e.textContent);
			e.textContent !== n && (Vb(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return At(H.afterSanitizeElements, e, null), Ft(e, n);
	}, Lt = function(e, t, r) {
		if (de[t] || bt(t, e) || Ee && (t === "id" || t === "name") && (r in n || r in Qe)) return !1;
		let i = K[t] || q.attributeCheck instanceof Function && q.attributeCheck(t, e);
		return pe && ox(W, t) || fe && ox(ne, t) ? !0 : i ? Y[t] || ox(oe, Zb(r, ie, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && Qb(r, "data:") === 0 && Fe[e] || me && !ox(re, Zb(r, ie, "")) ? !0 : !r : zt(e) && Mt(le.tagNameCheck, e) && Mt(le.attributeNameCheck, t, e) || t === "is" && le.allowCustomizedBuiltInElements && Mt(le.tagNameCheck, r);
	}, Rt = $({}, [
		"annotation-xml",
		"color-profile",
		"font-face",
		"font-face-format",
		"font-face-name",
		"font-face-src",
		"font-face-uri",
		"missing-glyph"
	]), zt = function(e) {
		return !Rt[Jb(e)] && ox(ae, e);
	}, Bt = function(e, t, n, r) {
		if (D && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return N(r);
			case "TrustedScriptURL": return P(r);
		}
		return r;
	}, Vt = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !Dt(e) || (st(e), !1);
		} catch {
			return ut(t, e), !1;
		}
	}, Ht = function(e, t) {
		return _e && ox(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : Jb(e) === "attributename" && Xb(t, "href") !== null;
	}, Ut = function(e, n) {
		if (At(H.beforeSanitizeAttributes, e, null), Ft(e, n)) return;
		let r = e.attributes;
		if (!r || Dt(e)) return;
		K = Pt(H.uponSanitizeAttribute, K, ce, xe);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: K,
			forceKeepAttr: void 0
		}, a = r.length, o = X(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = X(s), d = l, f = s === "value" ? d : $b(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, At(H.uponSanitizeAttribute, e, i), f = i.attrValue, De && (u === "id" || u === "name") && Qb(f, Oe) !== 0 && (ut(s, e, n), f = Oe + f, p = !0), Ht(s, f) ? ut(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !he && ox(Bx, f) ? ut(s, e, n) : (ge && (f = wt(f)), Lt(o, u, f) ? (f = Bt(o, u, c, f), f !== d && Vt(e, s, c, f) && p && Bb(t.removed)) : ut(s, e, n)) : ut(s, e, n));
		}
		At(H.afterSanitizeAttributes, e, null), Ft(e, n);
	}, Wt = function(e) {
		let t = null, n = Ct(e);
		for (At(H.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (At(H.uponSanitizeShadowNode, t, null), It(t, e), Ut(t, e), Ot(t.content) && Wt(t.content), C(t) === Vx.element) {
			let e = v(t);
			Ot(e) && (Gt(e), Wt(e));
		}
		ge && Et(e), At(H.afterSanitizeShadowDOM, e, null);
	}, Gt = function(e) {
		let t = [{
			node: e,
			shadow: null
		}];
		for (; t.length > 0;) {
			let e = t.pop();
			if (e.shadow) {
				Wt(e.shadow);
				continue;
			}
			let n = e.node, r = C(n) === Vx.element, i = g(n);
			if (i) for (let e = i.length - 1; e >= 0; --e) t.push({
				node: i[e],
				shadow: null
			});
			if (r) {
				let e = x ? x(n) : null;
				if (typeof e == "string" && X(e) === "template") {
					let e = n.content;
					Ot(e) && t.push({
						node: e,
						shadow: null
					});
				}
			}
			if (r) {
				let e = v(n);
				Ot(e) && t.push({
					node: null,
					shadow: e
				}, {
					node: e,
					shadow: null
				});
			}
		}
	};
	return t.sanitize = function(e) {
		let n = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {}, i = null, a = null, o = null, s = null;
		if (Ve = !e, Ve && (e = "<!-->"), typeof e != "string" && !kt(e) && (e = fx(e), typeof e != "string")) throw sx("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		ye ? (G = be, K = xe) : et(n), (H.uponSanitizeElement.length > 0 || H.uponSanitizeAttribute.length > 0) && (G = dx(G)), H.uponSanitizeAttribute.length > 0 && (K = dx(K)), t.removed = [];
		let c = Ae && typeof e != "string" && kt(e), l = [], u = je;
		if (c) {
			xt(e);
			let t = w(e);
			if (typeof t == "string") {
				let n = X(t);
				if (!G[n] || ue[n]) throw lt(e), sx("root node is forbidden and cannot be sanitized in-place");
			}
			if (Dt(e)) throw lt(e), sx("root node is clobbered and cannot be sanitized in-place");
			l = mt(e), je = Gb();
		} else if (kt(e)) i = St("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === Vx.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), Gt(i);
		else {
			if (!Ce && !ge && !ve && e.indexOf("<") === -1) return D && Te ? N(e) : e;
			if (i = St(e), !i) return Ce ? null : Te ? O : "";
		}
		i && Se && st(i.firstChild);
		let d = c ? e : i;
		try {
			c && Gt(e);
			let t = Ct(d);
			for (; o = t.nextNode();) It(o, d), Ut(o, d), Ot(o.content) && Wt(o.content);
			c && _t(e, l);
		} catch (t) {
			throw c && (lt(e), gt(e, l)), t;
		} finally {
			je = u;
		}
		if (c) {
			if (gt(e, l)) throw sx("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return ge && Tt(e), e;
		}
		if (Ce) {
			if (ge && Tt(i), we) for (s = z.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (K.shadowroot || K.shadowrootmode) && (s = V.call(r, s, !0)), s;
		}
		let f = ve ? i.outerHTML : i.innerHTML;
		return ve && G["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && ox(Fx, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), ge && (f = wt(f)), D && Te ? N(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		et(e), ye = !0, be = G, xe = K;
	}, t.clearConfig = function() {
		Ze = null, ye = !1, be = null, xe = null, D = k, O = "";
	}, t.isValidAttribute = function(e, t, n) {
		Ze || et({});
		let r = X(e), i = X(t);
		return typeof n == "string" && (Ht(t, n) || !he && ox(Bx, n)) ? !1 : Lt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && ix(H, e) && Vb(H[e], t);
	}, t.removeHook = function(e, t) {
		if (ix(H, e)) {
			if (t !== void 0) {
				let n = zb(H[e], t);
				return n === -1 ? void 0 : Ub(H[e], n, 1)[0];
			}
			return Bb(H[e]);
		}
	}, t.removeHooks = function(e) {
		ix(H, e) && (H[e] = []);
	}, t.removeAllHooks = function() {
		H = Gx();
	}, t;
}
var Yx = Jx(), Xx = /* @__PURE__ */ new Map();
function Zx(e) {
	if (!e || typeof window > "u") return "";
	let t = Xx.get(e);
	if (t !== void 0) return Xx.delete(e), Xx.set(e, t), t;
	let n = Yx.sanitize(e, { USE_PROFILES: { html: !0 } });
	return Xx.set(e, n), Xx.size > 1e3 && Xx.delete(Xx.keys().next().value), n;
}
function Qx(e) {
	return /<(div|p|pre|table|ul|ol|li|details|blockquote|h[1-6])/i.test(e);
}
// ../../../clicky-ui/packages/ui/src/data/CodeBlock.tsx
var $x = 500, eS = 2e3;
function tS({ language: e, source: t = "", highlightedHtml: n, className: r, jsonDefaultOpenDepth: i = 2, jsonFormat: o = "json", bare: s = !1, copyable: c = !1, downloadable: l = !1, downloadFilename: u, themeToggle: d = !1, diff: f, diffView: p, inline: m = !1, toolbar: h }) {
	let g = (e ?? "").toLowerCase().replace(/^\.+/, ""), _ = n ? Zx(n) : "", v = (0, Y.useMemo)(() => g === "json" ? iS(t) : rS, [g, t]), y = a(), [b, x] = (0, Y.useState)(null), S = b ?? y, [C, w] = (0, Y.useState)(null), T = v === rS && !_ && !!g && !!t, E = (0, Y.useMemo)(() => m ? t.replace(/\s+/g, " ").trim() : "", [m, t]), D = m ? E.slice(0, $x) : t;
	(0, Y.useEffect)(() => {
		if (!T) {
			w(null);
			return;
		}
		let e = !1;
		return Ky(D, {
			lang: g,
			theme: S === "dark" ? "github-dark" : "github-light"
		}).then((t) => {
			e || w(t);
		}), () => {
			e = !0;
		};
	}, [
		T,
		g,
		D,
		S
	]);
	let O = /* @__PURE__ */ (0, J.jsx)(vy, {
		source: t,
		language: e,
		copyable: c,
		downloadable: l,
		downloadFilename: u,
		theme: S,
		...d ? { onToggleTheme: () => x(S === "dark" ? "light" : "dark") } : {}
	}), k = b ? { "data-theme": b } : {};
	if (m) {
		let e = _ || C, t = E.length > eS ? `${E.slice(0, eS)}…` : E;
		return /* @__PURE__ */ (0, J.jsx)("code", {
			className: q("block max-w-full truncate font-mono text-xs text-foreground", r, e && "[&_pre]:inline [&_pre]:whitespace-pre [&_.chroma]:bg-transparent"),
			title: t,
			children: e ? /* @__PURE__ */ (0, J.jsx)("span", { dangerouslySetInnerHTML: { __html: e } }) : D
		});
	}
	if (f) return /* @__PURE__ */ (0, J.jsx)(rb, {
		language: e,
		view: p,
		bare: s,
		className: r,
		...f
	});
	if (v !== rS) {
		let e = /* @__PURE__ */ (0, J.jsx)(Ad, {
			data: v,
			format: o,
			defaultOpenDepth: i
		});
		return s ? /* @__PURE__ */ (0, J.jsx)("div", {
			className: q("overflow-auto text-xs", r),
			children: e
		}) : /* @__PURE__ */ (0, J.jsxs)("div", {
			className: q("overflow-hidden rounded-md border border-border bg-muted/40", r),
			...k,
			children: [/* @__PURE__ */ (0, J.jsx)(nS, {
				label: o,
				toolbar: h,
				actions: O
			}), /* @__PURE__ */ (0, J.jsx)("div", {
				className: "overflow-auto p-3 text-xs",
				children: e
			})]
		});
	}
	let A = g === "xml" || g === "xslt" || g === "html" ? " [&_.chroma_.nt]:text-pink-700 [&_.chroma_.nt]:dark:text-pink-300 [&_.chroma_.nt]:font-semibold [&_.chroma_.na]:text-amber-700 [&_.chroma_.na]:dark:text-amber-300 [&_.chroma_.s]:text-emerald-700 [&_.chroma_.s]:dark:text-emerald-300 [&_.chroma_.c]:text-slate-500 [&_.chroma_.c]:italic [&_.chroma_.cp]:text-slate-500 [&_.chroma_.cp]:italic [&_pre]:leading-relaxed" : "", j = s ? void 0 : "p-3", M = _ ? /* @__PURE__ */ (0, J.jsx)("div", {
		className: q("overflow-auto text-xs font-mono [&_.chroma]:bg-transparent [&_pre]:m-0 [&_pre]:whitespace-pre-wrap [&_pre]:bg-transparent" + A, j, s && r),
		dangerouslySetInnerHTML: { __html: _ }
	}) : C ? /* @__PURE__ */ (0, J.jsx)("div", {
		className: q("overflow-auto text-xs font-mono [&_pre]:m-0 [&_pre.shiki]:!bg-transparent [&_pre]:whitespace-pre-wrap", j, s && r),
		dangerouslySetInnerHTML: { __html: C }
	}) : /* @__PURE__ */ (0, J.jsx)("pre", {
		className: q("overflow-auto whitespace-pre-wrap break-words text-xs font-mono text-foreground", j, s && r),
		children: t
	});
	return s ? M : /* @__PURE__ */ (0, J.jsxs)("div", {
		className: q("overflow-hidden rounded-md border border-border bg-muted/40", r),
		...k,
		children: [/* @__PURE__ */ (0, J.jsx)(nS, {
			label: e || "text",
			toolbar: h,
			actions: O
		}), M]
	});
}
function nS({ label: e, toolbar: t, actions: n }) {
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex items-center justify-between gap-2 border-b border-border px-3 py-1.5",
		children: [/* @__PURE__ */ (0, J.jsxs)("span", {
			className: "flex min-w-0 items-center gap-2",
			children: [/* @__PURE__ */ (0, J.jsx)("span", {
				className: "text-[11px] uppercase tracking-wide text-muted-foreground",
				children: e
			}), t]
		}), n]
	});
}
var rS = Symbol("json-parse-failed");
function iS(e) {
	let t = e.trim();
	if (!t) return rS;
	try {
		return JSON.parse(t);
	} catch {
		return rS;
	}
}
// ../../../clicky-ui/packages/ui/src/data/delimited.ts
var aS = {
	csv: ",",
	tsv: "	"
};
function oS(e, t = ",") {
	let n = [], r = [], i = "", a = !1;
	for (let o = 0; o < e.length; o++) {
		let s = e[o];
		a ? s === "\"" && e[o + 1] === "\"" ? (i += "\"", o++) : s === "\"" ? a = !1 : i += s : s === "\"" && i === "" ? a = !0 : s === t ? (r.push(i), i = "") : s === "\n" || s === "\r" ? (s === "\r" && e[o + 1] === "\n" && o++, r.push(i), n.push(r), r = [], i = "") : i += s;
	}
	if (a) throw Error("CSV: unterminated quoted field");
	return (i !== "" || r.length > 0) && (r.push(i), n.push(r)), n;
}
// ../../../clicky-ui/packages/ui/src/data/markdown-source-blocks.ts
var sS = "data-source-start", cS = "data-source-end";
function lS() {
	return (e) => {
		for (let t of e.children ?? []) t.type === "element" && t.position && (t.properties = {
			...t.properties,
			dataSourceStart: t.position.start.line,
			dataSourceEnd: t.position.end.line
		});
	};
}
function uS(e) {
	let t = e[sS], n = e[cS];
	return {
		...typeof t == "string" || typeof t == "number" ? { [sS]: t } : {},
		...typeof n == "string" || typeof n == "number" ? { [cS]: n } : {}
	};
}
// ../../../clicky-ui/packages/ui/src/data/markdown-urls.ts
function dS(e) {
	return () => (t) => {
		let n = (t) => {
			let r = t.tagName === "a" ? "href" : t.tagName === "img" ? "src" : null;
			r && t.properties && typeof t.properties[r] == "string" && (t.properties[r] = e(t.properties[r], r === "href" ? "link" : "image"));
			for (let e of t.children ?? []) n(e);
		};
		n(t);
	};
}
// ../../../clicky-ui/packages/ui/src/data/Markdown.tsx
var fS = {
	h1: "mt-4 mb-2 text-lg font-semibold",
	h2: "mt-3 mb-1.5 text-base font-semibold",
	h3: "mt-3 mb-1 text-sm font-semibold",
	h4: "mt-2 mb-1 text-sm font-semibold",
	h5: "mt-2 mb-1 text-sm font-medium",
	h6: "mt-2 mb-1 text-sm font-medium"
};
function pS({ as: e, className: t, children: n, node: r, ...i }) {
	return /* @__PURE__ */ (0, J.jsx)(e, {
		className: q(fS[e], t),
		"data-streamdown": `heading-${e.slice(1)}`,
		...i,
		children: n
	});
}
var mS = null;
function hS() {
	return mS ||= import("./dist-Bd8Idnus.js").then((e) => ({
		Streamdown: e.Streamdown,
		defaultRehypePlugins: e.defaultRehypePlugins
	})).catch((e) => (console.warn("clicky-ui: streamdown not available; rendering markdown as plain text", e), null)), mS;
}
function gS(e) {
	return typeof e == "string" ? e : typeof e == "number" ? String(e) : Array.isArray(e) ? e.map(gS).join("") : (0, Y.isValidElement)(e) ? gS(e.props.children) : "";
}
function _S({ className: e, children: t, ...n }) {
	if (!("data-block" in n)) return /* @__PURE__ */ (0, J.jsx)("code", {
		className: q("rounded bg-muted px-1.5 py-0.5 font-mono text-sm", e),
		children: t
	});
	let r = /language-([\w-]+)/.exec(e ?? "")?.[1], i = gS(t).replace(/\n$/, ""), a = r ? aS[r.toLowerCase()] : void 0, o = a ? /* @__PURE__ */ (0, J.jsx)(vS, {
		source: i,
		delimiter: a
	}) : /* @__PURE__ */ (0, J.jsx)(tS, {
		source: i,
		language: r,
		className: "not-prose my-4",
		copyable: !0,
		downloadable: !0,
		themeToggle: !0
	}), s = uS(n);
	return Object.keys(s).length > 0 ? /* @__PURE__ */ (0, J.jsx)("div", {
		...s,
		children: o
	}) : o;
}
function vS({ source: e, delimiter: t }) {
	let [n = [], ...r] = oS(e, t), i = Array.from({ length: Math.max(n.length, ...r.map((e) => e.length)) }, (e, t) => t);
	return /* @__PURE__ */ (0, J.jsxs)(SS, { children: [/* @__PURE__ */ (0, J.jsx)(CS, { children: /* @__PURE__ */ (0, J.jsx)("tr", { children: i.map((e) => /* @__PURE__ */ (0, J.jsx)(TS, { children: n[e] ?? "" }, e)) }) }), /* @__PURE__ */ (0, J.jsx)(wS, { children: r.map((e, t) => /* @__PURE__ */ (0, J.jsx)("tr", { children: i.map((t) => /* @__PURE__ */ (0, J.jsx)(ES, { children: e[t] ?? "" }, t)) }, t)) })] });
}
function yS({ children: e, ...t }) {
	return (0, Y.isValidElement)(e) ? (0, Y.cloneElement)(e, {
		"data-block": "true",
		...uS(t)
	}) : e;
}
function bS({ className: e, children: t, ...n }) {
	let r = e?.includes("contains-task-list");
	return /* @__PURE__ */ (0, J.jsx)("ul", {
		className: q("whitespace-normal", r ? "list-none pl-0" : "list-disc pl-6", e),
		...uS(n),
		children: t
	});
}
function xS({ className: e, children: t, ...n }) {
	return /* @__PURE__ */ (0, J.jsx)("ol", {
		className: q("list-decimal whitespace-normal pl-6", e),
		...uS(n),
		children: t
	});
}
function SS({ className: e, children: t, ...n }) {
	return /* @__PURE__ */ (0, J.jsx)("div", {
		className: "not-prose my-4 max-w-full overflow-x-auto rounded-md border border-border",
		...uS(n),
		children: /* @__PURE__ */ (0, J.jsx)("table", {
			className: q("w-max min-w-full table-auto text-left text-sm", e),
			children: t
		})
	});
}
function CS({ className: e, children: t }) {
	return /* @__PURE__ */ (0, J.jsx)("thead", {
		className: q("border-b border-border bg-muted", e),
		children: t
	});
}
function wS({ className: e, children: t }) {
	return /* @__PURE__ */ (0, J.jsx)("tbody", {
		className: q("divide-y divide-border", e),
		children: t
	});
}
function TS({ className: e, style: t, children: n }) {
	return /* @__PURE__ */ (0, J.jsx)("th", {
		className: q("whitespace-nowrap px-density-3 py-density-2 font-semibold", e),
		style: t,
		children: n
	});
}
function ES({ className: e, style: t, children: n }) {
	return /* @__PURE__ */ (0, J.jsx)("td", {
		className: q("px-density-3 py-density-2 align-top", e),
		style: t,
		children: n
	});
}
function DS({ src: e, alt: t, title: n, className: r, ...i }) {
	return /* @__PURE__ */ (0, J.jsx)("img", {
		src: e,
		alt: t ?? "",
		title: n,
		className: r,
		loading: "lazy",
		...uS(i)
	});
}
function OS({ href: e, title: t, className: n, children: r }) {
	let { renderLink: i } = hy(), [a, o] = (0, Y.useState)(!1), s = q("wrap-anywhere font-medium text-primary underline", n);
	return e ? e.startsWith("#") ? /* @__PURE__ */ (0, J.jsx)("a", {
		href: e,
		title: t,
		className: s,
		children: r
	}) : e.startsWith("/") && !e.startsWith("//") ? i({
		to: e,
		className: s,
		children: r,
		...t ? { title: t } : {}
	}) : /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)("a", {
		href: e,
		title: t,
		target: "_blank",
		rel: "noopener noreferrer",
		className: s,
		onClick: (e) => {
			e.preventDefault(), o(!0);
		},
		children: r
	}), /* @__PURE__ */ (0, J.jsx)(uy, {
		open: a,
		title: "Open external link?",
		message: /* @__PURE__ */ (0, J.jsx)("code", {
			className: "break-all font-mono text-sm",
			children: e
		}),
		confirmLabel: "Open link",
		onCancel: () => o(!1),
		onConfirm: () => {
			o(!1), window.open(e, "_blank", "noopener,noreferrer");
		}
	})] }) : /* @__PURE__ */ (0, J.jsx)("span", {
		className: n,
		children: r
	});
}
var kS = {
	h1: (e) => /* @__PURE__ */ (0, J.jsx)(pS, {
		as: "h1",
		...e
	}),
	h2: (e) => /* @__PURE__ */ (0, J.jsx)(pS, {
		as: "h2",
		...e
	}),
	h3: (e) => /* @__PURE__ */ (0, J.jsx)(pS, {
		as: "h3",
		...e
	}),
	h4: (e) => /* @__PURE__ */ (0, J.jsx)(pS, {
		as: "h4",
		...e
	}),
	h5: (e) => /* @__PURE__ */ (0, J.jsx)(pS, {
		as: "h5",
		...e
	}),
	h6: (e) => /* @__PURE__ */ (0, J.jsx)(pS, {
		as: "h6",
		...e
	}),
	a: OS,
	img: DS,
	pre: yS,
	code: _S,
	ul: bS,
	ol: xS,
	table: SS,
	thead: CS,
	tbody: wS,
	th: TS,
	td: ES
};
function AS({ badge: e, children: t, className: n, emphasis: r, icon: i, label: a, source: o, title: s, variant: c }) {
	return /* @__PURE__ */ (0, J.jsx)(ae, {
		badge: e,
		className: n,
		emphasis: r,
		icon: i,
		label: a,
		source: o,
		title: s,
		variant: c,
		children: t
	});
}
var jS = {
	...kS,
	calloutbox: AS
}, MS = { calloutbox: [
	"variant",
	"label",
	"badge",
	"source",
	"icon",
	"emphasis",
	"title"
] };
function NS({ text: e, className: t, callouts: n = !1, sourceBlocks: r = !1, resolveUrl: i }) {
	if (r && n) throw Error("Markdown: sourceBlocks and callouts cannot be combined; callouts replaces the sanitiser schema that sourceBlocks keeps");
	if (i && n) throw Error("Markdown: resolveUrl and callouts cannot be combined because callouts owns the sanitizer chain");
	let [a, o] = (0, Y.useState)(null);
	(0, Y.useEffect)(() => {
		let e = !1;
		return hS().then((t) => {
			e || o(t);
		}), () => {
			e = !0;
		};
	}, []);
	let s = (0, Y.useMemo)(() => {
		if (!a) return;
		let { raw: e, sanitize: t, harden: n } = a.defaultRehypePlugins;
		if (!e || !t || !n) throw Error("Streamdown must provide its raw, sanitize and harden plugins");
		return [
			e,
			t,
			...i ? [dS(i)] : [],
			n,
			...r ? [lS] : []
		];
	}, [
		a,
		r,
		i
	]);
	if (!a) return /* @__PURE__ */ (0, J.jsx)("pre", {
		className: q("whitespace-pre-wrap text-sm", t),
		children: e
	});
	let { Streamdown: c } = a;
	return /* @__PURE__ */ (0, J.jsx)("div", {
		className: q("prose prose-sm max-w-none dark:prose-invert", t),
		children: r ? /* @__PURE__ */ (0, J.jsx)(c, {
			mode: "static",
			components: kS,
			rehypePlugins: s,
			children: e
		}) : /* @__PURE__ */ (0, J.jsx)(c, {
			parseIncompleteMarkdown: !0,
			components: n ? jS : kS,
			...i ? { rehypePlugins: s } : {},
			allowedTags: n ? MS : void 0,
			children: e
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-property-selection.tsx
function PS({ field: e, ctx: t }) {
	let n = e.lookup, r = ug(n), i = t.rootValue, [a, s] = (0, Y.useState)([]), [c, l] = (0, Y.useState)();
	(0, Y.useEffect)(() => {
		if (!n || !r) return;
		let t = !1;
		return s([]), l(void 0), r({
			descriptor: n,
			query: "",
			...i ? { rootValue: i } : {}
		}).then((e) => {
			t || s(e);
		}, (n) => {
			t || l(`Unable to load ${e.label}: ${String(n)}`);
		}), () => {
			t = !0;
		};
	}, [
		n,
		r,
		i,
		e.label
	]);
	let u = n && r ? a : e.options ?? [], d = e.kind === "array" || n?.multi === !0, f = Array.isArray(e.value) ? e.value.map(String) : e.value == null || e.value === "" ? [] : [String(e.value)], p = e.arrayDisplay === "filter-pills" && f.length === 0 ? u.map((e) => e.value) : f;
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex min-w-0 flex-wrap items-center gap-1",
		children: [
			p.length === 0 && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "text-muted-foreground",
				children: "—"
			}),
			p.map((t, n) => {
				let r = u.find((e) => e.value === t);
				return e.arrayDisplay === "filter-pills" ? /* @__PURE__ */ (0, J.jsx)(Io, {
					label: r?.label ?? t,
					mode: "active"
				}, `${t}-${n}`) : d ? /* @__PURE__ */ (0, J.jsx)(Vo, {
					value: t,
					...r ? { option: r } : {}
				}, `${t}-${n}`) : /* @__PURE__ */ (0, J.jsxs)("span", {
					className: "inline-flex min-w-0 items-center gap-1.5",
					children: [
						/* @__PURE__ */ (0, J.jsx)(o, {
							icon: r?.icon,
							className: "shrink-0 text-muted-foreground"
						}),
						/* @__PURE__ */ (0, J.jsx)("span", { children: r?.label ?? t }),
						r?.description && /* @__PURE__ */ (0, J.jsx)("span", {
							className: "text-muted-foreground",
							children: r.description
						})
					]
				}, t);
			}),
			c && /* @__PURE__ */ (0, J.jsx)("span", {
				role: "alert",
				className: "text-xs text-destructive",
				children: c
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-property-collection.tsx
function FS({ field: e, ctx: t }) {
	let n = {
		...t,
		presentation: !0,
		readOnly: !0,
		layout: {
			...t.layout,
			mode: "stacked"
		}
	};
	return e.kind === "array" ? /* @__PURE__ */ (0, J.jsx)(wv, {
		field: e,
		fieldId: W(t.instancePath, t.idPrefix),
		ctx: n
	}) : e.kind === "object" ? /* @__PURE__ */ (0, J.jsx)(Zv, {
		field: e,
		ctx: n
	}) : /* @__PURE__ */ (0, J.jsx)(Qv, {
		field: e,
		ctx: n
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-property-preview.tsx
function IS({ field: e, fieldId: t, ctx: n }) {
	if (e.kind === "markdown") return /* @__PURE__ */ (0, J.jsx)(NS, {
		text: u(e.value),
		className: q("break-words [&_p]:my-1 [&_h2]:my-2", O[n.size])
	});
	if (e.kind === "string" || e.kind === "date" || e.kind === "number" || e.kind === "boolean") return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex min-w-0 items-center gap-1.5",
		children: [
			e.prefix,
			!n.presentation && e.kind === "boolean" && typeof e.value == "boolean" ? /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [/* @__PURE__ */ (0, J.jsx)("label", {
				htmlFor: t,
				className: "sr-only",
				children: e.label
			}), /* @__PURE__ */ (0, J.jsx)(tv, {
				field: e,
				fieldId: t,
				size: n.size,
				readOnly: !0
			})] }) : /* @__PURE__ */ (0, J.jsx)(Y_, {
				field: e,
				fieldId: t,
				size: n.size
			}),
			e.suffix ?? (e.unit ? /* @__PURE__ */ (0, J.jsx)("span", {
				className: "text-muted-foreground",
				children: e.unit
			}) : null)
		]
	});
	let r = e.kind === "enum" || e.kind === "lookup" || e.kind === "array" && e.arrayDisplay !== "stacked" && e.arrayDisplay !== "list" && (!!ph(e.itemSchema) || !!e.options?.length);
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		id: t,
		"data-jsf-readonly": !0,
		className: q("flex min-w-0 items-center gap-1.5", ue[n.size], O[n.size]),
		children: [
			e.prefix,
			/* @__PURE__ */ (0, J.jsx)("div", {
				className: "min-w-0 flex-1",
				children: r ? /* @__PURE__ */ (0, J.jsx)(PS, {
					field: e,
					ctx: n
				}) : e.kind === "textarea" ? /* @__PURE__ */ (0, J.jsx)("span", {
					className: "whitespace-pre-wrap break-words",
					children: u(e.value) || "—"
				}) : /* @__PURE__ */ (0, J.jsx)(FS, {
					field: e,
					ctx: n
				})
			}),
			e.suffix
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-union.tsx
var LS = 6;
function RS({ field: e, fieldId: t, ctx: n }) {
	let r = Wm(Hm(e.schema));
	return r ? /* @__PURE__ */ (0, J.jsx)(zS, {
		field: e,
		list: r,
		ctx: n
	}) : /* @__PURE__ */ (0, J.jsx)(BS, {
		field: e,
		fieldId: t,
		ctx: n
	});
}
function zS({ field: e, list: t, ctx: n }) {
	let { oneOf: r, anyOf: i, ...a } = e.schema, o = {
		...a,
		...t,
		title: e.label
	}, s = n.render.renderFieldNodes({
		key: e.key,
		prop: o,
		required: e.required,
		value: Gm(e.value),
		onChange: (t) => e.onChange(Km(t)),
		instancePath: n.instancePath
	}, n);
	return /* @__PURE__ */ (0, J.jsx)(J.Fragment, { children: s?.value });
}
function BS({ field: e, fieldId: t, ctx: n }) {
	let r = (0, Y.useRef)(/* @__PURE__ */ new Map()), [i, a] = (0, Y.useState)(null), o = Hm(e.schema), s = i === null ? void 0 : o[i], c = s && (Zm(s, e.value) || qm(s) === "string" && typeof e.value == "string") ? i : o.findIndex((t) => Zm(t, e.value)), l = c < 0 ? 0 : c, u = o[l];
	if (!u) throw Error(`Union ${e.key} has no branch`);
	let { oneOf: d, anyOf: f, ...p } = e.schema, m = {
		...p,
		...u,
		title: e.label,
		...e.nullable ? { anyOf: [u, { type: "null" }] } : {}
	}, h = n.render.renderFieldNodes({
		key: e.key,
		prop: m,
		required: e.required,
		value: e.value,
		onChange: e.onChange,
		instancePath: n.instancePath
	}, n), g = o.every((e) => [
		"string",
		"number",
		"integer",
		"boolean"
	].includes(qm(e)) && Object.keys(e).every((e) => e === "type")), _ = n.readOnly || e.readOnly, v = o.map((e, t) => {
		let n = Jm(e);
		return {
			value: String(t),
			label: n && o.filter((e) => Jm(e) === n).length === 1 ? n : `${n ?? "Variant"} (${qm(e)} ${t + 1})`
		};
	}), y = (t) => {
		let n = Number(t), i = o[n];
		if (!i) throw Error(`Unknown union variant ${n}`);
		c >= 0 && r.current.set(l, e.value), a(n), e.onChange(r.current.has(n) ? r.current.get(n) : Xm(i));
	}, b = v.length <= LS ? /* @__PURE__ */ (0, J.jsx)(uf, {
		value: String(l),
		wrap: !0,
		"aria-label": `${e.label} format`,
		onChange: y,
		options: v.map((e) => ({
			id: e.value,
			label: e.label,
			..._ ? { disabled: _ } : {}
		}))
	}) : /* @__PURE__ */ (0, J.jsx)(gf, {
		id: `${t}-variant`,
		"aria-label": `${e.label} format`,
		value: String(l),
		disabled: _,
		options: v,
		onChange: (e) => y(e.target.value)
	});
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "flex min-w-0 flex-col gap-density-2",
		"data-jsf-union": !0,
		children: [!g && /* @__PURE__ */ (0, J.jsx)("div", {
			className: "min-w-0",
			children: b
		}), qm(u) === "null" ? /* @__PURE__ */ (0, J.jsx)("code", { children: "null" }) : h?.value]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-array-add-menu.tsx
function VS({ owner: e, items: t }) {
	let n = Rp(e), r = t.map((e) => ({
		label: e.label,
		onSelect: e.onSelect
	}));
	return n ? (0, Fe.createPortal)(/* @__PURE__ */ (0, J.jsx)(uo, {
		label: /* @__PURE__ */ (0, J.jsx)("span", {
			className: "sr-only",
			children: "Add"
		}),
		icon: X,
		hideChevron: !0,
		variant: "ghost",
		size: "icon",
		align: "right",
		title: "Add",
		menuLabel: "Add",
		items: r
	}), n) : /* @__PURE__ */ (0, J.jsx)("div", {
		"data-jsf-add-menu": "",
		className: "col-span-full",
		style: { gridColumn: "1 / -1" },
		children: /* @__PURE__ */ (0, J.jsx)(uo, {
			label: "Add",
			icon: X,
			variant: "outline",
			size: "sm",
			align: "left",
			menuLabel: "Add",
			items: r
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-exclusive-key-select.tsx
function HS({ keys: e, value: t, fieldId: n, disabled: r, onChange: i }) {
	let a = e.find((e) => e.key === t)?.label ?? t;
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "w-36 shrink-0",
		children: [/* @__PURE__ */ (0, J.jsx)("label", {
			htmlFor: n,
			className: "sr-only",
			children: a
		}), /* @__PURE__ */ (0, J.jsx)(gf, {
			id: `${n}-key`,
			"aria-label": e.map((e) => e.label).join(" / "),
			value: t,
			disabled: r,
			options: e.map((e) => ({
				value: e.key,
				label: e.label
			})),
			onChange: (e) => i(e.target.value),
			className: "h-7 text-xs font-medium"
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-json.tsx
function US({ field: e, fieldId: t, readOnly: n }) {
	let r = e.nullable && e.value == null ? "" : JSON.stringify(e.value, null, 2) ?? "", [i, a] = (0, Y.useState)({
		source: r,
		text: r
	}), o = i.source === r ? i.text : r, s = i.source === r ? i.error : void 0;
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: "min-w-0",
		children: [/* @__PURE__ */ (0, J.jsx)(po, {
			as: "textarea",
			id: t,
			"aria-label": e.label,
			value: o,
			readOnly: n,
			inputClassName: "min-h-28 font-mono text-xs",
			clearable: e.nullable === !0,
			clearLabel: `Clear ${e.label} to null`,
			onClear: () => {
				e.onChange(null), a({
					source: "",
					text: ""
				});
			},
			"aria-invalid": !!s,
			onChange: (t) => {
				try {
					let n = JSON.parse(t);
					e.onChange(n), a({
						source: JSON.stringify(n, null, 2) ?? "",
						text: t
					});
				} catch (e) {
					a({
						source: r,
						text: t,
						error: e instanceof Error ? e.message : String(e)
					});
				}
			}
		}), s && /* @__PURE__ */ (0, J.jsx)("p", {
			role: "alert",
			className: "text-xs text-destructive",
			children: s
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-nullable.tsx
function WS({ field: e, fieldId: t, ctx: n, clearInHeader: r = !1, children: i }) {
	let a = n.readOnly || e.readOnly === !0, o = e.value === null || e.value === void 0, s = i;
	return e.kind === "boolean" ? s = /* @__PURE__ */ (0, J.jsx)(gf, {
		clearable: !0,
		clearLabel: `Clear ${e.label} to null`,
		onClear: () => e.onChange(null),
		id: t,
		"aria-label": e.label,
		disabled: a,
		className: g[n.size],
		value: o ? "" : String(e.value),
		options: [
			{
				value: "",
				label: ""
			},
			{
				value: "true",
				label: "Yes"
			},
			{
				value: "false",
				label: "No"
			}
		],
		onChange: (t) => e.onChange(t.target.value === "" ? null : t.target.value === "true")
	}) : o && [
		"object",
		"array",
		"string-map"
	].includes(e.kind) && (s = a ? /* @__PURE__ */ (0, J.jsx)("span", {}) : /* @__PURE__ */ (0, J.jsxs)(fe, {
		variant: "outline",
		size: "sm",
		onClick: () => e.onChange(Xm(e.schema)),
		children: ["Add ", e.label]
	})), /* @__PURE__ */ (0, J.jsx)(F, {
		clearable: !r && ce(e, a),
		clearLabel: `Clear ${e.label} to null`,
		onClear: () => e.onChange(null),
		children: s
	});
}
// ../../../clicky-ui/packages/ui/src/layout/PanelFrame.tsx
var GS = {
	default: "",
	danger: "border-l-2 border-l-red-500",
	warning: "border-l-2 border-l-yellow-500",
	success: "border-l-2 border-l-green-500",
	info: "border-l-2 border-l-blue-500"
};
function KS({ header: e, footer: t, tone: n = "default", padded: r = !0, bodyVisible: i = !0, className: a, headerClassName: o, bodyClassName: s, children: c }) {
	return /* @__PURE__ */ (0, J.jsxs)("div", {
		className: q("overflow-hidden rounded-md border border-border bg-card", GS[n], a),
		children: [
			e !== void 0 && /* @__PURE__ */ (0, J.jsx)("div", {
				className: q("flex items-center gap-2 px-density-3 py-density-2", i && "border-b border-border", o),
				children: e
			}),
			i && /* @__PURE__ */ (0, J.jsx)("div", {
				className: q(r && "px-density-3 py-density-2", s),
				children: c
			}),
			i && t !== void 0 && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "border-t border-border px-density-3 py-density-2",
				children: t
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/layout/Section.tsx
function qS({ title: e, summary: t, collapsible: n = !0, defaultOpen: r = !1, open: i, onToggle: a, icon: o, tone: s = "default", className: l, headerClassName: u, bodyClassName: d, children: f }) {
	let p = i !== void 0, [m, h] = (0, Y.useState)(r), g = !n || (p ? i : m);
	function _() {
		let e = !g;
		p || h(e), a?.(e);
	}
	let v = /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [
		n && /* @__PURE__ */ (0, J.jsx)(c, {
			icon: g ? pe : he,
			className: "text-muted-foreground text-xs"
		}),
		o && /* @__PURE__ */ (0, J.jsx)(c, {
			...typeof o == "string" ? { name: o } : { icon: o },
			className: "text-base"
		}),
		/* @__PURE__ */ (0, J.jsx)("span", {
			className: "font-medium text-sm flex-1 truncate",
			children: e
		})
	] });
	return /* @__PURE__ */ (0, J.jsx)(KS, {
		header: /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [n ? /* @__PURE__ */ (0, J.jsx)("button", {
			type: "button",
			onClick: _,
			"aria-expanded": g,
			className: "flex min-w-0 flex-1 items-center gap-2 rounded-sm text-left transition-colors hover:bg-accent/50 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
			children: v
		}) : /* @__PURE__ */ (0, J.jsx)("div", {
			className: "flex min-w-0 flex-1 items-center gap-2",
			children: v
		}), t && /* @__PURE__ */ (0, J.jsx)("span", {
			className: "shrink-0 text-xs text-muted-foreground",
			children: t
		})] }),
		tone: s,
		bodyVisible: g,
		className: q("overflow-visible bg-background", l),
		headerClassName: q("w-full text-left", u),
		...d === void 0 ? {} : { bodyClassName: d },
		children: f
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-section.tsx
function JS({ title: e, actions: t, hasErrors: n, children: r }) {
	let [i, a] = (0, Y.useState)(n);
	return (0, Y.useEffect)(() => {
		n && a(!0);
	}, [n]), /* @__PURE__ */ (0, J.jsx)("div", {
		className: "col-span-full min-w-0",
		children: /* @__PURE__ */ (0, J.jsx)(qS, {
			title: e,
			summary: t,
			open: i,
			onToggle: a,
			children: r
		})
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-render.tsx
function YS(e, t, n) {
	let r = W(t.instancePath, t.idPrefix);
	if (e.nullable && !t.presentation) {
		let i = {
			...e,
			onChange: (t) => e.onChange(e.kind === "date" && t === "" ? null : t)
		};
		return /* @__PURE__ */ (0, J.jsx)(WS, {
			field: e,
			fieldId: r,
			ctx: t,
			clearInHeader: n?.clearInHeader === !0,
			children: XS(i, t)
		});
	}
	return XS(e, t);
}
function XS(e, t) {
	let n = W(t.instancePath, t.idPrefix);
	if (t.presentation && e.kind !== "display" && e.kind !== "link") return /* @__PURE__ */ (0, J.jsx)(IS, {
		field: e,
		fieldId: n,
		ctx: t
	});
	if (e.readOnly && !ZS(e.kind)) return /* @__PURE__ */ (0, J.jsx)(Y_, {
		field: e,
		fieldId: n,
		size: t.size
	});
	let r = t.readOnly || e.readOnly === !0, i = t.size;
	switch (e.kind) {
		case "union": return /* @__PURE__ */ (0, J.jsx)(RS, {
			field: e,
			fieldId: n,
			ctx: t
		});
		case "json": return /* @__PURE__ */ (0, J.jsx)(US, {
			field: e,
			fieldId: n,
			readOnly: r
		});
		case "enum": return /* @__PURE__ */ (0, J.jsx)(nv, {
			field: e,
			fieldId: n,
			readOnly: r,
			size: i
		});
		case "lookup": return /* @__PURE__ */ (0, J.jsx)(rv, {
			field: e,
			fieldId: n,
			readOnly: r,
			size: i,
			...t.rootValue ? { rootValue: t.rootValue } : {}
		});
		case "boolean": return /* @__PURE__ */ (0, J.jsx)(tv, {
			field: e,
			fieldId: n,
			readOnly: r,
			size: i
		});
		case "number": return /* @__PURE__ */ (0, J.jsx)(Z_, {
			field: e,
			fieldId: n,
			readOnly: r,
			size: i
		});
		case "date": return /* @__PURE__ */ (0, J.jsx)(ev, {
			field: e,
			fieldId: n,
			readOnly: r,
			size: i
		});
		case "textarea": return /* @__PURE__ */ (0, J.jsx)(av, {
			field: e,
			fieldId: n,
			readOnly: r,
			size: i
		});
		case "markdown": return /* @__PURE__ */ (0, J.jsx)(ov, {
			field: e,
			fieldId: n,
			readOnly: r,
			size: i
		});
		case "display": return /* @__PURE__ */ (0, J.jsx)(sv, {
			field: e,
			size: i
		});
		case "link": return /* @__PURE__ */ (0, J.jsx)(cv, {
			field: e,
			fieldId: n,
			size: i
		});
		case "array": return /* @__PURE__ */ (0, J.jsx)(wv, {
			field: e,
			fieldId: n,
			ctx: t
		});
		case "object": return /* @__PURE__ */ (0, J.jsx)(Zv, {
			field: e,
			ctx: t
		});
		case "string-map": return /* @__PURE__ */ (0, J.jsx)(Qv, {
			field: e,
			ctx: t
		});
		case "custom": return /* @__PURE__ */ (0, J.jsx)(ey, {
			field: e,
			fieldId: n,
			readOnly: r,
			ctx: t
		});
		default: return /* @__PURE__ */ (0, J.jsx)(X_, {
			field: e,
			fieldId: n,
			readOnly: r,
			size: i
		});
	}
}
function ZS(e) {
	return e === "array" || e === "object" || e === "string-map" || e === "textarea" || e === "markdown" || e === "display" || e === "link" || e === "custom";
}
function QS(e, t, n) {
	let r, i = (e) => !t.debug || (r ??= e, !1), a = e.prop["x-hidden-by"];
	if (e.prop["x-hidden"] && i(a ? `hidden by ${a}` : "hidden by the schema (x-hidden)") || e.prop.const !== void 0 && e.prop["x-hidden"] !== !1 && e.value === e.prop.const && i("hidden: const discriminator")) return null;
	let o = dh({
		...e,
		extensions: t.extensions
	});
	if (t.showTypeIcons && o.labelIcon == null) {
		let e = {
			string: yn,
			textarea: yn,
			markdown: yn,
			date: mt,
			boolean: Dt,
			number: qt,
			array: Mt,
			object: zt,
			"string-map": zt,
			enum: it
		};
		if (o.kind in e) {
			let t = e[o.kind];
			o.labelIcon = /* @__PURE__ */ (0, J.jsx)(t, {});
		}
	}
	let s = o, c = o;
	for (let n of t.pre) {
		if (!s) break;
		c = s, s = n(s, {
			key: e.key,
			prop: e.prop,
			value: e.value,
			...t.rootValue ? { rootValue: t.rootValue } : {},
			...t.onRootChange ? { onRootChange: t.onRootChange } : {},
			...e.parent ? { parent: e.parent } : {}
		});
	}
	let l = !s;
	if (!s) {
		if (i("dropped by a form extension")) return null;
		s = c;
	}
	if (t.hideReadOnlyFields && s.readOnly && i("hidden: read-only") || s.writeOnly && (t.viewOnly || s.readOnly) && i("hidden: write-only") || t.hideEmpty && P(s.value) && i("hidden: empty")) return null;
	let u = s.readOnly === !0;
	t.presentation && (s = {
		...s,
		readOnly: !0
	});
	let d = e.instancePath ?? Xh(t.instancePath, e.key), f = W(d, t.idPrefix), p = Qh(t.errors, d), h = p.length === 0 ? S(s) : void 0, g = p.length > 0 ? p : h ? [{
		instancePath: d,
		...h
	}] : [];
	g.length > 0 && (s = {
		...s,
		invalid: !0
	});
	let _ = s.helpDisplay ?? t.layout.help ?? "inline", v = t.extensions.slots.SectionHeading ?? lm, b = nm({
		fieldKey: e.key,
		fieldLayout: s.layout,
		formMode: t.layout.mode,
		layouts: t.extensions.layouts
	}), x = m(s, t.layout.mode) ? s.label === "" ? null : /* @__PURE__ */ (0, J.jsx)(v, {
		field: s,
		size: t.size,
		helpDisplay: _
	}) : /* @__PURE__ */ (0, J.jsx)(sm, {
		field: s,
		fieldId: f,
		size: t.layout.mode === "properties" ? Uv[t.size] : t.size,
		helpDisplay: _
	}), C = s.layout === "inline" ? "inline" : s.layout === "stack" ? "stacked" : void 0, w = t.layout.mode === "properties" && (s.kind === "object" || s.kind === "string-map"), T = {
		...t,
		...s.disabled ? { readOnly: !0 } : {},
		...u ? { viewOnly: !0 } : {},
		...t.layout.mode === "properties" && !w ? { size: Uv[t.size] } : {},
		instancePath: d,
		...C && !w ? { layout: {
			...t.layout,
			mode: C
		} } : {},
		...t.layout.mode === "properties" && s.kind === "array" ? { layout: {
			...t.layout,
			mode: C ?? "stacked"
		} } : {}
	}, E = n !== void 0 && s.nullable === !0 && !t.presentation && m(s, t.layout.mode) && (n.labelOverride ?? s.label) !== "", D = s, O = E && ce(s, T.readOnly || s.readOnly === !0) ? /* @__PURE__ */ (0, J.jsx)(y, {
		clearLabel: `Clear ${s.label} to null`,
		onClear: () => D.onChange(null)
	}) : null, k = YS(s, T, { clearInHeader: E }), A = {
		fieldId: f,
		...t.rootValue ? { rootValue: t.rootValue } : {},
		...t.onRootChange ? { onRootChange: t.onRootChange } : {},
		instancePath: d
	};
	if ({label: x, value: k} = gv(T.readOnly ? {
		...s,
		readOnly: !0
	} : s, {
		label: x,
		value: k
	}, l ? [] : t.post, A), t.layout.mode === "properties" && !w && s.kind !== "display" && s.kind !== "link" && (k = /* @__PURE__ */ (0, J.jsx)(Yv, {
		field: s,
		fieldId: f,
		ctx: T,
		renderPreview: (e) => gv({
			...e,
			readOnly: !0
		}, {
			label: x,
			value: /* @__PURE__ */ (0, J.jsx)(IS, {
				field: e,
				fieldId: `${f}-preview`,
				ctx: {
					...T,
					idPrefix: `${t.idPrefix ?? ""}-preview`
				}
			})
		}, l ? [] : t.post, A).value,
		renderEditor: (e, t) => gv(e, {
			label: x,
			value: YS(e, t)
		}, l ? [] : t.post, {
			...A,
			...t.rootValue ? { rootValue: t.rootValue } : {},
			...t.onRootChange ? { onRootChange: t.onRootChange } : {}
		}).value
	})), t.debug && s.kind !== "display") {
		let n = s.kind === "custom" ? t.extensions.formats.get(s.format ?? "")?.extension : void 0;
		({label: x, value: k} = u_({
			label: x,
			value: k
		}, {
			fieldKey: e.key,
			instancePath: d,
			field: s,
			prop: e.prop,
			hiddenReason: r,
			...n ? { source: n } : {}
		}));
	}
	return {
		field: s,
		fieldId: f,
		label: x,
		value: k,
		headerClear: O,
		messages: g,
		help: _,
		engine: b
	};
}
function $S(e, t) {
	let n = QS(e, t);
	if (!n) return null;
	let r = t.extensions.slots.FieldError ?? fm;
	return {
		label: n.label,
		value: n.messages.length > 0 ? /* @__PURE__ */ (0, J.jsxs)("div", {
			className: "flex min-w-0 flex-col gap-0.5",
			children: [n.value, /* @__PURE__ */ (0, J.jsx)(r, {
				id: re(n.fieldId),
				errors: n.messages
			})]
		}) : n.value
	};
}
function eC(e, t, n) {
	let r = QS(e, t, { labelOverride: n?.labelOverride });
	if (!r) return null;
	let { engine: i, ...a } = r, { field: s } = a, c = t.extensions.slots.FieldError ?? fm, l = t.extensions.slots.SectionHeading ?? lm;
	if (i) return i.Row(n?.labelOverride === void 0 ? a : {
		...a,
		label: n.labelOverride
	}, t);
	if (t.layout.mode === "properties" && s.kind !== "display") return /* @__PURE__ */ (0, J.jsx)(Xv, {
		...a,
		label: n?.labelOverride ?? a.label,
		ctx: t
	});
	if (s.kind === "object" && s.label !== "" && (s.schema["x-object-display"] ?? t.objectDisplay) === "accordion") {
		let n = e.instancePath ?? Xh(t.instancePath, s.key), r = a.messages.length > 0 || t.errors.some((e) => e.instancePath.startsWith(`${n}/`));
		return /* @__PURE__ */ (0, J.jsxs)(JS, {
			title: a.label,
			actions: a.headerClear,
			hasErrors: r,
			children: [a.messages.length > 0 && /* @__PURE__ */ (0, J.jsx)(c, {
				id: re(a.fieldId),
				errors: a.messages
			}), a.value]
		});
	}
	if (m(s, t.layout.mode)) return /* @__PURE__ */ (0, J.jsxs)(um, {
		heading: n?.labelOverride === void 0 ? a.label : n.labelOverride === "" ? null : /* @__PURE__ */ (0, J.jsx)(l, {
			field: s,
			label: n.labelOverride,
			size: t.size,
			helpDisplay: a.help
		}),
		actions: a.headerClear,
		actionsOwner: e.instancePath ?? Xh(t.instancePath, e.key),
		size: t.size,
		helpDisplay: a.help,
		...im(s) ? { helper: /* @__PURE__ */ (0, J.jsx)(rm, { field: s }) } : {},
		children: [a.messages.length > 0 && /* @__PURE__ */ (0, J.jsx)(c, {
			id: re(a.fieldId),
			errors: a.messages,
			className: "min-w-0 break-words"
		}), a.value]
	});
	if (s.kind === "display") return /* @__PURE__ */ (0, J.jsx)("div", {
		className: "col-span-full",
		children: a.value
	});
	let u = n?.labelNode === void 0 ? n?.labelOverride ? /* @__PURE__ */ (0, J.jsxs)("label", {
		htmlFor: a.fieldId,
		className: q("flex items-center gap-2 whitespace-nowrap font-medium", O[t.size]),
		children: [
			/* @__PURE__ */ (0, J.jsx)(o, {
				icon: s.labelIcon,
				className: "text-[15px] text-muted-foreground"
			}),
			/* @__PURE__ */ (0, J.jsx)("span", { children: n.labelOverride }),
			s.required && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "text-destructive",
				children: "*"
			}),
			s.badge && /* @__PURE__ */ (0, J.jsx)("span", {
				className: "rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide text-muted-foreground",
				children: s.badge
			}),
			a.help === "hover" && im(s) && /* @__PURE__ */ (0, J.jsx)(cm, {
				label: n.labelOverride,
				helper: /* @__PURE__ */ (0, J.jsx)(rm, { field: s })
			})
		]
	}) : a.label : n.labelNode, d = a.messages.length > 0, f = t.extensions.slots.FieldError, p = d && !f ? /* @__PURE__ */ (0, J.jsx)(dm, {
		id: re(a.fieldId),
		errors: a.messages
	}) : void 0;
	return /* @__PURE__ */ (0, J.jsx)(om, {
		layout: t.layout,
		size: t.size,
		label: u,
		value: a.value,
		...s.colSpan === "full" ? { maxWidth: "none" } : {},
		...im(s) && a.help !== "hover" ? { helper: /* @__PURE__ */ (0, J.jsx)(rm, { field: s }) } : {},
		...p ? { error: p } : {},
		...d && f ? { errorElement: /* @__PURE__ */ (0, J.jsx)(c, {
			id: re(a.fieldId),
			errors: a.messages
		}) } : {}
	});
}
function tC(e, t, n, r, i) {
	let a = {
		root: r.rootValue ?? t,
		...r.expressionEvaluator ? { evaluate: r.expressionEvaluator } : {}
	}, { properties: o, required: s } = Oh(e, t, a), c = new Set(i?.hiddenKeys ?? []), l = ny({
		properties: o,
		value: t,
		onChange: n,
		schema: e,
		listenerOptions: a,
		instancePath: r.instancePath
	}), u = Object.entries(o).filter(([e, t]) => !l.emptied.has(e) && !(r.depth > 0 && ty(t))), d = {
		schema: {
			...e,
			properties: o
		},
		value: t,
		onChange: n
	}, m = (i, o) => l.rows.get(i) ?? {
		key: i,
		prop: o,
		required: s.includes(i),
		value: t[i],
		onChange: (r) => n(Nh(e, {
			...a,
			value: t,
			key: i,
			next: r
		})),
		instancePath: Xh(r.instancePath, i),
		parent: d
	}, h = E(K([...u, ...Array.from(l.rows, ([e, t]) => [e, t.prop])], e["x-order"])), g = r.sortMode === "required-first" ? D(h, s) : r.sortMode === "priority" ? f(h, s, t) : h, _ = r.depth === 0 && r.fieldFilter ? g.filter(([e, t]) => Yh({
		key: e,
		prop: t,
		filter: r.fieldFilter,
		value: m(e, t).value
	})) : g, v = r.layout.mode === "stacked" ? I(e["x-columns"]) : 1, y = _.filter(([, e]) => e.type === "array"), b = r.layout.emptyArrays === "add-menu", x = $p(r.layout.mode) && r.layout.mode !== "properties" && y.length > (b ? 0 : 3), S = e["x-object-display"] === "row" && r.layout.mode !== "properties", C = Dv(e)?.filter((e) => !c.has(e) && _.some(([t]) => t === e)), w = C && C.length > 1 ? _.find(([e]) => C.includes(e))?.[0] : void 0, T = _.flatMap(([e, i]) => {
		if (c.has(e)) return [];
		if (C && w !== void 0 && C.includes(e)) {
			if (e !== w) return [];
			let i = C.find((e) => t[e] !== void 0) ?? C[0], a = o[i], s = eC({
				...m(i, a),
				required: !0
			}, r, { labelNode: /* @__PURE__ */ (0, J.jsx)(HS, {
				keys: C.map((e) => ({
					key: e,
					label: o[e]?.title || e
				})),
				value: i,
				fieldId: W(Xh(r.instancePath, i), r.idPrefix),
				disabled: r.readOnly || r.presentation === !0,
				onChange: (e) => n(Ov({
					value: t,
					keys: C,
					next: e
				}))
			}) });
			return s ? [/* @__PURE__ */ (0, J.jsx)("div", {
				className: "contents",
				"data-jsf-exclusive-keys": "",
				children: s
			}, `exclusive:${C.join("|")}`)] : [];
		}
		if (S) return nC(m(e, i), r);
		if (x && i.type === "array" && i["x-hidden"] !== !1 && (!Array.isArray(t[e]) || t[e].length === 0)) return [];
		let a = eC(m(e, i), r);
		if (!a) return [];
		if (v === "auto" || v > 1) {
			let t = rC(i) ? "full" : le(i["x-col-span"], v);
			return [/* @__PURE__ */ (0, J.jsx)("div", {
				style: { gridColumn: t === "full" ? "1 / -1" : `span ${t} / span ${t}` },
				children: a
			}, e)];
		}
		return [/* @__PURE__ */ (0, J.jsx)("div", {
			className: "contents",
			children: a
		}, e)];
	});
	if (x && !S && !r.readOnly && !r.presentation) {
		let e = y.filter(([e, n]) => !c.has(e) && n["x-hidden"] === void 0 && (!Array.isArray(t[e]) || n.maxItems === void 0 || t[e].length < n.maxItems)), i = (e) => {
			let r = o[e];
			if (!r || r.type !== "array") throw Error(`cannot add item to ${e}`);
			let i = t[e];
			if (i !== void 0 && !Array.isArray(i)) throw Error(`${e} is not an array`);
			n({
				...t,
				[e]: [...i ?? [], p(r.items ?? { type: "string" })]
			});
		}, a = e.filter(([e]) => !Array.isArray(t[e]) || t[e].length === 0);
		if (a.length > 0) {
			let e = /* @__PURE__ */ (0, J.jsx)(VS, {
				owner: r.instancePath ?? "",
				items: a.map(([e, t]) => ({
					key: e,
					label: t.title || e,
					onSelect: () => i(e)
				}))
			}, "keyed-array-add");
			b ? T.unshift(e) : T.push(e);
		}
	}
	return T;
}
function nC(e, t) {
	let n = $S(e, t);
	if (!n) return [];
	let r = e.prop.title || e.key, i = e.instancePath ?? Xh(t.instancePath, e.key);
	return [/* @__PURE__ */ (0, J.jsxs)("div", {
		"data-jsf-row-cell": e.key,
		title: r,
		className: "min-w-0 flex-1 basis-0",
		children: [/* @__PURE__ */ (0, J.jsxs)("label", {
			htmlFor: W(i, t.idPrefix),
			className: "sr-only",
			children: [r, e.required ? " (required)" : ""]
		}), n.value]
	}, e.key)];
}
function rC(e) {
	let t = x(e, {}).properties;
	return nh(e) && (e.properties !== void 0 || Object.keys(t).length > 0) || H(e.items) && e["x-array-display"] !== "stacked" ? !0 : e["x-layout"] === "table";
}
var iC = {
	renderFieldNodes: $S,
	renderFieldRow: eC,
	renderObjectFields: tC
};
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-chrome.tsx
function aC({ title: e, size: t }) {
	return /* @__PURE__ */ (0, J.jsx)("h3", {
		className: q("font-semibold", O[t]),
		children: e
	});
}
function oC({ filter: e, size: t }) {
	return /* @__PURE__ */ (0, J.jsxs)("p", {
		className: q("text-muted-foreground", O[t]),
		children: [
			"No fields match “",
			e,
			"”."
		]
	});
}
var sC = [
	{
		value: "xs",
		label: "Extra small"
	},
	{
		value: "sm",
		label: "Small"
	},
	{
		value: "md",
		label: "Medium"
	},
	{
		value: "lg",
		label: "Large"
	},
	{
		value: "xl",
		label: "Extra large"
	}
], cC = [
	{
		value: "schema",
		label: "Schema order"
	},
	{
		value: "required-first",
		label: "Required first"
	},
	{
		value: "priority",
		label: "Required & filled first"
	}
];
function lC({ size: e, layoutMode: t, layoutOptions: n, sortMode: r, fieldFilter: i, filterInMenu: a, onFilterChange: o, onSelectSize: s, onSelectLayout: l, onSelectSort: u, debug: d, onToggleDebug: f }) {
	let p = i.trim() !== "" || d;
	return /* @__PURE__ */ (0, J.jsx)(uo, {
		align: "right",
		menuLabel: "Form display options",
		className: "absolute right-0 top-0 z-10",
		trigger: /* @__PURE__ */ (0, J.jsx)("button", {
			type: "button",
			"aria-label": "Form display options",
			"aria-haspopup": "menu",
			className: q("inline-flex h-7 w-7 items-center justify-center rounded-md transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", p ? "text-primary" : "text-muted-foreground hover:text-foreground"),
			children: /* @__PURE__ */ (0, J.jsx)(c, {
				icon: qe,
				className: "text-lg"
			})
		}),
		children: (c) => /* @__PURE__ */ (0, J.jsxs)(J.Fragment, { children: [
			a && /* @__PURE__ */ (0, J.jsx)("div", {
				className: "px-2 pb-1.5 pt-1",
				children: /* @__PURE__ */ (0, J.jsx)(Mp, {
					value: i,
					onChange: o
				})
			}),
			/* @__PURE__ */ (0, J.jsx)(uC, { title: "Size" }),
			sC.map((t) => /* @__PURE__ */ (0, J.jsx)(dC, {
				label: t.label,
				selected: t.value === e,
				onSelect: () => {
					s(t.value), c();
				}
			}, t.value)),
			/* @__PURE__ */ (0, J.jsx)(uC, { title: "Layout" }),
			n.map((e) => /* @__PURE__ */ (0, J.jsx)(dC, {
				label: e.label,
				selected: e.value === t,
				onSelect: () => {
					l(e.value), c();
				}
			}, e.value)),
			/* @__PURE__ */ (0, J.jsx)(uC, { title: "Sort" }),
			cC.map((e) => /* @__PURE__ */ (0, J.jsx)(dC, {
				label: e.label,
				selected: e.value === r,
				onSelect: () => {
					u(e.value), c();
				}
			}, e.value)),
			/* @__PURE__ */ (0, J.jsx)(uC, { title: "Debug" }),
			/* @__PURE__ */ (0, J.jsx)(dC, {
				role: "menuitemcheckbox",
				label: "Show hidden fields & sources",
				selected: d,
				onSelect: () => {
					f(), c();
				}
			})
		] })
	});
}
function uC({ title: e }) {
	return /* @__PURE__ */ (0, J.jsx)("div", {
		className: "px-3 pb-1 pt-2 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground",
		children: e
	});
}
function dC({ label: e, selected: t, onSelect: n, role: r = "menuitemradio" }) {
	return /* @__PURE__ */ (0, J.jsxs)("button", {
		type: "button",
		role: r,
		"aria-checked": t,
		className: "flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs text-popover-foreground hover:bg-accent hover:text-accent-foreground",
		onClick: n,
		children: [/* @__PURE__ */ (0, J.jsx)("span", {
			className: "flex w-3.5 shrink-0 justify-center",
			children: t && /* @__PURE__ */ (0, J.jsx)(c, {
				icon: bt,
				className: "text-xs"
			})
		}), e]
	});
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-preferences.ts
var fC = [
	{
		value: "stacked",
		label: "Stacked"
	},
	{
		value: "inline",
		label: "Inline"
	},
	{
		value: "properties",
		label: "Properties"
	}
], pC = "clicky-ui-json-schema-form-preferences", mC = [
	"xs",
	"sm",
	"md",
	"lg",
	"xl"
];
function hC(e) {
	return typeof e == "string" && mC.includes(e);
}
function gC(e, t) {
	return e === "stacked" || e === "inline" || e === "properties" || typeof e == "string" && t.includes(e);
}
function _C(e) {
	return e === "schema" || e === "required-first" || e === "priority";
}
function vC(e, t = []) {
	let n;
	try {
		n = localStorage.getItem(e);
	} catch {
		return {};
	}
	if (!n) return {};
	let r;
	try {
		r = JSON.parse(n);
	} catch {
		return {};
	}
	if (typeof r != "object" || !r) return {};
	let i = r, a = {};
	return hC(i.size) && (a.size = i.size), gC(i.layoutMode, t) && (a.layoutMode = i.layoutMode), _C(i.sortMode) && (a.sortMode = i.sortMode), typeof i.debug == "boolean" && (a.debug = i.debug), a;
}
function yC(e, t) {
	try {
		localStorage.setItem(e, JSON.stringify(t));
	} catch {}
}
// ../../../clicky-ui/packages/ui/src/components/JsonSchemaForm.tsx
var bC = "40ch", xC = "600px", SC = [];
function CC(e, t, n, r) {
	let i = e?.mode ?? (t ? "inline" : "stacked"), a = n ?? i;
	tm(a, r);
	let o = e?.valueMaxWidth ?? xC, s = {
		...e?.help ? { help: e.help } : {},
		...e?.emptyArrays ? { emptyArrays: e.emptyArrays } : {}
	};
	return a === "properties" ? {
		...e,
		mode: a,
		...s
	} : $p(a) ? a === "inline" ? {
		mode: "inline",
		labelMaxWidth: e?.labelMaxWidth ?? bC,
		valueMaxWidth: o,
		...s
	} : {
		mode: "stacked",
		valueMaxWidth: o,
		...s
	} : {
		...e,
		mode: a,
		valueMaxWidth: o,
		...s
	};
}
function wC({ extensionLoader: e, extensions: t, schema: n, ...r }) {
	let i = Ip(n["x-clicky-extensions"], e), a = (0, Y.useMemo)(() => Fp(n), [n]), o = (0, Y.useMemo)(() => i === void 0 || i.length === 0 ? t : [...t ?? [], ...i], [t, i]);
	return i === void 0 ? /* @__PURE__ */ (0, J.jsx)(de, { label: "Loading form extensions" }) : /* @__PURE__ */ (0, J.jsx)(TC, {
		schema: a,
		...r,
		...o ? { extensions: o } : {}
	});
}
function TC({ showTypeIcons: e = !1, objectDisplay: t = "flat", schema: n, value: r, onChange: i, rootValue: a, onRootChange: o, instancePath: s = "", errors: c = [], readOnly: l = !1, inline: u = !1, layout: f, size: p = "md", idPrefix: m, hideReadOnlyFields: h = !1, hideEmpty: g = !1, autoSave: _ = !1, hiddenKeys: v, requiredFirst: y = !1, title: b, pre: x, post: S, extensions: C = SC, showPreferencesMenu: w = !0, showFilter: T, persistPreferences: E = !0, preferencesStorageKey: D = pC, lookupFetcher: O, connectionApiBase: k, expressionEvaluator: A, applyDefaults: j = !0 }) {
	let M = (0, Y.useMemo)(() => Qp(C), [C]), { slots: N } = M, P = N.Title ?? aC, F = N.PreferencesMenu ?? lC, L = N.FieldFilter ?? Mp, R = N.ErrorSummary ?? pm, z = N.NoMatches ?? oC, B = (0, Y.useMemo)(() => [...fC, ...[...M.layouts.keys()].map((e) => ({
		value: e,
		label: e
	}))], [M]), [V, H] = (0, Y.useState)(() => w && E ? vC(D, [...M.layouts.keys()]) : {}), [U, ee] = (0, Y.useState)(""), te = V.size ?? p, W = CC(f, u, V.layoutMode, M.layouts), re = T ?? (w && W.mode === "properties"), ie = V.sortMode ?? (y ? "required-first" : "schema"), ae = (e) => {
		H(e), E && yC(D, e);
	}, oe = (0, Y.useMemo)(() => bg(n), [n]), se = (0, Y.useMemo)(() => j ? mh(oe, r) : r, [
		j,
		oe,
		r
	]), K = (0, Y.useMemo)(() => [...x ?? [], ...M.pre], [x, M]), ce = (0, Y.useMemo)(() => [...S ?? [], ...M.post], [S, M]), le = (0, Y.useMemo)(() => $h({
		schema: oe,
		value: se,
		errors: c,
		viewOnly: l,
		hideReadOnlyFields: h,
		hideEmpty: g,
		pre: K,
		extensions: M,
		rootValue: a ?? se,
		instancePath: s,
		...v ? { hiddenKeys: v } : {},
		...U.trim() ? { fieldFilter: U.trim() } : {},
		...A ? { expressionEvaluator: A } : {}
	}), [
		se,
		c,
		A,
		M,
		U,
		v,
		g,
		h,
		s,
		K,
		l,
		oe,
		a
	]);
	(0, Y.useEffect)(() => {
		!l && se !== r && i(se);
	}, [
		se,
		i,
		l,
		r
	]);
	let ue = {
		showTypeIcons: e,
		objectDisplay: t,
		readOnly: l,
		viewOnly: l,
		hideReadOnlyFields: h,
		hideEmpty: g,
		autoSave: _,
		layout: W,
		size: te,
		sortMode: ie,
		pre: K,
		post: ce,
		extensions: M,
		rootValue: a ?? se,
		onRootChange: o ?? i,
		instancePath: s,
		errors: c,
		depth: 0,
		render: iC,
		...A ? { expressionEvaluator: A } : {},
		...m ? { idPrefix: m } : {},
		...U.trim() ? { fieldFilter: U.trim() } : {},
		...V.debug ? { debug: !0 } : {}
	}, de = typeof oe["x-discriminator"] == "string" ? oe["x-discriminator"] : void 0, fe = de != null && (se[de] == null || se[de] === ""), pe = de ? null : tC(oe, se, i, ue, v ? { hiddenKeys: v } : void 0), me = pe != null && pe.length === 0 && U.trim() !== "";
	return /* @__PURE__ */ (0, J.jsx)(mg, {
		handlers: M.lookups,
		...O ? { fetcher: O } : {},
		children: /* @__PURE__ */ (0, J.jsx)(hg.Provider, {
			value: k,
			children: /* @__PURE__ */ (0, J.jsx)(Wp.Provider, {
				value: M.layouts,
				children: /* @__PURE__ */ (0, J.jsxs)("div", {
					"data-json-schema-form": !0,
					className: q("relative flex flex-col", d[te]),
					children: [
						w && !fe && /* @__PURE__ */ (0, J.jsx)(F, {
							size: te,
							layoutMode: W.mode,
							layoutOptions: B,
							sortMode: ie,
							fieldFilter: U,
							filterInMenu: !re,
							onFilterChange: ee,
							onSelectSize: (e) => ae({
								...V,
								size: e
							}),
							onSelectLayout: (e) => ae({
								...V,
								layoutMode: e
							}),
							onSelectSort: (e) => ae({
								...V,
								sortMode: e
							}),
							debug: V.debug === !0,
							onToggleDebug: () => ae({
								...V,
								debug: !V.debug
							})
						}),
						b && /* @__PURE__ */ (0, J.jsx)(P, {
							title: b,
							size: te
						}),
						re && !fe && /* @__PURE__ */ (0, J.jsx)("div", {
							className: "max-w-md pr-9",
							children: /* @__PURE__ */ (0, J.jsx)(L, {
								value: U,
								onChange: ee
							})
						}),
						le.length > 0 && /* @__PURE__ */ (0, J.jsx)(R, { errors: le }),
						/* @__PURE__ */ (0, J.jsx)(am, {
							layout: W,
							size: te,
							columns: I(oe["x-columns"]),
							columnMinWidth: G(oe["x-column-min-width"], ne),
							...typeof oe["x-columns-max-width"] == "string" ? { columnsMaxWidth: oe["x-columns-max-width"] } : {},
							...typeof oe["x-classes"] == "string" ? { className: oe["x-classes"] } : {},
							children: de ? /* @__PURE__ */ (0, J.jsx)(_g, {
								schema: oe,
								value: se,
								onChange: i,
								ctx: ue,
								discKey: de,
								...v ? { hiddenKeys: v } : {}
							}) : pe
						}),
						me && /* @__PURE__ */ (0, J.jsx)(z, {
							filter: U.trim(),
							size: te
						})
					]
				})
			})
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/TabButton.tsx
var EC = {
	pill: {
		base: "rounded-md",
		active: "bg-primary text-primary-foreground",
		inactive: "text-muted-foreground hover:bg-accent hover:text-foreground"
	},
	underline: {
		base: "rounded-none border-b-2 -mb-px",
		active: "border-primary text-foreground font-medium",
		inactive: "border-transparent text-muted-foreground hover:text-foreground"
	}
};
function DC({ active: e, onClick: t, label: n, icon: r, count: i, countColor: a = "bg-muted-foreground", variant: o = "pill", className: s }) {
	let l = EC[o];
	return /* @__PURE__ */ (0, J.jsxs)("button", {
		type: "button",
		role: "tab",
		"aria-selected": e,
		onClick: t,
		className: q("inline-flex items-center gap-1.5 px-density-3 py-density-1.5 text-sm transition-colors", l.base, e ? l.active : l.inactive, s),
		children: [
			r && /* @__PURE__ */ (0, J.jsx)(c, { ...typeof r == "string" ? { name: r } : { icon: r } }),
			/* @__PURE__ */ (0, J.jsx)("span", { children: n }),
			i !== void 0 && i > 0 && /* @__PURE__ */ (0, J.jsx)("span", {
				className: q("inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded-full text-[10px] font-bold text-white", a),
				children: i
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/layout/Tabs.tsx
function OC({ tabs: e, value: t, onChange: n, variant: r = "underline", className: i }) {
	return /* @__PURE__ */ (0, J.jsx)("div", {
		role: "tablist",
		className: q("flex items-center gap-density-1", r === "underline" && "border-b border-border", i),
		children: e.map((e) => /* @__PURE__ */ (0, J.jsx)(DC, {
			active: e.id === t,
			onClick: () => {
				e.disabled || n(e.id);
			},
			label: e.label,
			variant: r,
			...e.icon === void 0 ? {} : { icon: e.icon },
			...e.count === void 0 ? {} : { count: e.count },
			...e.countColor === void 0 ? {} : { countColor: e.countColor },
			...e.disabled ? { className: "opacity-50 pointer-events-none" } : {}
		}, e.id))
	});
}
// ../../../clicky-ui/packages/ui/src/lib/duration.ts
var kC = [
	"ns",
	"us",
	"µs",
	"μs",
	"ms",
	"s",
	"m",
	"h"
], AC = {
	ns: 1 / 1e6,
	us: 1 / 1e3,
	µs: 1 / 1e3,
	μs: 1 / 1e3,
	ms: 1,
	s: 1e3,
	m: 6e4,
	h: 36e5
};
function jC(e, t) {
	let n = t, r = 0;
	for (; PC(e[n]);) r++, n++;
	let i = 0;
	if (e[n] === ".") for (n++; PC(e[n]);) i++, n++;
	return r === 0 && i === 0 ? null : {
		value: Number(e.slice(t, n)),
		next: n
	};
}
function MC(e, t) {
	for (let n of kC) if (e.startsWith(n, t)) return n;
	return null;
}
function NC(e) {
	let t = 0, n = 0;
	for (; t < e.length;) {
		let r = jC(e, t), i = r && MC(e, r.next);
		if (!r || !i) return null;
		n += r.value * AC[i], t = r.next + i.length;
	}
	return t === 0 ? null : n;
}
function PC(e) {
	return e !== void 0 && e >= "0" && e <= "9";
}
export { sf as $, jt as $t, Hp as A, Ye as An, _a as At, _f as B, En as Bt, A_ as C, nt as Cn, Ba as Ct, fg as D, Ze as Dn, Di as Dt, bg as E, Qe as En, ii as Et, $f as F, Ta as Ft, Af as G, Vt as Gt, kf as H, Wt as Ht, ep as I, Ma as It, Cf as J, Lt as Jt, Ff as K, zt as Kt, tp as L, ti as Lt, Ap as M, Ke as Mn, Ca as Mt, Cp as N, Ei as Nt, mh as O, X as On, na as Ot, cp as P, Ti as Pt, uf as Q, Mt as Qt, rp as R, Pa as Rt, F_ as S, rt as Sn, Ua as St, O_ as T, $e as Tn, pa as Tt, Pf as U, Ut, vf as V, Gt as Vt, Nf as W, Ht as Wt, pf as X, Ft as Xt, gf as Y, It as Yt, df as Z, Nt as Zt, hy as _, lt as _n, po as _t, DC as a, wt as an, cd as at, I_ as b, at as bn, no as bt, NS as c, bt as cn, qo as ct, Zx as d, vt as dn, ko as dt, At as en, Xd as et, rb as f, gt as fn, wo as ft, by as g, ut as gn, Mo as gt, Py as h, ft as hn, Do as ht, OC as i, Tt as in, od as it, jp as j, Je as jn, xa as jt, ym as k, Xe as kn, Oi as kt, tS as l, xt as ln, Jo as lt, qy as m, mt as mn, Oo as mt, NC as n, Ot as nn, Ad as nt, wC as o, Ct as on, Ds as ot, Qy as p, _t as pn, To as pt, yf as q, Rt as qt, jC as r, Et as rn, sd as rt, qS as s, St as sn, ss as st, AC as t, kt as tn, Hd as tt, Qx as u, yt as un, Io as ut, uy as v, ct as vn, uo as vt, E_ as w, et as wn, Ha as wt, R_ as x, it as xn, ro as xt, Pv as y, ot as yn, lo as yt, ap as z, On as zt };
