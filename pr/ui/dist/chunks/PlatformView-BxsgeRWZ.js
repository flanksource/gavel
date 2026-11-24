import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-a1ZmknZu.js";
import { t as r } from "./modern-wtRp5Z-b.js";
import { t as i } from "./clsx-BULO-GtV.js";
import { t as a } from "./button-Dbm88adW.js";
import { t as o } from "./Badge-BcEdZ4Rn.js";
import { $ as s, A as c, B as l, C as u, Cn as d, Ct as f, D as p, Dn as m, E as h, En as g, Et as _, F as v, G as ee, H as y, I as b, It as x, J as S, K as C, L as w, M as te, N as T, O as E, P as D, Q as ne, R as re, S as O, St as k, T as ie, U as ae, V as oe, W as se, X as A, Y as ce, Z as le, _ as ue, _t as de, a as j, at as fe, b as pe, bt as me, c as M, ct as he, d as ge, dt as _e, et as ve, f as ye, ft as N, g as be, gt as P, h as xe, ht as Se, i as Ce, it as we, j as Te, k as Ee, l as De, lt as Oe, m as ke, mt as Ae, nt as je, o as Me, ot as Ne, p as Pe, pt as Fe, q as Ie, r as Le, rt as Re, s as ze, st as Be, t as Ve, tt as He, u as Ue, ut as F, v as We, vt as I, w as Ge, wt as Ke, x as qe, xt as Je, y as Ye, yt as Xe, z as Ze, zn as L } from "./MeasuredUsagePanel-BdzUVnW5.js";
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/tiny-invariant/1.3.3/02ee75c29ce1796a59f3e3b45712899cff37750c6762c5ff2cce3192e6b76df0/node_modules/tiny-invariant/dist/esm/tiny-invariant.js
var Qe = "Invariant failed";
function $e(e, t) {
	if (!e) throw Error(Qe);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/recharts/3.8.1/34931137a9a572d88f21e6b00cc4044378ec5ec46ce2a08bf5baa19c6db21255/node_modules/recharts/es6/component/Cell.js
var et = (e) => null;
et.displayName = "Cell";
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/recharts/3.8.1/34931137a9a572d88f21e6b00cc4044378ec5ec46ce2a08bf5baa19c6db21255/node_modules/recharts/es6/context/tooltipContext.js
var tt = (e, t, n) => {
	var r = _e();
	return (i, a) => (o) => {
		e?.(i, a, o), r(te({
			activeIndex: String(a),
			activeDataKey: t,
			activeCoordinate: i.tooltipPosition,
			activeGraphicalItemId: n
		}));
	};
}, nt = (e) => {
	var t = _e();
	return (n, r) => (i) => {
		e?.(n, r, i), t(c());
	};
}, rt = (e, t, n) => {
	var r = _e();
	return (i, a) => (o) => {
		e?.(i, a, o), r(Te({
			activeIndex: String(a),
			activeDataKey: t,
			activeCoordinate: i.tooltipPosition,
			activeGraphicalItemId: n
		}));
	};
}, it = (e, t, n) => {
	var r = n ?? e;
	if (!I(r)) return Se(r, t, 0);
}, at = (e, t, n) => {
	var r = {}, i = e.filter(re), a = e.filter((e) => e.stackId == null), o = i.reduce((e, t) => {
		var n = e[t.stackId];
		return n ??= [], n.push(t), e[t.stackId] = n, e;
	}, r), s = Object.entries(o).map((e) => {
		var [r, i] = e;
		return {
			stackId: r,
			dataKeys: i.map((e) => e.dataKey),
			barSize: it(t, n, i[0]?.barSize)
		};
	}), c = a.map((e) => ({
		stackId: void 0,
		dataKeys: [e.dataKey].filter((e) => e != null),
		barSize: it(t, n, e.barSize)
	}));
	return [...s, ...c];
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/recharts/3.8.1/34931137a9a572d88f21e6b00cc4044378ec5ec46ce2a08bf5baa19c6db21255/node_modules/recharts/es6/state/selectors/combiners/combineAllBarPositions.js
function ot(e, t) {
	var n = Object.keys(e);
	if (Object.getOwnPropertySymbols) {
		var r = Object.getOwnPropertySymbols(e);
		t && (r = r.filter(function(t) {
			return Object.getOwnPropertyDescriptor(e, t).enumerable;
		})), n.push.apply(n, r);
	}
	return n;
}
function R(e) {
	for (var t = 1; t < arguments.length; t++) {
		var n = arguments[t] == null ? {} : arguments[t];
		t % 2 ? ot(Object(n), !0).forEach(function(t) {
			st(e, t, n[t]);
		}) : Object.getOwnPropertyDescriptors ? Object.defineProperties(e, Object.getOwnPropertyDescriptors(n)) : ot(Object(n)).forEach(function(t) {
			Object.defineProperty(e, t, Object.getOwnPropertyDescriptor(n, t));
		});
	}
	return e;
}
function st(e, t, n) {
	return (t = ct(t)) in e ? Object.defineProperty(e, t, {
		value: n,
		enumerable: !0,
		configurable: !0,
		writable: !0
	}) : e[t] = n, e;
}
function ct(e) {
	var t = lt(e, "string");
	return typeof t == "symbol" ? t : t + "";
}
function lt(e, t) {
	if (typeof e != "object" || !e) return e;
	var n = e[Symbol.toPrimitive];
	if (n !== void 0) {
		var r = n.call(e, t || "default");
		if (typeof r != "object") return r;
		throw TypeError("@@toPrimitive must return a primitive value.");
	}
	return (t === "string" ? String : Number)(e);
}
function ut(e, t, n, r, i) {
	var a = r.length;
	if (!(a < 1)) {
		var o = Se(e, n, 0, !0), s, c = [];
		if (Oe(r[0]?.barSize)) {
			var l = !1, u = n / a, d = r.reduce((e, t) => e + (t.barSize || 0), 0);
			d += (a - 1) * o, d >= n && (d -= (a - 1) * o, o = 0), d >= n && u > 0 && (l = !0, u *= .9, d = a * u);
			var f = {
				offset: ((n - d) / 2 | 0) - o,
				size: 0
			};
			s = r.reduce((e, t) => {
				var n = {
					stackId: t.stackId,
					dataKeys: t.dataKeys,
					position: {
						offset: f.offset + f.size + o,
						size: l ? u : t.barSize ?? 0
					}
				}, r = [...e, n];
				return f = n.position, r;
			}, c);
		} else {
			var p = Se(t, n, 0, !0);
			n - 2 * p - (a - 1) * o <= 0 && (o = 0);
			var m = (n - 2 * p - (a - 1) * o) / a;
			m > 1 && (m >>= 0);
			var h = Oe(i) ? Math.min(m, i) : m;
			s = r.reduce((e, t, n) => [...e, {
				stackId: t.stackId,
				dataKeys: t.dataKeys,
				position: {
					offset: p + (m + o) * n + (m - h) / 2,
					size: h
				}
			}], c);
		}
		return s;
	}
}
var dt = (e, t, n, r, i, a, o) => {
	var s = I(o) ? t : o, c = ut(n, r, i === a ? a : i, e, s);
	return i !== a && c != null && (c = c.map((e) => R(R({}, e), {}, { position: R(R({}, e.position), {}, { offset: e.position.offset - i / 2 }) }))), c;
}, ft = (e, t) => {
	var n = Ze(t);
	if (e && n != null && t != null) {
		var { stackId: r } = t;
		if (r != null) {
			var i = e[r];
			if (i) {
				var { stackedData: a } = i;
				if (a) return a.find((e) => e.key === n);
			}
		}
	}
}, pt = (e, t) => {
	if (e != null && t != null) {
		var n = e.find((e) => e.stackId === t.stackId && t.dataKey != null && e.dataKeys.includes(t.dataKey));
		if (n != null) return n.position;
	}
};
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/recharts/3.8.1/34931137a9a572d88f21e6b00cc4044378ec5ec46ce2a08bf5baa19c6db21255/node_modules/recharts/es6/zIndex/getZIndexFromUnknown.js
function mt(e, t) {
	return e && typeof e == "object" && "zIndex" in e && typeof e.zIndex == "number" && Oe(e.zIndex) ? e.zIndex : t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/recharts/3.8.1/34931137a9a572d88f21e6b00cc4044378ec5ec46ce2a08bf5baa19c6db21255/node_modules/recharts/es6/state/selectors/graphicalItemSelectors.js
function z(e, t) {
	return e.graphicalItems.cartesianItems.find((e) => e.id === t)?.xAxisId ?? 0;
}
function B(e, t) {
	return e.graphicalItems.cartesianItems.find((e) => e.id === t)?.yAxisId ?? 0;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/recharts/3.8.1/34931137a9a572d88f21e6b00cc4044378ec5ec46ce2a08bf5baa19c6db21255/node_modules/recharts/es6/util/BarUtils.js
var V = /* @__PURE__ */ e(t());
function ht() {
	return ht = Object.assign ? Object.assign.bind() : function(e) {
		for (var t = 1; t < arguments.length; t++) {
			var n = arguments[t];
			for (var r in n) ({}).hasOwnProperty.call(n, r) && (e[r] = n[r]);
		}
		return e;
	}, ht.apply(null, arguments);
}
function gt(e) {
	return /*#__PURE__*/ V.createElement(qe, ht({
		shapeType: "rectangle",
		activeClassName: "recharts-active-bar",
		inActiveClassName: "recharts-inactive-bar"
	}, e));
}
var _t = function(e) {
	var t = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : 0;
	return (n, r) => {
		if (Xe(e)) return e;
		var i = Xe(n) || I(n);
		return i ? e(n, r) : (!i && $e(!1, `minPointSize callback function received a value with type of ${typeof n}. Currently only numbers or null/undefined are supported.`), t);
	};
}, vt = (e, t, n) => n, H = F([w, (e, t) => t], (e, t) => e.filter((e) => e.type === "bar").find((e) => e.id === t)), yt = F([H], (e) => e?.maxBarSize), bt = (e, t, n, r) => r, xt = F([
	A,
	w,
	z,
	B,
	vt
], (e, t, n, r, i) => t.filter((t) => e === "horizontal" ? t.xAxisId === n : t.yAxisId === r).filter((e) => e.isPanorama === i).filter((e) => e.hide === !1).filter((e) => e.type === "bar")), St = (e, t, n) => {
	var r = A(e), i = z(e, t), a = B(e, t);
	if (i != null && a != null) return r === "horizontal" ? v(e, "yAxis", a, n) : v(e, "xAxis", i, n);
}, Ct = F([
	xt,
	ae,
	(e, t) => {
		var n = A(e), r = z(e, t), i = B(e, t);
		if (r != null && i != null) return n === "horizontal" ? D(e, "xAxis", r) : D(e, "yAxis", i);
	}
], at), wt = (e, t, n) => {
	var r = H(e, t);
	if (r == null) return 0;
	var i = z(e, t), a = B(e, t);
	if (i == null || a == null) return 0;
	var o = A(e), s = se(e), { maxBarSize: c } = r, l = I(c) ? s : c, u, d;
	return o === "horizontal" ? (u = T(e, "xAxis", i, n), d = b(e, "xAxis", i, n)) : (u = T(e, "yAxis", a, n), d = b(e, "yAxis", a, n)), je(u, d, !0) ?? l ?? 0;
}, Tt = (e, t, n) => {
	var r = A(e), i = z(e, t), a = B(e, t);
	if (i != null && a != null) {
		var o, s;
		return r === "horizontal" ? (o = T(e, "xAxis", i, n), s = b(e, "xAxis", i, n)) : (o = T(e, "yAxis", a, n), s = b(e, "yAxis", a, n)), je(o, s);
	}
}, Et = F([
	Ct,
	se,
	y,
	oe,
	wt,
	Tt,
	yt
], dt), Dt = (e, t, n) => {
	var r = z(e, t);
	if (r != null) return T(e, "xAxis", r, n);
}, Ot = (e, t, n) => {
	var r = B(e, t);
	if (r != null) return T(e, "yAxis", r, n);
}, kt = (e, t, n) => {
	var r = z(e, t);
	if (r != null) return b(e, "xAxis", r, n);
}, At = (e, t, n) => {
	var r = B(e, t);
	if (r != null) return b(e, "yAxis", r, n);
}, jt = F([Et, H], pt), Mt = F([St, H], ft), Nt = F([
	He,
	ve,
	Dt,
	Ot,
	kt,
	At,
	jt,
	A,
	ee,
	Tt,
	Mt,
	H,
	bt
], (e, t, n, r, i, a, o, s, c, l, u, d, f) => {
	var { chartData: p, dataStartIndex: m, dataEndIndex: h } = c;
	if (!(d == null || o == null || t == null || s !== "horizontal" && s !== "vertical" || n == null || r == null || i == null || a == null || l == null)) {
		var { data: g } = d, _ = g != null && g.length > 0 ? g : p?.slice(m, h + 1);
		if (_ != null) return pn({
			layout: s,
			barSettings: d,
			pos: o,
			parentViewBox: t,
			bandSize: l,
			xAxis: n,
			yAxis: r,
			xAxisTicks: i,
			yAxisTicks: a,
			stackedData: u,
			displayedData: _,
			offset: e,
			cells: f,
			dataStartIndex: m
		});
	}
}), Pt = ["index"];
function Ft() {
	return Ft = Object.assign ? Object.assign.bind() : function(e) {
		for (var t = 1; t < arguments.length; t++) {
			var n = arguments[t];
			for (var r in n) ({}).hasOwnProperty.call(n, r) && (e[r] = n[r]);
		}
		return e;
	}, Ft.apply(null, arguments);
}
function It(e, t) {
	if (e == null) return {};
	var n, r, i = Lt(e, t);
	if (Object.getOwnPropertySymbols) {
		var a = Object.getOwnPropertySymbols(e);
		for (r = 0; r < a.length; r++) n = a[r], t.indexOf(n) === -1 && {}.propertyIsEnumerable.call(e, n) && (i[n] = e[n]);
	}
	return i;
}
function Lt(e, t) {
	if (e == null) return {};
	var n = {};
	for (var r in e) if ({}.hasOwnProperty.call(e, r)) {
		if (t.indexOf(r) !== -1) continue;
		n[r] = e[r];
	}
	return n;
}
var Rt = /*#__PURE__*/ (0, V.createContext)(void 0), zt = (e) => {
	var t = (0, V.useContext)(Rt);
	if (t != null) return t.stackId;
	if (e != null) return fe(e);
}, Bt = (e, t) => `recharts-bar-stack-clip-path-${e}-${t}`, Vt = (e) => {
	var t = (0, V.useContext)(Rt);
	if (t != null) {
		var { stackId: n } = t;
		return `url(#${Bt(n, e)})`;
	}
}, Ht = (e) => {
	var { index: t } = e, n = It(e, Pt), r = Vt(t);
	return /*#__PURE__*/ V.createElement(k, Ft({
		className: "recharts-bar-stack-layer",
		clipPath: r
	}, n));
}, Ut = [
	"onMouseEnter",
	"onMouseLeave",
	"onClick"
], Wt = [
	"value",
	"background",
	"tooltipPosition"
], Gt = ["id"], Kt = [
	"onMouseEnter",
	"onClick",
	"onMouseLeave"
];
function U() {
	return U = Object.assign ? Object.assign.bind() : function(e) {
		for (var t = 1; t < arguments.length; t++) {
			var n = arguments[t];
			for (var r in n) ({}).hasOwnProperty.call(n, r) && (e[r] = n[r]);
		}
		return e;
	}, U.apply(null, arguments);
}
function qt(e, t) {
	var n = Object.keys(e);
	if (Object.getOwnPropertySymbols) {
		var r = Object.getOwnPropertySymbols(e);
		t && (r = r.filter(function(t) {
			return Object.getOwnPropertyDescriptor(e, t).enumerable;
		})), n.push.apply(n, r);
	}
	return n;
}
function W(e) {
	for (var t = 1; t < arguments.length; t++) {
		var n = arguments[t] == null ? {} : arguments[t];
		t % 2 ? qt(Object(n), !0).forEach(function(t) {
			Jt(e, t, n[t]);
		}) : Object.getOwnPropertyDescriptors ? Object.defineProperties(e, Object.getOwnPropertyDescriptors(n)) : qt(Object(n)).forEach(function(t) {
			Object.defineProperty(e, t, Object.getOwnPropertyDescriptor(n, t));
		});
	}
	return e;
}
function Jt(e, t, n) {
	return (t = Yt(t)) in e ? Object.defineProperty(e, t, {
		value: n,
		enumerable: !0,
		configurable: !0,
		writable: !0
	}) : e[t] = n, e;
}
function Yt(e) {
	var t = Xt(e, "string");
	return typeof t == "symbol" ? t : t + "";
}
function Xt(e, t) {
	if (typeof e != "object" || !e) return e;
	var n = e[Symbol.toPrimitive];
	if (n !== void 0) {
		var r = n.call(e, t || "default");
		if (typeof r != "object") return r;
		throw TypeError("@@toPrimitive must return a primitive value.");
	}
	return (t === "string" ? String : Number)(e);
}
function G(e, t) {
	if (e == null) return {};
	var n, r, i = Zt(e, t);
	if (Object.getOwnPropertySymbols) {
		var a = Object.getOwnPropertySymbols(e);
		for (r = 0; r < a.length; r++) n = a[r], t.indexOf(n) === -1 && {}.propertyIsEnumerable.call(e, n) && (i[n] = e[n]);
	}
	return i;
}
function Zt(e, t) {
	if (e == null) return {};
	var n = {};
	for (var r in e) if ({}.hasOwnProperty.call(e, r)) {
		if (t.indexOf(r) !== -1) continue;
		n[r] = e[r];
	}
	return n;
}
var Qt = (e) => {
	var { dataKey: t, name: n, fill: r, legendType: i, hide: a } = e;
	return [{
		inactive: a,
		dataKey: t,
		type: i,
		color: r,
		value: Ne(n, t),
		payload: e
	}];
}, $t = /*#__PURE__*/ V.memo((e) => {
	var { dataKey: t, stroke: n, strokeWidth: r, fill: i, name: a, hide: o, unit: s, tooltipType: c, id: l } = e, u = {
		dataDefinedOnItem: void 0,
		getPosition: Je,
		settings: {
			stroke: n,
			strokeWidth: r,
			fill: i,
			dataKey: t,
			nameKey: void 0,
			name: Ne(a, t),
			hide: o,
			type: c,
			color: i,
			unit: s,
			graphicalItemId: l
		}
	};
	return /*#__PURE__*/ V.createElement(pe, { tooltipEntrySettings: u });
});
function en(e) {
	var t = N(Ee), { data: n, dataKey: r, background: i, allOtherBarProps: a } = e, { onMouseEnter: o, onMouseLeave: s, onClick: c } = a, u = G(a, Ut), d = tt(o, r, a.id), f = nt(s), m = rt(c, r, a.id);
	if (!i || n == null) return null;
	var h = Ke(i);
	return /*#__PURE__*/ V.createElement(p, { zIndex: mt(i, l.barBackground) }, n.map((e, n) => {
		var { value: a, background: o, tooltipPosition: s } = e, c = G(e, Wt);
		if (!o) return null;
		var l = d(e, n), p = f(e, n), g = m(e, n), _ = W(W(W(W(W({
			option: i,
			isActive: String(n) === t
		}, c), {}, { fill: "#eee" }, o), h), Ae(u, e, n)), {}, {
			onMouseEnter: l,
			onMouseLeave: p,
			onClick: g,
			dataKey: r,
			index: n,
			className: "recharts-bar-background-rectangle"
		});
		return /*#__PURE__*/ V.createElement(gt, U({ key: `background-bar-${n}` }, _));
	}));
}
function tn(e) {
	var { showLabels: t, children: n, rects: r } = e, i = r?.map((e) => {
		var t = {
			x: e.x,
			y: e.y,
			width: e.width,
			lowerWidth: e.width,
			upperWidth: e.width,
			height: e.height
		};
		return W(W({}, t), {}, {
			value: e.value,
			payload: e.payload,
			parentViewBox: e.parentViewBox,
			viewBox: t,
			fill: e.fill
		});
	});
	return /*#__PURE__*/ V.createElement(u, { value: t ? i : void 0 }, n);
}
function nn(e) {
	var { shape: t, activeBar: n, baseProps: r, entry: i, index: a, dataKey: o } = e, s = N(Ee), c = N(E), u = n && String(i.originalDataIndex) === s && (c == null || o === c), [d, f] = (0, V.useState)(!1), [m, h] = (0, V.useState)(!1);
	(0, V.useEffect)(() => {
		var e;
		return u ? (f(!0), e = requestAnimationFrame(() => {
			h(!0);
		})) : h(!1), () => {
			cancelAnimationFrame(e);
		};
	}, [u]);
	var g = (0, V.useCallback)(() => {
		u || f(!1);
	}, [u]), _ = u && m, v = u || d, ee = u ? n === !0 ? t : n : t, y = /*#__PURE__*/ V.createElement(gt, U({}, r, { name: String(r.name) }, i, {
		isActive: _,
		option: ee,
		index: a,
		dataKey: o,
		onTransitionEnd: g
	}));
	return v ? /*#__PURE__*/ V.createElement(p, { zIndex: l.activeBar }, /*#__PURE__*/ V.createElement(Ht, { index: i.originalDataIndex }, y)) : y;
}
function rn(e) {
	var { shape: t, baseProps: n, entry: r, index: i, dataKey: a } = e;
	return /*#__PURE__*/ V.createElement(gt, U({}, n, { name: String(n.name) }, r, {
		isActive: !1,
		option: t,
		index: i,
		dataKey: a
	}));
}
function an(e) {
	var { data: t, props: n } = e, r = f(n) ?? {}, { id: i } = r, a = G(r, Gt), { shape: o, dataKey: s, activeBar: c } = n, { onMouseEnter: l, onClick: u, onMouseLeave: d } = n, p = G(n, Kt), m = tt(l, s, i), h = nt(d), g = rt(u, s, i);
	return t ? /*#__PURE__*/ V.createElement(V.Fragment, null, t.map((e, t) => /*#__PURE__*/ V.createElement(Ht, U({
		index: e.originalDataIndex,
		key: `rectangle-${e?.x}-${e?.y}-${e?.value}-${t}`,
		className: "recharts-bar-rectangle"
	}, Ae(p, e, t), {
		onMouseEnter: m(e, t),
		onMouseLeave: h(e, t),
		onClick: g(e, t)
	}), c ? /*#__PURE__*/ V.createElement(nn, {
		shape: o,
		activeBar: c,
		baseProps: a,
		entry: e,
		index: t,
		dataKey: s
	}) : /*#__PURE__*/ V.createElement(rn, {
		shape: o,
		baseProps: a,
		entry: e,
		index: t,
		dataKey: s
	})))) : null;
}
function on(e) {
	var { props: t, previousRectanglesRef: n } = e, { data: r, layout: i, isAnimationActive: a, animationBegin: o, animationDuration: s, animationEasing: c, onAnimationEnd: l, onAnimationStart: u } = t, d = n.current, f = C(t, "recharts-bar-"), [p, m] = (0, V.useState)(!1), h = !p, g = (0, V.useCallback)(() => {
		typeof l == "function" && l(), m(!1);
	}, [l]), _ = (0, V.useCallback)(() => {
		typeof u == "function" && u(), m(!0);
	}, [u]);
	return /*#__PURE__*/ V.createElement(tn, {
		showLabels: h,
		rects: r
	}, /*#__PURE__*/ V.createElement(Ie, {
		animationId: f,
		begin: o,
		duration: s,
		isActive: a,
		easing: c,
		onAnimationEnd: g,
		onAnimationStart: _,
		key: f
	}, (e) => {
		var a = e === 1 ? r : r?.map((t, n) => {
			var r = d && d[n];
			if (r) return W(W({}, t), {}, {
				x: P(r.x, t.x, e),
				y: P(r.y, t.y, e),
				width: P(r.width, t.width, e),
				height: P(r.height, t.height, e)
			});
			if (i === "horizontal") {
				var a = P(0, t.height, e), o = P(t.stackedBarStart, t.y, e);
				return W(W({}, t), {}, {
					y: o,
					height: a
				});
			}
			var s = P(0, t.width, e), c = P(t.stackedBarStart, t.x, e);
			return W(W({}, t), {}, {
				width: s,
				x: c
			});
		});
		return e > 0 && (n.current = a ?? null), a == null ? null : /*#__PURE__*/ V.createElement(k, null, /*#__PURE__*/ V.createElement(an, {
			props: t,
			data: a
		}));
	}), /*#__PURE__*/ V.createElement(Ge, { label: t.label }), t.children);
}
function sn(e) {
	var t = (0, V.useRef)(null);
	return /*#__PURE__*/ V.createElement(on, {
		previousRectanglesRef: t,
		props: e
	});
}
var cn = 0, ln = (e, t) => {
	var n = Array.isArray(e.value) ? e.value[1] : e.value;
	return {
		x: e.x,
		y: e.y,
		value: n,
		errorVal: Be(e, t)
	};
}, un = class extends V.PureComponent {
	render() {
		var { hide: e, data: t, dataKey: n, className: r, xAxisId: a, yAxisId: o, needClip: s, background: c, id: l } = this.props;
		if (e || t == null) return null;
		var u = i("recharts-bar", r), d = l;
		return /*#__PURE__*/ V.createElement(k, {
			className: u,
			id: l
		}, s && /*#__PURE__*/ V.createElement("defs", null, /*#__PURE__*/ V.createElement(Pe, {
			clipPathId: d,
			xAxisId: a,
			yAxisId: o
		})), /*#__PURE__*/ V.createElement(k, {
			className: "recharts-bar-rectangles",
			clipPath: s ? `url(#clipPath-${d})` : void 0
		}, /*#__PURE__*/ V.createElement(en, {
			data: t,
			dataKey: n,
			background: c,
			allOtherBarProps: this.props
		}), /*#__PURE__*/ V.createElement(sn, this.props)));
	}
}, dn = {
	activeBar: !1,
	animationBegin: 0,
	animationDuration: 400,
	animationEasing: "ease",
	background: !1,
	hide: !1,
	isAnimationActive: "auto",
	label: !1,
	legendType: "rect",
	minPointSize: cn,
	xAxisId: 0,
	yAxisId: 0,
	zIndex: l.bar
};
function fn(e) {
	var { xAxisId: t, yAxisId: n, hide: r, legendType: i, minPointSize: a, activeBar: o, animationBegin: c, animationDuration: l, animationEasing: u, isAnimationActive: d } = e, { needClip: f } = ke(t, n), p = le(), m = s(), h = O(e.children, et), g = N((t) => Nt(t, e.id, m, h));
	if (p !== "vertical" && p !== "horizontal") return null;
	var _, v = g?.[0];
	return _ = v == null || v.height == null || v.width == null ? 0 : p === "vertical" ? v.height / 2 : v.width / 2, /*#__PURE__*/ V.createElement(xe, {
		xAxisId: t,
		yAxisId: n,
		data: g,
		dataPointFormatter: ln,
		errorBarOffset: _
	}, /*#__PURE__*/ V.createElement(un, U({}, e, {
		layout: p,
		needClip: f,
		data: g,
		xAxisId: t,
		yAxisId: n,
		hide: r,
		legendType: i,
		minPointSize: a,
		activeBar: o,
		animationBegin: c,
		animationDuration: l,
		animationEasing: u,
		isAnimationActive: d
	})));
}
function pn(e) {
	var { layout: t, barSettings: { dataKey: n, minPointSize: r, hasCustomShape: i }, pos: a, bandSize: o, xAxis: s, yAxis: c, xAxisTicks: l, yAxisTicks: u, stackedData: d, displayedData: f, offset: p, cells: m, parentViewBox: h, dataStartIndex: g } = e, _ = t === "horizontal" ? c : s, v = d ? _.scale.domain() : null, ee = Re({ numericAxis: _ }), y = _.scale.map(ee);
	return f.map((e, f) => {
		var _, b, x, S, C, w;
		if (d) {
			var te = d[f + g];
			if (te == null) return null;
			_ = he(te, v);
		} else _ = Be(e, n), Array.isArray(_) || (_ = [ee, _]);
		var T = _t(r, cn)(_[1], f);
		if (t === "horizontal") {
			var E = c.scale.map(_[0]), D = c.scale.map(_[1]);
			if (E == null || D == null) return null;
			b = we({
				axis: s,
				ticks: l,
				bandSize: o,
				offset: a.offset,
				entry: e,
				index: f
			}), x = D ?? E ?? void 0, S = a.size;
			var ne = E - D;
			if (C = de(ne) ? 0 : ne, w = {
				x: b,
				y: p.top,
				width: S,
				height: p.height
			}, Math.abs(T) > 0 && Math.abs(C) < Math.abs(T)) {
				var re = me(C || T) * (Math.abs(T) - Math.abs(C));
				x -= re, C += re;
			}
		} else {
			var O = s.scale.map(_[0]), k = s.scale.map(_[1]);
			if (O == null || k == null) return null;
			if (b = O, x = we({
				axis: c,
				ticks: u,
				bandSize: o,
				offset: a.offset,
				entry: e,
				index: f
			}), S = k - O, C = a.size, w = {
				x: p.left,
				y: x,
				width: p.width,
				height: C
			}, Math.abs(T) > 0 && Math.abs(S) < Math.abs(T)) {
				var ie = me(S || T) * (Math.abs(T) - Math.abs(S));
				S += ie;
			}
		}
		return b == null || x == null || S == null || C == null || !i && (S === 0 || C === 0) ? null : W(W({}, e), {}, {
			stackedBarStart: y,
			x: b,
			y: x,
			width: S,
			height: C,
			value: d ? _ : _[1],
			payload: e,
			background: w,
			tooltipPosition: {
				x: b + S / 2,
				y: x + C / 2
			},
			parentViewBox: h,
			originalDataIndex: f
		}, m && m[f] && m[f].props);
	}).filter(Boolean);
}
function mn(e) {
	var t = Fe(e, dn), n = zt(t.stackId), r = s();
	return /*#__PURE__*/ V.createElement(We, {
		id: t.id,
		type: "bar"
	}, (e) => /*#__PURE__*/ V.createElement(V.Fragment, null, /*#__PURE__*/ V.createElement(Ye, { legendPayload: Qt(t) }), /*#__PURE__*/ V.createElement($t, {
		dataKey: t.dataKey,
		stroke: t.stroke,
		strokeWidth: t.strokeWidth,
		fill: t.fill,
		name: t.name,
		hide: t.hide,
		unit: t.unit,
		tooltipType: t.tooltipType,
		id: e
	}), /*#__PURE__*/ V.createElement(ue, {
		type: "bar",
		id: e,
		data: void 0,
		xAxisId: t.xAxisId,
		yAxisId: t.yAxisId,
		zAxisId: 0,
		dataKey: t.dataKey,
		stackId: n,
		hide: t.hide,
		barSize: t.barSize,
		minPointSize: t.minPointSize,
		maxBarSize: t.maxBarSize,
		isPanorama: r,
		hasCustomShape: t.shape != null
	}), /*#__PURE__*/ V.createElement(p, { zIndex: t.zIndex }, /*#__PURE__*/ V.createElement(fn, U({}, t, { id: e })))));
}
var hn = /*#__PURE__*/ V.memo(mn, ce);
hn.displayName = "Bar";
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/recharts/3.8.1/34931137a9a572d88f21e6b00cc4044378ec5ec46ce2a08bf5baa19c6db21255/node_modules/recharts/es6/chart/BarChart.js
var gn = ["axis", "item"], _n = /*#__PURE__*/ (0, V.forwardRef)((e, t) => /*#__PURE__*/ V.createElement(Ue, {
	chartName: "BarChart",
	defaultTooltipEventType: "axis",
	validateTooltipEventTypes: gn,
	tooltipPayloadSearcher: h,
	categoricalChartProps: e,
	ref: t
})), K = r(), q = /* @__PURE__ */ e(n(), 1), vn = new Intl.NumberFormat("en-US", {
	style: "currency",
	currency: "USD"
});
function yn(e) {
	return Object.keys(e).sort().map((t) => `${t}=${e[t]}`).join(", ");
}
function bn(e, t) {
	let n = Math.round(e / t * 100);
	return e > t ? `${n}% · over budget` : `${n}%`;
}
function xn(e) {
	return e.flatMap((e) => {
		let t = e.windowStart ? `${e.window} since ${Me(e.windowStart.slice(0, 10))}` : e.window, n = vn.format(e.amountUSD);
		return e.windowError ? [{
			key: e.id,
			cells: [
				e.name,
				t,
				e.windowError,
				"—",
				n,
				"—"
			]
		}] : e.groups.length === 0 ? [{
			key: e.id,
			cells: [
				e.name,
				t,
				"No spend in this window",
				vn.format(0),
				n,
				bn(0, e.amountUSD)
			]
		}] : e.groups.map((r) => {
			let i = yn(r.values), a = `${e.id}/${i}`;
			return r.error ? {
				key: a,
				cells: [
					e.name,
					t,
					i,
					r.error,
					n,
					"—"
				]
			} : {
				key: a,
				cells: [
					e.name,
					t,
					i,
					vn.format(r.spendUSD),
					n,
					bn(r.spendUSD, e.amountUSD)
				]
			};
		});
	});
}
function Sn() {
	let e = (0, K.useQuery)({
		queryKey: m.platform("budgets"),
		queryFn: ({ signal: e }) => d({
			url: "/api/platform/budgets",
			signal: e,
			context: "load budgets"
		})
	}), t = e.data?.budgets;
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "flex flex-col gap-3",
		children: [
			/* @__PURE__ */ (0, q.jsx)(M, { children: "Budgets" }),
			e.error && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: e.error.message
			}),
			t && t.length === 0 && /* @__PURE__ */ (0, q.jsx)("p", {
				className: "text-xs text-muted-foreground",
				children: "No budget rules. Add one with captain budget-rule create, or on captain serve's Budgets page."
			}),
			t && t.length > 0 && /* @__PURE__ */ (0, q.jsxs)("table", {
				"aria-label": "Budgets",
				className: "w-full border-collapse text-left text-xs",
				children: [/* @__PURE__ */ (0, q.jsx)("thead", {
					className: "text-muted-foreground",
					children: /* @__PURE__ */ (0, q.jsx)("tr", { children: [
						"Rule",
						"Window",
						"Group",
						"Spent",
						"Budget",
						"Used"
					].map((e) => /* @__PURE__ */ (0, q.jsx)("th", {
						className: "border-b border-border px-2 py-1.5 font-medium",
						children: e
					}, e)) })
				}), /* @__PURE__ */ (0, q.jsx)("tbody", { children: xn(t).map((e) => /* @__PURE__ */ (0, q.jsx)("tr", {
					className: "border-b border-border",
					children: e.cells.map((e, t) => /* @__PURE__ */ (0, q.jsx)("td", {
						className: `px-2 py-1.5${t === 0 ? " font-medium" : ""}`,
						children: e
					}, t))
				}, e.key)) })]
			})
		]
	});
}
// src/components/platform/formatSeconds.ts
function Cn(e) {
	return e < 60 ? `${Math.round(e)}s` : e < 3600 ? `${Math.round(e / 60)}m` : `${(e / 3600).toFixed(1)}h`;
}
// src/components/platform/StackedDayBars.tsx
function wn({ title: e, data: t, series: n, unit: r }) {
	let i = Object.fromEntries(n.map((e) => [e.key, e.label])), a = (e, t) => Number(e[t] ?? 0);
	return /* @__PURE__ */ (0, q.jsxs)("figure", {
		className: "flex flex-col gap-1",
		children: [
			/* @__PURE__ */ (0, q.jsx)("figcaption", {
				className: "text-xs text-muted-foreground",
				children: e
			}),
			/* @__PURE__ */ (0, q.jsx)("div", {
				className: "h-48 text-muted-foreground",
				children: /* @__PURE__ */ (0, q.jsx)(ne, {
					width: "100%",
					height: "100%",
					initialDimension: {
						width: 480,
						height: 192
					},
					children: /* @__PURE__ */ (0, q.jsxs)(_n, {
						data: t,
						margin: {
							top: 4,
							right: 8,
							bottom: 0,
							left: 0
						},
						children: [
							/* @__PURE__ */ (0, q.jsx)(be, {
								vertical: !1,
								stroke: "currentColor",
								strokeOpacity: .15
							}),
							/* @__PURE__ */ (0, q.jsx)(ye, {
								dataKey: "day",
								tickFormatter: ze,
								tick: {
									fill: "currentColor",
									fontSize: 11
								},
								tickLine: !1,
								axisLine: !1
							}),
							/* @__PURE__ */ (0, q.jsx)(ge, {
								allowDecimals: !1,
								width: 36,
								tick: {
									fill: "currentColor",
									fontSize: 11
								},
								tickLine: !1,
								axisLine: !1
							}),
							/* @__PURE__ */ (0, q.jsx)(ie, {
								formatter: (e, t) => [`${Math.round(Number(e))} ${r}`, i[String(t)] ?? String(t)],
								cursor: { fillOpacity: .08 }
							}),
							/* @__PURE__ */ (0, q.jsx)(S, {
								formatter: (e) => /* @__PURE__ */ (0, q.jsx)("span", {
									className: "text-xs text-foreground",
									children: i[e] ?? e
								}),
								iconSize: 8
							}),
							n.map((e) => /* @__PURE__ */ (0, q.jsx)(hn, {
								dataKey: e.key,
								stackId: "series",
								fill: e.color,
								stroke: "var(--background)",
								strokeWidth: 1,
								maxBarSize: 24
							}, e.key))
						]
					})
				})
			}),
			/* @__PURE__ */ (0, q.jsxs)("details", {
				className: "text-xs",
				children: [/* @__PURE__ */ (0, q.jsx)("summary", {
					className: "cursor-pointer text-muted-foreground",
					children: "Show as a table"
				}), /* @__PURE__ */ (0, q.jsxs)("table", {
					"aria-label": e,
					className: "mt-1 w-full border-collapse text-left",
					children: [/* @__PURE__ */ (0, q.jsx)("thead", {
						className: "text-muted-foreground",
						children: /* @__PURE__ */ (0, q.jsx)("tr", { children: [
							"Day",
							...n.map((e) => e.label),
							"Total"
						].map((e) => /* @__PURE__ */ (0, q.jsx)("th", {
							className: "border-b border-border px-2 py-1 font-medium",
							children: e
						}, e)) })
					}), /* @__PURE__ */ (0, q.jsx)("tbody", { children: t.map((e) => /* @__PURE__ */ (0, q.jsxs)("tr", {
						className: "border-b border-border",
						children: [
							/* @__PURE__ */ (0, q.jsx)("td", {
								className: "px-2 py-1 font-mono",
								children: Me(e.day)
							}),
							n.map((t) => /* @__PURE__ */ (0, q.jsx)("td", {
								className: "px-2 py-1",
								children: Math.round(a(e, t.key))
							}, t.key)),
							/* @__PURE__ */ (0, q.jsx)("td", {
								className: "px-2 py-1",
								children: Math.round(n.reduce((t, n) => t + a(e, n.key), 0))
							})
						]
					}, e.day)) })]
				})]
			})
		]
	});
}
// src/components/platform/windowDays.ts
function Tn(e, t) {
	let n = Date.UTC(t.getUTCFullYear(), t.getUTCMonth(), t.getUTCDate()), r = [];
	for (let t = e; t >= 0; t--) r.push((/* @__PURE__ */ new Date(n - t * 864e5)).toISOString().slice(0, 10));
	return r;
}
// src/components/platform/CaptainHistoryPanel.tsx
var En = 30, Dn = [
	{
		key: "succeeded",
		label: "Succeeded",
		color: "var(--series-1)"
	},
	{
		key: "failed",
		label: "Failed",
		color: "var(--series-2)"
	},
	{
		key: "cancelled",
		label: "Cancelled",
		color: "var(--series-3)"
	},
	{
		key: "active",
		label: "In progress",
		color: "var(--series-4)"
	}
];
function On(e) {
	return e === "succeeded" || e === "failed" || e === "cancelled" ? e : "active";
}
function kn(e, t, n) {
	let r = /* @__PURE__ */ new Map();
	for (let e of Tn(t, n)) r.set(e, {
		day: e,
		succeeded: 0,
		failed: 0,
		cancelled: 0,
		active: 0
	});
	for (let t of e) {
		let e = r.get(t.day);
		e && (e[On(t.state)] += t.runs);
	}
	return [...r.values()];
}
function An(e) {
	let t = /* @__PURE__ */ new Map(), n = (e, n) => {
		let r = e ?? "", i = t.get(r);
		return i || t.set(r, i = {
			key: r,
			name: n || "Unattributed",
			runs: 0,
			failed: 0,
			seconds: 0,
			live: 0,
			unhealthy: {}
		}), i;
	};
	for (let t of e.days) {
		let e = n(t.org, t.orgName);
		e.runs += t.runs, e.seconds += t.runSeconds, t.state === "failed" && (e.failed += t.runs);
	}
	for (let t of e.live) {
		let e = n(t.org, t.orgName);
		e.live += t.sessions, t.health !== "healthy" && (e.unhealthy[t.health] = (e.unhealthy[t.health] ?? 0) + t.sessions);
	}
	return [...t.values()].sort((e, t) => t.runs - e.runs || e.name.localeCompare(t.name));
}
function jn(e) {
	let t = Object.entries(e.unhealthy).map(([e, t]) => `${t} ${e}`);
	return t.length ? `${e.live} (${t.join(", ")})` : String(e.live);
}
function Mn({ title: e, headings: t, rows: n }) {
	return /* @__PURE__ */ (0, q.jsxs)("table", {
		"aria-label": e,
		className: "w-full border-collapse text-left text-xs",
		children: [/* @__PURE__ */ (0, q.jsx)("thead", {
			className: "text-muted-foreground",
			children: /* @__PURE__ */ (0, q.jsx)("tr", { children: t.map((e) => /* @__PURE__ */ (0, q.jsx)("th", {
				className: "border-b border-border px-2 py-1.5 font-medium",
				children: e
			}, e)) })
		}), /* @__PURE__ */ (0, q.jsx)("tbody", { children: n.map((e) => /* @__PURE__ */ (0, q.jsx)("tr", {
			className: "border-b border-border",
			children: e.cells.map((e, t) => /* @__PURE__ */ (0, q.jsx)("td", {
				className: "px-2 py-1.5",
				children: e
			}, t))
		}, e.key)) })]
	});
}
function Nn({ org: e } = {}) {
	let t = (0, K.useQuery)({
		queryKey: m.platform("captain-history", En),
		queryFn: ({ signal: e }) => d({
			url: `/api/platform/captain-history?days=${En}`,
			signal: e,
			context: "load Captain's history"
		})
	}), n = t.data && (e ? {
		days: t.data.days.filter((t) => t.org === e),
		failures: t.data.failures.filter((t) => t.org === e),
		live: t.data.live.filter((t) => t.org === e)
	} : t.data), r = n?.days.reduce((e, t) => e + t.runs, 0) ?? 0, i = n?.days.filter((e) => e.state === "failed").reduce((e, t) => e + t.runs, 0) ?? 0, a = n?.days.reduce((e, t) => e + t.runSeconds, 0) ?? 0, o = n?.live.reduce((e, t) => e + t.sessions, 0) ?? 0, s = n?.live.filter((e) => e.health !== "healthy").reduce((e, t) => e + t.sessions, 0) ?? 0;
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "flex flex-col gap-3",
		children: [
			/* @__PURE__ */ (0, q.jsxs)(M, { children: [
				"Captain history · ",
				En,
				" days"
			] }),
			t.error && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: t.error.message
			}),
			n && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
				/* @__PURE__ */ (0, q.jsx)(L, { items: [
					{
						label: "Prompt runs",
						value: `${r} prompt runs`
					},
					{
						label: "Failed",
						value: `${i} failed (${r ? Math.round(i / r * 100) : 0}%)`
					},
					{
						label: "Run time",
						value: `${Cn(a)} run time`
					},
					{
						label: "Live sessions",
						value: s ? `${o} sessions live, ${s} not healthy` : `${o} sessions live`
					}
				] }),
				/* @__PURE__ */ (0, q.jsx)(wn, {
					title: "Prompt runs per day",
					data: kn(n.days, En, /* @__PURE__ */ new Date()),
					series: Dn,
					unit: "runs"
				}),
				/* @__PURE__ */ (0, q.jsx)(Mn, {
					title: "Captain runs by organization",
					headings: [
						"Organization",
						"Runs",
						"Failed",
						"Run time",
						"Live sessions"
					],
					rows: An(n).map((e) => ({
						key: e.key,
						cells: [
							e.name,
							String(e.runs),
							String(e.failed),
							Cn(e.seconds),
							jn(e)
						]
					}))
				}),
				n.failures.length > 0 && /* @__PURE__ */ (0, q.jsx)(Mn, {
					title: "Latest failed prompt runs",
					headings: [
						"When",
						"Organization",
						"Origin",
						"Error"
					],
					rows: n.failures.map((e) => ({
						key: e.id,
						cells: [
							j(e.createdAt),
							e.orgName || "Unattributed",
							e.origin,
							e.error
						]
					}))
				})
			] })
		]
	});
}
// src/components/platform/GitHubPanel.tsx
function Pn(e) {
	return e.rateLimitError ? e.rateLimitError : e.rateLimit ? `${e.rateLimit.remaining} of ${e.rateLimit.limit} left` : "—";
}
function Fn(e) {
	return e.uninstalled ? "Uninstalled from GitHub" : e.suspendedAt ? `Suspended ${Ce(e.suspendedAt)}` : "";
}
function In({ org: e } = {}) {
	let t = (0, K.useQuery)({
		queryKey: m.platform("github"),
		queryFn: ({ signal: e }) => d({
			url: "/api/platform/github",
			signal: e,
			context: "load the GitHub App"
		})
	}), n = t.data && (e ? {
		...t.data,
		installations: t.data.installations.filter((t) => t.org?.id === e)
	} : t.data), r = n?.installations.filter((e) => e.org).length ?? 0, i = n?.installations.filter((e) => e.suspendedAt).length ?? 0;
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "flex flex-col gap-2",
		children: [
			/* @__PURE__ */ (0, q.jsx)(M, { children: n ? `GitHub App ${n.app.slug} · ${n.installations.length} installations, ${r} linked, ${i} suspended` : "GitHub App" }),
			t.error && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: t.error.message
			}),
			n?.githubError && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-amber-600",
				children: n.githubError
			}),
			!!n?.installations.length && /* @__PURE__ */ (0, q.jsxs)("table", {
				className: "w-full border-collapse text-left text-xs",
				children: [/* @__PURE__ */ (0, q.jsx)("thead", {
					className: "text-muted-foreground",
					children: /* @__PURE__ */ (0, q.jsx)("tr", { children: [
						"Account",
						"Organization",
						"Repositories",
						"Rate limit",
						"State",
						"Installed"
					].map((e) => /* @__PURE__ */ (0, q.jsx)("th", {
						className: "border-b border-border px-2 py-1.5 font-medium",
						children: e
					}, e)) })
				}), /* @__PURE__ */ (0, q.jsx)("tbody", { children: n.installations.map((e) => /* @__PURE__ */ (0, q.jsxs)("tr", {
					className: "border-b border-border",
					children: [
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5 font-medium",
							title: e.accountType,
							children: e.account
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5",
							children: e.org ? e.org.name || e.org.slug : "Not linked"
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5",
							children: e.repos
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: `px-2 py-1.5 ${e.rateLimitError ? "text-red-600" : ""}`,
							children: Pn(e)
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5 text-muted-foreground",
							children: Fn(e)
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5 text-muted-foreground",
							children: e.createdAt ? Ce(e.createdAt) : "—"
						})
					]
				}, e.id)) })]
			})
		]
	});
}
// src/components/platform/PlatformSections.tsx
var Ln = 16;
function J(e) {
	return `platform-${e}`;
}
function Rn({ sections: e, header: t }) {
	let n = e.flatMap((e) => [{
		id: e.id,
		label: e.label,
		depth: 0
	}, ...(e.subsections ?? []).map((e) => ({
		...e,
		depth: 1
	}))]), [r, i] = (0, V.useState)(""), a = n.some((e) => e.id === r) ? r : n[0]?.id, o = (0, V.useRef)(null), s = (0, V.useRef)(null), c = (e) => {
		s.current = e, i(e), document.getElementById(J(e))?.scrollIntoView({
			behavior: "smooth",
			block: "start"
		});
	}, l = () => {
		let e = o.current;
		if (!e || n.length === 0) return;
		if (e.scrollHeight > e.clientHeight && e.scrollTop + e.clientHeight >= e.scrollHeight - 2) {
			i(s.current ?? n[n.length - 1].id);
			return;
		}
		let t = e.getBoundingClientRect().top, r = n[0].id;
		for (let e of n) {
			let n = document.getElementById(J(e.id));
			n && n.getBoundingClientRect().top - t <= Ln && (r = e.id);
		}
		i(r);
	}, u = () => {
		s.current = null;
	};
	return /* @__PURE__ */ (0, q.jsxs)("div", {
		className: "flex h-full min-h-0 text-sm",
		children: [/* @__PURE__ */ (0, q.jsx)("nav", {
			"aria-label": "Platform sections",
			className: "hidden w-48 shrink-0 flex-col gap-0.5 overflow-y-auto border-r border-border p-2 md:flex",
			children: n.map((e) => /* @__PURE__ */ (0, q.jsx)("button", {
				type: "button",
				"data-depth": e.depth,
				"aria-current": e.id === a,
				onClick: () => c(e.id),
				className: `rounded py-1.5 pr-2 text-left hover:bg-muted ${e.depth ? "pl-6 text-xs" : "pl-2"} ${e.id === a ? "bg-primary/10 font-medium" : ""}`,
				children: e.label
			}, e.id))
		}), /* @__PURE__ */ (0, q.jsxs)("div", {
			ref: o,
			"data-testid": "platform-sections-scroll",
			onScroll: l,
			onWheel: u,
			onTouchStart: u,
			onKeyDown: u,
			className: "flex min-w-0 flex-1 flex-col gap-8 overflow-y-auto p-4",
			children: [t, e.map((e) => /* @__PURE__ */ (0, q.jsx)("div", {
				id: J(e.id),
				className: "scroll-mt-4",
				children: e.content
			}, e.id))]
		})]
	});
}
// src/components/platform/InfrastructurePanel.tsx
function zn(e) {
	return {
		healthy: "success",
		unhealthy: "danger",
		warning: "warning"
	}[e] ?? "neutral";
}
function Bn(e) {
	return {
		critical: "danger",
		high: "danger",
		medium: "warning",
		low: "info"
	}[e] ?? "neutral";
}
var Vn = "text-foreground dark:text-foreground";
function Hn({ cell: e }) {
	return e.tone ? e.text ? /* @__PURE__ */ (0, q.jsx)(o, {
		tone: e.tone,
		variant: "soft",
		size: "sm",
		className: Vn,
		children: e.text
	}) : null : e.href ? /* @__PURE__ */ (0, q.jsx)("a", {
		className: "underline",
		href: e.href,
		target: "_blank",
		rel: "noreferrer",
		children: e.text
	}) : /* @__PURE__ */ (0, q.jsx)(q.Fragment, { children: e.text });
}
function Un(e) {
	let t = /* @__PURE__ */ new Map([["healthy", 0], ["unhealthy", 0]]);
	for (let n of e) {
		let e = n.health || "unknown";
		t.set(e, (t.get(e) ?? 0) + 1);
	}
	return [...t].filter(([e, t]) => t > 0 || e === "healthy" || e === "unhealthy").map(([e, t]) => `${t} ${e}`).join(", ");
}
var Wn = "couldn't be read";
function Gn(e) {
	return e ? j(e) : "—";
}
function Kn(e) {
	return e ? [
		{
			id: "infrastructure-catalog",
			label: "Catalog"
		},
		...e.checks.length > 0 || e.checksError ? [{
			id: "infrastructure-checks",
			label: "Health checks"
		}] : [],
		{
			id: "infrastructure-changes",
			label: "Changes"
		}
	] : [];
}
function qn(e, t) {
	return e.type.localeCompare(t.type) || e.name.localeCompare(t.name);
}
function Jn({ label: e, error: t, empty: n, headings: r, rows: i }) {
	return t ? /* @__PURE__ */ (0, q.jsx)("p", {
		role: "alert",
		className: "text-red-600",
		children: t
	}) : i.length === 0 ? /* @__PURE__ */ (0, q.jsx)("p", {
		className: "text-xs text-muted-foreground",
		children: n
	}) : /* @__PURE__ */ (0, q.jsxs)("table", {
		"aria-label": e,
		className: "w-full border-collapse text-left text-xs",
		children: [/* @__PURE__ */ (0, q.jsx)("thead", {
			className: "text-muted-foreground",
			children: /* @__PURE__ */ (0, q.jsx)("tr", { children: r.map((e) => /* @__PURE__ */ (0, q.jsx)("th", {
				className: "border-b border-border px-2 py-1.5 font-medium",
				children: e
			}, e)) })
		}), /* @__PURE__ */ (0, q.jsx)("tbody", { children: i.map((e) => /* @__PURE__ */ (0, q.jsx)("tr", {
			className: "border-b border-border",
			children: e.cells.map((e, t) => /* @__PURE__ */ (0, q.jsx)("td", {
				className: "px-2 py-1.5",
				children: /* @__PURE__ */ (0, q.jsx)(Hn, { cell: e })
			}, t))
		}, e.id)) })]
	});
}
function Yn() {
	return (0, K.useQuery)({
		queryKey: m.platform("infrastructure"),
		queryFn: ({ signal: e }) => d({
			url: "/api/platform/infrastructure",
			signal: e,
			context: "load gavel's infrastructure"
		})
	});
}
function Xn() {
	let e = Yn(), t = e.data, n = new Set(Kn(t).map((e) => e.id)), r = (e) => ({
		text: e.health || "unknown",
		tone: zn(e.health ?? "")
	}), i = t?.missionControl.replace(/\/api\/?$/, "") ?? "", a = (e) => `${i}/catalog/${encodeURIComponent(e)}`;
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "flex flex-col gap-3",
		children: [
			/* @__PURE__ */ (0, q.jsxs)(M, { children: ["Infrastructure", t && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [" · ", /* @__PURE__ */ (0, q.jsx)("a", {
				className: "underline",
				href: i,
				target: "_blank",
				rel: "noreferrer",
				children: t.missionControl
			})] })] }),
			e.error && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: e.error.message
			}),
			t && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
				/* @__PURE__ */ (0, q.jsx)(L, { items: [
					{
						label: "Catalog",
						value: t.configsError ? Wn : Un(t.configs)
					},
					...n.has("infrastructure-checks") ? [{
						label: "Health checks",
						value: t.checksError ? Wn : Un(t.checks)
					}] : [],
					{
						label: "Changes",
						value: t.changesError ? Wn : `${t.changes.length} in ${t.changesDays} days`
					}
				] }),
				/* @__PURE__ */ (0, q.jsxs)("div", {
					id: J("infrastructure-catalog"),
					className: "flex scroll-mt-4 flex-col gap-2",
					children: [/* @__PURE__ */ (0, q.jsx)(De, { children: "Catalog" }), /* @__PURE__ */ (0, q.jsx)(Jn, {
						label: "Catalog",
						error: t.configsError,
						empty: "No config items match the search.",
						headings: [
							"Name",
							"Type",
							"Health",
							"Status",
							"Namespace",
							"Updated"
						],
						rows: [...t.configs].sort(qn).map((e) => ({
							id: e.id,
							cells: [
								{
									text: e.name,
									href: a(e.id)
								},
								{
									text: e.type,
									tone: "neutral"
								},
								r(e),
								{ text: e.status ?? "" },
								{ text: e.namespace },
								{ text: Gn(e.updated_at) }
							]
						}))
					})]
				}),
				n.has("infrastructure-checks") && /* @__PURE__ */ (0, q.jsxs)("div", {
					id: J("infrastructure-checks"),
					className: "flex scroll-mt-4 flex-col gap-2",
					children: [/* @__PURE__ */ (0, q.jsx)(De, { children: "Health checks" }), /* @__PURE__ */ (0, q.jsx)(Jn, {
						label: "Health checks",
						error: t.checksError,
						empty: "No health checks match the search.",
						headings: [
							"Name",
							"Type",
							"Health",
							"Status",
							"Namespace"
						],
						rows: t.checks.map((e) => ({
							id: e.id,
							cells: [
								{ text: e.name },
								{ text: e.type },
								r(e),
								{ text: e.status ?? "" },
								{ text: e.namespace }
							]
						}))
					})]
				}),
				/* @__PURE__ */ (0, q.jsxs)("div", {
					id: J("infrastructure-changes"),
					className: "flex scroll-mt-4 flex-col gap-2",
					children: [/* @__PURE__ */ (0, q.jsx)(De, { children: "Changes" }), /* @__PURE__ */ (0, q.jsx)(Jn, {
						label: "Changes",
						error: t.changesError,
						empty: `No changes to these config items in the last ${t.changesDays} days.`,
						headings: [
							"Config",
							"Type",
							"Change",
							"Summary",
							"Severity",
							"When"
						],
						rows: t.changes.map((e) => ({
							id: e.id,
							cells: [
								{
									text: e.name,
									href: a(e.config_id)
								},
								{
									text: e.type ?? "",
									tone: "neutral"
								},
								{
									text: e.change_type,
									tone: "neutral"
								},
								{ text: e.summary ?? "" },
								{
									text: e.severity ?? "",
									tone: Bn(e.severity ?? "")
								},
								{ text: Gn(e.created_at) }
							]
						}))
					})]
				})
			] })
		]
	});
}
// src/components/platform/ModelUsagePanel.tsx
var Y = 30, X = new Intl.NumberFormat("en-US", {
	style: "currency",
	currency: "USD"
}), Z = new Intl.NumberFormat("en-US");
function Zn(e) {
	return e.inputTokens + e.outputTokens + e.reasoningTokens + e.cacheReadTokens + e.cacheWriteTokens;
}
function Qn(e, t, n) {
	let r = /* @__PURE__ */ new Map();
	for (let e of Tn(t, n)) r.set(e, {
		day: e,
		costUSD: 0
	});
	for (let t of e) {
		let e = r.get(t.day);
		e && (e.costUSD += t.costUSD);
	}
	return [...r.values()];
}
function $n(e, t, n) {
	let r = /* @__PURE__ */ new Map();
	for (let i of e) {
		let e = t(i), a = r.get(e);
		a || r.set(e, a = {
			...n(i),
			calls: 0,
			tokens: 0,
			costUSD: 0
		}), a.calls += i.calls, a.tokens += Zn(i), a.costUSD += i.costUSD;
	}
	return [...r.values()].sort((e, t) => t.costUSD - e.costUSD || e.name.localeCompare(t.name));
}
function er({ title: e, first: t, rows: n }) {
	return /* @__PURE__ */ (0, q.jsxs)("table", {
		"aria-label": e,
		className: "w-full border-collapse text-left text-xs",
		children: [/* @__PURE__ */ (0, q.jsx)("thead", {
			className: "text-muted-foreground",
			children: /* @__PURE__ */ (0, q.jsx)("tr", { children: [
				...t,
				"Calls",
				"Tokens",
				"Cost"
			].map((e) => /* @__PURE__ */ (0, q.jsx)("th", {
				className: "border-b border-border px-2 py-1.5 font-medium",
				children: e
			}, e)) })
		}), /* @__PURE__ */ (0, q.jsx)("tbody", { children: n.map(({ cells: e, share: t }) => /* @__PURE__ */ (0, q.jsxs)("tr", {
			className: "border-b border-border",
			children: [
				e.map((e, t) => /* @__PURE__ */ (0, q.jsx)("td", {
					className: `px-2 py-1.5${t === 0 ? " font-medium" : ""}`,
					children: e
				}, t)),
				/* @__PURE__ */ (0, q.jsx)("td", {
					className: "px-2 py-1.5",
					children: Z.format(t.calls)
				}),
				/* @__PURE__ */ (0, q.jsx)("td", {
					className: "px-2 py-1.5",
					children: Z.format(t.tokens)
				}),
				/* @__PURE__ */ (0, q.jsx)("td", {
					className: "px-2 py-1.5",
					children: X.format(t.costUSD)
				})
			]
		}, t.key)) })]
	});
}
function tr({ data: e }) {
	let t = "Model spend per day (USD)";
	return /* @__PURE__ */ (0, q.jsxs)("figure", {
		className: "flex flex-col gap-1",
		children: [
			/* @__PURE__ */ (0, q.jsx)("figcaption", {
				className: "text-xs text-muted-foreground",
				children: t
			}),
			/* @__PURE__ */ (0, q.jsx)("div", {
				className: "h-48 text-muted-foreground",
				children: /* @__PURE__ */ (0, q.jsx)(ne, {
					width: "100%",
					height: "100%",
					initialDimension: {
						width: 480,
						height: 192
					},
					children: /* @__PURE__ */ (0, q.jsxs)(_n, {
						data: e,
						margin: {
							top: 4,
							right: 8,
							bottom: 0,
							left: 0
						},
						children: [
							/* @__PURE__ */ (0, q.jsx)(be, {
								vertical: !1,
								stroke: "currentColor",
								strokeOpacity: .15
							}),
							/* @__PURE__ */ (0, q.jsx)(ye, {
								dataKey: "day",
								tickFormatter: ze,
								tick: {
									fill: "currentColor",
									fontSize: 11
								},
								tickLine: !1,
								axisLine: !1
							}),
							/* @__PURE__ */ (0, q.jsx)(ge, {
								width: 44,
								tickFormatter: (e) => X.format(e),
								tick: {
									fill: "currentColor",
									fontSize: 11
								},
								tickLine: !1,
								axisLine: !1
							}),
							/* @__PURE__ */ (0, q.jsx)(ie, {
								formatter: (e) => [X.format(Number(e)), "Spent"],
								cursor: { fillOpacity: .08 }
							}),
							/* @__PURE__ */ (0, q.jsx)(hn, {
								dataKey: "costUSD",
								fill: "var(--series-1)",
								radius: [
									4,
									4,
									0,
									0
								],
								maxBarSize: 24,
								isAnimationActive: !1
							})
						]
					})
				})
			}),
			/* @__PURE__ */ (0, q.jsxs)("details", {
				className: "text-xs",
				children: [/* @__PURE__ */ (0, q.jsx)("summary", {
					className: "cursor-pointer text-muted-foreground",
					children: "Show as a table"
				}), /* @__PURE__ */ (0, q.jsxs)("table", {
					"aria-label": t,
					className: "mt-1 w-full border-collapse text-left",
					children: [/* @__PURE__ */ (0, q.jsx)("thead", {
						className: "text-muted-foreground",
						children: /* @__PURE__ */ (0, q.jsx)("tr", { children: ["Day", "Spent"].map((e) => /* @__PURE__ */ (0, q.jsx)("th", {
							className: "border-b border-border px-2 py-1 font-medium",
							children: e
						}, e)) })
					}), /* @__PURE__ */ (0, q.jsx)("tbody", { children: e.map((e) => /* @__PURE__ */ (0, q.jsxs)("tr", {
						className: "border-b border-border",
						children: [/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1 font-mono",
							children: Me(e.day)
						}), /* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1",
							children: X.format(e.costUSD)
						})]
					}, e.day)) })]
				})]
			})
		]
	});
}
function nr({ org: e } = {}) {
	let t = (0, K.useQuery)({
		queryKey: m.platform("model-usage", Y),
		queryFn: ({ signal: e }) => d({
			url: `/api/platform/model-usage?days=${Y}`,
			signal: e,
			context: "load model usage"
		})
	}), n = (t.data?.days ?? []).filter((t) => !e || t.org === e), r = n.reduce((e, t) => e + t.costUSD, 0), i = n.reduce((e, t) => e + t.calls, 0), a = n.reduce((e, t) => e + Zn(t), 0), o = n.reduce((e, t) => e + t.otherCurrencyCalls, 0), s = $n(n, (e) => e.org ?? "", (e) => ({
		key: e.org ?? "",
		name: e.orgName || "Unattributed"
	})), c = $n(n, (e) => `${e.provider}/${e.model}`, (e) => ({
		key: `${e.provider}/${e.model}`,
		name: e.model,
		provider: e.provider
	}));
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "flex flex-col gap-3",
		children: [
			/* @__PURE__ */ (0, q.jsxs)(M, { children: [
				"Model usage · ",
				Y,
				" days"
			] }),
			t.error && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: t.error.message
			}),
			t.data && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
				/* @__PURE__ */ (0, q.jsx)(L, { items: [
					{
						label: "Spent",
						value: `${X.format(r)} spent`
					},
					{
						label: "Calls",
						value: `${Z.format(i)} calls`
					},
					{
						label: "Tokens",
						value: `${Z.format(a)} tokens`
					}
				] }),
				o > 0 && /* @__PURE__ */ (0, q.jsx)("p", {
					role: "status",
					className: "text-xs text-muted-foreground",
					children: o === 1 ? "1 call was priced in another currency: its tokens count, its cost isn't in these totals." : `${Z.format(o)} calls were priced in another currency: their tokens count, their cost isn't in these totals.`
				}),
				/* @__PURE__ */ (0, q.jsx)(tr, { data: Qn(n, Y, /* @__PURE__ */ new Date()) }),
				/* @__PURE__ */ (0, q.jsxs)("div", {
					className: "grid gap-4 lg:grid-cols-2",
					children: [/* @__PURE__ */ (0, q.jsx)(er, {
						title: "Model spend by organization",
						first: ["Organization"],
						rows: s.map((e) => ({
							cells: [e.name],
							share: e
						}))
					}), /* @__PURE__ */ (0, q.jsx)(er, {
						title: "Model spend by model",
						first: ["Model", "Provider"],
						rows: c.map((e) => ({
							cells: [e.name, e.provider ?? ""],
							share: e
						}))
					})]
				})
			] })
		]
	});
}
// src/components/platform/SandboxSizePanel.tsx
var rr = 2 ** 30, ir = "w-28 rounded border border-border bg-background px-2 py-1 text-sm";
function ar(e) {
	return String(Number(e.toFixed(2)));
}
function or(e) {
	return `${ar(e.cpuMillicores / 1e3)} CPU · ${ar(e.memoryBytes / rr)} GiB`;
}
function sr(e) {
	return e instanceof Error ? e.message : String(e);
}
function cr({ org: e }) {
	let t = `/api/platform/orgs/${encodeURIComponent(e)}/sandbox-size`, n = m.platform(`orgs/${e}/sandbox-size`), r = (0, K.useQueryClient)(), i = (0, K.useQuery)({
		queryKey: n,
		queryFn: ({ signal: e }) => d({
			url: t,
			signal: e,
			context: "load the organization's sandbox size"
		})
	}), [o, s] = (0, V.useState)(""), [c, l] = (0, V.useState)(""), u = (e) => r.setQueryData(n, e), f = (0, K.useMutation)({
		mutationFn: () => g({
			url: t,
			method: "PUT",
			context: "save the sandbox size",
			body: {
				cpuMillicores: Math.round(Number(o) * 1e3),
				memoryBytes: Math.round(Number(c) * rr)
			}
		}),
		onSuccess: u
	}), p = (0, K.useMutation)({
		mutationFn: () => g({
			url: t,
			method: "DELETE",
			context: "return to the default sandbox size"
		}),
		onSuccess: u
	}), h = i.data, _ = i.error ?? f.error ?? p.error;
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "flex flex-col gap-3 text-sm",
		children: [
			/* @__PURE__ */ (0, q.jsx)(M, { children: "Sandbox size" }),
			/* @__PURE__ */ (0, q.jsx)("p", {
				className: "text-muted-foreground",
				children: "The CPU and memory substrate gives each of this organization's sandbox environments, and holds it to. Runs started from now on get the new size; runs already going keep theirs."
			}),
			h && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
				h.size ? /* @__PURE__ */ (0, q.jsx)("p", { children: `${or(h.size)} per environment, set for this organization` }) : h.default && /* @__PURE__ */ (0, q.jsx)("p", { children: `${or(h.default)} per environment, the default` }),
				h.size && h.default && /* @__PURE__ */ (0, q.jsx)("p", {
					className: "text-muted-foreground",
					children: `The default is ${or(h.default)}.`
				}),
				h.defaultError && /* @__PURE__ */ (0, q.jsx)("p", {
					className: "text-muted-foreground",
					children: `gavel couldn't read the default: ${h.defaultError}`
				}),
				/* @__PURE__ */ (0, q.jsxs)("form", {
					className: "flex flex-wrap items-end gap-3",
					onSubmit: (e) => {
						e.preventDefault(), f.mutate();
					},
					children: [
						/* @__PURE__ */ (0, q.jsxs)("label", {
							className: "flex flex-col gap-1",
							children: [/* @__PURE__ */ (0, q.jsx)("span", { children: "CPU (cores)" }), /* @__PURE__ */ (0, q.jsx)("input", {
								className: ir,
								type: "number",
								min: "0.1",
								step: "0.1",
								value: o,
								onChange: (e) => s(e.target.value)
							})]
						}),
						/* @__PURE__ */ (0, q.jsxs)("label", {
							className: "flex flex-col gap-1",
							children: [/* @__PURE__ */ (0, q.jsx)("span", { children: "Memory (GiB)" }), /* @__PURE__ */ (0, q.jsx)("input", {
								className: ir,
								type: "number",
								min: "0.25",
								step: "0.25",
								value: c,
								onChange: (e) => l(e.target.value)
							})]
						}),
						/* @__PURE__ */ (0, q.jsx)(a, {
							type: "submit",
							disabled: f.isPending || !o || !c,
							children: "Save size"
						}),
						h.size && /* @__PURE__ */ (0, q.jsx)(a, {
							type: "button",
							variant: "ghost",
							disabled: p.isPending,
							onClick: () => p.mutate(),
							children: "Use the default"
						})
					]
				})
			] }),
			_ && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: sr(_)
			})
		]
	});
}
// src/components/platform/SandboxUsagePanel.tsx
var Q = 30, lr = [
	"test",
	"lint",
	"push-check",
	"agent"
];
function ur(e, t, n, r) {
	let i = /* @__PURE__ */ new Map();
	for (let e of Tn(t, n)) i.set(e, {
		day: e,
		test: 0,
		lint: 0,
		"push-check": 0,
		agent: 0
	});
	for (let t of e) {
		let e = i.get(t.day), n = lr.find((e) => e === t.kind);
		e && n && (e[n] += r(t));
	}
	return [...i.values()];
}
function dr(e) {
	let t = /* @__PURE__ */ new Map(), n = (e, n) => {
		let r = t.get(e);
		return r || t.set(e, r = {
			org: e,
			name: n,
			runs: 0,
			failed: 0,
			seconds: 0,
			active: 0
		}), r;
	};
	for (let t of e.days) {
		let e = n(t.org, t.orgName);
		e.runs += t.runs, e.seconds += t.runSeconds, t.status === "failed" && (e.failed += t.runs);
	}
	for (let t of e.active) n(t.org, t.orgName).active += t.runs;
	return [...t.values()].sort((e, t) => t.runs - e.runs || e.name.localeCompare(t.name));
}
var fr = lr.map((e) => ({
	key: e,
	label: _[e] ?? e,
	color: `var(--kind-${e})`
}));
function pr({ org: e } = {}) {
	let t = (0, K.useQuery)({
		queryKey: m.platform("sandbox-runs", Q),
		queryFn: ({ signal: e }) => d({
			url: `/api/platform/sandbox-runs?days=${Q}`,
			signal: e,
			context: "load sandbox usage"
		})
	}), n = t.data && (e ? {
		days: t.data.days.filter((t) => t.org === e),
		active: t.data.active.filter((t) => t.org === e)
	} : t.data), r = n?.days.reduce((e, t) => e + t.runs, 0) ?? 0, i = n?.days.filter((e) => e.status === "failed").reduce((e, t) => e + t.runs, 0) ?? 0, a = n?.days.reduce((e, t) => e + t.runSeconds, 0) ?? 0, o = n?.active.reduce((e, t) => e + t.runs, 0) ?? 0, s = /* @__PURE__ */ new Date();
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "flex flex-col gap-3",
		children: [
			/* @__PURE__ */ (0, q.jsxs)(M, { children: [
				"Sandbox usage · ",
				Q,
				" days"
			] }),
			t.error && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: t.error.message
			}),
			n && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
				/* @__PURE__ */ (0, q.jsx)(L, { items: [
					{
						label: "Runs",
						value: `${r} runs`
					},
					{
						label: "Failed",
						value: `${i} failed (${r ? Math.round(i / r * 100) : 0}%)`
					},
					{
						label: "Sandbox time",
						value: `${Cn(a)} sandbox time`
					},
					{
						label: "Queued or running",
						value: `${o} active now`
					}
				] }),
				/* @__PURE__ */ (0, q.jsxs)("div", {
					className: "grid gap-4 lg:grid-cols-2",
					children: [/* @__PURE__ */ (0, q.jsx)(wn, {
						title: "Runs per day",
						data: ur(n.days, Q, s, (e) => e.runs),
						series: fr,
						unit: "runs"
					}), /* @__PURE__ */ (0, q.jsx)(wn, {
						title: "Sandbox minutes per day",
						data: ur(n.days, Q, s, (e) => e.runSeconds / 60),
						series: fr,
						unit: "min"
					})]
				}),
				/* @__PURE__ */ (0, q.jsxs)("table", {
					className: "w-full border-collapse text-left text-xs",
					children: [/* @__PURE__ */ (0, q.jsx)("thead", {
						className: "text-muted-foreground",
						children: /* @__PURE__ */ (0, q.jsx)("tr", { children: [
							"Organization",
							"Runs",
							"Failed",
							"Sandbox time",
							"Active now"
						].map((e) => /* @__PURE__ */ (0, q.jsx)("th", {
							className: "border-b border-border px-2 py-1.5 font-medium",
							children: e
						}, e)) })
					}), /* @__PURE__ */ (0, q.jsx)("tbody", { children: dr(n).map((e) => /* @__PURE__ */ (0, q.jsxs)("tr", {
						className: "border-b border-border",
						children: [
							/* @__PURE__ */ (0, q.jsx)("td", {
								className: "px-2 py-1.5 font-medium",
								children: e.name
							}),
							/* @__PURE__ */ (0, q.jsx)("td", {
								className: "px-2 py-1.5",
								children: e.runs
							}),
							/* @__PURE__ */ (0, q.jsx)("td", {
								className: "px-2 py-1.5",
								children: e.failed
							}),
							/* @__PURE__ */ (0, q.jsx)("td", {
								className: "px-2 py-1.5",
								children: Cn(e.seconds)
							}),
							/* @__PURE__ */ (0, q.jsx)("td", {
								className: "px-2 py-1.5",
								children: e.active
							})
						]
					}, e.org)) })]
				})
			] })
		]
	});
}
// src/components/platform/SubstrateInventoryPanel.tsx
var mr = 1 << 30;
function $(e) {
	return String(Number(e.toFixed(1)));
}
var hr = [
	"running",
	"suspended",
	"crashed"
];
function gr(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) {
		let [e, r, i] = n.org ? [
			n.org.org,
			n.org.orgName,
			0
		] : n.run ? [
			"unknown",
			"A run gavel doesn't know",
			1
		] : [
			`atespace/${n.atespace}`,
			`Not a gavel run, in ${n.atespace}`,
			2
		], a = t.get(e);
		a || t.set(e, a = {
			key: e,
			name: r,
			order: i,
			states: {},
			total: 0
		});
		let o = hr.includes(n.state) ? n.state : "other";
		a.states[o] = (a.states[o] ?? 0) + 1, a.total++;
	}
	return [...t.values()].sort((e, t) => e.order - t.order || e.name.localeCompare(t.name));
}
function _r({ title: e, headings: t, rows: n }) {
	return /* @__PURE__ */ (0, q.jsxs)("table", {
		"aria-label": e,
		className: "w-full border-collapse text-left text-xs",
		children: [/* @__PURE__ */ (0, q.jsx)("thead", {
			className: "text-muted-foreground",
			children: /* @__PURE__ */ (0, q.jsx)("tr", { children: t.map((e) => /* @__PURE__ */ (0, q.jsx)("th", {
				className: "border-b border-border px-2 py-1.5 font-medium",
				children: e
			}, e)) })
		}), /* @__PURE__ */ (0, q.jsx)("tbody", { children: n.map((e) => /* @__PURE__ */ (0, q.jsx)("tr", {
			className: "border-b border-border",
			children: e.cells.map((e, t) => /* @__PURE__ */ (0, q.jsx)("td", {
				className: `px-2 py-1.5${t === 0 ? " font-medium" : ""}`,
				children: e
			}, t))
		}, e.key)) })]
	});
}
function vr({ org: e } = {}) {
	let t = (0, K.useQuery)({
		queryKey: m.platform("substrate-inventory"),
		queryFn: ({ signal: e }) => d({
			url: "/api/platform/substrate/inventory",
			signal: e,
			context: "load substrate's inventory"
		})
	}), n = t.data && (e ? {
		...t.data,
		actors: t.data.actors.filter((t) => t.org?.org === e)
	} : t.data), r = n?.workers ?? [], i = (e) => r.reduce((t, n) => t + e(n), 0);
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "flex flex-col gap-3",
		children: [
			/* @__PURE__ */ (0, q.jsx)(M, { children: "Substrate inventory" }),
			t.error && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: t.error.message
			}),
			n?.workersError && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: n.workersError
			}),
			n?.actorsError && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: n.actorsError
			}),
			n && !e && !n.workersError && /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [/* @__PURE__ */ (0, q.jsx)(L, { items: [
				{
					label: "Workers",
					value: `${r.filter((e) => e.state === "active").length} of ${r.length} workers active`
				},
				{
					label: "Actor slots",
					value: `${i((e) => e.allocated.actors)} of ${i((e) => e.capacity.actors)} actor slots used`
				},
				{
					label: "CPU",
					value: `${$(i((e) => e.allocated.cpuMillicores) / 1e3)} of ${$(i((e) => e.capacity.cpuMillicores) / 1e3)} cores allocated`
				},
				{
					label: "Memory",
					value: `${$(i((e) => e.allocated.memoryBytes) / mr)} of ${$(i((e) => e.capacity.memoryBytes) / mr)} GiB allocated`
				}
			] }), /* @__PURE__ */ (0, q.jsx)(_r, {
				title: "Substrate workers",
				headings: [
					"Worker",
					"Pool",
					"Node",
					"State",
					"Cores",
					"Memory (GiB)",
					"Actors"
				],
				rows: r.map((e) => ({
					key: e.name,
					cells: [
						e.name,
						e.pool,
						e.node,
						e.state,
						`${$(e.allocated.cpuMillicores / 1e3)} / ${$(e.capacity.cpuMillicores / 1e3)}`,
						`${$(e.allocated.memoryBytes / mr)} / ${$(e.capacity.memoryBytes / mr)}`,
						`${e.allocated.actors} / ${e.capacity.actors}`
					]
				}))
			})] }),
			n && !n.actorsError && (n.actors.length === 0 ? /* @__PURE__ */ (0, q.jsx)("p", {
				className: "text-xs text-muted-foreground",
				children: "No actors in any atespace."
			}) : /* @__PURE__ */ (0, q.jsx)(_r, {
				title: "Substrate actors by owner",
				headings: [
					"Owner",
					"Running",
					"Suspended",
					"Crashed",
					"Other",
					"Total"
				],
				rows: gr(n.actors).map((e) => ({
					key: e.key,
					cells: [
						e.name,
						...[...hr, "other"].map((t) => String(e.states[t] ?? 0)),
						String(e.total)
					]
				}))
			}))
		]
	});
}
function yr(e) {
	return e.clerkMembers === void 0 ? `${e.members}` : `${e.members} of ${e.clerkMembers}`;
}
function br(e) {
	return e.pushedRepos > 0 ? `${e.repos} (${e.pushedRepos} pushed)` : `${e.repos}`;
}
function xr(e, t) {
	return e.inGavel ? t && e.clerkMembers === void 0 ? "Not in Clerk" : "" : "Clerk only";
}
function Sr() {
	return (0, K.useQuery)({
		queryKey: m.platform("tenants", 30),
		queryFn: ({ signal: e }) => d({
			url: "/api/platform/tenants?days=30",
			signal: e,
			context: "load tenants"
		})
	});
}
function Cr({ onSelect: e } = {}) {
	let t = Sr(), n = !t.data?.clerkError;
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "flex flex-col gap-2",
		children: [
			/* @__PURE__ */ (0, q.jsxs)(M, { children: ["Tenants", t.data ? ` · ${t.data.tenants.length}` : ""] }),
			t.error && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: t.error.message
			}),
			t.data?.clerkError && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-amber-600",
				children: t.data.clerkError
			}),
			!!t.data?.tenants.length && /* @__PURE__ */ (0, q.jsxs)("table", {
				className: "w-full border-collapse text-left text-xs",
				children: [/* @__PURE__ */ (0, q.jsx)("thead", {
					className: "text-muted-foreground",
					children: /* @__PURE__ */ (0, q.jsx)("tr", { children: [
						"Organization",
						"Members (gavel of Clerk)",
						"Repositories",
						"GitHub installs",
						"Runs (30d)",
						"Last run",
						"Created",
						""
					].map((e) => /* @__PURE__ */ (0, q.jsx)("th", {
						className: "border-b border-border px-2 py-1.5 font-medium",
						children: e
					}, e)) })
				}), /* @__PURE__ */ (0, q.jsx)("tbody", { children: t.data.tenants.map((t) => /* @__PURE__ */ (0, q.jsxs)("tr", {
					className: "border-b border-border",
					children: [
						/* @__PURE__ */ (0, q.jsxs)("td", {
							className: "px-2 py-1.5",
							children: [t.id && e ? /* @__PURE__ */ (0, q.jsx)("button", {
								type: "button",
								className: "font-medium underline",
								onClick: () => e(t.id),
								children: t.name || t.clerkOrgId
							}) : /* @__PURE__ */ (0, q.jsx)("span", {
								className: "font-medium",
								children: t.name || t.clerkOrgId
							}), t.slug && /* @__PURE__ */ (0, q.jsx)("span", {
								className: "ml-1 text-muted-foreground",
								children: t.slug
							})]
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5",
							children: yr(t)
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5",
							children: br(t)
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5",
							children: t.githubInstallations
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5",
							children: t.sandboxRuns
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5 text-muted-foreground",
							children: t.lastRunAt ? j(t.lastRunAt) : "—"
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5 text-muted-foreground",
							children: Ce(t.createdAt)
						}),
						/* @__PURE__ */ (0, q.jsx)("td", {
							className: "px-2 py-1.5 text-muted-foreground",
							children: xr(t, n)
						})
					]
				}, t.clerkOrgId)) })]
			})
		]
	});
}
// src/components/platform/OrganizationView.tsx
function wr({ org: e, onBack: t }) {
	let n = Sr(), r = n.data?.tenants.find((t) => t.id === e), i = r && /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "flex flex-col gap-2",
		children: [/* @__PURE__ */ (0, q.jsxs)("h2", {
			className: "text-xl font-semibold",
			children: [
				r.name || r.clerkOrgId,
				" ",
				/* @__PURE__ */ (0, q.jsx)("span", {
					className: "text-xs font-normal text-muted-foreground",
					children: r.slug
				})
			]
		}), /* @__PURE__ */ (0, q.jsx)(L, { items: [
			{
				label: "Members",
				value: r.clerkMembers === void 0 ? `${r.members} members` : `${r.members} of ${r.clerkMembers} members`
			},
			{
				label: "Repositories",
				value: r.pushedRepos > 0 ? `${r.repos} repositories (${r.pushedRepos} pushed)` : `${r.repos} repositories`
			},
			{
				label: "GitHub",
				value: `${r.githubInstallations} GitHub installations`
			},
			{
				label: "Last run",
				value: r.lastRunAt ? `Last run ${j(r.lastRunAt)}` : "No runs in 30 days"
			}
		] })]
	});
	return /* @__PURE__ */ (0, q.jsx)(Rn, {
		header: /* @__PURE__ */ (0, q.jsxs)(q.Fragment, { children: [
			/* @__PURE__ */ (0, q.jsx)("button", {
				type: "button",
				className: "self-start text-xs underline text-muted-foreground",
				onClick: t,
				children: "All organizations"
			}),
			n.error && /* @__PURE__ */ (0, q.jsx)("p", {
				role: "alert",
				className: "text-red-600",
				children: n.error.message
			}),
			n.data && !r && /* @__PURE__ */ (0, q.jsxs)("p", {
				role: "alert",
				className: "text-red-600",
				children: [
					"gavel has no organization ",
					e,
					" in this Clerk instance."
				]
			})
		] }),
		sections: r ? [
			{
				id: "summary",
				label: r.name || r.clerkOrgId,
				content: i
			},
			{
				id: "sandbox-size",
				label: "Sandbox size",
				content: /* @__PURE__ */ (0, q.jsx)(cr, { org: e })
			},
			{
				id: "sandbox-usage",
				label: "Sandbox usage",
				content: /* @__PURE__ */ (0, q.jsx)(pr, { org: e })
			},
			{
				id: "measured-usage",
				label: "Measured CPU and memory",
				content: /* @__PURE__ */ (0, q.jsx)(Ve, { org: e })
			},
			{
				id: "model-usage",
				label: "Model usage",
				content: /* @__PURE__ */ (0, q.jsx)(nr, { org: e })
			},
			{
				id: "captain-history",
				label: "Captain history",
				content: /* @__PURE__ */ (0, q.jsx)(Nn, { org: e })
			},
			{
				id: "substrate-inventory",
				label: "Substrate inventory",
				content: /* @__PURE__ */ (0, q.jsx)(vr, { org: e })
			},
			{
				id: "github-app",
				label: "GitHub App",
				content: /* @__PURE__ */ (0, q.jsx)(In, { org: e })
			}
		] : []
	});
}
// src/components/platform/PlatformView.tsx
function Tr({ selectedOrg: e, onSelectOrg: t }) {
	let n = (0, K.useQuery)(x());
	return n.error ? /* @__PURE__ */ (0, q.jsx)("p", {
		role: "alert",
		className: "p-4 text-sm text-red-600",
		children: n.error.message
	}) : n.data ? n.data.user?.platformAdmin ? e ? /* @__PURE__ */ (0, q.jsx)(wr, {
		org: e,
		onBack: () => t("")
	}) : /* @__PURE__ */ (0, q.jsx)(Er, { onSelectOrg: t }) : /* @__PURE__ */ (0, q.jsx)(Dr, {}) : null;
}
function Er({ onSelectOrg: e }) {
	let t = Yn();
	return /* @__PURE__ */ (0, q.jsx)(Rn, { sections: [
		{
			id: "tenants",
			label: "Tenants",
			content: /* @__PURE__ */ (0, q.jsx)(Cr, { onSelect: e })
		},
		{
			id: "sandbox-usage",
			label: "Sandbox usage",
			content: /* @__PURE__ */ (0, q.jsx)(pr, {})
		},
		{
			id: "substrate-inventory",
			label: "Substrate inventory",
			content: /* @__PURE__ */ (0, q.jsx)(vr, {})
		},
		{
			id: "substrate-metrics",
			label: "Substrate metrics",
			content: /* @__PURE__ */ (0, q.jsx)(Le, {})
		},
		{
			id: "model-usage",
			label: "Model usage",
			content: /* @__PURE__ */ (0, q.jsx)(nr, {})
		},
		{
			id: "budgets",
			label: "Budgets",
			content: /* @__PURE__ */ (0, q.jsx)(Sn, {})
		},
		{
			id: "captain-history",
			label: "Captain history",
			content: /* @__PURE__ */ (0, q.jsx)(Nn, {})
		},
		{
			id: "github-app",
			label: "GitHub App",
			content: /* @__PURE__ */ (0, q.jsx)(In, {})
		},
		{
			id: "infrastructure",
			label: "Infrastructure",
			content: /* @__PURE__ */ (0, q.jsx)(Xn, {}),
			subsections: Kn(t.data)
		}
	] });
}
function Dr() {
	return /* @__PURE__ */ (0, q.jsxs)("section", {
		className: "flex flex-col gap-3 p-4 text-sm",
		children: [
			/* @__PURE__ */ (0, q.jsx)("h2", {
				className: "text-base font-semibold",
				children: "Platform admins only"
			}),
			/* @__PURE__ */ (0, q.jsx)("p", { children: "The platform dashboard shows every organization's use of gavel, so only platform admins can open it." }),
			/* @__PURE__ */ (0, q.jsx)("p", { children: "To become one, add this to your user's public metadata in Clerk (Users → your user → Metadata):" }),
			/* @__PURE__ */ (0, q.jsx)("pre", {
				className: "self-start rounded border border-border bg-muted px-3 py-2 font-mono text-xs",
				children: "{ \"gavel_role\": \"platform_admin\" }"
			}),
			/* @__PURE__ */ (0, q.jsx)("p", {
				className: "text-muted-foreground",
				children: "gavel reads it within 5 minutes; reload this page then."
			})
		]
	});
}
export { Tr as PlatformView };
