import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-B4NxTvNR.js";
import { i as r } from "./UiInfo-CCiCR0Fr.js";
import { it as i, rt as a } from "./SplitButton-w2X8sfgS.js";
import { L as o, c as s, d as c, f as l, l as u, m as d, p as f, u as p } from "./locations-CUxJY9f0.js";
import { t as ee } from "./ProblemActions-MzUA9awO.js";
// ../../../clicky-ui/packages/ui/src/data/test-runner/routePath.ts
var m = /* @__PURE__ */ e(t(), 1), h = n();
function g(e) {
	let t = (e || "").toLowerCase().replace(/[^a-z0-9]+/g, "-");
	return t.startsWith("-") && (t = t.slice(1)), t.endsWith("-") && (t = t.slice(0, -1)), t || "node";
}
function _(e, t = []) {
	let n = /* @__PURE__ */ new Map();
	for (let t of e) {
		let e = g(t.name);
		n.set(e, (n.get(e) || 0) + 1);
	}
	let r = /* @__PURE__ */ new Map();
	return e.map((e) => {
		let i = g(e.name), a = (r.get(i) || 0) + 1;
		r.set(i, a);
		let o = (n.get(i) || 0) > 1 ? `${i}~${a}` : i, s = [...t, o], c = {
			...e,
			route_path: s.join("/")
		};
		return e.children && (c.children = _(e.children, s)), c;
	});
}
function v(e, t) {
	if (!t) return null;
	for (let n of e) {
		if (n.route_path === t) return n;
		if (n.children) {
			let e = v(n.children, t);
			if (e) return e;
		}
	}
	return null;
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/use-problem-expansion.ts
function te(e) {
	let [t, n] = (0, m.useState)([]);
	return (0, m.useEffect)(() => {
		n((t) => {
			let n = t.filter((t) => e.includes(t));
			return n.length === t.length ? t : n;
		});
	}, [e]), [t, n];
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/ProblemRow.tsx
function ne({ message: e, severity: t, metadata: n, location: o, actions: s, children: c, expanded: l, onExpandedChange: u }) {
	let [f, p] = (0, m.useState)(!1), g = l ?? f;
	return /* @__PURE__ */ (0, h.jsxs)("div", {
		className: "border-b border-border/50",
		"data-problem-row": "",
		children: [/* @__PURE__ */ (0, h.jsxs)("div", {
			className: "flex min-w-0 items-center hover:bg-muted/50",
			children: [/* @__PURE__ */ (0, h.jsxs)("button", {
				type: "button",
				"aria-expanded": g,
				title: e,
				onClick: () => l === void 0 ? p(!g) : u?.(!g),
				className: "flex min-h-7 min-w-0 flex-1 items-center gap-2 px-2 py-1 text-left text-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring [@media(pointer:coarse)]:min-h-11",
				children: [
					/* @__PURE__ */ (0, h.jsx)(r, {
						icon: g ? i : a,
						className: "shrink-0 text-muted-foreground"
					}),
					/* @__PURE__ */ (0, h.jsx)(d, { severity: t }),
					/* @__PURE__ */ (0, h.jsx)("span", {
						className: "min-w-0 flex-1 truncate",
						children: e
					}),
					/* @__PURE__ */ (0, h.jsx)("span", {
						className: "hidden max-w-[30%] truncate text-muted-foreground sm:block",
						children: n
					}),
					/* @__PURE__ */ (0, h.jsx)("span", {
						className: "max-w-[25%] shrink-0 truncate font-mono text-muted-foreground",
						"data-problem-ln": "",
						children: o
					})
				]
			}), /* @__PURE__ */ (0, h.jsx)(ee, { items: s })]
		}), g && /* @__PURE__ */ (0, h.jsxs)("div", {
			className: "space-y-2 px-3 pb-3 pl-8 text-xs",
			children: [
				/* @__PURE__ */ (0, h.jsx)("p", {
					className: "whitespace-pre-wrap break-words",
					children: e
				}),
				/* @__PURE__ */ (0, h.jsx)("div", {
					className: "text-muted-foreground sm:hidden",
					children: n
				}),
				c
			]
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/lint/stats.ts
var y = {
	info: 1,
	warning: 2,
	error: 3
};
function b(e) {
	return e.severity ?? "error";
}
function re(e) {
	let t = "info";
	for (let n of e ?? []) {
		let e = b(n);
		y[e] > y[t] && (t = e);
	}
	return t;
}
function x(e) {
	return e.status !== void 0;
}
function S(e) {
	return !e.skipped && !x(e) && (!!e.error || !!e.timed_out);
}
function ie(e) {
	return e.error || "timed out";
}
function ae(e) {
	let t = /* @__PURE__ */ new Set();
	for (let n of e) ((n.violations ?? []).length > 0 || S(n) || x(n)) && t.add(n.linter);
	return Array.from(t).sort();
}
function oe(e, t) {
	let n = {
		error: 0,
		warning: 0,
		info: 0
	};
	for (let r of e) if (o(r.linter, t)) for (let e of r.violations ?? []) n[b(e)] += 1;
	return n;
}
function se(e, t) {
	let n = {};
	for (let r of e) for (let e of r.violations ?? []) o(b(e), t) && (n[r.linter] = (n[r.linter] ?? 0) + 1);
	return n;
}
function ce(e) {
	return e.reduce((e, t) => e + (t.violations ?? []).length, 0);
}
function le(e) {
	return e.nestedUnder && e.linterName ? e.linterName.slice(e.nestedUnder.length + 1) : e.linterName;
}
function C(e) {
	return e.violations ? e.violations.length : (e.children ?? []).reduce((e, t) => e + C(t), 0);
}
// ../../../clicky-ui/packages/ui/src/data/lint/grouping.ts
var ue = "(no rule)", de = 5;
function w(e) {
	return e.rule?.method || ue;
}
function fe(e) {
	return e.slice().sort((e, t) => (e.line ?? 0) - (t.line ?? 0));
}
function pe(e) {
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
	let n = e.filter((e) => o(e.linter, t.linter)), r = new Set(n.flatMap((e) => e.work_dir ? [c(e.work_dir)] : [])).size > 1, i = [];
	for (let e of n) for (let n of e.violations ?? []) {
		if (!o(b(n), t.severity)) continue;
		let a = l(n.file, e.work_dir), s = p(a, e.work_dir, r);
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
	let i = u(t), a = u(n);
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
	let d = i[i.length - 1], f = l.leaves.get(d);
	return f || (f = r(), l.leaves.set(d, f)), f;
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
function N(e) {
	let t = re(e.violations) !== "info";
	return {
		name: e.name,
		kind: e.kind,
		file: e.file,
		target_path: e.rawFile,
		work_dir: e.workDir,
		linterName: e.linterName,
		violations: fe(e.violations),
		failed: t,
		passed: !t,
		skipped: !1
	};
}
function P(e, t, n) {
	return {
		...j(e, "linter", n),
		linter: t,
		linterName: e
	};
}
var F = {
	entity: 0,
	url: 1,
	other: 2
};
function I(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) {
		if (n.file) continue;
		let e = s(n.v), r = `${e.kind}\u0000${e.name}`, i = t.get(r);
		i ? i.flats.push(n) : t.set(r, {
			group: e,
			flats: [n]
		});
	}
	return Array.from(t.values()).sort((e, t) => F[e.group.kind] - F[t.group.kind] || e.group.name.localeCompare(t.group.name));
}
function L(e, t, n) {
	return I(e).map(({ group: e, flats: r }) => ({
		...N({
			name: e.name,
			kind: "lint-target",
			linterName: t,
			violations: r.map((e) => e.v),
			workDir: r[0].workDir
		}),
		targetKind: e.kind,
		...n ? { ruleName: n } : {}
	}));
}
function me(e) {
	let t = ie(e);
	return {
		name: pe(t),
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
function R(e, t, n) {
	let r = E(e.folders).map((r) => {
		let i = e.folders.get(r), a = j(i.name, "lint-folder", R(i, t, n));
		return {
			...a,
			file: i.path,
			target_path: i.rawPath,
			linterName: t,
			...n ? { ruleName: n } : {},
			work_dir: M(a.children)
		};
	}), i = E(e.leaves).map((r) => {
		let i = e.leaves.get(r), a = N({
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
function z(e, t) {
	let n = E(e.folders).map((n) => {
		let r = e.folders.get(n), i = j(r.name, "lint-folder", z(r, t));
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
					...N({
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
				...P(e, t.get(e), o),
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
function B(e) {
	let t = k();
	for (let n of e) n.file && A(t, n.file, n.rawFile ?? n.file, () => ({
		rawFile: n.rawFile ?? n.file,
		workDir: n.workDir,
		violations: []
	})).violations.push(n.v);
	return t;
}
function he(e, t) {
	let n = O(e), r = T(D(e, t), (e) => e.linter);
	return Array.from(r.keys()).sort().map((e) => {
		let t = T(r.get(e), (e) => w(e.v)), i = Array.from(t.keys()).sort().map((r) => {
			let i = t.get(r), a = [...R(B(i), e, r), ...L(i, e, r)];
			return {
				...j(r, "lint-rule-group", a),
				linterName: e,
				ruleName: r,
				work_dir: M(a) ?? n.get(e)?.work_dir
			};
		});
		return P(e, n.get(e), i);
	});
}
function ge(e, t) {
	let n = O(e), r = T(D(e, t), (e) => e.linter);
	return Array.from(r.keys()).sort().map((e) => {
		let t = r.get(e);
		return P(e, n.get(e), [...R(B(t), e), ...L(t, e)]);
	});
}
function _e(e, t) {
	let n = O(e), r = D(e, t), i = k();
	for (let e of r) {
		if (!e.file) continue;
		let t = A(i, e.file, e.rawFile ?? e.file, () => /* @__PURE__ */ new Map()), n = t.get(e.linter);
		n || (n = {
			workDir: e.workDir,
			rules: /* @__PURE__ */ new Map()
		}, t.set(e.linter, n));
		let r = w(e.v);
		n.rules.set(r, [...n.rules.get(r) ?? [], e.v]);
	}
	let a = I(r).map(({ group: e, flats: t }) => {
		let r = T(t, (e) => e.linter), i = E(r).map((e) => {
			let t = T(r.get(e), (e) => w(e.v)), i = n.get(e)?.work_dir, a = E(t).map((n) => ({
				...N({
					name: n,
					kind: "lint-rule",
					linterName: e,
					violations: t.get(n).map((e) => e.v),
					workDir: i
				}),
				ruleName: n
			}));
			return {
				...P(e, n.get(e), a),
				work_dir: i
			};
		});
		return {
			...j(e.name, "lint-target", i),
			targetKind: e.kind,
			work_dir: M(i)
		};
	});
	return [...z(i, n), ...a];
}
function ve(e, t) {
	if (t.length > 1) return `${e} (${t.length})`;
	let n = t[0];
	return n?.line ? n.column ? `${e}:${n.line}:${n.column}` : `${e}:${n.line}` : e;
}
function ye(e) {
	if (e.file) return `file\u0000${e.file}`;
	let t = s(e.v);
	return `${t.kind}\u0000${t.name}`;
}
function be(e, t, n) {
	let r = n[0], i = n.map((e) => e.v);
	if (!r.file) {
		let n = s(r.v);
		return {
			...N({
				name: i.length > 1 ? `${n.name} (${i.length})` : n.name,
				kind: "lint-target",
				linterName: t,
				violations: i,
				workDir: r.workDir
			}),
			targetKind: n.kind,
			ruleName: e
		};
	}
	return {
		...N({
			name: ve(r.file, i),
			kind: "lint-file",
			file: r.file,
			rawFile: r.rawFile,
			linterName: t,
			violations: i,
			workDir: r.workDir
		}),
		ruleName: e
	};
}
function xe(e, t, n) {
	let r = T(n, ye), i = Array.from(r.keys()), a = i.slice(0, de).map((n) => be(e, t, r.get(n))), o = i.length - a.length;
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
function Se(e, t) {
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
		let i = T(t, (e) => w(e.v)), o = Array.from(i.keys()).sort((e, t) => i.get(t).length - i.get(e).length || e.localeCompare(t));
		return P(e, n.get(e), o.map((t) => xe(t, e, i.get(t))));
	});
}
function Ce(e, t, n) {
	let r = T(t.filter((e) => S(e) && o(e.linter, n.linter)), (e) => e.linter);
	if (r.size === 0) return e;
	let i = new Map(e.filter((e) => e.kind === "linter").map((e) => [e.name, e]));
	for (let [e, t] of r) {
		let n = t.map(me), r = i.get(e);
		i.set(e, r ? {
			...r,
			children: [...n, ...r.children ?? []],
			failed: !0,
			passed: !1
		} : P(e, t[0], n));
	}
	return [...e.filter((e) => e.kind !== "linter"), ...E(i).map((e) => i.get(e))];
}
function V(e, t = /* @__PURE__ */ new Set()) {
	for (let n of e) n.kind === "linter" && n.linterName && t.add(n.linterName), V(n.children ?? [], t);
	return t;
}
function we(e, t, n) {
	let r = V(e), i = /* @__PURE__ */ new Map();
	for (let e of t) x(e) && o(e.linter, n.linter) && !r.has(e.linter) && i.set(e.linter, e);
	let a = (e) => e.map((e) => {
		let t = e.children ? a(e.children) : void 0, n = e.kind === "linter" && e.linter !== void 0 && x(e.linter);
		return {
			...e,
			...t ? { children: t } : {},
			...n ? { passed: !1 } : {}
		};
	});
	if (i.size === 0) return a(e);
	let s = Array.from(i.values()).map((e) => P(e.linter, e, [])), c = new Map([...e.filter((e) => e.kind === "linter"), ...s].map((e) => [e.name, e]));
	return a([...e.filter((e) => e.kind !== "linter"), ...E(c).map((e) => c.get(e))]);
}
var Te = [
	"running",
	"pending",
	"failed",
	"findings",
	"passed"
];
function H(e) {
	let t = e.linter?.status === "running" ? "running" : e.linter?.status === "pending" ? "pending" : e.children?.some((e) => e.kind === "linter-error") ? "failed" : C(e) > 0 ? "findings" : "passed", n = (e.children ?? []).filter((e) => e.kind === "linter").map(H);
	return Math.min(Te.indexOf(t), ...n);
}
function U(e) {
	return e.map((e) => ({
		node: e,
		bucket: H(e)
	})).sort((e, t) => e.bucket - t.bucket || (e.node.name < t.node.name ? -1 : +(e.node.name > t.node.name))).map(({ node: e }) => e);
}
function Ee(e, t, n) {
	let r = O(t.filter((e) => o(e.linter, n.linter))), i = (e) => {
		for (let t = e.lastIndexOf("/"); t > 0; t = e.lastIndexOf("/", t - 1)) if (r.has(e.slice(0, t))) return e.slice(0, t);
	}, a = new Map(e.filter((e) => e.kind === "linter").map((e) => [e.name, e])), s = T(Array.from(r.keys()).filter((e) => i(e) !== void 0), (e) => i(e));
	if (s.size === 0) return e;
	let c = (e, t) => {
		let n = a.get(e) ?? P(e, r.get(e), []), i = (s.get(e) ?? []).map((t) => c(t, e)), o = t ? {
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
				...U(i),
				...l.filter((e) => e.kind !== "linter-error")
			],
			failed: u,
			passed: n.passed && !u && i.every((e) => e.passed || e.skipped)
		};
	}, l = /* @__PURE__ */ new Set([...a.keys(), ...s.keys()]);
	return [...e.filter((e) => e.kind !== "linter"), ...Array.from(l).filter((e) => i(e) === void 0).map((e) => c(e))];
}
function De(e, t) {
	let n = e.trim(), r = t.trim();
	return n && r ? `${n.replace(/\/+$/, "")}/${r.replace(/^\/+/, "")}` : n || r;
}
function W(e, t) {
	return e.map((e) => {
		if (!e.children || e.children.length === 0) return e;
		let n = W(e.children, t);
		if (n.length === 1 && e.kind === t) {
			let [t] = n;
			return {
				...t,
				name: De(e.name, t.name)
			};
		}
		return {
			...e,
			children: n
		};
	});
}
function G(e, t = []) {
	let n = e.map((e) => g(e.name)), r = T(n, (e) => e), i = /* @__PURE__ */ new Map();
	return e.map((e, a) => {
		let o = n[a], s = (i.get(o) ?? 0) + 1;
		i.set(o, s);
		let c = r.get(o).length > 1 ? `${o}~${s}` : o, l = [...t, c], { children: u, ...d } = e;
		return {
			...d,
			route_path: l.join("/"),
			...u ? { children: G(u, l) } : {}
		};
	});
}
var Oe = {
	"linter-rule-file": he,
	"linter-file": ge,
	"file-linter-rule": _e,
	summary: Se
};
function ke({ results: e, grouping: t, filters: n }) {
	let r = we(Ce(Oe[t](e, n), e, n), e, n), i = t === "file-linter-rule" ? r : U(Ee(r, e, n));
	return G(t === "summary" ? i : W(i, "lint-folder"));
}
function K(e, t) {
	if (!t) return null;
	for (let n of e) {
		if (n.route_path === t) return n;
		let e = n.children ? K(n.children, t) : null;
		if (e) return e;
	}
	return null;
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/diagnostic-identity.ts
function Ae(e) {
	let t = /* @__PURE__ */ new Map();
	return e.map((e) => {
		let n = JSON.stringify(e), r = t.get(n) ?? 0;
		return t.set(n, r + 1), `${n}:${r}`;
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/types.ts
function je(e, t) {
	return !e || !t ? e === t : e.kind === t.kind && Me(e.ref, t.ref);
}
function q(e) {
	return `${e.workDir ?? ""}\u0000${e.file}`;
}
function Me(e, t) {
	return !e || !t ? e === t : e.file === t.file && (e.workDir ?? "") === (t.workDir ?? "");
}
// ../../../clicky-ui/packages/ui/src/file-findings/adapters.ts
function Ne(e) {
	return e.flatMap((e) => (e.violations ?? []).flatMap((t) => {
		let n = {
			workDir: e.work_dir,
			severity: b(t),
			source: t.source ?? e.linter,
			rule: t.rule?.method,
			kind: "lint"
		}, r = t.message ?? "", i = (t.locations ?? []).filter((e) => e.file);
		if (i.length > 1) return i.map((t) => ({
			...n,
			file: l(t.file, e.work_dir),
			line: Math.max(t.line ?? 1, 1),
			column: t.column,
			message: t.message ? `${r} (${t.message})` : r
		}));
		let a = t.raw_file ?? l(t.file, e.work_dir);
		return a ? [{
			...n,
			file: a,
			line: Math.max(t.line ?? 1, 1),
			column: t.column,
			endLine: t.end_line,
			endColumn: t.end_column,
			message: r,
			path: t.path
		}] : [];
	}));
}
function J(e) {
	return e.failure_detail?.summary || e.message || `${e.name} failed`;
}
function Y(e) {
	let t = (e.children ?? []).flatMap(Y);
	return t.length > 0 ? t : !e.failed || !e.file || !e.line ? [] : [{
		file: l(e.file, e.work_dir),
		workDir: e.work_dir,
		line: e.line,
		endLine: e.end_line,
		severity: "error",
		message: J(e),
		source: e.framework ?? "test",
		rule: e.name,
		kind: "test"
	}];
}
function Pe(e) {
	return e.flatMap(Y);
}
function X(e, t) {
	return {
		name: e,
		path: t,
		folders: /* @__PURE__ */ new Map(),
		files: /* @__PURE__ */ new Map()
	};
}
function Z(e, t, n) {
	let r = u(t);
	if (r.length === 0) throw Error(`findings: empty file path (${JSON.stringify(t)})`);
	let i = e;
	for (let e of r.slice(0, -1)) {
		let t = i.folders.get(e);
		t || (t = X(e, i.path ? `${i.path}/${e}` : e), i.folders.set(e, t)), i = t;
	}
	let a = r[r.length - 1], o = i.files.get(a) ?? {
		file: t,
		findings: []
	};
	n && o.findings.push(n), i.files.set(a, o);
}
var Q = ([e], [t]) => e.localeCompare(t);
function $(e, t) {
	let n = Array.from(e.folders.entries()).sort(Q).map(([e, n]) => {
		let r = $(n, t);
		return {
			key: `folder\u0000${t ?? ""}\u0000${n.path}`,
			name: e,
			kind: "folder",
			ref: {
				workDir: t,
				file: n.path
			},
			children: r,
			findings: r.flatMap((e) => e.findings)
		};
	}), r = Array.from(e.files.entries()).sort(Q).map(([e, { file: n, findings: r }]) => {
		let i = {
			workDir: t,
			file: n
		};
		return {
			key: q(i),
			name: e,
			kind: "file",
			ref: i,
			findings: r.slice().sort((e, t) => e.line - t.line || (e.column ?? 0) - (t.column ?? 0))
		};
	});
	return [...n, ...r];
}
function Fe(e) {
	let t = new Map(e.map((e) => [e, e ? f(e) : "(no work dir)"])), n = /* @__PURE__ */ new Map();
	for (let e of t.values()) n.set(e, (n.get(e) ?? 0) + 1);
	return new Map(Array.from(t, ([e, t]) => [e, n.get(t) > 1 && e ? c(e) : t]));
}
function Ie(e, t = [], n = []) {
	let r = /* @__PURE__ */ new Map(), i = (e) => {
		let t = r.get(e);
		if (t) return t;
		let n = {
			findings: [],
			files: [],
			folders: []
		};
		return r.set(e, n), n;
	};
	for (let t of e) i(t.workDir).findings.push(t);
	for (let e of t) i(e.workDir).files.push(e.file);
	for (let e of n) i(e.workDir).folders.push(e.file);
	let a = r.size > 1, o = Fe(Array.from(r.keys()));
	return W(Array.from(r.entries()).map(([e, { findings: t, files: n, folders: r }]) => {
		let i = X("", "");
		for (let e of r) {
			let t = i;
			for (let n of u(e)) {
				let e = t.folders.get(n) ?? X(n, t.path ? `${t.path}/${n}` : n);
				t.folders.set(n, e), t = e;
			}
		}
		for (let e of n) Z(i, e);
		for (let e of t) Z(i, e.file, e);
		let s = $(i, e);
		return a ? [{
			key: `root\u0000${e ?? ""}`,
			name: o.get(e),
			kind: "folder",
			ref: {
				workDir: e,
				file: ""
			},
			children: s,
			findings: t
		}] : s;
	}).flat().sort((e, t) => a ? e.name.localeCompare(t.name) : 0), "folder");
}
export { ce as _, je as a, _ as b, K as c, se as d, oe as f, b as g, le as h, q as i, D as l, C as m, Ne as n, Ae as o, S as p, Pe as r, ke as s, Ie as t, ae as u, ne as v, v as x, te as y };
