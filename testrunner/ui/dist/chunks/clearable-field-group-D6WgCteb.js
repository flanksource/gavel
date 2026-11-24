import { t as e } from "./jsx-runtime-B4NxTvNR.js";
import { a as t } from "./button-BLoR2pgR.js";
import { a as n, m as r, u as i } from "./Callout-BqeseRSA.js";
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-conditionals.ts
function a(e, t) {
	let n = !1, i = !0;
	if ("const" in e && (n = !0, t !== e.const && (i = !1)), e.enum !== void 0 && (n = !0, e.enum.some((e) => e === t) || (i = !1)), e.not !== void 0) {
		let r = a(e.not, t);
		if (r === void 0) return;
		n = !0, r && (i = !1);
	}
	return (e.properties !== void 0 || e.required !== void 0) && (n = !0, r(t) && o(e, t) || (i = !1)), n ? i : void 0;
}
function o(e, t) {
	if (!e) return !1;
	for (let n of e.required ?? []) if (!(n in t)) return !1;
	for (let [n, r] of Object.entries(e.properties ?? {})) if (a(r, t[n]) !== !0) return !1;
	return !0;
}
function s(e, t, n) {
	if (e) {
		for (let [n, r] of Object.entries(e.properties ?? {})) t[n] = r;
		for (let t of e.required ?? []) n.add(t);
	}
}
function c(e, t, n = "") {
	if (e.if === void 0 && e.then === void 0 && e.else === void 0) return [];
	let r = o(e.if, t) ? "then" : "else", i = e[r];
	if (!i) return [];
	let a = `${n}/${r}`;
	return [{
		branch: i,
		pointer: a
	}, ...c(i, t, a)];
}
function l(e) {
	return [e.then, e.else].flatMap((e) => e ? [e, ...l(e)] : []);
}
function u(e, t) {
	for (let { branch: n } of c(e, t.value)) s(n, t.properties, t.required);
}
function d(e, t) {
	let n = {
		value: t,
		properties: { ...e.properties },
		required: new Set(e.required ?? [])
	};
	u(e, n);
	for (let t of e.allOf ?? []) s(t, n.properties, n.required), u(t, n);
	return {
		properties: n.properties,
		required: [...n.required]
	};
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-utils.ts
var f = "w-full min-w-0 rounded-md border border-input bg-background text-foreground outline-none placeholder:text-placeholder disabled:cursor-not-allowed disabled:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring";
function p(e) {
	return t(f, i[e]);
}
function m(e) {
	return e === "" ? "_" : e.replace(/[^A-Za-z0-9]/g, (e) => `_${e.charCodeAt(0).toString(16)}_`);
}
function h(e, t) {
	let n = (e === "" ? [] : e.replace(/^\//, "").split("/")).map((e) => m(e.replaceAll("~1", "/").replaceAll("~0", "~"))).join("-");
	return t ? `jsf-${m(t)}-${n}` : `jsf-${n}`;
}
function g(e) {
	return `${e}-error`;
}
function _(e, t) {
	return {
		...e.required ? { "aria-required": !0 } : {},
		...e.invalid ? {
			"aria-invalid": !0,
			"aria-describedby": g(t)
		} : {}
	};
}
function v(e, t) {
	return {
		...e.required ? { ariaRequired: !0 } : {},
		...e.invalid ? {
			invalid: !0,
			describedBy: g(t)
		} : {}
	};
}
function y(e) {
	if (e.default != null && typeof e.default != "object") return String(e.default);
	let t = e.examples?.[0];
	return t != null && typeof t != "object" ? String(t) : "";
}
function b(e) {
	return e == null || typeof e == "object" ? "" : String(e);
}
function x(e) {
	return Array.isArray(e) ? e.filter((e) => typeof e == "string") : [];
}
function S(e, t) {
	return t !== "properties" && (e.kind === "object" && e.schema["x-object-display"] !== "row" || e.layout === "table" || e.arrayDisplay === "accordion" || e.arrayDisplay === "cards" || e.kind === "array" && e.arrayDisplay !== "stacked" && w(e.itemSchema));
}
function C(e, t) {
	if (t || e.value === null || e.value === void 0) return !1;
	let n = e.kind === "array" && e.arrayDisplay === void 0 && (!!e.itemSchema?.enum || [
		"string",
		"integer",
		"number"
	].includes(String(e.itemSchema?.type)));
	return [
		"object",
		"array",
		"string-map"
	].includes(e.kind) && !n || e.kind === "lookup" && !!(e.lookup?.multi && e.lookup?.hierarchy);
}
function w(e) {
	return e ? Object.keys(d(e, {}).properties).length > 0 : !1;
}
function T(e) {
	if (e === "auto") return "auto";
	let t = typeof e == "number" ? e : Number(e);
	return !Number.isFinite(t) || t < 1 ? 1 : Math.min(Math.trunc(t), 12);
}
function E(e, t) {
	if (e === "full") return "full";
	let n = typeof e == "number" ? e : Number(e);
	return !Number.isFinite(n) || n < 1 ? 1 : Math.min(Math.trunc(n), t === "auto" ? 12 : t);
}
var D = "15rem";
function O(e, t) {
	return typeof e == "string" && e.trim() ? e.trim() : t;
}
function k(e, t) {
	return !t || e.some((e) => e.value === t) ? e : [{
		value: t,
		label: t
	}, ...e];
}
function A(e, t, n) {
	let r = new Set(t.filter((e) => e !== n));
	return e.filter((e) => !r.has(e.value));
}
function j(e, t) {
	if (!Array.isArray(t) || t.length === 0) return e;
	let n = new Map(t.filter((e) => typeof e == "string").map((e, t) => [e, t])), r = e.filter(([e]) => n.has(e)).sort((e, t) => n.get(e[0]) - n.get(t[0])), i = e.filter(([e]) => !n.has(e));
	return [...r, ...i];
}
function M(e) {
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
function N(e, t) {
	let n = new Set(t), r = e.filter(([e]) => n.has(e)), i = e.filter(([e]) => !n.has(e));
	return [...r, ...i];
}
function P(e) {
	return e == null || e === "" ? !0 : Array.isArray(e) ? e.every(P) : typeof e == "object" && Object.values(e).every(P);
}
function F(e, t, n) {
	let r = new Set(t), i = (e) => (r.has(e) ? 2 : 0) + +!P(n[e]);
	return e.map((e, t) => ({
		entry: e,
		index: t,
		score: i(e[0])
	})).sort((e, t) => t.score - e.score || e.index - t.index).map((e) => e.entry);
}
function I(e) {
	let t = e.value, n = t == null || t === "";
	if (e.required && n && !(t === null && e.nullable)) return {
		keyword: "required",
		message: "Required"
	};
	if (e.kind === "number" && typeof e.minimum == "number" && typeof t == "number" && t < e.minimum) return {
		keyword: "minimum",
		message: `Must be ≥ ${e.minimum}`
	};
	if (e.kind === "enum" && !e.allowCustomValue && !n && !(e.options ?? []).some((e) => e.value === String(t))) return {
		keyword: "enum",
		message: "Unknown value (allowed)"
	};
}
function L(e, t) {
	return e.maxItems === void 0 || t < e.maxItems;
}
function R(e, t) {
	return e.minItems === void 0 || t > e.minItems;
}
function z(e) {
	if (e.default !== void 0) return e.default;
	let t = Array.isArray(e.type) ? e.type[0] : e.type;
	return t === "array" ? [] : t === "object" ? {} : t !== "boolean" && "";
}
// ../../../clicky-ui/packages/ui/src/components/clearable-field-group.tsx
var B = e();
function V({ children: e, clearable: t, onClear: r, clearLabel: i }) {
	return t ? /* @__PURE__ */ (0, B.jsxs)("div", {
		"data-clearable-field": !0,
		className: "relative min-w-0 rounded-md border border-input bg-background p-2 pt-9",
		children: [/* @__PURE__ */ (0, B.jsx)("span", {
			className: "absolute right-1 top-1",
			children: /* @__PURE__ */ (0, B.jsx)(n, {
				...r ? { onClear: r } : {},
				...i ? { clearLabel: i } : {}
			})
		}), e]
	}) : e;
}
export { a as A, z as C, k as D, b as E, c as M, l as O, S, x as T, C as _, v as a, j as b, _ as c, w as d, p as f, T as g, E as h, R as i, o as j, d as k, g as l, A as m, D as n, O as o, P as p, L as r, y as s, V as t, h as u, M as v, I as w, N as x, F as y };
