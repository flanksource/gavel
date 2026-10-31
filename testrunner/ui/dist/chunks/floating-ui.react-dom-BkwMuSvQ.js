import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t } from "./jsx-runtime-lVm_Myqd.js";
import { t as n } from "./react-dom-BBUmKE87.js";
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@floating-ui/utils/0.2.11/87057d931cf909b87bee50274fc4edf36424eacbed3ab4a94fba7ad140b175bb/node_modules/@floating-ui/utils/dist/floating-ui.utils.dom.mjs
function r() {
	return typeof window < "u";
}
function i(e) {
	return s(e) ? (e.nodeName || "").toLowerCase() : "#document";
}
function a(e) {
	var t;
	return (e == null || (t = e.ownerDocument) == null ? void 0 : t.defaultView) || window;
}
function o(e) {
	return ((s(e) ? e.ownerDocument : e.document) || window.document)?.documentElement;
}
function s(e) {
	return r() ? e instanceof Node || e instanceof a(e).Node : !1;
}
function c(e) {
	return r() ? e instanceof Element || e instanceof a(e).Element : !1;
}
function l(e) {
	return r() ? e instanceof HTMLElement || e instanceof a(e).HTMLElement : !1;
}
function u(e) {
	return !r() || typeof ShadowRoot > "u" ? !1 : e instanceof ShadowRoot || e instanceof a(e).ShadowRoot;
}
function d(e) {
	let { overflow: t, overflowX: n, overflowY: r, display: i } = S(e);
	return /auto|scroll|overlay|hidden|clip/.test(t + r + n) && i !== "inline" && i !== "contents";
}
function f(e) {
	return /^(table|td|th)$/.test(i(e));
}
function p(e) {
	try {
		if (e.matches(":popover-open")) return !0;
	} catch {}
	try {
		return e.matches(":modal");
	} catch {
		return !1;
	}
}
var m = /transform|translate|scale|rotate|perspective|filter/, h = /paint|layout|strict|content/, g = (e) => !!e && e !== "none", _;
function v(e) {
	let t = c(e) ? S(e) : e;
	return g(t.transform) || g(t.translate) || g(t.scale) || g(t.rotate) || g(t.perspective) || !b() && (g(t.backdropFilter) || g(t.filter)) || m.test(t.willChange || "") || h.test(t.contain || "");
}
function y(e) {
	let t = w(e);
	for (; l(t) && !x(t);) {
		if (v(t)) return t;
		if (p(t)) return null;
		t = w(t);
	}
	return null;
}
function b() {
	return _ ??= typeof CSS < "u" && CSS.supports && CSS.supports("-webkit-backdrop-filter", "none"), _;
}
function x(e) {
	return /^(html|body|#document)$/.test(i(e));
}
function S(e) {
	return a(e).getComputedStyle(e);
}
function C(e) {
	return c(e) ? {
		scrollLeft: e.scrollLeft,
		scrollTop: e.scrollTop
	} : {
		scrollLeft: e.scrollX,
		scrollTop: e.scrollY
	};
}
function w(e) {
	if (i(e) === "html") return e;
	let t = e.assignedSlot || e.parentNode || u(e) && e.host || o(e);
	return u(t) ? t.host : t;
}
function T(e) {
	let t = w(e);
	return x(t) ? e.ownerDocument ? e.ownerDocument.body : e.body : l(t) && d(t) ? t : T(t);
}
function E(e, t, n) {
	t === void 0 && (t = []), n === void 0 && (n = !0);
	let r = T(e), i = r === e.ownerDocument?.body, o = a(r);
	if (i) {
		let e = D(o);
		return t.concat(o, o.visualViewport || [], d(r) ? r : [], e && n ? E(e) : []);
	}
	return t.concat(r, E(r, [], n));
}
function D(e) {
	return e.parent && Object.getPrototypeOf(e.parent) ? e.frameElement : null;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@floating-ui/utils/0.2.11/87057d931cf909b87bee50274fc4edf36424eacbed3ab4a94fba7ad140b175bb/node_modules/@floating-ui/utils/dist/floating-ui.utils.mjs
var O = [
	"top",
	"right",
	"bottom",
	"left"
], k = Math.min, A = Math.max, j = Math.round, M = Math.floor, N = (e) => ({
	x: e,
	y: e
}), P = {
	left: "right",
	right: "left",
	bottom: "top",
	top: "bottom"
};
function F(e, t, n) {
	return A(e, k(t, n));
}
function I(e, t) {
	return typeof e == "function" ? e(t) : e;
}
function L(e) {
	return e.split("-")[0];
}
function R(e) {
	return e.split("-")[1];
}
function z(e) {
	return e === "x" ? "y" : "x";
}
function B(e) {
	return e === "y" ? "height" : "width";
}
function V(e) {
	let t = e[0];
	return t === "t" || t === "b" ? "y" : "x";
}
function H(e) {
	return z(V(e));
}
function ee(e, t, n) {
	n === void 0 && (n = !1);
	let r = R(e), i = H(e), a = B(i), o = i === "x" ? r === (n ? "end" : "start") ? "right" : "left" : r === "start" ? "bottom" : "top";
	return t.reference[a] > t.floating[a] && (o = W(o)), [o, W(o)];
}
function te(e) {
	let t = W(e);
	return [
		U(e),
		t,
		U(t)
	];
}
function U(e) {
	return e.includes("start") ? e.replace("start", "end") : e.replace("end", "start");
}
var ne = ["left", "right"], re = ["right", "left"], ie = ["top", "bottom"], ae = ["bottom", "top"];
function oe(e, t, n) {
	switch (e) {
		case "top":
		case "bottom": return n ? t ? re : ne : t ? ne : re;
		case "left":
		case "right": return t ? ie : ae;
		default: return [];
	}
}
function se(e, t, n, r) {
	let i = R(e), a = oe(L(e), n === "start", r);
	return i && (a = a.map((e) => e + "-" + i), t && (a = a.concat(a.map(U)))), a;
}
function W(e) {
	let t = L(e);
	return P[t] + e.slice(t.length);
}
function ce(e) {
	return {
		top: 0,
		right: 0,
		bottom: 0,
		left: 0,
		...e
	};
}
function le(e) {
	return typeof e == "number" ? {
		top: e,
		right: e,
		bottom: e,
		left: e
	} : ce(e);
}
function G(e) {
	let { x: t, y: n, width: r, height: i } = e;
	return {
		width: r,
		height: i,
		top: n,
		left: t,
		right: t + r,
		bottom: n + i,
		x: t,
		y: n
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@floating-ui/core/1.7.5/24175fea28b0720fe8ec9199be79a66ff8370b61629ec7fe34bbf5103d0215a0/node_modules/@floating-ui/core/dist/floating-ui.core.mjs
function ue(e, t, n) {
	let { reference: r, floating: i } = e, a = V(t), o = H(t), s = B(o), c = L(t), l = a === "y", u = r.x + r.width / 2 - i.width / 2, d = r.y + r.height / 2 - i.height / 2, f = r[s] / 2 - i[s] / 2, p;
	switch (c) {
		case "top":
			p = {
				x: u,
				y: r.y - i.height
			};
			break;
		case "bottom":
			p = {
				x: u,
				y: r.y + r.height
			};
			break;
		case "right":
			p = {
				x: r.x + r.width,
				y: d
			};
			break;
		case "left":
			p = {
				x: r.x - i.width,
				y: d
			};
			break;
		default: p = {
			x: r.x,
			y: r.y
		};
	}
	switch (R(t)) {
		case "start":
			p[o] -= f * (n && l ? -1 : 1);
			break;
		case "end": p[o] += f * (n && l ? -1 : 1);
	}
	return p;
}
async function de(e, t) {
	t === void 0 && (t = {});
	let { x: n, y: r, platform: i, rects: a, elements: o, strategy: s } = e, { boundary: c = "clippingAncestors", rootBoundary: l = "viewport", elementContext: u = "floating", altBoundary: d = !1, padding: f = 0 } = I(t, e), p = le(f), m = o[d ? u === "floating" ? "reference" : "floating" : u], h = G(await i.getClippingRect({
		element: await (i.isElement == null ? void 0 : i.isElement(m)) ?? !0 ? m : m.contextElement || await (i.getDocumentElement == null ? void 0 : i.getDocumentElement(o.floating)),
		boundary: c,
		rootBoundary: l,
		strategy: s
	})), g = u === "floating" ? {
		x: n,
		y: r,
		width: a.floating.width,
		height: a.floating.height
	} : a.reference, _ = await (i.getOffsetParent == null ? void 0 : i.getOffsetParent(o.floating)), v = await (i.isElement == null ? void 0 : i.isElement(_)) && await (i.getScale == null ? void 0 : i.getScale(_)) || {
		x: 1,
		y: 1
	}, y = G(i.convertOffsetParentRelativeRectToViewportRelativeRect ? await i.convertOffsetParentRelativeRectToViewportRelativeRect({
		elements: o,
		rect: g,
		offsetParent: _,
		strategy: s
	}) : g);
	return {
		top: (h.top - y.top + p.top) / v.y,
		bottom: (y.bottom - h.bottom + p.bottom) / v.y,
		left: (h.left - y.left + p.left) / v.x,
		right: (y.right - h.right + p.right) / v.x
	};
}
var fe = 50, pe = async (e, t, n) => {
	let { placement: r = "bottom", strategy: i = "absolute", middleware: a = [], platform: o } = n, s = o.detectOverflow ? o : {
		...o,
		detectOverflow: de
	}, c = await (o.isRTL == null ? void 0 : o.isRTL(t)), l = await o.getElementRects({
		reference: e,
		floating: t,
		strategy: i
	}), { x: u, y: d } = ue(l, r, c), f = r, p = 0, m = {};
	for (let n = 0; n < a.length; n++) {
		let h = a[n];
		if (!h) continue;
		let { name: g, fn: _ } = h, { x: v, y, data: b, reset: x } = await _({
			x: u,
			y: d,
			initialPlacement: r,
			placement: f,
			strategy: i,
			middlewareData: m,
			rects: l,
			platform: s,
			elements: {
				reference: e,
				floating: t
			}
		});
		u = v ?? u, d = y ?? d, m[g] = {
			...m[g],
			...b
		}, x && p < fe && (p++, typeof x == "object" && (x.placement && (f = x.placement), x.rects && (l = x.rects === !0 ? await o.getElementRects({
			reference: e,
			floating: t,
			strategy: i
		}) : x.rects), {x: u, y: d} = ue(l, f, c)), n = -1);
	}
	return {
		x: u,
		y: d,
		placement: f,
		strategy: i,
		middlewareData: m
	};
}, me = (e) => ({
	name: "arrow",
	options: e,
	async fn(t) {
		let { x: n, y: r, placement: i, rects: a, platform: o, elements: s, middlewareData: c } = t, { element: l, padding: u = 0 } = I(e, t) || {};
		if (l == null) return {};
		let d = le(u), f = {
			x: n,
			y: r
		}, p = H(i), m = B(p), h = await o.getDimensions(l), g = p === "y", _ = g ? "top" : "left", v = g ? "bottom" : "right", y = g ? "clientHeight" : "clientWidth", b = a.reference[m] + a.reference[p] - f[p] - a.floating[m], x = f[p] - a.reference[p], S = await (o.getOffsetParent == null ? void 0 : o.getOffsetParent(l)), C = S ? S[y] : 0;
		C && await (o.isElement == null ? void 0 : o.isElement(S)) || (C = s.floating[y] || a.floating[m]);
		let w = b / 2 - x / 2, T = C / 2 - h[m] / 2 - 1, E = k(d[_], T), D = k(d[v], T), O = E, A = C - h[m] - D, j = C / 2 - h[m] / 2 + w, M = F(O, j, A), N = !c.arrow && R(i) != null && j !== M && a.reference[m] / 2 - (j < O ? E : D) - h[m] / 2 < 0, P = N ? j < O ? j - O : j - A : 0;
		return {
			[p]: f[p] + P,
			data: {
				[p]: M,
				centerOffset: j - M - P,
				...N && { alignmentOffset: P }
			},
			reset: N
		};
	}
}), he = function(e) {
	return e === void 0 && (e = {}), {
		name: "flip",
		options: e,
		async fn(t) {
			var n;
			let { placement: r, middlewareData: i, rects: a, initialPlacement: o, platform: s, elements: c } = t, { mainAxis: l = !0, crossAxis: u = !0, fallbackPlacements: d, fallbackStrategy: f = "bestFit", fallbackAxisSideDirection: p = "none", flipAlignment: m = !0, ...h } = I(e, t);
			if ((n = i.arrow) != null && n.alignmentOffset) return {};
			let g = L(r), _ = V(o), v = L(o) === o, y = await (s.isRTL == null ? void 0 : s.isRTL(c.floating)), b = d || (v || !m ? [W(o)] : te(o)), x = p !== "none";
			!d && x && b.push(...se(o, m, p, y));
			let S = [o, ...b], C = await s.detectOverflow(t, h), w = [], T = i.flip?.overflows || [];
			if (l && w.push(C[g]), u) {
				let e = ee(r, a, y);
				w.push(C[e[0]], C[e[1]]);
			}
			if (T = [...T, {
				placement: r,
				overflows: w
			}], !w.every((e) => e <= 0)) {
				let e = (i.flip?.index || 0) + 1, t = S[e];
				if (t && (u !== "alignment" || _ === V(t) || T.every((e) => V(e.placement) !== _ || e.overflows[0] > 0))) return {
					data: {
						index: e,
						overflows: T
					},
					reset: { placement: t }
				};
				let n = T.filter((e) => e.overflows[0] <= 0).sort((e, t) => e.overflows[1] - t.overflows[1])[0]?.placement;
				if (!n) switch (f) {
					case "bestFit": {
						let e = T.filter((e) => {
							if (x) {
								let t = V(e.placement);
								return t === _ || t === "y";
							}
							return !0;
						}).map((e) => [e.placement, e.overflows.filter((e) => e > 0).reduce((e, t) => e + t, 0)]).sort((e, t) => e[1] - t[1])[0]?.[0];
						e && (n = e);
						break;
					}
					case "initialPlacement": n = o;
				}
				if (r !== n) return { reset: { placement: n } };
			}
			return {};
		}
	};
};
function ge(e, t) {
	return {
		top: e.top - t.height,
		right: e.right - t.width,
		bottom: e.bottom - t.height,
		left: e.left - t.width
	};
}
function _e(e) {
	return O.some((t) => e[t] >= 0);
}
var ve = function(e) {
	return e === void 0 && (e = {}), {
		name: "hide",
		options: e,
		async fn(t) {
			let { rects: n, platform: r } = t, { strategy: i = "referenceHidden", ...a } = I(e, t);
			switch (i) {
				case "referenceHidden": {
					let e = ge(await r.detectOverflow(t, {
						...a,
						elementContext: "reference"
					}), n.reference);
					return { data: {
						referenceHiddenOffsets: e,
						referenceHidden: _e(e)
					} };
				}
				case "escaped": {
					let e = ge(await r.detectOverflow(t, {
						...a,
						altBoundary: !0
					}), n.floating);
					return { data: {
						escapedOffsets: e,
						escaped: _e(e)
					} };
				}
				default: return {};
			}
		}
	};
}, ye = /*#__PURE__*/ new Set(["left", "top"]);
async function be(e, t) {
	let { placement: n, platform: r, elements: i } = e, a = await (r.isRTL == null ? void 0 : r.isRTL(i.floating)), o = L(n), s = R(n), c = V(n) === "y", l = ye.has(o) ? -1 : 1, u = a && c ? -1 : 1, d = I(t, e), { mainAxis: f, crossAxis: p, alignmentAxis: m } = typeof d == "number" ? {
		mainAxis: d,
		crossAxis: 0,
		alignmentAxis: null
	} : {
		mainAxis: d.mainAxis || 0,
		crossAxis: d.crossAxis || 0,
		alignmentAxis: d.alignmentAxis
	};
	return s && typeof m == "number" && (p = s === "end" ? m * -1 : m), c ? {
		x: p * u,
		y: f * l
	} : {
		x: f * l,
		y: p * u
	};
}
var xe = function(e) {
	return e === void 0 && (e = 0), {
		name: "offset",
		options: e,
		async fn(t) {
			var n;
			let { x: r, y: i, placement: a, middlewareData: o } = t, s = await be(t, e);
			return a === o.offset?.placement && (n = o.arrow) != null && n.alignmentOffset ? {} : {
				x: r + s.x,
				y: i + s.y,
				data: {
					...s,
					placement: a
				}
			};
		}
	};
}, Se = function(e) {
	return e === void 0 && (e = {}), {
		name: "shift",
		options: e,
		async fn(t) {
			let { x: n, y: r, placement: i, platform: a } = t, { mainAxis: o = !0, crossAxis: s = !1, limiter: c = { fn: (e) => {
				let { x: t, y: n } = e;
				return {
					x: t,
					y: n
				};
			} }, ...l } = I(e, t), u = {
				x: n,
				y: r
			}, d = await a.detectOverflow(t, l), f = V(L(i)), p = z(f), m = u[p], h = u[f];
			if (o) {
				let e = p === "y" ? "top" : "left", t = p === "y" ? "bottom" : "right", n = m + d[e], r = m - d[t];
				m = F(n, m, r);
			}
			if (s) {
				let e = f === "y" ? "top" : "left", t = f === "y" ? "bottom" : "right", n = h + d[e], r = h - d[t];
				h = F(n, h, r);
			}
			let g = c.fn({
				...t,
				[p]: m,
				[f]: h
			});
			return {
				...g,
				data: {
					x: g.x - n,
					y: g.y - r,
					enabled: {
						[p]: o,
						[f]: s
					}
				}
			};
		}
	};
}, Ce = function(e) {
	return e === void 0 && (e = {}), {
		options: e,
		fn(t) {
			let { x: n, y: r, placement: i, rects: a, middlewareData: o } = t, { offset: s = 0, mainAxis: c = !0, crossAxis: l = !0 } = I(e, t), u = {
				x: n,
				y: r
			}, d = V(i), f = z(d), p = u[f], m = u[d], h = I(s, t), g = typeof h == "number" ? {
				mainAxis: h,
				crossAxis: 0
			} : {
				mainAxis: 0,
				crossAxis: 0,
				...h
			};
			if (c) {
				let e = f === "y" ? "height" : "width", t = a.reference[f] - a.floating[e] + g.mainAxis, n = a.reference[f] + a.reference[e] - g.mainAxis;
				p < t ? p = t : p > n && (p = n);
			}
			if (l) {
				let e = f === "y" ? "width" : "height", t = ye.has(L(i)), n = a.reference[d] - a.floating[e] + (t && o.offset?.[d] || 0) + (t ? 0 : g.crossAxis), r = a.reference[d] + a.reference[e] + (t ? 0 : o.offset?.[d] || 0) - (t ? g.crossAxis : 0);
				m < n ? m = n : m > r && (m = r);
			}
			return {
				[f]: p,
				[d]: m
			};
		}
	};
}, we = function(e) {
	return e === void 0 && (e = {}), {
		name: "size",
		options: e,
		async fn(t) {
			var n, r;
			let { placement: i, rects: a, platform: o, elements: s } = t, { apply: c = () => {}, ...l } = I(e, t), u = await o.detectOverflow(t, l), d = L(i), f = R(i), p = V(i) === "y", { width: m, height: h } = a.floating, g, _;
			d === "top" || d === "bottom" ? (g = d, _ = f === (await (o.isRTL == null ? void 0 : o.isRTL(s.floating)) ? "start" : "end") ? "left" : "right") : (_ = d, g = f === "end" ? "top" : "bottom");
			let v = h - u.top - u.bottom, y = m - u.left - u.right, b = k(h - u[g], v), x = k(m - u[_], y), S = !t.middlewareData.shift, C = b, w = x;
			if ((n = t.middlewareData.shift) != null && n.enabled.x && (w = y), (r = t.middlewareData.shift) != null && r.enabled.y && (C = v), S && !f) {
				let e = A(u.left, 0), t = A(u.right, 0), n = A(u.top, 0), r = A(u.bottom, 0);
				p ? w = m - 2 * (e !== 0 || t !== 0 ? e + t : A(u.left, u.right)) : C = h - 2 * (n !== 0 || r !== 0 ? n + r : A(u.top, u.bottom));
			}
			await c({
				...t,
				availableWidth: w,
				availableHeight: C
			});
			let T = await o.getDimensions(s.floating);
			return m !== T.width || h !== T.height ? { reset: { rects: !0 } } : {};
		}
	};
};
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@floating-ui/dom/1.7.6/3c31067665f9052c9761c253844668811ad258c48969e1367af619e17b7587d9/node_modules/@floating-ui/dom/dist/floating-ui.dom.mjs
function Te(e) {
	let t = S(e), n = parseFloat(t.width) || 0, r = parseFloat(t.height) || 0, i = l(e), a = i ? e.offsetWidth : n, o = i ? e.offsetHeight : r, s = j(n) !== a || j(r) !== o;
	return s && (n = a, r = o), {
		width: n,
		height: r,
		$: s
	};
}
function K(e) {
	return c(e) ? e : e.contextElement;
}
function q(e) {
	let t = K(e);
	if (!l(t)) return N(1);
	let n = t.getBoundingClientRect(), { width: r, height: i, $: a } = Te(t), o = (a ? j(n.width) : n.width) / r, s = (a ? j(n.height) : n.height) / i;
	return o && Number.isFinite(o) || (o = 1), s && Number.isFinite(s) || (s = 1), {
		x: o,
		y: s
	};
}
var Ee = /*#__PURE__*/ N(0);
function De(e) {
	let t = a(e);
	return !b() || !t.visualViewport ? Ee : {
		x: t.visualViewport.offsetLeft,
		y: t.visualViewport.offsetTop
	};
}
function Oe(e, t, n) {
	return t === void 0 && (t = !1), !n || t && n !== a(e) ? !1 : t;
}
function J(e, t, n, r) {
	t === void 0 && (t = !1), n === void 0 && (n = !1);
	let i = e.getBoundingClientRect(), o = K(e), s = N(1);
	t && (r ? c(r) && (s = q(r)) : s = q(e));
	let l = Oe(o, n, r) ? De(o) : N(0), u = (i.left + l.x) / s.x, d = (i.top + l.y) / s.y, f = i.width / s.x, p = i.height / s.y;
	if (o) {
		let e = a(o), t = r && c(r) ? a(r) : r, n = e, i = D(n);
		for (; i && r && t !== n;) {
			let e = q(i), t = i.getBoundingClientRect(), r = S(i), o = t.left + (i.clientLeft + parseFloat(r.paddingLeft)) * e.x, s = t.top + (i.clientTop + parseFloat(r.paddingTop)) * e.y;
			u *= e.x, d *= e.y, f *= e.x, p *= e.y, u += o, d += s, n = a(i), i = D(n);
		}
	}
	return G({
		width: f,
		height: p,
		x: u,
		y: d
	});
}
function Y(e, t) {
	let n = C(e).scrollLeft;
	return t ? t.left + n : J(o(e)).left + n;
}
function ke(e, t) {
	let n = e.getBoundingClientRect();
	return {
		x: n.left + t.scrollLeft - Y(e, n),
		y: n.top + t.scrollTop
	};
}
function Ae(e) {
	let { elements: t, rect: n, offsetParent: r, strategy: a } = e, s = a === "fixed", c = o(r), u = t ? p(t.floating) : !1;
	if (r === c || u && s) return n;
	let f = {
		scrollLeft: 0,
		scrollTop: 0
	}, m = N(1), h = N(0), g = l(r);
	if ((g || !g && !s) && ((i(r) !== "body" || d(c)) && (f = C(r)), g)) {
		let e = J(r);
		m = q(r), h.x = e.x + r.clientLeft, h.y = e.y + r.clientTop;
	}
	let _ = c && !g && !s ? ke(c, f) : N(0);
	return {
		width: n.width * m.x,
		height: n.height * m.y,
		x: n.x * m.x - f.scrollLeft * m.x + h.x + _.x,
		y: n.y * m.y - f.scrollTop * m.y + h.y + _.y
	};
}
function je(e) {
	return Array.from(e.getClientRects());
}
function Me(e) {
	let t = o(e), n = C(e), r = e.ownerDocument.body, i = A(t.scrollWidth, t.clientWidth, r.scrollWidth, r.clientWidth), a = A(t.scrollHeight, t.clientHeight, r.scrollHeight, r.clientHeight), s = -n.scrollLeft + Y(e), c = -n.scrollTop;
	return S(r).direction === "rtl" && (s += A(t.clientWidth, r.clientWidth) - i), {
		width: i,
		height: a,
		x: s,
		y: c
	};
}
var Ne = 25;
function Pe(e, t) {
	let n = a(e), r = o(e), i = n.visualViewport, s = r.clientWidth, c = r.clientHeight, l = 0, u = 0;
	if (i) {
		s = i.width, c = i.height;
		let e = b();
		(!e || e && t === "fixed") && (l = i.offsetLeft, u = i.offsetTop);
	}
	let d = Y(r);
	if (d <= 0) {
		let e = r.ownerDocument, t = e.body, n = getComputedStyle(t), i = e.compatMode === "CSS1Compat" && parseFloat(n.marginLeft) + parseFloat(n.marginRight) || 0, a = Math.abs(r.clientWidth - t.clientWidth - i);
		a <= Ne && (s -= a);
	} else d <= Ne && (s += d);
	return {
		width: s,
		height: c,
		x: l,
		y: u
	};
}
function Fe(e, t) {
	let n = J(e, !0, t === "fixed"), r = n.top + e.clientTop, i = n.left + e.clientLeft, a = l(e) ? q(e) : N(1);
	return {
		width: e.clientWidth * a.x,
		height: e.clientHeight * a.y,
		x: i * a.x,
		y: r * a.y
	};
}
function Ie(e, t, n) {
	let r;
	if (t === "viewport") r = Pe(e, n);
	else if (t === "document") r = Me(o(e));
	else if (c(t)) r = Fe(t, n);
	else {
		let n = De(e);
		r = {
			x: t.x - n.x,
			y: t.y - n.y,
			width: t.width,
			height: t.height
		};
	}
	return G(r);
}
function Le(e, t) {
	let n = w(e);
	return n === t || !c(n) || x(n) ? !1 : S(n).position === "fixed" || Le(n, t);
}
function Re(e, t) {
	let n = t.get(e);
	if (n) return n;
	let r = E(e, [], !1).filter((e) => c(e) && i(e) !== "body"), a = null, o = S(e).position === "fixed", s = o ? w(e) : e;
	for (; c(s) && !x(s);) {
		let t = S(s), n = v(s);
		!n && t.position === "fixed" && (a = null), (o ? !n && !a : !n && t.position === "static" && a && (a.position === "absolute" || a.position === "fixed") || d(s) && !n && Le(e, s)) ? r = r.filter((e) => e !== s) : a = t, s = w(s);
	}
	return t.set(e, r), r;
}
function ze(e) {
	let { element: t, boundary: n, rootBoundary: r, strategy: i } = e, a = [...n === "clippingAncestors" ? p(t) ? [] : Re(t, this._c) : [].concat(n), r], o = Ie(t, a[0], i), s = o.top, c = o.right, l = o.bottom, u = o.left;
	for (let e = 1; e < a.length; e++) {
		let n = Ie(t, a[e], i);
		s = A(n.top, s), c = k(n.right, c), l = k(n.bottom, l), u = A(n.left, u);
	}
	return {
		width: c - u,
		height: l - s,
		x: u,
		y: s
	};
}
function Be(e) {
	let { width: t, height: n } = Te(e);
	return {
		width: t,
		height: n
	};
}
function Ve(e, t, n) {
	let r = l(t), a = o(t), s = n === "fixed", c = J(e, !0, s, t), u = {
		scrollLeft: 0,
		scrollTop: 0
	}, f = N(0);
	function p() {
		f.x = Y(a);
	}
	if (r || !r && !s) {
		if ((i(t) !== "body" || d(a)) && (u = C(t)), r) {
			let e = J(t, !0, s, t);
			f.x = e.x + t.clientLeft, f.y = e.y + t.clientTop;
		} else a && p();
	}
	s && !r && a && p();
	let m = a && !r && !s ? ke(a, u) : N(0);
	return {
		x: c.left + u.scrollLeft - f.x - m.x,
		y: c.top + u.scrollTop - f.y - m.y,
		width: c.width,
		height: c.height
	};
}
function X(e) {
	return S(e).position === "static";
}
function He(e, t) {
	if (!l(e) || S(e).position === "fixed") return null;
	if (t) return t(e);
	let n = e.offsetParent;
	return o(e) === n && (n = n.ownerDocument.body), n;
}
function Ue(e, t) {
	let n = a(e);
	if (p(e)) return n;
	if (!l(e)) {
		let t = w(e);
		for (; t && !x(t);) {
			if (c(t) && !X(t)) return t;
			t = w(t);
		}
		return n;
	}
	let r = He(e, t);
	for (; r && f(r) && X(r);) r = He(r, t);
	return r && x(r) && X(r) && !v(r) ? n : r || y(e) || n;
}
var We = async function(e) {
	let t = this.getOffsetParent || Ue, n = this.getDimensions, r = await n(e.floating);
	return {
		reference: Ve(e.reference, await t(e.floating), e.strategy),
		floating: {
			x: 0,
			y: 0,
			width: r.width,
			height: r.height
		}
	};
};
function Ge(e) {
	return S(e).direction === "rtl";
}
var Ke = {
	convertOffsetParentRelativeRectToViewportRelativeRect: Ae,
	getDocumentElement: o,
	getClippingRect: ze,
	getOffsetParent: Ue,
	getElementRects: We,
	getClientRects: je,
	getDimensions: Be,
	getScale: q,
	isElement: c,
	isRTL: Ge
};
function qe(e, t) {
	return e.x === t.x && e.y === t.y && e.width === t.width && e.height === t.height;
}
function Je(e, t) {
	let n = null, r, i = o(e);
	function a() {
		var e;
		clearTimeout(r), (e = n) == null || e.disconnect(), n = null;
	}
	function s(o, c) {
		o === void 0 && (o = !1), c === void 0 && (c = 1), a();
		let l = e.getBoundingClientRect(), { left: u, top: d, width: f, height: p } = l;
		if (o || t(), !f || !p) return;
		let m = M(d), h = M(i.clientWidth - (u + f)), g = M(i.clientHeight - (d + p)), _ = M(u), v = {
			rootMargin: -m + "px " + -h + "px " + -g + "px " + -_ + "px",
			threshold: A(0, k(1, c)) || 1
		}, y = !0;
		function b(t) {
			let n = t[0].intersectionRatio;
			if (n !== c) {
				if (!y) return s();
				n ? s(!1, n) : r = setTimeout(() => {
					s(!1, 1e-7);
				}, 1e3);
			}
			n === 1 && !qe(l, e.getBoundingClientRect()) && s(), y = !1;
		}
		try {
			n = new IntersectionObserver(b, {
				...v,
				root: i.ownerDocument
			});
		} catch {
			n = new IntersectionObserver(b, v);
		}
		n.observe(e);
	}
	return s(!0), a;
}
function Ye(e, t, n, r) {
	r === void 0 && (r = {});
	let { ancestorScroll: i = !0, ancestorResize: a = !0, elementResize: o = typeof ResizeObserver == "function", layoutShift: s = typeof IntersectionObserver == "function", animationFrame: c = !1 } = r, l = K(e), u = i || a ? [...l ? E(l) : [], ...t ? E(t) : []] : [];
	u.forEach((e) => {
		i && e.addEventListener("scroll", n, { passive: !0 }), a && e.addEventListener("resize", n);
	});
	let d = l && s ? Je(l, n) : null, f = -1, p = null;
	o && (p = new ResizeObserver((e) => {
		let [r] = e;
		r && r.target === l && p && t && (p.unobserve(t), cancelAnimationFrame(f), f = requestAnimationFrame(() => {
			var e;
			(e = p) == null || e.observe(t);
		})), n();
	}), l && !c && p.observe(l), t && p.observe(t));
	let m, h = c ? J(e) : null;
	c && g();
	function g() {
		let t = J(e);
		h && !qe(h, t) && n(), h = t, m = requestAnimationFrame(g);
	}
	return n(), () => {
		var e;
		u.forEach((e) => {
			i && e.removeEventListener("scroll", n), a && e.removeEventListener("resize", n);
		}), d?.(), (e = p) == null || e.disconnect(), p = null, c && cancelAnimationFrame(m);
	};
}
var Xe = xe, Ze = Se, Qe = he, $e = we, et = ve, tt = me, nt = Ce, rt = (e, t, n) => {
	let r = /* @__PURE__ */ new Map(), i = {
		platform: Ke,
		...n
	}, a = {
		...i.platform,
		_c: r
	};
	return pe(e, t, {
		...i,
		platform: a
	});
}, Z = /* @__PURE__ */ e(t(), 1), it = /* @__PURE__ */ e(n(), 1), Q = typeof document < "u" ? Z.useLayoutEffect : function() {};
function $(e, t) {
	if (e === t) return !0;
	if (typeof e != typeof t) return !1;
	if (typeof e == "function" && e.toString() === t.toString()) return !0;
	let n, r, i;
	if (e && t && typeof e == "object") {
		if (Array.isArray(e)) {
			if (n = e.length, n !== t.length) return !1;
			for (r = n; r-- !== 0;) if (!$(e[r], t[r])) return !1;
			return !0;
		}
		if (i = Object.keys(e), n = i.length, n !== Object.keys(t).length) return !1;
		for (r = n; r-- !== 0;) if (!{}.hasOwnProperty.call(t, i[r])) return !1;
		for (r = n; r-- !== 0;) {
			let n = i[r];
			if (!(n === "_owner" && e.$$typeof) && !$(e[n], t[n])) return !1;
		}
		return !0;
	}
	return e !== e && t !== t;
}
function at(e) {
	return typeof window > "u" ? 1 : (e.ownerDocument.defaultView || window).devicePixelRatio || 1;
}
function ot(e, t) {
	let n = at(e);
	return Math.round(t * n) / n;
}
function st(e) {
	let t = Z.useRef(e);
	return Q(() => {
		t.current = e;
	}), t;
}
function ct(e) {
	e === void 0 && (e = {});
	let { placement: t = "bottom", strategy: n = "absolute", middleware: r = [], platform: i, elements: { reference: a, floating: o } = {}, transform: s = !0, whileElementsMounted: c, open: l } = e, [u, d] = Z.useState({
		x: 0,
		y: 0,
		strategy: n,
		placement: t,
		middlewareData: {},
		isPositioned: !1
	}), [f, p] = Z.useState(r);
	$(f, r) || p(r);
	let [m, h] = Z.useState(null), [g, _] = Z.useState(null), v = Z.useCallback((e) => {
		e !== S.current && (S.current = e, h(e));
	}, []), y = Z.useCallback((e) => {
		e !== C.current && (C.current = e, _(e));
	}, []), b = a || m, x = o || g, S = Z.useRef(null), C = Z.useRef(null), w = Z.useRef(u), T = c != null, E = st(c), D = st(i), O = st(l), k = Z.useCallback(() => {
		if (!S.current || !C.current) return;
		let e = {
			placement: t,
			strategy: n,
			middleware: f
		};
		D.current && (e.platform = D.current), rt(S.current, C.current, e).then((e) => {
			let t = {
				...e,
				isPositioned: O.current !== !1
			};
			A.current && !$(w.current, t) && (w.current = t, it.flushSync(() => {
				d(t);
			}));
		});
	}, [
		f,
		t,
		n,
		D,
		O
	]);
	Q(() => {
		l === !1 && w.current.isPositioned && (w.current.isPositioned = !1, d((e) => ({
			...e,
			isPositioned: !1
		})));
	}, [l]);
	let A = Z.useRef(!1);
	Q(() => (A.current = !0, () => {
		A.current = !1;
	}), []), Q(() => {
		if (b && (S.current = b), x && (C.current = x), b && x) {
			if (E.current) return E.current(b, x, k);
			k();
		}
	}, [
		b,
		x,
		k,
		E,
		T
	]);
	let j = Z.useMemo(() => ({
		reference: S,
		floating: C,
		setReference: v,
		setFloating: y
	}), [v, y]), M = Z.useMemo(() => ({
		reference: b,
		floating: x
	}), [b, x]), N = Z.useMemo(() => {
		let e = {
			position: n,
			left: 0,
			top: 0
		};
		if (!M.floating) return e;
		let t = ot(M.floating, u.x), r = ot(M.floating, u.y);
		return s ? {
			...e,
			transform: "translate(" + t + "px, " + r + "px)",
			...at(M.floating) >= 1.5 && { willChange: "transform" }
		} : {
			position: n,
			left: t,
			top: r
		};
	}, [
		n,
		s,
		M.floating,
		u.x,
		u.y
	]);
	return Z.useMemo(() => ({
		...u,
		update: k,
		refs: j,
		elements: M,
		floatingStyles: N
	}), [
		u,
		k,
		j,
		M,
		N
	]);
}
var lt = (e) => {
	function t(e) {
		return {}.hasOwnProperty.call(e, "current");
	}
	return {
		name: "arrow",
		options: e,
		fn(n) {
			let { element: r, padding: i } = typeof e == "function" ? e(n) : e;
			return r && t(r) ? r.current == null ? {} : tt({
				element: r.current,
				padding: i
			}).fn(n) : r ? tt({
				element: r,
				padding: i
			}).fn(n) : {};
		}
	};
}, ut = (e, t) => {
	let n = Xe(e);
	return {
		name: n.name,
		fn: n.fn,
		options: [e, t]
	};
}, dt = (e, t) => {
	let n = Ze(e);
	return {
		name: n.name,
		fn: n.fn,
		options: [e, t]
	};
}, ft = (e, t) => ({
	fn: nt(e).fn,
	options: [e, t]
}), pt = (e, t) => {
	let n = Qe(e);
	return {
		name: n.name,
		fn: n.fn,
		options: [e, t]
	};
}, mt = (e, t) => {
	let n = $e(e);
	return {
		name: n.name,
		fn: n.fn,
		options: [e, t]
	};
}, ht = (e, t) => {
	let n = et(e);
	return {
		name: n.name,
		fn: n.fn,
		options: [e, t]
	};
}, gt = (e, t) => {
	let n = lt(e);
	return {
		name: n.name,
		fn: n.fn,
		options: [e, t]
	};
};
export { x as _, ut as a, b, ct as c, S as d, i as f, l as g, c as h, ft as i, Ye as l, w as m, pt as n, dt as o, E as p, ht as r, mt as s, gt as t, M as u, s as v, u as y };
