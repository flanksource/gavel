import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-B4NxTvNR.js";
import { t as r } from "./react-dom-BBpWcFtU.js";
import { a as i, i as a, t as o } from "./button-BLoR2pgR.js";
import { i as s, r as c } from "./UiInfo-CCiCR0Fr.js";
import { n as l } from "./use-theme-sEqHIo2i.js";
import { t as u } from "./dist-DCJEOtHh.js";
import { _ as d, a as f, b as p, c as m, d as h, f as g, g as _, h as v, l as y, m as b, n as x, o as S, p as C, s as w, u as T, v as E, y as D } from "./floating-ui.react-dom-CJoIjj1G.js";
// ../../../clicky-ui/packages/ui/src/data/badge-variants.ts
var O = /* @__PURE__ */ e(t(), 1), k = a("inline-flex items-center gap-1 rounded-full border border-transparent font-medium whitespace-nowrap align-middle", {
	variants: {
		tone: {
			neutral: "",
			success: "",
			danger: "",
			warning: "",
			info: ""
		},
		variant: {
			soft: "",
			solid: "",
			outline: "bg-transparent"
		},
		size: {
			xxs: "min-h-4 px-1.5 py-0.5 text-[9px] leading-none",
			xs: "min-h-[18px] px-1.5 py-0.5 text-[10px] leading-none",
			sm: "h-4 px-1.5 py-0 text-[10px] leading-none",
			md: "px-2 py-0.5 text-xs",
			lg: "px-2.5 py-1 text-sm"
		}
	},
	compoundVariants: [
		{
			tone: "neutral",
			variant: "soft",
			class: "bg-muted text-foreground"
		},
		{
			tone: "neutral",
			variant: "solid",
			class: "bg-foreground text-background"
		},
		{
			tone: "neutral",
			variant: "outline",
			class: "border-border text-foreground"
		},
		{
			tone: "success",
			variant: "soft",
			class: "bg-green-100 text-green-800 dark:bg-green-500/20 dark:text-green-300"
		},
		{
			tone: "success",
			variant: "solid",
			class: "bg-green-500 text-white"
		},
		{
			tone: "success",
			variant: "outline",
			class: "border-green-500 text-green-700 dark:text-green-400"
		},
		{
			tone: "danger",
			variant: "soft",
			class: "bg-red-100 text-red-800 dark:bg-red-500/20 dark:text-red-300"
		},
		{
			tone: "danger",
			variant: "solid",
			class: "bg-red-500 text-white"
		},
		{
			tone: "danger",
			variant: "outline",
			class: "border-red-500 text-red-700 dark:text-red-400"
		},
		{
			tone: "warning",
			variant: "soft",
			class: "bg-yellow-100 text-yellow-800 dark:bg-yellow-500/20 dark:text-yellow-300"
		},
		{
			tone: "warning",
			variant: "solid",
			class: "bg-yellow-400 text-yellow-950"
		},
		{
			tone: "warning",
			variant: "outline",
			class: "border-yellow-500 text-yellow-700 dark:text-yellow-400"
		},
		{
			tone: "info",
			variant: "soft",
			class: "bg-blue-100 text-blue-800 dark:bg-blue-500/20 dark:text-blue-300"
		},
		{
			tone: "info",
			variant: "solid",
			class: "bg-blue-500 text-white"
		},
		{
			tone: "info",
			variant: "outline",
			class: "border-blue-500 text-blue-700 dark:text-blue-400"
		}
	],
	defaultVariants: {
		tone: "neutral",
		variant: "soft",
		size: "md"
	}
}), A = n(), j = /* @__PURE__ */ new Set([
	"soft",
	"solid",
	"outline"
]), M = /* @__PURE__ */ new Set([
	"status",
	"metric",
	"custom",
	"outlined",
	"label"
]), N = {
	xxs: {
		frame: "min-h-4",
		segment: "px-1.5 py-0.5",
		text: "text-[10px] leading-4",
		icon: "h-3 w-3",
		gap: "gap-1"
	},
	xs: {
		frame: "min-h-[18px]",
		segment: "px-2 py-0.5",
		text: "text-[11px] leading-4",
		icon: "h-3.5 w-3.5",
		gap: "gap-1"
	},
	sm: {
		frame: "min-h-5",
		segment: "px-2 py-0.5",
		text: "text-xs leading-4",
		icon: "h-3.5 w-3.5",
		gap: "gap-1.5"
	},
	md: {
		frame: "min-h-6",
		segment: "px-2.5 py-1",
		text: "text-sm leading-4",
		icon: "h-4 w-4",
		gap: "gap-1.5"
	},
	lg: {
		frame: "min-h-7",
		segment: "px-3 py-1.5",
		text: "text-base leading-5",
		icon: "h-5 w-5",
		gap: "gap-2"
	}
}, P = {
	success: {
		soft: "bg-emerald-500/12 text-emerald-700 border-emerald-200 dark:text-emerald-300 dark:border-emerald-500/30",
		outlined: "bg-background text-emerald-700 border-emerald-300 dark:text-emerald-400 dark:border-emerald-500/40"
	},
	error: {
		soft: "bg-rose-500/12 text-rose-700 border-rose-200 dark:text-rose-300 dark:border-rose-500/30",
		outlined: "bg-background text-rose-700 border-rose-300 dark:text-rose-400 dark:border-rose-500/40"
	},
	warning: {
		soft: "bg-amber-400/15 text-amber-800 border-amber-200 dark:text-amber-300 dark:border-amber-500/30",
		outlined: "bg-background text-amber-800 border-amber-300 dark:text-amber-400 dark:border-amber-500/40"
	},
	info: {
		soft: "bg-sky-500/12 text-sky-700 border-sky-200 dark:text-sky-300 dark:border-sky-500/30",
		outlined: "bg-background text-sky-700 border-sky-300 dark:text-sky-400 dark:border-sky-500/40"
	}
};
function F(e) {
	return e != null && j.has(e);
}
function I(e) {
	return e != null && /^(#|rgb|hsl|oklch|var\(|color\()/i.test(e);
}
function L(e) {
	switch (e) {
		case "success": return "success";
		case "danger": return "error";
		case "warning": return "warning";
		case "info": return "info";
		default: return;
	}
}
function R(e) {
	switch (e) {
		case "rounded": return "rounded-md";
		case "square": return "rounded-sm";
		default: return "rounded-full";
	}
}
function ee(e) {
	return e != null && M.has(e);
}
function z(e, t, n, r) {
	e != null && (I(e) ? n[t] = e : r.push(e));
}
function B(e) {
	if (e != null) return typeof e == "number" ? `${Math.max(1, Math.floor(e))}ch` : e;
}
function V(e, t = 24) {
	if (typeof e == "number" && Number.isFinite(e) && e > 0) return Math.max(4, Math.floor(e));
	if (typeof e == "string") {
		let t = e.trim(), n = t.match(/^(\d+(?:\.\d+)?)ch$/i);
		if (n) return Math.max(4, Math.floor(Number(n[1])));
		let r = t.match(/^(\d+(?:\.\d+)?)$/);
		if (r) return Math.max(4, Math.floor(Number(r[1])));
	}
	return t;
}
function H({ maxWidth: e, truncate: t, wrap: n }) {
	return e != null || t != null || n;
}
function U(e) {
	if (e == null || typeof e == "boolean") return null;
	if (typeof e == "string" || typeof e == "number") return String(e);
	if (Array.isArray(e)) {
		let t = O.Children.toArray(e).map(U);
		return t.every((e) => e != null) ? t.join("") : null;
	}
	return (0, O.isValidElement)(e) ? U(e.props.children) : null;
}
function W(e, t) {
	return e.length <= t ? e : t <= 1 ? "…" : `${e.slice(0, t - 1)}…`;
}
function te(e, t) {
	return e.length <= t ? e : t <= 1 ? "…" : `…${e.slice(-(t - 1))}`;
}
function G(e, t) {
	if (e.length <= t) return e;
	if (t <= 3) return W(e, t);
	let n = t - 1, r = Math.max(1, Math.ceil(n / 2)), i = Math.max(1, n - r);
	return `${e.slice(0, r)}…${e.slice(-i)}`;
}
function ne(e, t) {
	return e.length <= t ? e : t <= 2 ? W(e, t) : `${e.slice(0, t - 1)}…`;
}
function re(e, t) {
	return e.length <= t ? e : t <= 2 ? te(e, t) : `…${e.slice(-(t - 1))}`;
}
function ie(e) {
	return e == null || e.length === 0 ? "…" : e.length <= 2 ? `${e}…` : `${e[0]}…`;
}
function K(e, t, n, r, i) {
	let a = [e];
	t && a.push(t), a.push(n);
	let o = a.join(i);
	if (o.length <= r) return o;
	let s = a.length - 1, c = i.length * s + (t ? t.length : 0), l = Math.max(4, r - c), u = Math.max(3, Math.min(e.length, Math.floor(l * .4))), d = Math.max(4, l - u), f = ne(e, u), p = re(n, d), m = [f];
	t && m.push(t), m.push(p);
	let h = m.join(i);
	return h.length <= r ? h : G(o, r);
}
function q(e) {
	return Math.max(e, 32);
}
function ae(e, t) {
	let n = q(t), r = e.startsWith("/"), i = e.split("/").filter(Boolean);
	if (i.length < 2) return G(e, n);
	let a = `${r ? "/" : ""}${i[0]}`, o = i[i.length - 1] ?? e;
	return K(a, i.length > 2 ? ie(i.slice(1, -1).join("/")) : void 0, o, n, "/");
}
function oe(e, t) {
	let n = q(t);
	try {
		let t = new URL(e), r = t.host, i = t.pathname.split("/").filter(Boolean);
		if (i.length === 0) return r.length <= n ? r : G(`${r}${t.search}${t.hash}`, n);
		if (i.length === 1) {
			let e = `${r}/${i[0] ?? r}`;
			if (e.length <= n) return e;
			let t = `${r}/…`;
			return t.length <= n ? t : G(e, n);
		}
		let a = i[i.length - 1] ?? r;
		return [
			[
				r,
				i[0] ?? "",
				i.length > 2 ? ie(i.slice(1, -1).join("/")) : void 0,
				a
			].filter(Boolean).join("/"),
			`${r}/…/${a}`,
			`${r}/…`
		].find((e) => e.length <= n) ?? G(`${r}/${a}`, n);
	} catch {
		return G(e, n);
	}
}
function se(e) {
	return /^[a-z][a-z\d+\-.]*:\/\//i.test(e) || e.startsWith("//");
}
function ce(e) {
	if (/\s/.test(e) || se(e) || e.startsWith("/") || !e.includes("/")) return !1;
	let t = e.slice(e.lastIndexOf("/") + 1);
	return t.includes(":") || t.includes("@sha256:");
}
function le(e) {
	let t = e.indexOf("@sha256:");
	if (t >= 0) return {
		name: e.slice(0, t),
		suffix: e.slice(t)
	};
	let n = e.lastIndexOf(":");
	return n > 0 ? {
		name: e.slice(0, n),
		suffix: e.slice(n)
	} : {
		name: e,
		suffix: ""
	};
}
function ue(e, t) {
	let n = q(t), r = e.split("/").filter(Boolean);
	if (r.length < 2) return G(e, n);
	let i = r[r.length - 1] ?? e, a = le(i), o = `…${a.name}`;
	if (o.length <= n) return o;
	if (r.length === 2) return K(r[0] ?? e, void 0, a.name || i, n, "/");
	let s = `${r[0]}/${r[1] ?? ""}`.replace(/\/$/, ""), c = a.suffix.length > 0 && n >= i.length ? i : a.name || i;
	return K(s, r.length > 3 ? ie(r.slice(2, -1).join("/")) : r.length === 3 ? "…" : void 0, c, n, "/");
}
function de(e, t) {
	let n = q(t), r = e.split(":");
	if (r.length < 6) return G(e, n);
	let i = r[2] ?? r[1] ?? "arn", a = r.slice(5).join(":"), o = a.split(/[/:]/).filter(Boolean), s = o.length > 1 ? o[0] : void 0, c = o[o.length - 1] ?? a, l = c.includes("-") && c.split("-").length > 2 ? c.split("-").slice(-2).join("-") : c, u = `${i}:…${s != null && s !== l ? `${s}/${l}` : l}`;
	if (u.length <= n) return u;
	let d = Math.max(4, Math.min(l.length, n - 2 - 2));
	return `${ne(i, Math.max(2, n - d - 2))}:…${G(l, d)}`;
}
function fe(e) {
	return e.startsWith("arn:") ? "arn" : se(e) ? "url" : ce(e) ? "image" : e.includes("/") ? "path" : "suffix";
}
function pe(e, t, n) {
	let r = V(n);
	switch (t === "auto" ? fe(e) : t) {
		case "prefix": return te(e, r);
		case "suffix": return W(e, r);
		case "arn": return de(e, r);
		case "image": return ue(e, r);
		case "path": return ae(e, r);
		case "url": return oe(e, r);
		default: return W(e, r);
	}
}
function me(e, t, n, r = !1) {
	let i = U(e);
	if (i == null || t == null) return {
		content: e,
		fullText: i,
		title: r ? i ?? void 0 : void 0
	};
	let a = pe(i, t, n);
	return {
		content: a,
		fullText: i,
		title: r || a !== i ? i : void 0
	};
}
function he(e) {
	let t = globalThis.navigator?.clipboard;
	t?.writeText != null && t.writeText(e).catch(() => void 0);
}
function ge({ href: e, target: t, rel: n, className: r, style: i, content: a, canCopy: o, copyText: s }) {
	return e == null ? o && s != null ? /* @__PURE__ */ (0, A.jsx)("button", {
		type: "button",
		className: r,
		style: i,
		onClick: () => he(s),
		children: a
	}) : /* @__PURE__ */ (0, A.jsx)("span", {
		className: r,
		style: i,
		children: a
	}) : /* @__PURE__ */ (0, A.jsx)("a", {
		href: e,
		target: t,
		rel: n ?? (t === "_blank" ? "noopener noreferrer" : void 0),
		className: r,
		style: i,
		children: a
	});
}
function _e({ tone: e, variant: t, size: n = "md", icon: r, count: a, children: o, className: c, status: l, label: u, value: d, color: f, textColor: p, borderColor: m, shape: h, href: g, target: _ = "_self", rel: v, wrap: y = !1, maxWidth: b, truncate: x, clickToCopy: S, labelClassName: C, valueClassName: w }) {
	let T = t ?? (u != null || d != null || l != null || h != null || g != null || y === !0 || f != null || p != null || m != null || C != null || w != null ? "metric" : "soft"), E = B(b), D = y && x == null, O = H({
		maxWidth: b,
		truncate: x,
		wrap: y
	}), j = D ? "min-w-0 whitespace-normal break-words [overflow-wrap:anywhere]" : E != null || x != null ? "min-w-0 overflow-hidden text-ellipsis whitespace-nowrap" : "whitespace-nowrap";
	if (F(T)) {
		let t = {};
		E != null && (t.maxWidth = E);
		let l = me(o, x, b, O), u = l.fullText ?? (a === void 0 ? void 0 : String(a)), d = g == null && (S ?? !0) && u != null;
		return ge({
			href: g,
			target: _,
			rel: v,
			canCopy: d,
			copyText: u,
			className: i(k({
				tone: e,
				variant: T,
				size: n
			}), (E != null || D || x != null) && "min-w-0 max-w-full", D && "whitespace-normal", d && "cursor-pointer text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1", c),
			style: t,
			content: /* @__PURE__ */ (0, A.jsxs)(A.Fragment, { children: [
				r && /* @__PURE__ */ (0, A.jsx)(s, {
					...typeof r == "string" ? { name: r } : { icon: r },
					className: "text-[1em]"
				}),
				a !== void 0 && /* @__PURE__ */ (0, A.jsx)("span", { children: a }),
				o != null && /* @__PURE__ */ (0, A.jsx)("span", {
					className: j,
					title: l.title,
					children: l.content
				})
			] })
		});
	}
	if (!ee(T)) return null;
	let M = l ?? L(e) ?? "info", I = N[n], V = u ?? o, W = d ?? (V == null && a !== void 0 ? a : void 0), te = W == null ? me(V, x, b, O) : {
		content: V,
		fullText: U(V),
		title: O ? U(V) ?? void 0 : void 0
	}, G = W == null ? {
		content: W,
		fullText: void 0,
		title: void 0
	} : me(W, x, b, O), ne = G.fullText ?? te.fullText ?? (a === void 0 ? void 0 : String(a)), re = g == null && (S ?? !0) && ne != null, ie = R(h ?? "pill"), K = {}, q = [
		"inline-flex align-middle items-stretch border font-medium shadow-none",
		I.frame,
		I.text,
		E != null || D || x != null ? "min-w-0 max-w-full" : "",
		ie,
		D ? "whitespace-normal" : "whitespace-nowrap",
		g ? "transition-opacity hover:opacity-80" : "",
		re ? "cursor-pointer text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1" : ""
	];
	E != null && (K.maxWidth = E), T === "status" ? q.push(P[M].soft) : T === "metric" ? q.push("bg-muted/60 text-foreground border-border") : T === "outlined" ? l != null || L(e) != null ? q.push(P[M].outlined) : q.push("bg-background text-foreground border-border") : T === "custom" ? q.push("border-transparent bg-accent text-accent-foreground") : q.push("border-border bg-background text-foreground"), (T === "custom" || T === "outlined" && l == null && L(e) == null) && (z(f, "backgroundColor", K, q), z(p, "color", K, q), z(m, "borderColor", K, q)), T !== "label" && m != null && l == null && L(e) == null && z(m, "borderColor", K, q);
	let ae = r ? /* @__PURE__ */ (0, A.jsx)(s, {
		...typeof r == "string" ? { name: r } : { icon: r },
		className: i("shrink-0", I.icon)
	}) : null;
	if (T === "label") {
		let e = V != null || ae != null, t = [
			"inline-flex items-center self-stretch",
			W == null ? "min-w-0" : "shrink-0",
			I.segment,
			I.gap,
			W == null ? "" : "border-r border-border/70"
		], n = [
			"inline-flex min-w-0 items-center self-stretch",
			p == null || e ? "text-foreground" : "",
			e ? "flex-1" : "",
			I.segment
		], r = {};
		return e ? (t.push("bg-secondary text-secondary-foreground"), z(f, "backgroundColor", r, t), z(p, "color", r, t)) : (z(f, "backgroundColor", K, q), z(p, "color", K, q)), z(m, "borderColor", K, q), ge({
			href: g,
			target: _,
			rel: v,
			canCopy: re,
			copyText: ne,
			className: i(q, c),
			style: K,
			content: /* @__PURE__ */ (0, A.jsxs)(A.Fragment, { children: [e && /* @__PURE__ */ (0, A.jsxs)("span", {
				className: i(t, C),
				style: r,
				children: [ae, V != null && /* @__PURE__ */ (0, A.jsx)("span", {
					className: j,
					title: te.title,
					children: te.content
				})]
			}), W != null && /* @__PURE__ */ (0, A.jsx)("span", {
				className: i(n, w),
				children: /* @__PURE__ */ (0, A.jsx)("span", {
					className: j,
					title: G.title,
					children: G.content
				})
			})] })
		});
	}
	let oe = [
		"inline-flex items-center",
		W == null ? "min-w-0" : "shrink-0",
		I.segment,
		I.gap,
		W == null ? "" : "border-r border-current/10"
	], se = [
		"inline-flex min-w-0 items-center font-semibold",
		V != null || ae != null ? "flex-1" : "",
		I.segment
	];
	return T === "metric" && (oe.push("text-muted-foreground"), se.push("text-foreground")), ge({
		href: g,
		target: _,
		rel: v,
		canCopy: re,
		copyText: ne,
		className: i(q, c),
		style: K,
		content: /* @__PURE__ */ (0, A.jsxs)(A.Fragment, { children: [(V != null || ae != null) && /* @__PURE__ */ (0, A.jsxs)("span", {
			className: i(oe, C),
			children: [ae, V != null && /* @__PURE__ */ (0, A.jsx)("span", {
				className: j,
				title: te.title,
				children: te.content
			})]
		}), W != null && /* @__PURE__ */ (0, A.jsx)("span", {
			className: i(se, w),
			children: /* @__PURE__ */ (0, A.jsx)("span", {
				className: j,
				title: G.title,
				children: G.content
			})
		})] })
	});
}
// ../../../clicky-ui/packages/ui/src/icons/components/UiChevronDown.tsx
var ve = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
		fill: "currentColor",
		d: "m212.24 100.24l-80 80a6 6 0 0 1-8.48 0l-80-80a6 6 0 0 1 8.48-8.48L128 167.51l75.76-75.75a6 6 0 0 1 8.48 8.48"
	})
}), {
	__source: "ph:caret-down-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "chevron-down",
	displayName: "UiChevronDown"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
		fill: "currentColor",
		d: "m213.66 101.66l-80 80a8 8 0 0 1-11.32 0l-80-80A8 8 0 0 1 48 88h160a8 8 0 0 1 5.66 13.66"
	})
}), {
	__source: "ph:caret-down-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "chevron-down",
	displayName: "UiChevronDownFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiChevronRight.tsx
var ye = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
		fill: "currentColor",
		d: "m180.24 132.24l-80 80a6 6 0 0 1-8.48-8.48L167.51 128L91.76 52.24a6 6 0 0 1 8.48-8.48l80 80a6 6 0 0 1 0 8.48"
	})
}), {
	__source: "ph:caret-right-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "chevron-right",
	displayName: "UiChevronRight"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
		fill: "currentColor",
		d: "m181.66 133.66l-80 80A8 8 0 0 1 88 208V48a8 8 0 0 1 13.66-5.66l80 80a8 8 0 0 1 0 11.32"
	})
}), {
	__source: "ph:caret-right-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "chevron-right",
	displayName: "UiChevronRightFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiChevronUp.tsx
var be = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiAdd.tsx
var xe = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiFullscreen.tsx
var Se = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
		fill: "currentColor",
		d: "M214 48v40a6 6 0 0 1-12 0V54h-34a6 6 0 0 1 0-12h40a6 6 0 0 1 6 6M88 202H54v-34a6 6 0 0 0-12 0v40a6 6 0 0 0 6 6h40a6 6 0 0 0 0-12m120-40a6 6 0 0 0-6 6v34h-34a6 6 0 0 0 0 12h40a6 6 0 0 0 6-6v-40a6 6 0 0 0-6-6M88 42H48a6 6 0 0 0-6 6v40a6 6 0 0 0 12 0V54h34a6 6 0 0 0 0-12"
	})
}), {
	__source: "ph:corners-out-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "fullscreen",
	displayName: "UiFullscreen"
}), Ce = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
		fill: "currentColor",
		d: "M93.66 202.34A8 8 0 0 1 88 216H48a8 8 0 0 1-8-8v-40a8 8 0 0 1 13.66-5.66ZM88 40H48a8 8 0 0 0-8 8v40a8 8 0 0 0 13.66 5.66l40-40A8 8 0 0 0 88 40m123.06 120.61a8 8 0 0 0-8.72 1.73l-40 40A8 8 0 0 0 168 216h40a8 8 0 0 0 8-8v-40a8 8 0 0 0-4.94-7.39M208 40h-40a8 8 0 0 0-5.66 13.66l40 40A8 8 0 0 0 216 88V48a8 8 0 0 0-8-8"
	})
}), {
	__source: "ph:corners-out-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "fullscreen",
	displayName: "UiFullscreenFilled"
}), we = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
var Te = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiCheck.tsx
var Ee = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#059669",
		...r
	};
	return /* @__PURE__ */ (0, A.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, A.jsx)("path", {
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
}), De = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#059669",
		...r
	};
	return /* @__PURE__ */ (0, A.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, A.jsx)("path", {
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
}), Oe = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiRows.tsx
var ke = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiColumns.tsx
var Ae = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
		fill: "currentColor",
		d: "M104 34H64a14 14 0 0 0-14 14v160a14 14 0 0 0 14 14h40a14 14 0 0 0 14-14V48a14 14 0 0 0-14-14m2 174a2 2 0 0 1-2 2H64a2 2 0 0 1-2-2V48a2 2 0 0 1 2-2h40a2 2 0 0 1 2 2Zm86-174h-40a14 14 0 0 0-14 14v160a14 14 0 0 0 14 14h40a14 14 0 0 0 14-14V48a14 14 0 0 0-14-14m2 174a2 2 0 0 1-2 2h-40a2 2 0 0 1-2-2V48a2 2 0 0 1 2-2h40a2 2 0 0 1 2 2Z"
	})
}), {
	__source: "ph:columns-light",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "columns",
	displayName: "UiColumns"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
		fill: "currentColor",
		d: "M120 48v160a16 16 0 0 1-16 16H64a16 16 0 0 1-16-16V48a16 16 0 0 1 16-16h40a16 16 0 0 1 16 16m72-16h-40a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h40a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16"
	})
}), {
	__source: "ph:columns-fill",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "columns",
	displayName: "UiColumnsFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiFile.tsx
var je = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiCopy.tsx
var Me = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiTrash.tsx
var Ne = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, A.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, A.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/components/SegmentedControl.tsx
var Pe = {
	sm: "text-xs px-density-2 py-density-1 gap-1",
	md: "h-full px-control-px text-sm gap-1.5",
	lg: "min-h-[4.25rem] min-w-44 flex-1 px-density-3 py-density-3 gap-density-2 text-sm"
};
function Fe({ value: e, options: t, onChange: n, size: r = "md", wrap: a = !1, className: o, "aria-label": c }) {
	let l = (0, O.useRef)([]), u = (0, O.useId)(), d = t.some((t) => t.id === e), f = (e, r) => {
		let i = t.length, a = e;
		for (let e = 0; e < i; e++) {
			a = (a + r + i) % i;
			let e = t[a];
			if (e && !e.disabled) {
				n(e.id), l.current[a]?.focus();
				return;
			}
		}
	};
	return /* @__PURE__ */ (0, A.jsx)("div", {
		role: "radiogroup",
		"aria-label": c,
		className: i("inline-flex items-center rounded-md bg-muted p-0.5", r === "md" && (a ? "min-h-control-h" : "h-control-h"), r === "lg" && "gap-density-2 bg-transparent p-0", a && "flex-wrap", o),
		children: t.map((t, o) => {
			let c = t.id === e, p = !d && t.implied === !0, m = t.marker ? `${u}-marker-${o}` : void 0;
			return /* @__PURE__ */ (0, A.jsxs)("button", {
				ref: (e) => {
					l.current[o] = e;
				},
				type: "button",
				role: "radio",
				"aria-checked": c,
				"aria-describedby": m,
				"data-implied": p ? "true" : void 0,
				title: t.title,
				disabled: t.disabled,
				tabIndex: c || p ? 0 : -1,
				onClick: () => !t.disabled && n(t.id),
				onKeyDown: (e) => {
					e.key === "ArrowRight" || e.key === "ArrowDown" ? (e.preventDefault(), f(o, 1)) : (e.key === "ArrowLeft" || e.key === "ArrowUp") && (e.preventDefault(), f(o, -1));
				},
				className: i("inline-flex items-center rounded-[5px] border border-transparent font-medium transition-colors", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", "disabled:cursor-not-allowed disabled:opacity-50", r === "lg" && "rounded-lg text-left", t.description ? "whitespace-normal" : "whitespace-nowrap", Pe[r], r === "md" && a && "h-auto py-1", c ? r === "lg" ? "border-primary/70 bg-primary/10 text-foreground shadow-sm ring-1 ring-primary/25" : "border-border bg-card text-foreground shadow-sm" : p ? r === "lg" ? "border-dashed border-primary/60 bg-card text-foreground hover:shadow-sm" : "border-dashed border-primary/60 text-foreground hover:bg-background/60" : r === "lg" ? "border-border bg-card text-muted-foreground hover:border-muted-foreground/40 hover:bg-card hover:text-foreground hover:shadow-sm" : "text-foreground/80 hover:border-border hover:bg-background/60 hover:text-foreground", c && t.activeClassName),
				children: [(0, O.isValidElement)(t.icon) ? /* @__PURE__ */ (0, A.jsx)("span", {
					className: i("inline-flex shrink-0", t.iconClassName),
					children: t.icon
				}) : t.icon && /* @__PURE__ */ (0, A.jsx)(s, {
					...typeof t.icon == "string" ? { name: t.icon } : { icon: t.icon },
					className: i(r === "lg" ? "size-4 shrink-0" : "size-3.5 shrink-0", r === "lg" && (c ? "text-primary" : "text-muted-foreground"), t.iconClassName)
				}), /* @__PURE__ */ (0, A.jsxs)("span", {
					className: i("min-w-0", r === "lg" && "flex flex-col items-start text-left leading-tight"),
					children: [t.marker ? /* @__PURE__ */ (0, A.jsxs)("span", {
						className: "inline-flex min-w-0 max-w-full items-baseline gap-1",
						children: [/* @__PURE__ */ (0, A.jsx)("span", {
							className: "min-w-0 truncate",
							children: t.label
						}), /* @__PURE__ */ (0, A.jsx)("span", {
							id: m,
							"aria-hidden": "true",
							className: "shrink-0 text-[11px] font-normal opacity-70",
							children: t.marker
						})]
					}) : /* @__PURE__ */ (0, A.jsx)("span", {
						className: "min-w-0 truncate",
						children: t.label
					}), t.description && /* @__PURE__ */ (0, A.jsx)("span", {
						className: i("mt-0.5 max-w-56 whitespace-normal break-words text-[11px] font-normal leading-snug", c ? "text-muted-foreground" : "text-muted-foreground/80"),
						children: t.description
					})]
				})]
			}, t.id);
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/code-diff.ts
function Ie(e) {
	return e === "" ? [] : (e.endsWith("\n") ? e.slice(0, -1) : e).split("\n");
}
function Le(e, t) {
	let n = Re(Ie(e), Ie(t));
	return n.length > 0 ? [{ lines: n }] : [];
}
function Re(e, t) {
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
var ze = /^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@/;
function Be(e) {
	let t = e.split("	")[0].trim();
	if (t !== "" && t !== "/dev/null") return t.startsWith("a/") || t.startsWith("b/") ? t.slice(2) : t;
}
function Ve(e) {
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
		let c = ze.exec(e);
		if (!c) throw Error(`Malformed unified diff hunk header: ${e}`);
		r = Number(c[1]), i = Number(c[3]), a = c[2] ? Number(c[2]) : 1, o = c[4] ? Number(c[4]) : 1, n = {
			header: e,
			lines: [],
			...s ? { path: s } : {}
		}, t.push(n), a <= 0 && o <= 0 && (n = null);
	} else e.startsWith("diff --git ") ? (s = void 0, c = void 0) : e.startsWith("--- ") ? c = Be(e.slice(4)) : e.startsWith("+++ ") && (s = Be(e.slice(4)) ?? c);
	return t;
}
// ../../../clicky-ui/packages/ui/src/data/code-diff-selection.ts
function He(e) {
	if (e.type === "add" && e.newNumber !== void 0) return {
		side: "new",
		line: e.newNumber
	};
	if (e.type === "remove" && e.oldNumber !== void 0) return {
		side: "old",
		line: e.oldNumber
	};
}
function Ue(e) {
	return e.flatMap((e) => He(e) ?? []);
}
function We(e, t, n) {
	let [r, i] = t <= n ? [t, n] : [n, t];
	return Ue(e.slice(r, i + 1));
}
function Ge(e, t) {
	let n = e.filter(t).length;
	return n === 0 ? "none" : n === e.length ? "all" : "partial";
}
// ../../../clicky-ui/packages/ui/src/data/code-highlight.ts
var Ke = [
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
], qe = ["github-light", "github-dark"], Je = "github-light", Ye = null;
function Xe() {
	return Ye ??= (async () => {
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
	})(), Ye;
}
function Ze(e) {
	if (!e) return;
	let t = e.slice(e.lastIndexOf("/") + 1), n = t.lastIndexOf(".");
	if (!(n <= 0)) return t.slice(n + 1).toLowerCase();
}
function Qe(e) {
	let t = e.toLowerCase();
	return Ke.includes(t) ? t : t === "shell" || t === "sh" || t === "zsh" ? "bash" : t === "ts" ? "typescript" : t === "js" ? "javascript" : t === "py" ? "python" : t === "yml" ? "yaml" : null;
}
function $e(e) {
	return e && qe.includes(e) ? e : Je;
}
var et = 8, tt = [], nt = !1;
function rt(e) {
	return new Promise((t, n) => {
		tt.push(() => {
			try {
				t(e());
			} catch (e) {
				n(e);
			}
		}), nt || (nt = !0, setTimeout(it, 0));
	});
}
function it() {
	let e = performance.now() + et;
	for (; tt.length > 0 && performance.now() < e;) tt.shift()();
	tt.length > 0 ? setTimeout(it, 0) : nt = !1;
}
var at = 2e6, ot = /* @__PURE__ */ new Map(), st = 0;
function ct(e, t) {
	ot.set(e, t), st += t.length;
	for (let [e, t] of ot) {
		if (st <= at) break;
		ot.delete(e), st -= t.length;
	}
}
async function lt(e, t) {
	if (!e || !t.lang) return null;
	let n = Qe(t.lang);
	if (!n) return null;
	let r = $e(t.theme), i = t.transformers ? void 0 : `${n} ${r} ${e}`, a = i === void 0 ? void 0 : ot.get(i);
	if (i !== void 0 && a !== void 0) return ot.delete(i), ot.set(i, a), a;
	try {
		let a = await Xe(), o = await rt(() => a.codeToHtml(e, {
			lang: n,
			theme: r,
			...t.transformers ? { transformers: t.transformers } : {}
		}));
		return i !== void 0 && ct(i, o), o;
	} catch {
		return null;
	}
}
async function ut(e, t) {
	if (!e || !t.lang) return null;
	let n = Qe(t.lang);
	if (!n) return null;
	try {
		let r = await Xe(), { tokens: i } = await rt(() => r.codeToTokens(e, {
			lang: n,
			theme: $e(t.theme)
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
var dt = "flex w-7 shrink-0 select-none items-center justify-center self-stretch [@media(pointer:coarse)]:min-h-8 [@media(pointer:coarse)]:w-8";
function ft() {
	return /* @__PURE__ */ (0, A.jsx)("span", {
		"aria-hidden": !0,
		className: dt
	});
}
function pt({ state: e, label: t, disabled: n, onToggle: r }) {
	return /* @__PURE__ */ (0, A.jsx)("button", {
		type: "button",
		role: "checkbox",
		"aria-checked": e === "partial" ? "mixed" : e === "all",
		"aria-label": t,
		title: t,
		disabled: n,
		onClick: r,
		className: i(dt, "focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring disabled:cursor-not-allowed disabled:opacity-50"),
		children: /* @__PURE__ */ (0, A.jsxs)("span", {
			"aria-hidden": !0,
			className: i("inline-flex size-3.5 shrink-0 items-center justify-center rounded-[3px] border", e === "all" && "border-primary bg-primary text-primary-foreground", e === "partial" && "border-primary bg-background text-primary", e === "none" && "border-muted-foreground/60 bg-background"),
			children: [e === "all" && /* @__PURE__ */ (0, A.jsx)(Ee, { className: "size-3 text-current" }), e === "partial" && /* @__PURE__ */ (0, A.jsx)("span", { className: "h-0.5 w-2 rounded bg-current" })]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/token-style.ts
function mt(e) {
	let t = {};
	e.color && (t.color = e.color);
	let n = e.fontStyle ?? 0;
	return n & 1 && (t.fontStyle = "italic"), n & 2 && (t.fontWeight = 700), n & 4 && (t.textDecorationLine = "underline"), t;
}
// ../../../clicky-ui/packages/ui/src/data/HighlightedTokens.tsx
function ht({ tokens: e, content: t }) {
	return !e || e.length === 0 ? /* @__PURE__ */ (0, A.jsx)(A.Fragment, { children: t }) : /* @__PURE__ */ (0, A.jsx)(A.Fragment, { children: e.map((e, t) => /* @__PURE__ */ (0, A.jsx)("span", {
		style: mt(e),
		children: e.content
	}, t)) });
}
// ../../../clicky-ui/packages/ui/src/data/CodeDiff.tsx
var gt = [{
	id: "unified",
	label: "Unified",
	icon: ke,
	title: "Unified view"
}, {
	id: "split",
	label: "Split",
	icon: Ae,
	title: "Split view"
}], _t = {
	context: void 0,
	add: "bg-emerald-500/10",
	remove: "bg-rose-500/10"
}, vt = {
	context: "text-muted-foreground",
	add: "text-emerald-700 [[data-theme=dark]_&]:text-emerald-300",
	remove: "text-rose-700 [[data-theme=dark]_&]:text-rose-300"
}, yt = {
	context: "",
	add: "+",
	remove: "-"
};
function bt(e) {
	let { language: t, view: n = "unified", showLineNumbers: r = !0, bare: a = !1, className: o } = e, { unified: s, original: c, modified: l, originalLabel: u, modifiedLabel: d, onLineAction: f, lineWidgets: p, lineSelection: m } = e, [h, g] = (0, O.useState)(null), [_, v] = (0, O.useState)(n), y = (0, O.useMemo)(() => s === void 0 ? Le(c ?? "", l ?? "") : Ve(s), [
		s,
		c,
		l
	]), { oldSource: b, newSource: x } = (0, O.useMemo)(() => xt(y), [y]), [S, C] = (0, O.useState)({
		old: null,
		mod: null
	});
	(0, O.useEffect)(() => {
		if (!t) {
			C({
				old: null,
				mod: null
			});
			return;
		}
		let e = !1;
		return Promise.all([ut(b, { lang: t }), ut(x, { lang: t })]).then(([t, n]) => {
			e || C({
				old: t,
				mod: n
			});
		}), () => {
			e = !0;
		};
	}, [
		t,
		b,
		x
	]);
	let w = (0, O.useMemo)(() => St(y, S.old, S.mod), [y, S]), T = {
		onLineAction: f,
		lineWidgets: p,
		selection: (0, O.useMemo)(() => {
			if (!m) return;
			let e = (e) => {
				let t = y[e];
				if (!t) throw Error(`CodeDiff selection: no hunk ${e}`);
				return t.lines;
			};
			return {
				selection: m,
				onLineClick: (t, n, r) => {
					let i = e(t), a = i[n], o = a && He(a);
					if (!o) throw Error(`CodeDiff selection: line ${t}:${n} is not a changed line`);
					let s = h?.hunks === y && h.hunk === t ? h.line : void 0, c = r && s !== void 0 ? We(i, s, n) : [o];
					m.onChange(c, !m.isSelected(o)), g({
						hunks: y,
						hunk: t,
						line: n
					});
				},
				onHunkClick: (t) => {
					let n = Ue(e(t));
					m.onChange(n, Ge(n, m.isSelected) !== "all");
				}
			};
		}, [
			m,
			y,
			h
		])
	}, E = new Set(w.map((e) => e.path).filter((e) => e !== void 0)).size > 1, D, k = w.map((e, t) => {
		let n = E && e.path !== void 0 && e.path !== D ? e.path : void 0;
		return D = e.path, /* @__PURE__ */ (0, A.jsxs)(O.Fragment, { children: [n && /* @__PURE__ */ (0, A.jsx)(kt, { path: n }), _ === "split" ? /* @__PURE__ */ (0, A.jsx)(Et, {
			hunk: e,
			index: t,
			showLineNumbers: r,
			annotations: T
		}) : /* @__PURE__ */ (0, A.jsx)(Tt, {
			hunk: e,
			index: t,
			showLineNumbers: r,
			annotations: T
		})] }, t);
	});
	return a ? /* @__PURE__ */ (0, A.jsx)("div", {
		"data-scroll-x": !0,
		className: i("overflow-x-auto font-mono text-xs leading-relaxed", o),
		children: k
	}) : /* @__PURE__ */ (0, A.jsxs)("div", {
		className: i("overflow-hidden rounded-md border border-border bg-muted/40", o),
		children: [
			/* @__PURE__ */ (0, A.jsxs)("div", {
				className: "flex items-center justify-between gap-2 border-b border-border px-3 py-1.5",
				children: [/* @__PURE__ */ (0, A.jsx)("span", {
					className: "text-[11px] uppercase tracking-wide text-muted-foreground",
					children: t || "diff"
				}), /* @__PURE__ */ (0, A.jsx)(Fe, {
					size: "sm",
					"aria-label": "Diff view",
					value: _,
					onChange: v,
					options: gt
				})]
			}),
			u != null || d != null ? /* @__PURE__ */ (0, A.jsxs)("div", {
				className: i("border-b border-border text-xs font-medium", _ === "split" ? "grid grid-cols-2 divide-x divide-border" : "flex flex-wrap gap-x-4"),
				children: [/* @__PURE__ */ (0, A.jsxs)("div", {
					"data-diff-side": "original",
					className: "min-w-0 break-words px-3 py-1.5",
					children: [_ === "unified" ? "− " : null, u]
				}), /* @__PURE__ */ (0, A.jsxs)("div", {
					"data-diff-side": "modified",
					className: "min-w-0 break-words px-3 py-1.5",
					children: [_ === "unified" ? "+ " : null, d]
				})]
			}) : null,
			/* @__PURE__ */ (0, A.jsx)("div", {
				"data-scroll-x": !0,
				className: "overflow-x-auto font-mono text-xs leading-relaxed",
				children: k
			})
		]
	});
}
function xt(e) {
	let t = [], n = [];
	for (let r of e) for (let e of r.lines) e.type !== "add" && t.push(e.content), e.type !== "remove" && n.push(e.content);
	return {
		oldSource: t.join("\n"),
		newSource: n.join("\n")
	};
}
function St(e, t, n) {
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
function Ct(e, t, n) {
	return t === "old" ? e.type !== "add" && e.oldNumber === n : e.type !== "remove" && e.newNumber === n;
}
function wt({ widgets: e }) {
	return /* @__PURE__ */ (0, A.jsx)(A.Fragment, { children: e.map((e) => /* @__PURE__ */ (0, A.jsx)("div", {
		"data-diff-widget": e.key,
		className: "whitespace-normal border-y border-border bg-background px-3 py-2 font-sans text-sm leading-normal text-foreground",
		children: e.node
	}, e.key)) });
}
function Tt({ hunk: e, index: t, showLineNumbers: n, annotations: r }) {
	let { onLineAction: a, lineWidgets: o, selection: s } = r;
	return /* @__PURE__ */ (0, A.jsxs)(A.Fragment, { children: [/* @__PURE__ */ (0, A.jsx)(Nt, {
		hunk: e,
		index: t,
		selection: s
	}), e.lines.map((e, r) => {
		let { line: c } = e, l = c.type === "remove" ? e.oldTokens : e.newTokens, u = c.type === "remove" ? "old" : "new", d = o?.filter((e) => Ct(c, e.side, e.line)), f = s ? jt(s, c) : !1;
		return /* @__PURE__ */ (0, A.jsxs)(O.Fragment, { children: [/* @__PURE__ */ (0, A.jsxs)("div", {
			className: i("group flex", _t[c.type], f && At),
			"data-diff-line": c.type,
			"data-selected": f || void 0,
			children: [
				s && /* @__PURE__ */ (0, A.jsx)(Mt, {
					controls: s,
					hunk: t,
					resolved: e
				}),
				n && /* @__PURE__ */ (0, A.jsxs)(A.Fragment, { children: [/* @__PURE__ */ (0, A.jsx)(Pt, {
					value: c.oldNumber,
					line: c,
					side: "old",
					actionSide: u,
					onLineAction: a
				}), /* @__PURE__ */ (0, A.jsx)(Pt, {
					value: c.newNumber,
					line: c,
					side: "new",
					actionSide: u,
					onLineAction: a
				})] }),
				/* @__PURE__ */ (0, A.jsx)(Ft, { type: c.type }),
				/* @__PURE__ */ (0, A.jsx)("code", {
					className: "whitespace-pre pl-1 pr-3",
					children: /* @__PURE__ */ (0, A.jsx)(ht, {
						tokens: l,
						content: c.content
					})
				})
			]
		}), d && /* @__PURE__ */ (0, A.jsx)(wt, { widgets: d })] }, r);
	})] });
}
function Et({ hunk: e, index: t, showLineNumbers: n, annotations: r }) {
	let i = (0, O.useMemo)(() => Dt(e.lines), [e.lines]), { onLineAction: a, lineWidgets: o, selection: s } = r;
	return /* @__PURE__ */ (0, A.jsxs)(A.Fragment, { children: [/* @__PURE__ */ (0, A.jsx)(Nt, {
		hunk: e,
		index: t,
		selection: s
	}), i.map((e, r) => {
		let i = o?.filter((t) => e.left !== void 0 && Ct(e.left.line, t.side, t.line) && t.side === "old" || e.right !== void 0 && Ct(e.right.line, t.side, t.line) && t.side === "new");
		return /* @__PURE__ */ (0, A.jsxs)(O.Fragment, { children: [/* @__PURE__ */ (0, A.jsxs)("div", {
			className: "flex",
			children: [
				/* @__PURE__ */ (0, A.jsx)(Ot, {
					resolved: e.left,
					side: "old",
					hunk: t,
					showLineNumbers: n,
					onLineAction: a,
					selection: s
				}),
				/* @__PURE__ */ (0, A.jsx)("span", {
					"aria-hidden": !0,
					className: "w-px shrink-0 bg-border"
				}),
				/* @__PURE__ */ (0, A.jsx)(Ot, {
					resolved: e.right,
					side: "new",
					hunk: t,
					showLineNumbers: n,
					onLineAction: a,
					selection: s
				})
			]
		}), i && /* @__PURE__ */ (0, A.jsx)(wt, { widgets: i })] }, r);
	})] });
}
function Dt(e) {
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
function Ot({ resolved: e, side: t, hunk: n, showLineNumbers: r, onLineAction: a, selection: o }) {
	if (!e) return /* @__PURE__ */ (0, A.jsx)("div", { className: "min-w-0 flex-1 basis-0" });
	let { line: s } = e, c = t === "old" && s.type === "remove" || t === "new" && s.type === "add", l = c ? s.type : "context", u = t === "old" ? s.oldNumber : s.newNumber, d = t === "old" ? e.oldTokens : e.newTokens, f = c && o ? jt(o, s) : !1;
	return /* @__PURE__ */ (0, A.jsxs)("div", {
		className: i("group flex min-w-0 flex-1 basis-0", c && _t[s.type], f && At),
		"data-diff-line": l,
		"data-selected": f || void 0,
		children: [
			o && (c ? /* @__PURE__ */ (0, A.jsx)(Mt, {
				controls: o,
				hunk: n,
				resolved: e
			}) : /* @__PURE__ */ (0, A.jsx)(ft, {})),
			r && /* @__PURE__ */ (0, A.jsx)(Pt, {
				value: u,
				line: s,
				side: t,
				actionSide: t,
				onLineAction: a
			}),
			/* @__PURE__ */ (0, A.jsx)(Ft, { type: l }),
			/* @__PURE__ */ (0, A.jsx)("code", {
				className: "min-w-0 flex-1 whitespace-pre-wrap pl-1 pr-3 [overflow-wrap:anywhere]",
				children: /* @__PURE__ */ (0, A.jsx)(ht, {
					tokens: d,
					content: s.content
				})
			})
		]
	});
}
function kt({ path: e }) {
	return /* @__PURE__ */ (0, A.jsxs)("div", {
		"data-diff-file": e,
		className: "flex items-center gap-1.5 border-b border-border bg-muted px-2 py-1 font-medium text-foreground",
		children: [/* @__PURE__ */ (0, A.jsx)(je, { className: "size-3.5 shrink-0 text-muted-foreground" }), /* @__PURE__ */ (0, A.jsx)("span", {
			className: "truncate",
			children: e
		})]
	});
}
var At = "relative before:absolute before:inset-y-0 before:left-0 before:w-[3px] before:bg-emerald-500";
function jt(e, t) {
	let n = He(t);
	return n !== void 0 && e.selection.isSelected(n);
}
function Mt({ controls: e, hunk: t, resolved: n }) {
	let r = He(n.line);
	if (!r) return /* @__PURE__ */ (0, A.jsx)(ft, {});
	let { selection: i } = e, a = i.isSelected(r);
	return /* @__PURE__ */ (0, A.jsx)(pt, {
		state: a ? "all" : "none",
		label: i.label?.(r, a) ?? `${a ? "Deselect" : "Select"} ${r.side === "new" ? "added" : "removed"} line ${r.line}`,
		disabled: i.disabled === !0,
		onToggle: (r) => e.onLineClick(t, n.index, r.shiftKey)
	});
}
function Nt({ hunk: e, index: t, selection: n }) {
	if (!n) return e.header ? /* @__PURE__ */ (0, A.jsx)("div", {
		className: "select-none bg-muted/60 px-2 py-0.5 text-muted-foreground",
		children: e.header
	}) : null;
	let r = Ue(e.lines.map((e) => e.line)), i = Ge(r, n.selection.isSelected), a = i === "all", o = t + 1;
	return /* @__PURE__ */ (0, A.jsxs)("div", {
		"data-diff-hunk": o,
		className: "flex select-none items-center bg-muted/60 pr-2 text-muted-foreground",
		children: [r.length === 0 ? /* @__PURE__ */ (0, A.jsx)(ft, {}) : /* @__PURE__ */ (0, A.jsx)(pt, {
			state: i,
			label: n.selection.hunkLabel?.(o, a) ?? `${a ? "Deselect" : "Select"} hunk ${o}`,
			disabled: n.selection.disabled === !0,
			onToggle: () => n.onHunkClick(t)
		}), /* @__PURE__ */ (0, A.jsx)("span", {
			className: "min-w-0 flex-1 truncate py-0.5",
			children: e.header ?? `Hunk ${o}`
		})]
	});
}
function Pt({ value: e, line: t, side: n, actionSide: r, onLineAction: i }) {
	return /* @__PURE__ */ (0, A.jsxs)("span", {
		className: "relative w-10 shrink-0 select-none px-2 text-right tabular-nums text-muted-foreground",
		children: [e ?? "", i !== void 0 && e !== void 0 && n === r && /* @__PURE__ */ (0, A.jsx)("button", {
			type: "button",
			"aria-label": `Comment on line ${e}`,
			onClick: (r) => i({
				side: n,
				line: e,
				content: t.content,
				element: r.currentTarget
			}),
			className: "absolute inset-y-0 left-0.5 flex items-center rounded-sm text-primary opacity-0 hover:text-primary/80 focus-visible:opacity-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring group-hover:opacity-100",
			children: /* @__PURE__ */ (0, A.jsx)(xe, { className: "size-3.5" })
		})]
	});
}
function Ft({ type: e }) {
	return /* @__PURE__ */ (0, A.jsx)("span", {
		className: i("w-4 shrink-0 select-none text-center", vt[e]),
		children: yt[e]
	});
}
// ../../../clicky-ui/packages/ui/src/lib/breakpoints.ts
var It = /* @__PURE__ */ e(r(), 1), Lt = {
	sm: 640,
	md: 768,
	lg: 1024,
	xl: 1280,
	"2xl": 1536,
	"3xl": 1920
};
function Rt(e) {
	return `(max-width: ${Lt[e] - 1}px)`;
}
// ../../../clicky-ui/packages/ui/src/overlay/zIndex.ts
var zt = {
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
}, Bt = [], Vt = /* @__PURE__ */ new Set(), Ht = [], Ut = null, Wt = 0;
function Gt() {
	for (let e of Vt) e();
}
function Kt(e) {
	return Vt.add(e), () => {
		Vt.delete(e);
	};
}
function qt() {
	return `${Wt}:` + Bt.join("\0");
}
function Jt(e) {
	if (e.defaultPrevented || e.key !== "Escape" || e.isComposing || e.keyCode === 229) return;
	let t = Ht.findLast((e) => !e.dismissed);
	t && (e.preventDefault(), e.stopPropagation(), e.stopImmediatePropagation(), t.dismissed = !0, t.onEscape());
}
function Yt() {
	Ut || typeof document > "u" || (Ut = document, Ut.addEventListener("keydown", Jt));
}
function Xt(e) {
	let t = Ht.findIndex((t) => t.id === e);
	t !== -1 && Ht.splice(t, 1), Ht.length === 0 && Ut && (Ut.removeEventListener("keydown", Jt), Ut = null);
}
function Zt(e) {
	let t = (0, O.useId)();
	(0, O.useSyncExternalStore)(Kt, qt, qt), (0, O.useEffect)(() => {
		if (e) return Bt.push(t), Gt(), () => {
			let e = Bt.indexOf(t);
			e !== -1 && Bt.splice(e, 1), Gt();
		};
	}, [e, t]);
	let n = Bt.indexOf(t);
	return {
		isTop: n !== -1 && n === Bt.length - 1,
		depth: n === -1 ? 0 : n
	};
}
function Qt(e, t, n = !0) {
	let r = (0, O.useId)(), i = (0, O.useRef)(t);
	i.current = t;
	let a = (0, O.useRef)(null);
	(0, O.useEffect)(() => {
		a.current && (a.current.dismissed = !1);
	}), (0, O.useEffect)(() => {
		if (!e || !n) return;
		Yt();
		let t = {
			id: r,
			dismissed: !1,
			onEscape: () => i.current()
		};
		return a.current = t, Ht.push(t), () => {
			a.current = null, Xt(r);
		};
	}, [
		n,
		r,
		e
	]);
}
function $t() {
	(0, O.useSyncExternalStore)(Kt, qt, qt);
	let e = Bt.length === 0 ? zt.popover : zt.modal + (Bt.length - 1) * zt.modalStep + zt.popoverOverModalOffset;
	return Wt === 0 ? e : Math.max(e, zt.tour + zt.tourCardOffset + zt.popoverOverModalOffset);
}
// ../../../clicky-ui/packages/ui/src/overlay/Modal.tsx
var en = {
	sm: "max-w-sm",
	md: "max-w-md",
	lg: "max-w-2xl",
	xl: "max-w-4xl",
	"2xl": "max-w-6xl",
	full: "max-w-[95vw]"
}, tn = {
	title: "Discard changes?",
	message: "You have unsaved changes. Closing now will discard them.",
	confirmLabel: "Discard",
	cancelLabel: "Keep editing"
}, nn = Rt("sm"), rn = "(prefers-reduced-motion: reduce)", an = 200, on = "a[href],button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),[tabindex]:not([tabindex='-1'])";
function sn(e) {
	return Array.from(e.querySelectorAll(on)).filter((e) => {
		let t = window.getComputedStyle(e);
		return !e.hidden && t.display !== "none" && t.visibility !== "hidden";
	});
}
function cn({ open: e, minimized: t = !1, onClose: n, confirmClose: r = !1, title: a, "aria-label": u, subtitle: d, size: f = "md", closeOnBackdrop: p = !1, closeOnEsc: m = !0, hideClose: h = !1, expandable: g = !0, className: _, headerSlot: v, footer: y, error: b, scrollBody: x = !0, children: S }) {
	let C = (0, O.useRef)(null), w = (0, O.useRef)(null), [T, E] = (0, O.useState)();
	(0, O.useLayoutEffect)(() => {
		let e = w.current?.textContent?.replace(/\s+/g, " ").trim() || void 0;
		e !== T && E(e);
	});
	let D = (0, O.useRef)(void 0), k = (0, O.useRef)(void 0), [j, M] = (0, O.useState)(e), [N, P] = (0, O.useState)(!1), F = j && !t, I = e && F, { depth: L } = Zt(F), [R, ee] = (0, O.useState)(!1), [z, B] = (0, O.useState)(!1), V = () => {
		r ? B(!0) : n();
	};
	if ((0, O.useEffect)(() => {
		window.clearTimeout(D.current), window.cancelAnimationFrame(k.current ?? 0);
		let t = l(nn) && !l(rn);
		return e ? (M(!0), t ? (P(!1), k.current = window.requestAnimationFrame(() => {
			P(!0);
		})) : P(!0)) : t ? (P(!1), D.current = window.setTimeout(() => M(!1), an)) : (P(!1), M(!1)), () => {
			window.clearTimeout(D.current), window.cancelAnimationFrame(k.current ?? 0);
		};
	}, [e]), (0, O.useEffect)(() => {
		e || (ee(!1), B(!1));
	}, [e]), (0, O.useEffect)(() => {
		r || B(!1);
	}, [r]), Qt(F, () => {
		e && m && (z ? B(!1) : V());
	}), (0, O.useLayoutEffect)(() => {
		if (!I) return;
		let e = document.activeElement;
		return C.current?.focus({ preventScroll: !0 }), () => {
			e instanceof HTMLElement && e.isConnected && e.focus({ preventScroll: !0 });
		};
	}, [I]), !F) return null;
	let H = /* @__PURE__ */ (0, A.jsx)("div", {
		className: i("fixed inset-0 flex items-center justify-center p-density-2 sm:p-density-4 max-sm:items-stretch max-sm:justify-end max-sm:p-0 max-sm:!bg-transparent", L === 0 ? "bg-black/40" : "bg-black/20"),
		style: { zIndex: zt.modal + L * zt.modalStep },
		onClick: p ? V : void 0,
		role: "presentation",
		children: /* @__PURE__ */ (0, A.jsxs)("div", {
			ref: C,
			tabIndex: -1,
			role: "dialog",
			"aria-modal": "true",
			"aria-label": u ?? (typeof a == "string" ? a : T),
			"data-state": N ? "open" : "closed",
			className: i("relative flex w-full flex-col overflow-hidden rounded-lg border border-border bg-background shadow-xl", "max-sm:!h-dvh max-sm:!max-h-dvh max-sm:!max-w-none max-sm:!rounded-none max-sm:!border-0 max-sm:shadow-none", "max-sm:transition-transform max-sm:duration-200 max-sm:ease-out max-sm:motion-reduce:transition-none", "max-sm:data-[state=closed]:translate-x-full max-sm:data-[state=open]:translate-x-0", R ? en.full : en[f], _),
			style: {
				maxHeight: "calc(100dvh - 2rem)",
				...R ? { height: "calc(100dvh - 2rem)" } : {}
			},
			onClick: (e) => e.stopPropagation(),
			onKeyDown: (e) => {
				if (e.key !== "Tab") return;
				let t = z ? C.current?.querySelector("[role='alertdialog']") : C.current;
				if (!t) return;
				let n = sn(t), r = n[0], i = n[n.length - 1];
				!r || !i ? (e.preventDefault(), t.focus()) : e.shiftKey && document.activeElement === r ? (e.preventDefault(), i.focus()) : !e.shiftKey && document.activeElement === i && (e.preventDefault(), r.focus());
			},
			children: [
				(a || d || v || g || !h) && /* @__PURE__ */ (0, A.jsxs)("div", {
					className: "shrink-0 px-density-4 py-density-3 border-b border-border max-sm:min-h-14 max-sm:bg-background",
					children: [/* @__PURE__ */ (0, A.jsxs)("div", {
						className: "flex items-center gap-density-2",
						children: [
							!h && /* @__PURE__ */ (0, A.jsx)(o, {
								type: "button",
								onClick: V,
								variant: "ghost",
								size: "icon",
								"aria-label": "Back",
								className: "-ml-2 text-foreground sm:hidden",
								children: /* @__PURE__ */ (0, A.jsx)(s, { icon: Te })
							}),
							a ? /* @__PURE__ */ (0, A.jsx)("h2", {
								ref: w,
								className: "text-sm font-semibold flex-1 max-sm:text-base max-sm:leading-6",
								children: a
							}) : /* @__PURE__ */ (0, A.jsx)("span", { className: "flex-1" }),
							v,
							g && /* @__PURE__ */ (0, A.jsx)("button", {
								type: "button",
								onClick: () => ee((e) => !e),
								"aria-label": R ? "Restore size" : "Expand to fullscreen",
								className: "text-muted-foreground hover:text-foreground max-sm:hidden",
								children: /* @__PURE__ */ (0, A.jsx)(s, { icon: R ? Ce : Se })
							}),
							!h && /* @__PURE__ */ (0, A.jsx)("button", {
								type: "button",
								onClick: V,
								"aria-label": "Close",
								className: "text-muted-foreground hover:text-foreground max-sm:hidden",
								children: /* @__PURE__ */ (0, A.jsx)(s, { icon: c })
							})
						]
					}), d ? /* @__PURE__ */ (0, A.jsx)("div", {
						className: "mt-density-2",
						children: d
					}) : null]
				}),
				/* @__PURE__ */ (0, A.jsx)("div", {
					"data-slot": "modal-body",
					className: i("min-h-0 flex-1 px-density-4 py-density-3 max-sm:py-density-4", x ? "overflow-auto" : "flex flex-col overflow-hidden"),
					children: S
				}),
				b != null && b !== !1 && b !== "" && /* @__PURE__ */ (0, A.jsx)("div", {
					role: "alert",
					"data-slot": "modal-error",
					className: "mx-density-4 mb-density-3 max-h-32 shrink-0 overflow-auto rounded-md border border-destructive/40 bg-destructive/5 px-density-3 py-density-2 text-sm wrap-break-word text-destructive",
					children: b
				}),
				y && /* @__PURE__ */ (0, A.jsx)("div", {
					"data-slot": "modal-footer",
					className: "shrink-0 px-density-4 py-density-3 border-t border-border max-sm:py-density-4",
					children: y
				}),
				z && /* @__PURE__ */ (0, A.jsx)(ln, {
					options: r === !0 ? tn : {
						...tn,
						...r
					},
					onConfirm: () => {
						B(!1), n();
					},
					onCancel: () => B(!1)
				})
			]
		})
	});
	return typeof document < "u" ? (0, It.createPortal)(H, document.body) : null;
}
function ln({ options: e, onConfirm: t, onCancel: n }) {
	return /* @__PURE__ */ (0, A.jsx)("div", {
		className: "absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-background/80 p-density-4 max-sm:rounded-none",
		role: "alertdialog",
		"aria-modal": "true",
		"aria-label": typeof e.title == "string" ? e.title : "Confirm close",
		onClick: (e) => e.stopPropagation(),
		children: /* @__PURE__ */ (0, A.jsxs)("div", {
			className: "w-full max-w-sm rounded-lg border border-border bg-background p-density-4 shadow-xl",
			children: [
				/* @__PURE__ */ (0, A.jsx)("h3", {
					className: "text-sm font-semibold",
					children: e.title
				}),
				/* @__PURE__ */ (0, A.jsx)("p", {
					className: "mt-density-2 text-sm text-muted-foreground",
					children: e.message
				}),
				/* @__PURE__ */ (0, A.jsxs)("div", {
					className: "mt-density-4 flex justify-end gap-density-2",
					children: [/* @__PURE__ */ (0, A.jsx)(o, {
						variant: "ghost",
						size: "sm",
						onClick: n,
						children: e.cancelLabel
					}), /* @__PURE__ */ (0, A.jsx)(o, {
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
var un = "input:not([inert]):not([inert] *),select:not([inert]):not([inert] *),textarea:not([inert]):not([inert] *),a[href]:not([inert]):not([inert] *),area[href]:not([inert]):not([inert] *),button:not([inert]):not([inert] *),[tabindex]:not(slot):not([inert]):not([inert] *),audio[controls]:not([inert]):not([inert] *),video[controls]:not([inert]):not([inert] *),[contenteditable]:not([contenteditable=\"false\"]):not([inert]):not([inert] *),details>summary:first-of-type:not([inert]):not([inert] *),details:not([inert]):not([inert] *)", dn = typeof Element > "u", fn = dn ? function() {} : Element.prototype.matches || Element.prototype.msMatchesSelector || Element.prototype.webkitMatchesSelector, pn = !dn && Element.prototype.getRootNode ? function(e) {
	return e == null || e.getRootNode == null ? void 0 : e.getRootNode();
} : function(e) {
	return e?.ownerDocument;
}, mn = function(e, t) {
	t === void 0 && (t = !0);
	var n = e == null || e.getAttribute == null ? void 0 : e.getAttribute("inert");
	return n === "" || n === "true" || t && e && (typeof e.closest == "function" ? e.closest("[inert]") : mn(e.parentNode));
}, hn = function(e) {
	var t, n = e == null || e.getAttribute == null || (t = e.getAttribute("contenteditable")) == null ? void 0 : t.toLowerCase();
	return n === "" || n === "true" || n === "plaintext-only";
}, gn = function(e, t, n) {
	if (mn(e)) return [];
	var r = Array.prototype.slice.apply(e.querySelectorAll(un));
	return t && fn.call(e, un) && r.unshift(e), r = r.filter(n), r;
}, _n = function(e, t, n) {
	for (var r = [], i = Array.from(e); i.length;) {
		var a = i.shift();
		if (!mn(a, !1)) {
			if (a.tagName === "SLOT") {
				var o = a.assignedElements(), s = _n(o.length ? o : a.children, !0, n);
				n.flatten ? r.push.apply(r, s) : r.push({
					scopeParent: a,
					candidates: s
				});
			} else {
				fn.call(a, un) && n.filter(a) && (t || !e.includes(a)) && r.push(a);
				var c = a.shadowRoot || typeof n.getShadowRoot == "function" && n.getShadowRoot(a), l = !mn(c, !1) && (!n.shadowRootFilter || n.shadowRootFilter(a));
				if (c && l) {
					var u = _n(c === !0 ? a.children : c.children, !0, n);
					n.flatten ? r.push.apply(r, u) : r.push({
						scopeParent: a,
						candidates: u
					});
				} else i.unshift.apply(i, a.children);
			}
		}
	}
	return r;
}, vn = function(e) {
	return !isNaN(parseInt(e.getAttribute("tabindex"), 10));
}, yn = function(e) {
	if (!e) throw Error("No node provided");
	return e.tabIndex < 0 && (/^(AUDIO|VIDEO|DETAILS)$/.test(e.tagName) || hn(e)) && !vn(e) ? 0 : e.tabIndex;
}, bn = function(e, t) {
	var n = yn(e);
	return n < 0 && t && !vn(e) ? 0 : n;
}, xn = function(e, t) {
	return e.tabIndex === t.tabIndex ? e.documentOrder - t.documentOrder : e.tabIndex - t.tabIndex;
}, Sn = function(e) {
	return e.tagName === "INPUT";
}, Cn = function(e) {
	return Sn(e) && e.type === "hidden";
}, wn = function(e) {
	return e.tagName === "DETAILS" && Array.prototype.slice.apply(e.children).some(function(e) {
		return e.tagName === "SUMMARY";
	});
}, Tn = function(e, t) {
	for (var n = 0; n < e.length; n++) if (e[n].checked && e[n].form === t) return e[n];
}, En = function(e) {
	if (!e.name) return !0;
	var t = e.form || pn(e), n = function(e) {
		return t.querySelectorAll("input[type=\"radio\"][name=\"" + e + "\"]");
	}, r;
	if (typeof window < "u" && window.CSS !== void 0 && typeof window.CSS.escape == "function") r = n(window.CSS.escape(e.name));
	else try {
		r = n(e.name);
	} catch (e) {
		return console.error("Looks like you have a radio button with a name attribute containing invalid CSS selector characters and need the CSS.escape polyfill: %s", e.message), !1;
	}
	var i = Tn(r, e.form);
	return !i || i === e;
}, Dn = function(e) {
	return Sn(e) && e.type === "radio";
}, On = function(e) {
	return Dn(e) && !En(e);
}, kn = function(e) {
	var t = e && pn(e), n = t?.host, r = !1;
	if (t && t !== e) {
		var i, a;
		for (r = !!((i = n) != null && (i = i.ownerDocument) != null && i.contains(n) || e != null && (a = e.ownerDocument) != null && a.contains(e)); !r && n;) {
			var o;
			t = pn(n), n = t?.host, r = !!((o = n) != null && (o = o.ownerDocument) != null && o.contains(n));
		}
	}
	return r;
}, An = function(e) {
	var t = e.getBoundingClientRect(), n = t.width, r = t.height;
	return n === 0 && r === 0;
}, jn = function(e, t) {
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
	var a = fn.call(e, "details>summary:first-of-type") ? e.parentElement : e;
	if (fn.call(a, "details:not([open]) *")) return !0;
	if (!n || n === "full" || n === "full-native" || n === "legacy-full") {
		if (typeof r == "function") {
			for (var o = e; e;) {
				var s = e.parentElement, c = pn(e);
				if (s && !s.shadowRoot && r(s) === !0) return An(e);
				e = e.assignedSlot ? e.assignedSlot : !s && c !== e.ownerDocument ? c.host : s;
			}
			e = o;
		}
		if (kn(e)) return !e.getClientRects().length;
		if (n !== "legacy-full") return !0;
	} else if (n === "non-zero-area") return An(e);
	return !1;
}, Mn = function(e) {
	if (/^(INPUT|BUTTON|SELECT|TEXTAREA)$/.test(e.tagName)) for (var t = e.parentElement; t;) {
		if (t.tagName === "FIELDSET" && t.disabled) {
			for (var n = 0; n < t.children.length; n++) {
				var r = t.children.item(n);
				if (r.tagName === "LEGEND") return fn.call(t, "fieldset[disabled] *") ? !0 : !r.contains(e);
			}
			return !0;
		}
		t = t.parentElement;
	}
	return !1;
}, Nn = function(e, t) {
	return !(t.disabled || Cn(t) || jn(t, e) || wn(t) || Mn(t));
}, Pn = function(e, t) {
	return !(On(t) || yn(t) < 0 || !Nn(e, t));
}, Fn = function(e) {
	var t = parseInt(e.getAttribute("tabindex"), 10);
	return !!(isNaN(t) || t >= 0);
}, In = function(e) {
	var t = [], n = [];
	return e.forEach(function(e, r) {
		var i = !!e.scopeParent, a = i ? e.scopeParent : e, o = bn(a, i), s = i ? In(e.candidates) : a;
		o === 0 ? i ? t.push.apply(t, s) : t.push(a) : n.push({
			documentOrder: r,
			tabIndex: o,
			item: e,
			isScope: i,
			content: s
		});
	}), n.sort(xn).reduce(function(e, t) {
		return t.isScope ? e.push.apply(e, t.content) : e.push(t.content), e;
	}, []).concat(t);
}, Ln = function(e, t) {
	return t ||= {}, In(t.getShadowRoot ? _n([e], t.includeContainer, {
		filter: Pn.bind(null, t),
		flatten: !1,
		getShadowRoot: t.getShadowRoot,
		shadowRootFilter: Fn
	}) : gn(e, t.includeContainer, Pn.bind(null, t)));
}, Rn = function(e, t) {
	return t ||= {}, t.getShadowRoot ? _n([e], t.includeContainer, {
		filter: Nn.bind(null, t),
		flatten: !0,
		getShadowRoot: t.getShadowRoot
	}) : gn(e, t.includeContainer, Nn.bind(null, t));
}, zn = function(e, t) {
	if (t ||= {}, !e) throw Error("No node provided");
	return fn.call(e, un) !== !1 && Pn(t, e);
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@floating-ui/react/0.27.20/a5461da18b9b5ae2ae73018151621914115cfa402d046f06e6fc049e6b6c54d2/node_modules/@floating-ui/react/dist/floating-ui.react.utils.mjs
function Bn() {
	let e = navigator.userAgentData;
	return e != null && e.platform ? e.platform : navigator.platform;
}
function Vn() {
	let e = navigator.userAgentData;
	return e && Array.isArray(e.brands) ? e.brands.map((e) => {
		let { brand: t, version: n } = e;
		return t + "/" + n;
	}).join(" ") : navigator.userAgent;
}
function Hn() {
	return /apple/i.test(navigator.vendor);
}
function Un() {
	let e = /android/i;
	return e.test(Bn()) || e.test(Vn());
}
function Wn() {
	return Vn().includes("jsdom/");
}
var Gn = "data-floating-ui-focusable", Kn = "input:not([type='hidden']):not([disabled]),[contenteditable]:not([contenteditable='false']),textarea:not([disabled])", qn = "ArrowLeft", Jn = "ArrowRight", Yn = "ArrowUp", Xn = "ArrowDown";
function Zn(e) {
	let t = e.activeElement;
	for (; ((n = t) == null || (n = n.shadowRoot) == null ? void 0 : n.activeElement) != null;) {
		var n;
		t = t.shadowRoot.activeElement;
	}
	return t;
}
function Qn(e, t) {
	if (!e || !t) return !1;
	let n = t.getRootNode == null ? void 0 : t.getRootNode();
	if (e.contains(t)) return !0;
	if (n && D(n)) {
		let n = t;
		for (; n;) {
			if (e === n) return !0;
			n = n.parentNode || n.host;
		}
	}
	return !1;
}
function $n(e) {
	return "composedPath" in e ? e.composedPath()[0] : e.target;
}
function er(e, t) {
	if (t == null) return !1;
	if ("composedPath" in e) return e.composedPath().includes(t);
	let n = e;
	return n.target != null && t.contains(n.target);
}
function tr(e) {
	return e.matches("html,body");
}
function J(e) {
	return e?.ownerDocument || document;
}
function nr(e) {
	return _(e) && e.matches(Kn);
}
function rr(e) {
	return e ? e.getAttribute("role") === "combobox" && nr(e) : !1;
}
function ir(e) {
	return e ? e.hasAttribute(Gn) ? e : e.querySelector("[" + Gn + "]") || e : null;
}
function ar(e, t, n) {
	return n === void 0 && (n = !0), e.filter((e) => e.parentId === t && (!n || e.context?.open)).flatMap((t) => [t, ...ar(e, t.id, n)]);
}
function or(e, t) {
	let n, r = -1;
	function i(t, a) {
		a > r && (n = t, r = a), ar(e, t).forEach((e) => {
			i(e.id, a + 1);
		});
	}
	return i(t, 0), e.find((e) => e.id === n);
}
function sr(e, t) {
	let n = [], r = e.find((e) => e.id === t)?.parentId;
	for (; r;) {
		let t = e.find((e) => e.id === r);
		r = t?.parentId, t && (n = n.concat(t));
	}
	return n;
}
function Y(e) {
	e.preventDefault(), e.stopPropagation();
}
function cr(e) {
	return "nativeEvent" in e;
}
function lr(e) {
	return e.mozInputSource === 0 && e.isTrusted ? !0 : Un() && e.pointerType ? e.type === "click" && e.buttons === 1 : e.detail === 0 && !e.pointerType;
}
function ur(e) {
	return Wn() ? !1 : !Un() && e.width === 0 && e.height === 0 || Un() && e.width === 1 && e.height === 1 && e.pressure === 0 && e.detail === 0 && e.pointerType === "mouse" || e.width < 1 && e.height < 1 && e.pressure === 0 && e.detail === 0 && e.pointerType === "touch";
}
function dr(e, t) {
	let n = ["mouse", "pen"];
	return t || n.push("", void 0), n.includes(e);
}
var X = typeof document < "u" ? O.useLayoutEffect : function() {}, fr = { ...O };
function pr(e) {
	let t = O.useRef(e);
	return X(() => {
		t.current = e;
	}), t;
}
var mr = fr.useInsertionEffect || ((e) => e());
function Z(e) {
	let t = O.useRef(() => {});
	return mr(() => {
		t.current = e;
	}), O.useCallback(function() {
		var e = [...arguments];
		return t.current == null ? void 0 : t.current(...e);
	}, []);
}
function hr(e, t, n) {
	return Math.floor(e / t) !== n;
}
function gr(e, t) {
	return t < 0 || t >= e.current.length;
}
function _r(e, t) {
	return Q(e, { disabledIndices: t });
}
function vr(e, t) {
	return Q(e, {
		decrement: !0,
		startingIndex: e.current.length,
		disabledIndices: t
	});
}
function Q(e, t) {
	let { startingIndex: n = -1, decrement: r = !1, disabledIndices: i, amount: a = 1 } = t === void 0 ? {} : t, o = n;
	do
		o += r ? -a : a;
	while (o >= 0 && o <= e.current.length - 1 && Cr(e, o, i));
	return o;
}
function yr(e, t) {
	let { event: n, orientation: r, loop: i, rtl: a, cols: o, disabledIndices: s, minIndex: c, maxIndex: l, prevIndex: u, stopEvent: d = !1 } = t, f = u;
	if (n.key === Yn) {
		if (d && Y(n), u === -1) f = l;
		else if (f = Q(e, {
			startingIndex: f,
			amount: o,
			decrement: !0,
			disabledIndices: s
		}), i && (u - o < c || f < 0)) {
			let e = u % o, t = l % o, n = l - (t - e);
			f = t === e ? l : t > e ? n : n - o;
		}
		gr(e, f) && (f = u);
	}
	if (n.key === Xn && (d && Y(n), u === -1 ? f = c : (f = Q(e, {
		startingIndex: u,
		amount: o,
		disabledIndices: s
	}), i && u + o > l && (f = Q(e, {
		startingIndex: u % o - o,
		amount: o,
		disabledIndices: s
	}))), gr(e, f) && (f = u)), r === "both") {
		let t = T(u / o);
		n.key === (a ? qn : Jn) && (d && Y(n), u % o === o - 1 ? i && (f = Q(e, {
			startingIndex: u - u % o - 1,
			disabledIndices: s
		})) : (f = Q(e, {
			startingIndex: u,
			disabledIndices: s
		}), i && hr(f, o, t) && (f = Q(e, {
			startingIndex: u - u % o - 1,
			disabledIndices: s
		}))), hr(f, o, t) && (f = u)), n.key === (a ? Jn : qn) && (d && Y(n), u % o === 0 ? i && (f = Q(e, {
			startingIndex: u + (o - u % o),
			decrement: !0,
			disabledIndices: s
		})) : (f = Q(e, {
			startingIndex: u,
			decrement: !0,
			disabledIndices: s
		}), i && hr(f, o, t) && (f = Q(e, {
			startingIndex: u + (o - u % o),
			decrement: !0,
			disabledIndices: s
		}))), hr(f, o, t) && (f = u));
		let r = T(l / o) === t;
		gr(e, f) && (f = i && r ? n.key === (a ? Jn : qn) ? l : Q(e, {
			startingIndex: u - u % o - 1,
			disabledIndices: s
		}) : u);
	}
	return f;
}
function br(e, t, n) {
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
function xr(e, t, n, r, i) {
	if (e === -1) return -1;
	let a = n.indexOf(e), o = t[e];
	switch (i) {
		case "tl": return a;
		case "tr": return o ? a + o.width - 1 : a;
		case "bl": return o ? a + (o.height - 1) * r : a;
		case "br": return n.lastIndexOf(e);
	}
}
function Sr(e, t) {
	return t.flatMap((t, n) => e.includes(t) ? [n] : []);
}
function Cr(e, t, n) {
	if (typeof n == "function") return n(t);
	if (n) return n.includes(t);
	let r = e.current[t];
	return r == null || r.hasAttribute("disabled") || r.getAttribute("aria-disabled") === "true";
}
var wr = () => ({
	getShadowRoot: !0,
	displayCheck: typeof ResizeObserver == "function" && ResizeObserver.toString().includes("[native code]") ? "full" : "none"
});
function Tr(e, t) {
	let n = Ln(e, wr()), r = n.length;
	if (r === 0) return;
	let i = Zn(J(e)), a = n.indexOf(i);
	return n[a === -1 ? t === 1 ? 0 : r - 1 : a + t];
}
function Er(e) {
	return Tr(J(e).body, 1) || e;
}
function Dr(e) {
	return Tr(J(e).body, -1) || e;
}
function Or(e, t) {
	let n = t || e.currentTarget, r = e.relatedTarget;
	return !r || !Qn(n, r);
}
function kr(e) {
	Ln(e, wr()).forEach((e) => {
		e.dataset.tabindex = e.getAttribute("tabindex") || "", e.setAttribute("tabindex", "-1");
	});
}
function Ar(e) {
	e.querySelectorAll("[data-tabindex]").forEach((e) => {
		let t = e.dataset.tabindex;
		delete e.dataset.tabindex, t ? e.setAttribute("tabindex", t) : e.removeAttribute("tabindex");
	});
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@floating-ui/react/0.27.20/a5461da18b9b5ae2ae73018151621914115cfa402d046f06e6fc049e6b6c54d2/node_modules/@floating-ui/react/dist/floating-ui.react.mjs
function jr(e) {
	let t = O.useRef(void 0), n = O.useCallback((t) => {
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
	return O.useMemo(() => e.every((e) => e == null) ? null : (e) => {
		t.current &&= (t.current(), void 0), e != null && (t.current = n(e));
	}, e);
}
function Mr(e, t) {
	let n = e.compareDocumentPosition(t);
	return n & Node.DOCUMENT_POSITION_FOLLOWING || n & Node.DOCUMENT_POSITION_CONTAINED_BY ? -1 : n & Node.DOCUMENT_POSITION_PRECEDING || n & Node.DOCUMENT_POSITION_CONTAINS ? 1 : 0;
}
var Nr = /*#__PURE__*/ O.createContext({
	register: () => {},
	unregister: () => {},
	map: /*#__PURE__*/ new Map(),
	elementsRef: { current: [] }
});
function Pr(e) {
	let { children: t, elementsRef: n, labelsRef: r } = e, [i, a] = O.useState(() => /* @__PURE__ */ new Set()), o = O.useCallback((e) => {
		a((t) => new Set(t).add(e));
	}, []), s = O.useCallback((e) => {
		a((t) => {
			let n = new Set(t);
			return n.delete(e), n;
		});
	}, []), c = O.useMemo(() => {
		let e = /* @__PURE__ */ new Map();
		return Array.from(i.keys()).sort(Mr).forEach((t, n) => {
			e.set(t, n);
		}), e;
	}, [i]);
	return /*#__PURE__*/ (0, A.jsx)(Nr.Provider, {
		value: O.useMemo(() => ({
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
function Fr(e) {
	e === void 0 && (e = {});
	let { label: t } = e, { register: n, unregister: r, map: i, elementsRef: a, labelsRef: o } = O.useContext(Nr), [s, c] = O.useState(null), l = O.useRef(null), u = O.useCallback((e) => {
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
	return X(() => {
		let e = l.current;
		if (e) return n(e), () => {
			r(e);
		};
	}, [n, r]), X(() => {
		let e = l.current ? i.get(l.current) : null;
		e != null && c(e);
	}, [i]), O.useMemo(() => ({
		ref: u,
		index: s ?? -1
	}), [s, u]);
}
var Ir = "data-floating-ui-focusable", Lr = "active", Rr = "selected", zr = "ArrowLeft", Br = "ArrowRight", Vr = "ArrowUp", Hr = "ArrowDown", Ur = [zr, Br], Wr = [Vr, Hr];
[...Ur, ...Wr];
var Gr = { ...O }, Kr = !1, qr = 0, Jr = () => "floating-ui-" + Math.random().toString(36).slice(2, 6) + qr++;
function Yr() {
	let [e, t] = O.useState(() => Kr ? Jr() : void 0);
	return X(() => {
		e ?? t(Jr());
	}, []), O.useEffect(() => {
		Kr = !0;
	}, []), e;
}
var Xr = Gr.useId || Yr;
function Zr() {
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
var Qr = /*#__PURE__*/ O.createContext(null), $r = /*#__PURE__*/ O.createContext(null), ei = () => O.useContext(Qr)?.id || null, ti = () => O.useContext($r);
function ni(e) {
	let t = Xr(), n = ti(), r = ei(), i = e || r;
	return X(() => {
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
function ri(e) {
	let { children: t, id: n } = e, r = ei();
	return /*#__PURE__*/ (0, A.jsx)(Qr.Provider, {
		value: O.useMemo(() => ({
			id: n,
			parentId: r
		}), [n, r]),
		children: t
	});
}
function ii(e) {
	let { children: t } = e, n = O.useRef([]), r = O.useCallback((e) => {
		n.current = [...n.current, e];
	}, []), i = O.useCallback((e) => {
		n.current = n.current.filter((t) => t !== e);
	}, []), [a] = O.useState(() => Zr());
	return /*#__PURE__*/ (0, A.jsx)($r.Provider, {
		value: O.useMemo(() => ({
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
function ai(e) {
	return "data-floating-ui-" + e;
}
function $(e) {
	e.current !== -1 && (clearTimeout(e.current), e.current = -1);
}
var oi = /*#__PURE__*/ ai("safe-polygon");
function si(e, t, n) {
	if (n && !dr(n)) return 0;
	if (typeof e == "number") return e;
	if (typeof e == "function") {
		let n = e();
		return typeof n == "number" ? n : n?.[t];
	}
	return e?.[t];
}
function ci(e) {
	return typeof e == "function" ? e() : e;
}
function li(e, t) {
	t === void 0 && (t = {});
	let { open: n, onOpenChange: r, dataRef: i, events: a, elements: o } = e, { enabled: s = !0, delay: c = 0, handleClose: l = null, mouseOnly: u = !1, restMs: d = 0, move: f = !0 } = t, p = ti(), m = ei(), h = pr(l), g = pr(c), _ = pr(n), y = pr(d), b = O.useRef(), x = O.useRef(-1), S = O.useRef(), C = O.useRef(-1), w = O.useRef(!0), T = O.useRef(!1), E = O.useRef(() => {}), D = O.useRef(!1), k = Z(() => {
		let e = i.current.openEvent?.type;
		return e?.includes("mouse") && e !== "mousedown";
	});
	O.useEffect(() => {
		if (!s) return;
		function e(e) {
			let { open: t } = e;
			t || ($(x), $(C), w.current = !0, D.current = !1);
		}
		return a.on("openchange", e), () => {
			a.off("openchange", e);
		};
	}, [s, a]), O.useEffect(() => {
		if (!s || !h.current || !n) return;
		function e(e) {
			k() && r(!1, e, "hover");
		}
		let t = J(o.floating).documentElement;
		return t.addEventListener("mouseleave", e), () => {
			t.removeEventListener("mouseleave", e);
		};
	}, [
		o.floating,
		n,
		r,
		s,
		h,
		k
	]);
	let A = O.useCallback(function(e, t, n) {
		t === void 0 && (t = !0), n === void 0 && (n = "hover");
		let i = si(g.current, "close", b.current);
		i && !S.current ? ($(x), x.current = window.setTimeout(() => r(!1, e, n), i)) : t && ($(x), r(!1, e, n));
	}, [g, r]), j = Z(() => {
		E.current(), S.current = void 0;
	}), M = Z(() => {
		if (T.current) {
			let e = J(o.floating).body;
			e.style.pointerEvents = "", e.removeAttribute(oi), T.current = !1;
		}
	}), N = Z(() => i.current.openEvent ? ["click", "mousedown"].includes(i.current.openEvent.type) : !1);
	O.useEffect(() => {
		if (!s) return;
		function e(e) {
			if ($(x), w.current = !1, u && !dr(b.current) || ci(y.current) > 0 && !si(g.current, "open")) return;
			let t = si(g.current, "open", b.current);
			t ? x.current = window.setTimeout(() => {
				_.current || r(!0, e, "hover");
			}, t) : n || r(!0, e, "hover");
		}
		function t(e) {
			if (N()) {
				M();
				return;
			}
			E.current();
			let t = J(o.floating);
			if ($(C), D.current = !1, h.current && i.current.floatingContext) {
				n || $(x), S.current = h.current({
					...i.current.floatingContext,
					tree: p,
					x: e.clientX,
					y: e.clientY,
					onClose() {
						M(), j(), N() || A(e, !0, "safe-polygon");
					}
				});
				let r = S.current;
				t.addEventListener("mousemove", r), E.current = () => {
					t.removeEventListener("mousemove", r);
				};
			} else b.current === "touch" && Qn(o.floating, e.relatedTarget) || A(e);
		}
		function a(e) {
			N() || i.current.floatingContext && (h.current == null || h.current({
				...i.current.floatingContext,
				tree: p,
				x: e.clientX,
				y: e.clientY,
				onClose() {
					M(), j(), N() || A(e);
				}
			})(e));
		}
		function c() {
			$(x);
		}
		function l(e) {
			N() || A(e, !1);
		}
		if (v(o.domReference)) {
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
		A,
		j,
		M,
		r,
		n,
		_,
		p,
		g,
		h,
		i,
		N,
		y
	]), X(() => {
		var e;
		if (s && n && (e = h.current) != null && (e = e.__options) != null && e.blockPointerEvents && k()) {
			T.current = !0;
			let e = o.floating;
			if (v(o.domReference) && e) {
				var t;
				let n = J(o.floating).body;
				n.setAttribute(oi, "");
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
		k
	]), X(() => {
		n || (b.current = void 0, D.current = !1, j(), M());
	}, [
		n,
		j,
		M
	]), O.useEffect(() => () => {
		j(), $(x), $(C), M();
	}, [
		s,
		o.domReference,
		j,
		M
	]);
	let P = O.useMemo(() => {
		function e(e) {
			b.current = e.pointerType;
		}
		return {
			onPointerDown: e,
			onPointerEnter: e,
			onMouseMove(e) {
				let { nativeEvent: t } = e;
				function i() {
					!w.current && !_.current && r(!0, t, "hover");
				}
				u && !dr(b.current) || n || ci(y.current) === 0 || D.current && e.movementX ** 2 + e.movementY ** 2 < 2 || ($(C), b.current === "touch" ? i() : (D.current = !0, C.current = window.setTimeout(i, ci(y.current))));
			}
		};
	}, [
		u,
		r,
		n,
		_,
		y
	]);
	return O.useMemo(() => s ? { reference: P } : {}, [s, P]);
}
var ui = 0;
function di(e, t) {
	t === void 0 && (t = {});
	let { preventScroll: n = !1, cancelPrevious: r = !0, sync: i = !1 } = t;
	r && cancelAnimationFrame(ui);
	let a = () => e?.focus({ preventScroll: n });
	i ? a() : ui = requestAnimationFrame(a);
}
function fi(e, t) {
	if (!e || !t) return !1;
	let n = t.getRootNode == null ? void 0 : t.getRootNode();
	if (e.contains(t)) return !0;
	if (n && D(n)) {
		let n = t;
		for (; n;) {
			if (e === n) return !0;
			n = n.parentNode || n.host;
		}
	}
	return !1;
}
function pi(e) {
	return "composedPath" in e ? e.composedPath()[0] : e.target;
}
function mi(e) {
	return e?.ownerDocument || document;
}
var hi = {
	inert: /*#__PURE__*/ new WeakMap(),
	"aria-hidden": /*#__PURE__*/ new WeakMap(),
	none: /*#__PURE__*/ new WeakMap()
};
function gi(e) {
	return e === "inert" ? hi.inert : e === "aria-hidden" ? hi["aria-hidden"] : hi.none;
}
var _i = /*#__PURE__*/ new WeakSet(), vi = {}, yi = 0, bi = () => typeof HTMLElement < "u" && "inert" in HTMLElement.prototype;
function xi(e) {
	return e ? D(e) ? e.host : xi(e.parentNode) : null;
}
var Si = (e, t) => t.map((t) => {
	if (e.contains(t)) return t;
	let n = xi(t);
	return e.contains(n) ? n : null;
}).filter((e) => e != null);
function Ci(e, t, n, r) {
	let i = "data-floating-ui-inert", a = r ? "inert" : n ? "aria-hidden" : null, o = Si(t, e), s = /* @__PURE__ */ new Set(), c = new Set(o), l = [];
	vi[i] || (vi[i] = /* @__PURE__ */ new WeakMap());
	let u = vi[i];
	o.forEach(d), f(t), s.clear();
	function d(e) {
		e && !s.has(e) && (s.add(e), e.parentNode && d(e.parentNode));
	}
	function f(e) {
		e && !c.has(e) && [].forEach.call(e.children, (e) => {
			if (g(e) !== "script") {
				if (s.has(e)) f(e);
				else {
					let t = a ? e.getAttribute(a) : null, n = t !== null && t !== "false", r = gi(a), o = (r.get(e) || 0) + 1, s = (u.get(e) || 0) + 1;
					r.set(e, o), u.set(e, s), l.push(e), o === 1 && n && _i.add(e), s === 1 && e.setAttribute(i, ""), !n && a && e.setAttribute(a, a === "inert" ? "" : "true");
				}
			}
		});
	}
	return yi++, () => {
		l.forEach((e) => {
			let t = gi(a), n = (t.get(e) || 0) - 1, r = (u.get(e) || 0) - 1;
			t.set(e, n), u.set(e, r), n || (!_i.has(e) && a && e.removeAttribute(a), _i.delete(e)), r || e.removeAttribute(i);
		}), yi--, yi || (hi.inert = /* @__PURE__ */ new WeakMap(), hi["aria-hidden"] = /* @__PURE__ */ new WeakMap(), hi.none = /* @__PURE__ */ new WeakMap(), _i = /* @__PURE__ */ new WeakSet(), vi = {});
	};
}
function wi(e, t, n) {
	t === void 0 && (t = !1), n === void 0 && (n = !1);
	let r = mi(e[0]).body;
	return Ci(e.concat(Array.from(r.querySelectorAll("[aria-live],[role=\"status\"],output"))), r, t, n);
}
var Ti = {
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
}, Ei = /*#__PURE__*/ O.forwardRef(function(e, t) {
	let [n, r] = O.useState();
	X(() => {
		Hn() && r("button");
	}, []);
	let i = {
		ref: t,
		tabIndex: 0,
		role: n,
		"aria-hidden": !n || void 0,
		[ai("focus-guard")]: "",
		style: Ti
	};
	return /*#__PURE__*/ (0, A.jsx)("span", {
		...e,
		...i
	});
}), Di = {
	clipPath: "inset(50%)",
	position: "fixed",
	top: 0,
	left: 0
}, Oi = /*#__PURE__*/ O.createContext(null), ki = /*#__PURE__*/ ai("portal");
function Ai(e) {
	e === void 0 && (e = {});
	let { id: t, root: n } = e, r = Xr(), i = Mi(), [a, o] = O.useState(null), s = O.useRef(null);
	return X(() => () => {
		a?.remove(), queueMicrotask(() => {
			s.current = null;
		});
	}, [a]), X(() => {
		if (!r || s.current) return;
		let e = t ? document.getElementById(t) : null;
		if (!e) return;
		let n = document.createElement("div");
		n.id = r, n.setAttribute(ki, ""), e.appendChild(n), s.current = n, o(n);
	}, [t, r]), X(() => {
		if (n === null || !r || s.current) return;
		let e = n || i?.portalNode;
		e && !E(e) && (e = e.current), e ||= document.body;
		let a = null;
		t && (a = document.createElement("div"), a.id = t, e.appendChild(a));
		let c = document.createElement("div");
		c.id = r, c.setAttribute(ki, ""), e = a || e, e.appendChild(c), s.current = c, o(c);
	}, [
		t,
		n,
		r,
		i
	]), a;
}
function ji(e) {
	let { children: t, id: n, root: r, preserveTabOrder: i = !0 } = e, a = Ai({
		id: n,
		root: r
	}), [o, s] = O.useState(null), c = O.useRef(null), l = O.useRef(null), u = O.useRef(null), d = O.useRef(null), f = o?.modal, p = o?.open, m = !!o && !o.modal && o.open && i && !!(r || a);
	return O.useEffect(() => {
		if (!a || !i || f) return;
		function e(e) {
			a && Or(e) && (e.type === "focusin" ? Ar : kr)(a);
		}
		return a.addEventListener("focusin", e, !0), a.addEventListener("focusout", e, !0), () => {
			a.removeEventListener("focusin", e, !0), a.removeEventListener("focusout", e, !0);
		};
	}, [
		a,
		i,
		f
	]), O.useEffect(() => {
		a && (p || Ar(a));
	}, [p, a]), /*#__PURE__*/ (0, A.jsxs)(Oi.Provider, {
		value: O.useMemo(() => ({
			preserveTabOrder: i,
			beforeOutsideRef: c,
			afterOutsideRef: l,
			beforeInsideRef: u,
			afterInsideRef: d,
			portalNode: a,
			setFocusManagerState: s
		}), [i, a]),
		children: [
			m && a && /*#__PURE__*/ (0, A.jsx)(Ei, {
				"data-type": "outside",
				ref: c,
				onFocus: (e) => {
					if (Or(e, a)) {
						var t;
						(t = u.current) == null || t.focus();
					} else Dr(o ? o.domReference : null)?.focus();
				}
			}),
			m && a && /*#__PURE__*/ (0, A.jsx)("span", {
				"aria-owns": a.id,
				style: Di
			}),
			a && /*#__PURE__*/ It.createPortal(t, a),
			m && a && /*#__PURE__*/ (0, A.jsx)(Ei, {
				"data-type": "outside",
				ref: l,
				onFocus: (e) => {
					if (Or(e, a)) {
						var t;
						(t = d.current) == null || t.focus();
					} else Er(o ? o.domReference : null)?.focus(), o != null && o.closeOnFocusOut && o?.onOpenChange(!1, e.nativeEvent, "focus-out");
				}
			})
		]
	});
}
var Mi = () => O.useContext(Oi);
function Ni(e) {
	return O.useMemo(() => (t) => {
		e.forEach((e) => {
			e && (e.current = t);
		});
	}, e);
}
var Pi = 20, Fi = [];
function Ii() {
	Fi = Fi.filter((e) => e.deref()?.isConnected);
}
function Li(e) {
	Ii(), e && g(e) !== "body" && (Fi.push(new WeakRef(e)), Fi.length > Pi && (Fi = Fi.slice(-20)));
}
function Ri() {
	return Ii(), Fi[Fi.length - 1]?.deref();
}
function zi(e) {
	let t = wr();
	return zn(e, t) ? e : Ln(e, t)[0] || e;
}
function Bi(e, t) {
	var n;
	if (!t.current.includes("floating") && !((n = e.getAttribute("role")) != null && n.includes("dialog"))) return;
	let r = wr(), i = Rn(e, r).filter((e) => {
		let t = e.getAttribute("data-tabindex") || "";
		return zn(e, r) || e.hasAttribute("data-tabindex") && !t.startsWith("-");
	}), a = e.getAttribute("tabindex");
	t.current.includes("floating") || i.length === 0 ? a !== "0" && e.setAttribute("tabindex", "0") : (a !== "-1" || e.hasAttribute("data-tabindex") && e.getAttribute("data-tabindex") !== "-1") && (e.setAttribute("tabindex", "-1"), e.setAttribute("data-tabindex", "-1"));
}
var Vi = /*#__PURE__*/ O.forwardRef(function(e, t) {
	return /*#__PURE__*/ (0, A.jsx)("button", {
		...e,
		type: "button",
		ref: t,
		tabIndex: -1,
		style: Ti
	});
});
function Hi(e) {
	let { context: t, children: n, disabled: r = !1, order: i = ["content"], guards: a = !0, initialFocus: o = 0, returnFocus: s = !0, restoreFocus: c = !1, modal: l = !0, visuallyHiddenDismiss: u = !1, closeOnFocusOut: d = !0, outsideElementsInert: f = !1, getInsideElements: p = () => [] } = e, { open: m, onOpenChange: h, events: g, dataRef: v, elements: { domReference: y, floating: b } } = t, x = Z(() => v.current.floatingContext?.nodeId), S = Z(p), C = typeof o == "number" && o < 0, w = rr(y) && C, T = bi(), E = !T || a, D = !E || T && f, k = pr(i), j = pr(o), M = pr(s), N = ti(), P = Mi(), F = O.useRef(null), I = O.useRef(null), L = O.useRef(!1), R = O.useRef(!1), ee = O.useRef(-1), z = O.useRef(-1), B = P != null, V = ir(b), H = Z(function(e) {
		return e === void 0 && (e = V), e ? Ln(e, wr()) : [];
	}), U = Z((e) => {
		let t = H(e);
		return k.current.map((e) => y && e === "reference" ? y : V && e === "floating" ? V : t).filter(Boolean).flat();
	});
	O.useEffect(() => {
		if (r || !l) return;
		function e(e) {
			if (e.key === "Tab") {
				Qn(V, Zn(J(V))) && H().length === 0 && !w && Y(e);
				let t = U(), n = $n(e);
				k.current[0] === "reference" && n === y && (Y(e), e.shiftKey ? di(t[t.length - 1]) : di(t[1])), k.current[1] === "floating" && n === V && e.shiftKey && (Y(e), di(t[0]));
			}
		}
		let t = J(V);
		return t.addEventListener("keydown", e), () => {
			t.removeEventListener("keydown", e);
		};
	}, [
		r,
		y,
		V,
		l,
		k,
		w,
		H,
		U
	]), O.useEffect(() => {
		if (r || !b) return;
		function e(e) {
			let t = $n(e), n = H().indexOf(t);
			n !== -1 && (ee.current = n);
		}
		return b.addEventListener("focusin", e), () => {
			b.removeEventListener("focusin", e);
		};
	}, [
		r,
		b,
		H
	]), O.useEffect(() => {
		if (r || !d) return;
		function e() {
			R.current = !0, setTimeout(() => {
				R.current = !1;
			});
		}
		function t(e) {
			let t = e.relatedTarget, n = e.currentTarget, r = $n(e);
			queueMicrotask(() => {
				let i = x(), a = !(Qn(y, t) || Qn(b, t) || Qn(t, b) || Qn(P?.portalNode, t) || t != null && t.hasAttribute(ai("focus-guard")) || N && (ar(N.nodesRef.current, i).find((e) => Qn(e.context?.elements.floating, t) || Qn(e.context?.elements.domReference, t)) || sr(N.nodesRef.current, i).find((e) => [e.context?.elements.floating, ir(e.context?.elements.floating)].includes(t) || e.context?.elements.domReference === t)));
				if (n === y && V && Bi(V, k), c && n !== y && !(r != null && r.isConnected) && Zn(J(V)) === J(V).body) {
					_(V) && V.focus();
					let e = ee.current, t = H(), n = t[e] || t[t.length - 1] || V;
					_(n) && n.focus();
				}
				v.current.insideReactTree ? v.current.insideReactTree = !1 : (w || !l) && t && a && !R.current && t !== Ri() && (L.current = !0, h(!1, e, "focus-out"));
			});
		}
		let n = !(N || !P);
		function i() {
			$(z), v.current.insideReactTree = !0, z.current = window.setTimeout(() => {
				v.current.insideReactTree = !1;
			});
		}
		if (b && _(y)) return y.addEventListener("focusout", t), y.addEventListener("pointerdown", e), b.addEventListener("focusout", t), n && b.addEventListener("focusout", i, !0), () => {
			y.removeEventListener("focusout", t), y.removeEventListener("pointerdown", e), b.removeEventListener("focusout", t), n && b.removeEventListener("focusout", i, !0);
		};
	}, [
		r,
		y,
		b,
		V,
		l,
		N,
		P,
		h,
		d,
		c,
		H,
		w,
		x,
		k,
		v
	]);
	let W = O.useRef(null), te = O.useRef(null), G = Ni([W, P?.beforeInsideRef]), ne = Ni([te, P?.afterInsideRef]);
	O.useEffect(() => {
		var e, t;
		if (r || !b) return;
		let n = Array.from((P == null || (e = P.portalNode) == null ? void 0 : e.querySelectorAll("[" + ai("portal") + "]")) || []), i = (t = (N ? sr(N.nodesRef.current, x()) : []).find((e) => rr(e.context?.elements.domReference || null))) == null || (t = t.context) == null ? void 0 : t.elements.domReference, a = [
			b,
			i,
			...n,
			...S(),
			F.current,
			I.current,
			W.current,
			te.current,
			P?.beforeOutsideRef.current,
			P?.afterOutsideRef.current,
			k.current.includes("reference") || w ? y : null
		].filter((e) => e != null), o = l || w ? wi(a, !D, D) : wi(a);
		return () => {
			o();
		};
	}, [
		r,
		y,
		b,
		l,
		k,
		P,
		w,
		E,
		D,
		N,
		x,
		S
	]), X(() => {
		if (r || !_(V)) return;
		let e = Zn(J(V));
		queueMicrotask(() => {
			let t = U(V), n = j.current, r = (typeof n == "number" ? t[n] : n.current) || V, i = Qn(V, e);
			!C && !i && m && di(r, { preventScroll: r === V });
		});
	}, [
		r,
		m,
		V,
		C,
		U,
		j
	]), X(() => {
		if (r || !V) return;
		let e = J(V);
		Li(Zn(e));
		function t(e) {
			let { reason: t, event: n, nested: r } = e;
			if (["hover", "safe-polygon"].includes(t) && n.type === "mouseleave" && (L.current = !0), t === "outside-press") {
				if (r) L.current = !1;
				else if (lr(n) || ur(n)) L.current = !1;
				else {
					let e = !1;
					document.createElement("div").focus({ get preventScroll() {
						return e = !0, !1;
					} }), e ? L.current = !1 : L.current = !0;
				}
			}
		}
		g.on("openchange", t);
		let n = e.createElement("span");
		n.setAttribute("tabindex", "-1"), n.setAttribute("aria-hidden", "true"), Object.assign(n.style, Ti), B && y && y.insertAdjacentElement("afterend", n);
		function i() {
			if (typeof M.current == "boolean") {
				let e = y || Ri();
				return e && e.isConnected ? e : n;
			}
			return M.current.current || n;
		}
		return () => {
			g.off("openchange", t);
			let r = Zn(e), a = Qn(b, r) || N && ar(N.nodesRef.current, x(), !1).some((e) => Qn(e.context?.elements.floating, r)), o = i();
			queueMicrotask(() => {
				let t = zi(o);
				M.current && !L.current && _(t) && (t === r || r === e.body || a) && t.focus({ preventScroll: !0 }), n.remove();
			});
		};
	}, [
		r,
		b,
		V,
		M,
		v,
		g,
		N,
		B,
		y,
		x
	]), O.useEffect(() => (queueMicrotask(() => {
		L.current = !1;
	}), () => {
		queueMicrotask(Ii);
	}), [r]), X(() => {
		if (!r && P) return P.setFocusManagerState({
			modal: l,
			closeOnFocusOut: d,
			open: m,
			onOpenChange: h,
			domReference: y
		}), () => {
			P.setFocusManagerState(null);
		};
	}, [
		r,
		P,
		l,
		m,
		h,
		d,
		y
	]), X(() => {
		r || V && Bi(V, k);
	}, [
		r,
		V,
		k
	]);
	function re(e) {
		return r || !u || !l ? null : /*#__PURE__*/ (0, A.jsx)(Vi, {
			ref: e === "start" ? F : I,
			onClick: (e) => h(!1, e.nativeEvent),
			children: typeof u == "string" ? u : "Dismiss"
		});
	}
	let ie = !r && E && (!l || !w) && (B || l);
	return /*#__PURE__*/ (0, A.jsxs)(A.Fragment, { children: [
		ie && /*#__PURE__*/ (0, A.jsx)(Ei, {
			"data-type": "inside",
			ref: G,
			onFocus: (e) => {
				if (l) {
					let e = U();
					di(i[0] === "reference" ? e[0] : e[e.length - 1]);
				} else if (P != null && P.preserveTabOrder && P.portalNode) {
					if (L.current = !1, Or(e, P.portalNode)) Er(y)?.focus();
					else {
						var t;
						(t = P.beforeOutsideRef.current) == null || t.focus();
					}
				}
			}
		}),
		!w && re("start"),
		n,
		re("end"),
		ie && /*#__PURE__*/ (0, A.jsx)(Ei, {
			"data-type": "inside",
			ref: ne,
			onFocus: (e) => {
				if (l) di(U()[0]);
				else if (P != null && P.preserveTabOrder && P.portalNode) {
					if (d && (L.current = !0), Or(e, P.portalNode)) Dr(y)?.focus();
					else {
						var t;
						(t = P.afterOutsideRef.current) == null || t.focus();
					}
				}
			}
		})
	] });
}
function Ui(e) {
	return _(e.target) && e.target.tagName === "BUTTON";
}
function Wi(e) {
	return _(e.target) && e.target.tagName === "A";
}
function Gi(e) {
	return nr(e);
}
function Ki(e, t) {
	t === void 0 && (t = {});
	let { open: n, onOpenChange: r, dataRef: i, elements: { domReference: a } } = e, { enabled: o = !0, event: s = "click", toggle: c = !0, ignoreMouse: l = !1, keyboardHandlers: u = !0, stickIfOpen: d = !0 } = t, f = O.useRef(), p = O.useRef(!1), m = O.useMemo(() => ({
		onPointerDown(e) {
			f.current = e.pointerType;
		},
		onMouseDown(e) {
			let t = f.current;
			e.button === 0 && s !== "click" && (dr(t, !0) && l || (n && c && (!(i.current.openEvent && d) || i.current.openEvent.type === "mousedown") ? r(!1, e.nativeEvent, "click") : (e.preventDefault(), r(!0, e.nativeEvent, "click"))));
		},
		onClick(e) {
			let t = f.current;
			s === "mousedown" && f.current ? f.current = void 0 : dr(t, !0) && l || (n && c && (!(i.current.openEvent && d) || i.current.openEvent.type === "click") ? r(!1, e.nativeEvent, "click") : r(!0, e.nativeEvent, "click"));
		},
		onKeyDown(e) {
			f.current = void 0, !(e.defaultPrevented || !u || Ui(e)) && (e.key === " " && !Gi(a) && (e.preventDefault(), p.current = !0), !Wi(e) && e.key === "Enter" && r(!(n && c), e.nativeEvent, "click"));
		},
		onKeyUp(e) {
			e.defaultPrevented || !u || Ui(e) || Gi(a) || e.key === " " && p.current && (p.current = !1, r(!(n && c), e.nativeEvent, "click"));
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
	return O.useMemo(() => o ? { reference: m } : {}, [o, m]);
}
var qi = {
	pointerdown: "onPointerDown",
	mousedown: "onMouseDown",
	click: "onClick"
}, Ji = {
	pointerdown: "onPointerDownCapture",
	mousedown: "onMouseDownCapture",
	click: "onClickCapture"
}, Yi = (e) => ({
	escapeKey: typeof e == "boolean" ? e : e?.escapeKey ?? !1,
	outsidePress: typeof e == "boolean" ? e : e?.outsidePress ?? !0
});
function Xi(e, t) {
	t === void 0 && (t = {});
	let { open: n, onOpenChange: r, elements: i, dataRef: a } = e, { enabled: o = !0, escapeKey: s = !0, outsidePress: c = !0, outsidePressEvent: l = "pointerdown", referencePress: u = !1, referencePressEvent: f = "pointerdown", ancestorScroll: m = !1, bubbles: g, capture: y } = t, x = ti(), S = Z(typeof c == "function" ? c : () => !1), w = typeof c == "function" ? S : c, T = O.useRef(!1), { escapeKey: E, outsidePress: D } = Yi(g), { escapeKey: k, outsidePress: A } = Yi(y), j = O.useRef(!1), M = Z((e) => {
		if (!n || !o || !s || e.key !== "Escape" || j.current) return;
		let t = a.current.floatingContext?.nodeId, i = x ? ar(x.nodesRef.current, t) : [];
		if (!E && (e.stopPropagation(), i.length > 0)) {
			let e = !0;
			if (i.forEach((t) => {
				var n;
				(n = t.context) != null && n.open && !t.context.dataRef.current.__escapeKeyBubbles && (e = !1);
			}), !e) return;
		}
		r(!1, cr(e) ? e.nativeEvent : e, "escape-key");
	}), N = Z((e) => {
		var t;
		let n = () => {
			var t;
			M(e), (t = $n(e)) == null || t.removeEventListener("keydown", n);
		};
		(t = $n(e)) == null || t.addEventListener("keydown", n);
	}), P = Z((e) => {
		let t = a.current.insideReactTree;
		a.current.insideReactTree = !1;
		let n = T.current;
		if (T.current = !1, l === "click" && n || t || typeof w == "function" && !w(e)) return;
		let o = $n(e), s = "[" + ai("inert") + "]", c = J(i.floating).querySelectorAll(s), u = v(o) ? o : null;
		for (; u && !d(u);) {
			let e = b(u);
			if (d(e) || !v(e)) break;
			u = e;
		}
		if (c.length && v(o) && !tr(o) && !Qn(o, i.floating) && Array.from(c).every((e) => !Qn(u, e))) return;
		if (_(o) && L) {
			let t = d(o), n = h(o), r = /auto|scroll/, i = t || r.test(n.overflowX), a = t || r.test(n.overflowY), s = i && o.clientWidth > 0 && o.scrollWidth > o.clientWidth, c = a && o.clientHeight > 0 && o.scrollHeight > o.clientHeight, l = n.direction === "rtl", u = c && (l ? e.offsetX <= o.offsetWidth - o.clientWidth : e.offsetX > o.clientWidth), f = s && e.offsetY > o.clientHeight;
			if (u || f) return;
		}
		let f = a.current.floatingContext?.nodeId, p = x && ar(x.nodesRef.current, f).some((t) => er(e, t.context?.elements.floating));
		if (er(e, i.floating) || er(e, i.domReference) || p) return;
		let m = x ? ar(x.nodesRef.current, f) : [];
		if (m.length > 0) {
			let e = !0;
			if (m.forEach((t) => {
				var n;
				(n = t.context) != null && n.open && !t.context.dataRef.current.__outsidePressBubbles && (e = !1);
			}), !e) return;
		}
		r(!1, e, "outside-press");
	}), F = Z((e) => {
		var t;
		let n = () => {
			var t;
			P(e), (t = $n(e)) == null || t.removeEventListener(l, n);
		};
		(t = $n(e)) == null || t.addEventListener(l, n);
	});
	O.useEffect(() => {
		if (!n || !o) return;
		a.current.__escapeKeyBubbles = E, a.current.__outsidePressBubbles = D;
		let e = -1;
		function t(e) {
			r(!1, e, "ancestor-scroll");
		}
		function c() {
			window.clearTimeout(e), j.current = !0;
		}
		function u() {
			e = window.setTimeout(() => {
				j.current = !1;
			}, p() ? 5 : 0);
		}
		let d = J(i.floating);
		s && (d.addEventListener("keydown", k ? N : M, k), d.addEventListener("compositionstart", c), d.addEventListener("compositionend", u)), w && d.addEventListener(l, A ? F : P, A);
		let f = [];
		return m && (v(i.domReference) && (f = C(i.domReference)), v(i.floating) && (f = f.concat(C(i.floating))), !v(i.reference) && i.reference && i.reference.contextElement && (f = f.concat(C(i.reference.contextElement)))), f = f.filter((e) => e !== d.defaultView?.visualViewport), f.forEach((e) => {
			e.addEventListener("scroll", t);
		}), () => {
			s && (d.removeEventListener("keydown", k ? N : M, k), d.removeEventListener("compositionstart", c), d.removeEventListener("compositionend", u)), w && d.removeEventListener(l, A ? F : P, A), f.forEach((e) => {
				e.removeEventListener("scroll", t);
			}), window.clearTimeout(e);
		};
	}, [
		a,
		i,
		s,
		w,
		l,
		n,
		r,
		m,
		o,
		E,
		D,
		M,
		k,
		N,
		P,
		A,
		F
	]), O.useEffect(() => {
		a.current.insideReactTree = !1;
	}, [
		a,
		w,
		l
	]);
	let I = O.useMemo(() => ({
		onKeyDown: M,
		...u && {
			[qi[f]]: (e) => {
				r(!1, e.nativeEvent, "reference-press");
			},
			...f !== "click" && { onClick(e) {
				r(!1, e.nativeEvent, "reference-press");
			} }
		}
	}), [
		M,
		r,
		u,
		f
	]), L = O.useMemo(() => {
		function e(e) {
			e.button === 0 && (T.current = !0);
		}
		return {
			onKeyDown: M,
			onMouseDown: e,
			onMouseUp: e,
			[Ji[l]]: () => {
				a.current.insideReactTree = !0;
			}
		};
	}, [
		M,
		l,
		a
	]);
	return O.useMemo(() => o ? {
		reference: I,
		floating: L
	} : {}, [
		o,
		I,
		L
	]);
}
function Zi(e) {
	let { open: t = !1, onOpenChange: n, elements: r } = e, i = Xr(), a = O.useRef({}), [o] = O.useState(() => Zr()), s = ei() != null, [c, l] = O.useState(r.reference), u = Z((e, t, r) => {
		a.current.openEvent = e ? t : void 0, o.emit("openchange", {
			open: e,
			event: t,
			reason: r,
			nested: s
		}), n?.(e, t, r);
	}), d = O.useMemo(() => ({ setPositionReference: l }), []), f = O.useMemo(() => ({
		reference: c || r.reference || null,
		floating: r.floating || null,
		domReference: r.reference
	}), [
		c,
		r.reference,
		r.floating
	]);
	return O.useMemo(() => ({
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
function Qi(e) {
	let { elements: t, ...n } = e === void 0 ? {} : e, { nodeId: r } = n, i = Zi({
		...n,
		elements: {
			reference: t?.reference ?? null,
			floating: t?.floating ?? null
		}
	}), a = n.rootContext || i, o = a.elements, [s, c] = O.useState(null), [l, u] = O.useState(null), d = o?.domReference || s, f = O.useRef(null), p = ti();
	X(() => {
		d && (f.current = d);
	}, [d]);
	let h = m({
		...n,
		elements: {
			...o,
			...l && { reference: l }
		}
	}), g = O.useCallback((e) => {
		let t = v(e) ? {
			getBoundingClientRect: () => e.getBoundingClientRect(),
			getClientRects: () => e.getClientRects(),
			contextElement: e
		} : e;
		u(t), h.refs.setReference(t);
	}, [h.refs]), _ = O.useCallback((e) => {
		(v(e) || e === null) && (f.current = e, c(e)), (v(h.refs.reference.current) || h.refs.reference.current === null || e !== null && !v(e)) && h.refs.setReference(e);
	}, [h.refs]), y = O.useMemo(() => ({
		...h.refs,
		setReference: _,
		setPositionReference: g,
		domReference: f
	}), [
		h.refs,
		_,
		g
	]), b = O.useMemo(() => ({
		...h.elements,
		domReference: d
	}), [h.elements, d]), x = O.useMemo(() => ({
		...h,
		...a,
		refs: y,
		elements: b,
		nodeId: r
	}), [
		h,
		y,
		b,
		r,
		a
	]);
	return X(() => {
		a.dataRef.current.floatingContext = x;
		let e = p?.nodesRef.current.find((e) => e.id === r);
		e && (e.context = x);
	}), O.useMemo(() => ({
		...h,
		context: x,
		refs: y,
		elements: b
	}), [
		h,
		y,
		b,
		x
	]);
}
function $i(e, t, n) {
	let r = /* @__PURE__ */ new Map(), i = n === "item", a = e;
	if (i && e) {
		let { [Lr]: t, [Rr]: n, ...r } = e;
		a = r;
	}
	return {
		...n === "floating" && {
			tabIndex: -1,
			[Ir]: ""
		},
		...a,
		...t.map((t) => {
			let r = t ? t[n] : null;
			return typeof r == "function" ? e ? r(e) : null : r;
		}).concat(e).reduce((e, t) => (t && Object.entries(t).forEach((t) => {
			let [n, a] = t;
			if (!(i && [Lr, Rr].includes(n))) {
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
function ea(e) {
	e === void 0 && (e = []);
	let t = e.map((e) => e?.reference), n = e.map((e) => e?.floating), r = e.map((e) => e?.item), i = O.useCallback((t) => $i(t, e, "reference"), t), a = O.useCallback((t) => $i(t, e, "floating"), n), o = O.useCallback((t) => $i(t, e, "item"), r);
	return O.useMemo(() => ({
		getReferenceProps: i,
		getFloatingProps: a,
		getItemProps: o
	}), [
		i,
		a,
		o
	]);
}
var ta = "Escape";
function na(e, t, n) {
	switch (e) {
		case "vertical": return t;
		case "horizontal": return n;
		default: return t || n;
	}
}
function ra(e, t) {
	return na(t, e === Vr || e === Hr, e === zr || e === Br);
}
function ia(e, t, n) {
	return na(t, e === Hr, n ? e === zr : e === Br) || e === "Enter" || e === " " || e === "";
}
function aa(e, t, n) {
	return na(t, n ? e === zr : e === Br, e === Hr);
}
function oa(e, t, n, r) {
	return t === "both" || t === "horizontal" && r && r > 1 ? e === ta : na(t, n ? e === Br : e === zr, e === Vr);
}
function sa(e, t) {
	let { open: n, onOpenChange: r, elements: i, floatingId: a } = e, { listRef: o, activeIndex: s, onNavigate: c = () => {}, enabled: l = !0, selectedIndex: u = null, allowEscape: d = !1, loop: f = !1, nested: p = !1, rtl: m = !1, virtual: h = !1, focusItemOnOpen: g = "auto", focusItemOnHover: v = !0, openOnArrowKeyDown: y = !0, disabledIndices: b = void 0, orientation: x = "vertical", parentOrientation: S, cols: C = 1, scrollItemIntoView: w = !0, virtualItemRef: T, itemSizes: E, dense: D = !1 } = t, k = pr(ir(i.floating)), A = ei(), j = ti();
	X(() => {
		e.dataRef.current.orientation = x;
	}, [e, x]);
	let M = Z(() => {
		c(F.current === -1 ? null : F.current);
	}), N = rr(i.domReference), P = O.useRef(g), F = O.useRef(u ?? -1), I = O.useRef(null), L = O.useRef(!0), R = O.useRef(M), ee = O.useRef(!!i.floating), z = O.useRef(n), B = O.useRef(!1), V = O.useRef(!1), H = pr(b), U = pr(n), W = pr(w), te = pr(u), [G, ne] = O.useState(), [re, ie] = O.useState(), K = Z(() => {
		function e(e) {
			if (h) {
				var t;
				(t = e.id) != null && t.endsWith("-fui-option") && (e.id = a + "-" + Math.random().toString(16).slice(2, 10)), ne(e.id), j?.events.emit("virtualfocus", e), T && (T.current = e);
			} else di(e, {
				sync: B.current,
				preventScroll: !0
			});
		}
		let t = o.current[F.current], n = V.current;
		t && e(t), (B.current ? (e) => e() : requestAnimationFrame)(() => {
			let r = o.current[F.current] || t;
			if (!r) return;
			t || e(r);
			let i = W.current;
			i && ae && (n || !L.current) && (r.scrollIntoView == null || r.scrollIntoView(typeof i == "boolean" ? {
				block: "nearest",
				inline: "nearest"
			} : i));
		});
	});
	X(() => {
		l && (n && i.floating ? P.current && u != null && (V.current = !0, F.current = u, M()) : ee.current && (F.current = -1, R.current()));
	}, [
		l,
		n,
		i.floating,
		u,
		M
	]), X(() => {
		if (l && n && i.floating) {
			if (s == null) {
				if (B.current = !1, te.current != null) return;
				if (ee.current && (F.current = -1, K()), (!z.current || !ee.current) && P.current && (I.current != null || P.current === !0 && I.current == null)) {
					let e = 0, t = () => {
						o.current[0] == null ? (e < 2 && (e ? requestAnimationFrame : queueMicrotask)(t), e++) : (F.current = I.current == null || ia(I.current, x, m) || p ? _r(o, H.current) : vr(o, H.current), I.current = null, M());
					};
					t();
				}
			} else gr(o, s) || (F.current = s, K(), V.current = !1);
		}
	}, [
		l,
		n,
		i.floating,
		s,
		te,
		p,
		o,
		x,
		m,
		M,
		K,
		H
	]), X(() => {
		var e;
		if (!l || i.floating || !j || h || !ee.current) return;
		let t = j.nodesRef.current, n = (e = t.find((e) => e.id === A)) == null || (e = e.context) == null ? void 0 : e.elements.floating, r = Zn(J(i.floating)), a = t.some((e) => e.context && Qn(e.context.elements.floating, r));
		n && !a && L.current && n.focus({ preventScroll: !0 });
	}, [
		l,
		i.floating,
		j,
		A,
		h
	]), X(() => {
		if (!l || !j || !h || A) return;
		function e(e) {
			ie(e.id), T && (T.current = e);
		}
		return j.events.on("virtualfocus", e), () => {
			j.events.off("virtualfocus", e);
		};
	}, [
		l,
		j,
		h,
		A,
		T
	]), X(() => {
		R.current = M, z.current = n, ee.current = !!i.floating;
	}), X(() => {
		n || (I.current = null, P.current = g);
	}, [n, g]);
	let q = s != null, ae = O.useMemo(() => {
		function e(e) {
			if (!U.current) return;
			let t = o.current.indexOf(e);
			t !== -1 && F.current !== t && (F.current = t, M());
		}
		return {
			onFocus(t) {
				let { currentTarget: n } = t;
				B.current = !0, e(n);
			},
			onClick: (e) => {
				let { currentTarget: t } = e;
				return t.focus({ preventScroll: !0 });
			},
			onMouseMove(t) {
				let { currentTarget: n } = t;
				B.current = !0, V.current = !1, v && e(n);
			},
			onPointerLeave(e) {
				let { pointerType: t } = e;
				if (L.current && t !== "touch" && (B.current = !0, v && (F.current = -1, M(), !h))) {
					var n;
					(n = k.current) == null || n.focus({ preventScroll: !0 });
				}
			}
		};
	}, [
		U,
		k,
		v,
		o,
		M,
		h
	]), oe = O.useCallback(() => {
		var e;
		return S ?? (j == null || (e = j.nodesRef.current.find((e) => e.id === A)) == null || (e = e.context) == null || (e = e.dataRef) == null ? void 0 : e.current.orientation);
	}, [
		A,
		j,
		S
	]), se = Z((e) => {
		if (L.current = !1, B.current = !0, e.which === 229 || !U.current && e.currentTarget === k.current) return;
		if (p && oa(e.key, x, m, C)) {
			ra(e.key, oe()) || Y(e), r(!1, e.nativeEvent, "list-navigation"), _(i.domReference) && (h ? j?.events.emit("virtualfocus", i.domReference) : i.domReference.focus());
			return;
		}
		let t = F.current, a = _r(o, b), s = vr(o, b);
		if (N || (e.key === "Home" && (Y(e), F.current = a, M()), e.key === "End" && (Y(e), F.current = s, M())), C > 1) {
			let t = E || Array.from({ length: o.current.length }, () => ({
				width: 1,
				height: 1
			})), n = br(t, C, D), r = n.findIndex((e) => e != null && !Cr(o, e, b)), i = n.reduce((e, t, n) => t != null && !Cr(o, t, b) ? n : e, -1), c = n[yr({ current: n.map((e) => e == null ? null : o.current[e]) }, {
				event: e,
				orientation: x,
				loop: f,
				rtl: m,
				cols: C,
				disabledIndices: Sr([...(typeof b == "function" ? null : b) || o.current.map((e, t) => Cr(o, t, b) ? t : void 0), void 0], n),
				minIndex: r,
				maxIndex: i,
				prevIndex: xr(F.current > s ? a : F.current, t, n, C, e.key === Hr ? "bl" : e.key === (m ? zr : Br) ? "tr" : "tl"),
				stopEvent: !0
			})];
			if (c != null && (F.current = c, M()), x === "both") return;
		}
		if (ra(e.key, x)) {
			if (Y(e), n && !h && Zn(e.currentTarget.ownerDocument) === e.currentTarget) {
				F.current = ia(e.key, x, m) ? a : s, M();
				return;
			}
			ia(e.key, x, m) ? f ? F.current = t >= s ? d && t !== o.current.length ? -1 : a : Q(o, {
				startingIndex: t,
				disabledIndices: b
			}) : F.current = Math.min(s, Q(o, {
				startingIndex: t,
				disabledIndices: b
			})) : f ? F.current = t <= a ? d && t !== -1 ? o.current.length : s : Q(o, {
				startingIndex: t,
				decrement: !0,
				disabledIndices: b
			}) : F.current = Math.max(a, Q(o, {
				startingIndex: t,
				decrement: !0,
				disabledIndices: b
			})), gr(o, F.current) && (F.current = -1), M();
		}
	}), ce = O.useMemo(() => h && n && q && { "aria-activedescendant": re || G }, [
		h,
		n,
		q,
		re,
		G
	]), le = O.useMemo(() => ({
		"aria-orientation": x === "both" ? void 0 : x,
		...N ? {} : ce,
		onKeyDown: se,
		onPointerMove() {
			L.current = !0;
		}
	}), [
		ce,
		se,
		x,
		N
	]), ue = O.useMemo(() => {
		function e(e) {
			g === "auto" && lr(e.nativeEvent) && (P.current = !0);
		}
		function t(e) {
			P.current = g, g === "auto" && ur(e.nativeEvent) && (P.current = !0);
		}
		return {
			...ce,
			onKeyDown(e) {
				L.current = !1;
				let t = e.key.startsWith("Arrow"), i = ["Home", "End"].includes(e.key), a = t || i, s = aa(e.key, x, m), c = oa(e.key, x, m, C), l = aa(e.key, oe(), m), d = ra(e.key, x), f = (p ? l : d) || e.key === "Enter" || e.key.trim() === "";
				if (h && n) {
					let t = j?.nodesRef.current.find((e) => e.parentId == null), n = j && t ? or(j.nodesRef.current, t.id) : null;
					if (a && n && T) {
						let t = new KeyboardEvent("keydown", {
							key: e.key,
							bubbles: !0
						});
						if (s || c) {
							let r = n.context?.elements.domReference === e.currentTarget, i = c && !r ? n.context?.elements.domReference : s ? o.current.find((e) => e?.id === G) : null;
							i && (Y(e), i.dispatchEvent(t), ie(void 0));
						}
						if ((d || i) && n.context && n.context.open && n.parentId && e.currentTarget !== n.context.elements.domReference) {
							var g;
							Y(e), (g = n.context.elements.domReference) == null || g.dispatchEvent(t);
							return;
						}
					}
					return se(e);
				}
				if (n || y || !t) {
					if (f) {
						let t = ra(e.key, oe());
						I.current = p && t ? null : e.key;
					}
					p ? l && (Y(e), n ? (F.current = _r(o, H.current), M()) : r(!0, e.nativeEvent, "list-navigation")) : d && (u != null && (F.current = u), Y(e), !n && y ? r(!0, e.nativeEvent, "list-navigation") : se(e), n && M());
				}
			},
			onFocus() {
				n && !h && (F.current = -1, M());
			},
			onPointerDown: t,
			onPointerEnter: t,
			onMouseDown: e,
			onClick: e
		};
	}, [
		G,
		ce,
		C,
		se,
		H,
		g,
		o,
		p,
		M,
		r,
		n,
		y,
		x,
		oe,
		m,
		u,
		j,
		h,
		T
	]);
	return O.useMemo(() => l ? {
		reference: ue,
		floating: le,
		item: ae
	} : {}, [
		l,
		ue,
		le,
		ae
	]);
}
var ca = /*#__PURE__*/ new Map([
	["select", "listbox"],
	["combobox", "listbox"],
	["label", !1]
]);
function la(e, t) {
	t === void 0 && (t = {});
	let { open: n, elements: r, floatingId: i } = e, { enabled: a = !0, role: o = "dialog" } = t, s = Xr(), c = r.domReference?.id || s, l = O.useMemo(() => ir(r.floating)?.id || i, [r.floating, i]), u = ca.get(o) ?? o, d = ei() != null, f = O.useMemo(() => u === "tooltip" || o === "label" ? { ["aria-" + (o === "label" ? "labelledby" : "describedby")]: n ? l : void 0 } : {
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
	]), p = O.useMemo(() => {
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
	]), m = O.useCallback((e) => {
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
	return O.useMemo(() => a ? {
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
function ua(e, t) {
	let { open: n, dataRef: r } = e, { listRef: i, activeIndex: a, onMatch: o, onTypingChange: s, enabled: c = !0, findMatch: l = null, resetMs: u = 750, ignoreKeys: d = [], selectedIndex: f = null } = t, p = O.useRef(-1), m = O.useRef(""), h = O.useRef(f ?? a ?? -1), g = O.useRef(null), _ = Z(o), v = Z(s), y = pr(l), b = pr(d);
	X(() => {
		n && ($(p), g.current = null, m.current = "");
	}, [n]), X(() => {
		n && m.current === "" && (h.current = f ?? a ?? -1);
	}, [
		n,
		f,
		a
	]);
	let x = Z((e) => {
		e ? r.current.typing || (r.current.typing = e, v(e)) : r.current.typing && (r.current.typing = e, v(e));
	}), S = Z((e) => {
		function t(e, t, n) {
			let r = y.current ? y.current(t, n) : t.find((e) => e?.toLocaleLowerCase().indexOf(n.toLocaleLowerCase()) === 0);
			return r ? e.indexOf(r) : -1;
		}
		let r = i.current;
		if (m.current.length > 0 && m.current[0] !== " " && (t(r, r, m.current) === -1 ? x(!1) : e.key === " " && Y(e)), r == null || b.current.includes(e.key) || e.key.length !== 1 || e.ctrlKey || e.metaKey || e.altKey) return;
		n && e.key !== " " && (Y(e), x(!0)), r.every((e) => !e || e[0]?.toLocaleLowerCase() !== e[1]?.toLocaleLowerCase()) && m.current === e.key && (m.current = "", h.current = g.current), m.current += e.key, $(p), p.current = window.setTimeout(() => {
			m.current = "", h.current = g.current, x(!1);
		}, u);
		let a = h.current, o = t(r, [...r.slice((a || 0) + 1), ...r.slice(0, (a || 0) + 1)], m.current);
		o === -1 ? e.key !== " " && (m.current = "", x(!1)) : (_(o), g.current = o);
	}), C = O.useMemo(() => ({ onKeyDown: S }), [S]), w = O.useMemo(() => ({
		onKeyDown: S,
		onKeyUp(e) {
			e.key === " " && x(!1);
		}
	}), [S, x]);
	return O.useMemo(() => c ? {
		reference: C,
		floating: w
	} : {}, [
		c,
		C,
		w
	]);
}
function da(e, t, n) {
	return n === void 0 && (n = !0), e.filter((e) => e.parentId === t && (!n || e.context?.open)).flatMap((t) => [t, ...da(e, t.id, n)]);
}
function fa(e, t) {
	let [n, r] = e, i = !1, a = t.length;
	for (let e = 0, o = a - 1; e < a; o = e++) {
		let [a, s] = t[e] || [0, 0], [c, l] = t[o] || [0, 0];
		s >= r != l >= r && n <= (c - a) * (r - s) / (l - s) + a && (i = !i);
	}
	return i;
}
function pa(e, t) {
	return e[0] >= t.x && e[0] <= t.x + t.width && e[1] >= t.y && e[1] <= t.y + t.height;
}
function ma(e) {
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
				$(i), u();
			}
			if ($(i), !c.domReference || !c.floating || s == null || n == null || o == null) return;
			let { clientX: m, clientY: h } = e, g = [m, h], _ = pi(e), y = e.type === "mouseleave", b = fi(c.floating, _), x = fi(c.domReference, _), S = c.domReference.getBoundingClientRect(), C = c.floating.getBoundingClientRect(), w = s.split("-")[0], T = n > C.right - C.width / 2, E = o > C.bottom - C.height / 2, D = pa(g, S), O = C.width > S.width, k = C.height > S.height, A = (O ? S : C).left, j = (O ? S : C).right, M = (k ? S : C).top, N = (k ? S : C).bottom;
			if (b && (a = !0, !y)) return;
			if (x && (a = !1), x && !y) {
				a = !0;
				return;
			}
			if (y && v(e.relatedTarget) && fi(c.floating, e.relatedTarget) || f && da(f.nodesRef.current, d).length) return;
			if (w === "top" && o >= S.bottom - 1 || w === "bottom" && o <= S.top + 1 || w === "left" && n >= S.right - 1 || w === "right" && n <= S.left + 1) return p();
			let P = [];
			switch (w) {
				case "top":
					P = [
						[A, S.top + 1],
						[A, C.bottom - 1],
						[j, C.bottom - 1],
						[j, S.top + 1]
					];
					break;
				case "bottom":
					P = [
						[A, C.top + 1],
						[A, S.bottom - 1],
						[j, S.bottom - 1],
						[j, C.top + 1]
					];
					break;
				case "left":
					P = [
						[C.right - 1, N],
						[C.right - 1, M],
						[S.left + 1, M],
						[S.left + 1, N]
					];
					break;
				case "right": P = [
					[S.right - 1, N],
					[S.right - 1, M],
					[C.left + 1, M],
					[C.left + 1, N]
				];
			}
			function F(e) {
				let [n, r] = e;
				switch (w) {
					case "top": return [
						[O ? n + t / 2 : T ? n + t * 4 : n - t * 4, r + t + 1],
						[O ? n - t / 2 : T ? n + t * 4 : n - t * 4, r + t + 1],
						[C.left, T || O ? C.bottom - t : C.top],
						[C.right, T ? O ? C.bottom - t : C.top : C.bottom - t]
					];
					case "bottom": return [
						[O ? n + t / 2 : T ? n + t * 4 : n - t * 4, r - t],
						[O ? n - t / 2 : T ? n + t * 4 : n - t * 4, r - t],
						[C.left, T || O ? C.top + t : C.bottom],
						[C.right, T ? O ? C.top + t : C.bottom : C.top + t]
					];
					case "left": {
						let e = [n + t + 1, k ? r + t / 2 : E ? r + t * 4 : r - t * 4], i = [n + t + 1, k ? r - t / 2 : E ? r + t * 4 : r - t * 4];
						return [
							[E || k ? C.right - t : C.left, C.top],
							[E ? k ? C.right - t : C.left : C.right - t, C.bottom],
							e,
							i
						];
					}
					case "right": return [
						[n - t, k ? r + t / 2 : E ? r + t * 4 : r - t * 4],
						[n - t, k ? r - t / 2 : E ? r + t * 4 : r - t * 4],
						[E || k ? C.left + t : C.right, C.top],
						[E ? k ? C.left + t : C.right : C.left + t, C.bottom]
					];
				}
			}
			if (!fa([m, h], P)) {
				if (a && !D) return p();
				if (!y && r) {
					let t = l(e.clientX, e.clientY);
					if (t !== null && t < .1) return p();
				}
				fa([m, h], F([n, o])) ? !a && r && (i.current = window.setTimeout(p, 40)) : p();
			}
		};
	};
	return u.__options = { blockPointerEvents: n }, u;
}
// ../../../clicky-ui/packages/ui/src/overlay/dropdownMenuContext.ts
var ha = "min-w-[8rem] overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-md", ga = w({
	padding: 8,
	apply({ availableHeight: e, elements: t }) {
		t.floating.style.maxHeight = `${Math.max(0, e)}px`;
	}
}), _a = (0, O.createContext)({
	getItemProps: () => ({}),
	activeIndex: null,
	setActiveIndex: () => {},
	setHasFocusInside: () => {},
	isOpen: !1
}), va = "flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs text-popover-foreground hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50", ya = "select-none px-3 pb-1 pt-2 text-[10px] font-medium uppercase tracking-wide text-muted-foreground";
function ba({ icon: e, iconColor: t, iconClassName: n }) {
	return e ? /* @__PURE__ */ (0, A.jsx)("span", {
		className: i("inline-flex shrink-0 text-base", n),
		style: t ? { color: t } : void 0,
		children: /* @__PURE__ */ (0, A.jsx)(s, { ...typeof e == "string" ? { name: e } : { icon: e } })
	}) : null;
}
function xa({ item: e }) {
	let t = (0, O.useContext)(_a), n = ti(), r = Fr(), i = r.index === t.activeIndex, a = () => {
		e.disabled || (e.onSelect(), n?.events.emit("click"));
	};
	return /* @__PURE__ */ (0, A.jsxs)("button", {
		ref: r.ref,
		type: "button",
		role: "menuitem",
		tabIndex: i ? 0 : -1,
		disabled: e.disabled,
		title: e.title,
		className: va,
		...t.getItemProps({
			onClick: a,
			onFocus: () => t.setHasFocusInside(!0),
			onKeyDown: (e) => {
				(e.key === "Enter" || e.key === " ") && (e.preventDefault(), a());
			}
		}),
		children: [/* @__PURE__ */ (0, A.jsx)(ba, {
			icon: e.icon,
			iconColor: e.iconColor,
			iconClassName: e.iconClassName
		}), e.label]
	});
}
function Sa({ item: e }) {
	let t = (0, O.useContext)(_a), n = ti(), r = ni(), i = ei(), a = Fr(), [o, c] = (0, O.useState)(!1), [l, u] = (0, O.useState)(null), [d, p] = (0, O.useState)(!1), m = (0, O.useRef)([]), h = $t(), { refs: g, floatingStyles: _, context: v } = Qi({
		...r ? { nodeId: r } : {},
		open: o,
		onOpenChange: c,
		placement: "right-start",
		middleware: [
			f({
				mainAxis: 0,
				alignmentAxis: -4
			}),
			x({ padding: 8 }),
			S({
				padding: 8,
				crossAxis: !0
			}),
			ga
		],
		whileElementsMounted: y
	}), { getReferenceProps: b, getFloatingProps: C, getItemProps: w } = ea([
		li(v, {
			delay: { open: 75 },
			handleClose: ma()
		}),
		Ki(v, { toggle: !1 }),
		la(v, { role: "menu" }),
		Xi(v, {
			escapeKey: !1,
			bubbles: !0
		}),
		sa(v, {
			listRef: m,
			activeIndex: l,
			nested: !0,
			onNavigate: u,
			loop: !0
		})
	]);
	Qt(o, () => v.onOpenChange(!1)), (0, O.useEffect)(() => {
		if (!n) return;
		let e = () => v.onOpenChange(!1), t = (e) => {
			e.nodeId !== r && e.parentId === i && v.onOpenChange(!1);
		};
		return n.events.on("click", e), n.events.on("menuopen", t), () => {
			n.events.off("click", e), n.events.off("menuopen", t);
		};
	}, [
		n,
		r,
		i,
		v.onOpenChange
	]), (0, O.useEffect)(() => {
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
	let T = a.index === t.activeIndex, E = jr([g.setReference, a.ref]), D = typeof e.label == "string" ? e.label : void 0;
	return /* @__PURE__ */ (0, A.jsxs)(ri, {
		id: r,
		children: [/* @__PURE__ */ (0, A.jsxs)("button", {
			ref: E,
			type: "button",
			role: "menuitem",
			tabIndex: T ? 0 : -1,
			disabled: e.disabled,
			title: e.title,
			"aria-haspopup": "menu",
			"aria-expanded": o,
			"data-focus-inside": d ? "" : void 0,
			className: va,
			...t.getItemProps(b({ onFocus: () => {
				p(!1), t.setHasFocusInside(!0);
			} })),
			children: [
				/* @__PURE__ */ (0, A.jsx)(ba, {
					icon: e.icon,
					iconColor: e.iconColor,
					iconClassName: e.iconClassName
				}),
				/* @__PURE__ */ (0, A.jsx)("span", {
					className: "min-w-0 flex-1 truncate",
					children: e.label
				}),
				/* @__PURE__ */ (0, A.jsx)(s, {
					icon: ye,
					className: "ml-auto shrink-0 opacity-60"
				})
			]
		}), /* @__PURE__ */ (0, A.jsx)(_a.Provider, {
			value: {
				getItemProps: w,
				activeIndex: l,
				setActiveIndex: u,
				setHasFocusInside: p,
				isOpen: o
			},
			children: /* @__PURE__ */ (0, A.jsx)(Pr, {
				elementsRef: m,
				children: o && /* @__PURE__ */ (0, A.jsx)(ji, { children: /* @__PURE__ */ (0, A.jsx)(Hi, {
					context: v,
					modal: !1,
					initialFocus: -1,
					returnFocus: !1,
					children: /* @__PURE__ */ (0, A.jsxs)("div", {
						ref: g.setFloating,
						role: "menu",
						style: {
							..._,
							zIndex: h
						},
						className: "min-w-[8rem] overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-md",
						...C(),
						"aria-label": D,
						...D ? { "aria-labelledby": void 0 } : {},
						children: [e.header != null && /* @__PURE__ */ (0, A.jsx)("div", {
							className: "border-b border-border px-3 py-1.5",
							children: e.header
						}), /* @__PURE__ */ (0, A.jsx)(Ca, { items: e.children ?? [] })]
					})
				}) })
			})
		})]
	});
}
function Ca({ items: e }) {
	return /* @__PURE__ */ (0, A.jsx)(A.Fragment, { children: e.map((t, n) => {
		let r = t.group != null && t.group !== e[n - 1]?.group, i = !!t.children && t.children.length > 0;
		return /* @__PURE__ */ (0, A.jsxs)(O.Fragment, { children: [r && /* @__PURE__ */ (0, A.jsx)("div", {
			role: "presentation",
			className: ya,
			children: t.group
		}), i ? /* @__PURE__ */ (0, A.jsx)(Sa, { item: t }) : /* @__PURE__ */ (0, A.jsx)(xa, { item: t })] }, n);
	}) });
}
// ../../../clicky-ui/packages/ui/src/overlay/DropdownMenu.tsx
function wa(e) {
	return /* @__PURE__ */ (0, A.jsx)(ii, { children: /* @__PURE__ */ (0, A.jsx)(Ta, { ...e }) });
}
function Ta({ role: e = "menu", label: t, icon: n, hideChevron: r = !1, variant: a = "outline", size: c = "sm", trigger: l, items: d, children: p, header: m, footer: h, align: g = "right", onOpenChange: _, title: v, menuLabel: b, className: C, menuClassName: w }) {
	let [T, E] = (0, O.useState)(!1), [D, k] = (0, O.useState)(null), j = $t(), M = ti(), N = ni(), P = (0, O.useRef)([]), F = (e) => {
		e !== T && (E(e), _?.(e));
	}, { refs: I, floatingStyles: L, context: R } = Qi({
		...N ? { nodeId: N } : {},
		open: T,
		onOpenChange: F,
		placement: g === "right" ? "bottom-end" : "bottom-start",
		whileElementsMounted: y,
		middleware: [
			f(4),
			x({ padding: 8 }),
			S({ padding: 8 }),
			ga
		]
	}), ee = Ki(R), z = Xi(R, { escapeKey: !1 }), B = la(R, { role: e }), V = sa(R, {
		listRef: P,
		activeIndex: D,
		onNavigate: k,
		loop: !0
	}), { getReferenceProps: H, getFloatingProps: U, getItemProps: W } = ea([
		ee,
		z,
		B,
		...e === "menu" ? [V] : []
	]), te = () => {
		F(!1), I.domReference.current instanceof HTMLElement && I.domReference.current.focus();
	};
	Qt(T, te);
	let G = (0, O.useRef)(te);
	return (0, O.useLayoutEffect)(() => {
		G.current = te;
	}), (0, O.useEffect)(() => {
		if (!M) return;
		let e = () => G.current();
		return M.events.on("click", e), () => M.events.off("click", e);
	}, [M]), /* @__PURE__ */ (0, A.jsx)(ri, {
		id: N,
		children: /* @__PURE__ */ (0, A.jsxs)("div", {
			className: i("relative inline-flex", C),
			children: [l ? /* @__PURE__ */ (0, A.jsx)(u, {
				ref: I.setReference,
				className: "inline-flex",
				"aria-haspopup": e,
				"aria-expanded": T,
				...H(),
				children: l
			}) : /* @__PURE__ */ (0, A.jsxs)(o, {
				ref: I.setReference,
				type: "button",
				variant: a,
				size: c,
				title: v,
				"aria-haspopup": e,
				"aria-expanded": T,
				...H(),
				children: [
					n && /* @__PURE__ */ (0, A.jsx)(s, { ...typeof n == "string" ? { name: n } : { icon: n } }),
					t,
					!r && /* @__PURE__ */ (0, A.jsx)(s, { icon: ve })
				]
			}), T && /* @__PURE__ */ (0, A.jsx)(ji, { children: /* @__PURE__ */ (0, A.jsx)(Hi, {
				context: R,
				modal: !1,
				closeOnFocusOut: !1,
				children: /* @__PURE__ */ (0, A.jsxs)("div", {
					ref: I.setFloating,
					...U(),
					role: e,
					"aria-modal": e !== "dialog" && void 0,
					"aria-label": b,
					...b ? { "aria-labelledby": void 0 } : {},
					style: {
						...L,
						zIndex: j
					},
					className: i("min-w-[8rem] overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-md", w),
					children: [
						m != null && /* @__PURE__ */ (0, A.jsx)("div", {
							className: "border-b border-border px-3 py-1.5",
							children: m
						}),
						p ? p(te) : /* @__PURE__ */ (0, A.jsx)(_a.Provider, {
							value: {
								getItemProps: W,
								activeIndex: D,
								setActiveIndex: k,
								setHasFocusInside: () => {},
								isOpen: T
							},
							children: /* @__PURE__ */ (0, A.jsx)(Pr, {
								elementsRef: P,
								children: /* @__PURE__ */ (0, A.jsx)(Ca, { items: d ?? [] })
							})
						}),
						h != null && /* @__PURE__ */ (0, A.jsx)("div", {
							className: "border-t border-border px-3 py-1.5",
							children: h
						})
					]
				})
			}) })]
		})
	});
}
var Ea = [
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
function Da(e) {
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
var Oa = {
	s: 1e3,
	m: 6e4,
	h: 36e5,
	d: 864e5,
	w: 6048e5
}, ka = /^now\s*([+-])\s*(\d+)\s*([smhdwMqy])$/, Aa = /^now\/([wMqy])(?:\s*([+-])\s*(\d+)\s*([wMqy]))?$/;
function ja(e, t = /* @__PURE__ */ new Date()) {
	if (e == null) return null;
	if (typeof e == "string") {
		let n = e.trim();
		if (!n) return null;
		if (n.toLowerCase() === "now") return new Date(t.getTime());
		let r = n.match(Aa);
		if (r) {
			let [, e, n, i, a] = r, o = Na(t, e);
			return n && i && a && (o = Pa(o, a, Number(i) * (n === "-" ? -1 : 1))), o;
		}
		let i = n.match(ka);
		if (i) {
			let [, e, n, r] = i, a = r;
			if (Ma(a)) return Pa(t, a, Number(n) * (e === "-" ? -1 : 1));
			let o = Oa[a.toLowerCase()];
			if (!o) return null;
			let s = Number(n) * o * (e === "-" ? -1 : 1);
			return new Date(t.getTime() + s);
		}
	}
	return Da(e);
}
function Ma(e) {
	return e === "M" || e === "q" || e === "y";
}
function Na(e, t) {
	let n = e.getUTCFullYear(), r = e.getUTCMonth(), i = e.getUTCDate();
	if (t === "w") {
		let e = new Date(Date.UTC(n, r, i)), t = (e.getUTCDay() + 6) % 7;
		return e.setUTCDate(e.getUTCDate() - t), e;
	}
	return t === "M" ? new Date(Date.UTC(n, r, 1)) : t === "q" ? new Date(Date.UTC(n, Math.floor(r / 3) * 3, 1)) : t === "y" ? new Date(Date.UTC(n, 0, 1)) : new Date(e.getTime());
}
function Pa(e, t, n) {
	let r = new Date(e.getTime());
	if (t === "w") r.setUTCDate(r.getUTCDate() + n * 7);
	else if (t === "M") r.setUTCMonth(r.getUTCMonth() + n);
	else if (t === "q") r.setUTCMonth(r.getUTCMonth() + n * 3);
	else if (t === "y") r.setUTCFullYear(r.getUTCFullYear() + n);
	else {
		let e = Oa[t.toLowerCase()];
		e && r.setTime(r.getTime() + n * e);
	}
	return r;
}
var Fa = 6e4, Ia = 864e5;
function La(e) {
	if (e.length === 0) return "iso";
	let t = Infinity, n = -Infinity;
	for (let r of e) {
		let e = r.getTime();
		e < t && (t = e), e > n && (n = e);
	}
	let r = n - t;
	if (r <= Fa) return "relative";
	let i = new Date(t), a = new Date(n);
	return i.getFullYear() === a.getFullYear() && i.getMonth() === a.getMonth() && i.getDate() === a.getDate() ? "time" : i.getFullYear() === a.getFullYear() && r <= Ia * 365 ? "short" : "iso";
}
function Ra(e, t, n = /* @__PURE__ */ new Date()) {
	switch (t) {
		case "relative": return Va(e, n);
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
		default: return za(e);
	}
}
function za(e) {
	return `${e.getFullYear()}-${String(e.getMonth() + 1).padStart(2, "0")}-${String(e.getDate()).padStart(2, "0")} ${String(e.getHours()).padStart(2, "0")}:${String(e.getMinutes()).padStart(2, "0")}`;
}
var Ba = [
	["y", 31536e3],
	["w", 604800],
	["d", 86400],
	["h", 3600],
	["m", 60]
];
function Va(e, t = /* @__PURE__ */ new Date(), n = {}) {
	let r = t.getTime() - e.getTime(), i = r < 0, a = Math.abs(r);
	if (n.style === "token") {
		let e = Math.floor(a / 1e3), t = Ba.find(([, t]) => e >= t);
		if (!t) return "now";
		let n = `${Math.floor(e / t[1])}${t[0]}`;
		return i ? `in ${n}` : n;
	}
	let o = Math.round(a / 1e3), s;
	return s = o < 1 ? "just now" : o < 60 ? `${o}s` : o < 3600 ? `${Math.round(o / 60)}m` : o < 86400 ? `${Math.round(o / 3600)}h` : `${Math.round(o / 86400)}d`, s === "just now" ? s : i ? `in ${s}` : `${s} ago`;
}
function Ha(e, t = /* @__PURE__ */ new Date()) {
	let n = Da(e);
	if (!n) return typeof e == "string" ? e : "";
	let r = n.toLocaleString(void 0, {
		year: "numeric",
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit"
	}), i = Va(n, t);
	return i === "just now" ? r : `${r} (${i})`;
}
function Ua(e, t) {
	switch (e) {
		case "absolute": return "iso";
		case "relative": return "relative";
		case "time": return "time";
		default: return t;
	}
}
// ../../../clicky-ui/packages/ui/src/overlay/HoverCard.tsx
var Wa = 6, Ga = 120;
function Ka({ trigger: e, children: t, placement: n = "top", delay: r = 0, arrow: a = !0, className: o, cardClassName: s }) {
	let [c, l] = (0, O.useState)(!1), [u, d] = (0, O.useState)(null), f = $t(), p = (0, O.useRef)(null), m = (0, O.useRef)(null), h = (0, O.useRef)(null), g = (0, O.useRef)(null), _ = () => {
		g.current !== null && (window.clearTimeout(g.current), g.current = null);
	};
	function v() {
		_(), r > 0 ? (h.current !== null && window.clearTimeout(h.current), h.current = window.setTimeout(() => l(!0), r)) : l(!0);
	}
	function y() {
		h.current !== null && (window.clearTimeout(h.current), h.current = null), _(), g.current = window.setTimeout(() => l(!1), Ga);
	}
	(0, O.useLayoutEffect)(() => {
		if (!c) return;
		let e = () => {
			let e = p.current, t = m.current;
			if (!e || !t) return;
			let r = e.getBoundingClientRect(), i = t.getBoundingClientRect(), a = 0, o = 0;
			switch (n) {
				case "top":
					a = r.top - i.height - Wa, o = r.left + r.width / 2 - i.width / 2;
					break;
				case "bottom":
					a = r.bottom + Wa, o = r.left + r.width / 2 - i.width / 2;
					break;
				case "left":
					a = r.top + r.height / 2 - i.height / 2, o = r.left - i.width - Wa;
					break;
				case "right": a = r.top + r.height / 2 - i.height / 2, o = r.right + Wa;
			}
			o = Math.min(Math.max(4, o), window.innerWidth - i.width - 4), a = Math.min(Math.max(4, a), window.innerHeight - i.height - 4), d({
				top: a,
				left: o
			});
		};
		return e(), window.addEventListener("scroll", e, !0), window.addEventListener("resize", e), () => {
			window.removeEventListener("scroll", e, !0), window.removeEventListener("resize", e);
		};
	}, [
		c,
		n,
		t
	]), (0, O.useEffect)(() => () => {
		h.current !== null && window.clearTimeout(h.current), g.current !== null && window.clearTimeout(g.current);
	}, []);
	let b = c && typeof document < "u" ? (0, It.createPortal)(/* @__PURE__ */ (0, A.jsxs)("div", {
		ref: m,
		role: "tooltip",
		onMouseEnter: _,
		onMouseLeave: y,
		style: {
			...u ? {
				position: "fixed",
				top: u.top,
				left: u.left
			} : {
				position: "fixed",
				top: -9999,
				left: -9999,
				visibility: "hidden"
			},
			zIndex: f
		},
		className: i("rounded-md border border-border bg-background px-2.5 py-1.5 text-[11px] shadow-lg", "whitespace-nowrap", s),
		children: [t, a && u && /* @__PURE__ */ (0, A.jsx)(qa, { placement: n })]
	}), document.body) : null;
	return /* @__PURE__ */ (0, A.jsxs)("span", {
		ref: p,
		className: i("relative inline-flex items-center", o),
		onMouseEnter: v,
		onMouseLeave: y,
		onFocus: v,
		onBlur: y,
		children: [e, b]
	});
}
function qa({ placement: e }) {
	let t = "absolute h-1.5 w-1.5 rotate-45 border bg-background border-border";
	switch (e) {
		case "top": return /* @__PURE__ */ (0, A.jsx)("span", {
			"aria-hidden": !0,
			className: i(t, "left-1/2 -translate-x-1/2 -bottom-0.5 border-l-0 border-t-0")
		});
		case "bottom": return /* @__PURE__ */ (0, A.jsx)("span", {
			"aria-hidden": !0,
			className: i(t, "left-1/2 -translate-x-1/2 -top-0.5 border-r-0 border-b-0")
		});
		case "left": return /* @__PURE__ */ (0, A.jsx)("span", {
			"aria-hidden": !0,
			className: i(t, "top-1/2 -translate-y-1/2 -right-0.5 border-l-0 border-b-0")
		});
		case "right": return /* @__PURE__ */ (0, A.jsx)("span", {
			"aria-hidden": !0,
			className: i(t, "top-1/2 -translate-y-1/2 -left-0.5 border-r-0 border-t-0")
		});
	}
}
// ../../../clicky-ui/packages/ui/src/data/TabButton.tsx
var Ja = {
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
function Ya({ active: e, onClick: t, label: n, icon: r, count: a, countColor: o = "bg-muted-foreground", variant: c = "pill", className: l }) {
	let u = Ja[c];
	return /* @__PURE__ */ (0, A.jsxs)("button", {
		type: "button",
		role: "tab",
		"aria-selected": e,
		onClick: t,
		className: i("inline-flex items-center gap-1.5 px-density-3 py-density-1.5 text-sm transition-colors", u.base, e ? u.active : u.inactive, l),
		children: [
			r && /* @__PURE__ */ (0, A.jsx)(s, { ...typeof r == "string" ? { name: r } : { icon: r } }),
			/* @__PURE__ */ (0, A.jsx)("span", { children: n }),
			a !== void 0 && a > 0 && /* @__PURE__ */ (0, A.jsx)("span", {
				className: i("inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded-full text-[10px] font-bold text-white", o),
				children: a
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/SplitButton.tsx
function Xa({ label: e, onClick: t, items: n, icon: r, variant: a = "default", size: c = "default", loading: l = !1, disabled: u = !1, primaryDisabled: d = !1, align: f = "right", title: p, className: m, menuClassName: h }) {
	let [g, _] = (0, O.useState)(!1);
	return n.length === 0 ? /* @__PURE__ */ (0, A.jsxs)(o, {
		variant: a,
		size: c,
		loading: l,
		disabled: u || d,
		onClick: t,
		className: m,
		children: [!l && r && /* @__PURE__ */ (0, A.jsx)(s, { ...typeof r == "string" ? { name: r } : { icon: r } }), e]
	}) : /* @__PURE__ */ (0, A.jsxs)("div", {
		className: i("inline-flex items-stretch", m),
		children: [/* @__PURE__ */ (0, A.jsxs)(o, {
			variant: a,
			size: c,
			loading: l,
			disabled: u || d,
			onClick: t,
			className: "rounded-r-none",
			children: [!l && r && /* @__PURE__ */ (0, A.jsx)(s, { ...typeof r == "string" ? { name: r } : { icon: r } }), e]
		}), /* @__PURE__ */ (0, A.jsx)(wa, {
			items: n,
			align: f,
			className: "items-stretch",
			...h ? { menuClassName: h } : {},
			onOpenChange: _,
			trigger: /* @__PURE__ */ (0, A.jsx)(o, {
				variant: a,
				size: c === "default" ? "icon" : c,
				disabled: u,
				title: p,
				"aria-label": p ?? "Open menu",
				className: i("h-auto self-stretch rounded-l-none border-l border-l-current/25 px-2", c === "xs" ? "w-6 px-0" : c !== "icon" && "w-8"),
				children: /* @__PURE__ */ (0, A.jsx)(s, { icon: g ? be : ve })
			})
		})]
	});
}
export { we as $, ua as A, ut as B, Qi as C, sa as D, ea as E, Rt as F, Me as G, Le as H, bt as I, ke as J, je as K, ht as L, Qt as M, $t as N, jr as O, Lt as P, Te as Q, mt as R, Xi as S, ti as T, Fe as U, Ze as V, Ne as W, Ee as X, Oe as Y, De as Z, Pr as _, La as a, _e as at, ii as b, Ra as c, ja as d, Se as et, wa as f, Hi as g, _a as h, Ea as i, ve as it, cn as j, la as k, Ua as l, ha as m, Ya as n, be as nt, Ha as o, Ca as p, Ae as q, Ka as r, ye as rt, Va as s, Xa as t, xe as tt, Da as u, ri as v, ni as w, Ki as x, ji as y, lt as z };
