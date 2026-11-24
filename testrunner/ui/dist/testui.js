import { a as e, t } from "./chunks/rolldown-runtime-DUcC5eAW.js";
import { n, t as r } from "./chunks/jsx-runtime-B4NxTvNR.js";
import { t as i } from "./chunks/react-dom-BBpWcFtU.js";
import { B as a, C as o, D as s, F as c, H as l, M as u, N as d, P as f, T as p, _ as m, a as h, b as g, c as _, d as v, f as y, g as b, h as x, i as ee, j as te, l as ne, m as re, o as S, p as ie, r as ae, s as oe, t as se, u as ce, v as le, w as ue, y as de } from "./chunks/adapters-U8xOlcS2.js";
import { a as C, t as fe } from "./chunks/button-BLoR2pgR.js";
import { i as w, r as pe } from "./chunks/UiInfo-CCiCR0Fr.js";
import { G as me, K as T, Q as he, U as ge, V as _e, W as E, X as ve, at as D, f as ye, it as be, j as xe, rt as O, s as k, t as Se } from "./chunks/SplitButton-w2X8sfgS.js";
import { a as Ce, d as we, f as Te, l as Ee, m as A, o as De, p as Oe, r as ke } from "./chunks/ErrorDetails-CbwPqfQb.js";
import { H as Ae, I as je, R as Me, U as Ne, d as Pe, l as Fe, n as Ie, t as Le, z as Re } from "./chunks/JsonSchemaForm-C40jV51n.js";
import { C as ze, E as Be, I as Ve, J as He, L as Ue, R as We, S as Ge, f as Ke, h as qe, k as Je, n as Ye, x as Xe, z as Ze } from "./chunks/Markdown-s_A3Eic8.js";
import { $ as Qe, A as $e, C as et, E as tt, I as nt, K as rt, L as it, P as at, R as ot, S as st, T as ct, U as lt, W as ut, _ as dt, a as ft, at as pt, b as mt, ct as ht, dt as gt, et as _t, f as vt, g as yt, h as bt, m as xt, s as St, v as Ct } from "./chunks/locations-CUxJY9f0.js";
import { t as wt, y as Tt } from "./chunks/Callout-BqeseRSA.js";
import { _ as Et, b as Dt, c as Ot, d as kt, f as At, g as jt, h as Mt, l as Nt, m as Pt, n as Ft, o as It, p as Lt, r as Rt, s as zt, u as Bt, v as Vt, x as Ht, y as Ut } from "./chunks/adapters-BoPmjJjW.js";
import { n as Wt, t as Gt } from "./chunks/ProblemActions-MzUA9awO.js";
import { t as Kt } from "./chunks/Tabs-rMrrxeic.js";
import { t as qt } from "./chunks/lookupOptionPresentation-Dm9WE30P.js";
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/scheduler/0.28.0/26f61da90a7f44735cfb745f2b035a58334b2fcd58be6a514787c454b12dc803/node_modules/scheduler/cjs/scheduler.production.js
var Jt = /* @__PURE__ */ t(((e) => {
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
			if (n(c) !== null) m = !0, ee || (ee = !0, ae());
			else {
				var t = n(l);
				t !== null && ce(x, t.startTime - e);
			}
		}
	}
	var ee = !1, te = -1, ne = 5, re = -1;
	function S() {
		return g ? !0 : !(e.unstable_now() - re < ne);
	}
	function ie() {
		if (g = !1, ee) {
			var t = e.unstable_now();
			re = t;
			var i = !0;
			try {
				a: {
					m = !1, h && (h = !1, v(te), te = -1), p = !0;
					var a = f;
					try {
						b: {
							for (b(t), d = n(c); d !== null && !(d.expirationTime > t && S());) {
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
								u !== null && ce(x, u.startTime - t), i = !1;
							}
						}
						break a;
					} finally {
						d = null, f = a, p = !1;
					}
					i = void 0;
				}
			} finally {
				i ? ae() : ee = !1;
			}
		}
	}
	var ae;
	if (typeof y == "function") ae = function() {
		y(ie);
	};
	else if (typeof MessageChannel < "u") {
		var oe = new MessageChannel(), se = oe.port2;
		oe.port1.onmessage = ie, ae = function() {
			se.postMessage(null);
		};
	} else ae = function() {
		_(ie, 0);
	};
	function ce(t, n) {
		te = _(function() {
			t(e.unstable_now());
		}, n);
	}
	e.unstable_IdlePriority = 5, e.unstable_ImmediatePriority = 1, e.unstable_LowPriority = 4, e.unstable_NormalPriority = 3, e.unstable_Profiling = null, e.unstable_UserBlockingPriority = 2, e.unstable_cancelCallback = function(e) {
		e.callback = null;
	}, e.unstable_forceFrameRate = function(e) {
		0 > e || 125 < e ? console.error("forceFrameRate takes a positive int between 0 and 125, forcing frame rates higher than 125 fps is not supported") : ne = 0 < e ? Math.floor(1e3 / e) : 5;
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
		}, a > o ? (r.sortIndex = a, t(l, r), n(c) === null && r === n(l) && (h ? (v(te), te = -1) : h = !0, ce(x, a - o))) : (r.sortIndex = s, t(c, r), m || p || (m = !0, ee || (ee = !0, ae()))), r;
	}, e.unstable_shouldYield = S, e.unstable_wrapCallback = function(e) {
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
})), Yt = /* @__PURE__ */ t(((e, t) => {
	t.exports = Jt();
})), Xt = /* @__PURE__ */ t(((e) => {
	var t = Yt(), r = n(), a = i();
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
		for (var t = e, n = t; n && !n.alternate;) t = n, t.flags & 4098 && (e = t.return), n = t.return;
		for (; t.return;) t = t.return;
		return t.tag === 3 ? e : null;
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
	function m(e, t, n, r, i, a) {
		for (; e !== null;) {
			if ((e.tag === 5 || e.tag === 27 || e.tag === 6) && n(e, r, i, a) || (e.tag !== 22 || e.memoizedState === null) && (t || e.tag !== 5 && e.tag !== 27) && m(e.child, t, n, r, i, a)) return !0;
			e = e.sibling;
		}
		return !1;
	}
	function h(e) {
		for (e = e.return; e !== null;) {
			if (e.tag === 3 || e.tag === 5 || e.tag === 27) return e;
			e = e.return;
		}
		return null;
	}
	function g(e) {
		var t = !1;
		for (e = e.return; e !== null && (e.tag === 4 && (t = !0), e.tag !== 3 && e.tag !== 5 && e.tag !== 27);) e = e.return;
		return t;
	}
	function _(e) {
		var t = [null, null], n = h(e);
		return n === null || v(t, e, n.child, { foundSelf: !1 }), t;
	}
	function v(e, t, n, r) {
		for (; n !== null;) {
			if (n === t) r.foundSelf = !0;
			else if (n.tag === 5 || n.tag === 27 || n.tag === 6) {
				if (r.foundSelf) return e[1] = n, !0;
				e[0] = n;
			} else if ((n.tag !== 22 || n.memoizedState === null) && v(e, t, n.child, r)) return !0;
			n = n.sibling;
		}
		return !1;
	}
	function y(e) {
		switch (e.tag) {
			case 5:
			case 27:
			case 6: return e.stateNode;
			case 3: return e.stateNode.containerInfo;
			default: throw Error(o(559));
		}
	}
	var b = null, x = null;
	function ee(e, t, n) {
		return e === n || e === t && (b = e, !0);
	}
	function te(e, t, n) {
		return e === n ? (x = e, !1) : e === t && (x !== null && (b = e), !0);
	}
	function ne(e) {
		if (e === null) return null;
		do
			e = e === null ? null : e.return;
		while (e && e.tag !== 5 && e.tag !== 27 && e.tag !== 3);
		return e || null;
	}
	function re(e, t, n) {
		for (var r = 0, i = e; i; i = n(i)) r++;
		i = 0;
		for (var a = t; a; a = n(a)) i++;
		for (; 0 < r - i;) e = n(e), r--;
		for (; 0 < i - r;) t = n(t), i--;
		for (; r--;) {
			if (e === t || t !== null && e === t.alternate) return e;
			e = n(e), t = n(t);
		}
		return null;
	}
	var S = Object.assign, ie = Symbol.for("react.element"), ae = Symbol.for("react.transitional.element"), oe = Symbol.for("react.portal"), se = Symbol.for("react.fragment"), ce = Symbol.for("react.strict_mode"), le = Symbol.for("react.profiler"), ue = Symbol.for("react.consumer"), de = Symbol.for("react.context"), C = Symbol.for("react.forward_ref"), fe = Symbol.for("react.suspense"), w = Symbol.for("react.suspense_list"), pe = Symbol.for("react.memo"), me = Symbol.for("react.lazy"), T = Symbol.for("react.activity"), he = Symbol.for("react.legacy_hidden"), ge = Symbol.for("react.memo_cache_sentinel"), _e = Symbol.for("react.view_transition"), E = Symbol.for("react.recoverable"), ve = Symbol.iterator;
	function D(e) {
		return typeof e != "object" || !e ? null : (e = ve && e[ve] || e["@@iterator"], typeof e == "function" ? e : null);
	}
	var ye = Symbol.for("react.client.reference");
	function be(e) {
		if (e == null) return null;
		if (typeof e == "function") return e.$$typeof === ye ? null : e.displayName || e.name || null;
		if (typeof e == "string") return e;
		switch (e) {
			case se: return "Fragment";
			case le: return "Profiler";
			case ce: return "StrictMode";
			case fe: return "Suspense";
			case w: return "SuspenseList";
			case T: return "Activity";
			case _e: return "ViewTransition";
		}
		if (typeof e == "object") switch (e.$$typeof) {
			case oe: return "Portal";
			case de: return e.displayName || "Context";
			case ue: return (e._context.displayName || "Context") + ".Consumer";
			case C:
				var t = e.render;
				return e = e.displayName, e ||= (e = t.displayName || t.name || "", e === "" ? "ForwardRef" : "ForwardRef(" + e + ")"), e;
			case pe: return t = e.displayName || null, t === null ? be(e.type) || "Memo" : t;
			case me:
				t = e._payload, e = e._init;
				try {
					return be(e(t));
				} catch {}
		}
		return null;
	}
	var xe = Array.isArray, O = r.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE, k = a.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE, Se = {
		pending: !1,
		data: null,
		method: null,
		action: null
	}, Ce = [], we = -1;
	function Te(e) {
		return { current: e };
	}
	function Ee(e) {
		0 > we || (e.current = Ce[we], Ce[we] = null, we--);
	}
	function A(e, t) {
		we++, Ce[we] = e.current, e.current = t;
	}
	var De = Te(null), Oe = Te(null), ke = Te(null), Ae = Te(null);
	function je(e, t) {
		switch (A(ke, t), A(Oe, e), A(De, null), t.nodeType) {
			case 9:
			case 11:
				e = (e = t.documentElement) && (e = e.namespaceURI) ? up(e) : 0;
				break;
			default: if (e = t.tagName, t = t.namespaceURI) t = up(t), e = dp(t, e);
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
		Ee(De), A(De, e);
	}
	function Me() {
		Ee(De), Ee(Oe), Ee(ke);
	}
	function Ne(e) {
		var t = e.memoizedState;
		t !== null && (sh._currentValue = t.memoizedState, A(Ae, e)), t = De.current;
		var n = dp(t, e.type);
		t !== n && (A(Oe, e), A(De, n));
	}
	function Pe(e) {
		Oe.current === e && (Ee(De), Ee(Oe)), Ae.current === e && (Ee(Ae), sh._currentValue = Se);
	}
	var Fe, Ie;
	function Le(e) {
		if (Fe === void 0) try {
			throw Error();
		} catch (e) {
			var t = e.stack.trim().match(/\n( *(at )?)/);
			Fe = t && t[1] || "", Ie = -1 < e.stack.indexOf("\n    at") ? " (<anonymous>)" : -1 < e.stack.indexOf("@") ? "@unknown:0:0" : "";
		}
		return "\n" + Fe + e + Ie;
	}
	var Re = !1;
	function ze(e, t) {
		if (!e || Re) return "";
		Re = !0;
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
							n = !1;
							try {
								var i = Object.getOwnPropertyDescriptor(e.prototype, "props");
								Object.defineProperty(e.prototype, "props", {
									configurable: !0,
									set: function() {
										throw Error();
									}
								}), n = !0, new e();
							} finally {
								n && (i === void 0 ? delete e.prototype.props : Object.defineProperty(e.prototype, "props", i));
							}
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
			Re = !1, Error.prepareStackTrace = n;
		}
		return (n = e ? e.displayName || e.name : "") ? Le(n) : "";
	}
	function Be(e, t) {
		switch (e.tag) {
			case 26:
			case 27:
			case 5: return Le(e.type);
			case 16: return Le("Lazy");
			case 13: return e.child !== t && t !== null ? Le("Suspense Fallback") : Le("Suspense");
			case 19: return Le("SuspenseList");
			case 0:
			case 15: return ze(e.type, !1);
			case 11: return ze(e.type.render, !1);
			case 1: return ze(e.type, !0);
			case 31: return Le("Activity");
			case 30: return Le("ViewTransition");
			default: return "";
		}
	}
	function Ve(e) {
		try {
			var t = "", n = null;
			do
				t += Be(e, n), n = e, e = e.return;
			while (e);
			return t;
		} catch (e) {
			return "\nError generating stack: " + e.message + "\n" + e.stack;
		}
	}
	var He = Object.prototype.hasOwnProperty, Ue = t.unstable_scheduleCallback, We = t.unstable_cancelCallback, Ge = t.unstable_shouldYield, Ke = t.unstable_requestPaint, qe = t.unstable_now, Je = t.unstable_getCurrentPriorityLevel, Ye = t.unstable_ImmediatePriority, Xe = t.unstable_UserBlockingPriority, Ze = t.unstable_NormalPriority, Qe = t.unstable_LowPriority, $e = t.unstable_IdlePriority, et = t.log, tt = t.unstable_setDisableYieldValue, nt = null, rt = null;
	function it(e) {
		if (typeof et == "function" && tt(e), rt && typeof rt.setStrictMode == "function") try {
			rt.setStrictMode(nt, e);
		} catch {}
	}
	var at = Math.clz32 ? Math.clz32 : ct, ot = Math.log, st = Math.LN2;
	function ct(e) {
		return e >>>= 0, e === 0 ? 32 : 31 - (ot(e) / st | 0) | 0;
	}
	var lt = 256, ut = 262144, dt = 4194304;
	function ft(e) {
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
			case 131072: return e & -e;
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
	function pt(e, t, n) {
		var r = e.pendingLanes;
		if (r === 0) return 0;
		var i = 0, a = e.suspendedLanes, o = e.pingedLanes;
		e = e.warmLanes;
		var s = r & 134217727;
		return s === 0 ? (s = r & ~a, s === 0 ? o === 0 ? n || (n = r & ~e, n !== 0 && (i = ft(n))) : i = ft(o) : i = ft(s)) : (r = s & ~a, r === 0 ? (o &= s, o === 0 ? n || (n = s & ~e, n !== 0 && (i = ft(n))) : i = ft(o)) : i = ft(r)), i === 0 ? 0 : t !== 0 && t !== i && (t & a) === 0 && (a = i & -i, n = t & -t, a >= n || a === 32 && n & 4194048) ? t : i;
	}
	function mt(e, t) {
		return (e.pendingLanes & ~(e.suspendedLanes & ~e.pingedLanes) & t) === 0;
	}
	function ht(e, t) {
		t & 8 && (t |= t & 32);
		var n = e.entangledLanes;
		if (n !== 0) for (e = e.entanglements, n &= t; 0 < n;) {
			var r = 31 - at(n), i = 1 << r;
			t |= e[r], n &= ~i;
		}
		return t;
	}
	function gt(e, t) {
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
	function _t() {
		var e = dt;
		return dt <<= 1, !(dt & 62914560) && (dt = 4194304), e;
	}
	function vt(e) {
		for (var t = [], n = 0; 31 > n; n++) t.push(e);
		return t;
	}
	function yt(e, t) {
		e.pendingLanes |= t, t !== 268435456 && (e.suspendedLanes = 0, e.pingedLanes = 0, e.warmLanes = 0);
	}
	function bt(e, t, n, r, i, a) {
		var o = e.pendingLanes;
		e.pendingLanes = n, e.suspendedLanes = 0, e.pingedLanes = 0, e.warmLanes = 0, e.expiredLanes &= n, e.entangledLanes &= n, e.errorRecoveryDisabledLanes &= n, e.shellSuspendCounter = 0;
		var s = e.entanglements, c = e.expirationTimes, l = e.hiddenUpdates;
		for (n = o & ~n; 0 < n;) {
			var u = 31 - at(n), d = 1 << u;
			s[u] = 0, c[u] = -1;
			var f = l[u];
			if (f !== null) for (l[u] = null, u = 0; u < f.length; u++) {
				var p = f[u];
				p !== null && (p.lane &= -536870913);
			}
			n &= ~d;
		}
		r !== 0 && xt(e, r, 0), a !== 0 && i === 0 && e.tag !== 0 && (e.suspendedLanes |= a & ~(o & ~t));
	}
	function xt(e, t, n) {
		e.pendingLanes |= t, e.suspendedLanes &= ~t;
		var r = 31 - at(t);
		e.entangledLanes |= t, e.entanglements[r] = e.entanglements[r] | 1073741824 | n & 261930;
	}
	function St(e, t) {
		var n = e.entangledLanes |= t;
		for (e = e.entanglements; n;) {
			var r = 31 - at(n), i = 1 << r;
			i & t | e[r] & t && (e[r] |= t), n &= ~i;
		}
	}
	function Ct(e, t) {
		var n = t & -t;
		return n = n & 42 ? 1 : wt(n), (n & (e.suspendedLanes | t)) === 0 ? n : 0;
	}
	function wt(e) {
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
	function Tt(e) {
		return e &= -e, 2 < e ? 8 < e ? e & 134217727 ? 32 : 268435456 : 8 : 2;
	}
	function Et() {
		var e = k.p;
		return e === 0 ? (e = window.event, e === void 0 ? 32 : Ch(e.type)) : e;
	}
	function Dt(e, t) {
		var n = k.p;
		try {
			return k.p = e, t();
		} finally {
			k.p = n;
		}
	}
	var Ot = Math.random().toString(36).slice(2), kt = "__reactFiber$" + Ot, At = "__reactProps$" + Ot, jt = "__reactContainer$" + Ot, Mt = "__reactEvents$" + Ot, Nt = "__reactListeners$" + Ot, Pt = "__reactHandles$" + Ot, Ft = "__reactResources$" + Ot, It = "__reactMarker$" + Ot, Lt = "__reactLoad$" + Ot;
	function Rt(e) {
		delete e[kt], delete e[At], delete e[Nt], delete e[Pt];
	}
	function zt(e) {
		var t;
		if (t = e[kt]) return t;
		for (var n = e.parentNode; n;) {
			if (t = n[jt] || n[kt]) {
				if (n = t.alternate, t.child !== null || n !== null && n.child !== null) for (e = fm(e); e !== null;) {
					if (n = e[kt]) return n;
					e = fm(e);
				}
				return t;
			}
			e = n, n = e.parentNode;
		}
		return null;
	}
	function Bt(e) {
		if (e = e[kt] || e[jt]) {
			var t = e.tag;
			if (t === 5 || t === 6 || t === 13 || t === 31 || t === 26 || t === 27 || t === 3) return e;
		}
		return null;
	}
	function Vt(e) {
		var t = e.tag;
		if (t === 5 || t === 26 || t === 27 || t === 6) return e.stateNode;
		throw Error(o(33));
	}
	function Ht(e) {
		var t = e[Ft];
		return t ||= e[Ft] = {
			hoistableStyles: /* @__PURE__ */ new Map(),
			hoistableScripts: /* @__PURE__ */ new Map()
		}, t;
	}
	function Ut(e) {
		e[It] = !0;
	}
	function Wt(e) {
		e[Lt] = void 0;
	}
	var Gt = /* @__PURE__ */ new Set(), Kt = {};
	function qt(e, t) {
		Jt(e, t), Jt(e + "Capture", t);
	}
	function Jt(e, t) {
		for (Kt[e] = t, e = 0; e < t.length; e++) Gt.add(t[e]);
	}
	var Xt = RegExp("^[:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD][:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD\\-.0-9\\u00B7\\u0300-\\u036F\\u203F-\\u2040]*$"), Zt = {}, Qt = {};
	function j(e) {
		return He.call(Qt, e) ? !0 : He.call(Zt, e) ? !1 : Xt.test(e) ? Qt[e] = !0 : (Zt[e] = !0, !1);
	}
	var M = !1;
	function $t() {
		var e = M;
		return M = !1, e;
	}
	function en(e, t, n) {
		if (j(t)) {
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
				e.setAttribute(t, n);
			}
		}
	}
	function tn(e, t, n) {
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
			e.setAttribute(t, n);
		}
	}
	function nn(e, t, n, r) {
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
			e.setAttributeNS(t, n, r);
		}
	}
	function rn(e) {
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
	function an(e) {
		var t = e.type;
		return (e = e.nodeName) && e.toLowerCase() === "input" && (t === "checkbox" || t === "radio");
	}
	function on(e, t, n) {
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
	function sn(e) {
		if (!e._valueTracker) {
			var t = an(e) ? "checked" : "value";
			e._valueTracker = on(e, t, "" + e[t]);
		}
	}
	function cn(e) {
		if (!e) return !1;
		var t = e._valueTracker;
		if (!t) return !0;
		var n = t.getValue(), r = "";
		return e && (r = an(e) ? e.checked ? "true" : "false" : e.value), e = r, e !== n && (t.setValue(e), !0);
	}
	var ln = /[\n"\\]/g;
	function un(e) {
		return e.replace(ln, function(e) {
			return "\\" + e.charCodeAt(0).toString(16) + " ";
		});
	}
	function dn(e, t, n, r, i, a, o, s) {
		e.name = "", o != null && typeof o != "function" && typeof o != "symbol" && typeof o != "boolean" ? e.type = o : e.removeAttribute("type"), t == null ? o !== "submit" && o !== "reset" || e.removeAttribute("value") : o === "number" ? (t === 0 && e.value === "" || e.value != t) && (e.value = "" + rn(t)) : e.value !== "" + rn(t) && (e.value = "" + rn(t)), t == null ? n == null ? r != null && e.removeAttribute("value") : pn(e, rn(n)) : o === "number" && e.value == t ? pn(e, rn(e.value)) : pn(e, rn(t)), i == null && a != null && (e.defaultChecked = !!a), i != null && (e.checked = i && typeof i != "function" && typeof i != "symbol"), s != null && typeof s != "function" && typeof s != "symbol" && typeof s != "boolean" ? e.name = "" + rn(s) : e.removeAttribute("name");
	}
	function fn(e, t, n, r, i, a, o, s) {
		if (a != null && typeof a != "function" && typeof a != "symbol" && typeof a != "boolean" && (e.type = a), t != null || n != null) {
			if (!(a !== "submit" && a !== "reset" || t != null)) {
				sn(e);
				return;
			}
			n = n == null ? "" : "" + rn(n), t = t == null ? n : "" + rn(t), s || t === e.value || (e.value = t), e.defaultValue = t;
		}
		r ??= i, r = typeof r != "function" && typeof r != "symbol" && !!r, e.checked = s ? e.checked : !!r, e.defaultChecked = !!r, o != null && typeof o != "function" && typeof o != "symbol" && typeof o != "boolean" && (e.name = o), sn(e);
	}
	function pn(e, t) {
		e.defaultValue !== "" + t && (e.defaultValue = "" + t);
	}
	function mn(e, t, n, r) {
		if (e = e.options, t) {
			t = {};
			for (var i = 0; i < n.length; i++) t["$" + n[i]] = !0;
			for (n = 0; n < e.length; n++) i = t.hasOwnProperty("$" + e[n].value), e[n].selected !== i && (e[n].selected = i), i && r && (e[n].defaultSelected = !0);
		} else {
			for (n = "" + rn(n), t = null, i = 0; i < e.length; i++) {
				if (e[i].value === n) {
					e[i].selected = !0, r && (e[i].defaultSelected = !0);
					return;
				}
				t !== null || e[i].disabled || (t = e[i]);
			}
			t !== null && (t.selected = !0);
		}
	}
	function hn(e, t, n) {
		t != null && (t = "" + rn(t), t !== e.value && (e.value = t), n == null) ? e.defaultValue !== t && (e.defaultValue = t) : e.defaultValue = n == null ? "" : "" + rn(n);
	}
	function gn(e, t, n, r) {
		if (t == null) {
			if (r != null) {
				if (n != null) throw Error(o(92));
				if (xe(r)) {
					if (1 < r.length) throw Error(o(93));
					r = r[0];
				}
				n = r;
			}
			n ??= "", t = n;
		}
		n = rn(t), e.defaultValue = n, r = e.textContent, r === n && r !== "" && r !== null && (e.value = r), sn(e);
	}
	function _n(e, t) {
		if (t) {
			var n = e.firstChild;
			if (n && n === e.lastChild && n.nodeType === 3) {
				n.nodeValue = t;
				return;
			}
		}
		e.textContent = t;
	}
	var vn = new Set("animationIterationCount aspectRatio borderImageOutset borderImageSlice borderImageWidth boxFlex boxFlexGroup boxOrdinalGroup columnCount columns flex flexGrow flexPositive flexShrink flexNegative flexOrder gridArea gridRow gridRowEnd gridRowSpan gridRowStart gridColumn gridColumnEnd gridColumnSpan gridColumnStart fontWeight lineClamp lineHeight opacity order orphans scale tabSize widows zIndex zoom fillOpacity floodOpacity stopOpacity strokeDasharray strokeDashoffset strokeMiterlimit strokeOpacity strokeWidth MozAnimationIterationCount MozBoxFlex MozBoxFlexGroup MozLineClamp msAnimationIterationCount msFlex msZoom msFlexGrow msFlexNegative msFlexOrder msFlexPositive msFlexShrink msGridColumn msGridColumnSpan msGridRow msGridRowSpan WebkitAnimationIterationCount WebkitBoxFlex WebKitBoxFlexGroup WebkitBoxOrdinalGroup WebkitColumnCount WebkitColumns WebkitFlex WebkitFlexGrow WebkitFlexPositive WebkitFlexShrink WebkitLineClamp".split(" "));
	function yn(e, t, n) {
		var r = t.indexOf("--") === 0;
		n == null || typeof n == "boolean" || n === "" ? r ? e.setProperty(t, "") : t === "float" ? e.cssFloat = "" : e[t] = "" : r ? e.setProperty(t, n) : typeof n != "number" || n === 0 || vn.has(t) ? t === "float" ? e.cssFloat = n : e[t] = ("" + n).trim() : e[t] = n + "px";
	}
	function bn(e, t, n) {
		if (t != null && typeof t != "object") throw Error(o(62));
		if (e = e.style, n != null) {
			for (var r in n) !n.hasOwnProperty(r) || t != null && t.hasOwnProperty(r) || (r.indexOf("--") === 0 ? e.setProperty(r, "") : r === "float" ? e.cssFloat = "" : e[r] = "", M = !0);
			for (var i in t) r = t[i], t.hasOwnProperty(i) && n[i] !== r && (yn(e, i, r), M = !0);
		} else for (var a in t) t.hasOwnProperty(a) && yn(e, a, t[a]);
	}
	function xn(e) {
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
	var Sn = /* @__PURE__ */ new Map([
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
		["maskType", "mask-type"],
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
	]), Cn = /^[\u0000-\u001F ]*j[\r\n\t]*a[\r\n\t]*v[\r\n\t]*a[\r\n\t]*s[\r\n\t]*c[\r\n\t]*r[\r\n\t]*i[\r\n\t]*p[\r\n\t]*t[\r\n\t]*:/i;
	function wn(e) {
		return Cn.test("" + e) ? "javascript:throw new Error('React has blocked a javascript: URL as a security precaution.')" : e;
	}
	function Tn() {}
	var En = null;
	function Dn(e) {
		return e = e.target || e.srcElement || window, e.correspondingUseElement && (e = e.correspondingUseElement), e.nodeType === 3 ? e.parentNode : e;
	}
	var On = null, kn = null;
	function An(e) {
		var t = Bt(e);
		if (t && (e = t.stateNode)) {
			var n = e[At] || null;
			a: switch (e = t.stateNode, t.type) {
				case "input":
					if (dn(e, n.value, n.defaultValue, n.defaultValue, n.checked, n.defaultChecked, n.type, n.name), t = n.name, n.type === "radio" && t != null) {
						for (n = e; n.parentNode;) n = n.parentNode;
						for (n = n.querySelectorAll("input[name=\"" + un("" + t) + "\"][type=\"radio\"]"), t = 0; t < n.length; t++) {
							var r = n[t];
							if (r !== e && r.form === e.form) {
								var i = r[At] || null;
								if (!i) throw Error(o(90));
								dn(r, i.value, i.defaultValue, i.defaultValue, i.checked, i.defaultChecked, i.type, i.name);
							}
						}
						for (t = 0; t < n.length; t++) r = n[t], r.form === e.form && cn(r);
					}
					break a;
				case "textarea":
					hn(e, n.value, n.defaultValue);
					break a;
				case "select": t = n.value, t != null && mn(e, !!n.multiple, t, !1);
			}
		}
	}
	var jn = !1;
	function Mn(e, t, n) {
		if (jn) return e(t, n);
		jn = !0;
		try {
			return e(t);
		} finally {
			if (jn = !1, (On !== null || kn !== null) && (zd(), On && (t = On, e = kn, kn = On = null, An(t), e))) for (t = 0; t < e.length; t++) An(e[t]);
		}
	}
	function Nn(e, t) {
		var n = e.stateNode;
		if (n === null) return null;
		var r = n[At] || null;
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
	var Pn = typeof window < "u" && window.document !== void 0 && window.document.createElement !== void 0, Fn = !1;
	if (Pn) try {
		var In = {};
		Object.defineProperty(In, "passive", { get: function() {
			Fn = !0;
		} }), window.addEventListener("test", In, In), window.removeEventListener("test", In, In);
	} catch {
		Fn = !1;
	}
	var Ln = null, Rn = null, zn = null;
	function Bn() {
		if (zn) return zn;
		var e, t = Rn, n = t.length, r, i = "value" in Ln ? Ln.value : Ln.textContent, a = i.length;
		for (e = 0; e < n && t[e] === i[e]; e++);
		var o = n - e;
		for (r = 1; r <= o && t[n - r] === i[a - r]; r++);
		return zn = i.slice(e, 1 < r ? 1 - r : void 0);
	}
	function Vn(e) {
		var t = e.keyCode;
		return "charCode" in e ? (e = e.charCode, e === 0 && t === 13 && (e = 13)) : e = t, e === 10 && (e = 13), 32 <= e || e === 13 ? e : 0;
	}
	function Hn() {
		return !0;
	}
	function Un() {
		return !1;
	}
	function Wn(e) {
		function t(t, n, r, i, a) {
			for (var o in this._reactName = t, this._targetInst = r, this.type = n, this.nativeEvent = i, this.target = a, this.currentTarget = null, e) e.hasOwnProperty(o) && (t = e[o], this[o] = t ? t(i) : i[o]);
			return this.isDefaultPrevented = (i.defaultPrevented == null ? !1 === i.returnValue : i.defaultPrevented) ? Hn : Un, this.isPropagationStopped = Un, this;
		}
		return S(t.prototype, {
			preventDefault: function() {
				this.defaultPrevented = !0;
				var e = this.nativeEvent;
				e && (e.preventDefault ? e.preventDefault() : typeof e.returnValue != "unknown" && (e.returnValue = !1), this.isDefaultPrevented = Hn);
			},
			stopPropagation: function() {
				var e = this.nativeEvent;
				e && (e.stopPropagation ? e.stopPropagation() : typeof e.cancelBubble != "unknown" && (e.cancelBubble = !0), this.isPropagationStopped = Hn);
			},
			persist: function() {},
			isPersistent: Hn
		}), t;
	}
	var Gn = {
		eventPhase: 0,
		bubbles: 0,
		cancelable: 0,
		timeStamp: function(e) {
			return e.timeStamp || Date.now();
		},
		defaultPrevented: 0,
		isTrusted: 0
	}, Kn = Wn(Gn), qn = S({}, Gn, {
		view: 0,
		detail: 0
	}), Jn = Wn(qn), Yn, Xn, Zn, Qn = S({}, qn, {
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
		getModifierState: lr,
		button: 0,
		buttons: 0,
		relatedTarget: function(e) {
			return e.relatedTarget === void 0 ? e.fromElement === e.srcElement ? e.toElement : e.fromElement : e.relatedTarget;
		},
		movementX: function(e) {
			return "movementX" in e ? e.movementX : (e !== Zn && (Zn && e.type === "mousemove" ? (Yn = e.screenX - Zn.screenX, Xn = e.screenY - Zn.screenY) : Xn = Yn = 0, Zn = e), Yn);
		},
		movementY: function(e) {
			return "movementY" in e ? e.movementY : Xn;
		}
	}), $n = Wn(Qn), er = Wn(S({}, Qn, { dataTransfer: 0 })), tr = Wn(S({}, qn, { relatedTarget: 0 })), nr = Wn(S({}, Gn, {
		animationName: 0,
		elapsedTime: 0,
		pseudoElement: 0
	})), rr = Wn(S({}, Gn, { clipboardData: function(e) {
		return "clipboardData" in e ? e.clipboardData : window.clipboardData;
	} })), ir = Wn(S({}, Gn, { data: 0 })), ar = {
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
	}, or = {
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
	}, sr = {
		Alt: "altKey",
		Control: "ctrlKey",
		Meta: "metaKey",
		Shift: "shiftKey"
	};
	function cr(e) {
		var t = this.nativeEvent;
		return t.getModifierState ? t.getModifierState(e) : (e = sr[e]) ? !!t[e] : !1;
	}
	function lr() {
		return cr;
	}
	var ur = Wn(S({}, qn, {
		key: function(e) {
			if (e.key) {
				var t = ar[e.key] || e.key;
				if (t !== "Unidentified") return t;
			}
			return e.type === "keypress" ? (e = Vn(e), e === 13 ? "Enter" : String.fromCharCode(e)) : e.type === "keydown" || e.type === "keyup" ? or[e.keyCode] || "Unidentified" : "";
		},
		code: 0,
		location: 0,
		ctrlKey: 0,
		shiftKey: 0,
		altKey: 0,
		metaKey: 0,
		repeat: 0,
		locale: 0,
		getModifierState: lr,
		charCode: function(e) {
			return e.type === "keypress" ? Vn(e) : 0;
		},
		keyCode: function(e) {
			return e.type === "keydown" || e.type === "keyup" ? e.keyCode : 0;
		},
		which: function(e) {
			return e.type === "keypress" ? Vn(e) : e.type === "keydown" || e.type === "keyup" ? e.keyCode : 0;
		}
	})), dr = Wn(S({}, Qn, {
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
	})), fr = Wn(S({}, Gn, { submitter: 0 })), pr = Wn(S({}, qn, {
		touches: 0,
		targetTouches: 0,
		changedTouches: 0,
		altKey: 0,
		metaKey: 0,
		ctrlKey: 0,
		shiftKey: 0,
		getModifierState: lr
	})), mr = Wn(S({}, Gn, {
		propertyName: 0,
		elapsedTime: 0,
		pseudoElement: 0
	})), hr = Wn(S({}, Qn, {
		deltaX: function(e) {
			return "deltaX" in e ? e.deltaX : "wheelDeltaX" in e ? -e.wheelDeltaX : 0;
		},
		deltaY: function(e) {
			return "deltaY" in e ? e.deltaY : "wheelDeltaY" in e ? -e.wheelDeltaY : "wheelDelta" in e ? -e.wheelDelta : 0;
		},
		deltaZ: 0,
		deltaMode: 0
	})), gr = Wn(S({}, Gn, {
		newState: 0,
		oldState: 0,
		source: 0
	})), _r = [
		9,
		13,
		27,
		32
	], vr = Pn && "CompositionEvent" in window, yr = null;
	Pn && "documentMode" in document && (yr = document.documentMode);
	var br = Pn && "TextEvent" in window && !yr, xr = Pn && (!vr || yr && 8 < yr && 11 >= yr), Sr = " ", Cr = !1;
	function wr(e, t) {
		switch (e) {
			case "keyup": return _r.indexOf(t.keyCode) !== -1;
			case "keydown": return t.keyCode !== 229;
			case "keypress":
			case "mousedown":
			case "focusout": return !0;
			default: return !1;
		}
	}
	function Tr(e) {
		return e = e.detail, typeof e == "object" && "data" in e ? e.data : null;
	}
	var Er = !1;
	function Dr(e, t) {
		switch (e) {
			case "compositionend": return Tr(t);
			case "keypress": return t.which === 32 ? (Cr = !0, Sr) : null;
			case "textInput": return e = t.data, e === Sr && Cr ? null : e;
			default: return null;
		}
	}
	function Or(e, t) {
		if (Er) return e === "compositionend" || !vr && wr(e, t) ? (e = Bn(), zn = Rn = Ln = null, Er = !1, e) : null;
		switch (e) {
			case "paste": return null;
			case "keypress":
				if (!(t.ctrlKey || t.altKey || t.metaKey) || t.ctrlKey && t.altKey) {
					if (t.char && 1 < t.char.length) return t.char;
					if (t.which) return String.fromCharCode(t.which);
				}
				return null;
			case "compositionend": return xr && t.locale !== "ko" ? null : t.data;
			default: return null;
		}
	}
	var kr = {
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
	function Ar(e) {
		var t = e && e.nodeName && e.nodeName.toLowerCase();
		return t === "input" ? !!kr[e.type] : t === "textarea";
	}
	function jr(e, t, n, r) {
		On ? kn ? kn.push(r) : kn = [r] : On = r, t = Jf(t, "onChange"), 0 < t.length && (n = new Kn("onChange", "change", null, n, r), e.push({
			event: n,
			listeners: t
		}));
	}
	var Mr = null, Nr = null;
	function Pr(e) {
		Vf(e, 0);
	}
	function Fr(e) {
		if (cn(Vt(e))) return e;
	}
	function Ir(e, t) {
		if (e === "change") return t;
	}
	var Lr = !1;
	if (Pn) {
		var Rr;
		if (Pn) {
			var zr = "oninput" in document;
			if (!zr) {
				var Br = document.createElement("div");
				Br.setAttribute("oninput", "return;"), zr = typeof Br.oninput == "function";
			}
			Rr = zr;
		} else Rr = !1;
		Lr = Rr && (!document.documentMode || 9 < document.documentMode);
	}
	function Vr() {
		Mr && (Mr.detachEvent("onpropertychange", Hr), Nr = Mr = null);
	}
	function Hr(e) {
		if (e.propertyName === "value" && Fr(Nr)) {
			var t = [];
			jr(t, Nr, e, Dn(e)), Mn(Pr, t);
		}
	}
	function Ur(e, t, n) {
		e === "focusin" ? (Vr(), Mr = t, Nr = n, Mr.attachEvent("onpropertychange", Hr)) : e === "focusout" && Vr();
	}
	function Wr(e) {
		if (e === "selectionchange" || e === "keyup" || e === "keydown") return Fr(Nr);
	}
	function Gr(e, t) {
		if (e === "click") return Fr(t);
	}
	function Kr(e, t) {
		if (e === "input" || e === "change") return Fr(t);
	}
	function qr(e, t) {
		return e === t && (e !== 0 || 1 / e == 1 / t) || e !== e && t !== t;
	}
	var Jr = typeof Object.is == "function" ? Object.is : qr;
	function Yr(e, t) {
		if (Jr(e, t)) return !0;
		if (typeof e != "object" || !e || typeof t != "object" || !t) return !1;
		var n = Object.keys(e), r = Object.keys(t);
		if (n.length !== r.length) return !1;
		for (r = 0; r < n.length; r++) {
			var i = n[r];
			if (!He.call(t, i) || !Jr(e[i], t[i])) return !1;
		}
		return !0;
	}
	function Xr(e) {
		if (e ||= typeof document < "u" ? document : void 0, e === void 0) return null;
		try {
			return e.activeElement || e.body;
		} catch {
			return e.body;
		}
	}
	function Zr(e) {
		for (; e && e.firstChild;) e = e.firstChild;
		return e;
	}
	function Qr(e, t) {
		var n = Zr(e);
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
			n = Zr(n);
		}
	}
	function $r(e, t) {
		return e && t ? e === t ? !0 : e && e.nodeType === 3 ? !1 : t && t.nodeType === 3 ? $r(e, t.parentNode) : "contains" in e ? e.contains(t) : e.compareDocumentPosition ? !!(e.compareDocumentPosition(t) & 16) : !1 : !1;
	}
	function ei(e) {
		e = e != null && e.ownerDocument != null && e.ownerDocument.defaultView != null ? e.ownerDocument.defaultView : window;
		for (var t = Xr(e.document); t instanceof e.HTMLIFrameElement;) {
			try {
				var n = typeof t.contentWindow.location.href == "string";
			} catch {
				n = !1;
			}
			if (n) e = t.contentWindow;
			else break;
			t = Xr(e.document);
		}
		return t;
	}
	function ti(e) {
		var t = e && e.nodeName && e.nodeName.toLowerCase();
		return t && (t === "input" && (e.type === "text" || e.type === "search" || e.type === "tel" || e.type === "url" || e.type === "password") || t === "textarea" || e.contentEditable === "true");
	}
	var ni = Pn && "documentMode" in document && 11 >= document.documentMode, ri = null, ii = null, ai = null, oi = !1;
	function si(e, t, n) {
		var r = n.window === n ? n.document : n.nodeType === 9 ? n : n.ownerDocument;
		oi || ri == null || ri !== Xr(r) || (r = ri, "selectionStart" in r && ti(r) ? r = {
			start: r.selectionStart,
			end: r.selectionEnd
		} : (r = (r.ownerDocument && r.ownerDocument.defaultView || window).getSelection(), r = {
			anchorNode: r.anchorNode,
			anchorOffset: r.anchorOffset,
			focusNode: r.focusNode,
			focusOffset: r.focusOffset
		}), ai && Yr(ai, r) || (ai = r, r = Jf(ii, "onSelect"), 0 < r.length && (t = new Kn("onSelect", "select", null, t, n), e.push({
			event: t,
			listeners: r
		}), t.target = ri)));
	}
	function ci(e, t) {
		var n = {};
		return n[e.toLowerCase()] = t.toLowerCase(), n["Webkit" + e] = "webkit" + t, n["Moz" + e] = "moz" + t, n;
	}
	var li = {
		animationend: ci("Animation", "AnimationEnd"),
		animationiteration: ci("Animation", "AnimationIteration"),
		animationstart: ci("Animation", "AnimationStart"),
		transitionrun: ci("Transition", "TransitionRun"),
		transitionstart: ci("Transition", "TransitionStart"),
		transitioncancel: ci("Transition", "TransitionCancel"),
		transitionend: ci("Transition", "TransitionEnd")
	}, ui = {}, di = {};
	Pn && (di = document.createElement("div").style, "AnimationEvent" in window || (delete li.animationend.animation, delete li.animationiteration.animation, delete li.animationstart.animation), "TransitionEvent" in window || delete li.transitionend.transition);
	function fi(e) {
		if (ui[e]) return ui[e];
		if (!li[e]) return e;
		var t = li[e], n;
		for (n in t) if (t.hasOwnProperty(n) && n in di) return ui[e] = t[n];
		return e;
	}
	var pi = fi("animationend"), mi = fi("animationiteration"), hi = fi("animationstart"), gi = fi("transitionrun"), _i = fi("transitionstart"), vi = fi("transitioncancel"), yi = fi("transitionend"), bi = /* @__PURE__ */ new Map(), xi = "abort auxClick beforeToggle cancel canPlay canPlayThrough click close contextMenu copy cut drag dragEnd dragEnter dragExit dragLeave dragOver dragStart drop durationChange emptied encrypted ended error fullscreenChange fullscreenError gotPointerCapture input invalid keyDown keyPress keyUp load loadedData loadedMetadata loadStart lostPointerCapture mouseDown mouseMove mouseOut mouseOver mouseUp paste pause play playing pointerCancel pointerDown pointerMove pointerOut pointerOver pointerUp progress rateChange reset resize seeked seeking stalled submit suspend timeUpdate touchCancel touchEnd touchStart volumeChange scroll toggle touchMove waiting wheel".split(" ");
	xi.push("scrollEnd");
	function Si(e, t) {
		bi.set(e, t), qt(t, [e]);
	}
	var Ci = 0;
	function wi(e, t) {
		if (e.name != null && e.name !== "auto") return e.name;
		if (t.autoName !== null) return t.autoName;
		e = bd.identifierPrefix;
		var n = Ci++;
		return e = "_" + e + "t_" + n.toString(32) + "_", t.autoName = e;
	}
	function Ti(e) {
		if (e == null || typeof e == "string") return e;
		var t = null, n = Od;
		if (n !== null) for (var r = 0; r < n.length; r++) {
			var i = e[n[r]];
			if (i != null) {
				if (i === "none") return "none";
				t = t == null ? i : t + (" " + i);
			}
		}
		return t ?? e.default;
	}
	function Ei(e, t) {
		return e = Ti(e), t = Ti(t), t == null ? e === "auto" ? null : e : t === "auto" ? null : t;
	}
	var Di = typeof reportError == "function" ? reportError : function(e) {
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
	}, Oi = [], ki = 0, Ai = 0;
	function ji() {
		for (var e = ki, t = Ai = ki = 0; t < e;) {
			var n = Oi[t];
			Oi[t++] = null;
			var r = Oi[t];
			Oi[t++] = null;
			var i = Oi[t];
			Oi[t++] = null;
			var a = Oi[t];
			if (Oi[t++] = null, r !== null && i !== null) {
				var o = r.pending;
				o === null ? i.next = i : (i.next = o.next, o.next = i), r.pending = i;
			}
			a !== 0 && Fi(n, i, a);
		}
	}
	function Mi(e, t, n, r) {
		Oi[ki++] = e, Oi[ki++] = t, Oi[ki++] = n, Oi[ki++] = r, Ai |= r, e.lanes |= r, e = e.alternate, e !== null && (e.lanes |= r);
	}
	function Ni(e, t, n, r) {
		return Mi(e, t, n, r), Ii(e);
	}
	function Pi(e, t) {
		return Mi(e, null, null, t), Ii(e);
	}
	function Fi(e, t, n) {
		e.lanes |= n;
		var r = e.alternate;
		r !== null && (r.lanes |= n);
		for (var i = !1, a = e.return; a !== null;) a.childLanes |= n, r = a.alternate, r !== null && (r.childLanes |= n), a.tag === 22 && (e = a.stateNode, e === null || e._visibility & 1 || (i = !0)), e = a, a = a.return;
		return e.tag === 3 ? (a = e.stateNode, i && t !== null && (i = 31 - at(n), e = a.hiddenUpdates, r = e[i], r === null ? e[i] = [t] : r.push(t), t.lane = n | 536870912), a) : null;
	}
	function Ii(e) {
		if (50 < kd) throw kd = 0, Ad = null, Error(o(185));
		for (var t = e.return; t !== null;) e = t, t = e.return;
		return e.tag === 3 ? e.stateNode : null;
	}
	var Li = {};
	function Ri(e, t, n, r) {
		this.tag = e, this.key = n, this.sibling = this.child = this.return = this.stateNode = this.type = this.elementType = null, this.index = 0, this.refCleanup = this.ref = null, this.pendingProps = t, this.dependencies = this.memoizedState = this.updateQueue = this.memoizedProps = null, this.mode = r, this.subtreeFlags = this.flags = 0, this.deletions = null, this.childLanes = this.lanes = 0, this.alternate = null;
	}
	function zi(e, t, n, r) {
		return new Ri(e, t, n, r);
	}
	function Bi(e) {
		return e = e.prototype, !(!e || !e.isReactComponent);
	}
	function Vi(e, t) {
		var n = e.alternate;
		return n === null ? (n = zi(e.tag, t, e.key, e.mode), n.elementType = e.elementType, n.type = e.type, n.stateNode = e.stateNode, n.alternate = e, e.alternate = n) : (n.pendingProps = t, n.type = e.type, n.flags = 0, n.subtreeFlags = 0, n.deletions = null), n.flags = e.flags & 1206910976, n.childLanes = e.childLanes, n.lanes = e.lanes, n.child = e.child, n.memoizedProps = e.memoizedProps, n.memoizedState = e.memoizedState, n.updateQueue = e.updateQueue, t = e.dependencies, n.dependencies = t === null ? null : {
			lanes: t.lanes,
			firstContext: t.firstContext
		}, n.sibling = e.sibling, n.index = e.index, n.ref = e.ref, n.refCleanup = e.refCleanup, n;
	}
	function Hi(e, t) {
		e.flags &= 1206910978;
		var n = e.alternate;
		return n === null ? (e.childLanes = 0, e.lanes = t, e.child = null, e.subtreeFlags = 0, e.memoizedProps = null, e.memoizedState = null, e.updateQueue = null, e.dependencies = null, e.stateNode = null) : (e.childLanes = n.childLanes, e.lanes = n.lanes, e.child = n.child, e.subtreeFlags = 0, e.deletions = null, e.memoizedProps = n.memoizedProps, e.memoizedState = n.memoizedState, e.updateQueue = n.updateQueue, e.type = n.type, t = n.dependencies, e.dependencies = t === null ? null : {
			lanes: t.lanes,
			firstContext: t.firstContext
		}), e;
	}
	function Ui(e, t, n, r, i, a) {
		var s = 0;
		if (r = e, typeof r == "function") Bi(r) && (s = 1);
		else if (typeof r == "string") s = qm(e, n, De.current) ? 26 : e === "html" || e === "head" || e === "body" ? 27 : 5;
		else a: switch (r) {
			case T: return e = zi(31, n, t, i), e.elementType = T, e.lanes = a, e;
			case se: return Wi(n.children, i, a, t);
			case ce:
				s = 8, i |= 24;
				break;
			case le: return e = zi(12, n, t, i | 2), e.elementType = le, e.lanes = a, e;
			case fe: return e = zi(13, n, t, i), e.elementType = fe, e.lanes = a, e;
			case w: return e = zi(19, n, t, i), e.elementType = w, e.lanes = a, e;
			case he:
			case _e: return e = i | 32, e = zi(30, n, t, e), e.elementType = _e, e.lanes = a, e.stateNode = {
				autoName: null,
				paired: null,
				clones: null,
				ref: null
			}, e;
			default:
				if (typeof r == "object" && r) switch (r.$$typeof) {
					case de:
						s = 10;
						break a;
					case ue:
						s = 9;
						break a;
					case C:
						s = 11;
						break a;
					case pe:
						s = 14;
						break a;
					case me:
						s = 16, r = null;
						break a;
				}
				s = 29, n = Error(o(130, e === null ? "null" : typeof e, "")), r = null;
		}
		return t = zi(s, n, t, i), t.elementType = e, t.type = r, t.lanes = a, t;
	}
	function Wi(e, t, n, r) {
		return e = zi(7, e, r, t), e.lanes = n, e;
	}
	function Gi(e, t, n) {
		return e = zi(6, e, null, t), e.lanes = n, e;
	}
	function Ki(e) {
		var t = zi(18, null, null, 0);
		return t.stateNode = e, t;
	}
	function qi(e, t, n) {
		return t = zi(4, e.children === null ? [] : e.children, e.key, t), t.lanes = n, t.stateNode = {
			containerInfo: e.containerInfo,
			pendingChildren: null,
			implementation: e.implementation
		}, t;
	}
	var Ji = /* @__PURE__ */ new WeakMap();
	function Yi(e, t) {
		if (typeof e == "object" && e) {
			var n = Ji.get(e);
			return n === void 0 ? (t = {
				value: e,
				source: t,
				stack: Ve(t)
			}, Ji.set(e, t), t) : n;
		}
		return {
			value: e,
			source: t,
			stack: Ve(t)
		};
	}
	var Xi = [], Zi = 0, Qi = null, $i = 0, ea = [], ta = 0, na = null, ra = 1, ia = "";
	function aa(e, t) {
		Xi[Zi++] = $i, Xi[Zi++] = Qi, Qi = e, $i = t;
	}
	function oa(e, t, n) {
		ea[ta++] = ra, ea[ta++] = ia, ea[ta++] = na, na = e;
		var r = ra;
		e = ia;
		var i = 32 - at(r) - 1;
		r &= ~(1 << i), n += 1;
		var a = 32 - at(t) + i;
		if (30 < a) {
			var o = i - i % 5;
			a = (r & (1 << o) - 1).toString(32), r >>= o, i -= o, ra = 1 << 32 - at(t) + i | n << i | r, ia = a + e;
		} else ra = 1 << a | n << i | r, ia = e;
	}
	function sa(e) {
		e.return !== null && (aa(e, 1), oa(e, 1, 0));
	}
	function ca(e) {
		for (; e === Qi;) Qi = Xi[--Zi], Xi[Zi] = null, $i = Xi[--Zi], Xi[Zi] = null;
		for (; e === na;) na = ea[--ta], ea[ta] = null, ia = ea[--ta], ea[ta] = null, ra = ea[--ta], ea[ta] = null;
	}
	function la(e, t) {
		ea[ta++] = ra, ea[ta++] = ia, ea[ta++] = na, ra = t.id, ia = t.overflow, na = e;
	}
	var ua = null, N = null, P = !1, da = null, fa = !1, pa = Error(o(519));
	function ma(e) {
		throw ba(Yi(Error(o(418, 1 < arguments.length && arguments[1] !== void 0 && arguments[1] ? "text" : "HTML", "")), e)), pa;
	}
	function ha(e) {
		var t = e.stateNode, n = e.type, r = e.memoizedProps;
		switch (t[kt] = e, t[At] = r, n) {
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
				for (n = 0; n < zf.length; n++) Q(zf[n], t);
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
				Q("invalid", t), fn(t, r.value, r.defaultValue, r.checked, r.defaultChecked, r.type, r.name, !0);
				break;
			case "select":
				Q("invalid", t);
				break;
			case "textarea": Q("invalid", t), gn(t, r.value, r.defaultValue, r.children);
		}
		n = r.children, typeof n != "string" && typeof n != "number" && typeof n != "bigint" || t.textContent === "" + n || !0 === r.suppressHydrationWarning || ep(t.textContent, n) ? (r.popover != null && (Q("beforetoggle", t), Q("toggle", t)), r.onScroll != null && Q("scroll", t), r.onScrollEnd != null && Q("scrollend", t), r.onClick != null && (t.onclick = Tn), t = !0) : t = !1, t || ma(e, !0);
	}
	function ga(e) {
		for (ua = e.return; ua;) switch (ua.tag) {
			case 5:
			case 31:
			case 13:
				fa = !1;
				return;
			case 27:
			case 3:
				fa = !0;
				return;
			default: ua = ua.return;
		}
	}
	function _a(e) {
		if (e !== ua) return !1;
		if (!P) return ga(e), P = !0, !1;
		var t = e.tag, n;
		if ((n = t !== 3 && t !== 27) && ((n = t === 5) && (n = e.type, n = n === "form" || n === "button" || pp(e.type, e.memoizedProps)), n = !n), n && N && ma(e), ga(e), t === 13) {
			if (e = e.memoizedState, e = e === null ? null : e.dehydrated, !e) throw Error(o(317));
			N = dm(e);
		} else if (t === 31) {
			if (e = e.memoizedState, e = e === null ? null : e.dehydrated, !e) throw Error(o(317));
			N = dm(e);
		} else t === 27 ? (t = N, Sp(e.type) ? (e = um, um = null, N = e) : N = t) : N = ua ? lm(e.stateNode.nextSibling) : null;
		return !0;
	}
	function va() {
		N = ua = null, P = !1;
	}
	function ya() {
		var e = da;
		return e !== null && (pd === null ? pd = e : pd.push.apply(pd, e), da = null), e;
	}
	function ba(e) {
		da === null ? da = [e] : da.push(e);
	}
	var xa = Te(null), Sa = null, Ca = null;
	function wa(e, t, n) {
		A(xa, t._currentValue), t._currentValue = n;
	}
	function Ta(e) {
		e._currentValue = xa.current, Ee(xa);
	}
	function Ea(e, t, n) {
		for (; e !== null;) {
			var r = e.alternate;
			if ((e.childLanes & t) === t ? r !== null && (r.childLanes & t) !== t && (r.childLanes |= t) : (e.childLanes |= t, r !== null && (r.childLanes |= t)), e === n) break;
			e = e.return;
		}
	}
	function Da(e, t, n, r) {
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
						a.lanes |= n, c = a.alternate, c !== null && (c.lanes |= n), Ea(a.return, n, e), r || (s = null);
						break a;
					}
					a = c.next;
				}
			} else if (i.tag === 18) {
				if (s = i.return, s === null) throw Error(o(341));
				s.lanes |= n, a = s.alternate, a !== null && (a.lanes |= n), Ea(s, n, e), s = null;
			} else i.tag === 13 && i.memoizedState !== null && i.memoizedState.dehydrated === null ? (i.lanes |= n, s = i.alternate, s !== null && (s.lanes |= n), Ea(i.return, n, e), s = i.child, s = s === null ? null : s.sibling) : s = i.child;
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
	function Oa(e, t, n, r) {
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
					Jr(i.pendingProps.value, s.value) || (e === null ? e = [c] : e.push(c));
				}
			} else if (i === Ae.current) {
				if (s = i.alternate, s === null) throw Error(o(387));
				s.memoizedState.memoizedState !== i.memoizedState.memoizedState && (e === null ? e = [sh] : e.push(sh));
			}
			i = i.return;
		}
		return e !== null && Da(t, e, n, r), t.flags |= 262144, e !== null;
	}
	function ka(e) {
		for (e = e.firstContext; e !== null;) {
			if (!Jr(e.context._currentValue, e.memoizedValue)) return !0;
			e = e.next;
		}
		return !1;
	}
	function Aa(e) {
		Sa = e, Ca = null, e = e.dependencies, e !== null && (e.firstContext = null);
	}
	function ja(e) {
		return Na(Sa, e);
	}
	function Ma(e, t) {
		return Sa === null && Aa(e), Na(e, t);
	}
	function Na(e, t) {
		var n = t._currentValue;
		if (t = {
			context: t,
			memoizedValue: n,
			next: null
		}, Ca === null) {
			if (e === null) throw Error(o(308));
			Ca = t, e.dependencies = {
				lanes: 0,
				firstContext: t
			}, e.flags |= 524288;
		} else Ca = Ca.next = t;
		return n;
	}
	var Pa = typeof AbortController < "u" ? AbortController : function() {
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
	}, Fa = t.unstable_scheduleCallback, Ia = t.unstable_NormalPriority, La = {
		$$typeof: de,
		Consumer: null,
		Provider: null,
		_currentValue: null,
		_currentValue2: null,
		_threadCount: 0
	};
	function Ra() {
		return {
			controller: new Pa(),
			data: /* @__PURE__ */ new Map(),
			refCount: 0
		};
	}
	function za(e) {
		e.refCount--, e.refCount === 0 && Fa(Ia, function() {
			e.controller.abort();
		});
	}
	function Ba(e, t) {
		if (e.pendingLanes & 4194048) {
			var n = e.transitionTypes;
			for (n === null && (n = e.transitionTypes = []), e = 0; e < t.length; e++) {
				var r = t[e];
				n.indexOf(r) === -1 && n.push(r);
			}
		}
	}
	var Va = null;
	function Ha(e) {
		var t = e.transitionTypes;
		return e.transitionTypes = null, t;
	}
	var Ua = null, Wa = 0, Ga = 0, Ka = null;
	function qa(e, t) {
		if (Ua === null) {
			var n = Ua = [];
			Wa = 0, Ga = Pf(), Ka = {
				status: "pending",
				value: void 0,
				then: function(e) {
					n.push(e);
				}
			};
		}
		return Wa++, t.then(Ja, Ja), t;
	}
	function Ja() {
		if (--Wa === 0 && (Va = null, Ua !== null)) {
			Ka !== null && (Ka.status = "fulfilled");
			var e = Ua;
			Ua = null, Ga = 0, Ka = null;
			for (var t = 0; t < e.length; t++) (0, e[t])();
		}
	}
	function Ya(e, t) {
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
	var Xa = O.S;
	O.S = function(e, t) {
		if (gd = qe(), typeof t == "object" && t && typeof t.then == "function" && qa(e, t), Va !== null) for (var n = bf; n !== null;) Ba(n, Va), n = n.next;
		if (n = e.types, n !== null) {
			for (var r = bf; r !== null;) Ba(r, n), r = r.next;
			if (Ga !== 0) {
				r = Va, r === null && (r = Va = []);
				for (var i = 0; i < n.length; i++) {
					var a = n[i];
					r.indexOf(a) === -1 && r.push(a);
				}
			}
		}
		Xa !== null && Xa(e, t);
	};
	var Za = Te(null);
	function Qa() {
		var e = Za.current;
		return e === null ? G.pooledCache : e;
	}
	function $a(e, t) {
		t === null ? A(Za, Za.current) : A(Za, t.pool);
	}
	function eo() {
		var e = Qa();
		return e === null ? null : {
			parent: La._currentValue,
			pool: e
		};
	}
	var to = Error(o(460)), no = Error(o(474)), ro = Error(o(542)), io = { then: function() {} };
	function ao(e) {
		return e = e.status, e === "fulfilled" || e === "rejected";
	}
	function oo(e, t, n) {
		switch (n = e[n], n === void 0 ? e.push(t) : n !== t && (t.then(Tn, Tn), t = n), t.status) {
			case "fulfilled": return t.value;
			case "rejected": throw e = t.reason, uo(e), e === void 0 && !("reason" in t) ? Error(o(600)) : e;
			default:
				if (typeof t.status == "string") t.then(Tn, Tn);
				else {
					if (e = G, e !== null && 100 < e.shellSuspendCounter) throw Error(o(482));
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
					case "rejected": throw e = t.reason, uo(e), e;
				}
				throw co = t, to;
		}
	}
	function so(e) {
		try {
			var t = e._init;
			return t(e._payload);
		} catch (e) {
			throw typeof e == "object" && e && typeof e.then == "function" ? (co = e, to) : e;
		}
	}
	var co = null;
	function lo() {
		if (co === null) throw Error(o(459));
		var e = co;
		return co = null, e;
	}
	function uo(e) {
		if (e === to || e === ro) throw Error(o(483));
	}
	var fo = null, po = 0;
	function mo(e) {
		var t = po;
		return po += 1, fo === null && (fo = []), oo(fo, e, t);
	}
	function ho(e, t) {
		t = t.props.ref, e.ref = t === void 0 ? null : t;
	}
	function go(e, t) {
		throw t.$$typeof === ie ? Error(o(525)) : (e = Object.prototype.toString.call(t), Error(o(31, e === "[object Object]" ? "object with keys {" + Object.keys(t).join(", ") + "}" : e)));
	}
	function _o(e) {
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
			return e = Vi(e, t), e.index = 0, e.sibling = null, e;
		}
		function a(t, n, r) {
			return t.index = r, e ? (r = t.alternate, r === null ? (t.flags |= 134217730, n) : (r = r.index, r < n ? (t.flags |= 2, n) : r)) : (t.flags |= 1048576, n);
		}
		function s(t) {
			return e && t.alternate === null && (t.flags |= 134217730), t;
		}
		function c(e, t, n, r) {
			return t === null || t.tag !== 6 ? (t = Gi(n, e.mode, r), t.return = e, t) : (t = i(t, n), t.return = e, t);
		}
		function l(e, t, n, r) {
			var a = n.type;
			return a === se ? (e = d(e, t, n.props.children, r, n.key), ho(e, n), e) : t !== null && (t.elementType === a || typeof a == "object" && a && a.$$typeof === me && so(a) === t.type) ? (t = i(t, n.props), ho(t, n), t.return = e, t) : (t = Ui(n.type, n.key, n.props, null, e.mode, r), ho(t, n), t.return = e, t);
		}
		function u(e, t, n, r) {
			return t === null || t.tag !== 4 || t.stateNode.containerInfo !== n.containerInfo || t.stateNode.implementation !== n.implementation ? (t = qi(n, e.mode, r), t.return = e, t) : (t = i(t, n.children || []), t.return = e, t);
		}
		function d(e, t, n, r, a) {
			return t === null || t.tag !== 7 ? (t = Wi(n, e.mode, r, a), t.return = e, t) : (t = i(t, n), t.return = e, t);
		}
		function f(e, t, n) {
			if (typeof t == "string" && t !== "" || typeof t == "number" || typeof t == "bigint") return t = Gi("" + t, e.mode, n), t.return = e, t;
			if (typeof t == "object" && t) {
				switch (t.$$typeof) {
					case ae: return n = Ui(t.type, t.key, t.props, null, e.mode, n), ho(n, t), n.return = e, n;
					case oe: return t = qi(t, e.mode, n), t.return = e, t;
					case me: return t = so(t), f(e, t, n);
				}
				if (xe(t) || D(t)) return t = Wi(t, e.mode, n, null), t.return = e, t;
				if (typeof t.then == "function") return f(e, mo(t), n);
				if (t.$$typeof === de) return f(e, Ma(e, t), n);
				go(e, t);
			}
			return null;
		}
		function p(e, t, n, r) {
			var i = t === null ? null : t.key;
			if (typeof n == "string" && n !== "" || typeof n == "number" || typeof n == "bigint") return i === null ? c(e, t, "" + n, r) : null;
			if (typeof n == "object" && n) {
				switch (n.$$typeof) {
					case ae: return n.key === i ? l(e, t, n, r) : null;
					case oe: return n.key === i ? u(e, t, n, r) : null;
					case me: return n = so(n), p(e, t, n, r);
				}
				if (xe(n) || D(n)) return i === null ? d(e, t, n, r, null) : null;
				if (typeof n.then == "function") return p(e, t, mo(n), r);
				if (n.$$typeof === de) return p(e, t, Ma(e, n), r);
				go(e, n);
			}
			return null;
		}
		function m(e, t, n, r, i) {
			if (typeof r == "string" && r !== "" || typeof r == "number" || typeof r == "bigint") return e = e.get(n) || null, c(t, e, "" + r, i);
			if (typeof r == "object" && r) {
				switch (r.$$typeof) {
					case ae: return e = e.get(r.key === null ? n : r.key) || null, l(t, e, r, i);
					case oe: return e = e.get(r.key === null ? n : r.key) || null, u(t, e, r, i);
					case me: return r = so(r), m(e, t, n, r, i);
				}
				if (xe(r) || D(r)) return e = e.get(n) || null, d(t, e, r, i, null);
				if (typeof r.then == "function") return m(e, t, n, mo(r), i);
				if (r.$$typeof === de) return m(e, t, n, Ma(t, r), i);
				go(t, r);
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
			if (h === s.length) return n(i, d), P && aa(i, h), l;
			if (d === null) {
				for (; h < s.length; h++) d = f(i, s[h], c), d !== null && (o = a(d, o, h), u === null ? l = d : u.sibling = d, u = d);
				return P && aa(i, h), l;
			}
			for (d = r(d); h < s.length; h++) g = m(d, i, h, s[h], c), g !== null && (e && (_ = g.alternate, _ !== null && d.delete(_.key === null ? h : _.key)), o = a(g, o, h), u === null ? l = g : u.sibling = g, u = g);
			return e && d.forEach(function(e) {
				return t(i, e);
			}), P && aa(i, h), l;
		}
		function g(i, s, c, l) {
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
			if (v.done) return n(i, h), P && aa(i, g), u;
			if (h === null) {
				for (; !v.done; g++, v = c.next()) v = f(i, v.value, l), v !== null && (s = a(v, s, g), d === null ? u = v : d.sibling = v, d = v);
				return P && aa(i, g), u;
			}
			for (h = r(h); !v.done; g++, v = c.next()) v = m(h, i, g, v.value, l), v !== null && (e && (_ = v.alternate, _ !== null && h.delete(_.key === null ? g : _.key)), s = a(v, s, g), d === null ? u = v : d.sibling = v, d = v);
			return e && h.forEach(function(e) {
				return t(i, e);
			}), P && aa(i, g), u;
		}
		function _(e, r, a, c) {
			if (typeof a == "object" && a && a.type === se && a.key === null && a.props.ref === void 0 && (a = a.props.children), typeof a == "object" && a) {
				switch (a.$$typeof) {
					case ae:
						a: {
							for (var l = a.key; r !== null;) {
								if (r.key === l) {
									if (l = a.type, l === se) {
										if (r.tag === 7) {
											n(e, r.sibling), c = i(r, a.props.children), ho(c, a), c.return = e, e = c;
											break a;
										}
									} else if (r.elementType === l || typeof l == "object" && l && l.$$typeof === me && so(l) === r.type) {
										n(e, r.sibling), c = i(r, a.props), ho(c, a), c.return = e, e = c;
										break a;
									}
									n(e, r);
									break;
								}
								t(e, r), r = r.sibling;
							}
							a.type === se ? (c = Wi(a.props.children, e.mode, c, a.key), ho(c, a), c.return = e, e = c) : (c = Ui(a.type, a.key, a.props, null, e.mode, c), ho(c, a), c.return = e, e = c);
						}
						return s(e);
					case oe:
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
							c = qi(a, e.mode, c), c.return = e, e = c;
						}
						return s(e);
					case me: return a = so(a), _(e, r, a, c);
				}
				if (xe(a)) return h(e, r, a, c);
				if (D(a)) {
					if (l = D(a), typeof l != "function") throw Error(o(150));
					return a = l.call(a), g(e, r, a, c);
				}
				if (typeof a.then == "function") return _(e, r, mo(a), c);
				if (a.$$typeof === de) return _(e, r, Ma(e, a), c);
				go(e, a);
			}
			return typeof a == "string" && a !== "" || typeof a == "number" || typeof a == "bigint" ? (a = "" + a, r !== null && r.tag === 6 ? (n(e, r.sibling), c = i(r, a), c.return = e, e = c) : (n(e, r), c = Gi(a, e.mode, c), c.return = e, e = c), s(e)) : n(e, r);
		}
		return function(e, t, n, r) {
			try {
				po = 0;
				var i = _(e, t, n, r);
				return fo = null, i;
			} catch (t) {
				if (t === to || t === ro) throw t;
				var a = zi(29, t, null, e.mode);
				return a.lanes = r, a.return = e, a;
			}
		};
	}
	var vo = _o(!0), yo = _o(!1), bo = !1;
	function xo(e) {
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
	function So(e, t) {
		e = e.updateQueue, t.updateQueue === e && (t.updateQueue = {
			baseState: e.baseState,
			firstBaseUpdate: e.firstBaseUpdate,
			lastBaseUpdate: e.lastBaseUpdate,
			shared: e.shared,
			callbacks: null
		});
	}
	function Co(e) {
		return {
			lane: e,
			tag: 0,
			payload: null,
			callback: null,
			next: null
		};
	}
	function wo(e, t, n) {
		var r = e.updateQueue;
		if (r === null) return null;
		if (r = r.shared, W & 2) {
			var i = r.pending;
			return i === null ? t.next = t : (t.next = i.next, i.next = t), r.pending = t, t = Ii(e), Fi(e, null, n), t;
		}
		return Mi(e, r, t, n), Ii(e);
	}
	function To(e, t, n) {
		if (t = t.updateQueue, t !== null && (t = t.shared, n & 4194048)) {
			var r = t.lanes;
			r &= e.pendingLanes, n |= r, t.lanes = n, St(e, n);
		}
	}
	function Eo(e, t) {
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
	var Do = !1;
	function Oo() {
		if (Do) {
			var e = Ka;
			if (e !== null) throw e;
		}
	}
	function ko(e, t, n, r) {
		Do = !1;
		var i = e.updateQueue;
		bo = !1;
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
				if (p ? (q & f) === f : (r & f) === f) {
					f !== 0 && f === Ga && (Do = !0), u !== null && (u = u.next = {
						lane: 0,
						tag: s.tag,
						payload: s.payload,
						callback: null,
						next: null
					});
					a: {
						var m = e, h = s;
						f = t;
						var g = n;
						switch (h.tag) {
							case 1:
								if (m = h.payload, typeof m == "function") {
									d = m.call(g, d, f);
									break a;
								}
								d = m;
								break a;
							case 3: m.flags = m.flags & -65537 | 128;
							case 0:
								if (m = h.payload, f = typeof m == "function" ? m.call(g, d, f) : m, f == null) break a;
								d = S({}, d, f);
								break a;
							case 2: bo = !0;
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
			u === null && (c = d), i.baseState = c, i.firstBaseUpdate = l, i.lastBaseUpdate = u, a === null && (i.shared.lanes = 0), sd |= o, e.lanes = o, e.memoizedState = d;
		}
	}
	function Ao(e, t) {
		if (typeof e != "function") throw Error(o(191, e));
		e.call(t);
	}
	function jo(e, t) {
		var n = e.callbacks;
		if (n !== null) for (e.callbacks = null, e = 0; e < n.length; e++) Ao(n[e], t);
	}
	var Mo = Te(null), No = Te(0);
	function Po(e, t) {
		e = od, A(No, e), A(Mo, t), od = e | t.baseLanes;
	}
	function Fo() {
		A(No, od), A(Mo, Mo.current);
	}
	function Io() {
		od = No.current, Ee(Mo), Ee(No);
	}
	var Lo = Te(null), Ro = null;
	function zo(e) {
		var t = e.alternate;
		A(Wo, Wo.current & 1), A(Lo, e), Ro === null && (t === null || Mo.current !== null || t.memoizedState !== null) && (Ro = e);
	}
	function Bo(e) {
		A(Wo, Wo.current), A(Lo, e), Ro === null && (Ro = e);
	}
	function Vo(e) {
		e.tag === 22 ? (A(Wo, Wo.current), A(Lo, e), Ro === null && (Ro = e)) : Ho();
	}
	function Ho() {
		A(Wo, Wo.current), A(Lo, Lo.current);
	}
	function Uo(e) {
		Ee(Lo), Ro === e && (Ro = null), Ee(Wo);
	}
	var Wo = Te(0);
	function Go(e, t) {
		A(Lo, Lo.current), A(Wo, t);
	}
	function F(e) {
		Ee(Wo), Ee(Lo), Ro === e && (Ro = null);
	}
	function Ko(e) {
		for (var t = e; t !== null;) {
			if (t.tag === 13) {
				var n = t.memoizedState;
				if (n !== null && (n = n.dehydrated, n === null || om(n) || sm(n))) return t;
			} else if (t.tag === 19 && t.memoizedProps.revealOrder !== "independent") {
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
	var qo = 0, I = null, L = null, R = null, Jo = !1, Yo = !1, Xo = !1, Zo = 0, Qo = 0, $o = null, es = 0;
	function z() {
		throw Error(o(321));
	}
	function ts(e, t) {
		if (t === null) return !1;
		for (var n = 0; n < t.length && n < e.length; n++) if (!Jr(e[n], t[n])) return !1;
		return !0;
	}
	function ns(e, t, n, r, i, a) {
		return qo = a, I = t, t.memoizedState = null, t.updateQueue = null, t.lanes = 0, O.H = e === null || e.memoizedState === null ? vc : yc, Xo = !1, a = n(r, i), Xo = !1, Yo && (a = is(t, n, r, i)), rs(e), a;
	}
	function rs(e) {
		O.H = _c;
		var t = L !== null && L.next !== null;
		if (qo = 0, R = L = I = null, Jo = !1, Qo = 0, $o = null, t) throw Error(o(300));
		e === null || Ic || (e = e.dependencies, e !== null && ka(e) && (Ic = !0));
	}
	function is(e, t, n, r) {
		I = e;
		var i = 0;
		do {
			if (Yo && ($o = null), Qo = 0, Yo = !1, 25 <= i) throw Error(o(301));
			if (i += 1, R = L = null, e.updateQueue != null) {
				var a = e.updateQueue;
				a.lastEffect = null, a.events = null, a.stores = null, a.memoCache != null && (a.memoCache.index = 0);
			}
			O.H = bc, a = t(n, r);
		} while (Yo);
		return a;
	}
	function as() {
		var e = O.H, t = e.useState()[0];
		return t = typeof t.then == "function" ? ds(t) : t, e = e.useState()[0], (L === null ? null : L.memoizedState) !== e && (I.flags |= 1024), t;
	}
	function os() {
		var e = Zo !== 0;
		return Zo = 0, e;
	}
	function ss(e, t, n) {
		t.updateQueue = e.updateQueue, t.flags &= -2053, e.lanes &= ~n;
	}
	function cs(e) {
		if (Jo) {
			for (e = e.memoizedState; e !== null;) {
				var t = e.queue;
				t !== null && (t.pending = null), e = e.next;
			}
			Jo = !1;
		}
		qo = 0, R = L = I = null, Yo = !1, Qo = Zo = 0, $o = null;
	}
	function ls() {
		var e = {
			memoizedState: null,
			baseState: null,
			baseQueue: null,
			queue: null,
			next: null
		};
		return R === null ? I.memoizedState = R = e : R = R.next = e, R;
	}
	function B() {
		if (L === null) {
			var e = I.alternate;
			e = e === null ? null : e.memoizedState;
		} else e = L.next;
		var t = R === null ? I.memoizedState : R.next;
		if (t !== null) R = t, L = e;
		else {
			if (e === null) throw I.alternate === null ? Error(o(467)) : Error(o(310));
			L = e, e = {
				memoizedState: L.memoizedState,
				baseState: L.baseState,
				baseQueue: L.baseQueue,
				queue: L.queue,
				next: null
			}, R === null ? I.memoizedState = R = e : R = R.next = e;
		}
		return R;
	}
	function us() {
		return {
			lastEffect: null,
			events: null,
			stores: null,
			memoCache: null
		};
	}
	function ds(e) {
		var t = Qo;
		return Qo += 1, $o === null && ($o = []), e = oo($o, e, t), t = I, (R === null ? t.memoizedState : R.next) === null && (t = t.alternate, O.H = t === null || t.memoizedState === null ? vc : yc), e;
	}
	function fs(e) {
		if (typeof e == "object" && e) {
			if (typeof e.then == "function") return ds(e);
			if (e.$$typeof === E) return;
			if (e.$$typeof === de) return ja(e);
		}
		throw Error(o(438, String(e)));
	}
	function ps(e) {
		var t = null, n = I.updateQueue;
		if (n !== null && (t = n.memoCache), t == null) {
			var r = I.alternate;
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
		}, n === null && (n = us(), I.updateQueue = n), n.memoCache = t, n = t.data[t.index], n === void 0) for (n = t.data[t.index] = Array(e), r = 0; r < e; r++) n[r] = ge;
		return t.index++, n;
	}
	function ms(e, t) {
		return typeof t == "function" ? t(e) : t;
	}
	function hs(e) {
		return gs(B(), L, e);
	}
	function gs(e, t, n) {
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
				if (f === u.lane ? (qo & f) === f : (q & f) === f) {
					var p = u.revertLane;
					if (p === 0) l !== null && (l = l.next = {
						lane: 0,
						revertLane: 0,
						gesture: null,
						action: u.action,
						hasEagerState: u.hasEagerState,
						eagerState: u.eagerState,
						next: null
					}), f === Ga && (d = !0);
					else if ((qo & p) === p) {
						u = u.next, p === Ga && (d = !0);
						continue;
					} else f = {
						lane: 0,
						revertLane: u.revertLane,
						gesture: null,
						action: u.action,
						hasEagerState: u.hasEagerState,
						eagerState: u.eagerState,
						next: null
					}, l === null ? (c = l = f, s = a) : l = l.next = f, I.lanes |= p, sd |= p;
					f = u.action, Xo && n(a, f), a = u.hasEagerState ? u.eagerState : n(a, f);
				} else p = {
					lane: f,
					revertLane: u.revertLane,
					gesture: u.gesture,
					action: u.action,
					hasEagerState: u.hasEagerState,
					eagerState: u.eagerState,
					next: null
				}, l === null ? (c = l = p, s = a) : l = l.next = p, I.lanes |= f, sd |= f;
				u = u.next;
			} while (u !== null && u !== t);
			if (l === null ? s = a : l.next = c, !Jr(a, e.memoizedState) && (Ic = !0, d && (n = Ka, n !== null))) throw n;
			e.memoizedState = a, e.baseState = s, e.baseQueue = l, r.lastRenderedState = a;
		}
		return i === null && (r.lanes = 0), [e.memoizedState, r.dispatch];
	}
	function _s(e) {
		var t = B(), n = t.queue;
		if (n === null) throw Error(o(311));
		n.lastRenderedReducer = e;
		var r = n.dispatch, i = n.pending, a = t.memoizedState;
		if (i !== null) {
			n.pending = null;
			var s = i = i.next;
			do
				a = e(a, s.action), s = s.next;
			while (s !== i);
			Jr(a, t.memoizedState) || (Ic = !0), t.memoizedState = a, t.baseQueue === null && (t.baseState = a), n.lastRenderedState = a;
		}
		return [a, r];
	}
	function vs(e, t, n) {
		var r = I, i = B(), a = P;
		if (a) {
			if (n === void 0) throw Error(o(407));
			n = n();
		} else n = t();
		var s = !Jr((L || i).memoizedState, n);
		if (s && (i.memoizedState = n, Ic = !0), i = i.queue, Us(xs.bind(null, r, i, e), [e]), e = i.getSnapshot !== t || s || R !== null && !!(R.memoizedState.tag & 1), Rs(e ? 9 : 8, { destroy: void 0 }, bs.bind(null, r, i, n, t), null), e) {
			if (r.flags |= 2048, G === null) throw Error(o(349));
			a || qo & 127 || ys(r, t, n);
		}
		return n;
	}
	function ys(e, t, n) {
		e.flags |= 16384, e = {
			getSnapshot: t,
			value: n
		}, t = I.updateQueue, t === null ? (t = us(), I.updateQueue = t, t.stores = [e]) : (n = t.stores, n === null ? t.stores = [e] : n.push(e));
	}
	function bs(e, t, n, r) {
		t.value = n, t.getSnapshot = r, Ss(t) && Cs(e);
	}
	function xs(e, t, n) {
		return n(function() {
			Ss(t) && Cs(e);
		});
	}
	function Ss(e) {
		var t = e.getSnapshot;
		e = e.value;
		try {
			var n = t();
			return !Jr(e, n);
		} catch {
			return !0;
		}
	}
	function Cs(e) {
		var t = Pi(e, 2);
		t !== null && Pd(t, e, 2);
	}
	function ws(e) {
		var t = ls();
		if (typeof e == "function") {
			var n = e;
			if (e = n(), Xo) {
				it(!0);
				try {
					n();
				} finally {
					it(!1);
				}
			}
		}
		return t.memoizedState = t.baseState = e, t.queue = {
			pending: null,
			lanes: 0,
			dispatch: null,
			lastRenderedReducer: ms,
			lastRenderedState: e
		}, t;
	}
	function Ts(e, t, n, r) {
		return e.baseState = n, gs(e, L, typeof r == "function" ? r : ms);
	}
	function Es(e, t, n, r, i) {
		if (mc(e)) throw Error(o(485));
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
			O.T === null ? a.isTransition = !1 : n(!0), r(a), n = t.pending, n === null ? (a.next = t.pending = a, Ds(t, a)) : (a.next = n.next, t.pending = n.next = a);
		}
	}
	function Ds(e, t) {
		var n = t.action, r = t.payload, i = e.state;
		if (t.isTransition) {
			var a = O.T, o = {};
			o.types = a === null ? null : a.types, O.T = o;
			try {
				var s = n(i, r), c = O.S;
				c !== null && c(o, s), Os(e, t, s);
			} catch (n) {
				As(e, t, n);
			} finally {
				a !== null && o.types !== null && (a.types = o.types), O.T = a;
			}
		} else try {
			a = n(i, r), Os(e, t, a);
		} catch (n) {
			As(e, t, n);
		}
	}
	function Os(e, t, n) {
		typeof n == "object" && n && typeof n.then == "function" ? n.then(function(n) {
			ks(e, t, n);
		}, function(n) {
			return As(e, t, n);
		}) : ks(e, t, n);
	}
	function ks(e, t, n) {
		t.status = "fulfilled", t.value = n, js(t), e.state = n, t = e.pending, t !== null && (n = t.next, n === t ? e.pending = null : (n = n.next, t.next = n, Ds(e, n)));
	}
	function As(e, t, n) {
		var r = e.pending;
		if (e.pending = null, r !== null) {
			r = r.next;
			do
				t.status = "rejected", t.reason = n, js(t), t = t.next;
			while (t !== r);
		}
		e.action = null;
	}
	function js(e) {
		e = e.listeners;
		for (var t = 0; t < e.length; t++) (0, e[t])();
	}
	function Ms(e, t) {
		return t;
	}
	function Ns(e, t) {
		if (P) {
			var n = G.formState;
			if (n !== null) {
				a: {
					var r = I;
					if (P) {
						if (N) {
							b: {
								for (var i = N, a = fa; i.nodeType !== 8;) if (!a || (i = lm(i.nextSibling), i === null)) {
									i = null;
									break b;
								}
								a = i.data, i = a === "F!" || a === "F" ? i : null;
							}
							if (i) {
								N = lm(i.nextSibling), r = i.data === "F!";
								break a;
							}
						}
						ma(r);
					}
					r = !1;
				}
				r && (t = n[0]);
			}
		}
		return n = ls(), n.memoizedState = n.baseState = t, r = {
			pending: null,
			lanes: 0,
			dispatch: null,
			lastRenderedReducer: Ms,
			lastRenderedState: t
		}, n.queue = r, n = dc.bind(null, I, r), r.dispatch = n, r = ws(!1), a = pc.bind(null, I, !1, r.queue), r = ls(), i = {
			state: t,
			dispatch: null,
			action: e,
			pending: null
		}, r.queue = i, n = Es.bind(null, I, i, a, n), i.dispatch = n, r.memoizedState = e, [
			t,
			n,
			!1
		];
	}
	function Ps(e) {
		return Fs(B(), L, e);
	}
	function Fs(e, t, n) {
		if (t = gs(e, t, Ms)[0], e = hs(ms)[0], typeof t == "object" && t && typeof t.then == "function") try {
			var r = ds(t);
		} catch (e) {
			throw e === to ? ro : e;
		}
		else r = t;
		t = B();
		var i = t.queue, a = i.dispatch;
		return n !== t.memoizedState && (I.flags |= 2048, Rs(9, { destroy: void 0 }, Is.bind(null, i, n), null)), [
			r,
			a,
			e
		];
	}
	function Is(e, t) {
		e.action = t;
	}
	function Ls(e) {
		var t = B(), n = L;
		if (n !== null) return Fs(t, n, e);
		B(), t = t.memoizedState, n = B();
		var r = n.queue.dispatch;
		return n.memoizedState = e, [
			t,
			r,
			!1
		];
	}
	function Rs(e, t, n, r) {
		return e = {
			tag: e,
			create: n,
			deps: r,
			inst: t,
			next: null
		}, t = I.updateQueue, t === null && (t = us(), I.updateQueue = t), n = t.lastEffect, n === null ? t.lastEffect = e.next = e : (r = n.next, n.next = e, e.next = r, t.lastEffect = e), e;
	}
	function zs() {
		return B().memoizedState;
	}
	function Bs(e, t, n, r) {
		var i = ls();
		I.flags |= e, i.memoizedState = Rs(1 | t, { destroy: void 0 }, n, r === void 0 ? null : r);
	}
	function Vs(e, t, n, r) {
		var i = B();
		r = r === void 0 ? null : r;
		var a = i.memoizedState.inst;
		L !== null && r !== null && ts(r, L.memoizedState.deps) ? i.memoizedState = Rs(t, a, n, r) : (I.flags |= e, i.memoizedState = Rs(1 | t, a, n, r));
	}
	function Hs(e, t) {
		Bs(8390656, 8, e, t);
	}
	function Us(e, t) {
		Vs(2048, 8, e, t);
	}
	function Ws(e) {
		I.flags |= 4;
		var t = I.updateQueue;
		if (t === null) t = us(), I.updateQueue = t, t.events = [e];
		else {
			var n = t.events;
			n === null ? t.events = [e] : n.push(e);
		}
	}
	function Gs(e) {
		var t = B().memoizedState;
		return Ws({
			ref: t,
			nextImpl: e
		}), function() {
			if (W & 2) throw Error(o(440));
			return t.impl.apply(void 0, arguments);
		};
	}
	function Ks(e, t) {
		return Vs(4, 2, e, t);
	}
	function qs(e, t) {
		return Vs(4, 4, e, t);
	}
	function Js(e, t) {
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
	function Ys(e, t, n) {
		n = n == null ? null : n.concat([e]), Vs(4, 4, Js.bind(null, t, e), n);
	}
	function Xs() {}
	function Zs(e, t) {
		var n = B();
		t = t === void 0 ? null : t;
		var r = n.memoizedState;
		return t !== null && ts(t, r[1]) ? r[0] : (n.memoizedState = [e, t], e);
	}
	function Qs(e, t) {
		var n = B();
		t = t === void 0 ? null : t;
		var r = n.memoizedState;
		if (t !== null && ts(t, r[1])) return r[0];
		if (r = e(), Xo) {
			it(!0);
			try {
				e();
			} finally {
				it(!1);
			}
		}
		return n.memoizedState = [r, t], r;
	}
	function $s(e, t, n) {
		return n === void 0 || qo & 1073741824 && !(q & 261930) ? e.memoizedState = t : (e.memoizedState = n, e = Md(), I.lanes |= e, sd |= e, n);
	}
	function ec(e, t, n, r) {
		return Jr(n, t) ? n : Mo.current === null ? !(qo & 106) || qo & 1073741824 && !(q & 261930) ? (Ic = !0, e.memoizedState = n) : (e = Md(), I.lanes |= e, sd |= e, t) : (e = $s(e, n, r), Jr(e, t) || (Ic = !0), e);
	}
	function tc(e, t, n, r, i) {
		var a = k.p;
		k.p = a !== 0 && 8 > a ? a : 8;
		var o = O.T, s = {};
		s.types = o === null ? null : o.types, O.T = s, pc(e, !1, t, n);
		try {
			var c = i(), l = O.S;
			l !== null && l(s, c), typeof c == "object" && c && typeof c.then == "function" ? fc(e, t, Ya(c, r), jd(e)) : fc(e, t, r, jd(e));
		} catch (n) {
			fc(e, t, {
				then: function() {},
				status: "rejected",
				reason: n
			}, jd());
		} finally {
			k.p = a, o !== null && s.types !== null && (o.types = s.types), O.T = o;
		}
	}
	function nc() {}
	function rc(e, t, n, r) {
		if (e.tag !== 5) throw Error(o(476));
		var i = ic(e).queue;
		tc(e, i, t, Se, n === null ? nc : function() {
			return ac(e), n(r);
		});
	}
	function ic(e) {
		var t = e.memoizedState;
		if (t !== null) return t;
		t = {
			memoizedState: Se,
			baseState: Se,
			baseQueue: null,
			queue: {
				pending: null,
				lanes: 0,
				dispatch: null,
				lastRenderedReducer: ms,
				lastRenderedState: Se
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
				lastRenderedReducer: ms,
				lastRenderedState: n
			},
			next: null
		}, e.memoizedState = t, e = e.alternate, e !== null && (e.memoizedState = t), t;
	}
	function ac(e) {
		var t = ic(e);
		t.next === null && (t = e.alternate.memoizedState), fc(e, t.next.queue, {}, jd());
	}
	function oc() {
		return ja(sh);
	}
	function sc() {
		return B().memoizedState;
	}
	function cc() {
		return B().memoizedState;
	}
	function lc(e) {
		for (var t = e.return; t !== null;) {
			switch (t.tag) {
				case 24:
				case 3:
					var n = jd();
					e = Co(n);
					var r = wo(t, e, n);
					r !== null && (Pd(r, t, n), To(r, t, n)), t = { cache: Ra() }, e.payload = t;
					return;
			}
			t = t.return;
		}
	}
	function uc(e, t, n) {
		var r = jd();
		n = {
			lane: r,
			revertLane: 0,
			gesture: null,
			action: n,
			hasEagerState: !1,
			eagerState: null,
			next: null
		}, mc(e) ? hc(t, n) : (n = Ni(e, t, n, r), n !== null && (Pd(n, e, r), gc(n, t, r)));
	}
	function dc(e, t, n) {
		fc(e, t, n, jd());
	}
	function fc(e, t, n, r) {
		var i = {
			lane: r,
			revertLane: 0,
			gesture: null,
			action: n,
			hasEagerState: !1,
			eagerState: null,
			next: null
		};
		if (mc(e)) hc(t, i);
		else {
			var a = e.alternate;
			if (e.lanes === 0 && (a === null || a.lanes === 0) && (a = t.lastRenderedReducer, a !== null)) try {
				var o = t.lastRenderedState, s = a(o, n);
				if (i.hasEagerState = !0, i.eagerState = s, Jr(s, o)) return Mi(e, t, i, 0), G === null && ji(), !1;
			} catch {}
			if (n = Ni(e, t, i, r), n !== null) return Pd(n, e, r), gc(n, t, r), !0;
		}
		return !1;
	}
	function pc(e, t, n, r) {
		if (r = {
			lane: 2,
			revertLane: Pf(),
			gesture: null,
			action: r,
			hasEagerState: !1,
			eagerState: null,
			next: null
		}, mc(e)) {
			if (t) throw Error(o(479));
		} else t = Ni(e, n, r, 2), t !== null && Pd(t, e, 2);
	}
	function mc(e) {
		var t = e.alternate;
		return e === I || t !== null && t === I;
	}
	function hc(e, t) {
		Yo = Jo = !0;
		var n = e.pending;
		n === null ? t.next = t : (t.next = n.next, n.next = t), e.pending = t;
	}
	function gc(e, t, n) {
		if (n & 4194048) {
			var r = t.lanes;
			r &= e.pendingLanes, n |= r, t.lanes = n, St(e, n);
		}
	}
	var _c = {
		readContext: ja,
		use: fs,
		useCallback: z,
		useContext: z,
		useEffect: z,
		useImperativeHandle: z,
		useLayoutEffect: z,
		useInsertionEffect: z,
		useMemo: z,
		useReducer: z,
		useRef: z,
		useState: z,
		useDebugValue: z,
		useDeferredValue: z,
		useTransition: z,
		useSyncExternalStore: z,
		useId: z,
		useHostTransitionStatus: z,
		useFormState: z,
		useActionState: z,
		useOptimistic: z,
		useMemoCache: z,
		useCacheRefresh: z,
		useEffectEvent: z
	}, vc = {
		readContext: ja,
		use: fs,
		useCallback: function(e, t) {
			return ls().memoizedState = [e, t === void 0 ? null : t], e;
		},
		useContext: ja,
		useEffect: Hs,
		useImperativeHandle: function(e, t, n) {
			n = n == null ? null : n.concat([e]), Bs(4194308, 4, Js.bind(null, t, e), n);
		},
		useLayoutEffect: function(e, t) {
			return Bs(4194308, 4, e, t);
		},
		useInsertionEffect: function(e, t) {
			Bs(4, 2, e, t);
		},
		useMemo: function(e, t) {
			var n = ls();
			t = t === void 0 ? null : t;
			var r = e();
			if (Xo) {
				it(!0);
				try {
					e();
				} finally {
					it(!1);
				}
			}
			return n.memoizedState = [r, t], r;
		},
		useReducer: function(e, t, n) {
			var r = ls();
			if (n !== void 0) {
				var i = n(t);
				if (Xo) {
					it(!0);
					try {
						n(t);
					} finally {
						it(!1);
					}
				}
			} else i = t;
			return r.memoizedState = r.baseState = i, e = {
				pending: null,
				lanes: 0,
				dispatch: null,
				lastRenderedReducer: e,
				lastRenderedState: i
			}, r.queue = e, e = e.dispatch = uc.bind(null, I, e), [r.memoizedState, e];
		},
		useRef: function(e) {
			var t = ls();
			return e = { current: e }, t.memoizedState = e;
		},
		useState: function(e) {
			e = ws(e);
			var t = e.queue, n = dc.bind(null, I, t);
			return t.dispatch = n, [e.memoizedState, n];
		},
		useDebugValue: Xs,
		useDeferredValue: function(e, t) {
			return $s(ls(), e, t);
		},
		useTransition: function() {
			var e = ws(!1);
			return e = tc.bind(null, I, e.queue, !0, !1), ls().memoizedState = e, [!1, e];
		},
		useSyncExternalStore: function(e, t, n) {
			var r = I, i = ls();
			if (P) {
				if (n === void 0) throw Error(o(407));
				n = n();
			} else {
				if (n = t(), G === null) throw Error(o(349));
				q & 127 || ys(r, t, n);
			}
			i.memoizedState = n;
			var a = {
				value: n,
				getSnapshot: t
			};
			return i.queue = a, Hs(xs.bind(null, r, a, e), [e]), r.flags |= 2048, Rs(9, { destroy: void 0 }, bs.bind(null, r, a, n, t), null), n;
		},
		useId: function() {
			var e = ls(), t = G.identifierPrefix;
			if (P) {
				var n = ia, r = ra;
				n = (r & ~(1 << 32 - at(r) - 1)).toString(32) + n, t = "_" + t + "R_" + n, n = Zo++, 0 < n && (t += "H" + n.toString(32)), t += "_";
			} else n = es++, t = "_" + t + "r_" + n.toString(32) + "_";
			return e.memoizedState = t;
		},
		useHostTransitionStatus: oc,
		useFormState: Ns,
		useActionState: Ns,
		useOptimistic: function(e) {
			var t = ls();
			t.memoizedState = t.baseState = e;
			var n = {
				pending: null,
				lanes: 0,
				dispatch: null,
				lastRenderedReducer: null,
				lastRenderedState: null
			};
			return t.queue = n, t = pc.bind(null, I, !0, n), n.dispatch = t, [e, t];
		},
		useMemoCache: ps,
		useCacheRefresh: function() {
			return ls().memoizedState = lc.bind(null, I);
		},
		useEffectEvent: function(e) {
			var t = ls(), n = { impl: e };
			return t.memoizedState = n, function() {
				if (W & 2) throw Error(o(440));
				return n.impl.apply(void 0, arguments);
			};
		}
	}, yc = {
		readContext: ja,
		use: fs,
		useCallback: Zs,
		useContext: ja,
		useEffect: Us,
		useImperativeHandle: Ys,
		useInsertionEffect: Ks,
		useLayoutEffect: qs,
		useMemo: Qs,
		useReducer: hs,
		useRef: zs,
		useState: function() {
			return hs(ms);
		},
		useDebugValue: Xs,
		useDeferredValue: function(e, t) {
			return ec(B(), L.memoizedState, e, t);
		},
		useTransition: function() {
			var e = hs(ms)[0], t = B().memoizedState;
			return [typeof e == "boolean" ? e : ds(e), t];
		},
		useSyncExternalStore: vs,
		useId: sc,
		useHostTransitionStatus: oc,
		useFormState: Ps,
		useActionState: Ps,
		useOptimistic: function(e, t) {
			return Ts(B(), L, e, t);
		},
		useMemoCache: ps,
		useCacheRefresh: cc,
		useEffectEvent: Gs
	}, bc = {
		readContext: ja,
		use: fs,
		useCallback: Zs,
		useContext: ja,
		useEffect: Us,
		useImperativeHandle: Ys,
		useInsertionEffect: Ks,
		useLayoutEffect: qs,
		useMemo: Qs,
		useReducer: _s,
		useRef: zs,
		useState: function() {
			return _s(ms);
		},
		useDebugValue: Xs,
		useDeferredValue: function(e, t) {
			var n = B();
			return L === null ? $s(n, e, t) : ec(n, L.memoizedState, e, t);
		},
		useTransition: function() {
			var e = _s(ms)[0], t = B().memoizedState;
			return [typeof e == "boolean" ? e : ds(e), t];
		},
		useSyncExternalStore: vs,
		useId: sc,
		useHostTransitionStatus: oc,
		useFormState: Ls,
		useActionState: Ls,
		useOptimistic: function(e, t) {
			var n = B();
			return L === null ? (n.baseState = e, [e, n.queue.dispatch]) : Ts(n, L, e, t);
		},
		useMemoCache: ps,
		useCacheRefresh: cc,
		useEffectEvent: Gs
	};
	function xc(e, t, n, r) {
		t = e.memoizedState, n = n(r, t), n = n == null ? t : S({}, t, n), e.memoizedState = n, e.lanes === 0 && (e.updateQueue.baseState = n);
	}
	var Sc = {
		enqueueSetState: function(e, t, n) {
			e = e._reactInternals;
			var r = jd(), i = Co(r);
			i.payload = t, n != null && (i.callback = n), t = wo(e, i, r), t !== null && (Pd(t, e, r), To(t, e, r));
		},
		enqueueReplaceState: function(e, t, n) {
			e = e._reactInternals;
			var r = jd(), i = Co(r);
			i.tag = 1, i.payload = t, n != null && (i.callback = n), t = wo(e, i, r), t !== null && (Pd(t, e, r), To(t, e, r));
		},
		enqueueForceUpdate: function(e, t) {
			e = e._reactInternals;
			var n = jd(), r = Co(n);
			r.tag = 2, t != null && (r.callback = t), t = wo(e, r, n), t !== null && (Pd(t, e, n), To(t, e, n));
		}
	};
	function Cc(e, t, n, r, i, a, o) {
		return e = e.stateNode, typeof e.shouldComponentUpdate == "function" ? e.shouldComponentUpdate(r, a, o) : t.prototype && t.prototype.isPureReactComponent ? !Yr(n, r) || !Yr(i, a) : !0;
	}
	function wc(e, t, n, r) {
		e = t.state, typeof t.componentWillReceiveProps == "function" && t.componentWillReceiveProps(n, r), typeof t.UNSAFE_componentWillReceiveProps == "function" && t.UNSAFE_componentWillReceiveProps(n, r), t.state !== e && Sc.enqueueReplaceState(t, t.state, null);
	}
	function Tc(e, t) {
		var n = t;
		if ("ref" in t) for (var r in n = {}, t) r !== "ref" && (n[r] = t[r]);
		if (e = e.defaultProps) for (var i in n === t && (n = S({}, n)), e) n[i] === void 0 && (n[i] = e[i]);
		return n;
	}
	function Ec(e) {
		Di(e);
	}
	function Dc(e) {
		console.error(e);
	}
	function Oc(e) {
		Di(e);
	}
	function kc(e, t) {
		try {
			var n = e.onUncaughtError;
			n(t.value, { componentStack: t.stack });
		} catch (e) {
			setTimeout(function() {
				throw e;
			});
		}
	}
	function Ac(e, t, n) {
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
	function jc(e, t, n) {
		return n = Co(n), n.tag = 3, n.payload = { element: null }, n.callback = function() {
			kc(e, t);
		}, n;
	}
	function Mc(e) {
		return e = Co(e), e.tag = 3, e;
	}
	function Nc(e, t, n, r) {
		var i = n.type.getDerivedStateFromError;
		if (typeof i == "function") {
			var a = r.value;
			e.payload = function() {
				return i(a);
			}, e.callback = function() {
				Ac(t, n, r);
			};
		}
		var o = n.stateNode;
		o !== null && typeof o.componentDidCatch == "function" && (e.callback = function() {
			Ac(t, n, r), typeof i != "function" && (yd === null ? yd = /* @__PURE__ */ new Set([this]) : yd.add(this));
			var e = r.stack;
			this.componentDidCatch(r.value, { componentStack: e === null ? "" : e });
		});
	}
	function Pc(e, t, n, r, i) {
		if (n.flags |= 32768, typeof r == "object" && r && typeof r.then == "function") {
			if (t = n.alternate, t !== null && Oa(t, n, i, !0), n = Lo.current, n !== null) {
				switch (n.tag) {
					case 31:
					case 13:
					case 19: return Ro === null ? Kd() : n.alternate === null && Y === 0 && (Y = 3), n.flags &= -257, n.flags |= 65536, n.lanes = i, r === io ? n.flags |= 16384 : (t = n.updateQueue, t === null ? n.updateQueue = /* @__PURE__ */ new Set([r]) : t.add(r), mf(e, r, i)), !1;
					case 22: return n.flags |= 65536, r === io ? n.flags |= 16384 : (t = n.updateQueue, t === null ? (t = {
						transitions: null,
						markerInstances: null,
						retryQueue: /* @__PURE__ */ new Set([r])
					}, n.updateQueue = t) : (n = t.retryQueue, n === null ? t.retryQueue = /* @__PURE__ */ new Set([r]) : n.add(r)), mf(e, r, i)), !1;
				}
				throw Error(o(435, n.tag));
			}
			return mf(e, r, i), Kd(), !1;
		}
		if (P) return t = Lo.current, t === null ? (r !== pa && (t = Error(o(423), { cause: r }), ba(Yi(t, n))), e = e.current.alternate, e.flags |= 65536, i &= -i, e.lanes |= i, r = Yi(r, n), i = jc(e.stateNode, r, i), Eo(e, i), Y !== 4 && (Y = 2)) : (!(t.flags & 65536) && (t.flags |= 256), t.flags |= 65536, t.lanes = i, r !== pa && (e = Error(o(422), { cause: r }), ba(Yi(e, n)))), !1;
		var a = Error(o(520), { cause: r });
		if (a = Yi(a, n), fd === null ? fd = [a] : fd.push(a), Y !== 4 && (Y = 2), t === null) return !0;
		r = Yi(r, n), n = t;
		do {
			switch (n.tag) {
				case 3: return n.flags |= 65536, e = i & -i, n.lanes |= e, e = jc(n.stateNode, r, e), Eo(n, e), !1;
				case 1:
					if (t = n.type, a = n.stateNode, !(n.flags & 128) && (typeof t.getDerivedStateFromError == "function" || a !== null && typeof a.componentDidCatch == "function" && (yd === null || !yd.has(a)))) return n.flags |= 65536, i &= -i, n.lanes |= i, i = Mc(i), Nc(i, e, n, r), Eo(n, i), !1;
					break;
				case 22: if (n.memoizedState !== null) return n.flags |= 65536, !1;
			}
			n = n.return;
		} while (n !== null);
		return !1;
	}
	var Fc = Error(o(461)), Ic = !1;
	function Lc(e, t, n, r) {
		t.child = e === null ? yo(t, null, n, r) : vo(t, e.child, n, r);
	}
	function Rc(e, t, n, r, i) {
		n = n.render;
		var a = t.ref;
		if ("ref" in r) {
			var o = {};
			for (var s in r) s !== "ref" && (o[s] = r[s]);
		} else o = r;
		return Aa(t), r = ns(e, t, n, o, a, i), s = os(), e !== null && !Ic ? (ss(e, t, i), fl(e, t, i)) : (P && s && sa(t), t.flags |= 1, Lc(e, t, r, i), t.child);
	}
	function zc(e, t, n, r, i) {
		if (e === null) {
			var a = n.type;
			return typeof a == "function" && !Bi(a) && a.defaultProps === void 0 && n.compare === null ? (t.tag = 15, t.type = a, Bc(e, t, a, r, i)) : (e = Ui(n.type, null, r, t, t.mode, i), e.ref = t.ref, e.return = t, t.child = e);
		}
		if (a = e.child, !pl(e, i)) {
			var o = a.memoizedProps;
			if (n = n.compare, n = n === null ? Yr : n, n(o, r) && e.ref === t.ref) return fl(e, t, i);
		}
		return t.flags |= 1, e = Vi(a, r), e.ref = t.ref, e.return = t, t.child = e;
	}
	function Bc(e, t, n, r, i) {
		if (e !== null) {
			var a = e.memoizedProps;
			if (Yr(a, r) && e.ref === t.ref) {
				if (Ic = !1, t.pendingProps = r = a, pl(e, i)) e.flags & 131072 && (Ic = !0);
				else return t.lanes = e.lanes, fl(e, t, i);
			}
		}
		return Jc(e, t, n, r, i);
	}
	function Vc(e, t, n, r) {
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
				return Uc(e, t, a, n, r);
			}
			if (n & 536870912) t.memoizedState = {
				baseLanes: 0,
				cachePool: null
			}, e !== null && $a(t, a === null ? null : a.cachePool), a === null ? Fo() : Po(t, a), Vo(t);
			else return r = t.lanes = 536870912, Uc(e, t, a === null ? n : a.baseLanes | n, n, r);
		} else a === null ? (e !== null && $a(t, null), Fo(), Ho()) : ($a(t, a.cachePool), Po(t, a), Ho(), t.memoizedState = null);
		return Lc(e, t, i, n), t.child;
	}
	function Hc(e, t) {
		return e !== null && e.tag === 22 || t.stateNode !== null || (t.stateNode = {
			_visibility: 1,
			_pendingMarkers: null,
			_retryCache: null,
			_transitions: null
		}), t.sibling;
	}
	function Uc(e, t, n, r, i) {
		var a = Qa();
		return a = a === null ? null : {
			parent: La._currentValue,
			pool: a
		}, t.memoizedState = {
			baseLanes: n,
			cachePool: a
		}, e !== null && $a(t, null), Fo(), Vo(t), e !== null && Oa(e, t, r, !0), t.childLanes = i, null;
	}
	function Wc(e, t) {
		return t = rl({
			mode: t.mode,
			children: t.children
		}, e.mode), t.ref = e.ref, e.child = t, t.return = e, t;
	}
	function Gc(e, t, n) {
		return vo(t, e.child, null, n), e = Wc(t, t.pendingProps), e.flags |= 2, Uo(t), t.memoizedState = null, e;
	}
	function Kc(e, t, n) {
		var r = t.pendingProps, i = !!(t.flags & 128);
		if (t.flags &= -129, e === null) {
			if (P) {
				if (r.mode === "hidden") return e = Wc(t, r), t.lanes = 536870912, e.memoizedState = {
					baseLanes: 0,
					cachePool: null
				}, Hc(null, e);
				if (Bo(t), (e = N) ? (e = am(e, fa), e = e !== null && e.data === "&" ? e : null, e !== null && (t.memoizedState = {
					dehydrated: e,
					treeContext: na === null ? null : {
						id: ra,
						overflow: ia
					},
					retryLane: 536870912,
					hydrationErrors: null
				}, n = Ki(e), n.return = t, t.child = n, ua = t, N = null)) : e = null, e === null) throw ma(t);
				return t.lanes = 536870912, null;
			}
			return Wc(t, r);
		}
		var a = e.memoizedState;
		if (a !== null) {
			var s = a.dehydrated;
			if (Bo(t), i) {
				if (t.flags & 256) t.flags &= -257, t = Gc(e, t, n);
				else if (t.memoizedState !== null) t.child = e.child, t.flags |= 128, t = null;
				else throw Error(o(558));
			} else if (Ic || Oa(e, t, n, !1), i = (n & e.childLanes) !== 0, Ic || i) {
				if (Mo.current === null) {
					if (r = G, r !== null && (s = Ct(r, n), s !== 0 && s !== a.retryLane)) throw a.retryLane = s, Pi(e, s), Pd(r, e, s), Fc;
					Kd();
				}
				t = Gc(e, t, n);
			} else e = a.treeContext, N = lm(s.nextSibling), ua = t, P = !0, da = null, fa = !1, e !== null && la(t, e), t = Wc(t, r), t.flags |= 134221824;
			return t;
		}
		return e = Vi(e.child, {
			mode: r.mode,
			children: r.children
		}), e.ref = t.ref, t.child = e, e.return = t, e;
	}
	function qc(e, t) {
		var n = t.ref;
		if (n === null) e !== null && e.ref !== null && (t.flags |= 4194816);
		else {
			if (typeof n != "function" && typeof n != "object") throw Error(o(284));
			(e === null || e.ref !== n) && (t.flags |= 4194816);
		}
	}
	function Jc(e, t, n, r, i) {
		return Aa(t), n = ns(e, t, n, r, void 0, i), r = os(), e !== null && !Ic ? (ss(e, t, i), fl(e, t, i)) : (P && r && sa(t), t.flags |= 1, Lc(e, t, n, i), t.child);
	}
	function Yc(e, t, n, r, i, a) {
		return Aa(t), t.updateQueue = null, n = is(t, r, n, i), rs(e), r = os(), e !== null && !Ic ? (ss(e, t, a), fl(e, t, a)) : (P && r && sa(t), t.flags |= 1, Lc(e, t, n, a), t.child);
	}
	function Xc(e, t, n, r, i) {
		if (Aa(t), t.stateNode === null) {
			var a = Li, o = n.contextType;
			typeof o == "object" && o && (a = ja(o)), a = new n(r, a), t.memoizedState = a.state !== null && a.state !== void 0 ? a.state : null, a.updater = Sc, t.stateNode = a, a._reactInternals = t, a = t.stateNode, a.props = r, a.state = t.memoizedState, a.refs = {}, xo(t), o = n.contextType, a.context = typeof o == "object" && o ? ja(o) : Li, a.state = t.memoizedState, o = n.getDerivedStateFromProps, typeof o == "function" && (xc(t, n, o, r), a.state = t.memoizedState), typeof n.getDerivedStateFromProps == "function" || typeof a.getSnapshotBeforeUpdate == "function" || typeof a.UNSAFE_componentWillMount != "function" && typeof a.componentWillMount != "function" || (o = a.state, typeof a.componentWillMount == "function" && a.componentWillMount(), typeof a.UNSAFE_componentWillMount == "function" && a.UNSAFE_componentWillMount(), o !== a.state && Sc.enqueueReplaceState(a, a.state, null), ko(t, r, a, i), Oo(), a.state = t.memoizedState), typeof a.componentDidMount == "function" && (t.flags |= 4194308), r = !0;
		} else if (e === null) {
			a = t.stateNode;
			var s = t.memoizedProps, c = Tc(n, s);
			a.props = c;
			var l = a.context, u = n.contextType;
			o = Li, typeof u == "object" && u && (o = ja(u));
			var d = n.getDerivedStateFromProps;
			u = typeof d == "function" || typeof a.getSnapshotBeforeUpdate == "function", s = t.pendingProps !== s, u || typeof a.UNSAFE_componentWillReceiveProps != "function" && typeof a.componentWillReceiveProps != "function" || (s || l !== o) && wc(t, a, r, o), bo = !1;
			var f = t.memoizedState;
			a.state = f, ko(t, r, a, i), Oo(), l = t.memoizedState, s || f !== l || bo ? (typeof d == "function" && (xc(t, n, d, r), l = t.memoizedState), (c = bo || Cc(t, n, c, r, f, l, o)) ? (u || typeof a.UNSAFE_componentWillMount != "function" && typeof a.componentWillMount != "function" || (typeof a.componentWillMount == "function" && a.componentWillMount(), typeof a.UNSAFE_componentWillMount == "function" && a.UNSAFE_componentWillMount()), typeof a.componentDidMount == "function" && (t.flags |= 4194308)) : (typeof a.componentDidMount == "function" && (t.flags |= 4194308), t.memoizedProps = r, t.memoizedState = l), a.props = r, a.state = l, a.context = o, r = c) : (typeof a.componentDidMount == "function" && (t.flags |= 4194308), r = !1);
		} else {
			a = t.stateNode, So(e, t), o = t.memoizedProps, u = Tc(n, o), a.props = u, d = t.pendingProps, f = a.context, l = n.contextType, c = Li, typeof l == "object" && l && (c = ja(l)), s = n.getDerivedStateFromProps, (l = typeof s == "function" || typeof a.getSnapshotBeforeUpdate == "function") || typeof a.UNSAFE_componentWillReceiveProps != "function" && typeof a.componentWillReceiveProps != "function" || (o !== d || f !== c) && wc(t, a, r, c), bo = !1, f = t.memoizedState, a.state = f, ko(t, r, a, i), Oo();
			var p = t.memoizedState;
			o !== d || f !== p || bo || e !== null && e.dependencies !== null && ka(e.dependencies) ? (typeof s == "function" && (xc(t, n, s, r), p = t.memoizedState), (u = bo || Cc(t, n, u, r, f, p, c) || e !== null && e.dependencies !== null && ka(e.dependencies)) ? (l || typeof a.UNSAFE_componentWillUpdate != "function" && typeof a.componentWillUpdate != "function" || (typeof a.componentWillUpdate == "function" && a.componentWillUpdate(r, p, c), typeof a.UNSAFE_componentWillUpdate == "function" && a.UNSAFE_componentWillUpdate(r, p, c)), typeof a.componentDidUpdate == "function" && (t.flags |= 4), typeof a.getSnapshotBeforeUpdate == "function" && (t.flags |= 1024)) : (typeof a.componentDidUpdate != "function" || o === e.memoizedProps && f === e.memoizedState || (t.flags |= 4), typeof a.getSnapshotBeforeUpdate != "function" || o === e.memoizedProps && f === e.memoizedState || (t.flags |= 1024), t.memoizedProps = r, t.memoizedState = p), a.props = r, a.state = p, a.context = c, r = u) : (typeof a.componentDidUpdate != "function" || o === e.memoizedProps && f === e.memoizedState || (t.flags |= 4), typeof a.getSnapshotBeforeUpdate != "function" || o === e.memoizedProps && f === e.memoizedState || (t.flags |= 1024), r = !1);
		}
		return a = r, qc(e, t), r = !!(t.flags & 128), a || r ? (a = t.stateNode, n = r && typeof n.getDerivedStateFromError != "function" ? null : a.render(), t.flags |= 1, e !== null && r ? (t.child = vo(t, e.child, null, i), t.child = vo(t, null, n, i)) : Lc(e, t, n, i), t.memoizedState = a.state, e = t.child) : e = fl(e, t, i), e;
	}
	function Zc(e, t, n, r) {
		return va(), t.flags |= 256, Lc(e, t, n, r), t.child;
	}
	var Qc = {
		dehydrated: null,
		treeContext: null,
		retryLane: 0,
		hydrationErrors: null
	};
	function $c(e) {
		return {
			baseLanes: e,
			cachePool: eo()
		};
	}
	function el(e, t, n) {
		return e = e === null ? 0 : e.childLanes & ~n, t && (e |= ud), e;
	}
	function tl(e, t, n) {
		var r = t.pendingProps, i = !1, a = !!(t.flags & 128), o;
		if ((o = a) || (o = e !== null && e.memoizedState === null ? !1 : !!(Wo.current & 2)), o && (i = !0, t.flags &= -129), o = !!(t.flags & 32), t.flags &= -33, e === null) {
			if (P) {
				if (i ? zo(t) : Ho(), (e = N) ? (e = am(e, fa), e = e !== null && e.data !== "&" ? e : null, e !== null && (t.memoizedState = {
					dehydrated: e,
					treeContext: na === null ? null : {
						id: ra,
						overflow: ia
					},
					retryLane: 536870912,
					hydrationErrors: null
				}, n = Ki(e), n.return = t, t.child = n, ua = t, N = null)) : e = null, e === null) throw ma(t);
				return t.lanes = sm(e) ? 32 : 536870912, null;
			}
			return a = r.children, r = r.fallback, i ? (Ho(), i = t.mode, a = rl({
				mode: "hidden",
				children: a
			}, i), r = Wi(r, i, n, null), a.return = t, r.return = t, a.sibling = r, t.child = a, r = t.child, r.memoizedState = $c(n), r.childLanes = el(e, o, n), t.memoizedState = Qc, Hc(null, r)) : (zo(t), nl(t, a));
		}
		var s = e.memoizedState;
		if (s !== null) {
			var c = s.dehydrated;
			if (c !== null) return al(e, t, a, o, r, c, s, n);
		}
		return i ? (Ho(), i = r.fallback, a = t.mode, s = e.child, c = s.sibling, r = Vi(s, {
			mode: "hidden",
			children: r.children
		}), r.subtreeFlags = s.subtreeFlags & 1206910976, c === null ? (i = Wi(i, a, n, null), i.flags |= 2) : i = Vi(c, i), i.return = t, r.return = t, r.sibling = i, t.child = r, Hc(null, r), r = t.child, i = e.child.memoizedState, i === null ? i = $c(n) : (a = i.cachePool, a === null ? a = eo() : (s = La._currentValue, a = a.parent === s ? a : {
			parent: s,
			pool: s
		}), i = {
			baseLanes: i.baseLanes | n,
			cachePool: a
		}), r.memoizedState = i, r.childLanes = el(e, o, n), t.memoizedState = Qc, Hc(e.child, r)) : (zo(t), n = e.child, e = n.sibling, n = Vi(n, {
			mode: "visible",
			children: r.children
		}), n.return = t, n.sibling = null, e !== null && (o = t.deletions, o === null ? (t.deletions = [e], t.flags |= 16) : o.push(e)), t.child = n, t.memoizedState = null, n);
	}
	function nl(e, t) {
		return t = rl({
			mode: "visible",
			children: t
		}, e.mode), t.return = e, e.child = t;
	}
	function rl(e, t) {
		return e = zi(22, e, null, t), e.lanes = 0, e;
	}
	function il(e, t, n) {
		return vo(t, e.child, null, n), e = nl(t, t.pendingProps.children), e.flags |= 2, t.memoizedState = null, e;
	}
	function al(e, t, n, r, i, a, s, c) {
		if (n) return t.flags & 256 ? (zo(t), t.flags &= -257, il(e, t, c)) : t.memoizedState === null ? (Ho(), a = i.fallback, s = t.mode, i = rl({
			mode: "visible",
			children: i.children
		}, s), a = Wi(a, s, c, null), a.flags |= 2, i.return = t, a.return = t, i.sibling = a, t.child = i, vo(t, e.child, null, c), i = t.child, i.memoizedState = $c(c), i.childLanes = el(e, r, c), t.memoizedState = Qc, Hc(null, i)) : (Ho(), t.child = e.child, t.flags |= 128, null);
		if (zo(t), sm(a)) {
			if (r = a.nextSibling && a.nextSibling.dataset, r) var l = r.dgst;
			return r = l, r !== "" && (i = Error(o(419)), i.stack = "", i.digest = r, ba({
				value: i,
				source: null,
				stack: null
			})), il(e, t, c);
		}
		if (Ic || Oa(e, t, c, !1), r = (c & e.childLanes) !== 0, Ic || r) {
			if (Mo.current !== null) return il(e, t, c);
			if (r = G, r !== null && (i = Ct(r, c), i !== 0 && i !== s.retryLane)) throw s.retryLane = i, Pi(e, i), Pd(r, e, i), Fc;
			return om(a) || Kd(), il(e, t, c);
		}
		return om(a) ? (t.flags |= 192, t.child = e.child, null) : (e = s.treeContext, N = lm(a.nextSibling), ua = t, P = !0, da = null, fa = !1, e !== null && la(t, e), t = nl(t, i.children), t.flags |= 134221824, t);
	}
	function ol(e, t, n) {
		e.lanes |= t;
		var r = e.alternate;
		r !== null && (r.lanes |= t), Ea(e.return, t, n);
	}
	function sl(e) {
		for (var t = null; e !== null;) {
			var n = e.alternate;
			n !== null && Ko(n) === null && (t = e), e = e.sibling;
		}
		return t;
	}
	function cl(e, t, n, r, i, a) {
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
	function ll(e) {
		var t = e.child;
		for (e.child = null; t !== null;) {
			var n = t.sibling;
			t.sibling = e.child, e.child = t, t = n;
		}
	}
	function ul(e, t, n) {
		var r = t.pendingProps, i = r.revealOrder, a = r.tail;
		r = r.children;
		var o = Wo.current;
		if (t.flags & 128) return Go(t, o), null;
		var s = !!(o & 2);
		if (s ? (o = o & 1 | 2, t.flags |= 128) : o &= 1, Go(t, o), i === "backwards" && e !== null ? (ll(e), Lc(e, t, r, n), ll(e)) : Lc(e, t, r, n), r = P ? $i : 0, !s && e !== null && e.flags & 128) a: for (e = t.child; e !== null;) {
			if (e.tag === 13) e.memoizedState !== null && ol(e, n, t);
			else if (e.tag === 19) ol(e, n, t);
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
			case "backwards":
				n = sl(t.child), n === null ? (i = t.child, t.child = null) : (i = n.sibling, n.sibling = null, ll(t)), cl(t, !0, i, null, a, r);
				break;
			case "unstable_legacy-backwards":
				for (n = null, i = t.child, t.child = null; i !== null;) {
					if (e = i.alternate, e !== null && Ko(e) === null) {
						t.child = i;
						break;
					}
					e = i.sibling, i.sibling = n, n = i, i = e;
				}
				cl(t, !0, n, null, a, r);
				break;
			case "together":
				cl(t, !1, null, null, void 0, r);
				break;
			case "independent":
				t.memoizedState = null;
				break;
			default: n = sl(t.child), n === null ? (i = t.child, t.child = null) : (i = n.sibling, n.sibling = null), cl(t, !1, i, n, a, r);
		}
		return t.child;
	}
	function dl(e, t, n) {
		var r = t.pendingProps;
		return wa(t, t.type, r.value), Lc(e, t, r.children, n), t.child;
	}
	function fl(e, t, n) {
		if (e !== null && (t.dependencies = e.dependencies), sd |= t.lanes, (n & t.childLanes) === 0) {
			if (e !== null) {
				if (Oa(e, t, n, !1), (n & t.childLanes) === 0) return null;
			} else return null;
		}
		if (e !== null && t.child !== e.child) throw Error(o(153));
		if (t.child !== null) {
			for (e = t.child, n = Vi(e, e.pendingProps), t.child = n, n.return = t; e.sibling !== null;) e = e.sibling, n = n.sibling = Vi(e, e.pendingProps), n.return = t;
			n.sibling = null;
		}
		return t.child;
	}
	function pl(e, t) {
		return (e.lanes & t) !== 0 || (e = e.dependencies, !!(e !== null && ka(e)));
	}
	function ml(e, t, n) {
		switch (t.tag) {
			case 3:
				je(t, t.stateNode.containerInfo), wa(t, La, e.memoizedState.cache), va();
				break;
			case 27:
			case 5:
				Ne(t);
				break;
			case 4:
				je(t, t.stateNode.containerInfo);
				break;
			case 10:
				wa(t, t.type, t.memoizedProps.value);
				break;
			case 31:
				if (t.memoizedState !== null) return t.flags |= 128, Bo(t), null;
				break;
			case 13:
				var r = t.memoizedState;
				if (r !== null) {
					if (r.dehydrated !== null) return zo(t), t.flags |= 128, null;
					r = Oa(e, t, n, !1);
					var i = t.child.childLanes;
					return r || (n & i) !== 0 ? tl(e, t, n) : (zo(t), e = fl(e, t, n), e === null ? null : e.sibling);
				}
				zo(t);
				break;
			case 19:
				if (t.flags & 128) return ul(e, t, n);
				if (i = !!(e.flags & 128), r = (n & t.childLanes) !== 0, r ||= (Oa(e, t, n, !1), (n & t.childLanes) !== 0), i) {
					if (r) return ul(e, t, n);
					t.flags |= 128;
				}
				if (i = t.memoizedState, i !== null && (i.rendering = null, i.tail = null, i.lastEffect = null), Go(t, Wo.current), r) break;
				return null;
			case 22: return t.lanes = 0, Vc(e, t, n, t.pendingProps);
			case 24: wa(t, La, e.memoizedState.cache);
		}
		return fl(e, t, n);
	}
	function hl(e, t, n) {
		if (e !== null) {
			if (e.memoizedProps !== t.pendingProps) Ic = !0;
			else {
				if (!pl(e, n) && !(t.flags & 128)) return Ic = !1, ml(e, t, n);
				Ic = !!(e.flags & 131072);
			}
		} else Ic = !1, P && t.flags & 1048576 && oa(t, $i, t.index);
		switch (t.lanes = 0, t.tag) {
			case 16:
				a: {
					var r = t.pendingProps;
					if (e = so(t.elementType), t.type = e, typeof e == "function") Bi(e) ? (r = Tc(e, r), t.tag = 1, t = Xc(null, t, e, r, n)) : (t.tag = 0, t = Jc(null, t, e, r, n));
					else {
						if (e != null) {
							var i = e.$$typeof;
							if (i === C) {
								t.tag = 11, t = Rc(null, t, e, r, n);
								break a;
							}
							if (i === pe) {
								t.tag = 14, t = zc(null, t, e, r, n);
								break a;
							}
							if (i === de) {
								t.tag = 10, t.type = e, t = dl(null, t, n);
								break a;
							}
						}
						throw t = be(e) || e, Error(o(306, t, ""));
					}
				}
				return t;
			case 0: return Jc(e, t, t.type, t.pendingProps, n);
			case 1: return r = t.type, i = Tc(r, t.pendingProps), Xc(e, t, r, i, n);
			case 3:
				a: {
					if (je(t, t.stateNode.containerInfo), e === null) throw Error(o(387));
					r = t.pendingProps;
					var a = t.memoizedState;
					i = a.element, So(e, t), ko(t, r, null, n);
					var s = t.memoizedState;
					if (r = s.cache, wa(t, La, r), r !== a.cache && Da(t, [La], n, !0), Oo(), r = s.element, a.isDehydrated) {
						if (a = {
							element: r,
							isDehydrated: !1,
							cache: s.cache
						}, t.updateQueue.baseState = a, t.memoizedState = a, t.flags & 256) {
							t = Zc(e, t, r, n);
							break a;
						}
						if (r !== i) {
							i = Yi(Error(o(424)), t), ba(i), t = Zc(e, t, r, n);
							break a;
						}
						switch (e = t.stateNode.containerInfo, e.nodeType) {
							case 9:
								e = e.body;
								break;
							default: e = e.nodeName === "HTML" ? e.ownerDocument.body : e;
						}
						for (N = lm(e.firstChild), ua = t, P = !0, da = null, fa = !0, n = yo(t, null, r, n), t.child = n; n;) n.flags = n.flags & -3 | 134221824, n = n.sibling;
					} else {
						if (va(), r === i) {
							t = fl(e, t, n);
							break a;
						}
						Lc(e, t, r, n);
					}
					t = t.child;
				}
				return t;
			case 26: return qc(e, t), e === null ? (n = Nm(t.type, null, t.pendingProps, null)) ? t.memoizedState = n : P || (t.stateNode = fp(t.type, t.pendingProps, ke.current, t)) : t.memoizedState = Nm(t.type, e.memoizedProps, t.pendingProps, e.memoizedState), null;
			case 27: return Ne(t), e === null && P && (r = t.stateNode = hm(t.type, t.pendingProps, ke.current), ua = t, fa = !0, i = N, Sp(t.type) ? (um = i, N = lm(r.firstChild)) : N = i), Lc(e, t, t.pendingProps.children, n), qc(e, t), e === null && (t.flags |= 4194304), t.child;
			case 5: return e === null && P && ((i = r = N) && (r = rm(r, t.type, t.pendingProps, fa), r === null ? i = !1 : (t.stateNode = r, ua = t, N = lm(r.firstChild), fa = !1, i = !0)), i || ma(t)), Ne(t), i = t.type, a = t.pendingProps, s = e === null ? null : e.memoizedProps, r = a.children, pp(i, a) ? r = null : s !== null && pp(i, s) && (t.flags |= 32), t.memoizedState !== null && (i = ns(e, t, as, null, null, n), sh._currentValue = i), qc(e, t), Lc(e, t, r, n), t.child;
			case 6: return e === null && P && ((e = n = N) && (n = im(n, t.pendingProps, fa), n === null ? e = !1 : (t.stateNode = n, ua = t, N = null, e = !0)), e || ma(t)), null;
			case 13: return tl(e, t, n);
			case 4: return je(t, t.stateNode.containerInfo), r = t.pendingProps, e === null ? t.child = vo(t, null, r, n) : Lc(e, t, r, n), t.child;
			case 11: return Rc(e, t, t.type, t.pendingProps, n);
			case 7: return r = t.pendingProps, qc(e, t), Lc(e, t, r, n), t.child;
			case 8: return Lc(e, t, t.pendingProps.children, n), t.child;
			case 12: return Lc(e, t, t.pendingProps.children, n), t.child;
			case 10: return dl(e, t, n);
			case 9: return i = t.type._context, r = t.pendingProps.children, Aa(t), i = ja(i), r = r(i), t.flags |= 1, Lc(e, t, r, n), t.child;
			case 14: return zc(e, t, t.type, t.pendingProps, n);
			case 15: return Bc(e, t, t.type, t.pendingProps, n);
			case 19: return ul(e, t, n);
			case 31: return Kc(e, t, n);
			case 22: return Vc(e, t, n, t.pendingProps);
			case 24: return Aa(t), r = ja(La), e === null ? (i = Qa(), i === null && (i = G, a = Ra(), i.pooledCache = a, a.refCount++, a !== null && (i.pooledCacheLanes |= n), i = a), t.memoizedState = {
				parent: r,
				cache: i
			}, xo(t), wa(t, La, i)) : ((e.lanes & n) !== 0 && (So(e, t), ko(t, null, null, n), Oo()), i = e.memoizedState, a = t.memoizedState, i.parent === r ? (r = a.cache, wa(t, La, r), r !== i.cache && Da(t, [La], n, !0)) : (i = {
				parent: r,
				cache: r
			}, t.memoizedState = i, t.lanes === 0 && (t.memoizedState = t.updateQueue.baseState = i), wa(t, La, r))), Lc(e, t, t.pendingProps.children, n), t.child;
			case 30: return t.stateNode === null && (t.stateNode = {
				autoName: null,
				paired: null,
				clones: null,
				ref: null
			}), r = t.pendingProps, r.name != null && r.name !== "auto" ? t.flags |= e === null ? 18882560 : 18874368 : P && sa(t), e !== null && e.memoizedProps.name !== r.name ? t.flags |= 4194816 : qc(e, t), Lc(e, t, r.children, n), t.child;
			case 29: throw t.pendingProps;
		}
		throw Error(o(156, t.tag));
	}
	function gl(e) {
		e.flags |= 4;
	}
	function _l(e, t, n, r, i) {
		var a;
		if ((a = !!(e.mode & 32)) && (a = n === null ? Jm(t, r) : Jm(t, r) && (r.src !== n.src || r.srcSet !== n.srcSet)), a) {
			if (e.flags |= 16777216, (i & 335544128) === i) {
				if (e.stateNode.complete) e.flags |= 8192;
				else if (Ud()) e.flags |= 8192;
				else throw co = io, no;
			}
		} else e.flags &= -16777217;
	}
	function vl(e, t) {
		if (t.type !== "stylesheet" || t.state.loading & 4) e.flags &= -16777217;
		else if (e.flags |= 16777216, !Ym(t)) {
			if (Ud()) e.flags |= 8192;
			else throw co = io, no;
		}
	}
	function yl(e, t) {
		t !== null && (e.flags |= 4), e.flags & 16384 && (t = e.tag === 22 ? 536870912 : _t(), e.lanes |= t, dd |= t);
	}
	function bl(e, t) {
		if (!P) switch (e.tailMode) {
			case "visible": break;
			case "collapsed":
				for (var n = e.tail, r = null; n !== null;) n.alternate !== null && (r = n), n = n.sibling;
				r === null ? t || e.tail === null ? e.tail = null : e.tail.sibling = null : r.sibling = null;
				break;
			default:
				for (t = e.tail, n = null; t !== null;) t.alternate !== null && (n = t), t = t.sibling;
				n === null ? e.tail = null : n.sibling = null;
		}
	}
	function V(e) {
		var t = e.alternate !== null && e.alternate.child === e.child, n = 0, r = 0;
		if (t) for (var i = e.child; i !== null;) n |= i.lanes | i.childLanes, r |= i.subtreeFlags & 1206910976, r |= i.flags & 1206910976, i.return = e, i = i.sibling;
		else for (i = e.child; i !== null;) n |= i.lanes | i.childLanes, r |= i.subtreeFlags, r |= i.flags, i.return = e, i = i.sibling;
		return e.subtreeFlags |= r, e.childLanes = n, t;
	}
	function xl(e, t, n) {
		var r = t.pendingProps;
		switch (ca(t), t.tag) {
			case 16:
			case 15:
			case 0:
			case 11:
			case 7:
			case 8:
			case 12:
			case 9:
			case 14: return V(t), null;
			case 1: return V(t), null;
			case 3: return n = t.stateNode, r = null, e !== null && (r = e.memoizedState.cache), t.memoizedState.cache !== r && (t.flags |= 2048), Ta(La), Me(), n.pendingContext && (n.context = n.pendingContext, n.pendingContext = null), (e === null || e.child === null) && (_a(t) ? gl(t) : e === null || e.memoizedState.isDehydrated && !(t.flags & 256) || (t.flags |= 1024, ya())), V(t), null;
			case 26:
				var i = t.type, a = t.memoizedState;
				return e === null ? (gl(t), a === null ? (V(t), _l(t, i, null, r, n)) : (V(t), vl(t, a))) : a ? a === e.memoizedState ? (V(t), t.flags &= -16777217) : (gl(t), V(t), vl(t, a)) : (e = e.memoizedProps, e !== r && gl(t), V(t), _l(t, i, e, r, n)), null;
			case 27:
				if (Pe(t), n = ke.current, i = t.type, e !== null && t.stateNode != null) e.memoizedProps !== r && gl(t);
				else {
					if (!r) {
						if (t.stateNode === null) throw Error(o(166));
						return V(t), t.subtreeFlags &= -33554433, null;
					}
					e = De.current, _a(t) ? ha(t, e) : (e = hm(i, r, n), t.stateNode = e, gl(t));
				}
				return V(t), t.subtreeFlags &= -33554433, null;
			case 5:
				if (Pe(t), i = t.type, e !== null && t.stateNode != null) e.memoizedProps !== r && gl(t);
				else {
					if (!r) {
						if (t.stateNode === null) throw Error(o(166));
						return V(t), t.subtreeFlags &= -33554433, null;
					}
					if (a = De.current, _a(t)) ha(t, a);
					else {
						var s = lp(ke.current);
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
						a[kt] = t, a[At] = r;
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
						a: switch (np(a, i, r), i) {
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
						r && gl(t);
					}
				}
				return V(t), t.subtreeFlags &= -33554433, _l(t, t.type, e === null ? null : e.memoizedProps, t.pendingProps, n), null;
			case 6:
				if (e && t.stateNode != null) e.memoizedProps !== r && gl(t);
				else {
					if (typeof r != "string" && t.stateNode === null) throw Error(o(166));
					if (e = ke.current, _a(t)) {
						if (e = t.stateNode, n = t.memoizedProps, r = null, i = ua, i !== null) switch (i.tag) {
							case 27:
							case 5: r = i.memoizedProps;
						}
						e[kt] = t, e = !!(e.nodeValue === n || r !== null && !0 === r.suppressHydrationWarning || ep(e.nodeValue, n)), e || ma(t, !0);
					} else e = lp(e).createTextNode(r), e[kt] = t, t.stateNode = e;
				}
				return V(t), null;
			case 31:
				if (n = t.memoizedState, e === null || e.memoizedState !== null) {
					if (r = _a(t), n !== null) {
						if (e === null) {
							if (!r) throw Error(o(318));
							if (e = t.memoizedState, e = e === null ? null : e.dehydrated, !e) throw Error(o(557));
							e[kt] = t;
						} else va(), !(t.flags & 128) && (t.memoizedState = null), t.flags |= 4;
						V(t), e = !1;
					} else n = ya(), e !== null && e.memoizedState !== null && (e.memoizedState.hydrationErrors = n), e = !0;
					if (!e) return t.flags & 256 ? (Uo(t), t) : (Uo(t), null);
					if (t.flags & 128) throw Error(o(558));
				}
				return V(t), null;
			case 13:
				if (r = t.memoizedState, e === null || e.memoizedState !== null && e.memoizedState.dehydrated !== null) {
					if (i = _a(t), r !== null && r.dehydrated !== null) {
						if (e === null) {
							if (!i) throw Error(o(318));
							if (i = t.memoizedState, i = i === null ? null : i.dehydrated, !i) throw Error(o(317));
							i[kt] = t;
						} else va(), !(t.flags & 128) && (t.memoizedState = null), t.flags |= 4;
						V(t), i = !1;
					} else i = ya(), e !== null && e.memoizedState !== null && (e.memoizedState.hydrationErrors = i), i = !0;
					if (!i) return t.flags & 256 ? (Uo(t), t) : (Uo(t), null);
				}
				return Uo(t), t.flags & 128 ? (t.lanes = n, t) : (n = r !== null, e = e !== null && e.memoizedState !== null, n && (r = t.child, i = null, r.alternate !== null && r.alternate.memoizedState !== null && r.alternate.memoizedState.cachePool !== null && (i = r.alternate.memoizedState.cachePool.pool), a = null, r.memoizedState !== null && r.memoizedState.cachePool !== null && (a = r.memoizedState.cachePool.pool), a !== i && (r.flags |= 2048)), n !== e && n && (t.child.flags |= 8192), yl(t, t.updateQueue), V(t), null);
			case 4: return Me(), e === null && Wf(t.stateNode.containerInfo), t.flags |= 67108864, V(t), null;
			case 10: return Ta(t.type), V(t), null;
			case 19:
				if (F(t), r = t.memoizedState, r === null) return V(t), null;
				if (i = !!(t.flags & 128), a = r.rendering, a === null) {
					if (i) bl(r, !1);
					else {
						if (Y !== 0 || e !== null && e.flags & 128) for (e = t.child; e !== null;) {
							if (a = Ko(e), a !== null) {
								for (t.flags |= 128, bl(r, !1), e = a.updateQueue, t.updateQueue = e, yl(t, e), t.subtreeFlags = 0, e = n, n = t.child; n !== null;) Hi(n, e), n = n.sibling;
								return Go(t, Wo.current & 1 | 2), P && aa(t, r.treeForkCount), t.child;
							}
							e = e.sibling;
						}
						r.tail !== null && qe() > _d && (t.flags |= 128, i = !0, bl(r, !1), t.lanes = 4194304);
					}
				} else {
					if (!i) {
						if (e = Ko(a), e !== null) {
							if (t.flags |= 128, i = !0, e = e.updateQueue, t.updateQueue = e, yl(t, e), bl(r, !0), r.tail === null && r.tailMode !== "collapsed" && r.tailMode !== "visible" && !a.alternate && !P) return V(t), null;
						} else 2 * qe() - r.renderingStartTime > _d && n !== 536870912 && (t.flags |= 128, i = !0, bl(r, !1), t.lanes = 4194304);
					}
					r.isBackwards ? (a.sibling = t.child, t.child = a) : (e = r.last, e === null ? t.child = a : e.sibling = a, r.last = a);
				}
				if (r.tail !== null) {
					e = r.tail;
					a: {
						for (n = e; n !== null;) {
							if (n.alternate !== null) {
								n = !1;
								break a;
							}
							n = n.sibling;
						}
						n = !0;
					}
					return r.rendering = e, r.tail = e.sibling, r.renderingStartTime = qe(), e.sibling = null, a = Wo.current, a = i ? a & 1 | 2 : a & 1, r.tailMode === "visible" || r.tailMode === "collapsed" || !n || P ? Go(t, a) : (n = a, A(Lo, t), A(Wo, n), Ro === null && (Ro = t)), P && aa(t, r.treeForkCount), e;
				}
				return V(t), null;
			case 22:
			case 23: return Uo(t), Io(), r = t.memoizedState !== null, e === null ? r && (t.flags |= 8192) : e.memoizedState !== null !== r && (t.flags |= 8192), r ? n & 536870912 && !(t.flags & 128) && (V(t), t.subtreeFlags & 6 && (t.flags |= 8192)) : V(t), n = t.updateQueue, n !== null && yl(t, n.retryQueue), n = null, e !== null && e.memoizedState !== null && e.memoizedState.cachePool !== null && (n = e.memoizedState.cachePool.pool), r = null, t.memoizedState !== null && t.memoizedState.cachePool !== null && (r = t.memoizedState.cachePool.pool), r !== n && (t.flags |= 2048), e !== null && Ee(Za), null;
			case 24: return n = null, e !== null && (n = e.memoizedState.cache), t.memoizedState.cache !== n && (t.flags |= 2048), Ta(La), V(t), null;
			case 25: return null;
			case 30: return t.flags |= 33554432, V(t), null;
		}
		throw Error(o(156, t.tag));
	}
	function Sl(e, t) {
		switch (ca(t), t.tag) {
			case 1: return e = t.flags, e & 65536 ? (t.flags = e & -65537 | 128, t) : null;
			case 3: return Ta(La), Me(), e = t.flags, e & 65536 && !(e & 128) ? (t.flags = e & -65537 | 128, t) : null;
			case 26:
			case 27:
			case 5: return Pe(t), null;
			case 31:
				if (t.memoizedState !== null) {
					if (Uo(t), t.alternate === null) throw Error(o(340));
					va();
				}
				return e = t.flags, e & 65536 ? (t.flags = e & -65537 | 128, t) : null;
			case 13:
				if (Uo(t), e = t.memoizedState, e !== null && e.dehydrated !== null) {
					if (t.alternate === null) throw Error(o(340));
					va();
				}
				return e = t.flags, e & 65536 ? (t.flags = e & -65537 | 128, t) : null;
			case 19: return F(t), e = t.flags, e & 65536 ? (t.flags = e & -65537 | 128, e = t.memoizedState, e !== null && (e.rendering = null, e.tail = null), t.flags |= 4, t) : null;
			case 4: return Me(), null;
			case 10: return Ta(t.type), null;
			case 22:
			case 23: return Uo(t), Io(), e !== null && Ee(Za), e = t.flags, e & 65536 ? (t.flags = e & -65537 | 128, t) : null;
			case 24: return Ta(La), null;
			case 25: return null;
			default: return null;
		}
	}
	function Cl(e, t) {
		switch (ca(t), t.tag) {
			case 3:
				Ta(La), Me();
				break;
			case 26:
			case 27:
			case 5:
				Pe(t);
				break;
			case 4:
				Me();
				break;
			case 31:
				t.memoizedState !== null && Uo(t);
				break;
			case 13:
				Uo(t);
				break;
			case 19:
				F(t);
				break;
			case 10:
				Ta(t.type);
				break;
			case 22:
			case 23:
				Uo(t), Io(), e !== null && Ee(Za);
				break;
			case 24: Ta(La);
		}
	}
	function wl(e, t) {
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
	function Tl(e, t, n) {
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
	function El(e) {
		var t = e.updateQueue;
		if (t !== null) {
			var n = e.stateNode;
			try {
				jo(t, n);
			} catch (t) {
				Z(e, e.return, t);
			}
		}
	}
	function Dl(e, t, n) {
		n.props = Tc(e.type, e.memoizedProps), n.state = e.memoizedState;
		try {
			n.componentWillUnmount();
		} catch (n) {
			Z(e, t, n);
		}
	}
	function Ol(e, t) {
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
						var i = e.stateNode, a = wi(e.memoizedProps, i);
						(i.ref === null || i.ref.name !== a) && (i.ref = Pp(a)), r = i.ref;
						break;
					case 7:
						if (e.stateNode === null) {
							var o = new Fp(e);
							m(e.child, !1, Qp, o, void 0, void 0), e.stateNode = o;
						}
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
	function kl(e, t) {
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
	function Al(e, t) {
		if ((e.tag === 5 || e.tag === 27 || e.tag === 6) && e.alternate === null && t !== null) for (var n = 0; n < t.length; n++) em(e.stateNode, t[n]);
	}
	function jl(e) {
		for (var t = e.return; t !== null && (Pl(t) && em(e.stateNode, t.stateNode), !Nl(t));) t = t.return;
	}
	function Ml(e) {
		for (var t = e.return; t !== null && (Pl(t) && tm(e.stateNode, t.stateNode), !Nl(t));) t = t.return;
	}
	function Nl(e) {
		return e.tag === 5 || e.tag === 3 || e.tag === 27;
	}
	function Pl(e) {
		return e && e.tag === 7 && e.stateNode !== null;
	}
	function Fl(e) {
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
	function Il(e, t, n) {
		try {
			var r = e.stateNode;
			ip(r, e.type, n, t), r[At] = t;
		} catch (t) {
			Z(e, e.return, t);
		}
	}
	function Ll(e) {
		return e.tag === 5 || e.tag === 3 || e.tag === 26 || e.tag === 27 && Sp(e.type) || e.tag === 4;
	}
	function Rl(e) {
		a: for (;;) {
			for (; e.sibling === null;) {
				if (e.return === null || Ll(e.return)) return null;
				e = e.return;
			}
			for (e.sibling.return = e.return, e = e.sibling; e.tag !== 5 && e.tag !== 6 && e.tag !== 18;) {
				if (e.tag === 27 && Sp(e.type) || e.flags & 2 || e.child === null || e.tag === 4) continue a;
				e.child.return = e, e = e.child;
			}
			if (!(e.flags & 2)) return e.stateNode;
		}
	}
	function zl(e, t, n, r) {
		var i = e.tag;
		if (i === 5 || i === 6) i = e.stateNode, t ? (n.nodeType === 9 ? n.body : n.nodeName === "HTML" ? n.ownerDocument.body : n).insertBefore(i, t) : (t = n.nodeType === 9 ? n.body : n.nodeName === "HTML" ? n.ownerDocument.body : n, t.appendChild(i), n = n._reactRootContainer, n != null || t.onclick !== null || (t.onclick = Tn)), Al(e, r), M = !0;
		else if (i !== 4 && (i === 27 && (Al(e, r), r = null, Sp(e.type) && (n = e.stateNode, t = null)), e = e.child, e !== null)) for (zl(e, t, n, r), e = e.sibling; e !== null;) zl(e, t, n, r), e = e.sibling;
	}
	function Bl(e, t, n, r) {
		var i = e.tag;
		if (i === 5 || i === 6) i = e.stateNode, t ? n.insertBefore(i, t) : n.appendChild(i), Al(e, r), M = !0;
		else if (i !== 4 && (i === 27 && (Al(e, r), r = null, Sp(e.type) && (n = e.stateNode)), e = e.child, e !== null)) for (Bl(e, t, n, r), e = e.sibling; e !== null;) Bl(e, t, n, r), e = e.sibling;
	}
	function Vl(e) {
		var t = e.stateNode, n = e.memoizedProps;
		try {
			for (var r = e.type, i = t.attributes; i.length;) t.removeAttributeNode(i[0]);
			np(t, r, n), t[kt] = e, t[At] = n;
		} catch (t) {
			Z(e, e.return, t);
		}
	}
	var Hl = !1, Ul = null;
	function Wl(e) {
		(e.tag === 30 || e.subtreeFlags & 33554432) && (Hl = !0);
	}
	var Gl = null;
	function Kl() {
		var e = Gl;
		return Gl = null, e;
	}
	var ql = 0;
	function Jl(e, t, n, r, i) {
		return ql = 0, Yl(e.child, t, n, r, i);
	}
	function Yl(e, t, n, r, i) {
		for (var a = !1; e !== null;) {
			if (e.tag === 5) {
				var o = e.stateNode;
				if (r !== null) {
					var s = Op(o);
					r.push(s), s.view && (a = !0);
				} else a || Op(o).view && (a = !0);
				Hl = !0, Tp(o, ql === 0 ? t : t + "_" + ql, n), ql++;
			} else e.tag === 22 && e.memoizedState !== null || e.tag === 30 && i || Yl(e.child, t, n, r, i) && (a = !0);
			e = e.sibling;
		}
		return a;
	}
	function Xl(e, t) {
		for (; e !== null;) e.tag === 5 ? Ep(e.stateNode, e.memoizedProps) : e.tag === 22 && e.memoizedState !== null || e.tag === 30 && t || Xl(e.child, t), e = e.sibling;
	}
	function Zl(e) {
		if (e.subtreeFlags & 18874368) for (e = e.child; e !== null;) {
			if ((e.tag !== 22 || e.memoizedState === null) && (Zl(e), e.tag === 30 && e.flags & 18874368 && e.stateNode.paired)) {
				var t = e.memoizedProps;
				if (t.name == null || t.name === "auto") throw Error(o(544));
				var n = t.name;
				t = Ei(t.default, t.share), t !== "none" && (Jl(e, n, t, null, !1) || Xl(e.child, !1));
			}
			e = e.sibling;
		}
	}
	function Ql(e, t) {
		if (e.tag === 30) {
			var n = e.stateNode, r = e.memoizedProps, i = wi(r, n), a = Ei(r.default, n.paired ? r.share : r.enter);
			a === "none" ? Zl(e) : Jl(e, i, a, null, !1) ? (Zl(e), n.paired || t || Nd(e, r.onEnter)) : Xl(e.child, !1);
		} else if (e.subtreeFlags & 33554432) for (e = e.child; e !== null;) Ql(e, t), e = e.sibling;
		else Zl(e);
	}
	function $l(e) {
		if (Ul !== null && Ul.size !== 0) {
			var t = Ul;
			if (e.subtreeFlags & 18874368) for (e = e.child; e !== null;) {
				if (e.tag !== 22 || e.memoizedState === null) {
					if (e.tag === 30 && e.flags & 18874368) {
						var n = e.memoizedProps, r = n.name;
						if (r != null && r !== "auto") {
							var i = t.get(r);
							if (i !== void 0) {
								var a = Ei(n.default, n.share);
								if (a !== "none" && (Jl(e, r, a, null, !1) ? (a = e.stateNode, i.paired = a, a.paired = i, Nd(e, n.onShare)) : Xl(e.child, !1)), t.delete(r), t.size === 0) break;
							}
						}
					}
					$l(e);
				}
				e = e.sibling;
			}
		}
	}
	function eu(e) {
		if (e.tag === 30) {
			var t = e.memoizedProps, n = wi(t, e.stateNode), r = Ul === null ? void 0 : Ul.get(n), i = Ei(t.default, r === void 0 ? t.exit : t.share);
			i !== "none" && (Jl(e, n, i, null, !1) ? r === void 0 ? Nd(e, t.onExit) : (i = e.stateNode, r.paired = i, i.paired = r, Ul.delete(n), Nd(e, t.onShare)) : Xl(e.child, !1)), Ul !== null && $l(e);
		} else if (e.subtreeFlags & 33554432) for (e = e.child; e !== null;) eu(e), e = e.sibling;
		else Ul !== null && $l(e);
	}
	function tu(e) {
		for (e = e.child; e !== null;) {
			if (e.tag === 30) {
				var t = e.memoizedProps, n = wi(t, e.stateNode);
				t = Ei(t.default, t.update), e.flags &= -5, t !== "none" && Jl(e, n, t, e.memoizedState = [], !1);
			} else e.subtreeFlags & 33554432 && tu(e);
			e = e.sibling;
		}
	}
	function nu(e) {
		if (e.subtreeFlags & 18874368) for (e = e.child; e !== null;) {
			if (e.tag !== 22 || e.memoizedState === null) {
				if (e.tag === 30 && e.flags & 18874368) {
					var t = e.stateNode;
					t.paired !== null && (t.paired = null, Xl(e.child, !1));
				}
				nu(e);
			}
			e = e.sibling;
		}
	}
	function ru(e) {
		if (e.tag === 30) e.stateNode.paired = null, Xl(e.child, !1), nu(e);
		else if (e.subtreeFlags & 33554432) for (e = e.child; e !== null;) ru(e), e = e.sibling;
		else nu(e);
	}
	function iu(e) {
		for (e = e.child; e !== null;) e.tag === 30 ? Xl(e.child, !1) : e.subtreeFlags & 33554432 && iu(e), e = e.sibling;
	}
	function au(e, t, n, r, i, a, o) {
		for (var s = !1; t !== null;) {
			if (t.tag === 5) {
				var c = t.stateNode;
				if (a !== null && ql < a.length) {
					var l = a[ql], u = Op(c);
					(l.view || u.view) && (s = !0);
					var d;
					if (d = !(e.flags & 4)) {
						if (u.clip) d = !0;
						else {
							d = l.rect;
							var f = u.rect;
							d = d.y !== f.y || d.x !== f.x || d.height !== f.height || d.width !== f.width;
						}
					}
					d && (e.flags |= 4), u.abs ? u = !l.abs : (l = l.rect, u = u.rect, u = l.height !== u.height || l.width !== u.width), u && (e.flags |= 32);
				} else e.flags |= 32;
				e.flags & 4 && Tp(c, ql === 0 ? n : n + "_" + ql, i), s && e.flags & 4 || (Gl === null && (Gl = []), Gl.push(c, ql === 0 ? r : r + "_" + ql, t.memoizedProps)), ql++;
			} else (t.tag !== 22 || t.memoizedState === null) && (t.tag === 30 && o ? e.flags |= t.flags & 32 : au(e, t.child, n, r, i, a, o) && (s = !0));
			t = t.sibling;
		}
		return s;
	}
	function ou(e, t) {
		for (e = e.child; e !== null;) {
			if (e.tag === 30) {
				var n = e.memoizedProps, r = e.stateNode, i = wi(n, r), a = Ei(n.default, n.update);
				if (t) {
					r = r.clones;
					var o = r === null ? null : r.map(kp);
				} else o = e.memoizedState, e.memoizedState = null;
				r = e;
				var s = e.child;
				ql = 0, i = au(r, s, i, i, a, o, !1), e.flags & 4 && i && (t || Nd(e, n.onUpdate));
			} else e.subtreeFlags & 33554432 && ou(e, t);
			e = e.sibling;
		}
	}
	var su = !1, H = !1, cu = !1, lu = !1, uu = typeof WeakSet == "function" ? WeakSet : Set, du = null, fu = !1, pu = !1, mu = !1, hu = !1;
	function gu(e, t, n) {
		if (e = e.containerInfo, sp = gh, e = ei(e), ti(e)) {
			if ("selectionStart" in e) var r = {
				start: e.selectionStart,
				end: e.selectionEnd
			};
			else a: {
				r = (r = e.ownerDocument) && r.defaultView || window;
				var i = r.getSelection && r.getSelection();
				if (i && i.rangeCount !== 0) {
					r = i.anchorNode;
					var a = i.anchorOffset, o = i.focusNode;
					i = i.focusOffset;
					try {
						r.nodeType, o.nodeType;
					} catch {
						r = null;
						break a;
					}
					var s = 0, c = -1, l = -1, u = 0, d = 0, f = e, p = null;
					b: for (;;) {
						for (var m; f !== r || a !== 0 && f.nodeType !== 3 || (c = s + a), f !== o || i !== 0 && f.nodeType !== 3 || (l = s + i), f.nodeType === 3 && (s += f.nodeValue.length), (m = f.firstChild) !== null;) p = f, f = m;
						for (;;) {
							if (f === e) break b;
							if (p === r && ++u === a && (c = s), p === o && ++d === i && (l = s), (m = f.nextSibling) !== null) break;
							f = p, p = f.parentNode;
						}
						f = m;
					}
					r = c === -1 || l === -1 ? null : {
						start: c,
						end: l
					};
				} else r = null;
			}
			r ||= {
				start: 0,
				end: 0
			};
		} else r = null;
		for (cp = {
			focusedElem: e,
			selectionRange: r
		}, gh = !1, n = (n & 335544064) === n, du = t, t = n ? 9270 : 1024; du !== null;) {
			if (e = du, n && (r = e.deletions, r !== null)) for (a = 0; a < r.length; a++) n && eu(r[a]);
			if (e.alternate === null && e.flags & 2) n && Wl(e), _u(n);
			else {
				if (e.tag === 22) {
					if (r = e.alternate, e.memoizedState !== null) {
						r !== null && r.memoizedState === null && n && eu(r), _u(n);
						continue;
					}
					if (r !== null && r.memoizedState !== null) {
						n && Wl(e), _u(n);
						continue;
					}
				}
				r = e.child, (e.subtreeFlags & t) !== 0 && r !== null ? (r.return = e, du = r) : (n && tu(e), _u(n));
			}
		}
		Ul = null;
	}
	function _u(e) {
		for (; du !== null;) {
			var t = du, n = e, r = t.alternate, i = t.flags;
			switch (t.tag) {
				case 0:
				case 11:
				case 15: break;
				case 1:
					if (i & 1024 && r !== null) {
						n = void 0, i = r.memoizedProps, r = r.memoizedState;
						var a = t.stateNode;
						try {
							var s = Tc(t.type, i);
							n = a.getSnapshotBeforeUpdate(s, r), a.__reactInternalSnapshotBeforeUpdate = n;
						} catch (e) {
							Z(t, t.return, e);
						}
					}
					break;
				case 3:
					if (i & 1024) {
						if (r = t.stateNode.containerInfo, n = r.nodeType, n === 9) nm(r);
						else if (n === 1) switch (r.nodeName) {
							case "HEAD":
							case "HTML":
							case "BODY":
								nm(r);
								break;
							default: r.textContent = "";
						}
					}
					break;
				case 5:
				case 26:
				case 27:
				case 6:
				case 4:
				case 17: break;
				case 30:
					n && r !== null && (n = wi(r.memoizedProps, r.stateNode), i = t.memoizedProps, i = Ei(i.default, i.update), i !== "none" && Jl(r, n, i, r.memoizedState = [], !0));
					break;
				default: if (i & 1024) throw Error(o(163));
			}
			if (r = t.sibling, r !== null) {
				r.return = t.return, du = r;
				break;
			}
			du = t.return;
		}
	}
	function vu(e, t, n) {
		var r = n.flags;
		switch (n.tag) {
			case 0:
			case 11:
			case 15:
				Lu(e, n), r & 4 && wl(5, n);
				break;
			case 1:
				if (Lu(e, n), r & 4) {
					if (e = n.stateNode, t === null) try {
						e.componentDidMount();
					} catch (e) {
						Z(n, n.return, e);
					}
					else {
						var i = Tc(n.type, t.memoizedProps);
						t = t.memoizedState;
						try {
							e.componentDidUpdate(i, t, e.__reactInternalSnapshotBeforeUpdate);
						} catch (e) {
							Z(n, n.return, e);
						}
					}
				}
				r & 64 && El(n), r & 512 && Ol(n, n.return);
				break;
			case 3:
				if (Lu(e, n), r & 64 && (e = n.updateQueue, e !== null)) {
					if (t = null, n.child !== null) switch (n.child.tag) {
						case 27:
						case 5:
							t = n.child.stateNode;
							break;
						case 1: t = n.child.stateNode;
					}
					try {
						jo(e, t);
					} catch (e) {
						Z(n, n.return, e);
					}
				}
				break;
			case 27: t === null && r & 4 && Vl(n);
			case 26:
			case 5:
				Lu(e, n), t === null && r & 4 && Fl(n), r & 512 && Ol(n, n.return);
				break;
			case 12:
				Lu(e, n);
				break;
			case 31:
				Lu(e, n), r & 4 && Eu(e, n);
				break;
			case 13:
				Lu(e, n), r & 4 && Du(e, n), r & 64 && (e = n.memoizedState, e !== null && (e = e.dehydrated, e !== null && (n = _f.bind(null, n), cm(e, n))));
				break;
			case 22:
				if (r = n.memoizedState !== null || su, !r) {
					var a = t !== null && t.memoizedState !== null || H;
					t = su, i = H, su = r, (H = a) && !i ? (r = 2, n.subtreeFlags & 8772 && (r |= 1), zu(e, n, r)) : Lu(e, n), su = t, H = i;
				}
				break;
			case 30:
				Lu(e, n), r & 512 && Ol(n, n.return);
				break;
			case 7: r & 512 && Ol(n, n.return);
			default: Lu(e, n);
		}
	}
	function yu(e, t) {
		for (e = e.child; e !== null;) bu(e, t), e = e.sibling;
	}
	function bu(e, t) {
		switch (e.tag) {
			case 5:
			case 26:
				try {
					var n = e.stateNode;
					if (t) {
						var r = n.style;
						typeof r.setProperty == "function" ? r.setProperty("display", "none", "important") : r.display = "none";
					} else {
						var i = e.stateNode, a = e.memoizedProps.style, o = a != null && a.hasOwnProperty("display") ? a.display : null;
						i.style.display = o == null || typeof o == "boolean" ? "" : ("" + o).trim();
					}
				} catch (t) {
					Z(e, e.return, t);
				}
				xu(e, t);
				break;
			case 6:
				try {
					e.stateNode.nodeValue = t ? "" : e.memoizedProps, M = !0;
				} catch (t) {
					Z(e, e.return, t);
				}
				break;
			case 18:
				try {
					var s = e.stateNode;
					t ? wp(s, !0) : wp(e.stateNode, !1);
				} catch (t) {
					Z(e, e.return, t);
				}
				break;
			case 22:
			case 23:
				e.memoizedState === null && yu(e, t);
				break;
			default: yu(e, t);
		}
	}
	function xu(e, t) {
		if (e.subtreeFlags & 67108864) for (e = e.child; e !== null;) {
			a: {
				var n = e, r = t;
				switch (n.tag) {
					case 4:
						bu(n, r);
						break a;
					case 22:
						n.memoizedState === null && xu(n, r);
						break a;
					default: xu(n, r);
				}
			}
			e = e.sibling;
		}
	}
	function Su(e) {
		var t = e.alternate;
		t !== null && (e.alternate = null, Su(t)), e.child = null, e.deletions = null, e.sibling = null, e.tag === 5 && (t = e.stateNode, t !== null && Rt(t)), e.stateNode = null, e.return = null, e.dependencies = null, e.memoizedProps = null, e.memoizedState = null, e.pendingProps = null, e.stateNode = null, e.updateQueue = null;
	}
	var U = null, Cu = !1;
	function wu(e, t, n) {
		for (n = n.child; n !== null;) Tu(e, t, n), n = n.sibling;
	}
	function Tu(e, t, n) {
		if (rt && typeof rt.onCommitFiberUnmount == "function") try {
			rt.onCommitFiberUnmount(nt, n);
		} catch {}
		switch (n.tag) {
			case 26:
				H || kl(n, t), wu(e, t, n), n.memoizedState ? n.memoizedState.count-- : n.stateNode && !H && (n = n.stateNode, n.parentNode.removeChild(n));
				break;
			case 27:
				H || kl(n, t), Ml(n);
				var r = U, i = Cu;
				Sp(n.type) && (U = n.stateNode, Cu = !1), wu(e, t, n), gm(n.stateNode, n.type, n.memoizedProps), U = r, Cu = i;
				break;
			case 5: H || kl(n, t), Ml(n);
			case 6:
				if (n.tag === 6 && Ml(n), r = U, i = Cu, U = null, wu(e, t, n), U = r, Cu = i, U !== null) {
					if (Cu) try {
						(U.nodeType === 9 ? U.body : U.nodeName === "HTML" ? U.ownerDocument.body : U).removeChild(n.stateNode), M = !0;
					} catch (e) {
						Z(n, t, e);
					}
					else try {
						U.removeChild(n.stateNode), M = !0;
					} catch (e) {
						Z(n, t, e);
					}
				}
				break;
			case 18:
				U !== null && (Cu ? (e = U, Cp(e.nodeType === 9 ? e.body : e.nodeName === "HTML" ? e.ownerDocument.body : e, n.stateNode), Hh(e)) : Cp(U, n.stateNode));
				break;
			case 4:
				r = U, i = Cu, U = n.stateNode.containerInfo, Cu = !0, wu(e, t, n), U = r, Cu = i;
				break;
			case 0:
			case 11:
			case 14:
			case 15:
				Tl(2, n, t), H || Tl(4, n, t), wu(e, t, n);
				break;
			case 1:
				H || (kl(n, t), r = n.stateNode, typeof r.componentWillUnmount == "function" && Dl(n, t, r)), wu(e, t, n);
				break;
			case 21:
				wu(e, t, n);
				break;
			case 22:
				H = (r = H) || n.memoizedState !== null, wu(e, t, n), H = r;
				break;
			case 30:
				kl(n, t), wu(e, t, n);
				break;
			case 7:
				H || kl(n, t), wu(e, t, n);
				break;
			default: wu(e, t, n);
		}
	}
	function Eu(e, t) {
		if (t.memoizedState === null && (e = t.alternate, e !== null && (e = e.memoizedState, e !== null))) {
			e = e.dehydrated;
			try {
				Hh(e);
			} catch (e) {
				Z(t, t.return, e);
			}
		}
	}
	function Du(e, t) {
		if (t.memoizedState === null && (e = t.alternate, e !== null && (e = e.memoizedState, e !== null && (e = e.dehydrated, e !== null)))) try {
			Hh(e);
		} catch (e) {
			Z(t, t.return, e);
		}
	}
	function Ou(e) {
		switch (e.tag) {
			case 31:
			case 13:
			case 19:
				var t = e.stateNode;
				return t === null && (t = e.stateNode = new uu()), t;
			case 22: return e = e.stateNode, t = e._retryCache, t === null && (t = e._retryCache = new uu()), t;
			default: throw Error(o(435, e.tag));
		}
	}
	function ku(e, t) {
		var n = Ou(e);
		t.forEach(function(t) {
			if (!n.has(t)) {
				n.add(t);
				var r = vf.bind(null, e, t);
				t.then(r, r);
			}
		});
	}
	function Au(e, t, n) {
		var r = t.deletions;
		if (r !== null) for (var i = 0; i < r.length; i++) {
			var a = r[i], s = e, c = t, l = c;
			a: for (; l !== null;) {
				switch (l.tag) {
					case 27:
						if (Sp(l.type)) {
							U = l.stateNode, Cu = !1;
							break a;
						}
						break;
					case 5:
						U = l.stateNode, Cu = !1;
						break a;
					case 3:
					case 4:
						U = l.stateNode.containerInfo, Cu = !0;
						break a;
				}
				l = l.return;
			}
			if (U === null) throw Error(o(160));
			Tu(s, c, a), U = null, Cu = !1, s = a.alternate, s !== null && (s.return = null), a.return = null;
		}
		if (t.subtreeFlags & 13886) for (t = t.child; t !== null;) Mu(t, e, n), t = t.sibling;
	}
	var ju = null;
	function Mu(e, t, n) {
		var r = e.alternate, i = e.flags;
		switch (e.tag) {
			case 0:
			case 11:
			case 14:
			case 15:
				if (i & 4 && (r = e.updateQueue, r = r === null ? null : r.events, r !== null)) for (var a = 0; a < r.length; a++) {
					var s = r[a];
					s.ref.impl = s.nextImpl;
				}
				Au(t, e, n), Nu(e), i & 4 && (Tl(3, e, e.return), wl(3, e), Tl(5, e, e.return));
				break;
			case 1:
				Au(t, e, n), Nu(e), i & 512 && (H || r === null || kl(r, r.return)), i & 64 && su && (e = e.updateQueue, e !== null && (t = e.callbacks, t !== null && (n = e.shared.hiddenCallbacks, e.shared.hiddenCallbacks = n === null ? t : n.concat(t))));
				break;
			case 26:
				if (a = ju, Au(t, e, n), Nu(e), i & 512 && (H || r === null || kl(r, r.return)), i & 4) {
					if (i = r === null ? null : r.memoizedState, n = e.memoizedState, r === null) {
						if (n === null) {
							if (e.stateNode === null) {
								if (su) e.stateNode = fp(e.type, e.memoizedProps, t.containerInfo, e);
								else {
									a: {
										t = e.type, n = e.memoizedProps, i = a.ownerDocument || a;
										b: switch (t) {
											case "title":
												r = i.getElementsByTagName("title")[0], (!r || r[It] || r[kt] || r.namespaceURI === "http://www.w3.org/2000/svg" || r.hasAttribute("itemprop")) && (r = i.createElement(t), i.head.insertBefore(r, i.querySelector("head > title"))), np(r, t, n), r[kt] = e, Ut(r), t = r;
												break a;
											case "link":
												if (a = Gm("link", "href", i).get(t + (n.href || ""))) {
													for (s = 0; s < a.length; s++) if (r = a[s], r.getAttribute("href") === (n.href == null || n.href === "" ? null : n.href) && r.getAttribute("rel") === (n.rel == null ? null : n.rel) && r.getAttribute("title") === (n.title == null ? null : n.title) && r.getAttribute("crossorigin") === (n.crossOrigin == null ? null : n.crossOrigin)) {
														a.splice(s, 1);
														break b;
													}
												}
												r = i.createElement(t), np(r, t, n), i.head.appendChild(r);
												break;
											case "meta":
												if (a = Gm("meta", "content", i).get(t + (n.content || ""))) {
													for (s = 0; s < a.length; s++) if (r = a[s], r.getAttribute("content") === (n.content == null ? null : "" + n.content) && r.getAttribute("name") === (n.name == null ? null : n.name) && r.getAttribute("property") === (n.property == null ? null : n.property) && r.getAttribute("http-equiv") === (n.httpEquiv == null ? null : n.httpEquiv) && r.getAttribute("charset") === (n.charSet == null ? null : n.charSet)) {
														a.splice(s, 1);
														break b;
													}
												}
												r = i.createElement(t), np(r, t, n), i.head.appendChild(r);
												break;
											default: throw Error(o(468, t));
										}
										r[kt] = e, Ut(r), t = r;
									}
									e.stateNode = t;
								}
							} else su || Km(a, e.type, e.stateNode);
						} else e.stateNode = Bm(a, n, e.memoizedProps);
					} else i === n ? n === null && e.stateNode !== null && Il(e, e.memoizedProps, r.memoizedProps) : (i === null ? (t = r.stateNode, t === null || H || t.parentNode.removeChild(t)) : i.count--, n === null ? su || Km(a, e.type, e.stateNode) : Bm(a, n, e.memoizedProps));
				}
				break;
			case 27:
				Au(t, e, n), Nu(e), i & 512 && (H || r === null || kl(r, r.return)), r !== null && i & 4 && Il(e, e.memoizedProps, r.memoizedProps);
				break;
			case 5:
				if (a = cu, cu = !1, Au(t, e, n), cu = a, Nu(e), i & 512 && (H || r === null || kl(r, r.return)), e.flags & 32) {
					t = e.stateNode;
					try {
						_n(t, ""), M = !0;
					} catch (t) {
						Z(e, e.return, t);
					}
				}
				i & 4 && e.stateNode != null && (t = e.memoizedProps, Il(e, t, r === null ? t : r.memoizedProps)), i & 1024 && (lu = !0);
				break;
			case 6:
				if (Au(t, e, n), Nu(e), i & 4) {
					if (e.stateNode === null) throw Error(o(162));
					t = e.memoizedProps, n = e.stateNode;
					try {
						n.nodeValue = t, M = !0;
					} catch (t) {
						Z(e, e.return, t);
					}
				}
				break;
			case 3:
				if (M = !1, Wm = null, a = ju, ju = bm(t.containerInfo), Au(t, e, n), ju = a, Nu(e), i & 4 && r !== null && r.memoizedState.isDehydrated) try {
					Hh(t.containerInfo);
				} catch (t) {
					Z(e, e.return, t);
				}
				lu && (lu = !1, Pu(e)), M = !1;
				break;
			case 4:
				i = cu, cu = su, r = $t(), a = ju, ju = bm(e.stateNode.containerInfo), Au(t, e, n), Nu(e), ju = a, M && pu && (mu = !0), M = r, cu = i;
				break;
			case 12:
				Au(t, e, n), Nu(e);
				break;
			case 31:
				Au(t, e, n), Nu(e), i & 4 && (t = e.updateQueue, t !== null && (e.updateQueue = null, ku(e, t)));
				break;
			case 13:
				Au(t, e, n), Nu(e), e.child.flags & 8192 && e.memoizedState !== null != (r !== null && r.memoizedState !== null) && (hd = qe()), i & 4 && (t = e.updateQueue, t !== null && (e.updateQueue = null, ku(e, t)));
				break;
			case 22:
				a = e.memoizedState !== null, s = r !== null && r.memoizedState !== null;
				var c = su, l = H, u = cu;
				su = c || a, cu = u || a, H = l || s, Au(t, e, n), H = l, cu = u, su = c, Nu(e), i & 8192 && (t = e.stateNode, t._visibility = a ? t._visibility & -2 : t._visibility | 1, !a || r === null || s || su || H || (t = s || H, n = su, r = H, su = a || su, H = t, Ru(e, 2), su = n, H = r), !a && cu || yu(e, a)), i & 4 && (t = e.updateQueue, t !== null && (n = t.retryQueue, n !== null && (t.retryQueue = null, ku(e, n))));
				break;
			case 19:
				Au(t, e, n), Nu(e), i & 4 && (t = e.updateQueue, t !== null && (e.updateQueue = null, ku(e, t)));
				break;
			case 30:
				i & 512 && (H || r === null || kl(r, r.return)), i = $t(), a = pu, s = (n & 335544064) === n, c = e.memoizedProps, pu = s && Ei(c.default, c.update) !== "none", Au(t, e, n), Nu(e), s && r !== null && M && (e.flags |= 4), pu = a, M = i;
				break;
			case 21: break;
			case 7: i & 512 && (H || r === null || kl(r, r.return)), r && r.stateNode !== null && (r.stateNode._fragmentFiber = e);
			default: Au(t, e, n), Nu(e);
		}
	}
	function Nu(e) {
		var t = e.flags;
		if (t & 2) {
			try {
				for (var n, r = e.return; r !== null;) {
					if (Ll(r)) {
						n = r;
						break;
					}
					r = r.return;
				}
				r = null;
				for (var i = e.return; i !== null;) {
					if (Pl(i)) {
						var a = i.stateNode;
						r === null ? r = [a] : r.push(a);
					}
					if (Nl(i)) break;
					i = i.return;
				}
				var s = r;
				if (n == null) throw Error(o(160));
				switch (n.tag) {
					case 27:
						var c = n.stateNode;
						Bl(e, Rl(e), c, s);
						break;
					case 5:
						var l = n.stateNode;
						n.flags & 32 && (_n(l, ""), n.flags &= -33), Bl(e, Rl(e), l, s);
						break;
					case 3:
					case 4:
						var u = n.stateNode.containerInfo;
						zl(e, Rl(e), u, s);
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
	function Pu(e) {
		if (e.subtreeFlags & 1024) for (e = e.child; e !== null;) {
			var t = e;
			Pu(t), t.tag === 5 && t.flags & 1024 && (t = t.stateNode, gh = !0, t.reset(), gh = !1), e = e.sibling;
		}
	}
	function Fu(e, t) {
		if (t.subtreeFlags & 9270) for (t = t.child; t !== null;) Iu(t, e), t = t.sibling;
		else ou(t, !1);
	}
	function Iu(e, t) {
		var n = e.alternate;
		if (n === null) Ql(e, !1);
		else switch (e.tag) {
			case 3:
				if (hu = fu = !1, Kl(), Fu(t, e), !fu && !mu) {
					if (e = Gl, e !== null) for (var r = 0; r < e.length; r += 3) {
						n = e[r];
						var i = e[r + 1];
						Ep(n, e[r + 2]), n = n.ownerDocument.documentElement, n !== null && n.animate({
							opacity: [0, 0],
							pointerEvents: ["none", "none"]
						}, {
							duration: 0,
							fill: "forwards",
							pseudoElement: "::view-transition-group(" + i + ")"
						});
					}
					e = t.containerInfo, e = e.nodeType === 9 ? e.documentElement : e.ownerDocument.documentElement, e !== null && e.style.viewTransitionName === "" && (e.style.viewTransitionName = "none", e.animate({
						opacity: [0, 0],
						pointerEvents: ["none", "none"]
					}, {
						duration: 0,
						fill: "forwards",
						pseudoElement: "::view-transition-group(root)"
					}), e.animate({
						width: [0, 0],
						height: [0, 0]
					}, {
						duration: 0,
						fill: "forwards",
						pseudoElement: "::view-transition"
					})), hu = !0;
				}
				Gl = null;
				break;
			case 5:
				Fu(t, e);
				break;
			case 4:
				r = fu, fu = !1, Fu(t, e), fu && (mu = !0), fu = r;
				break;
			case 22:
				e.memoizedState === null && (n.memoizedState === null ? Fu(t, e) : Ql(e, !1));
				break;
			case 30:
				r = fu, i = Kl(), fu = !1, Fu(t, e), fu && (e.flags |= 4);
				var a = e.memoizedProps, o = e.stateNode;
				t = wi(a, o), o = wi(n.memoizedProps, o);
				var s = Ei(a.default, a.update);
				s === "none" ? t = !1 : (a = n.memoizedState, n.memoizedState = null, n = e.child, ql = 0, t = au(e, n, t, o, s, a, !0), ql !== (a === null ? 0 : a.length) && (e.flags |= 32)), e.flags & 4 && t ? (Nd(e, e.memoizedProps.onUpdate), Gl = i) : i !== null && (i.push.apply(i, Gl), Gl = i), fu = e.flags & 32 ? !0 : r;
				break;
			default: Fu(t, e);
		}
	}
	function Lu(e, t) {
		if (t.subtreeFlags & 8772) for (t = t.child; t !== null;) vu(e, t.alternate, t), t = t.sibling;
	}
	function Ru(e, t) {
		for (e = e.child; e !== null;) {
			var n = e, r = t;
			switch (n.tag) {
				case 0:
				case 11:
				case 14:
				case 15:
					Tl(4, n, n.return), Ru(n, r);
					break;
				case 1:
					kl(n, n.return);
					var i = n.stateNode;
					typeof i.componentWillUnmount == "function" && Dl(n, n.return, i), Ru(n, r);
					break;
				case 27: r & 2 && gm(n.stateNode, n.type, n.memoizedProps);
				case 5:
					kl(n, n.return), n.tag !== 5 && n.tag !== 27 || Ml(n), Ru(n, r);
					break;
				case 6:
					Ml(n);
					break;
				case 26:
					kl(n, n.return), i = n.stateNode, n.memoizedState !== null || i === null || H || i.parentNode.removeChild(i), Ru(n, r);
					break;
				case 22:
					n.memoizedState === null && Ru(n, r);
					break;
				case 30:
					kl(n, n.return), Ru(n, r);
					break;
				case 7: kl(n, n.return);
				default: Ru(n, r);
			}
			e = e.sibling;
		}
	}
	function zu(e, t, n) {
		for (n = t.subtreeFlags & 8772 ? n : n & -2, t = t.child; t !== null;) {
			var r = t.alternate, i = e, a = t, o = a.flags, s = !!(n & 1);
			switch (a.tag) {
				case 0:
				case 11:
				case 15:
					zu(i, a, n), wl(4, a);
					break;
				case 1:
					if (zu(i, a, n), r = a, i = r.stateNode, typeof i.componentDidMount == "function") try {
						i.componentDidMount();
					} catch (e) {
						Z(r, r.return, e);
					}
					if (r = a, i = r.updateQueue, i !== null) {
						var c = r.stateNode;
						try {
							var l = i.shared.hiddenCallbacks;
							if (l !== null) for (i.shared.hiddenCallbacks = null, i = 0; i < l.length; i++) Ao(l[i], c);
						} catch (e) {
							Z(r, r.return, e);
						}
					}
					s && o & 64 && El(a), Ol(a, a.return);
					break;
				case 27: n & 2 && Vl(a);
				case 5:
					a.tag !== 5 && a.tag !== 27 || jl(a), zu(i, a, n), s && r === null && o & 4 && Fl(a), Ol(a, a.return);
					break;
				case 6:
					jl(a);
					break;
				case 26:
					c = a.stateNode, a.memoizedState !== null || c === null || su || Km(bm(c.ownerDocument), a.type, c), zu(i, a, n), s && r === null && o & 4 && Fl(a), Ol(a, a.return);
					break;
				case 12:
					zu(i, a, n);
					break;
				case 31:
					zu(i, a, n), s && o & 4 && Eu(i, a);
					break;
				case 13:
					zu(i, a, n), s && o & 4 && Du(i, a);
					break;
				case 22:
					a.memoizedState === null && zu(i, a, n), Ol(a, a.return);
					break;
				case 30:
					zu(i, a, n), Ol(a, a.return);
					break;
				case 7: Ol(a, a.return);
				default: zu(i, a, n);
			}
			t = t.sibling;
		}
	}
	function Bu(e, t) {
		var n = null;
		e !== null && e.memoizedState !== null && e.memoizedState.cachePool !== null && (n = e.memoizedState.cachePool.pool), e = null, t.memoizedState !== null && t.memoizedState.cachePool !== null && (e = t.memoizedState.cachePool.pool), e !== n && (e != null && e.refCount++, n != null && za(n));
	}
	function Vu(e, t) {
		e = null, t.alternate !== null && (e = t.alternate.memoizedState.cache), t = t.memoizedState.cache, t !== e && (t.refCount++, e != null && za(e));
	}
	function Hu(e, t, n, r) {
		var i = (n & 335544064) === n;
		if (t.subtreeFlags & (i ? 10262 : 10256)) for (t = t.child; t !== null;) Uu(e, t, n, r), t = t.sibling;
		else i && iu(t);
	}
	function Uu(e, t, n, r) {
		var i = (n & 335544064) === n;
		i && t.alternate === null && t.return !== null && t.return.alternate !== null && ru(t);
		var a = t.flags;
		switch (t.tag) {
			case 0:
			case 11:
			case 15:
				Hu(e, t, n, r), a & 2048 && wl(9, t);
				break;
			case 1:
				Hu(e, t, n, r);
				break;
			case 3:
				Hu(e, t, n, r), i && hu && (e = e.containerInfo, e = e.nodeType === 9 ? e.body : e.nodeName === "HTML" ? e.ownerDocument.body : e, e.style.viewTransitionName === "root" && (e.style.viewTransitionName = ""), e = e.ownerDocument.documentElement, e !== null && e.style.viewTransitionName === "none" && (e.style.viewTransitionName = "")), a & 2048 && (a = null, t.alternate !== null && (a = t.alternate.memoizedState.cache), t = t.memoizedState.cache, t !== a && (t.refCount++, a != null && za(a)));
				break;
			case 12:
				if (a & 2048) {
					Hu(e, t, n, r), a = t.stateNode;
					try {
						var o = t.memoizedProps, s = o.id, c = o.onPostCommit;
						typeof c == "function" && c(s, t.alternate === null ? "mount" : "update", a.passiveEffectDuration, -0);
					} catch (e) {
						Z(t, t.return, e);
					}
				} else Hu(e, t, n, r);
				break;
			case 31:
				Hu(e, t, n, r);
				break;
			case 13:
				Hu(e, t, n, r);
				break;
			case 23: break;
			case 22:
				o = t.stateNode, s = t.alternate, t.memoizedState === null ? (i && s !== null && s.memoizedState !== null && ru(t), o._visibility & 2 ? Hu(e, t, n, r) : (o._visibility |= 2, Wu(e, t, n, r, !!(t.subtreeFlags & 10256) || !1))) : (i && s !== null && s.memoizedState === null && ru(s), o._visibility & 2 ? Hu(e, t, n, r) : Gu(e, t)), a & 2048 && Bu(s, t);
				break;
			case 24:
				Hu(e, t, n, r), a & 2048 && Vu(t.alternate, t);
				break;
			case 30:
				i && (a = t.alternate, a !== null && (Xl(a.child, !0), Xl(t.child, !0))), Hu(e, t, n, r);
				break;
			default: Hu(e, t, n, r);
		}
	}
	function Wu(e, t, n, r, i) {
		for (i &&= !!(t.subtreeFlags & 10256) || !1, t = t.child; t !== null;) {
			var a = e, o = t, s = n, c = r, l = o.flags;
			switch (o.tag) {
				case 0:
				case 11:
				case 15:
					Wu(a, o, s, c, i), wl(8, o);
					break;
				case 23: break;
				case 22:
					var u = o.stateNode;
					o.memoizedState === null ? (u._visibility |= 2, Wu(a, o, s, c, i)) : u._visibility & 2 ? Wu(a, o, s, c, i) : Gu(a, o), i && l & 2048 && Bu(o.alternate, o);
					break;
				case 24:
					Wu(a, o, s, c, i), i && l & 2048 && Vu(o.alternate, o);
					break;
				default: Wu(a, o, s, c, i);
			}
			t = t.sibling;
		}
	}
	function Gu(e, t) {
		if (t.subtreeFlags & 10256) for (t = t.child; t !== null;) {
			var n = e, r = t, i = r.flags;
			switch (r.tag) {
				case 22:
					Gu(n, r), i & 2048 && Bu(r.alternate, r);
					break;
				case 24:
					Gu(n, r), i & 2048 && Vu(r.alternate, r);
					break;
				default: Gu(n, r);
			}
			t = t.sibling;
		}
	}
	var Ku = 8192;
	function qu(e, t, n) {
		if (e.subtreeFlags & Ku) for (e = e.child; e !== null;) Ju(e, t, n), e = e.sibling;
	}
	function Ju(e, t, n) {
		switch (e.tag) {
			case 26:
				qu(e, t, n), e.flags & Ku && (e.memoizedState === null ? (e = e.stateNode, (t & 335544128) === t && Zm(n, e)) : Qm(n, ju, e.memoizedState, e.memoizedProps));
				break;
			case 5:
				qu(e, t, n), e.flags & Ku && (e = e.stateNode, (t & 335544128) === t && Zm(n, e));
				break;
			case 3:
			case 4:
				var r = ju;
				ju = bm(e.stateNode.containerInfo), qu(e, t, n), ju = r;
				break;
			case 22:
				e.memoizedState === null && (r = e.alternate, r !== null && r.memoizedState !== null ? (r = Ku, Ku = 16777216, qu(e, t, n), Ku = r) : qu(e, t, n));
				break;
			case 30:
				if ((e.flags & Ku) !== 0 && (r = e.memoizedProps.name, r != null && r !== "auto")) {
					var i = e.stateNode;
					i.paired = null, Ul === null && (Ul = /* @__PURE__ */ new Map()), Ul.set(r, i);
				}
				qu(e, t, n);
				break;
			default: qu(e, t, n);
		}
	}
	function Yu(e) {
		var t = e.alternate;
		if (t !== null && (e = t.child, e !== null)) {
			t.child = null;
			do
				t = e.sibling, e.sibling = null, e = t;
			while (e !== null);
		}
	}
	function Xu(e) {
		var t = e.deletions;
		if (e.flags & 16) {
			if (t !== null) for (var n = 0; n < t.length; n++) {
				var r = t[n];
				du = r, $u(r, e);
			}
			Yu(e);
		}
		if (e.subtreeFlags & 10256) for (e = e.child; e !== null;) Zu(e), e = e.sibling;
	}
	function Zu(e) {
		switch (e.tag) {
			case 0:
			case 11:
			case 15:
				Xu(e), e.flags & 2048 && Tl(9, e, e.return);
				break;
			case 3:
				Xu(e);
				break;
			case 12:
				Xu(e);
				break;
			case 22:
				var t = e.stateNode;
				e.memoizedState !== null && t._visibility & 2 && (e.return === null || e.return.tag !== 13) ? (t._visibility &= -3, Qu(e)) : Xu(e);
				break;
			default: Xu(e);
		}
	}
	function Qu(e) {
		var t = e.deletions;
		if (e.flags & 16) {
			if (t !== null) for (var n = 0; n < t.length; n++) {
				var r = t[n];
				du = r, $u(r, e);
			}
			Yu(e);
		}
		for (e = e.child; e !== null;) {
			switch (t = e, t.tag) {
				case 0:
				case 11:
				case 15:
					Tl(8, t, t.return), Qu(t);
					break;
				case 22:
					n = t.stateNode, n._visibility & 2 && (n._visibility &= -3, Qu(t));
					break;
				default: Qu(t);
			}
			e = e.sibling;
		}
	}
	function $u(e, t) {
		for (; du !== null;) {
			var n = du;
			switch (n.tag) {
				case 0:
				case 11:
				case 15:
					Tl(8, n, t);
					break;
				case 23:
				case 22:
					if (n.memoizedState !== null && n.memoizedState.cachePool !== null) {
						var r = n.memoizedState.cachePool.pool;
						r != null && r.refCount++;
					}
					break;
				case 24: za(n.memoizedState.cache);
			}
			if (r = n.child, r !== null) r.return = n, du = r;
			else a: for (n = e; du !== null;) {
				r = du;
				var i = r.sibling, a = r.return;
				if (Su(r), r === n) {
					du = null;
					break a;
				}
				if (i !== null) {
					i.return = a, du = i;
					break a;
				}
				du = a;
			}
		}
	}
	var ed = {
		getCacheForType: function(e) {
			var t = ja(La), n = t.data.get(e);
			return n === void 0 && (n = e(), t.data.set(e, n)), n;
		},
		cacheSignal: function() {
			return ja(La).controller.signal;
		}
	}, td = typeof WeakMap == "function" ? WeakMap : Map, W = 0, G = null, K = null, q = 0, J = 0, nd = null, rd = !1, id = !1, ad = !1, od = 0, Y = 0, sd = 0, cd = 0, ld = 0, ud = 0, dd = 0, fd = null, pd = null, md = !1, hd = 0, gd = 0, _d = Infinity, vd = null, yd = null, X = 0, bd = null, xd = null, Sd = 0, Cd = 0, wd = null, Td = null, Ed = null, Dd = null, Od = null, kd = 0, Ad = null;
	function jd() {
		return W & 2 && q !== 0 ? q & -q : O.T === null ? Et() : Pf();
	}
	function Md() {
		if (ud === 0) {
			if (!(q & 536870912) || P) {
				var e = ut;
				ut <<= 1, !(ut & 3932160) && (ut = 262144), ud = e;
			} else ud = 536870912;
		}
		return e = Lo.current, e !== null && (e.flags |= 32), ud;
	}
	function Nd(e, t) {
		if (t != null) {
			var n = e.stateNode, r = n.ref;
			r === null && (r = n.ref = Pp(wi(e.memoizedProps, n))), Dd === null && (Dd = []), Dd.push(t.bind(null, r));
		}
	}
	function Pd(e, t, n) {
		(e === G && (J === 2 || J === 9) || e.cancelPendingCommit !== null) && (Vd(e, 0), Rd(e, q, ud, !1)), yt(e, n), W & 2 && e === G || (e === G && (!(W & 2) && (cd |= n), Y === 4 && Rd(e, q, ud, !1)), Ef(e));
	}
	function Fd(e, t, n) {
		if (W & 6) throw Error(o(327));
		var r = !n && !(t & 127) && (t & e.expiredLanes) === 0 || mt(e, t), i = r ? Yd(e, t) : qd(e, t, !0), a = r;
		do {
			if (i === 0) {
				id && !r && Rd(e, t, 0, !1);
				break;
			}
			if (n = e.current.alternate, a && !Ld(n)) i = qd(e, t, !1), a = !1;
			else {
				if (i === 2) {
					if (a = t, e.errorRecoveryDisabledLanes & a) var s = 0;
					else s = e.pendingLanes & -536870913, s = s === 0 ? s & 536870912 ? 536870912 : 0 : s;
					if (s !== 0) {
						t = s;
						a: {
							var c = e;
							i = fd;
							var l = c.current.memoizedState.isDehydrated;
							if (l && (Vd(c, s).flags |= 256), s = qd(c, s, !1), s !== 2 && s !== 6) {
								if (ad && !l) {
									c.errorRecoveryDisabledLanes |= a, cd |= a, i = 4;
									break a;
								}
								a = pd, pd = i, a !== null && (pd === null ? pd = a : pd.push.apply(pd, a));
							}
							i = s;
						}
						if (a = !1, i !== 2) continue;
					}
				}
				if (i === 1) {
					Vd(e, 0), Rd(e, t, 0, !0);
					break;
				}
				a: {
					switch (r = e, a = i, a) {
						case 0:
						case 1: throw Error(o(345));
						case 4: if ((t & 4194048) !== t && (t & 62914560) !== t) break;
						case 6:
							Rd(r, t, ud, !rd);
							break a;
						case 2:
							pd = null;
							break;
						case 3:
						case 5: break;
						default: throw Error(o(329));
					}
					if ((t & 62914560) === t && (i = hd + 300 - qe(), 10 < i)) {
						if (Rd(r, t, ud, !rd), pt(r, 0, !0) !== 0) break a;
						Sd = t, r.timeoutHandle = gp(Id.bind(null, r, n, pd, vd, md, t, ud, cd, dd, rd, a, "Throttled", -0, 0), i);
						break a;
					}
					Id(r, n, pd, vd, md, t, ud, cd, dd, rd, a, null, -0, 0);
				}
				break;
			}
		} while (1);
		Ef(e);
	}
	function Id(e, t, n, r, i, a, o, s, c, l, u, d, f, p) {
		e.timeoutHandle = -1;
		var m = t.subtreeFlags, h = (a & 335544064) === a;
		d = null, (h || m & 8192 || (m & 16785408) == 16785408) && (d = {
			stylesheets: null,
			count: 0,
			imgCount: 0,
			imgBytes: 0,
			suspenseyImages: [],
			waitingForImages: !0,
			waitingForViewTransition: !1,
			unsuspend: Tn
		}, Ul = null, Ju(t, a, d), h && (m = d, h = e.containerInfo, h = (h.nodeType === 9 ? h : h.ownerDocument).__reactViewTransition, h != null && (m.count++, m.waitingForViewTransition = !0, m = nh.bind(m), h.finished.then(m, m))), m = (a & 62914560) === a ? hd - qe() : (a & 4194048) === a ? gd - qe() : 0, m = eh(d, m), m !== null) ? (Sd = a, e.cancelPendingCommit = m(nf.bind(null, e, t, a, n, r, i, o, s, c, l, u, d, null, f, p)), Rd(e, a, o, !l)) : nf(e, t, a, n, r, i, o, s, c, l, u, d);
	}
	function Ld(e) {
		for (var t = e;;) {
			var n = t.tag;
			if ((n === 0 || n === 11 || n === 15) && t.flags & 16384 && (n = t.updateQueue, n !== null && (n = n.stores, n !== null))) for (var r = 0; r < n.length; r++) {
				var i = n[r], a = i.getSnapshot;
				i = i.value;
				try {
					if (!Jr(a(), i)) return !1;
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
	function Rd(e, t, n, r) {
		t = ht(e, t), t &= ~ld, t &= ~cd, e.suspendedLanes |= t, e.pingedLanes &= ~t, r && (e.warmLanes |= t), r = e.expirationTimes;
		for (var i = t; 0 < i;) {
			var a = 31 - at(i), o = 1 << a;
			r[a] = -1, i &= ~o;
		}
		n !== 0 && xt(e, n, t);
	}
	function zd() {
		return W & 6 ? !0 : (Df(0, !1), !1);
	}
	function Bd() {
		if (K !== null) {
			if (J === 0) var e = K.return;
			else e = K, Ca = Sa = null, cs(e), fo = null, po = 0, e = K;
			for (; e !== null;) Cl(e.alternate, e), e = e.return;
			K = null;
		}
	}
	function Vd(e, t) {
		var n = e.timeoutHandle;
		return n !== -1 && (e.timeoutHandle = -1, _p(n)), n = e.cancelPendingCommit, n !== null && (e.cancelPendingCommit = null, n()), Sd = 0, Bd(), G = e, K = n = Vi(e.current, null), q = t, J = 0, nd = null, rd = !1, id = mt(e, t), ad = !1, dd = ud = ld = cd = sd = Y = 0, pd = fd = null, md = !1, od = ht(e, t), ji(), n;
	}
	function Hd(e, t) {
		I = null, O.H = _c, t === to || t === ro ? (t = lo(), J = 3) : t === no ? (t = lo(), J = 4) : J = t === Fc ? 8 : typeof t == "object" && t && typeof t.then == "function" ? 6 : 1, nd = t, K === null && (Y = 1, kc(e, Yi(t, e.current)));
	}
	function Ud() {
		var e = Lo.current;
		return e === null ? !0 : (q & 4194048) === q ? Ro === null : (q & 62914560) === q || q & 536870912 ? e === Ro : !1;
	}
	function Wd() {
		var e = O.H;
		return O.H = _c, e === null ? _c : e;
	}
	function Gd() {
		var e = O.A;
		return O.A = ed, e;
	}
	function Kd() {
		Y = 4, rd || (q & 4194048) !== q && Lo.current !== null || (id = !0), !(sd & 134217727) && !(cd & 134217727) || G === null || Rd(G, q, ud, !1);
	}
	function qd(e, t, n) {
		var r = W;
		W |= 2;
		var i = Wd(), a = Gd();
		(G !== e || q !== t) && (vd = null, Vd(e, t)), t = !1;
		var o = Y;
		a: do
			try {
				if (J !== 0 && K !== null) {
					var s = K, c = nd;
					switch (J) {
						case 8:
							Bd(), o = 6;
							break a;
						case 3:
						case 2:
						case 9:
						case 6:
							Lo.current === null && (t = !0);
							var l = J;
							if (J = 0, nd = null, $d(e, s, c, l), n && id) {
								o = 0;
								break a;
							}
							break;
						default: l = J, J = 0, nd = null, $d(e, s, c, l);
					}
				}
				Jd(), o = Y;
				break;
			} catch (t) {
				Hd(e, t);
			}
		while (1);
		return t && e.shellSuspendCounter++, Ca = Sa = null, W = r, O.H = i, O.A = a, K === null && (G = null, q = 0, ji()), o;
	}
	function Jd() {
		for (; K !== null;) Zd(K);
	}
	function Yd(e, t) {
		var n = W;
		W |= 2;
		var r = Wd(), i = Gd();
		G !== e || q !== t ? (vd = null, _d = qe() + 500, Vd(e, t)) : id = mt(e, t);
		a: do
			try {
				if (J !== 0 && K !== null) {
					t = K;
					var a = nd;
					b: switch (J) {
						case 1:
							J = 0, nd = null, $d(e, t, a, 1);
							break;
						case 2:
						case 9:
							if (ao(a)) {
								J = 0, nd = null, Qd(t);
								break;
							}
							t = function() {
								J !== 2 && J !== 9 || G !== e || (J = 7), Ef(e);
							}, a.then(t, t);
							break a;
						case 3:
							J = 7;
							break a;
						case 4:
							J = 5;
							break a;
						case 7:
							ao(a) ? (J = 0, nd = null, Qd(t)) : (J = 0, nd = null, $d(e, t, a, 7));
							break;
						case 5:
							var s = null;
							switch (K.tag) {
								case 26: s = K.memoizedState;
								case 5:
								case 27:
									var c = K;
									if (s ? Ym(s) : c.stateNode.complete) {
										J = 0, nd = null;
										var l = c.sibling;
										if (l !== null) K = l;
										else {
											var u = c.return;
											u === null ? K = null : (K = u, ef(u));
										}
										break b;
									}
							}
							J = 0, nd = null, $d(e, t, a, 5);
							break;
						case 6:
							J = 0, nd = null, $d(e, t, a, 6);
							break;
						case 8:
							Bd(), Y = 6;
							break a;
						default: throw Error(o(462));
					}
				}
				Xd();
				break;
			} catch (t) {
				Hd(e, t);
			}
		while (1);
		return Ca = Sa = null, O.H = r, O.A = i, W = n, K === null ? (G = null, q = 0, ji(), Y) : 0;
	}
	function Xd() {
		for (; K !== null && !Ge();) Zd(K);
	}
	function Zd(e) {
		var t = hl(e.alternate, e, od);
		e.memoizedProps = e.pendingProps, t === null ? ef(e) : K = t;
	}
	function Qd(e) {
		var t = e, n = t.alternate;
		switch (t.tag) {
			case 15:
			case 0:
				t = Yc(n, t, t.pendingProps, t.type, void 0, q);
				break;
			case 11:
				t = Yc(n, t, t.pendingProps, t.type.render, t.ref, q);
				break;
			case 5:
				cs(t);
				var r = t;
				r === ua && (P ? (ga(r), r.tag === 5 && r.stateNode != null && (N = r.stateNode)) : (ga(r), P = !0));
			default: Cl(n, t), t = K = Hi(t, od), t = hl(n, t, od);
		}
		e.memoizedProps = e.pendingProps, t === null ? ef(e) : K = t;
	}
	function $d(e, t, n, r) {
		Ca = Sa = null, cs(t), fo = null, po = 0;
		var i = t.return;
		try {
			if (Pc(e, i, t, n, q)) {
				Y = 1, kc(e, Yi(n, e.current)), K = null;
				return;
			}
		} catch (t) {
			if (i !== null) throw K = i, t;
			Y = 1, kc(e, Yi(n, e.current)), K = null;
			return;
		}
		t.flags & 32768 ? (P || r === 1 ? e = !0 : id || q & 536870912 ? e = !1 : (rd = e = !0, (r === 2 || r === 9 || r === 3 || r === 6) && (r = Lo.current, r !== null && r.tag === 13 && (r.flags |= 16384))), tf(t, e)) : ef(t);
	}
	function ef(e) {
		var t = e;
		do {
			if (t.flags & 32768) {
				tf(t, rd);
				return;
			}
			e = t.return;
			var n = xl(t.alternate, t, od);
			if (n !== null) {
				K = n;
				return;
			}
			if (t = t.sibling, t !== null) {
				K = t;
				return;
			}
			K = t = e;
		} while (t !== null);
		Y === 0 && (Y = 5);
	}
	function tf(e, t) {
		do {
			var n = Sl(e.alternate, e);
			if (n !== null) {
				n.flags &= 32767, K = n;
				return;
			}
			if (n = e.return, n !== null && (n.flags |= 32768, n.subtreeFlags = 0, n.deletions = null), !t && (e = e.sibling, e !== null)) {
				K = e;
				return;
			}
			K = e = n;
		} while (e !== null);
		Y = 6, K = null;
	}
	function nf(e, t, n, r, i, a, s, c, l, u, d, f) {
		e.cancelPendingCommit = null;
		do
			df();
		while (X !== 0);
		if (W & 6) throw Error(o(327));
		if (t !== null) {
			if (t === e.current) throw Error(o(177));
			e === G && (K = G = null, q = 0), xd = t, bd = e, Sd = n, wd = i, Td = r, rf(e, t, n, s, c, l, f);
		}
	}
	function rf(e, t, n, r, i, a, o) {
		var s = t.lanes | t.childLanes;
		if (Cd = s, s |= Ai, bt(e, n, s, r, i, a), Dd = null, (n & 335544064) === n ? (Od = Ha(e), r = 10262) : (Od = null, r = 10256), (t.subtreeFlags & r) !== 0 || (t.flags & r) !== 0 ? (e.callbackNode = null, e.callbackPriority = 0, yf(Ze, function() {
			return ff(), null;
		})) : (e.callbackNode = null, e.callbackPriority = 0), Hl = !1, r = !!(t.flags & 13878), t.subtreeFlags & 13878 || r) {
			r = O.T, O.T = null, i = k.p, k.p = 2, a = W, W |= 4;
			try {
				gu(e, t, n);
			} finally {
				W = a, k.p = i, O.T = r;
			}
		}
		X = 1, Hl ? Ed = Mp(o, e.containerInfo, Od, sf, cf, of, lf, ff, af, null, null) : (sf(), cf(), lf());
	}
	function af(e) {
		if (X !== 0) {
			var t = bd.onRecoverableError;
			t(e, { componentStack: null });
		}
	}
	function of() {
		X === 3 && (X = 0, Iu(xd, bd), X = 4);
	}
	function sf() {
		if (X === 1) {
			X = 0;
			var e = bd, t = xd, n = Sd, r = !!(t.flags & 13878);
			if (t.subtreeFlags & 13878 || r) {
				r = O.T, O.T = null;
				var i = k.p;
				k.p = 2;
				var a = W;
				W |= 4;
				try {
					pu = mu = !1, Mu(t, e, n), n = cp;
					var o = ei(e.containerInfo), s = n.focusedElem, c = n.selectionRange;
					if (o !== s && s && s.ownerDocument && $r(s.ownerDocument.documentElement, s)) {
						if (c !== null && ti(s)) {
							var l = c.start, u = c.end;
							if (u === void 0 && (u = l), "selectionStart" in s) s.selectionStart = l, s.selectionEnd = Math.min(u, s.value.length);
							else {
								var d = s.ownerDocument || document, f = d && d.defaultView || window;
								if (f.getSelection) {
									var p = f.getSelection(), m = s.textContent.length, h = Math.min(c.start, m), g = c.end === void 0 ? h : Math.min(c.end, m);
									!p.extend && h > g && (o = g, g = h, h = o);
									var _ = Qr(s, h), v = Qr(s, g);
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
					gh = !!sp, cp = sp = null;
				} finally {
					W = a, k.p = i, O.T = r;
				}
			}
			e.current = t, X = 2;
		}
	}
	function cf() {
		if (X === 2) {
			X = 0;
			var e = bd, t = xd, n = !!(t.flags & 8772);
			if (t.subtreeFlags & 8772 || n) {
				n = O.T, O.T = null;
				var r = k.p;
				k.p = 2;
				var i = W;
				W |= 4;
				try {
					vu(e, t.alternate, t);
				} finally {
					W = i, k.p = r, O.T = n;
				}
			}
			X = 3;
		}
	}
	function lf() {
		if (X === 4 || X === 3) {
			X = 0;
			var e = Ed;
			Ed = null, Ke();
			var t = bd, n = xd, r = Sd, i = Td, a = (r & 335544064) === r ? 10262 : 10256;
			if ((n.subtreeFlags & a) !== 0 || (n.flags & a) !== 0 ? X = 5 : (X = 0, xd = bd = null, uf(t, t.pendingLanes)), a = t.pendingLanes, a === 0 && (yd = null), Tt(r), n = n.stateNode, rt && typeof rt.onCommitFiberRoot == "function") try {
				rt.onCommitFiberRoot(nt, n, void 0, (n.current.flags & 128) == 128);
			} catch {}
			if (i !== null) {
				n = O.T, a = k.p, k.p = 2, O.T = null;
				try {
					for (var o = t.onRecoverableError, s = 0; s < i.length; s++) {
						var c = i[s];
						o(c.value, { componentStack: c.stack });
					}
				} finally {
					O.T = n, k.p = a;
				}
			}
			if (i = Dd, o = Od, Od = null, i !== null && (Dd = null, o === null && (o = []), e !== null)) for (c = 0; c < i.length; c++) n = (0, i[c])(o), n !== void 0 && e.finished.finally(n);
			Sd & 3 && df(), Ef(t), a = t.pendingLanes, r & 261930 && a & 42 ? t === Ad ? kd++ : (kd = 0, Ad = t) : (kd = 0, Ad = null), Df(0, !1);
		}
	}
	function uf(e, t) {
		(e.pooledCacheLanes &= t) === 0 && (t = e.pooledCache, t != null && (e.pooledCache = null, za(t)));
	}
	function df() {
		return Ed !== null && (Ed.skipTransition(), Ed = null), sf(), cf(), lf(), ff();
	}
	function ff() {
		if (X !== 5) return !1;
		var e = bd, t = Cd;
		Cd = 0;
		var n = Tt(Sd), r = O.T, i = k.p;
		try {
			k.p = 32 > n ? 32 : n, O.T = null, n = wd, wd = null;
			var a = bd, s = Sd;
			if (X = 0, xd = bd = null, Sd = 0, W & 6) throw Error(o(331));
			var c = W;
			if (W |= 4, Zu(a.current), Uu(a, a.current, s, n), W = c, Df(0, !1), rt && typeof rt.onPostCommitFiberRoot == "function") try {
				rt.onPostCommitFiberRoot(nt, a);
			} catch {}
			return !0;
		} finally {
			k.p = i, O.T = r, uf(e, t);
		}
	}
	function pf(e, t, n) {
		t = Yi(n, t), t = jc(e.stateNode, t, 2), e = wo(e, t, 2), e !== null && (yt(e, 2), Ef(e));
	}
	function Z(e, t, n) {
		if (e.tag === 3) pf(e, e, n);
		else for (; t !== null;) {
			if (t.tag === 3) {
				pf(t, e, n);
				break;
			}
			if (t.tag === 1) {
				var r = t.stateNode;
				if (typeof t.type.getDerivedStateFromError == "function" || typeof r.componentDidCatch == "function" && (yd === null || !yd.has(r))) {
					e = Yi(n, e), n = Mc(2), r = wo(t, n, 2), r !== null && (Nc(n, r, t, e), yt(r, 2), Ef(r));
					break;
				}
			}
			t = t.return;
		}
	}
	function mf(e, t, n) {
		var r = e.pingCache;
		if (r === null) {
			r = e.pingCache = new td();
			var i = /* @__PURE__ */ new Set();
			r.set(t, i);
		} else i = r.get(t), i === void 0 && (i = /* @__PURE__ */ new Set(), r.set(t, i));
		i.has(n) || (ad = !0, i.add(n), e = hf.bind(null, e, t, n), t.then(e, e));
	}
	function hf(e, t, n) {
		var r = e.pingCache;
		r !== null && r.delete(t), e.pingedLanes |= e.suspendedLanes & n, e.warmLanes &= ~n, G === e && (q & n) === n && (Y === 4 || Y === 3 && (q & 62914560) === q && 300 > qe() - hd ? W & 2 ? ld |= n : Vd(e, 0) : ld |= n, dd === q && (dd = 0)), Ef(e);
	}
	function gf(e, t) {
		t === 0 && (t = _t()), e = Pi(e, t), e !== null && (yt(e, t), Ef(e));
	}
	function _f(e) {
		var t = e.memoizedState, n = 0;
		t !== null && (n = t.retryLane), gf(e, n);
	}
	function vf(e, t) {
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
		r !== null && r.delete(t), gf(e, n);
	}
	function yf(e, t) {
		return Ue(e, t);
	}
	var bf = null, xf = null, Sf = !1, Cf = !1, wf = !1, Tf = 0;
	function Ef(e) {
		e !== xf && e.next === null && (xf === null ? bf = xf = e : xf = xf.next = e), Cf = !0, Sf || (Sf = !0, Nf());
	}
	function Df(e, t) {
		if (!wf && Cf) {
			wf = !0;
			do
				for (var n = !1, r = bf; r !== null;) {
					if (!t) {
						if (e !== 0) {
							var i = r.pendingLanes;
							if (i === 0) var a = 0;
							else {
								var o = r.suspendedLanes, s = r.pingedLanes;
								a = (1 << 31 - at(42 | e) + 1) - 1, a &= i & ~(o & ~s), a = a & 201326741 ? a & 201326741 | 1 : a ? a | 2 : 0;
							}
							a !== 0 && (n = !0, Mf(r, a));
						} else a = q, a = pt(r, r === G ? a : 0, r.cancelPendingCommit !== null || r.timeoutHandle !== -1), !(a & 3) || mt(r, a) || (n = !0, Mf(r, a));
					}
					r = r.next;
				}
			while (n);
			wf = !1;
		}
	}
	function Of() {
		kf();
	}
	function kf() {
		Cf = Sf = !1;
		var e = 0;
		Tf !== 0 && hp() && (e = Tf);
		for (var t = qe(), n = null, r = bf; r !== null;) {
			var i = r.next, a = Af(r, t);
			a === 0 ? (r.next = null, n === null ? bf = i : n.next = i, i === null && (xf = n)) : (n = r, (e !== 0 || a & 3) && (Cf = !0)), r = i;
		}
		X !== 0 && X !== 5 || Df(e, !1), Tf !== 0 && (Tf = 0);
	}
	function Af(e, t) {
		for (var n = e.suspendedLanes, r = e.pingedLanes, i = e.expirationTimes, a = e.pendingLanes & -62914561; 0 < a;) {
			var o = 31 - at(a), s = 1 << o, c = i[o];
			c === -1 ? ((s & n) === 0 || (s & r) !== 0) && (i[o] = gt(s, t)) : c <= t && (e.expiredLanes |= s), a &= ~s;
		}
		if (t = G, n = q, n = pt(e, e === t ? n : 0, e.cancelPendingCommit !== null || e.timeoutHandle !== -1), r = e.callbackNode, n === 0 || e === t && (J === 2 || J === 9) || e.cancelPendingCommit !== null) return r !== null && r !== null && We(r), e.callbackNode = null, e.callbackPriority = 0;
		if (!(n & 3) || mt(e, n)) {
			if (t = n & -n, t === e.callbackPriority) return t;
			switch (r !== null && We(r), Tt(n)) {
				case 2:
				case 8:
					n = Xe;
					break;
				case 32:
					n = Ze;
					break;
				case 268435456:
					n = $e;
					break;
				default: n = Ze;
			}
			return r = jf.bind(null, e), n = Ue(n, r), e.callbackPriority = t, e.callbackNode = n, t;
		}
		return r !== null && r !== null && We(r), e.callbackPriority = 2, e.callbackNode = null, 2;
	}
	function jf(e, t) {
		if (X !== 0 && X !== 5) return e.callbackNode = null, e.callbackPriority = 0, null;
		var n = e.callbackNode;
		if (df() && e.callbackNode !== n) return null;
		var r = q;
		return r = pt(e, e === G ? r : 0, e.cancelPendingCommit !== null || e.timeoutHandle !== -1), r === 0 ? null : (Fd(e, r, t), Af(e, qe()), e.callbackNode != null && e.callbackNode === n ? jf.bind(null, e) : null);
	}
	function Mf(e, t) {
		if (df()) return null;
		Fd(e, t, !0);
	}
	function Nf() {
		bp(function() {
			W & 6 ? Ue(Ye, Of) : kf();
		});
	}
	function Pf() {
		if (Tf === 0) {
			var e = Ga;
			e === 0 && (e = lt, lt <<= 1, !(lt & 261888) && (lt = 256)), Tf = e;
		}
		return Tf;
	}
	function Ff(e) {
		return e == null || typeof e == "symbol" || typeof e == "boolean" ? null : typeof e == "function" ? e : wn(e);
	}
	function If(e, t, n, r, i) {
		if (t === "submit" && n && n.stateNode === i) {
			var a = Ff((i[At] || null).action), o = r.submitter;
			o && (t = (t = o[At] || null) ? Ff(t.formAction) : o.getAttribute("formAction"), t !== null && (a = t, o = null));
			var s = new Kn("action", "action", null, r, i);
			e.push({
				event: s,
				listeners: [{
					instance: null,
					listener: function() {
						if (r.defaultPrevented) {
							if (Tf !== 0) {
								var e = new FormData(i, o);
								rc(n, {
									pending: !0,
									data: e,
									method: i.method,
									action: a
								}, null, e);
							}
						} else typeof a == "function" && (s.preventDefault(), e = new FormData(i, o), rc(n, {
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
	for (var Lf = 0; Lf < xi.length; Lf++) {
		var Rf = xi[Lf];
		Si(Rf.toLowerCase(), "on" + (Rf[0].toUpperCase() + Rf.slice(1)));
	}
	Si(pi, "onAnimationEnd"), Si(mi, "onAnimationIteration"), Si(hi, "onAnimationStart"), Si("dblclick", "onDoubleClick"), Si("focusin", "onFocus"), Si("focusout", "onBlur"), Si(gi, "onTransitionRun"), Si(_i, "onTransitionStart"), Si(vi, "onTransitionCancel"), Si(yi, "onTransitionEnd"), Jt("onMouseEnter", ["mouseout", "mouseover"]), Jt("onMouseLeave", ["mouseout", "mouseover"]), Jt("onPointerEnter", ["pointerout", "pointerover"]), Jt("onPointerLeave", ["pointerout", "pointerover"]), qt("onChange", "change click focusin focusout input keydown keyup selectionchange".split(" ")), qt("onSelect", "focusout contextmenu dragend focusin keydown keyup mousedown mouseup selectionchange".split(" ")), qt("onBeforeInput", [
		"compositionend",
		"keypress",
		"textInput",
		"paste"
	]), qt("onCompositionEnd", "compositionend focusout keydown keypress keyup mousedown".split(" ")), qt("onCompositionStart", "compositionstart focusout keydown keypress keyup mousedown".split(" ")), qt("onCompositionUpdate", "compositionupdate focusout keydown keypress keyup mousedown".split(" "));
	var zf = "abort canplay canplaythrough durationchange emptied encrypted ended error loadeddata loadedmetadata loadstart pause play playing progress ratechange resize seeked seeking stalled suspend timeupdate volumechange waiting".split(" "), Bf = new Set("beforetoggle cancel close invalid load scroll scrollend toggle".split(" ").concat(zf));
	function Vf(e, t) {
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
						Di(e);
					}
					i.currentTarget = null, a = c;
				}
				else for (o = 0; o < r.length; o++) {
					if (s = r[o], c = s.instance, l = s.currentTarget, s = s.listener, c !== a && i.isPropagationStopped()) break a;
					a = s, i.currentTarget = l;
					try {
						a(i);
					} catch (e) {
						Di(e);
					}
					i.currentTarget = null, a = c;
				}
			}
		}
	}
	function Q(e, t) {
		var n = t[Mt];
		n === void 0 && (n = t[Mt] = /* @__PURE__ */ new Set());
		var r = e + "__bubble";
		n.has(r) || (Gf(t, e, 2, !1), n.add(r));
	}
	function Hf(e, t, n) {
		var r = 0;
		t && (r |= 4), Gf(n, e, r, t);
	}
	var Uf = "_reactListening" + Math.random().toString(36).slice(2);
	function Wf(e) {
		if (!e[Uf]) {
			e[Uf] = !0, Gt.forEach(function(t) {
				t !== "selectionchange" && (Bf.has(t) || Hf(t, !1, e), Hf(t, !0, e));
			});
			var t = e.nodeType === 9 ? e : e.ownerDocument;
			t === null || t[Uf] || (t[Uf] = !0, Hf("selectionchange", !1, t));
		}
	}
	function Gf(e, t, n, r) {
		switch (Ch(t)) {
			case 2:
				var i = _h;
				break;
			case 8:
				i = vh;
				break;
			default: i = yh;
		}
		n = i.bind(null, t, n, e), i = void 0, !Fn || t !== "touchstart" && t !== "touchmove" && t !== "wheel" || (i = !0), r ? i === void 0 ? e.addEventListener(t, n, !0) : e.addEventListener(t, n, {
			capture: !0,
			passive: i
		}) : i === void 0 ? e.addEventListener(t, n, !1) : e.addEventListener(t, n, { passive: i });
	}
	function Kf(e, t, n, r, i) {
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
					if (o = zt(s), o === null) return;
					if (l = o.tag, l === 5 || l === 6 || l === 26 || l === 27) {
						r = a = o;
						continue a;
					}
					s = s.parentNode;
				}
			}
			r = r.return;
		}
		Mn(function() {
			var r = a, i = Dn(n), o = [];
			a: {
				var s = bi.get(e);
				if (s !== void 0) {
					var l = Kn, u = e;
					switch (e) {
						case "keypress": if (Vn(n) === 0) break a;
						case "keydown":
						case "keyup":
							l = ur;
							break;
						case "focusin":
							u = "focus", l = tr;
							break;
						case "focusout":
							u = "blur", l = tr;
							break;
						case "beforeblur":
						case "afterblur":
							l = tr;
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
							l = $n;
							break;
						case "drag":
						case "dragend":
						case "dragenter":
						case "dragexit":
						case "dragleave":
						case "dragover":
						case "dragstart":
						case "drop":
							l = er;
							break;
						case "touchcancel":
						case "touchend":
						case "touchmove":
						case "touchstart":
							l = pr;
							break;
						case pi:
						case mi:
						case hi:
							l = nr;
							break;
						case yi:
							l = mr;
							break;
						case "scroll":
						case "scrollend":
							l = Jn;
							break;
						case "wheel":
							l = hr;
							break;
						case "copy":
						case "cut":
						case "paste":
							l = rr;
							break;
						case "gotpointercapture":
						case "lostpointercapture":
						case "pointercancel":
						case "pointerdown":
						case "pointermove":
						case "pointerout":
						case "pointerover":
						case "pointerup":
							l = dr;
							break;
						case "submit":
							l = fr;
							break;
						case "toggle":
						case "beforetoggle": l = gr;
					}
					var d = !!(t & 4), f = !d && (e === "scroll" || e === "scrollend"), p = d ? s === null ? null : s + "Capture" : s;
					d = [];
					for (var m = r, h; m !== null;) {
						var g = m;
						if (h = g.stateNode, g = g.tag, g !== 5 && g !== 26 && g !== 27 || h === null || p === null || (g = Nn(m, p), g != null && d.push(qf(m, g, h))), f) break;
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
					if (l = e === "mouseover" || e === "pointerover", s = e === "mouseout" || e === "pointerout", l && n !== En && (u = n.relatedTarget || n.fromElement) && (zt(u) || u[jt])) break a;
					(s || l) && (u = i.window === i ? i : (l = i.ownerDocument) ? l.defaultView || l.parentWindow : window, s ? (l = n.relatedTarget || n.toElement, s = r, l = l ? zt(l) : null, l !== null && (f = c(l), d = l.tag, l !== f || d !== 5 && d !== 27 && d !== 6) && (l = null)) : (s = null, l = r), s !== l && (d = $n, g = "onMouseLeave", p = "onMouseEnter", m = "mouse", (e === "pointerout" || e === "pointerover") && (d = dr, g = "onPointerLeave", p = "onPointerEnter", m = "pointer"), f = s == null ? u : Vt(s), h = l == null ? u : Vt(l), u = new d(g, m + "leave", s, n, i), u.target = f, u.relatedTarget = h, g = null, zt(i) === r && (d = new d(p, m + "enter", l, n, i), d.target = h, d.relatedTarget = f, g = d), f = g, d = s && l ? re(s, l, Yf) : null, s !== null && Xf(o, u, s, d, !1), l !== null && f !== null && Xf(o, f, l, d, !0)));
				}
				a: {
					if (s = r ? Vt(r) : window, l = s.nodeName && s.nodeName.toLowerCase(), l === "select" || l === "input" && s.type === "file") var _ = Ir;
					else if (Ar(s)) {
						if (Lr) _ = Kr;
						else {
							_ = Wr;
							var v = Ur;
						}
					} else l = s.nodeName, !l || l.toLowerCase() !== "input" || s.type !== "checkbox" && s.type !== "radio" ? r && xn(r.elementType) && (_ = Ir) : _ = Gr;
					if (_ &&= _(e, r)) {
						jr(o, _, n, i);
						break a;
					}
					v && v(e, s, r);
				}
				switch (v = r ? Vt(r) : window, e) {
					case "focusin":
						(Ar(v) || v.contentEditable === "true") && (ri = v, ii = r, ai = null);
						break;
					case "focusout":
						ai = ii = ri = null;
						break;
					case "mousedown":
						oi = !0;
						break;
					case "contextmenu":
					case "mouseup":
					case "dragend":
						oi = !1, si(o, n, i);
						break;
					case "selectionchange": if (ni) break;
					case "keydown":
					case "keyup": si(o, n, i);
				}
				var y;
				if (vr) b: {
					switch (e) {
						case "compositionstart":
							var b = "onCompositionStart";
							break b;
						case "compositionend":
							b = "onCompositionEnd";
							break b;
						case "compositionupdate":
							b = "onCompositionUpdate";
							break b;
					}
					b = void 0;
				}
				else Er ? wr(e, n) && (b = "onCompositionEnd") : e === "keydown" && n.keyCode === 229 && (b = "onCompositionStart");
				b && (xr && n.locale !== "ko" && (Er || b !== "onCompositionStart" ? b === "onCompositionEnd" && Er && (y = Bn()) : (Ln = i, Rn = "value" in Ln ? Ln.value : Ln.textContent, Er = !0)), v = Jf(r, b), 0 < v.length && (b = new ir(b, e, null, n, i), o.push({
					event: b,
					listeners: v
				}), y ? b.data = y : (y = Tr(n), y !== null && (b.data = y)))), (y = br ? Dr(e, n) : Or(e, n)) && (b = Jf(r, "onBeforeInput"), 0 < b.length && (v = new ir("onBeforeInput", "beforeinput", null, n, i), o.push({
					event: v,
					listeners: b
				}), v.data = y)), If(o, e, r, n, i);
			}
			Vf(o, t);
		});
	}
	function qf(e, t, n) {
		return {
			instance: e,
			listener: t,
			currentTarget: n
		};
	}
	function Jf(e, t) {
		for (var n = t + "Capture", r = []; e !== null;) {
			var i = e, a = i.stateNode;
			if (i = i.tag, i !== 5 && i !== 26 && i !== 27 || a === null || (i = Nn(e, n), i != null && r.unshift(qf(e, i, a)), i = Nn(e, t), i != null && r.push(qf(e, i, a))), e.tag === 3) return r;
			e = e.return;
		}
		return [];
	}
	function Yf(e) {
		if (e === null) return null;
		do
			e = e.return;
		while (e && e.tag !== 5 && e.tag !== 27);
		return e || null;
	}
	function Xf(e, t, n, r, i) {
		for (var a = t._reactName, o = []; n !== null && n !== r;) {
			var s = n, c = s.alternate, l = s.stateNode;
			if (s = s.tag, c !== null && c === r) break;
			s !== 5 && s !== 26 && s !== 27 || l === null || (c = l, i ? (l = Nn(n, a), l != null && o.unshift(qf(n, l, c))) : i || (l = Nn(n, a), l != null && o.push(qf(n, l, c)))), n = n.return;
		}
		o.length !== 0 && e.push({
			event: t,
			listeners: o
		});
	}
	var Zf = /\r\n?/g, Qf = /\u0000|\uFFFD/g;
	function $f(e) {
		return (typeof e == "string" ? e : "" + e).replace(Zf, "\n").replace(Qf, "");
	}
	function ep(e, t) {
		return t = $f(t), $f(e) === t;
	}
	function $(e, t, n, r, i, a) {
		switch (n) {
			case "children":
				if (typeof r == "string") t === "body" || t === "textarea" && r === "" || _n(e, r);
				else if (typeof r == "number" || typeof r == "bigint") t !== "body" && _n(e, "" + r);
				else return;
				break;
			case "className":
				tn(e, "class", r);
				break;
			case "tabIndex":
				tn(e, "tabindex", r);
				break;
			case "dir":
			case "role":
			case "viewBox":
			case "width":
			case "height":
				tn(e, n, r);
				break;
			case "style":
				bn(e, r, a);
				return;
			case "data": if (t !== "object") {
				tn(e, "data", r);
				break;
			}
			case "src":
			case "href":
				if (r === "" && (t !== "a" || n !== "href") || r == null || typeof r == "function" || typeof r == "symbol" || typeof r == "boolean") {
					e.removeAttribute(n);
					break;
				}
				r = wn(r), e.setAttribute(n, r);
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
				r = wn(r), e.setAttribute(n, r);
				break;
			case "onClick":
				r != null && (e.onclick = Tn);
				return;
			case "onScroll":
				r != null && Q("scroll", e);
				return;
			case "onScrollEnd":
				r != null && Q("scrollend", e);
				return;
			case "dangerouslySetInnerHTML":
				if (r != null) {
					if (typeof r != "object" || !("__html" in r)) throw Error(o(61));
					if (n = r.__html, n != null) {
						if (i.children != null) throw Error(o(60));
						a?.__html !== n && (e.innerHTML = n);
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
				n = wn(r), e.setAttributeNS("http://www.w3.org/1999/xlink", "xlink:href", n);
				break;
			case "contentEditable":
			case "spellCheck":
			case "draggable":
			case "value":
			case "autoReverse":
			case "externalResourcesRequired":
			case "focusable":
			case "preserveAlpha":
				r != null && typeof r != "function" && typeof r != "symbol" ? e.setAttribute(n, r) : e.removeAttribute(n);
				break;
			case "inert":
			case "allowFullScreen":
			case "async":
			case "autoPlay":
			case "controls":
			case "credentialless":
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
				Q("beforetoggle", e), Q("toggle", e), en(e, "popover", r);
				break;
			case "xlinkActuate":
				nn(e, "http://www.w3.org/1999/xlink", "xlink:actuate", r);
				break;
			case "xlinkArcrole":
				nn(e, "http://www.w3.org/1999/xlink", "xlink:arcrole", r);
				break;
			case "xlinkRole":
				nn(e, "http://www.w3.org/1999/xlink", "xlink:role", r);
				break;
			case "xlinkShow":
				nn(e, "http://www.w3.org/1999/xlink", "xlink:show", r);
				break;
			case "xlinkTitle":
				nn(e, "http://www.w3.org/1999/xlink", "xlink:title", r);
				break;
			case "xlinkType":
				nn(e, "http://www.w3.org/1999/xlink", "xlink:type", r);
				break;
			case "xmlBase":
				nn(e, "http://www.w3.org/XML/1998/namespace", "xml:base", r);
				break;
			case "xmlLang":
				nn(e, "http://www.w3.org/XML/1998/namespace", "xml:lang", r);
				break;
			case "xmlSpace":
				nn(e, "http://www.w3.org/XML/1998/namespace", "xml:space", r);
				break;
			case "is":
				en(e, "is", r);
				break;
			case "innerText":
			case "textContent": return;
			default: if (!(2 < n.length) || n[0] !== "o" && n[0] !== "O" || n[1] !== "n" && n[1] !== "N") n = Sn.get(n) || n, en(e, n, r);
			else return;
		}
		M = !0;
	}
	function tp(e, t, n, r, i, a) {
		switch (n) {
			case "style":
				bn(e, r, a);
				return;
			case "dangerouslySetInnerHTML":
				if (r != null) {
					if (typeof r != "object" || !("__html" in r)) throw Error(o(61));
					if (n = r.__html, n != null) {
						if (i.children != null) throw Error(o(60));
						a?.__html !== n && (e.innerHTML = n);
					}
				}
				break;
			case "children":
				if (typeof r == "string") _n(e, r);
				else if (typeof r == "number" || typeof r == "bigint") _n(e, "" + r);
				else return;
				break;
			case "onScroll":
				r != null && Q("scroll", e);
				return;
			case "onScrollEnd":
				r != null && Q("scrollend", e);
				return;
			case "onClick":
				r != null && (e.onclick = Tn);
				return;
			case "suppressContentEditableWarning":
			case "suppressHydrationWarning":
			case "innerHTML":
			case "ref": return;
			case "innerText":
			case "textContent": return;
			default:
				if (!Kt.hasOwnProperty(n)) a: {
					if (n[0] === "o" && n[1] === "n" && (i = n.endsWith("Capture"), a = n.slice(2, i ? n.length - 7 : void 0), t = e[At] || null, t = t == null ? null : t[n], typeof t == "function" && e.removeEventListener(a, t, i), typeof r == "function")) {
						typeof t != "function" && t !== null && (n in e ? e[n] = null : e.hasAttribute(n) && e.removeAttribute(n)), e.addEventListener(a, r, i);
						break a;
					}
					M = !0, n in e ? e[n] = r : !0 === r ? e.setAttribute(n, "") : en(e, n, r);
				}
				return;
		}
		M = !0;
	}
	function np(e, t, n) {
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
				fn(e, a, c, l, u, s, i, !1);
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
				t = a, n = s, e.multiple = !!r, t == null ? n != null && mn(e, !!r, n, !0) : mn(e, !!r, t, !1);
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
				gn(e, r, i, a);
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
				for (r = 0; r < zf.length; r++) Q(zf[r], e);
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
			default: if (xn(t)) {
				for (d in n) n.hasOwnProperty(d) && (r = n[d], r !== void 0 && tp(e, t, d, r, n, void 0));
				return;
			}
		}
		for (c in n) n.hasOwnProperty(c) && (r = n[c], r != null && $(e, t, c, r, n, null));
	}
	var rp = {};
	function ip(e, t, n, r) {
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
							m !== f && (M = !0), a = m;
							break;
						case "name":
							m !== f && (M = !0), i = m;
							break;
						case "checked":
							m !== f && (M = !0), u = m;
							break;
						case "defaultChecked":
							m !== f && (M = !0), d = m;
							break;
						case "value":
							m !== f && (M = !0), s = m;
							break;
						case "defaultValue":
							m !== f && (M = !0), c = m;
							break;
						case "children":
						case "dangerouslySetInnerHTML":
							if (m != null) throw Error(o(137, t));
							break;
						default: m !== f && $(e, t, p, m, r, f);
					}
				}
				dn(e, s, c, l, u, d, a, i);
				return;
			case "select":
				for (a in m = s = c = p = null, n) if (l = n[a], n.hasOwnProperty(a) && l != null) switch (a) {
					case "value": break;
					case "multiple": m = l;
					default: r.hasOwnProperty(a) || $(e, t, a, null, r, l);
				}
				for (i in r) if (a = r[i], l = n[i], r.hasOwnProperty(i) && (a != null || l != null)) switch (i) {
					case "value":
						a !== l && (M = !0), p = a;
						break;
					case "defaultValue":
						a !== l && (M = !0), c = a;
						break;
					case "multiple": a !== l && (M = !0), s = a;
					default: a !== l && $(e, t, i, a, r, l);
				}
				t = c, n = s, r = m, p == null ? !!r != !!n && (t == null ? mn(e, !!n, n ? [] : "", !1) : mn(e, !!n, t, !0)) : mn(e, !!n, p, !1);
				return;
			case "textarea":
				for (c in m = p = null, n) if (i = n[c], n.hasOwnProperty(c) && i != null && !r.hasOwnProperty(c)) switch (c) {
					case "value": break;
					case "children": break;
					default: $(e, t, c, null, r, i);
				}
				for (s in r) if (i = r[s], a = n[s], r.hasOwnProperty(s) && (i != null || a != null)) switch (s) {
					case "value":
						i !== a && (M = !0), p = i;
						break;
					case "defaultValue":
						i !== a && (M = !0), m = i;
						break;
					case "children": break;
					case "dangerouslySetInnerHTML":
						if (i != null) throw Error(o(91));
						break;
					default: i !== a && $(e, t, s, i, r, a);
				}
				hn(e, p, m);
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
						p !== m && (M = !0), e.selected = p && typeof p != "function" && typeof p != "symbol";
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
			default: if (xn(t)) {
				for (var _ in n) p = n[_], n.hasOwnProperty(_) && p !== void 0 && !r.hasOwnProperty(_) && tp(e, t, _, void 0, r, p);
				for (d in r) p = r[d], m = n[d], !r.hasOwnProperty(d) || p === m || p === void 0 && m === void 0 || tp(e, t, d, p, r, m);
				return;
			}
		}
		for (var v in n) p = n[v], n.hasOwnProperty(v) && p != null && !r.hasOwnProperty(v) && $(e, t, v, null, r, p);
		for (f in r) p = r[f], m = n[f], !r.hasOwnProperty(f) || p === m || p == null && m == null || $(e, t, f, p, r, m);
	}
	function ap(e) {
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
	function op() {
		if (typeof performance.getEntriesByType == "function") {
			for (var e = 0, t = 0, n = performance.getEntriesByType("resource"), r = 0; r < n.length; r++) {
				var i = n[r], a = i.transferSize, o = i.initiatorType, s = i.duration;
				if (a && s && ap(o)) {
					for (o = 0, s = i.responseEnd, r += 1; r < n.length; r++) {
						var c = n[r], l = c.startTime;
						if (l > s) break;
						var u = c.transferSize, d = c.initiatorType;
						u && ap(d) && (c = c.responseEnd, o += u * (c < s ? 1 : (s - l) / (c - l)));
					}
					if (--r, t += 8 * (a + o) / (i.duration / 1e3), e++, 10 < e) break;
				}
			}
			if (0 < e) return t / e / 1e6;
		}
		return navigator.connection && (e = navigator.connection.downlink, typeof e == "number") ? e : 5;
	}
	var sp = null, cp = null;
	function lp(e) {
		return e.nodeType === 9 ? e : e.ownerDocument;
	}
	function up(e) {
		switch (e) {
			case "http://www.w3.org/2000/svg": return 1;
			case "http://www.w3.org/1998/Math/MathML": return 2;
			default: return 0;
		}
	}
	function dp(e, t) {
		if (e === 0) switch (t) {
			case "svg": return 1;
			case "math": return 2;
			default: return 0;
		}
		return e === 1 && t === "foreignObject" ? 0 : e;
	}
	function fp(e, t, n, r) {
		return n = lp(n).createElement(e), n[kt] = r, n[At] = t, np(n, e, t), Ut(n), n;
	}
	function pp(e, t) {
		return e === "textarea" || e === "noscript" || typeof t.children == "string" || typeof t.children == "number" || typeof t.children == "bigint" || typeof t.dangerouslySetInnerHTML == "object" && t.dangerouslySetInnerHTML !== null && t.dangerouslySetInnerHTML.__html != null;
	}
	var mp = null;
	function hp() {
		var e = window.event;
		return e && e.type === "popstate" ? e !== mp && (mp = e, !0) : (mp = null, !1);
	}
	var gp = typeof setTimeout == "function" ? setTimeout : void 0, _p = typeof clearTimeout == "function" ? clearTimeout : void 0, vp = typeof Promise == "function" ? Promise : void 0, yp = typeof requestAnimationFrame == "function" ? requestAnimationFrame : gp, bp = typeof queueMicrotask == "function" ? queueMicrotask : vp === void 0 ? gp : function(e) {
		return vp.resolve(null).then(e).catch(xp);
	};
	function xp(e) {
		setTimeout(function() {
			throw e;
		});
	}
	function Sp(e) {
		return e === "head";
	}
	function Cp(e, t) {
		var n = t, r = 0;
		do {
			var i = n.nextSibling;
			if (e.removeChild(n), i && i.nodeType === 8) {
				if (n = i.data, n === "/$" || n === "/&") {
					if (r === 0) {
						e.removeChild(i), Hh(t);
						return;
					}
					r--;
				} else if (n === "$" || n === "$?" || n === "$~" || n === "$!" || n === "&") r++;
				else if (n === "html") _m(e.ownerDocument.documentElement);
				else if (n === "head") {
					n = e.ownerDocument.head, _m(n);
					for (var a = n.firstChild; a;) {
						var o = a.nextSibling, s = a.nodeName;
						a[It] || s === "SCRIPT" || s === "STYLE" || s === "LINK" && a.rel.toLowerCase() === "stylesheet" || n.removeChild(a), a = o;
					}
				} else n === "body" && _m(e.ownerDocument.body);
			}
			n = i;
		} while (n);
		Hh(t);
	}
	function wp(e, t) {
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
	function Tp(e, t, n) {
		if (t = CSS.escape(t) === t ? t : "r-" + btoa(t).replace(/=/g, ""), e.style.viewTransitionName = t, n != null && (e.style.viewTransitionClass = n), n = getComputedStyle(e), n.display === "inline") {
			if (t = e.getClientRects(), t.length === 1) var r = 1;
			else for (var i = r = 0; i < t.length; i++) {
				var a = t[i];
				0 < a.width && 0 < a.height && r++;
			}
			r === 1 && (e = e.style, e.display = t.length === 1 ? "inline-block" : "block", e.marginTop = "-" + n.paddingTop, e.marginBottom = "-" + n.paddingBottom);
		}
	}
	function Ep(e, t) {
		e = e.style, t = t.style;
		var n = t == null ? null : t.hasOwnProperty("viewTransitionName") ? t.viewTransitionName : t.hasOwnProperty("view-transition-name") ? t["view-transition-name"] : null;
		e.viewTransitionName = n == null || typeof n == "boolean" ? "" : ("" + n).trim(), n = t == null ? null : t.hasOwnProperty("viewTransitionClass") ? t.viewTransitionClass : t.hasOwnProperty("view-transition-class") ? t["view-transition-class"] : null, e.viewTransitionClass = n == null || typeof n == "boolean" ? "" : ("" + n).trim(), e.display === "inline-block" && (t == null ? e.display = e.margin = "" : (n = t.display, e.display = n == null || typeof n == "boolean" ? "" : n, n = t.margin, n == null ? (n = t.hasOwnProperty("marginTop") ? t.marginTop : t["margin-top"], e.marginTop = n == null || typeof n == "boolean" ? "" : n, t = t.hasOwnProperty("marginBottom") ? t.marginBottom : t["margin-bottom"], e.marginBottom = t == null || typeof t == "boolean" ? "" : t) : e.margin = n));
	}
	function Dp(e, t, n) {
		return n = n.ownerDocument.defaultView, {
			rect: e,
			abs: t.position === "absolute" || t.position === "fixed",
			clip: t.clipPath !== "none" || t.overflow !== "visible" || t.filter !== "none" || t.mask !== "none" || t.mask !== "none" || t.borderRadius !== "0px",
			view: 0 <= e.bottom && 0 <= e.right && e.top <= n.innerHeight && e.left <= n.innerWidth
		};
	}
	function Op(e) {
		return Dp(e.getBoundingClientRect(), getComputedStyle(e), e);
	}
	function kp(e) {
		var t = e.getBoundingClientRect();
		t = new DOMRect(t.x + 2e4, t.y + 2e4, t.width, t.height);
		var n = getComputedStyle(e);
		return Dp(t, n, e);
	}
	function Ap(e) {
		return e.documentElement.clientHeight;
	}
	function jp(e) {
		this.addEventListener("load", e), this.addEventListener("error", e);
	}
	function Mp(e, t, n, r, i, a, o, s, c) {
		var l = t.nodeType === 9 ? t : t.ownerDocument;
		try {
			var u = l.startViewTransition({
				update: function() {
					var t = l.defaultView, n = t.navigation && t.navigation.transition, o = l.fonts.status;
					r();
					var s = [];
					if (o === "loaded" && (Ap(l), l.fonts.status === "loading" && s.push(l.fonts.ready)), o = s.length, e !== null) for (var c = e.suspenseyImages, u = 0, d = 0; d < c.length; d++) {
						var f = c[d];
						if (!f.complete) {
							var p = f.getBoundingClientRect();
							if (0 < p.bottom && 0 < p.right && p.top < t.innerHeight && p.left < t.innerWidth) {
								if (u += Xm(f), u > $m) {
									s.length = o;
									break;
								}
								f = new Promise(jp.bind(f)), s.push(f);
							}
						}
					}
					if (0 < s.length) return t = Promise.race([Promise.all(s), new Promise(function(e) {
						return setTimeout(e, 500);
					})]).then(i, i), (n ? Promise.allSettled([n.finished, t]) : t).then(a, a);
					if (i(), n) return n.finished.then(a, a);
					a();
				},
				types: n
			});
			l.__reactViewTransition = u;
			var d = [];
			return u.ready.then(function() {
				for (var e = l.documentElement.getAnimations({ subtree: !0 }), t = 0; t < e.length; t++) {
					var n = e[t], r = n.effect, i = r.pseudoElement;
					if (i != null && i.startsWith("::view-transition")) {
						d.push(n), n = r.getKeyframes();
						for (var a = i = void 0, s = !0, c = 0; c < n.length; c++) {
							var u = n[c], f = u.width;
							if (i === void 0) i = f;
							else if (i !== f) {
								s = !1;
								break;
							}
							if (f = u.height, a === void 0) a = f;
							else if (a !== f) {
								s = !1;
								break;
							}
							delete u.width, delete u.height, u.transform === "none" && delete u.transform;
						}
						s && i !== void 0 && a !== void 0 && (r.setKeyframes(n), s = getComputedStyle(r.target, r.pseudoElement), s.width !== i || s.height !== a) && (s = n[0], s.width = i, s.height = a, s = n[n.length - 1], s.width = i, s.height = a, r.setKeyframes(n));
					}
				}
				o();
			}, function(e) {
				l.__reactViewTransition === u && (l.__reactViewTransition = null);
				try {
					if (typeof e == "object" && e) switch (e.name) {
						case "InvalidStateError": (e.message === "View transition was skipped because document visibility state is hidden." || e.message === "Skipping view transition because document visibility state has become hidden." || e.message === "Skipping view transition because viewport size changed." || e.message === "Transition was aborted because of invalid state") && (e = null);
					}
					e !== null && c(e);
				} finally {
					r(), i(), o();
				}
			}), u.finished.finally(function() {
				for (var e = 0; e < d.length; e++) d[e].cancel();
				l.__reactViewTransition === u && (l.__reactViewTransition = null), s();
			}), u;
		} catch {
			return r(), i(), o(), null;
		}
	}
	function Np(e, t) {
		this._scope = document.documentElement, this._selector = "::view-transition-" + e + "(" + t + ")";
	}
	Np.prototype.animate = function(e, t) {
		return t = typeof t == "number" ? { duration: t } : S({}, t), t.pseudoElement = this._selector, this._scope.animate(e, t);
	}, Np.prototype.getAnimations = function() {
		for (var e = this._scope, t = this._selector, n = e.getAnimations({ subtree: !0 }), r = [], i = 0; i < n.length; i++) {
			var a = n[i].effect;
			a !== null && a.target === e && a.pseudoElement === t && r.push(n[i]);
		}
		return r;
	}, Np.prototype.getComputedStyle = function() {
		return getComputedStyle(this._scope, this._selector);
	};
	function Pp(e) {
		return {
			name: e,
			group: new Np("group", e),
			imagePair: new Np("image-pair", e),
			old: new Np("old", e),
			new: new Np("new", e)
		};
	}
	function Fp(e) {
		this._fragmentFiber = e, this._observers = this._eventListeners = null;
	}
	Fp.prototype.addEventListener = function(e, t, n) {
		var r = null, i = null;
		if (!(n != null && typeof n != "boolean" && (r = n.signal || null, r !== null && r.aborted))) {
			this._eventListeners === null && (this._eventListeners = []);
			var a = this._eventListeners;
			if (Bp(a, e, t, n) === -1) {
				var o = this, s = t;
				n != null && typeof n != "boolean" && !0 === n.once && (s = function(r) {
					o.removeEventListener(e, t, n), typeof t == "function" ? t.call(this, r) : t.handleEvent(r);
				}), r !== null && (i = o.removeEventListener.bind(o, e, t, n), r.addEventListener("abort", i, { once: !0 }), i = r.removeEventListener.bind(r, "abort", i)), r = Rp(n), a.push({
					type: e,
					listener: t,
					optionsOrUseCapture: n,
					attachedListener: s,
					cleanup: i
				}), m(this._fragmentFiber.child, !1, Ip, e, s, r);
			}
			this._eventListeners = a;
		}
	};
	function Ip(e, t, n, r) {
		return y(e).addEventListener(t, n, r), !1;
	}
	Fp.prototype.removeEventListener = function(e, t, n) {
		var r = this._eventListeners;
		if (r !== null && (t = Bp(r, e, t, n), t !== -1)) {
			var i = r[t];
			n = i.attachedListener;
			var a = i.cleanup;
			i = Rp(i.optionsOrUseCapture), m(this._fragmentFiber.child, !1, Lp, e, n, i), r.splice(t, 1), a !== null && a();
		}
	};
	function Lp(e, t, n, r) {
		return y(e).removeEventListener(t, n, r), !1;
	}
	function Rp(e) {
		return e != null && typeof e != "boolean" && (!0 === e.once || e.signal instanceof AbortSignal) ? {
			capture: e.capture,
			passive: e.passive
		} : e;
	}
	function zp(e) {
		return e == null ? "c=0" : typeof e == "boolean" ? "c=" + (e ? "1" : "0") : "c=" + (e.capture ? "1" : "0");
	}
	function Bp(e, t, n, r) {
		if (e.length === 0) return -1;
		r = zp(r);
		for (var i = 0; i < e.length; i++) {
			var a = e[i];
			if (a.type === t && a.listener === n && zp(a.optionsOrUseCapture) === r) return i;
		}
		return -1;
	}
	Fp.prototype.dispatchEvent = function(e) {
		var t = h(this._fragmentFiber);
		if (t === null) return !0;
		t = y(t);
		var n = this._eventListeners;
		if (n !== null && 0 < n.length || !e.bubbles) {
			var r = t.nodeType === 9 ? t.createComment("") : document.createTextNode("");
			if (n) for (var i = 0; i < n.length; i++) {
				var a = n[i];
				r.addEventListener(a.type, a.attachedListener, Rp(a.optionsOrUseCapture));
			}
			if (t.appendChild(r), e = r.dispatchEvent(e), n) for (i = 0; i < n.length; i++) a = n[i], r.removeEventListener(a.type, a.attachedListener, Rp(a.optionsOrUseCapture));
			return t.removeChild(r), e;
		}
		return t.dispatchEvent(e);
	}, Fp.prototype.focus = function(e) {
		m(this._fragmentFiber.child, !0, Vp, e, void 0, void 0);
	};
	function Vp(e, t) {
		return e.tag !== 6 && (e = y(e), pm(e, t));
	}
	Fp.prototype.focusLast = function(e) {
		var t = [];
		m(this._fragmentFiber.child, !0, Hp, t, void 0, void 0);
		for (var n = t.length - 1; 0 <= n && !Vp(t[n], e); n--);
	};
	function Hp(e, t) {
		return t.push(e), !1;
	}
	Fp.prototype.blur = function() {
		var e = h(this._fragmentFiber);
		e !== null && (e = y(e), e = lp(e).activeElement, e !== null && m(this._fragmentFiber.child, !1, Up, e, void 0, void 0));
	};
	function Up(e, t) {
		return e.tag !== 6 && (e = y(e), e === t || e.contains(t) ? (t.blur(), !0) : !1);
	}
	Fp.prototype.observeUsing = function(e) {
		this._observers === null && (this._observers = /* @__PURE__ */ new Set()), this._observers.add(e), m(this._fragmentFiber.child, !1, Wp, e, void 0, void 0);
	};
	function Wp(e, t) {
		return e.tag === 6 || (e = y(e), t.observe(e)), !1;
	}
	Fp.prototype.unobserveUsing = function(e) {
		var t = this._observers;
		if (t !== null && t.has(e)) {
			t.delete(e), m(this._fragmentFiber.child, !1, Gp, e, void 0, void 0);
			for (var n = t = 0; n < Kp.length; n++) {
				var r = Kp[n];
				r.fragmentInstance === this && r.observer === e ? e.unobserve(r.instance) : Kp[t++] = r;
			}
			Kp.length = t;
		}
	};
	function Gp(e, t) {
		return e.tag === 6 || (e = y(e), t.unobserve(e)), !1;
	}
	var Kp = [], qp = !1;
	function Jp(e, t, n) {
		Kp.push({
			fragmentInstance: e,
			observer: t,
			instance: n
		}), qp || (qp = !0, mm(function() {
			qp = !1;
			var e = Kp;
			Kp = [];
			for (var t = 0; t < e.length; t++) {
				var n = e[t];
				n.observer.unobserve(n.instance);
			}
		}));
	}
	Fp.prototype.getClientRects = function() {
		var e = [];
		return m(this._fragmentFiber.child, !1, Yp, e, void 0, void 0), e;
	};
	function Yp(e, t) {
		if (e.tag === 6) {
			e = e.stateNode;
			var n = e.ownerDocument.createRange();
			n.selectNodeContents(e), t.push.apply(t, n.getClientRects());
		} else e = y(e), t.push.apply(t, e.getClientRects());
		return !1;
	}
	Fp.prototype.getRootNode = function(e) {
		var t = h(this._fragmentFiber);
		return t === null ? this : y(t).getRootNode(e);
	}, Fp.prototype.compareDocumentPosition = function(e) {
		var t = h(this._fragmentFiber);
		if (t === null) return Node.DOCUMENT_POSITION_DISCONNECTED;
		var n = [];
		m(this._fragmentFiber.child, !1, Hp, n, void 0, void 0);
		var r = y(t);
		if (n.length === 0) {
			if (n = r, g(this._fragmentFiber)) {
				a: {
					for (t = this._fragmentFiber.return; t !== null;) {
						if (t.tag === 4) {
							t = t.stateNode.containerInfo;
							break a;
						}
						if (t.tag === 3 || t.tag === 5 || t.tag === 27) break;
						t = t.return;
					}
					t = null;
				}
				t != null && (n = t);
			}
			t = this._fragmentFiber;
			var i = r = n.compareDocumentPosition(e);
			return n === e ? i = Node.DOCUMENT_POSITION_CONTAINS : r & Node.DOCUMENT_POSITION_CONTAINED_BY && (n = _(t)[1], n === null ? i = Node.DOCUMENT_POSITION_PRECEDING : (e = y(n).compareDocumentPosition(e), i = e === 0 || e & Node.DOCUMENT_POSITION_FOLLOWING ? Node.DOCUMENT_POSITION_FOLLOWING : Node.DOCUMENT_POSITION_PRECEDING)), i |= Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC;
		}
		t = y(n[0]), i = y(n[n.length - 1]);
		var a = g(this._fragmentFiber) ? t.parentElement : r;
		if (a == null) return Node.DOCUMENT_POSITION_DISCONNECTED;
		r = a.compareDocumentPosition(t) & Node.DOCUMENT_POSITION_CONTAINED_BY, a = a.compareDocumentPosition(i) & Node.DOCUMENT_POSITION_CONTAINED_BY;
		var o = t.compareDocumentPosition(e), s = i.compareDocumentPosition(e), c = o & Node.DOCUMENT_POSITION_CONTAINED_BY || s & Node.DOCUMENT_POSITION_CONTAINED_BY;
		return s = r && a && o & Node.DOCUMENT_POSITION_FOLLOWING && s & Node.DOCUMENT_POSITION_PRECEDING, t = r && t === e || a && i === e || c || s ? Node.DOCUMENT_POSITION_CONTAINED_BY : !r && t === e || !a && i === e ? Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC : o, t & Node.DOCUMENT_POSITION_DISCONNECTED || t & Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC || Xp(t, this._fragmentFiber, n[0], n[n.length - 1], e) ? t : Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC;
	};
	function Xp(e, t, n, r, i) {
		var a = zt(i);
		if (e & Node.DOCUMENT_POSITION_CONTAINED_BY) {
			if (n = !!a) a: {
				for (; a !== null;) {
					if (a.tag === 7 && (a === t || a.alternate === t)) {
						n = !0;
						break a;
					}
					a = a.return;
				}
				n = !1;
			}
			return n;
		}
		if (e & Node.DOCUMENT_POSITION_CONTAINS) {
			if (a === null) return a = i.ownerDocument, i === a || i === a.documentElement || i === a.body;
			a: {
				for (a = t, t = h(t); a !== null;) {
					if (!(a.tag !== 5 && a.tag !== 3 && a.tag !== 27 || a !== t && a.alternate !== t)) {
						a = !0;
						break a;
					}
					a = a.return;
				}
				a = !1;
			}
			return a;
		}
		return e & Node.DOCUMENT_POSITION_PRECEDING ? ((t = !!a) && !(t = a === n) && (t = re(n, a, ne), t === null ? t = !1 : (m(t, !0, ee, a, n), a = b, b = null, t = a !== null)), t) : e & Node.DOCUMENT_POSITION_FOLLOWING ? ((t = !!a) && !(t = a === r) && (t = re(r, a, ne), t === null ? t = !1 : (m(t, !0, te, a, r), a = b, x = b = null, t = a !== null)), t) : !1;
	}
	function Zp(e, t) {
		var n = e.ownerDocument.createRange();
		n.selectNodeContents(e), e = n.getBoundingClientRect(), window.scrollTo(window.scrollX + e.left, t ? window.scrollY + e.top : window.scrollY + e.bottom - window.innerHeight);
	}
	Fp.prototype.scrollIntoView = function(e) {
		if (typeof e == "object") throw Error(o(566));
		var t = [];
		m(this._fragmentFiber.child, !1, Hp, t, void 0, void 0);
		var n = !1 !== e;
		if (t.length === 0) {
			var r = _(this._fragmentFiber);
			if (r = n ? r[1] || r[0] || h(this._fragmentFiber) : r[0] || r[1], r === null) return;
			if (r.tag === 6) {
				e = y(r), Zp(e, n);
				return;
			}
			if (r = y(r), r.nodeType !== 9) {
				if (r.nodeType === 11) {
					n = "host" in r ? r.host : null, n !== null && n.scrollIntoView(e);
					return;
				}
				r.scrollIntoView(e);
			}
		}
		for (r = n ? t.length - 1 : 0; r !== (n ? -1 : t.length);) {
			var i = t[r];
			i.tag === 6 ? (i = y(i), Zp(i, n)) : y(i).scrollIntoView(e), r += n ? -1 : 1;
		}
	};
	function Qp(e, t) {
		return e = y(e), $p(e, t), !1;
	}
	function $p(e, t) {
		e.reactFragments ??= /* @__PURE__ */ new Set(), e.reactFragments.add(t);
	}
	function em(e, t) {
		var n = t._eventListeners;
		if (n !== null) for (var r = 0; r < n.length; r++) {
			var i = n[r];
			e.addEventListener(i.type, i.attachedListener, Rp(i.optionsOrUseCapture));
		}
		e.nodeType !== 3 && (n = t._observers, n !== null && n.forEach(function(n) {
			for (var r = 0, i = 0; i < Kp.length; i++) {
				var a = Kp[i];
				(a.fragmentInstance !== t || a.observer !== n || a.instance !== e) && (Kp[r++] = a);
			}
			Kp.length = r, n.observe(e);
		}), $p(e, t));
	}
	function tm(e, t) {
		var n = t._eventListeners;
		if (n !== null) for (var r = 0; r < n.length; r++) {
			var i = n[r];
			e.removeEventListener(i.type, i.attachedListener, Rp(i.optionsOrUseCapture));
		}
		e.nodeType !== 3 && (n = t._observers, n !== null && n.forEach(function(n) {
			typeof n.rootMargin == "string" ? Jp(t, n, e) : n.unobserve(e);
		}), e.reactFragments != null && e.reactFragments.delete(t));
	}
	function nm(e) {
		var t = e.firstChild;
		for (t && t.nodeType === 10 && (t = t.nextSibling); t;) {
			var n = t;
			switch (t = t.nextSibling, n.nodeName) {
				case "HTML":
				case "HEAD":
				case "BODY":
					nm(n), Rt(n);
					continue;
				case "SCRIPT":
				case "STYLE": continue;
				case "LINK": if (n.rel.toLowerCase() === "stylesheet") continue;
			}
			e.removeChild(n);
		}
	}
	function rm(e, t, n, r) {
		for (; e.nodeType === 1;) {
			var i = n;
			if (e.nodeName.toLowerCase() !== t.toLowerCase()) {
				if (!r && (e.nodeName !== "INPUT" || e.type !== "hidden")) break;
			} else if (!r) {
				if (t === "input" && e.type === "hidden") {
					var a = i.name == null ? null : "" + i.name;
					if (i.type === "hidden" && e.getAttribute("name") === a) return e;
				} else return e;
			} else if (!e[It]) switch (t) {
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
			if (e = lm(e.nextSibling), e === null) break;
		}
		return null;
	}
	function im(e, t, n) {
		if (t === "") return null;
		for (; e.nodeType !== 3;) if ((e.nodeType !== 1 || e.nodeName !== "INPUT" || e.type !== "hidden") && !n || (e = lm(e.nextSibling), e === null)) return null;
		return e;
	}
	function am(e, t) {
		for (; e.nodeType !== 8;) if ((e.nodeType !== 1 || e.nodeName !== "INPUT" || e.type !== "hidden") && !t || (e = lm(e.nextSibling), e === null)) return null;
		return e;
	}
	function om(e) {
		return e.data === "$?" || e.data === "$~";
	}
	function sm(e) {
		return e.data === "$!" || e.data === "$?" && e.ownerDocument.readyState !== "loading";
	}
	function cm(e, t) {
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
	function lm(e) {
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
	var um = null;
	function dm(e) {
		e = e.nextSibling;
		for (var t = 0; e;) {
			if (e.nodeType === 8) {
				var n = e.data;
				if (n === "/$" || n === "/&") {
					if (t === 0) return lm(e.nextSibling);
					t--;
				} else n !== "$" && n !== "$!" && n !== "$?" && n !== "$~" && n !== "&" || t++;
			}
			e = e.nextSibling;
		}
		return null;
	}
	function fm(e) {
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
	function pm(e, t) {
		function n() {
			r = !0;
		}
		if (e.ownerDocument.activeElement === e) return !0;
		var r = !1;
		try {
			e.ownerDocument.addEventListener("focus", n, !0), (e.focus || HTMLElement.prototype.focus).call(e, t);
		} finally {
			e.ownerDocument.removeEventListener("focus", n, !0);
		}
		return r;
	}
	function mm(e) {
		yp(function() {
			yp(function(t) {
				return e(t);
			});
		});
	}
	function hm(e, t, n) {
		switch (t = lp(n), e) {
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
	function gm(e, t, n) {
		for (var r in n) {
			var i = n[r];
			n.hasOwnProperty(r) && i != null && $(e, t, r, null, rp, i);
		}
		n.dangerouslySetInnerHTML != null && (e.textContent = ""), e.onclick === Tn && (e.onclick = null), Rt(e);
	}
	function _m(e) {
		for (var t = e.attributes; t.length;) e.removeAttributeNode(t[0]);
		Rt(e);
	}
	var vm = /* @__PURE__ */ new Map(), ym = /* @__PURE__ */ new Set();
	function bm(e) {
		if (typeof e.getRootNode == "function") {
			var t = e.getRootNode();
			if (t.nodeType === 9 || t.nodeType === 11) return t;
		}
		return e.nodeType === 9 ? e : e.ownerDocument;
	}
	var xm = k.d;
	k.d = {
		f: Sm,
		r: Cm,
		D: Em,
		C: Dm,
		L: Om,
		m: km,
		X: jm,
		S: Am,
		M: Mm
	};
	function Sm() {
		var e = xm.f(), t = zd();
		return e || t;
	}
	function Cm(e) {
		var t = Bt(e);
		t !== null && t.tag === 5 && t.type === "form" ? ac(t) : xm.r(e);
	}
	var wm = typeof document > "u" ? null : document;
	function Tm(e, t, n) {
		var r = wm;
		if (r && typeof t == "string" && t) {
			var i = un(t);
			i = "link[rel=\"" + e + "\"][href=\"" + i + "\"]", typeof n == "string" && (i += "[crossorigin=\"" + n + "\"]"), ym.has(i) || (ym.add(i), e = {
				rel: e,
				crossOrigin: n,
				href: t
			}, r.querySelector(i) === null && (t = r.createElement("link"), np(t, "link", e), Ut(t), r.head.appendChild(t)));
		}
	}
	function Em(e) {
		xm.D(e), Tm("dns-prefetch", e, null);
	}
	function Dm(e, t) {
		xm.C(e, t), Tm("preconnect", e, t);
	}
	function Om(e, t, n) {
		xm.L(e, t, n);
		var r = wm;
		if (r && e && t) {
			var i = "link[rel=\"preload\"][as=\"" + un(t) + "\"]";
			t === "image" && n && n.imageSrcSet ? (i += "[imagesrcset=\"" + un(n.imageSrcSet) + "\"]", typeof n.imageSizes == "string" && (i += "[imagesizes=\"" + un(n.imageSizes) + "\"]")) : i += "[href=\"" + un(e) + "\"]";
			var a = i;
			switch (t) {
				case "style":
					a = Pm(e);
					break;
				case "script": a = Rm(e);
			}
			if (!(vm.has(a) || (e = S({
				rel: "preload",
				href: t === "image" && n && n.imageSrcSet ? void 0 : e,
				as: t
			}, n), vm.set(a, e), r.querySelector(i) !== null || t === "style" && r.querySelector(Fm(a)) || t === "script" && r.querySelector(zm(a))))) {
				var o = r.createElement("link");
				np(o, "link", e), t === "style" && (o[Lt] = !0, o.onload = o.onerror = function() {
					Wt(o);
				}), Ut(o), r.head.appendChild(o);
			}
		}
	}
	function km(e, t) {
		xm.m(e, t);
		var n = wm;
		if (n && e) {
			var r = t && typeof t.as == "string" ? t.as : "script", i = "link[rel=\"modulepreload\"][as=\"" + un(r) + "\"][href=\"" + un(e) + "\"]", a = i;
			switch (r) {
				case "audioworklet":
				case "paintworklet":
				case "serviceworker":
				case "sharedworker":
				case "worker":
				case "script": a = Rm(e);
			}
			if (!vm.has(a) && (e = S({
				rel: "modulepreload",
				href: e
			}, t), vm.set(a, e), n.querySelector(i) === null)) {
				switch (r) {
					case "audioworklet":
					case "paintworklet":
					case "serviceworker":
					case "sharedworker":
					case "worker":
					case "script": if (n.querySelector(zm(a))) return;
				}
				r = n.createElement("link"), np(r, "link", e), Ut(r), n.head.appendChild(r);
			}
		}
	}
	function Am(e, t, n) {
		xm.S(e, t, n);
		var r = wm;
		if (r && e) {
			var i = Ht(r).hoistableStyles, a = Pm(e);
			t ||= "default";
			var o = i.get(a);
			if (!o) {
				var s = {
					loading: 0,
					preload: null
				};
				if (o = r.querySelector(Fm(a))) s.loading = 5;
				else {
					e = S({
						rel: "stylesheet",
						href: e,
						"data-precedence": t
					}, n), (n = vm.get(a)) && Hm(e, n);
					var c = o = r.createElement("link");
					Ut(c), np(c, "link", e), c._p = new Promise(function(e, t) {
						c.onload = e, c.onerror = t;
					}), c.addEventListener("load", function() {
						s.loading |= 1;
					}), c.addEventListener("error", function() {
						s.loading |= 2;
					}), s.loading |= 4, Vm(o, t, r);
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
	function jm(e, t) {
		xm.X(e, t);
		var n = wm;
		if (n && e) {
			var r = Ht(n).hoistableScripts, i = Rm(e), a = r.get(i);
			a || (a = n.querySelector(zm(i)), a || (e = S({
				src: e,
				async: !0
			}, t), (t = vm.get(i)) && Um(e, t), a = n.createElement("script"), Ut(a), np(a, "link", e), n.head.appendChild(a)), a = {
				type: "script",
				instance: a,
				count: 1,
				state: null
			}, r.set(i, a));
		}
	}
	function Mm(e, t) {
		xm.M(e, t);
		var n = wm;
		if (n && e) {
			var r = Ht(n).hoistableScripts, i = Rm(e), a = r.get(i);
			a || (a = n.querySelector(zm(i)), a || (e = S({
				src: e,
				async: !0,
				type: "module"
			}, t), (t = vm.get(i)) && Um(e, t), a = n.createElement("script"), Ut(a), np(a, "link", e), n.head.appendChild(a)), a = {
				type: "script",
				instance: a,
				count: 1,
				state: null
			}, r.set(i, a));
		}
	}
	function Nm(e, t, n, r) {
		var i = (i = ke.current) ? bm(i) : null;
		if (!i) throw Error(o(446));
		switch (e) {
			case "meta":
			case "title": return null;
			case "style": return typeof n.precedence == "string" && typeof n.href == "string" ? (n = Pm(n.href), t = Ht(i).hoistableStyles, r = t.get(n), r || (r = {
				type: "style",
				instance: null,
				count: 0,
				state: null
			}, t.set(n, r)), r) : {
				type: "void",
				instance: null,
				count: 0,
				state: null
			};
			case "link":
				if (n.rel === "stylesheet" && typeof n.href == "string" && typeof n.precedence == "string") {
					e = Pm(n.href);
					var a = Ht(i).hoistableStyles, s = a.get(e);
					if (s || (i = i.ownerDocument || i, s = {
						type: "stylesheet",
						instance: null,
						count: 0,
						state: {
							loading: 0,
							preload: null
						}
					}, a.set(e, s), (a = i.querySelector(Fm(e))) ? a._p || (s.instance = a, s.state.loading = 5) : (a = vm.get(e), a || (a = {
						rel: "preload",
						as: "style",
						href: n.href,
						crossOrigin: n.crossOrigin,
						integrity: n.integrity,
						media: n.media,
						hrefLang: n.hrefLang,
						referrerPolicy: n.referrerPolicy
					}, vm.set(e, a)), Lm(i, e, a, s.state))), t && r === null) throw Error(o(528, ""));
					return s;
				}
				if (t && r !== null) throw Error(o(529, ""));
				return null;
			case "script": return t = n.async, n = n.src, typeof n == "string" && t && typeof t != "function" && typeof t != "symbol" ? (n = Rm(n), t = Ht(i).hoistableScripts, r = t.get(n), r || (r = {
				type: "script",
				instance: null,
				count: 0,
				state: null
			}, t.set(n, r)), r) : {
				type: "void",
				instance: null,
				count: 0,
				state: null
			};
			default: throw Error(o(444, e));
		}
	}
	function Pm(e) {
		return "href=\"" + un(e) + "\"";
	}
	function Fm(e) {
		return "link[rel=\"stylesheet\"][" + e + "]";
	}
	function Im(e) {
		return S({}, e, {
			"data-precedence": e.precedence,
			precedence: null
		});
	}
	function Lm(e, t, n, r) {
		if (t = e.querySelector("link[rel=\"preload\"][as=\"style\"][" + t + "]")) {
			if (!0 !== t[Lt]) {
				r.loading = 1;
				return;
			}
		} else t = e.createElement("link"), t[Lt] = !0, t.onload = t.onerror = Wt.bind(null, t), np(t, "link", n), Ut(t), e.head.appendChild(t);
		r.preload = t, t.addEventListener("load", function() {
			return r.loading |= 1;
		}), t.addEventListener("error", function() {
			return r.loading |= 2;
		});
	}
	function Rm(e) {
		return "[src=\"" + un(e) + "\"]";
	}
	function zm(e) {
		return "script[async]" + e;
	}
	function Bm(e, t, n) {
		if (t.count++, t.instance === null) switch (t.type) {
			case "style":
				var r = e.querySelector("style[data-href~=\"" + un(n.href) + "\"]");
				if (r) return t.instance = r, Ut(r), r;
				var i = S({}, n, {
					"data-href": n.href,
					"data-precedence": n.precedence,
					href: null,
					precedence: null
				});
				return r = (e.ownerDocument || e).createElement("style"), Ut(r), np(r, "style", i), Vm(r, n.precedence, e), t.instance = r;
			case "stylesheet":
				i = Pm(n.href);
				var a = e.querySelector(Fm(i));
				if (a) return t.state.loading |= 4, t.instance = a, Ut(a), a;
				r = Im(n), (i = vm.get(i)) && Hm(r, i), a = (e.ownerDocument || e).createElement("link"), Ut(a);
				var s = a;
				return s._p = new Promise(function(e, t) {
					s.onload = e, s.onerror = t;
				}), np(a, "link", r), t.state.loading |= 4, Vm(a, n.precedence, e), t.instance = a;
			case "script": return a = Rm(n.src), (i = e.querySelector(zm(a))) ? (t.instance = i, Ut(i), i) : (r = n, (i = vm.get(a)) && (r = S({}, n), Um(r, i)), e = e.ownerDocument || e, i = e.createElement("script"), Ut(i), np(i, "link", r), e.head.appendChild(i), t.instance = i);
			case "void": return null;
			default: throw Error(o(443, t.type));
		}
		else t.type === "stylesheet" && !(t.state.loading & 4) && (r = t.instance, t.state.loading |= 4, Vm(r, n.precedence, e));
		return t.instance;
	}
	function Vm(e, t, n) {
		for (var r = n.querySelectorAll("link[rel=\"stylesheet\"][data-precedence],style[data-precedence]"), i = r.length ? r[r.length - 1] : null, a = i, o = 0; o < r.length; o++) {
			var s = r[o];
			if (s.dataset.precedence === t) a = s;
			else if (a !== i) break;
		}
		a ? a.parentNode.insertBefore(e, a.nextSibling) : (t = n.nodeType === 9 ? n.head : n, t.insertBefore(e, t.firstChild));
	}
	function Hm(e, t) {
		e.crossOrigin ??= t.crossOrigin, e.referrerPolicy ??= t.referrerPolicy, e.title ??= t.title;
	}
	function Um(e, t) {
		e.crossOrigin ??= t.crossOrigin, e.referrerPolicy ??= t.referrerPolicy, e.integrity ??= t.integrity;
	}
	var Wm = null;
	function Gm(e, t, n) {
		if (Wm === null) {
			var r = /* @__PURE__ */ new Map(), i = Wm = /* @__PURE__ */ new Map();
			i.set(n, r);
		} else i = Wm, r = i.get(n), r || (r = /* @__PURE__ */ new Map(), i.set(n, r));
		if (r.has(e)) return r;
		for (r.set(e, null), n = n.getElementsByTagName(e), i = 0; i < n.length; i++) {
			var a = n[i];
			if (!(a[It] || a[kt] || e === "link" && a.getAttribute("rel") === "stylesheet") && a.namespaceURI !== "http://www.w3.org/2000/svg") {
				var o = a.getAttribute(t) || "";
				o = e + o;
				var s = r.get(o);
				s ? s.push(a) : r.set(o, [a]);
			}
		}
		return r;
	}
	function Km(e, t, n) {
		e = e.ownerDocument || e, e.head.insertBefore(n, t === "title" ? e.querySelector("head > title") : null);
	}
	function qm(e, t, n) {
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
	function Jm(e, t) {
		return e === "img" && t.src != null && t.src !== "" && t.onLoad == null && t.loading !== "lazy";
	}
	function Ym(e) {
		return !(e.type === "stylesheet" && !(e.state.loading & 3));
	}
	function Xm(e) {
		return (e.width || 100) * (e.height || 100) * (typeof devicePixelRatio == "number" ? devicePixelRatio : 1) * .25;
	}
	function Zm(e, t) {
		typeof t.decode == "function" && (e.imgCount++, t.complete || (e.imgBytes += Xm(t), e.suspenseyImages.push(t)), e = rh.bind(e), t.decode().then(e, e));
	}
	function Qm(e, t, n, r) {
		if (n.type === "stylesheet" && (typeof r.media != "string" || !1 !== matchMedia(r.media).matches) && !(n.state.loading & 4)) {
			if (n.instance === null) {
				var i = Pm(r.href), a = t.querySelector(Fm(i));
				if (a) {
					t = a._p, typeof t == "object" && t && typeof t.then == "function" && (e.count++, e = nh.bind(e), t.then(e, e)), n.state.loading |= 4, n.instance = a, Ut(a);
					return;
				}
				a = t.ownerDocument || t, r = Im(r), (i = vm.get(i)) && Hm(r, i), a = a.createElement("link"), Ut(a);
				var o = a;
				o._p = new Promise(function(e, t) {
					o.onload = e, o.onerror = t;
				}), np(a, "link", r), n.instance = a;
			}
			e.stylesheets === null && (e.stylesheets = /* @__PURE__ */ new Map()), e.stylesheets.set(n, t), (t = n.state.preload) && !(n.state.loading & 3) && (e.count++, n = nh.bind(e), t.addEventListener("load", n), t.addEventListener("error", n));
		}
	}
	var $m = 0;
	function eh(e, t) {
		return e.stylesheets && e.count === 0 && ah(e, e.stylesheets), 0 < e.count || 0 < e.imgCount ? function(n) {
			var r = setTimeout(function() {
				if (e.stylesheets && ah(e, e.stylesheets), e.unsuspend) {
					var t = e.unsuspend;
					e.unsuspend = null, t();
				}
			}, 6e4 + t);
			0 < e.imgBytes && $m === 0 && ($m = 62500 * op());
			var i = setTimeout(function() {
				if (e.waitingForImages = !1, e.count === 0 && (e.stylesheets && ah(e, e.stylesheets), e.unsuspend)) {
					var t = e.unsuspend;
					e.unsuspend = null, t();
				}
			}, (e.imgBytes > $m ? 50 : 800) + t);
			return e.unsuspend = n, function() {
				e.unsuspend = null, clearTimeout(r), clearTimeout(i);
			};
		} : null;
	}
	function th(e) {
		if (e.count === 0 && (e.imgCount === 0 || !e.waitingForImages)) {
			if (e.stylesheets) ah(e, e.stylesheets);
			else if (e.unsuspend) {
				var t = e.unsuspend;
				e.unsuspend = null, t();
			}
		}
	}
	function nh() {
		this.count--, th(this);
	}
	function rh() {
		this.imgCount--, th(this);
	}
	var ih = null;
	function ah(e, t) {
		e.stylesheets = null, e.unsuspend !== null && (e.count++, ih = /* @__PURE__ */ new Map(), t.forEach(oh, e), ih = null, nh.call(e));
	}
	function oh(e, t) {
		if (!(t.state.loading & 4)) {
			var n = ih.get(e);
			if (n) var r = n.get(null);
			else {
				n = /* @__PURE__ */ new Map(), ih.set(e, n);
				for (var i = e.querySelectorAll("link[data-precedence],style[data-precedence]"), a = 0; a < i.length; a++) {
					var o = i[a];
					(o.nodeName === "LINK" || o.getAttribute("media") !== "not all") && (n.set(o.dataset.precedence, o), r = o);
				}
				r && n.set(null, r);
			}
			i = t.instance, o = i.getAttribute("data-precedence"), a = n.get(o) || r, a === r && n.set(null, i), n.set(o, i), this.count++, r = nh.bind(this), i.addEventListener("load", r), i.addEventListener("error", r), a ? a.parentNode.insertBefore(i, a.nextSibling) : (e = e.nodeType === 9 ? e.head : e, e.insertBefore(i, e.firstChild)), t.state.loading |= 4;
		}
	}
	var sh = {
		$$typeof: de,
		Provider: null,
		Consumer: null,
		_currentValue: Se,
		_currentValue2: Se,
		_threadCount: 0
	};
	function ch(e, t, n, r, i, a, o, s, c) {
		this.tag = 1, this.containerInfo = e, this.pingCache = this.current = this.pendingChildren = null, this.timeoutHandle = -1, this.callbackNode = this.next = this.pendingContext = this.context = this.cancelPendingCommit = null, this.callbackPriority = 0, this.expirationTimes = vt(-1), this.entangledLanes = this.shellSuspendCounter = this.errorRecoveryDisabledLanes = this.expiredLanes = this.warmLanes = this.pingedLanes = this.suspendedLanes = this.pendingLanes = 0, this.entanglements = vt(0), this.hiddenUpdates = vt(null), this.identifierPrefix = r, this.onUncaughtError = i, this.onCaughtError = a, this.onRecoverableError = o, this.pooledCache = null, this.pooledCacheLanes = 0, this.formState = c, this.transitionTypes = null, this.incompleteTransitions = /* @__PURE__ */ new Map();
	}
	function lh(e, t, n, r, i, a, o, s, c, l, u, d) {
		return e = new ch(e, t, n, o, c, l, u, d, s), t = 1, !0 === a && (t |= 24), a = zi(3, null, null, t), e.current = a, a.stateNode = e, t = Ra(), t.refCount++, e.pooledCache = t, t.refCount++, a.memoizedState = {
			element: r,
			isDehydrated: n,
			cache: t
		}, xo(a), e;
	}
	function uh(e) {
		return e ? (e = Li, e) : Li;
	}
	function dh(e, t, n, r, i, a) {
		i = uh(i), r.context === null ? r.context = i : r.pendingContext = i, r = Co(t), r.payload = { element: n }, a = a === void 0 ? null : a, a !== null && (r.callback = a), n = wo(e, r, t), n !== null && (Pd(n, e, t), To(n, e, t));
	}
	function fh(e, t) {
		if (e = e.memoizedState, e !== null && e.dehydrated !== null) {
			var n = e.retryLane;
			e.retryLane = n !== 0 && n < t ? n : t;
		}
	}
	function ph(e, t) {
		fh(e, t), (e = e.alternate) && fh(e, t);
	}
	function mh(e) {
		if (e.tag === 13 || e.tag === 31) {
			var t = Pi(e, 67108864);
			t !== null && Pd(t, e, 67108864), ph(e, 67108864);
		}
	}
	function hh(e) {
		if (e.tag === 13 || e.tag === 31) {
			var t = jd();
			t = wt(t);
			var n = Pi(e, t);
			n !== null && Pd(n, e, t), ph(e, t);
		}
	}
	var gh = !0;
	function _h(e, t, n, r) {
		var i = O.T;
		O.T = null;
		var a = k.p;
		try {
			k.p = 2, yh(e, t, n, r);
		} finally {
			k.p = a, O.T = i;
		}
	}
	function vh(e, t, n, r) {
		var i = O.T;
		O.T = null;
		var a = k.p;
		try {
			k.p = 8, yh(e, t, n, r);
		} finally {
			k.p = a, O.T = i;
		}
	}
	function yh(e, t, n, r) {
		if (gh) {
			var i = bh(r);
			if (i === null) Kf(e, t, r, xh, n), Mh(e, r);
			else if (Ph(i, e, t, n, r)) r.stopPropagation();
			else if (Mh(e, r), t & 4 && -1 < jh.indexOf(e)) {
				for (; i !== null;) {
					var a = Bt(i);
					if (a !== null) switch (a.tag) {
						case 3:
							if (a = a.stateNode, a.current.memoizedState.isDehydrated) {
								var o = ft(a.pendingLanes);
								if (o !== 0) {
									var s = a;
									for (s.pendingLanes |= 2, s.entangledLanes |= 2; o;) {
										var c = 1 << 31 - at(o);
										s.entanglements[1] |= c, o &= ~c;
									}
									Ef(a), !(W & 6) && (_d = qe() + 500, Df(0, !1));
								}
							}
							break;
						case 31:
						case 13: s = Pi(a, 2), s !== null && Pd(s, a, 2), zd(), ph(a, 2);
					}
					if (a = bh(r), a === null && Kf(e, t, r, xh, n), a === i) break;
					i = a;
				}
				i !== null && r.stopPropagation();
			} else Kf(e, t, r, null, n);
		}
	}
	function bh(e) {
		return e = Dn(e), Sh(e);
	}
	var xh = null;
	function Sh(e) {
		if (xh = null, e = zt(e), e !== null) {
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
		return xh = e, null;
	}
	function Ch(e) {
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
			case "fullscreenerror":
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
			case "resize":
			case "scroll":
			case "touchmove":
			case "wheel":
			case "mouseenter":
			case "mouseleave":
			case "pointerenter":
			case "pointerleave": return 8;
			case "message": switch (Je()) {
				case Ye: return 2;
				case Xe: return 8;
				case Ze:
				case Qe: return 32;
				case $e: return 268435456;
				default: return 32;
			}
			default: return 32;
		}
	}
	var wh = !1, Th = null, Eh = null, Dh = null, Oh = /* @__PURE__ */ new Map(), kh = /* @__PURE__ */ new Map(), Ah = [], jh = "mousedown mouseup touchcancel touchend touchstart auxclick dblclick pointercancel pointerdown pointerup dragend dragstart drop compositionend compositionstart keydown keypress keyup input textInput copy cut paste click change contextmenu reset".split(" ");
	function Mh(e, t) {
		switch (e) {
			case "focusin":
			case "focusout":
				Th = null;
				break;
			case "dragenter":
			case "dragleave":
				Eh = null;
				break;
			case "mouseover":
			case "mouseout":
				Dh = null;
				break;
			case "pointerover":
			case "pointerout":
				Oh.delete(t.pointerId);
				break;
			case "gotpointercapture":
			case "lostpointercapture": kh.delete(t.pointerId);
		}
	}
	function Nh(e, t, n, r, i, a) {
		return e === null || e.nativeEvent !== a ? (e = {
			blockedOn: t,
			domEventName: n,
			eventSystemFlags: r,
			nativeEvent: a,
			targetContainers: [i]
		}, t !== null && (t = Bt(t), t !== null && mh(t)), e) : (e.eventSystemFlags |= r, t = e.targetContainers, i !== null && t.indexOf(i) === -1 && t.push(i), e);
	}
	function Ph(e, t, n, r, i) {
		switch (t) {
			case "focusin": return Th = Nh(Th, e, t, n, r, i), !0;
			case "dragenter": return Eh = Nh(Eh, e, t, n, r, i), !0;
			case "mouseover": return Dh = Nh(Dh, e, t, n, r, i), !0;
			case "pointerover":
				var a = i.pointerId;
				return Oh.set(a, Nh(Oh.get(a) || null, e, t, n, r, i)), !0;
			case "gotpointercapture": return a = i.pointerId, kh.set(a, Nh(kh.get(a) || null, e, t, n, r, i)), !0;
		}
		return !1;
	}
	function Fh(e) {
		var t = zt(e.target);
		if (t !== null) {
			var n = c(t);
			if (n !== null) {
				if (t = n.tag, t === 13) {
					if (t = l(n), t !== null) {
						e.blockedOn = t, Dt(e.priority, function() {
							hh(n);
						});
						return;
					}
				} else if (t === 31) {
					if (t = u(n), t !== null) {
						e.blockedOn = t, Dt(e.priority, function() {
							hh(n);
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
	function Ih(e) {
		if (e.blockedOn !== null) return !1;
		for (var t = e.targetContainers; 0 < t.length;) {
			var n = bh(e.nativeEvent);
			if (n === null) {
				n = e.nativeEvent;
				var r = new n.constructor(n.type, n);
				En = r, n.target.dispatchEvent(r), En = null;
			} else return t = Bt(n), t !== null && mh(t), e.blockedOn = n, !1;
			t.shift();
		}
		return !0;
	}
	function Lh(e, t, n) {
		Ih(e) && n.delete(t);
	}
	function Rh() {
		wh = !1, Th !== null && Ih(Th) && (Th = null), Eh !== null && Ih(Eh) && (Eh = null), Dh !== null && Ih(Dh) && (Dh = null), Oh.forEach(Lh), kh.forEach(Lh);
	}
	function zh(e, n) {
		e.blockedOn === n && (e.blockedOn = null, wh || (wh = !0, t.unstable_scheduleCallback(t.unstable_NormalPriority, Rh)));
	}
	var Bh = null;
	function Vh(e) {
		Bh !== e && (Bh = e, t.unstable_scheduleCallback(t.unstable_NormalPriority, function() {
			Bh === e && (Bh = null);
			for (var t = 0; t < e.length; t += 3) {
				var n = e[t], r = e[t + 1], i = e[t + 2];
				if (typeof r != "function") {
					if (Sh(r || n) === null) continue;
					break;
				}
				var a = Bt(n);
				a !== null && (e.splice(t, 3), t -= 3, rc(a, {
					pending: !0,
					data: i,
					method: n.method,
					action: r
				}, r, i));
			}
		}));
	}
	function Hh(e) {
		function t(t) {
			return zh(t, e);
		}
		Th !== null && zh(Th, e), Eh !== null && zh(Eh, e), Dh !== null && zh(Dh, e), Oh.forEach(t), kh.forEach(t);
		for (var n = 0; n < Ah.length; n++) {
			var r = Ah[n];
			r.blockedOn === e && (r.blockedOn = null);
		}
		for (; 0 < Ah.length && (n = Ah[0], n.blockedOn === null);) Fh(n), n.blockedOn === null && Ah.shift();
		if (n = (e.ownerDocument || e).$$reactFormReplay, n != null) for (r = 0; r < n.length; r += 3) {
			var i = n[r], a = n[r + 1], o = i[At] || null;
			if (typeof a == "function") o || Vh(n);
			else if (o) {
				var s = null;
				if (a && a.hasAttribute("formAction")) {
					if (i = a, o = a[At] || null) s = o.formAction;
					else if (Sh(i) !== null) continue;
				} else s = o.action;
				typeof s == "function" ? n[r + 1] = s : (n.splice(r, 3), r -= 3), Vh(n);
			}
		}
	}
	function Uh() {
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
	function Wh(e) {
		this._internalRoot = e;
	}
	Gh.prototype.render = Wh.prototype.render = function(e) {
		var t = this._internalRoot;
		if (t === null) throw Error(o(409));
		var n = t.current;
		dh(n, jd(), e, t, null, null);
	}, Gh.prototype.unmount = Wh.prototype.unmount = function() {
		var e = this._internalRoot;
		if (e !== null) {
			this._internalRoot = null;
			var t = e.containerInfo;
			dh(e.current, 2, null, e, null, null), zd(), t[jt] = null;
		}
	};
	function Gh(e) {
		this._internalRoot = e;
	}
	Gh.prototype.unstable_scheduleHydration = function(e) {
		if (e) {
			var t = Et();
			e = {
				blockedOn: null,
				target: e,
				priority: t
			};
			for (var n = 0; n < Ah.length && t !== 0 && t < Ah[n].priority; n++);
			Ah.splice(n, 0, e), n === 0 && Fh(e);
		}
	};
	var Kh = r.version;
	if (Kh !== "19.3.1") throw Error(o(527, Kh, "19.3.1"));
	k.findDOMNode = function(e) {
		var t = e._reactInternals;
		if (t === void 0) throw typeof e.render == "function" ? Error(o(188)) : (e = Object.keys(e).join(","), Error(o(268, e)));
		return e = f(t), e = e === null ? null : p(e), e = e === null ? null : e.stateNode, e;
	};
	var qh = {
		bundleType: 0,
		version: "19.3.1",
		rendererPackageName: "react-dom",
		currentDispatcherRef: O,
		reconcilerVersion: "19.3.1"
	};
	if (typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ < "u") {
		var Jh = __REACT_DEVTOOLS_GLOBAL_HOOK__;
		if (!Jh.isDisabled && Jh.supportsFiber) try {
			nt = Jh.inject(qh), rt = Jh;
		} catch {}
	}
	e.createRoot = function(e, t) {
		if (!s(e)) throw Error(o(299));
		var n = !1, r = "", i = Ec, a = Dc, c = Oc;
		return t != null && (!0 === t.unstable_strictMode && (n = !0), t.identifierPrefix !== void 0 && (r = t.identifierPrefix), t.onUncaughtError !== void 0 && (i = t.onUncaughtError), t.onCaughtError !== void 0 && (a = t.onCaughtError), t.onRecoverableError !== void 0 && (c = t.onRecoverableError)), t = lh(e, 1, !1, null, null, n, r, null, i, a, c, Uh), e[jt] = t.current, Wf(e), new Wh(t);
	};
})), Zt = (/* @__PURE__ */ t(((e, t) => {
	function n() {
		if (typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ < "u" && typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE == "function") try {
			__REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(n);
		} catch (e) {
			console.error(e);
		}
	}
	n(), t.exports = Xt();
})))(), Qt = "/*! tailwindcss v4.3.3 | MIT License | https://tailwindcss.com */\n@layer properties{@supports (((-webkit-hyphens:none)) and (not (margin-trim:inline))) or ((-moz-orient:inline) and (not (color:rgb(from red r g b)))){*,:before,:after,::backdrop{--tw-border-spacing-x:0;--tw-border-spacing-y:0;--tw-translate-x:0;--tw-translate-y:0;--tw-translate-z:0;--tw-scale-x:1;--tw-scale-y:1;--tw-scale-z:1;--tw-rotate-x:initial;--tw-rotate-y:initial;--tw-rotate-z:initial;--tw-skew-x:initial;--tw-skew-y:initial;--tw-space-y-reverse:0;--tw-divide-x-reverse:0;--tw-border-style:solid;--tw-divide-y-reverse:0;--tw-gradient-position:initial;--tw-gradient-from:#0000;--tw-gradient-via:#0000;--tw-gradient-to:#0000;--tw-gradient-stops:initial;--tw-gradient-via-stops:initial;--tw-gradient-from-position:0%;--tw-gradient-via-position:50%;--tw-gradient-to-position:100%;--tw-leading:initial;--tw-font-weight:initial;--tw-tracking:initial;--tw-ordinal:initial;--tw-slashed-zero:initial;--tw-numeric-figure:initial;--tw-numeric-spacing:initial;--tw-numeric-fraction:initial;--tw-shadow:0 0 #0000;--tw-shadow-color:initial;--tw-shadow-alpha:100%;--tw-inset-shadow:0 0 #0000;--tw-inset-shadow-color:initial;--tw-inset-shadow-alpha:100%;--tw-ring-color:initial;--tw-ring-shadow:0 0 #0000;--tw-inset-ring-color:initial;--tw-inset-ring-shadow:0 0 #0000;--tw-ring-inset:initial;--tw-ring-offset-width:0;--tw-ring-offset-color:#fff;--tw-ring-offset-shadow:0 0 #0000;--tw-outline-style:solid;--tw-blur:initial;--tw-brightness:initial;--tw-contrast:initial;--tw-grayscale:initial;--tw-hue-rotate:initial;--tw-invert:initial;--tw-opacity:initial;--tw-saturate:initial;--tw-sepia:initial;--tw-drop-shadow:initial;--tw-drop-shadow-color:initial;--tw-drop-shadow-alpha:100%;--tw-drop-shadow-size:initial;--tw-backdrop-blur:initial;--tw-backdrop-brightness:initial;--tw-backdrop-contrast:initial;--tw-backdrop-grayscale:initial;--tw-backdrop-hue-rotate:initial;--tw-backdrop-invert:initial;--tw-backdrop-opacity:initial;--tw-backdrop-saturate:initial;--tw-backdrop-sepia:initial;--tw-duration:initial;--tw-ease:initial;--tw-contain-size:initial;--tw-contain-layout:initial;--tw-contain-paint:initial;--tw-contain-style:initial;--tw-content:\"\"}}}@layer theme{:root,:host{--font-serif:ui-serif, Georgia, Cambria, \"Times New Roman\", Times, serif;--color-red-50:oklch(97.1% .013 17.38);--color-red-100:oklch(93.6% .032 17.717);--color-red-200:oklch(88.5% .062 18.334);--color-red-300:oklch(80.8% .114 19.571);--color-red-400:oklch(70.4% .191 22.216);--color-red-500:oklch(63.7% .237 25.331);--color-red-600:oklch(57.7% .245 27.325);--color-red-700:oklch(50.5% .213 27.518);--color-red-800:oklch(44.4% .177 26.899);--color-red-900:oklch(39.6% .141 25.723);--color-red-950:oklch(25.8% .092 26.042);--color-orange-50:oklch(98% .016 73.684);--color-orange-100:oklch(95.4% .038 75.164);--color-orange-200:oklch(90.1% .076 70.697);--color-orange-300:oklch(83.7% .128 66.29);--color-orange-400:oklch(75% .183 55.934);--color-orange-500:oklch(70.5% .213 47.604);--color-orange-600:oklch(64.6% .222 41.116);--color-orange-700:oklch(55.3% .195 38.402);--color-orange-800:oklch(47% .157 37.304);--color-orange-900:oklch(40.8% .123 38.172);--color-orange-950:oklch(26.6% .079 36.259);--color-amber-50:oklch(98.7% .022 95.277);--color-amber-100:oklch(96.2% .059 95.617);--color-amber-200:oklch(92.4% .12 95.746);--color-amber-300:oklch(87.9% .169 91.605);--color-amber-400:oklch(82.8% .189 84.429);--color-amber-500:oklch(76.9% .188 70.08);--color-amber-600:oklch(66.6% .179 58.318);--color-amber-700:oklch(55.5% .163 48.998);--color-amber-800:oklch(47.3% .137 46.201);--color-amber-900:oklch(41.4% .112 45.904);--color-amber-950:oklch(27.9% .077 45.635);--color-yellow-50:oklch(98.7% .026 102.212);--color-yellow-100:oklch(97.3% .071 103.193);--color-yellow-200:oklch(94.5% .129 101.54);--color-yellow-300:oklch(90.5% .182 98.111);--color-yellow-400:oklch(85.2% .199 91.936);--color-yellow-500:oklch(79.5% .184 86.047);--color-yellow-600:oklch(68.1% .162 75.834);--color-yellow-700:oklch(55.4% .135 66.442);--color-yellow-800:oklch(47.6% .114 61.907);--color-yellow-900:oklch(42.1% .095 57.708);--color-yellow-950:oklch(28.6% .066 53.813);--color-lime-50:oklch(98.6% .031 120.757);--color-lime-100:oklch(96.7% .067 122.328);--color-lime-200:oklch(93.8% .127 124.321);--color-lime-300:oklch(89.7% .196 126.665);--color-lime-400:oklch(84.1% .238 128.85);--color-lime-500:oklch(76.8% .233 130.85);--color-lime-600:oklch(64.8% .2 131.684);--color-lime-700:oklch(53.2% .157 131.589);--color-lime-800:oklch(45.3% .124 130.933);--color-lime-900:oklch(40.5% .101 131.063);--color-lime-950:oklch(27.4% .072 132.109);--color-green-50:oklch(98.2% .018 155.826);--color-green-100:oklch(96.2% .044 156.743);--color-green-200:oklch(92.5% .084 155.995);--color-green-300:oklch(87.1% .15 154.449);--color-green-400:oklch(79.2% .209 151.711);--color-green-500:oklch(72.3% .219 149.579);--color-green-600:oklch(62.7% .194 149.214);--color-green-700:oklch(52.7% .154 150.069);--color-green-800:oklch(44.8% .119 151.328);--color-green-900:oklch(39.3% .095 152.535);--color-green-950:oklch(26.6% .065 152.934);--color-emerald-50:oklch(97.9% .021 166.113);--color-emerald-100:oklch(95% .052 163.051);--color-emerald-200:oklch(90.5% .093 164.15);--color-emerald-300:oklch(84.5% .143 164.978);--color-emerald-400:oklch(76.5% .177 163.223);--color-emerald-500:oklch(69.6% .17 162.48);--color-emerald-600:oklch(59.6% .145 163.225);--color-emerald-700:oklch(50.8% .118 165.612);--color-emerald-800:oklch(43.2% .095 166.913);--color-emerald-900:oklch(37.8% .077 168.94);--color-emerald-950:oklch(26.2% .051 172.552);--color-teal-50:oklch(98.4% .014 180.72);--color-teal-100:oklch(95.3% .051 180.801);--color-teal-200:oklch(91% .096 180.426);--color-teal-300:oklch(85.5% .138 181.071);--color-teal-400:oklch(77.7% .152 181.912);--color-teal-500:oklch(70.4% .14 182.503);--color-teal-600:oklch(60% .118 184.704);--color-teal-700:oklch(51.1% .096 186.391);--color-teal-800:oklch(43.7% .078 188.216);--color-teal-900:oklch(38.6% .063 188.416);--color-teal-950:oklch(27.7% .046 192.524);--color-cyan-50:oklch(98.4% .019 200.873);--color-cyan-100:oklch(95.6% .045 203.388);--color-cyan-200:oklch(91.7% .08 205.041);--color-cyan-300:oklch(86.5% .127 207.078);--color-cyan-400:oklch(78.9% .154 211.53);--color-cyan-500:oklch(71.5% .143 215.221);--color-cyan-600:oklch(60.9% .126 221.723);--color-cyan-700:oklch(52% .105 223.128);--color-cyan-800:oklch(45% .085 224.283);--color-cyan-900:oklch(39.8% .07 227.392);--color-cyan-950:oklch(30.2% .056 229.695);--color-sky-50:oklch(97.7% .013 236.62);--color-sky-100:oklch(95.1% .026 236.824);--color-sky-200:oklch(90.1% .058 230.902);--color-sky-300:oklch(82.8% .111 230.318);--color-sky-400:oklch(74.6% .16 232.661);--color-sky-500:oklch(68.5% .169 237.323);--color-sky-600:oklch(58.8% .158 241.966);--color-sky-700:oklch(50% .134 242.749);--color-sky-800:oklch(44.3% .11 240.79);--color-sky-900:oklch(39.1% .09 240.876);--color-sky-950:oklch(29.3% .066 243.157);--color-blue-50:oklch(97% .014 254.604);--color-blue-100:oklch(93.2% .032 255.585);--color-blue-200:oklch(88.2% .059 254.128);--color-blue-300:oklch(80.9% .105 251.813);--color-blue-400:oklch(70.7% .165 254.624);--color-blue-500:oklch(62.3% .214 259.815);--color-blue-600:oklch(54.6% .245 262.881);--color-blue-700:oklch(48.8% .243 264.376);--color-blue-800:oklch(42.4% .199 265.638);--color-blue-900:oklch(37.9% .146 265.522);--color-blue-950:oklch(28.2% .091 267.935);--color-indigo-50:oklch(96.2% .018 272.314);--color-indigo-100:oklch(93% .034 272.788);--color-indigo-200:oklch(87% .065 274.039);--color-indigo-300:oklch(78.5% .115 274.713);--color-indigo-400:oklch(67.3% .182 276.935);--color-indigo-500:oklch(58.5% .233 277.117);--color-indigo-600:oklch(51.1% .262 276.966);--color-indigo-700:oklch(45.7% .24 277.023);--color-indigo-800:oklch(39.8% .195 277.366);--color-indigo-900:oklch(35.9% .144 278.697);--color-indigo-950:oklch(25.7% .09 281.288);--color-violet-50:oklch(96.9% .016 293.756);--color-violet-100:oklch(94.3% .029 294.588);--color-violet-200:oklch(89.4% .057 293.283);--color-violet-300:oklch(81.1% .111 293.571);--color-violet-400:oklch(70.2% .183 293.541);--color-violet-500:oklch(60.6% .25 292.717);--color-violet-600:oklch(54.1% .281 293.009);--color-violet-700:oklch(49.1% .27 292.581);--color-violet-800:oklch(43.2% .232 292.759);--color-violet-900:oklch(38% .189 293.745);--color-violet-950:oklch(28.3% .141 291.089);--color-purple-50:oklch(97.7% .014 308.299);--color-purple-100:oklch(94.6% .033 307.174);--color-purple-200:oklch(90.2% .063 306.703);--color-purple-300:oklch(82.7% .119 306.383);--color-purple-400:oklch(71.4% .203 305.504);--color-purple-500:oklch(62.7% .265 303.9);--color-purple-600:oklch(55.8% .288 302.321);--color-purple-700:oklch(49.6% .265 301.924);--color-purple-800:oklch(43.8% .218 303.724);--color-purple-900:oklch(38.1% .176 304.987);--color-purple-950:oklch(29.1% .149 302.717);--color-fuchsia-50:oklch(97.7% .017 320.058);--color-fuchsia-100:oklch(95.2% .037 318.852);--color-fuchsia-200:oklch(90.3% .076 319.62);--color-fuchsia-300:oklch(83.3% .145 321.434);--color-fuchsia-400:oklch(74% .238 322.16);--color-fuchsia-500:oklch(66.7% .295 322.15);--color-fuchsia-600:oklch(59.1% .293 322.896);--color-fuchsia-700:oklch(51.8% .253 323.949);--color-fuchsia-800:oklch(45.2% .211 324.591);--color-fuchsia-900:oklch(40.1% .17 325.612);--color-fuchsia-950:oklch(29.3% .136 325.661);--color-pink-50:oklch(97.1% .014 343.198);--color-pink-100:oklch(94.8% .028 342.258);--color-pink-200:oklch(89.9% .061 343.231);--color-pink-300:oklch(82.3% .12 346.018);--color-pink-400:oklch(71.8% .202 349.761);--color-pink-500:oklch(65.6% .241 354.308);--color-pink-600:oklch(59.2% .249 .584);--color-pink-700:oklch(52.5% .223 3.958);--color-pink-800:oklch(45.9% .187 3.815);--color-pink-900:oklch(40.8% .153 2.432);--color-pink-950:oklch(28.4% .109 3.907);--color-rose-50:oklch(96.9% .015 12.422);--color-rose-100:oklch(94.1% .03 12.58);--color-rose-200:oklch(89.2% .058 10.001);--color-rose-300:oklch(81% .117 11.638);--color-rose-400:oklch(71.2% .194 13.428);--color-rose-500:oklch(64.5% .246 16.439);--color-rose-600:oklch(58.6% .253 17.585);--color-rose-700:oklch(51.4% .222 16.935);--color-rose-800:oklch(45.5% .188 13.697);--color-rose-900:oklch(41% .159 10.272);--color-rose-950:oklch(27.1% .105 12.094);--color-slate-50:oklch(98.4% .003 247.858);--color-slate-100:oklch(96.8% .007 247.896);--color-slate-200:oklch(92.9% .013 255.508);--color-slate-300:oklch(86.9% .022 252.894);--color-slate-400:oklch(70.4% .04 256.788);--color-slate-500:oklch(55.4% .046 257.417);--color-slate-600:oklch(44.6% .043 257.281);--color-slate-700:oklch(37.2% .044 257.287);--color-slate-800:oklch(27.9% .041 260.031);--color-slate-900:oklch(20.8% .042 265.755);--color-slate-950:oklch(12.9% .042 264.695);--color-gray-50:oklch(98.5% .002 247.839);--color-gray-100:oklch(96.7% .003 264.542);--color-gray-200:oklch(92.8% .006 264.531);--color-gray-300:oklch(87.2% .01 258.338);--color-gray-400:oklch(70.7% .022 261.325);--color-gray-500:oklch(55.1% .027 264.364);--color-gray-600:oklch(44.6% .03 256.802);--color-gray-700:oklch(37.3% .034 259.733);--color-gray-800:oklch(27.8% .033 256.848);--color-gray-900:oklch(21% .034 264.665);--color-gray-950:oklch(13% .028 261.692);--color-zinc-50:oklch(98.5% 0 none);--color-zinc-100:oklch(96.7% .001 286.375);--color-zinc-200:oklch(92% .004 286.32);--color-zinc-300:oklch(87.1% .006 286.286);--color-zinc-400:oklch(70.5% .015 286.067);--color-zinc-500:oklch(55.2% .016 285.938);--color-zinc-600:oklch(44.2% .017 285.786);--color-zinc-700:oklch(37% .013 285.805);--color-zinc-800:oklch(27.4% .006 286.033);--color-zinc-900:oklch(21% .006 285.885);--color-zinc-950:oklch(14.1% .005 285.823);--color-neutral-50:oklch(98.5% 0 none);--color-neutral-100:oklch(97% 0 none);--color-neutral-200:oklch(92.2% 0 none);--color-neutral-300:oklch(87% 0 none);--color-neutral-400:oklch(70.8% 0 none);--color-neutral-500:oklch(55.6% 0 none);--color-neutral-600:oklch(43.9% 0 none);--color-neutral-700:oklch(37.1% 0 none);--color-neutral-800:oklch(26.9% 0 none);--color-neutral-900:oklch(20.5% 0 none);--color-neutral-950:oklch(14.5% 0 none);--color-stone-50:oklch(98.5% .001 106.423);--color-stone-100:oklch(97% .001 106.424);--color-stone-200:oklch(92.3% .003 48.717);--color-stone-300:oklch(86.9% .005 56.366);--color-stone-400:oklch(70.9% .01 56.259);--color-stone-500:oklch(55.3% .013 58.071);--color-stone-600:oklch(44.4% .011 73.639);--color-stone-700:oklch(37.4% .01 67.558);--color-stone-800:oklch(26.8% .007 34.298);--color-stone-900:oklch(21.6% .006 56.043);--color-stone-950:oklch(14.7% .004 49.25);--color-black:#000;--color-white:#fff;--spacing:.25rem;--container-xs:20rem;--container-sm:24rem;--container-md:28rem;--container-lg:32rem;--container-xl:36rem;--container-2xl:42rem;--container-3xl:48rem;--container-4xl:56rem;--container-5xl:64rem;--container-6xl:72rem;--container-7xl:80rem;--text-xs:.75rem;--text-xs--line-height:calc(1 / .75);--text-sm:.875rem;--text-sm--line-height:calc(1.25 / .875);--text-base:1rem;--text-base--line-height:calc(1.5 / 1);--text-lg:1.125rem;--text-lg--line-height:calc(1.75 / 1.125);--text-xl:1.25rem;--text-xl--line-height:calc(1.75 / 1.25);--text-2xl:1.5rem;--text-2xl--line-height:calc(2 / 1.5);--text-3xl:1.875rem;--text-3xl--line-height:calc(2.25 / 1.875);--text-4xl:2.25rem;--text-4xl--line-height:calc(2.5 / 2.25);--text-5xl:3rem;--text-5xl--line-height:1;--text-6xl:3.75rem;--text-6xl--line-height:1;--text-7xl:4.5rem;--text-7xl--line-height:1;--text-8xl:6rem;--text-8xl--line-height:1;--text-9xl:8rem;--text-9xl--line-height:1;--font-weight-thin:100;--font-weight-extralight:200;--font-weight-light:300;--font-weight-normal:400;--font-weight-medium:500;--font-weight-semibold:600;--font-weight-bold:700;--font-weight-extrabold:800;--font-weight-black:900;--tracking-tight:-.025em;--tracking-normal:0em;--tracking-wide:.025em;--tracking-wider:.05em;--leading-tight:1.25;--leading-snug:1.375;--leading-normal:1.5;--leading-relaxed:1.625;--radius-xl:.75rem;--radius-2xl:1rem;--ease-out:cubic-bezier(0, 0, .2, 1);--animate-spin:spin 1s linear infinite;--animate-pulse:pulse 2s cubic-bezier(.4, 0, .6, 1) infinite;--blur-sm:8px;--default-transition-duration:.15s;--default-transition-timing-function:cubic-bezier(.4, 0, .2, 1);--default-font-family:var(--clicky-font-sans);--default-font-feature-settings:normal;--default-font-variation-settings:normal;--default-mono-font-family:var(--clicky-font-mono);--default-mono-font-feature-settings:normal;--default-mono-font-variation-settings:normal;--font-sans:var(--clicky-font-sans);--font-mono:var(--clicky-font-mono);--radius-sm:.25rem;--radius-md:.375rem;--radius-lg:.5rem;--spacing-density-1:calc(var(--spacing-unit,.25rem) * 1);--spacing-density-2:calc(var(--spacing-unit,.25rem) * 2);--spacing-density-3:calc(var(--spacing-unit,.25rem) * 3);--spacing-density-4:calc(var(--spacing-unit,.25rem) * 4);--spacing-density-6:calc(var(--spacing-unit,.25rem) * 6);--spacing-density-8:calc(var(--spacing-unit,.25rem) * 8);--spacing-control-h:var(--control-height,2.25rem);--spacing-control-px:var(--control-padding-x,.75rem);--container-8xl:96rem;--container-9xl:120rem;--color-primary:var(--primary);--color-destructive:var(--destructive);--color-placeholder:var(--placeholder)}}@layer base{@supports (not ((-webkit-appearance:-apple-pay-button))) or (contain-intrinsic-size:1px){::placeholder{color:currentColor}@supports (color:color-mix(in lab, red, red)){::placeholder{color:color-mix(in oklab, currentcolor 50%, transparent)}}}*,:after,:before,::backdrop{box-sizing:border-box;border:0 solid;margin:0;padding:0}::file-selector-button{box-sizing:border-box;border:0 solid;margin:0;padding:0}html,:host{-webkit-text-size-adjust:100%;tab-size:4;line-height:1.5;font-family:var(--default-font-family,-apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", \"Noto Sans\", Arial, sans-serif, \"Apple Color Emoji\", \"Segoe UI Emoji\", \"Segoe UI Symbol\", \"Noto Color Emoji\");font-feature-settings:var(--default-font-feature-settings,normal);font-variation-settings:var(--default-font-variation-settings,normal);-webkit-tap-highlight-color:transparent}hr{height:0;color:inherit;border-top-width:1px}abbr:where([title]){-webkit-text-decoration:underline dotted;text-decoration:underline dotted}h1,h2,h3,h4,h5,h6{font-size:inherit;font-weight:inherit}a{color:inherit;-webkit-text-decoration:inherit;-webkit-text-decoration:inherit;-webkit-text-decoration:inherit;-webkit-text-decoration:inherit;text-decoration:inherit}b,strong{font-weight:bolder}code,kbd,samp,pre{font-family:var(--default-mono-font-family,ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, \"Liberation Mono\", \"Courier New\", monospace);font-feature-settings:var(--default-mono-font-feature-settings,normal);font-variation-settings:var(--default-mono-font-variation-settings,normal);font-size:1em}small{font-size:80%}sub,sup{vertical-align:baseline;font-size:75%;line-height:0;position:relative}sub{bottom:-.25em}sup{top:-.5em}table{text-indent:0;border-color:inherit;border-collapse:collapse}:-moz-focusring:where(:not(iframe)){outline:auto}progress{vertical-align:baseline}summary{display:list-item}ol,ul,menu{list-style:none}img,svg,video,canvas,audio,iframe,embed,object{vertical-align:middle;display:block}img,video{max-width:100%;height:auto}button,input,select,optgroup,textarea{font:inherit;font-feature-settings:inherit;font-variation-settings:inherit;letter-spacing:inherit;color:inherit;opacity:1;background-color:#0000;border-radius:0}::file-selector-button{font:inherit;font-feature-settings:inherit;font-variation-settings:inherit;letter-spacing:inherit;color:inherit;opacity:1;background-color:#0000;border-radius:0}:where(select:is([multiple],[size])) optgroup{font-weight:bolder}:where(select:is([multiple],[size])) optgroup option{padding-inline-start:20px}::file-selector-button{margin-inline-end:4px}::placeholder{opacity:1}@supports (not ((-webkit-appearance:-apple-pay-button))) or (contain-intrinsic-size:1px){::placeholder{color:currentColor}@supports (color:color-mix(in lab, red, red)){::placeholder{color:color-mix(in oklab, currentcolor 50%, transparent)}}}textarea{resize:vertical}::-webkit-search-decoration{-webkit-appearance:none}::-webkit-date-and-time-value{min-height:1lh;text-align:inherit}::-webkit-datetime-edit{display:inline-flex}::-webkit-datetime-edit-fields-wrapper{padding:0}::-webkit-datetime-edit{padding-block:0}::-webkit-datetime-edit-year-field{padding-block:0}::-webkit-datetime-edit-month-field{padding-block:0}::-webkit-datetime-edit-day-field{padding-block:0}::-webkit-datetime-edit-hour-field{padding-block:0}::-webkit-datetime-edit-minute-field{padding-block:0}::-webkit-datetime-edit-second-field{padding-block:0}::-webkit-datetime-edit-millisecond-field{padding-block:0}::-webkit-datetime-edit-meridiem-field{padding-block:0}::-webkit-calendar-picker-indicator{line-height:1}:-moz-ui-invalid{box-shadow:none}button,input:where([type=button],[type=reset],[type=submit]){appearance:button}::file-selector-button{appearance:button}::-webkit-inner-spin-button{height:auto}::-webkit-outer-spin-button{height:auto}[hidden]:where(:not([hidden=until-found])){display:none!important}*{border-color:var(--border)}}@layer components;@layer utilities{.end-1{inset-inline-end:calc(var(--spacing) * 1)}@media (width>=1920px){.container{max-width:1920px}}@media (width>=40rem){.container{max-width:40rem}}@media (width>=48rem){.container{max-width:48rem}}@media (width>=64rem){.container{max-width:64rem}}@media (width>=80rem){.container{max-width:80rem}}@media (width>=96rem){.container{max-width:96rem}}.max-h-\\[45\\%\\]{max-height:45%}.shrink{flex-shrink:1}.grow{flex-grow:1}.border-border{border-color:var(--border)}.border-primary{border-color:var(--primary)}.bg-background{background-color:var(--background)}.bg-muted-foreground,.bg-muted-foreground\\/40{background-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.bg-muted-foreground\\/40{background-color:color-mix(in oklab, var(--muted-foreground) 40%, transparent)}.bg-muted-foreground\\/60{background-color:color-mix(in oklab, var(--muted-foreground) 60%, transparent)}.bg-muted\\/30{background-color:color-mix(in oklab, var(--muted) 30%, transparent)}.bg-muted\\/50{background-color:color-mix(in oklab, var(--muted) 50%, transparent)}}.bg-primary\\/5{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/5{background-color:color-mix(in oklab, var(--primary) 5%, transparent)}.bg-primary\\/10{background-color:color-mix(in oklab, var(--primary) 10%, transparent)}}.pr-density-4{padding-right:calc(var(--spacing-unit) * 4)}.text-foreground{color:var(--foreground)}.text-muted-foreground{color:var(--muted-foreground)}.outline{outline-style:var(--tw-outline-style);outline-width:1px}@media (hover:hover){.hover\\:bg-muted:hover{background-color:var(--muted)}.hover\\:underline:hover{text-decoration-line:underline}}@media (width>=40rem){.sm\\:grid-cols-5{grid-template-columns:repeat(5,minmax(0,1fr))}}@media (width>=64rem){.lg\\:inline{display:inline}}.dark\\:bg-red-500\\/10:where([data-theme=dark],[data-theme=dark] *){background-color:#fb2c361a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-red-500\\/10:where([data-theme=dark],[data-theme=dark] *){background-color:color-mix(in oklab, var(--color-red-500) 10%, transparent)}}.dark\\:text-green-400:where([data-theme=dark],[data-theme=dark] *){color:var(--color-green-400)}.dark\\:text-red-300:where([data-theme=dark],[data-theme=dark] *){color:var(--color-red-300)}.dark\\:text-red-400:where([data-theme=dark],[data-theme=dark] *){color:var(--color-red-400)}.\\@container\\/app-content{container:app-content/inline-size}.\\@container\\/app-shell{container:app-shell/inline-size}.\\@container\\/combobox-options{container:combobox-options/inline-size}.\\@container\\/endpoint{container:endpoint/inline-size}.\\@container\\/error-details{container:error-details/inline-size}.\\@container\\/secret{container:secret/inline-size}.\\@container\\/session-viewer{container:session-viewer/inline-size}.\\@container\\/split-pane{container:split-pane/inline-size}.\\@container\\/workspace{container:workspace/inline-size}.\\@container{container-type:inline-size}.\\[container-type\\:size\\]{container-type:size}.pointer-events-auto{pointer-events:auto}.pointer-events-none{pointer-events:none}.collapse{visibility:collapse}.invisible{visibility:hidden}.visible{visibility:visible}.sr-only{clip-path:inset(50%);white-space:nowrap;border-width:0;width:1px;height:1px;margin:-1px;padding:0;position:absolute;overflow:hidden}.absolute{position:absolute}.fixed{position:fixed}.relative{position:relative}.static{position:static}.sticky{position:sticky}.inset-0{inset:0}.inset-4{inset:calc(var(--spacing) * 4)}.inset-x-0{inset-inline:0}.inset-x-3{inset-inline:calc(var(--spacing) * 3)}.inset-x-4{inset-inline:calc(var(--spacing) * 4)}.inset-y-0{inset-block:0}.-top-0\\.5{top:calc(var(--spacing) * -.5)}.-top-1{top:calc(var(--spacing) * -1)}.top-0{top:0}.top-0\\.5{top:calc(var(--spacing) * .5)}.top-1{top:var(--spacing)}.top-1\\.5{top:calc(var(--spacing) * 1.5)}.top-1\\/2{top:50%}.top-2{top:calc(var(--spacing) * 2)}.top-3{top:calc(var(--spacing) * 3)}.top-9{top:calc(var(--spacing) * 9)}.top-12{top:calc(var(--spacing) * 12)}.top-\\[2px\\]{top:2px}.top-\\[22px\\]{top:22px}.top-\\[calc\\(100\\%\\+0\\.375rem\\)\\]{top:calc(100% + .375rem)}.top-full{top:100%}.-right-0\\.5{right:calc(var(--spacing) * -.5)}.-right-1{right:calc(var(--spacing) * -1)}.right-0{right:0}.right-1{right:var(--spacing)}.right-1\\.5{right:calc(var(--spacing) * 1.5)}.right-2{right:calc(var(--spacing) * 2)}.right-3{right:calc(var(--spacing) * 3)}.right-4{right:calc(var(--spacing) * 4)}.right-6{right:calc(var(--spacing) * 6)}.right-7{right:calc(var(--spacing) * 7)}.right-14{right:calc(var(--spacing) * 14)}.right-\\[3\\.75rem\\]{right:3.75rem}.right-auto{right:auto}.right-full{right:100%}.-bottom-0\\.5{bottom:calc(var(--spacing) * -.5)}.-bottom-1{bottom:calc(var(--spacing) * -1)}.bottom-0{bottom:0}.bottom-1{bottom:var(--spacing)}.bottom-2{bottom:calc(var(--spacing) * 2)}.bottom-2\\.5{bottom:calc(var(--spacing) * 2.5)}.bottom-3{bottom:calc(var(--spacing) * 3)}.bottom-4{bottom:calc(var(--spacing) * 4)}.bottom-6{bottom:calc(var(--spacing) * 6)}.bottom-\\[2px\\]{bottom:2px}.bottom-\\[calc\\(1rem\\+env\\(safe-area-inset-bottom\\)\\)\\]{bottom:calc(1rem + env(safe-area-inset-bottom))}.bottom-density-3{bottom:var(--spacing-density-3)}.bottom-full{bottom:100%}.-left-0\\.5{left:calc(var(--spacing) * -.5)}.left-0{left:0}.left-0\\.5{left:calc(var(--spacing) * .5)}.left-1{left:var(--spacing)}.left-1\\.5{left:calc(var(--spacing) * 1.5)}.left-1\\/2{left:50%}.left-2{left:calc(var(--spacing) * 2)}.left-\\[10px\\]{left:10px}.left-auto{left:auto}.left-density-3{left:var(--spacing-density-3)}.left-full{left:100%}.isolate{isolation:isolate}.z-0{z-index:0}.z-1{z-index:1}.z-10{z-index:10}.z-20{z-index:20}.z-30{z-index:30}.z-50{z-index:50}.z-\\[1\\]{z-index:1}.z-\\[30\\]{z-index:30}.z-\\[999\\]{z-index:999}.z-\\[9999\\]{z-index:9999}.z-\\[N\\]{z-index:N}.z-\\[…\\]{z-index:…}.order-1{order:1}.order-2{order:2}.order-3{order:3}.order-4{order:4}.order-7{order:7}.order-last{order:9999}.\\[grid-column\\:span_var\\(--stat-last-span-sm\\)\\/span_var\\(--stat-last-span-sm\\)\\]{grid-column:span var(--stat-last-span-sm)/span var(--stat-last-span-sm)}.col-span-2{grid-column:span 2/span 2}.col-span-3{grid-column:span 3/span 3}.col-span-full{grid-column:1/-1}.col-start-1{grid-column-start:1}.col-start-2{grid-column-start:2}.row-0{grid-row:0}.row-1{grid-row:1}.row-7{grid-row:7}.row-start-1{grid-row-start:1}.\\!container{width:100%!important}@media (width>=40rem){.\\!container{max-width:40rem!important}}@media (width>=48rem){.\\!container{max-width:48rem!important}}@media (width>=64rem){.\\!container{max-width:64rem!important}}@media (width>=80rem){.\\!container{max-width:80rem!important}}@media (width>=96rem){.\\!container{max-width:96rem!important}}.container{width:100%}@media (width>=40rem){.container{max-width:40rem}}@media (width>=48rem){.container{max-width:48rem}}@media (width>=64rem){.container{max-width:64rem}}@media (width>=80rem){.container{max-width:80rem}}@media (width>=96rem){.container{max-width:96rem}}.-m-0\\.5{margin:calc(var(--spacing) * -.5)}.-m-4{margin:calc(var(--spacing) * -4)}.m-0{margin:0}.m-1{margin:var(--spacing)}.m-2{margin:calc(var(--spacing) * 2)}.m-3{margin:calc(var(--spacing) * 3)}.m-4{margin:calc(var(--spacing) * 4)}.m-9{margin:calc(var(--spacing) * 9)}.m-density-2{margin:var(--spacing-density-2)}.m-density-4{margin:var(--spacing-density-4)}.-mx-1{margin-inline:calc(var(--spacing) * -1)}.-mx-2{margin-inline:calc(var(--spacing) * -2)}.-mx-density-4{margin-inline:calc(var(--spacing-density-4) * -1)}.mx-0\\.5{margin-inline:calc(var(--spacing) * .5)}.mx-1{margin-inline:var(--spacing)}.mx-2{margin-inline:calc(var(--spacing) * 2)}.mx-3{margin-inline:calc(var(--spacing) * 3)}.mx-auto{margin-inline:auto}.mx-density-1{margin-inline:var(--spacing-density-1)}.mx-density-4{margin-inline:var(--spacing-density-4)}.-my-1{margin-block:calc(var(--spacing) * -1)}.-my-1\\.5{margin-block:calc(var(--spacing) * -1.5)}.-my-density-3{margin-block:calc(var(--spacing-density-3) * -1)}.my-0{margin-block:0}.my-1{margin-block:var(--spacing)}.my-2{margin-block:calc(var(--spacing) * 2)}.my-3{margin-block:calc(var(--spacing) * 3)}.my-4{margin-block:calc(var(--spacing) * 4)}.ms-1{margin-inline-start:var(--spacing)}.-mt-0\\.5{margin-top:calc(var(--spacing) * -.5)}.-mt-1{margin-top:calc(var(--spacing) * -1)}.-mt-\\[8px\\]{margin-top:-8px}.-mt-density-4{margin-top:calc(var(--spacing-density-4) * -1)}.mt-0{margin-top:0}.mt-0\\.5{margin-top:calc(var(--spacing) * .5)}.mt-1{margin-top:var(--spacing)}.mt-1\\.5{margin-top:calc(var(--spacing) * 1.5)}.mt-2{margin-top:calc(var(--spacing) * 2)}.mt-2\\.5{margin-top:calc(var(--spacing) * 2.5)}.mt-3{margin-top:calc(var(--spacing) * 3)}.mt-3\\.5{margin-top:calc(var(--spacing) * 3.5)}.mt-4{margin-top:calc(var(--spacing) * 4)}.mt-5{margin-top:calc(var(--spacing) * 5)}.mt-8{margin-top:calc(var(--spacing) * 8)}.mt-auto{margin-top:auto}.mt-density-1{margin-top:var(--spacing-density-1)}.mt-density-2{margin-top:var(--spacing-density-2)}.mt-density-3{margin-top:var(--spacing-density-3)}.mt-density-4{margin-top:var(--spacing-density-4)}.mt-density-6{margin-top:var(--spacing-density-6)}.mt-px{margin-top:1px}.-mr-1{margin-right:calc(var(--spacing) * -1)}.-mr-px{margin-right:-1px}.mr-0\\.5{margin-right:calc(var(--spacing) * .5)}.mr-1{margin-right:var(--spacing)}.mr-1\\.5{margin-right:calc(var(--spacing) * 1.5)}.mr-2{margin-right:calc(var(--spacing) * 2)}.mr-2\\.5{margin-right:calc(var(--spacing) * 2.5)}.mr-12{margin-right:calc(var(--spacing) * 12)}.mr-auto{margin-right:auto}.-mb-px{margin-bottom:-1px}.mb-0\\.5{margin-bottom:calc(var(--spacing) * .5)}.mb-1{margin-bottom:var(--spacing)}.mb-1\\.5{margin-bottom:calc(var(--spacing) * 1.5)}.mb-2{margin-bottom:calc(var(--spacing) * 2)}.mb-2\\.5{margin-bottom:calc(var(--spacing) * 2.5)}.mb-3{margin-bottom:calc(var(--spacing) * 3)}.mb-4{margin-bottom:calc(var(--spacing) * 4)}.mb-5{margin-bottom:calc(var(--spacing) * 5)}.mb-density-1{margin-bottom:var(--spacing-density-1)}.mb-density-2{margin-bottom:var(--spacing-density-2)}.mb-density-3{margin-bottom:var(--spacing-density-3)}.mb-density-4{margin-bottom:var(--spacing-density-4)}.mb-density-6{margin-bottom:var(--spacing-density-6)}.-ml-1\\.5{margin-left:calc(var(--spacing) * -1.5)}.-ml-2{margin-left:calc(var(--spacing) * -2)}.-ml-px{margin-left:-1px}.ml-0\\.5{margin-left:calc(var(--spacing) * .5)}.ml-1{margin-left:var(--spacing)}.ml-1\\.5{margin-left:calc(var(--spacing) * 1.5)}.ml-2{margin-left:calc(var(--spacing) * 2)}.ml-2\\.5{margin-left:calc(var(--spacing) * 2.5)}.ml-3{margin-left:calc(var(--spacing) * 3)}.ml-3\\.5{margin-left:calc(var(--spacing) * 3.5)}.ml-4{margin-left:calc(var(--spacing) * 4)}.ml-5{margin-left:calc(var(--spacing) * 5)}.ml-12{margin-left:calc(var(--spacing) * 12)}.ml-\\[4\\.5rem\\]{margin-left:4.5rem}.ml-\\[5px\\]{margin-left:5px}.ml-\\[18px\\]{margin-left:18px}.ml-auto{margin-left:auto}.ml-density-2{margin-left:var(--spacing-density-2)}.ml-px{margin-left:1px}.line-clamp-1{-webkit-line-clamp:1;-webkit-box-orient:vertical;display:-webkit-box;overflow:hidden}.line-clamp-2{-webkit-line-clamp:2;-webkit-box-orient:vertical;display:-webkit-box;overflow:hidden}.line-clamp-4{-webkit-line-clamp:4;-webkit-box-orient:vertical;display:-webkit-box;overflow:hidden}.line-clamp-5{-webkit-line-clamp:5;-webkit-box-orient:vertical;display:-webkit-box;overflow:hidden}.\\!inline{display:inline!important}.\\[display\\:-webkit-box\\]{display:-webkit-box}.block{display:block}.contents{display:contents}.flex{display:flex}.grid{display:grid}.hidden{display:none}.inline{display:inline}.inline-block{display:inline-block}.inline-flex{display:inline-flex}.list-item{display:list-item}.table{display:table}.aspect-square{aspect-ratio:1}.size-0{width:0;height:0}.size-1\\.5{width:calc(var(--spacing) * 1.5);height:calc(var(--spacing) * 1.5)}.size-2{width:calc(var(--spacing) * 2);height:calc(var(--spacing) * 2)}.size-3{width:calc(var(--spacing) * 3);height:calc(var(--spacing) * 3)}.size-3\\.5{width:calc(var(--spacing) * 3.5);height:calc(var(--spacing) * 3.5)}.size-4{width:calc(var(--spacing) * 4);height:calc(var(--spacing) * 4)}.size-5{width:calc(var(--spacing) * 5);height:calc(var(--spacing) * 5)}.size-6{width:calc(var(--spacing) * 6);height:calc(var(--spacing) * 6)}.size-7{width:calc(var(--spacing) * 7);height:calc(var(--spacing) * 7)}.size-8{width:calc(var(--spacing) * 8);height:calc(var(--spacing) * 8)}.size-9{width:calc(var(--spacing) * 9);height:calc(var(--spacing) * 9)}.size-10{width:calc(var(--spacing) * 10);height:calc(var(--spacing) * 10)}.size-12{width:calc(var(--spacing) * 12);height:calc(var(--spacing) * 12)}.size-\\[12px\\]{width:12px;height:12px}.size-\\[16px\\]{width:16px;height:16px}.size-\\[clamp\\(…\\)\\]{width:clamp(…);height:clamp(…)}.size-control-h{width:var(--spacing-control-h);height:var(--spacing-control-h)}.size-full{width:100%;height:100%}.h-0{height:0}.h-0\\.5{height:calc(var(--spacing) * .5)}.h-1{height:var(--spacing)}.h-1\\.5{height:calc(var(--spacing) * 1.5)}.h-2{height:calc(var(--spacing) * 2)}.h-2\\.5{height:calc(var(--spacing) * 2.5)}.h-3{height:calc(var(--spacing) * 3)}.h-3\\.5{height:calc(var(--spacing) * 3.5)}.h-4{height:calc(var(--spacing) * 4)}.h-5{height:calc(var(--spacing) * 5)}.h-6{height:calc(var(--spacing) * 6)}.h-7{height:calc(var(--spacing) * 7)}.h-8{height:calc(var(--spacing) * 8)}.h-9{height:calc(var(--spacing) * 9)}.h-10{height:calc(var(--spacing) * 10)}.h-11{height:calc(var(--spacing) * 11)}.h-12{height:calc(var(--spacing) * 12)}.h-14{height:calc(var(--spacing) * 14)}.h-16{height:calc(var(--spacing) * 16)}.h-20{height:calc(var(--spacing) * 20)}.h-24{height:calc(var(--spacing) * 24)}.h-28{height:calc(var(--spacing) * 28)}.h-40{height:calc(var(--spacing) * 40)}.h-64{height:calc(var(--spacing) * 64)}.h-72{height:calc(var(--spacing) * 72)}.h-80{height:calc(var(--spacing) * 80)}.h-96{height:calc(var(--spacing) * 96)}.h-\\[3px\\]{height:3px}.h-\\[8px\\]{height:8px}.h-\\[16px\\]{height:16px}.h-\\[17px\\]{height:17px}.h-\\[18px\\]{height:18px}.h-\\[20px\\]{height:20px}.h-\\[21px\\]{height:21px}.h-\\[24px\\]{height:24px}.h-\\[28rem\\]{height:28rem}.h-\\[30rem\\]{height:30rem}.h-\\[32rem\\]{height:32rem}.h-\\[33px\\]{height:33px}.h-\\[34\\%\\]{height:34%}.h-\\[34px\\]{height:34px}.h-\\[36rem\\]{height:36rem}.h-\\[42rem\\]{height:42rem}.h-\\[60vh\\]{height:60vh}.h-\\[70vh\\]{height:70vh}.h-\\[75vh\\]{height:75vh}.h-\\[80vh\\]{height:80vh}.h-\\[92vh\\]{height:92vh}.h-\\[95vh\\]{height:95vh}.h-\\[180px\\]{height:180px}.h-\\[240px\\]{height:240px}.h-\\[280px\\]{height:280px}.h-\\[320px\\]{height:320px}.h-\\[360px\\]{height:360px}.h-\\[400px\\]{height:400px}.h-\\[420px\\]{height:420px}.h-\\[440px\\]{height:440px}.h-\\[460px\\]{height:460px}.h-\\[480px\\]{height:480px}.h-\\[500px\\]{height:500px}.h-\\[520px\\]{height:520px}.h-\\[560px\\]{height:560px}.h-\\[600px\\]{height:600px}.h-\\[640px\\]{height:640px}.h-\\[680px\\]{height:680px}.h-\\[720px\\]{height:720px}.h-\\[1000px\\]{height:1000px}.h-\\[calc\\(100dvh-2rem\\)\\]{height:calc(100dvh - 2rem)}.h-\\[calc\\(100vh-15rem\\)\\]{height:calc(100vh - 15rem)}.h-\\[min\\(32rem\\,52vh\\)\\]{height:min(32rem,52vh)}.h-\\[min\\(58vh\\,34rem\\)\\]{height:min(58vh,34rem)}.h-\\[min\\(800px\\,90dvh\\)\\]{height:min(800px,90dvh)}.h-auto{height:auto}.h-control-h{height:var(--spacing-control-h)}.h-density-4{height:var(--spacing-density-4)}.h-dvh{height:100dvh}.h-fit{height:fit-content}.h-full{height:100%}.h-px{height:1px}.h-screen{height:100vh}.max-h-12{max-height:calc(var(--spacing) * 12)}.max-h-20{max-height:calc(var(--spacing) * 20)}.max-h-24{max-height:calc(var(--spacing) * 24)}.max-h-32{max-height:calc(var(--spacing) * 32)}.max-h-40{max-height:calc(var(--spacing) * 40)}.max-h-48{max-height:calc(var(--spacing) * 48)}.max-h-56{max-height:calc(var(--spacing) * 56)}.max-h-60{max-height:calc(var(--spacing) * 60)}.max-h-64{max-height:calc(var(--spacing) * 64)}.max-h-72{max-height:calc(var(--spacing) * 72)}.max-h-80{max-height:calc(var(--spacing) * 80)}.max-h-96{max-height:calc(var(--spacing) * 96)}.max-h-\\[26rem\\]{max-height:26rem}.max-h-\\[32rem\\]{max-height:32rem}.max-h-\\[33dvh\\]{max-height:33dvh}.max-h-\\[40\\%\\]{max-height:40%}.max-h-\\[40vh\\]{max-height:40vh}.max-h-\\[46rem\\]{max-height:46rem}.max-h-\\[50dvh\\]{max-height:50dvh}.max-h-\\[70vh\\]{max-height:70vh}.max-h-\\[80\\%\\]{max-height:80%}.max-h-\\[360px\\]{max-height:360px}.max-h-\\[520px\\]{max-height:520px}.max-h-\\[calc\\(100vh-1rem\\)\\]{max-height:calc(100vh - 1rem)}.max-h-\\[min\\(32rem\\,calc\\(100vh-2rem\\)\\)\\]{max-height:min(32rem,100vh - 2rem)}.max-h-\\[min\\(34rem\\,calc\\(100vh-2rem\\)\\)\\]{max-height:min(34rem,100vh - 2rem)}.max-h-\\[min\\(42rem\\,calc\\(100vh-5rem\\)\\)\\]{max-height:min(42rem,100vh - 5rem)}.max-h-\\[var\\(--rg-edit-list-max-h\\)\\]{max-height:var(--rg-edit-list-max-h)}.max-h-full{max-height:100%}.min-h-0{min-height:0}.min-h-4{min-height:calc(var(--spacing) * 4)}.min-h-5{min-height:calc(var(--spacing) * 5)}.min-h-6{min-height:calc(var(--spacing) * 6)}.min-h-7{min-height:calc(var(--spacing) * 7)}.min-h-8{min-height:calc(var(--spacing) * 8)}.min-h-9{min-height:calc(var(--spacing) * 9)}.min-h-10{min-height:calc(var(--spacing) * 10)}.min-h-11{min-height:calc(var(--spacing) * 11)}.min-h-12{min-height:calc(var(--spacing) * 12)}.min-h-14{min-height:calc(var(--spacing) * 14)}.min-h-16{min-height:calc(var(--spacing) * 16)}.min-h-20{min-height:calc(var(--spacing) * 20)}.min-h-28{min-height:calc(var(--spacing) * 28)}.min-h-32{min-height:calc(var(--spacing) * 32)}.min-h-36{min-height:calc(var(--spacing) * 36)}.min-h-40{min-height:calc(var(--spacing) * 40)}.min-h-44{min-height:calc(var(--spacing) * 44)}.min-h-48{min-height:calc(var(--spacing) * 48)}.min-h-56{min-height:calc(var(--spacing) * 56)}.min-h-64{min-height:calc(var(--spacing) * 64)}.min-h-72{min-height:calc(var(--spacing) * 72)}.min-h-80{min-height:calc(var(--spacing) * 80)}.min-h-96{min-height:calc(var(--spacing) * 96)}.min-h-\\[4\\.25rem\\]{min-height:4.25rem}.min-h-\\[5rem\\]{min-height:5rem}.min-h-\\[7rem\\]{min-height:7rem}.min-h-\\[8rem\\]{min-height:8rem}.min-h-\\[12rem\\]{min-height:12rem}.min-h-\\[14rem\\]{min-height:14rem}.min-h-\\[18px\\]{min-height:18px}.min-h-\\[20rem\\]{min-height:20rem}.min-h-\\[24rem\\]{min-height:24rem}.min-h-\\[28px\\]{min-height:28px}.min-h-\\[32rem\\]{min-height:32rem}.min-h-\\[34rem\\]{min-height:34rem}.min-h-\\[48px\\]{min-height:48px}.min-h-\\[640px\\]{min-height:640px}.min-h-control-h{min-height:var(--spacing-control-h)}.min-h-dvh{min-height:100dvh}.min-h-px{min-height:1px}.min-h-screen{min-height:100vh}.w-0{width:0}.w-0\\.5{width:calc(var(--spacing) * .5)}.w-1{width:var(--spacing)}.w-1\\.5{width:calc(var(--spacing) * 1.5)}.w-1\\/2{width:50%}.w-1\\/3{width:33.3333%}.w-2{width:calc(var(--spacing) * 2)}.w-2\\.5{width:calc(var(--spacing) * 2.5)}.w-2\\/3{width:66.6667%}.w-3{width:calc(var(--spacing) * 3)}.w-3\\.5{width:calc(var(--spacing) * 3.5)}.w-4{width:calc(var(--spacing) * 4)}.w-5{width:calc(var(--spacing) * 5)}.w-5\\/6{width:83.3333%}.w-6{width:calc(var(--spacing) * 6)}.w-7{width:calc(var(--spacing) * 7)}.w-8{width:calc(var(--spacing) * 8)}.w-9{width:calc(var(--spacing) * 9)}.w-10{width:calc(var(--spacing) * 10)}.w-12{width:calc(var(--spacing) * 12)}.w-14{width:calc(var(--spacing) * 14)}.w-16{width:calc(var(--spacing) * 16)}.w-20{width:calc(var(--spacing) * 20)}.w-24{width:calc(var(--spacing) * 24)}.w-28{width:calc(var(--spacing) * 28)}.w-32{width:calc(var(--spacing) * 32)}.w-36{width:calc(var(--spacing) * 36)}.w-40{width:calc(var(--spacing) * 40)}.w-44{width:calc(var(--spacing) * 44)}.w-48{width:calc(var(--spacing) * 48)}.w-52{width:calc(var(--spacing) * 52)}.w-56{width:calc(var(--spacing) * 56)}.w-64{width:calc(var(--spacing) * 64)}.w-72{width:calc(var(--spacing) * 72)}.w-80{width:calc(var(--spacing) * 80)}.w-96{width:calc(var(--spacing) * 96)}.w-\\[1\\.25em\\]{width:1.25em}.w-\\[2px\\]{width:2px}.w-\\[3px\\]{width:3px}.w-\\[5\\.5rem\\]{width:5.5rem}.w-\\[5\\.25rem\\]{width:5.25rem}.w-\\[15\\.5rem\\]{width:15.5rem}.w-\\[16ch\\]{width:16ch}.w-\\[18\\%\\]{width:18%}.w-\\[21px\\]{width:21px}.w-\\[22\\.5rem\\]{width:22.5rem}.w-\\[22rem\\]{width:22rem}.w-\\[24rem\\]{width:24rem}.w-\\[28\\%\\]{width:28%}.w-\\[28rem\\]{width:28rem}.w-\\[30rem\\]{width:30rem}.w-\\[34px\\]{width:34px}.w-\\[34rem\\]{width:34rem}.w-\\[42rem\\]{width:42rem}.w-\\[45rem\\]{width:45rem}.w-\\[52px\\]{width:52px}.w-\\[52rem\\]{width:52rem}.w-\\[58rem\\]{width:58rem}.w-\\[64rem\\]{width:64rem}.w-\\[200px\\]{width:200px}.w-\\[320px\\]{width:320px}.w-\\[360px\\]{width:360px}.w-\\[390px\\]{width:390px}.w-\\[420px\\]{width:420px}.w-\\[480px\\]{width:480px}.w-\\[560px\\]{width:560px}.w-\\[640px\\]{width:640px}.w-\\[720px\\]{width:720px}.w-\\[760px\\]{width:760px}.w-\\[880px\\]{width:880px}.w-\\[960px\\]{width:960px}.w-\\[calc\\(100\\%\\+16px\\)\\]{width:calc(100% + 16px)}.w-\\[calc\\(var\\(--dt-viewport-w\\)-2\\*var\\(--spacing-density-3\\)\\)\\]{width:calc(var(--dt-viewport-w) - 2 * var(--spacing-density-3))}.w-\\[min\\(20rem\\,85vw\\)\\]{width:min(20rem,85vw)}.w-\\[min\\(22rem\\,calc\\(100vw-2rem\\)\\)\\]{width:min(22rem,100vw - 2rem)}.w-\\[min\\(30rem\\,calc\\(100vw-2rem\\)\\)\\]{width:min(30rem,100vw - 2rem)}.w-\\[min\\(34rem\\,calc\\(100vw-2rem\\)\\)\\]{width:min(34rem,100vw - 2rem)}.w-\\[min\\(36rem\\,calc\\(100vw-1rem\\)\\)\\]{width:min(36rem,100vw - 1rem)}.w-\\[min\\(38rem\\,calc\\(100vw-4rem\\)\\)\\]{width:min(38rem,100vw - 4rem)}.w-\\[min\\(42rem\\,calc\\(100vw-1rem\\)\\)\\]{width:min(42rem,100vw - 1rem)}.w-\\[min\\(42rem\\,calc\\(100vw-2rem\\)\\)\\]{width:min(42rem,100vw - 2rem)}.w-\\[min\\(100\\%\\,28rem\\)\\]{width:min(100%,28rem)}.w-\\[min\\(320px\\,calc\\(100vw-16px\\)\\)\\]{width:min(320px,100vw - 16px)}.w-auto{width:auto}.w-control-h{width:var(--spacing-control-h)}.w-fit{width:fit-content}.w-full{width:100%}.w-max{width:max-content}.w-px{width:1px}.w-screen{width:100vw}.max-w-2xl{max-width:var(--container-2xl)}.max-w-3xl{max-width:var(--container-3xl)}.max-w-4{max-width:calc(var(--spacing) * 4)}.max-w-4xl{max-width:var(--container-4xl)}.max-w-5xl{max-width:var(--container-5xl)}.max-w-6xl{max-width:var(--container-6xl)}.max-w-7xl{max-width:var(--container-7xl)}.max-w-24{max-width:calc(var(--spacing) * 24)}.max-w-32{max-width:calc(var(--spacing) * 32)}.max-w-40{max-width:calc(var(--spacing) * 40)}.max-w-48{max-width:calc(var(--spacing) * 48)}.max-w-52{max-width:calc(var(--spacing) * 52)}.max-w-56{max-width:calc(var(--spacing) * 56)}.max-w-64{max-width:calc(var(--spacing) * 64)}.max-w-72{max-width:calc(var(--spacing) * 72)}.max-w-80{max-width:calc(var(--spacing) * 80)}.max-w-\\[8\\.5rem\\]{max-width:8.5rem}.max-w-\\[9\\.5rem\\]{max-width:9.5rem}.max-w-\\[12rem\\]{max-width:12rem}.max-w-\\[15rem\\]{max-width:15rem}.max-w-\\[16rem\\]{max-width:16rem}.max-w-\\[18rem\\]{max-width:18rem}.max-w-\\[20rem\\]{max-width:20rem}.max-w-\\[22rem\\]{max-width:22rem}.max-w-\\[24rem\\]{max-width:24rem}.max-w-\\[25\\%\\]{max-width:25%}.max-w-\\[28rem\\]{max-width:28rem}.max-w-\\[30\\%\\]{max-width:30%}.max-w-\\[32rem\\]{max-width:32rem}.max-w-\\[40\\%\\]{max-width:40%}.max-w-\\[40rem\\]{max-width:40rem}.max-w-\\[45\\%\\]{max-width:45%}.max-w-\\[48rem\\]{max-width:48rem}.max-w-\\[50\\%\\]{max-width:50%}.max-w-\\[55\\%\\]{max-width:55%}.max-w-\\[90vw\\]{max-width:90vw}.max-w-\\[95\\%\\]{max-width:95%}.max-w-\\[95vw\\]{max-width:95vw}.max-w-\\[100px\\]{max-width:100px}.max-w-\\[180px\\]{max-width:180px}.max-w-\\[200px\\]{max-width:200px}.max-w-\\[280px\\]{max-width:280px}.max-w-\\[320px\\]{max-width:320px}.max-w-\\[820px\\]{max-width:820px}.max-w-\\[calc\\(100\\%-1rem\\)\\]{max-width:calc(100% - 1rem)}.max-w-\\[calc\\(100vw-1rem\\)\\]{max-width:calc(100vw - 1rem)}.max-w-\\[calc\\(100vw-2rem\\)\\]{max-width:calc(100vw - 2rem)}.max-w-\\[min\\(40ch\\,100\\%\\)\\]{max-width:min(40ch,100%)}.max-w-\\[min\\(50ch\\,100\\%\\)\\]{max-width:min(50ch,100%)}.max-w-full{max-width:100%}.max-w-lg{max-width:var(--container-lg)}.max-w-md{max-width:var(--container-md)}.max-w-none{max-width:none}.max-w-sm{max-width:var(--container-sm)}.max-w-xl{max-width:var(--container-xl)}.max-w-xs{max-width:var(--container-xs)}.min-w-0{min-width:0}.min-w-4{min-width:calc(var(--spacing) * 4)}.min-w-8{min-width:calc(var(--spacing) * 8)}.min-w-9{min-width:calc(var(--spacing) * 9)}.min-w-10{min-width:calc(var(--spacing) * 10)}.min-w-12{min-width:calc(var(--spacing) * 12)}.min-w-20{min-width:calc(var(--spacing) * 20)}.min-w-24{min-width:calc(var(--spacing) * 24)}.min-w-28{min-width:calc(var(--spacing) * 28)}.min-w-32{min-width:calc(var(--spacing) * 32)}.min-w-36{min-width:calc(var(--spacing) * 36)}.min-w-40{min-width:calc(var(--spacing) * 40)}.min-w-44{min-width:calc(var(--spacing) * 44)}.min-w-48{min-width:calc(var(--spacing) * 48)}.min-w-56{min-width:calc(var(--spacing) * 56)}.min-w-64{min-width:calc(var(--spacing) * 64)}.min-w-72{min-width:calc(var(--spacing) * 72)}.min-w-\\[3px\\]{min-width:3px}.min-w-\\[8rem\\]{min-width:8rem}.min-w-\\[9rem\\]{min-width:9rem}.min-w-\\[11rem\\]{min-width:11rem}.min-w-\\[12rem\\]{min-width:12rem}.min-w-\\[14rem\\]{min-width:14rem}.min-w-\\[16px\\]{min-width:16px}.min-w-\\[16rem\\]{min-width:16rem}.min-w-\\[18px\\]{min-width:18px}.min-w-\\[18rem\\]{min-width:18rem}.min-w-\\[36px\\]{min-width:36px}.min-w-\\[40rem\\]{min-width:40rem}.min-w-\\[80px\\]{min-width:80px}.min-w-\\[100px\\]{min-width:100px}.min-w-\\[150px\\]{min-width:150px}.min-w-\\[170px\\]{min-width:170px}.min-w-\\[220px\\]{min-width:220px}.min-w-\\[520px\\]{min-width:520px}.min-w-\\[560px\\]{min-width:560px}.min-w-full{min-width:100%}.min-w-max{min-width:max-content}.flex-1{flex:1}.flex-\\[1_1_14rem\\]{flex:14rem}.flex-\\[1_1_16rem\\]{flex:16rem}.flex-none{flex:none}.flex-shrink,.shrink{flex-shrink:1}.shrink-0{flex-shrink:0}.shrink-\\[100\\]{flex-shrink:100}.flex-grow,.grow{flex-grow:1}.grow-0{flex-grow:0}.basis-0{flex-basis:0}.basis-28{flex-basis:calc(var(--spacing) * 28)}.basis-32{flex-basis:calc(var(--spacing) * 32)}.basis-40{flex-basis:calc(var(--spacing) * 40)}.basis-48{flex-basis:calc(var(--spacing) * 48)}.basis-64{flex-basis:calc(var(--spacing) * 64)}.basis-auto{flex-basis:auto}.basis-full{flex-basis:100%}.table-auto{table-layout:auto}.table-fixed{table-layout:fixed}.border-collapse{border-collapse:collapse}.border-separate{border-collapse:separate}.border-spacing-y-density-2{--tw-border-spacing-y:var(--spacing-density-2);border-spacing:var(--tw-border-spacing-x) var(--tw-border-spacing-y)}.origin-top-left{transform-origin:0 0}.-translate-x-1\\/2{--tw-translate-x:calc(calc(1 / 2 * 100%) * -1);translate:var(--tw-translate-x) var(--tw-translate-y)}.-translate-x-2{--tw-translate-x:calc(var(--spacing) * -2);translate:var(--tw-translate-x) var(--tw-translate-y)}.-translate-x-\\[calc\\(100\\%\\+0\\.375rem\\)\\]{--tw-translate-x:calc(calc(100% + .375rem) * -1);translate:var(--tw-translate-x) var(--tw-translate-y)}.-translate-x-full{--tw-translate-x:-100%;translate:var(--tw-translate-x) var(--tw-translate-y)}.translate-x-0{--tw-translate-x:0px;translate:var(--tw-translate-x) var(--tw-translate-y)}.translate-x-1\\.5{--tw-translate-x:calc(var(--spacing) * 1.5);translate:var(--tw-translate-x) var(--tw-translate-y)}.translate-x-4{--tw-translate-x:calc(var(--spacing) * 4);translate:var(--tw-translate-x) var(--tw-translate-y)}.translate-x-8{--tw-translate-x:calc(var(--spacing) * 8);translate:var(--tw-translate-x) var(--tw-translate-y)}.-translate-y-1\\/2{--tw-translate-y:calc(calc(1 / 2 * 100%) * -1);translate:var(--tw-translate-x) var(--tw-translate-y)}.translate-y-1\\/2{--tw-translate-y:calc(1 / 2 * 100%);translate:var(--tw-translate-x) var(--tw-translate-y)}.-scale-x-100{--tw-scale-x:calc(100% * -1);scale:var(--tw-scale-x) var(--tw-scale-y)}.scale-\\[0\\.2\\]{scale:.2}.scale-\\[0\\.66\\]{scale:.66}.-rotate-90{rotate:-90deg}.rotate-0{rotate:0deg}.rotate-45{rotate:45deg}.rotate-90{rotate:90deg}.rotate-180{rotate:180deg}.transform{transform:var(--tw-rotate-x,) var(--tw-rotate-y,) var(--tw-rotate-z,) var(--tw-skew-x,) var(--tw-skew-y,)}.animate-pulse{animation:var(--animate-pulse)}.animate-spin{animation:var(--animate-spin)}.cursor-col-resize{cursor:col-resize}.cursor-crosshair{cursor:crosshair}.cursor-default{cursor:default}.cursor-grab{cursor:grab}.cursor-grabbing{cursor:grabbing}.cursor-help{cursor:help}.cursor-move{cursor:move}.cursor-not-allowed{cursor:not-allowed}.cursor-pointer{cursor:pointer}.cursor-row-resize{cursor:row-resize}.cursor-text{cursor:text}.cursor-zoom-in{cursor:zoom-in}.cursor-zoom-out{cursor:zoom-out}.touch-none{touch-action:none}.resize{resize:both}.resize-none{resize:none}.resize-y{resize:vertical}.scroll-mt-4{scroll-margin-top:calc(var(--spacing) * 4)}.scroll-mt-24{scroll-margin-top:calc(var(--spacing) * 24)}.\\[scrollbar-width\\:none\\]{scrollbar-width:none}.list-decimal{list-style-type:decimal}.list-disc{list-style-type:disc}.list-none{list-style-type:none}.\\[appearance\\:textfield\\]{appearance:textfield}.appearance-none{appearance:none}.grid-cols-1{grid-template-columns:repeat(1,minmax(0,1fr))}.grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}.grid-cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}.grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}.grid-cols-\\[0fr\\]{grid-template-columns:0fr}.grid-cols-\\[1\\.25rem_2\\.5rem_1fr\\]{grid-template-columns:1.25rem 2.5rem 1fr}.grid-cols-\\[1\\.25rem_minmax\\(0\\,1fr\\)\\]{grid-template-columns:1.25rem minmax(0,1fr)}.grid-cols-\\[1\\.25rem_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:1.25rem minmax(0,1fr) auto}.grid-cols-\\[1fr\\]{grid-template-columns:1fr}.grid-cols-\\[1rem_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:1rem minmax(0,1fr) auto}.grid-cols-\\[2\\.5rem_1fr\\]{grid-template-columns:2.5rem 1fr}.grid-cols-\\[2\\.5rem_3\\.5rem_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:2.5rem 3.5rem minmax(0,1fr) auto}.grid-cols-\\[2rem_1fr\\]{grid-template-columns:2rem 1fr}.grid-cols-\\[2rem_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:2rem minmax(0,1fr) auto}.grid-cols-\\[3\\.5rem_1fr\\]{grid-template-columns:3.5rem 1fr}.grid-cols-\\[3\\.5rem_minmax\\(24rem\\,1fr\\)\\]{grid-template-columns:3.5rem minmax(24rem,1fr)}.grid-cols-\\[5\\.5rem_minmax\\(0\\,1fr\\)\\]{grid-template-columns:5.5rem minmax(0,1fr)}.grid-cols-\\[7rem_1fr\\]{grid-template-columns:7rem 1fr}.grid-cols-\\[8rem_1fr\\]{grid-template-columns:8rem 1fr}.grid-cols-\\[10rem_1fr\\]{grid-template-columns:10rem 1fr}.grid-cols-\\[10rem_1fr_9rem\\]{grid-template-columns:10rem 1fr 9rem}.grid-cols-\\[10rem_1fr_auto\\]{grid-template-columns:10rem 1fr auto}.grid-cols-\\[13rem_1\\.25rem_2\\.5rem_1fr\\]{grid-template-columns:13rem 1.25rem 2.5rem 1fr}.grid-cols-\\[18rem_24rem_30rem\\]{grid-template-columns:18rem 24rem 30rem}.grid-cols-\\[auto_1fr\\]{grid-template-columns:auto 1fr}.grid-cols-\\[auto_auto_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:auto auto minmax(0,1fr) auto}.grid-cols-\\[auto_minmax\\(0\\,1fr\\)\\]{grid-template-columns:auto minmax(0,1fr)}.grid-cols-\\[auto_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:auto minmax(0,1fr) auto}.grid-cols-\\[max-content_minmax\\(0\\,1fr\\)\\]{grid-template-columns:max-content minmax(0,1fr)}.grid-cols-\\[minmax\\(0\\,1fr\\)_7rem\\]{grid-template-columns:minmax(0,1fr) 7rem}.grid-cols-\\[minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:minmax(0,1fr) auto}.grid-cols-\\[minmax\\(0\\,1fr\\)_auto_auto\\]{grid-template-columns:minmax(0,1fr) auto auto}.grid-cols-\\[minmax\\(0\\,10rem\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(0,10rem) minmax(0,1fr)}.grid-cols-\\[minmax\\(0\\,14rem\\)_minmax\\(0\\,1fr\\)_4rem\\]{grid-template-columns:minmax(0,14rem) minmax(0,1fr) 4rem}.grid-cols-\\[minmax\\(7rem\\,12rem\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(7rem,12rem) minmax(0,1fr)}.grid-cols-\\[repeat\\(auto-fill\\,minmax\\(5\\.5rem\\,1fr\\)\\)\\]{grid-template-columns:repeat(auto-fill,minmax(5.5rem,1fr))}.grid-cols-\\[repeat\\(auto-fill\\,minmax\\(11rem\\,1fr\\)\\)\\]{grid-template-columns:repeat(auto-fill,minmax(11rem,1fr))}.grid-cols-\\[repeat\\(auto-fill\\,minmax\\(12rem\\,1fr\\)\\)\\]{grid-template-columns:repeat(auto-fill,minmax(12rem,1fr))}.grid-cols-\\[repeat\\(auto-fill\\,minmax\\(13rem\\,1fr\\)\\)\\]{grid-template-columns:repeat(auto-fill,minmax(13rem,1fr))}.grid-cols-\\[repeat\\(auto-fill\\,minmax\\(200px\\,1fr\\)\\)\\]{grid-template-columns:repeat(auto-fill,minmax(200px,1fr))}.grid-cols-\\[repeat\\(auto-fill\\,…\\)\\]{grid-template-columns:repeat(auto-fill,…)}.grid-cols-subgrid{grid-template-columns:subgrid}.grid-rows-\\[0fr\\]{grid-template-rows:0fr}.grid-rows-\\[1fr\\]{grid-template-rows:1fr}.grid-rows-\\[auto_minmax\\(0\\,1fr\\)\\]{grid-template-rows:auto minmax(0,1fr)}.flex-col{flex-direction:column}.flex-nowrap{flex-wrap:nowrap}.flex-wrap{flex-wrap:wrap}.place-items-center{place-items:center}.content-start{align-content:flex-start}.items-baseline{align-items:baseline}.items-center{align-items:center}.items-end{align-items:flex-end}.items-start{align-items:flex-start}.items-stretch{align-items:stretch}.justify-between{justify-content:space-between}.justify-center{justify-content:center}.justify-end{justify-content:flex-end}.justify-start{justify-content:flex-start}.justify-stretch{justify-content:stretch}.gap-0{gap:0}.gap-0\\.5{gap:calc(var(--spacing) * .5)}.gap-1{gap:var(--spacing)}.gap-1\\.5{gap:calc(var(--spacing) * 1.5)}.gap-2{gap:calc(var(--spacing) * 2)}.gap-2\\.5{gap:calc(var(--spacing) * 2.5)}.gap-3{gap:calc(var(--spacing) * 3)}.gap-4{gap:calc(var(--spacing) * 4)}.gap-5{gap:calc(var(--spacing) * 5)}.gap-6{gap:calc(var(--spacing) * 6)}.gap-8{gap:calc(var(--spacing) * 8)}.gap-\\[5px\\]{gap:5px}.gap-density-1{gap:var(--spacing-density-1)}.gap-density-2{gap:var(--spacing-density-2)}.gap-density-3{gap:var(--spacing-density-3)}.gap-density-4{gap:var(--spacing-density-4)}.gap-px{gap:1px}:where(.space-y-0\\.5>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * .5) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * .5) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-1>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(var(--spacing) * var(--tw-space-y-reverse));margin-block-end:calc(var(--spacing) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-1\\.5>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 1.5) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 1.5) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-2>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 2) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 2) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-2\\.5>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 2.5) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 2.5) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-3>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 3) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 3) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-4>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 4) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 4) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-5>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 5) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 5) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-6>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 6) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 6) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-8>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(calc(var(--spacing) * 8) * var(--tw-space-y-reverse));margin-block-end:calc(calc(var(--spacing) * 8) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-density-1>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(var(--spacing-density-1) * var(--tw-space-y-reverse));margin-block-end:calc(var(--spacing-density-1) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-density-2>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(var(--spacing-density-2) * var(--tw-space-y-reverse));margin-block-end:calc(var(--spacing-density-2) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-density-3>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(var(--spacing-density-3) * var(--tw-space-y-reverse));margin-block-end:calc(var(--spacing-density-3) * calc(1 - var(--tw-space-y-reverse)))}:where(.space-y-density-4>:not(:last-child)){--tw-space-y-reverse:0;margin-block-start:calc(var(--spacing-density-4) * var(--tw-space-y-reverse));margin-block-end:calc(var(--spacing-density-4) * calc(1 - var(--tw-space-y-reverse)))}.gap-x-1{column-gap:var(--spacing)}.gap-x-1\\.5{column-gap:calc(var(--spacing) * 1.5)}.gap-x-2{column-gap:calc(var(--spacing) * 2)}.gap-x-3{column-gap:calc(var(--spacing) * 3)}.gap-x-4{column-gap:calc(var(--spacing) * 4)}.gap-x-5{column-gap:calc(var(--spacing) * 5)}.gap-x-6{column-gap:calc(var(--spacing) * 6)}.gap-x-density-2{column-gap:var(--spacing-density-2)}.gap-x-density-3{column-gap:var(--spacing-density-3)}.gap-y-0\\.5{row-gap:calc(var(--spacing) * .5)}.gap-y-1{row-gap:var(--spacing)}.gap-y-1\\.5{row-gap:calc(var(--spacing) * 1.5)}.gap-y-2{row-gap:calc(var(--spacing) * 2)}.gap-y-density-2{row-gap:var(--spacing-density-2)}:where(.divide-x>:not(:last-child)){--tw-divide-x-reverse:0;border-inline-style:var(--tw-border-style);border-inline-start-width:calc(1px * var(--tw-divide-x-reverse));border-inline-end-width:calc(1px * calc(1 - var(--tw-divide-x-reverse)))}:where(.divide-y>:not(:last-child)){--tw-divide-y-reverse:0;border-bottom-style:var(--tw-border-style);border-top-style:var(--tw-border-style);border-top-width:calc(1px * var(--tw-divide-y-reverse));border-bottom-width:calc(1px * calc(1 - var(--tw-divide-y-reverse)))}:where(.divide-y-0>:not(:last-child)){--tw-divide-y-reverse:0;border-bottom-style:var(--tw-border-style);border-top-style:var(--tw-border-style);border-top-width:calc(0px * var(--tw-divide-y-reverse));border-bottom-width:calc(0px * calc(1 - var(--tw-divide-y-reverse)))}:where(.divide-border>:not(:last-child)),:where(.divide-border\\/60>:not(:last-child)){border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){:where(.divide-border\\/60>:not(:last-child)){border-color:color-mix(in oklab, var(--border) 60%, transparent)}}:where(.divide-border\\/70>:not(:last-child)){border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){:where(.divide-border\\/70>:not(:last-child)){border-color:color-mix(in oklab, var(--border) 70%, transparent)}}.self-center{align-self:center}.self-start{align-self:flex-start}.self-stretch{align-self:stretch}.justify-self-start{justify-self:flex-start}.truncate{text-overflow:ellipsis;white-space:nowrap;overflow:hidden}.overflow-auto{overflow:auto}.overflow-hidden{overflow:hidden}.overflow-visible{overflow:visible}.overflow-x-auto{overflow-x:auto}.overflow-y-auto{overflow-y:auto}.overscroll-contain{overscroll-behavior:contain}.overscroll-x-contain{overscroll-behavior-x:contain}.rounded{border-radius:.25rem}.rounded-2xl{border-radius:var(--radius-2xl)}.rounded-\\[1px\\]{border-radius:1px}.rounded-\\[2px\\]{border-radius:2px}.rounded-\\[3px\\]{border-radius:3px}.rounded-\\[4px\\]{border-radius:4px}.rounded-\\[5px\\]{border-radius:5px}.rounded-\\[6px\\]{border-radius:6px}.rounded-\\[13px\\]{border-radius:13px}.rounded-\\[14px\\]{border-radius:14px}.rounded-full{border-radius:2147483647px}.rounded-lg{border-radius:var(--radius-lg)}.rounded-md{border-radius:var(--radius-md)}.rounded-none{border-radius:0}.rounded-sm{border-radius:var(--radius-sm)}.rounded-xl{border-radius:var(--radius-xl)}.rounded-t{border-top-left-radius:.25rem;border-top-right-radius:.25rem}.rounded-t-\\[10px\\]{border-top-left-radius:10px;border-top-right-radius:10px}.rounded-t-md{border-top-left-radius:var(--radius-md);border-top-right-radius:var(--radius-md)}.rounded-t-none{border-top-left-radius:0;border-top-right-radius:0}.rounded-l{border-top-left-radius:.25rem;border-bottom-left-radius:.25rem}.rounded-l-\\[10px\\]{border-top-left-radius:10px;border-bottom-left-radius:10px}.rounded-l-none{border-top-left-radius:0;border-bottom-left-radius:0}.rounded-tl{border-top-left-radius:.25rem}.rounded-r{border-top-right-radius:.25rem;border-bottom-right-radius:.25rem}.rounded-r-\\[10px\\]{border-top-right-radius:10px;border-bottom-right-radius:10px}.rounded-r-none{border-top-right-radius:0;border-bottom-right-radius:0}.rounded-tr{border-top-right-radius:.25rem}.rounded-tr-md{border-top-right-radius:var(--radius-md)}.rounded-b{border-bottom-right-radius:.25rem;border-bottom-left-radius:.25rem}.rounded-b-\\[10px\\]{border-bottom-right-radius:10px;border-bottom-left-radius:10px}.rounded-b-none{border-bottom-right-radius:0;border-bottom-left-radius:0}.rounded-br{border-bottom-right-radius:.25rem}.rounded-bl{border-bottom-left-radius:.25rem}.border{border-style:var(--tw-border-style);border-width:1px}.border-0{border-style:var(--tw-border-style);border-width:0}.border-2{border-style:var(--tw-border-style);border-width:2px}.\\[border-width\\:0\\.5px\\]{border-width:.5px}.border-x{border-inline-style:var(--tw-border-style);border-inline-width:1px}.border-y{border-block-style:var(--tw-border-style);border-block-width:1px}.border-s{border-inline-start-style:var(--tw-border-style);border-inline-start-width:1px}.border-e{border-inline-end-style:var(--tw-border-style);border-inline-end-width:1px}.border-t{border-top-style:var(--tw-border-style);border-top-width:1px}.border-t-0{border-top-style:var(--tw-border-style);border-top-width:0}.border-t-2{border-top-style:var(--tw-border-style);border-top-width:2px}.border-t-4{border-top-style:var(--tw-border-style);border-top-width:4px}.\\[border-top-width\\:0\\.5px\\]{border-top-width:.5px}.border-r{border-right-style:var(--tw-border-style);border-right-width:1px}.border-r-0{border-right-style:var(--tw-border-style);border-right-width:0}.border-r-2{border-right-style:var(--tw-border-style);border-right-width:2px}.border-b{border-bottom-style:var(--tw-border-style);border-bottom-width:1px}.border-b-0{border-bottom-style:var(--tw-border-style);border-bottom-width:0}.border-b-2{border-bottom-style:var(--tw-border-style);border-bottom-width:2px}.border-l{border-left-style:var(--tw-border-style);border-left-width:1px}.border-l-0{border-left-style:var(--tw-border-style);border-left-width:0}.border-l-2{border-left-style:var(--tw-border-style);border-left-width:2px}.border-l-4{border-left-style:var(--tw-border-style);border-left-width:4px}.border-l-\\[3px\\]{border-left-style:var(--tw-border-style);border-left-width:3px}.border-dashed{--tw-border-style:dashed;border-style:dashed}.border-dotted{--tw-border-style:dotted;border-style:dotted}.border-none{--tw-border-style:none;border-style:none}.border-solid{--tw-border-style:solid;border-style:solid}.border-\\[var\\(--rg-border-button\\)\\]{border-color:var(--rg-border-button)}.border-amber-200{border-color:var(--color-amber-200)}.border-amber-300{border-color:var(--color-amber-300)}.border-amber-300\\/60{border-color:#ffd23699}@supports (color:color-mix(in lab, red, red)){.border-amber-300\\/60{border-color:color-mix(in oklab, var(--color-amber-300) 60%, transparent)}}.border-amber-400{border-color:var(--color-amber-400)}.border-amber-500{border-color:var(--color-amber-500)}.border-amber-500\\/30{border-color:#f99c004d}@supports (color:color-mix(in lab, red, red)){.border-amber-500\\/30{border-color:color-mix(in oklab, var(--color-amber-500) 30%, transparent)}}.border-amber-500\\/40{border-color:#f99c0066}@supports (color:color-mix(in lab, red, red)){.border-amber-500\\/40{border-color:color-mix(in oklab, var(--color-amber-500) 40%, transparent)}}.border-amber-500\\/50{border-color:#f99c0080}@supports (color:color-mix(in lab, red, red)){.border-amber-500\\/50{border-color:color-mix(in oklab, var(--color-amber-500) 50%, transparent)}}.border-amber-600\\/40{border-color:#dd740066}@supports (color:color-mix(in lab, red, red)){.border-amber-600\\/40{border-color:color-mix(in oklab, var(--color-amber-600) 40%, transparent)}}.border-blue-300{border-color:var(--color-blue-300)}.border-blue-500{border-color:var(--color-blue-500)}.border-blue-500\\/30{border-color:#3080ff4d}@supports (color:color-mix(in lab, red, red)){.border-blue-500\\/30{border-color:color-mix(in oklab, var(--color-blue-500) 30%, transparent)}}.border-border,.border-border\\/50{border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.border-border\\/50{border-color:color-mix(in oklab, var(--border) 50%, transparent)}}.border-border\\/60{border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.border-border\\/60{border-color:color-mix(in oklab, var(--border) 60%, transparent)}}.border-border\\/70{border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.border-border\\/70{border-color:color-mix(in oklab, var(--border) 70%, transparent)}}.border-border\\/80{border-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.border-border\\/80{border-color:color-mix(in oklab, var(--border) 80%, transparent)}}.border-current,.border-current\\/10{border-color:currentColor}@supports (color:color-mix(in lab, red, red)){.border-current\\/10{border-color:color-mix(in oklab, currentcolor 10%, transparent)}}.border-current\\/20{border-color:currentColor}@supports (color:color-mix(in lab, red, red)){.border-current\\/20{border-color:color-mix(in oklab, currentcolor 20%, transparent)}}.border-current\\/25{border-color:currentColor}@supports (color:color-mix(in lab, red, red)){.border-current\\/25{border-color:color-mix(in oklab, currentcolor 25%, transparent)}}.border-destructive,.border-destructive\\/20{border-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.border-destructive\\/20{border-color:color-mix(in oklab, var(--destructive) 20%, transparent)}}.border-destructive\\/30{border-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.border-destructive\\/30{border-color:color-mix(in oklab, var(--destructive) 30%, transparent)}}.border-destructive\\/40{border-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.border-destructive\\/40{border-color:color-mix(in oklab, var(--destructive) 40%, transparent)}}.border-destructive\\/50{border-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.border-destructive\\/50{border-color:color-mix(in oklab, var(--destructive) 50%, transparent)}}.border-emerald-200{border-color:var(--color-emerald-200)}.border-emerald-300{border-color:var(--color-emerald-300)}.border-emerald-500{border-color:var(--color-emerald-500)}.border-emerald-500\\/30{border-color:#00bb7f4d}@supports (color:color-mix(in lab, red, red)){.border-emerald-500\\/30{border-color:color-mix(in oklab, var(--color-emerald-500) 30%, transparent)}}.border-emerald-500\\/50{border-color:#00bb7f80}@supports (color:color-mix(in lab, red, red)){.border-emerald-500\\/50{border-color:color-mix(in oklab, var(--color-emerald-500) 50%, transparent)}}.border-emerald-600{border-color:var(--color-emerald-600)}.border-emerald-600\\/40{border-color:#00976766}@supports (color:color-mix(in lab, red, red)){.border-emerald-600\\/40{border-color:color-mix(in oklab, var(--color-emerald-600) 40%, transparent)}}.border-foreground\\/20{border-color:var(--foreground)}@supports (color:color-mix(in lab, red, red)){.border-foreground\\/20{border-color:color-mix(in oklab, var(--foreground) 20%, transparent)}}.border-fuchsia-300{border-color:var(--color-fuchsia-300)}.border-fuchsia-500{border-color:var(--color-fuchsia-500)}.border-green-500{border-color:var(--color-green-500)}.border-green-500\\/50{border-color:#00c75880}@supports (color:color-mix(in lab, red, red)){.border-green-500\\/50{border-color:color-mix(in oklab, var(--color-green-500) 50%, transparent)}}.border-green-600\\/40{border-color:#00a54466}@supports (color:color-mix(in lab, red, red)){.border-green-600\\/40{border-color:color-mix(in oklab, var(--color-green-600) 40%, transparent)}}.border-indigo-300{border-color:var(--color-indigo-300)}.border-indigo-500{border-color:var(--color-indigo-500)}.border-input{border-color:var(--input)}.border-muted-foreground,.border-muted-foreground\\/50{border-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.border-muted-foreground\\/50{border-color:color-mix(in oklab, var(--muted-foreground) 50%, transparent)}}.border-muted-foreground\\/60{border-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.border-muted-foreground\\/60{border-color:color-mix(in oklab, var(--muted-foreground) 60%, transparent)}}.border-orange-300{border-color:var(--color-orange-300)}.border-orange-500{border-color:var(--color-orange-500)}.border-pink-500{border-color:var(--color-pink-500)}.border-primary,.border-primary\\/30{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.border-primary\\/30{border-color:color-mix(in oklab, var(--primary) 30%, transparent)}}.border-primary\\/40{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.border-primary\\/40{border-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.border-primary\\/50{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.border-primary\\/50{border-color:color-mix(in oklab, var(--primary) 50%, transparent)}}.border-primary\\/60{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.border-primary\\/60{border-color:color-mix(in oklab, var(--primary) 60%, transparent)}}.border-primary\\/70{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.border-primary\\/70{border-color:color-mix(in oklab, var(--primary) 70%, transparent)}}.border-purple-500{border-color:var(--color-purple-500)}.border-purple-500\\/30{border-color:#ac4bff4d}@supports (color:color-mix(in lab, red, red)){.border-purple-500\\/30{border-color:color-mix(in oklab, var(--color-purple-500) 30%, transparent)}}.border-red-300{border-color:var(--color-red-300)}.border-red-400{border-color:var(--color-red-400)}.border-red-500{border-color:var(--color-red-500)}.border-red-500\\/30{border-color:#fb2c364d}@supports (color:color-mix(in lab, red, red)){.border-red-500\\/30{border-color:color-mix(in oklab, var(--color-red-500) 30%, transparent)}}.border-red-500\\/40{border-color:#fb2c3666}@supports (color:color-mix(in lab, red, red)){.border-red-500\\/40{border-color:color-mix(in oklab, var(--color-red-500) 40%, transparent)}}.border-red-500\\/50{border-color:#fb2c3680}@supports (color:color-mix(in lab, red, red)){.border-red-500\\/50{border-color:color-mix(in oklab, var(--color-red-500) 50%, transparent)}}.border-red-500\\/60{border-color:#fb2c3699}@supports (color:color-mix(in lab, red, red)){.border-red-500\\/60{border-color:color-mix(in oklab, var(--color-red-500) 60%, transparent)}}.border-red-600{border-color:var(--color-red-600)}.border-rose-200{border-color:var(--color-rose-200)}.border-rose-300{border-color:var(--color-rose-300)}.border-rose-500{border-color:var(--color-rose-500)}.border-rose-500\\/30{border-color:#ff23574d}@supports (color:color-mix(in lab, red, red)){.border-rose-500\\/30{border-color:color-mix(in oklab, var(--color-rose-500) 30%, transparent)}}.border-sidebar-border{border-color:var(--sidebar-border)}.border-sky-200{border-color:var(--color-sky-200)}.border-sky-300{border-color:var(--color-sky-300)}.border-sky-500{border-color:var(--color-sky-500)}.border-sky-500\\/20{border-color:#00a5ef33}@supports (color:color-mix(in lab, red, red)){.border-sky-500\\/20{border-color:color-mix(in oklab, var(--color-sky-500) 20%, transparent)}}.border-sky-500\\/25{border-color:#00a5ef40}@supports (color:color-mix(in lab, red, red)){.border-sky-500\\/25{border-color:color-mix(in oklab, var(--color-sky-500) 25%, transparent)}}.border-sky-500\\/30{border-color:#00a5ef4d}@supports (color:color-mix(in lab, red, red)){.border-sky-500\\/30{border-color:color-mix(in oklab, var(--color-sky-500) 30%, transparent)}}.border-sky-500\\/50{border-color:#00a5ef80}@supports (color:color-mix(in lab, red, red)){.border-sky-500\\/50{border-color:color-mix(in oklab, var(--color-sky-500) 50%, transparent)}}.border-sky-500\\/60{border-color:#00a5ef99}@supports (color:color-mix(in lab, red, red)){.border-sky-500\\/60{border-color:color-mix(in oklab, var(--color-sky-500) 60%, transparent)}}.border-sky-600{border-color:var(--color-sky-600)}.border-slate-600{border-color:var(--color-slate-600)}.border-teal-300{border-color:var(--color-teal-300)}.border-teal-500{border-color:var(--color-teal-500)}.border-transparent{border-color:#0000}.border-violet-300{border-color:var(--color-violet-300)}.border-violet-500{border-color:var(--color-violet-500)}.border-violet-500\\/30{border-color:#8d54ff4d}@supports (color:color-mix(in lab, red, red)){.border-violet-500\\/30{border-color:color-mix(in oklab, var(--color-violet-500) 30%, transparent)}}.border-white\\/10{border-color:#ffffff1a}@supports (color:color-mix(in lab, red, red)){.border-white\\/10{border-color:color-mix(in oklab, var(--color-white) 10%, transparent)}}.border-yellow-300{border-color:var(--color-yellow-300)}.border-yellow-400{border-color:var(--color-yellow-400)}.border-yellow-500{border-color:var(--color-yellow-500)}.border-yellow-500\\/30{border-color:#edb2004d}@supports (color:color-mix(in lab, red, red)){.border-yellow-500\\/30{border-color:color-mix(in oklab, var(--color-yellow-500) 30%, transparent)}}.border-zinc-800{border-color:var(--color-zinc-800)}.border-t-\\[var\\(--rg-border-subtle\\)\\]{border-top-color:var(--rg-border-subtle)}.border-t-primary{border-top-color:var(--primary)}.border-l-amber-400{border-left-color:var(--color-amber-400)}.border-l-amber-500{border-left-color:var(--color-amber-500)}.border-l-amber-500\\/60{border-left-color:#f99c0099}@supports (color:color-mix(in lab, red, red)){.border-l-amber-500\\/60{border-left-color:color-mix(in oklab, var(--color-amber-500) 60%, transparent)}}.border-l-blue-500{border-left-color:var(--color-blue-500)}.border-l-border{border-left-color:var(--border)}.border-l-current\\/25{border-left-color:currentColor}@supports (color:color-mix(in lab, red, red)){.border-l-current\\/25{border-left-color:color-mix(in oklab, currentcolor 25%, transparent)}}.border-l-emerald-400{border-left-color:var(--color-emerald-400)}.border-l-emerald-500{border-left-color:var(--color-emerald-500)}.border-l-green-500{border-left-color:var(--color-green-500)}.border-l-indigo-400{border-left-color:var(--color-indigo-400)}.border-l-purple-500{border-left-color:var(--color-purple-500)}.border-l-red-500{border-left-color:var(--color-red-500)}.border-l-rose-400{border-left-color:var(--color-rose-400)}.border-l-sky-400{border-left-color:var(--color-sky-400)}.border-l-slate-400{border-left-color:var(--color-slate-400)}.border-l-teal-400{border-left-color:var(--color-teal-400)}.border-l-transparent{border-left-color:#0000}.border-l-violet-400{border-left-color:var(--color-violet-400)}.border-l-yellow-500{border-left-color:var(--color-yellow-500)}.bg-\\[\\#000000\\]{background-color:#000}.bg-\\[\\#f00\\]{background-color:red}.bg-\\[rgb\\(0\\,0\\,0\\)\\]{background-color:#000}.bg-\\[var\\(--clicky-success-500\\)\\]{background-color:var(--clicky-success-500)}.bg-\\[var\\(--rg-error-bg\\)\\]{background-color:var(--rg-error-bg)}.bg-\\[var\\(--rg-panel-bg\\)\\]{background-color:var(--rg-panel-bg)}.bg-\\[var\\(--rg-submit-bg\\)\\]{background-color:var(--rg-submit-bg)}.bg-\\[var\\(--rg-surface-active\\)\\]{background-color:var(--rg-surface-active)}.bg-\\[var\\(--rg-surface-hover\\)\\]{background-color:var(--rg-surface-hover)}.bg-\\[var\\(--rg-text-primary\\)\\]{background-color:var(--rg-text-primary)}.bg-\\[var\\(--rg-text-secondary\\)\\]{background-color:var(--rg-text-secondary)}.bg-accent,.bg-accent\\/30{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.bg-accent\\/30{background-color:color-mix(in oklab, var(--accent) 30%, transparent)}}.bg-accent\\/40{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.bg-accent\\/40{background-color:color-mix(in oklab, var(--accent) 40%, transparent)}}.bg-accent\\/50{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.bg-accent\\/50{background-color:color-mix(in oklab, var(--accent) 50%, transparent)}}.bg-accent\\/60{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.bg-accent\\/60{background-color:color-mix(in oklab, var(--accent) 60%, transparent)}}.bg-accent\\/70{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.bg-accent\\/70{background-color:color-mix(in oklab, var(--accent) 70%, transparent)}}.bg-amber-50{background-color:var(--color-amber-50)}.bg-amber-50\\/40{background-color:#fffbeb66}@supports (color:color-mix(in lab, red, red)){.bg-amber-50\\/40{background-color:color-mix(in oklab, var(--color-amber-50) 40%, transparent)}}.bg-amber-50\\/85{background-color:#fffbebd9}@supports (color:color-mix(in lab, red, red)){.bg-amber-50\\/85{background-color:color-mix(in oklab, var(--color-amber-50) 85%, transparent)}}.bg-amber-100{background-color:var(--color-amber-100)}.bg-amber-100\\/30{background-color:#fef3c64d}@supports (color:color-mix(in lab, red, red)){.bg-amber-100\\/30{background-color:color-mix(in oklab, var(--color-amber-100) 30%, transparent)}}.bg-amber-100\\/70{background-color:#fef3c6b3}@supports (color:color-mix(in lab, red, red)){.bg-amber-100\\/70{background-color:color-mix(in oklab, var(--color-amber-100) 70%, transparent)}}.bg-amber-200{background-color:var(--color-amber-200)}.bg-amber-200\\/50{background-color:#fee68580}@supports (color:color-mix(in lab, red, red)){.bg-amber-200\\/50{background-color:color-mix(in oklab, var(--color-amber-200) 50%, transparent)}}.bg-amber-200\\/60{background-color:#fee68599}@supports (color:color-mix(in lab, red, red)){.bg-amber-200\\/60{background-color:color-mix(in oklab, var(--color-amber-200) 60%, transparent)}}.bg-amber-300{background-color:var(--color-amber-300)}.bg-amber-400{background-color:var(--color-amber-400)}.bg-amber-400\\/10{background-color:#fcbb001a}@supports (color:color-mix(in lab, red, red)){.bg-amber-400\\/10{background-color:color-mix(in oklab, var(--color-amber-400) 10%, transparent)}}.bg-amber-400\\/15{background-color:#fcbb0026}@supports (color:color-mix(in lab, red, red)){.bg-amber-400\\/15{background-color:color-mix(in oklab, var(--color-amber-400) 15%, transparent)}}.bg-amber-500{background-color:var(--color-amber-500)}.bg-amber-500\\/5{background-color:#f99c000d}@supports (color:color-mix(in lab, red, red)){.bg-amber-500\\/5{background-color:color-mix(in oklab, var(--color-amber-500) 5%, transparent)}}.bg-amber-500\\/10{background-color:#f99c001a}@supports (color:color-mix(in lab, red, red)){.bg-amber-500\\/10{background-color:color-mix(in oklab, var(--color-amber-500) 10%, transparent)}}.bg-amber-500\\/15{background-color:#f99c0026}@supports (color:color-mix(in lab, red, red)){.bg-amber-500\\/15{background-color:color-mix(in oklab, var(--color-amber-500) 15%, transparent)}}.bg-amber-500\\/\\[0\\.06\\]{background-color:#f99c000f}@supports (color:color-mix(in lab, red, red)){.bg-amber-500\\/\\[0\\.06\\]{background-color:color-mix(in oklab, var(--color-amber-500) 6%, transparent)}}.bg-amber-600{background-color:var(--color-amber-600)}.bg-amber-700{background-color:var(--color-amber-700)}.bg-amber-800{background-color:var(--color-amber-800)}.bg-amber-900{background-color:var(--color-amber-900)}.bg-amber-950{background-color:var(--color-amber-950)}.bg-background,.bg-background\\/80{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.bg-background\\/80{background-color:color-mix(in oklab, var(--background) 80%, transparent)}}.bg-background\\/90{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.bg-background\\/90{background-color:color-mix(in oklab, var(--background) 90%, transparent)}}.bg-background\\/95{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.bg-background\\/95{background-color:color-mix(in oklab, var(--background) 95%, transparent)}}.bg-black{background-color:var(--color-black)}.bg-black\\/20{background-color:#0003}@supports (color:color-mix(in lab, red, red)){.bg-black\\/20{background-color:color-mix(in oklab, var(--color-black) 20%, transparent)}}.bg-black\\/40{background-color:#0006}@supports (color:color-mix(in lab, red, red)){.bg-black\\/40{background-color:color-mix(in oklab, var(--color-black) 40%, transparent)}}.bg-black\\/60{background-color:#0009}@supports (color:color-mix(in lab, red, red)){.bg-black\\/60{background-color:color-mix(in oklab, var(--color-black) 60%, transparent)}}.bg-blue-50{background-color:var(--color-blue-50)}.bg-blue-100{background-color:var(--color-blue-100)}.bg-blue-200{background-color:var(--color-blue-200)}.bg-blue-300{background-color:var(--color-blue-300)}.bg-blue-400{background-color:var(--color-blue-400)}.bg-blue-500{background-color:var(--color-blue-500)}.bg-blue-500\\/5{background-color:#3080ff0d}@supports (color:color-mix(in lab, red, red)){.bg-blue-500\\/5{background-color:color-mix(in oklab, var(--color-blue-500) 5%, transparent)}}.bg-blue-500\\/10{background-color:#3080ff1a}@supports (color:color-mix(in lab, red, red)){.bg-blue-500\\/10{background-color:color-mix(in oklab, var(--color-blue-500) 10%, transparent)}}.bg-blue-600{background-color:var(--color-blue-600)}.bg-blue-700{background-color:var(--color-blue-700)}.bg-blue-800{background-color:var(--color-blue-800)}.bg-blue-900{background-color:var(--color-blue-900)}.bg-blue-950{background-color:var(--color-blue-950)}.bg-border{background-color:var(--border)}.bg-card{background-color:var(--card)}.bg-current{background-color:currentColor}.bg-cyan-50{background-color:var(--color-cyan-50)}.bg-cyan-100{background-color:var(--color-cyan-100)}.bg-cyan-200{background-color:var(--color-cyan-200)}.bg-cyan-300{background-color:var(--color-cyan-300)}.bg-cyan-400{background-color:var(--color-cyan-400)}.bg-cyan-500{background-color:var(--color-cyan-500)}.bg-cyan-600{background-color:var(--color-cyan-600)}.bg-cyan-700{background-color:var(--color-cyan-700)}.bg-cyan-800{background-color:var(--color-cyan-800)}.bg-cyan-900{background-color:var(--color-cyan-900)}.bg-cyan-950{background-color:var(--color-cyan-950)}.bg-destructive,.bg-destructive\\/5{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.bg-destructive\\/5{background-color:color-mix(in oklab, var(--destructive) 5%, transparent)}}.bg-destructive\\/10{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.bg-destructive\\/10{background-color:color-mix(in oklab, var(--destructive) 10%, transparent)}}.bg-destructive\\/20{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.bg-destructive\\/20{background-color:color-mix(in oklab, var(--destructive) 20%, transparent)}}.bg-destructive\\/\\[0\\.06\\]{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.bg-destructive\\/\\[0\\.06\\]{background-color:color-mix(in oklab, var(--destructive) 6%, transparent)}}.bg-emerald-50{background-color:var(--color-emerald-50)}.bg-emerald-50\\/80{background-color:#ecfdf5cc}@supports (color:color-mix(in lab, red, red)){.bg-emerald-50\\/80{background-color:color-mix(in oklab, var(--color-emerald-50) 80%, transparent)}}.bg-emerald-100{background-color:var(--color-emerald-100)}.bg-emerald-200{background-color:var(--color-emerald-200)}.bg-emerald-300{background-color:var(--color-emerald-300)}.bg-emerald-400{background-color:var(--color-emerald-400)}.bg-emerald-500{background-color:var(--color-emerald-500)}.bg-emerald-500\\/5{background-color:#00bb7f0d}@supports (color:color-mix(in lab, red, red)){.bg-emerald-500\\/5{background-color:color-mix(in oklab, var(--color-emerald-500) 5%, transparent)}}.bg-emerald-500\\/10{background-color:#00bb7f1a}@supports (color:color-mix(in lab, red, red)){.bg-emerald-500\\/10{background-color:color-mix(in oklab, var(--color-emerald-500) 10%, transparent)}}.bg-emerald-500\\/12{background-color:#00bb7f1f}@supports (color:color-mix(in lab, red, red)){.bg-emerald-500\\/12{background-color:color-mix(in oklab, var(--color-emerald-500) 12%, transparent)}}.bg-emerald-500\\/15{background-color:#00bb7f26}@supports (color:color-mix(in lab, red, red)){.bg-emerald-500\\/15{background-color:color-mix(in oklab, var(--color-emerald-500) 15%, transparent)}}.bg-emerald-600{background-color:var(--color-emerald-600)}.bg-emerald-700{background-color:var(--color-emerald-700)}.bg-emerald-800{background-color:var(--color-emerald-800)}.bg-emerald-900{background-color:var(--color-emerald-900)}.bg-emerald-950{background-color:var(--color-emerald-950)}.bg-foreground{background-color:var(--foreground)}.bg-fuchsia-50{background-color:var(--color-fuchsia-50)}.bg-fuchsia-100{background-color:var(--color-fuchsia-100)}.bg-fuchsia-200{background-color:var(--color-fuchsia-200)}.bg-fuchsia-300{background-color:var(--color-fuchsia-300)}.bg-fuchsia-400{background-color:var(--color-fuchsia-400)}.bg-fuchsia-500{background-color:var(--color-fuchsia-500)}.bg-fuchsia-600{background-color:var(--color-fuchsia-600)}.bg-fuchsia-700{background-color:var(--color-fuchsia-700)}.bg-fuchsia-800{background-color:var(--color-fuchsia-800)}.bg-fuchsia-900{background-color:var(--color-fuchsia-900)}.bg-fuchsia-950{background-color:var(--color-fuchsia-950)}.bg-gray-50{background-color:var(--color-gray-50)}.bg-gray-100{background-color:var(--color-gray-100)}.bg-gray-200{background-color:var(--color-gray-200)}.bg-gray-300{background-color:var(--color-gray-300)}.bg-gray-400{background-color:var(--color-gray-400)}.bg-gray-500{background-color:var(--color-gray-500)}.bg-gray-600{background-color:var(--color-gray-600)}.bg-gray-700{background-color:var(--color-gray-700)}.bg-gray-800{background-color:var(--color-gray-800)}.bg-gray-900{background-color:var(--color-gray-900)}.bg-gray-950{background-color:var(--color-gray-950)}.bg-green-50{background-color:var(--color-green-50)}.bg-green-100{background-color:var(--color-green-100)}.bg-green-200{background-color:var(--color-green-200)}.bg-green-300{background-color:var(--color-green-300)}.bg-green-400{background-color:var(--color-green-400)}.bg-green-500{background-color:var(--color-green-500)}.bg-green-500\\/5{background-color:#00c7580d}@supports (color:color-mix(in lab, red, red)){.bg-green-500\\/5{background-color:color-mix(in oklab, var(--color-green-500) 5%, transparent)}}.bg-green-500\\/10{background-color:#00c7581a}@supports (color:color-mix(in lab, red, red)){.bg-green-500\\/10{background-color:color-mix(in oklab, var(--color-green-500) 10%, transparent)}}.bg-green-500\\/20{background-color:#00c75833}@supports (color:color-mix(in lab, red, red)){.bg-green-500\\/20{background-color:color-mix(in oklab, var(--color-green-500) 20%, transparent)}}.bg-green-500\\/80{background-color:#00c758cc}@supports (color:color-mix(in lab, red, red)){.bg-green-500\\/80{background-color:color-mix(in oklab, var(--color-green-500) 80%, transparent)}}.bg-green-600{background-color:var(--color-green-600)}.bg-green-700{background-color:var(--color-green-700)}.bg-green-800{background-color:var(--color-green-800)}.bg-green-900{background-color:var(--color-green-900)}.bg-green-950{background-color:var(--color-green-950)}.bg-indigo-50{background-color:var(--color-indigo-50)}.bg-indigo-100{background-color:var(--color-indigo-100)}.bg-indigo-200{background-color:var(--color-indigo-200)}.bg-indigo-300{background-color:var(--color-indigo-300)}.bg-indigo-400{background-color:var(--color-indigo-400)}.bg-indigo-500{background-color:var(--color-indigo-500)}.bg-indigo-600{background-color:var(--color-indigo-600)}.bg-indigo-700{background-color:var(--color-indigo-700)}.bg-indigo-800{background-color:var(--color-indigo-800)}.bg-indigo-900{background-color:var(--color-indigo-900)}.bg-indigo-950{background-color:var(--color-indigo-950)}.bg-input{background-color:var(--input)}.bg-lime-50{background-color:var(--color-lime-50)}.bg-lime-100{background-color:var(--color-lime-100)}.bg-lime-200{background-color:var(--color-lime-200)}.bg-lime-300{background-color:var(--color-lime-300)}.bg-lime-400{background-color:var(--color-lime-400)}.bg-lime-500{background-color:var(--color-lime-500)}.bg-lime-600{background-color:var(--color-lime-600)}.bg-lime-700{background-color:var(--color-lime-700)}.bg-lime-800{background-color:var(--color-lime-800)}.bg-lime-900{background-color:var(--color-lime-900)}.bg-lime-950{background-color:var(--color-lime-950)}.bg-muted{background-color:var(--muted)}.bg-muted-foreground,.bg-muted-foreground\\/25{background-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.bg-muted-foreground\\/25{background-color:color-mix(in oklab, var(--muted-foreground) 25%, transparent)}}.bg-muted-foreground\\/30{background-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.bg-muted-foreground\\/30{background-color:color-mix(in oklab, var(--muted-foreground) 30%, transparent)}}.bg-muted-foreground\\/40{background-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.bg-muted-foreground\\/40{background-color:color-mix(in oklab, var(--muted-foreground) 40%, transparent)}}.bg-muted-foreground\\/50{background-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.bg-muted-foreground\\/50{background-color:color-mix(in oklab, var(--muted-foreground) 50%, transparent)}}.bg-muted-foreground\\/60{background-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.bg-muted-foreground\\/60{background-color:color-mix(in oklab, var(--muted-foreground) 60%, transparent)}}.bg-muted\\/10{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/10{background-color:color-mix(in oklab, var(--muted) 10%, transparent)}}.bg-muted\\/20{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/20{background-color:color-mix(in oklab, var(--muted) 20%, transparent)}}.bg-muted\\/25{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/25{background-color:color-mix(in oklab, var(--muted) 25%, transparent)}}.bg-muted\\/30{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/30{background-color:color-mix(in oklab, var(--muted) 30%, transparent)}}.bg-muted\\/35{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/35{background-color:color-mix(in oklab, var(--muted) 35%, transparent)}}.bg-muted\\/40{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/40{background-color:color-mix(in oklab, var(--muted) 40%, transparent)}}.bg-muted\\/50{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/50{background-color:color-mix(in oklab, var(--muted) 50%, transparent)}}.bg-muted\\/60{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/60{background-color:color-mix(in oklab, var(--muted) 60%, transparent)}}.bg-muted\\/70{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/70{background-color:color-mix(in oklab, var(--muted) 70%, transparent)}}.bg-muted\\/80{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/80{background-color:color-mix(in oklab, var(--muted) 80%, transparent)}}.bg-muted\\/90{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.bg-muted\\/90{background-color:color-mix(in oklab, var(--muted) 90%, transparent)}}.bg-neutral-50{background-color:var(--color-neutral-50)}.bg-neutral-100{background-color:var(--color-neutral-100)}.bg-neutral-200{background-color:var(--color-neutral-200)}.bg-neutral-300{background-color:var(--color-neutral-300)}.bg-neutral-400{background-color:var(--color-neutral-400)}.bg-neutral-500{background-color:var(--color-neutral-500)}.bg-neutral-600{background-color:var(--color-neutral-600)}.bg-neutral-700{background-color:var(--color-neutral-700)}.bg-neutral-800{background-color:var(--color-neutral-800)}.bg-neutral-900{background-color:var(--color-neutral-900)}.bg-neutral-950{background-color:var(--color-neutral-950)}.bg-orange-50{background-color:var(--color-orange-50)}.bg-orange-100{background-color:var(--color-orange-100)}.bg-orange-200{background-color:var(--color-orange-200)}.bg-orange-300{background-color:var(--color-orange-300)}.bg-orange-400{background-color:var(--color-orange-400)}.bg-orange-500{background-color:var(--color-orange-500)}.bg-orange-500\\/5{background-color:#fe6e000d}@supports (color:color-mix(in lab, red, red)){.bg-orange-500\\/5{background-color:color-mix(in oklab, var(--color-orange-500) 5%, transparent)}}.bg-orange-600{background-color:var(--color-orange-600)}.bg-orange-700{background-color:var(--color-orange-700)}.bg-orange-800{background-color:var(--color-orange-800)}.bg-orange-900{background-color:var(--color-orange-900)}.bg-orange-950{background-color:var(--color-orange-950)}.bg-pink-50{background-color:var(--color-pink-50)}.bg-pink-100{background-color:var(--color-pink-100)}.bg-pink-200{background-color:var(--color-pink-200)}.bg-pink-300{background-color:var(--color-pink-300)}.bg-pink-400{background-color:var(--color-pink-400)}.bg-pink-500{background-color:var(--color-pink-500)}.bg-pink-600{background-color:var(--color-pink-600)}.bg-pink-700{background-color:var(--color-pink-700)}.bg-pink-800{background-color:var(--color-pink-800)}.bg-pink-900{background-color:var(--color-pink-900)}.bg-pink-950{background-color:var(--color-pink-950)}.bg-popover{background-color:var(--popover)}.bg-primary,.bg-primary\\/5{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/5{background-color:color-mix(in oklab, var(--primary) 5%, transparent)}}.bg-primary\\/8{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/8{background-color:color-mix(in oklab, var(--primary) 8%, transparent)}}.bg-primary\\/10{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/10{background-color:color-mix(in oklab, var(--primary) 10%, transparent)}}.bg-primary\\/15{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/15{background-color:color-mix(in oklab, var(--primary) 15%, transparent)}}.bg-primary\\/20{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/20{background-color:color-mix(in oklab, var(--primary) 20%, transparent)}}.bg-primary\\/60{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/60{background-color:color-mix(in oklab, var(--primary) 60%, transparent)}}.bg-primary\\/70{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/70{background-color:color-mix(in oklab, var(--primary) 70%, transparent)}}.bg-primary\\/\\[0\\.04\\]{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/\\[0\\.04\\]{background-color:color-mix(in oklab, var(--primary) 4%, transparent)}}.bg-primary\\/\\[0\\.05\\]{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/\\[0\\.05\\]{background-color:color-mix(in oklab, var(--primary) 5%, transparent)}}.bg-primary\\/\\[0\\.06\\]{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/\\[0\\.06\\]{background-color:color-mix(in oklab, var(--primary) 6%, transparent)}}.bg-primary\\/\\[0\\.08\\]{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.bg-primary\\/\\[0\\.08\\]{background-color:color-mix(in oklab, var(--primary) 8%, transparent)}}.bg-purple-50{background-color:var(--color-purple-50)}.bg-purple-100{background-color:var(--color-purple-100)}.bg-purple-200{background-color:var(--color-purple-200)}.bg-purple-300{background-color:var(--color-purple-300)}.bg-purple-400{background-color:var(--color-purple-400)}.bg-purple-500{background-color:var(--color-purple-500)}.bg-purple-500\\/5{background-color:#ac4bff0d}@supports (color:color-mix(in lab, red, red)){.bg-purple-500\\/5{background-color:color-mix(in oklab, var(--color-purple-500) 5%, transparent)}}.bg-purple-500\\/10{background-color:#ac4bff1a}@supports (color:color-mix(in lab, red, red)){.bg-purple-500\\/10{background-color:color-mix(in oklab, var(--color-purple-500) 10%, transparent)}}.bg-purple-600{background-color:var(--color-purple-600)}.bg-purple-700{background-color:var(--color-purple-700)}.bg-purple-800{background-color:var(--color-purple-800)}.bg-purple-900{background-color:var(--color-purple-900)}.bg-purple-950{background-color:var(--color-purple-950)}.bg-red-50{background-color:var(--color-red-50)}.bg-red-100{background-color:var(--color-red-100)}.bg-red-200{background-color:var(--color-red-200)}.bg-red-300{background-color:var(--color-red-300)}.bg-red-400{background-color:var(--color-red-400)}.bg-red-500{background-color:var(--color-red-500)}.bg-red-500\\/5{background-color:#fb2c360d}@supports (color:color-mix(in lab, red, red)){.bg-red-500\\/5{background-color:color-mix(in oklab, var(--color-red-500) 5%, transparent)}}.bg-red-500\\/10{background-color:#fb2c361a}@supports (color:color-mix(in lab, red, red)){.bg-red-500\\/10{background-color:color-mix(in oklab, var(--color-red-500) 10%, transparent)}}.bg-red-500\\/20{background-color:#fb2c3633}@supports (color:color-mix(in lab, red, red)){.bg-red-500\\/20{background-color:color-mix(in oklab, var(--color-red-500) 20%, transparent)}}.bg-red-500\\/80{background-color:#fb2c36cc}@supports (color:color-mix(in lab, red, red)){.bg-red-500\\/80{background-color:color-mix(in oklab, var(--color-red-500) 80%, transparent)}}.bg-red-600{background-color:var(--color-red-600)}.bg-red-700{background-color:var(--color-red-700)}.bg-red-800{background-color:var(--color-red-800)}.bg-red-900{background-color:var(--color-red-900)}.bg-red-950{background-color:var(--color-red-950)}.bg-rose-50{background-color:var(--color-rose-50)}.bg-rose-50\\/85{background-color:#fff1f2d9}@supports (color:color-mix(in lab, red, red)){.bg-rose-50\\/85{background-color:color-mix(in oklab, var(--color-rose-50) 85%, transparent)}}.bg-rose-100{background-color:var(--color-rose-100)}.bg-rose-200{background-color:var(--color-rose-200)}.bg-rose-300{background-color:var(--color-rose-300)}.bg-rose-400{background-color:var(--color-rose-400)}.bg-rose-500{background-color:var(--color-rose-500)}.bg-rose-500\\/10{background-color:#ff23571a}@supports (color:color-mix(in lab, red, red)){.bg-rose-500\\/10{background-color:color-mix(in oklab, var(--color-rose-500) 10%, transparent)}}.bg-rose-500\\/12{background-color:#ff23571f}@supports (color:color-mix(in lab, red, red)){.bg-rose-500\\/12{background-color:color-mix(in oklab, var(--color-rose-500) 12%, transparent)}}.bg-rose-500\\/15{background-color:#ff235726}@supports (color:color-mix(in lab, red, red)){.bg-rose-500\\/15{background-color:color-mix(in oklab, var(--color-rose-500) 15%, transparent)}}.bg-rose-600{background-color:var(--color-rose-600)}.bg-rose-700{background-color:var(--color-rose-700)}.bg-rose-800{background-color:var(--color-rose-800)}.bg-rose-900{background-color:var(--color-rose-900)}.bg-rose-950{background-color:var(--color-rose-950)}.bg-secondary,.bg-secondary\\/50{background-color:var(--secondary)}@supports (color:color-mix(in lab, red, red)){.bg-secondary\\/50{background-color:color-mix(in oklab, var(--secondary) 50%, transparent)}}.bg-sidebar{background-color:var(--sidebar)}.bg-sidebar-accent,.bg-sidebar-accent\\/50{background-color:var(--sidebar-accent)}@supports (color:color-mix(in lab, red, red)){.bg-sidebar-accent\\/50{background-color:color-mix(in oklab, var(--sidebar-accent) 50%, transparent)}}.bg-sky-50{background-color:var(--color-sky-50)}.bg-sky-100{background-color:var(--color-sky-100)}.bg-sky-200{background-color:var(--color-sky-200)}.bg-sky-300{background-color:var(--color-sky-300)}.bg-sky-400{background-color:var(--color-sky-400)}.bg-sky-500{background-color:var(--color-sky-500)}.bg-sky-500\\/5{background-color:#00a5ef0d}@supports (color:color-mix(in lab, red, red)){.bg-sky-500\\/5{background-color:color-mix(in oklab, var(--color-sky-500) 5%, transparent)}}.bg-sky-500\\/10{background-color:#00a5ef1a}@supports (color:color-mix(in lab, red, red)){.bg-sky-500\\/10{background-color:color-mix(in oklab, var(--color-sky-500) 10%, transparent)}}.bg-sky-500\\/12{background-color:#00a5ef1f}@supports (color:color-mix(in lab, red, red)){.bg-sky-500\\/12{background-color:color-mix(in oklab, var(--color-sky-500) 12%, transparent)}}.bg-sky-500\\/15{background-color:#00a5ef26}@supports (color:color-mix(in lab, red, red)){.bg-sky-500\\/15{background-color:color-mix(in oklab, var(--color-sky-500) 15%, transparent)}}.bg-sky-600{background-color:var(--color-sky-600)}.bg-sky-700{background-color:var(--color-sky-700)}.bg-sky-800{background-color:var(--color-sky-800)}.bg-sky-900{background-color:var(--color-sky-900)}.bg-sky-950{background-color:var(--color-sky-950)}.bg-slate-50{background-color:var(--color-slate-50)}.bg-slate-100{background-color:var(--color-slate-100)}.bg-slate-200{background-color:var(--color-slate-200)}.bg-slate-300{background-color:var(--color-slate-300)}.bg-slate-400{background-color:var(--color-slate-400)}.bg-slate-500{background-color:var(--color-slate-500)}.bg-slate-600{background-color:var(--color-slate-600)}.bg-slate-700{background-color:var(--color-slate-700)}.bg-slate-800{background-color:var(--color-slate-800)}.bg-slate-900{background-color:var(--color-slate-900)}.bg-slate-950{background-color:var(--color-slate-950)}.bg-stone-50{background-color:var(--color-stone-50)}.bg-stone-100{background-color:var(--color-stone-100)}.bg-stone-200{background-color:var(--color-stone-200)}.bg-stone-300{background-color:var(--color-stone-300)}.bg-stone-400{background-color:var(--color-stone-400)}.bg-stone-500{background-color:var(--color-stone-500)}.bg-stone-600{background-color:var(--color-stone-600)}.bg-stone-700{background-color:var(--color-stone-700)}.bg-stone-800{background-color:var(--color-stone-800)}.bg-stone-900{background-color:var(--color-stone-900)}.bg-stone-950{background-color:var(--color-stone-950)}.bg-teal-50{background-color:var(--color-teal-50)}.bg-teal-100{background-color:var(--color-teal-100)}.bg-teal-200{background-color:var(--color-teal-200)}.bg-teal-300{background-color:var(--color-teal-300)}.bg-teal-400{background-color:var(--color-teal-400)}.bg-teal-500{background-color:var(--color-teal-500)}.bg-teal-600{background-color:var(--color-teal-600)}.bg-teal-700{background-color:var(--color-teal-700)}.bg-teal-800{background-color:var(--color-teal-800)}.bg-teal-900{background-color:var(--color-teal-900)}.bg-teal-950{background-color:var(--color-teal-950)}.bg-transparent{background-color:#0000}.bg-violet-50{background-color:var(--color-violet-50)}.bg-violet-100{background-color:var(--color-violet-100)}.bg-violet-200{background-color:var(--color-violet-200)}.bg-violet-300{background-color:var(--color-violet-300)}.bg-violet-400{background-color:var(--color-violet-400)}.bg-violet-500{background-color:var(--color-violet-500)}.bg-violet-500\\/5{background-color:#8d54ff0d}@supports (color:color-mix(in lab, red, red)){.bg-violet-500\\/5{background-color:color-mix(in oklab, var(--color-violet-500) 5%, transparent)}}.bg-violet-500\\/10{background-color:#8d54ff1a}@supports (color:color-mix(in lab, red, red)){.bg-violet-500\\/10{background-color:color-mix(in oklab, var(--color-violet-500) 10%, transparent)}}.bg-violet-600{background-color:var(--color-violet-600)}.bg-violet-700{background-color:var(--color-violet-700)}.bg-violet-800{background-color:var(--color-violet-800)}.bg-violet-900{background-color:var(--color-violet-900)}.bg-violet-950{background-color:var(--color-violet-950)}.bg-white{background-color:var(--color-white)}.bg-white\\/10{background-color:#ffffff1a}@supports (color:color-mix(in lab, red, red)){.bg-white\\/10{background-color:color-mix(in oklab, var(--color-white) 10%, transparent)}}.bg-yellow-50{background-color:var(--color-yellow-50)}.bg-yellow-100{background-color:var(--color-yellow-100)}.bg-yellow-200{background-color:var(--color-yellow-200)}.bg-yellow-300{background-color:var(--color-yellow-300)}.bg-yellow-400{background-color:var(--color-yellow-400)}.bg-yellow-500{background-color:var(--color-yellow-500)}.bg-yellow-500\\/5{background-color:#edb2000d}@supports (color:color-mix(in lab, red, red)){.bg-yellow-500\\/5{background-color:color-mix(in oklab, var(--color-yellow-500) 5%, transparent)}}.bg-yellow-600{background-color:var(--color-yellow-600)}.bg-yellow-700{background-color:var(--color-yellow-700)}.bg-yellow-800{background-color:var(--color-yellow-800)}.bg-yellow-900{background-color:var(--color-yellow-900)}.bg-yellow-950{background-color:var(--color-yellow-950)}.bg-zinc-50{background-color:var(--color-zinc-50)}.bg-zinc-100{background-color:var(--color-zinc-100)}.bg-zinc-200{background-color:var(--color-zinc-200)}.bg-zinc-300{background-color:var(--color-zinc-300)}.bg-zinc-400{background-color:var(--color-zinc-400)}.bg-zinc-500{background-color:var(--color-zinc-500)}.bg-zinc-600{background-color:var(--color-zinc-600)}.bg-zinc-700{background-color:var(--color-zinc-700)}.bg-zinc-800{background-color:var(--color-zinc-800)}.bg-zinc-900{background-color:var(--color-zinc-900)}.bg-zinc-950{background-color:var(--color-zinc-950)}.bg-gradient-to-l{--tw-gradient-position:to left in oklab;background-image:linear-gradient(var(--tw-gradient-stops))}.bg-gradient-to-r{--tw-gradient-position:to right in oklab;background-image:linear-gradient(var(--tw-gradient-stops))}.bg-gradient-to-t{--tw-gradient-position:to top in oklab;background-image:linear-gradient(var(--tw-gradient-stops))}.from-background{--tw-gradient-from:var(--background);--tw-gradient-stops:var(--tw-gradient-via-stops,var(--tw-gradient-position), var(--tw-gradient-from) var(--tw-gradient-from-position), var(--tw-gradient-to) var(--tw-gradient-to-position))}.from-border\\/30{--tw-gradient-from:var(--border)}@supports (color:color-mix(in lab, red, red)){.from-border\\/30{--tw-gradient-from:color-mix(in oklab, var(--border) 30%, transparent)}}.from-border\\/30{--tw-gradient-stops:var(--tw-gradient-via-stops,var(--tw-gradient-position), var(--tw-gradient-from) var(--tw-gradient-from-position), var(--tw-gradient-to) var(--tw-gradient-to-position))}.to-transparent{--tw-gradient-to:transparent;--tw-gradient-stops:var(--tw-gradient-via-stops,var(--tw-gradient-position), var(--tw-gradient-from) var(--tw-gradient-from-position), var(--tw-gradient-to) var(--tw-gradient-to-position))}.fill-current{fill:currentColor}.fill-foreground{fill:var(--foreground)}.fill-primary\\/20{fill:var(--primary)}@supports (color:color-mix(in lab, red, red)){.fill-primary\\/20{fill:color-mix(in oklab, var(--primary) 20%, transparent)}}.fill-primary\\/50{fill:var(--primary)}@supports (color:color-mix(in lab, red, red)){.fill-primary\\/50{fill:color-mix(in oklab, var(--primary) 50%, transparent)}}.fill-red-500{fill:var(--color-red-500)}.fill-sky-400{fill:var(--color-sky-400)}.stroke-amber-600{stroke:var(--color-amber-600)}.stroke-background{stroke:var(--background)}.stroke-primary,.stroke-primary\\/15{stroke:var(--primary)}@supports (color:color-mix(in lab, red, red)){.stroke-primary\\/15{stroke:color-mix(in oklab, var(--primary) 15%, transparent)}}.stroke-2{stroke-width:2px}.stroke-4{stroke-width:4px}.object-contain{object-fit:contain}.object-cover{object-fit:cover}.\\!p-0{padding:0!important}.\\!p-1{padding:var(--spacing)!important}.p-0{padding:0}.p-0\\.5{padding:calc(var(--spacing) * .5)}.p-1{padding:var(--spacing)}.p-1\\.5{padding:calc(var(--spacing) * 1.5)}.p-2{padding:calc(var(--spacing) * 2)}.p-2\\.5{padding:calc(var(--spacing) * 2.5)}.p-3{padding:calc(var(--spacing) * 3)}.p-4{padding:calc(var(--spacing) * 4)}.p-5{padding:calc(var(--spacing) * 5)}.p-6{padding:calc(var(--spacing) * 6)}.p-8{padding:calc(var(--spacing) * 8)}.p-10{padding:calc(var(--spacing) * 10)}.p-density-1{padding:var(--spacing-density-1)}.p-density-2{padding:var(--spacing-density-2)}.p-density-3{padding:var(--spacing-density-3)}.p-density-4{padding:var(--spacing-density-4)}.p-density-6{padding:var(--spacing-density-6)}.p-density-8{padding:var(--spacing-density-8)}.px-0{padding-inline:0}.px-0\\.5{padding-inline:calc(var(--spacing) * .5)}.px-0\\.25{padding-inline:calc(var(--spacing) * .25)}.px-1{padding-inline:var(--spacing)}.px-1\\.5{padding-inline:calc(var(--spacing) * 1.5)}.px-2{padding-inline:calc(var(--spacing) * 2)}.px-2\\.5{padding-inline:calc(var(--spacing) * 2.5)}.px-3{padding-inline:calc(var(--spacing) * 3)}.px-3\\.5{padding-inline:calc(var(--spacing) * 3.5)}.px-4{padding-inline:calc(var(--spacing) * 4)}.px-6{padding-inline:calc(var(--spacing) * 6)}.px-8{padding-inline:calc(var(--spacing) * 8)}.px-\\[3px\\]{padding-inline:3px}.px-control-px{padding-inline:var(--spacing-control-px)}.px-density-1{padding-inline:var(--spacing-density-1)}.px-density-2{padding-inline:var(--spacing-density-2)}.px-density-3{padding-inline:var(--spacing-density-3)}.px-density-4{padding-inline:var(--spacing-density-4)}.py-0{padding-block:0}.py-0\\.5{padding-block:calc(var(--spacing) * .5)}.py-0\\.25{padding-block:calc(var(--spacing) * .25)}.py-1{padding-block:var(--spacing)}.py-1\\.5{padding-block:calc(var(--spacing) * 1.5)}.py-2{padding-block:calc(var(--spacing) * 2)}.py-2\\.5{padding-block:calc(var(--spacing) * 2.5)}.py-3{padding-block:calc(var(--spacing) * 3)}.py-4{padding-block:calc(var(--spacing) * 4)}.py-5{padding-block:calc(var(--spacing) * 5)}.py-6{padding-block:calc(var(--spacing) * 6)}.py-8{padding-block:calc(var(--spacing) * 8)}.py-10{padding-block:calc(var(--spacing) * 10)}.py-\\[var\\(--json-view-row-space\\)\\]{padding-block:var(--json-view-row-space)}.py-density-1{padding-block:var(--spacing-density-1)}.py-density-2{padding-block:var(--spacing-density-2)}.py-density-3{padding-block:var(--spacing-density-3)}.py-density-4{padding-block:var(--spacing-density-4)}.py-px{padding-block:1px}.pt-0{padding-top:0}.pt-0\\.5{padding-top:calc(var(--spacing) * .5)}.pt-1{padding-top:var(--spacing)}.pt-1\\.5{padding-top:calc(var(--spacing) * 1.5)}.pt-2{padding-top:calc(var(--spacing) * 2)}.pt-3{padding-top:calc(var(--spacing) * 3)}.pt-3\\.5{padding-top:calc(var(--spacing) * 3.5)}.pt-4{padding-top:calc(var(--spacing) * 4)}.pt-9{padding-top:calc(var(--spacing) * 9)}.pt-\\[env\\(safe-area-inset-top\\)\\]{padding-top:env(safe-area-inset-top)}.pt-density-2{padding-top:var(--spacing-density-2)}.pt-density-3{padding-top:var(--spacing-density-3)}.pt-density-4{padding-top:var(--spacing-density-4)}.pt-px{padding-top:1px}.pr-0{padding-right:0}.pr-0\\.5{padding-right:calc(var(--spacing) * .5)}.pr-1{padding-right:var(--spacing)}.pr-1\\.5{padding-right:calc(var(--spacing) * 1.5)}.pr-2{padding-right:calc(var(--spacing) * 2)}.pr-3{padding-right:calc(var(--spacing) * 3)}.pr-4{padding-right:calc(var(--spacing) * 4)}.pr-5{padding-right:calc(var(--spacing) * 5)}.pr-6{padding-right:calc(var(--spacing) * 6)}.pr-7{padding-right:calc(var(--spacing) * 7)}.pr-8{padding-right:calc(var(--spacing) * 8)}.pr-9{padding-right:calc(var(--spacing) * 9)}.pr-10{padding-right:calc(var(--spacing) * 10)}.pr-12{padding-right:calc(var(--spacing) * 12)}.pr-14{padding-right:calc(var(--spacing) * 14)}.pr-16{padding-right:calc(var(--spacing) * 16)}.pr-20{padding-right:calc(var(--spacing) * 20)}.pr-24{padding-right:calc(var(--spacing) * 24)}.pr-28{padding-right:calc(var(--spacing) * 28)}.pr-\\[5\\.5rem\\]{padding-right:5.5rem}.pr-\\[env\\(safe-area-inset-right\\)\\]{padding-right:env(safe-area-inset-right)}.pr-\\[var\\(--workspace-top-right-reserve\\,0px\\)\\]{padding-right:var(--workspace-top-right-reserve,0px)}.pr-density-3{padding-right:var(--spacing-density-3)}.pb-1{padding-bottom:var(--spacing)}.pb-1\\.5{padding-bottom:calc(var(--spacing) * 1.5)}.pb-2{padding-bottom:calc(var(--spacing) * 2)}.pb-3{padding-bottom:calc(var(--spacing) * 3)}.pb-4{padding-bottom:calc(var(--spacing) * 4)}.pb-10{padding-bottom:calc(var(--spacing) * 10)}.pb-\\[max\\(1rem\\,env\\(safe-area-inset-bottom\\)\\)\\]{padding-bottom:max(1rem, env(safe-area-inset-bottom))}.pb-\\[max\\(calc\\(var\\(--spacing-unit\\)\\*3\\)\\,env\\(safe-area-inset-bottom\\)\\)\\]{padding-bottom:max(calc(var(--spacing-unit) * 3), env(safe-area-inset-bottom))}.pb-density-2{padding-bottom:var(--spacing-density-2)}.pb-density-3{padding-bottom:var(--spacing-density-3)}.pb-density-4{padding-bottom:var(--spacing-density-4)}.pl-0{padding-left:0}.pl-1{padding-left:var(--spacing)}.pl-2{padding-left:calc(var(--spacing) * 2)}.pl-3{padding-left:calc(var(--spacing) * 3)}.pl-4{padding-left:calc(var(--spacing) * 4)}.pl-5{padding-left:calc(var(--spacing) * 5)}.pl-6{padding-left:calc(var(--spacing) * 6)}.pl-7{padding-left:calc(var(--spacing) * 7)}.pl-8{padding-left:calc(var(--spacing) * 8)}.pl-\\[18px\\]{padding-left:18px}.pl-\\[env\\(safe-area-inset-left\\)\\]{padding-left:env(safe-area-inset-left)}.pl-\\[var\\(--json-view-indent\\)\\]{padding-left:var(--json-view-indent)}.pl-density-3{padding-left:var(--spacing-density-3)}.pl-px{padding-left:1px}.text-center{text-align:center}.text-end{text-align:end}.text-justify{text-align:justify}.text-left{text-align:left}.text-right{text-align:right}.text-start{text-align:start}.align-\\[-0\\.15em\\]{vertical-align:-.15em}.align-\\[-0\\.125em\\]{vertical-align:-.125em}.align-baseline{vertical-align:baseline}.align-bottom{vertical-align:bottom}.align-middle{vertical-align:middle}.align-super{vertical-align:super}.align-top{vertical-align:top}.font-mono{font-family:var(--clicky-font-mono)}.font-sans{font-family:var(--clicky-font-sans)}.font-serif{font-family:var(--font-serif)}.text-2xl{font-size:var(--text-2xl);line-height:var(--tw-leading,var(--text-2xl--line-height))}.text-3xl{font-size:var(--text-3xl);line-height:var(--tw-leading,var(--text-3xl--line-height))}.text-4xl{font-size:var(--text-4xl);line-height:var(--tw-leading,var(--text-4xl--line-height))}.text-5xl{font-size:var(--text-5xl);line-height:var(--tw-leading,var(--text-5xl--line-height))}.text-6xl{font-size:var(--text-6xl);line-height:var(--tw-leading,var(--text-6xl--line-height))}.text-7xl{font-size:var(--text-7xl);line-height:var(--tw-leading,var(--text-7xl--line-height))}.text-8xl{font-size:var(--text-8xl);line-height:var(--tw-leading,var(--text-8xl--line-height))}.text-9xl{font-size:var(--text-9xl);line-height:var(--tw-leading,var(--text-9xl--line-height))}.text-base{font-size:var(--text-base);line-height:var(--tw-leading,var(--text-base--line-height))}.text-lg{font-size:var(--text-lg);line-height:var(--tw-leading,var(--text-lg--line-height))}.text-sm{font-size:var(--text-sm);line-height:var(--tw-leading,var(--text-sm--line-height))}.text-xl{font-size:var(--text-xl);line-height:var(--tw-leading,var(--text-xl--line-height))}.text-xs{font-size:var(--text-xs);line-height:var(--tw-leading,var(--text-xs--line-height))}.text-\\[0\\.6rem\\]{font-size:.6rem}.text-\\[0\\.7rem\\]{font-size:.7rem}.text-\\[0\\.9em\\]{font-size:.9em}.text-\\[0\\.65em\\]{font-size:.65em}.text-\\[0\\.65rem\\]{font-size:.65rem}.text-\\[0\\.75em\\]{font-size:.75em}.text-\\[0\\.85em\\]{font-size:.85em}.text-\\[1em\\]{font-size:1em}.text-\\[9\\.5px\\]{font-size:9.5px}.text-\\[9px\\]{font-size:9px}.text-\\[10px\\]{font-size:10px}.text-\\[11\\.5px\\]{font-size:11.5px}.text-\\[11px\\]{font-size:11px}.text-\\[12px\\]{font-size:12px}.text-\\[13px\\]{font-size:13px}.text-\\[14px\\]{font-size:14px}.text-\\[15px\\]{font-size:15px}.text-\\[length\\:inherit\\]{font-size:inherit}.leading-3\\.5{--tw-leading:calc(var(--spacing) * 3.5);line-height:calc(var(--spacing) * 3.5)}.leading-4{--tw-leading:calc(var(--spacing) * 4);line-height:calc(var(--spacing) * 4)}.leading-5{--tw-leading:calc(var(--spacing) * 5);line-height:calc(var(--spacing) * 5)}.leading-6{--tw-leading:calc(var(--spacing) * 6);line-height:calc(var(--spacing) * 6)}.leading-7{--tw-leading:calc(var(--spacing) * 7);line-height:calc(var(--spacing) * 7)}.leading-\\[8px\\]{--tw-leading:8px;line-height:8px}.leading-\\[inherit\\]{--tw-leading:inherit;line-height:inherit}.leading-none{--tw-leading:1;line-height:1}.leading-normal{--tw-leading:var(--leading-normal);line-height:var(--leading-normal)}.leading-relaxed{--tw-leading:var(--leading-relaxed);line-height:var(--leading-relaxed)}.leading-snug{--tw-leading:var(--leading-snug);line-height:var(--leading-snug)}.leading-tight{--tw-leading:var(--leading-tight);line-height:var(--leading-tight)}.font-black{--tw-font-weight:var(--font-weight-black);font-weight:var(--font-weight-black)}.font-bold{--tw-font-weight:var(--font-weight-bold);font-weight:var(--font-weight-bold)}.font-extrabold{--tw-font-weight:var(--font-weight-extrabold);font-weight:var(--font-weight-extrabold)}.font-extralight{--tw-font-weight:var(--font-weight-extralight);font-weight:var(--font-weight-extralight)}.font-light{--tw-font-weight:var(--font-weight-light);font-weight:var(--font-weight-light)}.font-medium{--tw-font-weight:var(--font-weight-medium);font-weight:var(--font-weight-medium)}.font-normal{--tw-font-weight:var(--font-weight-normal);font-weight:var(--font-weight-normal)}.font-semibold{--tw-font-weight:var(--font-weight-semibold);font-weight:var(--font-weight-semibold)}.font-thin{--tw-font-weight:var(--font-weight-thin);font-weight:var(--font-weight-thin)}.tracking-\\[-0\\.02em\\]{--tw-tracking:-.02em;letter-spacing:-.02em}.tracking-\\[-0\\.05em\\]{--tw-tracking:-.05em;letter-spacing:-.05em}.tracking-\\[0\\.2em\\]{--tw-tracking:.2em;letter-spacing:.2em}.tracking-\\[0\\.03em\\]{--tw-tracking:.03em;letter-spacing:.03em}.tracking-\\[0\\.12em\\]{--tw-tracking:.12em;letter-spacing:.12em}.tracking-\\[0\\.14em\\]{--tw-tracking:.14em;letter-spacing:.14em}.tracking-\\[0\\.18em\\]{--tw-tracking:.18em;letter-spacing:.18em}.tracking-\\[0\\.24em\\]{--tw-tracking:.24em;letter-spacing:.24em}.tracking-normal{--tw-tracking:var(--tracking-normal);letter-spacing:var(--tracking-normal)}.tracking-tight{--tw-tracking:var(--tracking-tight);letter-spacing:var(--tracking-tight)}.tracking-wide{--tw-tracking:var(--tracking-wide);letter-spacing:var(--tracking-wide)}.tracking-wider{--tw-tracking:var(--tracking-wider);letter-spacing:var(--tracking-wider)}.break-normal{overflow-wrap:normal;word-break:normal}.\\[overflow-wrap\\:anywhere\\]{overflow-wrap:anywhere}.break-words{overflow-wrap:break-word}.wrap-anywhere{overflow-wrap:anywhere}.wrap-break-word{overflow-wrap:break-word}.break-all{word-break:break-all}.break-keep{word-break:keep-all}.text-ellipsis{text-overflow:ellipsis}.whitespace-break-spaces{white-space:break-spaces}.whitespace-normal{white-space:normal}.whitespace-nowrap{white-space:nowrap}.whitespace-pre{white-space:pre}.whitespace-pre-line{white-space:pre-line}.whitespace-pre-wrap{white-space:pre-wrap}.text-\\[\\#0F766E\\]{color:#0f766e}.text-\\[\\#0e7490\\]{color:#0e7490}.text-\\[\\#00add8\\]{color:#00add8}.text-\\[\\#2ead33\\]{color:#2ead33}.text-\\[\\#4D6BFE\\]{color:#4d6bfe}.text-\\[\\#4b32c3\\]{color:#4b32c3}.text-\\[\\#6e9f18\\]{color:#6e9f18}.text-\\[\\#111\\]{color:#111}.text-\\[\\#0668E1\\]{color:#0668e1}.text-\\[\\#0891b2\\]{color:#0891b2}.text-\\[\\#3178c6\\]{color:#3178c6}.text-\\[\\#3776ab\\]{color:#3776ab}.text-\\[\\#4285F4\\]{color:#4285f4}.text-\\[\\#261230\\]{color:#261230}.text-\\[\\#A16207\\]{color:#a16207}.text-\\[\\#C15F3C\\]{color:#c15f3c}.text-\\[\\#C2410C\\]{color:#c2410c}.text-\\[\\#c21325\\]{color:#c21325}.text-\\[\\#fff\\]{color:#fff}.text-\\[hsl\\(0\\,0\\%\\,0\\%\\)\\]{color:#000}.text-\\[var\\(--clicky-error\\)\\]{color:var(--clicky-error)}.text-\\[var\\(--clicky-info\\)\\]{color:var(--clicky-info)}.text-\\[var\\(--clicky-success\\)\\]{color:var(--clicky-success)}.text-\\[var\\(--clicky-warning\\)\\]{color:var(--clicky-warning)}.text-\\[var\\(--rg-error-text\\)\\]{color:var(--rg-error-text)}.text-\\[var\\(--rg-submit-fg\\)\\]{color:var(--rg-submit-fg)}.text-\\[var\\(--rg-text-primary\\)\\]{color:var(--rg-text-primary)}.text-\\[var\\(--rg-text-primary-85\\)\\]{color:var(--rg-text-primary-85)}.text-\\[var\\(--rg-text-secondary\\)\\]{color:var(--rg-text-secondary)}.text-accent-foreground{color:var(--accent-foreground)}.text-amber-50{color:var(--color-amber-50)}.text-amber-100{color:var(--color-amber-100)}.text-amber-200{color:var(--color-amber-200)}.text-amber-300{color:var(--color-amber-300)}.text-amber-400{color:var(--color-amber-400)}.text-amber-500{color:var(--color-amber-500)}.text-amber-600{color:var(--color-amber-600)}.text-amber-600\\/70{color:#dd7400b3}@supports (color:color-mix(in lab, red, red)){.text-amber-600\\/70{color:color-mix(in oklab, var(--color-amber-600) 70%, transparent)}}.text-amber-700{color:var(--color-amber-700)}.text-amber-800{color:var(--color-amber-800)}.text-amber-900{color:var(--color-amber-900)}.text-amber-950{color:var(--color-amber-950)}.text-amber-950\\/90{color:#461901e6}@supports (color:color-mix(in lab, red, red)){.text-amber-950\\/90{color:color-mix(in oklab, var(--color-amber-950) 90%, transparent)}}.text-background{color:var(--background)}.text-black{color:var(--color-black)}.text-blue-50{color:var(--color-blue-50)}.text-blue-100{color:var(--color-blue-100)}.text-blue-200{color:var(--color-blue-200)}.text-blue-300{color:var(--color-blue-300)}.text-blue-400{color:var(--color-blue-400)}.text-blue-500{color:var(--color-blue-500)}.text-blue-600{color:var(--color-blue-600)}.text-blue-600\\/70{color:#155dfcb3}@supports (color:color-mix(in lab, red, red)){.text-blue-600\\/70{color:color-mix(in oklab, var(--color-blue-600) 70%, transparent)}}.text-blue-700{color:var(--color-blue-700)}.text-blue-800{color:var(--color-blue-800)}.text-blue-900{color:var(--color-blue-900)}.text-blue-950{color:var(--color-blue-950)}.text-blue-950\\/90{color:#162456e6}@supports (color:color-mix(in lab, red, red)){.text-blue-950\\/90{color:color-mix(in oklab, var(--color-blue-950) 90%, transparent)}}.text-border{color:var(--border)}.text-card-foreground{color:var(--card-foreground)}.text-current{color:currentColor}.text-cyan-50{color:var(--color-cyan-50)}.text-cyan-100{color:var(--color-cyan-100)}.text-cyan-200{color:var(--color-cyan-200)}.text-cyan-300{color:var(--color-cyan-300)}.text-cyan-400{color:var(--color-cyan-400)}.text-cyan-500{color:var(--color-cyan-500)}.text-cyan-600{color:var(--color-cyan-600)}.text-cyan-700{color:var(--color-cyan-700)}.text-cyan-800{color:var(--color-cyan-800)}.text-cyan-900{color:var(--color-cyan-900)}.text-cyan-950{color:var(--color-cyan-950)}.text-destructive{color:var(--destructive)}.text-destructive-foreground{color:var(--destructive-foreground)}.text-emerald-50{color:var(--color-emerald-50)}.text-emerald-100{color:var(--color-emerald-100)}.text-emerald-200{color:var(--color-emerald-200)}.text-emerald-300{color:var(--color-emerald-300)}.text-emerald-400{color:var(--color-emerald-400)}.text-emerald-500{color:var(--color-emerald-500)}.text-emerald-600{color:var(--color-emerald-600)}.text-emerald-600\\/70{color:#009767b3}@supports (color:color-mix(in lab, red, red)){.text-emerald-600\\/70{color:color-mix(in oklab, var(--color-emerald-600) 70%, transparent)}}.text-emerald-700{color:var(--color-emerald-700)}.text-emerald-800{color:var(--color-emerald-800)}.text-emerald-900{color:var(--color-emerald-900)}.text-emerald-950{color:var(--color-emerald-950)}.text-emerald-950\\/90{color:#002c22e6}@supports (color:color-mix(in lab, red, red)){.text-emerald-950\\/90{color:color-mix(in oklab, var(--color-emerald-950) 90%, transparent)}}.text-foreground,.text-foreground\\/60{color:var(--foreground)}@supports (color:color-mix(in lab, red, red)){.text-foreground\\/60{color:color-mix(in oklab, var(--foreground) 60%, transparent)}}.text-foreground\\/70{color:var(--foreground)}@supports (color:color-mix(in lab, red, red)){.text-foreground\\/70{color:color-mix(in oklab, var(--foreground) 70%, transparent)}}.text-foreground\\/80{color:var(--foreground)}@supports (color:color-mix(in lab, red, red)){.text-foreground\\/80{color:color-mix(in oklab, var(--foreground) 80%, transparent)}}.text-fuchsia-50{color:var(--color-fuchsia-50)}.text-fuchsia-100{color:var(--color-fuchsia-100)}.text-fuchsia-200{color:var(--color-fuchsia-200)}.text-fuchsia-300{color:var(--color-fuchsia-300)}.text-fuchsia-400{color:var(--color-fuchsia-400)}.text-fuchsia-500{color:var(--color-fuchsia-500)}.text-fuchsia-600{color:var(--color-fuchsia-600)}.text-fuchsia-700{color:var(--color-fuchsia-700)}.text-fuchsia-800{color:var(--color-fuchsia-800)}.text-fuchsia-900{color:var(--color-fuchsia-900)}.text-fuchsia-950{color:var(--color-fuchsia-950)}.text-gray-50{color:var(--color-gray-50)}.text-gray-100{color:var(--color-gray-100)}.text-gray-200{color:var(--color-gray-200)}.text-gray-300{color:var(--color-gray-300)}.text-gray-400{color:var(--color-gray-400)}.text-gray-500{color:var(--color-gray-500)}.text-gray-600{color:var(--color-gray-600)}.text-gray-700{color:var(--color-gray-700)}.text-gray-800{color:var(--color-gray-800)}.text-gray-900{color:var(--color-gray-900)}.text-gray-950{color:var(--color-gray-950)}.text-green-50{color:var(--color-green-50)}.text-green-100{color:var(--color-green-100)}.text-green-200{color:var(--color-green-200)}.text-green-300{color:var(--color-green-300)}.text-green-400{color:var(--color-green-400)}.text-green-500{color:var(--color-green-500)}.text-green-600{color:var(--color-green-600)}.text-green-700{color:var(--color-green-700)}.text-green-800{color:var(--color-green-800)}.text-green-900{color:var(--color-green-900)}.text-green-950{color:var(--color-green-950)}.text-indigo-50{color:var(--color-indigo-50)}.text-indigo-100{color:var(--color-indigo-100)}.text-indigo-200{color:var(--color-indigo-200)}.text-indigo-300{color:var(--color-indigo-300)}.text-indigo-400{color:var(--color-indigo-400)}.text-indigo-500{color:var(--color-indigo-500)}.text-indigo-600{color:var(--color-indigo-600)}.text-indigo-700{color:var(--color-indigo-700)}.text-indigo-800{color:var(--color-indigo-800)}.text-indigo-900{color:var(--color-indigo-900)}.text-indigo-950{color:var(--color-indigo-950)}.text-inherit{color:inherit}.text-lime-50{color:var(--color-lime-50)}.text-lime-100{color:var(--color-lime-100)}.text-lime-200{color:var(--color-lime-200)}.text-lime-300{color:var(--color-lime-300)}.text-lime-400{color:var(--color-lime-400)}.text-lime-500{color:var(--color-lime-500)}.text-lime-600{color:var(--color-lime-600)}.text-lime-700{color:var(--color-lime-700)}.text-lime-800{color:var(--color-lime-800)}.text-lime-900{color:var(--color-lime-900)}.text-lime-950{color:var(--color-lime-950)}.text-muted{color:var(--muted)}.text-muted-foreground,.text-muted-foreground\\/20{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/20{color:color-mix(in oklab, var(--muted-foreground) 20%, transparent)}}.text-muted-foreground\\/45{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/45{color:color-mix(in oklab, var(--muted-foreground) 45%, transparent)}}.text-muted-foreground\\/50{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/50{color:color-mix(in oklab, var(--muted-foreground) 50%, transparent)}}.text-muted-foreground\\/60{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/60{color:color-mix(in oklab, var(--muted-foreground) 60%, transparent)}}.text-muted-foreground\\/70{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/70{color:color-mix(in oklab, var(--muted-foreground) 70%, transparent)}}.text-muted-foreground\\/80{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.text-muted-foreground\\/80{color:color-mix(in oklab, var(--muted-foreground) 80%, transparent)}}.text-neutral-50{color:var(--color-neutral-50)}.text-neutral-100{color:var(--color-neutral-100)}.text-neutral-200{color:var(--color-neutral-200)}.text-neutral-300{color:var(--color-neutral-300)}.text-neutral-400{color:var(--color-neutral-400)}.text-neutral-500{color:var(--color-neutral-500)}.text-neutral-600{color:var(--color-neutral-600)}.text-neutral-700{color:var(--color-neutral-700)}.text-neutral-800{color:var(--color-neutral-800)}.text-neutral-900{color:var(--color-neutral-900)}.text-neutral-950{color:var(--color-neutral-950)}.text-orange-50{color:var(--color-orange-50)}.text-orange-100{color:var(--color-orange-100)}.text-orange-200{color:var(--color-orange-200)}.text-orange-300{color:var(--color-orange-300)}.text-orange-400{color:var(--color-orange-400)}.text-orange-500{color:var(--color-orange-500)}.text-orange-600{color:var(--color-orange-600)}.text-orange-700{color:var(--color-orange-700)}.text-orange-800{color:var(--color-orange-800)}.text-orange-900{color:var(--color-orange-900)}.text-orange-950{color:var(--color-orange-950)}.text-pink-50{color:var(--color-pink-50)}.text-pink-100{color:var(--color-pink-100)}.text-pink-200{color:var(--color-pink-200)}.text-pink-300{color:var(--color-pink-300)}.text-pink-400{color:var(--color-pink-400)}.text-pink-500{color:var(--color-pink-500)}.text-pink-600{color:var(--color-pink-600)}.text-pink-700{color:var(--color-pink-700)}.text-pink-800{color:var(--color-pink-800)}.text-pink-900{color:var(--color-pink-900)}.text-pink-950{color:var(--color-pink-950)}.text-placeholder{color:var(--placeholder)}.text-popover-foreground{color:var(--popover-foreground)}.text-primary{color:var(--primary)}.text-primary-foreground{color:var(--primary-foreground)}.text-primary\\/70{color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.text-primary\\/70{color:color-mix(in oklab, var(--primary) 70%, transparent)}}.text-purple-50{color:var(--color-purple-50)}.text-purple-100{color:var(--color-purple-100)}.text-purple-200{color:var(--color-purple-200)}.text-purple-300{color:var(--color-purple-300)}.text-purple-400{color:var(--color-purple-400)}.text-purple-500{color:var(--color-purple-500)}.text-purple-600{color:var(--color-purple-600)}.text-purple-600\\/70{color:#9810fab3}@supports (color:color-mix(in lab, red, red)){.text-purple-600\\/70{color:color-mix(in oklab, var(--color-purple-600) 70%, transparent)}}.text-purple-700{color:var(--color-purple-700)}.text-purple-800{color:var(--color-purple-800)}.text-purple-900{color:var(--color-purple-900)}.text-purple-950{color:var(--color-purple-950)}.text-purple-950\\/90{color:#3c0366e6}@supports (color:color-mix(in lab, red, red)){.text-purple-950\\/90{color:color-mix(in oklab, var(--color-purple-950) 90%, transparent)}}.text-red-50{color:var(--color-red-50)}.text-red-100{color:var(--color-red-100)}.text-red-200{color:var(--color-red-200)}.text-red-300{color:var(--color-red-300)}.text-red-400{color:var(--color-red-400)}.text-red-500{color:var(--color-red-500)}.text-red-600{color:var(--color-red-600)}.text-red-600\\/70{color:#e40014b3}@supports (color:color-mix(in lab, red, red)){.text-red-600\\/70{color:color-mix(in oklab, var(--color-red-600) 70%, transparent)}}.text-red-700{color:var(--color-red-700)}.text-red-800{color:var(--color-red-800)}.text-red-900{color:var(--color-red-900)}.text-red-950{color:var(--color-red-950)}.text-red-950\\/90{color:#460809e6}@supports (color:color-mix(in lab, red, red)){.text-red-950\\/90{color:color-mix(in oklab, var(--color-red-950) 90%, transparent)}}.text-rose-50{color:var(--color-rose-50)}.text-rose-100{color:var(--color-rose-100)}.text-rose-200{color:var(--color-rose-200)}.text-rose-300{color:var(--color-rose-300)}.text-rose-400{color:var(--color-rose-400)}.text-rose-500{color:var(--color-rose-500)}.text-rose-600{color:var(--color-rose-600)}.text-rose-700{color:var(--color-rose-700)}.text-rose-800{color:var(--color-rose-800)}.text-rose-900{color:var(--color-rose-900)}.text-rose-950{color:var(--color-rose-950)}.text-secondary-foreground{color:var(--secondary-foreground)}.text-sidebar-foreground,.text-sidebar-foreground\\/55{color:var(--sidebar-foreground)}@supports (color:color-mix(in lab, red, red)){.text-sidebar-foreground\\/55{color:color-mix(in oklab, var(--sidebar-foreground) 55%, transparent)}}.text-sidebar-foreground\\/60{color:var(--sidebar-foreground)}@supports (color:color-mix(in lab, red, red)){.text-sidebar-foreground\\/60{color:color-mix(in oklab, var(--sidebar-foreground) 60%, transparent)}}.text-sidebar-foreground\\/70{color:var(--sidebar-foreground)}@supports (color:color-mix(in lab, red, red)){.text-sidebar-foreground\\/70{color:color-mix(in oklab, var(--sidebar-foreground) 70%, transparent)}}.text-sidebar-primary{color:var(--sidebar-primary)}.text-sky-50{color:var(--color-sky-50)}.text-sky-100{color:var(--color-sky-100)}.text-sky-200{color:var(--color-sky-200)}.text-sky-300{color:var(--color-sky-300)}.text-sky-400{color:var(--color-sky-400)}.text-sky-500{color:var(--color-sky-500)}.text-sky-600{color:var(--color-sky-600)}.text-sky-700{color:var(--color-sky-700)}.text-sky-800{color:var(--color-sky-800)}.text-sky-900{color:var(--color-sky-900)}.text-sky-950{color:var(--color-sky-950)}.text-slate-50{color:var(--color-slate-50)}.text-slate-100{color:var(--color-slate-100)}.text-slate-200{color:var(--color-slate-200)}.text-slate-300{color:var(--color-slate-300)}.text-slate-400{color:var(--color-slate-400)}.text-slate-500{color:var(--color-slate-500)}.text-slate-600{color:var(--color-slate-600)}.text-slate-700{color:var(--color-slate-700)}.text-slate-800{color:var(--color-slate-800)}.text-slate-900{color:var(--color-slate-900)}.text-slate-950{color:var(--color-slate-950)}.text-stone-50{color:var(--color-stone-50)}.text-stone-100{color:var(--color-stone-100)}.text-stone-200{color:var(--color-stone-200)}.text-stone-300{color:var(--color-stone-300)}.text-stone-400{color:var(--color-stone-400)}.text-stone-500{color:var(--color-stone-500)}.text-stone-600{color:var(--color-stone-600)}.text-stone-700{color:var(--color-stone-700)}.text-stone-800{color:var(--color-stone-800)}.text-stone-900{color:var(--color-stone-900)}.text-stone-950{color:var(--color-stone-950)}.text-teal-50{color:var(--color-teal-50)}.text-teal-100{color:var(--color-teal-100)}.text-teal-200{color:var(--color-teal-200)}.text-teal-300{color:var(--color-teal-300)}.text-teal-400{color:var(--color-teal-400)}.text-teal-500{color:var(--color-teal-500)}.text-teal-600{color:var(--color-teal-600)}.text-teal-700{color:var(--color-teal-700)}.text-teal-800{color:var(--color-teal-800)}.text-teal-900{color:var(--color-teal-900)}.text-teal-950{color:var(--color-teal-950)}.text-violet-50{color:var(--color-violet-50)}.text-violet-100{color:var(--color-violet-100)}.text-violet-200{color:var(--color-violet-200)}.text-violet-300{color:var(--color-violet-300)}.text-violet-400{color:var(--color-violet-400)}.text-violet-500{color:var(--color-violet-500)}.text-violet-600{color:var(--color-violet-600)}.text-violet-700{color:var(--color-violet-700)}.text-violet-800{color:var(--color-violet-800)}.text-violet-900{color:var(--color-violet-900)}.text-violet-950{color:var(--color-violet-950)}.text-white{color:var(--color-white)}.text-yellow-50{color:var(--color-yellow-50)}.text-yellow-100{color:var(--color-yellow-100)}.text-yellow-200{color:var(--color-yellow-200)}.text-yellow-300{color:var(--color-yellow-300)}.text-yellow-400{color:var(--color-yellow-400)}.text-yellow-500{color:var(--color-yellow-500)}.text-yellow-600{color:var(--color-yellow-600)}.text-yellow-700{color:var(--color-yellow-700)}.text-yellow-800{color:var(--color-yellow-800)}.text-yellow-900{color:var(--color-yellow-900)}.text-yellow-950{color:var(--color-yellow-950)}.text-zinc-50{color:var(--color-zinc-50)}.text-zinc-100{color:var(--color-zinc-100)}.text-zinc-200{color:var(--color-zinc-200)}.text-zinc-300{color:var(--color-zinc-300)}.text-zinc-400{color:var(--color-zinc-400)}.text-zinc-500{color:var(--color-zinc-500)}.text-zinc-600{color:var(--color-zinc-600)}.text-zinc-700{color:var(--color-zinc-700)}.text-zinc-800{color:var(--color-zinc-800)}.text-zinc-900{color:var(--color-zinc-900)}.text-zinc-950{color:var(--color-zinc-950)}.capitalize{text-transform:capitalize}.lowercase{text-transform:lowercase}.normal-case{text-transform:none}.uppercase{text-transform:uppercase}.italic{font-style:italic}.not-italic{font-style:normal}.lining-nums{--tw-numeric-figure:lining-nums;font-variant-numeric:var(--tw-ordinal,) var(--tw-slashed-zero,) var(--tw-numeric-figure,) var(--tw-numeric-spacing,) var(--tw-numeric-fraction,)}.oldstyle-nums{--tw-numeric-figure:oldstyle-nums;font-variant-numeric:var(--tw-ordinal,) var(--tw-slashed-zero,) var(--tw-numeric-figure,) var(--tw-numeric-spacing,) var(--tw-numeric-fraction,)}.ordinal{--tw-ordinal:ordinal;font-variant-numeric:var(--tw-ordinal,) var(--tw-slashed-zero,) var(--tw-numeric-figure,) var(--tw-numeric-spacing,) var(--tw-numeric-fraction,)}.proportional-nums{--tw-numeric-spacing:proportional-nums;font-variant-numeric:var(--tw-ordinal,) var(--tw-slashed-zero,) var(--tw-numeric-figure,) var(--tw-numeric-spacing,) var(--tw-numeric-fraction,)}.slashed-zero{--tw-slashed-zero:slashed-zero;font-variant-numeric:var(--tw-ordinal,) var(--tw-slashed-zero,) var(--tw-numeric-figure,) var(--tw-numeric-spacing,) var(--tw-numeric-fraction,)}.tabular-nums{--tw-numeric-spacing:tabular-nums;font-variant-numeric:var(--tw-ordinal,) var(--tw-slashed-zero,) var(--tw-numeric-figure,) var(--tw-numeric-spacing,) var(--tw-numeric-fraction,)}.line-through{text-decoration-line:line-through}.no-underline{text-decoration-line:none}.underline{text-decoration-line:underline}.underline-offset-2{text-underline-offset:2px}.underline-offset-4{text-underline-offset:4px}.antialiased{-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}.accent-primary{accent-color:var(--primary)}.opacity-0{opacity:0}.opacity-10{opacity:.1}.opacity-20{opacity:.2}.opacity-30{opacity:.3}.opacity-40{opacity:.4}.opacity-50{opacity:.5}.opacity-60{opacity:.6}.opacity-70{opacity:.7}.opacity-75{opacity:.75}.opacity-80{opacity:.8}.opacity-85{opacity:.85}.opacity-90{opacity:.9}.opacity-100{opacity:1}.shadow{--tw-shadow:0 1px 3px 0 var(--tw-shadow-color,#0000001a), 0 1px 2px -1px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-\\[0_1px_0_0_var\\(--tw-shadow-color\\)\\]{--tw-shadow:0 1px 0 0 var(--tw-shadow-color,var(--tw-shadow-color));box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-\\[0_4px_16px_rgba\\(0\\,0\\,0\\,0\\.12\\)\\]{--tw-shadow:0 4px 16px var(--tw-shadow-color,#0000001f);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-\\[inset_0_-2px_0_0_var\\(--color-primary\\)\\]{--tw-shadow:inset 0 -2px 0 0 var(--tw-shadow-color,var(--color-primary));box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-\\[inset_0_2px_0_0_var\\(--color-primary\\)\\]{--tw-shadow:inset 0 2px 0 0 var(--tw-shadow-color,var(--color-primary));box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-lg{--tw-shadow:0 10px 15px -3px var(--tw-shadow-color,#0000001a), 0 4px 6px -4px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-md{--tw-shadow:0 4px 6px -1px var(--tw-shadow-color,#0000001a), 0 2px 4px -2px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-none{--tw-shadow:0 0 #0000;box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-sm{--tw-shadow:0 1px 3px 0 var(--tw-shadow-color,#0000001a), 0 1px 2px -1px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.shadow-xl{--tw-shadow:0 20px 25px -5px var(--tw-shadow-color,#0000001a), 0 8px 10px -6px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.ring,.ring-1{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.ring-2{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.inset-shadow-sm{--tw-inset-shadow:inset 0 2px 4px var(--tw-inset-shadow-color,#0000000d);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.\\[box-shadow\\:var\\(--rg-shadow\\)\\]{box-shadow:var(--rg-shadow)}.shadow-black\\/5{--tw-shadow-color:#0000000d}@supports (color:color-mix(in lab, red, red)){.shadow-black\\/5{--tw-shadow-color:color-mix(in oklab, color-mix(in oklab, var(--color-black) 5%, transparent) var(--tw-shadow-alpha), transparent)}}.shadow-black\\/10{--tw-shadow-color:#0000001a}@supports (color:color-mix(in lab, red, red)){.shadow-black\\/10{--tw-shadow-color:color-mix(in oklab, color-mix(in oklab, var(--color-black) 10%, transparent) var(--tw-shadow-alpha), transparent)}}.shadow-border{--tw-shadow-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.shadow-border{--tw-shadow-color:color-mix(in oklab, var(--border) var(--tw-shadow-alpha), transparent)}}.ring-amber-200\\/80{--tw-ring-color:#fee685cc}@supports (color:color-mix(in lab, red, red)){.ring-amber-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-amber-200) 80%, transparent)}}.ring-amber-300\\/60{--tw-ring-color:#ffd23699}@supports (color:color-mix(in lab, red, red)){.ring-amber-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-amber-300) 60%, transparent)}}.ring-black\\/10{--tw-ring-color:#0000001a}@supports (color:color-mix(in lab, red, red)){.ring-black\\/10{--tw-ring-color:color-mix(in oklab, var(--color-black) 10%, transparent)}}.ring-border,.ring-border\\/60{--tw-ring-color:var(--border)}@supports (color:color-mix(in lab, red, red)){.ring-border\\/60{--tw-ring-color:color-mix(in oklab, var(--border) 60%, transparent)}}.ring-emerald-200\\/80{--tw-ring-color:#a4f4cfcc}@supports (color:color-mix(in lab, red, red)){.ring-emerald-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-emerald-200) 80%, transparent)}}.ring-emerald-300\\/60{--tw-ring-color:#5ee9b599}@supports (color:color-mix(in lab, red, red)){.ring-emerald-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-emerald-300) 60%, transparent)}}.ring-foreground{--tw-ring-color:var(--foreground)}.ring-indigo-300\\/60{--tw-ring-color:#a4b3ff99}@supports (color:color-mix(in lab, red, red)){.ring-indigo-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-indigo-300) 60%, transparent)}}.ring-primary,.ring-primary\\/20{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.ring-primary\\/20{--tw-ring-color:color-mix(in oklab, var(--primary) 20%, transparent)}}.ring-primary\\/25{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.ring-primary\\/25{--tw-ring-color:color-mix(in oklab, var(--primary) 25%, transparent)}}.ring-primary\\/40{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.ring-primary\\/40{--tw-ring-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.ring-ring{--tw-ring-color:var(--ring)}.ring-rose-200\\/80{--tw-ring-color:#ffccd3cc}@supports (color:color-mix(in lab, red, red)){.ring-rose-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-rose-200) 80%, transparent)}}.ring-rose-300\\/60{--tw-ring-color:#ffa2ae99}@supports (color:color-mix(in lab, red, red)){.ring-rose-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-rose-300) 60%, transparent)}}.ring-sky-200\\/80{--tw-ring-color:#b8e6fecc}@supports (color:color-mix(in lab, red, red)){.ring-sky-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-sky-200) 80%, transparent)}}.ring-sky-300\\/60{--tw-ring-color:#77d4ff99}@supports (color:color-mix(in lab, red, red)){.ring-sky-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-sky-300) 60%, transparent)}}.ring-slate-200\\/80{--tw-ring-color:#e2e8f0cc}@supports (color:color-mix(in lab, red, red)){.ring-slate-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-slate-200) 80%, transparent)}}.ring-slate-300\\/60{--tw-ring-color:#cad5e299}@supports (color:color-mix(in lab, red, red)){.ring-slate-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-slate-300) 60%, transparent)}}.ring-teal-300\\/60{--tw-ring-color:#46ecd599}@supports (color:color-mix(in lab, red, red)){.ring-teal-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-teal-300) 60%, transparent)}}.ring-violet-200\\/80{--tw-ring-color:#ddd6ffcc}@supports (color:color-mix(in lab, red, red)){.ring-violet-200\\/80{--tw-ring-color:color-mix(in oklab, var(--color-violet-200) 80%, transparent)}}.ring-violet-300\\/60{--tw-ring-color:#c4b4ff99}@supports (color:color-mix(in lab, red, red)){.ring-violet-300\\/60{--tw-ring-color:color-mix(in oklab, var(--color-violet-300) 60%, transparent)}}.ring-offset-1{--tw-ring-offset-width:1px;--tw-ring-offset-shadow:var(--tw-ring-inset,) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)}.ring-offset-2{--tw-ring-offset-width:2px;--tw-ring-offset-shadow:var(--tw-ring-inset,) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)}.ring-offset-background{--tw-ring-offset-color:var(--background)}.outline,.outline-1{outline-style:var(--tw-outline-style);outline-width:1px}.-outline-offset-1{outline-offset:calc(1px * -1)}.outline-offset-2{outline-offset:2px}.outline-muted-foreground\\/60{outline-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.outline-muted-foreground\\/60{outline-color:color-mix(in oklab, var(--muted-foreground) 60%, transparent)}}.outline-primary\\/60{outline-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.outline-primary\\/60{outline-color:color-mix(in oklab, var(--primary) 60%, transparent)}}.outline-rose-500{outline-color:var(--color-rose-500)}.blur{--tw-blur:blur(8px);filter:var(--tw-blur,) var(--tw-brightness,) var(--tw-contrast,) var(--tw-grayscale,) var(--tw-hue-rotate,) var(--tw-invert,) var(--tw-saturate,) var(--tw-sepia,) var(--tw-drop-shadow,)}.invert{--tw-invert:invert(100%);filter:var(--tw-blur,) var(--tw-brightness,) var(--tw-contrast,) var(--tw-grayscale,) var(--tw-hue-rotate,) var(--tw-invert,) var(--tw-saturate,) var(--tw-sepia,) var(--tw-drop-shadow,)}.\\[filter\\:var\\(--rg-drop-shadow\\)\\]{filter:var(--rg-drop-shadow)}.filter{filter:var(--tw-blur,) var(--tw-brightness,) var(--tw-contrast,) var(--tw-grayscale,) var(--tw-hue-rotate,) var(--tw-invert,) var(--tw-saturate,) var(--tw-sepia,) var(--tw-drop-shadow,)}.backdrop-blur{--tw-backdrop-blur:blur(8px);-webkit-backdrop-filter:var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-opacity,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,);backdrop-filter:var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-opacity,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,)}.backdrop-blur-sm{--tw-backdrop-blur:blur(var(--blur-sm));-webkit-backdrop-filter:var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-opacity,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,);backdrop-filter:var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-opacity,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,)}.backdrop-filter{-webkit-backdrop-filter:var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-opacity,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,);backdrop-filter:var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-opacity,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,)}.transition{transition-property:color,background-color,border-color,outline-color,text-decoration-color,fill,stroke,--tw-gradient-from,--tw-gradient-via,--tw-gradient-to,opacity,box-shadow,transform,translate,scale,rotate,filter,-webkit-backdrop-filter,backdrop-filter,display,content-visibility,overlay,pointer-events;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[border-color\\,box-shadow\\]{transition-property:border-color,box-shadow;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[clip-path\\]{transition-property:clip-path;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[grid-template-columns\\]{transition-property:grid-template-columns;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[grid-template-rows\\]{transition-property:grid-template-rows;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[height\\]{transition-property:height;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[opacity\\,transform\\]{transition-property:opacity,transform;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[padding\\,border-radius\\,transform\\]{transition-property:padding,border-radius,transform;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[stroke-dashoffset\\]{transition-property:stroke-dashoffset;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[top\\,left\\,width\\,height\\,opacity\\,border-radius\\]{transition-property:top,left,width,height,opacity,border-radius;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[top\\,left\\,width\\,transform\\,border-radius\\,box-shadow\\]{transition-property:top,left,width,transform,border-radius,box-shadow;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[top\\,transform\\]{transition-property:top,transform;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[transform\\,color\\]{transition-property:transform,color;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[transform\\,opacity\\]{transition-property:transform,opacity;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[width\\]{transition-property:width;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-all{transition-property:all;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-colors{transition-property:color,background-color,border-color,outline-color,text-decoration-color,fill,stroke,--tw-gradient-from,--tw-gradient-via,--tw-gradient-to;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-opacity{transition-property:opacity;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-shadow{transition-property:box-shadow;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-transform{transition-property:transform,translate,scale,rotate;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.delay-\\[80ms\\]{transition-delay:80ms}.duration-60{--tw-duration:60ms;transition-duration:60ms}.duration-75{--tw-duration:75ms;transition-duration:75ms}.duration-100{--tw-duration:.1s;transition-duration:.1s}.duration-120{--tw-duration:.12s;transition-duration:.12s}.duration-140{--tw-duration:.14s;transition-duration:.14s}.duration-150{--tw-duration:.15s;transition-duration:.15s}.duration-180{--tw-duration:.18s;transition-duration:.18s}.duration-200{--tw-duration:.2s;transition-duration:.2s}.duration-220{--tw-duration:.22s;transition-duration:.22s}.duration-300{--tw-duration:.3s;transition-duration:.3s}.duration-400{--tw-duration:.4s;transition-duration:.4s}.duration-500{--tw-duration:.5s;transition-duration:.5s}.ease-\\[cubic-bezier\\(0\\,0\\,0\\.2\\,1\\)\\]{--tw-ease:cubic-bezier(0,0,.2,1);transition-timing-function:cubic-bezier(0,0,.2,1)}.ease-out{--tw-ease:var(--ease-out);transition-timing-function:var(--ease-out)}.will-change-\\[opacity\\,transform\\]{will-change:opacity,transform}.will-change-transform{will-change:transform}.contain-layout{--tw-contain-layout:layout;contain:var(--tw-contain-size,) var(--tw-contain-layout,) var(--tw-contain-paint,) var(--tw-contain-style,)}.\\[contain\\:inline-size\\]{contain:inline-size}.outline-dashed{--tw-outline-style:dashed;outline-style:dashed}.outline-none{--tw-outline-style:none;outline-style:none}.select-all{-webkit-user-select:all;user-select:all}.select-none{-webkit-user-select:none;user-select:none}.\\[--json-view-indent\\:0\\.75rem\\]{--json-view-indent:.75rem}.\\[--json-view-indent\\:1\\.25rem\\]{--json-view-indent:1.25rem}.\\[--json-view-indent\\:1rem\\]{--json-view-indent:1rem}.\\[--json-view-row-space\\:0px\\]{--json-view-row-space:0px}.\\[--json-view-row-space\\:1px\\]{--json-view-row-space:1px}.\\[--json-view-row-space\\:2px\\]{--json-view-row-space:2px}.\\[-webkit-box-orient\\:vertical\\]{-webkit-box-orient:vertical}.\\[-webkit-line-clamp\\:3\\]{-webkit-line-clamp:3}.\\[-webkit-user-drag\\:none\\]{-webkit-user-drag:none}.\\[corner-shape\\:superellipse\\(1\\.25\\)\\]{corner-shape:superellipse(1.25)}.\\[font-synthesis\\:none\\]{font-synthesis:none}.\\[profile\\:pod-cpu\\]{profile:pod-cpu}.\\[rows\\:0\\]{rows:0}.\\[rows\\:1\\]{rows:1}.\\[rows\\:25\\]{rows:25}.\\[text-shadow\\:0_0_2px_rgb\\(0_0_0\\/0\\.45\\)\\]{text-shadow:0 0 2px #00000073}.\\[writing-mode\\:vertical-rl\\]{writing-mode:vertical-rl}.ring-inset{--tw-ring-inset:inset}.group-open\\:hidden:is(:where(.group):is([open],:popover-open,:open) *){display:none}.group-open\\:inline:is(:where(.group):is([open],:popover-open,:open) *){display:inline}.group-open\\:rotate-90:is(:where(.group):is([open],:popover-open,:open) *){rotate:90deg}@media (hover:hover){.group-hover\\:text-\\[var\\(--rg-text-primary\\)\\]:is(:where(.group):hover *){color:var(--rg-text-primary)}.group-hover\\:text-foreground:is(:where(.group):hover *){color:var(--foreground)}.group-hover\\:opacity-100:is(:where(.group):hover *){opacity:1}.group-hover\\/add\\:border-primary:is(:where(.group\\/add):hover *){border-color:var(--primary)}.group-hover\\/add\\:text-primary:is(:where(.group\\/add):hover *){color:var(--primary)}.group-hover\\/cell-actions\\:opacity-95:is(:where(.group\\/cell-actions):hover *){opacity:.95}.group-hover\\/comment\\:opacity-70:is(:where(.group\\/comment):hover *){opacity:.7}.group-hover\\/header\\:bg-primary\\/70:is(:where(.group\\/header):hover *){background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.group-hover\\/header\\:bg-primary\\/70:is(:where(.group\\/header):hover *){background-color:color-mix(in oklab, var(--primary) 70%, transparent)}}.group-hover\\/thread\\:opacity-100:is(:where(.group\\/thread):hover *){opacity:1}}.placeholder\\:text-placeholder::placeholder{color:var(--placeholder)}.before\\:invisible:before{content:var(--tw-content);visibility:hidden}.before\\:absolute:before{content:var(--tw-content);position:absolute}.before\\:inset-y-0:before{content:var(--tw-content);inset-block:0}.before\\:left-0:before{content:var(--tw-content);left:0}.before\\:w-\\[3px\\]:before{content:var(--tw-content);width:3px}.before\\:bg-emerald-500:before{content:var(--tw-content);background-color:var(--color-emerald-500)}.before\\:content-\\[\\'\\.\\'\\]:before{--tw-content:\".\";content:var(--tw-content)}.after\\:absolute:after{content:var(--tw-content);position:absolute}.after\\:inset-0:after{content:var(--tw-content);inset:0}.after\\:content-\\[\\'\\'\\]:after{--tw-content:\"\";content:var(--tw-content)}.first\\:mt-1:first-child{margin-top:var(--spacing)}.first\\:mt-2:first-child{margin-top:calc(var(--spacing) * 2)}.first\\:border-t-0:first-child{border-top-style:var(--tw-border-style);border-top-width:0}.first\\:border-l-0:first-child{border-left-style:var(--tw-border-style);border-left-width:0}.first\\:pt-0:first-child{padding-top:0}.last\\:border-0:last-child{border-style:var(--tw-border-style);border-width:0}.last\\:border-b-0:last-child{border-bottom-style:var(--tw-border-style);border-bottom-width:0}.last\\:pb-0:last-child{padding-bottom:0}.even\\:bg-muted\\/40:nth-child(2n){background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.even\\:bg-muted\\/40:nth-child(2n){background-color:color-mix(in oklab, var(--muted) 40%, transparent)}}.open\\:w-full:is([open],:popover-open,:open){width:100%}.open\\:rounded-lg:is([open],:popover-open,:open){border-radius:var(--radius-lg)}.empty\\:hidden:empty{display:none}.focus-within\\:border-primary\\/40:focus-within{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus-within\\:border-primary\\/40:focus-within{border-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.focus-within\\:opacity-100:focus-within{opacity:1}.focus-within\\:shadow-md:focus-within{--tw-shadow:0 4px 6px -1px var(--tw-shadow-color,#0000001a), 0 2px 4px -2px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus-within\\:ring-2:focus-within{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus-within\\:ring-destructive:focus-within{--tw-ring-color:var(--destructive)}.focus-within\\:ring-primary\\/30:focus-within{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus-within\\:ring-primary\\/30:focus-within{--tw-ring-color:color-mix(in oklab, var(--primary) 30%, transparent)}}.focus-within\\:ring-ring:focus-within{--tw-ring-color:var(--ring)}.focus-within\\:ring-offset-1:focus-within{--tw-ring-offset-width:1px;--tw-ring-offset-shadow:var(--tw-ring-inset,) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)}@media (hover:hover){.hover\\:z-\\[999\\]:hover{z-index:999}.hover\\:border-\\[\\#abc\\]:hover{border-color:#abc}.hover\\:border-border:hover{border-color:var(--border)}.hover\\:border-emerald-200:hover{border-color:var(--color-emerald-200)}.hover\\:border-foreground\\/20:hover{border-color:var(--foreground)}@supports (color:color-mix(in lab, red, red)){.hover\\:border-foreground\\/20:hover{border-color:color-mix(in oklab, var(--foreground) 20%, transparent)}}.hover\\:border-input:hover{border-color:var(--input)}.hover\\:border-muted-foreground\\/40:hover{border-color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.hover\\:border-muted-foreground\\/40:hover{border-color:color-mix(in oklab, var(--muted-foreground) 40%, transparent)}}.hover\\:border-primary:hover,.hover\\:border-primary\\/30:hover{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:border-primary\\/30:hover{border-color:color-mix(in oklab, var(--primary) 30%, transparent)}}.hover\\:border-primary\\/40:hover{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:border-primary\\/40:hover{border-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.hover\\:border-primary\\/50:hover{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:border-primary\\/50:hover{border-color:color-mix(in oklab, var(--primary) 50%, transparent)}}.hover\\:bg-\\[var\\(--rg-error-bg-hover\\)\\]:hover{background-color:var(--rg-error-bg-hover)}.hover\\:bg-\\[var\\(--rg-surface-active\\)\\]:hover{background-color:var(--rg-surface-active)}.hover\\:bg-\\[var\\(--rg-surface-hover\\)\\]:hover{background-color:var(--rg-surface-hover)}.hover\\:bg-accent:hover,.hover\\:bg-accent\\/40:hover{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-accent\\/40:hover{background-color:color-mix(in oklab, var(--accent) 40%, transparent)}}.hover\\:bg-accent\\/50:hover{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-accent\\/50:hover{background-color:color-mix(in oklab, var(--accent) 50%, transparent)}}.hover\\:bg-amber-500\\/20:hover{background-color:#f99c0033}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-amber-500\\/20:hover{background-color:color-mix(in oklab, var(--color-amber-500) 20%, transparent)}}.hover\\:bg-background:hover,.hover\\:bg-background\\/60:hover{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-background\\/60:hover{background-color:color-mix(in oklab, var(--background) 60%, transparent)}}.hover\\:bg-background\\/70:hover{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-background\\/70:hover{background-color:color-mix(in oklab, var(--background) 70%, transparent)}}.hover\\:bg-border:hover{background-color:var(--border)}.hover\\:bg-card:hover{background-color:var(--card)}.hover\\:bg-destructive\\/10:hover{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-destructive\\/10:hover{background-color:color-mix(in oklab, var(--destructive) 10%, transparent)}}.hover\\:bg-destructive\\/90:hover{background-color:var(--destructive)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-destructive\\/90:hover{background-color:color-mix(in oklab, var(--destructive) 90%, transparent)}}.hover\\:bg-emerald-50\\/50:hover{background-color:#ecfdf580}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-emerald-50\\/50:hover{background-color:color-mix(in oklab, var(--color-emerald-50) 50%, transparent)}}.hover\\:bg-muted:hover,.hover\\:bg-muted\\/30:hover{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-muted\\/30:hover{background-color:color-mix(in oklab, var(--muted) 30%, transparent)}}.hover\\:bg-muted\\/40:hover{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-muted\\/40:hover{background-color:color-mix(in oklab, var(--muted) 40%, transparent)}}.hover\\:bg-muted\\/50:hover{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-muted\\/50:hover{background-color:color-mix(in oklab, var(--muted) 50%, transparent)}}.hover\\:bg-muted\\/60:hover{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-muted\\/60:hover{background-color:color-mix(in oklab, var(--muted) 60%, transparent)}}.hover\\:bg-muted\\/70:hover{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-muted\\/70:hover{background-color:color-mix(in oklab, var(--muted) 70%, transparent)}}.hover\\:bg-primary:hover,.hover\\:bg-primary\\/10:hover{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-primary\\/10:hover{background-color:color-mix(in oklab, var(--primary) 10%, transparent)}}.hover\\:bg-primary\\/15:hover{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-primary\\/15:hover{background-color:color-mix(in oklab, var(--primary) 15%, transparent)}}.hover\\:bg-primary\\/20:hover{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-primary\\/20:hover{background-color:color-mix(in oklab, var(--primary) 20%, transparent)}}.hover\\:bg-primary\\/40:hover{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-primary\\/40:hover{background-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.hover\\:bg-primary\\/90:hover{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-primary\\/90:hover{background-color:color-mix(in oklab, var(--primary) 90%, transparent)}}.hover\\:bg-secondary\\/80:hover{background-color:var(--secondary)}@supports (color:color-mix(in lab, red, red)){.hover\\:bg-secondary\\/80:hover{background-color:color-mix(in oklab, var(--secondary) 80%, transparent)}}.hover\\:bg-sidebar-accent:hover{background-color:var(--sidebar-accent)}.hover\\:from-primary\\/20:hover{--tw-gradient-from:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:from-primary\\/20:hover{--tw-gradient-from:color-mix(in oklab, var(--primary) 20%, transparent)}}.hover\\:from-primary\\/20:hover{--tw-gradient-stops:var(--tw-gradient-via-stops,var(--tw-gradient-position), var(--tw-gradient-from) var(--tw-gradient-from-position), var(--tw-gradient-to) var(--tw-gradient-to-position))}.hover\\:fill-primary\\/40:hover{fill:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:fill-primary\\/40:hover{fill:color-mix(in oklab, var(--primary) 40%, transparent)}}.hover\\:stroke-primary\\/15:hover{stroke:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:stroke-primary\\/15:hover{stroke:color-mix(in oklab, var(--primary) 15%, transparent)}}.hover\\:text-\\[var\\(--clicky-success-700\\)\\]:hover{color:var(--clicky-success-700)}.hover\\:text-\\[var\\(--rg-text-primary\\)\\]:hover{color:var(--rg-text-primary)}.hover\\:text-accent-foreground:hover{color:var(--accent-foreground)}.hover\\:text-destructive:hover{color:var(--destructive)}.hover\\:text-foreground:hover{color:var(--foreground)}.hover\\:text-gray-200:hover{color:var(--color-gray-200)}.hover\\:text-primary:hover,.hover\\:text-primary\\/80:hover{color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:text-primary\\/80:hover{color:color-mix(in oklab, var(--primary) 80%, transparent)}}.hover\\:text-sidebar-accent-foreground:hover{color:var(--sidebar-accent-foreground)}.hover\\:underline:hover{text-decoration-line:underline}.hover\\:opacity-80:hover{opacity:.8}.hover\\:opacity-100:hover{opacity:1}.hover\\:shadow-md:hover{--tw-shadow:0 4px 6px -1px var(--tw-shadow-color,#0000001a), 0 2px 4px -2px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.hover\\:shadow-sm:hover{--tw-shadow:0 1px 3px 0 var(--tw-shadow-color,#0000001a), 0 1px 2px -1px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.hover\\:ring-1:hover{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.hover\\:ring-2:hover{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.hover\\:ring-border:hover{--tw-ring-color:var(--border)}.hover\\:ring-primary\\/50:hover{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.hover\\:ring-primary\\/50:hover{--tw-ring-color:color-mix(in oklab, var(--primary) 50%, transparent)}}.hover\\:ring-ring:hover{--tw-ring-color:var(--ring)}.hover\\:brightness-95:hover{--tw-brightness:brightness(95%);filter:var(--tw-blur,) var(--tw-brightness,) var(--tw-contrast,) var(--tw-grayscale,) var(--tw-hue-rotate,) var(--tw-invert,) var(--tw-saturate,) var(--tw-sepia,) var(--tw-drop-shadow,)}}.focus\\:border-primary:focus,.focus\\:border-primary\\/40:focus{border-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus\\:border-primary\\/40:focus{border-color:color-mix(in oklab, var(--primary) 40%, transparent)}}.focus\\:bg-accent:focus{background-color:var(--accent)}.focus\\:bg-background:focus{background-color:var(--background)}.focus\\:stroke-primary:focus{stroke:var(--primary)}.focus\\:text-accent-foreground:focus{color:var(--accent-foreground)}.focus\\:text-foreground:focus{color:var(--foreground)}.focus\\:ring-1:focus{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus\\:ring-2:focus{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus\\:ring-primary\\/15:focus{--tw-ring-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus\\:ring-primary\\/15:focus{--tw-ring-color:color-mix(in oklab, var(--primary) 15%, transparent)}}.focus\\:ring-ring:focus{--tw-ring-color:var(--ring)}.focus\\:outline-none:focus{--tw-outline-style:none;outline-style:none}.focus-visible\\:relative:focus-visible{position:relative}.focus-visible\\:z-10:focus-visible{z-index:10}.focus-visible\\:bg-accent\\/50:focus-visible{background-color:var(--accent)}@supports (color:color-mix(in lab, red, red)){.focus-visible\\:bg-accent\\/50:focus-visible{background-color:color-mix(in oklab, var(--accent) 50%, transparent)}}.focus-visible\\:bg-muted:focus-visible{background-color:var(--muted)}.focus-visible\\:bg-primary:focus-visible,.focus-visible\\:bg-primary\\/10:focus-visible{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus-visible\\:bg-primary\\/10:focus-visible{background-color:color-mix(in oklab, var(--primary) 10%, transparent)}}.focus-visible\\:bg-primary\\/60:focus-visible{background-color:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus-visible\\:bg-primary\\/60:focus-visible{background-color:color-mix(in oklab, var(--primary) 60%, transparent)}}.focus-visible\\:stroke-primary\\/30:focus-visible{stroke:var(--primary)}@supports (color:color-mix(in lab, red, red)){.focus-visible\\:stroke-primary\\/30:focus-visible{stroke:color-mix(in oklab, var(--primary) 30%, transparent)}}.focus-visible\\:text-primary:focus-visible{color:var(--primary)}.focus-visible\\:opacity-100:focus-visible{opacity:1}.focus-visible\\:ring-0:focus-visible{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(0px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus-visible\\:ring-1:focus-visible{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus-visible\\:ring-2:focus-visible{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.focus-visible\\:ring-destructive:focus-visible{--tw-ring-color:var(--destructive)}.focus-visible\\:ring-primary:focus-visible{--tw-ring-color:var(--primary)}.focus-visible\\:ring-ring:focus-visible{--tw-ring-color:var(--ring)}.focus-visible\\:ring-offset-0:focus-visible{--tw-ring-offset-width:0px;--tw-ring-offset-shadow:var(--tw-ring-inset,) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)}.focus-visible\\:ring-offset-1:focus-visible{--tw-ring-offset-width:1px;--tw-ring-offset-shadow:var(--tw-ring-inset,) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)}.focus-visible\\:ring-offset-2:focus-visible{--tw-ring-offset-width:2px;--tw-ring-offset-shadow:var(--tw-ring-inset,) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color)}.focus-visible\\:outline:focus-visible{outline-style:var(--tw-outline-style);outline-width:1px}.focus-visible\\:outline-2:focus-visible{outline-style:var(--tw-outline-style);outline-width:2px}.focus-visible\\:outline-ring:focus-visible{outline-color:var(--ring)}.focus-visible\\:outline-none:focus-visible{--tw-outline-style:none;outline-style:none}.focus-visible\\:ring-inset:focus-visible{--tw-ring-inset:inset}.active\\:cursor-grabbing:active{cursor:grabbing}.disabled\\:pointer-events-none:disabled{pointer-events:none}.disabled\\:cursor-default:disabled{cursor:default}.disabled\\:cursor-not-allowed:disabled{cursor:not-allowed}.disabled\\:bg-muted\\/40:disabled{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.disabled\\:bg-muted\\/40:disabled{background-color:color-mix(in oklab, var(--muted) 40%, transparent)}}.disabled\\:text-muted-foreground:disabled,.disabled\\:text-muted-foreground\\/40:disabled{color:var(--muted-foreground)}@supports (color:color-mix(in lab, red, red)){.disabled\\:text-muted-foreground\\/40:disabled{color:color-mix(in oklab, var(--muted-foreground) 40%, transparent)}}.disabled\\:opacity-30:disabled{opacity:.3}.disabled\\:opacity-35:disabled{opacity:.35}.disabled\\:opacity-40:disabled{opacity:.4}.disabled\\:opacity-50:disabled{opacity:.5}.disabled\\:opacity-60:disabled{opacity:.6}@media (hover:hover){.disabled\\:hover\\:bg-transparent:disabled:hover{background-color:#0000}.disabled\\:hover\\:text-muted-foreground:disabled:hover{color:var(--muted-foreground)}}.has-\\[\\:disabled\\]\\:cursor-not-allowed:has(:disabled){cursor:not-allowed}.has-\\[\\:disabled\\]\\:opacity-50:has(:disabled){opacity:.5}.aria-expanded\\:bg-muted[aria-expanded=true]{background-color:var(--muted)}.aria-pressed\\:bg-accent[aria-pressed=true]{background-color:var(--accent)}.aria-pressed\\:bg-muted[aria-pressed=true]{background-color:var(--muted)}.aria-pressed\\:font-semibold[aria-pressed=true]{--tw-font-weight:var(--font-weight-semibold);font-weight:var(--font-weight-semibold)}.aria-pressed\\:text-foreground[aria-pressed=true]{color:var(--foreground)}.aria-pressed\\:underline[aria-pressed=true]{text-decoration-line:underline}.data-\\[nav-drag\\=source\\]\\:opacity-40[data-nav-drag=source]{opacity:.4}.data-\\[nav-drop\\=over\\]\\:rounded-md[data-nav-drop=over]{border-radius:var(--radius-md)}.data-\\[nav-drop\\=over\\]\\:bg-sidebar-accent[data-nav-drop=over]{background-color:var(--sidebar-accent)}.data-\\[nav-drop\\=over\\]\\:ring-1[data-nav-drop=over]{--tw-ring-shadow:var(--tw-ring-inset,) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.data-\\[nav-drop\\=over\\]\\:ring-sidebar-primary[data-nav-drop=over]{--tw-ring-color:var(--sidebar-primary)}.data-\\[nav-drop\\=over\\]\\:outline-1[data-nav-drop=over]{outline-style:var(--tw-outline-style);outline-width:1px}.data-\\[nav-drop\\=over\\]\\:outline-sidebar-primary[data-nav-drop=over]{outline-color:var(--sidebar-primary)}.data-\\[nav-drop\\=over\\]\\:outline-dashed[data-nav-drop=over]{--tw-outline-style:dashed;outline-style:dashed}.data-\\[nav-drop\\=over\\]\\:ring-inset[data-nav-drop=over]{--tw-ring-inset:inset}.data-\\[theme\\=dark\\]\\:bg-\\[var\\(--clicky-bg-subtle\\)\\][data-theme=dark]{background-color:var(--clicky-bg-subtle)}@supports ((-webkit-backdrop-filter:var(--tw)) or (backdrop-filter:var(--tw))){.supports-\\[backdrop-filter\\]\\:bg-background\\/80{background-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.supports-\\[backdrop-filter\\]\\:bg-background\\/80{background-color:color-mix(in oklab, var(--background) 80%, transparent)}}.supports-\\[backdrop-filter\\]\\:bg-sidebar{background-color:var(--sidebar)}}@media (prefers-reduced-motion:no-preference){.motion-safe\\:animate-spin{animation:var(--animate-spin)}}@media (prefers-reduced-motion:reduce){.motion-reduce\\:hidden{display:none}.motion-reduce\\:transition-none{transition-property:none}}@media not all and (width>=48rem){.max-md\\:hidden{display:none}.max-md\\:inline-flex{display:inline-flex}.max-md\\:h-8{height:calc(var(--spacing) * 8)}.max-md\\:h-9{height:calc(var(--spacing) * 9)}.max-md\\:flex-nowrap{flex-wrap:nowrap}}@media not all and (width>=40rem){.max-sm\\:sr-only{clip-path:inset(50%);white-space:nowrap;border-width:0;width:1px;height:1px;margin:-1px;padding:0;position:absolute;overflow:hidden}.max-sm\\:ml-0{margin-left:0}.max-sm\\:ml-6{margin-left:calc(var(--spacing) * 6)}.max-sm\\:hidden{display:none}.max-sm\\:\\!h-dvh{height:100dvh!important}.max-sm\\:h-9{height:calc(var(--spacing) * 9)}.max-sm\\:\\!max-h-dvh{max-height:100dvh!important}.max-sm\\:min-h-14{min-height:calc(var(--spacing) * 14)}.max-sm\\:w-9{width:calc(var(--spacing) * 9)}.max-sm\\:\\!max-w-none{max-width:none!important}.max-sm\\:basis-full{flex-basis:100%}.max-sm\\:flex-nowrap{flex-wrap:nowrap}.max-sm\\:items-stretch{align-items:stretch}.max-sm\\:justify-end{justify-content:flex-end}.max-sm\\:\\!rounded-none{border-radius:0!important}.max-sm\\:rounded-none{border-radius:0}.max-sm\\:\\!border-0{border-style:var(--tw-border-style)!important;border-width:0!important}.max-sm\\:border-0{border-style:var(--tw-border-style);border-width:0}.max-sm\\:\\!bg-transparent{background-color:#0000!important}.max-sm\\:bg-background{background-color:var(--background)}.max-sm\\:bg-transparent{background-color:#0000}.max-sm\\:p-0{padding:0}.max-sm\\:px-0{padding-inline:0}.max-sm\\:py-density-4{padding-block:var(--spacing-density-4)}.max-sm\\:text-base{font-size:var(--text-base);line-height:var(--tw-leading,var(--text-base--line-height))}.max-sm\\:leading-6{--tw-leading:calc(var(--spacing) * 6);line-height:calc(var(--spacing) * 6)}.max-sm\\:shadow-none{--tw-shadow:0 0 #0000;box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.max-sm\\:transition-transform{transition-property:transform,translate,scale,rotate;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.max-sm\\:duration-200{--tw-duration:.2s;transition-duration:.2s}.max-sm\\:ease-out{--tw-ease:var(--ease-out);transition-timing-function:var(--ease-out)}.max-sm\\:data-\\[state\\=closed\\]\\:translate-x-full[data-state=closed]{--tw-translate-x:100%;translate:var(--tw-translate-x) var(--tw-translate-y)}.max-sm\\:data-\\[state\\=open\\]\\:translate-x-0[data-state=open]{--tw-translate-x:0px;translate:var(--tw-translate-x) var(--tw-translate-y)}@media (prefers-reduced-motion:reduce){.max-sm\\:motion-reduce\\:transition-none{transition-property:none}}}@media (width>=40rem){.sm\\:left-auto{left:auto}.sm\\:order-none{order:0}.sm\\:col-span-1{grid-column:span 1/span 1}.sm\\:col-span-2{grid-column:span 2/span 2}.sm\\:block{display:block}.sm\\:hidden{display:none}.sm\\:h-7{height:calc(var(--spacing) * 7)}.sm\\:w-7{width:calc(var(--spacing) * 7)}.sm\\:w-48{width:calc(var(--spacing) * 48)}.sm\\:w-80{width:calc(var(--spacing) * 80)}.sm\\:max-w-xs{max-width:var(--container-xs)}.sm\\:basis-auto{flex-basis:auto}.sm\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}.sm\\:grid-cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}.sm\\:grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}.sm\\:grid-cols-6{grid-template-columns:repeat(6,minmax(0,1fr))}.sm\\:grid-cols-\\[8rem_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:8rem minmax(0,1fr) auto}.sm\\:grid-cols-\\[9rem_4\\.5rem_minmax\\(0\\,1fr\\)_2\\.25rem\\]{grid-template-columns:9rem 4.5rem minmax(0,1fr) 2.25rem}.sm\\:grid-cols-\\[auto_1fr\\]{grid-template-columns:auto 1fr}.sm\\:grid-cols-\\[minmax\\(0\\,1fr\\)_10rem\\]{grid-template-columns:minmax(0,1fr) 10rem}.sm\\:grid-cols-\\[minmax\\(0\\,1fr\\)_10rem_12rem\\]{grid-template-columns:minmax(0,1fr) 10rem 12rem}.sm\\:grid-cols-\\[minmax\\(0\\,1fr\\)_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:minmax(0,1fr) minmax(0,1fr) auto}.sm\\:grid-cols-\\[minmax\\(8rem\\,14rem\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(8rem,14rem) minmax(0,1fr)}.sm\\:grid-cols-\\[minmax\\(10rem\\,0\\.35fr\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(10rem,.35fr) minmax(0,1fr)}.sm\\:grid-cols-\\[minmax\\(10rem\\,16rem\\)\\]{grid-template-columns:minmax(10rem,16rem)}.sm\\:flex-row{flex-direction:row}.sm\\:flex-nowrap{flex-wrap:nowrap}.sm\\:flex-wrap{flex-wrap:wrap}.sm\\:items-center{align-items:center}.sm\\:items-start{align-items:flex-start}.sm\\:justify-between{justify-content:space-between}.sm\\:justify-end{justify-content:flex-end}.sm\\:gap-3{gap:calc(var(--spacing) * 3)}.sm\\:p-density-4{padding:var(--spacing-density-4)}}@media (width>=48rem){.md\\:z-\\[var\\(--x\\)\\]{z-index:var(--x)}.md\\:order-none{order:0}.md\\:\\[grid-column\\:span_var\\(--stat-last-span\\)\\/span_var\\(--stat-last-span\\)\\]{grid-column:span var(--stat-last-span)/span var(--stat-last-span)}.md\\:col-span-1{grid-column:span 1/span 1}.md\\:col-span-2{grid-column:span 2/span 2}.md\\:block{display:block}.md\\:flex{display:flex}.md\\:h-12{height:calc(var(--spacing) * 12)}.md\\:h-\\[calc\\(100dvh-14rem\\)\\]{height:calc(100dvh - 14rem)}.md\\:min-h-\\[30rem\\]{min-height:30rem}.md\\:w-24{width:calc(var(--spacing) * 24)}.md\\:w-40{width:calc(var(--spacing) * 40)}.md\\:w-64{width:calc(var(--spacing) * 64)}.md\\:w-auto{width:auto}.md\\:max-w-\\[24rem\\]{max-width:24rem}.md\\:flex-1{flex:1}.md\\:flex-\\[1_1_14rem\\]{flex:14rem}.md\\:flex-none{flex:none}.md\\:shrink-0{flex-shrink:0}.md\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}.md\\:grid-cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}.md\\:grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}.md\\:grid-cols-\\[1fr_auto_1fr\\]{grid-template-columns:1fr auto 1fr}.md\\:grid-cols-\\[minmax\\(0\\,1fr\\)_9rem_6rem\\]{grid-template-columns:minmax(0,1fr) 9rem 6rem}.md\\:grid-cols-\\[minmax\\(7rem\\,10rem\\)_auto_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:minmax(7rem,10rem) auto minmax(0,1fr) auto}.md\\:grid-cols-\\[minmax\\(8rem\\,12rem\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(8rem,12rem) minmax(0,1fr)}.md\\:grid-cols-\\[repeat\\(var\\(--stat-cols\\,4\\)\\,minmax\\(0\\,1fr\\)\\)\\]{grid-template-columns:repeat(var(--stat-cols,4),minmax(0,1fr))}.md\\:flex-row{flex-direction:row}.md\\:flex-nowrap{flex-wrap:nowrap}.md\\:flex-wrap{flex-wrap:wrap}.md\\:items-center{align-items:center}.md\\:p-2{padding:calc(var(--spacing) * 2)}}@media (width>=64rem){.lg\\:col-span-2{grid-column:span 2/span 2}.lg\\:block{display:block}.lg\\:flex{display:flex}.lg\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}.lg\\:grid-cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}.lg\\:grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}.lg\\:grid-cols-\\[1\\.2fr_0\\.8fr\\]{grid-template-columns:1.2fr .8fr}.lg\\:grid-cols-\\[1fr_18rem\\]{grid-template-columns:1fr 18rem}.lg\\:grid-cols-\\[minmax\\(0\\,1\\.4fr\\)_minmax\\(0\\,0\\.7fr\\)\\]{grid-template-columns:minmax(0,1.4fr) minmax(0,.7fr)}.lg\\:grid-cols-\\[minmax\\(0\\,1fr\\)_12rem\\]{grid-template-columns:minmax(0,1fr) 12rem}.lg\\:grid-cols-\\[minmax\\(0\\,1fr\\)_minmax\\(0\\,1\\.15fr\\)\\]{grid-template-columns:minmax(0,1fr) minmax(0,1.15fr)}.lg\\:grid-cols-\\[minmax\\(0\\,1fr\\)_minmax\\(0\\,1fr\\)\\]{grid-template-columns:minmax(0,1fr) minmax(0,1fr)}.lg\\:grid-cols-\\[minmax\\(12rem\\,1\\.1fr\\)_minmax\\(12rem\\,1fr\\)_minmax\\(14rem\\,1fr\\)_auto\\]{grid-template-columns:minmax(12rem,1.1fr) minmax(12rem,1fr) minmax(14rem,1fr) auto}.lg\\:grid-cols-\\[minmax\\(22rem\\,0\\.9fr\\)_minmax\\(24rem\\,1\\.1fr\\)\\]{grid-template-columns:minmax(22rem,.9fr) minmax(24rem,1.1fr)}.lg\\:items-center{align-items:center}}@media (width>=80rem){.xl\\:sticky{position:sticky}.xl\\:top-0{top:0}.xl\\:inline{display:inline}.xl\\:max-h-screen{max-height:100vh}.xl\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}.xl\\:grid-cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}.xl\\:grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}.xl\\:grid-cols-6{grid-template-columns:repeat(6,minmax(0,1fr))}.xl\\:grid-cols-\\[14rem_minmax\\(30rem\\,0\\.9fr\\)_minmax\\(36rem\\,1\\.1fr\\)\\]{grid-template-columns:14rem minmax(30rem,.9fr) minmax(36rem,1.1fr)}.xl\\:grid-cols-\\[minmax\\(0\\,1\\.2fr\\)_minmax\\(22rem\\,0\\.8fr\\)\\]{grid-template-columns:minmax(0,1.2fr) minmax(22rem,.8fr)}.xl\\:grid-cols-\\[minmax\\(0\\,1fr\\)_minmax\\(20rem\\,26rem\\)\\]{grid-template-columns:minmax(0,1fr) minmax(20rem,26rem)}.xl\\:grid-cols-\\[minmax\\(12rem\\,1fr\\)_auto\\]{grid-template-columns:minmax(12rem,1fr) auto}.xl\\:items-center{align-items:center}.xl\\:self-start{align-self:flex-start}.xl\\:overflow-auto{overflow:auto}}@container not (width>=60rem){.\\@max-\\[60rem\\]\\:hidden{display:none}}@container not (width>=48rem){.\\@max-\\[48rem\\]\\:hidden{display:none}.\\@max-\\[48rem\\]\\:inline-flex{display:inline-flex}}@container app-shell not (width>=48rem){.\\@max-3xl\\/app-shell\\:max-w-full{max-width:100%}.\\@max-3xl\\/app-shell\\:min-w-0{min-width:0}.\\@max-3xl\\/app-shell\\:shrink{flex-shrink:1}.\\@max-3xl\\/app-shell\\:overflow-x-auto{overflow-x:auto}}@container not (width>=28rem){.\\@max-md\\:\\!block{display:block!important}.\\@max-md\\:hidden{display:none}}@container secret (width>=22rem){.\\@min-\\[22rem\\]\\/secret\\:basis-40{flex-basis:calc(var(--spacing) * 40)}.\\@min-\\[22rem\\]\\/secret\\:basis-72{flex-basis:calc(var(--spacing) * 72)}.\\@min-\\[22rem\\]\\/secret\\:flex-row{flex-direction:row}}@container error-details (width>=24rem){.\\@sm\\/error-details\\:w-auto{width:auto}}@container combobox-options (width>=28rem){.\\@min-\\[28rem\\]\\/combobox-options\\:ml-auto{margin-left:auto}.\\@min-\\[28rem\\]\\/combobox-options\\:max-w-\\[55\\%\\]{max-width:55%}.\\@min-\\[28rem\\]\\/combobox-options\\:flex-1{flex:1}.\\@min-\\[28rem\\]\\/combobox-options\\:shrink-0{flex-shrink:0}}@container (width>=28rem){.\\@md\\:grid-cols-4{grid-template-columns:repeat(4,minmax(0,1fr))}}@container combobox-options (width>=28rem){.\\@min-\\[28rem\\]\\/combobox-options\\:flex-row{flex-direction:row}.\\@min-\\[28rem\\]\\/combobox-options\\:items-center{align-items:center}.\\@min-\\[28rem\\]\\/combobox-options\\:gap-2{gap:calc(var(--spacing) * 2)}.\\@min-\\[28rem\\]\\/combobox-options\\:truncate{text-overflow:ellipsis;white-space:nowrap;overflow:hidden}.\\@min-\\[28rem\\]\\/combobox-options\\:pl-2{padding-left:calc(var(--spacing) * 2)}.\\@min-\\[28rem\\]\\/combobox-options\\:text-right{text-align:right}}@container (width>=36rem){.\\@xl\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}}@container endpoint (width>=40rem){.\\@min-\\[40rem\\]\\/endpoint\\:col-span-1{grid-column:span 1/span 1}.\\@min-\\[40rem\\]\\/endpoint\\:col-span-3{grid-column:span 3/span 3}.\\@min-\\[40rem\\]\\/endpoint\\:grid-cols-\\[minmax\\(14rem\\,1fr\\)_auto_auto\\]{grid-template-columns:minmax(14rem,1fr) auto auto}}@container (width>=42rem){.\\@2xl\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:order-none{order:0}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:order-none{order:0}.\\@3xl\\/workspace\\:col-start-1{grid-column-start:1}.\\@3xl\\/workspace\\:col-start-2{grid-column-start:2}.\\@3xl\\/workspace\\:col-start-3{grid-column-start:3}.\\@3xl\\/workspace\\:row-span-2{grid-row:span 2/span 2}.\\@3xl\\/workspace\\:row-start-1{grid-row-start:1}.\\@3xl\\/workspace\\:row-start-2{grid-row-start:2}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:ml-0{margin-left:0}.\\@3xl\\/app-shell\\:block{display:block}.\\@3xl\\/app-shell\\:flex{display:flex}.\\@3xl\\/app-shell\\:hidden{display:none}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:block{display:block}.\\@3xl\\/split-pane\\:hidden{display:none}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:block{display:block}.\\@3xl\\/workspace\\:flex{display:flex}.\\@3xl\\/workspace\\:hidden{display:none}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:h-14{height:calc(var(--spacing) * 14)}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:h-\\[var\\(--workspace-bottom-height\\)\\]{height:var(--workspace-bottom-height)}.\\@3xl\\/workspace\\:h-full{height:100%}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:max-h-none{max-height:none}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:min-h-0{min-height:0}.\\@3xl\\/workspace\\:min-h-12{min-height:calc(var(--spacing) * 12)}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:w-\\[var\\(--split-pane-width\\)\\]{width:var(--split-pane-width)}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:w-\\[var\\(--workspace-pane-width\\)\\]{width:var(--workspace-pane-width)}.\\@3xl\\/workspace\\:w-\\[var\\(--workspace-region-width\\)\\]{width:var(--workspace-region-width)}}@container session-viewer (width>=48rem){.\\@min-\\[48rem\\]\\/session-viewer\\:max-w-\\[85\\%\\]{max-width:85%}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:flex-1{flex:1}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:flex-none{flex:none}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:flex-1{flex:1}.\\@3xl\\/workspace\\:flex-none{flex:none}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:basis-auto{flex-basis:auto}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:grid-cols-\\[auto_minmax\\(0\\,1fr\\)_auto\\]{grid-template-columns:auto minmax(0,1fr) auto}}@container (width>=48rem){.\\@3xl\\:grid-cols-2{grid-template-columns:repeat(2,minmax(0,1fr))}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:grid-rows-\\[minmax\\(0\\,1fr\\)_auto\\]{grid-template-rows:minmax(0,1fr) auto}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:flex-row{flex-direction:row}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:flex-row{flex-direction:row}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:flex-col{flex-direction:column}.\\@3xl\\/workspace\\:flex-row{flex-direction:row}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:flex-nowrap{flex-wrap:nowrap}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:flex-nowrap{flex-wrap:nowrap}.\\@3xl\\/workspace\\:content-stretch{align-content:stretch}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:items-start{align-items:flex-start}.\\@3xl\\/app-shell\\:justify-between{justify-content:space-between}.\\@3xl\\/app-shell\\:gap-density-3{gap:var(--spacing-density-3)}}@container session-viewer (width>=48rem){.\\@min-\\[48rem\\]\\/session-viewer\\:gap-density-3{gap:var(--spacing-density-3)}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:overflow-hidden{overflow:hidden}.\\@3xl\\/workspace\\:border-r{border-right-style:var(--tw-border-style);border-right-width:1px}}@container split-pane (width>=48rem){.\\@3xl\\/split-pane\\:border-b-0{border-bottom-style:var(--tw-border-style);border-bottom-width:0}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:border-l{border-left-style:var(--tw-border-style);border-left-width:1px}}@container session-viewer (width>=48rem){.\\@min-\\[48rem\\]\\/session-viewer\\:p-density-4{padding:var(--spacing-density-4)}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:px-density-4{padding-inline:var(--spacing-density-4)}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:px-1{padding-inline:var(--spacing)}}@container session-viewer (width>=48rem){.\\@min-\\[48rem\\]\\/session-viewer\\:px-density-4{padding-inline:var(--spacing-density-4)}}@container app-shell (width>=48rem){.\\@3xl\\/app-shell\\:py-0{padding-block:0}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:py-2{padding-block:calc(var(--spacing) * 2)}.\\@3xl\\/workspace\\:pr-2{padding-right:calc(var(--spacing) * 2)}.\\@3xl\\/workspace\\:pr-\\[var\\(--workspace-top-right-reserve\\,0px\\)\\]{padding-right:var(--workspace-top-right-reserve,0px)}}@container session-viewer (width>=48rem){.\\@min-\\[48rem\\]\\/session-viewer\\:pb-density-4{padding-bottom:var(--spacing-density-4)}}@container workspace (width>=48rem){.\\@3xl\\/workspace\\:\\[writing-mode\\:vertical-rl\\]{writing-mode:vertical-rl}}@container endpoint (width>=56rem){.\\@min-\\[56rem\\]\\/endpoint\\:col-span-1{grid-column:span 1/span 1}.\\@min-\\[56rem\\]\\/endpoint\\:grid-cols-\\[minmax\\(14rem\\,1fr\\)_auto_auto_minmax\\(14rem\\,1fr\\)\\]{grid-template-columns:minmax(14rem,1fr) auto auto minmax(14rem,1fr)}}@container (width>=60rem){.\\@min-\\[60rem\\]\\:hidden{display:none}.\\@min-\\[60rem\\]\\:gap-2{gap:calc(var(--spacing) * 2)}.\\@min-\\[60rem\\]\\:px-2{padding-inline:calc(var(--spacing) * 2)}.\\@min-\\[60rem\\]\\:px-2\\.5{padding-inline:calc(var(--spacing) * 2.5)}.\\@min-\\[60rem\\]\\:px-3{padding-inline:calc(var(--spacing) * 3)}}@container (width>=72rem){.\\@6xl\\:grid-cols-3{grid-template-columns:repeat(3,minmax(0,1fr))}}@container app-content (width>=96rem){.\\@8xl\\/app-content\\:max-w-8xl{max-width:var(--container-8xl)}}@container app-content (width>=120rem){.\\@9xl\\/app-content\\:max-w-9xl{max-width:var(--container-9xl)}}@media (prefers-color-scheme:dark){.dark\\:border-amber-400{border-color:var(--color-amber-400)}.dark\\:border-amber-500\\/30{border-color:#f99c004d}@supports (color:color-mix(in lab, red, red)){.dark\\:border-amber-500\\/30{border-color:color-mix(in oklab, var(--color-amber-500) 30%, transparent)}}.dark\\:border-amber-500\\/40{border-color:#f99c0066}@supports (color:color-mix(in lab, red, red)){.dark\\:border-amber-500\\/40{border-color:color-mix(in oklab, var(--color-amber-500) 40%, transparent)}}.dark\\:border-amber-800{border-color:var(--color-amber-800)}.dark\\:border-amber-800\\/60{border-color:#953d0099}@supports (color:color-mix(in lab, red, red)){.dark\\:border-amber-800\\/60{border-color:color-mix(in oklab, var(--color-amber-800) 60%, transparent)}}.dark\\:border-emerald-500{border-color:var(--color-emerald-500)}.dark\\:border-emerald-500\\/30{border-color:#00bb7f4d}@supports (color:color-mix(in lab, red, red)){.dark\\:border-emerald-500\\/30{border-color:color-mix(in oklab, var(--color-emerald-500) 30%, transparent)}}.dark\\:border-emerald-500\\/40{border-color:#00bb7f66}@supports (color:color-mix(in lab, red, red)){.dark\\:border-emerald-500\\/40{border-color:color-mix(in oklab, var(--color-emerald-500) 40%, transparent)}}.dark\\:border-emerald-800{border-color:var(--color-emerald-800)}.dark\\:border-red-500{border-color:var(--color-red-500)}.dark\\:border-red-800{border-color:var(--color-red-800)}.dark\\:border-rose-500\\/30{border-color:#ff23574d}@supports (color:color-mix(in lab, red, red)){.dark\\:border-rose-500\\/30{border-color:color-mix(in oklab, var(--color-rose-500) 30%, transparent)}}.dark\\:border-rose-500\\/40{border-color:#ff235766}@supports (color:color-mix(in lab, red, red)){.dark\\:border-rose-500\\/40{border-color:color-mix(in oklab, var(--color-rose-500) 40%, transparent)}}.dark\\:border-rose-800{border-color:var(--color-rose-800)}.dark\\:border-sky-500{border-color:var(--color-sky-500)}.dark\\:border-sky-500\\/30{border-color:#00a5ef4d}@supports (color:color-mix(in lab, red, red)){.dark\\:border-sky-500\\/30{border-color:color-mix(in oklab, var(--color-sky-500) 30%, transparent)}}.dark\\:border-sky-500\\/40{border-color:#00a5ef66}@supports (color:color-mix(in lab, red, red)){.dark\\:border-sky-500\\/40{border-color:color-mix(in oklab, var(--color-sky-500) 40%, transparent)}}.dark\\:border-sky-800{border-color:var(--color-sky-800)}.dark\\:border-slate-300{border-color:var(--color-slate-300)}.dark\\:border-yellow-700{border-color:var(--color-yellow-700)}.dark\\:border-l-amber-500{border-left-color:var(--color-amber-500)}.dark\\:border-l-border{border-left-color:var(--border)}.dark\\:border-l-emerald-500{border-left-color:var(--color-emerald-500)}.dark\\:border-l-indigo-500{border-left-color:var(--color-indigo-500)}.dark\\:border-l-rose-500{border-left-color:var(--color-rose-500)}.dark\\:border-l-sky-500{border-left-color:var(--color-sky-500)}.dark\\:border-l-slate-500{border-left-color:var(--color-slate-500)}.dark\\:border-l-teal-500{border-left-color:var(--color-teal-500)}.dark\\:border-l-violet-500{border-left-color:var(--color-violet-500)}.dark\\:bg-amber-400{background-color:var(--color-amber-400)}.dark\\:bg-amber-400\\/10{background-color:#fcbb001a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-400\\/10{background-color:color-mix(in oklab, var(--color-amber-400) 10%, transparent)}}.dark\\:bg-amber-500{background-color:var(--color-amber-500)}.dark\\:bg-amber-500\\/10{background-color:#f99c001a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-500\\/10{background-color:color-mix(in oklab, var(--color-amber-500) 10%, transparent)}}.dark\\:bg-amber-500\\/15{background-color:#f99c0026}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-500\\/15{background-color:color-mix(in oklab, var(--color-amber-500) 15%, transparent)}}.dark\\:bg-amber-500\\/20{background-color:#f99c0033}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-500\\/20{background-color:color-mix(in oklab, var(--color-amber-500) 20%, transparent)}}.dark\\:bg-amber-900{background-color:var(--color-amber-900)}.dark\\:bg-amber-950\\/20{background-color:#46190133}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-950\\/20{background-color:color-mix(in oklab, var(--color-amber-950) 20%, transparent)}}.dark\\:bg-amber-950\\/40{background-color:#46190166}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-amber-950\\/40{background-color:color-mix(in oklab, var(--color-amber-950) 40%, transparent)}}.dark\\:bg-blue-500\\/10{background-color:#3080ff1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-blue-500\\/10{background-color:color-mix(in oklab, var(--color-blue-500) 10%, transparent)}}.dark\\:bg-blue-500\\/20{background-color:#3080ff33}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-blue-500\\/20{background-color:color-mix(in oklab, var(--color-blue-500) 20%, transparent)}}.dark\\:bg-emerald-400\\/10{background-color:#00d2941a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-emerald-400\\/10{background-color:color-mix(in oklab, var(--color-emerald-400) 10%, transparent)}}.dark\\:bg-emerald-500{background-color:var(--color-emerald-500)}.dark\\:bg-emerald-500\\/15{background-color:#00bb7f26}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-emerald-500\\/15{background-color:color-mix(in oklab, var(--color-emerald-500) 15%, transparent)}}.dark\\:bg-emerald-900{background-color:var(--color-emerald-900)}.dark\\:bg-emerald-950\\/40{background-color:#002c2266}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-emerald-950\\/40{background-color:color-mix(in oklab, var(--color-emerald-950) 40%, transparent)}}.dark\\:bg-gray-500\\/20{background-color:#6a728233}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-gray-500\\/20{background-color:color-mix(in oklab, var(--color-gray-500) 20%, transparent)}}.dark\\:bg-green-500\\/20{background-color:#00c75833}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-green-500\\/20{background-color:color-mix(in oklab, var(--color-green-500) 20%, transparent)}}.dark\\:bg-indigo-400\\/10{background-color:#7d87ff1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-indigo-400\\/10{background-color:color-mix(in oklab, var(--color-indigo-400) 10%, transparent)}}.dark\\:bg-indigo-500{background-color:var(--color-indigo-500)}.dark\\:bg-indigo-950\\/50{background-color:#1e1a4d80}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-indigo-950\\/50{background-color:color-mix(in oklab, var(--color-indigo-950) 50%, transparent)}}.dark\\:bg-muted-foreground{background-color:var(--muted-foreground)}.dark\\:bg-red-500{background-color:var(--color-red-500)}.dark\\:bg-red-500\\/20{background-color:#fb2c3633}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-red-500\\/20{background-color:color-mix(in oklab, var(--color-red-500) 20%, transparent)}}.dark\\:bg-red-900{background-color:var(--color-red-900)}.dark\\:bg-red-950{background-color:var(--color-red-950)}.dark\\:bg-rose-400\\/10{background-color:#ff667f1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-rose-400\\/10{background-color:color-mix(in oklab, var(--color-rose-400) 10%, transparent)}}.dark\\:bg-rose-500{background-color:var(--color-rose-500)}.dark\\:bg-rose-500\\/15{background-color:#ff235726}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-rose-500\\/15{background-color:color-mix(in oklab, var(--color-rose-500) 15%, transparent)}}.dark\\:bg-rose-950\\/30{background-color:#4d02184d}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-rose-950\\/30{background-color:color-mix(in oklab, var(--color-rose-950) 30%, transparent)}}.dark\\:bg-rose-950\\/40{background-color:#4d021866}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-rose-950\\/40{background-color:color-mix(in oklab, var(--color-rose-950) 40%, transparent)}}.dark\\:bg-rose-950\\/50{background-color:#4d021880}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-rose-950\\/50{background-color:color-mix(in oklab, var(--color-rose-950) 50%, transparent)}}.dark\\:bg-sky-400\\/10{background-color:#00bcfe1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-sky-400\\/10{background-color:color-mix(in oklab, var(--color-sky-400) 10%, transparent)}}.dark\\:bg-sky-500{background-color:var(--color-sky-500)}.dark\\:bg-sky-500\\/15{background-color:#00a5ef26}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-sky-500\\/15{background-color:color-mix(in oklab, var(--color-sky-500) 15%, transparent)}}.dark\\:bg-sky-900{background-color:var(--color-sky-900)}.dark\\:bg-sky-950\\/40{background-color:#052f4a66}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-sky-950\\/40{background-color:color-mix(in oklab, var(--color-sky-950) 40%, transparent)}}.dark\\:bg-slate-300{background-color:var(--color-slate-300)}.dark\\:bg-slate-400\\/10{background-color:#90a1b91a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-slate-400\\/10{background-color:color-mix(in oklab, var(--color-slate-400) 10%, transparent)}}.dark\\:bg-slate-500{background-color:var(--color-slate-500)}.dark\\:bg-slate-600{background-color:var(--color-slate-600)}.dark\\:bg-teal-400\\/10{background-color:#00d3bd1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-teal-400\\/10{background-color:color-mix(in oklab, var(--color-teal-400) 10%, transparent)}}.dark\\:bg-teal-500{background-color:var(--color-teal-500)}.dark\\:bg-violet-400\\/10{background-color:#a685ff1a}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-violet-400\\/10{background-color:color-mix(in oklab, var(--color-violet-400) 10%, transparent)}}.dark\\:bg-violet-500{background-color:var(--color-violet-500)}.dark\\:bg-violet-500\\/20{background-color:#8d54ff33}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-violet-500\\/20{background-color:color-mix(in oklab, var(--color-violet-500) 20%, transparent)}}.dark\\:bg-yellow-500\\/20{background-color:#edb20033}@supports (color:color-mix(in lab, red, red)){.dark\\:bg-yellow-500\\/20{background-color:color-mix(in oklab, var(--color-yellow-500) 20%, transparent)}}.dark\\:bg-yellow-950{background-color:var(--color-yellow-950)}.dark\\:text-\\[\\#00f7f1\\]{color:#00f7f1}.dark\\:text-\\[\\#61dafb\\]{color:#61dafb}.dark\\:text-\\[\\#8080f2\\]{color:#8080f2}.dark\\:text-\\[\\#d7ff64\\]{color:#d7ff64}.dark\\:text-\\[\\#e5576a\\]{color:#e5576a}.dark\\:text-\\[\\#ffd43b\\]{color:#ffd43b}.dark\\:text-amber-100{color:var(--color-amber-100)}.dark\\:text-amber-100\\/90{color:#fef3c6e6}@supports (color:color-mix(in lab, red, red)){.dark\\:text-amber-100\\/90{color:color-mix(in oklab, var(--color-amber-100) 90%, transparent)}}.dark\\:text-amber-300{color:var(--color-amber-300)}.dark\\:text-amber-300\\/70{color:#ffd236b3}@supports (color:color-mix(in lab, red, red)){.dark\\:text-amber-300\\/70{color:color-mix(in oklab, var(--color-amber-300) 70%, transparent)}}.dark\\:text-amber-400{color:var(--color-amber-400)}.dark\\:text-amber-500{color:var(--color-amber-500)}.dark\\:text-amber-950{color:var(--color-amber-950)}.dark\\:text-blue-100{color:var(--color-blue-100)}.dark\\:text-blue-100\\/90{color:#dbeafee6}@supports (color:color-mix(in lab, red, red)){.dark\\:text-blue-100\\/90{color:color-mix(in oklab, var(--color-blue-100) 90%, transparent)}}.dark\\:text-blue-300{color:var(--color-blue-300)}.dark\\:text-blue-300\\/70{color:#90c5ffb3}@supports (color:color-mix(in lab, red, red)){.dark\\:text-blue-300\\/70{color:color-mix(in oklab, var(--color-blue-300) 70%, transparent)}}.dark\\:text-blue-400{color:var(--color-blue-400)}.dark\\:text-emerald-100{color:var(--color-emerald-100)}.dark\\:text-emerald-100\\/90{color:#d0fae5e6}@supports (color:color-mix(in lab, red, red)){.dark\\:text-emerald-100\\/90{color:color-mix(in oklab, var(--color-emerald-100) 90%, transparent)}}.dark\\:text-emerald-300{color:var(--color-emerald-300)}.dark\\:text-emerald-300\\/70{color:#5ee9b5b3}@supports (color:color-mix(in lab, red, red)){.dark\\:text-emerald-300\\/70{color:color-mix(in oklab, var(--color-emerald-300) 70%, transparent)}}.dark\\:text-emerald-400{color:var(--color-emerald-400)}.dark\\:text-emerald-950{color:var(--color-emerald-950)}.dark\\:text-green-300{color:var(--color-green-300)}.dark\\:text-green-400{color:var(--color-green-400)}.dark\\:text-indigo-200{color:var(--color-indigo-200)}.dark\\:text-indigo-300{color:var(--color-indigo-300)}.dark\\:text-orange-300{color:var(--color-orange-300)}.dark\\:text-purple-100{color:var(--color-purple-100)}.dark\\:text-purple-100\\/90{color:#f3e8ffe6}@supports (color:color-mix(in lab, red, red)){.dark\\:text-purple-100\\/90{color:color-mix(in oklab, var(--color-purple-100) 90%, transparent)}}.dark\\:text-purple-300{color:var(--color-purple-300)}.dark\\:text-purple-300\\/70{color:#d9b3ffb3}@supports (color:color-mix(in lab, red, red)){.dark\\:text-purple-300\\/70{color:color-mix(in oklab, var(--color-purple-300) 70%, transparent)}}.dark\\:text-purple-400{color:var(--color-purple-400)}.dark\\:text-red-100{color:var(--color-red-100)}.dark\\:text-red-100\\/90{color:#ffe2e2e6}@supports (color:color-mix(in lab, red, red)){.dark\\:text-red-100\\/90{color:color-mix(in oklab, var(--color-red-100) 90%, transparent)}}.dark\\:text-red-200{color:var(--color-red-200)}.dark\\:text-red-300{color:var(--color-red-300)}.dark\\:text-red-300\\/70{color:#ffa3a3b3}@supports (color:color-mix(in lab, red, red)){.dark\\:text-red-300\\/70{color:color-mix(in oklab, var(--color-red-300) 70%, transparent)}}.dark\\:text-red-400{color:var(--color-red-400)}.dark\\:text-red-950{color:var(--color-red-950)}.dark\\:text-rose-200{color:var(--color-rose-200)}.dark\\:text-rose-300{color:var(--color-rose-300)}.dark\\:text-rose-400{color:var(--color-rose-400)}.dark\\:text-sky-300{color:var(--color-sky-300)}.dark\\:text-sky-400{color:var(--color-sky-400)}.dark\\:text-sky-950{color:var(--color-sky-950)}.dark\\:text-slate-300{color:var(--color-slate-300)}.dark\\:text-slate-500{color:var(--color-slate-500)}.dark\\:text-slate-950{color:var(--color-slate-950)}.dark\\:text-teal-300{color:var(--color-teal-300)}.dark\\:text-violet-300{color:var(--color-violet-300)}.dark\\:text-violet-400{color:var(--color-violet-400)}.dark\\:text-yellow-200{color:var(--color-yellow-200)}.dark\\:text-yellow-300{color:var(--color-yellow-300)}.dark\\:text-yellow-400{color:var(--color-yellow-400)}.dark\\:ring-amber-400\\/25{--tw-ring-color:#fcbb0040}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-amber-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-amber-400) 25%, transparent)}}.dark\\:ring-amber-500\\/30{--tw-ring-color:#f99c004d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-amber-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-amber-500) 30%, transparent)}}.dark\\:ring-emerald-400\\/25{--tw-ring-color:#00d29440}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-emerald-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-emerald-400) 25%, transparent)}}.dark\\:ring-emerald-500\\/30{--tw-ring-color:#00bb7f4d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-emerald-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-emerald-500) 30%, transparent)}}.dark\\:ring-indigo-400\\/25{--tw-ring-color:#7d87ff40}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-indigo-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-indigo-400) 25%, transparent)}}.dark\\:ring-rose-400\\/25{--tw-ring-color:#ff667f40}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-rose-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-rose-400) 25%, transparent)}}.dark\\:ring-rose-500\\/30{--tw-ring-color:#ff23574d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-rose-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-rose-500) 30%, transparent)}}.dark\\:ring-sky-400\\/25{--tw-ring-color:#00bcfe40}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-sky-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-sky-400) 25%, transparent)}}.dark\\:ring-sky-500\\/30{--tw-ring-color:#00a5ef4d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-sky-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-sky-500) 30%, transparent)}}.dark\\:ring-slate-400\\/25{--tw-ring-color:#90a1b940}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-slate-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-slate-400) 25%, transparent)}}.dark\\:ring-slate-500\\/30{--tw-ring-color:#62748e4d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-slate-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-slate-500) 30%, transparent)}}.dark\\:ring-teal-400\\/25{--tw-ring-color:#00d3bd40}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-teal-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-teal-400) 25%, transparent)}}.dark\\:ring-violet-400\\/25{--tw-ring-color:#a685ff40}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-violet-400\\/25{--tw-ring-color:color-mix(in oklab, var(--color-violet-400) 25%, transparent)}}.dark\\:ring-violet-500\\/30{--tw-ring-color:#8d54ff4d}@supports (color:color-mix(in lab, red, red)){.dark\\:ring-violet-500\\/30{--tw-ring-color:color-mix(in oklab, var(--color-violet-500) 30%, transparent)}}@media (hover:hover){.dark\\:hover\\:border-emerald-500\\/40:hover{border-color:#00bb7f66}@supports (color:color-mix(in lab, red, red)){.dark\\:hover\\:border-emerald-500\\/40:hover{border-color:color-mix(in oklab, var(--color-emerald-500) 40%, transparent)}}.dark\\:hover\\:bg-emerald-500\\/10:hover{background-color:#00bb7f1a}@supports (color:color-mix(in lab, red, red)){.dark\\:hover\\:bg-emerald-500\\/10:hover{background-color:color-mix(in oklab, var(--color-emerald-500) 10%, transparent)}}}}@media print{.print\\:\\[-webkit-print-color-adjust\\:exact\\]{-webkit-print-color-adjust:exact}.print\\:\\[print-color-adjust\\:exact\\]{-webkit-print-color-adjust:exact;print-color-adjust:exact}}.\\[\\&_\\*\\]\\:pointer-events-none *{pointer-events:none}.\\[\\&_\\.chroma\\]\\:bg-transparent .chroma{background-color:#0000}.\\[\\&_\\.chroma_\\.c\\]\\:text-slate-500 .chroma .c{color:var(--color-slate-500)}.\\[\\&_\\.chroma_\\.c\\]\\:italic .chroma .c{font-style:italic}.\\[\\&_\\.chroma_\\.cp\\]\\:text-slate-500 .chroma .cp{color:var(--color-slate-500)}.\\[\\&_\\.chroma_\\.cp\\]\\:italic .chroma .cp{font-style:italic}.\\[\\&_\\.chroma_\\.na\\]\\:text-amber-700 .chroma .na{color:var(--color-amber-700)}@media (prefers-color-scheme:dark){.\\[\\&_\\.chroma_\\.na\\]\\:dark\\:text-amber-300 .chroma .na{color:var(--color-amber-300)}}.\\[\\&_\\.chroma_\\.nt\\]\\:font-semibold .chroma .nt{--tw-font-weight:var(--font-weight-semibold);font-weight:var(--font-weight-semibold)}.\\[\\&_\\.chroma_\\.nt\\]\\:text-pink-700 .chroma .nt{color:var(--color-pink-700)}@media (prefers-color-scheme:dark){.\\[\\&_\\.chroma_\\.nt\\]\\:dark\\:text-pink-300 .chroma .nt{color:var(--color-pink-300)}}.\\[\\&_\\.chroma_\\.s\\]\\:text-emerald-700 .chroma .s{color:var(--color-emerald-700)}@media (prefers-color-scheme:dark){.\\[\\&_\\.chroma_\\.s\\]\\:dark\\:text-emerald-300 .chroma .s{color:var(--color-emerald-300)}}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:absolute [data-profile-tree-item]:before{content:var(--tw-content);position:absolute}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:top-1\\/2 [data-profile-tree-item]:before{content:var(--tw-content);top:50%}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:-left-4 [data-profile-tree-item]:before{content:var(--tw-content);left:calc(var(--spacing) * -4)}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:h-px [data-profile-tree-item]:before{content:var(--tw-content);height:1px}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:w-4 [data-profile-tree-item]:before{content:var(--tw-content);width:calc(var(--spacing) * 4)}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:bg-border [data-profile-tree-item]:before{content:var(--tw-content);background-color:var(--border)}.\\[\\&_\\[data-profile-tree-item\\]\\]\\:before\\:content-\\[\\'\\'\\] [data-profile-tree-item]:before{--tw-content:\"\";content:var(--tw-content)}.\\[\\&_\\[data-slot\\=monaco-editor\\]\\]\\:h-full [data-slot=monaco-editor]{height:100%}.\\[\\&_\\[data-slot\\=monaco-editor\\]\\]\\:rounded-none [data-slot=monaco-editor]{border-radius:0}.\\[\\&_\\[data-slot\\=monaco-editor\\]\\]\\:border-0 [data-slot=monaco-editor]{border-style:var(--tw-border-style);border-width:0}.\\[\\&_\\[data-slot\\=monaco-editor\\]\\]\\:border-t [data-slot=monaco-editor]{border-top-style:var(--tw-border-style);border-top-width:1px}.\\[\\&_h2\\]\\:my-2 h2{margin-block:calc(var(--spacing) * 2)}.\\[\\&_p\\]\\:my-1 p{margin-block:var(--spacing)}.\\[\\&_pre\\]\\:m-0 pre{margin:0}.\\[\\&_pre\\]\\:inline pre{display:inline}.\\[\\&_pre\\]\\:bg-transparent pre{background-color:#0000}.\\[\\&_pre\\]\\:leading-relaxed pre{--tw-leading:var(--leading-relaxed);line-height:var(--leading-relaxed)}.\\[\\&_pre\\]\\:whitespace-pre pre{white-space:pre}.\\[\\&_pre\\]\\:whitespace-pre-wrap pre{white-space:pre-wrap}.\\[\\&_pre\\.shiki\\]\\:\\!bg-transparent pre.shiki{background-color:#0000!important}.\\[\\&_span\\[data-jsf-readonly\\]\\]\\:h-auto span[data-jsf-readonly]{height:auto}.\\[\\&_span\\[data-jsf-readonly\\]\\]\\:min-w-0 span[data-jsf-readonly]{min-width:0}.\\[\\&_svg\\]\\:pointer-events-none svg{pointer-events:none}.\\[\\&_svg\\]\\:size-3\\.5 svg{width:calc(var(--spacing) * 3.5);height:calc(var(--spacing) * 3.5)}.\\[\\&_svg\\]\\:size-4 svg{width:calc(var(--spacing) * 4);height:calc(var(--spacing) * 4)}.\\[\\&_svg\\]\\:size-5 svg{width:calc(var(--spacing) * 5);height:calc(var(--spacing) * 5)}.\\[\\&_svg\\]\\:size-\\[0\\.75em\\] svg{width:.75em;height:.75em}.\\[\\&_svg\\]\\:shrink-0 svg{flex-shrink:0}.\\[\\&_svg\\]\\:text-muted-foreground svg{color:var(--muted-foreground)}.\\[\\&\\:\\:-moz-range-thumb\\]\\:pointer-events-auto::-moz-range-thumb{pointer-events:auto}.\\[\\&\\:\\:-moz-range-thumb\\]\\:size-3\\.5::-moz-range-thumb{width:calc(var(--spacing) * 3.5);height:calc(var(--spacing) * 3.5)}.\\[\\&\\:\\:-moz-range-thumb\\]\\:appearance-none::-moz-range-thumb{appearance:none}.\\[\\&\\:\\:-moz-range-thumb\\]\\:rounded-full::-moz-range-thumb{border-radius:2147483647px}.\\[\\&\\:\\:-moz-range-thumb\\]\\:border::-moz-range-thumb{border-style:var(--tw-border-style);border-width:1px}.\\[\\&\\:\\:-moz-range-thumb\\]\\:border-0::-moz-range-thumb{border-style:var(--tw-border-style);border-width:0}.\\[\\&\\:\\:-moz-range-thumb\\]\\:border-background\\/80::-moz-range-thumb{border-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.\\[\\&\\:\\:-moz-range-thumb\\]\\:border-background\\/80::-moz-range-thumb{border-color:color-mix(in oklab, var(--background) 80%, transparent)}}.\\[\\&\\:\\:-moz-range-thumb\\]\\:bg-muted-foreground::-moz-range-thumb{background-color:var(--muted-foreground)}.\\[\\&\\:\\:-moz-range-thumb\\]\\:bg-primary::-moz-range-thumb{background-color:var(--primary)}.\\[\\&\\:\\:-moz-range-thumb\\]\\:shadow-sm::-moz-range-thumb{--tw-shadow:0 1px 3px 0 var(--tw-shadow-color,#0000001a), 0 1px 2px -1px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.\\[\\&\\:\\:-moz-range-track\\]\\:h-1::-moz-range-track{height:var(--spacing)}.\\[\\&\\:\\:-moz-range-track\\]\\:rounded-full::-moz-range-track{border-radius:2147483647px}.\\[\\&\\:\\:-moz-range-track\\]\\:bg-transparent::-moz-range-track{background-color:#0000}.\\[\\&\\:\\:-webkit-calendar-picker-indicator\\]\\:hidden::-webkit-calendar-picker-indicator{display:none}.\\[\\&\\:\\:-webkit-calendar-picker-indicator\\]\\:appearance-none::-webkit-calendar-picker-indicator{appearance:none}.\\[\\&\\:\\:-webkit-inner-spin-button\\]\\:appearance-none::-webkit-inner-spin-button{appearance:none}.\\[\\&\\:\\:-webkit-outer-spin-button\\]\\:appearance-none::-webkit-outer-spin-button{appearance:none}.\\[\\&\\:\\:-webkit-search-cancel-button\\]\\:appearance-none::-webkit-search-cancel-button{appearance:none}.\\[\\&\\:\\:-webkit-slider-runnable-track\\]\\:h-1::-webkit-slider-runnable-track{height:var(--spacing)}.\\[\\&\\:\\:-webkit-slider-runnable-track\\]\\:rounded-full::-webkit-slider-runnable-track{border-radius:2147483647px}.\\[\\&\\:\\:-webkit-slider-runnable-track\\]\\:bg-transparent::-webkit-slider-runnable-track{background-color:#0000}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:pointer-events-auto::-webkit-slider-thumb{pointer-events:auto}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:-mt-\\[5px\\]::-webkit-slider-thumb{margin-top:-5px}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:mt-\\[-5px\\]::-webkit-slider-thumb{margin-top:-5px}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:size-3\\.5::-webkit-slider-thumb{width:calc(var(--spacing) * 3.5);height:calc(var(--spacing) * 3.5)}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:appearance-none::-webkit-slider-thumb{appearance:none}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:rounded-full::-webkit-slider-thumb{border-radius:2147483647px}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:border::-webkit-slider-thumb{border-style:var(--tw-border-style);border-width:1px}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:border-background\\/80::-webkit-slider-thumb{border-color:var(--background)}@supports (color:color-mix(in lab, red, red)){.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:border-background\\/80::-webkit-slider-thumb{border-color:color-mix(in oklab, var(--background) 80%, transparent)}}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:bg-muted-foreground::-webkit-slider-thumb{background-color:var(--muted-foreground)}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:bg-primary::-webkit-slider-thumb{background-color:var(--primary)}.\\[\\&\\:\\:-webkit-slider-thumb\\]\\:shadow-sm::-webkit-slider-thumb{--tw-shadow:0 1px 3px 0 var(--tw-shadow-color,#0000001a), 0 1px 2px -1px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.\\[\\&\\:not\\(\\:first-child\\)\\]\\:ml-2:not(:first-child){margin-left:calc(var(--spacing) * 2)}.\\[\\&\\>\\*\\:first-child\\]\\:mt-0>:first-child{margin-top:0}.\\[\\&\\>\\*\\:last-child\\]\\:mb-0>:last-child{margin-bottom:0}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\]\\:flex>[data-slot=modal-body]{display:flex}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\]\\:min-h-0>[data-slot=modal-body]{min-height:0}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\]\\:overflow-hidden>[data-slot=modal-body]{overflow:hidden}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\>\\*\\]\\:h-auto>[data-slot=modal-body]>*{height:auto}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\>\\*\\]\\:min-h-0>[data-slot=modal-body]>*{min-height:0}.\\[\\&\\>\\[data-slot\\=modal-body\\]\\>\\*\\]\\:flex-1>[data-slot=modal-body]>*{flex:1}.\\[\\&\\>\\[data-slot\\=monaco-editor\\]\\]\\:h-full>[data-slot=monaco-editor]{height:100%}.\\[\\&\\>button\\]\\:w-full>button{width:100%}.\\[\\&\\>button\\]\\:justify-start>button{justify-content:flex-start}.\\[\\&\\>p\\]\\:inline>p{display:inline}.\\[\\&\\>span\\]\\:w-full>span{width:100%}.\\[\\&\\>span\\]\\:min-w-0>span{min-width:0}.\\[\\&\\>svg\\]\\:h-4>svg{height:calc(var(--spacing) * 4)}.\\[\\&\\>svg\\]\\:w-4>svg{width:calc(var(--spacing) * 4)}:hover>.\\[\\:hover\\>\\&\\]\\:opacity-100{opacity:1}@media (hover:none){.\\[\\@media\\(hover\\:none\\)\\]\\:opacity-70{opacity:.7}.\\[\\@media\\(hover\\:none\\)\\]\\:opacity-100{opacity:1}}@media (pointer:coarse){.\\[\\@media\\(pointer\\:coarse\\)\\]\\:static{position:static}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:top-1{top:var(--spacing)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:right-1{right:var(--spacing)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:mt-2{margin-top:calc(var(--spacing) * 2)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:size-11{width:calc(var(--spacing) * 11);height:calc(var(--spacing) * 11)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:min-h-8{min-height:calc(var(--spacing) * 8)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:min-h-11{min-height:calc(var(--spacing) * 11)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:w-8{width:calc(var(--spacing) * 8)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:min-w-11{min-width:calc(var(--spacing) * 11)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:flex-wrap{flex-wrap:wrap}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:gap-2{gap:calc(var(--spacing) * 2)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:p-2\\.5{padding:calc(var(--spacing) * 2.5)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:px-3{padding-inline:calc(var(--spacing) * 3)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:pr-12{padding-right:calc(var(--spacing) * 12)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:pb-2{padding-bottom:calc(var(--spacing) * 2)}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:text-base{font-size:var(--text-base);line-height:var(--tw-leading,var(--text-base--line-height))}.\\[\\@media\\(pointer\\:coarse\\)\\]\\:text-sm{font-size:var(--text-sm);line-height:var(--tw-leading,var(--text-sm--line-height))}}[aria-expanded=true] .\\[\\[aria-expanded\\=true\\]_\\&\\]\\:bg-muted{background-color:var(--muted)}[aria-selected=true]>div>.\\[\\[aria-selected\\=true\\]\\>div\\>\\&\\]\\:opacity-100{opacity:1}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-amber-400{border-color:var(--color-amber-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-emerald-400{border-color:var(--color-emerald-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-emerald-700{border-color:var(--color-emerald-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-fuchsia-400{border-color:var(--color-fuchsia-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-fuchsia-700{border-color:var(--color-fuchsia-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-indigo-400{border-color:var(--color-indigo-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-indigo-700{border-color:var(--color-indigo-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-orange-400{border-color:var(--color-orange-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-orange-700{border-color:var(--color-orange-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-pink-400{border-color:var(--color-pink-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-rose-400{border-color:var(--color-rose-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-rose-500\\/30{border-color:#ff23574d}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-rose-500\\/30{border-color:color-mix(in oklab, var(--color-rose-500) 30%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-rose-700{border-color:var(--color-rose-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-sky-400{border-color:var(--color-sky-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-sky-700{border-color:var(--color-sky-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-teal-400{border-color:var(--color-teal-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-teal-700{border-color:var(--color-teal-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-violet-400{border-color:var(--color-violet-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:border-violet-700{border-color:var(--color-violet-700)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-\\[var\\(--clicky-bg-subtle\\)\\]{background-color:var(--clicky-bg-subtle)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400{background-color:var(--color-amber-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/5{background-color:#fcbb000d}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/5{background-color:color-mix(in oklab, var(--color-amber-400) 5%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/10{background-color:#fcbb001a}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/10{background-color:color-mix(in oklab, var(--color-amber-400) 10%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/20{background-color:#fcbb0033}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/20{background-color:color-mix(in oklab, var(--color-amber-400) 20%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/25{background-color:#fcbb0040}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/25{background-color:color-mix(in oklab, var(--color-amber-400) 25%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/50{background-color:#fcbb0080}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-400\\/50{background-color:color-mix(in oklab, var(--color-amber-400) 50%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-500\\/15{background-color:#f99c0026}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-500\\/15{background-color:color-mix(in oklab, var(--color-amber-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-950\\/40{background-color:#46190166}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-amber-950\\/40{background-color:color-mix(in oklab, var(--color-amber-950) 40%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-emerald-400{background-color:var(--color-emerald-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-emerald-500\\/15{background-color:#00bb7f26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-emerald-500\\/15{background-color:color-mix(in oklab, var(--color-emerald-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-emerald-950\\/60{background-color:#002c2299}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-emerald-950\\/60{background-color:color-mix(in oklab, var(--color-emerald-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-fuchsia-500\\/15{background-color:#e12afb26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-fuchsia-500\\/15{background-color:color-mix(in oklab, var(--color-fuchsia-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-fuchsia-950\\/60{background-color:#4b004f99}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-fuchsia-950\\/60{background-color:color-mix(in oklab, var(--color-fuchsia-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-green-500\\/20{background-color:#00c75833}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-green-500\\/20{background-color:color-mix(in oklab, var(--color-green-500) 20%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-indigo-500\\/15{background-color:#625fff26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-indigo-500\\/15{background-color:color-mix(in oklab, var(--color-indigo-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-indigo-950\\/60{background-color:#1e1a4d99}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-indigo-950\\/60{background-color:color-mix(in oklab, var(--color-indigo-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-orange-500\\/15{background-color:#fe6e0026}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-orange-500\\/15{background-color:color-mix(in oklab, var(--color-orange-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-orange-950\\/60{background-color:#44130699}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-orange-950\\/60{background-color:color-mix(in oklab, var(--color-orange-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-pink-500\\/15{background-color:#f6339a26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-pink-500\\/15{background-color:color-mix(in oklab, var(--color-pink-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-red-400{background-color:var(--color-red-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-500\\/10{background-color:#ff23571a}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-500\\/10{background-color:color-mix(in oklab, var(--color-rose-500) 10%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-500\\/15{background-color:#ff235726}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-500\\/15{background-color:color-mix(in oklab, var(--color-rose-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-950\\/60{background-color:#4d021899}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-rose-950\\/60{background-color:color-mix(in oklab, var(--color-rose-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-sky-500\\/15{background-color:#00a5ef26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-sky-500\\/15{background-color:color-mix(in oklab, var(--color-sky-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-sky-950\\/60{background-color:#052f4a99}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-sky-950\\/60{background-color:color-mix(in oklab, var(--color-sky-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-teal-500\\/15{background-color:#00baa726}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-teal-500\\/15{background-color:color-mix(in oklab, var(--color-teal-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-teal-950\\/60{background-color:#022f2e99}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-teal-950\\/60{background-color:color-mix(in oklab, var(--color-teal-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-violet-500\\/15{background-color:#8d54ff26}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-violet-500\\/15{background-color:color-mix(in oklab, var(--color-violet-500) 15%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-violet-950\\/60{background-color:#2f0d6899}@supports (color:color-mix(in lab, red, red)){[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:bg-violet-950\\/60{background-color:color-mix(in oklab, var(--color-violet-950) 60%, transparent)}}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-\\[\\#20B2AA\\]{color:#20b2aa}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-\\[\\#D97757\\]{color:#d97757}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-\\[\\#F97316\\]{color:#f97316}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-\\[\\#FFB000\\]{color:#ffb000}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-amber-200{color:var(--color-amber-200)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-amber-300{color:var(--color-amber-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-amber-400{color:var(--color-amber-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-blue-300{color:var(--color-blue-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-blue-400{color:var(--color-blue-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-cyan-300{color:var(--color-cyan-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-emerald-300{color:var(--color-emerald-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-emerald-400{color:var(--color-emerald-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-fuchsia-300{color:var(--color-fuchsia-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-fuchsia-400{color:var(--color-fuchsia-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-green-300{color:var(--color-green-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-green-400{color:var(--color-green-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-indigo-300{color:var(--color-indigo-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-indigo-400{color:var(--color-indigo-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-orange-300{color:var(--color-orange-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-orange-400{color:var(--color-orange-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-pink-300{color:var(--color-pink-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-pink-400{color:var(--color-pink-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-purple-400{color:var(--color-purple-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-red-400{color:var(--color-red-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-rose-300{color:var(--color-rose-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-rose-400{color:var(--color-rose-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-sky-300{color:var(--color-sky-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-sky-400{color:var(--color-sky-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-slate-300{color:var(--color-slate-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-slate-400{color:var(--color-slate-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-teal-300{color:var(--color-teal-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-teal-400{color:var(--color-teal-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-violet-300{color:var(--color-violet-300)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-violet-400{color:var(--color-violet-400)}[data-theme=dark] .\\[\\[data-theme\\=dark\\]_\\&\\]\\:text-white{color:var(--color-white)}li .\\[li_\\&\\]\\:pl-6{padding-left:calc(var(--spacing) * 6)}}:root,[data-theme=light]{--clicky-primary-blue:#3578e5;--clicky-cta-blue:#0053f8;--clicky-dark-navy:#131f3b;--clicky-sky-500:#0ea5e9;--clicky-indigo-500:#6366f1;--clicky-gray-50:#f9fafb;--clicky-gray-100:#f3f4f6;--clicky-gray-200:#e5e7eb;--clicky-gray-300:#d1d5db;--clicky-gray-400:#9ca3af;--clicky-gray-500:#6b7280;--clicky-gray-600:#4b5563;--clicky-gray-700:#374151;--clicky-gray-800:#1f2937;--clicky-gray-900:#111827;--clicky-gray-950:#030712;--clicky-success-50:#ecfdf5;--clicky-success-100:#d1fae5;--clicky-success-200:#a7f3d0;--clicky-success-500:#10b981;--clicky-success-600:#059669;--clicky-success-700:#047857;--clicky-success:var(--clicky-success-600);--clicky-error-50:#fff1f2;--clicky-error-100:#ffe4e6;--clicky-error-200:#fecdd3;--clicky-error-500:#f43f5e;--clicky-error-600:#e11d48;--clicky-error-700:#be123c;--clicky-error:var(--clicky-error-600);--clicky-warning-50:#fffbeb;--clicky-warning-100:#fef3c7;--clicky-warning-200:#fde68a;--clicky-warning-500:#f59e0b;--clicky-warning-600:#d97706;--clicky-warning-700:#b45309;--clicky-warning:var(--clicky-warning-600);--clicky-info-50:#f0f9ff;--clicky-info-100:#e0f2fe;--clicky-info-200:#bae6fd;--clicky-info-500:#0ea5e9;--clicky-info-600:#0284c7;--clicky-info-700:#0369a1;--clicky-info:var(--clicky-info-600);--clicky-hint-50:#f0fdfa;--clicky-hint-100:#ccfbf1;--clicky-hint-200:#99f6e4;--clicky-hint-500:#14b8a6;--clicky-hint-600:#0d9488;--clicky-hint-700:#0f766e;--clicky-hint:var(--clicky-hint-600);--clicky-prereq-50:#f5f3ff;--clicky-prereq-100:#ede9fe;--clicky-prereq-200:#ddd6fe;--clicky-prereq-500:#8b5cf6;--clicky-prereq-600:#7c3aed;--clicky-prereq-700:#6d28d9;--clicky-prereq:var(--clicky-prereq-600);--clicky-fg-1:#061431;--clicky-fg-2:#030712;--clicky-fg-3:#374151;--clicky-fg-muted:#666;--clicky-fg-subtle:#9ca3af;--clicky-fg-on-brand:#fff;--clicky-fg-link:var(--clicky-primary-blue);--clicky-bg-canvas:#fff;--clicky-bg-subtle:#f9fafb;--clicky-bg-muted:#f3f4f6;--clicky-bg-inset:#e5e7eb;--clicky-bg-inverse:var(--clicky-dark-navy);--clicky-bg-brand:var(--clicky-primary-blue);--clicky-bg-cta:var(--clicky-cta-blue);--clicky-border-subtle:#e5e7eb;--clicky-border-default:#d1d5db;--clicky-border-strong:#9ca3af;--clicky-border-focus:var(--clicky-primary-blue);--clicky-font-sans:\"Open Sans\", -apple-system, BlinkMacSystemFont, Helvetica, Arial, sans-serif;--clicky-font-serif:\"Merriweather\", Georgia, serif;--clicky-font-mono:\"Fira Code\", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;--clicky-text-xs:.8125rem;--clicky-text-sm:.875rem;--clicky-text-base:1rem;--clicky-text-md:1rem;--clicky-text-lg:1.25rem;--clicky-text-xl:1.5rem;--clicky-text-2xl:2rem;--clicky-text-3xl:2.5rem;--clicky-text-4xl:3.5rem;--clicky-weight-regular:400;--clicky-weight-medium:500;--clicky-weight-semibold:600;--clicky-weight-bold:700;--clicky-lh-body:1.75;--clicky-lh-tight:1.5;--clicky-lh-heading:1.2;--clicky-tracking-body:.03rem;--clicky-tracking-heading:normal;--clicky-tracking-wide:.08rem;--clicky-radius-xs:2px;--clicky-radius-sm:4px;--clicky-radius-md:6px;--clicky-radius-lg:8px;--clicky-radius-xl:12px;--clicky-radius-2xl:16px;--clicky-radius-full:9999px;--clicky-shadow-xs:0 1px 2px 0 #0614310d;--clicky-shadow-sm:0 1px 2px 0 #0614310f, 0 1px 3px 0 #0614311a;--clicky-shadow-md:0 4px 6px -1px #06143114, 0 2px 4px -2px #0614310f;--clicky-shadow-lg:0 10px 15px -3px #06143114, 0 4px 6px -4px #0614310d;--clicky-shadow-xl:0 20px 25px -5px #0614311a, 0 8px 10px -6px #0614310f;--clicky-shadow-focus:0 0 0 3px #3578e540;--clicky-space-1:4px;--clicky-space-2:8px;--clicky-space-3:12px;--clicky-space-4:16px;--clicky-space-5:20px;--clicky-space-6:24px;--clicky-space-8:32px;--clicky-space-10:40px;--clicky-space-12:48px;--clicky-space-16:64px;--clicky-space-20:80px;--clicky-space-24:96px;--clicky-ease-out:cubic-bezier(.16, 1, .3, 1);--clicky-ease-in-out:cubic-bezier(.65, 0, .35, 1);--clicky-duration-fast:.12s;--clicky-duration-base:.18s;--clicky-duration-slow:.28s}[data-theme=dark]{--clicky-fg-1:#f9fafb;--clicky-fg-2:#fff;--clicky-fg-3:#d1d5db;--clicky-fg-muted:#9ca3af;--clicky-fg-subtle:#6b7280;--clicky-bg-canvas:#030712;--clicky-bg-subtle:#111827;--clicky-bg-muted:#1f2937;--clicky-bg-inset:#374151;--clicky-border-subtle:#1f2937;--clicky-border-default:#374151;--clicky-border-strong:#4b5563;--clicky-success-50:#022c22;--clicky-success-100:#064e3b;--clicky-success-200:#065f46;--clicky-success-500:#34d399;--clicky-success-600:#6ee7b7;--clicky-success-700:#a7f3d0;--clicky-success:var(--clicky-success-500);--clicky-error-50:#4c0519;--clicky-error-100:#881337;--clicky-error-200:#9f1239;--clicky-error-500:#fb7185;--clicky-error-600:#fda4af;--clicky-error-700:#fecdd3;--clicky-error:var(--clicky-error-500);--clicky-warning-50:#451a03;--clicky-warning-100:#78350f;--clicky-warning-200:#92400e;--clicky-warning-500:#fbbf24;--clicky-warning-600:#fcd34d;--clicky-warning-700:#fde68a;--clicky-warning:var(--clicky-warning-500);--clicky-info-50:#082f49;--clicky-info-100:#0c4a6e;--clicky-info-200:#075985;--clicky-info-500:#38bdf8;--clicky-info-600:#7dd3fc;--clicky-info-700:#bae6fd;--clicky-info:var(--clicky-info-500);--clicky-hint-50:#042f2e;--clicky-hint-100:#134e4a;--clicky-hint-200:#115e59;--clicky-hint-500:#2dd4bf;--clicky-hint-600:#5eead4;--clicky-hint-700:#99f6e4;--clicky-hint:var(--clicky-hint-500);--clicky-prereq-50:#2e1065;--clicky-prereq-100:#4c1d95;--clicky-prereq-200:#5b21b6;--clicky-prereq-500:#a78bfa;--clicky-prereq-600:#c4b5fd;--clicky-prereq-700:#ddd6fe;--clicky-prereq:var(--clicky-prereq-500)}@layer clicky-tokens{:root,[data-theme]{--background:var(--clicky-bg-canvas);--foreground:var(--clicky-fg-1);--card:var(--clicky-bg-canvas);--card-foreground:var(--clicky-fg-1);--popover:var(--clicky-bg-canvas);--popover-foreground:var(--clicky-fg-1);--primary:var(--clicky-primary-blue);--primary-foreground:var(--clicky-fg-on-brand);--secondary:var(--clicky-bg-muted);--secondary-foreground:var(--clicky-fg-1);--muted:var(--clicky-bg-muted);--muted-foreground:var(--clicky-fg-muted);--placeholder:var(--clicky-fg-subtle);--accent:var(--clicky-bg-muted);--accent-foreground:var(--clicky-fg-1);--destructive:var(--clicky-error);--destructive-foreground:var(--clicky-fg-on-brand);--border:var(--clicky-border-subtle);--input:var(--clicky-border-default);--ring:var(--clicky-primary-blue);--radius:var(--clicky-radius-lg);--sidebar:#1e2b48;--sidebar-foreground:#c2cce0;--sidebar-border:#33425f;--sidebar-accent:#2c3c5e;--sidebar-accent-foreground:#fff;--sidebar-primary:#7cb0ff;--sidebar-primary-foreground:#fff;--sidebar-ring:var(--clicky-primary-blue)}:root,[data-density=comfortable]{--spacing-unit:.25rem;--control-height:2.25rem;--control-padding-x:.75rem;--font-size-base:.875rem}[data-density=compact]{--spacing-unit:.1875rem;--control-height:1.875rem;--control-padding-x:.5rem;--font-size-base:.8125rem}[data-density=spacious]{--spacing-unit:.3125rem;--control-height:2.75rem;--control-padding-x:1rem;--font-size-base:1rem}body{background-color:var(--background);color:var(--foreground);font-family:var(--font-sans)}.clicky-mdx-editor-field{--accentBase:var(--background);--accentBgSubtle:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBgSubtle:color-mix(in srgb, var(--primary) 6%, var(--background))}}.clicky-mdx-editor-field{--accentBg:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBg:color-mix(in srgb, var(--primary) 10%, var(--background))}}.clicky-mdx-editor-field{--accentBgHover:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBgHover:color-mix(in srgb, var(--primary) 14%, var(--background))}}.clicky-mdx-editor-field{--accentBgActive:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBgActive:color-mix(in srgb, var(--primary) 18%, var(--background))}}.clicky-mdx-editor-field{--accentLine:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentLine:color-mix(in srgb, var(--primary) 24%, var(--border))}}.clicky-mdx-editor-field{--accentBorder:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBorder:color-mix(in srgb, var(--primary) 34%, var(--border))}}.clicky-mdx-editor-field{--accentBorderHover:var(--primary)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--accentBorderHover:color-mix(in srgb, var(--primary) 45%, var(--border))}}.clicky-mdx-editor-field{--accentSolid:var(--primary);--accentSolidHover:var(--primary);--accentText:var(--primary);--accentTextContrast:var(--primary-foreground);--basePageBg:var(--popover);--baseBase:var(--background);--baseBgSubtle:var(--muted)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--baseBgSubtle:color-mix(in srgb, var(--muted) 55%, var(--background))}}.clicky-mdx-editor-field{--baseBg:var(--muted);--baseBgHover:var(--accent);--baseBgActive:var(--accent)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field{--baseBgActive:color-mix(in srgb, var(--accent) 80%, var(--foreground) 8%)}}.clicky-mdx-editor-field{--baseLine:var(--border);--baseBorder:var(--border);--baseBorderHover:var(--input);--baseSolid:var(--muted-foreground);--baseSolidHover:var(--foreground);--baseText:var(--muted-foreground);--baseTextContrast:var(--foreground);--font-body:var(--clicky-font-sans);--font-mono:var(--clicky-font-mono);background-color:var(--background);border:1px solid var(--input);border-radius:var(--radius);color:var(--foreground);overflow:hidden}.clicky-mdx-editor-field:focus-within{box-shadow:0 0 0 2px var(--ring)}.clicky-mdx-editor-field-readonly{background-color:var(--muted)}@supports (color:color-mix(in lab, red, red)){.clicky-mdx-editor-field-readonly{background-color:color-mix(in srgb, var(--muted) 40%, transparent)}}.clicky-mdx-editor-field .mdxeditor-root-contenteditable{border-top:1px solid var(--border)}.clicky-mdx-editor-content{font-size:var(--font-size-base);line-height:1.55}.clicky-mdx-editor-content :where(p,ul,ol,blockquote,pre,table){margin-block:.5rem}.clicky-mdx-editor-content :where(h1,h2,h3,h4,h5,h6){color:var(--foreground);margin-block:.75rem .35rem;font-weight:600;line-height:1.25}.clicky-mdx-editor-content :where(code,pre){font-family:var(--clicky-font-mono)}}.clicky-mdx-editor-content :where(ul,ol){list-style:revert;margin-block:.5rem;padding-inline-start:1.5rem}.stacktrace-source pre.shiki{font-size:inherit;line-height:inherit;counter-reset:shiki-line var(--shiki-start,0);white-space:normal;margin:0;padding:0;background:0 0!important}.stacktrace-source pre.shiki code{white-space:normal;display:block}.stacktrace-source pre.shiki .line{counter-increment:shiki-line;white-space:pre;min-height:1lh;padding-left:3.5rem;display:block;position:relative}.stacktrace-source pre.shiki .line:before{content:counter(shiki-line);text-align:right;width:3rem;color:var(--muted-foreground);-webkit-user-select:none;user-select:none;padding-right:.5rem;position:absolute;left:0}.stacktrace-source pre.shiki.has-highlighted .line{opacity:.55}.stacktrace-source pre.shiki .line.error{color:#b91c1c;opacity:1;background-color:#ef44441a;font-weight:700}[data-theme=dark] .stacktrace-source pre.shiki .line.error{color:#f87171}.stacktrace-source pre.shiki .line.error:before{color:#b91c1c}[data-theme=dark] .stacktrace-source pre.shiki .line.error:before{color:#f87171}@property --tw-border-spacing-x{syntax:\"<length>\";inherits:false;initial-value:0}@property --tw-border-spacing-y{syntax:\"<length>\";inherits:false;initial-value:0}@property --tw-translate-x{syntax:\"*\";inherits:false;initial-value:0}@property --tw-translate-y{syntax:\"*\";inherits:false;initial-value:0}@property --tw-translate-z{syntax:\"*\";inherits:false;initial-value:0}@property --tw-scale-x{syntax:\"*\";inherits:false;initial-value:1}@property --tw-scale-y{syntax:\"*\";inherits:false;initial-value:1}@property --tw-scale-z{syntax:\"*\";inherits:false;initial-value:1}@property --tw-rotate-x{syntax:\"*\";inherits:false}@property --tw-rotate-y{syntax:\"*\";inherits:false}@property --tw-rotate-z{syntax:\"*\";inherits:false}@property --tw-skew-x{syntax:\"*\";inherits:false}@property --tw-skew-y{syntax:\"*\";inherits:false}@property --tw-divide-x-reverse{syntax:\"*\";inherits:false;initial-value:0}@property --tw-divide-y-reverse{syntax:\"*\";inherits:false;initial-value:0}@property --tw-gradient-position{syntax:\"*\";inherits:false}@property --tw-gradient-from{syntax:\"<color>\";inherits:false;initial-value:#0000}@property --tw-gradient-via{syntax:\"<color>\";inherits:false;initial-value:#0000}@property --tw-gradient-to{syntax:\"<color>\";inherits:false;initial-value:#0000}@property --tw-gradient-stops{syntax:\"*\";inherits:false}@property --tw-gradient-via-stops{syntax:\"*\";inherits:false}@property --tw-gradient-from-position{syntax:\"<length-percentage>\";inherits:false;initial-value:0%}@property --tw-gradient-via-position{syntax:\"<length-percentage>\";inherits:false;initial-value:50%}@property --tw-gradient-to-position{syntax:\"<length-percentage>\";inherits:false;initial-value:100%}@property --tw-shadow{syntax:\"*\";inherits:false;initial-value:0 0 #0000}@property --tw-shadow-color{syntax:\"*\";inherits:false}@property --tw-shadow-alpha{syntax:\"<percentage>\";inherits:false;initial-value:100%}@property --tw-inset-shadow{syntax:\"*\";inherits:false;initial-value:0 0 #0000}@property --tw-inset-shadow-color{syntax:\"*\";inherits:false}@property --tw-inset-shadow-alpha{syntax:\"<percentage>\";inherits:false;initial-value:100%}@property --tw-ring-color{syntax:\"*\";inherits:false}@property --tw-ring-shadow{syntax:\"*\";inherits:false;initial-value:0 0 #0000}@property --tw-inset-ring-color{syntax:\"*\";inherits:false}@property --tw-inset-ring-shadow{syntax:\"*\";inherits:false;initial-value:0 0 #0000}@property --tw-ring-inset{syntax:\"*\";inherits:false}@property --tw-ring-offset-width{syntax:\"<length>\";inherits:false;initial-value:0}@property --tw-ring-offset-color{syntax:\"*\";inherits:false;initial-value:#fff}@property --tw-ring-offset-shadow{syntax:\"*\";inherits:false;initial-value:0 0 #0000}@property --tw-backdrop-blur{syntax:\"*\";inherits:false}@property --tw-backdrop-brightness{syntax:\"*\";inherits:false}@property --tw-backdrop-contrast{syntax:\"*\";inherits:false}@property --tw-backdrop-grayscale{syntax:\"*\";inherits:false}@property --tw-backdrop-hue-rotate{syntax:\"*\";inherits:false}@property --tw-backdrop-invert{syntax:\"*\";inherits:false}@property --tw-backdrop-opacity{syntax:\"*\";inherits:false}@property --tw-backdrop-saturate{syntax:\"*\";inherits:false}@property --tw-backdrop-sepia{syntax:\"*\";inherits:false}@property --tw-duration{syntax:\"*\";inherits:false}@property --tw-ease{syntax:\"*\";inherits:false}@property --tw-contain-size{syntax:\"*\";inherits:false}@property --tw-contain-layout{syntax:\"*\";inherits:false}@property --tw-contain-paint{syntax:\"*\";inherits:false}@property --tw-contain-style{syntax:\"*\";inherits:false}@property --tw-content{syntax:\"*\";inherits:false;initial-value:\"\"}@keyframes spin{to{transform:rotate(360deg)}}@keyframes pulse{50%{opacity:.5}}html,body,#root{height:100%}@property --tw-space-y-reverse{syntax:\"*\";inherits:false;initial-value:0}@property --tw-border-style{syntax:\"*\";inherits:false;initial-value:solid}@property --tw-leading{syntax:\"*\";inherits:false}@property --tw-font-weight{syntax:\"*\";inherits:false}@property --tw-tracking{syntax:\"*\";inherits:false}@property --tw-ordinal{syntax:\"*\";inherits:false}@property --tw-slashed-zero{syntax:\"*\";inherits:false}@property --tw-numeric-figure{syntax:\"*\";inherits:false}@property --tw-numeric-spacing{syntax:\"*\";inherits:false}@property --tw-numeric-fraction{syntax:\"*\";inherits:false}@property --tw-outline-style{syntax:\"*\";inherits:false;initial-value:solid}@property --tw-blur{syntax:\"*\";inherits:false}@property --tw-brightness{syntax:\"*\";inherits:false}@property --tw-contrast{syntax:\"*\";inherits:false}@property --tw-grayscale{syntax:\"*\";inherits:false}@property --tw-hue-rotate{syntax:\"*\";inherits:false}@property --tw-invert{syntax:\"*\";inherits:false}@property --tw-opacity{syntax:\"*\";inherits:false}@property --tw-saturate{syntax:\"*\";inherits:false}@property --tw-sepia{syntax:\"*\";inherits:false}@property --tw-drop-shadow{syntax:\"*\";inherits:false}@property --tw-drop-shadow-color{syntax:\"*\";inherits:false}@property --tw-drop-shadow-alpha{syntax:\"<percentage>\";inherits:false;initial-value:100%}@property --tw-drop-shadow-size{syntax:\"*\";inherits:false}", j = /* @__PURE__ */ e(n(), 1), M = r(), $t = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("path", {
		fill: "currentColor",
		d: "m13.236 9.307l-4.429 2.491a1.5 1.5 0 0 0-.781-.704l.01-.01l4.709-2.649a.5.5 0 0 0 0-.872l-8-4.499A.5.5 0 0 0 4 3.5V7c-.354 0-.684.081-1 .194V3.5a1.5 1.5 0 0 1 2.235-1.307l8 4.5c1.019.573 1.019 2.041 0 2.615zM7 10.707V12h.5a.5.5 0 0 1 0 1H7c0 .384-.073.75-.205 1.088l1.059 1.059a.5.5 0 0 1-.708.707l-.885-.885a2.995 2.995 0 0 1-4.522 0l-.885.885a.5.5 0 0 1-.708 0a.5.5 0 0 1 0-.707l1.059-1.059A3 3 0 0 1 1 13H.5a.5.5 0 0 1 0-1H1v-1.293l-.854-.854a.5.5 0 0 1 .707-.707l.854.854H2c0-1.103.897-2 2-2s2 .897 2 2h.293l.854-.854a.5.5 0 0 1 .707.707zM3 10h2a1 1 0 0 0-2 0m3 1H2v2c0 1.103.897 2 2 2s2-.897 2-2z"
	})
}), {
	__source: "codicon:debug-alt-small",
	__viewBox: "0 0 16 16",
	__group: "ui-controls",
	__consumerName: "debug-small",
	displayName: "UiDebugSmall"
}), en = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiPlay.tsx
var tn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("g", {
		fill: "none",
		children: /* @__PURE__ */ (0, M.jsx)("path", {
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
}), nn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 16 16",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsxs)("g", {
		fill: "none",
		children: [/* @__PURE__ */ (0, M.jsx)("rect", {
			x: "3.5",
			y: "2.5",
			width: "3",
			height: "11",
			rx: "0.5",
			stroke: "#6C707E"
		}), /* @__PURE__ */ (0, M.jsx)("rect", {
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
}), rn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiTag.tsx
var an = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiServerProcess.tsx
var on = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/icons/components/UiChartLineUp.tsx
var sn = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, M.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, M.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/data/lint/actions.ts
var cn = "Add {source} to .gavel.yaml";
function ln(e) {
	if (!e) return "**";
	let t = e.length;
	for (; t > 0 && e[t - 1] === "/";) --t;
	let n = e.slice(0, t);
	return n ? `${n}/**` : "**";
}
function un(e, t) {
	let n = e.target_path ?? "", r = /* @__PURE__ */ new Map();
	for (let i of t) if (!(e.work_dir && i.work_dir && i.work_dir !== e.work_dir)) for (let e of i.violations ?? []) {
		let t = vt(e.file, i.work_dir);
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
function dn(e, t) {
	return {
		key: "linter-disable",
		label: `Disable ${e} entirely`,
		title: cn,
		req: {
			source: e,
			work_dir: t
		},
		variant: "subtle"
	};
}
function fn(e, t) {
	let n = ln(e.target_path), r = e.linterName ?? "", i = e.ruleName ?? "", a = r ? [] : un(e, t), o = new Set(a.flatMap((e) => e.workDir ? [e.workDir] : [])), s = e.work_dir || (o.size === 1 ? Array.from(o)[0] : "");
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
		dn(r, s)
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
		dn(r, s)
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
		...dn(e.linter, s),
		key: `linter-disable:${e.linter}`
	}])];
}
function pn(e) {
	let t = (e.children ?? []).flatMap((e) => e.kind === "linter" && e.linterName ? [e.linterName] : []);
	return Array.from(new Set(t));
}
function mn(e) {
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
	}, dn(t, r)] : [{
		key: "file-everything",
		label: "Ignore everything in this file",
		title: "Add {file} to .gavel.yaml",
		req: {
			file: n,
			work_dir: r
		}
	}, ...pn(e).flatMap((e) => [{
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
		...dn(e, r),
		key: `linter-disable:${e}`
	}])];
}
function hn(e) {
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
		dn(t, r)
	];
}
function gn(e) {
	let t = e.work_dir ?? "";
	return e.ruleName ? hn(e) : e.linterName ? [dn(e.linterName, t)] : pn(e).map((e) => ({
		...dn(e, t),
		key: `linter-disable:${e}`
	}));
}
function _n(e, t) {
	switch (e.kind) {
		case "linter": return e.linterName ? [dn(e.linterName, e.work_dir || e.linter?.work_dir || "")] : [];
		case "lint-folder": return fn(e, t);
		case "lint-file": return mn(e);
		case "lint-rule":
		case "lint-rule-group": return hn(e);
		case "lint-target": return gn(e);
		case "linter-error": return [];
	}
}
function vn({ violation: e, linter: t, file: n, workDir: r }) {
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
// ../../../clicky-ui/packages/ui/src/data/lint/ignore-menu-items.ts
function yn(e, t) {
	return t || !!e.disabledWithoutWorkDir && !e.req.work_dir;
}
function bn(e, t, n) {
	return e.map((e) => ({
		label: e.label,
		icon: Ne,
		title: e.title,
		disabled: yn(e, n),
		onSelect: () => void t(e)
	}));
}
// ../../../clicky-ui/packages/ui/src/data/lint/violation-menu-items.ts
function xn({ violation: e, linter: t, file: n, workDir: r, onIgnore: i, ignoreBusy: a, onOpenFile: o }) {
	let s = i ? vn({
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
		...i ? bn(s, i, !!a) : []
	];
}
var Sn = /* @__PURE__ */ new WeakMap();
function Cn(e, t, n) {
	let r = Sn.get(e) ?? /* @__PURE__ */ new Map();
	Sn.set(e, r);
	let i = `${n}\u0000${t}`, a = r.get(i);
	return a || (a = e({
		file: t,
		...n ? { workDir: n } : {}
	}), r.set(i, a), a.catch(() => r.delete(i))), a;
}
function wn(e, t, n) {
	let [r, i] = (0, j.useState)(null);
	return (0, j.useEffect)(() => {
		if (!e || !t) {
			i(null);
			return;
		}
		let r = !1;
		return i({ status: "loading" }), Cn(e, t, n).then((e) => !r && i({
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
function Tn({ text: e }) {
	return e ? /* @__PURE__ */ (0, M.jsx)("div", {
		className: "mt-0.5 whitespace-pre-wrap text-sm",
		children: e
	}) : null;
}
function En({ violation: e, file: t, workDir: n, loadSource: r, showMessage: i = !0 }) {
	let a = e.line ?? 0, o = wn(a > 0 ? r : void 0, t, n);
	if (!o || o.status !== "ready") return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
		i && /* @__PURE__ */ (0, M.jsx)(Tn, { text: e.message }),
		o?.status === "loading" && /* @__PURE__ */ (0, M.jsx)("div", {
			className: "mt-1 text-xs text-muted-foreground",
			children: "Loading source…"
		}),
		o?.status === "error" && /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "mt-1 text-xs text-red-700 dark:text-red-300",
			role: "alert",
			children: [
				"Could not load ",
				t,
				": ",
				o.message
			]
		}),
		!o && e.code && /* @__PURE__ */ (0, M.jsx)("pre", {
			className: "mt-1 overflow-x-auto rounded bg-muted p-2 font-mono text-xs",
			children: e.code
		})
	] });
	let s = o.text.split("\n"), c = Math.max(1, a - 3), l = Math.min(s.length, Math.max(a, e.end_line ?? a) + 3), u = _e(t);
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
		i && /* @__PURE__ */ (0, M.jsx)(Tn, { text: e.message }),
		a > s.length && /* @__PURE__ */ (0, M.jsx)("p", {
			role: "alert",
			className: "text-xs text-destructive",
			children: "Source location is outside this file."
		}),
		/* @__PURE__ */ (0, M.jsx)(le, {
			focalTone: jt(e),
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
// ../../../clicky-ui/packages/ui/src/data/lint/LintIgnoreMenu.tsx
function Dn({ actions: e, onIgnore: t, busy: n, align: r = "left" }) {
	return /* @__PURE__ */ (0, M.jsx)(ye, {
		label: "Actions",
		variant: "ghost",
		align: r,
		menuLabel: "Finding actions",
		items: bn(e, t, n)
	});
}
function On({ actions: e, onIgnore: t, busy: n, label: r }) {
	let [i, a] = (0, j.useState)(null), o = (0, j.useCallback)((e) => a(e?.parentElement ?? null), []);
	if (e.length === 0) return null;
	let s = bn(e, t, n), c = `Ignore options for ${r}`;
	return /* @__PURE__ */ (0, M.jsxs)("span", {
		ref: o,
		className: "flex shrink-0 opacity-0 focus-within:opacity-100 [:hover>&]:opacity-100 [[aria-selected=true]>div>&]:opacity-100",
		onClick: (e) => e.stopPropagation(),
		onKeyDown: (e) => e.stopPropagation(),
		"data-lint-row-menu": "",
		children: [/* @__PURE__ */ (0, M.jsx)(ye, {
			items: s,
			align: "right",
			menuLabel: c,
			trigger: /* @__PURE__ */ (0, M.jsx)(fe, {
				type: "button",
				variant: "ghost",
				size: "icon",
				className: "h-5 w-5 text-muted-foreground",
				"aria-label": c,
				title: "Ignore options",
				children: /* @__PURE__ */ (0, M.jsx)(w, { icon: A })
			})
		}), /* @__PURE__ */ (0, M.jsx)(Wt, {
			contextTarget: i,
			menuLabel: c,
			menuItems: s
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/lint/types.ts
function kn(e, t = /* @__PURE__ */ new Date()) {
	let n = e?.cache;
	if (n?.hit) return k(new Date(n.stored_at), t);
}
var An = [
	"linter-rule-file",
	"linter-file",
	"file-linter-rule",
	"summary"
];
// ../../../clicky-ui/packages/ui/src/data/lint/LintDetail.tsx
function jn({ title: e, children: t }) {
	return /* @__PURE__ */ (0, M.jsxs)("section", {
		className: "space-y-1.5",
		children: [/* @__PURE__ */ (0, M.jsx)("h3", {
			className: "text-[10px] uppercase tracking-wide text-muted-foreground",
			children: e
		}), t]
	});
}
function Mn({ children: e, className: t }) {
	return /* @__PURE__ */ (0, M.jsx)("span", {
		className: C("break-all font-mono text-sm", t),
		children: e
	});
}
function Nn({ children: e, tone: t }) {
	return /* @__PURE__ */ (0, M.jsx)("pre", {
		className: C("max-h-80 overflow-auto whitespace-pre-wrap break-words rounded bg-muted p-3 font-mono text-xs", t === "error" && "bg-red-500/10 text-red-700 dark:text-red-300"),
		children: e
	});
}
function Pn({ violation: e, file: t, workDir: n, loadSource: r, onOpenFile: i, entityHref: a, showFile: o, showMessage: s = !0 }) {
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "space-y-2",
		"data-lint-violation-detail": "",
		children: [
			o && /* @__PURE__ */ (0, M.jsx)("p", {
				className: "break-all font-mono text-xs text-muted-foreground",
				children: St(e, !0)
			}),
			/* @__PURE__ */ (0, M.jsx)(En, {
				violation: e,
				file: t,
				workDir: n,
				loadSource: r,
				showMessage: s
			}),
			ft(e) && /* @__PURE__ */ (0, M.jsx)(ae, {
				className: "mt-1",
				locations: e.locations,
				workDir: n,
				onOpenFile: i,
				entityHref: a
			})
		]
	});
}
function Fn(e) {
	return /* @__PURE__ */ (0, M.jsx)("div", {
		"data-lint-violation": "",
		children: /* @__PURE__ */ (0, M.jsx)(Vt, {
			...e.expanded === void 0 ? {} : { expanded: e.expanded },
			...e.onExpandedChange ? { onExpandedChange: e.onExpandedChange } : {},
			message: e.violation.message ?? "",
			severity: jt(e.violation),
			metadata: `${e.linter} ${e.violation.rule?.method ?? ""}`,
			location: St(e.violation, !!e.showFile),
			actions: xn(e),
			children: /* @__PURE__ */ (0, M.jsx)(Pn, {
				...e,
				showMessage: !1
			})
		})
	});
}
function In(e) {
	return e === "" ? "''" : /^[A-Za-z0-9_./:=@%+,-]+$/.test(e) ? e : `'${e.replace(/'/g, "'\\''")}'`;
}
function Ln(e) {
	return e.command ? [e.command, ...e.args ?? []].map(In).join(" ") : "";
}
function Rn(e) {
	return e.status === "pending" ? ["pending", "text-muted-foreground"] : e.status === "running" ? ["running", "text-blue-600 dark:text-blue-400"] : e.skipped ? ["skipped", "text-muted-foreground"] : e.timed_out ? ["timed out", "text-amber-600 dark:text-amber-400"] : Lt(e) ? ["failed", "text-red-600 dark:text-red-400"] : ["ran", "text-emerald-600 dark:text-emerald-400"];
}
function zn({ result: e }) {
	let [t, n] = Rn(e), r = (e.violations ?? []).length, i = kn(e);
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex flex-wrap gap-3 text-sm",
		children: [
			/* @__PURE__ */ (0, M.jsx)("span", {
				className: n,
				children: t
			}),
			i && e.cache && /* @__PURE__ */ (0, M.jsxs)("span", {
				className: "text-muted-foreground",
				title: `${e.cache.source} fingerprint ${e.cache.fingerprint ?? "unknown"}`,
				children: ["cached · linted ", i]
			}),
			/* @__PURE__ */ (0, M.jsxs)("span", {
				className: "text-muted-foreground",
				children: [
					r,
					" violation",
					r === 1 ? "" : "s"
				]
			}),
			e.file_count !== void 0 && /* @__PURE__ */ (0, M.jsxs)("span", {
				className: "text-muted-foreground",
				children: [e.file_count, " files"]
			}),
			e.rule_count !== void 0 && /* @__PURE__ */ (0, M.jsxs)("span", {
				className: "text-muted-foreground",
				children: [e.rule_count, " rules"]
			}),
			e.duration ? /* @__PURE__ */ (0, M.jsx)("span", {
				className: "text-muted-foreground",
				children: tt(e.duration)
			}) : null
		]
	});
}
function Bn({ label: e, value: t }) {
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex items-center gap-1.5 text-xs",
		children: [
			/* @__PURE__ */ (0, M.jsx)("span", {
				className: "text-muted-foreground",
				children: e
			}),
			/* @__PURE__ */ (0, M.jsx)("span", {
				className: "break-all font-mono text-xs",
				children: t
			}),
			/* @__PURE__ */ (0, M.jsx)(ee, {
				value: t,
				label: `Copy ${e}`
			})
		]
	});
}
function Vn({ result: e }) {
	return !e.run_id && !e.session_id ? null : /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "space-y-0.5",
		children: [e.run_id && /* @__PURE__ */ (0, M.jsx)(Bn, {
			label: "run id",
			value: e.run_id
		}), e.session_id && /* @__PURE__ */ (0, M.jsx)(Bn, {
			label: "session id",
			value: e.session_id
		})]
	});
}
function Hn({ result: e, workDir: t }) {
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, M.jsxs)(jn, {
		title: "Status",
		children: [/* @__PURE__ */ (0, M.jsx)(zn, { result: e }), /* @__PURE__ */ (0, M.jsx)(Vn, { result: e })]
	}), e.command && /* @__PURE__ */ (0, M.jsxs)(jn, {
		title: "Command",
		children: [/* @__PURE__ */ (0, M.jsx)(Nn, { children: Ln(e) }), t && /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "font-mono text-xs text-muted-foreground",
			children: ["cwd: ", t]
		})]
	})] });
}
function Un({ node: e, renderRun: t }) {
	let n = e.linter, r = e.work_dir ?? n?.work_dir ?? "", i = n && t?.(n);
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
		n && /* @__PURE__ */ (0, M.jsx)(Hn, {
			result: n,
			workDir: r
		}),
		i && /* @__PURE__ */ (0, M.jsx)(jn, {
			title: "Run",
			children: i
		}),
		n?.error && /* @__PURE__ */ (0, M.jsx)(jn, {
			title: "Error",
			children: /* @__PURE__ */ (0, M.jsx)(Nn, {
				tone: "error",
				children: n.error
			})
		}),
		n?.raw_output && /* @__PURE__ */ (0, M.jsx)(jn, {
			title: "Raw output",
			children: /* @__PURE__ */ (0, M.jsx)(Nn, { children: n.raw_output })
		})
	] });
}
function Wn({ node: e }) {
	let t = e.linter;
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
		/* @__PURE__ */ (0, M.jsx)(jn, {
			title: "Error",
			children: /* @__PURE__ */ (0, M.jsx)(Nn, {
				tone: "error",
				children: e.message
			})
		}),
		t && /* @__PURE__ */ (0, M.jsx)(Hn, {
			result: t,
			workDir: t.work_dir ?? ""
		}),
		t?.raw_output && /* @__PURE__ */ (0, M.jsx)(jn, {
			title: "Raw output",
			children: /* @__PURE__ */ (0, M.jsx)(Nn, { children: t.raw_output })
		})
	] });
}
function Gn({ node: e, results: t }) {
	let n = e.ruleName ? [] : un(e, t).filter((t) => !e.linterName || t.linter === e.linterName);
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, M.jsx)(jn, {
		title: "Folder",
		children: /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "flex items-center gap-2",
			children: [/* @__PURE__ */ (0, M.jsx)(Mn, { children: e.file || e.name }), e.target_path !== void 0 && /* @__PURE__ */ (0, M.jsx)("span", {
				className: "text-xs text-muted-foreground",
				children: ln(e.target_path)
			})]
		})
	}), e.ruleName ? /* @__PURE__ */ (0, M.jsx)(jn, {
		title: "Scope",
		children: /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "space-y-1 text-sm",
			children: [/* @__PURE__ */ (0, M.jsxs)("div", { children: [
				/* @__PURE__ */ (0, M.jsx)("span", {
					className: "text-muted-foreground",
					children: "Linter:"
				}),
				" ",
				/* @__PURE__ */ (0, M.jsx)(Mn, { children: e.linterName })
			] }), /* @__PURE__ */ (0, M.jsxs)("div", { children: [
				/* @__PURE__ */ (0, M.jsx)("span", {
					className: "text-muted-foreground",
					children: "Rule:"
				}),
				" ",
				/* @__PURE__ */ (0, M.jsx)(Mn, { children: e.ruleName })
			] })]
		})
	}) : /* @__PURE__ */ (0, M.jsx)(jn, {
		title: "Linters",
		children: /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "space-y-1",
			children: [n.map(({ linter: e, count: t }) => /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex items-center justify-between text-sm",
				children: [/* @__PURE__ */ (0, M.jsx)(Mn, { children: e }), /* @__PURE__ */ (0, M.jsxs)("span", {
					className: "text-xs text-muted-foreground",
					children: [t, " violations"]
				})]
			}, e)), n.length === 0 && /* @__PURE__ */ (0, M.jsx)("div", {
				className: "text-sm text-muted-foreground",
				children: "No violations found under this folder."
			})]
		})
	})] });
}
function Kn({ node: e, ...t }) {
	let n = e.violations ?? [];
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, M.jsx)(jn, {
		title: "File",
		children: /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "flex items-center gap-2",
			children: [/* @__PURE__ */ (0, M.jsx)(Mn, { children: e.file }), n.length > 0 && /* @__PURE__ */ (0, M.jsxs)("span", {
				className: "text-xs text-muted-foreground",
				children: [
					"(",
					n.length,
					" violations)"
				]
			})]
		})
	}), n.length > 0 && /* @__PURE__ */ (0, M.jsx)(jn, {
		title: "Violations",
		children: /* @__PURE__ */ (0, M.jsx)("div", {
			className: "space-y-2",
			children: n.map((n, r) => /* @__PURE__ */ (0, M.jsx)(Fn, {
				violation: n,
				linter: e.linterName ?? "",
				file: e.target_path || e.file || "",
				workDir: e.work_dir ?? "",
				...t
			}, r))
		})
	})] });
}
function qn({ node: e, ...t }) {
	let n = e.violations ?? [], r = e.linterName ?? "", i = e.work_dir ?? "";
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
		/* @__PURE__ */ (0, M.jsxs)(jn, {
			title: "Rule",
			children: [/* @__PURE__ */ (0, M.jsx)(Mn, { children: e.ruleName }), /* @__PURE__ */ (0, M.jsxs)("span", {
				className: "ml-2 text-xs text-muted-foreground",
				children: [
					"(",
					Pt(e),
					" violations)"
				]
			})]
		}),
		e.file && /* @__PURE__ */ (0, M.jsx)(jn, {
			title: "File",
			children: /* @__PURE__ */ (0, M.jsx)(Mn, { children: e.file })
		}),
		n.length > 0 && /* @__PURE__ */ (0, M.jsx)(jn, {
			title: "Violations",
			children: /* @__PURE__ */ (0, M.jsx)("div", {
				className: "space-y-2",
				children: n.map((n, a) => /* @__PURE__ */ (0, M.jsx)(Fn, {
					violation: n,
					linter: r,
					file: n.raw_file || e.target_path || e.file || "",
					workDir: i,
					showFile: !e.file,
					...t
				}, a))
			})
		})
	] });
}
var Jn = {
	entity: "Entity kind",
	url: "Host",
	other: "Target"
};
function Yn({ node: e, ...t }) {
	let n = e.violations ?? [];
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, M.jsxs)(jn, {
		title: Jn[e.targetKind ?? "other"],
		children: [/* @__PURE__ */ (0, M.jsx)(Mn, { children: e.name }), /* @__PURE__ */ (0, M.jsxs)("span", {
			className: "ml-2 text-xs text-muted-foreground",
			children: [
				"(",
				Pt(e),
				" violations)"
			]
		})]
	}), n.length > 0 && /* @__PURE__ */ (0, M.jsx)(jn, {
		title: "Violations",
		children: /* @__PURE__ */ (0, M.jsx)("div", {
			className: "space-y-2",
			children: n.map((n, r) => /* @__PURE__ */ (0, M.jsx)(Fn, {
				violation: n,
				linter: e.linterName ?? "",
				file: "",
				workDir: e.work_dir ?? "",
				showFile: !0,
				...t
			}, r))
		})
	})] });
}
var Xn = {
	linter: "Linter",
	"lint-folder": "Folder",
	"lint-file": "File",
	"lint-rule": "Rule",
	"lint-rule-group": "Rule",
	"lint-target": "Target",
	"linter-error": "Linter error"
};
function Zn({ node: e, actions: t, onIgnore: n, ignoreBusy: r }) {
	let i = e.kind === "linter-error" ? 0 : Pt(e);
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex min-w-0 items-center gap-2 border-b border-border px-density-4 py-density-2",
		"data-detail-header": "",
		children: [
			e.kind === "linter" ? /* @__PURE__ */ (0, M.jsx)(yt, {
				tool: Mt(e),
				className: "text-lg"
			}) : e.kind === "linter-error" ? /* @__PURE__ */ (0, M.jsx)(xt, {
				severity: "error",
				className: "text-lg"
			}) : null,
			/* @__PURE__ */ (0, M.jsx)("h2", {
				className: "min-w-0 shrink truncate text-base font-semibold",
				title: e.name,
				children: e.name
			}),
			/* @__PURE__ */ (0, M.jsx)("span", {
				className: "inline-flex shrink-0 items-center rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground",
				children: Xn[e.kind]
			}),
			i > 0 && /* @__PURE__ */ (0, M.jsxs)("span", {
				className: "shrink-0 text-xs text-muted-foreground",
				children: [
					i,
					" violation",
					i === 1 ? "" : "s"
				]
			}),
			n && t.length > 0 && /* @__PURE__ */ (0, M.jsx)("div", {
				className: "ml-auto shrink-0",
				"data-lint-node-actions": "",
				children: /* @__PURE__ */ (0, M.jsx)(Dn, {
					actions: t,
					onIgnore: n,
					busy: !!r,
					align: "right"
				})
			})
		]
	});
}
function Qn({ node: e, results: t, onIgnore: n, ignoreBusy: r, onOpenFile: i, loadSource: a, entityHref: o, renderRun: s }) {
	if (!e) return /* @__PURE__ */ (0, M.jsx)("div", {
		className: "flex h-full items-center justify-center text-sm text-muted-foreground",
		children: /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "text-center",
			children: [/* @__PURE__ */ (0, M.jsx)(w, {
				icon: Ce,
				className: "mb-2 block text-4xl"
			}), "Select a linter, rule or file to view its findings."]
		})
	});
	let c = {
		onIgnore: n,
		ignoreBusy: r,
		onOpenFile: i,
		loadSource: a,
		entityHref: o
	};
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		children: [/* @__PURE__ */ (0, M.jsx)(Zn, {
			node: e,
			actions: _n(e, t),
			onIgnore: n,
			ignoreBusy: r
		}), /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "min-h-0 flex-1 space-y-4 overflow-y-auto p-density-4",
			children: [
				e.kind === "linter" && /* @__PURE__ */ (0, M.jsx)(Un, {
					node: e,
					renderRun: s
				}),
				e.kind === "linter-error" && /* @__PURE__ */ (0, M.jsx)(Wn, { node: e }),
				e.kind === "lint-folder" && /* @__PURE__ */ (0, M.jsx)(Gn, {
					node: e,
					results: t
				}),
				e.kind === "lint-file" && /* @__PURE__ */ (0, M.jsx)(Kn, {
					node: e,
					...c
				}),
				(e.kind === "lint-rule" || e.kind === "lint-rule-group") && /* @__PURE__ */ (0, M.jsx)(qn, {
					node: e,
					...c
				}),
				e.kind === "lint-target" && /* @__PURE__ */ (0, M.jsx)(Yn, {
					node: e,
					...c
				})
			]
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/lint/LintTable.tsx
var $n = {
	error: 0,
	warning: 1,
	info: 2
};
function er(e, t) {
	let n = Nt(e, {
		severity: /* @__PURE__ */ new Map(),
		linter: /* @__PURE__ */ new Map()
	}), r = It(n.map((e) => [
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
		severity: jt(e.v),
		linter: e.linter,
		rule: e.v.rule?.method ?? "",
		location: e.file ? `${e.file}${e.v.line ? `:${e.v.line}` : ""}` : St(e.v, !0),
		message: e.v.message ?? "",
		flat: e
	})).filter(({ flat: e }) => it(e.linter, t.linter) && it(jt(e.v), t.severity));
}
var tr = [
	{
		key: "severity",
		label: "Severity",
		sortable: !0,
		sortValue: (e) => $n[e],
		shrink: !0,
		render: (e) => /* @__PURE__ */ (0, M.jsx)("span", {
			className: "inline-flex items-center gap-1",
			children: /* @__PURE__ */ (0, M.jsx)(xt, { severity: e })
		})
	},
	{
		key: "linter",
		label: "Linter",
		sortable: !0,
		shrink: !0,
		render: (e) => /* @__PURE__ */ (0, M.jsxs)("span", {
			className: "inline-flex items-center gap-1",
			children: [/* @__PURE__ */ (0, M.jsx)(yt, { tool: e }), mt(e)?.label ?? e]
		})
	},
	{
		key: "rule",
		label: "Rule",
		sortable: !0,
		render: (e) => /* @__PURE__ */ (0, M.jsx)("span", {
			className: "font-mono text-xs",
			children: e
		})
	},
	{
		key: "location",
		label: "Location",
		sortable: !0,
		render: (e) => /* @__PURE__ */ (0, M.jsx)("span", {
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
function nr({ results: e, filters: t, className: n, ...r }) {
	let { ref: i, wider: a } = lt(640), o = (0, j.useMemo)(() => [...tr, {
		key: "actions",
		label: "",
		shrink: !0,
		render: (e, t) => /* @__PURE__ */ (0, M.jsx)(Gt, { items: xn({
			violation: t.flat.v,
			linter: t.flat.linter,
			file: t.flat.rawFile ?? "",
			workDir: t.flat.workDir ?? "",
			...r
		}) })
	}], [r]), s = (0, j.useMemo)(() => er(e, {
		severity: /* @__PURE__ */ new Map(),
		linter: /* @__PURE__ */ new Map()
	}).map((e) => e.id), [e]), [c, l] = Ut(s), u = (0, j.useMemo)(() => er(e, t), [e, t]);
	return /* @__PURE__ */ (0, M.jsx)("div", {
		ref: i,
		className: "min-w-0",
		children: a ? /* @__PURE__ */ (0, M.jsx)(g, {
			data: u,
			columns: o,
			getRowId: (e) => e.id,
			defaultSort: {
				key: "severity",
				dir: "asc"
			},
			emptyMessage: "No violations match the current filters",
			detailStyle: "row",
			expandedRowIds: c,
			onExpandedRowIdsChange: l,
			renderExpandedRow: ({ flat: e }) => /* @__PURE__ */ (0, M.jsx)(Pn, {
				violation: e.v,
				linter: e.linter,
				file: e.rawFile ?? "",
				workDir: e.workDir ?? "",
				...r
			}),
			...n ? { className: n } : {}
		}) : /* @__PURE__ */ (0, M.jsx)("div", { children: u.slice().sort((e, t) => $n[e.severity] - $n[t.severity]).map(({ id: e, flat: t }) => /* @__PURE__ */ (0, M.jsx)(Fn, {
			expanded: c.includes(e),
			onExpandedChange: (t) => l((n) => t ? [...n, e] : n.filter((t) => t !== e)),
			violation: t.v,
			linter: t.linter,
			file: t.rawFile ?? "",
			workDir: t.workDir ?? "",
			showFile: !0,
			...r
		}, e)) })
	});
}
function rr(e, t) {
	let n = Nt(e, t), r = It(n.map((e) => [
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
	])), i = /* @__PURE__ */ new Map();
	return n.forEach((e, t) => {
		let n = i.get(e.linter) ?? /* @__PURE__ */ new Map();
		i.set(e.linter, n);
		let a = e.file ?? "";
		n.set(a, [...n.get(a) ?? [], {
			id: r[t],
			flat: e
		}]);
	}), [...i].map(([e, t]) => ({
		linter: e,
		count: [...t.values()].reduce((e, t) => e + t.length, 0),
		files: [...t].sort(([e], [t]) => e.localeCompare(t)).map(([e, t]) => ({
			file: e,
			violations: t.slice().sort((e, t) => (e.flat.v.line ?? 0) - (t.flat.v.line ?? 0))
		}))
	}));
}
// ../../../clicky-ui/packages/ui/src/data/lint/LintList.tsx
function ir(e) {
	let { line: t, column: n } = e.v;
	return t ? n ? `:${t}:${n}` : `:${t}` : "";
}
function ar({ file: e }) {
	let t = e.lastIndexOf("/") + 1;
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [t > 0 && /* @__PURE__ */ (0, M.jsx)("span", {
		className: "text-muted-foreground",
		children: e.slice(0, t)
	}), /* @__PURE__ */ (0, M.jsx)("span", {
		className: "text-blue-600 dark:text-blue-400",
		children: e.slice(t)
	})] });
}
function or({ results: e, filters: t, className: n, ...r }) {
	let i = (0, j.useMemo)(() => rr(e, {
		severity: /* @__PURE__ */ new Map(),
		linter: /* @__PURE__ */ new Map()
	}).flatMap((e) => e.files.flatMap((e) => e.violations.map((e) => e.id))), [e]), [a, o] = Ut(i), s = (0, j.useMemo)(() => rr(e, t), [e, t]), c = (e) => o((t) => t.includes(e) ? t.filter((t) => t !== e) : [...t, e]);
	return s.length === 0 ? /* @__PURE__ */ (0, M.jsx)("div", {
		className: "p-6 text-center text-sm text-muted-foreground",
		children: "No violations match the current filters"
	}) : /* @__PURE__ */ (0, M.jsx)("ul", {
		className: C("space-y-2 p-density-3 font-mono text-xs", n),
		"data-lint-list": "",
		children: s.map((e) => /* @__PURE__ */ (0, M.jsxs)("li", {
			"data-lint-list-linter": e.linter,
			children: [/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex items-center gap-1.5",
				children: [
					/* @__PURE__ */ (0, M.jsx)(yt, { tool: e.linter }),
					/* @__PURE__ */ (0, M.jsx)("span", {
						className: "font-semibold text-foreground",
						children: mt(e.linter)?.label ?? e.linter
					}),
					/* @__PURE__ */ (0, M.jsxs)("span", {
						className: "text-muted-foreground",
						children: [
							"(",
							e.count,
							" violation",
							e.count === 1 ? "" : "s",
							")"
						]
					})
				]
			}), /* @__PURE__ */ (0, M.jsx)("ul", { children: e.files.map(({ file: e, violations: t }) => /* @__PURE__ */ (0, M.jsx)("li", {
				className: "mt-0.5",
				"data-lint-list-file": e,
				children: /* @__PURE__ */ (0, M.jsx)("ul", { children: t.map(({ id: t, flat: n }) => {
					let i = a.includes(t);
					return /* @__PURE__ */ (0, M.jsxs)("li", {
						"data-lint-violation": "",
						children: [/* @__PURE__ */ (0, M.jsxs)("div", {
							className: "group flex items-start gap-1 rounded hover:bg-accent/50",
							children: [/* @__PURE__ */ (0, M.jsxs)("button", {
								type: "button",
								"aria-expanded": i,
								"data-lint-line": "",
								onClick: () => c(t),
								className: "min-w-0 flex-1 py-0.5 text-left",
								children: [/* @__PURE__ */ (0, M.jsxs)("span", {
									className: "flex min-w-0 items-center gap-1.5",
									children: [
										/* @__PURE__ */ (0, M.jsx)(w, {
											icon: O,
											className: C("shrink-0 text-muted-foreground transition-transform", i && "rotate-90")
										}),
										/* @__PURE__ */ (0, M.jsx)(xt, {
											severity: jt(n.v),
											className: "shrink-0"
										}),
										/* @__PURE__ */ (0, M.jsxs)("span", {
											className: "min-w-0 truncate",
											title: e === "" ? void 0 : `${e}${ir(n)}`,
											"data-lint-location": "",
											children: [
												e !== "" && /* @__PURE__ */ (0, M.jsx)(ar, { file: e }),
												/* @__PURE__ */ (0, M.jsx)("span", {
													className: "text-muted-foreground",
													children: ir(n)
												}),
												n.v.rule?.method && /* @__PURE__ */ (0, M.jsxs)("span", {
													className: "text-muted-foreground",
													children: [
														" (",
														n.v.rule.method,
														")"
													]
												})
											]
										})
									]
								}), /* @__PURE__ */ (0, M.jsx)("span", {
									className: "block break-words pl-[18px] text-foreground",
									"data-lint-message": "",
									children: n.v.message ?? ""
								})]
							}), /* @__PURE__ */ (0, M.jsx)("div", {
								className: "-my-1 opacity-0 focus-within:opacity-100 group-hover:opacity-100",
								children: /* @__PURE__ */ (0, M.jsx)(Gt, { items: xn({
									violation: n.v,
									linter: n.linter,
									file: n.rawFile ?? "",
									workDir: n.workDir ?? "",
									...r
								}) })
							})]
						}), i && /* @__PURE__ */ (0, M.jsx)("div", {
							className: "mb-1 ml-[18px] font-sans",
							children: /* @__PURE__ */ (0, M.jsx)(Pn, {
								violation: n.v,
								linter: n.linter,
								file: n.rawFile ?? "",
								workDir: n.workDir ?? "",
								showMessage: !1,
								...r
							})
						})]
					}, t);
				}) })
			}, e)) })]
		}, e.linter))
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/ProblemFacet.tsx
var sr = {
	autoSubmit: !1,
	compact: !1
};
function cr({ label: e, options: t, state: n, onChange: r }) {
	return /* @__PURE__ */ (0, M.jsxs)("section", {
		"aria-label": e,
		children: [/* @__PURE__ */ (0, M.jsx)("div", {
			className: "mb-1 text-xs font-medium text-muted-foreground",
			children: e
		}), /* @__PURE__ */ (0, M.jsx)(rt.Provider, {
			value: sr,
			children: /* @__PURE__ */ (0, M.jsx)(ut, {
				chrome: "embedded",
				filter: {
					key: e,
					kind: "multi",
					label: e,
					options: t,
					value: Object.fromEntries(n),
					onChange: (e) => r(new Map(Object.entries(e)))
				}
			})
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/ProblemViewOptions.tsx
function lr({ active: e, children: t }) {
	return /* @__PURE__ */ (0, M.jsx)(ye, {
		label: `View Options${e ? ` (${e})` : ""}`,
		menuLabel: "Problem view options",
		variant: "ghost",
		menuClassName: "max-h-[70vh] w-80 overflow-y-auto p-3",
		children: () => /* @__PURE__ */ (0, M.jsx)("div", {
			className: "space-y-3",
			children: t
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/lint/grouping-options.ts
var ur = [
	{
		key: "linter-rule-file",
		label: "Linter · Rule · File",
		icon: je
	},
	{
		key: "linter-file",
		label: "Linter · File",
		icon: Ue
	},
	{
		key: "file-linter-rule",
		label: "File · Linter · Rule",
		icon: We
	},
	{
		key: "summary",
		label: "Summary",
		icon: He
	}
], dr = [
	"error",
	"warning",
	"info"
];
function fr({ grouping: e, onGroupingChange: t, className: n }) {
	return /* @__PURE__ */ (0, M.jsx)("div", {
		role: "group",
		"aria-label": "Group by",
		className: C("space-y-1", n),
		children: ur.map((n) => /* @__PURE__ */ (0, M.jsx)("button", {
			type: "button",
			"aria-pressed": n.key === e,
			className: C("block w-full rounded px-2 py-1 text-left text-xs hover:bg-muted", n.key === e && "bg-muted font-medium"),
			title: `Group by ${n.label}`,
			onClick: () => t(n.key),
			children: n.label
		}, n.key))
	});
}
function pr({ results: e, filters: t, onFiltersChange: n, className: r }) {
	let i = At(e, t.linter), a = kt(e, t.severity), o = [.../* @__PURE__ */ new Set([...Bt(e), ...t.linter.keys()])], s = t.severity.size > 0 || t.linter.size > 0, c = dr.filter((e) => i[e] > 0 || t.severity.size > 0).map((e) => ({
		value: e,
		label: dt[e].label,
		icon: /* @__PURE__ */ (0, M.jsx)(xt, { severity: e }),
		count: i[e]
	})), l = o.map((e) => ({
		value: e,
		label: mt(e)?.label ?? e,
		title: e,
		icon: /* @__PURE__ */ (0, M.jsx)(yt, { tool: e }),
		count: a[e] ?? 0
	}));
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: C("space-y-3", r),
		"data-lint-filter-facets": "",
		children: [
			/* @__PURE__ */ (0, M.jsx)(cr, {
				label: "Severity",
				options: c,
				state: t.severity,
				onChange: (e) => n({
					...t,
					severity: e
				})
			}),
			(o.length > 1 || t.linter.size > 0) && /* @__PURE__ */ (0, M.jsx)(cr, {
				label: "Linter",
				options: l,
				state: t.linter,
				onChange: (e) => n({
					...t,
					linter: e
				})
			}),
			s && /* @__PURE__ */ (0, M.jsx)("button", {
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
// ../../../clicky-ui/packages/ui/src/data/lint/LintTreeNode.tsx
function mr(e) {
	if (e.kind === "linter") return e.children?.find((e) => e.kind === "linter-error")?.message;
}
function hr({ node: e }) {
	switch (e.kind) {
		case "linter": return e.linter?.status === "running" ? /* @__PURE__ */ (0, M.jsx)(bt, { outcome: "running" }) : e.linter?.status === "pending" ? /* @__PURE__ */ (0, M.jsx)(w, {
			icon: De,
			title: "Pending",
			className: "shrink-0 text-muted-foreground"
		}) : mr(e) === void 0 ? mt(Mt(e)) ? /* @__PURE__ */ (0, M.jsx)(yt, { tool: Mt(e) }) : /* @__PURE__ */ (0, M.jsx)(w, {
			icon: s,
			className: "shrink-0 text-muted-foreground"
		}) : /* @__PURE__ */ (0, M.jsx)(xt, { severity: "error" });
		case "lint-folder": return /* @__PURE__ */ (0, M.jsx)(w, {
			icon: Ue,
			className: "shrink-0 text-muted-foreground"
		});
		case "lint-file": return /* @__PURE__ */ (0, M.jsx)(w, {
			icon: We,
			className: "shrink-0 text-muted-foreground"
		});
		case "lint-rule":
		case "lint-rule-group": return /* @__PURE__ */ (0, M.jsx)(w, {
			icon: an,
			className: "shrink-0 text-muted-foreground"
		});
		case "lint-target": return /* @__PURE__ */ (0, M.jsx)(w, {
			icon: e.targetKind === "entity" ? c : e.targetKind === "url" ? gt : Me,
			className: "shrink-0 text-muted-foreground"
		});
		case "linter-error": return /* @__PURE__ */ (0, M.jsx)(xt, { severity: "error" });
	}
}
function gr({ node: e, selected: t, actions: n = [], onIgnore: r, ignoreBusy: i = !1 }) {
	let a = e.kind === "linter-error" ? 0 : Pt(e), o = mr(e), s = o !== void 0, c = e.skipped ? e.message : s ? "failed to run" : void 0, l = s ? o : c, u = e.kind === "linter" ? kn(e.linter) : void 0, d = u ? `cached · ${u.replace(/ ago$/, "")}` : e.kind === "linter" && !e.linter?.status && e.linter?.duration ? tt(e.linter.duration) : "";
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
		/* @__PURE__ */ (0, M.jsx)(hr, { node: e }),
		/* @__PURE__ */ (0, M.jsx)("span", {
			className: C("min-w-0 truncate", t ? "font-semibold" : "font-medium", e.kind === "linter-error" || s ? "text-red-600 dark:text-red-400" : "text-foreground", (e.skipped || e.kind === "linter" && e.linter?.status === "pending") && "text-muted-foreground"),
			title: e.name,
			children: e.name
		}),
		/* @__PURE__ */ (0, M.jsx)("span", { className: "flex-1" }),
		c && /* @__PURE__ */ (0, M.jsx)("span", {
			className: C("min-w-0 shrink-[100] truncate text-xs", s ? "text-red-600 dark:text-red-400" : "text-muted-foreground"),
			title: l,
			children: c
		}),
		d && /* @__PURE__ */ (0, M.jsx)("span", {
			className: "shrink-0 text-xs tabular-nums text-muted-foreground",
			children: d
		}),
		a > 0 && /* @__PURE__ */ (0, M.jsx)("span", {
			className: "flex shrink-0",
			title: `${a} violation${a === 1 ? "" : "s"}`,
			children: /* @__PURE__ */ (0, M.jsx)("span", {
				className: "text-xs tabular-nums text-muted-foreground",
				children: a
			})
		}),
		r && /* @__PURE__ */ (0, M.jsx)(On, {
			actions: n,
			onIgnore: r,
			busy: i,
			label: e.name
		})
	] });
}
// ../../../clicky-ui/packages/ui/src/data/lint/LintResults.tsx
var _r = [
	{
		id: "tree",
		label: "Tree",
		icon: je
	},
	{
		id: "list",
		label: "List",
		icon: Ae
	},
	{
		id: "table",
		label: "Table",
		icon: d
	}
], vr = 640, yr = [
	"error",
	"warning",
	"info"
];
function br({ results: e }) {
	let t = At(e, /* @__PURE__ */ new Map()), n = e.filter(Lt).length;
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "ml-auto flex shrink-0 items-center gap-3 text-xs text-muted-foreground",
		"data-lint-summary": "",
		children: [yr.filter((e) => t[e] > 0).map((e) => /* @__PURE__ */ (0, M.jsxs)("span", {
			className: "inline-flex items-center gap-1",
			children: [/* @__PURE__ */ (0, M.jsx)(xt, { severity: e }), /* @__PURE__ */ (0, M.jsx)("span", {
				className: "tabular-nums",
				children: t[e]
			})]
		}, e)), n > 0 && /* @__PURE__ */ (0, M.jsxs)("span", {
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
function xr(e, t) {
	return e.length === 0 ? "No lint results" : t.severity.size > 0 || t.linter.size > 0 ? "No violations match the current filters" : "No lint violations";
}
function Sr({ results: e, grouping: t, onGroupingChange: n, filters: r, onFiltersChange: i, selected: a, onSelect: o, expandAll: s, onExpandAllChange: c, onIgnore: l, ignoreBusy: u, onOpenFile: d, loadSource: f, entityHref: p, renderRun: m, headerSlot: h, title: g, view: _ = "tree", onViewChange: v, defaultSplit: y = 45, className: b }) {
	let x = (0, j.useMemo)(() => zt({
		results: e,
		grouping: t,
		filters: r
	}), [
		e,
		t,
		r
	]), ee = (0, j.useMemo)(() => a ? Ot(x, a) : null, [x, a]), { ref: te, wider: ne } = lt(vr), re = /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex flex-wrap items-center gap-2 border-b border-border bg-card px-density-4 py-density-2",
		"data-lint-filter-bar": "",
		children: [v && /* @__PURE__ */ (0, M.jsx)(ge, {
			size: "sm",
			"aria-label": "View",
			value: _,
			options: _r,
			onChange: v
		}), /* @__PURE__ */ (0, M.jsxs)(lr, {
			active: r.severity.size + r.linter.size,
			children: [_ === "tree" && /* @__PURE__ */ (0, M.jsx)(fr, {
				grouping: t,
				onGroupingChange: n
			}), /* @__PURE__ */ (0, M.jsx)(pr, {
				results: e,
				filters: r,
				onFiltersChange: i
			})]
		})]
	}), S = /* @__PURE__ */ (0, M.jsx)(qe, {
		className: "min-h-0",
		roots: x,
		getKey: (e) => e.route_path,
		getChildren: (e) => e.children,
		selected: ee,
		onSelect: (e) => o(e.route_path),
		expandAll: s,
		onExpandAllChange: c,
		defaultOpen: (e, t) => t < 1 || e.failed,
		showSearch: !0,
		searchLabel: "Filter findings",
		getSearchText: (e) => `${e.name} ${e.linterName ?? ""} ${e.ruleName ?? ""}`,
		empty: /* @__PURE__ */ (0, M.jsx)("div", {
			className: "p-8 text-center text-sm text-muted-foreground",
			children: xr(e, r)
		}),
		renderRow: ({ node: t, selected: n }) => /* @__PURE__ */ (0, M.jsx)(gr, {
			node: t,
			selected: n,
			actions: l ? _n(t, e) : [],
			onIgnore: l,
			ignoreBusy: u
		})
	}), ie = /* @__PURE__ */ (0, M.jsx)(Qn, {
		node: ee,
		results: e,
		onIgnore: l,
		ignoreBusy: u,
		onOpenFile: d,
		loadSource: f,
		entityHref: p,
		renderRun: m
	}), ae = g !== null || e.length > 0 || h != null;
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		ref: te,
		className: C("flex h-full min-h-0 flex-col bg-background", b),
		"data-lint-results": "",
		children: [ae && /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1 border-b border-border bg-card px-density-4 py-density-2",
			"data-lint-header": "",
			children: [
				g === void 0 ? /* @__PURE__ */ (0, M.jsxs)("h1", {
					className: "inline-flex shrink-0 items-center gap-1.5 text-base font-bold",
					children: [/* @__PURE__ */ (0, M.jsx)(w, {
						icon: Ce,
						className: "text-primary"
					}), "Lint Results"]
				}) : g,
				/* @__PURE__ */ (0, M.jsx)(br, { results: e }),
				h && /* @__PURE__ */ (0, M.jsx)("div", {
					className: "flex shrink-0 items-center gap-1",
					children: h
				})
			]
		}), _ === "list" ? /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [re, /* @__PURE__ */ (0, M.jsx)("div", {
			className: "min-h-0 flex-1 overflow-y-auto",
			children: /* @__PURE__ */ (0, M.jsx)(or, {
				results: e,
				filters: r,
				onIgnore: l,
				ignoreBusy: u,
				onOpenFile: d,
				loadSource: f,
				entityHref: p
			})
		})] }) : _ === "table" ? /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [re, /* @__PURE__ */ (0, M.jsx)("div", {
			className: "min-h-0 flex-1 overflow-y-auto",
			"data-lint-table": "",
			children: /* @__PURE__ */ (0, M.jsx)(nr, {
				results: e,
				filters: r,
				onIgnore: l,
				ignoreBusy: u,
				onOpenFile: d,
				loadSource: f,
				entityHref: p
			})
		})] }) : ne ? /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [re, /* @__PURE__ */ (0, M.jsx)(ot, {
			className: "min-h-0 flex-1",
			defaultSplit: y,
			minLeft: 25,
			minRight: 30,
			leftClass: "border-r-2 border-border bg-muted/30",
			left: S,
			right: ie
		})] }) : ee ? /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "flex min-h-0 flex-1 flex-col",
			children: [/* @__PURE__ */ (0, M.jsxs)("button", {
				type: "button",
				onClick: () => o(null),
				className: "inline-flex items-center gap-1.5 border-b border-border px-density-3 py-density-2 text-left text-sm text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
				children: [/* @__PURE__ */ (0, M.jsx)(w, { icon: he }), "Back to results"]
			}), /* @__PURE__ */ (0, M.jsx)("div", {
				className: "min-h-0 flex-1 overflow-hidden",
				children: ie
			})]
		}) : /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [re, /* @__PURE__ */ (0, M.jsx)("div", {
			className: "min-h-0 flex-1 overflow-y-auto",
			children: S
		})] })]
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/utils.ts
function Cr(e) {
	let t = (e || "").toLowerCase();
	return t.includes("run") ? a : t.includes("sleep") || t.includes("idle") ? Qe : t.includes("stop") || t.includes("halt") ? nn : t.includes("zombie") || t.includes("dead") ? we : t.includes("wait") || t.includes("block") ? Te : en;
}
function wr(e) {
	let t = (e || "").toLowerCase();
	return t.includes("run") ? "text-green-600" : t.includes("sleep") || t.includes("idle") ? "text-amber-500" : t.includes("stop") || t.includes("halt") ? "text-orange-600" : t.includes("zombie") || t.includes("dead") ? "text-red-600" : t.includes("wait") || t.includes("block") ? "text-blue-600" : "text-muted-foreground";
}
function Tr(e) {
	if (e.name) return e.name;
	if (e.command) {
		let [t] = e.command.split(/\s+/, 1);
		return t || `pid ${e.pid}`;
	}
	return `pid ${e.pid}`;
}
// ../../../clicky-ui/packages/ui/src/hooks/use-prompts.ts
var Er = "/api/todos", Dr = 2e3, Or = [];
function kr() {
	return typeof globalThis < "u" && globalThis.EventSource !== void 0;
}
function Ar(e) {
	let t = new URLSearchParams();
	e.owner && t.set("owner", e.owner), e.kind && t.set("kind", e.kind), e.state && t.set("state", e.state);
	for (let [n, r] of Object.entries(e.labels ?? {})) t.append("label", `${n}=${r}`);
	return t.toString();
}
async function jr(e, t, n) {
	let r = await fetch(`${e}/prompts/${encodeURIComponent(t)}/answer`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(n)
	});
	if (!r.ok) throw Error(await r.text() || `answer failed: ${r.status}`);
}
function Mr(e = {}) {
	let { owner: t, kind: n, state: r, labels: i, basePath: a = Er, enabled: o = !0, pollMs: s = Dr, forcePoll: c = !1 } = e, l = Ar({
		owner: t,
		kind: n,
		state: r,
		labels: i
	}), u = `${a}/prompts?${l}`, [d, f] = (0, j.useState)({
		scope: u,
		prompts: Or
	}), [p, m] = (0, j.useState)("idle"), h = b(), g = (0, j.useRef)(h);
	g.current = h, (0, j.useEffect)(() => {
		if (!o) return;
		let e = (e) => f({
			scope: u,
			prompts: e
		});
		if (c || !kr()) {
			let t = !1, n, r = async () => {
				try {
					let t = await fetch(`${a}/prompts${l ? `?${l}` : ""}`, { headers: { Accept: "application/json" } });
					t.ok && (e(await t.json()), m("polling"));
				} catch {
					m("connection lost — retrying");
				}
				t || (n = setTimeout(r, s));
			};
			return r(), () => {
				t = !0, n && clearTimeout(n);
			};
		}
		let t = g.current(`${a}/prompts/stream${l ? `?${l}` : ""}`);
		return m("connected"), t.addEventListener("prompts", (t) => {
			try {
				e(JSON.parse(t.data));
			} catch {}
		}), t.onerror = () => m("connection lost — retrying"), () => t.close();
	}, [
		u,
		l,
		a,
		o,
		s,
		c
	]);
	let _ = d.scope === u ? d.prompts : Or;
	return {
		prompts: _,
		pending: _.filter((e) => e.state === "pending"),
		status: p
	};
}
// ../../../clicky-ui/packages/ui/src/data/PromptForm.tsx
function Nr({ prompt: e, basePath: t, onResolved: n, lookupFetcher: r, connectionApiBase: i, idPrefix: a, frame: o }) {
	let [s, c] = (0, j.useState)(() => ({ ...e.value })), [l, u] = (0, j.useState)(!1), [d, f] = (0, j.useState)(null), [p, m] = (0, j.useState)(e.id);
	p !== e.id && (m(e.id), c({ ...e.value }), u(!1), f(null));
	let h = async (r) => {
		u(!0), f(null);
		try {
			await jr(t, e.id, r), n?.();
		} catch (e) {
			f(e instanceof Error ? e.message : String(e));
		} finally {
			u(!1);
		}
	}, g = /* @__PURE__ */ (0, M.jsx)(Le, {
		schema: e.schema,
		value: s,
		onChange: c,
		idPrefix: a ?? `prompt-${e.id}`,
		...r ? { lookupFetcher: r } : {},
		...i ? { connectionApiBase: i } : {}
	}), _ = /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex items-center justify-end gap-2",
		children: [
			d ? /* @__PURE__ */ (0, M.jsx)("span", {
				className: "mr-auto text-sm text-destructive",
				children: d
			}) : null,
			/* @__PURE__ */ (0, M.jsx)(fe, {
				variant: "ghost",
				disabled: l,
				onClick: () => void h({ cancelled: !0 }),
				children: "Cancel"
			}),
			/* @__PURE__ */ (0, M.jsx)(fe, {
				disabled: l,
				onClick: () => void h({ values: s }),
				children: "Submit"
			})
		]
	});
	return o ? /* @__PURE__ */ (0, M.jsx)(M.Fragment, { children: o({
		form: g,
		actions: _
	}) }) : /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "min-w-0 space-y-density-2",
		"data-prompt-form": e.id,
		children: [
			/* @__PURE__ */ (0, M.jsxs)("div", { children: [/* @__PURE__ */ (0, M.jsx)("div", {
				className: "text-sm font-medium",
				children: e.title
			}), e.description ? /* @__PURE__ */ (0, M.jsx)("div", {
				className: "text-xs text-muted-foreground",
				children: e.description
			}) : null] }),
			g,
			_
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/PromptDialog.tsx
function Pr({ prompt: e, basePath: t, open: n, onClose: r, onResolved: i, lookupFetcher: a, connectionApiBase: o, size: s = "lg" }) {
	return /* @__PURE__ */ (0, M.jsx)(Nr, {
		prompt: e,
		basePath: t,
		lookupFetcher: a,
		connectionApiBase: o,
		onResolved: () => {
			i?.(), r();
		},
		frame: ({ form: t, actions: i }) => /* @__PURE__ */ (0, M.jsx)(xe, {
			open: n,
			size: s,
			onClose: r,
			title: e.title,
			subtitle: e.description,
			footer: i,
			children: t
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/PromptBanner.model.ts
var Fr = [
	"decision",
	"choice",
	"action",
	"value",
	"approve",
	"approved"
], Ir = [
	"comment",
	"reason",
	"message",
	"note"
], Lr = /\b(approve|approved|allow|allowed|continue|yes|accept|accepted|proceed|run)\b/i, Rr = /\b(reject|rejected|deny|denied|cancel|cancelled|canceled|no|block|stop|decline|declined)\b/i;
function zr(e, t) {
	return e?.type ? Array.isArray(e.type) ? e.type.includes(t) : e.type === t : !1;
}
function Br(e) {
	return Object.entries(e.properties ?? {}).sort(([e], [t]) => {
		let n = Fr.indexOf(e), r = Fr.indexOf(t);
		return n === -1 && r === -1 ? 0 : n === -1 ? 1 : r === -1 ? -1 : n - r;
	});
}
function Vr(e, t) {
	let n = String(t);
	return [n, e["x-enum-labels"]?.[n]].filter(Boolean).join(" ");
}
function Hr(e) {
	let t = e.properties ?? {}, n = new Set(e.required ?? []);
	for (let e of Ir) {
		let r = t[e];
		if (r && zr(r, "string") && !r.enum) return {
			commentField: e,
			commentRequired: n.has(e),
			commentLabel: r.title ?? "Comment"
		};
	}
	return { commentRequired: !1 };
}
function Ur(e) {
	for (let [t, n] of Br(e)) {
		let r = n.enum ?? [];
		if (r.length === 0) continue;
		let i = r.find((e) => Lr.test(Vr(n, e))), a = r.find((e) => Rr.test(Vr(n, e)));
		if (i !== void 0 && a !== void 0) return {
			field: t,
			approveValue: i,
			rejectValue: a,
			...Hr(e)
		};
	}
	return null;
}
function Wr(e) {
	let t = e.properties ?? {}, n = [
		"value",
		"approve",
		"approved",
		"confirm",
		"confirmed",
		"allowed"
	].find((e) => zr(t[e], "boolean")) ?? Object.entries(t).find(([, e]) => zr(e, "boolean"))?.[0];
	return n ? {
		field: n,
		approveValue: !0,
		rejectValue: !1,
		...Hr(e)
	} : null;
}
function Gr(e) {
	let t = e.schema;
	return t?.properties ? Ur(t) ?? Wr(t) : null;
}
function Kr(e) {
	return e.reduce((e, t) => (e.total += 1, t.state === "pending" ? e.pending += 1 : t.state === "answered" ? e.answered += 1 : t.state === "cancelled" ? e.cancelled += 1 : t.state === "expired" && (e.expired += 1), e), {
		total: 0,
		pending: 0,
		answered: 0,
		cancelled: 0,
		expired: 0
	});
}
var qr = {
	pending: 0,
	answered: 1
};
function Jr(e) {
	let t = (e) => qr[e.state] ?? 2;
	return [...e].sort((e, n) => t(e) - t(n));
}
function Yr(e) {
	if (e.cancelled || e.state === "cancelled") return "Cancelled";
	if (e.state === "expired") return "Expired";
	if (e.state === "pending") return e.kind ? `Pending ${e.kind}` : "Pending";
	let t = e.value ?? {}, n = Gr(e), r = [];
	if (n && n.field in t && (Object.is(t[n.field], n.approveValue) ? r.push("Approved") : Object.is(t[n.field], n.rejectValue) && r.push("Rejected")), n?.commentField) {
		let e = t[n.commentField];
		typeof e == "string" && e && r.push(e);
	}
	return r.length > 0 ? r.join(" - ") : Object.entries(t).filter(([, e]) => [
		"string",
		"number",
		"boolean"
	].includes(typeof e)).slice(0, 2).map(([e, t]) => `${e}: ${String(t)}`).join(" - ") || e.state;
}
// ../../../clicky-ui/packages/ui/src/data/PromptList.tsx
function Xr({ prompts: e, activeId: t, onSelect: n, className: r }) {
	return /* @__PURE__ */ (0, M.jsx)("ol", {
		className: C("divide-y divide-border overflow-hidden rounded-md border border-border bg-background text-xs", r),
		children: e.map((e) => {
			let r = C("flex w-full min-w-0 items-start gap-2 px-2 py-1.5 text-left", e.id === t && "bg-muted/60");
			return /* @__PURE__ */ (0, M.jsx)("li", { children: n && e.state === "pending" ? /* @__PURE__ */ (0, M.jsx)("button", {
				type: "button",
				className: C(r, "hover:bg-muted focus-visible:bg-muted focus-visible:outline-none"),
				onClick: () => n(e),
				children: /* @__PURE__ */ (0, M.jsx)(Zr, { prompt: e })
			}) : /* @__PURE__ */ (0, M.jsx)("div", {
				className: r,
				children: /* @__PURE__ */ (0, M.jsx)(Zr, { prompt: e })
			}) }, e.id);
		})
	});
}
function Zr({ prompt: e }) {
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, M.jsx)(Qr, { prompt: e }), /* @__PURE__ */ (0, M.jsxs)("span", {
		className: "block min-w-0 flex-1",
		children: [/* @__PURE__ */ (0, M.jsxs)("span", {
			className: "flex min-w-0 items-center gap-2",
			children: [/* @__PURE__ */ (0, M.jsx)("span", {
				className: "truncate font-medium text-foreground",
				"data-testid": "prompt-list-title",
				children: e.title
			}), e.kind ? /* @__PURE__ */ (0, M.jsx)("span", {
				className: "shrink-0 text-muted-foreground",
				children: e.kind
			}) : null]
		}), /* @__PURE__ */ (0, M.jsx)("span", {
			className: "mt-0.5 block truncate text-muted-foreground",
			children: Yr(e)
		})]
	})] });
}
function Qr({ prompt: e }) {
	return e.state === "answered" ? /* @__PURE__ */ (0, M.jsx)(w, {
		icon: ve,
		className: "mt-0.5 size-3.5 shrink-0 text-emerald-600"
	}) : e.state === "cancelled" || e.state === "expired" ? /* @__PURE__ */ (0, M.jsx)(w, {
		icon: pe,
		className: "mt-0.5 size-3.5 shrink-0 text-rose-600"
	}) : e.kind === "comment" ? /* @__PURE__ */ (0, M.jsx)(w, {
		icon: Ee,
		className: "mt-0.5 size-3.5 shrink-0"
	}) : /* @__PURE__ */ (0, M.jsx)(w, {
		icon: ue,
		className: "mt-0.5 size-3.5 shrink-0"
	});
}
// ../../../clicky-ui/packages/ui/src/data/PromptIndicator.tsx
function $r({ basePath: e = "/api/todos", enabled: t, className: n, lookupFetcher: r, connectionApiBase: i, autoOpen: a = !1, state: o, ...s }) {
	let { prompts: c, pending: l } = Mr({
		...s,
		basePath: e,
		...t === void 0 ? {} : { enabled: t }
	}), [u, d] = (0, j.useState)(null), [f, p] = (0, j.useState)(e);
	f !== e && (p(e), d(null));
	let m = (0, j.useMemo)(() => Jr(c), [c]), h = (0, j.useMemo)(() => Kr(c).pending, [c]), g = (0, j.useRef)(/* @__PURE__ */ new Set()), _ = a ? l.find((e) => !g.current.has(e.id)) : void 0;
	(0, j.useEffect)(() => {
		_ && !u && (g.current.add(_.id), d(_));
	}, [_, u]);
	let v = (e) => d((t) => t?.id === e ? null : t), y = h > 0 ? `${h} pending` : "Prompts";
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, M.jsx)(ye, {
		role: "dialog",
		menuLabel: "Prompts",
		align: "right",
		menuClassName: "w-80 max-w-[calc(100vw-2rem)] p-1",
		trigger: /* @__PURE__ */ (0, M.jsxs)(fe, {
			type: "button",
			variant: "ghost",
			size: "sm",
			title: h > 0 ? `${y} — answer from the list` : "No pending prompts",
			className: C("gap-1.5", h > 0 ? Ct.warned.className : "text-muted-foreground", n),
			children: [/* @__PURE__ */ (0, M.jsxs)("span", {
				className: "relative inline-flex",
				children: [/* @__PURE__ */ (0, M.jsx)(w, {
					icon: Me,
					className: "size-4"
				}), h > 0 ? /* @__PURE__ */ (0, M.jsx)("span", {
					"data-prompt-pulse": "",
					"aria-hidden": !0,
					className: "absolute -right-0.5 -top-0.5 size-1.5 animate-pulse rounded-full bg-current"
				}) : null]
			}), y]
		}),
		children: (e) => m.length === 0 ? /* @__PURE__ */ (0, M.jsx)("div", {
			className: "px-2 py-1.5 text-xs text-muted-foreground",
			children: "No prompts"
		}) : /* @__PURE__ */ (0, M.jsx)(Xr, {
			prompts: m,
			activeId: u?.id,
			className: "border-0",
			onSelect: (t) => {
				e(), d(t);
			}
		})
	}), u ? /* @__PURE__ */ (0, M.jsx)(Pr, {
		prompt: u,
		basePath: e,
		open: !0,
		onClose: () => v(u.id),
		lookupFetcher: r,
		connectionApiBase: i
	}, u.id) : null] });
}
// ../../../clicky-ui/packages/ui/src/data/har/HarPanel.tsx
var ei = 80, ti = (0, j.createContext)(void 0), ni = [
	{
		key: "request.method",
		label: "Method",
		shrink: !0
	},
	{
		key: "request.url",
		label: "URL",
		grow: !0,
		render: (e, t) => /* @__PURE__ */ (0, M.jsxs)("span", {
			className: "inline-flex min-w-0 items-baseline gap-density-2",
			children: [/* @__PURE__ */ (0, M.jsx)("span", {
				title: String(e ?? ""),
				children: String(e ?? "")
			}), t._error ? /* @__PURE__ */ (0, M.jsx)("span", {
				className: "text-destructive text-xs",
				title: t._error,
				children: ci(t._error)
			}) : null]
		})
	},
	{
		key: "response.status",
		label: "Status",
		shrink: !0,
		render: (e, t) => /* @__PURE__ */ (0, M.jsx)(ii, {
			status: Number(e ?? 0),
			entry: t
		}),
		sortValue: (e) => Number(e ?? 0),
		filterValue: (e, t) => si(Number(e ?? 0), t)
	},
	{
		key: "time",
		label: "Time",
		align: "right",
		shrink: !0,
		render: (e, t) => t._pending ? /* @__PURE__ */ (0, M.jsx)(ai, { entry: t }) : `${Number(e ?? 0).toFixed(0)}ms`,
		sortValue: (e, t) => t._pending ? oi(t, Date.now()) : Number(e ?? 0)
	},
	{
		key: "response.bodySize",
		label: "Size",
		align: "right",
		shrink: !0,
		render: (e) => _i(Number(e ?? 0)),
		sortValue: (e) => Number(e ?? 0)
	},
	{
		key: "response.content.mimeType",
		label: "Type",
		shrink: !0
	}
];
function ri({ entries: e, search: t, emptyLabel: n = "No HTTP traffic captured", className: r, bodySource: i, menuActions: a }) {
	let o = (0, j.useMemo)(() => t === void 0 || t.trim() === "" ? e : e.filter((e) => gi(t.trim().toLowerCase(), e)), [e, t]), s = (0, j.useMemo)(() => e.filter((e) => e._pending).length, [e]), c = y(s > 0);
	return !e || e.length === 0 ? /* @__PURE__ */ (0, M.jsx)("div", {
		className: "p-density-6 text-center text-muted-foreground text-sm",
		children: n
	}) : /* @__PURE__ */ (0, M.jsx)("div", {
		className: `flex h-full min-h-0 flex-col ${r ?? ""}`,
		children: /* @__PURE__ */ (0, M.jsx)(ti.Provider, {
			value: c,
			children: /* @__PURE__ */ (0, M.jsx)(g, {
				data: o,
				columns: ni,
				autoFilter: !0,
				showGlobalFilter: t === void 0,
				globalFilterPlaceholder: "Filter URL, method, or body…",
				defaultSort: {
					key: "time",
					dir: "asc"
				},
				emptyMessage: n,
				...a ? { menuActions: a } : {},
				...s > 0 ? { footer: ({ visibleRowCount: e, totalRowCount: t }) => `${e} of ${t} row${t === 1 ? "" : "s"} · ${s} running` } : {},
				getRowId: (e, t) => e._id ?? `${e.startedDateTime ?? t}-${e.request.method}-${e.request.url}${e._pending ? "-pending" : `-${e.time}`}`,
				renderExpandedRow: (e) => /* @__PURE__ */ (0, M.jsx)(li, {
					entry: e,
					bodySource: i
				})
			})
		})
	});
}
function ii({ status: e, entry: t }) {
	return t._pending ? /* @__PURE__ */ (0, M.jsxs)("span", {
		className: "inline-flex items-center gap-1 text-muted-foreground",
		children: [/* @__PURE__ */ (0, M.jsx)(w, {
			icon: ht,
			className: "animate-spin"
		}), "running"]
	}) : t._error ? /* @__PURE__ */ (0, M.jsx)("span", {
		className: "font-semibold text-destructive",
		children: "ERR"
	}) : /* @__PURE__ */ (0, M.jsx)("span", {
		className: vi(e),
		children: String(e)
	});
}
function ai({ entry: e }) {
	let t = (0, j.useContext)(ti);
	if (t === void 0) throw Error("PendingElapsed must render inside HarPanel's clock");
	return /* @__PURE__ */ (0, M.jsx)("span", {
		className: "tabular-nums",
		children: Ge(oi(e, t))
	});
}
function oi(e, t) {
	let n = Date.parse(e.startedDateTime ?? "");
	return Number.isNaN(n) ? e.time : Math.max(0, t - n);
}
function si(e, t) {
	return t._pending ? "running" : t._error ? "ERR" : String(e);
}
function ci(e) {
	return e.length > ei ? `${e.slice(0, ei)}…` : e;
}
function li({ entry: e, bodySource: t }) {
	let n = (0, j.useContext)(ti), r = y(!!e._pending && n === void 0);
	return /* @__PURE__ */ (0, M.jsx)(ti.Provider, {
		value: n ?? r,
		children: /* @__PURE__ */ (0, M.jsx)(ui, {
			entry: e,
			bodySource: t
		})
	});
}
function ui({ entry: e, bodySource: t }) {
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
		e._error && /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "mb-density-3",
			children: [/* @__PURE__ */ (0, M.jsx)("div", {
				className: "font-semibold text-foreground mb-1",
				children: "Transport Error"
			}), /* @__PURE__ */ (0, M.jsx)("pre", {
				className: "whitespace-pre-wrap break-all text-destructive",
				children: e._error
			})]
		}),
		e._pending && /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "mb-density-3 text-muted-foreground",
			children: ["Waiting for response — in flight for ", /* @__PURE__ */ (0, M.jsx)(ai, { entry: e })]
		}),
		/* @__PURE__ */ (0, M.jsxs)("div", {
			className: "grid grid-cols-2 gap-density-4",
			children: [/* @__PURE__ */ (0, M.jsxs)("div", { children: [/* @__PURE__ */ (0, M.jsx)(fi, {
				title: "Request Headers",
				...e.request.headers ? { headers: e.request.headers } : {}
			}), (e.request.postData?.text || t?.hasBody(e, "request")) && /* @__PURE__ */ (0, M.jsx)(di, {
				entry: e,
				part: "request",
				text: e.request.postData?.text,
				mimeType: e.request.postData?.mimeType,
				bodySource: t
			})] }), /* @__PURE__ */ (0, M.jsx)("div", { children: /* @__PURE__ */ (0, M.jsx)(fi, {
				title: "Response Headers",
				...e.response.headers ? { headers: e.response.headers } : {}
			}) })]
		}),
		(e.response.content?.text || t?.hasBody(e, "response")) && /* @__PURE__ */ (0, M.jsx)(di, {
			entry: e,
			part: "response",
			text: e.response.content?.text,
			mimeType: e.response.content?.mimeType,
			bodySource: t
		})
	] });
}
function di({ entry: e, part: t, text: n, mimeType: r, bodySource: i }) {
	let a = !n && i?.hasBody(e, t), o = e._id;
	if (a && !o) throw Error("Private HAR body requires a request ID");
	let [s, c] = (0, j.useState)(), [l, u] = (0, j.useState)();
	return (0, j.useEffect)(() => {
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
	]), /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "mt-density-3",
		children: [
			/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "mb-1 flex items-center gap-2 font-semibold text-foreground",
				children: [t === "request" ? "Request Body" : "Response Body", a && i && /* @__PURE__ */ (0, M.jsx)("a", {
					className: "text-sm font-normal text-primary underline",
					href: i.downloadHref(e, t),
					download: !0,
					children: "Download full body"
				})]
			}),
			/* @__PURE__ */ (0, M.jsx)("div", {
				className: "bg-background p-density-2 rounded border border-border overflow-auto max-h-64",
				children: l ? /* @__PURE__ */ (0, M.jsx)("span", {
					role: "alert",
					className: "text-destructive",
					children: l.message
				}) : a && s === void 0 ? /* @__PURE__ */ (0, M.jsx)("span", {
					className: "text-muted-foreground",
					children: "Loading preview…"
				}) : /* @__PURE__ */ (0, M.jsx)(pi, {
					text: n || s || "",
					...r ? { mimeType: r } : {}
				})
			}),
			a && /* @__PURE__ */ (0, M.jsx)("p", {
				className: "mt-1 text-xs text-muted-foreground",
				children: "Preview shows at most the first 64 KiB."
			})
		]
	});
}
function fi({ title: e, headers: t }) {
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, M.jsx)("div", {
		className: "font-semibold text-foreground mb-1",
		children: e
	}), /* @__PURE__ */ (0, M.jsx)("div", {
		className: "space-y-0.5",
		children: t?.map((e, t) => /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "whitespace-nowrap",
			children: [
				/* @__PURE__ */ (0, M.jsxs)("span", {
					className: "text-purple-600 dark:text-purple-400",
					children: [e.name, ":"]
				}),
				" ",
				e.value
			]
		}, t))
	})] });
}
function pi({ text: e, mimeType: t }) {
	if (hi(t)) {
		let t = mi(e);
		if (t !== null) return /* @__PURE__ */ (0, M.jsx)(Je, { data: t });
	}
	return /* @__PURE__ */ (0, M.jsx)("pre", {
		className: "whitespace-pre-wrap break-all",
		children: e
	});
}
function mi(e) {
	try {
		return JSON.parse(e);
	} catch {
		return null;
	}
}
function hi(e) {
	return !!e && (e.includes("json") || e.includes("javascript"));
}
function gi(e, t) {
	return [
		t.request.method,
		t.request.url,
		t.request.postData?.text,
		t.response.content?.text,
		t._error
	].some((t) => !!t && t.toLowerCase().includes(e));
}
function _i(e) {
	return e < 0 ? "" : e < 1024 ? `${e}B` : e < 1048576 ? `${(e / 1024).toFixed(1)}KB` : `${(e / 1024 / 1024).toFixed(1)}MB`;
}
function vi(e) {
	return e >= 500 ? "text-red-600" : e >= 400 ? "text-amber-600" : e >= 300 ? "text-blue-600" : e >= 200 ? "text-green-600" : "text-muted-foreground";
}
// ../../../clicky-ui/packages/ui/src/data/traces/profile-aggregate.ts
function yi(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.has(n.sampleType) || t.set(n.sampleType, {
		sampleType: n.sampleType,
		unit: n.unit
	});
	return [...t.values()];
}
function bi(e, t) {
	let n = e.filter((e) => e.sampleType === t), r = new Set(n.map((e) => e.unit));
	if (r.size > 1) throw Error(`sample type "${t}" has mixed units ${[...r].map((e) => `"${e}"`).join(", ")}`);
	for (let e of n) if (e.frames.length === 0) throw Error(`a "${t}" sample has no frames`);
	return n;
}
function xi(e, t) {
	let n = bi(e, t), r = /* @__PURE__ */ new Map(), i = 0;
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
function Si(e, t, n, r) {
	let i = /* @__PURE__ */ new Map();
	for (let r of bi(e, t)) {
		if (!r.frames.some((e) => e.func === n)) continue;
		let e = r.frames.map((e) => `${e.func}@${e.file}:${e.line}`).join("\n"), t = i.get(e);
		t ? t.value += r.value : i.set(e, {
			frames: r.frames,
			value: r.value
		});
	}
	return [...i.values()].sort((e, t) => t.value - e.value).slice(0, r);
}
var Ci = 1e6;
function wi(e, t) {
	switch (t) {
		case "bytes": return Xe(e);
		case "nanoseconds":
		case "ns": return Math.abs(e) < 1e3 ? `${ze(e)} ns` : Math.abs(e) < Ci ? `${ze(e / 1e3)} µs` : Ge(e / Ci);
		case "microseconds":
		case "us": return Math.abs(e) < 1e3 ? `${ze(e)} µs` : Ge(e / 1e3);
		case "milliseconds":
		case "ms": return Ge(e);
		case "seconds": return Ge(e, { from: "s" });
		case "count":
		case "samples":
		case "": return ze(e);
		case "percent": return `${ze(e)}%`;
		default: return `${ze(e)} ${t}`;
	}
}
// ../../../clicky-ui/packages/ui/src/data/traces/process-tree.ts
var Ti = (e, t) => Date.parse(e.row.startedAt) - Date.parse(t.row.startedAt) || e.row.pid - t.row.pid;
function Ei(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) {
		if (t.has(n.key)) throw Error(`process tree key "${n.key}" appears twice`);
		t.set(n.key, {
			row: n,
			children: [],
			exited: Ai(n)
		});
	}
	let n = [];
	for (let e of t.values()) {
		let r = e.row.parentKey ? t.get(e.row.parentKey) : void 0;
		(r ? r.children : n).push(e);
	}
	n.sort((e, t) => Number(!!t.row.isRoot) - Number(!!e.row.isRoot) || Ti(e, t));
	let r = ji(n);
	if (r.length !== t.size) {
		let e = [...t.keys()].filter((e) => !r.some((t) => t.row.key === e));
		throw Error(`process tree has a parent cycle among keys ${e.map((e) => `"${e}"`).join(", ")}`);
	}
	for (let e of t.values()) e.children.sort(Ti);
	return n;
}
function Di(e) {
	return e.flatMap((e) => {
		let t = Di(e.children);
		return e.exited && t.length === 0 ? [] : [{
			...e,
			children: t
		}];
	});
}
function Oi(e) {
	return Date.parse(e.exitedAt ?? e.lastSeenAt) - Date.parse(e.startedAt);
}
function ki(e) {
	let t = new Date(e);
	if (Number.isNaN(t.getTime())) return e;
	let n = (e) => String(e).padStart(2, "0");
	return `${t.getFullYear()}-${n(t.getMonth() + 1)}-${n(t.getDate())} ${n(t.getHours())}:${n(t.getMinutes())}:${n(t.getSeconds())}`;
}
function Ai(e) {
	return e.status === "exited" || !!e.exitedAt;
}
function ji(e) {
	let t = [], n = (e) => {
		t.push(e), e.children.forEach(n);
	};
	return e.forEach(n), t;
}
function Mi(e) {
	return e == null ? "—" : `${e.toFixed(1)}%`;
}
function Ni(e) {
	return e.flatMap((e) => e.cpuPercent == null ? [] : [e.cpuPercent]);
}
function Pi(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.key, [...t.get(n.key) ?? [], n]);
	return new Map([...t].map(([e, t]) => [e, Ni(t.sort((e, t) => e.tick - t.tick))]));
}
function Fi(e, t) {
	let n = e.filter((e) => e.key === t).sort((e, t) => e.tick - t.tick);
	return {
		cpu: Ni(n),
		rss: n.map((e) => e.rssBytes)
	};
}
function Ii(e, t, n) {
	if (e.length < 2) return;
	let r = Math.max(...e);
	return e.map((i, a) => {
		let o = a / (e.length - 1) * t, s = r === 0 ? n : n - i / r * n;
		return `${o.toFixed(1)},${s.toFixed(1)}`;
	}).join(" ");
}
// ../../../clicky-ui/packages/ui/src/data/traces/GoProfileView.tsx
var Li = 20, Ri = 5;
function zi(e) {
	return e.file ? `${e.file}${e.line ? `:${e.line}` : ""}` : "";
}
function Bi({ value: e, total: t }) {
	return /* @__PURE__ */ (0, M.jsx)("span", {
		className: "text-muted-foreground",
		children: t > 0 ? `${(e / t * 100).toFixed(1)}%` : "—"
	});
}
function Vi({ samples: e, sampleType: t, unit: n, func: r, total: i }) {
	let a = Si(e, t, r, Ri);
	return /* @__PURE__ */ (0, M.jsx)("div", {
		className: "space-y-density-2",
		"data-profile-stacks": r,
		children: a.map((e, t) => /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "rounded border border-border bg-muted/30 p-density-2",
			children: [/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "mb-1 text-[11px] text-muted-foreground",
				children: [
					wi(e.value, n),
					" ",
					/* @__PURE__ */ (0, M.jsx)(Bi, {
						value: e.value,
						total: i
					})
				]
			}), /* @__PURE__ */ (0, M.jsx)("ol", {
				className: "space-y-0.5 font-mono text-[11px] leading-4",
				children: e.frames.map((e, t) => {
					let n = zi(e).split(/[\\/]/).at(-1);
					return /* @__PURE__ */ (0, M.jsxs)("li", {
						className: C("break-all", e.func === r ? "font-semibold text-foreground" : "text-muted-foreground"),
						children: [e.func, n && /* @__PURE__ */ (0, M.jsx)("span", {
							className: "ml-2 text-[10px] font-normal opacity-80",
							children: n
						})]
					}, t);
				})
			})]
		}, t))
	});
}
function Hi({ samples: e, topN: t = Li, className: n }) {
	let r = (0, j.useMemo)(() => yi(e), [e]), [i, a] = (0, j.useState)(null), o = r.find((e) => e.sampleType === i)?.sampleType ?? r[r.length - 1]?.sampleType, [s, c] = (0, j.useState)("flat"), [l, u] = (0, j.useState)(!1), [d, f] = (0, j.useState)(null), p = (0, j.useMemo)(() => o ? xi(e, o) : void 0, [e, o]);
	if (!o || !p) return /* @__PURE__ */ (0, M.jsx)("div", {
		className: "text-sm text-muted-foreground",
		children: "The profile has no samples"
	});
	let m = s === "flat" ? p.functions : [...p.functions].sort((e, t) => t.cum - e.cum || t.flat - e.flat), h = l ? m : m.slice(0, t), g = (e, t) => /* @__PURE__ */ (0, M.jsx)("th", {
		scope: "col",
		className: "px-2 py-1 text-right font-medium",
		"aria-sort": s === e ? "descending" : "none",
		children: /* @__PURE__ */ (0, M.jsx)("button", {
			type: "button",
			className: C("hover:text-foreground", s === e && "text-foreground"),
			onClick: () => c(e),
			children: t
		})
	});
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: C("space-y-density-2", n),
		"data-profile-top": "",
		children: [
			/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex flex-wrap items-center justify-between gap-density-2 text-xs text-muted-foreground",
				children: [/* @__PURE__ */ (0, M.jsxs)("span", { children: [
					"Total ",
					wi(p.total, p.unit),
					" across ",
					p.functions.length,
					" functions"
				] }), r.length > 1 && /* @__PURE__ */ (0, M.jsx)("div", {
					className: "min-w-[12rem]",
					children: /* @__PURE__ */ (0, M.jsx)(Ke, {
						"aria-label": "Sample type",
						value: o,
						onChange: (e) => {
							a(e.target.value), f(null);
						},
						options: r.map((e) => ({
							value: e.sampleType,
							label: `${e.sampleType} (${e.unit})`
						}))
					})
				})]
			}),
			/* @__PURE__ */ (0, M.jsxs)("table", {
				className: "w-full table-fixed border-collapse text-xs",
				children: [/* @__PURE__ */ (0, M.jsx)("thead", {
					className: "text-muted-foreground",
					children: /* @__PURE__ */ (0, M.jsxs)("tr", {
						className: "border-b border-border",
						children: [
							/* @__PURE__ */ (0, M.jsx)("th", {
								scope: "col",
								className: "px-2 py-1 text-left font-medium",
								children: "Function"
							}),
							g("flat", "Flat"),
							/* @__PURE__ */ (0, M.jsx)("th", {
								scope: "col",
								className: "w-16 px-2 py-1 text-right font-medium",
								children: "Flat %"
							}),
							g("cum", "Cum"),
							/* @__PURE__ */ (0, M.jsx)("th", {
								scope: "col",
								className: "w-16 px-2 py-1 text-right font-medium",
								children: "Cum %"
							})
						]
					})
				}), /* @__PURE__ */ (0, M.jsx)("tbody", { children: h.map((t) => {
					let n = d === t.func;
					return /* @__PURE__ */ (0, M.jsxs)(j.Fragment, { children: [/* @__PURE__ */ (0, M.jsxs)("tr", {
						className: "border-b border-border/60 hover:bg-muted/40",
						"data-profile-function": t.func,
						children: [
							/* @__PURE__ */ (0, M.jsx)("td", {
								className: "px-2 py-1",
								children: /* @__PURE__ */ (0, M.jsxs)("button", {
									type: "button",
									"aria-expanded": n,
									className: "flex w-full items-start gap-1 text-left",
									onClick: () => f(n ? null : t.func),
									children: [/* @__PURE__ */ (0, M.jsx)(w, {
										icon: n ? be : O,
										className: "mt-0.5 shrink-0 text-[11px]"
									}), /* @__PURE__ */ (0, M.jsxs)("span", {
										className: "min-w-0",
										children: [/* @__PURE__ */ (0, M.jsx)("span", {
											className: "block break-all font-mono text-[11px] font-semibold",
											children: t.func
										}), zi(t) && /* @__PURE__ */ (0, M.jsx)("span", {
											className: "block break-all text-[10px] text-muted-foreground",
											children: zi(t)
										})]
									})]
								})
							}),
							/* @__PURE__ */ (0, M.jsx)("td", {
								className: "px-2 py-1 text-right tabular-nums",
								children: wi(t.flat, p.unit)
							}),
							/* @__PURE__ */ (0, M.jsx)("td", {
								className: "px-2 py-1 text-right tabular-nums",
								children: /* @__PURE__ */ (0, M.jsx)(Bi, {
									value: t.flat,
									total: p.total
								})
							}),
							/* @__PURE__ */ (0, M.jsx)("td", {
								className: "px-2 py-1 text-right tabular-nums",
								children: wi(t.cum, p.unit)
							}),
							/* @__PURE__ */ (0, M.jsx)("td", {
								className: "px-2 py-1 text-right tabular-nums",
								children: /* @__PURE__ */ (0, M.jsx)(Bi, {
									value: t.cum,
									total: p.total
								})
							})
						]
					}), n && /* @__PURE__ */ (0, M.jsx)("tr", { children: /* @__PURE__ */ (0, M.jsx)("td", {
						colSpan: 5,
						className: "px-2 py-density-2",
						children: /* @__PURE__ */ (0, M.jsx)(Vi, {
							samples: e,
							sampleType: o,
							unit: p.unit,
							func: t.func,
							total: p.total
						})
					}) })] }, t.func);
				}) })]
			}),
			!l && m.length > h.length && /* @__PURE__ */ (0, M.jsxs)("button", {
				type: "button",
				className: "text-xs text-primary hover:underline",
				onClick: () => u(!0),
				children: [
					"Show all ",
					m.length,
					" functions"
				]
			})
		]
	});
}
function Ui(e) {
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
function Wi({ records: e, processKey: t, topN: n, className: r }) {
	let i = (0, j.useMemo)(() => new Set(e.map((e) => e.key)), [e]);
	if (i.size > 1 && t === void 0) throw Error(`GoProfileView got records of ${i.size} processes; pass processKey to choose one`);
	let a = (0, j.useMemo)(() => t === void 0 ? e : e.filter((e) => e.key === t), [e, t]), o = (0, j.useMemo)(() => Ui(a), [a]), [s, c] = (0, j.useState)(null), l = o.find((e) => e.checkpointId === s) ?? o[o.length - 1];
	return l ? /* @__PURE__ */ (0, M.jsxs)("div", {
		className: C("space-y-density-2", r),
		"data-go-profile": l.profile,
		children: [o.length > 1 ? /* @__PURE__ */ (0, M.jsx)("div", {
			className: "max-w-md",
			children: /* @__PURE__ */ (0, M.jsx)(Ke, {
				"aria-label": "Profile checkpoint",
				value: l.checkpointId,
				onChange: (e) => c(e.target.value),
				options: [...o].reverse().map((e) => ({
					value: e.checkpointId,
					label: [
						ki(e.takenAt),
						e.profile,
						e.reason,
						e.source
					].filter(Boolean).join(" · ")
				}))
			})
		}) : /* @__PURE__ */ (0, M.jsx)("div", {
			className: "text-xs text-muted-foreground",
			children: [
				l.profile,
				ki(l.takenAt),
				l.reason,
				l.source
			].filter(Boolean).join(" · ")
		}), /* @__PURE__ */ (0, M.jsx)(Hi, {
			samples: l.records,
			...n === void 0 ? {} : { topN: n }
		}, l.checkpointId)]
	}) : /* @__PURE__ */ (0, M.jsx)("div", {
		className: "text-sm text-muted-foreground",
		children: "No profile collected yet"
	});
}
// ../../../clicky-ui/packages/ui/src/data/traces/stack-checkpoints.ts
function Gi(e, t) {
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
// ../../../clicky-ui/packages/ui/src/data/traces/ProcessRecordViews.tsx
function Ki(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.key, [...t.get(n.key) ?? [], n]);
	return [...t];
}
function qi({ processes: e, render: t }) {
	return /* @__PURE__ */ (0, M.jsx)("div", {
		className: "min-h-0 flex-1 space-y-density-3 overflow-auto p-density-2",
		children: e.map(([n, r]) => /* @__PURE__ */ (0, M.jsxs)("div", {
			"data-trace-process": n,
			children: [e.length > 1 && /* @__PURE__ */ (0, M.jsx)("div", {
				className: "mb-1 font-mono text-xs text-muted-foreground",
				children: n
			}), t(n, r)]
		}, n))
	});
}
function Ji({ rows: e }) {
	let t = (0, j.useMemo)(() => Ki(e), [e]);
	return /* @__PURE__ */ (0, M.jsx)(qi, {
		processes: t,
		render: (e, t) => /* @__PURE__ */ (0, M.jsx)(Wi, {
			records: t,
			processKey: e
		})
	});
}
function Yi({ rows: e }) {
	let t = (0, j.useMemo)(() => Ki(e).map(([t]) => [t, Gi(e, t).reverse()]), [e]);
	return /* @__PURE__ */ (0, M.jsx)(qi, {
		processes: t,
		render: (e, t) => /* @__PURE__ */ (0, M.jsx)("div", {
			className: "space-y-density-2",
			children: t.map((e) => /* @__PURE__ */ (0, M.jsx)("section", {
				className: "rounded border border-border px-density-2",
				"data-stack-checkpoint": e.checkpointId,
				children: /* @__PURE__ */ (0, M.jsx)(v, {
					records: e.records,
					headerActions: /* @__PURE__ */ (0, M.jsx)("span", {
						className: "font-mono text-[11px] text-muted-foreground",
						children: e.transport
					})
				})
			}, e.checkpointId))
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/traces/Sparkline.tsx
function Xi({ values: e, label: t, width: n = 64, height: r = 16, fill: i = !1, className: a }) {
	let o = Ii(e, n, r);
	return o ? /* @__PURE__ */ (0, M.jsx)("svg", {
		role: "img",
		"aria-label": t,
		viewBox: `0 0 ${n} ${r}`,
		...i ? { preserveAspectRatio: "none" } : {
			width: n,
			height: r
		},
		className: C(i ? "w-full" : "shrink-0", "overflow-visible text-primary", a),
		style: i ? { height: r } : void 0,
		children: /* @__PURE__ */ (0, M.jsx)("polyline", {
			points: o,
			fill: "none",
			stroke: "currentColor",
			strokeWidth: 1.25,
			strokeLinejoin: "round",
			...i ? { vectorEffect: "non-scaling-stroke" } : {}
		})
	}) : null;
}
function Zi({ label: e, values: t, format: n, className: r }) {
	let i = Math.max(...t, 0), a = t[0];
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: C("rounded-lg border border-border bg-muted/50 px-2.5 py-density-2", r),
		"data-sparkline": e,
		children: [/* @__PURE__ */ (0, M.jsxs)("div", {
			className: "flex items-center justify-between text-[10px] uppercase tracking-wide text-muted-foreground",
			children: [/* @__PURE__ */ (0, M.jsx)("span", { children: e }), /* @__PURE__ */ (0, M.jsx)("span", { children: t.length > 0 ? `max ${n(i)}` : "" })]
		}), t.length < 2 ? /* @__PURE__ */ (0, M.jsx)("div", {
			className: "mt-1 text-[11px] text-muted-foreground",
			children: a === void 0 ? "No samples" : `${n(a)} (1 sample)`
		}) : /* @__PURE__ */ (0, M.jsx)(Xi, {
			fill: !0,
			values: t,
			label: `${e} over time`,
			width: 160,
			height: 36,
			className: "mt-1"
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/traces/ProcessUsageBars.tsx
function Qi({ resource: e, pid: t, value: n, percent: r, title: i }) {
	let a = `${e} for pid ${t}`;
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "w-28 shrink-0 text-xs tabular-nums",
		title: i,
		children: [/* @__PURE__ */ (0, M.jsxs)("div", {
			className: "mb-0.5 flex justify-between gap-1 text-muted-foreground",
			children: [/* @__PURE__ */ (0, M.jsx)("span", { children: e }), /* @__PURE__ */ (0, M.jsx)("span", { children: n })]
		}), r == null ? /* @__PURE__ */ (0, M.jsx)("div", {
			role: "progressbar",
			"aria-label": a,
			"aria-valuetext": "Waiting for sample",
			"aria-valuemin": 0,
			"aria-valuemax": 100,
			className: "h-1 rounded-full bg-muted"
		}) : /* @__PURE__ */ (0, M.jsx)(de, {
			label: a,
			total: 100,
			height: "h-1",
			segments: [{
				count: Math.max(0, Math.min(100, r)),
				color: "bg-primary",
				label: n
			}]
		})]
	});
}
function $i({ row: e, memoryPercent: t }) {
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex shrink-0 gap-3",
		children: [/* @__PURE__ */ (0, M.jsx)(Qi, {
			resource: "CPU",
			pid: e.pid,
			value: Mi(e.cpuPercent),
			percent: e.cpuPercent,
			title: `Peak ${Mi(e.peakCpuPercent)} · CPU per core`
		}), /* @__PURE__ */ (0, M.jsx)(Qi, {
			resource: "RAM",
			pid: e.pid,
			value: Xe(e.rssBytes),
			percent: t,
			title: `Peak ${Xe(e.peakRssBytes)}${t === void 0 ? "" : ` · ${t.toFixed(1)}% of host RAM`}`
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/traces/ProcessTraceTree.tsx
var ea = 2, ta = () => {};
function na({ node: e, selected: t, series: n, showUsageBars: r, memoryPercent: i }) {
	let { row: a, exited: o } = e;
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		"data-process-key": a.key,
		...o ? { "data-exited": "true" } : {},
		className: C("flex min-w-0 flex-1 items-center gap-1.5", r && "flex-wrap sm:flex-nowrap", o && "opacity-60"),
		title: a.command,
		children: [
			/* @__PURE__ */ (0, M.jsx)(w, {
				icon: a.isRoot ? on : f,
				className: a.isRoot ? "text-base text-blue-600" : "text-base text-muted-foreground"
			}),
			/* @__PURE__ */ (0, M.jsx)("span", {
				className: C("truncate", t ? "font-semibold text-primary" : "font-medium text-foreground"),
				children: Tr({
					pid: a.pid,
					name: a.name,
					command: a.command
				})
			}),
			/* @__PURE__ */ (0, M.jsxs)("span", {
				className: "shrink-0 text-xs text-muted-foreground",
				children: ["pid ", a.pid]
			}),
			/* @__PURE__ */ (0, M.jsx)("span", {
				className: C("min-w-0 flex-1 truncate font-mono text-xs text-muted-foreground", r && "order-last basis-full sm:order-none sm:basis-auto"),
				title: a.command,
				"data-process-command": "",
				children: a.command
			}),
			n ? /* @__PURE__ */ (0, M.jsx)(Xi, {
				values: n,
				label: `CPU % over ${n.length} samples`
			}) : null,
			/* @__PURE__ */ (0, M.jsxs)("span", {
				className: C("inline-flex shrink-0 items-center gap-1 text-xs", wr(a.status)),
				children: [/* @__PURE__ */ (0, M.jsx)(w, { icon: Cr(a.status) }), a.status]
			}),
			r ? /* @__PURE__ */ (0, M.jsx)($i, {
				row: a,
				memoryPercent: i
			}) : o ? /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
				/* @__PURE__ */ (0, M.jsx)("span", {
					className: "shrink-0 text-xs tabular-nums text-muted-foreground",
					title: "Lifetime",
					children: Ge(Oi(a))
				}),
				/* @__PURE__ */ (0, M.jsx)("span", {
					className: "shrink-0 text-xs tabular-nums text-muted-foreground",
					title: "Peak CPU",
					children: `peak ${Mi(a.peakCpuPercent)}`
				}),
				/* @__PURE__ */ (0, M.jsx)("span", {
					className: "shrink-0 text-xs tabular-nums text-muted-foreground",
					title: "Peak RSS",
					children: `peak ${Xe(a.peakRssBytes)}`
				})
			] }) : /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, M.jsx)("span", {
				className: "shrink-0 text-xs tabular-nums text-muted-foreground",
				title: `Peak ${Mi(a.peakCpuPercent)}`,
				children: Mi(a.cpuPercent)
			}), /* @__PURE__ */ (0, M.jsx)("span", {
				className: "shrink-0 text-xs tabular-nums text-muted-foreground",
				title: `Peak ${Xe(a.peakRssBytes)}`,
				children: Xe(a.rssBytes)
			})] })
		]
	});
}
function ra({ nodes: e, samples: t, onSelect: n, selectedKey: r, emptyLabel: i = "No processes recorded", className: a, expandAll: o, onExpandAllChange: s, hideExited: c = !1, onHideExitedChange: l, loading: u = !1, error: d, showUsageBars: f = !1 }) {
	let p = (0, j.useMemo)(() => Ei(e), [e]), m = (0, j.useMemo)(() => c ? Di(p) : p, [p, c]), h = (0, j.useMemo)(() => Pi(t ?? []), [t]), g = (0, j.useMemo)(() => {
		let e = /* @__PURE__ */ new Map();
		for (let n of t ?? []) (!e.has(n.key) || e.get(n.key).tick < n.tick) && e.set(n.key, n);
		return e;
	}, [t]), _ = (0, j.useMemo)(() => r ? ji(m).find((e) => e.row.key === r) ?? null : null, [m, r]), v = e.length === 0 ? d ? null : u ? /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "p-density-6 text-center text-sm text-muted-foreground",
		role: "status",
		children: [/* @__PURE__ */ (0, M.jsx)(w, {
			icon: ht,
			className: "animate-spin text-3xl text-blue-500"
		}), /* @__PURE__ */ (0, M.jsx)("p", {
			className: "mt-2",
			children: "Waiting for process trace…"
		})]
	}) : /* @__PURE__ */ (0, M.jsx)("div", {
		className: "p-density-6 text-center text-sm text-muted-foreground",
		children: i
	}) : /* @__PURE__ */ (0, M.jsx)("div", {
		className: "p-density-6 text-center text-sm text-muted-foreground",
		children: "Every process has exited"
	}), y = o !== void 0 && s === void 0;
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: C("flex min-h-0 flex-col", a),
		children: [
			l && /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex items-center justify-between gap-2 border-b border-border bg-background px-3 py-1.5 text-xs text-muted-foreground",
				"data-process-toolbar": "",
				children: [/* @__PURE__ */ (0, M.jsxs)("span", { children: [
					e.length,
					" ",
					e.length === 1 ? "process" : "processes"
				] }), /* @__PURE__ */ (0, M.jsxs)("label", {
					className: "inline-flex cursor-pointer items-center gap-1.5",
					children: [/* @__PURE__ */ (0, M.jsx)("input", {
						type: "checkbox",
						checked: c,
						onChange: (e) => l(e.target.checked)
					}), "Hide exited"]
				})]
			}),
			d && /* @__PURE__ */ (0, M.jsx)("div", {
				className: "whitespace-pre-wrap px-3 py-2 text-xs text-red-700 dark:text-red-400",
				role: "alert",
				children: d
			}),
			/* @__PURE__ */ (0, M.jsx)(qe, {
				roots: m,
				className: "min-h-0 flex-1",
				ariaLabel: "Process tree",
				empty: v,
				showControls: !y,
				...o === void 0 ? {} : {
					expandAll: o,
					onExpandAllChange: s ?? ta
				},
				...o === void 0 && s ? { onExpandAllChange: s } : {},
				getChildren: (e) => e.children,
				getKey: (e) => e.row.key,
				getSearchText: (e) => `${e.row.name} ${e.row.command} ${e.row.pid}`,
				selected: _,
				revealSelected: !0,
				defaultOpen: (e, t) => t < ea,
				onSelect: (e) => n(e.row.key),
				renderRow: ({ node: e, selected: t }) => /* @__PURE__ */ (0, M.jsx)(na, {
					node: e,
					selected: t,
					series: h.get(e.row.key),
					showUsageBars: f,
					memoryPercent: g.get(e.row.key)?.memoryPercent
				}),
				rowClass: (e, t) => t ? "bg-primary/10 border-l-2 border-primary" : "hover:bg-accent"
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/traces/sql-trace.ts
function ia(e) {
	return e >= 100 ? Ge(e) : `${Number(e.toFixed(1))} ms`;
}
function aa(e) {
	let t = e.filter((e) => e.rows >= 0);
	return t.length === 0 ? -1 : t.reduce((e, t) => e + t.rows, 0);
}
function oa(e, t) {
	let [n] = t;
	return {
		id: e,
		startedAt: n.startedAt,
		op: n.op,
		tables: n.tables,
		sql: n.sql,
		durationMs: t.reduce((e, t) => e + t.durationMs, 0),
		rows: aa(t),
		error: t.find((e) => e.error)?.error ?? "",
		failed: t.filter((e) => e.error).length,
		count: t.length,
		occurrences: t
	};
}
function sa(e, t) {
	let n = /* @__PURE__ */ new Map();
	return e.forEach((e, r) => {
		let i = t.group && e.repeatKey ? `repeat:${e.repeatKey}` : `row:${r}`;
		n.set(i, [...n.get(i) ?? [], e]);
	}), [...n].map(([e, t]) => oa(e, t));
}
// ../../../clicky-ui/packages/ui/src/data/traces/SqlTraceTable.tsx
var ca = [
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
		render: (e) => /* @__PURE__ */ (0, M.jsx)(D, {
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
		render: (e, t) => /* @__PURE__ */ (0, M.jsxs)("span", {
			className: "flex min-w-0 items-center gap-density-2",
			children: [/* @__PURE__ */ (0, M.jsx)(Be, {
				inline: !0,
				language: "sql",
				source: t.sql
			}), t.count > 1 ? /* @__PURE__ */ (0, M.jsx)("span", {
				title: `${t.count} executions`,
				children: /* @__PURE__ */ (0, M.jsx)(D, {
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
		render: (e, t) => /* @__PURE__ */ (0, M.jsx)("span", {
			className: "tabular-nums",
			title: t.count > 1 ? `${t.count} executions in total` : void 0,
			children: ia(Number(e))
		}),
		sortValue: (e) => Number(e)
	},
	{
		key: "rows",
		label: "Rows",
		align: "right",
		shrink: !0,
		render: (e) => Number(e) < 0 ? /* @__PURE__ */ (0, M.jsx)("span", {
			className: "text-muted-foreground",
			children: "—"
		}) : /* @__PURE__ */ (0, M.jsx)("span", {
			className: "tabular-nums",
			children: String(e)
		}),
		sortValue: (e) => Number(e)
	},
	{
		key: "error",
		label: "Error",
		grow: !0,
		render: (e, t) => e ? /* @__PURE__ */ (0, M.jsxs)("span", {
			className: "truncate text-destructive",
			title: String(e),
			children: [String(e), t.count > 1 ? ` (${t.failed} of ${t.count} failed)` : ""]
		}) : null
	}
];
function la({ row: e }) {
	let [t] = e.occurrences;
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "space-y-density-3",
		children: [
			/* @__PURE__ */ (0, M.jsx)(Be, {
				language: "sql",
				source: e.sql,
				copyable: !0
			}),
			e.error ? /* @__PURE__ */ (0, M.jsxs)("div", { children: [/* @__PURE__ */ (0, M.jsx)("div", {
				className: "mb-1 font-semibold text-foreground",
				children: "Error"
			}), /* @__PURE__ */ (0, M.jsx)("pre", {
				className: "whitespace-pre-wrap break-all text-destructive",
				children: e.error
			})] }) : null,
			/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "text-xs text-muted-foreground",
				children: [
					t.connection,
					" · ",
					t.driver,
					t.truncated?.length ? ` · truncated: ${t.truncated.join(", ")}` : ""
				]
			}),
			e.count > 1 ? /* @__PURE__ */ (0, M.jsx)("ul", {
				className: "space-y-1 text-xs",
				"aria-label": "Executions",
				children: e.occurrences.map((e, t) => /* @__PURE__ */ (0, M.jsxs)("li", {
					className: C("flex items-baseline gap-density-3", e.error && "text-destructive"),
					children: [
						/* @__PURE__ */ (0, M.jsx)("span", {
							className: "tabular-nums text-muted-foreground",
							children: e.startedAt
						}),
						/* @__PURE__ */ (0, M.jsx)("span", {
							className: "tabular-nums",
							children: ia(e.durationMs)
						}),
						/* @__PURE__ */ (0, M.jsx)("span", {
							className: "tabular-nums",
							children: e.rows < 0 ? "— rows" : `${e.rows} rows`
						}),
						e.args ? /* @__PURE__ */ (0, M.jsx)("code", {
							className: "font-mono",
							children: JSON.stringify(e.args)
						}) : null,
						e.error ? /* @__PURE__ */ (0, M.jsx)("span", { children: e.error }) : null
					]
				}, t))
			}) : t.args ? /* @__PURE__ */ (0, M.jsxs)("div", { children: [/* @__PURE__ */ (0, M.jsx)("div", {
				className: "mb-1 font-semibold text-foreground",
				children: "Arguments"
			}), /* @__PURE__ */ (0, M.jsx)(Je, { data: t.args })] }) : null
		]
	});
}
function ua({ rows: e, groupRepeated: t = !0, emptyLabel: n = "No SQL statements recorded", className: r }) {
	let [i, a] = (0, j.useState)(t), o = (0, j.useMemo)(() => sa(e, { group: i }), [e, i]);
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: C("flex h-full min-h-0 flex-col gap-density-2", r),
		children: [/* @__PURE__ */ (0, M.jsx)(Fe, {
			checked: i,
			onChange: a,
			label: "Group repeated statements",
			className: "self-start px-density-2"
		}), /* @__PURE__ */ (0, M.jsx)(g, {
			data: o,
			columns: ca,
			autoFilter: !0,
			emptyMessage: n,
			globalFilterPlaceholder: "Filter SQL, table, or error…",
			getRowId: (e) => e.id,
			renderExpandedRow: (e) => /* @__PURE__ */ (0, M.jsx)(la, { row: e })
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/traces/cast.ts
var N = -1, P = /* @__PURE__ */ new Set(["i", "m"]);
function da(e, t) {
	try {
		return JSON.parse(e);
	} catch (e) {
		throw Error(`cast ${t} is not valid JSON: ${e.message}`);
	}
}
function fa(e) {
	let t = da(e, "header");
	if (typeof t != "object" || !t || Array.isArray(t)) throw Error("cast header is not a JSON object");
	if (t.version !== 2) throw Error(`cast header: only asciinema version 2 is supported, got ${String(t.version)}`);
	if (typeof t.width != "number" || typeof t.height != "number") throw Error("cast header has no numeric width and height");
	return {
		...t,
		width: t.width,
		height: t.height
	};
}
function pa(e, t) {
	let n = typeof e == "string" ? /^(\d+)x(\d+)$/.exec(e) : null;
	if (!n) throw Error(`cast line ${t}: resize data must be "<width>x<height>", got ${JSON.stringify(e)}`);
	return {
		width: Number(n[1]),
		height: Number(n[2])
	};
}
function ma(e) {
	let [t, ...n] = e.split("\n").filter((e) => e.trim() !== "");
	if (t === void 0) throw Error("cast text is empty");
	let r = fa(t), i = {
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
		if (o = Math.max(o, s), !P.has(c)) {
			if (c === "o") a.push({
				offset: s,
				kind: "o",
				data: String(l),
				...i
			});
			else if (c === "r") i = pa(l, n), a.push({
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
		...s.exit_code !== void 0 && s.exit_code !== N ? { exitCode: s.exit_code } : {},
		...s.truncated ? { truncated: !0 } : {},
		...r.command ? { command: r.command } : {}
	}), a;
}
// ../../../clicky-ui/packages/ui/src/data/traces/terminal-screen.ts
var ha = {
	ch: " ",
	sgr: ""
}, ga = 8, _a = 4096, va = (e, t, n) => Math.min(Math.max(e, t), n);
function ya(e, t) {
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
function ba(e, t, n) {
	if (!(t > 0) || !(n > 0)) throw Error(`terminal screen needs a positive size, got ${t}x${n}`);
	if (n < e.rows.length) {
		let t = Math.max(0, e.row - n + 1);
		e.rows.splice(0, t), e.row -= t, e.rows.length = n;
	}
	for (; e.rows.length < n;) e.rows.push([]);
	for (let n of e.rows) n.length > t && (n.length = t);
	e.width = t, e.height = n, e.row = va(e.row, 0, n - 1), e.col = va(e.col, 0, t - 1);
}
function xa(e) {
	e.row < e.height - 1 ? e.row++ : (e.rows.shift(), e.rows.push([]));
}
function Sa(e, t) {
	e.col >= e.width && (e.col = 0, xa(e));
	let n = e.rows[e.row];
	for (; n.length < e.col;) n.push(ha);
	n[e.col] = {
		ch: t,
		sgr: e.sgr
	}, e.col++;
}
function Ca(e, t) {
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
	for (; n.length < r;) n.push(ha);
	for (let e = 0; e < r; e++) n[e] = ha;
}
function wa(e, t) {
	if (t === 0) {
		Ca(e, 0);
		for (let t = e.row + 1; t < e.height; t++) e.rows[t] = [];
	} else if (t === 1) {
		for (let t = 0; t < e.row; t++) e.rows[t] = [];
		Ca(e, 1);
	} else for (let t = 0; t < e.height; t++) e.rows[t] = [];
}
function Ta(e, t) {
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
function Ea(e, t, n) {
	if (/^[?>=<!]/.test(t)) return;
	let r = t === "" ? [] : t.split(";"), i = (e, t) => {
		let n = Number.parseInt(r[e] ?? "", 10);
		return Number.isNaN(n) || n === 0 ? t : n;
	}, a = Number.parseInt(r[0] ?? "0", 10) || 0, o = e.rows[e.row];
	switch (n) {
		case "m":
			Ta(e, r);
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
			e.col = va(i(0, 1) - 1, 0, e.width - 1);
			break;
		case "d":
			e.row = va(i(0, 1) - 1, 0, e.height - 1);
			break;
		case "H":
		case "f":
			e.row = va(i(0, 1) - 1, 0, e.height - 1), e.col = va(i(1, 1) - 1, 0, e.width - 1);
			break;
		case "J":
			wa(e, a);
			break;
		case "K":
			Ca(e, a);
			break;
		case "P":
			o.splice(e.col, i(0, 1));
			break;
		case "X":
			for (let t = 0; t < i(0, 1) && e.col + t < o.length; t++) o[e.col + t] = ha;
			break;
		case "@":
			o.length > e.col && o.splice(e.col, 0, ...Array.from({ length: i(0, 1) }, () => ha)), o.length = Math.min(o.length, e.width);
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
function Da(e, t, n) {
	let r = t[n + 1];
	if (r === void 0) return -1;
	if (r === "[") {
		let r = n + 2;
		for (; r < t.length && /[\x20-\x3f]/.test(t[r]);) r++;
		return r >= t.length ? -1 : (Ea(e, t.slice(n + 2, r).replace(/[\x20-\x2f]/g, ""), t[r]), r + 1);
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
function Oa(e, t) {
	let n = e.pending + t;
	e.pending = "";
	let r = 0;
	for (; r < n.length;) {
		let t = n[r];
		if (t === "\x1B") {
			let t = Da(e, n, r);
			if (t === -1) {
				if (n.length - r <= _a) {
					e.pending = n.slice(r);
					return;
				}
				r++;
				continue;
			}
			r = t;
		} else if (r++, t === "\r") e.col = 0;
		else if (t === "\n") xa(e);
		else if (t === "\b") e.col = Math.max(0, e.col - 1);
		else if (t === "	") e.col = Math.min(e.width - 1, (Math.floor(e.col / ga) + 1) * ga);
		else if (t < " " || t === "") continue;
		else {
			let t = n.codePointAt(r - 1), i = String.fromCodePoint(t);
			r += i.length - 1, Sa(e, i);
		}
	}
}
function ka(e) {
	let t = e.length;
	for (; t > 0 && e[t - 1].ch === " ";) t--;
	return e.slice(0, t);
}
function Aa(e, t) {
	return t === "" ? "\x1B[0m" : e === "" ? `\u001b[${t}m` : `\u001b[0;${t}m`;
}
function ja(e) {
	let t = "";
	return e.rows.map((e) => {
		let n = "";
		for (let r of ka(e)) r.sgr !== t && (n += Aa(t, r.sgr), t = r.sgr), n += r.ch;
		return n;
	}).join("\n");
}
// ../../../clicky-ui/packages/ui/src/data/traces/terminal-timeline.ts
var Ma = [
	"o",
	"r",
	"snapshot",
	"exit"
], Na = {
	width: 80,
	height: 24
};
function Pa(e) {
	e.forEach((e, t) => {
		if (!Ma.includes(e.kind)) throw Error(`terminal event ${t} has unknown kind "${e.kind}"`);
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
		width: n?.width ?? Na.width,
		height: n?.height ?? Na.height
	};
}
function Fa(e, t) {
	let n = 0, r = e.length;
	for (; n < r;) {
		let i = n + r >>> 1;
		e[i].offset <= t ? n = i + 1 : r = i;
	}
	return n;
}
function Ia(e, t, n) {
	let r = e !== void 0 && e.applied <= n, i = r ? e.screen : ya(t.width, t.height);
	for (let a = r ? e.applied : 0; a < n; a++) {
		let e = t.steps[a];
		e.kind === "r" ? ba(i, e.width, e.height) : Oa(i, e.data);
	}
	return r && e.applied === n ? e : {
		screen: i,
		applied: n
	};
}
// ../../../clicky-ui/packages/ui/src/data/traces/TerminalCastPlayer.tsx
var La = 50, Ra = [
	"1",
	"2",
	"4"
].map((e) => ({
	id: e,
	label: `${e}x`
})), za = [{
	id: "replay",
	label: "Replay"
}, {
	id: "settled",
	label: "Settled screen"
}];
function Ba(e, t) {
	if (e === void 0 == (t === void 0)) throw Error("TerminalCastPlayer takes either events or castText, not both or neither");
	return e ?? ma(t);
}
function Va(e) {
	return `${e.toFixed(1)}s`;
}
function Ha({ exit: e }) {
	return e.exitCode === void 0 ? /* @__PURE__ */ (0, M.jsx)(D, {
		variant: "outline",
		tone: "neutral",
		size: "sm",
		children: "exited"
	}) : /* @__PURE__ */ (0, M.jsx)(D, {
		variant: "soft",
		tone: e.exitCode === 0 ? "success" : "danger",
		size: "sm",
		className: "font-mono",
		children: `exit ${e.exitCode}`
	});
}
function Ua(e, t) {
	let n = (0, j.useRef)(void 0), r = Fa(e.steps, t);
	return (0, j.useMemo)(() => {
		let t = Ia(n.current?.timeline === e ? n.current.replay : void 0, e, r);
		return n.current = {
			timeline: e,
			replay: t
		}, {
			ansi: ja(t.screen),
			cols: t.screen.width,
			rows: t.screen.height
		};
	}, [e, r]);
}
function Wa({ events: e, castText: t, autoPlay: n = !1, emptyLabel: r = "No terminal output recorded", className: i }) {
	let a = (0, j.useMemo)(() => Pa(Ba(e, t)), [e, t]), [o, s] = (0, j.useState)(0), [c, u] = (0, j.useState)(n), [d, f] = (0, j.useState)("1"), [p, m] = (0, j.useState)("replay"), h = Ua(a, o);
	if ((0, j.useEffect)(() => {
		if (!c) return;
		let e = Date.now(), t = setInterval(() => {
			let t = Date.now(), n = (t - e) / 1e3;
			e = t, s((e) => Math.min(a.duration, e + n * Number(d)));
		}, La);
		return () => clearInterval(t);
	}, [
		c,
		d,
		a.duration
	]), (0, j.useEffect)(() => {
		c && o >= a.duration && u(!1);
	}, [
		c,
		o,
		a.duration
	]), a.steps.length === 0 && a.settled === void 0) return /* @__PURE__ */ (0, M.jsx)("div", {
		className: "p-density-6 text-center text-muted-foreground text-sm",
		children: r
	});
	let g = a.settled !== void 0, _ = p === "settled" && g;
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: C("flex min-h-0 flex-col gap-density-2", i),
		children: [
			/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex flex-wrap items-center gap-density-2 text-xs text-muted-foreground",
				children: [
					a.exit?.command ? /* @__PURE__ */ (0, M.jsx)("code", {
						className: "font-mono text-foreground",
						children: a.exit.command
					}) : null,
					a.exit ? /* @__PURE__ */ (0, M.jsx)(Ha, { exit: a.exit }) : null,
					a.exit?.truncated ? /* @__PURE__ */ (0, M.jsx)("span", {
						title: "Size caps or a full buffer stopped some output events being recorded",
						children: /* @__PURE__ */ (0, M.jsx)(D, {
							variant: "soft",
							tone: "warning",
							size: "sm",
							children: "output truncated"
						})
					}) : null,
					/* @__PURE__ */ (0, M.jsx)("span", {
						className: "tabular-nums",
						children: `${a.width}x${a.height}`
					}),
					/* @__PURE__ */ (0, M.jsx)("span", { className: "flex-1" }),
					g && a.steps.length > 0 ? /* @__PURE__ */ (0, M.jsx)(ge, {
						size: "sm",
						"aria-label": "Screen view",
						value: p,
						options: za,
						onChange: m
					}) : null
				]
			}),
			/* @__PURE__ */ (0, M.jsx)("div", {
				role: "region",
				"aria-label": "Terminal screen",
				"data-cols": _ ? void 0 : h.cols,
				className: "overflow-auto rounded border bg-black p-density-2",
				children: /* @__PURE__ */ (0, M.jsx)(l, {
					text: _ ? a.settled : h.ansi,
					className: "font-mono text-xs leading-5 whitespace-pre text-gray-100"
				})
			}),
			a.steps.length > 0 && !_ ? /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex items-center gap-density-2",
				children: [
					/* @__PURE__ */ (0, M.jsx)(fe, {
						variant: "ghost",
						size: "icon",
						"aria-label": c ? "Pause" : "Play",
						onClick: c ? () => u(!1) : () => {
							o >= a.duration && s(0), u(!0);
						},
						children: /* @__PURE__ */ (0, M.jsx)(w, { icon: c ? nn : tn })
					}),
					/* @__PURE__ */ (0, M.jsx)("input", {
						type: "range",
						"aria-label": "Seek",
						className: "min-w-0 flex-1 accent-primary",
						min: 0,
						max: a.duration,
						step: "any",
						value: o,
						onChange: (e) => s(Number(e.target.value))
					}),
					/* @__PURE__ */ (0, M.jsx)("span", {
						className: "shrink-0 text-xs tabular-nums text-muted-foreground",
						children: `${Va(o)} / ${Va(a.duration)}`
					}),
					/* @__PURE__ */ (0, M.jsx)(ge, {
						size: "sm",
						"aria-label": "Playback speed",
						value: d,
						options: Ra,
						onChange: f
					}),
					/* @__PURE__ */ (0, M.jsx)(fe, {
						variant: "ghost",
						size: "icon",
						"aria-label": "Jump to end",
						onClick: () => {
							u(!1), s(a.duration);
						},
						children: /* @__PURE__ */ (0, M.jsx)(w, { icon: pt })
					})
				]
			}) : null
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/traces/har-adapter.ts
var Ga = "request.postData.text", Ka = "response.content.text";
function qa(e) {
	if (e != null) return typeof e == "string" ? e : JSON.stringify(e, null, 2);
}
function Ja(e, t) {
	return t || (e != null && typeof e != "string" ? "application/json" : void 0);
}
function Ya(e) {
	return e ? { headers: e } : {};
}
function Xa(e, t, n) {
	if (typeof e != "object" || !e) throw Error(`HTTP exchange ${n.method ?? "?"} ${n.url ?? "?"} has no ${t} object`);
	return e;
}
function Za(e, t) {
	let n = qa(e.postData?.text);
	return {
		method: e.method,
		url: e.url,
		...e.httpVersion ? { httpVersion: e.httpVersion } : {},
		...Ya(e.headers),
		...e.queryString ? { queryString: e.queryString } : {},
		...n === void 0 ? {} : { postData: {
			mimeType: Ja(e.postData?.text, e.postData?.mimeType) ?? "",
			text: n
		} },
		...e.headersSize === void 0 ? {} : { headersSize: e.headersSize },
		...t === void 0 ? {} : { bodySize: t }
	};
}
function Qa(e, t, n, r, i) {
	return {
		status: e.status,
		...e.statusText ? { statusText: e.statusText } : {},
		...e.httpVersion ? { httpVersion: e.httpVersion } : {},
		...Ya(e.headers),
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
function $a(e) {
	return `Tunnelled CONNECT relayed without decrypting it: ${Xe(e.bytesIn ?? 0)} in from the server, ${Xe(e.bytesOut ?? 0)} out from the client.`;
}
function eo(e, t = {}) {
	let n = Xa(e.request, "request", e), r = Xa(e.response, "response", e), i = new Set(e.truncated ?? []), a = e.tunnelled === !0, o = a ? $a(e) : qa(r.content?.text), s = a ? "text/plain" : Ja(r.content?.text, r.content?.mimeType), c = a ? e.bytesIn ?? 0 : r.bodySize ?? r.content?.size ?? 0;
	return {
		startedDateTime: e.startedAt,
		time: e.durationMs,
		request: Za(n, a ? e.bytesOut ?? 0 : n.bodySize),
		response: Qa(r, o, s, c, i.has(Ka)),
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
		...i.has(Ga) ? { _requestBodyTruncated: !0 } : {}
	};
}
// ../../../clicky-ui/packages/ui/src/data/traces/TraceHarView.tsx
function to({ rows: e, emptyLabel: t = "No HTTP exchanges recorded", ...n }) {
	let r = (0, j.useMemo)(() => e.map((e, t) => eo(e, { id: `${e.startedAt}#${t}` })), [e]);
	return /* @__PURE__ */ (0, M.jsx)(ri, {
		entries: r,
		emptyLabel: t,
		...n
	});
}
// ../../../clicky-ui/packages/ui/src/data/traces/TraceView.tsx
function no(e, t, n) {
	t.forEach((t, r) => {
		if (t[n] === void 0 || t[n] === null) throw Error(`trace kind "${e}" row ${r} has no "${n}"`);
	});
}
function ro({ kind: e, rows: t }) {
	let n = (0, j.useMemo)(() => o(t, { literalKeys: !0 }), [t]);
	return /* @__PURE__ */ (0, M.jsx)(g, {
		data: t,
		columns: n,
		autoFilter: !0,
		emptyMessage: `No ${e} records`
	});
}
function io({ rows: e, samples: t }) {
	let [n, r] = (0, j.useState)(null);
	return /* @__PURE__ */ (0, M.jsx)(ra, {
		nodes: e,
		selectedKey: n,
		onSelect: r,
		...t ? { samples: t } : {}
	});
}
function ao({ kind: e, rows: t, samples: n, className: r }) {
	let i = (() => {
		switch (e) {
			case "http":
			case "http_proxy": return no(e, t, "request"), no(e, t, "response"), /* @__PURE__ */ (0, M.jsx)(to, { rows: t });
			case "pg_proxy":
			case "sql_statements": return no(e, t, "sql"), /* @__PURE__ */ (0, M.jsx)(ua, { rows: t });
			case "terminal": return no(e, t, "kind"), no(e, t, "offset"), /* @__PURE__ */ (0, M.jsx)(Wa, { events: t });
			case "process_tree": return no(e, t, "key"), n && no("process_samples", n, "key"), /* @__PURE__ */ (0, M.jsx)(io, {
				rows: t,
				samples: n
			});
			case "goprof": return no(e, t, "key"), no(e, t, "checkpointId"), /* @__PURE__ */ (0, M.jsx)(Ji, { rows: t });
			case "stacks": return no(e, t, "key"), no(e, t, "checkpointId"), /* @__PURE__ */ (0, M.jsx)(Yi, { rows: t });
			default: return /* @__PURE__ */ (0, M.jsx)(ro, {
				kind: e,
				rows: t
			});
		}
	})();
	return /* @__PURE__ */ (0, M.jsx)("div", {
		className: C("flex h-full min-h-0 flex-col", r),
		children: i
	});
}
// ../../../clicky-ui/packages/ui/src/data/traces/process-facets.ts
var oo = {
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
}, so = [
	"openFiles",
	"sockets",
	"ports",
	"connections"
], co = (e, t) => e.filter((e) => e.key === t).sort((e, t) => e.tick - t.tick);
function lo(e) {
	return so.filter((t) => e.some((e) => e[oo[t].countKey] !== void 0));
}
function uo(e, t, n) {
	let { countKey: r } = oo[n];
	return co(e, t).flatMap((e) => e[r] === void 0 ? [] : [e[r]]);
}
function fo(e, t, n) {
	let { countKey: r, listKey: i } = oo[n], a;
	for (let n of co(e, t)) {
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
function po(e, t) {
	let n = co(e, t);
	return {
		read: n.flatMap((e) => e.diskReadBytes === void 0 ? [] : [e.diskReadBytes]),
		write: n.flatMap((e) => e.diskWriteBytes === void 0 ? [] : [e.diskWriteBytes])
	};
}
// ../../../clicky-ui/packages/ui/src/data/traces/ProcessFacetTable.tsx
var mo = [
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
], ho = {
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
	sockets: mo,
	connections: mo,
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
}, go = [{
	key: "name",
	label: "Name",
	sortable: !0
}, {
	key: "value",
	label: "Value",
	grow: !0
}], _o = {
	autoFilter: !0,
	showThemeControl: !1,
	showFullscreenControl: !1
};
function vo(e) {
	if (e.facet === "env") {
		let t = Object.entries(e.environment ?? {}).map(([e, t]) => ({
			name: e,
			value: t
		})).sort((e, t) => e.name.localeCompare(t.name));
		return /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "space-y-2",
			"data-process-facet": "env",
			children: [e.error && /* @__PURE__ */ (0, M.jsx)("div", {
				className: "whitespace-pre-wrap text-xs text-red-700 dark:text-red-400",
				role: "alert",
				children: e.error
			}), e.environment && /* @__PURE__ */ (0, M.jsx)(g, {
				data: t,
				columns: go,
				getRowId: (e) => e.name,
				emptyMessage: "The process has no environment variables",
				..._o
			})]
		});
	}
	let { reading: t, facet: n } = e;
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "space-y-2",
		"data-process-facet": n,
		children: [/* @__PURE__ */ (0, M.jsx)("div", {
			className: "text-xs text-muted-foreground",
			children: t ? `Read at ${ki(t.readAt)}` : "No reading stored for this process yet"
		}), t && /* @__PURE__ */ (0, M.jsx)(g, {
			data: t.rows,
			columns: ho[n],
			getRowId: (e, t) => `${e.fd}:${t}`,
			emptyMessage: oo[n].empty,
			..._o
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/traces/ProcessTraceDetail.tsx
var yo = {
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
function bo(e) {
	let t = yo[e];
	if (!t) throw Error(`unknown stack transport "${e}"; expected one of ${Object.keys(yo).join(", ")}`);
	return t;
}
var xo = [];
function So({ title: e, children: t }) {
	return /* @__PURE__ */ (0, M.jsxs)("section", { children: [/* @__PURE__ */ (0, M.jsx)("div", {
		className: "mb-1.5 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground",
		children: e
	}), t] });
}
function Co({ label: e, value: t, wide: n }) {
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: `rounded-lg border border-border bg-muted/50 px-2.5 py-density-2 ${n ? "col-span-2" : ""}`,
		children: [/* @__PURE__ */ (0, M.jsx)("div", {
			className: "text-[10px] uppercase tracking-wide text-muted-foreground",
			children: e
		}), /* @__PURE__ */ (0, M.jsx)("div", {
			className: "mt-0.5 break-all text-xs font-medium text-foreground",
			children: t
		})]
	});
}
function wo({ node: e }) {
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "min-w-0",
		children: [/* @__PURE__ */ (0, M.jsxs)("div", {
			className: "flex items-center gap-density-2",
			children: [/* @__PURE__ */ (0, M.jsx)(w, {
				icon: e.isRoot ? on : f,
				className: "shrink-0 text-2xl text-blue-600"
			}), /* @__PURE__ */ (0, M.jsx)("h2", {
				className: "break-words text-lg font-bold text-foreground",
				children: Tr(e)
			})]
		}), /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "mt-1 flex flex-wrap items-center gap-density-2 text-xs text-muted-foreground",
			children: [
				/* @__PURE__ */ (0, M.jsxs)("span", {
					className: "font-mono",
					children: ["pid ", e.pid]
				}),
				e.ppid ? /* @__PURE__ */ (0, M.jsxs)("span", {
					className: "font-mono",
					children: ["ppid ", e.ppid]
				}) : null,
				/* @__PURE__ */ (0, M.jsxs)("span", {
					className: `inline-flex items-center gap-1 rounded-full bg-muted px-2 py-0.5 ${wr(e.status)}`,
					children: [/* @__PURE__ */ (0, M.jsx)(w, { icon: Cr(e.status) }), e.status]
				})
			]
		})]
	});
}
function To({ transports: e, hasStack: t, collecting: n, onCollect: r }) {
	return /* @__PURE__ */ (0, M.jsx)("div", {
		className: "flex items-center gap-density-1",
		"data-stack-actions": "",
		children: e.map((e) => {
			let i = bo(e.transport);
			return /* @__PURE__ */ (0, M.jsxs)(fe, {
				size: "sm",
				variant: e.destructive ? "destructive" : t ? "outline" : "default",
				disabled: n,
				title: i.title,
				onClick: () => r(e),
				children: [/* @__PURE__ */ (0, M.jsx)(w, {
					icon: n ? ht : $t,
					className: n ? "mr-1 animate-spin" : "mr-1"
				}), i.label(t)]
			}, e.transport);
		})
	});
}
function Eo({ node: e, stacks: t, onCollectStack: n, collecting: r, collectError: i }) {
	let [a, o] = (0, j.useState)(null), [s, c] = (0, j.useState)(null), l = (0, j.useMemo)(() => Gi(t, e.key), [t, e.key]), u = e.stackTransports ?? [], d = !!n && u.length > 0;
	if (l.length === 0 && !d) return null;
	let f = l.find((e) => e.checkpointId === s) ?? l[l.length - 1], p = d ? /* @__PURE__ */ (0, M.jsx)(To, {
		transports: u,
		hasStack: !!f,
		collecting: !!r,
		onCollect: (t) => t.destructive ? o(t) : void n?.(e.key, t.transport)
	}) : null, m = a ? bo(a.transport) : null;
	return /* @__PURE__ */ (0, M.jsxs)(So, {
		title: "Stack",
		children: [
			i && /* @__PURE__ */ (0, M.jsx)("div", {
				className: "mb-2 whitespace-pre-wrap text-xs text-red-700 dark:text-red-400",
				role: "alert",
				children: i
			}),
			f ? /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "space-y-density-1",
				children: [l.length > 1 && /* @__PURE__ */ (0, M.jsx)("div", {
					className: "max-w-md",
					children: /* @__PURE__ */ (0, M.jsx)(Ke, {
						"aria-label": "Stack checkpoint",
						value: f.checkpointId,
						onChange: (e) => c(e.target.value),
						options: [...l].reverse().map((e) => ({
							value: e.checkpointId,
							label: [
								ki(e.takenAt),
								e.reason,
								e.transport,
								e.failed ? "failed" : ""
							].filter(Boolean).join(" · ")
						}))
					})
				}), /* @__PURE__ */ (0, M.jsx)(v, {
					records: f.records,
					...p ? { headerActions: p } : {}
				}, `${e.key}:${f.checkpointId}`)]
			}) : /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex items-center justify-between gap-3",
				children: [/* @__PURE__ */ (0, M.jsx)("div", {
					className: "text-xs text-muted-foreground",
					children: r ? "Collecting stack trace..." : "No stack trace collected yet."
				}), p]
			}),
			/* @__PURE__ */ (0, M.jsx)(Ye, {
				open: a !== null,
				title: `${m?.label(!1) ?? ""}?`,
				warning: m?.warning?.(e.pid) ?? `Collecting this stack over ${a?.transport ?? ""} terminates pid ${e.pid}.`,
				message: "The stack is shown here once the dump arrives. Use a non-destructive transport to look without killing.",
				confirmLabel: m?.label(!1) ?? "Collect",
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
function Do({ node: e, samples: t, links: n, ...r }) {
	let i = Ai(e), a = Fi(t, e.key), o = po(t, e.key), s = {
		read: o.read[o.read.length - 1],
		write: o.write[o.write.length - 1]
	};
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "space-y-density-3",
		children: [
			n,
			/* @__PURE__ */ (0, M.jsx)(So, {
				title: "Command",
				children: /* @__PURE__ */ (0, M.jsx)("pre", {
					className: "whitespace-pre-wrap break-all rounded bg-blue-50 p-2 font-mono text-xs text-foreground dark:bg-blue-500/10",
					children: e.command
				})
			}),
			/* @__PURE__ */ (0, M.jsx)(So, {
				title: "Process",
				children: /* @__PURE__ */ (0, M.jsxs)("div", {
					className: "grid grid-cols-2 gap-density-2 text-sm",
					children: [
						/* @__PURE__ */ (0, M.jsx)(Co, {
							label: "Cwd",
							value: e.cwd || "Unavailable",
							wide: !0
						}),
						/* @__PURE__ */ (0, M.jsx)(Co, {
							label: "Started",
							value: ki(e.startedAt)
						}),
						/* @__PURE__ */ (0, M.jsx)(Co, {
							label: "Exited",
							value: i ? ki(e.exitedAt ?? e.lastSeenAt) : "running"
						}),
						/* @__PURE__ */ (0, M.jsx)(Co, {
							label: "Lifetime",
							value: `${Ge(Oi(e))}${i ? "" : " so far"}`
						})
					]
				})
			}),
			/* @__PURE__ */ (0, M.jsxs)(So, {
				title: "Resources",
				children: [/* @__PURE__ */ (0, M.jsxs)("div", {
					className: "grid grid-cols-2 gap-density-1 sm:grid-cols-4",
					children: [
						/* @__PURE__ */ (0, M.jsx)(x, {
							label: "CPU",
							value: Mi(e.cpuPercent),
							barPct: Math.min(100, e.cpuPercent ?? 0),
							tone: _(e.cpuPercent ?? 0)
						}),
						/* @__PURE__ */ (0, M.jsx)(x, {
							label: "RSS",
							value: Xe(e.rssBytes)
						}),
						/* @__PURE__ */ (0, M.jsx)(x, {
							label: "Peak CPU",
							value: Mi(e.peakCpuPercent),
							tone: _(e.peakCpuPercent ?? 0)
						}),
						/* @__PURE__ */ (0, M.jsx)(x, {
							label: "Peak RSS",
							value: Xe(e.peakRssBytes)
						}),
						s.read !== void 0 && /* @__PURE__ */ (0, M.jsx)(x, {
							label: "Disk read",
							value: Xe(s.read)
						}),
						s.write !== void 0 && /* @__PURE__ */ (0, M.jsx)(x, {
							label: "Disk write",
							value: Xe(s.write)
						})
					]
				}), /* @__PURE__ */ (0, M.jsxs)("div", {
					className: "mt-2 grid grid-cols-2 gap-density-2",
					children: [
						/* @__PURE__ */ (0, M.jsx)(Zi, {
							label: "CPU %",
							values: a.cpu,
							format: (e) => `${e.toFixed(1)}%`
						}),
						/* @__PURE__ */ (0, M.jsx)(Zi, {
							label: "RSS",
							values: a.rss,
							format: Xe
						}),
						o.read.length > 0 && /* @__PURE__ */ (0, M.jsx)(Zi, {
							label: "Disk read",
							values: o.read,
							format: Xe
						}),
						o.write.length > 0 && /* @__PURE__ */ (0, M.jsx)(Zi, {
							label: "Disk write",
							values: o.write,
							format: Xe
						})
					]
				})]
			}),
			/* @__PURE__ */ (0, M.jsx)(Eo, {
				node: e,
				...r
			}, e.key)
		]
	});
}
function Oo({ node: e, samples: t, facet: n }) {
	let { label: r } = oo[n];
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "space-y-density-3",
		children: [/* @__PURE__ */ (0, M.jsx)(Zi, {
			label: r,
			values: uo(t, e.key, n),
			format: String
		}), /* @__PURE__ */ (0, M.jsx)(vo, {
			facet: n,
			reading: fo(t, e.key, n)
		})]
	});
}
function ko({ node: e, samples: t, stacks: n = xo, links: r, onCollectStack: i, collecting: a, collectError: o, className: s }) {
	let [c, l] = (0, j.useState)("overview");
	if (!e) return /* @__PURE__ */ (0, M.jsx)("div", {
		className: "flex h-full items-center justify-center text-sm text-muted-foreground",
		children: /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "text-center",
			children: [/* @__PURE__ */ (0, M.jsx)(w, {
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
		...lo(t).map((e) => ({
			id: e,
			label: oo[e].label
		}))
	], d = u.some((e) => e.id === c) ? c : "overview";
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: C("h-full space-y-density-3 overflow-y-auto p-density-4", s),
		"data-process-detail": e.key,
		children: [
			/* @__PURE__ */ (0, M.jsx)(wo, { node: e }),
			u.length > 1 && /* @__PURE__ */ (0, M.jsx)(Kt, {
				tabs: u,
				value: d,
				onChange: (e) => l(e)
			}),
			d === "overview" ? /* @__PURE__ */ (0, M.jsx)(Do, {
				node: e,
				samples: t,
				stacks: n,
				links: r,
				...i ? { onCollectStack: i } : {},
				...a === void 0 ? {} : { collecting: a },
				...o ? { collectError: o } : {}
			}) : d === "env" ? /* @__PURE__ */ (0, M.jsx)(vo, {
				facet: "env",
				...e.environment ? { environment: e.environment } : {},
				...e.environmentError ? { error: e.environmentError } : {}
			}) : /* @__PURE__ */ (0, M.jsx)(Oo, {
				node: e,
				samples: t,
				facet: d
			}, `${e.key}:${d}`)
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/traces/TraceWindowList.tsx
var Ao = "This run's trace store is not served, so its windows cannot be opened.", jo = 6, Mo = (e) => `${e.run_id}:${e.kind}:${e.stream}:${e.afterSeq ?? 0}`, No = (e) => typeof e == "string" || typeof e == "number" || typeof e == "boolean";
function Po({ summary: e }) {
	let t = Object.entries(e ?? {}).filter(([e, t]) => e !== "count" && No(t)).slice(0, jo);
	return /* @__PURE__ */ (0, M.jsx)(M.Fragment, { children: t.map(([e, t]) => /* @__PURE__ */ (0, M.jsxs)("span", {
		className: "rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground",
		"data-summary-chip": e,
		children: [
			e,
			": ",
			/* @__PURE__ */ (0, M.jsx)("span", {
				className: "font-mono",
				children: String(t)
			})
		]
	}, e)) });
}
var Fo = class extends j.Component {
	state = {};
	static getDerivedStateFromError(e) {
		return { error: e instanceof Error ? e.message : String(e) };
	}
	render() {
		return this.state.error ? /* @__PURE__ */ (0, M.jsx)("div", {
			className: "px-2 py-1.5 text-xs text-red-700 dark:text-red-400",
			role: "alert",
			children: this.state.error
		}) : this.props.children;
	}
};
function Io({ window: e, fetchRecords: t }) {
	let [n, r] = (0, j.useState)(null), [i, a] = (0, j.useState)(void 0);
	return (0, j.useEffect)(() => {
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
	]), i ? /* @__PURE__ */ (0, M.jsx)("div", {
		className: "whitespace-pre-wrap border-t border-border px-2 py-1.5 text-xs text-red-700 dark:text-red-400",
		role: "alert",
		children: i
	}) : n ? /* @__PURE__ */ (0, M.jsx)("div", {
		className: "border-t border-border",
		style: { height: 420 },
		children: /* @__PURE__ */ (0, M.jsx)(Fo, { children: /* @__PURE__ */ (0, M.jsx)(ao, {
			kind: e.kind,
			rows: n
		}) })
	}) : /* @__PURE__ */ (0, M.jsx)("div", {
		className: "border-t border-border px-2 py-1.5 text-xs text-muted-foreground",
		children: "Loading records..."
	});
}
function Lo({ window: e, mount: t, formats: n, fetchRecords: r, exportUrl: i, openUrl: a }) {
	let [o, s] = (0, j.useState)(!1), c = "text-xs text-primary hover:underline";
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "rounded border border-border",
		"data-trace-kind": e.kind,
		children: [
			/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex flex-wrap items-center gap-2 px-2 py-1.5 text-sm",
				children: [
					/* @__PURE__ */ (0, M.jsx)("span", {
						className: "font-mono font-medium text-foreground",
						children: e.kind
					}),
					/* @__PURE__ */ (0, M.jsxs)("span", {
						className: "text-xs text-muted-foreground",
						children: [
							e.count,
							" ",
							e.count === 1 ? "record" : "records"
						]
					}),
					/* @__PURE__ */ (0, M.jsx)(Po, { summary: e.summary }),
					/* @__PURE__ */ (0, M.jsx)("span", { className: "flex-1" }),
					t && /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
						/* @__PURE__ */ (0, M.jsx)("a", {
							className: c,
							href: a(e),
							target: "_blank",
							rel: "noreferrer",
							children: "Open in query"
						}),
						n.map((t) => /* @__PURE__ */ (0, M.jsx)("a", {
							className: c,
							href: i(e, t.name),
							download: !0,
							children: t.name.toUpperCase()
						}, t.name)),
						/* @__PURE__ */ (0, M.jsx)("button", {
							type: "button",
							className: c,
							onClick: () => s(!o),
							children: o ? "Hide" : "View"
						})
					] })
				]
			}),
			e.error && /* @__PURE__ */ (0, M.jsx)("div", {
				className: "whitespace-pre-wrap px-2 pb-1.5 text-xs text-red-700 dark:text-red-400",
				children: e.error
			}),
			o && t && /* @__PURE__ */ (0, M.jsx)(Io, {
				window: e,
				fetchRecords: r
			})
		]
	});
}
function Ro({ windows: e, resolve: t, formats: n, fetchRecords: r, exportUrl: i, openUrl: a, error: o }) {
	let s = e.map((e) => t(e)), c = [...new Set(s.filter((e) => !e.mount).map((e) => e.unavailable || Ao))];
	return /* @__PURE__ */ (0, M.jsxs)("div", { children: [
		/* @__PURE__ */ (0, M.jsx)("h3", {
			className: "mb-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground",
			children: "Traces"
		}),
		c.map((e) => /* @__PURE__ */ (0, M.jsx)("div", {
			className: "mb-2 rounded border border-amber-300 bg-amber-50 px-2 py-1 text-xs text-amber-800 dark:border-amber-500/40 dark:bg-amber-500/10 dark:text-amber-300",
			role: "status",
			children: e
		}, e)),
		/* @__PURE__ */ (0, M.jsx)("div", {
			className: "space-y-2",
			children: e.map((e, t) => /* @__PURE__ */ (0, M.jsx)(Lo, {
				window: e,
				mount: s[t].mount,
				formats: n(e.kind),
				fetchRecords: r,
				exportUrl: i,
				openUrl: a
			}, Mo(e)))
		}),
		o && /* @__PURE__ */ (0, M.jsx)("div", {
			className: "mt-1 text-xs text-red-700 dark:text-red-400",
			children: o
		})
	] });
}
// src/config.ts
var zo = typeof window < "u" && window.__gavelBasePath || "";
function Bo({ baseUrl: e = zo, runId: t } = {}) {
	return t ? `${e}/${encodeURIComponent(t)}` : e;
}
function Vo(e, t) {
	return Bo(t) + e;
}
// src/adapters/FixtureFileSource.tsx
var Ho = {
	yaml: {
		language: "yaml",
		label: "YAML"
	},
	yml: {
		language: "yaml",
		label: "YAML"
	},
	json: {
		language: "json",
		label: "JSON"
	},
	md: {
		language: "markdown",
		label: "Markdown"
	},
	mdx: {
		language: "markdown",
		label: "Markdown"
	}
};
function Uo(e) {
	if (e.work_dir) return e.work_dir;
	for (let t of e.children ?? []) {
		let e = Uo(t);
		if (e) return e;
	}
}
function Wo(e) {
	let t = e.detail;
	if (e.framework !== "fixture" || t?.kind !== "file" || !e.file) return;
	let n = Ho[e.file.split(".").pop()?.toLowerCase() ?? ""], r = Uo(e);
	if (n && r) return {
		file: e.file,
		workDir: r,
		...n
	};
}
function Go({ source: e, scope: t = {} }) {
	let [n, r] = (0, j.useState)({});
	return (0, j.useEffect)(() => {
		let n = !0, i = new URLSearchParams({
			work_dir: e.workDir,
			path: e.file
		});
		return fetch(Vo(`/api/file?${i}`, t)).then(async (e) => {
			let t = await e.text();
			if (!e.ok) throw Error(`${e.status} ${t.trim() || e.statusText}`);
			return t;
		}).then((e) => n && r({ text: e }), (e) => n && r({ error: e instanceof Error ? e.message : String(e) })), () => {
			n = !1;
		};
	}, [
		e.file,
		e.workDir,
		t.runId
	]), n.error ? /* @__PURE__ */ (0, M.jsxs)("div", {
		role: "alert",
		className: "p-density-4 text-xs text-red-600",
		children: [
			e.file,
			": ",
			n.error
		]
	}) : n.text === void 0 ? /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "p-density-4 text-xs text-muted-foreground",
		children: [
			"Loading ",
			e.file,
			"…"
		]
	}) : /* @__PURE__ */ (0, M.jsx)("div", {
		className: "min-w-0 p-density-4",
		"data-fixture-file-source": "",
		children: /* @__PURE__ */ (0, M.jsx)(Be, {
			language: e.language,
			source: n.text,
			copyable: !0
		})
	});
}
// src/components/Field.tsx
function F({ title: e, children: t }) {
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "space-y-1",
		children: [/* @__PURE__ */ (0, M.jsx)("div", {
			className: "text-[10px] uppercase tracking-wide text-muted-foreground",
			children: e
		}), t]
	});
}
// src/adapters/GavelDetailBody.tsx
function Ko(e) {
	return typeof e == "object" && !!e && "version" in e && "node" in e;
}
function qo(e) {
	if (e.framework !== "task" || !e.context || typeof e.context != "object") return null;
	let t = e.context;
	return {
		status: typeof t.status == "string" ? t.status : void 0,
		type: typeof t.type == "string" ? t.type : void 0
	};
}
function I({ ctx: e }) {
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [e.parent_test && /* @__PURE__ */ (0, M.jsx)(F, {
		title: "Parent Test",
		children: /* @__PURE__ */ (0, M.jsx)("span", {
			className: "font-mono text-sm",
			children: e.parent_test
		})
	}), e.import_path && /* @__PURE__ */ (0, M.jsx)(F, {
		title: "Import Path",
		children: /* @__PURE__ */ (0, M.jsx)("span", {
			className: "font-mono text-sm",
			children: e.import_path
		})
	})] });
}
function L({ ctx: e, failureLocation: t, suiteAbove: n }) {
	let r = e.failure_location && e.failure_location !== t, i = e.suite_description && e.suite_description !== n;
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
		(i || e.suite_path) && /* @__PURE__ */ (0, M.jsxs)(F, {
			title: "Ginkgo suite",
			children: [i && /* @__PURE__ */ (0, M.jsx)("div", {
				className: "text-sm font-medium",
				children: e.suite_description
			}), e.suite_path && /* @__PURE__ */ (0, M.jsx)("div", {
				className: "font-mono text-xs text-muted-foreground",
				children: e.suite_path
			})]
		}),
		r && /* @__PURE__ */ (0, M.jsx)(F, {
			title: "Failure Location",
			children: /* @__PURE__ */ (0, M.jsx)("span", {
				className: "font-mono text-sm text-red-600 dark:text-red-400",
				children: e.failure_location
			})
		}),
		!!e.labels?.length && /* @__PURE__ */ (0, M.jsx)(F, {
			title: "Labels",
			children: /* @__PURE__ */ (0, M.jsx)("div", {
				className: "flex flex-wrap gap-1",
				children: e.labels.map((e) => /* @__PURE__ */ (0, M.jsx)("span", {
					className: "rounded bg-muted px-2 py-0.5 text-xs",
					children: e
				}, e))
			})
		}),
		!!e.num_attempts && e.num_attempts > 1 && /* @__PURE__ */ (0, M.jsx)(F, {
			title: "Ginkgo attempts",
			children: /* @__PURE__ */ (0, M.jsx)("span", {
				className: "text-sm",
				children: e.num_attempts
			})
		}),
		!!e.report_entries?.length && /* @__PURE__ */ (0, M.jsx)(F, {
			title: "Report entries",
			children: /* @__PURE__ */ (0, M.jsx)("div", {
				className: "space-y-2",
				children: e.report_entries.map((e, t) => /* @__PURE__ */ (0, M.jsxs)("div", {
					className: "rounded border border-border bg-muted/50 p-2",
					children: [
						/* @__PURE__ */ (0, M.jsx)("div", {
							className: "text-sm font-medium",
							children: e.name
						}),
						e.file && /* @__PURE__ */ (0, M.jsxs)("div", {
							className: "font-mono text-xs text-muted-foreground",
							children: [e.file, e.line ? `:${e.line}` : ""]
						}),
						/* @__PURE__ */ (0, M.jsx)("div", {
							className: "mt-1 break-words",
							children: /* @__PURE__ */ (0, M.jsx)(Je, { data: e.value ?? e.representation })
						})
					]
				}, `${e.name}-${t}`))
			})
		})
	] });
}
function R({ node: e }) {
	let t = e, n = ie(), r = t.attempts?.at(-1), i = qo(t), a = (t.children?.length ?? 0) > 0 ? $e(e) : null, o = t.suite?.join(" > "), s = t.failure_detail && t.failure_detail.kind && t.failure_detail.kind !== "raw" ? t.failure_detail : void 0, c = !!t.message && !!s && t.message !== s.summary && t.message.includes("\n");
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "space-y-4 p-density-4",
		children: [
			t.cached && /* @__PURE__ */ (0, M.jsx)("span", {
				className: "inline-block rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground",
				children: "cached"
			}),
			o && /* @__PURE__ */ (0, M.jsx)(F, {
				title: "Suite",
				children: /* @__PURE__ */ (0, M.jsx)("span", {
					className: "text-sm",
					children: o
				})
			}),
			t.package_path && /* @__PURE__ */ (0, M.jsx)(F, {
				title: "Package",
				children: /* @__PURE__ */ (0, M.jsx)("span", {
					className: "font-mono text-sm",
					children: t.package_path
				})
			}),
			s ? /* @__PURE__ */ (0, M.jsxs)(F, {
				title: "Error",
				children: [/* @__PURE__ */ (0, M.jsx)(ne, { failure: s }), c && /* @__PURE__ */ (0, M.jsxs)("details", {
					className: "rounded bg-muted/50 p-2",
					children: [/* @__PURE__ */ (0, M.jsx)("summary", {
						className: "cursor-pointer text-xs font-semibold text-muted-foreground",
						children: "Full message"
					}), /* @__PURE__ */ (0, M.jsx)(l, {
						text: t.message,
						className: "mt-2 block max-h-64 overflow-y-auto whitespace-pre-wrap font-mono text-xs"
					})]
				})]
			}) : t.message ? /* @__PURE__ */ (0, M.jsx)(F, {
				title: "Error",
				children: /* @__PURE__ */ (0, M.jsx)(l, {
					text: t.message,
					className: "block max-h-64 overflow-y-auto whitespace-pre-wrap rounded bg-red-50 p-3 font-mono text-sm text-red-700 dark:bg-red-500/10 dark:text-red-300"
				})
			}) : null,
			t.command && /* @__PURE__ */ (0, M.jsx)(F, {
				title: "Command",
				children: /* @__PURE__ */ (0, M.jsxs)("pre", {
					className: "whitespace-pre-wrap rounded bg-muted p-2 font-mono text-sm",
					children: [/* @__PURE__ */ (0, M.jsx)("span", {
						className: "text-muted-foreground",
						children: "$ "
					}), t.command]
				})
			}),
			i && (i.status || i.type) && /* @__PURE__ */ (0, M.jsx)(F, {
				title: "Task",
				children: /* @__PURE__ */ (0, M.jsxs)("div", {
					className: "flex gap-4 text-sm",
					children: [i.type && /* @__PURE__ */ (0, M.jsxs)("span", { children: ["type: ", /* @__PURE__ */ (0, M.jsx)("span", {
						className: "font-mono",
						children: i.type
					})] }), i.status && /* @__PURE__ */ (0, M.jsxs)("span", { children: ["status: ", /* @__PURE__ */ (0, M.jsx)("span", {
						className: "font-mono",
						children: i.status
					})] })]
				})
			}),
			t.framework === "go test" && t.context && /* @__PURE__ */ (0, M.jsx)(I, { ctx: t.context }),
			t.framework === "ginkgo" && t.context && /* @__PURE__ */ (0, M.jsx)(L, {
				ctx: t.context,
				failureLocation: t.failure_detail?.location,
				suiteAbove: o
			}),
			r && /* @__PURE__ */ (0, M.jsx)(F, {
				title: "Attempt",
				children: /* @__PURE__ */ (0, M.jsx)(S, {
					attempt: r,
					...n.fetchProcessMetrics ? { fetchProcessMetrics: n.fetchProcessMetrics } : {}
				}, `${r.run_kind ?? ""}:${r.sequence}`)
			}),
			/* @__PURE__ */ (0, M.jsx)(ce, { node: e }),
			a && a.total > 0 && /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex flex-wrap gap-3 text-xs text-muted-foreground",
				children: [
					/* @__PURE__ */ (0, M.jsxs)("span", { children: [a.total, " total"] }),
					a.passed > 0 && /* @__PURE__ */ (0, M.jsxs)("span", { children: [a.passed, " passed"] }),
					a.failed > 0 && /* @__PURE__ */ (0, M.jsxs)("span", { children: [a.failed, " failed"] }),
					a.warned > 0 && /* @__PURE__ */ (0, M.jsxs)("span", { children: [a.warned, " warned"] }),
					a.aborted > 0 && /* @__PURE__ */ (0, M.jsxs)("span", { children: [a.aborted, " aborted"] }),
					a.skipped > 0 && /* @__PURE__ */ (0, M.jsxs)("span", { children: [a.skipped, " skipped"] }),
					a.pending > 0 && /* @__PURE__ */ (0, M.jsxs)("span", { children: [a.pending, " pending"] })
				]
			}),
			t.detail != null && /* @__PURE__ */ (0, M.jsx)(F, {
				title: "Detail",
				children: Ko(t.detail) ? /* @__PURE__ */ (0, M.jsx)(m, { data: t.detail }) : /* @__PURE__ */ (0, M.jsx)(Je, { data: t.detail })
			})
		]
	});
}
// src/routes.ts
var Jo = [
	"tests",
	"lint",
	"files",
	"bench",
	"processes",
	"history"
];
function Yo(e) {
	return e ? e.split(",").map((e) => e.trim()).filter(Boolean) : [];
}
function Xo(e) {
	return e || "preview";
}
function Zo(e) {
	let t = e.pathname;
	zo && t.startsWith(zo) && (t = t.slice(zo.length));
	let n = t.replace(/^\/+|\/+$/g, ""), r = n ? n.split("/").map(decodeURIComponent) : [], i = "tests", a = "", o = r[0];
	o && Jo.includes(o) && (i = o, a = r.slice(1).join("/"));
	let s = new URLSearchParams(e.search), c = s.get("sample"), l = s.get("line"), u = s.get("dir");
	return {
		tab: i,
		selectedPath: a,
		filters: {
			status: at(Yo(s.get("status"))),
			framework: at(Yo(s.get("framework")))
		},
		lintGrouping: es(s.get("grouping")),
		lintFilters: {
			severity: at(Yo(s.get("severity"))),
			linter: at(Yo(s.get("linter")))
		},
		problemFilters: {
			severity: at(Yo(s.get("problemSeverity"))),
			type: at(Yo(s.get("problemType")))
		},
		sampleIndex: c !== null && /^\d+$/.test(c) ? Number(c) : void 0,
		hideExited: s.get("hideExited") === "1",
		...u ? { workDir: u } : {},
		...i === "files" && s.get("kind") === "folder" ? { folder: !0 } : {},
		view: Xo(s.get("view")),
		line: l !== null && /^\d+$/.test(l) && Number(l) > 0 ? Number(l) : void 0
	};
}
function Qo(e) {
	let t = [e.tab];
	e.selectedPath && t.push(...e.selectedPath.split("/").map(encodeURIComponent));
	let n = new URLSearchParams();
	e.tab === "tests" && (e.filters.status.size > 0 && n.set("status", nt(e.filters.status).join(",")), e.filters.framework.size > 0 && n.set("framework", nt(e.filters.framework).join(","))), e.tab === "lint" && (e.lintGrouping !== "linter-rule-file" && n.set("grouping", e.lintGrouping), e.lintFilters.severity.size > 0 && n.set("severity", nt(e.lintFilters.severity).join(",")), e.lintFilters.linter.size > 0 && n.set("linter", nt(e.lintFilters.linter).join(","))), e.tab === "files" && (e.workDir && n.set("dir", e.workDir), e.folder && n.set("kind", "folder"), e.view && e.view !== "preview" && n.set("view", e.view), e.line !== void 0 && n.set("line", String(e.line)), e.problemFilters.severity.size > 0 && n.set("problemSeverity", nt(e.problemFilters.severity).join(",")), e.problemFilters.type.size > 0 && n.set("problemType", nt(e.problemFilters.type).join(","))), e.tab === "bench" && e.sampleIndex !== void 0 && n.set("sample", String(e.sampleIndex)), e.tab === "processes" && e.hideExited && n.set("hideExited", "1");
	let r = n.toString();
	return `${zo}/${t.join("/")}${r ? `?${r}` : ""}`;
}
function $o(e, t) {
	let [n, r = ""] = Qo(e).split("?", 2);
	return `${n}.${t}${r ? `?${r}` : ""}`;
}
function es(e) {
	return An.find((t) => t === e) ?? "linter-rule-file";
}
// src/export.ts
function z(e, t) {
	return `${(e.selectedPath ? e.selectedPath.replace(/\//g, "-") : e.tab) || e.tab}.${t}`;
}
function ts(e, t) {
	let n = document.createElement("a");
	n.href = $o(e, t), n.download = z(e, t), document.body.appendChild(n), n.click(), n.remove();
}
function ns(e) {
	return e.tab === "lint" ? "Use this gavel lint report as context for your analysis or code changes." : e.tab === "bench" ? "Use this gavel benchmark report as context for your analysis or code changes." : "Use this gavel test report as context for your analysis or code changes.";
}
async function rs(e) {
	let t = $o(e, "md"), n = new URL(t, window.location.origin).toString(), r = await fetch(t, { headers: { Accept: "text/markdown" } });
	if (!r.ok) throw Error(`Markdown export failed (${r.status})`);
	let i = (await r.text()).trimEnd(), a = [
		ns(e),
		"",
		`Source URL: ${n}`,
		"",
		"```md",
		i,
		"```"
	].join("\n");
	await ke(a);
}
// src/hooks/use-copy-feedback.ts
function is({ copiedMs: e = 2e3, errorMs: t = 3e3 } = {}) {
	let [n, r] = (0, j.useState)("idle"), [i, a] = (0, j.useState)(""), o = (0, j.useRef)(null), s = (0, j.useCallback)(() => {
		o.current !== null && (window.clearTimeout(o.current), o.current = null);
	}, []);
	return (0, j.useEffect)(() => s, [s]), {
		copyState: n,
		copyError: i,
		beginCopy: (0, j.useCallback)(() => {
			r("copying"), a(""), s();
		}, [s]),
		resetCopyFeedback: (0, j.useCallback)((n, i = "") => {
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
function as(e) {
	let { copyState: t, copyError: n, beginCopy: r, resetCopyFeedback: i } = is();
	return {
		copyState: t,
		copyError: n,
		run: (0, j.useCallback)(async () => {
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
function os({ routeState: e, title: t }) {
	return /* @__PURE__ */ (0, M.jsx)(ye, {
		label: "Download",
		icon: Ve,
		variant: "outline",
		size: "sm",
		items: [{
			label: "JSON",
			icon: u,
			title: $o(e, "json"),
			onSelect: () => ts(e, "json")
		}, {
			label: "Markdown",
			icon: te,
			title: $o(e, "md"),
			onSelect: () => ts(e, "md")
		}],
		...t ? { title: t } : {}
	});
}
var ss = {
	idle: "Copy AI Prompt",
	copying: "Copying...",
	copied: "Copied",
	error: "Copy failed"
}, cs = {
	idle: me,
	copying: ht,
	copied: ve,
	error: Re
};
function ls({ routeState: e, disabled: t, disabledTitle: n, title: r }) {
	let { copyState: i, copyError: a, run: o } = as(() => rs(e)), s = cs[i];
	return /* @__PURE__ */ (0, M.jsxs)(fe, {
		type: "button",
		variant: "outline",
		size: "sm",
		disabled: t || i === "copying",
		onClick: () => {
			o();
		},
		title: t ? n : a || r || $o(e, "md"),
		children: [/* @__PURE__ */ (0, M.jsx)(s, { className: i === "copying" ? "mr-1 animate-spin" : "mr-1" }), ss[i]]
	});
}
// src/adapters/NodeActions.tsx
var B = /* @__PURE__ */ new Set([
	"go test",
	"ginkgo",
	"vitest"
]);
function us(e) {
	if (e.framework !== "fixture") return;
	let t = e.detail?.step?.id;
	return typeof t == "string" && t !== "" ? t : void 0;
}
function ds(e, t) {
	return t && e.framework !== "task" && B.has(e.framework ?? "") && !!e.file;
}
function fs(e) {
	if (e.failed || e.timed_out) return !0;
	if ((e.children?.length ?? 0) === 0) return !1;
	let t = $e(e);
	return t.failed > 0 || t.timedout > 0;
}
function ps({ node: e, busy: t, onEdit: n }) {
	let [r, i] = (0, j.useState)(null), a = r?.scope === "file" ? e.file || "file" : e.name || "test", o = r?.action === "skip" ? "Skip" : "Delete", s = r?.scope === "file" ? "file" : "test", c = (e, t) => () => i({
		action: e,
		scope: t
	});
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, M.jsx)(ye, {
		label: "Source",
		title: "Edit the test source",
		items: [
			{
				label: "Skip Test",
				icon: pt,
				disabled: t,
				onSelect: c("skip", "test")
			},
			{
				label: "Delete Test",
				icon: E,
				disabled: t,
				onSelect: c("delete", "test")
			},
			{
				label: "Skip File",
				icon: T,
				disabled: t,
				onSelect: c("skip", "file")
			},
			{
				label: "Delete File",
				icon: E,
				disabled: t,
				onSelect: c("delete", "file")
			}
		]
	}), /* @__PURE__ */ (0, M.jsx)(Ye, {
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
function ms({ node: e, config: t }) {
	let n = e.framework === "task" ? void 0 : t.nodeRouteState(e), r = t.failingOnlyRouteState(e), i = fs(e), a = us(e);
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
		a && /* @__PURE__ */ (0, M.jsx)(Se, {
			label: "Retry from here",
			icon: Oe,
			variant: "outline",
			size: "sm",
			primaryDisabled: t.rerunBusy,
			onClick: () => t.onRetryFromStep(e, a),
			title: "More ways to rerun this fixture from this step",
			items: [{
				label: "Copy rerun command",
				icon: me,
				disabled: t.runInProgress,
				title: t.runInProgress ? "Available once the run finishes" : "Copy the command that reruns this fixture from this step, reusing the earlier steps' results",
				onSelect: () => t.onCopyRerunCommand(e, a)
			}]
		}),
		n && /* @__PURE__ */ (0, M.jsx)(os, {
			routeState: n,
			title: "Download this node as JSON or Markdown"
		}),
		n && r && /* @__PURE__ */ (0, M.jsx)(ls, {
			routeState: r,
			disabled: !i,
			disabledTitle: "No failures to copy",
			title: "Copy failing output and repro steps as AI prompt"
		}),
		ds(e, t.testEditSupported) && /* @__PURE__ */ (0, M.jsx)(ps, {
			node: e,
			busy: t.testEditBusy,
			onEdit: t.onTestEdit
		})
	] });
}
// src/components/ResourceParts.tsx
function hs(e) {
	return e >= 1e3 ? `${(e / 1e3).toFixed(2)} s` : `${e.toFixed(1)} ms`;
}
function gs(e) {
	return e >= 1048576 ? `${(e / 1048576).toFixed(1)} MiB` : e >= 1024 ? `${(e / 1024).toFixed(1)} KiB` : `${e} B`;
}
function _s(e) {
	return `${e >= 0 ? "+" : "-"}${gs(Math.abs(e))}`;
}
function vs({ profile: e }) {
	let t = e.statements;
	return t?.length ? /* @__PURE__ */ (0, M.jsx)(Ie, {
		title: `${t.length} ${t.length === 1 ? "query" : "queries"}`,
		defaultOpen: !1,
		children: /* @__PURE__ */ (0, M.jsx)("div", {
			className: "max-h-80 space-y-2 overflow-auto",
			children: t.map((e, t) => /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "rounded border border-border bg-muted/50 p-2 text-xs",
				children: [
					/* @__PURE__ */ (0, M.jsxs)("div", {
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
					/* @__PURE__ */ (0, M.jsx)("pre", {
						className: "whitespace-pre-wrap break-all font-mono",
						children: e.sql
					}),
					e.params.length > 0 && /* @__PURE__ */ (0, M.jsxs)("div", {
						className: "mt-1",
						children: [/* @__PURE__ */ (0, M.jsx)("span", {
							className: "text-muted-foreground",
							children: "Params"
						}), /* @__PURE__ */ (0, M.jsx)("ol", {
							className: "list-decimal pl-5 font-mono",
							children: e.params.map((e, t) => /* @__PURE__ */ (0, M.jsx)("li", {
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
function ys({ profile: e }) {
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "space-y-2",
		children: [
			e.sample_count > 0 ? /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "grid grid-cols-3 gap-density-1",
				children: [
					/* @__PURE__ */ (0, M.jsx)(x, {
						label: "Peak CPU",
						value: `${e.peak_cpu_percent.toFixed(1)}%`
					}),
					/* @__PURE__ */ (0, M.jsx)(x, {
						label: "Peak RSS",
						value: Xe(e.peak_rss_bytes)
					}),
					/* @__PURE__ */ (0, M.jsx)(x, {
						label: "Peak memory",
						value: `${e.peak_memory_percent.toFixed(1)}%`
					})
				]
			}) : /* @__PURE__ */ (0, M.jsx)("p", {
				className: "text-sm text-muted-foreground",
				children: "No process samples were captured during this test."
			}),
			e.disk_io && /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "grid grid-cols-2 gap-density-1",
				children: [/* @__PURE__ */ (0, M.jsx)(x, {
					label: "Observed disk read",
					value: Xe(e.disk_io.disk_read_bytes)
				}), /* @__PURE__ */ (0, M.jsx)(x, {
					label: "Observed disk write",
					value: Xe(e.disk_io.disk_write_bytes)
				})]
			}),
			/* @__PURE__ */ (0, M.jsxs)("p", {
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
function bs({ node: e }) {
	let t = e.sql_profile;
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "space-y-4 p-density-4",
		children: [e.fixture_profile && /* @__PURE__ */ (0, M.jsx)(F, {
			title: "Resources",
			children: /* @__PURE__ */ (0, M.jsx)(ys, { profile: e.fixture_profile })
		}), t && /* @__PURE__ */ (0, M.jsxs)(F, {
			title: "SQL profile",
			children: [/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "grid grid-cols-3 gap-density-1",
				children: [
					/* @__PURE__ */ (0, M.jsx)(x, {
						label: "Queries",
						value: String(t.query_count)
					}),
					/* @__PURE__ */ (0, M.jsx)(x, {
						label: "Total SQL",
						value: `${t.total_duration_ms.toFixed(1)} ms`
					}),
					/* @__PURE__ */ (0, M.jsx)(x, {
						label: "Slow queries",
						value: String(t.slow_query_count)
					})
				]
			}), /* @__PURE__ */ (0, M.jsx)(vs, { profile: t })]
		})]
	});
}
// src/prompts.ts
function xs(e = {}) {
	return Vo("/api", e);
}
function Ss(e = {}) {
	return Vo("/api/v1", e);
}
function Cs(e = {}) {
	return async ({ descriptor: t, query: n, rootValue: r }) => {
		if (!t.url) throw Error(`lookup ${t.filter} has no url`);
		let i = new URL(Vo(t.url, e), window.location.origin);
		for (let [e, n] of Object.entries(Pe(t, r))) i.searchParams.set(e, n);
		i.searchParams.set("__lookup_filter", t.filter), i.searchParams.set(t.searchParam ?? "__lookup_q", n);
		let a = await fetch(`${i.pathname}${i.search}`, { headers: { Accept: "application/json" } });
		if (!a.ok) throw Error(`lookup ${t.filter}: ${a.status} ${(await a.text()).trim()}`);
		let o = (await a.json()).filters?.[t.filter];
		if (!o) throw Error(`lookup ${t.filter}: the response has no ${t.filter} filter`);
		return Object.entries(o.options ?? {}).map(([e, t]) => ({
			value: e,
			label: t.plain || t.text || e,
			...qt(e, t)
		}));
	};
}
// src/components/StepPendingForm.tsx
function ws({ node: e, owner: t, scope: n = {} }) {
	let r = (0, j.useMemo)(() => Cs(n), [n.runId]), i = xs(n), { pending: a } = Mr({
		basePath: i,
		owner: t,
		labels: { node: e.name },
		state: "pending",
		enabled: !!t
	});
	return t ? a.length === 0 ? /* @__PURE__ */ (0, M.jsx)("p", {
		className: "text-xs text-muted-foreground",
		children: "No pending form for this step"
	}) : /* @__PURE__ */ (0, M.jsx)("div", {
		className: "space-y-density-3",
		children: a.map((e) => /* @__PURE__ */ (0, M.jsx)(Nr, {
			prompt: e,
			basePath: i,
			lookupFetcher: r,
			connectionApiBase: Ss(n),
			idPrefix: `step-prompt-${e.id}`
		}, e.id))
	}) : /* @__PURE__ */ (0, M.jsx)("p", {
		className: "text-xs text-muted-foreground",
		children: "No pending form for this step: the run publishes no prompts"
	});
}
// src/adapters/StepProcesses.tsx
function Ts({ processes: e, onOpenProcess: t }) {
	return /* @__PURE__ */ (0, M.jsx)("div", {
		className: "flex h-80 min-h-0 flex-col px-3 pt-2",
		"data-step-processes": "",
		children: /* @__PURE__ */ (0, M.jsx)(ra, {
			nodes: e,
			selectedKey: null,
			expandAll: !0,
			onSelect: t,
			emptyLabel: "This step started no processes the trace recorded"
		})
	});
}
// src/traces.ts
var Es = class extends Error {};
function Ds(e, t) {
	return e?.runs[t] || { unavailable: `The trace store of run ${t || "(unnamed)"} is not served by this view.` };
}
function Os(e) {
	if (e?.process_run && e.process) return e.runs[e.process_run]?.mount;
}
function ks(e, t) {
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
function As(e, t) {
	return `${e}${t}`;
}
function js(e, t = {}) {
	let n = new URLSearchParams({ stream: e.stream });
	e.afterSeq && n.set("afterSeq", String(e.afterSeq)), e.toSeq && n.set("toSeq", String(e.toSeq));
	for (let [e, r] of Object.entries(t)) n.set(e, r);
	return n.toString();
}
function Ms(e, t, n) {
	return As(e, `/api/v1/traces/${encodeURIComponent(t.kind)}/export?${js(t, { format: n })}`);
}
function Ns(e, t) {
	return As(e, `/traces/${encodeURIComponent(t.kind)}?${js(t)}`);
}
async function Ps(e, t) {
	let n = (await e.text()).trim();
	return /* @__PURE__ */ Error(`${t} failed (${e.status})${n ? `: ${n}` : ""}`);
}
function Fs(e) {
	return e.split("\n").filter((e) => e.trim() !== "").map((e) => JSON.parse(e));
}
async function Is(e, t, n) {
	let r = await fetch(Ms(e, t, "ndjson"), n ? { signal: n } : void 0);
	if (r.status === 404) throw new Es(`${t.kind} stream ${t.stream} has no records`);
	if (!r.ok) throw await Ps(r, `export ${t.kind} window`);
	return Fs(await r.text());
}
function Ls(e) {
	return `profile-${`traces/${e}`.toLowerCase().replace(/[ _/.-]/g, "-").replace(/[^a-z0-9-]/g, "").replace(/^-+|-+$/g, "")}`;
}
async function Rs(e, t) {
	let n = new URLSearchParams({
		stream: t.stream,
		sort: "seq",
		order: "desc",
		limit: "1"
	}), r = await fetch(As(e, `/api/v1/profile/${Ls(t.kind)}?${n}`), { headers: { Accept: "application/json" } });
	if (!r.ok) throw await Ps(r, `read ${t.kind} stream`);
	return (await r.json())[0]?.seq ?? 0;
}
async function zs(e) {
	let t = await fetch(As(e, "/api/v1/traces/kinds"), { headers: { Accept: "application/json" } });
	if (!t.ok) throw await Ps(t, "list trace kinds");
	return t.json();
}
async function Bs({ mount: e, session: t, params: n }) {
	let r = await fetch(As(e, `/api/v1/sessions/${encodeURIComponent(t)}/checkpoint`), {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(n)
	});
	if (!r.ok) throw await Ps(r, "checkpoint");
}
// src/adapters/TracesPanel.tsx
function Vs(e) {
	let [t, n] = (0, j.useState)({}), [r, i] = (0, j.useState)(void 0);
	return (0, j.useEffect)(() => {
		e && zs(e).then((e) => n(Object.fromEntries(e.map((e) => [e.name, e.exportFormats])))).catch((e) => i(e instanceof Error ? e.message : String(e)));
	}, [e]), {
		byKind: t,
		error: r
	};
}
function Hs({ windows: e, traces: t }) {
	let n = Vs(e.map((e) => Ds(t, e.run_id).mount).find(Boolean)), r = (0, j.useCallback)((e) => Ds(t, e.run_id), [t]), i = (0, j.useCallback)((e) => {
		let n = Ds(t, e.run_id).mount;
		if (!n) throw Error(`trace store of run ${e.run_id} is not served`);
		return n;
	}, [t]), a = (0, j.useMemo)(() => ({
		formats: (e) => n.byKind[e] ?? [],
		fetchRecords: (e, t) => Is(i(e), e, t),
		exportUrl: (e, t) => Ms(i(e), e, t),
		openUrl: (e) => Ns(i(e), e)
	}), [n.byKind, i]);
	return /* @__PURE__ */ (0, M.jsx)("div", {
		className: "space-y-density-3 p-density-4",
		children: /* @__PURE__ */ (0, M.jsx)(Ro, {
			windows: e,
			resolve: r,
			...a,
			...n.error ? { error: n.error } : {}
		})
	});
}
// src/adapters/index.tsx
var Us = (e) => e;
function Ws(e) {
	let t = Us(e);
	return t.framework === "fixture" ? t.context?.traces ?? [] : [];
}
function Gs(e, t) {
	let n = new Set(Us(e).fixture?.pids ?? []);
	return n.size === 0 ? [] : t.filter((e) => n.has(e.pid));
}
function Ks() {
	return {
		id: "gavel-detail",
		match: (e) => Us(e).framework !== "fixture",
		renderDetail: ({ node: e }) => /* @__PURE__ */ (0, M.jsx)(R, { node: e })
	};
}
function qs(e) {
	return {
		id: "gavel-traces",
		match: (e) => Ws(e).length > 0,
		detailTabs: ({ node: t }) => [{
			id: "traces",
			label: "Traces",
			render: () => /* @__PURE__ */ (0, M.jsx)(Hs, {
				windows: Ws(t),
				traces: e
			})
		}]
	};
}
function Js({ processes: e, onOpenProcess: t }) {
	return {
		id: "gavel-processes",
		match: (t) => Gs(t, e).length > 0,
		detailTabs: ({ node: n }) => [{
			id: "processes",
			label: "Processes",
			render: () => /* @__PURE__ */ (0, M.jsx)(Ts, {
				processes: Gs(n, e),
				onOpenProcess: t
			})
		}]
	};
}
function Ys() {
	return {
		id: "gavel-resources",
		match: (e) => {
			let t = Us(e);
			return !!t.fixture_profile || !!t.sql_profile;
		},
		detailTabs: ({ node: e }) => [{
			id: "resources",
			label: "Resources",
			render: () => /* @__PURE__ */ (0, M.jsx)(bs, { node: Us(e) })
		}]
	};
}
function Xs(e = {}) {
	return {
		id: "gavel-fixture-file",
		match: (e) => Wo(Us(e)) !== void 0,
		detailTabs: ({ node: t }) => {
			let n = Wo(Us(t));
			return n ? [{
				id: "fixture-file",
				label: n.label,
				render: () => /* @__PURE__ */ (0, M.jsx)(Go, {
					source: n,
					scope: e
				})
			}] : [];
		}
	};
}
function Zs(e) {
	return {
		id: "gavel-actions",
		match: (e) => Us(e).framework !== "task",
		nodeActions: ({ node: t }) => /* @__PURE__ */ (0, M.jsx)(ms, {
			node: Us(t),
			config: e
		})
	};
}
function Qs({ traces: e, processes: t, onOpenProcess: n, actions: r, scope: i }) {
	return [
		Ks(),
		Xs(i),
		Zs(r),
		qs(e),
		Js({
			processes: t,
			onOpenProcess: n
		}),
		Ys()
	];
}
function $s(e) {
	let { promptOwner: t, scope: n } = e, r = se({ renderPendingPrompt: (e) => /* @__PURE__ */ (0, M.jsx)(ws, {
		node: e,
		owner: t,
		...n ? { scope: n } : {}
	}) });
	return re([...Qs(e), ...r.adapters]);
}
// src/components/BenchView.tsx
function ec(e) {
	return e >= 1e9 ? `${(e / 1e9).toFixed(2)}s` : e >= 1e6 ? `${(e / 1e6).toFixed(2)}ms` : e >= 1e3 ? `${(e / 1e3).toFixed(2)}µs` : `${e.toFixed(2)}ns`;
}
function tc(e, t) {
	return e.only_in ? "only" : e.significant && e.delta_pct > t ? "regression" : e.significant && e.delta_pct < -t ? "improvement" : "neutral";
}
function nc(e, t) {
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
		category: tc(e, t),
		onlyIn: e.only_in ?? null,
		threshold: t,
		significant: !!e.significant
	};
}
var rc = {
	regression: "text-red-600 dark:text-red-400",
	improvement: "text-green-600 dark:text-green-400",
	neutral: "text-muted-foreground",
	only: "text-muted-foreground"
}, ic = {
	regression: "bg-red-500",
	improvement: "bg-green-500",
	neutral: "bg-muted-foreground/40",
	only: "bg-muted-foreground/40"
};
function ac({ row: e }) {
	if (e.delta === null) return /* @__PURE__ */ (0, M.jsxs)("span", {
		className: "text-xs italic text-muted-foreground",
		children: ["only in ", e.onlyIn]
	});
	let t = Math.max(-50, Math.min(50, e.delta)), n = Math.abs(t);
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex w-full min-w-48 items-center gap-2",
		children: [/* @__PURE__ */ (0, M.jsxs)("div", {
			className: "relative h-4 min-w-20 flex-1 overflow-hidden rounded bg-muted",
			children: [/* @__PURE__ */ (0, M.jsx)("div", { className: "absolute inset-y-0 left-1/2 w-px bg-muted-foreground/60" }), /* @__PURE__ */ (0, M.jsx)("div", {
				className: C("absolute inset-y-0", ic[e.category]),
				style: e.delta >= 0 ? {
					left: "50%",
					width: `${n}%`
				} : {
					right: "50%",
					width: `${n}%`
				}
			})]
		}), /* @__PURE__ */ (0, M.jsxs)("span", {
			className: C("w-16 text-right text-xs tabular-nums", rc[e.category], e.significant && "font-semibold"),
			children: [
				e.delta >= 0 ? "+" : "",
				e.delta.toFixed(2),
				"%"
			]
		})]
	});
}
var oc = [
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
		render: (e, t) => t.base === null ? "" : /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "text-xs tabular-nums",
			children: [ec(t.base), t.baseSpread && /* @__PURE__ */ (0, M.jsx)("div", {
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
		render: (e, t) => t.head === null ? "" : /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "text-xs tabular-nums",
			children: [ec(t.head), t.headSpread && /* @__PURE__ */ (0, M.jsx)("div", {
				className: "text-[10px] text-muted-foreground",
				children: t.headSpread
			})]
		})
	},
	{
		key: "delta",
		label: "Δ",
		sortable: !0,
		render: (e, t) => /* @__PURE__ */ (0, M.jsx)(ac, { row: t })
	},
	{
		key: "pValue",
		label: "p-value",
		sortable: !0,
		align: "right",
		render: (e, t) => /* @__PURE__ */ (0, M.jsxs)("span", {
			className: "text-xs tabular-nums text-muted-foreground",
			children: [t.pValue === null ? "—" : t.pValue.toPrecision(2), t.samples !== null && /* @__PURE__ */ (0, M.jsxs)("span", {
				className: "ml-1",
				children: ["n=", t.samples]
			})]
		})
	}
];
function sc({ bench: e }) {
	let t = (0, j.useMemo)(() => (e?.deltas ?? []).map((t) => nc(t, e.threshold)), [e]), n = (0, j.useMemo)(() => {
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
	return e ? /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col gap-density-2 p-density-4",
		children: [/* @__PURE__ */ (0, M.jsxs)("div", {
			className: "flex flex-wrap items-center gap-4 text-sm",
			children: [
				/* @__PURE__ */ (0, M.jsxs)("div", { children: [
					/* @__PURE__ */ (0, M.jsx)("span", {
						className: "font-mono",
						children: e.base_label || "base"
					}),
					/* @__PURE__ */ (0, M.jsx)("span", {
						className: "mx-2 text-muted-foreground",
						children: "→"
					}),
					/* @__PURE__ */ (0, M.jsx)("span", {
						className: "font-mono",
						children: e.head_label || "head"
					})
				] }),
				/* @__PURE__ */ (0, M.jsxs)("div", {
					className: "text-xs text-muted-foreground",
					children: [
						"threshold ±",
						e.threshold.toFixed(1),
						"%"
					]
				}),
				n.regression > 0 && /* @__PURE__ */ (0, M.jsxs)(D, {
					tone: "danger",
					variant: "soft",
					size: "xs",
					children: [
						n.regression,
						" regression",
						n.regression > 1 ? "s" : ""
					]
				}),
				n.improvement > 0 && /* @__PURE__ */ (0, M.jsxs)(D, {
					tone: "success",
					variant: "soft",
					size: "xs",
					children: [
						n.improvement,
						" improvement",
						n.improvement > 1 ? "s" : ""
					]
				}),
				n.neutral > 0 && /* @__PURE__ */ (0, M.jsxs)(D, {
					tone: "neutral",
					variant: "soft",
					size: "xs",
					children: [n.neutral, " unchanged"]
				}),
				/* @__PURE__ */ (0, M.jsxs)("div", {
					className: C("ml-auto font-mono text-sm", e.geomean_delta > e.threshold ? "font-bold text-red-600 dark:text-red-400" : e.geomean_delta < -e.threshold ? "font-bold text-green-600 dark:text-green-400" : "text-muted-foreground"),
					children: [
						"geomean: ",
						e.geomean_delta >= 0 ? "+" : "",
						e.geomean_delta.toFixed(2),
						"%"
					]
				})
			]
		}), /* @__PURE__ */ (0, M.jsx)(g, {
			data: t,
			columns: oc,
			defaultSort: {
				key: "delta",
				dir: "desc"
			},
			emptyMessage: "No benchmarks were compared",
			className: "min-h-0 flex-1"
		})]
	}) : /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "p-8 text-center text-sm text-muted-foreground",
		children: [
			"No benchmark comparison loaded. Run ",
			/* @__PURE__ */ (0, M.jsx)("code", {
				className: "rounded bg-muted px-1",
				children: "gavel bench compare --ui"
			}),
			"."
		]
	});
}
// src/components/FixtureBenchView.tsx
var cc = [
	"setup",
	"build",
	"daemon",
	"daemon-stop",
	"cleanup"
];
function lc(e) {
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
			cc.includes(r.kind) ? t.push(r) : n.push(r);
		}
		e.children?.forEach(r);
	};
	return e.forEach(r), {
		phases: t,
		fixtures: n
	};
}
function uc(e, t) {
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
function dc(e) {
	let t = "font-mono tabular-nums", n = [
		{
			key: "name",
			label: "Step",
			sortable: !0,
			grow: !0,
			render: (e, t) => /* @__PURE__ */ (0, M.jsxs)("div", { children: [
				/* @__PURE__ */ (0, M.jsx)("span", {
					className: "font-medium",
					children: t.name
				}),
				/* @__PURE__ */ (0, M.jsx)("span", {
					className: "ml-2 text-xs text-muted-foreground",
					children: t.kind
				}),
				t.entry.violations?.map((e) => /* @__PURE__ */ (0, M.jsx)("div", {
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
			render: (e, t) => hs(t.duration)
		},
		{
			key: "command",
			label: "Command",
			sortable: !0,
			align: "right",
			cellClassName: `${t} text-muted-foreground`,
			render: (e, t) => t.command === null ? "—" : hs(t.command)
		},
		{
			key: "diskRead",
			label: "Disk read",
			sortable: !0,
			align: "right",
			cellClassName: t,
			render: (e, t) => t.diskRead === null ? "—" : gs(t.diskRead)
		},
		{
			key: "diskWrite",
			label: "Disk write",
			sortable: !0,
			align: "right",
			cellClassName: t,
			render: (e, t) => t.diskWrite === null ? "—" : gs(t.diskWrite)
		},
		{
			key: "sql",
			label: "SQL",
			sortable: !0,
			align: "right",
			cellClassName: t,
			render: (e, t) => {
				let n = t.entry.sql_profile;
				return n ? /* @__PURE__ */ (0, M.jsxs)("div", { children: [
					hs(n.total_duration_ms),
					/* @__PURE__ */ (0, M.jsxs)("div", {
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
					/* @__PURE__ */ (0, M.jsxs)("div", {
						className: "text-xs text-muted-foreground",
						children: ["Max ", hs(n.max_query_ms)]
					}),
					/* @__PURE__ */ (0, M.jsx)(vs, { profile: n })
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
		render: (e, t) => t.baseline === null ? "—" : hs(t.baseline)
	}, {
		key: "change",
		label: "Change",
		sortable: !0,
		align: "right",
		cellClassName: t,
		render: (e, t) => {
			let n = t.delta;
			return n ? /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [/* @__PURE__ */ (0, M.jsx)("span", {
				className: C(n.delta_ms > 0 && "text-red-700 dark:text-red-400", n.delta_ms < 0 && "text-green-700 dark:text-green-400"),
				children: n.delta_pct === void 0 ? "—" : `${n.delta_pct >= 0 ? "+" : ""}${n.delta_pct.toFixed(1)}%`
			}), n.disk_read_bytes_delta !== void 0 && n.disk_write_bytes_delta !== void 0 && /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "text-xs text-muted-foreground",
				children: [
					"Disk R ",
					_s(n.disk_read_bytes_delta),
					" · W ",
					_s(n.disk_write_bytes_delta)
				]
			})] }) : "—";
		}
	}), n;
}
function fc({ title: e, entries: t, deltas: n }) {
	let r = (0, j.useMemo)(() => uc(t, n), [t, n]), i = (0, j.useMemo)(() => dc(!!n), [n]);
	return t.length ? /* @__PURE__ */ (0, M.jsxs)("section", {
		className: "space-y-1",
		children: [/* @__PURE__ */ (0, M.jsx)("h2", {
			className: "text-sm font-semibold",
			children: e
		}), /* @__PURE__ */ (0, M.jsx)(g, {
			data: r,
			columns: i,
			emptyMessage: `No ${e.toLowerCase()}`
		})]
	}) : null;
}
function pc(e) {
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
function mc({ performance: e, sampleIndex: t, onSampleChange: n }) {
	let r = e.profile, i = r?.samples ?? [], a = t !== void 0 && t < i.length ? t : 0, o = i[a], s = (0, j.useMemo)(() => o ? pc(o) : [], [o]);
	return /* @__PURE__ */ (0, M.jsxs)("section", {
		className: "space-y-3",
		children: [/* @__PURE__ */ (0, M.jsx)("h2", {
			className: "text-sm font-semibold",
			children: "Process profile"
		}), r ? i.length === 0 ? /* @__PURE__ */ (0, M.jsx)("p", {
			className: "text-sm text-muted-foreground",
			children: "No process samples were observed."
		}) : /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
			/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "grid grid-cols-2 gap-density-1 sm:grid-cols-5",
				children: [
					/* @__PURE__ */ (0, M.jsx)(x, {
						label: "Peak CPU",
						value: `${r.peak_cpu_percent.toFixed(1)}%`
					}),
					/* @__PURE__ */ (0, M.jsx)(x, {
						label: "Peak RSS",
						value: gs(r.peak_rss_bytes)
					}),
					r.disk_io && /* @__PURE__ */ (0, M.jsx)(x, {
						label: "Observed disk read",
						value: gs(r.disk_io.disk_read_bytes)
					}),
					r.disk_io && /* @__PURE__ */ (0, M.jsx)(x, {
						label: "Observed disk write",
						value: gs(r.disk_io.disk_write_bytes)
					}),
					/* @__PURE__ */ (0, M.jsx)(x, {
						label: "Samples",
						value: `${i.length} at ${r.interval_ms} ms`
					})
				]
			}),
			/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "grid grid-cols-2 gap-density-2",
				children: [/* @__PURE__ */ (0, M.jsx)(Zi, {
					label: "CPU %",
					values: i.map((e) => e.cpu_percent),
					format: (e) => `${e.toFixed(1)}%`
				}), /* @__PURE__ */ (0, M.jsx)(Zi, {
					label: "RSS",
					values: i.map((e) => e.rss_bytes),
					format: gs
				})]
			}),
			/* @__PURE__ */ (0, M.jsx)("div", {
				className: "flex flex-wrap gap-1",
				"aria-label": "Profile samples",
				children: i.map((e, t) => /* @__PURE__ */ (0, M.jsx)(fe, {
					type: "button",
					size: "sm",
					variant: t === a ? "default" : "outline",
					"aria-label": `Sample ${t + 1}`,
					onClick: () => n(t),
					children: new Date(e.sampled_at).toLocaleTimeString()
				}, e.sampled_at))
			}),
			o && /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "space-y-1",
				children: [/* @__PURE__ */ (0, M.jsxs)("div", {
					className: "text-sm text-muted-foreground",
					children: [
						"Sample ",
						a + 1,
						" · CPU ",
						o.cpu_percent.toFixed(1),
						"% · RSS ",
						gs(o.rss_bytes)
					]
				}), /* @__PURE__ */ (0, M.jsx)("div", {
					className: "h-64 rounded border border-border",
					children: /* @__PURE__ */ (0, M.jsx)(ra, {
						className: "h-full",
						nodes: s,
						selectedKey: null,
						expandAll: !0,
						onSelect: () => {},
						emptyLabel: "No processes were sampled"
					})
				})]
			})
		] }) : /* @__PURE__ */ (0, M.jsx)("p", {
			className: "text-sm text-muted-foreground",
			children: "Sampling in progress. Profile appears when the run finishes."
		})]
	});
}
function hc({ performance: e, tests: t, error: n, sampleIndex: r, onSampleChange: i }) {
	let { phases: a, fixtures: o } = (0, j.useMemo)(() => lc(t), [t]), s = e.comparison, c = e.profile;
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "mx-auto max-w-7xl space-y-5 p-density-4",
		children: [
			/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex flex-wrap items-baseline gap-3",
				children: [
					/* @__PURE__ */ (0, M.jsxs)("h2", {
						className: "text-lg font-semibold",
						children: ["Fixture ", e.mode === "profile" ? "profile" : "benchmark"]
					}),
					/* @__PURE__ */ (0, M.jsx)("span", {
						className: "text-sm text-muted-foreground",
						children: e.status === "running" ? "Running" : `${e.status} · ${hs(e.duration_ms)}`
					}),
					e.artifact_path && /* @__PURE__ */ (0, M.jsx)("span", {
						className: "break-all font-mono text-xs text-muted-foreground",
						children: e.artifact_path
					})
				]
			}),
			n && /* @__PURE__ */ (0, M.jsx)("div", {
				role: "alert",
				children: /* @__PURE__ */ (0, M.jsx)(wt, {
					variant: "caution",
					children: n
				})
			}),
			s && /* @__PURE__ */ (0, M.jsx)(wt, {
				variant: "note",
				children: /* @__PURE__ */ (0, M.jsxs)("div", {
					className: "space-y-1 text-sm",
					children: [
						/* @__PURE__ */ (0, M.jsxs)("div", { children: [
							"Compared with ",
							/* @__PURE__ */ (0, M.jsx)("span", {
								className: "font-mono",
								children: s.baseline
							}),
							". Changes describe these two runs only."
						] }),
						!!s.added?.length && /* @__PURE__ */ (0, M.jsxs)("div", { children: ["Added: ", s.added.join(", ")] }),
						!!s.missing?.length && /* @__PURE__ */ (0, M.jsxs)("div", { children: ["Missing: ", s.missing.join(", ")] }),
						!!s.unmeasured?.length && /* @__PURE__ */ (0, M.jsxs)("div", { children: ["Unmeasured: ", s.unmeasured.join(", ")] }),
						(s.peak_cpu_percent_delta !== void 0 || s.peak_rss_bytes_delta !== void 0) && /* @__PURE__ */ (0, M.jsxs)("div", { children: [
							"Peak change: CPU ",
							s.peak_cpu_percent_delta === void 0 ? "—" : `${s.peak_cpu_percent_delta >= 0 ? "+" : ""}${s.peak_cpu_percent_delta.toFixed(1)} points`,
							"; RSS ",
							s.peak_rss_bytes_delta === void 0 ? "—" : _s(s.peak_rss_bytes_delta)
						] })
					]
				})
			}),
			/* @__PURE__ */ (0, M.jsx)(fc, {
				title: "Run phases",
				entries: a,
				deltas: s?.deltas
			}),
			/* @__PURE__ */ (0, M.jsx)(fc, {
				title: "Fixture steps",
				entries: o,
				deltas: s?.deltas
			}),
			e.mode === "profile" || c ? /* @__PURE__ */ (0, M.jsx)(mc, {
				performance: e,
				sampleIndex: r,
				onSampleChange: i
			}) : null,
			!a.length && !o.length && /* @__PURE__ */ (0, M.jsx)("p", {
				className: "text-sm text-muted-foreground",
				children: "Waiting for setup or fixture steps…"
			})
		]
	});
}
// src/utils.ts
var gc = "Running tests across packages";
function _c(e, t) {
	return e.framework === "task" && e.name === t;
}
function vc(e) {
	return e.filter((e) => !_c(e, gc));
}
function yc(e) {
	return e.context?.status;
}
function bc(e) {
	let t = yc(e);
	return t === "running" || t === "pending";
}
function xc(e) {
	let t = e.find((e) => _c(e, gc));
	return !!t && bc(t);
}
function Sc(e, t) {
	return t && !xc(e);
}
function Cc(e, t) {
	return t && !xc(e) && e.every((e) => e.framework === "task");
}
function wc(e) {
	let t = {
		running: 0,
		done: 0,
		failed: 0
	};
	for (let n of e ?? []) n.status ? t.running++ : !n.skipped && (n.error || n.timed_out || !n.success && !n.violations?.length) ? t.failed++ : t.done++;
	return t;
}
function Tc(e) {
	return e.filter((e) => e.framework === "task" && e.name !== "Running tests across packages" && e.name !== "Running linters" && bc(e)).map((e) => ({
		...e,
		children: (e.children ?? []).filter(bc)
	}));
}
function Ec(e, t) {
	if (!e) return null;
	let n = e[t];
	if (typeof n != "string") return null;
	let r = n.trim();
	return r === "" ? null : r;
}
function Dc(e) {
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
function Oc(e) {
	return String(e.context?.status ?? "");
}
function kc({ tests: e }) {
	let t = Tc(e);
	return t.length === 0 ? null : /* @__PURE__ */ (0, M.jsxs)("section", {
		"aria-label": "Tasks",
		className: "flex flex-col gap-1 border-b border-border px-density-3 py-density-2 text-xs",
		children: [/* @__PURE__ */ (0, M.jsx)("h3", {
			className: "text-xs font-medium uppercase tracking-wide text-muted-foreground",
			children: "Tasks"
		}), t.map((e) => {
			let t = e.children?.length ? e.children : [e];
			return /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex flex-col gap-0.5",
				children: [/* @__PURE__ */ (0, M.jsx)("span", {
					className: "text-muted-foreground",
					children: e.name
				}), /* @__PURE__ */ (0, M.jsx)("ul", {
					className: "flex flex-col gap-0.5",
					children: t.map((e) => {
						let t = Oc(e) === "running";
						return /* @__PURE__ */ (0, M.jsxs)("li", {
							className: "flex items-center gap-2",
							children: [
								/* @__PURE__ */ (0, M.jsx)(w, {
									icon: t ? ht : Ze,
									className: t ? "animate-spin text-blue-500" : "text-muted-foreground"
								}),
								/* @__PURE__ */ (0, M.jsx)("span", {
									className: "truncate font-medium",
									children: e.name
								}),
								/* @__PURE__ */ (0, M.jsx)("span", {
									className: "truncate text-muted-foreground",
									children: e.message || Oc(e)
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
function Ac({ commands: e }) {
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex items-start gap-2 border-b border-border bg-muted px-4 py-2",
		"data-rerun-commands": "",
		children: [/* @__PURE__ */ (0, M.jsx)("div", {
			className: "min-w-0 flex-1 space-y-0.5 font-mono text-xs",
			children: e.map((e, t) => /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "truncate",
				title: e.cwd ? `${e.line}\ncwd: ${e.cwd}` : e.line,
				children: [/* @__PURE__ */ (0, M.jsx)("span", {
					className: "text-muted-foreground",
					children: "$ "
				}), e.line]
			}, t))
		}), /* @__PURE__ */ (0, M.jsx)(ee, {
			value: () => e.map((e) => e.shell).join("\n"),
			label: e.length > 1 ? "Copy rerun commands" : "Copy rerun command"
		})]
	});
}
var jc = 20, Mc = 250, Nc = {
	running: "info",
	success: "success",
	failed: "danger",
	canceled: "warning"
};
function Pc(e) {
	return e.stream === "stderr" ? `\x1b[31m${e.text}\x1b[0m` : e.text;
}
function Fc({ open: e, onClose: t, streamUrl: n }) {
	let [r, i] = (0, j.useState)([]), [a, o] = (0, j.useState)("running"), [s, c] = (0, j.useState)(""), [u, d] = (0, j.useState)([]), [f, p] = (0, j.useState)(void 0), m = (0, j.useRef)(null), h = (0, j.useRef)(!0), g = (0, j.useRef)("running"), _ = (0, j.useCallback)(() => {
		let e = m.current;
		e && (h.current = e.scrollHeight - e.scrollTop - e.clientHeight < 40);
	}, []), [v, y] = (0, j.useState)(0), b = (0, j.useRef)(null);
	return (0, j.useEffect)(() => {
		e && (o("running"), g.current = "running", c(""), d([]), p(void 0), y(0), h.current = !0);
	}, [e, n]), (0, j.useEffect)(() => {
		if (!e) return;
		let t = new EventSource(n);
		return t.addEventListener("open", () => i([])), t.addEventListener("message", (e) => {
			let t = JSON.parse(e.data);
			t.command && c(t.command), t.commands?.length && d(t.commands), g.current = t.status, o(t.status), t.lines?.length && i((e) => [...e, ...t.lines]);
		}), t.addEventListener("done", () => t.close()), t.onerror = () => {
			g.current === "running" ? t.readyState === EventSource.CLOSED && (v >= jc ? p("The rerun output is not available") : b.current = window.setTimeout(() => y((e) => e + 1), Mc)) : t.close();
		}, () => {
			t.close(), b.current !== null && window.clearTimeout(b.current);
		};
	}, [
		e,
		n,
		v
	]), (0, j.useEffect)(() => {
		h.current && m.current && (m.current.scrollTop = m.current.scrollHeight);
	}, [r]), /* @__PURE__ */ (0, M.jsxs)(xe, {
		open: e,
		onClose: t,
		size: "xl",
		closeOnBackdrop: !0,
		scrollBody: !1,
		title: /* @__PURE__ */ (0, M.jsx)("span", {
			className: "truncate font-mono text-sm",
			title: s || "rerun",
			children: s || "rerun"
		}),
		headerSlot: /* @__PURE__ */ (0, M.jsx)("span", {
			"data-rerun-status": a,
			children: /* @__PURE__ */ (0, M.jsx)(D, {
				tone: Nc[a],
				variant: "soft",
				size: "xs",
				children: a
			})
		}),
		children: [u.length > 0 && /* @__PURE__ */ (0, M.jsx)(Ac, { commands: u }), /* @__PURE__ */ (0, M.jsxs)("pre", {
			ref: m,
			onScroll: _,
			className: "h-[70vh] overflow-auto bg-zinc-950 p-4 font-mono text-sm leading-relaxed text-zinc-100",
			"data-rerun-output": "",
			children: [
				r.length === 0 && a === "running" && !f && /* @__PURE__ */ (0, M.jsx)("span", {
					className: "text-zinc-500",
					children: "Waiting for output..."
				}),
				r.length === 0 && a !== "running" && !f && /* @__PURE__ */ (0, M.jsx)("span", {
					className: "text-zinc-500",
					children: "The rerun wrote no output."
				}),
				f && /* @__PURE__ */ (0, M.jsx)("span", {
					className: "text-red-400",
					children: f
				}),
				/* @__PURE__ */ (0, M.jsx)(l, {
					as: "span",
					text: r.map(Pc).join("")
				})
			]
		})]
	});
}
// src/components/RunActions.tsx
var Ic = [
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
function Lc({ runMeta: e }) {
	let t = Ic.flatMap((t) => {
		let n = Ec(e?.args, t.key);
		return n ? [{
			...t,
			value: n
		}] : [];
	});
	return t.length === 0 ? null : /* @__PURE__ */ (0, M.jsxs)("span", {
		className: "inline-flex items-center gap-1",
		"data-timeout-pills": "",
		children: [/* @__PURE__ */ (0, M.jsx)(Qe, { className: "text-muted-foreground" }), t.map((e) => /* @__PURE__ */ (0, M.jsx)("span", {
			title: e.title,
			children: /* @__PURE__ */ (0, M.jsxs)(D, {
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
function Rc(e) {
	return [
		e.pid ? `PID ${e.pid}` : "",
		e.frameworks?.length ? `Frameworks ${e.frameworks.join(", ")}` : "",
		e.command ? `Command ${e.command}` : ""
	].filter(Boolean).join("\n");
}
function zc({ runMeta: e, canStop: t, stopping: n, onStop: r, exportState: i }) {
	let a = e ? Rc(e) : "";
	return /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
		a && /* @__PURE__ */ (0, M.jsx)("span", {
			className: "hidden text-xs text-muted-foreground lg:inline",
			title: a,
			children: "run info"
		}),
		/* @__PURE__ */ (0, M.jsx)(Lc, { runMeta: e }),
		t && /* @__PURE__ */ (0, M.jsxs)(fe, {
			type: "button",
			size: "sm",
			variant: "destructive",
			disabled: n,
			onClick: r,
			title: "Stop the active run",
			children: [/* @__PURE__ */ (0, M.jsx)(Tt, { className: "mr-1" }), n ? "Stopping..." : "Stop"]
		}),
		i && /* @__PURE__ */ (0, M.jsx)(os, {
			routeState: i,
			title: "Download the whole result as JSON or Markdown"
		}),
		i && /* @__PURE__ */ (0, M.jsx)(ls, { routeState: i })
	] });
}
// src/components/RunPrompts.tsx
function Bc({ owner: e, scope: t = {} }) {
	let n = (0, j.useMemo)(() => Cs(t), [t.runId]);
	return e ? /* @__PURE__ */ (0, M.jsx)($r, {
		basePath: xs(t),
		owner: e,
		lookupFetcher: n,
		connectionApiBase: Ss(t),
		autoOpen: !0
	}) : null;
}
// src/runHistory.ts
async function Vc({ before: e, signal: t } = {}) {
	let n = e ? `?before=${encodeURIComponent(e)}` : "", r = await fetch(Vo(`/api/runs${n}`), t ? { signal: t } : {});
	if (r.status === 404) return null;
	if (!r.ok) throw Error(`GET /api/runs: ${r.status} ${(await r.text()).trim()}`);
	return await r.json();
}
function Hc(e) {
	return e.kind === "lint" ? `${e.href}lint` : e.href;
}
function Uc(e) {
	return e.started ? new Date(e.started).toLocaleString(void 0, {
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit"
	}) : e.runId;
}
function Wc(e) {
	let t = [];
	return e.total > 0 && t.push(e.failed > 0 ? `${e.failed} failed / ${e.total} tests` : `${e.total} tests`), (e.kind === "lint" || e.kind === "test+lint") && t.push(`${e.lintViolations} lint`), t.join(" · ") || "no results";
}
function Gc(e) {
	if (!e.started || !e.ended) return;
	let t = Date.parse(e.ended) - Date.parse(e.started);
	return t >= 0 ? t : void 0;
}
// src/components/RunHistory.tsx
function Kc({ run: e }) {
	return /* @__PURE__ */ (0, M.jsxs)("span", {
		className: "inline-flex flex-wrap items-center gap-1",
		children: [
			e.total > 0 && /* @__PURE__ */ (0, M.jsx)(D, {
				tone: e.failed > 0 ? "danger" : "success",
				variant: "outline",
				size: "xs",
				children: e.failed > 0 ? `${e.failed} failed` : `${e.passed} passed`
			}),
			e.total > 0 && /* @__PURE__ */ (0, M.jsxs)("span", {
				className: "text-xs text-muted-foreground",
				children: ["of ", e.total]
			}),
			(e.kind === "lint" || e.kind === "test+lint") && /* @__PURE__ */ (0, M.jsxs)(D, {
				tone: e.lintViolations > 0 ? "warning" : "neutral",
				variant: "outline",
				size: "xs",
				children: [e.lintViolations, " lint"]
			})
		]
	});
}
function qc({ history: e, navigate: t }) {
	let n = (0, j.useMemo)(() => [
		{
			key: "started",
			label: "Started",
			kind: "timestamp",
			sortable: !0,
			render: (t, n) => /* @__PURE__ */ (0, M.jsxs)("span", {
				className: "inline-flex items-center gap-1.5",
				children: [
					/* @__PURE__ */ (0, M.jsx)("a", {
						href: Hc(n),
						className: "text-primary no-underline hover:underline",
						onClick: (e) => e.stopPropagation(),
						children: n.started ? new Date(n.started).toLocaleString() : n.runId
					}),
					e.current === n.runId && /* @__PURE__ */ (0, M.jsx)(D, {
						tone: "info",
						variant: "soft",
						size: "xs",
						children: "viewing"
					}),
					!e.current && e.latest.includes(n.runId) && /* @__PURE__ */ (0, M.jsx)(D, {
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
			key: "fixtureFiles",
			label: "Fixture",
			filterable: !0,
			accessor: (e) => (e.fixtureFiles ?? []).join(", "),
			render: (e) => /* @__PURE__ */ (0, M.jsx)("span", {
				className: "break-all",
				children: String(e)
			})
		},
		{
			key: "results",
			label: "Results",
			accessor: (e) => e.failed + e.lintViolations,
			sortable: !0,
			render: (e, t) => /* @__PURE__ */ (0, M.jsx)(Kc, { run: t })
		},
		{
			key: "duration",
			label: "Duration",
			accessor: (e) => Gc(e) ?? null,
			sortable: !0,
			align: "right",
			shrink: !0,
			render: (e) => typeof e == "number" ? Ge(e) : ""
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
			render: (e) => /* @__PURE__ */ (0, M.jsx)("code", {
				className: "text-xs",
				children: String(e ?? "").slice(0, 8)
			})
		}
	], [e.current, e.latest]);
	return /* @__PURE__ */ (0, M.jsx)(g, {
		data: e.runs,
		columns: n,
		getRowId: (e) => e.runId,
		onRowClick: (e) => t(Hc(e)),
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
function Jc({ trace: e, selectedKey: t, onSelect: n, hideExited: r, onHideExitedChange: i, expandAll: a, onExpandAllChange: o, stepLinks: s, onOpenStep: c, className: l }) {
	let u = (0, j.useMemo)(() => e.nodes.find((e) => e.key === t) ?? null, [e.nodes, t]), d = u ? s?.get(u.pid) : void 0, f = e.canCollect && lo(e.samples).length > 0, m = !!u && e.goprof.some((e) => e.key === u.key);
	return /* @__PURE__ */ (0, M.jsx)(ot, {
		className: l ?? "h-full min-h-0",
		defaultSplit: 45,
		minLeft: 25,
		minRight: 30,
		leftClass: "border-r-2 border-border bg-muted/30",
		left: /* @__PURE__ */ (0, M.jsx)(ra, {
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
		right: /* @__PURE__ */ (0, M.jsxs)("div", {
			className: "flex h-full min-h-0 flex-col",
			children: [
				u && f && /* @__PURE__ */ (0, M.jsxs)("div", {
					className: "flex shrink-0 items-center justify-end gap-2 border-b border-border px-density-4 py-density-1",
					children: [e.refreshError && /* @__PURE__ */ (0, M.jsx)("span", {
						className: "min-w-0 truncate text-xs text-red-700 dark:text-red-400",
						role: "alert",
						title: e.refreshError,
						children: e.refreshError
					}), /* @__PURE__ */ (0, M.jsxs)(fe, {
						type: "button",
						size: "sm",
						variant: "outline",
						disabled: e.refreshing,
						title: "Read the open files, sockets, ports and connections of this process in full on the next sample",
						onClick: () => void e.refreshFacets(u.key),
						children: [e.refreshing ? /* @__PURE__ */ (0, M.jsx)(ht, { className: "mr-1 animate-spin" }) : /* @__PURE__ */ (0, M.jsx)(p, { className: "mr-1" }), "Refresh facets"]
					})]
				}),
				/* @__PURE__ */ (0, M.jsx)(ko, {
					className: "min-h-0 flex-1",
					node: u,
					samples: e.samples,
					stacks: e.stacks,
					...e.canCollect ? { onCollectStack: e.collectStack } : {},
					collecting: e.collecting,
					...e.collectError ? { collectError: e.collectError } : {},
					...d && c ? { links: /* @__PURE__ */ (0, M.jsx)(F, {
						title: "Step",
						children: /* @__PURE__ */ (0, M.jsxs)(fe, {
							type: "button",
							variant: "link",
							size: "sm",
							className: "h-auto p-0",
							onClick: () => c(d),
							children: [/* @__PURE__ */ (0, M.jsx)(_t, { className: "mr-1" }), d.name]
						})
					}) } : {}
				}),
				u && m && /* @__PURE__ */ (0, M.jsx)("div", {
					className: "max-h-[45%] shrink-0 overflow-y-auto border-t border-border px-density-4 py-density-2",
					"data-go-profiles": "",
					children: /* @__PURE__ */ (0, M.jsxs)(Ie, {
						title: "Go profiles",
						defaultOpen: !0,
						children: [/* @__PURE__ */ (0, M.jsx)(Wi, {
							records: e.goprof,
							processKey: u.key
						}), e.goprofExportUrl && /* @__PURE__ */ (0, M.jsx)("a", {
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
var Yc = "";
function Xc({ history: e, navigate: t }) {
	let n = e.current ?? Yc, r = [{
		value: Yc,
		label: "Latest results"
	}, ...e.runs.map((t) => ({
		value: t.runId,
		label: `${Uc(t)} · ${t.kind}${t.fixtureFiles?.length ? ` · ${t.fixtureFiles.join(", ")}` : ""} · ${Wc(t)}${e.latest.includes(t.runId) ? " (latest)" : ""}`
	}))];
	return n !== Yc && !e.runs.some((e) => e.runId === n) && r.splice(1, 0, {
		value: n,
		label: n
	}), /* @__PURE__ */ (0, M.jsxs)("label", {
		className: "inline-flex min-w-0 items-center gap-1 text-sm",
		"data-run-switcher": "",
		children: [
			/* @__PURE__ */ (0, M.jsx)(rn, {
				className: "shrink-0 text-muted-foreground",
				"aria-hidden": !0
			}),
			/* @__PURE__ */ (0, M.jsx)("span", {
				className: "sr-only",
				children: "Run"
			}),
			/* @__PURE__ */ (0, M.jsx)(Ke, {
				className: "h-7 max-w-[22rem] truncate py-0 text-xs",
				value: n,
				options: r,
				onChange: (r) => {
					let i = r.target.value;
					if (i === n) return;
					let a = e.runs.find((e) => e.runId === i);
					t(i === Yc ? e.latestHref : a ? Hc(a) : e.latestHref);
				}
			})
		]
	});
}
// src/components/ShardFilter.tsx
function Zc(e, t, n) {
	let r = n.pathname.startsWith(n.basePath) ? n.pathname.slice(n.basePath.length) : "", i = e.shards.map((e) => e.artifactId).filter((e) => t.includes(e));
	return `${e.base}/${i.join(",")}${r}${n.search}`;
}
function Qc(e, t) {
	let n = e.shards.filter((e) => e.selected).map((e) => e.artifactId);
	if (!n.includes(t)) return [...n, t];
	let r = n.filter((e) => e !== t);
	return r.length > 0 ? r : null;
}
var $c = "inline-flex items-center rounded-full border px-2 py-0.5 text-xs no-underline", el = `${$c} border-primary bg-primary/10 text-primary`, tl = `${$c} border-border text-muted-foreground hover:bg-muted`;
function nl({ page: e, location: t }) {
	if (e.shards.length < 2) return null;
	let n = e.shards.map((e) => e.artifactId), r = e.shards.every((e) => e.selected);
	return /* @__PURE__ */ (0, M.jsxs)("nav", {
		"aria-label": "Shards",
		className: "flex min-w-0 flex-wrap items-center gap-1",
		children: [
			/* @__PURE__ */ (0, M.jsx)("span", {
				className: "text-xs text-muted-foreground",
				children: "Shards"
			}),
			/* @__PURE__ */ (0, M.jsx)("a", {
				href: Zc(e, n, t),
				className: r ? el : tl,
				"aria-current": r ? "true" : void 0,
				children: "All"
			}),
			e.shards.map((n) => {
				let r = Qc(e, n.artifactId), i = n.selected ? el : tl;
				return r ? /* @__PURE__ */ (0, M.jsx)("a", {
					href: Zc(e, r, t),
					className: i,
					"aria-current": n.selected ? "true" : void 0,
					children: n.label
				}, n.artifactId) : /* @__PURE__ */ (0, M.jsx)("span", {
					className: i,
					"aria-current": n.selected ? "true" : void 0,
					title: "At least one shard stays selected",
					children: n.label
				}, n.artifactId);
			})
		]
	});
}
// src/processMetrics.ts
var rl = (e) => e.status === "exited" || !!e.exitedAt;
function il({ samples: e, nodes: t }) {
	return async (n) => {
		let r = t.filter((e) => e.pid === n);
		if (r.length > 0 && r.every(rl)) throw new oe(n);
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
var al = 1e3, ol = 2e3, sl = "The run has ended, so its process trace can no longer be checkpointed";
async function cl(e, t) {
	try {
		return await Is(e, t);
	} catch (e) {
		if (e instanceof Es) return [];
		throw e;
	}
}
async function ll(e, t) {
	let n = await Rs(e, t);
	return n === 0 ? [] : cl(e, {
		...t,
		afterSeq: Math.max(0, n - ol)
	});
}
function ul(e) {
	return e instanceof Error ? e.message : String(e);
}
function dl(e, t, n, r) {
	let [i, a] = (0, j.useState)(null), [o, s] = (0, j.useState)(void 0), c = (0, j.useRef)(n);
	(0, j.useEffect)(() => {
		c.current = n;
	}, [n]);
	let l = (0, j.useCallback)(async () => {
		try {
			a(await c.current()), s(void 0);
		} catch (e) {
			s(ul(e));
		}
	}, []);
	return (0, j.useEffect)(() => {
		if (!e) return;
		let n = !1, r = () => {
			c.current().then((e) => {
				n || (a(e), s(void 0));
			}).catch((e) => {
				n || s(ul(e));
			});
		};
		if (r(), !t) return () => {
			n = !0;
		};
		let i = window.setInterval(r, al);
		return () => {
			n = !0, window.clearInterval(i);
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
function fl(e, t, n) {
	let [r, i] = (0, j.useState)(!1), [a, o] = (0, j.useState)(void 0);
	return {
		busy: r,
		error: a,
		checkpoint: (0, j.useCallback)(async (r) => {
			i(!0), o(void 0);
			try {
				if (!e || !t) {
					o(sl);
					return;
				}
				await Bs({
					mount: e,
					session: t,
					params: r
				}), await n();
			} catch (e) {
				o(ul(e));
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
function pl({ traces: e, done: t, selectedKey: n, base: r }) {
	let i = r ?? Bo(), a = (0, j.useMemo)(() => ks(e, i), [e, i]), o = Os(a), s = a?.process, c = !t, l = !!o && !!s, u = c ? e?.sessions : void 0, d = dl(l, c, () => cl(o, s.tree), `${o}|${s?.tree.stream}`), f = dl(l, c, () => ll(o, s.samples), `${o}|${s?.samples.stream}`), p = dl(l && !!n, c, () => cl(o, s.stacks), `${o}|${s?.stacks.stream}`), m = dl(l && !!n && !!s?.goprof, c, () => cl(o, s.goprof), `${o}|${s?.goprof?.stream}`), h = fl(o, u?.stacks, p.reload), g = fl(o, u?.samples, f.reload), _ = (0, j.useCallback)((e, t) => h.checkpoint({
		key: e,
		transport: t,
		reason: "manual"
	}), [h.checkpoint]), v = (0, j.useCallback)((e) => g.checkpoint({ key: e }), [g.checkpoint]);
	return {
		nodes: (0, j.useMemo)(() => d.value ?? [], [d.value]),
		samples: (0, j.useMemo)(() => f.value ?? [], [f.value]),
		stacks: (0, j.useMemo)(() => p.value ?? [], [p.value]),
		goprof: (0, j.useMemo)(() => m.value ?? [], [m.value]),
		...o && s?.goprof ? { goprofExportUrl: Ms(o, s.goprof, "pprof") } : {},
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
function ml() {
	let [e, t] = (0, j.useState)([]), [n, r] = (0, j.useState)(!1), [i, a] = (0, j.useState)(!0), [o, s] = (0, j.useState)(), c = (0, j.useRef)(!1), l = (0, j.useCallback)(async (e) => {
		if (!c.current) {
			c.current = !0, a(!0);
			try {
				let n = await Vc(e ? { before: e } : {});
				n && (r(!0), t((t) => e ? [...t, n] : [n])), s(void 0);
			} catch (e) {
				s(e instanceof Error ? e.message : String(e));
			} finally {
				c.current = !1, a(!1);
			}
		}
	}, []);
	(0, j.useEffect)(() => {
		l();
	}, [l]);
	let u = e[0], d = e.at(-1)?.next, f = (0, j.useCallback)(() => {
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
var hl = 2e3, gl = { running: !1 };
function _l(e) {
	return new EventSource(e);
}
function vl() {
	return typeof window < "u" && window.__gavelBasePath || "";
}
function yl() {
	return typeof globalThis < "u" && globalThis.EventSource !== void 0;
}
function bl(e) {
	let t = e.status || gl, n = e.tests || [];
	if (!t.running) {
		if (t.stopped) return t.stop_message || "Stopped by user";
		if (e.metadata?.kind === "rerun") return "Rerun complete";
		if (Cc(n, !!t.lint_run)) {
			let { done: t, failed: n } = wc(e.lint);
			return `Lint run complete (${t} done · ${n} failed)`;
		}
		return "Test run complete";
	}
	if (e.metadata?.kind === "rerun") return `Running rerun #${e.metadata.sequence || 1}...`;
	if (e.performance) return "Running fixtures...";
	if (Sc(n, t.running)) {
		if (!e.lint?.length) return "Running linters...";
		let { running: t, done: n, failed: r } = wc(e.lint);
		return `Running linters... (${t} running · ${n} done · ${r} failed)`;
	}
	return "Running tests...";
}
function V(e, t, n, r) {
	let i = new AbortController(), { signal: a } = i, o, s = async () => {
		try {
			let t = await fetch(`${e}/api/tests`, {
				headers: { Accept: "application/json" },
				signal: a
			});
			if (a.aborted) return;
			if (t.ok) {
				let e = await t.json();
				if (a.aborted || (n(e), !e.status?.running)) return;
			}
		} catch {
			a.aborted || r();
		}
		a.aborted || (o = setTimeout(s, t));
	};
	return s(), () => {
		i.abort(), clearTimeout(o);
	};
}
function xl(e = {}) {
	let { baseUrl: t, basePath: n, runId: r, enabled: i = !0, pollMs: a = hl, forcePoll: o = !1, createEventSource: s } = e, c = t ?? n ?? vl(), l = r ? `${c}/${encodeURIComponent(r)}` : c, [u, d] = (0, j.useState)(void 0), [f, p] = (0, j.useState)("Loading..."), [m, h] = (0, j.useState)(!1), [g, _] = (0, j.useState)(void 0), [v, y] = (0, j.useState)(0), b = (0, j.useRef)(!1), x = (0, j.useRef)(s ?? _l);
	(0, j.useEffect)(() => {
		x.current = s ?? _l;
	}, [s]);
	let ee = (0, j.useCallback)((e) => {
		let t = !!(e.status && e.status.running);
		d(e), p(bl(e)), _(void 0), t ? (b.current = !1, h(!1)) : (b.current = !0, h(!0));
	}, []);
	(0, j.useEffect)(() => {
		if (!i) return;
		if (b.current = !1, h(!1), o || !yl()) return V(l, a, ee, () => {
			b.current || _("Connection lost — retrying...");
		});
		let e = !1;
		fetch(`${l}/api/tests`, { headers: { Accept: "application/json" } }).then((e) => {
			if (!e.ok) throw Error(`GET /api/tests: ${e.status}`);
			return e.json();
		}).then((t) => {
			e || ee(t);
		}).catch(() => {});
		let t = x.current(`${l}/api/tests/stream`), n = (e) => {
			let n = JSON.parse(e.data);
			ee(n), n.status?.running || t.close();
		}, r = () => {
			b.current = !0, h(!0), d((e) => e && {
				...e,
				status: {
					...e.status,
					running: !1
				}
			}), t.close();
		};
		return t.addEventListener("message", n), t.addEventListener("done", r), t.onerror = () => {
			b.current || _("Connection lost — retrying...");
		}, () => {
			e = !0, t.removeEventListener("message", n), t.removeEventListener("done", r), t.close();
		};
	}, [
		l,
		i,
		a,
		o,
		v,
		ee
	]);
	let te = (0, j.useCallback)(async (e = {}) => {
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
	}, [l]), ne = (0, j.useCallback)(async (e, t) => {
		let n = await fetch(`${l}/api/rerun/command?${new URLSearchParams({
			fixture: e,
			from_step: t
		})}`);
		if (!n.ok) throw Error((await n.text()).trim());
		return (await n.json()).shell;
	}, [l]), re = (0, j.useCallback)(async (e) => {
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
	}, [l]), S = (0, j.useCallback)(async (e) => {
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
	}, [l]), ie = (0, j.useCallback)(() => y((e) => e + 1), []);
	return {
		snapshot: u,
		tests: u?.tests ?? [],
		lint: u?.lint,
		bench: u?.bench,
		status: u?.status ?? gl,
		statusText: f,
		runMeta: u?.metadata,
		done: m,
		error: g,
		rerun: te,
		rerunCommand: ne,
		stop: re,
		editTest: S,
		refetch: ie
	};
}
// src/lintRerun.ts
function Sl(e) {
	return e.replace(/\\/g, "/").replace(/\/+/g, "/").replace(/\/$/, "");
}
function Cl(e, t) {
	if (!e) return;
	let n = Sl(e);
	if (!t) return n;
	let r = `${Sl(t)}/`;
	return n.startsWith(r) ? n.slice(r.length) : n;
}
function wl(e, t) {
	e.work_dir && t.add(e.work_dir), (e.children ?? []).forEach((e) => wl(e, t));
}
function Tl(e, t) {
	let n = /* @__PURE__ */ new Set(), r = /* @__PURE__ */ new Set(), i = /* @__PURE__ */ new Set(), a = (a, o, s) => {
		for (let c of t) {
			if (e?.work_dir && c.work_dir && c.work_dir !== e.work_dir || o && c.linter !== o) continue;
			let t = !1;
			for (let e of c.violations ?? []) {
				let n = Cl(e.file, c.work_dir);
				s && (e.rule?.method || "(no rule)") !== s || (n ? (!a || n === a || n.startsWith(`${a}/`)) && (t = !0, i.add(n)) : t = s !== void 0 && a === void 0);
			}
			t && (c.work_dir && n.add(c.work_dir), r.add(c.linter));
		}
	};
	if (e) e.kind === "linter" || e.kind === "linter-error" ? (e.linterName && r.add(e.linterName), wl(e, n)) : ((e.kind === "lint-rule" || e.kind === "lint-rule-group") && e.linterName && r.add(e.linterName), a(e.target_path, e.linterName, e.ruleName));
	else for (let e of t) r.add(e.linter), e.work_dir && n.add(e.work_dir);
	return e?.work_dir && n.add(e.work_dir), {
		lint: !0,
		work_dir: n.size === 1 ? Array.from(n)[0] : "",
		lint_linters: Array.from(r),
		lint_files: Array.from(i)
	};
}
// src/exportPaths.ts
var El = "go test";
function Dl(e) {
	let t = "";
	for (let n = 0; n < e.length; n++) {
		let r = e[n], i = n > 0 ? e[n - 1] : "";
		n > 0 && /\p{Lu}/u.test(r) && (/\p{Ll}/u.test(i) || /\p{Nd}/u.test(i)) && (t += " "), t += r;
	}
	return t;
}
function Ol(e, t) {
	return t !== El || e.endsWith("/") ? e : e.split("/").map((e, t) => t === 0 ? Dl(e.replace(/^Test/, "")) : e.replaceAll("_", " ")).join(" / ");
}
function kl(e) {
	return e.toLowerCase().replace(/[^\p{L}\p{N}]+/gu, "-").replace(/^-+|-+$/g, "") || "node";
}
function Al(e, t = []) {
	let n = e.map((e) => kl(Ol(e.name, e.framework))), r = /* @__PURE__ */ new Map();
	n.forEach((e) => r.set(e, (r.get(e) ?? 0) + 1));
	let i = /* @__PURE__ */ new Map();
	return e.map((e, a) => {
		let o = n[a], s = (i.get(o) ?? 0) + 1;
		i.set(o, s);
		let c = [...t, r.get(o) > 1 ? `${o}~${s}` : o];
		return {
			...e,
			export_path: c.join("/"),
			...e.children ? { children: Al(e.children, c) } : {}
		};
	});
}
// src/App.tsx
function jl(e) {
	return e.reduce((e, t) => e + +!!t.fixture + jl(t.children || []), 0);
}
var Ml = (e) => e, Nl = (0, j.lazy)(() => import("./chunks/FilesTab-B1iR4D-6.js").then((e) => e.t)), Pl = (0, j.lazy)(() => import("./chunks/PromptRunView-Jl9BoF04.js"));
function Fl(e) {
	return e.run_id ? /* @__PURE__ */ (0, M.jsx)(j.Suspense, {
		fallback: /* @__PURE__ */ (0, M.jsx)("p", {
			role: "status",
			className: "text-sm text-muted-foreground",
			children: "Loading the prompt run…"
		}),
		children: /* @__PURE__ */ (0, M.jsx)(Pl, {
			result: e,
			baseUrl: Bo()
		})
	}) : null;
}
function Il() {
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex h-full items-center justify-center text-sm text-muted-foreground",
		role: "status",
		children: [/* @__PURE__ */ (0, M.jsx)(w, {
			icon: ht,
			className: "mr-2 animate-spin text-blue-500"
		}), "Loading files..."]
	});
}
function Ll(e, t) {
	return t ? e.map((e) => ({
		...e,
		work_dir: e.work_dir || t,
		...e.children ? { children: Ll(e.children, t) } : {}
	})) : e;
}
function Rl(e) {
	return [
		e.workDir,
		e.file,
		e.line,
		e.rule,
		e.message
	].join("\0");
}
function zl(e, t) {
	if (!e) return;
	if (!t) return e;
	let n = new URL(e);
	return n.searchParams.has("line") || n.searchParams.set("line", String(t)), n.toString();
}
async function Bl(e, t) {
	let n = new URLSearchParams();
	e.workDir && n.set("work_dir", e.workDir), n.set("path", e.file);
	let r = await fetch(Vo(`/api/file?${n}`, t)), i = await r.text();
	if (!r.ok) throw Error(`${r.status} ${i.trim() || r.statusText}`);
	return i;
}
function Vl(e) {
	return e.map((e) => ({
		...e,
		own_name: e.name,
		...e.children ? { children: Vl(e.children) } : {}
	}));
}
function Hl(e) {
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
function Ul({ icon: e, title: t, status: n, actions: r }) {
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex min-w-0 items-center gap-3 border-b border-border bg-card px-density-4 py-density-2",
		"data-page-header": "",
		children: [
			/* @__PURE__ */ (0, M.jsxs)("h1", {
				className: "inline-flex shrink-0 items-center gap-1.5 text-base font-bold",
				children: [/* @__PURE__ */ (0, M.jsx)(w, {
					icon: e,
					className: "text-primary"
				}), t]
			}),
			n && /* @__PURE__ */ (0, M.jsx)("span", {
				className: "truncate text-sm text-muted-foreground",
				children: n
			}),
			r && /* @__PURE__ */ (0, M.jsx)("div", {
				className: "ml-auto flex shrink-0 items-center gap-1",
				children: r
			})
		]
	});
}
function Wl({ done: e }) {
	return /* @__PURE__ */ (0, M.jsx)("div", {
		className: "flex h-full items-center justify-center p-8 text-center text-sm text-muted-foreground",
		children: e ? "No tests were run" : /* @__PURE__ */ (0, M.jsxs)("div", { children: [/* @__PURE__ */ (0, M.jsx)(w, {
			icon: ht,
			className: "animate-spin text-3xl text-blue-500"
		}), /* @__PURE__ */ (0, M.jsx)("p", {
			className: "mt-2",
			children: "Waiting for test results..."
		})] })
	});
}
var Gl = 8e3, Kl = "__global__", ql = (e) => window.location.assign(e);
function Jl({ runId: e, navigate: t = ql } = {}) {
	let n = (0, j.useMemo)(() => e ? { runId: e } : {}, [e]), r = Bo(n), i = xl(n.runId ? { runId: n.runId } : {}), { snapshot: a, tests: o, lint: s, status: c, runMeta: l, done: u, statusText: d, error: f } = i, m = a?.performance, g = a?.error, _ = !!c.lint_run, v = ml(), [y, b] = (0, j.useState)(() => Zo(window.location)), [x, ee] = (0, j.useState)(null), [te, ne] = (0, j.useState)(!1), [re, S] = (0, j.useState)(!1), [ie, ae] = (0, j.useState)(!1), [oe, se] = (0, j.useState)(!1), [ce, le] = (0, j.useState)(null), [ue, de] = (0, j.useState)(void 0), C = (0, j.useRef)(null), w = (0, j.useRef)(Date.now()), [pe, me] = (0, j.useState)(Date.now()), T = (0, j.useCallback)((e) => {
		de(e), C.current !== null && window.clearTimeout(C.current), C.current = window.setTimeout(() => de(void 0), Gl);
	}, []);
	(0, j.useEffect)(() => () => {
		C.current !== null && window.clearTimeout(C.current);
	}, []), (0, j.useEffect)(() => {
		if (u) return;
		let e = window.setInterval(() => me(Date.now()), 1e3);
		return () => window.clearInterval(e);
	}, [u]);
	let ge = l?.started ? Date.parse(l.started) : w.current, _e = l?.ended ? Date.parse(l.ended) : null, E = (0, j.useCallback)((e, t = "push") => {
		b(e);
		let n = Qo(e);
		n !== `${window.location.pathname}${window.location.search}` && (t === "replace" ? window.history.replaceState({}, "", n) : window.history.pushState({}, "", n));
	}, []);
	(0, j.useEffect)(() => {
		let e = () => b(Zo(window.location));
		return window.addEventListener("popstate", e), () => window.removeEventListener("popstate", e);
	}, []), (0, j.useEffect)(() => {
		m && y.tab === "tests" && window.location.pathname === `${Bo()}/` && E({
			...y,
			tab: "bench",
			selectedPath: ""
		}, "replace");
	}, [
		m,
		y,
		E
	]);
	let ve = (0, j.useMemo)(() => Sc(o, !!c.running) ? vc(o) : o, [o, c.running]), D = (0, j.useMemo)(() => st(ve), [ve]), ye = (0, j.useMemo)(() => Dt(et(Al(ct(Vl(ve), y.filters.status, y.filters.framework)))), [ve, y.filters]), be = (0, j.useMemo)(() => y.tab === "tests" && y.selectedPath ? Ht(ye, y.selectedPath) : null, [
		y.tab,
		y.selectedPath,
		ye
	]), xe = (0, j.useMemo)(() => zt({
		results: s ?? [],
		grouping: y.lintGrouping,
		filters: y.lintFilters
	}), [
		s,
		y.lintGrouping,
		y.lintFilters
	]), O = y.tab === "lint" && y.selectedPath ? Ot(xe, y.selectedPath) : null, k = (0, j.useMemo)(() => Et(s ?? []), [s]), Se = a?.git?.root, { findings: we, linterByFinding: Te, sourceUrlByFinding: Ee } = (0, j.useMemo)(() => {
		let e = /* @__PURE__ */ new Map(), t = /* @__PURE__ */ new Map();
		return {
			findings: [...(s ?? []).flatMap((n) => {
				let r = n.work_dir || Se;
				return Ft([{
					...n,
					...r ? { work_dir: r } : {}
				}]).map((r) => {
					e.set(Rl(r), n.linter);
					let i = zl(n.sources?.[r.file]?.url, r.line);
					return i && t.set(Rl(r), i), r;
				});
			}), ...Rt(Ll(o, Se))],
			linterByFinding: e,
			sourceUrlByFinding: t
		};
	}, [
		s,
		o,
		Se
	]), A = (0, j.useMemo)(() => {
		if (y.tab !== "files" || !y.selectedPath && !y.folder) return null;
		let e = {
			file: y.selectedPath,
			...y.workDir ? { workDir: y.workDir } : {}
		};
		return {
			kind: y.folder ? "folder" : "file",
			ref: e
		};
	}, [
		y.tab,
		y.selectedPath,
		y.folder,
		y.workDir
	]), De = (0, j.useMemo)(() => ks(c.traces, r), [c.traces, r]), Oe = !!Os(c.traces), Ae = pl({
		traces: c.traces,
		done: u,
		base: r,
		...y.tab === "processes" && y.selectedPath ? { selectedKey: y.selectedPath } : {}
	}), je = (0, j.useMemo)(() => Ae.nodes.filter((e) => !Ai(e)).length, [Ae.nodes]), Me = (0, j.useMemo)(() => Ae.nodes.find((e) => e.isRoot) ?? null, [Ae.nodes]), Ne = y.tab === "processes" ? Ae.nodes.find((e) => e.key === y.selectedPath) ?? null : null, Pe = (0, j.useMemo)(() => Oe ? Dc(Ml(Dt(et(ve)))) : /* @__PURE__ */ new Map(), [Oe, ve]);
	(0, j.useEffect)(() => {
		y.tab === "processes" && !y.selectedPath && Me && E({
			...y,
			selectedPath: Me.key
		}, "replace");
	}, [
		y,
		Me,
		E
	]), (0, j.useEffect)(() => {
		if (!y.selectedPath) return;
		let e = y.tab === "tests" ? o.length > 0 || u : y.tab === "lint" ? _ || u || !!s : y.tab !== "processes" || Ae.nodes.length > 0, t = y.tab === "tests" ? be : y.tab === "lint" ? O : y.tab !== "processes" || Ne;
		e && !t && E({
			...y,
			selectedPath: y.tab === "processes" && Me ? Me.key : ""
		}, "replace");
	}, [
		y,
		be,
		O,
		Ne,
		o.length,
		u,
		_,
		s,
		Ae.nodes.length,
		Me,
		E
	]);
	let Fe = (0, j.useCallback)((e) => {
		let { workDir: t, line: n, folder: r, ...i } = y;
		E({
			...i,
			tab: e,
			selectedPath: ""
		});
	}, [y, E]), Ie = (0, j.useCallback)((e) => E({
		...y,
		selectedPath: e?.route_path ?? ""
	}), [y, E]), Le = (0, j.useCallback)((e) => E({
		...y,
		tab: "tests",
		filters: e,
		selectedPath: ""
	}), [y, E]), Re = (0, j.useCallback)((e) => E({
		...y,
		tab: "lint",
		lintFilters: e,
		selectedPath: ""
	}), [y, E]), ze = (0, j.useCallback)((e) => E({
		...y,
		tab: "lint",
		lintGrouping: e,
		selectedPath: ""
	}), [y, E]), Be = (0, j.useCallback)((e) => E({
		...y,
		selectedPath: e ?? ""
	}), [y, E]), Ve = (0, j.useCallback)((e) => E({
		...y,
		tab: "processes",
		selectedPath: e
	}), [y, E]), He = (0, j.useCallback)((e) => E({
		...y,
		tab: "processes",
		hideExited: e
	}, "replace"), [y, E]), Ue = (0, j.useCallback)((e, t) => {
		let { workDir: n, line: r, folder: i, ...a } = y, { ref: o } = e;
		E({
			...a,
			tab: "files",
			selectedPath: o.file,
			...e.kind === "folder" ? { folder: !0 } : {},
			...o.workDir ? { workDir: o.workDir } : {},
			...t ? { line: t } : {}
		});
	}, [y, E]), Ge = (0, j.useCallback)((e, t) => Ue({
		kind: "file",
		ref: e
	}, t), [Ue]), Ke = (0, j.useCallback)((e) => Ge({
		file: e.file,
		...e.workDir || Se ? { workDir: e.workDir || Se } : {}
	}, e.line), [Ge, Se]), qe = (0, j.useCallback)((e) => Ue(e), [Ue]), Je = (0, j.useCallback)((e) => {
		let { line: t, ...n } = y;
		E({
			...n,
			...e ? { line: e } : {}
		}, "replace");
	}, [y, E]), Ye = (0, j.useCallback)((e) => E({
		...y,
		view: e
	}, "replace"), [y, E]), Xe = (0, j.useCallback)((e) => E({
		...y,
		problemFilters: e
	}, "replace"), [y, E]), Ze = (0, j.useCallback)((e) => Bl(e, n), [n]), Qe = (0, j.useCallback)((e) => E({
		...y,
		sampleIndex: e
	}), [y, E]), $e = (0, j.useCallback)((e) => {
		E({
			...y,
			tab: "tests",
			filters: {
				status: /* @__PURE__ */ new Map(),
				framework: /* @__PURE__ */ new Map()
			},
			selectedPath: e.routePath
		});
	}, [y, E]), tt = (0, j.useCallback)(async (e, t) => {
		if (!te) {
			ne(!0), S(!0), T(`Rerunning ${t}...`);
			try {
				let t = await i.rerun(e);
				t.conflict ? T("Another rerun is in progress") : t.ok || T(`Rerun failed: ${t.message ?? t.status}`);
			} catch (e) {
				T(`Rerun error: ${e instanceof Error ? e.message : String(e)}`);
			} finally {
				ne(!1);
			}
		}
	}, [
		te,
		i,
		T
	]), nt = (0, j.useCallback)((e) => {
		if (e.framework === "task") return;
		let t = Hl(e);
		"error" in t ? T(t.error) : tt(t.body, e.name);
	}, [tt, T]), rt = (0, j.useCallback)((e, t) => {
		if (!e.file) {
			T(`Cannot retry from step ${t}: it has no fixture file`);
			return;
		}
		let n = Hl(e);
		"error" in n ? T(n.error) : tt({
			fixture: e.file,
			from_step: t,
			work_dir: n.body.work_dir ?? ""
		}, `${e.name} from step ${t}`);
	}, [tt, T]), it = (0, j.useCallback)(async (e, t) => {
		try {
			await ke(await i.rerunCommand(e.file ?? "", t)), T("Rerun command copied");
		} catch (e) {
			T(`Cannot copy the rerun command: ${e instanceof Error ? e.message : String(e)}`);
		}
	}, [i, T]), at = (0, j.useCallback)(() => {
		let e = Tl(O, s ?? []);
		e.work_dir ? tt(e, O?.name ?? "lint") : T("Rerun across multiple lint roots is not supported");
	}, [
		O,
		s,
		tt,
		T
	]), ot = (0, j.useCallback)(async ({ taskId: e, label: t }) => {
		if (!ce) {
			le(e ?? Kl), T(`Stopping ${t}...`);
			try {
				let t = await i.stop(e);
				t.ok || T(`Stop failed: ${t.message ?? t.status}`);
			} catch (e) {
				T(`Stop error: ${e instanceof Error ? e.message : String(e)}`);
			} finally {
				le(null);
			}
		}
	}, [
		ce,
		i,
		T
	]), lt = (0, j.useCallback)((e) => {
		e.task_id && ot({
			taskId: e.task_id,
			label: e.name
		});
	}, [ot]), ut = (0, j.useCallback)(() => {
		ot({ label: "the run" });
	}, [ot]), dt = (0, j.useCallback)(async (e) => {
		if (!ie) {
			ae(!0);
			try {
				let t = await fetch(Vo("/api/lint/ignore", n), {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(e.req)
				});
				if (!t.ok) {
					T(`Ignore failed: ${(await t.text()).trim()}`);
					return;
				}
				T("Ignore rule saved to .gavel.yaml"), y.tab !== "files" && E({
					...y,
					selectedPath: ""
				}, "replace");
			} catch (e) {
				T(`Ignore error: ${e instanceof Error ? e.message : String(e)}`);
			} finally {
				ae(!1);
			}
		}
	}, [
		ie,
		n,
		y,
		E,
		T
	]), ft = (0, j.useCallback)((e) => {
		if (e.kind !== "lint") return [];
		let t = Te.get(Rl(e));
		if (!t) throw Error(`no linter recorded for the finding at ${e.file}:${e.line}`);
		let n = vn({
			violation: {
				...e.rule ? { rule: { method: e.rule } } : {},
				...e.line ? { line: e.line } : {}
			},
			linter: t,
			file: e.file,
			workDir: e.workDir ?? ""
		}), r = Ee.get(Rl(e));
		return [...r ? [{
			label: "Open original source",
			onSelect: () => {
				window.open(r, "_blank", "noopener,noreferrer");
			}
		}] : [], ...n.map((e) => ({
			label: e.label,
			title: e.title,
			disabled: ie || !!e.disabledWithoutWorkDir && !e.req.work_dir,
			onSelect: () => void dt(e)
		}))];
	}, [
		Te,
		Ee,
		ie,
		dt
	]), pt = (0, j.useCallback)(async (e, t, n) => {
		if (oe || c.test_edit_supported !== !0) return;
		let r = n === "file" ? e.file || "file" : e.name || "test";
		se(!0), T(t === "skip" ? `Skipping ${r}...` : `Deleting ${r}...`);
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
				T(`Test edit failed: ${r.message ?? r.status}`);
				return;
			}
			T(r.message || `${t === "skip" ? "Skip" : "Delete"} ${n === "file" ? "file" : "test"} saved`), t === "delete" && E({
				...y,
				selectedPath: ""
			}, "replace");
		} catch (e) {
			T(`Test edit error: ${e instanceof Error ? e.message : String(e)}`);
		} finally {
			se(!1);
		}
	}, [
		oe,
		c.test_edit_supported,
		i,
		y,
		E,
		T
	]), mt = (0, j.useMemo)(() => ({
		...y,
		selectedPath: ""
	}), [y]), ht = (0, j.useMemo)(() => ({
		nodeRouteState: (e) => e.export_path ? {
			...y,
			tab: "tests",
			selectedPath: e.export_path
		} : void 0,
		failingOnlyRouteState: (e) => e.export_path ? {
			...y,
			tab: "tests",
			selectedPath: e.export_path,
			filters: {
				...y.filters,
				status: /* @__PURE__ */ new Map([["failed", "include"], ["timedout", "include"]])
			}
		} : void 0,
		testEditSupported: c.test_edit_supported === !0,
		onTestEdit: pt,
		testEditBusy: oe,
		onRetryFromStep: rt,
		onCopyRerunCommand: (e, t) => void it(e, t),
		rerunBusy: te,
		runInProgress: c.running
	}), [
		y,
		c.test_edit_supported,
		c.running,
		pt,
		oe,
		rt,
		it,
		te
	]), gt = (0, j.useMemo)(() => $s({
		traces: De,
		processes: Ae.nodes,
		onOpenProcess: Ve,
		actions: ht,
		scope: n,
		promptOwner: l?.prompt_owner
	}), [
		De,
		Ae.nodes,
		Ve,
		ht,
		n,
		l?.prompt_owner
	]), vt = (0, j.useMemo)(() => il({
		samples: Ae.samples,
		nodes: Ae.nodes
	}), [Ae.samples, Ae.nodes]), yt = _, bt = we.length > 0 || y.tab === "files", xt = !!i.bench || !!m, St = Oe, Ct = i.bench?.deltas?.filter((e) => e.significant && e.delta_pct > i.bench.threshold).length || 0, wt = y.tab === "tests" && ve.length > 0 || y.tab === "lint" && _ || y.tab === "bench" && xt, Tt = c.running && !!c.stop_supported, kt = [
		{
			id: "tests",
			label: "Tests",
			icon: _t,
			count: D.failed > 0 ? D.failed : D.total,
			countColor: D.failed > 0 ? "bg-red-500" : "bg-muted-foreground"
		},
		...yt ? [{
			id: "lint",
			label: "Lint",
			icon: Ce,
			count: k,
			countColor: k > 0 ? "bg-yellow-500" : "bg-muted-foreground"
		}] : [],
		...bt ? [{
			id: "files",
			label: "Files",
			icon: We,
			count: we.length,
			countColor: we.length > 0 ? "bg-yellow-500" : "bg-muted-foreground"
		}] : [],
		...xt ? [{
			id: "bench",
			label: "Bench",
			icon: sn,
			count: m ? jl(o) : Ct > 0 ? Ct : i.bench?.deltas?.length || 0,
			countColor: Ct > 0 ? "bg-red-500" : "bg-muted-foreground"
		}] : [],
		...St ? [{
			id: "processes",
			label: "Processes",
			icon: on,
			count: je,
			countColor: "bg-blue-500"
		}] : [],
		...v.available ? [{
			id: "history",
			label: "History",
			icon: rn
		}] : []
	], At = kt.length > 1 || v.available, jt = y.tab === "tests" || !At, Mt = /* @__PURE__ */ (0, M.jsx)(Bc, {
		owner: l?.prompt_owner,
		scope: n
	}), Nt = /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [jt && Mt, /* @__PURE__ */ (0, M.jsx)(zc, {
		runMeta: l,
		canStop: Tt,
		stopping: ce === Kl,
		onStop: ut,
		exportState: wt ? mt : void 0
	})] }), Pt = ue ?? f ?? d, { __gavelBackTo: It, __gavelShards: Lt } = window;
	return /* @__PURE__ */ (0, M.jsxs)("div", {
		className: "flex h-screen flex-col bg-background text-foreground",
		children: [
			(It || Lt) && /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex shrink-0 items-center gap-3 border-b border-border bg-card px-4 py-1.5 text-sm",
				children: [It && /* @__PURE__ */ (0, M.jsxs)("a", {
					href: It,
					className: "inline-flex shrink-0 items-center gap-1 text-primary no-underline hover:underline",
					children: [/* @__PURE__ */ (0, M.jsx)(he, {}), "Back to PR"]
				}), Lt && /* @__PURE__ */ (0, M.jsx)(nl, {
					page: Lt,
					location: {
						basePath: Bo(),
						pathname: window.location.pathname,
						search: window.location.search
					}
				})]
			}),
			At && /* @__PURE__ */ (0, M.jsxs)("div", {
				className: "flex shrink-0 items-center gap-2 pr-density-4",
				children: [
					/* @__PURE__ */ (0, M.jsx)(Kt, {
						tabs: kt,
						value: y.tab,
						onChange: Fe,
						className: "min-w-0 flex-1 px-density-4"
					}),
					v.available && /* @__PURE__ */ (0, M.jsx)(Xc, {
						history: v,
						navigate: t
					}),
					!jt && Mt
				]
			}),
			/* @__PURE__ */ (0, M.jsxs)("div", {
				className: "min-h-0 flex-1",
				children: [
					y.tab === "tests" && (ve.length === 0 ? /* @__PURE__ */ (0, M.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, M.jsx)(Ul, {
							icon: _t,
							title: "Test Results",
							status: Pt,
							actions: Nt
						}), /* @__PURE__ */ (0, M.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, M.jsx)(Wl, { done: u })
						})]
					}) : /* @__PURE__ */ (0, M.jsx)(h, {
						tests: ye,
						allTests: ve,
						selected: be,
						filters: y.filters,
						expandAll: x,
						done: u,
						...l ? { runMeta: l } : {},
						status: c,
						statusText: ye.length === 0 ? `${Pt} · No tests match the current filters` : Pt,
						now: pe,
						startTime: ge,
						endTime: _e,
						busy: {
							rerun: te,
							stop: ce !== null
						},
						onSelect: Ie,
						onFiltersChange: Le,
						onExpandAllChange: ee,
						onRerun: nt,
						onStop: lt,
						fetchProcessMetrics: vt,
						adapters: gt,
						onOpenFile: Ke,
						headerSlot: Nt
					})),
					y.tab === "lint" && /* @__PURE__ */ (0, M.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, M.jsx)(kc, { tests: o }), /* @__PURE__ */ (0, M.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, M.jsx)(Sr, {
								results: s ?? [],
								grouping: y.lintGrouping,
								onGroupingChange: ze,
								filters: y.lintFilters,
								onFiltersChange: Re,
								selected: O ? y.selectedPath : null,
								onSelect: Be,
								expandAll: x,
								onExpandAllChange: ee,
								onIgnore: dt,
								ignoreBusy: ie,
								onOpenFile: Ke,
								loadSource: Ze,
								renderRun: Fl,
								headerSlot: /* @__PURE__ */ (0, M.jsxs)(M.Fragment, { children: [
									/* @__PURE__ */ (0, M.jsx)("span", {
										className: "truncate text-sm text-muted-foreground",
										children: Pt
									}),
									/* @__PURE__ */ (0, M.jsxs)(fe, {
										type: "button",
										size: "sm",
										disabled: te,
										onClick: at,
										title: O ? "Rerun the linters for this selection" : "Rerun the linters",
										children: [/* @__PURE__ */ (0, M.jsx)(p, { className: "mr-1" }), te ? "Running..." : "Rerun"]
									}),
									Nt
								] })
							})
						})]
					}),
					y.tab === "files" && /* @__PURE__ */ (0, M.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, M.jsx)(Ul, {
							icon: We,
							title: "Files",
							status: Pt,
							actions: Nt
						}), /* @__PURE__ */ (0, M.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, M.jsx)(j.Suspense, {
								fallback: /* @__PURE__ */ (0, M.jsx)(Il, {}),
								children: /* @__PURE__ */ (0, M.jsx)(Nl, {
									findings: we,
									selected: A,
									onSelect: qe,
									line: y.line,
									onLineChange: Je,
									view: y.view ?? "preview",
									onViewChange: Ye,
									problemFilters: y.problemFilters,
									onProblemFiltersChange: Xe,
									loadFile: Ze,
									getFindingActions: ft
								})
							})
						})]
					}),
					y.tab === "bench" && /* @__PURE__ */ (0, M.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, M.jsx)(Ul, {
							icon: sn,
							title: m ? "Fixture Benchmark" : "Benchmark Comparison",
							status: Pt,
							actions: Nt
						}), /* @__PURE__ */ (0, M.jsx)("div", {
							className: "min-h-0 flex-1 overflow-auto",
							children: m ? /* @__PURE__ */ (0, M.jsx)(hc, {
								performance: m,
								tests: o,
								error: g,
								sampleIndex: y.sampleIndex,
								onSampleChange: Qe
							}) : /* @__PURE__ */ (0, M.jsx)(sc, { bench: i.bench })
						})]
					}),
					y.tab === "history" && /* @__PURE__ */ (0, M.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, M.jsx)(Ul, {
							icon: rn,
							title: "Run History",
							status: v.current ? `Viewing ${v.current}` : "Viewing the latest results"
						}), /* @__PURE__ */ (0, M.jsx)("div", {
							className: "flex min-h-0 flex-1 flex-col p-density-2",
							children: /* @__PURE__ */ (0, M.jsx)(qc, {
								history: v,
								navigate: t
							})
						})]
					}),
					y.tab === "processes" && /* @__PURE__ */ (0, M.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, M.jsx)(Ul, {
							icon: on,
							title: "Processes",
							status: Pt,
							actions: Nt
						}), /* @__PURE__ */ (0, M.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, M.jsx)(Jc, {
								trace: Ae,
								selectedKey: Ne?.key ?? null,
								onSelect: Ve,
								hideExited: !!y.hideExited,
								onHideExitedChange: He,
								expandAll: x,
								onExpandAllChange: ee,
								stepLinks: Pe,
								onOpenStep: $e
							})
						})]
					})
				]
			}),
			/* @__PURE__ */ (0, M.jsx)(Fc, {
				open: re,
				onClose: () => S(!1),
				streamUrl: Vo("/api/rerun/stream", n)
			})
		]
	});
}
// src/index.tsx
var Yl = document.createElement("style");
Yl.textContent = Qt, document.head.appendChild(Yl), (0, Zt.createRoot)(document.getElementById("root")).render(/* @__PURE__ */ (0, M.jsx)(Jl, {}));
