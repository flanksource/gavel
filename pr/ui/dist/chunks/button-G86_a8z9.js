import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./react-DHvKIBht.js";
import { t as r } from "./clsx-BitgEGBo.js";
import { t as i } from "./bundle-mjs-CD_lylFG.js";
// ../../../clicky-ui/packages/ui/src/lib/utils.ts
var a = /* @__PURE__ */ e(t(), 1), o = /* @__PURE__ */ e(n(), 1), s = [
	"density-1",
	"density-2",
	"density-3",
	"density-4"
], c = i({ extend: { classGroups: {
	h: [{ h: ["control-h", ...s] }],
	w: [{ w: ["control-h"] }],
	p: [{ p: s }],
	px: [{ px: ["control-px", ...s] }],
	py: [{ py: s }],
	m: [{ m: s }],
	mt: [{ mt: s }],
	mb: [{ mb: s }],
	gap: [{ gap: s }],
	"font-size": [{ text: ["density-base"] }]
} } });
function l(...e) {
	return c(r(e));
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@radix-ui/react-compose-refs/1.1.2/0ef1c5f33f439299023182b9a87b403bf76b779dc9de0fa656bb2e643e96eaf4/node_modules/@radix-ui/react-compose-refs/dist/index.mjs
function u(e, t) {
	if (typeof e == "function") return e(t);
	e != null && (e.current = t);
}
function d(...e) {
	return (t) => {
		let n = !1, r = e.map((e) => {
			let r = u(e, t);
			return !n && typeof r == "function" && (n = !0), r;
		});
		if (n) return () => {
			for (let t = 0; t < r.length; t++) {
				let n = r[t];
				typeof n == "function" ? n() : u(e[t], null);
			}
		};
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@radix-ui/react-slot/1.2.4/caad07b7528bb5b6b91006e6e66386e8b1aab5b3ecfd0f08f840f58bcc254d65/node_modules/@radix-ui/react-slot/dist/index.mjs
var f = Symbol.for("react.lazy"), p = o.use;
function m(e) {
	return typeof e == "object" && !!e && "then" in e;
}
function h(e) {
	return typeof e == "object" && !!e && "$$typeof" in e && e.$$typeof === f && "_payload" in e && m(e._payload);
}
// @__NO_SIDE_EFFECTS__
function g(e) {
	let t = /* @__PURE__ */ v(e), n = o.forwardRef((e, n) => {
		let { children: r, ...i } = e;
		h(r) && typeof p == "function" && (r = p(r._payload));
		let s = o.Children.toArray(r), c = s.find(b);
		if (c) {
			let e = c.props.children, r = s.map((t) => t === c ? o.Children.count(e) > 1 ? o.Children.only(null) : o.isValidElement(e) ? e.props.children : null : t);
			return /* @__PURE__ */ (0, a.jsx)(t, {
				...i,
				ref: n,
				children: o.isValidElement(e) ? o.cloneElement(e, void 0, r) : null
			});
		}
		return /* @__PURE__ */ (0, a.jsx)(t, {
			...i,
			ref: n,
			children: r
		});
	});
	return n.displayName = `${e}.Slot`, n;
}
var _ = /* @__PURE__ */ g("Slot");
// @__NO_SIDE_EFFECTS__
function v(e) {
	let t = o.forwardRef((e, t) => {
		let { children: n, ...r } = e;
		if (h(n) && typeof p == "function" && (n = p(n._payload)), o.isValidElement(n)) {
			let e = S(n), i = x(r, n.props);
			return n.type !== o.Fragment && (i.ref = t ? d(t, e) : e), o.cloneElement(n, i);
		}
		return o.Children.count(n) > 1 ? o.Children.only(null) : null;
	});
	return t.displayName = `${e}.SlotClone`, t;
}
var y = Symbol("radix.slottable");
function b(e) {
	return o.isValidElement(e) && typeof e.type == "function" && "__radixId" in e.type && e.type.__radixId === y;
}
function x(e, t) {
	let n = { ...t };
	for (let r in t) {
		let i = e[r], a = t[r];
		/^on[A-Z]/.test(r) ? i && a ? n[r] = (...e) => {
			let t = a(...e);
			return i(...e), t;
		} : i && (n[r] = i) : r === "style" ? n[r] = {
			...i,
			...a
		} : r === "className" && (n[r] = [i, a].filter(Boolean).join(" "));
	}
	return {
		...e,
		...n
	};
}
function S(e) {
	let t = Object.getOwnPropertyDescriptor(e.props, "ref")?.get, n = t && "isReactWarning" in t && t.isReactWarning;
	return n ? e.ref : (t = Object.getOwnPropertyDescriptor(e, "ref")?.get, n = t && "isReactWarning" in t && t.isReactWarning, n ? e.props.ref : e.props.ref || e.ref);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/class-variance-authority/0.7.1/085c9fb7731f516d7c41efcb924b7bd4b30ca77d3f239b7b053ed8b2e26e6903/node_modules/class-variance-authority/dist/index.mjs
var C = (e) => typeof e == "boolean" ? `${e}` : e === 0 ? "0" : e, w = r, T = (e, t) => (n) => {
	if (t?.variants == null) return w(e, n?.class, n?.className);
	let { variants: r, defaultVariants: i } = t, a = Object.keys(r).map((e) => {
		let t = n?.[e], a = i?.[e];
		if (t === null) return null;
		let o = C(t) || C(a);
		return r[e][o];
	}), o = n && Object.entries(n).reduce((e, t) => {
		let [n, r] = t;
		return r === void 0 || (e[n] = r), e;
	}, {});
	return w(e, a, t?.compoundVariants?.reduce((e, t) => {
		let { class: n, className: r, ...a } = t;
		return Object.entries(a).every((e) => {
			let [t, n] = e;
			return Array.isArray(n) ? n.includes({
				...i,
				...o
			}[t]) : {
				...i,
				...o
			}[t] === n;
		}) ? [
			...e,
			n,
			r
		] : e;
	}, []), n?.class, n?.className);
}, E = T("inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0", {
	variants: {
		variant: {
			default: "bg-primary text-primary-foreground hover:bg-primary/90",
			destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
			outline: "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
			secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
			ghost: "hover:bg-accent hover:text-accent-foreground",
			link: "text-primary underline-offset-4 hover:underline"
		},
		size: {
			default: "h-control-h px-control-px py-2",
			sm: "h-8 rounded-md px-3 text-xs",
			lg: "h-10 rounded-md px-8",
			icon: "h-control-h w-control-h"
		}
	},
	defaultVariants: {
		variant: "default",
		size: "default"
	}
}), D = {
	sm: "size-4",
	md: "size-6",
	lg: "size-10"
}, O = {
	width: "clamp(1.25rem, 18cqmin, 2.5rem)",
	height: "clamp(1.25rem, 18cqmin, 2.5rem)"
}, k = { fontSize: "clamp(0.72rem, 4.5cqmin, 0.9rem)" }, A = (0, o.forwardRef)(({ className: e, ...t }, n) => /* @__PURE__ */ (0, a.jsxs)("svg", {
	ref: n,
	xmlns: "http://www.w3.org/2000/svg",
	viewBox: "0 0 24 24",
	fill: "currentColor",
	"aria-hidden": "true",
	className: l("shrink-0", e),
	...t,
	children: [
		/* @__PURE__ */ (0, a.jsx)("circle", {
			cx: "4",
			cy: "12",
			r: "3",
			children: /* @__PURE__ */ (0, a.jsx)("animate", {
				id: "clickyLoadDotA",
				fill: "freeze",
				attributeName: "opacity",
				begin: "0;clickyLoadDotC.end-0.25s",
				dur: "0.75s",
				values: "1;0.2"
			})
		}),
		/* @__PURE__ */ (0, a.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "3",
			opacity: "0.4",
			children: /* @__PURE__ */ (0, a.jsx)("animate", {
				fill: "freeze",
				attributeName: "opacity",
				begin: "clickyLoadDotA.begin+0.15s",
				dur: "0.75s",
				values: "1;0.2"
			})
		}),
		/* @__PURE__ */ (0, a.jsx)("circle", {
			cx: "20",
			cy: "12",
			r: "3",
			opacity: "0.3",
			children: /* @__PURE__ */ (0, a.jsx)("animate", {
				id: "clickyLoadDotC",
				fill: "freeze",
				attributeName: "opacity",
				begin: "clickyLoadDotA.begin+0.3s",
				dur: "0.75s",
				values: "1;0.2"
			})
		})
	]
}));
A.displayName = "LoadingDots";
var j = (0, o.forwardRef)(({ className: e, ...t }, n) => /* @__PURE__ */ (0, a.jsx)("div", {
	ref: n,
	"aria-hidden": !0,
	className: l("pointer-events-none absolute inset-x-0 top-0 z-20 h-0.5 overflow-hidden bg-primary/20", e),
	...t,
	children: /* @__PURE__ */ (0, a.jsx)("div", { className: "h-full w-1/2 animate-pulse rounded-full bg-primary" })
}));
j.displayName = "LoadingBar";
var M = (0, o.forwardRef)(({ size: e = "responsive", variant: t = "inline", label: n, className: r, ...i }, o) => {
	let s = t === "inline" && e === "responsive" ? "sm" : e, c = s === "responsive", u = c ? "" : D[s], d = c ? O : void 0;
	return t === "centered" ? /* @__PURE__ */ (0, a.jsxs)("div", {
		ref: o,
		role: "status",
		"aria-label": "Loading",
		className: l("flex h-full min-h-32 w-full flex-col items-center justify-center gap-3 text-muted-foreground", "[container-type:size]", r),
		...i,
		children: [/* @__PURE__ */ (0, a.jsx)(A, {
			className: u,
			style: d
		}), n != null && /* @__PURE__ */ (0, a.jsx)("span", {
			className: c ? void 0 : "text-sm",
			style: c ? k : void 0,
			children: n
		})]
	}) : /* @__PURE__ */ (0, a.jsxs)("div", {
		ref: o,
		role: "status",
		"aria-label": "Loading",
		className: l("inline-flex items-center gap-2 text-muted-foreground", r),
		...i,
		children: [/* @__PURE__ */ (0, a.jsx)(A, {
			className: u,
			style: d
		}), n != null && /* @__PURE__ */ (0, a.jsx)("span", {
			className: "text-sm",
			children: n
		})]
	});
});
M.displayName = "Loading";
// ../../../clicky-ui/packages/ui/src/components/button.tsx
var N = (0, o.forwardRef)(({ children: e, className: t, disabled: n, loading: r = !1, loadingLabel: i, variant: o, size: s, asChild: c = !1, ...u }, d) => {
	let f = c ? _ : "button", p = i === void 0 ? e : i;
	return c ? /* @__PURE__ */ (0, a.jsx)(f, {
		ref: d,
		"aria-busy": r || void 0,
		className: l(E({
			variant: o,
			size: s
		}), t),
		disabled: n || r,
		...u,
		children: e
	}) : /* @__PURE__ */ (0, a.jsxs)(f, {
		ref: d,
		"aria-busy": r || void 0,
		className: l(E({
			variant: o,
			size: s
		}), t),
		disabled: n || r,
		...u,
		children: [r && /* @__PURE__ */ (0, a.jsx)(A, { className: "size-4" }), p]
	});
});
N.displayName = "Button";
export { _ as a, T as i, M as n, l as o, j as r, N as t };
