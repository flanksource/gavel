import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-a1ZmknZu.js";
import { t as r } from "./clsx-BULO-GtV.js";
import { t as i } from "./bundle-mjs-CU1JUvhy.js";
import { t as a } from "./dist-CMVKLuPp.js";
// ../../../clicky-ui/packages/ui/src/lib/utils.ts
var o = /* @__PURE__ */ e(t(), 1), s = [
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/class-variance-authority/0.7.1/085c9fb7731f516d7c41efcb924b7bd4b30ca77d3f239b7b053ed8b2e26e6903/node_modules/class-variance-authority/dist/index.mjs
var u = (e) => typeof e == "boolean" ? `${e}` : e === 0 ? "0" : e, d = r, f = (e, t) => (n) => {
	if (t?.variants == null) return d(e, n?.class, n?.className);
	let { variants: r, defaultVariants: i } = t, a = Object.keys(r).map((e) => {
		let t = n?.[e], a = i?.[e];
		if (t === null) return null;
		let o = u(t) || u(a);
		return r[e][o];
	}), o = n && Object.entries(n).reduce((e, t) => {
		let [n, r] = t;
		return r === void 0 || (e[n] = r), e;
	}, {});
	return d(e, a, t?.compoundVariants?.reduce((e, t) => {
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
}, p = f("inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0", {
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
			xs: "h-6 gap-1 rounded-md px-2 text-[11px]",
			sm: "h-8 rounded-md px-3 text-xs",
			lg: "h-10 rounded-md px-8",
			icon: "h-control-h w-control-h"
		}
	},
	defaultVariants: {
		variant: "default",
		size: "default"
	}
}), m = /* @__PURE__ */ e(n(), 1), h = {
	sm: "size-4",
	md: "size-6",
	lg: "size-10"
}, g = {
	width: "clamp(1.25rem, 18cqmin, 2.5rem)",
	height: "clamp(1.25rem, 18cqmin, 2.5rem)"
}, _ = { fontSize: "clamp(0.72rem, 4.5cqmin, 0.9rem)" }, v = (0, o.forwardRef)(({ className: e, ...t }, n) => /* @__PURE__ */ (0, m.jsxs)("svg", {
	ref: n,
	xmlns: "http://www.w3.org/2000/svg",
	viewBox: "0 0 24 24",
	fill: "currentColor",
	"aria-hidden": "true",
	className: l("shrink-0", e),
	...t,
	children: [
		/* @__PURE__ */ (0, m.jsx)("circle", {
			cx: "4",
			cy: "12",
			r: "3",
			children: /* @__PURE__ */ (0, m.jsx)("animate", {
				id: "clickyLoadDotA",
				fill: "freeze",
				attributeName: "opacity",
				begin: "0;clickyLoadDotC.end-0.25s",
				dur: "0.75s",
				values: "1;0.2"
			})
		}),
		/* @__PURE__ */ (0, m.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "3",
			opacity: "0.4",
			children: /* @__PURE__ */ (0, m.jsx)("animate", {
				fill: "freeze",
				attributeName: "opacity",
				begin: "clickyLoadDotA.begin+0.15s",
				dur: "0.75s",
				values: "1;0.2"
			})
		}),
		/* @__PURE__ */ (0, m.jsx)("circle", {
			cx: "20",
			cy: "12",
			r: "3",
			opacity: "0.3",
			children: /* @__PURE__ */ (0, m.jsx)("animate", {
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
v.displayName = "LoadingDots";
var y = (0, o.forwardRef)(({ className: e, ...t }, n) => /* @__PURE__ */ (0, m.jsx)("div", {
	ref: n,
	"aria-hidden": !0,
	className: l("pointer-events-none absolute inset-x-0 top-0 z-20 h-0.5 overflow-hidden bg-primary/20", e),
	...t,
	children: /* @__PURE__ */ (0, m.jsx)("div", { className: "h-full w-1/2 animate-pulse rounded-full bg-primary" })
}));
y.displayName = "LoadingBar";
var b = (0, o.forwardRef)(({ size: e = "responsive", variant: t = "inline", label: n, className: r, ...i }, a) => {
	let o = t === "inline" && e === "responsive" ? "sm" : e, s = o === "responsive", c = s ? "" : h[o], u = s ? g : void 0;
	return t === "centered" ? /* @__PURE__ */ (0, m.jsxs)("div", {
		ref: a,
		role: "status",
		"aria-label": "Loading",
		className: l("flex h-full min-h-32 w-full flex-col items-center justify-center gap-3 text-muted-foreground", "[container-type:size]", r),
		...i,
		children: [/* @__PURE__ */ (0, m.jsx)(v, {
			className: c,
			style: u
		}), n != null && /* @__PURE__ */ (0, m.jsx)("span", {
			className: s ? void 0 : "text-sm",
			style: s ? _ : void 0,
			children: n
		})]
	}) : /* @__PURE__ */ (0, m.jsxs)("div", {
		ref: a,
		role: "status",
		"aria-label": "Loading",
		className: l("inline-flex items-center gap-2 text-muted-foreground", r),
		...i,
		children: [/* @__PURE__ */ (0, m.jsx)(v, {
			className: c,
			style: u
		}), n != null && /* @__PURE__ */ (0, m.jsx)("span", {
			className: "text-sm",
			children: n
		})]
	});
});
b.displayName = "Loading";
// ../../../clicky-ui/packages/ui/src/components/button.tsx
var x = (0, o.forwardRef)(({ children: e, className: t, disabled: n, loading: r = !1, loadingLabel: i, variant: o, size: s, asChild: c = !1, ...u }, d) => {
	let f = c ? a : "button", h = r && i !== void 0 ? i : e;
	return c ? /* @__PURE__ */ (0, m.jsx)(f, {
		ref: d,
		"aria-busy": r || void 0,
		className: l(p({
			variant: o,
			size: s
		}), t),
		disabled: n || r,
		...u,
		children: e
	}) : /* @__PURE__ */ (0, m.jsxs)(f, {
		ref: d,
		"aria-busy": r || void 0,
		className: l(p({
			variant: o,
			size: s
		}), t),
		disabled: n || r,
		...u,
		children: [r && /* @__PURE__ */ (0, m.jsx)(v, { className: "size-4" }), h]
	});
});
x.displayName = "Button";
export { f as a, v as i, b as n, l as o, y as r, x as t };
