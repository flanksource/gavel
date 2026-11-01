import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./jsx-runtime-f3E3oPuc.js";
import { T as r, _ as i } from "./Callout-DALwzckz.js";
import { A as a, C as o, D as s, Et as c, I as l, L as u, M as d, O as f, P as p, S as m, T as h, _t as g, f as _, j as v, k as y, l as b, o as x, s as S, u as C, w, x as T } from "./clipboard-CMVj3cUp.js";
import { a as E, l as D, n as O, o as k } from "./floating-ui.react-dom-Dh-Ge66-.js";
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/overlay/ContextMenu.tsx
var A = /* @__PURE__ */ e(t(), 1), j = n();
function M(e, t, n) {
	return {
		contextElement: n,
		getBoundingClientRect: () => ({
			x: e,
			y: t,
			top: t,
			left: e,
			right: e,
			bottom: t,
			width: 0,
			height: 0,
			toJSON: () => ({})
		})
	};
}
function N({ contextTarget: e, menuLabel: t, menuItems: n }) {
	let [r, i] = (0, A.useState)(!1), [c, h] = (0, A.useState)(null), g = (0, A.useRef)(null), b = (0, A.useRef)([]), x = u(), S = a(), N = y(), P = (0, A.useCallback)(() => {
		i(!1), queueMicrotask(() => g.current?.focus({ preventScroll: !0 }));
	}, []), { refs: F, floatingStyles: I, context: L } = f({
		...N ? { nodeId: N } : {},
		open: r,
		onOpenChange: (e) => e ? i(!0) : P(),
		placement: "right-start",
		middleware: [
			E(4),
			O({ padding: 8 }),
			k({ padding: 8 })
		],
		whileElementsMounted: D
	}), R = s(L, { escapeKey: !1 }), z = p(L, { role: "menu" }), B = d(L, {
		listRef: b,
		activeIndex: c,
		onNavigate: h,
		loop: !0
	}), { getFloatingProps: V, getItemProps: H } = v([
		R,
		z,
		B
	]), U = (0, A.useCallback)((t, r) => {
		if (!e || n.length === 0) return;
		let a = document.activeElement;
		g.current = a instanceof HTMLElement ? a : e, F.setPositionReference(M(t, r, e)), h(null), i(!0);
	}, [
		e,
		n.length,
		F
	]);
	return (0, A.useEffect)(() => {
		if (!e || n.length === 0) return;
		let t = (e) => {
			e.preventDefault(), e.stopPropagation(), U(e.clientX, e.clientY);
		}, r = (t) => {
			if (t.key !== "ContextMenu" && !(t.key === "F10" && t.shiftKey)) return;
			t.preventDefault(), t.stopPropagation();
			let n = e.getBoundingClientRect();
			U(n.left + 8, n.top + Math.min(n.height, 24));
		};
		return e.addEventListener("contextmenu", t), e.addEventListener("keydown", r), () => {
			e.removeEventListener("contextmenu", t), e.removeEventListener("keydown", r);
		};
	}, [
		e,
		n.length,
		U
	]), (0, A.useEffect)(() => {
		if (S) return S.events.on("click", P), () => S.events.off("click", P);
	}, [P, S]), l(r, P), /* @__PURE__ */ (0, j.jsx)(o, {
		id: N,
		children: r && /* @__PURE__ */ (0, j.jsx)(w, { children: /* @__PURE__ */ (0, j.jsx)(T, {
			context: L,
			modal: !1,
			returnFocus: !1,
			children: /* @__PURE__ */ (0, j.jsx)("div", {
				ref: F.setFloating,
				role: "menu",
				"aria-label": t,
				style: {
					...I,
					zIndex: x
				},
				className: "min-w-[8rem] overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-md",
				...V(),
				children: /* @__PURE__ */ (0, j.jsx)(_.Provider, {
					value: {
						getItemProps: H,
						activeIndex: c,
						setActiveIndex: h,
						setHasFocusInside: () => {},
						isOpen: r
					},
					children: /* @__PURE__ */ (0, j.jsx)(m, {
						elementsRef: b,
						children: /* @__PURE__ */ (0, j.jsx)(C, { items: n })
					})
				})
			})
		}) })
	});
}
function P(e) {
	return /* @__PURE__ */ (0, j.jsx)(h, { children: /* @__PURE__ */ (0, j.jsx)(N, { ...e }) });
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/ProblemActions.tsx
function F({ items: e, label: t = "Problem actions" }) {
	let [n, r] = (0, A.useState)(null), a = (0, A.useCallback)((e) => r(e?.closest("tr") ?? e?.parentElement ?? null), []);
	return e.length ? /* @__PURE__ */ (0, j.jsxs)("span", {
		ref: a,
		className: "shrink-0",
		onClick: (e) => e.stopPropagation(),
		onKeyDown: (e) => e.stopPropagation(),
		children: [/* @__PURE__ */ (0, j.jsx)(b, {
			menuLabel: t,
			items: e,
			trigger: /* @__PURE__ */ (0, j.jsx)(i, {
				variant: "ghost",
				size: "icon",
				className: "size-7 text-muted-foreground",
				"aria-label": t,
				children: /* @__PURE__ */ (0, j.jsx)(c, { icon: g })
			})
		}), /* @__PURE__ */ (0, j.jsx)(P, {
			contextTarget: n,
			menuLabel: t,
			menuItems: e
		})]
	}) : null;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/stats.ts
var I = {
	info: 1,
	warning: 2,
	error: 3
};
function L(e) {
	return I[e];
}
function R(e) {
	return e.severity ?? "error";
}
function z(e) {
	let t = "info";
	for (let n of e ?? []) {
		let e = R(n);
		I[e] > I[t] && (t = e);
	}
	return t;
}
function B(e) {
	return e.status !== void 0;
}
function V(e) {
	return !e.skipped && !B(e) && (!!e.error || !!e.timed_out);
}
function H(e) {
	return e.error || "timed out";
}
function U(e) {
	let t = /* @__PURE__ */ new Set();
	for (let n of e) ((n.violations ?? []).length > 0 || V(n) || B(n)) && t.add(n.linter);
	return Array.from(t).sort();
}
function W(e, t) {
	let n = {
		error: 0,
		warning: 0,
		info: 0
	};
	for (let r of e) if (x(r.linter, t)) for (let e of r.violations ?? []) n[R(e)] += 1;
	return n;
}
function G(e, t) {
	let n = {};
	for (let r of e) for (let e of r.violations ?? []) x(R(e), t) && (n[r.linter] = (n[r.linter] ?? 0) + 1);
	return n;
}
function K(e) {
	return e.reduce((e, t) => e + (t.violations ?? []).length, 0);
}
function q(e) {
	return e.nestedUnder && e.linterName ? e.linterName.slice(e.nestedUnder.length + 1) : e.linterName;
}
function J(e) {
	return e.violations ? e.violations.length : (e.noFileViolations?.length ?? 0) + (e.children ?? []).reduce((e, t) => e + J(t), 0);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/layout/Tabs.tsx
function Y({ tabs: e, value: t, onChange: n, variant: i = "underline", className: a }) {
	return /* @__PURE__ */ (0, j.jsx)("div", {
		role: "tablist",
		className: r("flex items-center gap-density-1", i === "underline" && "border-b border-border", a),
		children: e.map((e) => /* @__PURE__ */ (0, j.jsx)(S, {
			active: e.id === t,
			onClick: () => {
				e.disabled || n(e.id);
			},
			label: e.label,
			variant: i,
			...e.icon === void 0 ? {} : { icon: e.icon },
			...e.count === void 0 ? {} : { count: e.count },
			...e.countColor === void 0 ? {} : { countColor: e.countColor },
			...e.disabled ? { className: "opacity-50 pointer-events-none" } : {}
		}, e.id))
	});
}
export { V as a, H as c, L as d, K as f, P as h, W as i, q as l, F as m, U as n, B as o, z as p, G as r, J as s, Y as t, R as u };
