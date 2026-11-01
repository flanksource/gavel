import { r as e, t } from "./chunks/rolldown-runtime-QGRDbDUS.js";
import { n, t as r } from "./chunks/jsx-runtime-f3E3oPuc.js";
import { t as i } from "./chunks/react-dom-D8HDTZ2P.js";
import { A as a, C as o, E as s, F as c, G as l, I as u, J as d, K as f, N as p, O as m, P as h, Q as g, R as _, T as v, U as y, V as b, W as x, X as S, Y as C, Z as w, _ as T, a as E, b as D, c as O, ct as ee, d as te, dt as ne, et as re, f as ie, g as k, h as A, i as j, k as ae, l as oe, m as M, o as se, ot as ce, p as N, q as le, r as ue, s as de, st as fe, t as pe, u as me, ut as he, v as ge, x as _e, y as ve } from "./chunks/adapters-DVesfYVL.js";
import { C as ye, T as P, _ as be, i as xe, t as Se, x as Ce } from "./chunks/Callout-DALwzckz.js";
import { $ as we, Et as F, F as Te, H as Ee, I as De, Tt as Oe, U as ke, V as Ae, W as je, Z as Me, _t as Ne, a as Pe, at as Fe, bt as Ie, c as Le, et as Re, g as ze, ht as Be, it as Ve, l as He, lt as Ue, nt as We, o as Ge, p as Ke, r as qe, rt as Je, t as Ye, ut as Xe, xt as Ze } from "./chunks/clipboard-CMVj3cUp.js";
import { $ as Qe, E as $e, K as et, Q as tt, S as nt, T as rt, U as it, Z as at, _ as ot, c as st, dt as ct, f as lt, h as ut, i as dt, k as ft, l as pt, m as mt, nt as ht, o as gt, rt as _t, s as vt, st as yt, v as bt } from "./chunks/paths-D_tC_cYe.js";
import { c as xt, d as St, f as Ct, g as wt, h as Tt, l as Et, m as Dt, n as Ot, o as kt, p as At, r as jt, s as Mt, u as Nt } from "./chunks/adapters-B0rO-m88.js";
import { a as Pt, f as Ft, h as It, i as Lt, l as Rt, m as zt, n as Bt, r as Vt, s as Ht, t as Ut, u as Wt } from "./chunks/Tabs-Dd-b3nbz.js";
// ../../../../../../pnpm/store/v11/links/@/scheduler/0.27.0/7fc31d4cd7fa696a9f8118a856b7786dcf00c3d63451f7e09fad708c96f7d341/node_modules/scheduler/cjs/scheduler.production.js
var Gt = /* @__PURE__ */ t(((e) => {
	function t(e, t) {
		var n = e.length;
		e.push(t);
		a: for (; 0 < n;) {
			var r = n - 1 >>> 1, a = e[r];
			if (0 < i(a, t)) e[r] = t, e[n] = a, n = r;
			else break a;
		}
	}
	function n(e) {
		return e.length === 0 ? null : e[0];
	}
	function r(e) {
		if (e.length === 0) return null;
		var t = e[0], n = e.pop();
		if (n !== t) {
			e[0] = n;
			a: for (var r = 0, a = e.length, o = a >>> 1; r < o;) {
				var s = 2 * (r + 1) - 1, c = e[s], l = s + 1, u = e[l];
				if (0 > i(c, n)) l < a && 0 > i(u, c) ? (e[r] = u, e[l] = n, r = l) : (e[r] = c, e[s] = n, r = s);
				else if (l < a && 0 > i(u, n)) e[r] = u, e[l] = n, r = l;
				else break a;
			}
		}
		return t;
	}
	function i(e, t) {
		var n = e.sortIndex - t.sortIndex;
		return n === 0 ? e.id - t.id : n;
	}
	if (e.unstable_now = void 0, typeof performance == "object" && typeof performance.now == "function") {
		var a = performance;
		e.unstable_now = function() {
			return a.now();
		};
	} else {
		var o = Date, s = o.now();
		e.unstable_now = function() {
			return o.now() - s;
		};
	}
	var c = [], l = [], u = 1, d = null, f = 3, p = !1, m = !1, h = !1, g = !1, _ = typeof setTimeout == "function" ? setTimeout : null, v = typeof clearTimeout == "function" ? clearTimeout : null, y = typeof setImmediate < "u" ? setImmediate : null;
	function b(e) {
		for (var i = n(l); i !== null;) {
			if (i.callback === null) r(l);
			else if (i.startTime <= e) r(l), i.sortIndex = i.expirationTime, t(c, i);
			else break;
			i = n(l);
		}
	}
	function x(e) {
		if (h = !1, b(e), !m) {
			if (n(c) !== null) m = !0, S || (S = !0, O());
			else {
				var t = n(l);
				t !== null && ne(x, t.startTime - e);
			}
		}
	}
	var S = !1, C = -1, w = 5, T = -1;
	function E() {
		return g ? !0 : !(e.unstable_now() - T < w);
	}
	function D() {
		if (g = !1, S) {
			var t = e.unstable_now();
			T = t;
			var i = !0;
			try {
				a: {
					m = !1, h && (h = !1, v(C), C = -1), p = !0;
					var a = f;
					try {
						b: {
							for (b(t), d = n(c); d !== null && !(d.expirationTime > t && E());) {
								var o = d.callback;
								if (typeof o == "function") {
									d.callback = null, f = d.priorityLevel;
									var s = o(d.expirationTime <= t);
									if (t = e.unstable_now(), typeof s == "function") {
										d.callback = s, b(t), i = !0;
										break b;
									}
									d === n(c) && r(c), b(t);
								} else r(c);
								d = n(c);
							}
							if (d !== null) i = !0;
							else {
								var u = n(l);
								u !== null && ne(x, u.startTime - t), i = !1;
							}
						}
						break a;
					} finally {
						d = null, f = a, p = !1;
					}
					i = void 0;
				}
			} finally {
				i ? O() : S = !1;
			}
		}
	}
	var O;
	if (typeof y == "function") O = function() {
		y(D);
	};
	else if (typeof MessageChannel < "u") {
		var ee = new MessageChannel(), te = ee.port2;
		ee.port1.onmessage = D, O = function() {
			te.postMessage(null);
		};
	} else O = function() {
		_(D, 0);
	};
	function ne(t, n) {
		C = _(function() {
			t(e.unstable_now());
		}, n);
	}
	e.unstable_IdlePriority = 5, e.unstable_ImmediatePriority = 1, e.unstable_LowPriority = 4, e.unstable_NormalPriority = 3, e.unstable_Profiling = null, e.unstable_UserBlockingPriority = 2, e.unstable_cancelCallback = function(e) {
		e.callback = null;
	}, e.unstable_forceFrameRate = function(e) {
		0 > e || 125 < e ? console.error("forceFrameRate takes a positive int between 0 and 125, forcing frame rates higher than 125 fps is not supported") : w = 0 < e ? Math.floor(1e3 / e) : 5;
	}, e.unstable_getCurrentPriorityLevel = function() {
		return f;
	}, e.unstable_next = function(e) {
		switch (f) {
			case 1:
			case 2:
			case 3:
				var t = 3;
				break;
			default: t = f;
		}
		var n = f;
		f = t;
		try {
			return e();
		} finally {
			f = n;
		}
	}, e.unstable_requestPaint = function() {
		g = !0;
	}, e.unstable_runWithPriority = function(e, t) {
		switch (e) {
			case 1:
			case 2:
			case 3:
			case 4:
			case 5: break;
			default: e = 3;
		}
		var n = f;
		f = e;
		try {
			return t();
		} finally {
			f = n;
		}
	}, e.unstable_scheduleCallback = function(r, i, a) {
		var o = e.unstable_now();
		switch (typeof a == "object" && a ? (a = a.delay, a = typeof a == "number" && 0 < a ? o + a : o) : a = o, r) {
			case 1:
				var s = -1;
				break;
			case 2:
				s = 250;
				break;
			case 5:
				s = 1073741823;
				break;
			case 4:
				s = 1e4;
				break;
			default: s = 5e3;
		}
		return s = a + s, r = {
			id: u++,
			callback: i,
			priorityLevel: r,
			startTime: a,
			expirationTime: s,
			sortIndex: -1
		}, a > o ? (r.sortIndex = a, t(l, r), n(c) === null && r === n(l) && (h ? (v(C), C = -1) : h = !0, ne(x, a - o))) : (r.sortIndex = s, t(c, r), m || p || (m = !0, S || (S = !0, O()))), r;
	}, e.unstable_shouldYield = E, e.unstable_wrapCallback = function(e) {
		var t = f;
		return function() {
			var n = f;
			f = t;
			try {
				return e.apply(this, arguments);
			} finally {
				f = n;
			}
		};
	};
})), Kt = /* @__PURE__ */ t(((e, t) => {
	t.exports = Gt();
})), qt = /* @__PURE__ */ t(((e) => {
	var t = Kt(), r = n(), a = i();
	function o(e) {
		var t = "https://react.dev/errors/" + e;
		if (1 < arguments.length) {
			t += "?args[]=" + encodeURIComponent(arguments[1]);
			for (var n = 2; n < arguments.length; n++) t += "&args[]=" + encodeURIComponent(arguments[n]);
		}
		return "Minified React error #" + e + "; visit " + t + " for the full message or use the non-minified dev environment for full errors and additional helpful warnings.";
	}
	function s(e) {
		return !(!e || e.nodeType !== 1 && e.nodeType !== 9 && e.nodeType !== 11);
	}
	function c(e) {
		var t = e, n = e;
		if (e.alternate) for (; t.return;) t = t.return;
		else {
			e = t;
			do
				t = e, t.flags & 4098 && (n = t.return), e = t.return;
			while (e);
		}
		return t.tag === 3 ? n : null;
	}
	function l(e) {
		if (e.tag === 13) {
			var t = e.memoizedState;
			if (t === null && (e = e.alternate, e !== null && (t = e.memoizedState)), t !== null) return t.dehydrated;
		}
		return null;
	}
	function u(e) {
		if (e.tag === 31) {
			var t = e.memoizedState;
			if (t === null && (e = e.alternate, e !== null && (t = e.memoizedState)), t !== null) return t.dehydrated;
		}
		return null;
	}
	function d(e) {
		if (c(e) !== e) throw Error(o(188));
	}
	function f(e) {
		var t = e.alternate;
		if (!t) {
			if (t = c(e), t === null) throw Error(o(188));
			return t === e ? e : null;
		}
		for (var n = e, r = t;;) {
			var i = n.return;
			if (i === null) break;
			var a = i.alternate;
			if (a === null) {
				if (r = i.return, r !== null) {
					n = r;
					continue;
				}
				break;
			}
			if (i.child === a.child) {
				for (a = i.child; a;) {
					if (a === n) return d(i), e;
					if (a === r) return d(i), t;
					a = a.sibling;
				}
				throw Error(o(188));
			}
			if (n.return !== r.return) n = i, r = a;
			else {
				for (var s = !1, l = i.child; l;) {
					if (l === n) {
						s = !0, n = i, r = a;
						break;
					}
					if (l === r) {
						s = !0, r = i, n = a;
						break;
					}
					l = l.sibling;
				}
				if (!s) {
					for (l = a.child; l;) {
						if (l === n) {
							s = !0, n = a, r = i;
							break;
						}
						if (l === r) {
							s = !0, r = a, n = i;
							break;
						}
						l = l.sibling;
					}
					if (!s) throw Error(o(189));
				}
			}
			if (n.alternate !== r) throw Error(o(190));
		}
		if (n.tag !== 3) throw Error(o(188));
		return n.stateNode.current === n ? e : t;
	}
	function p(e) {
		var t = e.tag;
		if (t === 5 || t === 26 || t === 27 || t === 6) return e;
		for (e = e.child; e !== null;) {
			if (t = p(e), t !== null) return t;
			e = e.sibling;
		}
		return null;
	}
	var m = Object.assign, h = Symbol.for("react.element"), g = Symbol.for("react.transitional.element"), _ = Symbol.for("react.portal"), v = Symbol.for("react.fragment"), y = Symbol.for("react.strict_mode"), b = Symbol.for("react.profiler"), x = Symbol.for("react.consumer"), S = Symbol.for("react.context"), C = Symbol.for("react.forward_ref"), w = Symbol.for("react.suspense"), T = Symbol.for("react.suspense_list"), E = Symbol.for("react.memo"), D = Symbol.for("react.lazy"), O = Symbol.for("react.activity"), ee = Symbol.for("react.memo_cache_sentinel"), te = Symbol.iterator;
	function ne(e) {
		return typeof e != "object" || !e ? null : (e = te && e[te] || e["@@iterator"], typeof e == "function" ? e : null);
	}
	var re = Symbol.for("react.client.reference");
	function ie(e) {
		if (e == null) return null;
		if (typeof e == "function") return e.$$typeof === re ? null : e.displayName || e.name || null;
		if (typeof e == "string") return e;
		switch (e) {
			case v: return "Fragment";
			case b: return "Profiler";
			case y: return "StrictMode";
			case w: return "Suspense";
			case T: return "SuspenseList";
			case O: return "Activity";
		}
		if (typeof e == "object") switch (e.$$typeof) {
			case _: return "Portal";
			case S: return e.displayName || "Context";
			case x: return (e._context.displayName || "Context") + ".Consumer";
			case C:
				var t = e.render;
				return e = e.displayName, e ||= (e = t.displayName || t.name || "", e === "" ? "ForwardRef" : "ForwardRef(" + e + ")"), e;
			case E: return t = e.displayName || null, t === null ? ie(e.type) || "Memo" : t;
			case D:
				t = e._payload, e = e._init;
				try {
					return ie(e(t));
				} catch {}
		}
		return null;
	}
	var k = Array.isArray, A = r.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE, j = a.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE, ae = {
		pending: !1,
		data: null,
		method: null,
		action: null
	}, oe = [], M = -1;
	function se(e) {
		return { current: e };
	}
	function ce(e) {
		0 > M || (e.current = oe[M], oe[M] = null, M--);
	}
	function N(e, t) {
		M++, oe[M] = e.current, e.current = t;
	}
	var le = se(null), ue = se(null), de = se(null), fe = se(null);
	function pe(e, t) {
		switch (N(de, t), N(ue, e), N(le, null), t.nodeType) {
			case 9:
			case 11:
				e = (e = t.documentElement) && (e = e.namespaceURI) ? Vd(e) : 0;
				break;
			default: if (e = t.tagName, t = t.namespaceURI) t = Vd(t), e = Hd(t, e);
			else switch (e) {
				case "svg":
					e = 1;
					break;
				case "math":
					e = 2;
					break;
				default: e = 0;
			}
		}
		ce(le), N(le, e);
	}
	function me() {
		ce(le), ce(ue), ce(de);
	}
	function he(e) {
		e.memoizedState !== null && N(fe, e);
		var t = le.current, n = Hd(t, e.type);
		t !== n && (N(ue, e), N(le, n));
	}
	function ge(e) {
		ue.current === e && (ce(le), ce(ue)), fe.current === e && (ce(fe), Qf._currentValue = ae);
	}
	var _e, ve;
	function ye(e) {
		if (_e === void 0) try {
			throw Error();
		} catch (e) {
			var t = e.stack.trim().match(/\n( *(at )?)/);
			_e = t && t[1] || "", ve = -1 < e.stack.indexOf("\n    at") ? " (<anonymous>)" : -1 < e.stack.indexOf("@") ? "@unknown:0:0" : "";
		}
		return "\n" + _e + e + ve;
	}
	var P = !1;
	function be(e, t) {
		if (!e || P) return "";
		P = !0;
		var n = Error.prepareStackTrace;
		Error.prepareStackTrace = void 0;
		try {
			var r = { DetermineComponentFrameRoot: function() {
				try {
					if (t) {
						var n = function() {
							throw Error();
						};
						if (Object.defineProperty(n.prototype, "props", { set: function() {
							throw Error();
						} }), typeof Reflect == "object" && Reflect.construct) {
							try {
								Reflect.construct(n, []);
							} catch (e) {
								var r = e;
							}
							Reflect.construct(e, [], n);
						} else {
							try {
								n.call();
							} catch (e) {
								r = e;
							}
							e.call(n.prototype);
						}
					} else {
						try {
							throw Error();
						} catch (e) {
							r = e;
						}
						(n = e()) && typeof n.catch == "function" && n.catch(function() {});
					}
				} catch (e) {
					if (e && r && typeof e.stack == "string") return [e.stack, r.stack];
				}
				return [null, null];
			} };
			r.DetermineComponentFrameRoot.displayName = "DetermineComponentFrameRoot";
			var i = Object.getOwnPropertyDescriptor(r.DetermineComponentFrameRoot, "name");
			i && i.configurable && Object.defineProperty(r.DetermineComponentFrameRoot, "name", { value: "DetermineComponentFrameRoot" });
			var a = r.DetermineComponentFrameRoot(), o = a[0], s = a[1];
			if (o && s) {
				var c = o.split("\n"), l = s.split("\n");
				for (i = r = 0; r < c.length && !c[r].includes("DetermineComponentFrameRoot");) r++;
				for (; i < l.length && !l[i].includes("DetermineComponentFrameRoot");) i++;
				if (r === c.length || i === l.length) for (r = c.length - 1, i = l.length - 1; 1 <= r && 0 <= i && c[r] !== l[i];) i--;
				for (; 1 <= r && 0 <= i; r--, i--) if (c[r] !== l[i]) {
					if (r !== 1 || i !== 1) do
						if (r--, i--, 0 > i || c[r] !== l[i]) {
							var u = "\n" + c[r].replace(" at new ", " at ");
							return e.displayName && u.includes("<anonymous>") && (u = u.replace("<anonymous>", e.displayName)), u;
						}
					while (1 <= r && 0 <= i);
					break;
				}
			}
		} finally {
			P = !1, Error.prepareStackTrace = n;
		}
		return (n = e ? e.displayName || e.name : "") ? ye(n) : "";
	}
	function xe(e, t) {
		switch (e.tag) {
			case 26:
			case 27:
			case 5: return ye(e.type);
			case 16: return ye("Lazy");
			case 13: return e.child !== t && t !== null ? ye("Suspense Fallback") : ye("Suspense");
			case 19: return ye("SuspenseList");
			case 0:
			case 15: return be(e.type, !1);
			case 11: return be(e.type.render, !1);
			case 1: return be(e.type, !0);
			case 31: return ye("Activity");
			default: return "";
		}
	}
	function Se(e) {
		try {
			var t = "", n = null;
			do
				t += xe(e, n), n = e, e = e.return;
			while (e);
			return t;
		} catch (e) {
			return "\nError generating stack: " + e.message + "\n" + e.stack;
		}
	}
	var Ce = Object.prototype.hasOwnProperty, we = t.unstable_scheduleCallback, F = t.unstable_cancelCallback, Te = t.unstable_shouldYield, Ee = t.unstable_requestPaint, De = t.unstable_now, Oe = t.unstable_getCurrentPriorityLevel, ke = t.unstable_ImmediatePriority, Ae = t.unstable_UserBlockingPriority, je = t.unstable_NormalPriority, Me = t.unstable_LowPriority, Ne = t.unstable_IdlePriority, Pe = t.log, Fe = t.unstable_setDisableYieldValue, Ie = null, Le = null;
	function Re(e) {
		if (typeof Pe == "function" && Fe(e), Le && typeof Le.setStrictMode == "function") try {
			Le.setStrictMode(Ie, e);
		} catch {}
	}
	var ze = Math.clz32 ? Math.clz32 : He, Be = Math.log, Ve = Math.LN2;
	function He(e) {
		return e >>>= 0, e === 0 ? 32 : 31 - (Be(e) / Ve | 0) | 0;
	}
	var Ue = 256, We = 262144, Ge = 4194304;
	function Ke(e) {
		var t = e & 42;
		if (t !== 0) return t;
		switch (e & -e) {
			case 1: return 1;
			case 2: return 2;
			case 4: return 4;
			case 8: return 8;
			case 16: return 16;
			case 32: return 32;
			case 64: return 64;
			case 128: return 128;
			case 256:
			case 512:
			case 1024:
			case 2048:
			case 4096:
			case 8192:
			case 16384:
			case 32768:
			case 65536:
			case 131072: return e & 261888;
			case 262144:
			case 524288:
			case 1048576:
			case 2097152: return e & 3932160;
			case 4194304:
			case 8388608:
			case 16777216:
			case 33554432: return e & 62914560;
			case 67108864: return 67108864;
			case 134217728: return 134217728;
			case 268435456: return 268435456;
			case 536870912: return 536870912;
			case 1073741824: return 0;
			default: return e;
		}
	}
	function qe(e, t, n) {
		var r = e.pendingLanes;
		if (r === 0) return 0;
		var i = 0, a = e.suspendedLanes, o = e.pingedLanes;
		e = e.warmLanes;
		var s = r & 134217727;
		return s === 0 ? (s = r & ~a, s === 0 ? o === 0 ? n || (n = r & ~e, n !== 0 && (i = Ke(n))) : i = Ke(o) : i = Ke(s)) : (r = s & ~a, r === 0 ? (o &= s, o === 0 ? n || (n = s & ~e, n !== 0 && (i = Ke(n))) : i = Ke(o)) : i = Ke(r)), i === 0 ? 0 : t !== 0 && t !== i && (t & a) === 0 && (a = i & -i, n = t & -t, a >= n || a === 32 && n & 4194048) ? t : i;
	}
	function Je(e, t) {
		return (e.pendingLanes & ~(e.suspendedLanes & ~e.pingedLanes) & t) === 0;
	}
	function Ye(e, t) {
		switch (e) {
			case 1:
			case 2:
			case 4:
			case 8:
			case 64: return t + 250;
			case 16:
			case 32:
			case 128:
			case 256:
			case 512:
			case 1024:
			case 2048:
			case 4096:
			case 8192:
			case 16384:
			case 32768:
			case 65536:
			case 131072:
			case 262144:
			case 524288:
			case 1048576:
			case 2097152: return t + 5e3;
			case 4194304:
			case 8388608:
			case 16777216:
			case 33554432: return -1;
			case 67108864:
			case 134217728:
			case 268435456:
			case 536870912:
			case 1073741824: return -1;
			default: return -1;
		}
	}
	function Xe() {
		var e = Ge;
		return Ge <<= 1, !(Ge & 62914560) && (Ge = 4194304), e;
	}
	function Ze(e) {
		for (var t = [], n = 0; 31 > n; n++) t.push(e);
		return t;
	}
	function Qe(e, t) {
		e.pendingLanes |= t, t !== 268435456 && (e.suspendedLanes = 0, e.pingedLanes = 0, e.warmLanes = 0);
	}
	function $e(e, t, n, r, i, a) {
		var o = e.pendingLanes;
		e.pendingLanes = n, e.suspendedLanes = 0, e.pingedLanes = 0, e.warmLanes = 0, e.expiredLanes &= n, e.entangledLanes &= n, e.errorRecoveryDisabledLanes &= n, e.shellSuspendCounter = 0;
		var s = e.entanglements, c = e.expirationTimes, l = e.hiddenUpdates;
		for (n = o & ~n; 0 < n;) {
			var u = 31 - ze(n), d = 1 << u;
			s[u] = 0, c[u] = -1;
			var f = l[u];
			if (f !== null) for (l[u] = null, u = 0; u < f.length; u++) {
				var p = f[u];
				p !== null && (p.lane &= -536870913);
			}
			n &= ~d;
		}
		r !== 0 && et(e, r, 0), a !== 0 && i === 0 && e.tag !== 0 && (e.suspendedLanes |= a & ~(o & ~t));
	}
	function et(e, t, n) {
		e.pendingLanes |= t, e.suspendedLanes &= ~t;
		var r = 31 - ze(t);
		e.entangledLanes |= t, e.entanglements[r] = e.entanglements[r] | 1073741824 | n & 261930;
	}
	function tt(e, t) {
		var n = e.entangledLanes |= t;
		for (e = e.entanglements; n;) {
			var r = 31 - ze(n), i = 1 << r;
			i & t | e[r] & t && (e[r] |= t), n &= ~i;
		}
	}
	function nt(e, t) {
		var n = t & -t;
		return n = n & 42 ? 1 : rt(n), (n & (e.suspendedLanes | t)) === 0 ? n : 0;
	}
	function rt(e) {
		switch (e) {
			case 2:
				e = 1;
				break;
			case 8:
				e = 4;
				break;
			case 32:
				e = 16;
				break;
			case 256:
			case 512:
			case 1024:
			case 2048:
			case 4096:
			case 8192:
			case 16384:
			case 32768:
			case 65536:
			case 131072:
			case 262144:
			case 524288:
			case 1048576:
			case 2097152:
			case 4194304:
			case 8388608:
			case 16777216:
			case 33554432:
				e = 128;
				break;
			case 268435456:
				e = 134217728;
				break;
			default: e = 0;
		}
		return e;
	}
	function it(e) {
		return e &= -e, 2 < e ? 8 < e ? e & 134217727 ? 32 : 268435456 : 8 : 2;
	}
	function at() {
		var e = j.p;
		return e === 0 ? (e = window.event, e === void 0 ? 32 : mp(e.type)) : e;
	}
	function ot(e, t) {
		var n = j.p;
		try {
			return j.p = e, t();
		} finally {
			j.p = n;
		}
	}
	var st = Math.random().toString(36).slice(2), ct = "__reactFiber$" + st, lt = "__reactProps$" + st, ut = "__reactContainer$" + st, dt = "__reactEvents$" + st, ft = "__reactListeners$" + st, pt = "__reactHandles$" + st, mt = "__reactResources$" + st, ht = "__reactMarker$" + st;
	function gt(e) {
		delete e[ct], delete e[lt], delete e[dt], delete e[ft], delete e[pt];
	}
	function _t(e) {
		var t = e[ct];
		if (t) return t;
		for (var n = e.parentNode; n;) {
			if (t = n[ut] || n[ct]) {
				if (n = t.alternate, t.child !== null || n !== null && n.child !== null) for (e = df(e); e !== null;) {
					if (n = e[ct]) return n;
					e = df(e);
				}
				return t;
			}
			e = n, n = e.parentNode;
		}
		return null;
	}
	function vt(e) {
		if (e = e[ct] || e[ut]) {
			var t = e.tag;
			if (t === 5 || t === 6 || t === 13 || t === 31 || t === 26 || t === 27 || t === 3) return e;
		}
		return null;
	}
	function yt(e) {
		var t = e.tag;
		if (t === 5 || t === 26 || t === 27 || t === 6) return e.stateNode;
		throw Error(o(33));
	}
	function bt(e) {
		var t = e[mt];
		return t ||= e[mt] = {
			hoistableStyles: /* @__PURE__ */ new Map(),
			hoistableScripts: /* @__PURE__ */ new Map()
		}, t;
	}
	function xt(e) {
		e[ht] = !0;
	}
	var St = /* @__PURE__ */ new Set(), Ct = {};
	function wt(e, t) {
		Tt(e, t), Tt(e + "Capture", t);
	}
	function Tt(e, t) {
		for (Ct[e] = t, e = 0; e < t.length; e++) St.add(t[e]);
	}
	var Et = RegExp("^[:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD][:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD\\-.0-9\\u00B7\\u0300-\\u036F\\u203F-\\u2040]*$"), Dt = {}, Ot = {};
	function kt(e) {
		return Ce.call(Ot, e) ? !0 : Ce.call(Dt, e) ? !1 : Et.test(e) ? Ot[e] = !0 : (Dt[e] = !0, !1);
	}
	function At(e, t, n) {
		if (kt(t)) {
			if (n === null) e.removeAttribute(t);
			else {
				switch (typeof n) {
					case "undefined":
					case "function":
					case "symbol":
						e.removeAttribute(t);
						return;
					case "boolean":
						var r = t.toLowerCase().slice(0, 5);
						if (r !== "data-" && r !== "aria-") {
							e.removeAttribute(t);
							return;
						}
				}
				e.setAttribute(t, "" + n);
			}
		}
	}
	function jt(e, t, n) {
		if (n === null) e.removeAttribute(t);
		else {
			switch (typeof n) {
				case "undefined":
				case "function":
				case "symbol":
				case "boolean":
					e.removeAttribute(t);
					return;
			}
			e.setAttribute(t, "" + n);
		}
	}
	function Mt(e, t, n, r) {
		if (r === null) e.removeAttribute(n);
		else {
			switch (typeof r) {
				case "undefined":
				case "function":
				case "symbol":
				case "boolean":
					e.removeAttribute(n);
					return;
			}
			e.setAttributeNS(t, n, "" + r);
		}
	}
	function Nt(e) {
		switch (typeof e) {
			case "bigint":
			case "boolean":
			case "number":
			case "string":
			case "undefined": return e;
			case "object": return e;
			default: return "";
		}
	}
	function Pt(e) {
		var t = e.type;
		return (e = e.nodeName) && e.toLowerCase() === "input" && (t === "checkbox" || t === "radio");
	}
	function Ft(e, t, n) {
		var r = Object.getOwnPropertyDescriptor(e.constructor.prototype, t);
		if (!e.hasOwnProperty(t) && r !== void 0 && typeof r.get == "function" && typeof r.set == "function") {
			var i = r.get, a = r.set;
			return Object.defineProperty(e, t, {
				configurable: !0,
				get: function() {
					return i.call(this);
				},
				set: function(e) {
					n = "" + e, a.call(this, e);
				}
			}), Object.defineProperty(e, t, { enumerable: r.enumerable }), {
				getValue: function() {
					return n;
				},
				setValue: function(e) {
					n = "" + e;
				},
				stopTracking: function() {
					e._valueTracker = null, delete e[t];
				}
			};
		}
	}
	function It(e) {
		if (!e._valueTracker) {
			var t = Pt(e) ? "checked" : "value";
			e._valueTracker = Ft(e, t, "" + e[t]);
		}
	}
	function Lt(e) {
		if (!e) return !1;
		var t = e._valueTracker;
		if (!t) return !0;
		var n = t.getValue(), r = "";
		return e && (r = Pt(e) ? e.checked ? "true" : "false" : e.value), e = r, e !== n && (t.setValue(e), !0);
	}
	function Rt(e) {
		if (e ||= typeof document < "u" ? document : void 0, e === void 0) return null;
		try {
			return e.activeElement || e.body;
		} catch {
			return e.body;
		}
	}
	var zt = /[\n"\\]/g;
	function Bt(e) {
		return e.replace(zt, function(e) {
			return "\\" + e.charCodeAt(0).toString(16) + " ";
		});
	}
	function Vt(e, t, n, r, i, a, o, s) {
		e.name = "", o != null && typeof o != "function" && typeof o != "symbol" && typeof o != "boolean" ? e.type = o : e.removeAttribute("type"), t == null ? o !== "submit" && o !== "reset" || e.removeAttribute("value") : o === "number" ? (t === 0 && e.value === "" || e.value != t) && (e.value = "" + Nt(t)) : e.value !== "" + Nt(t) && (e.value = "" + Nt(t)), t == null ? n == null ? r != null && e.removeAttribute("value") : Ut(e, o, Nt(n)) : Ut(e, o, Nt(t)), i == null && a != null && (e.defaultChecked = !!a), i != null && (e.checked = i && typeof i != "function" && typeof i != "symbol"), s != null && typeof s != "function" && typeof s != "symbol" && typeof s != "boolean" ? e.name = "" + Nt(s) : e.removeAttribute("name");
	}
	function Ht(e, t, n, r, i, a, o, s) {
		if (a != null && typeof a != "function" && typeof a != "symbol" && typeof a != "boolean" && (e.type = a), t != null || n != null) {
			if (!(a !== "submit" && a !== "reset" || t != null)) {
				It(e);
				return;
			}
			n = n == null ? "" : "" + Nt(n), t = t == null ? n : "" + Nt(t), s || t === e.value || (e.value = t), e.defaultValue = t;
		}
		r ??= i, r = typeof r != "function" && typeof r != "symbol" && !!r, e.checked = s ? e.checked : !!r, e.defaultChecked = !!r, o != null && typeof o != "function" && typeof o != "symbol" && typeof o != "boolean" && (e.name = o), It(e);
	}
	function Ut(e, t, n) {
		t === "number" && Rt(e.ownerDocument) === e || e.defaultValue === "" + n || (e.defaultValue = "" + n);
	}
	function Wt(e, t, n, r) {
		if (e = e.options, t) {
			t = {};
			for (var i = 0; i < n.length; i++) t["$" + n[i]] = !0;
			for (n = 0; n < e.length; n++) i = t.hasOwnProperty("$" + e[n].value), e[n].selected !== i && (e[n].selected = i), i && r && (e[n].defaultSelected = !0);
		} else {
			for (n = "" + Nt(n), t = null, i = 0; i < e.length; i++) {
				if (e[i].value === n) {
					e[i].selected = !0, r && (e[i].defaultSelected = !0);
					return;
				}
				t !== null || e[i].disabled || (t = e[i]);
			}
			t !== null && (t.selected = !0);
		}
	}
	function Gt(e, t, n) {
		t != null && (t = "" + Nt(t), t !== e.value && (e.value = t), n == null) ? e.defaultValue !== t && (e.defaultValue = t) : e.defaultValue = n == null ? "" : "" + Nt(n);
	}
	function qt(e, t, n, r) {
		if (t == null) {
			if (r != null) {
				if (n != null) throw Error(o(92));
				if (k(r)) {
					if (1 < r.length) throw Error(o(93));
					r = r[0];
				}
				n = r;
			}
			n ??= "", t = n;
		}
		n = Nt(t), e.defaultValue = n, r = e.textContent, r === n && r !== "" && r !== null && (e.value = r), It(e);
	}
	function Jt(e, t) {
		if (t) {
			var n = e.firstChild;
			if (n && n === e.lastChild && n.nodeType === 3) {
				n.nodeValue = t;
				return;
			}
		}
		e.textContent = t;
	}
	var Yt = new Set("animationIterationCount aspectRatio borderImageOutset borderImageSlice borderImageWidth boxFlex boxFlexGroup boxOrdinalGroup columnCount columns flex flexGrow flexPositive flexShrink flexNegative flexOrder gridArea gridRow gridRowEnd gridRowSpan gridRowStart gridColumn gridColumnEnd gridColumnSpan gridColumnStart fontWeight lineClamp lineHeight opacity order orphans scale tabSize widows zIndex zoom fillOpacity floodOpacity stopOpacity strokeDasharray strokeDashoffset strokeMiterlimit strokeOpacity strokeWidth MozAnimationIterationCount MozBoxFlex MozBoxFlexGroup MozLineClamp msAnimationIterationCount msFlex msZoom msFlexGrow msFlexNegative msFlexOrder msFlexPositive msFlexShrink msGridColumn msGridColumnSpan msGridRow msGridRowSpan WebkitAnimationIterationCount WebkitBoxFlex WebKitBoxFlexGroup WebkitBoxOrdinalGroup WebkitColumnCount WebkitColumns WebkitFlex WebkitFlexGrow WebkitFlexPositive WebkitFlexShrink WebkitLineClamp".split(" "));
	function I(e, t, n) {
		var r = t.indexOf("--") === 0;
		n == null || typeof n == "boolean" || n === "" ? r ? e.setProperty(t, "") : t === "float" ? e.cssFloat = "" : e[t] = "" : r ? e.setProperty(t, n) : typeof n != "number" || n === 0 || Yt.has(t) ? t === "float" ? e.cssFloat = n : e[t] = ("" + n).trim() : e[t] = n + "px";
	}
	function L(e, t, n) {
		if (t != null && typeof t != "object") throw Error(o(62));
		if (e = e.style, n != null) {
			for (var r in n) !n.hasOwnProperty(r) || t != null && t.hasOwnProperty(r) || (r.indexOf("--") === 0 ? e.setProperty(r, "") : r === "float" ? e.cssFloat = "" : e[r] = "");
			for (var i in t) r = t[i], t.hasOwnProperty(i) && n[i] !== r && I(e, i, r);
		} else for (var a in t) t.hasOwnProperty(a) && I(e, a, t[a]);
	}
	function Xt(e) {
		if (e.indexOf("-") === -1) return !1;
		switch (e) {
			case "annotation-xml":
			case "color-profile":
			case "font-face":
			case "font-face-src":
			case "font-face-uri":
			case "font-face-format":
			case "font-face-name":
			case "missing-glyph": return !1;
			default: return !0;
		}
	}
	var Zt = /* @__PURE__ */ new Map([
		["acceptCharset", "accept-charset"],
		["htmlFor", "for"],
		["httpEquiv", "http-equiv"],
		["crossOrigin", "crossorigin"],
		["accentHeight", "accent-height"],
		["alignmentBaseline", "alignment-baseline"],
		["arabicForm", "arabic-form"],
		["baselineShift", "baseline-shift"],
		["capHeight", "cap-height"],
		["clipPath", "clip-path"],
		["clipRule", "clip-rule"],
		["colorInterpolation", "color-interpolation"],
		["colorInterpolationFilters", "color-interpolation-filters"],
		["colorProfile", "color-profile"],
		["colorRendering", "color-rendering"],
		["dominantBaseline", "dominant-baseline"],
		["enableBackground", "enable-background"],
		["fillOpacity", "fill-opacity"],
		["fillRule", "fill-rule"],
		["floodColor", "flood-color"],
		["floodOpacity", "flood-opacity"],
		["fontFamily", "font-family"],
		["fontSize", "font-size"],
		["fontSizeAdjust", "font-size-adjust"],
		["fontStretch", "font-stretch"],
		["fontStyle", "font-style"],
		["fontVariant", "font-variant"],
		["fontWeight", "font-weight"],
		["glyphName", "glyph-name"],
		["glyphOrientationHorizontal", "glyph-orientation-horizontal"],
		["glyphOrientationVertical", "glyph-orientation-vertical"],
		["horizAdvX", "horiz-adv-x"],
		["horizOriginX", "horiz-origin-x"],
		["imageRendering", "image-rendering"],
		["letterSpacing", "letter-spacing"],
		["lightingColor", "lighting-color"],
		["markerEnd", "marker-end"],
		["markerMid", "marker-mid"],
		["markerStart", "marker-start"],
		["overlinePosition", "overline-position"],
		["overlineThickness", "overline-thickness"],
		["paintOrder", "paint-order"],
		["panose-1", "panose-1"],
		["pointerEvents", "pointer-events"],
		["renderingIntent", "rendering-intent"],
		["shapeRendering", "shape-rendering"],
		["stopColor", "stop-color"],
		["stopOpacity", "stop-opacity"],
		["strikethroughPosition", "strikethrough-position"],
		["strikethroughThickness", "strikethrough-thickness"],
		["strokeDasharray", "stroke-dasharray"],
		["strokeDashoffset", "stroke-dashoffset"],
		["strokeLinecap", "stroke-linecap"],
		["strokeLinejoin", "stroke-linejoin"],
		["strokeMiterlimit", "stroke-miterlimit"],
		["strokeOpacity", "stroke-opacity"],
		["strokeWidth", "stroke-width"],
		["textAnchor", "text-anchor"],
		["textDecoration", "text-decoration"],
		["textRendering", "text-rendering"],
		["transformOrigin", "transform-origin"],
		["underlinePosition", "underline-position"],
		["underlineThickness", "underline-thickness"],
		["unicodeBidi", "unicode-bidi"],
		["unicodeRange", "unicode-range"],
		["unitsPerEm", "units-per-em"],
		["vAlphabetic", "v-alphabetic"],
		["vHanging", "v-hanging"],
		["vIdeographic", "v-ideographic"],
		["vMathematical", "v-mathematical"],
		["vectorEffect", "vector-effect"],
		["vertAdvY", "vert-adv-y"],
		["vertOriginX", "vert-origin-x"],
		["vertOriginY", "vert-origin-y"],
		["wordSpacing", "word-spacing"],
		["writingMode", "writing-mode"],
		["xmlnsXlink", "xmlns:xlink"],
		["xHeight", "x-height"]
	]), Qt = /^[\u0000-\u001F ]*j[\r\n\t]*a[\r\n\t]*v[\r\n\t]*a[\r\n\t]*s[\r\n\t]*c[\r\n\t]*r[\r\n\t]*i[\r\n\t]*p[\r\n\t]*t[\r\n\t]*:/i;
	function $t(e) {
		return Qt.test("" + e) ? "javascript:throw new Error('React has blocked a javascript: URL as a security precaution.')" : e;
	}
	function en() {}
	var tn = null;
	function nn(e) {
		return e = e.target || e.srcElement || window, e.correspondingUseElement && (e = e.correspondingUseElement), e.nodeType === 3 ? e.parentNode : e;
	}
	var rn = null, an = null;
	function on(e) {
		var t = vt(e);
		if (t && (e = t.stateNode)) {
			var n = e[lt] || null;
			a: switch (e = t.stateNode, t.type) {
				case "input":
					if (Vt(e, n.value, n.defaultValue, n.defaultValue, n.checked, n.defaultChecked, n.type, n.name), t = n.name, n.type === "radio" && t != null) {
						for (n = e; n.parentNode;) n = n.parentNode;
						for (n = n.querySelectorAll("input[name=\"" + Bt("" + t) + "\"][type=\"radio\"]"), t = 0; t < n.length; t++) {
							var r = n[t];
							if (r !== e && r.form === e.form) {
								var i = r[lt] || null;
								if (!i) throw Error(o(90));
								Vt(r, i.value, i.defaultValue, i.defaultValue, i.checked, i.defaultChecked, i.type, i.name);
							}
						}
						for (t = 0; t < n.length; t++) r = n[t], r.form === e.form && Lt(r);
					}
					break a;
				case "textarea":
					Gt(e, n.value, n.defaultValue);
					break a;
				case "select": t = n.value, t != null && Wt(e, !!n.multiple, t, !1);
			}
		}
	}
	var sn = !1;
	function cn(e, t, n) {
		if (sn) return e(t, n);
		sn = !0;
		try {
			return e(t);
		} finally {
			if (sn = !1, (rn !== null || an !== null) && (bu(), rn && (t = rn, e = an, an = rn = null, on(t), e))) for (t = 0; t < e.length; t++) on(e[t]);
		}
	}
	function ln(e, t) {
		var n = e.stateNode;
		if (n === null) return null;
		var r = n[lt] || null;
		if (r === null) return null;
		n = r[t];
		a: switch (t) {
			case "onClick":
			case "onClickCapture":
			case "onDoubleClick":
			case "onDoubleClickCapture":
			case "onMouseDown":
			case "onMouseDownCapture":
			case "onMouseMove":
			case "onMouseMoveCapture":
			case "onMouseUp":
			case "onMouseUpCapture":
			case "onMouseEnter":
				(r = !r.disabled) || (e = e.type, r = e !== "button" && e !== "input" && e !== "select" && e !== "textarea"), e = !r;
				break a;
			default: e = !1;
		}
		if (e) return null;
		if (n && typeof n != "function") throw Error(o(231, t, typeof n));
		return n;
	}
	var un = typeof window < "u" && window.document !== void 0 && window.document.createElement !== void 0, dn = !1;
	if (un) try {
		var fn = {};
		Object.defineProperty(fn, "passive", { get: function() {
			dn = !0;
		} }), window.addEventListener("test", fn, fn), window.removeEventListener("test", fn, fn);
	} catch {
		dn = !1;
	}
	var pn = null, mn = null, hn = null;
	function gn() {
		if (hn) return hn;
		var e, t = mn, n = t.length, r, i = "value" in pn ? pn.value : pn.textContent, a = i.length;
		for (e = 0; e < n && t[e] === i[e]; e++);
		var o = n - e;
		for (r = 1; r <= o && t[n - r] === i[a - r]; r++);
		return hn = i.slice(e, 1 < r ? 1 - r : void 0);
	}
	function _n(e) {
		var t = e.keyCode;
		return "charCode" in e ? (e = e.charCode, e === 0 && t === 13 && (e = 13)) : e = t, e === 10 && (e = 13), 32 <= e || e === 13 ? e : 0;
	}
	function vn() {
		return !0;
	}
	function yn() {
		return !1;
	}
	function bn(e) {
		function t(t, n, r, i, a) {
			for (var o in this._reactName = t, this._targetInst = r, this.type = n, this.nativeEvent = i, this.target = a, this.currentTarget = null, e) e.hasOwnProperty(o) && (t = e[o], this[o] = t ? t(i) : i[o]);
			return this.isDefaultPrevented = (i.defaultPrevented == null ? !1 === i.returnValue : i.defaultPrevented) ? vn : yn, this.isPropagationStopped = yn, this;
		}
		return m(t.prototype, {
			preventDefault: function() {
				this.defaultPrevented = !0;
				var e = this.nativeEvent;
				e && (e.preventDefault ? e.preventDefault() : typeof e.returnValue != "unknown" && (e.returnValue = !1), this.isDefaultPrevented = vn);
			},
			stopPropagation: function() {
				var e = this.nativeEvent;
				e && (e.stopPropagation ? e.stopPropagation() : typeof e.cancelBubble != "unknown" && (e.cancelBubble = !0), this.isPropagationStopped = vn);
			},
			persist: function() {},
			isPersistent: vn
		}), t;
	}
	var xn = {
		eventPhase: 0,
		bubbles: 0,
		cancelable: 0,
		timeStamp: function(e) {
			return e.timeStamp || Date.now();
		},
		defaultPrevented: 0,
		isTrusted: 0
	}, Sn = bn(xn), Cn = m({}, xn, {
		view: 0,
		detail: 0
	}), wn = bn(Cn), Tn, En, Dn, On = m({}, Cn, {
		screenX: 0,
		screenY: 0,
		clientX: 0,
		clientY: 0,
		pageX: 0,
		pageY: 0,
		ctrlKey: 0,
		shiftKey: 0,
		altKey: 0,
		metaKey: 0,
		getModifierState: zn,
		button: 0,
		buttons: 0,
		relatedTarget: function(e) {
			return e.relatedTarget === void 0 ? e.fromElement === e.srcElement ? e.toElement : e.fromElement : e.relatedTarget;
		},
		movementX: function(e) {
			return "movementX" in e ? e.movementX : (e !== Dn && (Dn && e.type === "mousemove" ? (Tn = e.screenX - Dn.screenX, En = e.screenY - Dn.screenY) : En = Tn = 0, Dn = e), Tn);
		},
		movementY: function(e) {
			return "movementY" in e ? e.movementY : En;
		}
	}), kn = bn(On), An = bn(m({}, On, { dataTransfer: 0 })), jn = bn(m({}, Cn, { relatedTarget: 0 })), Mn = bn(m({}, xn, {
		animationName: 0,
		elapsedTime: 0,
		pseudoElement: 0
	})), Nn = bn(m({}, xn, { clipboardData: function(e) {
		return "clipboardData" in e ? e.clipboardData : window.clipboardData;
	} })), Pn = bn(m({}, xn, { data: 0 })), Fn = {
		Esc: "Escape",
		Spacebar: " ",
		Left: "ArrowLeft",
		Up: "ArrowUp",
		Right: "ArrowRight",
		Down: "ArrowDown",
		Del: "Delete",
		Win: "OS",
		Menu: "ContextMenu",
		Apps: "ContextMenu",
		Scroll: "ScrollLock",
		MozPrintableKey: "Unidentified"
	}, In = {
		8: "Backspace",
		9: "Tab",
		12: "Clear",
		13: "Enter",
		16: "Shift",
		17: "Control",
		18: "Alt",
		19: "Pause",
		20: "CapsLock",
		27: "Escape",
		32: " ",
		33: "PageUp",
		34: "PageDown",
		35: "End",
		36: "Home",
		37: "ArrowLeft",
		38: "ArrowUp",
		39: "ArrowRight",
		40: "ArrowDown",
		45: "Insert",
		46: "Delete",
		112: "F1",
		113: "F2",
		114: "F3",
		115: "F4",
		116: "F5",
		117: "F6",
		118: "F7",
		119: "F8",
		120: "F9",
		121: "F10",
		122: "F11",
		123: "F12",
		144: "NumLock",
		145: "ScrollLock",
		224: "Meta"
	}, Ln = {
		Alt: "altKey",
		Control: "ctrlKey",
		Meta: "metaKey",
		Shift: "shiftKey"
	};
	function Rn(e) {
		var t = this.nativeEvent;
		return t.getModifierState ? t.getModifierState(e) : (e = Ln[e]) ? !!t[e] : !1;
	}
	function zn() {
		return Rn;
	}
	var Bn = bn(m({}, Cn, {
		key: function(e) {
			if (e.key) {
				var t = Fn[e.key] || e.key;
				if (t !== "Unidentified") return t;
			}
			return e.type === "keypress" ? (e = _n(e), e === 13 ? "Enter" : String.fromCharCode(e)) : e.type === "keydown" || e.type === "keyup" ? In[e.keyCode] || "Unidentified" : "";
		},
		code: 0,
		location: 0,
		ctrlKey: 0,
		shiftKey: 0,
		altKey: 0,
		metaKey: 0,
		repeat: 0,
		locale: 0,
		getModifierState: zn,
		charCode: function(e) {
			return e.type === "keypress" ? _n(e) : 0;
		},
		keyCode: function(e) {
			return e.type === "keydown" || e.type === "keyup" ? e.keyCode : 0;
		},
		which: function(e) {
			return e.type === "keypress" ? _n(e) : e.type === "keydown" || e.type === "keyup" ? e.keyCode : 0;
		}
	})), Vn = bn(m({}, On, {
		pointerId: 0,
		width: 0,
		height: 0,
		pressure: 0,
		tangentialPressure: 0,
		tiltX: 0,
		tiltY: 0,
		twist: 0,
		pointerType: 0,
		isPrimary: 0
	})), Hn = bn(m({}, Cn, {
		touches: 0,
		targetTouches: 0,
		changedTouches: 0,
		altKey: 0,
		metaKey: 0,
		ctrlKey: 0,
		shiftKey: 0,
		getModifierState: zn
	})), Un = bn(m({}, xn, {
		propertyName: 0,
		elapsedTime: 0,
		pseudoElement: 0
	})), Wn = bn(m({}, On, {
		deltaX: function(e) {
			return "deltaX" in e ? e.deltaX : "wheelDeltaX" in e ? -e.wheelDeltaX : 0;
		},
		deltaY: function(e) {
			return "deltaY" in e ? e.deltaY : "wheelDeltaY" in e ? -e.wheelDeltaY : "wheelDelta" in e ? -e.wheelDelta : 0;
		},
		deltaZ: 0,
		deltaMode: 0
	})), Gn = bn(m({}, xn, {
		newState: 0,
		oldState: 0
	})), Kn = [
		9,
		13,
		27,
		32
	], qn = un && "CompositionEvent" in window, Jn = null;
	un && "documentMode" in document && (Jn = document.documentMode);
	var Yn = un && "TextEvent" in window && !Jn, Xn = un && (!qn || Jn && 8 < Jn && 11 >= Jn), Zn = " ", Qn = !1;
	function $n(e, t) {
		switch (e) {
			case "keyup": return Kn.indexOf(t.keyCode) !== -1;
			case "keydown": return t.keyCode !== 229;
			case "keypress":
			case "mousedown":
			case "focusout": return !0;
			default: return !1;
		}
	}
	function er(e) {
		return e = e.detail, typeof e == "object" && "data" in e ? e.data : null;
	}
	var tr = !1;
	function nr(e, t) {
		switch (e) {
			case "compositionend": return er(t);
			case "keypress": return t.which === 32 ? (Qn = !0, Zn) : null;
			case "textInput": return e = t.data, e === Zn && Qn ? null : e;
			default: return null;
		}
	}
	function rr(e, t) {
		if (tr) return e === "compositionend" || !qn && $n(e, t) ? (e = gn(), hn = mn = pn = null, tr = !1, e) : null;
		switch (e) {
			case "paste": return null;
			case "keypress":
				if (!(t.ctrlKey || t.altKey || t.metaKey) || t.ctrlKey && t.altKey) {
					if (t.char && 1 < t.char.length) return t.char;
					if (t.which) return String.fromCharCode(t.which);
				}
				return null;
			case "compositionend": return Xn && t.locale !== "ko" ? null : t.data;
			default: return null;
		}
	}
	var ir = {
		color: !0,
		date: !0,
		datetime: !0,
		"datetime-local": !0,
		email: !0,
		month: !0,
		number: !0,
		password: !0,
		range: !0,
		search: !0,
		tel: !0,
		text: !0,
		time: !0,
		url: !0,
		week: !0
	};
	function ar(e) {
		var t = e && e.nodeName && e.nodeName.toLowerCase();
		return t === "input" ? !!ir[e.type] : t === "textarea";
	}
	function or(e, t, n, r) {
		rn ? an ? an.push(r) : an = [r] : rn = r, t = Ed(t, "onChange"), 0 < t.length && (n = new Sn("onChange", "change", null, n, r), e.push({
			event: n,
			listeners: t
		}));
	}
	var sr = null, cr = null;
	function lr(e) {
		yd(e, 0);
	}
	function ur(e) {
		if (Lt(yt(e))) return e;
	}
	function dr(e, t) {
		if (e === "change") return t;
	}
	var fr = !1;
	if (un) {
		var pr;
		if (un) {
			var mr = "oninput" in document;
			if (!mr) {
				var hr = document.createElement("div");
				hr.setAttribute("oninput", "return;"), mr = typeof hr.oninput == "function";
			}
			pr = mr;
		} else pr = !1;
		fr = pr && (!document.documentMode || 9 < document.documentMode);
	}
	function gr() {
		sr && (sr.detachEvent("onpropertychange", _r), cr = sr = null);
	}
	function _r(e) {
		if (e.propertyName === "value" && ur(cr)) {
			var t = [];
			or(t, cr, e, nn(e)), cn(lr, t);
		}
	}
	function vr(e, t, n) {
		e === "focusin" ? (gr(), sr = t, cr = n, sr.attachEvent("onpropertychange", _r)) : e === "focusout" && gr();
	}
	function yr(e) {
		if (e === "selectionchange" || e === "keyup" || e === "keydown") return ur(cr);
	}
	function br(e, t) {
		if (e === "click") return ur(t);
	}
	function xr(e, t) {
		if (e === "input" || e === "change") return ur(t);
	}
	function Sr(e, t) {
		return e === t && (e !== 0 || 1 / e == 1 / t) || e !== e && t !== t;
	}
	var Cr = typeof Object.is == "function" ? Object.is : Sr;
	function wr(e, t) {
		if (Cr(e, t)) return !0;
		if (typeof e != "object" || !e || typeof t != "object" || !t) return !1;
		var n = Object.keys(e), r = Object.keys(t);
		if (n.length !== r.length) return !1;
		for (r = 0; r < n.length; r++) {
			var i = n[r];
			if (!Ce.call(t, i) || !Cr(e[i], t[i])) return !1;
		}
		return !0;
	}
	function Tr(e) {
		for (; e && e.firstChild;) e = e.firstChild;
		return e;
	}
	function Er(e, t) {
		var n = Tr(e);
		e = 0;
		for (var r; n;) {
			if (n.nodeType === 3) {
				if (r = e + n.textContent.length, e <= t && r >= t) return {
					node: n,
					offset: t - e
				};
				e = r;
			}
			a: {
				for (; n;) {
					if (n.nextSibling) {
						n = n.nextSibling;
						break a;
					}
					n = n.parentNode;
				}
				n = void 0;
			}
			n = Tr(n);
		}
	}
	function Dr(e, t) {
		return e && t ? e === t ? !0 : e && e.nodeType === 3 ? !1 : t && t.nodeType === 3 ? Dr(e, t.parentNode) : "contains" in e ? e.contains(t) : e.compareDocumentPosition ? !!(e.compareDocumentPosition(t) & 16) : !1 : !1;
	}
	function Or(e) {
		e = e != null && e.ownerDocument != null && e.ownerDocument.defaultView != null ? e.ownerDocument.defaultView : window;
		for (var t = Rt(e.document); t instanceof e.HTMLIFrameElement;) {
			try {
				var n = typeof t.contentWindow.location.href == "string";
			} catch {
				n = !1;
			}
			if (n) e = t.contentWindow;
			else break;
			t = Rt(e.document);
		}
		return t;
	}
	function kr(e) {
		var t = e && e.nodeName && e.nodeName.toLowerCase();
		return t && (t === "input" && (e.type === "text" || e.type === "search" || e.type === "tel" || e.type === "url" || e.type === "password") || t === "textarea" || e.contentEditable === "true");
	}
	var Ar = un && "documentMode" in document && 11 >= document.documentMode, jr = null, Mr = null, Nr = null, Pr = !1;
	function Fr(e, t, n) {
		var r = n.window === n ? n.document : n.nodeType === 9 ? n : n.ownerDocument;
		Pr || jr == null || jr !== Rt(r) || (r = jr, "selectionStart" in r && kr(r) ? r = {
			start: r.selectionStart,
			end: r.selectionEnd
		} : (r = (r.ownerDocument && r.ownerDocument.defaultView || window).getSelection(), r = {
			anchorNode: r.anchorNode,
			anchorOffset: r.anchorOffset,
			focusNode: r.focusNode,
			focusOffset: r.focusOffset
		}), Nr && wr(Nr, r) || (Nr = r, r = Ed(Mr, "onSelect"), 0 < r.length && (t = new Sn("onSelect", "select", null, t, n), e.push({
			event: t,
			listeners: r
		}), t.target = jr)));
	}
	function Ir(e, t) {
		var n = {};
		return n[e.toLowerCase()] = t.toLowerCase(), n["Webkit" + e] = "webkit" + t, n["Moz" + e] = "moz" + t, n;
	}
	var Lr = {
		animationend: Ir("Animation", "AnimationEnd"),
		animationiteration: Ir("Animation", "AnimationIteration"),
		animationstart: Ir("Animation", "AnimationStart"),
		transitionrun: Ir("Transition", "TransitionRun"),
		transitionstart: Ir("Transition", "TransitionStart"),
		transitioncancel: Ir("Transition", "TransitionCancel"),
		transitionend: Ir("Transition", "TransitionEnd")
	}, Rr = {}, zr = {};
	un && (zr = document.createElement("div").style, "AnimationEvent" in window || (delete Lr.animationend.animation, delete Lr.animationiteration.animation, delete Lr.animationstart.animation), "TransitionEvent" in window || delete Lr.transitionend.transition);
	function Br(e) {
		if (Rr[e]) return Rr[e];
		if (!Lr[e]) return e;
		var t = Lr[e], n;
		for (n in t) if (t.hasOwnProperty(n) && n in zr) return Rr[e] = t[n];
		return e;
	}
	var Vr = Br("animationend"), Hr = Br("animationiteration"), Ur = Br("animationstart"), Wr = Br("transitionrun"), Gr = Br("transitionstart"), Kr = Br("transitioncancel"), qr = Br("transitionend"), Jr = /* @__PURE__ */ new Map(), Yr = "abort auxClick beforeToggle cancel canPlay canPlayThrough click close contextMenu copy cut drag dragEnd dragEnter dragExit dragLeave dragOver dragStart drop durationChange emptied encrypted ended error gotPointerCapture input invalid keyDown keyPress keyUp load loadedData loadedMetadata loadStart lostPointerCapture mouseDown mouseMove mouseOut mouseOver mouseUp paste pause play playing pointerCancel pointerDown pointerMove pointerOut pointerOver pointerUp progress rateChange reset resize seeked seeking stalled submit suspend timeUpdate touchCancel touchEnd touchStart volumeChange scroll toggle touchMove waiting wheel".split(" ");
	Yr.push("scrollEnd");
	function Xr(e, t) {
		Jr.set(e, t), wt(t, [e]);
	}
	var Zr = typeof reportError == "function" ? reportError : function(e) {
		if (typeof window == "object" && typeof window.ErrorEvent == "function") {
			var t = new window.ErrorEvent("error", {
				bubbles: !0,
				cancelable: !0,
				message: typeof e == "object" && e && typeof e.message == "string" ? String(e.message) : String(e),
				error: e
			});
			if (!window.dispatchEvent(t)) return;
		} else if (typeof process == "object" && typeof process.emit == "function") {
			process.emit("uncaughtException", e);
			return;
		}
		console.error(e);
	}, Qr = [], $r = 0, ei = 0;
	function ti() {
		for (var e = $r, t = ei = $r = 0; t < e;) {
			var n = Qr[t];
			Qr[t++] = null;
			var r = Qr[t];
			Qr[t++] = null;
			var i = Qr[t];
			Qr[t++] = null;
			var a = Qr[t];
			if (Qr[t++] = null, r !== null && i !== null) {
				var o = r.pending;
				o === null ? i.next = i : (i.next = o.next, o.next = i), r.pending = i;
			}
			a !== 0 && ai(n, i, a);
		}
	}
	function ni(e, t, n, r) {
		Qr[$r++] = e, Qr[$r++] = t, Qr[$r++] = n, Qr[$r++] = r, ei |= r, e.lanes |= r, e = e.alternate, e !== null && (e.lanes |= r);
	}
	function ri(e, t, n, r) {
		return ni(e, t, n, r), oi(e);
	}
	function ii(e, t) {
		return ni(e, null, null, t), oi(e);
	}
	function ai(e, t, n) {
		e.lanes |= n;
		var r = e.alternate;
		r !== null && (r.lanes |= n);
		for (var i = !1, a = e.return; a !== null;) a.childLanes |= n, r = a.alternate, r !== null && (r.childLanes |= n), a.tag === 22 && (e = a.stateNode, e === null || e._visibility & 1 || (i = !0)), e = a, a = a.return;
		return e.tag === 3 ? (a = e.stateNode, i && t !== null && (i = 31 - ze(n), e = a.hiddenUpdates, r = e[i], r === null ? e[i] = [t] : r.push(t), t.lane = n | 536870912), a) : null;
	}
	function oi(e) {
		if (50 < du) throw du = 0, fu = null, Error(o(185));
		for (var t = e.return; t !== null;) e = t, t = e.return;
		return e.tag === 3 ? e.stateNode : null;
	}
	var si = {};
	function ci(e, t, n, r) {
		this.tag = e, this.key = n, this.sibling = this.child = this.return = this.stateNode = this.type = this.elementType = null, this.index = 0, this.refCleanup = this.ref = null, this.pendingProps = t, this.dependencies = this.memoizedState = this.updateQueue = this.memoizedProps = null, this.mode = r, this.subtreeFlags = this.flags = 0, this.deletions = null, this.childLanes = this.lanes = 0, this.alternate = null;
	}
	function li(e, t, n, r) {
		return new ci(e, t, n, r);
	}
	function ui(e) {
		return e = e.prototype, !(!e || !e.isReactComponent);
	}
	function di(e, t) {
		var n = e.alternate;
		return n === null ? (n = li(e.tag, t, e.key, e.mode), n.elementType = e.elementType, n.type = e.type, n.stateNode = e.stateNode, n.alternate = e, e.alternate = n) : (n.pendingProps = t, n.type = e.type, n.flags = 0, n.subtreeFlags = 0, n.deletions = null), n.flags = e.flags & 65011712, n.childLanes = e.childLanes, n.lanes = e.lanes, n.child = e.child, n.memoizedProps = e.memoizedProps, n.memoizedState = e.memoizedState, n.updateQueue = e.updateQueue, t = e.dependencies, n.dependencies = t === null ? null : {
			lanes: t.lanes,
			firstContext: t.firstContext
		}, n.sibling = e.sibling, n.index = e.index, n.ref = e.ref, n.refCleanup = e.refCleanup, n;
	}
	function fi(e, t) {
		e.flags &= 65011714;
		var n = e.alternate;
		return n === null ? (e.childLanes = 0, e.lanes = t, e.child = null, e.subtreeFlags = 0, e.memoizedProps = null, e.memoizedState = null, e.updateQueue = null, e.dependencies = null, e.stateNode = null) : (e.childLanes = n.childLanes, e.lanes = n.lanes, e.child = n.child, e.subtreeFlags = 0, e.deletions = null, e.memoizedProps = n.memoizedProps, e.memoizedState = n.memoizedState, e.updateQueue = n.updateQueue, e.type = n.type, t = n.dependencies, e.dependencies = t === null ? null : {
			lanes: t.lanes,
			firstContext: t.firstContext
		}), e;
	}
	function pi(e, t, n, r, i, a) {
		var s = 0;
		if (r = e, typeof e == "function") ui(e) && (s = 1);
		else if (typeof e == "string") s = Uf(e, n, le.current) ? 26 : e === "html" || e === "head" || e === "body" ? 27 : 5;
		else a: switch (e) {
			case O: return e = li(31, n, t, i), e.elementType = O, e.lanes = a, e;
			case v: return mi(n.children, i, a, t);
			case y:
				s = 8, i |= 24;
				break;
			case b: return e = li(12, n, t, i | 2), e.elementType = b, e.lanes = a, e;
			case w: return e = li(13, n, t, i), e.elementType = w, e.lanes = a, e;
			case T: return e = li(19, n, t, i), e.elementType = T, e.lanes = a, e;
			default:
				if (typeof e == "object" && e) switch (e.$$typeof) {
					case S:
						s = 10;
						break a;
					case x:
						s = 9;
						break a;
					case C:
						s = 11;
						break a;
					case E:
						s = 14;
						break a;
					case D:
						s = 16, r = null;
						break a;
				}
				s = 29, n = Error(o(130, e === null ? "null" : typeof e, "")), r = null;
		}
		return t = li(s, n, t, i), t.elementType = e, t.type = r, t.lanes = a, t;
	}
	function mi(e, t, n, r) {
		return e = li(7, e, r, t), e.lanes = n, e;
	}
	function hi(e, t, n) {
		return e = li(6, e, null, t), e.lanes = n, e;
	}
	function gi(e) {
		var t = li(18, null, null, 0);
		return t.stateNode = e, t;
	}
	function _i(e, t, n) {
		return t = li(4, e.children === null ? [] : e.children, e.key, t), t.lanes = n, t.stateNode = {
			containerInfo: e.containerInfo,
			pendingChildren: null,
			implementation: e.implementation
		}, t;
	}
	var vi = /* @__PURE__ */ new WeakMap();
	function yi(e, t) {
		if (typeof e == "object" && e) {
			var n = vi.get(e);
			return n === void 0 ? (t = {
				value: e,
				source: t,
				stack: Se(t)
			}, vi.set(e, t), t) : n;
		}
		return {
			value: e,
			source: t,
			stack: Se(t)
		};
	}
	var bi = [], xi = 0, Si = null, Ci = 0, wi = [], Ti = 0, Ei = null, Di = 1, Oi = "";
	function ki(e, t) {
		bi[xi++] = Ci, bi[xi++] = Si, Si = e, Ci = t;
	}
	function Ai(e, t, n) {
		wi[Ti++] = Di, wi[Ti++] = Oi, wi[Ti++] = Ei, Ei = e;
		var r = Di;
		e = Oi;
		var i = 32 - ze(r) - 1;
		r &= ~(1 << i), n += 1;
		var a = 32 - ze(t) + i;
		if (30 < a) {
			var o = i - i % 5;
			a = (r & (1 << o) - 1).toString(32), r >>= o, i -= o, Di = 1 << 32 - ze(t) + i | n << i | r, Oi = a + e;
		} else Di = 1 << a | n << i | r, Oi = e;
	}
	function ji(e) {
		e.return !== null && (ki(e, 1), Ai(e, 1, 0));
	}
	function Mi(e) {
		for (; e === Si;) Si = bi[--xi], bi[xi] = null, Ci = bi[--xi], bi[xi] = null;
		for (; e === Ei;) Ei = wi[--Ti], wi[Ti] = null, Oi = wi[--Ti], wi[Ti] = null, Di = wi[--Ti], wi[Ti] = null;
	}
	function Ni(e, t) {
		wi[Ti++] = Di, wi[Ti++] = Oi, wi[Ti++] = Ei, Di = t.id, Oi = t.overflow, Ei = e;
	}
	var Pi = null, Fi = null, R = !1, Ii = null, Li = !1, Ri = Error(o(519));
	function zi(e) {
		throw Gi(yi(Error(o(418, 1 < arguments.length && arguments[1] !== void 0 && arguments[1] ? "text" : "HTML", "")), e)), Ri;
	}
	function Bi(e) {
		var t = e.stateNode, n = e.type, r = e.memoizedProps;
		switch (t[ct] = e, t[lt] = r, n) {
			case "dialog":
				Q("cancel", t), Q("close", t);
				break;
			case "iframe":
			case "object":
			case "embed":
				Q("load", t);
				break;
			case "video":
			case "audio":
				for (n = 0; n < _d.length; n++) Q(_d[n], t);
				break;
			case "source":
				Q("error", t);
				break;
			case "img":
			case "image":
			case "link":
				Q("error", t), Q("load", t);
				break;
			case "details":
				Q("toggle", t);
				break;
			case "input":
				Q("invalid", t), Ht(t, r.value, r.defaultValue, r.checked, r.defaultChecked, r.type, r.name, !0);
				break;
			case "select":
				Q("invalid", t);
				break;
			case "textarea": Q("invalid", t), qt(t, r.value, r.defaultValue, r.children);
		}
		n = r.children, typeof n != "string" && typeof n != "number" && typeof n != "bigint" || t.textContent === "" + n || !0 === r.suppressHydrationWarning || Md(t.textContent, n) ? (r.popover != null && (Q("beforetoggle", t), Q("toggle", t)), r.onScroll != null && Q("scroll", t), r.onScrollEnd != null && Q("scrollend", t), r.onClick != null && (t.onclick = en), t = !0) : t = !1, t || zi(e, !0);
	}
	function Vi(e) {
		for (Pi = e.return; Pi;) switch (Pi.tag) {
			case 5:
			case 31:
			case 13:
				Li = !1;
				return;
			case 27:
			case 3:
				Li = !0;
				return;
			default: Pi = Pi.return;
		}
	}
	function Hi(e) {
		if (e !== Pi) return !1;
		if (!R) return Vi(e), R = !0, !1;
		var t = e.tag, n;
		if ((n = t !== 3 && t !== 27) && ((n = t === 5) && (n = e.type, n = n === "form" || n === "button" || Ud(e.type, e.memoizedProps)), n = !n), n && Fi && zi(e), Vi(e), t === 13) {
			if (e = e.memoizedState, e = e === null ? null : e.dehydrated, !e) throw Error(o(317));
			Fi = uf(e);
		} else if (t === 31) {
			if (e = e.memoizedState, e = e === null ? null : e.dehydrated, !e) throw Error(o(317));
			Fi = uf(e);
		} else t === 27 ? (t = Fi, Zd(e.type) ? (e = lf, lf = null, Fi = e) : Fi = t) : Fi = Pi ? cf(e.stateNode.nextSibling) : null;
		return !0;
	}
	function Ui() {
		Fi = Pi = null, R = !1;
	}
	function Wi() {
		var e = Ii;
		return e !== null && (Zl === null ? Zl = e : Zl.push.apply(Zl, e), Ii = null), e;
	}
	function Gi(e) {
		Ii === null ? Ii = [e] : Ii.push(e);
	}
	var Ki = se(null), qi = null, Ji = null;
	function Yi(e, t, n) {
		N(Ki, t._currentValue), t._currentValue = n;
	}
	function Xi(e) {
		e._currentValue = Ki.current, ce(Ki);
	}
	function Zi(e, t, n) {
		for (; e !== null;) {
			var r = e.alternate;
			if ((e.childLanes & t) === t ? r !== null && (r.childLanes & t) !== t && (r.childLanes |= t) : (e.childLanes |= t, r !== null && (r.childLanes |= t)), e === n) break;
			e = e.return;
		}
	}
	function Qi(e, t, n, r) {
		var i = e.child;
		for (i !== null && (i.return = e); i !== null;) {
			var a = i.dependencies;
			if (a !== null) {
				var s = i.child;
				a = a.firstContext;
				a: for (; a !== null;) {
					var c = a;
					a = i;
					for (var l = 0; l < t.length; l++) if (c.context === t[l]) {
						a.lanes |= n, c = a.alternate, c !== null && (c.lanes |= n), Zi(a.return, n, e), r || (s = null);
						break a;
					}
					a = c.next;
				}
			} else if (i.tag === 18) {
				if (s = i.return, s === null) throw Error(o(341));
				s.lanes |= n, a = s.alternate, a !== null && (a.lanes |= n), Zi(s, n, e), s = null;
			} else s = i.child;
			if (s !== null) s.return = i;
			else for (s = i; s !== null;) {
				if (s === e) {
					s = null;
					break;
				}
				if (i = s.sibling, i !== null) {
					i.return = s.return, s = i;
					break;
				}
				s = s.return;
			}
			i = s;
		}
	}
	function $i(e, t, n, r) {
		e = null;
		for (var i = t, a = !1; i !== null;) {
			if (!a) {
				if (i.flags & 524288) a = !0;
				else if (i.flags & 262144) break;
			}
			if (i.tag === 10) {
				var s = i.alternate;
				if (s === null) throw Error(o(387));
				if (s = s.memoizedProps, s !== null) {
					var c = i.type;
					Cr(i.pendingProps.value, s.value) || (e === null ? e = [c] : e.push(c));
				}
			} else if (i === fe.current) {
				if (s = i.alternate, s === null) throw Error(o(387));
				s.memoizedState.memoizedState !== i.memoizedState.memoizedState && (e === null ? e = [Qf] : e.push(Qf));
			}
			i = i.return;
		}
		e !== null && Qi(t, e, n, r), t.flags |= 262144;
	}
	function ea(e) {
		for (e = e.firstContext; e !== null;) {
			if (!Cr(e.context._currentValue, e.memoizedValue)) return !0;
			e = e.next;
		}
		return !1;
	}
	function ta(e) {
		qi = e, Ji = null, e = e.dependencies, e !== null && (e.firstContext = null);
	}
	function na(e) {
		return ia(qi, e);
	}
	function ra(e, t) {
		return qi === null && ta(e), ia(e, t);
	}
	function ia(e, t) {
		var n = t._currentValue;
		if (t = {
			context: t,
			memoizedValue: n,
			next: null
		}, Ji === null) {
			if (e === null) throw Error(o(308));
			Ji = t, e.dependencies = {
				lanes: 0,
				firstContext: t
			}, e.flags |= 524288;
		} else Ji = Ji.next = t;
		return n;
	}
	var aa = typeof AbortController < "u" ? AbortController : function() {
		var e = [], t = this.signal = {
			aborted: !1,
			addEventListener: function(t, n) {
				e.push(n);
			}
		};
		this.abort = function() {
			t.aborted = !0, e.forEach(function(e) {
				return e();
			});
		};
	}, oa = t.unstable_scheduleCallback, sa = t.unstable_NormalPriority, ca = {
		$$typeof: S,
		Consumer: null,
		Provider: null,
		_currentValue: null,
		_currentValue2: null,
		_threadCount: 0
	};
	function la() {
		return {
			controller: new aa(),
			data: /* @__PURE__ */ new Map(),
			refCount: 0
		};
	}
	function ua(e) {
		e.refCount--, e.refCount === 0 && oa(sa, function() {
			e.controller.abort();
		});
	}
	var da = null, fa = 0, pa = 0, ma = null;
	function ha(e, t) {
		if (da === null) {
			var n = da = [];
			fa = 0, pa = dd(), ma = {
				status: "pending",
				value: void 0,
				then: function(e) {
					n.push(e);
				}
			};
		}
		return fa++, t.then(ga, ga), t;
	}
	function ga() {
		if (--fa === 0 && da !== null) {
			ma !== null && (ma.status = "fulfilled");
			var e = da;
			da = null, pa = 0, ma = null;
			for (var t = 0; t < e.length; t++) (0, e[t])();
		}
	}
	function _a(e, t) {
		var n = [], r = {
			status: "pending",
			value: null,
			reason: null,
			then: function(e) {
				n.push(e);
			}
		};
		return e.then(function() {
			r.status = "fulfilled", r.value = t;
			for (var e = 0; e < n.length; e++) (0, n[e])(t);
		}, function(e) {
			for (r.status = "rejected", r.reason = e, e = 0; e < n.length; e++) (0, n[e])(void 0);
		}), r;
	}
	var va = A.S;
	A.S = function(e, t) {
		eu = De(), typeof t == "object" && t && typeof t.then == "function" && ha(e, t), va !== null && va(e, t);
	};
	var ya = se(null);
	function ba() {
		var e = ya.current;
		return e === null ? q.pooledCache : e;
	}
	function xa(e, t) {
		t === null ? N(ya, ya.current) : N(ya, t.pool);
	}
	function Sa() {
		var e = ba();
		return e === null ? null : {
			parent: ca._currentValue,
			pool: e
		};
	}
	var Ca = Error(o(460)), wa = Error(o(474)), Ta = Error(o(542)), Ea = { then: function() {} };
	function Da(e) {
		return e = e.status, e === "fulfilled" || e === "rejected";
	}
	function Oa(e, t, n) {
		switch (n = e[n], n === void 0 ? e.push(t) : n !== t && (t.then(en, en), t = n), t.status) {
			case "fulfilled": return t.value;
			case "rejected": throw e = t.reason, Ma(e), e;
			default:
				if (typeof t.status == "string") t.then(en, en);
				else {
					if (e = q, e !== null && 100 < e.shellSuspendCounter) throw Error(o(482));
					e = t, e.status = "pending", e.then(function(e) {
						if (t.status === "pending") {
							var n = t;
							n.status = "fulfilled", n.value = e;
						}
					}, function(e) {
						if (t.status === "pending") {
							var n = t;
							n.status = "rejected", n.reason = e;
						}
					});
				}
				switch (t.status) {
					case "fulfilled": return t.value;
					case "rejected": throw e = t.reason, Ma(e), e;
				}
				throw Aa = t, Ca;
		}
	}
	function ka(e) {
		try {
			var t = e._init;
			return t(e._payload);
		} catch (e) {
			throw typeof e == "object" && e && typeof e.then == "function" ? (Aa = e, Ca) : e;
		}
	}
	var Aa = null;
	function ja() {
		if (Aa === null) throw Error(o(459));
		var e = Aa;
		return Aa = null, e;
	}
	function Ma(e) {
		if (e === Ca || e === Ta) throw Error(o(483));
	}
	var Na = null, Pa = 0;
	function Fa(e) {
		var t = Pa;
		return Pa += 1, Na === null && (Na = []), Oa(Na, e, t);
	}
	function Ia(e, t) {
		t = t.props.ref, e.ref = t === void 0 ? null : t;
	}
	function La(e, t) {
		throw t.$$typeof === h ? Error(o(525)) : (e = Object.prototype.toString.call(t), Error(o(31, e === "[object Object]" ? "object with keys {" + Object.keys(t).join(", ") + "}" : e)));
	}
	function Ra(e) {
		function t(t, n) {
			if (e) {
				var r = t.deletions;
				r === null ? (t.deletions = [n], t.flags |= 16) : r.push(n);
			}
		}
		function n(n, r) {
			if (!e) return null;
			for (; r !== null;) t(n, r), r = r.sibling;
			return null;
		}
		function r(e) {
			for (var t = /* @__PURE__ */ new Map(); e !== null;) e.key === null ? t.set(e.index, e) : t.set(e.key, e), e = e.sibling;
			return t;
		}
		function i(e, t) {
			return e = di(e, t), e.index = 0, e.sibling = null, e;
		}
		function a(t, n, r) {
			return t.index = r, e ? (r = t.alternate, r === null ? (t.flags |= 67108866, n) : (r = r.index, r < n ? (t.flags |= 67108866, n) : r)) : (t.flags |= 1048576, n);
		}
		function s(t) {
			return e && t.alternate === null && (t.flags |= 67108866), t;
		}
		function c(e, t, n, r) {
			return t === null || t.tag !== 6 ? (t = hi(n, e.mode, r), t.return = e, t) : (t = i(t, n), t.return = e, t);
		}
		function l(e, t, n, r) {
			var a = n.type;
			return a === v ? d(e, t, n.props.children, r, n.key) : t !== null && (t.elementType === a || typeof a == "object" && a && a.$$typeof === D && ka(a) === t.type) ? (t = i(t, n.props), Ia(t, n), t.return = e, t) : (t = pi(n.type, n.key, n.props, null, e.mode, r), Ia(t, n), t.return = e, t);
		}
		function u(e, t, n, r) {
			return t === null || t.tag !== 4 || t.stateNode.containerInfo !== n.containerInfo || t.stateNode.implementation !== n.implementation ? (t = _i(n, e.mode, r), t.return = e, t) : (t = i(t, n.children || []), t.return = e, t);
		}
		function d(e, t, n, r, a) {
			return t === null || t.tag !== 7 ? (t = mi(n, e.mode, r, a), t.return = e, t) : (t = i(t, n), t.return = e, t);
		}
		function f(e, t, n) {
			if (typeof t == "string" && t !== "" || typeof t == "number" || typeof t == "bigint") return t = hi("" + t, e.mode, n), t.return = e, t;
			if (typeof t == "object" && t) {
				switch (t.$$typeof) {
					case g: return n = pi(t.type, t.key, t.props, null, e.mode, n), Ia(n, t), n.return = e, n;
					case _: return t = _i(t, e.mode, n), t.return = e, t;
					case D: return t = ka(t), f(e, t, n);
				}
				if (k(t) || ne(t)) return t = mi(t, e.mode, n, null), t.return = e, t;
				if (typeof t.then == "function") return f(e, Fa(t), n);
				if (t.$$typeof === S) return f(e, ra(e, t), n);
				La(e, t);
			}
			return null;
		}
		function p(e, t, n, r) {
			var i = t === null ? null : t.key;
			if (typeof n == "string" && n !== "" || typeof n == "number" || typeof n == "bigint") return i === null ? c(e, t, "" + n, r) : null;
			if (typeof n == "object" && n) {
				switch (n.$$typeof) {
					case g: return n.key === i ? l(e, t, n, r) : null;
					case _: return n.key === i ? u(e, t, n, r) : null;
					case D: return n = ka(n), p(e, t, n, r);
				}
				if (k(n) || ne(n)) return i === null ? d(e, t, n, r, null) : null;
				if (typeof n.then == "function") return p(e, t, Fa(n), r);
				if (n.$$typeof === S) return p(e, t, ra(e, n), r);
				La(e, n);
			}
			return null;
		}
		function m(e, t, n, r, i) {
			if (typeof r == "string" && r !== "" || typeof r == "number" || typeof r == "bigint") return e = e.get(n) || null, c(t, e, "" + r, i);
			if (typeof r == "object" && r) {
				switch (r.$$typeof) {
					case g: return e = e.get(r.key === null ? n : r.key) || null, l(t, e, r, i);
					case _: return e = e.get(r.key === null ? n : r.key) || null, u(t, e, r, i);
					case D: return r = ka(r), m(e, t, n, r, i);
				}
				if (k(r) || ne(r)) return e = e.get(n) || null, d(t, e, r, i, null);
				if (typeof r.then == "function") return m(e, t, n, Fa(r), i);
				if (r.$$typeof === S) return m(e, t, n, ra(t, r), i);
				La(t, r);
			}
			return null;
		}
		function h(i, o, s, c) {
			for (var l = null, u = null, d = o, h = o = 0, g = null; d !== null && h < s.length; h++) {
				d.index > h ? (g = d, d = null) : g = d.sibling;
				var _ = p(i, d, s[h], c);
				if (_ === null) {
					d === null && (d = g);
					break;
				}
				e && d && _.alternate === null && t(i, d), o = a(_, o, h), u === null ? l = _ : u.sibling = _, u = _, d = g;
			}
			if (h === s.length) return n(i, d), R && ki(i, h), l;
			if (d === null) {
				for (; h < s.length; h++) d = f(i, s[h], c), d !== null && (o = a(d, o, h), u === null ? l = d : u.sibling = d, u = d);
				return R && ki(i, h), l;
			}
			for (d = r(d); h < s.length; h++) g = m(d, i, h, s[h], c), g !== null && (e && g.alternate !== null && d.delete(g.key === null ? h : g.key), o = a(g, o, h), u === null ? l = g : u.sibling = g, u = g);
			return e && d.forEach(function(e) {
				return t(i, e);
			}), R && ki(i, h), l;
		}
		function y(i, s, c, l) {
			if (c == null) throw Error(o(151));
			for (var u = null, d = null, h = s, g = s = 0, _ = null, v = c.next(); h !== null && !v.done; g++, v = c.next()) {
				h.index > g ? (_ = h, h = null) : _ = h.sibling;
				var y = p(i, h, v.value, l);
				if (y === null) {
					h === null && (h = _);
					break;
				}
				e && h && y.alternate === null && t(i, h), s = a(y, s, g), d === null ? u = y : d.sibling = y, d = y, h = _;
			}
			if (v.done) return n(i, h), R && ki(i, g), u;
			if (h === null) {
				for (; !v.done; g++, v = c.next()) v = f(i, v.value, l), v !== null && (s = a(v, s, g), d === null ? u = v : d.sibling = v, d = v);
				return R && ki(i, g), u;
			}
			for (h = r(h); !v.done; g++, v = c.next()) v = m(h, i, g, v.value, l), v !== null && (e && v.alternate !== null && h.delete(v.key === null ? g : v.key), s = a(v, s, g), d === null ? u = v : d.sibling = v, d = v);
			return e && h.forEach(function(e) {
				return t(i, e);
			}), R && ki(i, g), u;
		}
		function b(e, r, a, c) {
			if (typeof a == "object" && a && a.type === v && a.key === null && (a = a.props.children), typeof a == "object" && a) {
				switch (a.$$typeof) {
					case g:
						a: {
							for (var l = a.key; r !== null;) {
								if (r.key === l) {
									if (l = a.type, l === v) {
										if (r.tag === 7) {
											n(e, r.sibling), c = i(r, a.props.children), c.return = e, e = c;
											break a;
										}
									} else if (r.elementType === l || typeof l == "object" && l && l.$$typeof === D && ka(l) === r.type) {
										n(e, r.sibling), c = i(r, a.props), Ia(c, a), c.return = e, e = c;
										break a;
									}
									n(e, r);
									break;
								}
								t(e, r), r = r.sibling;
							}
							a.type === v ? (c = mi(a.props.children, e.mode, c, a.key), c.return = e, e = c) : (c = pi(a.type, a.key, a.props, null, e.mode, c), Ia(c, a), c.return = e, e = c);
						}
						return s(e);
					case _:
						a: {
							for (l = a.key; r !== null;) {
								if (r.key === l) {
									if (r.tag === 4 && r.stateNode.containerInfo === a.containerInfo && r.stateNode.implementation === a.implementation) {
										n(e, r.sibling), c = i(r, a.children || []), c.return = e, e = c;
										break a;
									}
									n(e, r);
									break;
								}
								t(e, r), r = r.sibling;
							}
							c = _i(a, e.mode, c), c.return = e, e = c;
						}
						return s(e);
					case D: return a = ka(a), b(e, r, a, c);
				}
				if (k(a)) return h(e, r, a, c);
				if (ne(a)) {
					if (l = ne(a), typeof l != "function") throw Error(o(150));
					return a = l.call(a), y(e, r, a, c);
				}
				if (typeof a.then == "function") return b(e, r, Fa(a), c);
				if (a.$$typeof === S) return b(e, r, ra(e, a), c);
				La(e, a);
			}
			return typeof a == "string" && a !== "" || typeof a == "number" || typeof a == "bigint" ? (a = "" + a, r !== null && r.tag === 6 ? (n(e, r.sibling), c = i(r, a), c.return = e, e = c) : (n(e, r), c = hi(a, e.mode, c), c.return = e, e = c), s(e)) : n(e, r);
		}
		return function(e, t, n, r) {
			try {
				Pa = 0;
				var i = b(e, t, n, r);
				return Na = null, i;
			} catch (t) {
				if (t === Ca || t === Ta) throw t;
				var a = li(29, t, null, e.mode);
				return a.lanes = r, a.return = e, a;
			}
		};
	}
	var za = Ra(!0), Ba = Ra(!1), Va = !1;
	function Ha(e) {
		e.updateQueue = {
			baseState: e.memoizedState,
			firstBaseUpdate: null,
			lastBaseUpdate: null,
			shared: {
				pending: null,
				lanes: 0,
				hiddenCallbacks: null
			},
			callbacks: null
		};
	}
	function Ua(e, t) {
		e = e.updateQueue, t.updateQueue === e && (t.updateQueue = {
			baseState: e.baseState,
			firstBaseUpdate: e.firstBaseUpdate,
			lastBaseUpdate: e.lastBaseUpdate,
			shared: e.shared,
			callbacks: null
		});
	}
	function Wa(e) {
		return {
			lane: e,
			tag: 0,
			payload: null,
			callback: null,
			next: null
		};
	}
	function Ga(e, t, n) {
		var r = e.updateQueue;
		if (r === null) return null;
		if (r = r.shared, K & 2) {
			var i = r.pending;
			return i === null ? t.next = t : (t.next = i.next, i.next = t), r.pending = t, t = oi(e), ai(e, null, n), t;
		}
		return ni(e, r, t, n), oi(e);
	}
	function Ka(e, t, n) {
		if (t = t.updateQueue, t !== null && (t = t.shared, n & 4194048)) {
			var r = t.lanes;
			r &= e.pendingLanes, n |= r, t.lanes = n, tt(e, n);
		}
	}
	function qa(e, t) {
		var n = e.updateQueue, r = e.alternate;
		if (r !== null && (r = r.updateQueue, n === r)) {
			var i = null, a = null;
			if (n = n.firstBaseUpdate, n !== null) {
				do {
					var o = {
						lane: n.lane,
						tag: n.tag,
						payload: n.payload,
						callback: null,
						next: null
					};
					a === null ? i = a = o : a = a.next = o, n = n.next;
				} while (n !== null);
				a === null ? i = a = t : a = a.next = t;
			} else i = a = t;
			n = {
				baseState: r.baseState,
				firstBaseUpdate: i,
				lastBaseUpdate: a,
				shared: r.shared,
				callbacks: r.callbacks
			}, e.updateQueue = n;
		} else e = n.lastBaseUpdate, e === null ? n.firstBaseUpdate = t : e.next = t, n.lastBaseUpdate = t;
	}
	var Ja = !1;
	function Ya() {
		if (Ja) {
			var e = ma;
			if (e !== null) throw e;
		}
	}
	function Xa(e, t, n, r) {
		Ja = !1;
		var i = e.updateQueue;
		Va = !1;
		var a = i.firstBaseUpdate, o = i.lastBaseUpdate, s = i.shared.pending;
		if (s !== null) {
			i.shared.pending = null;
			var c = s, l = c.next;
			c.next = null, o === null ? a = l : o.next = l, o = c;
			var u = e.alternate;
			u !== null && (u = u.updateQueue, s = u.lastBaseUpdate, s !== o && (s === null ? u.firstBaseUpdate = l : s.next = l, u.lastBaseUpdate = c));
		}
		if (a !== null) {
			var d = i.baseState;
			o = 0, u = l = c = null, s = a;
			do {
				var f = s.lane & -536870913, p = f !== s.lane;
				if (p ? (Y & f) === f : (r & f) === f) {
					f !== 0 && f === pa && (Ja = !0), u !== null && (u = u.next = {
						lane: 0,
						tag: s.tag,
						payload: s.payload,
						callback: null,
						next: null
					});
					a: {
						var h = e, g = s;
						f = t;
						var _ = n;
						switch (g.tag) {
							case 1:
								if (h = g.payload, typeof h == "function") {
									d = h.call(_, d, f);
									break a;
								}
								d = h;
								break a;
							case 3: h.flags = h.flags & -65537 | 128;
							case 0:
								if (h = g.payload, f = typeof h == "function" ? h.call(_, d, f) : h, f == null) break a;
								d = m({}, d, f);
								break a;
							case 2: Va = !0;
						}
					}
					f = s.callback, f !== null && (e.flags |= 64, p && (e.flags |= 8192), p = i.callbacks, p === null ? i.callbacks = [f] : p.push(f));
				} else p = {
					lane: f,
					tag: s.tag,
					payload: s.payload,
					callback: s.callback,
					next: null
				}, u === null ? (l = u = p, c = d) : u = u.next = p, o |= f;
				if (s = s.next, s === null) {
					if (s = i.shared.pending, s === null) break;
					p = s, s = p.next, p.next = null, i.lastBaseUpdate = p, i.shared.pending = null;
				}
			} while (1);
			u === null && (c = d), i.baseState = c, i.firstBaseUpdate = l, i.lastBaseUpdate = u, a === null && (i.shared.lanes = 0), Gl |= o, e.lanes = o, e.memoizedState = d;
		}
	}
	function Za(e, t) {
		if (typeof e != "function") throw Error(o(191, e));
		e.call(t);
	}
	function Qa(e, t) {
		var n = e.callbacks;
		if (n !== null) for (e.callbacks = null, e = 0; e < n.length; e++) Za(n[e], t);
	}
	var $a = se(null), eo = se(0);
	function to(e, t) {
		e = Ul, N(eo, e), N($a, t), Ul = e | t.baseLanes;
	}
	function no() {
		N(eo, Ul), N($a, $a.current);
	}
	function ro() {
		Ul = eo.current, ce($a), ce(eo);
	}
	var io = se(null), ao = null;
	function oo(e) {
		var t = e.alternate;
		N(fo, fo.current & 1), N(io, e), ao === null && (t === null || $a.current !== null || t.memoizedState !== null) && (ao = e);
	}
	function so(e) {
		N(fo, fo.current), N(io, e), ao === null && (ao = e);
	}
	function co(e) {
		e.tag === 22 ? (N(fo, fo.current), N(io, e), ao === null && (ao = e)) : lo(e);
	}
	function lo() {
		N(fo, fo.current), N(io, io.current);
	}
	function uo(e) {
		ce(io), ao === e && (ao = null), ce(fo);
	}
	var fo = se(0);
	function po(e) {
		for (var t = e; t !== null;) {
			if (t.tag === 13) {
				var n = t.memoizedState;
				if (n !== null && (n = n.dehydrated, n === null || af(n) || of(n))) return t;
			} else if (t.tag === 19 && (t.memoizedProps.revealOrder === "forwards" || t.memoizedProps.revealOrder === "backwards" || t.memoizedProps.revealOrder === "unstable_legacy-backwards" || t.memoizedProps.revealOrder === "together")) {
				if (t.flags & 128) return t;
			} else if (t.child !== null) {
				t.child.return = t, t = t.child;
				continue;
			}
			if (t === e) break;
			for (; t.sibling === null;) {
				if (t.return === null || t.return === e) return null;
				t = t.return;
			}
			t.sibling.return = t.return, t = t.sibling;
		}
		return null;
	}
	var z = 0, B = null, V = null, H = null, mo = !1, ho = !1, go = !1, _o = 0, vo = 0, yo = null, bo = 0;
	function xo() {
		throw Error(o(321));
	}
	function So(e, t) {
		if (t === null) return !1;
		for (var n = 0; n < t.length && n < e.length; n++) if (!Cr(e[n], t[n])) return !1;
		return !0;
	}
	function Co(e, t, n, r, i, a) {
		return z = a, B = t, t.memoizedState = null, t.updateQueue = null, t.lanes = 0, A.H = e === null || e.memoizedState === null ? zs : Bs, go = !1, a = n(r, i), go = !1, ho && (a = To(t, n, r, i)), wo(e), a;
	}
	function wo(e) {
		A.H = U;
		var t = V !== null && V.next !== null;
		if (z = 0, H = V = B = null, mo = !1, vo = 0, yo = null, t) throw Error(o(300));
		e === null || rc || (e = e.dependencies, e !== null && ea(e) && (rc = !0));
	}
	function To(e, t, n, r) {
		B = e;
		var i = 0;
		do {
			if (ho && (yo = null), vo = 0, ho = !1, 25 <= i) throw Error(o(301));
			if (i += 1, H = V = null, e.updateQueue != null) {
				var a = e.updateQueue;
				a.lastEffect = null, a.events = null, a.stores = null, a.memoCache != null && (a.memoCache.index = 0);
			}
			A.H = Vs, a = t(n, r);
		} while (ho);
		return a;
	}
	function Eo() {
		var e = A.H, t = e.useState()[0];
		return t = typeof t.then == "function" ? No(t) : t, e = e.useState()[0], (V === null ? null : V.memoizedState) !== e && (B.flags |= 1024), t;
	}
	function Do() {
		var e = _o !== 0;
		return _o = 0, e;
	}
	function Oo(e, t, n) {
		t.updateQueue = e.updateQueue, t.flags &= -2053, e.lanes &= ~n;
	}
	function ko(e) {
		if (mo) {
			for (e = e.memoizedState; e !== null;) {
				var t = e.queue;
				t !== null && (t.pending = null), e = e.next;
			}
			mo = !1;
		}
		z = 0, H = V = B = null, ho = !1, vo = _o = 0, yo = null;
	}
	function Ao() {
		var e = {
			memoizedState: null,
			baseState: null,
			baseQueue: null,
			queue: null,
			next: null
		};
		return H === null ? B.memoizedState = H = e : H = H.next = e, H;
	}
	function jo() {
		if (V === null) {
			var e = B.alternate;
			e = e === null ? null : e.memoizedState;
		} else e = V.next;
		var t = H === null ? B.memoizedState : H.next;
		if (t !== null) H = t, V = e;
		else {
			if (e === null) throw B.alternate === null ? Error(o(467)) : Error(o(310));
			V = e, e = {
				memoizedState: V.memoizedState,
				baseState: V.baseState,
				baseQueue: V.baseQueue,
				queue: V.queue,
				next: null
			}, H === null ? B.memoizedState = H = e : H = H.next = e;
		}
		return H;
	}
	function Mo() {
		return {
			lastEffect: null,
			events: null,
			stores: null,
			memoCache: null
		};
	}
	function No(e) {
		var t = vo;
		return vo += 1, yo === null && (yo = []), e = Oa(yo, e, t), t = B, (H === null ? t.memoizedState : H.next) === null && (t = t.alternate, A.H = t === null || t.memoizedState === null ? zs : Bs), e;
	}
	function Po(e) {
		if (typeof e == "object" && e) {
			if (typeof e.then == "function") return No(e);
			if (e.$$typeof === S) return na(e);
		}
		throw Error(o(438, String(e)));
	}
	function Fo(e) {
		var t = null, n = B.updateQueue;
		if (n !== null && (t = n.memoCache), t == null) {
			var r = B.alternate;
			r !== null && (r = r.updateQueue, r !== null && (r = r.memoCache, r != null && (t = {
				data: r.data.map(function(e) {
					return e.slice();
				}),
				index: 0
			})));
		}
		if (t ??= {
			data: [],
			index: 0
		}, n === null && (n = Mo(), B.updateQueue = n), n.memoCache = t, n = t.data[t.index], n === void 0) for (n = t.data[t.index] = Array(e), r = 0; r < e; r++) n[r] = ee;
		return t.index++, n;
	}
	function Io(e, t) {
		return typeof t == "function" ? t(e) : t;
	}
	function Lo(e) {
		return Ro(jo(), V, e);
	}
	function Ro(e, t, n) {
		var r = e.queue;
		if (r === null) throw Error(o(311));
		r.lastRenderedReducer = n;
		var i = e.baseQueue, a = r.pending;
		if (a !== null) {
			if (i !== null) {
				var s = i.next;
				i.next = a.next, a.next = s;
			}
			t.baseQueue = i = a, r.pending = null;
		}
		if (a = e.baseState, i === null) e.memoizedState = a;
		else {
			t = i.next;
			var c = s = null, l = null, u = t, d = !1;
			do {
				var f = u.lane & -536870913;
				if (f === u.lane ? (z & f) === f : (Y & f) === f) {
					var p = u.revertLane;
					if (p === 0) l !== null && (l = l.next = {
						lane: 0,
						revertLane: 0,
						gesture: null,
						action: u.action,
						hasEagerState: u.hasEagerState,
						eagerState: u.eagerState,
						next: null
					}), f === pa && (d = !0);
					else if ((z & p) === p) {
						u = u.next, p === pa && (d = !0);
						continue;
					} else f = {
						lane: 0,
						revertLane: u.revertLane,
						gesture: null,
						action: u.action,
						hasEagerState: u.hasEagerState,
						eagerState: u.eagerState,
						next: null
					}, l === null ? (c = l = f, s = a) : l = l.next = f, B.lanes |= p, Gl |= p;
					f = u.action, go && n(a, f), a = u.hasEagerState ? u.eagerState : n(a, f);
				} else p = {
					lane: f,
					revertLane: u.revertLane,
					gesture: u.gesture,
					action: u.action,
					hasEagerState: u.hasEagerState,
					eagerState: u.eagerState,
					next: null
				}, l === null ? (c = l = p, s = a) : l = l.next = p, B.lanes |= f, Gl |= f;
				u = u.next;
			} while (u !== null && u !== t);
			if (l === null ? s = a : l.next = c, !Cr(a, e.memoizedState) && (rc = !0, d && (n = ma, n !== null))) throw n;
			e.memoizedState = a, e.baseState = s, e.baseQueue = l, r.lastRenderedState = a;
		}
		return i === null && (r.lanes = 0), [e.memoizedState, r.dispatch];
	}
	function zo(e) {
		var t = jo(), n = t.queue;
		if (n === null) throw Error(o(311));
		n.lastRenderedReducer = e;
		var r = n.dispatch, i = n.pending, a = t.memoizedState;
		if (i !== null) {
			n.pending = null;
			var s = i = i.next;
			do
				a = e(a, s.action), s = s.next;
			while (s !== i);
			Cr(a, t.memoizedState) || (rc = !0), t.memoizedState = a, t.baseQueue === null && (t.baseState = a), n.lastRenderedState = a;
		}
		return [a, r];
	}
	function Bo(e, t, n) {
		var r = B, i = jo(), a = R;
		if (a) {
			if (n === void 0) throw Error(o(407));
			n = n();
		} else n = t();
		var s = !Cr((V || i).memoizedState, n);
		if (s && (i.memoizedState = n, rc = !0), i = i.queue, ds(Uo.bind(null, r, i, e), [e]), i.getSnapshot !== t || s || H !== null && H.memoizedState.tag & 1) {
			if (r.flags |= 2048, os(9, { destroy: void 0 }, Ho.bind(null, r, i, n, t), null), q === null) throw Error(o(349));
			a || z & 127 || Vo(r, t, n);
		}
		return n;
	}
	function Vo(e, t, n) {
		e.flags |= 16384, e = {
			getSnapshot: t,
			value: n
		}, t = B.updateQueue, t === null ? (t = Mo(), B.updateQueue = t, t.stores = [e]) : (n = t.stores, n === null ? t.stores = [e] : n.push(e));
	}
	function Ho(e, t, n, r) {
		t.value = n, t.getSnapshot = r, Wo(t) && Go(e);
	}
	function Uo(e, t, n) {
		return n(function() {
			Wo(t) && Go(e);
		});
	}
	function Wo(e) {
		var t = e.getSnapshot;
		e = e.value;
		try {
			var n = t();
			return !Cr(e, n);
		} catch {
			return !0;
		}
	}
	function Go(e) {
		var t = ii(e, 2);
		t !== null && hu(t, e, 2);
	}
	function Ko(e) {
		var t = Ao();
		if (typeof e == "function") {
			var n = e;
			if (e = n(), go) {
				Re(!0);
				try {
					n();
				} finally {
					Re(!1);
				}
			}
		}
		return t.memoizedState = t.baseState = e, t.queue = {
			pending: null,
			lanes: 0,
			dispatch: null,
			lastRenderedReducer: Io,
			lastRenderedState: e
		}, t;
	}
	function qo(e, t, n, r) {
		return e.baseState = n, Ro(e, V, typeof r == "function" ? r : Io);
	}
	function Jo(e, t, n, r, i) {
		if (Is(e)) throw Error(o(485));
		if (e = t.action, e !== null) {
			var a = {
				payload: i,
				action: e,
				next: null,
				isTransition: !0,
				status: "pending",
				value: null,
				reason: null,
				listeners: [],
				then: function(e) {
					a.listeners.push(e);
				}
			};
			A.T === null ? a.isTransition = !1 : n(!0), r(a), n = t.pending, n === null ? (a.next = t.pending = a, Yo(t, a)) : (a.next = n.next, t.pending = n.next = a);
		}
	}
	function Yo(e, t) {
		var n = t.action, r = t.payload, i = e.state;
		if (t.isTransition) {
			var a = A.T, o = {};
			A.T = o;
			try {
				var s = n(i, r), c = A.S;
				c !== null && c(o, s), Xo(e, t, s);
			} catch (n) {
				Qo(e, t, n);
			} finally {
				a !== null && o.types !== null && (a.types = o.types), A.T = a;
			}
		} else try {
			a = n(i, r), Xo(e, t, a);
		} catch (n) {
			Qo(e, t, n);
		}
	}
	function Xo(e, t, n) {
		typeof n == "object" && n && typeof n.then == "function" ? n.then(function(n) {
			Zo(e, t, n);
		}, function(n) {
			return Qo(e, t, n);
		}) : Zo(e, t, n);
	}
	function Zo(e, t, n) {
		t.status = "fulfilled", t.value = n, $o(t), e.state = n, t = e.pending, t !== null && (n = t.next, n === t ? e.pending = null : (n = n.next, t.next = n, Yo(e, n)));
	}
	function Qo(e, t, n) {
		var r = e.pending;
		if (e.pending = null, r !== null) {
			r = r.next;
			do
				t.status = "rejected", t.reason = n, $o(t), t = t.next;
			while (t !== r);
		}
		e.action = null;
	}
	function $o(e) {
		e = e.listeners;
		for (var t = 0; t < e.length; t++) (0, e[t])();
	}
	function es(e, t) {
		return t;
	}
	function ts(e, t) {
		if (R) {
			var n = q.formState;
			if (n !== null) {
				a: {
					var r = B;
					if (R) {
						if (Fi) {
							b: {
								for (var i = Fi, a = Li; i.nodeType !== 8;) if (!a || (i = cf(i.nextSibling), i === null)) {
									i = null;
									break b;
								}
								a = i.data, i = a === "F!" || a === "F" ? i : null;
							}
							if (i) {
								Fi = cf(i.nextSibling), r = i.data === "F!";
								break a;
							}
						}
						zi(r);
					}
					r = !1;
				}
				r && (t = n[0]);
			}
		}
		return n = Ao(), n.memoizedState = n.baseState = t, r = {
			pending: null,
			lanes: 0,
			dispatch: null,
			lastRenderedReducer: es,
			lastRenderedState: t
		}, n.queue = r, n = Ns.bind(null, B, r), r.dispatch = n, r = Ko(!1), a = Fs.bind(null, B, !1, r.queue), r = Ao(), i = {
			state: t,
			dispatch: null,
			action: e,
			pending: null
		}, r.queue = i, n = Jo.bind(null, B, i, a, n), i.dispatch = n, r.memoizedState = e, [
			t,
			n,
			!1
		];
	}
	function ns(e) {
		return rs(jo(), V, e);
	}
	function rs(e, t, n) {
		if (t = Ro(e, t, es)[0], e = Lo(Io)[0], typeof t == "object" && t && typeof t.then == "function") try {
			var r = No(t);
		} catch (e) {
			throw e === Ca ? Ta : e;
		}
		else r = t;
		t = jo();
		var i = t.queue, a = i.dispatch;
		return n !== t.memoizedState && (B.flags |= 2048, os(9, { destroy: void 0 }, is.bind(null, i, n), null)), [
			r,
			a,
			e
		];
	}
	function is(e, t) {
		e.action = t;
	}
	function as(e) {
		var t = jo(), n = V;
		if (n !== null) return rs(t, n, e);
		jo(), t = t.memoizedState, n = jo();
		var r = n.queue.dispatch;
		return n.memoizedState = e, [
			t,
			r,
			!1
		];
	}
	function os(e, t, n, r) {
		return e = {
			tag: e,
			create: n,
			deps: r,
			inst: t,
			next: null
		}, t = B.updateQueue, t === null && (t = Mo(), B.updateQueue = t), n = t.lastEffect, n === null ? t.lastEffect = e.next = e : (r = n.next, n.next = e, e.next = r, t.lastEffect = e), e;
	}
	function ss() {
		return jo().memoizedState;
	}
	function cs(e, t, n, r) {
		var i = Ao();
		B.flags |= e, i.memoizedState = os(1 | t, { destroy: void 0 }, n, r === void 0 ? null : r);
	}
	function ls(e, t, n, r) {
		var i = jo();
		r = r === void 0 ? null : r;
		var a = i.memoizedState.inst;
		V !== null && r !== null && So(r, V.memoizedState.deps) ? i.memoizedState = os(t, a, n, r) : (B.flags |= e, i.memoizedState = os(1 | t, a, n, r));
	}
	function us(e, t) {
		cs(8390656, 8, e, t);
	}
	function ds(e, t) {
		ls(2048, 8, e, t);
	}
	function fs(e) {
		B.flags |= 4;
		var t = B.updateQueue;
		if (t === null) t = Mo(), B.updateQueue = t, t.events = [e];
		else {
			var n = t.events;
			n === null ? t.events = [e] : n.push(e);
		}
	}
	function ps(e) {
		var t = jo().memoizedState;
		return fs({
			ref: t,
			nextImpl: e
		}), function() {
			if (K & 2) throw Error(o(440));
			return t.impl.apply(void 0, arguments);
		};
	}
	function ms(e, t) {
		return ls(4, 2, e, t);
	}
	function hs(e, t) {
		return ls(4, 4, e, t);
	}
	function gs(e, t) {
		if (typeof t == "function") {
			e = e();
			var n = t(e);
			return function() {
				typeof n == "function" ? n() : t(null);
			};
		}
		if (t != null) return e = e(), t.current = e, function() {
			t.current = null;
		};
	}
	function _s(e, t, n) {
		n = n == null ? null : n.concat([e]), ls(4, 4, gs.bind(null, t, e), n);
	}
	function vs() {}
	function ys(e, t) {
		var n = jo();
		t = t === void 0 ? null : t;
		var r = n.memoizedState;
		return t !== null && So(t, r[1]) ? r[0] : (n.memoizedState = [e, t], e);
	}
	function bs(e, t) {
		var n = jo();
		t = t === void 0 ? null : t;
		var r = n.memoizedState;
		if (t !== null && So(t, r[1])) return r[0];
		if (r = e(), go) {
			Re(!0);
			try {
				e();
			} finally {
				Re(!1);
			}
		}
		return n.memoizedState = [r, t], r;
	}
	function xs(e, t, n) {
		return n === void 0 || z & 1073741824 && !(Y & 261930) ? e.memoizedState = t : (e.memoizedState = n, e = mu(), B.lanes |= e, Gl |= e, n);
	}
	function Ss(e, t, n, r) {
		return Cr(n, t) ? n : $a.current === null ? !(z & 42) || z & 1073741824 && !(Y & 261930) ? (rc = !0, e.memoizedState = n) : (e = mu(), B.lanes |= e, Gl |= e, t) : (e = xs(e, n, r), Cr(e, t) || (rc = !0), e);
	}
	function Cs(e, t, n, r, i) {
		var a = j.p;
		j.p = a !== 0 && 8 > a ? a : 8;
		var o = A.T, s = {};
		A.T = s, Fs(e, !1, t, n);
		try {
			var c = i(), l = A.S;
			l !== null && l(s, c), typeof c == "object" && c && typeof c.then == "function" ? Ps(e, t, _a(c, r), pu(e)) : Ps(e, t, r, pu(e));
		} catch (n) {
			Ps(e, t, {
				then: function() {},
				status: "rejected",
				reason: n
			}, pu());
		} finally {
			j.p = a, o !== null && s.types !== null && (o.types = s.types), A.T = o;
		}
	}
	function ws() {}
	function Ts(e, t, n, r) {
		if (e.tag !== 5) throw Error(o(476));
		var i = Es(e).queue;
		Cs(e, i, t, ae, n === null ? ws : function() {
			return Ds(e), n(r);
		});
	}
	function Es(e) {
		var t = e.memoizedState;
		if (t !== null) return t;
		t = {
			memoizedState: ae,
			baseState: ae,
			baseQueue: null,
			queue: {
				pending: null,
				lanes: 0,
				dispatch: null,
				lastRenderedReducer: Io,
				lastRenderedState: ae
			},
			next: null
		};
		var n = {};
		return t.next = {
			memoizedState: n,
			baseState: n,
			baseQueue: null,
			queue: {
				pending: null,
				lanes: 0,
				dispatch: null,
				lastRenderedReducer: Io,
				lastRenderedState: n
			},
			next: null
		}, e.memoizedState = t, e = e.alternate, e !== null && (e.memoizedState = t), t;
	}
	function Ds(e) {
		var t = Es(e);
		t.next === null && (t = e.alternate.memoizedState), Ps(e, t.next.queue, {}, pu());
	}
	function Os() {
		return na(Qf);
	}
	function ks() {
		return jo().memoizedState;
	}
	function As() {
		return jo().memoizedState;
	}
	function js(e) {
		for (var t = e.return; t !== null;) {
			switch (t.tag) {
				case 24:
				case 3:
					var n = pu();
					e = Wa(n);
					var r = Ga(t, e, n);
					r !== null && (hu(r, t, n), Ka(r, t, n)), t = { cache: la() }, e.payload = t;
					return;
			}
			t = t.return;
		}
	}
	function Ms(e, t, n) {
		var r = pu();
		n = {
			lane: r,
			revertLane: 0,
			gesture: null,
			action: n,
			hasEagerState: !1,
			eagerState: null,
			next: null
		}, Is(e) ? Ls(t, n) : (n = ri(e, t, n, r), n !== null && (hu(n, e, r), Rs(n, t, r)));
	}
	function Ns(e, t, n) {
		Ps(e, t, n, pu());
	}
	function Ps(e, t, n, r) {
		var i = {
			lane: r,
			revertLane: 0,
			gesture: null,
			action: n,
			hasEagerState: !1,
			eagerState: null,
			next: null
		};
		if (Is(e)) Ls(t, i);
		else {
			var a = e.alternate;
			if (e.lanes === 0 && (a === null || a.lanes === 0) && (a = t.lastRenderedReducer, a !== null)) try {
				var o = t.lastRenderedState, s = a(o, n);
				if (i.hasEagerState = !0, i.eagerState = s, Cr(s, o)) return ni(e, t, i, 0), q === null && ti(), !1;
			} catch {}
			if (n = ri(e, t, i, r), n !== null) return hu(n, e, r), Rs(n, t, r), !0;
		}
		return !1;
	}
	function Fs(e, t, n, r) {
		if (r = {
			lane: 2,
			revertLane: dd(),
			gesture: null,
			action: r,
			hasEagerState: !1,
			eagerState: null,
			next: null
		}, Is(e)) {
			if (t) throw Error(o(479));
		} else t = ri(e, n, r, 2), t !== null && hu(t, e, 2);
	}
	function Is(e) {
		var t = e.alternate;
		return e === B || t !== null && t === B;
	}
	function Ls(e, t) {
		ho = mo = !0;
		var n = e.pending;
		n === null ? t.next = t : (t.next = n.next, n.next = t), e.pending = t;
	}
	function Rs(e, t, n) {
		if (n & 4194048) {
			var r = t.lanes;
			r &= e.pendingLanes, n |= r, t.lanes = n, tt(e, n);
		}
	}
	var U = {
		readContext: na,
		use: Po,
		useCallback: xo,
		useContext: xo,
		useEffect: xo,
		useImperativeHandle: xo,
		useLayoutEffect: xo,
		useInsertionEffect: xo,
		useMemo: xo,
		useReducer: xo,
		useRef: xo,
		useState: xo,
		useDebugValue: xo,
		useDeferredValue: xo,
		useTransition: xo,
		useSyncExternalStore: xo,
		useId: xo,
		useHostTransitionStatus: xo,
		useFormState: xo,
		useActionState: xo,
		useOptimistic: xo,
		useMemoCache: xo,
		useCacheRefresh: xo
	};
	U.useEffectEvent = xo;
	var zs = {
		readContext: na,
		use: Po,
		useCallback: function(e, t) {
			return Ao().memoizedState = [e, t === void 0 ? null : t], e;
		},
		useContext: na,
		useEffect: us,
		useImperativeHandle: function(e, t, n) {
			n = n == null ? null : n.concat([e]), cs(4194308, 4, gs.bind(null, t, e), n);
		},
		useLayoutEffect: function(e, t) {
			return cs(4194308, 4, e, t);
		},
		useInsertionEffect: function(e, t) {
			cs(4, 2, e, t);
		},
		useMemo: function(e, t) {
			var n = Ao();
			t = t === void 0 ? null : t;
			var r = e();
			if (go) {
				Re(!0);
				try {
					e();
				} finally {
					Re(!1);
				}
			}
			return n.memoizedState = [r, t], r;
		},
		useReducer: function(e, t, n) {
			var r = Ao();
			if (n !== void 0) {
				var i = n(t);
				if (go) {
					Re(!0);
					try {
						n(t);
					} finally {
						Re(!1);
					}
				}
			} else i = t;
			return r.memoizedState = r.baseState = i, e = {
				pending: null,
				lanes: 0,
				dispatch: null,
				lastRenderedReducer: e,
				lastRenderedState: i
			}, r.queue = e, e = e.dispatch = Ms.bind(null, B, e), [r.memoizedState, e];
		},
		useRef: function(e) {
			var t = Ao();
			return e = { current: e }, t.memoizedState = e;
		},
		useState: function(e) {
			e = Ko(e);
			var t = e.queue, n = Ns.bind(null, B, t);
			return t.dispatch = n, [e.memoizedState, n];
		},
		useDebugValue: vs,
		useDeferredValue: function(e, t) {
			return xs(Ao(), e, t);
		},
		useTransition: function() {
			var e = Ko(!1);
			return e = Cs.bind(null, B, e.queue, !0, !1), Ao().memoizedState = e, [!1, e];
		},
		useSyncExternalStore: function(e, t, n) {
			var r = B, i = Ao();
			if (R) {
				if (n === void 0) throw Error(o(407));
				n = n();
			} else {
				if (n = t(), q === null) throw Error(o(349));
				Y & 127 || Vo(r, t, n);
			}
			i.memoizedState = n;
			var a = {
				value: n,
				getSnapshot: t
			};
			return i.queue = a, us(Uo.bind(null, r, a, e), [e]), r.flags |= 2048, os(9, { destroy: void 0 }, Ho.bind(null, r, a, n, t), null), n;
		},
		useId: function() {
			var e = Ao(), t = q.identifierPrefix;
			if (R) {
				var n = Oi, r = Di;
				n = (r & ~(1 << 32 - ze(r) - 1)).toString(32) + n, t = "_" + t + "R_" + n, n = _o++, 0 < n && (t += "H" + n.toString(32)), t += "_";
			} else n = bo++, t = "_" + t + "r_" + n.toString(32) + "_";
			return e.memoizedState = t;
		},
		useHostTransitionStatus: Os,
		useFormState: ts,
		useActionState: ts,
		useOptimistic: function(e) {
			var t = Ao();
			t.memoizedState = t.baseState = e;
			var n = {
				pending: null,
				lanes: 0,
				dispatch: null,
				lastRenderedReducer: null,
				lastRenderedState: null
			};
			return t.queue = n, t = Fs.bind(null, B, !0, n), n.dispatch = t, [e, t];
		},
		useMemoCache: Fo,
		useCacheRefresh: function() {
			return Ao().memoizedState = js.bind(null, B);
		},
		useEffectEvent: function(e) {
			var t = Ao(), n = { impl: e };
			return t.memoizedState = n, function() {
				if (K & 2) throw Error(o(440));
				return n.impl.apply(void 0, arguments);
			};
		}
	}, Bs = {
		readContext: na,
		use: Po,
		useCallback: ys,
		useContext: na,
		useEffect: ds,
		useImperativeHandle: _s,
		useInsertionEffect: ms,
		useLayoutEffect: hs,
		useMemo: bs,
		useReducer: Lo,
		useRef: ss,
		useState: function() {
			return Lo(Io);
		},
		useDebugValue: vs,
		useDeferredValue: function(e, t) {
			return Ss(jo(), V.memoizedState, e, t);
		},
		useTransition: function() {
			var e = Lo(Io)[0], t = jo().memoizedState;
			return [typeof e == "boolean" ? e : No(e), t];
		},
		useSyncExternalStore: Bo,
		useId: ks,
		useHostTransitionStatus: Os,
		useFormState: ns,
		useActionState: ns,
		useOptimistic: function(e, t) {
			return qo(jo(), V, e, t);
		},
		useMemoCache: Fo,
		useCacheRefresh: As
	};
	Bs.useEffectEvent = ps;
	var Vs = {
		readContext: na,
		use: Po,
		useCallback: ys,
		useContext: na,
		useEffect: ds,
		useImperativeHandle: _s,
		useInsertionEffect: ms,
		useLayoutEffect: hs,
		useMemo: bs,
		useReducer: zo,
		useRef: ss,
		useState: function() {
			return zo(Io);
		},
		useDebugValue: vs,
		useDeferredValue: function(e, t) {
			var n = jo();
			return V === null ? xs(n, e, t) : Ss(n, V.memoizedState, e, t);
		},
		useTransition: function() {
			var e = zo(Io)[0], t = jo().memoizedState;
			return [typeof e == "boolean" ? e : No(e), t];
		},
		useSyncExternalStore: Bo,
		useId: ks,
		useHostTransitionStatus: Os,
		useFormState: as,
		useActionState: as,
		useOptimistic: function(e, t) {
			var n = jo();
			return V === null ? (n.baseState = e, [e, n.queue.dispatch]) : qo(n, V, e, t);
		},
		useMemoCache: Fo,
		useCacheRefresh: As
	};
	Vs.useEffectEvent = ps;
	function Hs(e, t, n, r) {
		t = e.memoizedState, n = n(r, t), n = n == null ? t : m({}, t, n), e.memoizedState = n, e.lanes === 0 && (e.updateQueue.baseState = n);
	}
	var Us = {
		enqueueSetState: function(e, t, n) {
			e = e._reactInternals;
			var r = pu(), i = Wa(r);
			i.payload = t, n != null && (i.callback = n), t = Ga(e, i, r), t !== null && (hu(t, e, r), Ka(t, e, r));
		},
		enqueueReplaceState: function(e, t, n) {
			e = e._reactInternals;
			var r = pu(), i = Wa(r);
			i.tag = 1, i.payload = t, n != null && (i.callback = n), t = Ga(e, i, r), t !== null && (hu(t, e, r), Ka(t, e, r));
		},
		enqueueForceUpdate: function(e, t) {
			e = e._reactInternals;
			var n = pu(), r = Wa(n);
			r.tag = 2, t != null && (r.callback = t), t = Ga(e, r, n), t !== null && (hu(t, e, n), Ka(t, e, n));
		}
	};
	function Ws(e, t, n, r, i, a, o) {
		return e = e.stateNode, typeof e.shouldComponentUpdate == "function" ? e.shouldComponentUpdate(r, a, o) : t.prototype && t.prototype.isPureReactComponent ? !wr(n, r) || !wr(i, a) : !0;
	}
	function Gs(e, t, n, r) {
		e = t.state, typeof t.componentWillReceiveProps == "function" && t.componentWillReceiveProps(n, r), typeof t.UNSAFE_componentWillReceiveProps == "function" && t.UNSAFE_componentWillReceiveProps(n, r), t.state !== e && Us.enqueueReplaceState(t, t.state, null);
	}
	function Ks(e, t) {
		var n = t;
		if ("ref" in t) for (var r in n = {}, t) r !== "ref" && (n[r] = t[r]);
		if (e = e.defaultProps) for (var i in n === t && (n = m({}, n)), e) n[i] === void 0 && (n[i] = e[i]);
		return n;
	}
	function qs(e) {
		Zr(e);
	}
	function Js(e) {
		console.error(e);
	}
	function Ys(e) {
		Zr(e);
	}
	function Xs(e, t) {
		try {
			var n = e.onUncaughtError;
			n(t.value, { componentStack: t.stack });
		} catch (e) {
			setTimeout(function() {
				throw e;
			});
		}
	}
	function Zs(e, t, n) {
		try {
			var r = e.onCaughtError;
			r(n.value, {
				componentStack: n.stack,
				errorBoundary: t.tag === 1 ? t.stateNode : null
			});
		} catch (e) {
			setTimeout(function() {
				throw e;
			});
		}
	}
	function Qs(e, t, n) {
		return n = Wa(n), n.tag = 3, n.payload = { element: null }, n.callback = function() {
			Xs(e, t);
		}, n;
	}
	function $s(e) {
		return e = Wa(e), e.tag = 3, e;
	}
	function ec(e, t, n, r) {
		var i = n.type.getDerivedStateFromError;
		if (typeof i == "function") {
			var a = r.value;
			e.payload = function() {
				return i(a);
			}, e.callback = function() {
				Zs(t, n, r);
			};
		}
		var o = n.stateNode;
		o !== null && typeof o.componentDidCatch == "function" && (e.callback = function() {
			Zs(t, n, r), typeof i != "function" && (ru === null ? ru = /* @__PURE__ */ new Set([this]) : ru.add(this));
			var e = r.stack;
			this.componentDidCatch(r.value, { componentStack: e === null ? "" : e });
		});
	}
	function tc(e, t, n, r, i) {
		if (n.flags |= 32768, typeof r == "object" && r && typeof r.then == "function") {
			if (t = n.alternate, t !== null && $i(t, n, i, !0), n = io.current, n !== null) {
				switch (n.tag) {
					case 31:
					case 13: return ao === null ? Du() : n.alternate === null && Wl === 0 && (Wl = 3), n.flags &= -257, n.flags |= 65536, n.lanes = i, r === Ea ? n.flags |= 16384 : (t = n.updateQueue, t === null ? n.updateQueue = /* @__PURE__ */ new Set([r]) : t.add(r), Gu(e, r, i)), !1;
					case 22: return n.flags |= 65536, r === Ea ? n.flags |= 16384 : (t = n.updateQueue, t === null ? (t = {
						transitions: null,
						markerInstances: null,
						retryQueue: /* @__PURE__ */ new Set([r])
					}, n.updateQueue = t) : (n = t.retryQueue, n === null ? t.retryQueue = /* @__PURE__ */ new Set([r]) : n.add(r)), Gu(e, r, i)), !1;
				}
				throw Error(o(435, n.tag));
			}
			return Gu(e, r, i), Du(), !1;
		}
		if (R) return t = io.current, t === null ? (r !== Ri && (t = Error(o(423), { cause: r }), Gi(yi(t, n))), e = e.current.alternate, e.flags |= 65536, i &= -i, e.lanes |= i, r = yi(r, n), i = Qs(e.stateNode, r, i), qa(e, i), Wl !== 4 && (Wl = 2)) : (!(t.flags & 65536) && (t.flags |= 256), t.flags |= 65536, t.lanes = i, r !== Ri && (e = Error(o(422), { cause: r }), Gi(yi(e, n)))), !1;
		var a = Error(o(520), { cause: r });
		if (a = yi(a, n), Xl === null ? Xl = [a] : Xl.push(a), Wl !== 4 && (Wl = 2), t === null) return !0;
		r = yi(r, n), n = t;
		do {
			switch (n.tag) {
				case 3: return n.flags |= 65536, e = i & -i, n.lanes |= e, e = Qs(n.stateNode, r, e), qa(n, e), !1;
				case 1: if (t = n.type, a = n.stateNode, !(n.flags & 128) && (typeof t.getDerivedStateFromError == "function" || a !== null && typeof a.componentDidCatch == "function" && (ru === null || !ru.has(a)))) return n.flags |= 65536, i &= -i, n.lanes |= i, i = $s(i), ec(i, e, n, r), qa(n, i), !1;
			}
			n = n.return;
		} while (n !== null);
		return !1;
	}
	var nc = Error(o(461)), rc = !1;
	function ic(e, t, n, r) {
		t.child = e === null ? Ba(t, null, n, r) : za(t, e.child, n, r);
	}
	function ac(e, t, n, r, i) {
		n = n.render;
		var a = t.ref;
		if ("ref" in r) {
			var o = {};
			for (var s in r) s !== "ref" && (o[s] = r[s]);
		} else o = r;
		return ta(t), r = Co(e, t, n, o, a, i), s = Do(), e !== null && !rc ? (Oo(e, t, i), kc(e, t, i)) : (R && s && ji(t), t.flags |= 1, ic(e, t, r, i), t.child);
	}
	function oc(e, t, n, r, i) {
		if (e === null) {
			var a = n.type;
			return typeof a == "function" && !ui(a) && a.defaultProps === void 0 && n.compare === null ? (t.tag = 15, t.type = a, sc(e, t, a, r, i)) : (e = pi(n.type, null, r, t, t.mode, i), e.ref = t.ref, e.return = t, t.child = e);
		}
		if (a = e.child, !Ac(e, i)) {
			var o = a.memoizedProps;
			if (n = n.compare, n = n === null ? wr : n, n(o, r) && e.ref === t.ref) return kc(e, t, i);
		}
		return t.flags |= 1, e = di(a, r), e.ref = t.ref, e.return = t, t.child = e;
	}
	function sc(e, t, n, r, i) {
		if (e !== null) {
			var a = e.memoizedProps;
			if (wr(a, r) && e.ref === t.ref) {
				if (rc = !1, t.pendingProps = r = a, Ac(e, i)) e.flags & 131072 && (rc = !0);
				else return t.lanes = e.lanes, kc(e, t, i);
			}
		}
		return hc(e, t, n, r, i);
	}
	function cc(e, t, n, r) {
		var i = r.children, a = e === null ? null : e.memoizedState;
		if (e === null && t.stateNode === null && (t.stateNode = {
			_visibility: 1,
			_pendingMarkers: null,
			_retryCache: null,
			_transitions: null
		}), r.mode === "hidden") {
			if (t.flags & 128) {
				if (a = a === null ? n : a.baseLanes | n, e !== null) {
					for (r = t.child = e.child, i = 0; r !== null;) i = i | r.lanes | r.childLanes, r = r.sibling;
					r = i & ~a;
				} else r = 0, t.child = null;
				return uc(e, t, a, n, r);
			}
			if (n & 536870912) t.memoizedState = {
				baseLanes: 0,
				cachePool: null
			}, e !== null && xa(t, a === null ? null : a.cachePool), a === null ? no() : to(t, a), co(t);
			else return r = t.lanes = 536870912, uc(e, t, a === null ? n : a.baseLanes | n, n, r);
		} else a === null ? (e !== null && xa(t, null), no(), lo(t)) : (xa(t, a.cachePool), to(t, a), lo(t), t.memoizedState = null);
		return ic(e, t, i, n), t.child;
	}
	function lc(e, t) {
		return e !== null && e.tag === 22 || t.stateNode !== null || (t.stateNode = {
			_visibility: 1,
			_pendingMarkers: null,
			_retryCache: null,
			_transitions: null
		}), t.sibling;
	}
	function uc(e, t, n, r, i) {
		var a = ba();
		return a = a === null ? null : {
			parent: ca._currentValue,
			pool: a
		}, t.memoizedState = {
			baseLanes: n,
			cachePool: a
		}, e !== null && xa(t, null), no(), co(t), e !== null && $i(e, t, r, !0), t.childLanes = i, null;
	}
	function dc(e, t) {
		return t = wc({
			mode: t.mode,
			children: t.children
		}, e.mode), t.ref = e.ref, e.child = t, t.return = e, t;
	}
	function fc(e, t, n) {
		return za(t, e.child, null, n), e = dc(t, t.pendingProps), e.flags |= 2, uo(t), t.memoizedState = null, e;
	}
	function pc(e, t, n) {
		var r = t.pendingProps, i = !!(t.flags & 128);
		if (t.flags &= -129, e === null) {
			if (R) {
				if (r.mode === "hidden") return e = dc(t, r), t.lanes = 536870912, lc(null, e);
				if (so(t), (e = Fi) ? (e = rf(e, Li), e = e !== null && e.data === "&" ? e : null, e !== null && (t.memoizedState = {
					dehydrated: e,
					treeContext: Ei === null ? null : {
						id: Di,
						overflow: Oi
					},
					retryLane: 536870912,
					hydrationErrors: null
				}, n = gi(e), n.return = t, t.child = n, Pi = t, Fi = null)) : e = null, e === null) throw zi(t);
				return t.lanes = 536870912, null;
			}
			return dc(t, r);
		}
		var a = e.memoizedState;
		if (a !== null) {
			var s = a.dehydrated;
			if (so(t), i) {
				if (t.flags & 256) t.flags &= -257, t = fc(e, t, n);
				else if (t.memoizedState !== null) t.child = e.child, t.flags |= 128, t = null;
				else throw Error(o(558));
			} else if (rc || $i(e, t, n, !1), i = (n & e.childLanes) !== 0, rc || i) {
				if (r = q, r !== null && (s = nt(r, n), s !== 0 && s !== a.retryLane)) throw a.retryLane = s, ii(e, s), hu(r, e, s), nc;
				Du(), t = fc(e, t, n);
			} else e = a.treeContext, Fi = cf(s.nextSibling), Pi = t, R = !0, Ii = null, Li = !1, e !== null && Ni(t, e), t = dc(t, r), t.flags |= 4096;
			return t;
		}
		return e = di(e.child, {
			mode: r.mode,
			children: r.children
		}), e.ref = t.ref, t.child = e, e.return = t, e;
	}
	function mc(e, t) {
		var n = t.ref;
		if (n === null) e !== null && e.ref !== null && (t.flags |= 4194816);
		else {
			if (typeof n != "function" && typeof n != "object") throw Error(o(284));
			(e === null || e.ref !== n) && (t.flags |= 4194816);
		}
	}
	function hc(e, t, n, r, i) {
		return ta(t), n = Co(e, t, n, r, void 0, i), r = Do(), e !== null && !rc ? (Oo(e, t, i), kc(e, t, i)) : (R && r && ji(t), t.flags |= 1, ic(e, t, n, i), t.child);
	}
	function gc(e, t, n, r, i, a) {
		return ta(t), t.updateQueue = null, n = To(t, r, n, i), wo(e), r = Do(), e !== null && !rc ? (Oo(e, t, a), kc(e, t, a)) : (R && r && ji(t), t.flags |= 1, ic(e, t, n, a), t.child);
	}
	function _c(e, t, n, r, i) {
		if (ta(t), t.stateNode === null) {
			var a = si, o = n.contextType;
			typeof o == "object" && o && (a = na(o)), a = new n(r, a), t.memoizedState = a.state !== null && a.state !== void 0 ? a.state : null, a.updater = Us, t.stateNode = a, a._reactInternals = t, a = t.stateNode, a.props = r, a.state = t.memoizedState, a.refs = {}, Ha(t), o = n.contextType, a.context = typeof o == "object" && o ? na(o) : si, a.state = t.memoizedState, o = n.getDerivedStateFromProps, typeof o == "function" && (Hs(t, n, o, r), a.state = t.memoizedState), typeof n.getDerivedStateFromProps == "function" || typeof a.getSnapshotBeforeUpdate == "function" || typeof a.UNSAFE_componentWillMount != "function" && typeof a.componentWillMount != "function" || (o = a.state, typeof a.componentWillMount == "function" && a.componentWillMount(), typeof a.UNSAFE_componentWillMount == "function" && a.UNSAFE_componentWillMount(), o !== a.state && Us.enqueueReplaceState(a, a.state, null), Xa(t, r, a, i), Ya(), a.state = t.memoizedState), typeof a.componentDidMount == "function" && (t.flags |= 4194308), r = !0;
		} else if (e === null) {
			a = t.stateNode;
			var s = t.memoizedProps, c = Ks(n, s);
			a.props = c;
			var l = a.context, u = n.contextType;
			o = si, typeof u == "object" && u && (o = na(u));
			var d = n.getDerivedStateFromProps;
			u = typeof d == "function" || typeof a.getSnapshotBeforeUpdate == "function", s = t.pendingProps !== s, u || typeof a.UNSAFE_componentWillReceiveProps != "function" && typeof a.componentWillReceiveProps != "function" || (s || l !== o) && Gs(t, a, r, o), Va = !1;
			var f = t.memoizedState;
			a.state = f, Xa(t, r, a, i), Ya(), l = t.memoizedState, s || f !== l || Va ? (typeof d == "function" && (Hs(t, n, d, r), l = t.memoizedState), (c = Va || Ws(t, n, c, r, f, l, o)) ? (u || typeof a.UNSAFE_componentWillMount != "function" && typeof a.componentWillMount != "function" || (typeof a.componentWillMount == "function" && a.componentWillMount(), typeof a.UNSAFE_componentWillMount == "function" && a.UNSAFE_componentWillMount()), typeof a.componentDidMount == "function" && (t.flags |= 4194308)) : (typeof a.componentDidMount == "function" && (t.flags |= 4194308), t.memoizedProps = r, t.memoizedState = l), a.props = r, a.state = l, a.context = o, r = c) : (typeof a.componentDidMount == "function" && (t.flags |= 4194308), r = !1);
		} else {
			a = t.stateNode, Ua(e, t), o = t.memoizedProps, u = Ks(n, o), a.props = u, d = t.pendingProps, f = a.context, l = n.contextType, c = si, typeof l == "object" && l && (c = na(l)), s = n.getDerivedStateFromProps, (l = typeof s == "function" || typeof a.getSnapshotBeforeUpdate == "function") || typeof a.UNSAFE_componentWillReceiveProps != "function" && typeof a.componentWillReceiveProps != "function" || (o !== d || f !== c) && Gs(t, a, r, c), Va = !1, f = t.memoizedState, a.state = f, Xa(t, r, a, i), Ya();
			var p = t.memoizedState;
			o !== d || f !== p || Va || e !== null && e.dependencies !== null && ea(e.dependencies) ? (typeof s == "function" && (Hs(t, n, s, r), p = t.memoizedState), (u = Va || Ws(t, n, u, r, f, p, c) || e !== null && e.dependencies !== null && ea(e.dependencies)) ? (l || typeof a.UNSAFE_componentWillUpdate != "function" && typeof a.componentWillUpdate != "function" || (typeof a.componentWillUpdate == "function" && a.componentWillUpdate(r, p, c), typeof a.UNSAFE_componentWillUpdate == "function" && a.UNSAFE_componentWillUpdate(r, p, c)), typeof a.componentDidUpdate == "function" && (t.flags |= 4), typeof a.getSnapshotBeforeUpdate == "function" && (t.flags |= 1024)) : (typeof a.componentDidUpdate != "function" || o === e.memoizedProps && f === e.memoizedState || (t.flags |= 4), typeof a.getSnapshotBeforeUpdate != "function" || o === e.memoizedProps && f === e.memoizedState || (t.flags |= 1024), t.memoizedProps = r, t.memoizedState = p), a.props = r, a.state = p, a.context = c, r = u) : (typeof a.componentDidUpdate != "function" || o === e.memoizedProps && f === e.memoizedState || (t.flags |= 4), typeof a.getSnapshotBeforeUpdate != "function" || o === e.memoizedProps && f === e.memoizedState || (t.flags |= 1024), r = !1);
		}
		return a = r, mc(e, t), r = !!(t.flags & 128), a || r ? (a = t.stateNode, n = r && typeof n.getDerivedStateFromError != "function" ? null : a.render(), t.flags |= 1, e !== null && r ? (t.child = za(t, e.child, null, i), t.child = za(t, null, n, i)) : ic(e, t, n, i), t.memoizedState = a.state, e = t.child) : e = kc(e, t, i), e;
	}
	function vc(e, t, n, r) {
		return Ui(), t.flags |= 256, ic(e, t, n, r), t.child;
	}
	var yc = {
		dehydrated: null,
		treeContext: null,
		retryLane: 0,
		hydrationErrors: null
	};
	function bc(e) {
		return {
			baseLanes: e,
			cachePool: Sa()
		};
	}
	function xc(e, t, n) {
		return e = e === null ? 0 : e.childLanes & ~n, t && (e |= Jl), e;
	}
	function Sc(e, t, n) {
		var r = t.pendingProps, i = !1, a = !!(t.flags & 128), s;
		if ((s = a) || (s = e !== null && e.memoizedState === null ? !1 : !!(fo.current & 2)), s && (i = !0, t.flags &= -129), s = !!(t.flags & 32), t.flags &= -33, e === null) {
			if (R) {
				if (i ? oo(t) : lo(t), (e = Fi) ? (e = rf(e, Li), e = e !== null && e.data !== "&" ? e : null, e !== null && (t.memoizedState = {
					dehydrated: e,
					treeContext: Ei === null ? null : {
						id: Di,
						overflow: Oi
					},
					retryLane: 536870912,
					hydrationErrors: null
				}, n = gi(e), n.return = t, t.child = n, Pi = t, Fi = null)) : e = null, e === null) throw zi(t);
				return of(e) ? t.lanes = 32 : t.lanes = 536870912, null;
			}
			var c = r.children;
			return r = r.fallback, i ? (lo(t), i = t.mode, c = wc({
				mode: "hidden",
				children: c
			}, i), r = mi(r, i, n, null), c.return = t, r.return = t, c.sibling = r, t.child = c, r = t.child, r.memoizedState = bc(n), r.childLanes = xc(e, s, n), t.memoizedState = yc, lc(null, r)) : (oo(t), Cc(t, c));
		}
		var l = e.memoizedState;
		if (l !== null && (c = l.dehydrated, c !== null)) {
			if (a) t.flags & 256 ? (oo(t), t.flags &= -257, t = Tc(e, t, n)) : t.memoizedState === null ? (lo(t), c = r.fallback, i = t.mode, r = wc({
				mode: "visible",
				children: r.children
			}, i), c = mi(c, i, n, null), c.flags |= 2, r.return = t, c.return = t, r.sibling = c, t.child = r, za(t, e.child, null, n), r = t.child, r.memoizedState = bc(n), r.childLanes = xc(e, s, n), t.memoizedState = yc, t = lc(null, r)) : (lo(t), t.child = e.child, t.flags |= 128, t = null);
			else if (oo(t), of(c)) {
				if (s = c.nextSibling && c.nextSibling.dataset, s) var u = s.dgst;
				s = u, r = Error(o(419)), r.stack = "", r.digest = s, Gi({
					value: r,
					source: null,
					stack: null
				}), t = Tc(e, t, n);
			} else if (rc || $i(e, t, n, !1), s = (n & e.childLanes) !== 0, rc || s) {
				if (s = q, s !== null && (r = nt(s, n), r !== 0 && r !== l.retryLane)) throw l.retryLane = r, ii(e, r), hu(s, e, r), nc;
				af(c) || Du(), t = Tc(e, t, n);
			} else af(c) ? (t.flags |= 192, t.child = e.child, t = null) : (e = l.treeContext, Fi = cf(c.nextSibling), Pi = t, R = !0, Ii = null, Li = !1, e !== null && Ni(t, e), t = Cc(t, r.children), t.flags |= 4096);
			return t;
		}
		return i ? (lo(t), c = r.fallback, i = t.mode, l = e.child, u = l.sibling, r = di(l, {
			mode: "hidden",
			children: r.children
		}), r.subtreeFlags = l.subtreeFlags & 65011712, u === null ? (c = mi(c, i, n, null), c.flags |= 2) : c = di(u, c), c.return = t, r.return = t, r.sibling = c, t.child = r, lc(null, r), r = t.child, c = e.child.memoizedState, c === null ? c = bc(n) : (i = c.cachePool, i === null ? i = Sa() : (l = ca._currentValue, i = i.parent === l ? i : {
			parent: l,
			pool: l
		}), c = {
			baseLanes: c.baseLanes | n,
			cachePool: i
		}), r.memoizedState = c, r.childLanes = xc(e, s, n), t.memoizedState = yc, lc(e.child, r)) : (oo(t), n = e.child, e = n.sibling, n = di(n, {
			mode: "visible",
			children: r.children
		}), n.return = t, n.sibling = null, e !== null && (s = t.deletions, s === null ? (t.deletions = [e], t.flags |= 16) : s.push(e)), t.child = n, t.memoizedState = null, n);
	}
	function Cc(e, t) {
		return t = wc({
			mode: "visible",
			children: t
		}, e.mode), t.return = e, e.child = t;
	}
	function wc(e, t) {
		return e = li(22, e, null, t), e.lanes = 0, e;
	}
	function Tc(e, t, n) {
		return za(t, e.child, null, n), e = Cc(t, t.pendingProps.children), e.flags |= 2, t.memoizedState = null, e;
	}
	function Ec(e, t, n) {
		e.lanes |= t;
		var r = e.alternate;
		r !== null && (r.lanes |= t), Zi(e.return, t, n);
	}
	function Dc(e, t, n, r, i, a) {
		var o = e.memoizedState;
		o === null ? e.memoizedState = {
			isBackwards: t,
			rendering: null,
			renderingStartTime: 0,
			last: r,
			tail: n,
			tailMode: i,
			treeForkCount: a
		} : (o.isBackwards = t, o.rendering = null, o.renderingStartTime = 0, o.last = r, o.tail = n, o.tailMode = i, o.treeForkCount = a);
	}
	function Oc(e, t, n) {
		var r = t.pendingProps, i = r.revealOrder, a = r.tail;
		r = r.children;
		var o = fo.current, s = !!(o & 2);
		if (s ? (o = o & 1 | 2, t.flags |= 128) : o &= 1, N(fo, o), ic(e, t, r, n), r = R ? Ci : 0, !s && e !== null && e.flags & 128) a: for (e = t.child; e !== null;) {
			if (e.tag === 13) e.memoizedState !== null && Ec(e, n, t);
			else if (e.tag === 19) Ec(e, n, t);
			else if (e.child !== null) {
				e.child.return = e, e = e.child;
				continue;
			}
			if (e === t) break a;
			for (; e.sibling === null;) {
				if (e.return === null || e.return === t) break a;
				e = e.return;
			}
			e.sibling.return = e.return, e = e.sibling;
		}
		switch (i) {
			case "forwards":
				for (n = t.child, i = null; n !== null;) e = n.alternate, e !== null && po(e) === null && (i = n), n = n.sibling;
				n = i, n === null ? (i = t.child, t.child = null) : (i = n.sibling, n.sibling = null), Dc(t, !1, i, n, a, r);
				break;
			case "backwards":
			case "unstable_legacy-backwards":
				for (n = null, i = t.child, t.child = null; i !== null;) {
					if (e = i.alternate, e !== null && po(e) === null) {
						t.child = i;
						break;
					}
					e = i.sibling, i.sibling = n, n = i, i = e;
				}
				Dc(t, !0, n, null, a, r);
				break;
			case "together":
				Dc(t, !1, null, null, void 0, r);
				break;
			default: t.memoizedState = null;
		}
		return t.child;
	}
	function kc(e, t, n) {
		if (e !== null && (t.dependencies = e.dependencies), Gl |= t.lanes, (n & t.childLanes) === 0) {
			if (e !== null) {
				if ($i(e, t, n, !1), (n & t.childLanes) === 0) return null;
			} else return null;
		}
		if (e !== null && t.child !== e.child) throw Error(o(153));
		if (t.child !== null) {
			for (e = t.child, n = di(e, e.pendingProps), t.child = n, n.return = t; e.sibling !== null;) e = e.sibling, n = n.sibling = di(e, e.pendingProps), n.return = t;
			n.sibling = null;
		}
		return t.child;
	}
	function Ac(e, t) {
		return (e.lanes & t) !== 0 || (e = e.dependencies, !!(e !== null && ea(e)));
	}
	function jc(e, t, n) {
		switch (t.tag) {
			case 3:
				pe(t, t.stateNode.containerInfo), Yi(t, ca, e.memoizedState.cache), Ui();
				break;
			case 27:
			case 5:
				he(t);
				break;
			case 4:
				pe(t, t.stateNode.containerInfo);
				break;
			case 10:
				Yi(t, t.type, t.memoizedProps.value);
				break;
			case 31:
				if (t.memoizedState !== null) return t.flags |= 128, so(t), null;
				break;
			case 13:
				var r = t.memoizedState;
				if (r !== null) return r.dehydrated === null ? (n & t.child.childLanes) === 0 ? (oo(t), e = kc(e, t, n), e === null ? null : e.sibling) : Sc(e, t, n) : (oo(t), t.flags |= 128, null);
				oo(t);
				break;
			case 19:
				var i = !!(e.flags & 128);
				if (r = (n & t.childLanes) !== 0, r ||= ($i(e, t, n, !1), (n & t.childLanes) !== 0), i) {
					if (r) return Oc(e, t, n);
					t.flags |= 128;
				}
				if (i = t.memoizedState, i !== null && (i.rendering = null, i.tail = null, i.lastEffect = null), N(fo, fo.current), r) break;
				return null;
			case 22: return t.lanes = 0, cc(e, t, n, t.pendingProps);
			case 24: Yi(t, ca, e.memoizedState.cache);
		}
		return kc(e, t, n);
	}
	function Mc(e, t, n) {
		if (e !== null) {
			if (e.memoizedProps !== t.pendingProps) rc = !0;
			else {
				if (!Ac(e, n) && !(t.flags & 128)) return rc = !1, jc(e, t, n);
				rc = !!(e.flags & 131072);
			}
		} else rc = !1, R && t.flags & 1048576 && Ai(t, Ci, t.index);
		switch (t.lanes = 0, t.tag) {
			case 16:
				a: {
					var r = t.pendingProps;
					if (e = ka(t.elementType), t.type = e, typeof e == "function") ui(e) ? (r = Ks(e, r), t.tag = 1, t = _c(null, t, e, r, n)) : (t.tag = 0, t = hc(null, t, e, r, n));
					else {
						if (e != null) {
							var i = e.$$typeof;
							if (i === C) {
								t.tag = 11, t = ac(null, t, e, r, n);
								break a;
							}
							if (i === E) {
								t.tag = 14, t = oc(null, t, e, r, n);
								break a;
							}
						}
						throw t = ie(e) || e, Error(o(306, t, ""));
					}
				}
				return t;
			case 0: return hc(e, t, t.type, t.pendingProps, n);
			case 1: return r = t.type, i = Ks(r, t.pendingProps), _c(e, t, r, i, n);
			case 3:
				a: {
					if (pe(t, t.stateNode.containerInfo), e === null) throw Error(o(387));
					r = t.pendingProps;
					var a = t.memoizedState;
					i = a.element, Ua(e, t), Xa(t, r, null, n);
					var s = t.memoizedState;
					if (r = s.cache, Yi(t, ca, r), r !== a.cache && Qi(t, [ca], n, !0), Ya(), r = s.element, a.isDehydrated) {
						if (a = {
							element: r,
							isDehydrated: !1,
							cache: s.cache
						}, t.updateQueue.baseState = a, t.memoizedState = a, t.flags & 256) {
							t = vc(e, t, r, n);
							break a;
						}
						if (r !== i) {
							i = yi(Error(o(424)), t), Gi(i), t = vc(e, t, r, n);
							break a;
						}
						switch (e = t.stateNode.containerInfo, e.nodeType) {
							case 9:
								e = e.body;
								break;
							default: e = e.nodeName === "HTML" ? e.ownerDocument.body : e;
						}
						for (Fi = cf(e.firstChild), Pi = t, R = !0, Ii = null, Li = !0, n = Ba(t, null, r, n), t.child = n; n;) n.flags = n.flags & -3 | 4096, n = n.sibling;
					} else {
						if (Ui(), r === i) {
							t = kc(e, t, n);
							break a;
						}
						ic(e, t, r, n);
					}
					t = t.child;
				}
				return t;
			case 26: return mc(e, t), e === null ? (n = kf(t.type, null, t.pendingProps, null)) ? t.memoizedState = n : R || (n = t.type, e = t.pendingProps, r = Bd(de.current).createElement(n), r[ct] = t, r[lt] = e, Pd(r, n, e), xt(r), t.stateNode = r) : t.memoizedState = kf(t.type, e.memoizedProps, t.pendingProps, e.memoizedState), null;
			case 27: return he(t), e === null && R && (r = t.stateNode = ff(t.type, t.pendingProps, de.current), Pi = t, Li = !0, i = Fi, Zd(t.type) ? (lf = i, Fi = cf(r.firstChild)) : Fi = i), ic(e, t, t.pendingProps.children, n), mc(e, t), e === null && (t.flags |= 4194304), t.child;
			case 5: return e === null && R && ((i = r = Fi) && (r = tf(r, t.type, t.pendingProps, Li), r === null ? i = !1 : (t.stateNode = r, Pi = t, Fi = cf(r.firstChild), Li = !1, i = !0)), i || zi(t)), he(t), i = t.type, a = t.pendingProps, s = e === null ? null : e.memoizedProps, r = a.children, Ud(i, a) ? r = null : s !== null && Ud(i, s) && (t.flags |= 32), t.memoizedState !== null && (i = Co(e, t, Eo, null, null, n), Qf._currentValue = i), mc(e, t), ic(e, t, r, n), t.child;
			case 6: return e === null && R && ((e = n = Fi) && (n = nf(n, t.pendingProps, Li), n === null ? e = !1 : (t.stateNode = n, Pi = t, Fi = null, e = !0)), e || zi(t)), null;
			case 13: return Sc(e, t, n);
			case 4: return pe(t, t.stateNode.containerInfo), r = t.pendingProps, e === null ? t.child = za(t, null, r, n) : ic(e, t, r, n), t.child;
			case 11: return ac(e, t, t.type, t.pendingProps, n);
			case 7: return ic(e, t, t.pendingProps, n), t.child;
			case 8: return ic(e, t, t.pendingProps.children, n), t.child;
			case 12: return ic(e, t, t.pendingProps.children, n), t.child;
			case 10: return r = t.pendingProps, Yi(t, t.type, r.value), ic(e, t, r.children, n), t.child;
			case 9: return i = t.type._context, r = t.pendingProps.children, ta(t), i = na(i), r = r(i), t.flags |= 1, ic(e, t, r, n), t.child;
			case 14: return oc(e, t, t.type, t.pendingProps, n);
			case 15: return sc(e, t, t.type, t.pendingProps, n);
			case 19: return Oc(e, t, n);
			case 31: return pc(e, t, n);
			case 22: return cc(e, t, n, t.pendingProps);
			case 24: return ta(t), r = na(ca), e === null ? (i = ba(), i === null && (i = q, a = la(), i.pooledCache = a, a.refCount++, a !== null && (i.pooledCacheLanes |= n), i = a), t.memoizedState = {
				parent: r,
				cache: i
			}, Ha(t), Yi(t, ca, i)) : ((e.lanes & n) !== 0 && (Ua(e, t), Xa(t, null, null, n), Ya()), i = e.memoizedState, a = t.memoizedState, i.parent === r ? (r = a.cache, Yi(t, ca, r), r !== i.cache && Qi(t, [ca], n, !0)) : (i = {
				parent: r,
				cache: r
			}, t.memoizedState = i, t.lanes === 0 && (t.memoizedState = t.updateQueue.baseState = i), Yi(t, ca, r))), ic(e, t, t.pendingProps.children, n), t.child;
			case 29: throw t.pendingProps;
		}
		throw Error(o(156, t.tag));
	}
	function Nc(e) {
		e.flags |= 4;
	}
	function Pc(e, t, n, r, i) {
		if ((t = !!(e.mode & 32)) && (t = !1), t) {
			if (e.flags |= 16777216, (i & 335544128) === i) {
				if (e.stateNode.complete) e.flags |= 8192;
				else if (wu()) e.flags |= 8192;
				else throw Aa = Ea, wa;
			}
		} else e.flags &= -16777217;
	}
	function Fc(e, t) {
		if (t.type !== "stylesheet" || t.state.loading & 4) e.flags &= -16777217;
		else if (e.flags |= 16777216, !Wf(t)) {
			if (wu()) e.flags |= 8192;
			else throw Aa = Ea, wa;
		}
	}
	function Ic(e, t) {
		t !== null && (e.flags |= 4), e.flags & 16384 && (t = e.tag === 22 ? 536870912 : Xe(), e.lanes |= t, Yl |= t);
	}
	function Lc(e, t) {
		if (!R) switch (e.tailMode) {
			case "hidden":
				t = e.tail;
				for (var n = null; t !== null;) t.alternate !== null && (n = t), t = t.sibling;
				n === null ? e.tail = null : n.sibling = null;
				break;
			case "collapsed":
				n = e.tail;
				for (var r = null; n !== null;) n.alternate !== null && (r = n), n = n.sibling;
				r === null ? t || e.tail === null ? e.tail = null : e.tail.sibling = null : r.sibling = null;
		}
	}
	function W(e) {
		var t = e.alternate !== null && e.alternate.child === e.child, n = 0, r = 0;
		if (t) for (var i = e.child; i !== null;) n |= i.lanes | i.childLanes, r |= i.subtreeFlags & 65011712, r |= i.flags & 65011712, i.return = e, i = i.sibling;
		else for (i = e.child; i !== null;) n |= i.lanes | i.childLanes, r |= i.subtreeFlags, r |= i.flags, i.return = e, i = i.sibling;
		return e.subtreeFlags |= r, e.childLanes = n, t;
	}
	function Rc(e, t, n) {
		var r = t.pendingProps;
		switch (Mi(t), t.tag) {
			case 16:
			case 15:
			case 0:
			case 11:
			case 7:
			case 8:
			case 12:
			case 9:
			case 14: return W(t), null;
			case 1: return W(t), null;
			case 3: return n = t.stateNode, r = null, e !== null && (r = e.memoizedState.cache), t.memoizedState.cache !== r && (t.flags |= 2048), Xi(ca), me(), n.pendingContext && (n.context = n.pendingContext, n.pendingContext = null), (e === null || e.child === null) && (Hi(t) ? Nc(t) : e === null || e.memoizedState.isDehydrated && !(t.flags & 256) || (t.flags |= 1024, Wi())), W(t), null;
			case 26:
				var i = t.type, a = t.memoizedState;
				return e === null ? (Nc(t), a === null ? (W(t), Pc(t, i, null, r, n)) : (W(t), Fc(t, a))) : a ? a === e.memoizedState ? (W(t), t.flags &= -16777217) : (Nc(t), W(t), Fc(t, a)) : (e = e.memoizedProps, e !== r && Nc(t), W(t), Pc(t, i, e, r, n)), null;
			case 27:
				if (ge(t), n = de.current, i = t.type, e !== null && t.stateNode != null) e.memoizedProps !== r && Nc(t);
				else {
					if (!r) {
						if (t.stateNode === null) throw Error(o(166));
						return W(t), null;
					}
					e = le.current, Hi(t) ? Bi(t, e) : (e = ff(i, r, n), t.stateNode = e, Nc(t));
				}
				return W(t), null;
			case 5:
				if (ge(t), i = t.type, e !== null && t.stateNode != null) e.memoizedProps !== r && Nc(t);
				else {
					if (!r) {
						if (t.stateNode === null) throw Error(o(166));
						return W(t), null;
					}
					if (a = le.current, Hi(t)) Bi(t, a);
					else {
						var s = Bd(de.current);
						switch (a) {
							case 1:
								a = s.createElementNS("http://www.w3.org/2000/svg", i);
								break;
							case 2:
								a = s.createElementNS("http://www.w3.org/1998/Math/MathML", i);
								break;
							default: switch (i) {
								case "svg":
									a = s.createElementNS("http://www.w3.org/2000/svg", i);
									break;
								case "math":
									a = s.createElementNS("http://www.w3.org/1998/Math/MathML", i);
									break;
								case "script":
									a = s.createElement("div"), a.innerHTML = "<script><\/script>", a = a.removeChild(a.firstChild);
									break;
								case "select":
									a = typeof r.is == "string" ? s.createElement("select", { is: r.is }) : s.createElement("select"), r.multiple ? a.multiple = !0 : r.size && (a.size = r.size);
									break;
								default: a = typeof r.is == "string" ? s.createElement(i, { is: r.is }) : s.createElement(i);
							}
						}
						a[ct] = t, a[lt] = r;
						a: for (s = t.child; s !== null;) {
							if (s.tag === 5 || s.tag === 6) a.appendChild(s.stateNode);
							else if (s.tag !== 4 && s.tag !== 27 && s.child !== null) {
								s.child.return = s, s = s.child;
								continue;
							}
							if (s === t) break a;
							for (; s.sibling === null;) {
								if (s.return === null || s.return === t) break a;
								s = s.return;
							}
							s.sibling.return = s.return, s = s.sibling;
						}
						t.stateNode = a;
						a: switch (Pd(a, i, r), i) {
							case "button":
							case "input":
							case "select":
							case "textarea":
								r = !!r.autoFocus;
								break a;
							case "img":
								r = !0;
								break a;
							default: r = !1;
						}
						r && Nc(t);
					}
				}
				return W(t), Pc(t, t.type, e === null ? null : e.memoizedProps, t.pendingProps, n), null;
			case 6:
				if (e && t.stateNode != null) e.memoizedProps !== r && Nc(t);
				else {
					if (typeof r != "string" && t.stateNode === null) throw Error(o(166));
					if (e = de.current, Hi(t)) {
						if (e = t.stateNode, n = t.memoizedProps, r = null, i = Pi, i !== null) switch (i.tag) {
							case 27:
							case 5: r = i.memoizedProps;
						}
						e[ct] = t, e = !!(e.nodeValue === n || r !== null && !0 === r.suppressHydrationWarning || Md(e.nodeValue, n)), e || zi(t, !0);
					} else e = Bd(e).createTextNode(r), e[ct] = t, t.stateNode = e;
				}
				return W(t), null;
			case 31:
				if (n = t.memoizedState, e === null || e.memoizedState !== null) {
					if (r = Hi(t), n !== null) {
						if (e === null) {
							if (!r) throw Error(o(318));
							if (e = t.memoizedState, e = e === null ? null : e.dehydrated, !e) throw Error(o(557));
							e[ct] = t;
						} else Ui(), !(t.flags & 128) && (t.memoizedState = null), t.flags |= 4;
						W(t), e = !1;
					} else n = Wi(), e !== null && e.memoizedState !== null && (e.memoizedState.hydrationErrors = n), e = !0;
					if (!e) return t.flags & 256 ? (uo(t), t) : (uo(t), null);
					if (t.flags & 128) throw Error(o(558));
				}
				return W(t), null;
			case 13:
				if (r = t.memoizedState, e === null || e.memoizedState !== null && e.memoizedState.dehydrated !== null) {
					if (i = Hi(t), r !== null && r.dehydrated !== null) {
						if (e === null) {
							if (!i) throw Error(o(318));
							if (i = t.memoizedState, i = i === null ? null : i.dehydrated, !i) throw Error(o(317));
							i[ct] = t;
						} else Ui(), !(t.flags & 128) && (t.memoizedState = null), t.flags |= 4;
						W(t), i = !1;
					} else i = Wi(), e !== null && e.memoizedState !== null && (e.memoizedState.hydrationErrors = i), i = !0;
					if (!i) return t.flags & 256 ? (uo(t), t) : (uo(t), null);
				}
				return uo(t), t.flags & 128 ? (t.lanes = n, t) : (n = r !== null, e = e !== null && e.memoizedState !== null, n && (r = t.child, i = null, r.alternate !== null && r.alternate.memoizedState !== null && r.alternate.memoizedState.cachePool !== null && (i = r.alternate.memoizedState.cachePool.pool), a = null, r.memoizedState !== null && r.memoizedState.cachePool !== null && (a = r.memoizedState.cachePool.pool), a !== i && (r.flags |= 2048)), n !== e && n && (t.child.flags |= 8192), Ic(t, t.updateQueue), W(t), null);
			case 4: return me(), e === null && Sd(t.stateNode.containerInfo), W(t), null;
			case 10: return Xi(t.type), W(t), null;
			case 19:
				if (ce(fo), r = t.memoizedState, r === null) return W(t), null;
				if (i = !!(t.flags & 128), a = r.rendering, a === null) {
					if (i) Lc(r, !1);
					else {
						if (Wl !== 0 || e !== null && e.flags & 128) for (e = t.child; e !== null;) {
							if (a = po(e), a !== null) {
								for (t.flags |= 128, Lc(r, !1), e = a.updateQueue, t.updateQueue = e, Ic(t, e), t.subtreeFlags = 0, e = n, n = t.child; n !== null;) fi(n, e), n = n.sibling;
								return N(fo, fo.current & 1 | 2), R && ki(t, r.treeForkCount), t.child;
							}
							e = e.sibling;
						}
						r.tail !== null && De() > tu && (t.flags |= 128, i = !0, Lc(r, !1), t.lanes = 4194304);
					}
				} else {
					if (!i) {
						if (e = po(a), e !== null) {
							if (t.flags |= 128, i = !0, e = e.updateQueue, t.updateQueue = e, Ic(t, e), Lc(r, !0), r.tail === null && r.tailMode === "hidden" && !a.alternate && !R) return W(t), null;
						} else 2 * De() - r.renderingStartTime > tu && n !== 536870912 && (t.flags |= 128, i = !0, Lc(r, !1), t.lanes = 4194304);
					}
					r.isBackwards ? (a.sibling = t.child, t.child = a) : (e = r.last, e === null ? t.child = a : e.sibling = a, r.last = a);
				}
				return r.tail === null ? (W(t), null) : (e = r.tail, r.rendering = e, r.tail = e.sibling, r.renderingStartTime = De(), e.sibling = null, n = fo.current, N(fo, i ? n & 1 | 2 : n & 1), R && ki(t, r.treeForkCount), e);
			case 22:
			case 23: return uo(t), ro(), r = t.memoizedState !== null, e === null ? r && (t.flags |= 8192) : e.memoizedState !== null !== r && (t.flags |= 8192), r ? n & 536870912 && !(t.flags & 128) && (W(t), t.subtreeFlags & 6 && (t.flags |= 8192)) : W(t), n = t.updateQueue, n !== null && Ic(t, n.retryQueue), n = null, e !== null && e.memoizedState !== null && e.memoizedState.cachePool !== null && (n = e.memoizedState.cachePool.pool), r = null, t.memoizedState !== null && t.memoizedState.cachePool !== null && (r = t.memoizedState.cachePool.pool), r !== n && (t.flags |= 2048), e !== null && ce(ya), null;
			case 24: return n = null, e !== null && (n = e.memoizedState.cache), t.memoizedState.cache !== n && (t.flags |= 2048), Xi(ca), W(t), null;
			case 25: return null;
			case 30: return null;
		}
		throw Error(o(156, t.tag));
	}
	function zc(e, t) {
		switch (Mi(t), t.tag) {
			case 1: return e = t.flags, e & 65536 ? (t.flags = e & -65537 | 128, t) : null;
			case 3: return Xi(ca), me(), e = t.flags, e & 65536 && !(e & 128) ? (t.flags = e & -65537 | 128, t) : null;
			case 26:
			case 27:
			case 5: return ge(t), null;
			case 31:
				if (t.memoizedState !== null) {
					if (uo(t), t.alternate === null) throw Error(o(340));
					Ui();
				}
				return e = t.flags, e & 65536 ? (t.flags = e & -65537 | 128, t) : null;
			case 13:
				if (uo(t), e = t.memoizedState, e !== null && e.dehydrated !== null) {
					if (t.alternate === null) throw Error(o(340));
					Ui();
				}
				return e = t.flags, e & 65536 ? (t.flags = e & -65537 | 128, t) : null;
			case 19: return ce(fo), null;
			case 4: return me(), null;
			case 10: return Xi(t.type), null;
			case 22:
			case 23: return uo(t), ro(), e !== null && ce(ya), e = t.flags, e & 65536 ? (t.flags = e & -65537 | 128, t) : null;
			case 24: return Xi(ca), null;
			case 25: return null;
			default: return null;
		}
	}
	function Bc(e, t) {
		switch (Mi(t), t.tag) {
			case 3:
				Xi(ca), me();
				break;
			case 26:
			case 27:
			case 5:
				ge(t);
				break;
			case 4:
				me();
				break;
			case 31:
				t.memoizedState !== null && uo(t);
				break;
			case 13:
				uo(t);
				break;
			case 19:
				ce(fo);
				break;
			case 10:
				Xi(t.type);
				break;
			case 22:
			case 23:
				uo(t), ro(), e !== null && ce(ya);
				break;
			case 24: Xi(ca);
		}
	}
	function Vc(e, t) {
		try {
			var n = t.updateQueue, r = n === null ? null : n.lastEffect;
			if (r !== null) {
				var i = r.next;
				n = i;
				do {
					if ((n.tag & e) === e) {
						r = void 0;
						var a = n.create, o = n.inst;
						r = a(), o.destroy = r;
					}
					n = n.next;
				} while (n !== i);
			}
		} catch (e) {
			Z(t, t.return, e);
		}
	}
	function Hc(e, t, n) {
		try {
			var r = t.updateQueue, i = r === null ? null : r.lastEffect;
			if (i !== null) {
				var a = i.next;
				r = a;
				do {
					if ((r.tag & e) === e) {
						var o = r.inst, s = o.destroy;
						if (s !== void 0) {
							o.destroy = void 0, i = t;
							var c = n, l = s;
							try {
								l();
							} catch (e) {
								Z(i, c, e);
							}
						}
					}
					r = r.next;
				} while (r !== a);
			}
		} catch (e) {
			Z(t, t.return, e);
		}
	}
	function Uc(e) {
		var t = e.updateQueue;
		if (t !== null) {
			var n = e.stateNode;
			try {
				Qa(t, n);
			} catch (t) {
				Z(e, e.return, t);
			}
		}
	}
	function Wc(e, t, n) {
		n.props = Ks(e.type, e.memoizedProps), n.state = e.memoizedState;
		try {
			n.componentWillUnmount();
		} catch (n) {
			Z(e, t, n);
		}
	}
	function Gc(e, t) {
		try {
			var n = e.ref;
			if (n !== null) {
				switch (e.tag) {
					case 26:
					case 27:
					case 5:
						var r = e.stateNode;
						break;
					case 30:
						r = e.stateNode;
						break;
					default: r = e.stateNode;
				}
				typeof n == "function" ? e.refCleanup = n(r) : n.current = r;
			}
		} catch (n) {
			Z(e, t, n);
		}
	}
	function Kc(e, t) {
		var n = e.ref, r = e.refCleanup;
		if (n !== null) {
			if (typeof r == "function") try {
				r();
			} catch (n) {
				Z(e, t, n);
			} finally {
				e.refCleanup = null, e = e.alternate, e != null && (e.refCleanup = null);
			}
			else if (typeof n == "function") try {
				n(null);
			} catch (n) {
				Z(e, t, n);
			}
			else n.current = null;
		}
	}
	function qc(e) {
		var t = e.type, n = e.memoizedProps, r = e.stateNode;
		try {
			a: switch (t) {
				case "button":
				case "input":
				case "select":
				case "textarea":
					n.autoFocus && r.focus();
					break a;
				case "img": n.src ? r.src = n.src : n.srcSet && (r.srcset = n.srcSet);
			}
		} catch (t) {
			Z(e, e.return, t);
		}
	}
	function Jc(e, t, n) {
		try {
			var r = e.stateNode;
			Fd(r, e.type, n, t), r[lt] = t;
		} catch (t) {
			Z(e, e.return, t);
		}
	}
	function Yc(e) {
		return e.tag === 5 || e.tag === 3 || e.tag === 26 || e.tag === 27 && Zd(e.type) || e.tag === 4;
	}
	function Xc(e) {
		a: for (;;) {
			for (; e.sibling === null;) {
				if (e.return === null || Yc(e.return)) return null;
				e = e.return;
			}
			for (e.sibling.return = e.return, e = e.sibling; e.tag !== 5 && e.tag !== 6 && e.tag !== 18;) {
				if (e.tag === 27 && Zd(e.type) || e.flags & 2 || e.child === null || e.tag === 4) continue a;
				e.child.return = e, e = e.child;
			}
			if (!(e.flags & 2)) return e.stateNode;
		}
	}
	function Zc(e, t, n) {
		var r = e.tag;
		if (r === 5 || r === 6) e = e.stateNode, t ? (n.nodeType === 9 ? n.body : n.nodeName === "HTML" ? n.ownerDocument.body : n).insertBefore(e, t) : (t = n.nodeType === 9 ? n.body : n.nodeName === "HTML" ? n.ownerDocument.body : n, t.appendChild(e), n = n._reactRootContainer, n != null || t.onclick !== null || (t.onclick = en));
		else if (r !== 4 && (r === 27 && Zd(e.type) && (n = e.stateNode, t = null), e = e.child, e !== null)) for (Zc(e, t, n), e = e.sibling; e !== null;) Zc(e, t, n), e = e.sibling;
	}
	function Qc(e, t, n) {
		var r = e.tag;
		if (r === 5 || r === 6) e = e.stateNode, t ? n.insertBefore(e, t) : n.appendChild(e);
		else if (r !== 4 && (r === 27 && Zd(e.type) && (n = e.stateNode), e = e.child, e !== null)) for (Qc(e, t, n), e = e.sibling; e !== null;) Qc(e, t, n), e = e.sibling;
	}
	function $c(e) {
		var t = e.stateNode, n = e.memoizedProps;
		try {
			for (var r = e.type, i = t.attributes; i.length;) t.removeAttributeNode(i[0]);
			Pd(t, r, n), t[ct] = e, t[lt] = n;
		} catch (t) {
			Z(e, e.return, t);
		}
	}
	var el = !1, tl = !1, nl = !1, rl = typeof WeakSet == "function" ? WeakSet : Set, il = null;
	function al(e, t) {
		if (e = e.containerInfo, Rd = sp, e = Or(e), kr(e)) {
			if ("selectionStart" in e) var n = {
				start: e.selectionStart,
				end: e.selectionEnd
			};
			else a: {
				n = (n = e.ownerDocument) && n.defaultView || window;
				var r = n.getSelection && n.getSelection();
				if (r && r.rangeCount !== 0) {
					n = r.anchorNode;
					var i = r.anchorOffset, a = r.focusNode;
					r = r.focusOffset;
					try {
						n.nodeType, a.nodeType;
					} catch {
						n = null;
						break a;
					}
					var s = 0, c = -1, l = -1, u = 0, d = 0, f = e, p = null;
					b: for (;;) {
						for (var m; f !== n || i !== 0 && f.nodeType !== 3 || (c = s + i), f !== a || r !== 0 && f.nodeType !== 3 || (l = s + r), f.nodeType === 3 && (s += f.nodeValue.length), (m = f.firstChild) !== null;) p = f, f = m;
						for (;;) {
							if (f === e) break b;
							if (p === n && ++u === i && (c = s), p === a && ++d === r && (l = s), (m = f.nextSibling) !== null) break;
							f = p, p = f.parentNode;
						}
						f = m;
					}
					n = c === -1 || l === -1 ? null : {
						start: c,
						end: l
					};
				} else n = null;
			}
			n ||= {
				start: 0,
				end: 0
			};
		} else n = null;
		for (zd = {
			focusedElem: e,
			selectionRange: n
		}, sp = !1, il = t; il !== null;) if (t = il, e = t.child, t.subtreeFlags & 1028 && e !== null) e.return = t, il = e;
		else for (; il !== null;) {
			switch (t = il, a = t.alternate, e = t.flags, t.tag) {
				case 0:
					if (e & 4 && (e = t.updateQueue, e = e === null ? null : e.events, e !== null)) for (n = 0; n < e.length; n++) i = e[n], i.ref.impl = i.nextImpl;
					break;
				case 11:
				case 15: break;
				case 1:
					if (e & 1024 && a !== null) {
						e = void 0, n = t, i = a.memoizedProps, a = a.memoizedState, r = n.stateNode;
						try {
							var h = Ks(n.type, i);
							e = r.getSnapshotBeforeUpdate(h, a), r.__reactInternalSnapshotBeforeUpdate = e;
						} catch (e) {
							Z(n, n.return, e);
						}
					}
					break;
				case 3:
					if (e & 1024) {
						if (e = t.stateNode.containerInfo, n = e.nodeType, n === 9) ef(e);
						else if (n === 1) switch (e.nodeName) {
							case "HEAD":
							case "HTML":
							case "BODY":
								ef(e);
								break;
							default: e.textContent = "";
						}
					}
					break;
				case 5:
				case 26:
				case 27:
				case 6:
				case 4:
				case 17: break;
				default: if (e & 1024) throw Error(o(163));
			}
			if (e = t.sibling, e !== null) {
				e.return = t.return, il = e;
				break;
			}
			il = t.return;
		}
	}
	function ol(e, t, n) {
		var r = n.flags;
		switch (n.tag) {
			case 0:
			case 11:
			case 15:
				bl(e, n), r & 4 && Vc(5, n);
				break;
			case 1:
				if (bl(e, n), r & 4) {
					if (e = n.stateNode, t === null) try {
						e.componentDidMount();
					} catch (e) {
						Z(n, n.return, e);
					}
					else {
						var i = Ks(n.type, t.memoizedProps);
						t = t.memoizedState;
						try {
							e.componentDidUpdate(i, t, e.__reactInternalSnapshotBeforeUpdate);
						} catch (e) {
							Z(n, n.return, e);
						}
					}
				}
				r & 64 && Uc(n), r & 512 && Gc(n, n.return);
				break;
			case 3:
				if (bl(e, n), r & 64 && (e = n.updateQueue, e !== null)) {
					if (t = null, n.child !== null) switch (n.child.tag) {
						case 27:
						case 5:
							t = n.child.stateNode;
							break;
						case 1: t = n.child.stateNode;
					}
					try {
						Qa(e, t);
					} catch (e) {
						Z(n, n.return, e);
					}
				}
				break;
			case 27: t === null && r & 4 && $c(n);
			case 26:
			case 5:
				bl(e, n), t === null && r & 4 && qc(n), r & 512 && Gc(n, n.return);
				break;
			case 12:
				bl(e, n);
				break;
			case 31:
				bl(e, n), r & 4 && dl(e, n);
				break;
			case 13:
				bl(e, n), r & 4 && fl(e, n), r & 64 && (e = n.memoizedState, e !== null && (e = e.dehydrated, e !== null && (n = Ju.bind(null, n), sf(e, n))));
				break;
			case 22:
				if (r = n.memoizedState !== null || el, !r) {
					t = t !== null && t.memoizedState !== null || tl, i = el;
					var a = tl;
					el = r, (tl = t) && !a ? Sl(e, n, !!(n.subtreeFlags & 8772)) : bl(e, n), el = i, tl = a;
				}
				break;
			case 30: break;
			default: bl(e, n);
		}
	}
	function sl(e) {
		var t = e.alternate;
		t !== null && (e.alternate = null, sl(t)), e.child = null, e.deletions = null, e.sibling = null, e.tag === 5 && (t = e.stateNode, t !== null && gt(t)), e.stateNode = null, e.return = null, e.dependencies = null, e.memoizedProps = null, e.memoizedState = null, e.pendingProps = null, e.stateNode = null, e.updateQueue = null;
	}
	var G = null, cl = !1;
	function ll(e, t, n) {
		for (n = n.child; n !== null;) ul(e, t, n), n = n.sibling;
	}
	function ul(e, t, n) {
		if (Le && typeof Le.onCommitFiberUnmount == "function") try {
			Le.onCommitFiberUnmount(Ie, n);
		} catch {}
		switch (n.tag) {
			case 26:
				tl || Kc(n, t), ll(e, t, n), n.memoizedState ? n.memoizedState.count-- : n.stateNode && (n = n.stateNode, n.parentNode.removeChild(n));
				break;
			case 27:
				tl || Kc(n, t);
				var r = G, i = cl;
				Zd(n.type) && (G = n.stateNode, cl = !1), ll(e, t, n), pf(n.stateNode), G = r, cl = i;
				break;
			case 5: tl || Kc(n, t);
			case 6:
				if (r = G, i = cl, G = null, ll(e, t, n), G = r, cl = i, G !== null) {
					if (cl) try {
						(G.nodeType === 9 ? G.body : G.nodeName === "HTML" ? G.ownerDocument.body : G).removeChild(n.stateNode);
					} catch (e) {
						Z(n, t, e);
					}
					else try {
						G.removeChild(n.stateNode);
					} catch (e) {
						Z(n, t, e);
					}
				}
				break;
			case 18:
				G !== null && (cl ? (e = G, Qd(e.nodeType === 9 ? e.body : e.nodeName === "HTML" ? e.ownerDocument.body : e, n.stateNode), Np(e)) : Qd(G, n.stateNode));
				break;
			case 4:
				r = G, i = cl, G = n.stateNode.containerInfo, cl = !0, ll(e, t, n), G = r, cl = i;
				break;
			case 0:
			case 11:
			case 14:
			case 15:
				Hc(2, n, t), tl || Hc(4, n, t), ll(e, t, n);
				break;
			case 1:
				tl || (Kc(n, t), r = n.stateNode, typeof r.componentWillUnmount == "function" && Wc(n, t, r)), ll(e, t, n);
				break;
			case 21:
				ll(e, t, n);
				break;
			case 22:
				tl = (r = tl) || n.memoizedState !== null, ll(e, t, n), tl = r;
				break;
			default: ll(e, t, n);
		}
	}
	function dl(e, t) {
		if (t.memoizedState === null && (e = t.alternate, e !== null && (e = e.memoizedState, e !== null))) {
			e = e.dehydrated;
			try {
				Np(e);
			} catch (e) {
				Z(t, t.return, e);
			}
		}
	}
	function fl(e, t) {
		if (t.memoizedState === null && (e = t.alternate, e !== null && (e = e.memoizedState, e !== null && (e = e.dehydrated, e !== null)))) try {
			Np(e);
		} catch (e) {
			Z(t, t.return, e);
		}
	}
	function pl(e) {
		switch (e.tag) {
			case 31:
			case 13:
			case 19:
				var t = e.stateNode;
				return t === null && (t = e.stateNode = new rl()), t;
			case 22: return e = e.stateNode, t = e._retryCache, t === null && (t = e._retryCache = new rl()), t;
			default: throw Error(o(435, e.tag));
		}
	}
	function ml(e, t) {
		var n = pl(e);
		t.forEach(function(t) {
			if (!n.has(t)) {
				n.add(t);
				var r = Yu.bind(null, e, t);
				t.then(r, r);
			}
		});
	}
	function hl(e, t) {
		var n = t.deletions;
		if (n !== null) for (var r = 0; r < n.length; r++) {
			var i = n[r], a = e, s = t, c = s;
			a: for (; c !== null;) {
				switch (c.tag) {
					case 27:
						if (Zd(c.type)) {
							G = c.stateNode, cl = !1;
							break a;
						}
						break;
					case 5:
						G = c.stateNode, cl = !1;
						break a;
					case 3:
					case 4:
						G = c.stateNode.containerInfo, cl = !0;
						break a;
				}
				c = c.return;
			}
			if (G === null) throw Error(o(160));
			ul(a, s, i), G = null, cl = !1, a = i.alternate, a !== null && (a.return = null), i.return = null;
		}
		if (t.subtreeFlags & 13886) for (t = t.child; t !== null;) _l(t, e), t = t.sibling;
	}
	var gl = null;
	function _l(e, t) {
		var n = e.alternate, r = e.flags;
		switch (e.tag) {
			case 0:
			case 11:
			case 14:
			case 15:
				hl(t, e), vl(e), r & 4 && (Hc(3, e, e.return), Vc(3, e), Hc(5, e, e.return));
				break;
			case 1:
				hl(t, e), vl(e), r & 512 && (tl || n === null || Kc(n, n.return)), r & 64 && el && (e = e.updateQueue, e !== null && (r = e.callbacks, r !== null && (n = e.shared.hiddenCallbacks, e.shared.hiddenCallbacks = n === null ? r : n.concat(r))));
				break;
			case 26:
				var i = gl;
				if (hl(t, e), vl(e), r & 512 && (tl || n === null || Kc(n, n.return)), r & 4) {
					var a = n === null ? null : n.memoizedState;
					if (r = e.memoizedState, n === null) {
						if (r === null) {
							if (e.stateNode === null) {
								a: {
									r = e.type, n = e.memoizedProps, i = i.ownerDocument || i;
									b: switch (r) {
										case "title":
											a = i.getElementsByTagName("title")[0], (!a || a[ht] || a[ct] || a.namespaceURI === "http://www.w3.org/2000/svg" || a.hasAttribute("itemprop")) && (a = i.createElement(r), i.head.insertBefore(a, i.querySelector("head > title"))), Pd(a, r, n), a[ct] = e, xt(a), r = a;
											break a;
										case "link":
											var s = Vf("link", "href", i).get(r + (n.href || ""));
											if (s) {
												for (var c = 0; c < s.length; c++) if (a = s[c], a.getAttribute("href") === (n.href == null || n.href === "" ? null : n.href) && a.getAttribute("rel") === (n.rel == null ? null : n.rel) && a.getAttribute("title") === (n.title == null ? null : n.title) && a.getAttribute("crossorigin") === (n.crossOrigin == null ? null : n.crossOrigin)) {
													s.splice(c, 1);
													break b;
												}
											}
											a = i.createElement(r), Pd(a, r, n), i.head.appendChild(a);
											break;
										case "meta":
											if (s = Vf("meta", "content", i).get(r + (n.content || ""))) {
												for (c = 0; c < s.length; c++) if (a = s[c], a.getAttribute("content") === (n.content == null ? null : "" + n.content) && a.getAttribute("name") === (n.name == null ? null : n.name) && a.getAttribute("property") === (n.property == null ? null : n.property) && a.getAttribute("http-equiv") === (n.httpEquiv == null ? null : n.httpEquiv) && a.getAttribute("charset") === (n.charSet == null ? null : n.charSet)) {
													s.splice(c, 1);
													break b;
												}
											}
											a = i.createElement(r), Pd(a, r, n), i.head.appendChild(a);
											break;
										default: throw Error(o(468, r));
									}
									a[ct] = e, xt(a), r = a;
								}
								e.stateNode = r;
							} else Hf(i, e.type, e.stateNode);
						} else e.stateNode = If(i, r, e.memoizedProps);
					} else a === r ? r === null && e.stateNode !== null && Jc(e, e.memoizedProps, n.memoizedProps) : (a === null ? n.stateNode !== null && (n = n.stateNode, n.parentNode.removeChild(n)) : a.count--, r === null ? Hf(i, e.type, e.stateNode) : If(i, r, e.memoizedProps));
				}
				break;
			case 27:
				hl(t, e), vl(e), r & 512 && (tl || n === null || Kc(n, n.return)), n !== null && r & 4 && Jc(e, e.memoizedProps, n.memoizedProps);
				break;
			case 5:
				if (hl(t, e), vl(e), r & 512 && (tl || n === null || Kc(n, n.return)), e.flags & 32) {
					i = e.stateNode;
					try {
						Jt(i, "");
					} catch (t) {
						Z(e, e.return, t);
					}
				}
				r & 4 && e.stateNode != null && (i = e.memoizedProps, Jc(e, i, n === null ? i : n.memoizedProps)), r & 1024 && (nl = !0);
				break;
			case 6:
				if (hl(t, e), vl(e), r & 4) {
					if (e.stateNode === null) throw Error(o(162));
					r = e.memoizedProps, n = e.stateNode;
					try {
						n.nodeValue = r;
					} catch (t) {
						Z(e, e.return, t);
					}
				}
				break;
			case 3:
				if (Bf = null, i = gl, gl = gf(t.containerInfo), hl(t, e), gl = i, vl(e), r & 4 && n !== null && n.memoizedState.isDehydrated) try {
					Np(t.containerInfo);
				} catch (t) {
					Z(e, e.return, t);
				}
				nl && (nl = !1, yl(e));
				break;
			case 4:
				r = gl, gl = gf(e.stateNode.containerInfo), hl(t, e), vl(e), gl = r;
				break;
			case 12:
				hl(t, e), vl(e);
				break;
			case 31:
				hl(t, e), vl(e), r & 4 && (r = e.updateQueue, r !== null && (e.updateQueue = null, ml(e, r)));
				break;
			case 13:
				hl(t, e), vl(e), e.child.flags & 8192 && e.memoizedState !== null != (n !== null && n.memoizedState !== null) && ($l = De()), r & 4 && (r = e.updateQueue, r !== null && (e.updateQueue = null, ml(e, r)));
				break;
			case 22:
				i = e.memoizedState !== null;
				var l = n !== null && n.memoizedState !== null, u = el, d = tl;
				if (el = u || i, tl = d || l, hl(t, e), tl = d, el = u, vl(e), r & 8192) a: for (t = e.stateNode, t._visibility = i ? t._visibility & -2 : t._visibility | 1, i && (n === null || l || el || tl || xl(e)), n = null, t = e;;) {
					if (t.tag === 5 || t.tag === 26) {
						if (n === null) {
							l = n = t;
							try {
								if (a = l.stateNode, i) s = a.style, typeof s.setProperty == "function" ? s.setProperty("display", "none", "important") : s.display = "none";
								else {
									c = l.stateNode;
									var f = l.memoizedProps.style, p = f != null && f.hasOwnProperty("display") ? f.display : null;
									c.style.display = p == null || typeof p == "boolean" ? "" : ("" + p).trim();
								}
							} catch (e) {
								Z(l, l.return, e);
							}
						}
					} else if (t.tag === 6) {
						if (n === null) {
							l = t;
							try {
								l.stateNode.nodeValue = i ? "" : l.memoizedProps;
							} catch (e) {
								Z(l, l.return, e);
							}
						}
					} else if (t.tag === 18) {
						if (n === null) {
							l = t;
							try {
								var m = l.stateNode;
								i ? $d(m, !0) : $d(l.stateNode, !1);
							} catch (e) {
								Z(l, l.return, e);
							}
						}
					} else if ((t.tag !== 22 && t.tag !== 23 || t.memoizedState === null || t === e) && t.child !== null) {
						t.child.return = t, t = t.child;
						continue;
					}
					if (t === e) break a;
					for (; t.sibling === null;) {
						if (t.return === null || t.return === e) break a;
						n === t && (n = null), t = t.return;
					}
					n === t && (n = null), t.sibling.return = t.return, t = t.sibling;
				}
				r & 4 && (r = e.updateQueue, r !== null && (n = r.retryQueue, n !== null && (r.retryQueue = null, ml(e, n))));
				break;
			case 19:
				hl(t, e), vl(e), r & 4 && (r = e.updateQueue, r !== null && (e.updateQueue = null, ml(e, r)));
				break;
			case 30: break;
			case 21: break;
			default: hl(t, e), vl(e);
		}
	}
	function vl(e) {
		var t = e.flags;
		if (t & 2) {
			try {
				for (var n, r = e.return; r !== null;) {
					if (Yc(r)) {
						n = r;
						break;
					}
					r = r.return;
				}
				if (n == null) throw Error(o(160));
				switch (n.tag) {
					case 27:
						var i = n.stateNode;
						Qc(e, Xc(e), i);
						break;
					case 5:
						var a = n.stateNode;
						n.flags & 32 && (Jt(a, ""), n.flags &= -33), Qc(e, Xc(e), a);
						break;
					case 3:
					case 4:
						var s = n.stateNode.containerInfo;
						Zc(e, Xc(e), s);
						break;
					default: throw Error(o(161));
				}
			} catch (t) {
				Z(e, e.return, t);
			}
			e.flags &= -3;
		}
		t & 4096 && (e.flags &= -4097);
	}
	function yl(e) {
		if (e.subtreeFlags & 1024) for (e = e.child; e !== null;) {
			var t = e;
			yl(t), t.tag === 5 && t.flags & 1024 && t.stateNode.reset(), e = e.sibling;
		}
	}
	function bl(e, t) {
		if (t.subtreeFlags & 8772) for (t = t.child; t !== null;) ol(e, t.alternate, t), t = t.sibling;
	}
	function xl(e) {
		for (e = e.child; e !== null;) {
			var t = e;
			switch (t.tag) {
				case 0:
				case 11:
				case 14:
				case 15:
					Hc(4, t, t.return), xl(t);
					break;
				case 1:
					Kc(t, t.return);
					var n = t.stateNode;
					typeof n.componentWillUnmount == "function" && Wc(t, t.return, n), xl(t);
					break;
				case 27: pf(t.stateNode);
				case 26:
				case 5:
					Kc(t, t.return), xl(t);
					break;
				case 22:
					t.memoizedState === null && xl(t);
					break;
				case 30:
					xl(t);
					break;
				default: xl(t);
			}
			e = e.sibling;
		}
	}
	function Sl(e, t, n) {
		for (n &&= !!(t.subtreeFlags & 8772), t = t.child; t !== null;) {
			var r = t.alternate, i = e, a = t, o = a.flags;
			switch (a.tag) {
				case 0:
				case 11:
				case 15:
					Sl(i, a, n), Vc(4, a);
					break;
				case 1:
					if (Sl(i, a, n), r = a, i = r.stateNode, typeof i.componentDidMount == "function") try {
						i.componentDidMount();
					} catch (e) {
						Z(r, r.return, e);
					}
					if (r = a, i = r.updateQueue, i !== null) {
						var s = r.stateNode;
						try {
							var c = i.shared.hiddenCallbacks;
							if (c !== null) for (i.shared.hiddenCallbacks = null, i = 0; i < c.length; i++) Za(c[i], s);
						} catch (e) {
							Z(r, r.return, e);
						}
					}
					n && o & 64 && Uc(a), Gc(a, a.return);
					break;
				case 27: $c(a);
				case 26:
				case 5:
					Sl(i, a, n), n && r === null && o & 4 && qc(a), Gc(a, a.return);
					break;
				case 12:
					Sl(i, a, n);
					break;
				case 31:
					Sl(i, a, n), n && o & 4 && dl(i, a);
					break;
				case 13:
					Sl(i, a, n), n && o & 4 && fl(i, a);
					break;
				case 22:
					a.memoizedState === null && Sl(i, a, n), Gc(a, a.return);
					break;
				case 30: break;
				default: Sl(i, a, n);
			}
			t = t.sibling;
		}
	}
	function Cl(e, t) {
		var n = null;
		e !== null && e.memoizedState !== null && e.memoizedState.cachePool !== null && (n = e.memoizedState.cachePool.pool), e = null, t.memoizedState !== null && t.memoizedState.cachePool !== null && (e = t.memoizedState.cachePool.pool), e !== n && (e != null && e.refCount++, n != null && ua(n));
	}
	function wl(e, t) {
		e = null, t.alternate !== null && (e = t.alternate.memoizedState.cache), t = t.memoizedState.cache, t !== e && (t.refCount++, e != null && ua(e));
	}
	function Tl(e, t, n, r) {
		if (t.subtreeFlags & 10256) for (t = t.child; t !== null;) El(e, t, n, r), t = t.sibling;
	}
	function El(e, t, n, r) {
		var i = t.flags;
		switch (t.tag) {
			case 0:
			case 11:
			case 15:
				Tl(e, t, n, r), i & 2048 && Vc(9, t);
				break;
			case 1:
				Tl(e, t, n, r);
				break;
			case 3:
				Tl(e, t, n, r), i & 2048 && (e = null, t.alternate !== null && (e = t.alternate.memoizedState.cache), t = t.memoizedState.cache, t !== e && (t.refCount++, e != null && ua(e)));
				break;
			case 12:
				if (i & 2048) {
					Tl(e, t, n, r), e = t.stateNode;
					try {
						var a = t.memoizedProps, o = a.id, s = a.onPostCommit;
						typeof s == "function" && s(o, t.alternate === null ? "mount" : "update", e.passiveEffectDuration, -0);
					} catch (e) {
						Z(t, t.return, e);
					}
				} else Tl(e, t, n, r);
				break;
			case 31:
				Tl(e, t, n, r);
				break;
			case 13:
				Tl(e, t, n, r);
				break;
			case 23: break;
			case 22:
				a = t.stateNode, o = t.alternate, t.memoizedState === null ? a._visibility & 2 ? Tl(e, t, n, r) : (a._visibility |= 2, Dl(e, t, n, r, !!(t.subtreeFlags & 10256) || !1)) : a._visibility & 2 ? Tl(e, t, n, r) : Ol(e, t), i & 2048 && Cl(o, t);
				break;
			case 24:
				Tl(e, t, n, r), i & 2048 && wl(t.alternate, t);
				break;
			default: Tl(e, t, n, r);
		}
	}
	function Dl(e, t, n, r, i) {
		for (i &&= !!(t.subtreeFlags & 10256) || !1, t = t.child; t !== null;) {
			var a = e, o = t, s = n, c = r, l = o.flags;
			switch (o.tag) {
				case 0:
				case 11:
				case 15:
					Dl(a, o, s, c, i), Vc(8, o);
					break;
				case 23: break;
				case 22:
					var u = o.stateNode;
					o.memoizedState === null ? (u._visibility |= 2, Dl(a, o, s, c, i)) : u._visibility & 2 ? Dl(a, o, s, c, i) : Ol(a, o), i && l & 2048 && Cl(o.alternate, o);
					break;
				case 24:
					Dl(a, o, s, c, i), i && l & 2048 && wl(o.alternate, o);
					break;
				default: Dl(a, o, s, c, i);
			}
			t = t.sibling;
		}
	}
	function Ol(e, t) {
		if (t.subtreeFlags & 10256) for (t = t.child; t !== null;) {
			var n = e, r = t, i = r.flags;
			switch (r.tag) {
				case 22:
					Ol(n, r), i & 2048 && Cl(r.alternate, r);
					break;
				case 24:
					Ol(n, r), i & 2048 && wl(r.alternate, r);
					break;
				default: Ol(n, r);
			}
			t = t.sibling;
		}
	}
	var kl = 8192;
	function Al(e, t, n) {
		if (e.subtreeFlags & kl) for (e = e.child; e !== null;) jl(e, t, n), e = e.sibling;
	}
	function jl(e, t, n) {
		switch (e.tag) {
			case 26:
				Al(e, t, n), e.flags & kl && e.memoizedState !== null && Gf(n, gl, e.memoizedState, e.memoizedProps);
				break;
			case 5:
				Al(e, t, n);
				break;
			case 3:
			case 4:
				var r = gl;
				gl = gf(e.stateNode.containerInfo), Al(e, t, n), gl = r;
				break;
			case 22:
				e.memoizedState === null && (r = e.alternate, r !== null && r.memoizedState !== null ? (r = kl, kl = 16777216, Al(e, t, n), kl = r) : Al(e, t, n));
				break;
			default: Al(e, t, n);
		}
	}
	function Ml(e) {
		var t = e.alternate;
		if (t !== null && (e = t.child, e !== null)) {
			t.child = null;
			do
				t = e.sibling, e.sibling = null, e = t;
			while (e !== null);
		}
	}
	function Nl(e) {
		var t = e.deletions;
		if (e.flags & 16) {
			if (t !== null) for (var n = 0; n < t.length; n++) {
				var r = t[n];
				il = r, Il(r, e);
			}
			Ml(e);
		}
		if (e.subtreeFlags & 10256) for (e = e.child; e !== null;) Pl(e), e = e.sibling;
	}
	function Pl(e) {
		switch (e.tag) {
			case 0:
			case 11:
			case 15:
				Nl(e), e.flags & 2048 && Hc(9, e, e.return);
				break;
			case 3:
				Nl(e);
				break;
			case 12:
				Nl(e);
				break;
			case 22:
				var t = e.stateNode;
				e.memoizedState !== null && t._visibility & 2 && (e.return === null || e.return.tag !== 13) ? (t._visibility &= -3, Fl(e)) : Nl(e);
				break;
			default: Nl(e);
		}
	}
	function Fl(e) {
		var t = e.deletions;
		if (e.flags & 16) {
			if (t !== null) for (var n = 0; n < t.length; n++) {
				var r = t[n];
				il = r, Il(r, e);
			}
			Ml(e);
		}
		for (e = e.child; e !== null;) {
			switch (t = e, t.tag) {
				case 0:
				case 11:
				case 15:
					Hc(8, t, t.return), Fl(t);
					break;
				case 22:
					n = t.stateNode, n._visibility & 2 && (n._visibility &= -3, Fl(t));
					break;
				default: Fl(t);
			}
			e = e.sibling;
		}
	}
	function Il(e, t) {
		for (; il !== null;) {
			var n = il;
			switch (n.tag) {
				case 0:
				case 11:
				case 15:
					Hc(8, n, t);
					break;
				case 23:
				case 22:
					if (n.memoizedState !== null && n.memoizedState.cachePool !== null) {
						var r = n.memoizedState.cachePool.pool;
						r != null && r.refCount++;
					}
					break;
				case 24: ua(n.memoizedState.cache);
			}
			if (r = n.child, r !== null) r.return = n, il = r;
			else a: for (n = e; il !== null;) {
				r = il;
				var i = r.sibling, a = r.return;
				if (sl(r), r === n) {
					il = null;
					break a;
				}
				if (i !== null) {
					i.return = a, il = i;
					break a;
				}
				il = a;
			}
		}
	}
	var Ll = {
		getCacheForType: function(e) {
			var t = na(ca), n = t.data.get(e);
			return n === void 0 && (n = e(), t.data.set(e, n)), n;
		},
		cacheSignal: function() {
			return na(ca).controller.signal;
		}
	}, Rl = typeof WeakMap == "function" ? WeakMap : Map, K = 0, q = null, J = null, Y = 0, X = 0, zl = null, Bl = !1, Vl = !1, Hl = !1, Ul = 0, Wl = 0, Gl = 0, Kl = 0, ql = 0, Jl = 0, Yl = 0, Xl = null, Zl = null, Ql = !1, $l = 0, eu = 0, tu = Infinity, nu = null, ru = null, iu = 0, au = null, ou = null, su = 0, cu = 0, lu = null, uu = null, du = 0, fu = null;
	function pu() {
		return K & 2 && Y !== 0 ? Y & -Y : A.T === null ? at() : dd();
	}
	function mu() {
		if (Jl === 0) {
			if (!(Y & 536870912) || R) {
				var e = We;
				We <<= 1, !(We & 3932160) && (We = 262144), Jl = e;
			} else Jl = 536870912;
		}
		return e = io.current, e !== null && (e.flags |= 32), Jl;
	}
	function hu(e, t, n) {
		(e === q && (X === 2 || X === 9) || e.cancelPendingCommit !== null) && (Su(e, 0), yu(e, Y, Jl, !1)), Qe(e, n), K & 2 && e === q || (e === q && (!(K & 2) && (Kl |= n), Wl === 4 && yu(e, Y, Jl, !1)), rd(e));
	}
	function gu(e, t, n) {
		if (K & 6) throw Error(o(327));
		var r = !n && !(t & 127) && (t & e.expiredLanes) === 0 || Je(e, t), i = r ? Au(e, t) : Ou(e, t, !0), a = r;
		do {
			if (i === 0) {
				Vl && !r && yu(e, t, 0, !1);
				break;
			}
			if (n = e.current.alternate, a && !vu(n)) i = Ou(e, t, !1), a = !1;
			else {
				if (i === 2) {
					if (a = t, e.errorRecoveryDisabledLanes & a) var s = 0;
					else s = e.pendingLanes & -536870913, s = s === 0 ? s & 536870912 ? 536870912 : 0 : s;
					if (s !== 0) {
						t = s;
						a: {
							var c = e;
							i = Xl;
							var l = c.current.memoizedState.isDehydrated;
							if (l && (Su(c, s).flags |= 256), s = Ou(c, s, !1), s !== 2) {
								if (Hl && !l) {
									c.errorRecoveryDisabledLanes |= a, Kl |= a, i = 4;
									break a;
								}
								a = Zl, Zl = i, a !== null && (Zl === null ? Zl = a : Zl.push.apply(Zl, a));
							}
							i = s;
						}
						if (a = !1, i !== 2) continue;
					}
				}
				if (i === 1) {
					Su(e, 0), yu(e, t, 0, !0);
					break;
				}
				a: {
					switch (r = e, a = i, a) {
						case 0:
						case 1: throw Error(o(345));
						case 4: if ((t & 4194048) !== t) break;
						case 6:
							yu(r, t, Jl, !Bl);
							break a;
						case 2:
							Zl = null;
							break;
						case 3:
						case 5: break;
						default: throw Error(o(329));
					}
					if ((t & 62914560) === t && (i = $l + 300 - De(), 10 < i)) {
						if (yu(r, t, Jl, !Bl), qe(r, 0, !0) !== 0) break a;
						su = t, r.timeoutHandle = Kd(_u.bind(null, r, n, Zl, nu, Ql, t, Jl, Kl, Yl, Bl, a, "Throttled", -0, 0), i);
						break a;
					}
					_u(r, n, Zl, nu, Ql, t, Jl, Kl, Yl, Bl, a, null, -0, 0);
				}
				break;
			}
		} while (1);
		rd(e);
	}
	function _u(e, t, n, r, i, a, o, s, c, l, u, d, f, p) {
		if (e.timeoutHandle = -1, d = t.subtreeFlags, d & 8192 || (d & 16785408) == 16785408) {
			d = {
				stylesheets: null,
				count: 0,
				imgCount: 0,
				imgBytes: 0,
				suspenseyImages: [],
				waitingForImages: !0,
				waitingForViewTransition: !1,
				unsuspend: en
			}, jl(t, a, d);
			var m = (a & 62914560) === a ? $l - De() : (a & 4194048) === a ? eu - De() : 0;
			if (m = qf(d, m), m !== null) {
				su = a, e.cancelPendingCommit = m(Lu.bind(null, e, t, a, n, r, i, o, s, c, u, d, null, f, p)), yu(e, a, o, !l);
				return;
			}
		}
		Lu(e, t, a, n, r, i, o, s, c);
	}
	function vu(e) {
		for (var t = e;;) {
			var n = t.tag;
			if ((n === 0 || n === 11 || n === 15) && t.flags & 16384 && (n = t.updateQueue, n !== null && (n = n.stores, n !== null))) for (var r = 0; r < n.length; r++) {
				var i = n[r], a = i.getSnapshot;
				i = i.value;
				try {
					if (!Cr(a(), i)) return !1;
				} catch {
					return !1;
				}
			}
			if (n = t.child, t.subtreeFlags & 16384 && n !== null) n.return = t, t = n;
			else {
				if (t === e) break;
				for (; t.sibling === null;) {
					if (t.return === null || t.return === e) return !0;
					t = t.return;
				}
				t.sibling.return = t.return, t = t.sibling;
			}
		}
		return !0;
	}
	function yu(e, t, n, r) {
		t &= ~ql, t &= ~Kl, e.suspendedLanes |= t, e.pingedLanes &= ~t, r && (e.warmLanes |= t), r = e.expirationTimes;
		for (var i = t; 0 < i;) {
			var a = 31 - ze(i), o = 1 << a;
			r[a] = -1, i &= ~o;
		}
		n !== 0 && et(e, n, t);
	}
	function bu() {
		return K & 6 ? !0 : (id(0, !1), !1);
	}
	function xu() {
		if (J !== null) {
			if (X === 0) var e = J.return;
			else e = J, Ji = qi = null, ko(e), Na = null, Pa = 0, e = J;
			for (; e !== null;) Bc(e.alternate, e), e = e.return;
			J = null;
		}
	}
	function Su(e, t) {
		var n = e.timeoutHandle;
		n !== -1 && (e.timeoutHandle = -1, qd(n)), n = e.cancelPendingCommit, n !== null && (e.cancelPendingCommit = null, n()), su = 0, xu(), q = e, J = n = di(e.current, null), Y = t, X = 0, zl = null, Bl = !1, Vl = Je(e, t), Hl = !1, Yl = Jl = ql = Kl = Gl = Wl = 0, Zl = Xl = null, Ql = !1, t & 8 && (t |= t & 32);
		var r = e.entangledLanes;
		if (r !== 0) for (e = e.entanglements, r &= t; 0 < r;) {
			var i = 31 - ze(r), a = 1 << i;
			t |= e[i], r &= ~a;
		}
		return Ul = t, ti(), n;
	}
	function Cu(e, t) {
		B = null, A.H = U, t === Ca || t === Ta ? (t = ja(), X = 3) : t === wa ? (t = ja(), X = 4) : X = t === nc ? 8 : typeof t == "object" && t && typeof t.then == "function" ? 6 : 1, zl = t, J === null && (Wl = 1, Xs(e, yi(t, e.current)));
	}
	function wu() {
		var e = io.current;
		return e === null ? !0 : (Y & 4194048) === Y ? ao === null : (Y & 62914560) === Y || Y & 536870912 ? e === ao : !1;
	}
	function Tu() {
		var e = A.H;
		return A.H = U, e === null ? U : e;
	}
	function Eu() {
		var e = A.A;
		return A.A = Ll, e;
	}
	function Du() {
		Wl = 4, Bl || (Y & 4194048) !== Y && io.current !== null || (Vl = !0), !(Gl & 134217727) && !(Kl & 134217727) || q === null || yu(q, Y, Jl, !1);
	}
	function Ou(e, t, n) {
		var r = K;
		K |= 2;
		var i = Tu(), a = Eu();
		(q !== e || Y !== t) && (nu = null, Su(e, t)), t = !1;
		var o = Wl;
		a: do
			try {
				if (X !== 0 && J !== null) {
					var s = J, c = zl;
					switch (X) {
						case 8:
							xu(), o = 6;
							break a;
						case 3:
						case 2:
						case 9:
						case 6:
							io.current === null && (t = !0);
							var l = X;
							if (X = 0, zl = null, Pu(e, s, c, l), n && Vl) {
								o = 0;
								break a;
							}
							break;
						default: l = X, X = 0, zl = null, Pu(e, s, c, l);
					}
				}
				ku(), o = Wl;
				break;
			} catch (t) {
				Cu(e, t);
			}
		while (1);
		return t && e.shellSuspendCounter++, Ji = qi = null, K = r, A.H = i, A.A = a, J === null && (q = null, Y = 0, ti()), o;
	}
	function ku() {
		for (; J !== null;) Mu(J);
	}
	function Au(e, t) {
		var n = K;
		K |= 2;
		var r = Tu(), i = Eu();
		q !== e || Y !== t ? (nu = null, tu = De() + 500, Su(e, t)) : Vl = Je(e, t);
		a: do
			try {
				if (X !== 0 && J !== null) {
					t = J;
					var a = zl;
					b: switch (X) {
						case 1:
							X = 0, zl = null, Pu(e, t, a, 1);
							break;
						case 2:
						case 9:
							if (Da(a)) {
								X = 0, zl = null, Nu(t);
								break;
							}
							t = function() {
								X !== 2 && X !== 9 || q !== e || (X = 7), rd(e);
							}, a.then(t, t);
							break a;
						case 3:
							X = 7;
							break a;
						case 4:
							X = 5;
							break a;
						case 7:
							Da(a) ? (X = 0, zl = null, Nu(t)) : (X = 0, zl = null, Pu(e, t, a, 7));
							break;
						case 5:
							var s = null;
							switch (J.tag) {
								case 26: s = J.memoizedState;
								case 5:
								case 27:
									var c = J;
									if (s ? Wf(s) : c.stateNode.complete) {
										X = 0, zl = null;
										var l = c.sibling;
										if (l !== null) J = l;
										else {
											var u = c.return;
											u === null ? J = null : (J = u, Fu(u));
										}
										break b;
									}
							}
							X = 0, zl = null, Pu(e, t, a, 5);
							break;
						case 6:
							X = 0, zl = null, Pu(e, t, a, 6);
							break;
						case 8:
							xu(), Wl = 6;
							break a;
						default: throw Error(o(462));
					}
				}
				ju();
				break;
			} catch (t) {
				Cu(e, t);
			}
		while (1);
		return Ji = qi = null, A.H = r, A.A = i, K = n, J === null ? (q = null, Y = 0, ti(), Wl) : 0;
	}
	function ju() {
		for (; J !== null && !Te();) Mu(J);
	}
	function Mu(e) {
		var t = Mc(e.alternate, e, Ul);
		e.memoizedProps = e.pendingProps, t === null ? Fu(e) : J = t;
	}
	function Nu(e) {
		var t = e, n = t.alternate;
		switch (t.tag) {
			case 15:
			case 0:
				t = gc(n, t, t.pendingProps, t.type, void 0, Y);
				break;
			case 11:
				t = gc(n, t, t.pendingProps, t.type.render, t.ref, Y);
				break;
			case 5: ko(t);
			default: Bc(n, t), t = J = fi(t, Ul), t = Mc(n, t, Ul);
		}
		e.memoizedProps = e.pendingProps, t === null ? Fu(e) : J = t;
	}
	function Pu(e, t, n, r) {
		Ji = qi = null, ko(t), Na = null, Pa = 0;
		var i = t.return;
		try {
			if (tc(e, i, t, n, Y)) {
				Wl = 1, Xs(e, yi(n, e.current)), J = null;
				return;
			}
		} catch (t) {
			if (i !== null) throw J = i, t;
			Wl = 1, Xs(e, yi(n, e.current)), J = null;
			return;
		}
		t.flags & 32768 ? (R || r === 1 ? e = !0 : Vl || Y & 536870912 ? e = !1 : (Bl = e = !0, (r === 2 || r === 9 || r === 3 || r === 6) && (r = io.current, r !== null && r.tag === 13 && (r.flags |= 16384))), Iu(t, e)) : Fu(t);
	}
	function Fu(e) {
		var t = e;
		do {
			if (t.flags & 32768) {
				Iu(t, Bl);
				return;
			}
			e = t.return;
			var n = Rc(t.alternate, t, Ul);
			if (n !== null) {
				J = n;
				return;
			}
			if (t = t.sibling, t !== null) {
				J = t;
				return;
			}
			J = t = e;
		} while (t !== null);
		Wl === 0 && (Wl = 5);
	}
	function Iu(e, t) {
		do {
			var n = zc(e.alternate, e);
			if (n !== null) {
				n.flags &= 32767, J = n;
				return;
			}
			if (n = e.return, n !== null && (n.flags |= 32768, n.subtreeFlags = 0, n.deletions = null), !t && (e = e.sibling, e !== null)) {
				J = e;
				return;
			}
			J = e = n;
		} while (e !== null);
		Wl = 6, J = null;
	}
	function Lu(e, t, n, r, i, a, s, c, l) {
		e.cancelPendingCommit = null;
		do
			Hu();
		while (iu !== 0);
		if (K & 6) throw Error(o(327));
		if (t !== null) {
			if (t === e.current) throw Error(o(177));
			if (a = t.lanes | t.childLanes, a |= ei, $e(e, n, a, s, c, l), e === q && (J = q = null, Y = 0), ou = t, au = e, su = n, cu = a, lu = i, uu = r, t.subtreeFlags & 10256 || t.flags & 10256 ? (e.callbackNode = null, e.callbackPriority = 0, Xu(je, function() {
				return Uu(), null;
			})) : (e.callbackNode = null, e.callbackPriority = 0), r = !!(t.flags & 13878), t.subtreeFlags & 13878 || r) {
				r = A.T, A.T = null, i = j.p, j.p = 2, s = K, K |= 4;
				try {
					al(e, t, n);
				} finally {
					K = s, j.p = i, A.T = r;
				}
			}
			iu = 1, Ru(), zu(), Bu();
		}
	}
	function Ru() {
		if (iu === 1) {
			iu = 0;
			var e = au, t = ou, n = !!(t.flags & 13878);
			if (t.subtreeFlags & 13878 || n) {
				n = A.T, A.T = null;
				var r = j.p;
				j.p = 2;
				var i = K;
				K |= 4;
				try {
					_l(t, e);
					var a = zd, o = Or(e.containerInfo), s = a.focusedElem, c = a.selectionRange;
					if (o !== s && s && s.ownerDocument && Dr(s.ownerDocument.documentElement, s)) {
						if (c !== null && kr(s)) {
							var l = c.start, u = c.end;
							if (u === void 0 && (u = l), "selectionStart" in s) s.selectionStart = l, s.selectionEnd = Math.min(u, s.value.length);
							else {
								var d = s.ownerDocument || document, f = d && d.defaultView || window;
								if (f.getSelection) {
									var p = f.getSelection(), m = s.textContent.length, h = Math.min(c.start, m), g = c.end === void 0 ? h : Math.min(c.end, m);
									!p.extend && h > g && (o = g, g = h, h = o);
									var _ = Er(s, h), v = Er(s, g);
									if (_ && v && (p.rangeCount !== 1 || p.anchorNode !== _.node || p.anchorOffset !== _.offset || p.focusNode !== v.node || p.focusOffset !== v.offset)) {
										var y = d.createRange();
										y.setStart(_.node, _.offset), p.removeAllRanges(), h > g ? (p.addRange(y), p.extend(v.node, v.offset)) : (y.setEnd(v.node, v.offset), p.addRange(y));
									}
								}
							}
						}
						for (d = [], p = s; p = p.parentNode;) p.nodeType === 1 && d.push({
							element: p,
							left: p.scrollLeft,
							top: p.scrollTop
						});
						for (typeof s.focus == "function" && s.focus(), s = 0; s < d.length; s++) {
							var b = d[s];
							b.element.scrollLeft = b.left, b.element.scrollTop = b.top;
						}
					}
					sp = !!Rd, zd = Rd = null;
				} finally {
					K = i, j.p = r, A.T = n;
				}
			}
			e.current = t, iu = 2;
		}
	}
	function zu() {
		if (iu === 2) {
			iu = 0;
			var e = au, t = ou, n = !!(t.flags & 8772);
			if (t.subtreeFlags & 8772 || n) {
				n = A.T, A.T = null;
				var r = j.p;
				j.p = 2;
				var i = K;
				K |= 4;
				try {
					ol(e, t.alternate, t);
				} finally {
					K = i, j.p = r, A.T = n;
				}
			}
			iu = 3;
		}
	}
	function Bu() {
		if (iu === 4 || iu === 3) {
			iu = 0, Ee();
			var e = au, t = ou, n = su, r = uu;
			t.subtreeFlags & 10256 || t.flags & 10256 ? iu = 5 : (iu = 0, ou = au = null, Vu(e, e.pendingLanes));
			var i = e.pendingLanes;
			if (i === 0 && (ru = null), it(n), t = t.stateNode, Le && typeof Le.onCommitFiberRoot == "function") try {
				Le.onCommitFiberRoot(Ie, t, void 0, (t.current.flags & 128) == 128);
			} catch {}
			if (r !== null) {
				t = A.T, i = j.p, j.p = 2, A.T = null;
				try {
					for (var a = e.onRecoverableError, o = 0; o < r.length; o++) {
						var s = r[o];
						a(s.value, { componentStack: s.stack });
					}
				} finally {
					A.T = t, j.p = i;
				}
			}
			su & 3 && Hu(), rd(e), i = e.pendingLanes, n & 261930 && i & 42 ? e === fu ? du++ : (du = 0, fu = e) : du = 0, id(0, !1);
		}
	}
	function Vu(e, t) {
		(e.pooledCacheLanes &= t) === 0 && (t = e.pooledCache, t != null && (e.pooledCache = null, ua(t)));
	}
	function Hu() {
		return Ru(), zu(), Bu(), Uu();
	}
	function Uu() {
		if (iu !== 5) return !1;
		var e = au, t = cu;
		cu = 0;
		var n = it(su), r = A.T, i = j.p;
		try {
			j.p = 32 > n ? 32 : n, A.T = null, n = lu, lu = null;
			var a = au, s = su;
			if (iu = 0, ou = au = null, su = 0, K & 6) throw Error(o(331));
			var c = K;
			if (K |= 4, Pl(a.current), El(a, a.current, s, n), K = c, id(0, !1), Le && typeof Le.onPostCommitFiberRoot == "function") try {
				Le.onPostCommitFiberRoot(Ie, a);
			} catch {}
			return !0;
		} finally {
			j.p = i, A.T = r, Vu(e, t);
		}
	}
	function Wu(e, t, n) {
		t = yi(n, t), t = Qs(e.stateNode, t, 2), e = Ga(e, t, 2), e !== null && (Qe(e, 2), rd(e));
	}
	function Z(e, t, n) {
		if (e.tag === 3) Wu(e, e, n);
		else for (; t !== null;) {
			if (t.tag === 3) {
				Wu(t, e, n);
				break;
			}
			if (t.tag === 1) {
				var r = t.stateNode;
				if (typeof t.type.getDerivedStateFromError == "function" || typeof r.componentDidCatch == "function" && (ru === null || !ru.has(r))) {
					e = yi(n, e), n = $s(2), r = Ga(t, n, 2), r !== null && (ec(n, r, t, e), Qe(r, 2), rd(r));
					break;
				}
			}
			t = t.return;
		}
	}
	function Gu(e, t, n) {
		var r = e.pingCache;
		if (r === null) {
			r = e.pingCache = new Rl();
			var i = /* @__PURE__ */ new Set();
			r.set(t, i);
		} else i = r.get(t), i === void 0 && (i = /* @__PURE__ */ new Set(), r.set(t, i));
		i.has(n) || (Hl = !0, i.add(n), e = Ku.bind(null, e, t, n), t.then(e, e));
	}
	function Ku(e, t, n) {
		var r = e.pingCache;
		r !== null && r.delete(t), e.pingedLanes |= e.suspendedLanes & n, e.warmLanes &= ~n, q === e && (Y & n) === n && (Wl === 4 || Wl === 3 && (Y & 62914560) === Y && 300 > De() - $l ? !(K & 2) && Su(e, 0) : ql |= n, Yl === Y && (Yl = 0)), rd(e);
	}
	function qu(e, t) {
		t === 0 && (t = Xe()), e = ii(e, t), e !== null && (Qe(e, t), rd(e));
	}
	function Ju(e) {
		var t = e.memoizedState, n = 0;
		t !== null && (n = t.retryLane), qu(e, n);
	}
	function Yu(e, t) {
		var n = 0;
		switch (e.tag) {
			case 31:
			case 13:
				var r = e.stateNode, i = e.memoizedState;
				i !== null && (n = i.retryLane);
				break;
			case 19:
				r = e.stateNode;
				break;
			case 22:
				r = e.stateNode._retryCache;
				break;
			default: throw Error(o(314));
		}
		r !== null && r.delete(t), qu(e, n);
	}
	function Xu(e, t) {
		return we(e, t);
	}
	var Zu = null, Qu = null, $u = !1, ed = !1, td = !1, nd = 0;
	function rd(e) {
		e !== Qu && e.next === null && (Qu === null ? Zu = Qu = e : Qu = Qu.next = e), ed = !0, $u || ($u = !0, ud());
	}
	function id(e, t) {
		if (!td && ed) {
			td = !0;
			do
				for (var n = !1, r = Zu; r !== null;) {
					if (!t) {
						if (e !== 0) {
							var i = r.pendingLanes;
							if (i === 0) var a = 0;
							else {
								var o = r.suspendedLanes, s = r.pingedLanes;
								a = (1 << 31 - ze(42 | e) + 1) - 1, a &= i & ~(o & ~s), a = a & 201326741 ? a & 201326741 | 1 : a ? a | 2 : 0;
							}
							a !== 0 && (n = !0, ld(r, a));
						} else a = Y, a = qe(r, r === q ? a : 0, r.cancelPendingCommit !== null || r.timeoutHandle !== -1), !(a & 3) || Je(r, a) || (n = !0, ld(r, a));
					}
					r = r.next;
				}
			while (n);
			td = !1;
		}
	}
	function ad() {
		od();
	}
	function od() {
		ed = $u = !1;
		var e = 0;
		nd !== 0 && Gd() && (e = nd);
		for (var t = De(), n = null, r = Zu; r !== null;) {
			var i = r.next, a = sd(r, t);
			a === 0 ? (r.next = null, n === null ? Zu = i : n.next = i, i === null && (Qu = n)) : (n = r, (e !== 0 || a & 3) && (ed = !0)), r = i;
		}
		iu !== 0 && iu !== 5 || id(e, !1), nd !== 0 && (nd = 0);
	}
	function sd(e, t) {
		for (var n = e.suspendedLanes, r = e.pingedLanes, i = e.expirationTimes, a = e.pendingLanes & -62914561; 0 < a;) {
			var o = 31 - ze(a), s = 1 << o, c = i[o];
			c === -1 ? ((s & n) === 0 || (s & r) !== 0) && (i[o] = Ye(s, t)) : c <= t && (e.expiredLanes |= s), a &= ~s;
		}
		if (t = q, n = Y, n = qe(e, e === t ? n : 0, e.cancelPendingCommit !== null || e.timeoutHandle !== -1), r = e.callbackNode, n === 0 || e === t && (X === 2 || X === 9) || e.cancelPendingCommit !== null) return r !== null && r !== null && F(r), e.callbackNode = null, e.callbackPriority = 0;
		if (!(n & 3) || Je(e, n)) {
			if (t = n & -n, t === e.callbackPriority) return t;
			switch (r !== null && F(r), it(n)) {
				case 2:
				case 8:
					n = Ae;
					break;
				case 32:
					n = je;
					break;
				case 268435456:
					n = Ne;
					break;
				default: n = je;
			}
			return r = cd.bind(null, e), n = we(n, r), e.callbackPriority = t, e.callbackNode = n, t;
		}
		return r !== null && r !== null && F(r), e.callbackPriority = 2, e.callbackNode = null, 2;
	}
	function cd(e, t) {
		if (iu !== 0 && iu !== 5) return e.callbackNode = null, e.callbackPriority = 0, null;
		var n = e.callbackNode;
		if (Hu() && e.callbackNode !== n) return null;
		var r = Y;
		return r = qe(e, e === q ? r : 0, e.cancelPendingCommit !== null || e.timeoutHandle !== -1), r === 0 ? null : (gu(e, r, t), sd(e, De()), e.callbackNode != null && e.callbackNode === n ? cd.bind(null, e) : null);
	}
	function ld(e, t) {
		if (Hu()) return null;
		gu(e, t, !0);
	}
	function ud() {
		Yd(function() {
			K & 6 ? we(ke, ad) : od();
		});
	}
	function dd() {
		if (nd === 0) {
			var e = pa;
			e === 0 && (e = Ue, Ue <<= 1, !(Ue & 261888) && (Ue = 256)), nd = e;
		}
		return nd;
	}
	function fd(e) {
		return e == null || typeof e == "symbol" || typeof e == "boolean" ? null : typeof e == "function" ? e : $t("" + e);
	}
	function pd(e, t) {
		var n = t.ownerDocument.createElement("input");
		return n.name = t.name, n.value = t.value, e.id && n.setAttribute("form", e.id), t.parentNode.insertBefore(n, t), e = new FormData(e), n.parentNode.removeChild(n), e;
	}
	function md(e, t, n, r, i) {
		if (t === "submit" && n && n.stateNode === i) {
			var a = fd((i[lt] || null).action), o = r.submitter;
			o && (t = (t = o[lt] || null) ? fd(t.formAction) : o.getAttribute("formAction"), t !== null && (a = t, o = null));
			var s = new Sn("action", "action", null, r, i);
			e.push({
				event: s,
				listeners: [{
					instance: null,
					listener: function() {
						if (r.defaultPrevented) {
							if (nd !== 0) {
								var e = o ? pd(i, o) : new FormData(i);
								Ts(n, {
									pending: !0,
									data: e,
									method: i.method,
									action: a
								}, null, e);
							}
						} else typeof a == "function" && (s.preventDefault(), e = o ? pd(i, o) : new FormData(i), Ts(n, {
							pending: !0,
							data: e,
							method: i.method,
							action: a
						}, a, e));
					},
					currentTarget: i
				}]
			});
		}
	}
	for (var hd = 0; hd < Yr.length; hd++) {
		var gd = Yr[hd];
		Xr(gd.toLowerCase(), "on" + (gd[0].toUpperCase() + gd.slice(1)));
	}
	Xr(Vr, "onAnimationEnd"), Xr(Hr, "onAnimationIteration"), Xr(Ur, "onAnimationStart"), Xr("dblclick", "onDoubleClick"), Xr("focusin", "onFocus"), Xr("focusout", "onBlur"), Xr(Wr, "onTransitionRun"), Xr(Gr, "onTransitionStart"), Xr(Kr, "onTransitionCancel"), Xr(qr, "onTransitionEnd"), Tt("onMouseEnter", ["mouseout", "mouseover"]), Tt("onMouseLeave", ["mouseout", "mouseover"]), Tt("onPointerEnter", ["pointerout", "pointerover"]), Tt("onPointerLeave", ["pointerout", "pointerover"]), wt("onChange", "change click focusin focusout input keydown keyup selectionchange".split(" ")), wt("onSelect", "focusout contextmenu dragend focusin keydown keyup mousedown mouseup selectionchange".split(" ")), wt("onBeforeInput", [
		"compositionend",
		"keypress",
		"textInput",
		"paste"
	]), wt("onCompositionEnd", "compositionend focusout keydown keypress keyup mousedown".split(" ")), wt("onCompositionStart", "compositionstart focusout keydown keypress keyup mousedown".split(" ")), wt("onCompositionUpdate", "compositionupdate focusout keydown keypress keyup mousedown".split(" "));
	var _d = "abort canplay canplaythrough durationchange emptied encrypted ended error loadeddata loadedmetadata loadstart pause play playing progress ratechange resize seeked seeking stalled suspend timeupdate volumechange waiting".split(" "), vd = new Set("beforetoggle cancel close invalid load scroll scrollend toggle".split(" ").concat(_d));
	function yd(e, t) {
		t = !!(t & 4);
		for (var n = 0; n < e.length; n++) {
			var r = e[n], i = r.event;
			r = r.listeners;
			a: {
				var a = void 0;
				if (t) for (var o = r.length - 1; 0 <= o; o--) {
					var s = r[o], c = s.instance, l = s.currentTarget;
					if (s = s.listener, c !== a && i.isPropagationStopped()) break a;
					a = s, i.currentTarget = l;
					try {
						a(i);
					} catch (e) {
						Zr(e);
					}
					i.currentTarget = null, a = c;
				}
				else for (o = 0; o < r.length; o++) {
					if (s = r[o], c = s.instance, l = s.currentTarget, s = s.listener, c !== a && i.isPropagationStopped()) break a;
					a = s, i.currentTarget = l;
					try {
						a(i);
					} catch (e) {
						Zr(e);
					}
					i.currentTarget = null, a = c;
				}
			}
		}
	}
	function Q(e, t) {
		var n = t[dt];
		n === void 0 && (n = t[dt] = /* @__PURE__ */ new Set());
		var r = e + "__bubble";
		n.has(r) || (Cd(t, e, 2, !1), n.add(r));
	}
	function bd(e, t, n) {
		var r = 0;
		t && (r |= 4), Cd(n, e, r, t);
	}
	var xd = "_reactListening" + Math.random().toString(36).slice(2);
	function Sd(e) {
		if (!e[xd]) {
			e[xd] = !0, St.forEach(function(t) {
				t !== "selectionchange" && (vd.has(t) || bd(t, !1, e), bd(t, !0, e));
			});
			var t = e.nodeType === 9 ? e : e.ownerDocument;
			t === null || t[xd] || (t[xd] = !0, bd("selectionchange", !1, t));
		}
	}
	function Cd(e, t, n, r) {
		switch (mp(t)) {
			case 2:
				var i = cp;
				break;
			case 8:
				i = lp;
				break;
			default: i = up;
		}
		n = i.bind(null, t, n, e), i = void 0, !dn || t !== "touchstart" && t !== "touchmove" && t !== "wheel" || (i = !0), r ? i === void 0 ? e.addEventListener(t, n, !0) : e.addEventListener(t, n, {
			capture: !0,
			passive: i
		}) : i === void 0 ? e.addEventListener(t, n, !1) : e.addEventListener(t, n, { passive: i });
	}
	function wd(e, t, n, r, i) {
		var a = r;
		if (!(t & 1) && !(t & 2) && r !== null) a: for (;;) {
			if (r === null) return;
			var o = r.tag;
			if (o === 3 || o === 4) {
				var s = r.stateNode.containerInfo;
				if (s === i) break;
				if (o === 4) for (o = r.return; o !== null;) {
					var l = o.tag;
					if ((l === 3 || l === 4) && o.stateNode.containerInfo === i) return;
					o = o.return;
				}
				for (; s !== null;) {
					if (o = _t(s), o === null) return;
					if (l = o.tag, l === 5 || l === 6 || l === 26 || l === 27) {
						r = a = o;
						continue a;
					}
					s = s.parentNode;
				}
			}
			r = r.return;
		}
		cn(function() {
			var r = a, i = nn(n), o = [];
			a: {
				var s = Jr.get(e);
				if (s !== void 0) {
					var l = Sn, u = e;
					switch (e) {
						case "keypress": if (_n(n) === 0) break a;
						case "keydown":
						case "keyup":
							l = Bn;
							break;
						case "focusin":
							u = "focus", l = jn;
							break;
						case "focusout":
							u = "blur", l = jn;
							break;
						case "beforeblur":
						case "afterblur":
							l = jn;
							break;
						case "click": if (n.button === 2) break a;
						case "auxclick":
						case "dblclick":
						case "mousedown":
						case "mousemove":
						case "mouseup":
						case "mouseout":
						case "mouseover":
						case "contextmenu":
							l = kn;
							break;
						case "drag":
						case "dragend":
						case "dragenter":
						case "dragexit":
						case "dragleave":
						case "dragover":
						case "dragstart":
						case "drop":
							l = An;
							break;
						case "touchcancel":
						case "touchend":
						case "touchmove":
						case "touchstart":
							l = Hn;
							break;
						case Vr:
						case Hr:
						case Ur:
							l = Mn;
							break;
						case qr:
							l = Un;
							break;
						case "scroll":
						case "scrollend":
							l = wn;
							break;
						case "wheel":
							l = Wn;
							break;
						case "copy":
						case "cut":
						case "paste":
							l = Nn;
							break;
						case "gotpointercapture":
						case "lostpointercapture":
						case "pointercancel":
						case "pointerdown":
						case "pointermove":
						case "pointerout":
						case "pointerover":
						case "pointerup":
							l = Vn;
							break;
						case "toggle":
						case "beforetoggle": l = Gn;
					}
					var d = !!(t & 4), f = !d && (e === "scroll" || e === "scrollend"), p = d ? s === null ? null : s + "Capture" : s;
					d = [];
					for (var m = r, h; m !== null;) {
						var g = m;
						if (h = g.stateNode, g = g.tag, g !== 5 && g !== 26 && g !== 27 || h === null || p === null || (g = ln(m, p), g != null && d.push(Td(m, g, h))), f) break;
						m = m.return;
					}
					0 < d.length && (s = new l(s, u, null, n, i), o.push({
						event: s,
						listeners: d
					}));
				}
			}
			if (!(t & 7)) {
				a: {
					if (s = e === "mouseover" || e === "pointerover", l = e === "mouseout" || e === "pointerout", s && n !== tn && (u = n.relatedTarget || n.fromElement) && (_t(u) || u[ut])) break a;
					if ((l || s) && (s = i.window === i ? i : (s = i.ownerDocument) ? s.defaultView || s.parentWindow : window, l ? (u = n.relatedTarget || n.toElement, l = r, u = u ? _t(u) : null, u !== null && (f = c(u), d = u.tag, u !== f || d !== 5 && d !== 27 && d !== 6) && (u = null)) : (l = null, u = r), l !== u)) {
						if (d = kn, g = "onMouseLeave", p = "onMouseEnter", m = "mouse", (e === "pointerout" || e === "pointerover") && (d = Vn, g = "onPointerLeave", p = "onPointerEnter", m = "pointer"), f = l == null ? s : yt(l), h = u == null ? s : yt(u), s = new d(g, m + "leave", l, n, i), s.target = f, s.relatedTarget = h, g = null, _t(i) === r && (d = new d(p, m + "enter", u, n, i), d.target = h, d.relatedTarget = f, g = d), f = g, l && u) b: {
							for (d = Dd, p = l, m = u, h = 0, g = p; g; g = d(g)) h++;
							g = 0;
							for (var _ = m; _; _ = d(_)) g++;
							for (; 0 < h - g;) p = d(p), h--;
							for (; 0 < g - h;) m = d(m), g--;
							for (; h--;) {
								if (p === m || m !== null && p === m.alternate) {
									d = p;
									break b;
								}
								p = d(p), m = d(m);
							}
							d = null;
						}
						else d = null;
						l !== null && Od(o, s, l, d, !1), u !== null && f !== null && Od(o, f, u, d, !0);
					}
				}
				a: {
					if (s = r ? yt(r) : window, l = s.nodeName && s.nodeName.toLowerCase(), l === "select" || l === "input" && s.type === "file") var v = dr;
					else if (ar(s)) {
						if (fr) v = xr;
						else {
							v = yr;
							var y = vr;
						}
					} else l = s.nodeName, !l || l.toLowerCase() !== "input" || s.type !== "checkbox" && s.type !== "radio" ? r && Xt(r.elementType) && (v = dr) : v = br;
					if (v &&= v(e, r)) {
						or(o, v, n, i);
						break a;
					}
					y && y(e, s, r), e === "focusout" && r && s.type === "number" && r.memoizedProps.value != null && Ut(s, "number", s.value);
				}
				switch (y = r ? yt(r) : window, e) {
					case "focusin":
						(ar(y) || y.contentEditable === "true") && (jr = y, Mr = r, Nr = null);
						break;
					case "focusout":
						Nr = Mr = jr = null;
						break;
					case "mousedown":
						Pr = !0;
						break;
					case "contextmenu":
					case "mouseup":
					case "dragend":
						Pr = !1, Fr(o, n, i);
						break;
					case "selectionchange": if (Ar) break;
					case "keydown":
					case "keyup": Fr(o, n, i);
				}
				var b;
				if (qn) b: {
					switch (e) {
						case "compositionstart":
							var x = "onCompositionStart";
							break b;
						case "compositionend":
							x = "onCompositionEnd";
							break b;
						case "compositionupdate":
							x = "onCompositionUpdate";
							break b;
					}
					x = void 0;
				}
				else tr ? $n(e, n) && (x = "onCompositionEnd") : e === "keydown" && n.keyCode === 229 && (x = "onCompositionStart");
				x && (Xn && n.locale !== "ko" && (tr || x !== "onCompositionStart" ? x === "onCompositionEnd" && tr && (b = gn()) : (pn = i, mn = "value" in pn ? pn.value : pn.textContent, tr = !0)), y = Ed(r, x), 0 < y.length && (x = new Pn(x, e, null, n, i), o.push({
					event: x,
					listeners: y
				}), b ? x.data = b : (b = er(n), b !== null && (x.data = b)))), (b = Yn ? nr(e, n) : rr(e, n)) && (x = Ed(r, "onBeforeInput"), 0 < x.length && (y = new Pn("onBeforeInput", "beforeinput", null, n, i), o.push({
					event: y,
					listeners: x
				}), y.data = b)), md(o, e, r, n, i);
			}
			yd(o, t);
		});
	}
	function Td(e, t, n) {
		return {
			instance: e,
			listener: t,
			currentTarget: n
		};
	}
	function Ed(e, t) {
		for (var n = t + "Capture", r = []; e !== null;) {
			var i = e, a = i.stateNode;
			if (i = i.tag, i !== 5 && i !== 26 && i !== 27 || a === null || (i = ln(e, n), i != null && r.unshift(Td(e, i, a)), i = ln(e, t), i != null && r.push(Td(e, i, a))), e.tag === 3) return r;
			e = e.return;
		}
		return [];
	}
	function Dd(e) {
		if (e === null) return null;
		do
			e = e.return;
		while (e && e.tag !== 5 && e.tag !== 27);
		return e || null;
	}
	function Od(e, t, n, r, i) {
		for (var a = t._reactName, o = []; n !== null && n !== r;) {
			var s = n, c = s.alternate, l = s.stateNode;
			if (s = s.tag, c !== null && c === r) break;
			s !== 5 && s !== 26 && s !== 27 || l === null || (c = l, i ? (l = ln(n, a), l != null && o.unshift(Td(n, l, c))) : i || (l = ln(n, a), l != null && o.push(Td(n, l, c)))), n = n.return;
		}
		o.length !== 0 && e.push({
			event: t,
			listeners: o
		});
	}
	var kd = /\r\n?/g, Ad = /\u0000|\uFFFD/g;
	function jd(e) {
		return (typeof e == "string" ? e : "" + e).replace(kd, "\n").replace(Ad, "");
	}
	function Md(e, t) {
		return t = jd(t), jd(e) === t;
	}
	function $(e, t, n, r, i, a) {
		switch (n) {
			case "children":
				typeof r == "string" ? t === "body" || t === "textarea" && r === "" || Jt(e, r) : (typeof r == "number" || typeof r == "bigint") && t !== "body" && Jt(e, "" + r);
				break;
			case "className":
				jt(e, "class", r);
				break;
			case "tabIndex":
				jt(e, "tabindex", r);
				break;
			case "dir":
			case "role":
			case "viewBox":
			case "width":
			case "height":
				jt(e, n, r);
				break;
			case "style":
				L(e, r, a);
				break;
			case "data": if (t !== "object") {
				jt(e, "data", r);
				break;
			}
			case "src":
			case "href":
				if (r === "" && (t !== "a" || n !== "href") || r == null || typeof r == "function" || typeof r == "symbol" || typeof r == "boolean") {
					e.removeAttribute(n);
					break;
				}
				r = $t("" + r), e.setAttribute(n, r);
				break;
			case "action":
			case "formAction":
				if (typeof r == "function") {
					e.setAttribute(n, "javascript:throw new Error('A React form was unexpectedly submitted. If you called form.submit() manually, consider using form.requestSubmit() instead. If you\\'re trying to use event.stopPropagation() in a submit event handler, consider also calling event.preventDefault().')");
					break;
				}
				if (typeof a == "function" && (n === "formAction" ? (t !== "input" && $(e, t, "name", i.name, i, null), $(e, t, "formEncType", i.formEncType, i, null), $(e, t, "formMethod", i.formMethod, i, null), $(e, t, "formTarget", i.formTarget, i, null)) : ($(e, t, "encType", i.encType, i, null), $(e, t, "method", i.method, i, null), $(e, t, "target", i.target, i, null))), r == null || typeof r == "symbol" || typeof r == "boolean") {
					e.removeAttribute(n);
					break;
				}
				r = $t("" + r), e.setAttribute(n, r);
				break;
			case "onClick":
				r != null && (e.onclick = en);
				break;
			case "onScroll":
				r != null && Q("scroll", e);
				break;
			case "onScrollEnd":
				r != null && Q("scrollend", e);
				break;
			case "dangerouslySetInnerHTML":
				if (r != null) {
					if (typeof r != "object" || !("__html" in r)) throw Error(o(61));
					if (n = r.__html, n != null) {
						if (i.children != null) throw Error(o(60));
						e.innerHTML = n;
					}
				}
				break;
			case "multiple":
				e.multiple = r && typeof r != "function" && typeof r != "symbol";
				break;
			case "muted":
				e.muted = r && typeof r != "function" && typeof r != "symbol";
				break;
			case "suppressContentEditableWarning":
			case "suppressHydrationWarning":
			case "defaultValue":
			case "defaultChecked":
			case "innerHTML":
			case "ref": break;
			case "autoFocus": break;
			case "xlinkHref":
				if (r == null || typeof r == "function" || typeof r == "boolean" || typeof r == "symbol") {
					e.removeAttribute("xlink:href");
					break;
				}
				n = $t("" + r), e.setAttributeNS("http://www.w3.org/1999/xlink", "xlink:href", n);
				break;
			case "contentEditable":
			case "spellCheck":
			case "draggable":
			case "value":
			case "autoReverse":
			case "externalResourcesRequired":
			case "focusable":
			case "preserveAlpha":
				r != null && typeof r != "function" && typeof r != "symbol" ? e.setAttribute(n, "" + r) : e.removeAttribute(n);
				break;
			case "inert":
			case "allowFullScreen":
			case "async":
			case "autoPlay":
			case "controls":
			case "default":
			case "defer":
			case "disabled":
			case "disablePictureInPicture":
			case "disableRemotePlayback":
			case "formNoValidate":
			case "hidden":
			case "loop":
			case "noModule":
			case "noValidate":
			case "open":
			case "playsInline":
			case "readOnly":
			case "required":
			case "reversed":
			case "scoped":
			case "seamless":
			case "itemScope":
				r && typeof r != "function" && typeof r != "symbol" ? e.setAttribute(n, "") : e.removeAttribute(n);
				break;
			case "capture":
			case "download":
				!0 === r ? e.setAttribute(n, "") : !1 !== r && r != null && typeof r != "function" && typeof r != "symbol" ? e.setAttribute(n, r) : e.removeAttribute(n);
				break;
			case "cols":
			case "rows":
			case "size":
			case "span":
				r != null && typeof r != "function" && typeof r != "symbol" && !isNaN(r) && 1 <= r ? e.setAttribute(n, r) : e.removeAttribute(n);
				break;
			case "rowSpan":
			case "start":
				r == null || typeof r == "function" || typeof r == "symbol" || isNaN(r) ? e.removeAttribute(n) : e.setAttribute(n, r);
				break;
			case "popover":
				Q("beforetoggle", e), Q("toggle", e), At(e, "popover", r);
				break;
			case "xlinkActuate":
				Mt(e, "http://www.w3.org/1999/xlink", "xlink:actuate", r);
				break;
			case "xlinkArcrole":
				Mt(e, "http://www.w3.org/1999/xlink", "xlink:arcrole", r);
				break;
			case "xlinkRole":
				Mt(e, "http://www.w3.org/1999/xlink", "xlink:role", r);
				break;
			case "xlinkShow":
				Mt(e, "http://www.w3.org/1999/xlink", "xlink:show", r);
				break;
			case "xlinkTitle":
				Mt(e, "http://www.w3.org/1999/xlink", "xlink:title", r);
				break;
			case "xlinkType":
				Mt(e, "http://www.w3.org/1999/xlink", "xlink:type", r);
				break;
			case "xmlBase":
				Mt(e, "http://www.w3.org/XML/1998/namespace", "xml:base", r);
				break;
			case "xmlLang":
				Mt(e, "http://www.w3.org/XML/1998/namespace", "xml:lang", r);
				break;
			case "xmlSpace":
				Mt(e, "http://www.w3.org/XML/1998/namespace", "xml:space", r);
				break;
			case "is":
				At(e, "is", r);
				break;
			case "innerText":
			case "textContent": break;
			default: (!(2 < n.length) || n[0] !== "o" && n[0] !== "O" || n[1] !== "n" && n[1] !== "N") && (n = Zt.get(n) || n, At(e, n, r));
		}
	}
	function Nd(e, t, n, r, i, a) {
		switch (n) {
			case "style":
				L(e, r, a);
				break;
			case "dangerouslySetInnerHTML":
				if (r != null) {
					if (typeof r != "object" || !("__html" in r)) throw Error(o(61));
					if (n = r.__html, n != null) {
						if (i.children != null) throw Error(o(60));
						e.innerHTML = n;
					}
				}
				break;
			case "children":
				typeof r == "string" ? Jt(e, r) : (typeof r == "number" || typeof r == "bigint") && Jt(e, "" + r);
				break;
			case "onScroll":
				r != null && Q("scroll", e);
				break;
			case "onScrollEnd":
				r != null && Q("scrollend", e);
				break;
			case "onClick":
				r != null && (e.onclick = en);
				break;
			case "suppressContentEditableWarning":
			case "suppressHydrationWarning":
			case "innerHTML":
			case "ref": break;
			case "innerText":
			case "textContent": break;
			default: if (!Ct.hasOwnProperty(n)) a: {
				if (n[0] === "o" && n[1] === "n" && (i = n.endsWith("Capture"), t = n.slice(2, i ? n.length - 7 : void 0), a = e[lt] || null, a = a == null ? null : a[n], typeof a == "function" && e.removeEventListener(t, a, i), typeof r == "function")) {
					typeof a != "function" && a !== null && (n in e ? e[n] = null : e.hasAttribute(n) && e.removeAttribute(n)), e.addEventListener(t, r, i);
					break a;
				}
				n in e ? e[n] = r : !0 === r ? e.setAttribute(n, "") : At(e, n, r);
			}
		}
	}
	function Pd(e, t, n) {
		switch (t) {
			case "div":
			case "span":
			case "svg":
			case "path":
			case "a":
			case "g":
			case "p":
			case "li": break;
			case "img":
				Q("error", e), Q("load", e);
				var r = !1, i = !1, a;
				for (a in n) if (n.hasOwnProperty(a)) {
					var s = n[a];
					if (s != null) switch (a) {
						case "src":
							r = !0;
							break;
						case "srcSet":
							i = !0;
							break;
						case "children":
						case "dangerouslySetInnerHTML": throw Error(o(137, t));
						default: $(e, t, a, s, n, null);
					}
				}
				i && $(e, t, "srcSet", n.srcSet, n, null), r && $(e, t, "src", n.src, n, null);
				return;
			case "input":
				Q("invalid", e);
				var c = a = s = i = null, l = null, u = null;
				for (r in n) if (n.hasOwnProperty(r)) {
					var d = n[r];
					if (d != null) switch (r) {
						case "name":
							i = d;
							break;
						case "type":
							s = d;
							break;
						case "checked":
							l = d;
							break;
						case "defaultChecked":
							u = d;
							break;
						case "value":
							a = d;
							break;
						case "defaultValue":
							c = d;
							break;
						case "children":
						case "dangerouslySetInnerHTML":
							if (d != null) throw Error(o(137, t));
							break;
						default: $(e, t, r, d, n, null);
					}
				}
				Ht(e, a, c, l, u, s, i, !1);
				return;
			case "select":
				for (i in Q("invalid", e), r = s = a = null, n) if (n.hasOwnProperty(i) && (c = n[i], c != null)) switch (i) {
					case "value":
						a = c;
						break;
					case "defaultValue":
						s = c;
						break;
					case "multiple": r = c;
					default: $(e, t, i, c, n, null);
				}
				t = a, n = s, e.multiple = !!r, t == null ? n != null && Wt(e, !!r, n, !0) : Wt(e, !!r, t, !1);
				return;
			case "textarea":
				for (s in Q("invalid", e), a = i = r = null, n) if (n.hasOwnProperty(s) && (c = n[s], c != null)) switch (s) {
					case "value":
						r = c;
						break;
					case "defaultValue":
						i = c;
						break;
					case "children":
						a = c;
						break;
					case "dangerouslySetInnerHTML":
						if (c != null) throw Error(o(91));
						break;
					default: $(e, t, s, c, n, null);
				}
				qt(e, r, i, a);
				return;
			case "option":
				for (l in n) if (n.hasOwnProperty(l) && (r = n[l], r != null)) switch (l) {
					case "selected":
						e.selected = r && typeof r != "function" && typeof r != "symbol";
						break;
					default: $(e, t, l, r, n, null);
				}
				return;
			case "dialog":
				Q("beforetoggle", e), Q("toggle", e), Q("cancel", e), Q("close", e);
				break;
			case "iframe":
			case "object":
				Q("load", e);
				break;
			case "video":
			case "audio":
				for (r = 0; r < _d.length; r++) Q(_d[r], e);
				break;
			case "image":
				Q("error", e), Q("load", e);
				break;
			case "details":
				Q("toggle", e);
				break;
			case "embed":
			case "source":
			case "link": Q("error", e), Q("load", e);
			case "area":
			case "base":
			case "br":
			case "col":
			case "hr":
			case "keygen":
			case "meta":
			case "param":
			case "track":
			case "wbr":
			case "menuitem":
				for (u in n) if (n.hasOwnProperty(u) && (r = n[u], r != null)) switch (u) {
					case "children":
					case "dangerouslySetInnerHTML": throw Error(o(137, t));
					default: $(e, t, u, r, n, null);
				}
				return;
			default: if (Xt(t)) {
				for (d in n) n.hasOwnProperty(d) && (r = n[d], r !== void 0 && Nd(e, t, d, r, n, void 0));
				return;
			}
		}
		for (c in n) n.hasOwnProperty(c) && (r = n[c], r != null && $(e, t, c, r, n, null));
	}
	function Fd(e, t, n, r) {
		switch (t) {
			case "div":
			case "span":
			case "svg":
			case "path":
			case "a":
			case "g":
			case "p":
			case "li": break;
			case "input":
				var i = null, a = null, s = null, c = null, l = null, u = null, d = null;
				for (m in n) {
					var f = n[m];
					if (n.hasOwnProperty(m) && f != null) switch (m) {
						case "checked": break;
						case "value": break;
						case "defaultValue": l = f;
						default: r.hasOwnProperty(m) || $(e, t, m, null, r, f);
					}
				}
				for (var p in r) {
					var m = r[p];
					if (f = n[p], r.hasOwnProperty(p) && (m != null || f != null)) switch (p) {
						case "type":
							a = m;
							break;
						case "name":
							i = m;
							break;
						case "checked":
							u = m;
							break;
						case "defaultChecked":
							d = m;
							break;
						case "value":
							s = m;
							break;
						case "defaultValue":
							c = m;
							break;
						case "children":
						case "dangerouslySetInnerHTML":
							if (m != null) throw Error(o(137, t));
							break;
						default: m !== f && $(e, t, p, m, r, f);
					}
				}
				Vt(e, s, c, l, u, d, a, i);
				return;
			case "select":
				for (a in m = s = c = p = null, n) if (l = n[a], n.hasOwnProperty(a) && l != null) switch (a) {
					case "value": break;
					case "multiple": m = l;
					default: r.hasOwnProperty(a) || $(e, t, a, null, r, l);
				}
				for (i in r) if (a = r[i], l = n[i], r.hasOwnProperty(i) && (a != null || l != null)) switch (i) {
					case "value":
						p = a;
						break;
					case "defaultValue":
						c = a;
						break;
					case "multiple": s = a;
					default: a !== l && $(e, t, i, a, r, l);
				}
				t = c, n = s, r = m, p == null ? !!r != !!n && (t == null ? Wt(e, !!n, n ? [] : "", !1) : Wt(e, !!n, t, !0)) : Wt(e, !!n, p, !1);
				return;
			case "textarea":
				for (c in m = p = null, n) if (i = n[c], n.hasOwnProperty(c) && i != null && !r.hasOwnProperty(c)) switch (c) {
					case "value": break;
					case "children": break;
					default: $(e, t, c, null, r, i);
				}
				for (s in r) if (i = r[s], a = n[s], r.hasOwnProperty(s) && (i != null || a != null)) switch (s) {
					case "value":
						p = i;
						break;
					case "defaultValue":
						m = i;
						break;
					case "children": break;
					case "dangerouslySetInnerHTML":
						if (i != null) throw Error(o(91));
						break;
					default: i !== a && $(e, t, s, i, r, a);
				}
				Gt(e, p, m);
				return;
			case "option":
				for (var h in n) if (p = n[h], n.hasOwnProperty(h) && p != null && !r.hasOwnProperty(h)) switch (h) {
					case "selected":
						e.selected = !1;
						break;
					default: $(e, t, h, null, r, p);
				}
				for (l in r) if (p = r[l], m = n[l], r.hasOwnProperty(l) && p !== m && (p != null || m != null)) switch (l) {
					case "selected":
						e.selected = p && typeof p != "function" && typeof p != "symbol";
						break;
					default: $(e, t, l, p, r, m);
				}
				return;
			case "img":
			case "link":
			case "area":
			case "base":
			case "br":
			case "col":
			case "embed":
			case "hr":
			case "keygen":
			case "meta":
			case "param":
			case "source":
			case "track":
			case "wbr":
			case "menuitem":
				for (var g in n) p = n[g], n.hasOwnProperty(g) && p != null && !r.hasOwnProperty(g) && $(e, t, g, null, r, p);
				for (u in r) if (p = r[u], m = n[u], r.hasOwnProperty(u) && p !== m && (p != null || m != null)) switch (u) {
					case "children":
					case "dangerouslySetInnerHTML":
						if (p != null) throw Error(o(137, t));
						break;
					default: $(e, t, u, p, r, m);
				}
				return;
			default: if (Xt(t)) {
				for (var _ in n) p = n[_], n.hasOwnProperty(_) && p !== void 0 && !r.hasOwnProperty(_) && Nd(e, t, _, void 0, r, p);
				for (d in r) p = r[d], m = n[d], !r.hasOwnProperty(d) || p === m || p === void 0 && m === void 0 || Nd(e, t, d, p, r, m);
				return;
			}
		}
		for (var v in n) p = n[v], n.hasOwnProperty(v) && p != null && !r.hasOwnProperty(v) && $(e, t, v, null, r, p);
		for (f in r) p = r[f], m = n[f], !r.hasOwnProperty(f) || p === m || p == null && m == null || $(e, t, f, p, r, m);
	}
	function Id(e) {
		switch (e) {
			case "css":
			case "script":
			case "font":
			case "img":
			case "image":
			case "input":
			case "link": return !0;
			default: return !1;
		}
	}
	function Ld() {
		if (typeof performance.getEntriesByType == "function") {
			for (var e = 0, t = 0, n = performance.getEntriesByType("resource"), r = 0; r < n.length; r++) {
				var i = n[r], a = i.transferSize, o = i.initiatorType, s = i.duration;
				if (a && s && Id(o)) {
					for (o = 0, s = i.responseEnd, r += 1; r < n.length; r++) {
						var c = n[r], l = c.startTime;
						if (l > s) break;
						var u = c.transferSize, d = c.initiatorType;
						u && Id(d) && (c = c.responseEnd, o += u * (c < s ? 1 : (s - l) / (c - l)));
					}
					if (--r, t += 8 * (a + o) / (i.duration / 1e3), e++, 10 < e) break;
				}
			}
			if (0 < e) return t / e / 1e6;
		}
		return navigator.connection && (e = navigator.connection.downlink, typeof e == "number") ? e : 5;
	}
	var Rd = null, zd = null;
	function Bd(e) {
		return e.nodeType === 9 ? e : e.ownerDocument;
	}
	function Vd(e) {
		switch (e) {
			case "http://www.w3.org/2000/svg": return 1;
			case "http://www.w3.org/1998/Math/MathML": return 2;
			default: return 0;
		}
	}
	function Hd(e, t) {
		if (e === 0) switch (t) {
			case "svg": return 1;
			case "math": return 2;
			default: return 0;
		}
		return e === 1 && t === "foreignObject" ? 0 : e;
	}
	function Ud(e, t) {
		return e === "textarea" || e === "noscript" || typeof t.children == "string" || typeof t.children == "number" || typeof t.children == "bigint" || typeof t.dangerouslySetInnerHTML == "object" && t.dangerouslySetInnerHTML !== null && t.dangerouslySetInnerHTML.__html != null;
	}
	var Wd = null;
	function Gd() {
		var e = window.event;
		return e && e.type === "popstate" ? e !== Wd && (Wd = e, !0) : (Wd = null, !1);
	}
	var Kd = typeof setTimeout == "function" ? setTimeout : void 0, qd = typeof clearTimeout == "function" ? clearTimeout : void 0, Jd = typeof Promise == "function" ? Promise : void 0, Yd = typeof queueMicrotask == "function" ? queueMicrotask : Jd === void 0 ? Kd : function(e) {
		return Jd.resolve(null).then(e).catch(Xd);
	};
	function Xd(e) {
		setTimeout(function() {
			throw e;
		});
	}
	function Zd(e) {
		return e === "head";
	}
	function Qd(e, t) {
		var n = t, r = 0;
		do {
			var i = n.nextSibling;
			if (e.removeChild(n), i && i.nodeType === 8) {
				if (n = i.data, n === "/$" || n === "/&") {
					if (r === 0) {
						e.removeChild(i), Np(t);
						return;
					}
					r--;
				} else if (n === "$" || n === "$?" || n === "$~" || n === "$!" || n === "&") r++;
				else if (n === "html") pf(e.ownerDocument.documentElement);
				else if (n === "head") {
					n = e.ownerDocument.head, pf(n);
					for (var a = n.firstChild; a;) {
						var o = a.nextSibling, s = a.nodeName;
						a[ht] || s === "SCRIPT" || s === "STYLE" || s === "LINK" && a.rel.toLowerCase() === "stylesheet" || n.removeChild(a), a = o;
					}
				} else n === "body" && pf(e.ownerDocument.body);
			}
			n = i;
		} while (n);
		Np(t);
	}
	function $d(e, t) {
		var n = e;
		e = 0;
		do {
			var r = n.nextSibling;
			if (n.nodeType === 1 ? t ? (n._stashedDisplay = n.style.display, n.style.display = "none") : (n.style.display = n._stashedDisplay || "", n.getAttribute("style") === "" && n.removeAttribute("style")) : n.nodeType === 3 && (t ? (n._stashedText = n.nodeValue, n.nodeValue = "") : n.nodeValue = n._stashedText || ""), r && r.nodeType === 8) {
				if (n = r.data, n === "/$") {
					if (e === 0) break;
					e--;
				} else n !== "$" && n !== "$?" && n !== "$~" && n !== "$!" || e++;
			}
			n = r;
		} while (n);
	}
	function ef(e) {
		var t = e.firstChild;
		for (t && t.nodeType === 10 && (t = t.nextSibling); t;) {
			var n = t;
			switch (t = t.nextSibling, n.nodeName) {
				case "HTML":
				case "HEAD":
				case "BODY":
					ef(n), gt(n);
					continue;
				case "SCRIPT":
				case "STYLE": continue;
				case "LINK": if (n.rel.toLowerCase() === "stylesheet") continue;
			}
			e.removeChild(n);
		}
	}
	function tf(e, t, n, r) {
		for (; e.nodeType === 1;) {
			var i = n;
			if (e.nodeName.toLowerCase() !== t.toLowerCase()) {
				if (!r && (e.nodeName !== "INPUT" || e.type !== "hidden")) break;
			} else if (!r) {
				if (t === "input" && e.type === "hidden") {
					var a = i.name == null ? null : "" + i.name;
					if (i.type === "hidden" && e.getAttribute("name") === a) return e;
				} else return e;
			} else if (!e[ht]) switch (t) {
				case "meta":
					if (!e.hasAttribute("itemprop")) break;
					return e;
				case "link":
					if (a = e.getAttribute("rel"), a === "stylesheet" && e.hasAttribute("data-precedence") || a !== i.rel || e.getAttribute("href") !== (i.href == null || i.href === "" ? null : i.href) || e.getAttribute("crossorigin") !== (i.crossOrigin == null ? null : i.crossOrigin) || e.getAttribute("title") !== (i.title == null ? null : i.title)) break;
					return e;
				case "style":
					if (e.hasAttribute("data-precedence")) break;
					return e;
				case "script":
					if (a = e.getAttribute("src"), (a !== (i.src == null ? null : i.src) || e.getAttribute("type") !== (i.type == null ? null : i.type) || e.getAttribute("crossorigin") !== (i.crossOrigin == null ? null : i.crossOrigin)) && a && e.hasAttribute("async") && !e.hasAttribute("itemprop")) break;
					return e;
				default: return e;
			}
			if (e = cf(e.nextSibling), e === null) break;
		}
		return null;
	}
	function nf(e, t, n) {
		if (t === "") return null;
		for (; e.nodeType !== 3;) if ((e.nodeType !== 1 || e.nodeName !== "INPUT" || e.type !== "hidden") && !n || (e = cf(e.nextSibling), e === null)) return null;
		return e;
	}
	function rf(e, t) {
		for (; e.nodeType !== 8;) if ((e.nodeType !== 1 || e.nodeName !== "INPUT" || e.type !== "hidden") && !t || (e = cf(e.nextSibling), e === null)) return null;
		return e;
	}
	function af(e) {
		return e.data === "$?" || e.data === "$~";
	}
	function of(e) {
		return e.data === "$!" || e.data === "$?" && e.ownerDocument.readyState !== "loading";
	}
	function sf(e, t) {
		var n = e.ownerDocument;
		if (e.data === "$~") e._reactRetry = t;
		else if (e.data !== "$?" || n.readyState !== "loading") t();
		else {
			var r = function() {
				t(), n.removeEventListener("DOMContentLoaded", r);
			};
			n.addEventListener("DOMContentLoaded", r), e._reactRetry = r;
		}
	}
	function cf(e) {
		for (; e != null; e = e.nextSibling) {
			var t = e.nodeType;
			if (t === 1 || t === 3) break;
			if (t === 8) {
				if (t = e.data, t === "$" || t === "$!" || t === "$?" || t === "$~" || t === "&" || t === "F!" || t === "F") break;
				if (t === "/$" || t === "/&") return null;
			}
		}
		return e;
	}
	var lf = null;
	function uf(e) {
		e = e.nextSibling;
		for (var t = 0; e;) {
			if (e.nodeType === 8) {
				var n = e.data;
				if (n === "/$" || n === "/&") {
					if (t === 0) return cf(e.nextSibling);
					t--;
				} else n !== "$" && n !== "$!" && n !== "$?" && n !== "$~" && n !== "&" || t++;
			}
			e = e.nextSibling;
		}
		return null;
	}
	function df(e) {
		e = e.previousSibling;
		for (var t = 0; e;) {
			if (e.nodeType === 8) {
				var n = e.data;
				if (n === "$" || n === "$!" || n === "$?" || n === "$~" || n === "&") {
					if (t === 0) return e;
					t--;
				} else n !== "/$" && n !== "/&" || t++;
			}
			e = e.previousSibling;
		}
		return null;
	}
	function ff(e, t, n) {
		switch (t = Bd(n), e) {
			case "html":
				if (e = t.documentElement, !e) throw Error(o(452));
				return e;
			case "head":
				if (e = t.head, !e) throw Error(o(453));
				return e;
			case "body":
				if (e = t.body, !e) throw Error(o(454));
				return e;
			default: throw Error(o(451));
		}
	}
	function pf(e) {
		for (var t = e.attributes; t.length;) e.removeAttributeNode(t[0]);
		gt(e);
	}
	var mf = /* @__PURE__ */ new Map(), hf = /* @__PURE__ */ new Set();
	function gf(e) {
		return typeof e.getRootNode == "function" ? e.getRootNode() : e.nodeType === 9 ? e : e.ownerDocument;
	}
	var _f = j.d;
	j.d = {
		f: vf,
		r: yf,
		D: Sf,
		C: Cf,
		L: wf,
		m: Tf,
		X: Df,
		S: Ef,
		M: Of
	};
	function vf() {
		var e = _f.f(), t = bu();
		return e || t;
	}
	function yf(e) {
		var t = vt(e);
		t !== null && t.tag === 5 && t.type === "form" ? Ds(t) : _f.r(e);
	}
	var bf = typeof document > "u" ? null : document;
	function xf(e, t, n) {
		var r = bf;
		if (r && typeof t == "string" && t) {
			var i = Bt(t);
			i = "link[rel=\"" + e + "\"][href=\"" + i + "\"]", typeof n == "string" && (i += "[crossorigin=\"" + n + "\"]"), hf.has(i) || (hf.add(i), e = {
				rel: e,
				crossOrigin: n,
				href: t
			}, r.querySelector(i) === null && (t = r.createElement("link"), Pd(t, "link", e), xt(t), r.head.appendChild(t)));
		}
	}
	function Sf(e) {
		_f.D(e), xf("dns-prefetch", e, null);
	}
	function Cf(e, t) {
		_f.C(e, t), xf("preconnect", e, t);
	}
	function wf(e, t, n) {
		_f.L(e, t, n);
		var r = bf;
		if (r && e && t) {
			var i = "link[rel=\"preload\"][as=\"" + Bt(t) + "\"]";
			t === "image" && n && n.imageSrcSet ? (i += "[imagesrcset=\"" + Bt(n.imageSrcSet) + "\"]", typeof n.imageSizes == "string" && (i += "[imagesizes=\"" + Bt(n.imageSizes) + "\"]")) : i += "[href=\"" + Bt(e) + "\"]";
			var a = i;
			switch (t) {
				case "style":
					a = Af(e);
					break;
				case "script": a = Pf(e);
			}
			mf.has(a) || (e = m({
				rel: "preload",
				href: t === "image" && n && n.imageSrcSet ? void 0 : e,
				as: t
			}, n), mf.set(a, e), r.querySelector(i) !== null || t === "style" && r.querySelector(jf(a)) || t === "script" && r.querySelector(Ff(a)) || (t = r.createElement("link"), Pd(t, "link", e), xt(t), r.head.appendChild(t)));
		}
	}
	function Tf(e, t) {
		_f.m(e, t);
		var n = bf;
		if (n && e) {
			var r = t && typeof t.as == "string" ? t.as : "script", i = "link[rel=\"modulepreload\"][as=\"" + Bt(r) + "\"][href=\"" + Bt(e) + "\"]", a = i;
			switch (r) {
				case "audioworklet":
				case "paintworklet":
				case "serviceworker":
				case "sharedworker":
				case "worker":
				case "script": a = Pf(e);
			}
			if (!mf.has(a) && (e = m({
				rel: "modulepreload",
				href: e
			}, t), mf.set(a, e), n.querySelector(i) === null)) {
				switch (r) {
					case "audioworklet":
					case "paintworklet":
					case "serviceworker":
					case "sharedworker":
					case "worker":
					case "script": if (n.querySelector(Ff(a))) return;
				}
				r = n.createElement("link"), Pd(r, "link", e), xt(r), n.head.appendChild(r);
			}
		}
	}
	function Ef(e, t, n) {
		_f.S(e, t, n);
		var r = bf;
		if (r && e) {
			var i = bt(r).hoistableStyles, a = Af(e);
			t ||= "default";
			var o = i.get(a);
			if (!o) {
				var s = {
					loading: 0,
					preload: null
				};
				if (o = r.querySelector(jf(a))) s.loading = 5;
				else {
					e = m({
						rel: "stylesheet",
						href: e,
						"data-precedence": t
					}, n), (n = mf.get(a)) && Rf(e, n);
					var c = o = r.createElement("link");
					xt(c), Pd(c, "link", e), c._p = new Promise(function(e, t) {
						c.onload = e, c.onerror = t;
					}), c.addEventListener("load", function() {
						s.loading |= 1;
					}), c.addEventListener("error", function() {
						s.loading |= 2;
					}), s.loading |= 4, Lf(o, t, r);
				}
				o = {
					type: "stylesheet",
					instance: o,
					count: 1,
					state: s
				}, i.set(a, o);
			}
		}
	}
	function Df(e, t) {
		_f.X(e, t);
		var n = bf;
		if (n && e) {
			var r = bt(n).hoistableScripts, i = Pf(e), a = r.get(i);
			a || (a = n.querySelector(Ff(i)), a || (e = m({
				src: e,
				async: !0
			}, t), (t = mf.get(i)) && zf(e, t), a = n.createElement("script"), xt(a), Pd(a, "link", e), n.head.appendChild(a)), a = {
				type: "script",
				instance: a,
				count: 1,
				state: null
			}, r.set(i, a));
		}
	}
	function Of(e, t) {
		_f.M(e, t);
		var n = bf;
		if (n && e) {
			var r = bt(n).hoistableScripts, i = Pf(e), a = r.get(i);
			a || (a = n.querySelector(Ff(i)), a || (e = m({
				src: e,
				async: !0,
				type: "module"
			}, t), (t = mf.get(i)) && zf(e, t), a = n.createElement("script"), xt(a), Pd(a, "link", e), n.head.appendChild(a)), a = {
				type: "script",
				instance: a,
				count: 1,
				state: null
			}, r.set(i, a));
		}
	}
	function kf(e, t, n, r) {
		var i = (i = de.current) ? gf(i) : null;
		if (!i) throw Error(o(446));
		switch (e) {
			case "meta":
			case "title": return null;
			case "style": return typeof n.precedence == "string" && typeof n.href == "string" ? (t = Af(n.href), n = bt(i).hoistableStyles, r = n.get(t), r || (r = {
				type: "style",
				instance: null,
				count: 0,
				state: null
			}, n.set(t, r)), r) : {
				type: "void",
				instance: null,
				count: 0,
				state: null
			};
			case "link":
				if (n.rel === "stylesheet" && typeof n.href == "string" && typeof n.precedence == "string") {
					e = Af(n.href);
					var a = bt(i).hoistableStyles, s = a.get(e);
					if (s || (i = i.ownerDocument || i, s = {
						type: "stylesheet",
						instance: null,
						count: 0,
						state: {
							loading: 0,
							preload: null
						}
					}, a.set(e, s), (a = i.querySelector(jf(e))) && !a._p && (s.instance = a, s.state.loading = 5), mf.has(e) || (n = {
						rel: "preload",
						as: "style",
						href: n.href,
						crossOrigin: n.crossOrigin,
						integrity: n.integrity,
						media: n.media,
						hrefLang: n.hrefLang,
						referrerPolicy: n.referrerPolicy
					}, mf.set(e, n), a || Nf(i, e, n, s.state))), t && r === null) throw Error(o(528, ""));
					return s;
				}
				if (t && r !== null) throw Error(o(529, ""));
				return null;
			case "script": return t = n.async, n = n.src, typeof n == "string" && t && typeof t != "function" && typeof t != "symbol" ? (t = Pf(n), n = bt(i).hoistableScripts, r = n.get(t), r || (r = {
				type: "script",
				instance: null,
				count: 0,
				state: null
			}, n.set(t, r)), r) : {
				type: "void",
				instance: null,
				count: 0,
				state: null
			};
			default: throw Error(o(444, e));
		}
	}
	function Af(e) {
		return "href=\"" + Bt(e) + "\"";
	}
	function jf(e) {
		return "link[rel=\"stylesheet\"][" + e + "]";
	}
	function Mf(e) {
		return m({}, e, {
			"data-precedence": e.precedence,
			precedence: null
		});
	}
	function Nf(e, t, n, r) {
		e.querySelector("link[rel=\"preload\"][as=\"style\"][" + t + "]") ? r.loading = 1 : (t = e.createElement("link"), r.preload = t, t.addEventListener("load", function() {
			return r.loading |= 1;
		}), t.addEventListener("error", function() {
			return r.loading |= 2;
		}), Pd(t, "link", n), xt(t), e.head.appendChild(t));
	}
	function Pf(e) {
		return "[src=\"" + Bt(e) + "\"]";
	}
	function Ff(e) {
		return "script[async]" + e;
	}
	function If(e, t, n) {
		if (t.count++, t.instance === null) switch (t.type) {
			case "style":
				var r = e.querySelector("style[data-href~=\"" + Bt(n.href) + "\"]");
				if (r) return t.instance = r, xt(r), r;
				var i = m({}, n, {
					"data-href": n.href,
					"data-precedence": n.precedence,
					href: null,
					precedence: null
				});
				return r = (e.ownerDocument || e).createElement("style"), xt(r), Pd(r, "style", i), Lf(r, n.precedence, e), t.instance = r;
			case "stylesheet":
				i = Af(n.href);
				var a = e.querySelector(jf(i));
				if (a) return t.state.loading |= 4, t.instance = a, xt(a), a;
				r = Mf(n), (i = mf.get(i)) && Rf(r, i), a = (e.ownerDocument || e).createElement("link"), xt(a);
				var s = a;
				return s._p = new Promise(function(e, t) {
					s.onload = e, s.onerror = t;
				}), Pd(a, "link", r), t.state.loading |= 4, Lf(a, n.precedence, e), t.instance = a;
			case "script": return a = Pf(n.src), (i = e.querySelector(Ff(a))) ? (t.instance = i, xt(i), i) : (r = n, (i = mf.get(a)) && (r = m({}, n), zf(r, i)), e = e.ownerDocument || e, i = e.createElement("script"), xt(i), Pd(i, "link", r), e.head.appendChild(i), t.instance = i);
			case "void": return null;
			default: throw Error(o(443, t.type));
		}
		else t.type === "stylesheet" && !(t.state.loading & 4) && (r = t.instance, t.state.loading |= 4, Lf(r, n.precedence, e));
		return t.instance;
	}
	function Lf(e, t, n) {
		for (var r = n.querySelectorAll("link[rel=\"stylesheet\"][data-precedence],style[data-precedence]"), i = r.length ? r[r.length - 1] : null, a = i, o = 0; o < r.length; o++) {
			var s = r[o];
			if (s.dataset.precedence === t) a = s;
			else if (a !== i) break;
		}
		a ? a.parentNode.insertBefore(e, a.nextSibling) : (t = n.nodeType === 9 ? n.head : n, t.insertBefore(e, t.firstChild));
	}
	function Rf(e, t) {
		e.crossOrigin ??= t.crossOrigin, e.referrerPolicy ??= t.referrerPolicy, e.title ??= t.title;
	}
	function zf(e, t) {
		e.crossOrigin ??= t.crossOrigin, e.referrerPolicy ??= t.referrerPolicy, e.integrity ??= t.integrity;
	}
	var Bf = null;
	function Vf(e, t, n) {
		if (Bf === null) {
			var r = /* @__PURE__ */ new Map(), i = Bf = /* @__PURE__ */ new Map();
			i.set(n, r);
		} else i = Bf, r = i.get(n), r || (r = /* @__PURE__ */ new Map(), i.set(n, r));
		if (r.has(e)) return r;
		for (r.set(e, null), n = n.getElementsByTagName(e), i = 0; i < n.length; i++) {
			var a = n[i];
			if (!(a[ht] || a[ct] || e === "link" && a.getAttribute("rel") === "stylesheet") && a.namespaceURI !== "http://www.w3.org/2000/svg") {
				var o = a.getAttribute(t) || "";
				o = e + o;
				var s = r.get(o);
				s ? s.push(a) : r.set(o, [a]);
			}
		}
		return r;
	}
	function Hf(e, t, n) {
		e = e.ownerDocument || e, e.head.insertBefore(n, t === "title" ? e.querySelector("head > title") : null);
	}
	function Uf(e, t, n) {
		if (n === 1 || t.itemProp != null) return !1;
		switch (e) {
			case "meta":
			case "title": return !0;
			case "style":
				if (typeof t.precedence != "string" || typeof t.href != "string" || t.href === "") break;
				return !0;
			case "link":
				if (typeof t.rel != "string" || typeof t.href != "string" || t.href === "" || t.onLoad || t.onError) break;
				switch (t.rel) {
					case "stylesheet": return e = t.disabled, typeof t.precedence == "string" && e == null;
					default: return !0;
				}
			case "script": if (t.async && typeof t.async != "function" && typeof t.async != "symbol" && !t.onLoad && !t.onError && t.src && typeof t.src == "string") return !0;
		}
		return !1;
	}
	function Wf(e) {
		return !(e.type === "stylesheet" && !(e.state.loading & 3));
	}
	function Gf(e, t, n, r) {
		if (n.type === "stylesheet" && (typeof r.media != "string" || !1 !== matchMedia(r.media).matches) && !(n.state.loading & 4)) {
			if (n.instance === null) {
				var i = Af(r.href), a = t.querySelector(jf(i));
				if (a) {
					t = a._p, typeof t == "object" && t && typeof t.then == "function" && (e.count++, e = Jf.bind(e), t.then(e, e)), n.state.loading |= 4, n.instance = a, xt(a);
					return;
				}
				a = t.ownerDocument || t, r = Mf(r), (i = mf.get(i)) && Rf(r, i), a = a.createElement("link"), xt(a);
				var o = a;
				o._p = new Promise(function(e, t) {
					o.onload = e, o.onerror = t;
				}), Pd(a, "link", r), n.instance = a;
			}
			e.stylesheets === null && (e.stylesheets = /* @__PURE__ */ new Map()), e.stylesheets.set(n, t), (t = n.state.preload) && !(n.state.loading & 3) && (e.count++, n = Jf.bind(e), t.addEventListener("load", n), t.addEventListener("error", n));
		}
	}
	var Kf = 0;
	function qf(e, t) {
		return e.stylesheets && e.count === 0 && Xf(e, e.stylesheets), 0 < e.count || 0 < e.imgCount ? function(n) {
			var r = setTimeout(function() {
				if (e.stylesheets && Xf(e, e.stylesheets), e.unsuspend) {
					var t = e.unsuspend;
					e.unsuspend = null, t();
				}
			}, 6e4 + t);
			0 < e.imgBytes && Kf === 0 && (Kf = 62500 * Ld());
			var i = setTimeout(function() {
				if (e.waitingForImages = !1, e.count === 0 && (e.stylesheets && Xf(e, e.stylesheets), e.unsuspend)) {
					var t = e.unsuspend;
					e.unsuspend = null, t();
				}
			}, (e.imgBytes > Kf ? 50 : 800) + t);
			return e.unsuspend = n, function() {
				e.unsuspend = null, clearTimeout(r), clearTimeout(i);
			};
		} : null;
	}
	function Jf() {
		if (this.count--, this.count === 0 && (this.imgCount === 0 || !this.waitingForImages)) {
			if (this.stylesheets) Xf(this, this.stylesheets);
			else if (this.unsuspend) {
				var e = this.unsuspend;
				this.unsuspend = null, e();
			}
		}
	}
	var Yf = null;
	function Xf(e, t) {
		e.stylesheets = null, e.unsuspend !== null && (e.count++, Yf = /* @__PURE__ */ new Map(), t.forEach(Zf, e), Yf = null, Jf.call(e));
	}
	function Zf(e, t) {
		if (!(t.state.loading & 4)) {
			var n = Yf.get(e);
			if (n) var r = n.get(null);
			else {
				n = /* @__PURE__ */ new Map(), Yf.set(e, n);
				for (var i = e.querySelectorAll("link[data-precedence],style[data-precedence]"), a = 0; a < i.length; a++) {
					var o = i[a];
					(o.nodeName === "LINK" || o.getAttribute("media") !== "not all") && (n.set(o.dataset.precedence, o), r = o);
				}
				r && n.set(null, r);
			}
			i = t.instance, o = i.getAttribute("data-precedence"), a = n.get(o) || r, a === r && n.set(null, i), n.set(o, i), this.count++, r = Jf.bind(this), i.addEventListener("load", r), i.addEventListener("error", r), a ? a.parentNode.insertBefore(i, a.nextSibling) : (e = e.nodeType === 9 ? e.head : e, e.insertBefore(i, e.firstChild)), t.state.loading |= 4;
		}
	}
	var Qf = {
		$$typeof: S,
		Provider: null,
		Consumer: null,
		_currentValue: ae,
		_currentValue2: ae,
		_threadCount: 0
	};
	function $f(e, t, n, r, i, a, o, s, c) {
		this.tag = 1, this.containerInfo = e, this.pingCache = this.current = this.pendingChildren = null, this.timeoutHandle = -1, this.callbackNode = this.next = this.pendingContext = this.context = this.cancelPendingCommit = null, this.callbackPriority = 0, this.expirationTimes = Ze(-1), this.entangledLanes = this.shellSuspendCounter = this.errorRecoveryDisabledLanes = this.expiredLanes = this.warmLanes = this.pingedLanes = this.suspendedLanes = this.pendingLanes = 0, this.entanglements = Ze(0), this.hiddenUpdates = Ze(null), this.identifierPrefix = r, this.onUncaughtError = i, this.onCaughtError = a, this.onRecoverableError = o, this.pooledCache = null, this.pooledCacheLanes = 0, this.formState = c, this.incompleteTransitions = /* @__PURE__ */ new Map();
	}
	function ep(e, t, n, r, i, a, o, s, c, l, u, d) {
		return e = new $f(e, t, n, o, c, l, u, d, s), t = 1, !0 === a && (t |= 24), a = li(3, null, null, t), e.current = a, a.stateNode = e, t = la(), t.refCount++, e.pooledCache = t, t.refCount++, a.memoizedState = {
			element: r,
			isDehydrated: n,
			cache: t
		}, Ha(a), e;
	}
	function tp(e) {
		return e ? (e = si, e) : si;
	}
	function np(e, t, n, r, i, a) {
		i = tp(i), r.context === null ? r.context = i : r.pendingContext = i, r = Wa(t), r.payload = { element: n }, a = a === void 0 ? null : a, a !== null && (r.callback = a), n = Ga(e, r, t), n !== null && (hu(n, e, t), Ka(n, e, t));
	}
	function rp(e, t) {
		if (e = e.memoizedState, e !== null && e.dehydrated !== null) {
			var n = e.retryLane;
			e.retryLane = n !== 0 && n < t ? n : t;
		}
	}
	function ip(e, t) {
		rp(e, t), (e = e.alternate) && rp(e, t);
	}
	function ap(e) {
		if (e.tag === 13 || e.tag === 31) {
			var t = ii(e, 67108864);
			t !== null && hu(t, e, 67108864), ip(e, 67108864);
		}
	}
	function op(e) {
		if (e.tag === 13 || e.tag === 31) {
			var t = pu();
			t = rt(t);
			var n = ii(e, t);
			n !== null && hu(n, e, t), ip(e, t);
		}
	}
	var sp = !0;
	function cp(e, t, n, r) {
		var i = A.T;
		A.T = null;
		var a = j.p;
		try {
			j.p = 2, up(e, t, n, r);
		} finally {
			j.p = a, A.T = i;
		}
	}
	function lp(e, t, n, r) {
		var i = A.T;
		A.T = null;
		var a = j.p;
		try {
			j.p = 8, up(e, t, n, r);
		} finally {
			j.p = a, A.T = i;
		}
	}
	function up(e, t, n, r) {
		if (sp) {
			var i = dp(r);
			if (i === null) wd(e, t, r, fp, n), Cp(e, r);
			else if (Tp(i, e, t, n, r)) r.stopPropagation();
			else if (Cp(e, r), t & 4 && -1 < Sp.indexOf(e)) {
				for (; i !== null;) {
					var a = vt(i);
					if (a !== null) switch (a.tag) {
						case 3:
							if (a = a.stateNode, a.current.memoizedState.isDehydrated) {
								var o = Ke(a.pendingLanes);
								if (o !== 0) {
									var s = a;
									for (s.pendingLanes |= 2, s.entangledLanes |= 2; o;) {
										var c = 1 << 31 - ze(o);
										s.entanglements[1] |= c, o &= ~c;
									}
									rd(a), !(K & 6) && (tu = De() + 500, id(0, !1));
								}
							}
							break;
						case 31:
						case 13: s = ii(a, 2), s !== null && hu(s, a, 2), bu(), ip(a, 2);
					}
					if (a = dp(r), a === null && wd(e, t, r, fp, n), a === i) break;
					i = a;
				}
				i !== null && r.stopPropagation();
			} else wd(e, t, r, null, n);
		}
	}
	function dp(e) {
		return e = nn(e), pp(e);
	}
	var fp = null;
	function pp(e) {
		if (fp = null, e = _t(e), e !== null) {
			var t = c(e);
			if (t === null) e = null;
			else {
				var n = t.tag;
				if (n === 13) {
					if (e = l(t), e !== null) return e;
					e = null;
				} else if (n === 31) {
					if (e = u(t), e !== null) return e;
					e = null;
				} else if (n === 3) {
					if (t.stateNode.current.memoizedState.isDehydrated) return t.tag === 3 ? t.stateNode.containerInfo : null;
					e = null;
				} else t !== e && (e = null);
			}
		}
		return fp = e, null;
	}
	function mp(e) {
		switch (e) {
			case "beforetoggle":
			case "cancel":
			case "click":
			case "close":
			case "contextmenu":
			case "copy":
			case "cut":
			case "auxclick":
			case "dblclick":
			case "dragend":
			case "dragstart":
			case "drop":
			case "focusin":
			case "focusout":
			case "input":
			case "invalid":
			case "keydown":
			case "keypress":
			case "keyup":
			case "mousedown":
			case "mouseup":
			case "paste":
			case "pause":
			case "play":
			case "pointercancel":
			case "pointerdown":
			case "pointerup":
			case "ratechange":
			case "reset":
			case "resize":
			case "seeked":
			case "submit":
			case "toggle":
			case "touchcancel":
			case "touchend":
			case "touchstart":
			case "volumechange":
			case "change":
			case "selectionchange":
			case "textInput":
			case "compositionstart":
			case "compositionend":
			case "compositionupdate":
			case "beforeblur":
			case "afterblur":
			case "beforeinput":
			case "blur":
			case "fullscreenchange":
			case "focus":
			case "hashchange":
			case "popstate":
			case "select":
			case "selectstart": return 2;
			case "drag":
			case "dragenter":
			case "dragexit":
			case "dragleave":
			case "dragover":
			case "mousemove":
			case "mouseout":
			case "mouseover":
			case "pointermove":
			case "pointerout":
			case "pointerover":
			case "scroll":
			case "touchmove":
			case "wheel":
			case "mouseenter":
			case "mouseleave":
			case "pointerenter":
			case "pointerleave": return 8;
			case "message": switch (Oe()) {
				case ke: return 2;
				case Ae: return 8;
				case je:
				case Me: return 32;
				case Ne: return 268435456;
				default: return 32;
			}
			default: return 32;
		}
	}
	var hp = !1, gp = null, _p = null, vp = null, yp = /* @__PURE__ */ new Map(), bp = /* @__PURE__ */ new Map(), xp = [], Sp = "mousedown mouseup touchcancel touchend touchstart auxclick dblclick pointercancel pointerdown pointerup dragend dragstart drop compositionend compositionstart keydown keypress keyup input textInput copy cut paste click change contextmenu reset".split(" ");
	function Cp(e, t) {
		switch (e) {
			case "focusin":
			case "focusout":
				gp = null;
				break;
			case "dragenter":
			case "dragleave":
				_p = null;
				break;
			case "mouseover":
			case "mouseout":
				vp = null;
				break;
			case "pointerover":
			case "pointerout":
				yp.delete(t.pointerId);
				break;
			case "gotpointercapture":
			case "lostpointercapture": bp.delete(t.pointerId);
		}
	}
	function wp(e, t, n, r, i, a) {
		return e === null || e.nativeEvent !== a ? (e = {
			blockedOn: t,
			domEventName: n,
			eventSystemFlags: r,
			nativeEvent: a,
			targetContainers: [i]
		}, t !== null && (t = vt(t), t !== null && ap(t)), e) : (e.eventSystemFlags |= r, t = e.targetContainers, i !== null && t.indexOf(i) === -1 && t.push(i), e);
	}
	function Tp(e, t, n, r, i) {
		switch (t) {
			case "focusin": return gp = wp(gp, e, t, n, r, i), !0;
			case "dragenter": return _p = wp(_p, e, t, n, r, i), !0;
			case "mouseover": return vp = wp(vp, e, t, n, r, i), !0;
			case "pointerover":
				var a = i.pointerId;
				return yp.set(a, wp(yp.get(a) || null, e, t, n, r, i)), !0;
			case "gotpointercapture": return a = i.pointerId, bp.set(a, wp(bp.get(a) || null, e, t, n, r, i)), !0;
		}
		return !1;
	}
	function Ep(e) {
		var t = _t(e.target);
		if (t !== null) {
			var n = c(t);
			if (n !== null) {
				if (t = n.tag, t === 13) {
					if (t = l(n), t !== null) {
						e.blockedOn = t, ot(e.priority, function() {
							op(n);
						});
						return;
					}
				} else if (t === 31) {
					if (t = u(n), t !== null) {
						e.blockedOn = t, ot(e.priority, function() {
							op(n);
						});
						return;
					}
				} else if (t === 3 && n.stateNode.current.memoizedState.isDehydrated) {
					e.blockedOn = n.tag === 3 ? n.stateNode.containerInfo : null;
					return;
				}
			}
		}
		e.blockedOn = null;
	}
	function Dp(e) {
		if (e.blockedOn !== null) return !1;
		for (var t = e.targetContainers; 0 < t.length;) {
			var n = dp(e.nativeEvent);
			if (n === null) {
				n = e.nativeEvent;
				var r = new n.constructor(n.type, n);
				tn = r, n.target.dispatchEvent(r), tn = null;
			} else return t = vt(n), t !== null && ap(t), e.blockedOn = n, !1;
			t.shift();
		}
		return !0;
	}
	function Op(e, t, n) {
		Dp(e) && n.delete(t);
	}
	function kp() {
		hp = !1, gp !== null && Dp(gp) && (gp = null), _p !== null && Dp(_p) && (_p = null), vp !== null && Dp(vp) && (vp = null), yp.forEach(Op), bp.forEach(Op);
	}
	function Ap(e, n) {
		e.blockedOn === n && (e.blockedOn = null, hp || (hp = !0, t.unstable_scheduleCallback(t.unstable_NormalPriority, kp)));
	}
	var jp = null;
	function Mp(e) {
		jp !== e && (jp = e, t.unstable_scheduleCallback(t.unstable_NormalPriority, function() {
			jp === e && (jp = null);
			for (var t = 0; t < e.length; t += 3) {
				var n = e[t], r = e[t + 1], i = e[t + 2];
				if (typeof r != "function") {
					if (pp(r || n) === null) continue;
					break;
				}
				var a = vt(n);
				a !== null && (e.splice(t, 3), t -= 3, Ts(a, {
					pending: !0,
					data: i,
					method: n.method,
					action: r
				}, r, i));
			}
		}));
	}
	function Np(e) {
		function t(t) {
			return Ap(t, e);
		}
		gp !== null && Ap(gp, e), _p !== null && Ap(_p, e), vp !== null && Ap(vp, e), yp.forEach(t), bp.forEach(t);
		for (var n = 0; n < xp.length; n++) {
			var r = xp[n];
			r.blockedOn === e && (r.blockedOn = null);
		}
		for (; 0 < xp.length && (n = xp[0], n.blockedOn === null);) Ep(n), n.blockedOn === null && xp.shift();
		if (n = (e.ownerDocument || e).$$reactFormReplay, n != null) for (r = 0; r < n.length; r += 3) {
			var i = n[r], a = n[r + 1], o = i[lt] || null;
			if (typeof a == "function") o || Mp(n);
			else if (o) {
				var s = null;
				if (a && a.hasAttribute("formAction")) {
					if (i = a, o = a[lt] || null) s = o.formAction;
					else if (pp(i) !== null) continue;
				} else s = o.action;
				typeof s == "function" ? n[r + 1] = s : (n.splice(r, 3), r -= 3), Mp(n);
			}
		}
	}
	function Pp() {
		function e(e) {
			e.canIntercept && e.info === "react-transition" && e.intercept({
				handler: function() {
					return new Promise(function(e) {
						return i = e;
					});
				},
				focusReset: "manual",
				scroll: "manual"
			});
		}
		function t() {
			i !== null && (i(), i = null), r || setTimeout(n, 20);
		}
		function n() {
			if (!r && !navigation.transition) {
				var e = navigation.currentEntry;
				e && e.url != null && navigation.navigate(e.url, {
					state: e.getState(),
					info: "react-transition",
					history: "replace"
				});
			}
		}
		if (typeof navigation == "object") {
			var r = !1, i = null;
			return navigation.addEventListener("navigate", e), navigation.addEventListener("navigatesuccess", t), navigation.addEventListener("navigateerror", t), setTimeout(n, 100), function() {
				r = !0, navigation.removeEventListener("navigate", e), navigation.removeEventListener("navigatesuccess", t), navigation.removeEventListener("navigateerror", t), i !== null && (i(), i = null);
			};
		}
	}
	function Fp(e) {
		this._internalRoot = e;
	}
	Ip.prototype.render = Fp.prototype.render = function(e) {
		var t = this._internalRoot;
		if (t === null) throw Error(o(409));
		var n = t.current;
		np(n, pu(), e, t, null, null);
	}, Ip.prototype.unmount = Fp.prototype.unmount = function() {
		var e = this._internalRoot;
		if (e !== null) {
			this._internalRoot = null;
			var t = e.containerInfo;
			np(e.current, 2, null, e, null, null), bu(), t[ut] = null;
		}
	};
	function Ip(e) {
		this._internalRoot = e;
	}
	Ip.prototype.unstable_scheduleHydration = function(e) {
		if (e) {
			var t = at();
			e = {
				blockedOn: null,
				target: e,
				priority: t
			};
			for (var n = 0; n < xp.length && t !== 0 && t < xp[n].priority; n++);
			xp.splice(n, 0, e), n === 0 && Ep(e);
		}
	};
	var Lp = r.version;
	if (Lp !== "19.2.7") throw Error(o(527, Lp, "19.2.7"));
	j.findDOMNode = function(e) {
		var t = e._reactInternals;
		if (t === void 0) throw typeof e.render == "function" ? Error(o(188)) : (e = Object.keys(e).join(","), Error(o(268, e)));
		return e = f(t), e = e === null ? null : p(e), e = e === null ? null : e.stateNode, e;
	};
	var Rp = {
		bundleType: 0,
		version: "19.2.7",
		rendererPackageName: "react-dom",
		currentDispatcherRef: A,
		reconcilerVersion: "19.2.7"
	};
	if (typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ < "u") {
		var zp = __REACT_DEVTOOLS_GLOBAL_HOOK__;
		if (!zp.isDisabled && zp.supportsFiber) try {
			Ie = zp.inject(Rp), Le = zp;
		} catch {}
	}
	e.createRoot = function(e, t) {
		if (!s(e)) throw Error(o(299));
		var n = !1, r = "", i = qs, a = Js, c = Ys;
		return t != null && (!0 === t.unstable_strictMode && (n = !0), t.identifierPrefix !== void 0 && (r = t.identifierPrefix), t.onUncaughtError !== void 0 && (i = t.onUncaughtError), t.onCaughtError !== void 0 && (a = t.onCaughtError), t.onRecoverableError !== void 0 && (c = t.onRecoverableError)), t = ep(e, 1, !1, null, null, n, r, null, i, a, c, Pp), e[ut] = t.current, Sd(e), new Fp(t);
	};
})), Jt = (/* @__PURE__ */ t(((e, t) => {
	function n() {
		if (typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ < "u" && typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE == "function") try {
			__REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(n);
		} catch (e) {
			console.error(e);
		}
	}
	n(), t.exports = qt();
})))(), Yt = "/*! tailwindcss v4.3.3 | MIT License | https://tailwindcss.com */\n@layer properties{@supports (((-webkit-hyphens:none)) and (not (margin-trim:inline))) or ((-moz-orient:inline) and (not (color:rgb(from red r g b)))){*,:before,:after,::backdrop{--tw-border-spacing-x:0;--tw-border-spacing-y:0;--tw-translate-x:0;--tw-translate-y:0;--tw-translate-z:0;--tw-scale-x:1;--tw-scale-y:1;--tw-scale-z:1;--tw-rotate-x:initial;--tw-rotate-y:initial;--tw-rotate-z:initial;--tw-skew-x:initial;--tw-skew-y:initial;--tw-space-y-reverse:0;--tw-divide-x-reverse:0;--tw-border-style:solid;--tw-divide-y-reverse:0;--tw-gradient-position:initial;--tw-gradient-from:#0000;--tw-gradient-via:#0000;--tw-gradient-to:#0000;--tw-gradient-stops:initial;--tw-gradient-via-stops:initial;--tw-gradient-from-position:0%;--tw-gradient-via-position:50%;--tw-gradient-to-position:100%;--tw-leading:initial;--tw-font-weight:initial;--tw-tracking:initial;--tw-ordinal:initial;--tw-slashed-zero:initial;--tw-numeric-figure:initial;--tw-numeric-spacing:initial;--tw-numeric-fraction:initial;--tw-shadow:0 0 #0000;--tw-shadow-color:initial;--tw-shadow-alpha:100%;--tw-inset-shadow:0 0 #0000;--tw-inset-shadow-color:initial;--tw-inset-shadow-alpha:100%;--tw-ring-color:initial;--tw-ring-shadow:0 0 #0000;--tw-inset-ring-color:initial;--tw-inset-ring-shadow:0 0 #0000;--tw-ring-inset:initial;--tw-ring-offset-width:0;--tw-ring-offset-color:#fff;--tw-ring-offset-shadow:0 0 #0000;--tw-outline-style:solid;--tw-blur:initial;--tw-brightness:initial;--tw-contrast:initial;--tw-grayscale:initial;--tw-hue-rotate:initial;--tw-invert:initial;--tw-opacity:initial;--tw-saturate:initial;--tw-sepia:initial;--tw-drop-shadow:initial;--tw-drop-shadow-color:initial;--tw-drop-shadow-alpha:100%;--tw-drop-shadow-size:initial;--tw-backdrop-blur:initial;--tw-backdrop-brightness:initial;--tw-backdrop-contrast:initial;--tw-backdrop-grayscale:initial;--tw-backdrop-hue-rotate:initial;--tw-backdrop-invert:initial;--tw-backdrop-opacity:initial;--tw-backdrop-saturate:initial;--tw-backdrop-sepia:initial;--tw-duration:initial;--tw-ease:initial;--tw-content:\"\"}}}@layer theme{:root,:host{--color-red-50:oklch(97.1% .013 17.38);--color-red-100:oklch(93.6% .032 17.717);--color-red-200:oklch(88.5% .062 18.334);--color-red-300:oklch(80.8% .114 19.571);--color-red-400:oklch(70.4% .191 22.216);--color-red-500:oklch(63.7% .237 25.331);--color-red-600:oklch(57.7% .245 27.325);--color-red-700:oklch(50.5% .213 27.518);--color-red-800:oklch(44.4% .177 26.899);--color-red-900:oklch(39.6% .141 25.723);--color-red-950:oklch(25.8% .092 26.042);--color-orange-50:oklch(98% .016 73.684);--color-orange-100:oklch(95.4% .038 75.164);--color-orange-200:oklch(90.1% .076 70.697);--color-orange-300:oklch(83.7% .128 66.29);--color-orange-400:oklch(75% .183 55.934);--color-orange-500:oklch(70.5% .213 47.604);--color-orange-600:oklch(64.6% .222 41.116);--color-orange-700:oklch(55.3% .195 38.402);--color-orange-800:oklch(47% .157 37.304);--color-orange-900:oklch(40.8% .123 38.172);--color-orange-950:oklch(26.6% .079 36.259);--color-amber-50:oklch(98.7% .022 95.277);--color-amber-100:oklch(96.2% .059 95.617);--color-amber-200:oklch(92.4% .12 95.746);--color-amber-300:oklch(87.9% .169 91.605);--color-amber-400:oklch(82.8% .189 84.429);--color-amber-500:oklch(76.9% .188 70.08);--color-amber-600:oklch(66.6% .179 58.318);--color-amber-700:oklch(55.5% .163 48.998);--color-amber-800:oklch(47.3% .137 46.201);--color-amber-900:oklch(41.4% .112 45.904);--color-amber-950:oklch(27.9% .077 45.635);--color-yellow-50:oklch(98.7% .026 102.212);--color-yellow-100:oklch(97.3% .071 103.193);--color-yellow-200:oklch(94.5% .129 101.54);--color-yellow-300:oklch(90.5% .182 98.111);--color-yellow-400:oklch(85.2% .199 91.936);--color-yellow-500:oklch(79.5% .184 86.047);--color-yellow-600:oklch(68.1% .162 75.834);--color-yellow-700:oklch(55.4% .135 66.442);--color-yellow-800:oklch(47.6% .114 61.907);--color-yellow-900:oklch(42.1% .095 57.708);--color-yellow-950:oklch(28.6% .066 53.813);--color-lime-50:oklch(98.6% .031 120.757);--color-lime-100:oklch(96.7% .067 122.328);--color-lime-200:oklch(93.8% .127 124.321);--color-lime-300:oklch(89.7% .196 126.665);--color-lime-400:oklch(84.1% .238 128.85);--color-lime-500:oklch(76.8% .233 130.85);--color-lime-600:oklch(64.8% .2 131.684);--color-lime-700:oklch(53.2% .157 131.589);--color-lime-800:oklch(45.3% .124 130.933);--color-lime-900:oklch(40.5% .101 131.063);--color-lime-950:oklch(27.4% .072 132.109);--color-green-50:oklch(98.2% .018 155.826);--color-green-100:oklch(96.2% .044 156.743);--color-green-200:oklch(92.5% .084 155.995);--color-green-300:oklch(87.1% .15 154.449);--color-green-400:oklch(79.2% .209 151.711);--color-green-500:oklch(72.3% .219 149.579);--color-green-600:oklch(62.7% .194 149.214);--color-green-700:oklch(52.7% .154 150.069);--color-green-800:oklch(44.8% .119 151.328);--color-green-900:oklch(39.3% .095 152.535);--color-green-950:oklch(26.6% .065 152.934);--color-emerald-50:oklch(97.9% .021 166.113);--color-emerald-100:oklch(95% .052 163.051);--color-emerald-200:oklch(90.5% .093 164.15);--color-emerald-300:oklch(84.5% .143 164.978);--color-emerald-400:oklch(76.5% .177 163.223);--color-emerald-500:oklch(69.6% .17 162.48);--color-emerald-600:oklch(59.6% .145 163.225);--color-emerald-700:oklch(50.8% .118 165.612);--color-emerald-800:oklch(43.2% .095 166.913);--color-emerald-900:oklch(37.8% .077 168.94);--color-emerald-950:oklch(26.2% .051 172.552);--color-teal-50:oklch(98.4% .014 180.72);--color-teal-100:oklch(95.3% .051 180.801);--color-teal-200:oklch(91% .096 180.426);--color-teal-300:oklch(85.5% .138 181.071);--color-teal-400:oklch(77.7% .152 181.912);--color-teal-500:oklch(70.4% .14 182.503);--color-teal-600:oklch(60% .118 184.704);--color-teal-700:oklch(51.1% .096 186.391);--color-teal-800:oklch(43.7% .078 188.216);--color-teal-900:oklch(38.6% .063 188.416);--color-teal-950:oklch(27.7% .046 192.524);--color-cyan-50:oklch(98.4% .019 200.873);--color-cyan-100:oklch(95.6% .045 203.388);--color-cyan-200:oklch(91.7% .08 205.041);--color-cyan-300:oklch(86.5% .127 207.078);--color-cyan-400:oklch(78.9% .154 211.53);--color-cyan-500:oklch(71.5% .143 215.221);--color-cyan-600:oklch(60.9% .126 221.723);--color-cyan-700:oklch(52% .105 223.128);--color-cyan-800:oklch(45% .085 224.283);--color-cyan-900:oklch(39.8% .07 227.392);--color-cyan-950:oklch(30.2% .056 229.695);--color-sky-50:oklch(97.7% .013 236.62);--color-sky-100:oklch(95.1% .026 236.824);--color-sky-200:oklch(90.1% .058 230.902);--color-sky-300:oklch(82.8% .111 230.318);--color-sky-400:oklch(74.6% .16 232.661);--color-sky-500:oklch(68.5% .169 237.323);--color-sky-600:oklch(58.8% .158 241.966);--color-sky-700:oklch(50% .134 242.749);--color-sky-800:oklch(44.3% .11 240.79);--color-sky-900:oklch(39.1% .09 240.876);--color-sky-950:oklch(29.3% .066 243.157);--color-blue-50:oklch(97% .014 254.604);--color-blue-100:oklch(93.2% .032 255.585);--color-blue-200:oklch(88.2% .059 254.128);--color-blue-300:oklch(80.9% .105 251.813);--color-blue-400:oklch(70.7% .165 254.624);--color-blue-500:oklch(62.3% .214 259.815);--color-blue-600:oklch(54.6% .245 262.881);--color-blue-700:oklch(48.8% .243 264.376);--color-blue-800:oklch(42.4% .199 265.638);--color-blue-900:oklch(37.9% .146 265.522);--color-blue-950:oklch(28.2% .091 267.935);--color-indigo-50:oklch(96.2% .018 272.314);--color-indigo-100:oklch(93% .034 272.788);--color-indigo-200:oklch(87% .065 274.039);--color-indigo-300:oklch(78.5% .115 274.713);--color-indigo-400:oklch(67.3% .182 276.935);--color-indigo-500:oklch(58.5% .233 277.117);--color-indigo-600:oklch(51.1% .262 276.966);--color-indigo-700:oklch(45.7% .24 277.023);--color-indigo-800:oklch(39.8% .195 277.366);--color-indigo-900:oklch(35.9% .144 278.697);--color-indigo-950:oklch(25.7% .09 281.288);--color-violet-50:oklch(96.9% .016 293.756);--color-violet-100:oklch(94.3% .029 294.588);--color-violet-200:oklch(89.4% .057 293.283);--color-violet-300:oklch(81.1% .111 293.571);--color-violet-400:oklch(70.2% .183 293.541);--color-violet-500:oklch(60.6% .25 292.717);--color-violet-600:oklch(54.1% .281 293.009);--color-violet-700:oklch(49.1% .27 292.581);--color-violet-800:oklch(43.2% .232 292.759);--color-violet-900:oklch(38% .189 293.745);--color-violet-950:oklch(28.3% .141 291.089);--color-purple-50:oklch(97.7% .014 308.299);--color-purple-100:oklch(94.6% .033 307.174);--color-purple-200:oklch(90.2% .063 306.703);--color-purple-300:oklch(82.7% .119 306.383);--color-purple-400:oklch(71.4% .203 305.504);--color-purple-500:oklch(62.7% .265 303.9);--color-purple-600:oklch(55.8% .288 302.321);--color-purple-700:oklch(49.6% .265 301.924);--color-purple-800:oklch(43.8% .218 303.724);--color-purple-900:oklch(38.1% .176 304.987);--color-purple-950:oklch(29.1% .149 302.717);--color-fuchsia-50:oklch(97.7% .017 320.058);--color-fuchsia-100:oklch(95.2% .037 318.852);--color-fuchsia-200:oklch(90.3% .076 319.62);--color-fuchsia-300:oklch(83.3% .145 321.434);--color-fuchsia-400:oklch(74% .238 322.16);--color-fuchsia-500:oklch(66.7% .295 322.15);--color-fuchsia-600:oklch(59.1% .293 322.896);--color-fuchsia-700:oklch(51.8% .253 323.949);--color-fuchsia-800:oklch(45.2% .211 324.591);--color-fuchsia-900:oklch(40.1% .17 325.612);--color-fuchsia-950:oklch(29.3% .136 325.661);--color-pink-50:oklch(97.1% .014 343.198);--color-pink-100:oklch(94.8% .028 342.258);--color-pink-200:oklch(89.9% .061 343.231);--color-pink-300:oklch(82.3% .12 346.018);--color-pink-400:oklch(71.8% .202 349.761);--color-pink-500:oklch(65.6% .241 354.308);--color-pink-600:oklch(59.2% .249 .584);--color-pink-700:oklch(52.5% .223 3.958);--color-pink-800:oklch(45.9% .187 3.815);--color-pink-900:oklch(40.8% .153 2.432);--color-pink-950:oklch(28.4% .109 3.907);--color-rose-50:oklch(96.9% .015 12.422);--color-rose-100:oklch(94.1% .03 12.58);--color-rose-200:oklch(89.2% .058 10.001);--color-rose-300:oklch(81% .117 11.638);--color-rose-400:oklch(71.2% .194 13.428);--color-rose-500:oklch(64.5% .246 16.439);--color-rose-600:oklch(58.6% .253 17.585);--color-rose-700:oklch(51.4% .222 16.935);--color-rose-800:oklch(45.5% .188 13.697);--color-rose-900:oklch(41% .159 10.272);--color-rose-950:oklch(27.1% .105 12.094);--color-slate-50:oklch(98.4% .003 247.858);--color-slate-100:oklch(96.8% .007 247.896);--color-slate-200:oklch(92.9% .013 255.508);--color-slate-300:oklch(86.9% .022 252.894);--color-slate-400:oklch(70.4% .04 256.788);--color-slate-500:oklch(55.4% .046 257.417);--color-slate-600:oklch(44.6% .043 257.281);--color-slate-700:oklch(37.2% .044 257.287);--color-slate-800:oklch(27.9% .041 260.031);--color-slate-900:oklch(20.8% .042 265.755);--color-slate-950:oklch(12.9% .042 264.695);--color-gray-50:oklch(98.5% .002 247.839);--color-gray-100:oklch(96.7% .003 264.542);--color-gray-200:oklch(92.8% .006 264.531);--color-gray-300:oklch(87.2% .01 258.338);--color-gray-400:oklch(70.7% .022 261.325);--color-gray-500:oklch(55.1% .027 264.364);--color-gray-600:oklch(44.6% .03 256.802);--color-gray-700:oklch(37.3% .034 259.733);--color-gray-800:oklch(27.8% .033 256.848);--color-gray-900:oklch(21% .034 264.665);--color-gray-950:oklch(13% .028 261.692);--color-zinc-50:oklch(98.5% 0 none);--color-zinc-100:oklch(96.7% .001 286.375);--color-zinc-200:oklch(92% .004 286.32);--color-zinc-300:oklch(87.1% .006 286.286);--color-zinc-400:oklch(70.5% .015 286.067);--color-zinc-500:oklch(55.2% .016 285.938);--color-zinc-600:oklch(44.2% .017 285.786);--color-zinc-700:oklch(37% .013 285.805);--color-zinc-800:oklch(27.4% .006 286.033);--color-zinc-900:oklch(21% .006 285.885);--color-zinc-950:oklch(14.1% .005 285.823);--color-neutral-50:oklch(98.5% 0 none);--color-neutral-100:oklch(97% 0 none);--color-neutral-200:oklch(92.2% 0 none);--color-neutral-300:oklch(87% 0 none);--color-neutral-400:oklch(70.8% 0 none);--color-neutral-500:oklch(55.6% 0 none);--color-neutral-600:oklch(43.9% 0 none);--color-neutral-700:oklch(37.1% 0 none);--color-neutral-800:oklch(26.9% 0 none);--color-neutral-900:oklch(20.5% 0 none);--color-neutral-950:oklch(14.5% 0 none);--color-stone-50:oklch(98.5% .001 106.423);--color-stone-100:oklch(97% .001 106.424);--color-stone-200:oklch(92.3% .003 48.717);--color-stone-300:oklch(86.9% .005 56.366);--color-stone-400:oklch(70.9% .01 56.259);--color-stone-500:oklch(55.3% .013 58.071);--color-stone-600:oklch(44.4% .011 73.639);--color-stone-700:oklch(37.4% .01 67.558);--color-stone-800:oklch(26.8% .007 34.298);--color-stone-900:oklch(21.6% .006 56.043);--color-stone-950:oklch(14.7% .004 49.25);--color-black:#000;--color-white:#fff;--spacing:.25rem;--container-xs:20rem;--container-sm:24rem;--container-md:28rem;--container-lg:32rem;--container-xl:36rem;--container-2xl:42rem;--container-3xl:48rem;--container-4xl:56rem;--container-5xl:64rem;--container-6xl:72rem;--container-7xl:80rem;--text-xs:.75rem;--text-xs--line-height:calc(1 / .75);--text-sm:.875rem;--text-sm--line-height:calc(1.25 / .875);--text-base:1rem;--text-base--line-height:calc(1.5 / 1);--text-lg:1.125rem;--text-lg--line-height:calc(1.75 / 1.125);--text-xl:1.25rem;--text-xl--line-height:calc(1.75 / 1.25);--text-2xl:1.5rem;--text-2xl--line-height:calc(2 / 1.5);--text-3xl:1.875rem;--text-3xl--line-height:calc(2.25 / 1.875);--text-4xl:2.25rem;--text-4xl--line-height:calc(2.5 / 2.25);--font-weight-normal:400;--font-weight-medium:500;--font-weight-semibold:600;--font-weight-bold:700;--tracking-tight:-.025em;--tracking-normal:0em;--tracking-wide:.025em;--tracking-wider:.05em;--leading-tight:1.25;--leading-snug:1.375;--leading-normal:1.5;--leading-relaxed:1.625;--radius-xl:.75rem;--radius-2xl:1rem;--ease-out:cubic-bezier(0, 0, .2, 1);--animate-spin:spin 1s linear infinite;--animate-pulse:pulse 2s cubic-bezier(.4, 0, .6, 1) infinite;--blur-sm:8px;--default-transition-duration:.15s;--default-transition-timing-function:cubic-bezier(.4, 0, .2, 1);--default-font-family:var(--fs-font-sans);--default-font-feature-settings:normal;--default-font-variation-settings:normal;--default-mono-font-family:var(--fs-font-mono);--default-mono-font-feature-settings:normal;--default-mono-font-variation-settings:normal;--font-sans:var(--fs-font-sans);--font-mono:var(--fs-font-mono);--radius-sm:.25rem;--radius-md:.375rem;--radius-lg:.5rem;--spacing-density-1:calc(var(--spacing-unit,.25rem) * 1);--spacing-density-2:calc(var(--spacing-unit,.25rem) * 2);--spacing-density-3:calc(var(--spacing-unit,.25rem) * 3);--spacing-density-4:calc(var(--spacing-unit,.25rem) * 4);--spacing-density-6:calc(var(--spacing-unit,.25rem) * 6);--spacing-density-8:calc(var(--spacing-unit,.25rem) * 8);--spacing-control-h:var(--control-height,2.25rem);--spacing-control-px:var(--control-padding-x,.75rem);--container-8xl:96rem;--container-9xl:120rem;--color-primary:var(--primary);--color-destructive:var(--destructive);--color-placeholder:var(--placeholder)}}@layer base{@supports (not ((-webkit-appearance:-apple-pay-button))) or (contain-intrinsic-size:1px){::placeholder{color:currentColor}@supports (color:color-mix(in lab, red, red)){::placeholder{color:color-mix(in oklab, currentcolor 50%, transparent)}}}*,:after,:before,::backdrop{box-sizing:border-box;border:0 solid;margin:0;padding:0}::file-selector-button{box-sizing:border-box;border:0 solid;margin:0;padding:0}html,:host{-webkit-text-size-adjust:100%;tab-size:4;line-height:1.5;font-family:var(--default-font-family,-apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", \"Noto Sans\", Arial, sans-serif, \"Apple Color Emoji\", \"Segoe UI Emoji\", \"Segoe UI Symbol\", \"Noto Color Emoji\");font-feature-settings:var(--default-font-feature-settings,normal);font-variation-settings:var(--default-font-variation-settings,normal);-webkit-tap-highlight-color:transparent}hr{height:0;color:inherit;border-top-width:1px}abbr:where([title]){-webkit-text-decoration:underline dotted;text-decoration:underline dotted}h1,h2,h3,h4,h5,h6{font-size:inherit;font-weight:inherit}a{color:inherit;-webkit-text-decoration:inherit;-webkit-text-decoration:inherit;-webkit-text-decoration:inherit;-webkit-text-decoration:inherit;text-decoration:inherit}b,strong{font-weight:bolder}code,kbd,samp,pre{font-family:var(--default-mono-font-family,ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, \"Liberation Mono\", \"Courier New\", monospace);font-feature-settings:var(--default-mono-font-feature-settings,normal);font-variation-settings:var(--default-mono-font-variation-settings,normal);font-size:1em}small{font-size:80%}sub,sup{vertical-align:baseline;font-size:75%;line-height:0;position:relative}sub{bottom:-.25em}sup{top:-.5em}table{text-indent:0;border-color:inherit;border-collapse:collapse}:-moz-focusring:where(:not(iframe)){outline:auto}progress{vertical-align:baseline}summary{display:list-item}ol,ul,menu{list-style:none}img,svg,video,canvas,audio,iframe,embed,object{vertical-align:middle;display:block}img,video{max-width:100%;height:auto}button,input,select,optgroup,textarea{font:inherit;font-feature-settings:inherit;font-variation-settings:inherit;letter-spacing:inherit;color:inherit;opacity:1;background-color:#0000;border-radius:0}::file-selector-button{font:inherit;font-feature-settings:inherit;font-variation-settings:inherit;letter-spacing:inherit;color:inherit;opacity:1;background-color:#0000;border-radius:0}:where(select:is([multiple],[size])) optgroup{font-weight:bolder}:where(select:is([multiple],[size])) optgroup option{padding-inline-start:20px}::file-selector-button{margin-inline-end:4px}::placeholder{opacity:1}@supports (not ((-webkit-appearance:-apple-pay-button))) or (contain-intrinsic-size:1px){::placeholder{color:currentColor}@supports (color:color-mix(in lab, red, red)){::placeholder{color:color-mix(in oklab, currentcolor 50%, transparent)}}}textarea{resize:vertical}::-webkit-search-decoration{-webkit-appearance:none}::-webkit-date-and-time-value{min-height:1lh;text-align:inherit}::-webkit-datetime-edit{display:inline-flex}::-webkit-datetime-edit-fields-wrapper{padding:0}::-webkit-datetime-edit{padding-block:0}::-webkit-datetime-edit-year-field{padding-block:0}::-webkit-datetime-edit-month-field{padding-block:0}::-webkit-datetime-edit-day-field{padding-block:0}::-webkit-datetime-edit-hour-field{padding-block:0}::-webkit-datetime-edit-minute-field{padding-block:0}::-webkit-datetime-edit-second-field{padding-block:0}::-webkit-datetime-edit-millisecond-field{padding-block:0}::-webkit-datetime-edit-meridiem-field{padding-block:0}::-webkit-calendar-picker-indicator{line-height:1}:-moz-ui-invalid{box-shadow:none}button,input:where([type=button],[type=reset],[type=submit]){appearance:button}::file-selector-button{appearance:button}::-webkit-inner-spin-button{height:auto}::-webkit-outer-spin-button{height:auto}[hidden]:where(:not([hidden=until-found])){display:none!important}*{border-color:var(--border)}}@layer components;@layer utilities{.end-1{inset-inline-end:calc(var(--spacing) * 1)}@media (width>=1920px){.container{max-width:1920px}}@media (width>=40rem){.container{max-width:40rem}}@media (width>=48rem){.container{max-width:48rem}}@media (width>=64rem){.container{max-width:64rem}}@media (width>=80rem){.container{max-width:80rem}}@media (width>=96rem){.container{max-width:96rem}}.max-h-\\[45\\%\\]{max-height:45%}.border-border{border-color:var(--border)}.bg-background{background-color:var(--background)}.bg-muted-foreground,.bg-muted-foreground\\/40{background-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.bg-muted-foreground\\/40{background-color:color-mix(in oklab, var(--muted-foreground) 40%, transparent)}.bg-muted-foreground\\/60{background-color:color-mix(in oklab, var(--muted-foreground) 60%, transparent)}.bg-muted\\/30{background-color:color-mix(in oklab, var(--muted) 30%, transparent)}.bg-muted\\/50{background-color:color-mix(in oklab, var(--muted) 50%, transparent)}}.bg-primary\\/5{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/5{background-color:color-mix(in oklab, var(--primary) 5%, transparent)}}.pr-density-4{padding-right:calc(var(--spacing-unit) * 4)}.text-foreground{color:var(--foreground)}.text-muted-foreground{color:var(--muted-foreground)}.no-underline{text-decoration-line:none}.outline{outline-style:var(--tw-outline-style);outline-width:1px}@media (hover:hover){.hover\\:underline:hover{text-decoration-line:underline}}@media (width>=40rem){.sm\\:grid-cols-5{grid-template-columns:repeat(5,minmax(0,1fr))}}@media (width>=64rem){.lg\\:inline{display:inline}}.dark\\:bg-red-500\\/10:where([data-theme=dark],[data-theme=dark] *){background-color:#fb2c361a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-red-500\\/10:where([data-theme=dark],[data-theme=dark] *){background-color:color-mix(in oklab, var(--color-red-500) 10%, transparent)}}.dark\\:text-green-400:where([data-theme=dark],[data-theme=dark] *){color:var(--color-green-400)}.dark\\:text-red-300:where([data-theme=dark],[data-theme=dark] *){color:var(--color-red-300)}.dark\\:text-red-400:where([data-theme=dark],[data-theme=dark] *){color:var(--color-red-400)}.\\@container\\/app-content{container:app-content/inline-size}.\\@container\\/app-shell{container:app-shell/inline-size}.\\@container\\/combobox-options{container:combobox-options/inline-size}.\\@container\\/endpoint{container:endpoint/inline-size}.\\@container\\/secret{container:secret/inline-size}.\\@container\\/session-viewer{container:session-viewer/inline-size}.\\@container\\/split-pane{container:split-pane/inline-size}.\\@container\\/workspace{container:workspace/inline-size}.\\@container{container-type:inline-size}.\\[container-type\\:size\\]{container-type:size}.pointer-events-auto{pointer-events:auto}.pointer-events-none{pointer-events:none}.collapse{visibility:collapse}.invisible{visibility:hidden}.visible{visibility:visible}.sr-only{clip-path:inset(50%);white-space:nowrap;border-width:0;width:1px;height:1px;margin:-1px;padding:0;position:absolute;overflow:hidden}.absolute{position:absolute}.fixed{position:fixed}.relative{position:relative}.static{position:static}.sticky{position:sticky}.inset-0{inset:0}.inset-4{inset:calc(var(--spacing) * 4)}.inset-x-0{inset-inline:0}.inset-x-3{inset-inline:calc(var(--spacing) * 3)}.inset-y-0{inset-block:0}.-top-0\\.5{top:calc(var(--spacing) * -.5)}.-top-1{top:calc(var(--spacing) * -1)}.top-0{top:0}.top-0\\.5{top:calc(var(--spacing) * .5)}.top-1{top:var(--spacing)}.top-1\\/2{top:50%}.top-2{top:calc(var(--spacing) * 2)}.top-3{top:calc(var(--spacing) * 3)}.top-9{top:calc(var(--spacing) * 9)}.top-10{top:calc(var(--spacing) * 10)}.top-\\[22px\\]{top:22px}.top-\\[calc\\(100\\%\\+0\\.375rem\\)\\]{top:calc(100% + .375rem)}.-right-0\\.5{right:calc(var(--spacing) * -.5)}.-right-1{right:calc(var(--spacing) * -1)}.right-0{right:0}.right-1{right:var(--spacing)}.right-1\\.5{right:calc(var(--spacing) * 1.5)}.right-2{right:calc(var(--spacing) * 2)}.right-3{right:calc(var(--spacing) * 3)}.right-4{right:calc(var(--spacing) * 4)}.right-6{right:calc(var(--spacing) * 6)}.right-7{right:calc(var(--spacing) * 7)}.right-\\[3\\.75rem\\]{right:3.75rem}.right-auto{right:auto}.-bottom-0\\.5{bottom:calc(var(--spacing) * -.5)}.-bottom-1{bottom:calc(var(--spacing) * -1)}.bottom-0{bottom:0}.bottom-1{bottom:var(--spacing)}.bottom-2{bottom:calc(var(--spacing) * 2)}.bottom-2\\.5{bottom:calc(var(--spacing) * 2.5)}.bottom-3{bottom:calc(var(--spacing) * 3)}.bottom-6{bottom:calc(var(--spacing) * 6)}.bottom-\\[calc\\(1rem\\+env\\(safe-area-inset-bottom\\)\\)\\]{bottom:calc(1rem + env(safe-area-inset-bottom))}.bottom-density-3{bottom:var(--spacing-density-3)}.-left-0\\.5{left:calc(var(--spacing) * -.5)}.left-0{left:0}.left-0\\.5{left:calc(var(--spacing) * .5)}.left-1{left:var(--spacing)}.left-1\\.5{left:calc(var(--spacing) * 1.5)}.left-1\\/2{left:50%}.left-2{left:calc(var(--spacing) * 2)}.left-\\[10px\\]{left:10px}.left-auto{left:auto}.left-density-3{left:var(--spacing-density-3)}.isolate{isolation:isolate}.z-0{z-index:0}.z-10{z-index:10}.z-20{z-index:20}.z-30{z-index:30}.z-50{z-index:50}.z-\\[1\\]{z-index:1}.z-\\[30\\]{z-index:30}.z-\\[999\\]{z-index:999}.z-\\[9999\\]{z-index:9999}.z-\\[N\\]{z-index:N}.z-\\[…\\]{z-index:…}.order-1{order:1}.order-2{order:2}.order-3{order:3}.order-4{order:4}.col-span-2{grid-column:span 2/span 2}.col-span-3{grid-column:span 3/span 3}.col-span-full{grid-column:1/-1}.col-start-1{grid-column-start:1}.col-start-2{grid-column-start:2}.row-0{grid-row:0}.row-1{grid-row:1}.row-7{grid-row:7}.row-start-1{grid-row-start:1}.\\!container{width:100%!important}@media (width>=40rem){.\\!container{max-width:40rem!important}}@media (width>=48rem){.\\!container{max-width:48rem!important}}@media (width>=64rem){.\\!container{max-width:64rem!important}}@media (width>=80rem){.\\!container{max-width:80rem!important}}@media (width>=96rem){.\\!container{max-width:96rem!important}}.container{width:100%}@media (width>=40rem){.container{max-width:40rem}}@media (width>=48rem){.container{max-width:48rem}}@media (width>=64rem){.container{max-width:64rem}}@media (width>=80rem){.container{max-width:80rem}}@media (width>=96rem){.container{max-width:96rem}}.m-0{margin:0}.m-1{margin:var(--spacing)}.m-2{margin:calc(var(--spacing) * 2)}.m-4{margin:calc(var(--spacing) * 4)}.m-density-2{margin:var(--spacing-density-2)}.m-density-4{margin:var(--spacing-density-4)}.-mx-1{margin-inline:calc(var(--spacing) * -1)}.-mx-2{margin-inline:calc(var(--spacing) * -2)}.-mx-density-4{margin-inline:calc(var(--spacing-density-4) * -1)}.mx-0\\.5{margin-inline:calc(var(--spacing) * .5)}.mx-1{margin-inline:var(--spacing)}.mx-2{margin-inline:calc(var(--spacing) * 2)}.mx-3{margin-inline:calc(var(--spacing) * 3)}.mx-auto{margin-inline:auto}.mx-density-1{margin-inline:var(--spacing-density-1)}.mx-density-4{margin-inline:var(--spacing-density-4)}.-my-density-3{margin-block:calc(var(--spacing-density-3) * -1)}.my-0{margin-block:0}.my-1{margin-block:var(--spacing)}.my-2{margin-block:calc(var(--spacing) * 2)}.my-3{margin-block:calc(var(--spacing) * 3)}.my-4{margin-block:calc(var(--spacing) * 4)}.ms-1{margin-inline-start:var(--spacing)}.-mt-0\\.5{margin-top:calc(var(--spacing) * -.5)}.-mt-1{margin-top:calc(var(--spacing) * -1)}.-mt-density-4{margin-top:calc(var(--spacing-density-4) * -1)}.mt-0{margin-top:0}.mt-0\\.5{margin-top:calc(var(--spacing) * .5)}.mt-1{margin-top:var(--spacing)}.mt-1\\.5{margin-top:calc(var(--spacing) * 1.5)}.mt-2{margin-top:calc(var(--spacing) * 2)}.mt-3{margin-top:calc(var(--spacing) * 3)}.mt-4{margin-top:calc(var(--spacing) * 4)}.mt-5{margin-top:calc(var(--spacing) * 5)}.mt-auto{margin-top:auto}.mt-density-1{margin-top:var(--spacing-density-1)}.mt-density-2{margin-top:var(--spacing-density-2)}.mt-density-3{margin-top:var(--spacing-density-3)}.mt-density-4{margin-top:var(--spacing-density-4)}.mt-density-6{margin-top:var(--spacing-density-6)}.mt-px{margin-top:1px}.-mr-1{margin-right:calc(var(--spacing) * -1)}.-mr-px{margin-right:-1px}.mr-1{margin-right:var(--spacing)}.mr-1\\.5{margin-right:calc(var(--spacing) * 1.5)}.mr-2{margin-right:calc(var(--spacing) * 2)}.mr-auto{margin-right:auto}.-mb-px{margin-bottom:-1px}.mb-0\\.5{margin-bottom:calc(var(--spacing) * .5)}.mb-1{margin-bottom:var(--spacing)}.mb-1\\.5{margin-bottom:calc(var(--spacing) * 1.5)}.mb-2{margin-bottom:calc(var(--spacing) * 2)}.mb-3{margin-bottom:calc(var(--spacing) * 3)}.mb-4{margin-bottom:calc(var(--spacing) * 4)}.mb-5{margin-bottom:calc(var(--spacing) * 5)}.mb-density-1{margin-bottom:var(--spacing-density-1)}.mb-density-2{margin-bottom:var(--spacing-density-2)}.mb-density-3{margin-bottom:var(--spacing-density-3)}.mb-density-4{margin-bottom:var(--spacing-density-4)}.mb-density-6{margin-bottom:var(--spacing-density-6)}.-ml-1\\.5{margin-left:calc(var(--spacing) * -1.5)}.-ml-2{margin-left:calc(var(--spacing) * -2)}.-ml-px{margin-left:-1px}.ml-0\\.5{margin-left:calc(var(--spacing) * .5)}.ml-1{margin-left:var(--spacing)}.ml-1\\.5{margin-left:calc(var(--spacing) * 1.5)}.ml-2{margin-left:calc(var(--spacing) * 2)}.ml-3{margin-left:calc(var(--spacing) * 3)}.ml-3\\.5{margin-left:calc(var(--spacing) * 3.5)}.ml-4{margin-left:calc(var(--spacing) * 4)}.ml-5{margin-left:calc(var(--spacing) * 5)}.ml-6{margin-left:calc(var(--spacing) * 6)}.ml-12{margin-left:calc(var(--spacing) * 12)}.ml-\\[5px\\]{margin-left:5px}.ml-auto{margin-left:auto}.ml-density-2{margin-left:var(--spacing-density-2)}.line-clamp-1{-webkit-line-clamp:1;-webkit-box-orient:vertical;display:-webkit-box;overflow:hidden}.line-clamp-2{-webkit-line-clamp:2;-webkit-box-orient:vertical;display:-webkit-box;overflow:hidden}.line-clamp-4{-webkit-line-clamp:4;-webkit-box-orient:vertical;display:-webkit-box;overflow:hidden}.\\[display\\:-webkit-box\\]{display:-webkit-box}.block{display:block}.contents{display:contents}.flex{display:flex}.grid{display:grid}.hidden{display:none}.inline{display:inline}.inline-block{display:inline-block}.inline-flex{display:inline-flex}.list-item{display:list-item}.table{display:table}.aspect-square{aspect-ratio:1}.size-1\\.5{width:calc(var(--spacing) * 1.5);height:calc(var(--spacing) * 1.5)}.size-2{width:calc(var(--spacing) * 2);height:calc(var(--spacing) * 2)}.size-3{width:calc(var(--spacing) * 3);height:calc(var(--spacing) * 3)}.size-3\\.5{width:calc(var(--spacing) * 3.5);height:calc(var(--spacing) * 3.5)}.size-4{width:calc(var(--spacing) * 4);height:calc(var(--spacing) * 4)}.size-5{width:calc(var(--spacing) * 5);height:calc(var(--spacing) * 5)}.size-6{width:calc(var(--spacing) * 6);height:calc(var(--spacing) * 6)}.size-7{width:calc(var(--spacing) * 7);height:calc(var(--spacing) * 7)}.size-8{width:calc(var(--spacing) * 8);height:calc(var(--spacing) * 8)}.size-9{width:calc(var(--spacing) * 9);height:calc(var(--spacing) * 9)}.size-10{width:calc(var(--spacing) * 10);height:calc(var(--spacing) * 10)}.size-12{width:calc(var(--spacing) * 12);height:calc(var(--spacing) * 12)}.size-\\[clamp\\(…\\)\\]{width:clamp(…);height:clamp(…)}.size-control-h{width:var(--spacing-control-h);height:var(--spacing-control-h)}.size-full{width:100%;height:100%}.h-0{height:0}.h-0\\.5{height:calc(var(--spacing) * .5)}.h-1{height:var(--spacing)}.h-1\\.5{height:calc(var(--spacing) * 1.5)}.h-2{height:calc(var(--spacing) * 2)}.h-2\\.5{height:calc(var(--spacing) * 2.5)}.h-3{height:calc(var(--spacing) * 3)}.h-3\\.5{height:calc(var(--spacing) * 3.5)}.h-4{height:calc(var(--spacing) * 4)}.h-5{height:calc(var(--spacing) * 5)}.h-6{height:calc(var(--spacing) * 6)}.h-7{height:calc(var(--spacing) * 7)}.h-8{height:calc(var(--spacing) * 8)}.h-9{height:calc(var(--spacing) * 9)}.h-10{height:calc(var(--spacing) * 10)}.h-11{height:calc(var(--spacing) * 11)}.h-12{height:calc(var(--spacing) * 12)}.h-14{height:calc(var(--spacing) * 14)}.h-16{height:calc(var(--spacing) * 16)}.h-24{height:calc(var(--spacing) * 24)}.h-28{height:calc(var(--spacing) * 28)}.h-40{height:calc(var(--spacing) * 40)}.h-64{height:calc(var(--spacing) * 64)}.h-72{height:calc(var(--spacing) * 72)}.h-80{height:calc(var(--spacing) * 80)}.h-96{height:calc(var(--spacing) * 96)}.h-\\[3px\\]{height:3px}.h-\\[16px\\]{height:16px}.h-\\[18px\\]{height:18px}.h-\\[21px\\]{height:21px}.h-\\[30rem\\]{height:30rem}.h-\\[32rem\\]{height:32rem}.h-\\[33px\\]{height:33px}.h-\\[34\\%\\]{height:34%}.h-\\[34px\\]{height:34px}.h-\\[36rem\\]{height:36rem}.h-\\[42rem\\]{height:42rem}.h-\\[60vh\\]{height:60vh}.h-\\[70vh\\]{height:70vh}.h-\\[75vh\\]{height:75vh}.h-\\[80vh\\]{height:80vh}.h-\\[92vh\\]{height:92vh}.h-\\[95vh\\]{height:95vh}.h-\\[240px\\]{height:240px}.h-\\[280px\\]{height:280px}.h-\\[320px\\]{height:320px}.h-\\[360px\\]{height:360px}.h-\\[400px\\]{height:400px}.h-\\[420px\\]{height:420px}.h-\\[440px\\]{height:440px}.h-\\[460px\\]{height:460px}.h-\\[480px\\]{height:480px}.h-\\[500px\\]{height:500px}.h-\\[520px\\]{height:520px}.h-\\[560px\\]{height:560px}.h-\\[600px\\]{height:600px}.h-\\[640px\\]{height:640px}.h-\\[720px\\]{height:720px}.h-\\[calc\\(100dvh-2rem\\)\\]{height:calc(100dvh - 2rem)}.h-\\[calc\\(100vh-15rem\\)\\]{height:calc(100vh - 15rem)}.h-\\[min\\(32rem\\,52vh\\)\\]{height:min(32rem,52vh)}.h-\\[min\\(58vh\\,34rem\\)\\]{height:min(58vh,34rem)}.h-auto{height:auto}.h-control-h{height:var(--spacing-control-h)}.h-density-4{height:var(--spacing-density-4)}.h-dvh{height:100dvh}.h-fit{height:fit-content}.h-full{height:100%}.h-px{height:1px}.h-screen{height:100vh}.max-h-20{max-height:calc(var(--spacing) * 20)}.max-h-24{max-height:calc(var(--spacing) * 24)}.max-h-32{max-height:calc(var(--spacing) * 32)}.max-h-40{max-height:calc(var(--spacing) * 40)}.max-h-48{max-height:calc(var(--spacing) * 48)}.max-h-56{max-height:calc(var(--spacing) * 56)}.max-h-64{max-height:calc(var(--spacing) * 64)}.max-h-72{max-height:calc(var(--spacing) * 72)}.max-h-80{max-height:calc(var(--spacing) * 80)}.max-h-96{max-height:calc(var(--spacing) * 96)}.max-h-\\[26rem\\]{max-height:26rem}.max-h-\\[32rem\\]{max-height:32rem}.max-h-\\[40\\%\\]{max-height:40%}.max-h-\\[40vh\\]{max-height:40vh}.max-h-\\[46rem\\]{max-height:46rem}.max-h-\\[70vh\\]{max-height:70vh}.max-h-\\[360px\\]{max-height:360px}.max-h-\\[520px\\]{max-height:520px}.max-h-\\[calc\\(100vh-1rem\\)\\]{max-height:calc(100vh - 1rem)}.max-h-\\[min\\(32rem\\,calc\\(100vh-2rem\\)\\)\\]{max-height:min(32rem,100vh - 2rem)}.max-h-\\[min\\(34rem\\,calc\\(100vh-2rem\\)\\)\\]{max-height:min(34rem,100vh - 2rem)}.max-h-\\[min\\(42rem\\,calc\\(100vh-5rem\\)\\)\\]{max-height:min(42rem,100vh - 5rem)}.min-h-0{min-height:0}.min-h-4{min-height:calc(var(--spacing) * 4)}.min-h-5{min-height:calc(var(--spacing) * 5)}.min-h-6{min-height:calc(var(--spacing) * 6)}.min-h-7{min-height:calc(var(--spacing) * 7)}.min-h-8{min-height:calc(var(--spacing) * 8)}.min-h-9{min-height:calc(var(--spacing) * 9)}.min-h-10{min-height:calc(var(--spacing) * 10)}.min-h-11{min-height:calc(var(--spacing) * 11)}.min-h-12{min-height:calc(var(--spacing) * 12)}.min-h-14{min-height:calc(var(--spacing) * 14)}.min-h-16{min-height:calc(var(--spacing) * 16)}.min-h-20{min-height:calc(var(--spacing) * 20)}.min-h-28{min-height:calc(var(--spacing) * 28)}.min-h-32{min-height:calc(var(--spacing) * 32)}.min-h-36{min-height:calc(var(--spacing) * 36)}.min-h-40{min-height:calc(var(--spacing) * 40)}.min-h-44{min-height:calc(var(--spacing) * 44)}.min-h-48{min-height:calc(var(--spacing) * 48)}.min-h-56{min-height:calc(var(--spacing) * 56)}.min-h-64{min-height:calc(var(--spacing) * 64)}.min-h-72{min-height:calc(var(--spacing) * 72)}.min-h-80{min-height:calc(var(--spacing) * 80)}.min-h-96{min-height:calc(var(--spacing) * 96)}.min-h-\\[4\\.25rem\\]{min-height:4.25rem}.min-h-\\[5rem\\]{min-height:5rem}.min-h-\\[7rem\\]{min-height:7rem}.min-h-\\[8rem\\]{min-height:8rem}.min-h-\\[12rem\\]{min-height:12rem}.min-h-\\[14rem\\]{min-height:14rem}.min-h-\\[18px\\]{min-height:18px}.min-h-\\[20rem\\]{min-height:20rem}.min-h-\\[24rem\\]{min-height:24rem}.min-h-\\[32rem\\]{min-height:32rem}.min-h-\\[34rem\\]{min-height:34rem}.min-h-\\[48px\\]{min-height:48px}.min-h-\\[640px\\]{min-height:640px}.min-h-control-h{min-height:var(--spacing-control-h)}.min-h-dvh{min-height:100dvh}.min-h-px{min-height:1px}.min-h-screen{min-height:100vh}.w-0{width:0}.w-0\\.5{width:calc(var(--spacing) * .5)}.w-1{width:var(--spacing)}.w-1\\.5{width:calc(var(--spacing) * 1.5)}.w-1\\/2{width:50%}.w-1\\/3{width:33.3333%}.w-2{width:calc(var(--spacing) * 2)}.w-2\\.5{width:calc(var(--spacing) * 2.5)}.w-2\\/3{width:66.6667%}.w-3{width:calc(var(--spacing) * 3)}.w-3\\.5{width:calc(var(--spacing) * 3.5)}.w-4{width:calc(var(--spacing) * 4)}.w-5{width:calc(var(--spacing) * 5)}.w-5\\/6{width:83.3333%}.w-6{width:calc(var(--spacing) * 6)}.w-7{width:calc(var(--spacing) * 7)}.w-8{width:calc(var(--spacing) * 8)}.w-9{width:calc(var(--spacing) * 9)}.w-10{width:calc(var(--spacing) * 10)}.w-12{width:calc(var(--spacing) * 12)}.w-14{width:calc(var(--spacing) * 14)}.w-16{width:calc(var(--spacing) * 16)}.w-20{width:calc(var(--spacing) * 20)}.w-24{width:calc(var(--spacing) * 24)}.w-28{width:calc(var(--spacing) * 28)}.w-32{width:calc(var(--spacing) * 32)}.w-36{width:calc(var(--spacing) * 36)}.w-40{width:calc(var(--spacing) * 40)}.w-44{width:calc(var(--spacing) * 44)}.w-48{width:calc(var(--spacing) * 48)}.w-52{width:calc(var(--spacing) * 52)}.w-56{width:calc(var(--spacing) * 56)}.w-64{width:calc(var(--spacing) * 64)}.w-72{width:calc(var(--spacing) * 72)}.w-80{width:calc(var(--spacing) * 80)}.w-96{width:calc(var(--spacing) * 96)}.w-\\[1\\.25em\\]{width:1.25em}.w-\\[3px\\]{width:3px}.w-\\[5\\.5rem\\]{width:5.5rem}.w-\\[5\\.25rem\\]{width:5.25rem}.w-\\[15\\.5rem\\]{width:15.5rem}.w-\\[18\\%\\]{width:18%}.w-\\[21px\\]{width:21px}.w-\\[22\\.5rem\\]{width:22.5rem}.w-\\[22rem\\]{width:22rem}.w-\\[24rem\\]{width:24rem}.w-\\[28\\%\\]{width:28%}.w-\\[28rem\\]{width:28rem}.w-\\[30rem\\]{width:30rem}.w-\\[34px\\]{width:34px}.w-\\[34rem\\]{width:34rem}.w-\\[42rem\\]{width:42rem}.w-\\[45rem\\]{width:45rem}.w-\\[52px\\]{width:52px}.w-\\[52rem\\]{width:52rem}.w-\\[58rem\\]{width:58rem}.w-\\[64rem\\]{width:64rem}.w-\\[200px\\]{width:200px}.w-\\[316px\\]{width:316px}.w-\\[320px\\]{width:320px}.w-\\[360px\\]{width:360px}.w-\\[390px\\]{width:390px}.w-\\[420px\\]{width:420px}.w-\\[480px\\]{width:480px}.w-\\[560px\\]{width:560px}.w-\\[640px\\]{width:640px}.w-\\[720px\\]{width:720px}.w-\\[760px\\]{width:760px}.w-\\[880px\\]{width:880px}.w-\\[calc\\(var\\(--dt-viewport-w\\)-2\\*var\\(--spacing-density-3\\)\\)\\]{width:calc(var(--dt-viewport-w) - 2 * var(--spacing-density-3))}.w-\\[min\\(20rem\\,85vw\\)\\]{width:min(20rem,85vw)}.w-\\[min\\(22rem\\,calc\\(100vw-2rem\\)\\)\\]{width:min(22rem,100vw - 2rem)}.w-\\[min\\(30rem\\,calc\\(100vw-2rem\\)\\)\\]{width:min(30rem,100vw - 2rem)}.w-\\[min\\(34rem\\,calc\\(100vw-2rem\\)\\)\\]{width:min(34rem,100vw - 2rem)}.w-\\[min\\(38rem\\,calc\\(100vw-4rem\\)\\)\\]{width:min(38rem,100vw - 4rem)}.w-\\[min\\(42rem\\,calc\\(100vw-1rem\\)\\)\\]{width:min(42rem,100vw - 1rem)}.w-\\[min\\(42rem\\,calc\\(100vw-2rem\\)\\)\\]{width:min(42rem,100vw - 2rem)}.w-\\[min\\(100\\%\\,28rem\\)\\]{width:min(100%,28rem)}.w-auto{width:auto}.w-control-h{width:var(--spacing-control-h)}.w-fit{width:fit-content}.w-full{width:100%}.w-max{width:max-content}.w-px{width:1px}.w-screen{width:100vw}.max-w-2xl{max-width:var(--container-2xl)}.max-w-3xl{max-width:var(--container-3xl)}.max-w-4{max-width:calc(var(--spacing) * 4)}.max-w-4xl{max-width:var(--container-4xl)}.max-w-5xl{max-width:var(--container-5xl)}.max-w-6xl{max-width:var(--container-6xl)}.max-w-7xl{max-width:var(--container-7xl)}.max-w-32{max-width:calc(var(--spacing) * 32)}.max-w-40{max-width:calc(var(--spacing) * 40)}.max-w-48{max-width:calc(var(--spacing) * 48)}.max-w-52{max-width:calc(var(--spacing) * 52)}.max-w-56{max-width:calc(var(--spacing) * 56)}.max-w-64{max-width:calc(var(--spacing) * 64)}.max-w-72{max-width:calc(var(--spacing) * 72)}.max-w-80{max-width:calc(var(--spacing) * 80)}.max-w-\\[8\\.5rem\\]{max-width:8.5rem}.max-w-\\[9\\.5rem\\]{max-width:9.5rem}.max-w-\\[12rem\\]{max-width:12rem}.max-w-\\[15rem\\]{max-width:15rem}.max-w-\\[16rem\\]{max-width:16rem}.max-w-\\[18rem\\]{max-width:18rem}.max-w-\\[20rem\\]{max-width:20rem}.max-w-\\[22rem\\]{max-width:22rem}.max-w-\\[24rem\\]{max-width:24rem}.max-w-\\[25\\%\\]{max-width:25%}.max-w-\\[28rem\\]{max-width:28rem}.max-w-\\[30\\%\\]{max-width:30%}.max-w-\\[32rem\\]{max-width:32rem}.max-w-\\[40\\%\\]{max-width:40%}.max-w-\\[40rem\\]{max-width:40rem}.max-w-\\[45\\%\\]{max-width:45%}.max-w-\\[48rem\\]{max-width:48rem}.max-w-\\[55\\%\\]{max-width:55%}.max-w-\\[90vw\\]{max-width:90vw}.max-w-\\[95\\%\\]{max-width:95%}.max-w-\\[95vw\\]{max-width:95vw}.max-w-\\[100px\\]{max-width:100px}.max-w-\\[180px\\]{max-width:180px}.max-w-\\[200px\\]{max-width:200px}.max-w-\\[320px\\]{max-width:320px}.max-w-\\[820px\\]{max-width:820px}.max-w-\\[calc\\(100\\%-1rem\\)\\]{max-width:calc(100% - 1rem)}.max-w-\\[calc\\(100vw-1rem\\)\\]{max-width:calc(100vw - 1rem)}.max-w-\\[calc\\(100vw-2rem\\)\\]{max-width:calc(100vw - 2rem)}.max-w-\\[min\\(40ch\\,100\\%\\)\\]{max-width:min(40ch,100%)}.max-w-\\[min\\(50ch\\,100\\%\\)\\]{max-width:min(50ch,100%)}.max-w-full{max-width:100%}.max-w-lg{max-width:var(--container-lg)}.max-w-md{max-width:var(--container-md)}.max-w-none{max-width:none}.max-w-sm{max-width:var(--container-sm)}.max-w-xl{max-width:var(--container-xl)}.max-w-xs{max-width:var(--container-xs)}.min-w-0{min-width:0}.min-w-4{min-width:calc(var(--spacing) * 4)}.min-w-8{min-width:calc(var(--spacing) * 8)}.min-w-9{min-width:calc(var(--spacing) * 9)}.min-w-10{min-width:calc(var(--spacing) * 10)}.min-w-12{min-width:calc(var(--spacing) * 12)}.min-w-20{min-width:calc(var(--spacing) * 20)}.min-w-24{min-width:calc(var(--spacing) * 24)}.min-w-28{min-width:calc(var(--spacing) * 28)}.min-w-32{min-width:calc(var(--spacing) * 32)}.min-w-36{min-width:calc(var(--spacing) * 36)}.min-w-40{min-width:calc(var(--spacing) * 40)}.min-w-44{min-width:calc(var(--spacing) * 44)}.min-w-48{min-width:calc(var(--spacing) * 48)}.min-w-56{min-width:calc(var(--spacing) * 56)}.min-w-64{min-width:calc(var(--spacing) * 64)}.min-w-72{min-width:calc(var(--spacing) * 72)}.min-w-\\[3px\\]{min-width:3px}.min-w-\\[8rem\\]{min-width:8rem}.min-w-\\[9rem\\]{min-width:9rem}.min-w-\\[11rem\\]{min-width:11rem}.min-w-\\[12rem\\]{min-width:12rem}.min-w-\\[14rem\\]{min-width:14rem}.min-w-\\[16px\\]{min-width:16px}.min-w-\\[16rem\\]{min-width:16rem}.min-w-\\[18px\\]{min-width:18px}.min-w-\\[18rem\\]{min-width:18rem}.min-w-\\[80px\\]{min-width:80px}.min-w-\\[150px\\]{min-width:150px}.min-w-\\[170px\\]{min-width:170px}.min-w-\\[220px\\]{min-width:220px}.min-w-\\[520px\\]{min-width:520px}.min-w-\\[560px\\]{min-width:560px}.min-w-full{min-width:100%}.min-w-max{min-width:max-content}.flex-1{flex:1}.flex-\\[1_1_14rem\\]{flex:14rem}.flex-none{flex:none}.shrink{flex-shrink:1}.shrink-0{flex-shrink:0}.shrink-\\[100\\]{flex-shrink:100}.grow{flex-grow:1}.grow-0{flex-grow:0}.basis-0{flex-basis:0}.basis-28{flex-basis:calc(var(--spacing) * 28)}.basis-32{flex-basis:calc(var(--spacing) * 32)}.basis-40{flex-basis:calc(var(--spacing) * 40)}.basis-64{flex-basis:calc(var(--spacing) * 64)}.basis-auto{flex-basis:auto}.basis-full{flex-basis:100%}.table-auto{table-layout:auto}.table-fixed{table-layout:fixed}.border-collapse{border-collapse:collapse}.border-separate{border-collapse:separate}.border-spacing-y-density-2{--tw-border-spacing-y:var(--spacing-density-2);border-spacing:var(--tw-border-spacing-x) var(--tw-border-spacing-y)}.-translate-x-1\\/2{--tw-translate-x:calc(calc(1 / 2 * 100%) * -1);translate:var(--tw-translate-x) var(--tw-translate-y)}.-translate-x-2{--tw-translate-x:calc(var(--spacing) * -2);translate:var(--tw-translate-x) var(--tw-translate-y)}.-translate-x-\\[calc\\(100\\%\\+0\\.375rem\\)\\]{--tw-translate-x:calc(calc(100% + .375rem) * -1);translate:var(--tw-translate-x) var(--tw-translate-y)}.translate-x-0{--tw-translate-x:0px;translate:var(--tw-translate-x) var(--tw-translate-y)}.translate-x-1\\.5{--tw-translate-x:calc(var(--spacing) * 1.5);translate:var(--tw-translate-x) var(--tw-translate-y)}.translate-x-4{--tw-translate-x:calc(var(--spacing) * 4);translate:var(--tw-translate-x) var(--tw-translate-y)}.translate-x-8{--tw-translate-x:calc(var(--spacing) * 8);translate:var(--tw-translate-x) var(--tw-translate-y)}.-translate-y-1\\/2{--tw-translate-y:calc(calc(1 / 2 * 100%) * -1);translate:var(--tw-translate-x) var(--tw-translate-y)}.translate-y-1\\/2{--tw-translate-y:calc(1 / 2 * 100%);translate:var(--tw-translate-x) var(--tw-translate-y)}.-scale-x-100{--tw-scale-x:calc(100% * -1);scale:var(--tw-scale-x) var(--tw-scale-y)}.-rotate-90{rotate:-90deg}.rotate-45{rotate:45deg}.rotate-90{rotate:90deg}.rotate-180{rotate:180deg}.transform{transform:var(--tw-rotate-x,) var(--tw-rotate-y,) var(--tw-rotate-z,) var(--tw-skew-x,) var(--tw-skew-y,)}.animate-pulse{animation:var(--animate-pulse)}.animate-spin{animation:var(--animate-spin)}.cursor-col-resize{cursor:col-resize}.cursor-default{cursor:default}.cursor-grab{cursor:grab}.cursor-grabbing{cursor:grabbing}.cursor-help{cursor:help}.cursor-move{cursor:move}.cursor-not-allowed{cursor:not-allowed}.cursor-pointer{cursor:pointer}.cursor-row-resize{cursor:row-resize}.cursor-text{cursor:text}.touch-none{touch-action:none}.resize{resize:both}.resize-none{resize:none}.resize-y{resize:vertical}.scroll-mt-4{scroll-margin-top:calc(var(--spacing) * 4)}.scroll-mt-24{scroll-margin-top:calc(var(--spacing) * 24)}.list-decimal{list-style-type:decimal}.list-disc{list-style-type:disc}.list-none{list-style-type:none}.appearance-none{appearance:none}.grid-cols-1{grid-template-columns:repeat(1,minmax(0,1fr))}.grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}.grid-cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}.grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}.grid-cols-\\[1\\.25rem_2\\.5rem_1fr\\]{grid-template-columns:1.25rem 2.5rem 1fr}.grid-cols-\\[1\\.25rem_minmax\\(0\\,1fr\\)\\]{grid-template-columns:1.25rem minmax(0,1fr)}.grid-cols-\\[1\\.25rem_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:1.25rem minmax(0,1fr) auto}.grid-cols-\\[1fr_auto\\]{grid-template-columns:1fr auto}.grid-cols-\\[1rem_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:1rem minmax(0,1fr) auto}.grid-cols-\\[2\\.5rem_1fr\\]{grid-template-columns:2.5rem 1fr}.grid-cols-\\[2\\.5rem_3\\.5rem_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:2.5rem 3.5rem minmax(0,1fr) auto}.grid-cols-\\[2rem_1fr\\]{grid-template-columns:2rem 1fr}.grid-cols-\\[2rem_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:2rem minmax(0,1fr) auto}.grid-cols-\\[3\\.5rem_minmax\\(24rem\\,1fr\\)\\]{grid-template-columns:3.5rem minmax(24rem,1fr)}.grid-cols-\\[5\\.5rem_minmax\\(0\\,1fr\\)\\]{grid-template-columns:5.5rem minmax(0,1fr)}.grid-cols-\\[7rem_1fr\\]{grid-template-columns:7rem 1fr}.grid-cols-\\[10rem_1fr\\]{grid-template-columns:10rem 1fr}.grid-cols-\\[10rem_1fr_9rem\\]{grid-template-columns:10rem 1fr 9rem}.grid-cols-\\[10rem_1fr_auto\\]{grid-template-columns:10rem 1fr auto}.grid-cols-\\[13rem_1\\.25rem_2\\.5rem_1fr\\]{grid-template-columns:13rem 1.25rem 2.5rem 1fr}.grid-cols-\\[18rem_24rem_30rem\\]{grid-template-columns:18rem 24rem 30rem}.grid-cols-\\[auto_1fr\\]{grid-template-columns:auto 1fr}.grid-cols-\\[auto_auto_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:auto auto minmax(0,1fr) auto}.grid-cols-\\[auto_minmax\\(0\\,1fr\\)\\]{grid-template-columns:auto minmax(0,1fr)}.grid-cols-\\[auto_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:auto minmax(0,1fr) auto}.grid-cols-\\[max-content_minmax\\(0\\,1fr\\)\\]{grid-template-columns:max-content minmax(0,1fr)}.grid-cols-\\[minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:minmax(0,1fr) auto}.grid-cols-\\[minmax\\(0\\,1fr\\)_auto_auto\\]{grid-template-columns:minmax(0,1fr) auto auto}.grid-cols-\\[minmax\\(0\\,10rem\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(0,10rem) minmax(0,1fr)}.grid-cols-\\[minmax\\(0\\,14rem\\)_minmax\\(0\\,1fr\\)_4rem\\]{grid-template-columns:minmax(0,14rem) minmax(0,1fr) 4rem}.grid-cols-\\[minmax\\(7rem\\,12rem\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(7rem,12rem) minmax(0,1fr)}.grid-cols-\\[repeat\\(auto-fill\\,minmax\\(12rem\\,1fr\\)\\)\\]{grid-template-columns:repeat(auto-fill,minmax(12rem,1fr))}.grid-cols-\\[repeat\\(auto-fill\\,minmax\\(13rem\\,1fr\\)\\)\\]{grid-template-columns:repeat(auto-fill,minmax(13rem,1fr))}.grid-cols-\\[repeat\\(auto-fill\\,minmax\\(200px\\,1fr\\)\\)\\]{grid-template-columns:repeat(auto-fill,minmax(200px,1fr))}.grid-cols-\\[repeat\\(auto-fill\\,…\\)\\]{grid-template-columns:repeat(auto-fill,…)}.grid-cols-subgrid{grid-template-columns:subgrid}.grid-rows-\\[auto_auto_auto\\]{grid-template-rows:auto auto auto}.grid-rows-\\[auto_minmax\\(0\\,1fr\\)\\]{grid-template-rows:auto minmax(0,1fr)}.flex-col{flex-direction:column}.flex-nowrap{flex-wrap:nowrap}.flex-wrap{flex-wrap:wrap}.place-items-center{place-items:center}.content-start{align-content:flex-start}.items-baseline{align-items:baseline}.items-center{align-items:center}.items-end{align-items:flex-end}.items-start{align-items:flex-start}.items-stretch{align-items:stretch}.justify-between{justify-content:space-between}.justify-center{justify-content:center}.justify-end{justify-content:flex-end}.justify-start{justify-content:flex-start}.justify-stretch{justify-content:stretch}.gap-0{gap:0}.gap-0\\.5{gap:calc(var(--spacing) * .5)}.gap-1{gap:var(--spacing)}.gap-1\\.5{gap:calc(var(--spacing) * 1.5)}.gap-2{gap:calc(var(--spacing) * 2)}.gap-2\\.5{gap:calc(var(--spacing) * 2.5)}.gap-3{gap:calc(var(--spacing) * 3)}.gap-4{gap:calc(var(--spacing) * 4)}.gap-5{gap:calc(var(--spacing) * 5)}.gap-6{gap:calc(var(--spacing) * 6)}.gap-8{gap:calc(var(--spacing) * 8)}.gap-density-1{gap:var(--spacing-density-1)}.gap-density-2{gap:var(--spacing-density-2)}.gap-density-3{gap:var(--spacing-density-3)}.gap-density-4{gap:var(--spacing-density-4)}.gap-px{gap:1px}:where(.space-y-0\\.5>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * .5) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * .5) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-1>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(var(--spacing) * var(--tw-space-y-reverse));margin-block-end:calc(var(--spacing) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-1\\.5>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 1.5) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 1.5) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-2>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 2) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 2) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-2\\.5>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 2.5) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 2.5) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-3>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 3) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 3) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-4>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 4) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 4) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-5>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 5) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 5) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-6>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 6) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 6) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-8>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 8) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 8) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-density-1>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(var(--spacing-density-1) * var(--tw-space-y-reverse));margin-block-end:calc(var(--spacing-density-1) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-density-2>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(var(--spacing-density-2) * var(--tw-space-y-reverse));margin-block-end:calc(var(--spacing-density-2) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-density-3>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(var(--spacing-density-3) * var(--tw-space-y-reverse));margin-block-end:calc(var(--spacing-density-3) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-density-4>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(var(--spacing-density-4) * var(--tw-space-y-reverse));margin-block-end:calc(var(--spacing-density-4) * calc(1 - var(--tw-space-y-reverse)))}.gap-x-1\\.5{column-gap:calc(var(--spacing) * 1.5)}.gap-x-2{column-gap:calc(var(--spacing) * 2)}.gap-x-3{column-gap:calc(var(--spacing) * 3)}.gap-x-4{column-gap:calc(var(--spacing) * 4)}.gap-x-5{column-gap:calc(var(--spacing) * 5)}.gap-x-6{column-gap:calc(var(--spacing) * 6)}.gap-x-density-2{column-gap:var(--spacing-density-2)}.gap-x-density-3{column-gap:var(--spacing-density-3)}.gap-y-0\\.5{row-gap:calc(var(--spacing) * .5)}.gap-y-1{row-gap:var(--spacing)}.gap-y-1\\.5{row-gap:calc(var(--spacing) * 1.5)}.gap-y-2{row-gap:calc(var(--spacing) * 2)}.gap-y-density-2{row-gap:var(--spacing-density-2)}:where(.divide-x>:not(:last-child)){--tw-divide-x-reverse:0;border-inline-style:var(--tw-border-style);border-inline-start-width:calc(1px * var(--tw-divide-x-reverse));border-inline-end-width:calc(1px * calc(1 - var(--tw-divide-x-reverse)))}:where(.divide-y>:not(:last-child)){--tw-divide-y-reverse:0;border-bottom-style:var(--tw-border-style);border-top-style:var(--tw-border-style);border-top-width:calc(1px * var(--tw-divide-y-reverse));border-bottom-width:calc(1px * calc(1 - var(--tw-divide-y-reverse)))}:where(.divide-y-0>:not(:last-child)){--tw-divide-y-reverse:0;border-bottom-style:var(--tw-border-style);border-top-style:var(--tw-border-style);border-top-width:calc(0px * var(--tw-divide-y-reverse));border-bottom-width:calc(0px * calc(1 - var(--tw-divide-y-reverse)))}:where(.divide-border>:not(:last-child)),:where(.divide-border\\/60>:not(:last-child)){border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){:where(.divide-border\\/60>:not(:last-child)){border-color:color-mix(in oklab, var(--border) 60%, transparent)}}:where(.divide-border\\/70>:not(:last-child)){border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){:where(.divide-border\\/70>:not(:last-child)){border-color:color-mix(in oklab, var(--border) 70%, transparent)}}.self-center{align-self:center}.self-start{align-self:flex-start}.self-stretch{align-self:stretch}.justify-self-start{justify-self:flex-start}.truncate{text-overflow:ellipsis;white-space:nowrap;overflow:hidden}.overflow-auto{overflow:auto}.overflow-hidden{overflow:hidden}.overflow-visible{overflow:visible}.overflow-x-auto{overflow-x:auto}.overflow-y-auto{overflow-y:auto}.overscroll-contain{overscroll-behavior:contain}.overscroll-x-contain{overscroll-behavior-x:contain}.rounded{border-radius:.25rem}.rounded-2xl{border-radius:var(--radius-2xl)}.rounded-\\[1px\\]{border-radius:1px}.rounded-\\[2px\\]{border-radius:2px}.rounded-\\[5px\\]{border-radius:5px}.rounded-full{border-radius:2147483647px}.rounded-lg{border-radius:var(--radius-lg)}.rounded-md{border-radius:var(--radius-md)}.rounded-none{border-radius:0}.rounded-sm{border-radius:var(--radius-sm)}.rounded-xl{border-radius:var(--radius-xl)}.rounded-t{border-top-left-radius:.25rem;border-top-right-radius:.25rem}.rounded-t-md{border-top-left-radius:var(--radius-md);border-top-right-radius:var(--radius-md)}.rounded-l-none{border-top-left-radius:0;border-bottom-left-radius:0}.rounded-r{border-top-right-radius:.25rem;border-bottom-right-radius:.25rem}.rounded-r-none{border-top-right-radius:0;border-bottom-right-radius:0}.rounded-tr-md{border-top-right-radius:var(--radius-md)}.border{border-style:var(--tw-border-style);border-width:1px}.border-0{border-style:var(--tw-border-style);border-width:0}.border-2{border-style:var(--tw-border-style);border-width:2px}.border-y{border-block-style:var(--tw-border-style);border-block-width:1px}.border-t{border-top-style:var(--tw-border-style);border-top-width:1px}.border-t-0{border-top-style:var(--tw-border-style);border-top-width:0}.border-t-2{border-top-style:var(--tw-border-style);border-top-width:2px}.border-r{border-right-style:var(--tw-border-style);border-right-width:1px}.border-r-0{border-right-style:var(--tw-border-style);border-right-width:0}.border-r-2{border-right-style:var(--tw-border-style);border-right-width:2px}.border-b{border-bottom-style:var(--tw-border-style);border-bottom-width:1px}.border-b-0{border-bottom-style:var(--tw-border-style);border-bottom-width:0}.border-b-2{border-bottom-style:var(--tw-border-style);border-bottom-width:2px}.border-l{border-left-style:var(--tw-border-style);border-left-width:1px}.border-l-0{border-left-style:var(--tw-border-style);border-left-width:0}.border-l-2{border-left-style:var(--tw-border-style);border-left-width:2px}.border-l-4{border-left-style:var(--tw-border-style);border-left-width:4px}.border-l-\\[3px\\]{border-left-style:var(--tw-border-style);border-left-width:3px}.border-dashed{--tw-border-style:dashed;border-style:dashed}.border-dotted{--tw-border-style:dotted;border-style:dotted}.border-none{--tw-border-style:none;border-style:none}.border-amber-200{border-color:var(--color-amber-200)}.border-amber-300{border-color:var(--color-amber-300)}.border-amber-300\\/60{border-color:#ffd23699}@supports (color:color-mix(in lab, red, red)){.border-amber-300\\/60{border-color:color-mix(in oklab, var(--color-amber-300) 60%, transparent)}}.border-amber-400{border-color:var(--color-amber-400)}.border-amber-500{border-color:var(--color-amber-500)}.border-amber-500\\/30{border-color:#f99c004d}@supports (color:color-mix(in lab, red, red)){.border-amber-500\\/30{border-color:color-mix(in oklab, var(--color-amber-500) 30%, transparent)}}.border-amber-500\\/40{border-color:#f99c0066}@supports (color:color-mix(in lab, red, red)){.border-amber-500\\/40{border-color:color-mix(in oklab, var(--color-amber-500) 40%, transparent)}}.border-amber-500\\/50{border-color:#f99c0080}@supports (color:color-mix(in lab, red, red)){.border-amber-500\\/50{border-color:color-mix(in oklab, var(--color-amber-500) 50%, transparent)}}.border-amber-600\\/40{border-color:#dd740066}@supports (color:color-mix(in lab, red, red)){.border-amber-600\\/40{border-color:color-mix(in oklab, var(--color-amber-600) 40%, transparent)}}.border-blue-300{border-color:var(--color-blue-300)}.border-blue-500{border-color:var(--color-blue-500)}.border-blue-500\\/30{border-color:#3080ff4d}@supports (color:color-mix(in lab, red, red)){.border-blue-500\\/30{border-color:color-mix(in oklab, var(--color-blue-500) 30%, transparent)}}.border-border,.border-border\\/50{border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.border-border\\/50{border-color:color-mix(in oklab, var(--border) 50%, transparent)}}.border-border\\/60{border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.border-border\\/60{border-color:color-mix(in oklab, var(--border) 60%, transparent)}}.border-border\\/70{border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.border-border\\/70{border-color:color-mix(in oklab, var(--border) 70%, transparent)}}.border-border\\/80{border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.border-border\\/80{border-color:color-mix(in oklab, var(--border) 80%, transparent)}}.border-current,.border-current\\/10{border-color:currentColor}@supports (color:color-mix(in lab, red, red)){.border-current\\/10{border-color:color-mix(in oklab, currentcolor 10%, transparent)}}.border-current\\/25{border-color:currentColor}@supports (color:color-mix(in lab, red, red)){.border-current\\/25{border-color:color-mix(in oklab, currentcolor 25%, transparent)}}.border-destructive,.border-destructive\\/20{border-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.border-destructive\\/20{border-color:color-mix(in oklab, var(--destructive) 20%, transparent)}}.border-destructive\\/30{border-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.border-destructive\\/30{border-color:color-mix(in oklab, var(--destructive) 30%, transparent)}}.border-destructive\\/40{border-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.border-destructive\\/40{border-color:color-mix(in oklab, var(--destructive) 40%, transparent)}}.border-destructive\\/50{border-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.border-destructive\\/50{border-color:color-mix(in oklab, var(--destructive) 50%, transparent)}}.border-emerald-200{border-color:var(--color-emerald-200)}.border-emerald-300{border-color:var(--color-emerald-300)}.border-emerald-500{border-color:var(--color-emerald-500)}.border-emerald-500\\/30{border-color:#00bb7f4d}@supports (color:color-mix(in lab, red, red)){.border-emerald-500\\/30{border-color:color-mix(in oklab, var(--color-emerald-500) 30%, transparent)}}.border-emerald-500\\/50{border-color:#00bb7f80}@supports (color:color-mix(in lab, red, red)){.border-emerald-500\\/50{border-color:color-mix(in oklab, var(--color-emerald-500) 50%, transparent)}}.border-emerald-600{border-color:var(--color-emerald-600)}.border-emerald-600\\/40{border-color:#00976766}@supports (color:color-mix(in lab, red, red)){.border-emerald-600\\/40{border-color:color-mix(in oklab, var(--color-emerald-600) 40%, transparent)}}.border-foreground\\/20{border-color:var(--foreground)}@supports (color:color-mix(in lab, red, red)){.border-foreground\\/20{border-color:color-mix(in oklab, var(--foreground) 20%, transparent)}}.border-fuchsia-300{border-color:var(--color-fuchsia-300)}.border-fuchsia-500{border-color:var(--color-fuchsia-500)}.border-green-500{border-color:var(--color-green-500)}.border-green-500\\/50{border-color:#00c75880}@supports (color:color-mix(in lab, red, red)){.border-green-500\\/50{border-color:color-mix(in oklab, var(--color-green-500) 50%, transparent)}}.border-green-600\\/40{border-color:#00a54466}@supports (color:color-mix(in lab, red, red)){.border-green-600\\/40{border-color:color-mix(in oklab, var(--color-green-600) 40%, transparent)}}.border-indigo-300{border-color:var(--color-indigo-300)}.border-indigo-500{border-color:var(--color-indigo-500)}.border-input{border-color:var(--input)}.border-muted-foreground,.border-muted-foreground\\/50{border-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.border-muted-foreground\\/50{border-color:color-mix(in oklab, var(--muted-foreground) 50%, transparent)}}.border-orange-300{border-color:var(--color-orange-300)}.border-orange-500{border-color:var(--color-orange-500)}.border-pink-500{border-color:var(--color-pink-500)}.border-primary,.border-primary\\/30{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.border-primary\\/30{border-color:color-mix(in oklab, var(--primary) 30%, transparent)}}.border-primary\\/40{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.border-primary\\/40{border-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.border-primary\\/50{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.border-primary\\/50{border-color:color-mix(in oklab, var(--primary) 50%, transparent)}}.border-primary\\/60{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.border-primary\\/60{border-color:color-mix(in oklab, var(--primary) 60%, transparent)}}.border-primary\\/70{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.border-primary\\/70{border-color:color-mix(in oklab, var(--primary) 70%, transparent)}}.border-purple-500{border-color:var(--color-purple-500)}.border-purple-500\\/30{border-color:#ac4bff4d}@supports (color:color-mix(in lab, red, red)){.border-purple-500\\/30{border-color:color-mix(in oklab, var(--color-purple-500) 30%, transparent)}}.border-red-300{border-color:var(--color-red-300)}.border-red-400{border-color:var(--color-red-400)}.border-red-500{border-color:var(--color-red-500)}.border-red-500\\/30{border-color:#fb2c364d}@supports (color:color-mix(in lab, red, red)){.border-red-500\\/30{border-color:color-mix(in oklab, var(--color-red-500) 30%, transparent)}}.border-red-500\\/40{border-color:#fb2c3666}@supports (color:color-mix(in lab, red, red)){.border-red-500\\/40{border-color:color-mix(in oklab, var(--color-red-500) 40%, transparent)}}.border-red-500\\/50{border-color:#fb2c3680}@supports (color:color-mix(in lab, red, red)){.border-red-500\\/50{border-color:color-mix(in oklab, var(--color-red-500) 50%, transparent)}}.border-red-500\\/60{border-color:#fb2c3699}@supports (color:color-mix(in lab, red, red)){.border-red-500\\/60{border-color:color-mix(in oklab, var(--color-red-500) 60%, transparent)}}.border-red-500\\/70{border-color:#fb2c36b3}@supports (color:color-mix(in lab, red, red)){.border-red-500\\/70{border-color:color-mix(in oklab, var(--color-red-500) 70%, transparent)}}.border-red-600{border-color:var(--color-red-600)}.border-rose-200{border-color:var(--color-rose-200)}.border-rose-300{border-color:var(--color-rose-300)}.border-rose-500{border-color:var(--color-rose-500)}.border-rose-500\\/30{border-color:#ff23574d}@supports (color:color-mix(in lab, red, red)){.border-rose-500\\/30{border-color:color-mix(in oklab, var(--color-rose-500) 30%, transparent)}}.border-sidebar-border{border-color:var(--sidebar-border)}.border-sky-200{border-color:var(--color-sky-200)}.border-sky-300{border-color:var(--color-sky-300)}.border-sky-500{border-color:var(--color-sky-500)}.border-sky-500\\/20{border-color:#00a5ef33}@supports (color:color-mix(in lab, red, red)){.border-sky-500\\/20{border-color:color-mix(in oklab, var(--color-sky-500) 20%, transparent)}}.border-sky-500\\/25{border-color:#00a5ef40}@supports (color:color-mix(in lab, red, red)){.border-sky-500\\/25{border-color:color-mix(in oklab, var(--color-sky-500) 25%, transparent)}}.border-sky-500\\/30{border-color:#00a5ef4d}@supports (color:color-mix(in lab, red, red)){.border-sky-500\\/30{border-color:color-mix(in oklab, var(--color-sky-500) 30%, transparent)}}.border-sky-500\\/50{border-color:#00a5ef80}@supports (color:color-mix(in lab, red, red)){.border-sky-500\\/50{border-color:color-mix(in oklab, var(--color-sky-500) 50%, transparent)}}.border-sky-500\\/60{border-color:#00a5ef99}@supports (color:color-mix(in lab, red, red)){.border-sky-500\\/60{border-color:color-mix(in oklab, var(--color-sky-500) 60%, transparent)}}.border-sky-600{border-color:var(--color-sky-600)}.border-slate-600{border-color:var(--color-slate-600)}.border-teal-300{border-color:var(--color-teal-300)}.border-teal-500{border-color:var(--color-teal-500)}.border-transparent{border-color:#0000}.border-violet-300{border-color:var(--color-violet-300)}.border-violet-500{border-color:var(--color-violet-500)}.border-violet-500\\/30{border-color:#8d54ff4d}@supports (color:color-mix(in lab, red, red)){.border-violet-500\\/30{border-color:color-mix(in oklab, var(--color-violet-500) 30%, transparent)}}.border-white\\/10{border-color:#ffffff1a}@supports (color:color-mix(in lab, red, red)){.border-white\\/10{border-color:color-mix(in oklab, var(--color-white) 10%, transparent)}}.border-yellow-300{border-color:var(--color-yellow-300)}.border-yellow-400{border-color:var(--color-yellow-400)}.border-yellow-500{border-color:var(--color-yellow-500)}.border-yellow-500\\/30{border-color:#edb2004d}@supports (color:color-mix(in lab, red, red)){.border-yellow-500\\/30{border-color:color-mix(in oklab, var(--color-yellow-500) 30%, transparent)}}.border-zinc-800{border-color:var(--color-zinc-800)}.border-t-primary{border-top-color:var(--primary)}.border-l-amber-400{border-left-color:var(--color-amber-400)}.border-l-amber-500{border-left-color:var(--color-amber-500)}.border-l-amber-500\\/60{border-left-color:#f99c0099}@supports (color:color-mix(in lab, red, red)){.border-l-amber-500\\/60{border-left-color:color-mix(in oklab, var(--color-amber-500) 60%, transparent)}}.border-l-blue-500{border-left-color:var(--color-blue-500)}.border-l-border{border-left-color:var(--border)}.border-l-current\\/25{border-left-color:currentColor}@supports (color:color-mix(in lab, red, red)){.border-l-current\\/25{border-left-color:color-mix(in oklab, currentcolor 25%, transparent)}}.border-l-emerald-400{border-left-color:var(--color-emerald-400)}.border-l-emerald-500{border-left-color:var(--color-emerald-500)}.border-l-green-500{border-left-color:var(--color-green-500)}.border-l-indigo-400{border-left-color:var(--color-indigo-400)}.border-l-purple-500{border-left-color:var(--color-purple-500)}.border-l-red-500{border-left-color:var(--color-red-500)}.border-l-rose-400{border-left-color:var(--color-rose-400)}.border-l-sky-400{border-left-color:var(--color-sky-400)}.border-l-slate-400{border-left-color:var(--color-slate-400)}.border-l-teal-400{border-left-color:var(--color-teal-400)}.border-l-transparent{border-left-color:#0000}.border-l-violet-400{border-left-color:var(--color-violet-400)}.border-l-yellow-500{border-left-color:var(--color-yellow-500)}.bg-\\[\\#000000\\],.bg-\\[rgb\\(0\\,0\\,0\\)\\]{background-color:#000}.bg-\\[var\\(--fs-success-500\\)\\]{background-color:var(--fs-success-500)}.bg-accent,.bg-accent\\/30{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.bg-accent\\/30{background-color:color-mix(in oklab, var(--accent) 30%, transparent)}}.bg-accent\\/50{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.bg-accent\\/50{background-color:color-mix(in oklab, var(--accent) 50%, transparent)}}.bg-accent\\/60{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.bg-accent\\/60{background-color:color-mix(in oklab, var(--accent) 60%, transparent)}}.bg-accent\\/70{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.bg-accent\\/70{background-color:color-mix(in oklab, var(--accent) 70%, transparent)}}.bg-amber-50{background-color:var(--color-amber-50)}.bg-amber-50\\/40{background-color:#fffbeb66}@supports (color:color-mix(in lab, red, red)){.bg-amber-50\\/40{background-color:color-mix(in oklab, var(--color-amber-50) 40%, transparent)}}.bg-amber-50\\/85{background-color:#fffbebd9}@supports (color:color-mix(in lab, red, red)){.bg-amber-50\\/85{background-color:color-mix(in oklab, var(--color-amber-50) 85%, transparent)}}.bg-amber-100{background-color:var(--color-amber-100)}.bg-amber-100\\/30{background-color:#fef3c64d}@supports (color:color-mix(in lab, red, red)){.bg-amber-100\\/30{background-color:color-mix(in oklab, var(--color-amber-100) 30%, transparent)}}.bg-amber-100\\/70{background-color:#fef3c6b3}@supports (color:color-mix(in lab, red, red)){.bg-amber-100\\/70{background-color:color-mix(in oklab, var(--color-amber-100) 70%, transparent)}}.bg-amber-200{background-color:var(--color-amber-200)}.bg-amber-200\\/50{background-color:#fee68580}@supports (color:color-mix(in lab, red, red)){.bg-amber-200\\/50{background-color:color-mix(in oklab, var(--color-amber-200) 50%, transparent)}}.bg-amber-200\\/60{background-color:#fee68599}@supports (color:color-mix(in lab, red, red)){.bg-amber-200\\/60{background-color:color-mix(in oklab, var(--color-amber-200) 60%, transparent)}}.bg-amber-300{background-color:var(--color-amber-300)}.bg-amber-400{background-color:var(--color-amber-400)}.bg-amber-400\\/10{background-color:#fcbb001a}@supports (color:color-mix(in lab, red, red)){.bg-amber-400\\/10{background-color:color-mix(in oklab, var(--color-amber-400) 10%, transparent)}}.bg-amber-400\\/15{background-color:#fcbb0026}@supports (color:color-mix(in lab, red, red)){.bg-amber-400\\/15{background-color:color-mix(in oklab, var(--color-amber-400) 15%, transparent)}}.bg-amber-500{background-color:var(--color-amber-500)}.bg-amber-500\\/5{background-color:#f99c000d}@supports (color:color-mix(in lab, red, red)){.bg-amber-500\\/5{background-color:color-mix(in oklab, var(--color-amber-500) 5%, transparent)}}.bg-amber-500\\/10{background-color:#f99c001a}@supports (color:color-mix(in lab, red, red)){.bg-amber-500\\/10{background-color:color-mix(in oklab, var(--color-amber-500) 10%, transparent)}}.bg-amber-500\\/15{background-color:#f99c0026}@supports (color:color-mix(in lab, red, red)){.bg-amber-500\\/15{background-color:color-mix(in oklab, var(--color-amber-500) 15%, transparent)}}.bg-amber-500\\/\\[0\\.06\\]{background-color:#f99c000f}@supports (color:color-mix(in lab, red, red)){.bg-amber-500\\/\\[0\\.06\\]{background-color:color-mix(in oklab, var(--color-amber-500) 6%, transparent)}}.bg-amber-600{background-color:var(--color-amber-600)}.bg-amber-700{background-color:var(--color-amber-700)}.bg-amber-800{background-color:var(--color-amber-800)}.bg-amber-900{background-color:var(--color-amber-900)}.bg-amber-950{background-color:var(--color-amber-950)}.bg-background,.bg-background\\/80{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.bg-background\\/80{background-color:color-mix(in oklab, var(--background) 80%, transparent)}}.bg-background\\/90{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.bg-background\\/90{background-color:color-mix(in oklab, var(--background) 90%, transparent)}}.bg-background\\/95{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.bg-background\\/95{background-color:color-mix(in oklab, var(--background) 95%, transparent)}}.bg-black{background-color:var(--color-black)}.bg-black\\/20{background-color:#0003}@supports (color:color-mix(in lab, red, red)){.bg-black\\/20{background-color:color-mix(in oklab, var(--color-black) 20%, transparent)}}.bg-black\\/40{background-color:#0006}@supports (color:color-mix(in lab, red, red)){.bg-black\\/40{background-color:color-mix(in oklab, var(--color-black) 40%, transparent)}}.bg-black\\/60{background-color:#0009}@supports (color:color-mix(in lab, red, red)){.bg-black\\/60{background-color:color-mix(in oklab, var(--color-black) 60%, transparent)}}.bg-blue-50{background-color:var(--color-blue-50)}.bg-blue-100{background-color:var(--color-blue-100)}.bg-blue-200{background-color:var(--color-blue-200)}.bg-blue-300{background-color:var(--color-blue-300)}.bg-blue-400{background-color:var(--color-blue-400)}.bg-blue-500{background-color:var(--color-blue-500)}.bg-blue-500\\/5{background-color:#3080ff0d}@supports (color:color-mix(in lab, red, red)){.bg-blue-500\\/5{background-color:color-mix(in oklab, var(--color-blue-500) 5%, transparent)}}.bg-blue-500\\/10{background-color:#3080ff1a}@supports (color:color-mix(in lab, red, red)){.bg-blue-500\\/10{background-color:color-mix(in oklab, var(--color-blue-500) 10%, transparent)}}.bg-blue-600{background-color:var(--color-blue-600)}.bg-blue-700{background-color:var(--color-blue-700)}.bg-blue-800{background-color:var(--color-blue-800)}.bg-blue-900{background-color:var(--color-blue-900)}.bg-blue-950{background-color:var(--color-blue-950)}.bg-border{background-color:var(--border)}.bg-card{background-color:var(--card)}.bg-current{background-color:currentColor}.bg-cyan-50{background-color:var(--color-cyan-50)}.bg-cyan-100{background-color:var(--color-cyan-100)}.bg-cyan-200{background-color:var(--color-cyan-200)}.bg-cyan-300{background-color:var(--color-cyan-300)}.bg-cyan-400{background-color:var(--color-cyan-400)}.bg-cyan-500{background-color:var(--color-cyan-500)}.bg-cyan-600{background-color:var(--color-cyan-600)}.bg-cyan-700{background-color:var(--color-cyan-700)}.bg-cyan-800{background-color:var(--color-cyan-800)}.bg-cyan-900{background-color:var(--color-cyan-900)}.bg-cyan-950{background-color:var(--color-cyan-950)}.bg-destructive,.bg-destructive\\/5{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.bg-destructive\\/5{background-color:color-mix(in oklab, var(--destructive) 5%, transparent)}}.bg-destructive\\/10{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.bg-destructive\\/10{background-color:color-mix(in oklab, var(--destructive) 10%, transparent)}}.bg-destructive\\/20{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.bg-destructive\\/20{background-color:color-mix(in oklab, var(--destructive) 20%, transparent)}}.bg-destructive\\/\\[0\\.06\\]{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.bg-destructive\\/\\[0\\.06\\]{background-color:color-mix(in oklab, var(--destructive) 6%, transparent)}}.bg-emerald-50{background-color:var(--color-emerald-50)}.bg-emerald-50\\/80{background-color:#ecfdf5cc}@supports (color:color-mix(in lab, red, red)){.bg-emerald-50\\/80{background-color:color-mix(in oklab, var(--color-emerald-50) 80%, transparent)}}.bg-emerald-100{background-color:var(--color-emerald-100)}.bg-emerald-200{background-color:var(--color-emerald-200)}.bg-emerald-300{background-color:var(--color-emerald-300)}.bg-emerald-400{background-color:var(--color-emerald-400)}.bg-emerald-500{background-color:var(--color-emerald-500)}.bg-emerald-500\\/5{background-color:#00bb7f0d}@supports (color:color-mix(in lab, red, red)){.bg-emerald-500\\/5{background-color:color-mix(in oklab, var(--color-emerald-500) 5%, transparent)}}.bg-emerald-500\\/10{background-color:#00bb7f1a}@supports (color:color-mix(in lab, red, red)){.bg-emerald-500\\/10{background-color:color-mix(in oklab, var(--color-emerald-500) 10%, transparent)}}.bg-emerald-500\\/12{background-color:#00bb7f1f}@supports (color:color-mix(in lab, red, red)){.bg-emerald-500\\/12{background-color:color-mix(in oklab, var(--color-emerald-500) 12%, transparent)}}.bg-emerald-500\\/15{background-color:#00bb7f26}@supports (color:color-mix(in lab, red, red)){.bg-emerald-500\\/15{background-color:color-mix(in oklab, var(--color-emerald-500) 15%, transparent)}}.bg-emerald-600{background-color:var(--color-emerald-600)}.bg-emerald-700{background-color:var(--color-emerald-700)}.bg-emerald-800{background-color:var(--color-emerald-800)}.bg-emerald-900{background-color:var(--color-emerald-900)}.bg-emerald-950{background-color:var(--color-emerald-950)}.bg-foreground{background-color:var(--foreground)}.bg-fuchsia-50{background-color:var(--color-fuchsia-50)}.bg-fuchsia-100{background-color:var(--color-fuchsia-100)}.bg-fuchsia-200{background-color:var(--color-fuchsia-200)}.bg-fuchsia-300{background-color:var(--color-fuchsia-300)}.bg-fuchsia-400{background-color:var(--color-fuchsia-400)}.bg-fuchsia-500{background-color:var(--color-fuchsia-500)}.bg-fuchsia-600{background-color:var(--color-fuchsia-600)}.bg-fuchsia-700{background-color:var(--color-fuchsia-700)}.bg-fuchsia-800{background-color:var(--color-fuchsia-800)}.bg-fuchsia-900{background-color:var(--color-fuchsia-900)}.bg-fuchsia-950{background-color:var(--color-fuchsia-950)}.bg-gray-50{background-color:var(--color-gray-50)}.bg-gray-100{background-color:var(--color-gray-100)}.bg-gray-200{background-color:var(--color-gray-200)}.bg-gray-300{background-color:var(--color-gray-300)}.bg-gray-400{background-color:var(--color-gray-400)}.bg-gray-500{background-color:var(--color-gray-500)}.bg-gray-600{background-color:var(--color-gray-600)}.bg-gray-700{background-color:var(--color-gray-700)}.bg-gray-800{background-color:var(--color-gray-800)}.bg-gray-900{background-color:var(--color-gray-900)}.bg-gray-950{background-color:var(--color-gray-950)}.bg-green-50{background-color:var(--color-green-50)}.bg-green-100{background-color:var(--color-green-100)}.bg-green-200{background-color:var(--color-green-200)}.bg-green-300{background-color:var(--color-green-300)}.bg-green-400{background-color:var(--color-green-400)}.bg-green-500{background-color:var(--color-green-500)}.bg-green-500\\/5{background-color:#00c7580d}@supports (color:color-mix(in lab, red, red)){.bg-green-500\\/5{background-color:color-mix(in oklab, var(--color-green-500) 5%, transparent)}}.bg-green-500\\/10{background-color:#00c7581a}@supports (color:color-mix(in lab, red, red)){.bg-green-500\\/10{background-color:color-mix(in oklab, var(--color-green-500) 10%, transparent)}}.bg-green-500\\/20{background-color:#00c75833}@supports (color:color-mix(in lab, red, red)){.bg-green-500\\/20{background-color:color-mix(in oklab, var(--color-green-500) 20%, transparent)}}.bg-green-500\\/80{background-color:#00c758cc}@supports (color:color-mix(in lab, red, red)){.bg-green-500\\/80{background-color:color-mix(in oklab, var(--color-green-500) 80%, transparent)}}.bg-green-600{background-color:var(--color-green-600)}.bg-green-700{background-color:var(--color-green-700)}.bg-green-800{background-color:var(--color-green-800)}.bg-green-900{background-color:var(--color-green-900)}.bg-green-950{background-color:var(--color-green-950)}.bg-indigo-50{background-color:var(--color-indigo-50)}.bg-indigo-100{background-color:var(--color-indigo-100)}.bg-indigo-200{background-color:var(--color-indigo-200)}.bg-indigo-300{background-color:var(--color-indigo-300)}.bg-indigo-400{background-color:var(--color-indigo-400)}.bg-indigo-500{background-color:var(--color-indigo-500)}.bg-indigo-600{background-color:var(--color-indigo-600)}.bg-indigo-700{background-color:var(--color-indigo-700)}.bg-indigo-800{background-color:var(--color-indigo-800)}.bg-indigo-900{background-color:var(--color-indigo-900)}.bg-indigo-950{background-color:var(--color-indigo-950)}.bg-input{background-color:var(--input)}.bg-lime-50{background-color:var(--color-lime-50)}.bg-lime-100{background-color:var(--color-lime-100)}.bg-lime-200{background-color:var(--color-lime-200)}.bg-lime-300{background-color:var(--color-lime-300)}.bg-lime-400{background-color:var(--color-lime-400)}.bg-lime-500{background-color:var(--color-lime-500)}.bg-lime-600{background-color:var(--color-lime-600)}.bg-lime-700{background-color:var(--color-lime-700)}.bg-lime-800{background-color:var(--color-lime-800)}.bg-lime-900{background-color:var(--color-lime-900)}.bg-lime-950{background-color:var(--color-lime-950)}.bg-muted{background-color:var(--muted)}.bg-muted-foreground,.bg-muted-foreground\\/30{background-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.bg-muted-foreground\\/30{background-color:color-mix(in oklab, var(--muted-foreground) 30%, transparent)}}.bg-muted-foreground\\/40{background-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.bg-muted-foreground\\/40{background-color:color-mix(in oklab, var(--muted-foreground) 40%, transparent)}}.bg-muted-foreground\\/50{background-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.bg-muted-foreground\\/50{background-color:color-mix(in oklab, var(--muted-foreground) 50%, transparent)}}.bg-muted-foreground\\/60{background-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.bg-muted-foreground\\/60{background-color:color-mix(in oklab, var(--muted-foreground) 60%, transparent)}}.bg-muted\\/20{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/20{background-color:color-mix(in oklab, var(--muted) 20%, transparent)}}.bg-muted\\/25{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/25{background-color:color-mix(in oklab, var(--muted) 25%, transparent)}}.bg-muted\\/30{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/30{background-color:color-mix(in oklab, var(--muted) 30%, transparent)}}.bg-muted\\/35{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/35{background-color:color-mix(in oklab, var(--muted) 35%, transparent)}}.bg-muted\\/40{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/40{background-color:color-mix(in oklab, var(--muted) 40%, transparent)}}.bg-muted\\/50{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/50{background-color:color-mix(in oklab, var(--muted) 50%, transparent)}}.bg-muted\\/60{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/60{background-color:color-mix(in oklab, var(--muted) 60%, transparent)}}.bg-muted\\/70{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/70{background-color:color-mix(in oklab, var(--muted) 70%, transparent)}}.bg-muted\\/80{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/80{background-color:color-mix(in oklab, var(--muted) 80%, transparent)}}.bg-muted\\/90{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/90{background-color:color-mix(in oklab, var(--muted) 90%, transparent)}}.bg-neutral-50{background-color:var(--color-neutral-50)}.bg-neutral-100{background-color:var(--color-neutral-100)}.bg-neutral-200{background-color:var(--color-neutral-200)}.bg-neutral-300{background-color:var(--color-neutral-300)}.bg-neutral-400{background-color:var(--color-neutral-400)}.bg-neutral-500{background-color:var(--color-neutral-500)}.bg-neutral-600{background-color:var(--color-neutral-600)}.bg-neutral-700{background-color:var(--color-neutral-700)}.bg-neutral-800{background-color:var(--color-neutral-800)}.bg-neutral-900{background-color:var(--color-neutral-900)}.bg-neutral-950{background-color:var(--color-neutral-950)}.bg-orange-50{background-color:var(--color-orange-50)}.bg-orange-100{background-color:var(--color-orange-100)}.bg-orange-200{background-color:var(--color-orange-200)}.bg-orange-300{background-color:var(--color-orange-300)}.bg-orange-400{background-color:var(--color-orange-400)}.bg-orange-500{background-color:var(--color-orange-500)}.bg-orange-500\\/5{background-color:#fe6e000d}@supports (color:color-mix(in lab, red, red)){.bg-orange-500\\/5{background-color:color-mix(in oklab, var(--color-orange-500) 5%, transparent)}}.bg-orange-600{background-color:var(--color-orange-600)}.bg-orange-700{background-color:var(--color-orange-700)}.bg-orange-800{background-color:var(--color-orange-800)}.bg-orange-900{background-color:var(--color-orange-900)}.bg-orange-950{background-color:var(--color-orange-950)}.bg-pink-50{background-color:var(--color-pink-50)}.bg-pink-100{background-color:var(--color-pink-100)}.bg-pink-200{background-color:var(--color-pink-200)}.bg-pink-300{background-color:var(--color-pink-300)}.bg-pink-400{background-color:var(--color-pink-400)}.bg-pink-500{background-color:var(--color-pink-500)}.bg-pink-600{background-color:var(--color-pink-600)}.bg-pink-700{background-color:var(--color-pink-700)}.bg-pink-800{background-color:var(--color-pink-800)}.bg-pink-900{background-color:var(--color-pink-900)}.bg-pink-950{background-color:var(--color-pink-950)}.bg-popover{background-color:var(--popover)}.bg-primary,.bg-primary\\/5{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/5{background-color:color-mix(in oklab, var(--primary) 5%, transparent)}}.bg-primary\\/8{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/8{background-color:color-mix(in oklab, var(--primary) 8%, transparent)}}.bg-primary\\/10{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/10{background-color:color-mix(in oklab, var(--primary) 10%, transparent)}}.bg-primary\\/15{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/15{background-color:color-mix(in oklab, var(--primary) 15%, transparent)}}.bg-primary\\/20{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/20{background-color:color-mix(in oklab, var(--primary) 20%, transparent)}}.bg-primary\\/60{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/60{background-color:color-mix(in oklab, var(--primary) 60%, transparent)}}.bg-primary\\/70{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/70{background-color:color-mix(in oklab, var(--primary) 70%, transparent)}}.bg-primary\\/\\[0\\.04\\]{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/\\[0\\.04\\]{background-color:color-mix(in oklab, var(--primary) 4%, transparent)}}.bg-primary\\/\\[0\\.05\\]{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/\\[0\\.05\\]{background-color:color-mix(in oklab, var(--primary) 5%, transparent)}}.bg-primary\\/\\[0\\.06\\]{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/\\[0\\.06\\]{background-color:color-mix(in oklab, var(--primary) 6%, transparent)}}.bg-primary\\/\\[0\\.08\\]{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/\\[0\\.08\\]{background-color:color-mix(in oklab, var(--primary) 8%, transparent)}}.bg-purple-50{background-color:var(--color-purple-50)}.bg-purple-100{background-color:var(--color-purple-100)}.bg-purple-200{background-color:var(--color-purple-200)}.bg-purple-300{background-color:var(--color-purple-300)}.bg-purple-400{background-color:var(--color-purple-400)}.bg-purple-500{background-color:var(--color-purple-500)}.bg-purple-500\\/5{background-color:#ac4bff0d}@supports (color:color-mix(in lab, red, red)){.bg-purple-500\\/5{background-color:color-mix(in oklab, var(--color-purple-500) 5%, transparent)}}.bg-purple-500\\/10{background-color:#ac4bff1a}@supports (color:color-mix(in lab, red, red)){.bg-purple-500\\/10{background-color:color-mix(in oklab, var(--color-purple-500) 10%, transparent)}}.bg-purple-600{background-color:var(--color-purple-600)}.bg-purple-700{background-color:var(--color-purple-700)}.bg-purple-800{background-color:var(--color-purple-800)}.bg-purple-900{background-color:var(--color-purple-900)}.bg-purple-950{background-color:var(--color-purple-950)}.bg-red-50{background-color:var(--color-red-50)}.bg-red-100{background-color:var(--color-red-100)}.bg-red-200{background-color:var(--color-red-200)}.bg-red-300{background-color:var(--color-red-300)}.bg-red-400{background-color:var(--color-red-400)}.bg-red-500{background-color:var(--color-red-500)}.bg-red-500\\/5{background-color:#fb2c360d}@supports (color:color-mix(in lab, red, red)){.bg-red-500\\/5{background-color:color-mix(in oklab, var(--color-red-500) 5%, transparent)}}.bg-red-500\\/10{background-color:#fb2c361a}@supports (color:color-mix(in lab, red, red)){.bg-red-500\\/10{background-color:color-mix(in oklab, var(--color-red-500) 10%, transparent)}}.bg-red-500\\/20{background-color:#fb2c3633}@supports (color:color-mix(in lab, red, red)){.bg-red-500\\/20{background-color:color-mix(in oklab, var(--color-red-500) 20%, transparent)}}.bg-red-500\\/80{background-color:#fb2c36cc}@supports (color:color-mix(in lab, red, red)){.bg-red-500\\/80{background-color:color-mix(in oklab, var(--color-red-500) 80%, transparent)}}.bg-red-600{background-color:var(--color-red-600)}.bg-red-700{background-color:var(--color-red-700)}.bg-red-800{background-color:var(--color-red-800)}.bg-red-900{background-color:var(--color-red-900)}.bg-red-950{background-color:var(--color-red-950)}.bg-rose-50{background-color:var(--color-rose-50)}.bg-rose-50\\/85{background-color:#fff1f2d9}@supports (color:color-mix(in lab, red, red)){.bg-rose-50\\/85{background-color:color-mix(in oklab, var(--color-rose-50) 85%, transparent)}}.bg-rose-100{background-color:var(--color-rose-100)}.bg-rose-200{background-color:var(--color-rose-200)}.bg-rose-300{background-color:var(--color-rose-300)}.bg-rose-400{background-color:var(--color-rose-400)}.bg-rose-500{background-color:var(--color-rose-500)}.bg-rose-500\\/10{background-color:#ff23571a}@supports (color:color-mix(in lab, red, red)){.bg-rose-500\\/10{background-color:color-mix(in oklab, var(--color-rose-500) 10%, transparent)}}.bg-rose-500\\/12{background-color:#ff23571f}@supports (color:color-mix(in lab, red, red)){.bg-rose-500\\/12{background-color:color-mix(in oklab, var(--color-rose-500) 12%, transparent)}}.bg-rose-500\\/15{background-color:#ff235726}@supports (color:color-mix(in lab, red, red)){.bg-rose-500\\/15{background-color:color-mix(in oklab, var(--color-rose-500) 15%, transparent)}}.bg-rose-600{background-color:var(--color-rose-600)}.bg-rose-700{background-color:var(--color-rose-700)}.bg-rose-800{background-color:var(--color-rose-800)}.bg-rose-900{background-color:var(--color-rose-900)}.bg-rose-950{background-color:var(--color-rose-950)}.bg-secondary,.bg-secondary\\/50{background-color:var(--secondary)}@supports (color:color-mix(in lab, red, red)){.bg-secondary\\/50{background-color:color-mix(in oklab, var(--secondary) 50%, transparent)}}.bg-sidebar{background-color:var(--sidebar)}.bg-sidebar-accent,.bg-sidebar-accent\\/50{background-color:var(--sidebar-accent)}@supports (color:color-mix(in lab, red, red)){.bg-sidebar-accent\\/50{background-color:color-mix(in oklab, var(--sidebar-accent) 50%, transparent)}}.bg-sky-50{background-color:var(--color-sky-50)}.bg-sky-100{background-color:var(--color-sky-100)}.bg-sky-200{background-color:var(--color-sky-200)}.bg-sky-300{background-color:var(--color-sky-300)}.bg-sky-400{background-color:var(--color-sky-400)}.bg-sky-500{background-color:var(--color-sky-500)}.bg-sky-500\\/5{background-color:#00a5ef0d}@supports (color:color-mix(in lab, red, red)){.bg-sky-500\\/5{background-color:color-mix(in oklab, var(--color-sky-500) 5%, transparent)}}.bg-sky-500\\/10{background-color:#00a5ef1a}@supports (color:color-mix(in lab, red, red)){.bg-sky-500\\/10{background-color:color-mix(in oklab, var(--color-sky-500) 10%, transparent)}}.bg-sky-500\\/12{background-color:#00a5ef1f}@supports (color:color-mix(in lab, red, red)){.bg-sky-500\\/12{background-color:color-mix(in oklab, var(--color-sky-500) 12%, transparent)}}.bg-sky-500\\/15{background-color:#00a5ef26}@supports (color:color-mix(in lab, red, red)){.bg-sky-500\\/15{background-color:color-mix(in oklab, var(--color-sky-500) 15%, transparent)}}.bg-sky-600{background-color:var(--color-sky-600)}.bg-sky-700{background-color:var(--color-sky-700)}.bg-sky-800{background-color:var(--color-sky-800)}.bg-sky-900{background-color:var(--color-sky-900)}.bg-sky-950{background-color:var(--color-sky-950)}.bg-slate-50{background-color:var(--color-slate-50)}.bg-slate-100{background-color:var(--color-slate-100)}.bg-slate-200{background-color:var(--color-slate-200)}.bg-slate-300{background-color:var(--color-slate-300)}.bg-slate-400{background-color:var(--color-slate-400)}.bg-slate-500{background-color:var(--color-slate-500)}.bg-slate-600{background-color:var(--color-slate-600)}.bg-slate-700{background-color:var(--color-slate-700)}.bg-slate-800{background-color:var(--color-slate-800)}.bg-slate-900{background-color:var(--color-slate-900)}.bg-slate-950{background-color:var(--color-slate-950)}.bg-stone-50{background-color:var(--color-stone-50)}.bg-stone-100{background-color:var(--color-stone-100)}.bg-stone-200{background-color:var(--color-stone-200)}.bg-stone-300{background-color:var(--color-stone-300)}.bg-stone-400{background-color:var(--color-stone-400)}.bg-stone-500{background-color:var(--color-stone-500)}.bg-stone-600{background-color:var(--color-stone-600)}.bg-stone-700{background-color:var(--color-stone-700)}.bg-stone-800{background-color:var(--color-stone-800)}.bg-stone-900{background-color:var(--color-stone-900)}.bg-stone-950{background-color:var(--color-stone-950)}.bg-teal-50{background-color:var(--color-teal-50)}.bg-teal-100{background-color:var(--color-teal-100)}.bg-teal-200{background-color:var(--color-teal-200)}.bg-teal-300{background-color:var(--color-teal-300)}.bg-teal-400{background-color:var(--color-teal-400)}.bg-teal-500{background-color:var(--color-teal-500)}.bg-teal-600{background-color:var(--color-teal-600)}.bg-teal-700{background-color:var(--color-teal-700)}.bg-teal-800{background-color:var(--color-teal-800)}.bg-teal-900{background-color:var(--color-teal-900)}.bg-teal-950{background-color:var(--color-teal-950)}.bg-transparent{background-color:#0000}.bg-violet-50{background-color:var(--color-violet-50)}.bg-violet-100{background-color:var(--color-violet-100)}.bg-violet-200{background-color:var(--color-violet-200)}.bg-violet-300{background-color:var(--color-violet-300)}.bg-violet-400{background-color:var(--color-violet-400)}.bg-violet-500{background-color:var(--color-violet-500)}.bg-violet-500\\/5{background-color:#8d54ff0d}@supports (color:color-mix(in lab, red, red)){.bg-violet-500\\/5{background-color:color-mix(in oklab, var(--color-violet-500) 5%, transparent)}}.bg-violet-500\\/10{background-color:#8d54ff1a}@supports (color:color-mix(in lab, red, red)){.bg-violet-500\\/10{background-color:color-mix(in oklab, var(--color-violet-500) 10%, transparent)}}.bg-violet-600{background-color:var(--color-violet-600)}.bg-violet-700{background-color:var(--color-violet-700)}.bg-violet-800{background-color:var(--color-violet-800)}.bg-violet-900{background-color:var(--color-violet-900)}.bg-violet-950{background-color:var(--color-violet-950)}.bg-white{background-color:var(--color-white)}.bg-white\\/10{background-color:#ffffff1a}@supports (color:color-mix(in lab, red, red)){.bg-white\\/10{background-color:color-mix(in oklab, var(--color-white) 10%, transparent)}}.bg-yellow-50{background-color:var(--color-yellow-50)}.bg-yellow-100{background-color:var(--color-yellow-100)}.bg-yellow-200{background-color:var(--color-yellow-200)}.bg-yellow-300{background-color:var(--color-yellow-300)}.bg-yellow-400{background-color:var(--color-yellow-400)}.bg-yellow-500{background-color:var(--color-yellow-500)}.bg-yellow-500\\/5{background-color:#edb2000d}@supports (color:color-mix(in lab, red, red)){.bg-yellow-500\\/5{background-color:color-mix(in oklab, var(--color-yellow-500) 5%, transparent)}}.bg-yellow-600{background-color:var(--color-yellow-600)}.bg-yellow-700{background-color:var(--color-yellow-700)}.bg-yellow-800{background-color:var(--color-yellow-800)}.bg-yellow-900{background-color:var(--color-yellow-900)}.bg-yellow-950{background-color:var(--color-yellow-950)}.bg-zinc-50{background-color:var(--color-zinc-50)}.bg-zinc-100{background-color:var(--color-zinc-100)}.bg-zinc-200{background-color:var(--color-zinc-200)}.bg-zinc-300{background-color:var(--color-zinc-300)}.bg-zinc-400{background-color:var(--color-zinc-400)}.bg-zinc-500{background-color:var(--color-zinc-500)}.bg-zinc-600{background-color:var(--color-zinc-600)}.bg-zinc-700{background-color:var(--color-zinc-700)}.bg-zinc-800{background-color:var(--color-zinc-800)}.bg-zinc-900{background-color:var(--color-zinc-900)}.bg-zinc-950{background-color:var(--color-zinc-950)}.bg-gradient-to-l{--tw-gradient-position:to left in oklab;background-image:linear-gradient(var(--tw-gradient-stops))}.bg-gradient-to-r{--tw-gradient-position:to right in oklab;background-image:linear-gradient(var(--tw-gradient-stops))}.bg-gradient-to-t{--tw-gradient-position:to top in oklab;background-image:linear-gradient(var(--tw-gradient-stops))}.from-background{--tw-gradient-from:var(--background);--tw-gradient-stops:var(--tw-gradient-via-stops,var(--tw-gradient-position), var(--tw-gradient-from) var(--tw-gradient-from-position), var(--tw-gradient-to) var(--tw-gradient-to-position))}.from-border\\/30{--tw-gradient-from:var(--border)}@supports (color:color-mix(in lab, red, red)){.from-border\\/30{--tw-gradient-from:color-mix(in oklab, var(--border) 30%, transparent)}}.from-border\\/30{--tw-gradient-stops:var(--tw-gradient-via-stops,var(--tw-gradient-position), var(--tw-gradient-from) var(--tw-gradient-from-position), var(--tw-gradient-to) var(--tw-gradient-to-position))}.to-transparent{--tw-gradient-to:transparent;--tw-gradient-stops:var(--tw-gradient-via-stops,var(--tw-gradient-position), var(--tw-gradient-from) var(--tw-gradient-from-position), var(--tw-gradient-to) var(--tw-gradient-to-position))}.fill-current{fill:currentColor}.fill-sky-400{fill:var(--color-sky-400)}.stroke-amber-600{stroke:var(--color-amber-600)}.stroke-primary\\/15{stroke:var(--primary)}@supports (color:color-mix(in lab, red, red)){.stroke-primary\\/15{stroke:color-mix(in oklab, var(--primary) 15%, transparent)}}.object-cover{object-fit:cover}.\\!p-0{padding:0!important}.\\!p-1{padding:var(--spacing)!important}.p-0{padding:0}.p-0\\.5{padding:calc(var(--spacing) * .5)}.p-1{padding:var(--spacing)}.p-1\\.5{padding:calc(var(--spacing) * 1.5)}.p-2{padding:calc(var(--spacing) * 2)}.p-2\\.5{padding:calc(var(--spacing) * 2.5)}.p-3{padding:calc(var(--spacing) * 3)}.p-4{padding:calc(var(--spacing) * 4)}.p-5{padding:calc(var(--spacing) * 5)}.p-6{padding:calc(var(--spacing) * 6)}.p-8{padding:calc(var(--spacing) * 8)}.p-10{padding:calc(var(--spacing) * 10)}.p-density-1{padding:var(--spacing-density-1)}.p-density-2{padding:var(--spacing-density-2)}.p-density-3{padding:var(--spacing-density-3)}.p-density-4{padding:var(--spacing-density-4)}.p-density-6{padding:var(--spacing-density-6)}.p-density-8{padding:var(--spacing-density-8)}.px-0{padding-inline:0}.px-0\\.5{padding-inline:calc(var(--spacing) * .5)}.px-1{padding-inline:var(--spacing)}.px-1\\.5{padding-inline:calc(var(--spacing) * 1.5)}.px-2{padding-inline:calc(var(--spacing) * 2)}.px-2\\.5{padding-inline:calc(var(--spacing) * 2.5)}.px-3{padding-inline:calc(var(--spacing) * 3)}.px-3\\.5{padding-inline:calc(var(--spacing) * 3.5)}.px-4{padding-inline:calc(var(--spacing) * 4)}.px-6{padding-inline:calc(var(--spacing) * 6)}.px-8{padding-inline:calc(var(--spacing) * 8)}.px-control-px{padding-inline:var(--spacing-control-px)}.px-density-1{padding-inline:var(--spacing-density-1)}.px-density-2{padding-inline:var(--spacing-density-2)}.px-density-3{padding-inline:var(--spacing-density-3)}.px-density-4{padding-inline:var(--spacing-density-4)}.py-0{padding-block:0}.py-0\\.5{padding-block:calc(var(--spacing) * .5)}.py-1{padding-block:var(--spacing)}.py-1\\.5{padding-block:calc(var(--spacing) * 1.5)}.py-2{padding-block:calc(var(--spacing) * 2)}.py-2\\.5{padding-block:calc(var(--spacing) * 2.5)}.py-3{padding-block:calc(var(--spacing) * 3)}.py-4{padding-block:calc(var(--spacing) * 4)}.py-5{padding-block:calc(var(--spacing) * 5)}.py-6{padding-block:calc(var(--spacing) * 6)}.py-8{padding-block:calc(var(--spacing) * 8)}.py-10{padding-block:calc(var(--spacing) * 10)}.py-\\[var\\(--json-view-row-space\\)\\]{padding-block:var(--json-view-row-space)}.py-density-1{padding-block:var(--spacing-density-1)}.py-density-2{padding-block:var(--spacing-density-2)}.py-density-3{padding-block:var(--spacing-density-3)}.py-density-4{padding-block:var(--spacing-density-4)}.py-px{padding-block:1px}.pt-0{padding-top:0}.pt-0\\.5{padding-top:calc(var(--spacing) * .5)}.pt-1{padding-top:var(--spacing)}.pt-1\\.5{padding-top:calc(var(--spacing) * 1.5)}.pt-2{padding-top:calc(var(--spacing) * 2)}.pt-3{padding-top:calc(var(--spacing) * 3)}.pt-3\\.5{padding-top:calc(var(--spacing) * 3.5)}.pt-4{padding-top:calc(var(--spacing) * 4)}.pt-\\[env\\(safe-area-inset-top\\)\\]{padding-top:env(safe-area-inset-top)}.pt-density-2{padding-top:var(--spacing-density-2)}.pt-density-3{padding-top:var(--spacing-density-3)}.pt-density-4{padding-top:var(--spacing-density-4)}.pt-px{padding-top:1px}.pr-0{padding-right:0}.pr-0\\.5{padding-right:calc(var(--spacing) * .5)}.pr-1{padding-right:var(--spacing)}.pr-2{padding-right:calc(var(--spacing) * 2)}.pr-3{padding-right:calc(var(--spacing) * 3)}.pr-4{padding-right:calc(var(--spacing) * 4)}.pr-5{padding-right:calc(var(--spacing) * 5)}.pr-6{padding-right:calc(var(--spacing) * 6)}.pr-7{padding-right:calc(var(--spacing) * 7)}.pr-8{padding-right:calc(var(--spacing) * 8)}.pr-9{padding-right:calc(var(--spacing) * 9)}.pr-10{padding-right:calc(var(--spacing) * 10)}.pr-12{padding-right:calc(var(--spacing) * 12)}.pr-14{padding-right:calc(var(--spacing) * 14)}.pr-20{padding-right:calc(var(--spacing) * 20)}.pr-28{padding-right:calc(var(--spacing) * 28)}.pr-\\[5\\.5rem\\]{padding-right:5.5rem}.pr-\\[env\\(safe-area-inset-right\\)\\]{padding-right:env(safe-area-inset-right)}.pr-density-3{padding-right:var(--spacing-density-3)}.pb-1{padding-bottom:var(--spacing)}.pb-1\\.5{padding-bottom:calc(var(--spacing) * 1.5)}.pb-2{padding-bottom:calc(var(--spacing) * 2)}.pb-3{padding-bottom:calc(var(--spacing) * 3)}.pb-4{padding-bottom:calc(var(--spacing) * 4)}.pb-10{padding-bottom:calc(var(--spacing) * 10)}.pb-\\[max\\(1rem\\,env\\(safe-area-inset-bottom\\)\\)\\]{padding-bottom:max(1rem, env(safe-area-inset-bottom))}.pb-\\[max\\(calc\\(var\\(--spacing-unit\\)\\*3\\)\\,env\\(safe-area-inset-bottom\\)\\)\\]{padding-bottom:max(calc(var(--spacing-unit) * 3), env(safe-area-inset-bottom))}.pb-density-2{padding-bottom:var(--spacing-density-2)}.pb-density-3{padding-bottom:var(--spacing-density-3)}.pb-density-4{padding-bottom:var(--spacing-density-4)}.pl-0{padding-left:0}.pl-1{padding-left:var(--spacing)}.pl-2{padding-left:calc(var(--spacing) * 2)}.pl-3{padding-left:calc(var(--spacing) * 3)}.pl-4{padding-left:calc(var(--spacing) * 4)}.pl-5{padding-left:calc(var(--spacing) * 5)}.pl-6{padding-left:calc(var(--spacing) * 6)}.pl-7{padding-left:calc(var(--spacing) * 7)}.pl-8{padding-left:calc(var(--spacing) * 8)}.pl-\\[env\\(safe-area-inset-left\\)\\]{padding-left:env(safe-area-inset-left)}.pl-\\[var\\(--json-view-indent\\)\\]{padding-left:var(--json-view-indent)}.pl-density-3{padding-left:var(--spacing-density-3)}.pl-px{padding-left:1px}.text-center{text-align:center}.text-end{text-align:end}.text-left{text-align:left}.text-right{text-align:right}.text-start{text-align:start}.align-\\[-0\\.15em\\]{vertical-align:-.15em}.align-\\[-0\\.125em\\]{vertical-align:-.125em}.align-baseline{vertical-align:baseline}.align-bottom{vertical-align:bottom}.align-middle{vertical-align:middle}.align-super{vertical-align:super}.align-top{vertical-align:top}.font-mono{font-family:var(--fs-font-mono)}.font-sans{font-family:var(--fs-font-sans)}.text-2xl{font-size:var(--text-2xl);line-height:var(--tw-leading,var(--text-2xl--line-height))}.text-3xl{font-size:var(--text-3xl);line-height:var(--tw-leading,var(--text-3xl--line-height))}.text-4xl{font-size:var(--text-4xl);line-height:var(--tw-leading,var(--text-4xl--line-height))}.text-base{font-size:var(--text-base);line-height:var(--tw-leading,var(--text-base--line-height))}.text-lg{font-size:var(--text-lg);line-height:var(--tw-leading,var(--text-lg--line-height))}.text-sm{font-size:var(--text-sm);line-height:var(--tw-leading,var(--text-sm--line-height))}.text-xl{font-size:var(--text-xl);line-height:var(--tw-leading,var(--text-xl--line-height))}.text-xs{font-size:var(--text-xs);line-height:var(--tw-leading,var(--text-xs--line-height))}.text-\\[0\\.6rem\\]{font-size:.6rem}.text-\\[0\\.7rem\\]{font-size:.7rem}.text-\\[0\\.9em\\]{font-size:.9em}.text-\\[0\\.65em\\]{font-size:.65em}.text-\\[0\\.65rem\\]{font-size:.65rem}.text-\\[0\\.75em\\]{font-size:.75em}.text-\\[0\\.85em\\]{font-size:.85em}.text-\\[1em\\]{font-size:1em}.text-\\[9\\.5px\\]{font-size:9.5px}.text-\\[9px\\]{font-size:9px}.text-\\[10px\\]{font-size:10px}.text-\\[11\\.5px\\]{font-size:11.5px}.text-\\[11px\\]{font-size:11px}.text-\\[12px\\]{font-size:12px}.text-\\[13px\\]{font-size:13px}.text-\\[14px\\]{font-size:14px}.text-\\[15px\\]{font-size:15px}.text-\\[length\\:inherit\\]{font-size:inherit}.leading-4{--tw-leading:calc(var(--spacing) * 4);line-height:calc(var(--spacing) * 4)}.leading-5{--tw-leading:calc(var(--spacing) * 5);line-height:calc(var(--spacing) * 5)}.leading-6{--tw-leading:calc(var(--spacing) * 6);line-height:calc(var(--spacing) * 6)}.leading-7{--tw-leading:calc(var(--spacing) * 7);line-height:calc(var(--spacing) * 7)}.leading-\\[8px\\]{--tw-leading:8px;line-height:8px}.leading-\\[inherit\\]{--tw-leading:inherit;line-height:inherit}.leading-none{--tw-leading:1;line-height:1}.leading-normal{--tw-leading:var(--leading-normal);line-height:var(--leading-normal)}.leading-relaxed{--tw-leading:var(--leading-relaxed);line-height:var(--leading-relaxed)}.leading-snug{--tw-leading:var(--leading-snug);line-height:var(--leading-snug)}.leading-tight{--tw-leading:var(--leading-tight);line-height:var(--leading-tight)}.font-bold{--tw-font-weight:var(--font-weight-bold);font-weight:var(--font-weight-bold)}.font-medium{--tw-font-weight:var(--font-weight-medium);font-weight:var(--font-weight-medium)}.font-normal{--tw-font-weight:var(--font-weight-normal);font-weight:var(--font-weight-normal)}.font-semibold{--tw-font-weight:var(--font-weight-semibold);font-weight:var(--font-weight-semibold)}.tracking-\\[-0\\.02em\\]{--tw-tracking:-.02em;letter-spacing:-.02em}.tracking-\\[-0\\.05em\\]{--tw-tracking:-.05em;letter-spacing:-.05em}.tracking-\\[0\\.2em\\]{--tw-tracking:.2em;letter-spacing:.2em}.tracking-\\[0\\.03em\\]{--tw-tracking:.03em;letter-spacing:.03em}.tracking-\\[0\\.12em\\]{--tw-tracking:.12em;letter-spacing:.12em}.tracking-\\[0\\.14em\\]{--tw-tracking:.14em;letter-spacing:.14em}.tracking-\\[0\\.18em\\]{--tw-tracking:.18em;letter-spacing:.18em}.tracking-\\[0\\.24em\\]{--tw-tracking:.24em;letter-spacing:.24em}.tracking-normal{--tw-tracking:var(--tracking-normal);letter-spacing:var(--tracking-normal)}.tracking-tight{--tw-tracking:var(--tracking-tight);letter-spacing:var(--tracking-tight)}.tracking-wide{--tw-tracking:var(--tracking-wide);letter-spacing:var(--tracking-wide)}.tracking-wider{--tw-tracking:var(--tracking-wider);letter-spacing:var(--tracking-wider)}.\\[overflow-wrap\\:anywhere\\]{overflow-wrap:anywhere}.break-words{overflow-wrap:break-word}.wrap-anywhere{overflow-wrap:anywhere}.break-all{word-break:break-all}.text-ellipsis{text-overflow:ellipsis}.whitespace-normal{white-space:normal}.whitespace-nowrap{white-space:nowrap}.whitespace-pre{white-space:pre}.whitespace-pre-line{white-space:pre-line}.whitespace-pre-wrap{white-space:pre-wrap}.text-\\[\\#0F766E\\]{color:#0f766e}.text-\\[\\#0e7490\\]{color:#0e7490}.text-\\[\\#00add8\\]{color:#00add8}.text-\\[\\#2ead33\\]{color:#2ead33}.text-\\[\\#4D6BFE\\]{color:#4d6bfe}.text-\\[\\#4b32c3\\]{color:#4b32c3}.text-\\[\\#6e9f18\\]{color:#6e9f18}.text-\\[\\#111\\]{color:#111}.text-\\[\\#0668E1\\]{color:#0668e1}.text-\\[\\#0891b2\\]{color:#0891b2}.text-\\[\\#3178c6\\]{color:#3178c6}.text-\\[\\#3776ab\\]{color:#3776ab}.text-\\[\\#4285F4\\]{color:#4285f4}.text-\\[\\#261230\\]{color:#261230}.text-\\[\\#A16207\\]{color:#a16207}.text-\\[\\#C15F3C\\]{color:#c15f3c}.text-\\[\\#C2410C\\]{color:#c2410c}.text-\\[\\#c21325\\]{color:#c21325}.text-\\[\\#fff\\]{color:#fff}.text-\\[hsl\\(0\\,0\\%\\,0\\%\\)\\]{color:#000}.text-\\[var\\(--fs-error\\)\\]{color:var(--fs-error)}.text-\\[var\\(--fs-info\\)\\]{color:var(--fs-info)}.text-\\[var\\(--fs-success\\)\\]{color:var(--fs-success)}.text-\\[var\\(--fs-warning\\)\\]{color:var(--fs-warning)}.text-accent-foreground{color:var(--accent-foreground)}.text-amber-50{color:var(--color-amber-50)}.text-amber-100{color:var(--color-amber-100)}.text-amber-200{color:var(--color-amber-200)}.text-amber-300{color:var(--color-amber-300)}.text-amber-400{color:var(--color-amber-400)}.text-amber-500{color:var(--color-amber-500)}.text-amber-600{color:var(--color-amber-600)}.text-amber-600\\/70{color:#dd7400b3}@supports (color:color-mix(in lab, red, red)){.text-amber-600\\/70{color:color-mix(in oklab, var(--color-amber-600) 70%, transparent)}}.text-amber-700{color:var(--color-amber-700)}.text-amber-800{color:var(--color-amber-800)}.text-amber-900{color:var(--color-amber-900)}.text-amber-950{color:var(--color-amber-950)}.text-amber-950\\/90{color:#461901e6}@supports (color:color-mix(in lab, red, red)){.text-amber-950\\/90{color:color-mix(in oklab, var(--color-amber-950) 90%, transparent)}}.text-background{color:var(--background)}.text-black{color:var(--color-black)}.text-blue-50{color:var(--color-blue-50)}.text-blue-100{color:var(--color-blue-100)}.text-blue-200{color:var(--color-blue-200)}.text-blue-300{color:var(--color-blue-300)}.text-blue-400{color:var(--color-blue-400)}.text-blue-500{color:var(--color-blue-500)}.text-blue-600{color:var(--color-blue-600)}.text-blue-600\\/70{color:#155dfcb3}@supports (color:color-mix(in lab, red, red)){.text-blue-600\\/70{color:color-mix(in oklab, var(--color-blue-600) 70%, transparent)}}.text-blue-700{color:var(--color-blue-700)}.text-blue-800{color:var(--color-blue-800)}.text-blue-900{color:var(--color-blue-900)}.text-blue-950{color:var(--color-blue-950)}.text-blue-950\\/90{color:#162456e6}@supports (color:color-mix(in lab, red, red)){.text-blue-950\\/90{color:color-mix(in oklab, var(--color-blue-950) 90%, transparent)}}.text-border{color:var(--border)}.text-card-foreground{color:var(--card-foreground)}.text-cyan-50{color:var(--color-cyan-50)}.text-cyan-100{color:var(--color-cyan-100)}.text-cyan-200{color:var(--color-cyan-200)}.text-cyan-300{color:var(--color-cyan-300)}.text-cyan-400{color:var(--color-cyan-400)}.text-cyan-500{color:var(--color-cyan-500)}.text-cyan-600{color:var(--color-cyan-600)}.text-cyan-700{color:var(--color-cyan-700)}.text-cyan-800{color:var(--color-cyan-800)}.text-cyan-900{color:var(--color-cyan-900)}.text-cyan-950{color:var(--color-cyan-950)}.text-destructive{color:var(--destructive)}.text-destructive-foreground{color:var(--destructive-foreground)}.text-emerald-50{color:var(--color-emerald-50)}.text-emerald-100{color:var(--color-emerald-100)}.text-emerald-200{color:var(--color-emerald-200)}.text-emerald-300{color:var(--color-emerald-300)}.text-emerald-400{color:var(--color-emerald-400)}.text-emerald-500{color:var(--color-emerald-500)}.text-emerald-600{color:var(--color-emerald-600)}.text-emerald-600\\/70{color:#009767b3}@supports (color:color-mix(in lab, red, red)){.text-emerald-600\\/70{color:color-mix(in oklab, var(--color-emerald-600) 70%, transparent)}}.text-emerald-700{color:var(--color-emerald-700)}.text-emerald-800{color:var(--color-emerald-800)}.text-emerald-900{color:var(--color-emerald-900)}.text-emerald-950{color:var(--color-emerald-950)}.text-emerald-950\\/90{color:#002c22e6}@supports (color:color-mix(in lab, red, red)){.text-emerald-950\\/90{color:color-mix(in oklab, var(--color-emerald-950) 90%, transparent)}}.text-foreground,.text-foreground\\/60{color:var(--foreground)}@supports (color:color-mix(in lab, red, red)){.text-foreground\\/60{color:color-mix(in oklab, var(--foreground) 60%, transparent)}}.text-foreground\\/70{color:var(--foreground)}@supports (color:color-mix(in lab, red, red)){.text-foreground\\/70{color:color-mix(in oklab, var(--foreground) 70%, transparent)}}.text-foreground\\/80{color:var(--foreground)}@supports (color:color-mix(in lab, red, red)){.text-foreground\\/80{color:color-mix(in oklab, var(--foreground) 80%, transparent)}}.text-fuchsia-50{color:var(--color-fuchsia-50)}.text-fuchsia-100{color:var(--color-fuchsia-100)}.text-fuchsia-200{color:var(--color-fuchsia-200)}.text-fuchsia-300{color:var(--color-fuchsia-300)}.text-fuchsia-400{color:var(--color-fuchsia-400)}.text-fuchsia-500{color:var(--color-fuchsia-500)}.text-fuchsia-600{color:var(--color-fuchsia-600)}.text-fuchsia-700{color:var(--color-fuchsia-700)}.text-fuchsia-800{color:var(--color-fuchsia-800)}.text-fuchsia-900{color:var(--color-fuchsia-900)}.text-fuchsia-950{color:var(--color-fuchsia-950)}.text-gray-50{color:var(--color-gray-50)}.text-gray-100{color:var(--color-gray-100)}.text-gray-200{color:var(--color-gray-200)}.text-gray-300{color:var(--color-gray-300)}.text-gray-400{color:var(--color-gray-400)}.text-gray-500{color:var(--color-gray-500)}.text-gray-600{color:var(--color-gray-600)}.text-gray-700{color:var(--color-gray-700)}.text-gray-800{color:var(--color-gray-800)}.text-gray-900{color:var(--color-gray-900)}.text-gray-950{color:var(--color-gray-950)}.text-green-50{color:var(--color-green-50)}.text-green-100{color:var(--color-green-100)}.text-green-200{color:var(--color-green-200)}.text-green-300{color:var(--color-green-300)}.text-green-400{color:var(--color-green-400)}.text-green-500{color:var(--color-green-500)}.text-green-600{color:var(--color-green-600)}.text-green-700{color:var(--color-green-700)}.text-green-800{color:var(--color-green-800)}.text-green-900{color:var(--color-green-900)}.text-green-950{color:var(--color-green-950)}.text-indigo-50{color:var(--color-indigo-50)}.text-indigo-100{color:var(--color-indigo-100)}.text-indigo-200{color:var(--color-indigo-200)}.text-indigo-300{color:var(--color-indigo-300)}.text-indigo-400{color:var(--color-indigo-400)}.text-indigo-500{color:var(--color-indigo-500)}.text-indigo-600{color:var(--color-indigo-600)}.text-indigo-700{color:var(--color-indigo-700)}.text-indigo-800{color:var(--color-indigo-800)}.text-indigo-900{color:var(--color-indigo-900)}.text-indigo-950{color:var(--color-indigo-950)}.text-inherit{color:inherit}.text-lime-50{color:var(--color-lime-50)}.text-lime-100{color:var(--color-lime-100)}.text-lime-200{color:var(--color-lime-200)}.text-lime-300{color:var(--color-lime-300)}.text-lime-400{color:var(--color-lime-400)}.text-lime-500{color:var(--color-lime-500)}.text-lime-600{color:var(--color-lime-600)}.text-lime-700{color:var(--color-lime-700)}.text-lime-800{color:var(--color-lime-800)}.text-lime-900{color:var(--color-lime-900)}.text-lime-950{color:var(--color-lime-950)}.text-muted{color:var(--muted)}.text-muted-foreground,.text-muted-foreground\\/20{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/20{color:color-mix(in oklab, var(--muted-foreground) 20%, transparent)}}.text-muted-foreground\\/45{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/45{color:color-mix(in oklab, var(--muted-foreground) 45%, transparent)}}.text-muted-foreground\\/50{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/50{color:color-mix(in oklab, var(--muted-foreground) 50%, transparent)}}.text-muted-foreground\\/60{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/60{color:color-mix(in oklab, var(--muted-foreground) 60%, transparent)}}.text-muted-foreground\\/70{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/70{color:color-mix(in oklab, var(--muted-foreground) 70%, transparent)}}.text-muted-foreground\\/80{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/80{color:color-mix(in oklab, var(--muted-foreground) 80%, transparent)}}.text-neutral-50{color:var(--color-neutral-50)}.text-neutral-100{color:var(--color-neutral-100)}.text-neutral-200{color:var(--color-neutral-200)}.text-neutral-300{color:var(--color-neutral-300)}.text-neutral-400{color:var(--color-neutral-400)}.text-neutral-500{color:var(--color-neutral-500)}.text-neutral-600{color:var(--color-neutral-600)}.text-neutral-700{color:var(--color-neutral-700)}.text-neutral-800{color:var(--color-neutral-800)}.text-neutral-900{color:var(--color-neutral-900)}.text-neutral-950{color:var(--color-neutral-950)}.text-orange-50{color:var(--color-orange-50)}.text-orange-100{color:var(--color-orange-100)}.text-orange-200{color:var(--color-orange-200)}.text-orange-300{color:var(--color-orange-300)}.text-orange-400{color:var(--color-orange-400)}.text-orange-500{color:var(--color-orange-500)}.text-orange-600{color:var(--color-orange-600)}.text-orange-700{color:var(--color-orange-700)}.text-orange-800{color:var(--color-orange-800)}.text-orange-900{color:var(--color-orange-900)}.text-orange-950{color:var(--color-orange-950)}.text-pink-50{color:var(--color-pink-50)}.text-pink-100{color:var(--color-pink-100)}.text-pink-200{color:var(--color-pink-200)}.text-pink-300{color:var(--color-pink-300)}.text-pink-400{color:var(--color-pink-400)}.text-pink-500{color:var(--color-pink-500)}.text-pink-600{color:var(--color-pink-600)}.text-pink-700{color:var(--color-pink-700)}.text-pink-800{color:var(--color-pink-800)}.text-pink-900{color:var(--color-pink-900)}.text-pink-950{color:var(--color-pink-950)}.text-placeholder{color:var(--placeholder)}.text-popover-foreground{color:var(--popover-foreground)}.text-primary{color:var(--primary)}.text-primary-foreground{color:var(--primary-foreground)}.text-primary\\/70{color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.text-primary\\/70{color:color-mix(in oklab, var(--primary) 70%, transparent)}}.text-purple-50{color:var(--color-purple-50)}.text-purple-100{color:var(--color-purple-100)}.text-purple-200{color:var(--color-purple-200)}.text-purple-300{color:var(--color-purple-300)}.text-purple-400{color:var(--color-purple-400)}.text-purple-500{color:var(--color-purple-500)}.text-purple-600{color:var(--color-purple-600)}.text-purple-600\\/70{color:#9810fab3}@supports (color:color-mix(in lab, red, red)){.text-purple-600\\/70{color:color-mix(in oklab, var(--color-purple-600) 70%, transparent)}}.text-purple-700{color:var(--color-purple-700)}.text-purple-800{color:var(--color-purple-800)}.text-purple-900{color:var(--color-purple-900)}.text-purple-950{color:var(--color-purple-950)}.text-purple-950\\/90{color:#3c0366e6}@supports (color:color-mix(in lab, red, red)){.text-purple-950\\/90{color:color-mix(in oklab, var(--color-purple-950) 90%, transparent)}}.text-red-50{color:var(--color-red-50)}.text-red-100{color:var(--color-red-100)}.text-red-200{color:var(--color-red-200)}.text-red-300{color:var(--color-red-300)}.text-red-400{color:var(--color-red-400)}.text-red-500{color:var(--color-red-500)}.text-red-600{color:var(--color-red-600)}.text-red-600\\/70{color:#e40014b3}@supports (color:color-mix(in lab, red, red)){.text-red-600\\/70{color:color-mix(in oklab, var(--color-red-600) 70%, transparent)}}.text-red-700{color:var(--color-red-700)}.text-red-800{color:var(--color-red-800)}.text-red-900{color:var(--color-red-900)}.text-red-950{color:var(--color-red-950)}.text-red-950\\/90{color:#460809e6}@supports (color:color-mix(in lab, red, red)){.text-red-950\\/90{color:color-mix(in oklab, var(--color-red-950) 90%, transparent)}}.text-rose-50{color:var(--color-rose-50)}.text-rose-100{color:var(--color-rose-100)}.text-rose-200{color:var(--color-rose-200)}.text-rose-300{color:var(--color-rose-300)}.text-rose-400{color:var(--color-rose-400)}.text-rose-500{color:var(--color-rose-500)}.text-rose-600{color:var(--color-rose-600)}.text-rose-700{color:var(--color-rose-700)}.text-rose-800{color:var(--color-rose-800)}.text-rose-900{color:var(--color-rose-900)}.text-rose-950{color:var(--color-rose-950)}.text-secondary-foreground{color:var(--secondary-foreground)}.text-sidebar-foreground,.text-sidebar-foreground\\/55{color:var(--sidebar-foreground)}@supports (color:color-mix(in lab, red, red)){.text-sidebar-foreground\\/55{color:color-mix(in oklab, var(--sidebar-foreground) 55%, transparent)}}.text-sidebar-foreground\\/60{color:var(--sidebar-foreground)}@supports (color:color-mix(in lab, red, red)){.text-sidebar-foreground\\/60{color:color-mix(in oklab, var(--sidebar-foreground) 60%, transparent)}}.text-sidebar-foreground\\/70{color:var(--sidebar-foreground)}@supports (color:color-mix(in lab, red, red)){.text-sidebar-foreground\\/70{color:color-mix(in oklab, var(--sidebar-foreground) 70%, transparent)}}.text-sidebar-primary{color:var(--sidebar-primary)}.text-sky-50{color:var(--color-sky-50)}.text-sky-100{color:var(--color-sky-100)}.text-sky-200{color:var(--color-sky-200)}.text-sky-300{color:var(--color-sky-300)}.text-sky-400{color:var(--color-sky-400)}.text-sky-500{color:var(--color-sky-500)}.text-sky-600{color:var(--color-sky-600)}.text-sky-700{color:var(--color-sky-700)}.text-sky-800{color:var(--color-sky-800)}.text-sky-900{color:var(--color-sky-900)}.text-sky-950{color:var(--color-sky-950)}.text-slate-50{color:var(--color-slate-50)}.text-slate-100{color:var(--color-slate-100)}.text-slate-200{color:var(--color-slate-200)}.text-slate-300{color:var(--color-slate-300)}.text-slate-400{color:var(--color-slate-400)}.text-slate-500{color:var(--color-slate-500)}.text-slate-600{color:var(--color-slate-600)}.text-slate-700{color:var(--color-slate-700)}.text-slate-800{color:var(--color-slate-800)}.text-slate-900{color:var(--color-slate-900)}.text-slate-950{color:var(--color-slate-950)}.text-stone-50{color:var(--color-stone-50)}.text-stone-100{color:var(--color-stone-100)}.text-stone-200{color:var(--color-stone-200)}.text-stone-300{color:var(--color-stone-300)}.text-stone-400{color:var(--color-stone-400)}.text-stone-500{color:var(--color-stone-500)}.text-stone-600{color:var(--color-stone-600)}.text-stone-700{color:var(--color-stone-700)}.text-stone-800{color:var(--color-stone-800)}.text-stone-900{color:var(--color-stone-900)}.text-stone-950{color:var(--color-stone-950)}.text-teal-50{color:var(--color-teal-50)}.text-teal-100{color:var(--color-teal-100)}.text-teal-200{color:var(--color-teal-200)}.text-teal-300{color:var(--color-teal-300)}.text-teal-400{color:var(--color-teal-400)}.text-teal-500{color:var(--color-teal-500)}.text-teal-600{color:var(--color-teal-600)}.text-teal-700{color:var(--color-teal-700)}.text-teal-800{color:var(--color-teal-800)}.text-teal-900{color:var(--color-teal-900)}.text-teal-950{color:var(--color-teal-950)}.text-violet-50{color:var(--color-violet-50)}.text-violet-100{color:var(--color-violet-100)}.text-violet-200{color:var(--color-violet-200)}.text-violet-300{color:var(--color-violet-300)}.text-violet-400{color:var(--color-violet-400)}.text-violet-500{color:var(--color-violet-500)}.text-violet-600{color:var(--color-violet-600)}.text-violet-700{color:var(--color-violet-700)}.text-violet-800{color:var(--color-violet-800)}.text-violet-900{color:var(--color-violet-900)}.text-violet-950{color:var(--color-violet-950)}.text-white{color:var(--color-white)}.text-yellow-50{color:var(--color-yellow-50)}.text-yellow-100{color:var(--color-yellow-100)}.text-yellow-200{color:var(--color-yellow-200)}.text-yellow-300{color:var(--color-yellow-300)}.text-yellow-400{color:var(--color-yellow-400)}.text-yellow-500{color:var(--color-yellow-500)}.text-yellow-600{color:var(--color-yellow-600)}.text-yellow-700{color:var(--color-yellow-700)}.text-yellow-800{color:var(--color-yellow-800)}.text-yellow-900{color:var(--color-yellow-900)}.text-yellow-950{color:var(--color-yellow-950)}.text-zinc-50{color:var(--color-zinc-50)}.text-zinc-100{color:var(--color-zinc-100)}.text-zinc-200{color:var(--color-zinc-200)}.text-zinc-300{color:var(--color-zinc-300)}.text-zinc-400{color:var(--color-zinc-400)}.text-zinc-500{color:var(--color-zinc-500)}.text-zinc-600{color:var(--color-zinc-600)}.text-zinc-700{color:var(--color-zinc-700)}.text-zinc-800{color:var(--color-zinc-800)}.text-zinc-900{color:var(--color-zinc-900)}.text-zinc-950{color:var(--color-zinc-950)}.capitalize{text-transform:capitalize}.lowercase{text-transform:lowercase}.normal-case{text-transform:none}.uppercase{text-transform:uppercase}.italic{font-style:italic}.ordinal{--tw-ordinal:ordinal;font-variant-numeric:var(--tw-ordinal,) var(--tw-slashed-zero,) var(--tw-numeric-figure,) var(--tw-numeric-spacing,) var(--tw-numeric-fraction,)}.tabular-nums{--tw-numeric-spacing:tabular-nums;font-variant-numeric:var(--tw-ordinal,) var(--tw-slashed-zero,) var(--tw-numeric-figure,) var(--tw-numeric-spacing,) var(--tw-numeric-fraction,)}.line-through{text-decoration-line:line-through}.underline{text-decoration-line:underline}.underline-offset-2{text-underline-offset:2px}.underline-offset-4{text-underline-offset:4px}.accent-primary{accent-color:var(--primary)}.opacity-0{opacity:0}.opacity-10{opacity:.1}.opacity-20{opacity:.2}.opacity-30{opacity:.3}.opacity-40{opacity:.4}.opacity-50{opacity:.5}.opacity-60{opacity:.6}.opacity-70{opacity:.7}.opacity-75{opacity:.75}.opacity-80{opacity:.8}.opacity-85{opacity:.85}.opacity-90{opacity:.9}.opacity-100{opacity:1}.shadow{--tw-shadow:0 1px 3px 0 var(--tw-shadow-color,#0000001a), 0 1px 2px -1px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-\\[0_1px_0_0_var\\(--tw-shadow-color\\)\\]{--tw-shadow:0 1px 0 0 var(--tw-shadow-color,var(--tw-shadow-color));box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-\\[0_4px_16px_rgba\\(0\\,0\\,0\\,0\\.12\\)\\]{--tw-shadow:0 4px 16px var(--tw-shadow-color,#0000001f);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-\\[inset_0_-2px_0_0_var\\(--color-primary\\)\\]{--tw-shadow:inset 0 -2px 0 0 var(--tw-shadow-color,var(--color-primary));box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-\\[inset_0_2px_0_0_var\\(--color-primary\\)\\]{--tw-shadow:inset 0 2px 0 0 var(--tw-shadow-color,var(--color-primary));box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-lg{--tw-shadow:0 10px 15px -3px var(--tw-shadow-color,#0000001a), 0 4px 6px -4px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-md{--tw-shadow:0 4px 6px -1px var(--tw-shadow-color,#0000001a), 0 2px 4px -2px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-none{--tw-shadow:0 0 #0000;box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-sm{--tw-shadow:0 1px 3px 0 var(--tw-shadow-color,#0000001a), 0 1px 2px -1px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-xl{--tw-shadow:0 20px 25px -5px var(--tw-shadow-color,#0000001a), 0 8px 10px -6px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.ring,.ring-1{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.ring-2{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.inset-shadow-sm{--tw-inset-shadow:inset 0 2px 4px var(--tw-inset-shadow-color,#0000000d);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-black\\/5{--tw-shadow-color:#0000000d}@supports (color:color-mix(in lab, red, red)){.shadow-black\\/5{--tw-shadow-color:color-mix(in oklab, color-mix(in oklab, var(--color-black) 5%, transparent) var(--tw-shadow-alpha), transparent)}}.shadow-black\\/10{--tw-shadow-color:#0000001a}@supports (color:color-mix(in lab, red, red)){.shadow-black\\/10{--tw-shadow-color:color-mix(in oklab, color-mix(in oklab, var(--color-black) 10%, transparent) var(--tw-shadow-alpha), transparent)}}.shadow-border{--tw-shadow-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.shadow-border{--tw-shadow-color:color-mix(in oklab, var(--border) var(--tw-shadow-alpha), transparent)}}.ring-amber-200\\/80{--tw-ring-color:#fee685cc}@supports (color:color-mix(in lab, red, red)){.ring-amber-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-amber-200) 80%, transparent)}}.ring-amber-300\\/60{--tw-ring-color:#ffd23699}@supports (color:color-mix(in lab, red, red)){.ring-amber-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-amber-300) 60%, transparent)}}.ring-black\\/10{--tw-ring-color:#0000001a}@supports (color:color-mix(in lab, red, red)){.ring-black\\/10{--tw-ring-color:color-mix(in oklab, var(--color-black) 10%, transparent)}}.ring-border,.ring-border\\/60{--tw-ring-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.ring-border\\/60{--tw-ring-color:color-mix(in oklab, var(--border) 60%, transparent)}}.ring-emerald-200\\/80{--tw-ring-color:#a4f4cfcc}@supports (color:color-mix(in lab, red, red)){.ring-emerald-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-emerald-200) 80%, transparent)}}.ring-emerald-300\\/60{--tw-ring-color:#5ee9b599}@supports (color:color-mix(in lab, red, red)){.ring-emerald-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-emerald-300) 60%, transparent)}}.ring-indigo-300\\/60{--tw-ring-color:#a4b3ff99}@supports (color:color-mix(in lab, red, red)){.ring-indigo-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-indigo-300) 60%, transparent)}}.ring-primary,.ring-primary\\/20{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.ring-primary\\/20{--tw-ring-color:color-mix(in oklab, var(--primary) 20%, transparent)}}.ring-primary\\/25{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.ring-primary\\/25{--tw-ring-color:color-mix(in oklab, var(--primary) 25%, transparent)}}.ring-primary\\/40{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.ring-primary\\/40{--tw-ring-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.ring-ring{--tw-ring-color:var(--ring)}.ring-rose-200\\/80{--tw-ring-color:#ffccd3cc}@supports (color:color-mix(in lab, red, red)){.ring-rose-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-rose-200) 80%, transparent)}}.ring-rose-300\\/60{--tw-ring-color:#ffa2ae99}@supports (color:color-mix(in lab, red, red)){.ring-rose-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-rose-300) 60%, transparent)}}.ring-sky-200\\/80{--tw-ring-color:#b8e6fecc}@supports (color:color-mix(in lab, red, red)){.ring-sky-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-sky-200) 80%, transparent)}}.ring-sky-300\\/60{--tw-ring-color:#77d4ff99}@supports (color:color-mix(in lab, red, red)){.ring-sky-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-sky-300) 60%, transparent)}}.ring-slate-200\\/80{--tw-ring-color:#e2e8f0cc}@supports (color:color-mix(in lab, red, red)){.ring-slate-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-slate-200) 80%, transparent)}}.ring-slate-300\\/60{--tw-ring-color:#cad5e299}@supports (color:color-mix(in lab, red, red)){.ring-slate-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-slate-300) 60%, transparent)}}.ring-teal-300\\/60{--tw-ring-color:#46ecd599}@supports (color:color-mix(in lab, red, red)){.ring-teal-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-teal-300) 60%, transparent)}}.ring-violet-200\\/80{--tw-ring-color:#ddd6ffcc}@supports (color:color-mix(in lab, red, red)){.ring-violet-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-violet-200) 80%, transparent)}}.ring-violet-300\\/60{--tw-ring-color:#c4b4ff99}@supports (color:color-mix(in lab, red, red)){.ring-violet-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-violet-300) 60%, transparent)}}.ring-offset-1{--tw-ring-offset-width:1px;--tw-ring-offset-shadow:var(--tw-ring-inset,) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)}.ring-offset-2{--tw-ring-offset-width:2px;--tw-ring-offset-shadow:var(--tw-ring-inset,) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)}.ring-offset-background{--tw-ring-offset-color:var(--background)}.outline,.outline-1{outline-style:var(--tw-outline-style);outline-width:1px}.-outline-offset-1{outline-offset:calc(1px * -1)}.outline-offset-2{outline-offset:2px}.outline-muted-foreground\\/60{outline-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.outline-muted-foreground\\/60{outline-color:color-mix(in oklab, var(--muted-foreground) 60%, transparent)}}.outline-primary\\/60{outline-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.outline-primary\\/60{outline-color:color-mix(in oklab, var(--primary) 60%, transparent)}}.outline-rose-500{outline-color:var(--color-rose-500)}.blur{--tw-blur:blur(8px);filter:var(--tw-blur,) var(--tw-brightness,) var(--tw-contrast,) var(--tw-grayscale,) var(--tw-hue-rotate,) var(--tw-invert,) var(--tw-saturate,) var(--tw-sepia,) var(--tw-drop-shadow,)}.invert{--tw-invert:invert(100%);filter:var(--tw-blur,) var(--tw-brightness,) var(--tw-contrast,) var(--tw-grayscale,) var(--tw-hue-rotate,) var(--tw-invert,) var(--tw-saturate,) var(--tw-sepia,) var(--tw-drop-shadow,)}.filter{filter:var(--tw-blur,) var(--tw-brightness,) var(--tw-contrast,) var(--tw-grayscale,) var(--tw-hue-rotate,) var(--tw-invert,) var(--tw-saturate,) var(--tw-sepia,) var(--tw-drop-shadow,)}.backdrop-blur{--tw-backdrop-blur:blur(8px);-webkit-backdrop-filter:var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-opacity,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,);backdrop-filter:var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-opacity,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,)}.backdrop-blur-sm{--tw-backdrop-blur:blur(var(--blur-sm));-webkit-backdrop-filter:var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-opacity,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,);backdrop-filter:var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-opacity,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,)}.transition{transition-property:color,background-color,border-color,outline-color,text-decoration-color,fill,stroke,--tw-gradient-from,--tw-gradient-via,--tw-gradient-to,opacity,box-shadow,transform,translate,scale,rotate,filter,-webkit-backdrop-filter,backdrop-filter,display,content-visibility,overlay,pointer-events;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[border-color\\,box-shadow\\]{transition-property:border-color,box-shadow;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[clip-path\\]{transition-property:clip-path;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[height\\]{transition-property:height;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[stroke-dashoffset\\]{transition-property:stroke-dashoffset;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[top\\,left\\,width\\,transform\\,border-radius\\,box-shadow\\]{transition-property:top,left,width,transform,border-radius,box-shadow;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[top\\,transform\\]{transition-property:top,transform;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[width\\]{transition-property:width;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-all{transition-property:all;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-colors{transition-property:color,background-color,border-color,outline-color,text-decoration-color,fill,stroke,--tw-gradient-from,--tw-gradient-via,--tw-gradient-to;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-opacity{transition-property:opacity;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-shadow{transition-property:box-shadow;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-transform{transition-property:transform,translate,scale,rotate;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.duration-150{--tw-duration:.15s;transition-duration:.15s}.duration-200{--tw-duration:.2s;transition-duration:.2s}.duration-300{--tw-duration:.3s;transition-duration:.3s}.duration-500{--tw-duration:.5s;transition-duration:.5s}.ease-out{--tw-ease:var(--ease-out);transition-timing-function:var(--ease-out)}.outline-dashed{--tw-outline-style:dashed;outline-style:dashed}.outline-none{--tw-outline-style:none;outline-style:none}.select-all{-webkit-user-select:all;user-select:all}.select-none{-webkit-user-select:none;user-select:none}.\\[--json-view-indent\\:0\\.75rem\\]{--json-view-indent:.75rem}.\\[--json-view-indent\\:1\\.25rem\\]{--json-view-indent:1.25rem}.\\[--json-view-indent\\:1rem\\]{--json-view-indent:1rem}.\\[--json-view-row-space\\:0px\\]{--json-view-row-space:0px}.\\[--json-view-row-space\\:1px\\]{--json-view-row-space:1px}.\\[--json-view-row-space\\:2px\\]{--json-view-row-space:2px}.\\[-webkit-box-orient\\:vertical\\]{-webkit-box-orient:vertical}.\\[-webkit-line-clamp\\:3\\]{-webkit-line-clamp:3}.\\[-webkit-user-drag\\:none\\]{-webkit-user-drag:none}.\\[profile\\:pod-cpu\\]{profile:pod-cpu}.\\[rows\\:0\\]{rows:0}.\\[rows\\:1\\]{rows:1}.\\[rows\\:25\\]{rows:25}.\\[text-shadow\\:0_0_2px_rgb\\(0_0_0\\/0\\.45\\)\\]{text-shadow:0 0 2px #00000073}.\\[writing-mode\\:vertical-rl\\]{writing-mode:vertical-rl}.ring-inset{--tw-ring-inset:inset}.group-open\\:hidden:is(:where(.group):is([open],:popover-open,:open) *){display:none}.group-open\\:inline:is(:where(.group):is([open],:popover-open,:open) *){display:inline}.group-open\\:rotate-90:is(:where(.group):is([open],:popover-open,:open) *){rotate:90deg}@media (hover:hover){.group-hover\\:text-foreground:is(:where(.group):hover *){color:var(--foreground)}.group-hover\\:opacity-100:is(:where(.group):hover *){opacity:1}.group-hover\\/add\\:border-primary:is(:where(.group\\/add):hover *){border-color:var(--primary)}.group-hover\\/add\\:text-primary:is(:where(.group\\/add):hover *){color:var(--primary)}.group-hover\\/cell-actions\\:opacity-95:is(:where(.group\\/cell-actions):hover *){opacity:.95}.group-hover\\/comment\\:opacity-70:is(:where(.group\\/comment):hover *){opacity:.7}.group-hover\\/header\\:bg-primary\\/70:is(:where(.group\\/header):hover *){background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.group-hover\\/header\\:bg-primary\\/70:is(:where(.group\\/header):hover *){background-color:color-mix(in oklab, var(--primary) 70%, transparent)}}.group-hover\\/thread\\:opacity-100:is(:where(.group\\/thread):hover *){opacity:1}}.placeholder\\:text-placeholder::placeholder{color:var(--placeholder)}.before\\:invisible:before{content:var(--tw-content);visibility:hidden}.before\\:content-\\[\\'\\.\\'\\]:before{--tw-content:\".\";content:var(--tw-content)}.after\\:absolute:after{content:var(--tw-content);position:absolute}.after\\:inset-0:after{content:var(--tw-content);inset:0}.after\\:content-\\[\\'\\'\\]:after{--tw-content:\"\";content:var(--tw-content)}.first\\:mt-1:first-child{margin-top:var(--spacing)}.first\\:mt-2:first-child{margin-top:calc(var(--spacing) * 2)}.first\\:border-t-0:first-child{border-top-style:var(--tw-border-style);border-top-width:0}.first\\:border-l-0:first-child{border-left-style:var(--tw-border-style);border-left-width:0}.first\\:pt-0:first-child{padding-top:0}.last\\:border-0:last-child{border-style:var(--tw-border-style);border-width:0}.last\\:border-b-0:last-child{border-bottom-style:var(--tw-border-style);border-bottom-width:0}.last\\:pb-0:last-child{padding-bottom:0}.even\\:bg-muted\\/40:nth-child(2n){background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.even\\:bg-muted\\/40:nth-child(2n){background-color:color-mix(in oklab, var(--muted) 40%, transparent)}}.open\\:w-full:is([open],:popover-open,:open){width:100%}.open\\:rounded-lg:is([open],:popover-open,:open){border-radius:var(--radius-lg)}.empty\\:hidden:empty{display:none}.focus-within\\:border-primary\\/40:focus-within{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus-within\\:border-primary\\/40:focus-within{border-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.focus-within\\:opacity-100:focus-within{opacity:1}.focus-within\\:shadow-md:focus-within{--tw-shadow:0 4px 6px -1px var(--tw-shadow-color,#0000001a), 0 2px 4px -2px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus-within\\:ring-0:focus-within{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(0px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus-within\\:ring-2:focus-within{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus-within\\:ring-destructive:focus-within{--tw-ring-color:var(--destructive)}.focus-within\\:ring-primary\\/30:focus-within{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus-within\\:ring-primary\\/30:focus-within{--tw-ring-color:color-mix(in oklab, var(--primary) 30%, transparent)}}.focus-within\\:ring-ring:focus-within{--tw-ring-color:var(--ring)}.focus-within\\:ring-offset-1:focus-within{--tw-ring-offset-width:1px;--tw-ring-offset-shadow:var(--tw-ring-inset,) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)}@media (hover:hover){.hover\\:z-\\[999\\]:hover{z-index:999}.hover\\:border-\\[\\#abc\\]:hover{border-color:#abc}.hover\\:border-border:hover{border-color:var(--border)}.hover\\:border-emerald-200:hover{border-color:var(--color-emerald-200)}.hover\\:border-foreground\\/20:hover{border-color:var(--foreground)}@supports (color:color-mix(in lab, red, red)){.hover\\:border-foreground\\/20:hover{border-color:color-mix(in oklab, var(--foreground) 20%, transparent)}}.hover\\:border-input:hover{border-color:var(--input)}.hover\\:border-muted-foreground\\/40:hover{border-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.hover\\:border-muted-foreground\\/40:hover{border-color:color-mix(in oklab, var(--muted-foreground) 40%, transparent)}}.hover\\:border-primary:hover,.hover\\:border-primary\\/30:hover{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:border-primary\\/30:hover{border-color:color-mix(in oklab, var(--primary) 30%, transparent)}}.hover\\:border-primary\\/40:hover{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:border-primary\\/40:hover{border-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.hover\\:border-primary\\/50:hover{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:border-primary\\/50:hover{border-color:color-mix(in oklab, var(--primary) 50%, transparent)}}.hover\\:bg-accent:hover,.hover\\:bg-accent\\/40:hover{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-accent\\/40:hover{background-color:color-mix(in oklab, var(--accent) 40%, transparent)}}.hover\\:bg-accent\\/50:hover{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-accent\\/50:hover{background-color:color-mix(in oklab, var(--accent) 50%, transparent)}}.hover\\:bg-amber-500\\/20:hover{background-color:#f99c0033}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-amber-500\\/20:hover{background-color:color-mix(in oklab, var(--color-amber-500) 20%, transparent)}}.hover\\:bg-background:hover,.hover\\:bg-background\\/60:hover{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-background\\/60:hover{background-color:color-mix(in oklab, var(--background) 60%, transparent)}}.hover\\:bg-background\\/70:hover{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-background\\/70:hover{background-color:color-mix(in oklab, var(--background) 70%, transparent)}}.hover\\:bg-border:hover{background-color:var(--border)}.hover\\:bg-card:hover{background-color:var(--card)}.hover\\:bg-destructive\\/10:hover{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-destructive\\/10:hover{background-color:color-mix(in oklab, var(--destructive) 10%, transparent)}}.hover\\:bg-destructive\\/90:hover{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-destructive\\/90:hover{background-color:color-mix(in oklab, var(--destructive) 90%, transparent)}}.hover\\:bg-emerald-50\\/50:hover{background-color:#ecfdf580}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-emerald-50\\/50:hover{background-color:color-mix(in oklab, var(--color-emerald-50) 50%, transparent)}}.hover\\:bg-muted:hover,.hover\\:bg-muted\\/30:hover{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-muted\\/30:hover{background-color:color-mix(in oklab, var(--muted) 30%, transparent)}}.hover\\:bg-muted\\/40:hover{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-muted\\/40:hover{background-color:color-mix(in oklab, var(--muted) 40%, transparent)}}.hover\\:bg-muted\\/50:hover{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-muted\\/50:hover{background-color:color-mix(in oklab, var(--muted) 50%, transparent)}}.hover\\:bg-muted\\/60:hover{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-muted\\/60:hover{background-color:color-mix(in oklab, var(--muted) 60%, transparent)}}.hover\\:bg-muted\\/70:hover{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-muted\\/70:hover{background-color:color-mix(in oklab, var(--muted) 70%, transparent)}}.hover\\:bg-primary:hover,.hover\\:bg-primary\\/10:hover{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-primary\\/10:hover{background-color:color-mix(in oklab, var(--primary) 10%, transparent)}}.hover\\:bg-primary\\/15:hover{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-primary\\/15:hover{background-color:color-mix(in oklab, var(--primary) 15%, transparent)}}.hover\\:bg-primary\\/40:hover{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-primary\\/40:hover{background-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.hover\\:bg-primary\\/90:hover{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-primary\\/90:hover{background-color:color-mix(in oklab, var(--primary) 90%, transparent)}}.hover\\:bg-secondary\\/80:hover{background-color:var(--secondary)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-secondary\\/80:hover{background-color:color-mix(in oklab, var(--secondary) 80%, transparent)}}.hover\\:bg-sidebar-accent:hover{background-color:var(--sidebar-accent)}.hover\\:from-primary\\/20:hover{--tw-gradient-from:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:from-primary\\/20:hover{--tw-gradient-from:color-mix(in oklab, var(--primary) 20%, transparent)}}.hover\\:from-primary\\/20:hover{--tw-gradient-stops:var(--tw-gradient-via-stops,var(--tw-gradient-position), var(--tw-gradient-from) var(--tw-gradient-from-position), var(--tw-gradient-to) var(--tw-gradient-to-position))}.hover\\:stroke-primary\\/15:hover{stroke:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:stroke-primary\\/15:hover{stroke:color-mix(in oklab, var(--primary) 15%, transparent)}}.hover\\:text-\\[var\\(--fs-success-700\\)\\]:hover{color:var(--fs-success-700)}.hover\\:text-accent-foreground:hover{color:var(--accent-foreground)}.hover\\:text-destructive:hover{color:var(--destructive)}.hover\\:text-foreground:hover{color:var(--foreground)}.hover\\:text-gray-200:hover{color:var(--color-gray-200)}.hover\\:text-primary:hover,.hover\\:text-primary\\/80:hover{color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:text-primary\\/80:hover{color:color-mix(in oklab, var(--primary) 80%, transparent)}}.hover\\:text-sidebar-accent-foreground:hover{color:var(--sidebar-accent-foreground)}.hover\\:underline:hover{text-decoration-line:underline}.hover\\:opacity-80:hover{opacity:.8}.hover\\:opacity-100:hover{opacity:1}.hover\\:shadow-md:hover{--tw-shadow:0 4px 6px -1px var(--tw-shadow-color,#0000001a), 0 2px 4px -2px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.hover\\:shadow-sm:hover{--tw-shadow:0 1px 3px 0 var(--tw-shadow-color,#0000001a), 0 1px 2px -1px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.hover\\:ring-1:hover{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.hover\\:ring-2:hover{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.hover\\:ring-border:hover{--tw-ring-color:var(--border)}.hover\\:ring-primary\\/50:hover{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:ring-primary\\/50:hover{--tw-ring-color:color-mix(in oklab, var(--primary) 50%, transparent)}}.hover\\:brightness-95:hover{--tw-brightness:brightness(95%);filter:var(--tw-blur,) var(--tw-brightness,) var(--tw-contrast,) var(--tw-grayscale,) var(--tw-hue-rotate,) var(--tw-invert,) var(--tw-saturate,) var(--tw-sepia,) var(--tw-drop-shadow,)}}.focus\\:border-primary:focus,.focus\\:border-primary\\/40:focus{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus\\:border-primary\\/40:focus{border-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.focus\\:bg-accent:focus{background-color:var(--accent)}.focus\\:bg-background:focus{background-color:var(--background)}.focus\\:text-accent-foreground:focus{color:var(--accent-foreground)}.focus\\:text-foreground:focus{color:var(--foreground)}.focus\\:ring-1:focus{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus\\:ring-2:focus{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus\\:ring-primary\\/15:focus{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus\\:ring-primary\\/15:focus{--tw-ring-color:color-mix(in oklab, var(--primary) 15%, transparent)}}.focus\\:ring-ring:focus{--tw-ring-color:var(--ring)}.focus\\:outline-none:focus{--tw-outline-style:none;outline-style:none}.focus-visible\\:relative:focus-visible{position:relative}.focus-visible\\:z-10:focus-visible{z-index:10}.focus-visible\\:bg-accent\\/50:focus-visible{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.focus-visible\\:bg-accent\\/50:focus-visible{background-color:color-mix(in oklab, var(--accent) 50%, transparent)}}.focus-visible\\:bg-primary:focus-visible,.focus-visible\\:bg-primary\\/10:focus-visible{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus-visible\\:bg-primary\\/10:focus-visible{background-color:color-mix(in oklab, var(--primary) 10%, transparent)}}.focus-visible\\:bg-primary\\/60:focus-visible{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus-visible\\:bg-primary\\/60:focus-visible{background-color:color-mix(in oklab, var(--primary) 60%, transparent)}}.focus-visible\\:stroke-primary\\/30:focus-visible{stroke:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus-visible\\:stroke-primary\\/30:focus-visible{stroke:color-mix(in oklab, var(--primary) 30%, transparent)}}.focus-visible\\:text-primary:focus-visible{color:var(--primary)}.focus-visible\\:opacity-100:focus-visible{opacity:1}.focus-visible\\:ring-0:focus-visible{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(0px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus-visible\\:ring-1:focus-visible{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus-visible\\:ring-2:focus-visible{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus-visible\\:ring-destructive:focus-visible{--tw-ring-color:var(--destructive)}.focus-visible\\:ring-primary:focus-visible{--tw-ring-color:var(--primary)}.focus-visible\\:ring-ring:focus-visible{--tw-ring-color:var(--ring)}.focus-visible\\:ring-offset-1:focus-visible{--tw-ring-offset-width:1px;--tw-ring-offset-shadow:var(--tw-ring-inset,) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)}.focus-visible\\:ring-offset-2:focus-visible{--tw-ring-offset-width:2px;--tw-ring-offset-shadow:var(--tw-ring-inset,) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)}.focus-visible\\:outline:focus-visible{outline-style:var(--tw-outline-style);outline-width:1px}.focus-visible\\:outline-2:focus-visible{outline-style:var(--tw-outline-style);outline-width:2px}.focus-visible\\:outline-ring:focus-visible{outline-color:var(--ring)}.focus-visible\\:outline-none:focus-visible{--tw-outline-style:none;outline-style:none}.focus-visible\\:ring-inset:focus-visible{--tw-ring-inset:inset}.active\\:cursor-grabbing:active{cursor:grabbing}.disabled\\:pointer-events-none:disabled{pointer-events:none}.disabled\\:cursor-default:disabled{cursor:default}.disabled\\:cursor-not-allowed:disabled{cursor:not-allowed}.disabled\\:bg-muted:disabled,.disabled\\:bg-muted\\/40:disabled{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.disabled\\:bg-muted\\/40:disabled{background-color:color-mix(in oklab, var(--muted) 40%, transparent)}}.disabled\\:text-muted-foreground:disabled,.disabled\\:text-muted-foreground\\/40:disabled{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.disabled\\:text-muted-foreground\\/40:disabled{color:color-mix(in oklab, var(--muted-foreground) 40%, transparent)}}.disabled\\:opacity-30:disabled{opacity:.3}.disabled\\:opacity-35:disabled{opacity:.35}.disabled\\:opacity-50:disabled{opacity:.5}.disabled\\:opacity-60:disabled{opacity:.6}@media (hover:hover){.disabled\\:hover\\:bg-transparent:disabled:hover{background-color:#0000}.disabled\\:hover\\:text-muted-foreground:disabled:hover{color:var(--muted-foreground)}}.has-\\[\\:disabled\\]\\:cursor-not-allowed:has(:disabled){cursor:not-allowed}.has-\\[\\:disabled\\]\\:opacity-50:has(:disabled){opacity:.5}.aria-expanded\\:bg-muted[aria-expanded=true],.aria-pressed\\:bg-muted[aria-pressed=true]{background-color:var(--muted)}.aria-pressed\\:font-semibold[aria-pressed=true]{--tw-font-weight:var(--font-weight-semibold);font-weight:var(--font-weight-semibold)}.aria-pressed\\:underline[aria-pressed=true]{text-decoration-line:underline}.data-\\[nav-drag\\=source\\]\\:opacity-40[data-nav-drag=source]{opacity:.4}.data-\\[nav-drop\\=over\\]\\:rounded-md[data-nav-drop=over]{border-radius:var(--radius-md)}.data-\\[nav-drop\\=over\\]\\:bg-sidebar-accent[data-nav-drop=over]{background-color:var(--sidebar-accent)}.data-\\[nav-drop\\=over\\]\\:ring-1[data-nav-drop=over]{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.data-\\[nav-drop\\=over\\]\\:ring-sidebar-primary[data-nav-drop=over]{--tw-ring-color:var(--sidebar-primary)}.data-\\[nav-drop\\=over\\]\\:outline-1[data-nav-drop=over]{outline-style:var(--tw-outline-style);outline-width:1px}.data-\\[nav-drop\\=over\\]\\:outline-sidebar-primary[data-nav-drop=over]{outline-color:var(--sidebar-primary)}.data-\\[nav-drop\\=over\\]\\:outline-dashed[data-nav-drop=over]{--tw-outline-style:dashed;outline-style:dashed}.data-\\[nav-drop\\=over\\]\\:ring-inset[data-nav-drop=over]{--tw-ring-inset:inset}.data-\\[theme\\=dark\\]\\:bg-\\[var\\(--fs-bg-subtle\\)\\][data-theme=dark]{background-color:var(--fs-bg-subtle)}@supports ((-webkit-backdrop-filter:var(--tw)) or (backdrop-filter:var(--tw))){.supports-\\[backdrop-filter\\]\\:bg-background\\/80{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.supports-\\[backdrop-filter\\]\\:bg-background\\/80{background-color:color-mix(in oklab, var(--background) 80%, transparent)}}.supports-\\[backdrop-filter\\]\\:bg-sidebar{background-color:var(--sidebar)}}@media (prefers-reduced-motion:no-preference){.motion-safe\\:animate-spin{animation:var(--animate-spin)}}@media (prefers-reduced-motion:reduce){.motion-reduce\\:transition-none{transition-property:none}}@media not all and (width>=48rem){.max-md\\:hidden{display:none}.max-md\\:inline-flex{display:inline-flex}.max-md\\:h-8{height:calc(var(--spacing) * 8)}.max-md\\:h-9{height:calc(var(--spacing) * 9)}.max-md\\:flex-nowrap{flex-wrap:nowrap}}@media not all and (width>=40rem){.max-sm\\:sr-only{clip-path:inset(50%);white-space:nowrap;border-width:0;width:1px;height:1px;margin:-1px;padding:0;position:absolute;overflow:hidden}.max-sm\\:ml-0{margin-left:0}.max-sm\\:ml-6{margin-left:calc(var(--spacing) * 6)}.max-sm\\:hidden{display:none}.max-sm\\:\\!h-dvh{height:100dvh!important}.max-sm\\:h-9{height:calc(var(--spacing) * 9)}.max-sm\\:\\!max-h-dvh{max-height:100dvh!important}.max-sm\\:min-h-14{min-height:calc(var(--spacing) * 14)}.max-sm\\:w-9{width:calc(var(--spacing) * 9)}.max-sm\\:\\!max-w-none{max-width:none!important}.max-sm\\:basis-full{flex-basis:100%}.max-sm\\:flex-nowrap{flex-wrap:nowrap}.max-sm\\:items-stretch{align-items:stretch}.max-sm\\:justify-end{justify-content:flex-end}.max-sm\\:\\!rounded-none{border-radius:0!important}.max-sm\\:rounded-none{border-radius:0}.max-sm\\:\\!border-0{border-style:var(--tw-border-style)!important;border-width:0!important}.max-sm\\:border-0{border-style:var(--tw-border-style);border-width:0}.max-sm\\:\\!bg-transparent{background-color:#0000!important}.max-sm\\:bg-background{background-color:var(--background)}.max-sm\\:bg-transparent{background-color:#0000}.max-sm\\:p-0{padding:0}.max-sm\\:px-0{padding-inline:0}.max-sm\\:py-density-4{padding-block:var(--spacing-density-4)}.max-sm\\:text-base{font-size:var(--text-base);line-height:var(--tw-leading,var(--text-base--line-height))}.max-sm\\:leading-6{--tw-leading:calc(var(--spacing) * 6);line-height:calc(var(--spacing) * 6)}.max-sm\\:shadow-none{--tw-shadow:0 0 #0000;box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.max-sm\\:transition-transform{transition-property:transform,translate,scale,rotate;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.max-sm\\:duration-200{--tw-duration:.2s;transition-duration:.2s}.max-sm\\:ease-out{--tw-ease:var(--ease-out);transition-timing-function:var(--ease-out)}.max-sm\\:data-\\[state\\=closed\\]\\:translate-x-full[data-state=closed]{--tw-translate-x:100%;translate:var(--tw-translate-x) var(--tw-translate-y)}.max-sm\\:data-\\[state\\=open\\]\\:translate-x-0[data-state=open]{--tw-translate-x:0px;translate:var(--tw-translate-x) var(--tw-translate-y)}@media (prefers-reduced-motion:reduce){.max-sm\\:motion-reduce\\:transition-none{transition-property:none}}}@media (width>=40rem){.sm\\:col-span-2{grid-column:span 2/span 2}.sm\\:block{display:block}.sm\\:hidden{display:none}.sm\\:h-7{height:calc(var(--spacing) * 7)}.sm\\:w-7{width:calc(var(--spacing) * 7)}.sm\\:w-48{width:calc(var(--spacing) * 48)}.sm\\:max-w-xs{max-width:var(--container-xs)}.sm\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}.sm\\:grid-cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}.sm\\:grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}.sm\\:grid-cols-6{grid-template-columns:repeat(6,minmax(0,1fr))}.sm\\:grid-cols-\\[8rem_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:8rem minmax(0,1fr) auto}.sm\\:grid-cols-\\[9rem_4\\.5rem_minmax\\(0\\,1fr\\)_2\\.25rem\\]{grid-template-columns:9rem 4.5rem minmax(0,1fr) 2.25rem}.sm\\:grid-cols-\\[auto_1fr\\]{grid-template-columns:auto 1fr}.sm\\:grid-cols-\\[minmax\\(0\\,1fr\\)_10rem\\]{grid-template-columns:minmax(0,1fr) 10rem}.sm\\:grid-cols-\\[minmax\\(0\\,1fr\\)_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:minmax(0,1fr) minmax(0,1fr) auto}.sm\\:grid-cols-\\[minmax\\(8rem\\,14rem\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(8rem,14rem) minmax(0,1fr)}.sm\\:grid-cols-\\[minmax\\(10rem\\,0\\.35fr\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(10rem,.35fr) minmax(0,1fr)}.sm\\:grid-cols-\\[minmax\\(10rem\\,16rem\\)\\]{grid-template-columns:minmax(10rem,16rem)}.sm\\:flex-row{flex-direction:row}.sm\\:flex-wrap{flex-wrap:wrap}.sm\\:items-center{align-items:center}.sm\\:gap-3{gap:calc(var(--spacing) * 3)}.sm\\:p-density-4{padding:var(--spacing-density-4)}}@media (width>=48rem){.md\\:z-\\[var\\(--x\\)\\]{z-index:var(--x)}.md\\:col-span-1{grid-column:span 1/span 1}.md\\:col-span-2{grid-column:span 2/span 2}.md\\:block{display:block}.md\\:flex{display:flex}.md\\:h-12{height:calc(var(--spacing) * 12)}.md\\:w-40{width:calc(var(--spacing) * 40)}.md\\:max-w-\\[24rem\\]{max-width:24rem}.md\\:min-w-\\[14rem\\]{min-width:14rem}.md\\:flex-\\[1_1_14rem\\]{flex:14rem}.md\\:shrink-0{flex-shrink:0}.md\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}.md\\:grid-cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}.md\\:grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}.md\\:grid-cols-\\[1fr_auto_1fr\\]{grid-template-columns:1fr auto 1fr}.md\\:grid-cols-\\[minmax\\(0\\,1fr\\)_9rem_6rem\\]{grid-template-columns:minmax(0,1fr) 9rem 6rem}.md\\:grid-cols-\\[minmax\\(7rem\\,10rem\\)_auto_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:minmax(7rem,10rem) auto minmax(0,1fr) auto}.md\\:grid-cols-\\[minmax\\(8rem\\,12rem\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(8rem,12rem) minmax(0,1fr)}.md\\:grid-cols-\\[repeat\\(var\\(--stat-cols\\,4\\)\\,minmax\\(0\\,1fr\\)\\)\\]{grid-template-columns:repeat(var(--stat-cols,4),minmax(0,1fr))}.md\\:flex-nowrap{flex-wrap:nowrap}.md\\:flex-wrap{flex-wrap:wrap}.md\\:items-center{align-items:center}.md\\:p-2{padding:calc(var(--spacing) * 2)}}@media (width>=64rem){.lg\\:col-span-2{grid-column:span 2/span 2}.lg\\:block{display:block}.lg\\:flex{display:flex}.lg\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}.lg\\:grid-cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}.lg\\:grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}.lg\\:grid-cols-\\[1\\.2fr_0\\.8fr\\]{grid-template-columns:1.2fr .8fr}.lg\\:grid-cols-\\[1fr_18rem\\]{grid-template-columns:1fr 18rem}.lg\\:grid-cols-\\[minmax\\(0\\,1\\.4fr\\)_minmax\\(0\\,0\\.7fr\\)\\]{grid-template-columns:minmax(0,1.4fr) minmax(0,.7fr)}.lg\\:grid-cols-\\[minmax\\(0\\,1fr\\)_12rem\\]{grid-template-columns:minmax(0,1fr) 12rem}.lg\\:grid-cols-\\[minmax\\(0\\,1fr\\)_minmax\\(0\\,1\\.15fr\\)\\]{grid-template-columns:minmax(0,1fr) minmax(0,1.15fr)}.lg\\:grid-cols-\\[minmax\\(0\\,1fr\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(0,1fr) minmax(0,1fr)}.lg\\:grid-cols-\\[minmax\\(12rem\\,1\\.1fr\\)_minmax\\(12rem\\,1fr\\)_minmax\\(14rem\\,1fr\\)_auto\\]{grid-template-columns:minmax(12rem,1.1fr) minmax(12rem,1fr) minmax(14rem,1fr) auto}.lg\\:grid-cols-\\[minmax\\(22rem\\,0\\.9fr\\)_minmax\\(24rem\\,1\\.1fr\\)\\]{grid-template-columns:minmax(22rem,.9fr) minmax(24rem,1.1fr)}.lg\\:items-center{align-items:center}}@media (width>=80rem){.xl\\:sticky{position:sticky}.xl\\:top-0{top:0}.xl\\:inline{display:inline}.xl\\:max-h-screen{max-height:100vh}.xl\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}.xl\\:grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}.xl\\:grid-cols-6{grid-template-columns:repeat(6,minmax(0,1fr))}.xl\\:grid-cols-\\[14rem_minmax\\(30rem\\,0\\.9fr\\)_minmax\\(36rem\\,1\\.1fr\\)\\]{grid-template-columns:14rem minmax(30rem,.9fr) minmax(36rem,1.1fr)}.xl\\:grid-cols-\\[minmax\\(0\\,1\\.2fr\\)_minmax\\(22rem\\,0\\.8fr\\)\\]{grid-template-columns:minmax(0,1.2fr) minmax(22rem,.8fr)}.xl\\:grid-cols-\\[minmax\\(0\\,1fr\\)_minmax\\(20rem\\,26rem\\)\\]{grid-template-columns:minmax(0,1fr) minmax(20rem,26rem)}.xl\\:grid-cols-\\[minmax\\(12rem\\,1fr\\)_auto\\]{grid-template-columns:minmax(12rem,1fr) auto}.xl\\:items-center{align-items:center}.xl\\:self-start{align-self:flex-start}.xl\\:overflow-auto{overflow:auto}}@container not (width>=60rem){.\\@max-\\[60rem\\]\\:hidden{display:none}}@container not (width>=48rem){.\\@max-\\[48rem\\]\\:hidden{display:none}.\\@max-\\[48rem\\]\\:inline-flex{display:inline-flex}}@container app-shell not (width>=48rem){.\\@max-3xl\\/app-shell\\:max-w-full{max-width:100%}.\\@max-3xl\\/app-shell\\:min-w-0{min-width:0}.\\@max-3xl\\/app-shell\\:shrink{flex-shrink:1}.\\@max-3xl\\/app-shell\\:overflow-x-auto{overflow-x:auto}}@container not (width>=28rem){.\\@max-md\\:\\!block{display:block!important}.\\@max-md\\:hidden{display:none}}@container secret (width>=22rem){.\\@min-\\[22rem\\]\\/secret\\:basis-40{flex-basis:calc(var(--spacing) * 40)}.\\@min-\\[22rem\\]\\/secret\\:basis-72{flex-basis:calc(var(--spacing) * 72)}.\\@min-\\[22rem\\]\\/secret\\:flex-row{flex-direction:row}}@container combobox-options (width>=28rem){.\\@min-\\[28rem\\]\\/combobox-options\\:ml-auto{margin-left:auto}.\\@min-\\[28rem\\]\\/combobox-options\\:max-w-\\[55\\%\\]{max-width:55%}.\\@min-\\[28rem\\]\\/combobox-options\\:flex-1{flex:1}.\\@min-\\[28rem\\]\\/combobox-options\\:shrink-0{flex-shrink:0}}@container (width>=28rem){.\\@md\\:grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}}@container combobox-options (width>=28rem){.\\@min-\\[28rem\\]\\/combobox-options\\:flex-row{flex-direction:row}.\\@min-\\[28rem\\]\\/combobox-options\\:items-center{align-items:center}.\\@min-\\[28rem\\]\\/combobox-options\\:gap-2{gap:calc(var(--spacing) * 2)}.\\@min-\\[28rem\\]\\/combobox-options\\:truncate{text-overflow:ellipsis;white-space:nowrap;overflow:hidden}.\\@min-\\[28rem\\]\\/combobox-options\\:pl-2{padding-left:calc(var(--spacing) * 2)}.\\@min-\\[28rem\\]\\/combobox-options\\:text-right{text-align:right}}@container (width>=36rem){.\\@xl\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}}@container endpoint (width>=40rem){.\\@min-\\[40rem\\]\\/endpoint\\:col-span-1{grid-column:span 1/span 1}.\\@min-\\[40rem\\]\\/endpoint\\:col-span-3{grid-column:span 3/span 3}.\\@min-\\[40rem\\]\\/endpoint\\:grid-cols-\\[minmax\\(14rem\\,1fr\\)_auto_auto\\]{grid-template-columns:minmax(14rem,1fr) auto auto}}@container (width>=42rem){.\\@2xl\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:order-none{order:0}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:order-none{order:0}.\\@3xl\\/workspace\\:col-start-1{grid-column-start:1}.\\@3xl\\/workspace\\:col-start-2{grid-column-start:2}.\\@3xl\\/workspace\\:col-start-3{grid-column-start:3}.\\@3xl\\/workspace\\:row-span-2{grid-row:span 2/span 2}.\\@3xl\\/workspace\\:row-start-1{grid-row-start:1}.\\@3xl\\/workspace\\:row-start-2{grid-row-start:2}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:ml-0{margin-left:0}.\\@3xl\\/app-shell\\:block{display:block}.\\@3xl\\/app-shell\\:flex{display:flex}.\\@3xl\\/app-shell\\:hidden{display:none}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:block{display:block}.\\@3xl\\/split-pane\\:hidden{display:none}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:block{display:block}.\\@3xl\\/workspace\\:flex{display:flex}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:h-14{height:calc(var(--spacing) * 14)}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:h-\\[var\\(--workspace-bottom-height\\)\\]{height:var(--workspace-bottom-height)}.\\@3xl\\/workspace\\:h-full{height:100%}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:max-h-none{max-height:none}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:min-h-0{min-height:0}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:w-\\[var\\(--split-pane-width\\)\\]{width:var(--split-pane-width)}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:w-\\[var\\(--workspace-region-width\\)\\]{width:var(--workspace-region-width)}}@container session-viewer (width>=48rem){.\\@min-\\[48rem\\]\\/session-viewer\\:max-w-\\[85\\%\\]{max-width:85%}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:flex-1{flex:1}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:flex-none{flex:none}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:basis-auto{flex-basis:auto}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:grid-cols-\\[auto_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:auto minmax(0,1fr) auto}.\\@3xl\\/workspace\\:grid-rows-\\[minmax\\(0\\,1fr\\)_auto\\]{grid-template-rows:minmax(0,1fr) auto}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:flex-row{flex-direction:row}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:flex-row{flex-direction:row}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:flex-nowrap{flex-wrap:nowrap}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:content-stretch{align-content:stretch}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:items-start{align-items:flex-start}.\\@3xl\\/app-shell\\:justify-between{justify-content:space-between}.\\@3xl\\/app-shell\\:gap-density-3{gap:var(--spacing-density-3)}}@container session-viewer (width>=48rem){.\\@min-\\[48rem\\]\\/session-viewer\\:gap-density-3{gap:var(--spacing-density-3)}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:overflow-hidden{overflow:hidden}.\\@3xl\\/workspace\\:border-r{border-right-style:var(--tw-border-style);border-right-width:1px}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:border-b-0{border-bottom-style:var(--tw-border-style);border-bottom-width:0}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:border-l{border-left-style:var(--tw-border-style);border-left-width:1px}}@container session-viewer (width>=48rem){.\\@min-\\[48rem\\]\\/session-viewer\\:p-density-4{padding:var(--spacing-density-4)}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:px-density-4{padding-inline:var(--spacing-density-4)}}@container session-viewer (width>=48rem){.\\@min-\\[48rem\\]\\/session-viewer\\:px-density-4{padding-inline:var(--spacing-density-4)}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:py-0{padding-block:0}}@container session-viewer (width>=48rem){.\\@min-\\[48rem\\]\\/session-viewer\\:pb-density-4{padding-bottom:var(--spacing-density-4)}}@container endpoint (width>=56rem){.\\@min-\\[56rem\\]\\/endpoint\\:col-span-1{grid-column:span 1/span 1}.\\@min-\\[56rem\\]\\/endpoint\\:grid-cols-\\[minmax\\(14rem\\,1fr\\)_auto_auto_minmax\\(14rem\\,1fr\\)\\]{grid-template-columns:minmax(14rem,1fr) auto auto minmax(14rem,1fr)}}@container (width>=60rem){.\\@min-\\[60rem\\]\\:hidden{display:none}.\\@min-\\[60rem\\]\\:gap-2{gap:calc(var(--spacing) * 2)}.\\@min-\\[60rem\\]\\:px-2{padding-inline:calc(var(--spacing) * 2)}.\\@min-\\[60rem\\]\\:px-2\\.5{padding-inline:calc(var(--spacing) * 2.5)}.\\@min-\\[60rem\\]\\:px-3{padding-inline:calc(var(--spacing) * 3)}}@container app-content (width>=96rem){.\\@8xl\\/app-content\\:max-w-8xl{max-width:var(--container-8xl)}}@container app-content (width>=120rem){.\\@9xl\\/app-content\\:max-w-9xl{max-width:var(--container-9xl)}}@media (prefers-color-scheme:dark){.dark\\:border-amber-400{border-color:var(--color-amber-400)}.dark\\:border-amber-500\\/30{border-color:#f99c004d}@supports (color:color-mix(in lab, red, red)){.dark\\:border-amber-500\\/30{border-color:color-mix(in oklab, var(--color-amber-500) 30%, transparent)}}.dark\\:border-amber-500\\/40{border-color:#f99c0066}@supports (color:color-mix(in lab, red, red)){.dark\\:border-amber-500\\/40{border-color:color-mix(in oklab, var(--color-amber-500) 40%, transparent)}}.dark\\:border-amber-800{border-color:var(--color-amber-800)}.dark\\:border-amber-800\\/60{border-color:#953d0099}@supports (color:color-mix(in lab, red, red)){.dark\\:border-amber-800\\/60{border-color:color-mix(in oklab, var(--color-amber-800) 60%, transparent)}}.dark\\:border-emerald-500{border-color:var(--color-emerald-500)}.dark\\:border-emerald-500\\/30{border-color:#00bb7f4d}@supports (color:color-mix(in lab, red, red)){.dark\\:border-emerald-500\\/30{border-color:color-mix(in oklab, var(--color-emerald-500) 30%, transparent)}}.dark\\:border-emerald-500\\/40{border-color:#00bb7f66}@supports (color:color-mix(in lab, red, red)){.dark\\:border-emerald-500\\/40{border-color:color-mix(in oklab, var(--color-emerald-500) 40%, transparent)}}.dark\\:border-emerald-800{border-color:var(--color-emerald-800)}.dark\\:border-red-500{border-color:var(--color-red-500)}.dark\\:border-red-800{border-color:var(--color-red-800)}.dark\\:border-rose-500\\/30{border-color:#ff23574d}@supports (color:color-mix(in lab, red, red)){.dark\\:border-rose-500\\/30{border-color:color-mix(in oklab, var(--color-rose-500) 30%, transparent)}}.dark\\:border-rose-500\\/40{border-color:#ff235766}@supports (color:color-mix(in lab, red, red)){.dark\\:border-rose-500\\/40{border-color:color-mix(in oklab, var(--color-rose-500) 40%, transparent)}}.dark\\:border-rose-800{border-color:var(--color-rose-800)}.dark\\:border-sky-500{border-color:var(--color-sky-500)}.dark\\:border-sky-500\\/30{border-color:#00a5ef4d}@supports (color:color-mix(in lab, red, red)){.dark\\:border-sky-500\\/30{border-color:color-mix(in oklab, var(--color-sky-500) 30%, transparent)}}.dark\\:border-sky-500\\/40{border-color:#00a5ef66}@supports (color:color-mix(in lab, red, red)){.dark\\:border-sky-500\\/40{border-color:color-mix(in oklab, var(--color-sky-500) 40%, transparent)}}.dark\\:border-sky-800{border-color:var(--color-sky-800)}.dark\\:border-slate-300{border-color:var(--color-slate-300)}.dark\\:border-yellow-700{border-color:var(--color-yellow-700)}.dark\\:border-l-amber-500{border-left-color:var(--color-amber-500)}.dark\\:border-l-border{border-left-color:var(--border)}.dark\\:border-l-emerald-500{border-left-color:var(--color-emerald-500)}.dark\\:border-l-indigo-500{border-left-color:var(--color-indigo-500)}.dark\\:border-l-rose-500{border-left-color:var(--color-rose-500)}.dark\\:border-l-sky-500{border-left-color:var(--color-sky-500)}.dark\\:border-l-slate-500{border-left-color:var(--color-slate-500)}.dark\\:border-l-teal-500{border-left-color:var(--color-teal-500)}.dark\\:border-l-violet-500{border-left-color:var(--color-violet-500)}.dark\\:bg-amber-400{background-color:var(--color-amber-400)}.dark\\:bg-amber-400\\/10{background-color:#fcbb001a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-400\\/10{background-color:color-mix(in oklab, var(--color-amber-400) 10%, transparent)}}.dark\\:bg-amber-500{background-color:var(--color-amber-500)}.dark\\:bg-amber-500\\/10{background-color:#f99c001a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-500\\/10{background-color:color-mix(in oklab, var(--color-amber-500) 10%, transparent)}}.dark\\:bg-amber-500\\/15{background-color:#f99c0026}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-500\\/15{background-color:color-mix(in oklab, var(--color-amber-500) 15%, transparent)}}.dark\\:bg-amber-500\\/20{background-color:#f99c0033}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-500\\/20{background-color:color-mix(in oklab, var(--color-amber-500) 20%, transparent)}}.dark\\:bg-amber-900{background-color:var(--color-amber-900)}.dark\\:bg-amber-950\\/20{background-color:#46190133}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-950\\/20{background-color:color-mix(in oklab, var(--color-amber-950) 20%, transparent)}}.dark\\:bg-amber-950\\/40{background-color:#46190166}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-950\\/40{background-color:color-mix(in oklab, var(--color-amber-950) 40%, transparent)}}.dark\\:bg-amber-950\\/50{background-color:#46190180}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-950\\/50{background-color:color-mix(in oklab, var(--color-amber-950) 50%, transparent)}}.dark\\:bg-blue-500\\/10{background-color:#3080ff1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-blue-500\\/10{background-color:color-mix(in oklab, var(--color-blue-500) 10%, transparent)}}.dark\\:bg-blue-500\\/20{background-color:#3080ff33}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-blue-500\\/20{background-color:color-mix(in oklab, var(--color-blue-500) 20%, transparent)}}.dark\\:bg-emerald-400\\/10{background-color:#00d2941a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-emerald-400\\/10{background-color:color-mix(in oklab, var(--color-emerald-400) 10%, transparent)}}.dark\\:bg-emerald-500{background-color:var(--color-emerald-500)}.dark\\:bg-emerald-500\\/15{background-color:#00bb7f26}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-emerald-500\\/15{background-color:color-mix(in oklab, var(--color-emerald-500) 15%, transparent)}}.dark\\:bg-emerald-900{background-color:var(--color-emerald-900)}.dark\\:bg-emerald-950\\/40{background-color:#002c2266}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-emerald-950\\/40{background-color:color-mix(in oklab, var(--color-emerald-950) 40%, transparent)}}.dark\\:bg-emerald-950\\/50{background-color:#002c2280}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-emerald-950\\/50{background-color:color-mix(in oklab, var(--color-emerald-950) 50%, transparent)}}.dark\\:bg-gray-500\\/20{background-color:#6a728233}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-gray-500\\/20{background-color:color-mix(in oklab, var(--color-gray-500) 20%, transparent)}}.dark\\:bg-green-500\\/20{background-color:#00c75833}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-green-500\\/20{background-color:color-mix(in oklab, var(--color-green-500) 20%, transparent)}}.dark\\:bg-indigo-400\\/10{background-color:#7d87ff1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-indigo-400\\/10{background-color:color-mix(in oklab, var(--color-indigo-400) 10%, transparent)}}.dark\\:bg-indigo-500{background-color:var(--color-indigo-500)}.dark\\:bg-indigo-950\\/50{background-color:#1e1a4d80}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-indigo-950\\/50{background-color:color-mix(in oklab, var(--color-indigo-950) 50%, transparent)}}.dark\\:bg-muted-foreground{background-color:var(--muted-foreground)}.dark\\:bg-red-500{background-color:var(--color-red-500)}.dark\\:bg-red-500\\/20{background-color:#fb2c3633}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-red-500\\/20{background-color:color-mix(in oklab, var(--color-red-500) 20%, transparent)}}.dark\\:bg-red-800{background-color:var(--color-red-800)}.dark\\:bg-red-900{background-color:var(--color-red-900)}.dark\\:bg-red-950{background-color:var(--color-red-950)}.dark\\:bg-rose-400\\/10{background-color:#ff667f1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-rose-400\\/10{background-color:color-mix(in oklab, var(--color-rose-400) 10%, transparent)}}.dark\\:bg-rose-500{background-color:var(--color-rose-500)}.dark\\:bg-rose-500\\/15{background-color:#ff235726}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-rose-500\\/15{background-color:color-mix(in oklab, var(--color-rose-500) 15%, transparent)}}.dark\\:bg-rose-950\\/30{background-color:#4d02184d}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-rose-950\\/30{background-color:color-mix(in oklab, var(--color-rose-950) 30%, transparent)}}.dark\\:bg-rose-950\\/40{background-color:#4d021866}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-rose-950\\/40{background-color:color-mix(in oklab, var(--color-rose-950) 40%, transparent)}}.dark\\:bg-rose-950\\/50{background-color:#4d021880}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-rose-950\\/50{background-color:color-mix(in oklab, var(--color-rose-950) 50%, transparent)}}.dark\\:bg-sky-400\\/10{background-color:#00bcfe1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-sky-400\\/10{background-color:color-mix(in oklab, var(--color-sky-400) 10%, transparent)}}.dark\\:bg-sky-500{background-color:var(--color-sky-500)}.dark\\:bg-sky-500\\/15{background-color:#00a5ef26}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-sky-500\\/15{background-color:color-mix(in oklab, var(--color-sky-500) 15%, transparent)}}.dark\\:bg-sky-900{background-color:var(--color-sky-900)}.dark\\:bg-sky-950\\/40{background-color:#052f4a66}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-sky-950\\/40{background-color:color-mix(in oklab, var(--color-sky-950) 40%, transparent)}}.dark\\:bg-sky-950\\/50{background-color:#052f4a80}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-sky-950\\/50{background-color:color-mix(in oklab, var(--color-sky-950) 50%, transparent)}}.dark\\:bg-slate-300{background-color:var(--color-slate-300)}.dark\\:bg-slate-400\\/10{background-color:#90a1b91a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-slate-400\\/10{background-color:color-mix(in oklab, var(--color-slate-400) 10%, transparent)}}.dark\\:bg-slate-500{background-color:var(--color-slate-500)}.dark\\:bg-slate-600{background-color:var(--color-slate-600)}.dark\\:bg-teal-400\\/10{background-color:#00d3bd1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-teal-400\\/10{background-color:color-mix(in oklab, var(--color-teal-400) 10%, transparent)}}.dark\\:bg-teal-500{background-color:var(--color-teal-500)}.dark\\:bg-violet-400\\/10{background-color:#a685ff1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-violet-400\\/10{background-color:color-mix(in oklab, var(--color-violet-400) 10%, transparent)}}.dark\\:bg-violet-500{background-color:var(--color-violet-500)}.dark\\:bg-violet-500\\/20{background-color:#8d54ff33}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-violet-500\\/20{background-color:color-mix(in oklab, var(--color-violet-500) 20%, transparent)}}.dark\\:bg-violet-950\\/50{background-color:#2f0d6880}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-violet-950\\/50{background-color:color-mix(in oklab, var(--color-violet-950) 50%, transparent)}}.dark\\:bg-yellow-500\\/20{background-color:#edb20033}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-yellow-500\\/20{background-color:color-mix(in oklab, var(--color-yellow-500) 20%, transparent)}}.dark\\:bg-yellow-950{background-color:var(--color-yellow-950)}.dark\\:text-\\[\\#00f7f1\\]{color:#00f7f1}.dark\\:text-\\[\\#61dafb\\]{color:#61dafb}.dark\\:text-\\[\\#8080f2\\]{color:#8080f2}.dark\\:text-\\[\\#d7ff64\\]{color:#d7ff64}.dark\\:text-\\[\\#e5576a\\]{color:#e5576a}.dark\\:text-\\[\\#ffd43b\\]{color:#ffd43b}.dark\\:text-amber-100{color:var(--color-amber-100)}.dark\\:text-amber-100\\/90{color:#fef3c6e6}@supports (color:color-mix(in lab, red, red)){.dark\\:text-amber-100\\/90{color:color-mix(in oklab, var(--color-amber-100) 90%, transparent)}}.dark\\:text-amber-200{color:var(--color-amber-200)}.dark\\:text-amber-300{color:var(--color-amber-300)}.dark\\:text-amber-300\\/70{color:#ffd236b3}@supports (color:color-mix(in lab, red, red)){.dark\\:text-amber-300\\/70{color:color-mix(in oklab, var(--color-amber-300) 70%, transparent)}}.dark\\:text-amber-400{color:var(--color-amber-400)}.dark\\:text-amber-500{color:var(--color-amber-500)}.dark\\:text-amber-950{color:var(--color-amber-950)}.dark\\:text-blue-100{color:var(--color-blue-100)}.dark\\:text-blue-100\\/90{color:#dbeafee6}@supports (color:color-mix(in lab, red, red)){.dark\\:text-blue-100\\/90{color:color-mix(in oklab, var(--color-blue-100) 90%, transparent)}}.dark\\:text-blue-300{color:var(--color-blue-300)}.dark\\:text-blue-300\\/70{color:#90c5ffb3}@supports (color:color-mix(in lab, red, red)){.dark\\:text-blue-300\\/70{color:color-mix(in oklab, var(--color-blue-300) 70%, transparent)}}.dark\\:text-blue-400{color:var(--color-blue-400)}.dark\\:text-emerald-100{color:var(--color-emerald-100)}.dark\\:text-emerald-100\\/90{color:#d0fae5e6}@supports (color:color-mix(in lab, red, red)){.dark\\:text-emerald-100\\/90{color:color-mix(in oklab, var(--color-emerald-100) 90%, transparent)}}.dark\\:text-emerald-200{color:var(--color-emerald-200)}.dark\\:text-emerald-300{color:var(--color-emerald-300)}.dark\\:text-emerald-300\\/70{color:#5ee9b5b3}@supports (color:color-mix(in lab, red, red)){.dark\\:text-emerald-300\\/70{color:color-mix(in oklab, var(--color-emerald-300) 70%, transparent)}}.dark\\:text-emerald-400{color:var(--color-emerald-400)}.dark\\:text-emerald-950{color:var(--color-emerald-950)}.dark\\:text-green-300{color:var(--color-green-300)}.dark\\:text-green-400{color:var(--color-green-400)}.dark\\:text-indigo-200{color:var(--color-indigo-200)}.dark\\:text-indigo-300{color:var(--color-indigo-300)}.dark\\:text-orange-300{color:var(--color-orange-300)}.dark\\:text-purple-100{color:var(--color-purple-100)}.dark\\:text-purple-100\\/90{color:#f3e8ffe6}@supports (color:color-mix(in lab, red, red)){.dark\\:text-purple-100\\/90{color:color-mix(in oklab, var(--color-purple-100) 90%, transparent)}}.dark\\:text-purple-300{color:var(--color-purple-300)}.dark\\:text-purple-300\\/70{color:#d9b3ffb3}@supports (color:color-mix(in lab, red, red)){.dark\\:text-purple-300\\/70{color:color-mix(in oklab, var(--color-purple-300) 70%, transparent)}}.dark\\:text-purple-400{color:var(--color-purple-400)}.dark\\:text-red-100{color:var(--color-red-100)}.dark\\:text-red-100\\/90{color:#ffe2e2e6}@supports (color:color-mix(in lab, red, red)){.dark\\:text-red-100\\/90{color:color-mix(in oklab, var(--color-red-100) 90%, transparent)}}.dark\\:text-red-200{color:var(--color-red-200)}.dark\\:text-red-300{color:var(--color-red-300)}.dark\\:text-red-300\\/70{color:#ffa3a3b3}@supports (color:color-mix(in lab, red, red)){.dark\\:text-red-300\\/70{color:color-mix(in oklab, var(--color-red-300) 70%, transparent)}}.dark\\:text-red-400{color:var(--color-red-400)}.dark\\:text-red-950{color:var(--color-red-950)}.dark\\:text-rose-200{color:var(--color-rose-200)}.dark\\:text-rose-300{color:var(--color-rose-300)}.dark\\:text-rose-400{color:var(--color-rose-400)}.dark\\:text-sky-200{color:var(--color-sky-200)}.dark\\:text-sky-300{color:var(--color-sky-300)}.dark\\:text-sky-400{color:var(--color-sky-400)}.dark\\:text-sky-950{color:var(--color-sky-950)}.dark\\:text-slate-300{color:var(--color-slate-300)}.dark\\:text-slate-500{color:var(--color-slate-500)}.dark\\:text-slate-950{color:var(--color-slate-950)}.dark\\:text-teal-300{color:var(--color-teal-300)}.dark\\:text-violet-200{color:var(--color-violet-200)}.dark\\:text-violet-300{color:var(--color-violet-300)}.dark\\:text-violet-400{color:var(--color-violet-400)}.dark\\:text-yellow-200{color:var(--color-yellow-200)}.dark\\:text-yellow-300{color:var(--color-yellow-300)}.dark\\:text-yellow-400{color:var(--color-yellow-400)}.dark\\:ring-amber-400\\/25{--tw-ring-color:#fcbb0040}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-amber-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-amber-400) 25%, transparent)}}.dark\\:ring-amber-500\\/30{--tw-ring-color:#f99c004d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-amber-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-amber-500) 30%, transparent)}}.dark\\:ring-emerald-400\\/25{--tw-ring-color:#00d29440}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-emerald-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-emerald-400) 25%, transparent)}}.dark\\:ring-emerald-500\\/30{--tw-ring-color:#00bb7f4d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-emerald-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-emerald-500) 30%, transparent)}}.dark\\:ring-indigo-400\\/25{--tw-ring-color:#7d87ff40}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-indigo-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-indigo-400) 25%, transparent)}}.dark\\:ring-rose-400\\/25{--tw-ring-color:#ff667f40}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-rose-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-rose-400) 25%, transparent)}}.dark\\:ring-rose-500\\/30{--tw-ring-color:#ff23574d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-rose-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-rose-500) 30%, transparent)}}.dark\\:ring-sky-400\\/25{--tw-ring-color:#00bcfe40}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-sky-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-sky-400) 25%, transparent)}}.dark\\:ring-sky-500\\/30{--tw-ring-color:#00a5ef4d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-sky-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-sky-500) 30%, transparent)}}.dark\\:ring-slate-400\\/25{--tw-ring-color:#90a1b940}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-slate-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-slate-400) 25%, transparent)}}.dark\\:ring-slate-500\\/30{--tw-ring-color:#62748e4d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-slate-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-slate-500) 30%, transparent)}}.dark\\:ring-teal-400\\/25{--tw-ring-color:#00d3bd40}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-teal-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-teal-400) 25%, transparent)}}.dark\\:ring-violet-400\\/25{--tw-ring-color:#a685ff40}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-violet-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-violet-400) 25%, transparent)}}.dark\\:ring-violet-500\\/30{--tw-ring-color:#8d54ff4d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-violet-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-violet-500) 30%, transparent)}}@media (hover:hover){.dark\\:hover\\:border-emerald-500\\/40:hover{border-color:#00bb7f66}@supports (color:color-mix(in lab, red, red)){.dark\\:hover\\:border-emerald-500\\/40:hover{border-color:color-mix(in oklab, var(--color-emerald-500) 40%, transparent)}}.dark\\:hover\\:bg-emerald-500\\/10:hover{background-color:#00bb7f1a}@supports (color:color-mix(in lab, red, red)){.dark\\:hover\\:bg-emerald-500\\/10:hover{background-color:color-mix(in oklab, var(--color-emerald-500) 10%, transparent)}}}}.\\[\\&_\\*\\]\\:pointer-events-none *{pointer-events:none}.\\[\\&_\\.chroma\\]\\:bg-transparent .chroma{background-color:#0000}.\\[\\&_\\.chroma_\\.c\\]\\:text-slate-500 .chroma .c{color:var(--color-slate-500)}.\\[\\&_\\.chroma_\\.c\\]\\:italic .chroma .c{font-style:italic}.\\[\\&_\\.chroma_\\.cp\\]\\:text-slate-500 .chroma .cp{color:var(--color-slate-500)}.\\[\\&_\\.chroma_\\.cp\\]\\:italic .chroma .cp{font-style:italic}.\\[\\&_\\.chroma_\\.na\\]\\:text-amber-700 .chroma .na{color:var(--color-amber-700)}@media (prefers-color-scheme:dark){.\\[\\&_\\.chroma_\\.na\\]\\:dark\\:text-amber-300 .chroma .na{color:var(--color-amber-300)}}.\\[\\&_\\.chroma_\\.nt\\]\\:font-semibold .chroma .nt{--tw-font-weight:var(--font-weight-semibold);font-weight:var(--font-weight-semibold)}.\\[\\&_\\.chroma_\\.nt\\]\\:text-pink-700 .chroma .nt{color:var(--color-pink-700)}@media (prefers-color-scheme:dark){.\\[\\&_\\.chroma_\\.nt\\]\\:dark\\:text-pink-300 .chroma .nt{color:var(--color-pink-300)}}.\\[\\&_\\.chroma_\\.s\\]\\:text-emerald-700 .chroma .s{color:var(--color-emerald-700)}@media (prefers-color-scheme:dark){.\\[\\&_\\.chroma_\\.s\\]\\:dark\\:text-emerald-300 .chroma .s{color:var(--color-emerald-300)}}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:absolute [data-profile-tree-item]:before{content:var(--tw-content);position:absolute}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:top-1\\/2 [data-profile-tree-item]:before{content:var(--tw-content);top:50%}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:-left-4 [data-profile-tree-item]:before{content:var(--tw-content);left:calc(var(--spacing) * -4)}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:h-px [data-profile-tree-item]:before{content:var(--tw-content);height:1px}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:w-4 [data-profile-tree-item]:before{content:var(--tw-content);width:calc(var(--spacing) * 4)}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:bg-border [data-profile-tree-item]:before{content:var(--tw-content);background-color:var(--border)}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:content-\\[\\'\\'\\] [data-profile-tree-item]:before{--tw-content:\"\";content:var(--tw-content)}.\\[\\&_\\[data-slot\\=monaco-editor\\]\\]\\:h-full [data-slot=monaco-editor]{height:100%}.\\[\\&_\\[data-slot\\=monaco-editor\\]\\]\\:rounded-none [data-slot=monaco-editor]{border-radius:0}.\\[\\&_\\[data-slot\\=monaco-editor\\]\\]\\:border-0 [data-slot=monaco-editor]{border-style:var(--tw-border-style);border-width:0}.\\[\\&_\\[data-slot\\=monaco-editor\\]\\]\\:border-t [data-slot=monaco-editor]{border-top-style:var(--tw-border-style);border-top-width:1px}.\\[\\&_h2\\]\\:my-2 h2{margin-block:calc(var(--spacing) * 2)}.\\[\\&_p\\]\\:my-1 p{margin-block:var(--spacing)}.\\[\\&_pre\\]\\:m-0 pre{margin:0}.\\[\\&_pre\\]\\:inline pre{display:inline}.\\[\\&_pre\\]\\:bg-transparent pre{background-color:#0000}.\\[\\&_pre\\]\\:leading-relaxed pre{--tw-leading:var(--leading-relaxed);line-height:var(--leading-relaxed)}.\\[\\&_pre\\]\\:whitespace-pre pre{white-space:pre}.\\[\\&_pre\\]\\:whitespace-pre-wrap pre{white-space:pre-wrap}.\\[\\&_pre\\.shiki\\]\\:\\!bg-transparent pre.shiki{background-color:#0000!important}.\\[\\&_span\\[data-jsf-readonly\\]\\]\\:h-auto span[data-jsf-readonly]{height:auto}.\\[\\&_span\\[data-jsf-readonly\\]\\]\\:min-w-0 span[data-jsf-readonly]{min-width:0}.\\[\\&_svg\\]\\:pointer-events-none svg{pointer-events:none}.\\[\\&_svg\\]\\:size-3\\.5 svg{width:calc(var(--spacing) * 3.5);height:calc(var(--spacing) * 3.5)}.\\[\\&_svg\\]\\:size-4 svg{width:calc(var(--spacing) * 4);height:calc(var(--spacing) * 4)}.\\[\\&_svg\\]\\:size-5 svg{width:calc(var(--spacing) * 5);height:calc(var(--spacing) * 5)}.\\[\\&_svg\\]\\:size-\\[0\\.75em\\] svg{width:.75em;height:.75em}.\\[\\&_svg\\]\\:shrink-0 svg{flex-shrink:0}.\\[\\&_svg\\]\\:text-muted-foreground svg{color:var(--muted-foreground)}.\\[\\&\\:\\:-moz-range-thumb\\]\\:pointer-events-auto::-moz-range-thumb{pointer-events:auto}.\\[\\&\\:\\:-moz-range-thumb\\]\\:size-3\\.5::-moz-range-thumb{width:calc(var(--spacing) * 3.5);height:calc(var(--spacing) * 3.5)}.\\[\\&\\:\\:-moz-range-thumb\\]\\:appearance-none::-moz-range-thumb{appearance:none}.\\[\\&\\:\\:-moz-range-thumb\\]\\:rounded-full::-moz-range-thumb{border-radius:2147483647px}.\\[\\&\\:\\:-moz-range-thumb\\]\\:border::-moz-range-thumb{border-style:var(--tw-border-style);border-width:1px}.\\[\\&\\:\\:-moz-range-thumb\\]\\:border-0::-moz-range-thumb{border-style:var(--tw-border-style);border-width:0}.\\[\\&\\:\\:-moz-range-thumb\\]\\:border-background\\/80::-moz-range-thumb{border-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.\\[\\&\\:\\:-moz-range-thumb\\]\\:border-background\\/80::-moz-range-thumb{border-color:color-mix(in oklab, var(--background) 80%, transparent)}}.\\[\\&\\:\\:-moz-range-thumb\\]\\:bg-muted-foreground::-moz-range-thumb{background-color:var(--muted-foreground)}.\\[\\&\\:\\:-moz-range-thumb\\]\\:bg-primary::-moz-range-thumb{background-color:var(--primary)}.\\[\\&\\:\\:-moz-range-thumb\\]\\:shadow-sm::-moz-range-thumb{--tw-shadow:0 1px 3px 0 var(--tw-shadow-color,#0000001a), 0 1px 2px -1px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.\\[\\&\\:\\:-moz-range-track\\]\\:h-1::-moz-range-track{height:var(--spacing)}.\\[\\&\\:\\:-moz-range-track\\]\\:rounded-full::-moz-range-track{border-radius:2147483647px}.\\[\\&\\:\\:-moz-range-track\\]\\:bg-transparent::-moz-range-track{background-color:#0000}.\\[\\&\\:\\:-webkit-calendar-picker-indicator\\]\\:hidden::-webkit-calendar-picker-indicator{display:none}.\\[\\&\\:\\:-webkit-calendar-picker-indicator\\]\\:appearance-none::-webkit-calendar-picker-indicator{appearance:none}.\\[\\&\\:\\:-webkit-search-cancel-button\\]\\:appearance-none::-webkit-search-cancel-button{appearance:none}.\\[\\&\\:\\:-webkit-slider-runnable-track\\]\\:h-1::-webkit-slider-runnable-track{height:var(--spacing)}.\\[\\&\\:\\:-webkit-slider-runnable-track\\]\\:rounded-full::-webkit-slider-runnable-track{border-radius:2147483647px}.\\[\\&\\:\\:-webkit-slider-runnable-track\\]\\:bg-transparent::-webkit-slider-runnable-track{background-color:#0000}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:pointer-events-auto::-webkit-slider-thumb{pointer-events:auto}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:-mt-\\[5px\\]::-webkit-slider-thumb{margin-top:-5px}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:mt-\\[-5px\\]::-webkit-slider-thumb{margin-top:-5px}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:size-3\\.5::-webkit-slider-thumb{width:calc(var(--spacing) * 3.5);height:calc(var(--spacing) * 3.5)}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:appearance-none::-webkit-slider-thumb{appearance:none}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:rounded-full::-webkit-slider-thumb{border-radius:2147483647px}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:border::-webkit-slider-thumb{border-style:var(--tw-border-style);border-width:1px}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:border-background\\/80::-webkit-slider-thumb{border-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:border-background\\/80::-webkit-slider-thumb{border-color:color-mix(in oklab, var(--background) 80%, transparent)}}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:bg-muted-foreground::-webkit-slider-thumb{background-color:var(--muted-foreground)}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:bg-primary::-webkit-slider-thumb{background-color:var(--primary)}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:shadow-sm::-webkit-slider-thumb{--tw-shadow:0 1px 3px 0 var(--tw-shadow-color,#0000001a), 0 1px 2px -1px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.\\[\\&\\:not\\(\\:first-child\\)\\]\\:ml-2:not(:first-child){margin-left:calc(var(--spacing) * 2)}.\\[\\&\\>\\*\\:first-child\\]\\:mt-0>:first-child{margin-top:0}.\\[\\&\\>\\*\\:last-child\\]\\:mb-0>:last-child{margin-bottom:0}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\]\\:flex>[data-slot=modal-body]{display:flex}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\]\\:min-h-0>[data-slot=modal-body]{min-height:0}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\]\\:overflow-hidden>[data-slot=modal-body]{overflow:hidden}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\>\\*\\]\\:h-auto>[data-slot=modal-body]>*{height:auto}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\>\\*\\]\\:min-h-0>[data-slot=modal-body]>*{min-height:0}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\>\\*\\]\\:flex-1>[data-slot=modal-body]>*{flex:1}.\\[\\&\\>\\[data-slot\\=monaco-editor\\]\\]\\:h-full>[data-slot=monaco-editor]{height:100%}.\\[\\&\\>button\\]\\:w-full>button{width:100%}.\\[\\&\\>button\\]\\:justify-start>button{justify-content:flex-start}.\\[\\&\\>span\\]\\:w-full>span{width:100%}.\\[\\&\\>span\\]\\:min-w-0>span{min-width:0}.\\[\\&\\>svg\\]\\:h-4>svg{height:calc(var(--spacing) * 4)}.\\[\\&\\>svg\\]\\:w-4>svg{width:calc(var(--spacing) * 4)}:hover>.\\[\\:hover\\>\\&\\]\\:opacity-100{opacity:1}@media (hover:none){.\\[\\@media\\(hover\\:none\\)\\]\\:opacity-70{opacity:.7}.\\[\\@media\\(hover\\:none\\)\\]\\:opacity-100{opacity:1}}@media (pointer:coarse){.\\[\\@media\\(pointer\\:coarse\\)\\]\\:static{position:static}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:top-1{top:var(--spacing)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:right-1{right:var(--spacing)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:mt-2{margin-top:calc(var(--spacing) * 2)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:size-11{width:calc(var(--spacing) * 11);height:calc(var(--spacing) * 11)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:h-11{height:calc(var(--spacing) * 11)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:min-h-11{min-height:calc(var(--spacing) * 11)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:min-w-11{min-width:calc(var(--spacing) * 11)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:flex-wrap{flex-wrap:wrap}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:justify-end{justify-content:flex-end}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:gap-2{gap:calc(var(--spacing) * 2)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:p-2\\.5{padding:calc(var(--spacing) * 2.5)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:px-3{padding-inline:calc(var(--spacing) * 3)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:px-4{padding-inline:calc(var(--spacing) * 4)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:pr-12{padding-right:calc(var(--spacing) * 12)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:pb-2{padding-bottom:calc(var(--spacing) * 2)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:text-base{font-size:var(--text-base);line-height:var(--tw-leading,var(--text-base--line-height))}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:text-sm{font-size:var(--text-sm);line-height:var(--tw-leading,var(--text-sm--line-height))}}[aria-expanded=true] .\\[\\[aria-expanded\\=true\\]_\\&\\]\\:bg-muted{background-color:var(--muted)}[aria-selected=true]>div>.\\[\\[aria-selected\\=true\\]\\>div\\>\\&\\]\\:opacity-100{opacity:1}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-amber-400{border-color:var(--color-amber-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-emerald-400{border-color:var(--color-emerald-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-emerald-700{border-color:var(--color-emerald-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-fuchsia-400{border-color:var(--color-fuchsia-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-fuchsia-700{border-color:var(--color-fuchsia-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-indigo-400{border-color:var(--color-indigo-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-indigo-700{border-color:var(--color-indigo-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-orange-400{border-color:var(--color-orange-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-orange-700{border-color:var(--color-orange-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-pink-400{border-color:var(--color-pink-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-rose-400{border-color:var(--color-rose-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-rose-500\\/30{border-color:#ff23574d}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-rose-500\\/30{border-color:color-mix(in oklab, var(--color-rose-500) 30%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-rose-700{border-color:var(--color-rose-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-sky-400{border-color:var(--color-sky-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-sky-700{border-color:var(--color-sky-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-teal-400{border-color:var(--color-teal-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-teal-700{border-color:var(--color-teal-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-violet-400{border-color:var(--color-violet-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-violet-700{border-color:var(--color-violet-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-\\[var\\(--fs-bg-subtle\\)\\]{background-color:var(--fs-bg-subtle)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400{background-color:var(--color-amber-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/5{background-color:#fcbb000d}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/5{background-color:color-mix(in oklab, var(--color-amber-400) 5%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/10{background-color:#fcbb001a}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/10{background-color:color-mix(in oklab, var(--color-amber-400) 10%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/20{background-color:#fcbb0033}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/20{background-color:color-mix(in oklab, var(--color-amber-400) 20%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/25{background-color:#fcbb0040}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/25{background-color:color-mix(in oklab, var(--color-amber-400) 25%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/50{background-color:#fcbb0080}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/50{background-color:color-mix(in oklab, var(--color-amber-400) 50%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-500\\/15{background-color:#f99c0026}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-500\\/15{background-color:color-mix(in oklab, var(--color-amber-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-950\\/40{background-color:#46190166}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-950\\/40{background-color:color-mix(in oklab, var(--color-amber-950) 40%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-emerald-400{background-color:var(--color-emerald-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-emerald-500\\/15{background-color:#00bb7f26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-emerald-500\\/15{background-color:color-mix(in oklab, var(--color-emerald-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-emerald-950\\/60{background-color:#002c2299}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-emerald-950\\/60{background-color:color-mix(in oklab, var(--color-emerald-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-fuchsia-500\\/15{background-color:#e12afb26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-fuchsia-500\\/15{background-color:color-mix(in oklab, var(--color-fuchsia-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-fuchsia-950\\/60{background-color:#4b004f99}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-fuchsia-950\\/60{background-color:color-mix(in oklab, var(--color-fuchsia-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-green-500\\/20{background-color:#00c75833}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-green-500\\/20{background-color:color-mix(in oklab, var(--color-green-500) 20%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-indigo-500\\/15{background-color:#625fff26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-indigo-500\\/15{background-color:color-mix(in oklab, var(--color-indigo-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-indigo-950\\/60{background-color:#1e1a4d99}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-indigo-950\\/60{background-color:color-mix(in oklab, var(--color-indigo-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-orange-500\\/15{background-color:#fe6e0026}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-orange-500\\/15{background-color:color-mix(in oklab, var(--color-orange-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-orange-950\\/60{background-color:#44130699}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-orange-950\\/60{background-color:color-mix(in oklab, var(--color-orange-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-pink-500\\/15{background-color:#f6339a26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-pink-500\\/15{background-color:color-mix(in oklab, var(--color-pink-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-red-400{background-color:var(--color-red-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-red-500\\/20{background-color:#fb2c3633}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-red-500\\/20{background-color:color-mix(in oklab, var(--color-red-500) 20%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-500\\/10{background-color:#ff23571a}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-500\\/10{background-color:color-mix(in oklab, var(--color-rose-500) 10%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-500\\/15{background-color:#ff235726}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-500\\/15{background-color:color-mix(in oklab, var(--color-rose-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-950\\/60{background-color:#4d021899}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-950\\/60{background-color:color-mix(in oklab, var(--color-rose-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-sky-500\\/15{background-color:#00a5ef26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-sky-500\\/15{background-color:color-mix(in oklab, var(--color-sky-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-sky-950\\/60{background-color:#052f4a99}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-sky-950\\/60{background-color:color-mix(in oklab, var(--color-sky-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-teal-500\\/15{background-color:#00baa726}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-teal-500\\/15{background-color:color-mix(in oklab, var(--color-teal-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-teal-950\\/60{background-color:#022f2e99}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-teal-950\\/60{background-color:color-mix(in oklab, var(--color-teal-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-violet-500\\/15{background-color:#8d54ff26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-violet-500\\/15{background-color:color-mix(in oklab, var(--color-violet-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-violet-950\\/60{background-color:#2f0d6899}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-violet-950\\/60{background-color:color-mix(in oklab, var(--color-violet-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-\\[\\#20B2AA\\]{color:#20b2aa}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-\\[\\#D97757\\]{color:#d97757}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-\\[\\#F97316\\]{color:#f97316}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-\\[\\#FFB000\\]{color:#ffb000}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-amber-200{color:var(--color-amber-200)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-amber-300{color:var(--color-amber-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-amber-400{color:var(--color-amber-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-blue-300{color:var(--color-blue-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-blue-400{color:var(--color-blue-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-cyan-300{color:var(--color-cyan-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-emerald-300{color:var(--color-emerald-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-emerald-400{color:var(--color-emerald-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-fuchsia-300{color:var(--color-fuchsia-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-fuchsia-400{color:var(--color-fuchsia-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-green-300{color:var(--color-green-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-green-400{color:var(--color-green-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-indigo-300{color:var(--color-indigo-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-indigo-400{color:var(--color-indigo-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-orange-300{color:var(--color-orange-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-orange-400{color:var(--color-orange-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-pink-300{color:var(--color-pink-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-pink-400{color:var(--color-pink-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-purple-400{color:var(--color-purple-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-red-300{color:var(--color-red-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-red-400{color:var(--color-red-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-rose-300{color:var(--color-rose-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-rose-400{color:var(--color-rose-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-sky-300{color:var(--color-sky-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-sky-400{color:var(--color-sky-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-slate-300{color:var(--color-slate-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-slate-400{color:var(--color-slate-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-teal-300{color:var(--color-teal-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-teal-400{color:var(--color-teal-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-violet-300{color:var(--color-violet-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-violet-400{color:var(--color-violet-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-white{color:var(--color-white)}li .\\[li_\\&\\]\\:pl-6{padding-left:calc(var(--spacing) * 6)}}:root,[data-theme=light]{--fs-primary-blue:#3578e5;--fs-cta-blue:#0053f8;--fs-dark-navy:#131f3b;--fs-sky-500:#0ea5e9;--fs-indigo-500:#6366f1;--fs-gray-50:#f9fafb;--fs-gray-100:#f3f4f6;--fs-gray-200:#e5e7eb;--fs-gray-300:#d1d5db;--fs-gray-400:#9ca3af;--fs-gray-500:#6b7280;--fs-gray-600:#4b5563;--fs-gray-700:#374151;--fs-gray-800:#1f2937;--fs-gray-900:#111827;--fs-gray-950:#030712;--fs-success-50:#ecfdf5;--fs-success-100:#d1fae5;--fs-success-200:#a7f3d0;--fs-success-500:#10b981;--fs-success-600:#059669;--fs-success-700:#047857;--fs-success:var(--fs-success-600);--fs-error-50:#fff1f2;--fs-error-100:#ffe4e6;--fs-error-200:#fecdd3;--fs-error-500:#f43f5e;--fs-error-600:#e11d48;--fs-error-700:#be123c;--fs-error:var(--fs-error-600);--fs-warning-50:#fffbeb;--fs-warning-100:#fef3c7;--fs-warning-200:#fde68a;--fs-warning-500:#f59e0b;--fs-warning-600:#d97706;--fs-warning-700:#b45309;--fs-warning:var(--fs-warning-600);--fs-info-50:#f0f9ff;--fs-info-100:#e0f2fe;--fs-info-200:#bae6fd;--fs-info-500:#0ea5e9;--fs-info-600:#0284c7;--fs-info-700:#0369a1;--fs-info:var(--fs-info-600);--fs-hint-50:#f0fdfa;--fs-hint-100:#ccfbf1;--fs-hint-200:#99f6e4;--fs-hint-500:#14b8a6;--fs-hint-600:#0d9488;--fs-hint-700:#0f766e;--fs-hint:var(--fs-hint-600);--fs-prereq-50:#f5f3ff;--fs-prereq-100:#ede9fe;--fs-prereq-200:#ddd6fe;--fs-prereq-500:#8b5cf6;--fs-prereq-600:#7c3aed;--fs-prereq-700:#6d28d9;--fs-prereq:var(--fs-prereq-600);--fs-fg-1:#061431;--fs-fg-2:#030712;--fs-fg-3:#374151;--fs-fg-muted:#666;--fs-fg-subtle:#9ca3af;--fs-fg-on-brand:#fff;--fs-fg-link:var(--fs-primary-blue);--fs-bg-canvas:#fff;--fs-bg-subtle:#f9fafb;--fs-bg-muted:#f3f4f6;--fs-bg-inset:#e5e7eb;--fs-bg-inverse:var(--fs-dark-navy);--fs-bg-brand:var(--fs-primary-blue);--fs-bg-cta:var(--fs-cta-blue);--fs-border-subtle:#e5e7eb;--fs-border-default:#d1d5db;--fs-border-strong:#9ca3af;--fs-border-focus:var(--fs-primary-blue);--fs-font-sans:\"Open Sans\", -apple-system, BlinkMacSystemFont, Helvetica, Arial, sans-serif;--fs-font-serif:\"Merriweather\", Georgia, serif;--fs-font-mono:\"Fira Code\", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;--fs-text-xs:.8125rem;--fs-text-sm:.875rem;--fs-text-base:1rem;--fs-text-md:1rem;--fs-text-lg:1.25rem;--fs-text-xl:1.5rem;--fs-text-2xl:2rem;--fs-text-3xl:2.5rem;--fs-text-4xl:3.5rem;--fs-weight-regular:400;--fs-weight-medium:500;--fs-weight-semibold:600;--fs-weight-bold:700;--fs-lh-body:1.75;--fs-lh-tight:1.5;--fs-lh-heading:1.2;--fs-tracking-body:.03rem;--fs-tracking-heading:normal;--fs-tracking-wide:.08rem;--fs-radius-xs:2px;--fs-radius-sm:4px;--fs-radius-md:6px;--fs-radius-lg:8px;--fs-radius-xl:12px;--fs-radius-2xl:16px;--fs-radius-full:9999px;--fs-shadow-xs:0 1px 2px 0 #0614310d;--fs-shadow-sm:0 1px 2px 0 #0614310f, 0 1px 3px 0 #0614311a;--fs-shadow-md:0 4px 6px -1px #06143114, 0 2px 4px -2px #0614310f;--fs-shadow-lg:0 10px 15px -3px #06143114, 0 4px 6px -4px #0614310d;--fs-shadow-xl:0 20px 25px -5px #0614311a, 0 8px 10px -6px #0614310f;--fs-shadow-focus:0 0 0 3px #3578e540;--fs-space-1:4px;--fs-space-2:8px;--fs-space-3:12px;--fs-space-4:16px;--fs-space-5:20px;--fs-space-6:24px;--fs-space-8:32px;--fs-space-10:40px;--fs-space-12:48px;--fs-space-16:64px;--fs-space-20:80px;--fs-space-24:96px;--fs-ease-out:cubic-bezier(.16, 1, .3, 1);--fs-ease-in-out:cubic-bezier(.65, 0, .35, 1);--fs-duration-fast:.12s;--fs-duration-base:.18s;--fs-duration-slow:.28s}[data-theme=dark]{--fs-fg-1:#f9fafb;--fs-fg-2:#fff;--fs-fg-3:#d1d5db;--fs-fg-muted:#9ca3af;--fs-fg-subtle:#6b7280;--fs-bg-canvas:#030712;--fs-bg-subtle:#111827;--fs-bg-muted:#1f2937;--fs-bg-inset:#374151;--fs-border-subtle:#1f2937;--fs-border-default:#374151;--fs-border-strong:#4b5563;--fs-success-50:#022c22;--fs-success-100:#064e3b;--fs-success-200:#065f46;--fs-success-500:#34d399;--fs-success-600:#6ee7b7;--fs-success-700:#a7f3d0;--fs-success:var(--fs-success-500);--fs-error-50:#4c0519;--fs-error-100:#881337;--fs-error-200:#9f1239;--fs-error-500:#fb7185;--fs-error-600:#fda4af;--fs-error-700:#fecdd3;--fs-error:var(--fs-error-500);--fs-warning-50:#451a03;--fs-warning-100:#78350f;--fs-warning-200:#92400e;--fs-warning-500:#fbbf24;--fs-warning-600:#fcd34d;--fs-warning-700:#fde68a;--fs-warning:var(--fs-warning-500);--fs-info-50:#082f49;--fs-info-100:#0c4a6e;--fs-info-200:#075985;--fs-info-500:#38bdf8;--fs-info-600:#7dd3fc;--fs-info-700:#bae6fd;--fs-info:var(--fs-info-500);--fs-hint-50:#042f2e;--fs-hint-100:#134e4a;--fs-hint-200:#115e59;--fs-hint-500:#2dd4bf;--fs-hint-600:#5eead4;--fs-hint-700:#99f6e4;--fs-hint:var(--fs-hint-500);--fs-prereq-50:#2e1065;--fs-prereq-100:#4c1d95;--fs-prereq-200:#5b21b6;--fs-prereq-500:#a78bfa;--fs-prereq-600:#c4b5fd;--fs-prereq-700:#ddd6fe;--fs-prereq:var(--fs-prereq-500)}@layer clicky-tokens{:root,[data-theme]{--background:var(--fs-bg-canvas);--foreground:var(--fs-fg-1);--card:var(--fs-bg-canvas);--card-foreground:var(--fs-fg-1);--popover:var(--fs-bg-canvas);--popover-foreground:var(--fs-fg-1);--primary:var(--fs-primary-blue);--primary-foreground:var(--fs-fg-on-brand);--secondary:var(--fs-bg-muted);--secondary-foreground:var(--fs-fg-1);--muted:var(--fs-bg-muted);--muted-foreground:var(--fs-fg-muted);--placeholder:var(--fs-fg-subtle);--accent:var(--fs-bg-muted);--accent-foreground:var(--fs-fg-1);--destructive:var(--fs-error);--destructive-foreground:var(--fs-fg-on-brand);--border:var(--fs-border-subtle);--input:var(--fs-border-default);--ring:var(--fs-primary-blue);--radius:var(--fs-radius-lg);--sidebar:#1e2b48;--sidebar-foreground:#c2cce0;--sidebar-border:#33425f;--sidebar-accent:#2c3c5e;--sidebar-accent-foreground:#fff;--sidebar-primary:#7cb0ff;--sidebar-primary-foreground:#fff;--sidebar-ring:var(--fs-primary-blue)}:root,[data-density=comfortable]{--spacing-unit:.25rem;--control-height:2.25rem;--control-padding-x:.75rem;--font-size-base:.875rem}[data-density=compact]{--spacing-unit:.1875rem;--control-height:1.875rem;--control-padding-x:.5rem;--font-size-base:.8125rem}[data-density=spacious]{--spacing-unit:.3125rem;--control-height:2.75rem;--control-padding-x:1rem;--font-size-base:1rem}body{background-color:var(--background);color:var(--foreground);font-family:var(--font-sans)}.clicky-mdx-editor-field{--accentBase:var(--background);--accentBgSubtle:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBgSubtle:color-mix(in srgb, var(--primary) 6%, var(--background))}}.clicky-mdx-editor-field{--accentBg:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBg:color-mix(in srgb, var(--primary) 10%, var(--background))}}.clicky-mdx-editor-field{--accentBgHover:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBgHover:color-mix(in srgb, var(--primary) 14%, var(--background))}}.clicky-mdx-editor-field{--accentBgActive:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBgActive:color-mix(in srgb, var(--primary) 18%, var(--background))}}.clicky-mdx-editor-field{--accentLine:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentLine:color-mix(in srgb, var(--primary) 24%, var(--border))}}.clicky-mdx-editor-field{--accentBorder:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBorder:color-mix(in srgb, var(--primary) 34%, var(--border))}}.clicky-mdx-editor-field{--accentBorderHover:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBorderHover:color-mix(in srgb, var(--primary) 45%, var(--border))}}.clicky-mdx-editor-field{--accentSolid:var(--primary);--accentSolidHover:var(--primary);--accentText:var(--primary);--accentTextContrast:var(--primary-foreground);--basePageBg:var(--popover);--baseBase:var(--background);--baseBgSubtle:var(--muted)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--baseBgSubtle:color-mix(in srgb, var(--muted) 55%, var(--background))}}.clicky-mdx-editor-field{--baseBg:var(--muted);--baseBgHover:var(--accent);--baseBgActive:var(--accent)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--baseBgActive:color-mix(in srgb, var(--accent) 80%, var(--foreground) 8%)}}.clicky-mdx-editor-field{--baseLine:var(--border);--baseBorder:var(--border);--baseBorderHover:var(--input);--baseSolid:var(--muted-foreground);--baseSolidHover:var(--foreground);--baseText:var(--muted-foreground);--baseTextContrast:var(--foreground);--font-body:var(--fs-font-sans);--font-mono:var(--fs-font-mono);background-color:var(--background);border:1px solid var(--input);border-radius:var(--radius);color:var(--foreground);overflow:hidden}.clicky-mdx-editor-field:focus-within{box-shadow:0 0 0 2px var(--ring)}.clicky-mdx-editor-field-readonly{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field-readonly{background-color:color-mix(in srgb, var(--muted) 40%, transparent)}}.clicky-mdx-editor-field .mdxeditor-root-contenteditable{border-top:1px solid var(--border)}.clicky-mdx-editor-content{font-size:var(--font-size-base);line-height:1.55}.clicky-mdx-editor-content :where(p,ul,ol,blockquote,pre,table){margin-block:.5rem}.clicky-mdx-editor-content :where(h1,h2,h3,h4,h5,h6){color:var(--foreground);margin-block:.75rem .35rem;font-weight:600;line-height:1.25}.clicky-mdx-editor-content :where(code,pre){font-family:var(--fs-font-mono)}}.clicky-mdx-editor-content :where(ul,ol){list-style:revert;margin-block:.5rem;padding-inline-start:1.5rem}.stacktrace-source pre.shiki{font-size:inherit;line-height:inherit;counter-reset:shiki-line var(--shiki-start,0);white-space:normal;margin:0;padding:0;background:0 0!important}.stacktrace-source pre.shiki code{white-space:normal;display:block}.stacktrace-source pre.shiki .line{counter-increment:shiki-line;white-space:pre;min-height:1lh;padding-left:3.5rem;display:block;position:relative}.stacktrace-source pre.shiki .line:before{content:counter(shiki-line);text-align:right;width:3rem;color:var(--muted-foreground);-webkit-user-select:none;user-select:none;padding-right:.5rem;position:absolute;left:0}.stacktrace-source pre.shiki.has-highlighted .line{opacity:.55}.stacktrace-source pre.shiki .line.error{color:#b91c1c;opacity:1;background-color:#ef44441a;font-weight:700}[data-theme=dark] .stacktrace-source pre.shiki .line.error{color:#f87171}.stacktrace-source pre.shiki .line.error:before{color:#b91c1c}[data-theme=dark] .stacktrace-source pre.shiki .line.error:before{color:#f87171}@property --tw-border-spacing-x{syntax:\"<length>\";inherits:false;initial-value:0}@property --tw-border-spacing-y{syntax:\"<length>\";inherits:false;initial-value:0}@property --tw-translate-x{syntax:\"*\";inherits:false;initial-value:0}@property --tw-translate-y{syntax:\"*\";inherits:false;initial-value:0}@property --tw-translate-z{syntax:\"*\";inherits:false;initial-value:0}@property --tw-scale-x{syntax:\"*\";inherits:false;initial-value:1}@property --tw-scale-y{syntax:\"*\";inherits:false;initial-value:1}@property --tw-scale-z{syntax:\"*\";inherits:false;initial-value:1}@property --tw-rotate-x{syntax:\"*\";inherits:false}@property --tw-rotate-y{syntax:\"*\";inherits:false}@property --tw-rotate-z{syntax:\"*\";inherits:false}@property --tw-skew-x{syntax:\"*\";inherits:false}@property --tw-skew-y{syntax:\"*\";inherits:false}@property --tw-divide-x-reverse{syntax:\"*\";inherits:false;initial-value:0}@property --tw-divide-y-reverse{syntax:\"*\";inherits:false;initial-value:0}@property --tw-gradient-position{syntax:\"*\";inherits:false}@property --tw-gradient-from{syntax:\"<color>\";inherits:false;initial-value:#0000}@property --tw-gradient-via{syntax:\"<color>\";inherits:false;initial-value:#0000}@property --tw-gradient-to{syntax:\"<color>\";inherits:false;initial-value:#0000}@property --tw-gradient-stops{syntax:\"*\";inherits:false}@property --tw-gradient-via-stops{syntax:\"*\";inherits:false}@property --tw-gradient-from-position{syntax:\"<length-percentage>\";inherits:false;initial-value:0%}@property --tw-gradient-via-position{syntax:\"<length-percentage>\";inherits:false;initial-value:50%}@property --tw-gradient-to-position{syntax:\"<length-percentage>\";inherits:false;initial-value:100%}@property --tw-shadow{syntax:\"*\";inherits:false;initial-value:0 0 #0000}@property --tw-shadow-color{syntax:\"*\";inherits:false}@property --tw-shadow-alpha{syntax:\"<percentage>\";inherits:false;initial-value:100%}@property --tw-inset-shadow{syntax:\"*\";inherits:false;initial-value:0 0 #0000}@property --tw-inset-shadow-color{syntax:\"*\";inherits:false}@property --tw-inset-shadow-alpha{syntax:\"<percentage>\";inherits:false;initial-value:100%}@property --tw-ring-color{syntax:\"*\";inherits:false}@property --tw-ring-shadow{syntax:\"*\";inherits:false;initial-value:0 0 #0000}@property --tw-inset-ring-color{syntax:\"*\";inherits:false}@property --tw-inset-ring-shadow{syntax:\"*\";inherits:false;initial-value:0 0 #0000}@property --tw-ring-inset{syntax:\"*\";inherits:false}@property --tw-ring-offset-width{syntax:\"<length>\";inherits:false;initial-value:0}@property --tw-ring-offset-color{syntax:\"*\";inherits:false;initial-value:#fff}@property --tw-ring-offset-shadow{syntax:\"*\";inherits:false;initial-value:0 0 #0000}@property --tw-backdrop-blur{syntax:\"*\";inherits:false}@property --tw-backdrop-brightness{syntax:\"*\";inherits:false}@property --tw-backdrop-contrast{syntax:\"*\";inherits:false}@property --tw-backdrop-grayscale{syntax:\"*\";inherits:false}@property --tw-backdrop-hue-rotate{syntax:\"*\";inherits:false}@property --tw-backdrop-invert{syntax:\"*\";inherits:false}@property --tw-backdrop-opacity{syntax:\"*\";inherits:false}@property --tw-backdrop-saturate{syntax:\"*\";inherits:false}@property --tw-backdrop-sepia{syntax:\"*\";inherits:false}@property --tw-duration{syntax:\"*\";inherits:false}@property --tw-ease{syntax:\"*\";inherits:false}@property --tw-content{syntax:\"*\";inherits:false;initial-value:\"\"}@keyframes spin{to{transform:rotate(360deg)}}@keyframes pulse{50%{opacity:.5}}html,body,#root{height:100%}@property --tw-space-y-reverse{syntax:\"*\";inherits:false;initial-value:0}@property --tw-border-style{syntax:\"*\";inherits:false;initial-value:solid}@property --tw-leading{syntax:\"*\";inherits:false}@property --tw-font-weight{syntax:\"*\";inherits:false}@property --tw-tracking{syntax:\"*\";inherits:false}@property --tw-ordinal{syntax:\"*\";inherits:false}@property --tw-slashed-zero{syntax:\"*\";inherits:false}@property --tw-numeric-figure{syntax:\"*\";inherits:false}@property --tw-numeric-spacing{syntax:\"*\";inherits:false}@property --tw-numeric-fraction{syntax:\"*\";inherits:false}@property --tw-outline-style{syntax:\"*\";inherits:false;initial-value:solid}@property --tw-blur{syntax:\"*\";inherits:false}@property --tw-brightness{syntax:\"*\";inherits:false}@property --tw-contrast{syntax:\"*\";inherits:false}@property --tw-grayscale{syntax:\"*\";inherits:false}@property --tw-hue-rotate{syntax:\"*\";inherits:false}@property --tw-invert{syntax:\"*\";inherits:false}@property --tw-opacity{syntax:\"*\";inherits:false}@property --tw-saturate{syntax:\"*\";inherits:false}@property --tw-sepia{syntax:\"*\";inherits:false}@property --tw-drop-shadow{syntax:\"*\";inherits:false}@property --tw-drop-shadow-color{syntax:\"*\";inherits:false}@property --tw-drop-shadow-alpha{syntax:\"<percentage>\";inherits:false;initial-value:100%}@property --tw-drop-shadow-size{syntax:\"*\";inherits:false}", I = /* @__PURE__ */ e(n(), 1), L = r(), Xt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "m13.236 9.307l-4.429 2.491a1.5 1.5 0 0 0-.781-.704l.01-.01l4.709-2.649a.5.5 0 0 0 0-.872l-8-4.499A.5.5 0 0 0 4 3.5V7c-.354 0-.684.081-1 .194V3.5a1.5 1.5 0 0 1 2.235-1.307l8 4.5c1.019.573 1.019 2.041 0 2.615zM7 10.707V12h.5a.5.5 0 0 1 0 1H7c0 .384-.073.75-.205 1.088l1.059 1.059a.5.5 0 0 1-.708.707l-.885-.885a2.995 2.995 0 0 1-4.522 0l-.885.885a.5.5 0 0 1-.708 0a.5.5 0 0 1 0-.707l1.059-1.059A3 3 0 0 1 1 13H.5a.5.5 0 0 1 0-1H1v-1.293l-.854-.854a.5.5 0 0 1 .707-.707l.854.854H2c0-1.103.897-2 2-2s2 .897 2 2h.293l.854-.854a.5.5 0 0 1 .707.707zM3 10h2a1 1 0 0 0-2 0m3 1H2v2c0 1.103.897 2 2 2s2-.897 2-2z"
	})
}), {
	__source: "codicon:debug-alt-small",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "debug-small",
	displayName: "UiDebugSmall"
}), Zt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M128 26a102 102 0 1 0 102 102A102.12 102.12 0 0 0 128 26m0 192a90 90 0 1 1 90-90a90.1 90.1 0 0 1-90 90"
	})
}), {
	__source: "ph:circle-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "circle-filled",
	displayName: "UiCircleFilled"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M232 128A104 104 0 1 1 128 24a104.13 104.13 0 0 1 104 104"
	})
}), {
	__source: "ph:circle-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "circle-filled",
	displayName: "UiCircleFilledFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiDebug.tsx
var Qt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M146 92a10 10 0 1 1 10 10a10 10 0 0 1-10-10m-46-10a10 10 0 1 0 10 10a10 10 0 0 0-10-10m146 46a6 6 0 0 1-6 6h-26v10a85.9 85.9 0 0 1-3.45 24.08l23.85 10.42a6 6 0 0 1-4.8 11l-23.23-10.15a86 86 0 0 1-156.74 0L26.4 189.5a6 6 0 1 1-4.8-11l23.85-10.42A85.9 85.9 0 0 1 42 144v-10H16a6 6 0 0 1 0-12h26v-10a85.9 85.9 0 0 1 3.45-24.08L21.6 77.5a6 6 0 0 1 4.8-11l23.23 10.15a86 86 0 0 1 156.74 0L229.6 66.5a6 6 0 1 1 4.8 11l-23.85 10.42A85.9 85.9 0 0 1 214 112v10h26a6 6 0 0 1 6 6m-192-6h148v-10a74 74 0 0 0-148 0Zm68 95.74V134H54v10a74.09 74.09 0 0 0 68 73.74M202 134h-68v83.74A74.09 74.09 0 0 0 202 144Z"
	})
}), {
	__source: "ph:bug-light",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "debug",
	displayName: "UiDebug"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M168 92a12 12 0 1 1-12-12a12 12 0 0 1 12 12m-68-12a12 12 0 1 0 12 12a12 12 0 0 0-12-12m116 64a87.8 87.8 0 0 1-3 23l22.24 9.72A8 8 0 0 1 232 192a7.9 7.9 0 0 1-3.2-.67L207.38 182a88 88 0 0 1-158.76 0l-21.42 9.33a7.9 7.9 0 0 1-3.2.67a8 8 0 0 1-3.2-15.33L43 167a87.8 87.8 0 0 1-3-23v-8H16a8 8 0 0 1 0-16h24v-8a87.8 87.8 0 0 1 3-23l-22.2-9.67a8 8 0 1 1 6.4-14.66L48.62 74a88 88 0 0 1 158.76 0l21.42-9.36a8 8 0 0 1 6.4 14.66L213 89.05a87.8 87.8 0 0 1 3 23v8h24a8 8 0 0 1 0 16h-24Zm-80 0a8 8 0 0 0-16 0v64a8 8 0 0 0 16 0Zm64-32a72 72 0 0 0-144 0v8h144Z"
	})
}), {
	__source: "ph:bug-fill",
	__viewBox: "0 0 256 256",
	__group: "health-status",
	__consumerName: "debug",
	displayName: "UiDebugFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiListTree.tsx
var $t = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M160 110h48a14 14 0 0 0 14-14V48a14 14 0 0 0-14-14h-48a14 14 0 0 0-14 14v18h-18a22 22 0 0 0-22 22v34H70v-10a14 14 0 0 0-14-14H24a14 14 0 0 0-14 14v32a14 14 0 0 0 14 14h32a14 14 0 0 0 14-14v-10h36v34a22 22 0 0 0 22 22h18v18a14 14 0 0 0 14 14h48a14 14 0 0 0 14-14v-48a14 14 0 0 0-14-14h-48a14 14 0 0 0-14 14v18h-18a10 10 0 0 1-10-10V88a10 10 0 0 1 10-10h18v18a14 14 0 0 0 14 14M58 144a2 2 0 0 1-2 2H24a2 2 0 0 1-2-2v-32a2 2 0 0 1 2-2h32a2 2 0 0 1 2 2Zm100 16a2 2 0 0 1 2-2h48a2 2 0 0 1 2 2v48a2 2 0 0 1-2 2h-48a2 2 0 0 1-2-2Zm0-112a2 2 0 0 1 2-2h48a2 2 0 0 1 2 2v48a2 2 0 0 1-2 2h-48a2 2 0 0 1-2-2Z"
	})
}), {
	__source: "ph:tree-structure-light",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "list-tree",
	displayName: "UiListTree"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M144 96V80h-16a8 8 0 0 0-8 8v80a8 8 0 0 0 8 8h16v-16a16 16 0 0 1 16-16h48a16 16 0 0 1 16 16v48a16 16 0 0 1-16 16h-48a16 16 0 0 1-16-16v-16h-16a24 24 0 0 1-24-24v-32H72v8a16 16 0 0 1-16 16H24a16 16 0 0 1-16-16v-32a16 16 0 0 1 16-16h32a16 16 0 0 1 16 16v8h32V88a24 24 0 0 1 24-24h16V48a16 16 0 0 1 16-16h48a16 16 0 0 1 16 16v48a16 16 0 0 1-16 16h-48a16 16 0 0 1-16-16"
	})
}), {
	__source: "ph:tree-structure-fill",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "list-tree",
	displayName: "UiListTreeFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiTable.tsx
var en = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M224 50H32a6 6 0 0 0-6 6v136a14 14 0 0 0 14 14h176a14 14 0 0 0 14-14V56a6 6 0 0 0-6-6M38 110h44v36H38Zm56 0h124v36H94Zm124-48v36H38V62ZM38 192v-34h44v36H40a2 2 0 0 1-2-2m178 2H94v-36h124v34a2 2 0 0 1-2 2"
	})
}), {
	__source: "ph:table-light",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "table",
	displayName: "UiTable"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M224 48H32a8 8 0 0 0-8 8v136a16 16 0 0 0 16 16h176a16 16 0 0 0 16-16V56a8 8 0 0 0-8-8M40 112h40v32H40Zm56 0h120v32H96Zm-56 48h40v32H40Zm176 32H96v-32h120z"
	})
}), {
	__source: "ph:table-fill",
	__viewBox: "0 0 256 256",
	__group: "trees-lists-tables",
	__consumerName: "table",
	displayName: "UiTableFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiPlay.tsx
var tn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("g", {
		fill: "none",
		children: /* @__PURE__ */ (0, L.jsx)("path", {
			d: "M13.5 7.13397C14.1667 7.51888 14.1667 8.48112 13.5 8.86602L4.5 14.0622C3.83333 14.4471 3 13.966 3 13.1962L3 2.80385C3 2.03405 3.83333 1.55292 4.5 1.93782L13.5 7.13397Z",
			fill: "#F2FCF3",
			stroke: "#208A3C"
		})
	})
}), {
	__source: "jb-expui-run:run",
	__viewBox: "0 0 16 16",
	__group: "runtime-process",
	__consumerName: "play",
	displayName: "UiPlay"
}), nn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, L.jsx)("rect", {
			x: "3.5",
			y: "2.5",
			width: "3",
			height: "11",
			rx: "0.5",
			stroke: "#6C707E"
		}), /* @__PURE__ */ (0, L.jsx)("rect", {
			x: "9.5",
			y: "2.5",
			width: "3",
			height: "11",
			rx: "0.5",
			stroke: "#6C707E"
		})]
	})
}), {
	__source: "jb-expui-run:pause",
	__viewBox: "0 0 16 16",
	__group: "runtime-process",
	__consumerName: "pause",
	displayName: "UiPause"
}), rn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M134 80v44.6l37.09 22.25a6 6 0 0 1-6.18 10.3l-40-24A6 6 0 0 1 122 128V80a6 6 0 0 1 12 0m-6-46a93.4 93.4 0 0 0-66.49 27.56c-8.58 8.68-16 17-23.51 25.8V64a6 6 0 0 0-12 0v40a6 6 0 0 0 6 6h40a6 6 0 0 0 0-12H44.73C52.86 88.29 60.79 79.35 70 70a82 82 0 1 1 1.7 117.62a6 6 0 1 0-8.24 8.72A94 94 0 1 0 128 34"
	})
}), {
	__source: "ph:clock-counter-clockwise-light",
	__viewBox: "0 0 256 256",
	__group: "runtime-process",
	__consumerName: "history",
	displayName: "UiHistory"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M224 128a96 96 0 0 1-161.89 69.82a8 8 0 1 1 11-11.64a80 80 0 1 0-1.68-114.75A439 439 0 0 0 61.35 82l16.31 16.34A8 8 0 0 1 72 112H32a8 8 0 0 1-8-8V64a8 8 0 0 1 13.66-5.66L50 70.7c3.22-3.49 6.54-7 10.06-10.55A96 96 0 0 1 224 128m-96-56a8 8 0 0 0-8 8v48a8 8 0 0 0 3.88 6.86l40 24a8 8 0 1 0 8.24-13.72L136 123.47V80a8 8 0 0 0-8-8"
	})
}), {
	__source: "ph:clock-counter-clockwise-fill",
	__viewBox: "0 0 256 256",
	__group: "runtime-process",
	__consumerName: "history",
	displayName: "UiHistoryFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiTag.tsx
var an = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M241.91 137.42L142.59 38.1a13.94 13.94 0 0 0-9.9-4.1H40a6 6 0 0 0-6 6v92.69a13.94 13.94 0 0 0 4.1 9.9l99.32 99.32a14 14 0 0 0 19.8 0l84.69-84.69a14 14 0 0 0 0-19.8m-8.49 11.31l-84.69 84.69a2 2 0 0 1-2.83 0L46.59 134.1a2 2 0 0 1-.59-1.41V46h86.69a2 2 0 0 1 1.41.59l99.32 99.31a2 2 0 0 1 0 2.83M94 84a10 10 0 1 1-10-10a10 10 0 0 1 10 10"
	})
}), {
	__source: "ph:tag-light",
	__viewBox: "0 0 256 256",
	__group: "configs-metadata",
	__consumerName: "tag",
	displayName: "UiTag"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M243.31 136L144 36.69A15.86 15.86 0 0 0 132.69 32H40a8 8 0 0 0-8 8v92.69A15.86 15.86 0 0 0 36.69 144L136 243.31a16 16 0 0 0 22.63 0l84.68-84.68a16 16 0 0 0 0-22.63M84 96a12 12 0 1 1 12-12a12 12 0 0 1-12 12"
	})
}), {
	__source: "ph:tag-fill",
	__viewBox: "0 0 256 256",
	__group: "configs-metadata",
	__consumerName: "tag",
	displayName: "UiTagFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiServerProcess.tsx
var on = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M208 138H48a14 14 0 0 0-14 14v48a14 14 0 0 0 14 14h160a14 14 0 0 0 14-14v-48a14 14 0 0 0-14-14m2 62a2 2 0 0 1-2 2H48a2 2 0 0 1-2-2v-48a2 2 0 0 1 2-2h160a2 2 0 0 1 2 2Zm-2-158H48a14 14 0 0 0-14 14v48a14 14 0 0 0 14 14h160a14 14 0 0 0 14-14V56a14 14 0 0 0-14-14m2 62a2 2 0 0 1-2 2H48a2 2 0 0 1-2-2V56a2 2 0 0 1 2-2h160a2 2 0 0 1 2 2Zm-20-24a10 10 0 1 1-10-10a10 10 0 0 1 10 10m0 96a10 10 0 1 1-10-10a10 10 0 0 1 10 10"
	})
}), {
	__source: "ph:hard-drives-light",
	__viewBox: "0 0 256 256",
	__group: "infrastructure",
	__consumerName: "server-process",
	displayName: "UiServerProcess"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M208 40H48a16 16 0 0 0-16 16v48a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16V56a16 16 0 0 0-16-16m-28 52a12 12 0 1 1 12-12a12 12 0 0 1-12 12m28 44H48a16 16 0 0 0-16 16v48a16 16 0 0 0 16 16h160a16 16 0 0 0 16-16v-48a16 16 0 0 0-16-16m-28 52a12 12 0 1 1 12-12a12 12 0 0 1-12 12"
	})
}), {
	__source: "ph:hard-drives-fill",
	__viewBox: "0 0 256 256",
	__group: "infrastructure",
	__consumerName: "server-process",
	displayName: "UiServerProcessFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiChartLineUp.tsx
var sn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M230 208a6 6 0 0 1-6 6H32a6 6 0 0 1-6-6V48a6 6 0 0 1 12 0v113.52l53.76-53.76a6 6 0 0 1 8.48 0L128 135.51L185.52 78H160a6 6 0 0 1 0-12h40a6 6 0 0 1 6 6v40a6 6 0 0 1-12 0V86.48l-61.76 61.76a6 6 0 0 1-8.48 0L96 120.49l-58 58V202h186a6 6 0 0 1 6 6"
	})
}), {
	__source: "ph:chart-line-up-light",
	__viewBox: "0 0 256 256",
	__group: "data-analytics",
	__consumerName: "chart-line-up",
	displayName: "UiChartLineUp"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, L.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, L.jsx)("path", {
		fill: "currentColor",
		d: "M216 40H40a16 16 0 0 0-16 16v144a16 16 0 0 0 16 16h176a16 16 0 0 0 16-16V56a16 16 0 0 0-16-16m-16 152H56a8 8 0 0 1-8-8V72a8 8 0 0 1 16 0v76.69l34.34-34.35a8 8 0 0 1 11.32 0L128 132.69L172.69 88H144a8 8 0 0 1 0-16h48a8 8 0 0 1 8 8v48a8 8 0 0 1-16 0V99.31l-50.34 50.35a8 8 0 0 1-11.32 0L104 131.31l-40 40V176h136a8 8 0 0 1 0 16"
	})
}), {
	__source: "ph:chart-line-up-fill",
	__viewBox: "0 0 256 256",
	__group: "data-analytics",
	__consumerName: "chart-line-up",
	displayName: "UiChartLineUpFilled"
});
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/lib/duration.ts
var cn = [
	"ns",
	"us",
	"µs",
	"μs",
	"ms",
	"s",
	"m",
	"h"
], ln = {
	ns: 1 / 1e6,
	us: 1 / 1e3,
	µs: 1 / 1e3,
	μs: 1 / 1e3,
	ms: 1,
	s: 1e3,
	m: 6e4,
	h: 36e5
};
function un(e, t) {
	let n = t, r = 0;
	for (; pn(e[n]);) r++, n++;
	let i = 0;
	if (e[n] === ".") for (n++; pn(e[n]);) i++, n++;
	return r === 0 && i === 0 ? null : {
		value: Number(e.slice(t, n)),
		next: n
	};
}
function dn(e, t) {
	for (let n of cn) if (e.startsWith(n, t)) return n;
	return null;
}
function fn(e) {
	let t = 0, n = 0;
	for (; t < e.length;) {
		let r = un(e, t), i = r && dn(e, r.next);
		if (!r || !i) return null;
		n += r.value * ln[i], t = r.next + i.length;
	}
	return t === 0 ? null : n;
}
function pn(e) {
	return e !== void 0 && e >= "0" && e <= "9";
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/data-table-filter-values.ts
function mn(e) {
	return e === "\\" || e === "!" || e === "," || /\s/.test(e);
}
function hn(e) {
	return !e.escaped && /\s/.test(e.char);
}
function gn(e) {
	let t = [[]];
	for (let n = 0; n < e.length; n++) {
		let r = e[n], i = e[n + 1];
		r === "\\" && i !== void 0 && mn(i) ? (t[t.length - 1].push({
			char: i,
			escaped: !0
		}), n++) : r === "," ? t.push([]) : t[t.length - 1].push({
			char: r,
			escaped: !1
		});
	}
	return t.map((e) => {
		let t = 0, n = e.length;
		for (; t < n && hn(e[t]);) t++;
		for (; n > t && hn(e[n - 1]);) n--;
		return e.slice(t, n);
	});
}
function _n(e) {
	let t = {};
	for (let n of gn(e)) {
		if (n.length === 0) continue;
		let e = n.length > 1 && n[0].char === "!" && !n[0].escaped, r = (e ? n.slice(1) : n).map((e) => e.char).join("");
		r && (t[r] = e ? "exclude" : "include");
	}
	return t;
}
function vn(e) {
	let t = "";
	for (let n = 0; n < e.length; n++) {
		let r = e[n], i = n === 0 || n === e.length - 1;
		r === "," ? t += "\\," : r === "\\" && mn(e[n + 1] ?? "") ? t += "\\\\" : i && /\s/.test(r) ? t += `\\${r}` : t += r;
	}
	return t;
}
function yn(e) {
	return Object.entries(e).flatMap(([e, t]) => {
		if (!e) return [];
		let n = vn(e);
		return t === "include" ? [e.startsWith("!") ? `\\${n}` : n] : t === "exclude" ? [`!${n}`] : [];
	}).join(",");
}
function bn(e) {
	return e.split(",").map((e) => e.trim()).filter(Boolean);
}
function xn(e, t) {
	let n = {};
	t && !e.trim() && (t.startsWith(">") ? n.minOperator = t : n.maxOperator = t);
	for (let r of e.split(",")) {
		let e = r.trim();
		if (!e) continue;
		let i = /^(>=|>|<=|<)(.+)$/.exec(e) ?? (t ? [
			e,
			t,
			e
		] : null);
		if (!i) continue;
		let [, a, o] = i;
		a === ">" || a === ">=" ? (n.min = o.trim(), n.minOperator = a) : (n.max = o.trim(), n.maxOperator = a);
	}
	return n;
}
function Sn(e) {
	let t = [];
	return e.min?.trim() && t.push(`${e.minOperator ?? ">="}${e.min.trim()}`), e.max?.trim() && t.push(`${e.maxOperator ?? "<="}${e.max.trim()}`), t.join(",");
}
function Cn(e, t = "ms", n) {
	let r = xn(e, n), i = {
		...r.minOperator ? { minOperator: r.minOperator } : {},
		...r.maxOperator ? { maxOperator: r.maxOperator } : {}
	};
	return r.min !== void 0 && (i.min = Tn(r.min, t), i.minUnit = t), r.max !== void 0 && (i.max = Tn(r.max, t), i.maxUnit = t), i;
}
function wn(e, t = "ms") {
	return Sn({
		...e.min?.trim() ? {
			min: `${e.min.trim()}${e.minUnit ?? t}`,
			...e.minOperator ? { minOperator: e.minOperator } : {}
		} : {},
		...e.max?.trim() ? {
			max: `${e.max.trim()}${e.maxUnit ?? t}`,
			...e.maxOperator ? { maxOperator: e.maxOperator } : {}
		} : {}
	});
}
function Tn(e, t) {
	let n = e.trim(), r = n[0] === "+" || n[0] === "-", i = n[0] === "-" ? -1 : 1, a = r ? n.slice(1) : n;
	if (un(a, 0)?.next === a.length) return n;
	let o = fn(a);
	if (o === null) throw Error(`Invalid duration bound ${e}`);
	return En(i * o / ln[t]);
}
function En(e) {
	return String(Number.isInteger(e) ? e : Number(e.toFixed(9)));
}
// ../../../../../../pnpm/store/v11/links/@tanstack/query-core/5.101.1/7e19765689cbccf24c1f78df4f522110979ff2e4db631a13f55549f0e91ee24f/node_modules/@tanstack/query-core/build/modern/subscribable.cjs
var Dn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { Subscribable: () => u }), t.exports = c(l);
	var u = class {
		constructor() {
			this.listeners = /* @__PURE__ */ new Set(), this.subscribe = this.subscribe.bind(this);
		}
		subscribe(e) {
			return this.listeners.add(e), this.onSubscribe(), () => {
				this.listeners.delete(e), this.onUnsubscribe();
			};
		}
		hasListeners() {
			return this.listeners.size > 0;
		}
		onSubscribe() {}
		onUnsubscribe() {}
	};
	0 && (t.exports = { Subscribable: u });
})), On = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		FocusManager: () => d,
		focusManager: () => f
	}), t.exports = c(l);
	var u = Dn(), d = class extends u.Subscribable {
		#e;
		#t;
		#n;
		constructor() {
			super(), this.#n = (e) => {
				if (typeof window < "u" && window.addEventListener) {
					let t = () => e();
					return window.addEventListener("visibilitychange", t, !1), () => {
						window.removeEventListener("visibilitychange", t);
					};
				}
			};
		}
		onSubscribe() {
			this.#t || this.setEventListener(this.#n);
		}
		onUnsubscribe() {
			this.hasListeners() || (this.#t?.(), this.#t = void 0);
		}
		setEventListener(e) {
			this.#n = e, this.#t?.(), this.#t = e((e) => {
				typeof e == "boolean" ? this.setFocused(e) : this.onFocus();
			});
		}
		setFocused(e) {
			this.#e !== e && (this.#e = e, this.onFocus());
		}
		onFocus() {
			let e = this.isFocused();
			this.listeners.forEach((t) => {
				t(e);
			});
		}
		isFocused() {
			return typeof this.#e == "boolean" ? this.#e : globalThis.document?.visibilityState !== "hidden";
		}
	}, f = new d();
	0 && (t.exports = {
		FocusManager: d,
		focusManager: f
	});
})), kn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		TimeoutManager: () => d,
		defaultTimeoutProvider: () => u,
		systemSetTimeoutZero: () => p,
		timeoutManager: () => f
	}), t.exports = c(l);
	var u = {
		setTimeout: (e, t) => setTimeout(e, t),
		clearTimeout: (e) => clearTimeout(e),
		setInterval: (e, t) => setInterval(e, t),
		clearInterval: (e) => clearInterval(e)
	}, d = class {
		#e = u;
		setTimeoutProvider(e) {
			this.#e = e;
		}
		setTimeout(e, t) {
			return this.#e.setTimeout(e, t);
		}
		clearTimeout(e) {
			this.#e.clearTimeout(e);
		}
		setInterval(e, t) {
			return this.#e.setInterval(e, t);
		}
		clearInterval(e) {
			this.#e.clearInterval(e);
		}
	}, f = new d();
	function p(e) {
		setTimeout(e, 0);
	}
	0 && (t.exports = {
		TimeoutManager: d,
		defaultTimeoutProvider: u,
		systemSetTimeoutZero: p,
		timeoutManager: f
	});
})), An = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		addConsumeAwareSignal: () => ae,
		addToEnd: () => re,
		addToStart: () => ie,
		ensureQueryFn: () => A,
		functionalUpdate: () => p,
		hashKey: () => x,
		hashQueryKeyByOptions: () => b,
		isPlainArray: () => E,
		isPlainObject: () => D,
		isServer: () => d,
		isValidTimeout: () => m,
		keepPreviousData: () => ne,
		matchMutation: () => y,
		matchQuery: () => v,
		noop: () => f,
		partialMatchKey: () => S,
		replaceData: () => te,
		replaceEqualDeep: () => w,
		resolveQueryBoolean: () => _,
		resolveStaleTime: () => g,
		shallowEqualObjects: () => T,
		shouldThrowError: () => j,
		skipToken: () => k,
		sleep: () => ee,
		timeUntilStale: () => h
	}), t.exports = c(l);
	var u = kn(), d = typeof window > "u" || "Deno" in globalThis;
	function f() {}
	function p(e, t) {
		return typeof e == "function" ? e(t) : e;
	}
	function m(e) {
		return typeof e == "number" && e >= 0 && e !== Infinity;
	}
	function h(e, t) {
		return Math.max(e + (t || 0) - Date.now(), 0);
	}
	function g(e, t) {
		return typeof e == "function" ? e(t) : e;
	}
	function _(e, t) {
		return typeof e == "function" ? e(t) : e;
	}
	function v(e, t) {
		let { type: n = "all", exact: r, fetchStatus: i, predicate: a, queryKey: o, stale: s } = e;
		if (o) {
			if (r) {
				if (t.queryHash !== b(o, t.options)) return !1;
			} else if (!S(t.queryKey, o)) return !1;
		}
		if (n !== "all") {
			let e = t.isActive();
			if (n === "active" && !e || n === "inactive" && e) return !1;
		}
		return !(typeof s == "boolean" && t.isStale() !== s || i && i !== t.state.fetchStatus || a && !a(t));
	}
	function y(e, t) {
		let { exact: n, status: r, predicate: i, mutationKey: a } = e;
		if (a) {
			if (!t.options.mutationKey) return !1;
			if (n) {
				if (x(t.options.mutationKey) !== x(a)) return !1;
			} else if (!S(t.options.mutationKey, a)) return !1;
		}
		return !(r && t.state.status !== r || i && !i(t));
	}
	function b(e, t) {
		return (t?.queryKeyHashFn || x)(e);
	}
	function x(e) {
		return JSON.stringify(e, (e, t) => D(t) ? Object.keys(t).sort().reduce((e, n) => (e[n] = t[n], e), {}) : t);
	}
	function S(e, t) {
		return e === t ? !0 : typeof e == typeof t && e && t && typeof e == "object" && typeof t == "object" ? Object.keys(t).every((n) => S(e[n], t[n])) : !1;
	}
	var C = Object.prototype.hasOwnProperty;
	function w(e, t, n = 0) {
		if (e === t) return e;
		if (n > 500) return t;
		let r = E(e) && E(t);
		if (!r && !(D(e) && D(t))) return t;
		let i = (r ? e : Object.keys(e)).length, a = r ? t : Object.keys(t), o = a.length, s = r ? Array(o) : {}, c = 0;
		for (let l = 0; l < o; l++) {
			let o = r ? l : a[l], u = e[o], d = t[o];
			if (u === d) {
				s[o] = u, (r ? l < i : C.call(e, o)) && c++;
				continue;
			}
			if (u === null || d === null || typeof u != "object" || typeof d != "object") {
				s[o] = d;
				continue;
			}
			let f = w(u, d, n + 1);
			s[o] = f, f === u && c++;
		}
		return i === o && c === i ? e : s;
	}
	function T(e, t) {
		if (!t || Object.keys(e).length !== Object.keys(t).length) return !1;
		for (let n in e) if (e[n] !== t[n]) return !1;
		return !0;
	}
	function E(e) {
		return Array.isArray(e) && e.length === Object.keys(e).length;
	}
	function D(e) {
		if (!O(e)) return !1;
		let t = e.constructor;
		if (t === void 0) return !0;
		let n = t.prototype;
		return !(!O(n) || !n.hasOwnProperty("isPrototypeOf") || Object.getPrototypeOf(e) !== Object.prototype);
	}
	function O(e) {
		return Object.prototype.toString.call(e) === "[object Object]";
	}
	function ee(e) {
		return new Promise((t) => {
			u.timeoutManager.setTimeout(t, e);
		});
	}
	function te(e, t, n) {
		return typeof n.structuralSharing == "function" ? n.structuralSharing(e, t) : n.structuralSharing === !1 ? t : w(e, t);
	}
	function ne(e) {
		return e;
	}
	function re(e, t, n = 0) {
		let r = [...e, t];
		return n && r.length > n ? r.slice(1) : r;
	}
	function ie(e, t, n = 0) {
		let r = [t, ...e];
		return n && r.length > n ? r.slice(0, -1) : r;
	}
	var k = /* @__PURE__ */ Symbol();
	function A(e, t) {
		return !e.queryFn && t?.initialPromise ? () => t.initialPromise : !e.queryFn || e.queryFn === k ? () => Promise.reject(/* @__PURE__ */ Error(`Missing queryFn: '${e.queryHash}'`)) : e.queryFn;
	}
	function j(e, t) {
		return typeof e == "function" ? e(...t) : !!e;
	}
	function ae(e, t, n) {
		let r = !1, i;
		return Object.defineProperty(e, "signal", {
			enumerable: !0,
			get: () => (i ??= t(), r || (r = !0, i.aborted ? n() : i.addEventListener("abort", n, { once: !0 })), i)
		}), e;
	}
	0 && (t.exports = {
		addConsumeAwareSignal: ae,
		addToEnd: re,
		addToStart: ie,
		ensureQueryFn: A,
		functionalUpdate: p,
		hashKey: x,
		hashQueryKeyByOptions: b,
		isPlainArray: E,
		isPlainObject: D,
		isServer: d,
		isValidTimeout: m,
		keepPreviousData: ne,
		matchMutation: y,
		matchQuery: v,
		noop: f,
		partialMatchKey: S,
		replaceData: te,
		replaceEqualDeep: w,
		resolveQueryBoolean: _,
		resolveStaleTime: g,
		shallowEqualObjects: T,
		shouldThrowError: j,
		skipToken: k,
		sleep: ee,
		timeUntilStale: h
	});
})), jn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { environmentManager: () => d }), t.exports = c(l);
	var u = An(), d = /* @__PURE__ */ (() => {
		let e = () => u.isServer;
		return {
			isServer() {
				return e();
			},
			setIsServer(t) {
				e = t;
			}
		};
	})();
	0 && (t.exports = { environmentManager: d });
})), Mn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		pendingThenable: () => d,
		tryResolveSync: () => f
	}), t.exports = c(l);
	var u = An();
	function d() {
		let e, t, n = new Promise((n, r) => {
			e = n, t = r;
		});
		n.status = "pending", n.catch(() => {});
		function r(e) {
			Object.assign(n, e), delete n.resolve, delete n.reject;
		}
		return n.resolve = (t) => {
			r({
				status: "fulfilled",
				value: t
			}), e(t);
		}, n.reject = (e) => {
			r({
				status: "rejected",
				reason: e
			}), t(e);
		}, n;
	}
	function f(e) {
		let t;
		if (e.then((e) => (t = e, e), u.noop)?.catch(u.noop), t !== void 0) return { data: t };
	}
	0 && (t.exports = {
		pendingThenable: d,
		tryResolveSync: f
	});
})), Nn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		defaultShouldDehydrateMutation: () => h,
		defaultShouldDehydrateQuery: () => g,
		dehydrate: () => v,
		hydrate: () => y
	}), t.exports = c(l);
	var u = Mn(), d = An();
	function f(e) {
		return e;
	}
	function p(e) {
		return {
			mutationKey: e.options.mutationKey,
			state: e.state,
			...e.options.scope && { scope: e.options.scope },
			...e.meta && { meta: e.meta }
		};
	}
	function m(e, t, n) {
		let r = () => {
			let r = e.promise?.then(t).catch((e) => n(e) ? Promise.reject(/* @__PURE__ */ Error("redacted")) : Promise.reject(e));
			return r?.catch(d.noop), r;
		};
		return {
			dehydratedAt: Date.now(),
			state: {
				...e.state,
				...e.state.data !== void 0 && { data: t(e.state.data) }
			},
			queryKey: e.queryKey,
			queryHash: e.queryHash,
			...e.state.status === "pending" && { promise: r() },
			...e.meta && { meta: e.meta },
			...e.queryType && { queryType: e.queryType }
		};
	}
	function h(e) {
		return e.state.isPaused;
	}
	function g(e) {
		return e.state.status === "success";
	}
	function _(e) {
		return !0;
	}
	function v(e, t = {}) {
		let n = t.shouldDehydrateMutation ?? e.getDefaultOptions().dehydrate?.shouldDehydrateMutation ?? h, r = e.getMutationCache().getAll().flatMap((e) => n(e) ? [p(e)] : []), i = t.shouldDehydrateQuery ?? e.getDefaultOptions().dehydrate?.shouldDehydrateQuery ?? g, a = t.shouldRedactErrors ?? e.getDefaultOptions().dehydrate?.shouldRedactErrors ?? _, o = t.serializeData ?? e.getDefaultOptions().dehydrate?.serializeData ?? f;
		return {
			mutations: r,
			queries: e.getQueryCache().getAll().flatMap((e) => i(e) ? [m(e, o, a)] : [])
		};
	}
	function y(e, t, n) {
		if (typeof t != "object" || !t) return;
		let r = e.getMutationCache(), i = e.getQueryCache(), a = n?.defaultOptions?.deserializeData ?? e.getDefaultOptions().hydrate?.deserializeData ?? f, o = t.mutations || [], s = t.queries || [];
		o.forEach(({ state: t, ...i }) => {
			r.build(e, {
				...e.getDefaultOptions().hydrate?.mutations,
				...n?.defaultOptions?.mutations,
				...i
			}, t);
		}), s.forEach(({ queryKey: t, state: r, queryHash: o, meta: s, promise: c, dehydratedAt: l, queryType: f }) => {
			let p = c ? (0, u.tryResolveSync)(c) : void 0, m = r.data === void 0 ? p?.data : r.data, h = m === void 0 ? m : a(m), g = i.get(o), _ = g?.state.status === "pending", v = g?.state.fetchStatus === "fetching";
			if (g) {
				let e = p && l !== void 0 && l > g.state.dataUpdatedAt;
				if (r.dataUpdatedAt > g.state.dataUpdatedAt || e) {
					let { fetchStatus: e, ...t } = r;
					g.setState({
						...t,
						data: h,
						...r.status === "pending" && h !== void 0 && {
							status: "success",
							dataUpdatedAt: l ?? Date.now(),
							...!v && { fetchStatus: "idle" }
						}
					});
				}
			} else g = i.build(e, {
				...e.getDefaultOptions().hydrate?.queries,
				...n?.defaultOptions?.queries,
				queryKey: t,
				queryHash: o,
				meta: s,
				_type: f
			}, {
				...r,
				data: h,
				fetchStatus: "idle",
				status: r.status === "pending" && h !== void 0 ? "success" : r.status,
				...r.status === "pending" && h !== void 0 && { dataUpdatedAt: l ?? Date.now() }
			});
			c && !p && !_ && !v && (l === void 0 || l > g.state.dataUpdatedAt) && g.fetch(void 0, { initialPromise: Promise.resolve(c).then(a) }).catch(d.noop);
		});
	}
	0 && (t.exports = {
		defaultShouldDehydrateMutation: h,
		defaultShouldDehydrateQuery: g,
		dehydrate: v,
		hydrate: y
	});
})), Pn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		createNotifyManager: () => d,
		defaultScheduler: () => u,
		notifyManager: () => f
	}), t.exports = c(l);
	var u = kn().systemSetTimeoutZero;
	function d() {
		let e = [], t = 0, n = (e) => {
			e();
		}, r = (e) => {
			e();
		}, i = u, a = (r) => {
			t ? e.push(r) : i(() => {
				n(r);
			});
		}, o = () => {
			let t = e;
			e = [], t.length && i(() => {
				r(() => {
					t.forEach((e) => {
						n(e);
					});
				});
			});
		};
		return {
			batch: (e) => {
				let n;
				t++;
				try {
					n = e();
				} finally {
					t--, t || o();
				}
				return n;
			},
			batchCalls: (e) => (...t) => {
				a(() => {
					e(...t);
				});
			},
			schedule: a,
			setNotifyFunction: (e) => {
				n = e;
			},
			setBatchNotifyFunction: (e) => {
				r = e;
			},
			setScheduler: (e) => {
				i = e;
			}
		};
	}
	var f = d();
	0 && (t.exports = {
		createNotifyManager: d,
		defaultScheduler: u,
		notifyManager: f
	});
})), Fn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		OnlineManager: () => d,
		onlineManager: () => f
	}), t.exports = c(l);
	var u = Dn(), d = class extends u.Subscribable {
		#e = !0;
		#t;
		#n;
		constructor() {
			super(), this.#n = (e) => {
				if (typeof window < "u" && window.addEventListener) {
					let t = () => e(!0), n = () => e(!1);
					return window.addEventListener("online", t, !1), window.addEventListener("offline", n, !1), () => {
						window.removeEventListener("online", t), window.removeEventListener("offline", n);
					};
				}
			};
		}
		onSubscribe() {
			this.#t || this.setEventListener(this.#n);
		}
		onUnsubscribe() {
			this.hasListeners() || (this.#t?.(), this.#t = void 0);
		}
		setEventListener(e) {
			this.#n = e, this.#t?.(), this.#t = e(this.setOnline.bind(this));
		}
		setOnline(e) {
			this.#e !== e && (this.#e = e, this.listeners.forEach((t) => {
				t(e);
			}));
		}
		isOnline() {
			return this.#e;
		}
	}, f = new d();
	0 && (t.exports = {
		OnlineManager: d,
		onlineManager: f
	});
})), In = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		CancelledError: () => _,
		canFetch: () => g,
		createRetryer: () => y,
		isCancelledError: () => v
	}), t.exports = c(l);
	var u = On(), d = Fn(), f = Mn(), p = jn(), m = An();
	function h(e) {
		return Math.min(1e3 * 2 ** e, 3e4);
	}
	function g(e) {
		return (e ?? "online") !== "online" || d.onlineManager.isOnline();
	}
	var _ = class extends Error {
		constructor(e) {
			super("CancelledError"), this.revert = e?.revert, this.silent = e?.silent;
		}
	};
	function v(e) {
		return e instanceof _;
	}
	function y(e) {
		let t = !1, n = 0, r, i = (0, f.pendingThenable)(), a = () => i.status !== "pending", o = (t) => {
			if (!a()) {
				let n = new _(t);
				b(n), e.onCancel?.(n);
			}
		}, s = () => {
			t = !0;
		}, c = () => {
			t = !1;
		}, l = () => u.focusManager.isFocused() && (e.networkMode === "always" || d.onlineManager.isOnline()) && e.canRun(), v = () => g(e.networkMode) && e.canRun(), y = (e) => {
			a() || (r?.(), i.resolve(e));
		}, b = (e) => {
			a() || (r?.(), i.reject(e));
		}, x = () => new Promise((t) => {
			r = (e) => {
				(a() || l()) && t(e);
			}, e.onPause?.();
		}).then(() => {
			r = void 0, a() || e.onContinue?.();
		}), S = () => {
			if (a()) return;
			let r, i = n === 0 ? e.initialPromise : void 0;
			try {
				r = i ?? e.fn();
			} catch (e) {
				r = Promise.reject(e);
			}
			Promise.resolve(r).then(y).catch((r) => {
				if (a()) return;
				let i = e.retry ?? (p.environmentManager.isServer() ? 0 : 3), o = e.retryDelay ?? h, s = typeof o == "function" ? o(n, r) : o, c = i === !0 || typeof i == "number" && n < i || typeof i == "function" && i(n, r);
				t || !c ? b(r) : (n++, e.onFail?.(n, r), (0, m.sleep)(s).then(() => l() ? void 0 : x()).then(() => {
					t ? b(r) : S();
				}));
			});
		};
		return {
			promise: i,
			status: () => i.status,
			cancel: o,
			continue: () => (r?.(), i),
			cancelRetry: s,
			continueRetry: c,
			canStart: v,
			start: () => (v() ? S() : x().then(S), i)
		};
	}
	0 && (t.exports = {
		CancelledError: _,
		canFetch: g,
		createRetryer: y,
		isCancelledError: v
	});
})), Ln = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { Removable: () => p }), t.exports = c(l);
	var u = kn(), d = jn(), f = An(), p = class {
		#e;
		destroy() {
			this.clearGcTimeout();
		}
		scheduleGc() {
			this.clearGcTimeout(), (0, f.isValidTimeout)(this.gcTime) && (this.#e = u.timeoutManager.setTimeout(() => {
				this.optionalRemove();
			}, this.gcTime));
		}
		updateGcTime(e) {
			this.gcTime = Math.max(this.gcTime || 0, e ?? (d.environmentManager.isServer() ? Infinity : 3e5));
		}
		clearGcTimeout() {
			this.#e !== void 0 && (u.timeoutManager.clearTimeout(this.#e), this.#e = void 0);
		}
	};
	0 && (t.exports = { Removable: p });
})), Rn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		hasNextPage: () => m,
		hasPreviousPage: () => h,
		infiniteQueryBehavior: () => d
	}), t.exports = c(l);
	var u = An();
	function d(e) {
		return { onFetch: (t, n) => {
			let r = t.options, i = t.fetchOptions?.meta?.fetchMore?.direction, a = t.state.data?.pages || [], o = t.state.data?.pageParams || [], s = {
				pages: [],
				pageParams: []
			}, c = 0, l = async () => {
				let n = !1, l = (e) => {
					(0, u.addConsumeAwareSignal)(e, () => t.signal, () => n = !0);
				}, d = (0, u.ensureQueryFn)(t.options, t.fetchOptions), m = async (e, r, i) => {
					if (n) return Promise.reject(t.signal.reason);
					if (r == null && e.pages.length) return Promise.resolve(e);
					let a = (() => {
						let e = {
							client: t.client,
							queryKey: t.queryKey,
							pageParam: r,
							direction: i ? "backward" : "forward",
							meta: t.options.meta
						};
						return l(e), e;
					})(), o = await d(a), { maxPages: s } = t.options, c = i ? u.addToStart : u.addToEnd;
					return {
						pages: c(e.pages, o, s),
						pageParams: c(e.pageParams, r, s)
					};
				};
				if (i && a.length) {
					let e = i === "backward", t = e ? p : f, n = {
						pages: a,
						pageParams: o
					};
					s = await m(n, t(r, n), e);
				} else {
					let t = e ?? a.length;
					do {
						let e = c === 0 ? o[0] ?? r.initialPageParam : f(r, s);
						if (c > 0 && e == null) break;
						s = await m(s, e), c++;
					} while (c < t);
				}
				return s;
			};
			t.fetchFn = t.options.persister ? () => t.options.persister?.(l, {
				client: t.client,
				queryKey: t.queryKey,
				meta: t.options.meta,
				signal: t.signal
			}, n) : l;
		} };
	}
	function f(e, { pages: t, pageParams: n }) {
		let r = t.length - 1;
		return t.length > 0 ? e.getNextPageParam(t[r], t, n[r], n) : void 0;
	}
	function p(e, { pages: t, pageParams: n }) {
		return t.length > 0 ? e.getPreviousPageParam?.(t[0], t, n[0], n) : void 0;
	}
	function m(e, t) {
		return t ? f(e, t) != null : !1;
	}
	function h(e, t) {
		return !t || !e.getPreviousPageParam ? !1 : p(e, t) != null;
	}
	0 && (t.exports = {
		hasNextPage: m,
		hasPreviousPage: h,
		infiniteQueryBehavior: d
	});
})), zn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		Query: () => h,
		fetchState: () => g
	}), t.exports = c(l);
	var u = An(), d = Pn(), f = In(), p = Ln(), m = Rn(), h = class extends p.Removable {
		#e;
		#t;
		#n;
		#r;
		#i;
		#a;
		#o;
		#s;
		constructor(e) {
			super(), this.#s = !1, this.#o = e.defaultOptions, this.setOptions(e.options), this.observers = [], this.#i = e.client, this.#r = this.#i.getQueryCache(), this.queryKey = e.queryKey, this.queryHash = e.queryHash, this.#t = v(this.options), this.state = e.state ?? this.#t, this.scheduleGc();
		}
		get meta() {
			return this.options.meta;
		}
		get queryType() {
			return this.#e;
		}
		get promise() {
			return this.#a?.promise;
		}
		setOptions(e) {
			if (this.options = {
				...this.#o,
				...e
			}, e?._type && (this.#e = e._type), this.updateGcTime(this.options.gcTime), this.state && this.state.data === void 0) {
				let e = v(this.options);
				e.data !== void 0 && (this.setState(_(e.data, e.dataUpdatedAt)), this.#t = e);
			}
		}
		optionalRemove() {
			!this.observers.length && this.state.fetchStatus === "idle" && this.#r.remove(this);
		}
		setData(e, t) {
			let n = (0, u.replaceData)(this.state.data, e, this.options);
			return this.#l({
				data: n,
				type: "success",
				dataUpdatedAt: t?.updatedAt,
				manual: t?.manual
			}), n;
		}
		setState(e) {
			this.#l({
				type: "setState",
				state: e
			});
		}
		cancel(e) {
			let t = this.#a?.promise;
			return this.#a?.cancel(e), t ? t.then(u.noop).catch(u.noop) : Promise.resolve();
		}
		destroy() {
			super.destroy(), this.cancel({ silent: !0 });
		}
		get resetState() {
			return this.#t;
		}
		reset() {
			this.destroy(), this.setState(this.resetState);
		}
		isActive() {
			return this.observers.some((e) => (0, u.resolveQueryBoolean)(e.options.enabled, this) !== !1);
		}
		isDisabled() {
			return this.getObserversCount() > 0 ? !this.isActive() : this.options.queryFn === u.skipToken || !this.isFetched();
		}
		isFetched() {
			return this.state.dataUpdateCount + this.state.errorUpdateCount > 0;
		}
		isStatic() {
			return this.getObserversCount() > 0 && this.observers.some((e) => (0, u.resolveStaleTime)(e.options.staleTime, this) === "static");
		}
		isStale() {
			return this.getObserversCount() > 0 ? this.observers.some((e) => e.getCurrentResult().isStale) : this.state.data === void 0 || this.state.isInvalidated;
		}
		isStaleByTime(e = 0) {
			return this.state.data === void 0 ? !0 : e === "static" ? !1 : this.state.isInvalidated ? !0 : !(0, u.timeUntilStale)(this.state.dataUpdatedAt, e);
		}
		onFocus() {
			this.observers.find((e) => e.shouldFetchOnWindowFocus())?.refetch({ cancelRefetch: !1 }), this.#a?.continue();
		}
		onOnline() {
			this.observers.find((e) => e.shouldFetchOnReconnect())?.refetch({ cancelRefetch: !1 }), this.#a?.continue();
		}
		addObserver(e) {
			this.observers.includes(e) || (this.observers.push(e), this.clearGcTimeout(), this.#r.notify({
				type: "observerAdded",
				query: this,
				observer: e
			}));
		}
		removeObserver(e) {
			this.observers.includes(e) && (this.observers = this.observers.filter((t) => t !== e), this.observers.length || (this.#a && (this.#s || this.#c() ? this.#a.cancel({ revert: !0 }) : this.#a.cancelRetry()), this.scheduleGc()), this.#r.notify({
				type: "observerRemoved",
				query: this,
				observer: e
			}));
		}
		getObserversCount() {
			return this.observers.length;
		}
		#c() {
			return this.state.fetchStatus === "paused" && this.state.status === "pending";
		}
		invalidate() {
			this.state.isInvalidated || this.#l({ type: "invalidate" });
		}
		async fetch(e, t) {
			if (this.state.fetchStatus !== "idle" && this.#a?.status() !== "rejected") {
				if (this.state.data !== void 0 && t?.cancelRefetch) this.cancel({ silent: !0 });
				else if (this.#a) return this.#a.continueRetry(), this.#a.promise;
			}
			if (e && this.setOptions(e), !this.options.queryFn) {
				let e = this.observers.find((e) => e.options.queryFn);
				e && this.setOptions(e.options);
			}
			let n = new AbortController(), r = (e) => {
				Object.defineProperty(e, "signal", {
					enumerable: !0,
					get: () => (this.#s = !0, n.signal)
				});
			}, i = () => {
				let e = (0, u.ensureQueryFn)(this.options, t), n = (() => {
					let e = {
						client: this.#i,
						queryKey: this.queryKey,
						meta: this.meta
					};
					return r(e), e;
				})();
				return this.#s = !1, this.options.persister ? this.options.persister(e, n, this) : e(n);
			}, a = (() => {
				let e = {
					fetchOptions: t,
					options: this.options,
					queryKey: this.queryKey,
					client: this.#i,
					state: this.state,
					fetchFn: i
				};
				return r(e), e;
			})();
			(this.#e === "infinite" ? (0, m.infiniteQueryBehavior)(this.options.pages) : this.options.behavior)?.onFetch(a, this), this.#n = this.state, (this.state.fetchStatus === "idle" || this.state.fetchMeta !== a.fetchOptions?.meta) && this.#l({
				type: "fetch",
				meta: a.fetchOptions?.meta
			}), this.#a = (0, f.createRetryer)({
				initialPromise: t?.initialPromise,
				fn: a.fetchFn,
				onCancel: (e) => {
					e instanceof f.CancelledError && e.revert && this.setState({
						...this.#n,
						fetchStatus: "idle"
					}), n.abort();
				},
				onFail: (e, t) => {
					this.#l({
						type: "failed",
						failureCount: e,
						error: t
					});
				},
				onPause: () => {
					this.#l({ type: "pause" });
				},
				onContinue: () => {
					this.#l({ type: "continue" });
				},
				retry: a.options.retry,
				retryDelay: a.options.retryDelay,
				networkMode: a.options.networkMode,
				canRun: () => !0
			});
			try {
				let e = await this.#a.start();
				if (e === void 0) throw Error(`${this.queryHash} data is undefined`);
				return this.setData(e), this.#r.config.onSuccess?.(e, this), this.#r.config.onSettled?.(e, this.state.error, this), e;
			} catch (e) {
				if (e instanceof f.CancelledError) {
					if (e.silent) return this.#a.promise;
					if (e.revert) {
						if (this.state.data === void 0) throw e;
						return this.state.data;
					}
				}
				throw this.#l({
					type: "error",
					error: e
				}), this.#r.config.onError?.(e, this), this.#r.config.onSettled?.(this.state.data, e, this), e;
			} finally {
				this.scheduleGc();
			}
		}
		#l(e) {
			let t = (t) => {
				switch (e.type) {
					case "failed": return {
						...t,
						fetchFailureCount: e.failureCount,
						fetchFailureReason: e.error
					};
					case "pause": return {
						...t,
						fetchStatus: "paused"
					};
					case "continue": return {
						...t,
						fetchStatus: "fetching"
					};
					case "fetch": return {
						...t,
						...g(t.data, this.options),
						fetchMeta: e.meta ?? null
					};
					case "success":
						let n = {
							...t,
							..._(e.data, e.dataUpdatedAt),
							dataUpdateCount: t.dataUpdateCount + 1,
							...!e.manual && {
								fetchStatus: "idle",
								fetchFailureCount: 0,
								fetchFailureReason: null
							}
						};
						return this.#n = e.manual ? n : void 0, n;
					case "error":
						let r = e.error;
						return {
							...t,
							error: r,
							errorUpdateCount: t.errorUpdateCount + 1,
							errorUpdatedAt: Date.now(),
							fetchFailureCount: t.fetchFailureCount + 1,
							fetchFailureReason: r,
							fetchStatus: "idle",
							status: "error",
							isInvalidated: !0
						};
					case "invalidate": return {
						...t,
						isInvalidated: !0
					};
					case "setState": return {
						...t,
						...e.state
					};
				}
			};
			this.state = t(this.state), d.notifyManager.batch(() => {
				this.observers.forEach((e) => {
					e.onQueryUpdate();
				}), this.#r.notify({
					query: this,
					type: "updated",
					action: e
				});
			});
		}
	};
	function g(e, t) {
		return {
			fetchFailureCount: 0,
			fetchFailureReason: null,
			fetchStatus: (0, f.canFetch)(t.networkMode) ? "fetching" : "paused",
			...e === void 0 && {
				error: null,
				status: "pending"
			}
		};
	}
	function _(e, t) {
		return {
			data: e,
			dataUpdatedAt: t ?? Date.now(),
			error: null,
			isInvalidated: !1,
			status: "success"
		};
	}
	function v(e) {
		let t = typeof e.initialData == "function" ? e.initialData() : e.initialData, n = t !== void 0, r = n ? typeof e.initialDataUpdatedAt == "function" ? e.initialDataUpdatedAt() : e.initialDataUpdatedAt : 0;
		return {
			data: t,
			dataUpdateCount: 0,
			dataUpdatedAt: n ? r ?? Date.now() : 0,
			error: null,
			errorUpdateCount: 0,
			errorUpdatedAt: 0,
			fetchFailureCount: 0,
			fetchFailureReason: null,
			fetchMeta: null,
			isInvalidated: !1,
			status: n ? "success" : "pending",
			fetchStatus: "idle"
		};
	}
	0 && (t.exports = {
		Query: h,
		fetchState: g
	});
})), Bn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { QueryObserver: () => v }), t.exports = c(l);
	var u = On(), d = jn(), f = Pn(), p = zn(), m = Dn(), h = Mn(), g = An(), _ = kn(), v = class extends m.Subscribable {
		constructor(e, t) {
			super(), this.options = t, this.#e = e, this.#s = null, this.#o = (0, h.pendingThenable)(), this.bindMethods(), this.setOptions(t);
		}
		#e;
		#t = void 0;
		#n = void 0;
		#r = void 0;
		#i;
		#a;
		#o;
		#s;
		#c;
		#l;
		#u;
		#d;
		#f;
		#p;
		#m = /* @__PURE__ */ new Set();
		bindMethods() {
			this.refetch = this.refetch.bind(this);
		}
		onSubscribe() {
			this.listeners.size === 1 && (this.#t.addObserver(this), b(this.#t, this.options) ? this.#h() : this.updateResult(), this.#y());
		}
		onUnsubscribe() {
			this.hasListeners() || this.destroy();
		}
		shouldFetchOnReconnect() {
			return x(this.#t, this.options, this.options.refetchOnReconnect);
		}
		shouldFetchOnWindowFocus() {
			return x(this.#t, this.options, this.options.refetchOnWindowFocus);
		}
		destroy() {
			this.listeners = /* @__PURE__ */ new Set(), this.#b(), this.#x(), this.#t.removeObserver(this);
		}
		setOptions(e) {
			let t = this.options, n = this.#t;
			if (this.options = this.#e.defaultQueryOptions(e), this.options.enabled !== void 0 && typeof this.options.enabled != "boolean" && typeof this.options.enabled != "function" && typeof (0, g.resolveQueryBoolean)(this.options.enabled, this.#t) != "boolean") throw Error("Expected enabled to be a boolean or a callback that returns a boolean");
			this.#S(), this.#t.setOptions(this.options), t._defaulted && !(0, g.shallowEqualObjects)(this.options, t) && this.#e.getQueryCache().notify({
				type: "observerOptionsUpdated",
				query: this.#t,
				observer: this
			});
			let r = this.hasListeners();
			r && S(this.#t, n, this.options, t) && this.#h(), this.updateResult(), r && (this.#t !== n || (0, g.resolveQueryBoolean)(this.options.enabled, this.#t) !== (0, g.resolveQueryBoolean)(t.enabled, this.#t) || (0, g.resolveStaleTime)(this.options.staleTime, this.#t) !== (0, g.resolveStaleTime)(t.staleTime, this.#t)) && this.#g();
			let i = this.#_();
			r && (this.#t !== n || (0, g.resolveQueryBoolean)(this.options.enabled, this.#t) !== (0, g.resolveQueryBoolean)(t.enabled, this.#t) || i !== this.#p) && this.#v(i);
		}
		getOptimisticResult(e) {
			let t = this.#e.getQueryCache().build(this.#e, e), n = this.createResult(t, e);
			return w(this, n) && (this.#r = n, this.#a = this.options, this.#i = this.#t.state), n;
		}
		getCurrentResult() {
			return this.#r;
		}
		trackResult(e, t) {
			return new Proxy(e, { get: (e, n) => (this.trackProp(n), t?.(n), n === "promise" && (this.trackProp("data"), !this.options.experimental_prefetchInRender && this.#o.status === "pending" && this.#o.reject(/* @__PURE__ */ Error("experimental_prefetchInRender feature flag is not enabled"))), Reflect.get(e, n)) });
		}
		trackProp(e) {
			this.#m.add(e);
		}
		getCurrentQuery() {
			return this.#t;
		}
		refetch({ ...e } = {}) {
			return this.fetch({ ...e });
		}
		fetchOptimistic(e) {
			let t = this.#e.defaultQueryOptions(e), n = this.#e.getQueryCache().build(this.#e, t);
			return n.fetch().then(() => this.createResult(n, t));
		}
		fetch(e) {
			return this.#h({
				...e,
				cancelRefetch: e.cancelRefetch ?? !0
			}).then(() => (this.updateResult(), this.#r));
		}
		#h(e) {
			this.#S();
			let t = this.#t.fetch(this.options, e);
			return e?.throwOnError || (t = t.catch(g.noop)), t;
		}
		#g() {
			this.#b();
			let e = (0, g.resolveStaleTime)(this.options.staleTime, this.#t);
			if (d.environmentManager.isServer() || this.#r.isStale || !(0, g.isValidTimeout)(e)) return;
			let t = (0, g.timeUntilStale)(this.#r.dataUpdatedAt, e) + 1;
			this.#d = _.timeoutManager.setTimeout(() => {
				this.#r.isStale || this.updateResult();
			}, t);
		}
		#_() {
			return (typeof this.options.refetchInterval == "function" ? this.options.refetchInterval(this.#t) : this.options.refetchInterval) ?? !1;
		}
		#v(e) {
			this.#x(), this.#p = e, !d.environmentManager.isServer() && (0, g.resolveQueryBoolean)(this.options.enabled, this.#t) !== !1 && (0, g.isValidTimeout)(this.#p) && this.#p !== 0 && (this.#f = _.timeoutManager.setInterval(() => {
				(this.options.refetchIntervalInBackground || u.focusManager.isFocused()) && this.#h();
			}, this.#p));
		}
		#y() {
			this.#g(), this.#v(this.#_());
		}
		#b() {
			this.#d !== void 0 && (_.timeoutManager.clearTimeout(this.#d), this.#d = void 0);
		}
		#x() {
			this.#f !== void 0 && (_.timeoutManager.clearInterval(this.#f), this.#f = void 0);
		}
		createResult(e, t) {
			let n = this.#t, r = this.options, i = this.#r, a = this.#i, o = this.#a, s = e === n ? this.#n : e.state, { state: c } = e, l = { ...c }, u = !1, d;
			if (t._optimisticResults) {
				let i = this.hasListeners(), a = !i && b(e, t), o = i && S(e, n, t, r);
				(a || o) && (l = {
					...l,
					...(0, p.fetchState)(c.data, e.options)
				}), t._optimisticResults === "isRestoring" && (l.fetchStatus = "idle");
			}
			let { error: f, errorUpdatedAt: m, status: _ } = l;
			d = l.data;
			let v = !1;
			if (t.placeholderData !== void 0 && d === void 0 && _ === "pending") {
				let e;
				i?.isPlaceholderData && t.placeholderData === o?.placeholderData ? (e = i.data, v = !0) : e = typeof t.placeholderData == "function" ? t.placeholderData(this.#u?.state.data, this.#u) : t.placeholderData, e !== void 0 && (_ = "success", d = (0, g.replaceData)(i?.data, e, t), u = !0);
			}
			if (t.select && d !== void 0 && !v) {
				if (i && d === a?.data && t.select === this.#c) d = this.#l;
				else try {
					this.#c = t.select, d = t.select(d), d = (0, g.replaceData)(i?.data, d, t), this.#l = d, this.#s = null;
				} catch (e) {
					this.#s = e;
				}
			}
			this.#s && (f = this.#s, d = this.#l, m = Date.now(), _ = "error");
			let y = l.fetchStatus === "fetching", x = _ === "pending", w = _ === "error", T = x && y, E = d !== void 0, D = {
				status: _,
				fetchStatus: l.fetchStatus,
				isPending: x,
				isSuccess: _ === "success",
				isError: w,
				isInitialLoading: T,
				isLoading: T,
				data: d,
				dataUpdatedAt: l.dataUpdatedAt,
				error: f,
				errorUpdatedAt: m,
				failureCount: l.fetchFailureCount,
				failureReason: l.fetchFailureReason,
				errorUpdateCount: l.errorUpdateCount,
				isFetched: e.isFetched(),
				isFetchedAfterMount: l.dataUpdateCount > s.dataUpdateCount || l.errorUpdateCount > s.errorUpdateCount,
				isFetching: y,
				isRefetching: y && !x,
				isLoadingError: w && !E,
				isPaused: l.fetchStatus === "paused",
				isPlaceholderData: u,
				isRefetchError: w && E,
				isStale: C(e, t),
				refetch: this.refetch,
				promise: this.#o,
				isEnabled: (0, g.resolveQueryBoolean)(t.enabled, e) !== !1
			};
			if (this.options.experimental_prefetchInRender) {
				let t = D.data !== void 0, r = D.status === "error" && !t, i = (e) => {
					r ? e.reject(D.error) : t && e.resolve(D.data);
				}, a = () => {
					let e = this.#o = D.promise = (0, h.pendingThenable)();
					i(e);
				}, o = this.#o;
				switch (o.status) {
					case "pending":
						e.queryHash === n.queryHash && i(o);
						break;
					case "fulfilled":
						(r || D.data !== o.value) && a();
						break;
					case "rejected": r && D.error === o.reason || a();
				}
			}
			return D;
		}
		updateResult() {
			let e = this.#r, t = this.createResult(this.#t, this.options);
			this.#i = this.#t.state, this.#a = this.options, this.#i.data !== void 0 && (this.#u = this.#t), !(0, g.shallowEqualObjects)(t, e) && (this.#r = t, this.#C({ listeners: (() => {
				if (!e) return !0;
				let { notifyOnChangeProps: t } = this.options, n = typeof t == "function" ? t() : t;
				if (n === "all" || !n && !this.#m.size) return !0;
				let r = new Set(n ?? this.#m);
				return this.options.throwOnError && r.add("error"), Object.keys(this.#r).some((t) => {
					let n = t;
					return this.#r[n] !== e[n] && r.has(n);
				});
			})() }));
		}
		#S() {
			let e = this.#e.getQueryCache().build(this.#e, this.options);
			if (e === this.#t) return;
			let t = this.#t;
			this.#t = e, this.#n = e.state, this.hasListeners() && (t?.removeObserver(this), e.addObserver(this));
		}
		onQueryUpdate() {
			this.updateResult(), this.hasListeners() && this.#y();
		}
		#C(e) {
			f.notifyManager.batch(() => {
				e.listeners && this.listeners.forEach((e) => {
					e(this.#r);
				}), this.#e.getQueryCache().notify({
					query: this.#t,
					type: "observerResultsUpdated"
				});
			});
		}
	};
	function y(e, t) {
		return (0, g.resolveQueryBoolean)(t.enabled, e) !== !1 && e.state.data === void 0 && (e.state.status !== "error" || (0, g.resolveQueryBoolean)(t.retryOnMount, e) !== !1);
	}
	function b(e, t) {
		return y(e, t) || e.state.data !== void 0 && x(e, t, t.refetchOnMount);
	}
	function x(e, t, n) {
		if ((0, g.resolveQueryBoolean)(t.enabled, e) !== !1 && (0, g.resolveStaleTime)(t.staleTime, e) !== "static") {
			let r = typeof n == "function" ? n(e) : n;
			return r === "always" || r !== !1 && C(e, t);
		}
		return !1;
	}
	function S(e, t, n, r) {
		return (e !== t || (0, g.resolveQueryBoolean)(r.enabled, e) === !1) && (!n.suspense || e.state.status !== "error") && C(e, n);
	}
	function C(e, t) {
		return (0, g.resolveQueryBoolean)(t.enabled, e) !== !1 && e.isStaleByTime((0, g.resolveStaleTime)(t.staleTime, e));
	}
	function w(e, t) {
		return !(0, g.shallowEqualObjects)(e.getCurrentResult(), t);
	}
	0 && (t.exports = { QueryObserver: v });
})), Vn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { InfiniteQueryObserver: () => f }), t.exports = c(l);
	var u = Bn(), d = Rn(), f = class extends u.QueryObserver {
		constructor(e, t) {
			super(e, t);
		}
		bindMethods() {
			super.bindMethods(), this.fetchNextPage = this.fetchNextPage.bind(this), this.fetchPreviousPage = this.fetchPreviousPage.bind(this);
		}
		setOptions(e) {
			e._type = "infinite", super.setOptions(e);
		}
		getOptimisticResult(e) {
			return e._type = "infinite", super.getOptimisticResult(e);
		}
		fetchNextPage(e) {
			return this.fetch({
				...e,
				meta: { fetchMore: { direction: "forward" } }
			});
		}
		fetchPreviousPage(e) {
			return this.fetch({
				...e,
				meta: { fetchMore: { direction: "backward" } }
			});
		}
		createResult(e, t) {
			let { state: n } = e, r = super.createResult(e, t), { isFetching: i, isRefetching: a, isError: o, isRefetchError: s } = r, c = n.fetchMeta?.fetchMore?.direction, l = o && c === "forward", u = i && c === "forward", f = o && c === "backward", p = i && c === "backward";
			return {
				...r,
				fetchNextPage: this.fetchNextPage,
				fetchPreviousPage: this.fetchPreviousPage,
				hasNextPage: (0, d.hasNextPage)(t, n.data),
				hasPreviousPage: (0, d.hasPreviousPage)(t, n.data),
				isFetchNextPageError: l,
				isFetchingNextPage: u,
				isFetchPreviousPageError: f,
				isFetchingPreviousPage: p,
				isRefetchError: s && !l && !f,
				isRefetching: a && !u && !p
			};
		}
	};
	0 && (t.exports = { InfiniteQueryObserver: f });
})), Hn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		Mutation: () => p,
		getDefaultState: () => m
	}), t.exports = c(l);
	var u = Pn(), d = Ln(), f = In(), p = class extends d.Removable {
		#e;
		#t;
		#n;
		#r;
		constructor(e) {
			super(), this.#e = e.client, this.mutationId = e.mutationId, this.#n = e.mutationCache, this.#t = [], this.state = e.state || m(), this.setOptions(e.options), this.scheduleGc();
		}
		setOptions(e) {
			this.options = e, this.updateGcTime(this.options.gcTime);
		}
		get meta() {
			return this.options.meta;
		}
		addObserver(e) {
			this.#t.includes(e) || (this.#t.push(e), this.clearGcTimeout(), this.#n.notify({
				type: "observerAdded",
				mutation: this,
				observer: e
			}));
		}
		removeObserver(e) {
			this.#t = this.#t.filter((t) => t !== e), this.scheduleGc(), this.#n.notify({
				type: "observerRemoved",
				mutation: this,
				observer: e
			});
		}
		optionalRemove() {
			this.#t.length || (this.state.status === "pending" ? this.scheduleGc() : this.#n.remove(this));
		}
		continue() {
			return this.#r?.continue() ?? this.execute(this.state.variables);
		}
		async execute(e) {
			let t = () => {
				this.#i({ type: "continue" });
			}, n = {
				client: this.#e,
				meta: this.options.meta,
				mutationKey: this.options.mutationKey
			};
			this.#r = (0, f.createRetryer)({
				fn: () => this.options.mutationFn ? this.options.mutationFn(e, n) : Promise.reject(/* @__PURE__ */ Error("No mutationFn found")),
				onFail: (e, t) => {
					this.#i({
						type: "failed",
						failureCount: e,
						error: t
					});
				},
				onPause: () => {
					this.#i({ type: "pause" });
				},
				onContinue: t,
				retry: this.options.retry ?? 0,
				retryDelay: this.options.retryDelay,
				networkMode: this.options.networkMode,
				canRun: () => this.#n.canRun(this)
			});
			let r = this.state.status === "pending", i = !this.#r.canStart();
			try {
				if (r) t();
				else {
					this.#i({
						type: "pending",
						variables: e,
						isPaused: i
					}), this.#n.config.onMutate && await this.#n.config.onMutate(e, this, n);
					let t = await this.options.onMutate?.(e, n);
					t !== this.state.context && this.#i({
						type: "pending",
						context: t,
						variables: e,
						isPaused: i
					});
				}
				let a = await this.#r.start();
				return await this.#n.config.onSuccess?.(a, e, this.state.context, this, n), await this.options.onSuccess?.(a, e, this.state.context, n), await this.#n.config.onSettled?.(a, null, this.state.variables, this.state.context, this, n), await this.options.onSettled?.(a, null, e, this.state.context, n), this.#i({
					type: "success",
					data: a
				}), a;
			} catch (t) {
				try {
					await this.#n.config.onError?.(t, e, this.state.context, this, n);
				} catch (e) {
					Promise.reject(e);
				}
				try {
					await this.options.onError?.(t, e, this.state.context, n);
				} catch (e) {
					Promise.reject(e);
				}
				try {
					await this.#n.config.onSettled?.(void 0, t, this.state.variables, this.state.context, this, n);
				} catch (e) {
					Promise.reject(e);
				}
				try {
					await this.options.onSettled?.(void 0, t, e, this.state.context, n);
				} catch (e) {
					Promise.reject(e);
				}
				throw this.#i({
					type: "error",
					error: t
				}), t;
			} finally {
				this.#n.runNext(this);
			}
		}
		#i(e) {
			let t = (t) => {
				switch (e.type) {
					case "failed": return {
						...t,
						failureCount: e.failureCount,
						failureReason: e.error
					};
					case "pause": return {
						...t,
						isPaused: !0
					};
					case "continue": return {
						...t,
						isPaused: !1
					};
					case "pending": return {
						...t,
						context: e.context,
						data: void 0,
						failureCount: 0,
						failureReason: null,
						error: null,
						isPaused: e.isPaused,
						status: "pending",
						variables: e.variables,
						submittedAt: Date.now()
					};
					case "success": return {
						...t,
						data: e.data,
						failureCount: 0,
						failureReason: null,
						error: null,
						status: "success",
						isPaused: !1
					};
					case "error": return {
						...t,
						data: void 0,
						error: e.error,
						failureCount: t.failureCount + 1,
						failureReason: e.error,
						isPaused: !1,
						status: "error"
					};
				}
			};
			this.state = t(this.state), u.notifyManager.batch(() => {
				this.#t.forEach((t) => {
					t.onMutationUpdate(e);
				}), this.#n.notify({
					mutation: this,
					type: "updated",
					action: e
				});
			});
		}
	};
	function m() {
		return {
			context: void 0,
			data: void 0,
			error: null,
			failureCount: 0,
			failureReason: null,
			isPaused: !1,
			status: "idle",
			variables: void 0,
			submittedAt: 0
		};
	}
	0 && (t.exports = {
		Mutation: p,
		getDefaultState: m
	});
})), Un = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { MutationCache: () => m }), t.exports = c(l);
	var u = Pn(), d = Hn(), f = An(), p = Dn(), m = class extends p.Subscribable {
		constructor(e = {}) {
			super(), this.config = e, this.#e = /* @__PURE__ */ new Set(), this.#t = /* @__PURE__ */ new Map(), this.#n = 0;
		}
		#e;
		#t;
		#n;
		build(e, t, n) {
			let r = new d.Mutation({
				client: e,
				mutationCache: this,
				mutationId: ++this.#n,
				options: e.defaultMutationOptions(t),
				state: n
			});
			return this.add(r), r;
		}
		add(e) {
			this.#e.add(e);
			let t = h(e);
			if (typeof t == "string") {
				let n = this.#t.get(t);
				n ? n.push(e) : this.#t.set(t, [e]);
			}
			this.notify({
				type: "added",
				mutation: e
			});
		}
		remove(e) {
			if (this.#e.delete(e)) {
				let t = h(e);
				if (typeof t == "string") {
					let n = this.#t.get(t);
					if (n) {
						if (n.length > 1) {
							let t = n.indexOf(e);
							t !== -1 && n.splice(t, 1);
						} else n[0] === e && this.#t.delete(t);
					}
				}
			}
			this.notify({
				type: "removed",
				mutation: e
			});
		}
		canRun(e) {
			let t = h(e);
			if (typeof t == "string") {
				let n = this.#t.get(t)?.find((e) => e.state.status === "pending");
				return !n || n === e;
			}
			return !0;
		}
		runNext(e) {
			let t = h(e);
			return typeof t == "string" ? (this.#t.get(t)?.find((t) => t !== e && t.state.isPaused))?.continue() ?? Promise.resolve() : Promise.resolve();
		}
		clear() {
			u.notifyManager.batch(() => {
				this.#e.forEach((e) => {
					this.notify({
						type: "removed",
						mutation: e
					});
				}), this.#e.clear(), this.#t.clear();
			});
		}
		getAll() {
			return Array.from(this.#e);
		}
		find(e) {
			let t = {
				exact: !0,
				...e
			};
			return this.getAll().find((e) => (0, f.matchMutation)(t, e));
		}
		findAll(e = {}) {
			return this.getAll().filter((t) => (0, f.matchMutation)(e, t));
		}
		notify(e) {
			u.notifyManager.batch(() => {
				this.listeners.forEach((t) => {
					t(e);
				});
			});
		}
		resumePausedMutations() {
			let e = this.getAll().filter((e) => e.state.isPaused);
			return u.notifyManager.batch(() => Promise.all(e.map((e) => e.continue().catch(f.noop))));
		}
	};
	function h(e) {
		return e.options.scope?.id;
	}
	0 && (t.exports = { MutationCache: m });
})), Wn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { MutationObserver: () => m }), t.exports = c(l);
	var u = Hn(), d = Pn(), f = Dn(), p = An(), m = class extends f.Subscribable {
		#e;
		#t = void 0;
		#n;
		#r;
		constructor(e, t) {
			super(), this.#e = e, this.setOptions(t), this.bindMethods(), this.#i();
		}
		bindMethods() {
			this.mutate = this.mutate.bind(this), this.reset = this.reset.bind(this);
		}
		setOptions(e) {
			let t = this.options;
			this.options = this.#e.defaultMutationOptions(e), (0, p.shallowEqualObjects)(this.options, t) || this.#e.getMutationCache().notify({
				type: "observerOptionsUpdated",
				mutation: this.#n,
				observer: this
			}), t?.mutationKey && this.options.mutationKey && (0, p.hashKey)(t.mutationKey) !== (0, p.hashKey)(this.options.mutationKey) ? this.reset() : this.#n?.state.status === "pending" && this.#n.setOptions(this.options);
		}
		onUnsubscribe() {
			this.hasListeners() || this.#n?.removeObserver(this);
		}
		onMutationUpdate(e) {
			this.#i(), this.#a(e);
		}
		getCurrentResult() {
			return this.#t;
		}
		reset() {
			this.#n?.removeObserver(this), this.#n = void 0, this.#i(), this.#a();
		}
		mutate(e, t) {
			return this.#r = t, this.#n?.removeObserver(this), this.#n = this.#e.getMutationCache().build(this.#e, this.options), this.#n.addObserver(this), this.#n.execute(e);
		}
		#i() {
			let e = this.#n?.state ?? (0, u.getDefaultState)();
			this.#t = {
				...e,
				isPending: e.status === "pending",
				isSuccess: e.status === "success",
				isError: e.status === "error",
				isIdle: e.status === "idle",
				mutate: this.mutate,
				reset: this.reset
			};
		}
		#a(e) {
			d.notifyManager.batch(() => {
				if (this.#r && this.hasListeners()) {
					let t = this.#t.variables, n = this.#t.context, r = {
						client: this.#e,
						meta: this.options.meta,
						mutationKey: this.options.mutationKey
					};
					if (e?.type === "success") {
						try {
							this.#r.onSuccess?.(e.data, t, n, r);
						} catch (e) {
							Promise.reject(e);
						}
						try {
							this.#r.onSettled?.(e.data, null, t, n, r);
						} catch (e) {
							Promise.reject(e);
						}
					} else if (e?.type === "error") {
						try {
							this.#r.onError?.(e.error, t, n, r);
						} catch (e) {
							Promise.reject(e);
						}
						try {
							this.#r.onSettled?.(void 0, e.error, t, n, r);
						} catch (e) {
							Promise.reject(e);
						}
					}
				}
				this.listeners.forEach((e) => {
					e(this.#t);
				});
			});
		}
	};
	0 && (t.exports = { MutationObserver: m });
})), Gn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { QueriesObserver: () => g }), t.exports = c(l);
	var u = Pn(), d = Bn(), f = Dn(), p = An();
	function m(e, t) {
		let n = new Set(t);
		return e.filter((e) => !n.has(e));
	}
	function h(e, t, n) {
		let r = e.slice(0);
		return r[t] = n, r;
	}
	var g = class extends f.Subscribable {
		#e;
		#t;
		#n;
		#r;
		#i;
		#a;
		#o;
		#s;
		#c;
		#l = [];
		constructor(e, t, n) {
			super(), this.#e = e, this.#r = n, this.#n = [], this.#i = [], this.#t = [], this.setQueries(t);
		}
		onSubscribe() {
			this.listeners.size === 1 && this.#i.forEach((e) => {
				e.subscribe((t) => {
					this.#m(e, t);
				});
			});
		}
		onUnsubscribe() {
			this.listeners.size || this.destroy();
		}
		destroy() {
			this.listeners = /* @__PURE__ */ new Set(), this.#i.forEach((e) => {
				e.destroy();
			});
		}
		setQueries(e, t) {
			this.#n = e, this.#r = t, u.notifyManager.batch(() => {
				let e = this.#i, t = this.#p(this.#n);
				t.forEach((e) => e.observer.setOptions(e.defaultedQueryOptions));
				let n = t.map((e) => e.observer), r = n.map((e) => e.getCurrentResult()), i = e.length !== n.length, a = n.some((t, n) => t !== e[n]), o = i || a, s = o ? !0 : r.some((e, t) => {
					let n = this.#t[t];
					return !n || !(0, p.shallowEqualObjects)(e, n);
				});
				(o || s) && (o && (this.#l = t, this.#i = n), this.#t = r, this.hasListeners() && (o && (m(e, n).forEach((e) => {
					e.destroy();
				}), m(n, e).forEach((e) => {
					e.subscribe((t) => {
						this.#m(e, t);
					});
				})), this.#h()));
			});
		}
		getCurrentResult() {
			return this.#t;
		}
		getQueries() {
			return this.#i.map((e) => e.getCurrentQuery());
		}
		getObservers() {
			return this.#i;
		}
		getOptimisticResult(e, t) {
			let n = this.#p(e), r = n.map((e) => e.observer.getOptimisticResult(e.defaultedQueryOptions)), i = n.map((e) => e.defaultedQueryOptions.queryHash);
			return [
				r,
				(e) => this.#d(e ?? r, t, i),
				() => this.#u(r, n)
			];
		}
		#u(e, t) {
			return t.map((n, r) => {
				let i = e[r];
				return n.defaultedQueryOptions.notifyOnChangeProps ? i : n.observer.trackResult(i, (e) => {
					t.forEach((t) => {
						t.observer.trackProp(e);
					});
				});
			});
		}
		#d(e, t, n) {
			if (t) {
				let r = this.#c, i = n !== void 0 && r !== void 0 && (r.length !== n.length || n.some((e, t) => e !== r[t]));
				return (!this.#a || this.#t !== this.#s || i || t !== this.#o) && (this.#o = t, this.#s = this.#t, n !== void 0 && (this.#c = n), this.#a = (0, p.replaceEqualDeep)(this.#a, t(e))), this.#a;
			}
			return e;
		}
		#f() {
			return this.#r?.combine !== void 0 && this.#i.some((e, t) => e.options.suspense && this.#t[t]?.data === void 0);
		}
		#p(e) {
			let t = /* @__PURE__ */ new Map();
			this.#i.forEach((e) => {
				let n = e.options.queryHash;
				if (!n) return;
				let r = t.get(n);
				r ? r.push(e) : t.set(n, [e]);
			});
			let n = [];
			return e.forEach((e) => {
				let r = this.#e.defaultQueryOptions(e), i = t.get(r.queryHash)?.shift() ?? new d.QueryObserver(this.#e, r);
				n.push({
					defaultedQueryOptions: r,
					observer: i
				});
			}), n;
		}
		#m(e, t) {
			let n = this.#i.indexOf(e);
			n !== -1 && (this.#t = h(this.#t, n, t), this.#h());
		}
		#h() {
			if (this.hasListeners()) {
				let e = this.#u(this.#t, this.#l), t = this.#f(), n = this.#a, r = t ? n : this.#d(e, this.#r?.combine);
				(t || n !== r) && u.notifyManager.batch(() => {
					this.listeners.forEach((e) => {
						e(this.#t);
					});
				});
			}
		}
	};
	0 && (t.exports = { QueriesObserver: g });
})), Kn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { QueryCache: () => m }), t.exports = c(l);
	var u = An(), d = zn(), f = Pn(), p = Dn(), m = class extends p.Subscribable {
		constructor(e = {}) {
			super(), this.config = e, this.#e = /* @__PURE__ */ new Map();
		}
		#e;
		build(e, t, n) {
			let r = t.queryKey, i = t.queryHash ?? (0, u.hashQueryKeyByOptions)(r, t), a = this.get(i);
			return a || (a = new d.Query({
				client: e,
				queryKey: r,
				queryHash: i,
				options: e.defaultQueryOptions(t),
				state: n,
				defaultOptions: e.getQueryDefaults(r)
			}), this.add(a)), a;
		}
		add(e) {
			this.#e.has(e.queryHash) || (this.#e.set(e.queryHash, e), this.notify({
				type: "added",
				query: e
			}));
		}
		remove(e) {
			let t = this.#e.get(e.queryHash);
			t && (e.destroy(), t === e && this.#e.delete(e.queryHash), this.notify({
				type: "removed",
				query: e
			}));
		}
		clear() {
			f.notifyManager.batch(() => {
				this.getAll().forEach((e) => {
					this.remove(e);
				});
			});
		}
		get(e) {
			return this.#e.get(e);
		}
		getAll() {
			return [...this.#e.values()];
		}
		find(e) {
			let t = {
				exact: !0,
				...e
			};
			return this.getAll().find((e) => (0, u.matchQuery)(t, e));
		}
		findAll(e = {}) {
			let t = this.getAll();
			return Object.keys(e).length > 0 ? t.filter((t) => (0, u.matchQuery)(e, t)) : t;
		}
		notify(e) {
			f.notifyManager.batch(() => {
				this.listeners.forEach((t) => {
					t(e);
				});
			});
		}
		onFocus() {
			f.notifyManager.batch(() => {
				this.getAll().forEach((e) => {
					e.onFocus();
				});
			});
		}
		onOnline() {
			f.notifyManager.batch(() => {
				this.getAll().forEach((e) => {
					e.onOnline();
				});
			});
		}
	};
	0 && (t.exports = { QueryCache: m });
})), qn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { QueryClient: () => g }), t.exports = c(l);
	var u = An(), d = Kn(), f = Un(), p = On(), m = Fn(), h = Pn(), g = class {
		#e;
		#t;
		#n;
		#r;
		#i;
		#a;
		#o;
		#s;
		constructor(e = {}) {
			this.#e = e.queryCache || new d.QueryCache(), this.#t = e.mutationCache || new f.MutationCache(), this.#n = e.defaultOptions || {}, this.#r = /* @__PURE__ */ new Map(), this.#i = /* @__PURE__ */ new Map(), this.#a = 0;
		}
		mount() {
			this.#a++, this.#a === 1 && (this.#o = p.focusManager.subscribe(async (e) => {
				e && (await this.resumePausedMutations(), this.#e.onFocus());
			}), this.#s = m.onlineManager.subscribe(async (e) => {
				e && (await this.resumePausedMutations(), this.#e.onOnline());
			}));
		}
		unmount() {
			this.#a--, this.#a === 0 && (this.#o?.(), this.#o = void 0, this.#s?.(), this.#s = void 0);
		}
		isFetching(e) {
			return this.#e.findAll({
				...e,
				fetchStatus: "fetching"
			}).length;
		}
		isMutating(e) {
			return this.#t.findAll({
				...e,
				status: "pending"
			}).length;
		}
		getQueryData(e) {
			let t = this.defaultQueryOptions({ queryKey: e });
			return this.#e.get(t.queryHash)?.state.data;
		}
		ensureQueryData(e) {
			let t = this.defaultQueryOptions(e), n = this.#e.build(this, t), r = n.state.data;
			return r === void 0 ? this.fetchQuery(e) : (e.revalidateIfStale && n.isStaleByTime((0, u.resolveStaleTime)(t.staleTime, n)) && this.prefetchQuery(t), Promise.resolve(r));
		}
		getQueriesData(e) {
			return this.#e.findAll(e).map(({ queryKey: e, state: t }) => [e, t.data]);
		}
		setQueryData(e, t, n) {
			let r = this.defaultQueryOptions({ queryKey: e }), i = this.#e.get(r.queryHash)?.state.data, a = (0, u.functionalUpdate)(t, i);
			if (a !== void 0) return this.#e.build(this, r).setData(a, {
				...n,
				manual: !0
			});
		}
		setQueriesData(e, t, n) {
			return h.notifyManager.batch(() => this.#e.findAll(e).map(({ queryKey: e }) => [e, this.setQueryData(e, t, n)]));
		}
		getQueryState(e) {
			let t = this.defaultQueryOptions({ queryKey: e });
			return this.#e.get(t.queryHash)?.state;
		}
		removeQueries(e) {
			let t = this.#e;
			h.notifyManager.batch(() => {
				t.findAll(e).forEach((e) => {
					t.remove(e);
				});
			});
		}
		resetQueries(e, t) {
			let n = this.#e;
			return h.notifyManager.batch(() => (n.findAll(e).forEach((e) => {
				e.reset();
			}), this.refetchQueries({
				type: "active",
				...e
			}, t)));
		}
		cancelQueries(e, t = {}) {
			let n = {
				revert: !0,
				...t
			}, r = h.notifyManager.batch(() => this.#e.findAll(e).map((e) => e.cancel(n)));
			return Promise.all(r).then(u.noop).catch(u.noop);
		}
		invalidateQueries(e, t = {}) {
			return h.notifyManager.batch(() => (this.#e.findAll(e).forEach((e) => {
				e.invalidate();
			}), e?.refetchType === "none" ? Promise.resolve() : this.refetchQueries({
				...e,
				type: e?.refetchType ?? e?.type ?? "active"
			}, t)));
		}
		refetchQueries(e, t = {}) {
			let n = {
				...t,
				cancelRefetch: t.cancelRefetch ?? !0
			}, r = h.notifyManager.batch(() => this.#e.findAll(e).filter((e) => !e.isDisabled() && !e.isStatic()).map((e) => {
				let t = e.fetch(void 0, n);
				return n.throwOnError || (t = t.catch(u.noop)), e.state.fetchStatus === "paused" ? Promise.resolve() : t;
			}));
			return Promise.all(r).then(u.noop);
		}
		fetchQuery(e) {
			let t = this.defaultQueryOptions(e);
			t.retry === void 0 && (t.retry = !1);
			let n = this.#e.build(this, t);
			return n.isStaleByTime((0, u.resolveStaleTime)(t.staleTime, n)) ? n.fetch(t) : Promise.resolve(n.state.data);
		}
		prefetchQuery(e) {
			return this.fetchQuery(e).then(u.noop).catch(u.noop);
		}
		fetchInfiniteQuery(e) {
			return e._type = "infinite", this.fetchQuery(e);
		}
		prefetchInfiniteQuery(e) {
			return this.fetchInfiniteQuery(e).then(u.noop).catch(u.noop);
		}
		ensureInfiniteQueryData(e) {
			return e._type = "infinite", this.ensureQueryData(e);
		}
		resumePausedMutations() {
			return m.onlineManager.isOnline() ? this.#t.resumePausedMutations() : Promise.resolve();
		}
		getQueryCache() {
			return this.#e;
		}
		getMutationCache() {
			return this.#t;
		}
		getDefaultOptions() {
			return this.#n;
		}
		setDefaultOptions(e) {
			this.#n = e;
		}
		setQueryDefaults(e, t) {
			this.#r.set((0, u.hashKey)(e), {
				queryKey: e,
				defaultOptions: t
			});
		}
		getQueryDefaults(e) {
			let t = [...this.#r.values()], n = {};
			return t.forEach((t) => {
				(0, u.partialMatchKey)(e, t.queryKey) && Object.assign(n, t.defaultOptions);
			}), n;
		}
		setMutationDefaults(e, t) {
			this.#i.set((0, u.hashKey)(e), {
				mutationKey: e,
				defaultOptions: t
			});
		}
		getMutationDefaults(e) {
			let t = [...this.#i.values()], n = {};
			return t.forEach((t) => {
				(0, u.partialMatchKey)(e, t.mutationKey) && Object.assign(n, t.defaultOptions);
			}), n;
		}
		defaultQueryOptions(e) {
			if (e._defaulted) return e;
			let t = {
				...this.#n.queries,
				...this.getQueryDefaults(e.queryKey),
				...e,
				_defaulted: !0
			};
			return t.queryHash ||= (0, u.hashQueryKeyByOptions)(t.queryKey, t), t.refetchOnReconnect === void 0 && (t.refetchOnReconnect = t.networkMode !== "always"), t.throwOnError === void 0 && (t.throwOnError = !!t.suspense), !t.networkMode && t.persister && (t.networkMode = "offlineFirst"), t.queryFn === u.skipToken && (t.enabled = !1), t;
		}
		defaultMutationOptions(e) {
			return e?._defaulted ? e : {
				...this.#n.mutations,
				...e?.mutationKey && this.getMutationDefaults(e.mutationKey),
				...e,
				_defaulted: !0
			};
		}
		clear() {
			this.#e.clear(), this.#t.clear();
		}
	};
	0 && (t.exports = { QueryClient: g });
})), Jn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { streamedQuery: () => d }), t.exports = c(l);
	var u = An();
	function d({ streamFn: e, refetchMode: t = "reset", reducer: n = (e, t) => (0, u.addToEnd)(e, t), initialValue: r = [] }) {
		return async (i) => {
			let a = i.client.getQueryCache().find({
				queryKey: i.queryKey,
				exact: !0
			}), o = !!a && a.isFetched();
			o && t === "reset" && a.setState({
				...a.resetState,
				fetchStatus: "fetching"
			});
			let s = r, c = !1, l = await e((0, u.addConsumeAwareSignal)({
				client: i.client,
				meta: i.meta,
				queryKey: i.queryKey,
				pageParam: i.pageParam,
				direction: i.direction
			}, () => i.signal, () => c = !0)), d = o && t === "replace";
			for await (let e of l) {
				if (c) break;
				d ? s = n(s, e) : i.client.setQueryData(i.queryKey, (t) => n(t === void 0 ? r : t, e));
			}
			return d && !c && i.client.setQueryData(i.queryKey, s), i.client.getQueryData(i.queryKey) ?? r;
		};
	}
	0 && (t.exports = { streamedQuery: d });
})), Yn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		dataTagErrorSymbol: () => d,
		dataTagSymbol: () => u,
		unsetMarker: () => f
	}), t.exports = c(l);
	var u = /* @__PURE__ */ Symbol("dataTagSymbol"), d = /* @__PURE__ */ Symbol("dataTagErrorSymbol"), f = /* @__PURE__ */ Symbol("unsetMarker");
	0 && (t.exports = {
		dataTagErrorSymbol: d,
		dataTagSymbol: u,
		unsetMarker: f
	});
})), Xn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e, t, n) => (s(e, t, "default"), n && s(n, t, "default")), l = (e) => s(n({}, "__esModule", { value: !0 }), e), u = {};
	o(u, {
		CancelledError: () => C.CancelledError,
		InfiniteQueryObserver: () => m.InfiniteQueryObserver,
		Mutation: () => D.Mutation,
		MutationCache: () => h.MutationCache,
		MutationObserver: () => g.MutationObserver,
		QueriesObserver: () => y.QueriesObserver,
		Query: () => O.Query,
		QueryCache: () => b.QueryCache,
		QueryClient: () => x.QueryClient,
		QueryObserver: () => S.QueryObserver,
		defaultScheduler: () => _.defaultScheduler,
		defaultShouldDehydrateMutation: () => p.defaultShouldDehydrateMutation,
		defaultShouldDehydrateQuery: () => p.defaultShouldDehydrateQuery,
		dehydrate: () => p.dehydrate,
		environmentManager: () => f.environmentManager,
		experimental_streamedQuery: () => E.streamedQuery,
		focusManager: () => d.focusManager,
		hashKey: () => T.hashKey,
		hydrate: () => p.hydrate,
		isCancelledError: () => C.isCancelledError,
		isServer: () => T.isServer,
		keepPreviousData: () => T.keepPreviousData,
		matchMutation: () => T.matchMutation,
		matchQuery: () => T.matchQuery,
		noop: () => T.noop,
		notifyManager: () => _.notifyManager,
		onlineManager: () => v.onlineManager,
		partialMatchKey: () => T.partialMatchKey,
		replaceEqualDeep: () => T.replaceEqualDeep,
		shouldThrowError: () => T.shouldThrowError,
		skipToken: () => T.skipToken,
		timeoutManager: () => w.timeoutManager
	}), t.exports = l(u);
	var d = On(), f = jn(), p = Nn(), m = Vn(), h = Un(), g = Wn(), _ = Pn(), v = Fn(), y = Gn(), b = Kn(), x = qn(), S = Bn(), C = In(), w = kn(), T = An(), E = Jn(), D = Hn(), O = zn();
	c(u, Yn(), t.exports), 0 && (t.exports = {
		CancelledError,
		InfiniteQueryObserver,
		Mutation,
		MutationCache,
		MutationObserver,
		QueriesObserver,
		Query,
		QueryCache,
		QueryClient,
		QueryObserver,
		defaultScheduler,
		defaultShouldDehydrateMutation,
		defaultShouldDehydrateQuery,
		dehydrate,
		environmentManager,
		experimental_streamedQuery,
		focusManager,
		hashKey,
		hydrate,
		isCancelledError,
		isServer,
		keepPreviousData,
		matchMutation,
		matchQuery,
		noop,
		notifyManager,
		onlineManager,
		partialMatchKey,
		replaceEqualDeep,
		shouldThrowError,
		skipToken,
		timeoutManager,
		...Yn()
	});
})), Zn = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	};
	t.exports = ((e) => o(n({}, "__esModule", { value: !0 }), e))({});
})), Qn = /* @__PURE__ */ t(((e, t) => {
	var i = Object.create, a = Object.defineProperty, o = Object.getOwnPropertyDescriptor, s = Object.getOwnPropertyNames, c = Object.getPrototypeOf, l = Object.prototype.hasOwnProperty, u = (e, t) => {
		for (var n in t) a(e, n, {
			get: t[n],
			enumerable: !0
		});
	}, d = (e, t, n, r) => {
		if (t && typeof t == "object" || typeof t == "function") for (let i of s(t)) !l.call(e, i) && i !== n && a(e, i, {
			get: () => t[i],
			enumerable: !(r = o(t, i)) || r.enumerable
		});
		return e;
	}, f = (e, t, n) => (n = e == null ? {} : i(c(e)), d(t || !e || !e.__esModule ? a(n, "default", {
		value: e,
		enumerable: !0
	}) : n, e)), p = (e) => d(a({}, "__esModule", { value: !0 }), e), m = {};
	u(m, {
		QueryClientContext: () => _,
		QueryClientProvider: () => y,
		useQueryClient: () => v
	}), t.exports = p(m);
	var h = f(n(), 1), g = r(), _ = h.createContext(void 0), v = (e) => {
		let t = h.useContext(_);
		if (e) return e;
		if (!t) throw Error("No QueryClient set, use QueryClientProvider to set one");
		return t;
	}, y = ({ client: e, children: t }) => (h.useEffect(() => (e.mount(), () => {
		e.unmount();
	}), [e]), /* @__PURE__ */ (0, g.jsx)(_.Provider, {
		value: e,
		children: t
	}));
	0 && (t.exports = {
		QueryClientContext: _,
		QueryClientProvider: y,
		useQueryClient: v
	});
})), $n = /* @__PURE__ */ t(((e, t) => {
	var r = Object.create, i = Object.defineProperty, a = Object.getOwnPropertyDescriptor, o = Object.getOwnPropertyNames, s = Object.getPrototypeOf, c = Object.prototype.hasOwnProperty, l = (e, t) => {
		for (var n in t) i(e, n, {
			get: t[n],
			enumerable: !0
		});
	}, u = (e, t, n, r) => {
		if (t && typeof t == "object" || typeof t == "function") for (let s of o(t)) !c.call(e, s) && s !== n && i(e, s, {
			get: () => t[s],
			enumerable: !(r = a(t, s)) || r.enumerable
		});
		return e;
	}, d = (e, t, n) => (n = e == null ? {} : r(s(e)), u(t || !e || !e.__esModule ? i(n, "default", {
		value: e,
		enumerable: !0
	}) : n, e)), f = (e) => u(i({}, "__esModule", { value: !0 }), e), p = {};
	l(p, {
		IsRestoringProvider: () => _,
		useIsRestoring: () => g
	}), t.exports = f(p);
	var m = d(n(), 1), h = m.createContext(!1), g = () => m.useContext(h), _ = h.Provider;
	0 && (t.exports = {
		IsRestoringProvider: _,
		useIsRestoring: g
	});
})), er = /* @__PURE__ */ t(((e, t) => {
	var i = Object.create, a = Object.defineProperty, o = Object.getOwnPropertyDescriptor, s = Object.getOwnPropertyNames, c = Object.getPrototypeOf, l = Object.prototype.hasOwnProperty, u = (e, t) => {
		for (var n in t) a(e, n, {
			get: t[n],
			enumerable: !0
		});
	}, d = (e, t, n, r) => {
		if (t && typeof t == "object" || typeof t == "function") for (let i of s(t)) !l.call(e, i) && i !== n && a(e, i, {
			get: () => t[i],
			enumerable: !(r = o(t, i)) || r.enumerable
		});
		return e;
	}, f = (e, t, n) => (n = e == null ? {} : i(c(e)), d(t || !e || !e.__esModule ? a(n, "default", {
		value: e,
		enumerable: !0
	}) : n, e)), p = (e) => d(a({}, "__esModule", { value: !0 }), e), m = {};
	u(m, {
		QueryErrorResetBoundary: () => b,
		useQueryErrorResetBoundary: () => y
	}), t.exports = p(m);
	var h = f(n(), 1), g = r();
	function _() {
		let e = !1;
		return {
			clearReset: () => {
				e = !1;
			},
			reset: () => {
				e = !0;
			},
			isReset: () => e
		};
	}
	var v = h.createContext(_()), y = () => h.useContext(v), b = ({ children: e }) => {
		let [t] = h.useState(() => _());
		return /* @__PURE__ */ (0, g.jsx)(v.Provider, {
			value: t,
			children: typeof e == "function" ? e(t) : e
		});
	};
	0 && (t.exports = {
		QueryErrorResetBoundary: b,
		useQueryErrorResetBoundary: y
	});
})), tr = /* @__PURE__ */ t(((e, t) => {
	var r = Object.create, i = Object.defineProperty, a = Object.getOwnPropertyDescriptor, o = Object.getOwnPropertyNames, s = Object.getPrototypeOf, c = Object.prototype.hasOwnProperty, l = (e, t) => {
		for (var n in t) i(e, n, {
			get: t[n],
			enumerable: !0
		});
	}, u = (e, t, n, r) => {
		if (t && typeof t == "object" || typeof t == "function") for (let s of o(t)) !c.call(e, s) && s !== n && i(e, s, {
			get: () => t[s],
			enumerable: !(r = a(t, s)) || r.enumerable
		});
		return e;
	}, d = (e, t, n) => (n = e == null ? {} : r(s(e)), u(t || !e || !e.__esModule ? i(n, "default", {
		value: e,
		enumerable: !0
	}) : n, e)), f = (e) => u(i({}, "__esModule", { value: !0 }), e), p = {};
	l(p, {
		ensurePreventErrorBoundaryRetry: () => g,
		getHasError: () => v,
		useClearResetErrorBoundary: () => _
	}), t.exports = f(p);
	var m = d(n(), 1), h = Xn(), g = (e, t, n) => {
		let r = n?.state.error && typeof e.throwOnError == "function" ? (0, h.shouldThrowError)(e.throwOnError, [n.state.error, n]) : e.throwOnError;
		(e.suspense || e.experimental_prefetchInRender || r) && (t.isReset() || (e.retryOnMount = !1));
	}, _ = (e) => {
		m.useEffect(() => {
			e.clearReset();
		}, [e]);
	}, v = ({ result: e, errorResetBoundary: t, throwOnError: n, query: r, suspense: i }) => e.isError && !t.isReset() && !e.isFetching && r && (i && e.data === void 0 || (0, h.shouldThrowError)(n, [e.error, r]));
	0 && (t.exports = {
		ensurePreventErrorBoundaryRetry: g,
		getHasError: v,
		useClearResetErrorBoundary: _
	});
})), nr = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, {
		defaultThrowOnError: () => u,
		ensureSuspenseTimers: () => d,
		fetchOptimistic: () => m,
		shouldSuspend: () => p,
		willFetch: () => f
	}), t.exports = c(l);
	var u = (e, t) => t.state.data === void 0, d = (e) => {
		if (e.suspense) {
			let t = 1e3, n = (e) => e === "static" ? e : Math.max(e ?? t, t), r = e.staleTime;
			e.staleTime = typeof r == "function" ? (...e) => n(r(...e)) : n(r), typeof e.gcTime == "number" && (e.gcTime = Math.max(e.gcTime, t));
		}
	}, f = (e, t) => e.isLoading && e.isFetching && !t, p = (e, t) => e?.suspense && t.isPending, m = (e, t, n) => t.fetchOptimistic(e).catch(() => {
		n.clearReset();
	});
	0 && (t.exports = {
		defaultThrowOnError: u,
		ensureSuspenseTimers: d,
		fetchOptimistic: m,
		shouldSuspend: p,
		willFetch: f
	});
})), rr = /* @__PURE__ */ t(((e, t) => {
	var r = Object.create, i = Object.defineProperty, a = Object.getOwnPropertyDescriptor, o = Object.getOwnPropertyNames, s = Object.getPrototypeOf, c = Object.prototype.hasOwnProperty, l = (e, t) => {
		for (var n in t) i(e, n, {
			get: t[n],
			enumerable: !0
		});
	}, u = (e, t, n, r) => {
		if (t && typeof t == "object" || typeof t == "function") for (let s of o(t)) !c.call(e, s) && s !== n && i(e, s, {
			get: () => t[s],
			enumerable: !(r = a(t, s)) || r.enumerable
		});
		return e;
	}, d = (e, t, n) => (n = e == null ? {} : r(s(e)), u(t || !e || !e.__esModule ? i(n, "default", {
		value: e,
		enumerable: !0
	}) : n, e)), f = (e) => u(i({}, "__esModule", { value: !0 }), e), p = {};
	l(p, { useQueries: () => x }), t.exports = f(p);
	var m = d(n(), 1), h = Xn(), g = Qn(), _ = $n(), v = er(), y = tr(), b = nr();
	function x({ queries: e, ...t }, n) {
		let r = (0, g.useQueryClient)(n), i = (0, _.useIsRestoring)(), a = (0, v.useQueryErrorResetBoundary)(), o = m.useMemo(() => e.map((e) => {
			let t = r.defaultQueryOptions(e);
			return t._optimisticResults = i ? "isRestoring" : "optimistic", t;
		}), [
			e,
			r,
			i
		]);
		o.forEach((e) => {
			(0, b.ensureSuspenseTimers)(e);
			let t = r.getQueryCache().get(e.queryHash);
			(0, y.ensurePreventErrorBoundaryRetry)(e, a, t);
		}), (0, y.useClearResetErrorBoundary)(a);
		let [s] = m.useState(() => new h.QueriesObserver(r, o, t)), [c, l, u] = s.getOptimisticResult(o, t.combine), d = !i && t.subscribed !== !1;
		m.useSyncExternalStore(m.useCallback((e) => d ? s.subscribe(h.notifyManager.batchCalls(e)) : h.noop, [s, d]), () => s.getCurrentResult(), () => s.getCurrentResult()), m.useEffect(() => {
			s.setQueries(o, t);
		}, [
			o,
			t,
			s
		]);
		let f = c.some((e, t) => (0, b.shouldSuspend)(o[t], e)) ? c.flatMap((e, t) => {
			let n = o[t];
			if (n && (0, b.shouldSuspend)(n, e)) {
				let e = new h.QueryObserver(r, n);
				return (0, b.fetchOptimistic)(n, e, a);
			}
			return [];
		}) : [];
		if (f.length > 0) throw Promise.all(f);
		let p = c.find((e, t) => {
			let n = o[t];
			return n && (0, y.getHasError)({
				result: e,
				errorResetBoundary: a,
				throwOnError: n.throwOnError,
				query: r.getQueryCache().get(n.queryHash),
				suspense: n.suspense
			});
		});
		if (p?.error) throw p.error;
		return l(u());
	}
	0 && (t.exports = { useQueries: x });
})), ir = /* @__PURE__ */ t(((e, t) => {
	var r = Object.create, i = Object.defineProperty, a = Object.getOwnPropertyDescriptor, o = Object.getOwnPropertyNames, s = Object.getPrototypeOf, c = Object.prototype.hasOwnProperty, l = (e, t) => {
		for (var n in t) i(e, n, {
			get: t[n],
			enumerable: !0
		});
	}, u = (e, t, n, r) => {
		if (t && typeof t == "object" || typeof t == "function") for (let s of o(t)) !c.call(e, s) && s !== n && i(e, s, {
			get: () => t[s],
			enumerable: !(r = a(t, s)) || r.enumerable
		});
		return e;
	}, d = (e, t, n) => (n = e == null ? {} : r(s(e)), u(t || !e || !e.__esModule ? i(n, "default", {
		value: e,
		enumerable: !0
	}) : n, e)), f = (e) => u(i({}, "__esModule", { value: !0 }), e), p = {};
	l(p, { useBaseQuery: () => x }), t.exports = f(p);
	var m = d(n(), 1), h = Xn(), g = Qn(), _ = er(), v = tr(), y = $n(), b = nr();
	function x(e, t, n) {
		let r = (0, y.useIsRestoring)(), i = (0, _.useQueryErrorResetBoundary)(), a = (0, g.useQueryClient)(n), o = a.defaultQueryOptions(e);
		a.getDefaultOptions().queries?._experimental_beforeQuery?.(o);
		let s = a.getQueryCache().get(o.queryHash), c = e.subscribed !== !1;
		o._optimisticResults = r ? "isRestoring" : c ? "optimistic" : void 0, (0, b.ensureSuspenseTimers)(o), (0, v.ensurePreventErrorBoundaryRetry)(o, i, s), (0, v.useClearResetErrorBoundary)(i);
		let l = !a.getQueryCache().get(o.queryHash), [u] = m.useState(() => new t(a, o)), d = u.getOptimisticResult(o), f = !r && c;
		if (m.useSyncExternalStore(m.useCallback((e) => {
			let t = f ? u.subscribe(h.notifyManager.batchCalls(e)) : h.noop;
			return u.updateResult(), t;
		}, [u, f]), () => u.getCurrentResult(), () => u.getCurrentResult()), m.useEffect(() => {
			u.setOptions(o);
		}, [o, u]), (0, b.shouldSuspend)(o, d)) throw (0, b.fetchOptimistic)(o, u, i);
		if ((0, v.getHasError)({
			result: d,
			errorResetBoundary: i,
			throwOnError: o.throwOnError,
			query: s,
			suspense: o.suspense
		})) throw d.error;
		return a.getDefaultOptions().queries?._experimental_afterQuery?.(o, d), o.experimental_prefetchInRender && !h.environmentManager.isServer() && (0, b.willFetch)(d, r) && (l ? (0, b.fetchOptimistic)(o, u, i) : s?.promise)?.catch(h.noop).finally(() => {
			u.updateResult();
		}), o.notifyOnChangeProps ? d : u.trackResult(d);
	}
	0 && (t.exports = { useBaseQuery: x });
})), ar = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { useQuery: () => f }), t.exports = c(l);
	var u = Xn(), d = ir();
	function f(e, t) {
		return (0, d.useBaseQuery)(e, u.QueryObserver, t);
	}
	0 && (t.exports = { useQuery: f });
})), or = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { useSuspenseQuery: () => p }), t.exports = c(l);
	var u = Xn(), d = ir(), f = nr();
	function p(e, t) {
		return (0, d.useBaseQuery)({
			...e,
			enabled: !0,
			suspense: !0,
			throwOnError: f.defaultThrowOnError,
			placeholderData: void 0
		}, u.QueryObserver, t);
	}
	0 && (t.exports = { useSuspenseQuery: p });
})), sr = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { useSuspenseInfiniteQuery: () => p }), t.exports = c(l);
	var u = Xn(), d = ir(), f = nr();
	function p(e, t) {
		return (0, d.useBaseQuery)({
			...e,
			enabled: !0,
			suspense: !0,
			throwOnError: f.defaultThrowOnError
		}, u.InfiniteQueryObserver, t);
	}
	0 && (t.exports = { useSuspenseInfiniteQuery: p });
})), cr = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { useSuspenseQueries: () => f }), t.exports = c(l), Xn();
	var u = rr(), d = nr();
	function f(e, t) {
		return (0, u.useQueries)({
			...e,
			queries: e.queries.map((e) => ({
				...e,
				suspense: !0,
				throwOnError: d.defaultThrowOnError,
				enabled: !0,
				placeholderData: void 0
			}))
		}, t);
	}
	0 && (t.exports = { useSuspenseQueries: f });
})), lr = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { usePrefetchQuery: () => d }), t.exports = c(l);
	var u = Qn();
	function d(e, t) {
		let n = (0, u.useQueryClient)(t);
		n.getQueryState(e.queryKey) || n.prefetchQuery(e);
	}
	0 && (t.exports = { usePrefetchQuery: d });
})), ur = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { usePrefetchInfiniteQuery: () => d }), t.exports = c(l);
	var u = Qn();
	function d(e, t) {
		let n = (0, u.useQueryClient)(t);
		n.getQueryState(e.queryKey) || n.prefetchInfiniteQuery(e);
	}
	0 && (t.exports = { usePrefetchInfiniteQuery: d });
})), dr = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { queryOptions: () => u }), t.exports = c(l);
	function u(e) {
		return e;
	}
	0 && (t.exports = { queryOptions: u });
})), fr = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { infiniteQueryOptions: () => u }), t.exports = c(l);
	function u(e) {
		return e;
	}
	0 && (t.exports = { infiniteQueryOptions: u });
})), pr = /* @__PURE__ */ t(((e, t) => {
	var r = Object.create, i = Object.defineProperty, a = Object.getOwnPropertyDescriptor, o = Object.getOwnPropertyNames, s = Object.getPrototypeOf, c = Object.prototype.hasOwnProperty, l = (e, t) => {
		for (var n in t) i(e, n, {
			get: t[n],
			enumerable: !0
		});
	}, u = (e, t, n, r) => {
		if (t && typeof t == "object" || typeof t == "function") for (let s of o(t)) !c.call(e, s) && s !== n && i(e, s, {
			get: () => t[s],
			enumerable: !(r = a(t, s)) || r.enumerable
		});
		return e;
	}, d = (e, t, n) => (n = e == null ? {} : r(s(e)), u(t || !e || !e.__esModule ? i(n, "default", {
		value: e,
		enumerable: !0
	}) : n, e)), f = (e) => u(i({}, "__esModule", { value: !0 }), e), p = {};
	l(p, { HydrationBoundary: () => _ }), t.exports = f(p);
	var m = d(n(), 1), h = Xn(), g = Qn(), _ = ({ children: e, options: t = {}, state: n, queryClient: r }) => {
		let i = (0, g.useQueryClient)(r), a = m.useRef(t);
		m.useEffect(() => {
			a.current = t;
		});
		let o = m.useMemo(() => {
			if (n) {
				if (typeof n != "object") return;
				let e = i.getQueryCache(), t = n.queries || [], r = [], o = [];
				for (let n of t) {
					let t = e.get(n.queryHash);
					t ? (n.state.dataUpdatedAt > t.state.dataUpdatedAt || n.promise && t.state.status !== "pending" && t.state.fetchStatus !== "fetching" && n.dehydratedAt !== void 0 && n.dehydratedAt > t.state.dataUpdatedAt) && o.push(n) : r.push(n);
				}
				if (r.length > 0 && (0, h.hydrate)(i, { queries: r }, a.current), o.length > 0) return o;
			}
		}, [i, n]);
		return m.useEffect(() => {
			o && (0, h.hydrate)(i, { queries: o }, a.current);
		}, [i, o]), e;
	};
	0 && (t.exports = { HydrationBoundary: _ });
})), mr = /* @__PURE__ */ t(((e, t) => {
	var r = Object.create, i = Object.defineProperty, a = Object.getOwnPropertyDescriptor, o = Object.getOwnPropertyNames, s = Object.getPrototypeOf, c = Object.prototype.hasOwnProperty, l = (e, t) => {
		for (var n in t) i(e, n, {
			get: t[n],
			enumerable: !0
		});
	}, u = (e, t, n, r) => {
		if (t && typeof t == "object" || typeof t == "function") for (let s of o(t)) !c.call(e, s) && s !== n && i(e, s, {
			get: () => t[s],
			enumerable: !(r = a(t, s)) || r.enumerable
		});
		return e;
	}, d = (e, t, n) => (n = e == null ? {} : r(s(e)), u(t || !e || !e.__esModule ? i(n, "default", {
		value: e,
		enumerable: !0
	}) : n, e)), f = (e) => u(i({}, "__esModule", { value: !0 }), e), p = {};
	l(p, { useIsFetching: () => _ }), t.exports = f(p);
	var m = d(n(), 1), h = Xn(), g = Qn();
	function _(e, t) {
		let n = (0, g.useQueryClient)(t), r = n.getQueryCache();
		return m.useSyncExternalStore(m.useCallback((e) => r.subscribe(h.notifyManager.batchCalls(e)), [r]), () => n.isFetching(e), () => n.isFetching(e));
	}
	0 && (t.exports = { useIsFetching: _ });
})), hr = /* @__PURE__ */ t(((e, t) => {
	var r = Object.create, i = Object.defineProperty, a = Object.getOwnPropertyDescriptor, o = Object.getOwnPropertyNames, s = Object.getPrototypeOf, c = Object.prototype.hasOwnProperty, l = (e, t) => {
		for (var n in t) i(e, n, {
			get: t[n],
			enumerable: !0
		});
	}, u = (e, t, n, r) => {
		if (t && typeof t == "object" || typeof t == "function") for (let s of o(t)) !c.call(e, s) && s !== n && i(e, s, {
			get: () => t[s],
			enumerable: !(r = a(t, s)) || r.enumerable
		});
		return e;
	}, d = (e, t, n) => (n = e == null ? {} : r(s(e)), u(t || !e || !e.__esModule ? i(n, "default", {
		value: e,
		enumerable: !0
	}) : n, e)), f = (e) => u(i({}, "__esModule", { value: !0 }), e), p = {};
	l(p, {
		useIsMutating: () => _,
		useMutationState: () => y
	}), t.exports = f(p);
	var m = d(n(), 1), h = Xn(), g = Qn();
	function _(e, t) {
		let n = (0, g.useQueryClient)(t);
		return y({ filters: {
			...e,
			status: "pending"
		} }, n).length;
	}
	function v(e, t) {
		return e.findAll(t.filters).map((e) => t.select ? t.select(e) : e.state);
	}
	function y(e = {}, t) {
		let n = (0, g.useQueryClient)(t).getMutationCache(), r = m.useRef(e), i = m.useRef(null);
		return i.current === null && (i.current = v(n, e)), m.useEffect(() => {
			r.current = e;
		}), m.useSyncExternalStore(m.useCallback((e) => n.subscribe(() => {
			let t = (0, h.replaceEqualDeep)(i.current, v(n, r.current));
			i.current !== t && (i.current = t, h.notifyManager.schedule(e));
		}), [n]), () => i.current, () => i.current);
	}
	0 && (t.exports = {
		useIsMutating: _,
		useMutationState: y
	});
})), gr = /* @__PURE__ */ t(((e, t) => {
	var r = Object.create, i = Object.defineProperty, a = Object.getOwnPropertyDescriptor, o = Object.getOwnPropertyNames, s = Object.getPrototypeOf, c = Object.prototype.hasOwnProperty, l = (e, t) => {
		for (var n in t) i(e, n, {
			get: t[n],
			enumerable: !0
		});
	}, u = (e, t, n, r) => {
		if (t && typeof t == "object" || typeof t == "function") for (let s of o(t)) !c.call(e, s) && s !== n && i(e, s, {
			get: () => t[s],
			enumerable: !(r = a(t, s)) || r.enumerable
		});
		return e;
	}, d = (e, t, n) => (n = e == null ? {} : r(s(e)), u(t || !e || !e.__esModule ? i(n, "default", {
		value: e,
		enumerable: !0
	}) : n, e)), f = (e) => u(i({}, "__esModule", { value: !0 }), e), p = {};
	l(p, { useMutation: () => _ }), t.exports = f(p);
	var m = d(n(), 1), h = Xn(), g = Qn();
	function _(e, t) {
		let n = (0, g.useQueryClient)(t), [r] = m.useState(() => new h.MutationObserver(n, e));
		m.useEffect(() => {
			r.setOptions(e);
		}, [r, e]);
		let i = m.useSyncExternalStore(m.useCallback((e) => r.subscribe(h.notifyManager.batchCalls(e)), [r]), () => r.getCurrentResult(), () => r.getCurrentResult()), a = m.useCallback((e, t) => {
			r.mutate(e, t).catch(h.noop);
		}, [r]);
		if (i.error && (0, h.shouldThrowError)(r.options.throwOnError, [i.error])) throw i.error;
		return {
			...i,
			mutate: a,
			mutateAsync: i.mutate
		};
	}
	0 && (t.exports = { useMutation: _ });
})), _r = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { mutationOptions: () => u }), t.exports = c(l);
	function u(e) {
		return e;
	}
	0 && (t.exports = { mutationOptions: u });
})), vr = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e) => s(n({}, "__esModule", { value: !0 }), e), l = {};
	o(l, { useInfiniteQuery: () => f }), t.exports = c(l);
	var u = Xn(), d = ir();
	function f(e, t) {
		return (0, d.useBaseQuery)(e, u.InfiniteQueryObserver, t);
	}
	0 && (t.exports = { useInfiniteQuery: f });
})), yr = /* @__PURE__ */ t(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, s = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, c = (e, t, n) => (s(e, t, "default"), n && s(n, t, "default")), l = (e) => s(n({}, "__esModule", { value: !0 }), e), u = {};
	o(u, {
		HydrationBoundary: () => x.HydrationBoundary,
		IsRestoringProvider: () => O.IsRestoringProvider,
		QueryClientContext: () => b.QueryClientContext,
		QueryClientProvider: () => b.QueryClientProvider,
		QueryErrorResetBoundary: () => S.QueryErrorResetBoundary,
		infiniteQueryOptions: () => y.infiniteQueryOptions,
		mutationOptions: () => E.mutationOptions,
		queryOptions: () => v.queryOptions,
		useInfiniteQuery: () => D.useInfiniteQuery,
		useIsFetching: () => C.useIsFetching,
		useIsMutating: () => w.useIsMutating,
		useIsRestoring: () => O.useIsRestoring,
		useMutation: () => T.useMutation,
		useMutationState: () => w.useMutationState,
		usePrefetchInfiniteQuery: () => _.usePrefetchInfiniteQuery,
		usePrefetchQuery: () => g.usePrefetchQuery,
		useQueries: () => d.useQueries,
		useQuery: () => f.useQuery,
		useQueryClient: () => b.useQueryClient,
		useQueryErrorResetBoundary: () => S.useQueryErrorResetBoundary,
		useSuspenseInfiniteQuery: () => m.useSuspenseInfiniteQuery,
		useSuspenseQueries: () => h.useSuspenseQueries,
		useSuspenseQuery: () => p.useSuspenseQuery
	}), t.exports = l(u), c(u, Xn(), t.exports), c(u, Zn(), t.exports);
	var d = rr(), f = ar(), p = or(), m = sr(), h = cr(), g = lr(), _ = ur(), v = dr(), y = fr(), b = Qn(), x = pr(), S = er(), C = mr(), w = hr(), T = gr(), E = _r(), D = vr(), O = $n();
	0 && (t.exports = {
		HydrationBoundary,
		IsRestoringProvider,
		QueryClientContext,
		QueryClientProvider,
		QueryErrorResetBoundary,
		infiniteQueryOptions,
		mutationOptions,
		queryOptions,
		useInfiniteQuery,
		useIsFetching,
		useIsMutating,
		useIsRestoring,
		useMutation,
		useMutationState,
		usePrefetchInfiniteQuery,
		usePrefetchQuery,
		useQueries,
		useQuery,
		useQueryClient,
		useQueryErrorResetBoundary,
		useSuspenseInfiniteQuery,
		useSuspenseQueries,
		useSuspenseQuery,
		...Xn(),
		...Zn()
	});
}));
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/rpc/filterShapes.ts
function br(e, t) {
	let n = e["x-clicky-lookup"]?.$ref;
	if (!n || !t) return;
	let r = n.split("/").pop();
	return r ? t[r] : void 0;
}
function xr(e, t) {
	return new Map(e.map((e) => [e.name, br(e, t)]));
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/rpc/types.ts
function Sr(e) {
	return e["x-clicky-placeholder"] ?? e.placeholder;
}
function Cr(e) {
	return e.name === "args" || e.description?.toLowerCase().includes("positional argument") === !0;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/rpc/formMetadata.ts
function wr(e) {
	return e.truncated ? {
		truncated: !0,
		...e.total === void 0 ? {} : { total: e.total }
	} : {};
}
function Tr(e, t, n) {
	return e.truncated ? {
		...wr(e),
		...n ? { onSearch: (e) => n(t, e) } : {}
	} : {};
}
function Er(e) {
	return e.split(/[-_]/).filter(Boolean).map((e) => e.charAt(0).toUpperCase() + e.slice(1)).join(" ");
}
function Dr(e, t) {
	if (t.toUpperCase() === "GET" && e.in === "query" && !e.required) return "";
	let n = e.in === "path" ? "" : void 0, r = e.schema?.default ?? n;
	return typeof r == "boolean" ? r ? "true" : "false" : Array.isArray(r) || typeof r == "object" && r || r == null ? "" : String(r);
}
function Or(e, t, n = {}, r = {}) {
	return {
		...Object.fromEntries(e.map((e) => [e.name, Dr(e, t)])),
		...r,
		...n
	};
}
function kr(e) {
	return Object.fromEntries(Object.entries(e).filter(([, e]) => e !== ""));
}
function Ar(e, t) {
	let n = {}, r = [];
	for (let i of t) {
		let t = e[i.name];
		t && (i.in !== "path" && Cr(i) ? r.push(t) : n[i.name] = t);
	}
	return r.length > 0 && (n.args = r), n;
}
function jr(e, t) {
	return Ar(e, t.filter((e) => e["x-clicky"]?.role !== "limit" && e["x-clicky"]?.role !== "offset" && e["x-clicky"]?.role !== "cursor" && e["x-clicky"]?.role !== "sort" && e["x-clicky"]?.role !== "order"));
}
function Mr(e, t, n, r = {}) {
	let i = [], a = r.lookup?.filters ?? {}, o = xr(e, r.components), s = new Set(r.includeLocations ?? [
		"path",
		"query",
		"header"
	]), c = r.lockedValues ?? {}, l = r.hideLocked ?? !1, u = e.find((e) => e["x-clicky"]?.role === "limit"), d = e.find((e) => e["x-clicky"]?.role === "offset"), f = e.find((e) => e["x-clicky"]?.role === "cursor"), p = e.find((e) => e["x-clicky"]?.role === "sort"), m = e.find((e) => e["x-clicky"]?.role === "order");
	if (p == null != (m == null)) throw Error("Server sorting requires both sort and order parameters");
	let h = /* @__PURE__ */ new Set();
	u && h.add(u.name), d && h.add(d.name), f && h.add(f.name), p && h.add(p.name), m && h.add(m.name);
	let g = e.find((e) => e["x-clicky"]?.role === "search" && s.has(e.in)), _ = new Set(h);
	g && _.add(g.name);
	let v = (e) => ({
		...e,
		...d ? { [d.name]: "0" } : {},
		...f ? { [f.name]: "" } : {}
	}), y = e.find((e) => e["x-clicky"]?.role === "time-from"), b = e.find((e) => e["x-clicky"]?.role === "time-to"), x = (e) => o.get(e.name)?.type ?? a[e.name]?.type, S = y ?? e.find((e) => s.has(e.in) && e.in === "query" && x(e) === "from"), C = b ?? e.find((e) => s.has(e.in) && e.in === "query" && x(e) === "to"), w = S != null && C != null;
	for (let u of e) {
		if (!s.has(u.in) || _.has(u.name) || w && (u.name === S?.name || u.name === C?.name)) continue;
		let e = Object.prototype.hasOwnProperty.call(c, u.name);
		if (e && l) continue;
		let d = e ? c[u.name] ?? "" : t[u.name] ?? "", f = o.get(u.name), p = f?.label ?? a[u.name]?.label ?? Er(u.name), m = Sr(u), h = m === void 0 ? {} : { placeholder: m }, g = (t) => {
			if (e) return;
			let r = typeof t == "boolean" ? t ? "true" : "false" : t;
			n((e) => v({
				...e,
				[u.name]: r
			}));
		}, y = u.schema, b = a[u.name], x = f?.type ?? b?.type;
		if (x === "number" && u.in === "query") {
			let t = f?.unit ?? b?.unit;
			i.push({
				key: u.name,
				kind: "number",
				label: p,
				disabled: e,
				value: xn(d, f?.defaultOperator ?? b?.defaultOperator),
				...t ? { unit: t } : {},
				onChange: (t) => {
					e || n((e) => v({
						...e,
						[u.name]: Sn(t)
					}));
				}
			});
		} else if (x === "duration" && u.in === "query") {
			let t = Pr(f?.unit ?? b?.unit);
			i.push({
				key: u.name,
				kind: "duration",
				label: p,
				disabled: e,
				unit: t,
				value: Cn(d, t, f?.defaultOperator ?? b?.defaultOperator),
				onChange: (r) => {
					e || n((e) => v({
						...e,
						[u.name]: wn(r, t)
					}));
				}
			});
		} else if (x === "workload" && u.in === "query") i.push({
			key: u.name,
			kind: "workload",
			label: p,
			value: d,
			disabled: e,
			kinds: [
				"pod",
				"deployment",
				"statefulset",
				"daemonset"
			],
			loadWorkloads: async (e) => b ? Ir(b, e) : {},
			collapseSingleOption: !0,
			onChange: (e) => g(e)
		});
		else if (x === "labels" && u.in === "query") {
			let t = b ? Fr(b) : [], a = _n(d);
			t.some((e) => e.value.includes("=")) || Object.keys(a).some((e) => e.includes("=")) ? i.push({
				key: u.name,
				kind: "nested-multi",
				label: p,
				value: a,
				disabled: e,
				groups: Lr(t),
				onChange: (e) => n((t) => v({
					...t,
					[u.name]: yn(e)
				}))
			}) : i.push({
				key: u.name,
				kind: "multi",
				label: p,
				value: a,
				disabled: e,
				options: t,
				...b ? Tr(b, u.name, r.lookupSearch) : {},
				onChange: (e) => n((t) => v({
					...t,
					[u.name]: yn(e)
				}))
			});
		} else if ((x === "date-range" || x === "day-range") && u.in === "query") {
			let t = xn(d || String(u.schema?.default ?? ""));
			i.push({
				key: u.name,
				kind: "date-range",
				label: p,
				disabled: e,
				...t.min === void 0 ? {} : { from: t.min },
				...t.max === void 0 ? {} : { to: t.max },
				...b?.presets ? { presets: b.presets } : {},
				timeEnabled: x === "day-range" ? !1 : b?.timeEnabled ?? !0,
				...b?.timeZone ? { timeZone: b.timeZone } : {},
				...b?.timeZones ? { timeZones: b.timeZones } : {},
				onApply: (t, r) => {
					e || n((e) => v({
						...e,
						[u.name]: Sn({
							min: t,
							max: r
						})
					}));
				}
			});
		} else if (x === "value" && u.in === "query") i.push({
			key: u.name,
			kind: "text",
			label: p,
			value: d,
			disabled: e,
			...h,
			onChange: (e) => g(e)
		});
		else if (x === "multi-filter" && u.in === "query") i.push({
			key: u.name,
			kind: "multi",
			label: p,
			value: _n(d),
			disabled: e,
			options: b ? Fr(b) : [],
			...b ? Tr(b, u.name, r.lookupSearch) : {},
			onChange: (e) => n((t) => v({
				...t,
				[u.name]: yn(e)
			}))
		});
		else if (y?.enum) i.push({
			key: u.name,
			kind: "enum",
			label: p,
			value: d,
			disabled: e,
			options: y.enum.map((e) => ({
				value: String(e),
				label: String(e)
			})),
			onChange: (e) => g(e)
		});
		else if (x === "bool" || y?.type === "boolean") i.push({
			key: u.name,
			kind: "boolean",
			label: p,
			value: d === "true",
			disabled: e,
			onChange: (e) => g(e)
		});
		else if ((f != null || b != null) && u.in === "query") {
			if (f?.multi ?? b?.multi) {
				i.push({
					key: u.name,
					kind: "lookup-multi",
					label: p,
					value: bn(d),
					disabled: e,
					options: b ? Fr(b) : [],
					...b ? wr(b) : {},
					...h,
					onChange: (e) => n((t) => v({
						...t,
						[u.name]: e.join(",")
					}))
				});
				continue;
			}
			i.push({
				key: u.name,
				kind: "lookup",
				label: p,
				value: d,
				disabled: e,
				options: b ? Fr(b) : [],
				...h,
				inputType: x === "number" ? "number" : x === "date" ? "date" : "text",
				onChange: (e) => g(e)
			});
		} else i.push({
			key: u.name,
			kind: "text",
			label: p,
			value: d,
			disabled: e,
			...h,
			onChange: (e) => g(e)
		});
	}
	let T = { filters: i };
	if (p && m) {
		let e = t[p.name] ?? "", r = t[m.name] || "asc";
		if (r !== "asc" && r !== "desc") throw Error(`Unsupported server sort direction ${r}`);
		T.sort = {
			value: e ? {
				key: e,
				dir: r
			} : null,
			onChange: (e) => n((t) => {
				let n = { ...t };
				return e ? (n[p.name] = e.key, n[m.name] = e.dir) : (delete n[p.name], delete n[m.name]), v(n);
			})
		};
	}
	if (g) {
		let e = Object.prototype.hasOwnProperty.call(c, g.name), r = e ? c[g.name] ?? "" : t[g.name] ?? "", i = Sr(g);
		T.search = {
			value: r,
			onChange: (t) => {
				e || n((e) => v({
					...e,
					[g.name]: t
				}));
			},
			...i ? { placeholder: i } : {},
			ariaLabel: a[g.name]?.label ?? Er(g.name)
		};
	}
	if (u && (T.pagination = {
		limitParam: u.name,
		limitValue: t[u.name] ?? "",
		setLimit: (e) => n((t) => ({
			...t,
			[u.name]: e
		})),
		...d ? {
			offsetParam: d.name,
			offsetValue: t[d.name] ?? "",
			setOffset: (e) => n((t) => ({
				...t,
				[d.name]: e
			}))
		} : {},
		...f ? {
			cursorParam: f.name,
			cursorValue: t[f.name] ?? "",
			setCursor: (e) => n((t) => ({
				...t,
				[f.name]: e
			}))
		} : {}
	}), w && S != null && C != null) {
		let e = a[S.name], r = a[C.name], i = e ?? r, o = e?.timeEnabled ?? r?.timeEnabled;
		T.timeRange = {
			from: t[S.name] ?? "",
			to: t[C.name] ?? "",
			onApply: (e, t) => n((n) => v({
				...n,
				[S.name]: e,
				[C.name]: t
			})),
			...i?.presets ? { presets: i.presets } : {},
			timeEnabled: o ?? (S.schema?.format === "date-time" || C.schema?.format === "date-time"),
			...i?.timeZone ? { timeZone: i.timeZone } : {},
			...i?.timeZones ? { timeZones: i.timeZones } : {},
			...S.description ? { fromPlaceholder: S.description } : {},
			...C.description ? { toPlaceholder: C.description } : {}
		};
	}
	return T;
}
function Nr(e, t) {
	let [n, r] = (0, I.useState)(e);
	return (0, I.useEffect)(() => {
		let n = window.setTimeout(() => r(e), t);
		return () => window.clearTimeout(n);
	}, [t, e]), n;
}
function Pr(e) {
	let t = e || "ms";
	if (t !== "ms" && t !== "s" && t !== "m" && t !== "h") throw Error(`Unsupported duration storage unit ${t}`);
	return t;
}
function Fr(e) {
	let t = /* @__PURE__ */ new Map();
	for (let [n, r] of Object.entries(e.options ?? {})) t.set(n, {
		label: Rr(r) || n,
		title: Rr(r) || n
	});
	for (let [n, r] of Object.entries(e.selected ?? {})) t.has(n) || t.set(n, {
		label: Rr(r) || n,
		title: Rr(r) || n
	});
	let n = e.counts;
	return Array.from(t.entries()).map(([e, t]) => {
		let r = n?.[e];
		return {
			value: e,
			label: t.label ?? e,
			title: t.title ?? e,
			...r === void 0 ? {} : { count: r }
		};
	});
}
function Ir(e, t) {
	let n = new Set(t), r = {};
	for (let e of t) r[e] = [];
	for (let t of Fr(e)) {
		let e = t.value.split("/"), i = e.at(-1), a = e.at(-2)?.toLowerCase(), o = e.length > 2 ? e.slice(0, -2).join("/") : void 0;
		i && a && n.has(a) && r[a]?.push({
			name: i,
			...o ? { namespace: o } : {}
		});
	}
	return r;
}
function Lr(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) {
		let e = n.value.indexOf("=");
		if (e <= 0) continue;
		let r = n.value.slice(0, e), i = n.value.slice(e + 1), a = t.get(r) ?? [];
		a.push({
			...n,
			label: i || n.label
		}), t.set(r, a);
	}
	return Array.from(t, ([e, t]) => ({
		groupKey: e,
		label: e,
		options: t
	}));
}
function Rr(e) {
	return e == null ? "" : e.plain ? e.plain : e.text ? e.text : (e.children ?? []).map((e) => Rr(e)).join("");
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/rpc/useOperations.ts
var zr = yr();
function Br(e) {
	return (0, zr.useQuery)({
		queryKey: ["openapi-spec"],
		queryFn: () => e.getOpenAPISpec(),
		staleTime: 3e5
	});
}
function Vr(e) {
	let { data: t, ...n } = Br(e);
	return {
		operations: (0, I.useMemo)(() => {
			if (!t) return [];
			let e = [];
			for (let [n, r] of Object.entries(t.paths)) for (let [t, i] of Object.entries(r)) e.push({
				path: n,
				method: t,
				operation: i
			});
			return e;
		}, [t]),
		spec: t,
		...n
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/rpc/FilterFormFields.tsx
function Hr({ filters: e, timeRange: t, autoSubmit: n, isSubmitting: r, submitLabel: i, submittingLabel: a, className: o, showSubmit: s = !0 }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("space-y-3", o),
		children: [/* @__PURE__ */ (0, L.jsxs)("div", {
			className: "divide-y divide-border border-y border-border",
			children: [e.map((e) => /* @__PURE__ */ (0, L.jsx)(Ur, {
				filter: e,
				autoSubmit: n
			}, e.key)), t && /* @__PURE__ */ (0, L.jsx)(Kr, { timeRange: t })]
		}), !n && s && /* @__PURE__ */ (0, L.jsx)("div", {
			className: "flex justify-end",
			children: /* @__PURE__ */ (0, L.jsx)(be, {
				type: "submit",
				disabled: r,
				children: r ? a : i
			})
		})]
	});
}
function Ur({ filter: e, autoSubmit: t }) {
	let n = `clicky-param-${e.key.replace(/[^a-zA-Z0-9_-]/g, "-")}`;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		title: e.description,
		className: P("grid grid-cols-1 gap-2 py-2 sm:grid-cols-[minmax(8rem,14rem)_minmax(0,1fr)] sm:items-center", e.disabled && "opacity-60"),
		children: [/* @__PURE__ */ (0, L.jsx)("label", {
			htmlFor: n,
			className: "text-sm font-medium text-muted-foreground",
			children: e.label
		}), /* @__PURE__ */ (0, L.jsx)("div", {
			className: "min-w-0",
			children: Wr(e, n, t)
		})]
	});
}
function Wr(e, t, n) {
	if (e.kind === "date-range") return /* @__PURE__ */ (0, L.jsx)(c, {
		kind: e.timeEnabled === !1 ? "date" : "time",
		label: e.label,
		align: "left",
		from: e.from ?? "",
		to: e.to ?? "",
		onApply: e.onApply,
		...e.disabled === void 0 ? {} : { disabled: e.disabled },
		...e.presets ? { presets: e.presets } : {},
		...e.timeEnabled === void 0 ? {} : { timeEnabled: e.timeEnabled },
		...e.timeZone ? { timeZone: e.timeZone } : {},
		...e.timeZones ? { timeZones: e.timeZones } : {},
		...e.fromPlaceholder ? { fromPlaceholder: e.fromPlaceholder } : {},
		...e.toPlaceholder ? { toPlaceholder: e.toPlaceholder } : {}
	});
	if (e.kind === "enum") return /* @__PURE__ */ (0, L.jsxs)("select", {
		id: t,
		"aria-label": e.label,
		className: Jr,
		value: e.value,
		disabled: e.disabled,
		onChange: (t) => e.onChange(t.target.value),
		children: [/* @__PURE__ */ (0, L.jsx)("option", {
			value: "",
			children: e.placeholder ?? ""
		}), e.options.map((e) => /* @__PURE__ */ (0, L.jsx)("option", {
			value: e.value,
			children: e.label ?? e.value
		}, e.value))]
	});
	if (e.kind === "boolean") return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "flex h-8 items-center",
		children: /* @__PURE__ */ (0, L.jsx)("input", {
			id: t,
			type: "checkbox",
			"aria-label": e.label,
			className: "h-4 w-4 accent-primary disabled:cursor-not-allowed",
			checked: e.value,
			disabled: e.disabled,
			onChange: (t) => e.onChange(t.target.checked)
		})
	});
	if (e.kind === "lookup") {
		let n = `${t}-options`;
		return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsx)("input", {
			id: t,
			type: e.inputType === "number" ? "number" : "text",
			"aria-label": e.label,
			className: Jr,
			...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
			value: e.value,
			list: n,
			disabled: e.disabled,
			onChange: (t) => e.onChange(t.target.value)
		}), /* @__PURE__ */ (0, L.jsx)("datalist", {
			id: n,
			children: e.options.map((e) => /* @__PURE__ */ (0, L.jsx)("option", {
				value: e.value,
				label: e.label ?? e.value,
				disabled: e.disabled
			}, e.value))
		})] });
	}
	if (e.kind === "lookup-multi") {
		let n = `${t}-options`;
		return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsx)("input", {
			id: t,
			type: "text",
			"aria-label": e.label,
			className: Jr,
			...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
			value: e.value.join(", "),
			list: n,
			disabled: e.disabled,
			onChange: (t) => e.onChange(qr(t.target.value))
		}), /* @__PURE__ */ (0, L.jsx)("datalist", {
			id: n,
			children: e.options.map((e) => /* @__PURE__ */ (0, L.jsx)("option", {
				value: e.value,
				label: e.label ?? e.value,
				disabled: e.disabled
			}, e.value))
		})] });
	}
	return e.kind === "multi" ? /* @__PURE__ */ (0, L.jsx)(Gr, {
		filter: e,
		id: t
	}) : e.kind === "number" || e.kind === "duration" ? /* @__PURE__ */ (0, L.jsx)(h, {
		filter: e,
		chrome: "embedded",
		autoSubmit: n
	}) : e.kind === "text" ? /* @__PURE__ */ (0, L.jsx)("input", {
		id: t,
		type: "text",
		"aria-label": e.label,
		className: Jr,
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		value: e.value,
		disabled: e.disabled,
		onChange: (t) => e.onChange(t.target.value)
	}) : /* @__PURE__ */ (0, L.jsx)("input", {
		id: t,
		type: "text",
		"aria-label": e.label,
		className: Jr,
		value: JSON.stringify(e.value),
		disabled: !0,
		readOnly: !0
	});
}
function Gr({ filter: e, id: t }) {
	let n = e.truncated && e.total ? Math.max(e.total - e.options.length, 0) : 0;
	return /* @__PURE__ */ (0, L.jsx)(ft, {
		multiple: !0,
		id: t,
		ariaLabel: e.label,
		value: Object.keys(e.value),
		options: e.options.map((e) => ({
			value: e.value,
			label: typeof e.label == "string" ? e.label : e.title ?? e.value,
			...e.disabled === void 0 ? {} : { disabled: e.disabled },
			...e.title === void 0 ? {} : { title: e.title }
		})),
		onChange: (t) => {
			let n = {};
			for (let r of t) n[r] = e.value[r] ?? "include";
			e.onChange(n);
		},
		allowCustomValue: e.allowCustomValue ?? !1,
		...e.placeholder === void 0 ? {} : { placeholder: e.placeholder },
		...e.disabled === void 0 ? {} : { disabled: e.disabled },
		...e.className === void 0 ? {} : { className: e.className },
		...n > 0 ? { footer: `… and ${n.toLocaleString()} more` } : {}
	});
}
function Kr({ timeRange: e }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "grid grid-cols-1 gap-2 py-2 sm:grid-cols-[minmax(8rem,14rem)_minmax(0,1fr)] sm:items-center",
		children: [/* @__PURE__ */ (0, L.jsx)("div", {
			className: "text-sm font-medium text-muted-foreground",
			children: "Time range"
		}), /* @__PURE__ */ (0, L.jsx)("div", {
			className: "min-w-0",
			children: /* @__PURE__ */ (0, L.jsx)(c, {
				kind: e.timeEnabled ? "time" : "date",
				label: "Time range",
				align: "left",
				from: e.from ?? "",
				to: e.to ?? "",
				onApply: e.onApply,
				...e.presets ? { presets: e.presets } : {},
				...e.timeEnabled === void 0 ? {} : { timeEnabled: e.timeEnabled },
				...e.timeZone ? { timeZone: e.timeZone } : {},
				...e.timeZones ? { timeZones: e.timeZones } : {},
				...e.fromPlaceholder ? { fromPlaceholder: e.fromPlaceholder } : {},
				...e.toPlaceholder ? { toPlaceholder: e.toPlaceholder } : {}
			})
		})]
	});
}
function qr(e) {
	return e.split(",").map((e) => e.trim()).filter(Boolean);
}
var Jr = "h-8 w-full min-w-0 rounded-md border border-input bg-background px-2 text-sm text-foreground outline-none placeholder:text-placeholder disabled:cursor-not-allowed disabled:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring";
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/rpc/FilterForm.tsx
function Yr({ client: e, path: t, method: n, parameters: r = [], initialValues: i = {}, lockedValues: a = {}, hideLocked: o = !1, enableLookup: s = n.toUpperCase() === "GET", autoSubmit: c = !1, submitLabel: l = "Execute request", submittingLabel: u = "Executing…", emptyMessage: d = "This operation does not require input.", isSubmitting: f = !1, className: m, formId: h, showSubmit: g = !0, excludedParameterNames: _ = [], presentation: v = "form", onChange: y, onSubmit: b }) {
	let x = (0, I.useMemo)(() => `${n}:${t}:${JSON.stringify(i)}:${JSON.stringify(a)}`, [
		i,
		a,
		n,
		t
	]), S = (0, I.useMemo)(() => r.filter((e) => !_.includes(e.name)), [_, r]), [C, w] = (0, I.useState)(() => Or(r, n, a, i)), [T, E] = (0, I.useState)(""), D = Nr(C, 250), O = (0, I.useRef)(null), ee = (0, I.useRef)(y);
	(0, I.useEffect)(() => {
		ee.current = y;
	}), (0, I.useEffect)(() => {
		ee.current?.(C);
	}, [C]);
	let te = (0, I.useMemo)(() => jr(D, r), [D, r]), ne = (0, zr.useQuery)({
		queryKey: [
			"filter-form-lookup",
			n,
			t,
			te
		],
		queryFn: async () => await e.lookupFilters?.(t, n, te, { Accept: "application/json+clicky" }) ?? { filters: {} },
		enabled: s && !!e.lookupFilters && r.some((e) => e.in === "query"),
		placeholderData: zr.keepPreviousData,
		staleTime: 3e4,
		retry: 0
	}), { data: re } = Br(e), ie = re?.components?.["x-clicky-filters"], k = (0, I.useMemo)(() => Mr(S, C, w, {
		lookup: ne.data,
		components: ie,
		lockedValues: a,
		hideLocked: o
	}), [
		ie,
		o,
		a,
		ne.data,
		S,
		C
	]), A = k.filters.length > 0 || k.search != null || k.timeRange != null;
	async function j(e) {
		e?.preventDefault();
		let t = r.filter((e) => e.required ? (a[e.name] ?? C[e.name] ?? "").trim() === "" : !1);
		t.length > 0 ? E(`Missing required fields: ${t.map((e) => Er(e.name)).join(", ")}`) : (E(""), await b(kr(C)));
	}
	return (0, I.useEffect)(() => {
		w(Or(r, n, a, i)), E(""), O.current = null;
	}, [x]), (0, I.useEffect)(() => {
		if (!c || r.filter((e) => e.required ? (a[e.name] ?? D[e.name] ?? "").trim() === "" : !1).length > 0) return;
		let e = kr(D), t = JSON.stringify(e);
		O.current !== t && (O.current = t, E(""), b(e));
	}, [
		c,
		D,
		a,
		b,
		r
	]), /* @__PURE__ */ (0, L.jsxs)("form", {
		id: h,
		className: "space-y-3",
		onSubmit: j,
		children: [A ? v === "filter-bar" ? /* @__PURE__ */ (0, L.jsx)(p, {
			autoSubmit: c,
			filters: k.filters,
			overflowMode: "wrap",
			...k.search ? { search: k.search } : {},
			...k.timeRange ? { timeRange: k.timeRange } : {},
			...c ? {} : {
				onApply: () => void j(),
				applyLabel: l,
				isPending: f
			},
			...m ? { className: m } : {}
		}) : /* @__PURE__ */ (0, L.jsx)(Hr, {
			autoSubmit: c,
			filters: k.filters,
			isSubmitting: f,
			submitLabel: l,
			submittingLabel: u,
			showSubmit: g,
			...k.timeRange ? { timeRange: k.timeRange } : {},
			...m ? { className: m } : {}
		}) : /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex items-center justify-between gap-4",
			children: [/* @__PURE__ */ (0, L.jsx)("p", {
				className: "text-sm text-muted-foreground",
				children: d
			}), g ? /* @__PURE__ */ (0, L.jsx)(be, {
				type: "button",
				onClick: j,
				disabled: f,
				children: f ? u : l
			}) : null]
		}), T && /* @__PURE__ */ (0, L.jsx)("div", {
			className: "rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive",
			children: T
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/rpc/apiClient.ts
var Xr = class extends Error {
	status;
	method;
	url;
	responseBody;
	responseData;
	responseHeaders;
	constructor(e, t = {}) {
		super(e), this.name = "OperationsApiClientError", this.status = t.status, this.method = t.method, this.url = t.url, this.responseBody = t.responseBody, this.responseData = t.responseData, this.responseHeaders = t.responseHeaders;
	}
};
async function Zr(e, t) {
	let n = e.headers.get("Content-Type") || "";
	if (t === "blob") return {
		data: void 0,
		text: "[binary]",
		blob: await e.blob(),
		contentType: n
	};
	let r = await e.text();
	if (t === "text" || !ei(r, n)) return {
		data: r,
		text: r,
		contentType: n
	};
	try {
		return {
			data: JSON.parse(r),
			text: r,
			contentType: n
		};
	} catch {
		return {
			data: r,
			text: r,
			contentType: n
		};
	}
}
function Qr(e, t, n, r) {
	let i = $r(e.headers);
	return new Xr(i["x-error"] || ri(r.data) || `${t} ${n} failed with ${e.status}: ${r.text || e.statusText}`, {
		status: e.status,
		method: t,
		url: n,
		responseBody: r.text,
		responseData: r.data,
		responseHeaders: i
	});
}
function $r(e) {
	let t = {};
	return e.forEach((e, n) => {
		t[n.toLowerCase()] = e;
	}), t;
}
function ei(e, t) {
	let n = e.trim();
	return n !== "" && (t.toLowerCase().includes("json") || n.startsWith("{") || n.startsWith("["));
}
function ti(e) {
	if (!e || typeof e != "object") return;
	let t = e;
	if (typeof t.plain == "string" && t.plain !== "") return t.plain;
	if (typeof t.text == "string" && t.text !== "") return t.text;
	if (Array.isArray(t.children)) for (let e of t.children) {
		let t = ti(e);
		if (t) return t;
	}
}
function ni(e) {
	if (!e || typeof e != "object") return;
	let t = e.node;
	if (!t || typeof t != "object") return;
	let n = t;
	if (n.kind !== "map" || !Array.isArray(n.fields)) return;
	let r, i = !1;
	for (let e of n.fields) {
		if (!e || typeof e != "object") continue;
		let t = e;
		t.name === "Error" && (r = ti(t.value)), t.name === "Success" && ti(t.value) === "false" && (i = !0);
	}
	return i ? r : void 0;
}
function ri(e) {
	if (!e || typeof e != "object") return;
	let t = e;
	return typeof t.error == "string" && t.error ? t.error : typeof t.message == "string" && t.message ? t.message : ni(e);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/error-diagnostics.ts
function ii(e, t) {
	if (!e) return null;
	if (typeof e == "string") return e.trim() ? {
		message: e,
		context: []
	} : null;
	let n = mi(e);
	if (!n) return null;
	let r = mi(n.error);
	if (r && r !== n) return ii(r, t);
	let i = hi(n, [
		"error",
		"message",
		"msg",
		"reason",
		"detail",
		"details"
	]) ?? t, a = hi(n, [
		"trace",
		"trace_id",
		"traceId",
		"traceID"
	]), o = hi(n, [
		"stacktrace",
		"stack_trace",
		"stackTrace",
		"stack"
	]), s = hi(n, [
		"time",
		"timestamp",
		"created_at"
	]), c = hi(n, ["hint"]), l = [...c ? [["Hint", c]] : [], ...gi(n.context)], u = ai(n.details);
	if (!i && !a && !o && !s && l.length === 0) {
		let e = mi(n.diagnostics);
		if (e && e !== n) return ii(e, t);
	}
	return !i && !a && !o && !s && l.length === 0 ? null : {
		message: i ?? "Action failed",
		context: l,
		raw: e,
		...a === void 0 ? {} : { trace: a },
		...s === void 0 ? {} : { time: s },
		...o === void 0 ? {} : { stacktrace: o },
		...u.length > 0 ? { details: u } : {}
	};
}
function ai(e) {
	return Array.isArray(e) ? e.flatMap((e) => {
		let t = mi(e), n = hi(t ?? {}, ["label"]), r = hi(t ?? {}, ["value"]);
		return n && r ? [{
			label: n,
			value: r
		}] : [];
	}) : [];
}
function oi() {
	if (typeof window < "u") return `Page: ${window.location.origin}${window.location.pathname}`;
}
function si(e) {
	let t = e.split(/\r?\n/).map((e) => e.trim()).filter(Boolean), n = [], r = [], i;
	for (let e of t) {
		let t = di(e);
		t ? n.push(t) : !i && !e.startsWith("--- at ") ? i = e : r.push(e);
	}
	return {
		frames: n,
		unparsed: r,
		raw: e,
		...i === void 0 ? {} : { headline: i }
	};
}
function ci(e) {
	let t = e.trim();
	if (!t || !/^[{[]/.test(t)) return null;
	try {
		return JSON.parse(t);
	} catch {
		return null;
	}
}
function li(e) {
	return e.replace(/^github\.com\/flanksource\/incident-commander\//, "").replace(/^github\.com\/flanksource\//, "flanksource/").replace(/^.*\/go\/pkg\/mod\//, "pkg/mod/").replace(/^.*\/incident-commander\//, "");
}
function ui(e) {
	return e.includes("github.com/flanksource/incident-commander/") || e.includes("/incident-commander/");
}
function di(e) {
	if (!e.startsWith("--- at ")) return null;
	let t = e.slice(7), n = -1, r = -1, i = -1, a = -1;
	for (let e = 0; e < t.length; e++) {
		if (t[e] !== ":" || !fi(t[e + 1])) continue;
		let o = e + 1;
		for (; fi(t[o]);) o++;
		if (o === t.length) {
			n = e, r = e + 1, i = o, a = -1;
			continue;
		}
		if (!pi(t[o])) continue;
		let s = o;
		for (; pi(t[s]);) s++;
		s === t.length && s - o < 2 || (n = e, r = e + 1, i = o, a = s);
	}
	if (n <= 0 || r < 0 || i < 0) return null;
	let o = a >= 0 ? t.slice(a).trim() : "";
	return {
		raw: e,
		file: t.slice(0, n),
		line: Number(t.slice(r, i)),
		...o ? { functionName: o } : {}
	};
}
function fi(e) {
	return e !== void 0 && e >= "0" && e <= "9";
}
function pi(e) {
	return e !== void 0 && e.trim() === "";
}
function mi(e) {
	return !e || typeof e != "object" || Array.isArray(e) ? null : e;
}
function hi(e, t) {
	for (let n of t) {
		let t = e[n];
		if (typeof t == "string" && t.trim()) return t;
	}
}
function gi(e) {
	let t = mi(e);
	return t ? Object.entries(t).filter(([, e]) => e != null && e !== "").map(([e, t]) => [e, _i(t)]) : [];
}
function _i(e) {
	return e == null || e === "" ? "-" : typeof e == "string" ? e : typeof e == "number" || typeof e == "boolean" ? String(e) : JSON.stringify(e);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/ErrorDetails.tsx
function vi({ diagnostics: e, title: t = "Error", renderJsonContext: n, actions: r, className: i }) {
	let [a, o] = (0, I.useState)("idle"), s = e.context.filter(([, e]) => !ci(e)), c = e.context.map(([e, t]) => ({
		label: e,
		value: t,
		data: ci(t)
	})).filter((e) => e.data !== null), l = async (n) => {
		if (n.preventDefault(), n.stopPropagation(), !navigator.clipboard?.writeText) o("failed");
		else try {
			await navigator.clipboard.writeText(yi(e, t)), o("copied");
		} catch {
			o("failed");
		}
	};
	return /* @__PURE__ */ (0, L.jsxs)("details", {
		className: P("group rounded-md border border-destructive/30 bg-destructive/5", i),
		children: [/* @__PURE__ */ (0, L.jsxs)("summary", {
			className: "flex cursor-pointer list-none items-start gap-2 p-3",
			children: [
				/* @__PURE__ */ (0, L.jsx)(F, {
					icon: ye,
					className: "mt-0.5 shrink-0 text-destructive"
				}),
				/* @__PURE__ */ (0, L.jsxs)("div", {
					className: "min-w-0 flex-1",
					children: [/* @__PURE__ */ (0, L.jsx)("div", {
						className: "text-sm font-semibold text-destructive",
						children: t
					}), /* @__PURE__ */ (0, L.jsx)("div", {
						className: "mt-1 whitespace-pre-wrap text-sm text-destructive",
						children: e.message
					})]
				}),
				/* @__PURE__ */ (0, L.jsxs)("div", {
					className: "flex shrink-0 flex-col items-end gap-2 text-xs font-medium text-muted-foreground sm:flex-row sm:items-center sm:gap-3",
					children: [
						r && /* @__PURE__ */ (0, L.jsx)("span", {
							className: "inline-flex",
							onClick: (e) => e.preventDefault(),
							children: r
						}),
						/* @__PURE__ */ (0, L.jsxs)("button", {
							type: "button",
							"aria-label": a === "copied" ? "Copied" : a === "failed" ? "Copy failed" : "Copy error details",
							title: "Copy message, trace, time, context, and stack trace",
							className: "inline-flex items-center gap-1 hover:text-foreground",
							onClick: (e) => void l(e),
							children: [/* @__PURE__ */ (0, L.jsx)(F, { icon: a === "copied" ? Ue : Je }), /* @__PURE__ */ (0, L.jsx)("span", {
								"aria-live": "polite",
								children: a === "copied" ? "Copied" : a === "failed" ? "Copy failed" : "Copy"
							})]
						}),
						/* @__PURE__ */ (0, L.jsxs)("span", {
							className: "inline-flex items-center gap-1 hover:text-foreground",
							children: [
								/* @__PURE__ */ (0, L.jsx)("span", {
									className: "group-open:hidden",
									children: "More details"
								}),
								/* @__PURE__ */ (0, L.jsx)("span", {
									className: "hidden group-open:inline",
									children: "Less details"
								}),
								/* @__PURE__ */ (0, L.jsx)(F, {
									icon: Ie,
									className: "shrink-0 transition-transform group-open:rotate-90"
								})
							]
						})
					]
				})
			]
		}), /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "grid gap-3 border-t border-destructive/20 p-3 pt-2",
			children: [
				(e.trace || e.time) && /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "flex min-w-0 flex-wrap gap-2",
					children: [e.trace && /* @__PURE__ */ (0, L.jsx)(Ci, {
						label: "Trace",
						value: e.trace,
						className: "max-w-full"
					}), e.time && /* @__PURE__ */ (0, L.jsxs)("span", {
						className: "inline-flex max-w-full items-center overflow-hidden rounded-md border border-border bg-background/80 text-xs",
						children: [/* @__PURE__ */ (0, L.jsx)("span", {
							className: "shrink-0 bg-muted px-2 py-1 font-medium text-muted-foreground",
							children: "Time"
						}), /* @__PURE__ */ (0, L.jsx)("span", {
							className: "min-w-0 truncate px-2 py-1 font-mono text-foreground",
							children: e.time
						})]
					})]
				}),
				e.context.length > 0 && /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "min-w-0",
					children: [
						/* @__PURE__ */ (0, L.jsx)("div", {
							className: "mb-1.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground",
							children: "Context"
						}),
						s.length > 0 && /* @__PURE__ */ (0, L.jsx)("div", {
							className: "flex flex-wrap gap-1.5",
							children: s.map(([e, t]) => /* @__PURE__ */ (0, L.jsx)(Ci, {
								label: e,
								value: t
							}, `${e}:${t}`))
						}),
						c.length > 0 && /* @__PURE__ */ (0, L.jsx)("div", {
							className: "mt-2 grid gap-2",
							children: c.map((e) => n ? /* @__PURE__ */ (0, L.jsx)("span", { children: n(e) }, `${e.label}:${e.value}`) : /* @__PURE__ */ (0, L.jsx)(Ci, {
								label: e.label,
								value: e.value
							}, `${e.label}:${e.value}`))
						})
					]
				}),
				e.details?.map((e) => /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "min-w-0",
					children: [/* @__PURE__ */ (0, L.jsxs)("div", {
						className: "mb-1.5 flex min-w-0 items-center justify-between gap-3",
						children: [/* @__PURE__ */ (0, L.jsx)("div", {
							className: "text-xs font-semibold uppercase tracking-wide text-muted-foreground",
							children: e.label
						}), /* @__PURE__ */ (0, L.jsxs)("button", {
							type: "button",
							onClick: () => wi(e.value),
							className: "inline-flex items-center gap-1 text-xs font-medium text-muted-foreground hover:text-foreground",
							children: [/* @__PURE__ */ (0, L.jsx)(F, { icon: Je }), "copy"]
						})]
					}), /* @__PURE__ */ (0, L.jsx)("pre", {
						className: "max-h-64 overflow-auto whitespace-pre-wrap break-all rounded-md bg-muted p-2 font-mono text-xs leading-5 text-foreground",
						children: e.value
					})]
				}, e.label)),
				e.stacktrace && /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "min-w-0",
					children: [/* @__PURE__ */ (0, L.jsxs)("div", {
						className: "mb-1.5 flex min-w-0 items-center justify-between gap-3",
						children: [/* @__PURE__ */ (0, L.jsx)("div", {
							className: "text-xs font-semibold uppercase tracking-wide text-muted-foreground",
							children: "Stack trace"
						}), /* @__PURE__ */ (0, L.jsxs)("button", {
							type: "button",
							onClick: () => wi(e.stacktrace ?? ""),
							className: "inline-flex items-center gap-1 text-xs font-medium text-muted-foreground hover:text-foreground",
							children: [/* @__PURE__ */ (0, L.jsx)(F, { icon: Je }), "copy"]
						})]
					}), /* @__PURE__ */ (0, L.jsx)(bi, { stacktrace: e.stacktrace })]
				})
			]
		})]
	});
}
function yi(e, t) {
	let n = [`${t}: ${e.message}`], r = oi();
	r && n.push(r), e.trace && n.push(`Trace: ${e.trace}`), e.time && n.push(`Time: ${e.time}`), e.context.length > 0 && n.push("", "Context:", ...e.context.map(([e, t]) => `${e}: ${t}`));
	for (let t of e.details ?? []) n.push("", `${t.label}:`, t.value);
	return e.stacktrace && n.push("", "Stack trace:", e.stacktrace), n.join("\n");
}
function bi({ stacktrace: e }) {
	let t = si(e);
	return t.frames.length === 0 ? /* @__PURE__ */ (0, L.jsx)("pre", {
		className: "max-h-72 overflow-auto whitespace-pre-wrap rounded-md bg-slate-950 p-3 font-mono text-xs leading-5 text-slate-200",
		children: e
	}) : /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "max-h-96 min-h-40 overflow-auto rounded-md border border-border bg-background p-2",
		children: [
			/* @__PURE__ */ (0, L.jsx)("div", {
				className: "mb-2 flex items-center justify-between gap-2 text-[11px] text-muted-foreground",
				children: /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "flex items-center gap-1.5",
					children: [/* @__PURE__ */ (0, L.jsx)("span", {
						className: "rounded-full bg-red-100 px-2 py-0.5 text-red-700",
						children: "error"
					}), /* @__PURE__ */ (0, L.jsxs)("span", { children: [t.frames.length, " frames"] })]
				})
			}),
			t.headline && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "mb-2 rounded-md bg-red-50 px-2 py-1.5 font-mono text-[11px] leading-4 text-red-700",
				children: t.headline
			}),
			/* @__PURE__ */ (0, L.jsx)("div", {
				className: "space-y-0.5",
				children: t.frames.map((e, t) => /* @__PURE__ */ (0, L.jsx)(xi, {
					frame: e,
					index: t
				}, `${e.file}:${e.line}:${t}`))
			}),
			t.unparsed.length > 0 && /* @__PURE__ */ (0, L.jsx)("pre", {
				className: "mt-2 whitespace-pre-wrap rounded bg-muted p-2 font-mono text-[11px] leading-4 text-muted-foreground",
				children: t.unparsed.join("\n")
			})
		]
	});
}
function xi({ frame: e, index: t }) {
	let n = ui(e.file);
	return /* @__PURE__ */ (0, L.jsx)("button", {
		type: "button",
		onClick: () => wi(e.raw),
		title: "Copy stack frame",
		className: ["block w-full rounded px-1.5 py-1 text-left hover:bg-accent/50", n ? "text-foreground" : "text-muted-foreground"].join(" "),
		children: /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex min-w-0 items-start gap-1.5",
			children: [/* @__PURE__ */ (0, L.jsx)(F, {
				icon: n ? g : fe,
				className: "mt-0.5 shrink-0 text-[11px]"
			}), /* @__PURE__ */ (0, L.jsx)("div", {
				className: "min-w-0",
				children: /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "break-all font-mono text-[11px] font-semibold leading-4",
					children: [
						/* @__PURE__ */ (0, L.jsxs)("span", {
							className: "mr-2 text-[10px] font-normal opacity-60",
							children: ["#", t + 1]
						}),
						e.functionName || "unknown function",
						/* @__PURE__ */ (0, L.jsxs)("span", {
							className: "ml-2 text-[10px] font-normal opacity-80",
							children: [
								li(e.file),
								":",
								e.line
							]
						})
					]
				})
			})]
		})
	});
}
var Si = {
	clipped: {
		frame: "items-center overflow-hidden",
		label: "",
		value: "truncate",
		icon: ""
	},
	wrapped: {
		frame: "items-stretch",
		label: "flex items-center",
		value: "whitespace-normal break-all",
		icon: "self-center"
	}
};
function Ci({ label: e, value: t, className: n = "", wrap: r = !1 }) {
	let i = Si[r ? "wrapped" : "clipped"];
	return /* @__PURE__ */ (0, L.jsxs)("button", {
		type: "button",
		onClick: () => wi(t),
		title: `Copy ${e}`,
		className: [
			"inline-flex max-w-full rounded-md border border-border bg-background/80 text-left text-xs hover:border-primary/40 hover:bg-background",
			i.frame,
			n
		].filter(Boolean).join(" "),
		children: [
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: ["shrink-0 bg-muted px-2 py-1 font-medium text-muted-foreground", i.label].filter(Boolean).join(" "),
				children: e
			}),
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: ["min-w-0 px-2 py-1 font-mono text-foreground", i.value].join(" "),
				children: t
			}),
			/* @__PURE__ */ (0, L.jsx)(F, {
				icon: Je,
				className: ["mr-1.5 h-3 w-3 shrink-0 text-muted-foreground", i.icon].filter(Boolean).join(" ")
			})
		]
	});
}
function wi(e) {
	typeof navigator < "u" && navigator.clipboard && navigator.clipboard.writeText(e);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/rpc/operationErrorDiagnostics.tsx
function Ti(e, t) {
	return /* @__PURE__ */ (0, L.jsx)(vi, {
		title: t,
		diagnostics: Ei(e)
	});
}
function Ei(e) {
	let t = Ni(e), n = e instanceof Error ? e.message : String(e ?? ""), r = t?.responseData, i = Pi(t, "responseBody"), a = ii(r ?? e, n) ?? {
		message: n || "Request failed",
		context: []
	}, o = Di(r), s = [
		...Oi(t),
		...ki(r),
		...Ai(o),
		...a.context
	], c = [
		...ji(o),
		...Mi(i, a.message),
		...a.details ?? []
	], l = a.stacktrace ?? (e instanceof Error ? e.stack : void 0);
	return {
		...a,
		message: a.message || n || "Request failed",
		context: s,
		raw: e,
		...c.length > 0 ? { details: c } : {},
		...l === void 0 ? {} : { stacktrace: l }
	};
}
function Di(e) {
	return Ni(Ni(e)?.diagnostics) || null;
}
function Oi(e) {
	if (!e) return [];
	let t = [], n = Pi(e, "method"), r = Pi(e, "url"), i = e.status;
	return n && t.push(["Method", n]), r && t.push(["URL", r]), typeof i == "number" && t.push(["Status", String(i)]), t;
}
function ki(e) {
	let t = Pi(Ni(e), "code");
	return t ? [["Code", t]] : [];
}
function Ai(e) {
	if (!e) return [];
	let t = [];
	e.provider && t.push(["Provider", e.provider]);
	let n = e.response?.durationMs;
	typeof n == "number" && t.push(["Duration", `${n}ms`]);
	let r = e.response?.returnedRows;
	return typeof r == "number" && t.push(["Rows", String(r)]), t;
}
function ji(e) {
	if (!e?.request) return [];
	let t = [], n = e.request.query;
	typeof n == "string" && n.trim() && t.push({
		label: "Query",
		value: n
	});
	let r = e.request.arguments;
	Array.isArray(r) && r.length > 0 && t.push({
		label: "Arguments",
		value: JSON.stringify(r, null, 2)
	});
	let i = Ni(e.request.options);
	return i && Object.keys(i).length > 0 && t.push({
		label: "Options",
		value: JSON.stringify(i, null, 2)
	}), t;
}
function Mi(e, t) {
	return !e || e === t ? [] : [{
		label: "Response",
		value: e
	}];
}
function Ni(e) {
	return !e || typeof e != "object" || Array.isArray(e) ? null : e;
}
function Pi(e, t) {
	let n = e?.[t];
	return typeof n == "string" && n.trim() ? n : void 0;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/rpc/classify.ts
function Fi(e) {
	if (!e) return null;
	let t = e.stdout || e.output || "";
	if (!t.trim()) return null;
	try {
		return JSON.parse(t);
	} catch {
		return null;
	}
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/object-graph-nodes.ts
function R(e) {
	if (e.raw) return !1;
	if (e.value == null) return !0;
	if (typeof e.value != "string") return !1;
	let t = e.value.trim();
	return t === "" || t === "null";
}
function Ii(e) {
	return e.flatMap((e) => {
		let t = e.children ? Ii(e.children) : void 0;
		return t && t.length > 0 ? [{
			...e,
			children: t
		}] : e.expandable ? [t ? {
			...e,
			children: t
		} : e] : e.children && e.children.length > 0 || R(e) ? [] : [e];
	});
}
function Li(e) {
	return e.reduce((e, t) => e + 1 + Li(t.children ?? []), 0);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/ObjectGraph.tsx
function Ri({ roots: e, className: t, empty: n, showControls: r, defaultOpenDepth: i = 2, renderLabel: a, renderValue: o, renderType: s, loadChildren: c, onNodeSelect: l, selectedId: u, showOptionsMenu: d = !1, defaultHideNulls: f = !1, menuActions: p, showFullscreenControl: m = !1, fullscreenTitle: h = "Object", fullscreenButtonLabel: g = "Open object browser full screen" }) {
	let [_, v] = (0, I.useState)(f), [y, b] = (0, I.useState)(!1), x = (0, I.useMemo)(() => _ ? Ii(e) : e, [_, e]), S = (0, I.useMemo)(() => _ ? Li(e) - Li(x) : 0, [
		_,
		e,
		x
	]), C = [{
		label: _ && S > 0 ? `Hide null values (${S})` : "Hide null values",
		onSelect: () => v((e) => !e),
		..._ ? { icon: Ue } : {},
		title: "Drop fields whose value is null or blank"
	}, ...p ?? []], w = ({ inFullscreen: f }) => {
		let p = d || m && !f ? /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [m && !f && /* @__PURE__ */ (0, L.jsx)(zi, {
			label: g,
			onClick: () => b(!0),
			children: /* @__PURE__ */ (0, L.jsx)(F, {
				icon: Be,
				className: "text-sm"
			})
		}), d && /* @__PURE__ */ (0, L.jsx)(He, {
			items: C,
			align: "right",
			menuLabel: "Display options",
			trigger: /* @__PURE__ */ (0, L.jsx)(zi, {
				label: "Display options",
				children: /* @__PURE__ */ (0, L.jsx)(F, {
					icon: Ne,
					className: "text-sm"
				})
			})
		})] }) : void 0;
		return /* @__PURE__ */ (0, L.jsx)($e, {
			roots: x,
			getKey: (e) => e.id,
			getChildren: (e) => e.children,
			defaultOpen: (e, t) => t < i,
			...t !== void 0 && !f ? { className: t } : {},
			...p === void 0 ? {} : { actions: p },
			empty: _ && e.length > 0 ? /* @__PURE__ */ (0, L.jsx)("div", {
				className: "px-3 py-4 text-sm text-muted-foreground",
				children: "All values are null."
			}) : n,
			...r === void 0 ? {} : { showControls: r },
			...c ? {
				loadChildren: c,
				hasMoreChildren: (e) => e.expandable === !0
			} : {},
			renderRow: ({ node: e, loading: t, error: n }) => /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex min-w-0 flex-1 flex-wrap items-center gap-2 font-mono text-xs",
				children: [
					a ? a(e) : /* @__PURE__ */ (0, L.jsx)(Bi, {
						node: e,
						selected: u != null && e.id === u,
						...o ? { renderValue: o } : {},
						...s ? { renderType: s } : {},
						...l ? { onSelect: l } : {}
					}),
					t && /* @__PURE__ */ (0, L.jsx)("span", {
						className: "shrink-0 text-muted-foreground",
						children: "…"
					}),
					n != null && /* @__PURE__ */ (0, L.jsx)("span", {
						className: "break-all text-red-600",
						children: n instanceof Error ? n.message : String(n)
					})
				]
			})
		});
	};
	return m ? /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [w({ inFullscreen: !1 }), /* @__PURE__ */ (0, L.jsx)(Te, {
		open: y,
		onClose: () => b(!1),
		title: h,
		size: "full",
		className: "h-[95vh]",
		scrollBody: !1,
		children: /* @__PURE__ */ (0, L.jsx)("div", {
			className: "flex h-full min-h-0 flex-col",
			children: w({ inFullscreen: !0 })
		})
	})] }) : w({ inFullscreen: !1 });
}
function zi({ label: e, onClick: t, children: n }) {
	return /* @__PURE__ */ (0, L.jsx)("button", {
		type: "button",
		"aria-label": e,
		title: e,
		className: "inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-accent-foreground",
		...t ? { onClick: t } : {},
		children: n
	});
}
function Bi({ node: e, onSelect: t, renderValue: n, renderType: r, selected: i = !1 }) {
	let a = t ? /* @__PURE__ */ (0, L.jsx)("button", {
		type: "button",
		onClick: (n) => {
			n.stopPropagation(), t(e);
		},
		title: "Select this path",
		className: P("rounded px-0.5 text-left text-muted-foreground hover:text-foreground hover:underline", i && "bg-primary/15 text-foreground"),
		children: e.label
	}) : /* @__PURE__ */ (0, L.jsx)("span", {
		className: "text-muted-foreground",
		children: e.label
	}), o = n?.(e);
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		a,
		e.type && (r ? r(e) : /* @__PURE__ */ (0, L.jsxs)("span", {
			className: "text-muted-foreground/60",
			children: ["@", e.type]
		})),
		o ?? (e.value == null ? e.raw && !e.children ? /* @__PURE__ */ (0, L.jsx)("span", {
			className: "break-all italic text-muted-foreground/70",
			children: e.raw
		}) : null : /* @__PURE__ */ (0, L.jsx)("span", {
			className: "break-all text-foreground",
			children: String(e.value)
		}))
	] });
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/ExecutionTree.tsx
function Vi({ roots: e, className: t, empty: n, showControls: r, defaultOpenDepth: i = 1, costThreshold: a, renderRow: o, renderDetail: s, renderAfterChildren: c, loadChildren: l }) {
	return /* @__PURE__ */ (0, L.jsx)($e, {
		roots: e,
		getKey: (e) => e.id,
		getChildren: (e) => e.children,
		defaultOpen: (e, t) => t < i,
		...t === void 0 ? {} : { className: t },
		...n === void 0 ? {} : { empty: n },
		...r === void 0 ? {} : { showControls: r },
		...s ? { renderDetail: s } : {},
		...c ? { renderAfterChildren: c } : {},
		...l ? {
			loadChildren: l,
			hasMoreChildren: (e) => e.expandable === !0
		} : {},
		renderRow: ({ node: e }) => o ? o(e) : /* @__PURE__ */ (0, L.jsx)(Hi, {
			node: e,
			costThreshold: a
		})
	});
}
function Hi({ node: e, costThreshold: t }) {
	let n = t != null && e.cost != null && e.cost >= t;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex min-w-0 flex-1 flex-wrap items-center gap-2 font-mono text-xs",
		children: [
			e.status && /* @__PURE__ */ (0, L.jsx)(Gi, { status: e.status }),
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: "font-semibold text-foreground",
				children: e.label
			}),
			e.cost != null && /* @__PURE__ */ (0, L.jsx)("span", {
				className: n ? "text-red-600" : "text-muted-foreground",
				children: Ui(e)
			}),
			e.times != null && e.times > 1 && /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "text-muted-foreground/60",
				children: ["×", e.times]
			}),
			e.className && /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "break-all text-muted-foreground/60",
				children: [e.className, e.lineNumber != null && e.lineNumber > 0 ? `:${e.lineNumber}` : ""]
			}),
			e.detail && Object.entries(e.detail).map(([e, t]) => /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "text-muted-foreground/60",
				children: [
					e,
					"=",
					t
				]
			}, e))
		]
	});
}
function Ui(e) {
	let t = e.unit || "ms", n = e.cost ?? 0;
	return t === "ms" ? `${n.toFixed(3)}ms` : `${n} ${t}`;
}
var Wi = {
	ok: "text-green-600",
	success: "text-green-600",
	error: "text-red-600",
	warning: "text-amber-500",
	running: "text-blue-600",
	pending: "text-muted-foreground"
};
function Gi({ status: e }) {
	return /* @__PURE__ */ (0, L.jsx)("span", {
		className: `shrink-0 ${Wi[e] ?? "text-muted-foreground"}`,
		title: e,
		children: "●"
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/stacktrace-parse.ts
var Ki = /^\s*at\s+([\w$.]+)\.([\w$<>]+)\(([^)]+)\)/, qi = /^\.\.\.\s+\d+\s+more$/;
function Ji(e) {
	let t = {
		causedBy: [],
		frames: [],
		language: "java"
	};
	if (!e || !e.trim()) return t;
	let n = [];
	for (let r of e.split("\n")) {
		let e = r.trim();
		if (!e || qi.test(e)) continue;
		let i = Ki.exec(r);
		if (i && i[1] && i[2] && i[3] !== void 0) t.frames.push(Xi(i[1], i[2], i[3]));
		else if (e.startsWith("Caused by:")) t.causedBy.push(e.slice(10).trim());
		else if (e.startsWith("Internal Exception:")) t.causedBy.push(e.slice(19).trim());
		else {
			if (!t.exceptionClass) {
				let n = Yi(e);
				if (n) {
					t.exceptionClass = n.exceptionClass, n.message && (t.message = n.message);
					continue;
				}
			}
			n.push(e);
		}
	}
	return !t.message && n.length > 0 && (t.message = n.join(" ")), t;
}
function Yi(e) {
	let t = e.indexOf(":"), n = t === -1 ? e : e.slice(0, t);
	if (!n || ![
		"Exception",
		"Error",
		"Throwable"
	].some((e) => n.endsWith(e))) return null;
	for (let e of n) {
		let t = e.charCodeAt(0);
		if (!(t >= 48 && t <= 57 || t >= 65 && t <= 90 || t >= 97 && t <= 122 || e === "_" || e === "." || e === "$")) return null;
	}
	if (t === -1) return { exceptionClass: n };
	let r = e.slice(t + 1).trimStart();
	return r ? {
		exceptionClass: n,
		message: r
	} : { exceptionClass: n };
}
function Xi(e, t, n) {
	let r = n.trim(), i, a, o = !1;
	if (r === "Native Method") o = !0;
	else if (r !== "Unknown Source") {
		let e = (r.split(" ~[")[0] ?? r).trim(), t = e.lastIndexOf(":");
		if (t >= 0) {
			i = e.slice(0, t);
			let n = Number(e.slice(t + 1));
			Number.isFinite(n) && (a = n);
		} else i = e;
	}
	let s = `${e}.${t}`, c = T(i, a) ?? (o ? "Native Method" : void 0), l = {
		functionName: s,
		displayName: ve(e, t),
		kind: "frame",
		runtime: ge(e),
		nativeMethod: o,
		class: e,
		method: t
	};
	return i !== void 0 && (l.file = i), a !== void 0 && (l.line = a), c !== void 0 && (l.location = c), l;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/RenderedStackTrace.tsx
function Zi({ input: e, language: t = "java", resolver: n, contextLines: r = 3, hideRuntimeOnly: i = !1, include: a, exclude: o, frameActions: s, renderFramePart: c, renderFrameDetail: l, frameActionsVisibility: u = "hover", showFrameIndex: d = !0, frameFilter: f, renderCauseDetail: p, className: m }) {
	let h = (0, I.useMemo)(() => typeof e == "string" ? Ji(e) : e, [e, t]), g = (0, I.useMemo)(() => n ? h.frames.map((e) => {
		if (e.sourceLines && e.sourceLines.length > 0) return e;
		let t = n(e, r);
		if (!t) return e;
		let { line: i, ...a } = t;
		return {
			...e,
			...a
		};
	}) : h.frames, [
		h.frames,
		n,
		r
	]), _ = (0, I.useMemo)(() => g.filter((e) => i && e.runtime || f && !f(e) || o && o.some((t) => e.class?.startsWith(t)) ? !1 : a && a.length > 0 ? a.some((t) => e.class?.startsWith(t)) : !0), [
		g,
		i,
		a,
		o,
		f
	]);
	if (_.length === 0 && !h.exceptionClass) return null;
	let v = _.length !== g.length;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: ["overflow-hidden rounded-md border border-border bg-background text-foreground", m].filter(Boolean).join(" "),
		children: [
			(h.exceptionClass || h.message) && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "border-b border-border bg-red-500/5 px-3 py-2",
				children: /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "flex min-w-0 items-start gap-2",
					children: [/* @__PURE__ */ (0, L.jsx)(F, {
						icon: wt,
						className: "mt-0.5 shrink-0 text-sm text-red-600 dark:text-red-400"
					}), /* @__PURE__ */ (0, L.jsxs)("div", {
						className: "min-w-0 flex-1",
						children: [h.exceptionClass && /* @__PURE__ */ (0, L.jsx)("div", {
							className: "break-all font-mono text-xs font-semibold text-red-700 dark:text-red-300",
							children: h.exceptionClass
						}), h.message && /* @__PURE__ */ (0, L.jsx)("div", {
							className: "mt-0.5 whitespace-pre-wrap break-words text-xs text-foreground",
							children: h.message
						})]
					})]
				})
			}),
			h.causedBy.length > 0 && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "space-y-1 border-b border-border bg-orange-500/5 px-3 py-2",
				children: h.causedBy.map((e, t) => {
					let n = p?.(e, t);
					return /* @__PURE__ */ (0, L.jsxs)("div", {
						"data-cause-index": t,
						className: "min-w-0",
						children: [/* @__PURE__ */ (0, L.jsxs)("div", {
							className: "flex min-w-0 items-start gap-2 font-mono text-[11px] text-orange-700 dark:text-orange-300",
							children: [
								/* @__PURE__ */ (0, L.jsx)(F, {
									icon: ee,
									className: "mt-0.5 shrink-0 text-xs"
								}),
								/* @__PURE__ */ (0, L.jsx)("span", {
									className: "shrink-0 opacity-75",
									children: "Caused by"
								}),
								/* @__PURE__ */ (0, L.jsx)("span", {
									className: "min-w-0 break-all",
									children: e
								})
							]
						}), n ? /* @__PURE__ */ (0, L.jsx)("div", {
							className: "mt-1 pl-5",
							children: n
						}) : null]
					}, t);
				})
			}),
			/* @__PURE__ */ (0, L.jsx)("div", {
				className: "divide-y divide-border/60",
				children: _.map((e, t) => /* @__PURE__ */ (0, L.jsx)(A, {
					frame: e,
					index: t,
					showIndex: d,
					...s ? { frameActions: s } : {},
					...c ? { renderFramePart: c } : {},
					...l ? { renderFrameDetail: l } : {},
					frameActionsVisibility: u
				}, `${e.functionName}-${t}`))
			}),
			v && /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "border-t border-border bg-muted/30 px-3 py-1.5 text-[11px] text-muted-foreground",
				children: [
					"Showing ",
					_.length,
					" of ",
					g.length,
					" frames"
				]
			})
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/clicky-parse.ts
function Qi(e) {
	if (typeof e == "string") try {
		return $i(JSON.parse(e));
	} catch (t) {
		return {
			ok: !1,
			message: t instanceof Error ? t.message : "Failed to parse JSON",
			raw: e
		};
	}
	return $i(e);
}
function $i(e) {
	if (!e || typeof e != "object") return {
		ok: !1,
		message: "Payload must be an object",
		raw: String(e ?? "")
	};
	let t = e;
	return "version" in t && t.version === 1 && t.node && ea(t.node) ? {
		ok: !0,
		document: t
	} : ea(t) ? {
		ok: !0,
		document: {
			version: 1,
			node: t
		}
	} : {
		ok: !1,
		message: "Payload is neither a Clicky document nor a Clicky node",
		raw: JSON.stringify(e, null, 2)
	};
}
function ea(e) {
	return !!e && typeof e == "object" && typeof e.kind == "string";
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/debugConsoleSignal.ts
var ta = /* @__PURE__ */ new Set(), na = /* @__PURE__ */ new Set();
function ra(e) {
	return na.add(e), () => {
		na.delete(e);
	};
}
function ia() {
	return ta.size > 0;
}
function aa(e = {}) {
	for (let t of ta) t(e);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/useDebugAction.tsx
function oa() {
	let e = (0, I.useSyncExternalStore)(ra, ia, () => !1);
	return (0, I.useMemo)(() => e ? {
		id: "show-query",
		label: "Debug",
		icon: Qt,
		section: "",
		onSelect: () => aa({ tab: "queries" })
	} : void 0, [e]);
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/clicky-export/ClickyExportDialog.tsx
function sa({ open: e, onClose: t, title: n = "Export", formats: r, scopes: i, onDownload: a, pendingContent: o, allowCloseWhilePending: s = !1 }) {
	let [c, l] = (0, I.useState)(r[0]?.format), [u, d] = (0, I.useState)(i[0]?.scope), [f, p] = (0, I.useState)(!1), [m, h] = (0, I.useState)("");
	(0, I.useEffect)(() => {
		e && h("");
	}, [e]);
	let g = i.find((e) => e.scope === u), _ = async () => {
		if (c) {
			p(!0), h("");
			try {
				await a({
					format: c,
					scope: u
				}), t();
			} catch (e) {
				h(e instanceof Error ? e.message : "The export failed");
			} finally {
				p(!1);
			}
		}
	};
	return /* @__PURE__ */ (0, L.jsx)(Te, {
		open: e,
		onClose: t,
		title: n,
		size: "lg",
		closeOnEsc: !0,
		footer: /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex items-center justify-end gap-2",
			children: [/* @__PURE__ */ (0, L.jsx)(be, {
				variant: "outline",
				onClick: t,
				disabled: f && !s,
				children: f && s ? "Continue in background" : "Cancel"
			}), /* @__PURE__ */ (0, L.jsx)(be, {
				onClick: () => void _(),
				loading: f,
				disabled: !c,
				children: "Download"
			})]
		}),
		children: /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "space-y-4",
			children: [
				f && o,
				m && /* @__PURE__ */ (0, L.jsx)("div", {
					role: "alert",
					className: "rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive",
					children: m
				}),
				i.length > 1 ? /* @__PURE__ */ (0, L.jsxs)("fieldset", { children: [/* @__PURE__ */ (0, L.jsx)("legend", {
					className: "mb-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground",
					children: "Rows"
				}), /* @__PURE__ */ (0, L.jsx)("div", {
					className: "flex flex-wrap gap-2",
					children: i.map((e) => /* @__PURE__ */ (0, L.jsxs)("label", {
						className: P("cursor-pointer rounded-md border px-3 py-1.5 text-sm transition-colors", e.scope === u ? "border-foreground/20 bg-accent text-accent-foreground" : "border-input hover:bg-accent hover:text-accent-foreground"),
						children: [/* @__PURE__ */ (0, L.jsx)("input", {
							type: "radio",
							name: "clicky-export-scope",
							"aria-label": e.label,
							className: "sr-only",
							checked: e.scope === u,
							onChange: () => d(e.scope)
						}), e.label]
					}, e.scope ?? "default"))
				})] }) : i[0]?.label && /* @__PURE__ */ (0, L.jsx)("p", {
					className: "text-sm text-muted-foreground",
					children: i[0].label
				}),
				/* @__PURE__ */ (0, L.jsxs)("fieldset", { children: [/* @__PURE__ */ (0, L.jsx)("legend", {
					className: "mb-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground",
					children: "Format"
				}), /* @__PURE__ */ (0, L.jsx)("div", {
					className: "grid gap-1 sm:grid-cols-2",
					children: r.map((e) => {
						let t = g?.note?.(e.format);
						return /* @__PURE__ */ (0, L.jsxs)("label", {
							className: P("flex cursor-pointer items-start gap-2 rounded-md border px-3 py-2 text-left transition-colors", e.format === c ? "border-foreground/20 bg-accent text-accent-foreground" : "border-transparent hover:bg-accent hover:text-accent-foreground"),
							children: [
								/* @__PURE__ */ (0, L.jsx)("input", {
									type: "radio",
									name: "clicky-export-format",
									"aria-label": e.label,
									className: "sr-only",
									checked: e.format === c,
									onChange: () => l(e.format)
								}),
								e.icon && /* @__PURE__ */ (0, L.jsx)(F, {
									icon: e.icon,
									className: P("mt-0.5 shrink-0 text-sm", e.iconClassName ?? "text-muted-foreground")
								}),
								/* @__PURE__ */ (0, L.jsxs)("span", {
									className: "min-w-0",
									children: [
										/* @__PURE__ */ (0, L.jsx)("span", {
											className: "block text-sm font-medium",
											children: e.label
										}),
										e.description && /* @__PURE__ */ (0, L.jsx)("span", {
											className: "mt-0.5 block text-xs text-muted-foreground",
											children: e.description
										}),
										t && /* @__PURE__ */ (0, L.jsx)("span", {
											className: "mt-0.5 block text-xs text-amber-700 [[data-theme=dark]_&]:text-amber-400",
											children: t
										})
									]
								})
							]
						}, e.format);
					})
				})] })
			]
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/clicky-export/useClickyExport.tsx
function ca({ formats: e, scopes: t, onDownload: n, label: r = "Export…", title: i }) {
	let [a, o] = (0, I.useState)(!1), s = (0, I.useCallback)(() => o(!1), []), c = e.length > 0, l = (0, I.useMemo)(() => c ? {
		id: "export",
		label: r,
		icon: f,
		section: "",
		onSelect: () => o(!0)
	} : void 0, [c, r]);
	return c ? {
		action: l,
		dialog: /* @__PURE__ */ (0, L.jsx)(sa, {
			open: a,
			onClose: s,
			formats: e,
			scopes: t,
			onDownload: n,
			...i ? { title: i } : {}
		})
	} : {
		action: void 0,
		dialog: null
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/Clicky.tsx
var la = ["clicky", "json"], ua = [
	"pdf",
	"html",
	"markdown",
	"yaml",
	"csv",
	"pretty",
	"excel",
	"slack"
], da = [
	"yaml",
	"json",
	"csv",
	"pdf",
	"markdown"
], fa = {
	operations: [],
	operationsLoading: !1
}, pa = (0, I.createContext)(fa);
function ma(e) {
	let t = (0, I.useContext)(zr.QueryClientContext), [n] = (0, I.useState)(() => new zr.QueryClient({ defaultOptions: { queries: {
		staleTime: 3e4,
		retry: 1
	} } })), r = oa(), i = ca({
		formats: (0, I.useMemo)(() => Ia({
			url: e.url,
			download: e.download
		}), [e.download, e.url]),
		scopes: (0, I.useMemo)(() => La(e.download, e.pagination), [e.download, e.pagination]),
		onDownload: (0, I.useCallback)(({ format: t, scope: n }) => za({
			url: e.url ?? "",
			format: t,
			scope: n,
			download: e.download
		}), [e.download, e.url])
	}), a = (0, I.useMemo)(() => [i.action, r].filter((e) => e !== void 0), [i.action, r]), o = /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [i.dialog, /* @__PURE__ */ (0, L.jsx)(ha, {
		...e.commandRuntime ? { commandRuntime: e.commandRuntime } : {},
		...e.onTableRowClick ? { onTableRowClick: e.onTableRowClick } : {},
		...e.getTableRowHref ? { getTableRowHref: e.getTableRowHref } : {},
		...e.isTableRowClickable ? { isTableRowClickable: e.isTableRowClickable } : {},
		...e.search ? { tableSearch: e.search } : {},
		...e.timeRange ? { tableTimeRange: e.timeRange } : {},
		...e.externalFilters ? { tableExternalFilters: e.externalFilters } : {},
		...e.cellFilters ? { tableCellFilters: e.cellFilters } : {},
		...e.onCellFilterChange ? { onTableCellFilterChange: e.onCellFilterChange } : {},
		...e.pagination ? { tablePagination: e.pagination } : {},
		...e.sort === void 0 ? {} : { tableSort: e.sort },
		...e.onSortChange ? { onTableSortChange: e.onSortChange } : {},
		...e.infinite ? { tableInfinite: e.infinite } : {},
		...e.rowSelection ? { tableRowSelection: e.rowSelection } : {},
		...e.hiddenColumns ? { tableHiddenColumns: e.hiddenColumns } : {},
		...e.loading === void 0 ? {} : { tableLoading: e.loading },
		...e.renderRowDetail ? { tableRenderRowDetail: e.renderRowDetail } : {},
		...e.cellRenderers ? { tableCellRenderers: e.cellRenderers } : {},
		...e.renderRowCard ? { tableRenderRowCard: e.renderRowCard } : {},
		...e.detailStyle ? { tableDetailStyle: e.detailStyle } : {},
		...e.detailDialogSize ? { tableDetailDialogSize: e.detailDialogSize } : {},
		...e.detailDialogTitle ? { tableDetailDialogTitle: e.detailDialogTitle } : {},
		...a.length > 0 ? { tableMenuActions: a } : {},
		children: e.url ? /* @__PURE__ */ (0, L.jsx)(va, {
			...e,
			url: e.url
		}) : /* @__PURE__ */ (0, L.jsx)(_a, {
			data: e.data,
			...e.className ? { className: e.className } : {}
		})
	})] });
	return !t && (e.url || e.commandRuntime) ? /* @__PURE__ */ (0, L.jsx)(zr.QueryClientProvider, {
		client: n,
		children: o
	}) : o;
}
function ha({ commandRuntime: e, onTableRowClick: t, getTableRowHref: n, isTableRowClickable: r, tableSearch: i, tableTimeRange: a, tableExternalFilters: o, tableCellFilters: s, onTableCellFilterChange: c, tablePagination: l, tableSort: u, onTableSortChange: d, tableInfinite: f, tableRowSelection: p, tableHiddenColumns: m, tableMenuActions: h, tableLoading: g, tableRenderRowDetail: _, tableCellRenderers: v, tableRenderRowCard: y, tableDetailStyle: b, tableDetailDialogSize: x, tableDetailDialogTitle: S, children: C }) {
	if (!e) {
		let e = t || n || r || i || a || o || s || c || l || u !== void 0 || d || f || p || m || h || g || _ || v || y || b || x || S;
		return /* @__PURE__ */ (0, L.jsx)(pa.Provider, {
			value: e ? {
				...fa,
				onTableRowClick: t,
				getTableRowHref: n,
				isTableRowClickable: r,
				tableSearch: i,
				tableTimeRange: a,
				tableExternalFilters: o,
				tableCellFilters: s,
				onTableCellFilterChange: c,
				tablePagination: l,
				tableSort: u,
				onTableSortChange: d,
				tableInfinite: f,
				tableRowSelection: p,
				tableHiddenColumns: m,
				tableMenuActions: h,
				tableLoading: g,
				tableRenderRowDetail: _,
				tableCellRenderers: v,
				tableRenderRowCard: y,
				tableDetailStyle: b,
				tableDetailDialogSize: x,
				tableDetailDialogTitle: S
			} : fa,
			children: C
		});
	}
	return /* @__PURE__ */ (0, L.jsx)(ga, {
		commandRuntime: e,
		...t ? { onTableRowClick: t } : {},
		...n ? { getTableRowHref: n } : {},
		...r ? { isTableRowClickable: r } : {},
		...i ? { tableSearch: i } : {},
		...a ? { tableTimeRange: a } : {},
		...o ? { tableExternalFilters: o } : {},
		...s ? { tableCellFilters: s } : {},
		...c ? { onTableCellFilterChange: c } : {},
		...l ? { tablePagination: l } : {},
		...u === void 0 ? {} : { tableSort: u },
		...d ? { onTableSortChange: d } : {},
		...f ? { tableInfinite: f } : {},
		...p ? { tableRowSelection: p } : {},
		...m ? { tableHiddenColumns: m } : {},
		...h ? { tableMenuActions: h } : {},
		...g === void 0 ? {} : { tableLoading: g },
		..._ ? { tableRenderRowDetail: _ } : {},
		...v ? { tableCellRenderers: v } : {},
		...y ? { tableRenderRowCard: y } : {},
		...b ? { tableDetailStyle: b } : {},
		...x ? { tableDetailDialogSize: x } : {},
		...S ? { tableDetailDialogTitle: S } : {},
		children: C
	});
}
function ga({ commandRuntime: e, onTableRowClick: t, getTableRowHref: n, isTableRowClickable: r, tableSearch: i, tableTimeRange: a, tableExternalFilters: o, tableCellFilters: s, onTableCellFilterChange: c, tablePagination: l, tableSort: u, onTableSortChange: d, tableInfinite: f, tableRowSelection: p, tableHiddenColumns: m, tableMenuActions: h, tableLoading: g, tableRenderRowDetail: _, tableCellRenderers: v, tableRenderRowCard: y, tableDetailStyle: b, tableDetailDialogSize: x, tableDetailDialogTitle: S, children: C }) {
	let { operations: w, isLoading: T } = Vr(e.client), E = (0, I.useMemo)(() => ({
		commandRuntime: e,
		onTableRowClick: t,
		getTableRowHref: n,
		isTableRowClickable: r,
		tableSearch: i,
		tableTimeRange: a,
		tableExternalFilters: o,
		tableCellFilters: s,
		onTableCellFilterChange: c,
		tablePagination: l,
		tableSort: u,
		onTableSortChange: d,
		tableInfinite: f,
		tableRowSelection: p,
		tableHiddenColumns: m,
		tableMenuActions: h,
		tableLoading: g,
		tableRenderRowDetail: _,
		tableCellRenderers: v,
		tableRenderRowCard: y,
		tableDetailStyle: b,
		tableDetailDialogSize: x,
		tableDetailDialogTitle: S,
		operations: w,
		operationsLoading: T
	}), [
		e,
		n,
		T,
		r,
		t,
		w,
		i,
		a,
		o,
		s,
		c,
		l,
		u,
		d,
		f,
		p,
		m,
		h,
		g,
		_,
		v,
		y,
		b,
		x,
		S
	]);
	return /* @__PURE__ */ (0, L.jsx)(pa.Provider, {
		value: E,
		children: C
	});
}
function _a({ data: e, className: t }) {
	if (e === void 0) return /* @__PURE__ */ (0, L.jsx)(wa, {
		className: t,
		title: "No Clicky data",
		message: "Provide either a local data payload or a URL."
	});
	let n = Qi(e);
	return n.ok ? /* @__PURE__ */ (0, L.jsx)("div", {
		className: t,
		children: /* @__PURE__ */ (0, L.jsx)(z, { node: n.document.node })
	}) : /* @__PURE__ */ (0, L.jsx)(Ca, {
		parsed: n,
		className: t
	});
}
function va({ data: e, dataFormat: t, url: n, view: r, download: i, className: a }) {
	let o = (0, I.useMemo)(() => Ma({
		data: e,
		url: n,
		view: r
	}), [
		e,
		n,
		r
	]), s = (0, I.useMemo)(() => o.filter(Xa), [o]), c = (0, I.useMemo)(() => o.filter(Za), [o]), l = (0, I.useMemo)(() => Na({
		url: n,
		download: i
	}), [i, n]), [u, d] = (0, I.useState)(() => o[0] ?? "clicky");
	(0, I.useEffect)(() => {
		o.includes(u) || d(o[0] ?? "clicky");
	}, [u, o]);
	let f = Ua(n, u), p = e === void 0 ? void 0 : t, m = (0, zr.useQuery)({
		queryKey: [
			"clicky",
			u,
			f
		],
		enabled: Ya(u) && u !== p,
		queryFn: async () => Sa(f, u)
	}), h = u !== p && m.data?.kind === "text" ? m.data.text : void 0, g = u === "clicky" && h !== void 0 ? h : e, _ = (0, I.useMemo)(() => $a(e), [e]), v = u === "json" && h !== void 0 ? $a(h) : _, y = Za(u) ? u : null, b = eo(g), x = l.length > 0 && !b, S = `Fetching ${f}`, C = (0, I.useContext)(pa), w = b && u === "clicky", T = (0, I.useMemo)(() => {
		if (!w || o.length <= 1) return [];
		let e = Va(u);
		return [{
			id: "view",
			label: `View: ${Ha(u)}`,
			icon: e.icon,
			...e.iconClassName ? { iconClassName: e.iconClassName } : {},
			section: "",
			children: o.map((e) => {
				let t = Va(e);
				return {
					id: `view-${e}`,
					label: Ha(e),
					description: t.description,
					icon: t.icon,
					...t.iconClassName ? { iconClassName: t.iconClassName } : {},
					disabled: e === u,
					onSelect: () => d(e)
				};
			}),
			onSelect: () => void 0
		}];
	}, [
		w,
		o,
		u
	]), E = (0, I.useMemo)(() => [...T, ...C.tableMenuActions ?? []], [T, C.tableMenuActions]);
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("space-y-density-3", a),
		children: [!w && /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex flex-wrap items-center gap-2 rounded-md border border-border bg-muted/20 px-density-3 py-density-2",
			children: [
				s.length > 1 && /* @__PURE__ */ (0, L.jsx)("div", {
					role: "radiogroup",
					"aria-label": "Clicky view mode",
					className: "flex flex-wrap gap-1",
					children: s.map((e) => /* @__PURE__ */ (0, L.jsx)("button", {
						type: "button",
						role: "radio",
						"aria-checked": u === e,
						onClick: () => d(e),
						className: P("rounded-md border px-2.5 py-1 text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2", u === e ? "border-foreground/20 bg-accent text-foreground" : "border-transparent text-muted-foreground hover:border-foreground/20 hover:text-foreground"),
						children: Ha(e)
					}, e))
				}),
				c.length > 0 && /* @__PURE__ */ (0, L.jsx)(ya, {
					activeFormat: y,
					formats: c,
					onSelect: d
				}),
				/* @__PURE__ */ (0, L.jsxs)("div", {
					className: "ml-auto flex items-center gap-2",
					children: [Ya(u) && m.isFetching && /* @__PURE__ */ (0, L.jsx)("span", {
						className: "text-xs text-muted-foreground",
						children: "Refreshing…"
					}), x && /* @__PURE__ */ (0, L.jsx)(ba, {
						url: n,
						...i ? { download: i } : {}
					})]
				})
			]
		}), u === "pdf" ? /* @__PURE__ */ (0, L.jsx)(Ta, { src: f }) : u === "html" ? /* @__PURE__ */ (0, L.jsx)(Ea, { src: f }) : u === "excel" ? /* @__PURE__ */ (0, L.jsx)(Da, {
			title: "Excel preview",
			message: "Excel output is available for download, but not inline preview.",
			href: f
		}) : u === "clicky" ? m.isPending && g === void 0 ? /* @__PURE__ */ (0, L.jsx)(wa, {
			title: "Loading Clicky",
			message: S
		}) : m.isError && g === void 0 ? Ti(m.error, `Failed to load ${f}`) : /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [m.isError && e !== void 0 && Ti(m.error, `Refresh failed: ${f}`), /* @__PURE__ */ (0, L.jsx)(pa.Provider, {
			value: {
				...C,
				tableMenuActions: E
			},
			children: /* @__PURE__ */ (0, L.jsx)(_a, {
				data: g,
				className: "flex min-h-0 flex-1 flex-col"
			})
		})] }) : u === "json" ? m.isPending && v === void 0 ? /* @__PURE__ */ (0, L.jsx)(wa, {
			title: "Loading JSON",
			message: S
		}) : m.isError && v === void 0 ? Ti(m.error, `Failed to load ${f}`) : /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [m.isError && _ !== void 0 && Ti(m.error, `Refresh failed: ${f}`), /* @__PURE__ */ (0, L.jsx)(Aa, { value: v })] }) : m.isPending ? /* @__PURE__ */ (0, L.jsx)(wa, {
			title: `Loading ${Ha(u)}`,
			message: S
		}) : m.isError ? Ti(m.error, `Failed to load ${Ha(u)} from ${f}`) : /* @__PURE__ */ (0, L.jsx)(Oa, {
			format: u,
			response: m.data
		})]
	});
}
function ya({ activeFormat: e, formats: t, onSelect: n }) {
	let r = (0, I.useRef)(null), i = (0, I.useRef)(null), [a, o] = (0, I.useState)(!1);
	return xa(a, r, i, () => o(!1)), /* @__PURE__ */ (0, L.jsxs)("div", {
		ref: r,
		className: "relative",
		children: [/* @__PURE__ */ (0, L.jsx)("button", {
			ref: i,
			type: "button",
			"aria-label": "Open additional view menu",
			"aria-haspopup": "menu",
			"aria-expanded": a,
			"aria-pressed": e != null,
			className: P("inline-flex h-[34px] w-[34px] items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2", e && "border-foreground/20 bg-accent text-foreground"),
			onClick: () => o((e) => !e),
			children: /* @__PURE__ */ (0, L.jsx)(F, {
				icon: Ne,
				className: "text-sm"
			})
		}), a && /* @__PURE__ */ (0, L.jsx)("div", {
			role: "menu",
			className: "absolute left-0 top-[calc(100%+0.375rem)] z-50 min-w-[18rem] rounded-md border border-border bg-popover p-1.5 text-popover-foreground shadow-lg shadow-black/5",
			children: t.map((t) => {
				let r = t === e, i = Va(t);
				return /* @__PURE__ */ (0, L.jsxs)("button", {
					type: "button",
					role: "menuitemradio",
					"aria-checked": r,
					className: P("flex w-full items-start gap-3 rounded-md px-3 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:outline-none", r && "bg-accent/70"),
					onClick: () => {
						n(t), o(!1);
					},
					children: [/* @__PURE__ */ (0, L.jsx)(F, {
						icon: i.icon,
						className: P("mt-0.5 shrink-0 text-base", i.iconClassName ?? "text-muted-foreground", r && !i.iconClassName && "text-foreground")
					}), /* @__PURE__ */ (0, L.jsxs)("span", {
						className: "min-w-0 flex-1",
						children: [/* @__PURE__ */ (0, L.jsxs)("span", {
							className: "flex items-center justify-between gap-3",
							children: [/* @__PURE__ */ (0, L.jsx)("span", {
								className: "font-medium",
								children: i.label
							}), r && /* @__PURE__ */ (0, L.jsx)(F, {
								icon: Ue,
								className: "text-xs text-muted-foreground"
							})]
						}), /* @__PURE__ */ (0, L.jsx)("span", {
							className: "mt-0.5 block text-xs text-muted-foreground",
							children: i.description
						})]
					})]
				}, t);
			})
		})]
	});
}
function ba({ url: e, download: t }) {
	let [n, r] = (0, I.useState)(!1), i = (0, I.useMemo)(() => Ia({
		url: e,
		download: t
	}), [t, e]), a = (0, I.useMemo)(() => La(t), [t]);
	return i.length === 0 ? null : /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsxs)("button", {
		type: "button",
		"aria-label": t?.label ? `Export ${t.label}` : "Export",
		className: "inline-flex items-center gap-2 rounded-md border border-input bg-background px-3 py-1.5 text-xs font-medium transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
		onClick: () => r(!0),
		children: [/* @__PURE__ */ (0, L.jsx)(F, {
			icon: f,
			className: "text-sm"
		}), /* @__PURE__ */ (0, L.jsx)("span", { children: "Export" })]
	}), /* @__PURE__ */ (0, L.jsx)(sa, {
		open: n,
		onClose: () => r(!1),
		formats: i,
		scopes: a,
		onDownload: ({ format: n, scope: r }) => za({
			url: e,
			format: n,
			scope: r,
			...t ? { download: t } : {}
		})
	})] });
}
function xa(e, t, n, r) {
	De(e, () => {
		r(), n.current?.focus();
	}), (0, I.useEffect)(() => {
		if (!e) return;
		let n = (e) => {
			t.current?.contains(e.target) || r();
		};
		return document.addEventListener("mousedown", n), () => {
			document.removeEventListener("mousedown", n);
		};
	}, [
		r,
		e,
		t,
		n
	]);
}
function Sa(e, t) {
	return fetch(e, { headers: { Accept: Va(t).accept } }).then(async (n) => {
		if (!n.ok) throw Qr(n, "GET", e, await Zr(n, "json"));
		let r = n.headers.get("Content-Type") ?? "";
		return t === "excel" ? {
			kind: "blob",
			blob: await n.blob(),
			contentType: r
		} : {
			kind: "text",
			text: await n.text(),
			contentType: r
		};
	});
}
function Ca({ parsed: e, className: t }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("rounded-md border border-destructive/30 bg-destructive/5 p-density-3", t),
		children: [/* @__PURE__ */ (0, L.jsx)("div", {
			className: "text-sm font-medium text-destructive",
			children: "Invalid Clicky payload"
		}), /* @__PURE__ */ (0, L.jsxs)("pre", {
			className: "mt-2 whitespace-pre-wrap break-all text-xs text-muted-foreground",
			children: [e.message, e.raw ? `\n\n${e.raw}` : ""]
		})]
	});
}
function wa({ title: e, message: t, tone: n = "default", className: r }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("rounded-md border p-density-3", n === "destructive" && "border-destructive/30 bg-destructive/5", n === "warning" && "border-yellow-500/30 bg-yellow-500/5", n === "default" && "border-border bg-muted/30", r),
		children: [/* @__PURE__ */ (0, L.jsx)("div", {
			className: P("text-sm font-medium", n === "destructive" && "text-destructive", n === "warning" && "text-yellow-700 dark:text-yellow-400"),
			children: e
		}), /* @__PURE__ */ (0, L.jsx)("div", {
			className: "mt-1 text-xs text-muted-foreground",
			children: t
		})]
	});
}
function Ta({ src: e }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "overflow-hidden rounded-md border border-border bg-background",
		children: [/* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex items-center justify-between border-b border-border px-3 py-2 text-xs text-muted-foreground",
			children: [/* @__PURE__ */ (0, L.jsx)("span", { children: "PDF preview" }), /* @__PURE__ */ (0, L.jsx)("a", {
				href: e,
				target: "_blank",
				rel: "noreferrer",
				className: "text-primary hover:underline",
				children: "Open in new tab"
			})]
		}), /* @__PURE__ */ (0, L.jsx)("iframe", {
			title: "Clicky PDF preview",
			src: e,
			className: "h-[720px] w-full bg-background"
		})]
	});
}
function Ea({ src: e }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "overflow-hidden rounded-md border border-border bg-background",
		children: [/* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex items-center justify-between border-b border-border px-3 py-2 text-xs text-muted-foreground",
			children: [/* @__PURE__ */ (0, L.jsx)("span", { children: "HTML preview" }), /* @__PURE__ */ (0, L.jsx)("a", {
				href: e,
				target: "_blank",
				rel: "noreferrer",
				className: "text-primary hover:underline",
				children: "Open in new tab"
			})]
		}), /* @__PURE__ */ (0, L.jsx)("iframe", {
			title: "Clicky HTML preview",
			src: e,
			className: "h-[720px] w-full bg-background",
			sandbox: "allow-same-origin"
		})]
	});
}
function Da({ title: e, message: t, href: n }) {
	return /* @__PURE__ */ (0, L.jsx)(wa, {
		title: e,
		message: /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsxs)("span", { children: [t, " "] }), /* @__PURE__ */ (0, L.jsx)("a", {
			href: n,
			target: "_blank",
			rel: "noreferrer",
			className: "text-primary hover:underline",
			children: "Open format"
		})] })
	});
}
function Oa({ format: e, response: t }) {
	let n = Va(e);
	if (e === "slack") {
		let e = t?.kind === "text" ? $a(t.text) : void 0;
		return /* @__PURE__ */ (0, L.jsx)(Aa, {
			value: e,
			emptyLabel: n.label
		});
	}
	return /* @__PURE__ */ (0, L.jsx)(ka, {
		title: `${n.label} output`,
		content: t?.kind === "text" ? t.text : ""
	});
}
function ka({ title: e, content: t }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "overflow-hidden rounded-md border border-border bg-background",
		children: [/* @__PURE__ */ (0, L.jsx)("div", {
			className: "border-b border-border px-3 py-2 text-xs text-muted-foreground",
			children: e
		}), /* @__PURE__ */ (0, L.jsx)("pre", {
			"aria-label": "Clicky text preview",
			className: "overflow-auto whitespace-pre-wrap break-words px-3 py-3 font-mono text-xs text-foreground",
			children: t || "No content"
		})]
	});
}
function Aa({ value: e, emptyLabel: t = "JSON" }) {
	let n = (0, I.useMemo)(() => no(e), [e]);
	return e === void 0 ? /* @__PURE__ */ (0, L.jsx)(wa, {
		title: `No ${t} payload`,
		message: "The response did not include a JSON body to inspect."
	}) : /* @__PURE__ */ (0, L.jsx)("div", {
		"aria-label": "JSON tree",
		className: "overflow-hidden rounded-md border border-border bg-background",
		children: /* @__PURE__ */ (0, L.jsx)($e, {
			roots: n,
			className: "min-h-[12rem]",
			showControls: n.some((e) => (e.children?.length ?? 0) > 0),
			getKey: (e) => e.id,
			getChildren: (e) => e.children,
			renderRow: ({ node: e }) => /* @__PURE__ */ (0, L.jsx)(ja, { node: e }),
			rowClass: () => "hover:bg-accent",
			empty: /* @__PURE__ */ (0, L.jsx)("div", {
				className: "px-3 py-4 text-sm text-muted-foreground",
				children: "No JSON fields."
			}),
			toolbarClassName: "pl-3"
		})
	});
}
function ja({ node: e }) {
	let t = typeof e.value == "string" ? "text-emerald-700 dark:text-emerald-400" : typeof e.value == "number" ? "text-sky-700 dark:text-sky-400" : typeof e.value == "boolean" || e.value === null ? "text-violet-700 dark:text-violet-400" : "text-muted-foreground";
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex min-w-0 items-start gap-2 font-mono text-xs",
		children: [
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: "shrink-0 text-foreground",
				children: e.key
			}),
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: "shrink-0 text-muted-foreground",
				children: ":"
			}),
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: P("min-w-0 break-words", t),
				children: e.preview
			})
		]
	});
}
function Ma({ data: e, url: t, view: n }) {
	let r = Qa(n, "clicky", e, t), i = Qa(n, "json", e, t), a = [];
	r && (t || e !== void 0) && a.push("clicky"), i && (t || e !== void 0) && a.push("json");
	for (let r of ua) Qa(n, r, e, t) && t && a.push(r);
	return a.length === 0 && (t || e !== void 0) && a.push("clicky"), a;
}
function Na({ url: e, download: t }) {
	return !(t ? t.all ?? !0 : e) || !e ? [] : t?.formats?.length ? [...t.formats] : [...da];
}
function Pa(e) {
	return e?.scopes?.length ? e.scopes : [void 0];
}
function Fa(e, t, n) {
	let r = n?.formatMaxRows?.[e];
	if (r) return `Limited to ${r.toLocaleString()} rows`;
	if (t === "all" && n?.allRowsMode === "buffered") return "Compatibility mode buffers the complete result";
	if (t === "all" && n?.allRowsMode === "streaming") return "Streams rows as they are read";
}
function Ia({ url: e, download: t }) {
	return e ? Na({
		url: e,
		download: t
	}).map((e) => {
		let t = Va(e);
		return {
			format: e,
			label: t.label,
			description: t.description,
			icon: t.icon,
			...t.iconClassName ? { iconClassName: t.iconClassName } : {}
		};
	}) : [];
}
function La(e, t) {
	return Pa(e).map((n) => ({
		scope: n,
		label: n === "all" || n === void 0 ? Ra(t) : "Current page",
		note: (t) => Fa(t, n, e)
	}));
}
function Ra(e) {
	let t = e?.total;
	if (t === void 0) return "All rows";
	let n = t.toLocaleString();
	return e?.totalRelation === "gte" ? `All ${n}+ rows` : `All ${n} rows`;
}
async function za({ url: e, format: t, scope: n, download: r }) {
	let i = "";
	if (await Ba({
		url: e,
		format: t,
		...n ? { scope: n } : {},
		...r ? { download: r } : {},
		onError: (e) => {
			i = e;
		}
	}), i) throw Error(i);
}
async function Ba({ url: e, format: t, scope: n, download: r, onPending: i, onError: a }) {
	a?.(""), i?.(!0);
	try {
		let i = r?.prepare ? await r.prepare({
			format: t,
			...n ? { scope: n } : {}
		}) : {
			url: e,
			...r?.label ? { label: r.label } : {}
		};
		if (!i.url.trim()) throw Error("The prepared download URL is empty");
		co(Wa(i.url, t, i.label ?? r?.label, n));
	} catch (e) {
		a?.(e instanceof Error ? e.message : "Preparing the download failed");
	} finally {
		i?.(!1);
	}
}
function Va(e) {
	switch (e) {
		case "clicky": return {
			label: "Clicky",
			description: "Rendered Clicky JSON with the rich Clicky viewer",
			icon: yt,
			accept: "application/json+clicky, application/clicky+json;q=0.9, application/json;q=0.8,*/*;q=0.7"
		};
		case "json": return {
			label: "JSON",
			description: "Plain JSON for inspecting the raw response body",
			icon: w,
			accept: "application/json, text/plain;q=0.8,*/*;q=0.7"
		};
		case "ndjson": return {
			label: "NDJSON",
			description: "Newline-delimited JSON for streaming data pipelines",
			icon: w,
			accept: "application/x-ndjson, application/ndjson;q=0.9,*/*;q=0.7"
		};
		case "pdf": return {
			label: "PDF",
			description: "Portable document for sharing and printing",
			icon: C,
			iconClassName: "text-rose-600 dark:text-rose-400",
			accept: "application/pdf, */*;q=0.7"
		};
		case "html": return {
			label: "HTML",
			description: "Browser-ready HTML preview of the formatted output",
			icon: tt,
			accept: "text/html, */*;q=0.7"
		};
		case "markdown": return {
			label: "Markdown",
			description: "Markdown formatted for docs, comments, and chat",
			icon: S,
			accept: "text/markdown, text/plain;q=0.8,*/*;q=0.7"
		};
		case "yaml": return {
			label: "YAML",
			description: "YAML for config-friendly inspection and export",
			icon: tt,
			iconClassName: "text-violet-600 dark:text-violet-400",
			accept: "application/yaml, text/yaml;q=0.9, text/plain;q=0.8,*/*;q=0.7"
		};
		case "csv": return {
			label: "CSV",
			description: "Comma-separated values for spreadsheets and imports",
			icon: d,
			iconClassName: "text-emerald-600 dark:text-emerald-400",
			accept: "text/csv, text/plain;q=0.8,*/*;q=0.7"
		};
		case "pretty": return {
			label: "Pretty",
			description: "Human-readable plain text output from the formatter",
			icon: tt,
			accept: "text/plain, */*;q=0.7"
		};
		case "excel": return {
			label: "Excel",
			description: "Spreadsheet workbook for offline analysis",
			icon: en,
			accept: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet, */*;q=0.7"
		};
		case "slack": return {
			label: "Slack",
			description: "Slack Block Kit JSON for chat-native output",
			icon: Fe,
			accept: "application/vnd.slack.block-kit+json, application/json;q=0.8,*/*;q=0.7"
		};
	}
}
function Ha(e) {
	return Va(e).label;
}
function Ua(e, t) {
	let n = typeof window > "u" ? "http://localhost" : window.location.origin, r = new URL(e, n);
	return r.searchParams.set("format", t === "clicky" ? "clicky-json" : t), lo(e) ? r.toString() : `${r.pathname}${r.search}${r.hash}`;
}
function Wa(e, t, n, r) {
	let i = Ua(e, t), a = typeof window > "u" ? "http://localhost" : window.location.origin, o = new URL(i, a);
	return o.searchParams.set("filename", Ga(e, t, n)), o.searchParams.set("_download", Date.now().toString()), r && o.searchParams.set("scope", r), r === "all" && (o.searchParams.delete("limit"), o.searchParams.delete("offset"), o.searchParams.delete("page")), lo(i) ? o.toString() : `${o.pathname}${o.search}${o.hash}`;
}
function Ga(e, t, n) {
	return `${qa(n || Ka(e) || "clicky-output")}${Ja(t)}`;
}
function Ka(e) {
	let t = typeof window > "u" ? "http://localhost" : window.location.origin;
	return new URL(e, t).pathname.split("/").filter(Boolean).at(-1) ?? "";
}
function qa(e) {
	return _e(e.trim().replace(/\.[a-z0-9]+$/i, "").replace(/[^a-z0-9._-]+/gi, "-").replace(/-+/g, "-"), "._-") || "clicky-output";
}
function Ja(e) {
	switch (e) {
		case "clicky": return ".clicky.json";
		case "json": return ".json";
		case "ndjson": return ".ndjson";
		case "pdf": return ".pdf";
		case "html": return ".html";
		case "markdown": return ".md";
		case "yaml": return ".yaml";
		case "csv": return ".csv";
		case "pretty": return ".txt";
		case "excel": return ".xlsx";
		case "slack": return ".slack.json";
	}
}
function Ya(e) {
	return e !== "pdf" && e !== "html" && e !== "excel";
}
function Xa(e) {
	return la.includes(e);
}
function Za(e) {
	return ua.includes(e);
}
function Qa(e, t, n, r) {
	let i = t === "clicky" || t === "json" ? r != null || n !== void 0 : !!r;
	return e === void 0 ? i : Array.isArray(e) ? e.length === 0 ? !1 : t === "clicky" ? e.includes("clicky") || e.includes("json") : e.includes(t) : t in e ? !!e[t] : t === "clicky" && e.json === !0 || i;
}
function $a(e) {
	if (e === void 0) return;
	if (typeof e != "string") return e;
	let t = e.trim();
	if (!t) return "";
	try {
		return JSON.parse(t);
	} catch {
		return e;
	}
}
function eo(e) {
	if (e === void 0) return !1;
	let t = Qi(e);
	return t.ok ? to(t.document.node) : !1;
}
function to(e) {
	return e ? !!(e.kind === "table" || e.children?.some(to) || e.items?.some(to) || e.fields?.some((e) => to(e.value)) || e.label && to(e.label) || e.content && to(e.content) || e.tooltip && to(e.tooltip) || e.columns?.some((e) => e.header && to(e.header)) || e.rows?.some((e) => Object.values(e.cells).some((e) => to(e))) || e.rows?.some((e) => e.detail && to(e.detail))) : !1;
}
function no(e) {
	return Array.isArray(e) ? e.map((e, t) => ro(String(t), e, `$[${t}]`)) : oo(e) ? Object.entries(e).map(([e, t]) => ro(e, t, `$.${e}`)) : e === void 0 ? [] : [ro("$", e, "$")];
}
function ro(e, t, n) {
	let r = io(t, n);
	return {
		id: n,
		key: e,
		value: t,
		preview: ao(t),
		...r ? { children: r } : {}
	};
}
function io(e, t) {
	if (Array.isArray(e)) return e.map((e, n) => ro(String(n), e, `${t}[${n}]`));
	if (oo(e)) return Object.entries(e).map(([e, n]) => ro(e, n, `${t}.${e}`));
}
function ao(e) {
	if (Array.isArray(e)) return `[${e.length} item${e.length === 1 ? "" : "s"}]`;
	if (oo(e)) {
		let t = Object.keys(e).length;
		return `{${t} key${t === 1 ? "" : "s"}}`;
	}
	return typeof e == "string" ? JSON.stringify(e) : e === void 0 ? "undefined" : String(e);
}
function oo(e) {
	return typeof e == "object" && !!e && !Array.isArray(e);
}
var so = "clicky-download-frame";
function co(e) {
	if (typeof document > "u") return;
	let t = document.getElementById(so);
	t || (t = document.createElement("iframe"), t.id = so, t.name = so, t.title = "Clicky download target", t.setAttribute("aria-hidden", "true"), t.tabIndex = -1, t.style.display = "none", document.body.appendChild(t)), t.src = e;
}
function lo(e) {
	return /^[a-z][a-z\d+\-.]*:/i.test(e) || e.startsWith("//");
}
var uo = { "text-muted": "text-muted-foreground" };
function fo(e) {
	return e.split(/\s+/).map((e) => uo[e] ?? e).join(" ");
}
function po(e) {
	let t = e.style?.className;
	if (!t) return e;
	let n = fo(t);
	return n === t ? e : {
		...e,
		style: {
			...e.style,
			className: n
		}
	};
}
function z({ node: e }) {
	if (!e) return null;
	let t = po(e);
	switch (t.kind) {
		case "text": return /* @__PURE__ */ (0, L.jsx)(V, { node: t });
		case "link": return /* @__PURE__ */ (0, L.jsx)(mo, { node: t });
		case "link-command": return /* @__PURE__ */ (0, L.jsx)(ho, { node: t });
		case "icon": return /* @__PURE__ */ (0, L.jsx)(bo, { node: t });
		case "list": return /* @__PURE__ */ (0, L.jsx)(Io, { node: t });
		case "map": return /* @__PURE__ */ (0, L.jsx)(Lo, { node: t });
		case "table": return /* @__PURE__ */ (0, L.jsx)(zo, { node: t });
		case "tree": return /* @__PURE__ */ (0, L.jsx)(ms, { node: t });
		case "object-graph": return /* @__PURE__ */ (0, L.jsx)(Ri, { roots: t.objects ?? [] });
		case "execution-tree": return /* @__PURE__ */ (0, L.jsx)(Vi, { roots: t.executionRoots ?? [] });
		case "code": return /* @__PURE__ */ (0, L.jsx)(Ae, {
			language: t.language,
			source: t.source ?? t.plain ?? "",
			highlightedHtml: t.highlightedHtml
		});
		case "heading": return /* @__PURE__ */ (0, L.jsx)(xo, { node: t });
		case "blockquote": return /* @__PURE__ */ (0, L.jsx)(wo, { node: t });
		case "admonition": return /* @__PURE__ */ (0, L.jsx)(To, { node: t });
		case "footnote-ref": return /* @__PURE__ */ (0, L.jsx)(ko, { node: t });
		case "footnote": return /* @__PURE__ */ (0, L.jsx)(Ao, { node: t });
		case "footnotes": return /* @__PURE__ */ (0, L.jsx)(jo, { node: t });
		case "collapsed": return /* @__PURE__ */ (0, L.jsx)(vs, { node: t });
		case "stacktrace": return /* @__PURE__ */ (0, L.jsx)(_s, { node: t });
		case "button": return /* @__PURE__ */ (0, L.jsx)(ys, { node: t });
		case "button-group": return /* @__PURE__ */ (0, L.jsx)(bs, { node: t });
		case "html": return /* @__PURE__ */ (0, L.jsx)(xs, { node: t });
		case "comment": return /* @__PURE__ */ (0, L.jsx)(B, { node: t });
		case "badge": return /* @__PURE__ */ (0, L.jsx)(gs, { node: t });
		default: return /* @__PURE__ */ (0, L.jsx)("pre", {
			className: "rounded-md border border-border bg-muted p-density-3 text-xs",
			children: JSON.stringify(t, null, 2)
		});
	}
}
function B({ node: e }) {
	let t = e.text ?? e.plain, n = /* @__PURE__ */ (0, L.jsx)(H, { node: e }), r = e.style?.maxWidth && e.style.maxWidth > 0 ? Math.min(e.style.maxWidth, 50) : 50, i = e.style?.maxLines && e.style.maxLines > 0 ? Math.min(e.style.maxLines, 3) : 3, a = Ss({
		...e.style,
		maxWidth: r,
		maxLines: i
	}, t);
	return !t && !e.children?.length ? null : /* @__PURE__ */ (0, L.jsx)("div", {
		className: P("mt-1 text-xs leading-snug text-muted-foreground", e.style?.className),
		style: a,
		title: e.tooltip?.plain ?? t,
		children: n
	});
}
function V({ node: e }) {
	let t = Ss(e.style, e.text ?? e.plain), n = /* @__PURE__ */ (0, L.jsx)(H, { node: e });
	return !e.style && !e.tooltip ? /* @__PURE__ */ (0, L.jsx)(L.Fragment, { children: n }) : /* @__PURE__ */ (0, L.jsx)("span", {
		style: t,
		title: e.tooltip?.plain,
		className: P(e.style?.className, (e.text ?? "").includes("\n") && "whitespace-pre-wrap"),
		children: n
	});
}
function H({ node: e }) {
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [e.text, e.children?.map((e, t) => /* @__PURE__ */ (0, L.jsx)(I.Fragment, { children: /* @__PURE__ */ (0, L.jsx)(z, { node: e }) }, t))] });
}
function mo({ node: e }) {
	let t = /* @__PURE__ */ (0, L.jsx)(H, { node: e }), n = Ss(e.style, e.text ?? e.plain), r = Ns(e.target), i = r === "_blank" ? "noopener noreferrer" : void 0;
	return e.href ? /* @__PURE__ */ (0, L.jsx)("a", {
		href: e.href,
		target: r,
		rel: i,
		title: e.tooltip?.plain,
		style: n,
		className: P(e.style?.className, "inline cursor-pointer"),
		children: t
	}) : /* @__PURE__ */ (0, L.jsx)("span", {
		style: n,
		children: t
	});
}
function ho({ node: e }) {
	let t = (0, I.useContext)(pa), n = (0, I.useMemo)(() => ws(e), [e]), r = (0, I.useMemo)(() => Ts(n, t.commandRuntime, t.operations), [
		n,
		t.commandRuntime,
		t.operations
	]), i = n.target ?? "Dialog", a = t.commandRuntime != null, [o, s] = (0, I.useState)(!1), [c, l] = (0, I.useState)(!1);
	if (!a) return /* @__PURE__ */ (0, L.jsx)("span", {
		title: e.tooltip?.plain,
		className: P(e.style?.className),
		children: /* @__PURE__ */ (0, L.jsx)(H, { node: e })
	});
	if (i === "_self" || i === "_window" || i === "_tab") {
		let n = t.commandRuntime?.hrefForCommand?.(r) ?? js(r);
		return n ? /* @__PURE__ */ (0, L.jsx)("a", {
			href: n,
			target: Ns(i),
			rel: i === "_self" ? void 0 : "noopener noreferrer",
			title: e.tooltip?.plain,
			style: Ss(e.style, e.text ?? e.plain),
			className: P(e.style?.className, "inline cursor-pointer"),
			children: /* @__PURE__ */ (0, L.jsx)(H, { node: e })
		}) : /* @__PURE__ */ (0, L.jsx)("span", {
			title: e.tooltip?.plain,
			className: P(e.style?.className),
			children: /* @__PURE__ */ (0, L.jsx)(H, { node: e })
		});
	}
	return i === "_clicky" ? /* @__PURE__ */ (0, L.jsx)(go, {
		node: e,
		onClick: () => t.commandRuntime?.onNavigate?.(r)
	}) : i === "Hover" ? /* @__PURE__ */ (0, L.jsx)(Ke, {
		delay: 75,
		cardClassName: "w-[28rem] max-w-[calc(100vw-2rem)] whitespace-normal p-0",
		trigger: /* @__PURE__ */ (0, L.jsx)(go, { node: e }),
		children: /* @__PURE__ */ (0, L.jsx)("div", {
			className: "max-h-[26rem] overflow-auto p-density-3",
			children: /* @__PURE__ */ (0, L.jsx)(vo, { resolved: r })
		})
	}) : i === "Expand" ? /* @__PURE__ */ (0, L.jsxs)("span", {
		className: "inline",
		children: [/* @__PURE__ */ (0, L.jsx)(go, {
			node: e,
			ariaExpanded: c,
			onClick: () => l((e) => !e)
		}), c && /* @__PURE__ */ (0, L.jsx)("div", {
			className: "mt-2 rounded-md border border-border bg-background p-density-3",
			children: /* @__PURE__ */ (0, L.jsx)(vo, { resolved: r })
		})]
	}) : /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsx)(go, {
		node: e,
		ariaExpanded: o,
		onClick: () => s(!0)
	}), /* @__PURE__ */ (0, L.jsx)(_o, {
		open: o,
		onClose: () => s(!1),
		resolved: r
	})] });
}
function go({ node: e, onClick: t, ariaExpanded: n }) {
	let r = Ss(e.style, e.text ?? e.plain);
	return /* @__PURE__ */ (0, L.jsx)("button", {
		type: "button",
		title: e.tooltip?.plain,
		"aria-expanded": n,
		onClick: (e) => {
			e.preventDefault(), t?.();
		},
		style: r,
		className: P("inline cursor-pointer appearance-none border-0 bg-transparent p-0 text-left align-baseline text-inherit", e.style?.className),
		children: /* @__PURE__ */ (0, L.jsx)(H, { node: e })
	});
}
function _o({ open: e, onClose: t, resolved: n }) {
	let r = (0, I.useContext)(pa), i = n.operation, a = i?.operation.parameters ?? [], o = (0, I.useMemo)(() => Os(a, n.request), [a, n.request]), [s, c] = (0, I.useState)(null), [l, u] = (0, I.useState)(""), [d, f] = (0, I.useState)(!1), [p, m] = (0, I.useState)(!1), h = e && n.request.autoRun === !0 && i != null && ks(a, o).length === 0;
	(0, I.useEffect)(() => {
		e || (m(!1), c(null), u(""));
	}, [e, n.request.command]);
	async function g(e) {
		if (r.commandRuntime && i) {
			f(!0), u("");
			try {
				let t = await As(r.commandRuntime.client, i, e);
				c(t);
			} catch (e) {
				c(null), u(e instanceof Error ? e.message : String(e ?? "Unknown error"));
			} finally {
				f(!1);
			}
		}
	}
	return (0, I.useEffect)(() => {
		h && !p && (m(!0), g(o));
	}, [
		p,
		o,
		h
	]), /* @__PURE__ */ (0, L.jsx)(Te, {
		open: e,
		onClose: t,
		title: i?.operation.summary || n.request.command,
		size: "xl",
		children: /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "space-y-4",
			children: [
				i ? /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "flex flex-wrap items-center gap-3 text-xs text-muted-foreground",
					children: [/* @__PURE__ */ (0, L.jsx)("span", {
						className: "rounded-md border border-border bg-muted px-2 py-1 font-medium uppercase",
						children: i.method
					}), /* @__PURE__ */ (0, L.jsx)("code", { children: i.path })]
				}) : r.operationsLoading ? /* @__PURE__ */ (0, L.jsx)(wa, {
					title: "Loading commands",
					message: "Resolving the command definition."
				}) : /* @__PURE__ */ (0, L.jsx)(wa, {
					title: "Unknown command",
					message: `No operation matched "${n.request.command}".`,
					tone: "destructive"
				}),
				i && /* @__PURE__ */ (0, L.jsx)("div", {
					className: "rounded-md border border-border bg-muted/20 p-density-3",
					children: /* @__PURE__ */ (0, L.jsx)(Yr, {
						client: r.commandRuntime.client,
						path: i.path,
						method: i.method,
						parameters: a,
						initialValues: o,
						isSubmitting: d,
						submitLabel: "Run command",
						submittingLabel: "Running…",
						onSubmit: g
					})
				}),
				l ? /* @__PURE__ */ (0, L.jsx)(wa, {
					title: "Command failed",
					message: l,
					tone: "destructive"
				}) : /* @__PURE__ */ (0, L.jsx)(yo, {
					response: s,
					pending: d,
					emptyMessage: "Run the command to load a Clicky response."
				})
			]
		})
	});
}
function vo({ resolved: e }) {
	let t = (0, I.useContext)(pa), n = e.operation, r = n?.operation.parameters ?? [], i = (0, I.useMemo)(() => Os(r, e.request), [r, e.request]), a = (0, I.useMemo)(() => ks(r, i), [i, r]), o = (0, zr.useQuery)({
		queryKey: [
			"clicky-command",
			e.request.command,
			n?.method,
			n?.path,
			i
		],
		enabled: t.commandRuntime != null && n != null && a.length === 0,
		retry: 0,
		queryFn: async () => As(t.commandRuntime.client, n, i)
	});
	return t.commandRuntime ? n ? a.length > 0 ? /* @__PURE__ */ (0, L.jsx)(wa, {
		title: "Missing required parameters",
		message: `This link is missing: ${a.map((e) => Cs(e.name)).join(", ")}.`,
		tone: "warning"
	}) : o.isPending ? /* @__PURE__ */ (0, L.jsx)(wa, {
		title: "Running command",
		message: "Loading Clicky response…"
	}) : o.isError ? /* @__PURE__ */ (0, L.jsx)(wa, {
		title: "Command failed",
		message: o.error instanceof Error ? o.error.message : "Request failed",
		tone: "destructive"
	}) : /* @__PURE__ */ (0, L.jsx)(yo, {
		response: o.data,
		emptyMessage: "No response body returned."
	}) : t.operationsLoading ? /* @__PURE__ */ (0, L.jsx)(wa, {
		title: "Loading commands",
		message: "Resolving the command definition."
	}) : /* @__PURE__ */ (0, L.jsx)(wa, {
		title: "Unknown command",
		message: `No operation matched "${e.request.command}".`,
		tone: "destructive"
	}) : /* @__PURE__ */ (0, L.jsx)(wa, {
		title: "Command runtime unavailable",
		message: "This Clicky view was rendered without a command runtime.",
		tone: "warning"
	});
}
function yo({ response: e, pending: t = !1, emptyMessage: n }) {
	if (t) return /* @__PURE__ */ (0, L.jsx)(wa, {
		title: "Running command",
		message: "Loading Clicky response…"
	});
	if (!e) return /* @__PURE__ */ (0, L.jsx)(wa, {
		title: "No response",
		message: n
	});
	let r = Ps(e) ? e.parsed ?? Fi(e) : e, i = Ps(e) ? e.stdout || e.output || e.message || "" : typeof e == "string" ? e : JSON.stringify(e, null, 2), a = typeof r == "string" || typeof r == "object" && r ? r : i, o = a === "" ? null : Qi(a);
	return o?.ok ? /* @__PURE__ */ (0, L.jsx)("div", {
		className: "rounded-md border border-border bg-background p-density-3",
		children: /* @__PURE__ */ (0, L.jsx)(z, { node: o.document.node })
	}) : typeof r == "object" && r ? /* @__PURE__ */ (0, L.jsx)(Aa, { value: r }) : /* @__PURE__ */ (0, L.jsx)(ka, {
		title: "Response body",
		content: i
	});
}
function bo({ node: e }) {
	let t = Ss(e.style, e.plain ?? e.unicode);
	return /* @__PURE__ */ (0, L.jsx)("span", {
		style: t,
		title: e.tooltip?.plain,
		className: "inline-flex items-center",
		children: e.iconify ? /* @__PURE__ */ (0, L.jsx)(F, { name: e.iconify }) : /* @__PURE__ */ (0, L.jsx)("span", { children: e.unicode ?? e.plain })
	});
}
function xo({ node: e }) {
	let t = So(e.level), n = e.content ?? Fo(e), r = Ss(e.style, U(n)), i = P(Co(t), "text-foreground", e.style?.className);
	switch (t) {
		case 1: return /* @__PURE__ */ (0, L.jsx)("h1", {
			className: i,
			style: r,
			children: n && /* @__PURE__ */ (0, L.jsx)(z, { node: n })
		});
		case 2: return /* @__PURE__ */ (0, L.jsx)("h2", {
			className: i,
			style: r,
			children: n && /* @__PURE__ */ (0, L.jsx)(z, { node: n })
		});
		case 3: return /* @__PURE__ */ (0, L.jsx)("h3", {
			className: i,
			style: r,
			children: n && /* @__PURE__ */ (0, L.jsx)(z, { node: n })
		});
		case 4: return /* @__PURE__ */ (0, L.jsx)("h4", {
			className: i,
			style: r,
			children: n && /* @__PURE__ */ (0, L.jsx)(z, { node: n })
		});
		case 5: return /* @__PURE__ */ (0, L.jsx)("h5", {
			className: i,
			style: r,
			children: n && /* @__PURE__ */ (0, L.jsx)(z, { node: n })
		});
		default: return /* @__PURE__ */ (0, L.jsx)("h6", {
			className: i,
			style: r,
			children: n && /* @__PURE__ */ (0, L.jsx)(z, { node: n })
		});
	}
}
function So(e) {
	return Number.isFinite(e) ? Math.min(6, Math.max(1, Math.trunc(e ?? 1))) : 1;
}
function Co(e) {
	switch (e) {
		case 1: return "text-2xl font-semibold leading-tight";
		case 2: return "text-xl font-semibold leading-tight";
		case 3: return "text-lg font-semibold leading-snug";
		case 4: return "text-base font-semibold leading-snug";
		case 5: return "text-sm font-semibold leading-snug";
		default: return "text-xs font-semibold uppercase leading-snug text-muted-foreground";
	}
}
function wo({ node: e }) {
	let t = e.content ?? Fo(e);
	return t ? /* @__PURE__ */ (0, L.jsx)("blockquote", {
		className: "border-l-4 border-border pl-density-3 text-sm leading-relaxed text-muted-foreground",
		children: /* @__PURE__ */ (0, L.jsx)(z, { node: t })
	}) : null;
}
function To({ node: e }) {
	let t = Eo(e.severity), n = e.label ?? {
		kind: "text",
		text: Cs(t),
		plain: Cs(t)
	}, r = Oo(t);
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		role: t === "danger" || t === "warning" ? "alert" : "note",
		"data-severity": t,
		className: P("rounded-md border px-density-3 py-density-2", r.container),
		children: [/* @__PURE__ */ (0, L.jsx)("div", {
			className: P("text-sm font-semibold", r.title),
			children: /* @__PURE__ */ (0, L.jsx)(z, { node: n })
		}), e.content && /* @__PURE__ */ (0, L.jsx)("div", {
			className: P("mt-1 text-sm leading-relaxed", r.body),
			children: /* @__PURE__ */ (0, L.jsx)(z, { node: e.content })
		})]
	});
}
function Eo(e) {
	let t = (e ?? "note").trim().toLowerCase();
	return t === "info" || t === "tip" || t === "warning" || t === "danger" || t === "note" ? t : "note";
}
var Do = {
	info: "note",
	tip: "tip",
	warning: "warning",
	danger: "caution",
	note: "default"
};
function Oo(e) {
	let t = xe(Do[e] ?? "default");
	return {
		container: t.subtle,
		title: t.title,
		body: t.content
	};
}
function ko({ node: e }) {
	let t = No(e.id);
	return t ? /* @__PURE__ */ (0, L.jsx)("sup", {
		id: Po("fnref", t),
		className: "align-super text-[0.75em]",
		children: /* @__PURE__ */ (0, L.jsx)("a", {
			href: `#${Po("fn", t)}`,
			"aria-label": `Footnote ${t}`,
			className: "text-primary hover:underline",
			children: `[^${t}]`
		})
	}) : null;
}
function Ao({ node: e }) {
	let t = No(e.id);
	return t ? /* @__PURE__ */ (0, L.jsxs)("div", {
		id: Po("fn", t),
		className: "text-sm text-muted-foreground",
		children: [
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: "mr-2 font-mono text-xs text-foreground",
				children: `[^${t}]`
			}),
			e.content && /* @__PURE__ */ (0, L.jsx)(z, { node: e.content }),
			/* @__PURE__ */ (0, L.jsx)(Mo, { id: t })
		]
	}) : null;
}
function jo({ node: e }) {
	let t = (e.items ?? []).filter((e) => No(e.id));
	return t.length === 0 ? null : /* @__PURE__ */ (0, L.jsx)("section", {
		"aria-label": "Footnotes",
		className: "border-t border-border pt-density-2 text-xs text-muted-foreground",
		children: /* @__PURE__ */ (0, L.jsx)("ol", {
			className: "ml-5 list-decimal space-y-1",
			children: t.map((e, t) => /* @__PURE__ */ (0, L.jsxs)("li", {
				id: Po("fn", e.id ?? ""),
				children: [e.content && /* @__PURE__ */ (0, L.jsx)(z, { node: e.content }), /* @__PURE__ */ (0, L.jsx)(Mo, { id: e.id ?? "" })]
			}, `${e.id}-${t}`))
		})
	});
}
function Mo({ id: e }) {
	return /* @__PURE__ */ (0, L.jsx)("a", {
		href: `#${Po("fnref", e)}`,
		"aria-label": `Back to footnote reference ${e}`,
		className: "ml-2 text-primary hover:underline",
		children: "back"
	});
}
function No(e) {
	return (e ?? "").trim();
}
function Po(e, t) {
	return `${e}-${encodeURIComponent(t)}`;
}
function Fo(e) {
	let t = e.text ?? e.plain;
	return t ? {
		kind: "text",
		text: t,
		plain: t
	} : void 0;
}
function Io({ node: e }) {
	let t = e.items ?? [];
	if (t.length === 0) return null;
	if (e.inline) return /* @__PURE__ */ (0, L.jsx)("span", {
		className: "inline-flex flex-wrap items-start gap-1",
		children: t.map((t, n) => /* @__PURE__ */ (0, L.jsxs)(I.Fragment, { children: [n > 0 && /* @__PURE__ */ (0, L.jsx)("span", {
			className: "text-muted-foreground",
			children: ","
		}), /* @__PURE__ */ (0, L.jsxs)("span", {
			className: "inline-flex items-center gap-1",
			children: [
				!e.ordered && e.bullet && /* @__PURE__ */ (0, L.jsx)(z, { node: e.bullet }),
				e.ordered && /* @__PURE__ */ (0, L.jsxs)("span", {
					className: "text-muted-foreground",
					children: [n + 1, "."]
				}),
				/* @__PURE__ */ (0, L.jsx)(z, { node: t })
			]
		})] }, n))
	});
	if (e.unstyled) return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "space-y-1",
		children: t.map((t, n) => /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex items-start gap-2",
			children: [
				!e.ordered && e.bullet && /* @__PURE__ */ (0, L.jsx)(z, { node: e.bullet }),
				e.ordered && /* @__PURE__ */ (0, L.jsxs)("span", {
					className: "text-muted-foreground",
					children: [n + 1, "."]
				}),
				/* @__PURE__ */ (0, L.jsx)("div", {
					className: "min-w-0 flex-1",
					children: /* @__PURE__ */ (0, L.jsx)(z, { node: t })
				})
			]
		}, n))
	});
	let n = e.ordered ? "ol" : "ul", r = e.ordered ? "list-decimal" : "list-disc";
	return /* @__PURE__ */ (0, L.jsx)(n, {
		className: P("ml-5 space-y-1 text-sm", r),
		children: t.map((e, t) => /* @__PURE__ */ (0, L.jsx)("li", { children: /* @__PURE__ */ (0, L.jsx)(z, { node: e }) }, t))
	});
}
function Lo({ node: e }) {
	let t = (0, I.useContext)(pa), n = (e.fields ?? []).filter((e) => !(t.tablePagination && Ro(e)));
	if (n.length === 0) return null;
	let r = n.filter((e) => zs(e.value)), i = n.filter((e) => !zs(e.value));
	return r.length === 0 && i.length === 1 && i[0]?.value.kind === "table" ? /* @__PURE__ */ (0, L.jsx)(z, { node: i[0].value }) : /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "space-y-density-4",
		children: [r.length > 0 && /* @__PURE__ */ (0, L.jsx)("dl", {
			className: "grid gap-density-3 md:grid-cols-2",
			children: r.map((e) => /* @__PURE__ */ (0, L.jsxs)("div", { children: [/* @__PURE__ */ (0, L.jsx)("dt", {
				className: "text-xs font-semibold uppercase tracking-wide text-muted-foreground",
				children: e.label || Cs(e.name)
			}), /* @__PURE__ */ (0, L.jsx)("dd", {
				className: "mt-1 text-sm text-foreground",
				children: /* @__PURE__ */ (0, L.jsx)(z, { node: e.value })
			})] }, e.name))
		}), i.map((e) => /* @__PURE__ */ (0, L.jsxs)("section", {
			className: "space-y-2",
			children: [/* @__PURE__ */ (0, L.jsx)("div", {
				className: "text-xs font-semibold uppercase tracking-wide text-muted-foreground",
				children: e.label || Cs(e.name)
			}), /* @__PURE__ */ (0, L.jsx)("div", {
				className: "rounded-md border border-border bg-background p-density-3",
				children: /* @__PURE__ */ (0, L.jsx)(z, { node: e.value })
			})]
		}, e.name))]
	});
}
function Ro(e) {
	if ((e.label || e.name || "").toLowerCase() !== "page" || e.value.kind !== "map") return !1;
	let t = new Set((e.value.fields ?? []).map((e) => e.name.toLowerCase()));
	return t.has("limit") && t.has("offset") && t.has("total");
}
function zo({ node: e }) {
	return /* @__PURE__ */ (0, L.jsx)(Bo, {
		columns: e.columns ?? [],
		rows: e.rows ?? [],
		...e.autoFilter === void 0 ? {} : { autoFilter: e.autoFilter }
	});
}
function Bo({ columns: e, rows: t, autoFilter: n, onTableRowClick: r, getTableRowHref: i, isTableRowClickable: a, search: s, timeRange: c, externalFilters: l, cellFilters: u, onCellFilterChange: d, pagination: f, sort: p, onSortChange: m, infinite: h, rowSelection: g, hiddenColumns: _, menuActions: y, loading: b, renderRowDetail: x, cellRenderers: S, renderRowCard: C, detailStyle: w, detailDialogSize: T, detailDialogTitle: E }) {
	let D = (0, I.useContext)(pa), O = r ?? D.onTableRowClick, ee = i ?? D.getTableRowHref, te = a ?? D.isTableRowClickable, ne = s ?? D.tableSearch, re = c ?? D.tableTimeRange, ie = l ?? D.tableExternalFilters, k = u ?? D.tableCellFilters, A = d ?? D.onTableCellFilterChange, j = f ?? D.tablePagination, ae = p === void 0 ? D.tableSort : p, oe = m ?? D.onTableSortChange, M = h ?? D.tableInfinite, se = g ?? D.tableRowSelection, ce = _ ?? D.tableHiddenColumns, N = y ?? D.tableMenuActions, le = b ?? D.tableLoading, ue = x ?? D.tableRenderRowDetail, de = S ?? D.tableCellRenderers, fe = C ?? D.tableRenderRowCard, pe = w ?? D.tableDetailStyle, me = T ?? D.tableDetailDialogSize, he = E ?? D.tableDetailDialogTitle, ge = new Set(ce), _e = e.filter((e) => !ge.has(e.name));
	if (_e.length === 0) return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "text-sm text-muted-foreground",
		children: "No data"
	});
	if (Uo(_e, t)) return /* @__PURE__ */ (0, L.jsx)(Vo, {
		columns: _e,
		rows: t
	});
	let ve = new Map(e.map((e) => [e.name, e])), ye = _e.map((e) => {
		let t = Wo(e), n = Go(e), r = e.type === "json", i = {
			key: `cells.${e.name}`,
			accessor: (t) => t.cells[e.name],
			label: e.header ? /* @__PURE__ */ (0, L.jsx)(z, { node: e.header }) : e.label || Cs(e.name),
			...e.align ? { align: e.align } : {},
			...oe ? { sortable: !!e.sortKey } : e.sortable === void 0 ? {} : { sortable: e.sortable },
			...e.filterable === void 0 ? {} : { filterable: e.filterable },
			...e.filterKey ? { filterKey: e.filterKey } : {},
			serverFilterValue: (e) => {
				let t = e;
				return t.filterValue ?? U(t);
			},
			serverFilterLabel: (e) => U(e),
			...e.grow === void 0 ? {} : { grow: e.grow },
			...e.shrink === void 0 ? {} : { shrink: e.shrink },
			...e.minWidth === void 0 ? {} : { minWidth: e.minWidth },
			...e.maxWidth === void 0 ? {} : { maxWidth: e.maxWidth }
		};
		return e.kind === "timestamp" && !t && !n && !r ? {
			...i,
			kind: "timestamp",
			accessor: (t) => Ls(t.cells[e.name], e),
			serverFilterValue: (e) => typeof e == "string" || typeof e == "number" ? String(e) : void 0,
			serverFilterLabel: (t, n) => typeof t == "string" || typeof t == "number" ? String(t) : U(n.cells[e.name])
		} : n ? {
			...i,
			kind: "tags",
			render: (e) => /* @__PURE__ */ (0, L.jsx)(Ko, { node: e }),
			sortValue: (e) => $o(e),
			filterValue: (e) => Qo(e)
		} : r ? {
			...i,
			render: (e) => /* @__PURE__ */ (0, L.jsx)(qo, { node: e }),
			sortValue: (e) => ns(e),
			filterValue: (e) => ns(e)
		} : t ? {
			...i,
			kind: "tags",
			render: (t) => {
				let n = rs(t, e);
				return /* @__PURE__ */ (0, L.jsx)(v, {
					tags: n,
					maxVisible: 3
				});
			},
			sortValue: (t) => rs(t, e).length,
			filterValue: (t) => rs(t, e).map((e) => e.token)
		} : {
			...i,
			...e.kind ? { kind: e.kind } : {},
			render: (e) => {
				let t = e;
				return t.kind === "code" ? /* @__PURE__ */ (0, L.jsx)(Ho, { node: t }) : /* @__PURE__ */ (0, L.jsx)(z, { node: t });
			},
			sortValue: (e) => Fs(e),
			filterValue: (e) => U(e)
		};
	}).map((e, t) => {
		let n = de?.[_e[t].name];
		return n ? {
			...e,
			render: (e, r) => n(Ls(e, _e[t]), Rs(r, ve))
		} : e;
	}), P = j || oe ? void 0 : _e.find((e) => e.sortable !== !1) ?? _e[0], be = ae ? (() => {
		let e = _e.find((e) => e.sortKey === ae.key);
		if (!e) throw Error(`Server sort key ${ae.key} does not match a Clicky column`);
		return {
			key: `cells.${e.name}`,
			dir: ae.dir
		};
	})() : ae === null ? null : void 0, xe = oe ? (e) => {
		if (!e) {
			oe(null);
			return;
		}
		let t = e.key.startsWith("cells.") ? e.key.slice(6) : e.key, n = _e.find((e) => e.name === t);
		if (!n?.sortKey) throw Error(`Clicky column ${t} does not declare a server sort key`);
		oe({
			key: n.sortKey,
			dir: e.dir
		});
	} : void 0;
	return /* @__PURE__ */ (0, L.jsx)(o, {
		className: "min-h-0 flex-1",
		data: t,
		columns: ye,
		defaultHiddenColumns: _e.filter((e) => e.defaultHidden).map((e) => `cells.${e.name}`),
		...n === void 0 ? {} : { autoFilter: n },
		...P ? { defaultSort: {
			key: `cells.${P.name}`,
			dir: "asc"
		} } : {},
		...be === void 0 ? {} : { sort: be },
		...xe ? {
			onSortChange: xe,
			manualSort: !0
		} : {},
		getRowId: (e, t) => se?.getRowId(e, t) ?? (fe && typeof Rs(e, ve).id == "string" ? String(Rs(e, ve).id) : void 0) ?? `${t}-${_e.map((t) => U(e.cells[t.name])).filter(Boolean).join("|")}`,
		...se ? { rowSelection: se } : {},
		...O ? { onRowClick: O } : {},
		...ee ? { getRowHref: ee } : {},
		...O && (te || ee) ? { isRowClickable: te ?? ((e) => !!ee?.(e)) } : {},
		...ne ? { externalSearch: ne } : {},
		...re ? { externalTimeRange: re } : {},
		...ie ? { externalFilters: ie } : {},
		...k ? { cellFilters: k } : {},
		...A ? { onCellFilterChange: A } : {},
		...j ? { pagination: j } : {},
		...M ? { infinite: M } : {},
		...N ? { menuActions: N } : {},
		...le === void 0 ? {} : { loading: le },
		...ue ? { renderExpandedRow: (e) => ue(Rs(e, ve)) } : {},
		...fe ? { renderCard: (e) => fe(Rs(e, ve)) } : {},
		...pe ? { detailStyle: pe } : {},
		...me ? { detailDialogSize: me } : {},
		...he ? { detailDialogTitle: (e) => he(Rs(e, ve)) } : {}
	});
}
function Vo({ columns: e, rows: t }) {
	return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "space-y-2",
		children: t.map((t, n) => {
			let r = ss(t, e, n), i = os(t, e);
			return /* @__PURE__ */ (0, L.jsxs)("details", {
				className: "rounded-md border border-border bg-background",
				children: [/* @__PURE__ */ (0, L.jsx)("summary", {
					className: "cursor-pointer px-3 py-2 text-sm font-medium text-foreground",
					children: r
				}), /* @__PURE__ */ (0, L.jsx)("div", {
					className: "border-t border-border p-3",
					children: /* @__PURE__ */ (0, L.jsx)(Lo, { node: i })
				})]
			}, cs(t, e, n));
		})
	});
}
function Ho({ node: e }) {
	return /* @__PURE__ */ (0, L.jsx)(Ae, {
		language: e.language,
		source: e.source ?? e.plain ?? "",
		highlightedHtml: e.highlightedHtml,
		inline: !0
	});
}
function Uo(e, t) {
	return t.length === 0 || e.some((e) => Go(e) || e.type === "json") ? !1 : t.every((t) => {
		let n = e.map((e) => t.cells[e.name]).filter(Boolean);
		if (n.length === 0) return !1;
		let r = n.filter(as);
		return r.length > 0 && r.length === n.length;
	});
}
function Wo(e) {
	if (e.kind === "tags") return !0;
	let t = e.name.toLowerCase();
	return /(^|[._-])(tags?|labels?|annotations?)$/.test(t);
}
function Go(e) {
	return e.type === "key_value" || e.type === "key_values";
}
function Ko({ node: e }) {
	let t = Jo(e);
	if (t.length === 0) {
		let t = U(e).trim();
		return t ? /* @__PURE__ */ (0, L.jsx)("span", {
			className: "text-sm text-foreground",
			children: t
		}) : /* @__PURE__ */ (0, L.jsx)("span", {
			className: "text-muted-foreground",
			children: "—"
		});
	}
	return /* @__PURE__ */ (0, L.jsx)("dl", {
		className: "grid grid-cols-[max-content_minmax(0,1fr)] gap-x-2 gap-y-1 text-xs",
		children: t.map((e, t) => /* @__PURE__ */ (0, L.jsxs)(I.Fragment, { children: [/* @__PURE__ */ (0, L.jsx)("dt", {
			className: "font-medium text-muted-foreground",
			children: e.key
		}), /* @__PURE__ */ (0, L.jsx)("dd", {
			className: "min-w-0 break-words text-foreground",
			children: /* @__PURE__ */ (0, L.jsx)(z, { node: e.value })
		})] }, `${e.key}-${t}`))
	});
}
function qo({ node: e }) {
	let t = ts(e);
	return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "max-h-40 min-w-40 max-w-xl overflow-auto text-xs",
		onClick: (e) => e.stopPropagation(),
		children: /* @__PURE__ */ (0, L.jsx)(je, {
			data: t,
			defaultOpenDepth: 0
		})
	});
}
function Jo(e) {
	if (!e) return [];
	if (e.kind === "map") return Yo(e.fields ?? []);
	if (e.kind === "list") return (e.items ?? []).flatMap((e) => Jo(e));
	let t = e.kind === "code" ? e.source : e.text ?? e.plain;
	if (!t) return [];
	try {
		return Xo(JSON.parse(t));
	} catch {
		return [];
	}
}
function Yo(e) {
	let t = e.find((e) => /^(key|name)$/i.test(e.name)), n = e.find((e) => /^value$/i.test(e.name));
	if (t && n) {
		let e = U(t.value).trim();
		if (e) return [{
			key: e,
			value: n.value
		}];
	}
	return e.map((e) => ({
		key: e.label || e.name,
		value: e.value
	}));
}
function Xo(e) {
	if (Array.isArray(e)) return e.flatMap(Xo);
	if (!e || typeof e != "object") return [];
	let t = e, n = Object.keys(t).find((e) => /^(key|name)$/i.test(e)), r = Object.keys(t).find((e) => /^value$/i.test(e));
	return n && r && typeof t[n] == "string" ? [{
		key: t[n],
		value: Zo(t[r])
	}] : Object.keys(t).sort((e, t) => e.localeCompare(t)).map((e) => ({
		key: e,
		value: Zo(t[e])
	}));
}
function Zo(e) {
	if (Array.isArray(e)) return {
		kind: "list",
		plain: JSON.stringify(e),
		items: e.map(Zo)
	};
	if (e && typeof e == "object") {
		let t = e;
		return {
			kind: "map",
			plain: JSON.stringify(e),
			fields: Object.keys(t).sort((e, t) => e.localeCompare(t)).map((e) => ({
				name: e,
				value: Zo(t[e])
			}))
		};
	}
	let t = e === null ? "null" : String(e ?? "");
	return {
		kind: "text",
		text: t,
		plain: t
	};
}
function Qo(e) {
	return Jo(e).map((e) => `${e.key}=${es(e.value)}`).filter((e) => !e.endsWith("="));
}
function $o(e) {
	return Qo(e).join(", ");
}
function es(e) {
	let t = U(e).trim();
	if (t) return t;
	let n = ts(e);
	return n === void 0 ? "" : JSON.stringify(n);
}
function ts(e) {
	if (!e) return;
	if (e.kind === "map") return Object.fromEntries((e.fields ?? []).map((e) => [e.name, ts(e.value)]));
	if (e.kind === "list") return (e.items ?? []).map(ts);
	let t = e.kind === "code" ? e.source ?? e.plain : e.text ?? e.plain;
	if (t === void 0) return U(e);
	try {
		return JSON.parse(t);
	} catch {
		return t;
	}
}
function ns(e) {
	let t = ts(e);
	return t === void 0 ? "" : JSON.stringify(t);
}
function rs(e, t) {
	if (!Wo(t)) return [];
	let n = is(e);
	return s(n);
}
function is(e) {
	if (!e) return [];
	if (e.kind === "map" && e.fields?.length) return Object.fromEntries(e.fields.map((e) => [e.name, U(e.value)]).filter(([, e]) => e.trim() !== ""));
	if (e.kind === "list" && e.items?.length) return e.items.map((e) => U(e)).filter(Boolean);
	let t = U(e).trim();
	return t ? t.split(/[,\s]+/).map((e) => e.trim()).filter(Boolean) : [];
}
function as(e) {
	if (!e) return !1;
	if (e.kind === "map") return (e.fields ?? []).length > 0;
	if (e.kind === "table") return (e.rows ?? []).length > 0;
	if (e.kind !== "list") return !1;
	let t = e.items ?? [];
	return t.length > 0 && t.every(as);
}
function os(e, t) {
	return {
		kind: "map",
		fields: t.flatMap((t) => {
			let n = e.cells[t.name];
			return n ? [{
				name: t.name,
				label: U(t.header) || t.label || Cs(t.name),
				value: n
			}] : [];
		})
	};
}
function ss(e, t, n) {
	let r = ps(t.find((t) => ds(e.cells[t.name]))?.label ?? t[0]?.label ?? t[0]?.name ?? "Item"), i = ls(e, t);
	return i ? `${r} ${n + 1}: ${i}` : `${r} ${n + 1}`;
}
function cs(e, t, n) {
	return `${n}-${ls(e, t) || t.map((t) => U(e.cells[t.name])).join("|")}`;
}
function ls(e, t) {
	for (let n of [
		"name",
		"title",
		"agreementname",
		"companyname",
		"customernumber",
		"id",
		"guid"
	]) for (let r of t) {
		let t = us(e.cells[r.name], n);
		if (t) return t;
	}
	for (let n of t) {
		let t = U(e.cells[n.name]).trim();
		if (t) return t;
	}
	return "";
}
function us(e, t) {
	if (!e) return "";
	if (e.kind === "map") {
		for (let n of e.fields ?? []) if (fs(n.name) === t) {
			let e = U(n.value).trim();
			if (e) return e;
		}
	}
	for (let n of e.fields ?? []) {
		let e = us(n.value, t);
		if (e) return e;
	}
	for (let n of e.items ?? []) {
		let e = us(n, t);
		if (e) return e;
	}
	return "";
}
function ds(e) {
	return !!(U(e).trim() || (e?.fields ?? []).length > 0);
}
function fs(e) {
	return e.toLowerCase().replace(/[^a-z0-9]/g, "");
}
function ps(e) {
	let t = Cs(e).trim();
	return t.endsWith("s") ? t.slice(0, -1) : t;
}
function ms({ node: e }) {
	let t = e.roots ?? [];
	return t.length === 0 ? null : /* @__PURE__ */ (0, L.jsx)($e, {
		roots: t,
		getChildren: (e) => e.children,
		getKey: (e) => e.id,
		defaultOpen: (e, t) => t < 1,
		renderRow: ({ node: e }) => /* @__PURE__ */ (0, L.jsx)("div", {
			className: "min-w-0 flex-1 truncate",
			children: /* @__PURE__ */ (0, L.jsx)(z, { node: e.label })
		})
	});
}
function hs(e) {
	return e === "pill" || e === "rounded" || e === "square";
}
function gs({ node: e }) {
	let t = { variant: "label" };
	return e.badgeLabel !== void 0 && (t.label = e.badgeLabel), e.badgeValue !== void 0 && (t.value = e.badgeValue), e.badgeColor !== void 0 && (t.color = e.badgeColor), e.badgeText !== void 0 && (t.textColor = e.badgeText), e.badgeIcon !== void 0 && (t.icon = e.badgeIcon), hs(e.badgeShape) && (t.shape = e.badgeShape), /* @__PURE__ */ (0, L.jsx)(Oe, { ...t });
}
function _s({ node: e }) {
	let t = {
		causedBy: e.causedBy ?? [],
		frames: e.frames ?? [],
		language: "java",
		...e.exceptionClass === void 0 ? {} : { exceptionClass: e.exceptionClass },
		...e.message === void 0 ? {} : { message: e.message }
	};
	return /* @__PURE__ */ (0, L.jsx)(Zi, {
		input: t,
		className: "space-y-1"
	});
}
function vs({ node: e }) {
	return /* @__PURE__ */ (0, L.jsxs)("details", {
		className: "rounded-md border border-border bg-background",
		children: [/* @__PURE__ */ (0, L.jsx)("summary", {
			className: "cursor-pointer px-3 py-2 text-sm font-medium text-foreground",
			children: /* @__PURE__ */ (0, L.jsx)(z, { node: e.label })
		}), e.content && /* @__PURE__ */ (0, L.jsx)("div", {
			className: "border-t border-border px-3 py-3",
			children: /* @__PURE__ */ (0, L.jsx)(z, { node: e.content })
		})]
	});
}
function ys({ node: e }) {
	let t = [e.id, e.payload].filter(Boolean).join("\n") || void 0, n = e.label ? /* @__PURE__ */ (0, L.jsx)(z, { node: e.label }) : e.text;
	return e.href ? /* @__PURE__ */ (0, L.jsx)("a", {
		href: e.href,
		target: "_blank",
		rel: "noopener noreferrer",
		title: t,
		className: "inline-flex items-center rounded-md border border-border bg-background px-3 py-1.5 text-sm font-medium text-foreground hover:bg-accent",
		children: n
	}) : /* @__PURE__ */ (0, L.jsx)("button", {
		type: "button",
		title: t,
		className: "inline-flex items-center rounded-md border border-border bg-muted px-3 py-1.5 text-sm font-medium text-foreground",
		children: n
	});
}
function bs({ node: e }) {
	let t = e.items ?? [];
	return t.length === 0 ? null : /* @__PURE__ */ (0, L.jsx)("div", {
		className: "flex flex-wrap gap-2",
		children: t.map((e, t) => /* @__PURE__ */ (0, L.jsx)(z, { node: e }, t))
	});
}
function xs({ node: e }) {
	let t = ke(e.html ?? "");
	if (!t) return null;
	let n = Ee(t) ? "div" : "span";
	return /* @__PURE__ */ (0, L.jsx)(n, { dangerouslySetInnerHTML: { __html: t } });
}
function Ss(e, t) {
	if (!e) return t?.includes("\n") ? { whiteSpace: "pre-wrap" } : void 0;
	let n = {};
	e.color && (n.color = e.color), e.backgroundColor && (n.backgroundColor = e.backgroundColor), e.bold && (n.fontWeight = 700), e.faint && (n.opacity = .7), e.italic && (n.fontStyle = "italic");
	let r = [];
	return e.underline && r.push("underline"), e.strikethrough && r.push("line-through"), r.length > 0 && (n.textDecoration = r.join(" ")), (e.textTransform === "uppercase" || e.textTransform === "lowercase" || e.textTransform === "capitalize") && (n.textTransform = e.textTransform), e.maxWidth && e.maxWidth > 0 && (n.maxWidth = `${e.maxWidth}ch`), e.maxLines && e.maxLines > 0 && (n.display = "-webkit-box", n.overflow = "hidden", n.WebkitLineClamp = e.maxLines, n.WebkitBoxOrient = "vertical", n.whiteSpace = "pre-wrap"), e.truncateMode === "suffix" && e.maxWidth && (!e.maxLines || e.maxLines <= 1) && (n.overflow = "hidden", n.textOverflow = "ellipsis", n.whiteSpace = "nowrap", n.display = "inline-block", n.verticalAlign = "bottom"), e.monospace && (n.fontFamily = "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, \"Liberation Mono\", \"Courier New\", monospace"), !n.whiteSpace && t?.includes("\n") && (n.whiteSpace = "pre-wrap"), Object.keys(n).length > 0 ? n : void 0;
}
function Cs(e) {
	return e.replace(/([a-z])([A-Z])/g, "$1 $2").replace(/_/g, " ").replace(/\b\w/g, (e) => e.toUpperCase());
}
function ws(e) {
	return {
		command: e.command ?? "",
		args: e.args ?? [],
		flags: e.flags ?? {},
		target: e.target ?? "Dialog",
		autoRun: e.autoRun ?? !1
	};
}
function Ts(e, t, n) {
	return {
		request: e,
		operation: t?.resolveCommand?.(e, n) ?? Es(e, n)
	};
}
function Es(e, t) {
	let n = Ds(e.command);
	if (n) return t.find((e) => {
		let t = Ds(e.operation["x-clicky"]?.command);
		return t && t === n ? !0 : Ds(e.operation.operationId?.replaceAll("_", "/")) === n;
	});
}
function Ds(e) {
	return (e ?? "").trim().replace(/^\//, "").replace(/\s+/g, "/").replace(/\/+/g, "/").toLowerCase();
}
function Os(e, t) {
	let n = { ...t.flags }, r = [...t.args ?? []], i = e.filter((e) => Cr(e)), a = 0;
	for (let e of i) if (e.name === "args") {
		let e = r.slice(a);
		e.length > 0 && (n.args = e.join(",")), a = r.length;
	} else a < r.length && (n[e.name] = r[a] ?? "", a += 1);
	return n;
}
function ks(e, t) {
	return e.filter((e) => e.required && (t[e.name] ?? "").trim() === "");
}
async function As(e, t, n) {
	return e.executeCommand(t.path, t.method, Ar(kr(n), t.operation.parameters ?? []), { Accept: "application/json+clicky" });
}
function js(e) {
	let t = e.operation;
	if (!t) return;
	let n = Ar(kr(Os(t.operation.parameters ?? [], e.request)), t.operation.parameters ?? []), r = Ms(t.path, n);
	return t.method.toUpperCase(), r;
}
function Ms(e, t) {
	let n = new URL(e, "http://localhost"), r = { ...t };
	for (let [e, i] of Object.entries(t)) if (n.pathname.includes(`{${e}}`)) {
		if (Array.isArray(i)) throw Error(`Path parameter ${e} must be a scalar value`);
		n.pathname = n.pathname.replace(`{${e}}`, encodeURIComponent(i)), delete r[e];
	}
	for (let [e, t] of Object.entries(r)) n.searchParams.set(e, Array.isArray(t) ? t.join(",") : t);
	return `${n.pathname}${n.search}`;
}
function Ns(e) {
	if (e === "_self") return "_self";
	if (e === "_window" || e === "_tab") return "_blank";
}
function Ps(e) {
	return typeof e == "object" && !!e && ("stdout" in e || "output" in e || "message" in e);
}
function Fs(e) {
	return e?.filterValue ?? U(e);
}
var Is = /* @__PURE__ */ new Set([
	"number",
	"int",
	"integer",
	"float",
	"decimal",
	"currency",
	"bytes",
	"duration",
	"boolean",
	"bool",
	"date",
	"datetime",
	"time",
	"timestamp"
]);
function Ls(e, t) {
	if (e == null) return;
	if (e.kind === "execution-tree") return e.executionRoots ?? [];
	if (e.kind === "object-graph") return e.objects ?? [];
	if (e.kind === "map" || e.kind === "list") return ts(e);
	if (e.kind === "code" && e.language === "json") return JSON.parse(e.source ?? "null");
	if (e.filterValue !== void 0) return e.filterValue;
	let n = U(e);
	return n === "" && t?.type !== void 0 && Is.has(t.type) ? null : n;
}
function Rs(e, t) {
	return Object.fromEntries(Object.entries(e.cells).map(([e, n]) => [e, Ls(n, t.get(e))]));
}
function U(e) {
	if (!e) return "";
	if (e.kind === "heading" || e.kind === "blockquote") return U(e.content) || e.plain || "";
	if (e.kind === "admonition") return [U(e.label), U(e.content)].filter(Boolean).join(" ") || e.plain || "";
	if (e.kind === "footnote-ref") return e.id ? `[^${e.id}]` : e.plain || "";
	if (e.kind === "footnote") return U(e.content) || e.plain || "";
	if (e.kind === "footnotes") return (e.items ?? []).map((e) => U(e)).join(" ") || e.plain || "";
	if (e.plain) return e.plain;
	if (e.text) return e.text;
	if (e.kind === "html") return D(ke(e.html ?? "")).trim();
	if (e.kind === "code") return e.source ?? "";
	if ((e.kind === "link" || e.kind === "link-command") && e.children?.length) return [e.text, ...e.children.map((e) => U(e))].join("");
	if (e.kind === "button" && e.label) return U(e.label);
	if (e.kind === "button-group") return (e.items ?? []).map((e) => U(e)).join(" ");
	if (e.kind === "text" && e.children?.length) return [e.text, ...e.children.map((e) => U(e))].join("");
	if (e.kind === "list" && e.items?.length) return e.items.map((e) => U(e)).join(" ");
	if (e.kind === "map" && e.fields?.length) return e.fields.map((e) => {
		let t = U(e.value), n = e.label ?? e.name;
		return n ? `${n}: ${t}` : t;
	}).join(" ");
	if (e.kind === "tree" && e.roots?.length) {
		let t = (e) => [U(e.label), ...(e.children ?? []).map((e) => t(e))].filter(Boolean).join(" ");
		return e.roots.map((e) => t(e)).join(" ");
	}
	return e.kind === "table" && e.rows?.length ? e.rows.map((e) => Object.values(e.cells ?? {}).map((e) => U(e)).filter(Boolean).join(" ")).filter(Boolean).join(" ") : "";
}
function zs(e) {
	return e.kind === "text" || e.kind === "link" || e.kind === "link-command" || e.kind === "icon" || e.kind === "html" || e.kind === "footnote-ref" || e.kind === "badge" || e.kind === "button" || e.kind === "button-group";
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/actions.ts
var Bs = "Add {source} to .gavel.yaml";
function Vs(e) {
	if (!e) return "**";
	let t = e.replace(/\/+$/, "");
	return t ? `${t}/**` : "**";
}
function Hs(e, t) {
	let n = e.target_path ?? "", r = /* @__PURE__ */ new Map();
	for (let i of t) if (!(e.work_dir && i.work_dir && i.work_dir !== e.work_dir)) for (let e of i.violations ?? []) {
		let t = dt(e.file, i.work_dir);
		if (!t || n !== "" && t !== n && !t.startsWith(`${n}/`)) continue;
		let a = r.get(i.linter);
		a ? a.count += 1 : r.set(i.linter, {
			linter: i.linter,
			count: 1,
			workDir: i.work_dir
		});
	}
	return Array.from(r.values()).sort((e, t) => t.count - e.count || e.linter.localeCompare(t.linter));
}
function Us(e, t) {
	return {
		key: "linter-disable",
		label: `Disable ${e} entirely`,
		title: Bs,
		req: {
			source: e,
			work_dir: t
		},
		variant: "subtle"
	};
}
function Ws(e, t) {
	let n = Vs(e.target_path), r = e.linterName ?? "", i = e.ruleName ?? "", a = r ? [] : Hs(e, t), o = new Set(a.flatMap((e) => e.workDir ? [e.workDir] : [])), s = e.work_dir || (o.size === 1 ? Array.from(o)[0] : "");
	if (r && i) return [
		{
			key: "rule-folder",
			label: `Ignore ${i} in this folder`,
			title: "Add {source, rule, file} to .gavel.yaml",
			req: {
				source: r,
				rule: i,
				file: n,
				work_dir: s
			},
			disabledWithoutWorkDir: !0
		},
		{
			key: "rule-repo",
			label: `Ignore rule ${i} everywhere`,
			title: "Add {source, rule} to .gavel.yaml",
			req: {
				source: r,
				rule: i,
				work_dir: s
			}
		},
		Us(r, s)
	];
	let c = {
		key: "folder-everything",
		label: "Ignore everything in this folder",
		title: "Add {file} to .gavel.yaml",
		req: {
			file: n,
			work_dir: s
		},
		disabledWithoutWorkDir: !0
	};
	return r ? [
		c,
		{
			key: "linter-folder",
			label: `Ignore ${r} in this folder`,
			title: "Add {source, file} to .gavel.yaml",
			req: {
				source: r,
				file: n,
				work_dir: s
			},
			disabledWithoutWorkDir: !0
		},
		Us(r, s)
	] : [c, ...a.flatMap((e) => [{
		key: `linter-folder:${e.linter}`,
		label: `Ignore ${e.linter} in this folder`,
		title: "Add {source, file} to .gavel.yaml",
		req: {
			source: e.linter,
			file: n,
			work_dir: s
		},
		variant: "subtle",
		disabledWithoutWorkDir: !0
	}, {
		...Us(e.linter, s),
		key: `linter-disable:${e.linter}`
	}])];
}
function Gs(e) {
	let t = (e.children ?? []).flatMap((e) => e.kind === "linter" && e.linterName ? [e.linterName] : []);
	return Array.from(new Set(t));
}
function Ks(e) {
	let t = e.linterName ?? "", n = e.target_path || e.file || "", r = e.work_dir ?? "";
	return t ? [{
		key: "file-everything-linter",
		label: `Ignore all ${t} in this file`,
		title: "Add {source, file} to .gavel.yaml",
		req: {
			source: t,
			file: n,
			work_dir: r
		}
	}, Us(t, r)] : [{
		key: "file-everything",
		label: "Ignore everything in this file",
		title: "Add {file} to .gavel.yaml",
		req: {
			file: n,
			work_dir: r
		}
	}, ...Gs(e).flatMap((e) => [{
		key: `file-linter:${e}`,
		label: `Ignore all ${e} in this file`,
		title: "Add {source, file} to .gavel.yaml",
		req: {
			source: e,
			file: n,
			work_dir: r
		},
		variant: "subtle"
	}, {
		...Us(e, r),
		key: `linter-disable:${e}`
	}])];
}
function qs(e) {
	let t = e.linterName ?? "", n = e.ruleName ?? "", r = e.work_dir ?? "", i = e.target_path || e.file || "";
	return [
		...e.file ? [{
			key: "rule-file",
			label: `Ignore ${n} in this file`,
			title: "Add {source, rule, file} to .gavel.yaml",
			req: {
				source: t,
				rule: n,
				file: i,
				work_dir: r
			}
		}] : [],
		{
			key: "rule-repo",
			label: `Ignore rule ${n} everywhere`,
			title: "Add {source, rule} to .gavel.yaml",
			req: {
				source: t,
				rule: n,
				work_dir: r
			}
		},
		Us(t, r)
	];
}
function Js(e, t) {
	switch (e.kind) {
		case "linter": return e.linterName ? [Us(e.linterName, e.work_dir || e.linter?.work_dir || "")] : [];
		case "lint-folder": return Ws(e, t);
		case "lint-file": return Ks(e);
		case "lint-rule":
		case "lint-rule-group": return qs(e);
		case "linter-error": return [];
	}
}
function Ys({ violation: e, linter: t, file: n, workDir: r }) {
	let i = e.rule?.method, a = e.line;
	return [
		...n && a ? [{
			key: "violation-occurrence",
			label: "Ignore this occurrence",
			title: `Add {source${i ? ", rule" : ""}, file, line} to .gavel.yaml`,
			req: {
				source: t,
				...i ? { rule: i } : {},
				file: n,
				line: a,
				work_dir: r
			}
		}] : [],
		...i ? [{
			key: "violation-rule",
			label: `Ignore ${i} everywhere`,
			title: "Add {source, rule} to .gavel.yaml",
			variant: "subtle",
			req: {
				source: t,
				rule: i,
				work_dir: r
			}
		}] : [],
		...n ? [{
			key: "violation-file",
			label: `Ignore all ${t} in this file`,
			title: "Add {source, file} to .gavel.yaml",
			variant: "subtle",
			req: {
				source: t,
				file: n,
				work_dir: r
			}
		}] : []
	];
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/ignore-menu-items.ts
function Xs(e, t) {
	return t || !!e.disabledWithoutWorkDir && !e.req.work_dir;
}
function Zs(e, t, n) {
	return e.map((e) => ({
		label: e.label,
		icon: he,
		title: e.title,
		disabled: Xs(e, n),
		onSelect: () => void t(e)
	}));
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/violation-menu-items.ts
function Qs({ violation: e, linter: t, file: n, workDir: r, onIgnore: i, ignoreBusy: a, onOpenFile: o }) {
	let s = i ? Ys({
		violation: e,
		linter: t,
		file: n,
		workDir: r
	}) : [];
	return [
		{
			label: "Copy description",
			onSelect: () => {
				navigator.clipboard.writeText(e.message ?? "");
			}
		},
		...o && n ? [{
			label: "Open in Files",
			onSelect: () => o({
				file: n,
				...r ? { workDir: r } : {},
				...e.line ? { line: e.line } : {}
			})
		}] : [],
		...i ? Zs(s, i, !!a) : []
	];
}
var $s = /* @__PURE__ */ new WeakMap();
function ec(e, t, n) {
	let r = $s.get(e) ?? /* @__PURE__ */ new Map();
	$s.set(e, r);
	let i = `${n}\u0000${t}`, a = r.get(i);
	return a || (a = e({
		file: t,
		...n ? { workDir: n } : {}
	}), r.set(i, a), a.catch(() => r.delete(i))), a;
}
function tc(e, t, n) {
	let [r, i] = (0, I.useState)(null);
	return (0, I.useEffect)(() => {
		if (!e || !t) {
			i(null);
			return;
		}
		let r = !1;
		return i({ status: "loading" }), ec(e, t, n).then((e) => !r && i({
			status: "ready",
			text: e
		}), (e) => !r && i({
			status: "error",
			message: e instanceof Error ? e.message : String(e)
		})), () => {
			r = !0;
		};
	}, [
		e,
		t,
		n
	]), r;
}
function nc({ text: e }) {
	return e ? /* @__PURE__ */ (0, L.jsx)("div", {
		className: "mt-0.5 whitespace-pre-wrap text-sm",
		children: e
	}) : null;
}
function rc({ violation: e, file: t, workDir: n, loadSource: r, showMessage: i = !0 }) {
	let a = e.line ?? 0, o = tc(a > 0 ? r : void 0, t, n);
	if (!o || o.status !== "ready") return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		i && /* @__PURE__ */ (0, L.jsx)(nc, { text: e.message }),
		o?.status === "loading" && /* @__PURE__ */ (0, L.jsx)("div", {
			className: "mt-1 text-xs text-muted-foreground",
			children: "Loading source…"
		}),
		o?.status === "error" && /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "mt-1 text-xs text-red-700 dark:text-red-300",
			role: "alert",
			children: [
				"Could not load ",
				t,
				": ",
				o.message
			]
		}),
		!o && e.code && /* @__PURE__ */ (0, L.jsx)("pre", {
			className: "mt-1 overflow-x-auto rounded bg-muted p-2 font-mono text-xs",
			children: e.code
		})
	] });
	let s = o.text.split("\n"), c = Math.max(1, a - 3), l = Math.min(s.length, Math.max(a, e.end_line ?? a) + 3), u = Me(t);
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		i && /* @__PURE__ */ (0, L.jsx)(nc, { text: e.message }),
		a > s.length && /* @__PURE__ */ (0, L.jsx)("p", {
			role: "alert",
			className: "text-xs text-destructive",
			children: "Source location is outside this file."
		}),
		/* @__PURE__ */ (0, L.jsx)(k, {
			focalTone: Wt(e),
			className: "m-0 mt-1.5 max-sm:ml-0",
			frame: {
				sourceLines: s.slice(c - 1, l),
				sourceStartLine: c,
				line: a,
				...u ? { sourceLanguage: u } : {}
			}
		})
	] });
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/LintIgnoreMenu.tsx
function ic({ actions: e, onIgnore: t, busy: n, align: r = "left" }) {
	return /* @__PURE__ */ (0, L.jsx)(He, {
		label: "Actions",
		variant: "ghost",
		align: r,
		menuLabel: "Finding actions",
		items: Zs(e, t, n)
	});
}
function ac({ actions: e, onIgnore: t, busy: n, label: r }) {
	let [i, a] = (0, I.useState)(null), o = (0, I.useCallback)((e) => a(e?.parentElement ?? null), []);
	if (e.length === 0) return null;
	let s = Zs(e, t, n), c = `Ignore options for ${r}`;
	return /* @__PURE__ */ (0, L.jsxs)("span", {
		ref: o,
		className: "flex shrink-0 opacity-0 focus-within:opacity-100 [:hover>&]:opacity-100 [[aria-selected=true]>div>&]:opacity-100",
		onClick: (e) => e.stopPropagation(),
		onKeyDown: (e) => e.stopPropagation(),
		"data-lint-row-menu": "",
		children: [/* @__PURE__ */ (0, L.jsx)(He, {
			items: s,
			align: "right",
			menuLabel: c,
			trigger: /* @__PURE__ */ (0, L.jsx)(be, {
				type: "button",
				variant: "ghost",
				size: "icon",
				className: "h-5 w-5 text-muted-foreground",
				"aria-label": c,
				title: "Ignore options",
				children: /* @__PURE__ */ (0, L.jsx)(F, { icon: Ne })
			})
		}), /* @__PURE__ */ (0, L.jsx)(It, {
			contextTarget: i,
			menuLabel: c,
			menuItems: s
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/types.ts
function oc(e, t = /* @__PURE__ */ new Date()) {
	let n = e?.cache;
	if (n?.hit) return ze(new Date(n.stored_at), t);
}
var sc = [
	"linter-rule-file",
	"linter-file",
	"file-linter-rule",
	"summary"
];
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/LintDetail.tsx
function cc({ title: e, children: t }) {
	return /* @__PURE__ */ (0, L.jsxs)("section", {
		className: "space-y-1.5",
		children: [/* @__PURE__ */ (0, L.jsx)("h3", {
			className: "text-[10px] uppercase tracking-wide text-muted-foreground",
			children: e
		}), t]
	});
}
function lc({ children: e, className: t }) {
	return /* @__PURE__ */ (0, L.jsx)("span", {
		className: P("break-all font-mono text-sm", t),
		children: e
	});
}
function uc({ children: e, tone: t }) {
	return /* @__PURE__ */ (0, L.jsx)("pre", {
		className: P("max-h-80 overflow-auto whitespace-pre-wrap break-words rounded bg-muted p-3 font-mono text-xs", t === "error" && "bg-red-500/10 text-red-700 dark:text-red-300"),
		children: e
	});
}
function dc(e, t) {
	let n = `:${e.line ?? 0}${e.column ? `:${e.column}` : ""}`;
	return t && e.file ? `${e.file}${n}` : n;
}
function fc({ violation: e, workDir: t, onOpenFile: n }) {
	let r = e.locations ?? [];
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "mt-1 space-y-0.5",
		children: [/* @__PURE__ */ (0, L.jsx)("div", {
			className: "text-[10px] uppercase tracking-wide text-muted-foreground",
			children: `${r.length} locations`
		}), /* @__PURE__ */ (0, L.jsx)("ul", {
			className: "space-y-0.5",
			children: r.map((e, r) => {
				let i = dt(e.file, t) ?? e.file;
				return /* @__PURE__ */ (0, L.jsxs)("li", {
					className: "flex flex-wrap items-baseline gap-2 text-xs",
					"data-testid": "lint-violation-location",
					children: [
						/* @__PURE__ */ (0, L.jsx)("span", {
							className: "font-mono text-muted-foreground",
							children: `${i}${e.line ? `:${e.line}` : ""}`
						}),
						e.message && /* @__PURE__ */ (0, L.jsx)("span", { children: e.message }),
						n && /* @__PURE__ */ (0, L.jsx)(N, {
							target: {
								file: i,
								...t ? { workDir: t } : {},
								...e.line ? { line: e.line } : {}
							},
							onOpenFile: n
						})
					]
				}, r);
			})
		})]
	});
}
function pc({ violation: e, file: t, workDir: n, loadSource: r, onOpenFile: i, showFile: a, showMessage: o = !0 }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "space-y-2",
		"data-lint-violation-detail": "",
		children: [
			a && /* @__PURE__ */ (0, L.jsx)("p", {
				className: "break-all font-mono text-xs text-muted-foreground",
				children: dc(e, !0)
			}),
			/* @__PURE__ */ (0, L.jsx)(rc, {
				violation: e,
				file: t,
				workDir: n,
				loadSource: r,
				showMessage: o
			}),
			(e.locations?.length ?? 0) > 1 && /* @__PURE__ */ (0, L.jsx)(fc, {
				violation: e,
				workDir: n,
				onOpenFile: i
			})
		]
	});
}
function mc(e) {
	return /* @__PURE__ */ (0, L.jsx)("div", {
		"data-lint-violation": "",
		children: /* @__PURE__ */ (0, L.jsx)(Ct, {
			...e.expanded === void 0 ? {} : { expanded: e.expanded },
			...e.onExpandedChange ? { onExpandedChange: e.onExpandedChange } : {},
			message: e.violation.message ?? "",
			severity: Wt(e.violation),
			metadata: `${e.linter} ${e.violation.rule?.method ?? ""}`,
			location: dc(e.violation, !!e.showFile),
			actions: Qs(e),
			children: /* @__PURE__ */ (0, L.jsx)(pc, {
				...e,
				showMessage: !1
			})
		})
	});
}
function hc(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) {
		let e = n.rule?.method || "(no rule)";
		t.set(e, [...t.get(e) ?? [], n]);
	}
	return Array.from(t.keys()).sort().map((e) => ({
		rule: e,
		violations: t.get(e).slice().sort((e, t) => (e.line ?? 0) - (t.line ?? 0))
	}));
}
function gc(e) {
	return e === "" ? "''" : /^[A-Za-z0-9_./:=@%+,-]+$/.test(e) ? e : `'${e.replace(/'/g, "'\\''")}'`;
}
function _c(e) {
	return e.command ? [e.command, ...e.args ?? []].map(gc).join(" ") : "";
}
function vc(e) {
	return e.status === "pending" ? ["pending", "text-muted-foreground"] : e.status === "running" ? ["running", "text-blue-600 dark:text-blue-400"] : e.skipped ? ["skipped", "text-muted-foreground"] : e.timed_out ? ["timed out", "text-amber-600 dark:text-amber-400"] : Pt(e) ? ["failed", "text-red-600 dark:text-red-400"] : ["ran", "text-emerald-600 dark:text-emerald-400"];
}
function yc({ result: e }) {
	let [t, n] = vc(e), r = (e.violations ?? []).length, i = oc(e);
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex flex-wrap gap-3 text-sm",
		children: [
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: n,
				children: t
			}),
			i && e.cache && /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "text-muted-foreground",
				title: `${e.cache.source} fingerprint ${e.cache.fingerprint ?? "unknown"}`,
				children: ["cached · linted ", i]
			}),
			/* @__PURE__ */ (0, L.jsxs)("span", {
				className: "text-muted-foreground",
				children: [
					r,
					" violation",
					r === 1 ? "" : "s"
				]
			}),
			e.file_count !== void 0 && /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "text-muted-foreground",
				children: [e.file_count, " files"]
			}),
			e.rule_count !== void 0 && /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "text-muted-foreground",
				children: [e.rule_count, " rules"]
			}),
			e.duration ? /* @__PURE__ */ (0, L.jsx)("span", {
				className: "text-muted-foreground",
				children: bt(e.duration)
			}) : null
		]
	});
}
function bc({ label: e, value: t }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex items-center gap-1.5 text-xs",
		children: [
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: "text-muted-foreground",
				children: e
			}),
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: "break-all font-mono text-xs",
				children: t
			}),
			/* @__PURE__ */ (0, L.jsx)(ue, {
				value: t,
				label: `Copy ${e}`
			})
		]
	});
}
function xc({ result: e }) {
	return !e.run_id && !e.session_id ? null : /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "space-y-0.5",
		children: [e.run_id && /* @__PURE__ */ (0, L.jsx)(bc, {
			label: "run id",
			value: e.run_id
		}), e.session_id && /* @__PURE__ */ (0, L.jsx)(bc, {
			label: "session id",
			value: e.session_id
		})]
	});
}
function Sc({ result: e, workDir: t }) {
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsxs)(cc, {
		title: "Status",
		children: [/* @__PURE__ */ (0, L.jsx)(yc, { result: e }), /* @__PURE__ */ (0, L.jsx)(xc, { result: e })]
	}), e.command && /* @__PURE__ */ (0, L.jsxs)(cc, {
		title: "Command",
		children: [/* @__PURE__ */ (0, L.jsx)(uc, { children: _c(e) }), t && /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "font-mono text-xs text-muted-foreground",
			children: ["cwd: ", t]
		})]
	})] });
}
function Cc({ violations: e, linter: t, workDir: n }) {
	return e.length === 0 ? null : /* @__PURE__ */ (0, L.jsx)(cc, {
		title: "Violations Without File",
		children: /* @__PURE__ */ (0, L.jsx)("div", {
			className: "space-y-3",
			children: hc(e).map(({ rule: e, violations: r }) => /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "space-y-2 rounded border border-border bg-muted/40 p-3",
				children: [/* @__PURE__ */ (0, L.jsxs)("div", {
					className: "font-mono text-sm",
					children: [
						e,
						" ",
						/* @__PURE__ */ (0, L.jsxs)("span", {
							className: "text-xs text-muted-foreground",
							children: [
								"(",
								r.length,
								")"
							]
						})
					]
				}), r.map((e, r) => /* @__PURE__ */ (0, L.jsx)(mc, {
					violation: e,
					linter: t,
					file: "",
					workDir: n
				}, r))]
			}, e))
		})
	});
}
function wc({ node: e, renderRun: t }) {
	let n = e.linter, r = e.linterName ?? n?.linter ?? "", i = e.work_dir ?? n?.work_dir ?? "", a = n && t?.(n);
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		n && /* @__PURE__ */ (0, L.jsx)(Sc, {
			result: n,
			workDir: i
		}),
		a && /* @__PURE__ */ (0, L.jsx)(cc, {
			title: "Run",
			children: a
		}),
		n?.error && /* @__PURE__ */ (0, L.jsx)(cc, {
			title: "Error",
			children: /* @__PURE__ */ (0, L.jsx)(uc, {
				tone: "error",
				children: n.error
			})
		}),
		n?.raw_output && /* @__PURE__ */ (0, L.jsx)(cc, {
			title: "Raw output",
			children: /* @__PURE__ */ (0, L.jsx)(uc, { children: n.raw_output })
		}),
		/* @__PURE__ */ (0, L.jsx)(Cc, {
			violations: e.noFileViolations ?? [],
			linter: r,
			workDir: i
		})
	] });
}
function Tc({ node: e }) {
	let t = e.linter;
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		/* @__PURE__ */ (0, L.jsx)(cc, {
			title: "Error",
			children: /* @__PURE__ */ (0, L.jsx)(uc, {
				tone: "error",
				children: e.message
			})
		}),
		t && /* @__PURE__ */ (0, L.jsx)(Sc, {
			result: t,
			workDir: t.work_dir ?? ""
		}),
		t?.raw_output && /* @__PURE__ */ (0, L.jsx)(cc, {
			title: "Raw output",
			children: /* @__PURE__ */ (0, L.jsx)(uc, { children: t.raw_output })
		})
	] });
}
function Ec({ node: e, results: t }) {
	let n = e.ruleName ? [] : Hs(e, t).filter((t) => !e.linterName || t.linter === e.linterName);
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsx)(cc, {
		title: "Folder",
		children: /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex items-center gap-2",
			children: [/* @__PURE__ */ (0, L.jsx)(lc, { children: e.file || e.name }), e.target_path !== void 0 && /* @__PURE__ */ (0, L.jsx)("span", {
				className: "text-xs text-muted-foreground",
				children: Vs(e.target_path)
			})]
		})
	}), e.ruleName ? /* @__PURE__ */ (0, L.jsx)(cc, {
		title: "Scope",
		children: /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "space-y-1 text-sm",
			children: [/* @__PURE__ */ (0, L.jsxs)("div", { children: [
				/* @__PURE__ */ (0, L.jsx)("span", {
					className: "text-muted-foreground",
					children: "Linter:"
				}),
				" ",
				/* @__PURE__ */ (0, L.jsx)(lc, { children: e.linterName })
			] }), /* @__PURE__ */ (0, L.jsxs)("div", { children: [
				/* @__PURE__ */ (0, L.jsx)("span", {
					className: "text-muted-foreground",
					children: "Rule:"
				}),
				" ",
				/* @__PURE__ */ (0, L.jsx)(lc, { children: e.ruleName })
			] })]
		})
	}) : /* @__PURE__ */ (0, L.jsx)(cc, {
		title: "Linters",
		children: /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "space-y-1",
			children: [n.map(({ linter: e, count: t }) => /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex items-center justify-between text-sm",
				children: [/* @__PURE__ */ (0, L.jsx)(lc, { children: e }), /* @__PURE__ */ (0, L.jsxs)("span", {
					className: "text-xs text-muted-foreground",
					children: [t, " violations"]
				})]
			}, e)), n.length === 0 && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "text-sm text-muted-foreground",
				children: "No violations found under this folder."
			})]
		})
	})] });
}
function Dc({ node: e, ...t }) {
	let n = e.violations ?? [];
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsx)(cc, {
		title: "File",
		children: /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex items-center gap-2",
			children: [/* @__PURE__ */ (0, L.jsx)(lc, { children: e.file }), n.length > 0 && /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "text-xs text-muted-foreground",
				children: [
					"(",
					n.length,
					" violations)"
				]
			})]
		})
	}), n.length > 0 && /* @__PURE__ */ (0, L.jsx)(cc, {
		title: "Violations",
		children: /* @__PURE__ */ (0, L.jsx)("div", {
			className: "space-y-2",
			children: n.map((n, r) => /* @__PURE__ */ (0, L.jsx)(mc, {
				violation: n,
				linter: e.linterName ?? "",
				file: e.target_path || e.file || "",
				workDir: e.work_dir ?? "",
				...t
			}, r))
		})
	})] });
}
function Oc({ node: e, ...t }) {
	let n = e.violations ?? [], r = e.linterName ?? "", i = e.work_dir ?? "";
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		/* @__PURE__ */ (0, L.jsxs)(cc, {
			title: "Rule",
			children: [/* @__PURE__ */ (0, L.jsx)(lc, { children: e.ruleName }), /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "ml-2 text-xs text-muted-foreground",
				children: [
					"(",
					Ht(e),
					" violations)"
				]
			})]
		}),
		e.file && /* @__PURE__ */ (0, L.jsx)(cc, {
			title: "File",
			children: /* @__PURE__ */ (0, L.jsx)(lc, { children: e.file })
		}),
		n.length > 0 && /* @__PURE__ */ (0, L.jsx)(cc, {
			title: "Violations",
			children: /* @__PURE__ */ (0, L.jsx)("div", {
				className: "space-y-2",
				children: n.map((n, a) => /* @__PURE__ */ (0, L.jsx)(mc, {
					violation: n,
					linter: r,
					file: n.raw_file || e.target_path || e.file || "",
					workDir: i,
					showFile: !e.file,
					...t
				}, a))
			})
		}),
		/* @__PURE__ */ (0, L.jsx)(Cc, {
			violations: e.noFileViolations ?? [],
			linter: r,
			workDir: i
		})
	] });
}
var kc = {
	linter: "Linter",
	"lint-folder": "Folder",
	"lint-file": "File",
	"lint-rule": "Rule",
	"lint-rule-group": "Rule",
	"linter-error": "Linter error"
};
function Ac({ node: e, actions: t, onIgnore: n, ignoreBusy: r }) {
	let i = e.kind === "linter-error" ? 0 : Ht(e);
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex min-w-0 items-center gap-2 border-b border-border px-density-4 py-density-2",
		"data-detail-header": "",
		children: [
			e.kind === "linter" ? /* @__PURE__ */ (0, L.jsx)(st, {
				tool: Rt(e),
				className: "text-lg"
			}) : e.kind === "linter-error" ? /* @__PURE__ */ (0, L.jsx)(gt, {
				severity: "error",
				className: "text-lg"
			}) : null,
			/* @__PURE__ */ (0, L.jsx)("h2", {
				className: "min-w-0 shrink truncate text-base font-semibold",
				title: e.name,
				children: e.name
			}),
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: "inline-flex shrink-0 items-center rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground",
				children: kc[e.kind]
			}),
			i > 0 && /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "shrink-0 text-xs text-muted-foreground",
				children: [
					i,
					" violation",
					i === 1 ? "" : "s"
				]
			}),
			n && t.length > 0 && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "ml-auto shrink-0",
				"data-lint-node-actions": "",
				children: /* @__PURE__ */ (0, L.jsx)(ic, {
					actions: t,
					onIgnore: n,
					busy: !!r,
					align: "right"
				})
			})
		]
	});
}
function jc({ node: e, results: t, onIgnore: n, ignoreBusy: r, onOpenFile: i, loadSource: a, renderRun: o }) {
	if (!e) return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "flex h-full items-center justify-center text-sm text-muted-foreground",
		children: /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "text-center",
			children: [/* @__PURE__ */ (0, L.jsx)(F, {
				icon: it,
				className: "mb-2 block text-4xl"
			}), "Select a linter, rule or file to view its findings."]
		})
	});
	let s = {
		onIgnore: n,
		ignoreBusy: r,
		onOpenFile: i,
		loadSource: a
	};
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		children: [/* @__PURE__ */ (0, L.jsx)(Ac, {
			node: e,
			actions: Js(e, t),
			onIgnore: n,
			ignoreBusy: r
		}), /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "min-h-0 flex-1 space-y-4 overflow-y-auto p-density-4",
			children: [
				e.kind === "linter" && /* @__PURE__ */ (0, L.jsx)(wc, {
					node: e,
					renderRun: o
				}),
				e.kind === "linter-error" && /* @__PURE__ */ (0, L.jsx)(Tc, { node: e }),
				e.kind === "lint-folder" && /* @__PURE__ */ (0, L.jsx)(Ec, {
					node: e,
					results: t
				}),
				e.kind === "lint-file" && /* @__PURE__ */ (0, L.jsx)(Dc, {
					node: e,
					...s
				}),
				(e.kind === "lint-rule" || e.kind === "lint-rule-group") && /* @__PURE__ */ (0, L.jsx)(Oc, {
					node: e,
					...s
				})
			]
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/LintTable.tsx
var Mc = {
	error: 0,
	warning: 1,
	info: 2
};
function Nc(e, t) {
	let n = St(e, {
		severity: /* @__PURE__ */ new Map(),
		linter: /* @__PURE__ */ new Map()
	}), r = xt(n.map((e) => [
		e.workDir,
		e.rawFile,
		e.linter,
		e.v.rule?.method,
		e.v.line,
		e.v.column,
		e.v.end_line,
		e.v.end_column,
		e.v.message,
		e.v.severity
	]));
	return n.map((e, t) => ({
		id: r[t],
		severity: Wt(e.v),
		linter: e.linter,
		rule: e.v.rule?.method ?? "",
		location: e.file ? `${e.file}${e.v.line ? `:${e.v.line}` : ""}` : "",
		message: e.v.message ?? "",
		flat: e
	})).filter(({ flat: e }) => Ge(e.linter, t.linter) && Ge(Wt(e.v), t.severity));
}
var Pc = [
	{
		key: "severity",
		label: "Severity",
		sortable: !0,
		sortValue: (e) => Mc[e],
		shrink: !0,
		render: (e) => /* @__PURE__ */ (0, L.jsx)("span", {
			className: "inline-flex items-center gap-1",
			children: /* @__PURE__ */ (0, L.jsx)(gt, { severity: e })
		})
	},
	{
		key: "linter",
		label: "Linter",
		sortable: !0,
		shrink: !0,
		render: (e) => /* @__PURE__ */ (0, L.jsxs)("span", {
			className: "inline-flex items-center gap-1",
			children: [/* @__PURE__ */ (0, L.jsx)(st, { tool: e }), lt(e)?.label ?? e]
		})
	},
	{
		key: "rule",
		label: "Rule",
		sortable: !0,
		render: (e) => /* @__PURE__ */ (0, L.jsx)("span", {
			className: "font-mono text-xs",
			children: e
		})
	},
	{
		key: "location",
		label: "Location",
		sortable: !0,
		render: (e) => /* @__PURE__ */ (0, L.jsx)("span", {
			title: e,
			className: "block max-w-40 truncate font-mono text-xs",
			children: e
		})
	},
	{
		key: "message",
		label: "Message",
		grow: !0
	}
];
function Fc({ results: e, filters: t, className: n, ...r }) {
	let { ref: i, wider: a } = _(640), s = (0, I.useMemo)(() => [...Pc, {
		key: "actions",
		label: "",
		shrink: !0,
		render: (e, t) => /* @__PURE__ */ (0, L.jsx)(zt, { items: Qs({
			violation: t.flat.v,
			linter: t.flat.linter,
			file: t.flat.rawFile ?? "",
			workDir: t.flat.workDir ?? "",
			...r
		}) })
	}], [r]), c = (0, I.useMemo)(() => Nc(e, {
		severity: /* @__PURE__ */ new Map(),
		linter: /* @__PURE__ */ new Map()
	}).map((e) => e.id), [e]), [l, u] = At(c), d = (0, I.useMemo)(() => Nc(e, t), [e, t]);
	return /* @__PURE__ */ (0, L.jsx)("div", {
		ref: i,
		className: "min-w-0",
		children: a ? /* @__PURE__ */ (0, L.jsx)(o, {
			data: d,
			columns: s,
			getRowId: (e) => e.id,
			defaultSort: {
				key: "severity",
				dir: "asc"
			},
			emptyMessage: "No violations match the current filters",
			detailStyle: "row",
			expandedRowIds: l,
			onExpandedRowIdsChange: u,
			renderExpandedRow: ({ flat: e }) => /* @__PURE__ */ (0, L.jsx)(pc, {
				violation: e.v,
				linter: e.linter,
				file: e.rawFile ?? "",
				workDir: e.workDir ?? "",
				...r
			}),
			...n ? { className: n } : {}
		}) : /* @__PURE__ */ (0, L.jsx)("div", { children: d.slice().sort((e, t) => Mc[e.severity] - Mc[t.severity]).map(({ id: e, flat: t }) => /* @__PURE__ */ (0, L.jsx)(mc, {
			expanded: l.includes(e),
			onExpandedChange: (t) => u((n) => t ? [...n, e] : n.filter((t) => t !== e)),
			violation: t.v,
			linter: t.linter,
			file: t.rawFile ?? "",
			workDir: t.workDir ?? "",
			showFile: !0,
			...r
		}, e)) })
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/grouping-options.ts
var Ic = [
	{
		key: "linter-rule-file",
		label: "Linter · Rule · File",
		icon: $t
	},
	{
		key: "linter-file",
		label: "Linter · File",
		icon: at
	},
	{
		key: "file-linter-rule",
		label: "File · Linter · Rule",
		icon: tt
	},
	{
		key: "summary",
		label: "Summary",
		icon: ct
	}
], Lc = [
	"error",
	"warning",
	"info"
];
function W({ grouping: e, onGroupingChange: t, className: n }) {
	return /* @__PURE__ */ (0, L.jsx)("div", {
		role: "group",
		"aria-label": "Group by",
		className: P("space-y-1", n),
		children: Ic.map((n) => /* @__PURE__ */ (0, L.jsx)("button", {
			type: "button",
			"aria-pressed": n.key === e,
			className: P("block w-full rounded px-2 py-1 text-left text-xs hover:bg-muted", n.key === e && "bg-muted font-medium"),
			title: `Group by ${n.label}`,
			onClick: () => t(n.key),
			children: n.label
		}, n.key))
	});
}
function Rc({ results: e, filters: t, onFiltersChange: n, className: r }) {
	let i = Lt(e, t.linter), a = Vt(e, t.severity), o = [.../* @__PURE__ */ new Set([...Bt(e), ...t.linter.keys()])], s = t.severity.size > 0 || t.linter.size > 0, c = Lc.filter((e) => i[e] > 0 || t.severity.size > 0).map((e) => ({
		value: e,
		label: pt[e].label,
		icon: /* @__PURE__ */ (0, L.jsx)(gt, { severity: e }),
		count: i[e]
	})), l = o.map((e) => ({
		value: e,
		label: lt(e)?.label ?? e,
		title: e,
		icon: /* @__PURE__ */ (0, L.jsx)(st, { tool: e }),
		count: a[e] ?? 0
	}));
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("space-y-3", r),
		"data-lint-filter-facets": "",
		children: [
			/* @__PURE__ */ (0, L.jsx)(Mt, {
				label: "Severity",
				options: c,
				state: t.severity,
				onChange: (e) => n({
					...t,
					severity: e
				})
			}),
			(o.length > 1 || t.linter.size > 0) && /* @__PURE__ */ (0, L.jsx)(Mt, {
				label: "Linter",
				options: l,
				state: t.linter,
				onChange: (e) => n({
					...t,
					linter: e
				})
			}),
			s && /* @__PURE__ */ (0, L.jsx)("button", {
				type: "button",
				className: "ml-1 inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
				onClick: () => n({
					severity: /* @__PURE__ */ new Map(),
					linter: /* @__PURE__ */ new Map()
				}),
				children: "Clear filters"
			})
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/LintTreeNode.tsx
function zc(e) {
	if (e.kind === "linter") return e.children?.find((e) => e.kind === "linter-error")?.message;
}
function Bc({ node: e }) {
	switch (e.kind) {
		case "linter": return e.linter?.status === "running" ? /* @__PURE__ */ (0, L.jsx)(vt, { outcome: "running" }) : e.linter?.status === "pending" ? /* @__PURE__ */ (0, L.jsx)(F, {
			icon: x,
			title: "Pending",
			className: "shrink-0 text-muted-foreground"
		}) : zc(e) === void 0 ? lt(Rt(e)) ? /* @__PURE__ */ (0, L.jsx)(st, { tool: Rt(e) }) : /* @__PURE__ */ (0, L.jsx)(F, {
			icon: le,
			className: "shrink-0 text-muted-foreground"
		}) : /* @__PURE__ */ (0, L.jsx)(gt, { severity: "error" });
		case "lint-folder": return /* @__PURE__ */ (0, L.jsx)(F, {
			icon: at,
			className: "shrink-0 text-muted-foreground"
		});
		case "lint-file": return /* @__PURE__ */ (0, L.jsx)(F, {
			icon: tt,
			className: "shrink-0 text-muted-foreground"
		});
		case "lint-rule":
		case "lint-rule-group": return /* @__PURE__ */ (0, L.jsx)(F, {
			icon: an,
			className: "shrink-0 text-muted-foreground"
		});
		case "linter-error": return /* @__PURE__ */ (0, L.jsx)(gt, { severity: "error" });
	}
}
function Vc({ node: e, selected: t, actions: n = [], onIgnore: r, ignoreBusy: i = !1 }) {
	let a = e.kind === "linter-error" ? 0 : Ht(e), o = zc(e), s = o !== void 0, c = e.skipped ? e.message : s ? "failed to run" : void 0, l = s ? o : c, u = e.kind === "linter" ? oc(e.linter) : void 0, d = u ? `cached · ${u.replace(/ ago$/, "")}` : e.kind === "linter" && !e.linter?.status && e.linter?.duration ? bt(e.linter.duration) : "";
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		/* @__PURE__ */ (0, L.jsx)(Bc, { node: e }),
		/* @__PURE__ */ (0, L.jsx)("span", {
			className: P("min-w-0 truncate", t ? "font-semibold" : "font-medium", e.kind === "linter-error" || s ? "text-red-600 dark:text-red-400" : "text-foreground", (e.skipped || e.kind === "linter" && e.linter?.status === "pending") && "text-muted-foreground"),
			title: e.name,
			children: e.name
		}),
		/* @__PURE__ */ (0, L.jsx)("span", { className: "flex-1" }),
		c && /* @__PURE__ */ (0, L.jsx)("span", {
			className: P("min-w-0 shrink-[100] truncate text-xs", s ? "text-red-600 dark:text-red-400" : "text-muted-foreground"),
			title: l,
			children: c
		}),
		d && /* @__PURE__ */ (0, L.jsx)("span", {
			className: "shrink-0 text-xs tabular-nums text-muted-foreground",
			children: d
		}),
		a > 0 && /* @__PURE__ */ (0, L.jsx)("span", {
			className: "flex shrink-0",
			title: `${a} violation${a === 1 ? "" : "s"}`,
			children: /* @__PURE__ */ (0, L.jsx)("span", {
				className: "text-xs tabular-nums text-muted-foreground",
				children: a
			})
		}),
		r && /* @__PURE__ */ (0, L.jsx)(ac, {
			actions: n,
			onIgnore: r,
			busy: i,
			label: e.name
		})
	] });
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/lint/LintResults.tsx
var Hc = [{
	id: "tree",
	label: "Tree",
	icon: $t
}, {
	id: "table",
	label: "Table",
	icon: en
}], Uc = 640, Wc = [
	"error",
	"warning",
	"info"
];
function Gc({ results: e }) {
	let t = Lt(e, /* @__PURE__ */ new Map()), n = e.filter(Pt).length;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "ml-auto flex shrink-0 items-center gap-3 text-xs text-muted-foreground",
		"data-lint-summary": "",
		children: [Wc.filter((e) => t[e] > 0).map((e) => /* @__PURE__ */ (0, L.jsxs)("span", {
			className: "inline-flex items-center gap-1",
			children: [/* @__PURE__ */ (0, L.jsx)(gt, { severity: e }), /* @__PURE__ */ (0, L.jsx)("span", {
				className: "tabular-nums",
				children: t[e]
			})]
		}, e)), n > 0 && /* @__PURE__ */ (0, L.jsxs)("span", {
			className: "text-red-600 dark:text-red-400",
			children: [
				n,
				" linter",
				n === 1 ? "" : "s",
				" failed to run"
			]
		})]
	});
}
function Kc(e, t) {
	return e.length === 0 ? "No lint results" : t.severity.size > 0 || t.linter.size > 0 ? "No violations match the current filters" : "No lint violations";
}
function qc({ results: e, grouping: t, onGroupingChange: n, filters: r, onFiltersChange: i, selected: a, onSelect: o, expandAll: s, onExpandAllChange: c, onIgnore: l, ignoreBusy: u, onOpenFile: d, loadSource: f, renderRun: p, headerSlot: m, title: h, view: g = "tree", onViewChange: v, defaultSplit: y = 45, className: b }) {
	let x = (0, I.useMemo)(() => Et({
		results: e,
		grouping: t,
		filters: r
	}), [
		e,
		t,
		r
	]), S = (0, I.useMemo)(() => a ? Nt(x, a) : null, [x, a]), { ref: C, wider: w } = _(Uc), T = /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex flex-wrap items-center gap-2 border-b border-border bg-card px-density-4 py-density-2",
		"data-lint-filter-bar": "",
		children: [v && /* @__PURE__ */ (0, L.jsx)(we, {
			size: "sm",
			"aria-label": "View",
			value: g,
			options: Hc,
			onChange: v
		}), /* @__PURE__ */ (0, L.jsxs)(kt, {
			active: r.severity.size + r.linter.size,
			children: [g === "tree" && /* @__PURE__ */ (0, L.jsx)(W, {
				grouping: t,
				onGroupingChange: n
			}), /* @__PURE__ */ (0, L.jsx)(Rc, {
				results: e,
				filters: r,
				onFiltersChange: i
			})]
		})]
	}), E = /* @__PURE__ */ (0, L.jsx)($e, {
		className: "min-h-0",
		roots: x,
		getKey: (e) => e.route_path,
		getChildren: (e) => e.children,
		selected: S,
		onSelect: (e) => o(e.route_path),
		expandAll: s,
		onExpandAllChange: c,
		defaultOpen: (e, t) => t < 1 || e.failed,
		showSearch: !0,
		searchLabel: "Filter findings",
		getSearchText: (e) => `${e.name} ${e.linterName ?? ""} ${e.ruleName ?? ""}`,
		empty: /* @__PURE__ */ (0, L.jsx)("div", {
			className: "p-8 text-center text-sm text-muted-foreground",
			children: Kc(e, r)
		}),
		renderRow: ({ node: t, selected: n }) => /* @__PURE__ */ (0, L.jsx)(Vc, {
			node: t,
			selected: n,
			actions: l ? Js(t, e) : [],
			onIgnore: l,
			ignoreBusy: u
		})
	}), D = /* @__PURE__ */ (0, L.jsx)(jc, {
		node: S,
		results: e,
		onIgnore: l,
		ignoreBusy: u,
		onOpenFile: d,
		loadSource: f,
		renderRun: p
	}), O = h !== null || e.length > 0 || m != null;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		ref: C,
		className: P("flex h-full min-h-0 flex-col bg-background", b),
		"data-lint-results": "",
		children: [O && /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1 border-b border-border bg-card px-density-4 py-density-2",
			"data-lint-header": "",
			children: [
				h === void 0 ? /* @__PURE__ */ (0, L.jsxs)("h1", {
					className: "inline-flex shrink-0 items-center gap-1.5 text-base font-bold",
					children: [/* @__PURE__ */ (0, L.jsx)(F, {
						icon: it,
						className: "text-primary"
					}), "Lint Results"]
				}) : h,
				/* @__PURE__ */ (0, L.jsx)(Gc, { results: e }),
				m && /* @__PURE__ */ (0, L.jsx)("div", {
					className: "flex shrink-0 items-center gap-1",
					children: m
				})
			]
		}), g === "table" ? /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [T, /* @__PURE__ */ (0, L.jsx)("div", {
			className: "min-h-0 flex-1 overflow-y-auto",
			"data-lint-table": "",
			children: /* @__PURE__ */ (0, L.jsx)(Fc, {
				results: e,
				filters: r,
				onIgnore: l,
				ignoreBusy: u,
				onOpenFile: d,
				loadSource: f
			})
		})] }) : w ? /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [T, /* @__PURE__ */ (0, L.jsx)(rt, {
			className: "min-h-0 flex-1",
			defaultSplit: y,
			minLeft: 25,
			minRight: 30,
			leftClass: "border-r-2 border-border bg-muted/30",
			left: E,
			right: D
		})] }) : S ? /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex min-h-0 flex-1 flex-col",
			children: [/* @__PURE__ */ (0, L.jsxs)("button", {
				type: "button",
				onClick: () => o(null),
				className: "inline-flex items-center gap-1.5 border-b border-border px-density-3 py-density-2 text-left text-sm text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
				children: [/* @__PURE__ */ (0, L.jsx)(F, { icon: Xe }), "Back to results"]
			}), /* @__PURE__ */ (0, L.jsx)("div", {
				className: "min-h-0 flex-1 overflow-hidden",
				children: D
			})]
		}) : /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [T, /* @__PURE__ */ (0, L.jsx)("div", {
			className: "min-h-0 flex-1 overflow-y-auto",
			children: E
		})] })]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/layout/PanelFrame.tsx
var Jc = {
	default: "",
	danger: "border-l-2 border-l-red-500",
	warning: "border-l-2 border-l-yellow-500",
	success: "border-l-2 border-l-green-500",
	info: "border-l-2 border-l-blue-500"
};
function Yc({ header: e, footer: t, tone: n = "default", padded: r = !0, bodyVisible: i = !0, className: a, headerClassName: o, bodyClassName: s, children: c }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("overflow-hidden rounded-md border border-border bg-card", Jc[n], a),
		children: [
			e !== void 0 && /* @__PURE__ */ (0, L.jsx)("div", {
				className: P("flex items-center gap-2 px-density-3 py-density-2", i && "border-b border-border", o),
				children: e
			}),
			i && /* @__PURE__ */ (0, L.jsx)("div", {
				className: P(r && "px-density-3 py-density-2", s),
				children: c
			}),
			i && t !== void 0 && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "border-t border-border px-density-3 py-density-2",
				children: t
			})
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/layout/Section.tsx
function Xc({ title: e, summary: t, collapsible: n = !0, defaultOpen: r = !1, open: i, onToggle: a, icon: o, tone: s = "default", className: c, headerClassName: l, bodyClassName: u, children: d }) {
	let f = i !== void 0, [p, m] = (0, I.useState)(r), h = !n || (f ? i : p);
	function g() {
		let e = !h;
		f || m(e), a?.(e);
	}
	let _ = /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		n && /* @__PURE__ */ (0, L.jsx)(F, {
			icon: h ? Ze : Ie,
			className: "text-muted-foreground text-xs"
		}),
		o && /* @__PURE__ */ (0, L.jsx)(F, {
			...typeof o == "string" ? { name: o } : { icon: o },
			className: "text-base"
		}),
		/* @__PURE__ */ (0, L.jsx)("span", {
			className: "font-medium text-sm flex-1 truncate",
			children: e
		})
	] });
	return /* @__PURE__ */ (0, L.jsx)(Yc, {
		header: /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [n ? /* @__PURE__ */ (0, L.jsx)("button", {
			type: "button",
			onClick: g,
			"aria-expanded": h,
			className: "flex min-w-0 flex-1 items-center gap-2 rounded-sm text-left transition-colors hover:bg-accent/50 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
			children: _
		}) : /* @__PURE__ */ (0, L.jsx)("div", {
			className: "flex min-w-0 flex-1 items-center gap-2",
			children: _
		}), t && /* @__PURE__ */ (0, L.jsx)("span", {
			className: "shrink-0 text-xs text-muted-foreground",
			children: t
		})] }),
		tone: s,
		bodyVisible: h,
		className: P("overflow-visible bg-background", c),
		headerClassName: P("w-full text-left", l),
		...u === void 0 ? {} : { bodyClassName: u },
		children: d
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/components/Switch.tsx
function Zc({ checked: e, onChange: t, disabled: n, label: r, className: i, "aria-label": a }) {
	let o = (0, I.useId)(), s = /* @__PURE__ */ (0, L.jsx)("button", {
		type: "button",
		role: "switch",
		"aria-checked": e,
		"aria-label": r ? void 0 : a,
		"aria-labelledby": r ? o : void 0,
		disabled: n,
		onClick: () => t(!e),
		className: P("inline-flex shrink-0 items-center rounded-full px-0.5 transition-colors", "h-5 w-9 density-compact:h-4 density-compact:w-7 density-spacious:h-6 density-spacious:w-11", "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2", "disabled:cursor-not-allowed disabled:opacity-50", e ? "justify-end bg-primary" : "justify-start bg-input", !r && i),
		children: /* @__PURE__ */ (0, L.jsx)("span", { className: P("pointer-events-none rounded-full bg-background shadow transition-all", "h-4 w-4 density-compact:h-3 density-compact:w-3 density-spacious:h-5 density-spacious:w-5") })
	});
	return r ? /* @__PURE__ */ (0, L.jsxs)("label", {
		className: P("inline-flex cursor-pointer items-center gap-density-2", i),
		children: [s, /* @__PURE__ */ (0, L.jsx)("span", {
			id: o,
			className: "text-sm text-foreground",
			children: r
		})]
	}) : s;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/diagnostics/utils.ts
function Qc(e) {
	let t = (e || "").toLowerCase();
	return t.includes("run") ? ce : t.includes("sleep") || t.includes("idle") ? y : t.includes("stop") || t.includes("halt") ? nn : t.includes("zombie") || t.includes("dead") ? wt : t.includes("wait") || t.includes("block") ? fe : Zt;
}
function $c(e) {
	let t = (e || "").toLowerCase();
	return t.includes("run") ? "text-green-600" : t.includes("sleep") || t.includes("idle") ? "text-amber-500" : t.includes("stop") || t.includes("halt") ? "text-orange-600" : t.includes("zombie") || t.includes("dead") ? "text-red-600" : t.includes("wait") || t.includes("block") ? "text-blue-600" : "text-muted-foreground";
}
function el(e) {
	if (e.name) return e.name;
	if (e.command) {
		let [t] = e.command.split(/\s+/, 1);
		return t || `pid ${e.pid}`;
	}
	return `pid ${e.pid}`;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/hooks/use-resource-clock.ts
function tl(e) {
	let [t, n] = (0, I.useState)(Date.now);
	return (0, I.useEffect)(() => {
		if (!e) return;
		let t = setInterval(() => n(Date.now()), 1e3);
		return () => clearInterval(t);
	}, [e]), t;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/har/HarPanel.tsx
var nl = 80, rl = (0, I.createContext)(void 0), il = [
	{
		key: "request.method",
		label: "Method",
		shrink: !0
	},
	{
		key: "request.url",
		label: "URL",
		grow: !0,
		render: (e, t) => /* @__PURE__ */ (0, L.jsxs)("span", {
			className: "inline-flex min-w-0 items-baseline gap-density-2",
			children: [/* @__PURE__ */ (0, L.jsx)("span", {
				title: String(e ?? ""),
				children: String(e ?? "")
			}), t._error ? /* @__PURE__ */ (0, L.jsx)("span", {
				className: "text-destructive text-xs",
				title: t._error,
				children: ll(t._error)
			}) : null]
		})
	},
	{
		key: "response.status",
		label: "Status",
		shrink: !0,
		render: (e, t) => /* @__PURE__ */ (0, L.jsx)(ol, {
			status: Number(e ?? 0),
			entry: t
		}),
		sortValue: (e) => Number(e ?? 0),
		filterValue: (e, t) => cl(Number(e ?? 0), t)
	},
	{
		key: "time",
		label: "Time",
		align: "right",
		shrink: !0,
		render: (e, t) => t._pending ? /* @__PURE__ */ (0, L.jsx)(sl, { entry: t }) : `${Number(e ?? 0).toFixed(0)}ms`,
		sortValue: (e, t) => t._pending ? G(t, Date.now()) : Number(e ?? 0)
	},
	{
		key: "response.bodySize",
		label: "Size",
		align: "right",
		shrink: !0,
		render: (e) => vl(Number(e ?? 0)),
		sortValue: (e) => Number(e ?? 0)
	},
	{
		key: "response.content.mimeType",
		label: "Type",
		shrink: !0
	}
];
function al({ entries: e, search: t, emptyLabel: n = "No HTTP traffic captured", className: r, bodySource: i, menuActions: a }) {
	let s = (0, I.useMemo)(() => t === void 0 || t.trim() === "" ? e : e.filter((e) => _l(t.trim().toLowerCase(), e)), [e, t]), c = (0, I.useMemo)(() => e.filter((e) => e._pending).length, [e]), l = tl(c > 0);
	return !e || e.length === 0 ? /* @__PURE__ */ (0, L.jsx)("div", {
		className: "p-density-6 text-center text-muted-foreground text-sm",
		children: n
	}) : /* @__PURE__ */ (0, L.jsx)("div", {
		className: `flex h-full min-h-0 flex-col ${r ?? ""}`,
		children: /* @__PURE__ */ (0, L.jsx)(rl.Provider, {
			value: l,
			children: /* @__PURE__ */ (0, L.jsx)(o, {
				data: s,
				columns: il,
				autoFilter: !0,
				showGlobalFilter: t === void 0,
				globalFilterPlaceholder: "Filter URL, method, or body…",
				defaultSort: {
					key: "time",
					dir: "asc"
				},
				emptyMessage: n,
				...a ? { menuActions: a } : {},
				...c > 0 ? { footer: ({ visibleRowCount: e, totalRowCount: t }) => `${e} of ${t} row${t === 1 ? "" : "s"} · ${c} running` } : {},
				getRowId: (e, t) => e._id ?? `${e.startedDateTime ?? t}-${e.request.method}-${e.request.url}${e._pending ? "-pending" : `-${e.time}`}`,
				renderExpandedRow: (e) => /* @__PURE__ */ (0, L.jsx)(ul, {
					entry: e,
					bodySource: i
				})
			})
		})
	});
}
function ol({ status: e, entry: t }) {
	return t._pending ? /* @__PURE__ */ (0, L.jsxs)("span", {
		className: "inline-flex items-center gap-1 text-muted-foreground",
		children: [/* @__PURE__ */ (0, L.jsx)(F, {
			icon: ht,
			className: "animate-spin"
		}), "running"]
	}) : t._error ? /* @__PURE__ */ (0, L.jsx)("span", {
		className: "font-semibold text-destructive",
		children: "ERR"
	}) : /* @__PURE__ */ (0, L.jsx)("span", {
		className: yl(e),
		children: String(e)
	});
}
function sl({ entry: e }) {
	let t = (0, I.useContext)(rl);
	if (t === void 0) throw Error("PendingElapsed must render inside HarPanel's clock");
	return /* @__PURE__ */ (0, L.jsx)("span", {
		className: "tabular-nums",
		children: ae(G(e, t))
	});
}
function G(e, t) {
	let n = Date.parse(e.startedDateTime ?? "");
	return Number.isNaN(n) ? e.time : Math.max(0, t - n);
}
function cl(e, t) {
	return t._pending ? "running" : t._error ? "ERR" : String(e);
}
function ll(e) {
	return e.length > nl ? `${e.slice(0, nl)}…` : e;
}
function ul({ entry: e, bodySource: t }) {
	let n = (0, I.useContext)(rl), r = tl(!!e._pending && n === void 0);
	return /* @__PURE__ */ (0, L.jsx)(rl.Provider, {
		value: n ?? r,
		children: /* @__PURE__ */ (0, L.jsx)(dl, {
			entry: e,
			bodySource: t
		})
	});
}
function dl({ entry: e, bodySource: t }) {
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		e._error && /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "mb-density-3",
			children: [/* @__PURE__ */ (0, L.jsx)("div", {
				className: "font-semibold text-foreground mb-1",
				children: "Transport Error"
			}), /* @__PURE__ */ (0, L.jsx)("pre", {
				className: "whitespace-pre-wrap break-all text-destructive",
				children: e._error
			})]
		}),
		e._pending && /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "mb-density-3 text-muted-foreground",
			children: ["Waiting for response — in flight for ", /* @__PURE__ */ (0, L.jsx)(sl, { entry: e })]
		}),
		/* @__PURE__ */ (0, L.jsxs)("div", {
			className: "grid grid-cols-2 gap-density-4",
			children: [/* @__PURE__ */ (0, L.jsxs)("div", { children: [/* @__PURE__ */ (0, L.jsx)(pl, {
				title: "Request Headers",
				...e.request.headers ? { headers: e.request.headers } : {}
			}), (e.request.postData?.text || t?.hasBody(e, "request")) && /* @__PURE__ */ (0, L.jsx)(fl, {
				entry: e,
				part: "request",
				text: e.request.postData?.text,
				mimeType: e.request.postData?.mimeType,
				bodySource: t
			})] }), /* @__PURE__ */ (0, L.jsx)("div", { children: /* @__PURE__ */ (0, L.jsx)(pl, {
				title: "Response Headers",
				...e.response.headers ? { headers: e.response.headers } : {}
			}) })]
		}),
		(e.response.content?.text || t?.hasBody(e, "response")) && /* @__PURE__ */ (0, L.jsx)(fl, {
			entry: e,
			part: "response",
			text: e.response.content?.text,
			mimeType: e.response.content?.mimeType,
			bodySource: t
		})
	] });
}
function fl({ entry: e, part: t, text: n, mimeType: r, bodySource: i }) {
	let a = !n && i?.hasBody(e, t), o = e._id;
	if (a && !o) throw Error("Private HAR body requires a request ID");
	let [s, c] = (0, I.useState)(), [l, u] = (0, I.useState)();
	return (0, I.useEffect)(() => {
		if (!a || !i || !o) return;
		let e = !0;
		return c(void 0), u(void 0), i.preview(o, t).then((t) => {
			e && c(t);
		}, (t) => {
			e && u(t instanceof Error ? t : Error(String(t)));
		}), () => {
			e = !1;
		};
	}, [
		i,
		o,
		t,
		a
	]), /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "mt-density-3",
		children: [
			/* @__PURE__ */ (0, L.jsxs)("div", {
				className: "mb-1 flex items-center gap-2 font-semibold text-foreground",
				children: [t === "request" ? "Request Body" : "Response Body", a && i && /* @__PURE__ */ (0, L.jsx)("a", {
					className: "text-sm font-normal text-primary underline",
					href: i.downloadHref(e, t),
					download: !0,
					children: "Download full body"
				})]
			}),
			/* @__PURE__ */ (0, L.jsx)("div", {
				className: "bg-background p-density-2 rounded border border-border overflow-auto max-h-64",
				children: l ? /* @__PURE__ */ (0, L.jsx)("span", {
					role: "alert",
					className: "text-destructive",
					children: l.message
				}) : a && s === void 0 ? /* @__PURE__ */ (0, L.jsx)("span", {
					className: "text-muted-foreground",
					children: "Loading preview…"
				}) : /* @__PURE__ */ (0, L.jsx)(ml, {
					text: n || s || "",
					...r ? { mimeType: r } : {}
				})
			}),
			a && /* @__PURE__ */ (0, L.jsx)("p", {
				className: "mt-1 text-xs text-muted-foreground",
				children: "Preview shows at most the first 64 KiB."
			})
		]
	});
}
function pl({ title: e, headers: t }) {
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsx)("div", {
		className: "font-semibold text-foreground mb-1",
		children: e
	}), /* @__PURE__ */ (0, L.jsx)("div", {
		className: "space-y-0.5",
		children: t?.map((e, t) => /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "whitespace-nowrap",
			children: [
				/* @__PURE__ */ (0, L.jsxs)("span", {
					className: "text-purple-600 dark:text-purple-400",
					children: [e.name, ":"]
				}),
				" ",
				e.value
			]
		}, t))
	})] });
}
function ml({ text: e, mimeType: t }) {
	if (gl(t)) {
		let t = hl(e);
		if (t !== null) return /* @__PURE__ */ (0, L.jsx)(je, { data: t });
	}
	return /* @__PURE__ */ (0, L.jsx)("pre", {
		className: "whitespace-pre-wrap break-all",
		children: e
	});
}
function hl(e) {
	try {
		return JSON.parse(e);
	} catch {
		return null;
	}
}
function gl(e) {
	return !!e && (e.includes("json") || e.includes("javascript"));
}
function _l(e, t) {
	return [
		t.request.method,
		t.request.url,
		t.request.postData?.text,
		t.response.content?.text,
		t._error
	].some((t) => !!t && t.toLowerCase().includes(e));
}
function vl(e) {
	return e < 0 ? "" : e < 1024 ? `${e}B` : e < 1048576 ? `${(e / 1024).toFixed(1)}KB` : `${(e / 1024 / 1024).toFixed(1)}MB`;
}
function yl(e) {
	return e >= 500 ? "text-red-600" : e >= 400 ? "text-amber-600" : e >= 300 ? "text-blue-600" : e >= 200 ? "text-green-600" : "text-muted-foreground";
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/profile-aggregate.ts
function bl(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.has(n.sampleType) || t.set(n.sampleType, {
		sampleType: n.sampleType,
		unit: n.unit
	});
	return [...t.values()];
}
function xl(e, t) {
	let n = e.filter((e) => e.sampleType === t), r = new Set(n.map((e) => e.unit));
	if (r.size > 1) throw Error(`sample type "${t}" has mixed units ${[...r].map((e) => `"${e}"`).join(", ")}`);
	for (let e of n) if (e.frames.length === 0) throw Error(`a "${t}" sample has no frames`);
	return n;
}
function Sl(e, t) {
	let n = xl(e, t), r = /* @__PURE__ */ new Map(), i = 0;
	for (let e of n) {
		i += e.value;
		for (let t of e.frames) r.has(t.func) || r.set(t.func, {
			func: t.func,
			file: t.file,
			line: t.line,
			flat: 0,
			cum: 0
		});
		for (let t of new Set(e.frames.map((e) => e.func))) r.get(t).cum += e.value;
		r.get(e.frames[0].func).flat += e.value;
	}
	return {
		total: i,
		unit: n[0]?.unit ?? "",
		functions: [...r.values()].sort((e, t) => t.flat - e.flat || t.cum - e.cum || e.func.localeCompare(t.func))
	};
}
function Cl(e, t, n, r) {
	let i = /* @__PURE__ */ new Map();
	for (let r of xl(e, t)) {
		if (!r.frames.some((e) => e.func === n)) continue;
		let e = r.frames.map((e) => `${e.func}@${e.file}:${e.line}`).join("\n"), t = i.get(e);
		t ? t.value += r.value : i.set(e, {
			frames: r.frames,
			value: r.value
		});
	}
	return [...i.values()].sort((e, t) => t.value - e.value).slice(0, r);
}
var wl = 1e6;
function Tl(e, t) {
	switch (t) {
		case "bytes": return m(e);
		case "nanoseconds": return Math.abs(e) < wl ? `${a(e / 1e3)} µs` : ae(e / wl);
		case "microseconds": return Math.abs(e) < 1e3 ? `${a(e)} µs` : ae(e / 1e3);
		case "milliseconds": return ae(e);
		case "seconds": return ae(e, { from: "s" });
		case "count":
		case "": return a(e);
		default: return `${a(e)} ${t}`;
	}
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/process-tree.ts
var El = (e, t) => Date.parse(e.row.startedAt) - Date.parse(t.row.startedAt) || e.row.pid - t.row.pid;
function Dl(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) {
		if (t.has(n.key)) throw Error(`process tree key "${n.key}" appears twice`);
		t.set(n.key, {
			row: n,
			children: [],
			exited: jl(n)
		});
	}
	let n = [];
	for (let e of t.values()) {
		let r = e.row.parentKey ? t.get(e.row.parentKey) : void 0;
		(r ? r.children : n).push(e);
	}
	n.sort((e, t) => Number(!!t.row.isRoot) - Number(!!e.row.isRoot) || El(e, t));
	let r = Ml(n);
	if (r.length !== t.size) {
		let e = [...t.keys()].filter((e) => !r.some((t) => t.row.key === e));
		throw Error(`process tree has a parent cycle among keys ${e.map((e) => `"${e}"`).join(", ")}`);
	}
	for (let e of t.values()) e.children.sort(El);
	return n;
}
function Ol(e) {
	return e.flatMap((e) => {
		let t = Ol(e.children);
		return e.exited && t.length === 0 ? [] : [{
			...e,
			children: t
		}];
	});
}
function kl(e) {
	return Date.parse(e.exitedAt ?? e.lastSeenAt) - Date.parse(e.startedAt);
}
function Al(e) {
	let t = new Date(e);
	if (Number.isNaN(t.getTime())) return e;
	let n = (e) => String(e).padStart(2, "0");
	return `${t.getFullYear()}-${n(t.getMonth() + 1)}-${n(t.getDate())} ${n(t.getHours())}:${n(t.getMinutes())}:${n(t.getSeconds())}`;
}
function jl(e) {
	return e.status === "exited" || !!e.exitedAt;
}
function Ml(e) {
	let t = [], n = (e) => {
		t.push(e), e.children.forEach(n);
	};
	return e.forEach(n), t;
}
function Nl(e) {
	return e == null ? "—" : `${e.toFixed(1)}%`;
}
function Pl(e) {
	return e.flatMap((e) => e.cpuPercent == null ? [] : [e.cpuPercent]);
}
function Fl(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.key, [...t.get(n.key) ?? [], n]);
	return new Map([...t].map(([e, t]) => [e, Pl(t.sort((e, t) => e.tick - t.tick))]));
}
function Il(e, t) {
	let n = e.filter((e) => e.key === t).sort((e, t) => e.tick - t.tick);
	return {
		cpu: Pl(n),
		rss: n.map((e) => e.rssBytes)
	};
}
function Ll(e, t, n) {
	if (e.length < 2) return;
	let r = Math.max(...e);
	return e.map((i, a) => {
		let o = a / (e.length - 1) * t, s = r === 0 ? n : n - i / r * n;
		return `${o.toFixed(1)},${s.toFixed(1)}`;
	}).join(" ");
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/GoProfileView.tsx
var Rl = 20, K = 5;
function q(e) {
	return e.file ? `${e.file}${e.line ? `:${e.line}` : ""}` : "";
}
function J({ value: e, total: t }) {
	return /* @__PURE__ */ (0, L.jsx)("span", {
		className: "text-muted-foreground",
		children: t > 0 ? `${(e / t * 100).toFixed(1)}%` : "—"
	});
}
function Y({ samples: e, sampleType: t, unit: n, func: r, total: i }) {
	let a = Cl(e, t, r, K);
	return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "space-y-density-2",
		"data-profile-stacks": r,
		children: a.map((e, t) => /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "rounded border border-border bg-muted/30 p-density-2",
			children: [/* @__PURE__ */ (0, L.jsxs)("div", {
				className: "mb-1 text-[11px] text-muted-foreground",
				children: [
					Tl(e.value, n),
					" ",
					/* @__PURE__ */ (0, L.jsx)(J, {
						value: e.value,
						total: i
					})
				]
			}), /* @__PURE__ */ (0, L.jsx)("ol", {
				className: "space-y-0.5 font-mono text-[11px] leading-4",
				children: e.frames.map((e, t) => /* @__PURE__ */ (0, L.jsxs)("li", {
					className: P("break-all", e.func === r ? "font-semibold text-foreground" : "text-muted-foreground"),
					children: [e.func, q(e) && /* @__PURE__ */ (0, L.jsx)("span", {
						className: "ml-2 text-[10px] font-normal opacity-80",
						children: q(e)
					})]
				}, t))
			})]
		}, t))
	});
}
function X({ samples: e, topN: t = Rl, className: n }) {
	let r = (0, I.useMemo)(() => bl(e), [e]), [i, a] = (0, I.useState)(null), o = r.find((e) => e.sampleType === i)?.sampleType ?? r[r.length - 1]?.sampleType, [s, c] = (0, I.useState)("flat"), [l, d] = (0, I.useState)(!1), [f, p] = (0, I.useState)(null), m = (0, I.useMemo)(() => o ? Sl(e, o) : void 0, [e, o]);
	if (!o || !m) return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "text-sm text-muted-foreground",
		children: "The profile has no samples"
	});
	let h = s === "flat" ? m.functions : [...m.functions].sort((e, t) => t.cum - e.cum || t.flat - e.flat), g = l ? h : h.slice(0, t), _ = (e, t) => /* @__PURE__ */ (0, L.jsx)("th", {
		scope: "col",
		className: "px-2 py-1 text-right font-medium",
		"aria-sort": s === e ? "descending" : "none",
		children: /* @__PURE__ */ (0, L.jsx)("button", {
			type: "button",
			className: P("hover:text-foreground", s === e && "text-foreground"),
			onClick: () => c(e),
			children: t
		})
	});
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("space-y-density-2", n),
		"data-profile-top": "",
		children: [
			/* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex flex-wrap items-center justify-between gap-density-2 text-xs text-muted-foreground",
				children: [/* @__PURE__ */ (0, L.jsxs)("span", { children: [
					"Total ",
					Tl(m.total, m.unit),
					" across ",
					m.functions.length,
					" functions"
				] }), r.length > 1 && /* @__PURE__ */ (0, L.jsx)("div", {
					className: "min-w-[12rem]",
					children: /* @__PURE__ */ (0, L.jsx)(u, {
						"aria-label": "Sample type",
						value: o,
						onChange: (e) => {
							a(e.target.value), p(null);
						},
						options: r.map((e) => ({
							value: e.sampleType,
							label: `${e.sampleType} (${e.unit})`
						}))
					})
				})]
			}),
			/* @__PURE__ */ (0, L.jsxs)("table", {
				className: "w-full table-fixed border-collapse text-xs",
				children: [/* @__PURE__ */ (0, L.jsx)("thead", {
					className: "text-muted-foreground",
					children: /* @__PURE__ */ (0, L.jsxs)("tr", {
						className: "border-b border-border",
						children: [
							/* @__PURE__ */ (0, L.jsx)("th", {
								scope: "col",
								className: "px-2 py-1 text-left font-medium",
								children: "Function"
							}),
							_("flat", "Flat"),
							/* @__PURE__ */ (0, L.jsx)("th", {
								scope: "col",
								className: "w-16 px-2 py-1 text-right font-medium",
								children: "Flat %"
							}),
							_("cum", "Cum"),
							/* @__PURE__ */ (0, L.jsx)("th", {
								scope: "col",
								className: "w-16 px-2 py-1 text-right font-medium",
								children: "Cum %"
							})
						]
					})
				}), /* @__PURE__ */ (0, L.jsx)("tbody", { children: g.map((t) => {
					let n = f === t.func;
					return /* @__PURE__ */ (0, L.jsxs)(I.Fragment, { children: [/* @__PURE__ */ (0, L.jsxs)("tr", {
						className: "border-b border-border/60 hover:bg-muted/40",
						"data-profile-function": t.func,
						children: [
							/* @__PURE__ */ (0, L.jsx)("td", {
								className: "px-2 py-1",
								children: /* @__PURE__ */ (0, L.jsxs)("button", {
									type: "button",
									"aria-expanded": n,
									className: "flex w-full items-start gap-1 text-left",
									onClick: () => p(n ? null : t.func),
									children: [/* @__PURE__ */ (0, L.jsx)(F, {
										icon: n ? Ze : Ie,
										className: "mt-0.5 shrink-0 text-[11px]"
									}), /* @__PURE__ */ (0, L.jsxs)("span", {
										className: "min-w-0",
										children: [/* @__PURE__ */ (0, L.jsx)("span", {
											className: "block break-all font-mono text-[11px] font-semibold",
											children: t.func
										}), q(t) && /* @__PURE__ */ (0, L.jsx)("span", {
											className: "block break-all text-[10px] text-muted-foreground",
											children: q(t)
										})]
									})]
								})
							}),
							/* @__PURE__ */ (0, L.jsx)("td", {
								className: "px-2 py-1 text-right tabular-nums",
								children: Tl(t.flat, m.unit)
							}),
							/* @__PURE__ */ (0, L.jsx)("td", {
								className: "px-2 py-1 text-right tabular-nums",
								children: /* @__PURE__ */ (0, L.jsx)(J, {
									value: t.flat,
									total: m.total
								})
							}),
							/* @__PURE__ */ (0, L.jsx)("td", {
								className: "px-2 py-1 text-right tabular-nums",
								children: Tl(t.cum, m.unit)
							}),
							/* @__PURE__ */ (0, L.jsx)("td", {
								className: "px-2 py-1 text-right tabular-nums",
								children: /* @__PURE__ */ (0, L.jsx)(J, {
									value: t.cum,
									total: m.total
								})
							})
						]
					}), n && /* @__PURE__ */ (0, L.jsx)("tr", { children: /* @__PURE__ */ (0, L.jsx)("td", {
						colSpan: 5,
						className: "px-2 py-density-2",
						children: /* @__PURE__ */ (0, L.jsx)(Y, {
							samples: e,
							sampleType: o,
							unit: m.unit,
							func: t.func,
							total: m.total
						})
					}) })] }, t.func);
				}) })]
			}),
			!l && h.length > g.length && /* @__PURE__ */ (0, L.jsxs)("button", {
				type: "button",
				className: "text-xs text-primary hover:underline",
				onClick: () => d(!0),
				children: [
					"Show all ",
					h.length,
					" functions"
				]
			})
		]
	});
}
function zl(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.checkpointId, [...t.get(n.checkpointId) ?? [], n]);
	return [...t].map(([e, t]) => {
		let [n] = t;
		return {
			checkpointId: e,
			takenAt: n.takenAt,
			reason: n.reason,
			source: n.source,
			profile: n.profile,
			records: t
		};
	}).sort((e, t) => Date.parse(e.takenAt) - Date.parse(t.takenAt));
}
function Bl({ records: e, processKey: t, topN: n, className: r }) {
	let i = (0, I.useMemo)(() => new Set(e.map((e) => e.key)), [e]);
	if (i.size > 1 && t === void 0) throw Error(`GoProfileView got records of ${i.size} processes; pass processKey to choose one`);
	let a = (0, I.useMemo)(() => t === void 0 ? e : e.filter((e) => e.key === t), [e, t]), o = (0, I.useMemo)(() => zl(a), [a]), [s, c] = (0, I.useState)(null), l = o.find((e) => e.checkpointId === s) ?? o[o.length - 1];
	return l ? /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("space-y-density-2", r),
		"data-go-profile": l.profile,
		children: [o.length > 1 ? /* @__PURE__ */ (0, L.jsx)("div", {
			className: "max-w-md",
			children: /* @__PURE__ */ (0, L.jsx)(u, {
				"aria-label": "Profile checkpoint",
				value: l.checkpointId,
				onChange: (e) => c(e.target.value),
				options: [...o].reverse().map((e) => ({
					value: e.checkpointId,
					label: [
						Al(e.takenAt),
						e.profile,
						e.reason,
						e.source
					].filter(Boolean).join(" · ")
				}))
			})
		}) : /* @__PURE__ */ (0, L.jsx)("div", {
			className: "text-xs text-muted-foreground",
			children: [
				l.profile,
				Al(l.takenAt),
				l.reason,
				l.source
			].filter(Boolean).join(" · ")
		}), /* @__PURE__ */ (0, L.jsx)(X, {
			samples: l.records,
			...n === void 0 ? {} : { topN: n }
		}, l.checkpointId)]
	}) : /* @__PURE__ */ (0, L.jsx)("div", {
		className: "text-sm text-muted-foreground",
		children: "No profile collected yet"
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/stack-checkpoints.ts
function Vl(e, t) {
	let n = /* @__PURE__ */ new Map();
	for (let r of e) r.key === t && n.set(r.checkpointId, [...n.get(r.checkpointId) ?? [], r]);
	return [...n].map(([e, t]) => {
		let [n] = t;
		return {
			checkpointId: e,
			takenAt: n.takenAt,
			reason: n.reason,
			transport: n.transport,
			failed: t.some((e) => e.error),
			records: t
		};
	}).sort((e, t) => Date.parse(e.takenAt) - Date.parse(t.takenAt));
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/ProcessRecordViews.tsx
function Hl(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.key, [...t.get(n.key) ?? [], n]);
	return [...t];
}
function Ul({ processes: e, render: t }) {
	return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "min-h-0 flex-1 space-y-density-3 overflow-auto p-density-2",
		children: e.map(([n, r]) => /* @__PURE__ */ (0, L.jsxs)("div", {
			"data-trace-process": n,
			children: [e.length > 1 && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "mb-1 font-mono text-xs text-muted-foreground",
				children: n
			}), t(n, r)]
		}, n))
	});
}
function Wl({ rows: e }) {
	let t = (0, I.useMemo)(() => Hl(e), [e]);
	return /* @__PURE__ */ (0, L.jsx)(Ul, {
		processes: t,
		render: (e, t) => /* @__PURE__ */ (0, L.jsx)(Bl, {
			records: t,
			processKey: e
		})
	});
}
function Gl({ rows: e }) {
	let t = (0, I.useMemo)(() => Hl(e).map(([t]) => [t, Vl(e, t).reverse()]), [e]);
	return /* @__PURE__ */ (0, L.jsx)(Ul, {
		processes: t,
		render: (e, t) => /* @__PURE__ */ (0, L.jsx)("div", {
			className: "space-y-density-2",
			children: t.map((e) => /* @__PURE__ */ (0, L.jsx)("section", {
				className: "rounded border border-border px-density-2",
				"data-stack-checkpoint": e.checkpointId,
				children: /* @__PURE__ */ (0, L.jsx)(me, {
					records: e.records,
					headerActions: /* @__PURE__ */ (0, L.jsx)("span", {
						className: "font-mono text-[11px] text-muted-foreground",
						children: e.transport
					})
				})
			}, e.checkpointId))
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/Sparkline.tsx
function Kl({ values: e, label: t, width: n = 64, height: r = 16, fill: i = !1, className: a }) {
	let o = Ll(e, n, r);
	return o ? /* @__PURE__ */ (0, L.jsx)("svg", {
		role: "img",
		"aria-label": t,
		viewBox: `0 0 ${n} ${r}`,
		...i ? { preserveAspectRatio: "none" } : {
			width: n,
			height: r
		},
		className: P(i ? "w-full" : "shrink-0", "overflow-visible text-primary", a),
		style: i ? { height: r } : void 0,
		children: /* @__PURE__ */ (0, L.jsx)("polyline", {
			points: o,
			fill: "none",
			stroke: "currentColor",
			strokeWidth: 1.25,
			strokeLinejoin: "round",
			...i ? { vectorEffect: "non-scaling-stroke" } : {}
		})
	}) : null;
}
function ql({ label: e, values: t, format: n, className: r }) {
	let i = Math.max(...t, 0), a = t[0];
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("rounded-lg border border-border bg-muted/50 px-2.5 py-density-2", r),
		"data-sparkline": e,
		children: [/* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex items-center justify-between text-[10px] uppercase tracking-wide text-muted-foreground",
			children: [/* @__PURE__ */ (0, L.jsx)("span", { children: e }), /* @__PURE__ */ (0, L.jsx)("span", { children: t.length > 0 ? `max ${n(i)}` : "" })]
		}), t.length < 2 ? /* @__PURE__ */ (0, L.jsx)("div", {
			className: "mt-1 text-[11px] text-muted-foreground",
			children: a === void 0 ? "No samples" : `${n(a)} (1 sample)`
		}) : /* @__PURE__ */ (0, L.jsx)(Kl, {
			fill: !0,
			values: t,
			label: `${e} over time`,
			width: 160,
			height: 36,
			className: "mt-1"
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/ProcessTraceTree.tsx
var Jl = 2, Yl = () => {};
function Xl({ node: e, selected: t, series: n }) {
	let { row: r, exited: i } = e;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		"data-process-key": r.key,
		...i ? { "data-exited": "true" } : {},
		className: P("flex min-w-0 flex-1 items-center gap-1.5", i && "opacity-60"),
		title: r.command,
		children: [
			/* @__PURE__ */ (0, L.jsx)(F, {
				icon: r.isRoot ? on : Qt,
				className: r.isRoot ? "text-base text-blue-600" : "text-base text-muted-foreground"
			}),
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: P("truncate", t ? "font-semibold text-primary" : "font-medium text-foreground"),
				children: el({
					pid: r.pid,
					name: r.name,
					command: r.command
				})
			}),
			/* @__PURE__ */ (0, L.jsxs)("span", {
				className: "shrink-0 text-xs text-muted-foreground",
				children: ["pid ", r.pid]
			}),
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: "min-w-0 flex-1 truncate font-mono text-xs text-muted-foreground",
				title: r.command,
				"data-process-command": "",
				children: r.command
			}),
			n ? /* @__PURE__ */ (0, L.jsx)(Kl, {
				values: n,
				label: `CPU % over ${n.length} samples`
			}) : null,
			/* @__PURE__ */ (0, L.jsxs)("span", {
				className: P("inline-flex shrink-0 items-center gap-1 text-xs", $c(r.status)),
				children: [/* @__PURE__ */ (0, L.jsx)(F, { icon: Qc(r.status) }), r.status]
			}),
			i ? /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
				/* @__PURE__ */ (0, L.jsx)("span", {
					className: "shrink-0 text-xs tabular-nums text-muted-foreground",
					title: "Lifetime",
					children: ae(kl(r))
				}),
				/* @__PURE__ */ (0, L.jsx)("span", {
					className: "shrink-0 text-xs tabular-nums text-muted-foreground",
					title: "Peak CPU",
					children: `peak ${Nl(r.peakCpuPercent)}`
				}),
				/* @__PURE__ */ (0, L.jsx)("span", {
					className: "shrink-0 text-xs tabular-nums text-muted-foreground",
					title: "Peak RSS",
					children: `peak ${m(r.peakRssBytes)}`
				})
			] }) : /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsx)("span", {
				className: "shrink-0 text-xs tabular-nums text-muted-foreground",
				title: `Peak ${Nl(r.peakCpuPercent)}`,
				children: Nl(r.cpuPercent)
			}), /* @__PURE__ */ (0, L.jsx)("span", {
				className: "shrink-0 text-xs tabular-nums text-muted-foreground",
				title: `Peak ${m(r.peakRssBytes)}`,
				children: m(r.rssBytes)
			})] })
		]
	});
}
function Zl({ nodes: e, samples: t, onSelect: n, selectedKey: r, emptyLabel: i = "No processes recorded", className: a, expandAll: o, onExpandAllChange: s, hideExited: c = !1, onHideExitedChange: l, loading: u = !1, error: d }) {
	let f = (0, I.useMemo)(() => Dl(e), [e]), p = (0, I.useMemo)(() => c ? Ol(f) : f, [f, c]), m = (0, I.useMemo)(() => Fl(t ?? []), [t]), h = (0, I.useMemo)(() => r ? Ml(p).find((e) => e.row.key === r) ?? null : null, [p, r]), g = e.length === 0 ? d ? null : u ? /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "p-density-6 text-center text-sm text-muted-foreground",
		role: "status",
		children: [/* @__PURE__ */ (0, L.jsx)(F, {
			icon: ht,
			className: "animate-spin text-3xl text-blue-500"
		}), /* @__PURE__ */ (0, L.jsx)("p", {
			className: "mt-2",
			children: "Waiting for process trace…"
		})]
	}) : /* @__PURE__ */ (0, L.jsx)("div", {
		className: "p-density-6 text-center text-sm text-muted-foreground",
		children: i
	}) : /* @__PURE__ */ (0, L.jsx)("div", {
		className: "p-density-6 text-center text-sm text-muted-foreground",
		children: "Every process has exited"
	}), _ = o !== void 0 && s === void 0;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("flex min-h-0 flex-col", a),
		children: [
			l && /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex items-center justify-between gap-2 border-b border-border bg-background px-3 py-1.5 text-xs text-muted-foreground",
				"data-process-toolbar": "",
				children: [/* @__PURE__ */ (0, L.jsxs)("span", { children: [
					e.length,
					" ",
					e.length === 1 ? "process" : "processes"
				] }), /* @__PURE__ */ (0, L.jsxs)("label", {
					className: "inline-flex cursor-pointer items-center gap-1.5",
					children: [/* @__PURE__ */ (0, L.jsx)("input", {
						type: "checkbox",
						checked: c,
						onChange: (e) => l(e.target.checked)
					}), "Hide exited"]
				})]
			}),
			d && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "whitespace-pre-wrap px-3 py-2 text-xs text-red-700 dark:text-red-400",
				role: "alert",
				children: d
			}),
			/* @__PURE__ */ (0, L.jsx)($e, {
				roots: p,
				className: "min-h-0 flex-1",
				ariaLabel: "Process tree",
				empty: g,
				showControls: !_,
				...o === void 0 ? {} : {
					expandAll: o,
					onExpandAllChange: s ?? Yl
				},
				...o === void 0 && s ? { onExpandAllChange: s } : {},
				getChildren: (e) => e.children,
				getKey: (e) => e.row.key,
				getSearchText: (e) => `${e.row.name} ${e.row.command} ${e.row.pid}`,
				selected: h,
				revealSelected: !0,
				defaultOpen: (e, t) => t < Jl,
				onSelect: (e) => n(e.row.key),
				renderRow: ({ node: e, selected: t }) => /* @__PURE__ */ (0, L.jsx)(Xl, {
					node: e,
					selected: t,
					series: m.get(e.row.key)
				}),
				rowClass: (e, t) => t ? "bg-primary/10 border-l-2 border-primary" : "hover:bg-accent"
			})
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/sql-trace.ts
function Ql(e) {
	return e >= 100 ? ae(e) : `${Number(e.toFixed(1))} ms`;
}
function $l(e) {
	let t = e.filter((e) => e.rows >= 0);
	return t.length === 0 ? -1 : t.reduce((e, t) => e + t.rows, 0);
}
function eu(e, t) {
	let [n] = t;
	return {
		id: e,
		startedAt: n.startedAt,
		op: n.op,
		tables: n.tables,
		sql: n.sql,
		durationMs: t.reduce((e, t) => e + t.durationMs, 0),
		rows: $l(t),
		error: t.find((e) => e.error)?.error ?? "",
		failed: t.filter((e) => e.error).length,
		count: t.length,
		occurrences: t
	};
}
function tu(e, t) {
	let n = /* @__PURE__ */ new Map();
	return e.forEach((e, r) => {
		let i = t.group && e.repeatKey ? `repeat:${e.repeatKey}` : `row:${r}`;
		n.set(i, [...n.get(i) ?? [], e]);
	}), [...n].map(([e, t]) => eu(e, t));
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/SqlTraceTable.tsx
var nu = [
	{
		key: "startedAt",
		label: "Time",
		kind: "timestamp",
		shrink: !0
	},
	{
		key: "op",
		label: "Op",
		shrink: !0,
		filterable: !0,
		render: (e) => /* @__PURE__ */ (0, L.jsx)(Oe, {
			variant: "soft",
			tone: "neutral",
			size: "xs",
			className: "font-mono",
			children: String(e)
		})
	},
	{
		key: "tables",
		label: "Tables",
		kind: "tags",
		shrink: !0
	},
	{
		key: "sql",
		label: "SQL",
		grow: !0,
		render: (e, t) => /* @__PURE__ */ (0, L.jsxs)("span", {
			className: "flex min-w-0 items-center gap-density-2",
			children: [/* @__PURE__ */ (0, L.jsx)(Ae, {
				inline: !0,
				language: "sql",
				source: t.sql
			}), t.count > 1 ? /* @__PURE__ */ (0, L.jsx)("span", {
				title: `${t.count} executions`,
				children: /* @__PURE__ */ (0, L.jsx)(Oe, {
					variant: "soft",
					tone: "info",
					size: "xs",
					children: `×${t.count}`
				})
			}) : null]
		})
	},
	{
		key: "durationMs",
		label: "Duration",
		align: "right",
		shrink: !0,
		render: (e, t) => /* @__PURE__ */ (0, L.jsx)("span", {
			className: "tabular-nums",
			title: t.count > 1 ? `${t.count} executions in total` : void 0,
			children: Ql(Number(e))
		}),
		sortValue: (e) => Number(e)
	},
	{
		key: "rows",
		label: "Rows",
		align: "right",
		shrink: !0,
		render: (e) => Number(e) < 0 ? /* @__PURE__ */ (0, L.jsx)("span", {
			className: "text-muted-foreground",
			children: "—"
		}) : /* @__PURE__ */ (0, L.jsx)("span", {
			className: "tabular-nums",
			children: String(e)
		}),
		sortValue: (e) => Number(e)
	},
	{
		key: "error",
		label: "Error",
		grow: !0,
		render: (e, t) => e ? /* @__PURE__ */ (0, L.jsxs)("span", {
			className: "truncate text-destructive",
			title: String(e),
			children: [String(e), t.count > 1 ? ` (${t.failed} of ${t.count} failed)` : ""]
		}) : null
	}
];
function ru({ row: e }) {
	let [t] = e.occurrences;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "space-y-density-3",
		children: [
			/* @__PURE__ */ (0, L.jsx)(Ae, {
				language: "sql",
				source: e.sql,
				copyable: !0
			}),
			e.error ? /* @__PURE__ */ (0, L.jsxs)("div", { children: [/* @__PURE__ */ (0, L.jsx)("div", {
				className: "mb-1 font-semibold text-foreground",
				children: "Error"
			}), /* @__PURE__ */ (0, L.jsx)("pre", {
				className: "whitespace-pre-wrap break-all text-destructive",
				children: e.error
			})] }) : null,
			/* @__PURE__ */ (0, L.jsxs)("div", {
				className: "text-xs text-muted-foreground",
				children: [
					t.connection,
					" · ",
					t.driver,
					t.truncated?.length ? ` · truncated: ${t.truncated.join(", ")}` : ""
				]
			}),
			e.count > 1 ? /* @__PURE__ */ (0, L.jsx)("ul", {
				className: "space-y-1 text-xs",
				"aria-label": "Executions",
				children: e.occurrences.map((e, t) => /* @__PURE__ */ (0, L.jsxs)("li", {
					className: P("flex items-baseline gap-density-3", e.error && "text-destructive"),
					children: [
						/* @__PURE__ */ (0, L.jsx)("span", {
							className: "tabular-nums text-muted-foreground",
							children: e.startedAt
						}),
						/* @__PURE__ */ (0, L.jsx)("span", {
							className: "tabular-nums",
							children: Ql(e.durationMs)
						}),
						/* @__PURE__ */ (0, L.jsx)("span", {
							className: "tabular-nums",
							children: e.rows < 0 ? "— rows" : `${e.rows} rows`
						}),
						e.args ? /* @__PURE__ */ (0, L.jsx)("code", {
							className: "font-mono",
							children: JSON.stringify(e.args)
						}) : null,
						e.error ? /* @__PURE__ */ (0, L.jsx)("span", { children: e.error }) : null
					]
				}, t))
			}) : t.args ? /* @__PURE__ */ (0, L.jsxs)("div", { children: [/* @__PURE__ */ (0, L.jsx)("div", {
				className: "mb-1 font-semibold text-foreground",
				children: "Arguments"
			}), /* @__PURE__ */ (0, L.jsx)(je, { data: t.args })] }) : null
		]
	});
}
function iu({ rows: e, groupRepeated: t = !0, emptyLabel: n = "No SQL statements recorded", className: r }) {
	let [i, a] = (0, I.useState)(t), s = (0, I.useMemo)(() => tu(e, { group: i }), [e, i]);
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("flex h-full min-h-0 flex-col gap-density-2", r),
		children: [/* @__PURE__ */ (0, L.jsx)(Zc, {
			checked: i,
			onChange: a,
			label: "Group repeated statements",
			className: "self-start px-density-2"
		}), /* @__PURE__ */ (0, L.jsx)(o, {
			data: s,
			columns: nu,
			autoFilter: !0,
			emptyMessage: n,
			globalFilterPlaceholder: "Filter SQL, table, or error…",
			getRowId: (e) => e.id,
			renderExpandedRow: (e) => /* @__PURE__ */ (0, L.jsx)(ru, { row: e })
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/cast.ts
var au = -1, ou = /* @__PURE__ */ new Set(["i", "m"]);
function su(e, t) {
	try {
		return JSON.parse(e);
	} catch (e) {
		throw Error(`cast ${t} is not valid JSON: ${e.message}`);
	}
}
function cu(e) {
	let t = su(e, "header");
	if (typeof t != "object" || !t || Array.isArray(t)) throw Error("cast header is not a JSON object");
	if (t.version !== 2) throw Error(`cast header: only asciinema version 2 is supported, got ${String(t.version)}`);
	if (typeof t.width != "number" || typeof t.height != "number") throw Error("cast header has no numeric width and height");
	return {
		...t,
		width: t.width,
		height: t.height
	};
}
function lu(e, t) {
	let n = typeof e == "string" ? /^(\d+)x(\d+)$/.exec(e) : null;
	if (!n) throw Error(`cast line ${t}: resize data must be "<width>x<height>", got ${JSON.stringify(e)}`);
	return {
		width: Number(n[1]),
		height: Number(n[2])
	};
}
function uu(e) {
	let [t, ...n] = e.split("\n").filter((e) => e.trim() !== "");
	if (t === void 0) throw Error("cast text is empty");
	let r = cu(t), i = {
		width: r.width,
		height: r.height
	}, a = [], o = 0;
	n.forEach((e, t) => {
		let n = t + 2, r;
		try {
			r = JSON.parse(e);
		} catch (e) {
			throw Error(`cast line ${n} is not valid JSON: ${e.message}`);
		}
		if (!Array.isArray(r) || r.length < 3 || typeof r[0] != "number" || typeof r[1] != "string") throw Error(`cast line ${n} is not a [seconds, type, data] event`);
		let [s, c, l] = r;
		if (o = Math.max(o, s), !ou.has(c)) {
			if (c === "o") a.push({
				offset: s,
				kind: "o",
				data: String(l),
				...i
			});
			else if (c === "r") i = lu(l, n), a.push({
				offset: s,
				kind: "r",
				data: "",
				...i
			});
			else throw Error(`cast line ${n}: unknown event type "${c}"`);
		}
	});
	let s = r._gavel;
	if (!s) return a;
	let c = r.duration ?? o, l = s.snapshots ?? [];
	for (let [e, t] of l) a.push({
		offset: e / 1e3,
		kind: "snapshot",
		data: t,
		...i
	});
	return l.length === 0 && s.final !== void 0 && a.push({
		offset: c,
		kind: "snapshot",
		data: s.final,
		...i
	}), a.push({
		offset: c,
		kind: "exit",
		data: "",
		...i,
		...s.exit_code !== void 0 && s.exit_code !== au ? { exitCode: s.exit_code } : {},
		...s.truncated ? { truncated: !0 } : {},
		...r.command ? { command: r.command } : {}
	}), a;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/terminal-screen.ts
var du = {
	ch: " ",
	sgr: ""
}, fu = 8, pu = 4096, mu = (e, t, n) => Math.min(Math.max(e, t), n);
function hu(e, t) {
	if (!(e > 0) || !(t > 0)) throw Error(`terminal screen needs a positive size, got ${e}x${t}`);
	return {
		width: e,
		height: t,
		rows: Array.from({ length: t }, () => []),
		row: 0,
		col: 0,
		style: {
			bold: !1,
			dim: !1,
			italic: !1,
			underline: !1,
			fg: null
		},
		sgr: "",
		saved: null,
		pending: ""
	};
}
function gu(e, t, n) {
	if (!(t > 0) || !(n > 0)) throw Error(`terminal screen needs a positive size, got ${t}x${n}`);
	if (n < e.rows.length) {
		let t = Math.max(0, e.row - n + 1);
		e.rows.splice(0, t), e.row -= t, e.rows.length = n;
	}
	for (; e.rows.length < n;) e.rows.push([]);
	for (let n of e.rows) n.length > t && (n.length = t);
	e.width = t, e.height = n, e.row = mu(e.row, 0, n - 1), e.col = mu(e.col, 0, t - 1);
}
function _u(e) {
	e.row < e.height - 1 ? e.row++ : (e.rows.shift(), e.rows.push([]));
}
function vu(e, t) {
	e.col >= e.width && (e.col = 0, _u(e));
	let n = e.rows[e.row];
	for (; n.length < e.col;) n.push(du);
	n[e.col] = {
		ch: t,
		sgr: e.sgr
	}, e.col++;
}
function yu(e, t) {
	let n = e.rows[e.row];
	if (t === 0) {
		n.length = Math.min(n.length, e.col);
		return;
	}
	if (t === 2) {
		n.length = 0;
		return;
	}
	let r = Math.min(e.col + 1, e.width);
	for (; n.length < r;) n.push(du);
	for (let e = 0; e < r; e++) n[e] = du;
}
function bu(e, t) {
	if (t === 0) {
		yu(e, 0);
		for (let t = e.row + 1; t < e.height; t++) e.rows[t] = [];
	} else if (t === 1) {
		for (let t = 0; t < e.row; t++) e.rows[t] = [];
		yu(e, 1);
	} else for (let t = 0; t < e.height; t++) e.rows[t] = [];
}
function xu(e, t) {
	let n = e.style, r = t.length === 0 ? [""] : t;
	for (let e = 0; e < r.length; e++) {
		let t = r[e];
		if (t.includes(":")) continue;
		let i = t === "" ? 0 : Number.parseInt(t, 10);
		if (i === 0) n.bold = n.dim = n.italic = n.underline = !1, n.fg = null;
		else if (i === 1) n.bold = !0;
		else if (i === 2) n.dim = !0;
		else if (i === 3) n.italic = !0;
		else if (i === 4) n.underline = !0;
		else if (i === 22) n.bold = n.dim = !1;
		else if (i === 23) n.italic = !1;
		else if (i === 24) n.underline = !1;
		else if (i >= 30 && i <= 37 || i >= 90 && i <= 97) n.fg = i;
		else if (i === 39) n.fg = null;
		else if (i === 38 || i === 48 || i === 58) {
			let t = r[e + 1];
			e += t === "5" ? 2 : t === "2" ? 4 : 0;
		}
	}
	e.sgr = [
		n.bold ? "1" : "",
		n.dim ? "2" : "",
		n.italic ? "3" : "",
		n.underline ? "4" : "",
		n.fg === null ? "" : String(n.fg)
	].filter(Boolean).join(";");
}
function Su(e, t, n) {
	if (/^[?>=<!]/.test(t)) return;
	let r = t === "" ? [] : t.split(";"), i = (e, t) => {
		let n = Number.parseInt(r[e] ?? "", 10);
		return Number.isNaN(n) || n === 0 ? t : n;
	}, a = Number.parseInt(r[0] ?? "0", 10) || 0, o = e.rows[e.row];
	switch (n) {
		case "m":
			xu(e, r);
			break;
		case "A":
			e.row = Math.max(0, e.row - i(0, 1));
			break;
		case "B":
			e.row = Math.min(e.height - 1, e.row + i(0, 1));
			break;
		case "C":
			e.col = Math.min(e.width - 1, e.col + i(0, 1));
			break;
		case "D":
			e.col = Math.max(0, e.col - i(0, 1));
			break;
		case "E":
			e.row = Math.min(e.height - 1, e.row + i(0, 1)), e.col = 0;
			break;
		case "F":
			e.row = Math.max(0, e.row - i(0, 1)), e.col = 0;
			break;
		case "G":
			e.col = mu(i(0, 1) - 1, 0, e.width - 1);
			break;
		case "d":
			e.row = mu(i(0, 1) - 1, 0, e.height - 1);
			break;
		case "H":
		case "f":
			e.row = mu(i(0, 1) - 1, 0, e.height - 1), e.col = mu(i(1, 1) - 1, 0, e.width - 1);
			break;
		case "J":
			bu(e, a);
			break;
		case "K":
			yu(e, a);
			break;
		case "P":
			o.splice(e.col, i(0, 1));
			break;
		case "X":
			for (let t = 0; t < i(0, 1) && e.col + t < o.length; t++) o[e.col + t] = du;
			break;
		case "@":
			o.length > e.col && o.splice(e.col, 0, ...Array.from({ length: i(0, 1) }, () => du)), o.length = Math.min(o.length, e.width);
			break;
		case "L":
			e.rows.splice(e.row, 0, ...Array.from({ length: i(0, 1) }, () => [])), e.rows.length = e.height;
			break;
		case "M":
			for (e.rows.splice(e.row, i(0, 1)); e.rows.length < e.height;) e.rows.push([]);
			break;
		case "s":
			e.saved = {
				row: e.row,
				col: e.col
			};
			break;
		case "u": e.saved && ({row: e.row, col: e.col} = e.saved);
	}
}
function Cu(e, t, n) {
	let r = t[n + 1];
	if (r === void 0) return -1;
	if (r === "[") {
		let r = n + 2;
		for (; r < t.length && /[\x20-\x3f]/.test(t[r]);) r++;
		return r >= t.length ? -1 : (Su(e, t.slice(n + 2, r).replace(/[\x20-\x2f]/g, ""), t[r]), r + 1);
	}
	if (r === "]") {
		for (let e = n + 2; e < t.length; e++) {
			if (t[e] === "\x07") return e + 1;
			if (t[e] === "\x1B" && t[e + 1] === "\\") return e + 2;
		}
		return -1;
	}
	return "()*+".includes(r) ? n + 3 > t.length ? -1 : n + 3 : (r === "7" ? e.saved = {
		row: e.row,
		col: e.col
	} : r === "8" && e.saved ? {row: e.row, col: e.col} = e.saved : r === "M" && (e.row = Math.max(0, e.row - 1)), n + 2);
}
function wu(e, t) {
	let n = e.pending + t;
	e.pending = "";
	let r = 0;
	for (; r < n.length;) {
		let t = n[r];
		if (t === "\x1B") {
			let t = Cu(e, n, r);
			if (t === -1) {
				if (n.length - r <= pu) {
					e.pending = n.slice(r);
					return;
				}
				r++;
				continue;
			}
			r = t;
		} else if (r++, t === "\r") e.col = 0;
		else if (t === "\n") _u(e);
		else if (t === "\b") e.col = Math.max(0, e.col - 1);
		else if (t === "	") e.col = Math.min(e.width - 1, (Math.floor(e.col / fu) + 1) * fu);
		else if (t < " " || t === "") continue;
		else {
			let t = n.codePointAt(r - 1), i = String.fromCodePoint(t);
			r += i.length - 1, vu(e, i);
		}
	}
}
function Tu(e) {
	let t = e.length;
	for (; t > 0 && e[t - 1].ch === " ";) t--;
	return e.slice(0, t);
}
function Eu(e, t) {
	return t === "" ? "\x1B[0m" : e === "" ? `\u001b[${t}m` : `\u001b[0;${t}m`;
}
function Du(e) {
	let t = "";
	return e.rows.map((e) => {
		let n = "";
		for (let r of Tu(e)) r.sgr !== t && (n += Eu(t, r.sgr), t = r.sgr), n += r.ch;
		return n;
	}).join("\n");
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/terminal-timeline.ts
var Ou = [
	"o",
	"r",
	"snapshot",
	"exit"
], ku = {
	width: 80,
	height: 24
};
function Au(e) {
	e.forEach((e, t) => {
		if (!Ou.includes(e.kind)) throw Error(`terminal event ${t} has unknown kind "${e.kind}"`);
	});
	let t = e.map((e, t) => ({
		event: e,
		index: t
	})).sort((e, t) => e.event.offset - t.event.offset || e.index - t.index).map(({ event: e }) => e), n = t[0];
	if (n && !(n.width > 0 && n.height > 0)) throw Error(`terminal events carry no terminal size (first row is ${n.width}x${n.height})`);
	return {
		steps: t.filter((e) => e.kind === "o" || e.kind === "r"),
		settled: t.findLast((e) => e.kind === "snapshot")?.data,
		exit: t.findLast((e) => e.kind === "exit"),
		duration: t.reduce((e, t) => Math.max(e, t.offset), 0),
		width: n?.width ?? ku.width,
		height: n?.height ?? ku.height
	};
}
function ju(e, t) {
	let n = 0, r = e.length;
	for (; n < r;) {
		let i = n + r >>> 1;
		e[i].offset <= t ? n = i + 1 : r = i;
	}
	return n;
}
function Mu(e, t, n) {
	let r = e !== void 0 && e.applied <= n, i = r ? e.screen : hu(t.width, t.height);
	for (let a = r ? e.applied : 0; a < n; a++) {
		let e = t.steps[a];
		e.kind === "r" ? gu(i, e.width, e.height) : wu(i, e.data);
	}
	return r && e.applied === n ? e : {
		screen: i,
		applied: n
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/TerminalCastPlayer.tsx
var Nu = 50, Pu = [
	"1",
	"2",
	"4"
].map((e) => ({
	id: e,
	label: `${e}x`
})), Fu = [{
	id: "replay",
	label: "Replay"
}, {
	id: "settled",
	label: "Settled screen"
}];
function Iu(e, t) {
	if (e === void 0 == (t === void 0)) throw Error("TerminalCastPlayer takes either events or castText, not both or neither");
	return e ?? uu(t);
}
function Lu(e) {
	return `${e.toFixed(1)}s`;
}
function Ru({ exit: e }) {
	return e.exitCode === void 0 ? /* @__PURE__ */ (0, L.jsx)(Oe, {
		variant: "outline",
		tone: "neutral",
		size: "sm",
		children: "exited"
	}) : /* @__PURE__ */ (0, L.jsx)(Oe, {
		variant: "soft",
		tone: e.exitCode === 0 ? "success" : "danger",
		size: "sm",
		className: "font-mono",
		children: `exit ${e.exitCode}`
	});
}
function zu(e, t) {
	let n = (0, I.useRef)(void 0), r = ju(e.steps, t);
	return (0, I.useMemo)(() => {
		let t = Mu(n.current?.timeline === e ? n.current.replay : void 0, e, r);
		return n.current = {
			timeline: e,
			replay: t
		}, {
			ansi: Du(t.screen),
			cols: t.screen.width,
			rows: t.screen.height
		};
	}, [e, r]);
}
function Bu({ events: e, castText: t, autoPlay: n = !1, emptyLabel: r = "No terminal output recorded", className: i }) {
	let a = (0, I.useMemo)(() => Au(Iu(e, t)), [e, t]), [o, s] = (0, I.useState)(0), [c, l] = (0, I.useState)(n), [u, d] = (0, I.useState)("1"), [f, p] = (0, I.useState)("replay"), m = zu(a, o);
	if ((0, I.useEffect)(() => {
		if (!c) return;
		let e = Date.now(), t = setInterval(() => {
			let t = Date.now(), n = (t - e) / 1e3;
			e = t, s((e) => Math.min(a.duration, e + n * Number(u)));
		}, Nu);
		return () => clearInterval(t);
	}, [
		c,
		u,
		a.duration
	]), (0, I.useEffect)(() => {
		c && o >= a.duration && l(!1);
	}, [
		c,
		o,
		a.duration
	]), a.steps.length === 0 && a.settled === void 0) return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "p-density-6 text-center text-muted-foreground text-sm",
		children: r
	});
	let h = a.settled !== void 0, g = f === "settled" && h;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("flex min-h-0 flex-col gap-density-2", i),
		children: [
			/* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex flex-wrap items-center gap-density-2 text-xs text-muted-foreground",
				children: [
					a.exit?.command ? /* @__PURE__ */ (0, L.jsx)("code", {
						className: "font-mono text-foreground",
						children: a.exit.command
					}) : null,
					a.exit ? /* @__PURE__ */ (0, L.jsx)(Ru, { exit: a.exit }) : null,
					a.exit?.truncated ? /* @__PURE__ */ (0, L.jsx)("span", {
						title: "Size caps or a full buffer stopped some output events being recorded",
						children: /* @__PURE__ */ (0, L.jsx)(Oe, {
							variant: "soft",
							tone: "warning",
							size: "sm",
							children: "output truncated"
						})
					}) : null,
					/* @__PURE__ */ (0, L.jsx)("span", {
						className: "tabular-nums",
						children: `${a.width}x${a.height}`
					}),
					/* @__PURE__ */ (0, L.jsx)("span", { className: "flex-1" }),
					h && a.steps.length > 0 ? /* @__PURE__ */ (0, L.jsx)(we, {
						size: "sm",
						"aria-label": "Screen view",
						value: f,
						options: Fu,
						onChange: p
					}) : null
				]
			}),
			/* @__PURE__ */ (0, L.jsx)("div", {
				role: "region",
				"aria-label": "Terminal screen",
				"data-cols": g ? void 0 : m.cols,
				className: "overflow-auto rounded border bg-black p-density-2",
				children: /* @__PURE__ */ (0, L.jsx)(ne, {
					text: g ? a.settled : m.ansi,
					className: "font-mono text-xs leading-5 whitespace-pre text-gray-100"
				})
			}),
			a.steps.length > 0 && !g ? /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex items-center gap-density-2",
				children: [
					/* @__PURE__ */ (0, L.jsx)(be, {
						variant: "ghost",
						size: "icon",
						"aria-label": c ? "Pause" : "Play",
						onClick: c ? () => l(!1) : () => {
							o >= a.duration && s(0), l(!0);
						},
						children: /* @__PURE__ */ (0, L.jsx)(F, { icon: c ? nn : tn })
					}),
					/* @__PURE__ */ (0, L.jsx)("input", {
						type: "range",
						"aria-label": "Seek",
						className: "min-w-0 flex-1 accent-primary",
						min: 0,
						max: a.duration,
						step: "any",
						value: o,
						onChange: (e) => s(Number(e.target.value))
					}),
					/* @__PURE__ */ (0, L.jsx)("span", {
						className: "shrink-0 text-xs tabular-nums text-muted-foreground",
						children: `${Lu(o)} / ${Lu(a.duration)}`
					}),
					/* @__PURE__ */ (0, L.jsx)(we, {
						size: "sm",
						"aria-label": "Playback speed",
						value: u,
						options: Pu,
						onChange: d
					}),
					/* @__PURE__ */ (0, L.jsx)(be, {
						variant: "ghost",
						size: "icon",
						"aria-label": "Jump to end",
						onClick: () => {
							l(!1), s(a.duration);
						},
						children: /* @__PURE__ */ (0, L.jsx)(F, { icon: Qe })
					})
				]
			}) : null
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/har-adapter.ts
var Vu = "request.postData.text", Hu = "response.content.text";
function Uu(e) {
	if (e != null) return typeof e == "string" ? e : JSON.stringify(e, null, 2);
}
function Wu(e, t) {
	return t || (e != null && typeof e != "string" ? "application/json" : void 0);
}
function Z(e) {
	return e ? { headers: e } : {};
}
function Gu(e, t, n) {
	if (typeof e != "object" || !e) throw Error(`HTTP exchange ${n.method ?? "?"} ${n.url ?? "?"} has no ${t} object`);
	return e;
}
function Ku(e, t) {
	let n = Uu(e.postData?.text);
	return {
		method: e.method,
		url: e.url,
		...e.httpVersion ? { httpVersion: e.httpVersion } : {},
		...Z(e.headers),
		...e.queryString ? { queryString: e.queryString } : {},
		...n === void 0 ? {} : { postData: {
			mimeType: Wu(e.postData?.text, e.postData?.mimeType) ?? "",
			text: n
		} },
		...e.headersSize === void 0 ? {} : { headersSize: e.headersSize },
		...t === void 0 ? {} : { bodySize: t }
	};
}
function qu(e, t, n, r, i) {
	return {
		status: e.status,
		...e.statusText ? { statusText: e.statusText } : {},
		...e.httpVersion ? { httpVersion: e.httpVersion } : {},
		...Z(e.headers),
		content: {
			size: e.content?.size ?? r,
			...n ? { mimeType: n } : {},
			...t === void 0 ? {} : { text: t },
			...i ? { truncated: i } : {}
		},
		...e.redirectURL ? { redirectURL: e.redirectURL } : {},
		...e.headersSize === void 0 ? {} : { headersSize: e.headersSize },
		bodySize: r
	};
}
function Ju(e) {
	return `Tunnelled CONNECT relayed without decrypting it: ${m(e.bytesIn ?? 0)} in from the server, ${m(e.bytesOut ?? 0)} out from the client.`;
}
function Yu(e, t = {}) {
	let n = Gu(e.request, "request", e), r = Gu(e.response, "response", e), i = new Set(e.truncated ?? []), a = e.tunnelled === !0, o = a ? Ju(e) : Uu(r.content?.text), s = a ? "text/plain" : Wu(r.content?.text, r.content?.mimeType), c = a ? e.bytesIn ?? 0 : r.bodySize ?? r.content?.size ?? 0;
	return {
		startedDateTime: e.startedAt,
		time: e.durationMs,
		request: Ku(n, a ? e.bytesOut ?? 0 : n.bodySize),
		response: qu(r, o, s, c, i.has(Hu)),
		...e.timings ? { timings: e.timings } : {},
		...t.id ? { _id: t.id } : {},
		...e.error ? { _error: e.error } : {},
		...e.serverIPAddress ? { serverIPAddress: e.serverIPAddress } : {},
		...a ? {
			_tunnelled: !0,
			_bytesIn: e.bytesIn ?? 0,
			_bytesOut: e.bytesOut ?? 0
		} : {},
		...e.comment ? { _comment: e.comment } : {},
		...e.encoding ? { _encoding: e.encoding } : {},
		...i.has(Vu) ? { _requestBodyTruncated: !0 } : {}
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/TraceHarView.tsx
function Xu({ rows: e, emptyLabel: t = "No HTTP exchanges recorded", ...n }) {
	let r = (0, I.useMemo)(() => e.map((e, t) => Yu(e, { id: `${e.startedAt}#${t}` })), [e]);
	return /* @__PURE__ */ (0, L.jsx)(al, {
		entries: r,
		emptyLabel: t,
		...n
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/TraceView.tsx
function Zu(e, t, n) {
	t.forEach((t, r) => {
		if (t[n] === void 0 || t[n] === null) throw Error(`trace kind "${e}" row ${r} has no "${n}"`);
	});
}
function Qu({ kind: e, rows: t }) {
	let n = (0, I.useMemo)(() => b(t, { literalKeys: !0 }), [t]);
	return /* @__PURE__ */ (0, L.jsx)(o, {
		data: t,
		columns: n,
		autoFilter: !0,
		emptyMessage: `No ${e} records`
	});
}
function $u({ rows: e, samples: t }) {
	let [n, r] = (0, I.useState)(null);
	return /* @__PURE__ */ (0, L.jsx)(Zl, {
		nodes: e,
		selectedKey: n,
		onSelect: r,
		...t ? { samples: t } : {}
	});
}
function ed({ kind: e, rows: t, samples: n, className: r }) {
	let i = (() => {
		switch (e) {
			case "http":
			case "http_proxy": return Zu(e, t, "request"), Zu(e, t, "response"), /* @__PURE__ */ (0, L.jsx)(Xu, { rows: t });
			case "pg_proxy":
			case "sql_statements": return Zu(e, t, "sql"), /* @__PURE__ */ (0, L.jsx)(iu, { rows: t });
			case "terminal": return Zu(e, t, "kind"), Zu(e, t, "offset"), /* @__PURE__ */ (0, L.jsx)(Bu, { events: t });
			case "process_tree": return Zu(e, t, "key"), n && Zu("process_samples", n, "key"), /* @__PURE__ */ (0, L.jsx)($u, {
				rows: t,
				samples: n
			});
			case "goprof": return Zu(e, t, "key"), Zu(e, t, "checkpointId"), /* @__PURE__ */ (0, L.jsx)(Wl, { rows: t });
			case "stacks": return Zu(e, t, "key"), Zu(e, t, "checkpointId"), /* @__PURE__ */ (0, L.jsx)(Gl, { rows: t });
			default: return /* @__PURE__ */ (0, L.jsx)(Qu, {
				kind: e,
				rows: t
			});
		}
	})();
	return /* @__PURE__ */ (0, L.jsx)("div", {
		className: P("flex h-full min-h-0 flex-col", r),
		children: i
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/process-facets.ts
var td = {
	openFiles: {
		label: "Open files",
		countKey: "openFiles",
		listKey: "openFileList",
		empty: "No open files"
	},
	sockets: {
		label: "Sockets",
		countKey: "sockets",
		listKey: "socketList",
		empty: "No open sockets"
	},
	ports: {
		label: "Ports",
		countKey: "ports",
		listKey: "portList",
		empty: "Not listening on any port"
	},
	connections: {
		label: "Connections",
		countKey: "connections",
		listKey: "connectionList",
		empty: "No network connections"
	}
}, nd = [
	"openFiles",
	"sockets",
	"ports",
	"connections"
], rd = (e, t) => e.filter((e) => e.key === t).sort((e, t) => e.tick - t.tick);
function id(e) {
	return nd.filter((t) => e.some((e) => e[td[t].countKey] !== void 0));
}
function ad(e, t, n) {
	let { countKey: r } = td[n];
	return rd(e, t).flatMap((e) => e[r] === void 0 ? [] : [e[r]]);
}
function od(e, t, n) {
	let { countKey: r, listKey: i } = td[n], a;
	for (let n of rd(e, t)) {
		let e = n[i];
		e ? a = {
			rows: e,
			readAt: n.sampledAt
		} : n[r] === 0 && (!a || a.rows.length > 0) && (a = {
			rows: [],
			readAt: n.sampledAt
		});
	}
	return a;
}
function sd(e, t) {
	let n = rd(e, t);
	return {
		read: n.flatMap((e) => e.diskReadBytes === void 0 ? [] : [e.diskReadBytes]),
		write: n.flatMap((e) => e.diskWriteBytes === void 0 ? [] : [e.diskWriteBytes])
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/ProcessFacetTable.tsx
var cd = [
	{
		key: "fd",
		label: "FD",
		sortable: !0
	},
	{
		key: "protocol",
		label: "Protocol",
		sortable: !0,
		filterable: !0,
		accessor: (e) => e.protocol || e.type
	},
	{
		key: "state",
		label: "State",
		sortable: !0,
		filterable: !0
	},
	{
		key: "localAddr",
		label: "Local",
		sortable: !0,
		grow: !0,
		accessor: (e) => e.localAddr || e.name
	},
	{
		key: "remoteAddr",
		label: "Remote",
		sortable: !0
	}
], ld = {
	openFiles: [
		{
			key: "fd",
			label: "FD",
			sortable: !0
		},
		{
			key: "type",
			label: "Type",
			sortable: !0,
			filterable: !0
		},
		{
			key: "name",
			label: "Path",
			sortable: !0,
			grow: !0
		}
	],
	sockets: cd,
	connections: cd,
	ports: [
		{
			key: "port",
			label: "Port",
			sortable: !0
		},
		{
			key: "protocol",
			label: "Protocol",
			sortable: !0,
			filterable: !0
		},
		{
			key: "localAddr",
			label: "Address",
			sortable: !0,
			grow: !0
		},
		{
			key: "fd",
			label: "FD",
			sortable: !0
		}
	]
}, ud = [{
	key: "name",
	label: "Name",
	sortable: !0
}, {
	key: "value",
	label: "Value",
	grow: !0
}], dd = {
	autoFilter: !0,
	showThemeControl: !1,
	showFullscreenControl: !1
};
function fd(e) {
	if (e.facet === "env") {
		let t = Object.entries(e.environment ?? {}).map(([e, t]) => ({
			name: e,
			value: t
		})).sort((e, t) => e.name.localeCompare(t.name));
		return /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "space-y-2",
			"data-process-facet": "env",
			children: [e.error && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "whitespace-pre-wrap text-xs text-red-700 dark:text-red-400",
				role: "alert",
				children: e.error
			}), e.environment && /* @__PURE__ */ (0, L.jsx)(o, {
				data: t,
				columns: ud,
				getRowId: (e) => e.name,
				emptyMessage: "The process has no environment variables",
				...dd
			})]
		});
	}
	let { reading: t, facet: n } = e;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "space-y-2",
		"data-process-facet": n,
		children: [/* @__PURE__ */ (0, L.jsx)("div", {
			className: "text-xs text-muted-foreground",
			children: t ? `Read at ${Al(t.readAt)}` : "No reading stored for this process yet"
		}), t && /* @__PURE__ */ (0, L.jsx)(o, {
			data: t.rows,
			columns: ld[n],
			getRowId: (e, t) => `${e.fd}:${t}`,
			emptyMessage: td[n].empty,
			...dd
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/ProcessTraceDetail.tsx
var pd = {
	inprocess: {
		label: (e) => e ? "Refresh stack trace" : "Collect stack trace",
		title: "Take a fresh observation and dump the goroutine stacks"
	},
	native: {
		label: () => "Native sample",
		title: "Sample the process's native threads without interrupting it"
	},
	pprof: {
		label: () => "Collect via pprof",
		title: "Read the goroutine profile from the process's pprof endpoint"
	},
	gops: {
		label: () => "Collect via gops",
		title: "Read the goroutine stacks from the process's gops agent"
	},
	sigquit: {
		label: () => "Dump goroutines & kill",
		title: "Send SIGQUIT: the process prints every goroutine's stack and exits",
		warning: (e) => `This sends SIGQUIT to pid ${e}. The process prints every goroutine's stack and exits.`
	}
};
function md(e) {
	let t = pd[e];
	if (!t) throw Error(`unknown stack transport "${e}"; expected one of ${Object.keys(pd).join(", ")}`);
	return t;
}
var hd = [];
function gd({ title: e, children: t }) {
	return /* @__PURE__ */ (0, L.jsxs)("section", { children: [/* @__PURE__ */ (0, L.jsx)("div", {
		className: "mb-1.5 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground",
		children: e
	}), t] });
}
function _d({ label: e, value: t, wide: n }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: `rounded-lg border border-border bg-muted/50 px-2.5 py-density-2 ${n ? "col-span-2" : ""}`,
		children: [/* @__PURE__ */ (0, L.jsx)("div", {
			className: "text-[10px] uppercase tracking-wide text-muted-foreground",
			children: e
		}), /* @__PURE__ */ (0, L.jsx)("div", {
			className: "mt-0.5 break-all text-xs font-medium text-foreground",
			children: t
		})]
	});
}
function vd({ node: e }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "min-w-0",
		children: [/* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex items-center gap-density-2",
			children: [/* @__PURE__ */ (0, L.jsx)(F, {
				icon: e.isRoot ? on : Qt,
				className: "shrink-0 text-2xl text-blue-600"
			}), /* @__PURE__ */ (0, L.jsx)("h2", {
				className: "break-words text-lg font-bold text-foreground",
				children: el(e)
			})]
		}), /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "mt-1 flex flex-wrap items-center gap-density-2 text-xs text-muted-foreground",
			children: [
				/* @__PURE__ */ (0, L.jsxs)("span", {
					className: "font-mono",
					children: ["pid ", e.pid]
				}),
				e.ppid ? /* @__PURE__ */ (0, L.jsxs)("span", {
					className: "font-mono",
					children: ["ppid ", e.ppid]
				}) : null,
				/* @__PURE__ */ (0, L.jsxs)("span", {
					className: `inline-flex items-center gap-1 rounded-full bg-muted px-2 py-0.5 ${$c(e.status)}`,
					children: [/* @__PURE__ */ (0, L.jsx)(F, { icon: Qc(e.status) }), e.status]
				})
			]
		})]
	});
}
function yd({ transports: e, hasStack: t, collecting: n, onCollect: r }) {
	return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "flex items-center gap-density-1",
		"data-stack-actions": "",
		children: e.map((e) => {
			let i = md(e.transport);
			return /* @__PURE__ */ (0, L.jsxs)(be, {
				size: "sm",
				variant: e.destructive ? "destructive" : t ? "outline" : "default",
				disabled: n,
				title: i.title,
				onClick: () => r(e),
				children: [/* @__PURE__ */ (0, L.jsx)(F, {
					icon: n ? ht : Xt,
					className: n ? "mr-1 animate-spin" : "mr-1"
				}), i.label(t)]
			}, e.transport);
		})
	});
}
function Q({ node: e, stacks: t, onCollectStack: n, collecting: r, collectError: i }) {
	let [a, o] = (0, I.useState)(null), [s, c] = (0, I.useState)(null), l = (0, I.useMemo)(() => Vl(t, e.key), [t, e.key]), d = e.stackTransports ?? [], f = !!n && d.length > 0;
	if (l.length === 0 && !f) return null;
	let p = l.find((e) => e.checkpointId === s) ?? l[l.length - 1], m = f ? /* @__PURE__ */ (0, L.jsx)(yd, {
		transports: d,
		hasStack: !!p,
		collecting: !!r,
		onCollect: (t) => t.destructive ? o(t) : void n?.(e.key, t.transport)
	}) : null, h = a ? md(a.transport) : null;
	return /* @__PURE__ */ (0, L.jsxs)(gd, {
		title: "Stack",
		children: [
			i && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "mb-2 whitespace-pre-wrap text-xs text-red-700 dark:text-red-400",
				role: "alert",
				children: i
			}),
			p ? /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "space-y-density-1",
				children: [l.length > 1 && /* @__PURE__ */ (0, L.jsx)("div", {
					className: "max-w-md",
					children: /* @__PURE__ */ (0, L.jsx)(u, {
						"aria-label": "Stack checkpoint",
						value: p.checkpointId,
						onChange: (e) => c(e.target.value),
						options: [...l].reverse().map((e) => ({
							value: e.checkpointId,
							label: [
								Al(e.takenAt),
								e.reason,
								e.transport,
								e.failed ? "failed" : ""
							].filter(Boolean).join(" · ")
						}))
					})
				}), /* @__PURE__ */ (0, L.jsx)(me, {
					records: p.records,
					...m ? { headerActions: m } : {}
				}, `${e.key}:${p.checkpointId}`)]
			}) : /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex items-center justify-between gap-3",
				children: [/* @__PURE__ */ (0, L.jsx)("div", {
					className: "text-xs text-muted-foreground",
					children: r ? "Collecting stack trace..." : "No stack trace collected yet."
				}), m]
			}),
			/* @__PURE__ */ (0, L.jsx)(Le, {
				open: a !== null,
				title: `${h?.label(!1) ?? ""}?`,
				warning: h?.warning?.(e.pid) ?? `Collecting this stack over ${a?.transport ?? ""} terminates pid ${e.pid}.`,
				message: "The stack is shown here once the dump arrives. Use a non-destructive transport to look without killing.",
				confirmLabel: h?.label(!1) ?? "Collect",
				variant: "destructive",
				onCancel: () => o(null),
				onConfirm: () => {
					let t = a;
					o(null), t && n?.(e.key, t.transport);
				}
			})
		]
	});
}
function bd({ node: e, samples: t, links: n, ...r }) {
	let i = jl(e), a = Il(t, e.key), o = sd(t, e.key), s = {
		read: o.read[o.read.length - 1],
		write: o.write[o.write.length - 1]
	};
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "space-y-density-3",
		children: [
			n,
			/* @__PURE__ */ (0, L.jsx)(gd, {
				title: "Command",
				children: /* @__PURE__ */ (0, L.jsx)("pre", {
					className: "whitespace-pre-wrap break-all rounded bg-blue-50 p-2 font-mono text-xs text-foreground dark:bg-blue-500/10",
					children: e.command
				})
			}),
			/* @__PURE__ */ (0, L.jsx)(gd, {
				title: "Process",
				children: /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "grid grid-cols-2 gap-density-2 text-sm",
					children: [
						/* @__PURE__ */ (0, L.jsx)(_d, {
							label: "Cwd",
							value: e.cwd || "Unavailable",
							wide: !0
						}),
						/* @__PURE__ */ (0, L.jsx)(_d, {
							label: "Started",
							value: Al(e.startedAt)
						}),
						/* @__PURE__ */ (0, L.jsx)(_d, {
							label: "Exited",
							value: i ? Al(e.exitedAt ?? e.lastSeenAt) : "running"
						}),
						/* @__PURE__ */ (0, L.jsx)(_d, {
							label: "Lifetime",
							value: `${ae(kl(e))}${i ? "" : " so far"}`
						})
					]
				})
			}),
			/* @__PURE__ */ (0, L.jsxs)(gd, {
				title: "Resources",
				children: [/* @__PURE__ */ (0, L.jsxs)("div", {
					className: "grid grid-cols-2 gap-density-1 sm:grid-cols-4",
					children: [
						/* @__PURE__ */ (0, L.jsx)(M, {
							label: "CPU",
							value: Nl(e.cpuPercent),
							barPct: Math.min(100, e.cpuPercent ?? 0),
							tone: de(e.cpuPercent ?? 0)
						}),
						/* @__PURE__ */ (0, L.jsx)(M, {
							label: "RSS",
							value: m(e.rssBytes)
						}),
						/* @__PURE__ */ (0, L.jsx)(M, {
							label: "Peak CPU",
							value: Nl(e.peakCpuPercent),
							tone: de(e.peakCpuPercent ?? 0)
						}),
						/* @__PURE__ */ (0, L.jsx)(M, {
							label: "Peak RSS",
							value: m(e.peakRssBytes)
						}),
						s.read !== void 0 && /* @__PURE__ */ (0, L.jsx)(M, {
							label: "Disk read",
							value: m(s.read)
						}),
						s.write !== void 0 && /* @__PURE__ */ (0, L.jsx)(M, {
							label: "Disk write",
							value: m(s.write)
						})
					]
				}), /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "mt-2 grid grid-cols-2 gap-density-2",
					children: [
						/* @__PURE__ */ (0, L.jsx)(ql, {
							label: "CPU %",
							values: a.cpu,
							format: (e) => `${e.toFixed(1)}%`
						}),
						/* @__PURE__ */ (0, L.jsx)(ql, {
							label: "RSS",
							values: a.rss,
							format: m
						}),
						o.read.length > 0 && /* @__PURE__ */ (0, L.jsx)(ql, {
							label: "Disk read",
							values: o.read,
							format: m
						}),
						o.write.length > 0 && /* @__PURE__ */ (0, L.jsx)(ql, {
							label: "Disk write",
							values: o.write,
							format: m
						})
					]
				})]
			}),
			/* @__PURE__ */ (0, L.jsx)(Q, {
				node: e,
				...r
			}, e.key)
		]
	});
}
function xd({ node: e, samples: t, facet: n }) {
	let { label: r } = td[n];
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "space-y-density-3",
		children: [/* @__PURE__ */ (0, L.jsx)(ql, {
			label: r,
			values: ad(t, e.key, n),
			format: String
		}), /* @__PURE__ */ (0, L.jsx)(fd, {
			facet: n,
			reading: od(t, e.key, n)
		})]
	});
}
function Sd({ node: e, samples: t, stacks: n = hd, links: r, onCollectStack: i, collecting: a, collectError: o, className: s }) {
	let [c, l] = (0, I.useState)("overview");
	if (!e) return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "flex h-full items-center justify-center text-sm text-muted-foreground",
		children: /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "text-center",
			children: [/* @__PURE__ */ (0, L.jsx)(F, {
				icon: on,
				className: "mb-2 block text-4xl"
			}), "Select a process to view its trace"]
		})
	});
	let u = [
		{
			id: "overview",
			label: "Overview"
		},
		...e.environment || e.environmentError ? [{
			id: "env",
			label: "Environment"
		}] : [],
		...id(t).map((e) => ({
			id: e,
			label: td[e].label
		}))
	], d = u.some((e) => e.id === c) ? c : "overview";
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: P("h-full space-y-density-3 overflow-y-auto p-density-4", s),
		"data-process-detail": e.key,
		children: [
			/* @__PURE__ */ (0, L.jsx)(vd, { node: e }),
			u.length > 1 && /* @__PURE__ */ (0, L.jsx)(Ut, {
				tabs: u,
				value: d,
				onChange: (e) => l(e)
			}),
			d === "overview" ? /* @__PURE__ */ (0, L.jsx)(bd, {
				node: e,
				samples: t,
				stacks: n,
				links: r,
				...i ? { onCollectStack: i } : {},
				...a === void 0 ? {} : { collecting: a },
				...o ? { collectError: o } : {}
			}) : d === "env" ? /* @__PURE__ */ (0, L.jsx)(fd, {
				facet: "env",
				...e.environment ? { environment: e.environment } : {},
				...e.environmentError ? { error: e.environmentError } : {}
			}) : /* @__PURE__ */ (0, L.jsx)(xd, {
				node: e,
				samples: t,
				facet: d
			}, `${e.key}:${d}`)
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/traces/TraceWindowList.tsx
var Cd = "This run's trace store is not served, so its windows cannot be opened.", wd = 6, Td = (e) => `${e.run_id}:${e.kind}:${e.stream}:${e.afterSeq ?? 0}`, Ed = (e) => typeof e == "string" || typeof e == "number" || typeof e == "boolean";
function Dd({ summary: e }) {
	let t = Object.entries(e ?? {}).filter(([e, t]) => e !== "count" && Ed(t)).slice(0, wd);
	return /* @__PURE__ */ (0, L.jsx)(L.Fragment, { children: t.map(([e, t]) => /* @__PURE__ */ (0, L.jsxs)("span", {
		className: "rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground",
		"data-summary-chip": e,
		children: [
			e,
			": ",
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: "font-mono",
				children: String(t)
			})
		]
	}, e)) });
}
var Od = class extends I.Component {
	state = {};
	static getDerivedStateFromError(e) {
		return { error: e instanceof Error ? e.message : String(e) };
	}
	render() {
		return this.state.error ? /* @__PURE__ */ (0, L.jsx)("div", {
			className: "px-2 py-1.5 text-xs text-red-700 dark:text-red-400",
			role: "alert",
			children: this.state.error
		}) : this.props.children;
	}
};
function kd({ window: e, fetchRecords: t }) {
	let [n, r] = (0, I.useState)(null), [i, a] = (0, I.useState)(void 0);
	return (0, I.useEffect)(() => {
		let n = new AbortController();
		return t(e, n.signal).then(r, (e) => {
			n.signal.aborted || a(e instanceof Error ? e.message : String(e));
		}), () => n.abort();
	}, [
		e.run_id,
		e.kind,
		e.stream,
		e.afterSeq,
		e.toSeq
	]), i ? /* @__PURE__ */ (0, L.jsx)("div", {
		className: "whitespace-pre-wrap border-t border-border px-2 py-1.5 text-xs text-red-700 dark:text-red-400",
		role: "alert",
		children: i
	}) : n ? /* @__PURE__ */ (0, L.jsx)("div", {
		className: "border-t border-border",
		style: { height: 420 },
		children: /* @__PURE__ */ (0, L.jsx)(Od, { children: /* @__PURE__ */ (0, L.jsx)(ed, {
			kind: e.kind,
			rows: n
		}) })
	}) : /* @__PURE__ */ (0, L.jsx)("div", {
		className: "border-t border-border px-2 py-1.5 text-xs text-muted-foreground",
		children: "Loading records..."
	});
}
function Ad({ window: e, mount: t, formats: n, fetchRecords: r, exportUrl: i, openUrl: a }) {
	let [o, s] = (0, I.useState)(!1), c = "text-xs text-primary hover:underline";
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "rounded border border-border",
		"data-trace-kind": e.kind,
		children: [
			/* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex flex-wrap items-center gap-2 px-2 py-1.5 text-sm",
				children: [
					/* @__PURE__ */ (0, L.jsx)("span", {
						className: "font-mono font-medium text-foreground",
						children: e.kind
					}),
					/* @__PURE__ */ (0, L.jsxs)("span", {
						className: "text-xs text-muted-foreground",
						children: [
							e.count,
							" ",
							e.count === 1 ? "record" : "records"
						]
					}),
					/* @__PURE__ */ (0, L.jsx)(Dd, { summary: e.summary }),
					/* @__PURE__ */ (0, L.jsx)("span", { className: "flex-1" }),
					t && /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
						/* @__PURE__ */ (0, L.jsx)("a", {
							className: c,
							href: a(e),
							target: "_blank",
							rel: "noreferrer",
							children: "Open in query"
						}),
						n.map((t) => /* @__PURE__ */ (0, L.jsx)("a", {
							className: c,
							href: i(e, t.name),
							download: !0,
							children: t.name.toUpperCase()
						}, t.name)),
						/* @__PURE__ */ (0, L.jsx)("button", {
							type: "button",
							className: c,
							onClick: () => s(!o),
							children: o ? "Hide" : "View"
						})
					] })
				]
			}),
			e.error && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "whitespace-pre-wrap px-2 pb-1.5 text-xs text-red-700 dark:text-red-400",
				children: e.error
			}),
			o && t && /* @__PURE__ */ (0, L.jsx)(kd, {
				window: e,
				fetchRecords: r
			})
		]
	});
}
function jd({ windows: e, resolve: t, formats: n, fetchRecords: r, exportUrl: i, openUrl: a, error: o }) {
	let s = e.map((e) => t(e)), c = [...new Set(s.filter((e) => !e.mount).map((e) => e.unavailable || Cd))];
	return /* @__PURE__ */ (0, L.jsxs)("div", { children: [
		/* @__PURE__ */ (0, L.jsx)("h3", {
			className: "mb-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground",
			children: "Traces"
		}),
		c.map((e) => /* @__PURE__ */ (0, L.jsx)("div", {
			className: "mb-2 rounded border border-amber-300 bg-amber-50 px-2 py-1 text-xs text-amber-800 dark:border-amber-500/40 dark:bg-amber-500/10 dark:text-amber-300",
			role: "status",
			children: e
		}, e)),
		/* @__PURE__ */ (0, L.jsx)("div", {
			className: "space-y-2",
			children: e.map((e, t) => /* @__PURE__ */ (0, L.jsx)(Ad, {
				window: e,
				mount: s[t].mount,
				formats: n(e.kind),
				fetchRecords: r,
				exportUrl: i,
				openUrl: a
			}, Td(e)))
		}),
		o && /* @__PURE__ */ (0, L.jsx)("div", {
			className: "mt-1 text-xs text-red-700 dark:text-red-400",
			children: o
		})
	] });
}
// src/components/Field.tsx
function Md({ title: e, children: t }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "space-y-1",
		children: [/* @__PURE__ */ (0, L.jsx)("div", {
			className: "text-[10px] uppercase tracking-wide text-muted-foreground",
			children: e
		}), t]
	});
}
// src/adapters/GavelDetailBody.tsx
function $(e) {
	return typeof e == "object" && !!e && "version" in e && "node" in e;
}
function Nd(e) {
	if (e.framework !== "task" || !e.context || typeof e.context != "object") return null;
	let t = e.context;
	return {
		status: typeof t.status == "string" ? t.status : void 0,
		type: typeof t.type == "string" ? t.type : void 0
	};
}
function Pd({ ctx: e }) {
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [e.parent_test && /* @__PURE__ */ (0, L.jsx)(Md, {
		title: "Parent Test",
		children: /* @__PURE__ */ (0, L.jsx)("span", {
			className: "font-mono text-sm",
			children: e.parent_test
		})
	}), e.import_path && /* @__PURE__ */ (0, L.jsx)(Md, {
		title: "Import Path",
		children: /* @__PURE__ */ (0, L.jsx)("span", {
			className: "font-mono text-sm",
			children: e.import_path
		})
	})] });
}
function Fd({ ctx: e, failureLocation: t, suiteAbove: n }) {
	let r = e.failure_location && e.failure_location !== t, i = e.suite_description && e.suite_description !== n;
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		(i || e.suite_path) && /* @__PURE__ */ (0, L.jsxs)(Md, {
			title: "Ginkgo suite",
			children: [i && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "text-sm font-medium",
				children: e.suite_description
			}), e.suite_path && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "font-mono text-xs text-muted-foreground",
				children: e.suite_path
			})]
		}),
		r && /* @__PURE__ */ (0, L.jsx)(Md, {
			title: "Failure Location",
			children: /* @__PURE__ */ (0, L.jsx)("span", {
				className: "font-mono text-sm text-red-600 dark:text-red-400",
				children: e.failure_location
			})
		}),
		!!e.labels?.length && /* @__PURE__ */ (0, L.jsx)(Md, {
			title: "Labels",
			children: /* @__PURE__ */ (0, L.jsx)("div", {
				className: "flex flex-wrap gap-1",
				children: e.labels.map((e) => /* @__PURE__ */ (0, L.jsx)("span", {
					className: "rounded bg-muted px-2 py-0.5 text-xs",
					children: e
				}, e))
			})
		}),
		!!e.num_attempts && e.num_attempts > 1 && /* @__PURE__ */ (0, L.jsx)(Md, {
			title: "Ginkgo attempts",
			children: /* @__PURE__ */ (0, L.jsx)("span", {
				className: "text-sm",
				children: e.num_attempts
			})
		}),
		!!e.report_entries?.length && /* @__PURE__ */ (0, L.jsx)(Md, {
			title: "Report entries",
			children: /* @__PURE__ */ (0, L.jsx)("div", {
				className: "space-y-2",
				children: e.report_entries.map((e, t) => /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "rounded border border-border bg-muted/50 p-2",
					children: [
						/* @__PURE__ */ (0, L.jsx)("div", {
							className: "text-sm font-medium",
							children: e.name
						}),
						e.file && /* @__PURE__ */ (0, L.jsxs)("div", {
							className: "font-mono text-xs text-muted-foreground",
							children: [e.file, e.line ? `:${e.line}` : ""]
						}),
						/* @__PURE__ */ (0, L.jsx)("div", {
							className: "mt-1 break-words",
							children: /* @__PURE__ */ (0, L.jsx)(je, { data: e.value ?? e.representation })
						})
					]
				}, `${e.name}-${t}`))
			})
		})
	] });
}
function Id({ node: e }) {
	let t = e, n = te(), r = t.attempts?.at(-1), i = Nd(t), a = (t.children?.length ?? 0) > 0 ? nt(e) : null, o = t.suite?.join(" > "), s = t.failure_detail && t.failure_detail.kind && t.failure_detail.kind !== "raw" ? t.failure_detail : void 0, c = !!t.message && !!s && t.message !== s.summary && t.message.includes("\n");
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "space-y-4 p-density-4",
		children: [
			t.cached && /* @__PURE__ */ (0, L.jsx)("span", {
				className: "inline-block rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground",
				children: "cached"
			}),
			o && /* @__PURE__ */ (0, L.jsx)(Md, {
				title: "Suite",
				children: /* @__PURE__ */ (0, L.jsx)("span", {
					className: "text-sm",
					children: o
				})
			}),
			t.package_path && /* @__PURE__ */ (0, L.jsx)(Md, {
				title: "Package",
				children: /* @__PURE__ */ (0, L.jsx)("span", {
					className: "font-mono text-sm",
					children: t.package_path
				})
			}),
			s ? /* @__PURE__ */ (0, L.jsxs)(Md, {
				title: "Error",
				children: [/* @__PURE__ */ (0, L.jsx)(O, { failure: s }), c && /* @__PURE__ */ (0, L.jsxs)("details", {
					className: "rounded bg-muted/50 p-2",
					children: [/* @__PURE__ */ (0, L.jsx)("summary", {
						className: "cursor-pointer text-xs font-semibold text-muted-foreground",
						children: "Full message"
					}), /* @__PURE__ */ (0, L.jsx)(ne, {
						text: t.message,
						className: "mt-2 block max-h-64 overflow-y-auto whitespace-pre-wrap font-mono text-xs"
					})]
				})]
			}) : t.message ? /* @__PURE__ */ (0, L.jsx)(Md, {
				title: "Error",
				children: /* @__PURE__ */ (0, L.jsx)(ne, {
					text: t.message,
					className: "block max-h-64 overflow-y-auto whitespace-pre-wrap rounded bg-red-50 p-3 font-mono text-sm text-red-700 dark:bg-red-500/10 dark:text-red-300"
				})
			}) : null,
			t.command && /* @__PURE__ */ (0, L.jsx)(Md, {
				title: "Command",
				children: /* @__PURE__ */ (0, L.jsxs)("pre", {
					className: "whitespace-pre-wrap rounded bg-muted p-2 font-mono text-sm",
					children: [/* @__PURE__ */ (0, L.jsx)("span", {
						className: "text-muted-foreground",
						children: "$ "
					}), t.command]
				})
			}),
			i && (i.status || i.type) && /* @__PURE__ */ (0, L.jsx)(Md, {
				title: "Task",
				children: /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "flex gap-4 text-sm",
					children: [i.type && /* @__PURE__ */ (0, L.jsxs)("span", { children: ["type: ", /* @__PURE__ */ (0, L.jsx)("span", {
						className: "font-mono",
						children: i.type
					})] }), i.status && /* @__PURE__ */ (0, L.jsxs)("span", { children: ["status: ", /* @__PURE__ */ (0, L.jsx)("span", {
						className: "font-mono",
						children: i.status
					})] })]
				})
			}),
			t.framework === "go test" && t.context && /* @__PURE__ */ (0, L.jsx)(Pd, { ctx: t.context }),
			t.framework === "ginkgo" && t.context && /* @__PURE__ */ (0, L.jsx)(Fd, {
				ctx: t.context,
				failureLocation: t.failure_detail?.location,
				suiteAbove: o
			}),
			r && /* @__PURE__ */ (0, L.jsx)(Md, {
				title: "Attempt",
				children: /* @__PURE__ */ (0, L.jsx)(E, {
					attempt: r,
					...n.fetchProcessMetrics ? { fetchProcessMetrics: n.fetchProcessMetrics } : {}
				}, `${r.run_kind ?? ""}:${r.sequence}`)
			}),
			/* @__PURE__ */ (0, L.jsx)(oe, { node: e }),
			a && a.total > 0 && /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex flex-wrap gap-3 text-xs text-muted-foreground",
				children: [
					/* @__PURE__ */ (0, L.jsxs)("span", { children: [a.total, " total"] }),
					a.passed > 0 && /* @__PURE__ */ (0, L.jsxs)("span", { children: [a.passed, " passed"] }),
					a.failed > 0 && /* @__PURE__ */ (0, L.jsxs)("span", { children: [a.failed, " failed"] }),
					a.warned > 0 && /* @__PURE__ */ (0, L.jsxs)("span", { children: [a.warned, " warned"] }),
					a.aborted > 0 && /* @__PURE__ */ (0, L.jsxs)("span", { children: [a.aborted, " aborted"] }),
					a.skipped > 0 && /* @__PURE__ */ (0, L.jsxs)("span", { children: [a.skipped, " skipped"] }),
					a.pending > 0 && /* @__PURE__ */ (0, L.jsxs)("span", { children: [a.pending, " pending"] })
				]
			}),
			t.detail != null && /* @__PURE__ */ (0, L.jsx)(Md, {
				title: "Detail",
				children: $(t.detail) ? /* @__PURE__ */ (0, L.jsx)(ma, { data: t.detail }) : /* @__PURE__ */ (0, L.jsx)(je, { data: t.detail })
			})
		]
	});
}
// src/config.ts
var Ld = typeof window < "u" && window.__gavelBasePath || "";
function Rd({ baseUrl: e = Ld, runId: t } = {}) {
	return t ? `${e}/${encodeURIComponent(t)}` : e;
}
function zd(e, t) {
	return Rd(t) + e;
}
// src/routes.ts
var Bd = ["!passed", "!skipped"], Vd = "all";
function Hd() {
	return qe(Bd);
}
function Ud(e) {
	let t = Pe(e);
	if (t.length !== Bd.length) return !1;
	let n = [...Bd].sort(), r = [...t].sort();
	return n.every((e, t) => e === r[t]);
}
var Wd = [
	"tests",
	"lint",
	"files",
	"bench",
	"processes",
	"history"
];
function Gd(e) {
	return e ? e.split(",").map((e) => e.trim()).filter(Boolean) : [];
}
function Kd(e) {
	return e === "source" || e === "problems" ? e : "preview";
}
function qd(e) {
	let t = e.pathname;
	Ld && t.startsWith(Ld) && (t = t.slice(Ld.length));
	let n = t.replace(/^\/+|\/+$/g, ""), r = n ? n.split("/").map(decodeURIComponent) : [], i = "tests", a = "", o = r[0];
	o && Wd.includes(o) && (i = o, a = r.slice(1).join("/"));
	let s = new URLSearchParams(e.search), c = s.get("sample"), l = s.get("line"), u = s.get("dir"), d = s.get("status"), f;
	return f = d === null ? Hd() : d === Vd ? /* @__PURE__ */ new Map() : qe(Gd(d)), {
		tab: i,
		selectedPath: a,
		filters: {
			status: f,
			framework: qe(Gd(s.get("framework")))
		},
		lintGrouping: Xd(s.get("grouping")),
		lintFilters: {
			severity: qe(Gd(s.get("severity"))),
			linter: qe(Gd(s.get("linter")))
		},
		sampleIndex: c !== null && /^\d+$/.test(c) ? Number(c) : void 0,
		hideExited: s.get("hideExited") === "1",
		...u ? { workDir: u } : {},
		view: Kd(s.get("view")),
		line: l !== null && /^\d+$/.test(l) && Number(l) > 0 ? Number(l) : void 0
	};
}
function Jd(e) {
	let t = [e.tab];
	e.selectedPath && t.push(...e.selectedPath.split("/").map(encodeURIComponent));
	let n = new URLSearchParams();
	e.tab === "tests" && (Ud(e.filters.status) || (e.filters.status.size === 0 ? n.set("status", Vd) : n.set("status", Pe(e.filters.status).join(","))), e.filters.framework.size > 0 && n.set("framework", Pe(e.filters.framework).join(","))), e.tab === "lint" && (e.lintGrouping !== "linter-rule-file" && n.set("grouping", e.lintGrouping), e.lintFilters.severity.size > 0 && n.set("severity", Pe(e.lintFilters.severity).join(",")), e.lintFilters.linter.size > 0 && n.set("linter", Pe(e.lintFilters.linter).join(","))), e.tab === "files" && (e.workDir && n.set("dir", e.workDir), e.view && e.view !== "preview" && n.set("view", e.view), e.line !== void 0 && n.set("line", String(e.line))), e.tab === "bench" && e.sampleIndex !== void 0 && n.set("sample", String(e.sampleIndex)), e.tab === "processes" && e.hideExited && n.set("hideExited", "1");
	let r = n.toString();
	return `${Ld}/${t.join("/")}${r ? `?${r}` : ""}`;
}
function Yd(e, t) {
	let [n, r = ""] = Jd(e).split("?", 2);
	return `${n}.${t}${r ? `?${r}` : ""}`;
}
function Xd(e) {
	return sc.find((t) => t === e) ?? "linter-rule-file";
}
// src/export.ts
function Zd(e, t) {
	return `${(e.selectedPath ? e.selectedPath.replace(/\//g, "-") : e.tab) || e.tab}.${t}`;
}
function Qd(e, t) {
	let n = document.createElement("a");
	n.href = Yd(e, t), n.download = Zd(e, t), document.body.appendChild(n), n.click(), n.remove();
}
function $d(e) {
	return e.tab === "lint" ? "Use this gavel lint report as context for your analysis or code changes." : e.tab === "bench" ? "Use this gavel benchmark report as context for your analysis or code changes." : "Use this gavel test report as context for your analysis or code changes.";
}
async function ef(e) {
	let t = Yd(e, "md"), n = new URL(t, window.location.origin).toString(), r = await fetch(t, { headers: { Accept: "text/markdown" } });
	if (!r.ok) throw Error(`Markdown export failed (${r.status})`);
	let i = (await r.text()).trimEnd(), a = [
		$d(e),
		"",
		`Source URL: ${n}`,
		"",
		"```md",
		i,
		"```"
	].join("\n");
	await Ye(a);
}
// src/hooks/use-copy-feedback.ts
function tf({ copiedMs: e = 2e3, errorMs: t = 3e3 } = {}) {
	let [n, r] = (0, I.useState)("idle"), [i, a] = (0, I.useState)(""), o = (0, I.useRef)(null), s = (0, I.useCallback)(() => {
		o.current !== null && (window.clearTimeout(o.current), o.current = null);
	}, []);
	return (0, I.useEffect)(() => s, [s]), {
		copyState: n,
		copyError: i,
		beginCopy: (0, I.useCallback)(() => {
			r("copying"), a(""), s();
		}, [s]),
		resetCopyFeedback: (0, I.useCallback)((n, i = "") => {
			r(n), a(i), s(), o.current = window.setTimeout(() => {
				o.current = null, r("idle"), a("");
			}, n === "copied" ? e : t);
		}, [
			s,
			e,
			t
		])
	};
}
function nf(e) {
	let { copyState: t, copyError: n, beginCopy: r, resetCopyFeedback: i } = tf();
	return {
		copyState: t,
		copyError: n,
		run: (0, I.useCallback)(async () => {
			if (t !== "copying") {
				r();
				try {
					await e(), i("copied");
				} catch (e) {
					i("error", e instanceof Error ? e.message : "Copy failed");
				}
			}
		}, [
			e,
			t,
			r,
			i
		])
	};
}
// src/components/ExportControls.tsx
function rf({ routeState: e, title: t }) {
	return /* @__PURE__ */ (0, L.jsx)(He, {
		label: "Download",
		icon: We,
		variant: "outline",
		size: "sm",
		items: [{
			label: "JSON",
			icon: w,
			title: Yd(e, "json"),
			onSelect: () => Qd(e, "json")
		}, {
			label: "Markdown",
			icon: S,
			title: Yd(e, "md"),
			onSelect: () => Qd(e, "md")
		}],
		...t ? { title: t } : {}
	});
}
var af = {
	idle: "Copy AI Prompt",
	copying: "Copying...",
	copied: "Copied",
	error: "Copy failed"
}, of = {
	idle: Je,
	copying: ht,
	copied: Ue,
	error: re
};
function sf({ routeState: e, disabled: t, disabledTitle: n, title: r }) {
	let { copyState: i, copyError: a, run: o } = nf(() => ef(e)), s = of[i];
	return /* @__PURE__ */ (0, L.jsxs)(be, {
		type: "button",
		variant: "outline",
		size: "sm",
		disabled: t || i === "copying",
		onClick: () => {
			o();
		},
		title: t ? n : a || r || Yd(e, "md"),
		children: [/* @__PURE__ */ (0, L.jsx)(s, { className: i === "copying" ? "mr-1 animate-spin" : "mr-1" }), af[i]]
	});
}
// src/adapters/NodeActions.tsx
var cf = /* @__PURE__ */ new Set([
	"go test",
	"ginkgo",
	"vitest"
]);
function lf(e, t) {
	return t && e.framework !== "task" && cf.has(e.framework ?? "") && !!e.file;
}
function uf(e) {
	if (e.failed || e.timed_out) return !0;
	if ((e.children?.length ?? 0) === 0) return !1;
	let t = nt(e);
	return t.failed > 0 || t.timedout > 0;
}
function df({ node: e, busy: t, onEdit: n }) {
	let [r, i] = (0, I.useState)(null), a = r?.scope === "file" ? e.file || "file" : e.name || "test", o = r?.action === "skip" ? "Skip" : "Delete", s = r?.scope === "file" ? "file" : "test", c = (e, t) => () => i({
		action: e,
		scope: t
	});
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsx)(He, {
		label: "Source",
		title: "Edit the test source",
		items: [
			{
				label: "Skip Test",
				icon: Qe,
				disabled: t,
				onSelect: c("skip", "test")
			},
			{
				label: "Delete Test",
				icon: Re,
				disabled: t,
				onSelect: c("delete", "test")
			},
			{
				label: "Skip File",
				icon: Ve,
				disabled: t,
				onSelect: c("skip", "file")
			},
			{
				label: "Delete File",
				icon: Re,
				disabled: t,
				onSelect: c("delete", "file")
			}
		]
	}), /* @__PURE__ */ (0, L.jsx)(Le, {
		open: r !== null,
		title: `${o} ${s} ${a}?`,
		message: "This edits the source file on disk.",
		confirmLabel: o,
		variant: r?.action === "delete" ? "destructive" : "default",
		onCancel: () => i(null),
		onConfirm: () => {
			let t = r;
			i(null), t && n(e, t.action, t.scope);
		}
	})] });
}
function ff({ node: e, config: t }) {
	let n = e.framework === "task" ? void 0 : t.nodeRouteState(e), r = t.failingOnlyRouteState(e), i = uf(e);
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		n && /* @__PURE__ */ (0, L.jsx)(rf, {
			routeState: n,
			title: "Download this node as JSON or Markdown"
		}),
		n && r && /* @__PURE__ */ (0, L.jsx)(sf, {
			routeState: r,
			disabled: !i,
			disabledTitle: "No failures to copy",
			title: "Copy failing output and repro steps as AI prompt"
		}),
		lf(e, t.testEditSupported) && /* @__PURE__ */ (0, L.jsx)(df, {
			node: e,
			busy: t.testEditBusy,
			onEdit: t.onTestEdit
		})
	] });
}
// src/components/ResourceParts.tsx
function pf(e) {
	return e >= 1e3 ? `${(e / 1e3).toFixed(2)} s` : `${e.toFixed(1)} ms`;
}
function mf(e) {
	return e >= 1048576 ? `${(e / 1048576).toFixed(1)} MiB` : e >= 1024 ? `${(e / 1024).toFixed(1)} KiB` : `${e} B`;
}
function hf(e) {
	return `${e >= 0 ? "+" : "-"}${mf(Math.abs(e))}`;
}
function gf({ profile: e }) {
	let t = e.statements;
	return t?.length ? /* @__PURE__ */ (0, L.jsx)(Xc, {
		title: `${t.length} ${t.length === 1 ? "query" : "queries"}`,
		defaultOpen: !1,
		children: /* @__PURE__ */ (0, L.jsx)("div", {
			className: "max-h-80 space-y-2 overflow-auto",
			children: t.map((e, t) => /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "rounded border border-border bg-muted/50 p-2 text-xs",
				children: [
					/* @__PURE__ */ (0, L.jsxs)("div", {
						className: "mb-1 text-muted-foreground",
						children: [
							e.duration_ms.toFixed(1),
							" ms · ",
							e.rows,
							" ",
							e.rows === 1 ? "row" : "rows",
							e.slow ? " · slow" : "",
							e.error ? " · error" : ""
						]
					}),
					/* @__PURE__ */ (0, L.jsx)("pre", {
						className: "whitespace-pre-wrap break-all font-mono",
						children: e.sql
					}),
					e.params.length > 0 && /* @__PURE__ */ (0, L.jsxs)("div", {
						className: "mt-1",
						children: [/* @__PURE__ */ (0, L.jsx)("span", {
							className: "text-muted-foreground",
							children: "Params"
						}), /* @__PURE__ */ (0, L.jsx)("ol", {
							className: "list-decimal pl-5 font-mono",
							children: e.params.map((e, t) => /* @__PURE__ */ (0, L.jsx)("li", {
								className: "break-all",
								children: e
							}, t))
						})]
					})
				]
			}, t))
		})
	}) : null;
}
function _f({ profile: e }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "space-y-2",
		children: [
			e.sample_count > 0 ? /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "grid grid-cols-3 gap-density-1",
				children: [
					/* @__PURE__ */ (0, L.jsx)(M, {
						label: "Peak CPU",
						value: `${e.peak_cpu_percent.toFixed(1)}%`
					}),
					/* @__PURE__ */ (0, L.jsx)(M, {
						label: "Peak RSS",
						value: m(e.peak_rss_bytes)
					}),
					/* @__PURE__ */ (0, L.jsx)(M, {
						label: "Peak memory",
						value: `${e.peak_memory_percent.toFixed(1)}%`
					})
				]
			}) : /* @__PURE__ */ (0, L.jsx)("p", {
				className: "text-sm text-muted-foreground",
				children: "No process samples were captured during this test."
			}),
			e.disk_io && /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "grid grid-cols-2 gap-density-1",
				children: [/* @__PURE__ */ (0, L.jsx)(M, {
					label: "Observed disk read",
					value: m(e.disk_io.disk_read_bytes)
				}), /* @__PURE__ */ (0, L.jsx)(M, {
					label: "Observed disk write",
					value: m(e.disk_io.disk_write_bytes)
				})]
			}),
			/* @__PURE__ */ (0, L.jsxs)("p", {
				className: "text-xs text-muted-foreground",
				children: [
					e.sample_count,
					" samples · ",
					e.scope === "process_tree" ? `Command process tree · PID ${e.pid}` : "Run process tree during this test"
				]
			})
		]
	});
}
// src/adapters/ResourcesPanel.tsx
function vf({ node: e }) {
	let t = e.sql_profile;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "space-y-4 p-density-4",
		children: [e.fixture_profile && /* @__PURE__ */ (0, L.jsx)(Md, {
			title: "Resources",
			children: /* @__PURE__ */ (0, L.jsx)(_f, { profile: e.fixture_profile })
		}), t && /* @__PURE__ */ (0, L.jsxs)(Md, {
			title: "SQL profile",
			children: [/* @__PURE__ */ (0, L.jsxs)("div", {
				className: "grid grid-cols-3 gap-density-1",
				children: [
					/* @__PURE__ */ (0, L.jsx)(M, {
						label: "Queries",
						value: String(t.query_count)
					}),
					/* @__PURE__ */ (0, L.jsx)(M, {
						label: "Total SQL",
						value: `${t.total_duration_ms.toFixed(1)} ms`
					}),
					/* @__PURE__ */ (0, L.jsx)(M, {
						label: "Slow queries",
						value: String(t.slow_query_count)
					})
				]
			}), /* @__PURE__ */ (0, L.jsx)(gf, { profile: t })]
		})]
	});
}
// src/adapters/StepProcesses.tsx
function yf({ processes: e, onOpenProcess: t }) {
	return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "flex h-80 min-h-0 flex-col px-3 pt-2",
		"data-step-processes": "",
		children: /* @__PURE__ */ (0, L.jsx)(Zl, {
			nodes: e,
			selectedKey: null,
			expandAll: !0,
			onSelect: t,
			emptyLabel: "This step started no processes the trace recorded"
		})
	});
}
// src/traces.ts
var bf = class extends Error {};
function xf(e, t) {
	return e?.runs[t] || { unavailable: `The trace store of run ${t || "(unnamed)"} is not served by this view.` };
}
function Sf(e) {
	if (e?.process_run && e.process) return e.runs[e.process_run]?.mount;
}
function Cf(e, t) {
	if (!e) return;
	if (!t) return e;
	let n = Object.fromEntries(Object.entries(e.runs).map(([e, n]) => [e, n.mount ? {
		...n,
		mount: `${t}${n.mount}`
	} : n]));
	return {
		...e,
		runs: n
	};
}
function wf(e, t) {
	return `${e}${t}`;
}
function Tf(e, t = {}) {
	let n = new URLSearchParams({ stream: e.stream });
	e.afterSeq && n.set("afterSeq", String(e.afterSeq)), e.toSeq && n.set("toSeq", String(e.toSeq));
	for (let [e, r] of Object.entries(t)) n.set(e, r);
	return n.toString();
}
function Ef(e, t, n) {
	return wf(e, `/api/v1/traces/${encodeURIComponent(t.kind)}/export?${Tf(t, { format: n })}`);
}
function Df(e, t) {
	return wf(e, `/traces/${encodeURIComponent(t.kind)}?${Tf(t)}`);
}
async function Of(e, t) {
	let n = (await e.text()).trim();
	return /* @__PURE__ */ Error(`${t} failed (${e.status})${n ? `: ${n}` : ""}`);
}
function kf(e) {
	return e.split("\n").filter((e) => e.trim() !== "").map((e) => JSON.parse(e));
}
async function Af(e, t, n) {
	let r = await fetch(Ef(e, t, "ndjson"), n ? { signal: n } : void 0);
	if (r.status === 404) throw new bf(`${t.kind} stream ${t.stream} has no records`);
	if (!r.ok) throw await Of(r, `export ${t.kind} window`);
	return kf(await r.text());
}
function jf(e) {
	return `profile-${`traces/${e}`.toLowerCase().replace(/[ _/.-]/g, "-").replace(/[^a-z0-9-]/g, "").replace(/^-+|-+$/g, "")}`;
}
async function Mf(e, t) {
	let n = new URLSearchParams({
		stream: t.stream,
		sort: "seq",
		order: "desc",
		limit: "1"
	}), r = await fetch(wf(e, `/api/v1/profile/${jf(t.kind)}?${n}`), { headers: { Accept: "application/json" } });
	if (!r.ok) throw await Of(r, `read ${t.kind} stream`);
	return (await r.json())[0]?.seq ?? 0;
}
async function Nf(e) {
	let t = await fetch(wf(e, "/api/v1/traces/kinds"), { headers: { Accept: "application/json" } });
	if (!t.ok) throw await Of(t, "list trace kinds");
	return t.json();
}
async function Pf({ mount: e, session: t, params: n }) {
	let r = await fetch(wf(e, `/api/v1/sessions/${encodeURIComponent(t)}/checkpoint`), {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(n)
	});
	if (!r.ok) throw await Of(r, "checkpoint");
}
// src/adapters/TracesPanel.tsx
function Ff(e) {
	let [t, n] = (0, I.useState)({}), [r, i] = (0, I.useState)(void 0);
	return (0, I.useEffect)(() => {
		e && Nf(e).then((e) => n(Object.fromEntries(e.map((e) => [e.name, e.exportFormats])))).catch((e) => i(e instanceof Error ? e.message : String(e)));
	}, [e]), {
		byKind: t,
		error: r
	};
}
function If({ windows: e, traces: t }) {
	let n = Ff(e.map((e) => xf(t, e.run_id).mount).find(Boolean)), r = (0, I.useCallback)((e) => xf(t, e.run_id), [t]), i = (0, I.useCallback)((e) => {
		let n = xf(t, e.run_id).mount;
		if (!n) throw Error(`trace store of run ${e.run_id} is not served`);
		return n;
	}, [t]), a = (0, I.useMemo)(() => ({
		formats: (e) => n.byKind[e] ?? [],
		fetchRecords: (e, t) => Af(i(e), e, t),
		exportUrl: (e, t) => Ef(i(e), e, t),
		openUrl: (e) => Df(i(e), e)
	}), [n.byKind, i]);
	return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "space-y-density-3 p-density-4",
		children: /* @__PURE__ */ (0, L.jsx)(jd, {
			windows: e,
			resolve: r,
			...a,
			...n.error ? { error: n.error } : {}
		})
	});
}
// src/adapters/index.tsx
var Lf = (e) => e;
function Rf(e) {
	let t = Lf(e);
	return t.framework === "fixture" ? t.context?.traces ?? [] : [];
}
function zf(e, t) {
	let n = new Set(Lf(e).fixture?.pids ?? []);
	return n.size === 0 ? [] : t.filter((e) => n.has(e.pid));
}
function Bf() {
	return {
		id: "gavel-detail",
		match: (e) => Lf(e).framework !== "fixture",
		renderDetail: ({ node: e }) => /* @__PURE__ */ (0, L.jsx)(Id, { node: e })
	};
}
function Vf(e) {
	return {
		id: "gavel-traces",
		match: (e) => Rf(e).length > 0,
		detailTabs: ({ node: t }) => [{
			id: "traces",
			label: "Traces",
			render: () => /* @__PURE__ */ (0, L.jsx)(If, {
				windows: Rf(t),
				traces: e
			})
		}]
	};
}
function Hf({ processes: e, onOpenProcess: t }) {
	return {
		id: "gavel-processes",
		match: (t) => zf(t, e).length > 0,
		detailTabs: ({ node: n }) => [{
			id: "processes",
			label: "Processes",
			render: () => /* @__PURE__ */ (0, L.jsx)(yf, {
				processes: zf(n, e),
				onOpenProcess: t
			})
		}]
	};
}
function Uf() {
	return {
		id: "gavel-resources",
		match: (e) => {
			let t = Lf(e);
			return !!t.fixture_profile || !!t.sql_profile;
		},
		detailTabs: ({ node: e }) => [{
			id: "resources",
			label: "Resources",
			render: () => /* @__PURE__ */ (0, L.jsx)(vf, { node: Lf(e) })
		}]
	};
}
function Wf(e) {
	return {
		id: "gavel-actions",
		match: (e) => Lf(e).framework !== "task",
		nodeActions: ({ node: t }) => /* @__PURE__ */ (0, L.jsx)(ff, {
			node: Lf(t),
			config: e
		})
	};
}
function Gf({ traces: e, processes: t, onOpenProcess: n, actions: r }) {
	return [
		Bf(),
		Wf(r),
		Vf(e),
		Hf({
			processes: t,
			onOpenProcess: n
		}),
		Uf()
	];
}
function Kf(e) {
	return ie([...Gf(e), ...pe().adapters]);
}
// src/components/BenchView.tsx
function qf(e) {
	return e >= 1e9 ? `${(e / 1e9).toFixed(2)}s` : e >= 1e6 ? `${(e / 1e6).toFixed(2)}ms` : e >= 1e3 ? `${(e / 1e3).toFixed(2)}µs` : `${e.toFixed(2)}ns`;
}
function Jf(e, t) {
	return e.only_in ? "only" : e.significant && e.delta_pct > t ? "regression" : e.significant && e.delta_pct < -t ? "improvement" : "neutral";
}
function Yf(e, t) {
	let n = !!e.only_in;
	return {
		name: e.name,
		base: n ? null : e.base_mean,
		head: n ? null : e.head_mean,
		baseSpread: e.base_stddev ? `±${e.base_stddev.toFixed(1)}%` : "",
		headSpread: e.head_stddev ? `±${e.head_stddev.toFixed(1)}%` : "",
		delta: n ? null : e.delta_pct,
		pValue: e.p_value !== void 0 && e.p_value > 0 ? e.p_value : null,
		samples: e.samples ?? null,
		category: Jf(e, t),
		onlyIn: e.only_in ?? null,
		threshold: t,
		significant: !!e.significant
	};
}
var Xf = {
	regression: "text-red-600 dark:text-red-400",
	improvement: "text-green-600 dark:text-green-400",
	neutral: "text-muted-foreground",
	only: "text-muted-foreground"
}, Zf = {
	regression: "bg-red-500",
	improvement: "bg-green-500",
	neutral: "bg-muted-foreground/40",
	only: "bg-muted-foreground/40"
};
function Qf({ row: e }) {
	if (e.delta === null) return /* @__PURE__ */ (0, L.jsxs)("span", {
		className: "text-xs italic text-muted-foreground",
		children: ["only in ", e.onlyIn]
	});
	let t = Math.max(-50, Math.min(50, e.delta)), n = Math.abs(t);
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex w-full min-w-48 items-center gap-2",
		children: [/* @__PURE__ */ (0, L.jsxs)("div", {
			className: "relative h-4 min-w-20 flex-1 overflow-hidden rounded bg-muted",
			children: [/* @__PURE__ */ (0, L.jsx)("div", { className: "absolute inset-y-0 left-1/2 w-px bg-muted-foreground/60" }), /* @__PURE__ */ (0, L.jsx)("div", {
				className: P("absolute inset-y-0", Zf[e.category]),
				style: e.delta >= 0 ? {
					left: "50%",
					width: `${n}%`
				} : {
					right: "50%",
					width: `${n}%`
				}
			})]
		}), /* @__PURE__ */ (0, L.jsxs)("span", {
			className: P("w-16 text-right text-xs tabular-nums", Xf[e.category], e.significant && "font-semibold"),
			children: [
				e.delta >= 0 ? "+" : "",
				e.delta.toFixed(2),
				"%"
			]
		})]
	});
}
var $f = [
	{
		key: "name",
		label: "Benchmark",
		sortable: !0,
		grow: !0,
		cellClassName: "font-mono text-xs"
	},
	{
		key: "base",
		label: "Base",
		sortable: !0,
		align: "right",
		render: (e, t) => t.base === null ? "" : /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "text-xs tabular-nums",
			children: [qf(t.base), t.baseSpread && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "text-[10px] text-muted-foreground",
				children: t.baseSpread
			})]
		})
	},
	{
		key: "head",
		label: "Head",
		sortable: !0,
		align: "right",
		render: (e, t) => t.head === null ? "" : /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "text-xs tabular-nums",
			children: [qf(t.head), t.headSpread && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "text-[10px] text-muted-foreground",
				children: t.headSpread
			})]
		})
	},
	{
		key: "delta",
		label: "Δ",
		sortable: !0,
		render: (e, t) => /* @__PURE__ */ (0, L.jsx)(Qf, { row: t })
	},
	{
		key: "pValue",
		label: "p-value",
		sortable: !0,
		align: "right",
		render: (e, t) => /* @__PURE__ */ (0, L.jsxs)("span", {
			className: "text-xs tabular-nums text-muted-foreground",
			children: [t.pValue === null ? "—" : t.pValue.toPrecision(2), t.samples !== null && /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "ml-1",
				children: ["n=", t.samples]
			})]
		})
	}
];
function ep({ bench: e }) {
	let t = (0, I.useMemo)(() => (e?.deltas ?? []).map((t) => Yf(t, e.threshold)), [e]), n = (0, I.useMemo)(() => {
		let e = {
			regression: 0,
			improvement: 0,
			neutral: 0,
			only: 0
		};
		return t.forEach((t) => {
			e[t.category] += 1;
		}), e;
	}, [t]);
	return e ? /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col gap-density-2 p-density-4",
		children: [/* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex flex-wrap items-center gap-4 text-sm",
			children: [
				/* @__PURE__ */ (0, L.jsxs)("div", { children: [
					/* @__PURE__ */ (0, L.jsx)("span", {
						className: "font-mono",
						children: e.base_label || "base"
					}),
					/* @__PURE__ */ (0, L.jsx)("span", {
						className: "mx-2 text-muted-foreground",
						children: "→"
					}),
					/* @__PURE__ */ (0, L.jsx)("span", {
						className: "font-mono",
						children: e.head_label || "head"
					})
				] }),
				/* @__PURE__ */ (0, L.jsxs)("div", {
					className: "text-xs text-muted-foreground",
					children: [
						"threshold ±",
						e.threshold.toFixed(1),
						"%"
					]
				}),
				n.regression > 0 && /* @__PURE__ */ (0, L.jsxs)(Oe, {
					tone: "danger",
					variant: "soft",
					size: "xs",
					children: [
						n.regression,
						" regression",
						n.regression > 1 ? "s" : ""
					]
				}),
				n.improvement > 0 && /* @__PURE__ */ (0, L.jsxs)(Oe, {
					tone: "success",
					variant: "soft",
					size: "xs",
					children: [
						n.improvement,
						" improvement",
						n.improvement > 1 ? "s" : ""
					]
				}),
				n.neutral > 0 && /* @__PURE__ */ (0, L.jsxs)(Oe, {
					tone: "neutral",
					variant: "soft",
					size: "xs",
					children: [n.neutral, " unchanged"]
				}),
				/* @__PURE__ */ (0, L.jsxs)("div", {
					className: P("ml-auto font-mono text-sm", e.geomean_delta > e.threshold ? "font-bold text-red-600 dark:text-red-400" : e.geomean_delta < -e.threshold ? "font-bold text-green-600 dark:text-green-400" : "text-muted-foreground"),
					children: [
						"geomean: ",
						e.geomean_delta >= 0 ? "+" : "",
						e.geomean_delta.toFixed(2),
						"%"
					]
				})
			]
		}), /* @__PURE__ */ (0, L.jsx)(o, {
			data: t,
			columns: $f,
			defaultSort: {
				key: "delta",
				dir: "desc"
			},
			emptyMessage: "No benchmarks were compared",
			className: "min-h-0 flex-1"
		})]
	}) : /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "p-8 text-center text-sm text-muted-foreground",
		children: [
			"No benchmark comparison loaded. Run ",
			/* @__PURE__ */ (0, L.jsx)("code", {
				className: "rounded bg-muted px-1",
				children: "gavel bench compare --ui"
			}),
			"."
		]
	});
}
// src/components/FixtureBenchView.tsx
var tp = [
	"setup",
	"build",
	"daemon",
	"daemon-stop",
	"cleanup"
];
function np(e) {
	let t = [], n = [], r = (e) => {
		if (e.fixture) {
			let r = {
				key: e.fixture.key,
				name: e.name,
				kind: e.fixture.kind,
				status: e.fixture.state,
				duration_ms: (e.duration || 0) / 1e6,
				command_ms: e.fixture.command_ms,
				profile: e.fixture_profile,
				sql_profile: e.sql_profile,
				violations: e.fixture.violations
			};
			tp.includes(r.kind) ? t.push(r) : n.push(r);
		}
		e.children?.forEach(r);
	};
	return e.forEach(r), {
		phases: t,
		fixtures: n
	};
}
function rp(e, t) {
	let n = new Map(t?.map((e) => [e.key, e]));
	return e.map((e) => {
		let t = n.get(e.key) ?? null;
		return {
			key: e.key,
			name: e.name,
			kind: e.kind,
			status: e.status,
			duration: e.duration_ms,
			command: e.command_ms ?? null,
			diskRead: e.profile?.disk_io?.disk_read_bytes ?? null,
			diskWrite: e.profile?.disk_io?.disk_write_bytes ?? null,
			sql: e.sql_profile?.total_duration_ms ?? null,
			baseline: t?.baseline_ms ?? null,
			change: t?.delta_pct ?? null,
			entry: e,
			delta: t
		};
	});
}
function ip(e) {
	let t = "font-mono tabular-nums", n = [
		{
			key: "name",
			label: "Step",
			sortable: !0,
			grow: !0,
			render: (e, t) => /* @__PURE__ */ (0, L.jsxs)("div", { children: [
				/* @__PURE__ */ (0, L.jsx)("span", {
					className: "font-medium",
					children: t.name
				}),
				/* @__PURE__ */ (0, L.jsx)("span", {
					className: "ml-2 text-xs text-muted-foreground",
					children: t.kind
				}),
				t.entry.violations?.map((e) => /* @__PURE__ */ (0, L.jsx)("div", {
					className: "text-xs text-red-700 dark:text-red-400",
					children: e
				}, e))
			] })
		},
		{
			key: "status",
			label: "Status",
			sortable: !0,
			filterable: !0
		},
		{
			key: "duration",
			label: "Duration",
			sortable: !0,
			align: "right",
			cellClassName: t,
			render: (e, t) => pf(t.duration)
		},
		{
			key: "command",
			label: "Command",
			sortable: !0,
			align: "right",
			cellClassName: `${t} text-muted-foreground`,
			render: (e, t) => t.command === null ? "—" : pf(t.command)
		},
		{
			key: "diskRead",
			label: "Disk read",
			sortable: !0,
			align: "right",
			cellClassName: t,
			render: (e, t) => t.diskRead === null ? "—" : mf(t.diskRead)
		},
		{
			key: "diskWrite",
			label: "Disk write",
			sortable: !0,
			align: "right",
			cellClassName: t,
			render: (e, t) => t.diskWrite === null ? "—" : mf(t.diskWrite)
		},
		{
			key: "sql",
			label: "SQL",
			sortable: !0,
			align: "right",
			cellClassName: t,
			render: (e, t) => {
				let n = t.entry.sql_profile;
				return n ? /* @__PURE__ */ (0, L.jsxs)("div", { children: [
					pf(n.total_duration_ms),
					/* @__PURE__ */ (0, L.jsxs)("div", {
						className: "text-xs text-muted-foreground",
						children: [
							n.query_count,
							" ",
							n.query_count === 1 ? "query" : "queries",
							" · ",
							n.slow_query_count,
							" slow"
						]
					}),
					/* @__PURE__ */ (0, L.jsxs)("div", {
						className: "text-xs text-muted-foreground",
						children: ["Max ", pf(n.max_query_ms)]
					}),
					/* @__PURE__ */ (0, L.jsx)(gf, { profile: n })
				] }) : "—";
			}
		}
	];
	return e && n.push({
		key: "baseline",
		label: "Baseline",
		sortable: !0,
		align: "right",
		cellClassName: t,
		render: (e, t) => t.baseline === null ? "—" : pf(t.baseline)
	}, {
		key: "change",
		label: "Change",
		sortable: !0,
		align: "right",
		cellClassName: t,
		render: (e, t) => {
			let n = t.delta;
			return n ? /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [/* @__PURE__ */ (0, L.jsx)("span", {
				className: P(n.delta_ms > 0 && "text-red-700 dark:text-red-400", n.delta_ms < 0 && "text-green-700 dark:text-green-400"),
				children: n.delta_pct === void 0 ? "—" : `${n.delta_pct >= 0 ? "+" : ""}${n.delta_pct.toFixed(1)}%`
			}), n.disk_read_bytes_delta !== void 0 && n.disk_write_bytes_delta !== void 0 && /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "text-xs text-muted-foreground",
				children: [
					"Disk R ",
					hf(n.disk_read_bytes_delta),
					" · W ",
					hf(n.disk_write_bytes_delta)
				]
			})] }) : "—";
		}
	}), n;
}
function ap({ title: e, entries: t, deltas: n }) {
	let r = (0, I.useMemo)(() => rp(t, n), [t, n]), i = (0, I.useMemo)(() => ip(!!n), [n]);
	return t.length ? /* @__PURE__ */ (0, L.jsxs)("section", {
		className: "space-y-1",
		children: [/* @__PURE__ */ (0, L.jsx)("h2", {
			className: "text-sm font-semibold",
			children: e
		}), /* @__PURE__ */ (0, L.jsx)(o, {
			data: r,
			columns: i,
			emptyMessage: `No ${e.toLowerCase()}`
		})]
	}) : null;
}
function op(e) {
	let t = new Set(e.processes.map((e) => e.pid));
	return e.processes.map((n) => ({
		key: String(n.pid),
		...n.ppid !== void 0 && t.has(n.ppid) ? { parentKey: String(n.ppid) } : {},
		pid: n.pid,
		ppid: n.ppid ?? 0,
		name: n.command?.split(/\s+/, 1)[0] ?? "",
		command: n.command || `PID ${n.pid}`,
		status: n.status || "active",
		startedAt: e.sampled_at,
		lastSeenAt: e.sampled_at,
		cpuPercent: n.cpuPercent ?? 0,
		rssBytes: n.rssBytes ?? 0,
		peakCpuPercent: n.cpuPercent ?? 0,
		peakRssBytes: n.rssBytes ?? 0
	}));
}
function sp({ performance: e, sampleIndex: t, onSampleChange: n }) {
	let r = e.profile, i = r?.samples ?? [], a = t !== void 0 && t < i.length ? t : 0, o = i[a], s = (0, I.useMemo)(() => o ? op(o) : [], [o]);
	return /* @__PURE__ */ (0, L.jsxs)("section", {
		className: "space-y-3",
		children: [/* @__PURE__ */ (0, L.jsx)("h2", {
			className: "text-sm font-semibold",
			children: "Process profile"
		}), r ? i.length === 0 ? /* @__PURE__ */ (0, L.jsx)("p", {
			className: "text-sm text-muted-foreground",
			children: "No process samples were observed."
		}) : /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
			/* @__PURE__ */ (0, L.jsxs)("div", {
				className: "grid grid-cols-2 gap-density-1 sm:grid-cols-5",
				children: [
					/* @__PURE__ */ (0, L.jsx)(M, {
						label: "Peak CPU",
						value: `${r.peak_cpu_percent.toFixed(1)}%`
					}),
					/* @__PURE__ */ (0, L.jsx)(M, {
						label: "Peak RSS",
						value: mf(r.peak_rss_bytes)
					}),
					r.disk_io && /* @__PURE__ */ (0, L.jsx)(M, {
						label: "Observed disk read",
						value: mf(r.disk_io.disk_read_bytes)
					}),
					r.disk_io && /* @__PURE__ */ (0, L.jsx)(M, {
						label: "Observed disk write",
						value: mf(r.disk_io.disk_write_bytes)
					}),
					/* @__PURE__ */ (0, L.jsx)(M, {
						label: "Samples",
						value: `${i.length} at ${r.interval_ms} ms`
					})
				]
			}),
			/* @__PURE__ */ (0, L.jsxs)("div", {
				className: "grid grid-cols-2 gap-density-2",
				children: [/* @__PURE__ */ (0, L.jsx)(ql, {
					label: "CPU %",
					values: i.map((e) => e.cpu_percent),
					format: (e) => `${e.toFixed(1)}%`
				}), /* @__PURE__ */ (0, L.jsx)(ql, {
					label: "RSS",
					values: i.map((e) => e.rss_bytes),
					format: mf
				})]
			}),
			/* @__PURE__ */ (0, L.jsx)("div", {
				className: "flex flex-wrap gap-1",
				"aria-label": "Profile samples",
				children: i.map((e, t) => /* @__PURE__ */ (0, L.jsx)(be, {
					type: "button",
					size: "sm",
					variant: t === a ? "default" : "outline",
					"aria-label": `Sample ${t + 1}`,
					onClick: () => n(t),
					children: new Date(e.sampled_at).toLocaleTimeString()
				}, e.sampled_at))
			}),
			o && /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "space-y-1",
				children: [/* @__PURE__ */ (0, L.jsxs)("div", {
					className: "text-sm text-muted-foreground",
					children: [
						"Sample ",
						a + 1,
						" · CPU ",
						o.cpu_percent.toFixed(1),
						"% · RSS ",
						mf(o.rss_bytes)
					]
				}), /* @__PURE__ */ (0, L.jsx)("div", {
					className: "h-64 rounded border border-border",
					children: /* @__PURE__ */ (0, L.jsx)(Zl, {
						className: "h-full",
						nodes: s,
						selectedKey: null,
						expandAll: !0,
						onSelect: () => {},
						emptyLabel: "No processes were sampled"
					})
				})]
			})
		] }) : /* @__PURE__ */ (0, L.jsx)("p", {
			className: "text-sm text-muted-foreground",
			children: "Sampling in progress. Profile appears when the run finishes."
		})]
	});
}
function cp({ performance: e, tests: t, error: n, sampleIndex: r, onSampleChange: i }) {
	let { phases: a, fixtures: o } = (0, I.useMemo)(() => np(t), [t]), s = e.comparison, c = e.profile;
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "mx-auto max-w-7xl space-y-5 p-density-4",
		children: [
			/* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex flex-wrap items-baseline gap-3",
				children: [
					/* @__PURE__ */ (0, L.jsxs)("h2", {
						className: "text-lg font-semibold",
						children: ["Fixture ", e.mode === "profile" ? "profile" : "benchmark"]
					}),
					/* @__PURE__ */ (0, L.jsx)("span", {
						className: "text-sm text-muted-foreground",
						children: e.status === "running" ? "Running" : `${e.status} · ${pf(e.duration_ms)}`
					}),
					e.artifact_path && /* @__PURE__ */ (0, L.jsx)("span", {
						className: "break-all font-mono text-xs text-muted-foreground",
						children: e.artifact_path
					})
				]
			}),
			n && /* @__PURE__ */ (0, L.jsx)("div", {
				role: "alert",
				children: /* @__PURE__ */ (0, L.jsx)(Se, {
					variant: "caution",
					children: n
				})
			}),
			s && /* @__PURE__ */ (0, L.jsx)(Se, {
				variant: "note",
				children: /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "space-y-1 text-sm",
					children: [
						/* @__PURE__ */ (0, L.jsxs)("div", { children: [
							"Compared with ",
							/* @__PURE__ */ (0, L.jsx)("span", {
								className: "font-mono",
								children: s.baseline
							}),
							". Changes describe these two runs only."
						] }),
						!!s.added?.length && /* @__PURE__ */ (0, L.jsxs)("div", { children: ["Added: ", s.added.join(", ")] }),
						!!s.missing?.length && /* @__PURE__ */ (0, L.jsxs)("div", { children: ["Missing: ", s.missing.join(", ")] }),
						!!s.unmeasured?.length && /* @__PURE__ */ (0, L.jsxs)("div", { children: ["Unmeasured: ", s.unmeasured.join(", ")] }),
						(s.peak_cpu_percent_delta !== void 0 || s.peak_rss_bytes_delta !== void 0) && /* @__PURE__ */ (0, L.jsxs)("div", { children: [
							"Peak change: CPU ",
							s.peak_cpu_percent_delta === void 0 ? "—" : `${s.peak_cpu_percent_delta >= 0 ? "+" : ""}${s.peak_cpu_percent_delta.toFixed(1)} points`,
							"; RSS ",
							s.peak_rss_bytes_delta === void 0 ? "—" : hf(s.peak_rss_bytes_delta)
						] })
					]
				})
			}),
			/* @__PURE__ */ (0, L.jsx)(ap, {
				title: "Run phases",
				entries: a,
				deltas: s?.deltas
			}),
			/* @__PURE__ */ (0, L.jsx)(ap, {
				title: "Fixture steps",
				entries: o,
				deltas: s?.deltas
			}),
			e.mode === "profile" || c ? /* @__PURE__ */ (0, L.jsx)(sp, {
				performance: e,
				sampleIndex: r,
				onSampleChange: i
			}) : null,
			!a.length && !o.length && /* @__PURE__ */ (0, L.jsx)("p", {
				className: "text-sm text-muted-foreground",
				children: "Waiting for setup or fixture steps…"
			})
		]
	});
}
// src/utils.ts
var lp = "Running tests across packages";
function up(e, t) {
	return e.framework === "task" && e.name === t;
}
function dp(e) {
	return e.filter((e) => !up(e, lp));
}
function fp(e) {
	return e.context?.status;
}
function pp(e) {
	let t = fp(e);
	return t === "running" || t === "pending";
}
function mp(e) {
	let t = e.find((e) => up(e, lp));
	return !!t && pp(t);
}
function hp(e, t) {
	return t && !mp(e);
}
function gp(e, t) {
	return t && !mp(e) && e.every((e) => e.framework === "task");
}
function _p(e) {
	let t = {
		running: 0,
		done: 0,
		failed: 0
	};
	for (let n of e ?? []) n.status ? t.running++ : !n.skipped && (n.error || n.timed_out || !n.success && !n.violations?.length) ? t.failed++ : t.done++;
	return t;
}
function vp(e) {
	return e.filter((e) => e.framework === "task" && e.name !== "Running tests across packages" && e.name !== "Running linters" && pp(e)).map((e) => ({
		...e,
		children: (e.children ?? []).filter(pp)
	}));
}
function yp(e, t) {
	if (!e) return null;
	let n = e[t];
	if (typeof n != "string") return null;
	let r = n.trim();
	return r === "" ? null : r;
}
function bp(e) {
	let t = /* @__PURE__ */ new Map(), n = (e) => {
		for (let r of e) {
			for (let e of r.fixture?.pids || []) !t.has(e) && r.route_path && t.set(e, {
				pid: e,
				routePath: r.route_path,
				name: r.name
			});
			r.children && n(r.children);
		}
	};
	return n(e), t;
}
// src/components/LintTasks.tsx
function xp(e) {
	return String(e.context?.status ?? "");
}
function Sp({ tests: e }) {
	let t = vp(e);
	return t.length === 0 ? null : /* @__PURE__ */ (0, L.jsxs)("section", {
		"aria-label": "Tasks",
		className: "flex flex-col gap-1 border-b border-border px-density-3 py-density-2 text-xs",
		children: [/* @__PURE__ */ (0, L.jsx)("h3", {
			className: "text-xs font-medium uppercase tracking-wide text-muted-foreground",
			children: "Tasks"
		}), t.map((e) => {
			let t = e.children?.length ? e.children : [e];
			return /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex flex-col gap-0.5",
				children: [/* @__PURE__ */ (0, L.jsx)("span", {
					className: "text-muted-foreground",
					children: e.name
				}), /* @__PURE__ */ (0, L.jsx)("ul", {
					className: "flex flex-col gap-0.5",
					children: t.map((e) => {
						let t = xp(e) === "running";
						return /* @__PURE__ */ (0, L.jsxs)("li", {
							className: "flex items-center gap-2",
							children: [
								/* @__PURE__ */ (0, L.jsx)(F, {
									icon: t ? ht : _t,
									className: t ? "animate-spin text-blue-500" : "text-muted-foreground"
								}),
								/* @__PURE__ */ (0, L.jsx)("span", {
									className: "truncate font-medium",
									children: e.name
								}),
								/* @__PURE__ */ (0, L.jsx)("span", {
									className: "truncate text-muted-foreground",
									children: e.message || xp(e)
								})
							]
						}, e.task_id ?? e.name);
					})
				})]
			}, e.name);
		})]
	});
}
// src/components/RerunDialog.tsx
var Cp = 20, wp = 250, Tp = {
	running: "info",
	success: "success",
	failed: "danger",
	canceled: "warning"
};
function Ep(e) {
	return e.stream === "stderr" ? `\x1b[31m${e.text}\x1b[0m` : e.text;
}
function Dp({ open: e, onClose: t, streamUrl: n }) {
	let [r, i] = (0, I.useState)([]), [a, o] = (0, I.useState)("running"), [s, c] = (0, I.useState)(""), [l, u] = (0, I.useState)(void 0), d = (0, I.useRef)(null), f = (0, I.useRef)(!0), p = (0, I.useRef)("running"), m = (0, I.useCallback)(() => {
		let e = d.current;
		e && (f.current = e.scrollHeight - e.scrollTop - e.clientHeight < 40);
	}, []), [h, g] = (0, I.useState)(0), _ = (0, I.useRef)(null);
	return (0, I.useEffect)(() => {
		e && (o("running"), p.current = "running", c(""), u(void 0), g(0), f.current = !0);
	}, [e, n]), (0, I.useEffect)(() => {
		if (!e) return;
		let t = new EventSource(n);
		return t.addEventListener("open", () => i([])), t.addEventListener("message", (e) => {
			let t = JSON.parse(e.data);
			t.command && c(t.command), p.current = t.status, o(t.status), t.lines?.length && i((e) => [...e, ...t.lines]);
		}), t.addEventListener("done", () => t.close()), t.onerror = () => {
			p.current === "running" ? t.readyState === EventSource.CLOSED && (h >= Cp ? u("The rerun output is not available") : _.current = window.setTimeout(() => g((e) => e + 1), wp)) : t.close();
		}, () => {
			t.close(), _.current !== null && window.clearTimeout(_.current);
		};
	}, [
		e,
		n,
		h
	]), (0, I.useEffect)(() => {
		f.current && d.current && (d.current.scrollTop = d.current.scrollHeight);
	}, [r]), /* @__PURE__ */ (0, L.jsx)(Te, {
		open: e,
		onClose: t,
		size: "xl",
		closeOnBackdrop: !0,
		scrollBody: !1,
		title: /* @__PURE__ */ (0, L.jsx)("span", {
			className: "truncate font-mono text-sm",
			title: s || "rerun",
			children: s || "rerun"
		}),
		headerSlot: /* @__PURE__ */ (0, L.jsx)("span", {
			"data-rerun-status": a,
			children: /* @__PURE__ */ (0, L.jsx)(Oe, {
				tone: Tp[a],
				variant: "soft",
				size: "xs",
				children: a
			})
		}),
		children: /* @__PURE__ */ (0, L.jsxs)("pre", {
			ref: d,
			onScroll: m,
			className: "h-[70vh] overflow-auto bg-zinc-950 p-4 font-mono text-sm leading-relaxed text-zinc-100",
			"data-rerun-output": "",
			children: [
				r.length === 0 && a === "running" && !l && /* @__PURE__ */ (0, L.jsx)("span", {
					className: "text-zinc-500",
					children: "Waiting for output..."
				}),
				r.length === 0 && a !== "running" && !l && /* @__PURE__ */ (0, L.jsx)("span", {
					className: "text-zinc-500",
					children: "The rerun wrote no output."
				}),
				l && /* @__PURE__ */ (0, L.jsx)("span", {
					className: "text-red-400",
					children: l
				}),
				/* @__PURE__ */ (0, L.jsx)(ne, {
					as: "span",
					text: r.map(Ep).join("")
				})
			]
		})
	});
}
// src/components/RunActions.tsx
var Op = [
	{
		key: "timeout",
		label: "global",
		title: "Global --timeout"
	},
	{
		key: "test_timeout",
		label: "per-test",
		title: "--test-timeout (per-package)"
	},
	{
		key: "lint_timeout",
		label: "per-lint",
		title: "--lint-timeout (per-linter)"
	}
];
function kp({ runMeta: e }) {
	let t = Op.flatMap((t) => {
		let n = yp(e?.args, t.key);
		return n ? [{
			...t,
			value: n
		}] : [];
	});
	return t.length === 0 ? null : /* @__PURE__ */ (0, L.jsxs)("span", {
		className: "inline-flex items-center gap-1",
		"data-timeout-pills": "",
		children: [/* @__PURE__ */ (0, L.jsx)(y, { className: "text-muted-foreground" }), t.map((e) => /* @__PURE__ */ (0, L.jsx)("span", {
			title: e.title,
			children: /* @__PURE__ */ (0, L.jsxs)(Oe, {
				tone: "neutral",
				variant: "outline",
				size: "xs",
				children: [
					e.label,
					" ",
					e.value
				]
			})
		}, e.key))]
	});
}
function Ap(e) {
	return [
		e.pid ? `PID ${e.pid}` : "",
		e.frameworks?.length ? `Frameworks ${e.frameworks.join(", ")}` : "",
		e.command ? `Command ${e.command}` : ""
	].filter(Boolean).join("\n");
}
function jp({ runMeta: e, canStop: t, stopping: n, onStop: r, exportState: i }) {
	let a = e ? Ap(e) : "";
	return /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
		a && /* @__PURE__ */ (0, L.jsx)("span", {
			className: "hidden text-xs text-muted-foreground lg:inline",
			title: a,
			children: "run info"
		}),
		/* @__PURE__ */ (0, L.jsx)(kp, { runMeta: e }),
		t && /* @__PURE__ */ (0, L.jsxs)(be, {
			type: "button",
			size: "sm",
			variant: "destructive",
			disabled: n,
			onClick: r,
			title: "Stop the active run",
			children: [/* @__PURE__ */ (0, L.jsx)(Ce, { className: "mr-1" }), n ? "Stopping..." : "Stop"]
		}),
		i && /* @__PURE__ */ (0, L.jsx)(rf, {
			routeState: i,
			title: "Download the whole result as JSON or Markdown"
		}),
		i && /* @__PURE__ */ (0, L.jsx)(sf, { routeState: i })
	] });
}
// src/runHistory.ts
async function Mp({ before: e, signal: t } = {}) {
	let n = e ? `?before=${encodeURIComponent(e)}` : "", r = await fetch(zd(`/api/runs${n}`), t ? { signal: t } : {});
	if (r.status === 404) return null;
	if (!r.ok) throw Error(`GET /api/runs: ${r.status} ${(await r.text()).trim()}`);
	return await r.json();
}
function Np(e) {
	return e.kind === "lint" ? `${e.href}lint` : e.href;
}
function Pp(e) {
	return e.started ? new Date(e.started).toLocaleString(void 0, {
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit"
	}) : e.runId;
}
function Fp(e) {
	let t = [];
	return e.total > 0 && t.push(e.failed > 0 ? `${e.failed} failed / ${e.total} tests` : `${e.total} tests`), e.kind !== "test" && t.push(`${e.lintViolations} lint`), t.join(" · ") || "no results";
}
function Ip(e) {
	if (!e.started || !e.ended) return;
	let t = Date.parse(e.ended) - Date.parse(e.started);
	return t >= 0 ? t : void 0;
}
// src/components/RunHistory.tsx
function Lp({ run: e }) {
	return /* @__PURE__ */ (0, L.jsxs)("span", {
		className: "inline-flex flex-wrap items-center gap-1",
		children: [
			e.total > 0 && /* @__PURE__ */ (0, L.jsx)(Oe, {
				tone: e.failed > 0 ? "danger" : "success",
				variant: "outline",
				size: "xs",
				children: e.failed > 0 ? `${e.failed} failed` : `${e.passed} passed`
			}),
			e.total > 0 && /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "text-xs text-muted-foreground",
				children: ["of ", e.total]
			}),
			e.kind !== "test" && /* @__PURE__ */ (0, L.jsxs)(Oe, {
				tone: e.lintViolations > 0 ? "warning" : "neutral",
				variant: "outline",
				size: "xs",
				children: [e.lintViolations, " lint"]
			})
		]
	});
}
function Rp({ history: e, navigate: t }) {
	let n = (0, I.useMemo)(() => [
		{
			key: "started",
			label: "Started",
			kind: "timestamp",
			sortable: !0,
			render: (t, n) => /* @__PURE__ */ (0, L.jsxs)("span", {
				className: "inline-flex items-center gap-1.5",
				children: [
					/* @__PURE__ */ (0, L.jsx)("a", {
						href: Np(n),
						className: "text-primary no-underline hover:underline",
						onClick: (e) => e.stopPropagation(),
						children: n.started ? new Date(n.started).toLocaleString() : n.runId
					}),
					e.current === n.runId && /* @__PURE__ */ (0, L.jsx)(Oe, {
						tone: "info",
						variant: "soft",
						size: "xs",
						children: "viewing"
					}),
					!e.current && e.latest.includes(n.runId) && /* @__PURE__ */ (0, L.jsx)(Oe, {
						tone: "info",
						variant: "soft",
						size: "xs",
						children: "latest"
					})
				]
			})
		},
		{
			key: "kind",
			label: "Kind",
			filterable: !0,
			shrink: !0
		},
		{
			key: "results",
			label: "Results",
			accessor: (e) => e.failed + e.lintViolations,
			sortable: !0,
			render: (e, t) => /* @__PURE__ */ (0, L.jsx)(Lp, { run: t })
		},
		{
			key: "duration",
			label: "Duration",
			accessor: (e) => Ip(e) ?? null,
			sortable: !0,
			align: "right",
			shrink: !0,
			render: (e) => typeof e == "number" ? ae(e) : ""
		},
		{
			key: "frameworks",
			label: "Frameworks",
			kind: "tags",
			accessor: (e) => e.frameworks ?? []
		},
		{
			key: "sha",
			label: "Commit",
			shrink: !0,
			render: (e) => /* @__PURE__ */ (0, L.jsx)("code", {
				className: "text-xs",
				children: String(e ?? "").slice(0, 8)
			})
		}
	], [e.current, e.latest]);
	return /* @__PURE__ */ (0, L.jsx)(o, {
		data: e.runs,
		columns: n,
		getRowId: (e) => e.runId,
		onRowClick: (e) => t(Np(e)),
		getRowClassName: (t) => e.current === t.runId ? "bg-primary/5" : void 0,
		autoFilter: !0,
		infinite: {
			hasMore: e.hasMore,
			loading: e.loading,
			onLoadMore: e.loadMore
		},
		loading: e.loading && e.runs.length === 0,
		loadingMessage: "Loading run history",
		error: e.error,
		emptyMessage: "No runs recorded in .gavel yet"
	});
}
// src/components/RunProcesses.tsx
function zp({ trace: e, selectedKey: t, onSelect: n, hideExited: r, onHideExitedChange: i, expandAll: a, onExpandAllChange: o, stepLinks: s, onOpenStep: c, className: u }) {
	let d = (0, I.useMemo)(() => e.nodes.find((e) => e.key === t) ?? null, [e.nodes, t]), f = d ? s?.get(d.pid) : void 0, p = e.canCollect && id(e.samples).length > 0, m = !!d && e.goprof.some((e) => e.key === d.key);
	return /* @__PURE__ */ (0, L.jsx)(rt, {
		className: u ?? "h-full min-h-0",
		defaultSplit: 45,
		minLeft: 25,
		minRight: 30,
		leftClass: "border-r-2 border-border bg-muted/30",
		left: /* @__PURE__ */ (0, L.jsx)(Zl, {
			className: "h-full",
			nodes: e.nodes,
			samples: e.samples,
			selectedKey: t,
			onSelect: n,
			expandAll: a,
			onExpandAllChange: o,
			hideExited: r,
			onHideExitedChange: i,
			loading: e.loading,
			...e.error ? { error: e.error } : {}
		}),
		right: /* @__PURE__ */ (0, L.jsxs)("div", {
			className: "flex h-full min-h-0 flex-col",
			children: [
				d && p && /* @__PURE__ */ (0, L.jsxs)("div", {
					className: "flex shrink-0 items-center justify-end gap-2 border-b border-border px-density-4 py-density-1",
					children: [e.refreshError && /* @__PURE__ */ (0, L.jsx)("span", {
						className: "min-w-0 truncate text-xs text-red-700 dark:text-red-400",
						role: "alert",
						title: e.refreshError,
						children: e.refreshError
					}), /* @__PURE__ */ (0, L.jsxs)(be, {
						type: "button",
						size: "sm",
						variant: "outline",
						disabled: e.refreshing,
						title: "Read the open files, sockets, ports and connections of this process in full on the next sample",
						onClick: () => void e.refreshFacets(d.key),
						children: [e.refreshing ? /* @__PURE__ */ (0, L.jsx)(ht, { className: "mr-1 animate-spin" }) : /* @__PURE__ */ (0, L.jsx)(l, { className: "mr-1" }), "Refresh facets"]
					})]
				}),
				/* @__PURE__ */ (0, L.jsx)(Sd, {
					className: "min-h-0 flex-1",
					node: d,
					samples: e.samples,
					stacks: e.stacks,
					...e.canCollect ? { onCollectStack: e.collectStack } : {},
					collecting: e.collecting,
					...e.collectError ? { collectError: e.collectError } : {},
					...f && c ? { links: /* @__PURE__ */ (0, L.jsx)(Md, {
						title: "Step",
						children: /* @__PURE__ */ (0, L.jsxs)(be, {
							type: "button",
							variant: "link",
							size: "sm",
							className: "h-auto p-0",
							onClick: () => c(f),
							children: [/* @__PURE__ */ (0, L.jsx)(et, { className: "mr-1" }), f.name]
						})
					}) } : {}
				}),
				d && m && /* @__PURE__ */ (0, L.jsx)("div", {
					className: "max-h-[45%] shrink-0 overflow-y-auto border-t border-border px-density-4 py-density-2",
					"data-go-profiles": "",
					children: /* @__PURE__ */ (0, L.jsxs)(Xc, {
						title: "Go profiles",
						defaultOpen: !0,
						children: [/* @__PURE__ */ (0, L.jsx)(Bl, {
							records: e.goprof,
							processKey: d.key
						}), e.goprofExportUrl && /* @__PURE__ */ (0, L.jsx)("a", {
							className: "mt-2 inline-block text-xs text-primary hover:underline",
							href: e.goprofExportUrl,
							download: !0,
							children: "Download pprof"
						})]
					})
				})
			]
		})
	});
}
// src/components/RunSwitcher.tsx
var Bp = "";
function Vp({ history: e, navigate: t }) {
	let n = e.current ?? Bp, r = [{
		value: Bp,
		label: "Latest results"
	}, ...e.runs.map((t) => ({
		value: t.runId,
		label: `${Pp(t)} · ${t.kind} · ${Fp(t)}${e.latest.includes(t.runId) ? " (latest)" : ""}`
	}))];
	return n !== Bp && !e.runs.some((e) => e.runId === n) && r.splice(1, 0, {
		value: n,
		label: n
	}), /* @__PURE__ */ (0, L.jsxs)("label", {
		className: "inline-flex min-w-0 items-center gap-1 text-sm",
		"data-run-switcher": "",
		children: [
			/* @__PURE__ */ (0, L.jsx)(rn, {
				className: "shrink-0 text-muted-foreground",
				"aria-hidden": !0
			}),
			/* @__PURE__ */ (0, L.jsx)("span", {
				className: "sr-only",
				children: "Run"
			}),
			/* @__PURE__ */ (0, L.jsx)(u, {
				className: "h-7 max-w-[22rem] truncate py-0 text-xs",
				value: n,
				options: r,
				onChange: (r) => {
					let i = r.target.value;
					if (i === n) return;
					let a = e.runs.find((e) => e.runId === i);
					t(i === Bp ? e.latestHref : a ? Np(a) : e.latestHref);
				}
			})
		]
	});
}
// src/processMetrics.ts
var Hp = (e) => e.status === "exited" || !!e.exitedAt;
function Up({ samples: e, nodes: t }) {
	return async (n) => {
		let r = t.filter((e) => e.pid === n);
		if (r.length > 0 && r.every(Hp)) throw new se(n);
		let i = e.filter((e) => e.pid === n).sort((e, t) => t.tick - e.tick)[0];
		if (!i) throw Error(`no process sample for pid ${n} yet`);
		return {
			pid: n,
			rss: i.rssBytes,
			...i.cpuPercent == null ? {} : { cpu_percent: i.cpuPercent },
			...i.openFiles === void 0 ? {} : { open_files: i.openFiles }
		};
	};
}
// src/hooks/use-process-trace.ts
var Wp = 1e3, Gp = 2e3, Kp = "The run has ended, so its process trace can no longer be checkpointed";
async function qp(e, t) {
	try {
		return await Af(e, t);
	} catch (e) {
		if (e instanceof bf) return [];
		throw e;
	}
}
async function Jp(e, t) {
	let n = await Mf(e, t);
	return n === 0 ? [] : qp(e, {
		...t,
		afterSeq: Math.max(0, n - Gp)
	});
}
function Yp(e) {
	return e instanceof Error ? e.message : String(e);
}
function Xp(e, t, n, r) {
	let [i, a] = (0, I.useState)(null), [o, s] = (0, I.useState)(void 0), c = (0, I.useRef)(n);
	(0, I.useEffect)(() => {
		c.current = n;
	});
	let l = (0, I.useCallback)(async () => {
		try {
			a(await c.current()), s(void 0);
		} catch (e) {
			s(Yp(e));
		}
	}, []);
	return (0, I.useEffect)(() => {
		if (!e) return;
		let n = !1, r = () => {
			c.current().then((e) => {
				n || (a(e), s(void 0));
			}).catch((e) => {
				n || s(Yp(e));
			});
		};
		r();
		let i = t ? window.setInterval(r, Wp) : void 0;
		return () => {
			n = !0, i !== void 0 && window.clearInterval(i);
		};
	}, [
		e,
		t,
		r
	]), {
		value: i,
		error: o,
		reload: l
	};
}
function Zp(e, t, n) {
	let [r, i] = (0, I.useState)(!1), [a, o] = (0, I.useState)(void 0);
	return {
		busy: r,
		error: a,
		checkpoint: (0, I.useCallback)(async (r) => {
			i(!0), o(void 0);
			try {
				if (!e || !t) {
					o(Kp);
					return;
				}
				await Pf({
					mount: e,
					session: t,
					params: r
				}), await n();
			} catch (e) {
				o(Yp(e));
			} finally {
				i(!1);
			}
		}, [
			e,
			t,
			n
		])
	};
}
function Qp({ traces: e, done: t, selectedKey: n, base: r }) {
	let i = r ?? Rd(), a = (0, I.useMemo)(() => Cf(e, i), [e, i]), o = Sf(a), s = a?.process, c = !t, l = !!o && !!s, u = c ? e?.sessions : void 0, d = Xp(l, c, () => qp(o, s.tree), `${o}|${s?.tree.stream}`), f = Xp(l, c, () => Jp(o, s.samples), `${o}|${s?.samples.stream}`), p = Xp(l && !!n, c, () => qp(o, s.stacks), `${o}|${s?.stacks.stream}`), m = Xp(l && !!n && !!s?.goprof, c, () => qp(o, s.goprof), `${o}|${s?.goprof?.stream}`), h = Zp(o, u?.stacks, p.reload), g = Zp(o, u?.samples, f.reload), _ = (0, I.useCallback)((e, t) => h.checkpoint({
		key: e,
		transport: t,
		reason: "manual"
	}), [h.checkpoint]), v = (0, I.useCallback)((e) => g.checkpoint({ key: e }), [g.checkpoint]);
	return {
		nodes: (0, I.useMemo)(() => d.value ?? [], [d.value]),
		samples: (0, I.useMemo)(() => f.value ?? [], [f.value]),
		stacks: (0, I.useMemo)(() => p.value ?? [], [p.value]),
		goprof: (0, I.useMemo)(() => m.value ?? [], [m.value]),
		...o && s?.goprof ? { goprofExportUrl: Ef(o, s.goprof, "pprof") } : {},
		loading: l && d.value === null && !d.error,
		error: d.error ?? f.error ?? p.error ?? m.error,
		canCollect: l && !!u,
		collecting: h.busy,
		...h.error ? { collectError: h.error } : {},
		collectStack: _,
		refreshing: g.busy,
		...g.error ? { refreshError: g.error } : {},
		refreshFacets: v
	};
}
// src/hooks/use-run-history.ts
function $p() {
	let [e, t] = (0, I.useState)([]), [n, r] = (0, I.useState)(!1), [i, a] = (0, I.useState)(!0), [o, s] = (0, I.useState)(), c = (0, I.useRef)(!1), l = (0, I.useCallback)(async (e) => {
		if (!c.current) {
			c.current = !0, a(!0);
			try {
				let n = await Mp(e ? { before: e } : {});
				n && (r(!0), t((t) => e ? [...t, n] : [n])), s(void 0);
			} catch (e) {
				s(e instanceof Error ? e.message : String(e));
			} finally {
				c.current = !1, a(!1);
			}
		}
	}, []);
	(0, I.useEffect)(() => {
		l();
	}, [l]);
	let u = e[0], d = e.at(-1)?.next, f = (0, I.useCallback)(() => {
		d && l(d);
	}, [d, l]);
	return {
		available: n,
		...u?.current ? { current: u.current } : {},
		latest: u?.latest ?? [],
		latestHref: u?.latestHref ?? "/",
		runs: e.flatMap((e) => e.runs),
		hasMore: !!d,
		loading: i,
		...o ? { error: o } : {},
		loadMore: f
	};
}
// src/hooks/use-test-run.ts
var em = 2e3, tm = { running: !1 };
function nm(e, t) {
	return t.aborted ? Promise.resolve() : new Promise((n) => {
		let r = () => {
			clearTimeout(i), n();
		}, i = setTimeout(() => {
			t.removeEventListener("abort", r), n();
		}, e);
		t.addEventListener("abort", r, { once: !0 });
	});
}
function rm(e) {
	return new EventSource(e);
}
function im() {
	return typeof window < "u" && window.__gavelBasePath || "";
}
function am() {
	return typeof globalThis < "u" && globalThis.EventSource !== void 0;
}
function om(e) {
	let t = e.status || tm, n = e.tests || [];
	if (!t.running) {
		if (t.stopped) return t.stop_message || "Stopped by user";
		if (e.metadata?.kind === "rerun") return "Rerun complete";
		if (gp(n, !!t.lint_run)) {
			let { done: t, failed: n } = _p(e.lint);
			return `Lint run complete (${t} done · ${n} failed)`;
		}
		return "Test run complete";
	}
	if (e.metadata?.kind === "rerun") return `Running rerun #${e.metadata.sequence || 1}...`;
	if (e.performance) return "Running fixtures...";
	if (hp(n, t.running)) {
		if (!e.lint?.length) return "Running linters...";
		let { running: t, done: n, failed: r } = _p(e.lint);
		return `Running linters... (${t} running · ${n} done · ${r} failed)`;
	}
	return "Running tests...";
}
function sm(e = {}) {
	let { baseUrl: t, basePath: n, runId: r, enabled: i = !0, pollMs: a = em, forcePoll: o = !1, createEventSource: s } = e, c = t ?? n ?? im(), l = r ? `${c}/${encodeURIComponent(r)}` : c, [u, d] = (0, I.useState)(void 0), [f, p] = (0, I.useState)("Loading..."), [m, h] = (0, I.useState)(!1), [g, _] = (0, I.useState)(void 0), [v, y] = (0, I.useState)(0), b = (0, I.useRef)(!1), x = (0, I.useRef)(s ?? rm);
	(0, I.useEffect)(() => {
		x.current = s ?? rm;
	});
	let S = (0, I.useCallback)((e) => {
		let t = !!(e.status && e.status.running);
		d(e), p(om(e)), _(void 0), t ? (b.current = !1, h(!1)) : (b.current = !0, h(!0));
	}, []);
	(0, I.useEffect)(() => {
		if (!i) return;
		if (b.current = !1, h(!1), o || !am()) {
			let e = new AbortController();
			return (async () => {
				for (; !e.signal.aborted;) {
					try {
						let t = await fetch(`${l}/api/tests`, {
							headers: { Accept: "application/json" },
							signal: e.signal
						});
						if (e.signal.aborted) return;
						if (t.ok) {
							let n = await t.json();
							if (e.signal.aborted || (S(n), !n.status?.running)) return;
						}
					} catch {
						!e.signal.aborted && !b.current && _("Connection lost — retrying...");
					}
					await nm(a, e.signal);
				}
			})(), () => e.abort();
		}
		fetch(`${l}/api/tests`, { headers: { Accept: "application/json" } }).then((e) => e.json()).then((e) => S(e)).catch(() => {});
		let e = x.current(`${l}/api/tests/stream`), t = (t) => {
			let n = JSON.parse(t.data);
			S(n), n.status?.running || e.close();
		}, n = () => {
			b.current = !0, h(!0), d((e) => e && {
				...e,
				status: {
					...e.status,
					running: !1
				}
			}), e.close();
		};
		return e.addEventListener("message", t), e.addEventListener("done", n), e.onerror = () => {
			b.current || _("Connection lost — retrying...");
		}, () => {
			e.removeEventListener("message", t), e.removeEventListener("done", n), e.close();
		};
	}, [
		l,
		i,
		a,
		o,
		v,
		S
	]);
	let C = (0, I.useCallback)(async (e = {}) => {
		b.current = !1, h(!1), y((e) => e + 1);
		let t = await fetch(`${l}/api/rerun`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(e)
		});
		if (t.status === 409) return {
			ok: !1,
			status: 409,
			conflict: !0,
			message: "Another rerun is in progress"
		};
		if (!t.ok) {
			let e = (await t.text()).trim();
			return _(`Rerun failed: ${e}`), {
				ok: !1,
				status: t.status,
				conflict: !1,
				message: e
			};
		}
		return {
			ok: !0,
			status: t.status,
			conflict: !1
		};
	}, [l]), w = (0, I.useCallback)(async (e) => {
		let t = e ? {
			scope: "task",
			task_id: e
		} : { scope: "global" }, n = await fetch(`${l}/api/stop`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(t)
		});
		if (!n.ok) {
			let e = (await n.text()).trim();
			return _(`Stop failed: ${e}`), {
				ok: !1,
				status: n.status,
				conflict: !1,
				message: e
			};
		}
		return {
			ok: !0,
			status: n.status,
			conflict: !1
		};
	}, [l]), T = (0, I.useCallback)(async (e) => {
		let t = await fetch(`${l}/api/tests/edit`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(e)
		});
		if (!t.ok) {
			let e = (await t.text()).trim();
			return _(`Test edit failed: ${e}`), {
				ok: !1,
				status: t.status,
				conflict: !1,
				message: e
			};
		}
		let n = await t.json();
		return y((e) => e + 1), {
			ok: !0,
			status: t.status,
			conflict: !1,
			...n.message ? { message: n.message } : {}
		};
	}, [l]), E = (0, I.useCallback)(() => y((e) => e + 1), []);
	return {
		snapshot: u,
		tests: u?.tests ?? [],
		lint: u?.lint,
		bench: u?.bench,
		status: u?.status ?? tm,
		statusText: f,
		runMeta: u?.metadata,
		done: m,
		error: g,
		rerun: C,
		stop: w,
		editTest: T,
		refetch: E
	};
}
// src/lintRerun.ts
function cm(e) {
	return e.replace(/\\/g, "/").replace(/\/+/g, "/").replace(/\/$/, "");
}
function lm(e, t) {
	if (!e) return;
	let n = cm(e);
	if (!t) return n;
	let r = `${cm(t)}/`;
	return n.startsWith(r) ? n.slice(r.length) : n;
}
function um(e, t) {
	e.work_dir && t.add(e.work_dir), (e.children ?? []).forEach((e) => um(e, t));
}
function dm(e, t) {
	let n = /* @__PURE__ */ new Set(), r = /* @__PURE__ */ new Set(), i = /* @__PURE__ */ new Set(), a = (a, o, s) => {
		for (let c of t) {
			if (e?.work_dir && c.work_dir && c.work_dir !== e.work_dir || o && c.linter !== o) continue;
			let t = !1;
			for (let e of c.violations ?? []) {
				let n = lm(e.file, c.work_dir);
				s && (e.rule?.method || "(no rule)") !== s || (n ? (!a || n === a || n.startsWith(`${a}/`)) && (t = !0, i.add(n)) : t = s !== void 0 && a === void 0);
			}
			t && (c.work_dir && n.add(c.work_dir), r.add(c.linter));
		}
	};
	if (e) e.kind === "linter" || e.kind === "linter-error" ? (e.linterName && r.add(e.linterName), um(e, n)) : ((e.kind === "lint-rule" || e.kind === "lint-rule-group") && e.linterName && r.add(e.linterName), a(e.target_path, e.linterName, e.ruleName));
	else for (let e of t) r.add(e.linter), e.work_dir && n.add(e.work_dir);
	return e?.work_dir && n.add(e.work_dir), {
		lint: !0,
		work_dir: n.size === 1 ? Array.from(n)[0] : "",
		lint_linters: Array.from(r),
		lint_files: Array.from(i)
	};
}
// src/exportPaths.ts
var fm = "go test";
function pm(e) {
	let t = "";
	for (let n = 0; n < e.length; n++) {
		let r = e[n], i = n > 0 ? e[n - 1] : "";
		n > 0 && /\p{Lu}/u.test(r) && (/\p{Ll}/u.test(i) || /\p{Nd}/u.test(i)) && (t += " "), t += r;
	}
	return t;
}
function mm(e, t) {
	return t !== fm || e.endsWith("/") ? e : e.split("/").map((e, t) => t === 0 ? pm(e.replace(/^Test/, "")) : e.replaceAll("_", " ")).join(" / ");
}
function hm(e) {
	return e.toLowerCase().replace(/[^\p{L}\p{N}]+/gu, "-").replace(/^-+|-+$/g, "") || "node";
}
function gm(e, t = []) {
	let n = e.map((e) => hm(mm(e.name, e.framework))), r = /* @__PURE__ */ new Map();
	n.forEach((e) => r.set(e, (r.get(e) ?? 0) + 1));
	let i = /* @__PURE__ */ new Map();
	return e.map((e, a) => {
		let o = n[a], s = (i.get(o) ?? 0) + 1;
		i.set(o, s);
		let c = [...t, r.get(o) > 1 ? `${o}~${s}` : o];
		return {
			...e,
			export_path: c.join("/"),
			...e.children ? { children: gm(e.children, c) } : {}
		};
	});
}
// src/App.tsx
function _m(e) {
	return e.reduce((e, t) => e + +!!t.fixture + _m(t.children || []), 0);
}
var vm = (e) => e, ym = (0, I.lazy)(() => import("./chunks/FilesTab-ylDZtrKn.js").then((e) => e.t)), bm = (0, I.lazy)(() => import("./chunks/PromptRunView-BqG3fTg5.js").then((e) => e.t));
function xm(e) {
	return e.run_id ? /* @__PURE__ */ (0, L.jsx)(I.Suspense, {
		fallback: /* @__PURE__ */ (0, L.jsx)("p", {
			role: "status",
			className: "text-sm text-muted-foreground",
			children: "Loading the prompt run…"
		}),
		children: /* @__PURE__ */ (0, L.jsx)(bm, {
			result: e,
			baseUrl: Rd()
		})
	}) : null;
}
function Sm() {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex h-full items-center justify-center text-sm text-muted-foreground",
		role: "status",
		children: [/* @__PURE__ */ (0, L.jsx)(F, {
			icon: ht,
			className: "mr-2 animate-spin text-blue-500"
		}), "Loading files..."]
	});
}
function Cm(e, t) {
	return t ? e.map((e) => ({
		...e,
		work_dir: e.work_dir || t,
		...e.children ? { children: Cm(e.children, t) } : {}
	})) : e;
}
function wm(e) {
	return [
		e.workDir,
		e.file,
		e.line,
		e.rule,
		e.message
	].join("\0");
}
function Tm(e, t) {
	if (!e) return;
	if (!t) return e;
	let n = new URL(e);
	return n.searchParams.has("line") || n.searchParams.set("line", String(t)), n.toString();
}
async function Em(e, t) {
	let n = new URLSearchParams();
	e.workDir && n.set("work_dir", e.workDir), n.set("path", e.file);
	let r = await fetch(zd(`/api/file?${n}`, t)), i = await r.text();
	if (!r.ok) throw Error(`${r.status} ${i.trim() || r.statusText}`);
	return i;
}
function Dm(e) {
	return e.map((e) => ({
		...e,
		own_name: e.name,
		...e.children ? { children: Dm(e.children) } : {}
	}));
}
function Om(e) {
	let t = /* @__PURE__ */ new Set(), n = /* @__PURE__ */ new Set(), r = (e) => {
		e.package_path && t.add(e.package_path), e.work_dir && n.add(e.work_dir), (e.children || []).forEach(r);
	};
	if (r(e), n.size > 1) return { error: "Rerun across multiple module roots is not supported" };
	let i = !e.children || e.children.length === 0;
	return { body: {
		package_paths: Array.from(t),
		work_dir: n.size === 1 ? Array.from(n)[0] : "",
		test_name: i ? e.own_name ?? e.name : "",
		suite: e.suite || [],
		framework: e.framework || ""
	} };
}
function km({ icon: e, title: t, status: n, actions: r }) {
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex min-w-0 items-center gap-3 border-b border-border bg-card px-density-4 py-density-2",
		"data-page-header": "",
		children: [
			/* @__PURE__ */ (0, L.jsxs)("h1", {
				className: "inline-flex shrink-0 items-center gap-1.5 text-base font-bold",
				children: [/* @__PURE__ */ (0, L.jsx)(F, {
					icon: e,
					className: "text-primary"
				}), t]
			}),
			n && /* @__PURE__ */ (0, L.jsx)("span", {
				className: "truncate text-sm text-muted-foreground",
				children: n
			}),
			r && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "ml-auto flex shrink-0 items-center gap-1",
				children: r
			})
		]
	});
}
function Am({ done: e }) {
	return /* @__PURE__ */ (0, L.jsx)("div", {
		className: "flex h-full items-center justify-center p-8 text-center text-sm text-muted-foreground",
		children: e ? "No tests were run" : /* @__PURE__ */ (0, L.jsxs)("div", { children: [/* @__PURE__ */ (0, L.jsx)(F, {
			icon: ht,
			className: "animate-spin text-3xl text-blue-500"
		}), /* @__PURE__ */ (0, L.jsx)("p", {
			className: "mt-2",
			children: "Waiting for test results..."
		})] })
	});
}
var jm = 8e3, Mm = "__global__", Nm = (e) => window.location.assign(e);
function Pm({ runId: e, navigate: t = Nm } = {}) {
	let n = (0, I.useMemo)(() => e ? { runId: e } : {}, [e]), r = Rd(n), i = sm(n.runId ? { runId: n.runId } : {}), { snapshot: a, tests: o, lint: s, status: c, runMeta: u, done: d, statusText: f, error: p } = i, m = a?.performance, h = a?.error, g = !!c.lint_run, _ = $p(), [v, y] = (0, I.useState)(() => qd(window.location)), [b, x] = (0, I.useState)(null), [S, C] = (0, I.useState)(!1), [w, T] = (0, I.useState)(!1), [E, D] = (0, I.useState)(!1), [O, ee] = (0, I.useState)(!1), [te, ne] = (0, I.useState)(null), [re, ie] = (0, I.useState)(void 0), k = (0, I.useRef)(null), A = (0, I.useRef)(Date.now()), [ae, oe] = (0, I.useState)(Date.now()), M = (0, I.useCallback)((e) => {
		ie(e), k.current !== null && window.clearTimeout(k.current), k.current = window.setTimeout(() => ie(void 0), jm);
	}, []);
	(0, I.useEffect)(() => () => {
		k.current !== null && window.clearTimeout(k.current);
	}, []), (0, I.useEffect)(() => {
		if (d) return;
		let e = window.setInterval(() => oe(Date.now()), 1e3);
		return () => window.clearInterval(e);
	}, [d]);
	let se = u?.started ? Date.parse(u.started) : A.current, ce = u?.ended ? Date.parse(u.ended) : null, N = (0, I.useCallback)((e, t = "push") => {
		y(e);
		let n = Jd(e);
		n !== `${window.location.pathname}${window.location.search}` && (t === "replace" ? window.history.replaceState({}, "", n) : window.history.pushState({}, "", n));
	}, []);
	(0, I.useEffect)(() => {
		let e = () => y(qd(window.location));
		return window.addEventListener("popstate", e), () => window.removeEventListener("popstate", e);
	}, []), (0, I.useEffect)(() => {
		m && v.tab === "tests" && window.location.pathname === `${Rd()}/` && N({
			...v,
			tab: "bench",
			selectedPath: ""
		}, "replace");
	}, [
		m,
		v,
		N
	]);
	let le = (0, I.useMemo)(() => hp(o, !!c.running) ? dp(o) : o, [o, c.running]), ue = (0, I.useMemo)(() => mt(le), [le]), de = (0, I.useMemo)(() => Dt(ut(gm(ot(Dm(le), v.filters.status, v.filters.framework)))), [le, v.filters]), fe = (0, I.useMemo)(() => v.tab === "tests" && v.selectedPath ? Tt(de, v.selectedPath) : null, [
		v.tab,
		v.selectedPath,
		de
	]), pe = (0, I.useMemo)(() => Et({
		results: s ?? [],
		grouping: v.lintGrouping,
		filters: v.lintFilters
	}), [
		s,
		v.lintGrouping,
		v.lintFilters
	]), me = v.tab === "lint" && v.selectedPath ? Nt(pe, v.selectedPath) : null, ge = (0, I.useMemo)(() => Ft(s ?? []), [s]), _e = a?.git?.root, { findings: ve, linterByFinding: ye, sourceUrlByFinding: P } = (0, I.useMemo)(() => {
		let e = /* @__PURE__ */ new Map(), t = /* @__PURE__ */ new Map();
		return {
			findings: [...(s ?? []).flatMap((n) => {
				let r = n.work_dir || _e;
				return Ot([{
					...n,
					...r ? { work_dir: r } : {}
				}]).map((r) => {
					e.set(wm(r), n.linter);
					let i = Tm(n.sources?.[r.file]?.url, r.line);
					return i && t.set(wm(r), i), r;
				});
			}), ...jt(Cm(o, _e))],
			linterByFinding: e,
			sourceUrlByFinding: t
		};
	}, [
		s,
		o,
		_e
	]), xe = (0, I.useMemo)(() => v.tab === "files" && v.selectedPath ? {
		file: v.selectedPath,
		...v.workDir ? { workDir: v.workDir } : {}
	} : null, [
		v.tab,
		v.selectedPath,
		v.workDir
	]), Se = (0, I.useMemo)(() => Cf(c.traces, r), [c.traces, r]), Ce = !!Sf(c.traces), we = Qp({
		traces: c.traces,
		done: d,
		base: r,
		...v.tab === "processes" && v.selectedPath ? { selectedKey: v.selectedPath } : {}
	}), F = (0, I.useMemo)(() => we.nodes.filter((e) => !jl(e)).length, [we.nodes]), Te = (0, I.useMemo)(() => we.nodes.find((e) => e.isRoot) ?? null, [we.nodes]), Ee = v.tab === "processes" ? we.nodes.find((e) => e.key === v.selectedPath) ?? null : null, De = (0, I.useMemo)(() => Ce ? bp(vm(Dt(ut(le)))) : /* @__PURE__ */ new Map(), [Ce, le]);
	(0, I.useEffect)(() => {
		v.tab === "processes" && !v.selectedPath && Te && N({
			...v,
			selectedPath: Te.key
		}, "replace");
	}, [
		v,
		Te,
		N
	]), (0, I.useEffect)(() => {
		if (!v.selectedPath) return;
		let e = v.tab === "tests" ? o.length > 0 || d : v.tab === "lint" ? g || d || !!s : v.tab !== "processes" || we.nodes.length > 0, t = v.tab === "tests" ? fe : v.tab === "lint" ? me : v.tab !== "processes" || Ee;
		e && !t && N({
			...v,
			selectedPath: v.tab === "processes" && Te ? Te.key : ""
		}, "replace");
	}, [
		v,
		fe,
		me,
		Ee,
		o.length,
		d,
		g,
		s,
		we.nodes.length,
		Te,
		N
	]);
	let Oe = (0, I.useCallback)((e) => {
		let { workDir: t, line: n, ...r } = v;
		N({
			...r,
			tab: e,
			selectedPath: ""
		});
	}, [v, N]), ke = (0, I.useCallback)((e) => N({
		...v,
		selectedPath: e?.route_path ?? ""
	}), [v, N]), Ae = (0, I.useCallback)((e) => N({
		...v,
		tab: "tests",
		filters: e,
		selectedPath: ""
	}), [v, N]), je = (0, I.useCallback)((e) => N({
		...v,
		tab: "lint",
		lintFilters: e,
		selectedPath: ""
	}), [v, N]), Me = (0, I.useCallback)((e) => N({
		...v,
		tab: "lint",
		lintGrouping: e,
		selectedPath: ""
	}), [v, N]), Ne = (0, I.useCallback)((e) => N({
		...v,
		selectedPath: e ?? ""
	}), [v, N]), Pe = (0, I.useCallback)((e) => N({
		...v,
		tab: "processes",
		selectedPath: e
	}), [v, N]), Fe = (0, I.useCallback)((e) => N({
		...v,
		tab: "processes",
		hideExited: e
	}, "replace"), [v, N]), Ie = (0, I.useCallback)((e, t) => {
		let { workDir: n, line: r, ...i } = v;
		N({
			...i,
			tab: "files",
			selectedPath: e.file,
			...e.workDir ? { workDir: e.workDir } : {},
			...t ? { line: t } : {}
		});
	}, [v, N]), Le = (0, I.useCallback)((e) => Ie({
		file: e.file,
		...e.workDir || _e ? { workDir: e.workDir || _e } : {}
	}, e.line), [Ie, _e]), Re = (0, I.useCallback)((e) => Ie(e), [Ie]), ze = (0, I.useCallback)((e) => {
		let { line: t, ...n } = v;
		N({
			...n,
			...e ? { line: e } : {}
		}, "replace");
	}, [v, N]), Be = (0, I.useCallback)((e) => N({
		...v,
		view: e
	}, "replace"), [v, N]), Ve = (0, I.useCallback)((e) => Em(e, n), [n]), Ue = (0, I.useCallback)((e) => N({
		...v,
		sampleIndex: e
	}), [v, N]), We = (0, I.useCallback)((e) => {
		N({
			...v,
			tab: "tests",
			filters: {
				status: /* @__PURE__ */ new Map(),
				framework: /* @__PURE__ */ new Map()
			},
			selectedPath: e.routePath
		});
	}, [v, N]), Ge = (0, I.useCallback)(async (e, t) => {
		if (!S) {
			C(!0), T(!0), M(`Rerunning ${t}...`);
			try {
				let t = await i.rerun(e);
				t.conflict ? M("Another rerun is in progress") : t.ok || M(`Rerun failed: ${t.message ?? t.status}`);
			} catch (e) {
				M(`Rerun error: ${e instanceof Error ? e.message : String(e)}`);
			} finally {
				C(!1);
			}
		}
	}, [
		S,
		i,
		M
	]), Ke = (0, I.useCallback)((e) => {
		if (e.framework === "task") return;
		let t = Om(e);
		"error" in t ? M(t.error) : Ge(t.body, e.name);
	}, [Ge, M]), qe = (0, I.useCallback)(() => {
		let e = dm(me, s ?? []);
		e.work_dir ? Ge(e, me?.name ?? "lint") : M("Rerun across multiple lint roots is not supported");
	}, [
		me,
		s,
		Ge,
		M
	]), Je = (0, I.useCallback)(async ({ taskId: e, label: t }) => {
		if (!te) {
			ne(e ?? Mm), M(`Stopping ${t}...`);
			try {
				let t = await i.stop(e);
				t.ok || M(`Stop failed: ${t.message ?? t.status}`);
			} catch (e) {
				M(`Stop error: ${e instanceof Error ? e.message : String(e)}`);
			} finally {
				ne(null);
			}
		}
	}, [
		te,
		i,
		M
	]), Ye = (0, I.useCallback)((e) => {
		e.task_id && Je({
			taskId: e.task_id,
			label: e.name
		});
	}, [Je]), Ze = (0, I.useCallback)(() => {
		Je({ label: "the run" });
	}, [Je]), Qe = (0, I.useCallback)(async (e) => {
		if (!E) {
			D(!0);
			try {
				let t = await fetch(zd("/api/lint/ignore", n), {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(e.req)
				});
				if (!t.ok) {
					M(`Ignore failed: ${(await t.text()).trim()}`);
					return;
				}
				M("Ignore rule saved to .gavel.yaml"), v.tab !== "files" && N({
					...v,
					selectedPath: ""
				}, "replace");
			} catch (e) {
				M(`Ignore error: ${e instanceof Error ? e.message : String(e)}`);
			} finally {
				D(!1);
			}
		}
	}, [
		E,
		n,
		v,
		N,
		M
	]), $e = (0, I.useCallback)((e) => {
		if (e.kind !== "lint") return null;
		let t = ye.get(wm(e));
		if (!t) throw Error(`no linter recorded for the finding at ${e.file}:${e.line}`);
		let n = Ys({
			violation: {
				...e.rule ? { rule: { method: e.rule } } : {},
				...e.line ? { line: e.line } : {}
			},
			linter: t,
			file: e.file,
			workDir: e.workDir ?? ""
		}), r = P.get(wm(e));
		return n.length === 0 && !r ? null : /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [r && /* @__PURE__ */ (0, L.jsx)(be, {
			asChild: !0,
			size: "sm",
			variant: "outline",
			className: "h-6 gap-1 px-2 text-xs text-muted-foreground",
			children: /* @__PURE__ */ (0, L.jsx)("a", {
				href: r,
				target: "_blank",
				rel: "noreferrer",
				title: `Open ${e.file}:${e.line ?? ""} at its source`,
				children: "Open source"
			})
		}), n.length > 0 && /* @__PURE__ */ (0, L.jsx)(He, {
			label: "Ignore",
			icon: he,
			variant: "outline",
			size: "sm",
			align: "right",
			title: "Add an ignore rule to .gavel.yaml",
			menuLabel: "Ignore scope",
			items: n.map((e) => ({
				label: e.label,
				title: e.title,
				disabled: E,
				onSelect: () => void Qe(e)
			}))
		})] });
	}, [
		ye,
		P,
		E,
		Qe
	]), nt = (0, I.useCallback)(async (e, t, n) => {
		if (O || c.test_edit_supported !== !0) return;
		let r = n === "file" ? e.file || "file" : e.name || "test";
		ee(!0), M(t === "skip" ? `Skipping ${r}...` : `Deleting ${r}...`);
		try {
			let r = await i.editTest({
				action: t,
				scope: n,
				framework: e.framework || "",
				work_dir: e.work_dir || "",
				package_path: e.package_path || "",
				file: e.file || "",
				line: e.line || 0,
				test_name: e.name || "",
				suite: e.suite || []
			});
			if (!r.ok) {
				M(`Test edit failed: ${r.message ?? r.status}`);
				return;
			}
			M(r.message || `${t === "skip" ? "Skip" : "Delete"} ${n === "file" ? "file" : "test"} saved`), t === "delete" && N({
				...v,
				selectedPath: ""
			}, "replace");
		} catch (e) {
			M(`Test edit error: ${e instanceof Error ? e.message : String(e)}`);
		} finally {
			ee(!1);
		}
	}, [
		O,
		c.test_edit_supported,
		i,
		v,
		N,
		M
	]), rt = (0, I.useMemo)(() => ({
		...v,
		selectedPath: ""
	}), [v]), at = (0, I.useMemo)(() => ({
		nodeRouteState: (e) => e.export_path ? {
			...v,
			tab: "tests",
			selectedPath: e.export_path
		} : void 0,
		failingOnlyRouteState: (e) => e.export_path ? {
			...v,
			tab: "tests",
			selectedPath: e.export_path,
			filters: {
				...v.filters,
				status: /* @__PURE__ */ new Map([["failed", "include"], ["timedout", "include"]])
			}
		} : void 0,
		testEditSupported: c.test_edit_supported === !0,
		onTestEdit: nt,
		testEditBusy: O
	}), [
		v,
		c.test_edit_supported,
		nt,
		O
	]), st = (0, I.useMemo)(() => Kf({
		traces: Se,
		processes: we.nodes,
		onOpenProcess: Pe,
		actions: at
	}), [
		Se,
		we.nodes,
		Pe,
		at
	]), ct = (0, I.useMemo)(() => Up({
		samples: we.samples,
		nodes: we.nodes
	}), [we.samples, we.nodes]), lt = g, dt = ve.length > 0 || v.tab === "files", ft = !!i.bench || !!m, pt = Ce, ht = i.bench?.deltas?.filter((e) => e.significant && e.delta_pct > i.bench.threshold).length || 0, gt = v.tab === "tests" && le.length > 0 || v.tab === "lint" && g || v.tab === "bench" && ft, _t = c.running && !!c.stop_supported, vt = /* @__PURE__ */ (0, L.jsx)(jp, {
		runMeta: u,
		canStop: _t,
		stopping: te === Mm,
		onStop: Ze,
		exportState: gt ? rt : void 0
	}), yt = re ?? p ?? f, bt = window.__gavelBackTo, xt = [
		{
			id: "tests",
			label: "Tests",
			icon: et,
			count: ue.failed > 0 ? ue.failed : ue.total,
			countColor: ue.failed > 0 ? "bg-red-500" : "bg-muted-foreground"
		},
		...lt ? [{
			id: "lint",
			label: "Lint",
			icon: it,
			count: ge,
			countColor: ge > 0 ? "bg-yellow-500" : "bg-muted-foreground"
		}] : [],
		...dt ? [{
			id: "files",
			label: "Files",
			icon: tt,
			count: ve.length,
			countColor: ve.length > 0 ? "bg-yellow-500" : "bg-muted-foreground"
		}] : [],
		...ft ? [{
			id: "bench",
			label: "Bench",
			icon: sn,
			count: m ? _m(o) : ht > 0 ? ht : i.bench?.deltas?.length || 0,
			countColor: ht > 0 ? "bg-red-500" : "bg-muted-foreground"
		}] : [],
		...pt ? [{
			id: "processes",
			label: "Processes",
			icon: on,
			count: F,
			countColor: "bg-blue-500"
		}] : [],
		..._.available ? [{
			id: "history",
			label: "History",
			icon: rn
		}] : []
	];
	return /* @__PURE__ */ (0, L.jsxs)("div", {
		className: "flex h-screen flex-col bg-background text-foreground",
		children: [
			bt && /* @__PURE__ */ (0, L.jsx)("div", {
				className: "flex shrink-0 items-center gap-2 border-b border-border bg-card px-4 py-1.5 text-sm",
				children: /* @__PURE__ */ (0, L.jsxs)("a", {
					href: bt,
					className: "inline-flex items-center gap-1 text-primary no-underline hover:underline",
					children: [/* @__PURE__ */ (0, L.jsx)(Xe, {}), "Back to PR"]
				})
			}),
			(xt.length > 1 || _.available) && /* @__PURE__ */ (0, L.jsxs)("div", {
				className: "flex shrink-0 items-center gap-2 pr-density-4",
				children: [/* @__PURE__ */ (0, L.jsx)(Ut, {
					tabs: xt,
					value: v.tab,
					onChange: Oe,
					className: "min-w-0 flex-1 px-density-4"
				}), _.available && /* @__PURE__ */ (0, L.jsx)(Vp, {
					history: _,
					navigate: t
				})]
			}),
			/* @__PURE__ */ (0, L.jsxs)("div", {
				className: "min-h-0 flex-1",
				children: [
					v.tab === "tests" && (le.length === 0 ? /* @__PURE__ */ (0, L.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, L.jsx)(km, {
							icon: et,
							title: "Test Results",
							status: yt,
							actions: vt
						}), /* @__PURE__ */ (0, L.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, L.jsx)(Am, { done: d })
						})]
					}) : /* @__PURE__ */ (0, L.jsx)(j, {
						tests: de,
						allTests: le,
						selected: fe,
						filters: v.filters,
						expandAll: b,
						done: d,
						...u ? { runMeta: u } : {},
						status: c,
						statusText: de.length === 0 ? `${yt} · No tests match the current filters` : yt,
						now: ae,
						startTime: se,
						endTime: ce,
						busy: {
							rerun: S,
							stop: te !== null
						},
						onSelect: ke,
						onFiltersChange: Ae,
						onExpandAllChange: x,
						onRerun: Ke,
						onStop: Ye,
						fetchProcessMetrics: ct,
						adapters: st,
						onOpenFile: Le,
						headerSlot: vt
					})),
					v.tab === "lint" && /* @__PURE__ */ (0, L.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, L.jsx)(Sp, { tests: o }), /* @__PURE__ */ (0, L.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, L.jsx)(qc, {
								results: s ?? [],
								grouping: v.lintGrouping,
								onGroupingChange: Me,
								filters: v.lintFilters,
								onFiltersChange: je,
								selected: me ? v.selectedPath : null,
								onSelect: Ne,
								expandAll: b,
								onExpandAllChange: x,
								onIgnore: Qe,
								ignoreBusy: E,
								onOpenFile: Le,
								renderRun: xm,
								headerSlot: /* @__PURE__ */ (0, L.jsxs)(L.Fragment, { children: [
									/* @__PURE__ */ (0, L.jsx)("span", {
										className: "truncate text-sm text-muted-foreground",
										children: yt
									}),
									/* @__PURE__ */ (0, L.jsxs)(be, {
										type: "button",
										size: "sm",
										disabled: S,
										onClick: qe,
										title: me ? "Rerun the linters for this selection" : "Rerun the linters",
										children: [/* @__PURE__ */ (0, L.jsx)(l, { className: "mr-1" }), S ? "Running..." : "Rerun"]
									}),
									vt
								] })
							})
						})]
					}),
					v.tab === "files" && /* @__PURE__ */ (0, L.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, L.jsx)(km, {
							icon: tt,
							title: "Files",
							status: yt,
							actions: vt
						}), /* @__PURE__ */ (0, L.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, L.jsx)(I.Suspense, {
								fallback: /* @__PURE__ */ (0, L.jsx)(Sm, {}),
								children: /* @__PURE__ */ (0, L.jsx)(ym, {
									findings: ve,
									selectedFile: xe,
									onSelectFile: Re,
									line: v.line,
									onLineChange: ze,
									view: v.view ?? "preview",
									onViewChange: Be,
									loadFile: Ve,
									renderFindingActions: $e
								})
							})
						})]
					}),
					v.tab === "bench" && /* @__PURE__ */ (0, L.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, L.jsx)(km, {
							icon: sn,
							title: m ? "Fixture Benchmark" : "Benchmark Comparison",
							status: yt,
							actions: vt
						}), /* @__PURE__ */ (0, L.jsx)("div", {
							className: "min-h-0 flex-1 overflow-auto",
							children: m ? /* @__PURE__ */ (0, L.jsx)(cp, {
								performance: m,
								tests: o,
								error: h,
								sampleIndex: v.sampleIndex,
								onSampleChange: Ue
							}) : /* @__PURE__ */ (0, L.jsx)(ep, { bench: i.bench })
						})]
					}),
					v.tab === "history" && /* @__PURE__ */ (0, L.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, L.jsx)(km, {
							icon: rn,
							title: "Run History",
							status: _.current ? `Viewing ${_.current}` : "Viewing the latest results"
						}), /* @__PURE__ */ (0, L.jsx)("div", {
							className: "flex min-h-0 flex-1 flex-col p-density-2",
							children: /* @__PURE__ */ (0, L.jsx)(Rp, {
								history: _,
								navigate: t
							})
						})]
					}),
					v.tab === "processes" && /* @__PURE__ */ (0, L.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, L.jsx)(km, {
							icon: on,
							title: "Processes",
							status: yt,
							actions: vt
						}), /* @__PURE__ */ (0, L.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, L.jsx)(zp, {
								trace: we,
								selectedKey: Ee?.key ?? null,
								onSelect: Pe,
								hideExited: !!v.hideExited,
								onHideExitedChange: Fe,
								expandAll: b,
								onExpandAllChange: x,
								stepLinks: De,
								onOpenStep: We
							})
						})]
					})
				]
			}),
			/* @__PURE__ */ (0, L.jsx)(Dp, {
				open: w,
				onClose: () => T(!1),
				streamUrl: zd("/api/rerun/stream", n)
			})
		]
	});
}
// src/index.tsx
var Fm = document.createElement("style");
Fm.textContent = Yt, document.head.appendChild(Fm), (0, Jt.createRoot)(document.getElementById("root")).render(/* @__PURE__ */ (0, L.jsx)(Pm, {}));
