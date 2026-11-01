import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./jsx-runtime-f3E3oPuc.js";
import { Et as r, bt as i, l as a, o, xt as s } from "./clipboard-CMVj3cUp.js";
import { B as c, O as l, a as u, i as d, n as f, o as ee, r as p, t as m } from "./paths-D_tC_cYe.js";
import { a as te, c as ne, m as re, o as h, p as g, s as ie, u as _ } from "./Tabs-Dd-b3nbz.js";
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiError.tsx
var v = /* @__PURE__ */ e(t(), 1), y = n(), ae = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#dc2626",
		...r
	};
	return /* @__PURE__ */ (0, y.jsx)("svg", {
		xmlns: "http://www.w3.org/2000/svg",
		width: e,
		height: e,
		viewBox: "0 0 32 32",
		role: n ? "img" : "presentation",
		"aria-label": n,
		"aria-hidden": !n || void 0,
		className: t,
		style: a,
		...i,
		children: /* @__PURE__ */ (0, y.jsx)("path", {
			fill: "currentColor",
			d: "M31.86 29.023 16.554 2.516c-.23-.395-.88-.395-1.11 0L.085 29.12a.639.639 0 0 0 .556.957h30.73c.356 0 .64-.281.64-.637 0-.156-.054-.3-.152-.418M15.035 12.328c0-.137.094-.226.23-.226h1.47c.136 0 .226.09.226.226v9.258c0 .137-.09.23-.227.23h-1.468a.21.21 0 0 1-.168-.066.212.212 0 0 1-.063-.164Zm1.977 13.23c0 .141-.094.231-.23.231h-1.563c-.137 0-.227-.094-.227-.23v-1.676c0-.137.094-.227.227-.227h1.562c.14 0 .23.09.23.227Zm0 0"
		})
	});
}, {
	__source: "incumbent",
	__viewBox: "0 0 32 32",
	__group: "health-status",
	__consumerName: "error",
	__defaultColor: "#dc2626",
	displayName: "UiError"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/test-runner/routePath.ts
function b(e) {
	let t = (e || "").toLowerCase().replace(/[^a-z0-9]+/g, "-");
	return t.startsWith("-") && (t = t.slice(1)), t.endsWith("-") && (t = t.slice(0, -1)), t || "node";
}
function x(e, t = []) {
	let n = /* @__PURE__ */ new Map();
	for (let t of e) {
		let e = b(t.name);
		n.set(e, (n.get(e) || 0) + 1);
	}
	let r = /* @__PURE__ */ new Map();
	return e.map((e) => {
		let i = b(e.name), a = (r.get(i) || 0) + 1;
		r.set(i, a);
		let o = (n.get(i) || 0) > 1 ? `${i}~${a}` : i, s = [...t, o], c = {
			...e,
			route_path: s.join("/")
		};
		return e.children && (c.children = x(e.children, s)), c;
	});
}
function S(e, t) {
	if (!t) return null;
	for (let n of e) {
		if (n.route_path === t) return n;
		if (n.children) {
			let e = S(n.children, t);
			if (e) return e;
		}
	}
	return null;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/use-problem-expansion.ts
function oe(e) {
	let [t, n] = (0, v.useState)([]);
	return (0, v.useEffect)(() => {
		n((t) => {
			let n = t.filter((t) => e.includes(t));
			return n.length === t.length ? t : n;
		});
	}, [e]), [t, n];
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/ProblemRow.tsx
function se({ message: e, severity: t, metadata: n, location: a, actions: o, children: c, expanded: l, onExpandedChange: u }) {
	let [d, f] = (0, v.useState)(!1), p = l ?? d;
	return /* @__PURE__ */ (0, y.jsxs)("div", {
		className: "border-b border-border/50",
		"data-problem-row": "",
		children: [/* @__PURE__ */ (0, y.jsxs)("div", {
			className: "flex min-w-0 items-center hover:bg-muted/50",
			children: [/* @__PURE__ */ (0, y.jsxs)("button", {
				type: "button",
				"aria-expanded": p,
				title: e,
				onClick: () => l === void 0 ? f(!p) : u?.(!p),
				className: "flex min-h-7 min-w-0 flex-1 items-center gap-2 px-2 py-1 text-left text-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring [@media(pointer:coarse)]:min-h-11",
				children: [
					/* @__PURE__ */ (0, y.jsx)(r, {
						icon: p ? s : i,
						className: "shrink-0 text-muted-foreground"
					}),
					/* @__PURE__ */ (0, y.jsx)(ee, { severity: t }),
					/* @__PURE__ */ (0, y.jsx)("span", {
						className: "min-w-0 flex-1 truncate",
						children: e
					}),
					/* @__PURE__ */ (0, y.jsx)("span", {
						className: "hidden max-w-[30%] truncate text-muted-foreground sm:block",
						children: n
					}),
					/* @__PURE__ */ (0, y.jsx)("span", {
						className: "max-w-[25%] shrink-0 truncate font-mono text-muted-foreground",
						"data-problem-ln": "",
						children: a
					})
				]
			}), /* @__PURE__ */ (0, y.jsx)(re, { items: o })]
		}), p && /* @__PURE__ */ (0, y.jsxs)("div", {
			className: "space-y-2 px-3 pb-3 pl-8 text-xs",
			children: [
				/* @__PURE__ */ (0, y.jsx)("p", {
					className: "whitespace-pre-wrap break-words",
					children: e
				}),
				/* @__PURE__ */ (0, y.jsx)("div", {
					className: "text-muted-foreground sm:hidden",
					children: n
				}),
				c
			]
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/grouping.ts
var ce = "(no rule)", le = 5;
function C(e) {
	return e.rule?.method || ce;
}
function w(e) {
	return e.slice().sort((e, t) => (e.line ?? 0) - (t.line ?? 0));
}
function ue(e) {
	let t = e.indexOf("\n");
	return t === -1 ? e : e.slice(0, t);
}
function T(e, t) {
	let n = /* @__PURE__ */ new Map();
	for (let r of e) {
		let e = n.get(t(r));
		e ? e.push(r) : n.set(t(r), [r]);
	}
	return n;
}
function E(e) {
	return Array.from(e.keys()).sort();
}
function D(e, t) {
	let n = e.filter((e) => o(e.linter, t.linter)), r = new Set(n.flatMap((e) => e.work_dir ? [p(e.work_dir)] : [])).size > 1, i = [];
	for (let e of n) for (let n of e.violations ?? []) {
		if (!o(_(n), t.severity)) continue;
		let a = d(n.file, e.work_dir), s = f(a, e.work_dir, r);
		i.push({
			linter: e.linter,
			result: e,
			v: {
				...n,
				raw_file: a,
				file: s
			},
			file: s,
			rawFile: a,
			workDir: e.work_dir
		});
	}
	return i;
}
function O(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.has(n.linter) || t.set(n.linter, n);
	return t;
}
function k() {
	return {
		name: "",
		path: void 0,
		rawPath: "",
		folders: /* @__PURE__ */ new Map(),
		leaves: /* @__PURE__ */ new Map()
	};
}
function A(e, t, n, r) {
	let i = m(t), a = m(n);
	if (i.length === 0 || a.length === 0) throw Error(`lint: cannot place violation, empty path (file=${JSON.stringify(t)}, raw=${JSON.stringify(n)})`);
	let o = i.slice(0, -1), s = a.slice(0, -1), c = o.length - s.length, l = e;
	o.forEach((e, t) => {
		let n = l.folders.get(e);
		if (!n) {
			let r = Math.max(0, Math.min(s.length, t + 1 - c));
			n = {
				name: e,
				path: l.path ? `${l.path}/${e}` : e,
				rawPath: s.slice(0, r).join("/"),
				folders: /* @__PURE__ */ new Map(),
				leaves: /* @__PURE__ */ new Map()
			}, l.folders.set(e, n);
		}
		l = n;
	});
	let u = i[i.length - 1], d = l.leaves.get(u);
	return d || (d = r(), l.leaves.set(u, d)), d;
}
function j(e, t, n) {
	let r = n.some((e) => e.failed);
	return {
		name: e,
		kind: t,
		children: n,
		failed: r,
		passed: !r,
		skipped: !1
	};
}
function M(e) {
	let t = new Set(e.flatMap((e) => e.work_dir ? [e.work_dir] : []));
	return t.size === 1 ? Array.from(t)[0] : void 0;
}
function N(e, t) {
	if (t.length === 0) return e;
	let n = e.failed || g(t) !== "info";
	return {
		...e,
		noFileViolations: w(t),
		failed: n,
		passed: !n
	};
}
function P(e) {
	let t = g(e.violations) !== "info";
	return {
		name: e.name,
		kind: e.kind,
		file: e.file,
		target_path: e.rawFile,
		work_dir: e.workDir,
		linterName: e.linterName,
		violations: w(e.violations),
		failed: t,
		passed: !t,
		skipped: !1
	};
}
function F(e, t, n, r = []) {
	return N({
		...j(e, "linter", n),
		linter: t,
		linterName: e
	}, r);
}
function de(e) {
	let t = ne(e);
	return {
		name: ue(t),
		kind: "linter-error",
		linterName: e.linter,
		linter: e,
		work_dir: e.work_dir,
		failed: !0,
		passed: !1,
		skipped: !1,
		message: t
	};
}
function I(e, t, n) {
	let r = E(e.folders).map((r) => {
		let i = e.folders.get(r), a = j(i.name, "lint-folder", I(i, t, n));
		return {
			...a,
			file: i.path,
			target_path: i.rawPath,
			linterName: t,
			...n ? { ruleName: n } : {},
			work_dir: M(a.children)
		};
	}), i = E(e.leaves).map((r) => {
		let i = e.leaves.get(r), a = P({
			name: r,
			kind: "lint-file",
			file: e.path ? `${e.path}/${r}` : r,
			rawFile: i.rawFile,
			linterName: t,
			violations: i.violations,
			workDir: i.workDir
		});
		return n ? {
			...a,
			ruleName: n
		} : a;
	});
	return [...r, ...i];
}
function L(e, t) {
	let n = E(e.folders).map((n) => {
		let r = e.folders.get(n), i = j(r.name, "lint-folder", L(r, t));
		return {
			...i,
			file: r.path,
			target_path: r.rawPath,
			work_dir: M(i.children)
		};
	}), r = E(e.leaves).map((n) => {
		let r = e.leaves.get(n), i = e.path ? `${e.path}/${n}` : n, a = E(r).map((e) => {
			let { workDir: n, rules: a } = r.get(e), o = E(a).map((t) => {
				let r = a.get(t);
				return {
					...P({
						name: t,
						kind: "lint-rule",
						file: i,
						rawFile: r[0]?.raw_file ?? i,
						linterName: e,
						violations: r,
						workDir: n
					}),
					ruleName: t
				};
			});
			return {
				...F(e, t.get(e), o),
				file: i,
				target_path: o[0]?.target_path,
				work_dir: n
			};
		});
		return {
			...j(n, "lint-file", a),
			file: i,
			target_path: a[0]?.target_path,
			work_dir: M(a)
		};
	});
	return [...n, ...r];
}
function R(e) {
	let t = k();
	for (let n of e) n.file && A(t, n.file, n.rawFile ?? n.file, () => ({
		rawFile: n.rawFile ?? n.file,
		workDir: n.workDir,
		violations: []
	})).violations.push(n.v);
	return t;
}
function fe(e, t) {
	let n = O(e), r = T(D(e, t), (e) => e.linter);
	return Array.from(r.keys()).sort().map((e) => {
		let t = T(r.get(e), (e) => C(e.v)), i = Array.from(t.keys()).sort().map((r) => {
			let i = t.get(r), a = I(R(i), e, r);
			return N({
				...j(r, "lint-rule-group", a),
				linterName: e,
				ruleName: r,
				work_dir: M(a) ?? n.get(e)?.work_dir
			}, i.filter((e) => !e.file).map((e) => e.v));
		});
		return F(e, n.get(e), i);
	});
}
function z(e, t) {
	let n = O(e), r = T(D(e, t), (e) => e.linter);
	return Array.from(r.keys()).sort().map((e) => {
		let t = r.get(e);
		return F(e, n.get(e), I(R(t), e), t.filter((e) => !e.file).map((e) => e.v));
	});
}
function B(e, t) {
	let n = O(e), r = D(e, t), i = k();
	for (let e of r) {
		if (!e.file) continue;
		let t = A(i, e.file, e.rawFile ?? e.file, () => /* @__PURE__ */ new Map()), n = t.get(e.linter);
		n || (n = {
			workDir: e.workDir,
			rules: /* @__PURE__ */ new Map()
		}, t.set(e.linter, n));
		let r = C(e.v);
		n.rules.set(r, [...n.rules.get(r) ?? [], e.v]);
	}
	let a = T(r.filter((e) => !e.file), (e) => e.linter), o = Array.from(a.keys()).sort().map((e) => {
		let t = T(a.get(e), (e) => C(e.v)), r = n.get(e)?.work_dir, i = Array.from(t.keys()).sort().map((n) => ({
			...P({
				name: n,
				kind: "lint-rule",
				file: "",
				rawFile: "",
				linterName: e,
				violations: t.get(n).map((e) => e.v),
				workDir: r
			}),
			ruleName: n,
			file: void 0,
			target_path: void 0
		}));
		return F(e, n.get(e), i);
	});
	return [...L(i, n), ...o];
}
function V(e, t) {
	if (!e) return "(no file)";
	if (t.length > 1) return `${e} (${t.length})`;
	let n = t[0];
	return n?.line ? n.column ? `${e}:${n.line}:${n.column}` : `${e}:${n.line}` : e;
}
function pe(e, t, n) {
	let r = T(n, (e) => e.file ?? ""), i = Array.from(r.keys()), a = i.slice(0, le).map((n) => {
		let i = r.get(n);
		return {
			...P({
				name: V(i[0].file, i.map((e) => e.v)),
				kind: "lint-file",
				file: i[0].file ?? "",
				rawFile: i[0].rawFile ?? "",
				linterName: t,
				violations: i.map((e) => e.v),
				workDir: i[0].workDir
			}),
			file: i[0].file,
			target_path: i[0].rawFile,
			ruleName: e
		};
	}), o = i.length - a.length;
	o > 0 && a.push({
		name: `… ${o} more file${o === 1 ? "" : "s"}`,
		kind: "lint-file",
		linterName: t,
		ruleName: e,
		failed: !1,
		passed: !0,
		skipped: !1
	});
	let s = n.find((e) => e.v.message)?.v.message;
	return {
		...j(s ? `${e} (${n.length}) — ${s}` : `${e} (${n.length})`, "lint-rule-group", a),
		linterName: t,
		ruleName: e
	};
}
function me(e, t) {
	let n = O(e), r = T(D(e, t), (e) => e.linter), i = e.filter((e) => o(e.linter, t.linter)), a = new Map(i.filter((e) => e.skipped).map((e) => [e.linter, e.error || "skipped"]));
	return Array.from(new Set(i.map((e) => e.linter))).sort().map((e) => {
		let t = r.get(e) ?? [];
		if (t.length === 0 && a.has(e)) return {
			name: e,
			kind: "linter",
			linter: n.get(e),
			linterName: e,
			skipped: !0,
			passed: !1,
			failed: !1,
			message: `skipped: ${a.get(e)}`
		};
		let i = T(t, (e) => C(e.v)), o = Array.from(i.keys()).sort((e, t) => i.get(t).length - i.get(e).length || e.localeCompare(t));
		return F(e, n.get(e), o.map((t) => pe(t, e, i.get(t))));
	});
}
function H(e, t, n) {
	let r = T(t.filter((e) => te(e) && o(e.linter, n.linter)), (e) => e.linter);
	if (r.size === 0) return e;
	let i = new Map(e.filter((e) => e.kind === "linter").map((e) => [e.name, e]));
	for (let [e, t] of r) {
		let n = t.map(de), r = i.get(e);
		i.set(e, r ? {
			...r,
			children: [...n, ...r.children ?? []],
			failed: !0,
			passed: !1
		} : F(e, t[0], n));
	}
	return [...e.filter((e) => e.kind !== "linter"), ...E(i).map((e) => i.get(e))];
}
function U(e, t = /* @__PURE__ */ new Set()) {
	for (let n of e) n.kind === "linter" && n.linterName && t.add(n.linterName), U(n.children ?? [], t);
	return t;
}
function he(e, t, n) {
	let r = U(e), i = /* @__PURE__ */ new Map();
	for (let e of t) h(e) && o(e.linter, n.linter) && !r.has(e.linter) && i.set(e.linter, e);
	let a = (e) => e.map((e) => {
		let t = e.children ? a(e.children) : void 0, n = e.kind === "linter" && e.linter !== void 0 && h(e.linter);
		return {
			...e,
			...t ? { children: t } : {},
			...n ? { passed: !1 } : {}
		};
	});
	if (i.size === 0) return a(e);
	let s = Array.from(i.values()).map((e) => F(e.linter, e, [])), c = new Map([...e.filter((e) => e.kind === "linter"), ...s].map((e) => [e.name, e]));
	return a([...e.filter((e) => e.kind !== "linter"), ...E(c).map((e) => c.get(e))]);
}
var ge = [
	"running",
	"pending",
	"failed",
	"findings",
	"passed"
];
function W(e) {
	let t = e.linter?.status === "running" ? "running" : e.linter?.status === "pending" ? "pending" : e.children?.some((e) => e.kind === "linter-error") ? "failed" : ie(e) > 0 ? "findings" : "passed", n = (e.children ?? []).filter((e) => e.kind === "linter").map(W);
	return Math.min(ge.indexOf(t), ...n);
}
function G(e) {
	return e.map((e) => ({
		node: e,
		bucket: W(e)
	})).sort((e, t) => e.bucket - t.bucket || (e.node.name < t.node.name ? -1 : +(e.node.name > t.node.name))).map(({ node: e }) => e);
}
function _e(e, t, n) {
	let r = O(t.filter((e) => o(e.linter, n.linter))), i = (e) => {
		for (let t = e.lastIndexOf("/"); t > 0; t = e.lastIndexOf("/", t - 1)) if (r.has(e.slice(0, t))) return e.slice(0, t);
	}, a = new Map(e.filter((e) => e.kind === "linter").map((e) => [e.name, e])), s = T(Array.from(r.keys()).filter((e) => i(e) !== void 0), (e) => i(e));
	if (s.size === 0) return e;
	let c = (e, t) => {
		let n = a.get(e) ?? F(e, r.get(e), []), i = (s.get(e) ?? []).map((t) => c(t, e)), o = t ? {
			name: e.slice(t.length + 1),
			nestedUnder: t
		} : {};
		if (i.length === 0) return {
			...n,
			...o
		};
		let l = n.children ?? [], u = n.failed || i.some((e) => e.failed);
		return {
			...n,
			...o,
			children: [
				...l.filter((e) => e.kind === "linter-error"),
				...G(i),
				...l.filter((e) => e.kind !== "linter-error")
			],
			failed: u,
			passed: n.passed && !u && i.every((e) => e.passed || e.skipped)
		};
	}, l = /* @__PURE__ */ new Set([...a.keys(), ...s.keys()]);
	return [...e.filter((e) => e.kind !== "linter"), ...Array.from(l).filter((e) => i(e) === void 0).map((e) => c(e))];
}
function ve(e, t) {
	let n = e.trim(), r = t.trim();
	return n && r ? `${n.replace(/\/+$/, "")}/${r.replace(/^\/+/, "")}` : n || r;
}
function K(e, t) {
	return e.map((e) => {
		if (!e.children || e.children.length === 0) return e;
		let n = K(e.children, t);
		if (n.length === 1 && e.kind === t) {
			let [t] = n;
			return {
				...t,
				name: ve(e.name, t.name)
			};
		}
		return {
			...e,
			children: n
		};
	});
}
function q(e, t = []) {
	let n = e.map((e) => b(e.name)), r = T(n, (e) => e), i = /* @__PURE__ */ new Map();
	return e.map((e, a) => {
		let o = n[a], s = (i.get(o) ?? 0) + 1;
		i.set(o, s);
		let c = r.get(o).length > 1 ? `${o}~${s}` : o, l = [...t, c], { children: u, ...d } = e;
		return {
			...d,
			route_path: l.join("/"),
			...u ? { children: q(u, l) } : {}
		};
	});
}
var ye = {
	"linter-rule-file": fe,
	"linter-file": z,
	"file-linter-rule": B,
	summary: me
};
function be({ results: e, grouping: t, filters: n }) {
	let r = he(H(ye[t](e, n), e, n), e, n), i = t === "file-linter-rule" ? r : G(_e(r, e, n));
	return q(t === "summary" ? i : K(i, "lint-folder"));
}
function J(e, t) {
	if (!t) return null;
	for (let n of e) {
		if (n.route_path === t) return n;
		let e = n.children ? J(n.children, t) : null;
		if (e) return e;
	}
	return null;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/diagnostic-identity.ts
function xe(e) {
	let t = /* @__PURE__ */ new Map();
	return e.map((e) => {
		let n = JSON.stringify(e), r = t.get(n) ?? 0;
		return t.set(n, r + 1), `${n}:${r}`;
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/ProblemFacet.tsx
var Se = {
	autoSubmit: !1,
	compact: !1
};
function Ce({ label: e, options: t, state: n, onChange: r }) {
	return /* @__PURE__ */ (0, y.jsxs)("section", {
		"aria-label": e,
		children: [/* @__PURE__ */ (0, y.jsx)("div", {
			className: "mb-1 text-xs font-medium text-muted-foreground",
			children: e
		}), /* @__PURE__ */ (0, y.jsx)(c.Provider, {
			value: Se,
			children: /* @__PURE__ */ (0, y.jsx)(l, {
				chrome: "embedded",
				filter: {
					key: e,
					kind: "multi",
					label: e,
					options: t,
					value: Object.fromEntries(n),
					onChange: (e) => r(new Map(Object.entries(e)))
				}
			})
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/ProblemViewOptions.tsx
function we({ active: e, children: t }) {
	return /* @__PURE__ */ (0, y.jsx)(a, {
		label: `View Options${e ? ` (${e})` : ""}`,
		menuLabel: "Problem view options",
		variant: "ghost",
		menuClassName: "max-h-[70vh] w-80 overflow-y-auto p-3",
		children: () => /* @__PURE__ */ (0, y.jsx)("div", {
			className: "space-y-3",
			children: t
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/types.ts
function Y(e) {
	return `${e.workDir ?? ""}\u0000${e.file}`;
}
function Te(e, t) {
	return !e || !t ? e === t : e.file === t.file && (e.workDir ?? "") === (t.workDir ?? "");
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/adapters.ts
function Ee(e) {
	return e.flatMap((e) => (e.violations ?? []).flatMap((t) => {
		let n = {
			workDir: e.work_dir,
			severity: _(t),
			source: t.source ?? e.linter,
			rule: t.rule?.method,
			kind: "lint"
		}, r = t.message ?? "";
		if ((t.locations?.length ?? 0) > 1) return t.locations.map((t) => ({
			...n,
			file: d(t.file, e.work_dir),
			line: Math.max(t.line ?? 1, 1),
			column: t.column,
			message: t.message ? `${r} (${t.message})` : r
		}));
		let i = t.raw_file ?? d(t.file, e.work_dir);
		return i ? [{
			...n,
			file: i,
			line: Math.max(t.line ?? 1, 1),
			column: t.column,
			endLine: t.end_line,
			endColumn: t.end_column,
			message: r
		}] : [];
	}));
}
function De(e) {
	return e.failure_detail?.summary || e.message || `${e.name} failed`;
}
function X(e) {
	let t = (e.children ?? []).flatMap(X);
	return t.length > 0 ? t : !e.failed || !e.file || !e.line ? [] : [{
		file: d(e.file, e.work_dir),
		workDir: e.work_dir,
		line: e.line,
		endLine: e.end_line,
		severity: "error",
		message: De(e),
		source: e.framework ?? "test",
		rule: e.name,
		kind: "test"
	}];
}
function Oe(e) {
	return e.flatMap(X);
}
function Z(e, t) {
	return {
		name: e,
		path: t,
		folders: /* @__PURE__ */ new Map(),
		files: /* @__PURE__ */ new Map()
	};
}
function ke(e, t) {
	let n = m(t.file);
	if (n.length === 0) throw Error(`findings: empty file path (${JSON.stringify(t.file)})`);
	let r = e;
	for (let e of n.slice(0, -1)) {
		let t = r.folders.get(e);
		t || (t = Z(e, r.path ? `${r.path}/${e}` : e), r.folders.set(e, t)), r = t;
	}
	let i = n[n.length - 1];
	r.files.set(i, [...r.files.get(i) ?? [], t]);
}
var Q = ([e], [t]) => e.localeCompare(t);
function $(e, t) {
	let n = Array.from(e.folders.entries()).sort(Q).map(([e, n]) => {
		let r = $(n, t);
		return {
			key: `folder\u0000${t ?? ""}\u0000${n.path}`,
			name: e,
			kind: "folder",
			children: r,
			findings: r.flatMap((e) => e.findings)
		};
	}), r = Array.from(e.files.entries()).sort(Q).map(([e, n]) => {
		let r = {
			workDir: t,
			file: n[0].file
		};
		return {
			key: Y(r),
			name: e,
			kind: "file",
			ref: r,
			findings: n.slice().sort((e, t) => e.line - t.line || (e.column ?? 0) - (t.column ?? 0))
		};
	});
	return [...n, ...r];
}
function Ae(e) {
	let t = new Map(e.map((e) => [e, e ? u(e) : "(no work dir)"])), n = /* @__PURE__ */ new Map();
	for (let e of t.values()) n.set(e, (n.get(e) ?? 0) + 1);
	return new Map(Array.from(t, ([e, t]) => [e, n.get(t) > 1 && e ? p(e) : t]));
}
function je(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.workDir, [...t.get(n.workDir) ?? [], n]);
	let n = t.size > 1, r = Ae(Array.from(t.keys()));
	return K(Array.from(t.entries()).map(([e, t]) => {
		let i = Z("", "");
		for (let e of t) ke(i, e);
		let a = $(i, e);
		return n ? [{
			key: `root\u0000${e ?? ""}`,
			name: r.get(e),
			kind: "folder",
			children: a,
			findings: t
		}] : a;
	}).flat().sort((e, t) => n ? e.name.localeCompare(t.name) : 0), "folder");
}
export { Te as a, xe as c, D as d, se as f, ae as g, S as h, Y as i, be as l, x as m, Ee as n, we as o, oe as p, Oe as r, Ce as s, je as t, J as u };
