import "./rolldown-runtime-QGRDbDUS.js";
import { n as e, t } from "./jsx-runtime-lVm_Myqd.js";
import { s as n } from "./clipboard-C0PFseuO.js";
import { a as r, i, n as a, r as o, t as s } from "./paths-RfM6xXIw.js";
import { a as c, d as l, l as u, m as d, o as f, s as p } from "./Tabs-Cbv7DUex.js";
e();
var m = t(), h = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#dc2626",
		...r
	};
	return /* @__PURE__ */ (0, m.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, m.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/data/test-runner/routePath.ts
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
// ../../../clicky-ui/packages/ui/src/data/lint/grouping.ts
var ee = "(no rule)", te = 5;
function y(e) {
	return e.rule?.method || ee;
}
function b(e) {
	return e.slice().sort((e, t) => (e.line ?? 0) - (t.line ?? 0));
}
function x(e) {
	let t = e.indexOf("\n");
	return t === -1 ? e : e.slice(0, t);
}
function S(e, t) {
	let n = /* @__PURE__ */ new Map();
	for (let r of e) {
		let e = n.get(t(r));
		e ? e.push(r) : n.set(t(r), [r]);
	}
	return n;
}
function C(e) {
	return Array.from(e.keys()).sort();
}
function w(e, t) {
	let r = e.filter((e) => n(e.linter, t.linter)), s = new Set(r.flatMap((e) => e.work_dir ? [o(e.work_dir)] : [])).size > 1, c = [];
	for (let e of r) for (let r of e.violations ?? []) {
		if (!n(l(r), t.severity)) continue;
		let o = i(r.file, e.work_dir), u = a(o, e.work_dir, s);
		c.push({
			linter: e.linter,
			result: e,
			v: {
				...r,
				raw_file: o,
				file: u
			},
			file: u,
			rawFile: o,
			workDir: e.work_dir
		});
	}
	return c;
}
function T(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.has(n.linter) || t.set(n.linter, n);
	return t;
}
function E() {
	return {
		name: "",
		path: void 0,
		rawPath: "",
		folders: /* @__PURE__ */ new Map(),
		leaves: /* @__PURE__ */ new Map()
	};
}
function D(e, t, n, r) {
	let i = s(t), a = s(n);
	if (i.length === 0 || a.length === 0) throw Error(`lint: cannot place violation, empty path (file=${JSON.stringify(t)}, raw=${JSON.stringify(n)})`);
	let o = i.slice(0, -1), c = a.slice(0, -1), l = o.length - c.length, u = e;
	o.forEach((e, t) => {
		let n = u.folders.get(e);
		if (!n) {
			let r = Math.max(0, Math.min(c.length, t + 1 - l));
			n = {
				name: e,
				path: u.path ? `${u.path}/${e}` : e,
				rawPath: c.slice(0, r).join("/"),
				folders: /* @__PURE__ */ new Map(),
				leaves: /* @__PURE__ */ new Map()
			}, u.folders.set(e, n);
		}
		u = n;
	});
	let d = i[i.length - 1], f = u.leaves.get(d);
	return f || (f = r(), u.leaves.set(d, f)), f;
}
function O(e, t, n) {
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
function k(e) {
	let t = new Set(e.flatMap((e) => e.work_dir ? [e.work_dir] : []));
	return t.size === 1 ? Array.from(t)[0] : void 0;
}
function A(e, t) {
	if (t.length === 0) return e;
	let n = e.failed || d(t) !== "info";
	return {
		...e,
		noFileViolations: b(t),
		failed: n,
		passed: !n
	};
}
function j(e) {
	let t = d(e.violations) !== "info";
	return {
		name: e.name,
		kind: e.kind,
		file: e.file,
		target_path: e.rawFile,
		work_dir: e.workDir,
		linterName: e.linterName,
		violations: b(e.violations),
		failed: t,
		passed: !t,
		skipped: !1
	};
}
function M(e, t, n, r = []) {
	return A({
		...O(e, "linter", n),
		linter: t,
		linterName: e
	}, r);
}
function N(e) {
	let t = u(e);
	return {
		name: x(t),
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
function P(e, t, n) {
	let r = C(e.folders).map((r) => {
		let i = e.folders.get(r), a = O(i.name, "lint-folder", P(i, t, n));
		return {
			...a,
			file: i.path,
			target_path: i.rawPath,
			linterName: t,
			...n ? { ruleName: n } : {},
			work_dir: k(a.children)
		};
	}), i = C(e.leaves).map((r) => {
		let i = e.leaves.get(r), a = j({
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
function F(e, t) {
	let n = C(e.folders).map((n) => {
		let r = e.folders.get(n), i = O(r.name, "lint-folder", F(r, t));
		return {
			...i,
			file: r.path,
			target_path: r.rawPath,
			work_dir: k(i.children)
		};
	}), r = C(e.leaves).map((n) => {
		let r = e.leaves.get(n), i = e.path ? `${e.path}/${n}` : n, a = C(r).map((e) => {
			let { workDir: n, rules: a } = r.get(e), o = C(a).map((t) => {
				let r = a.get(t);
				return {
					...j({
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
				...M(e, t.get(e), o),
				file: i,
				target_path: o[0]?.target_path,
				work_dir: n
			};
		});
		return {
			...O(n, "lint-file", a),
			file: i,
			target_path: a[0]?.target_path,
			work_dir: k(a)
		};
	});
	return [...n, ...r];
}
function I(e) {
	let t = E();
	for (let n of e) n.file && D(t, n.file, n.rawFile ?? n.file, () => ({
		rawFile: n.rawFile ?? n.file,
		workDir: n.workDir,
		violations: []
	})).violations.push(n.v);
	return t;
}
function ne(e, t) {
	let n = T(e), r = S(w(e, t), (e) => e.linter);
	return Array.from(r.keys()).sort().map((e) => {
		let t = S(r.get(e), (e) => y(e.v)), i = Array.from(t.keys()).sort().map((r) => {
			let i = t.get(r), a = P(I(i), e, r);
			return A({
				...O(r, "lint-rule-group", a),
				linterName: e,
				ruleName: r,
				work_dir: k(a) ?? n.get(e)?.work_dir
			}, i.filter((e) => !e.file).map((e) => e.v));
		});
		return M(e, n.get(e), i);
	});
}
function re(e, t) {
	let n = T(e), r = S(w(e, t), (e) => e.linter);
	return Array.from(r.keys()).sort().map((e) => {
		let t = r.get(e);
		return M(e, n.get(e), P(I(t), e), t.filter((e) => !e.file).map((e) => e.v));
	});
}
function ie(e, t) {
	let n = T(e), r = w(e, t), i = E();
	for (let e of r) {
		if (!e.file) continue;
		let t = D(i, e.file, e.rawFile ?? e.file, () => /* @__PURE__ */ new Map()), n = t.get(e.linter);
		n || (n = {
			workDir: e.workDir,
			rules: /* @__PURE__ */ new Map()
		}, t.set(e.linter, n));
		let r = y(e.v);
		n.rules.set(r, [...n.rules.get(r) ?? [], e.v]);
	}
	let a = S(r.filter((e) => !e.file), (e) => e.linter), o = Array.from(a.keys()).sort().map((e) => {
		let t = S(a.get(e), (e) => y(e.v)), r = n.get(e)?.work_dir, i = Array.from(t.keys()).sort().map((n) => ({
			...j({
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
		return M(e, n.get(e), i);
	});
	return [...F(i, n), ...o];
}
function ae(e, t) {
	if (!e) return "(no file)";
	if (t.length > 1) return `${e} (${t.length})`;
	let n = t[0];
	return n?.line ? n.column ? `${e}:${n.line}:${n.column}` : `${e}:${n.line}` : e;
}
function oe(e, t, n) {
	let r = S(n, (e) => e.file ?? ""), i = Array.from(r.keys()), a = i.slice(0, te).map((n) => {
		let i = r.get(n);
		return {
			...j({
				name: ae(i[0].file, i.map((e) => e.v)),
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
		...O(s ? `${e} (${n.length}) — ${s}` : `${e} (${n.length})`, "lint-rule-group", a),
		linterName: t,
		ruleName: e
	};
}
function L(e, t) {
	let r = T(e), i = S(w(e, t), (e) => e.linter), a = e.filter((e) => n(e.linter, t.linter)), o = new Map(a.filter((e) => e.skipped).map((e) => [e.linter, e.error || "skipped"]));
	return Array.from(new Set(a.map((e) => e.linter))).sort().map((e) => {
		let t = i.get(e) ?? [];
		if (t.length === 0 && o.has(e)) return {
			name: e,
			kind: "linter",
			linter: r.get(e),
			linterName: e,
			skipped: !0,
			passed: !1,
			failed: !1,
			message: `skipped: ${o.get(e)}`
		};
		let n = S(t, (e) => y(e.v)), a = Array.from(n.keys()).sort((e, t) => n.get(t).length - n.get(e).length || e.localeCompare(t));
		return M(e, r.get(e), a.map((t) => oe(t, e, n.get(t))));
	});
}
function R(e, t, r) {
	let i = S(t.filter((e) => c(e) && n(e.linter, r.linter)), (e) => e.linter);
	if (i.size === 0) return e;
	let a = new Map(e.filter((e) => e.kind === "linter").map((e) => [e.name, e]));
	for (let [e, t] of i) {
		let n = t.map(N), r = a.get(e);
		a.set(e, r ? {
			...r,
			children: [...n, ...r.children ?? []],
			failed: !0,
			passed: !1
		} : M(e, t[0], n));
	}
	return [...e.filter((e) => e.kind !== "linter"), ...C(a).map((e) => a.get(e))];
}
function z(e, t = /* @__PURE__ */ new Set()) {
	for (let n of e) n.kind === "linter" && n.linterName && t.add(n.linterName), z(n.children ?? [], t);
	return t;
}
function B(e, t, r) {
	let i = z(e), a = /* @__PURE__ */ new Map();
	for (let e of t) f(e) && n(e.linter, r.linter) && !i.has(e.linter) && a.set(e.linter, e);
	let o = (e) => e.map((e) => {
		let t = e.children ? o(e.children) : void 0, n = e.kind === "linter" && e.linter !== void 0 && f(e.linter);
		return {
			...e,
			...t ? { children: t } : {},
			...n ? { passed: !1 } : {}
		};
	});
	if (a.size === 0) return o(e);
	let s = Array.from(a.values()).map((e) => M(e.linter, e, [])), c = new Map([...e.filter((e) => e.kind === "linter"), ...s].map((e) => [e.name, e]));
	return o([...e.filter((e) => e.kind !== "linter"), ...C(c).map((e) => c.get(e))]);
}
var V = [
	"running",
	"pending",
	"failed",
	"findings",
	"passed"
];
function H(e) {
	let t = e.linter?.status === "running" ? "running" : e.linter?.status === "pending" ? "pending" : e.children?.some((e) => e.kind === "linter-error") ? "failed" : p(e) > 0 ? "findings" : "passed", n = (e.children ?? []).filter((e) => e.kind === "linter").map(H);
	return Math.min(V.indexOf(t), ...n);
}
function U(e) {
	return e.map((e) => ({
		node: e,
		bucket: H(e)
	})).sort((e, t) => e.bucket - t.bucket || (e.node.name < t.node.name ? -1 : +(e.node.name > t.node.name))).map(({ node: e }) => e);
}
function W(e, t, r) {
	let i = T(t.filter((e) => n(e.linter, r.linter))), a = (e) => {
		for (let t = e.lastIndexOf("/"); t > 0; t = e.lastIndexOf("/", t - 1)) if (i.has(e.slice(0, t))) return e.slice(0, t);
	}, o = new Map(e.filter((e) => e.kind === "linter").map((e) => [e.name, e])), s = S(Array.from(i.keys()).filter((e) => a(e) !== void 0), (e) => a(e));
	if (s.size === 0) return e;
	let c = (e, t) => {
		let n = o.get(e) ?? M(e, i.get(e), []), r = (s.get(e) ?? []).map((t) => c(t, e)), a = t ? {
			name: e.slice(t.length + 1),
			nestedUnder: t
		} : {};
		if (r.length === 0) return {
			...n,
			...a
		};
		let l = n.children ?? [], u = n.failed || r.some((e) => e.failed);
		return {
			...n,
			...a,
			children: [
				...l.filter((e) => e.kind === "linter-error"),
				...U(r),
				...l.filter((e) => e.kind !== "linter-error")
			],
			failed: u,
			passed: n.passed && !u && r.every((e) => e.passed || e.skipped)
		};
	}, l = /* @__PURE__ */ new Set([...o.keys(), ...s.keys()]);
	return [...e.filter((e) => e.kind !== "linter"), ...Array.from(l).filter((e) => a(e) === void 0).map((e) => c(e))];
}
function G(e, t) {
	let n = e.trim(), r = t.trim();
	return n && r ? `${n} > ${r}` : n || r;
}
function K(e, t) {
	return e.map((e) => {
		if (!e.children || e.children.length === 0) return e;
		let n = K(e.children, t);
		if (n.length === 1 && e.kind === t) {
			let [t] = n;
			return {
				...t,
				name: G(e.name, t.name)
			};
		}
		return {
			...e,
			children: n
		};
	});
}
function q(e, t = []) {
	let n = e.map((e) => g(e.name)), r = S(n, (e) => e), i = /* @__PURE__ */ new Map();
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
var se = {
	"linter-rule-file": ne,
	"linter-file": re,
	"file-linter-rule": ie,
	summary: L
};
function ce({ results: e, grouping: t, filters: n }) {
	let r = B(R(se[t](e, n), e, n), e, n), i = t === "file-linter-rule" ? r : U(W(r, e, n));
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
// ../../../clicky-ui/packages/ui/src/file-findings/types.ts
function Y(e) {
	return `${e.workDir ?? ""}\u0000${e.file}`;
}
function le(e, t) {
	return !e || !t ? e === t : e.file === t.file && (e.workDir ?? "") === (t.workDir ?? "");
}
// ../../../clicky-ui/packages/ui/src/file-findings/adapters.ts
function ue(e) {
	return e.flatMap((e) => (e.violations ?? []).flatMap((t) => {
		let n = {
			workDir: e.work_dir,
			severity: l(t),
			source: t.source ?? e.linter,
			rule: t.rule?.method,
			kind: "lint"
		}, r = t.message ?? "";
		if ((t.locations?.length ?? 0) > 1) return t.locations.map((t) => ({
			...n,
			file: i(t.file, e.work_dir),
			line: Math.max(t.line ?? 1, 1),
			column: t.column,
			message: t.message ? `${r} (${t.message})` : r
		}));
		let a = t.raw_file ?? i(t.file, e.work_dir);
		return a ? [{
			...n,
			file: a,
			line: Math.max(t.line ?? 1, 1),
			column: t.column,
			endLine: t.end_line,
			endColumn: t.end_column,
			message: r
		}] : [];
	}));
}
function de(e) {
	return e.failure_detail?.summary || e.message || `${e.name} failed`;
}
function X(e) {
	let t = (e.children ?? []).flatMap(X);
	return t.length > 0 ? t : !e.failed || !e.file || !e.line ? [] : [{
		file: i(e.file, e.work_dir),
		workDir: e.work_dir,
		line: e.line,
		endLine: e.end_line,
		severity: "error",
		message: de(e),
		source: e.framework ?? "test",
		rule: e.name,
		kind: "test"
	}];
}
function fe(e) {
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
function pe(e, t) {
	let n = s(t.file);
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
function me(e) {
	let t = new Map(e.map((e) => [e, e ? r(e) : "(no work dir)"])), n = /* @__PURE__ */ new Map();
	for (let e of t.values()) n.set(e, (n.get(e) ?? 0) + 1);
	return new Map(Array.from(t, ([e, t]) => [e, n.get(t) > 1 && e ? o(e) : t]));
}
function he(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.workDir, [...t.get(n.workDir) ?? [], n]);
	let n = t.size > 1, r = me(Array.from(t.keys()));
	return K(Array.from(t.entries()).map(([e, t]) => {
		let i = Z("", "");
		for (let e of t) pe(i, e);
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
export { le as a, _ as c, Y as i, v as l, ue as n, ce as o, fe as r, J as s, he as t, h as u };
