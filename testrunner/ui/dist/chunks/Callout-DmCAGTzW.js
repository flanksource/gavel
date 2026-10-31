import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./jsx-runtime-lVm_Myqd.js";
import { r, t as i } from "./bundle-mjs--8zjrAdI.js";
// ../../../clicky-ui/packages/ui/src/lib/utils.ts
var a = /* @__PURE__ */ e(t(), 1), o = n(), s = [
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
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/class-variance-authority/0.7.1/085c9fb7731f516d7c41efcb924b7bd4b30ca77d3f239b7b053ed8b2e26e6903/node_modules/class-variance-authority/dist/index.mjs
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
}, p = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#d97706",
		...r
	};
	return /* @__PURE__ */ (0, o.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, o.jsx)("path", {
			fill: "currentColor",
			d: "M235.07 189.09L147.61 37.22a22.75 22.75 0 0 0-39.22 0L20.93 189.09a21.53 21.53 0 0 0 0 21.72A22.35 22.35 0 0 0 40.55 222h174.9a22.35 22.35 0 0 0 19.6-11.19a21.53 21.53 0 0 0 .02-21.72m-10.41 15.71a10.46 10.46 0 0 1-9.21 5.2H40.55a10.46 10.46 0 0 1-9.21-5.2a9.51 9.51 0 0 1 0-9.72l87.45-151.87a10.75 10.75 0 0 1 18.42 0l87.46 151.87a9.51 9.51 0 0 1-.01 9.72M122 144v-40a6 6 0 0 1 12 0v40a6 6 0 0 1-12 0m16 36a10 10 0 1 1-10-10a10 10 0 0 1 10 10"
		})
	});
}, {
	__source: "ph:warning-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "warning-triangle",
	__defaultColor: "#d97706",
	displayName: "UiWarningTriangle"
});
Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#d97706",
		...r
	};
	return /* @__PURE__ */ (0, o.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, o.jsx)("path", {
			fill: "currentColor",
			d: "M236.8 188.09L149.35 36.22a24.76 24.76 0 0 0-42.7 0L19.2 188.09a23.51 23.51 0 0 0 0 23.72A24.35 24.35 0 0 0 40.55 224h174.9a24.35 24.35 0 0 0 21.33-12.19a23.51 23.51 0 0 0 .02-23.72M120 104a8 8 0 0 1 16 0v40a8 8 0 0 1-16 0Zm8 88a12 12 0 1 1 12-12a12 12 0 0 1-12 12"
		})
	});
}, {
	__source: "ph:warning-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "warning-triangle",
	__defaultColor: "#d97706",
	displayName: "UiWarningTriangleFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiInfo.tsx
var m = Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#0ea5e9",
		...r
	};
	return /* @__PURE__ */ (0, o.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, o.jsx)("path", {
			fill: "currentColor",
			d: "M142 176a6 6 0 0 1-6 6a14 14 0 0 1-14-14v-40a2 2 0 0 0-2-2a6 6 0 0 1 0-12a14 14 0 0 1 14 14v40a2 2 0 0 0 2 2a6 6 0 0 1 6 6m-18-82a10 10 0 1 0-10-10a10 10 0 0 0 10 10m106 34A102 102 0 1 1 128 26a102.12 102.12 0 0 1 102 102m-12 0a90 90 0 1 0-90 90a90.1 90.1 0 0 0 90-90"
		})
	});
}, {
	__source: "ph:info-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "info",
	__defaultColor: "#0ea5e9",
	displayName: "UiInfo"
});
Object.assign(({ size: e = "1em", className: t, title: n, style: r, ...i }) => {
	let a = typeof t == "string" && /\btext-/.test(t) || r && r.color != null ? r : {
		color: "#0ea5e9",
		...r
	};
	return /* @__PURE__ */ (0, o.jsx)("svg", {
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
		children: /* @__PURE__ */ (0, o.jsx)("path", {
			fill: "currentColor",
			d: "M128 24a104 104 0 1 0 104 104A104.11 104.11 0 0 0 128 24m-4 48a12 12 0 1 1-12 12a12 12 0 0 1 12-12m12 112a16 16 0 0 1-16-16v-40a8 8 0 0 1 0-16a16 16 0 0 1 16 16v40a8 8 0 0 1 0 16"
		})
	});
}, {
	__source: "ph:info-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "info",
	__defaultColor: "#0ea5e9",
	displayName: "UiInfoFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiLightbulb.tsx
var h = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, o.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, o.jsx)("path", {
		fill: "currentColor",
		d: "M174 232a6 6 0 0 1-6 6H88a6 6 0 0 1 0-12h80a6 6 0 0 1 6 6m40-128a85.56 85.56 0 0 1-32.88 67.64A18.23 18.23 0 0 0 174 186v6a14 14 0 0 1-14 14H96a14 14 0 0 1-14-14v-6a18 18 0 0 0-7-14.23a85.59 85.59 0 0 1-33-67.24C41.74 57.91 79.39 19.12 125.93 18A86 86 0 0 1 214 104m-12 0a74 74 0 0 0-75.79-74C86.17 31 53.78 64.34 54 104.42a73.67 73.67 0 0 0 28.4 57.87A29.92 29.92 0 0 1 94 186v6a2 2 0 0 0 2 2h64a2 2 0 0 0 2-2v-6a30.18 30.18 0 0 1 11.7-23.78A73.59 73.59 0 0 0 202 104m-20.08-9A55.58 55.58 0 0 0 137 50.08a6 6 0 1 0-2 11.84c17.38 2.92 32.13 17.68 35.08 35.08a6 6 0 0 0 5.91 5a7 7 0 0 0 1-.08a6 6 0 0 0 4.93-6.92"
	})
}), {
	__source: "ph:lightbulb-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "lightbulb",
	displayName: "UiLightbulb"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, o.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, o.jsx)("path", {
		fill: "currentColor",
		d: "M176 232a8 8 0 0 1-8 8H88a8 8 0 0 1 0-16h80a8 8 0 0 1 8 8m40-128a87.55 87.55 0 0 1-33.64 69.21A16.24 16.24 0 0 0 176 186v6a16 16 0 0 1-16 16H96a16 16 0 0 1-16-16v-6a16 16 0 0 0-6.23-12.66A87.59 87.59 0 0 1 40 104.49C39.74 56.83 78.26 17.14 125.88 16A88 88 0 0 1 216 104m-32.11-9.34a57.6 57.6 0 0 0-46.56-46.55a8 8 0 0 0-2.66 15.78c16.57 2.79 30.63 16.85 33.44 33.45A8 8 0 0 0 176 104a9 9 0 0 0 1.35-.11a8 8 0 0 0 6.54-9.23"
	})
}), {
	__source: "ph:lightbulb-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "lightbulb",
	displayName: "UiLightbulbFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiStop.tsx
var g = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, o.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, o.jsx)("g", {
		fill: "none",
		children: /* @__PURE__ */ (0, o.jsx)("rect", {
			x: "2.5",
			y: "2.5",
			width: "11",
			height: "11",
			rx: "1.5",
			fill: "#FFF7F7",
			stroke: "#DB3B4B"
		})
	})
}), {
	__source: "jb-expui-run:stop",
	__viewBox: "0 0 16 16",
	__group: "runtime-process",
	__consumerName: "stop",
	displayName: "UiStop"
}), _ = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, o.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, o.jsx)("path", {
		fill: "currentColor",
		d: "M154.64 26.61a6 6 0 0 0-6.32.65L77.94 82H32a14 14 0 0 0-14 14v64a14 14 0 0 0 14 14h45.94l70.38 54.74A6 6 0 0 0 158 224V32a6 6 0 0 0-3.36-5.39M30 160V96a2 2 0 0 1 2-2h42v68H32a2 2 0 0 1-2-2m116 51.73l-60-46.66V90.93l60-46.66Zm50.53-108.85a38 38 0 0 1 0 50.24a6 6 0 1 1-9-7.94a26 26 0 0 0 0-34.37a6 6 0 0 1 9-7.93M246 128a77.86 77.86 0 0 1-19.86 52a6 6 0 1 1-8.94-8a66 66 0 0 0 0-88a6 6 0 1 1 8.94-8A77.86 77.86 0 0 1 246 128"
	})
}), {
	__source: "ph:speaker-high-light",
	__viewBox: "0 0 256 256",
	__group: "media",
	__consumerName: "speaker",
	displayName: "UiSpeaker"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, o.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, o.jsx)("path", {
		fill: "currentColor",
		d: "M160 32.25v191.44a8.29 8.29 0 0 1-3.91 7.18a8 8 0 0 1-9-.56l-65.57-51a4 4 0 0 1-1.52-3.15V79.84a4 4 0 0 1 1.55-3.15l65.57-51a8 8 0 0 1 10 .16a8.27 8.27 0 0 1 2.88 6.4M60 80H32a16 16 0 0 0-16 16v64a16 16 0 0 0 16 16h28a4 4 0 0 0 4-4V84a4 4 0 0 0-4-4m126.77 20.84a8 8 0 0 0-.72 11.3a24 24 0 0 1 0 31.72a8 8 0 1 0 12 10.58a40 40 0 0 0 0-52.88a8 8 0 0 0-11.31-.72Zm40.89-26.17a8 8 0 1 0-11.92 10.66a64 64 0 0 1 0 85.34a8 8 0 1 0 11.92 10.66a80 80 0 0 0 0-106.66"
	})
}), {
	__source: "ph:speaker-high-fill",
	__viewBox: "0 0 256 256",
	__group: "media",
	__consumerName: "speaker",
	displayName: "UiSpeakerFilled"
});
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@radix-ui/react-compose-refs/1.1.2/0ef1c5f33f439299023182b9a87b403bf76b779dc9de0fa656bb2e643e96eaf4/node_modules/@radix-ui/react-compose-refs/dist/index.mjs
function v(e, t) {
	if (typeof e == "function") return e(t);
	e != null && (e.current = t);
}
function y(...e) {
	return (t) => {
		let n = !1, r = e.map((e) => {
			let r = v(e, t);
			return !n && typeof r == "function" && (n = !0), r;
		});
		if (n) return () => {
			for (let t = 0; t < r.length; t++) {
				let n = r[t];
				typeof n == "function" ? n() : v(e[t], null);
			}
		};
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@radix-ui/react-slot/1.2.4/caad07b7528bb5b6b91006e6e66386e8b1aab5b3ecfd0f08f840f58bcc254d65/node_modules/@radix-ui/react-slot/dist/index.mjs
var b = Symbol.for("react.lazy"), x = a.use;
function ee(e) {
	return typeof e == "object" && !!e && "then" in e;
}
function S(e) {
	return typeof e == "object" && !!e && "$$typeof" in e && e.$$typeof === b && "_payload" in e && ee(e._payload);
}
// @__NO_SIDE_EFFECTS__
function C(e) {
	let t = /* @__PURE__ */ T(e), n = a.forwardRef((e, n) => {
		let { children: r, ...i } = e;
		S(r) && typeof x == "function" && (r = x(r._payload));
		let s = a.Children.toArray(r), c = s.find(D);
		if (c) {
			let e = c.props.children, r = s.map((t) => t === c ? a.Children.count(e) > 1 ? a.Children.only(null) : a.isValidElement(e) ? e.props.children : null : t);
			return /* @__PURE__ */ (0, o.jsx)(t, {
				...i,
				ref: n,
				children: a.isValidElement(e) ? a.cloneElement(e, void 0, r) : null
			});
		}
		return /* @__PURE__ */ (0, o.jsx)(t, {
			...i,
			ref: n,
			children: r
		});
	});
	return n.displayName = `${e}.Slot`, n;
}
var w = /* @__PURE__ */ C("Slot");
// @__NO_SIDE_EFFECTS__
function T(e) {
	let t = a.forwardRef((e, t) => {
		let { children: n, ...r } = e;
		if (S(n) && typeof x == "function" && (n = x(n._payload)), a.isValidElement(n)) {
			let e = te(n), i = O(r, n.props);
			return n.type !== a.Fragment && (i.ref = t ? y(t, e) : e), a.cloneElement(n, i);
		}
		return a.Children.count(n) > 1 ? a.Children.only(null) : null;
	});
	return t.displayName = `${e}.SlotClone`, t;
}
var E = Symbol("radix.slottable");
function D(e) {
	return a.isValidElement(e) && typeof e.type == "function" && "__radixId" in e.type && e.type.__radixId === E;
}
function O(e, t) {
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
function te(e) {
	let t = Object.getOwnPropertyDescriptor(e.props, "ref")?.get, n = t && "isReactWarning" in t && t.isReactWarning;
	return n ? e.ref : (t = Object.getOwnPropertyDescriptor(e, "ref")?.get, n = t && "isReactWarning" in t && t.isReactWarning, n ? e.props.ref : e.props.ref || e.ref);
}
// ../../../clicky-ui/packages/ui/src/components/button-variants.ts
var k = f("inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0", {
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
}), A = {
	sm: "size-4",
	md: "size-6",
	lg: "size-10"
}, j = {
	width: "clamp(1.25rem, 18cqmin, 2.5rem)",
	height: "clamp(1.25rem, 18cqmin, 2.5rem)"
}, M = { fontSize: "clamp(0.72rem, 4.5cqmin, 0.9rem)" }, N = (0, a.forwardRef)(({ className: e, ...t }, n) => /* @__PURE__ */ (0, o.jsxs)("svg", {
	ref: n,
	xmlns: "http://www.w3.org/2000/svg",
	viewBox: "0 0 24 24",
	fill: "currentColor",
	"aria-hidden": "true",
	className: l("shrink-0", e),
	...t,
	children: [
		/* @__PURE__ */ (0, o.jsx)("circle", {
			cx: "4",
			cy: "12",
			r: "3",
			children: /* @__PURE__ */ (0, o.jsx)("animate", {
				id: "clickyLoadDotA",
				fill: "freeze",
				attributeName: "opacity",
				begin: "0;clickyLoadDotC.end-0.25s",
				dur: "0.75s",
				values: "1;0.2"
			})
		}),
		/* @__PURE__ */ (0, o.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "3",
			opacity: "0.4",
			children: /* @__PURE__ */ (0, o.jsx)("animate", {
				fill: "freeze",
				attributeName: "opacity",
				begin: "clickyLoadDotA.begin+0.15s",
				dur: "0.75s",
				values: "1;0.2"
			})
		}),
		/* @__PURE__ */ (0, o.jsx)("circle", {
			cx: "20",
			cy: "12",
			r: "3",
			opacity: "0.3",
			children: /* @__PURE__ */ (0, o.jsx)("animate", {
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
N.displayName = "LoadingDots";
var P = (0, a.forwardRef)(({ className: e, ...t }, n) => /* @__PURE__ */ (0, o.jsx)("div", {
	ref: n,
	"aria-hidden": !0,
	className: l("pointer-events-none absolute inset-x-0 top-0 z-20 h-0.5 overflow-hidden bg-primary/20", e),
	...t,
	children: /* @__PURE__ */ (0, o.jsx)("div", { className: "h-full w-1/2 animate-pulse rounded-full bg-primary" })
}));
P.displayName = "LoadingBar";
var F = (0, a.forwardRef)(({ size: e = "responsive", variant: t = "inline", label: n, className: r, ...i }, a) => {
	let s = t === "inline" && e === "responsive" ? "sm" : e, c = s === "responsive", u = c ? "" : A[s], d = c ? j : void 0;
	return t === "centered" ? /* @__PURE__ */ (0, o.jsxs)("div", {
		ref: a,
		role: "status",
		"aria-label": "Loading",
		className: l("flex h-full min-h-32 w-full flex-col items-center justify-center gap-3 text-muted-foreground", "[container-type:size]", r),
		...i,
		children: [/* @__PURE__ */ (0, o.jsx)(N, {
			className: u,
			style: d
		}), n != null && /* @__PURE__ */ (0, o.jsx)("span", {
			className: c ? void 0 : "text-sm",
			style: c ? M : void 0,
			children: n
		})]
	}) : /* @__PURE__ */ (0, o.jsxs)("div", {
		ref: a,
		role: "status",
		"aria-label": "Loading",
		className: l("inline-flex items-center gap-2 text-muted-foreground", r),
		...i,
		children: [/* @__PURE__ */ (0, o.jsx)(N, {
			className: u,
			style: d
		}), n != null && /* @__PURE__ */ (0, o.jsx)("span", {
			className: "text-sm",
			children: n
		})]
	});
});
F.displayName = "Loading";
// ../../../clicky-ui/packages/ui/src/components/button.tsx
var I = (0, a.forwardRef)(({ children: e, className: t, disabled: n, loading: r = !1, loadingLabel: i, variant: a, size: s, asChild: c = !1, ...u }, d) => {
	let f = c ? w : "button", p = i === void 0 ? e : i;
	return c ? /* @__PURE__ */ (0, o.jsx)(f, {
		ref: d,
		"aria-busy": r || void 0,
		className: l(k({
			variant: a,
			size: s
		}), t),
		disabled: n || r,
		...u,
		children: e
	}) : /* @__PURE__ */ (0, o.jsxs)(f, {
		ref: d,
		"aria-busy": r || void 0,
		className: l(k({
			variant: a,
			size: s
		}), t),
		disabled: n || r,
		...u,
		children: [r && /* @__PURE__ */ (0, o.jsx)(N, { className: "size-4" }), p]
	});
});
I.displayName = "Button";
// ../../../clicky-ui/packages/ui/src/lib/collections.ts
function L(e) {
	return typeof e == "object" && !!e && !Array.isArray(e);
}
function R(e, t, n) {
	return e.map((e, r) => r === t ? n : e);
}
function z(e, t) {
	return e.filter((e, n) => n !== t);
}
function B(e, t, n) {
	if (n < 0 || n >= e.length) return e;
	let r = [...e], [i] = r.splice(t, 1);
	return r.splice(n, 0, i), r;
}
function V(e, t, n) {
	let r = e[t], i = n ? n(r) : Array.isArray(r) ? [...r] : L(r) ? { ...r } : r, a = [...e];
	return a.splice(t + 1, 0, i), a;
}
// ../../../clicky-ui/packages/ui/src/components/json-schema-form-size.ts
var H = {
	xs: "h-7 px-1.5 text-xs",
	sm: "h-8 px-2 text-xs",
	md: "h-9 px-2 text-sm",
	lg: "h-10 px-3 text-base",
	xl: "h-11 px-3.5 text-lg"
}, U = {
	xs: "text-xs",
	sm: "text-xs",
	md: "text-sm",
	lg: "text-base",
	xl: "text-lg"
}, W = {
	xs: "h-7",
	sm: "h-8",
	md: "h-9",
	lg: "h-10",
	xl: "h-11"
}, G = {
	xs: "min-h-7",
	sm: "min-h-8",
	md: "min-h-9",
	lg: "min-h-10",
	xl: "min-h-11"
}, K = {
	xs: "gap-1",
	sm: "gap-1.5",
	md: "gap-2.5",
	lg: "gap-4",
	xl: "gap-4"
}, q = {
	xs: "gap-0.5",
	sm: "gap-1",
	md: "gap-1.5",
	lg: "gap-2",
	xl: "gap-2"
}, J = {
	xs: "gap-1",
	sm: "gap-1",
	md: "gap-1.5",
	lg: "gap-2",
	xl: "gap-2"
}, Y = [
	"note",
	"tip",
	"important",
	"warning",
	"caution"
], X = [...Y, "default"], Z = {
	note: {
		accent: "border-l-[3px] border-l-blue-500 bg-blue-500/5 rounded-r",
		emphasis: "border-2 border-blue-500 bg-blue-500/10 rounded",
		subtle: "border-blue-500/30 bg-blue-500/5",
		badge: "bg-blue-600 text-white",
		label: "text-blue-700 dark:text-blue-300",
		source: "text-blue-600/70 dark:text-blue-300/70",
		icon: "text-blue-600 dark:text-blue-400",
		title: "text-blue-900 dark:text-blue-100",
		content: "text-blue-950/90 dark:text-blue-100/90"
	},
	tip: {
		accent: "border-l-[3px] border-l-emerald-500 bg-emerald-500/5 rounded-r",
		emphasis: "border-2 border-emerald-500 bg-emerald-500/10 rounded",
		subtle: "border-emerald-500/30 bg-emerald-500/5",
		badge: "bg-emerald-600 text-white",
		label: "text-emerald-700 dark:text-emerald-300",
		source: "text-emerald-600/70 dark:text-emerald-300/70",
		icon: "text-emerald-600 dark:text-emerald-400",
		title: "text-emerald-900 dark:text-emerald-100",
		content: "text-emerald-950/90 dark:text-emerald-100/90"
	},
	important: {
		accent: "border-l-[3px] border-l-purple-500 bg-purple-500/5 rounded-r",
		emphasis: "border-2 border-purple-500 bg-purple-500/10 rounded",
		subtle: "border-purple-500/30 bg-purple-500/5",
		badge: "bg-purple-600 text-white",
		label: "text-purple-700 dark:text-purple-300",
		source: "text-purple-600/70 dark:text-purple-300/70",
		icon: "text-purple-600 dark:text-purple-400",
		title: "text-purple-900 dark:text-purple-100",
		content: "text-purple-950/90 dark:text-purple-100/90"
	},
	warning: {
		accent: "border-l-[3px] border-l-amber-500 bg-amber-500/10 rounded-r",
		emphasis: "border-2 border-amber-500 bg-amber-500/15 rounded",
		subtle: "border-amber-500/40 bg-amber-500/10",
		badge: "bg-amber-600 text-white",
		label: "text-amber-700 dark:text-amber-300",
		source: "text-amber-600/70 dark:text-amber-300/70",
		icon: "text-amber-600 dark:text-amber-400",
		title: "text-amber-900 dark:text-amber-100",
		content: "text-amber-950/90 dark:text-amber-100/90"
	},
	caution: {
		accent: "border-l-[3px] border-l-red-500 bg-red-500/5 rounded-r",
		emphasis: "border-2 border-red-500 bg-red-500/10 rounded",
		subtle: "border-red-500/40 bg-red-500/10",
		badge: "bg-red-600 text-white",
		label: "text-red-700 dark:text-red-300",
		source: "text-red-600/70 dark:text-red-300/70",
		icon: "text-red-600 dark:text-red-400",
		title: "text-red-900 dark:text-red-100",
		content: "text-red-950/90 dark:text-red-100/90"
	},
	default: {
		accent: "border-l-[3px] border-l-border bg-muted/30 rounded-r",
		emphasis: "border-2 border-border bg-muted/50 rounded",
		subtle: "border-border bg-muted/30",
		badge: "bg-foreground text-background",
		label: "text-foreground",
		source: "text-muted-foreground",
		icon: "text-muted-foreground",
		title: "text-foreground",
		content: "text-muted-foreground"
	}
}, ne = {
	note: "Note",
	tip: "Tip",
	important: "Important",
	warning: "Warning",
	caution: "Caution",
	default: ""
};
function Q(e) {
	return Z[e];
}
function $(e, t, n) {
	if (!e.includes(t)) throw Error(`Unknown Callout ${n} "${t}" — expected one of: ${e.join(", ")}`);
	return t;
}
// ../../../clicky-ui/packages/ui/src/data/Callout.tsx
var re = {
	note: m,
	tip: h,
	important: _,
	warning: p,
	caution: g
};
function ie({ children: e, variant: t = "default", label: n, badge: r, source: i, icon: a, emphasis: s = !1, title: c, className: u }) {
	let d = $(X, t, "variant"), f = a === void 0 ? void 0 : $(Y, a, "icon"), p = Q(d), m = s !== !1 && s !== "false", h = n ?? ne[d], g = f ?? (d === "default" ? void 0 : d), _ = g && re[g];
	return /* @__PURE__ */ (0, o.jsxs)("div", {
		className: l("not-prose my-4 p-density-3", m ? p.emphasis : p.accent, u),
		children: [
			!!(r || h || i || _) && /* @__PURE__ */ (0, o.jsxs)("div", {
				className: "mb-1 flex items-baseline gap-2",
				children: [
					r && /* @__PURE__ */ (0, o.jsx)("span", {
						className: l("rounded px-1.5 py-0.5 text-xs font-bold", p.badge),
						children: r
					}),
					_ && /* @__PURE__ */ (0, o.jsx)("span", {
						className: l("self-center", p.icon),
						children: /* @__PURE__ */ (0, o.jsx)(_, {
							className: p.icon,
							size: "1em"
						})
					}),
					h && /* @__PURE__ */ (0, o.jsx)("span", {
						className: l("text-xs font-semibold uppercase tracking-wide", p.label),
						children: h
					}),
					i && /* @__PURE__ */ (0, o.jsx)("span", {
						className: l("text-xs", p.source),
						children: i
					})
				]
			}),
			c && /* @__PURE__ */ (0, o.jsx)("h3", {
				className: l("mb-2 text-sm font-bold", p.title),
				children: c
			}),
			/* @__PURE__ */ (0, o.jsx)("div", {
				className: l("text-sm leading-snug", p.content),
				children: e
			})
		]
	});
}
export { p as C, m as S, l as T, I as _, W as a, _ as b, q as c, K as d, V as f, R as g, z as h, Q as i, H as l, B as m, Y as n, G as o, L as p, X as r, J as s, ie as t, U as u, P as v, f as w, g as x, w as y };
