import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./jsx-runtime-lVm_Myqd.js";
import { t as r } from "./react-dom-BBUmKE87.js";
import { T as i, _ as a, a as o, f as s, g as c, h as l, m as u, t as d, u as f, w as p, y as m } from "./Callout-DmCAGTzW.js";
import { _ as h, a as g, b as _, c as v, d as y, f as b, g as x, h as S, l as C, m as w, n as T, o as E, p as D, s as O, u as k, v as A, y as j } from "./floating-ui.react-dom-BkwMuSvQ.js";
// ../../../clicky-ui/packages/ui/src/lib/size.ts
var M = /* @__PURE__ */ e(t(), 1), N = n(), P = {
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
function F(e, t = "comfortable") {
	return P[t][e];
}
// ../../../clicky-ui/packages/ui/src/hooks/use-density.ts
var I = (0, M.createContext)(null);
function L() {
	let e = (0, M.useContext)(I);
	if (e) return e.density;
	if (typeof document > "u") return "comfortable";
	let t = document.documentElement.getAttribute("data-density");
	return t === "compact" || t === "spacious" ? t : "comfortable";
}
// ../../../clicky-ui/packages/ui/src/data/icon-fallback.ts
var ee = null;
function R() {
	return ee;
}
// ../../../clicky-ui/packages/ui/src/data/Icon.tsx
var z = {
	emerald: "text-emerald-700 ring-emerald-200/80 dark:text-emerald-400 dark:ring-emerald-500/30",
	amber: "text-amber-700 ring-amber-200/80 dark:text-amber-400 dark:ring-amber-500/30",
	rose: "text-rose-700 ring-rose-200/80 dark:text-rose-400 dark:ring-rose-500/30",
	slate: "text-slate-700 ring-slate-200/80 dark:text-slate-300 dark:ring-slate-500/30",
	sky: "text-sky-700 ring-sky-200/80 dark:text-sky-400 dark:ring-sky-500/30",
	violet: "text-violet-700 ring-violet-200/80 dark:text-violet-400 dark:ring-violet-500/30",
	neutral: "text-foreground ring-border"
}, te = "1em";
function B(e, t, n, r, i, a, o, s) {
	let c = [
		typeof r == "number" ? `rotate(${r}deg)` : typeof r == "string" ? r : "",
		i === "horizontal" || i === "both" ? "scaleX(-1)" : "",
		i === "vertical" || i === "both" ? "scaleY(-1)" : ""
	].filter(Boolean).join(" ");
	return /* @__PURE__ */ (0, N.jsx)(e, {
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
function V(e, t, n, r, a) {
	let o = R();
	if (o) {
		let i = o, s = typeof t == "number" ? t : n;
		return /* @__PURE__ */ (0, N.jsx)(i, {
			name: e,
			...r ? { className: r } : {},
			...s == null ? {} : { size: s },
			...a ? { alt: a } : {}
		});
	}
	return /* @__PURE__ */ (0, N.jsx)("span", {
		className: i("inline-flex shrink-0 items-center justify-center rounded-sm border border-dashed border-muted-foreground/50 text-[0.65em] leading-none text-muted-foreground", r),
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
function H({ icon: e, name: t, className: n, width: r, height: a, rotate: o, flip: s, inline: c, title: l, style: u = "plain", size: d, tone: f = "neutral" }) {
	let p = L(), m = d == null ? void 0 : F(d, p), h = m == null ? void 0 : Math.max(8, Math.round(m * .5)), g = r ?? a ?? (u === "badge" ? h : m ?? te), _ = a ?? g, v = i("inline-block shrink-0 align-[-0.125em]", u === "plain" ? n : void 0), y = u === "plain" ? l : void 0, b = e ? B(e, g, _, o, s, c, v, y) : V(t, g, _, v, y);
	if (u === "plain") return b;
	let x = m == null ? void 0 : {
		width: m,
		height: m
	};
	return /* @__PURE__ */ (0, N.jsx)("span", {
		title: l,
		"aria-label": l,
		role: l ? "img" : void 0,
		className: i("inline-flex items-center justify-center rounded-full bg-background shadow-sm ring-1 ring-inset", z[f], n),
		style: x,
		children: b
	});
}
function ne({ icon: e, className: t }) {
	return e == null || e === "" || e === !1 ? null : typeof e == "string" ? /* @__PURE__ */ (0, N.jsx)(H, {
		name: e,
		className: i("shrink-0", t)
	}) : /* @__PURE__ */ (0, N.jsx)(N.Fragment, { children: e });
}
// ../../../clicky-ui/packages/ui/src/data/badge-variants.ts
var U = p("inline-flex items-center gap-1 rounded-full border border-transparent font-medium whitespace-nowrap align-middle", {
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
}), re = /* @__PURE__ */ new Set([
	"soft",
	"solid",
	"outline"
]), W = /* @__PURE__ */ new Set([
	"status",
	"metric",
	"custom",
	"outlined",
	"label"
]), ie = {
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
}, G = {
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
function ae(e) {
	return e != null && re.has(e);
}
function oe(e) {
	return e != null && /^(#|rgb|hsl|oklch|var\(|color\()/i.test(e);
}
function se(e) {
	switch (e) {
		case "success": return "success";
		case "danger": return "error";
		case "warning": return "warning";
		case "info": return "info";
		default: return;
	}
}
function ce(e) {
	switch (e) {
		case "rounded": return "rounded-md";
		case "square": return "rounded-sm";
		default: return "rounded-full";
	}
}
function le(e) {
	return e != null && W.has(e);
}
function ue(e, t, n, r) {
	e != null && (oe(e) ? n[t] = e : r.push(e));
}
function de(e) {
	if (e != null) return typeof e == "number" ? `${Math.max(1, Math.floor(e))}ch` : e;
}
function fe(e, t = 24) {
	if (typeof e == "number" && Number.isFinite(e) && e > 0) return Math.max(4, Math.floor(e));
	if (typeof e == "string") {
		let t = e.trim(), n = t.match(/^(\d+(?:\.\d+)?)ch$/i);
		if (n) return Math.max(4, Math.floor(Number(n[1])));
		let r = t.match(/^(\d+(?:\.\d+)?)$/);
		if (r) return Math.max(4, Math.floor(Number(r[1])));
	}
	return t;
}
function pe({ maxWidth: e, truncate: t, wrap: n }) {
	return e != null || t != null || n;
}
function me(e) {
	if (e == null || typeof e == "boolean") return null;
	if (typeof e == "string" || typeof e == "number") return String(e);
	if (Array.isArray(e)) {
		let t = M.Children.toArray(e).map(me);
		return t.every((e) => e != null) ? t.join("") : null;
	}
	return (0, M.isValidElement)(e) ? me(e.props.children) : null;
}
function he(e, t) {
	return e.length <= t ? e : t <= 1 ? "…" : `${e.slice(0, t - 1)}…`;
}
function ge(e, t) {
	return e.length <= t ? e : t <= 1 ? "…" : `…${e.slice(-(t - 1))}`;
}
function _e(e, t) {
	if (e.length <= t) return e;
	if (t <= 3) return he(e, t);
	let n = t - 1, r = Math.max(1, Math.ceil(n / 2)), i = Math.max(1, n - r);
	return `${e.slice(0, r)}…${e.slice(-i)}`;
}
function ve(e, t) {
	return e.length <= t ? e : t <= 2 ? he(e, t) : `${e.slice(0, t - 1)}…`;
}
function ye(e, t) {
	return e.length <= t ? e : t <= 2 ? ge(e, t) : `…${e.slice(-(t - 1))}`;
}
function be(e) {
	return e == null || e.length === 0 ? "…" : e.length <= 2 ? `${e}…` : `${e[0]}…`;
}
function xe(e, t, n, r, i) {
	let a = [e];
	t && a.push(t), a.push(n);
	let o = a.join(i);
	if (o.length <= r) return o;
	let s = a.length - 1, c = i.length * s + (t ? t.length : 0), l = Math.max(4, r - c), u = Math.max(3, Math.min(e.length, Math.floor(l * .4))), d = Math.max(4, l - u), f = ve(e, u), p = ye(n, d), m = [f];
	t && m.push(t), m.push(p);
	let h = m.join(i);
	return h.length <= r ? h : _e(o, r);
}
function Se(e) {
	return Math.max(e, 32);
}
function Ce(e, t) {
	let n = Se(t), r = e.startsWith("/"), i = e.split("/").filter(Boolean);
	if (i.length < 2) return _e(e, n);
	let a = `${r ? "/" : ""}${i[0]}`, o = i[i.length - 1] ?? e;
	return xe(a, i.length > 2 ? be(i.slice(1, -1).join("/")) : void 0, o, n, "/");
}
function we(e, t) {
	let n = Se(t);
	try {
		let t = new URL(e), r = t.host, i = t.pathname.split("/").filter(Boolean);
		if (i.length === 0) return r.length <= n ? r : _e(`${r}${t.search}${t.hash}`, n);
		if (i.length === 1) {
			let e = `${r}/${i[0] ?? r}`;
			if (e.length <= n) return e;
			let t = `${r}/…`;
			return t.length <= n ? t : _e(e, n);
		}
		let a = i[i.length - 1] ?? r;
		return [
			[
				r,
				i[0] ?? "",
				i.length > 2 ? be(i.slice(1, -1).join("/")) : void 0,
				a
			].filter(Boolean).join("/"),
			`${r}/…/${a}`,
			`${r}/…`
		].find((e) => e.length <= n) ?? _e(`${r}/${a}`, n);
	} catch {
		return _e(e, n);
	}
}
function Te(e) {
	return /^[a-z][a-z\d+\-.]*:\/\//i.test(e) || e.startsWith("//");
}
function Ee(e) {
	if (/\s/.test(e) || Te(e) || e.startsWith("/") || !e.includes("/")) return !1;
	let t = e.slice(e.lastIndexOf("/") + 1);
	return t.includes(":") || t.includes("@sha256:");
}
function De(e) {
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
function Oe(e, t) {
	let n = Se(t), r = e.split("/").filter(Boolean);
	if (r.length < 2) return _e(e, n);
	let i = r[r.length - 1] ?? e, a = De(i), o = `…${a.name}`;
	if (o.length <= n) return o;
	if (r.length === 2) return xe(r[0] ?? e, void 0, a.name || i, n, "/");
	let s = `${r[0]}/${r[1] ?? ""}`.replace(/\/$/, ""), c = a.suffix.length > 0 && n >= i.length ? i : a.name || i;
	return xe(s, r.length > 3 ? be(r.slice(2, -1).join("/")) : r.length === 3 ? "…" : void 0, c, n, "/");
}
function ke(e, t) {
	let n = Se(t), r = e.split(":");
	if (r.length < 6) return _e(e, n);
	let i = r[2] ?? r[1] ?? "arn", a = r.slice(5).join(":"), o = a.split(/[/:]/).filter(Boolean), s = o.length > 1 ? o[0] : void 0, c = o[o.length - 1] ?? a, l = c.includes("-") && c.split("-").length > 2 ? c.split("-").slice(-2).join("-") : c, u = `${i}:…${s != null && s !== l ? `${s}/${l}` : l}`;
	if (u.length <= n) return u;
	let d = Math.max(4, Math.min(l.length, n - 2 - 2));
	return `${ve(i, Math.max(2, n - d - 2))}:…${_e(l, d)}`;
}
function Ae(e) {
	return e.startsWith("arn:") ? "arn" : Te(e) ? "url" : Ee(e) ? "image" : e.includes("/") ? "path" : "suffix";
}
function je(e, t, n) {
	let r = fe(n);
	switch (t === "auto" ? Ae(e) : t) {
		case "prefix": return ge(e, r);
		case "suffix": return he(e, r);
		case "arn": return ke(e, r);
		case "image": return Oe(e, r);
		case "path": return Ce(e, r);
		case "url": return we(e, r);
		default: return he(e, r);
	}
}
function Me(e, t, n, r = !1) {
	let i = me(e);
	if (i == null || t == null) return {
		content: e,
		fullText: i,
		title: r ? i ?? void 0 : void 0
	};
	let a = je(i, t, n);
	return {
		content: a,
		fullText: i,
		title: r || a !== i ? i : void 0
	};
}
function Ne(e) {
	let t = globalThis.navigator?.clipboard;
	t?.writeText != null && t.writeText(e).catch(() => void 0);
}
function Pe({ href: e, target: t, rel: n, className: r, style: i, content: a, canCopy: o, copyText: s }) {
	return e == null ? o && s != null ? /* @__PURE__ */ (0, N.jsx)("button", {
		type: "button",
		className: r,
		style: i,
		onClick: () => Ne(s),
		children: a
	}) : /* @__PURE__ */ (0, N.jsx)("span", {
		className: r,
		style: i,
		children: a
	}) : /* @__PURE__ */ (0, N.jsx)("a", {
		href: e,
		target: t,
		rel: n ?? (t === "_blank" ? "noopener noreferrer" : void 0),
		className: r,
		style: i,
		children: a
	});
}
function Fe({ tone: e, variant: t, size: n = "md", icon: r, count: a, children: o, className: s, status: c, label: l, value: u, color: d, textColor: f, borderColor: p, shape: m, href: h, target: g = "_self", rel: _, wrap: v = !1, maxWidth: y, truncate: b, clickToCopy: x, labelClassName: S, valueClassName: C }) {
	let w = t ?? (l != null || u != null || c != null || m != null || h != null || v === !0 || d != null || f != null || p != null || S != null || C != null ? "metric" : "soft"), T = de(y), E = v && b == null, D = pe({
		maxWidth: y,
		truncate: b,
		wrap: v
	}), O = E ? "min-w-0 whitespace-normal break-words [overflow-wrap:anywhere]" : T != null || b != null ? "min-w-0 overflow-hidden text-ellipsis whitespace-nowrap" : "whitespace-nowrap";
	if (ae(w)) {
		let t = {};
		T != null && (t.maxWidth = T);
		let c = Me(o, b, y, D), l = c.fullText ?? (a === void 0 ? void 0 : String(a)), u = h == null && (x ?? !0) && l != null;
		return Pe({
			href: h,
			target: g,
			rel: _,
			canCopy: u,
			copyText: l,
			className: i(U({
				tone: e,
				variant: w,
				size: n
			}), (T != null || E || b != null) && "min-w-0 max-w-full", E && "whitespace-normal", u && "cursor-pointer text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1", s),
			style: t,
			content: /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [
				r && /* @__PURE__ */ (0, N.jsx)(H, {
					...typeof r == "string" ? { name: r } : { icon: r },
					className: "text-[1em]"
				}),
				a !== void 0 && /* @__PURE__ */ (0, N.jsx)("span", { children: a }),
				o != null && /* @__PURE__ */ (0, N.jsx)("span", {
					className: O,
					title: c.title,
					children: c.content
				})
			] })
		});
	}
	if (!le(w)) return null;
	let k = c ?? se(e) ?? "info", A = ie[n], j = l ?? o, M = u ?? (j == null && a !== void 0 ? a : void 0), P = M == null ? Me(j, b, y, D) : {
		content: j,
		fullText: me(j),
		title: D ? me(j) ?? void 0 : void 0
	}, F = M == null ? {
		content: M,
		fullText: void 0,
		title: void 0
	} : Me(M, b, y, D), I = F.fullText ?? P.fullText ?? (a === void 0 ? void 0 : String(a)), L = h == null && (x ?? !0) && I != null, ee = ce(m ?? "pill"), R = {}, z = [
		"inline-flex align-middle items-stretch border font-medium shadow-none",
		A.frame,
		A.text,
		T != null || E || b != null ? "min-w-0 max-w-full" : "",
		ee,
		E ? "whitespace-normal" : "whitespace-nowrap",
		h ? "transition-opacity hover:opacity-80" : "",
		L ? "cursor-pointer text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1" : ""
	];
	T != null && (R.maxWidth = T), w === "status" ? z.push(G[k].soft) : w === "metric" ? z.push("bg-muted/60 text-foreground border-border") : w === "outlined" ? c != null || se(e) != null ? z.push(G[k].outlined) : z.push("bg-background text-foreground border-border") : w === "custom" ? z.push("border-transparent bg-accent text-accent-foreground") : z.push("border-border bg-background text-foreground"), (w === "custom" || w === "outlined" && c == null && se(e) == null) && (ue(d, "backgroundColor", R, z), ue(f, "color", R, z), ue(p, "borderColor", R, z)), w !== "label" && p != null && c == null && se(e) == null && ue(p, "borderColor", R, z);
	let te = r ? /* @__PURE__ */ (0, N.jsx)(H, {
		...typeof r == "string" ? { name: r } : { icon: r },
		className: i("shrink-0", A.icon)
	}) : null;
	if (w === "label") {
		let e = j != null || te != null, t = [
			"inline-flex items-center self-stretch",
			M == null ? "min-w-0" : "shrink-0",
			A.segment,
			A.gap,
			M == null ? "" : "border-r border-border/70"
		], n = [
			"inline-flex min-w-0 items-center self-stretch",
			f == null || e ? "text-foreground" : "",
			e ? "flex-1" : "",
			A.segment
		], r = {};
		return e ? (t.push("bg-secondary text-secondary-foreground"), ue(d, "backgroundColor", r, t), ue(f, "color", r, t)) : (ue(d, "backgroundColor", R, z), ue(f, "color", R, z)), ue(p, "borderColor", R, z), Pe({
			href: h,
			target: g,
			rel: _,
			canCopy: L,
			copyText: I,
			className: i(z, s),
			style: R,
			content: /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [e && /* @__PURE__ */ (0, N.jsxs)("span", {
				className: i(t, S),
				style: r,
				children: [te, j != null && /* @__PURE__ */ (0, N.jsx)("span", {
					className: O,
					title: P.title,
					children: P.content
				})]
			}), M != null && /* @__PURE__ */ (0, N.jsx)("span", {
				className: i(n, C),
				children: /* @__PURE__ */ (0, N.jsx)("span", {
					className: O,
					title: F.title,
					children: F.content
				})
			})] })
		});
	}
	let B = [
		"inline-flex items-center",
		M == null ? "min-w-0" : "shrink-0",
		A.segment,
		A.gap,
		M == null ? "" : "border-r border-current/10"
	], V = [
		"inline-flex min-w-0 items-center font-semibold",
		j != null || te != null ? "flex-1" : "",
		A.segment
	];
	return w === "metric" && (B.push("text-muted-foreground"), V.push("text-foreground")), Pe({
		href: h,
		target: g,
		rel: _,
		canCopy: L,
		copyText: I,
		className: i(z, s),
		style: R,
		content: /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [(j != null || te != null) && /* @__PURE__ */ (0, N.jsxs)("span", {
			className: i(B, S),
			children: [te, j != null && /* @__PURE__ */ (0, N.jsx)("span", {
				className: O,
				title: P.title,
				children: P.content
			})]
		}), M != null && /* @__PURE__ */ (0, N.jsx)("span", {
			className: i(V, C),
			children: /* @__PURE__ */ (0, N.jsx)("span", {
				className: O,
				title: F.title,
				children: F.content
			})
		})] })
	});
}
// ../../../clicky-ui/packages/ui/src/hooks/use-media-query.ts
function Ie(e) {
	let t = ze(e);
	return t ? t.matches : !1;
}
function Le(e, t) {
	let n = ze(e);
	if (!n) return () => {};
	let r = () => t(n.matches);
	return n.addEventListener ? (n.addEventListener("change", r), () => n.removeEventListener("change", r)) : (n.addListener?.(r), () => n.removeListener?.(r));
}
function Re(e) {
	let [t, n] = (0, M.useState)(() => Ie(e));
	return (0, M.useEffect)(() => (n(Ie(e)), Le(e, n)), [e]), t;
}
function ze(e) {
	return typeof window > "u" || typeof window.matchMedia != "function" ? null : window.matchMedia(e);
}
// ../../../clicky-ui/packages/ui/src/hooks/use-theme.ts
var Be = "(prefers-color-scheme: dark)", Ve = (0, M.createContext)(null);
function He() {
	return Ie(Be);
}
function Ue() {
	return (0, M.useContext)(Ve);
}
function We(e) {
	let t = Ue(), [n, r] = (0, M.useState)(() => He() ? "dark" : "light");
	return (0, M.useEffect)(() => Le(Be, (e) => r(e ? "dark" : "light")), []), e && e !== "system" ? e : e === "system" ? n : t ? t.resolvedTheme : n;
}
// ../../../clicky-ui/packages/ui/src/icons/components/UiChevronDown.tsx
var Ge = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var Ke = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var qe = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiClose.tsx
var K = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiAdd.tsx
var Je = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var Ye = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
		fill: "currentColor",
		d: "M214 48v40a6 6 0 0 1-12 0V54h-34a6 6 0 0 1 0-12h40a6 6 0 0 1 6 6M88 202H54v-34a6 6 0 0 0-12 0v40a6 6 0 0 0 6 6h40a6 6 0 0 0 0-12m120-40a6 6 0 0 0-6 6v34h-34a6 6 0 0 0 0 12h40a6 6 0 0 0 6-6v-40a6 6 0 0 0-6-6M88 42H48a6 6 0 0 0-6 6v40a6 6 0 0 0 12 0V54h34a6 6 0 0 0 0-12"
	})
}), {
	__source: "ph:corners-out-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "fullscreen",
	displayName: "UiFullscreen"
}), Xe = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
		fill: "currentColor",
		d: "M93.66 202.34A8 8 0 0 1 88 216H48a8 8 0 0 1-8-8v-40a8 8 0 0 1 13.66-5.66ZM88 40H48a8 8 0 0 0-8 8v40a8 8 0 0 0 13.66 5.66l40-40A8 8 0 0 0 88 40m123.06 120.61a8 8 0 0 0-8.72 1.73l-40 40A8 8 0 0 0 168 216h40a8 8 0 0 0 8-8v-40a8 8 0 0 0-4.94-7.39M208 40h-40a8 8 0 0 0-5.66 13.66l40 40A8 8 0 0 0 216 88V48a8 8 0 0 0-8-8"
	})
}), {
	__source: "ph:corners-out-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "fullscreen",
	displayName: "UiFullscreenFilled"
}), Ze = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var Qe = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiRestart.tsx
var $e = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
		fill: "currentColor",
		d: "M14 8a5.98 5.98 0 0 1-1.759 4.243A5.98 5.98 0 0 1 8 14a6 6 0 0 1-4.243-1.758a6 6 0 0 1-1.285-1.905a.75.75 0 1 1 1.38-.586c.229.537.553 1.019.966 1.43c.412.414.894.738 1.431.967a4.55 4.55 0 0 0 3.503 0a4.5 4.5 0 0 0 1.429-.965A4.5 4.5 0 0 0 12.501 8a4.48 4.48 0 0 0-1.319-3.181a4.5 4.5 0 0 0-1.431-.967a4.55 4.55 0 0 0-3.503 0A4.5 4.5 0 0 0 4.257 5.5h1.992a.75.75 0 0 1 0 1.5h-3.5a.75.75 0 0 1-.75-.75v-3.5a.75.75 0 0 1 1.5 0v1.282q.124-.142.259-.275a6 6 0 0 1 1.904-1.285a6.04 6.04 0 0 1 4.673 0a6 6 0 0 1 1.907 1.286A6 6 0 0 1 13.999 8z"
	})
}), {
	__source: "codicon:debug-restart",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "restart",
	displayName: "UiRestart"
}), et = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var tt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var nt = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#059669",
		...r
	};
	return /* @__PURE__ */ (0, N.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, N.jsx)("path", {
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
});
Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#059669",
		...r
	};
	return /* @__PURE__ */ (0, N.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, N.jsx)("path", {
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
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiLock.tsx
var rt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var it = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var at = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiComment.tsx
var ot = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsxs)("g", {
		fill: "none",
		children: [
			/* @__PURE__ */ (0, N.jsx)("path", {
				d: "M8.5 10C8.77614 10 9 10.2239 9 10.5C9 10.7761 8.77614 11 8.5 11H5.5C5.22386 11 5 10.7761 5 10.5C5 10.2239 5.22386 10 5.5 10H8.5Z",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, N.jsx)("path", {
				d: "M8.5 8C8.77614 8 9 8.22386 9 8.5C9 8.77614 8.77614 9 8.5 9H5.5C5.22386 9 5 8.77614 5 8.5C5 8.22386 5.22386 8 5.5 8H8.5Z",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, N.jsx)("path", {
				d: "M8.5 6C8.77614 6 9 6.22386 9 6.5C9 6.77614 8.77614 7 8.5 7H5.5C5.22386 7 5 6.77614 5 6.5C5 6.22386 5.22386 6 5.5 6H8.5Z",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, N.jsx)("path", {
				fillRule: "evenodd",
				clipRule: "evenodd",
				d: "M10 3C11.1046 3 12 3.89543 12 5V12L11.9893 12.2041C11.8938 13.1457 11.1457 13.8938 10.2041 13.9893L10 14H4L3.7959 13.9893C2.85435 13.8938 2.1062 13.1457 2.01074 12.2041L2 12V5C2 3.89543 2.89543 3 4 3H10ZM4 4C3.44772 4 3 4.44772 3 5V12C3 12.5523 3.44772 13 4 13H10C10.5523 13 11 12.5523 11 12V5C11 4.44772 10.5523 4 10 4H4Z",
				fill: "#6C707E"
			}),
			/* @__PURE__ */ (0, N.jsx)("path", {
				d: "M11 1C12.6569 1 14 2.34315 14 4V10.5C14 10.7761 13.7761 11 13.5 11C13.2239 11 13 10.7761 13 10.5V4C13 2.89543 12.1046 2 11 2H5.5C5.22386 2 5 1.77614 5 1.5C5 1.22386 5.22386 1 5.5 1H11Z",
				fill: "#6C707E"
			})
		]
	})
}), {
	__source: "jb-expui-toolwindows:documentation",
	__viewBox: "0 0 16 16",
	__group: "uir-ast-code",
	__consumerName: "uir-comment",
	displayName: "UiComment"
}), st = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var ct = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiDownload.tsx
var lt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiRobotAi.tsx
var ut = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
		fill: "currentColor",
		d: "M200 50h-66V16a6 6 0 0 0-12 0v34H56a30 30 0 0 0-30 30v112a30 30 0 0 0 30 30h144a30 30 0 0 0 30-30V80a30 30 0 0 0-30-30m18 142a18 18 0 0 1-18 18H56a18 18 0 0 1-18-18V80a18 18 0 0 1 18-18h144a18 18 0 0 1 18 18ZM74 108a10 10 0 1 1 10 10a10 10 0 0 1-10-10m88 0a10 10 0 1 1 10 10a10 10 0 0 1-10-10m2 30H92a26 26 0 0 0 0 52h72a26 26 0 0 0 0-52m-22 12v28h-28v-28Zm-64 14a14 14 0 0 1 14-14h10v28H92a14 14 0 0 1-14-14m86 14h-10v-28h10a14 14 0 0 1 0 28"
	})
}), {
	__source: "ph:robot-light",
	__viewBox: "0 0 256 256",
	__group: "ai-ml",
	__consumerName: "robot-ai",
	displayName: "UiRobotAi"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
		fill: "currentColor",
		d: "M200 48h-64V16a8 8 0 0 0-16 0v32H56a32 32 0 0 0-32 32v112a32 32 0 0 0 32 32h144a32 32 0 0 0 32-32V80a32 32 0 0 0-32-32m-28 48a12 12 0 1 1-12 12a12 12 0 0 1 12-12m-76 88H80a16 16 0 0 1 0-32h16Zm-12-64a12 12 0 1 1 12-12a12 12 0 0 1-12 12m60 64h-32v-32h32Zm32 0h-16v-32h16a16 16 0 0 1 0 32"
	})
}), {
	__source: "ph:robot-fill",
	__viewBox: "0 0 256 256",
	__group: "ai-ml",
	__consumerName: "robot-ai",
	displayName: "UiRobotAiFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiTrash.tsx
var dt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/data/code-block-actions.tsx
var ft = {
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
function pt(e, t, n) {
	let r = t && ft[t.toLowerCase()] || "txt", i = new Blob([e], { type: "text/plain;charset=utf-8" }), a = URL.createObjectURL(i), o = document.createElement("a");
	o.href = a, o.download = n || `snippet.${r}`, o.click(), URL.revokeObjectURL(a);
}
function mt({ source: e, language: t, copyable: n = !1, downloadable: r = !1, downloadFilename: a, theme: o, onToggleTheme: s }) {
	let [c, l] = (0, M.useState)(!1);
	return !n && !r && !s ? null : /* @__PURE__ */ (0, N.jsxs)("div", {
		className: "flex items-center gap-1 text-muted-foreground",
		children: [
			s && /* @__PURE__ */ (0, N.jsx)("button", {
				type: "button",
				"aria-label": o === "dark" ? "Switch to light theme" : "Switch to dark theme",
				title: "Toggle theme",
				onClick: s,
				className: "rounded p-1 hover:bg-accent hover:text-accent-foreground",
				children: /* @__PURE__ */ (0, N.jsx)(H, {
					icon: o === "dark" ? Qe : Ze,
					className: "size-3.5"
				})
			}),
			r && /* @__PURE__ */ (0, N.jsx)("button", {
				type: "button",
				"aria-label": "Download code",
				title: "Download file",
				onClick: () => pt(e, t, a),
				className: "rounded p-1 hover:bg-accent hover:text-accent-foreground",
				children: /* @__PURE__ */ (0, N.jsx)(H, {
					icon: lt,
					className: "size-3.5"
				})
			}),
			n && /* @__PURE__ */ (0, N.jsx)("button", {
				type: "button",
				"aria-label": c ? "Copied" : "Copy code",
				title: "Copy code",
				onClick: async () => {
					try {
						await navigator.clipboard.writeText(e), l(!0), setTimeout(() => l(!1), 1500);
					} catch (e) {
						console.warn("clicky-ui: copy to clipboard failed", e);
					}
				},
				className: "rounded p-1 hover:bg-accent hover:text-accent-foreground",
				children: /* @__PURE__ */ (0, N.jsx)(H, {
					icon: c ? nt : ct,
					className: i("size-3.5", c && "text-emerald-600")
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/SegmentedControl.tsx
var ht = {
	sm: "text-xs px-density-2 py-density-1 gap-1",
	md: "h-full px-control-px text-sm gap-1.5",
	lg: "min-h-[4.25rem] min-w-44 flex-1 px-density-3 py-density-3 gap-density-2 text-sm"
};
function gt({ value: e, options: t, onChange: n, size: r = "md", wrap: a = !1, className: o, "aria-label": s }) {
	let c = (0, M.useRef)([]), l = (0, M.useId)(), u = t.some((t) => t.id === e), d = (e, r) => {
		let i = t.length, a = e;
		for (let e = 0; e < i; e++) {
			a = (a + r + i) % i;
			let e = t[a];
			if (e && !e.disabled) {
				n(e.id), c.current[a]?.focus();
				return;
			}
		}
	};
	return /* @__PURE__ */ (0, N.jsx)("div", {
		role: "radiogroup",
		"aria-label": s,
		className: i("inline-flex items-center rounded-md bg-muted p-0.5", r === "md" && "h-control-h", r === "lg" && "gap-density-2 bg-transparent p-0", a && "flex-wrap", o),
		children: t.map((t, a) => {
			let o = t.id === e, s = !u && t.implied === !0, f = t.marker ? `${l}-marker-${a}` : void 0;
			return /* @__PURE__ */ (0, N.jsxs)("button", {
				ref: (e) => {
					c.current[a] = e;
				},
				type: "button",
				role: "radio",
				"aria-checked": o,
				"aria-describedby": f,
				"data-implied": s ? "true" : void 0,
				title: t.title,
				disabled: t.disabled,
				tabIndex: o || s ? 0 : -1,
				onClick: () => !t.disabled && n(t.id),
				onKeyDown: (e) => {
					e.key === "ArrowRight" || e.key === "ArrowDown" ? (e.preventDefault(), d(a, 1)) : (e.key === "ArrowLeft" || e.key === "ArrowUp") && (e.preventDefault(), d(a, -1));
				},
				className: i("inline-flex items-center rounded-[5px] border border-transparent font-medium transition-colors", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring", "disabled:cursor-not-allowed disabled:opacity-50", r === "lg" && "rounded-lg text-left", t.description ? "whitespace-normal" : "whitespace-nowrap", ht[r], o ? r === "lg" ? "border-primary/70 bg-primary/10 text-foreground shadow-sm ring-1 ring-primary/25" : "border-border bg-card text-foreground shadow-sm" : s ? r === "lg" ? "border-dashed border-primary/60 bg-card text-foreground hover:shadow-sm" : "border-dashed border-primary/60 text-foreground hover:bg-background/60" : r === "lg" ? "border-border bg-card text-muted-foreground hover:border-muted-foreground/40 hover:bg-card hover:text-foreground hover:shadow-sm" : "text-foreground/80 hover:border-border hover:bg-background/60 hover:text-foreground", o && t.activeClassName),
				children: [t.icon && /* @__PURE__ */ (0, N.jsx)(H, {
					...typeof t.icon == "string" ? { name: t.icon } : { icon: t.icon },
					className: i(r === "lg" ? "size-4 shrink-0" : "size-3.5 shrink-0", r === "lg" && (o ? "text-primary" : "text-muted-foreground"), t.iconClassName)
				}), /* @__PURE__ */ (0, N.jsxs)("span", {
					className: i("min-w-0", r === "lg" && "flex flex-col items-start text-left leading-tight"),
					children: [t.marker ? /* @__PURE__ */ (0, N.jsxs)("span", {
						className: "inline-flex min-w-0 max-w-full items-baseline gap-1",
						children: [/* @__PURE__ */ (0, N.jsx)("span", {
							className: "min-w-0 truncate",
							children: t.label
						}), /* @__PURE__ */ (0, N.jsx)("span", {
							id: f,
							"aria-hidden": "true",
							className: "shrink-0 text-[11px] font-normal opacity-70",
							children: t.marker
						})]
					}) : /* @__PURE__ */ (0, N.jsx)("span", {
						className: "min-w-0 truncate",
						children: t.label
					}), t.description && /* @__PURE__ */ (0, N.jsx)("span", {
						className: i("mt-0.5 max-w-56 whitespace-normal break-words text-[11px] font-normal leading-snug", o ? "text-muted-foreground" : "text-muted-foreground/80"),
						children: t.description
					})]
				})]
			}, t.id);
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/code-diff.ts
function _t(e) {
	return e === "" ? [] : (e.endsWith("\n") ? e.slice(0, -1) : e).split("\n");
}
function vt(e, t) {
	let n = yt(_t(e), _t(t));
	return n.length > 0 ? [{ lines: n }] : [];
}
function yt(e, t) {
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
var bt = /^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@/;
function xt(e) {
	let t = e.split("	")[0].trim();
	if (t !== "" && t !== "/dev/null") return t.startsWith("a/") || t.startsWith("b/") ? t.slice(2) : t;
}
function St(e) {
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
		let c = bt.exec(e);
		if (!c) throw Error(`Malformed unified diff hunk header: ${e}`);
		r = Number(c[1]), i = Number(c[3]), a = c[2] ? Number(c[2]) : 1, o = c[4] ? Number(c[4]) : 1, n = {
			header: e,
			lines: [],
			...s ? { path: s } : {}
		}, t.push(n), a <= 0 && o <= 0 && (n = null);
	} else e.startsWith("diff --git ") ? (s = void 0, c = void 0) : e.startsWith("--- ") ? c = xt(e.slice(4)) : e.startsWith("+++ ") && (s = xt(e.slice(4)) ?? c);
	return t;
}
// ../../../clicky-ui/packages/ui/src/data/code-highlight.ts
var Ct = [
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
], wt = ["github-light", "github-dark"], Tt = "github-light", Et = null;
function Dt() {
	return Et ??= (async () => {
		let [{ createHighlighterCore: e }, { createOnigurumaEngine: t }] = await Promise.all([import("./core-BLykqbjA.js"), import("./engine-oniguruma-CHcoBf8c.js")]);
		return e({
			themes: [import("./github-light-CSQAsfih.js"), import("./github-dark-pW_68VqL.js")],
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
				import("./xml-sNMacCiP.js"),
				import("./html-DGpqF2Oy.js"),
				import("./css-BXHpbd9g.js")
			],
			engine: t(import("./wasm-26cD-7Fr.js"))
		});
	})(), Et;
}
function Ot(e) {
	if (!e) return;
	let t = e.slice(e.lastIndexOf("/") + 1), n = t.lastIndexOf(".");
	if (!(n <= 0)) return t.slice(n + 1).toLowerCase();
}
function kt(e) {
	let t = e.toLowerCase();
	return Ct.includes(t) ? t : t === "shell" || t === "sh" || t === "zsh" ? "bash" : t === "ts" ? "typescript" : t === "js" ? "javascript" : t === "py" ? "python" : t === "yml" ? "yaml" : null;
}
function At(e) {
	return e && wt.includes(e) ? e : Tt;
}
var jt = 8, Mt = [], Nt = !1;
function Pt(e) {
	return new Promise((t, n) => {
		Mt.push(() => {
			try {
				t(e());
			} catch (e) {
				n(e);
			}
		}), Nt || (Nt = !0, setTimeout(Ft, 0));
	});
}
function Ft() {
	let e = performance.now() + jt;
	for (; Mt.length > 0 && performance.now() < e;) Mt.shift()();
	Mt.length > 0 ? setTimeout(Ft, 0) : Nt = !1;
}
var It = 2e6, Lt = /* @__PURE__ */ new Map(), Rt = 0;
function zt(e, t) {
	Lt.set(e, t), Rt += t.length;
	for (let [e, t] of Lt) {
		if (Rt <= It) break;
		Lt.delete(e), Rt -= t.length;
	}
}
async function Bt(e, t) {
	if (!e || !t.lang) return null;
	let n = kt(t.lang);
	if (!n) return null;
	let r = At(t.theme), i = t.transformers ? void 0 : `${n} ${r} ${e}`, a = i === void 0 ? void 0 : Lt.get(i);
	if (i !== void 0 && a !== void 0) return Lt.delete(i), Lt.set(i, a), a;
	try {
		let a = await Dt(), o = await Pt(() => a.codeToHtml(e, {
			lang: n,
			theme: r,
			...t.transformers ? { transformers: t.transformers } : {}
		}));
		return i !== void 0 && zt(i, o), o;
	} catch {
		return null;
	}
}
async function Vt(e, t) {
	if (!e || !t.lang) return null;
	let n = kt(t.lang);
	if (!n) return null;
	try {
		let r = await Dt(), { tokens: i } = await Pt(() => r.codeToTokens(e, {
			lang: n,
			theme: At(t.theme)
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
// ../../../clicky-ui/packages/ui/src/data/token-style.ts
function Ht(e) {
	let t = {};
	e.color && (t.color = e.color);
	let n = e.fontStyle ?? 0;
	return n & 1 && (t.fontStyle = "italic"), n & 2 && (t.fontWeight = 700), n & 4 && (t.textDecorationLine = "underline"), t;
}
// ../../../clicky-ui/packages/ui/src/data/HighlightedTokens.tsx
function Ut({ tokens: e, content: t }) {
	return !e || e.length === 0 ? /* @__PURE__ */ (0, N.jsx)(N.Fragment, { children: t }) : /* @__PURE__ */ (0, N.jsx)(N.Fragment, { children: e.map((e, t) => /* @__PURE__ */ (0, N.jsx)("span", {
		style: Ht(e),
		children: e.content
	}, t)) });
}
// ../../../clicky-ui/packages/ui/src/data/CodeDiff.tsx
var Wt = [{
	id: "unified",
	label: "Unified",
	icon: it,
	title: "Unified view"
}, {
	id: "split",
	label: "Split",
	icon: at,
	title: "Split view"
}], Gt = {
	context: void 0,
	add: "bg-emerald-500/10",
	remove: "bg-rose-500/10"
}, Kt = {
	context: "text-muted-foreground",
	add: "text-emerald-700 [[data-theme=dark]_&]:text-emerald-300",
	remove: "text-rose-700 [[data-theme=dark]_&]:text-rose-300"
}, qt = {
	context: "",
	add: "+",
	remove: "-"
};
function Jt(e) {
	let { language: t, view: n = "unified", showLineNumbers: r = !0, bare: a = !1, className: o } = e, { unified: s, original: c, modified: l, originalLabel: u, modifiedLabel: d, onLineAction: f, lineWidgets: p } = e, m = {
		onLineAction: f,
		lineWidgets: p
	}, [h, g] = (0, M.useState)(n), _ = (0, M.useMemo)(() => s === void 0 ? vt(c ?? "", l ?? "") : St(s), [
		s,
		c,
		l
	]), { oldSource: v, newSource: y } = (0, M.useMemo)(() => Yt(_), [_]), [b, x] = (0, M.useState)({
		old: null,
		mod: null
	});
	(0, M.useEffect)(() => {
		if (!t) {
			x({
				old: null,
				mod: null
			});
			return;
		}
		let e = !1;
		return Promise.all([Vt(v, { lang: t }), Vt(y, { lang: t })]).then(([t, n]) => {
			e || x({
				old: t,
				mod: n
			});
		}), () => {
			e = !0;
		};
	}, [
		t,
		v,
		y
	]);
	let S = (0, M.useMemo)(() => Xt(_, b.old, b.mod), [_, b]), C = new Set(S.map((e) => e.path).filter((e) => e !== void 0)).size > 1, w, T = S.map((e, t) => {
		let n = C && e.path !== void 0 && e.path !== w ? e.path : void 0;
		return w = e.path, /* @__PURE__ */ (0, N.jsxs)(M.Fragment, { children: [n && /* @__PURE__ */ (0, N.jsx)(rn, { path: n }), h === "split" ? /* @__PURE__ */ (0, N.jsx)(en, {
			hunk: e,
			showLineNumbers: r,
			annotations: m
		}) : /* @__PURE__ */ (0, N.jsx)($t, {
			hunk: e,
			showLineNumbers: r,
			annotations: m
		})] }, t);
	});
	return a ? /* @__PURE__ */ (0, N.jsx)("div", {
		"data-scroll-x": !0,
		className: i("overflow-x-auto font-mono text-xs leading-relaxed", o),
		children: T
	}) : /* @__PURE__ */ (0, N.jsxs)("div", {
		className: i("overflow-hidden rounded-md border border-border bg-muted/40", o),
		children: [
			/* @__PURE__ */ (0, N.jsxs)("div", {
				className: "flex items-center justify-between gap-2 border-b border-border px-3 py-1.5",
				children: [/* @__PURE__ */ (0, N.jsx)("span", {
					className: "text-[11px] uppercase tracking-wide text-muted-foreground",
					children: t || "diff"
				}), /* @__PURE__ */ (0, N.jsx)(gt, {
					size: "sm",
					"aria-label": "Diff view",
					value: h,
					onChange: g,
					options: Wt
				})]
			}),
			u != null || d != null ? /* @__PURE__ */ (0, N.jsxs)("div", {
				className: i("border-b border-border text-xs font-medium", h === "split" ? "grid grid-cols-2 divide-x divide-border" : "flex flex-wrap gap-x-4"),
				children: [/* @__PURE__ */ (0, N.jsxs)("div", {
					"data-diff-side": "original",
					className: "min-w-0 break-words px-3 py-1.5",
					children: [h === "unified" ? "− " : null, u]
				}), /* @__PURE__ */ (0, N.jsxs)("div", {
					"data-diff-side": "modified",
					className: "min-w-0 break-words px-3 py-1.5",
					children: [h === "unified" ? "+ " : null, d]
				})]
			}) : null,
			/* @__PURE__ */ (0, N.jsx)("div", {
				"data-scroll-x": !0,
				className: "overflow-x-auto font-mono text-xs leading-relaxed",
				children: T
			})
		]
	});
}
function Yt(e) {
	let t = [], n = [];
	for (let r of e) for (let e of r.lines) e.type !== "add" && t.push(e.content), e.type !== "remove" && n.push(e.content);
	return {
		oldSource: t.join("\n"),
		newSource: n.join("\n")
	};
}
function Xt(e, t, n) {
	let r = 0, i = 0;
	return e.map((e) => ({
		header: e.header,
		path: e.path,
		lines: e.lines.map((e) => {
			let a = {
				line: e,
				oldTokens: e.type === "add" ? void 0 : t?.[r],
				newTokens: e.type === "remove" ? void 0 : n?.[i]
			};
			return e.type !== "add" && (r += 1), e.type !== "remove" && (i += 1), a;
		})
	}));
}
function Zt(e, t, n) {
	return t === "old" ? e.type !== "add" && e.oldNumber === n : e.type !== "remove" && e.newNumber === n;
}
function Qt({ widgets: e }) {
	return /* @__PURE__ */ (0, N.jsx)(N.Fragment, { children: e.map((e) => /* @__PURE__ */ (0, N.jsx)("div", {
		"data-diff-widget": e.key,
		className: "whitespace-normal border-y border-border bg-background px-3 py-2 font-sans text-sm leading-normal text-foreground",
		children: e.node
	}, e.key)) });
}
function $t({ hunk: e, showLineNumbers: t, annotations: n }) {
	let { onLineAction: r, lineWidgets: a } = n;
	return /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [e.header && /* @__PURE__ */ (0, N.jsx)(an, { header: e.header }), e.lines.map((e, n) => {
		let { line: o } = e, s = o.type === "remove" ? e.oldTokens : e.newTokens, c = o.type === "remove" ? "old" : "new", l = a?.filter((e) => Zt(o, e.side, e.line));
		return /* @__PURE__ */ (0, N.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, N.jsxs)("div", {
			className: i("group flex", Gt[o.type]),
			"data-diff-line": o.type,
			children: [
				t && /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [/* @__PURE__ */ (0, N.jsx)(on, {
					value: o.oldNumber,
					line: o,
					side: "old",
					actionSide: c,
					onLineAction: r
				}), /* @__PURE__ */ (0, N.jsx)(on, {
					value: o.newNumber,
					line: o,
					side: "new",
					actionSide: c,
					onLineAction: r
				})] }),
				/* @__PURE__ */ (0, N.jsx)(sn, { type: o.type }),
				/* @__PURE__ */ (0, N.jsx)("code", {
					className: "whitespace-pre pl-1 pr-3",
					children: /* @__PURE__ */ (0, N.jsx)(Ut, {
						tokens: s,
						content: o.content
					})
				})
			]
		}), l && /* @__PURE__ */ (0, N.jsx)(Qt, { widgets: l })] }, n);
	})] });
}
function en({ hunk: e, showLineNumbers: t, annotations: n }) {
	let r = (0, M.useMemo)(() => tn(e.lines), [e.lines]), { onLineAction: i, lineWidgets: a } = n;
	return /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [e.header && /* @__PURE__ */ (0, N.jsx)(an, { header: e.header }), r.map((e, n) => {
		let r = a?.filter((t) => e.left !== void 0 && Zt(e.left.line, t.side, t.line) && t.side === "old" || e.right !== void 0 && Zt(e.right.line, t.side, t.line) && t.side === "new");
		return /* @__PURE__ */ (0, N.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, N.jsxs)("div", {
			className: "flex",
			children: [
				/* @__PURE__ */ (0, N.jsx)(nn, {
					resolved: e.left,
					side: "old",
					showLineNumbers: t,
					onLineAction: i
				}),
				/* @__PURE__ */ (0, N.jsx)("span", {
					"aria-hidden": !0,
					className: "w-px shrink-0 bg-border"
				}),
				/* @__PURE__ */ (0, N.jsx)(nn, {
					resolved: e.right,
					side: "new",
					showLineNumbers: t,
					onLineAction: i
				})
			]
		}), r && /* @__PURE__ */ (0, N.jsx)(Qt, { widgets: r })] }, n);
	})] });
}
function tn(e) {
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
function nn({ resolved: e, side: t, showLineNumbers: n, onLineAction: r }) {
	if (!e) return /* @__PURE__ */ (0, N.jsx)("div", { className: "min-w-0 flex-1 basis-0" });
	let { line: a } = e, o = t === "old" && a.type === "remove" || t === "new" && a.type === "add", s = o ? a.type : "context", c = t === "old" ? a.oldNumber : a.newNumber, l = t === "old" ? e.oldTokens : e.newTokens;
	return /* @__PURE__ */ (0, N.jsxs)("div", {
		className: i("group flex min-w-0 flex-1 basis-0", o && Gt[a.type]),
		"data-diff-line": s,
		children: [
			n && /* @__PURE__ */ (0, N.jsx)(on, {
				value: c,
				line: a,
				side: t,
				actionSide: t,
				onLineAction: r
			}),
			/* @__PURE__ */ (0, N.jsx)(sn, { type: s }),
			/* @__PURE__ */ (0, N.jsx)("code", {
				className: "min-w-0 flex-1 whitespace-pre-wrap pl-1 pr-3 [overflow-wrap:anywhere]",
				children: /* @__PURE__ */ (0, N.jsx)(Ut, {
					tokens: l,
					content: a.content
				})
			})
		]
	});
}
function rn({ path: e }) {
	return /* @__PURE__ */ (0, N.jsxs)("div", {
		"data-diff-file": e,
		className: "flex items-center gap-1.5 border-b border-border bg-muted px-2 py-1 font-medium text-foreground",
		children: [/* @__PURE__ */ (0, N.jsx)(st, { className: "size-3.5 shrink-0 text-muted-foreground" }), /* @__PURE__ */ (0, N.jsx)("span", {
			className: "truncate",
			children: e
		})]
	});
}
function an({ header: e }) {
	return /* @__PURE__ */ (0, N.jsx)("div", {
		className: "select-none bg-muted/60 px-2 py-0.5 text-muted-foreground",
		children: e
	});
}
function on({ value: e, line: t, side: n, actionSide: r, onLineAction: i }) {
	return /* @__PURE__ */ (0, N.jsxs)("span", {
		className: "relative w-10 shrink-0 select-none px-2 text-right tabular-nums text-muted-foreground",
		children: [e ?? "", i !== void 0 && e !== void 0 && n === r && /* @__PURE__ */ (0, N.jsx)("button", {
			type: "button",
			"aria-label": `Comment on line ${e}`,
			onClick: () => i({
				side: n,
				line: e,
				content: t.content
			}),
			className: "absolute inset-y-0 left-0.5 flex items-center rounded-sm bg-background text-primary opacity-0 hover:text-primary/80 focus-visible:opacity-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring group-hover:opacity-100",
			children: /* @__PURE__ */ (0, N.jsx)(Je, { className: "size-3.5" })
		})]
	});
}
function sn({ type: e }) {
	return /* @__PURE__ */ (0, N.jsx)("span", {
		className: i("w-4 shrink-0 select-none text-center", Kt[e]),
		children: qt[e]
	});
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/nodes/identity.js
var cn = Symbol.for("yaml.alias"), ln = Symbol.for("yaml.document"), un = Symbol.for("yaml.map"), dn = Symbol.for("yaml.pair"), fn = Symbol.for("yaml.scalar"), pn = Symbol.for("yaml.seq"), mn = Symbol.for("yaml.node.type"), hn = (e) => !!e && typeof e == "object" && e[mn] === cn, gn = (e) => !!e && typeof e == "object" && e[mn] === ln, _n = (e) => !!e && typeof e == "object" && e[mn] === un, vn = (e) => !!e && typeof e == "object" && e[mn] === dn, q = (e) => !!e && typeof e == "object" && e[mn] === fn, yn = (e) => !!e && typeof e == "object" && e[mn] === pn;
function J(e) {
	if (e && typeof e == "object") switch (e[mn]) {
		case un:
		case pn: return !0;
	}
	return !1;
}
function Y(e) {
	if (e && typeof e == "object") switch (e[mn]) {
		case cn:
		case un:
		case fn:
		case pn: return !0;
	}
	return !1;
}
var bn = (e) => (q(e) || J(e)) && !!e.anchor, xn = Symbol("break visit"), Sn = Symbol("skip children"), Cn = Symbol("remove node");
function wn(e, t) {
	let n = En(t);
	gn(e) ? Tn(null, e.contents, n, Object.freeze([e])) === Cn && (e.contents = null) : Tn(null, e, n, Object.freeze([]));
}
wn.BREAK = xn, wn.SKIP = Sn, wn.REMOVE = Cn;
function Tn(e, t, n, r) {
	let i = Dn(e, t, n, r);
	if (Y(i) || vn(i)) return On(e, r, i), Tn(e, i, n, r);
	if (typeof i != "symbol") {
		if (J(t)) {
			r = Object.freeze(r.concat(t));
			for (let e = 0; e < t.items.length; ++e) {
				let i = Tn(e, t.items[e], n, r);
				if (typeof i == "number") e = i - 1;
				else if (i === xn) return xn;
				else i === Cn && (t.items.splice(e, 1), --e);
			}
		} else if (vn(t)) {
			r = Object.freeze(r.concat(t));
			let e = Tn("key", t.key, n, r);
			if (e === xn) return xn;
			e === Cn && (t.key = null);
			let i = Tn("value", t.value, n, r);
			if (i === xn) return xn;
			i === Cn && (t.value = null);
		}
	}
	return i;
}
function En(e) {
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
function Dn(e, t, n, r) {
	if (typeof n == "function") return n(e, t, r);
	if (_n(t)) return n.Map?.(e, t, r);
	if (yn(t)) return n.Seq?.(e, t, r);
	if (vn(t)) return n.Pair?.(e, t, r);
	if (q(t)) return n.Scalar?.(e, t, r);
	if (hn(t)) return n.Alias?.(e, t, r);
}
function On(e, t, n) {
	let r = t[t.length - 1];
	if (J(r)) r.items[e] = n;
	else if (vn(r)) e === "key" ? r.key = n : r.value = n;
	else if (gn(r)) r.contents = n;
	else {
		let e = hn(r) ? "alias" : "scalar";
		throw Error(`Cannot replace node with ${e} parent`);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/doc/directives.js
var kn = {
	"!": "%21",
	",": "%2C",
	"[": "%5B",
	"]": "%5D",
	"{": "%7B",
	"}": "%7D"
}, An = (e) => e.replace(/[!,[\]{}]/g, (e) => kn[e]), jn = class e {
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
		for (let [t, n] of Object.entries(this.tags)) if (e.startsWith(n)) return t + An(e.substring(n.length));
		return e[0] === "!" ? e : `!<${e}>`;
	}
	toString(e) {
		let t = this.yaml.explicit ? [`%YAML ${this.yaml.version || "1.2"}`] : [], n = Object.entries(this.tags), r;
		if (e && n.length > 0 && Y(e.contents)) {
			let t = {};
			wn(e.contents, (e, n) => {
				Y(n) && n.tag && (t[n.tag] = !0);
			}), r = Object.keys(t);
		} else r = [];
		for (let [i, a] of n) i === "!!" && a === "tag:yaml.org,2002:" || e && !r.some((e) => e.startsWith(a)) || t.push(`%TAG ${i} ${a}`);
		return t.join("\n");
	}
};
jn.defaultYaml = {
	explicit: !1,
	version: "1.2"
}, jn.defaultTags = { "!!": "tag:yaml.org,2002:" };
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/doc/anchors.js
function Mn(e) {
	if (/[\x00-\x19\s,[\]{}]/.test(e)) {
		let t = `Anchor must not contain whitespace or control characters: ${JSON.stringify(e)}`;
		throw Error(t);
	}
	return !0;
}
function Nn(e) {
	let t = /* @__PURE__ */ new Set();
	return wn(e, { Value(e, n) {
		n.anchor && t.add(n.anchor);
	} }), t;
}
function Pn(e, t) {
	for (let n = 1;; ++n) {
		let r = `${e}${n}`;
		if (!t.has(r)) return r;
	}
}
function Fn(e, t) {
	let n = [], r = /* @__PURE__ */ new Map(), i = null;
	return {
		onAnchor: (r) => {
			n.push(r), i ??= Nn(e);
			let a = Pn(t, i);
			return i.add(a), a;
		},
		setAnchors: () => {
			for (let e of n) {
				let t = r.get(e);
				if (typeof t == "object" && t.anchor && (q(t.node) || J(t.node))) t.node.anchor = t.anchor;
				else {
					let t = /* @__PURE__ */ Error("Failed to resolve repeated object (this should not happen)");
					throw t.source = e, t;
				}
			}
		},
		sourceObjects: r
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/doc/applyReviver.js
function In(e, t, n, r) {
	if (r && typeof r == "object") {
		if (Array.isArray(r)) for (let t = 0, n = r.length; t < n; ++t) {
			let n = r[t], i = In(e, r, String(t), n);
			i === void 0 ? delete r[t] : i !== n && (r[t] = i);
		}
		else if (r instanceof Map) for (let t of Array.from(r.keys())) {
			let n = r.get(t), i = In(e, r, t, n);
			i === void 0 ? r.delete(t) : i !== n && r.set(t, i);
		}
		else if (r instanceof Set) for (let t of Array.from(r)) {
			let n = In(e, r, t, t);
			n === void 0 ? r.delete(t) : n !== t && (r.delete(t), r.add(n));
		}
		else for (let [t, n] of Object.entries(r)) {
			let i = In(e, r, t, n);
			i === void 0 ? delete r[t] : i !== n && (r[t] = i);
		}
	}
	return e.call(t, n, r);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/nodes/toJS.js
function Ln(e, t, n) {
	if (Array.isArray(e)) return e.map((e, t) => Ln(e, String(t), n));
	if (e && typeof e.toJSON == "function") {
		if (!n || !bn(e)) return e.toJSON(t, n);
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
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/nodes/Node.js
var Rn = class {
	constructor(e) {
		Object.defineProperty(this, mn, { value: e });
	}
	clone() {
		let e = Object.create(Object.getPrototypeOf(this), Object.getOwnPropertyDescriptors(this));
		return this.range && (e.range = this.range.slice()), e;
	}
	toJS(e, { mapAsMap: t, maxAliasCount: n, onAnchor: r, reviver: i } = {}) {
		if (!gn(e)) throw TypeError("A document argument is required");
		let a = {
			anchors: /* @__PURE__ */ new Map(),
			doc: e,
			keep: !0,
			mapAsMap: t === !0,
			mapKeyWarned: !1,
			maxAliasCount: typeof n == "number" ? n : 100
		}, o = Ln(this, "", a);
		if (typeof r == "function") for (let { count: e, res: t } of a.anchors.values()) r(t, e);
		return typeof i == "function" ? In(i, { "": o }, "", o) : o;
	}
}, zn = class extends Rn {
	constructor(e) {
		super(cn), this.source = e, Object.defineProperty(this, "tag", { set() {
			throw Error("Alias nodes cannot have tags");
		} });
	}
	resolve(e, t) {
		let n;
		t?.aliasResolveCache ? n = t.aliasResolveCache : (n = [], wn(e, { Node: (e, t) => {
			(hn(t) || bn(t)) && n.push(t);
		} }), t && (t.aliasResolveCache = n));
		let r;
		for (let e of n) {
			if (e === this) break;
			e.anchor === this.source && (r = e);
		}
		return r;
	}
	toJSON(e, t) {
		if (!t) return { source: this.source };
		let { anchors: n, doc: r, maxAliasCount: i } = t, a = this.resolve(r, t);
		if (!a) {
			let e = `Unresolved alias (the anchor must be set before the alias): ${this.source}`;
			throw ReferenceError(e);
		}
		let o = n.get(a);
		/* istanbul ignore if */
		if (o ||= (Ln(a, null, t), n.get(a)), o?.res === void 0) throw ReferenceError("This should not happen: Alias anchor was not resolved?");
		if (i >= 0 && (o.count += 1, o.aliasCount === 0 && (o.aliasCount = Bn(r, a, n)), o.count * o.aliasCount > i)) throw ReferenceError("Excessive alias count indicates a resource exhaustion attack");
		return o.res;
	}
	toString(e, t, n) {
		let r = `*${this.source}`;
		if (e) {
			if (Mn(this.source), e.options.verifyAliasOrder && !e.anchors.has(this.source)) {
				let e = `Unresolved alias (the anchor must be set before the alias): ${this.source}`;
				throw Error(e);
			}
			if (e.implicitKey) return `${r} `;
		}
		return r;
	}
};
function Bn(e, t, n) {
	if (hn(t)) {
		let r = t.resolve(e), i = n && r && n.get(r);
		return i ? i.count * i.aliasCount : 0;
	}
	if (J(t)) {
		let r = 0;
		for (let i of t.items) {
			let t = Bn(e, i, n);
			t > r && (r = t);
		}
		return r;
	}
	if (vn(t)) {
		let r = Bn(e, t.key, n), i = Bn(e, t.value, n);
		return Math.max(r, i);
	}
	return 1;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/nodes/Scalar.js
var Vn = (e) => !e || typeof e != "function" && typeof e != "object", X = class extends Rn {
	constructor(e) {
		super(fn), this.value = e;
	}
	toJSON(e, t) {
		return t?.keep ? this.value : Ln(this.value, e, t);
	}
	toString() {
		return String(this.value);
	}
};
X.BLOCK_FOLDED = "BLOCK_FOLDED", X.BLOCK_LITERAL = "BLOCK_LITERAL", X.PLAIN = "PLAIN", X.QUOTE_DOUBLE = "QUOTE_DOUBLE", X.QUOTE_SINGLE = "QUOTE_SINGLE";
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/doc/createNode.js
var Hn = "tag:yaml.org,2002:";
function Un(e, t, n) {
	if (t) {
		let e = n.filter((e) => e.tag === t), r = e.find((e) => !e.format) ?? e[0];
		if (!r) throw Error(`Tag ${t} not found`);
		return r;
	}
	return n.find((t) => t.identify?.(e) && !t.format);
}
function Wn(e, t, n) {
	if (gn(e) && (e = e.contents), Y(e)) return e;
	if (vn(e)) {
		let t = n.schema[un].createNode?.(n.schema, null, n);
		return t.items.push(e), t;
	}
	(e instanceof String || e instanceof Number || e instanceof Boolean || typeof BigInt < "u" && e instanceof BigInt) && (e = e.valueOf());
	let { aliasDuplicateObjects: r, onAnchor: i, onTagObj: a, schema: o, sourceObjects: s } = n, c;
	if (r && e && typeof e == "object") {
		if (c = s.get(e), c) return c.anchor ?? (c.anchor = i(e)), new zn(c.anchor);
		c = {
			anchor: null,
			node: null
		}, s.set(e, c);
	}
	t?.startsWith("!!") && (t = Hn + t.slice(2));
	let l = Un(e, t, o.tags);
	if (!l) {
		if (e && typeof e.toJSON == "function" && (e = e.toJSON()), !e || typeof e != "object") {
			let t = new X(e);
			return c && (c.node = t), t;
		}
		l = e instanceof Map ? o[un] : Symbol.iterator in Object(e) ? o[pn] : o[un];
	}
	a && (a(l), delete n.onTagObj);
	let u = l?.createNode ? l.createNode(n.schema, e, n) : typeof l?.nodeClass?.from == "function" ? l.nodeClass.from(n.schema, e, n) : new X(e);
	return t ? u.tag = t : l.default || (u.tag = l.tag), c && (c.node = u), u;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/nodes/Collection.js
function Gn(e, t, n) {
	let r = n;
	for (let e = t.length - 1; e >= 0; --e) {
		let n = t[e];
		if (typeof n == "number" && Number.isInteger(n) && n >= 0) {
			let e = [];
			e[n] = r, r = e;
		} else r = /* @__PURE__ */ new Map([[n, r]]);
	}
	return Wn(r, void 0, {
		aliasDuplicateObjects: !1,
		keepUndefined: !1,
		onAnchor: () => {
			throw Error("This should not happen, please report a bug.");
		},
		schema: e,
		sourceObjects: /* @__PURE__ */ new Map()
	});
}
var Kn = (e) => e == null || typeof e == "object" && !!e[Symbol.iterator]().next().done, qn = class extends Rn {
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
		return e && (t.schema = e), t.items = t.items.map((t) => Y(t) || vn(t) ? t.clone(e) : t), this.range && (t.range = this.range.slice()), t;
	}
	addIn(e, t) {
		if (Kn(e)) this.add(t);
		else {
			let [n, ...r] = e, i = this.get(n, !0);
			if (J(i)) i.addIn(r, t);
			else if (i === void 0 && this.schema) this.set(n, Gn(this.schema, r, t));
			else throw Error(`Expected YAML collection at ${n}. Remaining path: ${r}`);
		}
	}
	deleteIn(e) {
		let [t, ...n] = e;
		if (n.length === 0) return this.delete(t);
		let r = this.get(t, !0);
		if (J(r)) return r.deleteIn(n);
		throw Error(`Expected YAML collection at ${t}. Remaining path: ${n}`);
	}
	getIn(e, t) {
		let [n, ...r] = e, i = this.get(n, !0);
		return r.length === 0 ? !t && q(i) ? i.value : i : J(i) ? i.getIn(r, t) : void 0;
	}
	hasAllNullValues(e) {
		return this.items.every((t) => {
			if (!vn(t)) return !1;
			let n = t.value;
			return n == null || e && q(n) && n.value == null && !n.commentBefore && !n.comment && !n.tag;
		});
	}
	hasIn(e) {
		let [t, ...n] = e;
		if (n.length === 0) return this.has(t);
		let r = this.get(t, !0);
		return J(r) ? r.hasIn(n) : !1;
	}
	setIn(e, t) {
		let [n, ...r] = e;
		if (r.length === 0) this.set(n, t);
		else {
			let e = this.get(n, !0);
			if (J(e)) e.setIn(r, t);
			else if (e === void 0 && this.schema) this.set(n, Gn(this.schema, r, t));
			else throw Error(`Expected YAML collection at ${n}. Remaining path: ${r}`);
		}
	}
}, Jn = (e) => e.replace(/^(?!$)(?: $)?/gm, "#");
function Yn(e, t) {
	return /^\n+$/.test(e) ? e.substring(1) : t ? e.replace(/^(?! *$)/gm, t) : e;
}
var Xn = (e, t, n) => e.endsWith("\n") ? Yn(n, t) : n.includes("\n") ? "\n" + Yn(n, t) : (e.endsWith(" ") ? "" : " ") + n, Zn = "flow", Qn = "block", $n = "quoted";
function er(e, t, n = "flow", { indentAtStart: r, lineWidth: i = 80, minContentWidth: a = 20, onFold: o, onOverflow: s } = {}) {
	if (!i || i < 0) return e;
	i < a && (a = 0);
	let c = Math.max(1 + a, 1 + i - t.length);
	if (e.length <= c) return e;
	let l = [], u = {}, d = i - t.length;
	typeof r == "number" && (r > i - Math.max(2, a) ? l.push(0) : d = i - r);
	let f, p, m = !1, h = -1, g = -1, _ = -1;
	n === "block" && (h = tr(e, h, t.length), h !== -1 && (d = h + c));
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
		if (r === "\n") n === "block" && (h = tr(e, h, t.length)), d = h + t.length + c, f = void 0;
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
function tr(e, t, n) {
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
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/stringify/stringifyString.js
var nr = (e, t) => ({
	indentAtStart: t ? e.indent.length : e.indentAtStart,
	lineWidth: e.options.lineWidth,
	minContentWidth: e.options.minContentWidth
}), rr = (e) => /^(%|---|\.\.\.)/m.test(e);
function ir(e, t, n) {
	if (!t || t < 0) return !1;
	let r = t - n, i = e.length;
	if (i <= r) return !1;
	for (let t = 0, n = 0; t < i; ++t) if (e[t] === "\n") {
		if (t - n > r) return !0;
		if (n = t + 1, i - n <= r) return !1;
	}
	return !0;
}
function ar(e, t) {
	let n = JSON.stringify(e);
	if (t.options.doubleQuotedAsJSON) return n;
	let { implicitKey: r } = t, i = t.options.doubleQuotedMinMultiLineLength, a = t.indent || (rr(e) ? "  " : ""), o = "", s = 0;
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
	return o = s ? o + n.slice(s) : n, r ? o : er(o, a, $n, nr(t, !1));
}
function or(e, t) {
	if (t.options.singleQuote === !1 || t.implicitKey && e.includes("\n") || /[ \t]\n|\n[ \t]/.test(e)) return ar(e, t);
	let n = t.indent || (rr(e) ? "  " : ""), r = "'" + e.replace(/'/g, "''").replace(/\n+/g, `$&\n${n}`) + "'";
	return t.implicitKey ? r : er(r, n, Zn, nr(t, !1));
}
function sr(e, t) {
	let { singleQuote: n } = t.options, r;
	if (n === !1) r = ar;
	else {
		let t = e.includes("\""), i = e.includes("'");
		r = t && !i ? or : i && !t ? ar : n ? or : ar;
	}
	return r(e, t);
}
var cr;
try {
	cr = /* @__PURE__ */ RegExp("(^|(?<!\n))\n+(?!\n|$)", "g");
} catch {
	cr = /\n+(?!\n|$)/g;
}
function lr({ comment: e, type: t, value: n }, r, i, a) {
	let { blockQuote: o, commentString: s, lineWidth: c } = r.options;
	if (!o || /\n[\t ]+$/.test(n)) return sr(n, r);
	let l = r.indent || (r.forceBlockIndent || rr(n) ? "  " : ""), u = o === "literal" ? !0 : o === "folded" || t === X.BLOCK_FOLDED ? !1 : t === X.BLOCK_LITERAL || !ir(n, c, l.length);
	if (!n) return u ? "|\n" : ">\n";
	let d, f;
	for (f = n.length; f > 0; --f) {
		let e = n[f - 1];
		if (e !== "\n" && e !== "	" && e !== " ") break;
	}
	let p = n.substring(f), m = p.indexOf("\n");
	m === -1 ? d = "-" : n === p || m !== p.length - 1 ? (d = "+", a && a()) : d = "", p &&= (n = n.slice(0, -p.length), p[p.length - 1] === "\n" && (p = p.slice(0, -1)), p.replace(cr, `$&${l}`));
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
		let e = n.replace(/\n+/g, "\n$&").replace(/(?:^|\n)([\t ].*)(?:([\n\t ]*)\n(?![\n\t ]))?/g, "$1$2").replace(/\n+/g, `$&${l}`), i = !1, a = nr(r, !0);
		o !== "folded" && t !== X.BLOCK_FOLDED && (a.onOverflow = () => {
			i = !0;
		});
		let s = er(`${v}${e}${p}`, l, Qn, a);
		if (!i) return `>${y}\n${l}${s}`;
	}
	return n = n.replace(/\n+/g, `$&${l}`), `|${y}\n${l}${v}${n}${p}`;
}
function ur(e, t, n, r) {
	let { type: i, value: a } = e, { actualString: o, implicitKey: s, indent: c, indentStep: l, inFlow: u } = t;
	if (s && a.includes("\n") || u && /[[\]{},]/.test(a)) return sr(a, t);
	if (/^[\n\t ,[\]{}#&*!|>'"%@`]|^[?-]$|^[?-][ \t]|[\n:][ \t]|[ \t]\n|[\n\t ]#|[\n\t :]$/.test(a)) return s || u || !a.includes("\n") ? sr(a, t) : lr(e, t, n, r);
	if (!s && !u && i !== X.PLAIN && a.includes("\n")) return lr(e, t, n, r);
	if (rr(a)) {
		if (c === "") return t.forceBlockIndent = !0, lr(e, t, n, r);
		if (s && c === l) return sr(a, t);
	}
	let d = a.replace(/\n+/g, `$&\n${c}`);
	if (o) {
		let e = (e) => e.default && e.tag !== "tag:yaml.org,2002:str" && e.test?.test(d), { compat: n, tags: r } = t.doc.schema;
		if (r.some(e) || n?.some(e)) return sr(a, t);
	}
	return s ? d : er(d, c, Zn, nr(t, !1));
}
function dr(e, t, n, r) {
	let { implicitKey: i, inFlow: a } = t, o = typeof e.value == "string" ? e : Object.assign({}, e, { value: String(e.value) }), { type: s } = e;
	s !== X.QUOTE_DOUBLE && /[\x00-\x08\x0b-\x1f\x7f-\x9f\u{D800}-\u{DFFF}]/u.test(o.value) && (s = X.QUOTE_DOUBLE);
	let c = (e) => {
		switch (e) {
			case X.BLOCK_FOLDED:
			case X.BLOCK_LITERAL: return i || a ? sr(o.value, t) : lr(o, t, n, r);
			case X.QUOTE_DOUBLE: return ar(o.value, t);
			case X.QUOTE_SINGLE: return or(o.value, t);
			case X.PLAIN: return ur(o, t, n, r);
			default: return null;
		}
	}, l = c(s);
	if (l === null) {
		let { defaultKeyType: e, defaultStringType: n } = t.options, r = i && e || n;
		if (l = c(r), l === null) throw Error(`Unsupported default string type ${r}`);
	}
	return l;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/stringify/stringify.js
function fr(e, t) {
	let n = Object.assign({
		blockQuote: !0,
		commentString: Jn,
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
function pr(e, t) {
	if (t.tag) {
		let n = e.filter((e) => e.tag === t.tag);
		if (n.length > 0) return n.find((e) => e.format === t.format) ?? n[0];
	}
	let n, r;
	if (q(t)) {
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
function mr(e, t, { anchors: n, doc: r }) {
	if (!r.directives) return "";
	let i = [], a = (q(e) || J(e)) && e.anchor;
	a && Mn(a) && (n.add(a), i.push(`&${a}`));
	let o = e.tag ?? (t.default ? null : t.tag);
	return o && i.push(r.directives.tagString(o)), i.join(" ");
}
function hr(e, t, n, r) {
	if (vn(e)) return e.toString(t, n, r);
	if (hn(e)) {
		if (t.doc.directives) return e.toString(t);
		if (t.resolvedAliases?.has(e)) throw TypeError("Cannot stringify circular structure without alias nodes");
		t.resolvedAliases ? t.resolvedAliases.add(e) : t.resolvedAliases = /* @__PURE__ */ new Set([e]), e = e.resolve(t.doc);
	}
	let i, a = Y(e) ? e : t.doc.createNode(e, { onTagObj: (e) => i = e });
	i ??= pr(t.doc.schema.tags, a);
	let o = mr(a, i, t);
	o.length > 0 && (t.indentAtStart = (t.indentAtStart ?? 0) + o.length + 1);
	let s = typeof i.stringify == "function" ? i.stringify(a, t, n, r) : q(a) ? dr(a, t, n, r) : a.toString(t, n, r);
	return o ? q(a) || s[0] === "{" || s[0] === "[" ? `${o} ${s}` : `${o}\n${t.indent}${s}` : s;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/stringify/stringifyPair.js
function gr({ key: e, value: t }, n, r, i) {
	let { allNullValues: a, doc: o, indent: s, indentStep: c, options: { commentString: l, indentSeq: u, simpleKeys: d } } = n, f = Y(e) && e.comment || null;
	if (d) {
		if (f) throw Error("With simple keys, key nodes cannot have comments");
		if (J(e) || !Y(e) && typeof e == "object") throw Error("With simple keys, collection cannot be used as a key value");
	}
	let p = !d && (!e || f && t == null && !n.inFlow || J(e) || (q(e) ? e.type === X.BLOCK_FOLDED || e.type === X.BLOCK_LITERAL : typeof e == "object"));
	n = Object.assign({}, n, {
		allNullValues: !1,
		implicitKey: !p && (d || !a),
		indent: s + c
	});
	let m = !1, h = !1, g = hr(e, n, () => m = !0, () => h = !0);
	if (!p && !n.inFlow && g.length > 1024) {
		if (d) throw Error("With simple keys, single line scalar must not span more than 1024 characters");
		p = !0;
	}
	if (n.inFlow) {
		if (a || t == null) return m && r && r(), g === "" ? "?" : p ? `? ${g}` : g;
	} else if (a && !d || t == null && p) return g = `? ${g}`, f && !m ? g += Xn(g, n.indent, l(f)) : h && i && i(), g;
	m && (f = null), p ? (f && (g += Xn(g, n.indent, l(f))), g = `? ${g}\n${s}:`) : (g = `${g}:`, f && (g += Xn(g, n.indent, l(f))));
	let _, v, y;
	Y(t) ? (_ = !!t.spaceBefore, v = t.commentBefore, y = t.comment) : (_ = !1, v = null, y = null, t && typeof t == "object" && (t = o.createNode(t))), n.implicitKey = !1, !p && !f && q(t) && (n.indentAtStart = g.length + 1), h = !1, !u && c.length >= 2 && !n.inFlow && !p && yn(t) && !t.flow && !t.tag && !t.anchor && (n.indent = n.indent.substring(2));
	let b = !1, x = hr(t, n, () => b = !0, () => h = !0), S = " ";
	if (f || _ || v) {
		if (S = _ ? "\n" : "", v) {
			let e = l(v);
			S += `\n${Yn(e, n.indent)}`;
		}
		x === "" && !n.inFlow ? S === "\n" && y && (S = "\n\n") : S += `\n${n.indent}`;
	} else if (!p && J(t)) {
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
	return g += S + x, n.inFlow ? b && r && r() : y && !b ? g += Xn(g, n.indent, l(y)) : h && i && i(), g;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/log.js
function _r(e, t) {
	(e === "debug" || e === "warn") && console.warn(t);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/schema/yaml-1.1/merge.js
var vr = "<<", yr = {
	identify: (e) => e === vr || typeof e == "symbol" && e.description === vr,
	default: "key",
	tag: "tag:yaml.org,2002:merge",
	test: /^<<$/,
	resolve: () => Object.assign(new X(Symbol(vr)), { addToJSMap: xr }),
	stringify: () => vr
}, br = (e, t) => (yr.identify(t) || q(t) && (!t.type || t.type === X.PLAIN) && yr.identify(t.value)) && e?.doc.schema.tags.some((e) => e.tag === yr.tag && e.default);
function xr(e, t, n) {
	if (n = e && hn(n) ? n.resolve(e.doc) : n, yn(n)) for (let r of n.items) Sr(e, t, r);
	else if (Array.isArray(n)) for (let r of n) Sr(e, t, r);
	else Sr(e, t, n);
}
function Sr(e, t, n) {
	let r = e && hn(n) ? n.resolve(e.doc) : n;
	if (!_n(r)) throw Error("Merge sources must be maps or map aliases");
	let i = r.toJSON(null, e, Map);
	for (let [e, n] of i) t instanceof Map ? t.has(e) || t.set(e, n) : t instanceof Set ? t.add(e) : Object.prototype.hasOwnProperty.call(t, e) || Object.defineProperty(t, e, {
		value: n,
		writable: !0,
		enumerable: !0,
		configurable: !0
	});
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/nodes/addPairToJSMap.js
function Cr(e, t, { key: n, value: r }) {
	if (Y(n) && n.addToJSMap) n.addToJSMap(e, t, r);
	else if (br(e, n)) xr(e, t, r);
	else {
		let i = Ln(n, "", e);
		if (t instanceof Map) t.set(i, Ln(r, i, e));
		else if (t instanceof Set) t.add(i);
		else {
			let a = wr(n, i, e), o = Ln(r, a, e);
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
function wr(e, t, n) {
	if (t === null) return "";
	if (typeof t != "object") return String(t);
	if (Y(e) && n?.doc) {
		let t = fr(n.doc, {});
		t.anchors = /* @__PURE__ */ new Set();
		for (let e of n.anchors.keys()) t.anchors.add(e.anchor);
		t.inFlow = !0, t.inStringifyKey = !0;
		let r = e.toString(t);
		if (!n.mapKeyWarned) {
			let e = JSON.stringify(r);
			e.length > 40 && (e = e.substring(0, 36) + "...\""), _r(n.doc.options.logLevel, `Keys with collection values will be stringified due to JS Object restrictions: ${e}. Set mapAsMap: true to use object keys.`), n.mapKeyWarned = !0;
		}
		return r;
	}
	return JSON.stringify(t);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/nodes/Pair.js
function Tr(e, t, n) {
	return new Er(Wn(e, void 0, n), Wn(t, void 0, n));
}
var Er = class e {
	constructor(e, t = null) {
		Object.defineProperty(this, mn, { value: dn }), this.key = e, this.value = t;
	}
	clone(t) {
		let { key: n, value: r } = this;
		return Y(n) && (n = n.clone(t)), Y(r) && (r = r.clone(t)), new e(n, r);
	}
	toJSON(e, t) {
		return Cr(t, t?.mapAsMap ? /* @__PURE__ */ new Map() : {}, this);
	}
	toString(e, t, n) {
		return e?.doc ? gr(this, e, t, n) : JSON.stringify(this);
	}
};
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/stringify/stringifyCollection.js
function Dr(e, t, n) {
	return (t.inFlow ?? e.flow ? kr : Or)(e, t, n);
}
function Or({ comment: e, items: t }, n, { blockItemPrefix: r, flowChars: i, itemIndent: a, onChompKeep: o, onComment: s }) {
	let { indent: c, options: { commentString: l } } = n, u = Object.assign({}, n, {
		indent: a,
		type: null
	}), d = !1, f = [];
	for (let e = 0; e < t.length; ++e) {
		let i = t[e], o = null;
		if (Y(i)) !d && i.spaceBefore && f.push(""), Ar(n, f, i.commentBefore, d), i.comment && (o = i.comment);
		else if (vn(i)) {
			let e = Y(i.key) ? i.key : null;
			e && (!d && e.spaceBefore && f.push(""), Ar(n, f, e.commentBefore, d));
		}
		d = !1;
		let s = hr(i, u, () => o = null, () => d = !0);
		o && (s += Xn(s, a, l(o))), d && o && (d = !1), f.push(r + s);
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
	return e ? (p += "\n" + Yn(l(e), c), s && s()) : d && o && o(), p;
}
function kr({ items: e }, t, { flowChars: n, itemIndent: r }) {
	let { indent: i, indentStep: a, flowCollectionPadding: o, options: { commentString: s } } = t;
	r += a;
	let c = Object.assign({}, t, {
		indent: r,
		inFlow: !0,
		type: null
	}), l = !1, u = 0, d = [];
	for (let n = 0; n < e.length; ++n) {
		let i = e[n], a = null;
		if (Y(i)) i.spaceBefore && d.push(""), Ar(t, d, i.commentBefore, !1), i.comment && (a = i.comment);
		else if (vn(i)) {
			let e = Y(i.key) ? i.key : null;
			e && (e.spaceBefore && d.push(""), Ar(t, d, e.commentBefore, !1), e.comment && (l = !0));
			let n = Y(i.value) ? i.value : null;
			n ? (n.comment && (a = n.comment), n.commentBefore && (l = !0)) : i.value == null && e?.comment && (a = e.comment);
		}
		a && (l = !0);
		let o = hr(i, c, () => a = null);
		l ||= d.length > u || o.includes("\n"), n < e.length - 1 ? o += "," : t.options.trailingComma && (t.options.lineWidth > 0 && (l ||= d.reduce((e, t) => e + t.length + 2, 2) + (o.length + 2) > t.options.lineWidth), l && (o += ",")), a && (o += Xn(o, r, s(a))), d.push(o), u = d.length;
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
function Ar({ indent: e, options: { commentString: t } }, n, r, i) {
	if (r && i && (r = r.replace(/^\n+/, "")), r) {
		let i = Yn(t(r), e);
		n.push(i.trimStart());
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/nodes/YAMLMap.js
function jr(e, t) {
	let n = q(t) ? t.value : t;
	for (let r of e) if (vn(r) && (r.key === t || r.key === n || q(r.key) && r.key.value === n)) return r;
}
var Mr = class extends qn {
	static get tagName() {
		return "tag:yaml.org,2002:map";
	}
	constructor(e) {
		super(un, e), this.items = [];
	}
	static from(e, t, n) {
		let { keepUndefined: r, replacer: i } = n, a = new this(e), o = (e, o) => {
			if (typeof i == "function") o = i.call(t, e, o);
			else if (Array.isArray(i) && !i.includes(e)) return;
			(o !== void 0 || r) && a.items.push(Tr(e, o, n));
		};
		if (t instanceof Map) for (let [e, n] of t) o(e, n);
		else if (t && typeof t == "object") for (let e of Object.keys(t)) o(e, t[e]);
		return typeof e.sortMapEntries == "function" && a.items.sort(e.sortMapEntries), a;
	}
	add(e, t) {
		let n;
		n = vn(e) ? e : !e || typeof e != "object" || !("key" in e) ? new Er(e, e?.value) : new Er(e.key, e.value);
		let r = jr(this.items, n.key), i = this.schema?.sortMapEntries;
		if (r) {
			if (!t) throw Error(`Key ${n.key} already set`);
			q(r.value) && Vn(n.value) ? r.value.value = n.value : r.value = n.value;
		} else if (i) {
			let e = this.items.findIndex((e) => i(n, e) < 0);
			e === -1 ? this.items.push(n) : this.items.splice(e, 0, n);
		} else this.items.push(n);
	}
	delete(e) {
		let t = jr(this.items, e);
		return t ? this.items.splice(this.items.indexOf(t), 1).length > 0 : !1;
	}
	get(e, t) {
		let n = jr(this.items, e)?.value;
		return (!t && q(n) ? n.value : n) ?? void 0;
	}
	has(e) {
		return !!jr(this.items, e);
	}
	set(e, t) {
		this.add(new Er(e, t), !0);
	}
	toJSON(e, t, n) {
		let r = n ? new n() : t?.mapAsMap ? /* @__PURE__ */ new Map() : {};
		t?.onCreate && t.onCreate(r);
		for (let e of this.items) Cr(t, r, e);
		return r;
	}
	toString(e, t, n) {
		if (!e) return JSON.stringify(this);
		for (let e of this.items) if (!vn(e)) throw Error(`Map items must all be pairs; found ${JSON.stringify(e)} instead`);
		return !e.allNullValues && this.hasAllNullValues(!1) && (e = Object.assign({}, e, { allNullValues: !0 })), Dr(this, e, {
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
}, Nr = {
	collection: "map",
	default: !0,
	nodeClass: Mr,
	tag: "tag:yaml.org,2002:map",
	resolve(e, t) {
		return _n(e) || t("Expected a mapping for this tag"), e;
	},
	createNode: (e, t, n) => Mr.from(e, t, n)
}, Pr = class extends qn {
	static get tagName() {
		return "tag:yaml.org,2002:seq";
	}
	constructor(e) {
		super(pn, e), this.items = [];
	}
	add(e) {
		this.items.push(e);
	}
	delete(e) {
		let t = Fr(e);
		return typeof t == "number" && this.items.splice(t, 1).length > 0;
	}
	get(e, t) {
		let n = Fr(e);
		if (typeof n != "number") return;
		let r = this.items[n];
		return !t && q(r) ? r.value : r;
	}
	has(e) {
		let t = Fr(e);
		return typeof t == "number" && t < this.items.length;
	}
	set(e, t) {
		let n = Fr(e);
		if (typeof n != "number") throw Error(`Expected a valid index, not ${e}.`);
		let r = this.items[n];
		q(r) && Vn(t) ? r.value = t : this.items[n] = t;
	}
	toJSON(e, t) {
		let n = [];
		t?.onCreate && t.onCreate(n);
		let r = 0;
		for (let e of this.items) n.push(Ln(e, String(r++), t));
		return n;
	}
	toString(e, t, n) {
		return e ? Dr(this, e, {
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
				i.items.push(Wn(a, void 0, n));
			}
		}
		return i;
	}
};
function Fr(e) {
	let t = q(e) ? e.value : e;
	return t && typeof t == "string" && (t = Number(t)), typeof t == "number" && Number.isInteger(t) && t >= 0 ? t : null;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/schema/common/seq.js
var Ir = {
	collection: "seq",
	default: !0,
	nodeClass: Pr,
	tag: "tag:yaml.org,2002:seq",
	resolve(e, t) {
		return yn(e) || t("Expected a sequence for this tag"), e;
	},
	createNode: (e, t, n) => Pr.from(e, t, n)
}, Lr = {
	identify: (e) => typeof e == "string",
	default: !0,
	tag: "tag:yaml.org,2002:str",
	resolve: (e) => e,
	stringify(e, t, n, r) {
		return t = Object.assign({ actualString: !0 }, t), dr(e, t, n, r);
	}
}, Rr = {
	identify: (e) => e == null,
	createNode: () => new X(null),
	default: !0,
	tag: "tag:yaml.org,2002:null",
	test: /^(?:~|[Nn]ull|NULL)?$/,
	resolve: () => new X(null),
	stringify: ({ source: e }, t) => typeof e == "string" && Rr.test.test(e) ? e : t.options.nullStr
}, zr = {
	identify: (e) => typeof e == "boolean",
	default: !0,
	tag: "tag:yaml.org,2002:bool",
	test: /^(?:[Tt]rue|TRUE|[Ff]alse|FALSE)$/,
	resolve: (e) => new X(e[0] === "t" || e[0] === "T"),
	stringify({ source: e, value: t }, n) {
		return e && zr.test.test(e) && t === (e[0] === "t" || e[0] === "T") ? e : t ? n.options.trueStr : n.options.falseStr;
	}
};
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/stringify/stringifyNumber.js
function Br({ format: e, minFractionDigits: t, tag: n, value: r }) {
	if (typeof r == "bigint") return String(r);
	let i = typeof r == "number" ? r : Number(r);
	if (!isFinite(i)) return isNaN(i) ? ".nan" : i < 0 ? "-.inf" : ".inf";
	let a = Object.is(r, -0) ? "-0" : JSON.stringify(r);
	if (!e && t && (!n || n === "tag:yaml.org,2002:float") && /^\d/.test(a)) {
		let e = a.indexOf(".");
		e < 0 && (e = a.length, a += ".");
		let n = t - (a.length - e - 1);
		for (; n-- > 0;) a += "0";
	}
	return a;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/schema/core/float.js
var Vr = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	test: /^(?:[-+]?\.(?:inf|Inf|INF)|\.nan|\.NaN|\.NAN)$/,
	resolve: (e) => e.slice(-3).toLowerCase() === "nan" ? NaN : e[0] === "-" ? -Infinity : Infinity,
	stringify: Br
}, Hr = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	format: "EXP",
	test: /^[-+]?(?:\.[0-9]+|[0-9]+(?:\.[0-9]*)?)[eE][-+]?[0-9]+$/,
	resolve: (e) => parseFloat(e),
	stringify(e) {
		let t = Number(e.value);
		return isFinite(t) ? t.toExponential() : Br(e);
	}
}, Ur = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	test: /^[-+]?(?:\.[0-9]+|[0-9]+\.[0-9]*)$/,
	resolve(e) {
		let t = new X(parseFloat(e)), n = e.indexOf(".");
		return n !== -1 && e[e.length - 1] === "0" && (t.minFractionDigits = e.length - n - 1), t;
	},
	stringify: Br
}, Wr = (e) => typeof e == "bigint" || Number.isInteger(e), Gr = (e, t, n, { intAsBigInt: r }) => r ? BigInt(e) : parseInt(e.substring(t), n);
function Kr(e, t, n) {
	let { value: r } = e;
	return Wr(r) && r >= 0 ? n + r.toString(t) : Br(e);
}
var qr = {
	identify: (e) => Wr(e) && e >= 0,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "OCT",
	test: /^0o[0-7]+$/,
	resolve: (e, t, n) => Gr(e, 2, 8, n),
	stringify: (e) => Kr(e, 8, "0o")
}, Jr = {
	identify: Wr,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	test: /^[-+]?[0-9]+$/,
	resolve: (e, t, n) => Gr(e, 0, 10, n),
	stringify: Br
}, Yr = {
	identify: (e) => Wr(e) && e >= 0,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "HEX",
	test: /^0x[0-9a-fA-F]+$/,
	resolve: (e, t, n) => Gr(e, 2, 16, n),
	stringify: (e) => Kr(e, 16, "0x")
}, Xr = [
	Nr,
	Ir,
	Lr,
	Rr,
	zr,
	qr,
	Jr,
	Yr,
	Vr,
	Hr,
	Ur
];
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/schema/json/schema.js
function Zr(e) {
	return typeof e == "bigint" || Number.isInteger(e);
}
var Qr = ({ value: e }) => JSON.stringify(e), $r = [
	{
		identify: (e) => typeof e == "string",
		default: !0,
		tag: "tag:yaml.org,2002:str",
		resolve: (e) => e,
		stringify: Qr
	},
	{
		identify: (e) => e == null,
		createNode: () => new X(null),
		default: !0,
		tag: "tag:yaml.org,2002:null",
		test: /^null$/,
		resolve: () => null,
		stringify: Qr
	},
	{
		identify: (e) => typeof e == "boolean",
		default: !0,
		tag: "tag:yaml.org,2002:bool",
		test: /^true$|^false$/,
		resolve: (e) => e === "true",
		stringify: Qr
	},
	{
		identify: Zr,
		default: !0,
		tag: "tag:yaml.org,2002:int",
		test: /^-?(?:0|[1-9][0-9]*)$/,
		resolve: (e, t, { intAsBigInt: n }) => n ? BigInt(e) : parseInt(e, 10),
		stringify: ({ value: e }) => Zr(e) ? e.toString() : JSON.stringify(e)
	},
	{
		identify: (e) => typeof e == "number",
		default: !0,
		tag: "tag:yaml.org,2002:float",
		test: /^-?(?:0|[1-9][0-9]*)(?:\.[0-9]*)?(?:[eE][-+]?[0-9]+)?$/,
		resolve: (e) => parseFloat(e),
		stringify: Qr
	}
], ei = [Nr, Ir].concat($r, {
	default: !0,
	tag: "",
	test: /^/,
	resolve(e, t) {
		return t(`Unresolved plain scalar ${JSON.stringify(e)}`), e;
	}
}), ti = {
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
		if (t ??= X.BLOCK_LITERAL, t !== X.QUOTE_DOUBLE) {
			let e = Math.max(r.options.lineWidth - r.indent.length, r.options.minContentWidth), n = Math.ceil(s.length / e), i = Array(n);
			for (let t = 0, r = 0; t < n; ++t, r += e) i[t] = s.substr(r, e);
			s = i.join(t === X.BLOCK_LITERAL ? "\n" : " ");
		}
		return dr({
			comment: e,
			type: t,
			value: s
		}, r, i, a);
	}
};
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/schema/yaml-1.1/pairs.js
function ni(e, t) {
	if (yn(e)) for (let n = 0; n < e.items.length; ++n) {
		let r = e.items[n];
		if (!vn(r)) {
			if (_n(r)) {
				r.items.length > 1 && t("Each pair must have its own sequence indicator");
				let e = r.items[0] || new Er(new X(null));
				if (r.commentBefore && (e.key.commentBefore = e.key.commentBefore ? `${r.commentBefore}\n${e.key.commentBefore}` : r.commentBefore), r.comment) {
					let t = e.value ?? e.key;
					t.comment = t.comment ? `${r.comment}\n${t.comment}` : r.comment;
				}
				r = e;
			}
			e.items[n] = vn(r) ? r : new Er(r);
		}
	}
	else t("Expected a sequence for this tag");
	return e;
}
function ri(e, t, n) {
	let { replacer: r } = n, i = new Pr(e);
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
		i.items.push(Tr(o, s, n));
	}
	return i;
}
var ii = {
	collection: "seq",
	default: !1,
	tag: "tag:yaml.org,2002:pairs",
	resolve: ni,
	createNode: ri
}, ai = class e extends Pr {
	constructor() {
		super(), this.add = Mr.prototype.add.bind(this), this.delete = Mr.prototype.delete.bind(this), this.get = Mr.prototype.get.bind(this), this.has = Mr.prototype.has.bind(this), this.set = Mr.prototype.set.bind(this), this.tag = e.tag;
	}
	toJSON(e, t) {
		if (!t) return super.toJSON(e);
		let n = /* @__PURE__ */ new Map();
		t?.onCreate && t.onCreate(n);
		for (let e of this.items) {
			let r, i;
			if (vn(e) ? (r = Ln(e.key, "", t), i = Ln(e.value, r, t)) : r = Ln(e, "", t), n.has(r)) throw Error("Ordered maps must not include duplicate keys");
			n.set(r, i);
		}
		return n;
	}
	static from(e, t, n) {
		let r = ri(e, t, n), i = new this();
		return i.items = r.items, i;
	}
};
ai.tag = "tag:yaml.org,2002:omap";
var oi = {
	collection: "seq",
	identify: (e) => e instanceof Map,
	nodeClass: ai,
	default: !1,
	tag: "tag:yaml.org,2002:omap",
	resolve(e, t) {
		let n = ni(e, t), r = [];
		for (let { key: e } of n.items) q(e) && (r.includes(e.value) ? t(`Ordered maps must not include duplicate keys: ${e.value}`) : r.push(e.value));
		return Object.assign(new ai(), n);
	},
	createNode: (e, t, n) => ai.from(e, t, n)
};
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/schema/yaml-1.1/bool.js
function si({ value: e, source: t }, n) {
	return t && (e ? ci : li).test.test(t) ? t : e ? n.options.trueStr : n.options.falseStr;
}
var ci = {
	identify: (e) => e === !0,
	default: !0,
	tag: "tag:yaml.org,2002:bool",
	test: /^(?:Y|y|[Yy]es|YES|[Tt]rue|TRUE|[Oo]n|ON)$/,
	resolve: () => new X(!0),
	stringify: si
}, li = {
	identify: (e) => e === !1,
	default: !0,
	tag: "tag:yaml.org,2002:bool",
	test: /^(?:N|n|[Nn]o|NO|[Ff]alse|FALSE|[Oo]ff|OFF)$/,
	resolve: () => new X(!1),
	stringify: si
}, ui = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	test: /^(?:[-+]?\.(?:inf|Inf|INF)|\.nan|\.NaN|\.NAN)$/,
	resolve: (e) => e.slice(-3).toLowerCase() === "nan" ? NaN : e[0] === "-" ? -Infinity : Infinity,
	stringify: Br
}, di = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	format: "EXP",
	test: /^[-+]?(?:[0-9][0-9_]*)?(?:\.[0-9_]*)?[eE][-+]?[0-9]+$/,
	resolve: (e) => parseFloat(e.replace(/_/g, "")),
	stringify(e) {
		let t = Number(e.value);
		return isFinite(t) ? t.toExponential() : Br(e);
	}
}, fi = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	test: /^[-+]?(?:[0-9][0-9_]*)?\.[0-9_]*$/,
	resolve(e) {
		let t = new X(parseFloat(e.replace(/_/g, ""))), n = e.indexOf(".");
		if (n !== -1) {
			let r = e.substring(n + 1).replace(/_/g, "");
			r[r.length - 1] === "0" && (t.minFractionDigits = r.length);
		}
		return t;
	},
	stringify: Br
}, pi = (e) => typeof e == "bigint" || Number.isInteger(e);
function mi(e, t, n, { intAsBigInt: r }) {
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
function hi(e, t, n) {
	let { value: r } = e;
	if (pi(r)) {
		let e = r.toString(t);
		return r < 0 ? "-" + n + e.substr(1) : n + e;
	}
	return Br(e);
}
var gi = {
	identify: pi,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "BIN",
	test: /^[-+]?0b[0-1_]+$/,
	resolve: (e, t, n) => mi(e, 2, 2, n),
	stringify: (e) => hi(e, 2, "0b")
}, _i = {
	identify: pi,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "OCT",
	test: /^[-+]?0[0-7_]+$/,
	resolve: (e, t, n) => mi(e, 1, 8, n),
	stringify: (e) => hi(e, 8, "0")
}, vi = {
	identify: pi,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	test: /^[-+]?[0-9][0-9_]*$/,
	resolve: (e, t, n) => mi(e, 0, 10, n),
	stringify: Br
}, yi = {
	identify: pi,
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "HEX",
	test: /^[-+]?0x[0-9a-fA-F_]+$/,
	resolve: (e, t, n) => mi(e, 2, 16, n),
	stringify: (e) => hi(e, 16, "0x")
}, bi = class e extends Mr {
	constructor(t) {
		super(t), this.tag = e.tag;
	}
	add(e) {
		let t;
		t = vn(e) ? e : e && typeof e == "object" && "key" in e && "value" in e && e.value === null ? new Er(e.key, null) : new Er(e, null), jr(this.items, t.key) || this.items.push(t);
	}
	get(e, t) {
		let n = jr(this.items, e);
		return !t && vn(n) ? q(n.key) ? n.key.value : n.key : n;
	}
	set(e, t) {
		if (typeof t != "boolean") throw Error(`Expected boolean value for set(key, value) in a YAML set, not ${typeof t}`);
		let n = jr(this.items, e);
		n && !t ? this.items.splice(this.items.indexOf(n), 1) : !n && t && this.items.push(new Er(e));
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
		if (t && Symbol.iterator in Object(t)) for (let e of t) typeof r == "function" && (e = r.call(t, e, e)), i.items.push(Tr(e, null, n));
		return i;
	}
};
bi.tag = "tag:yaml.org,2002:set";
var xi = {
	collection: "map",
	identify: (e) => e instanceof Set,
	nodeClass: bi,
	default: !1,
	tag: "tag:yaml.org,2002:set",
	createNode: (e, t, n) => bi.from(e, t, n),
	resolve(e, t) {
		if (_n(e)) {
			if (e.hasAllNullValues(!0)) return Object.assign(new bi(), e);
			t("Set items must all have null values");
		} else t("Expected a mapping for this tag");
		return e;
	}
};
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/schema/yaml-1.1/timestamp.js
function Si(e, t) {
	let n = e[0], r = n === "-" || n === "+" ? e.substring(1) : e, i = (e) => t ? BigInt(e) : Number(e), a = r.replace(/_/g, "").split(":").reduce((e, t) => e * i(60) + i(t), i(0));
	return n === "-" ? i(-1) * a : a;
}
function Ci(e) {
	let { value: t } = e, n = (e) => e;
	if (typeof t == "bigint") n = (e) => BigInt(e);
	else if (isNaN(t) || !isFinite(t)) return Br(e);
	let r = "";
	t < 0 && (r = "-", t *= n(-1));
	let i = n(60), a = [t % i];
	return t < 60 ? a.unshift(0) : (t = (t - a[0]) / i, a.unshift(t % i), t >= 60 && (t = (t - a[0]) / i, a.unshift(t))), r + a.map((e) => String(e).padStart(2, "0")).join(":").replace(/000000\d*$/, "");
}
var wi = {
	identify: (e) => typeof e == "bigint" || Number.isInteger(e),
	default: !0,
	tag: "tag:yaml.org,2002:int",
	format: "TIME",
	test: /^[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+$/,
	resolve: (e, t, { intAsBigInt: n }) => Si(e, n),
	stringify: Ci
}, Ti = {
	identify: (e) => typeof e == "number",
	default: !0,
	tag: "tag:yaml.org,2002:float",
	format: "TIME",
	test: /^[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+\.[0-9_]*$/,
	resolve: (e) => Si(e, !1),
	stringify: Ci
}, Ei = {
	identify: (e) => e instanceof Date,
	default: !0,
	tag: "tag:yaml.org,2002:timestamp",
	test: RegExp("^([0-9]{4})-([0-9]{1,2})-([0-9]{1,2})(?:(?:t|T|[ \\t]+)([0-9]{1,2}):([0-9]{1,2}):([0-9]{1,2}(\\.[0-9]+)?)(?:[ \\t]*(Z|[-+][012]?[0-9](?::[0-9]{2})?))?)?$"),
	resolve(e) {
		let t = e.match(Ei.test);
		if (!t) throw Error("!!timestamp expects a date, starting with yyyy-mm-dd");
		let [, n, r, i, a, o, s] = t.map(Number), c = t[7] ? Number((t[7] + "00").substr(1, 3)) : 0, l = Date.UTC(n, r - 1, i, a || 0, o || 0, s || 0, c), u = t[8];
		if (u && u !== "Z") {
			let e = Si(u, !1);
			Math.abs(e) < 30 && (e *= 60), l -= 6e4 * e;
		}
		return new Date(l);
	},
	stringify: ({ value: e }) => e?.toISOString().replace(/(T00:00:00)?\.000Z$/, "") ?? ""
}, Di = [
	Nr,
	Ir,
	Lr,
	Rr,
	ci,
	li,
	gi,
	_i,
	vi,
	yi,
	ui,
	di,
	fi,
	ti,
	yr,
	oi,
	ii,
	xi,
	wi,
	Ti,
	Ei
], Oi = /* @__PURE__ */ new Map([
	["core", Xr],
	["failsafe", [
		Nr,
		Ir,
		Lr
	]],
	["json", ei],
	["yaml11", Di],
	["yaml-1.1", Di]
]), ki = {
	binary: ti,
	bool: zr,
	float: Ur,
	floatExp: Hr,
	floatNaN: Vr,
	floatTime: Ti,
	int: Jr,
	intHex: Yr,
	intOct: qr,
	intTime: wi,
	map: Nr,
	merge: yr,
	null: Rr,
	omap: oi,
	pairs: ii,
	seq: Ir,
	set: xi,
	timestamp: Ei
}, Ai = {
	"tag:yaml.org,2002:binary": ti,
	"tag:yaml.org,2002:merge": yr,
	"tag:yaml.org,2002:omap": oi,
	"tag:yaml.org,2002:pairs": ii,
	"tag:yaml.org,2002:set": xi,
	"tag:yaml.org,2002:timestamp": Ei
};
function ji(e, t, n) {
	let r = Oi.get(t);
	if (r && !e) return n && !r.includes(yr) ? r.concat(yr) : r.slice();
	let i = r;
	if (!i) {
		if (Array.isArray(e)) i = [];
		else {
			let e = Array.from(Oi.keys()).filter((e) => e !== "yaml11").map((e) => JSON.stringify(e)).join(", ");
			throw Error(`Unknown schema "${t}"; use one of ${e} or define customTags array`);
		}
	}
	if (Array.isArray(e)) for (let t of e) i = i.concat(t);
	else typeof e == "function" && (i = e(i.slice()));
	return n && (i = i.concat(yr)), i.reduce((e, t) => {
		let n = typeof t == "string" ? ki[t] : t;
		if (!n) {
			let e = JSON.stringify(t), n = Object.keys(ki).map((e) => JSON.stringify(e)).join(", ");
			throw Error(`Unknown custom tag ${e}; use one of ${n}`);
		}
		return e.includes(n) || e.push(n), e;
	}, []);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/schema/Schema.js
var Mi = (e, t) => e.key < t.key ? -1 : +(e.key > t.key), Ni = class e {
	constructor({ compat: e, customTags: t, merge: n, resolveKnownTags: r, schema: i, sortMapEntries: a, toStringDefaults: o }) {
		this.compat = Array.isArray(e) ? ji(e, "compat") : e ? ji(null, e) : null, this.name = typeof i == "string" && i || "core", this.knownTags = r ? Ai : {}, this.tags = ji(t, this.name, n), this.toStringOptions = o ?? null, Object.defineProperty(this, un, { value: Nr }), Object.defineProperty(this, fn, { value: Lr }), Object.defineProperty(this, pn, { value: Ir }), this.sortMapEntries = typeof a == "function" ? a : a === !0 ? Mi : null;
	}
	clone() {
		let t = Object.create(e.prototype, Object.getOwnPropertyDescriptors(this));
		return t.tags = this.tags.slice(), t;
	}
};
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/stringify/stringifyDocument.js
function Pi(e, t) {
	let n = [], r = t.directives === !0;
	if (t.directives !== !1 && e.directives) {
		let t = e.directives.toString(e);
		t ? (n.push(t), r = !0) : e.directives.docStart && (r = !0);
	}
	r && n.push("---");
	let i = fr(e, t), { commentString: a } = i.options;
	if (e.commentBefore) {
		n.length !== 1 && n.unshift("");
		let t = a(e.commentBefore);
		n.unshift(Yn(t, ""));
	}
	let o = !1, s = null;
	if (e.contents) {
		if (Y(e.contents)) {
			if (e.contents.spaceBefore && r && n.push(""), e.contents.commentBefore) {
				let t = a(e.contents.commentBefore);
				n.push(Yn(t, ""));
			}
			i.forceBlockIndent = !!e.comment, s = e.contents.comment;
		}
		let t = s ? void 0 : () => o = !0, c = hr(e.contents, i, () => s = null, t);
		s && (c += Xn(c, "", a(s))), (c[0] === "|" || c[0] === ">") && n[n.length - 1] === "---" ? n[n.length - 1] = `--- ${c}` : n.push(c);
	} else n.push(hr(e.contents, i));
	if (e.directives?.docEnd) {
		if (e.comment) {
			let t = a(e.comment);
			t.includes("\n") ? (n.push("..."), n.push(Yn(t, ""))) : n.push(`... ${t}`);
		} else n.push("...");
	} else {
		let t = e.comment;
		t && o && (t = t.replace(/^\n+/, "")), t && ((!o || s) && n[n.length - 1] !== "" && n.push(""), n.push(Yn(a(t), "")));
	}
	return n.join("\n") + "\n";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/doc/Document.js
var Fi = class e {
	constructor(e, t, n) {
		this.commentBefore = null, this.comment = null, this.errors = [], this.warnings = [], Object.defineProperty(this, mn, { value: ln });
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
		n?._directives ? (this.directives = n._directives.atDocument(), this.directives.yaml.explicit && (a = this.directives.yaml.version)) : this.directives = new jn({ version: a }), this.setSchema(a, n), this.contents = e === void 0 ? null : this.createNode(e, r, n);
	}
	clone() {
		let t = Object.create(e.prototype, { [mn]: { value: ln } });
		return t.commentBefore = this.commentBefore, t.comment = this.comment, t.errors = this.errors.slice(), t.warnings = this.warnings.slice(), t.options = Object.assign({}, this.options), this.directives && (t.directives = this.directives.clone()), t.schema = this.schema.clone(), t.contents = Y(this.contents) ? this.contents.clone(t.schema) : this.contents, this.range && (t.range = this.range.slice()), t;
	}
	add(e) {
		Ii(this.contents) && this.contents.add(e);
	}
	addIn(e, t) {
		Ii(this.contents) && this.contents.addIn(e, t);
	}
	createAlias(e, t) {
		if (!e.anchor) {
			let n = Nn(this);
			e.anchor = !t || n.has(t) ? Pn(t || "a", n) : t;
		}
		return new zn(e.anchor);
	}
	createNode(e, t, n) {
		let r;
		if (typeof t == "function") e = t.call({ "": e }, "", e), r = t;
		else if (Array.isArray(t)) {
			let e = t.filter((e) => typeof e == "number" || e instanceof String || e instanceof Number).map(String);
			e.length > 0 && (t = t.concat(e)), r = t;
		} else n === void 0 && t && (n = t, t = void 0);
		let { aliasDuplicateObjects: i, anchorPrefix: a, flow: o, keepUndefined: s, onTagObj: c, tag: l } = n ?? {}, { onAnchor: u, setAnchors: d, sourceObjects: f } = Fn(this, a || "a"), p = {
			aliasDuplicateObjects: i ?? !0,
			keepUndefined: s ?? !1,
			onAnchor: u,
			onTagObj: c,
			replacer: r,
			schema: this.schema,
			sourceObjects: f
		}, m = Wn(e, l, p);
		return o && J(m) && (m.flow = !0), d(), m;
	}
	createPair(e, t, n = {}) {
		return new Er(this.createNode(e, null, n), this.createNode(t, null, n));
	}
	delete(e) {
		return Ii(this.contents) ? this.contents.delete(e) : !1;
	}
	deleteIn(e) {
		return Kn(e) ? this.contents != null && (this.contents = null, !0) : Ii(this.contents) ? this.contents.deleteIn(e) : !1;
	}
	get(e, t) {
		return J(this.contents) ? this.contents.get(e, t) : void 0;
	}
	getIn(e, t) {
		return Kn(e) ? !t && q(this.contents) ? this.contents.value : this.contents : J(this.contents) ? this.contents.getIn(e, t) : void 0;
	}
	has(e) {
		return J(this.contents) ? this.contents.has(e) : !1;
	}
	hasIn(e) {
		return Kn(e) ? this.contents !== void 0 : J(this.contents) ? this.contents.hasIn(e) : !1;
	}
	set(e, t) {
		this.contents == null ? this.contents = Gn(this.schema, [e], t) : Ii(this.contents) && this.contents.set(e, t);
	}
	setIn(e, t) {
		Kn(e) ? this.contents = t : this.contents == null ? this.contents = Gn(this.schema, Array.from(e), t) : Ii(this.contents) && this.contents.setIn(e, t);
	}
	setSchema(e, t = {}) {
		typeof e == "number" && (e = String(e));
		let n;
		switch (e) {
			case "1.1":
				this.directives ? this.directives.yaml.version = "1.1" : this.directives = new jn({ version: "1.1" }), n = {
					resolveKnownTags: !1,
					schema: "yaml-1.1"
				};
				break;
			case "1.2":
			case "next":
				this.directives ? this.directives.yaml.version = e : this.directives = new jn({ version: e }), n = {
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
		else if (n) this.schema = new Ni(Object.assign(n, t));
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
		}, s = Ln(this.contents, t ?? "", o);
		if (typeof i == "function") for (let { count: e, res: t } of o.anchors.values()) i(t, e);
		return typeof a == "function" ? In(a, { "": s }, "", s) : s;
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
		return Pi(this, e);
	}
};
function Ii(e) {
	if (J(e)) return !0;
	throw Error("Expected a YAML collection as document contents");
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/parse/cst-visit.js
var Li = Symbol("break visit"), Ri = Symbol("skip children"), zi = Symbol("remove item");
function Bi(e, t) {
	"type" in e && e.type === "document" && (e = {
		start: e.start,
		value: e.value
	}), Vi(Object.freeze([]), e, t);
}
Bi.BREAK = Li, Bi.SKIP = Ri, Bi.REMOVE = zi, Bi.itemAtPath = (e, t) => {
	let n = e;
	for (let [e, r] of t) {
		let t = n?.[e];
		if (t && "items" in t) n = t.items[r];
		else return;
	}
	return n;
}, Bi.parentCollection = (e, t) => {
	let n = Bi.itemAtPath(e, t.slice(0, -1)), r = t[t.length - 1][0], i = n?.[r];
	if (i && "items" in i) return i;
	throw Error("Parent collection not found");
};
function Vi(e, t, n) {
	let r = n(t, e);
	if (typeof r == "symbol") return r;
	for (let i of ["key", "value"]) {
		let a = t[i];
		if (a && "items" in a) {
			for (let t = 0; t < a.items.length; ++t) {
				let r = Vi(Object.freeze(e.concat([[i, t]])), a.items[t], n);
				if (typeof r == "number") t = r - 1;
				else if (r === Li) return Li;
				else r === zi && (a.items.splice(t, 1), --t);
			}
			typeof r == "function" && i === "key" && (r = r(t, e));
		}
	}
	return typeof r == "function" ? r(t, e) : r;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/yaml/2.8.3/5368704e419f24a904000290efb263e6de32091b7f282feafee13695476a49fc/node_modules/yaml/browser/dist/public-api.js
function Hi(e, t, n) {
	let r = null;
	if (typeof t == "function" || Array.isArray(t) ? r = t : n === void 0 && t && (n = t), typeof n == "string" && (n = n.length), typeof n == "number") {
		let e = Math.round(n);
		n = e < 1 ? void 0 : e > 8 ? { indent: 8 } : { indent: e };
	}
	if (e === void 0) {
		let { keepUndefined: e } = n ?? t ?? {};
		if (!e) return;
	}
	return gn(e) && !r ? e.toString(n) : new Fi(e, r, n).toString(n);
}
// ../../../clicky-ui/packages/ui/src/components/ItemActions.tsx
function Ui({ label: e, index: t, count: n, onMove: r, onDuplicate: a, onRemove: s, leading: c, reveal: l = !0, size: u = "md", className: d }) {
	let f = i("inline-flex aspect-square items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-foreground disabled:opacity-30", o[u]);
	return /* @__PURE__ */ (0, N.jsxs)("div", {
		className: i("ml-auto flex shrink-0 items-center gap-0.5", l && "opacity-0 transition-opacity focus-within:opacity-100 group-hover:opacity-100 [@media(hover:none)]:opacity-100", d),
		children: [
			c,
			r && /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [/* @__PURE__ */ (0, N.jsx)("button", {
				type: "button",
				"aria-label": `Move ${e} up`,
				disabled: t === 0,
				className: f,
				onClick: () => r(t - 1),
				children: /* @__PURE__ */ (0, N.jsx)(H, {
					icon: qe,
					className: "text-sm"
				})
			}), /* @__PURE__ */ (0, N.jsx)("button", {
				type: "button",
				"aria-label": `Move ${e} down`,
				disabled: t >= n - 1,
				className: f,
				onClick: () => r(t + 1),
				children: /* @__PURE__ */ (0, N.jsx)(H, {
					icon: Ge,
					className: "text-sm"
				})
			})] }),
			a && /* @__PURE__ */ (0, N.jsx)("button", {
				type: "button",
				"aria-label": `Duplicate ${e}`,
				className: f,
				onClick: a,
				children: /* @__PURE__ */ (0, N.jsx)(H, {
					icon: ct,
					className: "text-sm"
				})
			}),
			s && /* @__PURE__ */ (0, N.jsx)("button", {
				type: "button",
				"aria-label": `Remove ${e}`,
				className: i(f, "hover:bg-destructive/10 hover:text-destructive"),
				onClick: s,
				"data-item-index": t,
				children: /* @__PURE__ */ (0, N.jsx)(H, {
					icon: dt,
					className: "text-sm"
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/components/AccordionList.tsx
var Wi = "[data-autofocus], input, select, textarea";
function Gi({ items: e, onChange: t, renderHeader: n, renderBody: r, renderActions: a, renderMeta: o, summary: d, toolbar: p, itemLabel: m, allowReorder: h = !1, allowDrag: g = !1, canDrag: _, onReorder: v, allowDuplicate: y = !1, allowRemove: b = !1, revealActions: x = !0, cloneItem: S, onCreate: C, addLabel: w = "Add item", addDescription: T, expanded: E, defaultExpanded: D = null, onExpandedChange: O, readOnly: k = !1, size: A = "md", idPrefix: j = "accordion", itemId: P, focusSelector: F = Wi, className: I, listClassName: L, rowClassName: ee, bodyClassName: R }) {
	let z = E !== void 0, [te, B] = (0, M.useState)(D), V = z ? E : te, [ne, U] = (0, M.useState)(null), [re, W] = (0, M.useState)(null), ie = (0, M.useRef)([]), G = (0, M.useRef)(null), ae = (0, M.useRef)(null);
	(0, M.useEffect)(() => {
		let e = ae.current;
		if (!e) return;
		if (ae.current = null, e.kind === "add") {
			G.current?.focus();
			return;
		}
		let t = ie.current[e.index];
		e.kind === "row" ? t?.focus() : (document.getElementById(se(e.index))?.querySelector(F) ?? t)?.focus();
	});
	function oe(e) {
		return P ? P(e) : `${j}-${e}`;
	}
	function se(e) {
		return `${oe(e)}-panel`;
	}
	function ce(e) {
		return `${oe(e)}-header`;
	}
	function le(e) {
		z || B(e), O?.(e);
	}
	function ue(e) {
		let t = e(V);
		t !== V && le(t);
	}
	function de(e) {
		t?.(e);
	}
	function fe() {
		C && (ae.current = {
			kind: "panel",
			index: e.length
		}, le(e.length), de([...e, C()]));
	}
	function pe(t) {
		ue((e) => e === t ? null : e !== null && e > t ? e - 1 : e), ae.current = e.length === 1 ? {
			kind: "add",
			index: 0
		} : {
			kind: "row",
			index: Math.max(t - 1, 0)
		}, de(l(e, t));
	}
	function me(t) {
		ue((e) => e !== null && e > t ? e + 1 : e), ae.current = {
			kind: "row",
			index: t + 1
		}, de(s(e, t, S));
	}
	function he(e) {
		return ne === null || re !== e || ne === e ? null : ne < e ? "bottom" : "top";
	}
	function ge() {
		U(null), W(null);
	}
	function _e(e, t) {
		e !== t && (v ? v(e, t) : ve(e, t));
	}
	function ve(t, n) {
		n < 0 || n >= e.length || (ue((e) => e === t ? n : e === n ? t : e), ae.current = {
			kind: "row",
			index: n
		}, de(u(e, t, n)));
	}
	function ye(e, t) {
		let n = ie.current.length, r = (t) => {
			e.preventDefault(), (t >= n ? G.current : ie.current[t])?.focus();
		};
		e.key === "ArrowDown" ? r(Math.min(t + 1, n)) : e.key === "ArrowUp" && t > 0 ? r(t - 1) : e.key === "Home" ? r(0) : e.key === "End" && r(n);
	}
	return /* @__PURE__ */ (0, N.jsxs)("div", {
		className: i("flex flex-col gap-2", I),
		children: [(d !== void 0 || p !== void 0) && /* @__PURE__ */ (0, N.jsxs)("div", {
			"data-accordion-summary": !0,
			className: "flex items-center justify-between gap-2",
			children: [d !== void 0 && /* @__PURE__ */ (0, N.jsx)("p", {
				className: i("text-muted-foreground", f[A]),
				children: d
			}), p !== void 0 && /* @__PURE__ */ (0, N.jsx)("div", {
				className: "ml-auto flex shrink-0 items-center",
				children: p
			})]
		}), /* @__PURE__ */ (0, N.jsxs)("div", {
			className: i("divide-y divide-border rounded-lg border border-border", L),
			children: [e.map((t, s) => {
				let l = V === s, u = {
					item: t,
					index: s,
					open: l,
					onChange: (t) => de(c(e, s, t)),
					headerId: ce(s),
					panelId: se(s)
				}, d = m?.({
					item: t,
					index: s
				}) ?? `Item ${s + 1}`, f = a?.(u), p = o?.(u), v = g && !k && (_?.({
					item: t,
					index: s
				}) ?? !0), S = v && ne !== null && ne !== s, C = S ? he(s) : null;
				return /* @__PURE__ */ (0, N.jsxs)("div", {
					className: "group",
					children: [/* @__PURE__ */ (0, N.jsxs)("div", {
						"data-accordion-row": !0,
						"data-drop-edge": C ?? void 0,
						onDragOver: (e) => {
							S && (e.preventDefault(), e.dataTransfer.dropEffect = "move", W(s));
						},
						onDrop: (e) => {
							S && ne !== null && (e.preventDefault(), _e(ne, s), ge());
						},
						className: i("flex w-full items-center gap-2 px-3 py-2 hover:bg-accent/40", s % 2 == 0 ? "bg-card" : "bg-muted/20", l && "bg-accent/30", ne === s && "opacity-40", C === "top" && "shadow-[inset_0_2px_0_0_var(--color-primary)]", C === "bottom" && "shadow-[inset_0_-2px_0_0_var(--color-primary)]", ee),
						children: [
							g && !k && /* @__PURE__ */ (0, N.jsx)("button", {
								type: "button",
								"aria-label": `Reorder ${d}`,
								title: `Reorder ${d}`,
								disabled: !v,
								draggable: v,
								onDragStart: (e) => {
									e.dataTransfer.effectAllowed = "move", e.dataTransfer.setData("text/plain", String(s));
									let t = e.currentTarget.closest("[data-accordion-row]");
									t instanceof HTMLElement && e.dataTransfer.setDragImage(t, 12, 12), U(s);
								},
								onDragEnd: ge,
								onKeyDown: (t) => {
									let n = t.key === "ArrowUp" ? s - 1 : t.key === "ArrowDown" ? s + 1 : s;
									n === s || n < 0 || n >= e.length || (t.preventDefault(), _e(s, n));
								},
								className: "shrink-0 cursor-grab rounded text-muted-foreground hover:text-foreground active:cursor-grabbing disabled:cursor-default disabled:opacity-30",
								children: /* @__PURE__ */ (0, N.jsx)(H, {
									icon: et,
									className: "text-sm"
								})
							}),
							/* @__PURE__ */ (0, N.jsxs)("button", {
								type: "button",
								id: ce(s),
								"aria-expanded": l,
								"aria-controls": se(s),
								ref: (e) => {
									ie.current[s] = e;
								},
								onKeyDown: (e) => ye(e, s),
								onClick: () => le(l ? null : s),
								className: "flex min-w-0 flex-1 items-center gap-2 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
								children: [/* @__PURE__ */ (0, N.jsx)(H, {
									icon: l ? Ge : Ke,
									className: "shrink-0 text-sm text-muted-foreground"
								}), n(u)]
							}),
							p,
							!k && (h || y || b || f) && /* @__PURE__ */ (0, N.jsx)(Ui, {
								label: d,
								index: s,
								count: e.length,
								size: A,
								reveal: x,
								...f ? { leading: f } : {},
								...h ? { onMove: (e) => ve(s, e) } : {},
								...y ? { onDuplicate: () => me(s) } : {},
								...b ? { onRemove: () => pe(s) } : {}
							})
						]
					}), l && /* @__PURE__ */ (0, N.jsx)("div", {
						id: se(s),
						role: "region",
						"aria-labelledby": ce(s),
						className: i("border-t border-border bg-muted/20 p-3", R),
						children: r(u)
					})]
				}, s);
			}), !k && C && /* @__PURE__ */ (0, N.jsx)(Ki, {
				buttonRef: G,
				label: w,
				empty: e.length === 0,
				onAdd: fe,
				onKeyDown: (t) => ye(t, e.length),
				...T ? { copy: T } : {}
			})]
		})]
	});
}
function Ki({ label: e, copy: t, empty: n, onAdd: r, onKeyDown: a, buttonRef: o }) {
	return /* @__PURE__ */ (0, N.jsxs)("button", {
		type: "button",
		ref: (e) => {
			o.current = e;
		},
		onClick: r,
		onKeyDown: a,
		className: i("group/add flex w-full items-center gap-2 px-3 py-2 text-left text-muted-foreground hover:bg-accent/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring", n && "items-start py-5"),
		children: [
			/* @__PURE__ */ (0, N.jsx)(H, {
				icon: Ke,
				className: "shrink-0 text-sm opacity-0"
			}),
			/* @__PURE__ */ (0, N.jsx)("span", {
				className: "inline-flex size-6 shrink-0 items-center justify-center rounded-md border border-dashed border-border text-foreground/60 group-hover/add:border-primary group-hover/add:text-primary",
				children: /* @__PURE__ */ (0, N.jsx)(H, {
					icon: Je,
					className: "text-[13px]"
				})
			}),
			/* @__PURE__ */ (0, N.jsxs)("span", {
				className: "flex min-w-0 flex-col gap-0.5",
				children: [/* @__PURE__ */ (0, N.jsx)("span", {
					className: "text-sm font-medium text-foreground",
					children: e
				}), t && /* @__PURE__ */ (0, N.jsx)("span", {
					className: i("text-xs", n ? "max-w-xl leading-snug" : "truncate"),
					children: t
				})]
			})
		]
	});
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/jsonc-parser/3.3.1/58003957d75a2d2251ecd31c04d4addb7d50a9ac64da586f7d3632efb5050498/node_modules/jsonc-parser/lib/esm/impl/scanner.js
function qi(e, t = !1) {
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
		else for (r++; r < e.length && Xi(e.charCodeAt(r));) r++;
		if (r < e.length && e.charCodeAt(r) === 46) {
			if (r++, r < e.length && Xi(e.charCodeAt(r))) for (r++; r < e.length && Xi(e.charCodeAt(r));) r++;
			else return d = 3, e.substring(t, r);
		}
		let n = r;
		if (r < e.length && (e.charCodeAt(r) === 69 || e.charCodeAt(r) === 101)) {
			if (r++, (r < e.length && e.charCodeAt(r) === 43 || e.charCodeAt(r) === 45) && r++, r < e.length && Xi(e.charCodeAt(r))) {
				for (r++; r < e.length && Xi(e.charCodeAt(r));) r++;
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
					if (Yi(a)) {
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
		if (Ji(t)) {
			do
				r++, i += String.fromCharCode(t), t = e.charCodeAt(r);
			while (Ji(t));
			return o = 15;
		}
		if (Yi(t)) return r++, i += String.fromCharCode(t), t === 13 && e.charCodeAt(r) === 10 && (r++, i += "\n"), s++, l = r, o = 14;
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
					for (r += 2; r < n && !Yi(e.charCodeAt(r));) r++;
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
						r++, Yi(t) && (t === 13 && e.charCodeAt(r) === 10 && r++, s++, l = r);
					}
					return a || (r++, d = 1), i = e.substring(c, r), o = 13;
				}
				return i += String.fromCharCode(t), r++, o = 16;
			case 45: if (i += String.fromCharCode(t), r++, r === n || !Xi(e.charCodeAt(r))) return o = 16;
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
		if (Ji(e) || Yi(e)) return !1;
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
function Ji(e) {
	return e === 32 || e === 9;
}
function Yi(e) {
	return e === 10 || e === 13;
}
function Xi(e) {
	return e >= 48 && e <= 57;
}
var Zi;
(function(e) {
	e[e.lineFeed = 10] = "lineFeed", e[e.carriageReturn = 13] = "carriageReturn", e[e.space = 32] = "space", e[e._0 = 48] = "_0", e[e._1 = 49] = "_1", e[e._2 = 50] = "_2", e[e._3 = 51] = "_3", e[e._4 = 52] = "_4", e[e._5 = 53] = "_5", e[e._6 = 54] = "_6", e[e._7 = 55] = "_7", e[e._8 = 56] = "_8", e[e._9 = 57] = "_9", e[e.a = 97] = "a", e[e.b = 98] = "b", e[e.c = 99] = "c", e[e.d = 100] = "d", e[e.e = 101] = "e", e[e.f = 102] = "f", e[e.g = 103] = "g", e[e.h = 104] = "h", e[e.i = 105] = "i", e[e.j = 106] = "j", e[e.k = 107] = "k", e[e.l = 108] = "l", e[e.m = 109] = "m", e[e.n = 110] = "n", e[e.o = 111] = "o", e[e.p = 112] = "p", e[e.q = 113] = "q", e[e.r = 114] = "r", e[e.s = 115] = "s", e[e.t = 116] = "t", e[e.u = 117] = "u", e[e.v = 118] = "v", e[e.w = 119] = "w", e[e.x = 120] = "x", e[e.y = 121] = "y", e[e.z = 122] = "z", e[e.A = 65] = "A", e[e.B = 66] = "B", e[e.C = 67] = "C", e[e.D = 68] = "D", e[e.E = 69] = "E", e[e.F = 70] = "F", e[e.G = 71] = "G", e[e.H = 72] = "H", e[e.I = 73] = "I", e[e.J = 74] = "J", e[e.K = 75] = "K", e[e.L = 76] = "L", e[e.M = 77] = "M", e[e.N = 78] = "N", e[e.O = 79] = "O", e[e.P = 80] = "P", e[e.Q = 81] = "Q", e[e.R = 82] = "R", e[e.S = 83] = "S", e[e.T = 84] = "T", e[e.U = 85] = "U", e[e.V = 86] = "V", e[e.W = 87] = "W", e[e.X = 88] = "X", e[e.Y = 89] = "Y", e[e.Z = 90] = "Z", e[e.asterisk = 42] = "asterisk", e[e.backslash = 92] = "backslash", e[e.closeBrace = 125] = "closeBrace", e[e.closeBracket = 93] = "closeBracket", e[e.colon = 58] = "colon", e[e.comma = 44] = "comma", e[e.dot = 46] = "dot", e[e.doubleQuote = 34] = "doubleQuote", e[e.minus = 45] = "minus", e[e.openBrace = 123] = "openBrace", e[e.openBracket = 91] = "openBracket", e[e.plus = 43] = "plus", e[e.slash = 47] = "slash", e[e.formFeed = 12] = "formFeed", e[e.tab = 9] = "tab";
})(Zi ||= {}), Array(20).fill(0).map((e, t) => " ".repeat(t));
var Qi = 200;
Array(Qi).fill(0).map((e, t) => "\n" + " ".repeat(t)), Array(Qi).fill(0).map((e, t) => "\r" + " ".repeat(t)), Array(Qi).fill(0).map((e, t) => "\r\n" + " ".repeat(t)), Array(Qi).fill(0).map((e, t) => "\n" + "	".repeat(t)), Array(Qi).fill(0).map((e, t) => "\r" + "	".repeat(t)), Array(Qi).fill(0).map((e, t) => "\r\n" + "	".repeat(t));
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/jsonc-parser/3.3.1/58003957d75a2d2251ecd31c04d4addb7d50a9ac64da586f7d3632efb5050498/node_modules/jsonc-parser/lib/esm/impl/parser.js
var $i;
(function(e) {
	e.DEFAULT = { allowTrailingComma: !1 };
})($i ||= {});
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/jsonc-parser/3.3.1/58003957d75a2d2251ecd31c04d4addb7d50a9ac64da586f7d3632efb5050498/node_modules/jsonc-parser/lib/esm/main.js
var ea = qi, ta;
(function(e) {
	e[e.None = 0] = "None", e[e.UnexpectedEndOfComment = 1] = "UnexpectedEndOfComment", e[e.UnexpectedEndOfString = 2] = "UnexpectedEndOfString", e[e.UnexpectedEndOfNumber = 3] = "UnexpectedEndOfNumber", e[e.InvalidUnicode = 4] = "InvalidUnicode", e[e.InvalidEscapeCharacter = 5] = "InvalidEscapeCharacter", e[e.InvalidCharacter = 6] = "InvalidCharacter";
})(ta ||= {});
var na;
(function(e) {
	e[e.OpenBraceToken = 1] = "OpenBraceToken", e[e.CloseBraceToken = 2] = "CloseBraceToken", e[e.OpenBracketToken = 3] = "OpenBracketToken", e[e.CloseBracketToken = 4] = "CloseBracketToken", e[e.CommaToken = 5] = "CommaToken", e[e.ColonToken = 6] = "ColonToken", e[e.NullKeyword = 7] = "NullKeyword", e[e.TrueKeyword = 8] = "TrueKeyword", e[e.FalseKeyword = 9] = "FalseKeyword", e[e.StringLiteral = 10] = "StringLiteral", e[e.NumericLiteral = 11] = "NumericLiteral", e[e.LineCommentTrivia = 12] = "LineCommentTrivia", e[e.BlockCommentTrivia = 13] = "BlockCommentTrivia", e[e.LineBreakTrivia = 14] = "LineBreakTrivia", e[e.Trivia = 15] = "Trivia", e[e.Unknown = 16] = "Unknown", e[e.EOF = 17] = "EOF";
})(na ||= {});
var ra;
(function(e) {
	e[e.InvalidSymbol = 1] = "InvalidSymbol", e[e.InvalidNumberFormat = 2] = "InvalidNumberFormat", e[e.PropertyNameExpected = 3] = "PropertyNameExpected", e[e.ValueExpected = 4] = "ValueExpected", e[e.ColonExpected = 5] = "ColonExpected", e[e.CommaExpected = 6] = "CommaExpected", e[e.CloseBraceExpected = 7] = "CloseBraceExpected", e[e.CloseBracketExpected = 8] = "CloseBracketExpected", e[e.EndOfFileExpected = 9] = "EndOfFileExpected", e[e.InvalidCommentToken = 10] = "InvalidCommentToken", e[e.UnexpectedEndOfComment = 11] = "UnexpectedEndOfComment", e[e.UnexpectedEndOfString = 12] = "UnexpectedEndOfString", e[e.UnexpectedEndOfNumber = 13] = "UnexpectedEndOfNumber", e[e.InvalidUnicode = 14] = "InvalidUnicode", e[e.InvalidEscapeCharacter = 15] = "InvalidEscapeCharacter", e[e.InvalidCharacter = 16] = "InvalidCharacter";
})(ra ||= {});
// ../../../clicky-ui/packages/ui/src/data/json-view/parse-prefix.ts
var ia = /^"(?:[^"\\]|\\["\\/bfnrt]|\\u[\da-fA-F]{4})*(?:\\(?:u[\da-fA-F]{0,3})?)?$/, aa = /^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d*)?$/, oa = /^-?(?:0|[1-9]\d*)\.$/;
function sa(e) {
	if (e.startsWith("\"")) {
		for (let t of e) if (t.charCodeAt(0) < 32) return !1;
		return ia.test(e);
	}
	return e === "-" || aa.test(e) || oa.test(e) || [
		"true",
		"false",
		"null"
	].some((t) => t.startsWith(e));
}
function ca(e) {
	let t = ea(e), n = () => e.slice(t.getTokenOffset(), t.getTokenOffset() + t.getTokenLength()), r = (e, n) => ({
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
			return r(n + a.length === e.length && sa(a) ? "incomplete" : "invalid", "Expected a complete JSON string, number, boolean, or null");
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
function la(e, t = "json") {
	switch (t) {
		case "json": return ua(e);
		case "ndjson": return da(e);
		default: throw Error(`Unsupported JSON input format: ${String(t)}`);
	}
}
function ua(e) {
	try {
		return {
			status: "complete",
			value: JSON.parse(e)
		};
	} catch (e) {
		if (!(e instanceof SyntaxError)) throw e;
	}
	let t = ca(e);
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
function da(e) {
	let t = e.split(/\r\n|\n/), n = t.findLastIndex((e) => e.trim().length > 0), r = [];
	for (let [e, i] of t.entries()) {
		if (i.trim().length === 0) continue;
		let t = ua(i);
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
function fa({ source: e, inputFormat: t, renderValue: n }) {
	let r = (0, M.useId)(), i = (0, M.useMemo)(() => la(e, t), [e, t]), a = t === "ndjson" ? "NDJSON" : "JSON";
	if (i.status !== "complete" && !i.diagnostic) throw Error("JSON source status has no diagnostic");
	return /* @__PURE__ */ (0, N.jsxs)("div", {
		className: "space-y-density-2",
		"data-source-status": i.status,
		children: [
			"value" in i && n(i.value),
			i.partialRecord && /* @__PURE__ */ (0, N.jsxs)("div", {
				className: "space-y-density-1",
				children: [/* @__PURE__ */ (0, N.jsxs)("div", {
					className: "text-muted-foreground font-sans text-sm",
					children: [
						"Line ",
						i.partialRecord.line,
						" (incomplete record)"
					]
				}), n(i.partialRecord.value)]
			}),
			i.status === "incomplete" && !("value" in i) && !i.partialRecord && /* @__PURE__ */ (0, N.jsx)("div", {
				className: "font-sans text-sm text-muted-foreground",
				children: "No completed values received."
			}),
			/* @__PURE__ */ (0, N.jsx)(Gi, {
				idPrefix: r,
				items: [e],
				readOnly: !0,
				renderHeader: () => /* @__PURE__ */ (0, N.jsxs)("span", {
					className: "flex min-w-0 flex-wrap items-center gap-density-2 font-sans text-sm",
					children: ["Original source", i.diagnostic && /* @__PURE__ */ (0, N.jsxs)("span", {
						role: i.status === "invalid" ? "alert" : "status",
						className: "inline-flex flex-wrap items-center gap-density-2",
						children: [/* @__PURE__ */ (0, N.jsxs)(Fe, {
							variant: "status",
							status: i.status === "invalid" ? "error" : "warning",
							size: "xs",
							clickToCopy: !1,
							children: [
								i.status === "invalid" ? "Invalid" : "Incomplete",
								" ",
								a
							]
						}), /* @__PURE__ */ (0, N.jsxs)("span", {
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
				renderBody: ({ item: e }) => /* @__PURE__ */ (0, N.jsxs)("div", {
					className: "space-y-density-2",
					children: [i.diagnostic && /* @__PURE__ */ (0, N.jsxs)("p", {
						className: "font-sans text-sm text-muted-foreground whitespace-normal",
						children: [
							i.diagnostic.message,
							".",
							i.status === "incomplete" && " Showing completed values only."
						]
					}), /* @__PURE__ */ (0, N.jsx)("pre", {
						className: "max-h-80 overflow-auto whitespace-pre-wrap break-words font-mono text-xs",
						children: e
					})]
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/JsonView.tsx
var pa = {
	compact: "leading-4 [--json-view-indent:0.75rem] [--json-view-row-space:0px]",
	comfortable: "leading-5 [--json-view-indent:1rem] [--json-view-row-space:1px]",
	spacious: "leading-6 [--json-view-indent:1.25rem] [--json-view-row-space:2px]"
};
function ma({ density: e, format: t = "yaml", ...n }) {
	let r = L(), a = e ?? r, o = Object.hasOwn(n, "source");
	if (o === Object.hasOwn(n, "data")) throw Error("JsonView requires exactly one of data or source");
	if (!o && Object.hasOwn(n, "inputFormat")) throw Error("JsonView inputFormat requires source");
	if (o && typeof n.source != "string") throw Error("JsonView source must be a string");
	let { data: s, source: c, inputFormat: l = "json", ...u } = n;
	return /* @__PURE__ */ (0, N.jsx)("div", {
		"data-format": t,
		"data-density": a,
		className: i("font-mono text-density-base whitespace-pre-wrap break-words", pa[a]),
		children: o && c !== void 0 ? /* @__PURE__ */ (0, N.jsx)(fa, {
			source: c,
			inputFormat: l,
			renderValue: (e) => /* @__PURE__ */ (0, N.jsx)(ha, {
				...u,
				data: e,
				format: t
			})
		}) : /* @__PURE__ */ (0, N.jsx)(ha, {
			...u,
			data: s,
			format: t
		})
	});
}
function ha({ data: e, name: t, depth: n = 0, defaultOpenDepth: r = 2, format: o, sequenceItem: s = !1 }) {
	let [c, l] = (0, M.useState)(n < r), u = Array.isArray(e), d = typeof e == "object" && e ? Object.entries(e) : [], f = /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [s && /* @__PURE__ */ (0, N.jsx)("span", {
		className: "text-muted-foreground",
		children: "- "
	}), t !== void 0 && /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [/* @__PURE__ */ (0, N.jsx)("span", {
		className: "text-purple-600 [[data-theme=dark]_&]:text-purple-400",
		children: o === "yaml" ? _a(t) : t
	}), /* @__PURE__ */ (0, N.jsx)("span", {
		className: "text-muted-foreground",
		children: ": "
	})] })] });
	if (d.length === 0) return /* @__PURE__ */ (0, N.jsxs)("div", {
		className: "py-[var(--json-view-row-space)]",
		children: [f, /* @__PURE__ */ (0, N.jsx)(ga, {
			data: e,
			format: o
		})]
	});
	let [p, m] = u ? ["[", "]"] : ["{", "}"], h = u ? `${d.length} ${d.length === 1 ? "item" : "items"}` : "Object";
	return /* @__PURE__ */ (0, N.jsxs)("div", { children: [/* @__PURE__ */ (0, N.jsxs)(a, {
		variant: "ghost",
		type: "button",
		"aria-expanded": c,
		"aria-label": `${c ? "Collapse" : "Expand"} ${t ?? "value"}`,
		className: "h-auto max-w-full justify-start gap-0 rounded px-0 py-[var(--json-view-row-space)] text-left font-normal text-[length:inherit] leading-[inherit] whitespace-pre-wrap [&_svg]:size-[0.75em]",
		onClick: () => l(!c),
		children: [/* @__PURE__ */ (0, N.jsx)(Ke, {
			"aria-hidden": "true",
			className: i("mr-1 shrink-0 text-muted-foreground", c && "rotate-90")
		}), /* @__PURE__ */ (0, N.jsxs)("span", { children: [f, /* @__PURE__ */ (0, N.jsx)("span", {
			className: "text-muted-foreground",
			children: o === "json" ? c ? p : u ? `${p} ${h} ${m}` : `${p}…${m}` : !c || t === void 0 && !s ? h : null
		})] })]
	}), c && /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [/* @__PURE__ */ (0, N.jsx)("div", {
		className: "border-l border-border pl-[var(--json-view-indent)]",
		children: d.map(([e, t]) => /* @__PURE__ */ (0, N.jsx)(ha, {
			data: t,
			...u ? {} : { name: e },
			sequenceItem: u && o === "yaml",
			depth: n + 1,
			defaultOpenDepth: r,
			format: o
		}, e))
	}), o === "json" && /* @__PURE__ */ (0, N.jsx)("div", {
		className: "text-muted-foreground py-[var(--json-view-row-space)]",
		children: m
	})] })] });
}
function ga({ data: e, format: t }) {
	return e == null ? /* @__PURE__ */ (0, N.jsx)("span", {
		className: "text-muted-foreground italic",
		children: "null"
	}) : typeof e == "string" ? /* @__PURE__ */ (0, N.jsx)("span", {
		className: "text-green-700 [[data-theme=dark]_&]:text-green-400",
		children: t === "yaml" ? _a(e) : JSON.stringify(e)
	}) : typeof e == "number" || typeof e == "boolean" ? /* @__PURE__ */ (0, N.jsx)("span", {
		className: "text-blue-700 [[data-theme=dark]_&]:text-blue-400",
		children: String(e)
	}) : /* @__PURE__ */ (0, N.jsx)("span", {
		className: "text-muted-foreground",
		children: typeof e == "object" ? Array.isArray(e) ? "[]" : "{}" : String(e)
	});
}
function _a(e) {
	return Hi(e, {
		lineWidth: 0,
		defaultStringType: e.includes("\n") ? "QUOTE_DOUBLE" : "PLAIN"
	}).trimEnd();
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/dompurify/3.4.11/2c5536b07b36896ff81fe8feb25a9a203f10cff0328d7a0904382bdccdfdadbf/node_modules/dompurify/dist/purify.es.mjs
function va(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function ya(e) {
	if (Array.isArray(e)) return e;
}
function ba(e, t) {
	var n = e == null ? null : typeof Symbol < "u" && e[Symbol.iterator] || e["@@iterator"];
	if (n != null) {
		var r, i, a, o, s = [], c = !0, l = !1;
		try {
			if (a = (n = n.call(e)).next, t !== 0) for (; !(c = (r = a.call(n)).done) && (s.push(r.value), s.length !== t); c = !0);
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
function xa() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function Sa(e, t) {
	return ya(e) || ba(e, t) || Ca(e, t) || xa();
}
function Ca(e, t) {
	if (e) {
		if (typeof e == "string") return va(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? va(e, t) : void 0;
	}
}
var wa = Object.entries, Ta = Object.setPrototypeOf, Ea = Object.isFrozen, Da = Object.getPrototypeOf, Oa = Object.getOwnPropertyDescriptor, ka = Object.freeze, Aa = Object.seal, ja = Object.create, Ma = typeof Reflect < "u" && Reflect, Na = Ma.apply, Pa = Ma.construct;
ka ||= function(e) {
	return e;
}, Aa ||= function(e) {
	return e;
}, Na ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, Pa ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var Fa = Z(Array.prototype.forEach), Ia = Z(Array.prototype.lastIndexOf), La = Z(Array.prototype.pop), Ra = Z(Array.prototype.push), za = Z(Array.prototype.splice), Ba = Array.isArray, Va = Z(String.prototype.toLowerCase), Ha = Z(String.prototype.toString), Ua = Z(String.prototype.match), Wa = Z(String.prototype.replace), Ga = Z(String.prototype.indexOf), Ka = Z(String.prototype.trim), qa = Z(Number.prototype.toString), Ja = Z(Boolean.prototype.toString), Ya = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), Xa = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), Za = Z(Object.prototype.hasOwnProperty), Qa = Z(Object.prototype.toString), $a = Z(RegExp.prototype.test), eo = to(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return Na(e, t, n);
	};
}
function to(e) {
	return function() {
		return Pa(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : Va;
	if (Ta && Ta(e, null), !Ba(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (Ea(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function no(e) {
	for (let t = 0; t < e.length; t++) Za(e, t) || (e[t] = null);
	return e;
}
function ro(e) {
	let t = ja(null);
	for (let r of wa(e)) {
		var n = Sa(r, 2);
		let i = n[0], a = n[1];
		Za(e, i) && (t[i] = Ba(a) ? no(a) : a && typeof a == "object" && a.constructor === Object ? ro(a) : a);
	}
	return t;
}
function io(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return qa(e);
		case "boolean": return Ja(e);
		case "bigint": return Ya ? Ya(e) : "0";
		case "symbol": return Xa ? Xa(e) : "Symbol()";
		case "undefined": return Qa(e);
		case "function":
		case "object": {
			if (e === null) return Qa(e);
			let t = e, n = ao(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : Qa(e);
			}
			return Qa(e);
		}
		default: return Qa(e);
	}
}
function ao(e, t) {
	for (; e !== null;) {
		let n = Oa(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = Da(e);
	}
	function n() {
		return null;
	}
	return n;
}
function oo(e) {
	try {
		return $a(e, ""), !0;
	} catch {
		return !1;
	}
}
var so = ka(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), co = ka(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), lo = ka([
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
]), uo = ka([
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
]), fo = ka(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), po = ka([
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
]), mo = ka(["#text"]), ho = ka(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), go = ka(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-rendering.textlength.type.u1.u2.unicode.values.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), _o = ka(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), vo = ka([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), yo = Aa(/{{[\w\W]*|^[\w\W]*}}/g), bo = Aa(/<%[\w\W]*|^[\w\W]*%>/g), xo = Aa(/\${[\w\W]*/g), So = Aa(/^data-[\-\w.\u00B7-\uFFFF]+$/), Co = Aa(/^aria-[\-\w]+$/), wo = Aa(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), To = Aa(/^(?:\w+script|data):/i), Eo = Aa(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), Do = Aa(/^html$/i), Oo = Aa(/^[a-z][.\w]*(-[.\w]+)+$/i), ko = Aa(/<[/\w!]/g), Ao = Aa(/<[/\w]/g), jo = Aa(/<\/no(script|embed|frames)/i), Mo = Aa(/\/>/i), No = {
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
}, Po = function() {
	return typeof window > "u" ? null : window;
}, Fo = function(e, t) {
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
}, Io = function() {
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
}, Lo = function(e, t, n, r) {
	return Za(e, t) && Ba(e[t]) ? Q(r.base ? ro(r.base) : {}, e[t], r.transform) : n;
};
function Ro() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : Po(), t = (e) => Ro(e);
	if (t.version = "3.4.11", t.removed = [], !e || !e.document || e.document.nodeType !== No.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = ao(d, "cloneNode"), p = ao(d, "remove"), m = ao(d, "nextSibling"), h = ao(d, "childNodes"), g = ao(d, "parentNode"), _ = ao(d, "shadowRoot"), v = ao(d, "attributes"), y = o && o.prototype ? ao(o.prototype, "nodeType") : null, b = o && o.prototype ? ao(o.prototype, "nodeName") : null;
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let x, S = "", C, w = !1, T = 0, E = function() {
		if (T > 0) throw eo("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
	}, D = function(e) {
		E(), T++;
		try {
			return x.createHTML(e);
		} finally {
			T--;
		}
	}, O = function(e) {
		E(), T++;
		try {
			return x.createScriptURL(e);
		} finally {
			T--;
		}
	}, k = function() {
		return w ||= (C = Fo(u, i), !0), C;
	}, A = n, j = A.implementation, M = A.createNodeIterator, N = A.createDocumentFragment, P = A.getElementsByTagName, F = r.importNode, I = Io();
	t.isSupported = typeof wa == "function" && typeof g == "function" && j && j.createHTMLDocument !== void 0;
	let L = yo, ee = bo, R = xo, z = So, te = Co, B = To, V = Eo, H = Oo, ne = wo, U = null, re = Q({}, [
		...so,
		...co,
		...lo,
		...fo,
		...mo
	]), W = null, ie = Q({}, [
		...ho,
		...go,
		..._o,
		...vo
	]), G = Object.seal(ja(null, {
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
	})), ae = null, oe = null, se = Object.seal(ja(null, {
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
	})), ce = !0, le = !0, ue = !1, de = !0, fe = !1, pe = !0, me = !1, he = !1, ge = null, _e = null, ve = !1, ye = !1, be = !1, xe = !1, Se = !0, Ce = !1, we = "user-content-", Te = !0, Ee = !1, De = {}, Oe = null, ke = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Ae = null, je = Q({}, [
		"audio",
		"video",
		"img",
		"source",
		"image",
		"track"
	]), Me = null, Ne = Q({}, [
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
	]), Pe = "http://www.w3.org/1998/Math/MathML", Fe = "http://www.w3.org/2000/svg", Ie = "http://www.w3.org/1999/xhtml", Le = Ie, Re = !1, ze = null, Be = Q({}, [
		Pe,
		Fe,
		Ie
	], Ha), Ve = ka([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), He = Q({}, Ve), Ue = ka(["annotation-xml"]), We = Q({}, Ue), Ge = Q({}, [
		"title",
		"style",
		"font",
		"a",
		"script"
	]), Ke = null, qe = ["application/xhtml+xml", "text/html"], K = null, Je = null, Ye = n.createElement("form"), Xe = function(e) {
		return e instanceof RegExp || e instanceof Function;
	}, Ze = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		if (Je && Je === e) return;
		e && typeof e == "object" || (e = {}), e = ro(e), Ke = qe.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, K = Ke === "application/xhtml+xml" ? Ha : Va, U = Lo(e, "ALLOWED_TAGS", re, { transform: K }), W = Lo(e, "ALLOWED_ATTR", ie, { transform: K }), ze = Lo(e, "ALLOWED_NAMESPACES", Be, { transform: Ha }), Me = Lo(e, "ADD_URI_SAFE_ATTR", Ne, {
			transform: K,
			base: Ne
		}), Ae = Lo(e, "ADD_DATA_URI_TAGS", je, {
			transform: K,
			base: je
		}), Oe = Lo(e, "FORBID_CONTENTS", ke, { transform: K }), ae = Lo(e, "FORBID_TAGS", ro({}), { transform: K }), oe = Lo(e, "FORBID_ATTR", ro({}), { transform: K }), De = Za(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? ro(e.USE_PROFILES) : e.USE_PROFILES : !1, ce = e.ALLOW_ARIA_ATTR !== !1, le = e.ALLOW_DATA_ATTR !== !1, ue = e.ALLOW_UNKNOWN_PROTOCOLS || !1, de = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, fe = e.SAFE_FOR_TEMPLATES || !1, pe = e.SAFE_FOR_XML !== !1, me = e.WHOLE_DOCUMENT || !1, ye = e.RETURN_DOM || !1, be = e.RETURN_DOM_FRAGMENT || !1, xe = e.RETURN_TRUSTED_TYPE || !1, ve = e.FORCE_BODY || !1, Se = e.SANITIZE_DOM !== !1, Ce = e.SANITIZE_NAMED_PROPS || !1, Te = e.KEEP_CONTENT !== !1, Ee = e.IN_PLACE || !1, ne = oo(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : wo, Le = typeof e.NAMESPACE == "string" ? e.NAMESPACE : Ie, He = Za(e, "MATHML_TEXT_INTEGRATION_POINTS") && e.MATHML_TEXT_INTEGRATION_POINTS && typeof e.MATHML_TEXT_INTEGRATION_POINTS == "object" ? ro(e.MATHML_TEXT_INTEGRATION_POINTS) : Q({}, Ve), We = Za(e, "HTML_INTEGRATION_POINTS") && e.HTML_INTEGRATION_POINTS && typeof e.HTML_INTEGRATION_POINTS == "object" ? ro(e.HTML_INTEGRATION_POINTS) : Q({}, Ue);
		let t = Za(e, "CUSTOM_ELEMENT_HANDLING") && e.CUSTOM_ELEMENT_HANDLING && typeof e.CUSTOM_ELEMENT_HANDLING == "object" ? ro(e.CUSTOM_ELEMENT_HANDLING) : ja(null);
		if (G = ja(null), Za(t, "tagNameCheck") && Xe(t.tagNameCheck) && (G.tagNameCheck = t.tagNameCheck), Za(t, "attributeNameCheck") && Xe(t.attributeNameCheck) && (G.attributeNameCheck = t.attributeNameCheck), Za(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (G.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), Aa(G), fe && (le = !1), be && (ye = !0), De && (U = Q({}, mo), W = ja(null), De.html === !0 && (Q(U, so), Q(W, ho)), De.svg === !0 && (Q(U, co), Q(W, go), Q(W, vo)), De.svgFilters === !0 && (Q(U, lo), Q(W, go), Q(W, vo)), De.mathMl === !0 && (Q(U, fo), Q(W, _o), Q(W, vo))), se.tagCheck = null, se.attributeCheck = null, Za(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? se.tagCheck = e.ADD_TAGS : Ba(e.ADD_TAGS) && (U === re && (U = ro(U)), Q(U, e.ADD_TAGS, K))), Za(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? se.attributeCheck = e.ADD_ATTR : Ba(e.ADD_ATTR) && (W === ie && (W = ro(W)), Q(W, e.ADD_ATTR, K))), Za(e, "ADD_URI_SAFE_ATTR") && Ba(e.ADD_URI_SAFE_ATTR) && Q(Me, e.ADD_URI_SAFE_ATTR, K), Za(e, "FORBID_CONTENTS") && Ba(e.FORBID_CONTENTS) && (Oe === ke && (Oe = ro(Oe)), Q(Oe, e.FORBID_CONTENTS, K)), Za(e, "ADD_FORBID_CONTENTS") && Ba(e.ADD_FORBID_CONTENTS) && (Oe === ke && (Oe = ro(Oe)), Q(Oe, e.ADD_FORBID_CONTENTS, K)), Te && (U["#text"] = !0), me && Q(U, [
			"html",
			"head",
			"body"
		]), U.table && (Q(U, ["tbody"]), delete ae.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw eo("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw eo("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = x;
			x = e.TRUSTED_TYPES_POLICY;
			try {
				S = D("");
			} catch (e) {
				throw x = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (x = void 0, S = "") : (x === void 0 && (x = k()), x && typeof S == "string" && (S = D("")));
		ka && ka(e), Je = e;
	}, Qe = Q({}, [
		...co,
		...lo,
		...uo
	]), $e = Q({}, [...fo, ...po]), et = function(e, t, n) {
		return t.namespaceURI === Ie ? e === "svg" : t.namespaceURI === Pe ? e === "svg" && (n === "annotation-xml" || He[n]) : !!Qe[e];
	}, tt = function(e, t, n) {
		return t.namespaceURI === Ie ? e === "math" : t.namespaceURI === Fe ? e === "math" && We[n] : !!$e[e];
	}, nt = function(e, t, n) {
		return t.namespaceURI === Fe && !We[n] || t.namespaceURI === Pe && !He[n] ? !1 : !$e[e] && (Ge[e] || !Qe[e]);
	}, rt = function(e) {
		let t = g(e);
		t && t.tagName || (t = {
			namespaceURI: Le,
			tagName: "template"
		});
		let n = Va(e.tagName), r = Va(t.tagName);
		return ze[e.namespaceURI] ? e.namespaceURI === Fe ? et(n, t, r) : e.namespaceURI === Pe ? tt(n, t, r) : e.namespaceURI === Ie ? nt(n, t, r) : !!(Ke === "application/xhtml+xml" && ze[e.namespaceURI]) : !1;
	}, it = function(e) {
		Ra(t.removed, { element: e });
		try {
			g(e).removeChild(e);
		} catch {
			if (p(e), !g(e)) throw eo("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
		}
	}, at = function(e) {
		let t = h(e);
		if (t) {
			let e = [];
			Fa(t, (t) => {
				Ra(e, t);
			}), Fa(e, (e) => {
				try {
					p(e);
				} catch {}
			});
		}
		let n = v(e);
		if (n) for (let t = n.length - 1; t >= 0; --t) {
			let r = n[t], i = r && r.name;
			if (typeof i == "string") try {
				e.removeAttribute(i);
			} catch {}
		}
	}, ot = function(e, n) {
		try {
			Ra(t.removed, {
				attribute: n.getAttributeNode(e),
				from: n
			});
		} catch {
			Ra(t.removed, {
				attribute: null,
				from: n
			});
		}
		if (n.removeAttribute(e), e === "is") {
			if (ye || be) try {
				it(n);
			} catch {}
			else try {
				n.setAttribute(e, "");
			} catch {}
		}
	}, st = function(e) {
		let t = v(e);
		if (t) for (let n = t.length - 1; n >= 0; --n) {
			let r = t[n], i = r && r.name;
			if (!(typeof i != "string" || W[K(i)])) try {
				e.removeAttribute(i);
			} catch {}
		}
	}, ct = function(e) {
		let t = [e];
		for (; t.length > 0;) {
			let e = t.pop();
			(y ? y(e) : e.nodeType) === No.element && st(e);
			let n = h(e);
			if (n) for (let e = n.length - 1; e >= 0; --e) t.push(n[e]);
		}
	}, lt = function(e) {
		let t = null, r = null;
		if (ve) e = "<remove></remove>" + e;
		else {
			let t = Ua(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		Ke === "application/xhtml+xml" && Le === Ie && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = x ? D(e) : e;
		if (Le === Ie) try {
			t = new l().parseFromString(i, Ke);
		} catch {}
		if (!t || !t.documentElement) {
			t = j.createDocument(Le, "template", null);
			try {
				t.documentElement.innerHTML = Re ? S : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), Le === Ie ? P.call(t, me ? "html" : "body")[0] : me ? t.documentElement : a;
	}, ut = function(e) {
		return M.call(e.ownerDocument || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, dt = function(e) {
		return e = Wa(e, L, " "), e = Wa(e, ee, " "), e = Wa(e, R, " "), e;
	}, ft = function(e) {
		e.normalize();
		let t = M.call(e.ownerDocument || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), n = t.nextNode();
		for (; n;) n.data = dt(n.data), n = t.nextNode();
		let r = e.querySelectorAll?.call(e, "template");
		r && Fa(r, (e) => {
			mt(e.content) && ft(e.content);
		});
	}, pt = function(e) {
		let t = b ? b(e) : null;
		return typeof t != "string" || K(t) !== "form" ? !1 : typeof e.nodeName != "string" || typeof e.textContent != "string" || typeof e.removeChild != "function" || e.attributes !== v(e) || typeof e.removeAttribute != "function" || typeof e.setAttribute != "function" || typeof e.namespaceURI != "string" || typeof e.insertBefore != "function" || typeof e.hasChildNodes != "function" || e.nodeType !== y(e) || e.childNodes !== h(e);
	}, mt = function(e) {
		if (!y || typeof e != "object" || !e) return !1;
		try {
			return y(e) === No.documentFragment;
		} catch {
			return !1;
		}
	}, ht = function(e) {
		if (!y || typeof e != "object" || !e) return !1;
		try {
			return typeof y(e) == "number";
		} catch {
			return !1;
		}
	};
	function gt(e, n, r) {
		e.length !== 0 && Fa(e, (e) => {
			e.call(t, n, r, Je);
		});
	}
	let _t = function(e, t) {
		return !!(pe && e.hasChildNodes() && !ht(e.firstElementChild) && $a(ko, e.textContent) && $a(ko, e.innerHTML) || pe && e.namespaceURI === Ie && t === "style" && ht(e.firstElementChild) || e.nodeType === No.processingInstruction || pe && e.nodeType === No.comment && $a(Ao, e.data));
	}, vt = function(e, t) {
		if (!ae[t] && St(t) && (G.tagNameCheck instanceof RegExp && $a(G.tagNameCheck, t) || G.tagNameCheck instanceof Function && G.tagNameCheck(t))) return !1;
		if (Te && !Oe[t]) {
			let t = g(e), n = h(e);
			if (n && t) {
				let r = n.length;
				for (let i = r - 1; i >= 0; --i) {
					let r = Ee ? n[i] : f(n[i], !0);
					t.insertBefore(r, m(e));
				}
			}
		}
		return it(e), !0;
	}, yt = function(e) {
		if (gt(I.beforeSanitizeElements, e, null), pt(e)) return it(e), !0;
		let n = K(b ? b(e) : e.nodeName);
		if (gt(I.uponSanitizeElement, e, {
			tagName: n,
			allowedTags: U
		}), _t(e, n)) return it(e), !0;
		if (ae[n] || !(se.tagCheck instanceof Function && se.tagCheck(n)) && !U[n]) return vt(e, n);
		if ((y ? y(e) : e.nodeType) === No.element && !rt(e) || (n === "noscript" || n === "noembed" || n === "noframes") && $a(jo, e.innerHTML)) return it(e), !0;
		if (fe && e.nodeType === No.text) {
			let n = dt(e.textContent);
			e.textContent !== n && (Ra(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return gt(I.afterSanitizeElements, e, null), !1;
	}, bt = function(e, t, r) {
		if (oe[t] || Se && (t === "id" || t === "name") && (r in n || r in Ye)) return !1;
		let i = W[t] || se.attributeCheck instanceof Function && se.attributeCheck(t, e);
		if (!(le && $a(z, t)) && !(ce && $a(te, t))) {
			if (!i) {
				if (!(St(e) && (G.tagNameCheck instanceof RegExp && $a(G.tagNameCheck, e) || G.tagNameCheck instanceof Function && G.tagNameCheck(e)) && (G.attributeNameCheck instanceof RegExp && $a(G.attributeNameCheck, t) || G.attributeNameCheck instanceof Function && G.attributeNameCheck(t, e)) || t === "is" && G.allowCustomizedBuiltInElements && (G.tagNameCheck instanceof RegExp && $a(G.tagNameCheck, r) || G.tagNameCheck instanceof Function && G.tagNameCheck(r)))) return !1;
			} else if (!Me[t] && !$a(ne, Wa(r, V, "")) && (t !== "src" && t !== "xlink:href" && t !== "href" || e === "script" || Ga(r, "data:") !== 0 || !Ae[e]) && (!ue || $a(B, Wa(r, V, ""))) && r) return !1;
		}
		return !0;
	}, xt = Q({}, [
		"annotation-xml",
		"color-profile",
		"font-face",
		"font-face-format",
		"font-face-name",
		"font-face-src",
		"font-face-uri",
		"missing-glyph"
	]), St = function(e) {
		return !xt[Va(e)] && $a(H, e);
	}, Ct = function(e, t, n, r) {
		if (x && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return D(r);
			case "TrustedScriptURL": return O(r);
		}
		return r;
	}, wt = function(e, n, r, i) {
		try {
			r ? e.setAttributeNS(r, n, i) : e.setAttribute(n, i), pt(e) ? it(e) : La(t.removed);
		} catch {
			ot(n, e);
		}
	}, Tt = function(e) {
		gt(I.beforeSanitizeAttributes, e, null);
		let t = e.attributes;
		if (!t || pt(e)) return;
		let n = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: W,
			forceKeepAttr: void 0
		}, r = t.length, i = K(e.nodeName);
		for (; r--;) {
			let a = t[r], o = a.name, s = a.namespaceURI, c = a.value, l = K(o), u = c, d = o === "value" ? u : Ka(u);
			n.attrName = l, n.attrValue = d, n.keepAttr = !0, n.forceKeepAttr = void 0, gt(I.uponSanitizeAttribute, e, n), d = n.attrValue, Ce && (l === "id" || l === "name") && Ga(d, we) !== 0 && (ot(o, e), d = we + d), pe && $a(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, d) || l === "attributename" && Ua(d, "href") ? ot(o, e) : n.forceKeepAttr || (n.keepAttr ? !de && $a(Mo, d) ? ot(o, e) : (fe && (d = dt(d)), bt(i, l, d) ? (d = Ct(i, l, s, d), d !== u && wt(e, o, s, d)) : ot(o, e)) : ot(o, e));
		}
		gt(I.afterSanitizeAttributes, e, null);
	}, Et = function(e) {
		let t = null, n = ut(e);
		for (gt(I.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (gt(I.uponSanitizeShadowNode, t, null), yt(t), Tt(t), mt(t.content) && Et(t.content), (y ? y(t) : t.nodeType) === No.element) {
			let e = _(t);
			mt(e) && (Dt(e), Et(e));
		}
		gt(I.afterSanitizeShadowDOM, e, null);
	}, Dt = function(e) {
		let t = [{
			node: e,
			shadow: null
		}];
		for (; t.length > 0;) {
			let e = t.pop();
			if (e.shadow) {
				Et(e.shadow);
				continue;
			}
			let n = e.node, r = (y ? y(n) : n.nodeType) === No.element, i = h(n);
			if (i) for (let e = i.length - 1; e >= 0; --e) t.push({
				node: i[e],
				shadow: null
			});
			if (r) {
				let e = b ? b(n) : null;
				if (typeof e == "string" && K(e) === "template") {
					let e = n.content;
					mt(e) && t.push({
						node: e,
						shadow: null
					});
				}
			}
			if (r) {
				let e = _(n);
				mt(e) && t.push({
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
		if (Re = !e, Re && (e = "<!-->"), typeof e != "string" && !ht(e) && (e = io(e), typeof e != "string")) throw eo("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		he ? (U = ge, W = _e) : Ze(n), (I.uponSanitizeElement.length > 0 || I.uponSanitizeAttribute.length > 0) && (U = ro(U)), I.uponSanitizeAttribute.length > 0 && (W = ro(W)), t.removed = [];
		let c = Ee && typeof e != "string" && ht(e);
		if (c) {
			let t = b ? b(e) : e.nodeName;
			if (typeof t == "string") {
				let e = K(t);
				if (!U[e] || ae[e]) throw eo("root node is forbidden and cannot be sanitized in-place");
			}
			if (pt(e)) throw eo("root node is clobbered and cannot be sanitized in-place");
			try {
				Dt(e);
			} catch (t) {
				throw at(e), t;
			}
		} else if (ht(e)) i = lt("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === No.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), Dt(a);
		else {
			if (!ye && !fe && !me && e.indexOf("<") === -1) return x && xe ? D(e) : e;
			if (i = lt(e), !i) return ye ? null : xe ? S : "";
		}
		i && ve && it(i.firstChild);
		let l = ut(c ? e : i);
		try {
			for (; o = l.nextNode();) yt(o), Tt(o), mt(o.content) && Et(o.content);
		} catch (t) {
			throw c && at(e), t;
		}
		if (c) return Fa(t.removed, (e) => {
			e.element && ct(e.element);
		}), fe && ft(e), e;
		if (ye) {
			if (fe && ft(i), be) for (s = N.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (W.shadowroot || W.shadowrootmode) && (s = F.call(r, s, !0)), s;
		}
		let u = me ? i.outerHTML : i.innerHTML;
		return me && U["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && $a(Do, i.ownerDocument.doctype.name) && (u = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + u), fe && (u = dt(u)), x && xe ? D(u) : u;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		Ze(e), he = !0, ge = U, _e = W;
	}, t.clearConfig = function() {
		Je = null, he = !1, ge = null, _e = null, x = C, S = "";
	}, t.isValidAttribute = function(e, t, n) {
		Je || Ze({});
		let r = K(e), i = K(t);
		return bt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && Za(I, e) && Ra(I[e], t);
	}, t.removeHook = function(e, t) {
		if (Za(I, e)) {
			if (t !== void 0) {
				let n = Ia(I[e], t);
				return n === -1 ? void 0 : za(I[e], n, 1)[0];
			}
			return La(I[e]);
		}
	}, t.removeHooks = function(e) {
		Za(I, e) && (I[e] = []);
	}, t.removeAllHooks = function() {
		I = Io();
	}, t;
}
var zo = Ro();
// ../../../clicky-ui/packages/ui/src/data/html-utils.ts
function Bo(e) {
	return !e || typeof window > "u" ? "" : zo.sanitize(e, { USE_PROFILES: { html: !0 } });
}
function Vo(e) {
	return /<(div|p|pre|table|ul|ol|li|details|blockquote|h[1-6])/i.test(e);
}
// ../../../clicky-ui/packages/ui/src/data/CodeBlock.tsx
var Ho = 500, Uo = 2e3;
function Wo({ language: e, source: t = "", highlightedHtml: n, className: r, jsonDefaultOpenDepth: a = 2, jsonFormat: o = "json", bare: s = !1, copyable: c = !1, downloadable: l = !1, downloadFilename: u, themeToggle: d = !1, diff: f, diffView: p, inline: m = !1 }) {
	let h = (e ?? "").toLowerCase().replace(/^\.+/, ""), g = n ? Bo(n) : "", _ = (0, M.useMemo)(() => h === "json" ? qo(t) : Ko, [h, t]), v = We(), [y, b] = (0, M.useState)(null), x = y ?? v, [S, C] = (0, M.useState)(null), w = _ === Ko && !g && !!h && !!t, T = (0, M.useMemo)(() => m ? t.replace(/\s+/g, " ").trim() : "", [m, t]), E = m ? T.slice(0, Ho) : t;
	(0, M.useEffect)(() => {
		if (!w) {
			C(null);
			return;
		}
		let e = !1;
		return Bt(E, {
			lang: h,
			theme: x === "dark" ? "github-dark" : "github-light"
		}).then((t) => {
			e || C(t);
		}), () => {
			e = !0;
		};
	}, [
		w,
		h,
		E,
		x
	]);
	let D = /* @__PURE__ */ (0, N.jsx)(mt, {
		source: t,
		language: e,
		copyable: c,
		downloadable: l,
		downloadFilename: u,
		theme: x,
		...d ? { onToggleTheme: () => b(x === "dark" ? "light" : "dark") } : {}
	}), O = y ? { "data-theme": y } : {};
	if (m) {
		let e = g || S, t = T.length > Uo ? `${T.slice(0, Uo)}…` : T;
		return /* @__PURE__ */ (0, N.jsx)("code", {
			className: i("block max-w-full truncate font-mono text-xs text-foreground", r, e && "[&_pre]:inline [&_pre]:whitespace-pre [&_.chroma]:bg-transparent"),
			title: t,
			children: e ? /* @__PURE__ */ (0, N.jsx)("span", { dangerouslySetInnerHTML: { __html: e } }) : E
		});
	}
	if (f) return /* @__PURE__ */ (0, N.jsx)(Jt, {
		language: e,
		view: p,
		bare: s,
		className: r,
		...f
	});
	if (_ !== Ko) {
		let e = /* @__PURE__ */ (0, N.jsx)(ma, {
			data: _,
			format: o,
			defaultOpenDepth: a
		});
		return s ? /* @__PURE__ */ (0, N.jsx)("div", {
			className: i("overflow-auto text-xs", r),
			children: e
		}) : /* @__PURE__ */ (0, N.jsxs)("div", {
			className: i("overflow-hidden rounded-md border border-border bg-muted/40", r),
			...O,
			children: [/* @__PURE__ */ (0, N.jsx)(Go, {
				label: o,
				actions: D
			}), /* @__PURE__ */ (0, N.jsx)("div", {
				className: "overflow-auto p-3 text-xs",
				children: e
			})]
		});
	}
	let k = h === "xml" || h === "xslt" || h === "html" ? " [&_.chroma_.nt]:text-pink-700 [&_.chroma_.nt]:dark:text-pink-300 [&_.chroma_.nt]:font-semibold [&_.chroma_.na]:text-amber-700 [&_.chroma_.na]:dark:text-amber-300 [&_.chroma_.s]:text-emerald-700 [&_.chroma_.s]:dark:text-emerald-300 [&_.chroma_.c]:text-slate-500 [&_.chroma_.c]:italic [&_.chroma_.cp]:text-slate-500 [&_.chroma_.cp]:italic [&_pre]:leading-relaxed" : "", A = s ? void 0 : "p-3", j = g ? /* @__PURE__ */ (0, N.jsx)("div", {
		className: i("overflow-auto text-xs font-mono [&_.chroma]:bg-transparent [&_pre]:m-0 [&_pre]:whitespace-pre-wrap [&_pre]:bg-transparent" + k, A, s && r),
		dangerouslySetInnerHTML: { __html: g }
	}) : S ? /* @__PURE__ */ (0, N.jsx)("div", {
		className: i("overflow-auto text-xs font-mono [&_pre]:m-0 [&_pre.shiki]:!bg-transparent [&_pre]:whitespace-pre-wrap", A, s && r),
		dangerouslySetInnerHTML: { __html: S }
	}) : /* @__PURE__ */ (0, N.jsx)("pre", {
		className: i("overflow-auto whitespace-pre-wrap break-words text-xs font-mono text-foreground", A, s && r),
		children: t
	});
	return s ? j : /* @__PURE__ */ (0, N.jsxs)("div", {
		className: i("overflow-hidden rounded-md border border-border bg-muted/40", r),
		...O,
		children: [/* @__PURE__ */ (0, N.jsx)(Go, {
			label: e || "text",
			actions: D
		}), j]
	});
}
function Go({ label: e, actions: t }) {
	return /* @__PURE__ */ (0, N.jsxs)("div", {
		className: "flex items-center justify-between gap-2 border-b border-border px-3 py-1.5",
		children: [/* @__PURE__ */ (0, N.jsx)("span", {
			className: "text-[11px] uppercase tracking-wide text-muted-foreground",
			children: e
		}), t]
	});
}
var Ko = Symbol("json-parse-failed");
function qo(e) {
	let t = e.trim();
	if (!t) return Ko;
	try {
		return JSON.parse(t);
	} catch {
		return Ko;
	}
}
// ../../../clicky-ui/packages/ui/src/rpc/router.tsx
var Jo = (0, M.createContext)(null);
function Yo(e) {
	return !e.defaultPrevented && e.button === 0 && !e.metaKey && !e.ctrlKey && !e.shiftKey && !e.altKey && e.currentTarget.target !== "_blank";
}
function Xo(e) {
	return window.addEventListener("popstate", e), () => window.removeEventListener("popstate", e);
}
function Zo() {
	let e = (0, M.useSyncExternalStore)(Xo, () => `${window.location.pathname}${window.location.search}${window.location.hash}`, () => "/").split(/[?#]/, 1)[0] || "/", t = (0, M.useCallback)((e, t) => {
		typeof window < "u" && (t?.replace ? window.history.replaceState(window.history.state, "", e) : window.history.pushState(window.history.state, "", e), window.dispatchEvent(new PopStateEvent("popstate")));
	}, []), n = (0, M.useCallback)(({ to: e, className: n, children: r, title: i, key: a, draggable: o, onDragStart: s }) => /* @__PURE__ */ (0, N.jsx)("a", {
		href: e,
		className: n,
		title: i,
		draggable: o,
		onDragStart: s,
		onClick: (n) => {
			Yo(n) && (n.preventDefault(), t(e));
		},
		children: r
	}, a), [t]);
	return (0, M.useMemo)(() => ({
		pathname: e,
		renderLink: n,
		navigate: t
	}), [
		e,
		n,
		t
	]);
}
function Qo() {
	let e = (0, M.useContext)(Jo), t = Zo();
	return e ?? t;
}
// ../../../clicky-ui/packages/ui/src/lib/breakpoints.ts
var $o = /* @__PURE__ */ e(r(), 1), es = {
	sm: 640,
	md: 768,
	lg: 1024,
	xl: 1280,
	"2xl": 1536,
	"3xl": 1920
};
function ts(e) {
	return `(max-width: ${es[e] - 1}px)`;
}
// ../../../clicky-ui/packages/ui/src/overlay/zIndex.ts
var ns = {
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
}, rs = [], is = /* @__PURE__ */ new Set(), as = [], os = null, ss = 0;
function cs() {
	for (let e of is) e();
}
function ls(e) {
	return is.add(e), () => {
		is.delete(e);
	};
}
function us() {
	return `${ss}:` + rs.join("\0");
}
function ds(e) {
	if (e.defaultPrevented || e.key !== "Escape" || e.isComposing || e.keyCode === 229) return;
	let t = as.findLast((e) => !e.dismissed);
	t && (e.preventDefault(), e.stopPropagation(), e.stopImmediatePropagation(), t.dismissed = !0, t.onEscape());
}
function fs() {
	os || typeof document > "u" || (os = document, os.addEventListener("keydown", ds));
}
function ps(e) {
	let t = as.findIndex((t) => t.id === e);
	t !== -1 && as.splice(t, 1), as.length === 0 && os && (os.removeEventListener("keydown", ds), os = null);
}
function ms(e) {
	let t = (0, M.useId)();
	(0, M.useSyncExternalStore)(ls, us, us), (0, M.useEffect)(() => {
		if (e) return rs.push(t), cs(), () => {
			let e = rs.indexOf(t);
			e !== -1 && rs.splice(e, 1), cs();
		};
	}, [e, t]);
	let n = rs.indexOf(t);
	return {
		isTop: n !== -1 && n === rs.length - 1,
		depth: n === -1 ? 0 : n
	};
}
function hs(e, t, n = !0) {
	let r = (0, M.useId)(), i = (0, M.useRef)(t);
	i.current = t;
	let a = (0, M.useRef)(null);
	(0, M.useEffect)(() => {
		a.current && (a.current.dismissed = !1);
	}), (0, M.useEffect)(() => {
		if (!e || !n) return;
		fs();
		let t = {
			id: r,
			dismissed: !1,
			onEscape: () => i.current()
		};
		return a.current = t, as.push(t), () => {
			a.current = null, ps(r);
		};
	}, [
		n,
		r,
		e
	]);
}
function gs() {
	(0, M.useSyncExternalStore)(ls, us, us);
	let e = rs.length === 0 ? ns.popover : ns.modal + (rs.length - 1) * ns.modalStep + ns.popoverOverModalOffset;
	return ss === 0 ? e : Math.max(e, ns.tour + ns.tourCardOffset + ns.popoverOverModalOffset);
}
// ../../../clicky-ui/packages/ui/src/overlay/Modal.tsx
var _s = {
	sm: "max-w-sm",
	md: "max-w-md",
	lg: "max-w-2xl",
	xl: "max-w-4xl",
	"2xl": "max-w-6xl",
	full: "max-w-[95vw]"
}, vs = {
	title: "Discard changes?",
	message: "You have unsaved changes. Closing now will discard them.",
	confirmLabel: "Discard",
	cancelLabel: "Keep editing"
}, ys = ts("sm"), bs = "(prefers-reduced-motion: reduce)", xs = 200, Ss = "a[href],button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),[tabindex]:not([tabindex='-1'])";
function Cs(e) {
	return Array.from(e.querySelectorAll(Ss)).filter((e) => {
		let t = window.getComputedStyle(e);
		return !e.hidden && t.display !== "none" && t.visibility !== "hidden";
	});
}
function ws({ open: e, onClose: t, confirmClose: n = !1, title: r, subtitle: o, size: s = "md", closeOnBackdrop: c = !1, closeOnEsc: l = !0, hideClose: u = !1, expandable: d = !0, className: f, headerSlot: p, footer: m, scrollBody: h = !0, children: g }) {
	let _ = (0, M.useRef)(null), v = (0, M.useRef)(void 0), y = (0, M.useRef)(void 0), [b, x] = (0, M.useState)(e), [S, C] = (0, M.useState)(!1), { depth: w } = ms(b), [T, E] = (0, M.useState)(!1), [D, O] = (0, M.useState)(!1), k = () => {
		n ? O(!0) : t();
	};
	if ((0, M.useEffect)(() => {
		window.clearTimeout(v.current), window.cancelAnimationFrame(y.current ?? 0);
		let t = Ie(ys) && !Ie(bs);
		return e ? (x(!0), t ? (C(!1), y.current = window.requestAnimationFrame(() => {
			C(!0);
		})) : C(!0)) : t ? (C(!1), v.current = window.setTimeout(() => x(!1), xs)) : (C(!1), x(!1)), () => {
			window.clearTimeout(v.current), window.cancelAnimationFrame(y.current ?? 0);
		};
	}, [e]), (0, M.useEffect)(() => {
		e || (E(!1), O(!1));
	}, [e]), (0, M.useEffect)(() => {
		n || O(!1);
	}, [n]), hs(b, () => {
		e && l && (D ? O(!1) : k());
	}), (0, M.useEffect)(() => {
		if (!e) return;
		let t = document.activeElement;
		return _.current?.focus(), () => t?.focus?.();
	}, [e]), !b) return null;
	let A = /* @__PURE__ */ (0, N.jsx)("div", {
		className: i("fixed inset-0 flex items-center justify-center p-density-2 sm:p-density-4 max-sm:items-stretch max-sm:justify-end max-sm:p-0 max-sm:!bg-transparent", w === 0 ? "bg-black/40" : "bg-black/20"),
		style: { zIndex: ns.modal + w * ns.modalStep },
		onClick: c ? k : void 0,
		role: "presentation",
		children: /* @__PURE__ */ (0, N.jsxs)("div", {
			ref: _,
			tabIndex: -1,
			role: "dialog",
			"aria-modal": "true",
			"aria-label": typeof r == "string" ? r : void 0,
			"data-state": S ? "open" : "closed",
			className: i("relative flex w-full flex-col overflow-hidden rounded-lg border border-border bg-background shadow-xl", "max-sm:!h-dvh max-sm:!max-h-dvh max-sm:!max-w-none max-sm:!rounded-none max-sm:!border-0 max-sm:shadow-none", "max-sm:transition-transform max-sm:duration-200 max-sm:ease-out max-sm:motion-reduce:transition-none", "max-sm:data-[state=closed]:translate-x-full max-sm:data-[state=open]:translate-x-0", T ? _s.full : _s[s], f),
			style: {
				maxHeight: "calc(100dvh - 2rem)",
				...T ? { height: "calc(100dvh - 2rem)" } : {}
			},
			onClick: (e) => e.stopPropagation(),
			onKeyDown: (e) => {
				if (e.key !== "Tab") return;
				let t = D ? _.current?.querySelector("[role='alertdialog']") : _.current;
				if (!t) return;
				let n = Cs(t), r = n[0], i = n[n.length - 1];
				!r || !i ? (e.preventDefault(), t.focus()) : e.shiftKey && document.activeElement === r ? (e.preventDefault(), i.focus()) : !e.shiftKey && document.activeElement === i && (e.preventDefault(), r.focus());
			},
			children: [
				(r || o || p || d || !u) && /* @__PURE__ */ (0, N.jsxs)("div", {
					className: "shrink-0 px-density-4 py-density-3 border-b border-border max-sm:min-h-14 max-sm:bg-background",
					children: [/* @__PURE__ */ (0, N.jsxs)("div", {
						className: "flex items-center gap-density-2",
						children: [
							!u && /* @__PURE__ */ (0, N.jsx)(a, {
								type: "button",
								onClick: k,
								variant: "ghost",
								size: "icon",
								"aria-label": "Back",
								className: "-ml-2 text-foreground sm:hidden",
								children: /* @__PURE__ */ (0, N.jsx)(H, { icon: tt })
							}),
							r ? /* @__PURE__ */ (0, N.jsx)("h2", {
								className: "text-sm font-semibold flex-1 max-sm:text-base max-sm:leading-6",
								children: r
							}) : /* @__PURE__ */ (0, N.jsx)("span", { className: "flex-1" }),
							p,
							d && /* @__PURE__ */ (0, N.jsx)("button", {
								type: "button",
								onClick: () => E((e) => !e),
								"aria-label": T ? "Restore size" : "Expand to fullscreen",
								className: "text-muted-foreground hover:text-foreground max-sm:hidden",
								children: /* @__PURE__ */ (0, N.jsx)(H, { icon: T ? Xe : Ye })
							}),
							!u && /* @__PURE__ */ (0, N.jsx)("button", {
								type: "button",
								onClick: k,
								"aria-label": "Close",
								className: "text-muted-foreground hover:text-foreground max-sm:hidden",
								children: /* @__PURE__ */ (0, N.jsx)(H, { icon: K })
							})
						]
					}), o ? /* @__PURE__ */ (0, N.jsx)("div", {
						className: "mt-density-2",
						children: o
					}) : null]
				}),
				/* @__PURE__ */ (0, N.jsx)("div", {
					"data-slot": "modal-body",
					className: i("min-h-0 flex-1 px-density-4 py-density-3 max-sm:py-density-4", h ? "overflow-auto" : "flex flex-col overflow-hidden"),
					children: g
				}),
				m && /* @__PURE__ */ (0, N.jsx)("div", {
					"data-slot": "modal-footer",
					className: "shrink-0 px-density-4 py-density-3 border-t border-border max-sm:py-density-4",
					children: m
				}),
				D && /* @__PURE__ */ (0, N.jsx)(Ts, {
					options: n === !0 ? vs : {
						...vs,
						...n
					},
					onConfirm: () => {
						O(!1), t();
					},
					onCancel: () => O(!1)
				})
			]
		})
	});
	return typeof document < "u" ? (0, $o.createPortal)(A, document.body) : null;
}
function Ts({ options: e, onConfirm: t, onCancel: n }) {
	return /* @__PURE__ */ (0, N.jsx)("div", {
		className: "absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-background/80 p-density-4 max-sm:rounded-none",
		role: "alertdialog",
		"aria-modal": "true",
		"aria-label": typeof e.title == "string" ? e.title : "Confirm close",
		onClick: (e) => e.stopPropagation(),
		children: /* @__PURE__ */ (0, N.jsxs)("div", {
			className: "w-full max-w-sm rounded-lg border border-border bg-background p-density-4 shadow-xl",
			children: [
				/* @__PURE__ */ (0, N.jsx)("h3", {
					className: "text-sm font-semibold",
					children: e.title
				}),
				/* @__PURE__ */ (0, N.jsx)("p", {
					className: "mt-density-2 text-sm text-muted-foreground",
					children: e.message
				}),
				/* @__PURE__ */ (0, N.jsxs)("div", {
					className: "mt-density-4 flex justify-end gap-density-2",
					children: [/* @__PURE__ */ (0, N.jsx)(a, {
						variant: "ghost",
						size: "sm",
						onClick: n,
						children: e.cancelLabel
					}), /* @__PURE__ */ (0, N.jsx)(a, {
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
var Es = "input:not([inert]):not([inert] *),select:not([inert]):not([inert] *),textarea:not([inert]):not([inert] *),a[href]:not([inert]):not([inert] *),button:not([inert]):not([inert] *),[tabindex]:not(slot):not([inert]):not([inert] *),audio[controls]:not([inert]):not([inert] *),video[controls]:not([inert]):not([inert] *),[contenteditable]:not([contenteditable=\"false\"]):not([inert]):not([inert] *),details>summary:first-of-type:not([inert]):not([inert] *),details:not([inert]):not([inert] *)", Ds = typeof Element > "u", Os = Ds ? function() {} : Element.prototype.matches || Element.prototype.msMatchesSelector || Element.prototype.webkitMatchesSelector, ks = !Ds && Element.prototype.getRootNode ? function(e) {
	return e?.getRootNode?.call(e);
} : function(e) {
	return e?.ownerDocument;
}, As = function(e, t) {
	t === void 0 && (t = !0);
	var n = e?.getAttribute?.call(e, "inert");
	return n === "" || n === "true" || t && e && (typeof e.closest == "function" ? e.closest("[inert]") : As(e.parentNode));
}, js = function(e) {
	var t = e?.getAttribute?.call(e, "contenteditable");
	return t === "" || t === "true";
}, Ms = function(e, t, n) {
	if (As(e)) return [];
	var r = Array.prototype.slice.apply(e.querySelectorAll(Es));
	return t && Os.call(e, Es) && r.unshift(e), r = r.filter(n), r;
}, Ns = function(e, t, n) {
	for (var r = [], i = Array.from(e); i.length;) {
		var a = i.shift();
		if (!As(a, !1)) {
			if (a.tagName === "SLOT") {
				var o = a.assignedElements(), s = Ns(o.length ? o : a.children, !0, n);
				n.flatten ? r.push.apply(r, s) : r.push({
					scopeParent: a,
					candidates: s
				});
			} else {
				Os.call(a, Es) && n.filter(a) && (t || !e.includes(a)) && r.push(a);
				var c = a.shadowRoot || typeof n.getShadowRoot == "function" && n.getShadowRoot(a), l = !As(c, !1) && (!n.shadowRootFilter || n.shadowRootFilter(a));
				if (c && l) {
					var u = Ns(c === !0 ? a.children : c.children, !0, n);
					n.flatten ? r.push.apply(r, u) : r.push({
						scopeParent: a,
						candidates: u
					});
				} else i.unshift.apply(i, a.children);
			}
		}
	}
	return r;
}, Ps = function(e) {
	return !isNaN(parseInt(e.getAttribute("tabindex"), 10));
}, Fs = function(e) {
	if (!e) throw Error("No node provided");
	return e.tabIndex < 0 && (/^(AUDIO|VIDEO|DETAILS)$/.test(e.tagName) || js(e)) && !Ps(e) ? 0 : e.tabIndex;
}, Is = function(e, t) {
	var n = Fs(e);
	return n < 0 && t && !Ps(e) ? 0 : n;
}, Ls = function(e, t) {
	return e.tabIndex === t.tabIndex ? e.documentOrder - t.documentOrder : e.tabIndex - t.tabIndex;
}, Rs = function(e) {
	return e.tagName === "INPUT";
}, zs = function(e) {
	return Rs(e) && e.type === "hidden";
}, Bs = function(e) {
	return e.tagName === "DETAILS" && Array.prototype.slice.apply(e.children).some(function(e) {
		return e.tagName === "SUMMARY";
	});
}, Vs = function(e, t) {
	for (var n = 0; n < e.length; n++) if (e[n].checked && e[n].form === t) return e[n];
}, Hs = function(e) {
	if (!e.name) return !0;
	var t = e.form || ks(e), n = function(e) {
		return t.querySelectorAll("input[type=\"radio\"][name=\"" + e + "\"]");
	}, r;
	if (typeof window < "u" && window.CSS !== void 0 && typeof window.CSS.escape == "function") r = n(window.CSS.escape(e.name));
	else try {
		r = n(e.name);
	} catch (e) {
		return console.error("Looks like you have a radio button with a name attribute containing invalid CSS selector characters and need the CSS.escape polyfill: %s", e.message), !1;
	}
	var i = Vs(r, e.form);
	return !i || i === e;
}, Us = function(e) {
	return Rs(e) && e.type === "radio";
}, Ws = function(e) {
	return Us(e) && !Hs(e);
}, Gs = function(e) {
	var t = e && ks(e), n = t?.host, r = !1;
	if (t && t !== e) {
		var i, a, o;
		for (r = !!((i = n) != null && (a = i.ownerDocument) != null && a.contains(n) || e != null && (o = e.ownerDocument) != null && o.contains(e)); !r && n;) {
			var s, c;
			t = ks(n), n = t?.host, r = !!((s = n) != null && (c = s.ownerDocument) != null && c.contains(n));
		}
	}
	return r;
}, Ks = function(e) {
	var t = e.getBoundingClientRect(), n = t.width, r = t.height;
	return n === 0 && r === 0;
}, qs = function(e, t) {
	var n = t.displayCheck, r = t.getShadowRoot;
	if (n === "full-native" && "checkVisibility" in e) return !e.checkVisibility({
		checkOpacity: !1,
		opacityProperty: !1,
		contentVisibilityAuto: !0,
		visibilityProperty: !0,
		checkVisibilityCSS: !0
	});
	if (getComputedStyle(e).visibility === "hidden") return !0;
	var i = Os.call(e, "details>summary:first-of-type") ? e.parentElement : e;
	if (Os.call(i, "details:not([open]) *")) return !0;
	if (!n || n === "full" || n === "full-native" || n === "legacy-full") {
		if (typeof r == "function") {
			for (var a = e; e;) {
				var o = e.parentElement, s = ks(e);
				if (o && !o.shadowRoot && r(o) === !0) return Ks(e);
				e = e.assignedSlot ? e.assignedSlot : !o && s !== e.ownerDocument ? s.host : o;
			}
			e = a;
		}
		if (Gs(e)) return !e.getClientRects().length;
		if (n !== "legacy-full") return !0;
	} else if (n === "non-zero-area") return Ks(e);
	return !1;
}, Js = function(e) {
	if (/^(INPUT|BUTTON|SELECT|TEXTAREA)$/.test(e.tagName)) for (var t = e.parentElement; t;) {
		if (t.tagName === "FIELDSET" && t.disabled) {
			for (var n = 0; n < t.children.length; n++) {
				var r = t.children.item(n);
				if (r.tagName === "LEGEND") return Os.call(t, "fieldset[disabled] *") ? !0 : !r.contains(e);
			}
			return !0;
		}
		t = t.parentElement;
	}
	return !1;
}, Ys = function(e, t) {
	return !(t.disabled || zs(t) || qs(t, e) || Bs(t) || Js(t));
}, Xs = function(e, t) {
	return !(Ws(t) || Fs(t) < 0 || !Ys(e, t));
}, Zs = function(e) {
	var t = parseInt(e.getAttribute("tabindex"), 10);
	return !!(isNaN(t) || t >= 0);
}, Qs = function(e) {
	var t = [], n = [];
	return e.forEach(function(e, r) {
		var i = !!e.scopeParent, a = i ? e.scopeParent : e, o = Is(a, i), s = i ? Qs(e.candidates) : a;
		o === 0 ? i ? t.push.apply(t, s) : t.push(a) : n.push({
			documentOrder: r,
			tabIndex: o,
			item: e,
			isScope: i,
			content: s
		});
	}), n.sort(Ls).reduce(function(e, t) {
		return t.isScope ? e.push.apply(e, t.content) : e.push(t.content), e;
	}, []).concat(t);
}, $s = function(e, t) {
	return t ||= {}, Qs(t.getShadowRoot ? Ns([e], t.includeContainer, {
		filter: Xs.bind(null, t),
		flatten: !1,
		getShadowRoot: t.getShadowRoot,
		shadowRootFilter: Zs
	}) : Ms(e, t.includeContainer, Xs.bind(null, t)));
}, ec = function(e, t) {
	return t ||= {}, t.getShadowRoot ? Ns([e], t.includeContainer, {
		filter: Ys.bind(null, t),
		flatten: !0,
		getShadowRoot: t.getShadowRoot
	}) : Ms(e, t.includeContainer, Ys.bind(null, t));
}, tc = function(e, t) {
	if (t ||= {}, !e) throw Error("No node provided");
	return Os.call(e, Es) !== !1 && Xs(t, e);
};
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@floating-ui/react/0.27.19/01ef8dae599de5e9aab8bea825624a735e00051541336625c5d3fd3b24195bac/node_modules/@floating-ui/react/dist/floating-ui.react.utils.mjs
function nc() {
	let e = navigator.userAgentData;
	return e != null && e.platform ? e.platform : navigator.platform;
}
function rc() {
	let e = navigator.userAgentData;
	return e && Array.isArray(e.brands) ? e.brands.map((e) => {
		let { brand: t, version: n } = e;
		return t + "/" + n;
	}).join(" ") : navigator.userAgent;
}
function ic() {
	return /apple/i.test(navigator.vendor);
}
function ac() {
	let e = /android/i;
	return e.test(nc()) || e.test(rc());
}
function oc() {
	return rc().includes("jsdom/");
}
var sc = "data-floating-ui-focusable", cc = "input:not([type='hidden']):not([disabled]),[contenteditable]:not([contenteditable='false']),textarea:not([disabled])", lc = "ArrowLeft", uc = "ArrowRight", dc = "ArrowUp", fc = "ArrowDown";
function pc(e) {
	let t = e.activeElement;
	for (; ((n = t) == null || (n = n.shadowRoot) == null ? void 0 : n.activeElement) != null;) {
		var n;
		t = t.shadowRoot.activeElement;
	}
	return t;
}
function mc(e, t) {
	if (!e || !t) return !1;
	let n = t.getRootNode == null ? void 0 : t.getRootNode();
	if (e.contains(t)) return !0;
	if (n && j(n)) {
		let n = t;
		for (; n;) {
			if (e === n) return !0;
			n = n.parentNode || n.host;
		}
	}
	return !1;
}
function hc(e) {
	return "composedPath" in e ? e.composedPath()[0] : e.target;
}
function gc(e, t) {
	if (t == null) return !1;
	if ("composedPath" in e) return e.composedPath().includes(t);
	let n = e;
	return n.target != null && t.contains(n.target);
}
function _c(e) {
	return e.matches("html,body");
}
function vc(e) {
	return e?.ownerDocument || document;
}
function yc(e) {
	return x(e) && e.matches(cc);
}
function bc(e) {
	return e ? e.getAttribute("role") === "combobox" && yc(e) : !1;
}
function xc(e) {
	return e ? e.hasAttribute(sc) ? e : e.querySelector("[" + sc + "]") || e : null;
}
function Sc(e, t, n) {
	return n === void 0 && (n = !0), e.filter((e) => e.parentId === t && (!n || e.context?.open)).flatMap((t) => [t, ...Sc(e, t.id, n)]);
}
function Cc(e, t) {
	let n, r = -1;
	function i(t, a) {
		a > r && (n = t, r = a), Sc(e, t).forEach((e) => {
			i(e.id, a + 1);
		});
	}
	return i(t, 0), e.find((e) => e.id === n);
}
function wc(e, t) {
	let n = [], r = e.find((e) => e.id === t)?.parentId;
	for (; r;) {
		let t = e.find((e) => e.id === r);
		r = t?.parentId, t && (n = n.concat(t));
	}
	return n;
}
function Tc(e) {
	e.preventDefault(), e.stopPropagation();
}
function Ec(e) {
	return "nativeEvent" in e;
}
function Dc(e) {
	return e.mozInputSource === 0 && e.isTrusted ? !0 : ac() && e.pointerType ? e.type === "click" && e.buttons === 1 : e.detail === 0 && !e.pointerType;
}
function Oc(e) {
	return oc() ? !1 : !ac() && e.width === 0 && e.height === 0 || ac() && e.width === 1 && e.height === 1 && e.pressure === 0 && e.detail === 0 && e.pointerType === "mouse" || e.width < 1 && e.height < 1 && e.pressure === 0 && e.detail === 0 && e.pointerType === "touch";
}
function kc(e, t) {
	let n = ["mouse", "pen"];
	return t || n.push("", void 0), n.includes(e);
}
var $ = typeof document < "u" ? M.useLayoutEffect : function() {}, Ac = { ...M };
function jc(e) {
	let t = M.useRef(e);
	return $(() => {
		t.current = e;
	}), t;
}
var Mc = Ac.useInsertionEffect || ((e) => e());
function Nc(e) {
	let t = M.useRef(() => {});
	return Mc(() => {
		t.current = e;
	}), M.useCallback(function() {
		var e = [...arguments];
		return t.current == null ? void 0 : t.current(...e);
	}, []);
}
function Pc(e, t, n) {
	return Math.floor(e / t) !== n;
}
function Fc(e, t) {
	return t < 0 || t >= e.current.length;
}
function Ic(e, t) {
	return Rc(e, { disabledIndices: t });
}
function Lc(e, t) {
	return Rc(e, {
		decrement: !0,
		startingIndex: e.current.length,
		disabledIndices: t
	});
}
function Rc(e, t) {
	let { startingIndex: n = -1, decrement: r = !1, disabledIndices: i, amount: a = 1 } = t === void 0 ? {} : t, o = n;
	do
		o += r ? -a : a;
	while (o >= 0 && o <= e.current.length - 1 && Uc(e, o, i));
	return o;
}
function zc(e, t) {
	let { event: n, orientation: r, loop: i, rtl: a, cols: o, disabledIndices: s, minIndex: c, maxIndex: l, prevIndex: u, stopEvent: d = !1 } = t, f = u;
	if (n.key === dc) {
		if (d && Tc(n), u === -1) f = l;
		else if (f = Rc(e, {
			startingIndex: f,
			amount: o,
			decrement: !0,
			disabledIndices: s
		}), i && (u - o < c || f < 0)) {
			let e = u % o, t = l % o, n = l - (t - e);
			f = t === e ? l : t > e ? n : n - o;
		}
		Fc(e, f) && (f = u);
	}
	if (n.key === fc && (d && Tc(n), u === -1 ? f = c : (f = Rc(e, {
		startingIndex: u,
		amount: o,
		disabledIndices: s
	}), i && u + o > l && (f = Rc(e, {
		startingIndex: u % o - o,
		amount: o,
		disabledIndices: s
	}))), Fc(e, f) && (f = u)), r === "both") {
		let t = k(u / o);
		n.key === (a ? lc : uc) && (d && Tc(n), u % o === o - 1 ? i && (f = Rc(e, {
			startingIndex: u - u % o - 1,
			disabledIndices: s
		})) : (f = Rc(e, {
			startingIndex: u,
			disabledIndices: s
		}), i && Pc(f, o, t) && (f = Rc(e, {
			startingIndex: u - u % o - 1,
			disabledIndices: s
		}))), Pc(f, o, t) && (f = u)), n.key === (a ? uc : lc) && (d && Tc(n), u % o === 0 ? i && (f = Rc(e, {
			startingIndex: u + (o - u % o),
			decrement: !0,
			disabledIndices: s
		})) : (f = Rc(e, {
			startingIndex: u,
			decrement: !0,
			disabledIndices: s
		}), i && Pc(f, o, t) && (f = Rc(e, {
			startingIndex: u + (o - u % o),
			decrement: !0,
			disabledIndices: s
		}))), Pc(f, o, t) && (f = u));
		let r = k(l / o) === t;
		Fc(e, f) && (f = i && r ? n.key === (a ? uc : lc) ? l : Rc(e, {
			startingIndex: u - u % o - 1,
			disabledIndices: s
		}) : u);
	}
	return f;
}
function Bc(e, t, n) {
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
function Vc(e, t, n, r, i) {
	if (e === -1) return -1;
	let a = n.indexOf(e), o = t[e];
	switch (i) {
		case "tl": return a;
		case "tr": return o ? a + o.width - 1 : a;
		case "bl": return o ? a + (o.height - 1) * r : a;
		case "br": return n.lastIndexOf(e);
	}
}
function Hc(e, t) {
	return t.flatMap((t, n) => e.includes(t) ? [n] : []);
}
function Uc(e, t, n) {
	if (typeof n == "function") return n(t);
	if (n) return n.includes(t);
	let r = e.current[t];
	return r == null || r.hasAttribute("disabled") || r.getAttribute("aria-disabled") === "true";
}
var Wc = () => ({
	getShadowRoot: !0,
	displayCheck: typeof ResizeObserver == "function" && ResizeObserver.toString().includes("[native code]") ? "full" : "none"
});
function Gc(e, t) {
	let n = $s(e, Wc()), r = n.length;
	if (r === 0) return;
	let i = pc(vc(e)), a = n.indexOf(i);
	return n[a === -1 ? t === 1 ? 0 : r - 1 : a + t];
}
function Kc(e) {
	return Gc(vc(e).body, 1) || e;
}
function qc(e) {
	return Gc(vc(e).body, -1) || e;
}
function Jc(e, t) {
	let n = t || e.currentTarget, r = e.relatedTarget;
	return !r || !mc(n, r);
}
function Yc(e) {
	$s(e, Wc()).forEach((e) => {
		e.dataset.tabindex = e.getAttribute("tabindex") || "", e.setAttribute("tabindex", "-1");
	});
}
function Xc(e) {
	e.querySelectorAll("[data-tabindex]").forEach((e) => {
		let t = e.dataset.tabindex;
		delete e.dataset.tabindex, t ? e.setAttribute("tabindex", t) : e.removeAttribute("tabindex");
	});
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@floating-ui/react/0.27.19/01ef8dae599de5e9aab8bea825624a735e00051541336625c5d3fd3b24195bac/node_modules/@floating-ui/react/dist/floating-ui.react.mjs
function Zc(e) {
	let t = M.useRef(void 0), n = M.useCallback((t) => {
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
	return M.useMemo(() => e.every((e) => e == null) ? null : (e) => {
		t.current &&= (t.current(), void 0), e != null && (t.current = n(e));
	}, e);
}
function Qc(e, t) {
	let n = e.compareDocumentPosition(t);
	return n & Node.DOCUMENT_POSITION_FOLLOWING || n & Node.DOCUMENT_POSITION_CONTAINED_BY ? -1 : n & Node.DOCUMENT_POSITION_PRECEDING || n & Node.DOCUMENT_POSITION_CONTAINS ? 1 : 0;
}
var $c = /*#__PURE__*/ M.createContext({
	register: () => {},
	unregister: () => {},
	map: /*#__PURE__*/ new Map(),
	elementsRef: { current: [] }
});
function el(e) {
	let { children: t, elementsRef: n, labelsRef: r } = e, [i, a] = M.useState(() => /* @__PURE__ */ new Set()), o = M.useCallback((e) => {
		a((t) => new Set(t).add(e));
	}, []), s = M.useCallback((e) => {
		a((t) => {
			let n = new Set(t);
			return n.delete(e), n;
		});
	}, []), c = M.useMemo(() => {
		let e = /* @__PURE__ */ new Map();
		return Array.from(i.keys()).sort(Qc).forEach((t, n) => {
			e.set(t, n);
		}), e;
	}, [i]);
	return /*#__PURE__*/ (0, N.jsx)($c.Provider, {
		value: M.useMemo(() => ({
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
function tl(e) {
	e === void 0 && (e = {});
	let { label: t } = e, { register: n, unregister: r, map: i, elementsRef: a, labelsRef: o } = M.useContext($c), [s, c] = M.useState(null), l = M.useRef(null), u = M.useCallback((e) => {
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
	return $(() => {
		let e = l.current;
		if (e) return n(e), () => {
			r(e);
		};
	}, [n, r]), $(() => {
		let e = l.current ? i.get(l.current) : null;
		e != null && c(e);
	}, [i]), M.useMemo(() => ({
		ref: u,
		index: s ?? -1
	}), [s, u]);
}
var nl = "data-floating-ui-focusable", rl = "active", il = "selected", al = "ArrowLeft", ol = "ArrowRight", sl = "ArrowUp", cl = "ArrowDown", ll = [al, ol], ul = [sl, cl];
[...ll, ...ul];
var dl = { ...M }, fl = !1, pl = 0, ml = () => "floating-ui-" + Math.random().toString(36).slice(2, 6) + pl++;
function hl() {
	let [e, t] = M.useState(() => fl ? ml() : void 0);
	return $(() => {
		e ?? t(ml());
	}, []), M.useEffect(() => {
		fl = !0;
	}, []), e;
}
var gl = dl.useId || hl;
function _l() {
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
var vl = /*#__PURE__*/ M.createContext(null), yl = /*#__PURE__*/ M.createContext(null), bl = () => M.useContext(vl)?.id || null, xl = () => M.useContext(yl);
function Sl(e) {
	let t = gl(), n = xl(), r = bl(), i = e || r;
	return $(() => {
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
function Cl(e) {
	let { children: t, id: n } = e, r = bl();
	return /*#__PURE__*/ (0, N.jsx)(vl.Provider, {
		value: M.useMemo(() => ({
			id: n,
			parentId: r
		}), [n, r]),
		children: t
	});
}
function wl(e) {
	let { children: t } = e, n = M.useRef([]), r = M.useCallback((e) => {
		n.current = [...n.current, e];
	}, []), i = M.useCallback((e) => {
		n.current = n.current.filter((t) => t !== e);
	}, []), [a] = M.useState(() => _l());
	return /*#__PURE__*/ (0, N.jsx)(yl.Provider, {
		value: M.useMemo(() => ({
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
function Tl(e) {
	return "data-floating-ui-" + e;
}
function El(e) {
	e.current !== -1 && (clearTimeout(e.current), e.current = -1);
}
var Dl = /*#__PURE__*/ Tl("safe-polygon");
function Ol(e, t, n) {
	if (n && !kc(n)) return 0;
	if (typeof e == "number") return e;
	if (typeof e == "function") {
		let n = e();
		return typeof n == "number" ? n : n?.[t];
	}
	return e?.[t];
}
function kl(e) {
	return typeof e == "function" ? e() : e;
}
function Al(e, t) {
	t === void 0 && (t = {});
	let { open: n, onOpenChange: r, dataRef: i, events: a, elements: o } = e, { enabled: s = !0, delay: c = 0, handleClose: l = null, mouseOnly: u = !1, restMs: d = 0, move: f = !0 } = t, p = xl(), m = bl(), h = jc(l), g = jc(c), _ = jc(n), v = jc(d), y = M.useRef(), b = M.useRef(-1), x = M.useRef(), C = M.useRef(-1), w = M.useRef(!0), T = M.useRef(!1), E = M.useRef(() => {}), D = M.useRef(!1), O = Nc(() => {
		let e = i.current.openEvent?.type;
		return e?.includes("mouse") && e !== "mousedown";
	});
	M.useEffect(() => {
		if (!s) return;
		function e(e) {
			let { open: t } = e;
			t || (El(b), El(C), w.current = !0, D.current = !1);
		}
		return a.on("openchange", e), () => {
			a.off("openchange", e);
		};
	}, [s, a]), M.useEffect(() => {
		if (!s || !h.current || !n) return;
		function e(e) {
			O() && r(!1, e, "hover");
		}
		let t = vc(o.floating).documentElement;
		return t.addEventListener("mouseleave", e), () => {
			t.removeEventListener("mouseleave", e);
		};
	}, [
		o.floating,
		n,
		r,
		s,
		h,
		O
	]);
	let k = M.useCallback(function(e, t, n) {
		t === void 0 && (t = !0), n === void 0 && (n = "hover");
		let i = Ol(g.current, "close", y.current);
		i && !x.current ? (El(b), b.current = window.setTimeout(() => r(!1, e, n), i)) : t && (El(b), r(!1, e, n));
	}, [g, r]), A = Nc(() => {
		E.current(), x.current = void 0;
	}), j = Nc(() => {
		if (T.current) {
			let e = vc(o.floating).body;
			e.style.pointerEvents = "", e.removeAttribute(Dl), T.current = !1;
		}
	}), N = Nc(() => i.current.openEvent ? ["click", "mousedown"].includes(i.current.openEvent.type) : !1);
	M.useEffect(() => {
		if (!s) return;
		function e(e) {
			if (El(b), w.current = !1, u && !kc(y.current) || kl(v.current) > 0 && !Ol(g.current, "open")) return;
			let t = Ol(g.current, "open", y.current);
			t ? b.current = window.setTimeout(() => {
				_.current || r(!0, e, "hover");
			}, t) : n || r(!0, e, "hover");
		}
		function t(e) {
			if (N()) {
				j();
				return;
			}
			E.current();
			let t = vc(o.floating);
			if (El(C), D.current = !1, h.current && i.current.floatingContext) {
				n || El(b), x.current = h.current({
					...i.current.floatingContext,
					tree: p,
					x: e.clientX,
					y: e.clientY,
					onClose() {
						j(), A(), N() || k(e, !0, "safe-polygon");
					}
				});
				let r = x.current;
				t.addEventListener("mousemove", r), E.current = () => {
					t.removeEventListener("mousemove", r);
				};
			} else y.current === "touch" && mc(o.floating, e.relatedTarget) || k(e);
		}
		function a(e) {
			N() || i.current.floatingContext && (h.current == null || h.current({
				...i.current.floatingContext,
				tree: p,
				x: e.clientX,
				y: e.clientY,
				onClose() {
					j(), A(), N() || k(e);
				}
			})(e));
		}
		function c() {
			El(b);
		}
		function l(e) {
			N() || k(e, !1);
		}
		if (S(o.domReference)) {
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
		k,
		A,
		j,
		r,
		n,
		_,
		p,
		g,
		h,
		i,
		N,
		v
	]), $(() => {
		var e;
		if (s && n && (e = h.current) != null && (e = e.__options) != null && e.blockPointerEvents && O()) {
			T.current = !0;
			let e = o.floating;
			if (S(o.domReference) && e) {
				var t;
				let n = vc(o.floating).body;
				n.setAttribute(Dl, "");
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
		O
	]), $(() => {
		n || (y.current = void 0, D.current = !1, A(), j());
	}, [
		n,
		A,
		j
	]), M.useEffect(() => () => {
		A(), El(b), El(C), j();
	}, [
		s,
		o.domReference,
		A,
		j
	]);
	let P = M.useMemo(() => {
		function e(e) {
			y.current = e.pointerType;
		}
		return {
			onPointerDown: e,
			onPointerEnter: e,
			onMouseMove(e) {
				let { nativeEvent: t } = e;
				function i() {
					!w.current && !_.current && r(!0, t, "hover");
				}
				u && !kc(y.current) || n || kl(v.current) === 0 || D.current && e.movementX ** 2 + e.movementY ** 2 < 2 || (El(C), y.current === "touch" ? i() : (D.current = !0, C.current = window.setTimeout(i, kl(v.current))));
			}
		};
	}, [
		u,
		r,
		n,
		_,
		v
	]);
	return M.useMemo(() => s ? { reference: P } : {}, [s, P]);
}
var jl = 0;
function Ml(e, t) {
	t === void 0 && (t = {});
	let { preventScroll: n = !1, cancelPrevious: r = !0, sync: i = !1 } = t;
	r && cancelAnimationFrame(jl);
	let a = () => e?.focus({ preventScroll: n });
	i ? a() : jl = requestAnimationFrame(a);
}
function Nl(e, t) {
	if (!e || !t) return !1;
	let n = t.getRootNode == null ? void 0 : t.getRootNode();
	if (e.contains(t)) return !0;
	if (n && j(n)) {
		let n = t;
		for (; n;) {
			if (e === n) return !0;
			n = n.parentNode || n.host;
		}
	}
	return !1;
}
function Pl(e) {
	return "composedPath" in e ? e.composedPath()[0] : e.target;
}
function Fl(e) {
	return e?.ownerDocument || document;
}
var Il = {
	inert: /*#__PURE__*/ new WeakMap(),
	"aria-hidden": /*#__PURE__*/ new WeakMap(),
	none: /*#__PURE__*/ new WeakMap()
};
function Ll(e) {
	return e === "inert" ? Il.inert : e === "aria-hidden" ? Il["aria-hidden"] : Il.none;
}
var Rl = /*#__PURE__*/ new WeakSet(), zl = {}, Bl = 0, Vl = () => typeof HTMLElement < "u" && "inert" in HTMLElement.prototype;
function Hl(e) {
	return e ? j(e) ? e.host : Hl(e.parentNode) : null;
}
var Ul = (e, t) => t.map((t) => {
	if (e.contains(t)) return t;
	let n = Hl(t);
	return e.contains(n) ? n : null;
}).filter((e) => e != null);
function Wl(e, t, n, r) {
	let i = "data-floating-ui-inert", a = r ? "inert" : n ? "aria-hidden" : null, o = Ul(t, e), s = /* @__PURE__ */ new Set(), c = new Set(o), l = [];
	zl[i] || (zl[i] = /* @__PURE__ */ new WeakMap());
	let u = zl[i];
	o.forEach(d), f(t), s.clear();
	function d(e) {
		e && !s.has(e) && (s.add(e), e.parentNode && d(e.parentNode));
	}
	function f(e) {
		e && !c.has(e) && [].forEach.call(e.children, (e) => {
			if (b(e) !== "script") {
				if (s.has(e)) f(e);
				else {
					let t = a ? e.getAttribute(a) : null, n = t !== null && t !== "false", r = Ll(a), o = (r.get(e) || 0) + 1, s = (u.get(e) || 0) + 1;
					r.set(e, o), u.set(e, s), l.push(e), o === 1 && n && Rl.add(e), s === 1 && e.setAttribute(i, ""), !n && a && e.setAttribute(a, a === "inert" ? "" : "true");
				}
			}
		});
	}
	return Bl++, () => {
		l.forEach((e) => {
			let t = Ll(a), n = (t.get(e) || 0) - 1, r = (u.get(e) || 0) - 1;
			t.set(e, n), u.set(e, r), n || (!Rl.has(e) && a && e.removeAttribute(a), Rl.delete(e)), r || e.removeAttribute(i);
		}), Bl--, Bl || (Il.inert = /* @__PURE__ */ new WeakMap(), Il["aria-hidden"] = /* @__PURE__ */ new WeakMap(), Il.none = /* @__PURE__ */ new WeakMap(), Rl = /* @__PURE__ */ new WeakSet(), zl = {});
	};
}
function Gl(e, t, n) {
	t === void 0 && (t = !1), n === void 0 && (n = !1);
	let r = Fl(e[0]).body;
	return Wl(e.concat(Array.from(r.querySelectorAll("[aria-live],[role=\"status\"],output"))), r, t, n);
}
var Kl = {
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
}, ql = /*#__PURE__*/ M.forwardRef(function(e, t) {
	let [n, r] = M.useState();
	$(() => {
		ic() && r("button");
	}, []);
	let i = {
		ref: t,
		tabIndex: 0,
		role: n,
		"aria-hidden": !n || void 0,
		[Tl("focus-guard")]: "",
		style: Kl
	};
	return /*#__PURE__*/ (0, N.jsx)("span", {
		...e,
		...i
	});
}), Jl = {
	clipPath: "inset(50%)",
	position: "fixed",
	top: 0,
	left: 0
}, Yl = /*#__PURE__*/ M.createContext(null), Xl = /*#__PURE__*/ Tl("portal");
function Zl(e) {
	e === void 0 && (e = {});
	let { id: t, root: n } = e, r = gl(), i = $l(), [a, o] = M.useState(null), s = M.useRef(null);
	return $(() => () => {
		a?.remove(), queueMicrotask(() => {
			s.current = null;
		});
	}, [a]), $(() => {
		if (!r || s.current) return;
		let e = t ? document.getElementById(t) : null;
		if (!e) return;
		let n = document.createElement("div");
		n.id = r, n.setAttribute(Xl, ""), e.appendChild(n), s.current = n, o(n);
	}, [t, r]), $(() => {
		if (n === null || !r || s.current) return;
		let e = n || i?.portalNode;
		e && !A(e) && (e = e.current), e ||= document.body;
		let a = null;
		t && (a = document.createElement("div"), a.id = t, e.appendChild(a));
		let c = document.createElement("div");
		c.id = r, c.setAttribute(Xl, ""), e = a || e, e.appendChild(c), s.current = c, o(c);
	}, [
		t,
		n,
		r,
		i
	]), a;
}
function Ql(e) {
	let { children: t, id: n, root: r, preserveTabOrder: i = !0 } = e, a = Zl({
		id: n,
		root: r
	}), [o, s] = M.useState(null), c = M.useRef(null), l = M.useRef(null), u = M.useRef(null), d = M.useRef(null), f = o?.modal, p = o?.open, m = !!o && !o.modal && o.open && i && !!(r || a);
	return M.useEffect(() => {
		if (!a || !i || f) return;
		function e(e) {
			a && Jc(e) && (e.type === "focusin" ? Xc : Yc)(a);
		}
		return a.addEventListener("focusin", e, !0), a.addEventListener("focusout", e, !0), () => {
			a.removeEventListener("focusin", e, !0), a.removeEventListener("focusout", e, !0);
		};
	}, [
		a,
		i,
		f
	]), M.useEffect(() => {
		a && (p || Xc(a));
	}, [p, a]), /*#__PURE__*/ (0, N.jsxs)(Yl.Provider, {
		value: M.useMemo(() => ({
			preserveTabOrder: i,
			beforeOutsideRef: c,
			afterOutsideRef: l,
			beforeInsideRef: u,
			afterInsideRef: d,
			portalNode: a,
			setFocusManagerState: s
		}), [i, a]),
		children: [
			m && a && /*#__PURE__*/ (0, N.jsx)(ql, {
				"data-type": "outside",
				ref: c,
				onFocus: (e) => {
					if (Jc(e, a)) {
						var t;
						(t = u.current) == null || t.focus();
					} else qc(o ? o.domReference : null)?.focus();
				}
			}),
			m && a && /*#__PURE__*/ (0, N.jsx)("span", {
				"aria-owns": a.id,
				style: Jl
			}),
			a && /*#__PURE__*/ $o.createPortal(t, a),
			m && a && /*#__PURE__*/ (0, N.jsx)(ql, {
				"data-type": "outside",
				ref: l,
				onFocus: (e) => {
					if (Jc(e, a)) {
						var t;
						(t = d.current) == null || t.focus();
					} else Kc(o ? o.domReference : null)?.focus(), o != null && o.closeOnFocusOut && o?.onOpenChange(!1, e.nativeEvent, "focus-out");
				}
			})
		]
	});
}
var $l = () => M.useContext(Yl);
function eu(e) {
	return M.useMemo(() => (t) => {
		e.forEach((e) => {
			e && (e.current = t);
		});
	}, e);
}
var tu = 20, nu = [];
function ru() {
	nu = nu.filter((e) => e.deref()?.isConnected);
}
function iu(e) {
	ru(), e && b(e) !== "body" && (nu.push(new WeakRef(e)), nu.length > tu && (nu = nu.slice(-20)));
}
function au() {
	return ru(), nu[nu.length - 1]?.deref();
}
function ou(e) {
	let t = Wc();
	return tc(e, t) ? e : $s(e, t)[0] || e;
}
function su(e, t) {
	var n;
	if (!t.current.includes("floating") && !((n = e.getAttribute("role")) != null && n.includes("dialog"))) return;
	let r = Wc(), i = ec(e, r).filter((e) => {
		let t = e.getAttribute("data-tabindex") || "";
		return tc(e, r) || e.hasAttribute("data-tabindex") && !t.startsWith("-");
	}), a = e.getAttribute("tabindex");
	t.current.includes("floating") || i.length === 0 ? a !== "0" && e.setAttribute("tabindex", "0") : (a !== "-1" || e.hasAttribute("data-tabindex") && e.getAttribute("data-tabindex") !== "-1") && (e.setAttribute("tabindex", "-1"), e.setAttribute("data-tabindex", "-1"));
}
var cu = /*#__PURE__*/ M.forwardRef(function(e, t) {
	return /*#__PURE__*/ (0, N.jsx)("button", {
		...e,
		type: "button",
		ref: t,
		tabIndex: -1,
		style: Kl
	});
});
function lu(e) {
	let { context: t, children: n, disabled: r = !1, order: i = ["content"], guards: a = !0, initialFocus: o = 0, returnFocus: s = !0, restoreFocus: c = !1, modal: l = !0, visuallyHiddenDismiss: u = !1, closeOnFocusOut: d = !0, outsideElementsInert: f = !1, getInsideElements: p = () => [] } = e, { open: m, onOpenChange: h, events: g, dataRef: _, elements: { domReference: v, floating: y } } = t, b = Nc(() => _.current.floatingContext?.nodeId), S = Nc(p), C = typeof o == "number" && o < 0, w = bc(v) && C, T = Vl(), E = !T || a, D = !E || T && f, O = jc(i), k = jc(o), A = jc(s), j = xl(), P = $l(), F = M.useRef(null), I = M.useRef(null), L = M.useRef(!1), ee = M.useRef(!1), R = M.useRef(-1), z = M.useRef(-1), te = P != null, B = xc(y), V = Nc(function(e) {
		return e === void 0 && (e = B), e ? $s(e, Wc()) : [];
	}), H = Nc((e) => {
		let t = V(e);
		return O.current.map((e) => v && e === "reference" ? v : B && e === "floating" ? B : t).filter(Boolean).flat();
	});
	M.useEffect(() => {
		if (r || !l) return;
		function e(e) {
			if (e.key === "Tab") {
				mc(B, pc(vc(B))) && V().length === 0 && !w && Tc(e);
				let t = H(), n = hc(e);
				O.current[0] === "reference" && n === v && (Tc(e), e.shiftKey ? Ml(t[t.length - 1]) : Ml(t[1])), O.current[1] === "floating" && n === B && e.shiftKey && (Tc(e), Ml(t[0]));
			}
		}
		let t = vc(B);
		return t.addEventListener("keydown", e), () => {
			t.removeEventListener("keydown", e);
		};
	}, [
		r,
		v,
		B,
		l,
		O,
		w,
		V,
		H
	]), M.useEffect(() => {
		if (r || !y) return;
		function e(e) {
			let t = hc(e), n = V().indexOf(t);
			n !== -1 && (R.current = n);
		}
		return y.addEventListener("focusin", e), () => {
			y.removeEventListener("focusin", e);
		};
	}, [
		r,
		y,
		V
	]), M.useEffect(() => {
		if (r || !d) return;
		function e() {
			ee.current = !0, setTimeout(() => {
				ee.current = !1;
			});
		}
		function t(e) {
			let t = e.relatedTarget, n = e.currentTarget, r = hc(e);
			queueMicrotask(() => {
				let i = b(), a = !(mc(v, t) || mc(y, t) || mc(t, y) || mc(P?.portalNode, t) || t != null && t.hasAttribute(Tl("focus-guard")) || j && (Sc(j.nodesRef.current, i).find((e) => mc(e.context?.elements.floating, t) || mc(e.context?.elements.domReference, t)) || wc(j.nodesRef.current, i).find((e) => [e.context?.elements.floating, xc(e.context?.elements.floating)].includes(t) || e.context?.elements.domReference === t)));
				if (n === v && B && su(B, O), c && n !== v && !(r != null && r.isConnected) && pc(vc(B)) === vc(B).body) {
					x(B) && B.focus();
					let e = R.current, t = V(), n = t[e] || t[t.length - 1] || B;
					x(n) && n.focus();
				}
				_.current.insideReactTree ? _.current.insideReactTree = !1 : (w || !l) && t && a && !ee.current && t !== au() && (L.current = !0, h(!1, e, "focus-out"));
			});
		}
		let n = !(j || !P);
		function i() {
			El(z), _.current.insideReactTree = !0, z.current = window.setTimeout(() => {
				_.current.insideReactTree = !1;
			});
		}
		if (y && x(v)) return v.addEventListener("focusout", t), v.addEventListener("pointerdown", e), y.addEventListener("focusout", t), n && y.addEventListener("focusout", i, !0), () => {
			v.removeEventListener("focusout", t), v.removeEventListener("pointerdown", e), y.removeEventListener("focusout", t), n && y.removeEventListener("focusout", i, !0);
		};
	}, [
		r,
		v,
		y,
		B,
		l,
		j,
		P,
		h,
		d,
		c,
		V,
		w,
		b,
		O,
		_
	]);
	let ne = M.useRef(null), U = M.useRef(null), re = eu([ne, P?.beforeInsideRef]), W = eu([U, P?.afterInsideRef]);
	M.useEffect(() => {
		var e, t;
		if (r || !y) return;
		let n = Array.from((P == null || (e = P.portalNode) == null ? void 0 : e.querySelectorAll("[" + Tl("portal") + "]")) || []), i = (t = (j ? wc(j.nodesRef.current, b()) : []).find((e) => bc(e.context?.elements.domReference || null))) == null || (t = t.context) == null ? void 0 : t.elements.domReference, a = [
			y,
			i,
			...n,
			...S(),
			F.current,
			I.current,
			ne.current,
			U.current,
			P?.beforeOutsideRef.current,
			P?.afterOutsideRef.current,
			O.current.includes("reference") || w ? v : null
		].filter((e) => e != null), o = l || w ? Gl(a, !D, D) : Gl(a);
		return () => {
			o();
		};
	}, [
		r,
		v,
		y,
		l,
		O,
		P,
		w,
		E,
		D,
		j,
		b,
		S
	]), $(() => {
		if (r || !x(B)) return;
		let e = pc(vc(B));
		queueMicrotask(() => {
			let t = H(B), n = k.current, r = (typeof n == "number" ? t[n] : n.current) || B, i = mc(B, e);
			!C && !i && m && Ml(r, { preventScroll: r === B });
		});
	}, [
		r,
		m,
		B,
		C,
		H,
		k
	]), $(() => {
		if (r || !B) return;
		let e = vc(B);
		iu(pc(e));
		function t(e) {
			let { reason: t, event: n, nested: r } = e;
			if (["hover", "safe-polygon"].includes(t) && n.type === "mouseleave" && (L.current = !0), t === "outside-press") {
				if (r) L.current = !1;
				else if (Dc(n) || Oc(n)) L.current = !1;
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
		n.setAttribute("tabindex", "-1"), n.setAttribute("aria-hidden", "true"), Object.assign(n.style, Kl), te && v && v.insertAdjacentElement("afterend", n);
		function i() {
			if (typeof A.current == "boolean") {
				let e = v || au();
				return e && e.isConnected ? e : n;
			}
			return A.current.current || n;
		}
		return () => {
			g.off("openchange", t);
			let r = pc(e), a = mc(y, r) || j && Sc(j.nodesRef.current, b(), !1).some((e) => mc(e.context?.elements.floating, r)), o = i();
			queueMicrotask(() => {
				let t = ou(o);
				A.current && !L.current && x(t) && (t === r || r === e.body || a) && t.focus({ preventScroll: !0 }), n.remove();
			});
		};
	}, [
		r,
		y,
		B,
		A,
		_,
		g,
		j,
		te,
		v,
		b
	]), M.useEffect(() => (queueMicrotask(() => {
		L.current = !1;
	}), () => {
		queueMicrotask(ru);
	}), [r]), $(() => {
		if (!r && P) return P.setFocusManagerState({
			modal: l,
			closeOnFocusOut: d,
			open: m,
			onOpenChange: h,
			domReference: v
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
		v
	]), $(() => {
		r || B && su(B, O);
	}, [
		r,
		B,
		O
	]);
	function ie(e) {
		return r || !u || !l ? null : /*#__PURE__*/ (0, N.jsx)(cu, {
			ref: e === "start" ? F : I,
			onClick: (e) => h(!1, e.nativeEvent),
			children: typeof u == "string" ? u : "Dismiss"
		});
	}
	let G = !r && E && (!l || !w) && (te || l);
	return /*#__PURE__*/ (0, N.jsxs)(N.Fragment, { children: [
		G && /*#__PURE__*/ (0, N.jsx)(ql, {
			"data-type": "inside",
			ref: re,
			onFocus: (e) => {
				if (l) {
					let e = H();
					Ml(i[0] === "reference" ? e[0] : e[e.length - 1]);
				} else if (P != null && P.preserveTabOrder && P.portalNode) {
					if (L.current = !1, Jc(e, P.portalNode)) Kc(v)?.focus();
					else {
						var t;
						(t = P.beforeOutsideRef.current) == null || t.focus();
					}
				}
			}
		}),
		!w && ie("start"),
		n,
		ie("end"),
		G && /*#__PURE__*/ (0, N.jsx)(ql, {
			"data-type": "inside",
			ref: W,
			onFocus: (e) => {
				if (l) Ml(H()[0]);
				else if (P != null && P.preserveTabOrder && P.portalNode) {
					if (d && (L.current = !0), Jc(e, P.portalNode)) qc(v)?.focus();
					else {
						var t;
						(t = P.afterOutsideRef.current) == null || t.focus();
					}
				}
			}
		})
	] });
}
function uu(e) {
	return x(e.target) && e.target.tagName === "BUTTON";
}
function du(e) {
	return x(e.target) && e.target.tagName === "A";
}
function fu(e) {
	return yc(e);
}
function pu(e, t) {
	t === void 0 && (t = {});
	let { open: n, onOpenChange: r, dataRef: i, elements: { domReference: a } } = e, { enabled: o = !0, event: s = "click", toggle: c = !0, ignoreMouse: l = !1, keyboardHandlers: u = !0, stickIfOpen: d = !0 } = t, f = M.useRef(), p = M.useRef(!1), m = M.useMemo(() => ({
		onPointerDown(e) {
			f.current = e.pointerType;
		},
		onMouseDown(e) {
			let t = f.current;
			e.button === 0 && s !== "click" && (kc(t, !0) && l || (n && c && (!(i.current.openEvent && d) || i.current.openEvent.type === "mousedown") ? r(!1, e.nativeEvent, "click") : (e.preventDefault(), r(!0, e.nativeEvent, "click"))));
		},
		onClick(e) {
			let t = f.current;
			s === "mousedown" && f.current ? f.current = void 0 : kc(t, !0) && l || (n && c && (!(i.current.openEvent && d) || i.current.openEvent.type === "click") ? r(!1, e.nativeEvent, "click") : r(!0, e.nativeEvent, "click"));
		},
		onKeyDown(e) {
			f.current = void 0, !(e.defaultPrevented || !u || uu(e)) && (e.key === " " && !fu(a) && (e.preventDefault(), p.current = !0), !du(e) && e.key === "Enter" && r(!(n && c), e.nativeEvent, "click"));
		},
		onKeyUp(e) {
			e.defaultPrevented || !u || uu(e) || fu(a) || e.key === " " && p.current && (p.current = !1, r(!(n && c), e.nativeEvent, "click"));
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
	return M.useMemo(() => o ? { reference: m } : {}, [o, m]);
}
var mu = {
	pointerdown: "onPointerDown",
	mousedown: "onMouseDown",
	click: "onClick"
}, hu = {
	pointerdown: "onPointerDownCapture",
	mousedown: "onMouseDownCapture",
	click: "onClickCapture"
}, gu = (e) => ({
	escapeKey: typeof e == "boolean" ? e : e?.escapeKey ?? !1,
	outsidePress: typeof e == "boolean" ? e : e?.outsidePress ?? !0
});
function _u(e, t) {
	t === void 0 && (t = {});
	let { open: n, onOpenChange: r, elements: i, dataRef: a } = e, { enabled: o = !0, escapeKey: s = !0, outsidePress: c = !0, outsidePressEvent: l = "pointerdown", referencePress: u = !1, referencePressEvent: d = "pointerdown", ancestorScroll: f = !1, bubbles: p, capture: m } = t, g = xl(), v = Nc(typeof c == "function" ? c : () => !1), b = typeof c == "function" ? v : c, C = M.useRef(!1), { escapeKey: T, outsidePress: E } = gu(p), { escapeKey: O, outsidePress: k } = gu(m), A = M.useRef(!1), j = Nc((e) => {
		if (!n || !o || !s || e.key !== "Escape" || A.current) return;
		let t = a.current.floatingContext?.nodeId, i = g ? Sc(g.nodesRef.current, t) : [];
		if (!T && (e.stopPropagation(), i.length > 0)) {
			let e = !0;
			if (i.forEach((t) => {
				var n;
				(n = t.context) != null && n.open && !t.context.dataRef.current.__escapeKeyBubbles && (e = !1);
			}), !e) return;
		}
		r(!1, Ec(e) ? e.nativeEvent : e, "escape-key");
	}), N = Nc((e) => {
		var t;
		let n = () => {
			var t;
			j(e), (t = hc(e)) == null || t.removeEventListener("keydown", n);
		};
		(t = hc(e)) == null || t.addEventListener("keydown", n);
	}), P = Nc((e) => {
		let t = a.current.insideReactTree;
		a.current.insideReactTree = !1;
		let n = C.current;
		if (C.current = !1, l === "click" && n || t || typeof b == "function" && !b(e)) return;
		let o = hc(e), s = "[" + Tl("inert") + "]", c = vc(i.floating).querySelectorAll(s), u = S(o) ? o : null;
		for (; u && !h(u);) {
			let e = w(u);
			if (h(e) || !S(e)) break;
			u = e;
		}
		if (c.length && S(o) && !_c(o) && !mc(o, i.floating) && Array.from(c).every((e) => !mc(u, e))) return;
		if (x(o) && L) {
			let t = h(o), n = y(o), r = /auto|scroll/, i = t || r.test(n.overflowX), a = t || r.test(n.overflowY), s = i && o.clientWidth > 0 && o.scrollWidth > o.clientWidth, c = a && o.clientHeight > 0 && o.scrollHeight > o.clientHeight, l = n.direction === "rtl", u = c && (l ? e.offsetX <= o.offsetWidth - o.clientWidth : e.offsetX > o.clientWidth), d = s && e.offsetY > o.clientHeight;
			if (u || d) return;
		}
		let d = a.current.floatingContext?.nodeId, f = g && Sc(g.nodesRef.current, d).some((t) => gc(e, t.context?.elements.floating));
		if (gc(e, i.floating) || gc(e, i.domReference) || f) return;
		let p = g ? Sc(g.nodesRef.current, d) : [];
		if (p.length > 0) {
			let e = !0;
			if (p.forEach((t) => {
				var n;
				(n = t.context) != null && n.open && !t.context.dataRef.current.__outsidePressBubbles && (e = !1);
			}), !e) return;
		}
		r(!1, e, "outside-press");
	}), F = Nc((e) => {
		var t;
		let n = () => {
			var t;
			P(e), (t = hc(e)) == null || t.removeEventListener(l, n);
		};
		(t = hc(e)) == null || t.addEventListener(l, n);
	});
	M.useEffect(() => {
		if (!n || !o) return;
		a.current.__escapeKeyBubbles = T, a.current.__outsidePressBubbles = E;
		let e = -1;
		function t(e) {
			r(!1, e, "ancestor-scroll");
		}
		function c() {
			window.clearTimeout(e), A.current = !0;
		}
		function u() {
			e = window.setTimeout(() => {
				A.current = !1;
			}, _() ? 5 : 0);
		}
		let d = vc(i.floating);
		s && (d.addEventListener("keydown", O ? N : j, O), d.addEventListener("compositionstart", c), d.addEventListener("compositionend", u)), b && d.addEventListener(l, k ? F : P, k);
		let p = [];
		return f && (S(i.domReference) && (p = D(i.domReference)), S(i.floating) && (p = p.concat(D(i.floating))), !S(i.reference) && i.reference && i.reference.contextElement && (p = p.concat(D(i.reference.contextElement)))), p = p.filter((e) => e !== d.defaultView?.visualViewport), p.forEach((e) => {
			e.addEventListener("scroll", t, { passive: !0 });
		}), () => {
			s && (d.removeEventListener("keydown", O ? N : j, O), d.removeEventListener("compositionstart", c), d.removeEventListener("compositionend", u)), b && d.removeEventListener(l, k ? F : P, k), p.forEach((e) => {
				e.removeEventListener("scroll", t);
			}), window.clearTimeout(e);
		};
	}, [
		a,
		i,
		s,
		b,
		l,
		n,
		r,
		f,
		o,
		T,
		E,
		j,
		O,
		N,
		P,
		k,
		F
	]), M.useEffect(() => {
		a.current.insideReactTree = !1;
	}, [
		a,
		b,
		l
	]);
	let I = M.useMemo(() => ({
		onKeyDown: j,
		...u && {
			[mu[d]]: (e) => {
				r(!1, e.nativeEvent, "reference-press");
			},
			...d !== "click" && { onClick(e) {
				r(!1, e.nativeEvent, "reference-press");
			} }
		}
	}), [
		j,
		r,
		u,
		d
	]), L = M.useMemo(() => {
		function e(e) {
			e.button === 0 && (C.current = !0);
		}
		return {
			onKeyDown: j,
			onMouseDown: e,
			onMouseUp: e,
			[hu[l]]: () => {
				a.current.insideReactTree = !0;
			}
		};
	}, [
		j,
		l,
		a
	]);
	return M.useMemo(() => o ? {
		reference: I,
		floating: L
	} : {}, [
		o,
		I,
		L
	]);
}
function vu(e) {
	let { open: t = !1, onOpenChange: n, elements: r } = e, i = gl(), a = M.useRef({}), [o] = M.useState(() => _l()), s = bl() != null, [c, l] = M.useState(r.reference), u = Nc((e, t, r) => {
		a.current.openEvent = e ? t : void 0, o.emit("openchange", {
			open: e,
			event: t,
			reason: r,
			nested: s
		}), n?.(e, t, r);
	}), d = M.useMemo(() => ({ setPositionReference: l }), []), f = M.useMemo(() => ({
		reference: c || r.reference || null,
		floating: r.floating || null,
		domReference: r.reference
	}), [
		c,
		r.reference,
		r.floating
	]);
	return M.useMemo(() => ({
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
function yu(e) {
	e === void 0 && (e = {});
	let { nodeId: t } = e, n = vu({
		...e,
		elements: {
			reference: null,
			floating: null,
			...e.elements
		}
	}), r = e.rootContext || n, i = r.elements, [a, o] = M.useState(null), [s, c] = M.useState(null), l = i?.domReference || a, u = M.useRef(null), d = xl();
	$(() => {
		l && (u.current = l);
	}, [l]);
	let f = v({
		...e,
		elements: {
			...i,
			...s && { reference: s }
		}
	}), p = M.useCallback((e) => {
		let t = S(e) ? {
			getBoundingClientRect: () => e.getBoundingClientRect(),
			getClientRects: () => e.getClientRects(),
			contextElement: e
		} : e;
		c(t), f.refs.setReference(t);
	}, [f.refs]), m = M.useCallback((e) => {
		(S(e) || e === null) && (u.current = e, o(e)), (S(f.refs.reference.current) || f.refs.reference.current === null || e !== null && !S(e)) && f.refs.setReference(e);
	}, [f.refs]), h = M.useMemo(() => ({
		...f.refs,
		setReference: m,
		setPositionReference: p,
		domReference: u
	}), [
		f.refs,
		m,
		p
	]), g = M.useMemo(() => ({
		...f.elements,
		domReference: l
	}), [f.elements, l]), _ = M.useMemo(() => ({
		...f,
		...r,
		refs: h,
		elements: g,
		nodeId: t
	}), [
		f,
		h,
		g,
		t,
		r
	]);
	return $(() => {
		r.dataRef.current.floatingContext = _;
		let e = d?.nodesRef.current.find((e) => e.id === t);
		e && (e.context = _);
	}), M.useMemo(() => ({
		...f,
		context: _,
		refs: h,
		elements: g
	}), [
		f,
		h,
		g,
		_
	]);
}
function bu(e, t, n) {
	let r = /* @__PURE__ */ new Map(), i = n === "item", a = e;
	if (i && e) {
		let { [rl]: t, [il]: n, ...r } = e;
		a = r;
	}
	return {
		...n === "floating" && {
			tabIndex: -1,
			[nl]: ""
		},
		...a,
		...t.map((t) => {
			let r = t ? t[n] : null;
			return typeof r == "function" ? e ? r(e) : null : r;
		}).concat(e).reduce((e, t) => (t && Object.entries(t).forEach((t) => {
			let [n, a] = t;
			if (!(i && [rl, il].includes(n))) {
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
function xu(e) {
	e === void 0 && (e = []);
	let t = e.map((e) => e?.reference), n = e.map((e) => e?.floating), r = e.map((e) => e?.item), i = M.useCallback((t) => bu(t, e, "reference"), t), a = M.useCallback((t) => bu(t, e, "floating"), n), o = M.useCallback((t) => bu(t, e, "item"), r);
	return M.useMemo(() => ({
		getReferenceProps: i,
		getFloatingProps: a,
		getItemProps: o
	}), [
		i,
		a,
		o
	]);
}
var Su = "Escape";
function Cu(e, t, n) {
	switch (e) {
		case "vertical": return t;
		case "horizontal": return n;
		default: return t || n;
	}
}
function wu(e, t) {
	return Cu(t, e === sl || e === cl, e === al || e === ol);
}
function Tu(e, t, n) {
	return Cu(t, e === cl, n ? e === al : e === ol) || e === "Enter" || e === " " || e === "";
}
function Eu(e, t, n) {
	return Cu(t, n ? e === al : e === ol, e === cl);
}
function Du(e, t, n, r) {
	return t === "both" || t === "horizontal" && r && r > 1 ? e === Su : Cu(t, n ? e === ol : e === al, e === sl);
}
function Ou(e, t) {
	let { open: n, onOpenChange: r, elements: i, floatingId: a } = e, { listRef: o, activeIndex: s, onNavigate: c = () => {}, enabled: l = !0, selectedIndex: u = null, allowEscape: d = !1, loop: f = !1, nested: p = !1, rtl: m = !1, virtual: h = !1, focusItemOnOpen: g = "auto", focusItemOnHover: _ = !0, openOnArrowKeyDown: v = !0, disabledIndices: y = void 0, orientation: b = "vertical", parentOrientation: S, cols: C = 1, scrollItemIntoView: w = !0, virtualItemRef: T, itemSizes: E, dense: D = !1 } = t, O = jc(xc(i.floating)), k = bl(), A = xl();
	$(() => {
		e.dataRef.current.orientation = b;
	}, [e, b]);
	let j = Nc(() => {
		c(F.current === -1 ? null : F.current);
	}), N = bc(i.domReference), P = M.useRef(g), F = M.useRef(u ?? -1), I = M.useRef(null), L = M.useRef(!0), ee = M.useRef(j), R = M.useRef(!!i.floating), z = M.useRef(n), te = M.useRef(!1), B = M.useRef(!1), V = jc(y), H = jc(n), ne = jc(w), U = jc(u), [re, W] = M.useState(), [ie, G] = M.useState(), ae = Nc(() => {
		function e(e) {
			if (h) {
				var t;
				(t = e.id) != null && t.endsWith("-fui-option") && (e.id = a + "-" + Math.random().toString(16).slice(2, 10)), W(e.id), A?.events.emit("virtualfocus", e), T && (T.current = e);
			} else Ml(e, {
				sync: te.current,
				preventScroll: !0
			});
		}
		let t = o.current[F.current], n = B.current;
		t && e(t), (te.current ? (e) => e() : requestAnimationFrame)(() => {
			let r = o.current[F.current] || t;
			if (!r) return;
			t || e(r);
			let i = ne.current;
			i && se && (n || !L.current) && (r.scrollIntoView == null || r.scrollIntoView(typeof i == "boolean" ? {
				block: "nearest",
				inline: "nearest"
			} : i));
		});
	});
	$(() => {
		l && (n && i.floating ? P.current && u != null && (B.current = !0, F.current = u, j()) : R.current && (F.current = -1, ee.current()));
	}, [
		l,
		n,
		i.floating,
		u,
		j
	]), $(() => {
		if (l && n && i.floating) {
			if (s == null) {
				if (te.current = !1, U.current != null) return;
				if (R.current && (F.current = -1, ae()), (!z.current || !R.current) && P.current && (I.current != null || P.current === !0 && I.current == null)) {
					let e = 0, t = () => {
						o.current[0] == null ? (e < 2 && (e ? requestAnimationFrame : queueMicrotask)(t), e++) : (F.current = I.current == null || Tu(I.current, b, m) || p ? Ic(o, V.current) : Lc(o, V.current), I.current = null, j());
					};
					t();
				}
			} else Fc(o, s) || (F.current = s, ae(), B.current = !1);
		}
	}, [
		l,
		n,
		i.floating,
		s,
		U,
		p,
		o,
		b,
		m,
		j,
		ae,
		V
	]), $(() => {
		var e;
		if (!l || i.floating || !A || h || !R.current) return;
		let t = A.nodesRef.current, n = (e = t.find((e) => e.id === k)) == null || (e = e.context) == null ? void 0 : e.elements.floating, r = pc(vc(i.floating)), a = t.some((e) => e.context && mc(e.context.elements.floating, r));
		n && !a && L.current && n.focus({ preventScroll: !0 });
	}, [
		l,
		i.floating,
		A,
		k,
		h
	]), $(() => {
		if (!l || !A || !h || k) return;
		function e(e) {
			G(e.id), T && (T.current = e);
		}
		return A.events.on("virtualfocus", e), () => {
			A.events.off("virtualfocus", e);
		};
	}, [
		l,
		A,
		h,
		k,
		T
	]), $(() => {
		ee.current = j, z.current = n, R.current = !!i.floating;
	}), $(() => {
		n || (I.current = null, P.current = g);
	}, [n, g]);
	let oe = s != null, se = M.useMemo(() => {
		function e(e) {
			if (!H.current) return;
			let t = o.current.indexOf(e);
			t !== -1 && F.current !== t && (F.current = t, j());
		}
		return {
			onFocus(t) {
				let { currentTarget: n } = t;
				te.current = !0, e(n);
			},
			onClick: (e) => {
				let { currentTarget: t } = e;
				return t.focus({ preventScroll: !0 });
			},
			onMouseMove(t) {
				let { currentTarget: n } = t;
				te.current = !0, B.current = !1, _ && e(n);
			},
			onPointerLeave(e) {
				let { pointerType: t } = e;
				if (L.current && t !== "touch" && (te.current = !0, _ && (F.current = -1, j(), !h))) {
					var n;
					(n = O.current) == null || n.focus({ preventScroll: !0 });
				}
			}
		};
	}, [
		H,
		O,
		_,
		o,
		j,
		h
	]), ce = M.useCallback(() => {
		var e;
		return S ?? (A == null || (e = A.nodesRef.current.find((e) => e.id === k)) == null || (e = e.context) == null || (e = e.dataRef) == null ? void 0 : e.current.orientation);
	}, [
		k,
		A,
		S
	]), le = Nc((e) => {
		if (L.current = !1, te.current = !0, e.which === 229 || !H.current && e.currentTarget === O.current) return;
		if (p && Du(e.key, b, m, C)) {
			wu(e.key, ce()) || Tc(e), r(!1, e.nativeEvent, "list-navigation"), x(i.domReference) && (h ? A?.events.emit("virtualfocus", i.domReference) : i.domReference.focus());
			return;
		}
		let t = F.current, a = Ic(o, y), s = Lc(o, y);
		if (N || (e.key === "Home" && (Tc(e), F.current = a, j()), e.key === "End" && (Tc(e), F.current = s, j())), C > 1) {
			let t = E || Array.from({ length: o.current.length }, () => ({
				width: 1,
				height: 1
			})), n = Bc(t, C, D), r = n.findIndex((e) => e != null && !Uc(o, e, y)), i = n.reduce((e, t, n) => t != null && !Uc(o, t, y) ? n : e, -1), c = n[zc({ current: n.map((e) => e == null ? null : o.current[e]) }, {
				event: e,
				orientation: b,
				loop: f,
				rtl: m,
				cols: C,
				disabledIndices: Hc([...(typeof y == "function" ? null : y) || o.current.map((e, t) => Uc(o, t, y) ? t : void 0), void 0], n),
				minIndex: r,
				maxIndex: i,
				prevIndex: Vc(F.current > s ? a : F.current, t, n, C, e.key === cl ? "bl" : e.key === (m ? al : ol) ? "tr" : "tl"),
				stopEvent: !0
			})];
			if (c != null && (F.current = c, j()), b === "both") return;
		}
		if (wu(e.key, b)) {
			if (Tc(e), n && !h && pc(e.currentTarget.ownerDocument) === e.currentTarget) {
				F.current = Tu(e.key, b, m) ? a : s, j();
				return;
			}
			Tu(e.key, b, m) ? f ? F.current = t >= s ? d && t !== o.current.length ? -1 : a : Rc(o, {
				startingIndex: t,
				disabledIndices: y
			}) : F.current = Math.min(s, Rc(o, {
				startingIndex: t,
				disabledIndices: y
			})) : f ? F.current = t <= a ? d && t !== -1 ? o.current.length : s : Rc(o, {
				startingIndex: t,
				decrement: !0,
				disabledIndices: y
			}) : F.current = Math.max(a, Rc(o, {
				startingIndex: t,
				decrement: !0,
				disabledIndices: y
			})), Fc(o, F.current) && (F.current = -1), j();
		}
	}), ue = M.useMemo(() => h && n && oe && { "aria-activedescendant": ie || re }, [
		h,
		n,
		oe,
		ie,
		re
	]), de = M.useMemo(() => ({
		"aria-orientation": b === "both" ? void 0 : b,
		...N ? {} : ue,
		onKeyDown: le,
		onPointerMove() {
			L.current = !0;
		}
	}), [
		ue,
		le,
		b,
		N
	]), fe = M.useMemo(() => {
		function e(e) {
			g === "auto" && Dc(e.nativeEvent) && (P.current = !0);
		}
		function t(e) {
			P.current = g, g === "auto" && Oc(e.nativeEvent) && (P.current = !0);
		}
		return {
			...ue,
			onKeyDown(e) {
				L.current = !1;
				let t = e.key.startsWith("Arrow"), i = ["Home", "End"].includes(e.key), a = t || i, s = Eu(e.key, b, m), c = Du(e.key, b, m, C), l = Eu(e.key, ce(), m), d = wu(e.key, b), f = (p ? l : d) || e.key === "Enter" || e.key.trim() === "";
				if (h && n) {
					let t = A?.nodesRef.current.find((e) => e.parentId == null), n = A && t ? Cc(A.nodesRef.current, t.id) : null;
					if (a && n && T) {
						let t = new KeyboardEvent("keydown", {
							key: e.key,
							bubbles: !0
						});
						if (s || c) {
							let r = n.context?.elements.domReference === e.currentTarget, i = c && !r ? n.context?.elements.domReference : s ? o.current.find((e) => e?.id === re) : null;
							i && (Tc(e), i.dispatchEvent(t), G(void 0));
						}
						if ((d || i) && n.context && n.context.open && n.parentId && e.currentTarget !== n.context.elements.domReference) {
							var g;
							Tc(e), (g = n.context.elements.domReference) == null || g.dispatchEvent(t);
							return;
						}
					}
					return le(e);
				}
				if (n || v || !t) {
					if (f) {
						let t = wu(e.key, ce());
						I.current = p && t ? null : e.key;
					}
					p ? l && (Tc(e), n ? (F.current = Ic(o, V.current), j()) : r(!0, e.nativeEvent, "list-navigation")) : d && (u != null && (F.current = u), Tc(e), !n && v ? r(!0, e.nativeEvent, "list-navigation") : le(e), n && j());
				}
			},
			onFocus() {
				n && !h && (F.current = -1, j());
			},
			onPointerDown: t,
			onPointerEnter: t,
			onMouseDown: e,
			onClick: e
		};
	}, [
		re,
		ue,
		C,
		le,
		V,
		g,
		o,
		p,
		j,
		r,
		n,
		v,
		b,
		ce,
		m,
		u,
		A,
		h,
		T
	]);
	return M.useMemo(() => l ? {
		reference: fe,
		floating: de,
		item: se
	} : {}, [
		l,
		fe,
		de,
		se
	]);
}
var ku = /*#__PURE__*/ new Map([
	["select", "listbox"],
	["combobox", "listbox"],
	["label", !1]
]);
function Au(e, t) {
	t === void 0 && (t = {});
	let { open: n, elements: r, floatingId: i } = e, { enabled: a = !0, role: o = "dialog" } = t, s = gl(), c = r.domReference?.id || s, l = M.useMemo(() => xc(r.floating)?.id || i, [r.floating, i]), u = ku.get(o) ?? o, d = bl() != null, f = M.useMemo(() => u === "tooltip" || o === "label" ? { ["aria-" + (o === "label" ? "labelledby" : "describedby")]: n ? l : void 0 } : {
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
	]), p = M.useMemo(() => {
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
	]), m = M.useCallback((e) => {
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
	return M.useMemo(() => a ? {
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
function ju(e, t, n) {
	return n === void 0 && (n = !0), e.filter((e) => e.parentId === t && (!n || e.context?.open)).flatMap((t) => [t, ...ju(e, t.id, n)]);
}
function Mu(e, t) {
	let [n, r] = e, i = !1, a = t.length;
	for (let e = 0, o = a - 1; e < a; o = e++) {
		let [a, s] = t[e] || [0, 0], [c, l] = t[o] || [0, 0];
		s >= r != l >= r && n <= (c - a) * (r - s) / (l - s) + a && (i = !i);
	}
	return i;
}
function Nu(e, t) {
	return e[0] >= t.x && e[0] <= t.x + t.width && e[1] >= t.y && e[1] <= t.y + t.height;
}
function Pu(e) {
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
				El(i), u();
			}
			if (El(i), !c.domReference || !c.floating || s == null || n == null || o == null) return;
			let { clientX: m, clientY: h } = e, g = [m, h], _ = Pl(e), v = e.type === "mouseleave", y = Nl(c.floating, _), b = Nl(c.domReference, _), x = c.domReference.getBoundingClientRect(), C = c.floating.getBoundingClientRect(), w = s.split("-")[0], T = n > C.right - C.width / 2, E = o > C.bottom - C.height / 2, D = Nu(g, x), O = C.width > x.width, k = C.height > x.height, A = (O ? x : C).left, j = (O ? x : C).right, M = (k ? x : C).top, N = (k ? x : C).bottom;
			if (y && (a = !0, !v)) return;
			if (b && (a = !1), b && !v) {
				a = !0;
				return;
			}
			if (v && S(e.relatedTarget) && Nl(c.floating, e.relatedTarget) || f && ju(f.nodesRef.current, d).length) return;
			if (w === "top" && o >= x.bottom - 1 || w === "bottom" && o <= x.top + 1 || w === "left" && n >= x.right - 1 || w === "right" && n <= x.left + 1) return p();
			let P = [];
			switch (w) {
				case "top":
					P = [
						[A, x.top + 1],
						[A, C.bottom - 1],
						[j, C.bottom - 1],
						[j, x.top + 1]
					];
					break;
				case "bottom":
					P = [
						[A, C.top + 1],
						[A, x.bottom - 1],
						[j, x.bottom - 1],
						[j, C.top + 1]
					];
					break;
				case "left":
					P = [
						[C.right - 1, N],
						[C.right - 1, M],
						[x.left + 1, M],
						[x.left + 1, N]
					];
					break;
				case "right": P = [
					[x.right - 1, N],
					[x.right - 1, M],
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
			if (!Mu([m, h], P)) {
				if (a && !D) return p();
				if (!v && r) {
					let t = l(e.clientX, e.clientY);
					if (t !== null && t < .1) return p();
				}
				Mu([m, h], F([n, o])) ? !a && r && (i.current = window.setTimeout(p, 40)) : p();
			}
		};
	};
	return u.__options = { blockPointerEvents: n }, u;
}
// ../../../clicky-ui/packages/ui/src/data/cells/timestamp-format.ts
function Fu(e) {
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
var Iu = {
	s: 1e3,
	m: 6e4,
	h: 36e5,
	d: 864e5,
	w: 6048e5
}, Lu = /^now\s*([+-])\s*(\d+)\s*([smhdwMqy])$/, Ru = /^now\/([wMqy])(?:\s*([+-])\s*(\d+)\s*([wMqy]))?$/;
function zu(e, t = /* @__PURE__ */ new Date()) {
	if (e == null) return null;
	if (typeof e == "string") {
		let n = e.trim();
		if (!n) return null;
		if (n.toLowerCase() === "now") return new Date(t.getTime());
		let r = n.match(Ru);
		if (r) {
			let [, e, n, i, a] = r, o = Vu(t, e);
			return n && i && a && (o = Hu(o, a, Number(i) * (n === "-" ? -1 : 1))), o;
		}
		let i = n.match(Lu);
		if (i) {
			let [, e, n, r] = i, a = r;
			if (Bu(a)) return Hu(t, a, Number(n) * (e === "-" ? -1 : 1));
			let o = Iu[a.toLowerCase()];
			if (!o) return null;
			let s = Number(n) * o * (e === "-" ? -1 : 1);
			return new Date(t.getTime() + s);
		}
	}
	return Fu(e);
}
function Bu(e) {
	return e === "M" || e === "q" || e === "y";
}
function Vu(e, t) {
	let n = e.getUTCFullYear(), r = e.getUTCMonth(), i = e.getUTCDate();
	if (t === "w") {
		let e = new Date(Date.UTC(n, r, i)), t = (e.getUTCDay() + 6) % 7;
		return e.setUTCDate(e.getUTCDate() - t), e;
	}
	return t === "M" ? new Date(Date.UTC(n, r, 1)) : t === "q" ? new Date(Date.UTC(n, Math.floor(r / 3) * 3, 1)) : t === "y" ? new Date(Date.UTC(n, 0, 1)) : new Date(e.getTime());
}
function Hu(e, t, n) {
	let r = new Date(e.getTime());
	if (t === "w") r.setUTCDate(r.getUTCDate() + n * 7);
	else if (t === "M") r.setUTCMonth(r.getUTCMonth() + n);
	else if (t === "q") r.setUTCMonth(r.getUTCMonth() + n * 3);
	else if (t === "y") r.setUTCFullYear(r.getUTCFullYear() + n);
	else {
		let e = Iu[t.toLowerCase()];
		e && r.setTime(r.getTime() + n * e);
	}
	return r;
}
var Uu = 6e4, Wu = 864e5;
function Gu(e) {
	if (e.length === 0) return "iso";
	let t = Infinity, n = -Infinity;
	for (let r of e) {
		let e = r.getTime();
		e < t && (t = e), e > n && (n = e);
	}
	let r = n - t;
	if (r <= Uu) return "relative";
	let i = new Date(t), a = new Date(n);
	return i.getFullYear() === a.getFullYear() && i.getMonth() === a.getMonth() && i.getDate() === a.getDate() ? "time" : i.getFullYear() === a.getFullYear() && r <= Wu * 365 ? "short" : "iso";
}
function Ku(e, t, n = /* @__PURE__ */ new Date()) {
	switch (t) {
		case "relative": return Ju(e, n);
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
		default: return qu(e);
	}
}
function qu(e) {
	return `${e.getFullYear()}-${String(e.getMonth() + 1).padStart(2, "0")}-${String(e.getDate()).padStart(2, "0")} ${String(e.getHours()).padStart(2, "0")}:${String(e.getMinutes()).padStart(2, "0")}`;
}
function Ju(e, t = /* @__PURE__ */ new Date()) {
	let n = t.getTime() - e.getTime(), r = n < 0, i = Math.round(Math.abs(n) / 1e3), a;
	return a = i < 1 ? "just now" : i < 60 ? `${i}s` : i < 3600 ? `${Math.round(i / 60)}m` : i < 86400 ? `${Math.round(i / 3600)}h` : `${Math.round(i / 86400)}d`, a === "just now" ? a : r ? `in ${a}` : `${a} ago`;
}
function Yu(e, t = /* @__PURE__ */ new Date()) {
	let n = Fu(e);
	if (!n) return typeof e == "string" ? e : "";
	let r = n.toLocaleString(void 0, {
		year: "numeric",
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit"
	}), i = Ju(n, t);
	return i === "just now" ? r : `${r} (${i})`;
}
function Xu(e, t) {
	switch (e) {
		case "absolute": return "iso";
		case "relative": return "relative";
		case "time": return "time";
		default: return t;
	}
}
// ../../../clicky-ui/packages/ui/src/overlay/HoverCard.tsx
var Zu = 6, Qu = 120;
function $u({ trigger: e, children: t, placement: n = "top", delay: r = 0, arrow: a = !0, className: o, cardClassName: s }) {
	let [c, l] = (0, M.useState)(!1), [u, d] = (0, M.useState)(null), f = gs(), p = (0, M.useRef)(null), m = (0, M.useRef)(null), h = (0, M.useRef)(null), g = (0, M.useRef)(null), _ = () => {
		g.current !== null && (window.clearTimeout(g.current), g.current = null);
	};
	function v() {
		_(), r > 0 ? (h.current !== null && window.clearTimeout(h.current), h.current = window.setTimeout(() => l(!0), r)) : l(!0);
	}
	function y() {
		h.current !== null && (window.clearTimeout(h.current), h.current = null), _(), g.current = window.setTimeout(() => l(!1), Qu);
	}
	(0, M.useLayoutEffect)(() => {
		if (!c) return;
		let e = () => {
			let e = p.current, t = m.current;
			if (!e || !t) return;
			let r = e.getBoundingClientRect(), i = t.getBoundingClientRect(), a = 0, o = 0;
			switch (n) {
				case "top":
					a = r.top - i.height - Zu, o = r.left + r.width / 2 - i.width / 2;
					break;
				case "bottom":
					a = r.bottom + Zu, o = r.left + r.width / 2 - i.width / 2;
					break;
				case "left":
					a = r.top + r.height / 2 - i.height / 2, o = r.left - i.width - Zu;
					break;
				case "right": a = r.top + r.height / 2 - i.height / 2, o = r.right + Zu;
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
	]), (0, M.useEffect)(() => () => {
		h.current !== null && window.clearTimeout(h.current), g.current !== null && window.clearTimeout(g.current);
	}, []);
	let b = c && typeof document < "u" ? (0, $o.createPortal)(/* @__PURE__ */ (0, N.jsxs)("div", {
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
		children: [t, a && u && /* @__PURE__ */ (0, N.jsx)(ed, { placement: n })]
	}), document.body) : null;
	return /* @__PURE__ */ (0, N.jsxs)("span", {
		ref: p,
		className: i("relative inline-flex items-center", o),
		onMouseEnter: v,
		onMouseLeave: y,
		onFocus: v,
		onBlur: y,
		children: [e, b]
	});
}
function ed({ placement: e }) {
	let t = "absolute h-1.5 w-1.5 rotate-45 border bg-background border-border";
	switch (e) {
		case "top": return /* @__PURE__ */ (0, N.jsx)("span", {
			"aria-hidden": !0,
			className: i(t, "left-1/2 -translate-x-1/2 -bottom-0.5 border-l-0 border-t-0")
		});
		case "bottom": return /* @__PURE__ */ (0, N.jsx)("span", {
			"aria-hidden": !0,
			className: i(t, "left-1/2 -translate-x-1/2 -top-0.5 border-r-0 border-b-0")
		});
		case "left": return /* @__PURE__ */ (0, N.jsx)("span", {
			"aria-hidden": !0,
			className: i(t, "top-1/2 -translate-y-1/2 -right-0.5 border-l-0 border-b-0")
		});
		case "right": return /* @__PURE__ */ (0, N.jsx)("span", {
			"aria-hidden": !0,
			className: i(t, "top-1/2 -translate-y-1/2 -left-0.5 border-r-0 border-t-0")
		});
	}
}
var td = O({
	padding: 8,
	apply({ availableHeight: e, elements: t }) {
		t.floating.style.maxHeight = `${Math.max(0, e)}px`;
	}
}), nd = (0, M.createContext)({
	getItemProps: () => ({}),
	activeIndex: null,
	setActiveIndex: () => {},
	setHasFocusInside: () => {},
	isOpen: !1
}), rd = "flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs text-popover-foreground hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50", id = "select-none px-3 pb-1 pt-2 text-[10px] font-medium uppercase tracking-wide text-muted-foreground";
function ad({ icon: e, iconColor: t, iconClassName: n }) {
	return e ? /* @__PURE__ */ (0, N.jsx)("span", {
		className: i("inline-flex shrink-0 text-base", n),
		style: t ? { color: t } : void 0,
		children: /* @__PURE__ */ (0, N.jsx)(H, { ...typeof e == "string" ? { name: e } : { icon: e } })
	}) : null;
}
function od({ item: e }) {
	let t = (0, M.useContext)(nd), n = xl(), r = tl(), i = r.index === t.activeIndex, a = () => {
		e.disabled || (e.onSelect(), n?.events.emit("click"));
	};
	return /* @__PURE__ */ (0, N.jsxs)("button", {
		ref: r.ref,
		type: "button",
		role: "menuitem",
		tabIndex: i ? 0 : -1,
		disabled: e.disabled,
		title: e.title,
		className: rd,
		...t.getItemProps({
			onClick: a,
			onFocus: () => t.setHasFocusInside(!0),
			onKeyDown: (e) => {
				(e.key === "Enter" || e.key === " ") && (e.preventDefault(), a());
			}
		}),
		children: [/* @__PURE__ */ (0, N.jsx)(ad, {
			icon: e.icon,
			iconColor: e.iconColor,
			iconClassName: e.iconClassName
		}), e.label]
	});
}
function sd({ item: e }) {
	let t = (0, M.useContext)(nd), n = xl(), r = Sl(), i = bl(), a = tl(), [o, s] = (0, M.useState)(!1), [c, l] = (0, M.useState)(null), [u, d] = (0, M.useState)(!1), f = (0, M.useRef)([]), p = gs(), { refs: m, floatingStyles: h, context: _ } = yu({
		...r ? { nodeId: r } : {},
		open: o,
		onOpenChange: s,
		placement: "right-start",
		middleware: [
			g({
				mainAxis: 0,
				alignmentAxis: -4
			}),
			T({ padding: 8 }),
			E({
				padding: 8,
				crossAxis: !0
			}),
			td
		],
		whileElementsMounted: C
	}), { getReferenceProps: v, getFloatingProps: y, getItemProps: b } = xu([
		Al(_, {
			delay: { open: 75 },
			handleClose: Pu()
		}),
		pu(_, { toggle: !1 }),
		Au(_, { role: "menu" }),
		_u(_, {
			escapeKey: !1,
			bubbles: !0
		}),
		Ou(_, {
			listRef: f,
			activeIndex: c,
			nested: !0,
			onNavigate: l,
			loop: !0
		})
	]);
	hs(o, () => _.onOpenChange(!1)), (0, M.useEffect)(() => {
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
	]), (0, M.useEffect)(() => {
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
	let x = a.index === t.activeIndex, S = Zc([m.setReference, a.ref]), w = typeof e.label == "string" ? e.label : void 0;
	return /* @__PURE__ */ (0, N.jsxs)(Cl, {
		id: r,
		children: [/* @__PURE__ */ (0, N.jsxs)("button", {
			ref: S,
			type: "button",
			role: "menuitem",
			tabIndex: x ? 0 : -1,
			disabled: e.disabled,
			title: e.title,
			"aria-haspopup": "menu",
			"aria-expanded": o,
			"data-focus-inside": u ? "" : void 0,
			className: rd,
			...t.getItemProps(v({ onFocus: () => {
				d(!1), t.setHasFocusInside(!0);
			} })),
			children: [
				/* @__PURE__ */ (0, N.jsx)(ad, {
					icon: e.icon,
					iconColor: e.iconColor,
					iconClassName: e.iconClassName
				}),
				/* @__PURE__ */ (0, N.jsx)("span", {
					className: "min-w-0 flex-1 truncate",
					children: e.label
				}),
				/* @__PURE__ */ (0, N.jsx)(H, {
					icon: Ke,
					className: "ml-auto shrink-0 opacity-60"
				})
			]
		}), /* @__PURE__ */ (0, N.jsx)(nd.Provider, {
			value: {
				getItemProps: b,
				activeIndex: c,
				setActiveIndex: l,
				setHasFocusInside: d,
				isOpen: o
			},
			children: /* @__PURE__ */ (0, N.jsx)(el, {
				elementsRef: f,
				children: o && /* @__PURE__ */ (0, N.jsx)(Ql, { children: /* @__PURE__ */ (0, N.jsx)(lu, {
					context: _,
					modal: !1,
					initialFocus: -1,
					returnFocus: !1,
					children: /* @__PURE__ */ (0, N.jsxs)("div", {
						ref: m.setFloating,
						role: "menu",
						style: {
							...h,
							zIndex: p
						},
						className: "min-w-[8rem] overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-md",
						...y(),
						"aria-label": w,
						...w ? { "aria-labelledby": void 0 } : {},
						children: [e.header != null && /* @__PURE__ */ (0, N.jsx)("div", {
							className: "border-b border-border px-3 py-1.5",
							children: e.header
						}), /* @__PURE__ */ (0, N.jsx)(cd, { items: e.children ?? [] })]
					})
				}) })
			})
		})]
	});
}
function cd({ items: e }) {
	return /* @__PURE__ */ (0, N.jsx)(N.Fragment, { children: e.map((t, n) => {
		let r = t.group != null && t.group !== e[n - 1]?.group, i = !!t.children && t.children.length > 0;
		return /* @__PURE__ */ (0, N.jsxs)(M.Fragment, { children: [r && /* @__PURE__ */ (0, N.jsx)("div", {
			role: "presentation",
			className: id,
			children: t.group
		}), i ? /* @__PURE__ */ (0, N.jsx)(sd, { item: t }) : /* @__PURE__ */ (0, N.jsx)(od, { item: t })] }, n);
	}) });
}
// ../../../clicky-ui/packages/ui/src/overlay/DropdownMenu.tsx
function ld(e) {
	return /* @__PURE__ */ (0, N.jsx)(wl, { children: /* @__PURE__ */ (0, N.jsx)(ud, { ...e }) });
}
function ud({ label: e, icon: t, hideChevron: n = !1, variant: r = "outline", size: o = "sm", trigger: s, items: c, children: l, header: u, footer: d, align: f = "right", onOpenChange: p, title: h, menuLabel: _, className: v, menuClassName: y }) {
	let [b, x] = (0, M.useState)(!1), [S, w] = (0, M.useState)(null), D = gs(), O = xl(), k = Sl(), A = (0, M.useRef)([]), j = (0, M.useRef)(p);
	j.current = p, (0, M.useEffect)(() => {
		j.current?.(b);
	}, [b]);
	let { refs: P, floatingStyles: F, context: I } = yu({
		...k ? { nodeId: k } : {},
		open: b,
		onOpenChange: x,
		placement: f === "right" ? "bottom-end" : "bottom-start",
		whileElementsMounted: C,
		middleware: [
			g(4),
			T({ padding: 8 }),
			E({ padding: 8 }),
			td
		]
	}), { getReferenceProps: L, getFloatingProps: ee, getItemProps: R } = xu([
		pu(I),
		_u(I, { escapeKey: !1 }),
		Au(I, { role: "menu" }),
		Ou(I, {
			listRef: A,
			activeIndex: S,
			onNavigate: w,
			loop: !0
		})
	]), z = () => {
		x(!1), P.domReference.current instanceof HTMLElement && P.domReference.current.focus();
	};
	hs(b, z);
	let te = (0, M.useRef)(z);
	return te.current = z, (0, M.useEffect)(() => {
		if (!O) return;
		let e = () => te.current();
		return O.events.on("click", e), () => O.events.off("click", e);
	}, [O]), /* @__PURE__ */ (0, N.jsx)(Cl, {
		id: k,
		children: /* @__PURE__ */ (0, N.jsxs)("div", {
			className: i("relative inline-flex", v),
			children: [s ? /* @__PURE__ */ (0, N.jsx)(m, {
				ref: P.setReference,
				className: "inline-flex",
				"aria-haspopup": "menu",
				"aria-expanded": b,
				...L(),
				children: s
			}) : /* @__PURE__ */ (0, N.jsxs)(a, {
				ref: P.setReference,
				type: "button",
				variant: r,
				size: o,
				title: h,
				"aria-haspopup": "menu",
				"aria-expanded": b,
				...L(),
				children: [
					t && /* @__PURE__ */ (0, N.jsx)(H, { ...typeof t == "string" ? { name: t } : { icon: t } }),
					e,
					!n && /* @__PURE__ */ (0, N.jsx)(H, { icon: Ge })
				]
			}), b && /* @__PURE__ */ (0, N.jsx)(Ql, { children: /* @__PURE__ */ (0, N.jsx)(lu, {
				context: I,
				modal: !1,
				closeOnFocusOut: !1,
				children: /* @__PURE__ */ (0, N.jsxs)("div", {
					ref: P.setFloating,
					...ee(),
					role: "menu",
					"aria-label": _,
					..._ ? { "aria-labelledby": void 0 } : {},
					style: {
						...F,
						zIndex: D
					},
					className: i("min-w-[8rem] overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-md", y),
					children: [
						u != null && /* @__PURE__ */ (0, N.jsx)("div", {
							className: "border-b border-border px-3 py-1.5",
							children: u
						}),
						l ? l(z) : /* @__PURE__ */ (0, N.jsx)(nd.Provider, {
							value: {
								getItemProps: R,
								activeIndex: S,
								setActiveIndex: w,
								setHasFocusInside: () => {},
								isOpen: b
							},
							children: /* @__PURE__ */ (0, N.jsx)(el, {
								elementsRef: A,
								children: /* @__PURE__ */ (0, N.jsx)(cd, { items: c ?? [] })
							})
						}),
						d != null && /* @__PURE__ */ (0, N.jsx)("div", {
							className: "border-t border-border px-3 py-1.5",
							children: d
						})
					]
				})
			}) })]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/overlay/confirmMemory.ts
var dd = "clicky-ui:confirm:", fd = "confirmed", pd = /* @__PURE__ */ new Set();
function md() {
	if (typeof window > "u") return null;
	try {
		let e = window.localStorage;
		return e.getItem(dd), e;
	} catch {
		return null;
	}
}
function hd() {
	return md() !== null;
}
function gd(e) {
	let t = md();
	if (!t) throw Error(`cannot remember confirmation "${e}": localStorage is unavailable`);
	t.setItem(dd + e, fd), _d();
}
function _d() {
	for (let e of pd) e();
}
// ../../../clicky-ui/packages/ui/src/overlay/ConfirmDialog.tsx
function vd({ open: e, title: t, message: n, warning: r, confirmLabel: i = "Confirm", cancelLabel: o = "Cancel", variant: s = "default", remember: c, onConfirm: l, onCancel: u }) {
	let [f, p] = (0, M.useState)(!1);
	(0, M.useEffect)(() => {
		e && p(!1);
	}, [e]);
	let m = c !== void 0 && hd();
	return /* @__PURE__ */ (0, N.jsx)(ws, {
		open: e,
		onClose: u,
		title: t,
		size: "sm",
		expandable: !1,
		footer: /* @__PURE__ */ (0, N.jsxs)("div", {
			className: "flex justify-end gap-density-2",
			children: [/* @__PURE__ */ (0, N.jsx)(a, {
				type: "button",
				variant: "ghost",
				size: "sm",
				onClick: u,
				children: o
			}), /* @__PURE__ */ (0, N.jsx)(a, {
				type: "button",
				variant: s,
				size: "sm",
				onClick: () => {
					m && f && gd(c.key), l();
				},
				children: i
			})]
		}),
		children: /* @__PURE__ */ (0, N.jsxs)("div", {
			className: "space-y-density-3",
			children: [
				r ? /* @__PURE__ */ (0, N.jsx)(d, {
					variant: "warning",
					className: "my-0",
					children: r
				}) : null,
				n ? /* @__PURE__ */ (0, N.jsx)("p", {
					className: "text-sm text-muted-foreground",
					children: n
				}) : null,
				m ? /* @__PURE__ */ (0, N.jsxs)("label", {
					className: "flex items-center gap-2 text-sm",
					children: [/* @__PURE__ */ (0, N.jsx)("input", {
						type: "checkbox",
						className: "size-4 shrink-0 rounded border-input",
						checked: f,
						onChange: (e) => p(e.target.checked)
					}), c.label ?? "Remember my choice"]
				}) : null
			]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/TabButton.tsx
var yd = {
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
function bd({ active: e, onClick: t, label: n, icon: r, count: a, countColor: o = "bg-muted-foreground", variant: s = "pill", className: c }) {
	let l = yd[s];
	return /* @__PURE__ */ (0, N.jsxs)("button", {
		type: "button",
		role: "tab",
		"aria-selected": e,
		onClick: t,
		className: i("inline-flex items-center gap-1.5 px-density-3 py-density-1.5 text-sm transition-colors", l.base, e ? l.active : l.inactive, c),
		children: [
			r && /* @__PURE__ */ (0, N.jsx)(H, { ...typeof r == "string" ? { name: r } : { icon: r } }),
			/* @__PURE__ */ (0, N.jsx)("span", { children: n }),
			a !== void 0 && a > 0 && /* @__PURE__ */ (0, N.jsx)("span", {
				className: i("inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded-full text-[10px] font-bold text-white", o),
				children: a
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/test-runner/filterState.ts
function xd() {
	return {
		status: /* @__PURE__ */ new Map(),
		framework: /* @__PURE__ */ new Map()
	};
}
function Sd(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) {
		let e = n.startsWith("!") ? n.slice(1) : n;
		e && t.set(e, n.startsWith("!") ? "exclude" : "include");
	}
	return t;
}
function Cd(e) {
	return Array.from(e.entries()).sort(([e], [t]) => String(e).localeCompare(String(t))).map(([e, t]) => t === "exclude" ? `!${e}` : String(e));
}
function wd(e, t) {
	let n = new Map(e), r = n.get(t);
	return r === "include" ? n.set(t, "exclude") : r === "exclude" ? n.delete(t) : n.set(t, "include"), n;
}
function Td(e, t) {
	if (t.size === 0) return !0;
	let n = /* @__PURE__ */ new Set(), r = /* @__PURE__ */ new Set();
	for (let [e, i] of t.entries()) i === "include" && n.add(e), i === "exclude" && r.add(e);
	return e && r.has(e) ? !1 : n.size === 0 || !!e && n.has(e);
}
// ../../../clicky-ui/packages/ui/src/components/clipboard.ts
function Ed(e, t, n) {
	return e.then((e) => {
		let r = n(e);
		if (typeof r != "string") throw Error(`Rich copy payload has no ${t === "text/html" ? "html" : "text"} string`);
		return new Blob([r], { type: t });
	});
}
async function Dd(e) {
	let t = Promise.resolve(e);
	await kd("Rich copy", {
		"text/html": Ed(t, "text/html", (e) => e.html),
		"text/plain": Ed(t, "text/plain", (e) => e.text)
	});
}
async function Od(e) {
	await kd("Text copy", { "text/plain": e.then((e) => {
		if (typeof e != "string") throw Error("Text copy payload is not a string");
		return new Blob([e], { type: "text/plain" });
	}) });
}
async function kd(e, t) {
	if (typeof ClipboardItem > "u") throw Error(`${e} is unsupported: this browser has no ClipboardItem`);
	if (typeof navigator > "u" || typeof navigator.clipboard?.write != "function") throw Error(`${e} is unsupported: navigator.clipboard.write is unavailable (insecure origin or blocked by the host)`);
	for (let e of Object.values(t)) e.catch(() => void 0);
	await navigator.clipboard.write([new ClipboardItem(t)]), await Promise.all(Object.values(t));
}
async function Ad(e) {
	if (typeof navigator < "u" && navigator.clipboard?.writeText) try {
		await navigator.clipboard.writeText(e);
		return;
	} catch {}
	if (typeof document > "u" || typeof document.execCommand != "function") throw Error("Browser rejected the copy request");
	let t = document.createElement("textarea");
	t.value = e, t.style.position = "fixed", t.style.opacity = "0", document.body.appendChild(t), t.focus(), t.select();
	let n = !1;
	try {
		n = document.execCommand("copy");
	} catch {
		n = !1;
	} finally {
		document.body.removeChild(t);
	}
	if (!n) throw Error("Browser rejected the copy request");
}
function jd(e = 1500) {
	let [t, n] = (0, M.useState)("idle"), r = (0, M.useRef)(null), i = (0, M.useRef)(0);
	(0, M.useEffect)(() => () => {
		i.current += 1, r.current !== null && clearTimeout(r.current);
	}, []);
	let a = (0, M.useCallback)(() => {
		r.current !== null && clearTimeout(r.current), r.current = setTimeout(() => {
			r.current = null, n("idle");
		}, e);
	}, [e]);
	return {
		state: t,
		copy: (0, M.useCallback)((e) => {
			let t = ++i.current;
			if (typeof e == "string") {
				Ad(e).then(() => t === i.current && n("copied"), () => t === i.current && n("error")), a();
				return;
			}
			r.current !== null && clearTimeout(r.current), r.current = null, n("pending");
			let o = (e) => {
				t === i.current && (n(e), a());
			};
			("plain" in e ? Od(e.plain) : Dd(e)).then(() => o("copied"), () => o("error"));
		}, [a])
	};
}
export { it as $, es as A, Ut as B, yu as C, ws as D, Au as E, Bo as F, gt as G, Vt as H, ma as I, lt as J, dt as K, Gi as L, Qo as M, Wo as N, hs as O, Vo as P, at as Q, Ui as R, _u as S, Zc as T, Ot as U, Ht as V, vt as W, st as X, ct as Y, ot as Z, Fu as _, H as _t, xd as a, Qe as at, Ql as b, L as bt, bd as c, Je as ct, $u as d, Ke as dt, rt as et, Gu as f, Ge as ft, Xu as g, Fe as gt, Ku as h, Re as ht, Sd as i, $e as it, ts as j, gs as k, vd as l, K as lt, Ju as m, Ie as mt, jd as n, tt as nt, Cd as o, Ze as ot, Yu as p, We as pt, ut as q, wd as r, et as rt, Td as s, Ye as st, Ad as t, nt as tt, ld as u, qe as ut, zu as v, ne as vt, xu as w, pu as x, F as xt, lu as y, I as yt, Jt as z };
