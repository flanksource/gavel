import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-B4NxTvNR.js";
import { t as r } from "./button-BLoR2pgR.js";
import { i } from "./UiInfo-CCiCR0Fr.js";
import { C as a, D as o, E as s, M as c, N as l, S as u, T as d, _ as f, b as p, f as m, g as h, h as g, k as _, p as v, v as y, w as b, y as x } from "./SplitButton-w2X8sfgS.js";
import { m as S } from "./ErrorDetails-CbwPqfQb.js";
import { a as C, l as w, n as T, o as E } from "./floating-ui.react-dom-CJoIjj1G.js";
// ../../../clicky-ui/packages/ui/src/overlay/ContextMenu.tsx
var D = /* @__PURE__ */ e(t(), 1), O = n();
function k(e, t, n) {
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
function A({ contextTarget: e, menuLabel: t, menuItems: n }) {
	let [r, i] = (0, D.useState)(!1), [p, m] = (0, D.useState)(null), S = (0, D.useRef)(null), A = (0, D.useRef)([]), j = l(), M = d(), N = b(), P = (0, D.useCallback)(() => {
		i(!1), queueMicrotask(() => S.current?.focus({ preventScroll: !0 }));
	}, []), { refs: F, floatingStyles: I, context: L } = a({
		...N ? { nodeId: N } : {},
		open: r,
		onOpenChange: (e) => e ? i(!0) : P(),
		placement: "right-start",
		middleware: [
			C(4),
			T({ padding: 8 }),
			E({ padding: 8 })
		],
		whileElementsMounted: w
	}), R = u(L, { escapeKey: !1 }), z = _(L, { role: "menu" }), B = o(L, {
		listRef: A,
		activeIndex: p,
		onNavigate: m,
		loop: !0
	}), { getFloatingProps: V, getItemProps: H } = s([
		R,
		z,
		B
	]), U = (0, D.useCallback)((t, r) => {
		if (!e || n.length === 0) return;
		let a = document.activeElement;
		S.current = a instanceof HTMLElement ? a : e, F.setPositionReference(k(t, r, e)), m(null), i(!0);
	}, [
		e,
		n.length,
		F
	]);
	return (0, D.useEffect)(() => {
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
	]), (0, D.useEffect)(() => {
		if (M) return M.events.on("click", P), () => M.events.off("click", P);
	}, [P, M]), c(r, P), /* @__PURE__ */ (0, O.jsx)(y, {
		id: N,
		children: r && /* @__PURE__ */ (0, O.jsx)(x, { children: /* @__PURE__ */ (0, O.jsx)(h, {
			context: L,
			modal: !1,
			returnFocus: !1,
			children: /* @__PURE__ */ (0, O.jsx)("div", {
				ref: F.setFloating,
				role: "menu",
				"aria-label": t,
				style: {
					...I,
					zIndex: j
				},
				className: "min-w-[8rem] overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-md",
				...V(),
				children: /* @__PURE__ */ (0, O.jsx)(g.Provider, {
					value: {
						getItemProps: H,
						activeIndex: p,
						setActiveIndex: m,
						setHasFocusInside: () => {},
						isOpen: r
					},
					children: /* @__PURE__ */ (0, O.jsx)(f, {
						elementsRef: A,
						children: /* @__PURE__ */ (0, O.jsx)(v, { items: n })
					})
				})
			})
		}) })
	});
}
function j(e) {
	return /* @__PURE__ */ (0, O.jsx)(p, { children: /* @__PURE__ */ (0, O.jsx)(A, { ...e }) });
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/ProblemActions.tsx
function M({ items: e, label: t = "Problem actions" }) {
	let [n, a] = (0, D.useState)(null), o = (0, D.useCallback)((e) => a(e?.closest("tr") ?? e?.parentElement ?? null), []);
	return e.length ? /* @__PURE__ */ (0, O.jsxs)("span", {
		ref: o,
		className: "shrink-0",
		onClick: (e) => e.stopPropagation(),
		onKeyDown: (e) => e.stopPropagation(),
		children: [/* @__PURE__ */ (0, O.jsx)(m, {
			menuLabel: t,
			items: e,
			trigger: /* @__PURE__ */ (0, O.jsx)(r, {
				variant: "ghost",
				size: "icon",
				className: "size-7 text-muted-foreground",
				"aria-label": t,
				children: /* @__PURE__ */ (0, O.jsx)(i, { icon: S })
			})
		}), /* @__PURE__ */ (0, O.jsx)(j, {
			contextTarget: n,
			menuLabel: t,
			menuItems: e
		})]
	}) : null;
}
export { j as n, M as t };
