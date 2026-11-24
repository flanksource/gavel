(function() {
	const n = `bippy-0.6.1`;
	const r = Object.defineProperty;
	const i = Object.prototype.hasOwnProperty;
	const a = () => {
		/**
		* @license bippy
		*
		* Copyright (c) Aiden Bai
		*
		* This source code is licensed under the MIT license found in the
		* LICENSE file in the root directory of this source tree.
		*/
	};
	const o = (e) => {
		try {
			Function.prototype.toString.call(e).indexOf(`^_^`) > -1 && setTimeout(() => {
				throw Error(`React is running in production mode, but dead code elimination has not been applied. Read how to correctly configure React for production: https://reactjs.org/link/perf-use-production-build`);
			});
		} catch {}
	};
	const s = (e = globalThis.__REACT_DEVTOOLS_GLOBAL_HOOK__) => !!(e && `getFiberRoots` in e);
	let c = !1;
	let l;
	const u = (e = globalThis.__REACT_DEVTOOLS_GLOBAL_HOOK__) => c ? !0 : (e && typeof e.inject == `function` && (l = e.inject.toString()), !!l?.includes(`(injected)`));
	const d = /* @__PURE__ */ new Set();
	const f = /* @__PURE__ */ new Set();
	const v = (e) => {
		e && d.add(e);
		let t = /* @__PURE__ */ new Map(), i = 0, s = {
			_instrumentationIsActive: !1,
			_instrumentationSource: n,
			checkDCE: o,
			hasUnsupportedRendererAttached: !1,
			inject(e) {
				let n = ++i;
				return t.set(n, e), f.add(e), s._instrumentationIsActive || (s._instrumentationIsActive = !0, d.forEach((e) => e())), n;
			},
			on: a,
			onCommitFiberRoot: a,
			onCommitFiberUnmount: a,
			onPostCommitFiberRoot: a,
			renderers: t,
			supportsFiber: !0,
			supportsFlight: !0
		};
		try {
			r(globalThis, `__REACT_DEVTOOLS_GLOBAL_HOOK__`, {
				configurable: !0,
				enumerable: !0,
				get() {
					return s;
				},
				set(t) {
					if (t && typeof t == `object`) {
						let n = s.renderers;
						s = t, n.size > 0 && (n.forEach((e, n) => {
							f.add(e), t.renderers.set(n, e);
						}), y(e));
					}
				}
			});
			let t = window.hasOwnProperty, n = !1;
			r(window, `hasOwnProperty`, {
				configurable: !0,
				value: function(...e) {
					try {
						if (!n && e[0] === `__REACT_DEVTOOLS_GLOBAL_HOOK__`) return globalThis.__REACT_DEVTOOLS_GLOBAL_HOOK__ = void 0, n = !0, -0;
					} catch {}
					return t.apply(this, e);
				},
				writable: !0
			});
		} catch {
			y(e);
		}
		return s;
	};
	const y = (e) => {
		e && d.add(e);
		try {
			let t = globalThis.__REACT_DEVTOOLS_GLOBAL_HOOK__;
			if (!t) return;
			if (!t._instrumentationSource) {
				t.checkDCE = o, t.supportsFiber = !0, t.supportsFlight = !0, t.hasUnsupportedRendererAttached = !1, t._instrumentationSource = n, t._instrumentationIsActive = !1;
				let e = s(t);
				if (e || (t.on = a), t.renderers.size) {
					t._instrumentationIsActive = !0, d.forEach((e) => e());
					return;
				}
				let r = t.inject, i = u(t);
				i && !e && (c = !0, t.inject({ scheduleRefresh() {} }) && (t._instrumentationIsActive = !0)), t.inject = (e) => {
					let n = r(e);
					return f.add(e), i && t.renderers.set(n, e), t._instrumentationIsActive = !0, d.forEach((e) => e()), n;
				};
			}
			(t.renderers.size || t._instrumentationIsActive || u()) && e?.();
		} catch {}
	};
	const b = () => i.call(globalThis, `__REACT_DEVTOOLS_GLOBAL_HOOK__`);
	const x = (e) => b() ? (y(e), globalThis.__REACT_DEVTOOLS_GLOBAL_HOOK__) : v(e);
	const S = () => !!(typeof window < `u` && (window.document?.createElement || window.navigator?.product === `ReactNative`));
	const C = () => {
		try {
			S() && x();
		} catch {}
	};
	// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/bippy/0.6.1/1130005d9ad5ea110510cf11f128db1358e7b7bbb7b2d18e02d0aee399b85c38/node_modules/bippy/dist/install-hook-only.js
	/**
	* @license bippy
	*
	* Copyright (c) Aiden Bai
	*
	* This source code is licensed under the MIT license found in the
	* LICENSE file in the root directory of this source tree.
	*/
	C();
})();
