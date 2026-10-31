import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./react-DHvKIBht.js";
import { Br as r, Cn as i, En as a, Er as o, Si as s, Tn as c, gn as l, i as u, t as d, v as f, vn as p } from "./SpecRuntimeBar-CZVuNKUs.js";
import { t as m } from "./button-G86_a8z9.js";
// ../../../clicky-ui/packages/ui/src/data/FixtureEditor/fixture-frontmatter.ts
var h = /* @__PURE__ */ e(n(), 1), g = /* @__PURE__ */ new Set([
	"mode",
	"budget",
	"effort",
	"id",
	"model",
	"noCache",
	"prompt",
	"temperature"
]), _ = /* @__PURE__ */ new Set([
	"mode",
	"budget",
	"effort",
	"id",
	"model",
	"noCache",
	"temperature"
]), v = /* @__PURE__ */ new Set(["maxConcurrent", "cacheTTL"]);
function y(e) {
	let t = M(e);
	if (!t.hasFrontmatter) return {
		ok: !0,
		hasFrontmatter: !1,
		frontmatter: {},
		raw: "",
		body: e
	};
	if (!t.closed) return {
		ok: !1,
		hasFrontmatter: !0,
		error: "Frontmatter closing delimiter not found",
		raw: t.raw,
		body: t.body
	};
	let n = F(t.raw);
	return n.ok ? {
		ok: !0,
		hasFrontmatter: !0,
		frontmatter: n.value,
		raw: t.raw,
		body: t.body
	} : {
		ok: !1,
		hasFrontmatter: !0,
		error: n.error,
		raw: t.raw,
		body: t.body
	};
}
function b(e) {
	let t = R(e.ai), n = w(t), r = k(e.env);
	return r.length > 0 && (n.setup = {
		...n.setup,
		envVars: r
	}), {
		runtime: n,
		aiExtras: D(t)
	};
}
function x(e, t, n = {}) {
	let r = y(e), i = r.body;
	return j(C(r.ok ? r.frontmatter : {}, t, n), i);
}
function S(e, t) {
	let n = M(e), r = n.hasFrontmatter ? n.body : e;
	return t.trim() === "" ? r : `---\n${t.endsWith("\n") ? t : `${t}\n`}---\n${r}`;
}
function C(e, t, n = {}) {
	let r = L(e, ["ai", "env"]), i = {
		...I(R(e.ai)),
		...T(t.runtime, n),
		...O(t.aiExtras)
	}, a = A(t.runtime.setup?.envVars);
	return H(i) && (r.ai = i), H(a) && (r.env = a), r;
}
function w(e) {
	let t = {}, n = B(e.model);
	n && (t.model = n);
	let r = B(e.mode);
	r && (t.mode = r);
	let i = B(e.id);
	i && (t.id = i);
	let a = B(e.effort);
	a && (t.effort = a);
	let o = V(e.temperature);
	o != null && (t.temperature = o), e.noCache === !0 && (t.noCache = !0);
	let s = R(e.budget), c = V(e.maxTokens) ?? V(s.maxTokens), l = V(s.cost), u = V(s.maxTurns), d = B(s.timeout);
	(c != null || l != null || u != null || d) && (t.budget = {
		...l == null ? {} : { cost: l },
		...c == null ? {} : { maxTokens: c },
		...u == null ? {} : { maxTurns: u },
		...d ? { timeout: d } : {}
	});
	let f = E(e.prompt);
	return f && (t.prompt = f), t;
}
function T(e, t = {}) {
	let n = f(e), r = {};
	for (let e of g) {
		if (t.inheritModel && _.has(e)) continue;
		let i = n[e];
		i != null && (e !== "budget" || H(R(i))) && (r[e] = i);
	}
	let i = n.budget?.maxTokens;
	return i != null && !t.inheritModel && (delete r.budget, r.maxTokens = i), r;
}
function E(e) {
	if (typeof e == "string") {
		let t = B(e);
		return t ? { user: t } : void 0;
	}
	let t = R(e), n = B(t.user), r = B(t.system), i = B(t.appendSystem), a = B(t.source), o = z(t.metadata), s = {
		...n ? { user: n } : {},
		...r ? { system: r } : {},
		...i ? { appendSystem: i } : {},
		...a ? { source: a } : {},
		...o ? { metadata: o } : {}
	};
	return H(s) ? s : void 0;
}
function D(e) {
	let t = V(e.maxConcurrent), n = B(e.cacheTTL);
	return {
		...t == null ? {} : { maxConcurrent: t },
		...n ? { cacheTTL: n } : {}
	};
}
function O(e) {
	return {
		...V(e.maxConcurrent) == null ? {} : { maxConcurrent: V(e.maxConcurrent) },
		...B(e.cacheTTL) ? { cacheTTL: B(e.cacheTTL) } : {}
	};
}
function k(e) {
	if (!U(e)) return [];
	let t = [];
	for (let [n, r] of Object.entries(e)) {
		let e = n.trim();
		e && typeof r == "string" && t.push({
			name: e,
			value: r
		});
	}
	return t;
}
function A(e) {
	let t = {};
	for (let n of e ?? []) {
		let e = B(n.name);
		if (!e) continue;
		let r = typeof n.value == "string" ? n.value : void 0, i = typeof n.valueFrom == "string" ? n.valueFrom : void 0;
		r == null ? i && (t[e] = i) : t[e] = r;
	}
	return t;
}
function j(e, t) {
	if (!H(e)) return t;
	let n = a(e, {
		lineWidth: 0,
		collectionStyle: "block"
	});
	return `---\n${n.endsWith("\n") ? n : `${n}\n`}---\n${t}`;
}
function M(e) {
	let t = N(e);
	if (P(t[0] ?? "") !== "---") return { hasFrontmatter: !1 };
	for (let e = 1; e < t.length; e += 1) {
		let n = P(t[e] ?? "");
		if (n === "---" || n === "...") return {
			hasFrontmatter: !0,
			closed: !0,
			raw: t.slice(1, e).join(""),
			body: t.slice(e + 1).join("")
		};
	}
	return {
		hasFrontmatter: !0,
		closed: !1,
		raw: t.slice(1).join(""),
		body: ""
	};
}
function N(e) {
	if (e === "") return [];
	let t = e.match(/[^\n]*(?:\n|$)/g) ?? [];
	return t.at(-1) === "" && t.pop(), t;
}
function P(e) {
	return e.replace(/\r?\n$/, "").trim();
}
function F(e) {
	if (e.trim() === "") return {
		ok: !0,
		value: {}
	};
	let t = c(e, { prettyErrors: !1 }), n = t.errors[0];
	if (n) return {
		ok: !1,
		error: n.message
	};
	let r = t.toJSON();
	return r == null ? {
		ok: !0,
		value: {}
	} : U(r) ? {
		ok: !0,
		value: r
	} : {
		ok: !1,
		error: "Frontmatter root must be an object"
	};
}
function I(e) {
	return L(e, [.../* @__PURE__ */ new Set([
		...g,
		...v,
		"maxTokens"
	])]);
}
function L(e, t) {
	let n = new Set(t);
	return Object.fromEntries(Object.entries(e).filter(([e]) => !n.has(e)));
}
function R(e) {
	return U(e) ? e : {};
}
function z(e) {
	if (!U(e)) return;
	let t = {};
	for (let [n, r] of Object.entries(e)) {
		let e = n.trim();
		e && typeof r == "string" && (t[e] = r);
	}
	return H(t) ? t : void 0;
}
function B(e) {
	return typeof e == "string" && e.trim() ? e.trim() : void 0;
}
function V(e) {
	return typeof e == "number" && Number.isFinite(e) ? e : void 0;
}
function H(e) {
	return Object.keys(e).length > 0;
}
function U(e) {
	return typeof e == "object" && !!e && !Array.isArray(e);
}
// ../../../clicky-ui/packages/ui/src/data/FixtureEditor/FixtureFrontmatterDialog.tsx
var W = /* @__PURE__ */ e(t(), 1), G = {
	runtime: {},
	aiExtras: {}
}, K = [{
	id: "same",
	label: "Use same model"
}, {
	id: "different",
	label: "Use different model for verification"
}], q = ["model", "prompt"], J = [
	"model",
	"prompt",
	"environment"
], Y = ["prompt", "environment"], X = ["environment"];
function Z({ open: e, markdown: t, onChange: n, onClose: a, options: c }) {
	let f = (0, h.useMemo)(() => y(t), [t]), g = (0, h.useMemo)(() => f.ok ? b(f.frontmatter) : G, [f]), _ = c?.mode === "verification", v = (0, h.useMemo)(() => Q(g.runtime) ? "different" : "same", [g.runtime]), [C, w] = (0, h.useState)(g), [T, E] = (0, h.useState)(f.raw), [D, O] = (0, h.useState)(v);
	(0, h.useEffect)(() => {
		e && (w(g), E(f.raw), O(v));
	}, [
		g,
		v,
		e,
		f.raw
	]);
	let k = f.ok ? JSON.stringify(C) !== JSON.stringify(g) || _ && D !== v : T !== f.raw, A = () => {
		n(f.ok ? x(t, C, { inheritModel: _ && D === "same" }) : S(t, T)), a();
	}, j = _ && D === "same" ? Y : _ ? J : q, M = _ ? /* @__PURE__ */ (0, W.jsx)("div", {
		className: "pb-density-3",
		children: /* @__PURE__ */ (0, W.jsx)(l, {
			"aria-label": "Verification model",
			value: D,
			options: K,
			onChange: (e) => O(e),
			size: "lg",
			wrap: !0,
			className: "w-full"
		})
	}) : void 0;
	return /* @__PURE__ */ (0, W.jsx)(p, {
		open: e,
		onClose: a,
		confirmClose: k,
		title: "Fixture frontmatter",
		size: "2xl",
		className: "h-[92vh]",
		footer: /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "flex items-center justify-between gap-3",
			children: [/* @__PURE__ */ (0, W.jsx)("span", {
				className: "text-xs text-muted-foreground",
				children: f.ok ? "ai / env" : "YAML source"
			}), /* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex items-center gap-2",
				children: [/* @__PURE__ */ (0, W.jsx)(m, {
					type: "button",
					variant: "outline",
					size: "sm",
					onClick: a,
					children: "Cancel"
				}), /* @__PURE__ */ (0, W.jsxs)(m, {
					type: "button",
					size: "sm",
					onClick: A,
					children: [/* @__PURE__ */ (0, W.jsx)(s, {
						icon: r,
						className: "size-4"
					}), "Save frontmatter"]
				})]
			})]
		}),
		children: f.ok ? /* @__PURE__ */ (0, W.jsx)("div", {
			className: "min-h-0",
			children: /* @__PURE__ */ (0, W.jsx)(u, {
				value: C.runtime,
				onChange: (e) => w((t) => ({
					...t,
					runtime: e
				})),
				models: c?.models,
				families: c?.families,
				tools: c?.tools,
				permissionCatalog: c?.permissionCatalog,
				secretSelector: c?.secretSelector,
				cliOptions: c?.cliOptions,
				sections: j,
				showHeader: !1,
				beforeSections: /* @__PURE__ */ (0, W.jsxs)("div", {
					className: "space-y-density-3",
					children: [M, j.includes("model") && /* @__PURE__ */ (0, W.jsx)(d, {
						value: C.runtime,
						onChange: (e) => w((t) => ({
							...t,
							runtime: e
						})),
						models: c?.models,
						families: c?.families,
						sections: j
					})]
				}),
				defaultCollapsedSections: _ ? X : void 0
			})
		}) : /* @__PURE__ */ (0, W.jsxs)("div", {
			className: "grid gap-3",
			children: [/* @__PURE__ */ (0, W.jsxs)("div", {
				className: "flex items-center gap-2",
				children: [/* @__PURE__ */ (0, W.jsx)(i, {
					variant: "status",
					status: "error",
					size: "xs",
					clickToCopy: !1,
					children: "YAML error"
				}), /* @__PURE__ */ (0, W.jsx)("span", {
					className: "text-xs text-muted-foreground",
					children: f.error
				})]
			}), /* @__PURE__ */ (0, W.jsxs)("label", {
				className: "grid gap-2 text-xs text-muted-foreground",
				children: [/* @__PURE__ */ (0, W.jsxs)("span", {
					className: "inline-flex items-center gap-1.5",
					children: [/* @__PURE__ */ (0, W.jsx)(s, {
						icon: o,
						className: "size-4"
					}), "Frontmatter source"]
				}), /* @__PURE__ */ (0, W.jsx)("textarea", {
					"aria-label": "Frontmatter source",
					value: T,
					onChange: (e) => E(e.currentTarget.value),
					spellCheck: !1,
					className: "min-h-80 w-full resize-y rounded-md border border-input bg-background px-3 py-2 font-mono text-xs leading-5 text-foreground shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
				})]
			})]
		})
	});
}
function Q(e) {
	return !!(e.model || e.mode || e.id || e.effort || e.temperature != null || e.noCache || e.budget);
}
export { Z as FixtureFrontmatterDialog };
