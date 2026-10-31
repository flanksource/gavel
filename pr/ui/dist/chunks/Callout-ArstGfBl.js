import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./react-DHvKIBht.js";
import { o as r } from "./button-G86_a8z9.js";
// ../../../clicky-ui/packages/ui/src/lib/collections.ts
var i = /* @__PURE__ */ e(t(), 1);
n();
function a(e) {
	return typeof e == "object" && !!e && !Array.isArray(e);
}
function o(e, t, n) {
	return e.map((e, r) => r === t ? n : e);
}
function s(e, t) {
	return e.filter((e, n) => n !== t);
}
function c(e, t, n) {
	if (n < 0 || n >= e.length) return e;
	let r = [...e], [i] = r.splice(t, 1);
	return r.splice(n, 0, i), r;
}
function l(e, t, n) {
	let r = e[t], i = n ? n(r) : Array.isArray(r) ? [...r] : a(r) ? { ...r } : r, o = [...e];
	return o.splice(t + 1, 0, i), o;
}
// ../../../clicky-ui/packages/ui/src/icons/components/UiWarningTriangle.tsx
var u = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...a }) => {
	let o = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#d97706",
		...r
	};
	return /* @__PURE__ */ (0, i.jsx)("svg", {
		xmlns: "http://www.w3.org/2000/svg",
		width: e,
		height: e,
		viewBox: "0 0 256 256",
		role: n ? "img" : "presentation",
		"aria-label": n,
		"aria-hidden": !n || void 0,
		className: t,
		style: o,
		...a,
		children: /* @__PURE__ */ (0, i.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...a }) => {
	let o = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#d97706",
		...r
	};
	return /* @__PURE__ */ (0, i.jsx)("svg", {
		xmlns: "http://www.w3.org/2000/svg",
		width: e,
		height: e,
		viewBox: "0 0 256 256",
		role: n ? "img" : "presentation",
		"aria-label": n,
		"aria-hidden": !n || void 0,
		className: t,
		style: o,
		...a,
		children: /* @__PURE__ */ (0, i.jsx)("path", {
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
var d = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...a }) => {
	let o = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#0ea5e9",
		...r
	};
	return /* @__PURE__ */ (0, i.jsx)("svg", {
		xmlns: "http://www.w3.org/2000/svg",
		width: e,
		height: e,
		viewBox: "0 0 256 256",
		role: n ? "img" : "presentation",
		"aria-label": n,
		"aria-hidden": !n || void 0,
		className: t,
		style: o,
		...a,
		children: /* @__PURE__ */ (0, i.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...a }) => {
	let o = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#0ea5e9",
		...r
	};
	return /* @__PURE__ */ (0, i.jsx)("svg", {
		xmlns: "http://www.w3.org/2000/svg",
		width: e,
		height: e,
		viewBox: "0 0 256 256",
		role: n ? "img" : "presentation",
		"aria-label": n,
		"aria-hidden": !n || void 0,
		className: t,
		style: o,
		...a,
		children: /* @__PURE__ */ (0, i.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiLightbulb.tsx
var f = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, i.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, i.jsx)("path", {
		fill: "currentColor",
		d: "M174 232a6 6 0 0 1-6 6H88a6 6 0 0 1 0-12h80a6 6 0 0 1 6 6m40-128a85.56 85.56 0 0 1-32.88 67.64A18.23 18.23 0 0 0 174 186v6a14 14 0 0 1-14 14H96a14 14 0 0 1-14-14v-6a18 18 0 0 0-7-14.23a85.59 85.59 0 0 1-33-67.24C41.74 57.91 79.39 19.12 125.93 18A86 86 0 0 1 214 104m-12 0a74 74 0 0 0-75.79-74C86.17 31 53.78 64.34 54 104.42a73.67 73.67 0 0 0 28.4 57.87A29.92 29.92 0 0 1 94 186v6a2 2 0 0 0 2 2h64a2 2 0 0 0 2-2v-6a30.18 30.18 0 0 1 11.7-23.78A73.59 73.59 0 0 0 202 104m-20.08-9A55.58 55.58 0 0 0 137 50.08a6 6 0 1 0-2 11.84c17.38 2.92 32.13 17.68 35.08 35.08a6 6 0 0 0 5.91 5a7 7 0 0 0 1-.08a6 6 0 0 0 4.93-6.92"
	})
}), {
	__source: "ph:lightbulb-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "lightbulb",
	displayName: "UiLightbulb"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, i.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, i.jsx)("path", {
		fill: "currentColor",
		d: "M176 232a8 8 0 0 1-8 8H88a8 8 0 0 1 0-16h80a8 8 0 0 1 8 8m40-128a87.55 87.55 0 0 1-33.64 69.21A16.24 16.24 0 0 0 176 186v6a16 16 0 0 1-16 16H96a16 16 0 0 1-16-16v-6a16 16 0 0 0-6.23-12.66A87.59 87.59 0 0 1 40 104.49C39.74 56.83 78.26 17.14 125.88 16A88 88 0 0 1 216 104m-32.11-9.34a57.6 57.6 0 0 0-46.56-46.55a8 8 0 0 0-2.66 15.78c16.57 2.79 30.63 16.85 33.44 33.45A8 8 0 0 0 176 104a9 9 0 0 0 1.35-.11a8 8 0 0 0 6.54-9.23"
	})
}), {
	__source: "ph:lightbulb-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "lightbulb",
	displayName: "UiLightbulbFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiStop.tsx
var p = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, i.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, i.jsx)("g", {
		fill: "none",
		children: /* @__PURE__ */ (0, i.jsx)("rect", {
			x: "2.5",
			y: "2.5",
			width: "11",
			height: "11",
			rx: "1.5",
			fill: "#FFF7F7",
			stroke: "#DB3B4B"
		})
	})
}), {
	__source: "jb-expui-run:stop",
	__viewBox: "0 0 16 16",
	__group: "runtime-process",
	__consumerName: "stop",
	displayName: "UiStop"
}), m = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, i.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, i.jsx)("path", {
		fill: "currentColor",
		d: "M154.64 26.61a6 6 0 0 0-6.32.65L77.94 82H32a14 14 0 0 0-14 14v64a14 14 0 0 0 14 14h45.94l70.38 54.74A6 6 0 0 0 158 224V32a6 6 0 0 0-3.36-5.39M30 160V96a2 2 0 0 1 2-2h42v68H32a2 2 0 0 1-2-2m116 51.73l-60-46.66V90.93l60-46.66Zm50.53-108.85a38 38 0 0 1 0 50.24a6 6 0 1 1-9-7.94a26 26 0 0 0 0-34.37a6 6 0 0 1 9-7.93M246 128a77.86 77.86 0 0 1-19.86 52a6 6 0 1 1-8.94-8a66 66 0 0 0 0-88a6 6 0 1 1 8.94-8A77.86 77.86 0 0 1 246 128"
	})
}), {
	__source: "ph:speaker-high-light",
	__viewBox: "0 0 256 256",
	__group: "media",
	__consumerName: "speaker",
	displayName: "UiSpeaker"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, i.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, i.jsx)("path", {
		fill: "currentColor",
		d: "M160 32.25v191.44a8.29 8.29 0 0 1-3.91 7.18a8 8 0 0 1-9-.56l-65.57-51a4 4 0 0 1-1.52-3.15V79.84a4 4 0 0 1 1.55-3.15l65.57-51a8 8 0 0 1 10 .16a8.27 8.27 0 0 1 2.88 6.4M60 80H32a16 16 0 0 0-16 16v64a16 16 0 0 0 16 16h28a4 4 0 0 0 4-4V84a4 4 0 0 0-4-4m126.77 20.84a8 8 0 0 0-.72 11.3a24 24 0 0 1 0 31.72a8 8 0 1 0 12 10.58a40 40 0 0 0 0-52.88a8 8 0 0 0-11.31-.72Zm40.89-26.17a8 8 0 1 0-11.92 10.66a64 64 0 0 1 0 85.34a8 8 0 1 0 11.92 10.66a80 80 0 0 0 0-106.66"
	})
}), {
	__source: "ph:speaker-high-fill",
	__viewBox: "0 0 256 256",
	__group: "media",
	__consumerName: "speaker",
	displayName: "UiSpeakerFilled"
});
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-size.ts
var h = {
	xs: "h-7 px-1.5 text-xs",
	sm: "h-8 px-2 text-xs",
	md: "h-9 px-2 text-sm",
	lg: "h-10 px-3 text-base",
	xl: "h-11 px-3.5 text-lg"
}, g = {
	xs: "text-xs",
	sm: "text-xs",
	md: "text-sm",
	lg: "text-base",
	xl: "text-lg"
}, _ = {
	xs: "h-7",
	sm: "h-8",
	md: "h-9",
	lg: "h-10",
	xl: "h-11"
}, v = {
	xs: "min-h-7",
	sm: "min-h-8",
	md: "min-h-9",
	lg: "min-h-10",
	xl: "min-h-11"
}, y = {
	xs: "gap-1",
	sm: "gap-1.5",
	md: "gap-2.5",
	lg: "gap-4",
	xl: "gap-4"
}, ee = {
	xs: "gap-0.5",
	sm: "gap-1",
	md: "gap-1.5",
	lg: "gap-2",
	xl: "gap-2"
}, te = {
	xs: "gap-1",
	sm: "gap-1",
	md: "gap-1.5",
	lg: "gap-2",
	xl: "gap-2"
};
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-conditionals.ts
function b(e, t) {
	let n = !1, r = !0;
	if ("const" in e && (n = !0, t !== e.const && (r = !1)), e.enum !== void 0 && (n = !0, e.enum.some((e) => e === t) || (r = !1)), e.not !== void 0) {
		let i = b(e.not, t);
		if (i === void 0) return;
		n = !0, i && (r = !1);
	}
	return (e.properties !== void 0 || e.required !== void 0) && (n = !0, a(t) && x(e, t) || (r = !1)), n ? r : void 0;
}
function x(e, t) {
	if (!e) return !1;
	for (let n of e.required ?? []) if (!(n in t)) return !1;
	for (let [n, r] of Object.entries(e.properties ?? {})) if (b(r, t[n]) !== !0) return !1;
	return !0;
}
function S(e, t, n) {
	if (e) {
		for (let [n, r] of Object.entries(e.properties ?? {})) t[n] = r;
		for (let t of e.required ?? []) n.add(t);
	}
}
function C(e, t) {
	let n = { ...e.properties }, r = new Set(e.required ?? []);
	for (let i of e.allOf ?? []) i.if === void 0 && i.then === void 0 ? S(i, n, r) : S(x(i.if, t) ? i.then : i.else, n, r);
	return {
		properties: n,
		required: [...r]
	};
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-utils.ts
var w = "w-full min-w-0 rounded-md border border-input bg-background text-foreground outline-none placeholder:text-placeholder disabled:cursor-not-allowed disabled:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring";
function T(e) {
	return r(w, h[e]);
}
function E(e) {
	return e === "" ? "_" : e.replace(/[^A-Za-z0-9]/g, (e) => `_${e.charCodeAt(0).toString(16)}_`);
}
function ne(e, t) {
	let n = (e === "" ? [] : e.replace(/^\//, "").split("/")).map((e) => E(e.replaceAll("~1", "/").replaceAll("~0", "~"))).join("-");
	return t ? `jsf-${E(t)}-${n}` : `jsf-${n}`;
}
function D(e) {
	return `${e}-error`;
}
function O(e, t) {
	return {
		...e.required ? { "aria-required": !0 } : {},
		...e.invalid ? {
			"aria-invalid": !0,
			"aria-describedby": D(t)
		} : {}
	};
}
function k(e, t) {
	return {
		...e.required ? { ariaRequired: !0 } : {},
		...e.invalid ? {
			invalid: !0,
			describedBy: D(t)
		} : {}
	};
}
function A(e) {
	return e.default != null && typeof e.default != "object" ? String(e.default) : "";
}
function j(e) {
	return e == null || typeof e == "object" ? "" : String(e);
}
function M(e) {
	return Array.isArray(e) ? e.filter((e) => typeof e == "string") : [];
}
function N(e, t) {
	return t !== "properties" && (e.kind === "object" || e.layout === "table" || e.arrayDisplay === "accordion" || e.arrayDisplay === "cards" || e.kind === "array" && e.arrayDisplay !== "stacked" && P(e.itemSchema));
}
function P(e) {
	return e ? Object.keys(C(e, {}).properties).length > 0 : !1;
}
function F(e) {
	if (e === "auto") return "auto";
	let t = typeof e == "number" ? e : Number(e);
	return !Number.isFinite(t) || t < 1 ? 1 : Math.min(Math.trunc(t), 12);
}
function I(e, t) {
	if (e === "full") return "full";
	let n = typeof e == "number" ? e : Number(e);
	return !Number.isFinite(n) || n < 1 ? 1 : Math.min(Math.trunc(n), t === "auto" ? 12 : t);
}
var L = "15rem";
function R(e, t) {
	return typeof e == "string" && e.trim() ? e.trim() : t;
}
function z(e, t) {
	return !t || e.some((e) => e.value === t) ? e : [{
		value: t,
		label: t
	}, ...e];
}
function B(e, t, n) {
	let r = new Set(t.filter((e) => e !== n));
	return e.filter((e) => !r.has(e.value));
}
function V(e, t) {
	if (!Array.isArray(t) || t.length === 0) return e;
	let n = new Map(t.filter((e) => typeof e == "string").map((e, t) => [e, t])), r = e.filter(([e]) => n.has(e)).sort((e, t) => n.get(e[0]) - n.get(t[0])), i = e.filter(([e]) => !n.has(e));
	return [...r, ...i];
}
function H(e) {
	let t = (e) => {
		let t = e["x-clicky-order"];
		return typeof t == "number" ? t : Infinity;
	};
	return e.map((e, n) => ({
		entry: e,
		index: n,
		rank: t(e[1])
	})).sort((e, t) => e.rank - t.rank || e.index - t.index).map((e) => e.entry);
}
function U(e, t) {
	let n = new Set(t), r = e.filter(([e]) => n.has(e)), i = e.filter(([e]) => !n.has(e));
	return [...r, ...i];
}
function W(e) {
	return e == null || e === "" ? !0 : Array.isArray(e) ? e.every(W) : typeof e == "object" && Object.values(e).every(W);
}
function G(e, t, n) {
	let r = new Set(t), i = (e) => (r.has(e) ? 2 : 0) + +!W(n[e]);
	return e.map((e, t) => ({
		entry: e,
		index: t,
		score: i(e[0])
	})).sort((e, t) => t.score - e.score || e.index - t.index).map((e) => e.entry);
}
function K(e) {
	let t = e.value, n = t == null || t === "";
	if (e.required && n) return "Required";
	if (e.kind === "number" && typeof e.minimum == "number" && typeof t == "number" && t < e.minimum) return `Must be ≥ ${e.minimum}`;
	if (e.kind === "enum" && !e.allowCustomValue && !n && !(e.options ?? []).some((e) => e.value === String(t))) return "Unknown value (allowed)";
}
function q(e, t) {
	return e.maxItems === void 0 || t < e.maxItems;
}
function J(e, t) {
	return e.minItems === void 0 || t > e.minItems;
}
function Y(e) {
	if (e.default !== void 0) return e.default;
	let t = Array.isArray(e.type) ? e.type[0] : e.type;
	return t === "array" ? [] : t === "object" ? {} : t !== "boolean" && "";
}
// ../../../clicky-ui/packages/ui/src/data/callout-tones.ts
var X = [
	"note",
	"tip",
	"important",
	"warning",
	"caution"
], Z = [...X, "default"], re = {
	note: {
		accent: "border-l-[3px] border-l-blue-500 bg-blue-500/5 rounded-r",
		emphasis: "border-2 border-blue-500 bg-blue-500/10 rounded",
		subtle: "border-blue-500/30 bg-blue-500/5",
		badge: "bg-blue-600 text-white",
		label: "text-blue-700 dark:text-blue-300",
		source: "text-blue-600/70 dark:text-blue-300/70",
		icon: "text-blue-600 dark:text-blue-400",
		title: "text-blue-900 dark:text-blue-100",
		content: "text-blue-950/90 dark:text-blue-100/90"
	},
	tip: {
		accent: "border-l-[3px] border-l-emerald-500 bg-emerald-500/5 rounded-r",
		emphasis: "border-2 border-emerald-500 bg-emerald-500/10 rounded",
		subtle: "border-emerald-500/30 bg-emerald-500/5",
		badge: "bg-emerald-600 text-white",
		label: "text-emerald-700 dark:text-emerald-300",
		source: "text-emerald-600/70 dark:text-emerald-300/70",
		icon: "text-emerald-600 dark:text-emerald-400",
		title: "text-emerald-900 dark:text-emerald-100",
		content: "text-emerald-950/90 dark:text-emerald-100/90"
	},
	important: {
		accent: "border-l-[3px] border-l-purple-500 bg-purple-500/5 rounded-r",
		emphasis: "border-2 border-purple-500 bg-purple-500/10 rounded",
		subtle: "border-purple-500/30 bg-purple-500/5",
		badge: "bg-purple-600 text-white",
		label: "text-purple-700 dark:text-purple-300",
		source: "text-purple-600/70 dark:text-purple-300/70",
		icon: "text-purple-600 dark:text-purple-400",
		title: "text-purple-900 dark:text-purple-100",
		content: "text-purple-950/90 dark:text-purple-100/90"
	},
	warning: {
		accent: "border-l-[3px] border-l-amber-500 bg-amber-500/10 rounded-r",
		emphasis: "border-2 border-amber-500 bg-amber-500/15 rounded",
		subtle: "border-amber-500/40 bg-amber-500/10",
		badge: "bg-amber-600 text-white",
		label: "text-amber-700 dark:text-amber-300",
		source: "text-amber-600/70 dark:text-amber-300/70",
		icon: "text-amber-600 dark:text-amber-400",
		title: "text-amber-900 dark:text-amber-100",
		content: "text-amber-950/90 dark:text-amber-100/90"
	},
	caution: {
		accent: "border-l-[3px] border-l-red-500 bg-red-500/5 rounded-r",
		emphasis: "border-2 border-red-500 bg-red-500/10 rounded",
		subtle: "border-red-500/40 bg-red-500/10",
		badge: "bg-red-600 text-white",
		label: "text-red-700 dark:text-red-300",
		source: "text-red-600/70 dark:text-red-300/70",
		icon: "text-red-600 dark:text-red-400",
		title: "text-red-900 dark:text-red-100",
		content: "text-red-950/90 dark:text-red-100/90"
	},
	default: {
		accent: "border-l-[3px] border-l-border bg-muted/30 rounded-r",
		emphasis: "border-2 border-border bg-muted/50 rounded",
		subtle: "border-border bg-muted/30",
		badge: "bg-foreground text-background",
		label: "text-foreground",
		source: "text-muted-foreground",
		icon: "text-muted-foreground",
		title: "text-foreground",
		content: "text-muted-foreground"
	}
}, ie = {
	note: "Note",
	tip: "Tip",
	important: "Important",
	warning: "Warning",
	caution: "Caution",
	default: ""
};
function Q(e) {
	return re[e];
}
function $(e, t, n) {
	if (!e.includes(t)) throw Error(`Unknown Callout ${n} "${t}" — expected one of: ${e.join(", ")}`);
	return t;
}
// ../../../clicky-ui/packages/ui/src/data/Callout.tsx
var ae = {
	note: d,
	tip: f,
	important: m,
	warning: u,
	caution: p
};
function oe({ children: e, variant: t = "default", label: n, badge: a, source: o, icon: s, emphasis: c = !1, title: l, className: u }) {
	let d = $(Z, t, "variant"), f = s === void 0 ? void 0 : $(X, s, "icon"), p = Q(d), m = c !== !1 && c !== "false", h = n ?? ie[d], g = f ?? (d === "default" ? void 0 : d), _ = g && ae[g];
	return /* @__PURE__ */ (0, i.jsxs)("div", {
		className: r("not-prose my-4 p-density-3", m ? p.emphasis : p.accent, u),
		children: [
			!!(a || h || o || _) && /* @__PURE__ */ (0, i.jsxs)("div", {
				className: "mb-1 flex items-baseline gap-2",
				children: [
					a && /* @__PURE__ */ (0, i.jsx)("span", {
						className: r("rounded px-1.5 py-0.5 text-xs font-bold", p.badge),
						children: a
					}),
					_ && /* @__PURE__ */ (0, i.jsx)("span", {
						className: r("self-center", p.icon),
						children: /* @__PURE__ */ (0, i.jsx)(_, {
							className: p.icon,
							size: "1em"
						})
					}),
					h && /* @__PURE__ */ (0, i.jsx)("span", {
						className: r("text-xs font-semibold uppercase tracking-wide", p.label),
						children: h
					}),
					o && /* @__PURE__ */ (0, i.jsx)("span", {
						className: r("text-xs", p.source),
						children: o
					})
				]
			}),
			l && /* @__PURE__ */ (0, i.jsx)("h3", {
				className: r("mb-2 text-sm font-bold", p.title),
				children: l
			}),
			/* @__PURE__ */ (0, i.jsx)("div", {
				className: r("text-sm leading-snug", p.content),
				children: e
			})
		]
	});
}
export { C as A, m as B, U as C, M as D, K as E, te as F, l as G, f as H, ee as I, s as J, a as K, h as L, x as M, _ as N, j as O, v as P, g as R, V as S, Y as T, d as U, p as V, u as W, o as Y, B as _, L as a, H as b, k as c, O as d, D as f, W as g, T as h, Q as i, b as j, z as k, R as l, P as m, X as n, q as o, ne as p, c as q, Z as r, J as s, oe as t, A as u, I as v, N as w, G as x, F as y, y as z };
