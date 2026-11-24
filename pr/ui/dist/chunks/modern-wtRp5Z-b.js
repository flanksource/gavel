import { t as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-a1ZmknZu.js";
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@tanstack/query-core/5.101.1/7e19765689cbccf24c1f78df4f522110979ff2e4db631a13f55549f0e91ee24f/node_modules/@tanstack/query-core/build/modern/subscribable.cjs
var r = /* @__PURE__ */ e(((e, t) => {
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
})), i = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, i = Object.getOwnPropertyDescriptor, a = Object.getOwnPropertyNames, o = Object.prototype.hasOwnProperty, s = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, c = (e, t, r, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of a(t)) !o.call(e, c) && c !== r && n(e, c, {
			get: () => t[c],
			enumerable: !(s = i(t, c)) || s.enumerable
		});
		return e;
	}, l = (e) => c(n({}, "__esModule", { value: !0 }), e), u = {};
	s(u, {
		FocusManager: () => f,
		focusManager: () => p
	}), t.exports = l(u);
	var d = r(), f = class extends d.Subscribable {
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
	}, p = new f();
	0 && (t.exports = {
		FocusManager: f,
		focusManager: p
	});
})), a = /* @__PURE__ */ e(((e, t) => {
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
})), o = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, o = Object.prototype.hasOwnProperty, s = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, c = (e, t, a, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !o.call(e, c) && c !== a && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, l = (e) => c(n({}, "__esModule", { value: !0 }), e), u = {};
	s(u, {
		addConsumeAwareSignal: () => R,
		addToEnd: () => N,
		addToStart: () => P,
		ensureQueryFn: () => I,
		functionalUpdate: () => m,
		hashKey: () => S,
		hashQueryKeyByOptions: () => x,
		isPlainArray: () => D,
		isPlainObject: () => O,
		isServer: () => f,
		isValidTimeout: () => h,
		keepPreviousData: () => M,
		matchMutation: () => b,
		matchQuery: () => y,
		noop: () => p,
		partialMatchKey: () => C,
		replaceData: () => j,
		replaceEqualDeep: () => T,
		resolveQueryBoolean: () => v,
		resolveStaleTime: () => _,
		shallowEqualObjects: () => E,
		shouldThrowError: () => L,
		skipToken: () => F,
		sleep: () => A,
		timeUntilStale: () => g
	}), t.exports = l(u);
	var d = a(), f = typeof window > "u" || "Deno" in globalThis;
	function p() {}
	function m(e, t) {
		return typeof e == "function" ? e(t) : e;
	}
	function h(e) {
		return typeof e == "number" && e >= 0 && e !== Infinity;
	}
	function g(e, t) {
		return Math.max(e + (t || 0) - Date.now(), 0);
	}
	function _(e, t) {
		return typeof e == "function" ? e(t) : e;
	}
	function v(e, t) {
		return typeof e == "function" ? e(t) : e;
	}
	function y(e, t) {
		let { type: n = "all", exact: r, fetchStatus: i, predicate: a, queryKey: o, stale: s } = e;
		if (o) {
			if (r) {
				if (t.queryHash !== x(o, t.options)) return !1;
			} else if (!C(t.queryKey, o)) return !1;
		}
		if (n !== "all") {
			let e = t.isActive();
			if (n === "active" && !e || n === "inactive" && e) return !1;
		}
		return !(typeof s == "boolean" && t.isStale() !== s || i && i !== t.state.fetchStatus || a && !a(t));
	}
	function b(e, t) {
		let { exact: n, status: r, predicate: i, mutationKey: a } = e;
		if (a) {
			if (!t.options.mutationKey) return !1;
			if (n) {
				if (S(t.options.mutationKey) !== S(a)) return !1;
			} else if (!C(t.options.mutationKey, a)) return !1;
		}
		return !(r && t.state.status !== r || i && !i(t));
	}
	function x(e, t) {
		return (t?.queryKeyHashFn || S)(e);
	}
	function S(e) {
		return JSON.stringify(e, (e, t) => O(t) ? Object.keys(t).sort().reduce((e, n) => (e[n] = t[n], e), {}) : t);
	}
	function C(e, t) {
		return e === t ? !0 : typeof e == typeof t && e && t && typeof e == "object" && typeof t == "object" ? Object.keys(t).every((n) => C(e[n], t[n])) : !1;
	}
	var w = Object.prototype.hasOwnProperty;
	function T(e, t, n = 0) {
		if (e === t) return e;
		if (n > 500) return t;
		let r = D(e) && D(t);
		if (!r && !(O(e) && O(t))) return t;
		let i = (r ? e : Object.keys(e)).length, a = r ? t : Object.keys(t), o = a.length, s = r ? Array(o) : {}, c = 0;
		for (let l = 0; l < o; l++) {
			let o = r ? l : a[l], u = e[o], d = t[o];
			if (u === d) {
				s[o] = u, (r ? l < i : w.call(e, o)) && c++;
				continue;
			}
			if (u === null || d === null || typeof u != "object" || typeof d != "object") {
				s[o] = d;
				continue;
			}
			let f = T(u, d, n + 1);
			s[o] = f, f === u && c++;
		}
		return i === o && c === i ? e : s;
	}
	function E(e, t) {
		if (!t || Object.keys(e).length !== Object.keys(t).length) return !1;
		for (let n in e) if (e[n] !== t[n]) return !1;
		return !0;
	}
	function D(e) {
		return Array.isArray(e) && e.length === Object.keys(e).length;
	}
	function O(e) {
		if (!k(e)) return !1;
		let t = e.constructor;
		if (t === void 0) return !0;
		let n = t.prototype;
		return !(!k(n) || !n.hasOwnProperty("isPrototypeOf") || Object.getPrototypeOf(e) !== Object.prototype);
	}
	function k(e) {
		return Object.prototype.toString.call(e) === "[object Object]";
	}
	function A(e) {
		return new Promise((t) => {
			d.timeoutManager.setTimeout(t, e);
		});
	}
	function j(e, t, n) {
		return typeof n.structuralSharing == "function" ? n.structuralSharing(e, t) : n.structuralSharing === !1 ? t : T(e, t);
	}
	function M(e) {
		return e;
	}
	function N(e, t, n = 0) {
		let r = [...e, t];
		return n && r.length > n ? r.slice(1) : r;
	}
	function P(e, t, n = 0) {
		let r = [t, ...e];
		return n && r.length > n ? r.slice(0, -1) : r;
	}
	var F = /* @__PURE__ */ Symbol();
	function I(e, t) {
		return !e.queryFn && t?.initialPromise ? () => t.initialPromise : !e.queryFn || e.queryFn === F ? () => Promise.reject(/* @__PURE__ */ Error(`Missing queryFn: '${e.queryHash}'`)) : e.queryFn;
	}
	function L(e, t) {
		return typeof e == "function" ? e(...t) : !!e;
	}
	function R(e, t, n) {
		let r = !1, i;
		return Object.defineProperty(e, "signal", {
			enumerable: !0,
			get: () => (i ??= t(), r || (r = !0, i.aborted ? n() : i.addEventListener("abort", n, { once: !0 })), i)
		}), e;
	}
	0 && (t.exports = {
		addConsumeAwareSignal: R,
		addToEnd: N,
		addToStart: P,
		ensureQueryFn: I,
		functionalUpdate: m,
		hashKey: S,
		hashQueryKeyByOptions: x,
		isPlainArray: D,
		isPlainObject: O,
		isServer: f,
		isValidTimeout: h,
		keepPreviousData: M,
		matchMutation: b,
		matchQuery: y,
		noop: p,
		partialMatchKey: C,
		replaceData: j,
		replaceEqualDeep: T,
		resolveQueryBoolean: v,
		resolveStaleTime: _,
		shallowEqualObjects: E,
		shouldThrowError: L,
		skipToken: F,
		sleep: A,
		timeUntilStale: g
	});
})), s = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, s = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, c = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, l = (e) => c(n({}, "__esModule", { value: !0 }), e), u = {};
	s(u, { environmentManager: () => f }), t.exports = l(u);
	var d = o(), f = /* @__PURE__ */ (() => {
		let e = () => d.isServer;
		return {
			isServer() {
				return e();
			},
			setIsServer(t) {
				e = t;
			}
		};
	})();
	0 && (t.exports = { environmentManager: f });
})), c = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, s = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, c = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, l = (e) => c(n({}, "__esModule", { value: !0 }), e), u = {};
	s(u, {
		pendingThenable: () => f,
		tryResolveSync: () => p
	}), t.exports = l(u);
	var d = o();
	function f() {
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
	function p(e) {
		let t;
		if (e.then((e) => (t = e, e), d.noop)?.catch(d.noop), t !== void 0) return { data: t };
	}
	0 && (t.exports = {
		pendingThenable: f,
		tryResolveSync: p
	});
})), l = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, s = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, l = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, u = (e) => l(n({}, "__esModule", { value: !0 }), e), d = {};
	s(d, {
		defaultShouldDehydrateMutation: () => _,
		defaultShouldDehydrateQuery: () => v,
		dehydrate: () => b,
		hydrate: () => x
	}), t.exports = u(d);
	var f = c(), p = o();
	function m(e) {
		return e;
	}
	function h(e) {
		return {
			mutationKey: e.options.mutationKey,
			state: e.state,
			...e.options.scope && { scope: e.options.scope },
			...e.meta && { meta: e.meta }
		};
	}
	function g(e, t, n) {
		let r = () => {
			let r = e.promise?.then(t).catch((e) => n(e) ? Promise.reject(/* @__PURE__ */ Error("redacted")) : Promise.reject(e));
			return r?.catch(p.noop), r;
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
	function _(e) {
		return e.state.isPaused;
	}
	function v(e) {
		return e.state.status === "success";
	}
	function y(e) {
		return !0;
	}
	function b(e, t = {}) {
		let n = t.shouldDehydrateMutation ?? e.getDefaultOptions().dehydrate?.shouldDehydrateMutation ?? _, r = e.getMutationCache().getAll().flatMap((e) => n(e) ? [h(e)] : []), i = t.shouldDehydrateQuery ?? e.getDefaultOptions().dehydrate?.shouldDehydrateQuery ?? v, a = t.shouldRedactErrors ?? e.getDefaultOptions().dehydrate?.shouldRedactErrors ?? y, o = t.serializeData ?? e.getDefaultOptions().dehydrate?.serializeData ?? m;
		return {
			mutations: r,
			queries: e.getQueryCache().getAll().flatMap((e) => i(e) ? [g(e, o, a)] : [])
		};
	}
	function x(e, t, n) {
		if (typeof t != "object" || !t) return;
		let r = e.getMutationCache(), i = e.getQueryCache(), a = n?.defaultOptions?.deserializeData ?? e.getDefaultOptions().hydrate?.deserializeData ?? m, o = t.mutations || [], s = t.queries || [];
		o.forEach(({ state: t, ...i }) => {
			r.build(e, {
				...e.getDefaultOptions().hydrate?.mutations,
				...n?.defaultOptions?.mutations,
				...i
			}, t);
		}), s.forEach(({ queryKey: t, state: r, queryHash: o, meta: s, promise: c, dehydratedAt: l, queryType: u }) => {
			let d = c ? (0, f.tryResolveSync)(c) : void 0, m = r.data === void 0 ? d?.data : r.data, h = m === void 0 ? m : a(m), g = i.get(o), _ = g?.state.status === "pending", v = g?.state.fetchStatus === "fetching";
			if (g) {
				let e = d && l !== void 0 && l > g.state.dataUpdatedAt;
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
				_type: u
			}, {
				...r,
				data: h,
				fetchStatus: "idle",
				status: r.status === "pending" && h !== void 0 ? "success" : r.status,
				...r.status === "pending" && h !== void 0 && { dataUpdatedAt: l ?? Date.now() }
			});
			c && !d && !_ && !v && (l === void 0 || l > g.state.dataUpdatedAt) && g.fetch(void 0, { initialPromise: Promise.resolve(c).then(a) }).catch(p.noop);
		});
	}
	0 && (t.exports = {
		defaultShouldDehydrateMutation: _,
		defaultShouldDehydrateQuery: v,
		dehydrate: b,
		hydrate: x
	});
})), u = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, o = Object.prototype.hasOwnProperty, s = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, c = (e, t, a, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !o.call(e, c) && c !== a && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, l = (e) => c(n({}, "__esModule", { value: !0 }), e), u = {};
	s(u, {
		createNotifyManager: () => f,
		defaultScheduler: () => d,
		notifyManager: () => p
	}), t.exports = l(u);
	var d = a().systemSetTimeoutZero;
	function f() {
		let e = [], t = 0, n = (e) => {
			e();
		}, r = (e) => {
			e();
		}, i = d, a = (r) => {
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
	var p = f();
	0 && (t.exports = {
		createNotifyManager: f,
		defaultScheduler: d,
		notifyManager: p
	});
})), d = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, i = Object.getOwnPropertyDescriptor, a = Object.getOwnPropertyNames, o = Object.prototype.hasOwnProperty, s = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, c = (e, t, r, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of a(t)) !o.call(e, c) && c !== r && n(e, c, {
			get: () => t[c],
			enumerable: !(s = i(t, c)) || s.enumerable
		});
		return e;
	}, l = (e) => c(n({}, "__esModule", { value: !0 }), e), u = {};
	s(u, {
		OnlineManager: () => f,
		onlineManager: () => p
	}), t.exports = l(u);
	var d = r(), f = class extends d.Subscribable {
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
	}, p = new f();
	0 && (t.exports = {
		OnlineManager: f,
		onlineManager: p
	});
})), f = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, a = Object.getOwnPropertyNames, l = Object.prototype.hasOwnProperty, u = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, f = (e, t, i, o) => {
		if (t && typeof t == "object" || typeof t == "function") for (let s of a(t)) !l.call(e, s) && s !== i && n(e, s, {
			get: () => t[s],
			enumerable: !(o = r(t, s)) || o.enumerable
		});
		return e;
	}, p = (e) => f(n({}, "__esModule", { value: !0 }), e), m = {};
	u(m, {
		CancelledError: () => S,
		canFetch: () => x,
		createRetryer: () => w,
		isCancelledError: () => C
	}), t.exports = p(m);
	var h = i(), g = d(), _ = c(), v = s(), y = o();
	function b(e) {
		return Math.min(1e3 * 2 ** e, 3e4);
	}
	function x(e) {
		return (e ?? "online") !== "online" || g.onlineManager.isOnline();
	}
	var S = class extends Error {
		constructor(e) {
			super("CancelledError"), this.revert = e?.revert, this.silent = e?.silent;
		}
	};
	function C(e) {
		return e instanceof S;
	}
	function w(e) {
		let t = !1, n = 0, r, i = (0, _.pendingThenable)(), a = () => i.status !== "pending", o = (t) => {
			if (!a()) {
				let n = new S(t);
				f(n), e.onCancel?.(n);
			}
		}, s = () => {
			t = !0;
		}, c = () => {
			t = !1;
		}, l = () => h.focusManager.isFocused() && (e.networkMode === "always" || g.onlineManager.isOnline()) && e.canRun(), u = () => x(e.networkMode) && e.canRun(), d = (e) => {
			a() || (r?.(), i.resolve(e));
		}, f = (e) => {
			a() || (r?.(), i.reject(e));
		}, p = () => new Promise((t) => {
			r = (e) => {
				(a() || l()) && t(e);
			}, e.onPause?.();
		}).then(() => {
			r = void 0, a() || e.onContinue?.();
		}), m = () => {
			if (a()) return;
			let r, i = n === 0 ? e.initialPromise : void 0;
			try {
				r = i ?? e.fn();
			} catch (e) {
				r = Promise.reject(e);
			}
			Promise.resolve(r).then(d).catch((r) => {
				if (a()) return;
				let i = e.retry ?? (v.environmentManager.isServer() ? 0 : 3), o = e.retryDelay ?? b, s = typeof o == "function" ? o(n, r) : o, c = i === !0 || typeof i == "number" && n < i || typeof i == "function" && i(n, r);
				t || !c ? f(r) : (n++, e.onFail?.(n, r), (0, y.sleep)(s).then(() => l() ? void 0 : p()).then(() => {
					t ? f(r) : m();
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
			canStart: u,
			start: () => (u() ? m() : p().then(m), i)
		};
	}
	0 && (t.exports = {
		CancelledError: S,
		canFetch: x,
		createRetryer: w,
		isCancelledError: C
	});
})), p = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, c = Object.prototype.hasOwnProperty, l = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, u = (e, t, a, o) => {
		if (t && typeof t == "object" || typeof t == "function") for (let s of i(t)) !c.call(e, s) && s !== a && n(e, s, {
			get: () => t[s],
			enumerable: !(o = r(t, s)) || o.enumerable
		});
		return e;
	}, d = (e) => u(n({}, "__esModule", { value: !0 }), e), f = {};
	l(f, { Removable: () => g }), t.exports = d(f);
	var p = a(), m = s(), h = o(), g = class {
		#e;
		destroy() {
			this.clearGcTimeout();
		}
		scheduleGc() {
			this.clearGcTimeout(), (0, h.isValidTimeout)(this.gcTime) && (this.#e = p.timeoutManager.setTimeout(() => {
				this.optionalRemove();
			}, this.gcTime));
		}
		updateGcTime(e) {
			this.gcTime = Math.max(this.gcTime || 0, e ?? (m.environmentManager.isServer() ? Infinity : 3e5));
		}
		clearGcTimeout() {
			this.#e !== void 0 && (p.timeoutManager.clearTimeout(this.#e), this.#e = void 0);
		}
	};
	0 && (t.exports = { Removable: g });
})), m = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, s = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, c = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, l = (e) => c(n({}, "__esModule", { value: !0 }), e), u = {};
	s(u, {
		hasNextPage: () => h,
		hasPreviousPage: () => g,
		infiniteQueryBehavior: () => f
	}), t.exports = l(u);
	var d = o();
	function f(e) {
		return { onFetch: (t, n) => {
			let r = t.options, i = t.fetchOptions?.meta?.fetchMore?.direction, a = t.state.data?.pages || [], o = t.state.data?.pageParams || [], s = {
				pages: [],
				pageParams: []
			}, c = 0, l = async () => {
				let n = !1, l = (e) => {
					(0, d.addConsumeAwareSignal)(e, () => t.signal, () => n = !0);
				}, u = (0, d.ensureQueryFn)(t.options, t.fetchOptions), f = async (e, r, i) => {
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
					})(), o = await u(a), { maxPages: s } = t.options, c = i ? d.addToStart : d.addToEnd;
					return {
						pages: c(e.pages, o, s),
						pageParams: c(e.pageParams, r, s)
					};
				};
				if (i && a.length) {
					let e = i === "backward", t = e ? m : p, n = {
						pages: a,
						pageParams: o
					};
					s = await f(n, t(r, n), e);
				} else {
					let t = e ?? a.length;
					do {
						let e = c === 0 ? o[0] ?? r.initialPageParam : p(r, s);
						if (c > 0 && e == null) break;
						s = await f(s, e), c++;
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
	function p(e, { pages: t, pageParams: n }) {
		let r = t.length - 1;
		return t.length > 0 ? e.getNextPageParam(t[r], t, n[r], n) : void 0;
	}
	function m(e, { pages: t, pageParams: n }) {
		return t.length > 0 ? e.getPreviousPageParam?.(t[0], t, n[0], n) : void 0;
	}
	function h(e, t) {
		return t ? p(e, t) != null : !1;
	}
	function g(e, t) {
		return !t || !e.getPreviousPageParam ? !1 : m(e, t) != null;
	}
	0 && (t.exports = {
		hasNextPage: h,
		hasPreviousPage: g,
		infiniteQueryBehavior: f
	});
})), h = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, s = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, c = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, l = (e) => c(n({}, "__esModule", { value: !0 }), e), d = {};
	s(d, {
		Query: () => b,
		fetchState: () => x
	}), t.exports = l(d);
	var h = o(), g = u(), _ = f(), v = p(), y = m(), b = class extends v.Removable {
		#e;
		#t;
		#n;
		#r;
		#i;
		#a;
		#o;
		#s;
		constructor(e) {
			super(), this.#s = !1, this.#o = e.defaultOptions, this.setOptions(e.options), this.observers = [], this.#i = e.client, this.#r = this.#i.getQueryCache(), this.queryKey = e.queryKey, this.queryHash = e.queryHash, this.#t = C(this.options), this.state = e.state ?? this.#t, this.scheduleGc();
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
				let e = C(this.options);
				e.data !== void 0 && (this.setState(S(e.data, e.dataUpdatedAt)), this.#t = e);
			}
		}
		optionalRemove() {
			!this.observers.length && this.state.fetchStatus === "idle" && this.#r.remove(this);
		}
		setData(e, t) {
			let n = (0, h.replaceData)(this.state.data, e, this.options);
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
			return this.#a?.cancel(e), t ? t.then(h.noop).catch(h.noop) : Promise.resolve();
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
			return this.observers.some((e) => (0, h.resolveQueryBoolean)(e.options.enabled, this) !== !1);
		}
		isDisabled() {
			return this.getObserversCount() > 0 ? !this.isActive() : this.options.queryFn === h.skipToken || !this.isFetched();
		}
		isFetched() {
			return this.state.dataUpdateCount + this.state.errorUpdateCount > 0;
		}
		isStatic() {
			return this.getObserversCount() > 0 && this.observers.some((e) => (0, h.resolveStaleTime)(e.options.staleTime, this) === "static");
		}
		isStale() {
			return this.getObserversCount() > 0 ? this.observers.some((e) => e.getCurrentResult().isStale) : this.state.data === void 0 || this.state.isInvalidated;
		}
		isStaleByTime(e = 0) {
			return this.state.data === void 0 ? !0 : e === "static" ? !1 : this.state.isInvalidated ? !0 : !(0, h.timeUntilStale)(this.state.dataUpdatedAt, e);
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
				let e = (0, h.ensureQueryFn)(this.options, t), n = (() => {
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
			(this.#e === "infinite" ? (0, y.infiniteQueryBehavior)(this.options.pages) : this.options.behavior)?.onFetch(a, this), this.#n = this.state, (this.state.fetchStatus === "idle" || this.state.fetchMeta !== a.fetchOptions?.meta) && this.#l({
				type: "fetch",
				meta: a.fetchOptions?.meta
			}), this.#a = (0, _.createRetryer)({
				initialPromise: t?.initialPromise,
				fn: a.fetchFn,
				onCancel: (e) => {
					e instanceof _.CancelledError && e.revert && this.setState({
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
				if (e instanceof _.CancelledError) {
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
						...x(t.data, this.options),
						fetchMeta: e.meta ?? null
					};
					case "success":
						let n = {
							...t,
							...S(e.data, e.dataUpdatedAt),
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
			this.state = t(this.state), g.notifyManager.batch(() => {
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
	function x(e, t) {
		return {
			fetchFailureCount: 0,
			fetchFailureReason: null,
			fetchStatus: (0, _.canFetch)(t.networkMode) ? "fetching" : "paused",
			...e === void 0 && {
				error: null,
				status: "pending"
			}
		};
	}
	function S(e, t) {
		return {
			data: e,
			dataUpdatedAt: t ?? Date.now(),
			error: null,
			isInvalidated: !1,
			status: "success"
		};
	}
	function C(e) {
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
		Query: b,
		fetchState: x
	});
})), g = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, l = Object.getOwnPropertyDescriptor, d = Object.getOwnPropertyNames, f = Object.prototype.hasOwnProperty, p = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, m = (e, t, r, i) => {
		if (t && typeof t == "object" || typeof t == "function") for (let a of d(t)) !f.call(e, a) && a !== r && n(e, a, {
			get: () => t[a],
			enumerable: !(i = l(t, a)) || i.enumerable
		});
		return e;
	}, g = (e) => m(n({}, "__esModule", { value: !0 }), e), _ = {};
	p(_, { QueryObserver: () => E }), t.exports = g(_);
	var v = i(), y = s(), b = u(), x = h(), S = r(), C = c(), w = o(), T = a(), E = class extends S.Subscribable {
		constructor(e, t) {
			super(), this.options = t, this.#e = e, this.#s = null, this.#o = (0, C.pendingThenable)(), this.bindMethods(), this.setOptions(t);
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
			this.listeners.size === 1 && (this.#t.addObserver(this), O(this.#t, this.options) ? this.#h() : this.updateResult(), this.#y());
		}
		onUnsubscribe() {
			this.hasListeners() || this.destroy();
		}
		shouldFetchOnReconnect() {
			return k(this.#t, this.options, this.options.refetchOnReconnect);
		}
		shouldFetchOnWindowFocus() {
			return k(this.#t, this.options, this.options.refetchOnWindowFocus);
		}
		destroy() {
			this.listeners = /* @__PURE__ */ new Set(), this.#b(), this.#x(), this.#t.removeObserver(this);
		}
		setOptions(e) {
			let t = this.options, n = this.#t;
			if (this.options = this.#e.defaultQueryOptions(e), this.options.enabled !== void 0 && typeof this.options.enabled != "boolean" && typeof this.options.enabled != "function" && typeof (0, w.resolveQueryBoolean)(this.options.enabled, this.#t) != "boolean") throw Error("Expected enabled to be a boolean or a callback that returns a boolean");
			this.#S(), this.#t.setOptions(this.options), t._defaulted && !(0, w.shallowEqualObjects)(this.options, t) && this.#e.getQueryCache().notify({
				type: "observerOptionsUpdated",
				query: this.#t,
				observer: this
			});
			let r = this.hasListeners();
			r && A(this.#t, n, this.options, t) && this.#h(), this.updateResult(), r && (this.#t !== n || (0, w.resolveQueryBoolean)(this.options.enabled, this.#t) !== (0, w.resolveQueryBoolean)(t.enabled, this.#t) || (0, w.resolveStaleTime)(this.options.staleTime, this.#t) !== (0, w.resolveStaleTime)(t.staleTime, this.#t)) && this.#g();
			let i = this.#_();
			r && (this.#t !== n || (0, w.resolveQueryBoolean)(this.options.enabled, this.#t) !== (0, w.resolveQueryBoolean)(t.enabled, this.#t) || i !== this.#p) && this.#v(i);
		}
		getOptimisticResult(e) {
			let t = this.#e.getQueryCache().build(this.#e, e), n = this.createResult(t, e);
			return M(this, n) && (this.#r = n, this.#a = this.options, this.#i = this.#t.state), n;
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
			return e?.throwOnError || (t = t.catch(w.noop)), t;
		}
		#g() {
			this.#b();
			let e = (0, w.resolveStaleTime)(this.options.staleTime, this.#t);
			if (y.environmentManager.isServer() || this.#r.isStale || !(0, w.isValidTimeout)(e)) return;
			let t = (0, w.timeUntilStale)(this.#r.dataUpdatedAt, e) + 1;
			this.#d = T.timeoutManager.setTimeout(() => {
				this.#r.isStale || this.updateResult();
			}, t);
		}
		#_() {
			return (typeof this.options.refetchInterval == "function" ? this.options.refetchInterval(this.#t) : this.options.refetchInterval) ?? !1;
		}
		#v(e) {
			this.#x(), this.#p = e, !y.environmentManager.isServer() && (0, w.resolveQueryBoolean)(this.options.enabled, this.#t) !== !1 && (0, w.isValidTimeout)(this.#p) && this.#p !== 0 && (this.#f = T.timeoutManager.setInterval(() => {
				(this.options.refetchIntervalInBackground || v.focusManager.isFocused()) && this.#h();
			}, this.#p));
		}
		#y() {
			this.#g(), this.#v(this.#_());
		}
		#b() {
			this.#d !== void 0 && (T.timeoutManager.clearTimeout(this.#d), this.#d = void 0);
		}
		#x() {
			this.#f !== void 0 && (T.timeoutManager.clearInterval(this.#f), this.#f = void 0);
		}
		createResult(e, t) {
			let n = this.#t, r = this.options, i = this.#r, a = this.#i, o = this.#a, s = e === n ? this.#n : e.state, { state: c } = e, l = { ...c }, u = !1, d;
			if (t._optimisticResults) {
				let i = this.hasListeners(), a = !i && O(e, t), o = i && A(e, n, t, r);
				(a || o) && (l = {
					...l,
					...(0, x.fetchState)(c.data, e.options)
				}), t._optimisticResults === "isRestoring" && (l.fetchStatus = "idle");
			}
			let { error: f, errorUpdatedAt: p, status: m } = l;
			d = l.data;
			let h = !1;
			if (t.placeholderData !== void 0 && d === void 0 && m === "pending") {
				let e;
				i?.isPlaceholderData && t.placeholderData === o?.placeholderData ? (e = i.data, h = !0) : e = typeof t.placeholderData == "function" ? t.placeholderData(this.#u?.state.data, this.#u) : t.placeholderData, e !== void 0 && (m = "success", d = (0, w.replaceData)(i?.data, e, t), u = !0);
			}
			if (t.select && d !== void 0 && !h) {
				if (i && d === a?.data && t.select === this.#c) d = this.#l;
				else try {
					this.#c = t.select, d = t.select(d), d = (0, w.replaceData)(i?.data, d, t), this.#l = d, this.#s = null;
				} catch (e) {
					this.#s = e;
				}
			}
			this.#s && (f = this.#s, d = this.#l, p = Date.now(), m = "error");
			let g = l.fetchStatus === "fetching", _ = m === "pending", v = m === "error", y = _ && g, b = d !== void 0, S = {
				status: m,
				fetchStatus: l.fetchStatus,
				isPending: _,
				isSuccess: m === "success",
				isError: v,
				isInitialLoading: y,
				isLoading: y,
				data: d,
				dataUpdatedAt: l.dataUpdatedAt,
				error: f,
				errorUpdatedAt: p,
				failureCount: l.fetchFailureCount,
				failureReason: l.fetchFailureReason,
				errorUpdateCount: l.errorUpdateCount,
				isFetched: e.isFetched(),
				isFetchedAfterMount: l.dataUpdateCount > s.dataUpdateCount || l.errorUpdateCount > s.errorUpdateCount,
				isFetching: g,
				isRefetching: g && !_,
				isLoadingError: v && !b,
				isPaused: l.fetchStatus === "paused",
				isPlaceholderData: u,
				isRefetchError: v && b,
				isStale: j(e, t),
				refetch: this.refetch,
				promise: this.#o,
				isEnabled: (0, w.resolveQueryBoolean)(t.enabled, e) !== !1
			};
			if (this.options.experimental_prefetchInRender) {
				let t = S.data !== void 0, r = S.status === "error" && !t, i = (e) => {
					r ? e.reject(S.error) : t && e.resolve(S.data);
				}, a = () => {
					let e = this.#o = S.promise = (0, C.pendingThenable)();
					i(e);
				}, o = this.#o;
				switch (o.status) {
					case "pending":
						e.queryHash === n.queryHash && i(o);
						break;
					case "fulfilled":
						(r || S.data !== o.value) && a();
						break;
					case "rejected": r && S.error === o.reason || a();
				}
			}
			return S;
		}
		updateResult() {
			let e = this.#r, t = this.createResult(this.#t, this.options);
			this.#i = this.#t.state, this.#a = this.options, this.#i.data !== void 0 && (this.#u = this.#t), !(0, w.shallowEqualObjects)(t, e) && (this.#r = t, this.#C({ listeners: (() => {
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
			b.notifyManager.batch(() => {
				e.listeners && this.listeners.forEach((e) => {
					e(this.#r);
				}), this.#e.getQueryCache().notify({
					query: this.#t,
					type: "observerResultsUpdated"
				});
			});
		}
	};
	function D(e, t) {
		return (0, w.resolveQueryBoolean)(t.enabled, e) !== !1 && e.state.data === void 0 && (e.state.status !== "error" || (0, w.resolveQueryBoolean)(t.retryOnMount, e) !== !1);
	}
	function O(e, t) {
		return D(e, t) || e.state.data !== void 0 && k(e, t, t.refetchOnMount);
	}
	function k(e, t, n) {
		if ((0, w.resolveQueryBoolean)(t.enabled, e) !== !1 && (0, w.resolveStaleTime)(t.staleTime, e) !== "static") {
			let r = typeof n == "function" ? n(e) : n;
			return r === "always" || r !== !1 && j(e, t);
		}
		return !1;
	}
	function A(e, t, n, r) {
		return (e !== t || (0, w.resolveQueryBoolean)(r.enabled, e) === !1) && (!n.suspense || e.state.status !== "error") && j(e, n);
	}
	function j(e, t) {
		return (0, w.resolveQueryBoolean)(t.enabled, e) !== !1 && e.isStaleByTime((0, w.resolveStaleTime)(t.staleTime, e));
	}
	function M(e, t) {
		return !(0, w.shallowEqualObjects)(e.getCurrentResult(), t);
	}
	0 && (t.exports = { QueryObserver: E });
})), _ = /* @__PURE__ */ e(((e, t) => {
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
	var u = g(), d = m(), f = class extends u.QueryObserver {
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
})), v = /* @__PURE__ */ e(((e, t) => {
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
		Mutation: () => g,
		getDefaultState: () => _
	}), t.exports = c(l);
	var d = u(), m = p(), h = f(), g = class extends m.Removable {
		#e;
		#t;
		#n;
		#r;
		constructor(e) {
			super(), this.#e = e.client, this.mutationId = e.mutationId, this.#n = e.mutationCache, this.#t = [], this.state = e.state || _(), this.setOptions(e.options), this.scheduleGc();
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
			this.#r = (0, h.createRetryer)({
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
			this.state = t(this.state), d.notifyManager.batch(() => {
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
	function _() {
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
		Mutation: g,
		getDefaultState: _
	});
})), y = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, i = Object.getOwnPropertyDescriptor, a = Object.getOwnPropertyNames, s = Object.prototype.hasOwnProperty, c = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, l = (e, t, r, o) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of a(t)) !s.call(e, c) && c !== r && n(e, c, {
			get: () => t[c],
			enumerable: !(o = i(t, c)) || o.enumerable
		});
		return e;
	}, d = (e) => l(n({}, "__esModule", { value: !0 }), e), f = {};
	c(f, { MutationCache: () => _ }), t.exports = d(f);
	var p = u(), m = v(), h = o(), g = r(), _ = class extends g.Subscribable {
		constructor(e = {}) {
			super(), this.config = e, this.#e = /* @__PURE__ */ new Set(), this.#t = /* @__PURE__ */ new Map(), this.#n = 0;
		}
		#e;
		#t;
		#n;
		build(e, t, n) {
			let r = new m.Mutation({
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
			let t = y(e);
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
				let t = y(e);
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
			let t = y(e);
			if (typeof t == "string") {
				let n = this.#t.get(t)?.find((e) => e.state.status === "pending");
				return !n || n === e;
			}
			return !0;
		}
		runNext(e) {
			let t = y(e);
			return typeof t == "string" ? (this.#t.get(t)?.find((t) => t !== e && t.state.isPaused))?.continue() ?? Promise.resolve() : Promise.resolve();
		}
		clear() {
			p.notifyManager.batch(() => {
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
			return this.getAll().find((e) => (0, h.matchMutation)(t, e));
		}
		findAll(e = {}) {
			return this.getAll().filter((t) => (0, h.matchMutation)(e, t));
		}
		notify(e) {
			p.notifyManager.batch(() => {
				this.listeners.forEach((t) => {
					t(e);
				});
			});
		}
		resumePausedMutations() {
			let e = this.getAll().filter((e) => e.state.isPaused);
			return p.notifyManager.batch(() => Promise.all(e.map((e) => e.continue().catch(h.noop))));
		}
	};
	function y(e) {
		return e.options.scope?.id;
	}
	0 && (t.exports = { MutationCache: _ });
})), b = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, i = Object.getOwnPropertyDescriptor, a = Object.getOwnPropertyNames, s = Object.prototype.hasOwnProperty, c = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, l = (e, t, r, o) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of a(t)) !s.call(e, c) && c !== r && n(e, c, {
			get: () => t[c],
			enumerable: !(o = i(t, c)) || o.enumerable
		});
		return e;
	}, d = (e) => l(n({}, "__esModule", { value: !0 }), e), f = {};
	c(f, { MutationObserver: () => _ }), t.exports = d(f);
	var p = v(), m = u(), h = r(), g = o(), _ = class extends h.Subscribable {
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
			this.options = this.#e.defaultMutationOptions(e), (0, g.shallowEqualObjects)(this.options, t) || this.#e.getMutationCache().notify({
				type: "observerOptionsUpdated",
				mutation: this.#n,
				observer: this
			}), t?.mutationKey && this.options.mutationKey && (0, g.hashKey)(t.mutationKey) !== (0, g.hashKey)(this.options.mutationKey) ? this.reset() : this.#n?.state.status === "pending" && this.#n.setOptions(this.options);
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
			let e = this.#n?.state ?? (0, p.getDefaultState)();
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
			m.notifyManager.batch(() => {
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
	0 && (t.exports = { MutationObserver: _ });
})), x = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, i = Object.getOwnPropertyDescriptor, a = Object.getOwnPropertyNames, s = Object.prototype.hasOwnProperty, c = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, l = (e, t, r, o) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of a(t)) !s.call(e, c) && c !== r && n(e, c, {
			get: () => t[c],
			enumerable: !(o = i(t, c)) || o.enumerable
		});
		return e;
	}, d = (e) => l(n({}, "__esModule", { value: !0 }), e), f = {};
	c(f, { QueriesObserver: () => b }), t.exports = d(f);
	var p = u(), m = g(), h = r(), _ = o();
	function v(e, t) {
		let n = new Set(t);
		return e.filter((e) => !n.has(e));
	}
	function y(e, t, n) {
		let r = e.slice(0);
		return r[t] = n, r;
	}
	var b = class extends h.Subscribable {
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
			this.#n = e, this.#r = t, p.notifyManager.batch(() => {
				let e = this.#i, t = this.#p(this.#n);
				t.forEach((e) => e.observer.setOptions(e.defaultedQueryOptions));
				let n = t.map((e) => e.observer), r = n.map((e) => e.getCurrentResult()), i = e.length !== n.length, a = n.some((t, n) => t !== e[n]), o = i || a, s = o ? !0 : r.some((e, t) => {
					let n = this.#t[t];
					return !n || !(0, _.shallowEqualObjects)(e, n);
				});
				(o || s) && (o && (this.#l = t, this.#i = n), this.#t = r, this.hasListeners() && (o && (v(e, n).forEach((e) => {
					e.destroy();
				}), v(n, e).forEach((e) => {
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
				return (!this.#a || this.#t !== this.#s || i || t !== this.#o) && (this.#o = t, this.#s = this.#t, n !== void 0 && (this.#c = n), this.#a = (0, _.replaceEqualDeep)(this.#a, t(e))), this.#a;
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
				let r = this.#e.defaultQueryOptions(e), i = t.get(r.queryHash)?.shift() ?? new m.QueryObserver(this.#e, r);
				n.push({
					defaultedQueryOptions: r,
					observer: i
				});
			}), n;
		}
		#m(e, t) {
			let n = this.#i.indexOf(e);
			n !== -1 && (this.#t = y(this.#t, n, t), this.#h());
		}
		#h() {
			if (this.hasListeners()) {
				let e = this.#u(this.#t, this.#l), t = this.#f(), n = this.#a, r = t ? n : this.#d(e, this.#r?.combine);
				(t || n !== r) && p.notifyManager.batch(() => {
					this.listeners.forEach((e) => {
						e(this.#t);
					});
				});
			}
		}
	};
	0 && (t.exports = { QueriesObserver: b });
})), S = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, i = Object.getOwnPropertyDescriptor, a = Object.getOwnPropertyNames, s = Object.prototype.hasOwnProperty, c = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, l = (e, t, r, o) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of a(t)) !s.call(e, c) && c !== r && n(e, c, {
			get: () => t[c],
			enumerable: !(o = i(t, c)) || o.enumerable
		});
		return e;
	}, d = (e) => l(n({}, "__esModule", { value: !0 }), e), f = {};
	c(f, { QueryCache: () => v }), t.exports = d(f);
	var p = o(), m = h(), g = u(), _ = r(), v = class extends _.Subscribable {
		constructor(e = {}) {
			super(), this.config = e, this.#e = /* @__PURE__ */ new Map();
		}
		#e;
		build(e, t, n) {
			let r = t.queryKey, i = t.queryHash ?? (0, p.hashQueryKeyByOptions)(r, t), a = this.get(i);
			return a || (a = new m.Query({
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
			g.notifyManager.batch(() => {
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
			return this.getAll().find((e) => (0, p.matchQuery)(t, e));
		}
		findAll(e = {}) {
			let t = this.getAll();
			return Object.keys(e).length > 0 ? t.filter((t) => (0, p.matchQuery)(e, t)) : t;
		}
		notify(e) {
			g.notifyManager.batch(() => {
				this.listeners.forEach((t) => {
					t(e);
				});
			});
		}
		onFocus() {
			g.notifyManager.batch(() => {
				this.getAll().forEach((e) => {
					e.onFocus();
				});
			});
		}
		onOnline() {
			g.notifyManager.batch(() => {
				this.getAll().forEach((e) => {
					e.onOnline();
				});
			});
		}
	};
	0 && (t.exports = { QueryCache: v });
})), C = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, a = Object.getOwnPropertyNames, s = Object.prototype.hasOwnProperty, c = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, l = (e, t, i, o) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of a(t)) !s.call(e, c) && c !== i && n(e, c, {
			get: () => t[c],
			enumerable: !(o = r(t, c)) || o.enumerable
		});
		return e;
	}, f = (e) => l(n({}, "__esModule", { value: !0 }), e), p = {};
	c(p, { QueryClient: () => x }), t.exports = f(p);
	var m = o(), h = S(), g = y(), _ = i(), v = d(), b = u(), x = class {
		#e;
		#t;
		#n;
		#r;
		#i;
		#a;
		#o;
		#s;
		constructor(e = {}) {
			this.#e = e.queryCache || new h.QueryCache(), this.#t = e.mutationCache || new g.MutationCache(), this.#n = e.defaultOptions || {}, this.#r = /* @__PURE__ */ new Map(), this.#i = /* @__PURE__ */ new Map(), this.#a = 0;
		}
		mount() {
			this.#a++, this.#a === 1 && (this.#o = _.focusManager.subscribe(async (e) => {
				e && (await this.resumePausedMutations(), this.#e.onFocus());
			}), this.#s = v.onlineManager.subscribe(async (e) => {
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
			return r === void 0 ? this.fetchQuery(e) : (e.revalidateIfStale && n.isStaleByTime((0, m.resolveStaleTime)(t.staleTime, n)) && this.prefetchQuery(t), Promise.resolve(r));
		}
		getQueriesData(e) {
			return this.#e.findAll(e).map(({ queryKey: e, state: t }) => [e, t.data]);
		}
		setQueryData(e, t, n) {
			let r = this.defaultQueryOptions({ queryKey: e }), i = this.#e.get(r.queryHash)?.state.data, a = (0, m.functionalUpdate)(t, i);
			if (a !== void 0) return this.#e.build(this, r).setData(a, {
				...n,
				manual: !0
			});
		}
		setQueriesData(e, t, n) {
			return b.notifyManager.batch(() => this.#e.findAll(e).map(({ queryKey: e }) => [e, this.setQueryData(e, t, n)]));
		}
		getQueryState(e) {
			let t = this.defaultQueryOptions({ queryKey: e });
			return this.#e.get(t.queryHash)?.state;
		}
		removeQueries(e) {
			let t = this.#e;
			b.notifyManager.batch(() => {
				t.findAll(e).forEach((e) => {
					t.remove(e);
				});
			});
		}
		resetQueries(e, t) {
			let n = this.#e;
			return b.notifyManager.batch(() => (n.findAll(e).forEach((e) => {
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
			}, r = b.notifyManager.batch(() => this.#e.findAll(e).map((e) => e.cancel(n)));
			return Promise.all(r).then(m.noop).catch(m.noop);
		}
		invalidateQueries(e, t = {}) {
			return b.notifyManager.batch(() => (this.#e.findAll(e).forEach((e) => {
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
			}, r = b.notifyManager.batch(() => this.#e.findAll(e).filter((e) => !e.isDisabled() && !e.isStatic()).map((e) => {
				let t = e.fetch(void 0, n);
				return n.throwOnError || (t = t.catch(m.noop)), e.state.fetchStatus === "paused" ? Promise.resolve() : t;
			}));
			return Promise.all(r).then(m.noop);
		}
		fetchQuery(e) {
			let t = this.defaultQueryOptions(e);
			t.retry === void 0 && (t.retry = !1);
			let n = this.#e.build(this, t);
			return n.isStaleByTime((0, m.resolveStaleTime)(t.staleTime, n)) ? n.fetch(t) : Promise.resolve(n.state.data);
		}
		prefetchQuery(e) {
			return this.fetchQuery(e).then(m.noop).catch(m.noop);
		}
		fetchInfiniteQuery(e) {
			return e._type = "infinite", this.fetchQuery(e);
		}
		prefetchInfiniteQuery(e) {
			return this.fetchInfiniteQuery(e).then(m.noop).catch(m.noop);
		}
		ensureInfiniteQueryData(e) {
			return e._type = "infinite", this.ensureQueryData(e);
		}
		resumePausedMutations() {
			return v.onlineManager.isOnline() ? this.#t.resumePausedMutations() : Promise.resolve();
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
			this.#r.set((0, m.hashKey)(e), {
				queryKey: e,
				defaultOptions: t
			});
		}
		getQueryDefaults(e) {
			let t = [...this.#r.values()], n = {};
			return t.forEach((t) => {
				(0, m.partialMatchKey)(e, t.queryKey) && Object.assign(n, t.defaultOptions);
			}), n;
		}
		setMutationDefaults(e, t) {
			this.#i.set((0, m.hashKey)(e), {
				mutationKey: e,
				defaultOptions: t
			});
		}
		getMutationDefaults(e) {
			let t = [...this.#i.values()], n = {};
			return t.forEach((t) => {
				(0, m.partialMatchKey)(e, t.mutationKey) && Object.assign(n, t.defaultOptions);
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
			return t.queryHash ||= (0, m.hashQueryKeyByOptions)(t.queryKey, t), t.refetchOnReconnect === void 0 && (t.refetchOnReconnect = t.networkMode !== "always"), t.throwOnError === void 0 && (t.throwOnError = !!t.suspense), !t.networkMode && t.persister && (t.networkMode = "offlineFirst"), t.queryFn === m.skipToken && (t.enabled = !1), t;
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
	0 && (t.exports = { QueryClient: x });
})), w = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, s = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, c = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	}, l = (e) => c(n({}, "__esModule", { value: !0 }), e), u = {};
	s(u, { streamedQuery: () => f }), t.exports = l(u);
	var d = o();
	function f({ streamFn: e, refetchMode: t = "reset", reducer: n = (e, t) => (0, d.addToEnd)(e, t), initialValue: r = [] }) {
		return async (i) => {
			let a = i.client.getQueryCache().find({
				queryKey: i.queryKey,
				exact: !0
			}), o = !!a && a.isFetched();
			o && t === "reset" && a.setState({
				...a.resetState,
				fetchStatus: "fetching"
			});
			let s = r, c = !1, l = await e((0, d.addConsumeAwareSignal)({
				client: i.client,
				meta: i.meta,
				queryKey: i.queryKey,
				pageParam: i.pageParam,
				direction: i.direction
			}, () => i.signal, () => c = !0)), u = o && t === "replace";
			for await (let e of l) {
				if (c) break;
				u ? s = n(s, e) : i.client.setQueryData(i.queryKey, (t) => n(t === void 0 ? r : t, e));
			}
			return u && !c && i.client.setQueryData(i.queryKey, s), i.client.getQueryData(i.queryKey) ?? r;
		};
	}
	0 && (t.exports = { streamedQuery: f });
})), T = /* @__PURE__ */ e(((e, t) => {
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
})), E = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, c = Object.getOwnPropertyNames, p = Object.prototype.hasOwnProperty, m = (e, t) => {
		for (var r in t) n(e, r, {
			get: t[r],
			enumerable: !0
		});
	}, E = (e, t, i, a) => {
		if (t && typeof t == "object" || typeof t == "function") for (let o of c(t)) !p.call(e, o) && o !== i && n(e, o, {
			get: () => t[o],
			enumerable: !(a = r(t, o)) || a.enumerable
		});
		return e;
	}, D = (e, t, n) => (E(e, t, "default"), n && E(n, t, "default")), O = (e) => E(n({}, "__esModule", { value: !0 }), e), k = {};
	m(k, {
		CancelledError: () => H.CancelledError,
		InfiniteQueryObserver: () => N.InfiniteQueryObserver,
		Mutation: () => K.Mutation,
		MutationCache: () => P.MutationCache,
		MutationObserver: () => F.MutationObserver,
		QueriesObserver: () => R.QueriesObserver,
		Query: () => q.Query,
		QueryCache: () => z.QueryCache,
		QueryClient: () => B.QueryClient,
		QueryObserver: () => V.QueryObserver,
		defaultScheduler: () => I.defaultScheduler,
		defaultShouldDehydrateMutation: () => M.defaultShouldDehydrateMutation,
		defaultShouldDehydrateQuery: () => M.defaultShouldDehydrateQuery,
		dehydrate: () => M.dehydrate,
		environmentManager: () => j.environmentManager,
		experimental_streamedQuery: () => G.streamedQuery,
		focusManager: () => A.focusManager,
		hashKey: () => W.hashKey,
		hydrate: () => M.hydrate,
		isCancelledError: () => H.isCancelledError,
		isServer: () => W.isServer,
		keepPreviousData: () => W.keepPreviousData,
		matchMutation: () => W.matchMutation,
		matchQuery: () => W.matchQuery,
		noop: () => W.noop,
		notifyManager: () => I.notifyManager,
		onlineManager: () => L.onlineManager,
		partialMatchKey: () => W.partialMatchKey,
		replaceEqualDeep: () => W.replaceEqualDeep,
		shouldThrowError: () => W.shouldThrowError,
		skipToken: () => W.skipToken,
		timeoutManager: () => U.timeoutManager
	}), t.exports = O(k);
	var A = i(), j = s(), M = l(), N = _(), P = y(), F = b(), I = u(), L = d(), R = x(), z = S(), B = C(), V = g(), H = f(), U = a(), W = o(), G = w(), K = v(), q = h();
	D(k, T(), t.exports), 0 && (t.exports = {
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
		...T()
	});
})), D = /* @__PURE__ */ e(((e, t) => {
	var n = Object.defineProperty, r = Object.getOwnPropertyDescriptor, i = Object.getOwnPropertyNames, a = Object.prototype.hasOwnProperty, o = (e, t, o, s) => {
		if (t && typeof t == "object" || typeof t == "function") for (let c of i(t)) !a.call(e, c) && c !== o && n(e, c, {
			get: () => t[c],
			enumerable: !(s = r(t, c)) || s.enumerable
		});
		return e;
	};
	t.exports = ((e) => o(n({}, "__esModule", { value: !0 }), e))({});
})), O = /* @__PURE__ */ e(((e, r) => {
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
	}), r.exports = p(m);
	var h = f(t(), 1), g = n(), _ = h.createContext(void 0), v = (e) => {
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
	0 && (r.exports = {
		QueryClientContext: _,
		QueryClientProvider: y,
		useQueryClient: v
	});
})), k = /* @__PURE__ */ e(((e, n) => {
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
	}), n.exports = f(p);
	var m = d(t(), 1), h = m.createContext(!1), g = () => m.useContext(h), _ = h.Provider;
	0 && (n.exports = {
		IsRestoringProvider: _,
		useIsRestoring: g
	});
})), A = /* @__PURE__ */ e(((e, r) => {
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
	}), r.exports = p(m);
	var h = f(t(), 1), g = n();
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
	0 && (r.exports = {
		QueryErrorResetBoundary: b,
		useQueryErrorResetBoundary: y
	});
})), j = /* @__PURE__ */ e(((e, n) => {
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
	}), n.exports = f(p);
	var m = d(t(), 1), h = E(), g = (e, t, n) => {
		let r = n?.state.error && typeof e.throwOnError == "function" ? (0, h.shouldThrowError)(e.throwOnError, [n.state.error, n]) : e.throwOnError;
		(e.suspense || e.experimental_prefetchInRender || r) && (t.isReset() || (e.retryOnMount = !1));
	}, _ = (e) => {
		m.useEffect(() => {
			e.clearReset();
		}, [e]);
	}, v = ({ result: e, errorResetBoundary: t, throwOnError: n, query: r, suspense: i }) => e.isError && !t.isReset() && !e.isFetching && r && (i && e.data === void 0 || (0, h.shouldThrowError)(n, [e.error, r]));
	0 && (n.exports = {
		ensurePreventErrorBoundaryRetry: g,
		getHasError: v,
		useClearResetErrorBoundary: _
	});
})), M = /* @__PURE__ */ e(((e, t) => {
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
})), N = /* @__PURE__ */ e(((e, n) => {
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
	l(p, { useQueries: () => x }), n.exports = f(p);
	var m = d(t(), 1), h = E(), g = O(), _ = k(), v = A(), y = j(), b = M();
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
	0 && (n.exports = { useQueries: x });
})), P = /* @__PURE__ */ e(((e, n) => {
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
	l(p, { useBaseQuery: () => x }), n.exports = f(p);
	var m = d(t(), 1), h = E(), g = O(), _ = A(), v = j(), y = k(), b = M();
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
	0 && (n.exports = { useBaseQuery: x });
})), F = /* @__PURE__ */ e(((e, t) => {
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
	var u = E(), d = P();
	function f(e, t) {
		return (0, d.useBaseQuery)(e, u.QueryObserver, t);
	}
	0 && (t.exports = { useQuery: f });
})), I = /* @__PURE__ */ e(((e, t) => {
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
	var u = E(), d = P(), f = M();
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
})), L = /* @__PURE__ */ e(((e, t) => {
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
	var u = E(), d = P(), f = M();
	function p(e, t) {
		return (0, d.useBaseQuery)({
			...e,
			enabled: !0,
			suspense: !0,
			throwOnError: f.defaultThrowOnError
		}, u.InfiniteQueryObserver, t);
	}
	0 && (t.exports = { useSuspenseInfiniteQuery: p });
})), R = /* @__PURE__ */ e(((e, t) => {
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
	o(l, { useSuspenseQueries: () => f }), t.exports = c(l), E();
	var u = N(), d = M();
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
})), z = /* @__PURE__ */ e(((e, t) => {
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
	var u = O();
	function d(e, t) {
		let n = (0, u.useQueryClient)(t);
		n.getQueryState(e.queryKey) || n.prefetchQuery(e);
	}
	0 && (t.exports = { usePrefetchQuery: d });
})), B = /* @__PURE__ */ e(((e, t) => {
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
	var u = O();
	function d(e, t) {
		let n = (0, u.useQueryClient)(t);
		n.getQueryState(e.queryKey) || n.prefetchInfiniteQuery(e);
	}
	0 && (t.exports = { usePrefetchInfiniteQuery: d });
})), V = /* @__PURE__ */ e(((e, t) => {
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
})), H = /* @__PURE__ */ e(((e, t) => {
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
})), U = /* @__PURE__ */ e(((e, n) => {
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
	l(p, { HydrationBoundary: () => _ }), n.exports = f(p);
	var m = d(t(), 1), h = E(), g = O(), _ = ({ children: e, options: t = {}, state: n, queryClient: r }) => {
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
	0 && (n.exports = { HydrationBoundary: _ });
})), W = /* @__PURE__ */ e(((e, n) => {
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
	l(p, { useIsFetching: () => _ }), n.exports = f(p);
	var m = d(t(), 1), h = E(), g = O();
	function _(e, t) {
		let n = (0, g.useQueryClient)(t), r = n.getQueryCache();
		return m.useSyncExternalStore(m.useCallback((e) => r.subscribe(h.notifyManager.batchCalls(e)), [r]), () => n.isFetching(e), () => n.isFetching(e));
	}
	0 && (n.exports = { useIsFetching: _ });
})), G = /* @__PURE__ */ e(((e, n) => {
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
	}), n.exports = f(p);
	var m = d(t(), 1), h = E(), g = O();
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
	0 && (n.exports = {
		useIsMutating: _,
		useMutationState: y
	});
})), K = /* @__PURE__ */ e(((e, n) => {
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
	l(p, { useMutation: () => _ }), n.exports = f(p);
	var m = d(t(), 1), h = E(), g = O();
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
	0 && (n.exports = { useMutation: _ });
})), q = /* @__PURE__ */ e(((e, t) => {
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
})), J = /* @__PURE__ */ e(((e, t) => {
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
	var u = E(), d = P();
	function f(e, t) {
		return (0, d.useBaseQuery)(e, u.InfiniteQueryObserver, t);
	}
	0 && (t.exports = { useInfiniteQuery: f });
})), Y = /* @__PURE__ */ e(((e, t) => {
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
		IsRestoringProvider: () => P.IsRestoringProvider,
		QueryClientContext: () => b.QueryClientContext,
		QueryClientProvider: () => b.QueryClientProvider,
		QueryErrorResetBoundary: () => S.QueryErrorResetBoundary,
		infiniteQueryOptions: () => y.infiniteQueryOptions,
		mutationOptions: () => j.mutationOptions,
		queryOptions: () => v.queryOptions,
		useInfiniteQuery: () => M.useInfiniteQuery,
		useIsFetching: () => C.useIsFetching,
		useIsMutating: () => w.useIsMutating,
		useIsRestoring: () => P.useIsRestoring,
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
	}), t.exports = l(u), c(u, E(), t.exports), c(u, D(), t.exports);
	var d = N(), f = F(), p = I(), m = L(), h = R(), g = z(), _ = B(), v = V(), y = H(), b = O(), x = U(), S = A(), C = W(), w = G(), T = K(), j = q(), M = J(), P = k();
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
		...E(),
		...D()
	});
}));
export { Y as t };
